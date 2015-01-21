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

func GetFeeds(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("GetFeeds\n"))
}

func PostFeed(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("PostFeed\n"))
}

func GetFeed(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("GetFeed\n"))
}

func PutFeed(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("PutFeed\n"))
}

func DelFeed(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("DelFeed\n"))
}

func GetArticles(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("GetArticles\n"))
}

func PostArticle(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("PostArticle\n"))
}

func GetArticle(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("GetArticle\n"))
}

func PutArticle(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("PutArticle\n"))
}

func DelArticle(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("DelArticle\n"))
}
