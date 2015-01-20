"use strict"

angular.module("adminPub", [
	"ngRoute"
])

.config(["$routeProvider", function($routeProvider) {
	$routeProvider.when("/pub/list", {
		controller:  "AdminPubController",
		templateUrl: "/app/pub/list/list.partial.html"
	})
}])

angular.module("adminPub").service("adminPubService", ["jsonrpc", function(jsonrpc) {
	this.fetch = function(page) {
		return jsonrpc("Pub.List", {
			Page: page
		})
	}
}])

angular.module("adminPub").controller("AdminPubController", [
	"$rootScope",
	"$scope",
	"adminPubService",
	function($rootScope, $scope, adminPubService) {
		$scope.show = function(page) {
			adminPubService.fetch(page)
				.success(function(data, status) {
					console.log(data)
				})
				.error(function(data, status) {
					console.error(data, status)
				})
		}
	}
])
