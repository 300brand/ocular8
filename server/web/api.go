package web

import (
	"encoding/json"
	"github.com/golang/glog"
	"gopkg.in/mgo.v2"
	"io"
	"net/http"
	"time"

	"github.com/300brand/ocular8/types"
	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2/bson"
)

type APIContext struct {
	Body io.ReadCloser
	DB   *mgo.Database
	Vars map[string]string
}
type APIFuncType func(*APIContext) (interface{}, error)

func APIHandler(f APIFuncType) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		ctx := &APIContext{
			Body: r.Body,
			DB:   mongo.Clone().DB(""),
			Vars: mux.Vars(r),
		}
		defer ctx.DB.Session.Close()

		out, err := f(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			enc.Encode(struct{ Error error }{err})
			return
		}
		switch r.Method {
		case "POST":
			w.WriteHeader(http.StatusCreated)
		}
		if err = enc.Encode(out); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			enc.Encode(struct{ Error error }{err})
		}
	}
}

func GetPubs(ctx *APIContext) (out interface{}, err error) {
	limit := 20
	pubs := make([]types.Pub, limit)
	err = ctx.DB.C("pubs").Find(nil).Sort("name").Limit(limit).All(&pubs)
	return pubs, err
}

func PostPub(ctx *APIContext) (out interface{}, err error) {
	pub := new(types.Pub)
	if err = json.NewDecoder(ctx.Body).Decode(pub); err != nil {
		return
	}
	pub.Id = bson.NewObjectId()
	pub.LastUpdate = time.Now()
	err = ctx.DB.C("pubs").Insert(pub)
	return
}

func GetPub(ctx *APIContext) (out interface{}, err error) {
	id := bson.ObjectIdHex(ctx.Vars["pubid"])
	out = new(types.Pub)
	err = ctx.DB.C("pubs").FindId(id).One(out)
	return
}

func PutPub(ctx *APIContext) (out interface{}, err error) {
	pub := new(types.Pub)
	if err = json.NewDecoder(ctx.Body).Decode(pub); err != nil {
		return
	}
	pub.LastUpdate = time.Now()
	err = ctx.DB.C("pubs").UpdateId(pub.Id, pub)
	return pub, err
}

func DelPub(ctx *APIContext) (out interface{}, err error) {
	return
}

func GetFeeds(ctx *APIContext) (out interface{}, err error) {
	limit := 20
	feeds := make([]types.Feed, limit)
	query := make(map[string]interface{})
	if pubid, ok := ctx.Vars["pubid"]; ok {
		query["pubid"] = bson.ObjectIdHex(pubid)
	}
	err = ctx.DB.C("feeds").Find(query).Sort("url").Limit(limit).All(&feeds)
	return feeds, err
}

func PostFeed(ctx *APIContext) (out interface{}, err error) {
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

func GetFeed(ctx *APIContext) (out interface{}, err error) {
	return
}

func PutFeed(ctx *APIContext) (out interface{}, err error) {
	return
}

func DelFeed(ctx *APIContext) (out interface{}, err error) {
	return
}

func GetArticles(ctx *APIContext) (out interface{}, err error) {
	return
}

func PostArticle(ctx *APIContext) (out interface{}, err error) {
	return
}

func GetArticle(ctx *APIContext) (out interface{}, err error) {
	return
}

func PutArticle(ctx *APIContext) (out interface{}, err error) {
	return
}

func DelArticle(ctx *APIContext) (out interface{}, err error) {
	return
}
