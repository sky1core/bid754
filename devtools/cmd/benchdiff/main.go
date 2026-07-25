// benchdiff is the Go benchmark regression gate. It parses two captured
// `go test -bench` logs — a deliberately saved baseline and a candidate run —
// aggregates the repeated samples of each benchmark (-count reruns) into a
// per-benchmark median ns/op, and compares candidate against baseline. It is
// the Go-matrix counterpart of the Criterion `pinned` baseline on the Rust
// leg: the baseline is only ever written by an explicit save step, so change
// verdicts are always against a chosen reference, never against whatever ran
// last.
//
// Policy:
//   - a benchmark present in the baseline but absent from the candidate fails
//     the gate (a vanished benchmark would otherwise hide a regression);
//   - a benchmark only present in the candidate is reported as
//     "new (no baseline)" and does not fail the gate;
//   - a median regression above the threshold (default 8%, overridable via
//     BENCH_REGRESSION_THRESHOLD; the default clears the ±3–4% run-to-run
//     noise measured on the Apple M1 reference machine) fails the gate;
//   - a row that clears the percentage threshold but whose absolute median
//     delta is at or below the minimum delta (default 0.25 ns/op, overridable
//     via BENCH_REGRESSION_MIN_DELTA_NS; 0 disables the floor and restores the
//     pure-percentage gate) does not fail and is reported as
//     "ok (below min delta)" — never as a plain "ok". The floor is global, so
//     the effective threshold of a row is max(8%, 0.25 ns / baseline median):
//     8% of 3.125 ns is exactly 0.25 ns, so at or above a 3.125 ns baseline
//     median the percentage rule always binds first and the sensitivity is
//     unchanged, while rows below that boundary are held to the absolute floor
//     instead (sub-nanosecond rows are the extreme case, not the only one —
//     the motivating row, the inline-budget canary
//     BenchmarkAlignedBID128/from_int64, read +26.56% on a 0.1834 ns wobble
//     of 0.6905 → 0.8739 ns/op). A held row is not evidence that it did not
//     regress; see docs/BUILD.md for the masking bounds this accepts;
//   - improvements are reported but never fail;
//   - the threshold and every ns/op sample must be finite and positive, and
//     the minimum delta must be finite and non-negative
//     (NaN/Inf/out-of-range values are input errors, never silently compared);
//   - the two logs must describe comparable runs: the BENCH-META count and go=
//     tokens and the goos/goarch lines must be present in both logs and,
//     together with any cpu lines, must match between baseline and candidate,
//     otherwise the comparison fails as an input error. go= records the Go
//     toolchain the samples were taken with, because a toolchain change redoes
//     inlining and code layout and so moves medians on unchanged source; a
//     BENCH-META line that omits a token carries "(none)" for it, so logs
//     predating the token still compare with each other but never silently
//     against a log that records one.
//
// Exit codes: 0 pass, 1 regression or vanished benchmark, 2 usage/input
// errors (unreadable log, no benchmark rows, bad threshold, bad minimum delta
// or bad sample value, incomparable run metadata).
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"slices"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
)

const (
	defaultRegressionThresholdPct = 8.0
	thresholdEnvVar               = "BENCH_REGRESSION_THRESHOLD"
	defaultRegressionMinDeltaNs   = 0.25
	minDeltaEnvVar                = "BENCH_REGRESSION_MIN_DELTA_NS"
	benchMetaPrefix               = "BENCH-META "
)

type diffStatus string

const (
	statusOK            diffStatus = "ok"
	statusBelowMinDelta diffStatus = "ok (below min delta)"
	statusImproved      diffStatus = "improved"
	statusRegression    diffStatus = "REGRESSION"
	statusMissing       diffStatus = "MISSING IN CANDIDATE"
	statusNew           diffStatus = "new (no baseline)"
)

type diffRow struct {
	name        string
	baselineNs  float64
	candidateNs float64
	changePct   float64
	status      diffStatus
}

func main() {
	baselinePath := flag.String("baseline", "", "path to the saved baseline `go test -bench` log")
	candidatePath := flag.String("candidate", "", "path to the candidate `go test -bench` log")
	flag.Parse()
	if *baselinePath == "" || *candidatePath == "" {
		fmt.Fprintln(os.Stderr, "benchdiff: -baseline and -candidate are required")
		os.Exit(2)
	}

	thresholdPct, err := regressionThresholdPct(os.Getenv(thresholdEnvVar))
	if err != nil {
		fatal("%v", err)
	}
	minDeltaNs, err := regressionMinDeltaNs(os.Getenv(minDeltaEnvVar))
	if err != nil {
		fatal("%v", err)
	}

	baseline, baselineMeta, err := parseBenchLogFile(*baselinePath)
	if err != nil {
		fatal("baseline: %v", err)
	}
	candidate, candidateMeta, err := parseBenchLogFile(*candidatePath)
	if err != nil {
		fatal("candidate: %v", err)
	}
	if err := requireComparableMeta(baselineMeta, candidateMeta); err != nil {
		fatal("baseline %s and candidate %s are not comparable: %v", *baselinePath, *candidatePath, err)
	}

	rows, failed := compareBenchmarks(baseline, candidate, thresholdPct, minDeltaNs)
	printReport(os.Stdout, rows, *baselinePath, *candidatePath, thresholdPct, minDeltaNs, failed)
	if failed {
		os.Exit(1)
	}
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "benchdiff: "+format+"\n", args...)
	os.Exit(2)
}

// regressionThresholdPct resolves the regression threshold in percent from
// the raw environment value ("" means unset → default). NaN and ±Inf pass a
// plain `< 0` comparison, so the finite/positive requirement is explicit: a
// non-finite or non-positive threshold silently disables or distorts the
// gate instead of configuring it.
func regressionThresholdPct(raw string) (float64, error) {
	if raw == "" {
		return defaultRegressionThresholdPct, nil
	}
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, fmt.Errorf("%s=%q is not a number", thresholdEnvVar, raw)
	}
	if math.IsNaN(value) || math.IsInf(value, 0) || value <= 0 {
		return 0, fmt.Errorf("%s=%q must be a finite number > 0 (percent)", thresholdEnvVar, raw)
	}
	return value, nil
}

// regressionMinDeltaNs resolves the absolute median-delta floor in ns/op from
// the raw environment value ("" means unset → default). Unlike the threshold,
// 0 is a legal setting that disables the floor and restores the pure
// percentage gate; NaN, ±Inf, and negative values are rejected for the same
// reason as there — they pass a plain `< 0` comparison and would disable or
// distort the gate instead of configuring it.
func regressionMinDeltaNs(raw string) (float64, error) {
	if raw == "" {
		return defaultRegressionMinDeltaNs, nil
	}
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, fmt.Errorf("%s=%q is not a number", minDeltaEnvVar, raw)
	}
	if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 {
		return 0, fmt.Errorf("%s=%q must be a finite number >= 0 (ns/op; 0 disables the floor)", minDeltaEnvVar, raw)
	}
	return value, nil
}

// benchLogMeta carries the comparability identity of one captured log: the
// BENCH-META count= and go= values and the goos/goarch/cpu environment lines,
// each as a sorted set of the distinct values seen.
type benchLogMeta struct {
	counts     []string
	toolchains []string
	goos       []string
	goarch     []string
	cpu        []string
}

func parseBenchLogFile(path string) (map[string][]float64, benchLogMeta, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, benchLogMeta{}, err
	}
	defer file.Close()
	samples, meta, err := parseBenchLog(file)
	if err != nil {
		return nil, benchLogMeta{}, fmt.Errorf("%s: %w", path, err)
	}
	if len(samples) == 0 {
		return nil, benchLogMeta{}, fmt.Errorf("%s contains no `go test -bench` result lines", path)
	}
	return samples, meta, nil
}

// parseBenchLog collects every ns/op sample per benchmark name (the full
// name including the -GOMAXPROCS suffix, so runs from a different parallelism
// setting never silently pair up) from a `go test -bench` log, plus the
// BENCH-META count=/go= and goos/goarch/cpu comparability metadata.
func parseBenchLog(r io.Reader) (map[string][]float64, benchLogMeta, error) {
	samples := make(map[string][]float64)
	var meta benchLogMeta
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, benchMetaPrefix) {
			meta.counts = appendDistinct(meta.counts, benchMetaToken(line, "count"))
			meta.toolchains = appendDistinct(meta.toolchains, benchMetaToken(line, "go"))
			continue
		}
		if value, ok := strings.CutPrefix(line, "goos: "); ok {
			meta.goos = appendDistinct(meta.goos, strings.TrimSpace(value))
			continue
		}
		if value, ok := strings.CutPrefix(line, "goarch: "); ok {
			meta.goarch = appendDistinct(meta.goarch, strings.TrimSpace(value))
			continue
		}
		if value, ok := strings.CutPrefix(line, "cpu: "); ok {
			meta.cpu = appendDistinct(meta.cpu, strings.TrimSpace(value))
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 4 || !strings.HasPrefix(fields[0], "Benchmark") {
			continue
		}
		for i := 2; i < len(fields); i++ {
			if fields[i] != "ns/op" {
				continue
			}
			value, err := strconv.ParseFloat(fields[i-1], 64)
			if err != nil {
				return nil, benchLogMeta{}, fmt.Errorf("benchmark line %q has unparsable ns/op value: %v", line, err)
			}
			// NaN passes a plain `<= 0` comparison; require finite and
			// positive explicitly so a corrupt sample cannot enter a median.
			if math.IsNaN(value) || math.IsInf(value, 0) || value <= 0 {
				return nil, benchLogMeta{}, fmt.Errorf("benchmark line %q has non-positive or non-finite ns/op value", line)
			}
			samples[fields[0]] = append(samples[fields[0]], value)
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, benchLogMeta{}, err
	}
	return samples, meta, nil
}

// benchMetaToken extracts the value of the `key=` token from a BENCH-META
// line, which the caller must already have identified by benchMetaPrefix.
//
// A BENCH-META line that omits the token still identifies the log as carrying
// run metadata, so the absence is recorded as the distinct value "(none)"
// rather than as nothing: a log missing the token then pairs only with another
// log missing it, and never silently with a log that carries it. That is what
// makes the go= token retroactively safe to add — two pre-token logs still
// compare, while a pre-token baseline against a post-token candidate is a
// reported mismatch instead of a silent comparison across toolchains.
func benchMetaToken(line, key string) string {
	for _, field := range strings.Fields(line)[1:] {
		if value, ok := strings.CutPrefix(field, key+"="); ok {
			return value
		}
	}
	return "(none)"
}

func appendDistinct(values []string, value string) []string {
	if slices.Contains(values, value) {
		return values
	}
	values = append(values, value)
	slices.Sort(values)
	return values
}

// toolchainMismatchHint is the actionable half of a BENCH-META go= mismatch.
// A Go toolchain change re-runs inlining and code-layout decisions, so it moves
// medians on unchanged source — IntelCBID128/minnum has been observed shifting
// 6.93 → 8.17 ns from layout alone — which silently invalidates a saved
// baseline. The token was added after the current baselines were saved, so the
// first check against a pre-token baseline lands here by design.
const toolchainMismatchHint = "toolchain provenance mismatch — the baseline predates toolchain recording (or was measured on a different toolchain), " +
	"and a Go toolchain change redoes inlining and code layout, so the saved medians are not comparable; " +
	"re-measure and re-save the baseline on the current toolchain on an idle host " +
	"(make bench-native && make bench-bidgo && make bench-go-baseline)"

// requireComparableMeta fails when the two logs disagree on the BENCH-META
// count or go= token or on any goos/goarch/cpu environment line: medians from
// different sample counts, different toolchains, or different machines are not
// a like-for-like comparison, and a silent pass here would launder an
// environment change as a performance result. BENCH-META count and go=, goos,
// and goarch must be present in both logs (every repo bench target and
// `go test -bench` run emits them), so stripping the header lines cannot bypass
// the guard; the cpu line may be absent on platforms Go cannot identify, but
// must then be absent on both sides. A log whose BENCH-META line omits a token
// carries "(none)" for it, so pre-token logs still compare with each other.
func requireComparableMeta(baseline, candidate benchLogMeta) error {
	for _, item := range []struct {
		name                string
		required            bool
		baseline, candidate []string
		hint                string
	}{
		{name: "BENCH-META count", required: true, baseline: baseline.counts, candidate: candidate.counts},
		{name: "BENCH-META go", required: true, baseline: baseline.toolchains, candidate: candidate.toolchains, hint: toolchainMismatchHint},
		{name: "goos", required: true, baseline: baseline.goos, candidate: candidate.goos},
		{name: "goarch", required: true, baseline: baseline.goarch, candidate: candidate.goarch},
		{name: "cpu", required: false, baseline: baseline.cpu, candidate: candidate.cpu},
	} {
		if item.required && (len(item.baseline) == 0 || len(item.candidate) == 0) {
			return withHint(fmt.Errorf("%s is missing from baseline and/or candidate (baseline %s, candidate %s); a log without run metadata cannot prove comparability",
				item.name, formatMetaValues(item.baseline), formatMetaValues(item.candidate)), item.hint)
		}
		if !slices.Equal(item.baseline, item.candidate) {
			return withHint(fmt.Errorf("%s mismatch: baseline %s vs candidate %s",
				item.name, formatMetaValues(item.baseline), formatMetaValues(item.candidate)), item.hint)
		}
	}
	return nil
}

// withHint appends an actionable remediation clause to err, when the failing
// comparability item carries one.
func withHint(err error, hint string) error {
	if hint == "" {
		return err
	}
	return fmt.Errorf("%w; %s", err, hint)
}

func formatMetaValues(values []string) string {
	if len(values) == 0 {
		return "(absent)"
	}
	return strings.Join(values, "|")
}

// median returns the median of values: the middle sample for an odd count,
// the mean of the two middle samples for an even count. values must be
// non-empty.
func median(values []float64) float64 {
	sorted := append([]float64(nil), values...)
	sort.Float64s(sorted)
	mid := len(sorted) / 2
	if len(sorted)%2 == 1 {
		return sorted[mid]
	}
	return (sorted[mid-1] + sorted[mid]) / 2
}

// compareBenchmarks pairs baseline and candidate medians by benchmark name.
// It fails (second return true) on any baseline benchmark missing from the
// candidate and on any median regression that is strictly above thresholdPct
// *and* strictly above minDeltaNs ns/op in absolute median delta. A row that
// clears only the percentage half of that conjunction keeps its own
// statusBelowMinDelta status, so a row that survived on the floor alone stays
// visible in the report instead of collapsing into a plain "ok"; the floor
// suppresses a failure verdict, it never suppresses the row or its change%.
func compareBenchmarks(baseline, candidate map[string][]float64, thresholdPct, minDeltaNs float64) ([]diffRow, bool) {
	names := make([]string, 0, len(baseline)+len(candidate))
	for name := range baseline {
		names = append(names, name)
	}
	for name := range candidate {
		if _, ok := baseline[name]; !ok {
			names = append(names, name)
		}
	}
	sort.Strings(names)

	rows := make([]diffRow, 0, len(names))
	failed := false
	for _, name := range names {
		baseSamples, inBaseline := baseline[name]
		candSamples, inCandidate := candidate[name]
		switch {
		case inBaseline && !inCandidate:
			rows = append(rows, diffRow{name: name, baselineNs: median(baseSamples), status: statusMissing})
			failed = true
		case !inBaseline && inCandidate:
			rows = append(rows, diffRow{name: name, candidateNs: median(candSamples), status: statusNew})
		default:
			baseMedian := median(baseSamples)
			candMedian := median(candSamples)
			deltaNs := candMedian - baseMedian
			changePct := deltaNs / baseMedian * 100
			status := statusOK
			switch {
			case changePct > thresholdPct && deltaNs > minDeltaNs:
				status = statusRegression
				failed = true
			case changePct > thresholdPct:
				status = statusBelowMinDelta
			case changePct < 0:
				status = statusImproved
			}
			rows = append(rows, diffRow{
				name:        name,
				baselineNs:  baseMedian,
				candidateNs: candMedian,
				changePct:   changePct,
				status:      status,
			})
		}
	}
	return rows, failed
}

func printReport(w io.Writer, rows []diffRow, baselinePath, candidatePath string, thresholdPct, minDeltaNs float64, failed bool) {
	fmt.Fprintf(w, "benchdiff: baseline=%s candidate=%s threshold=+%.4g%% min-delta=+%.4gns (median ns/op)\n",
		baselinePath, candidatePath, thresholdPct, minDeltaNs)
	tw := tabwriter.NewWriter(w, 2, 4, 2, ' ', 0)
	fmt.Fprintln(tw, "benchmark\tbaseline ns/op\tcandidate ns/op\tchange\tstatus")
	for _, row := range rows {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
			row.name, formatNs(row.baselineNs), formatNs(row.candidateNs), formatChange(row), row.status)
	}
	tw.Flush()
	if failed {
		fmt.Fprintf(w, "benchdiff: FAIL — regression above both +%.4g%% and +%.4gns, or benchmark missing from candidate\n",
			thresholdPct, minDeltaNs)
	} else {
		fmt.Fprintln(w, "benchdiff: PASS")
	}
}

func formatNs(value float64) string {
	if value == 0 {
		return "-"
	}
	return strconv.FormatFloat(value, 'f', 2, 64)
}

func formatChange(row diffRow) string {
	if row.status == statusMissing || row.status == statusNew {
		return "-"
	}
	return fmt.Sprintf("%+.2f%%", row.changePct)
}
