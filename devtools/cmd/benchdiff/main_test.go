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
		"BENCH-META target=bench-bidgo count=5 go=go1.26.5 tree=synthetic date=2026-07-13T00:00:00Z",
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
	rows, failed := compareBenchmarks(baseline, candidate, 8, defaultRegressionMinDeltaNs)
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
	rows, failed := compareBenchmarks(baseline, candidate, 8, defaultRegressionMinDeltaNs)
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
	rows, failed := compareBenchmarks(baseline, candidate, 8, defaultRegressionMinDeltaNs)
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
	rows, failed := compareBenchmarks(baseline, candidate, 8, defaultRegressionMinDeltaNs)
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
	rows, failed := compareBenchmarks(baseline, candidate, 8, defaultRegressionMinDeltaNs)
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
	rows, failed := compareBenchmarks(baseline, candidate, 8, defaultRegressionMinDeltaNs)
	if failed {
		t.Fatalf("compareBenchmarks failed although the median is unchanged: %+v", rows)
	}
	if rows[0].changePct != 0 {
		t.Fatalf("changePct = %v, want 0 (median comparison)", rows[0].changePct)
	}
}

func TestCompareDoesNotFailSubNanosecondRowWithinMinDelta(t *testing.T) {
	// The inline-budget canary row BenchmarkAlignedBID128/from_int64-10 moved
	// 0.6905 → 0.8739 ns/op on the Apple M1 reference machine: delta 0.1834 ns,
	// +26.56%. The percentage half of the gate is meaningless at that scale, so
	// the absolute floor must hold the row — but the row must stay visibly
	// distinct from an unchanged "ok" row.
	//
	// The raw `go test -bench` log of that run is not retained in the tree (it
	// exists only in a local bench_watch working record), so these two medians
	// are the transcribed values, not a checked-in fixture.
	baseline := mustParse(t, benchLog("BenchmarkAlignedBID128/from_int64-10 \t1000000000\t0.6905 ns/op"))
	candidate := mustParse(t, benchLog("BenchmarkAlignedBID128/from_int64-10 \t1000000000\t0.8739 ns/op"))
	rows, failed := compareBenchmarks(baseline, candidate, 8, defaultRegressionMinDeltaNs)
	if failed {
		t.Fatalf("compareBenchmarks failed on a 0.1834 ns sub-nanosecond delta: %+v", rows)
	}
	if len(rows) != 1 || rows[0].status != statusBelowMinDelta {
		t.Fatalf("rows = %+v, want one %q row", rows, statusBelowMinDelta)
	}
	if rows[0].changePct < 26.5 || rows[0].changePct > 26.6 {
		t.Fatalf("changePct = %v, want ~26.56 (still reported, just not failing)", rows[0].changePct)
	}
}

func TestCompareHoldsRowsBelowTheSensitivityBoundaryToTheFloor(t *testing.T) {
	// The floor is global, not a sub-nanosecond special case: it binds on every
	// row whose baseline median is under 3.125 ns. On the committed baselines
	// that is 12 of 259 rows — 2 sub-nanosecond rows and a 10-row 2.20–2.55 ns
	// band. This row sits in that band (BenchmarkFairBID64/to_decimal128-10,
	// median 2.217 ns): +9.16% clears the 8% threshold while the 0.203 ns delta
	// stays under the floor, so it must be held, not failed.
	baseline := mustParse(t, benchLog("BenchmarkFairBID64/to_decimal128-10 \t1000\t2.217 ns/op"))
	candidate := mustParse(t, benchLog("BenchmarkFairBID64/to_decimal128-10 \t1000\t2.420 ns/op"))
	rows, failed := compareBenchmarks(baseline, candidate, 8, defaultRegressionMinDeltaNs)
	if failed {
		t.Fatalf("compareBenchmarks failed on a 0.203 ns delta in the 2.20–2.55 ns band: %+v", rows)
	}
	if len(rows) != 1 || rows[0].status != statusBelowMinDelta {
		t.Fatalf("rows = %+v, want one %q row", rows, statusBelowMinDelta)
	}
	if rows[0].changePct <= 8 {
		t.Fatalf("changePct = %v, want > 8 (the row must clear the percentage threshold, so only the floor holds it)", rows[0].changePct)
	}
}

func TestCompareTreatsDeltaExactlyAtTheFloorAsBelowIt(t *testing.T) {
	// The floor comparison is strict (`delta > minDeltaNs`), mirroring the
	// strict percentage comparison. 2.0 → 2.25 is +12.5% on a delta of exactly
	// 0.25 ns — both values and the floor are exactly representable, so this
	// pins the boundary itself rather than a value near it.
	baseline := mustParse(t, benchLog("BenchmarkX-10 \t1000\t2.00 ns/op"))
	candidate := mustParse(t, benchLog("BenchmarkX-10 \t1000\t2.25 ns/op"))
	rows, failed := compareBenchmarks(baseline, candidate, 8, defaultRegressionMinDeltaNs)
	if failed {
		t.Fatalf("compareBenchmarks failed on a delta exactly at the floor: %+v", rows)
	}
	if len(rows) != 1 || rows[0].status != statusBelowMinDelta {
		t.Fatalf("rows = %+v, want one %q row (delta == floor is not above it)", rows, statusBelowMinDelta)
	}
	if rows[0].changePct != 12.5 {
		t.Fatalf("changePct = %v, want exactly 12.5", rows[0].changePct)
	}
}

func TestCompareFailsWhenBothThresholdAndMinDeltaAreExceeded(t *testing.T) {
	for _, tc := range []struct {
		name             string
		baseNs, candNs   string
		wantChangePctMin float64
	}{
		// 8% of 3.125 ns is exactly 0.25 ns, so 3.125 ns is the median at
		// which the two rules meet: at or above it, every row that clears the
		// percentage threshold also clears the floor and the 8% sensitivity is
		// unchanged. 3.125 → 3.38 is +8.16% on a 0.255 ns delta.
		{"at_the_sensitivity_boundary", "3.125", "3.380", 8.1},
		// The second REGRESSION row of the same real failing run
		// (BenchmarkIntelCBID128/minnum-10, 7.04 → 8.26): a 1.22 ns delta is a
		// real regression and the floor must not rescue it.
		{"well_above_the_floor", "7.040", "8.260", 17.3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			baseline := mustParse(t, benchLog("BenchmarkX-10 \t1000\t"+tc.baseNs+" ns/op"))
			candidate := mustParse(t, benchLog("BenchmarkX-10 \t1000\t"+tc.candNs+" ns/op"))
			rows, failed := compareBenchmarks(baseline, candidate, 8, defaultRegressionMinDeltaNs)
			if !failed {
				t.Fatalf("compareBenchmarks passed although both the threshold and the floor were exceeded: %+v", rows)
			}
			if len(rows) != 1 || rows[0].status != statusRegression {
				t.Fatalf("rows = %+v, want one regression row", rows)
			}
			if rows[0].changePct < tc.wantChangePctMin {
				t.Fatalf("changePct = %v, want >= %v", rows[0].changePct, tc.wantChangePctMin)
			}
		})
	}
}

func TestCompareWithZeroMinDeltaRestoresPurePercentageGate(t *testing.T) {
	// BENCH_REGRESSION_MIN_DELTA_NS=0 disables the floor: the same 0.1834 ns
	// sub-nanosecond move that passes under the default must fail again, so
	// the pre-floor behaviour stays reachable for deliberate sub-ns work.
	baseline := mustParse(t, benchLog("BenchmarkAlignedBID128/from_int64-10 \t1000000000\t0.6905 ns/op"))
	candidate := mustParse(t, benchLog("BenchmarkAlignedBID128/from_int64-10 \t1000000000\t0.8739 ns/op"))
	rows, failed := compareBenchmarks(baseline, candidate, 8, 0)
	if !failed {
		t.Fatalf("compareBenchmarks passed with the floor disabled: %+v", rows)
	}
	if len(rows) != 1 || rows[0].status != statusRegression {
		t.Fatalf("rows = %+v, want one regression row", rows)
	}
}

// reportRowStatus returns the trailing status cell of the printReport row for
// the named benchmark. Status texts contain spaces ("ok (below min delta)"), so
// the cell is taken as everything after the change column — the one field
// ending in "%" — rather than as the last whitespace-separated field.
func reportRowStatus(t *testing.T, report, name string) string {
	t.Helper()
	for _, line := range strings.Split(report, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 || fields[0] != name {
			continue
		}
		for i, field := range fields[1:] {
			if strings.HasSuffix(field, "%") {
				return strings.Join(fields[i+2:], " ")
			}
		}
		t.Fatalf("row %q has no change column: %q", name, line)
	}
	t.Fatalf("report has no row for %q:\n%s", name, report)
	return ""
}

func TestPrintReportLabelsBelowFloorRowDistinctlyFromUnchangedRow(t *testing.T) {
	// The floor must not launder a threshold-clearing row into a plain "ok".
	// Both kinds of row go into one report so the two labels are pinned against
	// each other, and the status is read as a parsed cell rather than inferred
	// from tabwriter padding.
	baseline := mustParse(t, benchLog(
		"BenchmarkAlignedBID128/from_int64-10 \t1000000000\t0.6905 ns/op",
		"BenchmarkFairBID64/add-10 \t1000\t51.00 ns/op",
	))
	candidate := mustParse(t, benchLog(
		"BenchmarkAlignedBID128/from_int64-10 \t1000000000\t0.8739 ns/op",
		"BenchmarkFairBID64/add-10 \t1000\t51.00 ns/op",
	))
	rows, failed := compareBenchmarks(baseline, candidate, 8, defaultRegressionMinDeltaNs)
	if failed {
		t.Fatalf("compareBenchmarks failed: %+v", rows)
	}
	var out strings.Builder
	printReport(&out, rows, "base.txt", "cand.txt", 8, defaultRegressionMinDeltaNs, failed)
	got := out.String()

	// The expected statuses are spelled as literals, not as string(statusFoo).
	// Routing them through the constants would make this table agree with any
	// text the constant happens to hold, including a statusBelowMinDelta
	// redefined to "ok" — which is exactly the collapse the package doc forbids
	// ("never as a plain \"ok\"") and which the bench_watch summary grep for
	// "below min delta" depends on. The literal text is part of the contract.
	for _, tc := range []struct{ name, want string }{
		{"BenchmarkAlignedBID128/from_int64-10", "ok (below min delta)"},
		{"BenchmarkFairBID64/add-10", "ok"},
	} {
		if status := reportRowStatus(t, got, tc.name); status != tc.want {
			t.Fatalf("row %s status = %q, want %q:\n%s", tc.name, status, tc.want, got)
		}
	}
	// The header must name the floor in force, and the held row must keep its
	// change% so a reader can judge the move the floor suppressed a verdict on.
	for _, want := range []string{"min-delta=+0.25ns", "+26.56%", "benchdiff: PASS"} {
		if !strings.Contains(got, want) {
			t.Fatalf("report is missing %q:\n%s", want, got)
		}
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

func TestRegressionMinDeltaNs(t *testing.T) {
	if got, err := regressionMinDeltaNs(""); err != nil || got != defaultRegressionMinDeltaNs {
		t.Fatalf("unset min delta = (%v, %v), want (%v, nil)", got, err, defaultRegressionMinDeltaNs)
	}
	if got, err := regressionMinDeltaNs("0.5"); err != nil || got != 0.5 {
		t.Fatalf("min delta 0.5 = (%v, %v), want (0.5, nil)", got, err)
	}
	// Unlike the threshold, 0 is a legal setting: it disables the floor and
	// restores the pure percentage gate.
	if got, err := regressionMinDeltaNs("0"); err != nil || got != 0 {
		t.Fatalf("min delta 0 = (%v, %v), want (0, nil)", got, err)
	}
	if _, err := regressionMinDeltaNs("quarter"); err == nil {
		t.Fatal("non-numeric min delta accepted, want error")
	}
	// ParseFloat accepts NaN/Inf spellings and a plain `< 0` comparison passes
	// them; every non-finite or negative floor must be an input error, not a
	// silently disabled or distorted gate.
	for _, raw := range []string{"-1", "-0.25", "NaN", "nan", "Inf", "+Inf", "-Inf"} {
		if _, err := regressionMinDeltaNs(raw); err == nil {
			t.Fatalf("min delta %q accepted, want error", raw)
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
	if len(meta.toolchains) != 1 || meta.toolchains[0] != "go1.26.5" {
		t.Fatalf("toolchains = %v, want [go1.26.5]", meta.toolchains)
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
		"BENCH-META target=bench-bidgo count=1 go=go1.26.5 tree=synthetic date=2026-07-13T00:00:00Z",
		"goos: darwin", "goarch: arm64", "cpu: Apple M1",
	}, "BenchmarkX-10 \t1\t1.0 ns/op"))
	err := requireComparableMeta(baseline, candidate)
	if err == nil || !strings.Contains(err.Error(), "BENCH-META count") {
		t.Fatalf("count mismatch err = %v, want BENCH-META count mismatch", err)
	}
	// The remediation hint belongs to the go= item only. Leaking it onto another
	// comparability item would tell the reader to re-measure for a toolchain
	// change that did not happen, mis-diagnosing the actual mismatch.
	if strings.Contains(err.Error(), "toolchain provenance mismatch") {
		t.Fatalf("count mismatch err carries the toolchain hint: %v", err)
	}
}

func TestRequireComparableMetaFailsOnEnvironmentMismatch(t *testing.T) {
	baseline := mustParseMeta(t, benchLog("BenchmarkX-10 \t1\t1.0 ns/op"))
	for _, tc := range []struct {
		name string
		meta []string
	}{
		{"goos", []string{
			"BENCH-META target=bench-bidgo count=5 go=go1.26.5 tree=synthetic date=2026-07-13T00:00:00Z",
			"goos: linux", "goarch: arm64", "cpu: Apple M1",
		}},
		{"goarch", []string{
			"BENCH-META target=bench-bidgo count=5 go=go1.26.5 tree=synthetic date=2026-07-13T00:00:00Z",
			"goos: darwin", "goarch: amd64", "cpu: Apple M1",
		}},
		{"cpu", []string{
			"BENCH-META target=bench-bidgo count=5 go=go1.26.5 tree=synthetic date=2026-07-13T00:00:00Z",
			"goos: darwin", "goarch: arm64", "cpu: Apple M4",
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			candidate := mustParseMeta(t, benchLogWithMeta(tc.meta, "BenchmarkX-10 \t1\t1.0 ns/op"))
			err := requireComparableMeta(baseline, candidate)
			if err == nil || !strings.Contains(err.Error(), tc.name) {
				t.Fatalf("%s mismatch err = %v, want %s mismatch", tc.name, err, tc.name)
			}
			// Same containment rule as the count item: the go= remediation hint
			// must not surface on an environment mismatch and send the reader
			// re-measuring for a toolchain change that did not happen.
			if strings.Contains(err.Error(), "toolchain provenance mismatch") {
				t.Fatalf("%s mismatch err carries the toolchain hint: %v", tc.name, err)
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
		"BENCH-META target=bench-bidgo count=5 go=go1.26.5 tree=synthetic date=2026-07-13T00:00:00Z",
		"goos: linux", "goarch: arm64",
	}
	baseline := mustParseMeta(t, benchLogWithMeta(meta, "BenchmarkX-10 \t1\t1.0 ns/op"))
	candidate := mustParseMeta(t, benchLogWithMeta(meta, "BenchmarkX-10 \t1\t2.0 ns/op"))
	if err := requireComparableMeta(baseline, candidate); err != nil {
		t.Fatalf("cpu-less but otherwise identical runs rejected: %v", err)
	}
}

// benchLogToolchainMeta builds the standard synthetic metadata block with a
// caller-chosen go= token, or with the token omitted entirely when goToken is
// empty — the shape of every baseline saved before toolchain recording existed.
func benchLogToolchainMeta(goToken string) []string {
	meta := "BENCH-META target=bench-bidgo count=5"
	if goToken != "" {
		meta += " go=" + goToken
	}
	meta += " tree=synthetic date=2026-07-13T00:00:00Z"
	return []string{meta, "goos: darwin", "goarch: arm64", "cpu: Apple M1"}
}

func TestRequireComparableMetaPassesOnMatchingToolchain(t *testing.T) {
	baseline := mustParseMeta(t, benchLogWithMeta(benchLogToolchainMeta("go1.26.5"), "BenchmarkX-10 \t1\t1.0 ns/op"))
	candidate := mustParseMeta(t, benchLogWithMeta(benchLogToolchainMeta("go1.26.5"), "BenchmarkX-10 \t1\t2.0 ns/op"))
	if err := requireComparableMeta(baseline, candidate); err != nil {
		t.Fatalf("runs on the same toolchain rejected: %v", err)
	}
}

func TestRequireComparableMetaFailsWhenBaselinePredatesToolchainRecording(t *testing.T) {
	// The concrete situation this token exists for: the saved baselines were
	// captured before go= was emitted, then the host Go toolchain was upgraded.
	// Without the token the comparison silently spanned two toolchains; with it
	// the pairing is "(none)" vs "go1.26.5" and must be an input error carrying
	// the re-measure instruction.
	baseline := mustParseMeta(t, benchLogWithMeta(benchLogToolchainMeta(""), "BenchmarkX-10 \t1\t1.0 ns/op"))
	candidate := mustParseMeta(t, benchLogWithMeta(benchLogToolchainMeta("go1.26.5"), "BenchmarkX-10 \t1\t1.0 ns/op"))
	err := requireComparableMeta(baseline, candidate)
	if err == nil {
		t.Fatal("pre-token baseline compared against a toolchain-recording candidate, want error")
	}
	for _, want := range []string{
		"BENCH-META go", "(none)", "go1.26.5",
		"toolchain provenance mismatch",
		"re-measure and re-save the baseline",
		"make bench-go-baseline",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("toolchain mismatch err = %v, missing %q", err, want)
		}
	}
}

func TestRequireComparableMetaAllowsToolchainAbsentOnBothSides(t *testing.T) {
	// Two logs that both predate the token still describe one toolchain era, so
	// they remain comparable; adding the token must not retroactively break
	// baseline/candidate pairs captured before it existed.
	baseline := mustParseMeta(t, benchLogWithMeta(benchLogToolchainMeta(""), "BenchmarkX-10 \t1\t1.0 ns/op"))
	candidate := mustParseMeta(t, benchLogWithMeta(benchLogToolchainMeta(""), "BenchmarkX-10 \t1\t2.0 ns/op"))
	if err := requireComparableMeta(baseline, candidate); err != nil {
		t.Fatalf("two pre-token logs rejected: %v", err)
	}
}

func TestRequireComparableMetaFailsOnToolchainVersionMismatch(t *testing.T) {
	baseline := mustParseMeta(t, benchLogWithMeta(benchLogToolchainMeta("go1.26.1"), "BenchmarkX-10 \t1\t1.0 ns/op"))
	candidate := mustParseMeta(t, benchLogWithMeta(benchLogToolchainMeta("go1.26.5"), "BenchmarkX-10 \t1\t1.0 ns/op"))
	err := requireComparableMeta(baseline, candidate)
	if err == nil {
		t.Fatal("medians from two different Go toolchains compared, want error")
	}
	for _, want := range []string{"BENCH-META go", "go1.26.1", "go1.26.5", "toolchain provenance mismatch"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("toolchain version mismatch err = %v, missing %q", err, want)
		}
	}
}

func TestRequireComparableMetaFailsOnCountlessBenchMetaLine(t *testing.T) {
	baseline := mustParseMeta(t, benchLog("BenchmarkX-10 \t1\t1.0 ns/op"))
	candidate := mustParseMeta(t, benchLogWithMeta([]string{
		"BENCH-META target=bench-bidgo go=go1.26.5 tree=synthetic date=2026-07-13T00:00:00Z",
		"goos: darwin", "goarch: arm64", "cpu: Apple M1",
	}, "BenchmarkX-10 \t1\t1.0 ns/op"))
	err := requireComparableMeta(baseline, candidate)
	if err == nil || !strings.Contains(err.Error(), "BENCH-META count") {
		t.Fatalf("countless BENCH-META err = %v, want BENCH-META count mismatch", err)
	}
}
