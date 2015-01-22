package web

import (
	"gopkg.in/mgo.v2"
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
