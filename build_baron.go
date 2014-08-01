package buildbaron

import (
	"10gen.com/mci"
	"10gen.com/mci/model"
	"10gen.com/mci/plugin"
	"10gen.com/mci/thirdparty"
	"10gen.com/mci/web"
	"fmt"
	"github.com/10gen-labs/slogger/v1"
	"github.com/gorilla/mux"
	"html/template"
	"net/http"
	"regexp"
	"strings"
)

func init() {
	plugin.Publish(&BuildBaronPlugin{})
}

const (
	BUILD_BARON_PLUGIN_NAME = "buildbaron"
	BUILD_BARON_UI_ENDPOINT = "jira_bf_search"
	JIRA_FAILURE            = "Error searching jira for ticket"
)

type BuildBaronPlugin struct{}

var eitherSlash *regexp.Regexp = regexp.MustCompile("[/\\\\]")

func (self *BuildBaronPlugin) Name() string {
	return BUILD_BARON_PLUGIN_NAME
}

// No API component, so return an empty list of api routes
func (self *BuildBaronPlugin) GetRoutes() []plugin.PluginRoute {
	return []plugin.PluginRoute{}
}

// We don't provide any Commands, so this just always returns an error
func (self *BuildBaronPlugin) NewPluginCommand(cmdName string) (plugin.PluginCommand, error) {
	switch cmdName {
	default:
		return nil, fmt.Errorf("%v has no commands, especially not %v", BUILD_BARON_PLUGIN_NAME, cmdName)
	}
}

func (self *BuildBaronPlugin) GetUIConfig() *plugin.UIConfig {
	panelHTML := template.HTML(fmt.Sprintf(
		`<div ng-include="'/ui/plugin/%v/static/partials/task_build_baron.html'"></div>`,
		BUILD_BARON_PLUGIN_NAME))
	includeJS := template.HTML(fmt.Sprintf(
		`<script type="text/javascript" src="/ui/plugin/%v/static/js/task_build_baron.js"></script>`,
		BUILD_BARON_PLUGIN_NAME))
	includeCSS := template.HTML(fmt.Sprintf(
		`<link href="/ui/plugin/%v/static/css/task_build_baron.css" rel="stylesheet"/>`,
		BUILD_BARON_PLUGIN_NAME))
	return &plugin.UIConfig{
		StaticRoot: plugin.StaticWebRootFromSourceFile(),
		Routes: []plugin.PluginRoute{
			plugin.PluginRoute{
				Path:    fmt.Sprintf("/%v/{task_id}", BUILD_BARON_UI_ENDPOINT),
				Handler: BuildBaronHandler,
				Methods: []string{"GET"},
			},
		},
		Panels: []plugin.UIPanel{
			{
				Page:      plugin.TASK_PAGE,
				Position:  plugin.PAGE_RIGHT,
				PanelHTML: panelHTML,
				Includes:  []template.HTML{includeCSS, includeJS},
			},
		},
	}
}

// BuildBaronHandler handles the requests of searching jira in the build
//  failures project
func BuildBaronHandler(ae web.HandlerApp, mciSettings *mci.MCISettings,
	r *http.Request) web.HTTPResponse {

	taskId := mux.Vars(r)["task_id"]
	task, err := model.FindTask(taskId)
	if err != nil {
		return web.JSONResponse{fmt.Sprintf("Error finding task: %v", err),
			http.StatusInternalServerError}
	}

	jiraHandler := thirdparty.NewJiraHandler(
		mciSettings.Jira.Host,
		mciSettings.Jira.Username,
		mciSettings.Jira.Password,
	)

	return buildBaronHandler(task, &jiraHandler)
}

// In order that we can write tests without an actual jira server handy
type jqlSearcher interface {
	JQLSearch(query string) (*thirdparty.JiraSearchResults, error)
}

func buildBaronHandler(task *model.Task, jiraHandler jqlSearcher) web.HTTPResponse {

	jql := taskToJQL(task)
	results, err := jiraHandler.JQLSearch(jql)
	if err != nil {
		message := fmt.Sprintf("%v: %v, %v", JIRA_FAILURE, err, jql)
		mci.LOGGER.Errorf(slogger.ERROR, message)
		return web.JSONResponse{message, http.StatusInternalServerError}
	}
	return web.JSONResponse{results, http.StatusOK}
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
		if testResult.Status == mci.TEST_FAILED_STATUS {
			fileParts := eitherSlash.Split(testResult.TestFile, -1)
			jqlParts = append(jqlParts, fmt.Sprintf("text~\"%v\"", fileParts[len(fileParts)-1]))
		}
	}
	if jqlParts != nil {
		jqlClause = strings.Join(jqlParts, " or ")
	} else {
		jqlClause = fmt.Sprintf("text~\"%v\"", task.DisplayName)
	}

	return fmt.Sprintf("project=BF and ( %v ) order by status asc", jqlClause)
}
