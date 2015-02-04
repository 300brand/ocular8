package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/300brand/ocular8/types"
	"github.com/golang/glog"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Entry struct {
	Id        bson.ObjectId `bson:"_id"`
	FeedId    bson.ObjectId
	PubId     bson.ObjectId
	Link      string
	Author    string
	Title     string
	Published string
}

const TOPIC = "article.id.extract.goose"

var (
	dsn      = flag.String("mongo", "mongodb://localhost:27017/ocular8", "Connection string to MongoDB")
	nsqdhttp = flag.String("nsqdhttp", "http://localhost:4151", "NSQd HTTP address")
)

var (
	db     *mgo.Database
	nsqURL *url.URL
)

func process(entry *Entry) (err error) {
	prefix := fmt.Sprintf("P:%s F:%s E:%s", entry.PubId.Hex(), entry.FeedId.Hex(), entry.Id.Hex())
	start := time.Now()

	clean, resp, err := Clean(entry.Link)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	a := &types.Article{
		Id:     entry.Id,
		FeedId: entry.FeedId,
		PubId:  entry.PubId,
		Url:    clean,
		Author: entry.Author,
		Title:  entry.Title,
	}
	defer func(a *types.Article) { a.LoadTime = time.Since(start) }(a)

	if a.HTML, err = ioutil.ReadAll(resp.Body); err != nil {
		return
	}

	if a.Published, err = time.Parse(time.RFC1123, entry.Published); err != nil {
		glog.Warningf("%s %s", prefix, err)
		if a.Published, err = time.Parse(time.RFC1123Z, entry.Published); err != nil {
			glog.Warningf("%s %s", prefix, err)
		}
		err = nil
	}

	a.Entry.Url = entry.Link
	a.Entry.Title = entry.Title
	a.Entry.Author = entry.Author
	a.Entry.Published = a.Published

	glog.Infof("%s Moving to articles collection", prefix)
	if err = db.C("articles").Insert(a); err != nil {
		return
	}
	if err = db.C("entries").RemoveId(entry.Id); err != nil {
		return
	}

	payload := bytes.NewBufferString(entry.Id.Hex())
	if _, err = http.Post(nsqURL.String(), "multipart/form-data", payload); err != nil {
		return
	}
	glog.Infof("%s Sent to %s", prefix, TOPIC)

	return
}

func main() {
	flag.Parse()

	var err error
	nsqURL, err = url.Parse(*nsqdhttp)
	if err != nil {
		glog.Fatalf("Error parsing %s: %s", *nsqdhttp, err)
		return
	}
	nsqURL.Path = "/pub"
	nsqURL.RawQuery = (url.Values{"topic": []string{TOPIC}}).Encode()

	s, err := mgo.Dial(*dsn)
	if err != nil {
		glog.Fatalf("mgo.Dial(%s): %s", *dsn, err)
	}
	defer s.Close()
	db = s.DB("")

	change := mgo.Change{
		Update:    bson.M{"$inc": bson.M{"attempts": 1}},
		ReturnNew: false,
	}
	for _, id := range flag.Args() {
		if !bson.IsObjectIdHex(id) {
			glog.Fatalf("Invalid BSON ObjectId: %s", id)
		}

		q := bson.M{"_id": bson.ObjectIdHex(id)}
		entry := new(Entry)
		if _, err := db.C("entries").Find(q).Apply(change, entry); err != nil {
			glog.Fatalf("Find(%+v): %s", q, err)
		}
		if err := process(entry); err != nil {
			glog.Errorf("process(%s): %s", entry.Id.Hex(), err)
			// TODO switch the error type
			db.C("article_errors").Insert(bson.M{
				"url":    entry.Link,
				"type":   fmt.Sprintf("%T", err),
				"code":   0,
				"reason": err.Error(),
				"entry":  entry,
			})
		}
	}
}
