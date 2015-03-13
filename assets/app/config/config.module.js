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

.controller("AdminConfigController", [
	"$rootScope",
	"$scope",
	"Configs",
	function($rootScope, $scope, Configs) {
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
