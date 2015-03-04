package main

import (
	"encoding/xml"
	"flag"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/300brand/ocular8/lib/etcd"
	"github.com/300brand/ocular8/lib/metabase"
	goetcd "github.com/coreos/go-etcd/etcd"
	"github.com/golang/glog"
	"gopkg.in/mgo.v2"
)

var (
	apikey        string
	canRun        int64
	db            *mgo.Database
	dsn           string
	etcdUrl       = flag.String("etcd", "http://localhost:4001", "Etcd URL")
	lastDownload  time.Time
	nsqURL        *url.URL
	sequenceId    string
	sequenceReset time.Duration
	store         = flag.String("store", "", "Store a copy of results")
	TOPIC         string
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
	TOPIC = configs[2].Value
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

func saveCopy(r *metabase.Response) {
	dir := *store
	if dir == "" {
		return
	}
	f, err := os.Create(filepath.Join(dir, time.Now().Format("20060102T150405.xml")))
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

	if canRun > 0 {
		glog.Warningf("Already running, will be able to run in %s", time.Duration(canRun)*time.Second)
		return
	}
	return

	nsqURL.Path = "/mpub"
	nsqURL.RawQuery = (url.Values{"topic": []string{TOPIC}}).Encode()

	s, err := mgo.Dial(dsn)
	if err != nil {
		glog.Fatalf("mgo.Dial(%s): %s", dsn, err)
	}
	defer s.Close()
	db = s.DB("")

	result, err := metabase.Fetch(apikey, sequenceId)
	if err != nil {
		glog.Fatalf("metabase.Fetch: %s", err)
	}

	saveCopy(result)

	glog.Infof("%#q", result)
}
