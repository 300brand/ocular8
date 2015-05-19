package main

import (
	"encoding/json"
	"github.com/300brand/ocular8/lib/config"
	"github.com/golang/glog"
	"github.com/mattbaird/elastigo/lib"
	"sort"
)

type Pub struct {
	Id   string
	Name string
}

type Pubs []*Pub

var _ sort.Interface = make(Pubs, 0)

func allPubs(conn *elastigo.Conn) (pubs []*Pub, err error) {
	dsl := elastigo.Search(config.ElasticIndex())
	dsl.Type("pub")
	dsl.Scroll("30s")
	dsl.Source(true)
	dsl.Size("1000")
	result, err := dsl.Result(conn)
	if err != nil {
		return
	}
	pubs = make([]*Pub, 0, 50000)
	scroll := elastigo.SearchResult{
		ScrollId: result.ScrollId,
	}
	args := map[string]interface{}{
		"scroll": "30s",
	}
	for {
		scroll, err = conn.Scroll(args, scroll.ScrollId)
		if err != nil {
			return
		}
		if scroll.Hits.Len() == 0 {
			break
		}
		glog.Infof("Processing %d / %d", scroll.Hits.Len(), scroll.Hits.Total)
		for _, hit := range scroll.Hits.Hits {
			if hit.Source == nil {
				glog.Fatalf("hit.Source is nil. id: %s len(pubs) = %d", hit.Id, len(pubs))
			}
			p := new(Pub)
			if err = json.Unmarshal(*hit.Source, p); err != nil {
				return
			}
			p.Id = hit.Id
			pubs = append(pubs, p)
		}
	}
	return
}

// must be sorted!
func findDupes(pubs []*Pub) (dupes [][]*Pub) {
	dupes = make([][]*Pub, 0, 1000)
	var lastName = ""
	var k = 0
	for i := range pubs {
		if pubs[i].Name == lastName {
			continue
		}
		if pubs[i].Name != lastName && i-k > 1 {
			glog.Infof("Found grouping: [%d:%d] = %q", k, i, pubs[k:i])
			dupes = append(dupes, pubs[k:i])
		}
		k = i
		lastName = pubs[i].Name
	}
	return
}

func main() {
	config.Parse()
	elastic := elastigo.NewConn()
	elastic.SetHosts(config.ElasticHosts())

	pubs, err := allPubs(elastic)
	if err != nil {
		glog.Fatalf("allPubs(): %s", err)
	}
	sort.Sort(Pubs(pubs))
	_ = findDupes(pubs)
}

func (p Pubs) Len() int {
	return len(p)
}

func (p Pubs) Less(i, j int) bool {
	if p[i].Name == p[j].Name {
		return p[i].Id < p[j].Id
	}
	return p[i].Name < p[j].Name
}

func (p Pubs) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}
