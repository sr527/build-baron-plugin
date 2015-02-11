mciModule.controller('TaskBuildBaronCtrl', function($scope, $http, $window) {
  $scope.getBuildBaronResults = function() {
    $http.get('/plugin/buildbaron/jira_bf_search/' + $scope.taskId ).
      success(function(data, status) {
          if (data && data.issues && data.issues.length > 0 ) {
            $scope.build_baron_results = data.issues;
            $scope.build_baron_status = "success";
          } else {
            $scope.build_baron_status = "nothing";
          }
      }).
      error(function(jqXHR, status, errorThrown) {
            $scope.build_baron_status = "error";
      });
  };
  $scope.have_user = $window.have_user

  $scope.setTask = function(task) {
    $scope.task = task;
    $scope.taskId = task.id;
  };

  $scope.setTask($window.task_data);

  if ( $scope.task.status == "failed" && ! $scope.task.task_end_details.timed_out ) {
    $scope.build_baron_status = "loading"; 
    $scope.getBuildBaronResults();
  }

});
