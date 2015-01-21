package web

import (
	"net/http"
)

func GetPubs(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("GetPubs\n"))
}

func PostPub(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("PostPub\n"))
}

func GetPub(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("GetPub\n"))
}

func PutPub(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("PutPub\n"))
}

func DelPub(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("DeletePub\n"))
}
