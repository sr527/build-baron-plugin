package buildbaron

import (
	"github.com/evergreen-ci/evergreen"
	"github.com/evergreen-ci/evergreen/model"
	"github.com/evergreen-ci/evergreen/plugin"
	"github.com/evergreen-ci/evergreen/thirdparty"
	"fmt"
	"github.com/10gen-labs/slogger/v1"
	"github.com/gorilla/mux"
	"github.com/mitchellh/mapstructure"
	"html/template"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

func init() {
	plugin.Publish(&BuildBaronPlugin{})
}

const (
	PluginName  = "buildbaron"
	JIRAFailure = "Error searching jira for ticket"
	JQLBFQuery  = "project=BF and ( %v ) order by status asc, updatedDate desc"
)

type jiraOptions struct {
	Host     string
	Username string
	Password string
}

type BuildBaronPlugin struct {
	opts *jiraOptions
}

// A regex that matches either / or \ for splitting directory paths
// on either windows or linux paths.
var eitherSlash *regexp.Regexp = regexp.MustCompile(`[/\\]`)

func (bbp *BuildBaronPlugin) Name() string {
	return PluginName
}

// GetUIHandler adds a path for looking up build failures in JIRA.
func (bbp *BuildBaronPlugin) GetUIHandler() http.Handler {
	if bbp.opts == nil {
		panic("build baron plugin missing configuration")
	}
	r := mux.NewRouter()
	r.Path("/jira_bf_search/{task_id}").HandlerFunc(bbp.buildFailuresSearch)
	return r
}

func (bbp *BuildBaronPlugin) Configure(conf map[string]interface{}) error {
	// pull out the JIRA stuff we need from the config file
	jiraParams := &jiraOptions{}
	err := mapstructure.Decode(conf, jiraParams)
	if err != nil {
		return err
	}
	if jiraParams.Host == "" || jiraParams.Username == "" || jiraParams.Password == "" {
		return fmt.Errorf("Host, username, and password in config must not be blank")
	}
	bbp.opts = jiraParams
	return nil
}

func (bbp *BuildBaronPlugin) GetPanelConfig() (*plugin.PanelConfig, error) {
	root := plugin.StaticWebRootFromSourceFile()
	panelHTML, err := ioutil.ReadFile(root + "/partials/ng_include_task_build_baron.html")
	if err != nil {
		return nil, fmt.Errorf("Can't load panel html file, %v, %v",
			root+"/partials/ng_include_task_build_baron.html", err)
	}

	includeJS, err := ioutil.ReadFile(root + "/partials/script_task_build_baron_js.html")
	if err != nil {
		return nil, fmt.Errorf("Can't load panel html file, %v, %v",
			root+"/partials/script_task_build_baron_js.html", err)
	}

	includeCSS, err := ioutil.ReadFile(root +
		"/partials/link_task_build_baron_css.html")
	if err != nil {
		return nil, fmt.Errorf("Can't load panel html file, %v, %v",
			root+"/partials/link_task_build_baron_css.html", err)
	}

	return &plugin.PanelConfig{
		StaticRoot: plugin.StaticWebRootFromSourceFile(),
		Panels: []plugin.UIPanel{
			{
				Page:      plugin.TaskPage,
				Position:  plugin.PageRight,
				PanelHTML: template.HTML(panelHTML),
				Includes:  []template.HTML{template.HTML(includeCSS), template.HTML(includeJS)},
			},
		},
	}, nil
}

// BuildFailuresSearchHandler handles the requests of searching jira in the build
//  failures project
func (bbp *BuildBaronPlugin) buildFailuresSearch(w http.ResponseWriter, r *http.Request) {
	taskId := mux.Vars(r)["task_id"]
	task, err := model.FindTask(taskId)
	if err != nil {
		plugin.WriteJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	jql := taskToJQL(task)

	jiraHandler := thirdparty.NewJiraHandler(
		bbp.opts.Host,
		bbp.opts.Username,
		bbp.opts.Password,
	)

	results, err := jiraHandler.JQLSearch(jql)
	if err != nil {
		message := fmt.Sprintf("%v: %v, %v", JIRAFailure, err, jql)
		evergreen.Logger.Errorf(slogger.ERROR, message)
		plugin.WriteJSON(w, http.StatusInternalServerError, message)
		return
	}
	plugin.WriteJSON(w, http.StatusOK, results)
}

// In order that we can write tests without an actual jira server handy
type jqlSearcher interface {
	JQLSearch(query string) (*thirdparty.JiraSearchResults, error)
}

// Generates a jira JQL string from the task
// When we search in jira for a task we search in the BF package
// If there are any test results, then we only search by test file
// name of all of the failed tests.
// Otherwise we search by the task name.
func taskToJQL(task *model.Task) string {
	var jqlParts []string
	var jqlClause string
	for _, testResult := range task.TestResults {
		if testResult.Status == evergreen.TestFailedStatus {
			fileParts := eitherSlash.Split(testResult.TestFile, -1)
			jqlParts = append(jqlParts, fmt.Sprintf("text~\"%v\"", fileParts[len(fileParts)-1]))
		}
	}
	if jqlParts != nil {
		jqlClause = strings.Join(jqlParts, " or ")
	} else {
		jqlClause = fmt.Sprintf("text~\"%v\"", task.DisplayName)
	}

	return fmt.Sprintf(JQLBFQuery, jqlClause)
}
