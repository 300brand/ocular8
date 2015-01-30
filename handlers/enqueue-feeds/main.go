package main

import (
	"bytes"
	"flag"
	"net/http"
	"time"

	"github.com/bitly/go-nsq"
	"github.com/golang/glog"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const TOPIC = "feed.id.download"

var (
	dsn     = flag.String("mongo", "mongodb://localhost:27017/ocular8", "Connection string to MongoDB")
	nsqAddr = flag.String("nsqd", "localhost:4150", "NSQd address")
	limit   = flag.Int("limit", 10, "Max number of feeds to enqueue per batch")
	useHTTP = flag.String("usehttp", "", "Use this HTTP address instead of TCP")
)

func main() {
	flag.Parse()

	s, err := mgo.Dial(*dsn)
	if err != nil {
		glog.Fatalf("mgo.Dial(%s): %s", *dsn, err)
	}
	defer s.Close()

	config := nsq.NewConfig()
	config.Set("client_id", "enqueue-feeds")
	config.Set("user_agent", "enqueue-feeds/v1.0")

	query := bson.M{
		"$or": []bson.M{
			bson.M{
				"nextdownload": bson.M{
					"$lt": time.Now(),
				},
			},
			bson.M{
				"nextdownload": bson.M{
					"$exists": false,
				},
			},
		},
	}
	sel := bson.M{"_id": true}
	sort := "lastdownload"
	ids := make([]struct {
		Id bson.ObjectId `bson:"_id"`
	}, 0, *limit)

	if err = s.DB("").C("feeds").Find(query).Limit(*limit).Select(sel).Sort(sort).All(&ids); err != nil {
		glog.Fatalf("mgo.Find: %s", err)
	}

	payload := make([][]byte, len(ids))

	for i, id := range ids {
		glog.Infof("ID: %s", id.Id.Hex())
		payload[i] = []byte(id.Id.Hex())
	}

	if *useHTTP != "" {
		body := make([]byte, 0, *limit*25)
		for _, p := range payload {
			body = append(body, p...)
			body = append(body, '\n')
		}
		http.Post(*useHTTP+"/mpub?topic="+TOPIC, "application/json;charset=UTF-8", bytes.NewReader(body))
		return
	}

	producer, err := nsq.NewProducer(*nsqAddr, config)
	if err != nil {
		glog.Fatalf("nsq.NewProducer(%s, config): %s", *nsqAddr, err)
	}
	defer producer.Stop()
	if err = producer.MultiPublish("feed.id.download", payload); err != nil {
		glog.Fatalf("producer.MultiPublish: %s", err)
	}
}
