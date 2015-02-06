package main

import (
	"flag"
	"github.com/golang/glog"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)

var dsn = flag.String("mongo", "mongodb://localhost:27017/ocular8", "Mongo connection string")

var config = &struct {
	LastRun time.Time
	LastId  bson.ObjectId
}{
	time.Unix(0, 0),
	bson.NewObjectIdWithTime(time.Unix(0, 0)),
}
var lastId = new(struct {
	Id bson.ObjectId `bson:"_id"`
})

func main() {
	flag.Parse()

	s, err := mgo.Dial(*dsn)
	if err != nil {
		glog.Fatalf("mgo.Dial(%s): %s", *dsn, err)
	}
	defer s.Close()
	s.SetSocketTimeout(5 * time.Minute)
	db := s.DB("")
	c := db.C("article_counts")
	a := db.C("articles")

	if err := c.FindId("config").One(config); err != nil && err != mgo.ErrNotFound {
		glog.Fatalf("get last run: %s", err)
	}

	switch cnt, err := a.Count(); true {
	case err != nil:
		glog.Fatalf("Article count: %s", err)
	case cnt == 0:
		glog.Errorf("No articles to count")
		return
	}

	if err := a.Find(nil).Sort("-_id").One(lastId); err != nil {
		glog.Fatalf("get max id: %s", err)
	}

	job := &mgo.MapReduce{
		Map: `
		function() {
			emit({ type: "pub", id: this.pubid }, 1)
			emit({ type: "feed", id: this.feedid }, 1)
		}
		`,
		Reduce: `function(key, values) { return Array.sum(values) }`,
		Out:    bson.M{"merge": "article_counts"},
	}
	query := bson.M{
		"_id": bson.M{
			"$gt":  config.LastId,
			"$lte": lastId.Id,
		},
	}
	glog.Info("Starting...")
	info, err := a.Find(query).MapReduce(job, nil)
	if err != nil {
		glog.Fatalf("MapReduce: %s", err)
	}

	config.LastRun = time.Now()
	config.LastId = lastId.Id
	if _, err := c.UpsertId("config", config); err != nil {
		glog.Fatalf("upsert config: %s", err)
	}

	glog.Infof("IN %d; EMIT %d; OUT %d; TOOK %s",
		info.InputCount,
		info.EmitCount,
		info.OutputCount,
		time.Duration(info.Time),
	)
}
