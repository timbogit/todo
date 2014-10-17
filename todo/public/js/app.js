/*global angular */
/*jshint unused:false */
'use strict';

/**
 * The main TodoMVC app module
 *
 * @type {angular.Module}
 */
var todomvc = angular.module('todomvc', ['ngRoute']);

todomvc
  .factory('authInterceptor',
  function() {
    return {
      request: function(config) {
        config.headers = config.headers || {};
        if (localStorage.auth_token) {
          config.headers.token = localStorage.auth_token;
        }
        return config;
      }
    }
  })
  .config(['$httpProvider',
  function($httpProvider) {
    $httpProvider.interceptors.push('authInterceptor');
  }
  ])
  .config(['$routeProvider',
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
