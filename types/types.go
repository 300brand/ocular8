package types

import (
	"labix.org/v2/mgo/bson"
	"time"
)

type Pub struct {
	Id          bson.ObjectId `bson:"_id"`
	Name        string
	Homepage    string
	Description string
	Readership  int
	NumFeeds    int
	NumArticles int
	LastUpdate  time.Time
}

type Feed struct {
	Id           bson.ObjectId `bson:"_id"`
	PubId        bson.ObjectId
	Url          string
	LastDownload time.Time
}
