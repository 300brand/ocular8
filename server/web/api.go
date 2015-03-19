package web

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/300brand/ocular8/lib/config"
	"github.com/300brand/ocular8/lib/etcd"
	"github.com/300brand/ocular8/types"
	"github.com/golang/glog"
	"gopkg.in/mgo.v2/bson"
)

func GetConfigs(ctx *Context) (out interface{}, err error) {
	c := etcd.New(config.Etcd())
	return c.GetList()
}

func PostConfig(ctx *Context) (out interface{}, err error) {
	glog.Error("Can not post new configs")
	return
}

func GetConfig(ctx *Context) (out interface{}, err error) {
	c := etcd.New(config.Etcd())
	return c.Get(ctx.Vars["key"], false, false)
}

func PutConfig(ctx *Context) (out interface{}, err error) {
	c := etcd.New(config.Etcd())

	// value used to retain compatibility with etcd
	value := ctx.Values.Get("value")

	// AngularJS $update on an Item
	item := new(etcd.Item)
	if err := json.NewDecoder(ctx.Body).Decode(item); err == nil {
		value = item.Value
	}

	return c.Set(ctx.Vars["key"], value, 0)
}

func DelConfig(ctx *Context) (out interface{}, err error) {
	glog.Error("Can not delete configs")
	return
}

func GetPubs(ctx *Context) (out interface{}, err error) {
	query := bson.M{"deleted": bson.M{"$exists": false}}
	if v := ctx.Values.Get("query"); v != "" {
		if bson.IsObjectIdHex(v) {
			query["_id"] = bson.ObjectIdHex(v)
		} else {
			query["name"] = bson.M{
				"$regex":   v,
				"$options": "i",
			}
		}
	}

	sort := "name"
	if v := ctx.Values.Get("sort"); v != "" {
		sort = v
	}

	limit := 20
	if v := ctx.Values.Get("limit"); v != "" {
		if i, _ := strconv.Atoi(v); i > 0 && i <= 1e4 {
			limit = i
		}
	}

	offset := 0
	if v := ctx.Values.Get("offset"); v != "" {
		if i, _ := strconv.Atoi(v); i > 0 {
			offset = i
		}
	}

	// Fetch total
	total, err := ctx.DB.C("pubs").Find(query).Count()
	if err != nil {
		return
	}

	// Fetch actual subset of pubs
	pubs := make([]types.Pub, limit)
	glog.Infof("SORT: %s LIMIT: %d OFFSET: %d QUERY: %+v", sort, limit, offset, query)
	err = ctx.DB.C("pubs").Find(query).Sort(sort).Limit(limit).Skip(offset).All(&pubs)
	if err != nil {
		return
	}

	// Set headers for totals and links
	headers := ctx.W.Header()
	pages := int(math.Ceil(float64(total) / float64(limit)))
	if pages > 1 {
		links := make([]string, 0, 14)
		v := ctx.Values
		f := fmt.Sprintf(`<%s?%%s>; rel="%%s"; title="%%s"`, ctx.R.URL.Path)
		// First / Prev if able
		if offset > 0 {
			// First
			v.Set("offset", "0")
			links = append(links, fmt.Sprintf(f, v.Encode(), "first", "First"))
			// Previous
			v.Set("offset", fmt.Sprint(offset-limit))
			links = append(links, fmt.Sprintf(f, v.Encode(), "prev", "Previous"))
		}
		// Next / Last if able
		if o := offset + limit; o < total {
			// Next
			v.Set("offset", fmt.Sprint(o))
			links = append(links, fmt.Sprintf(f, v.Encode(), "next", "Next"))
			// Last
			v.Set("offset", fmt.Sprint((pages-1)*limit))
			links = append(links, fmt.Sprintf(f, v.Encode(), "last", "Last"))
		}

		headers.Add("Link", strings.Join(links, ", "))
	}

	headers.Add("X-Total-Count", fmt.Sprint(total))
	headers.Add("X-Total-Pages", fmt.Sprint(pages))

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
