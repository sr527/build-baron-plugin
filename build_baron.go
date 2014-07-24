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
)

func init() {
	plugin.Publish(&BuildBaronPlugin{})
}

const (
	BUILD_BARON_PLUGIN_NAME = "buildbaron"
	BUILD_BARON_UI_ENDPOINT = "jira_bf_search"
)

type BuildBaronPlugin struct{}

func (self *BuildBaronPlugin) Name() string {
	return BUILD_BARON_PLUGIN_NAME
}

// No API component, so return an empty list of api routes
func (self *BuildBaronPlugin) GetRoutes() []plugin.PluginRoute {
	return []plugin.PluginRoute{}
}

func (self *BuildBaronPlugin) GetUIConfig() *plugin.UIConfig {
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
				Page:     plugin.TASK_PAGE,
				Position: plugin.PAGE_RIGHT,
				PanelHTML: "<div " +
					"ng-include=\"'/ui/plugin/buildbaron/static/partials/task_build_baron.html'\" " +
					// "ng-init='results=plugins.buildbaron' " +
					//			"ng-show='plugins.buildbaron.issues.length' " +
					"></div>",
				Includes: []template.HTML{"<script " +
					"type=\"text/javascript\"" +
					"src=\"/ui/plugin/buildbaron/static/js/task_build_baron.js\"" +
					"></script>"},
			},
		},
	}
}

func BuildBaronHandler(ae web.HandlerApp, mciSettings *mci.MCISettings,
	r *http.Request) web.HTTPResponse {
	//func BuildBaronHandler(r *http.Request) web.HTTPResponse {
	taskId := mux.Vars(r)["task_id"]
	taskFromDb, err := model.FindOneTask(bson.M{"_id": taskId}, nil, nil)
	if err != nil {
		return web.JSONResponse{fmt.Sprintf("Error finding task: %v", err),
			http.StatusInternalServerError}
	}
	jiraInterface := thirdparty.NewJiraInterface(
		mciSettings.Jira.Host,
		mciSettings.Jira.Username,
		mciSettings.Jira.Password,
	)
	results, err := jiraInterface.JQL(fmt.Sprintf("project=BF and text~\"%v\"", taskFromDb.DisplayName))
	if err != nil {
		message := fmt.Sprintf("Error searching jira for ticket: %v", err)
		mci.LOGGER.Errorf(slogger.ERROR, message)
		return web.JSONResponse{message, http.StatusInternalServerError}
	} else {
		return web.JSONResponse{results, http.StatusOK}
	}
}

func (self *BuildBaronPlugin) NewPluginCommand(cmdName string) (plugin.PluginCommand, error) {
	switch cmdName {
	default:
		return nil, fmt.Errorf("No such %v comman %v", BUILD_BARON_PLUGIN_NAME, cmdName)
	}
}
