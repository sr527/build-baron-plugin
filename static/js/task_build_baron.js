mciModule.controller('TaskBuildBaronCtrl', function($scope, $timeout, $http, $location, $window) {
  $scope.build_baron_status = "loading"; 
  $scope.getBuildBaronResults = function() {
    $http.get('/ui/plugin/buildbaron/jira_bf_search/' + $scope.taskId ).
      success(function(data, status) {
          if (data && data.issues) {
            $scope.build_baron_results = data.issues;
            $scope.build_baron_status = "success";
            $window.data.bb = data;
          } else {
            $scope.build_baron_status = "nothing";
          }
      }).
      error(function(jqXHR, status, errorThrown) {
            $scope.build_baron_status = "error";
      });
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
