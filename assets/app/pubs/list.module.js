"use strict"

angular.module("adminPubs", [
	"ngRoute"
])

.config(["$routeProvider", function($routeProvider) {
	$routeProvider.when("/pubs", {
		controller:  "AdminPubsController",
		templateUrl: "/app/pubs/list.partial.html"
	})
}])

angular.module("adminPubs").service("adminPubsService", ["jsonrpc", function(jsonrpc) {
	this.fetch = function(page) {
		return jsonrpc("Pub.List", {
			Page: page
		})
	}
}])

angular.module("adminPubs").controller("AdminPubsController", [
	"$rootScope",
	"$scope",
	"adminPubsService",
	function($rootScope, $scope, adminPubsService) {
		$scope.show = function(page) {
			adminPubsService.fetch(page)
				.success(function(data, status) {
					console.log(data)
				})
				.error(function(data, status) {
					console.error(data, status)
				})
		}
	}
])
