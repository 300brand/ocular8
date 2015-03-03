package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/300brand/ocular8/lib/etcd"
	"github.com/golang/glog"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	etcdUrl     = flag.String("etcd", "http://localhost:4001", "Etcd URL")
	dsn         string
	nsqURL      *url.URL
	limit       int
	threshold   int
	TOPIC       string
	ENTRY_TOPIC string
)

var (
	nsqURLURL *url.URL
)

func checkStats() (err error) {
	nsqURL.RawQuery = (url.Values{"format": []string{"json"}}).Encode()
	nsqURL.Path = "/stats"
	resp, err := http.Get(nsqURL.String())
	if err != nil {
		return
	}
	defer resp.Body.Close()

	stats := &struct {
		Data struct {
			Topics []struct {
				Name     string `json:"topic_name"`
				Depth    int
				Channels []struct {
					Depth int
				}
			}
		}
	}{}
	if err = json.NewDecoder(resp.Body).Decode(stats); err != nil {
		return
	}

	for _, topic := range stats.Data.Topics {
		switch topic.Name {
		case TOPIC, ENTRY_TOPIC:
			// keep processing
		default:
			continue
		}

		chanDepth := 0
		for _, c := range topic.Channels {
			chanDepth += c.Depth
		}

		switch {
		case len(topic.Channels) == 0:
			err = fmt.Errorf("%s No channels to handle topic", topic.Name)
		case topic.Depth > 0:
			err = fmt.Errorf("%s Topic handlers not active, %d in topic queue", topic.Name, topic.Depth)
		case chanDepth > threshold:
			err = fmt.Errorf("%s Channel depth (%d) exceeds threshold (%d)", topic.Name, chanDepth, threshold)
		}

		if err != nil {
			return
		}
	}

	return
}

func setConfigs() (err error) {
	client := etcd.New(*etcdUrl)
	configs := []*etcd.Item{
		&etcd.Item{
			Key:     "/config/mongo/dsn",
			Default: "mongodb://localhost:27017/ocular8",
			Desc:    "Connection string to MongoDB",
		},
		&etcd.Item{
			Key:     "/config/nsq/http",
			Default: "http://localhost:4151",
			Desc:    "NSQd HTTP address",
		},
		&etcd.Item{
			Key:     "/handlers/enqueue-feeds/limit",
			Default: "10",
			Desc:    "Max number of feeds to enqueue per batch",
		},
		&etcd.Item{
			Key:     "/handlers/enqueue-feeds/threshold",
			Default: "100",
			Desc:    "Entry threshold to avoid pushing more feeds into download queue. Applies to both feed and entry downloads.",
		},
		&etcd.Item{
			Key:     "/handlers/enqueue-feeds/topic",
			Default: "feed.id.download",
			Desc:    "Topic to post feed IDs to",
		},
		&etcd.Item{
			Key:     "/handlers/download-feed/topic",
			Default: "entry.id.download",
			Desc:    "Topic to post entry IDs to",
		},
	}
	if err = client.GetAll(configs); err != nil {
		return
	}
	dsn = configs[0].Value
	if nsqURL, err = url.Parse(configs[1].Value); err != nil {
		return
	}
	if limit, err = strconv.Atoi(configs[2].Value); err != nil {
		return
	}
	if threshold, err = strconv.Atoi(configs[3].Value); err != nil {
		return
	}
	TOPIC = configs[4].Value
	ENTRY_TOPIC = configs[5].Value
	return
}

func main() {
	flag.Parse()

	if err := setConfigs(); err != nil {
		glog.Fatalf("setConfigs: %s", err)
	}

	glog.Infof("About to checkStats()")
	if err := checkStats(); err != nil {
		glog.Warningf("%s. Exiting.", err)
		return
	}

	glog.Infof("About to mgo.Dial")
	s, err := mgo.Dial(dsn)
	if err != nil {
		glog.Fatalf("mgo.Dial(%s): %s", dsn, err)
	}
	defer s.Close()

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
	}, 0, limit)

	if err = s.DB("").C("feeds").Find(query).Limit(limit).Select(sel).Sort(sort).All(&ids); err != nil {
		glog.Fatalf("mgo.Find: %s", err)
	}

	payload := make([]byte, 0, len(ids)*25)
	for _, id := range ids {
		glog.Infof("ID: %s", id.Id.Hex())
		payload = append(payload, []byte(id.Id.Hex())...)
		payload = append(payload, '\n')
	}
	body := bytes.NewReader(payload)
	bodyType := "multipart/form-data"

	nsqURL.Path = "/mpub"
	nsqURL.RawQuery = (url.Values{"topic": []string{TOPIC}}).Encode()
	if _, err := http.Post(nsqURL.String(), bodyType, body); err != nil {
		glog.Fatalf("http.Post(%s): %s", nsqURL.String(), err)
	}
	glog.Infof("Sent %d Feed IDs to %s", len(ids), nsqURL)
}
