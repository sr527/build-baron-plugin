package buildbaron

import (
	"10gen.com/mci/model"
	"testing"
)

func TestTaskToJQL(t *testing.T) {
	task1 := model.Task{}
	task1.TestResults = []model.TestResult{
		model.TestResult{"fail", "foo.js", "", 0, 0, 0},
		model.TestResult{"success", "bar.js", "", 0, 0, 0},
		model.TestResult{"fail", "baz.js", "", 0, 0, 0},
	}
	task1.DisplayName = "foobar"
	jQL1 := taskToJQL(&task1)
	referenceJQL1 := "project=BF and ( text~\"foo.js\" and text~\"baz.js\" ) order by status asc"
	if jQL1 != referenceJQL1 {
		t.Errorf("taskToJQL failed to produce the correct output: %v != %v", jQL1, referenceJQL1)
	}

	task2 := model.Task{}
	task2.TestResults = []model.TestResult{}
	task2.DisplayName = "foobar"
	jQL2 := taskToJQL(&task2)
	referenceJQL2 := "project=BF and ( text~\"foobar\" ) order by status asc"
	if jQL2 != referenceJQL2 {
		t.Errorf("taskToJQL failed to produce the correct output: %v != %v", jQL2, referenceJQL2)
	}
}
