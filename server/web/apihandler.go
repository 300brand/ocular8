package web

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/300brand/ocular8/lib/config"
	"github.com/golang/glog"
	"github.com/gorilla/mux"
	"github.com/mattbaird/elastigo/lib"
)

type APIError struct {
	Error error
}

type APIFuncType func(*Context) (interface{}, error)

type APIHandler map[string]APIFuncType

type Context struct {
	Body   io.ReadCloser
	Conn   *elastigo.Conn
	Values url.Values
	Vars   map[string]string
	R      *http.Request
	W      http.ResponseWriter
}

var acceptHeaders = []string{
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
		if i := sort.SearchStrings(acceptHeaders, ct); i == len(acceptHeaders) || acceptHeaders[i] != ct {
			h.writeError(w, http.StatusUnsupportedMediaType, nil)
			return
		}
	}

	ctx := &Context{
		Body:   r.Body,
		Conn:   elastigo.NewConn(),
		Values: r.URL.Query(),
		Vars:   mux.Vars(r),
		R:      r,
		W:      w,
	}
	ctx.Conn.SetHosts(config.ElasticHosts())

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
	case "PUT", "DELETE":
		w.WriteHeader(http.StatusNoContent)
		return
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
