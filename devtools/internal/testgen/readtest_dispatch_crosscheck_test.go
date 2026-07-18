package testgen

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The {base, *WithFlags, *Raw} variant-selection policy is implemented twice:
// resolveGoportDispatch in this package picks the Go port function for the
// goport readtest gate, and resolveRustFunc in devtools/tools/codegen picks
// the generated Rust function for the Rust readtest gate. Both generators
// record their resolved names in the checked-in dispatch inventories; this test
// compares the two surfaces under the shared NormalizeReadtestFuncName so a
// policy divergence becomes a visible test failure instead of a silent
// asymmetry between the Go and Rust gates.

// rustReadtestVariantDivergenceExceptions lists the only readtest functions
// whose resolved Go and Rust variants may legitimately differ, each with the
// concrete reason. Entries must correspond to actual divergences in the
// checked-in inventories; stale entries fail the test.
var rustReadtestVariantDivergenceExceptions = map[string]string{}

type crosscheckGoportInventory struct {
	Functions []struct {
		Function   string `json:"function"`
		GoFunction string `json:"go_function"`
		Status     string `json:"status"`
	} `json:"functions"`
}

type crosscheckRustInventory struct {
	Functions []struct {
		Function     string `json:"function"`
		Compare      string `json:"compare"`
		Status       string `json:"status"`
		Route        string `json:"route"`
		RustFunction string `json:"rust_function"`
	} `json:"functions"`
}

func TestRustAndGoportReadtestDispatchVariantSelectionAgrees(t *testing.T) {
	goportRaw, err := os.ReadFile(filepath.Join("..", "..", "generated", "testspec", "goport_readtest_dispatch_inventory.json"))
	if err != nil {
		t.Fatalf("read goport readtest dispatch inventory: %v", err)
	}
	var goportInventory crosscheckGoportInventory
	if err := json.Unmarshal(goportRaw, &goportInventory); err != nil {
		t.Fatalf("unmarshal goport readtest dispatch inventory: %v", err)
	}
	goportFuncs := make(map[string]string, len(goportInventory.Functions))
	for _, row := range goportInventory.Functions {
		if row.Status != "dispatched" {
			continue
		}
		if row.GoFunction == "" {
			t.Errorf("goport inventory row %q is dispatched without go_function", row.Function)
			continue
		}
		goportFuncs[row.Function] = row.GoFunction
	}

	rustRaw, err := os.ReadFile(filepath.Join("..", "..", "generated", "testspec", "rust_readtest_dispatch_inventory.json"))
	if err != nil {
		t.Fatalf("read rust readtest dispatch inventory: %v", err)
	}
	var rustInventory crosscheckRustInventory
	if err := json.Unmarshal(rustRaw, &rustInventory); err != nil {
		t.Fatalf("unmarshal rust readtest dispatch inventory: %v", err)
	}

	usedExceptions := map[string]bool{}
	seen := map[string]bool{}
	genericRows := 0
	for _, row := range rustInventory.Functions {
		if row.Status != "dispatched" || seen[row.Function] {
			continue
		}
		seen[row.Function] = true
		switch row.Route {
		case "custom":
			// Custom adapter emissions do not go through resolveRustFunc, so
			// there is no resolved variant to cross-check.
			continue
		case "generic":
		default:
			t.Errorf("rust inventory row %q has unknown dispatch route %q", row.Function, row.Route)
			continue
		}
		genericRows++
		if row.RustFunction == "" {
			t.Errorf("rust inventory row %q is generic-dispatched without rust_function", row.Function)
			continue
		}
		goFunction, ok := goportFuncs[row.Function]
		if !ok {
			t.Errorf("rust generic-dispatched function %q has no dispatched goport counterpart", row.Function)
			continue
		}
		if NormalizeReadtestFuncName(row.RustFunction) == NormalizeReadtestFuncName(goFunction) {
			continue
		}
		if reason, allowed := rustReadtestVariantDivergenceExceptions[row.Function]; allowed {
			usedExceptions[row.Function] = true
			t.Logf("excepted variant divergence for %q: rust=%q goport=%q (%s)", row.Function, row.RustFunction, goFunction, reason)
			continue
		}
		t.Errorf("readtest variant selection diverged for %q: rust resolved %q, goport resolved %q", row.Function, row.RustFunction, goFunction)
	}
	if genericRows == 0 {
		t.Error("rust readtest dispatch inventory contains no generic-dispatched rows; the cross-check lost its subject")
	}
	for function := range rustReadtestVariantDivergenceExceptions {
		if !usedExceptions[function] {
			t.Errorf("variant divergence exceptions entry %q no longer diverges; remove it", function)
		}
	}
}

func TestGeneratedReadtestCasesCoverEveryRustSelectedSourceRow(t *testing.T) {
	indexPath := filepath.Join("..", "..", "generated", "testspec", "spec_index.json")
	spec, err := LoadGenerated(indexPath)
	if err != nil {
		t.Fatalf("load generated testspec: %v", err)
	}

	rustInventoryPath := filepath.Join("..", "..", "generated", "testspec", "rust_readtest_dispatch_inventory.json")
	rustRaw, err := os.ReadFile(rustInventoryPath)
	if err != nil {
		t.Fatalf("read Rust readtest dispatch inventory: %v", err)
	}
	var rustInventory crosscheckRustInventory
	if err := json.Unmarshal(rustRaw, &rustInventory); err != nil {
		t.Fatalf("unmarshal Rust readtest dispatch inventory: %v", err)
	}
	// The Rust dispatch inventory records one dispatched row per (function,
	// comparator) spec pair: Intel readtest.h holds independent if-blocks, so a
	// function can legitimately be exercised under more than one comparator
	// for the same source row (the fmod CMP_FUZZYSTATUS + CMP_RELATIVEERR
	// duplicate). The generated testspec mirrors that with one suite per
	// (function, comparator), so the coverage key here is
	// (function, line, compare group) and the per-function expected comparator
	// set comes from the same inventory.
	selected := make(map[string]map[string]struct{}, len(rustInventory.Functions))
	for _, row := range rustInventory.Functions {
		if row.Status != "dispatched" {
			continue
		}
		if row.Compare == "" {
			t.Fatalf("Rust readtest dispatch inventory row %q is dispatched without a compare group", row.Function)
		}
		if selected[row.Function] == nil {
			selected[row.Function] = map[string]struct{}{}
		}
		if _, duplicate := selected[row.Function][row.Compare]; duplicate {
			t.Errorf("Rust readtest dispatch inventory duplicates (%s, %s)", row.Function, row.Compare)
		}
		selected[row.Function][row.Compare] = struct{}{}
	}
	if len(selected) == 0 {
		t.Fatal("Rust readtest dispatch inventory contains no selected functions")
	}

	type sourceRow struct {
		function string
		line     int
	}
	const upstreamSource = "third_party/intel_dfp/TESTS/readtest.in"
	generated := make(map[sourceRow]map[string]struct{})
	generatedPairs := 0
	for _, tc := range spec.ReadCases {
		if tc.Source != upstreamSource {
			continue
		}
		key := sourceRow{function: tc.Function, line: tc.Line}
		if generated[key] == nil {
			generated[key] = map[string]struct{}{}
		}
		if _, duplicate := generated[key][tc.CompareGroup]; duplicate {
			t.Errorf("generated readtest duplicates upstream row %s:%d under compare group %s", tc.Function, tc.Line, tc.CompareGroup)
			continue
		}
		if _, ok := selected[tc.Function][tc.CompareGroup]; !ok {
			t.Errorf("generated readtest row %s:%d carries compare group %s that the Rust dispatch inventory does not select", tc.Function, tc.Line, tc.CompareGroup)
		}
		generated[key][tc.CompareGroup] = struct{}{}
		generatedPairs++
	}

	sourcePath := filepath.Join("..", "..", upstreamSource)
	file, err := os.Open(sourcePath)
	if os.IsNotExist(err) {
		t.Skipf("pinned Intel readtest input absent: %s", sourcePath)
	}
	if err != nil {
		t.Fatalf("open pinned Intel readtest input: %v", err)
	}
	defer file.Close()

	var missing []string
	selectedRows := 0
	selectedPairs := 0
	scanner := bufio.NewScanner(file)
	for lineNo := 1; scanner.Scan(); lineNo++ {
		line := scanner.Text()
		if comment := strings.Index(line, "--"); comment >= 0 {
			line = line[:comment]
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		function := fields[0]
		compares, ok := selected[function]
		if !ok {
			continue
		}
		if strings.Contains(strings.ToLower(line), "longintsize=32") {
			continue
		}
		selectedRows++
		selectedPairs += len(compares)
		for compare := range compares {
			if _, ok := generated[sourceRow{function: function, line: lineNo}][compare]; !ok {
				missing = append(missing, fmt.Sprintf("%s:%d (%s): %s", function, lineNo, compare, line))
			}
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan pinned Intel readtest input: %v", err)
	}
	if len(missing) > 0 {
		const reportLimit = 12
		reported := missing
		if len(reported) > reportLimit {
			reported = reported[:reportLimit]
		}
		t.Fatalf("generated testspec (upstream row set %d, %d row-comparator pairs) omitted %d of %d Rust-selected Intel readtest row-comparator pairs (first %d):\n%s",
			len(generated), generatedPairs, len(missing), selectedPairs, len(reported), strings.Join(reported, "\n"))
	}
	if len(generated) != selectedRows {
		t.Fatalf("generated upstream readtest row set = %d, Rust-selected source rows = %d", len(generated), selectedRows)
	}
	if generatedPairs != selectedPairs {
		t.Fatalf("generated upstream readtest row-comparator pairs = %d, Rust-selected source row-comparator pairs = %d", generatedPairs, selectedPairs)
	}
}
