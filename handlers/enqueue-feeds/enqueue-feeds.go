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
	etcdUrl    = flag.String("etcd", "http://localhost:4001", "Etcd URL")
	dsn        string
	nsqURL     *url.URL
	LIMIT      int
	THRESHOLD  int
	TOPIC      string
	FEED_TOPIC string
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
		case TOPIC, FEED_TOPIC:
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
		case chanDepth > THRESHOLD:
			err = fmt.Errorf("%s Channel depth (%d) exceeds threshold (%d)", topic.Name, chanDepth, THRESHOLD)
		}

		if err != nil {
			return
		}
	}

	return
}

func setConfigs() (err error) {
	var nsqhttp, limit, threshold string
	client := etcd.New(*etcdUrl)
	err = client.GetAll(map[string]*string{
		"/config/mongo":                     &dsn,
		"/config/nsqhttp":                   &nsqhttp,
		"/handlers/enqueue-feeds/limit":     &limit,
		"/handlers/enqueue-feeds/threshold": &threshold,
		"/handlers/enqueue-feeds/topic":     &TOPIC,
		"/handlers/download-feed/topic":     &FEED_TOPIC,
	})
	if err != nil {
		return
	}
	if nsqURL, err = url.Parse(nsqhttp); err != nil {
		return
	}
	if LIMIT, err = strconv.Atoi(limit); err != nil {
		return
	}
	if THRESHOLD, err = strconv.Atoi(threshold); err != nil {
		return
	}
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
	}, 0, LIMIT)

	if err = s.DB("").C("feeds").Find(query).Limit(LIMIT).Select(sel).Sort(sort).All(&ids); err != nil {
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
