"use strict"

angular.module("adminPubs", [
	"ngResource",
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
}])

.factory("Pubs", ["$resource", function($resource) {
	return $resource("/api/pubs/:pubid", null, {
		"update": { method: "PUT" }
	})
}])

.controller("AdminPubsListController", [
	"$log",
	"$rootScope",
	"$scope",
	"Pubs",
	function($log, $rootScope, $scope, Pubs) {
		$scope.Pubs = Pubs.query(function() {
			$log.log("Pubs.query")
		})
	}
])

.controller("AdminPubsNewController", [
	"$log",
	"$rootScope",
	"$scope",
	"Pubs",
	function($log, $rootScope, $scope, Pubs) {
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
			$scope.Feeds.push({ Url: $scope.NewFeed })
			$scope.NewFeed = ""
		}
		$scope.save = function() {
			$scope.Pub.$save(function() {
				$log.log("Saved", $scope.Pub)
			})
		}
	}
])
