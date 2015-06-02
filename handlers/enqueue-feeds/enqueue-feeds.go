package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/300brand/ocular8/lib/config"
	"github.com/300brand/ocular8/lib/etcd"
	"github.com/golang/glog"
	"github.com/mattbaird/elastigo/lib"
	"gopkg.in/mgo.v2/bson"
)

var (
	dsn       string
	nsqURL    *url.URL
	LIMIT     int
	THRESHOLD int
	TOPIC     string
)

var query = `{
	"_source": false,
	"query": {
		"filtered": {
			"query": {
				"term": {
					"bozo": 0
				}
			},
			"filter": {
				"and": [
					{
						"not": {
							"exists": {
								"field": "MetabaseId"
							}
						}
					},
					{
						"or": [
							{
								"term": {
									"NextDownload": "0001-01-01T00:00:00Z"
								}
							},
							{
								"missing": {
									"field": "NextDownload"
								}
							},
							{
								"range": {
									"NextDownload": {
										"lte": "now"
									}
								}
							}
						]
					}
				]
			}
		}
	},
	"sort": [
		{
			"NextDownload": {
				"missing": "_first",
				"order":   "asc"
			}
		}
	]
}`

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
		case TOPIC:
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
	var limit, threshold string
	client := etcd.New(config.Etcd())
	err = client.GetAll(map[string]*string{
		"/handlers/enqueue-feeds/limit":     &limit,
		"/handlers/enqueue-feeds/threshold": &threshold,
		"/handlers/download-feed/consume":   &TOPIC,
	})
	if err != nil {
		return
	}
	if nsqURL, err = url.Parse(config.Nsqhttp()); err != nil {
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
	config.Parse()

	if err := setConfigs(); err != nil {
		glog.Fatalf("setConfigs: %s", err)
	}

	glog.Infof("About to checkStats()")
	if err := checkStats(); err != nil {
		glog.Warningf("%s. Exiting.", err)
		return
	}

	conn := elastigo.NewConn()
	conn.SetHosts(config.ElasticHosts())
	args := bson.M{
		"size": LIMIT,
		"from": 0,
	}

	result, err := conn.Search(config.ElasticIndex(), "feed", args, query)
	if err != nil {
		glog.Fatalf("ElasticSearch.Search: %s", err)
	}

	ids := make([]bson.ObjectId, 0, LIMIT)
	for _, hit := range result.Hits.Hits {
		ids = append(ids, bson.ObjectIdHex(hit.Id))
	}

	payload := make([]byte, 0, len(ids)*25)
	for _, id := range ids {
		glog.Infof("ID: %s", id.Hex())
		payload = append(payload, []byte(id.Hex())...)
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
