package web

import (
	"github.com/300brand/ocular8/server/config"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"net/http"
	"os"
)

func Handler() http.Handler {
	Router := mux.NewRouter()
	Router.PathPrefix("/").Handler(http.FileServer(http.Dir(config.Config.Web.Assets)))
	return handlers.CombinedLoggingHandler(os.Stdout, Router)
}
