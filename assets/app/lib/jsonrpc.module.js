"use strict"

angular.module("jsonrpc", [])

.provider("jsonrpc", function() {
	var path = "/api/v1"

	this.setPath = function(rpcPath) {
		path = rpcPath
	}

	this.$get = [ "$http", "$rootScope", function($http, $rootScope) {
		$rootScope.RequestsActive = 0

		function doRequest(method, data) {
			console.log("doRequest %d", $rootScope.RequestsActive)
			$rootScope.RequestsActive++
			var id = method + " " + (new Date).toJSON()
			var payload = {
				jsonrpc: "1.0",
				id:      id,
				method:  method,
				params:  []
			}
			if (angular.isDefined(data)) {
				payload.params.push(data)
			}
			var apiCall = $http.post(path, payload)
			apiCall.success(function() {
				console.log("success %d", $rootScope.RequestsActive)
				$rootScope.RequestsActive--
			})
			return apiCall
		}
		return doRequest
	}]
})
