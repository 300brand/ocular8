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
		if err = row.Scan(&processing_id, &data); err != nil {
			glog.Fatalf("row.Scan(): %s", err)
		}
		glog.Infof("data: %s", data)
		if err = json.Unmarshal(data, &article); err != nil {
			glog.Fatalf("json.Unmarshal(): %s", err)
		}
		glog.Infof("article: %+v", article)
		if _, err = conn.Index(config.ElasticIndex(), "article", id, nil, &article); err != nil {
			glog.Fatalf("elasticsearch.Index(): %s", err)
		}
	}
}
