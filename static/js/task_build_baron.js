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

  $scope.getNote = function() {
    $http.get('/plugin/buildbaron/note/' + $scope.taskId ).
      success(function(data, status) {
        // the GET can return null, for empty notes
        if (data && data.content) {
          $scope.note = data.content;
          $scope.editTime = data.time;
        }
      }).
    error(function(jqXHR, status) {
      $scope.build_baron_status = "error";
    }).finally(function(){
      $scope.loaded = true;
    });
  };

  $scope.saveNote = _.debounce(function() {
    // we attach the previous editTime to ensure we 
    // don't overwrite more recent edits the user
    // might have missed
    $http.put('/plugin/buildbaron/note/' + $scope.taskId,
        {content: $scope.note, time: $scope.editTime}).
      success(function(data, status) {
        $scope.editTime = data.time;
        $scope.editing = false;
      }).
    error(function(jqXHR, status) {
      var err = "error saving note";
      if (jqXHR) {
        // append an error message if we get one
        err += ": " + jqXHR;
      }
      alert(err);
    });
  });

  $scope.loaded = false;
  $scope.have_user = $window.have_user;
  $scope.editing = false;
  $scope.editTime = 0;
  $scope.note = "";

  $scope.setTask = function(task) {
    $scope.task = task;
    $scope.taskId = task.id;
  };
  
  $scope.setTask($window.task_data);
  if ( $scope.task.status == "failed" && !$scope.task.task_end_details.timed_out ) {
    $scope.build_baron_status = "loading"; 
    $scope.getBuildBaronResults();
  }
  $scope.getNote();

});
