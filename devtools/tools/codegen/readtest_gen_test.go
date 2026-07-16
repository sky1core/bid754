package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
)

type intelMixedReadtestDirectRoute struct {
	goFunction         string
	rustFunction       string
	rustOperandParsers []string
	rustArgs           []string
	rustResultParser   string
	rustComparator     string
}

// intelMixedFMAAndSqrtDirectRoutes is the closed Intel BID C mixed-width
// FMA/sqrt surface enabled in the regular readtest gate. Keep the Go-port and
// generated-Rust names explicit: deriving either name from the readtest name
// would let the two generators make the same routing mistake and still pass.
var intelMixedFMAAndSqrtDirectRoutes = map[string]intelMixedReadtestDirectRoute{
	"bid64ddq_fma":  {goFunction: "Bid64ddqFma", rustFunction: "bid64ddq_fma", rustOperandParsers: []string{"parse_bid64", "parse_bid64", "parse_bid128"}, rustArgs: []string{"a0", "a1", "a2", "rm"}, rustResultParser: "parse_bid64_expected", rustComparator: "compare_u64"},
	"bid64dqd_fma":  {goFunction: "Bid64dqdFma", rustFunction: "bid64dqd_fma", rustOperandParsers: []string{"parse_bid64", "parse_bid128", "parse_bid64"}, rustArgs: []string{"a0", "a1", "a2", "rm"}, rustResultParser: "parse_bid64_expected", rustComparator: "compare_u64"},
	"bid64dqq_fma":  {goFunction: "Bid64dqqFma", rustFunction: "bid64dqq_fma", rustOperandParsers: []string{"parse_bid64", "parse_bid128", "parse_bid128"}, rustArgs: []string{"a0", "a1", "a2", "rm"}, rustResultParser: "parse_bid64_expected", rustComparator: "compare_u64"},
	"bid64qdd_fma":  {goFunction: "Bid64qddFma", rustFunction: "bid64qdd_fma", rustOperandParsers: []string{"parse_bid128", "parse_bid64", "parse_bid64"}, rustArgs: []string{"a0", "a1", "a2", "rm"}, rustResultParser: "parse_bid64_expected", rustComparator: "compare_u64"},
	"bid64qdq_fma":  {goFunction: "Bid64qdqFma", rustFunction: "bid64qdq_fma", rustOperandParsers: []string{"parse_bid128", "parse_bid64", "parse_bid128"}, rustArgs: []string{"a0", "a1", "a2", "rm"}, rustResultParser: "parse_bid64_expected", rustComparator: "compare_u64"},
	"bid64qqd_fma":  {goFunction: "Bid64qqdFma", rustFunction: "bid64qqd_fma", rustOperandParsers: []string{"parse_bid128", "parse_bid128", "parse_bid64"}, rustArgs: []string{"a0", "a1", "a2", "rm"}, rustResultParser: "parse_bid64_expected", rustComparator: "compare_u64"},
	"bid64qqq_fma":  {goFunction: "Bid64qqqFma", rustFunction: "bid64qqq_fma", rustOperandParsers: []string{"parse_bid128", "parse_bid128", "parse_bid128"}, rustArgs: []string{"a0", "a1", "a2", "rm"}, rustResultParser: "parse_bid64_expected", rustComparator: "compare_u64"},
	"bid128ddd_fma": {goFunction: "Bid128dddFma", rustFunction: "bid128ddd_fma", rustOperandParsers: []string{"parse_bid64", "parse_bid64", "parse_bid64"}, rustArgs: []string{"a0", "a1", "a2", "rm"}, rustResultParser: "parse_bid128_expected", rustComparator: "compare_bid128"},
	"bid128ddq_fma": {goFunction: "Bid128ddqFma", rustFunction: "bid128ddq_fma", rustOperandParsers: []string{"parse_bid64", "parse_bid64", "parse_bid128"}, rustArgs: []string{"a0", "a1", "a2", "rm"}, rustResultParser: "parse_bid128_expected", rustComparator: "compare_bid128"},
	"bid128dqd_fma": {goFunction: "Bid128dqdFma", rustFunction: "bid128dqd_fma", rustOperandParsers: []string{"parse_bid64", "parse_bid128", "parse_bid64"}, rustArgs: []string{"a0", "a1", "a2", "rm"}, rustResultParser: "parse_bid128_expected", rustComparator: "compare_bid128"},
	"bid128dqq_fma": {goFunction: "Bid128dqqFma", rustFunction: "bid128dqq_fma", rustOperandParsers: []string{"parse_bid64", "parse_bid128", "parse_bid128"}, rustArgs: []string{"a0", "a1", "a2", "rm"}, rustResultParser: "parse_bid128_expected", rustComparator: "compare_bid128"},
	"bid128qdd_fma": {goFunction: "Bid128qddFma", rustFunction: "bid128qdd_fma", rustOperandParsers: []string{"parse_bid128", "parse_bid64", "parse_bid64"}, rustArgs: []string{"a0", "a1", "a2", "rm"}, rustResultParser: "parse_bid128_expected", rustComparator: "compare_bid128"},
	"bid128qdq_fma": {goFunction: "Bid128qdqFma", rustFunction: "bid128qdq_fma", rustOperandParsers: []string{"parse_bid128", "parse_bid64", "parse_bid128"}, rustArgs: []string{"a0", "a1", "a2", "rm"}, rustResultParser: "parse_bid128_expected", rustComparator: "compare_bid128"},
	"bid128qqd_fma": {goFunction: "Bid128qqdFma", rustFunction: "bid128qqd_fma", rustOperandParsers: []string{"parse_bid128", "parse_bid128", "parse_bid64"}, rustArgs: []string{"a0", "a1", "a2", "rm"}, rustResultParser: "parse_bid128_expected", rustComparator: "compare_bid128"},
	"bid64q_sqrt":   {goFunction: "Bid64qSqrt", rustFunction: "bid64q_sqrt", rustOperandParsers: []string{"parse_bid128"}, rustArgs: []string{"a0", "rm"}, rustResultParser: "parse_bid64_expected", rustComparator: "compare_u64"},
	"bid128d_sqrt":  {goFunction: "Bid128dSqrt", rustFunction: "bid128d_sqrt", rustOperandParsers: []string{"parse_bid64"}, rustArgs: []string{"a0", "rm"}, rustResultParser: "parse_bid128_expected", rustComparator: "compare_bid128"},
}

func TestRustReadtestDispatchInventoryIsManifestBacked(t *testing.T) {
	projectRoot := filepath.Clean(filepath.Join("..", ".."))
	inventoryPath := filepath.Join(projectRoot, "generated", "testspec", "rust_readtest_dispatch_inventory.json")
	data, err := os.ReadFile(inventoryPath)
	if err != nil {
		t.Fatalf("read Rust readtest dispatch inventory: %v", err)
	}
	var inventory RustReadtestDispatchInventory
	if err := json.Unmarshal(data, &inventory); err != nil {
		t.Fatalf("parse Rust readtest dispatch inventory: %v", err)
	}
	if inventory.SkipManifest != filepath.Join("tools", "registry", "rust_readtest_skip_manifest.json") {
		t.Fatalf("skip manifest = %q", inventory.SkipManifest)
	}
	if inventory.Dispatched != 561 || inventory.Skipped != 0 {
		t.Fatalf("Rust readtest dispatch counts = dispatched %d skipped %d, want 561/0", inventory.Dispatched, inventory.Skipped)
	}
	skipped := 0
	for _, row := range inventory.Functions {
		if row.Status == "skipped" {
			skipped++
			if row.Function == "" || row.Compare == "" || row.ReasonCode == "" || row.Reason == "" || row.Classification == "" {
				t.Fatalf("incomplete Rust readtest skip inventory row: %+v", row)
			}
			continue
		}
		switch row.Route {
		case "generic":
			if row.RustFunction == "" {
				t.Fatalf("generic Rust readtest dispatch row lacks rust_function: %+v", row)
			}
		case "custom":
			if row.RustFunction != "" {
				t.Fatalf("custom Rust readtest dispatch row carries rust_function: %+v", row)
			}
		default:
			t.Fatalf("Rust readtest dispatch row has unknown route: %+v", row)
		}
	}
	if skipped != inventory.Skipped {
		t.Fatalf("counted skipped rows = %d, inventory skipped = %d", skipped, inventory.Skipped)
	}
	if len(inventory.SuitePasses) != len(readtestSuiteFilters) {
		t.Fatalf("inventory suite pass rows = %d, want %d", len(inventory.SuitePasses), len(readtestSuiteFilters))
	}
	for i, suite := range readtestSuiteFilters {
		row := inventory.SuitePasses[i]
		if row.Suite != suite.Display || row.Filter != suite.Filter {
			t.Fatalf("inventory suite pass row %d = %+v, want suite %q filter %q", i, row, suite.Display, suite.Filter)
		}
		if row.ExpectedPasses <= 0 {
			t.Fatalf("inventory suite pass row %d has non-positive expected passes: %+v", i, row)
		}
	}
}

func TestIntelMixedFMAAndSqrtRustReadtestRoutesStayDirect(t *testing.T) {
	if len(intelMixedFMAAndSqrtDirectRoutes) != 16 {
		t.Fatalf("closed Intel mixed FMA/sqrt direct-route list has %d entries, want 16", len(intelMixedFMAAndSqrtDirectRoutes))
	}
	projectRoot := filepath.Clean(filepath.Join("..", ".."))

	rustData, err := os.ReadFile(filepath.Join(projectRoot, "generated", "testspec", "rust_readtest_dispatch_inventory.json"))
	if err != nil {
		t.Fatalf("read Rust readtest dispatch inventory: %v", err)
	}
	var rustInventory RustReadtestDispatchInventory
	if err := json.Unmarshal(rustData, &rustInventory); err != nil {
		t.Fatalf("parse Rust readtest dispatch inventory: %v", err)
	}

	type goportInventory struct {
		Functions []struct {
			Function   string `json:"function"`
			GoFunction string `json:"go_function"`
			Status     string `json:"status"`
		} `json:"functions"`
	}
	goData, err := os.ReadFile(filepath.Join(projectRoot, "generated", "testspec", "goport_readtest_dispatch_inventory.json"))
	if err != nil {
		t.Fatalf("read Go-port readtest dispatch inventory: %v", err)
	}
	var goInventory goportInventory
	if err := json.Unmarshal(goData, &goInventory); err != nil {
		t.Fatalf("parse Go-port readtest dispatch inventory: %v", err)
	}

	rustSeen := make(map[string]int, len(intelMixedFMAAndSqrtDirectRoutes))
	for _, row := range rustInventory.Functions {
		expected, ok := intelMixedFMAAndSqrtDirectRoutes[row.Function]
		if !ok {
			continue
		}
		rustSeen[row.Function]++
		if row.Status != "dispatched" || row.Route != "generic" || row.RustFunction != expected.rustFunction {
			t.Errorf("Rust readtest route for %q = status %q route %q function %q, want dispatched generic %q", row.Function, row.Status, row.Route, row.RustFunction, expected.rustFunction)
		}
	}

	goSeen := make(map[string]int, len(intelMixedFMAAndSqrtDirectRoutes))
	for _, row := range goInventory.Functions {
		expected, ok := intelMixedFMAAndSqrtDirectRoutes[row.Function]
		if !ok {
			continue
		}
		goSeen[row.Function]++
		if row.Status != "dispatched" || row.GoFunction != expected.goFunction {
			t.Errorf("Go-port readtest route for %q = status %q function %q, want dispatched %q", row.Function, row.Status, row.GoFunction, expected.goFunction)
		}
	}

	for function := range intelMixedFMAAndSqrtDirectRoutes {
		if rustSeen[function] != 1 {
			t.Errorf("Rust readtest inventory rows for %q = %d, want exactly 1", function, rustSeen[function])
		}
		if goSeen[function] != 1 {
			t.Errorf("Go-port readtest inventory rows for %q = %d, want exactly 1", function, goSeen[function])
		}
	}
}

func TestIntelMixedFMAAndSqrtGeneratedRustDispatchArmsStayDirect(t *testing.T) {
	projectRoot := filepath.Clean(filepath.Join("..", ".."))
	runnerPath := filepath.Join(projectRoot, "..", "bid754-rs", "ffi-verify", "tests", "readtest_generated.rs")
	source, err := os.ReadFile(runnerPath)
	if err != nil {
		t.Fatalf("read generated Rust readtest runner: %v", err)
	}
	if problems := validateIntelMixedGeneratedRustDispatch(string(source)); len(problems) != 0 {
		t.Fatalf("generated Rust mixed FMA/sqrt dispatch is not direct:\n  %s", strings.Join(problems, "\n  "))
	}
}

func TestIntelMixedGeneratedRustDispatchRejectsCommentAndStringDecoys(t *testing.T) {
	projectRoot := filepath.Clean(filepath.Join("..", ".."))
	runnerPath := filepath.Join(projectRoot, "..", "bid754-rs", "ffi-verify", "tests", "readtest_generated.rs")
	sourceBytes, err := os.ReadFile(runnerPath)
	if err != nil {
		t.Fatalf("read generated Rust readtest runner: %v", err)
	}
	source := string(sourceBytes)
	direct := "            let (got, flags) = bid64ddq_fma(a0, a1, a2, rm);"
	decoy := strings.Join([]string{
		"            // let (got, flags) = bid64ddq_fma(a0, a1, a2, rm);",
		"            let _direct_call_decoy = \"bid64ddq_fma(a0, a1, a2, rm)\";",
		"            let (got, flags) = mixed_fma_adapter(a0, a1, a2, rm);",
	}, "\n")
	mutated := strings.Replace(source, direct, decoy, 1)
	if mutated == source {
		t.Fatalf("generated Rust mutation fixture did not find %q", direct)
	}
	problems := validateIntelMixedGeneratedRustDispatch(mutated)
	if len(problems) == 0 {
		t.Fatal("comment/string direct-call decoys hid a compositional adapter route")
	}
	joined := strings.Join(problems, "\n")
	if !strings.Contains(joined, "bid64ddq_fma") || !strings.Contains(joined, "mixed_fma_adapter") {
		t.Fatalf("mutation was rejected for the wrong reason:\n%s", joined)
	}
}

func TestIntelMixedGeneratedRustDispatchRejectsBrokenDataflowAndOperandSlots(t *testing.T) {
	projectRoot := filepath.Clean(filepath.Join("..", ".."))
	runnerPath := filepath.Join(projectRoot, "..", "bid754-rs", "ffi-verify", "tests", "readtest_generated.rs")
	sourceBytes, err := os.ReadFile(runnerPath)
	if err != nil {
		t.Fatalf("read generated Rust readtest runner: %v", err)
	}
	source := string(sourceBytes)
	const function = "bid64ddq_fma"
	mutations := []struct {
		name string
		old  string
		new  string
	}{
		{
			name: "expected compared to itself",
			old:  "let result = compare_u64(got, expected, flags, expected_flags, CmpMode::CmpFuzzy, rm, ulp_add);",
			new:  "let result = compare_u64(expected, expected, expected_flags, expected_flags, CmpMode::CmpFuzzy, rm, ulp_add);",
		},
		{
			name: "comparison failure ignored",
			old:  "if !matches!(result, DispatchResult::Pass) { return result; }",
			new:  "",
		},
		{
			name: "operand source slot swapped",
			old:  "let Some(a0) = parse_bid64(parts[2]) else { return DispatchResult::Skip };",
			new:  "let Some(a0) = parse_bid64(parts[3]) else { return DispatchResult::Skip };",
		},
	}
	for _, mutation := range mutations {
		mutation := mutation
		t.Run(mutation.name, func(t *testing.T) {
			mutated := mutateIntelMixedGeneratedRustArm(t, source, function, mutation.old, mutation.new)
			problems := validateIntelMixedGeneratedRustDispatch(mutated)
			if len(problems) == 0 {
				t.Fatalf("generated Rust validator accepted %s mutation", mutation.name)
			}
			if joined := strings.Join(problems, "\n"); !strings.Contains(joined, "exact operand/parser/direct-call/comparator/failure-return contract") {
				t.Fatalf("%s mutation was rejected for the wrong reason:\n%s", mutation.name, joined)
			}
		})
	}
}

func TestIntelMixedGeneratedRustDispatchRejectsDeadMatchDecoy(t *testing.T) {
	projectRoot := filepath.Clean(filepath.Join("..", ".."))
	runnerPath := filepath.Join(projectRoot, "..", "bid754-rs", "ffi-verify", "tests", "readtest_generated.rs")
	sourceBytes, err := os.ReadFile(runnerPath)
	if err != nil {
		t.Fatalf("read generated Rust readtest runner: %v", err)
	}
	source := string(sourceBytes)
	const originalPrefix = "fn dispatch(func_name: &str, parts: &[&str], rm: i64, ulp_add: f64) -> DispatchResult {\n    match func_name {"
	const decoyPrefix = "fn dispatch(func_name: &str, parts: &[&str], rm: i64, ulp_add: f64) -> DispatchResult {\n    if false {\n        match func_name {"
	const originalSuffix = "        _ => DispatchResult::Skip,\n    }\n}\n\nfn supported_readtest_func"
	const decoySuffix = "        _ => DispatchResult::Skip,\n        };\n    }\n    DispatchResult::Pass\n}\n\nfn supported_readtest_func"
	mutated := strings.Replace(source, originalPrefix, decoyPrefix, 1)
	mutated = strings.Replace(mutated, originalSuffix, decoySuffix, 1)
	if mutated == source || !strings.Contains(mutated, decoyPrefix) || !strings.Contains(mutated, decoySuffix) {
		t.Fatal("generated Rust dead-match mutation fixture did not rewrite dispatch")
	}
	problems := validateIntelMixedGeneratedRustDispatch(mutated)
	if joined := strings.Join(problems, "\n"); !strings.Contains(joined, "top-level tail `match func_name`") {
		t.Fatalf("dead match decoy was not rejected at the dispatch boundary:\n%s", joined)
	}
}

func TestIntelMixedGeneratedRustDispatchRejectsShadowingMatchArms(t *testing.T) {
	projectRoot := filepath.Clean(filepath.Join("..", ".."))
	runnerPath := filepath.Join(projectRoot, "..", "bid754-rs", "ffi-verify", "tests", "readtest_generated.rs")
	sourceBytes, err := os.ReadFile(runnerPath)
	if err != nil {
		t.Fatalf("read generated Rust readtest runner: %v", err)
	}
	source := string(sourceBytes)
	const matchStart = "    match func_name {\n"
	mutations := []struct {
		name       string
		shadowArm  string
		wantDetail string
	}{
		{
			name:       "wildcard pass before literals",
			shadowArm:  "        _ => DispatchResult::Pass,\n",
			wantDetail: "fallback token",
		},
		{
			name:       "guarded binding pass before literals",
			shadowArm:  "        name if name == \"bid64ddq_fma\" => DispatchResult::Pass,\n",
			wantDetail: "unsupported top-level match-arm token",
		},
		{
			name:       "escaped literal pass before canonical literal",
			shadowArm:  "        \"\\x62id64ddq_fma\" => { DispatchResult::Pass },\n",
			wantDetail: "escaped string label",
		},
	}
	for _, mutation := range mutations {
		mutation := mutation
		t.Run(mutation.name, func(t *testing.T) {
			mutated := strings.Replace(source, matchStart, matchStart+mutation.shadowArm, 1)
			if mutated == source {
				t.Fatal("generated Rust shadow-arm mutation fixture did not find dispatch match")
			}
			problems := validateIntelMixedGeneratedRustDispatch(mutated)
			if joined := strings.Join(problems, "\n"); !strings.Contains(joined, mutation.wantDetail) {
				t.Fatalf("shadowing arm was not rejected at the closed match boundary:\n%s", joined)
			}
		})
	}
}

func mutateIntelMixedGeneratedRustArm(t *testing.T, source, function, old, replacement string) string {
	t.Helper()
	marker := fmt.Sprintf("        %q => {", function)
	start := strings.Index(source, marker)
	if start < 0 {
		t.Fatalf("generated Rust mutation fixture lacks arm %q", function)
	}
	endRel := strings.Index(source[start:], "\n        },")
	if endRel < 0 {
		t.Fatalf("generated Rust mutation fixture cannot find end of arm %q", function)
	}
	end := start + endRel
	arm := source[start:end]
	if got := strings.Count(arm, old); got != 1 {
		t.Fatalf("generated Rust arm %q mutation source %q count = %d, want 1", function, old, got)
	}
	mutatedArm := strings.Replace(arm, old, replacement, 1)
	return source[:start] + mutatedArm + source[end:]
}

func TestIntelMixedGeneratedRustDispatchRejectsUnexpectedArm(t *testing.T) {
	projectRoot := filepath.Clean(filepath.Join("..", ".."))
	runnerPath := filepath.Join(projectRoot, "..", "bid754-rs", "ffi-verify", "tests", "readtest_generated.rs")
	sourceBytes, err := os.ReadFile(runnerPath)
	if err != nil {
		t.Fatalf("read generated Rust readtest runner: %v", err)
	}
	source := string(sourceBytes)
	marker := "        _ => DispatchResult::Skip,"
	extra := strings.Join([]string{
		"        \"bid64ddd_fma\" => {",
		"            let (got, flags) = bid64ddd_fma(a0, a1, a2, rm);",
		"            DispatchResult::Pass",
		"        },",
		marker,
	}, "\n")
	mutated := strings.Replace(source, marker, extra, 1)
	if mutated == source {
		t.Fatalf("generated Rust mutation fixture did not find %q", marker)
	}
	problems := validateIntelMixedGeneratedRustDispatch(mutated)
	joined := strings.Join(problems, "\n")
	if !strings.Contains(joined, `unexpected mixed FMA/sqrt arm "bid64ddd_fma"`) ||
		!strings.Contains(joined, "arm set has 17 entries, want exactly 16") {
		t.Fatalf("unexpected mixed arm did not break the closed set:\n%s", joined)
	}
}

func TestGeneratedRustReadtestRunnerAndSuitesStayLive(t *testing.T) {
	source := loadGeneratedRustReadtestSource(t)
	if problems := validateGeneratedRustReadtestExecutionStructure(source); len(problems) != 0 {
		t.Fatalf("generated Rust readtest execution structure is not live:\n  %s", strings.Join(problems, "\n  "))
	}
}

func TestGeneratedRustReadtestRunnerRejectsDispatchPassSubstitution(t *testing.T) {
	source := loadGeneratedRustReadtestSource(t)
	mutated := mutateRustModuleFunctionSource(t, source, "run_readtest",
		"dispatch(func_name, &parts, rm, ulp_add)", "DispatchResult::Pass")
	problems := validateGeneratedRustReadtestExecutionStructure(mutated)
	if joined := strings.Join(problems, "\n"); !strings.Contains(joined, "catch_unwind dispatch scrutinee") {
		t.Fatalf("dispatch-to-Pass mutation was not rejected at the live runner call boundary:\n%s", joined)
	}
}

func TestGeneratedRustReadtestRunnerRejectsEarlyReturnBeforeLoop(t *testing.T) {
	source := loadGeneratedRustReadtestSource(t)
	mutated := mutateRustModuleFunctionSource(t, source, "run_readtest", `
    for line in reader.lines() {`, `
    return RunSummary {
        passed: 21722, failed: 0, skipped: 0,
        by_func: BTreeMap::new(),
    };

    for line in reader.lines() {`)
	problems := validateGeneratedRustReadtestExecutionStructure(mutated)
	if joined := strings.Join(problems, "\n"); !strings.Contains(joined, "closed live body") {
		t.Fatalf("pre-loop early-return mutation was not rejected at the closed runner body boundary:\n%s", joined)
	}
}

func TestGeneratedRustReadtestSuitesRejectInactiveTestAttributes(t *testing.T) {
	source := loadGeneratedRustReadtestSource(t)
	const suite = "test_readtest_generated_decimal64"
	for _, mutation := range []struct {
		name        string
		replacement string
	}{
		{name: "test attribute removed", replacement: ""},
		{name: "ignored test", replacement: "#[test]\n#[ignore]"},
		{name: "false cfg test", replacement: "#[cfg(any())]\n#[test]"},
	} {
		mutation := mutation
		t.Run(mutation.name, func(t *testing.T) {
			mutated := mutateRustModuleFunctionSource(t, source, suite, "#[test]", mutation.replacement)
			problems := validateGeneratedRustReadtestExecutionStructure(mutated)
			if joined := strings.Join(problems, "\n"); !strings.Contains(joined, suite+" attributes") {
				t.Fatalf("inactive suite mutation was not rejected at the exact #[test] boundary:\n%s", joined)
			}
		})
	}
}

func TestGeneratedRustReadtestSuitesRejectSummaryShadowing(t *testing.T) {
	source := loadGeneratedRustReadtestSource(t)
	const suite = "test_readtest_generated_decimal64"
	mutated := mutateRustModuleFunctionSource(t, source, suite,
		`let s = run_readtest("bid64");`, `let s = run_readtest("bid64");
    let s = RunSummary {
        passed: 21722, failed: 0, skipped: 0,
        by_func: BTreeMap::new(),
    };`)
	problems := validateGeneratedRustReadtestExecutionStructure(mutated)
	if joined := strings.Join(problems, "\n"); !strings.Contains(joined, suite+" body") {
		t.Fatalf("suite summary-shadow mutation was not rejected at the closed suite body boundary:\n%s", joined)
	}
}

func TestGeneratedRustReadtestRejectsCfgDeadDispatchAlias(t *testing.T) {
	source := loadGeneratedRustReadtestSource(t)
	mutated := mutateRustModuleFunctionSource(t, source, "dispatch", "fn dispatch", "#[cfg(any())]\nfn dispatch")
	mutated = insertAfterRustModuleFunction(t, mutated, "dispatch", `
fn pass_dispatch(_func_name: &str, _parts: &[&str], _rm: i64, _ulp_add: f64) -> DispatchResult {
    DispatchResult::Pass
}
use self::pass_dispatch as dispatch;
`)
	problems := validateGeneratedRustReadtestExecutionStructure(mutated)
	if joined := strings.Join(problems, "\n"); !strings.Contains(joined, "dispatch attributes") {
		t.Fatalf("cfg-dead dispatch plus pass alias was not rejected at the active definition boundary:\n%s", joined)
	}
}

func TestGeneratedRustReadtestRejectsCfgDeadRunnerAlias(t *testing.T) {
	source := loadGeneratedRustReadtestSource(t)
	mutated := mutateRustModuleFunctionSource(t, source, "run_readtest", "fn run_readtest", "#[cfg(any())]\nfn run_readtest")
	mutated = insertAfterRustModuleFunction(t, mutated, "run_readtest", `
fn pass_run_readtest(_filter: &str) -> RunSummary {
    RunSummary { passed: 0, failed: 0, skipped: 0, by_func: BTreeMap::new() }
}
use self::pass_run_readtest as run_readtest;
`)
	problems := validateGeneratedRustReadtestExecutionStructure(mutated)
	if joined := strings.Join(problems, "\n"); !strings.Contains(joined, "run_readtest attributes") {
		t.Fatalf("cfg-dead runner plus pass alias was not rejected at the active definition boundary:\n%s", joined)
	}
}

func TestGeneratedRustReadtestRejectsCrateLevelFalseCfg(t *testing.T) {
	source := loadGeneratedRustReadtestSource(t)
	mutated := "#![cfg(any())]\n" + source
	problems := validateGeneratedRustReadtestExecutionStructure(mutated)
	if joined := strings.Join(problems, "\n"); !strings.Contains(joined, "crate-level inner attribute") {
		t.Fatalf("crate-level false cfg was not rejected:\n%s", joined)
	}
}

func loadGeneratedRustReadtestSource(t *testing.T) string {
	t.Helper()
	projectRoot := filepath.Clean(filepath.Join("..", ".."))
	runnerPath := filepath.Join(projectRoot, "..", "bid754-rs", "ffi-verify", "tests", "readtest_generated.rs")
	source, err := os.ReadFile(runnerPath)
	if err != nil {
		t.Fatalf("read generated Rust readtest runner: %v", err)
	}
	return string(source)
}

var intelMixedGeneratedRustArmName = regexp.MustCompile(`^(bid(64|128)[dq]{3}_fma|bid(64q|128d)_sqrt)$`)

type rustDispatchLexKind uint8

const (
	rustDispatchIdent rustDispatchLexKind = iota
	rustDispatchString
	rustDispatchPunct
)

type rustDispatchToken struct {
	kind    rustDispatchLexKind
	text    string
	pos     int
	escaped bool
}

type rustDispatchArm struct {
	name string
	body []rustDispatchToken
}

type rustDispatchCall struct {
	callee string
	args   [][]rustDispatchToken
	start  int
}

type rustModuleFunction struct {
	name           string
	attributes     [][]rustDispatchToken
	signature      []rustDispatchToken
	body           []rustDispatchToken
	itemStartToken int
	bodyCloseToken int
}

type rustReadtestSuiteExecutionContract struct {
	testName       string
	display        string
	filter         string
	expectedPasses int
}

// These four complete suite contracts duplicate the hand-maintained
// verification anchors on purpose: this validator must not derive its suite
// names, filters, labels, or assertions from the generator it is checking.
var rustReadtestSuiteExecutionContracts = []rustReadtestSuiteExecutionContract{
	{testName: "decimal64", display: "decimal64", filter: "bid64", expectedPasses: 21722},
	{testName: "decimal32", display: "decimal32", filter: "bid32", expectedPasses: 20875},
	{testName: "status_control", display: "status-control", filter: "bid", expectedPasses: 137},
	{testName: "decimal128", display: "decimal128", filter: "bid128", expectedPasses: 43910},
}

func validateGeneratedRustReadtestExecutionStructure(source string) []string {
	tokens, err := lexRustDispatchSource(source)
	if err != nil {
		return []string{err.Error()}
	}
	functions, err := parseRustModuleFunctions(tokens)
	if err != nil {
		return []string{err.Error()}
	}

	var problems []string
	var innerAttributes []string
	for i := 0; i+2 < len(tokens); i++ {
		if tokens[i].text == "#" && tokens[i+1].text == "!" && tokens[i+2].text == "[" && rustTokenIsAtModuleLevel(tokens, i) {
			close, err := matchingRustDispatchToken(tokens, i+2, "[", "]")
			if err != nil {
				problems = append(problems, fmt.Sprintf("generated Rust readtest crate-level inner attribute at byte %d is malformed: %v", tokens[i].pos, err))
				continue
			}
			innerAttributes = append(innerAttributes, rustDispatchTokensText(tokens[i:close+1]))
		}
	}
	if want := []string{"#![allow(dead_code,unused_variables)]"}; !equalRustDispatchStrings(innerAttributes, want) {
		problems = append(problems, fmt.Sprintf("generated Rust readtest crate-level inner attributes = %v, want exact non-activating allow attribute %v", innerAttributes, want))
	}

	dispatch, ok := functions["dispatch"]
	if !ok {
		problems = append(problems, "generated Rust readtest lacks module-level fn dispatch")
	} else {
		problems = append(problems, validateRustModuleFunctionHeader(dispatch,
			"fn dispatch(func_name: &str, parts: &[&str], rm: i64, ulp_add: f64) -> DispatchResult", nil)...)
	}
	if _, err := parseRustDispatchMatchArms(tokens); err != nil {
		problems = append(problems, fmt.Sprintf("generated Rust active dispatch body: %v", err))
	}

	runner, ok := functions["run_readtest"]
	if !ok {
		problems = append(problems, "generated Rust readtest lacks module-level fn run_readtest")
	} else {
		problems = append(problems, validateRustModuleFunctionHeader(runner,
			"fn run_readtest(filter: &str) -> RunSummary", nil)...)
		problems = append(problems, validateRustReadtestRunnerDataflow(runner.body)...)
	}
	problems = append(problems, validateRustReadtestSuiteFunctions(functions)...)
	sort.Strings(problems)
	return problems
}

func validateRustModuleFunctionHeader(function rustModuleFunction, expectedSignature string, expectedAttributes []string) []string {
	var problems []string
	expected, err := lexRustDispatchSource(expectedSignature)
	if err != nil {
		return []string{fmt.Sprintf("%s expected signature cannot be lexed: %v", function.name, err)}
	}
	if !equalRustDispatchTokens(function.signature, expected) {
		problems = append(problems, fmt.Sprintf("%s signature is not exact: %s", function.name,
			firstRustDispatchTokenDivergence(function.signature, expected)))
	}
	actualAttributes := make([]string, len(function.attributes))
	for i, attribute := range function.attributes {
		actualAttributes[i] = rustDispatchTokensText(attribute)
	}
	if !equalRustDispatchStrings(actualAttributes, expectedAttributes) {
		problems = append(problems, fmt.Sprintf("%s attributes = %v, want %v", function.name, actualAttributes, expectedAttributes))
	}
	return problems
}

func validateRustReadtestRunnerDataflow(body []rustDispatchToken) []string {
	var problems []string
	expectedBody := mustLexRustDispatchContract(`
    let path = find_readtest_in();
    let f = File::open(&path).expect("open readtest.in");
    let reader = BufReader::new(f);
    let mut summary = RunSummary {
        passed: 0, failed: 0, skipped: 0,
        by_func: BTreeMap::new(),
    };

    for line in reader.lines() {
        let line = line.expect("read line");
        let line = line.trim();
        if line.is_empty() || line.starts_with("--") { continue; }

        let parts: Vec<&str> = line.split_whitespace().collect();
        if parts.len() < 4 { continue; }

        let func_name = parts[0];
        if !readtest_suite_matches(filter, func_name) { continue; }
        if !supported_readtest_func(func_name) { continue; }
        if line.contains("longintsize=32") { continue; }

        let rm: i64 = parts[1].parse().unwrap_or(0);
        let ulp_add = parts.iter()
            .find_map(|part| part.strip_prefix("ulp=").and_then(|value| value.parse::<f64>().ok()))
            .unwrap_or(0.0);
        let entry = summary.by_func.entry(func_name.to_string()).or_insert((0, 0, 0));

        match panic::catch_unwind(AssertUnwindSafe(|| dispatch(func_name, &parts, rm, ulp_add))) {
            Ok(DispatchResult::Pass) => {
                summary.passed += 1;
                entry.0 += 1;
            }
            Ok(DispatchResult::Fail(msg)) => {
                summary.failed += 1;
                entry.1 += 1;
                if entry.1 <= 3 {
                    eprintln!("FAIL {}: {}", func_name, msg);
                }
            }
            Ok(DispatchResult::Skip) => {
                summary.skipped += 1;
                entry.2 += 1;
                if entry.2 <= 3 {
                    eprintln!("SKIP {}: {}", func_name, line);
                }
            }
            Err(_) => {
                summary.skipped += 1;
                entry.2 += 1;
                if entry.2 <= 3 {
                    eprintln!("PANIC {}: {}", func_name, line);
                }
            }
        }
    }

    summary
`)
	if !equalRustDispatchTokens(body, expectedBody) {
		problems = append(problems, fmt.Sprintf("run_readtest closed live body is not exact: %s",
			firstRustDispatchTokenDivergence(body, expectedBody)))
	}
	summaryInit := mustLexRustDispatchContract(`let mut summary = RunSummary {
        passed: 0, failed: 0, skipped: 0,
        by_func: BTreeMap::new(),
    };`)
	if countRustDirectTokenSequence(body, summaryInit) != 1 {
		problems = append(problems, "run_readtest lacks one direct RunSummary initialization")
	}

	forIndexes := rustDirectTokenIndexes(body, "for")
	if len(forIndexes) != 1 {
		return append(problems, fmt.Sprintf("run_readtest direct for-loop count = %d, want 1", len(forIndexes)))
	}
	forIndex := forIndexes[0]
	forPrefix := mustLexRustDispatchContract("for line in reader.lines()")
	if forIndex+len(forPrefix) >= len(body) || !equalRustDispatchTokens(body[forIndex:forIndex+len(forPrefix)], forPrefix) || body[forIndex+len(forPrefix)].text != "{" {
		return append(problems, "run_readtest direct loop is not exact `for line in reader.lines()`")
	}
	loopOpen := forIndex + len(forPrefix)
	loopClose, err := matchingRustDispatchToken(body, loopOpen, "{", "}")
	if err != nil {
		return append(problems, fmt.Sprintf("run_readtest loop: %v", err))
	}
	if !equalRustDispatchTokenTexts(body[loopClose+1:], []string{"summary"}) {
		problems = append(problems, "run_readtest does not return the mutated summary as its sole tail expression")
	}
	loopBody := body[loopOpen+1 : loopClose]
	entry := mustLexRustDispatchContract("let entry = summary.by_func.entry(func_name.to_string()).or_insert((0, 0, 0));")
	if countRustDirectTokenSequence(loopBody, entry) != 1 {
		problems = append(problems, "run_readtest lacks one direct summary.by_func entry binding")
	}

	matchIndexes := rustDirectTokenIndexes(loopBody, "match")
	if len(matchIndexes) != 1 {
		return append(problems, fmt.Sprintf("run_readtest direct DispatchResult match count = %d, want 1", len(matchIndexes)))
	}
	matchIndex := matchIndexes[0]
	matchOpen := nextRustDirectToken(loopBody, matchIndex+1, "{")
	if matchOpen < 0 {
		return append(problems, "run_readtest direct DispatchResult match has no body")
	}
	expectedScrutinee := mustLexRustDispatchContract("panic::catch_unwind(AssertUnwindSafe(|| dispatch(func_name, &parts, rm, ulp_add)))")
	if !equalRustDispatchTokens(loopBody[matchIndex+1:matchOpen], expectedScrutinee) {
		problems = append(problems, fmt.Sprintf("run_readtest catch_unwind dispatch scrutinee is not exact: %s",
			firstRustDispatchTokenDivergence(loopBody[matchIndex+1:matchOpen], expectedScrutinee)))
	}
	matchClose, err := matchingRustDispatchToken(loopBody, matchOpen, "{", "}")
	if err != nil {
		return append(problems, fmt.Sprintf("run_readtest DispatchResult match: %v", err))
	}
	if matchClose != len(loopBody)-1 {
		problems = append(problems, "run_readtest DispatchResult match is not the live loop-tail expression")
	}
	expectedArms := mustLexRustDispatchContract(`
            Ok(DispatchResult::Pass) => {
                summary.passed += 1;
                entry.0 += 1;
            }
            Ok(DispatchResult::Fail(msg)) => {
                summary.failed += 1;
                entry.1 += 1;
                if entry.1 <= 3 {
                    eprintln!("FAIL {}: {}", func_name, msg);
                }
            }
            Ok(DispatchResult::Skip) => {
                summary.skipped += 1;
                entry.2 += 1;
                if entry.2 <= 3 {
                    eprintln!("SKIP {}: {}", func_name, line);
                }
            }
            Err(_) => {
                summary.skipped += 1;
                entry.2 += 1;
                if entry.2 <= 3 {
                    eprintln!("PANIC {}: {}", func_name, line);
                }
            }
`)
	if !equalRustDispatchTokens(loopBody[matchOpen+1:matchClose], expectedArms) {
		problems = append(problems, fmt.Sprintf("run_readtest DispatchResult arms do not preserve exact summary dataflow: %s",
			firstRustDispatchTokenDivergence(loopBody[matchOpen+1:matchClose], expectedArms)))
	}
	return problems
}

func validateRustReadtestSuiteFunctions(functions map[string]rustModuleFunction) []string {
	expected := make(map[string]rustReadtestSuiteExecutionContract, len(rustReadtestSuiteExecutionContracts))
	for _, suite := range rustReadtestSuiteExecutionContracts {
		expected["test_readtest_generated_"+suite.testName] = suite
	}
	var problems []string
	if len(expected) != 4 || len(rustReadtestSuiteExecutionContracts) != 4 {
		problems = append(problems, fmt.Sprintf("generated Rust readtest independent suite contract has %d unique/%d total entries, want exactly 4",
			len(expected), len(rustReadtestSuiteExecutionContracts)))
	}
	actual := map[string]bool{}
	for name := range functions {
		if !strings.HasPrefix(name, "test_readtest_generated_") {
			continue
		}
		actual[name] = true
		if _, ok := expected[name]; !ok {
			problems = append(problems, fmt.Sprintf("generated Rust readtest has unexpected suite function %q", name))
		}
	}
	for name, suite := range expected {
		function, ok := functions[name]
		if !ok {
			problems = append(problems, fmt.Sprintf("generated Rust readtest lacks module-level suite function %q", name))
			continue
		}
		problems = append(problems, validateRustModuleFunctionHeader(function, "fn "+name+"()", []string{"#[test]"})...)
		expectedBody := mustLexRustDispatchContract(fmt.Sprintf(`
    let s = run_readtest(%q);
    println!("%s: passed={} failed={} skipped={}", s.passed, s.failed, s.skipped);
    for (func, (p, f, sk)) in &s.by_func {
        if *f > 0 || *sk > 0 {
            println!("  STAT {}: passed={} failed={} skipped={}", func, p, f, sk);
        }
    }
    assert_eq!(s.failed, 0, "%s readtest failures");
    assert_eq!(s.skipped, 0, "%s readtest skips");
    assert_eq!(s.passed, %d, "%s readtest passed-case count derived from readtest.in at generation time");
`, suite.filter, suite.display, suite.display, suite.display, suite.expectedPasses, suite.display))
		if !equalRustDispatchTokens(function.body, expectedBody) {
			problems = append(problems, fmt.Sprintf("%s body does not preserve the exact live runner-result assertions: %s",
				name, firstRustDispatchTokenDivergence(function.body, expectedBody)))
		}
	}
	if len(actual) != len(expected) {
		problems = append(problems, fmt.Sprintf("generated Rust readtest suite function set has %d entries, want exactly %d", len(actual), len(expected)))
	}
	return problems
}

func parseRustModuleFunctions(tokens []rustDispatchToken) (map[string]rustModuleFunction, error) {
	functions := map[string]rustModuleFunction{}
	depth := 0
	for i := 0; i < len(tokens); {
		switch tokens[i].text {
		case "{":
			depth++
			i++
			continue
		case "}":
			depth--
			if depth < 0 {
				return nil, fmt.Errorf("generated Rust source has unmatched module brace")
			}
			i++
			continue
		}
		if depth != 0 || tokens[i].text != "fn" {
			i++
			continue
		}
		if i+2 >= len(tokens) || tokens[i+1].kind != rustDispatchIdent || tokens[i+2].text != "(" {
			return nil, fmt.Errorf("generated Rust module fn at byte %d has unsupported signature", tokens[i].pos)
		}
		name := tokens[i+1].text
		paramsClose, err := matchingRustDispatchToken(tokens, i+2, "(", ")")
		if err != nil {
			return nil, fmt.Errorf("parse module fn %s params: %w", name, err)
		}
		bodyOpen := -1
		for j := paramsClose + 1; j < len(tokens); j++ {
			if tokens[j].text == ";" {
				return nil, fmt.Errorf("module fn %s has no body", name)
			}
			if tokens[j].text == "{" {
				bodyOpen = j
				break
			}
		}
		if bodyOpen < 0 {
			return nil, fmt.Errorf("module fn %s has no body opener", name)
		}
		bodyClose, err := matchingRustDispatchToken(tokens, bodyOpen, "{", "}")
		if err != nil {
			return nil, fmt.Errorf("parse module fn %s body: %w", name, err)
		}
		attributes, itemStart, err := rustContiguousOuterAttributes(tokens, i)
		if err != nil {
			return nil, fmt.Errorf("parse module fn %s attributes: %w", name, err)
		}
		if _, duplicate := functions[name]; duplicate {
			return nil, fmt.Errorf("generated Rust source has duplicate module fn %s", name)
		}
		functions[name] = rustModuleFunction{
			name:           name,
			attributes:     attributes,
			signature:      append([]rustDispatchToken(nil), tokens[i:bodyOpen]...),
			body:           append([]rustDispatchToken(nil), tokens[bodyOpen+1:bodyClose]...),
			itemStartToken: itemStart,
			bodyCloseToken: bodyClose,
		}
		i = bodyClose + 1
	}
	if depth != 0 {
		return nil, fmt.Errorf("generated Rust source has unclosed module brace")
	}
	return functions, nil
}

func rustContiguousOuterAttributes(tokens []rustDispatchToken, fnIndex int) ([][]rustDispatchToken, int, error) {
	itemStart := fnIndex
	var reversed [][]rustDispatchToken
	for cursor := fnIndex - 1; cursor >= 0 && tokens[cursor].text == "]"; {
		depth := 0
		open := -1
		for i := cursor; i >= 0; i-- {
			switch tokens[i].text {
			case "]":
				depth++
			case "[":
				depth--
				if depth == 0 {
					open = i
				}
			}
			if open >= 0 {
				break
			}
		}
		if open <= 0 || tokens[open-1].text != "#" {
			return nil, 0, fmt.Errorf("tokens before fn end in a non-attribute bracket")
		}
		reversed = append(reversed, append([]rustDispatchToken(nil), tokens[open-1:cursor+1]...))
		itemStart = open - 1
		cursor = open - 2
	}
	attributes := make([][]rustDispatchToken, len(reversed))
	for i := range reversed {
		attributes[len(reversed)-1-i] = reversed[i]
	}
	return attributes, itemStart, nil
}

func mutateRustModuleFunctionSource(t *testing.T, source, functionName, old, replacement string) string {
	t.Helper()
	tokens, err := lexRustDispatchSource(source)
	if err != nil {
		t.Fatalf("lex Rust mutation source: %v", err)
	}
	functions, err := parseRustModuleFunctions(tokens)
	if err != nil {
		t.Fatalf("parse Rust mutation source: %v", err)
	}
	function, ok := functions[functionName]
	if !ok {
		t.Fatalf("Rust mutation source lacks module fn %s", functionName)
	}
	start := tokens[function.itemStartToken].pos
	end := tokens[function.bodyCloseToken].pos + 1
	item := source[start:end]
	if count := strings.Count(item, old); count != 1 {
		t.Fatalf("Rust module fn %s mutation source %q count = %d, want 1", functionName, old, count)
	}
	return source[:start] + strings.Replace(item, old, replacement, 1) + source[end:]
}

func insertAfterRustModuleFunction(t *testing.T, source, functionName, insertion string) string {
	t.Helper()
	tokens, err := lexRustDispatchSource(source)
	if err != nil {
		t.Fatalf("lex Rust insertion source: %v", err)
	}
	functions, err := parseRustModuleFunctions(tokens)
	if err != nil {
		t.Fatalf("parse Rust insertion source: %v", err)
	}
	function, ok := functions[functionName]
	if !ok {
		t.Fatalf("Rust insertion source lacks module fn %s", functionName)
	}
	end := tokens[function.bodyCloseToken].pos + 1
	return source[:end] + insertion + source[end:]
}

func rustTokenIsAtModuleLevel(tokens []rustDispatchToken, index int) bool {
	depth := 0
	for i := 0; i < index; i++ {
		switch tokens[i].text {
		case "{":
			depth++
		case "}":
			depth--
		}
	}
	return depth == 0
}

func rustDirectTokenIndexes(tokens []rustDispatchToken, text string) []int {
	paren, bracket, brace := 0, 0, 0
	var indexes []int
	for i, token := range tokens {
		if paren == 0 && bracket == 0 && brace == 0 && token.text == text {
			indexes = append(indexes, i)
		}
		switch token.text {
		case "(":
			paren++
		case ")":
			paren--
		case "[":
			bracket++
		case "]":
			bracket--
		case "{":
			brace++
		case "}":
			brace--
		}
	}
	return indexes
}

func nextRustDirectToken(tokens []rustDispatchToken, start int, text string) int {
	paren, bracket, brace := 0, 0, 0
	for i := start; i < len(tokens); i++ {
		if paren == 0 && bracket == 0 && brace == 0 && tokens[i].text == text {
			return i
		}
		switch tokens[i].text {
		case "(":
			paren++
		case ")":
			paren--
		case "[":
			bracket++
		case "]":
			bracket--
		case "{":
			brace++
		case "}":
			brace--
		}
	}
	return -1
}

func countRustDirectTokenSequence(tokens, sequence []rustDispatchToken) int {
	direct := make(map[int]bool)
	for _, index := range rustDirectTokenIndexes(tokens, "let") {
		direct[index] = true
	}
	count := 0
	for index := range direct {
		if index+len(sequence) <= len(tokens) && equalRustDispatchTokens(tokens[index:index+len(sequence)], sequence) {
			count++
		}
	}
	return count
}

func mustLexRustDispatchContract(source string) []rustDispatchToken {
	tokens, err := lexRustDispatchSource(source)
	if err != nil {
		panic(fmt.Sprintf("invalid hard-coded Rust structural contract: %v", err))
	}
	return tokens
}

func rustDispatchTokensText(tokens []rustDispatchToken) string {
	var text strings.Builder
	for _, token := range tokens {
		text.WriteString(token.text)
	}
	return text.String()
}

func equalRustDispatchTokenTexts(tokens []rustDispatchToken, texts []string) bool {
	if len(tokens) != len(texts) {
		return false
	}
	for i := range tokens {
		if tokens[i].text != texts[i] {
			return false
		}
	}
	return true
}

func validateIntelMixedGeneratedRustDispatch(source string) []string {
	tokens, err := lexRustDispatchSource(source)
	if err != nil {
		return []string{err.Error()}
	}
	arms, err := parseRustDispatchMatchArms(tokens)
	if err != nil {
		return []string{err.Error()}
	}

	var problems []string
	actual := make(map[string]bool, len(intelMixedFMAAndSqrtDirectRoutes))
	for name := range arms {
		if !intelMixedGeneratedRustArmName.MatchString(name) {
			continue
		}
		actual[name] = true
		if _, ok := intelMixedFMAAndSqrtDirectRoutes[name]; !ok {
			problems = append(problems, fmt.Sprintf("generated Rust dispatch has unexpected mixed FMA/sqrt arm %q", name))
		}
	}
	for name, route := range intelMixedFMAAndSqrtDirectRoutes {
		arm, ok := arms[name]
		if !ok {
			problems = append(problems, fmt.Sprintf("generated Rust dispatch lacks mixed FMA/sqrt arm %q", name))
			continue
		}
		problems = append(problems, validateIntelMixedGeneratedRustArm(name, route, arm.body)...)
	}
	if len(actual) != len(intelMixedFMAAndSqrtDirectRoutes) {
		problems = append(problems, fmt.Sprintf("generated Rust mixed FMA/sqrt arm set has %d entries, want exactly %d", len(actual), len(intelMixedFMAAndSqrtDirectRoutes)))
	}
	sort.Strings(problems)
	return problems
}

func validateIntelMixedGeneratedRustArm(name string, route intelMixedReadtestDirectRoute, body []rustDispatchToken) []string {
	calls, err := collectRustDispatchCalls(body)
	if err != nil {
		return []string{fmt.Sprintf("%s: %v", name, err)}
	}

	expectedBody, err := expectedIntelMixedGeneratedRustArmBody(route)
	if err != nil {
		return []string{fmt.Sprintf("%s: invalid direct-route contract: %v", name, err)}
	}
	var problems []string
	if !equalRustDispatchTokens(body, expectedBody) {
		problems = append(problems, fmt.Sprintf("%s: generated arm does not match the exact operand/parser/direct-call/comparator/failure-return contract: %s",
			name, firstRustDispatchTokenDivergence(body, expectedBody)))
	}

	allowed := map[string]bool{
		"parts.len":            true,
		"Some":                 true,
		"parse_bid64":          true,
		"parse_bid128":         true,
		route.rustResultParser: true,
		"parse_flags":          true,
		route.rustFunction:     true,
		route.rustComparator:   true,
	}

	directCalls := 0
	boundDirectCalls := 0
	for _, call := range calls {
		if !allowed[call.callee] {
			problems = append(problems, fmt.Sprintf("%s: generated arm contains adapter/composition call %q", name, call.callee))
		}
		if call.callee != route.rustFunction {
			continue
		}
		directCalls++
		if rustDispatchCallBindsGotFlags(body, call.start) {
			boundDirectCalls++
		}
		gotArgs, ok := rustDispatchIdentifierArgs(call.args)
		if !ok || !equalRustDispatchStrings(gotArgs, route.rustArgs) {
			problems = append(problems, fmt.Sprintf("%s: direct call args = %v, want %v", name, gotArgs, route.rustArgs))
		}
	}
	if directCalls != 1 {
		problems = append(problems, fmt.Sprintf("%s: direct call count for %s = %d, want exactly 1", name, route.rustFunction, directCalls))
	}
	if boundDirectCalls != 1 {
		problems = append(problems, fmt.Sprintf("%s: `let (got, flags) = %s(...)` count = %d, want exactly 1", name, route.rustFunction, boundDirectCalls))
	}

	for i, token := range body {
		switch token.text {
		case "+", "-", "*", "/", "%":
			problems = append(problems, fmt.Sprintf("%s: generated arm contains composition operator %q", name, token.text))
		case "as":
			problems = append(problems, fmt.Sprintf("%s: generated arm contains conversion cast `as`", name))
		case "!":
			if i > 0 && i+1 < len(body) && body[i-1].kind == rustDispatchIdent && body[i+1].text == "(" && body[i-1].text != "matches" {
				problems = append(problems, fmt.Sprintf("%s: generated arm contains adapter/composition macro %q", name, body[i-1].text+"!"))
			}
		case "::":
			if i+1 < len(body) && body[i+1].text == "<" {
				problems = append(problems, fmt.Sprintf("%s: generated arm contains an unparsed turbofish call/conversion", name))
			}
		case "(":
			if i > 0 && body[i-1].text == ")" {
				problems = append(problems, fmt.Sprintf("%s: generated arm contains an indirect call", name))
			}
		}
	}
	return problems
}

func expectedIntelMixedGeneratedRustArmBody(route intelMixedReadtestDirectRoute) ([]rustDispatchToken, error) {
	arity := len(route.rustOperandParsers)
	if arity == 0 {
		return nil, fmt.Errorf("no operand parsers")
	}
	if len(route.rustArgs) != arity+1 || route.rustArgs[arity] != "rm" {
		return nil, fmt.Errorf("direct args %v do not end the %d operands with rm", route.rustArgs, arity)
	}
	if route.rustResultParser == "" || route.rustComparator == "" {
		return nil, fmt.Errorf("missing result parser or comparator")
	}

	expectedIndex := arity + 2
	flagsIndex := arity + 3
	var source strings.Builder
	fmt.Fprintf(&source, "if parts.len() <= %s { return DispatchResult::Skip; }\n", strconv.Itoa(flagsIndex))
	for i, parser := range route.rustOperandParsers {
		arg := fmt.Sprintf("a%d", i)
		if route.rustArgs[i] != arg {
			return nil, fmt.Errorf("direct arg %d = %q, want %q", i, route.rustArgs[i], arg)
		}
		fmt.Fprintf(&source, "let Some(%s) = %s(parts[%d]) else { return DispatchResult::Skip };\n", arg, parser, i+2)
	}
	fmt.Fprintf(&source, "let Some(expected) = %s(parts[%d], rm) else { return DispatchResult::Skip };\n", route.rustResultParser, expectedIndex)
	fmt.Fprintf(&source, "let expected_flags = parse_flags(parts[%d]);\n", flagsIndex)
	fmt.Fprintf(&source, "let (got, flags) = %s(%s);\n", route.rustFunction, strings.Join(route.rustArgs, ", "))
	fmt.Fprintf(&source, "let result = %s(got, expected, flags, expected_flags, CmpMode::CmpFuzzy, rm, ulp_add);\n", route.rustComparator)
	source.WriteString("if !matches!(result, DispatchResult::Pass) { return result; }\n")
	source.WriteString("DispatchResult::Pass\n")

	tokens, err := lexRustDispatchSource(source.String())
	if err != nil {
		return nil, fmt.Errorf("lex direct-route contract: %w", err)
	}
	return tokens, nil
}

func equalRustDispatchTokens(left, right []rustDispatchToken) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i].kind != right[i].kind || left[i].text != right[i].text {
			return false
		}
	}
	return true
}

func firstRustDispatchTokenDivergence(got, want []rustDispatchToken) string {
	limit := len(got)
	if len(want) < limit {
		limit = len(want)
	}
	for i := 0; i < limit; i++ {
		if got[i].kind != want[i].kind || got[i].text != want[i].text {
			return fmt.Sprintf("token %d got %q want %q", i, got[i].text, want[i].text)
		}
	}
	return fmt.Sprintf("token count got %d want %d", len(got), len(want))
}

func rustDispatchCallBindsGotFlags(tokens []rustDispatchToken, callStart int) bool {
	if callStart < 7 {
		return false
	}
	want := []string{"let", "(", "got", ",", "flags", ")", "="}
	for i, text := range want {
		if tokens[callStart-7+i].text != text {
			return false
		}
	}
	return true
}

func rustDispatchIdentifierArgs(args [][]rustDispatchToken) ([]string, bool) {
	out := make([]string, len(args))
	for i, arg := range args {
		if len(arg) != 1 || arg[0].kind != rustDispatchIdent {
			return nil, false
		}
		out[i] = arg[0].text
	}
	return out, true
}

func equalRustDispatchStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func parseRustDispatchMatchArms(tokens []rustDispatchToken) (map[string]rustDispatchArm, error) {
	dispatchBodyOpen := -1
	dispatchBodyClose := -1
	for i := 0; i+3 < len(tokens); i++ {
		if tokens[i].text != "fn" || tokens[i+1].text != "dispatch" || tokens[i+2].text != "(" {
			continue
		}
		if dispatchBodyOpen >= 0 {
			return nil, fmt.Errorf("generated Rust source contains multiple dispatch functions")
		}
		paramsClose, err := matchingRustDispatchToken(tokens, i+2, "(", ")")
		if err != nil {
			return nil, fmt.Errorf("parse generated Rust dispatch signature: %w", err)
		}
		bodyOpen := nextRustDispatchToken(tokens, paramsClose+1, "{")
		if bodyOpen < 0 {
			return nil, fmt.Errorf("generated Rust dispatch function has no body")
		}
		bodyClose, err := matchingRustDispatchToken(tokens, bodyOpen, "{", "}")
		if err != nil {
			return nil, fmt.Errorf("parse generated Rust dispatch body: %w", err)
		}
		dispatchBodyOpen = bodyOpen
		dispatchBodyClose = bodyClose
	}
	if dispatchBodyOpen < 0 {
		return nil, fmt.Errorf("generated Rust dispatch function not found")
	}
	if dispatchBodyOpen+3 >= dispatchBodyClose ||
		tokens[dispatchBodyOpen+1].text != "match" ||
		tokens[dispatchBodyOpen+2].text != "func_name" ||
		tokens[dispatchBodyOpen+3].text != "{" {
		return nil, fmt.Errorf("generated Rust dispatch body is not a top-level tail `match func_name`")
	}
	matchOpen := dispatchBodyOpen + 3
	matchClose, err := matchingRustDispatchToken(tokens, matchOpen, "{", "}")
	if err != nil {
		return nil, fmt.Errorf("parse generated Rust dispatch match: %w", err)
	}
	if matchClose != dispatchBodyClose-1 {
		return nil, fmt.Errorf("generated Rust dispatch `match func_name` is not the function's sole tail expression")
	}

	arms := map[string]rustDispatchArm{}
	fallbackSeen := false
	for i := matchOpen + 1; i < matchClose; {
		if tokens[i].text == "_" {
			want := []string{"_", "=>", "DispatchResult", "::", "Skip", ","}
			if fallbackSeen {
				return nil, fmt.Errorf("generated Rust dispatch repeats the fallback arm")
			}
			if i+len(want) > matchClose {
				return nil, fmt.Errorf("generated Rust dispatch has a truncated fallback arm")
			}
			for j, text := range want {
				if tokens[i+j].text != text {
					return nil, fmt.Errorf("generated Rust dispatch fallback token %d = %q, want %q", j, tokens[i+j].text, text)
				}
			}
			if i+len(want) != matchClose {
				return nil, fmt.Errorf("generated Rust dispatch fallback is not the final match arm")
			}
			fallbackSeen = true
			i += len(want)
			continue
		}
		if fallbackSeen {
			return nil, fmt.Errorf("generated Rust dispatch contains an arm after the fallback")
		}
		if tokens[i].kind != rustDispatchString || i+2 >= matchClose || tokens[i+1].text != "=>" || tokens[i+2].text != "{" {
			return nil, fmt.Errorf("generated Rust dispatch has unsupported top-level match-arm token %q at byte %d", tokens[i].text, tokens[i].pos)
		}
		if tokens[i].escaped {
			return nil, fmt.Errorf("generated Rust dispatch has escaped string label %q at byte %d", tokens[i].text, tokens[i].pos)
		}
		name := tokens[i].text
		bodyOpen := i + 2
		bodyClose, err := matchingRustDispatchToken(tokens, bodyOpen, "{", "}")
		if err != nil {
			return nil, fmt.Errorf("parse generated Rust dispatch arm %q: %w", name, err)
		}
		if _, duplicate := arms[name]; duplicate {
			return nil, fmt.Errorf("generated Rust dispatch duplicates arm %q", name)
		}
		if bodyClose+1 >= matchClose || tokens[bodyClose+1].text != "," {
			return nil, fmt.Errorf("generated Rust dispatch arm %q lacks its trailing comma", name)
		}
		arms[name] = rustDispatchArm{name: name, body: tokens[bodyOpen+1 : bodyClose]}
		i = bodyClose + 2
	}
	if !fallbackSeen {
		return nil, fmt.Errorf("generated Rust dispatch lacks the final `_ => DispatchResult::Skip` fallback")
	}
	return arms, nil
}

func collectRustDispatchCalls(tokens []rustDispatchToken) ([]rustDispatchCall, error) {
	var calls []rustDispatchCall
	for i := 0; i+1 < len(tokens); i++ {
		if tokens[i].kind != rustDispatchIdent || tokens[i+1].text != "(" {
			continue
		}
		if isRustDispatchControlKeyword(tokens[i].text) {
			continue
		}
		close, err := matchingRustDispatchToken(tokens, i+1, "(", ")")
		if err != nil {
			return nil, fmt.Errorf("parse call %q: %w", tokens[i].text, err)
		}
		start := i
		for start >= 2 && (tokens[start-1].text == "." || tokens[start-1].text == "::") && tokens[start-2].kind == rustDispatchIdent {
			start -= 2
		}
		var callee strings.Builder
		for j := start; j <= i; j++ {
			callee.WriteString(tokens[j].text)
		}
		args, err := splitRustDispatchCallArgs(tokens[i+2 : close])
		if err != nil {
			return nil, fmt.Errorf("parse call %q args: %w", callee.String(), err)
		}
		calls = append(calls, rustDispatchCall{callee: callee.String(), args: args, start: start})
	}
	return calls, nil
}

func isRustDispatchControlKeyword(text string) bool {
	switch text {
	case "if", "while", "for", "match", "loop", "return", "let":
		return true
	default:
		return false
	}
}

func splitRustDispatchCallArgs(tokens []rustDispatchToken) ([][]rustDispatchToken, error) {
	if len(tokens) == 0 {
		return nil, nil
	}
	var args [][]rustDispatchToken
	start := 0
	paren, bracket, brace := 0, 0, 0
	for i, token := range tokens {
		switch token.text {
		case "(":
			paren++
		case ")":
			paren--
		case "[":
			bracket++
		case "]":
			bracket--
		case "{":
			brace++
		case "}":
			brace--
		case ",":
			if paren == 0 && bracket == 0 && brace == 0 {
				args = append(args, tokens[start:i])
				start = i + 1
			}
		}
		if paren < 0 || bracket < 0 || brace < 0 {
			return nil, fmt.Errorf("unbalanced nested delimiters")
		}
	}
	if paren != 0 || bracket != 0 || brace != 0 {
		return nil, fmt.Errorf("unbalanced nested delimiters")
	}
	args = append(args, tokens[start:])
	return args, nil
}

func matchingRustDispatchToken(tokens []rustDispatchToken, open int, left, right string) (int, error) {
	if open < 0 || open >= len(tokens) || tokens[open].text != left {
		return 0, fmt.Errorf("expected %q opener", left)
	}
	depth := 0
	for i := open; i < len(tokens); i++ {
		switch tokens[i].text {
		case left:
			depth++
		case right:
			depth--
			if depth == 0 {
				return i, nil
			}
		}
	}
	return 0, fmt.Errorf("unbalanced %s%s delimiters", left, right)
}

func nextRustDispatchToken(tokens []rustDispatchToken, start int, text string) int {
	for i := start; i < len(tokens); i++ {
		if tokens[i].text == text {
			return i
		}
	}
	return -1
}

func lexRustDispatchSource(source string) ([]rustDispatchToken, error) {
	var tokens []rustDispatchToken
	for i := 0; i < len(source); {
		if isRustDispatchSpace(source[i]) {
			i++
			continue
		}
		if strings.HasPrefix(source[i:], "//") {
			if end := strings.IndexByte(source[i+2:], '\n'); end >= 0 {
				i += end + 3
			} else {
				break
			}
			continue
		}
		if strings.HasPrefix(source[i:], "/*") {
			end, err := skipNestedRustDispatchBlockComment(source, i)
			if err != nil {
				return nil, err
			}
			i = end
			continue
		}
		if contentStart, hashes, ok := rustDispatchRawStringStart(source, i); ok {
			contentEnd, end, err := scanRustDispatchRawString(source, contentStart, hashes)
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, rustDispatchToken{kind: rustDispatchString, text: source[contentStart:contentEnd], pos: i})
			i = end
			continue
		}
		if source[i] == '"' || (source[i] == 'b' && i+1 < len(source) && source[i+1] == '"') {
			quote := i
			if source[i] == 'b' {
				quote++
			}
			text, end, escaped, err := scanRustDispatchQuoted(source, quote, '"')
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, rustDispatchToken{kind: rustDispatchString, text: text, pos: i, escaped: escaped})
			i = end
			continue
		}
		if source[i] == '\'' || (source[i] == 'b' && i+1 < len(source) && source[i+1] == '\'') {
			quote := i
			if source[i] == 'b' {
				quote++
			}
			_, end, _, err := scanRustDispatchQuoted(source, quote, '\'')
			if err == nil {
				tokens = append(tokens, rustDispatchToken{kind: rustDispatchPunct, text: "<char>", pos: i})
				i = end
				continue
			}
		}
		if isRustDispatchIdentStart(source[i]) {
			end := i + 1
			for end < len(source) && isRustDispatchIdentContinue(source[end]) {
				end++
			}
			tokens = append(tokens, rustDispatchToken{kind: rustDispatchIdent, text: source[i:end], pos: i})
			i = end
			continue
		}
		punct := rustDispatchPunctuation(source[i:])
		tokens = append(tokens, rustDispatchToken{kind: rustDispatchPunct, text: punct, pos: i})
		i += len(punct)
	}
	return tokens, nil
}

func skipNestedRustDispatchBlockComment(source string, start int) (int, error) {
	depth := 0
	for i := start; i+1 < len(source); {
		switch source[i : i+2] {
		case "/*":
			depth++
			i += 2
		case "*/":
			depth--
			i += 2
			if depth == 0 {
				return i, nil
			}
		default:
			i++
		}
	}
	return 0, fmt.Errorf("unterminated Rust block comment at byte %d", start)
}

func rustDispatchRawStringStart(source string, start int) (contentStart, hashes int, ok bool) {
	i := start
	if i < len(source) && source[i] == 'b' {
		i++
	}
	if i >= len(source) || source[i] != 'r' {
		return 0, 0, false
	}
	i++
	for i < len(source) && source[i] == '#' {
		hashes++
		i++
	}
	if i >= len(source) || source[i] != '"' {
		return 0, 0, false
	}
	return i + 1, hashes, true
}

func scanRustDispatchRawString(source string, contentStart, hashes int) (contentEnd, end int, err error) {
	terminator := "\"" + strings.Repeat("#", hashes)
	rel := strings.Index(source[contentStart:], terminator)
	if rel < 0 {
		return 0, 0, fmt.Errorf("unterminated Rust raw string at byte %d", contentStart-1)
	}
	contentEnd = contentStart + rel
	return contentEnd, contentEnd + len(terminator), nil
}

func scanRustDispatchQuoted(source string, quote int, delimiter byte) (string, int, bool, error) {
	var text strings.Builder
	escaped := false
	hadEscape := false
	for i := quote + 1; i < len(source); i++ {
		if escaped {
			text.WriteByte(source[i])
			escaped = false
			continue
		}
		if source[i] == '\\' {
			escaped = true
			hadEscape = true
			continue
		}
		if source[i] == delimiter {
			return text.String(), i + 1, hadEscape, nil
		}
		if delimiter == '\'' && source[i] == '\n' {
			return "", 0, false, fmt.Errorf("unterminated Rust char literal at byte %d", quote)
		}
		text.WriteByte(source[i])
	}
	return "", 0, false, fmt.Errorf("unterminated Rust quoted literal at byte %d", quote)
}

func rustDispatchPunctuation(source string) string {
	for _, punct := range []string{"=>", "::", "<=", ">=", "==", "!=", "->", "&&", "||", "..", "+=", "-=", "*=", "/=", "%="} {
		if strings.HasPrefix(source, punct) {
			return punct
		}
	}
	return source[:1]
}

func isRustDispatchSpace(char byte) bool {
	switch char {
	case ' ', '\t', '\n', '\r':
		return true
	default:
		return false
	}
}

func isRustDispatchIdentStart(char byte) bool {
	return char == '_' || char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z'
}

func isRustDispatchIdentContinue(char byte) bool {
	return isRustDispatchIdentStart(char) || char >= '0' && char <= '9'
}

func TestCountExpectedReadtestPassesIncludesIntelMixedWidthPrefixes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "readtest.in")
	input := strings.Join([]string{
		"bid64_add 0 1 1 2 00",
		"bid64dq_add 0 1 1 2 00",
		"bid32_add 0 1 1 2 00",
		"bid_saveFlags 0 00 00",
		"bid128_add 0 1 1 2 00",
		"bid128qd_div 0 1 1 1 00",
	}, "\n") + "\n"
	if err := os.WriteFile(path, []byte(input), 0o600); err != nil {
		t.Fatal(err)
	}
	dispatched := map[string]bool{
		"bid64_add": true, "bid64dq_add": true, "bid32_add": true,
		"bid_saveFlags": true, "bid128_add": true, "bid128qd_div": true,
	}
	got, err := countExpectedReadtestPasses(path, dispatched)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]int{"bid64": 2, "bid32": 1, "bid": 1, "bid128": 2}
	for filter, count := range want {
		if got[filter] != count {
			t.Errorf("suite %q passes = %d, want %d", filter, got[filter], count)
		}
	}
}

func TestRustReadtestParserMirrorsIntelGetop64MissingBracket(t *testing.T) {
	generated := readtestParsers()
	if !strings.Contains(generated, "fn parse_bid64_scanned_bits") ||
		!strings.Contains(generated, "s.strip_prefix('[')?") ||
		!strings.Contains(generated, ".take(16)") {
		t.Fatal("generated Rust Decimal64 parser does not mirror Intel getop64 scanning")
	}
}

func TestRustRoundingParamAcceptsGeneratedIntWidth(t *testing.T) {
	sig, ok := parseRustFuncSigLine("pub fn bid128_add(mut x: BID_UINT128, mut y: BID_UINT128, mut rnd_mode: i64, pfpsf: &mut u32) -> BID_UINT128 {")
	if !ok {
		t.Fatal("failed to parse generated Rust signature")
	}
	if !rustSigHasRounding(sig) {
		t.Fatalf("generated Rust signature should expose rounding parameter: %+v", sig.Params)
	}
}

func TestRustReadtestBid128EqualityUsesNamedLimbs(t *testing.T) {
	generated := readtestCompareFuncs()
	if strings.Contains(generated, "got.w") || strings.Contains(generated, "expected.w") {
		t.Fatal("generated Rust readtest compares BID_UINT128 through removed array field")
	}
	if !strings.Contains(generated, "CmpMode::CmpFuzzy => got == expected") {
		t.Fatal("generated Rust readtest lacks whole-value BID_UINT128 equality")
	}
}
