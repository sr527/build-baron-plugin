mciModule.controller('TaskBuildBaronCtrl', function($scope, $http, $window) {
  $scope.conf = $window.plugins["buildbaron"];
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



  $scope.fileTicket = _.debounce(function(){
    $scope.creatingTicket = true;
    $http.post('/plugin/buildbaron/file_ticket',
        {task: $scope.taskId, tests: $scope.ticketTests}).
      success(function(data, status) {
        $scope.ticketKey = data.key
      }).
    error(function(jqXHR, status) {
      var err = "error filing ticket";
      if (jqXHR) {
        // append an error message if we get one
        err += ": " + jqXHR;
      }
      alert(err);
    }).finally(function(){
        $scope.creatingTicket = false;
    });
  });

  $scope.loaded = false;
  $scope.have_user = $window.have_user;
  $scope.editing = false;
  $scope.editTime = 0;
  $scope.note = "";

  $scope.newTicket = false;
  $scope.ticketTests = [];
  $scope.creatingTicket = false;
  $scope.ticketKey = "";

  $scope.setTask = function(task) {
    $scope.task = task;
    $scope.taskId = task.id;
    $scope.failed = _.filter(task.test_results, function(test){return test.status == 'fail'});
    // special case where we don't need user input when there is only one failure
    if ($scope.failed.length == 1) {
      $scope.ticketTests = [$scope.failed[0].test_file];
    }
  };
  
  $scope.setTask($window.task_data);
  if ( $scope.conf.enabled && $scope.task.status == "failed" ) {
    $scope.build_baron_status = "loading"; 
    $scope.getBuildBaronResults();
  }
  if($scope.conf.enabled){
    $scope.getNote();
  }

  $scope.clearTicket = function(){
    $scope.newTicket = true;
    $scope.ticketKey = "";
    $scope.ticketTests = [];
    $scope.setTask($window.task_data);
  }

});
