package main

import (
	"flag"
	"net/url"

	"github.com/300brand/ocular8/lib/metabase"
	"github.com/golang/glog"
	"gopkg.in/mgo.v2"
)

const TOPIC = "article.id.elastic"

var (
	dsn      = flag.String("mongo", "mongodb://localhost:27017/ocular8", "Connection string to MongoDB")
	nsqdhttp = flag.String("nsqdhttp", "http://localhost:4151", "NSQd HTTP address")
)

var (
	db     *mgo.Database
	nsqURL *url.URL
)

func main() {
	flag.Parse()

	nsqURL, err := url.Parse(*nsqdhttp)
	if err != nil {
		glog.Fatalf("Error parsing %s: %s", *nsqdhttp, err)
	}
	nsqURL.Path = "/mpub"
	nsqURL.RawQuery = (url.Values{"topic": []string{TOPIC}}).Encode()

	s, err := mgo.Dial(*dsn)
	if err != nil {
		glog.Fatalf("mgo.Dial(%s): %s", *dsn, err)
	}
	defer s.Close()
	db = s.DB("")

	apikey := "028505fb9ddb415781ff76c862b998d2"
	sequenceId := ""
	result, err := metabase.Fetch(apikey, sequenceId)
	if err != nil {
		glog.Fatalf("metabase.Fetch: %s", err)
	}

	glog.Infof("%#q", result)
}
