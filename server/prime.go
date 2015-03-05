package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"net/http"
	"time"

	"github.com/300brand/ocular8/types"
	"github.com/golang/glog"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	doPrime  = flag.Bool("prime", false, "Prime database with data from existing databases")
	primeRPC = flag.String("primerpc", "http://okcodev:52204/rpc", "Prime RPC address")
)

func prime(mongoDSN string) (err error) {
	s, err := mgo.Dial(mongoDSN)
	if err != nil {
		return
	}
	defer s.Close()

	db := s.DB("")
	cp := db.C("pubs")
	cf := db.C("feeds")
	ca := db.C("articles")

	glog.Infof("Empty pubs and feeds collections")
	cp.RemoveAll(nil)
	cf.RemoveAll(nil)

	glog.Infof("Ensure indexes")
	cf.EnsureIndex(mgo.Index{
		Key:        []string{"metabaseid"},
		Background: true,
		Sparse:     true,
		Unique:     true,
	})
	cf.EnsureIndexKey("pubid")
	cf.EnsureIndexKey("ignore")
	ca.EnsureIndexKey("feedid")
	ca.EnsureIndexKey("pubid")
	ca.EnsureIndex(mgo.Index{
		Key:        []string{"url"},
		Background: true,
		Sparse:     true,
		Unique:     true,
	})
	ca.EnsureIndex(mgo.Index{
		Key:        []string{"metabase.id"},
		Background: true,
		Sparse:     true,
		Unique:     true,
	})

	// Temporary structs to hold original data
	type P struct {
		Id          bson.ObjectId `json:"ID"`
		Title       string
		Url         string `json:"URL"`
		NumArticles int
		NumFeeds    int
		NumReaders  int
		Updated     time.Time
		XPaths      struct{ Author, Body, Date, Title []string }
	}
	type F struct {
		Id      bson.ObjectId `json:"ID"`
		Url     string        `json:"URL"`
		Updated time.Time
	}

	pubquery := bson.M{
		"id":     1,
		"method": "Publication.GetAll",
		"params": []bson.M{},
	}
	b, err := json.Marshal(pubquery)
	if err != nil {
		return
	}
	glog.Infof("Publication.GetAll from %s", *primeRPC)
	resp, err := http.Post(*primeRPC, "application/json", bytes.NewReader(b))
	if err != nil {
		return
	}
	defer resp.Body.Close()

	result := &struct {
		Result struct {
			Publications []P
		}
	}{}
	if err = json.NewDecoder(resp.Body).Decode(result); err != nil {
		return
	}

	pubFeeds := func(id bson.ObjectId) (err error) {
		feedquery := bson.M{
			"id":     id.Hex(),
			"method": "Publication.View",
			"params": []bson.M{
				bson.M{
					"Publication": id,
					"Feeds": bson.M{
						"Select": bson.M{
							"content": 0,
							"urls":    0,
							"log":     0,
						},
					},
					"Articles": bson.M{
						"Limit": 1,
						"Select": bson.M{
							"_id": 1,
						},
					},
				},
			},
		}

		b, err := json.Marshal(feedquery)
		if err != nil {
			return
		}
		resp, err := http.Post(*primeRPC, "application/json", bytes.NewReader(b))
		if err != nil {
			return
		}
		defer resp.Body.Close()

		result := &struct {
			Result struct {
				Feeds struct {
					Feeds []F
				}
			}
		}{}
		if err = json.NewDecoder(resp.Body).Decode(result); err != nil {
			return
		}

		for _, f := range result.Result.Feeds.Feeds {
			glog.Infof("FEED %s", f.Id.Hex())
			err = cf.Insert(types.Feed{
				Id:    f.Id,
				PubId: id,
				Url:   f.Url,
			})
			if err != nil {
				return
			}
		}
		cp.UpdateId(id, bson.M{"$set": bson.M{"numfeeds": len(result.Result.Feeds.Feeds)}})
		return
	}

	for _, p := range result.Result.Publications {
		glog.Infof("PUB  %s", p.Id.Hex())
		err = cp.Insert(types.Pub{
			Id:          p.Id,
			Name:        p.Title,
			Homepage:    p.Url,
			NumReaders:  p.NumReaders,
			LastUpdate:  p.Updated,
			XPathAuthor: p.XPaths.Author,
			XPathBody:   p.XPaths.Body,
			XPathDate:   p.XPaths.Date,
			XPathTitle:  p.XPaths.Title,
		})
		if err != nil {
			return
		}
		if err = pubFeeds(p.Id); err != nil {
			return
		}
	}

	return
}
