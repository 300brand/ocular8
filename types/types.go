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
	XPathBody   []string `bson:",omitempty"`
	XPathAuthor []string `bson:",omitempty"`
	XPathDate   []string `bson:",omitempty"`
	XPathTitle  []string `bson:",omitempty"`
	LastUpdate  time.Time
}

type Feed struct {
	Id           bson.ObjectId `bson:"_id"`
	PubId        bson.ObjectId
	MetabaseId   int64 `bson:",omitempty"`
	Url          string
	NumArticles  int
	LastDownload time.Time
	NextDownload time.Time
	Ignore       bool
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
	Entry     *Entry    `bson:",omitempty"`
	Goose     *Goose    `bson:",omitempty"`
	Metabase  *Metabase `bson:",omitempty"`
}

type Entry struct {
	Url       string
	Title     string
	Author    string
	Published time.Time
}

type Goose struct {
	BodyXPath string
	Title     string
	Published time.Time
	Authors   []string
}

type Metabase struct {
	Author        string
	AuthorHomeUrl string
	AuthorEmail   string
	SequenceId    string
	Id            string
}
