package testgen

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// The Rust public-API parity routing check verifies that the parity runner
// resolves each wrapper's port ADDRESS through apiemit.PortPathFor, the same
// table apiemit uses to write the wrapper call site. That shared routing
// source is a coverage gap: if PortPathFor mapped a symbol to the wrong port
// (e.g. Bid64Div -> bid64_mul), the wrapper generator and the parity runner
// would BOTH call bid64_mul and agree bit-for-bit -- a false pass. The Go
// parity leg does not have this axis coupled: its census bidgo_function
// (Bid64Div, independently verified by publicroute against the Go source) is
// the port name baked into the runner.
//
// This check restores that independence WITHOUT trusting PortPathFor: it
// extracts the actual crate::generated::<module>::<fn> calls emitted into the
// generated wrapper source (bid754-rs/src/generated/api/*.rs) and confirms,
// per emitted census symbol, that the wrapper really calls the port function
// named by the census bidgo_function -- with the census name mapped to a
// Rust function name by an INDEPENDENT camelCase->snake_case transform
// (bidgoFuncToRustFn below), never through PortPathFor. This is the Rust
// counterpart of publicroute's verification<->shim reached-call-set check. If
// PortPathFor is altered, the wrapper's emitted call diverges from the
// census-derived expectation and this test fails.
//
// Because it reads the generated wrapper source (not the parity runner), it
// adds no bytes to bid754-rs/tests/public_parity_generated.rs: the runner is
// unchanged; this routing logic lives in the verification path rather than the
// generated artifact.

// rustRoutingDelegatedShapes classifies the emitted shapes whose wrapper does
// NOT itself call the census bidgo_function's port function, with the reason
// each is exempt from the direct-call check. Every one either delegates to
// another (direct-checked) public method or has no port call at all; the map
// is exhaustive (a new shape that is neither listed here nor direct-checkable
// fails the test), so a future routing-bearing shape cannot silently escape.
var rustRoutingDelegatedShapes = map[string]string{
	"parse":                    "parse() delegates to parse_raw(); the Bid<w>FromString(Raw) routing is checked on the parse_raw shape",
	"parse_fold":               "NewDecimal<w>BIDDirect folds onto the same parse() method (delegates to parse_raw)",
	"parse_with_flags":         "parse_with_flags() delegates to parse_raw(); routing checked on parse_raw",
	"signaling_not_eq_compose": "signaling_ne() delegates to signaling_eq(); routing checked on the signaling_eq_compose shape",
	"radix_const":              "RADIX is a spec-fixed associated const with no port call in the wrapper (the parity runner calls bid<w>_radix only to prove the const equals the port)",
	"copy_fold":                "Copy is the derived Copy trait with no port call in the wrapper (the parity runner calls bid<w>_copy only to prove the derive equals the port)",
}

// rustGeneratedAPIWrapperFileOwners are the generated wrapper source files
// that hold every emitted owner's surface.
var rustGeneratedAPIWrapperFileOwners = map[string]string{
	"decimal64.rs":  "Decimal64",
	"decimal32.rs":  "Decimal32",
	"decimal128.rs": "Decimal128",
	"context.rs":    "Context",
}

var (
	rustGeneratedCallRe = regexp.MustCompile(`crate::generated::\w+::(\w+)`)
	// Matches an inherent `pub fn NAME` / `pub const fn NAME` (the trait-impl
	// `fn from` / `fn fmt` and the non-pub free helpers are handled separately
	// or ignored). NAME is captured for the surface key.
	rustPubFnRe = regexp.MustCompile(`pub (?:const )?fn (\w+)`)
	// Group 1 is the From<T> type parameter, group 2 is the implementing
	// Decimal<w> type (Decimal64 has From<i32>/From<u32>; Decimal128 has all
	// four of From<i32>/From<u32>/From<i64>/From<u64> since its 34 digits
	// represent every int64/uint64 exactly; Decimal32 has none -- BID32's 7
	// significant digits mean it has no *exact* integer widening, so it uses
	// named from_i32/from_u32 methods instead -- see from_i32_mode/
	// from_u32_mode in apiemit -- but the pattern stays width-generic so a
	// future Decimal32 From<T> impl is not silently unchecked).
	rustFromImplRe = regexp.MustCompile(`impl From<(\w+)> for (Decimal32|Decimal64|Decimal128)\s*\{`)
	// Group 1 is the implementing Decimal<w> type.
	rustDisplayImplRe = regexp.MustCompile(`impl fmt::Display for (Decimal32|Decimal64|Decimal128)\s*\{`)
)

// bidgoFuncToRustFn maps a bidgo (Go mechanical port) exported function name to
// the go2rs-convention Rust snake_case function name, re-derived here from the
// census bidgo_function INDEPENDENTLY of apiemit's portPath table: an uppercase
// letter starts a new word (preceded by '_' when it follows a lowercase letter
// or a digit), digits attach to the current run. This reproduces the same
// lowering go2rs applies (Bid64Div->bid64_div, Bid64IsNaN->bid64_is_na_n,
// Bid64ToInt16Rnint->bid64_to_int16_rnint, Bid64FromUint32->bid64_from_uint32)
// so the expectation never passes through the routing source under test. The
// happy-path test cross-checks this transform against the real wrapper calls,
// so a lowering drift surfaces as a routing mismatch rather than passing
// silently.
func bidgoFuncToRustFn(name string) string {
	var b strings.Builder
	var prevLower, prevDigit bool
	for _, r := range name {
		switch {
		case r >= 'A' && r <= 'Z':
			if b.Len() > 0 && (prevLower || prevDigit) {
				b.WriteByte('_')
			}
			b.WriteRune(r - 'A' + 'a')
			prevLower, prevDigit = false, false
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
			prevLower, prevDigit = true, false
		case r >= '0' && r <= '9':
			b.WriteRune(r)
			prevLower, prevDigit = false, true
		default:
			prevLower, prevDigit = false, false
		}
	}
	return b.String()
}

// extractBalancedBlock returns src[openIdx:] up to and including the '}' that
// balances the '{' at openIdx (which must be a '{'). The generated wrapper
// bodies contain no string/char literals or comment braces in the direct-
// checked method bodies (they are pure port calls, type conversions, and match
// arms), so a raw brace counter is sufficient; a violation of that assumption
// would surface as a parse/routing failure in the happy-path test rather than
// a silent miss.
func extractBalancedBlock(src string, openIdx int) string {
	depth := 0
	for i := openIdx; i < len(src); i++ {
		switch src[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return src[openIdx : i+1]
			}
		}
	}
	return src[openIdx:]
}

func extractBodyCalls(body string) map[string]bool {
	set := map[string]bool{}
	for _, m := range rustGeneratedCallRe.FindAllStringSubmatch(body, -1) {
		set[m[1]] = true
	}
	return set
}

// collectRustWrapperPortCalls parses one generated wrapper file and records,
// per surface key ("<owner>.<surface>"), the set of crate::generated::*
// function names its body calls. Inherent `pub fn` methods are keyed by their
// name; the `From<T>` and `Display` trait impls (whose methods are not `pub fn`)
// are keyed to match the census rust_surface spellings ("From<i32>",
// "Display").
func collectRustWrapperPortCalls(owner, src string, out map[string]map[string]bool) {
	for _, loc := range rustPubFnRe.FindAllStringSubmatchIndex(src, -1) {
		name := src[loc[2]:loc[3]]
		rel := strings.IndexByte(src[loc[1]:], '{')
		if rel < 0 {
			continue
		}
		body := extractBalancedBlock(src, loc[1]+rel)
		out[owner+"."+name] = extractBodyCalls(body)
	}
	// From<T>/Display trait impls key themselves by their own captured
	// implementing type (not the owner param), so both regexes are tried on
	// every file unconditionally: context.rs simply never matches either
	// (Context has neither), and this needs no per-owner branch to extend to
	// a new Decimal<w> file.
	for _, loc := range rustFromImplRe.FindAllStringSubmatchIndex(src, -1) {
		t := src[loc[2]:loc[3]]
		implType := src[loc[4]:loc[5]]
		body := extractBalancedBlock(src, loc[1]-1) // rustFromImplRe ends on '{'
		out[implType+".From<"+t+">"] = extractBodyCalls(body)
	}
	if loc := rustDisplayImplRe.FindStringSubmatchIndex(src); loc != nil {
		implType := src[loc[2]:loc[3]]
		body := extractBalancedBlock(src, loc[1]-1)
		out[implType+".Display"] = extractBodyCalls(body)
	}
}

// loadRustWrapperCallMap reads the generated wrapper files and returns the
// per-surface generated-call map, owner-tagged by which file each surface
// came from (rustGeneratedAPIWrapperFileOwners).
func loadRustWrapperCallMap(devRoot string) (map[string]map[string]bool, error) {
	apiDir := filepath.Join(devRoot, "..", "bid754-rs", "src", "generated", "api")
	out := map[string]map[string]bool{}
	for name, owner := range rustGeneratedAPIWrapperFileOwners {
		data, err := os.ReadFile(filepath.Join(apiDir, name))
		if err != nil {
			return nil, fmt.Errorf("read generated wrapper %q: %w", name, err)
		}
		collectRustWrapperPortCalls(owner, string(data), out)
	}
	return out, nil
}

// checkRustParityRouting is the pure verification core (so the negative test can
// feed it a synthetic altered call map). For every emitted census row that is
// not a documented delegated shape, it requires the wrapper's surface to
// actually call the port function the census bidgo_function names (mapped by the
// independent bidgoFuncToRustFn). It returns one problem string per violation;
// an empty slice means the routing matches the census.
func checkRustParityRouting(callMap map[string]map[string]bool, rows []rustParityInventoryRow) []string {
	var problems []string
	checked := 0
	for _, row := range rows {
		if row.Status != "emitted" {
			continue
		}
		if _, delegated := rustRoutingDelegatedShapes[row.Shape]; delegated {
			continue
		}
		key := row.RustOwner + "." + row.RustSurface
		calls, ok := callMap[key]
		if !ok {
			problems = append(problems, fmt.Sprintf(
				"go_symbol %q (owner %s, surface %q, shape %q): no wrapper method/impl found in the generated API source, so its port routing cannot be verified against the census (add the shape to rustRoutingDelegatedShapes with a reason if it legitimately makes no direct port call)",
				row.GoSymbol, row.RustOwner, row.RustSurface, row.Shape))
			continue
		}
		expected := bidgoFuncToRustFn(row.BidgoFunction)
		if !calls[expected] {
			problems = append(problems, fmt.Sprintf(
				"go_symbol %q: census bidgo_function %q maps (independently of PortPathFor) to port fn %q, but the generated wrapper %q calls {%s} -- the wrapper routes to a different port than the census requires (an altered routing source would surface here)",
				row.GoSymbol, row.BidgoFunction, expected, key, strings.Join(sortedStringSet(calls), ", ")))
			continue
		}
		checked++
	}
	if checked == 0 {
		problems = append(problems, "routing check verified zero direct-routed wrappers; the extractor or shape classification is broken")
	}
	return problems
}

func sortedStringSet(set map[string]bool) []string {
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func emittedRustParityRows(inventory rustParityInventory) []rustParityInventoryRow {
	var rows []rustParityInventoryRow
	for _, row := range inventory.Rows {
		if row.Status == "emitted" {
			rows = append(rows, row)
		}
	}
	return rows
}

func TestRustPublicParityRoutingMatchesCensus(t *testing.T) {
	devRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve devtools root: %v", err)
	}
	inventory, err := loadRustAPISurfaceInventory(filepath.Join(devRoot, rustAPISurfaceInventoryPath))
	if err != nil {
		t.Fatalf("load rust API surface inventory: %v", err)
	}
	callMap, err := loadRustWrapperCallMap(devRoot)
	if err != nil {
		t.Fatalf("load rust wrapper call map: %v", err)
	}
	rows := emittedRustParityRows(inventory)
	if len(rows) == 0 {
		t.Fatal("no emitted rows in the rust API surface inventory")
	}

	problems := checkRustParityRouting(callMap, rows)
	if len(problems) > 0 {
		t.Fatalf("rust parity routing does not match the census (%d problem(s)):\n  %s", len(problems), strings.Join(problems, "\n  "))
	}

	// Direct and delegated row counts are anchored outside generation so an
	// over-broad delegated-shape classification cannot shrink this check.
	directRows := 0
	for _, row := range rows {
		if _, delegated := rustRoutingDelegatedShapes[row.Shape]; !delegated {
			directRows++
		}
	}
	anchors := loadVerificationAnchors(t)
	delegatedRows := len(rows) - directRows
	if directRows != anchors.RustPublicAPIDirectRoutingWrappers {
		t.Errorf("direct-routed wrappers = %d, anchor = %d", directRows, anchors.RustPublicAPIDirectRoutingWrappers)
	}
	if delegatedRows != anchors.RustPublicAPIDelegatedRoutingWrappers {
		t.Errorf("delegated wrappers = %d, anchor = %d", delegatedRows, anchors.RustPublicAPIDelegatedRoutingWrappers)
	}
}

func TestRustPublicParityRoutingDetectsPortAlteration(t *testing.T) {
	// Synthetic census row: Decimal64BID.Div must route to bid64_div.
	rows := []rustParityInventoryRow{
		{GoSymbol: "Decimal64BID.Div", Status: "emitted", RustOwner: "Decimal64", RustSurface: "div", Shape: "binary", BidgoFunction: "Bid64Div"},
	}

	// Clean: the wrapper calls the census-named port.
	clean := map[string]map[string]bool{"Decimal64.div": {"bid64_div": true}}
	if p := checkRustParityRouting(clean, rows); len(p) != 0 {
		t.Fatalf("clean routing should pass, got: %v", p)
	}

	// Alteration 1: as if PortPathFor mapped Bid64Div -> bid64_mul, the wrapper
	// now calls bid64_mul. The check must fail (this is the shared-routing-source
	// coverage gap the whole gate exists to close).
	changedCalls := map[string]map[string]bool{"Decimal64.div": {"bid64_mul": true}}
	if p := checkRustParityRouting(changedCalls, rows); len(p) == 0 {
		t.Fatal("a wrapper routed to the wrong port (bid64_mul instead of bid64_div) must be detected")
	}

	// Alteration 2: a symmetric swap (Div->mul AND Mul->div) cannot hide, because
	// the check is per-symbol, not a global set membership.
	swapRows := []rustParityInventoryRow{
		{GoSymbol: "Decimal64BID.Div", Status: "emitted", RustOwner: "Decimal64", RustSurface: "div", Shape: "binary", BidgoFunction: "Bid64Div"},
		{GoSymbol: "Decimal64BID.Mul", Status: "emitted", RustOwner: "Decimal64", RustSurface: "mul", Shape: "binary", BidgoFunction: "Bid64Mul"},
	}
	swapped := map[string]map[string]bool{
		"Decimal64.div": {"bid64_mul": true},
		"Decimal64.mul": {"bid64_div": true},
	}
	if p := checkRustParityRouting(swapped, swapRows); len(p) == 0 {
		t.Fatal("a symmetric Div<->Mul port swap must be detected per-symbol")
	}

	// Alteration 3: the wrapper method is missing entirely.
	if p := checkRustParityRouting(map[string]map[string]bool{}, rows); len(p) == 0 {
		t.Fatal("a missing wrapper method must be detected")
	}
}

// TestRustPublicParityRoutingDetectsAlterationOnRealSource strengthens the
// synthetic negative above with the real extracted call map: it loads the
// actual wrapper calls, confirms they pass, then alters one real direct
// symbol's call set in-place and confirms the check catches it. This proves the
// detection works against the exact data the happy-path test consumes, not just
// hand-built maps.
func TestRustPublicParityRoutingDetectsAlterationOnRealSource(t *testing.T) {
	devRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve devtools root: %v", err)
	}
	inventory, err := loadRustAPISurfaceInventory(filepath.Join(devRoot, rustAPISurfaceInventoryPath))
	if err != nil {
		t.Fatalf("load rust API surface inventory: %v", err)
	}
	callMap, err := loadRustWrapperCallMap(devRoot)
	if err != nil {
		t.Fatalf("load rust wrapper call map: %v", err)
	}
	rows := emittedRustParityRows(inventory)

	if p := checkRustParityRouting(callMap, rows); len(p) != 0 {
		t.Fatalf("real source should pass before alteration, got: %v", p)
	}

	// Find a direct-routed row backed by a real wrapper entry and alter its
	// call set to a plausible-but-wrong port.
	var target string
	for _, row := range rows {
		if _, delegated := rustRoutingDelegatedShapes[row.Shape]; delegated {
			continue
		}
		key := row.RustOwner + "." + row.RustSurface
		if _, ok := callMap[key]; ok {
			target = key
			break
		}
	}
	if target == "" {
		t.Fatal("no direct-routed wrapper found to alter")
	}
	callMap[target] = map[string]bool{"bid64_definitely_not_the_right_port": true}
	if p := checkRustParityRouting(callMap, rows); len(p) == 0 {
		t.Fatalf("altering the real call set of %q must be detected", target)
	}
}

// TestBidgoFuncToRustFnLowering pins the independent camelCase->snake_case
// transform on representative bidgo names (including the NaN and digit-boundary
// edge cases), so a lowering regression is caught directly rather than only
// through a downstream routing mismatch.
func TestBidgoFuncToRustFnLowering(t *testing.T) {
	cases := map[string]string{
		"Bid64Div":                   "bid64_div",
		"Bid64Mul":                   "bid64_mul",
		"Bid64IsNaN":                 "bid64_is_na_n",
		"Bid64ToInt16Rnint":          "bid64_to_int16_rnint",
		"Bid64ToUint32Xceil":         "bid64_to_uint32_xceil",
		"Bid64FromUint32":            "bid64_from_uint32",
		"Bid64QuietGreaterEqual":     "bid64_quiet_greater_equal",
		"Bid64AddWithFlags":          "bid64_add_with_flags",
		"Bid64ToBid128":              "bid64_to_bid128",
		"Bid64RoundIntegralExact":    "bid64_round_integral_exact",
		"Bid64Scalbn":                "bid64_scalbn",
		"Bid64SignalingGreaterEqual": "bid64_signaling_greater_equal",
	}
	for in, want := range cases {
		if got := bidgoFuncToRustFn(in); got != want {
			t.Errorf("bidgoFuncToRustFn(%q) = %q, want %q", in, got, want)
		}
	}
}
