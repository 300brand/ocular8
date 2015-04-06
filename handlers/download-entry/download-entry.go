package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/300brand/ocular8/lib/config"
	"github.com/300brand/ocular8/lib/etcd"
	"github.com/300brand/ocular8/types"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang/glog"
	"gopkg.in/mgo.v2/bson"
)

type Entry struct {
	ArticleId bson.ObjectId
	FeedId    bson.ObjectId
	PubId     bson.ObjectId
	Link      string
	Author    string
	Title     string
	Published string
}

var (
	TOPIC     string
	SIZELIMIT int
)

var (
	db     *sql.DB
	nsqURL *url.URL
)

func process(entry *Entry) (err error) {
	prefix := fmt.Sprintf("P:%s F:%s A:%s", entry.PubId.Hex(), entry.FeedId.Hex(), entry.ArticleId.Hex())
	start := time.Now()

	clean, resp, err := Clean(entry.Link)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if ct := resp.Header.Get("Content-Type"); invalidContentType(ct) {
		return fmt.Errorf("Invalid Content-Type: %s", ct)
	}

	a := &types.Article{
		Id:     entry.ArticleId,
		FeedId: entry.FeedId,
		PubId:  entry.PubId,
		Url:    clean,
		Author: entry.Author,
		Title:  entry.Title,
		Entry:  new(types.Entry),
	}

	limitReader := io.LimitReader(resp.Body, int64(SIZELIMIT))
	if a.HTML, err = ioutil.ReadAll(limitReader); err != nil {
		return
	}

	if len(a.HTML) == SIZELIMIT {
		return fmt.Errorf("Received more than %d bytes", SIZELIMIT)
	}

	if a.Published, err = time.Parse(time.RFC1123, entry.Published); err != nil {
		glog.Warningf("%s %s", prefix, err)
		if a.Published, err = time.Parse(time.RFC1123Z, entry.Published); err != nil {
			glog.Warningf("%s %s", prefix, err)
		}
		err = nil
	}

	a.Entry.Url = entry.Link
	a.Entry.Title = entry.Title
	a.Entry.Author = entry.Author
	a.Entry.Published = a.Published
	a.LoadTime = time.Since(start)

	glog.Infof("%s Updating data", prefix)

	data, err := json.Marshal(a)
	if err != nil {
		return
	}
	_, err = db.Exec(`UPDATE processing SET queue = ?, data = ? WHERE article_id = ?`, TOPIC, data, a.Id.Hex())
	if err != nil {
		return
	}

	payload := bytes.NewBufferString(entry.ArticleId.Hex())
	if _, err = http.Post(nsqURL.String(), "multipart/form-data", payload); err != nil {
		return
	}
	glog.Infof("%s Sent to %s", prefix, TOPIC)

	return
}

func invalidContentType(ct string) (invalid bool) {
	types := []string{
		"audio/mpeg",
	}
	if i := sort.SearchStrings(types, ct); i < len(types) && types[i] == ct {
		invalid = true
	}
	return
}

func setConfigs() (err error) {
	var sizelimit string
	client := etcd.New(config.Etcd())
	err = client.GetAll(map[string]*string{
		"/handlers/download-entry/sizelimit": &sizelimit,
		"/handlers/extract-goose/consume":    &TOPIC,
	})
	if err != nil {
		return
	}
	if nsqURL, err = url.Parse(config.Nsqhttp()); err != nil {
		return
	}
	nsqURL.Path = "/pub"
	nsqURL.RawQuery = (url.Values{"topic": []string{TOPIC}}).Encode()
	if SIZELIMIT, err = strconv.Atoi(sizelimit); err != nil {
		return
	}
	return
}

func main() {
	var err error
	config.Parse()

	if err = setConfigs(); err != nil {
		glog.Fatalf("setConfigs(): %s", err)
	}

	db, err = sql.Open("mysql", config.MysqlDSN())
	if err != nil {
		return
	}

	ids := make([]interface{}, flag.NArg())
	for i := range ids {
		ids[i] = interface{}(flag.Arg(i))
	}

	glog.Infoln(`SELECT id, data
		FROM processing
		WHERE article_id IN(` + strings.Repeat(",?", flag.NArg())[1:] + `)`)

	rows, err := db.Query(`
		SELECT id, data
		FROM processing
		WHERE article_id IN(`+strings.Repeat(",?", flag.NArg())[1:]+`)`,
		ids...,
	)
	if err != nil {
		glog.Fatalf("db.Query(): %s", err)
	}
	for rows.Next() {
		glog.Infof("Row")
		var id uint64
		data := make([]byte, 0, 16777216)
		if err = rows.Scan(&id, &data); err != nil {
			glog.Errorf("rows.Scan(): %s", err)
			continue
		}
		entry := new(Entry)
		if err = json.Unmarshal(data, entry); err != nil {
			glog.Errorf("json.Unmarshal(): %s", err)
			continue
		}
		if err = process(entry); err != nil {
			glog.Errorf("process(%s): %s", entry.ArticleId, err)
			db.Exec(`INSERT INTO errors
				(article_id, feed_id, pub_id, link, queue, data, started, last_action, reason)
				SELECT article_id, feed_id, pub_id, link, queue, data, started, last_action, ? AS reason
				FROM processing
				WHERE id = ?
				LIMIT 1
			`, err.Error(), id)
		}
	}
}
