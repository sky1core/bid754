package main

import (
	"strings"
	"testing"
)

func TestTopLevelPassEvidenceRejectsIndentedSubtestUnderFailingParent(t *testing.T) {
	log := strings.Split(
		"--- FAIL: TestGeneratedReadCases (1.00s)\n"+
			"    --- PASS: TestGeneratedReadCases/case_1 (0.00s)\n"+
			"FAIL\n", "\n")
	if missing := missingEvidence(log, []evidence{topLevelPass("TestGeneratedReadCases")}); len(missing) != 1 {
		t.Fatalf("indented subtest PASS under a failing parent satisfied top-level PASS evidence: missing=%v", missing)
	}
	ok := strings.Split("--- PASS: TestGeneratedReadCases (1.00s)\n", "\n")
	if missing := missingEvidence(ok, []evidence{topLevelPass("TestGeneratedReadCases")}); len(missing) != 0 {
		t.Fatalf("top-level PASS line was not accepted: missing=%v", missing)
	}
}

func TestCountEvidenceRequiresExactTotalBoundary(t *testing.T) {
	want := countLine("decimal32 structured exact comparisons: 1108658/1108658")
	longer := strings.Split("    x_test.go:1: decimal32 structured exact comparisons: 1108658/11086589\n", "\n")
	if missing := missingEvidence(longer, []evidence{want}); len(missing) != 1 {
		t.Fatal("anchored total matched a longer executed/total number")
	}
	sharded := strings.Split("    x_test.go:1: decimal32 structured exact comparisons: 1083/1108658\n", "\n")
	if missing := missingEvidence(sharded, []evidence{want}); len(missing) != 1 {
		t.Fatal("sharded owned/total line satisfied the full-run evidence")
	}
	full := strings.Split("    x_test.go:1: decimal32 structured exact comparisons: 1108658/1108658\n", "\n")
	if missing := missingEvidence(full, []evidence{want}); len(missing) != 0 {
		t.Fatalf("full-run count line was not accepted")
	}
	trailing := strings.Split("Rust structured Tier 1 conversion exact comparisons: 5/5; convenience=57\n", "\n")
	if missing := missingEvidence(trailing, []evidence{countLine("Rust structured Tier 1 conversion exact comparisons: 5/5;")}); len(missing) != 0 {
		t.Fatalf("count line with trailing non-digit text was not accepted")
	}
}
