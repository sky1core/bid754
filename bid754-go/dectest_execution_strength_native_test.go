//go:build cgo && bid754_native

package bid754

import "testing"

func TestNativeDectestRunnerRejectsPartialExecution(t *testing.T) {
	assertDectestRunnerRejectsPartialExecution(t, "TestGeneratedDectestSuites", "native decTest executed suite count")
}
