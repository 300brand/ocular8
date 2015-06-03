package main

import (
	"encoding/json"
	"sort"
	"strings"
	"time"

	"github.com/300brand/ocular8/lib/config"
	"github.com/300brand/ocular8/types"
	"github.com/golang/glog"
	"github.com/mattbaird/elastigo/lib"
)

type Category struct {
	Category string
	Added    time.Time
}

var categories = make([]string, 0, 255)

func addCategory(cat string) {
	i := sort.SearchStrings(categories, cat)
	if i == len(categories) {
		// Add to end
		categories = append(categories, cat)
		return
	}
	if categories[i] == cat {
		// Already exists
		return
	}
	// Make a hole
	categories = append(categories[:i+1], categories[i:]...)
	// Set new value
	categories[i] = cat
}

func catId(cat string) (id string) {
	return strings.Replace(strings.ToLower(cat), " ", "_", -1)
}

func main() {
	config.Parse()
	conn := elastigo.NewConn()
	conn.SetHosts(config.ElasticHosts())

	pubDsl := elastigo.Search(config.ElasticIndex())
	pubDsl.Filter(elastigo.Filter().Exists("Categories"))
	pubDsl.Scroll("30s")
	pubDsl.Size("100")
	pubDsl.Source(true)
	pubDsl.Type("pub")
	pubResult, err := pubDsl.Result(conn)
	if err != nil {
		glog.Fatalf("pubDsl.Result(): %s", err)
	}
	scroll := *pubResult
	args := map[string]interface{}{
		"search_type": "scan",
		"scroll":      "30s",
	}
	pubCount := 0
	for {
		if scroll.Hits.Len() == 0 {
			break
		}
		pubCount += scroll.Hits.Len()
		glog.Infof("Processing %d / %d Categories: %d", pubCount, scroll.Hits.Total, len(categories))
		for _, hit := range scroll.Hits.Hits {
			if hit.Source == nil {
				glog.Fatalf("hit.Source is nil. id: %s", hit.Id)
			}
			p := new(types.Pub)
			if err = json.Unmarshal(*hit.Source, p); err != nil {
				glog.Fatalf("json.Unmarshal(): %s", err)
			}
			for _, cat := range p.Categories {
				addCategory(cat)
			}
		}
		scroll, err = conn.Scroll(args, scroll.ScrollId)
		if err != nil {
			glog.Fatalf("conn.Scroll(): %s", err)
		}
	}

	for _, cat := range categories {
		id := catId(cat)
		response, err := conn.Get("categories", "unique", id, nil)
		if err != nil && err != elastigo.RecordNotFound {
			glog.Fatalf("conn.Get(%s): %s", id, err)
		} else if err == nil {
			glog.Infof("Exists: [%s]: %q", id, cat)
			continue
		}
		response, err = conn.Index("categories", "unique", id, nil, Category{cat, time.Now()})
		if err != nil {
			glog.Fatalf("conn.Index(%s): %s", id, err)
		}
		glog.Infof("%v [%s]:%q", response.Created, id, cat)
	}
}
