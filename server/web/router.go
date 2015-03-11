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

	var (
		article  = APIHandler{"GET": GetArticle, "PUT": PutArticle, "DELETE": DelArticle}
		articles = APIHandler{"GET": GetArticles, "POST": PostArticle}
		config   = APIHandler{"GET": GetConfig, "PUT": PutConfig, "DELETE": DelConfig}
		configs  = APIHandler{"GET": GetConfigs, "POST": PostConfig}
		feed     = APIHandler{"GET": GetFeed, "PUT": PutFeed, "DELETE": DelFeed}
		feeds    = APIHandler{"GET": GetFeeds, "POST": PostFeed}
		pub      = APIHandler{"GET": GetPub, "PUT": PutPub, "DELETE": DelPub}
		pubs     = APIHandler{"GET": GetPubs, "POST": PostPub}
	)

	api := router.PathPrefix("/api/v1").Subrouter()
	// Config
	api.Handle("/config", configs)
	api.Handle("/config/{key}", config)
	// Pubs
	api.Handle("/pubs", pubs)
	api.Handle("/pubs/{pubid:[a-f0-9]{24}}", pub)
	// Feeds
	api.Handle("/feeds", feeds)
	api.Handle("/feeds/{feedid:[a-f0-9]{24}}", feed)
	api.Handle("/pubs/{pubid:[a-f0-9]{24}}/feeds", feeds)
	api.Handle("/pubs/{pubid:[a-f0-9]{24}}/feeds/{feedid:[a-f0-9]{24}}", feed)
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
