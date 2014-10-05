/*global todomvc */
'use strict';

/**
 * Services that persists and retrieves TODOs from localStorage
*/
todomvc.factory('todoServer', ['$http', function ($http) {

  return {
    get: function () {
      return $http.get('/task/');
    },
    put: function (todos) {
      return $http.put('/task/', { "Tasks": todos });
    }
  };
}]);
