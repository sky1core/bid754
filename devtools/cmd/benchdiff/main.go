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
//
// Exit codes: 0 pass, 1 regression or vanished benchmark, 2 usage/input
// errors (unreadable log, no benchmark rows, bad threshold, bad minimum delta
// or bad sample value, incomparable run metadata).
package main

import (
	"bufio"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
)

const (
	defaultRegressionThresholdPct = 8.0
	thresholdEnvVar               = "BENCH_REGRESSION_THRESHOLD"
	defaultRegressionMinDeltaNs   = 0.25
	minDeltaEnvVar                = "BENCH_REGRESSION_MIN_DELTA_NS"
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
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("benchdiff", flag.ContinueOnError)
	flags.SetOutput(stderr)
	baselinePath := flags.String("baseline", "", "path to the saved baseline `go test -bench` log")
	candidatePath := flags.String("candidate", "", "path to the candidate `go test -bench` log")
	expectedTree := flags.String("expected-candidate-tree", "", "require candidate BENCH-META tree to equal this exact source identifier; baseline may be historical")
	expectedTarget := flags.String("expected-target", "", "require both logs to use this BENCH-META target")
	if err := flags.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}
	inputError := func(format string, args ...any) int {
		fmt.Fprintf(stderr, "benchdiff: "+format+"\n", args...)
		return 2
	}
	if *baselinePath == "" || *candidatePath == "" || flags.NArg() != 0 {
		return inputError("-baseline and -candidate are required; positional arguments are not supported")
	}
	var flagErr error
	flags.Visit(func(f *flag.Flag) {
		if (f.Name == "expected-candidate-tree" || f.Name == "expected-target") && strings.TrimSpace(f.Value.String()) == "" {
			flagErr = fmt.Errorf("-%s must not be empty when supplied", f.Name)
		}
	})
	if flagErr != nil {
		return inputError("%v", flagErr)
	}
	thresholdPct, err := regressionThresholdPct(os.Getenv(thresholdEnvVar))
	if err != nil {
		return inputError("%v", err)
	}
	minDeltaNs, err := regressionMinDeltaNs(os.Getenv(minDeltaEnvVar))
	if err != nil {
		return inputError("%v", err)
	}
	baseline, baselineMeta, err := parseBenchLogFile(*baselinePath)
	if err != nil {
		return inputError("baseline: %v", err)
	}
	candidate, candidateMeta, err := parseBenchLogFile(*candidatePath)
	if err != nil {
		return inputError("candidate: %v", err)
	}
	if *expectedTree != "" && candidateMeta.tokens["tree"] != *expectedTree {
		return inputError("candidate BENCH-META tree mismatch: expected %q, got %q", *expectedTree, candidateMeta.tokens["tree"])
	}
	if *expectedTarget != "" && (baselineMeta.tokens["target"] != *expectedTarget || candidateMeta.tokens["target"] != *expectedTarget) {
		return inputError("BENCH-META target mismatch: expected %q, baseline %q, candidate %q", *expectedTarget, baselineMeta.tokens["target"], candidateMeta.tokens["target"])
	}
	if err := requireComparableMeta(baselineMeta, candidateMeta); err != nil {
		return inputError("baseline %s and candidate %s are not comparable: %v", *baselinePath, *candidatePath, err)
	}
	rows, failed := compareBenchmarks(baseline, candidate, thresholdPct, minDeltaNs)
	printReport(stdout, rows, *baselinePath, *candidatePath, thresholdPct, minDeltaNs, failed)
	if failed {
		return 1
	}
	return 0
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

type benchLogMeta struct {
	tokens map[string]string
	env    map[string]string
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
	return samples, meta, nil
}

func parseBenchLog(r io.Reader) (map[string][]float64, benchLogMeta, error) {
	samples := make(map[string][]float64)
	meta := benchLogMeta{env: make(map[string]string)}
	passed, completed := false, false
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		if fields[0] == "FAIL" || strings.HasPrefix(line, "--- FAIL:") || strings.HasPrefix(line, "panic:") || strings.HasPrefix(line, "fatal error:") || strings.HasPrefix(line, "exit status ") {
			return nil, meta, fmt.Errorf("failed benchmark run: %s", line)
		}
		if completed {
			return nil, meta, fmt.Errorf("unexpected output after completed benchmark run: %s", line)
		}
		if fields[0] == "BENCH-META" {
			if meta.tokens != nil || len(samples) != 0 || passed {
				return nil, meta, fmt.Errorf("BENCH-META must appear exactly once before benchmark samples")
			}
			meta.tokens = make(map[string]string)
			for _, field := range fields[1:] {
				key, value, ok := strings.Cut(field, "=")
				if !ok || key == "" || value == "" {
					return nil, meta, fmt.Errorf("malformed BENCH-META token %q; use non-empty whitespace-free key=value tokens", field)
				}
				if _, exists := meta.tokens[key]; exists {
					return nil, meta, fmt.Errorf("duplicate BENCH-META %s", key)
				}
				meta.tokens[key] = value
			}
			continue
		}
		switch fields[0] {
		case "goos:", "goarch:", "cpu:", "pkg:":
			key := strings.TrimSuffix(fields[0], ":")
			value := strings.TrimSpace(strings.TrimPrefix(line, fields[0]))
			if value == "" || (meta.env[key] != "" && meta.env[key] != value) || passed {
				return nil, meta, fmt.Errorf("empty, conflicting, or misplaced %s metadata: %q", key, line)
			}
			meta.env[key] = value
			continue
		case "PASS":
			if line != "PASS" || passed || len(samples) == 0 {
				return nil, meta, fmt.Errorf("unexpected PASS marker: %q", line)
			}
			passed = true
			continue
		case "ok":
			if !passed || len(fields) < 3 || fields[1] != meta.env["pkg"] {
				return nil, meta, fmt.Errorf("invalid package completion: %q", line)
			}
			duration, err := time.ParseDuration(fields[2])
			if err != nil || duration < 0 {
				return nil, meta, fmt.Errorf("invalid package elapsed time: %q", line)
			}
			completed = true
			continue
		}
		if !strings.HasPrefix(fields[0], "Benchmark") {
			continue
		}
		if passed || meta.tokens == nil {
			return nil, meta, fmt.Errorf("benchmark outside an active BENCH-META run: %q", line)
		}
		if len(fields) == 1 {
			continue
		}
		iterations, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil || iterations == 0 || len(fields) < 4 || len(fields)%2 != 0 {
			return nil, meta, fmt.Errorf("malformed benchmark measurement: %q", line)
		}
		foundNs := false
		for i := 3; i < len(fields); i += 2 {
			if fields[i] != "ns/op" {
				continue
			}
			if foundNs {
				return nil, meta, fmt.Errorf("duplicate ns/op metric: %q", line)
			}
			value, err := strconv.ParseFloat(fields[i-1], 64)
			if err != nil || math.IsNaN(value) || math.IsInf(value, 0) || value <= 0 {
				return nil, meta, fmt.Errorf("benchmark line %q has invalid ns/op value", line)
			}
			samples[fields[0]] = append(samples[fields[0]], value)
			foundNs = true
		}
		if !foundNs {
			return nil, meta, fmt.Errorf("benchmark measurement lacks ns/op: %q", line)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, meta, err
	}
	if err := requireValidMeta(meta); err != nil {
		return nil, meta, err
	}
	if !completed {
		return nil, meta, fmt.Errorf("incomplete benchmark run: require PASS followed by timed ok for pkg")
	}
	count, _ := strconv.Atoi(meta.tokens["count"])
	names := make([]string, 0, len(samples))
	for name := range samples {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if len(samples[name]) != count {
			return nil, meta, fmt.Errorf("benchmark %s has %d samples; BENCH-META count=%d", name, len(samples[name]), count)
		}
	}
	return samples, meta, nil
}

func requireValidMeta(meta benchLogMeta) error {
	for _, key := range []string{"count", "go", "target", "tree", "benchtime", "build"} {
		if value := meta.tokens[key]; value == "" || value == "(none)" || value == "unknown" {
			return fmt.Errorf("BENCH-META %s is missing or unknown", key)
		}
	}
	count, err := strconv.Atoi(meta.tokens["count"])
	if err != nil || count <= 0 {
		return fmt.Errorf("BENCH-META count must be a positive integer: %q", meta.tokens["count"])
	}
	build, err := hex.DecodeString(meta.tokens["build"])
	if err != nil || len(build) != 32 {
		return fmt.Errorf("BENCH-META build must be a SHA-256 fingerprint")
	}
	benchtime := meta.tokens["benchtime"]
	if strings.HasSuffix(benchtime, "x") {
		iterations, err := strconv.ParseUint(strings.TrimSuffix(benchtime, "x"), 10, 64)
		if err != nil || iterations == 0 {
			return fmt.Errorf("BENCH-META benchtime must specify a positive iteration count")
		}
	} else if duration, err := time.ParseDuration(benchtime); err != nil || duration <= 0 {
		return fmt.Errorf("BENCH-META benchtime must specify a positive duration")
	}
	for _, key := range []string{"goos", "goarch", "pkg"} {
		if meta.env[key] == "" {
			return fmt.Errorf("%s metadata is missing", key)
		}
	}
	return nil
}

func requireComparableMeta(baseline, candidate benchLogMeta) error {
	for _, meta := range []benchLogMeta{baseline, candidate} {
		if err := requireValidMeta(meta); err != nil {
			return err
		}
	}
	for _, item := range []struct {
		prefix              string
		baseline, candidate map[string]string
	}{
		{"BENCH-META ", baseline.tokens, candidate.tokens},
		{"", baseline.env, candidate.env},
	} {
		keys := make(map[string]bool)
		for key := range item.baseline {
			keys[key] = true
		}
		for key := range item.candidate {
			keys[key] = true
		}
		names := make([]string, 0, len(keys))
		for key := range keys {
			if item.prefix != "" && (key == "tree" || key == "date") {
				continue
			}
			names = append(names, key)
		}
		sort.Strings(names)
		for _, key := range names {
			if item.baseline[key] != item.candidate[key] {
				return fmt.Errorf("%s%s mismatch: baseline %q vs candidate %q", item.prefix, key, item.baseline[key], item.candidate[key])
			}
		}
	}
	return nil
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
