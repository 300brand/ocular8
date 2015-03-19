"use strict"

angular.module("adminHandler", [
	"ngRoute",
	"api"
])

.config(["$routeProvider", function($routeProvider) {
	$routeProvider.when("/handlers", {
		controller:  "AdminHandlerController",
		templateUrl: "/app/handlers/handlers.partial.html"
	})
}])

.controller("AdminHandlerController", [
	"$log",
	"$rootScope",
	"$scope",
	"Configs",
	function($log, $rootScope, $scope, Configs) {
		$scope.handlers = []

		$scope.update = function(c, v) {
			c.Value = (v) ? "true" : "false"
			c.$update(function() {
				c.Active = c.Value == "true"
				$log.log("Updated %s to %s", c.Key, c.Value)
			})
		}

		var activeRE = /^\/handlers\/([^\/]+)\/active$/
		var configs = Configs.query(function() {
			for (var i = 0; i < configs.length; i++) {
				var c = configs[i]
				if (!activeRE.test(c.Key)) {
					continue
				}
				var matches = activeRE.exec(c.Key)
				c.Name = matches[1]
				c.Active = c.Value == "true"
				$scope.handlers.push(c)
			}
		})
	}
])
