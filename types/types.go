package types

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

type Pub struct {
	Id          bson.ObjectId `bson:"_id"`
	Name        string
	Homepage    string
	Description string
	NumArticles int
	NumFeeds    int
	NumReaders  int
	XPathBody   []string
	XPathAuthor []string
	XPathDate   []string
	XPathTitle  []string
	LastUpdate  time.Time
}

type Feed struct {
	Id           bson.ObjectId `bson:"_id"`
	PubId        bson.ObjectId
	Url          string
	NumArticles  int
	LastDownload time.Time
	NextDownload time.Time
}

type Article struct {
	Id        bson.ObjectId `bson:"_id"`
	FeedId    bson.ObjectId
	PubId     bson.ObjectId
	Url       string
	Title     string
	Author    string
	Published time.Time
	BodyText  string
	BodyHTML  string
	HTML      []byte
	LoadTime  time.Duration
	Entry     struct {
		Url       string
		Title     string
		Author    string
		Published time.Time
	}
	Goose struct {
		BodyXPath string
		Title     string
		Published time.Time
		Authors   []string
	}
}
