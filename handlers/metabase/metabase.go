package main

import (
	"bytes"
	"encoding/xml"
	"flag"
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/300brand/ocular8/lib/etcd"
	"github.com/300brand/ocular8/lib/metabase"
	"github.com/300brand/ocular8/types"
	goetcd "github.com/coreos/go-etcd/etcd"
	"github.com/golang/glog"
	"gopkg.in/mgo.v2"
)

var (
	apikey        string
	cacheHit      int
	cacheMiss     int
	canRun        int64
	db            *mgo.Database
	dsn           string
	etcdUrl       = flag.String("etcd", "http://localhost:4001", "Etcd URL")
	nsqTopic      string
	nsqURL        *url.URL
	parentCache   = make(map[int64][2]bson.ObjectId)
	sequenceId    string
	sequenceReset time.Duration
	store         = flag.String("store", "", "Store a copy of results")
)

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
			Key:     "/handlers/metabase/topic",
			Default: "article.id.elastic",
			Desc:    "Topic to post article IDs to",
		},
		&etcd.Item{
			Key:     "/handlers/metabase/apikey",
			Default: "",
			Desc:    "Metabase API Key",
		},
		&etcd.Item{
			Key:     "/handlers/metabase/sequenceid",
			Default: "",
			Desc:    "Metabase Sequence ID - Do not touch unless things go pear-shaped,",
		},
		&etcd.Item{
			Key:     "/handlers/metabase/sequencereset",
			Default: "48h",
			Desc:    "How long to wait before cutting losses and resetting sequenceId. Used in the event of power/network loss",
		},
	}
	if err = client.GetAll(configs); err != nil {
		return
	}
	dsn = configs[0].Value
	if nsqURL, err = url.Parse(configs[1].Value); err != nil {
		return
	}
	nsqTopic = configs[2].Value
	apikey = configs[3].Value
	sequenceId = configs[4].Value
	if sequenceReset, err = time.ParseDuration(configs[5].Value); err != nil {
		return
	}
	// Check to see if running attribute has expired. If it has, we can
	// continue, otherwise we'll have to exit now and wait
	runKey := "/handlers/metabase/running"
	resp, err := client.Client.Get(runKey, false, false)
	if e, ok := err.(*goetcd.EtcdError); ok {
		if e.ErrorCode == 100 {
			glog.Info("No running key, we can run!")
			err = nil
			_, err = client.Client.Set(runKey, "1", 30)
		}
	} else if err != nil {
		// Not just a "key not found" err..
		return
	} else {
		// Still running, inform how long until we can run again
		canRun = resp.Node.TTL
	}
	return
}

func parents(a *metabase.Article) (pubId, feedId bson.ObjectId, err error) {
	idSet, ok := parentCache[a.Source.Feed.Id]
	if ok {
		cacheHit++
		pubId, feedId = idSet[0], idSet[1]
		return
	}

	cacheMiss++
	feed := new(types.Feed)
	feedQ := bson.M{"metabaseid": a.Source.Feed.Id}
	feedSel := bson.M{"pubid": 1}
	err = db.C("feeds").Find(feedQ).Select(feedSel).One(feed)
	if err == mgo.ErrNotFound {
		pub := &types.Pub{
			Id:          bson.NewObjectId(),
			Name:        a.Source.Feed.Name,
			Description: a.Source.Feed.Description,
			NumFeeds:    1,
			NeedsReview: true,
		}
		feed.Id = bson.NewObjectId()
		feed.MetabaseId = a.Source.Feed.Id
		feed.PubId = pub.Id
		feed.Ignore = true
		feed.Url = fmt.Sprintf("http://ocular8.com/feed/%d.xml", a.Source.Feed.Id)
		if err = db.C("feeds").Insert(feed); err != nil {
			return
		}
		if err = db.C("pubs").Insert(pub); err != nil {
			return
		}
	}
	if err != nil {
		return
	}
	pubId, feedId = feed.PubId, feed.Id
	parentCache[a.Source.Feed.Id] = [2]bson.ObjectId{pubId, feedId}
	return
}

func saveArticles(r *metabase.Response) (ids []bson.ObjectId, err error) {
	ids = make([]bson.ObjectId, 0, len(r.Articles))
	docs := make([]interface{}, 0, len(r.Articles))
	for i := range r.Articles {
		ra := &r.Articles[i]
		author := ra.Author.Name
		author = strings.TrimPrefix(author, "By ")
		author = strings.TrimPrefix(author, "By ") // Some have it twice..
		a := &types.Article{
			Id:        bson.NewObjectId(),
			Url:       ra.Url,
			Title:     ra.Title,
			Author:    author,
			Published: ra.Published(),
			BodyText:  ra.Content,
			BodyHTML:  ra.ContentWithMarkup,
			HTML:      ra.XML(),
			Metabase: &types.Metabase{
				Author:        author,
				AuthorHomeUrl: ra.Author.HomeUrl,
				AuthorEmail:   ra.Author.Email,
				SequenceId:    ra.SequenceId,
				Id:            ra.Id,
			},
		}
		if a.PubId, a.FeedId, err = parents(ra); err != nil {
			return
		}
		docs = append(docs, a)
		ids = append(ids, a.Id)
	}
	err = db.C("articles").Insert(docs...)
	return
}

func saveCopy(r *metabase.Response, dir string) {
	f, err := os.Create(filepath.Join(dir, time.Now().Format("20060102T150405.encoded.xml")))
	if err != nil {
		glog.Error(err)
		return
	}
	defer f.Close()
	enc := xml.NewEncoder(f)
	enc.Indent("", "\t")
	if err := enc.Encode(r); err != nil {
		glog.Errorf("xml.Encode: %s", err)
		return
	}
}

func main() {
	flag.Parse()

	if err := setConfigs(); err != nil {
		glog.Fatalf("setConfigs(): %s", err)
	}

	if apikey == "" {
		glog.Errorf("API Key undefined. Please provide key in /handlers/metabase/apikey")
		os.Exit(2)
	}

	if canRun > 0 {
		glog.Warningf("Already running, will be able to run in %s", time.Duration(canRun)*time.Second)
		return
	}

	s, err := mgo.Dial(dsn)
	if err != nil {
		glog.Fatalf("mgo.Dial(%s): %s", dsn, err)
	}
	defer s.Close()
	db = s.DB("")

	result, err := metabase.Fetch(apikey, sequenceId, *store)
	if err != nil {
		glog.Fatalf("metabase.Fetch: %s", err)
	}

	if dir := *store; dir != "" {
		saveCopy(result, dir)
	}

	if len(result.Articles) == 0 {
		glog.Warning("No new articles. Exiting")
		return
	}

	if id := result.NewSequenceId(); id != "" {
		glog.Infof("New SequenceId: %s", id)
		ttl := uint64(sequenceReset.Seconds())
		if _, err := etcd.New(*etcdUrl).Set("/handlers/metabase/sequenceid", id, ttl); err != nil {
			glog.Fatal(err)
		}
	}

	ids, err := saveArticles(result)
	if err != nil {
		glog.Fatalf("saveArticles: %s", err)
	}
	glog.Infof("saveArticles cache hit %d miss %d", cacheHit, cacheMiss)

	payload := make([]byte, 0, len(ids)*25)
	for _, id := range ids {
		payload = append(payload, []byte(id.Hex())...)
		payload = append(payload, '\n')
	}
	body := bytes.NewReader(payload)
	bodyType := "multipart/form-data"

	nsqURL.Path = "/mpub"
	nsqURL.RawQuery = (url.Values{"topic": []string{nsqTopic}}).Encode()
	if _, err := http.Post(nsqURL.String(), bodyType, body); err != nil {
		glog.Fatalf("http.Post(%s): %s", nsqURL.String(), err)
	}
	glog.Infof("Sent %d Article IDs to %s", len(ids), nsqURL)

	if _, err = etcd.New(*etcdUrl).Set("/handlers/metabase/lastrun", time.Now().Format(time.RFC3339), 0); err != nil {
		glog.Fatal(err)
	}
}
