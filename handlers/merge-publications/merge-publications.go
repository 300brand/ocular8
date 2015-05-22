package main

import (
	"encoding/json"
	"errors"
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

var (
	bulk  *elastigo.BulkIndexer
	conn  *elastigo.Conn
	index string
)

func allPubs() (pubs []*Pub, err error) {
	dsl := elastigo.Search(index)
	dsl.Type("pub")
	dsl.Scroll("30s")
	dsl.Source(true)
	dsl.Size("1000")
	result, err := dsl.Result(conn)
	if err != nil {
		return
	}
	pubs = make([]*Pub, 0, 50000)
	scroll := *result
	args := map[string]interface{}{
		"scroll": "30s",
	}
	for {
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
		scroll, err = conn.Scroll(args, scroll.ScrollId)
		if err != nil {
			return
		}
	}
	return
}

func deletePub(pub *Pub) {
	bulk.Delete(index, "pub", pub.Id, false)
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
			dupes = append(dupes, pubs[k:i])
		}
		k = i
		lastName = pubs[i].Name
	}
	return
}

func merge(pubs []*Pub) (err error) {
	if len(pubs) < 2 {
		return errors.New("Need 2 or more pubs to merge")
	}
	parent := pubs[0]
	children := pubs[1:]
	for _, child := range children {
		if err = updateFeeds(child.Id, parent.Id); err != nil {
			return
		}
		if err = updateArticles(child.Id, parent.Id); err != nil {
			return
		}
		deletePub(child)
	}
	return
}

func updateArticles(oldId, newId string) (err error) {
	q := elastigo.Query()
	q.Term("PubId", oldId)

	dsl := elastigo.Search(index)
	dsl.Type("article")
	dsl.Query(q)
	dsl.Source(false)
	dsl.Size("10")
	dsl.Scroll("30s")
	result, err := dsl.Result(conn)
	if err != nil {
		return
	}
	glog.Infof("updateArticles(%q, %q): Total: %d", oldId, newId, result.Hits.Total)
	scroll := *result
	args := map[string]interface{}{
		"scroll": "30s",
	}
	update := &struct{ PubId string }{newId}
	for {
		if scroll.Hits.Len() == 0 {
			break
		}
		glog.Infof("Processing %d / %d", scroll.Hits.Len(), scroll.Hits.Total)
		for _, hit := range scroll.Hits.Hits {
			glog.Infof("[%s] Updating %q -> %q", hit.Id, oldId, newId)
			err = bulk.UpdateWithPartialDoc(hit.Index, hit.Type, hit.Id, "", nil, update, false, false)
			if err != nil {
				return
			}
		}
		scroll, err = conn.Scroll(args, scroll.ScrollId)
		if err != nil {
			return
		}
	}
	return
}

func updateFeeds(oldId, newId string) (err error) {
	q := elastigo.Query()
	q.Term("PubId", oldId)

	dsl := elastigo.Search(index)
	dsl.Type("feed")
	dsl.Query(q)
	dsl.Source(false)
	result, err := dsl.Result(conn)
	if err != nil {
		return
	}
	glog.Infof("updateFeeds(%q, %q): Total: %d", oldId, newId, result.Hits.Total)
	update := &struct{ PubId string }{newId}
	for _, hit := range result.Hits.Hits {
		glog.Infof("[%s] [%s] [%s] Updating %q -> %q", hit.Index, hit.Type, hit.Id, oldId, newId)
		err = bulk.UpdateWithPartialDoc(hit.Index, hit.Type, hit.Id, "", nil, update, false, false)
		if err != nil {
			return
		}
	}
	return
}

func main() {
	config.Parse()
	conn = elastigo.NewConn()
	conn.SetHosts(config.ElasticHosts())
	index = config.ElasticIndex()
	bulk = conn.NewBulkIndexer(2)

	pubs, err := allPubs()
	if err != nil {
		glog.Fatalf("allPubs(): %s", err)
	}
	sort.Sort(Pubs(pubs))
	dupes := findDupes(pubs)
	glog.Infof("Found %d dupes", len(dupes))
	bulk.Start()
	for _, dupe := range dupes {
		if err = merge(dupe); err != nil {
			glog.Fatalf("merge(%q): %s", dupe, err)
		}
	}
	bulk.Stop()
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
