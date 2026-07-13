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
//   - improvements are reported but never fail;
//   - the threshold and every ns/op sample must be finite and positive
//     (NaN/Inf/non-positive values are input errors, never silently compared);
//   - the two logs must describe comparable runs: the BENCH-META count and
//     the goos/goarch lines must be present in both logs and, together with
//     any cpu lines, must match between baseline and candidate, otherwise
//     the comparison fails as an input error.
//
// Exit codes: 0 pass, 1 regression or vanished benchmark, 2 usage/input
// errors (unreadable log, no benchmark rows, bad threshold or sample value,
// incomparable run metadata).
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
)

type diffStatus string

const (
	statusOK         diffStatus = "ok"
	statusImproved   diffStatus = "improved"
	statusRegression diffStatus = "REGRESSION"
	statusMissing    diffStatus = "MISSING IN CANDIDATE"
	statusNew        diffStatus = "new (no baseline)"
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

	rows, failed := compareBenchmarks(baseline, candidate, thresholdPct)
	printReport(os.Stdout, rows, *baselinePath, *candidatePath, thresholdPct, failed)
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

// benchLogMeta carries the comparability identity of one captured log: the
// BENCH-META count= values and the goos/goarch/cpu environment lines, each
// as a sorted set of the distinct values seen.
type benchLogMeta struct {
	counts []string
	goos   []string
	goarch []string
	cpu    []string
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
// BENCH-META/goos/goarch/cpu comparability metadata.
func parseBenchLog(r io.Reader) (map[string][]float64, benchLogMeta, error) {
	samples := make(map[string][]float64)
	var meta benchLogMeta
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if value, ok := benchMetaCount(line); ok {
			meta.counts = appendDistinct(meta.counts, value)
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

// benchMetaCount extracts the count= token value from a BENCH-META line.
func benchMetaCount(line string) (string, bool) {
	if !strings.HasPrefix(line, "BENCH-META ") {
		return "", false
	}
	for _, field := range strings.Fields(line)[1:] {
		if value, ok := strings.CutPrefix(field, "count="); ok {
			return value, true
		}
	}
	// A BENCH-META line without a count token still identifies the log as
	// carrying run metadata; record the absence explicitly so a countless
	// baseline never pairs with a counted candidate.
	return "(none)", true
}

func appendDistinct(values []string, value string) []string {
	if slices.Contains(values, value) {
		return values
	}
	values = append(values, value)
	slices.Sort(values)
	return values
}

// requireComparableMeta fails when the two logs disagree on the BENCH-META
// count or on any goos/goarch/cpu environment line: medians from different
// sample counts or different machines are not a like-for-like comparison,
// and a silent pass here would launder an environment change as a
// performance result. BENCH-META count, goos, and goarch must be present in
// both logs (every repo bench target and `go test -bench` run emits them), so
// stripping the header lines cannot bypass the guard; the cpu line may be
// absent on platforms Go cannot identify, but must then be absent on both
// sides.
func requireComparableMeta(baseline, candidate benchLogMeta) error {
	for _, item := range []struct {
		name                string
		required            bool
		baseline, candidate []string
	}{
		{"BENCH-META count", true, baseline.counts, candidate.counts},
		{"goos", true, baseline.goos, candidate.goos},
		{"goarch", true, baseline.goarch, candidate.goarch},
		{"cpu", false, baseline.cpu, candidate.cpu},
	} {
		if item.required && (len(item.baseline) == 0 || len(item.candidate) == 0) {
			return fmt.Errorf("%s is missing from baseline and/or candidate (baseline %s, candidate %s); a log without run metadata cannot prove comparability",
				item.name, formatMetaValues(item.baseline), formatMetaValues(item.candidate))
		}
		if !slices.Equal(item.baseline, item.candidate) {
			return fmt.Errorf("%s mismatch: baseline %s vs candidate %s",
				item.name, formatMetaValues(item.baseline), formatMetaValues(item.candidate))
		}
	}
	return nil
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
// candidate and on any median regression strictly above thresholdPct.
func compareBenchmarks(baseline, candidate map[string][]float64, thresholdPct float64) ([]diffRow, bool) {
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
			changePct := (candMedian - baseMedian) / baseMedian * 100
			status := statusOK
			if changePct > thresholdPct {
				status = statusRegression
				failed = true
			} else if changePct < 0 {
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

func printReport(w io.Writer, rows []diffRow, baselinePath, candidatePath string, thresholdPct float64, failed bool) {
	fmt.Fprintf(w, "benchdiff: baseline=%s candidate=%s threshold=+%.4g%% (median ns/op)\n", baselinePath, candidatePath, thresholdPct)
	tw := tabwriter.NewWriter(w, 2, 4, 2, ' ', 0)
	fmt.Fprintln(tw, "benchmark\tbaseline ns/op\tcandidate ns/op\tchange\tstatus")
	for _, row := range rows {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
			row.name, formatNs(row.baselineNs), formatNs(row.candidateNs), formatChange(row), row.status)
	}
	tw.Flush()
	if failed {
		fmt.Fprintf(w, "benchdiff: FAIL — regression above +%.4g%% or benchmark missing from candidate\n", thresholdPct)
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
