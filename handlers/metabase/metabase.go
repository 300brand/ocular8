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
	backfill      = flag.String("backfill", "", "Backfill saved XML")
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
	var nsqhttp, reset string
	client := etcd.New(*etcdUrl)
	err = client.GetAll(map[string]*string{
		"/config/mongo":                    &dsn,
		"/config/nsqhttp":                  &nsqhttp,
		"/handlers/metabase/topic":         &nsqTopic,
		"/handlers/metabase/apikey":        &apikey,
		"/handlers/metabase/sequenceid":    &sequenceId,
		"/handlers/metabase/sequencereset": &reset,
	})
	if err != nil {
		return
	}
	if nsqURL, err = url.Parse(nsqhttp); err != nil {
		return
	}
	if sequenceReset, err = time.ParseDuration(reset); err != nil {
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
			Id:           bson.NewObjectId(),
			Url:          ra.Url,
			Title:        ra.Title,
			Author:       author,
			Published:    ra.Published(),
			BodyText:     ra.Content,
			BodyHTML:     ra.ContentWithMarkup,
			HTML:         ra.XML(),
			IsLexisNexis: true,
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

func fromAPI() (response *metabase.Response, err error) {
	response, err = metabase.Fetch(apikey, sequenceId, *store)
	return
}

func fromFile(filename string) (response *metabase.Response, err error) {
	f, err := os.Open(filename)
	if err != nil {
		return
	}
	defer f.Close()
	response = new(metabase.Response)
	err = xml.NewDecoder(f).Decode(response)
	return
}

func main() {
	flag.Parse()

	if err := setConfigs(); err != nil {
		glog.Fatalf("setConfigs(): %s", err)
	}

	s, err := mgo.Dial(dsn)
	if err != nil {
		glog.Fatalf("mgo.Dial(%s): %s", dsn, err)
	}
	defer s.Close()
	db = s.DB("")

	var result *metabase.Response
	if filename := *backfill; filename != "" {
		if result, err = fromFile(filename); err != nil {
			glog.Fatalf("fromFile(%s): %s", filename, err)
		}
	} else {
		if apikey == "" {
			glog.Errorf("API Key undefined. Please provide key in /handlers/metabase/apikey")
			os.Exit(2)
		}

		if canRun > 0 {
			glog.Warningf("Already running, will be able to run in %s", time.Duration(canRun)*time.Second)
			return
		}
		if result, err = fromAPI(); err != nil {
			glog.Fatalf("fromAPI: %s", err)
		}
	}

	if len(result.Articles) == 0 {
		glog.Warning("No new articles. Exiting")
		return
	}

	if id := result.NewSequenceId(); *backfill == "" && id != "" {
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

	// This payload is different from others to fascilitate bulk elastic inserts
	payload := make([]byte, 0, len(ids)*25)
	for _, id := range ids {
		payload = append(payload, []byte(id.Hex())...)
		payload = append(payload, ' ')
	}
	payload[len(payload)-1] = '\n'
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
