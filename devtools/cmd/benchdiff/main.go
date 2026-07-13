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
//   - improvements are reported but never fail.
//
// Exit codes: 0 pass, 1 regression or vanished benchmark, 2 usage/input
// errors (unreadable log, no benchmark rows, bad threshold).
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
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

	baseline, err := parseBenchLogFile(*baselinePath)
	if err != nil {
		fatal("baseline: %v", err)
	}
	candidate, err := parseBenchLogFile(*candidatePath)
	if err != nil {
		fatal("candidate: %v", err)
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
// the raw environment value ("" means unset → default).
func regressionThresholdPct(raw string) (float64, error) {
	if raw == "" {
		return defaultRegressionThresholdPct, nil
	}
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, fmt.Errorf("%s=%q is not a number", thresholdEnvVar, raw)
	}
	if value < 0 {
		return 0, fmt.Errorf("%s=%q must be >= 0 (percent)", thresholdEnvVar, raw)
	}
	return value, nil
}

func parseBenchLogFile(path string) (map[string][]float64, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	samples, err := parseBenchLog(file)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if len(samples) == 0 {
		return nil, fmt.Errorf("%s contains no `go test -bench` result lines", path)
	}
	return samples, nil
}

// parseBenchLog collects every ns/op sample per benchmark name (the full
// name including the -GOMAXPROCS suffix, so runs from a different parallelism
// setting never silently pair up) from a `go test -bench` log.
func parseBenchLog(r io.Reader) (map[string][]float64, error) {
	samples := make(map[string][]float64)
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 4 || !strings.HasPrefix(fields[0], "Benchmark") {
			continue
		}
		for i := 2; i < len(fields); i++ {
			if fields[i] != "ns/op" {
				continue
			}
			value, err := strconv.ParseFloat(fields[i-1], 64)
			if err != nil {
				return nil, fmt.Errorf("benchmark line %q has unparsable ns/op value: %v", scanner.Text(), err)
			}
			if value <= 0 {
				return nil, fmt.Errorf("benchmark line %q has non-positive ns/op value", scanner.Text())
			}
			samples[fields[0]] = append(samples[fields[0]], value)
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return samples, nil
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
