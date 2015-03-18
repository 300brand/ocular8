"use strict"

angular.module("admin", [
	"ngRoute",
	"jsonrpc",
	"xeditable",

	"adminConfig",
	"adminHandler",
	"adminPubs"
])

.config(["$routeProvider", "$locationProvider", function($routeProvider, $locationProvider) {
	$locationProvider.html5Mode(false)
	$locationProvider.hashPrefix("")
	// $routeProvider.otherwise({ redirectTo: "/" })
}])

.run(["editableOptions", "editableThemes", function(editableOptions, editableThemes) {
	editableOptions.theme           = "bs3"
	editableThemes.bs3.inputClass   = 'input-sm';
	editableThemes.bs3.buttonsClass = 'btn-sm';
}])
