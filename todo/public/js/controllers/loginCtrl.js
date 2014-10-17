/*global todomvc, angular */
'use strict';

/**
 * The controller for the login page of the application.
 * This controller:
 * - retrieves and persists the JWT token via the loginServer factory
 */
todomvc.controller('LoginCtrl', function LoginCtrl($scope, $location, loginServer) {
  $scope.login = function() {
    if ($scope.loginForm.$valid) {
      loginServer.login($scope.user).then(success, error);
    }
  };

  var success = function(response) {
    $scope.wrongCredentials = false;
    localStorage.setItem('auth_token', response.data.token);
    $location.path('/');
  };

  var error = function(response) {
    $scope.wrongCredentials = true;
  }
});
