package bid754

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGoportDectestRunnerRejectsPartialExecution(t *testing.T) {
	for _, suite := range goportLoadGeneratedDectestSpec(t).DectestSuites {
		if !isGoportDectestRunnerSuite(suite.TestType) {
			continue
		}
		for _, file := range suite.Files {
			if _, err := os.Stat(filepath.Join("..", "devtools", file)); os.IsNotExist(err) {
				t.Skipf("pinned decTest input absent: %s", file)
			} else if err != nil {
				t.Fatal(err)
			}
		}
	}
	assertDectestRunnerRejectsPartialExecution(t, "TestGeneratedDectestSuitesGoPort", "goport decTest executed suite count")
}

func assertDectestRunnerRejectsPartialExecution(t *testing.T, rootTest, message string) {
	t.Helper()
	for _, tc := range []struct {
		name string
		args []string
	}{
		{"all suites skipped", []string{"-test.run=^" + rootTest + "$", "-test.skip=" + rootTest + "/"}},
		{"one suite skipped", []string{"-test.run=^" + rootTest + "$", "-test.skip=" + rootTest + "/Decimal32$"}},
		{"one suite selected", []string{"-test.run=^" + rootTest + "$/^Decimal32$"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			args := append([]string{"-test.timeout=60s"}, tc.args...)
			out, err := exec.Command(os.Args[0], args...).CombinedOutput()
			if err == nil || !strings.Contains(string(out), message) {
				t.Fatalf("partial execution was not rejected by suite coverage: err=%v\n%s", err, out)
			}
		})
	}
}
