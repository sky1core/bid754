// verifylog is the execution-evidence gate for the canonical native
// verification targets. `go test` exits zero when a -run pattern matches no
// tests and `cargo test` exits zero when a test target contains no tests, so
// a renamed test function, a broken build-tag set, or a mangled -run regex
// can turn a full native gate into a green no-op. This command re-reads the
// captured gate log and fails unless the log carries the exact completion
// evidence promised by devtools/verification_anchors.json: the anchored
// executed/total comparison counts and the top-level PASS lines. It is a
// hand-maintained enforcement gate (like the marker-coverage and digest
// scripts), not part of any generated verification path.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
)

type anchors struct {
	Tier1ArithmeticStructured       map[string]uint64 `json:"tier1_arithmetic_long_structured_comparisons_by_width"`
	Tier1ArithmeticRandomCasesPerOp map[string]uint64 `json:"tier1_arithmetic_long_random_cases_per_operation_by_width"`
	Tier1ArithmeticRandomOperations uint64            `json:"tier1_arithmetic_long_random_operations"`
	Tier1CCComparisonStructured     map[string]uint64 `json:"tier1_compare_conversion_long_comparison_structured_by_width"`
	Tier1CCComparisonRandom         map[string]uint64 `json:"tier1_compare_conversion_long_comparison_random_by_width"`
	Tier1CCConversionStructured     uint64            `json:"tier1_compare_conversion_long_conversion_structured"`
	Tier1CCConversionRandom         uint64            `json:"tier1_compare_conversion_long_conversion_random"`
	ReadtestCasesTotal              uint64            `json:"readtest_cases_total"`
	ReadtestNativeCompareSkipCases  uint64            `json:"readtest_native_compare_skip_cases"`
	FFIBitcompareCasesTotal         uint64            `json:"ffi_bitcompare_cases_total"`
}

// sentinels mirrors the routing-sentinel row pin file; the row counts are
// len(rows) by design (no separate count scalar to drift).
type sentinels struct {
	Tier1ArithmeticRoutingRows        []string `json:"tier1_arithmetic_long_routing_sentinel_rows"`
	Tier1CompareConversionRoutingRows []string `json:"tier1_compare_conversion_long_routing_sentinel_rows"`
}

func main() {
	anchorsPath := flag.String("anchors", "verification_anchors.json", "path to verification_anchors.json")
	sentinelsPath := flag.String("sentinels", "verification_sentinels.json", "path to verification_sentinels.json (routing-sentinel row pins)")
	logPath := flag.String("log", "", "path to the captured gate log")
	domain := flag.String("domain", "", "gate domain: tier1-arithmetic-go, tier1-arithmetic-rust, tier1-compare-conversion-go, tier1-compare-conversion-rust, native-readtest, native-ffi")
	passes := flag.String("passes", "", "comma-separated top-level Go test names that must have '--- PASS:' evidence")
	flag.Parse()
	if *logPath == "" || (*domain == "" && *passes == "") {
		fmt.Fprintln(os.Stderr, "verifylog: -log plus -domain or -passes are required")
		os.Exit(2)
	}

	rawLog, err := os.ReadFile(*logPath)
	if err != nil {
		fail("read gate log: %v", err)
	}
	logLines := strings.Split(string(rawLog), "\n")

	required := []evidence{}
	for _, name := range strings.Split(*passes, ",") {
		if name = strings.TrimSpace(name); name != "" {
			required = append(required, topLevelPass(name))
		}
	}

	if *domain != "" {
		rawAnchors, err := os.ReadFile(*anchorsPath)
		if err != nil {
			fail("read anchors: %v", err)
		}
		var a anchors
		if err := json.Unmarshal(rawAnchors, &a); err != nil {
			fail("unmarshal anchors: %v", err)
		}
		widths := []string{"32", "64", "128"}
		switch *domain {
		case "tier1-arithmetic-go":
			required = append(required,
				topLevelPass("TestTier1ArithmeticCorpusContract"),
				topLevelPass("TestTier1ArithmeticRoutingSentinels"),
				topLevelPass("TestTier1ArithmeticStructuredNativeDifferential"),
				topLevelPass("TestTier1ArithmeticDeterministicRandomNativeDifferential"),
				sentinelCountEvidence(*sentinelsPath, "Tier 1 arithmetic routing sentinels"),
			)
			for _, w := range widths {
				structured := a.Tier1ArithmeticStructured["decimal"+w]
				random := a.Tier1ArithmeticRandomCasesPerOp["decimal"+w] * a.Tier1ArithmeticRandomOperations
				required = append(required,
					countLine(fmt.Sprintf("decimal%s structured exact comparisons: %d/%d", w, structured, structured)),
					countLine(fmt.Sprintf("decimal%s deterministic random exact comparisons: %d/%d", w, random, random)),
				)
			}
		case "tier1-arithmetic-rust":
			required = append(required,
				countLine("test result: ok. 4 passed; 0 failed;"),
				sentinelCountEvidence(*sentinelsPath, "Rust Tier 1 arithmetic routing sentinels"),
			)
			for _, w := range widths {
				structured := a.Tier1ArithmeticStructured["decimal"+w]
				random := a.Tier1ArithmeticRandomCasesPerOp["decimal"+w] * a.Tier1ArithmeticRandomOperations
				required = append(required,
					countLine(fmt.Sprintf("Rust Decimal%s structured Tier 1 exact comparisons: %d/%d", w, structured, structured)),
					countLine(fmt.Sprintf("Rust Decimal%s random Tier 1 exact comparisons: %d/%d", w, random, random)),
				)
			}
		case "tier1-compare-conversion-go":
			required = append(required,
				topLevelPass("TestTier1QuietComparisonSemanticMatrix"),
				topLevelPass("TestTier1CompareConversionRoutingSentinels"),
				topLevelPass("TestTier1ComparisonMinMaxStructuredNativeDifferential"),
				topLevelPass("TestTier1ComparisonMinMaxDeterministicRandomNativeDifferential"),
				topLevelPass("TestTier1ConversionStructuredNativeDifferential"),
				topLevelPass("TestTier1ConversionDeterministicRandomNativeDifferential"),
				sentinelCCCountEvidence(*sentinelsPath, "Tier 1 compare/conversion routing sentinels"),
				countLine(fmt.Sprintf("Tier 1 structured conversion exact comparisons: %d/%d", a.Tier1CCConversionStructured, a.Tier1CCConversionStructured)),
				countLine(fmt.Sprintf("Tier 1 deterministic random conversion exact comparisons: %d/%d", a.Tier1CCConversionRandom, a.Tier1CCConversionRandom)),
			)
			for _, w := range widths {
				structured := a.Tier1CCComparisonStructured["decimal"+w]
				random := a.Tier1CCComparisonRandom["decimal"+w]
				required = append(required,
					countLine(fmt.Sprintf("decimal%s structured compare/minmax exact comparisons: %d/%d", w, structured, structured)),
					countLine(fmt.Sprintf("decimal%s random compare/minmax exact comparisons: %d/%d", w, random, random)),
				)
			}
		case "tier1-compare-conversion-rust":
			required = append(required,
				countLine("test result: ok. 7 passed; 0 failed;"),
				sentinelCCCountEvidence(*sentinelsPath, "Rust Tier 1 compare/conversion routing sentinels"),
				countLine(fmt.Sprintf("Rust structured Tier 1 conversion exact comparisons: %d/%d;", a.Tier1CCConversionStructured, a.Tier1CCConversionStructured)),
				countLine(fmt.Sprintf("Rust deterministic random Tier 1 conversion exact comparisons: %d/%d", a.Tier1CCConversionRandom, a.Tier1CCConversionRandom)),
			)
			for _, w := range widths {
				structured := a.Tier1CCComparisonStructured["decimal"+w]
				random := a.Tier1CCComparisonRandom["decimal"+w]
				required = append(required,
					countLine(fmt.Sprintf("Rust Decimal%s structured compare/minmax: %d/%d", w, structured, structured)),
					countLine(fmt.Sprintf("Rust Decimal%s random compare/minmax: %d/%d", w, random, random)),
				)
			}
		case "native-readtest":
			nativeEvidence, err := nativeReadtestEvidence(a)
			if err != nil {
				fail("native readtest evidence: %v", err)
			}
			required = append(required, nativeEvidence...)
		case "native-ffi":
			nativeEvidence, err := nativeFFIEvidence(a)
			if err != nil {
				fail("native FFI evidence: %v", err)
			}
			required = append(required, nativeEvidence...)
		default:
			fail("unknown domain %q", *domain)
		}
	}

	missing := missingEvidence(logLines, required)
	if len(missing) > 0 {
		fmt.Fprintf(os.Stderr, "verifylog: gate log %s is missing required execution evidence (the gate may have run zero tests or a reduced corpus):\n", *logPath)
		for _, want := range missing {
			fmt.Fprintf(os.Stderr, "  missing: %q\n", want)
		}
		os.Exit(1)
	}
	fmt.Printf("verifylog: %s carries all %d required execution-evidence lines\n", *logPath, len(required))
}

// evidence is one required log line. kind "pass" requires an unindented
// top-level `--- PASS: <name> (` line so an indented subtest PASS under a
// failing parent cannot satisfy it; kind "count" requires the literal followed
// by a non-digit so an anchored total cannot match a longer number; kind
// "compact-summary" requires exactly one complete summary line for its root.
type evidence struct {
	kind     string
	literal  string
	rootTest string
}

func topLevelPass(name string) evidence { return evidence{kind: "pass", literal: name} }
func countLine(literal string) evidence { return evidence{kind: "count", literal: literal} }
func compactSummary(rootTest, literal string) evidence {
	return evidence{kind: "compact-summary", literal: literal, rootTest: rootTest}
}

func nativeReadtestEvidence(a anchors) ([]evidence, error) {
	return compactGoSubtestEvidence(
		"TestGeneratedReadCases",
		"native readtest",
		a.ReadtestCasesTotal,
		a.ReadtestNativeCompareSkipCases,
	)
}

func nativeFFIEvidence(a anchors) ([]evidence, error) {
	return compactGoSubtestEvidence(
		"TestGeneratedFFIBitCompareSubset",
		"native FFI",
		a.FFIBitcompareCasesTotal,
		0,
	)
}

func compactGoSubtestEvidence(rootTest, label string, total, skips uint64) ([]evidence, error) {
	if total == 0 {
		return nil, fmt.Errorf("%s total cases must be positive", label)
	}
	if skips > total {
		return nil, fmt.Errorf("%s skipped cases %d exceeds total %d", label, skips, total)
	}
	passes := total - skips
	suppressed := total + passes + skips
	summary := fmt.Sprintf(
		"testlogcompact: suppressed %d subtest lifecycle lines (run=%d pass=%d skip=%d) for %s",
		suppressed,
		total,
		passes,
		skips,
		rootTest,
	)
	return []evidence{
		topLevelPass(rootTest),
		compactSummary(rootTest, summary),
	}, nil
}

// loadSentinels reads the routing-sentinel row pin file for the count
// evidence below.
func loadSentinels(sentinelsPath string) sentinels {
	rawSentinels, err := os.ReadFile(sentinelsPath)
	if err != nil {
		fail("read sentinels: %v", err)
	}
	var s sentinels
	if err := json.Unmarshal(rawSentinels, &s); err != nil {
		fail("unmarshal sentinels: %v", err)
	}
	return s
}

// sentinelCountEvidence loads the routing-sentinel row pin and requires the
// runner's "<prefix>: N/N" full-count line, N = len(pinned rows). Zero pinned
// rows fail immediately: an empty pin would turn the sentinel gate into a
// green no-op.
func sentinelCountEvidence(sentinelsPath, prefix string) evidence {
	n := len(loadSentinels(sentinelsPath).Tier1ArithmeticRoutingRows)
	if n == 0 {
		fail("verification_sentinels.json pins zero Tier 1 arithmetic routing sentinel rows")
	}
	return countLine(fmt.Sprintf("%s: %d/%d", prefix, n, n))
}

// sentinelCCCountEvidence is the compare/conversion analogue of
// sentinelCountEvidence.
func sentinelCCCountEvidence(sentinelsPath, prefix string) evidence {
	n := len(loadSentinels(sentinelsPath).Tier1CompareConversionRoutingRows)
	if n == 0 {
		fail("verification_sentinels.json pins zero Tier 1 compare/conversion routing sentinel rows")
	}
	return countLine(fmt.Sprintf("%s: %d/%d", prefix, n, n))
}

func (e evidence) String() string {
	if e.kind == "pass" {
		return "--- PASS: " + e.literal
	}
	return e.literal
}

func (e evidence) matchesLine(line string) bool {
	switch e.kind {
	case "pass":
		return strings.HasPrefix(line, "--- PASS: "+e.literal+" (")
	case "compact-summary":
		return line == e.literal
	default:
		idx := strings.Index(line, e.literal)
		if idx < 0 {
			return false
		}
		rest := line[idx+len(e.literal):]
		return rest == "" || rest[0] < '0' || rest[0] > '9'
	}
}

func (e evidence) isSatisfiedBy(logLines []string) bool {
	if e.kind != "compact-summary" {
		for _, line := range logLines {
			if e.matchesLine(line) {
				return true
			}
		}
		return false
	}

	candidates := 0
	exact := 0
	suffix := " for " + e.rootTest
	for _, line := range logLines {
		if strings.HasPrefix(line, "testlogcompact: suppressed ") && strings.HasSuffix(line, suffix) {
			candidates++
			if e.matchesLine(line) {
				exact++
			}
		}
	}
	return candidates == 1 && exact == 1
}

func missingEvidence(logLines []string, required []evidence) []string {
	missing := []string{}
	for _, want := range required {
		if !want.isSatisfiedBy(logLines) {
			missing = append(missing, want.String())
		}
	}
	return missing
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "verifylog: "+format+"\n", args...)
	os.Exit(1)
}
