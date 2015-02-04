package main

import (
	"net/http"
	"net/url"
	"sort"
	"strings"
)

var badKeys = []string{
	"",
	"_r",
	"action_object_map",
	"action_ref_map",
	"action_type_map",
	"amp",
	"asrc",
	"beta",
	"CMP",
	"cmp",
	"cmpid",
	"comm_ref",
	"cpage",
	"dlvrit",
	"ex_cid",
	"f",
	"feedName",
	"feedType",
	"ft",
	"gplus",
	"kc",
	"lifehealth",
	"logvisit",
	"mbid",
	"ncid",
	"npu",
	"op",
	"rpc",
	"s_cid",
	"sc",
	"source",
	"subj",
	"tag",
	"tc",
	"urw",
	"virtualBrandChannel",
}

var badValues = []string{"rss"}

func Clean(dirty string) (clean string, resp *http.Response, err error) {
	if resp, err = resolve(dirty); err != nil {
		return
	}
	var resolved *url.URL
	if resolved, err = resp.Location(); err == http.ErrNoLocation {
		if resolved, err = url.Parse(dirty); err != nil {
			return
		}
	}
	cleanQuery(resolved)
	clean = resolved.String()
	return
}

func cleanQuery(u *url.URL) {
	query := u.Query()
	for k := range query {
		if !valid(k, query.Get(k)) {
			delete(query, k)
		}
	}
	u.RawQuery = query.Encode()
}

func resolve(orig string) (resp *http.Response, err error) {
	resp, err = http.Get(orig)
	return
}

func valid(key, value string) bool {
	if strings.HasPrefix(key, "utm_") {
		return false
	}
	if strings.HasPrefix(key, "fb") {
		return false
	}
	if strings.HasPrefix(key, "atc") {
		return false
	}
	if strings.Contains(key, ".99") {
		return false
	}
	if key == "ref" && value == "25" {
		return false
	}
	if key == "ana" && value == "RSS" {
		return false
	}
	if key == "s" && value == "article_search" {
		return false
	}
	if key == "attr" && value == "all" {
		return false
	}
	if i := sort.SearchStrings(badKeys, key); i < len(badKeys) && badKeys[i] == key {
		return false
	}
	if i := sort.SearchStrings(badValues, value); i < len(badValues) && badValues[i] == value {
		return false
	}
	return true
}
