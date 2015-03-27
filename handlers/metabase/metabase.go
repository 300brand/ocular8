package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/300brand/ocular8/lib/config"
	"github.com/300brand/ocular8/lib/etcd"
	"github.com/300brand/ocular8/lib/metabase"
	"github.com/300brand/ocular8/types"
	goetcd "github.com/coreos/go-etcd/etcd"
	"github.com/golang/glog"
	"github.com/mattbaird/elastigo/lib"
	"gopkg.in/mgo.v2/bson"
)

var (
	apikey        string
	cacheHit      int
	cacheMiss     int
	canRun        int64
	es            = elastigo.NewConn()
	etcdUrl       = flag.String("etcd", "http://localhost:4001", "Etcd URL")
	parentCache   = make(map[int64][2]bson.ObjectId)
	sequenceId    string
	sequenceReset time.Duration
	store         = flag.String("store", "", "Store a copy of results")
)

func setConfigs() (err error) {
	var reset string
	client := etcd.New(*etcdUrl)
	err = client.GetAll(map[string]*string{
		"/handlers/metabase/apikey":        &apikey,
		"/handlers/metabase/sequenceid":    &sequenceId,
		"/handlers/metabase/sequencereset": &reset,
	})
	if err != nil {
		return
	}
	if sequenceReset, err = time.ParseDuration(reset); err != nil {
		return
	}
	es.SetHosts(config.ElasticHosts())
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
	query := bson.M{
		"query": bson.M{
			"MetabaseId": a.Source.Feed.Id,
		},
	}
	result, err := es.Search(index, "feed", nil, query)
	feed := new(types.Feed)

	feedQ := bson.M{"metabaseid": a.Source.Feed.Id}
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

func saveArticles(r *metabase.Response) (batchId bson.ObjectId, err error) {
	batchId = bson.NewObjectId()
	docs := make([]interface{}, 0, len(r.Articles))
	for i := range r.Articles {
		ra := &r.Articles[i]
		author := ra.Author.Name
		author = strings.TrimPrefix(author, "By ")
		author = strings.TrimPrefix(author, "By ") // Some have it twice..
		a := &types.Article{
			Id:           bson.NewObjectId(),
			BatchId:      batchId,
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
	}
	err = db.C("articles").Insert(docs...)
	return
}

func fromAPI() (response *metabase.Response, err error) {
	response, err = metabase.Fetch(apikey, sequenceId, *store)
	return
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

	var result *metabase.Response
	if result, err = fromAPI(); err != nil {
		glog.Fatalf("fromAPI: %s", err)
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

	batchId, err := saveArticles(result)
	if err != nil {
		glog.Fatalf("saveArticles: %s", err)
	}
	glog.Infof("saveArticles cache hit %d miss %d", cacheHit, cacheMiss)

	body := strings.NewReader(batchId.Hex())
	bodyType := "multipart/form-data"

	nsqURL.Path = "/pub"
	nsqURL.RawQuery = (url.Values{"topic": []string{nsqTopic}}).Encode()
	if _, err := http.Post(nsqURL.String(), bodyType, body); err != nil {
		glog.Fatalf("http.Post(%s): %s", nsqURL.String(), err)
	}
	glog.Infof("Sent %s to %s", batchId.Hex(), nsqURL)

	if _, err = etcd.New(*etcdUrl).Set("/handlers/metabase/lastrun", time.Now().Format(time.RFC3339), 0); err != nil {
		glog.Fatal(err)
	}
}
