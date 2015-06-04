package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"os"

	"github.com/300brand/ocular8/types"
	"github.com/golang/glog"
	"github.com/mattbaird/elastigo/lib"
)

type PubUpdate struct {
	Genre      string
	Country    string
	Categories []string
}

var useProxy = flag.Bool("proxy", false, "Using proxy (will use localhost instead of 192.x.x.x)")

var conn *elastigo.Conn

var record = []string{"ID", "Pub Name", "Homepage", "Added"}

func main() {
	flag.Parse()
	conn = elastigo.NewConn()
	if !*useProxy {
		conn.SetHosts([]string{"192.168.20.18", "192.168.20.17", "192.168.20.19"})
	}

	pubDsl := elastigo.Search("ocular8")
	pubDsl.Query(elastigo.Query().All())
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
	out := csv.NewWriter(os.Stdout)
	out.Write(record)
	defer out.Flush()
	for {
		if scroll.Hits.Len() == 0 {
			break
		}
		pubCount += scroll.Hits.Len()
		glog.Infof("Processing %d / %d", pubCount, scroll.Hits.Total)
		for _, hit := range scroll.Hits.Hits {
			if hit.Source == nil {
				glog.Fatalf("hit.Source is nil. id: %s", hit.Id)
			}
			p := new(types.Pub)
			if err = json.Unmarshal(*hit.Source, p); err != nil {
				glog.Fatalf("json.Unmarshal(): %s", err)
			}
			record[0] = p.Id.Hex()
			record[1] = p.Name
			record[2] = p.Homepage
			record[3] = p.Added.Format("01/02/2006")
			out.Write(record)
		}
		scroll, err = conn.Scroll(args, scroll.ScrollId)
		if err != nil {
			glog.Fatalf("conn.Scroll(): %s", err)
		}
	}
}
