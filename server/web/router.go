package web

import (
	"net/http"
	"os"
	"strings"

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

	apiMap := map[string]APIFuncType{
		"GET    /pubs":                                                  GetPubs,
		"POST   /pubs":                                                  PostPub,
		"GET    /pubs/{pubid:[a-f0-9]{24}}":                             GetPub,
		"PUT    /pubs/{pubid:[a-f0-9]{24}}":                             PutPub,
		"DELETE /pubs/{pubid:[a-f0-9]{24}}":                             DelPub,
		"GET    /feeds":                                                 GetFeeds,
		"POST   /feeds":                                                 PostFeed,
		"GET    /feeds/{feedid:[a-f0-9]{24}}":                           GetFeed,
		"PUT    /feeds/{feedid:[a-f0-9]{24}}":                           PutFeed,
		"DELETE /feeds/{feedid:[a-f0-9]{24}}":                           DelFeed,
		"GET    /articles":                                              GetArticles,
		"POST   /articles":                                              PostArticle,
		"GET    /articles/{articleid:[a-f0-9]{24}}":                     GetArticle,
		"PUT    /articles/{articleid:[a-f0-9]{24}}":                     PutArticle,
		"DELETE /articles/{articleid:[a-f0-9]{24}}":                     DelArticle,
		"GET    /pubs/{pubid:[a-f0-9]{24}}/feeds":                       GetFeeds,
		"POST   /pubs/{pubid:[a-f0-9]{24}}/feeds":                       PostFeed,
		"GET    /pubs/{pubid:[a-f0-9]{24}}/feeds/{feedid:[a-f0-9]{24}}": GetFeed,
		"PUT    /pubs/{pubid:[a-f0-9]{24}}/feeds/{feedid:[a-f0-9]{24}}": PutFeed,
		"DELETE /pubs/{pubid:[a-f0-9]{24}}/feeds/{feedid:[a-f0-9]{24}}": DelFeed,
	}

	api := router.PathPrefix("/api").Subrouter()
	for req, f := range apiMap {
		parts := strings.Fields(req)
		r := api.Path(parts[1]).Methods(parts[0]).HandlerFunc(APIHandler(f))
		switch parts[0] {
		case "POST", "PUT":
			r.Headers(headers...)
		}
	}

	router.PathPrefix("/app/").Handler(http.FileServer(http.Dir(AssetsDir)))
	router.HandleFunc("/", HandleIndex)

	return handlers.CombinedLoggingHandler(os.Stdout, router)
}
