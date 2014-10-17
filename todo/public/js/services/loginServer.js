/*global todomvc */
'use strict';

/**
 * Services that logs the user in and retrieves the JWT token
*/
todomvc.service('loginServer', ['$http', function ($http) {

  this.login = function(user) {
    return $http.post('/login/', { user: user.email, password: user.password });
  };
}]);
