/*global angular */
/*jshint unused:false */
'use strict';

/**
 * The main TodoMVC app module
 *
 * @type {angular.Module}
 */
var todomvc = angular.module('todomvc', ['ngRoute']);

todomvc.config(['$routeProvider',
  function($routeProvider) {
    $routeProvider.
      when('/', {
        templateUrl: 'partials/todo-list.html',
        controller: 'TodoCtrl'
      }).
      when('/active', {
        templateUrl: 'partials/todo-list.html',
        controller: 'TodoCtrl'
      }).
      when('/completed', {
        templateUrl: 'partials/todo-list.html',
        controller: 'TodoCtrl'
      }).
      when('/auth', {
        templateUrl: 'partials/login.html',
        controller: 'LoginCtrl'
      }).
      otherwise({
        redirectTo: '/'
      });
  }]);
