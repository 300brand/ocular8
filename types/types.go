package types

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

type Pub struct {
	Id          bson.ObjectId `bson:"_id" json:"PubId"`
	MetabaseId  int64         `bson:",omitempty" json:",omitempty"`
	Name        string
	Homepage    string
	Description string
	NumArticles int
	NumFeeds    int
	NumReaders  int
	XPathBody   []string `bson:",omitempty" json:",omitempty"`
	XPathAuthor []string `bson:",omitempty" json:",omitempty"`
	XPathDate   []string `bson:",omitempty" json:",omitempty"`
	XPathTitle  []string `bson:",omitempty" json:",omitempty"`
	LastUpdate  time.Time
	Added       time.Time
}

type Feed struct {
	Id           bson.ObjectId `bson:"_id" json:"FeedId"`
	PubId        bson.ObjectId
	MetabaseId   int64 `bson:",omitempty" json:",omitempty"`
	Url          string
	NumArticles  int
	Added        time.Time
	LastDownload time.Time `json:",omitempty"`
	NextDownload time.Time `json:",omitempty"`
}

type Article struct {
	Id           bson.ObjectId `bson:"_id" json:"ArticleId"`
	FeedId       bson.ObjectId
	PubId        bson.ObjectId
	BatchId      bson.ObjectId `bson:",omitempty" json:",omitempty"`
	Url          string
	Title        string
	Author       string
	Published    time.Time
	BodyText     string
	BodyHTML     string
	HTML         []byte
	LoadTime     time.Duration
	IsLexisNexis bool
	Added        time.Time
	Entry        *Entry    `bson:",omitempty" json:",omitempty"`
	Goose        *Goose    `bson:",omitempty" json:",omitempty"`
	Metabase     *Metabase `bson:",omitempty" json:",omitempty"`
	XPath        *XPath    `bson:",omitempty" json:",omitempty"`
}

type Entry struct {
	Url       string
	Title     string
	Author    string
	Published *time.Time
}

type Goose struct {
	BodyXPath string
	Title     string
	Published *time.Time
	Authors   []string
}

type XPath struct {
	Title     string
	Published *time.Time
	Author    string
}

type Metabase struct {
	Author        string
	AuthorHomeUrl string
	AuthorEmail   string
	SequenceId    string
	Id            string
}
