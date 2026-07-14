package testgen

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// TestRustAndGoportDectestDispatchAgrees is the decTest counterpart of
// TestRustAndGoportReadtestDispatchVariantSelectionAgrees
// (readtest_dispatch_crosscheck_test.go): it cross-checks the Rust decTest
// portable leg (dectest_rust_codegen.go, bid754-rs/tests/dectest_generated.rs)
// against the Go decTest portable leg (dectest_goport_codegen.go,
// bid754-go/generated_dectest_goport_*_test.go) from checked-in state, so an
// adapter-semantics divergence between the two legs is caught mechanically
// instead of silently drifting.
//
// Unlike readtest's generated 545-function surface, decTest's oracle-dispatch set is a
// small, fixed 11-operation family with no per-function resolution ambiguity
// to score -- both legs' dispatch is a hand-authored static table/template, not
// a generic signature-matching resolver. The two things that CAN drift between
// two independently hand-authored dispatch tables are (1) a Rust function name
// that does not actually correspond to the same bidgo function the Go leg
// calls for that operation, and (2) the two legs' suite-level executed/skipped
// case accounting silently diverging. This test checks both.
func TestRustAndGoportDectestDispatchAgrees(t *testing.T) {
	t.Run("function_names", testRustDectestDispatchFunctionNamesMatchGoBidgoNames)
	t.Run("suite_coverage", testRustAndGoportDectestSuiteCoverageAgrees)
}

type crosscheckDectestRustInventory struct {
	Functions []struct {
		Operation    string `json:"operation"`
		Width        string `json:"width"`
		GoFunction   string `json:"go_function"`
		RustFunction string `json:"rust_function"`
		Status       string `json:"status"`
	} `json:"functions"`
}

// testRustDectestDispatchFunctionNamesMatchGoBidgoNames verifies, from the
// checked-in devtools/generated/testspec/dectest_rust_dispatch_inventory.json,
// that every dispatch row (a) is actually dispatched (generation itself
// already reports an error on this -- re-checking here checks against a stale
// checked-in inventory from a tree that has since changed), (b) names a real
// bid754-go/internal/bidgo exported function as its Go counterpart, and (c)
// normalizes to the same name as its paired Rust function under the shared
// go2rs camelToSnake convention, tying the Rust dispatch back to the exact
// bidgo function the Go leg exercises for the identical decTest operation.
func testRustDectestDispatchFunctionNamesMatchGoBidgoNames(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "generated", "testspec", "dectest_rust_dispatch_inventory.json"))
	if err != nil {
		t.Fatalf("read dectest rust dispatch inventory: %v", err)
	}
	var inventory crosscheckDectestRustInventory
	if err := json.Unmarshal(raw, &inventory); err != nil {
		t.Fatalf("unmarshal dectest rust dispatch inventory: %v", err)
	}
	if len(inventory.Functions) == 0 {
		t.Fatal("dectest rust dispatch inventory contains no rows; the cross-check lost its subject")
	}

	bidgoFuncs, err := scanBidgoExportedFunctionNames(filepath.Join("..", "..", "..", "bid754-go", "internal", "bidgo"))
	if err != nil {
		t.Fatalf("scan bid754-go/internal/bidgo exported functions: %v", err)
	}

	for _, row := range inventory.Functions {
		row := row
		t.Run(row.Operation+"/"+row.Width, func(t *testing.T) {
			if row.Status != "dispatched" {
				t.Errorf("dectest rust dispatch row %s/%s has status %q, want \"dispatched\"", row.Operation, row.Width, row.Status)
				return
			}
			if !bidgoFuncs[row.GoFunction] {
				t.Errorf("dectest rust dispatch row %s/%s names go_function %q, which is not an exported bid754-go/internal/bidgo function", row.Operation, row.Width, row.GoFunction)
			}
			if NormalizeReadtestFuncName(row.RustFunction) != NormalizeReadtestFuncName(row.GoFunction) {
				t.Errorf("dectest dispatch naming diverged for %s/%s: rust %q, go %q", row.Operation, row.Width, row.RustFunction, row.GoFunction)
			}
		})
	}
}

var bidgoExportedFuncPattern = regexp.MustCompile(`(?m)^func ([A-Z][A-Za-z0-9_]*)\(`)

func scanBidgoExportedFunctionNames(dir string) (map[string]bool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	funcs := map[string]bool{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, err
		}
		for _, match := range bidgoExportedFuncPattern.FindAllStringSubmatch(string(data), -1) {
			funcs[match[1]] = true
		}
	}
	return funcs, nil
}

// testRustAndGoportDectestSuiteCoverageAgrees recounts each fixed-width
// suite's cases/executed totals with countDectestGoportSuiteCoverage (the
// exact function both dectest_goport_codegen.go and dectest_rust_codegen.go
// call to build their respective legs' pinned expected-coverage tables) and
// compares the live recount against the values actually embedded in the
// checked-in Rust runner (bid754-rs/tests/dectest_generated.rs). Because both
// legs are generated from the same function call, this is primarily a
// regression check against a future edit that makes dectest_rust_codegen.go
// source its coverage from something else; make verify-generated separately
// proves the checked-in Rust file byte-matches a fresh regeneration.
func testRustAndGoportDectestSuiteCoverageAgrees(t *testing.T) {
	repoRoot := filepath.Join("..", "..")
	spec, err := LoadGenerated(filepath.Join(repoRoot, "generated", "testspec", "spec_index.json"))
	if err != nil {
		t.Fatalf("load generated spec: %v", err)
	}
	coverage, err := countDectestGoportSuiteCoverage(repoRoot, spec)
	if err != nil {
		t.Fatalf("count goport dectest suite coverage: %v", err)
	}
	if len(coverage) == 0 {
		t.Fatal("countDectestGoportSuiteCoverage returned no suites; the cross-check lost its subject")
	}

	runnerData, err := os.ReadFile(filepath.Join(repoRoot, "..", "bid754-rs", "tests", "dectest_generated.rs"))
	if err != nil {
		t.Fatalf("read bid754-rs/tests/dectest_generated.rs: %v", err)
	}
	embedded := parseRustSuiteCoverageCasesExecuted(string(runnerData))
	if len(embedded) == 0 {
		t.Fatal("found no SuiteCoverage entries in bid754-rs/tests/dectest_generated.rs; the extraction regex lost its subject")
	}

	if len(embedded) != len(coverage) {
		t.Errorf("bid754-rs/tests/dectest_generated.rs embeds %d SuiteCoverage entries, live recount has %d suites", len(embedded), len(coverage))
	}
	for _, item := range coverage {
		got, ok := embedded[item.Name]
		if !ok {
			t.Errorf("suite %q missing from bid754-rs/tests/dectest_generated.rs SuiteCoverage table", item.Name)
			continue
		}
		if got.cases != item.Cases {
			t.Errorf("suite %q: rust embedded cases = %d, live recount (== goport leg's own count) = %d", item.Name, got.cases, item.Cases)
		}
		if got.executed != item.Executed {
			t.Errorf("suite %q: rust embedded executed = %d, live recount (== goport leg's own count) = %d", item.Name, got.executed, item.Executed)
		}
	}
}

type rustSuiteCoverageCasesExecuted struct {
	cases    int
	executed int
}

// rustSuiteCoveragePattern extracts (name, cases, executed) triples from the
// SuiteCoverage { name: "...", cases: N, executed: M, ... } entries
// dectestRustSuiteCoverageLiteral emits. It is intentionally narrow (matching
// only the exact field order/spacing that emitter produces) rather than a
// general Rust struct-literal parser.
var rustSuiteCoveragePattern = regexp.MustCompile(`name:\s*"([^"]+)",\s*cases:\s*(\d+),\s*executed:\s*(\d+),`)

func parseRustSuiteCoverageCasesExecuted(source string) map[string]rustSuiteCoverageCasesExecuted {
	result := map[string]rustSuiteCoverageCasesExecuted{}
	for _, match := range rustSuiteCoveragePattern.FindAllStringSubmatch(source, -1) {
		cases, err := strconv.Atoi(match[2])
		if err != nil {
			continue
		}
		executed, err := strconv.Atoi(match[3])
		if err != nil {
			continue
		}
		result[match[1]] = rustSuiteCoverageCasesExecuted{cases: cases, executed: executed}
	}
	return result
}
