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
			var f = new PubFeeds()
			f.Url = $scope.NewFeed
			$scope.Feeds.push(f)
			$scope.NewFeed = ""
		}
		$scope.save = function() {
			$scope.Pub.$save(function(pub, headers) {
				for (var i = 0; i < $scope.Feeds.length; i++) {
					$scope.Feeds[i].$save({ pubid: pub.Id })
				}
			})
		}
	}
])
