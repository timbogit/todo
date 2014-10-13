/*global todomvc */
'use strict';

/**
 * Services that logs the user in and retrieves the JWT token
*/
todomvc.factory('loginServer', ['$http', function ($http) {

  return {
    post: function (usr, passwd) {
      return $http.post('/login/', { "user": usr, "password": passwd });
    }
  };
}]);
