package main

import (
	"os"
	"path/filepath"
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

func TestNativeReadtestEvidencePinsCompactLifecycleCounts(t *testing.T) {
	required, err := nativeReadtestEvidence(anchors{
		ReadtestCasesTotal:             10,
		ReadtestNativeCompareSkipCases: 2,
	})
	if err != nil {
		t.Fatalf("nativeReadtestEvidence: %v", err)
	}
	log := strings.Split(
		"--- PASS: TestGeneratedReadCases (1.00s)\n"+
			"testlogcompact: suppressed 20 subtest lifecycle lines (run=10 pass=8 skip=2) for TestGeneratedReadCases\n",
		"\n",
	)
	if missing := missingEvidence(log, required); len(missing) != 0 {
		t.Fatalf("valid compact native readtest evidence missing=%v", missing)
	}

	wrong := strings.Split(
		"--- PASS: TestGeneratedReadCases (1.00s)\n"+
			"testlogcompact: suppressed 18 subtest lifecycle lines (run=9 pass=7 skip=2) for TestGeneratedReadCases\n",
		"\n",
	)
	if missing := missingEvidence(wrong, required); len(missing) != 1 {
		t.Fatalf("reduced compact native readtest evidence missing=%v, want one count line", missing)
	}
}

func TestNativeReadtestEvidenceRejectsQuotedOrDuplicateCompactSummary(t *testing.T) {
	required, err := nativeReadtestEvidence(anchors{
		ReadtestCasesTotal:             10,
		ReadtestNativeCompareSkipCases: 2,
	})
	if err != nil {
		t.Fatalf("nativeReadtestEvidence: %v", err)
	}
	wantSummary := "testlogcompact: suppressed 20 subtest lifecycle lines (run=10 pass=8 skip=2) for TestGeneratedReadCases"

	quoted := strings.Split(
		"--- PASS: TestGeneratedReadCases (1.00s)\n"+
			"testlogcompact: suppressed 18 subtest lifecycle lines (run=9 pass=7 skip=2) for TestGeneratedReadCases\n"+
			"diagnostic: expected "+wantSummary+"\n",
		"\n",
	)
	if missing := missingEvidence(quoted, required); len(missing) != 1 {
		t.Fatalf("quoted expected summary satisfied compact evidence: missing=%v", missing)
	}

	duplicate := strings.Split(
		"--- PASS: TestGeneratedReadCases (1.00s)\n"+wantSummary+"\n"+wantSummary+"\n",
		"\n",
	)
	if missing := missingEvidence(duplicate, required); len(missing) != 1 {
		t.Fatalf("duplicate compact summaries satisfied unique evidence: missing=%v", missing)
	}
}

func TestNativeReadtestEvidenceRejectsImpossibleAnchors(t *testing.T) {
	for _, a := range []anchors{
		{},
		{ReadtestCasesTotal: 1, ReadtestNativeCompareSkipCases: 2},
	} {
		if _, err := nativeReadtestEvidence(a); err == nil {
			t.Fatalf("nativeReadtestEvidence(%+v) accepted impossible anchors", a)
		}
	}
}

func TestNativeFFIEvidencePinsCompactLifecycleCounts(t *testing.T) {
	required, err := nativeFFIEvidence(anchors{FFIBitcompareCasesTotal: 10})
	if err != nil {
		t.Fatalf("nativeFFIEvidence: %v", err)
	}
	log := strings.Split(
		"--- PASS: TestGeneratedFFIBitCompareSubset (1.00s)\n"+
			"testlogcompact: suppressed 20 subtest lifecycle lines (run=10 pass=10 skip=0) for TestGeneratedFFIBitCompareSubset\n",
		"\n",
	)
	if missing := missingEvidence(log, required); len(missing) != 0 {
		t.Fatalf("valid compact native FFI evidence missing=%v", missing)
	}
	if _, err := nativeFFIEvidence(anchors{}); err == nil {
		t.Fatal("nativeFFIEvidence accepted a zero FFI case anchor")
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

func TestSentinelCCCountEvidenceUsesCompareConversionRows(t *testing.T) {
	path := filepath.Join(t.TempDir(), "verification_sentinels.json")
	pin := `{"tier1_arithmetic_long_routing_sentinel_rows": ["row a"], "tier1_compare_conversion_long_routing_sentinel_rows": ["row a", "row b"]}`
	if err := os.WriteFile(path, []byte(pin), 0o644); err != nil {
		t.Fatalf("write sentinel pin fixture: %v", err)
	}
	want := sentinelCCCountEvidence(path, "Tier 1 compare/conversion routing sentinels")
	full := strings.Split("    x_test.go:1: Tier 1 compare/conversion routing sentinels: 2/2\n", "\n")
	if missing := missingEvidence(full, []evidence{want}); len(missing) != 0 {
		t.Fatalf("full cc sentinel count line was not accepted: missing=%v", missing)
	}
	wrongDomain := strings.Split("    x_test.go:1: Tier 1 compare/conversion routing sentinels: 1/1\n", "\n")
	if missing := missingEvidence(wrongDomain, []evidence{want}); len(missing) != 1 {
		t.Fatal("arithmetic-count line satisfied the compare/conversion sentinel evidence")
	}
}

func TestSentinelCountEvidenceRequiresPinnedFullCount(t *testing.T) {
	path := filepath.Join(t.TempDir(), "verification_sentinels.json")
	pin := `{"tier1_arithmetic_long_routing_sentinel_rows": ["row a", "row b", "row c"]}`
	if err := os.WriteFile(path, []byte(pin), 0o644); err != nil {
		t.Fatalf("write sentinel pin fixture: %v", err)
	}
	want := sentinelCountEvidence(path, "Tier 1 arithmetic routing sentinels")
	full := strings.Split("    x_test.go:1: Tier 1 arithmetic routing sentinels: 3/3\n", "\n")
	if missing := missingEvidence(full, []evidence{want}); len(missing) != 0 {
		t.Fatalf("full sentinel count line was not accepted: missing=%v", missing)
	}
	short := strings.Split("    x_test.go:1: Tier 1 arithmetic routing sentinels: 2/3\n", "\n")
	if missing := missingEvidence(short, []evidence{want}); len(missing) != 1 {
		t.Fatal("reduced sentinel count line satisfied the pinned full-count evidence")
	}
	longer := strings.Split("    x_test.go:1: Tier 1 arithmetic routing sentinels: 3/31\n", "\n")
	if missing := missingEvidence(longer, []evidence{want}); len(missing) != 1 {
		t.Fatal("longer sentinel total satisfied the pinned full-count evidence")
	}
}
