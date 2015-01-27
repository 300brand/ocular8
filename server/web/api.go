package web

import (
	"encoding/json"
	"time"

	"github.com/300brand/ocular8/types"
	"gopkg.in/mgo.v2/bson"
)

func GetPubs(ctx *Context) (out interface{}, err error) {
	limit := 20
	pubs := make([]types.Pub, limit)
	query := bson.M{"deleted": bson.M{"$exists": false}}
	err = ctx.DB.C("pubs").Find(query).Sort("name").Limit(limit).All(&pubs)
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
	return pub, err
}

func GetPub(ctx *Context) (out interface{}, err error) {
	id := bson.ObjectIdHex(ctx.Vars["pubid"])
	pub := new(types.Pub)
	err = ctx.DB.C("pubs").FindId(id).One(pub)
	return pub, err
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
	update := bson.M{"$set": bson.M{"deleted": true}}
	err = ctx.DB.C("pubs").UpdateId(bson.ObjectIdHex(ctx.Vars["pubid"]), update)
	return
}

func GetFeeds(ctx *Context) (out interface{}, err error) {
	limit := 20
	feeds := make([]types.Feed, limit)
	query := bson.M{"deleted": bson.M{"$exists": false}}
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
	feed.Id = bson.NewObjectId()
	if pubid, ok := ctx.Vars["pubid"]; ok {
		feed.PubId = bson.ObjectIdHex(pubid)
	}
	if err = ctx.DB.C("feeds").Insert(feed); err != nil {
		return
	}
	err = ctx.DB.C("pubs").UpdateId(feed.PubId, bson.M{"$inc": bson.M{"numfeeds": 1}})
	return feed, err
}

func GetFeed(ctx *Context) (out interface{}, err error) {
	feed := new(types.Feed)
	err = ctx.DB.C("feeds").FindId(bson.ObjectIdHex(ctx.Vars["feedid"])).One(feed)
	return feed, err
}

func PutFeed(ctx *Context) (out interface{}, err error) {
	return
}

func DelFeed(ctx *Context) (out interface{}, err error) {
	update := bson.M{"$set": bson.M{"deleted": true}}
	id := bson.ObjectIdHex(ctx.Vars["feedid"])
	err = ctx.DB.C("feeds").UpdateId(id, update)
	if err != nil {
		return
	}
	pubid := &struct{ PubId bson.ObjectId }{}
	ctx.DB.C("feeds").FindId(id).Select(bson.M{"pubid": true}).One(pubid)
	err = ctx.DB.C("pubs").UpdateId(pubid.PubId, bson.M{"$inc": bson.M{"numfeeds": -1}})
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
