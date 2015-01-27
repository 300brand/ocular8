package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"time"
)

var doPrime = flag.Bool("prime", false, "Prime database with data from existing databases")

func prime() (err error) {
	query := bson.M{
		"id":     1,
		"method": "Publication.GetAll",
		"params": []bson.M{
			bson.M{
				"Sort":  "title",
				"Limit": 1000,
				"Skip":  0,
				"Query": bson.M{},
			},
		},
	}
	buf := bytes.NewBuffer(make([]byte, 512))
	json.NewEncoder(buf).Encode(query)
	resp, err := http.Post("http://okcodev:52204/rpc", "application/json", buf)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	type P struct {
		Id          bson.ObjectId `json:"ID"`
		Title       string
		Url         string `json:"URL"`
		NumArticles int
		NumFeeds    int
		NumReaders  int
		Updated     time.Time
		XPaths      map[string][]string
	}
	result := &struct {
		Result struct {
			Publications []P
		}
	}{}
	if err = json.NewDecoder(resp.Body).Decode(result); err != nil {
		return
	}
	return
}
