package testgen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/sky1core/bid754/devtools/internal/genmarker"
)

// The goport readtest outputs run the full generated readtest case surface
// directly against the Go mechanical port (bid754-go/internal/bidgo) without
// cgo, so the portable CI legs exercise the product implementation against
// the pinned expected values.
const (
	readtestGoportDispatchPath  = "../bid754-go/generated_readtest_goport_dispatch_test.go"
	readtestGoportCasesPath     = "../bid754-go/generated_readtest_goport_cases_test.go"
	readtestGoportInventoryPath = "generated/testspec/goport_readtest_dispatch_inventory.json"
	readtestGoportBidgoSrcDir   = "../bid754-go/internal/bidgo"
)

// goportReadtestOperandDropExceptions enumerates the only readtest operands
// (function name -> operand index -> reason) that the generated goport
// dispatch may parse and then intentionally not pass to the Go mechanical
// port, because the matched Go function does not take the corresponding
// Intel C value parameter. Every other unconsumed value parameter is a
// generation-time hard failure.
var goportReadtestOperandDropExceptions = map[string]map[int]string{
	"bid32_radix":  {0: "Bid32Radix takes no operand; the Intel C parameter is an unused dummy input"},
	"bid64_radix":  {0: "Bid64Radix takes no operand; the Intel C parameter is an unused dummy input"},
	"bid128_radix": {0: "Bid128Radix takes no operand; the Intel C parameter is an unused dummy input"},
}

type bidgoParam struct {
	Name string
	Type string
}

type bidgoFuncSig struct {
	Name    string
	Params  []bidgoParam
	Results []bidgoParam
}

type goportFlagsKind string

const (
	goportFlagsNone    goportFlagsKind = "none"
	goportFlagsPointer goportFlagsKind = "pointer"
	goportFlagsResult  goportFlagsKind = "result"
)

type goportArgPlan struct {
	OperandIndex int
	InputType    string
	Kind         readtestParamKind
	GoType       string // empty means parsed for validation and dropped
}

// goportStatusControlShape pins the readtest row wiring and the expected Go
// mechanical-port signature of one IEEE 754-2019 section 5.7.4 status-control
// operation. classifyReadtestParameters does not model the by-value
// _IDEC_flags / _IDEC_round parameters these eight functions take (the native
// C generator special-cases them for the same reason), so the shape is declared
// here instead of inferred, and every part of it is verified against the parsed
// bidgo signature at generation time.
//
// Operands are the readtest row's operand slots. ValueOperands are passed to
// the port as uint32 value arguments in Go parameter order; InitialFlagsOperand
// (-1 when the port takes no status-word pointer) is the incoming status word,
// masked with BID_FLAG_MASK into a local status word exactly like the native C
// wrapper does. Any remaining operand is parsed for validation parity and
// dropped, mirroring the native dispatch.
type goportStatusControlShape struct {
	Operands            int
	ValueOperands       []int
	InitialFlagsOperand int
	PassRounding        bool
	HasFlagsPointer     bool
	HasResult           bool
}

// expectedParams derives the Go port parameter types this shape requires:
// one uint32 per value operand, then the rounding mode, then the status-word
// pointer.
func (shape goportStatusControlShape) expectedParams() []string {
	params := make([]string, 0, len(shape.ValueOperands)+2)
	for range shape.ValueOperands {
		params = append(params, "uint32")
	}
	if shape.PassRounding {
		params = append(params, "uint32")
	}
	if shape.HasFlagsPointer {
		params = append(params, "*uint32")
	}
	return params
}

func (shape goportStatusControlShape) expectedResults() []string {
	if shape.HasResult {
		return []string{"uint32"}
	}
	return nil
}

// goportStatusControlShapes is the closed world of status-control rows the
// goport readtest dispatch drives. A readtest status-control function without
// an entry here is a generation-time failure rather than a silent skip.
var goportStatusControlShapes = map[string]goportStatusControlShape{
	"bid_testFlags":                   {Operands: 2, ValueOperands: []int{0}, InitialFlagsOperand: 1, HasFlagsPointer: true, HasResult: true},
	"bid_lowerFlags":                  {Operands: 2, ValueOperands: []int{0}, InitialFlagsOperand: 1, HasFlagsPointer: true},
	"bid_signalException":             {Operands: 2, ValueOperands: []int{0}, InitialFlagsOperand: 1, HasFlagsPointer: true},
	"bid_saveFlags":                   {Operands: 2, ValueOperands: []int{0}, InitialFlagsOperand: 1, HasFlagsPointer: true, HasResult: true},
	"bid_restoreFlags":                {Operands: 3, ValueOperands: []int{0, 1}, InitialFlagsOperand: 2, HasFlagsPointer: true},
	"bid_testSavedFlags":              {Operands: 2, ValueOperands: []int{0, 1}, InitialFlagsOperand: -1, HasResult: true},
	"bid_getDecimalRoundingDirection": {Operands: 1, InitialFlagsOperand: -1, PassRounding: true, HasResult: true},
	"bid_setDecimalRoundingDirection": {Operands: 1, ValueOperands: []int{0}, InitialFlagsOperand: -1, PassRounding: true, HasResult: true},
}

type goportCallPlan struct {
	Function     string
	GoName       string
	OperandCount int
	Args         []goportArgPlan
	HasRounding  bool
	FlagsKind    goportFlagsKind
	CHasFlags    bool
	Results      []bidgoParam
	// StatusControl is non-nil only for the section 5.7.4 status-control rows,
	// which carry their own declared shape instead of an inferred Args plan.
	StatusControl *goportStatusControlShape
	// SecondaryKind/SecondaryOperandIndex describe the Intel in/out pointer
	// parameter (frexp exponent, modf integral part). The Go port returns
	// that output as its second result; the runner compares it against the
	// readtest row operand at SecondaryOperandIndex like readtest.c
	// check_results does. SecondaryKind is "none" when the C prototype has
	// no pointer output.
	SecondaryKind         string
	SecondaryOperandIndex int
}

type goportInventoryRow struct {
	Function   string `json:"function"`
	GoFunction string `json:"go_function,omitempty"`
	Status     string `json:"status"`
	Reason     string `json:"reason,omitempty"`
}

type goportInventory struct {
	Version    string               `json:"version"`
	Source     string               `json:"source"`
	Dispatched int                  `json:"dispatched"`
	Functions  []goportInventoryRow `json:"functions"`
}

func WriteReadtestGoportOutputs(repoRoot string, manifest Manifest, spec SharedSpec) error {
	files, err := GenerateReadtestGoportOutputs(repoRoot, manifest, spec)
	if err != nil {
		return err
	}
	for path, data := range files {
		fullPath := filepath.Join(repoRoot, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			return fmt.Errorf("mkdir %q: %w", filepath.Dir(fullPath), err)
		}
		if err := os.WriteFile(fullPath, data, 0o644); err != nil {
			return fmt.Errorf("write generated readtest goport output %q: %w", fullPath, err)
		}
	}
	return nil
}

func GenerateReadtestGoportOutputs(repoRoot string, manifest Manifest, spec SharedSpec) (map[string][]byte, error) {
	reads, err := collectReadtestDispatchSpecs(repoRoot, manifest)
	if err != nil {
		return nil, err
	}
	symbols, err := loadSymbolFile(filepath.Join(repoRoot, readtestSymbolSourcePath))
	if err != nil {
		return nil, err
	}
	symbolIndex := make(map[string]symbolSpec, len(symbols.Symbols))
	for _, symbol := range symbols.Symbols {
		symbolIndex[symbol.Name] = symbol
	}
	sigs, err := loadBidgoFuncSigs(filepath.Join(repoRoot, readtestGoportBidgoSrcDir))
	if err != nil {
		return nil, err
	}

	dispatch, err := generateReadtestGoportDispatch(reads, symbolIndex, sigs)
	if err != nil {
		return nil, err
	}
	cases := generateReadtestGoportCases(spec)
	inventory, err := generateReadtestGoportInventory(reads, symbolIndex, sigs)
	if err != nil {
		return nil, err
	}
	files, err := formatGeneratedGoOutputs(map[string][]byte{
		readtestGoportDispatchPath: dispatch,
		readtestGoportCasesPath:    cases,
	})
	if err != nil {
		return nil, err
	}
	files[readtestGoportInventoryPath] = inventory
	return files, nil
}

func loadBidgoFuncSigs(dir string) (map[string]bidgoFuncSig, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read bidgo source dir %q: %w", dir, err)
	}
	sigs := map[string]bidgoFuncSig{}
	fset := token.NewFileSet()
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		file, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if err != nil {
			return nil, fmt.Errorf("parse bidgo source %q: %w", path, err)
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv != nil || !fn.Name.IsExported() {
				continue
			}
			sig := bidgoFuncSig{Name: fn.Name.Name}
			if fn.Type.Params != nil {
				for _, field := range fn.Type.Params.List {
					typeName := bidgoExprString(field.Type)
					if len(field.Names) == 0 {
						sig.Params = append(sig.Params, bidgoParam{Type: typeName})
						continue
					}
					for _, name := range field.Names {
						sig.Params = append(sig.Params, bidgoParam{Name: name.Name, Type: typeName})
					}
				}
			}
			if fn.Type.Results != nil {
				for _, field := range fn.Type.Results.List {
					typeName := bidgoExprString(field.Type)
					if len(field.Names) == 0 {
						sig.Results = append(sig.Results, bidgoParam{Type: typeName})
						continue
					}
					for _, name := range field.Names {
						sig.Results = append(sig.Results, bidgoParam{Name: name.Name, Type: typeName})
					}
				}
			}
			sigs[sig.Name] = sig
		}
	}
	if len(sigs) == 0 {
		return nil, fmt.Errorf("no exported bidgo functions found under %q", dir)
	}
	return sigs, nil
}

func bidgoExprString(expr ast.Expr) string {
	switch typed := expr.(type) {
	case *ast.Ident:
		return typed.Name
	case *ast.StarExpr:
		return "*" + bidgoExprString(typed.X)
	default:
		return fmt.Sprintf("<unsupported %T>", expr)
	}
}

func normalizeGoportFuncName(name string) string {
	return NormalizeReadtestFuncName(name)
}

func goportCandidateSigs(function string, sigs map[string]bidgoFuncSig) []bidgoFuncSig {
	base := normalizeGoportFuncName(function)
	wanted := map[string]bool{
		base:               true,
		base + "withflags": true,
		base + "raw":       true,
	}
	var names []string
	for name := range sigs {
		if wanted[normalizeGoportFuncName(name)] {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	candidates := make([]bidgoFuncSig, 0, len(names))
	for _, name := range names {
		candidates = append(candidates, sigs[name])
	}
	return candidates
}

func isGoportRoundingParam(param bidgoParam) bool {
	if param.Type != "int" {
		return false
	}
	name := strings.ToLower(param.Name)
	return strings.Contains(name, "rnd") || strings.Contains(name, "round")
}

func goportValueParamTypes(kind readtestParamKind) []string {
	switch kind {
	case readtestParamU32:
		return []string{"uint32"}
	case readtestParamU64:
		return []string{"uint64"}
	case readtestParamU128:
		return []string{"BID_UINT128"}
	case readtestParamInt:
		return []string{"int", "int32"}
	case readtestParamUInt:
		return []string{"uint32", "uint"}
	case readtestParamLInt:
		return []string{"int64", "int"}
	case readtestParamS64:
		return []string{"int64"}
	case readtestParamCStr:
		return []string{"string"}
	default:
		return nil
	}
}

func goportPrimaryResultTypes(read ReadTestSpec) []string {
	switch {
	case isReadtestDecimalOutput(read.OutputType) && read.Format == "decimal32":
		return []string{"uint32"}
	case isReadtestDecimalOutput(read.OutputType) && read.Format == "decimal64":
		return []string{"uint64"}
	case isReadtestDecimalOutput(read.OutputType) && read.Format == "decimal128":
		return []string{"BID_UINT128"}
	case isReadtestBinary32Output(read.OutputType):
		return []string{"uint32", "float32"}
	case isReadtestBinary64Output(read.OutputType):
		return []string{"uint64", "float64"}
	case isReadtestBinary128Output(read.OutputType):
		return []string{"BID_UINT128"}
	case isReadtestUnsignedOutput(read.OutputType):
		return []string{"uint8", "uint16", "uint32", "uint64"}
	default:
		// bool covers Go port predicates that return bool where Intel C
		// returns int 0/1.
		return []string{"int", "int8", "int16", "int32", "int64", "bool"}
	}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

// goportSecondaryKindForParam maps an Intel in/out pointer parameter kind to
// the shared readtestSecondaryOutput kind and the Go port result type that
// must carry the output.
func goportSecondaryKindForParam(kind readtestParamKind) (secondaryKind, goResultType string, ok bool) {
	switch kind {
	case readtestParamP32:
		return "dec32", "uint32", true
	case readtestParamP64:
		return "dec64", "uint64", true
	case readtestParamP128:
		return "dec128", "BID_UINT128", true
	case readtestParamPInt:
		return "int", "int", true
	default:
		return "", "", false
	}
}

func buildGoportCallPlan(read ReadTestSpec, kinds []readtestParamKind, sig bidgoFuncSig) (goportCallPlan, error) {
	plan := goportCallPlan{Function: read.Function, GoName: sig.Name, OperandCount: len(read.InputTypes), Results: sig.Results, SecondaryKind: "none", SecondaryOperandIndex: -1}
	goIdx := 0
	operandIdx := 0
	cHasFlags := false
	for _, kind := range kinds {
		switch kind {
		case readtestParamU32, readtestParamU64, readtestParamU128, readtestParamInt, readtestParamUInt, readtestParamLInt, readtestParamS64, readtestParamCStr:
			if operandIdx >= len(read.InputTypes) {
				return plan, fmt.Errorf("%s: symbol value parameter %d has no readtest input type", read.Function, operandIdx)
			}
			arg := goportArgPlan{OperandIndex: operandIdx, InputType: read.InputTypes[operandIdx], Kind: kind}
			allowed := goportValueParamTypes(kind)
			if goIdx < len(sig.Params) && containsString(allowed, sig.Params[goIdx].Type) && !isGoportRoundingParam(sig.Params[goIdx]) {
				arg.GoType = sig.Params[goIdx].Type
				goIdx++
			} else if _, ok := goportReadtestOperandDropExceptions[read.Function][operandIdx]; ok {
				arg.GoType = ""
			} else {
				return plan, fmt.Errorf("%s: %s does not consume %s value operand %d", read.Function, sig.Name, kind, operandIdx)
			}
			plan.Args = append(plan.Args, arg)
			operandIdx++
		case readtestParamP32, readtestParamP64, readtestParamP128, readtestParamPInt:
			if operandIdx >= len(read.InputTypes) {
				return plan, fmt.Errorf("%s: symbol pointer parameter %d has no readtest input type", read.Function, operandIdx)
			}
			// Intel C uses these as in/out outputs; the Go port returns the
			// extra output as its second result instead, so the operand is
			// parsed for validation parity, not passed, and pinned as the
			// expected secondary output.
			if plan.SecondaryKind != "none" {
				return plan, fmt.Errorf("%s: multiple pointer output parameters are not supported", read.Function)
			}
			secondaryKind, _, ok := goportSecondaryKindForParam(kind)
			if !ok {
				return plan, fmt.Errorf("%s: unsupported pointer parameter kind %q", read.Function, kind)
			}
			plan.SecondaryKind = secondaryKind
			plan.SecondaryOperandIndex = operandIdx
			plan.Args = append(plan.Args, goportArgPlan{OperandIndex: operandIdx, InputType: read.InputTypes[operandIdx], Kind: kind})
			operandIdx++
		case readtestParamRound:
			if goIdx < len(sig.Params) && isGoportRoundingParam(sig.Params[goIdx]) {
				plan.HasRounding = true
				goIdx++
			} else {
				return plan, fmt.Errorf("%s: %s lacks the rounding parameter required by the symbol", read.Function, sig.Name)
			}
		case readtestParamFlags:
			cHasFlags = true
		default:
			return plan, fmt.Errorf("%s: unsupported symbol parameter kind %q", read.Function, kind)
		}
	}
	switch {
	case goIdx == len(sig.Params):
		plan.FlagsKind = goportFlagsNone
	case goIdx == len(sig.Params)-1 && sig.Params[goIdx].Type == "*uint32":
		plan.FlagsKind = goportFlagsPointer
		goIdx++
	default:
		return plan, fmt.Errorf("%s: %s has unconsumed parameters %v", read.Function, sig.Name, sig.Params[goIdx:])
	}
	// A trailing uint32 result carries the port's status flags only when the
	// Intel C prototype takes a *pfpsf parameter. Without that parameter the
	// upstream readtest harness compares expected_status against an untouched
	// *pfpsf (for example bid32_frexp, readtest.h NOFLAGS call), so a
	// port-side extra uint32 result must not be wired into the status
	// comparison.
	if plan.FlagsKind == goportFlagsNone && cHasFlags && len(sig.Results) >= 2 && sig.Results[len(sig.Results)-1].Type == "uint32" {
		plan.FlagsKind = goportFlagsResult
	}
	plan.CHasFlags = cHasFlags
	if len(sig.Results) == 0 {
		return plan, fmt.Errorf("%s: %s has no results", read.Function, sig.Name)
	}
	if !containsString(goportPrimaryResultTypes(read), sig.Results[0].Type) {
		return plan, fmt.Errorf("%s: %s primary result type %s does not match output %s/%s", read.Function, sig.Name, sig.Results[0].Type, read.OutputType, read.Format)
	}
	if plan.SecondaryKind != "none" {
		wantType := map[string]string{"dec32": "uint32", "dec64": "uint64", "dec128": "BID_UINT128", "int": "int"}[plan.SecondaryKind]
		if len(sig.Results) < 2 {
			return plan, fmt.Errorf("%s: %s does not return the pointer output as a second result", read.Function, sig.Name)
		}
		if sig.Results[1].Type != wantType {
			return plan, fmt.Errorf("%s: %s second result type %s does not match pointer output type %s", read.Function, sig.Name, sig.Results[1].Type, wantType)
		}
		if plan.FlagsKind == goportFlagsResult && len(sig.Results) < 3 {
			return plan, fmt.Errorf("%s: %s cannot carry both the pointer output and the flags in a single extra result", read.Function, sig.Name)
		}
	}
	return plan, nil
}

func resolveGoportDispatch(read ReadTestSpec, kinds []readtestParamKind, sigs map[string]bidgoFuncSig) (goportCallPlan, error) {
	candidates := goportCandidateSigs(read.Function, sigs)
	if len(candidates) == 0 {
		return goportCallPlan{}, fmt.Errorf("readtest goport dispatch: no bidgo candidate found for %q (normalized %q)", read.Function, normalizeGoportFuncName(read.Function))
	}
	var fits []goportCallPlan
	var failures []string
	for _, candidate := range candidates {
		plan, err := buildGoportCallPlan(read, kinds, candidate)
		if err != nil {
			failures = append(failures, err.Error())
			continue
		}
		fits = append(fits, plan)
	}
	// When the Intel C symbol reports flags, prefer port variants that
	// expose flags; a flagless port (reported as status 00) is accepted
	// only when no flag-bearing variant exists, and any expected non-zero
	// status then fails honestly at run time.
	if len(fits) > 1 {
		var flagged []goportCallPlan
		for _, fit := range fits {
			if fit.CHasFlags && fit.FlagsKind != goportFlagsNone {
				flagged = append(flagged, fit)
			}
		}
		if len(flagged) > 0 && len(flagged) < len(fits) {
			fits = flagged
		}
	}
	if len(fits) == 1 {
		return fits[0], nil
	}
	if len(fits) > 1 {
		base := normalizeGoportFuncName(read.Function)
		// Mirror the Rust readtest generator preference: when the row needs
		// status flags, the explicit *WithFlags companion is the flags-
		// faithful surface for ports whose base entry drops flag
		// propagation; otherwise prefer the exact base name.
		preferred := []string{base}
		if len(fits) > 0 && fits[0].CHasFlags {
			preferred = []string{base + "withflags", base}
		}
		for _, want := range preferred {
			var exact []goportCallPlan
			for _, fit := range fits {
				if normalizeGoportFuncName(fit.GoName) == want {
					exact = append(exact, fit)
				}
			}
			if len(exact) == 1 {
				return exact[0], nil
			}
		}
		names := make([]string, 0, len(fits))
		for _, fit := range fits {
			names = append(names, fit.GoName)
		}
		return goportCallPlan{}, fmt.Errorf("readtest goport dispatch: ambiguous bidgo candidates for %q: %v", read.Function, names)
	}
	return goportCallPlan{}, fmt.Errorf("readtest goport dispatch: no shape-fit bidgo candidate for %q:\n  %s", read.Function, strings.Join(failures, "\n  "))
}

// resolveGoportStatusControlDispatch resolves the Go mechanical-port function
// for one status-control readtest row and verifies it against the declared
// shape: the row's operand count and operand types, the exact port name (no
// *WithFlags/*Raw variant is accepted here — these operations have a single
// port entry), and the port's parameter and result types.
func resolveGoportStatusControlDispatch(read ReadTestSpec, sigs map[string]bidgoFuncSig) (goportCallPlan, error) {
	shape, ok := goportStatusControlShapes[read.Function]
	if !ok {
		return goportCallPlan{}, fmt.Errorf("readtest goport dispatch: status-control function %q has no declared shape", read.Function)
	}
	if read.Kind != "status_control" {
		return goportCallPlan{}, fmt.Errorf("readtest goport dispatch: %s is a status-control function but its spec kind is %q", read.Function, read.Kind)
	}
	if !isReadtestUnsignedOutput(read.OutputType) {
		return goportCallPlan{}, fmt.Errorf("readtest goport dispatch: %s status-control output type %q is not an unsigned scalar", read.Function, read.OutputType)
	}
	if len(read.InputTypes) != shape.Operands {
		return goportCallPlan{}, fmt.Errorf("readtest goport dispatch: %s has %d readtest operands, declared shape expects %d", read.Function, len(read.InputTypes), shape.Operands)
	}
	for index, inputType := range read.InputTypes {
		parser, err := goportOperandParser(readtestParamU32, inputType)
		if err != nil {
			return goportCallPlan{}, fmt.Errorf("%s: %w", read.Function, err)
		}
		if parser != "parseReadtestUint" {
			return goportCallPlan{}, fmt.Errorf("readtest goport dispatch: %s operand %d input type %q does not parse as an unsigned status/mode word", read.Function, index, inputType)
		}
	}
	for _, operand := range shape.ValueOperands {
		if operand < 0 || operand >= shape.Operands {
			return goportCallPlan{}, fmt.Errorf("readtest goport dispatch: %s declared value operand %d is outside its %d operands", read.Function, operand, shape.Operands)
		}
		if operand == shape.InitialFlagsOperand {
			return goportCallPlan{}, fmt.Errorf("readtest goport dispatch: %s operand %d cannot be both a value argument and the incoming status word", read.Function, operand)
		}
	}
	if shape.InitialFlagsOperand >= shape.Operands {
		return goportCallPlan{}, fmt.Errorf("readtest goport dispatch: %s declared status-word operand %d is outside its %d operands", read.Function, shape.InitialFlagsOperand, shape.Operands)
	}
	if shape.HasFlagsPointer != (shape.InitialFlagsOperand >= 0) {
		return goportCallPlan{}, fmt.Errorf("readtest goport dispatch: %s status-word pointer and incoming status-word operand disagree", read.Function)
	}

	base := normalizeGoportFuncName(read.Function)
	var fits []bidgoFuncSig
	for _, candidate := range goportCandidateSigs(read.Function, sigs) {
		if normalizeGoportFuncName(candidate.Name) == base {
			fits = append(fits, candidate)
		}
	}
	if len(fits) != 1 {
		return goportCallPlan{}, fmt.Errorf("readtest goport dispatch: expected exactly one bidgo entry named %q for status-control function %q, found %d", base, read.Function, len(fits))
	}
	sig := fits[0]
	gotParams := make([]string, 0, len(sig.Params))
	for _, param := range sig.Params {
		gotParams = append(gotParams, param.Type)
	}
	gotResults := make([]string, 0, len(sig.Results))
	for _, result := range sig.Results {
		gotResults = append(gotResults, result.Type)
	}
	if !equalStringSlices(gotParams, shape.expectedParams()) {
		return goportCallPlan{}, fmt.Errorf("readtest goport dispatch: %s port %s has parameters %v, declared shape expects %v", read.Function, sig.Name, gotParams, shape.expectedParams())
	}
	if !equalStringSlices(gotResults, shape.expectedResults()) {
		return goportCallPlan{}, fmt.Errorf("readtest goport dispatch: %s port %s has results %v, declared shape expects %v", read.Function, sig.Name, gotResults, shape.expectedResults())
	}
	flagsKind := goportFlagsNone
	if shape.HasFlagsPointer {
		flagsKind = goportFlagsPointer
	}
	// Args stays nil: StatusControl carries the operand wiring instead of the
	// inferred per-operand plan.
	return goportCallPlan{
		Function:              read.Function,
		GoName:                sig.Name,
		OperandCount:          shape.Operands,
		HasRounding:           shape.PassRounding,
		FlagsKind:             flagsKind,
		CHasFlags:             shape.HasFlagsPointer,
		Results:               sig.Results,
		SecondaryKind:         "none",
		SecondaryOperandIndex: -1,
		StatusControl:         &shape,
	}, nil
}

type goportStringEntry struct {
	Width  string // "32", "64", "128"
	GoName string
}

func resolveGoportFromString(width string, sigs map[string]bidgoFuncSig) (goportStringEntry, error) {
	function := "bid" + width + "_from_string"
	primary := map[string]string{"32": "uint32", "64": "uint64", "128": "BID_UINT128"}[width]
	candidates := goportCandidateSigs(function, sigs)
	var fits []bidgoFuncSig
	for _, candidate := range candidates {
		if len(candidate.Params) != 2 || candidate.Params[0].Type != "string" || !isGoportRoundingParam(candidate.Params[1]) {
			continue
		}
		if len(candidate.Results) != 2 || candidate.Results[0].Type != primary || candidate.Results[1].Type != "uint32" {
			continue
		}
		fits = append(fits, candidate)
	}
	if len(fits) != 1 {
		return goportStringEntry{}, fmt.Errorf("readtest goport dispatch: expected exactly one bidgo from-string entry for %q, found %d", function, len(fits))
	}
	return goportStringEntry{Width: width, GoName: fits[0].Name}, nil
}

func resolveGoportToString(width string, sigs map[string]bidgoFuncSig) (goportStringEntry, error) {
	function := "bid" + width + "_to_string"
	primary := map[string]string{"32": "uint32", "64": "uint64", "128": "BID_UINT128"}[width]
	candidates := goportCandidateSigs(function, sigs)
	var fits []bidgoFuncSig
	for _, candidate := range candidates {
		if len(candidate.Params) != 1 || candidate.Params[0].Type != primary {
			continue
		}
		if len(candidate.Results) != 1 || candidate.Results[0].Type != "string" {
			continue
		}
		fits = append(fits, candidate)
	}
	if len(fits) > 1 {
		base := normalizeGoportFuncName(function)
		var exact []bidgoFuncSig
		for _, fit := range fits {
			if normalizeGoportFuncName(fit.Name) == base {
				exact = append(exact, fit)
			}
		}
		if len(exact) == 1 {
			fits = exact
		}
	}
	if len(fits) != 1 {
		return goportStringEntry{}, fmt.Errorf("readtest goport dispatch: expected exactly one bidgo to-string entry for %q, found %d", function, len(fits))
	}
	return goportStringEntry{Width: width, GoName: fits[0].Name}, nil
}

type goportDispatchSet struct {
	Decimal32  []goportCallPlan
	Decimal64  []goportCallPlan
	Decimal128 []goportCallPlan
	Signed     []goportCallPlan
	Unsigned   []goportCallPlan
	Binary32   []goportCallPlan
	Binary64   []goportCallPlan
	Binary128  []goportCallPlan
	FromString []goportStringEntry
	ToString   []goportStringEntry
}

func resolveGoportDispatchSet(reads []ReadTestSpec, symbols map[string]symbolSpec, sigs map[string]bidgoFuncSig) (goportDispatchSet, error) {
	var set goportDispatchSet
	for _, read := range reads {
		if isReadtestStatusControlFunction(read.Function) {
			plan, err := resolveGoportStatusControlDispatch(read, sigs)
			if err != nil {
				return set, err
			}
			set.Unsigned = append(set.Unsigned, plan)
			continue
		}
		switch read.Kind {
		case "from_string":
			width := strings.TrimSuffix(strings.TrimPrefix(read.Function, "bid"), "_from_string")
			entry, err := resolveGoportFromString(width, sigs)
			if err != nil {
				return set, err
			}
			set.FromString = append(set.FromString, entry)
			continue
		case "to_string":
			width := strings.TrimSuffix(strings.TrimPrefix(read.Function, "bid"), "_to_string")
			entry, err := resolveGoportToString(width, sigs)
			if err != nil {
				return set, err
			}
			set.ToString = append(set.ToString, entry)
			continue
		}
		symbol, ok := symbols[read.Function]
		if !ok {
			return set, fmt.Errorf("readtest goport dispatch: symbol %q not found in %s", read.Function, readtestSymbolSourcePath)
		}
		kinds, err := classifyReadtestParameters(symbol.Parameters)
		if err != nil {
			return set, fmt.Errorf("%s: %w", read.Function, err)
		}
		plan, err := resolveGoportDispatch(read, kinds, sigs)
		if err != nil {
			return set, err
		}
		switch {
		case isReadtestDecimalOutput(read.OutputType) && read.Format == "decimal32":
			set.Decimal32 = append(set.Decimal32, plan)
		case isReadtestDecimalOutput(read.OutputType) && read.Format == "decimal64":
			set.Decimal64 = append(set.Decimal64, plan)
		case isReadtestDecimalOutput(read.OutputType) && read.Format == "decimal128":
			set.Decimal128 = append(set.Decimal128, plan)
		case isReadtestBinary32Output(read.OutputType):
			set.Binary32 = append(set.Binary32, plan)
		case isReadtestBinary64Output(read.OutputType):
			set.Binary64 = append(set.Binary64, plan)
		case isReadtestBinary128Output(read.OutputType):
			set.Binary128 = append(set.Binary128, plan)
		case isReadtestUnsignedOutput(read.OutputType):
			set.Unsigned = append(set.Unsigned, plan)
		default:
			set.Signed = append(set.Signed, plan)
		}
	}
	return set, nil
}

func generateReadtestGoportDispatch(reads []ReadTestSpec, symbols map[string]symbolSpec, sigs map[string]bidgoFuncSig) ([]byte, error) {
	set, err := resolveGoportDispatchSet(reads, symbols, sigs)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	buf.WriteString(genmarker.Line("testgen") + "\n\n")
	buf.WriteString("package bid754\n\n")
	buf.WriteString("import (\n")
	buf.WriteString("\t\"encoding/binary\"\n")
	buf.WriteString("\t\"fmt\"\n")
	buf.WriteString("\t\"math\"\n")
	buf.WriteString("\t\"strings\"\n\n")
	buf.WriteString("\t\"github.com/sky1core/bid754/bid754-go/internal/bidgo\"\n")
	buf.WriteString(")\n\n")

	for _, entry := range set.FromString {
		emitGoportFromString(&buf, entry)
	}
	for _, entry := range set.ToString {
		emitGoportToString(&buf, entry)
	}

	type dispatchGroup struct {
		name          string
		plans         []goportCallPlan
		category      string
		withSecondary bool
	}
	groups := []dispatchGroup{
		{"goportReadtestGeneratedBID32", set.Decimal32, "dec32", true},
		{"goportReadtestGeneratedBID64", set.Decimal64, "dec64", true},
		{"goportReadtestGeneratedBID128", set.Decimal128, "dec128", true},
		{"goportReadtestGeneratedSigned", set.Signed, "signed", false},
		{"goportReadtestGeneratedBinary32", set.Binary32, "bin32", false},
		{"goportReadtestGeneratedBinary64", set.Binary64, "bin64", false},
		{"goportReadtestGeneratedBinary128", set.Binary128, "bin128", false},
		{"goportReadtestGeneratedUnsigned", set.Unsigned, "unsigned", false},
	}
	for _, group := range groups {
		if err := emitGoportDispatchFunc(&buf, group.name, group.category, group.plans, group.withSecondary); err != nil {
			return nil, err
		}
	}

	buf.WriteString(readtestGoportHelpers)
	return buf.Bytes(), nil
}

func goportCategoryReturnType(category string) (goType, zero string) {
	switch category {
	case "dec32", "bin32":
		return "uint32", "0"
	case "dec64", "bin64":
		return "uint64", "0"
	case "dec128", "bin128":
		return "[16]byte", "[16]byte{}"
	case "signed":
		return "int64", "0"
	case "unsigned":
		return "uint64", "0"
	default:
		return "", ""
	}
}

func emitGoportDispatchFunc(buf *bytes.Buffer, name, category string, plans []goportCallPlan, withSecondary bool) error {
	returnType, zero := goportCategoryReturnType(category)
	if withSecondary {
		fmt.Fprintf(buf, "func %s(function string, rounding int, operands []string) (%s, readtestSecondaryOutput, string, error) {\n", name, returnType)
	} else {
		fmt.Fprintf(buf, "func %s(function string, rounding int, operands []string) (%s, string, error) {\n", name, returnType)
	}
	buf.WriteString("\tswitch function {\n")
	for _, plan := range plans {
		if err := emitGoportDispatchCase(buf, category, zero, plan, withSecondary); err != nil {
			return err
		}
	}
	zeroReturn := zero
	if withSecondary {
		zeroReturn = zero + ", readtestNoSecondaryOutput()"
	}
	fmt.Fprintf(buf, "\tdefault:\n\t\treturn %s, \"\", fmt.Errorf(\"unsupported goport readtest %s function %%q\", function)\n\t}\n}\n\n", zeroReturn, category)
	return nil
}

func goportOperandParser(kind readtestParamKind, inputType string) (string, error) {
	switch inputType {
	case "OP_DEC32":
		return "parseReadtestBits32", nil
	case "OP_DEC64":
		return "parseReadtestBits64", nil
	case "OP_DEC128":
		return "parseReadtestBits128", nil
	case "OP_INT8", "OP_INT16", "OP_INT32", "OP_INT64", "OP_LINT":
		return "parseReadtestInt", nil
	case "OP_BID_UINT8", "OP_BID_UINT16", "OP_BID_UINT32", "OP_BID_UINT64":
		return "parseReadtestUint", nil
	default:
		return "", fmt.Errorf("unsupported readtest input type %q for kind %q", inputType, kind)
	}
}

// goportArgExpr renders the call argument for a kept value operand: the
// parsed raw value converted to the matched Go parameter type.
func goportArgExpr(arg goportArgPlan, rawName string) (string, error) {
	parsedType := ""
	sourceType := ""
	sourceExpr := rawName
	switch arg.InputType {
	case "OP_DEC32":
		parsedType = "uint32"
		sourceType = parsedType
	case "OP_DEC64":
		parsedType = "uint64"
		sourceType = parsedType
	case "OP_DEC128":
		parsedType = "[16]byte"
		sourceType = parsedType
	case "OP_INT8":
		parsedType = "int64"
		sourceType = "int8"
		sourceExpr = fmt.Sprintf("int8(%s)", rawName)
	case "OP_INT16":
		parsedType = "int64"
		sourceType = "int16"
		sourceExpr = fmt.Sprintf("int16(%s)", rawName)
	case "OP_INT32":
		parsedType = "int64"
		sourceType = "int32"
		sourceExpr = fmt.Sprintf("int32(%s)", rawName)
	case "OP_INT64", "OP_LINT":
		parsedType = "int64"
		sourceType = parsedType
	case "OP_BID_UINT8":
		parsedType = "uint64"
		sourceType = "uint8"
		sourceExpr = fmt.Sprintf("uint8(%s)", rawName)
	case "OP_BID_UINT16":
		parsedType = "uint64"
		sourceType = "uint16"
		sourceExpr = fmt.Sprintf("uint16(%s)", rawName)
	case "OP_BID_UINT32":
		parsedType = "uint64"
		sourceType = "uint32"
		sourceExpr = fmt.Sprintf("uint32(%s)", rawName)
	case "OP_BID_UINT64":
		parsedType = "uint64"
		sourceType = parsedType
	default:
		return "", fmt.Errorf("unsupported readtest input type %q", arg.InputType)
	}
	if arg.GoType == "BID_UINT128" {
		if parsedType != "[16]byte" {
			return "", fmt.Errorf("operand %d: cannot convert %s to BID_UINT128", arg.OperandIndex, parsedType)
		}
		return fmt.Sprintf("goportReadtestToBidgo128(%s)", rawName), nil
	}
	if arg.GoType == sourceType {
		return sourceExpr, nil
	}
	switch arg.GoType {
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64":
		return fmt.Sprintf("%s(%s)", arg.GoType, sourceExpr), nil
	default:
		return "", fmt.Errorf("operand %d: unsupported Go parameter type %q", arg.OperandIndex, arg.GoType)
	}
}

// goportSecondaryEmit renders the readtestSecondaryOutput literal for a plan
// whose Go port returns the Intel pointer output as its second result.
func goportSecondaryEmit(plan goportCallPlan) (string, error) {
	switch plan.SecondaryKind {
	case "int":
		return fmt.Sprintf("readtestSecondaryOutput{Kind: readtestSecondaryInt, OperandIndex: %d, Int: int64(second)}", plan.SecondaryOperandIndex), nil
	case "dec32":
		return fmt.Sprintf("readtestSecondaryOutput{Kind: readtestSecondaryDec32, OperandIndex: %d, Bits32: second}", plan.SecondaryOperandIndex), nil
	case "dec64":
		return fmt.Sprintf("readtestSecondaryOutput{Kind: readtestSecondaryDec64, OperandIndex: %d, Bits64: second}", plan.SecondaryOperandIndex), nil
	case "dec128":
		return fmt.Sprintf("readtestSecondaryOutput{Kind: readtestSecondaryDec128, OperandIndex: %d, Bits128: goportReadtestFromBidgo128(second)}", plan.SecondaryOperandIndex), nil
	default:
		return "", fmt.Errorf("unsupported secondary output kind %q", plan.SecondaryKind)
	}
}

// emitGoportStatusControlDispatchCase emits one section 5.7.4 status-control
// case. It mirrors the native C wrapper (readtest_codegen.go
// emitReadtestStatusControlCWrapper) line for line with bidgo in place of the
// Intel C entry point: the incoming status word becomes a local word masked
// with BID_FLAG_MASK, the port is called with the row's value operands (and the
// row's rounding mode where the C prototype takes rnd_mode), and the compared
// status is that local word afterwards — 0 for the operations whose C prototype
// has no *pfpsf parameter, matching the upstream harness which compares
// expected_status against an untouched status word there.
func emitGoportStatusControlDispatchCase(buf *bytes.Buffer, category, zero string, plan goportCallPlan, withSecondary bool) error {
	shape := plan.StatusControl
	if withSecondary {
		return fmt.Errorf("%s: status-control rows have no secondary output group", plan.Function)
	}
	if category != "unsigned" {
		return fmt.Errorf("%s: status-control rows belong in the unsigned dispatch group, got %q", plan.Function, category)
	}
	fmt.Fprintf(buf, "\tcase %q:\n", plan.Function)
	fmt.Fprintf(buf, "\t\tif len(operands) != %d {\n", shape.Operands)
	fmt.Fprintf(buf, "\t\t\treturn %s, \"\", fmt.Errorf(\"%s expects %d operands, got %%d\", len(operands))\n", zero, plan.Function, shape.Operands)
	buf.WriteString("\t\t}\n")
	for operand := 0; operand < shape.Operands; operand++ {
		fmt.Fprintf(buf, "\t\targ%dRaw, err := parseReadtestUint(operands[%d])\n", operand, operand)
		buf.WriteString("\t\tif err != nil {\n")
		fmt.Fprintf(buf, "\t\t\treturn %s, \"\", err\n", zero)
		buf.WriteString("\t\t}\n")
	}
	for operand := 0; operand < shape.Operands; operand++ {
		if operand == shape.InitialFlagsOperand || containsInt(shape.ValueOperands, operand) {
			continue
		}
		fmt.Fprintf(buf, "\t\t_ = arg%dRaw\n", operand)
	}

	callArgs := make([]string, 0, len(shape.ValueOperands)+2)
	for _, operand := range shape.ValueOperands {
		callArgs = append(callArgs, fmt.Sprintf("uint32(arg%dRaw)", operand))
	}
	if shape.PassRounding {
		callArgs = append(callArgs, "uint32(rounding)")
	}
	statusExpr := "0"
	if shape.HasFlagsPointer {
		fmt.Fprintf(buf, "\t\tflags := uint32(arg%dRaw) & bidgo.BID_FLAG_MASK\n", shape.InitialFlagsOperand)
		callArgs = append(callArgs, "&flags")
		statusExpr = "flags"
	}
	call := fmt.Sprintf("bidgo.%s(%s)", plan.GoName, strings.Join(callArgs, ", "))
	resultExpr := zero
	if shape.HasResult {
		fmt.Fprintf(buf, "\t\tresult := %s\n", call)
		resultExpr = "uint64(result)"
	} else {
		fmt.Fprintf(buf, "\t\t%s\n", call)
	}
	fmt.Fprintf(buf, "\t\treturn %s, formatReadtestStatus(%s), nil\n", resultExpr, statusExpr)
	return nil
}

func containsInt(values []int, want int) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func emitGoportDispatchCase(buf *bytes.Buffer, category, zero string, plan goportCallPlan, withSecondary bool) error {
	if plan.StatusControl != nil {
		return emitGoportStatusControlDispatchCase(buf, category, zero, plan, withSecondary)
	}
	hasSecondary := plan.SecondaryKind != "none"
	if hasSecondary && !withSecondary {
		return fmt.Errorf("%s: pointer output parameter is only supported in the decimal dispatch groups", plan.Function)
	}
	zeroReturn := zero
	if withSecondary {
		zeroReturn = zero + ", readtestNoSecondaryOutput()"
	}
	// Functions like bid*_inf take no Intel C parameters but still carry a
	// dummy operand in the readtest rows, so the count comes from the spec
	// input types rather than the consumed value parameters.
	operandCount := plan.OperandCount
	if operandCount == 0 {
		operandCount = len(plan.Args)
	}
	fmt.Fprintf(buf, "\tcase %q:\n", plan.Function)
	fmt.Fprintf(buf, "\t\tif len(operands) != %d {\n", operandCount)
	fmt.Fprintf(buf, "\t\t\treturn %s, \"\", fmt.Errorf(\"%s expects %d operands, got %%d\", len(operands))\n", zeroReturn, plan.Function, operandCount)
	buf.WriteString("\t\t}\n")

	callArgs := []string{}
	for _, arg := range plan.Args {
		rawName := fmt.Sprintf("arg%dRaw", arg.OperandIndex)
		if arg.Kind == readtestParamCStr {
			if arg.GoType == "string" {
				callArgs = append(callArgs, fmt.Sprintf("goportReadtestCStringTag(operands[%d])", arg.OperandIndex))
			}
			// Dropped C string operands carry no parse validation in the
			// native dispatch either (NULL is a valid C input).
			continue
		}
		parserName, err := goportOperandParser(arg.Kind, arg.InputType)
		if err != nil {
			return fmt.Errorf("%s: %w", plan.Function, err)
		}
		fmt.Fprintf(buf, "\t\t%s, err := %s(operands[%d])\n", rawName, parserName, arg.OperandIndex)
		buf.WriteString("\t\tif err != nil {\n")
		fmt.Fprintf(buf, "\t\t\treturn %s, \"\", err\n", zeroReturn)
		buf.WriteString("\t\t}\n")
		if arg.GoType == "" {
			fmt.Fprintf(buf, "\t\t_ = %s\n", rawName)
			continue
		}
		expr, err := goportArgExpr(arg, rawName)
		if err != nil {
			return fmt.Errorf("%s: %w", plan.Function, err)
		}
		callArgs = append(callArgs, expr)
	}
	if plan.HasRounding {
		callArgs = append(callArgs, "rounding")
	}
	if plan.FlagsKind == goportFlagsPointer {
		buf.WriteString("\t\tvar flags uint32\n")
		callArgs = append(callArgs, "&flags")
	}

	// Assignment shape: result is always the first Go result; the Intel
	// pointer output (if any) is the second result; trailing uint32 flags
	// (if wired) are received; other extra outputs are blanked.
	lhs := []string{"result"}
	resultCount := len(plan.Results)
	flagsFromResult := plan.FlagsKind == goportFlagsResult
	for i := 1; i < resultCount; i++ {
		switch {
		case hasSecondary && i == 1:
			lhs = append(lhs, "second")
		case flagsFromResult && i == resultCount-1:
			lhs = append(lhs, "flags")
		default:
			lhs = append(lhs, "_")
		}
	}
	fmt.Fprintf(buf, "\t\t%s := bidgo.%s(%s)\n", strings.Join(lhs, ", "), plan.GoName, strings.Join(callArgs, ", "))

	flagsExpr := "0"
	if plan.FlagsKind == goportFlagsPointer || flagsFromResult {
		flagsExpr = "flags"
	}
	resultExpr, err := goportResultExpr(category, plan.Results[0].Type)
	if err != nil {
		return fmt.Errorf("%s: %w", plan.Function, err)
	}
	if withSecondary {
		secExpr := "readtestNoSecondaryOutput()"
		if hasSecondary {
			secExpr, err = goportSecondaryEmit(plan)
			if err != nil {
				return fmt.Errorf("%s: %w", plan.Function, err)
			}
		}
		fmt.Fprintf(buf, "\t\treturn %s, %s, formatReadtestStatus(%s), nil\n", resultExpr, secExpr, flagsExpr)
	} else {
		fmt.Fprintf(buf, "\t\treturn %s, formatReadtestStatus(%s), nil\n", resultExpr, flagsExpr)
	}
	return nil
}

func goportResultExpr(category, resultType string) (string, error) {
	switch category {
	case "dec32":
		return "result", nil
	case "dec64":
		return "result", nil
	case "dec128", "bin128":
		return "goportReadtestFromBidgo128(result)", nil
	case "bin32":
		if resultType == "float32" {
			return "math.Float32bits(result)", nil
		}
		return "result", nil
	case "bin64":
		if resultType == "float64" {
			return "math.Float64bits(result)", nil
		}
		return "result", nil
	case "signed":
		if resultType == "int64" {
			return "result", nil
		}
		if resultType == "bool" {
			return "goportReadtestBoolToInt64(result)", nil
		}
		return "int64(result)", nil
	case "unsigned":
		if resultType == "uint64" {
			return "result", nil
		}
		return "uint64(result)", nil
	default:
		return "", fmt.Errorf("unsupported dispatch category %q", category)
	}
}

func emitGoportFromString(buf *bytes.Buffer, entry goportStringEntry) {
	switch entry.Width {
	case "32":
		fmt.Fprintf(buf, "func goportReadtestBID32FromString(input string, rounding int) (uint32, string) {\n")
		fmt.Fprintf(buf, "\tvalue, flags := bidgo.%s(input, rounding)\n", entry.GoName)
		buf.WriteString("\treturn value, formatReadtestStatus(flags)\n}\n\n")
	case "64":
		fmt.Fprintf(buf, "func goportReadtestBID64FromString(input string, rounding int) (uint64, string) {\n")
		fmt.Fprintf(buf, "\tvalue, flags := bidgo.%s(input, rounding)\n", entry.GoName)
		buf.WriteString("\treturn value, formatReadtestStatus(flags)\n}\n\n")
	case "128":
		fmt.Fprintf(buf, "func goportReadtestBID128FromString(input string, rounding int) ([16]byte, string) {\n")
		fmt.Fprintf(buf, "\tvalue, flags := bidgo.%s(input, rounding)\n", entry.GoName)
		buf.WriteString("\treturn goportReadtestFromBidgo128(value), formatReadtestStatus(flags)\n}\n\n")
	}
}

// The to_string entries return the raw port output; readtest.c check_results
// compares that string by round-tripping it through bid*_from_string, so no
// display normalization is applied here.
func emitGoportToString(buf *bytes.Buffer, entry goportStringEntry) {
	switch entry.Width {
	case "32":
		fmt.Fprintf(buf, "func goportReadtestBID32ToString(a uint32) (string, string) {\n")
		fmt.Fprintf(buf, "\treturn bidgo.%s(a), formatReadtestStatus(0)\n}\n\n", entry.GoName)
	case "64":
		fmt.Fprintf(buf, "func goportReadtestBID64ToString(a uint64) (string, string) {\n")
		fmt.Fprintf(buf, "\treturn bidgo.%s(a), formatReadtestStatus(0)\n}\n\n", entry.GoName)
	case "128":
		fmt.Fprintf(buf, "func goportReadtestBID128ToString(a [16]byte) (string, string) {\n")
		fmt.Fprintf(buf, "\treturn bidgo.%s(goportReadtestToBidgo128(a)), formatReadtestStatus(0)\n}\n\n", entry.GoName)
	}
}

// readtestGoportHelpers holds only the bidgo-specific glue; the
// backend-independent parse/normalize/compare helpers live in the shared
// generated file emitted by the readtest dispatch generator
// (generated_readtest_shared.go).
const readtestGoportHelpers = `
// goportReadtestToBidgo128 / goportReadtestFromBidgo128 convert between the
// little-endian [16]byte 128-bit image and the port operand type through the
// explicit (hi, lo) word view, so the conversion is byte-order identical on
// every host endianness (a native-endian pointer reinterpretation would
// byte-swap the words on big-endian platforms).
func goportReadtestToBidgo128(raw [16]byte) bidgo.BID_UINT128 {
	return bidgo.Bid128FromWords(binary.LittleEndian.Uint64(raw[8:16]), binary.LittleEndian.Uint64(raw[0:8]))
}

func goportReadtestFromBidgo128(value bidgo.BID_UINT128) [16]byte {
	hi, lo := bidgo.Bid128Words(value)
	var raw [16]byte
	binary.LittleEndian.PutUint64(raw[0:8], lo)
	binary.LittleEndian.PutUint64(raw[8:16], hi)
	return raw
}

func goportReadtestBoolToInt64(value bool) int64 {
	if value {
		return 1
	}
	return 0
}

func goportReadtestCStringTag(input string) string {
	if strings.EqualFold(strings.TrimSpace(input), "NULL") {
		return ""
	}
	return input
}
`

func generateReadtestGoportInventory(reads []ReadTestSpec, symbols map[string]symbolSpec, sigs map[string]bidgoFuncSig) ([]byte, error) {
	inventory := goportInventory{
		Version: "1.0",
		Source:  "generated/testspec/spec_index.json + generated/json/intel_dfp_symbols.json + bid754-go/internal/bidgo",
	}
	for _, read := range reads {
		row := goportInventoryRow{Function: read.Function, Status: "dispatched"}
		if isReadtestStatusControlFunction(read.Function) {
			plan, err := resolveGoportStatusControlDispatch(read, sigs)
			if err != nil {
				return nil, err
			}
			row.GoFunction = plan.GoName
			inventory.Dispatched++
			inventory.Functions = append(inventory.Functions, row)
			continue
		}
		switch read.Kind {
		case "from_string":
			width := strings.TrimSuffix(strings.TrimPrefix(read.Function, "bid"), "_from_string")
			entry, err := resolveGoportFromString(width, sigs)
			if err != nil {
				return nil, err
			}
			row.GoFunction = entry.GoName
		case "to_string":
			width := strings.TrimSuffix(strings.TrimPrefix(read.Function, "bid"), "_to_string")
			entry, err := resolveGoportToString(width, sigs)
			if err != nil {
				return nil, err
			}
			row.GoFunction = entry.GoName
		default:
			symbol, ok := symbols[read.Function]
			if !ok {
				return nil, fmt.Errorf("readtest goport inventory generation: symbol %q not found", read.Function)
			}
			kinds, err := classifyReadtestParameters(symbol.Parameters)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", read.Function, err)
			}
			plan, err := resolveGoportDispatch(read, kinds, sigs)
			if err != nil {
				return nil, err
			}
			row.GoFunction = plan.GoName
			if drops, ok := goportReadtestOperandDropExceptions[read.Function]; ok {
				indices := make([]int, 0, len(drops))
				for idx := range drops {
					indices = append(indices, idx)
				}
				sort.Ints(indices)
				reasons := make([]string, 0, len(indices))
				for _, idx := range indices {
					reasons = append(reasons, fmt.Sprintf("operand %d dropped: %s", idx, drops[idx]))
				}
				row.Reason = strings.Join(reasons, "; ")
			}
		}
		inventory.Dispatched++
		inventory.Functions = append(inventory.Functions, row)
	}
	data, err := json.MarshalIndent(inventory, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal goport readtest dispatch inventory: %w", err)
	}
	return append(data, '\n'), nil
}

type readtestGoportCaseCounts struct {
	readtestGeneratedCaseCounts
	Executed         int
	CDivergeExecuted int
}

func countReadtestGoportCases(spec SharedSpec) readtestGoportCaseCounts {
	counts := readtestGoportCaseCounts{readtestGeneratedCaseCounts: countReadtestGeneratedCases(spec)}
	// Every generated readtest row runs against the Go mechanical port,
	// status_control rows included; the goport gate has no exclusion.
	counts.Executed = counts.Total
	for _, tc := range spec.ReadCases {
		if tc.NativeCompareSkipReason != "" {
			counts.CDivergeExecuted++
		}
	}
	return counts
}

func generateReadtestGoportCases(spec SharedSpec) []byte {
	counts := countReadtestGoportCases(spec)
	replacer := strings.NewReplacer(
		"@@TOTAL@@", fmt.Sprint(counts.Total),
		"@@DECIMAL32@@", fmt.Sprint(counts.Decimal32),
		"@@DECIMAL64@@", fmt.Sprint(counts.Decimal64),
		"@@DECIMAL128@@", fmt.Sprint(counts.Decimal128),
		"@@FROM_STRING@@", fmt.Sprint(counts.FromString),
		"@@TO_STRING@@", fmt.Sprint(counts.ToString),
		"@@UNARY_OP@@", fmt.Sprint(counts.UnaryOp),
		"@@BINARY_OP@@", fmt.Sprint(counts.BinaryOp),
		"@@TERNARY_OP@@", fmt.Sprint(counts.TernaryOp),
		"@@STATUS_CONTROL@@", fmt.Sprint(counts.StatusControl),
		"@@EXECUTED@@", fmt.Sprint(counts.Executed),
		"@@CDIVERGE_EXECUTED@@", fmt.Sprint(counts.CDivergeExecuted),
		"@@FUNCTION_COUNTS@@", stringIntMapLiteral(counts.Functions),
		"@@GROUP_COUNTS@@", stringIntMapLiteral(counts.Groups),
		"@@COMPARE_GROUP_COUNTS@@", stringIntMapLiteral(counts.CompareGroups),
	)
	return []byte(replacer.Replace(readtestGoportCasesTemplate))
}

var readtestGoportCasesTemplate = genmarker.Line("testgen") + `

package bid754

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/sky1core/bid754/bid754-go/internal/testspec"
)

type goportReadCaseCounts struct {
	Total         int
	Decimal32     int
	Decimal64     int
	Decimal128    int
	FromString    int
	ToString      int
	UnaryOp       int
	BinaryOp      int
	TernaryOp     int
	StatusControl int
	Functions     map[string]int
	Groups        map[string]int
	CompareGroups map[string]int
}

var expectedGoportReadCaseCounts = goportReadCaseCounts{
	Total:         @@TOTAL@@,
	Decimal32:     @@DECIMAL32@@,
	Decimal64:     @@DECIMAL64@@,
	Decimal128:    @@DECIMAL128@@,
	FromString:    @@FROM_STRING@@,
	ToString:      @@TO_STRING@@,
	UnaryOp:       @@UNARY_OP@@,
	BinaryOp:      @@BINARY_OP@@,
	TernaryOp:     @@TERNARY_OP@@,
	StatusControl: @@STATUS_CONTROL@@,
	Functions: map[string]int{
@@FUNCTION_COUNTS@@
	},
	Groups: map[string]int{
@@GROUP_COUNTS@@
	},
	CompareGroups: map[string]int{
@@COMPARE_GROUP_COUNTS@@
	},
}

// The goport gate runs every generated readtest case directly against the Go
// mechanical port, with no exclusion: the IEEE 754-2019 section 5.7.4
// status_control rows drive the ported bid_*Flags / bid_*DecimalRoundingDirection
// operations with the row's status word and rounding mode passed explicitly.
// Rows carrying a native-compare skip reason (cdiverge) pin intended IEEE
// behavior, so they must execute and pass here too.
const (
	expectedGoportExecutedReadCases         = @@EXECUTED@@
	expectedGoportCDivergeExecutedReadCases = @@CDIVERGE_EXECUTED@@
)

// goportReadtestStringBackend routes the readtest.c check_results string
// round-trips through the Go mechanical port bid*_from_string, mirroring the
// upstream harness which uses the library's own conversion.
var goportReadtestStringBackend = readtestStringBackend{
	FromString32:  goportReadtestBID32FromString,
	FromString64:  goportReadtestBID64FromString,
	FromString128: goportReadtestBID128FromString,
}

// goportReadtestOperationBackend routes the CMP_RELATIVEERR comparator's
// bid*_quantize / bid*_quiet_less calls through the Go mechanical-port
// dispatch, mirroring the upstream check32/64/128_rel BIDECIMAL_CALL2 calls.
var goportReadtestOperationBackend = readtestOperationBackend{
	Dec32:  goportReadtestGeneratedBID32,
	Dec64:  goportReadtestGeneratedBID64,
	Dec128: goportReadtestGeneratedBID128,
	Signed: goportReadtestGeneratedSigned,
}

func TestGeneratedReadCasesGoPort(t *testing.T) {
	if testing.Short() {
		t.Skip("goport readtest cases run in non-short mode; use make test-portable-readtest")
	}
	spec := goportLoadGeneratedReadSpec(t)
	if len(spec.ReadCases) == 0 {
		t.Fatal("expected generated read cases")
	}
	goportAssertReadCaseCounts(t, goportCountReadCases(spec.ReadCases), expectedGoportReadCaseCounts)

	executed := 0
	cdivergeExecuted := 0
	for _, tc := range spec.ReadCases {
		executed++
		if tc.NativeCompareSkipReason != "" {
			cdivergeExecuted++
		}
		tc := tc
		t.Run(tc.ID, func(t *testing.T) {
			switch tc.Kind {
			case "from_string":
				got, status, err := goportReadCaseBits(tc)
				if err != nil {
					t.Fatalf("goportReadCaseBits(%q): %v", tc.Operands[0], err)
				}
				if strings.HasPrefix(strings.TrimSpace(tc.Expected), "[") {
					width, err := readtestFormatBitWidth(tc.Format)
					if err != nil {
						t.Fatalf("readtestFormatBitWidth(%q): %v", tc.Format, err)
					}
					if normalizeReadtestBits(got, width) != normalizeReadtestBits(tc.Expected, width) {
						t.Fatalf("goport read case %s line %d: expected bits %q, got %q", tc.ID, tc.Line, normalizeReadtestBits(tc.Expected, width), normalizeReadtestBits(got, width))
					}
				} else {
					equal, err := readtestDecimalRowEqual(tc.Format, tc.Expected, got, tc.Rounding, goportReadtestStringBackend)
					if err != nil {
						t.Fatalf("readtestDecimalRowEqual(%q): %v", tc.Function, err)
					}
					if !equal {
						t.Fatalf("goport read case %s line %d: expected %q, got bits %q", tc.ID, tc.Line, tc.Expected, got)
					}
				}
				if normalizeReadtestStatus(status) != normalizeReadtestStatus(tc.Status) {
					t.Fatalf("goport read case %s line %d: expected status %q, got %q", tc.ID, tc.Line, normalizeReadtestStatus(tc.Status), normalizeReadtestStatus(status))
				}
			case "to_string":
				got, status, err := goportReadCaseString(tc)
				if err != nil {
					t.Fatalf("goportReadCaseString(%q): %v", tc.Operands[0], err)
				}
				equal, roundTripStatus, err := readtestToStringRowEqual(tc.Format, tc.Expected, got, tc.Rounding, goportReadtestStringBackend)
				if err != nil {
					t.Fatalf("readtestToStringRowEqual(%q): %v", tc.Function, err)
				}
				if !equal {
					t.Fatalf("goport read case %s line %d: expected %q, got %q", tc.ID, tc.Line, tc.Expected, got)
				}
				combined, err := readtestCombineStatus(status, roundTripStatus)
				if err != nil {
					t.Fatalf("readtestCombineStatus(%q, %q): %v", status, roundTripStatus, err)
				}
				if normalizeReadtestStatus(combined) != normalizeReadtestStatus(tc.Status) {
					t.Fatalf("goport read case %s line %d: expected status %q, got %q", tc.ID, tc.Line, normalizeReadtestStatus(tc.Status), normalizeReadtestStatus(combined))
				}
			case "binary_op", "unary_op", "ternary_op", "status_control":
				if isReadtestScalarOutput(tc.OutputType) {
					got, status, err := goportReadCaseScalarString(tc)
					if err != nil {
						t.Fatalf("goportReadCaseScalarString(%q): %v", tc.Function, err)
					}
					if !compareReadtestScalarOutput(tc.OutputType, tc.Expected, got) {
						t.Fatalf("goport read case %s line %d: expected %q, got %q", tc.ID, tc.Line, tc.Expected, got)
					}
					if normalizeReadtestStatus(status) != normalizeReadtestStatus(tc.Status) {
						t.Fatalf("goport read case %s line %d: expected status %q, got %q", tc.ID, tc.Line, normalizeReadtestStatus(tc.Status), normalizeReadtestStatus(status))
					}
					return
				}
				got, sec, status, err := goportReadCaseOperationBits(tc)
				if err != nil {
					t.Fatalf("goportReadCaseOperationBits(%q): %v", tc.Function, err)
				}
				switch {
				case tc.CompareGroup == "CMP_RELATIVEERR":
					// readtest.c CMP_RELATIVEERR rows compare check*_rel plus the
					// trans_flags_mask-masked status only (readtest.c:1477/1486/1495);
					// no secondary output and no exact status comparison apply.
					equal, err := readtestRelativeErrRowEqual(tc.Format, tc.Expected, got, tc.Rounding, tc.UlpAdd, goportReadtestStringBackend, goportReadtestOperationBackend)
					if err != nil {
						t.Fatalf("readtestRelativeErrRowEqual(%q): %v", tc.Function, err)
					}
					if !equal {
						t.Fatalf("goport read case %s line %d: expected relative-error match %q (ulp_add %v), got bits %q", tc.ID, tc.Line, tc.Expected, tc.UlpAdd, got)
					}
					statusEqual, err := readtestRelativeErrStatusEqual(tc.Status, status)
					if err != nil {
						t.Fatalf("readtestRelativeErrStatusEqual(%q, %q): %v", tc.Status, status, err)
					}
					if !statusEqual {
						t.Fatalf("goport read case %s line %d: expected masked status %q, got %q", tc.ID, tc.Line, normalizeReadtestStatus(tc.Status), normalizeReadtestStatus(status))
					}
					return
				case tc.CompareGroup == "CMP_EQUALSTATUS":
					// readtest.c check_results does not compare the frexp/modf
					// secondary output in its CMP_EQUALSTATUS branches, so the
					// secondary check applies only to the value branches below.
					equal, err := readtestQuietEqual(tc.Format, tc.Expected, got, tc.Rounding, goportReadtestStringBackend, goportReadtestGeneratedSigned)
					if err != nil {
						t.Fatalf("readtestQuietEqual(%q): %v", tc.Function, err)
					}
					if !equal {
						t.Fatalf("goport read case %s line %d: expected quiet-equal %q, got %q", tc.ID, tc.Line, tc.Expected, got)
					}
				case strings.HasPrefix(strings.TrimSpace(tc.Expected), "["):
					width, err := readtestFormatBitWidth(tc.Format)
					if err != nil {
						t.Fatalf("readtestFormatBitWidth(%q): %v", tc.Format, err)
					}
					if normalizeReadtestBits(got, width) != normalizeReadtestBits(tc.Expected, width) {
						t.Fatalf("goport read case %s line %d: expected bits %q, got %q", tc.ID, tc.Line, normalizeReadtestBits(tc.Expected, width), normalizeReadtestBits(got, width))
					}
					secondaryEqual, err := readtestSecondaryOutputEqual(sec, tc.Operands)
					if err != nil {
						t.Fatalf("readtestSecondaryOutputEqual(%q): %v", tc.Function, err)
					}
					if !secondaryEqual {
						t.Fatalf("goport read case %s line %d: secondary output %+v does not match operand %q", tc.ID, tc.Line, sec, tc.Operands[sec.OperandIndex])
					}
				default:
					equal, err := readtestDecimalRowEqual(tc.Format, tc.Expected, got, tc.Rounding, goportReadtestStringBackend)
					if err != nil {
						t.Fatalf("readtestDecimalRowEqual(%q): %v", tc.Function, err)
					}
					if !equal {
						t.Fatalf("goport read case %s line %d: expected %q, got bits %q", tc.ID, tc.Line, tc.Expected, got)
					}
					secondaryEqual, err := readtestSecondaryOutputEqual(sec, tc.Operands)
					if err != nil {
						t.Fatalf("readtestSecondaryOutputEqual(%q): %v", tc.Function, err)
					}
					if !secondaryEqual {
						t.Fatalf("goport read case %s line %d: secondary output %+v does not match operand %q", tc.ID, tc.Line, sec, tc.Operands[sec.OperandIndex])
					}
				}
				if normalizeReadtestStatus(status) != normalizeReadtestStatus(tc.Status) {
					t.Fatalf("goport read case %s line %d: expected status %q, got %q", tc.ID, tc.Line, normalizeReadtestStatus(tc.Status), normalizeReadtestStatus(status))
				}
			default:
				t.Fatalf("unsupported goport read kind %q", tc.Kind)
			}
		})
	}
	if executed != expectedGoportExecutedReadCases {
		t.Fatalf("goport executed read case count = %d, want %d", executed, expectedGoportExecutedReadCases)
	}
	if cdivergeExecuted != expectedGoportCDivergeExecutedReadCases {
		t.Fatalf("goport executed cdiverge read case count = %d, want %d", cdivergeExecuted, expectedGoportCDivergeExecutedReadCases)
	}
}

func goportCountReadCases(cases []testspec.GeneratedReadCase) goportReadCaseCounts {
	counts := goportReadCaseCounts{
		Functions:     map[string]int{},
		Groups:        map[string]int{},
		CompareGroups: map[string]int{},
	}
	for _, tc := range cases {
		counts.Total++
		counts.Functions[tc.Function]++
		counts.Groups[tc.Group]++
		counts.CompareGroups[tc.CompareGroup]++
		switch tc.Format {
		case "decimal32":
			counts.Decimal32++
		case "decimal64":
			counts.Decimal64++
		case "decimal128":
			counts.Decimal128++
		}
		switch tc.Kind {
		case "from_string":
			counts.FromString++
		case "to_string":
			counts.ToString++
		case "unary_op":
			counts.UnaryOp++
		case "binary_op":
			counts.BinaryOp++
		case "ternary_op":
			counts.TernaryOp++
		case "status_control":
			counts.StatusControl++
		}
	}
	return counts
}

func goportAssertReadCaseCounts(t *testing.T, got, want goportReadCaseCounts) {
	t.Helper()
	for _, item := range []struct {
		label string
		got   int
		want  int
	}{
		{"total", got.Total, want.Total},
		{"decimal32", got.Decimal32, want.Decimal32},
		{"decimal64", got.Decimal64, want.Decimal64},
		{"decimal128", got.Decimal128, want.Decimal128},
		{"from_string", got.FromString, want.FromString},
		{"to_string", got.ToString, want.ToString},
		{"unary_op", got.UnaryOp, want.UnaryOp},
		{"binary_op", got.BinaryOp, want.BinaryOp},
		{"ternary_op", got.TernaryOp, want.TernaryOp},
		{"status_control", got.StatusControl, want.StatusControl},
	} {
		if item.got != item.want {
			t.Fatalf("goport read case %s count = %d, want %d", item.label, item.got, item.want)
		}
	}
	goportAssertReadStringCounts(t, "function", got.Functions, want.Functions)
	goportAssertReadStringCounts(t, "group", got.Groups, want.Groups)
	goportAssertReadStringCounts(t, "compare_group", got.CompareGroups, want.CompareGroups)
}

func goportAssertReadStringCounts(t *testing.T, label string, got, want map[string]int) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("goport read %s bucket count = %d, want %d", label, len(got), len(want))
	}
	for key, wantValue := range want {
		if got[key] != wantValue {
			t.Fatalf("goport read %s count[%q] = %d, want %d", label, key, got[key], wantValue)
		}
	}
}

func goportLoadGeneratedReadSpec(t *testing.T) testspec.SharedSpec {
	t.Helper()
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("resolve generated readtest goport file path")
	}
	spec, err := testspec.LoadGenerated(filepath.Join(filepath.Dir(currentFile), "..", "devtools", "generated", "testspec", "spec_index.json"))
	if err != nil {
		t.Fatalf("load shared spec: %v", err)
	}
	return spec
}

func goportReadCaseBits(tc testspec.GeneratedReadCase) (string, string, error) {
	if len(tc.Operands) != 1 {
		return "", "", fmt.Errorf("%s expects 1 operand, got %d", tc.Kind, len(tc.Operands))
	}
	switch tc.Format {
	case "decimal32":
		raw, status := goportReadtestBID32FromString(tc.Operands[0], tc.Rounding)
		return fmt.Sprintf("[%08x]", raw), status, nil
	case "decimal64":
		raw, status := goportReadtestBID64FromString(tc.Operands[0], tc.Rounding)
		return fmt.Sprintf("[%016x]", raw), status, nil
	case "decimal128":
		raw, status := goportReadtestBID128FromString(tc.Operands[0], tc.Rounding)
		return formatReadtestBits128(raw), status, nil
	default:
		return "", "", fmt.Errorf("unsupported read format %q", tc.Format)
	}
}

func goportReadCaseString(tc testspec.GeneratedReadCase) (string, string, error) {
	if len(tc.Operands) != 1 {
		return "", "", fmt.Errorf("%s expects 1 operand, got %d", tc.Kind, len(tc.Operands))
	}
	switch tc.Format {
	case "decimal32":
		raw, err := parseReadtestHex(tc.Operands[0], 32)
		if err != nil {
			return "", "", err
		}
		value, status := goportReadtestBID32ToString(uint32(raw))
		return value, status, nil
	case "decimal64":
		raw, err := parseReadtestHex(tc.Operands[0], 64)
		if err != nil {
			return "", "", err
		}
		value, status := goportReadtestBID64ToString(raw)
		return value, status, nil
	case "decimal128":
		raw, err := parseReadtestBits128(tc.Operands[0])
		if err != nil {
			return "", "", err
		}
		value, status := goportReadtestBID128ToString(raw)
		return value, status, nil
	default:
		return "", "", fmt.Errorf("unsupported read format %q", tc.Format)
	}
}

func goportReadCaseOperationBits(tc testspec.GeneratedReadCase) (string, readtestSecondaryOutput, string, error) {
	switch tc.Format {
	case "decimal32":
		raw, sec, status, err := goportReadtestGeneratedBID32(tc.Function, tc.Rounding, tc.Operands)
		if err != nil {
			return "", readtestNoSecondaryOutput(), "", err
		}
		return fmt.Sprintf("[%08x]", raw), sec, status, nil
	case "decimal64":
		raw, sec, status, err := goportReadtestGeneratedBID64(tc.Function, tc.Rounding, tc.Operands)
		if err != nil {
			return "", readtestNoSecondaryOutput(), "", err
		}
		return fmt.Sprintf("[%016x]", raw), sec, status, nil
	case "decimal128":
		raw, sec, status, err := goportReadtestGeneratedBID128(tc.Function, tc.Rounding, tc.Operands)
		if err != nil {
			return "", readtestNoSecondaryOutput(), "", err
		}
		return formatReadtestBits128(raw), sec, status, nil
	default:
		return "", readtestNoSecondaryOutput(), "", fmt.Errorf("unsupported read format %q", tc.Format)
	}
}

func goportReadCaseScalarString(tc testspec.GeneratedReadCase) (string, string, error) {
	switch tc.OutputType {
	case "OP_BIN32":
		value, status, err := goportReadtestGeneratedBinary32(tc.Function, tc.Rounding, tc.Operands)
		if err != nil {
			return "", "", err
		}
		return fmt.Sprintf("[%08x]", value), status, nil
	case "OP_BIN64":
		value, status, err := goportReadtestGeneratedBinary64(tc.Function, tc.Rounding, tc.Operands)
		if err != nil {
			return "", "", err
		}
		return fmt.Sprintf("[%016x]", value), status, nil
	case "OP_BIN128":
		value, status, err := goportReadtestGeneratedBinary128(tc.Function, tc.Rounding, tc.Operands)
		if err != nil {
			return "", "", err
		}
		return formatReadtestBits128(value), status, nil
	case "OP_BID_UINT8", "OP_BID_UINT16", "OP_BID_UINT32", "OP_BID_UINT64":
		value, status, err := goportReadtestGeneratedUnsigned(tc.Function, tc.Rounding, tc.Operands)
		if err != nil {
			return "", "", err
		}
		return strconv.FormatUint(value, 10), status, nil
	case "OP_INT8", "OP_INT16", "OP_INT32", "OP_INT64", "OP_LINT":
		value, status, err := goportReadtestGeneratedSigned(tc.Function, tc.Rounding, tc.Operands)
		if err != nil {
			return "", "", err
		}
		return strconv.FormatInt(value, 10), status, nil
	default:
		return "", "", fmt.Errorf("unsupported scalar output %q", tc.OutputType)
	}
}
`
