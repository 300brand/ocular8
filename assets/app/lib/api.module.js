"use strict"

angular.module("api", [
	"ngResource"
])

.factory("Pubs", ["$resource", function($resource) {
	return $resource("/api/pubs/:pubid", null, {
		"update": { method: "PUT" }
	})
}])

.factory("PubFeeds", ["$resource", function($resource) {
	return $resource("/api/pubs/:pubid/feeds/:feedid", null, {
		"update": { method: "PUT" }
	})
}])

.factory("Feeds", ["$resource", function($resource) {
	return $resource("/api/feeds/:feedid", null, {
		"update": { method: "PUT" }
	})
}])
