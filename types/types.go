package types

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

type Pub struct {
	Id           bson.ObjectId `bson:"_id"`
	IsLexisNexis bool
	Name         string
	Homepage     string
	Description  string
	NumArticles  int
	NumFeeds     int
	NumReaders   int
	XPathBody    []string
	XPathAuthor  []string
	XPathDate    []string
	XPathTitle   []string
	LastUpdate   time.Time
	LexisNexis   *LexisNexisPub
}

type Feed struct {
	Id           bson.ObjectId `bson:"_id"`
	PubId        bson.ObjectId
	IsLexisNexis bool
	Url          string
	NumArticles  int
	LastDownload time.Time
	NextDownload time.Time
	LexisNexis   *LexisNexisFeed
}

type Article struct {
	Id           bson.ObjectId `bson:"_id"`
	FeedId       bson.ObjectId
	PubId        bson.ObjectId
	IsLexisNexis bool
	Url          string
	Title        string
	Author       string
	Published    time.Time
	BodyText     string
	BodyHTML     string
	HTML         []byte
	LoadTime     time.Duration
	Entry        *EntryInfo
	Goose        *GooseInfo
	LexisNexis   *LexisNexisArticle
}

type EntryInfo struct {
	Url       string
	Title     string
	Author    string
	Published time.Time
}

type GooseInfo struct {
	BodyXPath string
	Title     string
	Published time.Time
	Authors   []string
}
