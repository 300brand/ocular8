package main

import (
	"flag"
	"strings"
	"time"

	"github.com/300brand/ocular8/lib/etcd"
	"github.com/300brand/ocular8/types"
	"github.com/golang/glog"
	elastic "github.com/mattbaird/elastigo/lib"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Pub struct {
	Id   bson.ObjectId `bson:"_id"`
	Name string
}

var (
	machine = flag.String("etcd", "http://localhost:4001", "Etcd address")
	dsn     string
	hosts   string
)

func main() {
	flag.Parse()

	client := etcd.New(*machine)
	err := client.GetAll(map[string]*string{
		"/config/mongo":   &dsn,
		"/config/elastic": &hosts,
	})
	if err != nil {
		glog.Fatalf("Could not get configs: %s", err)
	}

	s, err := mgo.Dial(dsn)
	if err != nil {
		glog.Fatalf("mgo.Dial(%s): %s", dsn, err)
	}
	defer s.Close()
	db := s.DB("")

	conn := elastic.NewConn()
	conn.SetHosts(strings.Split(hosts, ","))

	args := flag.Args()
	if len(args) == 1 && strings.Contains(args[0], " ") {
		args = strings.Split(args[0], " ")
	}

	ids := make([]bson.ObjectId, 0, len(args))
	for _, id := range args {
		if !bson.IsObjectIdHex(id) {
			glog.Errorf("Invalid BSON ObjectId: %s", id)
			continue
		}
		ids = append(ids, bson.ObjectIdHex(id))
	}

	query := bson.M{
		"_id": bson.M{
			"$in": ids,
		},
	}
	elasticArticles := make([]types.ElasticArticle, 0, len(ids))
	err = db.C("articles").Find(query).All(&elasticArticles)
	if err != nil {
		glog.Fatalf("articles.Find(%d): %s", len(ids), err)
	}

	pubIds := make([]bson.ObjectId, 0, len(ids))
	for i := range elasticArticles {
		pubIds = append(pubIds, elasticArticles[i].PublicationId)
	}
	pubs := make([]Pub, 0, len(ids))
	err = db.C("pubs").Find(bson.M{"_id": bson.M{"$in": pubIds}}).Select(bson.M{"name": 1}).All(&pubs)
	if err != nil {
		glog.Fatalf("pubs.Find(%d): %s", len(pubIds), err)
	}

	pubMap := make(map[bson.ObjectId]string)
	for i := range pubs {
		pubMap[pubs[i].Id] = pubs[i].Name
	}

	bi := conn.NewBulkIndexer(5)
	bi.Start()
	for i := range elasticArticles {
		a := &elasticArticles[i]
		var ok bool
		if a.PublicationName, ok = pubMap[a.PublicationId]; !ok {
			glog.Errorf("[%s] No publication found for %s", a.ArticleId.Hex(), a.PublicationId)
			continue
		}
		if a.Published.IsZero() {
			glog.Warningf("[%s] Got zero-published date, setting to Id's timestamp (%s)", a.ArticleId.Hex(), a.ArticleId.Time())
			a.Published = a.ArticleId.Time()
		}
		args := map[string]interface{}{
			"timestamp": a.ArticleId.Time().Format(time.RFC3339),
		}
		if _, err = conn.Index("articles", "article", a.ArticleId.Hex(), args, a); err != nil {
			glog.Errorf("[%s] ElasticSearch.Index: %s", a.ArticleId.Hex(), err)
		}
	}
	bi.Stop()
}
