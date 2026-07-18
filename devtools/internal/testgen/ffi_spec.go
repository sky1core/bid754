package testgen

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const ffiRoundingModeCount = 5

const ffiProbeRoundingDiscriminant = "rounding-discriminant"

type ffiMixedDecimalShape struct {
	format       string
	operation    string
	resultBits   int
	operandBits  [3]int
	operandCount int
}

func (s ffiMixedDecimalShape) operandWidths() []int {
	return append([]int(nil), s.operandBits[:s.operandCount]...)
}

// ffiMixedDecimalShapeFor is the closed-world registry for Intel extension
// entrypoints whose decimal operand widths differ from the result width. The
// D/Q letters describe x, y, and z in order; keeping the widths explicit here
// prevents a result-width inference from silently swapping or truncating an
// operand when the extracted C declaration changes.
func ffiMixedDecimalShapeFor(function string) (ffiMixedDecimalShape, bool) {
	const (
		d = 64
		q = 128
	)
	switch function {
	case "bid64ddq_fma":
		return ffiMixedDecimalShape{format: "decimal64", operation: "fma", resultBits: d, operandBits: [3]int{d, d, q}, operandCount: 3}, true
	case "bid64dqd_fma":
		return ffiMixedDecimalShape{format: "decimal64", operation: "fma", resultBits: d, operandBits: [3]int{d, q, d}, operandCount: 3}, true
	case "bid64dqq_fma":
		return ffiMixedDecimalShape{format: "decimal64", operation: "fma", resultBits: d, operandBits: [3]int{d, q, q}, operandCount: 3}, true
	case "bid64qdd_fma":
		return ffiMixedDecimalShape{format: "decimal64", operation: "fma", resultBits: d, operandBits: [3]int{q, d, d}, operandCount: 3}, true
	case "bid64qdq_fma":
		return ffiMixedDecimalShape{format: "decimal64", operation: "fma", resultBits: d, operandBits: [3]int{q, d, q}, operandCount: 3}, true
	case "bid64qqd_fma":
		return ffiMixedDecimalShape{format: "decimal64", operation: "fma", resultBits: d, operandBits: [3]int{q, q, d}, operandCount: 3}, true
	case "bid64qqq_fma":
		return ffiMixedDecimalShape{format: "decimal64", operation: "fma", resultBits: d, operandBits: [3]int{q, q, q}, operandCount: 3}, true
	case "bid128ddd_fma":
		return ffiMixedDecimalShape{format: "decimal128", operation: "fma", resultBits: q, operandBits: [3]int{d, d, d}, operandCount: 3}, true
	case "bid128ddq_fma":
		return ffiMixedDecimalShape{format: "decimal128", operation: "fma", resultBits: q, operandBits: [3]int{d, d, q}, operandCount: 3}, true
	case "bid128dqd_fma":
		return ffiMixedDecimalShape{format: "decimal128", operation: "fma", resultBits: q, operandBits: [3]int{d, q, d}, operandCount: 3}, true
	case "bid128dqq_fma":
		return ffiMixedDecimalShape{format: "decimal128", operation: "fma", resultBits: q, operandBits: [3]int{d, q, q}, operandCount: 3}, true
	case "bid128qdd_fma":
		return ffiMixedDecimalShape{format: "decimal128", operation: "fma", resultBits: q, operandBits: [3]int{q, d, d}, operandCount: 3}, true
	case "bid128qdq_fma":
		return ffiMixedDecimalShape{format: "decimal128", operation: "fma", resultBits: q, operandBits: [3]int{q, d, q}, operandCount: 3}, true
	case "bid128qqd_fma":
		return ffiMixedDecimalShape{format: "decimal128", operation: "fma", resultBits: q, operandBits: [3]int{q, q, d}, operandCount: 3}, true
	case "bid64q_sqrt":
		return ffiMixedDecimalShape{format: "decimal64", operation: "sqrt", resultBits: d, operandBits: [3]int{q}, operandCount: 1}, true
	case "bid128d_sqrt":
		return ffiMixedDecimalShape{format: "decimal128", operation: "sqrt", resultBits: q, operandBits: [3]int{d}, operandCount: 1}, true
	default:
		return ffiMixedDecimalShape{}, false
	}
}

func buildFFICases(repoRoot string, spec FFITestSpec) ([]GeneratedFFICase, error) {
	symbols, err := loadSymbolFile(filepath.Join(repoRoot, spec.Symbols))
	if err != nil {
		return nil, err
	}

	index := make(map[string]symbolSpec, len(symbols.Symbols))
	for _, symbol := range symbols.Symbols {
		index[symbol.Name] = symbol
	}

	functions, err := expandFFIFunctions(spec, symbols.Symbols)
	if err != nil {
		return nil, err
	}

	out := make([]GeneratedFFICase, 0, len(functions)*spec.CasesPerFunction)
	witnessByFunction := ffiMutationWitnessIndex()
	for _, function := range functions {
		symbol, ok := index[function]
		if !ok {
			return nil, fmt.Errorf("ffi suite %q: symbol %q not found in %q", spec.Name, function, spec.Symbols)
		}

		format, operation, bits, arity, err := verifyFFISignature(function, symbol)
		if err != nil {
			return nil, fmt.Errorf("ffi suite %q: %w", spec.Name, err)
		}

		hasRoundingParam := ffiSymbolHasRoundingParam(symbol)
		mixedShape, isMixedDecimal := ffiMixedDecimalShapeFor(function)
		var generator *deterministicFFIGenerator
		var mixedGenerator *deterministicFFIMixedDecimalGenerator
		if isMixedDecimal {
			mixedGenerator = newDeterministicFFIMixedDecimalGenerator(spec.Seed, function, mixedShape)
		} else {
			generator = newDeterministicFFIGenerator(spec.Seed, function, bits)
		}
		for i := 0; i < spec.CasesPerFunction; i++ {
			var operands []string
			if isMixedDecimal {
				operands = mixedGenerator.nextOperands(i)
			} else {
				operands = generator.nextOperandsForOperation(i, operation, arity)
			}
			out = append(out, GeneratedFFICase{
				Suite:       spec.Name,
				ID:          fmt.Sprintf("%s_%s_%03d", spec.Name, function, i+1),
				Format:      format,
				ResultBits:  mixedShape.resultBits,
				OperandBits: mixedShape.operandWidths(),
				Operation:   operation,
				Function:    function,
				LinkName:    symbol.LinkName,
				Declaration: symbol.Declaration,
				Source:      filepath.ToSlash(spec.Symbols),
				Rounding:    ffiCaseRoundingMode(i, hasRoundingParam),
				Operands:    operands,
			})
		}
		caseIndex := spec.CasesPerFunction
		if isMixedDecimal {
			if !hasRoundingParam {
				return nil, fmt.Errorf("ffi suite %q: mixed decimal operation %q has no rounding parameter", spec.Name, function)
			}
			probes, err := ffiMixedDecimalRoundingProbeOperands(mixedShape)
			if err != nil {
				return nil, fmt.Errorf("ffi suite %q: build rounding probes for %q: %w", spec.Name, function, err)
			}
			for probeIndex, operands := range probes {
				probeGroup := fmt.Sprintf("%s/rounding/%d", function, probeIndex+1)
				for mode := 0; mode < ffiRoundingModeCount; mode++ {
					out = append(out, GeneratedFFICase{
						Suite:       spec.Name,
						ID:          fmt.Sprintf("%s_%s_%03d", spec.Name, function, caseIndex+1),
						Format:      format,
						ResultBits:  mixedShape.resultBits,
						OperandBits: mixedShape.operandWidths(),
						Operation:   operation,
						Function:    function,
						LinkName:    symbol.LinkName,
						Declaration: symbol.Declaration,
						Source:      filepath.ToSlash(spec.Symbols),
						Probe:       ffiProbeRoundingDiscriminant,
						ProbeGroup:  probeGroup,
						Rounding:    mode,
						Operands:    append([]string(nil), operands...),
					})
					caseIndex++
				}
			}
			if operation == "fma" {
				probe, ok := ffiMixedFMAFusednessProbeFor(function)
				if !ok {
					return nil, fmt.Errorf("ffi suite %q: mixed FMA %q has no closed fusedness probe", spec.Name, function)
				}
				if err := validateFFIFusednessProbe(probe, mixedShape); err != nil {
					return nil, fmt.Errorf("ffi suite %q: %w", spec.Name, err)
				}
				out = append(out, GeneratedFFICase{
					Suite:       spec.Name,
					ID:          fmt.Sprintf("%s_%s_%03d", spec.Name, function, caseIndex+1),
					Format:      format,
					ResultBits:  mixedShape.resultBits,
					OperandBits: mixedShape.operandWidths(),
					Operation:   operation,
					Function:    function,
					LinkName:    symbol.LinkName,
					Declaration: symbol.Declaration,
					Source:      filepath.ToSlash(spec.Symbols),
					Probe:       ffiProbeFusedness,
					ProbeGroup:  function + "/fusedness",
					Expected:    probe.expected.ffiString(),
					Forbidden:   probe.forbidden.ffiString(),
					Rounding:    probe.rounding,
					Operands:    probe.ffiOperands(),
				})
				caseIndex++
			}
		}

		edgeCaseCount := ffiTier1RoundingEdgeCaseCount(operation, bits)
		if edgeCaseCount > 0 {
			if !hasRoundingParam {
				return nil, fmt.Errorf("ffi suite %q: Tier 1 operation %q has no rounding parameter", spec.Name, function)
			}
			if spec.CasesPerFunction < edgeCaseCount {
				return nil, fmt.Errorf("ffi suite %q: Tier 1 operation %q baseline has %d cases, need at least %d edge cases", spec.Name, function, spec.CasesPerFunction, edgeCaseCount)
			}
			var edgeGenerator *deterministicFFIGenerator
			var edgeMixedGenerator *deterministicFFIMixedDecimalGenerator
			if isMixedDecimal {
				edgeMixedGenerator = newDeterministicFFIMixedDecimalGenerator(spec.Seed, function, mixedShape)
			} else {
				edgeGenerator = newDeterministicFFIGenerator(spec.Seed, function, bits)
			}
			for edgeIndex := 0; edgeIndex < edgeCaseCount; edgeIndex++ {
				var operands []string
				if isMixedDecimal {
					operands = edgeMixedGenerator.nextOperands(edgeIndex)
				} else {
					operands = edgeGenerator.nextOperandsForOperation(edgeIndex, operation, arity)
				}
				baselineMode := ffiCaseRoundingMode(edgeIndex, true)
				for mode := 0; mode < ffiRoundingModeCount; mode++ {
					if mode == baselineMode {
						continue
					}
					out = append(out, GeneratedFFICase{
						Suite:       spec.Name,
						ID:          fmt.Sprintf("%s_%s_%03d", spec.Name, function, caseIndex+1),
						Format:      format,
						ResultBits:  mixedShape.resultBits,
						OperandBits: mixedShape.operandWidths(),
						Operation:   operation,
						Function:    function,
						LinkName:    symbol.LinkName,
						Declaration: symbol.Declaration,
						Source:      filepath.ToSlash(spec.Symbols),
						Rounding:    mode,
						Operands:    append([]string(nil), operands...),
					})
					caseIndex++
				}
			}
		}

		// Mutation-audit witness rows: one pinned distinguishing input per
		// audited surviving mutant on this function's flags-exposing port
		// path (mutation_witness_corpus.go).
		for _, witness := range witnessByFunction[function] {
			if len(witness.Operands) != arity {
				return nil, fmt.Errorf("ffi suite %q: mutation witness %s has %d operands, function %q takes %d", spec.Name, witness.MutantID, len(witness.Operands), function, arity)
			}
			if witness.Rounding != 0 && !hasRoundingParam {
				return nil, fmt.Errorf("ffi suite %q: mutation witness %s pins rounding %d but %q has no rounding parameter", spec.Name, witness.MutantID, witness.Rounding, function)
			}
			out = append(out, GeneratedFFICase{
				Suite:       spec.Name,
				ID:          fmt.Sprintf("%s_%s_%03d", spec.Name, function, caseIndex+1),
				Format:      format,
				ResultBits:  mixedShape.resultBits,
				OperandBits: mixedShape.operandWidths(),
				Operation:   operation,
				Function:    function,
				LinkName:    symbol.LinkName,
				Declaration: symbol.Declaration,
				Source:      filepath.ToSlash(spec.Symbols),
				Rounding:    witness.Rounding,
				Operands:    append([]string(nil), witness.Operands...),
			})
			caseIndex++
		}

		// bid_factors32 exactness sweep: every row of the pinned Intel
		// factors table consumed with sign-sensitive weight
		// (mutation_witness_corpus.go).
		if function == "bid32_div" {
			sweep, err := bid32DivFactors32SweepCases()
			if err != nil {
				return nil, fmt.Errorf("ffi suite %q: %w", spec.Name, err)
			}
			for _, sc := range sweep {
				out = append(out, GeneratedFFICase{
					Suite:       spec.Name,
					ID:          fmt.Sprintf("%s_%s_%03d", spec.Name, function, caseIndex+1),
					Format:      format,
					Operation:   operation,
					Function:    function,
					LinkName:    symbol.LinkName,
					Declaration: symbol.Declaration,
					Source:      filepath.ToSlash(spec.Symbols),
					Rounding:    0,
					Operands:    []string{encodeBid32SweepOperand(sc.X), encodeBid32SweepOperand(sc.Y)},
				})
				caseIndex++
			}
		}
	}

	return out, nil
}

// verifyFFIMutationCorpusCoverage closes the witness/sweep world over the
// union of every manifest FFI suite: each mutation-audit witness row and the
// bid_factors32 sweep must have been emitted by some generated suite. It runs
// in the manifest build path (spec_build.go) so reduced fixture suites in unit
// tests do not have to carry the full corpus, while the production generation
// fails closed when a targeted function leaves the generated surface.
func verifyFFIMutationCorpusCoverage(cases []GeneratedFFICase) error {
	emitted := map[string]bool{}
	sweepRows := 0
	for _, tc := range cases {
		emitted[tc.Function] = true
		if tc.Function == "bid32_div" {
			sweepRows++
		}
	}
	if err := ffiVerifyWitnessConsumption(emitted); err != nil {
		return err
	}
	if sweep, err := bid32DivFactors32SweepCases(); err != nil {
		return err
	} else if sweepRows < len(sweep) {
		return fmt.Errorf("generated FFI suites carry %d bid32_div cases, fewer than the %d bid_factors32 sweep cases; the exactness sweep lost its home", sweepRows, len(sweep))
	}
	return nil
}

// ffiMixedDecimalRoundingProbeOperands returns inputs whose exact arithmetic
// result lies on a result-width rounding boundary. Every returned operand set
// is emitted unchanged under all five Intel rounding modes; the native runner
// separately requires the canonical Intel C result bits to split into at
// least two distinct values.
func ffiMixedDecimalRoundingProbeOperands(shape ffiMixedDecimalShape) ([][]string, error) {
	var probes [][]modeDiscOperand
	switch shape.operation {
	case "fma":
		triples, err := mixedModeTernaryDiscriminantOperands("FMA", shape.resultBits, shape.operandBits)
		if err != nil {
			return nil, err
		}
		probes = make([][]modeDiscOperand, len(triples))
		for i, triple := range triples {
			probes[i] = append([]modeDiscOperand(nil), triple[:]...)
		}
	case "sqrt":
		values, err := mixedModeUnaryDiscriminantOperands("Sqrt", shape.resultBits, shape.operandBits[0])
		if err != nil {
			return nil, err
		}
		probes = make([][]modeDiscOperand, len(values))
		for i, value := range values {
			probes[i] = []modeDiscOperand{value}
		}
	default:
		return nil, fmt.Errorf("unsupported mixed decimal operation %q", shape.operation)
	}
	if len(probes) == 0 {
		return nil, fmt.Errorf("mixed decimal operation %q has no rounding probe operands", shape.operation)
	}
	// The first FMA row is the result-width half-ULP tie and the first sqrt
	// row is sqrt(2). One probe group per function is sufficient to prove the
	// rounding argument is live; the native runner validates that Intel C
	// actually produces at least two result bit patterns instead of assuming
	// that every generally useful parity discriminant also discriminates every
	// Intel mixed-width extension entrypoint.
	probes = probes[:1]

	widths := shape.operandWidths()
	encoded := make([][]string, len(probes))
	for probeIndex, probe := range probes {
		if len(probe) != len(widths) {
			return nil, fmt.Errorf("probe %d operand count = %d, want %d", probeIndex, len(probe), len(widths))
		}
		encoded[probeIndex] = make([]string, len(probe))
		for operandIndex, operand := range probe {
			bits, err := ffiModeDiscOperandBits(widths[operandIndex], operand)
			if err != nil {
				return nil, fmt.Errorf("probe %d operand %d width %d: %w", probeIndex, operandIndex, widths[operandIndex], err)
			}
			encoded[probeIndex][operandIndex] = bits
		}
	}
	return encoded, nil
}

func ffiModeDiscOperandBits(width int, operand modeDiscOperand) (string, error) {
	switch width {
	case 32:
		bits, err := encodeModeDiscOperand32(operand)
		if err != nil {
			return "", err
		}
		return formatFFIBits(uint64(bits), width), nil
	case 64:
		bits, err := encodeModeDiscOperand64(operand)
		if err != nil {
			return "", err
		}
		return formatFFIBits(bits, width), nil
	case 128:
		raw, err := encodeModeDiscOperand128(operand)
		if err != nil {
			return "", err
		}
		return hex.EncodeToString(raw[:]), nil
	default:
		return "", fmt.Errorf("unsupported decimal width %d", width)
	}
}

func ffiTier1RoundingEdgeCaseCount(operation string, bits int) int {
	switch operation {
	case "add", "sub", "mul", "div", "quantize":
		return len(ffiBinaryEdgePairs(bits))
	case "scalbln":
		return len(ffiScaleBEdgeExponents)
	default:
		return 0
	}
}

func ffiSymbolHasRoundingParam(symbol symbolSpec) bool {
	for _, param := range symbol.Parameters {
		if strings.HasPrefix(param, "_IDEC_round ") {
			return true
		}
	}
	return false
}

func ffiCaseRoundingMode(index int, hasRoundingParam bool) int {
	if !hasRoundingParam {
		return 0
	}
	return index % ffiRoundingModeCount
}

func expandFFIFunctions(spec FFITestSpec, symbols []symbolSpec) ([]string, error) {
	seen := map[string]struct{}{}
	functions := make([]string, 0, len(spec.Functions))
	for _, function := range spec.Functions {
		if _, ok := seen[function]; ok {
			continue
		}
		seen[function] = struct{}{}
		functions = append(functions, function)
	}

	for _, pattern := range spec.FunctionPatterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("ffi suite %q: compile function pattern %q: %w", spec.Name, pattern, err)
		}
		var matched []string
		for _, symbol := range symbols {
			if !re.MatchString(symbol.Name) {
				continue
			}
			if _, ok := seen[symbol.Name]; ok {
				continue
			}
			matched = append(matched, symbol.Name)
		}
		if len(matched) == 0 {
			return nil, fmt.Errorf("ffi suite %q: function pattern %q matched no symbols", spec.Name, pattern)
		}
		sort.Strings(matched)
		for _, function := range matched {
			seen[function] = struct{}{}
			functions = append(functions, function)
		}
	}

	return functions, nil
}

func verifyFFISignature(function string, symbol symbolSpec) (string, string, int, int, error) {
	if shape, ok := ffiMixedDecimalShapeFor(function); ok {
		if err := verifyFFIMixedDecimalSignature(function, symbol, shape); err != nil {
			return "", "", 0, 0, err
		}
		return shape.format, shape.operation, shape.resultBits, shape.operandCount, nil
	}

	prefix, operation, ok := strings.Cut(function, "_")
	if !ok {
		return "", "", 0, 0, fmt.Errorf("ffi symbol %q: unsupported name", function)
	}

	var (
		format     string
		returnType string
		paramType  string
		bits       int
	)
	switch prefix {
	case "bid32":
		format = "decimal32"
		returnType = "BID_UINT32"
		paramType = "BID_UINT32"
		bits = 32
	case "bid64":
		format = "decimal64"
		returnType = "BID_UINT64"
		paramType = "BID_UINT64"
		bits = 64
	case "bid128":
		format = "decimal128"
		returnType = "BID_UINT128"
		paramType = "BID_UINT128"
		bits = 128
	default:
		return "", "", 0, 0, fmt.Errorf("ffi symbol %q: unsupported format prefix", function)
	}

	if integerBits, ok := ffiBaseIntegerFromBits(operation); ok {
		if err := verifyFFIBaseIntegerFromSignature(function, symbol, returnType, operation); err != nil {
			return "", "", 0, 0, err
		}
		return format, operation, integerBits, 1, nil
	}

	if targetBits, ok := ffiBIDConversionTargetBits(operation); ok {
		if err := verifyFFIBIDWidthConversionSignature(function, symbol, prefix, paramType, targetBits); err != nil {
			return "", "", 0, 0, err
		}
		return format, operation, bits, 1, nil
	}

	if targetBits, ok := ffiBinaryConversionTargetBits(operation); ok {
		if err := verifyFFIBinaryConversionSignature(function, symbol, paramType, targetBits); err != nil {
			return "", "", 0, 0, err
		}
		return format, operation, bits, 1, nil
	}

	operationKind, ok := classifyFFIOperation(operation)
	if !ok {
		return "", "", 0, 0, fmt.Errorf("ffi symbol %q: unsupported operation", function)
	}

	expectedReturnType := returnType
	if operationKind.returnType != "" {
		expectedReturnType = operationKind.returnType
	}
	switch operationKind.returnKind {
	case ffiReturnDecimal:
	case ffiReturnInt:
		if operationKind.returnType == "" {
			expectedReturnType = "int"
		}
	case ffiReturnLongLong:
		if operationKind.returnType == "" {
			expectedReturnType = "long long int"
		}
	case ffiReturnClass:
		expectedReturnType = "class_t"
	default:
		return "", "", 0, 0, fmt.Errorf("ffi symbol %q: unsupported return kind", function)
	}

	if symbol.ReturnType != expectedReturnType {
		return "", "", 0, 0, fmt.Errorf("ffi symbol %q: return type = %q, want %q", function, symbol.ReturnType, expectedReturnType)
	}

	expectedParams := []string{paramType + " x"}
	arity := operationKind.arity
	if arity == 3 {
		expectedParams = append(expectedParams, paramType+" y", paramType+" z")
	} else if operation == "scalbn" || operation == "ldexp" {
		expectedParams = append(expectedParams, "int n")
	} else if operation == "scalbln" {
		expectedParams = append(expectedParams, "long int n")
	} else if arity == 2 {
		expectedParams = append(expectedParams, paramType+" y")
	}
	if operationKind.rounding {
		expectedParams = append(expectedParams, "_IDEC_round rnd_mode")
	}
	if operationKind.flags {
		expectedParams = append(expectedParams, "_IDEC_flags*pfpsf")
	}

	if len(symbol.Parameters) != len(expectedParams) {
		return "", "", 0, 0, fmt.Errorf("ffi symbol %q: parameter count = %d, want %d", function, len(symbol.Parameters), len(expectedParams))
	}
	for i, want := range expectedParams {
		if symbol.Parameters[i] != want {
			return "", "", 0, 0, fmt.Errorf("ffi symbol %q: parameter %d = %q, want %q", function, i, symbol.Parameters[i], want)
		}
	}

	return format, operation, bits, arity, nil
}

func verifyFFIMixedDecimalSignature(function string, symbol symbolSpec, shape ffiMixedDecimalShape) error {
	returnType, ok := ffiDecimalReturnType(shape.resultBits)
	if !ok {
		return fmt.Errorf("ffi symbol %q: unsupported mixed decimal result width %d", function, shape.resultBits)
	}
	if symbol.ReturnType != returnType {
		return fmt.Errorf("ffi symbol %q: return type = %q, want %q", function, symbol.ReturnType, returnType)
	}

	paramNames := [...]string{"x", "y", "z"}
	expectedParams := make([]string, 0, shape.operandCount+2)
	for i, bits := range shape.operandWidths() {
		paramType, ok := ffiDecimalReturnType(bits)
		if !ok {
			return fmt.Errorf("ffi symbol %q: unsupported mixed decimal operand %d width %d", function, i, bits)
		}
		expectedParams = append(expectedParams, paramType+" "+paramNames[i])
	}
	expectedParams = append(expectedParams, "_IDEC_round rnd_mode", "_IDEC_flags*pfpsf")
	return verifyFFIParameters(function, symbol.Parameters, expectedParams)
}

func verifyFFIBaseIntegerFromSignature(function string, symbol symbolSpec, returnType, operation string) error {
	if symbol.ReturnType != returnType {
		return fmt.Errorf("ffi symbol %q: return type = %q, want %q", function, symbol.ReturnType, returnType)
	}

	var expectedParams []string
	switch operation {
	case "from_int32":
		expectedParams = []string{"int x"}
	case "from_int64":
		expectedParams = []string{"BID_SINT64 x"}
	case "from_uint32":
		expectedParams = []string{"unsigned int x"}
	case "from_uint64":
		// Intel's public declaration omits the parameter name for some widths;
		// keep the generated signature check exact to the extracted symbol.
		if function == "bid128_from_uint64" {
			expectedParams = []string{"BID_UINT64 x"}
		} else {
			expectedParams = []string{"BID_UINT64"}
		}
	default:
		return fmt.Errorf("ffi symbol %q: unsupported integer constructor operation %q", function, operation)
	}

	switch function {
	case "bid32_from_int32", "bid32_from_int64", "bid32_from_uint32", "bid32_from_uint64",
		"bid64_from_int64", "bid64_from_uint64":
		expectedParams = append(expectedParams, "_IDEC_round rnd_mode", "_IDEC_flags*pfpsf")
	case "bid64_from_int32", "bid64_from_uint32",
		"bid128_from_int32", "bid128_from_int64", "bid128_from_uint32", "bid128_from_uint64":
	default:
		return fmt.Errorf("ffi symbol %q: unsupported integer constructor", function)
	}

	return verifyFFIParameters(function, symbol.Parameters, expectedParams)
}

func verifyFFIBIDWidthConversionSignature(function string, symbol symbolSpec, prefix, paramType string, targetBits int) error {
	expectedReturnType, ok := ffiDecimalReturnType(targetBits)
	if !ok {
		return fmt.Errorf("ffi symbol %q: unsupported BID conversion target width %d", function, targetBits)
	}
	if symbol.ReturnType != expectedReturnType {
		return fmt.Errorf("ffi symbol %q: return type = %q, want %q", function, symbol.ReturnType, expectedReturnType)
	}

	expectedParams := []string{paramType + " x"}
	switch function {
	case "bid32_to_bid64", "bid32_to_bid128", "bid64_to_bid128":
		expectedParams = append(expectedParams, "_IDEC_flags*pfpsf")
	case "bid64_to_bid32", "bid128_to_bid32", "bid128_to_bid64":
		expectedParams = append(expectedParams, "_IDEC_round rnd_mode", "_IDEC_flags*pfpsf")
	default:
		return fmt.Errorf("ffi symbol %q: unsupported BID width conversion for prefix %q", function, prefix)
	}

	return verifyFFIParameters(function, symbol.Parameters, expectedParams)
}

func verifyFFIBinaryConversionSignature(function string, symbol symbolSpec, paramType string, targetBits int) error {
	var expectedReturnType string
	switch targetBits {
	case 32:
		expectedReturnType = "float"
	case 64:
		expectedReturnType = "double"
	case 128:
		expectedReturnType = "BID_UINT128"
	default:
		return fmt.Errorf("ffi symbol %q: unsupported binary conversion target width %d", function, targetBits)
	}
	if symbol.ReturnType != expectedReturnType {
		return fmt.Errorf("ffi symbol %q: return type = %q, want %q", function, symbol.ReturnType, expectedReturnType)
	}
	expectedParams := []string{paramType + " x", "_IDEC_round rnd_mode", "_IDEC_flags*pfpsf"}
	return verifyFFIParameters(function, symbol.Parameters, expectedParams)
}

func verifyFFIParameters(function string, got []string, want []string) error {
	if len(got) != len(want) {
		return fmt.Errorf("ffi symbol %q: parameter count = %d, want %d", function, len(got), len(want))
	}
	for i, expected := range want {
		if got[i] != expected {
			return fmt.Errorf("ffi symbol %q: parameter %d = %q, want %q", function, i, got[i], expected)
		}
	}
	return nil
}

func ffiDecimalReturnType(bits int) (string, bool) {
	switch bits {
	case 32:
		return "BID_UINT32", true
	case 64:
		return "BID_UINT64", true
	case 128:
		return "BID_UINT128", true
	default:
		return "", false
	}
}

type ffiReturnKind int

const (
	ffiReturnDecimal ffiReturnKind = iota
	ffiReturnInt
	ffiReturnLongLong
	ffiReturnClass
)

type ffiOperationKind struct {
	returnKind ffiReturnKind
	returnType string
	arity      int
	rounding   bool
	flags      bool
}

func classifyFFIOperation(operation string) (ffiOperationKind, bool) {
	if kind, ok := classifyFFIBaseIntegerConversionOperation(operation); ok {
		return kind, true
	}
	if _, ok := ffiBaseIntegerFromBits(operation); ok {
		return ffiOperationKind{returnKind: ffiReturnDecimal, arity: 1}, true
	}
	if _, ok := ffiBIDConversionTargetBits(operation); ok {
		return ffiOperationKind{returnKind: ffiReturnDecimal, arity: 1, flags: true}, true
	}
	if _, ok := ffiBinaryConversionTargetBits(operation); ok {
		return ffiOperationKind{returnKind: ffiReturnInt, arity: 1, rounding: true, flags: true}, true
	}
	switch operation {
	case "add", "sub", "mul", "div", "quantize":
		return ffiOperationKind{returnKind: ffiReturnDecimal, arity: 2, rounding: true, flags: true}, true
	case "fma":
		return ffiOperationKind{returnKind: ffiReturnDecimal, arity: 3, rounding: true, flags: true}, true
	case "round_integral_exact", "sqrt":
		return ffiOperationKind{returnKind: ffiReturnDecimal, arity: 1, rounding: true, flags: true}, true
	case "scalbn", "ldexp":
		return ffiOperationKind{returnKind: ffiReturnDecimal, arity: 2, rounding: true, flags: true}, true
	case "scalbln":
		return ffiOperationKind{returnKind: ffiReturnDecimal, arity: 2, rounding: true, flags: true}, true
	case "rem", "fmod", "minnum", "maxnum", "minnum_mag", "maxnum_mag":
		return ffiOperationKind{returnKind: ffiReturnDecimal, arity: 2, flags: true}, true
	case "logb", "nextup", "nextdown", "quantum":
		return ffiOperationKind{returnKind: ffiReturnDecimal, arity: 1, flags: true}, true
	case "copy", "negate", "abs":
		return ffiOperationKind{returnKind: ffiReturnDecimal, arity: 1}, true
	case "copySign":
		return ffiOperationKind{returnKind: ffiReturnDecimal, arity: 2}, true
	case "class":
		return ffiOperationKind{returnKind: ffiReturnClass, arity: 1}, true
	case "isSigned", "isNormal", "isSubnormal", "isFinite", "isZero", "isInf", "isNaN", "isSignaling", "isCanonical", "radix":
		return ffiOperationKind{returnKind: ffiReturnInt, arity: 1}, true
	case "quantexp", "ilogb":
		return ffiOperationKind{returnKind: ffiReturnInt, arity: 1, flags: true}, true
	case "llquantexp":
		return ffiOperationKind{returnKind: ffiReturnLongLong, arity: 1, flags: true}, true
	case "totalOrder", "totalOrderMag", "sameQuantum":
		return ffiOperationKind{returnKind: ffiReturnInt, arity: 2}, true
	case "quiet_equal", "quiet_greater", "quiet_greater_equal", "quiet_greater_unordered", "quiet_less", "quiet_less_equal", "quiet_less_unordered", "quiet_not_equal", "quiet_not_greater", "quiet_not_less", "quiet_ordered", "quiet_unordered":
		return ffiOperationKind{returnKind: ffiReturnInt, arity: 2, flags: true}, true
	case "signaling_greater", "signaling_greater_equal", "signaling_greater_unordered", "signaling_less", "signaling_less_equal", "signaling_less_unordered", "signaling_not_greater", "signaling_not_less":
		return ffiOperationKind{returnKind: ffiReturnInt, arity: 2, flags: true}, true
	default:
		return ffiOperationKind{}, false
	}
}

func ffiBaseIntegerFromBits(operation string) (int, bool) {
	switch operation {
	case "from_int32", "from_uint32":
		return 32, true
	case "from_int64", "from_uint64":
		return 64, true
	default:
		return 0, false
	}
}

func ffiBIDConversionTargetBits(operation string) (int, bool) {
	switch operation {
	case "to_bid32":
		return 32, true
	case "to_bid64":
		return 64, true
	case "to_bid128":
		return 128, true
	default:
		return 0, false
	}
}

func ffiBinaryConversionTargetBits(operation string) (int, bool) {
	switch operation {
	case "to_binary32":
		return 32, true
	case "to_binary64":
		return 64, true
	case "to_binary128":
		return 128, true
	default:
		return 0, false
	}
}

func classifyFFIBaseIntegerConversionOperation(operation string) (ffiOperationKind, bool) {
	parts := strings.Split(operation, "_")
	if len(parts) != 3 || parts[0] != "to" {
		return ffiOperationKind{}, false
	}
	switch parts[2] {
	case "ceil", "floor", "int", "rnint", "rninta", "xceil", "xfloor", "xint", "xrnint", "xrninta":
	default:
		return ffiOperationKind{}, false
	}
	switch parts[1] {
	case "int8":
		return ffiOperationKind{returnKind: ffiReturnInt, returnType: "char", arity: 1, flags: true}, true
	case "int16":
		return ffiOperationKind{returnKind: ffiReturnInt, returnType: "short", arity: 1, flags: true}, true
	case "int32":
		return ffiOperationKind{returnKind: ffiReturnInt, returnType: "int", arity: 1, flags: true}, true
	case "int64":
		return ffiOperationKind{returnKind: ffiReturnLongLong, returnType: "BID_SINT64", arity: 1, flags: true}, true
	case "uint8":
		return ffiOperationKind{returnKind: ffiReturnInt, returnType: "unsigned char", arity: 1, flags: true}, true
	case "uint16":
		return ffiOperationKind{returnKind: ffiReturnInt, returnType: "unsigned short", arity: 1, flags: true}, true
	case "uint32":
		return ffiOperationKind{returnKind: ffiReturnInt, returnType: "unsigned int", arity: 1, flags: true}, true
	case "uint64":
		return ffiOperationKind{returnKind: ffiReturnLongLong, returnType: "BID_UINT64", arity: 1, flags: true}, true
	default:
		return ffiOperationKind{}, false
	}
}

type deterministicFFIGenerator struct {
	state     uint64
	bits      int
	mask      uint64
	edges     []uint64
	wideEdges [][2]uint64
}

type deterministicFFIMixedDecimalGenerator struct {
	shape      ffiMixedDecimalShape
	generators []*deterministicFFIGenerator
}

func newDeterministicFFIMixedDecimalGenerator(seed uint64, function string, shape ffiMixedDecimalShape) *deterministicFFIMixedDecimalGenerator {
	widths := shape.operandWidths()
	generators := make([]*deterministicFFIGenerator, len(widths))
	for i, bits := range widths {
		// Salt random operands by their ordered x/y/z slot. Edge operands stay
		// aligned by the width-specific edge tables, while equal-width random
		// slots do not collapse to identical bit patterns.
		generators[i] = newDeterministicFFIGenerator(seed, fmt.Sprintf("%s#operand%d", function, i), bits)
	}
	return &deterministicFFIMixedDecimalGenerator{shape: shape, generators: generators}
}

func (g *deterministicFFIMixedDecimalGenerator) nextOperands(index int) []string {
	operands := make([]string, len(g.generators))
	for position, generator := range g.generators {
		operands[position] = generator.nextMixedDecimalOperand(index, g.shape.operandCount, position)
	}
	return operands
}

func (g *deterministicFFIGenerator) nextMixedDecimalOperand(index, arity, position int) string {
	switch arity {
	case 1:
		if index < g.decimalEdgeCount() {
			return g.formatDecimalEdge(index)
		}
	case 3:
		if triple, ok := ffiTernaryEdgeTriple(index, g.bits, g.decimalEdgeCount()); ok {
			return g.formatDecimalEdge(triple[position])
		}
	default:
		panic("unsupported mixed decimal ffi arity")
	}
	return g.nextRandomDecimalOperand()
}

func (g *deterministicFFIGenerator) decimalEdgeCount() int {
	if g.bits == 128 {
		return len(g.wideEdges)
	}
	return len(g.edges)
}

func (g *deterministicFFIGenerator) formatDecimalEdge(index int) string {
	if g.bits == 128 {
		hi, lo := g.wideEdge(index)
		return formatFFIWideBits(hi, lo, g.bits)
	}
	return formatFFIBits(g.edges[index], g.bits)
}

func (g *deterministicFFIGenerator) nextRandomDecimalOperand() string {
	if g.bits == 128 {
		return formatFFIWideBits(g.next(), g.next(), g.bits)
	}
	return formatFFIBits(g.next(), g.bits)
}

func newDeterministicFFIGenerator(seed uint64, function string, bits int) *deterministicFFIGenerator {
	state := seed ^ mixFFISeed(hashFFIString(function))
	if state == 0 {
		state = 0x9e3779b97f4a7c15
	}
	mask := uint64(^uint64(0))
	if bits < 64 {
		mask = (uint64(1) << bits) - 1
	}
	return &deterministicFFIGenerator{
		state:     state,
		bits:      bits,
		mask:      mask,
		edges:     ffiEdgeValues(bits),
		wideEdges: ffiWideEdgeValues(bits),
	}
}

func (g *deterministicFFIGenerator) nextPair(index int) (uint64, uint64) {
	if pair, ok := ffiBinaryEdgePair(index, g.bits, len(g.edges)); ok {
		return g.edges[pair[0]], g.edges[pair[1]]
	}
	return g.next(), g.next()
}

func (g *deterministicFFIGenerator) nextTriple(index int) (uint64, uint64, uint64) {
	if triple, ok := ffiTernaryEdgeTriple(index, g.bits, len(g.edges)); ok {
		return g.edges[triple[0]], g.edges[triple[1]], g.edges[triple[2]]
	}
	return g.next(), g.next(), g.next()
}

func (g *deterministicFFIGenerator) nextOperands(index int, arity int) []string {
	if g.bits == 128 {
		return g.nextOperands128(index, arity)
	}
	switch arity {
	case 1:
		return []string{formatFFIBits(g.nextSingle(index), g.bits)}
	case 2:
		a, b := g.nextPair(index)
		return []string{formatFFIBits(a, g.bits), formatFFIBits(b, g.bits)}
	case 3:
		a, b, c := g.nextTriple(index)
		return []string{formatFFIBits(a, g.bits), formatFFIBits(b, g.bits), formatFFIBits(c, g.bits)}
	default:
		panic("unsupported ffi arity")
	}
}

func (g *deterministicFFIGenerator) nextOperandsForOperation(index int, operation string, arity int) []string {
	if _, ok := ffiBaseIntegerFromBits(operation); ok {
		return []string{g.nextBaseIntegerFromOperand(index, operation)}
	}
	if operation == "quantum" {
		return []string{g.nextQuantumOperand(index)}
	}
	if operation != "scalbn" && operation != "ldexp" && operation != "scalbln" {
		return g.nextOperands(index, arity)
	}
	if g.bits == 128 {
		hi, lo := g.nextWide(index)
		return []string{
			formatFFIWideBits(hi, lo, 128),
			strconv.Itoa(ffiScaleBExponent(index)),
		}
	}
	return []string{
		formatFFIBits(g.nextSingle(index), g.bits),
		strconv.Itoa(ffiScaleBExponent(index)),
	}
}

func (g *deterministicFFIGenerator) nextBaseIntegerFromOperand(index int, operation string) string {
	edges := map[string][]string{
		"from_int32": {
			"0", "1", "-1", "2", "-2", "9", "-9", "10", "-10", "9999999", "-9999999", "10000000",
			"-10000000", "2147483647", "-2147483648", "123456789", "-123456789", "1000000000", "-1000000000",
			"999999995", "-999999995", "999999999", "-999999999", "7654321",
		},
		"from_uint32": {
			"0", "1", "2", "9", "10", "99", "100", "999", "1000", "9999999", "10000000", "99999999",
			"100000000", "123456789", "999999995", "999999999", "1000000000", "2147483647", "2147483648",
			"3000000000", "4000000000", "4294967294", "4294967295", "7654321",
		},
		"from_int64": {
			"0", "1", "-1", "2", "-2", "9", "-9", "10", "-10", "9999999999999999", "-9999999999999999",
			"10000000000000000", "-10000000000000000", "99999999999999995", "-99999999999999995",
			"999999999999999999", "-999999999999999999", "1000000000000000000", "-1000000000000000000",
			"9223372036854775807", "-9223372036854775808", "1234567890123456789", "-1234567890123456789",
			"76543210987654321",
		},
		"from_uint64": {
			"0", "1", "2", "9", "10", "99", "100", "999", "1000", "9999999999999999",
			"10000000000000000", "99999999999999995", "999999999999999999", "1000000000000000000",
			"9223372036854775807", "9223372036854775808", "9999999999999999995", "9999999999999999999",
			"10000000000000000000", "12345678901234567890", "18446744073709551614", "18446744073709551615",
			"76543210987654321", "4000000000000000000",
		},
	}
	if values := edges[operation]; index < len(values) {
		return values[index]
	}
	switch operation {
	case "from_int32":
		return strconv.FormatInt(int64(int32(g.next())), 10)
	case "from_uint32":
		return strconv.FormatUint(uint64(uint32(g.next())), 10)
	case "from_int64":
		return strconv.FormatInt(int64(g.next()), 10)
	case "from_uint64":
		return strconv.FormatUint(g.next(), 10)
	default:
		panic("unsupported ffi integer constructor operation")
	}
}

func (g *deterministicFFIGenerator) nextQuantumOperand(index int) string {
	if g.bits == 128 {
		// Keep BID128 quantum exact bit-compare away from NaN payload cases.
		// Intel BID C can return payload bits that are not a stable oracle for
		// this generated FFI profile; TEST_GENERATION_SPEC.md documents this
		// as a deliberate current-scope exclusion.
		values := []string{
			"00000000000000000000000000000000",
			"01000000000000000000000000000000",
			"02000000000000000000000000000000",
			"01000000000000300000000000000000",
			"01000000000000b00000000000000000",
			"00000000000000000100000000000030",
			"000000000000000001000000000000b0",
			"39300000000000300000000000000000",
			"39300000000000b00000000000000000",
			"00000000000000000000000000000010",
			"00000000000000000000000000000020",
			"00000000000000000000000000000040",
			"00000000000000000000000000000050",
			"00000000000000000000000000000060",
			"78563412000000300000000000000000",
			"78563412000000b00000000000000000",
			"ffffffffffff00300000000000000000",
			"ffffffffffff00b00000000000000000",
			"01000000000000100000000000000000",
			"01000000000000200000000000000000",
			"01000000000000400000000000000000",
			"01000000000000500000000000000000",
			"01000000000000600000000000000000",
			"01000000000000600100000000000030",
		}
		if index < len(values) {
			return values[index]
		}
		return values[index%len(values)]
	}

	if g.bits == 32 {
		values := []uint64{
			0x00000000,
			0x80000000,
			0x00000001,
			0x00000002,
			0x78000000,
			0xf8000000,
			0x32800001,
			0xb2800001,
			0x22800001,
			0xa2800001,
			0x60000000,
			0xe0000000,
		}
		if index < len(values) {
			return formatFFIBits(values[index], g.bits)
		}
		value := (g.next() & g.mask) &^ 0x7e000000
		value |= 0x32800000
		return formatFFIBits(value, g.bits)
	}

	values := []uint64{
		0x0000000000000000,
		0x8000000000000000,
		0x0000000000000001,
		0x0000000000000002,
		0x7800000000000000,
		0xf800000000000000,
		0x31c0000000000001,
		0xb1c0000000000001,
		0x2238000000000001,
		0xa238000000000001,
		0x6000000000000000,
		0xe000000000000000,
	}
	if index < len(values) {
		return formatFFIBits(values[index]&g.mask, g.bits)
	}
	value := (g.next() & g.mask) &^ uint64(0x7e00000000000000)
	value |= 0x31c0000000000000
	return formatFFIBits(value, g.bits)
}

func (g *deterministicFFIGenerator) nextOperands128(index int, arity int) []string {
	switch arity {
	case 1:
		hi, lo := g.nextWide(index)
		return []string{formatFFIWideBits(hi, lo, 128)}
	case 2:
		var ahi, alo, bhi, blo uint64
		if pair, ok := ffiBinaryEdgePair(index, g.bits, len(g.wideEdges)); ok {
			ahi, alo = g.wideEdge(pair[0])
			bhi, blo = g.wideEdge(pair[1])
		} else {
			ahi, alo = g.next(), g.next()
			bhi, blo = g.next(), g.next()
		}
		return []string{
			formatFFIWideBits(ahi, alo, 128),
			formatFFIWideBits(bhi, blo, 128),
		}
	case 3:
		var ahi, alo, bhi, blo, chi, clo uint64
		if triple, ok := ffiTernaryEdgeTriple(index, g.bits, len(g.wideEdges)); ok {
			ahi, alo = g.wideEdge(triple[0])
			bhi, blo = g.wideEdge(triple[1])
			chi, clo = g.wideEdge(triple[2])
		} else {
			ahi, alo = g.next(), g.next()
			bhi, blo = g.next(), g.next()
			chi, clo = g.next(), g.next()
		}
		return []string{
			formatFFIWideBits(ahi, alo, 128),
			formatFFIWideBits(bhi, blo, 128),
			formatFFIWideBits(chi, clo, 128),
		}
	default:
		panic("unsupported ffi arity")
	}
}

func (g *deterministicFFIGenerator) wideEdge(index int) (uint64, uint64) {
	return g.wideEdges[index][0], g.wideEdges[index][1]
}

func (g *deterministicFFIGenerator) nextWide(index int) (uint64, uint64) {
	if index < len(g.wideEdges) {
		return g.wideEdge(index)
	}
	return g.next(), g.next()
}

func (g *deterministicFFIGenerator) nextSingle(index int) uint64 {
	if index < len(g.edges) {
		return g.edges[index]
	}
	return g.next()
}

var ffiScaleBEdgeExponents = []int{
	0,
	1,
	-1,
	10,
	-10,
	398,
	-398,
	1000,
	-1000,
	2016,
	-2016,
	999999,
	-999999,
}

func ffiScaleBExponent(index int) int {
	if index < len(ffiScaleBEdgeExponents) {
		return ffiScaleBEdgeExponents[index]
	}
	state := mixFFISeed(uint64(index+1) * 0x9e3779b97f4a7c15)
	return int(state%4097) - 2048
}

func (g *deterministicFFIGenerator) next() uint64 {
	g.state += 0x9e3779b97f4a7c15
	value := mixFFISeed(g.state) & g.mask
	return value
}

func ffiEdgeValues(bits int) []uint64 {
	switch bits {
	case 32:
		edges := bid32BidCodecEdgeValues()
		values := make([]uint64, 0, len(edges))
		for _, value := range edges {
			values = append(values, uint64(value))
		}
		return values
	case 64:
		// The FFI edge matrix indexes this list by fixed position
		// (ffiBinaryEdgePairs and the ternary table), so it uses the pinned
		// format-level corpus, not the BID codec vector edge list that
		// extends it.
		edges := bid64FormatEdgeValues()
		values := make([]uint64, 0, len(edges))
		values = append(values, edges...)
		values = append(values,
			0x6000000000000000,
			0xe000000000000000,
			0x77fb86f26fc0ffff,
			0xf7fb86f26fc0ffff,
		)
		return values
	default:
		return nil
	}
}

func ffiWideEdgeValues(bits int) [][2]uint64 {
	if bits != 128 {
		return nil
	}
	// The FFI edge matrix indexes this list by fixed position
	// (ffiBinaryEdgePairs and the ternary table), so it uses the pinned
	// format-level corpus, not the BID codec vector edge list that extends it.
	edges := bid128FormatEdgeValues()
	values := make([][2]uint64, 0, len(edges))
	for _, value := range edges {
		values = append(values, [2]uint64{value.hi, value.lo})
	}
	values = append(values,
		[2]uint64{0x6000000000000000, 0x0000000000000000},
		[2]uint64{0xe000000000000000, 0x0000000000000000},
		[2]uint64{0xfc00000000000000, 0x0000000000000000},
		[2]uint64{0xfe00000000000000, 0x0000000000000000},
	)
	return values
}

func ffiBinaryEdgePair(index int, bits int, edgeCount int) ([2]int, bool) {
	pairs := ffiBinaryEdgePairs(bits)
	if index >= len(pairs) {
		return [2]int{}, false
	}
	pair := pairs[index]
	if pair[0] >= edgeCount || pair[1] >= edgeCount {
		return [2]int{}, false
	}
	return pair, true
}

func ffiBinaryEdgePairs(bits int) [][2]int {
	switch bits {
	case 32, 64:
		return [][2]int{
			{0, 1},
			{4, 5},
			{8, 9},
			{10, 4},
			{4, 10},
			{11, 5},
			{5, 11},
			{12, 4},
			{4, 12},
			{15, 4},
			{4, 15},
			{10, 11},
			{12, 15},
			{17, 4},
			{4, 17},
		}
	case 128:
		return [][2]int{
			{0, 1},
			{2, 3},
			{4, 2},
			{5, 2},
			{2, 5},
			{6, 3},
			{3, 6},
			{7, 2},
			{2, 7},
			{9, 2},
			{2, 9},
			{5, 6},
			{7, 9},
			{10, 2},
			{2, 10},
		}
	default:
		return nil
	}
}

func ffiTernaryEdgeTriple(index int, bits int, edgeCount int) ([3]int, bool) {
	triples := ffiTernaryEdgeTriples(bits)
	if index >= len(triples) {
		return [3]int{}, false
	}
	triple := triples[index]
	if triple[0] >= edgeCount || triple[1] >= edgeCount || triple[2] >= edgeCount {
		return [3]int{}, false
	}
	return triple, true
}

func ffiTernaryEdgeTriples(bits int) [][3]int {
	switch bits {
	case 32, 64:
		return [][3]int{
			{4, 4, 4},
			{4, 5, 0},
			{8, 4, 5},
			{10, 4, 0},
			{4, 10, 0},
			{12, 4, 4},
			{4, 12, 4},
			{15, 4, 4},
			{4, 15, 4},
			{17, 4, 4},
		}
	case 128:
		return [][3]int{
			{2, 2, 2},
			{2, 3, 0},
			{4, 2, 3},
			{5, 2, 0},
			{2, 5, 0},
			{7, 2, 2},
			{2, 7, 2},
			{9, 2, 2},
			{2, 9, 2},
			{10, 2, 2},
		}
	default:
		return nil
	}
}

func hashFFIString(input string) uint64 {
	var hash uint64 = 1469598103934665603
	for i := 0; i < len(input); i++ {
		hash ^= uint64(input[i])
		hash *= 1099511628211
	}
	return hash
}

func mixFFISeed(x uint64) uint64 {
	x ^= x >> 30
	x *= 0xbf58476d1ce4e5b9
	x ^= x >> 27
	x *= 0x94d049bb133111eb
	x ^= x >> 31
	return x
}

func formatFFIBits(value uint64, bits int) string {
	if bits == 32 {
		return fmt.Sprintf("%08x", uint32(value))
	}
	return fmt.Sprintf("%016x", value)
}

// formatFFIWideBits emits the raw little-endian 16-byte image of the
// (high, low) BID words. The sole runtime consumer of 128-bit FFI operand
// strings, parseFFIUint128Bits on the generated Go native runner, decodes
// them as that byte image (the same LE convention ffi_fusedness.go rawString
// and the value-level Rust parity emit already follow), so the emitted
// string must be the LE image for the declared words to be the values
// actually injected.
func formatFFIWideBits(high uint64, low uint64, bits int) string {
	if bits != 128 {
		panic("unsupported wide ffi width")
	}
	var raw [16]byte
	binary.LittleEndian.PutUint64(raw[0:8], low)
	binary.LittleEndian.PutUint64(raw[8:16], high)
	return hex.EncodeToString(raw[:])
}
