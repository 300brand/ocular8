package web

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/300brand/ocular8/types"
	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2/bson"
)

func GetPubs(w http.ResponseWriter, r *http.Request) {
	limit := 10
	pubs := make([]types.Pub, limit)
	err := mongodb.C("pubs").Find(nil).Sort("name").Limit(limit).All(&pubs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(pubs); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func PostPub(w http.ResponseWriter, r *http.Request) {
	pub := new(types.Pub)
	if err := json.NewDecoder(r.Body).Decode(pub); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	pub.Id = bson.NewObjectId()
	pub.LastUpdate = time.Now()
	if err := mongodb.C("pubs").Insert(pub); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(pub); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func GetPub(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := bson.ObjectIdHex(vars["pubid"])
	pub := new(types.Pub)
	if err := mongodb.C("pubs").FindId(id).One(pub); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(pub); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func PutPub(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("PutPub\n"))
}

func DelPub(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("DeletePub\n"))
}

func GetFeeds(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("GetFeeds\n"))
}

func PostFeed(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("PostFeed\n"))
}

func GetFeed(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("GetFeed\n"))
}

func PutFeed(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("PutFeed\n"))
}

func DelFeed(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("DelFeed\n"))
}

func GetArticles(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("GetArticles\n"))
}

func PostArticle(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("PostArticle\n"))
}

func GetArticle(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("GetArticle\n"))
}

func PutArticle(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("PutArticle\n"))
}

func DelArticle(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("DelArticle\n"))
}
