package main

import (
	"encoding/json"
	"encoding/xml"
	"flag"
	"github.com/300brand/ocular8/lib/metabase"
	"os"

	"github.com/300brand/ocular8/types"
	"github.com/golang/glog"
	"github.com/mattbaird/elastigo/lib"
)

var conn *elastigo.Conn

func main() {
	flag.Parse()
	conn = elastigo.NewConn()
	conn.SetHosts([]string{"192.168.20.18", "192.168.20.17", "192.168.20.19"})

	pubDsl := elastigo.Search("ocular8")
	pubDsl.Filter(elastigo.Filter().Missing("Categories"))
	pubDsl.Scroll("30s")
	pubDsl.Size("100")
	pubDsl.Source(true)
	pubDsl.Type("pub")
	json.NewEncoder(os.Stdout).Encode(pubDsl)

	pubResult, err := pubDsl.Result(conn)
	if err != nil {
		glog.Fatalf("pubDsl.Result(): %s", err)
	}
	pubs := make([]*types.Pub, 0, 50000)
	scroll := *pubResult
	args := map[string]interface{}{
		"search_type": "scan",
		"scroll":      "30s",
	}
	for {
		if scroll.Hits.Len() == 0 {
			break
		}
		glog.Infof("Processing %d / %d", len(pubs)+scroll.Hits.Len(), scroll.Hits.Total)
		for _, hit := range scroll.Hits.Hits {
			if hit.Source == nil {
				glog.Fatalf("hit.Source is nil. id: %s len(pubs) = %d", hit.Id, len(pubs))
			}
			p := new(types.Pub)
			if err = json.Unmarshal(*hit.Source, p); err != nil {
				glog.Fatalf("json.Unmarshal(): %s", err)
			}
			pubs = append(pubs, p)
		}
		scroll, err = conn.Scroll(args, scroll.ScrollId)
		if err != nil {
			glog.Fatalf("conn.Scroll(): %s", err)
		}
	}

	articleDsl := elastigo.Search("ocular8")
	articleDsl.Filter(elastigo.Filter().Terms("Origin", "webnews", "lexisnexis"))
	articleDsl.Size("1")
	articleDsl.Source(true)
	articleDsl.Sort(elastigo.Sort("ArticleId").Desc())
	articleDsl.Type("article")
	for _, pub := range pubs {
		articleDsl.Query(elastigo.Query().Term("PubId", pub.Id.Hex()))
		json.NewEncoder(os.Stdout).Encode(articleDsl)

		result, err := articleDsl.Result(conn)
		if err != nil {
			glog.Fatalf("articleDsl.Result(): %s", err)
		}
		if result.Hits.Total < 1 {
			glog.Warningf("No articles found")
			continue
		}
		article := new(types.Article)
		if err = json.Unmarshal(*result.Hits.Hits[0].Source, article); err != nil {
			glog.Fatalf("json.Unmarshal(): %s", err)
		}
		metabaseArticle := new(metabase.Article)
		if err = xml.Unmarshal(article.HTML, metabaseArticle); err != nil {
			glog.Fatalf("xml.Unmarshal(): %s", err)
		}

		glog.Infof("%q", metabaseArticle.Source.Feed.EditorialTopics)

	}
}
