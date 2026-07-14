package apiemit

import (
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
	for _, want := range []string{
		"pub fn add_dq_with_mode(left: Decimal64, right: Decimal128, mode: RoundingMode)",
		"left.to_bits()",
		"bid_uint128_from_le_bytes(right.to_le_bytes())",
		"crate::generated::bid128_add::bid64dq_add",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("mixed wrapper output missing %q:\n%s", want, got)
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
