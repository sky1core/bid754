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
	"sort"
	"strings"
)

type anchors struct {
	Tier1ArithmeticStructured           map[string]uint64 `json:"tier1_arithmetic_long_structured_comparisons_by_width"`
	Tier1ArithmeticRandomCasesPerOp     map[string]uint64 `json:"tier1_arithmetic_long_random_cases_per_operation_by_width"`
	Tier1ArithmeticRandomOperations     uint64            `json:"tier1_arithmetic_long_random_operations"`
	Tier1ArithmeticStructuredDeviations *uint64           `json:"tier1_arithmetic_long_intel003_structured_deviations"`
	Tier1ArithmeticRandomDeviations     *uint64           `json:"tier1_arithmetic_long_intel003_random_deviations"`
	Tier1CCComparisonStructured         map[string]uint64 `json:"tier1_compare_conversion_long_comparison_structured_by_width"`
	Tier1CCComparisonRandom             map[string]uint64 `json:"tier1_compare_conversion_long_comparison_random_by_width"`
	Tier1CCConversionStructured         uint64            `json:"tier1_compare_conversion_long_conversion_structured"`
	Tier1CCConversionRandom             uint64            `json:"tier1_compare_conversion_long_conversion_random"`
	ReadtestCasesTotal                  uint64            `json:"readtest_cases_total"`
	ReadtestNativeCompareSkipCases      uint64            `json:"readtest_native_compare_skip_cases"`
	GoportReadtestExecutedCases         uint64            `json:"goport_readtest_executed_cases"`
	DectestSuiteCases                   map[string]uint64 `json:"dectest_suite_cases"`
	NativeDectestSkippedCases           map[string]uint64 `json:"native_dectest_skipped_cases"`
	GoportDectestExecutedCases          map[string]uint64 `json:"goport_dectest_executed_cases"`
	GoportDectestSkippedCases           map[string]uint64 `json:"goport_dectest_skipped_cases"`
	GoportDectestFlagExemptCases        map[string]uint64 `json:"goport_dectest_flag_exempt_cases"`
	FFIBitcompareCasesTotal             uint64            `json:"ffi_bitcompare_cases_total"`
	FFIQuantumSteeringDeviations        *uint64           `json:"ffi_intel006_independent_cases"`
	DecnumberDiffStructured             map[string]uint64 `json:"decnumber_differential_structured_comparisons_by_width"`
	DecnumberDiffStructuredExcluded     map[string]uint64 `json:"decnumber_differential_structured_fma_excluded_by_width"`
	DecnumberDiffStructuredKnown        map[string]uint64 `json:"decnumber_differential_structured_known_divergences_by_width"`
	DecnumberDiffRandom                 map[string]uint64 `json:"decnumber_differential_random_comparisons_by_width"`
	DecnumberDiffRandomExcluded         map[string]uint64 `json:"decnumber_differential_random_fma_excluded_by_width"`
	DecnumberDiffRandomKnown            map[string]uint64 `json:"decnumber_differential_random_known_divergences_by_width"`
	D32ExhaustiveLanes                  uint64            `json:"d32_exhaustive_unary_lanes"`
	D32ExhaustiveCasesPerLane           uint64            `json:"d32_exhaustive_unary_cases_per_lane"`
	D32ExhaustiveTotalComparisons       uint64            `json:"d32_exhaustive_unary_total_comparisons"`
	D32ExhaustiveDigestByLane           map[string]uint64 `json:"d32_exhaustive_unary_result_digest_by_lane"`
}

// sentinels mirrors the routing-sentinel row pin file; the row counts are
// len(rows) by design (no separate count scalar to drift).
type sentinels struct {
	Tier1ArithmeticRoutingRows        []string `json:"tier1_arithmetic_long_routing_sentinel_rows"`
	Tier1CompareConversionRoutingRows []string `json:"tier1_compare_conversion_long_routing_sentinel_rows"`
	MixedFormatFFIRoutingRows         []string `json:"mixed_format_ffi_routing_sentinel_rows"`
	DecnumberDifferentialRows         []string `json:"decnumber_differential_sentinel_rows"`
	D32ExhaustiveRows                 []string `json:"d32_exhaustive_sentinel_rows"`
}

func main() {
	anchorsPath := flag.String("anchors", "verification_anchors.json", "path to verification_anchors.json")
	sentinelsPath := flag.String("sentinels", "verification_sentinels.json", "path to verification_sentinels.json (routing-sentinel row pins)")
	logPath := flag.String("log", "", "path to the captured gate log")
	domain := flag.String("domain", "", "gate domain: tier1-arithmetic-go, tier1-arithmetic-rust, tier1-compare-conversion-go, tier1-compare-conversion-rust, goport-readtest, native-readtest, native-dectest, goport-dectest, native-ffi, decnumber-differential, d32-exhaustive, d32-exhaustive-rust")
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
			arithmeticEvidence, err := tier1ArithmeticEvidence(a, *sentinelsPath, false)
			if err != nil {
				fail("Tier 1 arithmetic evidence: %v", err)
			}
			required = append(required, arithmeticEvidence...)
		case "tier1-arithmetic-rust":
			arithmeticEvidence, err := tier1ArithmeticEvidence(a, *sentinelsPath, true)
			if err != nil {
				fail("Tier 1 arithmetic evidence: %v", err)
			}
			required = append(required, arithmeticEvidence...)
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
		case "native-dectest":
			nativeEvidence, err := nativeDectestEvidence(a)
			if err != nil {
				fail("native decTest evidence: %v", err)
			}
			required = append(required, nativeEvidence...)
		case "goport-dectest":
			goportEvidence, err := goportDectestEvidence(a)
			if err != nil {
				fail("Go-port decTest evidence: %v", err)
			}
			required = append(required, goportEvidence...)
		case "goport-readtest":
			goportEvidence, err := goportReadtestEvidence(a)
			if err != nil {
				fail("Go-port readtest evidence: %v", err)
			}
			required = append(required, goportEvidence...)
		case "native-ffi":
			nativeEvidence, err := nativeFFIEvidence(a, *sentinelsPath)
			if err != nil {
				fail("native FFI evidence: %v", err)
			}
			required = append(required, nativeEvidence...)
		case "decnumber-differential":
			required = append(required,
				topLevelPass("TestGeneratedDecnumberDifferentialCorpusContract"),
				topLevelPass("TestGeneratedDecnumberDifferentialRoutingSentinels"),
				topLevelPass("TestGeneratedDecnumberDifferentialStructured"),
				topLevelPass("TestGeneratedDecnumberDifferentialDeterministicRandom"),
				decnumberDiffSentinelCountEvidence(*sentinelsPath, "decNumber differential routing sentinels"),
			)
			for _, w := range widths {
				structured := a.DecnumberDiffStructured["decimal"+w]
				structuredKnown := a.DecnumberDiffStructuredKnown["decimal"+w]
				random := a.DecnumberDiffRandom["decimal"+w]
				randomKnown := a.DecnumberDiffRandomKnown["decimal"+w]
				if structured == 0 || random == 0 {
					fail("decNumber differential anchors carry a zero comparison count for decimal%s (an empty gate would be a green no-op)", w)
				}
				if structuredKnown > structured || randomKnown > random {
					fail("decNumber differential known-divergence anchors exceed the comparison totals for decimal%s", w)
				}
				required = append(required,
					countLine(fmt.Sprintf("decimal%s decnumber differential structured exact comparisons: %d/%d", w, structured-structuredKnown, structured-structuredKnown)),
					countLine(fmt.Sprintf("decimal%s decnumber differential structured excluded fma_zero_inf_qnan_invalid_ieee_optional: %d", w, a.DecnumberDiffStructuredExcluded["decimal"+w])),
					countLine(fmt.Sprintf("decimal%s decnumber differential structured known divergences: %d/%d", w, structuredKnown, structuredKnown)),
					countLine(fmt.Sprintf("decimal%s decnumber differential random exact comparisons: %d/%d", w, random-randomKnown, random-randomKnown)),
					countLine(fmt.Sprintf("decimal%s decnumber differential random excluded fma_zero_inf_qnan_invalid_ieee_optional: %d", w, a.DecnumberDiffRandomExcluded["decimal"+w])),
					countLine(fmt.Sprintf("decimal%s decnumber differential random known divergences: %d/%d", w, randomKnown, randomKnown)),
				)
			}
		case "d32-exhaustive":
			required = append(required,
				topLevelPass("TestGeneratedD32ExhaustiveLaneContract"),
				topLevelPass("TestGeneratedD32ExhaustiveRoutingSentinels"),
				topLevelPass("TestGeneratedD32ExhaustiveUnaryDifferential"),
				d32ExhaustiveSentinelCountEvidence(*sentinelsPath, "d32 exhaustive routing sentinels"),
			)
			required = append(required, d32ExhaustiveDigestEvidence(a, "")...)
		case "d32-exhaustive-rust":
			// The generated Rust leg binds to the SAME hand-pinned per-lane
			// digests. Each runner is checked against the pins independently
			// (this command reads one log at a time and never compares the two
			// logs), and an in-run C-vs-port divergence fails its runner before
			// any digest is printed. What the shared pins add is that the Rust
			// port must reproduce, over the whole space, the digests a Go-port
			// run established — so a go2rs-side regression cannot be pinned
			// away on its own.
			required = append(required,
				countLine("test result: ok. 3 passed; 0 failed;"),
				d32ExhaustiveSentinelCountEvidence(*sentinelsPath, "Rust d32 exhaustive routing sentinels"),
			)
			required = append(required, d32ExhaustiveDigestEvidence(a, "Rust ")...)
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
// failing parent cannot satisfy it; kind "count" requires the literal's first
// occurrence in a line to be followed by a non-digit so an anchored total
// cannot match a longer number, and — when rejectPrefix is set — not to be
// preceded by that prefix; kind "compact-summary" requires exactly one
// complete summary line for its root.
type evidence struct {
	kind     string
	literal  string
	rootTest string
	// rejectPrefix, when set, disqualifies an occurrence of literal that is
	// immediately preceded by it. The d32 exhaustive Go-leg lines are a
	// literal substring of the Rust leg's ("Rust " + the same text), so
	// without this the Rust log would satisfy the Go domain's digest
	// evidence; leg separation must not rest on the Go domain's unrelated
	// top-level PASS requirements.
	rejectPrefix string
}

func topLevelPass(name string) evidence { return evidence{kind: "pass", literal: name} }
func countLine(literal string) evidence { return evidence{kind: "count", literal: literal} }
func rustTestRun(name string) evidence {
	return evidence{kind: "rust-test", literal: "test " + name + " ... "}
}

// countLineRejectingPrefix is countLine restricted to occurrences that are
// not immediately preceded by reject.
func countLineRejectingPrefix(literal, reject string) evidence {
	return evidence{kind: "count", literal: literal, rejectPrefix: reject}
}
func compactSummary(rootTest, literal string) evidence {
	return evidence{kind: "compact-summary", literal: literal, rootTest: rootTest}
}

func tier1ArithmeticEvidence(a anchors, sentinelsPath string, rust bool) ([]evidence, error) {
	if a.Tier1ArithmeticStructuredDeviations == nil || a.Tier1ArithmeticRandomDeviations == nil {
		return nil, fmt.Errorf("INTEL-BID-003 deviation anchors are missing")
	}
	structuredDeviations := *a.Tier1ArithmeticStructuredDeviations
	randomDeviations := *a.Tier1ArithmeticRandomDeviations
	if a.Tier1ArithmeticRandomOperations == 0 {
		return nil, fmt.Errorf("Tier 1 arithmetic random operation count must be positive")
	}
	for _, w := range []string{"32", "64", "128"} {
		key := "decimal" + w
		structured := a.Tier1ArithmeticStructured[key]
		perOp := a.Tier1ArithmeticRandomCasesPerOp[key]
		if structured == 0 || perOp == 0 || perOp > ^uint64(0)/a.Tier1ArithmeticRandomOperations {
			return nil, fmt.Errorf("invalid Tier 1 arithmetic comparison anchors for %s", key)
		}
	}
	structured128 := a.Tier1ArithmeticStructured["decimal128"]
	random128 := a.Tier1ArithmeticRandomCasesPerOp["decimal128"] * a.Tier1ArithmeticRandomOperations
	if structuredDeviations == 0 || structuredDeviations > structured128 || randomDeviations > random128 {
		return nil, fmt.Errorf("invalid INTEL-BID-003 deviation anchors")
	}

	required := []evidence{countLine("INTEL-BID-003 witness: C comparisons=90 independent deviations=90")}
	if rust {
		required = append(required,
			countLine("test result: ok. 5 passed; 0 failed;"),
			rustTestRun("tier1_arithmetic_scaleb_intel003_witness"),
			sentinelCountEvidence(sentinelsPath, "Rust Tier 1 arithmetic routing sentinels"),
		)
	} else {
		required = append(required,
			topLevelPass("TestTier1ArithmeticCorpusContract"),
			topLevelPass("TestTier1ArithmeticRoutingSentinels"),
			topLevelPass("TestTier1ArithmeticScaleBIntel003Witness"),
			topLevelPass("TestTier1ArithmeticStructuredNativeDifferential"),
			topLevelPass("TestTier1ArithmeticDeterministicRandomNativeDifferential"),
			sentinelCountEvidence(sentinelsPath, "Tier 1 arithmetic routing sentinels"),
		)
	}
	for _, w := range []string{"32", "64"} {
		structured := a.Tier1ArithmeticStructured["decimal"+w]
		random := a.Tier1ArithmeticRandomCasesPerOp["decimal"+w] * a.Tier1ArithmeticRandomOperations
		if rust {
			required = append(required,
				countLine(fmt.Sprintf("Rust Decimal%s structured Tier 1 exact comparisons: %d/%d", w, structured, structured)),
				countLine(fmt.Sprintf("Rust Decimal%s random Tier 1 exact comparisons: %d/%d", w, random, random)),
			)
		} else {
			required = append(required,
				countLine(fmt.Sprintf("decimal%s structured exact comparisons: %d/%d", w, structured, structured)),
				countLine(fmt.Sprintf("decimal%s deterministic random exact comparisons: %d/%d", w, random, random)),
			)
		}
	}
	if rust {
		required = append(required,
			countLine(fmt.Sprintf("Rust Decimal128 structured Tier 1 comparisons: C exact=%d independent INTEL-BID-003=%d total=%d/%d", structured128-structuredDeviations, structuredDeviations, structured128, structured128)),
			countLine(fmt.Sprintf("Rust Decimal128 random Tier 1 comparisons: C exact=%d independent INTEL-BID-003=%d total=%d/%d", random128-randomDeviations, randomDeviations, random128, random128)),
		)
	} else {
		required = append(required,
			countLine(fmt.Sprintf("decimal128 structured comparisons: C exact=%d independent INTEL-BID-003=%d total=%d/%d", structured128-structuredDeviations, structuredDeviations, structured128, structured128)),
			countLine(fmt.Sprintf("decimal128 deterministic random comparisons: C exact=%d independent INTEL-BID-003=%d total=%d/%d", random128-randomDeviations, randomDeviations, random128, random128)),
		)
	}
	return required, nil
}

func nativeReadtestEvidence(a anchors) ([]evidence, error) {
	return compactGoSubtestEvidence(
		"TestGeneratedReadCases",
		"native readtest",
		a.ReadtestCasesTotal,
		a.ReadtestNativeCompareSkipCases,
	)
}

func nativeDectestEvidence(a anchors) ([]evidence, error) {
	suites := []string{"Decimal32", "Decimal64", "Decimal128", "General"}
	if len(a.DectestSuiteCases) != len(suites) || len(a.NativeDectestSkippedCases) != len(suites) {
		return nil, fmt.Errorf("native decTest requires case and skip anchors for all four suites")
	}
	required := []evidence{
		topLevelPass("TestGeneratedDectestSuites"),
		topLevelPass("TestGeneratedDectestPlusMinusQuantumStrengthGoPort"),
		topLevelPass("TestGeneratedDectestPlusMinusQuantumStrengthUnaryAdapter"),
		topLevelPass("TestNativeDectestRunnerRejectsPartialExecution"),
	}
	for _, suite := range suites {
		cases := a.DectestSuiteCases[suite]
		skipped, ok := a.NativeDectestSkippedCases[suite]
		if cases == 0 || !ok || skipped >= cases {
			return nil, fmt.Errorf("native decTest %s requires positive executed coverage below the raw case count", suite)
		}
		required = append(required, countLine(fmt.Sprintf("native decTest %s: passed=%d failed=0 skipped=%d total=%d", suite, cases-skipped, skipped, cases)))
	}
	return required, nil
}

func goportDectestEvidence(a anchors) ([]evidence, error) {
	suites := []string{"Decimal32", "Decimal64", "Decimal128"}
	if len(a.DectestSuiteCases) != 4 || len(a.GoportDectestExecutedCases) != len(suites) || len(a.GoportDectestSkippedCases) != len(suites) || len(a.GoportDectestFlagExemptCases) != len(suites) {
		return nil, fmt.Errorf("Go-port decTest requires case, execution, skip and flag-exemption anchors")
	}
	required := []evidence{
		topLevelPass("TestGeneratedDectestSuitesGoPort"),
		topLevelPass("TestGeneratedDectestPlusMinusQuantumStrengthGoPort"),
		topLevelPass("TestGeneratedDectestPlusMinusQuantumStrengthUnaryAdapter"),
		topLevelPass("TestGoportDectestRunnerRejectsPartialExecution"),
	}
	for _, suite := range suites {
		cases := a.DectestSuiteCases[suite]
		executed := a.GoportDectestExecutedCases[suite]
		skipped, hasSkipped := a.GoportDectestSkippedCases[suite]
		exempt, hasExempt := a.GoportDectestFlagExemptCases[suite]
		if executed == 0 || executed > cases || !hasSkipped || skipped != cases-executed || !hasExempt || exempt > executed {
			return nil, fmt.Errorf("Go-port decTest %s requires positive and consistent execution coverage", suite)
		}
		required = append(required, countLine(fmt.Sprintf("goport decTest %s: executed=%d failed=0 skipped=%d flagExempt=%d total=%d", suite, executed, skipped, exempt, cases)))
	}
	return required, nil
}

func goportReadtestEvidence(a anchors) ([]evidence, error) {
	return compactGoSubtestEvidence(
		"TestGeneratedReadCasesGoPort",
		"Go-port readtest",
		a.GoportReadtestExecutedCases,
		0,
	)
}

func nativeFFIEvidence(a anchors, sentinelsPath string) ([]evidence, error) {
	if a.FFIQuantumSteeringDeviations == nil || *a.FFIQuantumSteeringDeviations == 0 || *a.FFIQuantumSteeringDeviations >= a.FFIBitcompareCasesTotal {
		return nil, fmt.Errorf("native FFI independent quantum count must be positive and below the total")
	}
	required, err := compactGoSubtestEvidence(
		"TestGeneratedFFIBitCompareSubset",
		"native FFI",
		a.FFIBitcompareCasesTotal,
		0,
	)
	if err != nil {
		return nil, err
	}
	// The mixed-format FFI operand-swap routing sentinels run in the same gate
	// as a separate top-level test; bind their top-level PASS and full-count
	// log line so the gate cannot silently drop them.
	required = append(required,
		topLevelPass("TestGeneratedMixedFormatFFIRoutingSentinels"),
		topLevelPass("TestGeneratedFFIQuantumSteeringAdjudicationStrength"),
		topLevelPass("TestGeneratedFFIQuantumSteeringSamples"),
		countLine(fmt.Sprintf("FFI comparisons: C exact=%d independent INTEL-BID-006=%d total=%d/%d", a.FFIBitcompareCasesTotal-*a.FFIQuantumSteeringDeviations, *a.FFIQuantumSteeringDeviations, a.FFIBitcompareCasesTotal, a.FFIBitcompareCasesTotal)),
		mixedFFIRoutingSentinelCountEvidence(sentinelsPath, "mixed-format FFI routing sentinels"),
	)
	return required, nil
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

// mixedFFIRoutingSentinelCountEvidence is the mixed-format FFI routing
// analogue of sentinelCountEvidence.
func mixedFFIRoutingSentinelCountEvidence(sentinelsPath, prefix string) evidence {
	n := len(loadSentinels(sentinelsPath).MixedFormatFFIRoutingRows)
	if n == 0 {
		fail("verification_sentinels.json pins zero mixed-format FFI routing sentinel rows")
	}
	return countLine(fmt.Sprintf("%s: %d/%d", prefix, n, n))
}

// decnumberDiffSentinelCountEvidence is the decNumber differential analogue
// of sentinelCountEvidence.
func decnumberDiffSentinelCountEvidence(sentinelsPath, prefix string) evidence {
	n := len(loadSentinels(sentinelsPath).DecnumberDifferentialRows)
	if n == 0 {
		fail("verification_sentinels.json pins zero decNumber differential sentinel rows")
	}
	return countLine(fmt.Sprintf("%s: %d/%d", prefix, n, n))
}

// d32ExhaustiveSentinelCountEvidence is the d32 exhaustive analogue of
// sentinelCountEvidence. The Go leg's line is a substring of the Rust leg's
// ("Rust " + the same text), so — like the per-lane digest evidence — the Go
// leg rejects occurrences carrying the Rust prefix; otherwise a Rust log
// would satisfy this row of the Go domain.
func d32ExhaustiveSentinelCountEvidence(sentinelsPath, prefix string) evidence {
	n := len(loadSentinels(sentinelsPath).D32ExhaustiveRows)
	if n == 0 {
		fail("verification_sentinels.json pins zero d32 exhaustive sentinel rows")
	}
	literal := fmt.Sprintf("%s: %d/%d", prefix, n, n)
	if !strings.HasPrefix(prefix, "Rust ") {
		return countLineRejectingPrefix(literal, "Rust ")
	}
	return countLine(literal)
}

// d32ExhaustiveDigestEvidence validates the d32 exhaustive anchors and
// returns the per-lane digest and full-count evidence lines shared by the Go
// leg (legPrefix "") and the generated Rust leg (legPrefix "Rust "). Both
// legs bind to the identical hand-pinned digest values.
func d32ExhaustiveDigestEvidence(a anchors, legPrefix string) []evidence {
	if a.D32ExhaustiveLanes == 0 || a.D32ExhaustiveCasesPerLane == 0 {
		fail("d32 exhaustive anchors carry a zero lane or per-lane case count (an empty gate would be a green no-op)")
	}
	if uint64(len(a.D32ExhaustiveDigestByLane)) != a.D32ExhaustiveLanes {
		fail("d32 exhaustive anchors pin %d lane digests for %d lanes", len(a.D32ExhaustiveDigestByLane), a.D32ExhaustiveLanes)
	}
	if a.D32ExhaustiveLanes*a.D32ExhaustiveCasesPerLane != a.D32ExhaustiveTotalComparisons {
		fail("d32 exhaustive anchors: lanes %d x cases-per-lane %d != total %d",
			a.D32ExhaustiveLanes, a.D32ExhaustiveCasesPerLane, a.D32ExhaustiveTotalComparisons)
	}
	laneNames := make([]string, 0, len(a.D32ExhaustiveDigestByLane))
	for lane := range a.D32ExhaustiveDigestByLane {
		laneNames = append(laneNames, lane)
	}
	sort.Strings(laneNames)
	// The Go leg's lines are a substring of the Rust leg's, so the Go leg
	// rejects occurrences carrying the Rust prefix; that keeps the two
	// domains mutually exclusive on the digest evidence itself.
	line := countLine
	if legPrefix == "" {
		line = func(literal string) evidence { return countLineRejectingPrefix(literal, "Rust ") }
	}
	required := []evidence{}
	for _, lane := range laneNames {
		digest := a.D32ExhaustiveDigestByLane[lane]
		if digest == 0 {
			fail("d32 exhaustive anchors pin a zero result digest for lane %q (an unpinned digest binds nothing)", lane)
		}
		required = append(required, line(fmt.Sprintf(
			"%sdecimal32 exhaustive lane %s: exact comparisons %d/%d digest=%d",
			legPrefix, lane, a.D32ExhaustiveCasesPerLane, a.D32ExhaustiveCasesPerLane, digest)))
	}
	required = append(required, line(fmt.Sprintf(
		"%sdecimal32 exhaustive unary total comparisons: %d/%d",
		legPrefix, a.D32ExhaustiveTotalComparisons, a.D32ExhaustiveTotalComparisons)))
	return required
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
	case "rust-test":
		return strings.HasPrefix(line, e.literal)
	case "compact-summary":
		return line == e.literal
	default:
		idx := strings.Index(line, e.literal)
		if idx < 0 {
			return false
		}
		if e.rejectPrefix != "" && strings.HasSuffix(line[:idx], e.rejectPrefix) {
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
