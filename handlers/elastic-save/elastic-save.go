package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"strings"

	"github.com/300brand/ocular8/lib/config"
	"github.com/300brand/ocular8/types"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang/glog"
	elastic "github.com/mattbaird/elastigo/lib"
	"gopkg.in/mgo.v2/bson"
)

type M map[string]interface{}

func urlExists(conn *elastic.Conn, a *types.Article) (exists bool) {
	query := M{
		"_source": false,
		"query": M{
			"bool": M{
				"must": []M{
					M{"term": M{"Url": a.Url}},
				},
				"must_not": []M{
					// This bit allows re-saving articles with the same ID
					M{"term": M{"ArticleId": a.Id}},
				},
			},
		},
	}
	result, err := conn.Search("ocular8", "article", nil, query)
	if err != nil {
		glog.Errorf("urlExists(%q): %s", a.Url, err)
		return
	}
	exists = result.Hits.Total > 0
	return
}

func main() {
	config.Parse()

	db, err := sql.Open("mysql", config.MysqlDSN())
	if err != nil {
		glog.Fatalf("sql.Open(): %s", err)
	}
	defer db.Close()

	conn := elastic.NewConn()
	conn.SetHosts(config.ElasticHosts())

	// Sometimes a space separated list of IDs gets quoted
	args := flag.Args()
	if len(args) == 1 && strings.Contains(args[0], " ") {
		args = strings.Split(args[0], " ")
	}

	// Filter out any IDs that are not BSON IDs
	ids := make([]string, 0, len(args))
	for _, id := range args {
		if !bson.IsObjectIdHex(id) {
			glog.Errorf("Invalid BSON ObjectId: %s", id)
			continue
		}
		ids = append(ids, id)
	}

	var processing_id uint64
	var data = make([]byte, 0, 16777216)
	var article types.Article
	for _, id := range ids {
		row := db.QueryRow(`SELECT id, data FROM processing WHERE article_id = ?`, id)
		err = row.Scan(&processing_id, &data)
		if err == sql.ErrNoRows {
			glog.Warningf("No processing record found for %s", id)
			continue
		}
		if err != nil {
			glog.Fatalf("%s - row.Scan(): %s", id, err)
		}
		if err = json.Unmarshal(data, &article); err != nil {
			db.Exec(`INSERT INTO errors
				(article_id, feed_id, pub_id, link, queue, data, started, last_action, reason)
				SELECT article_id, feed_id, pub_id, link, queue, data, started, last_action, ? AS reason
				FROM processing
				WHERE id = ?
				LIMIT 1`, "json.Unmarshal(): "+err.Error(), processing_id)
			db.Exec(`DELETE FROM processing WHERE id = ? LIMIT 1`, processing_id)
			glog.Errorf("%s - json.Unmarshal(): %s", id, err)
			continue
		}
		// Final checks before shoving into database
		if urlExists(conn, &article) {
			glog.Warningf("%s - Article URL already exists [%s]", id, article.Url)
			continue
		}
		if article.Published.IsZero() {
			glog.Warningf("%s - Published is a zero-date. Setting to %s", id, article.Id.Time())
			article.Published = article.Id.Time()
		}
		article.HasEmptyBody = len(article.BodyText) == 0
		// Send article to ElasticSearch
		if _, err = conn.Index(config.ElasticIndex(), "article", id, nil, &article); err != nil {
			glog.Fatalf("%s - elasticsearch.Index(): %s", id, err)
		}
		_, err = db.Exec(`DELETE FROM processing WHERE id = ? LIMIT 1`, processing_id)
		if err != nil {
			glog.Fatalf("%s - DELETE: %s", id, err)
		}
	}
}
