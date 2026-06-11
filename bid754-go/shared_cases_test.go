package bid754

import (
	"fmt"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/sky1core/bid754/bid754-go/internal/testspec"
)

func TestGeneratedDectestSuiteSelection(t *testing.T) {
	spec := loadSharedSpecForTest(t)

	testCases := []struct {
		pattern string
		want    int
	}{
		{pattern: "ds*.decTest", want: 1},
		{pattern: "dd*.decTest", want: 33},
		{pattern: "dq*.decTest", want: 33},
		{pattern: "*.decTest", want: 10},
	}

	for _, tc := range testCases {
		files, err := getTestFiles("../devtools/tests", tc.pattern)
		if err != nil {
			t.Fatalf("getTestFiles(%q) error: %v", tc.pattern, err)
		}
		if len(files) != tc.want {
			t.Fatalf("getTestFiles(%q) returned %d files, want %d", tc.pattern, len(files), tc.want)
		}
	}

	if len(spec.DectestSuites) != 4 {
		t.Fatalf("generated suite count = %d, want 4", len(spec.DectestSuites))
	}

	assertGeneratedSuiteContains(t, spec, "dd*.decTest", "tests/ddQuantize.decTest")
	assertGeneratedSuiteContains(t, spec, "dd*.decTest", "tests/ddToIntegral.decTest")
	assertGeneratedSuiteContains(t, spec, "dd*.decTest", "tests/ddCopySign.decTest")
	assertGeneratedSuiteContains(t, spec, "dd*.decTest", "tests/ddClass.decTest")
	assertGeneratedSuiteContains(t, spec, "dd*.decTest", "tests/ddSameQuantum.decTest")
	assertGeneratedSuiteContains(t, spec, "dd*.decTest", "tests/ddMin.decTest")
	assertGeneratedSuiteContains(t, spec, "dd*.decTest", "tests/ddMax.decTest")
	assertGeneratedSuiteContains(t, spec, "dd*.decTest", "tests/ddMinMag.decTest")
	assertGeneratedSuiteContains(t, spec, "dd*.decTest", "tests/ddMaxMag.decTest")
	assertGeneratedSuiteContains(t, spec, "dd*.decTest", "tests/ddCompareTotal.decTest")
	assertGeneratedSuiteContains(t, spec, "dd*.decTest", "tests/ddCompareTotalMag.decTest")
	assertGeneratedSuiteContains(t, spec, "dd*.decTest", "tests/ddAbs.decTest")
	assertGeneratedSuiteContains(t, spec, "dd*.decTest", "tests/ddPlus.decTest")
	assertGeneratedSuiteContains(t, spec, "dd*.decTest", "tests/ddMinus.decTest")
	assertGeneratedSuiteContains(t, spec, "dd*.decTest", "tests/ddFMA.decTest")
	assertGeneratedSuiteContains(t, spec, "dd*.decTest", "tests/ddLogB.decTest")
	assertGeneratedSuiteContains(t, spec, "dd*.decTest", "tests/ddScaleB.decTest")
	assertGeneratedSuiteContains(t, spec, "dd*.decTest", "tests/ddRemainder.decTest")
	assertGeneratedSuiteContains(t, spec, "dd*.decTest", "tests/ddRemainderNear.decTest")
	assertGeneratedSuiteContains(t, spec, "dd*.decTest", "tests/ddNextToward.decTest")
	assertGeneratedSuiteContains(t, spec, "dd*.decTest", "tests/ddNextPlus.decTest")
	assertGeneratedSuiteContains(t, spec, "dd*.decTest", "tests/ddNextMinus.decTest")
	assertGeneratedSuiteContains(t, spec, "dd*.decTest", "tests/ddReduce.decTest")
	assertGeneratedSuiteContains(t, spec, "dq*.decTest", "tests/dqQuantize.decTest")
	assertGeneratedSuiteContains(t, spec, "dq*.decTest", "tests/dqToIntegral.decTest")
	assertGeneratedSuiteContains(t, spec, "dq*.decTest", "tests/dqCopySign.decTest")
	assertGeneratedSuiteContains(t, spec, "dq*.decTest", "tests/dqClass.decTest")
	assertGeneratedSuiteContains(t, spec, "dq*.decTest", "tests/dqSameQuantum.decTest")
	assertGeneratedSuiteContains(t, spec, "dq*.decTest", "tests/dqMin.decTest")
	assertGeneratedSuiteContains(t, spec, "dq*.decTest", "tests/dqMax.decTest")
	assertGeneratedSuiteContains(t, spec, "dq*.decTest", "tests/dqMinMag.decTest")
	assertGeneratedSuiteContains(t, spec, "dq*.decTest", "tests/dqMaxMag.decTest")
	assertGeneratedSuiteContains(t, spec, "dq*.decTest", "tests/dqCompareTotal.decTest")
	assertGeneratedSuiteContains(t, spec, "dq*.decTest", "tests/dqCompareTotalMag.decTest")
	assertGeneratedSuiteContains(t, spec, "dq*.decTest", "tests/dqAbs.decTest")
	assertGeneratedSuiteContains(t, spec, "dq*.decTest", "tests/dqPlus.decTest")
	assertGeneratedSuiteContains(t, spec, "dq*.decTest", "tests/dqMinus.decTest")
	assertGeneratedSuiteContains(t, spec, "dq*.decTest", "tests/dqFMA.decTest")
	assertGeneratedSuiteContains(t, spec, "dq*.decTest", "tests/dqLogB.decTest")
	assertGeneratedSuiteContains(t, spec, "dq*.decTest", "tests/dqScaleB.decTest")
	assertGeneratedSuiteContains(t, spec, "dq*.decTest", "tests/dqRemainder.decTest")
	assertGeneratedSuiteContains(t, spec, "dq*.decTest", "tests/dqRemainderNear.decTest")
	assertGeneratedSuiteContains(t, spec, "dq*.decTest", "tests/dqNextToward.decTest")
	assertGeneratedSuiteContains(t, spec, "dq*.decTest", "tests/dqNextPlus.decTest")
	assertGeneratedSuiteContains(t, spec, "dq*.decTest", "tests/dqNextMinus.decTest")
	assertGeneratedSuiteContains(t, spec, "*.decTest", "tests/quantize.decTest")
	assertGeneratedSuiteContains(t, spec, "*.decTest", "tests/tointegral.decTest")
	assertGeneratedSuiteContains(t, spec, "*.decTest", "tests/tointegralx.decTest")
	assertGeneratedSuiteContains(t, spec, "*.decTest", "tests/comparesig.decTest")
	assertGeneratedSuiteMissing(t, spec, "*.decTest", "tests/randoms.decTest")
	assertGeneratedSuiteMissing(t, spec, "*.decTest", "tests/testall.decTest")
}

func assertGeneratedSuiteContains(t *testing.T, spec testspec.SharedSpec, pattern, wantFile string) {
	t.Helper()

	for _, suite := range spec.DectestSuites {
		if suite.Pattern != pattern {
			continue
		}
		for _, file := range suite.Files {
			if file == wantFile {
				return
			}
		}
		t.Fatalf("generated suite %q does not include %q", pattern, wantFile)
	}

	t.Fatalf("generated suite %q not found", pattern)
}

func assertGeneratedSuiteMissing(t *testing.T, spec testspec.SharedSpec, pattern, unwantedFile string) {
	t.Helper()

	for _, suite := range spec.DectestSuites {
		if suite.Pattern != pattern {
			continue
		}
		for _, file := range suite.Files {
			if file == unwantedFile {
				t.Fatalf("generated suite %q unexpectedly includes %q", pattern, unwantedFile)
			}
		}
		return
	}

	t.Fatalf("generated suite %q not found", pattern)
}

// FuzzGeneratedArithmeticResultOnlyNative is a native-only, result-string fuzz
// complement. It is not a regular generated verification domain because it
// does not compare decTest status or IEEE exception flags.
func FuzzGeneratedArithmeticResultOnlyNative(f *testing.F) {
	spec, err := loadSharedSpecFromDisk()
	if err != nil {
		f.Fatalf("load shared spec: %v", err)
	}

	for _, tc := range spec.FuzzCases {
		f.Add(tc.TestType, tc.Operation, tc.Operands[0], tc.Operands[1], tc.Precision, tc.Expected, tc.RoundingMode, tc.MaxExponent, tc.MinExponent, tc.Clamp)
	}

	f.Fuzz(func(t *testing.T, testType, op, operand1, operand2 string, precision int, expected, roundingMode string, maxExponent, minExponent, clamp int) {
		requireNative(t)

		execResult, err := executeDecTestOperation(decTestCase{
			Operation:    op,
			Operands:     []string{operand1, operand2},
			Result:       expected,
			Precision:    precision,
			RoundingMode: roundingMode,
			MaxExponent:  maxExponent,
			MinExponent:  minExponent,
			Clamp:        clamp,
		}, testType)
		if err != nil {
			t.Skipf("unsupported generated result-only seed %q: %v", op, err)
		}
		if !compareDecimalResults(expected, execResult.Result) {
			t.Fatalf("%s %s %s (test_type=%s precision=%d): expected %q, got %q", operand1, op, operand2, testType, precision, expected, execResult.Result)
		}
	})
}

func loadSharedSpecForTest(t *testing.T) testspec.SharedSpec {
	t.Helper()
	spec, err := loadSharedSpecFromDisk()
	if err != nil {
		t.Fatalf("load shared spec: %v", err)
	}
	return spec
}

func loadSharedSpecFromDisk() (testspec.SharedSpec, error) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return testspec.SharedSpec{}, fmt.Errorf("resolve shared_cases_test.go path")
	}
	return testspec.LoadGenerated(filepath.Join(filepath.Dir(currentFile), "..", "devtools", "generated", "testspec", "spec_index.json"))
}
