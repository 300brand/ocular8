package web

import (
	"net/http"
	"os"

	"github.com/golang/glog"
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
	glog.Infof("Unused headers: %+v", headers)

	var (
		article  = APIHandler{"GET": GetArticle, "PUT": PutArticle, "DELETE": DelArticle}
		articles = APIHandler{"GET": GetArticles, "POST": PostArticle}
		feed     = APIHandler{"GET": GetFeed, "PUT": PutFeed, "DELETE": DelFeed}
		feeds    = APIHandler{"GET": GetFeeds, "POST": PostFeed}
		pub      = APIHandler{"GET": GetPub, "PUT": PutPub, "DELETE": DelPub}
		pubs     = APIHandler{"GET": GetPubs, "POST": PostPub}
	)

	api := router.PathPrefix("/api/v1").Subrouter()
	// Pubs
	api.Handle("/pubs", pubs)
	api.Handle("/pubs/{pubid:[a-f0-9]{24}}", pub)
	// Feeds
	api.Handle("/feeds", feeds)
	api.Handle("/feeds/{feedid:[a-f0-9]{24}}", feed)
	api.Handle("/pubs/{pubid:[a-f0-9]{24}}/feeds", feeds)
	// Articles
	api.Handle("/articles", articles)
	api.Handle("/feeds/{feedid:[a-f0-9]{24}}/articles", articles)
	api.Handle("/pubs/{pubid:[a-f0-9]{24}}/articles", articles)
	api.Handle("/articles/{articleid:[a-f0-9]{24}}", article)
	// Frontend
	router.PathPrefix("/app/").Handler(http.FileServer(http.Dir(AssetsDir)))
	router.HandleFunc("/", HandleIndex)

	return handlers.LoggingHandler(os.Stdout, router)
}
