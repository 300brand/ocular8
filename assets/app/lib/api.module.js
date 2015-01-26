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
	return $resource("/api/v1/pubs/:pubid/feeds/:feedid", null, {
		"update": { method: "PUT" }
	})
}])

.factory("Feeds", ["$resource", function($resource) {
	return $resource("/api/v1/feeds/:feedid", null, {
		"update": { method: "PUT" }
	})
}])
