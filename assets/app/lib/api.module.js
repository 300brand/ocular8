"use strict"

angular.module("api", [
	"ngResource"
])

.factory("Pubs", ["$resource", function($resource) {
	return $resource("/api/v1/pubs/:pubid", { pubid: "@Id" }, {
		"update": { method: "PUT" }
	})
}])

.factory("PubFeeds", ["$resource", function($resource) {
	return $resource("/api/v1/pubs/:pubid/feeds/:feedid", { feedid: "@Id", pubid: "PubId" }, {
		"update": { method: "PUT" }
	})
}])

.factory("Feeds", ["$resource", function($resource) {
	return $resource("/api/v1/feeds/:feedid", { feedid: "@Id" }, {
		"update": { method: "PUT" }
	})
}])

.factory("Articles", ["$resource", function($resource) {
	return $resource("/api/v1/articles/:articleid", { articleid: "@Id" }, {
		"update": { method: "PUT" }
	})
}])

.factory("FeedArticles", ["$resource", function($resource) {
	return $resource("/api/v1/feeds/:feedid/articles/:articleid", { articleid: "@Id", feedid: "@FeedId" }, {
		"update": { method: "PUT" }
	})
}])

.factory("PubArticles", ["$resource", function($resource) {
	return $resource("/api/v1/pubs/:pubid/articles/:articleid", { articleid: "@Id", pubid: "@PubId" }, {
		"update": { method: "PUT" }
	})
}])
