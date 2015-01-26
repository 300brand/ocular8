package web

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strings"

	"github.com/golang/glog"
	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2"
)

type APIError struct {
	Error error
}

type APIFuncType func(*Context) (interface{}, error)

type APIHandler map[string]APIFuncType

var Accept = []string{
	"application/json",
	"application/json;charset=UTF-8",
}

func (h APIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")

	handler, ok := h[r.Method]
	if !ok {
		allow := []string{}
		for k := range h {
			allow = append(allow, k)
		}
		sort.Strings(allow)
		w.Header().Set("Allow", strings.Join(allow, ", "))
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
		} else {
			h.writeError(w, http.StatusMethodNotAllowed, nil)
		}
		return
	}

	if m := r.Method; m == "POST" || m == "PUT" {
		ct := r.Header.Get("Content-Type")
		if i := sort.SearchStrings(Accept, ct); i == len(Accept) || Accept[i] != ct {
			h.writeError(w, http.StatusUnsupportedMediaType, nil)
			return
		}
	}

	ctx := &Context{
		Body: r.Body,
		DB:   mongo.Clone().DB(""),
		Vars: mux.Vars(r),
	}
	defer ctx.DB.Session.Close()

	out, err := handler(ctx)
	if err != nil {
		status := http.StatusInternalServerError
		if err == mgo.ErrNotFound {
			status = http.StatusNotFound
		}
		h.writeError(w, status, err)
		return
	}
	switch r.Method {
	case "POST":
		w.WriteHeader(http.StatusCreated)
	case "PUT":
		w.WriteHeader(http.StatusNoContent)
		return
	case "DELETE":
		w.WriteHeader(http.StatusNoContent)
	}
	if err = json.NewEncoder(w).Encode(out); err != nil {
		h.writeError(w, http.StatusInternalServerError, err)
	}
}

func (h APIHandler) writeError(w http.ResponseWriter, status int, err error) {
	if err == nil {
		err = errors.New(http.StatusText(status))
	}
	glog.Errorf("API Error: %s", err)
	w.WriteHeader(status)
	if err = json.NewEncoder(w).Encode(APIError{Error: err}); err != nil {
		h.writeError(w, http.StatusInternalServerError, err)
	}
}
