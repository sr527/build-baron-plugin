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
	"labix.org/v2/mgo/bson"
	"net/http"
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
		return nil, fmt.Errorf("No such %v comman %v", BUILD_BARON_PLUGIN_NAME, cmdName)
	}
}

func (self *BuildBaronPlugin) GetUIConfig() *plugin.UIConfig {
	return &plugin.UIConfig{
		StaticRoot: plugin.StaticWebRootFromSourceFile(),
		Routes: []plugin.PluginRoute{
			plugin.PluginRoute{
				Path:    "/" + BUILD_BARON_UI_ENDPOINT + "/{task_id}",
				Handler: BuildBaronHandler,
				Methods: []string{"GET"},
			},
		},
		Panels: []plugin.UIPanel{
			{
				Page:     plugin.TASK_PAGE,
				Position: plugin.PAGE_RIGHT,
				PanelHTML: "<div " +
					"ng-include=\"'/ui/plugin/" + BUILD_BARON_PLUGIN_NAME + "/static/partials/task_build_baron.html'\" " +
					"></div>",
				Includes: []template.HTML{"<script " +
					"type=\"text/javascript\"" +
					"src=\"/ui/plugin/" + BUILD_BARON_PLUGIN_NAME + "/static/js/task_build_baron.js\"" +
					"></script>"},
			},
		},
	}
}

func BuildBaronHandler(ae web.HandlerApp, mciSettings *mci.MCISettings,
	r *http.Request) web.HTTPResponse {

	taskId := mux.Vars(r)["task_id"]
	task, err := model.FindOneTask(bson.M{"_id": taskId}, nil, nil)
	if err != nil {
		return web.JSONResponse{fmt.Sprintf("Error finding task: %v", err),
			http.StatusInternalServerError}
	}

	jiraInterface := thirdparty.NewJiraInterface(
		mciSettings.Jira.Host,
		mciSettings.Jira.Username,
		mciSettings.Jira.Password,
	)

	return buildBaronHandler(task, &jiraInterface)
}

type jQLSearcher interface {
	JQLSearch(query string) (*thirdparty.JiraSearchResults, error)
}

func buildBaronHandler(task *model.Task, jiraInterface jQLSearcher) web.HTTPResponse {

	jQL := taskToJQL(task)
	results, err := jiraInterface.JQLSearch(jQL)
	if err != nil {
		message := fmt.Sprintf("%v: %v, %v", JIRA_FAILURE, err, jQL)
		mci.LOGGER.Errorf(slogger.ERROR, message)
		return web.JSONResponse{message, http.StatusInternalServerError}
	} else {
		if results.Total > 10 {
			results.Total = 10
			results.Issues = results.Issues[:10]
		}
		return web.JSONResponse{results, http.StatusOK}
	}
}

func taskToJQL(task *model.Task) string {
	var jQLParts []string
	var jQLClause string
	for _, testResult := range task.TestResults {
		if testResult.Status == "fail" {
			jQLParts = append(jQLParts, fmt.Sprintf("text~\"%v\"", testResult.TestFile))
		}
	}
	if jQLParts != nil {
		jQLClause = strings.Join(jQLParts, " and ")
	} else {
		jQLClause = fmt.Sprintf("text~\"%v\"", task.DisplayName)
	}

	return fmt.Sprintf("project=BF and ( %v ) order by status asc", jQLClause)
}
