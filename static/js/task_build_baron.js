mciModule.controller('TaskBuildBaronCtrl', function($scope, $timeout, $http, $location, $window) {
  $scope.getBuildBaronResults = function() {
    $http.get('/ui/plugin/buildbaron/jira_bf_search/' + $scope.taskId ).
      success(function(data, status) {
          if (data && data.issues) {
            $scope.build_baron_results = data.issues
          } else {
            $scope.logs = "No logs found.";
          }
      }).
      error(function(jqXHR, status, errorThrown) {
        //alert('Error retrieving logs: ' + jqXHR);
      });

    // If we already have an outstanding timeout, cancel it
    if ($scope.getBuildBaronTimeout) {
      $timeout.cancel($scope.getBuildBaronTimeout);
    }

    $scope.getBuildBaronTimeout = $timeout(function() {
      $scope.getBuildBaronResults();
    }, 5000);
  };

  $scope.setTask = function(task) {
    $scope.task = task;
    $scope.taskId = task.id;
  };

  $scope.setTask($window.data.task_data);

  if ( $scope.task.status == "failed" ) {
    $scope.getBuildBaronResults();
  }

});
