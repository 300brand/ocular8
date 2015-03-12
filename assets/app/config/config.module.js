"use strict"

angular.module("adminConfig", [
	"ngRoute",
	"api"
])

.config(["$routeProvider", function($routeProvider) {
	$routeProvider.when("/config", {
		controller:  "AdminConfigController",
		templateUrl: "/app/config/config.partial.html"
	})
}])

angular.module("adminConfig").service("adminConfigService", ["$http", function($http) {
	var etcdHost = "http://localhost:4001"
	this.fullDirectory = function() {
		return $http({
			method: "GET",
			url:    etcdHost + "/v2/keys/",
			params: {
				recursive: true
			}
		})
	}
	this.updateKey = function(key, value) {
		return $http({
			method: "PUT",
			url:    etcdHost + "/v2/keys" + key,
			params: {
				value: value
			}
		})
	}
}])

.controller("AdminConfigController", [
	"$rootScope",
	"$scope",
	"Configs",
	function($rootScope, $scope, adminConfigService) {
		$scope.configs = Configs.query()

		// $scope.updateKey = function(key) {
		// 	console.log(key, $scope.configs[key].value)
		// 	adminConfigService.updateKey(key, $scope.configs[key].value)
		// }

		// var extractConfigs = function(node) {
		// 	if (!node.hasOwnProperty("dir") || !node.dir) {
		// 		$scope.configs[node.key] = node
		// 		return
		// 	}
		// 	for (var i = 0; i < node.nodes.length; i++) {
		// 		extractConfigs(node.nodes[i])
		// 	}
		// }

		// adminConfigService.fullDirectory()
		// 	.success(function(data, status) {
		// 		extractConfigs(data.node)
		// 	})
		// 	.error(function(data, status) {
		// 		console.error(data, status)
		// 	})
	}
])
