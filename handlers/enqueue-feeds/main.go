package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/bitly/go-nsq"
	"github.com/golang/glog"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const TOPIC = "feed.id.download"

var (
	dsn      = flag.String("mongo", "mongodb://localhost:27017/ocular8", "Connection string to MongoDB")
	nsqdHTTP = flag.String("nsqdhttp", "http://localhost:4151", "NSQd HTTP address")
	limit    = flag.Int("limit", 10, "Max number of feeds to enqueue per batch")
)

var (
	nsqdURL *url.URL
)

func checkStats() (err error) {
	statsURL := new(url.URL)
	*statsURL = *nsqdURL
	statsURL.RawQuery = (url.Values{"format": []string{"json"}}).Encode()
	statsURL.Path = "/stats"
	resp, err := http.Get(statsURL.String())
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

	glog.Infof("%+v", stats)

	for _, topic := range stats.Data.Topics {
		switch topic.Name {
		case TOPIC, "entry.id.download":
			// keep processing
		default:
			continue
		}

		if len(topic.Channels)

		if topic.Depth > 0 {
			return fmt.Errorf("%s Topic handlers not active, %d in topic queue", topic.Name, topic.Depth)
		}
	}

	return
}

func main() {
	flag.Parse()

	var err error
	nsqdURL, err = url.Parse(*nsqdHTTP)
	if err != nil {
		glog.Fatalf("Error parsing %s: %s", *nsqdHTTP, err)
		return
	}

	if err := checkStats(); err != nil {
		glog.Fatal(err, "Exiting")
		return
	}
	return

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

	pub := new(url.URL)
	*pub = *nsqdURL
	pub.Path = "/pub"
	pub.RawQuery = (url.Values{"topic": []string{TOPIC}}).Encode()

	payload := make([]byte, 0, len(ids)*25)
	for _, id := range ids {
		glog.Infof("ID: %s", id.Id.Hex())
		payload = append(payload, []byte(id.Id.Hex())...)
		payload = append(payload, '\n')
	}
	body := bytes.NewReader(payload)
	bodyType := "multipart/form-data"

	if _, err := http.Post(pub.String(), bodyType, body); err != nil {
		glog.Fatalf("http.Post(%s): %s", pub.String(), err)
	}
}
