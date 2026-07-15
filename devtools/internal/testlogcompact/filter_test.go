package testlogcompact

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

type terminalErrorReader struct {
	data []byte
	err  error
}

func (r *terminalErrorReader) Read(p []byte) (int, error) {
	if len(r.data) == 0 {
		return 0, r.err
	}
	n := copy(p, r.data)
	r.data = r.data[n:]
	if len(r.data) == 0 {
		return n, r.err
	}
	return n, nil
}

func TestFilterSuppressesSubtestLifecycleButPreservesFailureEvidence(t *testing.T) {
	input := strings.Join([]string{
		"=== RUN   TestGeneratedReadCases",
		"=== RUN   TestGeneratedReadCases/pass_case",
		"    --- PASS: TestGeneratedReadCases/pass_case (0.00s)",
		"=== RUN   TestGeneratedReadCases/skip_case",
		"    readtest_test.go:42: documented native comparison skip",
		"    --- SKIP: TestGeneratedReadCases/skip_case (0.00s)",
		"=== RUN   TestGeneratedReadCases/fail_case",
		"    readtest_test.go:43: expected 01, got 02",
		"    --- FAIL: TestGeneratedReadCases/fail_case (0.00s)",
		"--- FAIL: TestGeneratedReadCases (1.00s)",
		"FAIL",
	}, "\n") + "\n"

	var output bytes.Buffer
	stats, err := Filter(strings.NewReader(input), &output, "TestGeneratedReadCases")
	if err != nil {
		t.Fatalf("Filter: %v", err)
	}
	want := strings.Join([]string{
		"=== RUN   TestGeneratedReadCases",
		"    readtest_test.go:42: documented native comparison skip",
		"    readtest_test.go:43: expected 01, got 02",
		"=== RUN   TestGeneratedReadCases/fail_case",
		"    --- FAIL: TestGeneratedReadCases/fail_case (0.00s)",
		"--- FAIL: TestGeneratedReadCases (1.00s)",
		"FAIL",
	}, "\n") + "\n"
	if output.String() != want {
		t.Fatalf("filtered output mismatch\ngot:\n%s\nwant:\n%s", output.String(), want)
	}
	if stats != (Stats{Runs: 2, Passes: 1, Skips: 1}) {
		t.Fatalf("stats = %+v, want runs=2 passes=1 skips=1", stats)
	}
}

func TestFilterPreservesPendingRunOnAbruptEOF(t *testing.T) {
	input := "=== RUN   TestGeneratedReadCases/crash_case\nfatal error: unexpected signal"
	var output bytes.Buffer
	stats, err := Filter(strings.NewReader(input), &output, "TestGeneratedReadCases")
	if err != nil {
		t.Fatalf("Filter: %v", err)
	}
	if !strings.Contains(output.String(), "TestGeneratedReadCases/crash_case") {
		t.Fatalf("abrupt EOF lost pending subtest identity: %q", output.String())
	}
	if stats != (Stats{}) {
		t.Fatalf("abrupt pending run was reported as suppressed: %+v", stats)
	}
}

func TestFilterPreservesPendingRunOnReadError(t *testing.T) {
	wantErr := errors.New("injected read failure")
	input := &terminalErrorReader{
		data: []byte("=== RUN   TestGeneratedReadCases/read_error_case"),
		err:  wantErr,
	}
	var output bytes.Buffer
	stats, err := Filter(input, &output, "TestGeneratedReadCases")
	if !errors.Is(err, wantErr) {
		t.Fatalf("Filter error = %v, want wrapped %v", err, wantErr)
	}
	if output.String() != "=== RUN   TestGeneratedReadCases/read_error_case" {
		t.Fatalf("read error lost pending subtest identity: %q", output.String())
	}
	if stats != (Stats{}) {
		t.Fatalf("read-error pending run was reported as suppressed: %+v", stats)
	}
}

func TestFilterSuppressesNestedCompletedSubtests(t *testing.T) {
	input := strings.Join([]string{
		"=== RUN   TestGeneratedReadCases/parent",
		"=== RUN   TestGeneratedReadCases/parent/child",
		"        --- PASS: TestGeneratedReadCases/parent/child (0.00s)",
		"    --- PASS: TestGeneratedReadCases/parent (0.00s)",
	}, "\n") + "\n"
	var output bytes.Buffer
	stats, err := Filter(strings.NewReader(input), &output, "TestGeneratedReadCases")
	if err != nil {
		t.Fatalf("Filter: %v", err)
	}
	if output.Len() != 0 {
		t.Fatalf("completed nested lifecycle was not compacted: %q", output.String())
	}
	if stats != (Stats{Runs: 2, Passes: 2}) {
		t.Fatalf("nested stats = %+v, want runs=2 passes=2", stats)
	}
}

func TestFilterPreservesParallelLifecycleAsACompleteSequence(t *testing.T) {
	input := strings.Join([]string{
		"=== RUN   TestGeneratedReadCases/parallel_case",
		"=== PAUSE TestGeneratedReadCases/parallel_case",
		"=== CONT  TestGeneratedReadCases/parallel_case",
		"    --- PASS: TestGeneratedReadCases/parallel_case (0.00s)",
	}, "\n") + "\n"
	var output bytes.Buffer
	stats, err := Filter(strings.NewReader(input), &output, "TestGeneratedReadCases")
	if err != nil {
		t.Fatalf("Filter: %v", err)
	}
	if output.String() != input || stats != (Stats{}) {
		t.Fatalf("parallel lifecycle output/stats = %q/%+v, want unchanged/zero", output.String(), stats)
	}
}

func TestFilterHandlesCompletedPartialFinalLine(t *testing.T) {
	input := "=== RUN   TestGeneratedReadCases/final_case\n    --- PASS: TestGeneratedReadCases/final_case (0.00s)"
	var output bytes.Buffer
	stats, err := Filter(strings.NewReader(input), &output, "TestGeneratedReadCases")
	if err != nil {
		t.Fatalf("Filter: %v", err)
	}
	if output.Len() != 0 || stats != (Stats{Runs: 1, Passes: 1}) {
		t.Fatalf("partial final lifecycle output/stats = %q/%+v", output.String(), stats)
	}
}

func TestFilterPreservesOtherTestLifecycle(t *testing.T) {
	input := "=== RUN   TestOther/case\n    --- PASS: TestOther/case (0.00s)\n"
	var output bytes.Buffer
	stats, err := Filter(strings.NewReader(input), &output, "TestGeneratedReadCases")
	if err != nil {
		t.Fatalf("Filter: %v", err)
	}
	if output.String() != input || stats != (Stats{}) {
		t.Fatalf("other test output/stats = %q/%+v, want unchanged/zero", output.String(), stats)
	}
}

func TestFilterRejectsEmptyRootTest(t *testing.T) {
	var output bytes.Buffer
	if _, err := Filter(strings.NewReader("PASS\n"), &output, ""); err == nil {
		t.Fatal("Filter accepted an empty root test name")
	}
}
