package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"flag"
	"net/http"
	"time"

	"github.com/300brand/ocular8/types"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang/glog"
	"github.com/mattbaird/elastigo/lib"
	"gopkg.in/mgo.v2/bson"
)

var (
	doPrime  = flag.Bool("prime", false, "Prime database with data from existing databases")
	primeRPC = flag.String("primerpc", "http://okcodev:52204/rpc", "Prime RPC address")
)

func prime(elasticHosts []string, index, mysqldsn string) (err error) {
	db, err := sql.Open("mysql", mysqldsn)
	if err != nil {
		return
	}
	defer db.Close()

	creates := []string{
		`CREATE TABLE IF NOT EXISTS processing (
			id          SERIAL,
			            PRIMARY KEY (id),
			article_id  CHAR(24) NOT NULL,
			            UNIQUE (article_id),
			feed_id     CHAR(24) NOT NULL,
			            INDEX (feed_id),
			pub_id      CHAR(24) NOT NULL,
			            INDEX (pub_id),
			link        VARCHAR(255),
			            UNIQUE(link),
			queue       VARCHAR(64),
			data        MEDIUMBLOB,
			started     DATETIME(6),
			last_action TIMESTAMP(6)
		)`,
		`CREATE TABLE IF NOT EXISTS errors (
			id          SERIAL,
			            PRIMARY KEY (id),
			article_id  CHAR(24) NOT NULL,
			            INDEX (article_id),
			feed_id     CHAR(24) NOT NULL,
			            INDEX (feed_id),
			pub_id      CHAR(24) NOT NULL,
			            INDEX (pub_id),
			link        VARCHAR(255),
			queue       VARCHAR(64),
			data        MEDIUMBLOB,
			started     DATETIME(6),
			last_action DATETIME(6),
			added       TIMESTAMP(6) DEFAULT CURRENT_TIMESTAMP,
			reason      TEXT
		)`,
	}
	for _, query := range creates {
		if _, err = db.Exec(query); err != nil {
			return
		}
	}

	conn := elastigo.NewConn()
	conn.SetHosts(elasticHosts)

	settings := struct {
		Aliases  bson.M `json:"aliases"`
		Mappings bson.M `json:"mappings"`
	}{
		Aliases: bson.M{
			"nextdownload": bson.M{
				"filter": bson.M{
					"and": []bson.M{
						bson.M{
							"not": bson.M{
								"exists": bson.M{
									"field": "MetabaseId",
								},
							},
						},
						bson.M{
							"or": []bson.M{
								bson.M{
									"term": bson.M{
										"NextDownload": "0001-01-01T00:00:00Z",
									},
								},
								bson.M{
									"range": bson.M{
										"NextDownload": bson.M{
											"gte": "now",
										},
									},
								},
							},
						},
					},
				},
			},
		},
		Mappings: bson.M{
			"pub": bson.M{
				"properties": bson.M{
					"PubId":       bson.M{"type": "string", "index": "not_analyzed"},
					"MetabaseId":  bson.M{"type": "long"},
					"Name":        bson.M{"type": "string", "index": "analyzed"},
					"Categories":  bson.M{"type": "string", "index": "not_analyzed"},
					"Country":     bson.M{"type": "string", "index": "not_analyzed"},
					"Homepage":    bson.M{"type": "string", "index": "no"},
					"Description": bson.M{"type": "string", "index": "analyzed"},
					"NumArticles": bson.M{"type": "integer"},
					"NumFeeds":    bson.M{"type": "integer"},
					"NumReaders":  bson.M{"type": "integer"},
					"XPathBody":   bson.M{"type": "string", "index": "no"},
					"XPathAuthor": bson.M{"type": "string", "index": "no"},
					"XPathDate":   bson.M{"type": "string", "index": "no"},
					"XPathTitle":  bson.M{"type": "string", "index": "no"},
					"LastUpdate":  bson.M{"type": "date"},
					"Added":       bson.M{"type": "date"},
					"IsNew":       bson.M{"type": "boolean"},
				},
			},
			"feed": bson.M{
				"properties": bson.M{
					"FeedId":       bson.M{"type": "string", "index": "not_analyzed"},
					"PubId":        bson.M{"type": "string", "index": "not_analyzed"},
					"MetabaseId":   bson.M{"type": "long"},
					"Url":          bson.M{"type": "string", "index": "not_analyzed"},
					"NumArticles":  bson.M{"type": "integer"},
					"Added":        bson.M{"type": "date"},
					"LastDownload": bson.M{"type": "date"},
					"NextDownload": bson.M{"type": "date"},
				},
			},
			"article": bson.M{
				"properties": bson.M{
					"ArticleId":    bson.M{"type": "string", "index": "not_analyzed"},
					"FeedId":       bson.M{"type": "string", "index": "not_analyzed"},
					"PubId":        bson.M{"type": "string", "index": "not_analyzed"},
					"BatchId":      bson.M{"type": "string", "index": "not_analyzed"},
					"Url":          bson.M{"type": "string", "index": "not_analyzed"},
					"Title":        bson.M{"type": "string", "index": "analyzed"},
					"Author":       bson.M{"type": "string", "index": "analyzed"},
					"Published":    bson.M{"type": "date"},
					"BodyText":     bson.M{"type": "string", "index": "analyzed"},
					"BodyHTML":     bson.M{"type": "string", "index": "no"},
					"HTML":         bson.M{"type": "string", "index": "no"},
					"LoadTime":     bson.M{"type": "long"},
					"IsLexisNexis": bson.M{"type": "boolean"},
					"HasEmptyBody": bson.M{"type": "boolean"},
					"Added":        bson.M{"type": "date"},
					"Entry":        bson.M{"type": "object", "enabled": false},
					"Goose":        bson.M{"type": "object", "enabled": false},
					"Metabase":     bson.M{"type": "object", "enabled": false},
					"XPath":        bson.M{"type": "object", "enabled": false},
				},
			},
		},
	}

	if _, err = conn.CreateIndexWithSettings(index, settings); err != nil {
		return
	}

	// Temporary structs to hold original data
	type P struct {
		Id          bson.ObjectId `json:"ID"`
		Title       string
		Url         string `json:"URL"`
		NumArticles int
		NumFeeds    int
		NumReaders  int
		Updated     time.Time
		XPaths      struct{ Author, Body, Date, Title []string }
	}
	type F struct {
		Id      bson.ObjectId `json:"ID"`
		Url     string        `json:"URL"`
		Updated time.Time
	}

	pubquery := bson.M{
		"id":     1,
		"method": "Publication.GetAll",
		"params": []bson.M{},
	}
	b, err := json.Marshal(pubquery)
	if err != nil {
		return
	}
	glog.Infof("Publication.GetAll from %s", *primeRPC)
	resp, err := http.Post(*primeRPC, "application/json", bytes.NewReader(b))
	if err != nil {
		return
	}
	defer resp.Body.Close()

	result := &struct {
		Result struct {
			Publications []P
		}
	}{}
	if err = json.NewDecoder(resp.Body).Decode(result); err != nil {
		return
	}

	pubFeeds := func(id bson.ObjectId, indexer *elastigo.BulkIndexer) (numfeeds int, err error) {
		feedquery := bson.M{
			"id":     id.Hex(),
			"method": "Publication.View",
			"params": []bson.M{
				bson.M{
					"Publication": id,
					"Feeds": bson.M{
						"Select": bson.M{
							"content": 0,
							"urls":    0,
							"log":     0,
						},
					},
					"Articles": bson.M{
						"Limit": 1,
						"Select": bson.M{
							"_id": 1,
						},
					},
				},
			},
		}

		b, err := json.Marshal(feedquery)
		if err != nil {
			return
		}
		resp, err := http.Post(*primeRPC, "application/json", bytes.NewReader(b))
		if err != nil {
			return
		}
		defer resp.Body.Close()

		result := &struct {
			Result struct {
				Feeds struct {
					Feeds []F
				}
			}
		}{}
		if err = json.NewDecoder(resp.Body).Decode(result); err != nil {
			return
		}

		for _, f := range result.Result.Feeds.Feeds {
			glog.Infof("FEED %s", f.Id.Hex())
			feed := &types.Feed{
				Id:    f.Id,
				PubId: id,
				Url:   f.Url,
			}
			t := f.Id.Time()
			if indexer.Index(index, "feed", f.Id.Hex(), "", &t, feed, false); err != nil {
				return
			}
		}
		numfeeds = len(result.Result.Feeds.Feeds)
		return
	}

	indexer := conn.NewBulkIndexer(10)
	indexer.Start()
	defer indexer.Stop()

	for _, p := range result.Result.Publications {
		glog.Infof("PUB  %s", p.Id.Hex())
		numfeeds := 0
		if numfeeds, err = pubFeeds(p.Id, indexer); err != nil {
			return
		}
		pub := &types.Pub{
			Id:          p.Id,
			Name:        p.Title,
			Homepage:    p.Url,
			NumReaders:  p.NumReaders,
			NumFeeds:    numfeeds,
			LastUpdate:  p.Updated,
			XPathAuthor: p.XPaths.Author,
			XPathBody:   p.XPaths.Body,
			XPathDate:   p.XPaths.Date,
			XPathTitle:  p.XPaths.Title,
		}
		t := p.Id.Time()
		if err = indexer.Index(index, "pub", p.Id.Hex(), "", &t, pub, false); err != nil {
			return
		}
	}

	return
}
