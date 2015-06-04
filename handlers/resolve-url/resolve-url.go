package main

import (
	"errors"
	"flag"
	"net/http"
	"strings"

	"github.com/300brand/ocular8/lib/config"
	"github.com/300brand/ocular8/types"
	"github.com/golang/glog"
	"github.com/mattbaird/elastigo/lib"
	"gopkg.in/mgo.v2/bson"
)

const TargetDomain = "ct.moreover"

func follow(url string) (newurl string, err error) {
	if !strings.Contains(url, TargetDomain) {
		err = errors.New("URL does not contain " + TargetDomain)
		return
	}
	req, err := http.NewRequest("HEAD", url, nil)
	req.Header.Add("User-Agent", "Ocular8 URL-Resolver (http://ocular8.com)")
	if err != nil {
		return
	}
	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return
	}
	loc, err := resp.Location()
	if err != nil && err == http.ErrNoLocation {
		return
	}
	// Getting url.Parse() errors, maybe sidestep them a little
	if loc == nil {
		newurl = resp.Header.Get("Location")
		glog.Warningf("Using raw Location header value: %q", newurl)
		err = nil
		return
	}
	newurl = loc.String()
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
		a := new(types.Article)
		if err := elastic.GetSource(config.ElasticIndex(), "article", id, nil, a); err != nil {
			glog.Errorf("Error fetching article with ID %s: %s", id, err)
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
