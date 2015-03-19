"use strict"

angular.module("adminPubs", [
	"api",
	"ngRoute"
])

.config(["$routeProvider", function($routeProvider) {
	$routeProvider
		.when("/pubs", {
			controller:  "AdminPubsListController",
			templateUrl: "/app/pubs/list.partial.html"
		})
		.when("/pubs/new", {
			controller:  "AdminPubsNewController",
			templateUrl: "/app/pubs/form.partial.html"
		})
		.when("/pubs/:pubid", {
			controller:  "AdminPubsEditController",
			templateUrl: "/app/pubs/form.partial.html"
		})
}])

.controller("AdminPubsEditController", [
	"$log",
	"$rootScope",
	"$routeParams",
	"$scope",
	"Pubs",
	"PubFeeds",
	"Feeds",
	function($log, $rootScope, $routeParams, $scope, Pubs, PubFeeds, Feeds) {
		$scope.Title = "Update"
		$scope.Pub = Pubs.get({ pubid: $routeParams.pubid })
		$scope.Feeds = PubFeeds.query({ pubid: $routeParams.pubid })
		$scope.NewFeed = undefined
		$scope.addNewFeed = function() {
			if ($scope.NewFeed == undefined) {
				return
			}
			for (var i = 0; i < $scope.Feeds.length; i++) {
				if ($scope.Feeds[i].Url == $scope.NewFeed) {
					$scope.NewFeed = ""
					return
				}
			}
			var f = new PubFeeds()
			f.Url = $scope.NewFeed
			$scope.Feeds.push(f)
			$scope.NewFeed = ""
		}
		$scope.save = function() {
			$scope.Pub.$update(function(pub, headers) {
				$log.log("pub:", pub, $scope.Pub)
				for (var i = 0; i < $scope.Feeds.length; i++) {
					var f = $scope.Feeds[i]
					if (f.Id == "") {
						f.$save({ pubid: pub.Id })
					} else if (f.Id != "" && f.deleted) {
						f.$delete()
					}
				}
			})
		}

	}
])

.controller("AdminPubsListController", [
	"$log",
	"$rootScope",
	"$route",
	"$routeParams",
	"$scope",
	"Pubs",
	function($log, $rootScope, $route, $routeParams, $scope, Pubs) {
		$scope.Params = {
			query:  $routeParams.query,
			limit:  $routeParams.limit || 20,
			offset: $routeParams.offset || 0,
			sort:   $routeParams.sort || "name"
		}
		$scope.Total = 0
		$scope.TotalPages = 0
		$scope.Links = {}
		Pubs.query($scope.Params, function(data, responseHeaders) {
			$log.log(responseHeaders())

			$scope.Pubs = data
			$scope.Total = responseHeaders("X-Total-Count")
			$scope.TotalPages = responseHeaders("X-Total-Pages")

			$scope.Links = {
				"first": null,
				"prev":  null,
				"page":  [],
				"next":  null,
				"last":  null
			}
			var links = responseHeaders("Link").split(/, /)
			var relRE = /rel="([^"]+)"/
			var titleRE = /title="([^"]+)"/
			var offsetRE = /offset=(\d+)/
			var p = $scope.Params
			for (var i = 0; i < links.length; i++) {
				var rel = relRE.exec(links[i])[1]
				var l = {
					"title":  titleRE.exec(links[i])[1],
					"offset": offsetRE.exec(links[i])[1],
				}
				l["qs"] = "query="+escape(p.query)+"&limit="+p.limit+"&offset="+l["offset"]+"&sort="+p.sort
				$scope.Links[rel] = l
			}

			$log.log($scope.Links)
		})
		$scope.search = function() {
			$route.updateParams($scope.Params)
		}
	}
])

.controller("AdminPubsNewController", [
	"$log",
	"$rootScope",
	"$scope",
	"Pubs",
	"PubFeeds",
	function($log, $rootScope, $scope, Pubs, PubFeeds) {
		$scope.Title = "New"
		$scope.Pub = new Pubs()
		$scope.Feeds = []
		$scope.NewFeed = undefined
		$scope.addNewFeed = function() {
			if ($scope.NewFeed == undefined) {
				return
			}
			for (var i = 0; i < $scope.Feeds.length; i++) {
				if ($scope.Feeds[i].Url == $scope.NewFeed) {
					$scope.NewFeed = ""
					return
				}
			}
			var f = new PubFeeds()
			f.Url = $scope.NewFeed
			$scope.Feeds.push(f)
			$scope.NewFeed = ""
		}
		$scope.save = function() {
			$scope.Pub.$save(function(pub, headers) {
				$log.log("pub:", pub, $scope.Pub)
				for (var i = 0; i < $scope.Feeds.length; i++) {
					$scope.Feeds[i].$save({ pubid: pub.Id })
				}
			})
		}
	}
])
