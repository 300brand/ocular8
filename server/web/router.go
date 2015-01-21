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

	coll("/pubs", GetPubs, PostPub)
	item("/pubs/{id:[a-f0-9]{24}}", GetPub, PutPub, DelPub)

	router.PathPrefix("/app/").Handler(http.FileServer(http.Dir(config.Config.WebAssets)))
	router.HandleFunc("/", HandleIndex)

	return handlers.CombinedLoggingHandler(os.Stdout, router)
}
