package web

import (
	"encoding/json"
	"gopkg.in/mgo.v2"
	"io"
	"time"

	"github.com/300brand/ocular8/types"
	"github.com/golang/glog"
	"gopkg.in/mgo.v2/bson"
)

type Context struct {
	Body io.ReadCloser
	DB   *mgo.Database
	Vars map[string]string
}

func GetPubs(ctx *Context) (out interface{}, err error) {
	limit := 20
	pubs := make([]types.Pub, limit)
	err = ctx.DB.C("pubs").Find(nil).Sort("name").Limit(limit).All(&pubs)
	return pubs, err
}

func PostPub(ctx *Context) (out interface{}, err error) {
	pub := new(types.Pub)
	if err = json.NewDecoder(ctx.Body).Decode(pub); err != nil {
		return
	}
	pub.Id = bson.NewObjectId()
	pub.LastUpdate = time.Now()
	err = ctx.DB.C("pubs").Insert(pub)
	return
}

func GetPub(ctx *Context) (out interface{}, err error) {
	id := bson.ObjectIdHex(ctx.Vars["pubid"])
	out = new(types.Pub)
	err = ctx.DB.C("pubs").FindId(id).One(out)
	return
}

func PutPub(ctx *Context) (out interface{}, err error) {
	pub := new(types.Pub)
	if err = json.NewDecoder(ctx.Body).Decode(pub); err != nil {
		return
	}
	pub.LastUpdate = time.Now()
	err = ctx.DB.C("pubs").UpdateId(pub.Id, pub)
	return pub, err
}

func DelPub(ctx *Context) (out interface{}, err error) {
	return
}

func GetFeeds(ctx *Context) (out interface{}, err error) {
	limit := 20
	feeds := make([]types.Feed, limit)
	query := make(map[string]interface{})
	if pubid, ok := ctx.Vars["pubid"]; ok {
		query["pubid"] = bson.ObjectIdHex(pubid)
	}
	err = ctx.DB.C("feeds").Find(query).Sort("url").Limit(limit).All(&feeds)
	return feeds, err
}

func PostFeed(ctx *Context) (out interface{}, err error) {
	feed := new(types.Feed)
	if err = json.NewDecoder(ctx.Body).Decode(feed); err != nil {
		return
	}
	if pubid, ok := ctx.Vars["pubid"]; ok {
		feed.PubId = bson.ObjectIdHex(pubid)
	}
	glog.Infof("%+v", feed)
	return feed, nil
}

func GetFeed(ctx *Context) (out interface{}, err error) {
	return
}

func PutFeed(ctx *Context) (out interface{}, err error) {
	return
}

func DelFeed(ctx *Context) (out interface{}, err error) {
	return
}

func GetArticles(ctx *Context) (out interface{}, err error) {
	return
}

func PostArticle(ctx *Context) (out interface{}, err error) {
	return
}

func GetArticle(ctx *Context) (out interface{}, err error) {
	return
}

func PutArticle(ctx *Context) (out interface{}, err error) {
	return
}

func DelArticle(ctx *Context) (out interface{}, err error) {
	return
}
