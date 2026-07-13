package main

import (
	"strings"
	"testing"
)

// benchLog builds a synthetic `go test -bench` log with the surrounding
// non-benchmark lines a real run emits, so the parser is exercised against
// the production log shape rather than a bare row list.
func benchLog(rows ...string) string {
	return benchLogWithMeta([]string{
		"BENCH-META target=bench-bidgo count=5 tree=synthetic date=2026-07-13T00:00:00Z",
		"goos: darwin",
		"goarch: arm64",
		"pkg: github.com/sky1core/bid754/bid754-go/internal/bidgo",
		"cpu: Apple M1",
	}, rows...)
}

// benchLogWithMeta builds the same synthetic log shape with caller-chosen
// leading metadata lines, for the comparability-mismatch tests.
func benchLogWithMeta(metaLines []string, rows ...string) string {
	lines := append([]string(nil), metaLines...)
	lines = append(lines, rows...)
	lines = append(lines, "PASS", "ok  \tgithub.com/sky1core/bid754/bid754-go/internal/bidgo\t10.0s")
	return strings.Join(lines, "\n") + "\n"
}

func mustParse(t *testing.T, log string) map[string][]float64 {
	t.Helper()
	samples, _, err := parseBenchLog(strings.NewReader(log))
	if err != nil {
		t.Fatalf("parseBenchLog: %v", err)
	}
	return samples
}

func mustParseMeta(t *testing.T, log string) benchLogMeta {
	t.Helper()
	_, meta, err := parseBenchLog(strings.NewReader(log))
	if err != nil {
		t.Fatalf("parseBenchLog: %v", err)
	}
	return meta
}

func TestParseBenchLogCollectsRepeatedSamplesAndSkipsNoise(t *testing.T) {
	log := benchLog(
		"BenchmarkFairBID64/add-10 \t23305646\t51.00 ns/op\t0 B/op\t0 allocs/op",
		"BenchmarkFairBID64/add-10 \t23305646\t53.00 ns/op\t0 B/op\t0 allocs/op",
		"BenchmarkFairBID64/sqrt-10 \t133136900\t9.028 ns/op",
	)
	samples := mustParse(t, log)
	if len(samples) != 2 {
		t.Fatalf("parsed %d benchmarks, want 2: %v", len(samples), samples)
	}
	if got := samples["BenchmarkFairBID64/add-10"]; len(got) != 2 || got[0] != 51.00 || got[1] != 53.00 {
		t.Fatalf("add samples = %v, want [51 53]", got)
	}
	if got := samples["BenchmarkFairBID64/sqrt-10"]; len(got) != 1 || got[0] != 9.028 {
		t.Fatalf("sqrt samples = %v, want [9.028]", got)
	}
}

func TestMedianOddCountPicksMiddleSample(t *testing.T) {
	// Unsorted on purpose: median must sort, not trust input order.
	if got := median([]float64{90, 10, 50, 30, 70}); got != 50 {
		t.Fatalf("median(odd) = %v, want 50", got)
	}
}

func TestMedianEvenCountAveragesTwoMiddleSamples(t *testing.T) {
	if got := median([]float64{40, 10, 20, 30}); got != 25 {
		t.Fatalf("median(even) = %v, want 25", got)
	}
}

func TestCompareFailsOnRegressionAboveThreshold(t *testing.T) {
	baseline := mustParse(t, benchLog(
		"BenchmarkFairBID64/fma-10 \t1000\t100.0 ns/op",
		"BenchmarkFairBID64/fma-10 \t1000\t100.0 ns/op",
		"BenchmarkFairBID64/fma-10 \t1000\t100.0 ns/op",
	))
	candidate := mustParse(t, benchLog(
		// Median 110 = +10% > 8% threshold; the 90 outlier must not rescue it.
		"BenchmarkFairBID64/fma-10 \t1000\t110.0 ns/op",
		"BenchmarkFairBID64/fma-10 \t1000\t90.0 ns/op",
		"BenchmarkFairBID64/fma-10 \t1000\t120.0 ns/op",
	))
	rows, failed := compareBenchmarks(baseline, candidate, 8)
	if !failed {
		t.Fatal("compareBenchmarks passed, want regression failure")
	}
	if len(rows) != 1 || rows[0].status != statusRegression {
		t.Fatalf("rows = %+v, want one regression row", rows)
	}
	if rows[0].changePct < 9.99 || rows[0].changePct > 10.01 {
		t.Fatalf("changePct = %v, want ~10", rows[0].changePct)
	}
}

func TestComparePassesOnRegressionAtOrBelowThreshold(t *testing.T) {
	baseline := mustParse(t, benchLog("BenchmarkFairBID64/fma-10 \t1000\t100.0 ns/op"))
	candidate := mustParse(t, benchLog("BenchmarkFairBID64/fma-10 \t1000\t108.0 ns/op"))
	rows, failed := compareBenchmarks(baseline, candidate, 8)
	if failed {
		t.Fatalf("compareBenchmarks failed on +8%% at threshold 8%%: %+v", rows)
	}
	if len(rows) != 1 || rows[0].status != statusOK {
		t.Fatalf("rows = %+v, want one ok row", rows)
	}
}

func TestCompareReportsImprovementWithoutFailing(t *testing.T) {
	baseline := mustParse(t, benchLog("BenchmarkFairBID64/fma-10 \t1000\t100.0 ns/op"))
	candidate := mustParse(t, benchLog("BenchmarkFairBID64/fma-10 \t1000\t80.0 ns/op"))
	rows, failed := compareBenchmarks(baseline, candidate, 8)
	if failed {
		t.Fatalf("compareBenchmarks failed on an improvement: %+v", rows)
	}
	if len(rows) != 1 || rows[0].status != statusImproved {
		t.Fatalf("rows = %+v, want one improved row", rows)
	}
}

func TestCompareFailsWhenBaselineBenchmarkVanishes(t *testing.T) {
	baseline := mustParse(t, benchLog(
		"BenchmarkFairBID64/add-10 \t1000\t50.0 ns/op",
		"BenchmarkFairBID64/fma-10 \t1000\t100.0 ns/op",
	))
	candidate := mustParse(t, benchLog(
		"BenchmarkFairBID64/add-10 \t1000\t50.0 ns/op",
	))
	rows, failed := compareBenchmarks(baseline, candidate, 8)
	if !failed {
		t.Fatal("compareBenchmarks passed although a baseline benchmark vanished")
	}
	var missing *diffRow
	for i := range rows {
		if rows[i].name == "BenchmarkFairBID64/fma-10" {
			missing = &rows[i]
		}
	}
	if missing == nil || missing.status != statusMissing {
		t.Fatalf("rows = %+v, want missing row for vanished fma benchmark", rows)
	}
}

func TestCompareReportsCandidateOnlyBenchmarkAsNewWithoutFailing(t *testing.T) {
	baseline := mustParse(t, benchLog("BenchmarkFairBID64/add-10 \t1000\t50.0 ns/op"))
	candidate := mustParse(t, benchLog(
		"BenchmarkFairBID64/add-10 \t1000\t50.0 ns/op",
		"BenchmarkFairBID64/sqrt-10 \t1000\t9.0 ns/op",
	))
	rows, failed := compareBenchmarks(baseline, candidate, 8)
	if failed {
		t.Fatalf("compareBenchmarks failed on a candidate-only benchmark: %+v", rows)
	}
	var added *diffRow
	for i := range rows {
		if rows[i].name == "BenchmarkFairBID64/sqrt-10" {
			added = &rows[i]
		}
	}
	if added == nil || added.status != statusNew {
		t.Fatalf("rows = %+v, want new-row status for candidate-only benchmark", rows)
	}
}

func TestCompareUsesMedianNotMeanAcrossRepeatedSamples(t *testing.T) {
	// Mean of the candidate samples is 140 (+40%), but the median is 100
	// (unchanged): a single noisy outlier must not fail the gate.
	baseline := mustParse(t, benchLog(
		"BenchmarkFairBID64/fma-10 \t1000\t100.0 ns/op",
		"BenchmarkFairBID64/fma-10 \t1000\t100.0 ns/op",
		"BenchmarkFairBID64/fma-10 \t1000\t100.0 ns/op",
	))
	candidate := mustParse(t, benchLog(
		"BenchmarkFairBID64/fma-10 \t1000\t100.0 ns/op",
		"BenchmarkFairBID64/fma-10 \t1000\t220.0 ns/op",
		"BenchmarkFairBID64/fma-10 \t1000\t100.0 ns/op",
	))
	rows, failed := compareBenchmarks(baseline, candidate, 8)
	if failed {
		t.Fatalf("compareBenchmarks failed although the median is unchanged: %+v", rows)
	}
	if rows[0].changePct != 0 {
		t.Fatalf("changePct = %v, want 0 (median comparison)", rows[0].changePct)
	}
}

func TestRegressionThresholdPct(t *testing.T) {
	if got, err := regressionThresholdPct(""); err != nil || got != defaultRegressionThresholdPct {
		t.Fatalf("unset threshold = (%v, %v), want (%v, nil)", got, err, defaultRegressionThresholdPct)
	}
	if got, err := regressionThresholdPct("12.5"); err != nil || got != 12.5 {
		t.Fatalf("threshold 12.5 = (%v, %v), want (12.5, nil)", got, err)
	}
	if _, err := regressionThresholdPct("fast"); err == nil {
		t.Fatal("non-numeric threshold accepted, want error")
	}
	// ParseFloat accepts NaN/Inf spellings and a plain `< 0` comparison
	// passes them; every non-finite or non-positive threshold must be an
	// input error, not a silently disabled gate.
	for _, raw := range []string{"-1", "0", "NaN", "nan", "Inf", "+Inf", "-Inf"} {
		if _, err := regressionThresholdPct(raw); err == nil {
			t.Fatalf("threshold %q accepted, want error", raw)
		}
	}
}

func TestParseBenchLogRejectsNonPositiveOrNonFiniteNsPerOp(t *testing.T) {
	for _, value := range []string{"0", "-3.5", "NaN", "Inf", "+Inf", "-Inf"} {
		log := "BenchmarkX-10 \t1\t" + value + " ns/op\n"
		if _, _, err := parseBenchLog(strings.NewReader(log)); err == nil {
			t.Fatalf("ns/op value %q accepted, want error", value)
		}
	}
}

func TestParseBenchLogCollectsComparabilityMeta(t *testing.T) {
	meta := mustParseMeta(t, benchLog("BenchmarkX-10 \t1\t1.0 ns/op"))
	if len(meta.counts) != 1 || meta.counts[0] != "5" {
		t.Fatalf("counts = %v, want [5]", meta.counts)
	}
	if len(meta.goos) != 1 || meta.goos[0] != "darwin" {
		t.Fatalf("goos = %v, want [darwin]", meta.goos)
	}
	if len(meta.goarch) != 1 || meta.goarch[0] != "arm64" {
		t.Fatalf("goarch = %v, want [arm64]", meta.goarch)
	}
	if len(meta.cpu) != 1 || meta.cpu[0] != "Apple M1" {
		t.Fatalf("cpu = %v, want [Apple M1]", meta.cpu)
	}
}

func TestRequireComparableMetaPassesOnIdenticalRuns(t *testing.T) {
	baseline := mustParseMeta(t, benchLog("BenchmarkX-10 \t1\t1.0 ns/op"))
	candidate := mustParseMeta(t, benchLog("BenchmarkX-10 \t1\t2.0 ns/op"))
	if err := requireComparableMeta(baseline, candidate); err != nil {
		t.Fatalf("identical run metadata rejected: %v", err)
	}
}

func TestRequireComparableMetaFailsOnCountMismatch(t *testing.T) {
	baseline := mustParseMeta(t, benchLog("BenchmarkX-10 \t1\t1.0 ns/op"))
	candidate := mustParseMeta(t, benchLogWithMeta([]string{
		"BENCH-META target=bench-bidgo count=1 tree=synthetic date=2026-07-13T00:00:00Z",
		"goos: darwin", "goarch: arm64", "cpu: Apple M1",
	}, "BenchmarkX-10 \t1\t1.0 ns/op"))
	err := requireComparableMeta(baseline, candidate)
	if err == nil || !strings.Contains(err.Error(), "BENCH-META count") {
		t.Fatalf("count mismatch err = %v, want BENCH-META count mismatch", err)
	}
}

func TestRequireComparableMetaFailsOnEnvironmentMismatch(t *testing.T) {
	baseline := mustParseMeta(t, benchLog("BenchmarkX-10 \t1\t1.0 ns/op"))
	for _, tc := range []struct {
		name string
		meta []string
	}{
		{"goos", []string{
			"BENCH-META target=bench-bidgo count=5 tree=synthetic date=2026-07-13T00:00:00Z",
			"goos: linux", "goarch: arm64", "cpu: Apple M1",
		}},
		{"goarch", []string{
			"BENCH-META target=bench-bidgo count=5 tree=synthetic date=2026-07-13T00:00:00Z",
			"goos: darwin", "goarch: amd64", "cpu: Apple M1",
		}},
		{"cpu", []string{
			"BENCH-META target=bench-bidgo count=5 tree=synthetic date=2026-07-13T00:00:00Z",
			"goos: darwin", "goarch: arm64", "cpu: Apple M4",
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			candidate := mustParseMeta(t, benchLogWithMeta(tc.meta, "BenchmarkX-10 \t1\t1.0 ns/op"))
			err := requireComparableMeta(baseline, candidate)
			if err == nil || !strings.Contains(err.Error(), tc.name) {
				t.Fatalf("%s mismatch err = %v, want %s mismatch", tc.name, err, tc.name)
			}
		})
	}
}

func TestRequireComparableMetaFailsWhenOneSideLacksMeta(t *testing.T) {
	baseline := mustParseMeta(t, benchLog("BenchmarkX-10 \t1\t1.0 ns/op"))
	candidate := mustParseMeta(t, benchLogWithMeta(nil, "BenchmarkX-10 \t1\t1.0 ns/op"))
	if err := requireComparableMeta(baseline, candidate); err == nil {
		t.Fatal("metadata-less candidate paired with a metadata-carrying baseline, want error")
	}
}

func TestRequireComparableMetaFailsWhenBothSidesLackMeta(t *testing.T) {
	// Stripping the header lines from BOTH logs must not bypass the guard:
	// empty metadata proves nothing about comparability (a cross-machine
	// comparison could otherwise be laundered by filtering the headers out).
	baseline := mustParseMeta(t, benchLogWithMeta(nil, "BenchmarkX-10 \t1\t1.0 ns/op"))
	candidate := mustParseMeta(t, benchLogWithMeta(nil, "BenchmarkX-10 \t1\t1.0 ns/op"))
	if err := requireComparableMeta(baseline, candidate); err == nil {
		t.Fatal("two metadata-less logs compared as equal, want missing-metadata error")
	}
}

func TestRequireComparableMetaAllowsCpuAbsentOnBothSides(t *testing.T) {
	// Go omits the cpu line on platforms it cannot identify; absent-on-both
	// is legitimate, absent-on-one is a mismatch.
	meta := []string{
		"BENCH-META target=bench-bidgo count=5 tree=synthetic date=2026-07-13T00:00:00Z",
		"goos: linux", "goarch: arm64",
	}
	baseline := mustParseMeta(t, benchLogWithMeta(meta, "BenchmarkX-10 \t1\t1.0 ns/op"))
	candidate := mustParseMeta(t, benchLogWithMeta(meta, "BenchmarkX-10 \t1\t2.0 ns/op"))
	if err := requireComparableMeta(baseline, candidate); err != nil {
		t.Fatalf("cpu-less but otherwise identical runs rejected: %v", err)
	}
}

func TestRequireComparableMetaFailsOnCountlessBenchMetaLine(t *testing.T) {
	baseline := mustParseMeta(t, benchLog("BenchmarkX-10 \t1\t1.0 ns/op"))
	candidate := mustParseMeta(t, benchLogWithMeta([]string{
		"BENCH-META target=bench-bidgo tree=synthetic date=2026-07-13T00:00:00Z",
		"goos: darwin", "goarch: arm64", "cpu: Apple M1",
	}, "BenchmarkX-10 \t1\t1.0 ns/op"))
	err := requireComparableMeta(baseline, candidate)
	if err == nil || !strings.Contains(err.Error(), "BENCH-META count") {
		t.Fatalf("countless BENCH-META err = %v, want BENCH-META count mismatch", err)
	}
}
