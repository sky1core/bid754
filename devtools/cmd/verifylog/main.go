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
}

func main() {
	anchorsPath := flag.String("anchors", "verification_anchors.json", "path to verification_anchors.json")
	logPath := flag.String("log", "", "path to the captured gate log")
	domain := flag.String("domain", "", "gate domain: tier1-arithmetic-go, tier1-arithmetic-rust, tier1-compare-conversion-go, tier1-compare-conversion-rust")
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
				topLevelPass("TestTier1ArithmeticStructuredNativeDifferential"),
				topLevelPass("TestTier1ArithmeticDeterministicRandomNativeDifferential"),
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
			required = append(required, countLine("test result: ok. 3 passed; 0 failed;"))
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
				topLevelPass("TestTier1ComparisonMinMaxStructuredNativeDifferential"),
				topLevelPass("TestTier1ComparisonMinMaxDeterministicRandomNativeDifferential"),
				topLevelPass("TestTier1ConversionStructuredNativeDifferential"),
				topLevelPass("TestTier1ConversionDeterministicRandomNativeDifferential"),
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
				countLine("test result: ok. 6 passed; 0 failed;"),
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
// by a non-digit so an anchored total cannot match a longer number.
type evidence struct {
	kind    string
	literal string
}

func topLevelPass(name string) evidence { return evidence{kind: "pass", literal: name} }
func countLine(literal string) evidence { return evidence{kind: "count", literal: literal} }

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
	default:
		idx := strings.Index(line, e.literal)
		if idx < 0 {
			return false
		}
		rest := line[idx+len(e.literal):]
		return rest == "" || rest[0] < '0' || rest[0] > '9'
	}
}

func missingEvidence(logLines []string, required []evidence) []string {
	missing := []string{}
	for _, want := range required {
		found := false
		for _, line := range logLines {
			if want.matchesLine(line) {
				found = true
				break
			}
		}
		if !found {
			missing = append(missing, want.String())
		}
	}
	return missing
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "verifylog: "+format+"\n", args...)
	os.Exit(1)
}
