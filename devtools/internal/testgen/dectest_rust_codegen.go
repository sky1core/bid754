package testgen

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/sky1core/bid754/devtools/internal/genmarker"
)

// The Rust decTest portable leg is the Rust counterpart of the Go portable
// mechanical-port decTest leg (dectest_goport_codegen.go): it runs the same
// fixed-width Decimal32/64/128 oracle-dispatch operation set
// (add/subtract/multiply/divide/quantize/compare/comparesig/tosci/toeng/
// tointegral/tointegralx) directly against the go2rs-generated Rust engine
// (bid754-rs/src/generated, reached through generated::prelude) rather than
// the Go mechanical port, and cross-checks it against the same IBM decTest
// expected values. It reuses countDectestGoportSuiteCoverage (this package,
// dectest_goport_codegen.go) so the expected executed/skipped/flag-exempt
// counts are computed by the identical generator-side function the Go leg
// uses -- the two legs' pinned counts are equal by construction, not by
// independent re-derivation.
// Note: repoRoot throughout this package is the devtools/ directory itself
// (see devtools/cmd/testgen/main.go's filepath.Abs(".") with the Makefile's
// `cd devtools && go run ./cmd/testgen`), not the actual repository root, so
// every path below into a sibling published module is prefixed with "../".
const (
	dectestRustRunnerPath            = "../bid754-rs/tests/dectest_generated.rs"
	dectestRustDispatchInventoryPath = "generated/testspec/dectest_rust_dispatch_inventory.json"
	dectestRustSupportTemplatePath   = "internal/testgen/dectest_templates/dectest_goport_dispatch.rs.tmpl"
	rustGeneratedDirRelToRepoRoot    = "../bid754-rs/src/generated"
	rustLibRsRelToRepoRoot           = "../bid754-rs/src/lib.rs"
)

// dectestRustDispatchRow pins one (operation, width) oracle-dispatch mapping as
// a single hand-authored source of truth: the Go bidgo function the portable
// Go-port decTest leg calls for the identical operation
// (bid754-go/types_bidgo_runtime.go / types_bidgo_runtime_compare.go) paired
// with the go2rs-generated Rust function name the Rust decTest leg's static
// dispatch template (dectest_goport_dispatch.rs.tmpl) calls. This table is not
// used to emit the dispatch template itself (that template is a static,
// hand-authored file, exactly like its Go counterpart
// dectest_goport_dispatch.go.tmpl); it exists so
// TestRustDectestGoportFunctionNamesMatchGoBidgoNames can verify, for every
// entry, that (a) the Go name is an actual exported bid754-go/internal/bidgo
// function, (b) the two names normalize-equal under the shared go2rs
// camelToSnake convention, tying the Rust dispatch back to the exact bidgo
// function the Go leg exercises for the same case.
type dectestRustDispatchRow struct {
	Operation string // decTest oracle-op family, or a compare/round-integral sub-op
	Width     string // decimal32 / decimal64 / decimal128
	GoFunc    string // bid754-go/internal/bidgo exported function name
	RustFunc  string // the true go2rs-generated Rust function name (bare, generated::* origin)
}

var dectestRustDispatchTable = []dectestRustDispatchRow{
	// arithmetic
	{"add", "decimal32", "Bid32AddWithFlags", "bid32_add_with_flags"},
	{"add", "decimal64", "Bid64AddWithFlags", "bid64_add_with_flags"},
	{"add", "decimal128", "Bid128Add", "bid128_add"},
	{"subtract", "decimal32", "Bid32SubWithFlags", "bid32_sub_with_flags"},
	{"subtract", "decimal64", "Bid64SubWithFlags", "bid64_sub_with_flags"},
	{"subtract", "decimal128", "Bid128Sub", "bid128_sub"},
	{"multiply", "decimal32", "Bid32MulWithFlags", "bid32_mul_with_flags"},
	{"multiply", "decimal64", "Bid64MulWithFlags", "bid64_mul_with_flags"},
	{"multiply", "decimal128", "Bid128Mul", "bid128_mul"},
	{"divide", "decimal32", "Bid32DivWithFlags", "bid32_div_with_flags"},
	{"divide", "decimal64", "Bid64DivWithFlags", "bid64_div_with_flags"},
	{"divide", "decimal128", "Bid128Div", "bid128_div"},
	{"quantize", "decimal32", "Bid32Quantize", "bid32_quantize"},
	{"quantize", "decimal64", "Bid64Quantize", "bid64_quantize"},
	{"quantize", "decimal128", "Bid128Quantize", "bid128_quantize"},
	// compare predicates (compare/comparesig dispatch)
	{"quiet_less", "decimal32", "Bid32QuietLess", "bid32_quiet_less"},
	{"quiet_less", "decimal64", "Bid64QuietLess", "bid64_quiet_less"},
	{"quiet_less", "decimal128", "Bid128QuietLess", "bid128_quiet_less"},
	{"quiet_greater", "decimal32", "Bid32QuietGreater", "bid32_quiet_greater"},
	{"quiet_greater", "decimal64", "Bid64QuietGreater", "bid64_quiet_greater"},
	{"quiet_greater", "decimal128", "Bid128QuietGreater", "bid128_quiet_greater"},
	{"signaling_less", "decimal32", "Bid32SignalingLess", "bid32_signaling_less"},
	{"signaling_less", "decimal64", "Bid64SignalingLess", "bid64_signaling_less"},
	{"signaling_less", "decimal128", "Bid128SignalingLess", "bid128_signaling_less"},
	{"signaling_greater", "decimal32", "Bid32SignalingGreater", "bid32_signaling_greater"},
	{"signaling_greater", "decimal64", "Bid64SignalingGreater", "bid64_signaling_greater"},
	{"signaling_greater", "decimal128", "Bid128SignalingGreater", "bid128_signaling_greater"},
	// round-integral family (tointegral / tointegralx dispatch)
	{"round_integral_exact", "decimal32", "Bid32RoundIntegralExact", "bid32_round_integral_exact"},
	{"round_integral_exact", "decimal64", "Bid64RoundIntegralExact", "bid64_round_integral_exact"},
	{"round_integral_exact", "decimal128", "Bid128RoundIntegralExact", "bid128_round_integral_exact"},
	{"round_integral_nearest_even", "decimal32", "Bid32RoundIntegralNearestEven", "bid32_round_integral_nearest_even"},
	{"round_integral_nearest_even", "decimal64", "Bid64RoundIntegralNearestEven", "bid64_round_integral_nearest_even"},
	{"round_integral_nearest_even", "decimal128", "Bid128RoundIntegralNearestEven", "bid128_round_integral_nearest_even"},
	{"round_integral_nearest_away", "decimal32", "Bid32RoundIntegralNearestAway", "bid32_round_integral_nearest_away"},
	{"round_integral_nearest_away", "decimal64", "Bid64RoundIntegralNearestAway", "bid64_round_integral_nearest_away"},
	{"round_integral_nearest_away", "decimal128", "Bid128RoundIntegralNearestAway", "bid128_round_integral_nearest_away"},
	{"round_integral_zero", "decimal32", "Bid32RoundIntegralZero", "bid32_round_integral_zero"},
	{"round_integral_zero", "decimal64", "Bid64RoundIntegralZero", "bid64_round_integral_zero"},
	{"round_integral_zero", "decimal128", "Bid128RoundIntegralZero", "bid128_round_integral_zero"},
	{"round_integral_positive", "decimal32", "Bid32RoundIntegralPositive", "bid32_round_integral_positive"},
	{"round_integral_positive", "decimal64", "Bid64RoundIntegralPositive", "bid64_round_integral_positive"},
	{"round_integral_positive", "decimal128", "Bid128RoundIntegralPositive", "bid128_round_integral_positive"},
	{"round_integral_negative", "decimal32", "Bid32RoundIntegralNegative", "bid32_round_integral_negative"},
	{"round_integral_negative", "decimal64", "Bid64RoundIntegralNegative", "bid64_round_integral_negative"},
	{"round_integral_negative", "decimal128", "Bid128RoundIntegralNegative", "bid128_round_integral_negative"},
	// string conversion (tosci / toeng dispatch, and operand parsing shared by every op)
	{"from_string", "decimal32", "Bid32FromStringRaw", "bid32_from_string_raw"},
	{"from_string", "decimal64", "Bid64FromString", "bid64_from_string"},
	{"from_string", "decimal128", "Bid128FromString", "bid128_from_string"},
	{"to_string", "decimal32", "Bid32ToString", "bid32_to_string"},
	{"to_string", "decimal64", "Bid64ToString", "bid64_to_string"},
	{"to_string", "decimal128", "Bid128ToString", "bid128_to_string"},
}

// dectestRustBid64FromStringCrateRootShim documents the one naming exception in
// dectestRustDispatchTable: go2rs resolves the case-insensitive collision
// between bidgo's unexported bid64_from_string and its exported wrapper
// Bid64FromString onto a single Rust identifier, keeping only the
// crate-private (pub(crate)) definition. The Rust dispatch template therefore
// calls the existing hand-written crate-root shim bid754::bid64_from_string_raw
// (bid754-rs/src/lib.rs), a one-line delegation to that exact generated
// function, instead of the bare generated name.
const dectestRustBid64FromStringCrateRootShim = "bid64_from_string_raw"

func WriteDectestRustOutputs(repoRoot string, spec SharedSpec) error {
	files, err := GenerateDectestRustOutputs(repoRoot, spec)
	if err != nil {
		return err
	}
	for path, data := range files {
		fullPath := filepath.Join(repoRoot, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			return fmt.Errorf("mkdir %q: %w", filepath.Dir(fullPath), err)
		}
		if err := os.WriteFile(fullPath, data, 0o644); err != nil {
			return fmt.Errorf("write generated dectest rust output %q: %w", fullPath, err)
		}
	}
	return nil
}

func GenerateDectestRustOutputs(repoRoot string, spec SharedSpec) (map[string][]byte, error) {
	coverage, err := countDectestGoportSuiteCoverage(repoRoot, spec)
	if err != nil {
		return nil, err
	}

	supportData, err := os.ReadFile(filepath.Join(repoRoot, dectestRustSupportTemplatePath))
	if err != nil {
		return nil, fmt.Errorf("read generated dectest rust support template %q: %w", dectestRustSupportTemplatePath, err)
	}
	runnerSource := dectestGeneratedRustSourceFromTemplate(supportData) + dectestRustCasesSource(coverage)

	rustFuncs, err := scanRustGeneratedPublicFunctions(repoRoot)
	if err != nil {
		return nil, err
	}
	libRsFuncs, err := scanRustPublicFunctionsInFile(filepath.Join(repoRoot, rustLibRsRelToRepoRoot))
	if err != nil {
		return nil, err
	}
	if err := requireRustDectestCalleesResolved(runnerSource, rustFuncs, libRsFuncs); err != nil {
		return nil, err
	}

	// The inventory's per-row resolution check answers "does this generated function
	// exist at all" (any visibility), distinct from requireRustDectestCalleesResolved's
	// stricter "is it externally callable from the runner" check above -- the
	// bid64_from_string row's true generated target is pub(crate)-visible, only
	// reachable at runtime through the documented bid64_from_string_raw crate-root
	// shim, so a direct-call inventory would otherwise report it unresolved despite the runner
	// compiling and calling it successfully through that shim.
	anyVisibilityFuncs, err := scanRustGeneratedAnyVisibilityFunctions(repoRoot)
	if err != nil {
		return nil, err
	}
	inventoryData, err := dectestRustDispatchInventoryJSON(anyVisibilityFuncs)
	if err != nil {
		return nil, err
	}

	files := map[string][]byte{
		dectestRustRunnerPath:            []byte(runnerSource),
		dectestRustDispatchInventoryPath: inventoryData,
	}
	formatted, err := formatGeneratedGoOutputs(files) // no-op for non-.go paths; keeps the map allocation pattern uniform
	if err != nil {
		return nil, err
	}
	formatted[dectestRustRunnerPath], err = formatGeneratedRustOutput(formatted[dectestRustRunnerPath])
	if err != nil {
		return nil, err
	}
	return formatted, nil
}

func dectestGeneratedRustSourceFromTemplate(data []byte) string {
	body := strings.TrimLeft(string(data), "\n")
	return genmarker.Line("testgen") + "\n" + body
}

// rustFuncCallPattern extracts every bare bid32_/bid64_/bid128_-prefixed call
// site (`name(`) from the generated Rust runner source. It deliberately does
// not distinguish call sites from other occurrences of the same identifier
// (e.g. inside a comment); requireRustDectestCalleesResolved only needs a
// superset of the real callees to reject unresolved input on a genuinely missing one, and
// a comment mentioning a real generated function name is not a false
// resolution failure.
var rustFuncCallPattern = regexp.MustCompile(`\bbid(?:32|64|128)_[A-Za-z0-9_]+\s*\(`)

// requireRustDectestCalleesResolved is the generation-time strict gate:
// every bid32_/bid64_/bid128_-prefixed identifier the generated runner calls
// must exist as a `pub fn` in the checked-in bid754-rs/src/generated tree (the
// go2rs-generated engine), with the one documented exception
// (dectestRustBid64FromStringCrateRootShim) resolved against the bid754-rs
// crate root instead. An unresolved callee fails generation outright rather
// than emitting a runner that will not compile or, worse, silently resolves
// to an unintended function via a glob import.
func requireRustDectestCalleesResolved(source string, rustFuncs, libRsFuncs map[string]bool) error {
	seen := map[string]bool{}
	var unresolved []string
	for _, match := range rustFuncCallPattern.FindAllString(source, -1) {
		name := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(match), "("))
		if seen[name] {
			continue
		}
		seen[name] = true
		if name == dectestRustBid64FromStringCrateRootShim {
			if !libRsFuncs[name] {
				unresolved = append(unresolved, name+" (expected pub fn in "+rustLibRsRelToRepoRoot+")")
			}
			continue
		}
		if !rustFuncs[name] {
			unresolved = append(unresolved, name+" (expected pub fn in "+rustGeneratedDirRelToRepoRoot+")")
		}
	}
	if len(unresolved) > 0 {
		sort.Strings(unresolved)
		return fmt.Errorf("generated dectest rust runner references unresolved generated function(s): %s", strings.Join(unresolved, ", "))
	}
	if len(seen) == 0 {
		return fmt.Errorf("generated dectest rust runner calls no bid32_/bid64_/bid128_ generated function; the strict scan lost its subject")
	}
	return nil
}

// rustPubFnPattern matches a top-level `pub fn <name>(` declaration. It
// intentionally does not match `pub(crate) fn` (a narrower visibility that
// requireRustDectestCalleesResolved must not treat as externally callable).
var rustPubFnPattern = regexp.MustCompile(`(?m)^pub fn ([A-Za-z0-9_]+)\s*\(`)

// scanRustGeneratedPublicFunctions reads every top-level .rs file directly
// under bid754-rs/src/generated (the go2rs implementation-generation output;
// the api/ subtree is a separate public-API emitter and is
// intentionally excluded -- this leg exercises the generated engine, not the
// public API wrapper layer) and records every `pub fn` name it declares.
func scanRustGeneratedPublicFunctions(repoRoot string) (map[string]bool, error) {
	dir := filepath.Join(repoRoot, rustGeneratedDirRelToRepoRoot)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read %q: %w", dir, err)
	}
	funcs := map[string]bool{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".rs") {
			continue
		}
		found, err := scanRustPublicFunctionsInFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, err
		}
		for name := range found {
			funcs[name] = true
		}
	}
	if len(funcs) == 0 {
		return nil, fmt.Errorf("found no `pub fn` declarations under %q; the generated engine scan lost its subject", dir)
	}
	return funcs, nil
}

func scanRustPublicFunctionsInFile(path string) (map[string]bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %q: %w", path, err)
	}
	funcs := map[string]bool{}
	for _, match := range rustPubFnPattern.FindAllStringSubmatch(string(data), -1) {
		funcs[match[1]] = true
	}
	return funcs, nil
}

// rustAnyVisibilityFnPattern additionally matches `pub(crate) fn <name>(`,
// unlike rustPubFnPattern. It is used only for the dispatch-table existence
// inventory (dectestRustDispatchInventoryJSON), which asks "does the generated engine
// define this function at all", not "can the runner call it directly" --
// requireRustDectestCalleesResolved above is the strict gate for the
// latter, stricter question.
var rustAnyVisibilityFnPattern = regexp.MustCompile(`(?m)^pub(?:\(crate\))? fn ([A-Za-z0-9_]+)\s*\(`)

func scanRustGeneratedAnyVisibilityFunctions(repoRoot string) (map[string]bool, error) {
	dir := filepath.Join(repoRoot, rustGeneratedDirRelToRepoRoot)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read %q: %w", dir, err)
	}
	funcs := map[string]bool{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".rs") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("read %q: %w", entry.Name(), err)
		}
		for _, match := range rustAnyVisibilityFnPattern.FindAllStringSubmatch(string(data), -1) {
			funcs[match[1]] = true
		}
	}
	if len(funcs) == 0 {
		return nil, fmt.Errorf("found no `pub`/`pub(crate)` fn declarations under %q; the generated engine scan lost its subject", dir)
	}
	return funcs, nil
}

// dectestRustDispatchInventoryRow records one dectestRustDispatchTable row's
// generation-time resolution outcome.
type dectestRustDispatchInventoryRow struct {
	Operation    string `json:"operation"`
	Width        string `json:"width"`
	GoFunction   string `json:"go_function"`
	RustFunction string `json:"rust_function"`
	Status       string `json:"status"`
}

type dectestRustDispatchInventory struct {
	Functions []dectestRustDispatchInventoryRow `json:"functions"`
}

// dectestRustDispatchInventoryJSON renders dectestRustDispatchTable's resolution
// outcome as a checked-in JSON artifact:
// devtools/generated/testspec/dectest_rust_dispatch_inventory.json. It is not a
// generation input (nothing reads it back to decide what to generate); it
// exists so TestRustDectestGoportFunctionNamesMatchGoBidgoNames
// (dectest_rust_crosscheck_test.go) can assert, from checked-in state alone,
// that every dispatch row's Rust function genuinely exists in the generated
// engine and its Go bidgo counterpart genuinely exists in bid754-go/internal/bidgo.
func dectestRustDispatchInventoryJSON(anyVisibilityFuncs map[string]bool) ([]byte, error) {
	inventory := dectestRustDispatchInventory{Functions: make([]dectestRustDispatchInventoryRow, 0, len(dectestRustDispatchTable))}
	var unresolved []string
	for _, row := range dectestRustDispatchTable {
		status := "dispatched"
		if !anyVisibilityFuncs[row.RustFunc] {
			status = "unresolved"
			unresolved = append(unresolved, fmt.Sprintf("%s/%s -> %s", row.Operation, row.Width, row.RustFunc))
		}
		inventory.Functions = append(inventory.Functions, dectestRustDispatchInventoryRow{
			Operation:    row.Operation,
			Width:        row.Width,
			GoFunction:   row.GoFunc,
			RustFunction: row.RustFunc,
			Status:       status,
		})
	}
	if len(unresolved) > 0 {
		return nil, fmt.Errorf("dectestRustDispatchTable references generated function(s) absent from bid754-rs/src/generated: %s", strings.Join(unresolved, ", "))
	}
	data, err := json.MarshalIndent(inventory, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal dectest rust dispatch inventory: %w", err)
	}
	return append(data, '\n'), nil
}

func dectestRustCasesSource(coverage []dectestGoportSuiteCoverage) string {
	return strings.NewReplacer(
		"@@RUST_DECTEST_SUITE_COVERAGE@@", dectestRustSuiteCoverageLiteral(coverage),
	).Replace(dectestRustCasesTemplate)
}

func dectestRustSuiteCoverageLiteral(coverage []dectestGoportSuiteCoverage) string {
	var b strings.Builder
	for i, item := range coverage {
		if i > 0 {
			b.WriteByte('\n')
		}
		fmt.Fprintf(&b, "    SuiteCoverage {\n")
		fmt.Fprintf(&b, "        name: %q,\n", item.Name)
		fmt.Fprintf(&b, "        cases: %d,\n", item.Cases)
		fmt.Fprintf(&b, "        executed: %d,\n", item.Executed)
		fmt.Fprintf(&b, "        skip_reasons: &[%s],\n", rustStringInt64PairList(item.SkipReasons))
		fmt.Fprintf(&b, "        flag_exempt: &[%s],\n", rustStringInt64PairList(item.FlagExempt))
		b.WriteString("    },")
	}
	return b.String()
}

// rustStringInt64PairList renders a Go map[string]int as a sorted Rust
// `("key", n), ("key2", n2)` list (for embedding inside a `&[(&str, i64)]`
// slice literal), matching the deterministic key ordering
// stringIntMapLiteral already uses for the Go-target coverage literal so both
// legs' generated source is reproducible.
func rustStringInt64PairList(values map[string]int) string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var b strings.Builder
	for i, key := range keys {
		if i > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "(%q, %d)", key, values[key])
	}
	return b.String()
}

var dectestRustCasesTemplate = `

// ============================================================================
// Pinned expected coverage (Phase 2): the raw decTest case total, the executed
// oracle-op count, and the mechanical skip/flag-exemption buckets, per
// fixed-width suite. executed + sum(skip_reasons) == cases (exempt cases stay
// executed -- only their flag comparison is waived). Computed by the same
// generator-side function that pins the Go goport leg's own
// expectedGoportDectestSuiteCoverage
// (bid754-go/generated_dectest_goport_cases_test.go), so the two pinned
// tables must always agree; see
// devtools/internal/testgen/dectest_rust_crosscheck_test.go.
// ============================================================================

#[derive(Debug, Clone)]
struct SuiteCoverage {
    name: &'static str,
    cases: i64,
    executed: i64,
    skip_reasons: &'static [(&'static str, i64)],
    flag_exempt: &'static [(&'static str, i64)],
}

const EXPECTED_COVERAGE: &[SuiteCoverage] = &[
@@RUST_DECTEST_SUITE_COVERAGE@@
];

fn expected_coverage_for(name: &str) -> &'static SuiteCoverage {
    EXPECTED_COVERAGE
        .iter()
        .find(|c| c.name == name)
        .unwrap_or_else(|| panic!("rust dectest suite {:?} missing expected coverage", name))
}

fn map_sum(entries: &[(&str, i64)]) -> i64 {
    entries.iter().map(|(_, n)| n).sum()
}

/// assert_bucket_counts checks a live-recounted BTreeMap<String,i64> bucket
/// against its pinned (&str, i64) expected slice: same key set, same count per
/// key. Mirrors assertGoportDectestSkipReasons on the Go side.
fn assert_bucket_counts(
    suite: &str,
    label: &str,
    got: &std::collections::BTreeMap<String, i64>,
    want: &[(&str, i64)],
) {
    if got.len() != want.len() {
        panic!(
            "rust dectest suite {:?} {} bucket count = {}, want {} (got {:?}, want {:?})",
            suite,
            label,
            got.len(),
            want.len(),
            got,
            want
        );
    }
    for (key, want_value) in want {
        match got.get(*key) {
            Some(got_value) if got_value == want_value => {}
            Some(got_value) => panic!(
                "rust dectest suite {:?} {} count[{:?}] = {}, want {}",
                suite, label, key, got_value, want_value
            ),
            None => panic!(
                "rust dectest suite {:?} {} missing key {:?} (want {})",
                suite, label, key, want_value
            ),
        }
    }
}

fn record_failure(failed: &mut usize, logged: &mut usize, msg: String) {
    *failed += 1;
    if *logged < 25 {
        *logged += 1;
        eprintln!("{}", msg);
    }
}

#[derive(Debug)]
enum DectestInputPreflight {
    Ready,
    Missing(std::path::PathBuf),
    StatError {
        path: std::path::PathBuf,
        err: std::io::Error,
    },
}

// preflight_required_dectest_inputs is the actual gate used for both the
// checked-in spec index and every downloaded decTest input. The injected probe
// keeps its run/skip/error contract directly testable: only NotFound skips;
// permission and I/O failures remain hard errors.
fn preflight_required_dectest_inputs<F>(
    paths: &[std::path::PathBuf],
    mut probe: F,
) -> DectestInputPreflight
where
    F: FnMut(&std::path::Path) -> std::io::Result<()>,
{
    for path in paths {
        match probe(path) {
            Ok(()) => {}
            Err(err) if err.kind() == std::io::ErrorKind::NotFound => {
                return DectestInputPreflight::Missing(path.clone());
            }
            Err(err) => {
                return DectestInputPreflight::StatError {
                    path: path.clone(),
                    err,
                };
            }
        }
    }
    DectestInputPreflight::Ready
}

#[test]
fn generated_dectest_preflight_skips_only_not_found() {
    let spec = std::path::PathBuf::from("spec_index.json");
    let input = std::path::PathBuf::from("decimal64.decTest");
    let paths = vec![spec.clone(), input.clone()];

    assert!(matches!(
        preflight_required_dectest_inputs(&paths, |_| Ok(())),
        DectestInputPreflight::Ready
    ));

    match preflight_required_dectest_inputs(&paths, |path| {
        if path == input {
            Err(std::io::Error::from(std::io::ErrorKind::NotFound))
        } else {
            Ok(())
        }
    }) {
        DectestInputPreflight::Missing(path) => assert_eq!(path, input),
        other => panic!("NotFound must skip with the missing path, got {:?}", other),
    }

    match preflight_required_dectest_inputs(&paths, |path| {
        if path == spec {
            Err(std::io::Error::from(std::io::ErrorKind::PermissionDenied))
        } else {
            Ok(())
        }
    }) {
        DectestInputPreflight::StatError { path, err } => {
            assert_eq!(path, spec);
            assert_eq!(err.kind(), std::io::ErrorKind::PermissionDenied);
        }
        other => panic!("PermissionDenied must remain a stat error, got {:?}", other),
    }
}

/// generated_dectest_suites_go_port is the Rust counterpart of
/// TestGeneratedDectestSuitesGoPort. It cross-checks the go2rs-generated Rust
/// engine (crate::generated::*) against the IBM decTest expected values for
/// the fixed-width Decimal32/64/128 oracle-dispatch operation set, mirroring
/// bid754-go's Phase 2 comparison strength (value + quantum + BID five-flag
/// parity). It is portable (no cgo, no native prerequisite) and reads only
/// repo-relative pinned inputs (devtools/generated/testspec/spec_index.json,
/// devtools/tests/*.decTest), so it runs under "cargo test" inside the
/// checked-out repo (make test-rust / make test-portable-dectest) and does
/// not ship in the published package (Cargo.toml excludes tests/). It is an
/// independent second-source cross-validation of operations already anchored
/// by the Rust native readtest gate (make test-rust-native); it does not
/// replace it.
///
/// Input-absence policy (SPEC.md / docs/TEST_GENERATION_SPEC.md): the pinned
/// decTest inputs (devtools/tests/*.decTest) are gitignored downloads, absent
/// from a checkout that has not run 'make setup-generation-inputs'. When those
/// inputs -- or the checked-in spec_index.json -- are absent, this portable
/// leg SKIPS with a visible stderr message (the test still reports Ok) rather
/// than failing, matching the SPEC rule that the portable test path consumes
/// checked-in artifacts and does not require the untracked authoritative input
/// tree. Skip is keyed ONLY on genuine file absence: the up-front pinned-input
/// probe classifies each file with fs::metadata and skips only on a NotFound
/// error, failing on any other stat fault rather than absorbing it into a skip.
/// A present-but-malformed input still fails, parse/execution errors and
/// value/quantum/flag divergences are never absorbed into a skip, and whenever
/// the inputs ARE present the leg always executes its full corpus (no silent
/// green). The Go sibling leg TestGeneratedDectestSuitesGoPort applies the same
/// classification (os.IsNotExist skips, any other stat fault Fatalf's), so the
/// two legs are symmetric on absence handling. The executed/
/// skipped/flag-exempt counts are identical between the two legs whenever the
/// inputs are present (the only state in which either leg runs its corpus).
#[test]
fn generated_dectest_suites_go_port() {
    let repo_root = std::path::PathBuf::from(env!("CARGO_MANIFEST_DIR")).join("..");
    let spec_index_path = repo_root.join("devtools/generated/testspec/spec_index.json");
    match preflight_required_dectest_inputs(std::slice::from_ref(&spec_index_path), |path| {
        fs::metadata(path).map(|_| ())
    }) {
        DectestInputPreflight::Ready => {}
        DectestInputPreflight::Missing(path) => {
            eprintln!(
                "skipping generated_dectest_suites_go_port: checked-in spec {} absent -- run 'make setup-generation-inputs' and 'make verify-generated' first",
                path.display()
            );
            return;
        }
        DectestInputPreflight::StatError { path, err } => {
            panic!("stat required decTest input {}: {}", path.display(), err);
        }
    }
    let spec_raw = fs::read_to_string(&spec_index_path)
        .unwrap_or_else(|err| panic!("read {}: {}", spec_index_path.display(), err));
    let spec: SpecIndex = serde_json::from_str(&spec_raw)
        .unwrap_or_else(|err| panic!("parse {}: {}", spec_index_path.display(), err));
    if spec.dectest_suites.is_empty() {
        panic!("expected generated dectest suites in {}", spec_index_path.display());
    }

    // Second absence gate: the pinned decTest input files the suites reference.
    // spec_index.json is checked in, but the devtools/tests/*.decTest files it
    // names are gitignored downloads that may be absent. Probe every
    // fixed-width suite's referenced files up front; if any is absent, skip the
    // whole leg (an unset-up checkout) rather than letting parse_dec_test_file
    // panic on the missing file mid-run. Presence of every file means the leg
    // must execute -- absence is the ONLY skip trigger.
    let mut pinned_input_paths = Vec::new();
    for suite in &spec.dectest_suites {
        if !is_goport_dectest_runner_suite(&suite.test_type) {
            continue;
        }
        for file in &suite.files {
            pinned_input_paths.push(repo_root.join("devtools").join(file));
        }
    }
    match preflight_required_dectest_inputs(&pinned_input_paths, |path| {
        fs::metadata(path).map(|_| ())
    }) {
        DectestInputPreflight::Ready => {}
        DectestInputPreflight::Missing(path) => {
            eprintln!(
                "skipping generated_dectest_suites_go_port: pinned decTest input {} absent -- run 'make setup-generation-inputs' first",
                path.display()
            );
            return;
        }
        DectestInputPreflight::StatError { path, err } => {
            panic!("stat required decTest input {}: {}", path.display(), err);
        }
    }

    let mut total_failed = 0usize;
    let mut suites_checked = 0usize;
    for suite in &spec.dectest_suites {
        if !is_goport_dectest_runner_suite(&suite.test_type) {
            continue;
        }
        suites_checked += 1;
        let expected = expected_coverage_for(&suite.name);

        let mut executed: i64 = 0;
        let mut failed = 0usize;
        let mut logged = 0usize;
        let mut skip_reasons: std::collections::BTreeMap<String, i64> = std::collections::BTreeMap::new();
        let mut flag_exempt: std::collections::BTreeMap<String, i64> = std::collections::BTreeMap::new();

        for file in &suite.files {
            let path = repo_root.join("devtools").join(file);
            let cases = parse_dec_test_file(&path);
            for tc in &cases {
                if let Some(reason) = dectest_goport_skip_reason(&suite.ignored_operations, tc) {
                    *skip_reasons.entry(reason).or_insert(0) += 1;
                    continue;
                }
                executed += 1;
                match run_dectest_goport_case(tc, &suite.test_type) {
                    Err(err) => {
                        record_failure(
                            &mut failed,
                            &mut logged,
                            format!(
                                "rust dectest case {} ({}) execution error: {}",
                                tc.id, suite.test_type, err
                            ),
                        );
                        continue;
                    }
                    Ok((got, got_flags)) => {
                        if !compare_decimal_results(&tc.result, &got) {
                            record_failure(
                                &mut failed,
                                &mut logged,
                                format!(
                                    "rust dectest case {} ({}) value mismatch [rounding {}]: expected {:?}, port produced {:?}",
                                    tc.id, suite.test_type, tc.rounding_mode, tc.result, got
                                ),
                            );
                            continue;
                        }
                        let normalized_op = normalize_dec_test_operation(&tc.operation);
                        if quantum_compared_operation(&normalized_op) && !quantum_equal(&tc.result, &got) {
                            record_failure(
                                &mut failed,
                                &mut logged,
                                format!(
                                    "rust dectest case {} ({}) quantum mismatch [rounding {}]: expected {:?}, port produced {:?} (same value, different cohort member)",
                                    tc.id, suite.test_type, tc.rounding_mode, tc.result, got
                                ),
                            );
                            continue;
                        }
                        let expected_flags_raw = match parse_dec_test_flags(&tc.flags) {
                            Some(v) => v,
                            None => {
                                record_failure(
                                    &mut failed,
                                    &mut logged,
                                    format!(
                                        "rust dectest case {} ({}) carries an unrecognized condition token: {:?}",
                                        tc.id, suite.test_type, tc.flags
                                    ),
                                );
                                continue;
                            }
                        };
                        let expected_flags = expected_flags_raw & FLAG_MASK;
                        if let Some(reason) = dectest_goport_flag_exempt_reason(tc) {
                            *flag_exempt.entry(reason).or_insert(0) += 1;
                            continue;
                        }
                        if (got_flags & FLAG_MASK) != expected_flags {
                            record_failure(
                                &mut failed,
                                &mut logged,
                                format!(
                                    "rust dectest case {} ({}) flag mismatch [rounding {}]: conditions {:?} project to {:#x}, port raised {:#x}",
                                    tc.id,
                                    suite.test_type,
                                    tc.rounding_mode,
                                    tc.flags,
                                    expected_flags,
                                    got_flags & FLAG_MASK
                                ),
                            );
                        }
                    }
                }
            }
        }

        assert_bucket_counts(&suite.name, "skip_reasons", &skip_reasons, expected.skip_reasons);
        assert_bucket_counts(&suite.name, "flag_exempt", &flag_exempt, expected.flag_exempt);
        if executed != expected.executed {
            panic!(
                "rust dectest suite {:?} executed = {}, want {}",
                suite.name, executed, expected.executed
            );
        }
        let skipped_total: i64 = map_sum(expected.skip_reasons);
        if skipped_total + executed != expected.cases {
            panic!(
                "rust dectest suite {:?} accounting broke: executed {} + skipped {} != cases {}",
                suite.name, executed, skipped_total, expected.cases
            );
        }
        println!(
            "{}: executed={} failed={} skipped={} flagExempt={}",
            suite.name,
            executed,
            failed,
            map_sum(expected.skip_reasons),
            map_sum(expected.flag_exempt)
        );
        total_failed += failed;
    }
    if suites_checked == 0 {
        panic!("no fixed-width decTest suites found in spec_index.json; the rust decTest leg lost its subject");
    }
    assert_eq!(
        total_failed, 0,
        "rust dectest value/quantum/flag cross-check: {} case(s) diverged from the IBM expected case",
        total_failed
    );
}
`
