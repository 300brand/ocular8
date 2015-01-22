package web

import (
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

var AssetsDir string

func Handler(assetsDir string) http.Handler {
	AssetsDir = assetsDir

	router := mux.NewRouter()

	headers := []string{
		"Content-Type", "application/json",
		"Content-Type", "application/json;charset=UTF-8",
	}
	api := router.PathPrefix("/api").Subrouter()

	coll := func(path string, get, post http.HandlerFunc) {
		api.Path(path).HandlerFunc(get).Methods("GET")
		api.Path(path).HandlerFunc(post).Headers(headers...).Methods("POST")
	}

	item := func(path string, get, put, del http.HandlerFunc) {
		api.Path(path).HandlerFunc(get).Methods("GET")
		api.Path(path).HandlerFunc(put).Headers(headers...).Methods("PUT")
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

	router.PathPrefix("/app/").Handler(http.FileServer(http.Dir(AssetsDir)))
	router.HandleFunc("/", HandleIndex)

	return handlers.CombinedLoggingHandler(os.Stdout, router)
}
