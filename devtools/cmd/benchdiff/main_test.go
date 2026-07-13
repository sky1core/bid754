package main

import (
	"strings"
	"testing"
)

// benchLog builds a synthetic `go test -bench` log with the surrounding
// non-benchmark lines a real run emits, so the parser is exercised against
// the production log shape rather than a bare row list.
func benchLog(rows ...string) string {
	lines := []string{
		"BENCH-META target=bench-bidgo count=5 tree=synthetic date=2026-07-13T00:00:00Z",
		"goos: darwin",
		"goarch: arm64",
		"pkg: github.com/sky1core/bid754/bid754-go/internal/bidgo",
		"cpu: Apple M1",
	}
	lines = append(lines, rows...)
	lines = append(lines, "PASS", "ok  \tgithub.com/sky1core/bid754/bid754-go/internal/bidgo\t10.0s")
	return strings.Join(lines, "\n") + "\n"
}

func mustParse(t *testing.T, log string) map[string][]float64 {
	t.Helper()
	samples, err := parseBenchLog(strings.NewReader(log))
	if err != nil {
		t.Fatalf("parseBenchLog: %v", err)
	}
	return samples
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
	if _, err := regressionThresholdPct("-1"); err == nil {
		t.Fatal("negative threshold accepted, want error")
	}
}

func TestParseBenchLogRejectsNonPositiveNsPerOp(t *testing.T) {
	if _, err := parseBenchLog(strings.NewReader("BenchmarkX-10 \t1\t0 ns/op\n")); err == nil {
		t.Fatal("zero ns/op accepted, want error")
	}
}
