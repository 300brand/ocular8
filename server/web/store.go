package web

import (
	"labix.org/v2/mgo"
)

var (
	mongo   *mgo.Session
	mongodb *mgo.Database
)

func Mongo(dsn string) (err error) {
	if mongo, err = mgo.Dial(dsn); err != nil {
		return
	}
	mongodb = mongo.DB("")
	return
}

func Close() {
	mongo.Close()
}
