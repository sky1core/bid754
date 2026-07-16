package testgen

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestLoadPublicAPISigsRejectsBuildVariantKindConflict(t *testing.T) {
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

	_, err := loadPublicAPISigs(dir)
	if err == nil || !strings.Contains(err.Error(), "conflicting declarations") {
		t.Fatalf("loadPublicAPISigs(build-tag var/func conflict) error = %v, want conflict rejection", err)
	}
}

func TestLoadPublicAPISigsAllowsExactBuildVariantDuplicate(t *testing.T) {
	dir := t.TempDir()
	source := "package bid754\n\nfunc Transform(values map[string]int) map[string]int { return values }\n"
	for _, name := range []string{"a.go", "z.go"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(source), 0o600); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	symbols, err := loadPublicAPISigs(dir)
	if err != nil {
		t.Fatalf("loadPublicAPISigs(exact duplicate): %v", err)
	}
	if len(symbols) != 1 || symbols[0].Symbol != "Transform" {
		t.Fatalf("exact duplicate symbols = %+v, want one Transform entry", symbols)
	}
}

func TestLoadPublicAPISigsRejectsBuildVariantSignatureConflicts(t *testing.T) {
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
			_, err := loadPublicAPISigs(dir)
			if err == nil || !strings.Contains(err.Error(), "conflicting declarations") {
				t.Fatalf("loadPublicAPISigs(signature conflict) error = %v, want conflict rejection", err)
			}
		})
	}
}

func TestResolvePublicAPIRejectsExportedMutableVariables(t *testing.T) {
	symbols := []publicAPISymbol{{Symbol: "Zero64BID", Kind: "var", Name: "Zero64BID"}}
	spec := &PublicAPISpec{SymbolExclusions: []PublicSymbolExclusion{{
		Symbol: "Zero64BID",
		Class:  "constant_accessor",
		Reason: "test exclusion must not make an exported mutable variable acceptable",
	}}}

	_, err := resolvePublicAPI(symbols, nil, spec)
	if err == nil || !strings.Contains(err.Error(), "exported mutable variable") {
		t.Fatalf("resolvePublicAPI(exported var) error = %v, want exported-mutable-variable rejection", err)
	}
}

func TestResolvePublicAPIAcceptsConstantAccessorFunction(t *testing.T) {
	symbols := []publicAPISymbol{{
		Symbol:  "Zero64BID",
		Kind:    "func",
		Name:    "Zero64BID",
		Results: []string{"Decimal64BID"},
	}}
	spec := &PublicAPISpec{SymbolExclusions: []PublicSymbolExclusion{{
		Symbol: "Zero64BID",
		Class:  "constant_accessor",
		Reason: "returns a copy of the package-private decimal constant",
	}}}

	inventory, err := resolvePublicAPI(symbols, nil, spec)
	if err != nil {
		t.Fatalf("resolvePublicAPI(constant accessor): %v", err)
	}
	if inventory.Total != 1 || inventory.Excluded != 1 || len(inventory.Symbols) != 1 {
		t.Fatalf("constant accessor inventory = %+v, want one excluded public function", inventory)
	}
	row := inventory.Symbols[0]
	if row.Kind != "func" || row.Status != "excluded_constant_accessor" {
		t.Fatalf("constant accessor row = %+v, want func/excluded_constant_accessor", row)
	}
}

func TestResolvePublicAPIRejectsGenericConstantAccessor(t *testing.T) {
	symbols := []publicAPISymbol{{
		Symbol:     "Zero64BID",
		Kind:       "func",
		Name:       "Zero64BID",
		TypeParams: []string{"T any"},
		Results:    []string{"Decimal64BID"},
	}}
	spec := &PublicAPISpec{SymbolExclusions: []PublicSymbolExclusion{{
		Symbol: "Zero64BID",
		Class:  "constant_accessor",
		Reason: "test generic accessors are not immutable constant plumbing",
	}}}

	_, err := resolvePublicAPI(symbols, nil, spec)
	if err == nil || !strings.Contains(err.Error(), "must be a non-generic package function") {
		t.Fatalf("resolvePublicAPI(generic constant accessor) error = %v, want exact accessor-shape rejection", err)
	}
}

func TestPublicParityFiniteCohortOracle(t *testing.T) {
	tests := []struct {
		input       string
		wantQuantum string
		wantDigits  int
		ok          bool
	}{
		{input: "1", wantQuantum: "0", wantDigits: 1, ok: true},
		{input: "1.25", wantQuantum: "-2", wantDigits: 3, ok: true},
		{input: ".25e3", wantQuantum: "1", wantDigits: 2, ok: true},
		{input: "1.e-4", wantQuantum: "-4", wantDigits: 1, ok: true},
		{input: " \t-1.0E+91", wantQuantum: "90", wantDigits: 2, ok: true},
		{input: "0001000.00", wantQuantum: "-2", wantDigits: 6, ok: true},
		{input: "0.000", wantQuantum: "-3", wantDigits: 0, ok: true},
		{input: "1000000.0", wantQuantum: "-1", wantDigits: 8, ok: true},
		{input: "1e9223372036854775808", wantQuantum: "9223372036854775808", wantDigits: 1, ok: true},
		{input: "1e-9223372036854775809", wantQuantum: "-9223372036854775809", wantDigits: 1, ok: true},
		{input: "Infinity"},
		{input: "1e"},
		{input: "1.0 "},
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got, ok := publicParityFiniteCohort(tc.input)
			if ok != tc.ok {
				t.Fatalf("publicParityFiniteCohort(%q) ok = %v, want %v", tc.input, ok, tc.ok)
			}
			if !ok {
				return
			}
			if got.Quantum.String() != tc.wantQuantum || got.CoefficientDigits != tc.wantDigits {
				t.Fatalf("publicParityFiniteCohort(%q) = (quantum=%s, digits=%d), want (%s, %d)", tc.input, got.Quantum, got.CoefficientDigits, tc.wantQuantum, tc.wantDigits)
			}
		})
	}
}

func TestPublicParityStringCorpusMetadataMatchesIndependentOracles(t *testing.T) {
	if err := validatePublicParityStringCorpus(); err != nil {
		t.Fatal(err)
	}
}

func TestResolveParityUnitClassifiesMixedArithmeticExactly(t *testing.T) {
	sym := publicAPISymbol{
		Symbol:  "Add64DQBIDWithMode",
		Kind:    "func",
		Name:    "Add64DQBIDWithMode",
		Params:  []string{"Decimal64BID", "Decimal128BID", "RoundingMode"},
		Results: []string{"Decimal64BID", "ExceptionFlags"},
	}
	sigs := map[string]bidgoFuncSig{
		"Bid64dqAdd": {
			Name:    "Bid64dqAdd",
			Params:  []bidgoParam{{Name: "x", Type: "uint64"}, {Name: "y", Type: "BID_UINT128"}, {Name: "rnd", Type: "int"}},
			Results: []bidgoParam{{Type: "uint64"}, {Type: "uint32"}},
		},
	}
	unit, err := resolveParityUnit(sym, "Bid64dqAdd", sigs, publicParityCorpus{})
	if err != nil {
		t.Fatalf("resolve mixed parity unit: %v", err)
	}
	if unit.Shape != shapeFuncMixedModeBinary || unit.Width != 64 || unit.OperandWidths != [2]int{64, 128} || unit.Operation != "Add" {
		t.Fatalf("mixed parity unit = %+v", unit)
	}
	wantCases := (len(parityLabelPairs)+4)*len(parityModeOrder) + 1
	if unit.Cases != wantCases {
		t.Fatalf("mixed parity cases = %d, want %d", unit.Cases, wantCases)
	}
}

func TestResolveParityUnitRejectsMixedSuffixSignatureMismatch(t *testing.T) {
	sym := publicAPISymbol{
		Symbol:  "Div128QDBIDWithMode",
		Kind:    "func",
		Name:    "Div128QDBIDWithMode",
		Params:  []string{"Decimal64BID", "Decimal128BID", "RoundingMode"},
		Results: []string{"Decimal128BID", "ExceptionFlags"},
	}
	_, err := resolveParityUnit(sym, "Bid128qdDiv", nil, publicParityCorpus{})
	if err == nil || !strings.Contains(err.Error(), "want Decimal128BID from suffix QD") {
		t.Fatalf("suffix/signature mismatch error = %v", err)
	}
}

func TestResolveParityUnitClassifiesMixedFMAExactly(t *testing.T) {
	tests := []struct {
		resultWidth int
		operandCode string
	}{
		{64, "DDQ"}, {64, "DQD"}, {64, "DQQ"}, {64, "QDD"},
		{64, "QDQ"}, {64, "QQD"}, {64, "QQQ"},
		{128, "DDD"}, {128, "DDQ"}, {128, "DQD"}, {128, "DQQ"},
		{128, "QDD"}, {128, "QDQ"}, {128, "QQD"},
	}
	for _, tc := range tests {
		name := "FMA" + strconv.Itoa(tc.resultWidth) + tc.operandCode + "BIDWithMode"
		t.Run(name, func(t *testing.T) {
			params := make([]string, 0, 4)
			portParams := make([]bidgoParam, 0, 4)
			portValueTypes := make([]string, 0, 3)
			var wantWidths [3]int
			for i, code := range tc.operandCode {
				width := 64
				portType := "uint64"
				if code == 'Q' {
					width = 128
					portType = "BID_UINT128"
				}
				wantWidths[i] = width
				params = append(params, "Decimal"+strconv.Itoa(width)+"BID")
				portParams = append(portParams, bidgoParam{Type: portType})
				portValueTypes = append(portValueTypes, portType)
			}
			params = append(params, "RoundingMode")
			portParams = append(portParams, bidgoParam{Name: "rnd_mode", Type: "int"})
			portResult := "uint64"
			if tc.resultWidth == 128 {
				portResult = "BID_UINT128"
			}
			bidgoFn := "Bid" + strconv.Itoa(tc.resultWidth) + strings.ToLower(tc.operandCode) + "Fma"
			sym := publicAPISymbol{
				Symbol:  name,
				Kind:    "func",
				Name:    name,
				Params:  params,
				Results: []string{"Decimal" + strconv.Itoa(tc.resultWidth) + "BID", "ExceptionFlags"},
			}
			sigs := map[string]bidgoFuncSig{bidgoFn: {
				Name:    bidgoFn,
				Params:  portParams,
				Results: []bidgoParam{{Type: portResult}, {Type: "uint32"}},
			}}

			unit, err := resolveParityUnit(sym, bidgoFn, sigs, publicParityCorpus{})
			if err != nil {
				t.Fatalf("resolve mixed FMA parity unit: %v", err)
			}
			if unit.Shape != shapeFuncMixedModeTernary || unit.Width != tc.resultWidth || unit.TernaryOperandWidths != wantWidths || unit.Operation != "FMA" {
				t.Fatalf("mixed FMA parity unit = %+v", unit)
			}
			if strings.Join(unit.Port.ValueParams, ",") != strings.Join(portValueTypes, ",") || !unit.Port.HasRounding || unit.Port.FlagsKind != "result" || unit.Port.PrimaryResult != portResult {
				t.Fatalf("mixed FMA port plan = %+v, want value params %v, rounding, result flags, primary %s", unit.Port, portValueTypes, portResult)
			}
			wantCases := (len(parityLabelTriples)+4)*len(parityModeOrder) + 1
			if unit.Cases != wantCases {
				t.Fatalf("mixed FMA cases = %d, want %d", unit.Cases, wantCases)
			}
		})
	}
}

func TestResolveParityUnitClassifiesMixedSqrtExactly(t *testing.T) {
	tests := []struct {
		name         string
		resultWidth  int
		operandWidth int
		bidgoFn      string
		portParam    string
		portResult   string
	}{
		{"Sqrt64QBIDWithMode", 64, 128, "Bid64qSqrt", "BID_UINT128", "uint64"},
		{"Sqrt128DBIDWithMode", 128, 64, "Bid128dSqrt", "uint64", "BID_UINT128"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sym := publicAPISymbol{
				Symbol:  tc.name,
				Kind:    "func",
				Name:    tc.name,
				Params:  []string{"Decimal" + strconv.Itoa(tc.operandWidth) + "BID", "RoundingMode"},
				Results: []string{"Decimal" + strconv.Itoa(tc.resultWidth) + "BID", "ExceptionFlags"},
			}
			sigs := map[string]bidgoFuncSig{tc.bidgoFn: {
				Name: tc.bidgoFn,
				Params: []bidgoParam{
					{Type: tc.portParam},
					{Name: "rnd_mode", Type: "int"},
				},
				Results: []bidgoParam{{Type: tc.portResult}, {Type: "uint32"}},
			}}

			unit, err := resolveParityUnit(sym, tc.bidgoFn, sigs, publicParityCorpus{})
			if err != nil {
				t.Fatalf("resolve mixed sqrt parity unit: %v", err)
			}
			if unit.Shape != shapeFuncMixedModeUnary || unit.Width != tc.resultWidth || unit.UnaryOperandWidth != tc.operandWidth || unit.Operation != "Sqrt" {
				t.Fatalf("mixed sqrt parity unit = %+v", unit)
			}
			if strings.Join(unit.Port.ValueParams, ",") != tc.portParam || !unit.Port.HasRounding || unit.Port.FlagsKind != "result" || unit.Port.PrimaryResult != tc.portResult {
				t.Fatalf("mixed sqrt port plan = %+v, want value param %s, rounding, result flags, primary %s", unit.Port, tc.portParam, tc.portResult)
			}
			wantCases := (publicParityCorpusLen+4)*len(parityModeOrder) + 1
			if unit.Cases != wantCases {
				t.Fatalf("mixed sqrt cases = %d, want %d", unit.Cases, wantCases)
			}
		})
	}
}

func TestResolveParityUnitRejectsMixedFMASuffixSignatureMismatch(t *testing.T) {
	sym := publicAPISymbol{
		Symbol:  "FMA64DQQBIDWithMode",
		Kind:    "func",
		Name:    "FMA64DQQBIDWithMode",
		Params:  []string{"Decimal64BID", "Decimal64BID", "Decimal128BID", "RoundingMode"},
		Results: []string{"Decimal64BID", "ExceptionFlags"},
	}
	_, err := resolveParityUnit(sym, "Bid64dqqFma", nil, publicParityCorpus{})
	if err == nil || !strings.Contains(err.Error(), "want Decimal128BID from suffix DQQ") {
		t.Fatalf("mixed FMA suffix/signature mismatch error = %v", err)
	}
}

func TestResolveParityUnitRejectsMixedPortMappingMismatch(t *testing.T) {
	tests := []struct {
		name    string
		params  []string
		results []string
		gotPort string
	}{
		{
			name:    "FMA128QDQBIDWithMode",
			params:  []string{"Decimal128BID", "Decimal64BID", "Decimal128BID", "RoundingMode"},
			results: []string{"Decimal128BID", "ExceptionFlags"},
			gotPort: "Bid128dqqFma",
		},
		{
			name:    "Sqrt64QBIDWithMode",
			params:  []string{"Decimal128BID", "RoundingMode"},
			results: []string{"Decimal64BID", "ExceptionFlags"},
			gotPort: "Bid128dSqrt",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sym := publicAPISymbol{Symbol: tc.name, Kind: "func", Name: tc.name, Params: tc.params, Results: tc.results}
			_, err := resolveParityUnit(sym, tc.gotPort, nil, publicParityCorpus{})
			if err == nil || !strings.Contains(err.Error(), "must map to") {
				t.Fatalf("mixed port mapping mismatch error = %v", err)
			}
		})
	}
}

func TestResolveParityUnitRejectsUnsupportedMixedFMAAndSqrtSurfaces(t *testing.T) {
	tests := []publicAPISymbol{
		{
			Symbol: "FMA64DDDBIDWithMode", Kind: "func", Name: "FMA64DDDBIDWithMode",
			Params: []string{"Decimal64BID", "Decimal64BID", "Decimal64BID", "RoundingMode"}, Results: []string{"Decimal64BID", "ExceptionFlags"},
		},
		{
			Symbol: "FMA128QQQBIDWithMode", Kind: "func", Name: "FMA128QQQBIDWithMode",
			Params: []string{"Decimal128BID", "Decimal128BID", "Decimal128BID", "RoundingMode"}, Results: []string{"Decimal128BID", "ExceptionFlags"},
		},
		{
			Symbol: "Sqrt64DBIDWithMode", Kind: "func", Name: "Sqrt64DBIDWithMode",
			Params: []string{"Decimal64BID", "RoundingMode"}, Results: []string{"Decimal64BID", "ExceptionFlags"},
		},
		{
			Symbol: "Sqrt128QBIDWithMode", Kind: "func", Name: "Sqrt128QBIDWithMode",
			Params: []string{"Decimal128BID", "RoundingMode"}, Results: []string{"Decimal128BID", "ExceptionFlags"},
		},
	}
	for _, sym := range tests {
		t.Run(sym.Name, func(t *testing.T) {
			_, err := resolveParityUnit(sym, "unused", nil, publicParityCorpus{})
			if err == nil || !strings.Contains(err.Error(), "unsupported Intel mixed") {
				t.Fatalf("unsupported mixed surface error = %v", err)
			}
		})
	}
}

func TestEmitMixedFMAAndSqrtParityPinsDiscriminantsAndInvalidMode(t *testing.T) {
	tests := []struct {
		name               string
		unit               parityUnit
		discriminant       string
		expectedBoundCalls []generatedGoBoundCallExpectation
		canonicalCall      string
	}{
		{
			name: "FMA64DQQ",
			unit: parityUnit{
				Symbol: "FMA64DQQBIDWithMode", FuncName: "publicParity_FMA64DQQBIDWithMode",
				Shape: shapeFuncMixedModeTernary, Width: 64, Func: "FMA64DQQBIDWithMode", Operation: "FMA",
				TernaryOperandWidths: [3]int{64, 128, 128}, ResultClass: "dec64",
				Port: parityPortPlan{GoName: "Bid64dqqFma", HasRounding: true, FlagsKind: "result", PrimaryResult: "uint64"},
			},
			discriminant:  "discTriples",
			canonicalCall: "canonicalQNaN64BID",
			expectedBoundCalls: []generatedGoBoundCallExpectation{
				{binding: "pv, pf", callee: "FMA64DQQBIDWithMode", args: [][]string{
					{"a", "bv", "c", "mode.pub"},
					{"a", "bv", "c", "mode.pub"},
				}},
				{binding: "pr, prf", callee: "bidgo.Bid64dqqFma", args: [][]string{
					{"aBits", "publicParityToBidgo128(bBits)", "publicParityToBidgo128(cBits)", "mode.port"},
					{"triple.a", "publicParityToBidgo128(triple.b)", "publicParityToBidgo128(triple.c)", "mode.port"},
				}},
				{binding: "invalidValue, invalidFlags", callee: "FMA64DQQBIDWithMode", args: [][]string{
					{"invalidA", "invalidB", "invalidC", "RoundingMode(99)"},
				}},
			},
		},
		{
			name: "Sqrt128D",
			unit: parityUnit{
				Symbol: "Sqrt128DBIDWithMode", FuncName: "publicParity_Sqrt128DBIDWithMode",
				Shape: shapeFuncMixedModeUnary, Width: 128, Func: "Sqrt128DBIDWithMode", Operation: "Sqrt",
				UnaryOperandWidth: 64, ResultClass: "dec128",
				Port: parityPortPlan{GoName: "Bid128dSqrt", HasRounding: true, FlagsKind: "result", PrimaryResult: "BID_UINT128"},
			},
			discriminant:  "discVals",
			canonicalCall: "canonicalQNaN128BID",
			expectedBoundCalls: []generatedGoBoundCallExpectation{
				{binding: "pv, pf", callee: "Sqrt128DBIDWithMode", args: [][]string{
					{"operand", "mode.pub"},
					{"operand", "mode.pub"},
				}},
				{binding: "pr, prf", callee: "bidgo.Bid128dSqrt", args: [][]string{
					{"elem", "mode.port"},
					{"dv", "mode.port"},
				}},
				{binding: "invalidValue, invalidFlags", callee: "Sqrt128DBIDWithMode", args: [][]string{
					{"invalidOperand", "RoundingMode(99)"},
				}},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var b strings.Builder
			if err := emitParityUnitFunc(&b, tc.unit); err != nil {
				t.Fatalf("emit parity unit: %v", err)
			}
			structure := parseGeneratedGoStructure(t, b.String())
			if structure.identifiers[tc.discriminant] == 0 {
				t.Errorf("emitted parity body has no AST identifier %q", tc.discriminant)
			}
			for _, want := range tc.expectedBoundCalls {
				assertGeneratedGoBoundCallArgs(t, structure, want)
			}
			assertGeneratedGoCallArgs(t, structure.calls[tc.canonicalCall], tc.canonicalCall, [][]string{{}})
		})
	}
}

type generatedGoBoundCallExpectation struct {
	binding string
	callee  string
	args    [][]string
}

type generatedGoStructure struct {
	calls       map[string][][]string
	boundCalls  map[string]map[string][][]string
	identifiers map[string]int
}

func parseGeneratedGoStructure(t *testing.T, source string) generatedGoStructure {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "generated_parity.go", "package bid754\n"+source, 0)
	if err != nil {
		t.Fatalf("parse generated Go parity body: %v\n%s", err, source)
	}
	structure := generatedGoStructure{
		calls:       map[string][][]string{},
		boundCalls:  map[string]map[string][][]string{},
		identifiers: map[string]int{},
	}
	ast.Inspect(file, func(node ast.Node) bool {
		switch node := node.(type) {
		case *ast.Ident:
			structure.identifiers[node.Name]++
		case *ast.CallExpr:
			callee := formatGeneratedGoExpr(t, fset, node.Fun)
			args := make([]string, len(node.Args))
			for i, arg := range node.Args {
				args[i] = formatGeneratedGoExpr(t, fset, arg)
			}
			structure.calls[callee] = append(structure.calls[callee], args)
		case *ast.AssignStmt:
			if len(node.Rhs) != 1 {
				break
			}
			call, ok := node.Rhs[0].(*ast.CallExpr)
			if !ok {
				break
			}
			lhs := make([]string, len(node.Lhs))
			for i, expr := range node.Lhs {
				lhs[i] = formatGeneratedGoExpr(t, fset, expr)
			}
			binding := strings.Join(lhs, ", ")
			callee := formatGeneratedGoExpr(t, fset, call.Fun)
			args := make([]string, len(call.Args))
			for i, arg := range call.Args {
				args[i] = formatGeneratedGoExpr(t, fset, arg)
			}
			if structure.boundCalls[binding] == nil {
				structure.boundCalls[binding] = map[string][][]string{}
			}
			structure.boundCalls[binding][callee] = append(structure.boundCalls[binding][callee], args)
		}
		return true
	})
	return structure
}

func assertGeneratedGoBoundCallArgs(t *testing.T, structure generatedGoStructure, want generatedGoBoundCallExpectation) {
	t.Helper()
	byCallee := structure.boundCalls[want.binding]
	total := 0
	for _, calls := range byCallee {
		total += len(calls)
	}
	if total != len(want.args) {
		t.Fatalf("generated binding %q call count = %d, want %d; calls %v", want.binding, total, len(want.args), byCallee)
	}
	assertGeneratedGoCallArgs(t, byCallee[want.callee], want.binding+" = "+want.callee, want.args)
}

func formatGeneratedGoExpr(t *testing.T, fset *token.FileSet, expr ast.Expr) string {
	t.Helper()
	var b bytes.Buffer
	if err := format.Node(&b, fset, expr); err != nil {
		t.Fatalf("format generated Go expression: %v", err)
	}
	return b.String()
}

func assertGeneratedGoCallArgs(t *testing.T, got [][]string, callee string, want [][]string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("generated call %s count = %d, want %d; got args %v", callee, len(got), len(want), got)
	}
	for i := range want {
		if len(got[i]) != len(want[i]) {
			t.Errorf("generated call %s[%d] arg count = %d, want %d; got %v", callee, i, len(got[i]), len(want[i]), got[i])
			continue
		}
		for j := range want[i] {
			if got[i][j] != want[i][j] {
				t.Errorf("generated call %s[%d] arg %d = %q, want %q", callee, i, j, got[i][j], want[i][j])
			}
		}
	}
}
