package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
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
	index         string
	indexer       *elastigo.BulkIndexer
	sequenceId    string
	sequenceReset time.Duration

	es           = elastigo.NewConn()
	parentCache  = make(map[int64][2]bson.ObjectId)
	store        = flag.String("store", "", "Store a copy of results")
	isLexisNexis = flag.Bool("lexisnexis", false, "Data coming in is from LexisNexis")
)

func name() string {
	return filepath.Base(os.Args[0])
}

func setConfigs() (err error) {
	var reset string
	glog.Infof("config.Etcd(): %s", config.Etcd())
	client := etcd.New(config.Etcd())
	err = client.GetAll(map[string]*string{
		"/handlers/" + name() + "/apikey":        &apikey,
		"/handlers/" + name() + "/sequencereset": &reset,
	})
	if err != nil {
		glog.Errorf("Err: %s", err)
		return
	}
	if sequenceReset, err = time.ParseDuration(reset); err != nil {
		glog.Errorf("Err: %s", err)
		return
	}
	index = config.ElasticIndex()
	es.SetHosts(config.ElasticHosts())
	indexer = es.NewBulkIndexer(4)
	// Check to see if running attribute has expired. If it has, we can
	// continue, otherwise we'll have to exit now and wait
	runKey := "/handlers/" + name() + "/running"
	resp, err := client.Get(runKey, false, false)
	if e, ok := err.(*goetcd.EtcdError); ok {
		if e.ErrorCode == 100 {
			glog.Info("No running key, we can run!")
			err = nil
			_, err = client.Set(runKey, "1", 30)
		}
	} else if err != nil {
		// Not just a "key not found" err..
		glog.Errorf("Err: %s", err)
		return
	} else {
		// Still running, inform how long until we can run again
		canRun = resp.Node.TTL
	}
	// Pull out the sequence ID  separately since it's not managed in
	// config.json
	sidKey := "/handlers/" + name() + "/sequenceid"
	resp, err = client.Get(sidKey, false, false)
	if e, ok := err.(*goetcd.EtcdError); ok {
		if e.ErrorCode == 100 {
			// key not found, not an error, this is just the first time running
			err = nil
		}
	} else if err != nil {
		// Not just a "key not found" err..
		glog.Errorf("Err: %s", err)
		return
	} else {
		sequenceId = resp.Node.Value
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

	// Need to make a new feed and pub
	cacheMiss++
	query := bson.M{
		"size": 1,
		"query": bson.M{
			"term": bson.M{
				"MetabaseId": a.Source.Feed.Id,
			},
		},
	}
	result, err := es.Search(index, "feed", nil, query)
	if err != nil {
		return
	}
	feed := new(types.Feed)
	if result.Hits.Len() == 0 {
		pub := &types.Pub{
			Id:          bson.NewObjectId(),
			Name:        a.Source.Feed.Name,
			Categories:  a.Source.Feed.EditorialTopics,
			Country:     a.Source.Location.Country,
			Description: a.Source.Feed.Description,
			NumFeeds:    1,
			MetabaseId:  a.Source.Feed.Id,
			Added:       time.Now(),
		}
		feed.Id = bson.NewObjectId()
		feed.MetabaseId = a.Source.Feed.Id
		feed.PubId = pub.Id
		feed.Url = fmt.Sprintf("http://ocular8.com/feed/%d.xml", a.Source.Feed.Id)
		feed.Added = time.Now()
		feed.Genre = a.Source.Feed.Genre
		if *isLexisNexis {
			feed.Origin = "lexisnexis"
		} else {
			feed.Origin = "webnews"
		}
		if err = indexer.Index(index, "pub", pub.Id.Hex(), "", &pub.Added, pub, false); err != nil {
			return
		}
		if err = indexer.Index(index, "feed", feed.Id.Hex(), "", &feed.Added, feed, false); err != nil {
			return
		}
	} else {
		if err = json.Unmarshal(*result.Hits.Hits[0].Source, feed); err != nil {
			return
		}
	}
	pubId, feedId = feed.PubId, feed.Id
	if feedId.Hex() == "" {
		glog.Fatal("FeedId is blank")
	}
	if pubId.Hex() == "" {
		glog.Fatal("PubId is blank")
	}
	parentCache[a.Source.Feed.Id] = [2]bson.ObjectId{pubId, feedId}
	return
}

func saveArticles(r *metabase.Response) (batchId bson.ObjectId, err error) {
	batchId = bson.NewObjectId()
	for i := range r.Articles {
		ra := &r.Articles[i]
		author := ra.Author.Name
		author = strings.TrimPrefix(author, "By ")
		author = strings.TrimPrefix(author, "By ") // Some have it twice..

		a := &types.Article{
			Id:           bson.NewObjectId(),
			BatchId:      batchId,
			Url:          ra.OriginalUrl,
			Title:        ra.Title,
			Author:       author,
			Genre:        ra.Source.Feed.Genre,
			Published:    ra.Published(),
			BodyText:     ra.Content,
			BodyHTML:     ra.ContentWithMarkup,
			HTML:         ra.XML(),
			IsLexisNexis: *isLexisNexis,
			HasEmptyBody: len(ra.Content) < 10,
			Added:        time.Now(),
			Metabase: &types.Metabase{
				Author:                author,
				AuthorHomeUrl:         ra.Author.HomeUrl,
				AuthorEmail:           ra.Author.Email,
				SequenceId:            ra.SequenceId,
				Id:                    ra.Id,
				Lni:                   ra.PublishingPlatform.ItemId,
				Url:                   ra.Url,
				OriginalUrl:           ra.OriginalUrl,
				PublishedDate:         ra.PublishedDate,
				HarvestDate:           ra.HarvestDate,
				EmbargoDate:           ra.EmbargoDate,
				LicenseEndDate:        ra.LicenseEndDate,
				ContentLicenseEndDate: ra.ContentLicenseEndDate,
			},
		}
		if a.Url == "" {
			a.Url = a.Metabase.Url
		}
		if a.PubId, a.FeedId, err = parents(ra); err != nil {
			return
		}
		if a.IsLexisNexis {
			a.Origin = "lexisnexis"
		} else {
			a.Origin = "webnews"
		}
		if lni := a.Metabase.Lni; lni != "" {
			a.Url = fmt.Sprintf("http://www.ocular8.com/view/%s", lni)
		}
		if a.Published.IsZero() {
			glog.Errorf(
				"A:%s Zero Published (%q) - MetabaseId:%s SequenceId:%s LNI:%s HarvestDate:%q Title:%q Url:%q",
				a.Id.Hex(),
				ra.PublishedDate,
				ra.Id,
				ra.SequenceId,
				ra.PublishingPlatform.ItemId,
				ra.HarvestDate,
				ra.Title,
				ra.Url,
			)
		}
		now := time.Now()
		if err = indexer.Index(index, "article", a.Id.Hex(), "", &now, a, false); err != nil {
			return
		}
	}
	return
}

func main() {
	var err error

	if err = config.Parse(); err != nil {
		glog.Fatalf("config.Parse: %s", err)
	}

	if err = setConfigs(); err != nil {
		glog.Fatalf("setConfigs(): %s", err)
	}

	if apikey == "" {
		glog.Errorf("API Key undefined. Please provide key in /handlers/" + name() + "/apikey")
		os.Exit(2)
	}

	if canRun > 0 {
		glog.Warningf("Already running, will be able to run in %s", time.Duration(canRun)*time.Second)
		return
	}

	var result *metabase.Response
	if result, err = metabase.Fetch(apikey, sequenceId, *store); err != nil {
		glog.Fatalf("fromAPI: %s", err)
	}

	if len(result.Articles) == 0 {
		glog.Warning("No new articles. Exiting")
		return
	}

	if id := result.NewSequenceId(); id != "" {
		glog.Infof("New SequenceId: %s", id)
		ttl := uint64(sequenceReset.Seconds())
		if _, err := etcd.New(config.Etcd()).Set("/handlers/"+name()+"/sequenceid", id, ttl); err != nil {
			glog.Fatal(err)
		}
	}

	indexer.Start()
	defer indexer.Stop()

	batchId, err := saveArticles(result)
	if err != nil {
		glog.Fatalf("saveArticles: %s", err)
	}
	glog.Infof("saveArticles batch %s cache hit %d miss %d", batchId.Hex(), cacheHit, cacheMiss)

	if _, err = etcd.New(config.Etcd()).Set("/handlers/"+name()+"/lastrun", time.Now().Format(time.RFC3339), 0); err != nil {
		glog.Fatal(err)
	}
}
