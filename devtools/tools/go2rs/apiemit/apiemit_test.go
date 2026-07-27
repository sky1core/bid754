package apiemit

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// baseInputs builds a minimal consistent census/manifest/signature triple: two
// mapped symbols, one emitted and one deferred.
func baseInputs() (*inventoryFile, *manifestFile, map[string]goSig) {
	inventory := &inventoryFile{
		Total:    2,
		Mapped:   2,
		Excluded: 0,
		Symbols: []inventorySymbol{
			{Symbol: "Decimal64BID.Add", Kind: "method", Status: "mapped", BidgoFunction: "Bid64Add"},
			{Symbol: "Decimal64BID.Sub", Kind: "method", Status: "mapped", BidgoFunction: "Bid64Sub"},
		},
	}
	manifest := &manifestFile{
		Emit: []emitRule{
			{GoSymbol: "Decimal64BID.Add", RustOwner: "Decimal64", RustSurface: "add", Shape: "binary", BidgoFunction: "Bid64Add", Reason: "add wrapper"},
		},
		Deferred: deferredBlock{
			Category: "deferred",
			Symbols:  []string{"Decimal64BID.Sub"},
		},
	}
	sigs := map[string]goSig{
		// Add matches the "binary" shape: value-type method, one value-type
		// parameter, one value-type result.
		"Decimal64BID.Add": {Symbol: "Decimal64BID.Add", Kind: "method", Recv: "Decimal64BID", Name: "Add", Params: []string{"Decimal64BID"}, Results: []string{"Decimal64BID"}},
		// Sub is deferred (never emitted), so its signature is not shape-validated.
		"Decimal64BID.Sub": {Symbol: "Decimal64BID.Sub", Kind: "method", Recv: "Decimal64BID", Name: "Sub"},
	}
	return inventory, manifest, sigs
}

func TestResolveClosureAcceptsExactCover(t *testing.T) {
	inventory, manifest, sigs := baseInputs()
	rows, err := resolveClosure(inventory, manifest, sigs)
	if err != nil {
		t.Fatalf("expected exact cover to resolve, got: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 surface rows, got %d", len(rows))
	}
	byStatus := map[string]int{}
	for _, r := range rows {
		byStatus[r.Status]++
	}
	if byStatus["emitted"] != 1 || byStatus["deferred"] != 1 {
		t.Fatalf("expected 1 emitted + 1 deferred, got %v", byStatus)
	}
}

func TestResolveClosureRejectsUncoveredMappedSymbol(t *testing.T) {
	inventory, manifest, sigs := baseInputs()
	// Drop the deferred entry so Sub is mapped but neither emitted nor deferred.
	manifest.Deferred.Symbols = nil
	_, err := resolveClosure(inventory, manifest, sigs)
	if err == nil || !strings.Contains(err.Error(), "incomplete mapping") {
		t.Fatalf("expected incomplete mapping for an uncovered mapped symbol, got: %v", err)
	}
}

func TestResolveClosureRejectsEmittedAndDeferred(t *testing.T) {
	inventory, manifest, sigs := baseInputs()
	// List the emitted symbol in deferred too.
	manifest.Deferred.Symbols = []string{"Decimal64BID.Sub", "Decimal64BID.Add"}
	_, err := resolveClosure(inventory, manifest, sigs)
	if err == nil || !strings.Contains(err.Error(), "both emitted and deferred") {
		t.Fatalf("expected a both-emitted-and-deferred failure, got: %v", err)
	}
}

func TestResolveClosureRejectsUnknownShape(t *testing.T) {
	inventory, manifest, sigs := baseInputs()
	manifest.Emit[0].Shape = "not_a_shape"
	_, err := resolveClosure(inventory, manifest, sigs)
	if err == nil || !strings.Contains(err.Error(), "unknown shape") {
		t.Fatalf("expected an unknown-shape failure, got: %v", err)
	}
}

func TestResolveClosureRejectsNonMappedEmit(t *testing.T) {
	inventory, manifest, sigs := baseInputs()
	// Emit a symbol that is not a mapped census symbol.
	manifest.Emit[0].GoSymbol = "Decimal64BID.Nonexistent"
	sigs["Decimal64BID.Nonexistent"] = goSig{Symbol: "Decimal64BID.Nonexistent", Kind: "method"}
	_, err := resolveClosure(inventory, manifest, sigs)
	if err == nil || !strings.Contains(err.Error(), "not a mapped census symbol") {
		t.Fatalf("expected a non-mapped emit failure, got: %v", err)
	}
}

func TestResolveClosureRejectsBidgoMismatch(t *testing.T) {
	inventory, manifest, sigs := baseInputs()
	manifest.Emit[0].BidgoFunction = "Bid64Mul"
	_, err := resolveClosure(inventory, manifest, sigs)
	if err == nil || !strings.Contains(err.Error(), "disagrees with census") {
		t.Fatalf("expected a bidgo_function disagreement failure, got: %v", err)
	}
}

func TestResolveClosureRejectsMissingGoSignature(t *testing.T) {
	inventory, manifest, sigs := baseInputs()
	// Remove the Go signature for the emitted symbol.
	delete(sigs, "Decimal64BID.Add")
	_, err := resolveClosure(inventory, manifest, sigs)
	if err == nil || !strings.Contains(err.Error(), "no Go public signature") {
		t.Fatalf("expected a missing-signature failure, got: %v", err)
	}
}

func TestResolveClosureRejectsWrongDeferredCategory(t *testing.T) {
	inventory, manifest, sigs := baseInputs()
	manifest.Deferred.Category = "temporary"
	_, err := resolveClosure(inventory, manifest, sigs)
	if err == nil || !strings.Contains(err.Error(), "deferred.category") {
		t.Fatalf("expected a deferred category failure, got: %v", err)
	}
}

// The next four cases pin the shape<->Go-signature check: an emitted symbol
// whose Go AST signature diverges from the form its shape assumes must fail
// generation, forcing a manifest update when the Go public surface changes
// shape underneath a template.

func TestResolveClosureRejectsShapeSigReceiverMismatch(t *testing.T) {
	inventory, manifest, sigs := baseInputs()
	// The "binary" shape assumes a value-type method; make Add a free function.
	sigs["Decimal64BID.Add"] = goSig{Symbol: "Decimal64BID.Add", Kind: "func", Name: "Add", Params: []string{"Decimal64BID"}, Results: []string{"Decimal64BID"}}
	_, err := resolveClosure(inventory, manifest, sigs)
	if err == nil || !strings.Contains(err.Error(), "expects a value-type method") {
		t.Fatalf("expected a receiver-form failure, got: %v", err)
	}
}

func TestResolveClosureRejectsShapeSigParamArityMismatch(t *testing.T) {
	inventory, manifest, sigs := baseInputs()
	// e.g. Add gaining a rounding parameter: binary expects exactly one param.
	sigs["Decimal64BID.Add"] = goSig{Symbol: "Decimal64BID.Add", Kind: "method", Recv: "Decimal64BID", Name: "Add", Params: []string{"Decimal64BID", "RoundingMode"}, Results: []string{"Decimal64BID"}}
	_, err := resolveClosure(inventory, manifest, sigs)
	if err == nil || !strings.Contains(err.Error(), "expects 1 parameter(s)") {
		t.Fatalf("expected a parameter-arity failure, got: %v", err)
	}
}

func TestResolveClosureRejectsShapeSigResultArityMismatch(t *testing.T) {
	inventory, manifest, sigs := baseInputs()
	// e.g. Add starting to return flags: binary expects exactly one result.
	sigs["Decimal64BID.Add"] = goSig{Symbol: "Decimal64BID.Add", Kind: "method", Recv: "Decimal64BID", Name: "Add", Params: []string{"Decimal64BID"}, Results: []string{"Decimal64BID", "ExceptionFlags"}}
	_, err := resolveClosure(inventory, manifest, sigs)
	if err == nil || !strings.Contains(err.Error(), "expects 1 result(s)") {
		t.Fatalf("expected a result-arity failure, got: %v", err)
	}
}

func TestResolveClosureRejectsShapeSigParamFormMismatch(t *testing.T) {
	inventory, manifest, sigs := baseInputs()
	// Add's operand changing away from a decimal value type.
	sigs["Decimal64BID.Add"] = goSig{Symbol: "Decimal64BID.Add", Kind: "method", Recv: "Decimal64BID", Name: "Add", Params: []string{"int64"}, Results: []string{"Decimal64BID"}}
	_, err := resolveClosure(inventory, manifest, sigs)
	if err == nil || !strings.Contains(err.Error(), "expects parameter 0 to be a decimal value type") {
		t.Fatalf("expected a parameter-form failure, got: %v", err)
	}
}

func TestResolveClosureAcceptsExactMixedWidthShape(t *testing.T) {
	inventory := &inventoryFile{
		Total: 1, Mapped: 1,
		Symbols: []inventorySymbol{{Symbol: "Add64DQBIDWithMode", Kind: "func", Status: "mapped", BidgoFunction: "Bid64dqAdd"}},
	}
	manifest := &manifestFile{
		Emit: []emitRule{{
			GoSymbol: "Add64DQBIDWithMode", RustOwner: "Decimal64", RustSurface: "add_dq_with_mode",
			Shape: "mixed_binary_mode_flags_dq", BidgoFunction: "Bid64dqAdd", Reason: "test mixed wrapper",
		}},
		Deferred: deferredBlock{Category: "deferred"},
	}
	sigs := map[string]goSig{
		"Add64DQBIDWithMode": {
			Symbol: "Add64DQBIDWithMode", Kind: "func", Name: "Add64DQBIDWithMode",
			Params: []string{"Decimal64BID", "Decimal128BID", "RoundingMode"}, Results: []string{"Decimal64BID", "ExceptionFlags"},
		},
	}
	if _, err := resolveClosure(inventory, manifest, sigs); err != nil {
		t.Fatalf("resolve exact mixed shape: %v", err)
	}

	sigs["Add64DQBIDWithMode"] = goSig{
		Symbol: "Add64DQBIDWithMode", Kind: "func", Name: "Add64DQBIDWithMode",
		Params: []string{"Decimal128BID", "Decimal64BID", "RoundingMode"}, Results: []string{"Decimal64BID", "ExceptionFlags"},
	}
	if _, err := resolveClosure(inventory, manifest, sigs); err == nil || !strings.Contains(err.Error(), "requires operands") {
		t.Fatalf("mixed operand-order mismatch error = %v", err)
	}
}

func TestResolveClosureRejectsCommonModeMixedPortMismatch(t *testing.T) {
	for _, wrongPort := range []string{
		"Bid64dqSub", // operation disagrees with Add64DQBIDWithMode
		"Bid64qdAdd", // operand-width order disagrees with DQ
	} {
		t.Run(wrongPort, func(t *testing.T) {
			inventory := &inventoryFile{
				Total: 1, Mapped: 1,
				Symbols: []inventorySymbol{{Symbol: "Add64DQBIDWithMode", Kind: "func", Status: "mapped", BidgoFunction: wrongPort}},
			}
			manifest := &manifestFile{
				Emit: []emitRule{{
					GoSymbol: "Add64DQBIDWithMode", RustOwner: "Decimal64", RustSurface: "add_dq_with_mode",
					Shape: "mixed_binary_mode_flags_dq", BidgoFunction: wrongPort, Reason: "test common-mode mismatch",
				}},
				Deferred: deferredBlock{Category: "deferred"},
			}
			sigs := map[string]goSig{
				"Add64DQBIDWithMode": {
					Symbol: "Add64DQBIDWithMode", Kind: "func", Name: "Add64DQBIDWithMode",
					Params: []string{"Decimal64BID", "Decimal128BID", "RoundingMode"}, Results: []string{"Decimal64BID", "ExceptionFlags"},
				},
			}

			_, err := resolveClosure(inventory, manifest, sigs)
			if err == nil || !strings.Contains(err.Error(), `requires exact port "Bid64dqAdd"`) {
				t.Fatalf("common-mode mixed port mismatch error = %v", err)
			}
		})
	}
}

func TestEmitMixedWidthShapeUsesAssociatedFunctionAndForeignConversions(t *testing.T) {
	var b strings.Builder
	op := mixedDecOp{
		decOp: decOp{method: "add_dq_with_mode", module: "bid128_add", port: "bid64dq_add"},
		left:  decimal64Width,
		right: decimal128Width,
	}
	if err := emitMixedBinaryModeFlagsOps(&b, []mixedDecOp{op}, decimal64Width); err != nil {
		t.Fatal(err)
	}
	got := b.String()
	params, body := rustGeneratedFunction(t, got, "add_dq_with_mode")
	assertRustStringList(t, "add_dq_with_mode params", params, []string{
		"left: Decimal64", "right: Decimal128", "mode: RoundingMode",
	})
	assertRustGeneratedBoundCall(t, body, "(bits, raw)", "crate::generated::bid128_add::bid64dq_add", [][]string{{
		"left.to_bits()",
		"super::types::bid_uint128_from_le_bytes(right.to_le_bytes())",
		"super::types::to_bidgo_rounding(mode)",
	}})
}

func TestResolveClosureAcceptsExactMixedFMAAndSqrtShapes(t *testing.T) {
	tests := []struct {
		name       string
		symbol     string
		owner      string
		surface    string
		shape      string
		port       string
		params     []string
		results    []string
		wrongParam []string
	}{
		{
			name: "fma", symbol: "FMA64DQQBIDWithMode", owner: "Decimal64", surface: "fma_dqq_with_mode",
			shape: "mixed_ternary_mode_flags_dqq", port: "Bid64dqqFma",
			params: []string{"Decimal64BID", "Decimal128BID", "Decimal128BID", "RoundingMode"}, results: []string{"Decimal64BID", "ExceptionFlags"},
			wrongParam: []string{"Decimal128BID", "Decimal64BID", "Decimal128BID", "RoundingMode"},
		},
		{
			name: "sqrt", symbol: "Sqrt128DBIDWithMode", owner: "Decimal128", surface: "sqrt_d_with_mode",
			shape: "mixed_unary_mode_flags_d", port: "Bid128dSqrt",
			params: []string{"Decimal64BID", "RoundingMode"}, results: []string{"Decimal128BID", "ExceptionFlags"},
			wrongParam: []string{"Decimal128BID", "RoundingMode"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			inventory := &inventoryFile{Total: 1, Mapped: 1, Symbols: []inventorySymbol{{Symbol: tc.symbol, Kind: "func", Status: "mapped", BidgoFunction: tc.port}}}
			manifest := &manifestFile{Emit: []emitRule{{GoSymbol: tc.symbol, RustOwner: tc.owner, RustSurface: tc.surface, Shape: tc.shape, BidgoFunction: tc.port, Reason: "test mixed extension"}}, Deferred: deferredBlock{Category: "deferred"}}
			sigs := map[string]goSig{tc.symbol: {Symbol: tc.symbol, Kind: "func", Name: tc.symbol, Params: tc.params, Results: tc.results}}
			if _, err := resolveClosure(inventory, manifest, sigs); err != nil {
				t.Fatalf("resolve exact mixed extension shape: %v", err)
			}
			sigs[tc.symbol] = goSig{Symbol: tc.symbol, Kind: "func", Name: tc.symbol, Params: tc.wrongParam, Results: tc.results}
			if _, err := resolveClosure(inventory, manifest, sigs); err == nil || !strings.Contains(err.Error(), "requires operand") {
				t.Fatalf("mixed extension operand mismatch error = %v", err)
			}
		})
	}
}

func TestEmitMixedFMAAndSqrtUseAssociatedFunctionsAndOrderedConversions(t *testing.T) {
	var b strings.Builder
	fma := mixedTernaryDecOp{
		decOp:    decOp{method: "fma_dqq_with_mode", module: "bid128_fma", port: "bid64dqq_fma"},
		operands: [3]widthSpec{decimal64Width, decimal128Width, decimal128Width},
	}
	if err := emitMixedTernaryModeFlagsOps(&b, []mixedTernaryDecOp{fma}, decimal64Width); err != nil {
		t.Fatal(err)
	}
	sqrt := mixedUnaryDecOp{decOp: decOp{method: "sqrt_d_with_mode", module: "bid128_sqrt", port: "bid128d_sqrt"}, operand: decimal64Width}
	if err := emitMixedUnaryModeFlagsOps(&b, []mixedUnaryDecOp{sqrt}, decimal128Width); err != nil {
		t.Fatal(err)
	}
	got := b.String()
	fmaParams, fmaBody := rustGeneratedFunction(t, got, "fma_dqq_with_mode")
	assertRustStringList(t, "fma_dqq_with_mode params", fmaParams, []string{
		"x: Decimal64", "y: Decimal128", "z: Decimal128", "mode: RoundingMode",
	})
	assertRustGeneratedBoundCall(t, fmaBody, "(bits, raw)", "crate::generated::bid128_fma::bid64dqq_fma", [][]string{{
		"x.to_bits()",
		"super::types::bid_uint128_from_le_bytes(y.to_le_bytes())",
		"super::types::bid_uint128_from_le_bytes(z.to_le_bytes())",
		"super::types::to_bidgo_rounding(mode)",
	}})

	sqrtParams, sqrtBody := rustGeneratedFunction(t, got, "sqrt_d_with_mode")
	assertRustStringList(t, "sqrt_d_with_mode params", sqrtParams, []string{
		"value: Decimal64", "mode: RoundingMode",
	})
	assertRustGeneratedBoundCall(t, sqrtBody, "(bits, raw)", "crate::generated::bid128_sqrt::bid128d_sqrt", [][]string{{
		"value.to_bits()",
		"super::types::to_bidgo_rounding(mode)",
	}})
}

func rustGeneratedFunction(t *testing.T, source, name string) ([]string, string) {
	t.Helper()
	needle := "pub fn " + name + "("
	start := strings.Index(source, needle)
	if start < 0 {
		t.Fatalf("generated Rust has no function %q:\n%s", name, source)
	}
	if strings.Index(source[start+len(needle):], needle) >= 0 {
		t.Fatalf("generated Rust has duplicate function %q:\n%s", name, source)
	}
	paramsOpen := start + len(needle) - 1
	paramsClose, err := matchingGeneratedRustDelimiter(source, paramsOpen, '(', ')')
	if err != nil {
		t.Fatalf("parse generated Rust function %q params: %v", name, err)
	}
	params, err := splitGeneratedRustArgs(source[paramsOpen+1 : paramsClose])
	if err != nil {
		t.Fatalf("parse generated Rust function %q params: %v", name, err)
	}
	bodyRel := strings.IndexByte(source[paramsClose+1:], '{')
	if bodyRel < 0 {
		t.Fatalf("generated Rust function %q has no body", name)
	}
	bodyOpen := paramsClose + 1 + bodyRel
	bodyClose, err := matchingGeneratedRustDelimiter(source, bodyOpen, '{', '}')
	if err != nil {
		t.Fatalf("parse generated Rust function %q body: %v", name, err)
	}
	return params, source[bodyOpen+1 : bodyClose]
}

func rustGeneratedBoundCallArgs(source, binding, callee string) ([][]string, error) {
	if strings.Contains(source, "//") || strings.Contains(source, "/*") {
		return nil, fmt.Errorf("generated Rust function contains comments; structural call parser must be extended before accepting them")
	}
	needle := "let " + binding + " = " + callee + "("
	var calls [][]string
	for searchFrom := 0; ; {
		rel := strings.Index(source[searchFrom:], needle)
		if rel < 0 {
			break
		}
		start := searchFrom + rel
		lineStart := strings.LastIndex(source[:start], "\n") + 1
		if strings.TrimSpace(source[lineStart:start]) != "" {
			searchFrom = start + len(needle)
			continue
		}
		open := start + len(needle) - 1
		close, err := matchingGeneratedRustDelimiter(source, open, '(', ')')
		if err != nil {
			return nil, err
		}
		lineEnd := strings.IndexByte(source[close+1:], '\n')
		if lineEnd < 0 {
			lineEnd = len(source)
		} else {
			lineEnd += close + 1
		}
		if strings.TrimSpace(source[close+1:lineEnd]) != ";" {
			searchFrom = close + 1
			continue
		}
		args, err := splitGeneratedRustArgs(source[open+1 : close])
		if err != nil {
			return nil, err
		}
		calls = append(calls, args)
		searchFrom = close + 1
	}
	return calls, nil
}

func matchingGeneratedRustDelimiter(source string, open int, left, right byte) (int, error) {
	depth := 0
	for i := open; i < len(source); i++ {
		switch source[i] {
		case left:
			depth++
		case right:
			depth--
			if depth == 0 {
				return i, nil
			}
		}
	}
	return 0, fmt.Errorf("unbalanced %c%c delimiters", left, right)
}

func splitGeneratedRustArgs(source string) ([]string, error) {
	if strings.TrimSpace(source) == "" {
		return nil, nil
	}
	var args []string
	start := 0
	paren, bracket, brace := 0, 0, 0
	for i := 0; i < len(source); i++ {
		switch source[i] {
		case '(':
			paren++
		case ')':
			paren--
		case '[':
			bracket++
		case ']':
			bracket--
		case '{':
			brace++
		case '}':
			brace--
		case ',':
			if paren == 0 && bracket == 0 && brace == 0 {
				args = append(args, strings.TrimSpace(source[start:i]))
				start = i + 1
			}
		}
		if paren < 0 || bracket < 0 || brace < 0 {
			return nil, fmt.Errorf("unbalanced nested delimiters in %q", source)
		}
	}
	if paren != 0 || bracket != 0 || brace != 0 {
		return nil, fmt.Errorf("unbalanced nested delimiters in %q", source)
	}
	args = append(args, strings.TrimSpace(source[start:]))
	return args, nil
}

func assertRustGeneratedBoundCall(t *testing.T, source, binding, callee string, want [][]string) {
	t.Helper()
	if total := strings.Count(source, "let "+binding+" = "); total != len(want) {
		t.Fatalf("generated Rust binding %s call count = %d, want %d; every bound call must use %s", binding, total, len(want), callee)
	}
	got, err := rustGeneratedBoundCallArgs(source, binding, callee)
	if err != nil {
		t.Fatalf("parse generated Rust call %s: %v", callee, err)
	}
	assertRustStringMatrix(t, callee+" args", got, want)
}

func assertRustStringMatrix(t *testing.T, label string, got, want [][]string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s row count = %d, want %d; got %v", label, len(got), len(want), got)
	}
	for i := range want {
		assertRustStringList(t, fmt.Sprintf("%s[%d]", label, i), got[i], want[i])
	}
}

func assertRustStringList(t *testing.T, label string, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s length = %d, want %d; got %v", label, len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("%s[%d] = %q, want %q", label, i, got[i], want[i])
		}
	}
}

// TestShapeSigsCoverManifestShapes keeps shapeSigs closed against the checked-in
// manifest: every emit rule's shape must have a signature spec, so a new shape
// cannot be emitted without a validator.
func TestShapeSigsCoverManifestShapes(t *testing.T) {
	manifest, err := loadManifest(filepath.Join("..", "..", "registry", "rust_api_manifest.json"))
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	for _, r := range manifest.Emit {
		if _, ok := shapeSigs[r.Shape]; !ok {
			t.Errorf("manifest emit shape %q (go_symbol %q) has no shapeSigs signature spec", r.Shape, r.GoSymbol)
		}
	}
}

func TestLoadGoPublicSigsRejectsBuildVariantKindConflict(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"a_mutable.go":  "//go:build mutable\n\npackage bid754\n\nvar Zero64BID Decimal64BID\n",
		"z_accessor.go": "//go:build !mutable\n\npackage bid754\n\nfunc Zero64BID() Decimal64BID { return 0 }\n",
	}
	for name, source := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(source), 0o600); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	_, err := loadGoPublicSigs(dir)
	if err == nil || !strings.Contains(err.Error(), "conflicting declarations") {
		t.Fatalf("loadGoPublicSigs(build-tag var/func conflict) error = %v, want conflict rejection", err)
	}
}

func TestLoadGoPublicSigsAllowsExactBuildVariantDuplicate(t *testing.T) {
	dir := t.TempDir()
	source := "package bid754\n\nfunc Transform(values map[string]int) map[string]int { return values }\n"
	for _, name := range []string{"a.go", "z.go"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(source), 0o600); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	symbols, err := loadGoPublicSigs(dir)
	if err != nil {
		t.Fatalf("loadGoPublicSigs(exact duplicate): %v", err)
	}
	if len(symbols) != 1 || symbols["Transform"].Symbol != "Transform" {
		t.Fatalf("exact duplicate symbols = %+v, want one Transform entry", symbols)
	}
}

func TestLoadGoPublicSigsRejectsBuildVariantSignatureConflicts(t *testing.T) {
	tests := []struct {
		name string
		a    string
		z    string
	}{
		{
			name: "map types",
			a:    "package bid754\n\nfunc Transform(values map[string]int) {}\n",
			z:    "package bid754\n\nfunc Transform(values map[int]string) {}\n",
		},
		{
			name: "receiver pointer",
			a:    "package bid754\n\ntype Decimal64BID uint64\n\nfunc (d Decimal64BID) Transform() {}\n",
			z:    "package bid754\n\ntype Decimal64BID uint64\n\nfunc (d *Decimal64BID) Transform() {}\n",
		},
		{
			name: "generic constraint",
			a:    "package bid754\n\nfunc Transform[T ~int](value T) {}\n",
			z:    "package bid754\n\nfunc Transform[T ~string](value T) {}\n",
		},
		{
			name: "generic parameter order",
			a:    "package bid754\n\nfunc Transform[T, U any](x T, y U) {}\n",
			z:    "package bid754\n\nfunc Transform[U, T any](x T, y U) {}\n",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			for name, source := range map[string]string{"a.go": tc.a, "z.go": tc.z} {
				if err := os.WriteFile(filepath.Join(dir, name), []byte(source), 0o600); err != nil {
					t.Fatalf("write %s: %v", name, err)
				}
			}
			_, err := loadGoPublicSigs(dir)
			if err == nil || !strings.Contains(err.Error(), "conflicting declarations") {
				t.Fatalf("loadGoPublicSigs(signature conflict) error = %v, want conflict rejection", err)
			}
		})
	}
}

// TestConstDocBlurbCoversManifestConstants mirrors TestShapeSigsCoverManifestShapes
// for the constants closure (associated-constants): every constants row's rust_const must have a
// constDocBlurb entry, so a new constant kind cannot be emitted without rustdoc.
func TestConstDocBlurbCoversManifestConstants(t *testing.T) {
	manifest, err := loadManifest(filepath.Join("..", "..", "registry", "rust_api_manifest.json"))
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	if len(manifest.Constants) == 0 {
		t.Fatalf("manifest has no constants rows")
	}
	for _, r := range manifest.Constants {
		if _, ok := constDocBlurb[r.RustConst]; !ok {
			t.Errorf("manifest constants rust_const %q (go_symbol %q) has no constDocBlurb entry", r.RustConst, r.GoSymbol)
		}
	}
}

// -- resolveConstantsClosure --

// baseConstInputs builds a minimal consistent census/manifest/goLiterals
// triple for the constants closure: two excluded_constant_accessor symbols, one
// covered by a constants row and one by excluded_constants (mirrors
// baseInputs' one-emitted/one-deferred shape for the mapped-symbol closure).
func baseConstInputs() (*inventoryFile, *manifestFile, map[string]goConstantAccessorLiteral) {
	inventory := &inventoryFile{
		Total:    2,
		Mapped:   0,
		Excluded: 2,
		Symbols: []inventorySymbol{
			{Symbol: "Zero64BID", Kind: "func", Status: "excluded_constant_accessor"},
			{Symbol: "One64BID", Kind: "func", Status: "excluded_constant_accessor"},
		},
	}
	manifest := &manifestFile{
		Constants: []constantRule{
			{GoSymbol: "Zero64BID", RustOwner: "Decimal64", RustConst: "ZERO", Literal: "0", Reason: "zero constant"},
		},
		ExcludedConstants: deferredBlock{
			Category: "excluded_constants",
			Symbols:  []string{"One64BID"},
		},
	}
	goLiterals := map[string]goConstantAccessorLiteral{
		"Zero64BID": {Literal: "0", GoType: "Decimal64BID"},
		"One64BID":  {Literal: "1", GoType: "Decimal64BID"},
	}
	return inventory, manifest, goLiterals
}

func TestResolveConstantsClosureAcceptsExactCover(t *testing.T) {
	inventory, manifest, goLiterals := baseConstInputs()
	rows, err := resolveConstantsClosure(inventory, manifest, goLiterals)
	if err != nil {
		t.Fatalf("expected exact cover to resolve, got: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 constant surface rows, got %d", len(rows))
	}
	byStatus := map[string]int{}
	var bitsHex string
	for _, r := range rows {
		byStatus[r.Status]++
		if r.GoSymbol == "Zero64BID" {
			bitsHex = r.BitsHex
		}
	}
	if byStatus["emitted"] != 1 || byStatus["excluded"] != 1 {
		t.Fatalf("expected 1 emitted + 1 excluded, got %v", byStatus)
	}
	if bitsHex != "0x31c0000000000000" {
		t.Fatalf("expected Zero64BID bits_hex 0x31c0000000000000, got %q", bitsHex)
	}
}

func TestResolveConstantsClosureRejectsUncoveredSymbol(t *testing.T) {
	inventory, manifest, goLiterals := baseConstInputs()
	manifest.ExcludedConstants.Symbols = nil
	_, err := resolveConstantsClosure(inventory, manifest, goLiterals)
	if err == nil || !strings.Contains(err.Error(), "constants incomplete mapping") {
		t.Fatalf("expected a constants incomplete mapping for an uncovered symbol, got: %v", err)
	}
}

func TestResolveConstantsClosureRejectsBothEmittedAndExcluded(t *testing.T) {
	inventory, manifest, goLiterals := baseConstInputs()
	manifest.ExcludedConstants.Symbols = []string{"One64BID", "Zero64BID"}
	_, err := resolveConstantsClosure(inventory, manifest, goLiterals)
	if err == nil || !strings.Contains(err.Error(), "both a constants row and excluded_constants") {
		t.Fatalf("expected a both-emitted-and-excluded failure, got: %v", err)
	}
}

func TestResolveConstantsClosureRejectsNonCensusConstant(t *testing.T) {
	inventory, manifest, goLiterals := baseConstInputs()
	manifest.Constants[0].GoSymbol = "NonexistentBID"
	goLiterals["NonexistentBID"] = goConstantAccessorLiteral{Literal: "0", GoType: "Decimal64BID"}
	_, err := resolveConstantsClosure(inventory, manifest, goLiterals)
	if err == nil || !strings.Contains(err.Error(), "not an excluded_constant_accessor census symbol") {
		t.Fatalf("expected a non-census constants failure, got: %v", err)
	}
}

func TestResolveConstantsClosureRejectsUnknownOwner(t *testing.T) {
	inventory, manifest, goLiterals := baseConstInputs()
	manifest.Constants[0].RustOwner = "Decimal65"
	_, err := resolveConstantsClosure(inventory, manifest, goLiterals)
	if err == nil || !strings.Contains(err.Error(), "unsupported rust_owner") {
		t.Fatalf("expected an unsupported-owner failure, got: %v", err)
	}
}

func TestResolveConstantsClosureRejectsGoRustWidthMismatch(t *testing.T) {
	inventory, manifest, goLiterals := baseConstInputs()
	goLiterals["Zero64BID"] = goConstantAccessorLiteral{Literal: "0", GoType: "Decimal32BID"}
	_, err := resolveConstantsClosure(inventory, manifest, goLiterals)
	if err == nil || !strings.Contains(err.Error(), "requires Go result Decimal64BID") {
		t.Fatalf("expected a Go/Rust constant width mismatch, got: %v", err)
	}
}

func TestResolveConstantsClosureRejectsLiteralDriftFromGoSource(t *testing.T) {
	inventory, manifest, goLiterals := baseConstInputs()
	// bid754-go's actual Zero64BID literal ("0") disagrees with a stale
	// manifest literal.
	manifest.Constants[0].Literal = "0.0"
	_, err := resolveConstantsClosure(inventory, manifest, goLiterals)
	if err == nil || !strings.Contains(err.Error(), "disagrees with bid754-go source") {
		t.Fatalf("expected a literal-drift failure, got: %v", err)
	}
}

func TestResolveConstantsClosureRejectsMissingGoLiteral(t *testing.T) {
	inventory, manifest, goLiterals := baseConstInputs()
	delete(goLiterals, "Zero64BID")
	_, err := resolveConstantsClosure(inventory, manifest, goLiterals)
	if err == nil || !strings.Contains(err.Error(), "no parsed bid754-go accessor literal") {
		t.Fatalf("expected a missing-literal failure, got: %v", err)
	}
}

func TestResolveConstantsClosureRejectsWrongExcludedCategory(t *testing.T) {
	inventory, manifest, goLiterals := baseConstInputs()
	manifest.ExcludedConstants.Category = "temporary"
	_, err := resolveConstantsClosure(inventory, manifest, goLiterals)
	if err == nil || !strings.Contains(err.Error(), "excluded_constants.category") {
		t.Fatalf("expected an excluded_constants category failure, got: %v", err)
	}
}

func TestResolveConstantsClosureRejectsDuplicateGoSymbol(t *testing.T) {
	inventory, manifest, goLiterals := baseConstInputs()
	manifest.Constants = append(manifest.Constants, constantRule{
		GoSymbol: "Zero64BID", RustOwner: "Decimal64", RustConst: "ZERO", Literal: "0", Reason: "dup",
	})
	_, err := resolveConstantsClosure(inventory, manifest, goLiterals)
	if err == nil || !strings.Contains(err.Error(), "lists go_symbol") {
		t.Fatalf("expected a duplicate go_symbol failure, got: %v", err)
	}
}

// TestRustAPIManifestConstantsCoverCensus regenerates the constants closure
// against the checked-in census, manifest, and actual bid754-go accessor AST.
// This pins that the 12
// excluded_constant_accessor census symbols and the manifest's constants/
// excluded_constants sections stay in exact agreement.
func TestRustAPIManifestConstantsCoverCensus(t *testing.T) {
	inventory, err := loadInventory(filepath.Join("..", "..", "..", "generated", "testspec", "public_api_routing_inventory.json"))
	if err != nil {
		t.Fatalf("load inventory: %v", err)
	}
	manifest, err := loadManifest(filepath.Join("..", "..", "registry", "rust_api_manifest.json"))
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	goLiterals, err := loadGoConstantLiterals(filepath.Join("..", "..", "..", "..", "bid754-go"))
	if err != nil {
		t.Fatalf("load Go constant accessors: %v", err)
	}
	rows, err := resolveConstantsClosure(inventory, manifest, goLiterals)
	if err != nil {
		t.Fatalf("resolveConstantsClosure: %v", err)
	}
	if len(rows) != 12 {
		t.Fatalf("expected 12 constant surface rows, got %d", len(rows))
	}
	for _, r := range rows {
		if r.Status != "emitted" {
			t.Errorf("constants row %q has status %q, want emitted (no excluded_constants entries expected today)", r.GoSymbol, r.Status)
		}
	}
}

func TestLoadGoConstantLiteralsFindsImmutableAccessors(t *testing.T) {
	got, err := loadGoConstantLiterals(filepath.Join("..", "..", "..", "..", "bid754-go"))
	if err != nil {
		t.Fatalf("loadGoConstantLiterals: %v", err)
	}
	want := map[string]goConstantAccessorLiteral{
		"Zero32BID":  {Literal: "0", GoType: "Decimal32BID"},
		"Zero64BID":  {Literal: "0", GoType: "Decimal64BID"},
		"Zero128BID": {Literal: "0", GoType: "Decimal128BID"},
		"One32BID":   {Literal: "1", GoType: "Decimal32BID"},
		"One64BID":   {Literal: "1", GoType: "Decimal64BID"},
		"One128BID":  {Literal: "1", GoType: "Decimal128BID"},
		"Pi32BID":    {Literal: "3.141593", GoType: "Decimal32BID"},
		"Pi64BID":    {Literal: "3.141592653589793", GoType: "Decimal64BID"},
		"Pi128BID":   {Literal: "3.141592653589793238462643383279503", GoType: "Decimal128BID"},
		"E32BID":     {Literal: "2.718282", GoType: "Decimal32BID"},
		"E64BID":     {Literal: "2.718281828459045", GoType: "Decimal64BID"},
		"E128BID":    {Literal: "2.718281828459045235360287471352662", GoType: "Decimal128BID"},
	}
	if len(got) != len(want) {
		t.Fatalf("constant accessor literal count = %d, want %d: %v", len(got), len(want), got)
	}
	for symbol, expected := range want {
		if got[symbol] != expected {
			t.Errorf("constant accessor %s = %+v, want %+v", symbol, got[symbol], expected)
		}
	}
}

func TestLoadGoConstantLiteralsRejectsAccessorBackingWidthMismatch(t *testing.T) {
	dir := t.TempDir()
	source := "package bid754\n\nvar zero64BID, _ = NewDecimal32BIDDirect(\"0\")\n\nfunc Zero64BID() Decimal64BID { return zero64BID }\n"
	if err := os.WriteFile(filepath.Join(dir, "constants.go"), []byte(source), 0o600); err != nil {
		t.Fatalf("write constants.go: %v", err)
	}

	_, err := loadGoConstantLiterals(dir)
	if err == nil || !strings.Contains(err.Error(), "returns Decimal64BID") || !strings.Contains(err.Error(), "initialized as Decimal32BID") {
		t.Fatalf("loadGoConstantLiterals(width mismatch) error = %v, want accessor/backing width rejection", err)
	}
}

func TestLoadGoConstantLiteralsRejectsNamedResultShadow(t *testing.T) {
	dir := t.TempDir()
	source := "package bid754\n\nvar zero64BID, _ = NewDecimal64BIDDirect(\"0\")\n\nfunc Zero64BID() (zero64BID Decimal64BID) { return zero64BID }\n"
	if err := os.WriteFile(filepath.Join(dir, "constants.go"), []byte(source), 0o600); err != nil {
		t.Fatalf("write constants.go: %v", err)
	}

	_, err := loadGoConstantLiterals(dir)
	if err == nil || !strings.Contains(err.Error(), "must use an unnamed result") {
		t.Fatalf("loadGoConstantLiterals(named-result shadow) error = %v, want shadow rejection", err)
	}
}

func TestLoadGoConstantLiteralsRejectsGenericAccessor(t *testing.T) {
	dir := t.TempDir()
	source := "package bid754\n\nvar zero64BID, _ = NewDecimal64BIDDirect(\"0\")\n\nfunc Zero64BID[T any]() Decimal64BID { return zero64BID }\n"
	if err := os.WriteFile(filepath.Join(dir, "constants.go"), []byte(source), 0o600); err != nil {
		t.Fatalf("write constants.go: %v", err)
	}

	_, err := loadGoConstantLiterals(dir)
	if err == nil || !strings.Contains(err.Error(), "must not declare type parameters") {
		t.Fatalf("loadGoConstantLiterals(generic accessor) error = %v, want type-parameter rejection", err)
	}
}

func TestLoadGoConstantLiteralsRejectsBuildVariantLiteralConflict(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"a_zero.go": "//go:build zero\n\npackage bid754\n\nvar zero64BID, _ = NewDecimal64BIDDirect(\"0\")\n\nfunc Zero64BID() Decimal64BID { return zero64BID }\n",
		"z_one.go":  "//go:build !zero\n\npackage bid754\n\nvar zero64BID, _ = NewDecimal64BIDDirect(\"1\")\n\nfunc Zero64BID() Decimal64BID { return zero64BID }\n",
	}
	for name, source := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(source), 0o600); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	_, err := loadGoConstantLiterals(dir)
	if err == nil || !strings.Contains(err.Error(), "conflicting build-variant definitions") {
		t.Fatalf("loadGoConstantLiterals(build-tag literal conflict) error = %v, want conflict rejection", err)
	}
}

// The public API subtree carries a no-panic contract, and `&s[a..b]` on a &str
// panics on a non-boundary index. Clippy's type-aware string_slice restriction
// lint is what enforces that (make test-rust runs cargo clippy), so mod.rs must
// keep forbidding it at module scope: forbid, not deny, so no inner attribute
// can re-allow it, and placed before the first item so it covers the whole
// subtree. Losing the attribute would silently disarm that gate.
func TestBuildModRsForbidsStringSliceAtModuleScope(t *testing.T) {
	const want = "#![forbid(clippy::string_slice)]"

	out := buildModRs([]string{"context.rs", "types.rs"})

	lines := strings.Split(out, "\n")
	attrIndex, firstItemIndex := -1, -1
	for i, line := range lines {
		if line == want && attrIndex == -1 {
			attrIndex = i
		}
		if strings.HasPrefix(line, "mod ") && firstItemIndex == -1 {
			firstItemIndex = i
		}
	}
	if attrIndex == -1 {
		t.Fatalf("buildModRs output missing %q; the Clippy string-slice gate would be silently disarmed:\n%s", want, out)
	}
	if firstItemIndex == -1 {
		t.Fatalf("buildModRs output declared no `mod` item:\n%s", out)
	}
	if attrIndex > firstItemIndex {
		t.Fatalf("buildModRs put %q at line %d, after the first item at line %d; an inner attribute must precede all items to scope the whole subtree", want, attrIndex+1, firstItemIndex+1)
	}
	// A `deny`/`allow` variant would compile but let a later inner attribute
	// re-enable string slicing, so pin the level word too.
	if strings.Contains(out, "deny(clippy::string_slice)") || strings.Contains(out, "allow(clippy::string_slice)") {
		t.Fatalf("buildModRs weakened the string_slice level away from forbid:\n%s", out)
	}
}

// The forbid attribute above is only enforced if something actually runs
// Clippy, and the `-D unknown_lints -D renamed_and_removed_lints` tail is what
// keeps a renamed or removed lint from being swallowed by `-A warnings`. Pin
// the whole invocation inside the test-rust recipe: a matching comment or a
// match in some other target must not satisfy this test.
func TestMakefileTestRustRunsHardenedClippy(t *testing.T) {
	// Exact executed recipe line, tab and subshell wrapper included: a comment or
	// `@true # ...` line that merely mentions the command must not satisfy this.
	const wantLine = "\t@(cd bid754-rs && cargo clippy --locked --lib -- -A warnings -D unknown_lints -D renamed_and_removed_lints)"

	path := filepath.Join("..", "..", "..", "..", "Makefile")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	lines := strings.Split(string(data), "\n")
	start := -1
	for i, line := range lines {
		if line == "test-rust:" {
			start = i + 1
			break
		}
	}
	if start == -1 {
		t.Fatalf("%s has no `test-rust:` target line", path)
	}

	matches := 0
	body := 0
	for _, line := range lines[start:] {
		if strings.TrimSpace(line) == "" {
			continue
		}
		// A recipe line is indented; the first non-indented nonblank line ends
		// the target body.
		if !strings.HasPrefix(line, "\t") && !strings.HasPrefix(line, " ") {
			break
		}
		body++
		if line == wantLine {
			matches++
		}
	}
	if body == 0 {
		t.Fatalf("%s: test-rust target body is empty", path)
	}
	if matches != 1 {
		t.Fatalf("%s: test-rust body has %d lines exactly equal to %q, want exactly 1", path, matches, wantLine)
	}
}
