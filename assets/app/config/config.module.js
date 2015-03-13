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
	"$log",
	"$rootScope",
	"$scope",
	"Configs",
	function($log, $rootScope, $scope, Configs) {
		$scope.configs = []
		$scope.handlerSets = []
		$scope.update = function(c) {
			c.$update(function() {
				$log.log("Updated %s to %s", c.Key, c.Value)
			})
		}

		var configs = Configs.query(function() {
			for (var i = 0; i < configs.length; i++) {
				var c = configs[i]
				var s = c.Key.split(/\//)
				c.HumanKey = s[s.length-1]
				switch (s[1]) {
				case "config":
					$scope.configs.push(c)
					break
				case "handlers":
					if ($scope.handlerSets[s[2]] == undefined) {
						$scope.handlerSets[s[2]] = []
					}
					$scope.handlerSets[s[2]].push(c)
				}
				$log.log(c)
			}
		})
	}
])
