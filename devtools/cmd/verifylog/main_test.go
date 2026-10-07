package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func arithmeticLongTestAnchors() anchors {
	structuredDeviations := uint64(2)
	randomDeviations := uint64(0)
	return anchors{
		Tier1ArithmeticStructured:           map[string]uint64{"decimal32": 10, "decimal64": 20, "decimal128": 30},
		Tier1ArithmeticRandomCasesPerOp:     map[string]uint64{"decimal32": 4, "decimal64": 5, "decimal128": 6},
		Tier1ArithmeticRandomOperations:     2,
		Tier1ArithmeticStructuredDeviations: &structuredDeviations,
		Tier1ArithmeticRandomDeviations:     &randomDeviations,
	}
}

func arithmeticLongTestLog(rust bool) string {
	if rust {
		return "test result: ok. 5 passed; 0 failed;\n" +
			"test tier1_arithmetic_scaleb_intel003_witness ... INTEL-BID-003 witness: C comparisons=90 independent deviations=90\n" +
			"ok\n" +
			"Rust Tier 1 arithmetic routing sentinels: 1/1\n" +
			"Rust Decimal32 structured Tier 1 exact comparisons: 10/10\n" +
			"Rust Decimal32 random Tier 1 exact comparisons: 8/8\n" +
			"Rust Decimal64 structured Tier 1 exact comparisons: 20/20\n" +
			"Rust Decimal64 random Tier 1 exact comparisons: 10/10\n" +
			"Rust Decimal128 structured Tier 1 comparisons: C exact=28 independent INTEL-BID-003=2 total=30/30\n" +
			"Rust Decimal128 random Tier 1 comparisons: C exact=12 independent INTEL-BID-003=0 total=12/12\n"
	}
	return "--- PASS: TestTier1ArithmeticCorpusContract (0.01s)\n" +
		"--- PASS: TestTier1ArithmeticRoutingSentinels (0.01s)\n" +
		"--- PASS: TestTier1ArithmeticScaleBIntel003Witness (0.01s)\n" +
		"--- PASS: TestTier1ArithmeticStructuredNativeDifferential (0.01s)\n" +
		"--- PASS: TestTier1ArithmeticDeterministicRandomNativeDifferential (0.01s)\n" +
		"INTEL-BID-003 witness: C comparisons=90 independent deviations=90\n" +
		"Tier 1 arithmetic routing sentinels: 1/1\n" +
		"decimal32 structured exact comparisons: 10/10\n" +
		"decimal32 deterministic random exact comparisons: 8/8\n" +
		"decimal64 structured exact comparisons: 20/20\n" +
		"decimal64 deterministic random exact comparisons: 10/10\n" +
		"decimal128 structured comparisons: C exact=28 independent INTEL-BID-003=2 total=30/30\n" +
		"decimal128 deterministic random comparisons: C exact=12 independent INTEL-BID-003=0 total=12/12\n"
}

func TestTier1ArithmeticEvidenceRejectsIncompleteOrFalseLogs(t *testing.T) {
	path := filepath.Join(t.TempDir(), "verification_sentinels.json")
	if err := os.WriteFile(path, []byte(`{"tier1_arithmetic_long_routing_sentinel_rows":["row"]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, rust := range []bool{false, true} {
		a := arithmeticLongTestAnchors()
		required, err := tier1ArithmeticEvidence(a, path, rust)
		if err != nil {
			t.Fatalf("rust=%t: %v", rust, err)
		}
		valid := arithmeticLongTestLog(rust)
		if missing := missingEvidence(strings.Split(valid, "\n"), required); len(missing) != 0 {
			t.Fatalf("rust=%t valid log missing=%v", rust, missing)
		}
		prefix := "decimal128"
		randomPrefix := "decimal128 deterministic random"
		oldWidth := "decimal64 structured"
		if rust {
			prefix = "Rust Decimal128"
			randomPrefix = "Rust Decimal128 random Tier 1"
			oldWidth = "Rust Decimal64 structured"
		}
		for _, tc := range []struct{ name, old, replacement string }{
			{"missing old width", oldWidth, "old width absent"},
			{"all C exact", "C exact=28 independent INTEL-BID-003=2", "C exact=30 independent INTEL-BID-003=0"},
			{"wrong C exact", "C exact=28", "C exact=29"},
			{"wrong deviation", "INTEL-BID-003=2 total=30/30", "INTEL-BID-003=1 total=30/30"},
			{"partial structured", "total=30/30", "total=29/30"},
			{"missing random", randomPrefix, "missing random"},
			{"partial random", "total=12/12", "total=11/12"},
			{"missing witness count", "INTEL-BID-003 witness: C comparisons=90", "INTEL-BID-003 witness: C comparisons=89"},
		} {
			t.Run(fmt.Sprintf("rust=%t/%s", rust, tc.name), func(t *testing.T) {
				changed := strings.Replace(valid, tc.old, tc.replacement, 1)
				if changed == valid || len(missingEvidence(strings.Split(changed, "\n"), required)) == 0 {
					t.Fatalf("changed %q did not invalidate evidence for %s", tc.old, prefix)
				}
			})
		}
		witnessName := "--- PASS: TestTier1ArithmeticScaleBIntel003Witness"
		if rust {
			witnessName = "test tier1_arithmetic_scaleb_intel003_witness ... "
		}
		if len(missingEvidence(strings.Split(strings.Replace(valid, witnessName, "witness absent", 1), "\n"), required)) == 0 {
			t.Fatalf("rust=%t missing witness execution accepted", rust)
		}
	}
}

func TestTier1ArithmeticEvidenceRejectsInvalidAnchors(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func(*anchors)
	}{
		{"missing structured deviation", func(a *anchors) { a.Tier1ArithmeticStructuredDeviations = nil }},
		{"missing random deviation", func(a *anchors) { a.Tier1ArithmeticRandomDeviations = nil }},
		{"zero structured deviation", func(a *anchors) { zero := uint64(0); a.Tier1ArithmeticStructuredDeviations = &zero }},
		{"excess deviation", func(a *anchors) { excess := uint64(31); a.Tier1ArithmeticStructuredDeviations = &excess }},
		{"missing old count", func(a *anchors) { delete(a.Tier1ArithmeticStructured, "decimal32") }},
		{"missing random count", func(a *anchors) { delete(a.Tier1ArithmeticRandomCasesPerOp, "decimal64") }},
		{"excess random deviation", func(a *anchors) { excess := uint64(13); a.Tier1ArithmeticRandomDeviations = &excess }},
		{"zero random operation", func(a *anchors) { a.Tier1ArithmeticRandomOperations = 0 }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a := arithmeticLongTestAnchors()
			tc.change(&a)
			if _, err := tier1ArithmeticEvidence(a, "", false); err == nil {
				t.Fatal("invalid arithmetic anchor accepted")
			}
		})
	}
}

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

func TestNativeDectestEvidenceRejectsEmptyAndPartialSuites(t *testing.T) {
	a := anchors{
		DectestSuiteCases:         map[string]uint64{"Decimal32": 10, "Decimal64": 20, "Decimal128": 30, "General": 40},
		NativeDectestSkippedCases: map[string]uint64{"Decimal32": 0, "Decimal64": 2, "Decimal128": 3, "General": 4},
	}
	required, err := nativeDectestEvidence(a)
	if err != nil {
		t.Fatal(err)
	}
	parents := "--- PASS: TestGeneratedDectestSuites (0.01s)\n" +
		"--- PASS: TestGeneratedDectestPlusMinusQuantumStrengthGoPort (0.01s)\n" +
		"--- PASS: TestGeneratedDectestPlusMinusQuantumStrengthUnaryAdapter (0.01s)\n" +
		"--- PASS: TestNativeDectestRunnerRejectsPartialExecution (0.01s)\n"
	valid := parents +
		"    runner_test.go:1: native decTest Decimal32: passed=10 failed=0 skipped=0 total=10\n" +
		"    runner_test.go:1: native decTest Decimal64: passed=18 failed=0 skipped=2 total=20\n" +
		"    runner_test.go:1: native decTest Decimal128: passed=27 failed=0 skipped=3 total=30\n" +
		"    runner_test.go:1: native decTest General: passed=36 failed=0 skipped=4 total=40\n"
	for name, log := range map[string]string{
		"valid":              valid,
		"parents only":       parents,
		"missing suite":      strings.ReplaceAll(valid, "native decTest General:", "missing:"),
		"partial execution":  strings.ReplaceAll(valid, "passed=18", "passed=17"),
		"case failure":       strings.ReplaceAll(valid, "failed=0", "failed=1"),
		"wrong skip count":   strings.ReplaceAll(valid, "skipped=2", "skipped=3"),
		"shrunk raw cases":   strings.ReplaceAll(valid, "total=40", "total=39"),
		"missing regression": strings.ReplaceAll(valid, "--- PASS: TestNativeDectestRunnerRejectsPartialExecution", "--- SKIP: TestNativeDectestRunnerRejectsPartialExecution"),
	} {
		t.Run(name, func(t *testing.T) {
			missing := missingEvidence(strings.Split(log, "\n"), required)
			if (len(missing) == 0) != (name == "valid") {
				t.Fatalf("invalid execution verdict: missing=%v", missing)
			}
		})
	}
	delete(a.NativeDectestSkippedCases, "General")
	if _, err := nativeDectestEvidence(a); err == nil {
		t.Fatal("missing suite anchor accepted")
	}
	a.NativeDectestSkippedCases["General"] = 40
	if _, err := nativeDectestEvidence(a); err == nil {
		t.Fatal("zero executed cases accepted")
	}
}

func TestGoportDectestEvidenceRejectsEmptyAndPartialSuites(t *testing.T) {
	a := anchors{
		DectestSuiteCases:            map[string]uint64{"Decimal32": 10, "Decimal64": 20, "Decimal128": 30, "General": 40},
		GoportDectestExecutedCases:   map[string]uint64{"Decimal32": 10, "Decimal64": 18, "Decimal128": 27},
		GoportDectestSkippedCases:    map[string]uint64{"Decimal32": 0, "Decimal64": 2, "Decimal128": 3},
		GoportDectestFlagExemptCases: map[string]uint64{"Decimal32": 0, "Decimal64": 1, "Decimal128": 0},
	}
	required, err := goportDectestEvidence(a)
	if err != nil {
		t.Fatal(err)
	}
	parents := "--- PASS: TestGeneratedDectestSuitesGoPort (0.01s)\n" +
		"--- PASS: TestGeneratedDectestPlusMinusQuantumStrengthGoPort (0.01s)\n" +
		"--- PASS: TestGeneratedDectestPlusMinusQuantumStrengthUnaryAdapter (0.01s)\n" +
		"--- PASS: TestGoportDectestRunnerRejectsPartialExecution (0.01s)\n"
	valid := parents +
		"    runner_test.go:1: goport decTest Decimal32: executed=10 failed=0 skipped=0 flagExempt=0 total=10\n" +
		"    runner_test.go:1: goport decTest Decimal64: executed=18 failed=0 skipped=2 flagExempt=1 total=20\n" +
		"    runner_test.go:1: goport decTest Decimal128: executed=27 failed=0 skipped=3 flagExempt=0 total=30\n"
	for name, log := range map[string]string{
		"valid":             valid,
		"parents only":      parents,
		"missing suite":     strings.ReplaceAll(valid, "goport decTest Decimal128:", "missing:"),
		"partial execution": strings.ReplaceAll(valid, "executed=18", "executed=17"),
		"failed cases":      strings.ReplaceAll(valid, "failed=0", "failed=1"),
		"wrong skips":       strings.ReplaceAll(valid, "skipped=2", "skipped=3"),
		"weakened flags":    strings.ReplaceAll(valid, "flagExempt=1", "flagExempt=2"),
		"shrunk raw cases":  strings.ReplaceAll(valid, "total=30", "total=29"),
	} {
		t.Run(name, func(t *testing.T) {
			if missing := missingEvidence(strings.Split(log, "\n"), required); (len(missing) == 0) != (name == "valid") {
				t.Fatalf("invalid execution verdict: missing=%v", missing)
			}
		})
	}
	a.GoportDectestExecutedCases["Decimal32"] = 0
	if _, err := goportDectestEvidence(a); err == nil {
		t.Fatal("zero executed coverage accepted")
	}
	a.GoportDectestExecutedCases["Decimal32"] = 10
	delete(a.GoportDectestSkippedCases, "Decimal64")
	if _, err := goportDectestEvidence(a); err == nil {
		t.Fatal("missing skip anchor accepted")
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

func TestGoportReadtestEvidencePinsCompactLifecycleCounts(t *testing.T) {
	required, err := goportReadtestEvidence(anchors{GoportReadtestExecutedCases: 10})
	if err != nil {
		t.Fatalf("goportReadtestEvidence: %v", err)
	}
	log := strings.Split(
		"--- PASS: TestGeneratedReadCasesGoPort (1.00s)\n"+
			"testlogcompact: suppressed 20 subtest lifecycle lines (run=10 pass=10 skip=0) for TestGeneratedReadCasesGoPort\n",
		"\n",
	)
	if missing := missingEvidence(log, required); len(missing) != 0 {
		t.Fatalf("valid compact Go-port readtest evidence missing=%v", missing)
	}

	wrong := strings.Split(
		"--- PASS: TestGeneratedReadCasesGoPort (1.00s)\n"+
			"testlogcompact: suppressed 18 subtest lifecycle lines (run=9 pass=9 skip=0) for TestGeneratedReadCasesGoPort\n",
		"\n",
	)
	if missing := missingEvidence(wrong, required); len(missing) != 1 {
		t.Fatalf("reduced compact Go-port readtest evidence missing=%v, want one count line", missing)
	}
	if _, err := goportReadtestEvidence(anchors{}); err == nil {
		t.Fatal("goportReadtestEvidence accepted a zero executed-case anchor")
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
	path := filepath.Join(t.TempDir(), "verification_sentinels.json")
	pin := `{"mixed_format_ffi_routing_sentinel_rows": ["row a", "row b"]}`
	if err := os.WriteFile(path, []byte(pin), 0o644); err != nil {
		t.Fatalf("write sentinel pin fixture: %v", err)
	}
	deviations := uint64(2)
	required, err := nativeFFIEvidence(anchors{FFIBitcompareCasesTotal: 10, FFIQuantumSteeringDeviations: &deviations}, path)
	if err != nil {
		t.Fatalf("nativeFFIEvidence: %v", err)
	}
	valid := "--- PASS: TestGeneratedFFIBitCompareSubset (1.00s)\n" +
		"testlogcompact: suppressed 20 subtest lifecycle lines (run=10 pass=10 skip=0) for TestGeneratedFFIBitCompareSubset\n" +
		"--- PASS: TestGeneratedMixedFormatFFIRoutingSentinels (0.01s)\n" +
		"    x_test.go:1: mixed-format FFI routing sentinels: 2/2\n" +
		"--- PASS: TestGeneratedFFIQuantumSteeringAdjudicationStrength (0.01s)\n" +
		"--- PASS: TestGeneratedFFIQuantumSteeringSamples (0.01s)\n" +
		"    x_test.go:1: FFI comparisons: C exact=8 independent INTEL-BID-006=2 total=10/10\n"
	if missing := missingEvidence(strings.Split(valid, "\n"), required); len(missing) != 0 {
		t.Fatalf("valid compact native FFI evidence missing=%v", missing)
	}
	for _, line := range strings.Split(strings.TrimSpace(valid), "\n") {
		if missing := missingEvidence(strings.Split(strings.Replace(valid, line+"\n", "", 1), "\n"), required); len(missing) == 0 {
			t.Fatalf("accepted missing FFI evidence line %q", line)
		}
	}
	for _, wrong := range []string{
		"C exact=10 independent INTEL-BID-006=0 total=10/10",
		"C exact=8 independent INTEL-BID-006=1 total=9/10",
		"C exact=7 independent INTEL-BID-006=3 total=10/10",
	} {
		mutated := strings.Replace(valid, "C exact=8 independent INTEL-BID-006=2 total=10/10", wrong, 1)
		if missing := missingEvidence(strings.Split(mutated, "\n"), required); len(missing) == 0 {
			t.Fatalf("accepted incorrect FFI count evidence %q", wrong)
		}
	}
	for _, n := range []uint64{0, 10, 11} {
		if _, err := nativeFFIEvidence(anchors{FFIBitcompareCasesTotal: 10, FFIQuantumSteeringDeviations: &n}, path); err == nil {
			t.Fatalf("accepted invalid independent FFI count %d", n)
		}
	}
	for _, a := range []anchors{{}, {FFIBitcompareCasesTotal: 10}} {
		if _, err := nativeFFIEvidence(a, path); err == nil {
			t.Fatalf("accepted missing FFI count anchor: %+v", a)
		}
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

func TestD32ExhaustiveDigestEvidenceBindsBothLegsToTheSamePins(t *testing.T) {
	a := anchors{
		D32ExhaustiveLanes:            2,
		D32ExhaustiveCasesPerLane:     4,
		D32ExhaustiveTotalComparisons: 8,
		D32ExhaustiveDigestByLane: map[string]uint64{
			"sqrt_nearest_even": 11,
			"nextup":            22,
		},
	}
	goRequired := d32ExhaustiveDigestEvidence(a, "")
	goLog := strings.Split(
		"    x_test.go:1: decimal32 exhaustive lane nextup: exact comparisons 4/4 digest=22\n"+
			"    x_test.go:1: decimal32 exhaustive lane sqrt_nearest_even: exact comparisons 4/4 digest=11\n"+
			"    x_test.go:1: decimal32 exhaustive unary total comparisons: 8/8\n",
		"\n",
	)
	if missing := missingEvidence(goLog, goRequired); len(missing) != 0 {
		t.Fatalf("valid Go-leg digest evidence missing=%v", missing)
	}
	wrongDigest := strings.Split(
		"    x_test.go:1: decimal32 exhaustive lane nextup: exact comparisons 4/4 digest=23\n"+
			"    x_test.go:1: decimal32 exhaustive lane sqrt_nearest_even: exact comparisons 4/4 digest=11\n"+
			"    x_test.go:1: decimal32 exhaustive unary total comparisons: 8/8\n",
		"\n",
	)
	if missing := missingEvidence(wrongDigest, goRequired); len(missing) != 1 {
		t.Fatalf("moved lane digest still satisfied the pinned evidence: missing=%v", missing)
	}

	rustRequired := d32ExhaustiveDigestEvidence(a, "Rust ")
	if missing := missingEvidence(goLog, rustRequired); len(missing) != len(rustRequired) {
		t.Fatalf("Go-leg log satisfied Rust-leg digest evidence: missing=%v of %d", missing, len(rustRequired))
	}
	rustLog := strings.Split(
		"Rust decimal32 exhaustive lane nextup: exact comparisons 4/4 digest=22\n"+
			"Rust decimal32 exhaustive lane sqrt_nearest_even: exact comparisons 4/4 digest=11\n"+
			"Rust decimal32 exhaustive unary total comparisons: 8/8\n",
		"\n",
	)
	if missing := missingEvidence(rustLog, rustRequired); len(missing) != 0 {
		t.Fatalf("valid Rust-leg digest evidence missing=%v", missing)
	}
	// Reverse direction: the Go leg's required lines are a literal substring
	// of the Rust leg's, so without an explicit prefix rejection a Rust log
	// would satisfy the Go domain's digest evidence on its own.
	if missing := missingEvidence(rustLog, goRequired); len(missing) != len(goRequired) {
		t.Fatalf("Rust-leg log satisfied Go-leg digest evidence: missing=%v of %d", missing, len(goRequired))
	}
	// The rejection is evaluated per line, so a Rust-format line elsewhere in
	// the log must not stop a genuine Go-format line from counting.
	mixed := strings.Split(
		"Rust decimal32 exhaustive lane nextup: exact comparisons 4/4 digest=22\n"+
			"    x_test.go:1: decimal32 exhaustive lane nextup: exact comparisons 4/4 digest=22\n",
		"\n",
	)
	if missing := missingEvidence(mixed, goRequired[:1]); len(missing) != 0 {
		t.Fatalf("Go-format line was rejected because a Rust-format line was present: missing=%v", missing)
	}
	sharded := strings.Split(
		"Rust decimal32 exhaustive lane nextup: exact comparisons 2/4 (sharded run; lane digest suppressed)\n"+
			"Rust decimal32 exhaustive lane sqrt_nearest_even: exact comparisons 2/4 (sharded run; lane digest suppressed)\n"+
			"Rust decimal32 exhaustive unary total comparisons: 4/8\n",
		"\n",
	)
	if missing := missingEvidence(sharded, rustRequired); len(missing) != len(rustRequired) {
		t.Fatalf("sharded Rust-leg log satisfied full-run digest evidence: missing=%v of %d", missing, len(rustRequired))
	}
}

func TestD32ExhaustiveSentinelCountEvidenceSeparatesTheLegs(t *testing.T) {
	path := filepath.Join(t.TempDir(), "verification_sentinels.json")
	pin := `{"d32_exhaustive_sentinel_rows": ["row a", "row b"]}`
	if err := os.WriteFile(path, []byte(pin), 0o644); err != nil {
		t.Fatalf("write sentinel pin fixture: %v", err)
	}
	goWant := d32ExhaustiveSentinelCountEvidence(path, "d32 exhaustive routing sentinels")
	rustWant := d32ExhaustiveSentinelCountEvidence(path, "Rust d32 exhaustive routing sentinels")

	goLog := strings.Split("    x_test.go:1: d32 exhaustive routing sentinels: 2/2\n", "\n")
	rustLog := strings.Split("Rust d32 exhaustive routing sentinels: 2/2\n", "\n")

	if missing := missingEvidence(goLog, []evidence{goWant}); len(missing) != 0 {
		t.Fatalf("Go sentinel count line was not accepted: missing=%v", missing)
	}
	if missing := missingEvidence(rustLog, []evidence{rustWant}); len(missing) != 0 {
		t.Fatalf("Rust sentinel count line was not accepted: missing=%v", missing)
	}
	// The Go literal is a substring of the Rust line, so without the prefix
	// rejection a Rust-only log would satisfy the Go domain's sentinel row.
	if missing := missingEvidence(rustLog, []evidence{goWant}); len(missing) != 1 {
		t.Fatalf("Rust sentinel line satisfied the Go-leg sentinel evidence: missing=%v", missing)
	}
	if missing := missingEvidence(goLog, []evidence{rustWant}); len(missing) != 1 {
		t.Fatalf("Go sentinel line satisfied the Rust-leg sentinel evidence: missing=%v", missing)
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
