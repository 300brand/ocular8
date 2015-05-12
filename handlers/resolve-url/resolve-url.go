package main

import (
	"encoding/json"
	"errors"
	"flag"
	"github.com/300brand/ocular8/lib/config"
	"github.com/300brand/ocular8/types"
	"github.com/golang/glog"
	"github.com/mattbaird/elastigo/lib"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"strings"
)

const TargetDomain = "ct.moreover"

func follow(url string) (newurl string, err error) {
	if !strings.Contains(url, TargetDomain) {
		err = errors.New("URL does not contain " + TargetDomain)
		return
	}
	resp, err := http.Head(url)
	if err != nil {
		return
	}
	if newurl = resp.Request.URL.String(); newurl == "" {
		err = errors.New("No new URL found for " + url)
		return
	}
	return
}

func main() {
	config.Parse()
	elastic := elastigo.NewConn()
	elastic.SetHosts(config.ElasticHosts())
	for _, id := range flag.Args() {
		if !bson.IsObjectIdHex(id) {
			glog.Errorf("Invalid BSON ObjectId: %s", id)
			continue
		}
		response, err := elastic.Get(config.ElasticIndex(), "article", id, nil)
		if err != nil {
			glog.Errorf("Error fetching article with ID %s: %s", id, err)
			continue
		}
		a := new(types.Article)
		if err = json.Unmarshal(*response.Source, a); err != nil {
			glog.Errorf("[%s] json.Unmarshal: %s", id, err)
			continue
		}
		newurl, err := follow(a.Metabase.Url)
		if err != nil {
			glog.Errorf("[%s] follow(): %s", id, err)
			continue
		}
		glog.Infof("%s %s", id, newurl)
		doc := struct{ Url string }{newurl}
		if _, err := elastic.UpdateWithPartialDoc(config.ElasticIndex(), "article", id, nil, doc, false); err != nil {
			glog.Errorf("[%s] Update: %s", id, err)
			continue
		}
	}
}
