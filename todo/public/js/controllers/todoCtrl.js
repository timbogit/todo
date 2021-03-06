/*global todomvc, angular */
'use strict';

/**
 * The main controller for the app. The controller:
 * - retrieves and persists the model via the todoServer factory
 * - exposes the model to the template and provides event handlers
 */
todomvc.controller('TodoCtrl', function TodoCtrl($scope, $location, $filter, todoServer) {

	$scope.handleError = function(data, status) {
		console.log('code '+status+': '+data);
		//redirect to login page if unauthorized
		if (status == 401) {
			$location.path('/auth')
		}
	};

	$scope.todos = [];
	var todos  = [];
	todoServer.get()
		.success(function (data) {
				$scope.todos = data.Tasks || [];
				todos = $scope.todos;
				$scope.newTodo = '';
				$scope.remainingCount = $filter('filter')(todos, {completed: false}).length;
				$scope.editedTodo = null;

			})
		.error($scope.handleError);


	if ($location.path() === '') {
		$location.path('/');
	}

	$scope.location = $location;

	$scope.$watch('location.path()', function (path) {
		$scope.statusFilter = { '/active': {completed: false}, '/completed': {completed: true} }[path];
	});

	$scope.$watch('remainingCount == 0', function (val) {
		$scope.allChecked = val;
	});

	$scope.putTodos = function(todos) {
		todoServer.put(todos)
			.success(function (data) {
					$scope.todos = data.Tasks;
				})
			.error($scope.handleError);
	};

	$scope.addTodo = function () {
		var newTodo = $scope.newTodo.trim();
		if (newTodo.length === 0) {
			return;
		}

		todoServer.get()
			.success(function (data) {
					$scope.todos = data.Tasks || [];
					todos = $scope.todos;
					todos.push({
						title: newTodo,
						completed: false
					});
					$scope.putTodos(todos);

					$scope.newTodo = '';
					$scope.remainingCount = $filter('filter')(todos, {completed: false}).length;
					$scope.editedTodo = null;

				})
			.error($scope.handleError);

	};

	$scope.editTodo = function (todo) {
		$scope.editedTodo = todo;
		// Clone the original todo to restore it on demand.
		$scope.originalTodo = angular.extend({}, todo);
	};

	$scope.doneEditing = function (todo, index) {
		$scope.editedTodo = null;
		todo.title = todo.title.trim();

		if (!todo.title) {
			$scope.removeTodo(todo);
		}
		todos[index] = todo;
		$scope.putTodos(todos);
	};

	$scope.revertEditing = function (todo) {
		todos[todos.indexOf(todo)] = $scope.originalTodo;
		$scope.doneEditing($scope.originalTodo);
	};

	$scope.removeTodo = function (todo) {
		$scope.remainingCount -= todo.completed ? 0 : 1;
		todos.splice(todos.indexOf(todo), 1);
		$scope.putTodos(todos);
	};

	$scope.todoCompleted = function (todo,index) {
		$scope.remainingCount += todo.completed ? -1 : 1;
		todos[index] = todo;
		$scope.putTodos(todos);
	};

	$scope.clearCompletedTodos = function () {
		$scope.todos = todos = todos.filter(function (val) {
			return !val.completed;
		});
		$scope.putTodos(todos);
	};

	$scope.markAll = function (completed) {
		todos.forEach(function (todo) {
			todo.completed = !completed;
		});
		$scope.remainingCount = completed ? todos.length : 0;
		$scope.putTodos(todos);
	};

});
