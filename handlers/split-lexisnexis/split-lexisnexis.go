package main

import (
	"bufio"
	"bytes"
	"flag"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/golang/glog"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Doc struct {
	Id       bson.ObjectId `bson:"_id"`
	Filename string
	XML      []byte
}

const (
	COLLECTION = "lexisnexis"
	TOPIC      = "lexisnexis.id.extract"
)

var (
	dsn      = flag.String("mongo", "mongodb://localhost:27017/ocular8", "Connection string to MongoDB")
	nsqdhttp = flag.String("nsqdhttp", "http://localhost:4151", "NSQd HTTP address")
)

var (
	db     *mgo.Database
	nsqURL *url.URL

	LINEPREFIX = []byte(`<?xml`)
)

func process(filename string) (err error) {
	f, err := os.Open(filename)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)

	var data []byte
	var id bson.ObjectId
	var ids = make([]bson.ObjectId, 0, 1024)
	for scanner.Scan() {
		line := scanner.Bytes()
		if !bytes.HasPrefix(line, LINEPREFIX) {
			data = append(data, line...)
			continue
		}
		if len(data) > 0 {
			if id, err = save(filename, data); err != nil {
				return
			}
			ids = append(ids, id)
		}
		data = line
	}
	if id, err = save(filename, data); err != nil {
		return
	}
	ids = append(ids, id)

	// Generate payload for NSQd
	payload := make([]byte, 0, len(ids)*25)
	for _, id := range ids {
		payload = append(payload, []byte(id.Hex())...)
		payload = append(payload, '\n')
	}
	body := bytes.NewReader(payload)
	bodyType := "multipart/form-data"

	// Send payload to NSQd
	if _, err := http.Post(nsqURL.String(), bodyType, body); err != nil {
		glog.Fatalf("http.Post(%s): %s", nsqURL.String(), err)
	}
	glog.Infof("Sent %d to %s from %s", len(ids), TOPIC, filepath.Base(filename))
	return
}

func save(filename string, data []byte) (id bson.ObjectId, err error) {
	id = bson.NewObjectId()
	doc := &Doc{
		Id:       id,
		Filename: filepath.Base(filename),
		XML:      data,
	}
	if err = db.C(COLLECTION).Insert(doc); err != nil {
		glog.Errorf("db.C(%s).Insert({Id:%s, Filename:%s, XML:%d})", COLLECTION, doc.Id, doc.Filename, len(data))
		return
	}
	return
}

func main() {
	flag.Parse()

	if flag.NArg() == 0 {
		glog.Fatal("No files given")
	}

	var err error
	nsqURL, err = url.Parse(*nsqdhttp)
	if err != nil {
		glog.Fatalf("Error parsing %s: %s", *nsqdhttp, err)
		return
	}
	nsqURL.Path = "/mpub"
	nsqURL.RawQuery = (url.Values{"topic": []string{TOPIC}}).Encode()

	s, err := mgo.Dial(*dsn)
	if err != nil {
		glog.Fatalf("mgo.Dial(%s): %s", *dsn, err)
	}
	defer s.Close()
	db = s.DB("")

	for _, filename := range flag.Args() {
		if err := process(filename); err != nil {
			glog.Errorf("process(%s): %s", filename, err)
		}
	}
}
