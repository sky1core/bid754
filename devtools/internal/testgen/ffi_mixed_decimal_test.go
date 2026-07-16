package testgen

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
)

type expectedFFIMixedDecimalShape struct {
	format      string
	operation   string
	resultBits  int
	operandBits []int
	goPort      string
}

var expectedFFIMixedDecimalShapes = map[string]expectedFFIMixedDecimalShape{
	"bid64ddq_fma":  {format: "decimal64", operation: "fma", resultBits: 64, operandBits: []int{64, 64, 128}, goPort: "Bid64ddqFma"},
	"bid64dqd_fma":  {format: "decimal64", operation: "fma", resultBits: 64, operandBits: []int{64, 128, 64}, goPort: "Bid64dqdFma"},
	"bid64dqq_fma":  {format: "decimal64", operation: "fma", resultBits: 64, operandBits: []int{64, 128, 128}, goPort: "Bid64dqqFma"},
	"bid64qdd_fma":  {format: "decimal64", operation: "fma", resultBits: 64, operandBits: []int{128, 64, 64}, goPort: "Bid64qddFma"},
	"bid64qdq_fma":  {format: "decimal64", operation: "fma", resultBits: 64, operandBits: []int{128, 64, 128}, goPort: "Bid64qdqFma"},
	"bid64qqd_fma":  {format: "decimal64", operation: "fma", resultBits: 64, operandBits: []int{128, 128, 64}, goPort: "Bid64qqdFma"},
	"bid64qqq_fma":  {format: "decimal64", operation: "fma", resultBits: 64, operandBits: []int{128, 128, 128}, goPort: "Bid64qqqFma"},
	"bid128ddd_fma": {format: "decimal128", operation: "fma", resultBits: 128, operandBits: []int{64, 64, 64}, goPort: "Bid128dddFma"},
	"bid128ddq_fma": {format: "decimal128", operation: "fma", resultBits: 128, operandBits: []int{64, 64, 128}, goPort: "Bid128ddqFma"},
	"bid128dqd_fma": {format: "decimal128", operation: "fma", resultBits: 128, operandBits: []int{64, 128, 64}, goPort: "Bid128dqdFma"},
	"bid128dqq_fma": {format: "decimal128", operation: "fma", resultBits: 128, operandBits: []int{64, 128, 128}, goPort: "Bid128dqqFma"},
	"bid128qdd_fma": {format: "decimal128", operation: "fma", resultBits: 128, operandBits: []int{128, 64, 64}, goPort: "Bid128qddFma"},
	"bid128qdq_fma": {format: "decimal128", operation: "fma", resultBits: 128, operandBits: []int{128, 64, 128}, goPort: "Bid128qdqFma"},
	"bid128qqd_fma": {format: "decimal128", operation: "fma", resultBits: 128, operandBits: []int{128, 128, 64}, goPort: "Bid128qqdFma"},
	"bid64q_sqrt":   {format: "decimal64", operation: "sqrt", resultBits: 64, operandBits: []int{128}, goPort: "Bid64qSqrt"},
	"bid128d_sqrt":  {format: "decimal128", operation: "sqrt", resultBits: 128, operandBits: []int{64}, goPort: "Bid128dSqrt"},
}

var expectedFFIMixedDecimalFunctionOrder = []string{
	"bid64ddq_fma",
	"bid64dqd_fma",
	"bid64dqq_fma",
	"bid64qdd_fma",
	"bid64qdq_fma",
	"bid64qqd_fma",
	"bid64qqq_fma",
	"bid128ddd_fma",
	"bid128ddq_fma",
	"bid128dqd_fma",
	"bid128dqq_fma",
	"bid128qdd_fma",
	"bid128qdq_fma",
	"bid128qqd_fma",
	"bid64q_sqrt",
	"bid128d_sqrt",
}

func TestBuildFFICasesSupportsClosedWorldMixedDecimalShapes(t *testing.T) {
	const casesPerFunction = 48
	const roundingProbeCasesPerFunction = ffiRoundingModeCount
	repoRoot := filepath.Clean(filepath.Join("..", ".."))
	functions := make([]string, 0, len(expectedFFIMixedDecimalShapes))
	for function := range expectedFFIMixedDecimalShapes {
		functions = append(functions, function)
	}

	cases, err := buildFFICases(repoRoot, FFITestSpec{
		Name:             "mixed_decimal_shapes",
		Symbols:          "generated/json/intel_dfp_symbols.json",
		Functions:        functions,
		CasesPerFunction: casesPerFunction,
		Seed:             754,
	})
	if err != nil {
		t.Fatalf("buildFFICases mixed decimal shapes: %v", err)
	}

	counts := make(map[string]int, len(functions))
	for _, tc := range cases {
		shape, ok := expectedFFIMixedDecimalShapes[tc.Function]
		if !ok {
			t.Fatalf("unexpected mixed decimal function %q", tc.Function)
		}
		if tc.Format != shape.format || tc.Operation != shape.operation {
			t.Fatalf("%s classification = (%s, %s), want (%s, %s)", tc.Function, tc.Format, tc.Operation, shape.format, shape.operation)
		}
		if tc.ResultBits != shape.resultBits || !slices.Equal(tc.OperandBits, shape.operandBits) {
			t.Fatalf("%s generated shape = (%d, %v), want (%d, %v)", tc.Function, tc.ResultBits, tc.OperandBits, shape.resultBits, shape.operandBits)
		}
		if len(tc.Operands) != len(shape.operandBits) {
			t.Fatalf("%s operand count = %d, want %d", tc.Function, len(tc.Operands), len(shape.operandBits))
		}
		for i, bits := range shape.operandBits {
			if got, want := len(tc.Operands[i]), bits/4; got != want {
				t.Fatalf("%s operand %d hex width = %d, want %d (%d bits): %q", tc.Function, i, got, want, bits, tc.Operands[i])
			}
		}
		counts[tc.Function]++
	}
	fmaFunctions := 0
	for _, shape := range expectedFFIMixedDecimalShapes {
		if shape.operation == "fma" {
			fmaFunctions++
		}
	}
	if got, want := len(cases), len(expectedFFIMixedDecimalShapes)*(casesPerFunction+roundingProbeCasesPerFunction)+fmaFunctions; got != want {
		t.Fatalf("mixed decimal FFI case count = %d, want %d", got, want)
	}
	for function, shape := range expectedFFIMixedDecimalShapes {
		want := casesPerFunction + roundingProbeCasesPerFunction
		if shape.operation == "fma" {
			want++
		}
		if counts[function] != want {
			t.Fatalf("%s case count = %d, want %d", function, counts[function], want)
		}
	}
}

func TestBuildFFICasesRepeatsMixedDecimalRoundingProbesAcrossAllModes(t *testing.T) {
	const casesPerFunction = 48
	repoRoot := filepath.Clean(filepath.Join("..", ".."))
	functions := make([]string, 0, len(expectedFFIMixedDecimalShapes))
	for function := range expectedFFIMixedDecimalShapes {
		functions = append(functions, function)
	}

	cases, err := buildFFICases(repoRoot, FFITestSpec{
		Name:             "mixed_decimal_rounding_probes",
		Symbols:          "generated/json/intel_dfp_symbols.json",
		Functions:        functions,
		CasesPerFunction: casesPerFunction,
		Seed:             754,
	})
	if err != nil {
		t.Fatalf("buildFFICases mixed decimal rounding probes: %v", err)
	}

	type probeGroup struct {
		function string
		operands []string
		modes    map[int]int
	}
	groups := map[string]*probeGroup{}
	for _, tc := range cases {
		if tc.Probe == "" {
			if tc.ProbeGroup != "" {
				t.Fatalf("case %s has probe_group %q without a probe", tc.ID, tc.ProbeGroup)
			}
			continue
		}
		if tc.Probe == ffiProbeFusedness {
			continue
		}
		if tc.Probe != ffiProbeRoundingDiscriminant {
			t.Fatalf("case %s has unexpected probe %q", tc.ID, tc.Probe)
		}
		if tc.ProbeGroup == "" {
			t.Fatalf("case %s has no probe_group", tc.ID)
		}
		group := groups[tc.ProbeGroup]
		if group == nil {
			group = &probeGroup{
				function: tc.Function,
				operands: append([]string(nil), tc.Operands...),
				modes:    map[int]int{},
			}
			groups[tc.ProbeGroup] = group
		} else if group.function != tc.Function || !slices.Equal(group.operands, tc.Operands) {
			t.Fatalf("probe group %q mixes function/operands: first=(%s,%v), case %s=(%s,%v)", tc.ProbeGroup, group.function, group.operands, tc.ID, tc.Function, tc.Operands)
		}
		group.modes[tc.Rounding]++
	}

	groupsPerFunction := map[string]int{}
	for groupName, group := range groups {
		for mode := 0; mode < ffiRoundingModeCount; mode++ {
			if group.modes[mode] != 1 {
				t.Errorf("probe group %q mode %d count = %d, want 1", groupName, mode, group.modes[mode])
			}
		}
		groupsPerFunction[group.function]++
	}
	for function := range expectedFFIMixedDecimalShapes {
		if got, want := groupsPerFunction[function], 1; got != want {
			t.Errorf("%s rounding probe groups = %d, want %d", function, got, want)
		}
	}
}

func TestGeneratedSpecHasClosedMixedDecimalProbeCensus(t *testing.T) {
	repoRoot := filepath.Clean(filepath.Join("..", ".."))
	spec, err := LoadGenerated(filepath.Join(repoRoot, "generated", "testspec", "spec_index.json"))
	if err != nil {
		t.Fatalf("LoadGenerated: %v", err)
	}

	actualFunctions := map[string]expectedFFIMixedDecimalShape{}
	roundingGroups := map[string]map[string][]GeneratedFFICase{}
	fusednessGroups := map[string]map[string][]GeneratedFFICase{}
	for _, tc := range spec.FFICases {
		if len(tc.OperandBits) == 0 {
			continue
		}
		shape := expectedFFIMixedDecimalShape{
			format:      tc.Format,
			operation:   tc.Operation,
			resultBits:  tc.ResultBits,
			operandBits: append([]int(nil), tc.OperandBits...),
		}
		if prior, ok := actualFunctions[tc.Function]; ok {
			if prior.format != shape.format || prior.operation != shape.operation || prior.resultBits != shape.resultBits || !slices.Equal(prior.operandBits, shape.operandBits) {
				t.Fatalf("generated mixed function %s changes shape from (%s,%s,%d,%v) to (%s,%s,%d,%v)", tc.Function, prior.format, prior.operation, prior.resultBits, prior.operandBits, shape.format, shape.operation, shape.resultBits, shape.operandBits)
			}
		} else {
			actualFunctions[tc.Function] = shape
		}
		switch tc.Probe {
		case "":
		case ffiProbeRoundingDiscriminant:
			if roundingGroups[tc.Function] == nil {
				roundingGroups[tc.Function] = map[string][]GeneratedFFICase{}
			}
			roundingGroups[tc.Function][tc.ProbeGroup] = append(roundingGroups[tc.Function][tc.ProbeGroup], tc)
		case ffiProbeFusedness:
			if fusednessGroups[tc.Function] == nil {
				fusednessGroups[tc.Function] = map[string][]GeneratedFFICase{}
			}
			fusednessGroups[tc.Function][tc.ProbeGroup] = append(fusednessGroups[tc.Function][tc.ProbeGroup], tc)
		default:
			t.Fatalf("generated mixed function %s has unknown probe %q", tc.Function, tc.Probe)
		}
	}

	if len(actualFunctions) != len(expectedFFIMixedDecimalShapes) {
		t.Errorf("generated mixed function census = %d, want %d", len(actualFunctions), len(expectedFFIMixedDecimalShapes))
	}
	for function, want := range expectedFFIMixedDecimalShapes {
		got, ok := actualFunctions[function]
		if !ok {
			t.Errorf("generated mixed function census is missing %s", function)
			continue
		}
		if got.format != want.format || got.operation != want.operation || got.resultBits != want.resultBits || !slices.Equal(got.operandBits, want.operandBits) {
			t.Errorf("generated mixed function %s shape = (%s,%s,%d,%v), want (%s,%s,%d,%v)", function, got.format, got.operation, got.resultBits, got.operandBits, want.format, want.operation, want.resultBits, want.operandBits)
		}
	}
	for function := range actualFunctions {
		if _, ok := expectedFFIMixedDecimalShapes[function]; !ok {
			t.Errorf("generated mixed function census has unexpected %s", function)
		}
	}

	for function, shape := range expectedFFIMixedDecimalShapes {
		groups := roundingGroups[function]
		if len(groups) != 1 {
			t.Errorf("%s rounding probe groups = %d, want 1", function, len(groups))
		} else {
			for groupName, cases := range groups {
				if len(cases) != ffiRoundingModeCount {
					t.Errorf("%s rounding probe group %q cases = %d, want %d", function, groupName, len(cases), ffiRoundingModeCount)
				}
				modes := [ffiRoundingModeCount]int{}
				var operands []string
				for i, tc := range cases {
					if tc.Rounding < 0 || tc.Rounding >= ffiRoundingModeCount {
						t.Errorf("%s rounding probe group %q mode = %d, want 0..4", function, groupName, tc.Rounding)
						continue
					}
					modes[tc.Rounding]++
					if i == 0 {
						operands = tc.Operands
					} else if !slices.Equal(operands, tc.Operands) {
						t.Errorf("%s rounding probe group %q mixes operands %v and %v", function, groupName, operands, tc.Operands)
					}
				}
				for mode, count := range modes {
					if count != 1 {
						t.Errorf("%s rounding probe group %q mode %d count = %d, want 1", function, groupName, mode, count)
					}
				}
			}
		}

		fusedGroups := fusednessGroups[function]
		if shape.operation == "fma" {
			if len(fusedGroups) != 1 {
				t.Errorf("%s fusedness probe groups = %d, want 1", function, len(fusedGroups))
			} else {
				for groupName, cases := range fusedGroups {
					if groupName != function+"/fusedness" || len(cases) != 1 {
						t.Errorf("%s fusedness group = %q with %d cases, want %q with 1", function, groupName, len(cases), function+"/fusedness")
					}
				}
			}
		} else if len(fusedGroups) != 0 {
			t.Errorf("%s non-FMA fusedness probe groups = %d, want 0", function, len(fusedGroups))
		}
	}
}

func TestVerifyFFISignatureRejectsMixedDecimalOperandWidthMismatch(t *testing.T) {
	repoRoot := filepath.Clean(filepath.Join("..", ".."))
	symbols, err := loadSymbolFile(filepath.Join(repoRoot, "generated", "json", "intel_dfp_symbols.json"))
	if err != nil {
		t.Fatalf("load Intel symbols: %v", err)
	}
	var symbol symbolSpec
	for _, candidate := range symbols.Symbols {
		if candidate.Name == "bid64ddq_fma" {
			symbol = candidate
			break
		}
	}
	if symbol.Name == "" {
		t.Fatal("missing bid64ddq_fma Intel symbol")
	}
	symbol.Parameters = append([]string(nil), symbol.Parameters...)
	symbol.Parameters[2] = "BID_UINT64 z"

	_, _, _, _, err = verifyFFISignature(symbol.Name, symbol)
	if err == nil || !strings.Contains(err.Error(), "parameter 2") {
		t.Fatalf("mixed operand width mismatch error = %v, want parameter 2 rejection", err)
	}
}

func TestMixedDecimalFFIShapeSurvivesShardRoundTrip(t *testing.T) {
	want := GeneratedFFICase{
		Suite:       "mixed_decimal_shapes",
		ID:          "mixed_decimal_shapes_bid64ddq_fma_001",
		Format:      "decimal64",
		ResultBits:  64,
		OperandBits: []int{64, 64, 128},
		Operation:   "fma",
		Function:    "bid64ddq_fma",
		LinkName:    "__bid64ddq_fma",
		Declaration: "BID_UINT64 bid64ddq_fma(BID_UINT64 x, BID_UINT64 y, BID_UINT128 z, _IDEC_round rnd_mode, _IDEC_flags*pfpsf)",
		Source:      "generated/json/intel_dfp_symbols.json",
		Probe:       ffiProbeFusedness,
		ProbeGroup:  "bid64ddq_fma/fusedness",
		Expected:    "2fe38d7ea4c68001/00000020",
		Forbidden:   "2fe38d7ea4c68000/00000020",
		Rounding:    0,
		Operands:    []string{"31c0000000000001", "31c0000000000001", "30400000000000000000000000000001"},
	}
	manifest := Manifest{Output: "generated/testspec/spec_index.json"}
	files, err := EncodeSpecFiles(manifest, SharedSpec{FFICases: []GeneratedFFICase{want}})
	if err != nil {
		t.Fatalf("EncodeSpecFiles: %v", err)
	}

	root := t.TempDir()
	for relativePath, data := range files {
		path := filepath.Join(root, filepath.FromSlash(relativePath))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", path, err)
		}
		if err := os.WriteFile(path, data, 0o644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}
	gotSpec, err := LoadGenerated(filepath.Join(root, filepath.FromSlash(manifest.Output)))
	if err != nil {
		t.Fatalf("LoadGenerated: %v", err)
	}
	if len(gotSpec.FFICases) != 1 {
		t.Fatalf("round-trip FFI case count = %d, want 1", len(gotSpec.FFICases))
	}
	got := gotSpec.FFICases[0]
	if got.ResultBits != want.ResultBits || !slices.Equal(got.OperandBits, want.OperandBits) {
		t.Fatalf("round-trip shape = (%d, %v), want (%d, %v)", got.ResultBits, got.OperandBits, want.ResultBits, want.OperandBits)
	}
	if got.Probe != want.Probe || got.ProbeGroup != want.ProbeGroup {
		t.Fatalf("round-trip probe = (%q, %q), want (%q, %q)", got.Probe, got.ProbeGroup, want.Probe, want.ProbeGroup)
	}
	if got.Expected != want.Expected || got.Forbidden != want.Forbidden {
		t.Fatalf("round-trip outcomes = (%q, %q), want (%q, %q)", got.Expected, got.Forbidden, want.Expected, want.Forbidden)
	}
}

func TestBuildFFICasesAddsClosedMixedFMAFusednessProbes(t *testing.T) {
	repoRoot := filepath.Clean(filepath.Join("..", ".."))
	functions := make([]string, 0, len(expectedFFIMixedDecimalShapes))
	for function := range expectedFFIMixedDecimalShapes {
		functions = append(functions, function)
	}
	cases, err := buildFFICases(repoRoot, FFITestSpec{
		Name: "mixed_fma_fusedness", Symbols: "generated/json/intel_dfp_symbols.json",
		Functions: functions, CasesPerFunction: 1, Seed: 754,
	})
	if err != nil {
		t.Fatalf("buildFFICases mixed FMA fusedness: %v", err)
	}
	want := map[string]ffiFusednessProbe{}
	for _, probe := range ffiMixedFMAFusednessProbes {
		shape, ok := ffiMixedDecimalShapeFor(probe.function)
		if !ok {
			t.Fatalf("canonical fusedness function %q has no mixed shape", probe.function)
		}
		if err := validateFFIFusednessProbe(probe, shape); err != nil {
			t.Fatalf("canonical fusedness probe %s: %v", probe.function, err)
		}
		if _, duplicate := want[probe.function]; duplicate {
			t.Fatalf("duplicate canonical fusedness function %q", probe.function)
		}
		want[probe.function] = probe
	}
	if got := len(want); got != 14 {
		t.Fatalf("canonical fusedness census = %d, want 14", got)
	}
	seen := map[string]bool{}
	for _, tc := range cases {
		if tc.Probe != ffiProbeFusedness {
			continue
		}
		probe, ok := want[tc.Function]
		if !ok {
			t.Fatalf("unexpected fusedness probe function %q", tc.Function)
		}
		if seen[tc.Function] {
			t.Fatalf("duplicate generated fusedness probe %q", tc.Function)
		}
		seen[tc.Function] = true
		if tc.ProbeGroup != tc.Function+"/fusedness" || tc.Rounding != probe.rounding ||
			!slices.Equal(tc.Operands, probe.ffiOperands()) || tc.Expected != probe.expected.ffiString() || tc.Forbidden != probe.forbidden.ffiString() {
			t.Fatalf("%s fusedness case differs from canonical probe: %#v", tc.Function, tc)
		}
	}
	for function := range want {
		if !seen[function] {
			t.Errorf("missing fusedness probe for %s", function)
		}
	}
}

func TestGenerateFFINativeRunnerCallsMixedDecimalSymbolsAndPortsDirectly(t *testing.T) {
	outputs, err := GenerateFFITestOutputs(SharedSpec{})
	if err != nil {
		t.Fatalf("GenerateFFITestOutputs: %v", err)
	}
	fset, file := parseGeneratedGoOutput(t, ffiGeneratedNativeSupportPath, outputs[ffiGeneratedNativeSupportPath])
	caller := requireGeneratedFunc(t, file, "runGeneratedFFICase")
	if err := validateGeneratedMixedDecimalCallerRoute(fset, caller); err != nil {
		t.Fatal(err)
	}
	shapeResolver := requireGeneratedFunc(t, file, "generatedFFIMixedDecimalShapeFor")
	if err := validateGeneratedMixedDecimalShapeBody(fset, shapeResolver); err != nil {
		t.Fatal(err)
	}
	predicate := requireGeneratedFunc(t, file, "generatedFFIMixedDecimalFunction")
	if err := validateGeneratedMixedDecimalPredicateBody(fset, predicate); err != nil {
		t.Fatal(err)
	}
	parserFn := requireGeneratedFunc(t, file, "parseGeneratedFFIMixedDecimalOperands")
	if err := validateGeneratedMixedDecimalParserBody(fset, parserFn); err != nil {
		t.Fatal(err)
	}
	equalWidths := requireGeneratedFunc(t, file, "equalGeneratedFFIIntSlices")
	if err := validateGeneratedMixedDecimalWidthEqualityBody(fset, equalWidths); err != nil {
		t.Fatal(err)
	}
	runner := requireGeneratedFunc(t, file, "runGeneratedFFICaseMixedDecimal")
	if err := validateGeneratedMixedDecimalRunnerBody(fset, runner); err != nil {
		t.Fatal(err)
	}

	mutationFSet, mutationFile := parseGeneratedGoOutput(t, ffiGeneratedNativeSupportPath, outputs[ffiGeneratedNativeSupportPath])
	mutatedCaller := requireGeneratedFunc(t, mutationFile, "runGeneratedFFICase")
	mixedRoute := mutatedCaller.Body.List[0].(*ast.IfStmt)
	mixedRoute.Body.List = []ast.Stmt{
		&ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent("native"), ast.NewIdent("_"), ast.NewIdent("err")},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{&ast.CallExpr{
				Fun:  ast.NewIdent("runGeneratedFFICaseMixedDecimal"),
				Args: []ast.Expr{ast.NewIdent("tc")},
			}},
		},
		&ast.ReturnStmt{Results: []ast.Expr{ast.NewIdent("native"), ast.NewIdent("native"), ast.NewIdent("err")}},
	}
	if err := validateGeneratedMixedDecimalCallerRoute(mutationFSet, mutatedCaller); err == nil {
		t.Fatal("mixed decimal caller contract accepted a live runner call whose Go-port result was discarded for a C/C projection")
	}

	supportSource := string(outputs[ffiGeneratedNativeSupportPath])
	const parserHeader = "func parseGeneratedFFIMixedDecimalOperands(tc generatedFFICase) (generatedFFIMixedDecimalOperands, error) {\n"
	baselineZeroSource := strings.Replace(supportSource, parserHeader, parserHeader+"\tif tc.Probe == \"\" {\n\t\treturn generatedFFIMixedDecimalOperands{}, nil\n\t}\n", 1)
	if baselineZeroSource == supportSource {
		t.Fatal("mixed decimal parser mutation fixture did not find the parser declaration")
	}
	baselineZeroFSet, baselineZeroFile := parseGeneratedGoOutput(t, ffiGeneratedNativeSupportPath, []byte(baselineZeroSource))
	if err := validateGeneratedMixedDecimalParserBody(baselineZeroFSet, requireGeneratedFunc(t, baselineZeroFile, "parseGeneratedFFIMixedDecimalOperands")); err == nil {
		t.Fatal("mixed decimal parser contract accepted baseline cases rewritten to zero operands")
	}

	const slotAssignment = "\t\t\toperands.narrow[i] = value"
	slotRewriteSource := strings.Replace(supportSource, slotAssignment, "\t\t\toperands.narrow[0] = value", 1)
	if slotRewriteSource == supportSource {
		t.Fatal("mixed decimal parser mutation fixture did not find the width-indexed slot assignment")
	}
	slotRewriteFSet, slotRewriteFile := parseGeneratedGoOutput(t, ffiGeneratedNativeSupportPath, []byte(slotRewriteSource))
	if err := validateGeneratedMixedDecimalParserBody(slotRewriteFSet, requireGeneratedFunc(t, slotRewriteFile, "parseGeneratedFFIMixedDecimalOperands")); err == nil {
		t.Fatal("mixed decimal parser contract accepted a fixed operand slot in place of the live width-indexed slot")
	}

	const shapeRow = `return generatedFFIMixedDecimalShape{"decimal64", "fma", 64, []int{64, 64, 128}}, true`
	shapeRewriteSource := strings.Replace(supportSource, shapeRow, `return generatedFFIMixedDecimalShape{"decimal64", "fma", 64, []int{64, 128, 64}}, true`, 1)
	if shapeRewriteSource == supportSource {
		t.Fatal("mixed decimal shape mutation fixture did not find bid64ddq_fma")
	}
	shapeRewriteFSet, shapeRewriteFile := parseGeneratedGoOutput(t, ffiGeneratedNativeSupportPath, []byte(shapeRewriteSource))
	if err := validateGeneratedMixedDecimalShapeBody(shapeRewriteFSet, requireGeneratedFunc(t, shapeRewriteFile, "generatedFFIMixedDecimalShapeFor")); err == nil {
		t.Fatal("mixed decimal shape contract accepted a rewritten operand-width slot")
	}

	mutated := requireGeneratedFunc(t, mutationFile, "runGeneratedFFICaseMixedDecimal")
	switchStmt := mutated.Body.List[4].(*ast.SwitchStmt)
	firstArm := switchStmt.Body.List[0].(*ast.CaseClause)
	goPortAssign := firstArm.Body[1].(*ast.AssignStmt)
	deadGoPortDecoy := &ast.IfStmt{
		Cond: ast.NewIdent("false"),
		Body: &ast.BlockStmt{List: []ast.Stmt{goPortAssign}},
	}
	liveCProjection := &ast.AssignStmt{
		Lhs: goPortAssign.Lhs,
		Tok: token.DEFINE,
		Rhs: []ast.Expr{
			ast.NewIdent("native"),
			&ast.CallExpr{Fun: ast.NewIdent("uint32"), Args: []ast.Expr{ast.NewIdent("flags")}},
		},
	}
	firstArm.Body = []ast.Stmt{firstArm.Body[0], deadGoPortDecoy, liveCProjection, firstArm.Body[2]}
	if err := validateGeneratedMixedDecimalRunnerBody(mutationFSet, mutated); err == nil {
		t.Fatal("mixed decimal runner contract accepted a dead Go-port decoy with live C/C projection")
	}
}

func TestGenerateFFINativeRunnerRequiresCanonicalRoundingDiscrimination(t *testing.T) {
	inputSpec := SharedSpec{}
	outputs, err := GenerateFFITestOutputs(inputSpec)
	if err != nil {
		t.Fatalf("GenerateFFITestOutputs: %v", err)
	}
	fset, file := parseGeneratedGoOutput(t, ffiGeneratedNativeTestPath, outputs[ffiGeneratedNativeTestPath])
	validator := requireGeneratedFunc(t, file, "validateGeneratedFFIProbeContract")
	if got, want := generatedFuncResultTypes(t, fset, validator), []string{"generatedFFIRoundingProbeTracker", "error"}; !slices.Equal(got, want) {
		t.Fatalf("validateGeneratedFFIProbeContract results = %v, want %v", got, want)
	}
	if got := len(generatedCallsNamed(t, fset, validator, "fmt.Errorf")); got == 0 {
		t.Fatal("validateGeneratedFFIProbeContract has no error-return construction")
	}
	for _, forbidden := range []string{"t.Fatal", "t.Fatalf", "t.Error", "t.Errorf"} {
		if got := len(generatedCallsNamed(t, fset, validator, forbidden)); got != 0 {
			t.Errorf("validateGeneratedFFIProbeContract calls %s %d times, want 0", forbidden, got)
		}
	}

	recorder := requireGeneratedMethod(t, file, "generatedFFIRoundingProbeTracker", "recordCanonicalResult")
	if got := len(generatedCallsNamed(t, fset, recorder, "hex.DecodeString")); got != 2 {
		t.Errorf("recordCanonicalResult hex.DecodeString calls = %d, want 2", got)
	}
	discriminator := requireGeneratedMethod(t, file, "generatedFFIRoundingProbeTracker", "validateCanonicalDiscrimination")
	if !generatedFuncHasBinaryExpr(t, fset, discriminator, "len(group.nativeResultBits)", token.LSS, "2") {
		t.Error("validateCanonicalDiscrimination lacks the canonical distinct-result lower bound")
	}

	mutationTest := requireGeneratedFunc(t, file, "testGeneratedFFIProbeValidatorRejectsMutations")
	for _, name := range []string{
		"unknown probe", "orphan group", "empty group", "non-mixed target", "mode out of range",
		"duplicate mode", "missing mode", "mixed function", "mixed operands", "required function probe missing",
		"fused expected drift", "fused forbidden drift", "fused expected missing", "fused forbidden missing",
		"fusedness probe missing", "malformed native result", "canonical result does not discriminate",
	} {
		if !generatedFuncHasStringLiteral(mutationTest, name) {
			t.Errorf("generated validator mutation test is missing case %q", name)
		}
	}
	if got := len(generatedCallsNamed(t, fset, mutationTest, "validateGeneratedFFIProbeContract")); got < 3 {
		t.Errorf("validator mutation helper validation calls = %d, want at least 3", got)
	}
	if got := len(generatedCallsNamed(t, fset, mutationTest, "tracker.recordCanonicalResult")); got < 2 {
		t.Errorf("validator mutation helper canonical-record calls = %d, want at least 2", got)
	}
	if got := len(generatedCallsNamed(t, fset, mutationTest, "tracker.validateCanonicalDiscrimination")); got != 1 {
		t.Errorf("validator mutation helper discrimination calls = %d, want 1", got)
	}

	mainTest := requireGeneratedFunc(t, file, "TestGeneratedFFIBitCompareSubset")
	if got := len(generatedCallsNamed(t, fset, mainTest, "testGeneratedFFIProbeValidatorRejectsMutations")); got != 0 {
		t.Errorf("native FFI exact gate calls validator mutation helper %d times, want 0", got)
	}
	mutationTopLevel := requireGeneratedFunc(t, file, "TestGeneratedFFIProbeValidatorRejectsMutations")
	if got := len(generatedCallsNamed(t, fset, mutationTopLevel, "testGeneratedFFIProbeValidatorRejectsMutations")); got != 1 {
		t.Errorf("validator mutation top-level test calls helper %d times, want 1", got)
	}
	if err := validateGeneratedFFIMainLoopBody(fset, mainTest, len(inputSpec.FFICases)); err != nil {
		t.Fatal(err)
	}
	assertGeneratedFFIMainLoopMutationsRejected(t, outputs[ffiGeneratedNativeTestPath], len(inputSpec.FFICases))
}

func TestGenerateFFINativeRunnerRequiresCanonicalMixedFMAFusedness(t *testing.T) {
	outputs, err := GenerateFFITestOutputs(SharedSpec{})
	if err != nil {
		t.Fatalf("GenerateFFITestOutputs: %v", err)
	}
	supportFSet, supportFile := parseGeneratedGoOutput(t, ffiGeneratedNativeSupportPath, outputs[ffiGeneratedNativeSupportPath])
	widen := requireGeneratedFunc(t, supportFile, "generatedFFIWidenMixedDecimalOperand")
	if err := validateGeneratedMixedDecimalWidenBody(supportFSet, widen); err != nil {
		t.Fatal(err)
	}
	widenMutationFSet, widenMutationFile := parseGeneratedGoOutput(t, ffiGeneratedNativeSupportPath, outputs[ffiGeneratedNativeSupportPath])
	mutatedWiden := requireGeneratedFunc(t, widenMutationFile, "generatedFFIWidenMixedDecimalOperand")
	widenSwitch := mutatedWiden.Body.List[0].(*ast.SwitchStmt)
	wideReturn := widenSwitch.Body.List[1].(*ast.CaseClause).Body[0]
	mutatedWiden.Body.List = []ast.Stmt{
		&ast.IfStmt{Cond: ast.NewIdent("false"), Body: &ast.BlockStmt{List: []ast.Stmt{widenSwitch}}},
		wideReturn,
	}
	if err := validateGeneratedMixedDecimalWidenBody(widenMutationFSet, mutatedWiden); err == nil {
		t.Fatal("mixed decimal widen contract accepted a dead BID64 conversion decoy with a live BID128-only return")
	}
	composed := requireGeneratedFunc(t, supportFile, "runGeneratedFFIMixedFMAComposed")
	if err := validateGeneratedMixedFMAComposedBody(supportFSet, composed); err != nil {
		t.Fatal(err)
	}

	mutationFSet, mutationFile := parseGeneratedGoOutput(t, ffiGeneratedNativeSupportPath, outputs[ffiGeneratedNativeSupportPath])
	mutated := requireGeneratedFunc(t, mutationFile, "runGeneratedFFIMixedFMAComposed")
	deadComposedDecoy := &ast.IfStmt{
		Cond: ast.NewIdent("false"),
		Body: &ast.BlockStmt{List: append([]ast.Stmt(nil), mutated.Body.List[7:]...)},
	}
	forbiddenReturn := &ast.ReturnStmt{Results: []ast.Expr{
		&ast.SelectorExpr{X: ast.NewIdent("tc"), Sel: ast.NewIdent("Forbidden")},
		ast.NewIdent("nil"),
	}}
	mutated.Body.List = append(append([]ast.Stmt(nil), mutated.Body.List[:7]...), deadComposedDecoy, forbiddenReturn)
	if err := validateGeneratedMixedFMAComposedBody(mutationFSet, mutated); err == nil {
		t.Fatal("mixed FMA composed contract accepted dead arithmetic decoys with tc.Forbidden as the live result")
	}
}

func validateGeneratedMixedDecimalCallerRoute(fset *token.FileSet, fn *ast.FuncDecl) error {
	if len(fn.Body.List) == 0 {
		return fmt.Errorf("%s has no live routing statements", fn.Name.Name)
	}
	expectedFSet := token.NewFileSet()
	expectedFile, err := parser.ParseFile(expectedFSet, "expected_mixed_caller_route.go", `package expected
func expected(tc generatedFFICase) (string, string, error) {
	if generatedFFIMixedDecimalFunction(tc.Function) {
		return runGeneratedFFICaseMixedDecimal(tc)
	}
}`, parser.AllErrors)
	if err != nil {
		return fmt.Errorf("parse expected mixed decimal caller route: %w", err)
	}
	expectedFn := expectedFile.Decls[0].(*ast.FuncDecl)
	actualText, err := formatGeneratedASTNode(fset, fn.Body.List[0])
	if err != nil {
		return fmt.Errorf("format %s first routing statement: %w", fn.Name.Name, err)
	}
	expectedText, err := formatGeneratedASTNode(expectedFSet, expectedFn.Body.List[0])
	if err != nil {
		return fmt.Errorf("format expected mixed decimal caller route: %w", err)
	}
	if actualText != expectedText {
		return fmt.Errorf("%s first live route = %q, want exact mixed predicate to direct runner route %q", fn.Name.Name, actualText, expectedText)
	}
	return nil
}

func validateGeneratedMixedDecimalShapeBody(fset *token.FileSet, fn *ast.FuncDecl) error {
	var b strings.Builder
	b.WriteString("{\n\tswitch function {\n")
	for _, function := range expectedFFIMixedDecimalFunctionOrder {
		shape, ok := expectedFFIMixedDecimalShapes[function]
		if !ok {
			panic("missing expected mixed decimal shape for " + function)
		}
		fmt.Fprintf(&b, "\tcase %q:\n\t\treturn generatedFFIMixedDecimalShape{%q, %q, %d, %#v}, true\n",
			function, shape.format, shape.operation, shape.resultBits, shape.operandBits)
	}
	b.WriteString(`	default:
		return generatedFFIMixedDecimalShape{}, false
	}
}`)
	return validateGeneratedExactFunctionBody(fset, fn, b.String())
}

func validateGeneratedMixedDecimalPredicateBody(fset *token.FileSet, fn *ast.FuncDecl) error {
	return validateGeneratedExactFunctionBody(fset, fn, `{
	_, ok := generatedFFIMixedDecimalShapeFor(function)
	return ok
}`)
}

func validateGeneratedMixedDecimalParserBody(fset *token.FileSet, fn *ast.FuncDecl) error {
	return validateGeneratedExactFunctionBody(fset, fn, `{
	shape, ok := generatedFFIMixedDecimalShapeFor(tc.Function)
	if !ok {
		return generatedFFIMixedDecimalOperands{}, fmt.Errorf("unsupported mixed decimal ffi function %q", tc.Function)
	}
	if tc.Format != shape.format || tc.Operation != shape.operation || tc.ResultBits != shape.resultBits || !equalGeneratedFFIIntSlices(tc.OperandBits, shape.operandBits) {
		return generatedFFIMixedDecimalOperands{}, fmt.Errorf("%s mixed decimal shape = (%s, %s, %d, %v), want (%s, %s, %d, %v)", tc.Function, tc.Format, tc.Operation, tc.ResultBits, tc.OperandBits, shape.format, shape.operation, shape.resultBits, shape.operandBits)
	}
	if len(tc.Operands) != len(shape.operandBits) {
		return generatedFFIMixedDecimalOperands{}, fmt.Errorf("%s expects %d operands, got %d", tc.Function, len(shape.operandBits), len(tc.Operands))
	}

	var operands generatedFFIMixedDecimalOperands
	for i, bits := range shape.operandBits {
		switch bits {
		case 64:
			value, err := strconv.ParseUint(tc.Operands[i], 16, 64)
			if err != nil {
				return generatedFFIMixedDecimalOperands{}, fmt.Errorf("parse operand %d as BID64 %q: %w", i, tc.Operands[i], err)
			}
			operands.narrow[i] = value
		case 128:
			value, err := parseFFIUint128Bits(tc.Operands[i])
			if err != nil {
				return generatedFFIMixedDecimalOperands{}, fmt.Errorf("parse operand %d as BID128 %q: %w", i, tc.Operands[i], err)
			}
			operands.wide[i] = value
		default:
			return generatedFFIMixedDecimalOperands{}, fmt.Errorf("%s operand %d has unsupported width %d", tc.Function, i, bits)
		}
	}
	return operands, nil
}`)
}

func validateGeneratedMixedDecimalWidthEqualityBody(fset *token.FileSet, fn *ast.FuncDecl) error {
	return validateGeneratedExactFunctionBody(fset, fn, `{
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}`)
}

func validateGeneratedMixedDecimalRunnerBody(fset *token.FileSet, fn *ast.FuncDecl) error {
	return validateGeneratedExactFunctionBody(fset, fn, expectedGeneratedMixedDecimalRunnerBody())
}

func expectedGeneratedMixedDecimalRunnerBody() string {
	var b strings.Builder
	b.WriteString(`{
	op, err := parseGeneratedFFIMixedDecimalOperands(tc)
	if err != nil {
		return "", "", err
	}
	rounding := generatedFFICRoundingMode(tc.Rounding)
	var flags C._IDEC_flags

	switch tc.Function {
`)
	for _, function := range expectedFFIMixedDecimalFunctionOrder {
		shape, ok := expectedFFIMixedDecimalShapes[function]
		if !ok {
			panic("missing expected mixed decimal shape for " + function)
		}
		nativeCall := fmt.Sprintf("C.%s(%s)", function, strings.Join(expectedMixedDecimalCallArgs(shape, true), ", "))
		if shape.resultBits == 64 {
			nativeCall = "uint64(" + nativeCall + ")"
		} else {
			nativeCall = "ffiUint128FromC(" + nativeCall + ")"
		}
		goPortCall := fmt.Sprintf("bidgo.%s(%s)", shape.goPort, strings.Join(expectedMixedDecimalCallArgs(shape, false), ", "))
		fmt.Fprintf(&b, "\tcase %q:\n\t\tnative := %s\n\t\texposed, exposedFlags := %s\n", function, nativeCall, goPortCall)
		if shape.resultBits == 64 {
			b.WriteString("\t\treturn fmt.Sprintf(\"%016x/%08x\", native, uint32(flags)), fmt.Sprintf(\"%016x/%08x\", exposed, exposedFlags), nil\n")
		} else {
			b.WriteString("\t\treturn fmt.Sprintf(\"%s/%08x\", formatFFIUint128Bits(native), uint32(flags)), fmt.Sprintf(\"%s/%08x\", formatFFIUint128Bits(decimal128BIDFromBidgo(exposed)), exposedFlags), nil\n")
		}
	}
	b.WriteString(`	default:
		return "", "", fmt.Errorf("unsupported mixed decimal ffi function %q", tc.Function)
	}
}`)
	return b.String()
}

func validateGeneratedMixedDecimalWidenBody(fset *token.FileSet, fn *ast.FuncDecl) error {
	return validateGeneratedExactFunctionBody(fset, fn, `{
	switch width {
	case 64:
		return C.bid64_to_bid128(C.BID_UINT64(op.narrow[index]), flags), nil
	case 128:
		return ffiUint128ToC(op.wide[index]), nil
	default:
		var zero C.BID_UINT128
		return zero, fmt.Errorf("mixed FMA operand %d has unsupported width %d", index, width)
	}
}`)
}

func validateGeneratedMixedFMAComposedBody(fset *token.FileSet, fn *ast.FuncDecl) error {
	var forbidden string
	ast.Inspect(fn.Body, func(node ast.Node) bool {
		selector, ok := node.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		ident, ok := selector.X.(*ast.Ident)
		if ok && ident.Name == "tc" && (selector.Sel.Name == "Expected" || selector.Sel.Name == "Forbidden") {
			forbidden = "tc." + selector.Sel.Name
			return false
		}
		return true
	})
	if forbidden != "" {
		return fmt.Errorf("%s reads forbidden generated-case oracle field %s", fn.Name.Name, forbidden)
	}
	return validateGeneratedExactFunctionBody(fset, fn, `{
	shape, ok := generatedFFIMixedDecimalShapeFor(tc.Function)
	if !ok || shape.operation != "fma" || len(shape.operandBits) != 3 {
		return "", fmt.Errorf("unsupported mixed FMA composition %q", tc.Function)
	}
	op, err := parseGeneratedFFIMixedDecimalOperands(tc)
	if err != nil {
		return "", err
	}
	rounding := generatedFFICRoundingMode(tc.Rounding)
	var flags C._IDEC_flags
	operands := [3]C.BID_UINT128{}
	for i, width := range shape.operandBits {
		operands[i], err = generatedFFIWidenMixedDecimalOperand(op, width, i, &flags)
		if err != nil {
			return "", err
		}
	}
	product := C.bid128_mul(operands[0], operands[1], rounding, &flags)
	sum := C.bid128_add(product, operands[2], rounding, &flags)
	switch shape.resultBits {
	case 64:
		result := uint64(C.bid128_to_bid64(sum, rounding, &flags))
		return fmt.Sprintf("%016x/%08x", result, uint32(flags)), nil
	case 128:
		result := ffiUint128FromC(sum)
		return fmt.Sprintf("%s/%08x", formatFFIUint128Bits(result), uint32(flags)), nil
	default:
		return "", fmt.Errorf("mixed FMA %s has unsupported result width %d", tc.Function, shape.resultBits)
	}
}`)
}

func validateGeneratedFFIMainLoopBody(fset *token.FileSet, fn *ast.FuncDecl, expectedFFICaseCount int) error {
	expectedBody := strings.ReplaceAll(`{
	requireNative(t)
	if testing.Short() {
		t.Skip("generated FFI bit-compare cases require non-short native run; use make test-native-ffi")
	}
	spec := loadGeneratedFFISpecForTest(t)

	if len(spec.FFICases) == 0 {
		t.Fatal("expected generated ffi cases")
	}
	if len(spec.FFICases) != @@FFI_TOTAL@@ {
		t.Fatalf("generated ffi case count = %d, want @@FFI_TOTAL@@", len(spec.FFICases))
	}
	assertGeneratedFFICoverage(t, spec.FFICases)
	probeTracker, err := validateGeneratedFFIProbeContract(spec.FFICases)
	if err != nil {
		t.Fatalf("validate generated FFI probe contract: %v", err)
	}

	for _, tc := range spec.FFICases {
		tc := tc
		t.Run(tc.ID, func(t *testing.T) {
			generated := generatedFFICase(tc)
			gotNative, gotExposed, err := runGeneratedFFICase(generated)
			if err != nil {
				t.Fatalf("runGeneratedFFICase(%s): %v", tc.ID, err)
			}
			if gotNative != gotExposed {
				t.Fatalf("%s %s(%s): C=%s exposed=%s", tc.Declaration, tc.Function, strings.Join(tc.Operands, ", "), gotNative, gotExposed)
			}
			if tc.Probe == generatedFFIFusednessProbe {
				if gotNative != tc.Expected || gotExposed != tc.Expected {
					t.Fatalf("fusedness sentinel %s direct mismatch: pinned=%s C=%s Go-port=%s", tc.Function, tc.Expected, gotNative, gotExposed)
				}
				gotComposed, err := runGeneratedFFIMixedFMAComposed(generated)
				if err != nil {
					t.Fatalf("runGeneratedFFIMixedFMAComposed(%s): %v", tc.ID, err)
				}
				if gotComposed != tc.Forbidden {
					t.Fatalf("fusedness sentinel %s sequential composition drift: pinned forbidden=%s composed=%s", tc.Function, tc.Forbidden, gotComposed)
				}
				if gotComposed == gotNative {
					t.Fatalf("fusedness sentinel %s no longer discriminates direct FMA from sequential composition: %s", tc.Function, gotNative)
				}
			}
			if err := probeTracker.recordCanonicalResult(tc, gotNative); err != nil {
				t.Fatalf("record canonical FFI probe result: %v", err)
			}
		})
	}
	if err := probeTracker.validateCanonicalDiscrimination(); err != nil {
		t.Fatal(err)
	}
}`, "@@FFI_TOTAL@@", strconv.Itoa(expectedFFICaseCount))
	return validateGeneratedExactFunctionBody(fset, fn, expectedBody)
}

func validateGeneratedExactFunctionBody(fset *token.FileSet, fn *ast.FuncDecl, expectedBody string) error {
	expectedFSet := token.NewFileSet()
	expectedFile, err := parser.ParseFile(expectedFSet, "expected_generated_contract.go", "package expected\nfunc expected() "+expectedBody, parser.AllErrors)
	if err != nil {
		return fmt.Errorf("parse expected body for %s: %w", fn.Name.Name, err)
	}
	expected := expectedFile.Decls[0].(*ast.FuncDecl)
	actualText, err := formatGeneratedASTNode(fset, fn.Body)
	if err != nil {
		return fmt.Errorf("format actual body for %s: %w", fn.Name.Name, err)
	}
	expectedText, err := formatGeneratedASTNode(expectedFSet, expected.Body)
	if err != nil {
		return fmt.Errorf("format expected body for %s: %w", fn.Name.Name, err)
	}
	if actualText == expectedText {
		return nil
	}
	actualLines := strings.Split(actualText, "\n")
	expectedLines := strings.Split(expectedText, "\n")
	limit := min(len(actualLines), len(expectedLines))
	for i := 0; i < limit; i++ {
		if actualLines[i] != expectedLines[i] {
			return fmt.Errorf("%s live AST body diverges at line %d: got %q, want %q", fn.Name.Name, i+1, actualLines[i], expectedLines[i])
		}
	}
	return fmt.Errorf("%s live AST body line count = %d, want %d", fn.Name.Name, len(actualLines), len(expectedLines))
}

func formatGeneratedASTNode(fset *token.FileSet, node any) (string, error) {
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, node); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func assertGeneratedFFIMainLoopMutationsRejected(t *testing.T, source []byte, expectedFFICaseCount int) {
	t.Helper()
	mutations := []struct {
		name  string
		apply func(t *testing.T, fn *ast.FuncDecl)
	}{
		{
			name: "live self compare",
			apply: func(t *testing.T, fn *ast.FuncDecl) {
				_, closure := requireGeneratedFFIMainLoopNodes(t, fn)
				comparison := closure.Body.List[3].(*ast.IfStmt).Cond.(*ast.BinaryExpr)
				comparison.Y = ast.NewIdent("gotNative")
			},
		},
		{
			name: "correct comparison only in dead block",
			apply: func(t *testing.T, fn *ast.FuncDecl) {
				_, closure := requireGeneratedFFIMainLoopNodes(t, fn)
				statements := closure.Body.List
				original := statements[3].(*ast.IfStmt)
				dead := &ast.IfStmt{Cond: ast.NewIdent("false"), Body: &ast.BlockStmt{List: []ast.Stmt{original}}}
				live := &ast.IfStmt{
					Cond: &ast.BinaryExpr{X: ast.NewIdent("gotNative"), Op: token.NEQ, Y: ast.NewIdent("gotNative")},
					Body: original.Body,
				}
				closure.Body.List = append(append(append([]ast.Stmt(nil), statements[:3]...), dead, live), statements[4:]...)
			},
		},
		{
			name: "canonical record deleted",
			apply: func(t *testing.T, fn *ast.FuncDecl) {
				_, closure := requireGeneratedFFIMainLoopNodes(t, fn)
				closure.Body.List = append([]ast.Stmt(nil), closure.Body.List[:len(closure.Body.List)-1]...)
			},
		},
		{
			name: "canonical record moved before error handling",
			apply: func(t *testing.T, fn *ast.FuncDecl) {
				_, closure := requireGeneratedFFIMainLoopNodes(t, fn)
				statements := closure.Body.List
				record := statements[len(statements)-1]
				withoutRecord := append([]ast.Stmt(nil), statements[:len(statements)-1]...)
				closure.Body.List = append(append(append([]ast.Stmt(nil), withoutRecord[:2]...), record), withoutRecord[2:]...)
			},
		},
		{
			name: "final discrimination deleted",
			apply: func(t *testing.T, fn *ast.FuncDecl) {
				fn.Body.List = append([]ast.Stmt(nil), fn.Body.List[:len(fn.Body.List)-1]...)
			},
		},
		{
			name: "final discrimination moved inside case range",
			apply: func(t *testing.T, fn *ast.FuncDecl) {
				caseRange, _ := requireGeneratedFFIMainLoopNodes(t, fn)
				final := fn.Body.List[len(fn.Body.List)-1]
				fn.Body.List = append([]ast.Stmt(nil), fn.Body.List[:len(fn.Body.List)-1]...)
				caseRange.Body.List = append(caseRange.Body.List, final)
			},
		},
	}

	for _, mutation := range mutations {
		t.Run(mutation.name, func(t *testing.T) {
			fset, file := parseGeneratedGoOutput(t, ffiGeneratedNativeTestPath, source)
			fn := requireGeneratedFunc(t, file, "TestGeneratedFFIBitCompareSubset")
			mutation.apply(t, fn)
			if err := validateGeneratedFFIMainLoopBody(fset, fn, expectedFFICaseCount); err == nil {
				t.Fatalf("main FFI loop contract accepted mutation %q", mutation.name)
			}
		})
	}
}

func requireGeneratedFFIMainLoopNodes(t *testing.T, fn *ast.FuncDecl) (*ast.RangeStmt, *ast.FuncLit) {
	t.Helper()
	if len(fn.Body.List) < 10 {
		t.Fatalf("%s top-level statements = %d, want at least 10", fn.Name.Name, len(fn.Body.List))
	}
	caseRange, ok := fn.Body.List[8].(*ast.RangeStmt)
	if !ok {
		t.Fatalf("%s statement 8 is %T, want case range", fn.Name.Name, fn.Body.List[8])
	}
	if len(caseRange.Body.List) < 2 {
		t.Fatalf("%s case range statements = %d, want at least 2", fn.Name.Name, len(caseRange.Body.List))
	}
	runStmt, ok := caseRange.Body.List[1].(*ast.ExprStmt)
	if !ok {
		t.Fatalf("%s case range run statement is %T", fn.Name.Name, caseRange.Body.List[1])
	}
	runCall, ok := runStmt.X.(*ast.CallExpr)
	if !ok || len(runCall.Args) != 2 {
		t.Fatalf("%s case range t.Run expression is %T with invalid args", fn.Name.Name, runStmt.X)
	}
	closure, ok := runCall.Args[1].(*ast.FuncLit)
	if !ok {
		t.Fatalf("%s t.Run callback is %T, want func literal", fn.Name.Name, runCall.Args[1])
	}
	return caseRange, closure
}

func expectedMixedDecimalCallArgs(shape expectedFFIMixedDecimalShape, native bool) []string {
	args := make([]string, 0, len(shape.operandBits)+2)
	for i, bits := range shape.operandBits {
		index := strconv.Itoa(i)
		switch {
		case native && bits == 64:
			args = append(args, "C.BID_UINT64(op.narrow["+index+"])")
		case native && bits == 128:
			args = append(args, "ffiUint128ToC(op.wide["+index+"])")
		case !native && bits == 64:
			args = append(args, "op.narrow["+index+"]")
		case !native && bits == 128:
			args = append(args, "decimal128BIDAsBidgo(op.wide["+index+"])")
		default:
			panic("unsupported expected mixed decimal operand width")
		}
	}
	if native {
		return append(args, "rounding", "&flags")
	}
	return append(args, "tc.Rounding")
}

func parseGeneratedGoOutput(t *testing.T, name string, source []byte) (*token.FileSet, *ast.File) {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, name, source, parser.AllErrors)
	if err != nil {
		t.Fatalf("parse generated Go output %s: %v", name, err)
	}
	return fset, file
}

func requireGeneratedFunc(t *testing.T, file *ast.File, name string) *ast.FuncDecl {
	t.Helper()
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if ok && fn.Recv == nil && fn.Name.Name == name {
			return fn
		}
	}
	t.Fatalf("generated Go output is missing function %s", name)
	return nil
}

func requireGeneratedMethod(t *testing.T, file *ast.File, receiver, name string) *ast.FuncDecl {
	t.Helper()
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv == nil || fn.Name.Name != name || len(fn.Recv.List) != 1 {
			continue
		}
		if generatedReceiverName(fn.Recv.List[0].Type) == receiver {
			return fn
		}
	}
	t.Fatalf("generated Go output is missing method %s.%s", receiver, name)
	return nil
}

func generatedReceiverName(expr ast.Expr) string {
	switch expr := expr.(type) {
	case *ast.Ident:
		return expr.Name
	case *ast.StarExpr:
		return generatedReceiverName(expr.X)
	default:
		return ""
	}
}

func generatedCallsNamed(t *testing.T, fset *token.FileSet, node ast.Node, name string) []*ast.CallExpr {
	t.Helper()
	var calls []*ast.CallExpr
	ast.Inspect(node, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if ok && generatedNodeString(t, fset, call.Fun) == name {
			calls = append(calls, call)
		}
		return true
	})
	return calls
}

func generatedNodeString(t *testing.T, fset *token.FileSet, node any) string {
	t.Helper()
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, node); err != nil {
		t.Fatalf("format generated AST node %T: %v", node, err)
	}
	return buf.String()
}

func generatedFuncResultTypes(t *testing.T, fset *token.FileSet, fn *ast.FuncDecl) []string {
	t.Helper()
	if fn.Type.Results == nil {
		return nil
	}
	var results []string
	for _, field := range fn.Type.Results.List {
		count := len(field.Names)
		if count == 0 {
			count = 1
		}
		for range count {
			results = append(results, generatedNodeString(t, fset, field.Type))
		}
	}
	return results
}

func generatedFuncHasBinaryExpr(t *testing.T, fset *token.FileSet, fn *ast.FuncDecl, left string, op token.Token, right string) bool {
	t.Helper()
	found := false
	ast.Inspect(fn.Body, func(node ast.Node) bool {
		expr, ok := node.(*ast.BinaryExpr)
		if ok && expr.Op == op && generatedNodeString(t, fset, expr.X) == left && generatedNodeString(t, fset, expr.Y) == right {
			found = true
		}
		return !found
	})
	return found
}

func generatedFuncHasStringLiteral(fn *ast.FuncDecl, want string) bool {
	found := false
	ast.Inspect(fn.Body, func(node ast.Node) bool {
		literal, ok := node.(*ast.BasicLit)
		if !ok || literal.Kind != token.STRING {
			return true
		}
		value, err := strconv.Unquote(literal.Value)
		if err == nil && value == want {
			found = true
		}
		return !found
	})
	return found
}
