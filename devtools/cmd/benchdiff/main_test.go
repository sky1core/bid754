package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// benchLog builds a synthetic `go test -bench` log with the surrounding
// non-benchmark lines a real run emits, so the parser is exercised against
// the production log shape rather than a bare row list.
func benchLog(rows ...string) string {
	count := 0
	if len(rows) > 0 {
		name := strings.Fields(rows[0])[0]
		for _, row := range rows {
			fields := strings.Fields(row)
			if len(fields) > 1 && fields[0] == name {
				count++
			}
		}
	}
	return benchLogWithMeta([]string{
		fmt.Sprintf("BENCH-META target=bench-bidgo count=%d go=go1.26.5 tree=synthetic benchtime=1s build=aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa date=2026-07-13T00:00:00Z", count),
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
		"BenchmarkFairBID64/sqrt-10 \t133136900\t9.100 ns/op",
	)
	samples := mustParse(t, log)
	if len(samples) != 2 {
		t.Fatalf("parsed %d benchmarks, want 2: %v", len(samples), samples)
	}
	if got := samples["BenchmarkFairBID64/add-10"]; len(got) != 2 || got[0] != 51.00 || got[1] != 53.00 {
		t.Fatalf("add samples = %v, want [51 53]", got)
	}
	if got := samples["BenchmarkFairBID64/sqrt-10"]; len(got) != 2 || got[0] != 9.028 || got[1] != 9.100 {
		t.Fatalf("sqrt samples = %v, want [9.028 9.100]", got)
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
		log := benchLog("BenchmarkX-10 \t1\t" + value + " ns/op")
		if _, _, err := parseBenchLog(strings.NewReader(log)); err == nil {
			t.Fatalf("ns/op value %q accepted, want error", value)
		}
	}
}

func TestParseBenchLogValidFormats(t *testing.T) {
	row := "BenchmarkX/sub-10 100 1.25 ns/op 0 B/op 0 allocs/op 3 custom/op"
	log := benchLog(row, row, row, row, row)
	for _, tc := range []struct{ name, log string }{
		{"count five", log},
		{"CRLF", strings.ReplaceAll(log, "\n", "\r\n")},
		{"without final newline", strings.TrimSuffix(log, "\n")},
		{"rounded elapsed time", strings.Replace(log, "10.0s", "0.000s", 1)},
		{"without cpu", strings.Replace(log, "cpu: Apple M1\n", "", 1)},
		{"verbose", strings.Replace(log, row, "=== RUN   TestOperandContract\n--- PASS: TestOperandContract (0.00s)\nBenchmarkX\nBenchmarkX/sub\n"+row, 1)},
		{"same environment repeated", strings.Replace(log, "goos: darwin", "goos: darwin\ngoos: darwin", 1)},
		{"coverage summary", strings.Replace(log, "10.0s", "10.0s coverage: 20.0% of statements", 1)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			samples, meta, err := parseBenchLog(strings.NewReader(tc.log))
			if err != nil {
				t.Fatal(err)
			}
			if len(samples) != 1 || len(samples["BenchmarkX/sub-10"]) != 5 || meta.tokens["count"] != "5" || meta.tokens["go"] != "go1.26.5" || meta.env["goos"] != "darwin" {
				t.Fatalf("unexpected samples or metadata: %v %+v", samples, meta)
			}
		})
	}
}

func TestParseBenchLogRejectsInvalidEvidence(t *testing.T) {
	row := "BenchmarkX-10 100 1.25 ns/op"
	valid := benchLog(row)
	for _, tc := range []struct{ name, log, want string }{
		{"under count", strings.Replace(valid, "count=1", "count=5", 1), "has 1 samples; BENCH-META count=5"},
		{"over count", strings.Replace(valid, row, row+"\n"+row, 1), "has 2 samples"},
		{"uneven count", strings.Replace(benchLog(row, row), "PASS", "BenchmarkY-10 100 1.0 ns/op\nPASS", 1), "has 1 samples"},
		{"appended FAIL", valid + "FAIL\nFAIL simulated-incomplete-benchmark-run\n", "failed benchmark run"},
		{"failure before PASS", strings.Replace(valid, "PASS", "--- FAIL: BenchmarkX (1.00s)\nPASS", 1), "failed benchmark run"},
		{"panic", strings.Replace(valid, "PASS", "panic: benchmark crashed\nPASS", 1), "failed benchmark run"},
		{"fatal", valid + "fatal error: runtime failure\n", "failed benchmark run"},
		{"exit status", valid + "exit status 2\n", "failed benchmark run"},
		{"no footer", strings.Split(valid, "PASS")[0], "incomplete benchmark run"},
		{"PASS only", strings.Split(valid, "PASS")[0] + "PASS\n", "incomplete benchmark run"},
		{"ok without PASS", strings.Replace(valid, "PASS\n", "", 1), "invalid package completion"},
		{"wrong package", strings.Replace(valid, "ok  \tgithub.com/sky1core/bid754/bid754-go/internal/bidgo", "ok  \twrong/pkg", 1), "invalid package completion"},
		{"cached summary", strings.Replace(valid, "10.0s", "(cached)", 1), "invalid package elapsed time"},
		{"truncated summary", strings.Replace(valid, "10.0s", "", 1), "invalid package completion"},
		{"duplicate PASS", strings.Replace(valid, "PASS", "PASS\nPASS", 1), "unexpected PASS"},
		{"row after PASS", strings.Replace(valid, "PASS", "PASS\n"+row, 1), "outside an active"},
		{"trailing truncated run", valid + "BenchmarkUnfinished\n", "after completed"},
		{"concatenated runs", valid + valid, "after completed"},
		{"duplicate header", strings.Split(valid, "\n")[0] + "\n" + valid, "exactly once"},
		{"conflicting headers", strings.Replace(strings.Split(valid, "\n")[0], "go1.26.5", "go1.26.1", 1) + "\n" + valid, "exactly once"},
		{"duplicate token", strings.Replace(valid, "count=1", "count=1 count=1", 1), "duplicate BENCH-META count"},
		{"conflicting token", strings.Replace(valid, "count=1", "count=1 count=5", 1), "duplicate BENCH-META count"},
		{"malformed token", strings.Replace(valid, "count=1", "count 1", 1), "malformed BENCH-META"},
		{"empty go", strings.Replace(valid, "go=go1.26.5", "go=", 1), "malformed BENCH-META"},
		{"missing go", strings.Replace(valid, "go=go1.26.5 ", "", 1), "BENCH-META go"},
		{"placeholder go", strings.Replace(valid, "go=go1.26.5", "go=(none)", 1), "BENCH-META go"},
		{"missing target", strings.Replace(valid, "target=bench-bidgo ", "", 1), "BENCH-META target"},
		{"missing tree", strings.Replace(valid, "tree=synthetic ", "", 1), "BENCH-META tree"},
		{"unknown tree", strings.Replace(valid, "tree=synthetic", "tree=unknown", 1), "BENCH-META tree"},
		{"missing benchtime", strings.Replace(valid, "benchtime=1s ", "", 1), "BENCH-META benchtime"},
		{"invalid benchtime", strings.Replace(valid, "benchtime=1s", "benchtime=0x", 1), "BENCH-META benchtime"},
		{"negative benchtime", strings.Replace(valid, "benchtime=1s", "benchtime=-1s", 1), "BENCH-META benchtime"},
		{"missing build", strings.Replace(valid, "build="+strings.Repeat("a", 64)+" ", "", 1), "BENCH-META build"},
		{"invalid build", strings.Replace(valid, "build="+strings.Repeat("a", 64), "build=abc", 1), "BENCH-META build"},
		{"missing count", strings.Replace(valid, "count=1 ", "", 1), "BENCH-META count"},
		{"zero count", strings.Replace(valid, "count=1", "count=0", 1), "positive integer"},
		{"negative count", strings.Replace(valid, "count=1", "count=-1", 1), "positive integer"},
		{"fractional count", strings.Replace(valid, "count=1", "count=1.5", 1), "positive integer"},
		{"overflow count", strings.Replace(valid, "count=1", "count=99999999999999999999", 1), "positive integer"},
		{"missing header", strings.Join(strings.Split(valid, "\n")[1:], "\n"), "outside an active"},
		{"missing goos", strings.Replace(valid, "goos: darwin\n", "", 1), "goos metadata is missing"},
		{"missing goarch", strings.Replace(valid, "goarch: arm64\n", "", 1), "goarch metadata is missing"},
		{"conflicting goos", strings.Replace(valid, "goos: darwin", "goos: darwin\ngoos: linux", 1), "conflicting"},
		{"conflicting cpu", strings.Replace(valid, "cpu: Apple M1", "cpu: Apple M1\ncpu: Another CPU", 1), "conflicting"},
		{"conflicting package", strings.Replace(valid, "cpu: Apple M1", "pkg: wrong/pkg", 1), "conflicting"},
		{"empty cpu", strings.Replace(valid, "cpu: Apple M1", "cpu:", 1), "empty"},
		{"missing rows", strings.Replace(valid, row+"\n", "", 1), "unexpected PASS"},
		{"partial row", strings.Replace(valid, row, "BenchmarkX-10 100", 1), "malformed benchmark"},
		{"bad iterations", strings.Replace(valid, row, "BenchmarkX-10 zero 1.25 ns/op", 1), "malformed benchmark"},
		{"zero iterations", strings.Replace(valid, row, "BenchmarkX-10 0 1.25 ns/op", 1), "malformed benchmark"},
		{"missing ns/op", strings.Replace(valid, "ns/op", "B/op", 1), "lacks ns/op"},
		{"duplicate ns/op", strings.Replace(valid, row, row+" 1 ns/op", 1), "duplicate ns/op"},
		{"bad ns/op", strings.Replace(valid, "1.25", "bad", 1), "invalid ns/op"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := parseBenchLog(strings.NewReader(tc.log))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want %q", err, tc.want)
			}
		})
	}
}

func TestRequireComparableMeta(t *testing.T) {
	valid := benchLog("BenchmarkX-10 100 1.25 ns/op")
	for _, tc := range []struct{ name, baseline, candidate, want string }{
		{"matching", valid, valid, ""},
		{"historical tree and date", valid, strings.ReplaceAll(strings.ReplaceAll(valid, "synthetic", "current"), "2026-07-13", "2026-09-08"), ""},
		{"no cpu on both", strings.ReplaceAll(valid, "cpu: Apple M1\n", ""), strings.ReplaceAll(valid, "cpu: Apple M1\n", ""), ""},
		{"go", valid, strings.ReplaceAll(valid, "go1.26.5", "go1.26.1"), "BENCH-META go"},
		{"target", valid, strings.ReplaceAll(valid, "target=bench-bidgo", "target=bench-native"), "BENCH-META target"},
		{"count", valid, benchLog("BenchmarkX-10 100 1.25 ns/op", "BenchmarkX-10 100 1.25 ns/op"), "BENCH-META count"},
		{"goos", valid, strings.ReplaceAll(valid, "darwin", "linux"), "goos"},
		{"goarch", valid, strings.ReplaceAll(valid, "arm64", "amd64"), "goarch"},
		{"cpu", valid, strings.ReplaceAll(valid, "Apple M1", "Apple M1 Max"), "cpu"},
		{"cpu missing", valid, strings.ReplaceAll(valid, "cpu: Apple M1\n", ""), "cpu"},
		{"package", valid, strings.ReplaceAll(valid, "github.com/sky1core/bid754/bid754-go/internal/bidgo", "another/pkg"), "pkg"},
		{"benchtime", valid, strings.Replace(valid, "benchtime=1s", "benchtime=100x", 1), "BENCH-META benchtime"},
		{"build flags", strings.Replace(valid, "count=1", "count=1 buildflags=sha256:aaa", 1), strings.Replace(valid, "count=1", "count=1 buildflags=sha256:bbb", 1), "BENCH-META buildflags"},
		{"missing build flags", strings.Replace(valid, "count=1", "count=1 buildflags=default", 1), valid, "BENCH-META buildflags"},
		{"extra control", valid, strings.Replace(valid, "count=1", "count=1 cgo=1", 1), "BENCH-META cgo"},
		{"matching controls", strings.Replace(valid, "count=1", "count=1 buildflags=sha256:aaa", 1), strings.Replace(valid, "count=1", "count=1 buildflags=sha256:aaa", 1), ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := requireComparableMeta(mustParseMeta(t, tc.baseline), mustParseMeta(t, tc.candidate))
			if tc.want == "" && err != nil || tc.want != "" && (err == nil || !strings.Contains(err.Error(), tc.want)) {
				t.Fatalf("error = %v, want %q", err, tc.want)
			}
		})
	}
}

func TestRunEvidenceGate(t *testing.T) {
	t.Setenv(thresholdEnvVar, "")
	t.Setenv(minDeltaEnvVar, "")
	row := "BenchmarkX-10 100 100 ns/op"
	baseline := strings.Replace(benchLog(row), "synthetic", "historical", 1)
	candidate := strings.Replace(baseline, "historical", "current", 1)
	var auditRows []string
	for i := 0; i < 174; i++ {
		auditRows = append(auditRows, fmt.Sprintf("BenchmarkAudit/row%d-10 100 100 ns/op", i))
	}
	auditBaseline := strings.Replace(benchLog(auditRows...), "count=1", "count=5", 1)
	for _, tc := range []struct {
		name, baseline, candidate string
		extra                     []string
		code                      int
		want                      string
	}{
		{"historical baseline current candidate", baseline, candidate, []string{"-expected-candidate-tree=current", "-expected-target=bench-bidgo"}, 0, "benchdiff: PASS"},
		{"historical comparison without binding", baseline, baseline, nil, 0, "benchdiff: PASS"},
		{"stale candidate", baseline, baseline, []string{"-expected-candidate-tree=current"}, 2, "tree mismatch"},
		{"wrong target on both", baseline, candidate, []string{"-expected-target=bench-native"}, 2, "target mismatch"},
		{"empty tree flag", baseline, candidate, []string{"-expected-candidate-tree="}, 2, "must not be empty"},
		{"empty target flag", baseline, candidate, []string{"-expected-target="}, 2, "must not be empty"},
		{"positional argument", baseline, candidate, []string{"unexpected"}, 2, "positional arguments"},
		{"failed candidate", baseline, candidate + "FAIL\nFAIL simulated-incomplete-benchmark-run\n", nil, 2, "candidate:"},
		{"failed baseline", baseline + "FAIL\n", candidate, nil, 2, "baseline:"},
		{"missing benchtime on both", strings.ReplaceAll(baseline, "benchtime=1s ", ""), strings.ReplaceAll(candidate, "benchtime=1s ", ""), nil, 2, "BENCH-META benchtime"},
		{"missing build on both", strings.ReplaceAll(baseline, "build="+strings.Repeat("a", 64)+" ", ""), strings.ReplaceAll(candidate, "build="+strings.Repeat("a", 64)+" ", ""), nil, 2, "BENCH-META build"},
		{"missing go on both", strings.ReplaceAll(baseline, "go=go1.26.5 ", ""), strings.ReplaceAll(candidate, "go=go1.26.5 ", ""), nil, 2, "BENCH-META go"},
		{"ambiguous on both", strings.ReplaceAll(baseline, "count=1", "count=1 count=5"), strings.ReplaceAll(candidate, "count=1", "count=1 count=5"), nil, 2, "duplicate BENCH-META"},
		{"174 median-only rows", baseline, auditBaseline, nil, 2, "has 1 samples; BENCH-META count=5"},
		{"174 median-only rows plus FAIL", baseline, auditBaseline + "FAIL\nFAIL simulated-incomplete-benchmark-run\n", nil, 2, "failed benchmark run"},
		{"regression", baseline, strings.Replace(candidate, "100 ns/op", "120 ns/op", 1), nil, 1, "REGRESSION"},
		{"new row", baseline, strings.Replace(candidate, "PASS", "BenchmarkNew-10 100 100 ns/op\nPASS", 1), nil, 0, "new (no baseline)"},
		{"vanished row", baseline, strings.Replace(candidate, "BenchmarkX", "BenchmarkY", 1), nil, 1, "MISSING IN CANDIDATE"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			basePath, candPath := filepath.Join(dir, "baseline.log"), filepath.Join(dir, "candidate.log")
			for path, content := range map[string]string{basePath: tc.baseline, candPath: tc.candidate} {
				if err := os.WriteFile(path, []byte(content), 0600); err != nil {
					t.Fatal(err)
				}
			}
			args := append([]string{"-baseline", basePath, "-candidate", candPath}, tc.extra...)
			var stdout, stderr strings.Builder
			code := run(args, &stdout, &stderr)
			if code != tc.code || !strings.Contains(stdout.String()+stderr.String(), tc.want) {
				t.Fatalf("exit=%d stdout=%s stderr=%s; want exit=%d containing %q", code, &stdout, &stderr, tc.code, tc.want)
			}
			if tc.code != 0 && strings.Contains(stdout.String(), "benchdiff: PASS") {
				t.Fatal("invalid or regressed evidence printed PASS")
			}
		})
	}
}
