package buildbaron

import (
	"github.com/evergreen-ci/evergreen/model"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestDescriptionGeneration(t *testing.T) {
	Convey("With a set of details, a valid description should be generated", t, func() {
		out, err := getDescription(
			&model.Task{
				DisplayName:  "My Task",
				Id:           "mytaskid1",
				BuildVariant: "osx-108",
			},
			"myUser",
			[]jiraTestFailure{
				{Name: "1.js", URL: "path/to/1"},
				{Name: "2.js", URL: "path/to/2"},
			},
		)
		So(err, ShouldBeNil)
	})
}

func TestSummaryGeneration(t *testing.T) {
	Convey("With different amounts of failures", t, func() {
		taskName := "Test Task"

		Convey("a task with no failures should return the task name", func() {
			failures := []jiraTestFailure{}
			So(getSummary(taskName, failures), ShouldEqual, "Test Task failure")
		})

		Convey("a task with some failures should return those failures ", func() {
			Convey("for one failure", func() {
				failures := []jiraTestFailure{{Name: "1.js"}}
				So(getSummary(taskName, failures), ShouldEqual, "1.js")
				Convey("or two", func() {
					failures = append(failures, jiraTestFailure{Name: "2.js"})
					So(getSummary(taskName, failures), ShouldEqual, "1.js, 2.js")
					Convey("or three", func() {
						failures = append(failures, jiraTestFailure{Name: "3.js"})
						So(getSummary(taskName, failures), ShouldEqual, "1.js, 2.js, 3.js")
						Convey("or four", func() {
							failures = append(failures, jiraTestFailure{Name: "4.js"})
							So(getSummary(taskName, failures), ShouldEqual, "1.js, 2.js, 3.js, 4.js")
						})
					})
				})
			})
		})

		Convey("but a task with many failures (>4) should not summarize them", func() {
			failures := []jiraTestFailure{
				{Name: "1"},
				{Name: "2"},
				{Name: "3"},
				{Name: "4"},
				{Name: "5"},
			}
			So(getSummary(taskName, failures), ShouldEqual, "Test Task failures")
		})
	})

}