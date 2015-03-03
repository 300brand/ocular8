package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/300brand/ocular8/lib/etcd"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"time"

	"github.com/300brand/ocular8/types"
	"github.com/golang/glog"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Entry struct {
	Id        bson.ObjectId `bson:"_id"`
	FeedId    bson.ObjectId
	PubId     bson.ObjectId
	Link      string
	Author    string
	Title     string
	Published string
}

var (
	etcdUrl   = flag.String("etcd", "http://localhost:4001", "Etcd URL")
	dsn       string
	TOPIC     string
	SIZELIMIT int
)

var (
	db     *mgo.Database
	nsqURL *url.URL
)

func process(entry *Entry) (err error) {
	prefix := fmt.Sprintf("P:%s F:%s E:%s", entry.PubId.Hex(), entry.FeedId.Hex(), entry.Id.Hex())
	start := time.Now()

	clean, resp, err := Clean(entry.Link)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if ct := resp.Header.Get("Content-Type"); invalidContentType(ct) {
		return fmt.Errorf("Invalid Content-Type: %s", ct)
	}

	a := &types.Article{
		Id:     entry.Id,
		FeedId: entry.FeedId,
		PubId:  entry.PubId,
		Url:    clean,
		Author: entry.Author,
		Title:  entry.Title,
	}

	limitReader := io.LimitReader(resp.Body, int64(SIZELIMIT))
	if a.HTML, err = ioutil.ReadAll(limitReader); err != nil {
		return
	}

	if len(a.HTML) == SIZELIMIT {
		return fmt.Errorf("Received more than %d bytes", SIZELIMIT)
	}

	if a.Published, err = time.Parse(time.RFC1123, entry.Published); err != nil {
		glog.Warningf("%s %s", prefix, err)
		if a.Published, err = time.Parse(time.RFC1123Z, entry.Published); err != nil {
			glog.Warningf("%s %s", prefix, err)
		}
		err = nil
	}

	a.Entry.Url = entry.Link
	a.Entry.Title = entry.Title
	a.Entry.Author = entry.Author
	a.Entry.Published = a.Published
	a.LoadTime = time.Since(start)

	glog.Infof("%s Moving to articles collection", prefix)
	if err = db.C("articles").Insert(a); err != nil {
		return
	}
	if err = db.C("entries").RemoveId(entry.Id); err != nil {
		return
	}

	payload := bytes.NewBufferString(entry.Id.Hex())
	if _, err = http.Post(nsqURL.String(), "multipart/form-data", payload); err != nil {
		return
	}
	glog.Infof("%s Sent to %s", prefix, TOPIC)

	return
}

func invalidContentType(ct string) (invalid bool) {
	types := []string{
		"audio/mpeg",
	}
	if i := sort.SearchStrings(types, ct); i < len(types) && types[i] == ct {
		invalid = true
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
			Key:     "/handlers/download-entry/sizelimit",
			Default: "2097152",
			Desc:    "Size limit for downloaded entries",
		},
		&etcd.Item{
			Key:     "/handlers/download-entry/topic",
			Default: "article.id.extract.goose",
			Desc:    "Topic to post article IDs to",
		},
	}
	if err = client.GetAll(configs); err != nil {
		return
	}
	dsn = configs[0].Value
	if nsqURL, err = url.Parse(configs[1].Value); err != nil {
		return
	}
	if SIZELIMIT, err = strconv.Atoi(configs[2].Value); err != nil {
		return
	}
	TOPIC = configs[3].Value
	return
}

func main() {
	flag.Parse()

	if err := setConfigs(); err != nil {
		glog.Fatalf("setConfigs(): %s", err)
	}

	nsqURL.Path = "/pub"
	nsqURL.RawQuery = (url.Values{"topic": []string{TOPIC}}).Encode()

	s, err := mgo.Dial(dsn)
	if err != nil {
		glog.Fatalf("mgo.Dial(%s): %s", dsn, err)
	}
	defer s.Close()
	db = s.DB("")

	change := mgo.Change{
		Update:    bson.M{"$inc": bson.M{"attempts": 1}},
		ReturnNew: false,
	}
	for _, id := range flag.Args() {
		if !bson.IsObjectIdHex(id) {
			glog.Fatalf("Invalid BSON ObjectId: %s", id)
		}

		q := bson.M{"_id": bson.ObjectIdHex(id)}
		entry := new(Entry)
		if _, err := db.C("entries").Find(q).Apply(change, entry); err != nil {
			glog.Errorf("Find(%+v): %s", q, err)
			continue
		}
		if err := process(entry); err != nil {
			glog.Errorf("process(%s): %s", entry.Id.Hex(), err)
			// TODO switch the error type
			db.C("article_errors").Insert(bson.M{
				"url":    entry.Link,
				"type":   fmt.Sprintf("%T", err),
				"code":   0,
				"reason": err.Error(),
				"entry":  entry,
			})
		}
	}
}
