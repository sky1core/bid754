package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) {
	return 0, errors.New("injected write failure")
}

func TestRunRejectsSummaryWriteFailure(t *testing.T) {
	input := strings.NewReader(
		"=== RUN   TestGeneratedReadCases/pass_case\n" +
			"    --- PASS: TestGeneratedReadCases/pass_case (0.00s)\n",
	)
	var stderr bytes.Buffer
	if status := run(
		[]string{"-root", "TestGeneratedReadCases"},
		input,
		failingWriter{},
		&stderr,
	); status != 1 {
		t.Fatalf("run status = %d, want 1; stderr=%q", status, stderr.String())
	}
	if !strings.Contains(stderr.String(), "write summary") {
		t.Fatalf("summary write failure was not reported: %q", stderr.String())
	}
}
