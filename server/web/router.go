package web

import (
	"net/http"
	"os"

	"github.com/300brand/ocular8/server/config"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func Handler() http.Handler {
	router := mux.NewRouter()

	api := router.PathPrefix("/api").Subrouter()

	coll := func(path string, get, post http.HandlerFunc) {
		api.Path(path).HandlerFunc(get).Methods("GET")
		api.Path(path).HandlerFunc(post).Headers("Content-Type", "application/json").Methods("POST")
	}

	item := func(path string, get, put, del http.HandlerFunc) {
		api.Path(path).HandlerFunc(get).Methods("GET")
		api.Path(path).HandlerFunc(put).Headers("Content-Type", "application/json").Methods("PUT")
		api.Path(path).HandlerFunc(del).Methods("DELETE")
	}

	// Pubs
	coll("/pubs", GetPubs, PostPub)
	item("/pubs/{pubid:[a-f0-9]{24}}", GetPub, PutPub, DelPub)

	// Feeds
	coll("/feeds", GetFeeds, PostFeed)
	item("/feeds/{feedid:[a-f0-9]{24}}", GetFeed, PutFeed, DelFeed)
	coll("/pubs/{pubid:[a-f0-9]{24}}/feeds", GetFeeds, PostFeed)
	item("/pubs/{pubid:[a-f0-9]{24}}/feeds/{feedid:[a-f0-9]{24}}", GetFeed, PutFeed, DelFeed)

	// Articles
	coll("/articles", GetArticles, PostArticle)
	item("/articles/{articleid:[a-f0-9]{24}}", GetArticle, PutArticle, DelArticle)
	coll("/feeds/{feedid:[a-f0-9]{24}}/articles", GetArticles, PostArticle)
	item("/feeds/{feedid:[a-f0-9]{24}}/articles/{articleid:[a-f0-9]{24}}", GetArticle, PutArticle, DelArticle)
	coll("/pubs/{pubid:[a-f0-9]{24}}/feeds/{feedid:[a-f0-9]{24}}/articles", GetArticles, PostArticle)
	item("/pubs/{pubid:[a-f0-9]{24}}/feeds/{feedid:[a-f0-9]{24}}/articles/{articleid:[a-f0-9]{24}}", GetArticle, PutArticle, DelArticle)

	router.PathPrefix("/app/").Handler(http.FileServer(http.Dir(config.Config.WebAssets)))
	router.HandleFunc("/", HandleIndex)

	return handlers.CombinedLoggingHandler(os.Stdout, router)
}
