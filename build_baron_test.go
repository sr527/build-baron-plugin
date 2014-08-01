package buildbaron

import (
	. "github.com/smartystreets/goconvey/convey"

	"10gen.com/mci/model"
	"10gen.com/mci/thirdparty"
	"10gen.com/mci/web"
	"fmt"
	"testing"
)

const (
	jiraFailure = "fake jira failed"
)

func TestTaskToJQL(t *testing.T) {
	Convey("Given a task with with two failed tests and one successful test, "+
		"the jql should contian only the failed test names", t, func() {
		task1 := model.Task{}
		task1.TestResults = []model.TestResult{
			model.TestResult{"fail", "foo.js", "", 0, 0, 0},
			model.TestResult{"success", "bar.js", "", 0, 0, 0},
			model.TestResult{"fail", "baz.js", "", 0, 0, 0},
		}
		task1.DisplayName = "foobar"
		jQL1 := taskToJQL(&task1)
		referenceJQL1 := "project=BF and ( text~\"foo.js\" or text~\"baz.js\" ) order by status asc"
		So(jQL1, ShouldEqual, referenceJQL1)
	})

	Convey("Given a task with with oo failed tests, "+
		"the jql should contian only the failed task name", t, func() {
		task2 := model.Task{}
		task2.TestResults = []model.TestResult{}
		task2.DisplayName = "foobar"
		jQL2 := taskToJQL(&task2)
		referenceJQL2 := "project=BF and ( text~\"foobar\" ) order by status asc"
		So(jQL2, ShouldEqual, referenceJQL2)
	})
}

type fakeJira struct {
	total int
}

func (self *fakeJira) JQLSearch(query string) (*thirdparty.JiraSearchResults, error) {
	if self.total < 0 {
		return nil, fmt.Errorf("%v", jiraFailure)
	}
	jiraSearchResults := thirdparty.JiraSearchResults{}
	jiraSearchResults.Total = self.total
	jiraSearchResults.Issues = make([]thirdparty.JiraTicket, self.total, self.total)
	for i := 0; i < self.total; i++ {
		issue := thirdparty.JiraTicket{}
		issue.Fields = &thirdparty.TicketFields{}
		issue.Fields.Summary = fmt.Sprintf("foo %v", i)
		jiraSearchResults.Issues[i] = issue
	}
	return &jiraSearchResults, nil
}

func TestBuildBaronHandler(t *testing.T) {
	Convey("with an arbitrary fake task", t, func() {
		task3 := model.Task{}
		task3.TestResults = []model.TestResult{
			model.TestResult{"fail", "foo.js", "", 0, 0, 0},
			model.TestResult{"success", "bar.js", "", 0, 0, 0},
			model.TestResult{"fail", "baz.js", "", 0, 0, 0},
		}
		task3.DisplayName = "foobar"

		Convey("and 12 results from jira", func() {
			response := buildBaronHandler(&task3, &fakeJira{12})
			jsonResponse, ok := response.(web.JSONResponse)
			So(ok, ShouldBeTrue)
			jiraSearchResults, ok := jsonResponse.Data.(*thirdparty.JiraSearchResults)
			So(ok, ShouldBeTrue)
			So(jiraSearchResults.Total, ShouldEqual, 12)
		})

		Convey("and a error from jira", func() {
			response := buildBaronHandler(&task3, &fakeJira{-1})
			jsonResponse, ok := response.(web.JSONResponse)
			So(ok, ShouldBeTrue)
			message, ok := jsonResponse.Data.(string)
			So(ok, ShouldBeTrue)
			ShouldStartWith(message, JIRA_FAILURE)
		})
	})
}
