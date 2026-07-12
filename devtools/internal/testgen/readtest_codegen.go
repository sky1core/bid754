package testgen

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/sky1core/bid754/devtools/internal/genmarker"
)

const (
	readtestGeneratedNativePath = "../bid754-go/generated_readtest_dispatch_native.go"
	readtestGeneratedStubPath   = "../bid754-go/generated_readtest_dispatch_stub.go"
	readtestGeneratedSharedPath = "../bid754-go/generated_readtest_shared.go"
	readtestSymbolSourcePath    = "generated/json/intel_dfp_symbols.json"
)

type readtestParamKind string

const (
	readtestParamU32    readtestParamKind = "u32"
	readtestParamU64    readtestParamKind = "u64"
	readtestParamU128   readtestParamKind = "u128"
	readtestParamP32    readtestParamKind = "p32"
	readtestParamP64    readtestParamKind = "p64"
	readtestParamP128   readtestParamKind = "p128"
	readtestParamPInt   readtestParamKind = "pint"
	readtestParamCStr   readtestParamKind = "cstr"
	readtestParamInt    readtestParamKind = "int"
	readtestParamUInt   readtestParamKind = "uint"
	readtestParamLInt   readtestParamKind = "lint"
	readtestParamS64    readtestParamKind = "s64"
	readtestParamRound  readtestParamKind = "round"
	readtestParamFlags  readtestParamKind = "flags"
	readtestParamMasks  readtestParamKind = "masks"
	readtestParamInfo   readtestParamKind = "info"
	readtestParamCharPs readtestParamKind = "char_ps"
)

type readtestDispatchSpec struct {
	ReadTestSpec
	Symbol     symbolSpec
	ParamKinds []readtestParamKind
}

func WriteReadtestDispatchOutputs(repoRoot string, manifest Manifest) error {
	files, err := GenerateReadtestDispatchOutputs(repoRoot, manifest)
	if err != nil {
		return err
	}
	for path, data := range files {
		fullPath := filepath.Join(repoRoot, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			return fmt.Errorf("mkdir %q: %w", filepath.Dir(fullPath), err)
		}
		if err := os.WriteFile(fullPath, data, 0o644); err != nil {
			return fmt.Errorf("write generated readtest dispatch %q: %w", fullPath, err)
		}
	}
	return nil
}

func GenerateReadtestDispatchOutputs(repoRoot string, manifest Manifest) (map[string][]byte, error) {
	reads, err := collectReadtestDispatchSpecs(repoRoot, manifest)
	if err != nil {
		return nil, err
	}
	symbols, err := loadSymbolFile(filepath.Join(repoRoot, readtestSymbolSourcePath))
	if err != nil {
		return nil, err
	}
	index := make(map[string]symbolSpec, len(symbols.Symbols))
	for _, symbol := range symbols.Symbols {
		index[symbol.Name] = symbol
	}
	for _, read := range reads {
		if _, ok := index[read.Function]; !ok {
			return nil, fmt.Errorf("readtest dispatch: symbol %q not found in %s", read.Function, readtestSymbolSourcePath)
		}
	}

	native, err := generateReadtestNativeDispatch(reads, index)
	if err != nil {
		return nil, err
	}
	stub := generateReadtestStubDispatch()
	return formatGeneratedGoOutputs(map[string][]byte{
		readtestGeneratedNativePath: native,
		readtestGeneratedStubPath:   stub,
		readtestGeneratedSharedPath: []byte(readtestGeneratedSharedHelpers),
	})
}

func collectReadtestDispatchSpecs(repoRoot string, manifest Manifest) ([]ReadTestSpec, error) {
	var reads []ReadTestSpec
	for _, profile := range manifest.ReadProfiles {
		profileReads, err := expandReadTestProfile(repoRoot, profile)
		if err != nil {
			return nil, err
		}
		reads = append(reads, profileReads...)
	}
	seen := map[string]bool{}
	deduped := make([]ReadTestSpec, 0, len(reads))
	for _, read := range reads {
		if seen[read.Function] {
			continue
		}
		seen[read.Function] = true
		deduped = append(deduped, read)
	}
	sort.Slice(deduped, func(i, j int) bool { return deduped[i].Function < deduped[j].Function })
	return deduped, nil
}

func generateReadtestNativeDispatch(reads []ReadTestSpec, symbols map[string]symbolSpec) ([]byte, error) {
	var c bytes.Buffer
	var goCode bytes.Buffer

	c.WriteString(genmarker.Line("testgen") + "\n")
	c.WriteString("//go:build cgo && bid754_native\n\n")
	c.WriteString("package bid754\n\n")
	c.WriteString("/*\n")
	c.WriteString("#cgo CFLAGS: -DDECNUMDIGITS=34 -I${SRCDIR}/../devtools/third_party/intel_dfp/src -I${SRCDIR}/../devtools/third_party/intel_dfp/include\n")
	c.WriteString("#cgo LDFLAGS: -ldecnumber -L${SRCDIR}/../devtools/third_party/intel_dfp/lib -lbid -lm\n\n")
	c.WriteString("#include <stdint.h>\n")
	c.WriteString("#include <string.h>\n")
	c.WriteString("#include <stdlib.h>\n")
	c.WriteString("#include \"bid_conf.h\"\n")
	c.WriteString("#include \"bid_functions.h\"\n\n")

	var decimalOps32, decimalOps64, decimalOps128 []readtestDispatchSpec
	var scalarSignedOps, scalarUnsignedOps []readtestDispatchSpec
	var scalarBinary32Ops, scalarBinary64Ops, scalarBinary128Ops []readtestDispatchSpec
	hasFrom32, hasFrom64, hasFrom128 := false, false, false
	hasTo32, hasTo64, hasTo128 := false, false, false

	for _, read := range reads {
		symbol, ok := symbols[read.Function]
		if !ok {
			return nil, fmt.Errorf("readtest dispatch: symbol %q not found", read.Function)
		}
		switch read.Function {
		case "bid32_from_string":
			hasFrom32 = true
		case "bid64_from_string":
			hasFrom64 = true
		case "bid128_from_string":
			hasFrom128 = true
		case "bid32_to_string":
			hasTo32 = true
		case "bid64_to_string":
			hasTo64 = true
		case "bid128_to_string":
			hasTo128 = true
		}
		if read.Kind == "from_string" || read.Kind == "to_string" {
			continue
		}
		var paramKinds []readtestParamKind
		if !isReadtestStatusControlFunction(read.Function) {
			var err error
			paramKinds, err = classifyReadtestParameters(symbol.Parameters)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", read.Function, err)
			}
		}
		dispatch := readtestDispatchSpec{
			ReadTestSpec: read,
			Symbol:       symbol,
			ParamKinds:   paramKinds,
		}
		switch {
		case isReadtestDecimalOutput(read.OutputType) && read.Format == "decimal32":
			decimalOps32 = append(decimalOps32, dispatch)
		case isReadtestDecimalOutput(read.OutputType) && read.Format == "decimal64":
			decimalOps64 = append(decimalOps64, dispatch)
		case isReadtestDecimalOutput(read.OutputType) && read.Format == "decimal128":
			decimalOps128 = append(decimalOps128, dispatch)
		case isReadtestBinary32Output(read.OutputType):
			scalarBinary32Ops = append(scalarBinary32Ops, dispatch)
		case isReadtestBinary64Output(read.OutputType):
			scalarBinary64Ops = append(scalarBinary64Ops, dispatch)
		case isReadtestBinary128Output(read.OutputType):
			scalarBinary128Ops = append(scalarBinary128Ops, dispatch)
		case isReadtestUnsignedOutput(read.OutputType):
			scalarUnsignedOps = append(scalarUnsignedOps, dispatch)
		default:
			scalarSignedOps = append(scalarSignedOps, dispatch)
		}
	}

	if hasFrom32 {
		c.WriteString("static BID_UINT32 bid754_generated_readtest_bid32_from_string(const char* input, int rounding_mode, _IDEC_flags* out_flags) {\n")
		c.WriteString("\t_IDEC_flags flags = 0;\n")
		c.WriteString("\tBID_UINT32 result = bid32_from_string((char*)input, (_IDEC_round)rounding_mode, &flags);\n")
		c.WriteString("\tif (out_flags) { *out_flags = flags; }\n")
		c.WriteString("\treturn result;\n")
		c.WriteString("}\n\n")
	}
	if hasFrom64 {
		c.WriteString("static BID_UINT64 bid754_generated_readtest_bid64_from_string(const char* input, int rounding_mode, _IDEC_flags* out_flags) {\n")
		c.WriteString("\t_IDEC_flags flags = 0;\n")
		c.WriteString("\tBID_UINT64 result = bid64_from_string((char*)input, (_IDEC_round)rounding_mode, &flags);\n")
		c.WriteString("\tif (out_flags) { *out_flags = flags; }\n")
		c.WriteString("\treturn result;\n")
		c.WriteString("}\n\n")
	}
	if hasFrom128 {
		c.WriteString("static void bid754_generated_readtest_bid128_from_string(const char* input, int rounding_mode, _IDEC_flags* out_flags, unsigned char out[16]) {\n")
		c.WriteString("\t_IDEC_flags flags = 0;\n")
		c.WriteString("\tBID_UINT128 result = bid128_from_string((char*)input, (_IDEC_round)rounding_mode, &flags);\n")
		c.WriteString("\tif (out_flags) { *out_flags = flags; }\n")
		c.WriteString("\tmemcpy(out, &result, 16);\n")
		c.WriteString("}\n\n")
	}
	if hasTo32 {
		c.WriteString("static uint32_t bid754_generated_readtest_bid32_to_string(char* out, BID_UINT32 x) {\n")
		c.WriteString("\t_IDEC_flags flags = 0;\n")
		c.WriteString("\tbid32_to_string(out, x, &flags);\n")
		c.WriteString("\treturn (uint32_t)flags;\n")
		c.WriteString("}\n\n")
	}
	if hasTo64 {
		c.WriteString("static uint32_t bid754_generated_readtest_bid64_to_string(char* out, BID_UINT64 x) {\n")
		c.WriteString("\t_IDEC_flags flags = 0;\n")
		c.WriteString("\tbid64_to_string(out, x, &flags);\n")
		c.WriteString("\treturn (uint32_t)flags;\n")
		c.WriteString("}\n\n")
	}
	if hasTo128 {
		c.WriteString("static uint32_t bid754_generated_readtest_bid128_to_string(char* out, const unsigned char in[16]) {\n")
		c.WriteString("\t_IDEC_flags flags = 0;\n")
		c.WriteString("\tBID_UINT128 value;\n")
		c.WriteString("\tmemcpy(&value, in, 16);\n")
		c.WriteString("\tbid128_to_string(out, value, &flags);\n")
		c.WriteString("\treturn (uint32_t)flags;\n")
		c.WriteString("}\n\n")
	}
	for _, dispatch := range decimalOps32 {
		wrapper, err := emitReadtestCWrapper("BID_UINT32", dispatch, true)
		if err != nil {
			return nil, err
		}
		c.WriteString(wrapper)
		c.WriteString("\n")
	}
	for _, dispatch := range decimalOps64 {
		wrapper, err := emitReadtestCWrapper("BID_UINT64", dispatch, true)
		if err != nil {
			return nil, err
		}
		c.WriteString(wrapper)
		c.WriteString("\n")
	}
	for _, dispatch := range decimalOps128 {
		wrapper, err := emitReadtestCWrapper128(dispatch, true)
		if err != nil {
			return nil, err
		}
		c.WriteString(wrapper)
		c.WriteString("\n")
	}
	for _, dispatch := range scalarSignedOps {
		wrapper, err := emitReadtestCWrapper("int64_t", dispatch, false)
		if err != nil {
			return nil, err
		}
		c.WriteString(wrapper)
		c.WriteString("\n")
	}
	for _, dispatch := range scalarBinary32Ops {
		wrapper, err := emitReadtestCWrapper("float", dispatch, false)
		if err != nil {
			return nil, err
		}
		c.WriteString(wrapper)
		c.WriteString("\n")
	}
	for _, dispatch := range scalarBinary64Ops {
		wrapper, err := emitReadtestCWrapper("double", dispatch, false)
		if err != nil {
			return nil, err
		}
		c.WriteString(wrapper)
		c.WriteString("\n")
	}
	for _, dispatch := range scalarBinary128Ops {
		wrapper, err := emitReadtestCWrapper128(dispatch, false)
		if err != nil {
			return nil, err
		}
		c.WriteString(wrapper)
		c.WriteString("\n")
	}
	for _, dispatch := range scalarUnsignedOps {
		var wrapper string
		var err error
		if isReadtestStatusControlFunction(dispatch.Function) {
			wrapper, err = emitReadtestStatusControlCWrapper(dispatch)
		} else {
			wrapper, err = emitReadtestCWrapper("uint64_t", dispatch, false)
		}
		if err != nil {
			return nil, err
		}
		c.WriteString(wrapper)
		c.WriteString("\n")
	}
	c.WriteString("*/\n")
	c.WriteString("import \"C\"\n\n")
	c.WriteString("import (\n\t\"fmt\"\n\t\"math\"\n\t\"strings\"\n\t\"unsafe\"\n)\n\n")

	if hasFrom32 {
		goCode.WriteString("func nativeReadtestBID32FromString(input string, rounding int) (uint32, string) {\n")
		goCode.WriteString("\tcstr := C.CString(input)\n")
		goCode.WriteString("\tdefer C.free(unsafe.Pointer(cstr))\n")
		goCode.WriteString("\tvar flags C._IDEC_flags\n")
		goCode.WriteString("\tvalue := uint32(C.bid754_generated_readtest_bid32_from_string(cstr, C.int(rounding), &flags))\n")
		goCode.WriteString("\treturn value, formatReadtestStatus(uint32(flags))\n")
		goCode.WriteString("}\n\n")
	}
	if hasFrom64 {
		goCode.WriteString("func nativeReadtestBID64FromString(input string, rounding int) (uint64, string) {\n")
		goCode.WriteString("\tcstr := C.CString(input)\n")
		goCode.WriteString("\tdefer C.free(unsafe.Pointer(cstr))\n")
		goCode.WriteString("\tvar flags C._IDEC_flags\n")
		goCode.WriteString("\tvalue := uint64(C.bid754_generated_readtest_bid64_from_string(cstr, C.int(rounding), &flags))\n")
		goCode.WriteString("\treturn value, formatReadtestStatus(uint32(flags))\n")
		goCode.WriteString("}\n\n")
	}
	if hasFrom128 {
		goCode.WriteString("func nativeReadtestBID128FromString(input string, rounding int) ([16]byte, string) {\n")
		goCode.WriteString("\tcstr := C.CString(input)\n")
		goCode.WriteString("\tdefer C.free(unsafe.Pointer(cstr))\n")
		goCode.WriteString("\tvar flags C._IDEC_flags\n")
		goCode.WriteString("\tvar value [16]byte\n")
		goCode.WriteString("\tC.bid754_generated_readtest_bid128_from_string(cstr, C.int(rounding), &flags, (*C.uchar)(unsafe.Pointer(&value[0])))\n")
		goCode.WriteString("\treturn value, formatReadtestStatus(uint32(flags))\n")
		goCode.WriteString("}\n\n")
	}
	// The to_string entries return the raw library output; readtest.c
	// check_results compares that string by round-tripping it through
	// bid*_from_string, so no display normalization is applied here.
	if hasTo32 {
		goCode.WriteString("func nativeReadtestBID32ToString(a uint32) (string, string) {\n")
		goCode.WriteString("\tbuf := make([]byte, 128)\n")
		goCode.WriteString("\tflags := uint32(C.bid754_generated_readtest_bid32_to_string((*C.char)(unsafe.Pointer(&buf[0])), C.BID_UINT32(a)))\n")
		goCode.WriteString("\treturn C.GoString((*C.char)(unsafe.Pointer(&buf[0]))), formatReadtestStatus(flags)\n")
		goCode.WriteString("}\n\n")
	}
	if hasTo64 {
		goCode.WriteString("func nativeReadtestBID64ToString(a uint64) (string, string) {\n")
		goCode.WriteString("\tbuf := make([]byte, 128)\n")
		goCode.WriteString("\tflags := uint32(C.bid754_generated_readtest_bid64_to_string((*C.char)(unsafe.Pointer(&buf[0])), C.BID_UINT64(a)))\n")
		goCode.WriteString("\treturn C.GoString((*C.char)(unsafe.Pointer(&buf[0]))), formatReadtestStatus(flags)\n")
		goCode.WriteString("}\n\n")
	}
	if hasTo128 {
		goCode.WriteString("func nativeReadtestBID128ToString(a [16]byte) (string, string) {\n")
		goCode.WriteString("\tbuf := make([]byte, 128)\n")
		goCode.WriteString("\tflags := uint32(C.bid754_generated_readtest_bid128_to_string((*C.char)(unsafe.Pointer(&buf[0])), (*C.uchar)(unsafe.Pointer(&a[0]))))\n")
		goCode.WriteString("\treturn C.GoString((*C.char)(unsafe.Pointer(&buf[0]))), formatReadtestStatus(flags)\n")
		goCode.WriteString("}\n\n")
	}

	goCode.WriteString("func nativeReadtestGeneratedBID32(function string, rounding int, operands []string) (uint32, readtestSecondaryOutput, string, error) {\n")
	goCode.WriteString("\tswitch function {\n")
	for _, dispatch := range decimalOps32 {
		goCase, err := emitReadtestGoCase("32", "uint32", dispatch, true)
		if err != nil {
			return nil, err
		}
		goCode.WriteString(goCase)
	}
	goCode.WriteString("\tdefault:\n\t\treturn 0, readtestNoSecondaryOutput(), \"\", fmt.Errorf(\"unsupported generated readtest decimal32 function %q\", function)\n\t}\n}\n\n")

	goCode.WriteString("func nativeReadtestGeneratedBID64(function string, rounding int, operands []string) (uint64, readtestSecondaryOutput, string, error) {\n")
	goCode.WriteString("\tswitch function {\n")
	for _, dispatch := range decimalOps64 {
		goCase, err := emitReadtestGoCase("64", "uint64", dispatch, true)
		if err != nil {
			return nil, err
		}
		goCode.WriteString(goCase)
	}
	goCode.WriteString("\tdefault:\n\t\treturn 0, readtestNoSecondaryOutput(), \"\", fmt.Errorf(\"unsupported generated readtest decimal64 function %q\", function)\n\t}\n}\n\n")

	goCode.WriteString("func nativeReadtestGeneratedBID128(function string, rounding int, operands []string) ([16]byte, readtestSecondaryOutput, string, error) {\n")
	goCode.WriteString("\tswitch function {\n")
	for _, dispatch := range decimalOps128 {
		goCase, err := emitReadtestGoCase("128", "[16]byte", dispatch, true)
		if err != nil {
			return nil, err
		}
		goCode.WriteString(goCase)
	}
	goCode.WriteString("\tdefault:\n\t\treturn [16]byte{}, readtestNoSecondaryOutput(), \"\", fmt.Errorf(\"unsupported generated readtest decimal128 function %q\", function)\n\t}\n}\n\n")

	goCode.WriteString("func nativeReadtestGeneratedSigned(function string, rounding int, operands []string) (int64, string, error) {\n")
	goCode.WriteString("\tswitch function {\n")
	for _, dispatch := range scalarSignedOps {
		goCase, err := emitReadtestGoCase("scalar", "int64", dispatch, false)
		if err != nil {
			return nil, err
		}
		goCode.WriteString(goCase)
	}
	goCode.WriteString("\tdefault:\n\t\treturn 0, \"\", fmt.Errorf(\"unsupported generated readtest signed function %q\", function)\n\t}\n}\n\n")

	goCode.WriteString("func nativeReadtestGeneratedBinary32(function string, rounding int, operands []string) (uint32, string, error) {\n")
	goCode.WriteString("\tswitch function {\n")
	for _, dispatch := range scalarBinary32Ops {
		goCase, err := emitReadtestGoFloatCase("32", dispatch)
		if err != nil {
			return nil, err
		}
		goCode.WriteString(goCase)
	}
	goCode.WriteString("\tdefault:\n\t\treturn 0, \"\", fmt.Errorf(\"unsupported generated readtest binary32 function %q\", function)\n\t}\n}\n\n")

	goCode.WriteString("func nativeReadtestGeneratedBinary64(function string, rounding int, operands []string) (uint64, string, error) {\n")
	goCode.WriteString("\tswitch function {\n")
	for _, dispatch := range scalarBinary64Ops {
		goCase, err := emitReadtestGoFloatCase("64", dispatch)
		if err != nil {
			return nil, err
		}
		goCode.WriteString(goCase)
	}
	goCode.WriteString("\tdefault:\n\t\treturn 0, \"\", fmt.Errorf(\"unsupported generated readtest binary64 function %q\", function)\n\t}\n}\n\n")

	goCode.WriteString("func nativeReadtestGeneratedBinary128(function string, rounding int, operands []string) ([16]byte, string, error) {\n")
	goCode.WriteString("\tswitch function {\n")
	for _, dispatch := range scalarBinary128Ops {
		goCase, err := emitReadtestGoCase("128", "[16]byte", dispatch, false)
		if err != nil {
			return nil, err
		}
		goCode.WriteString(goCase)
	}
	goCode.WriteString("\tdefault:\n\t\treturn [16]byte{}, \"\", fmt.Errorf(\"unsupported generated readtest binary128 function %q\", function)\n\t}\n}\n\n")

	goCode.WriteString("func nativeReadtestGeneratedUnsigned(function string, rounding int, operands []string) (uint64, string, error) {\n")
	goCode.WriteString("\tswitch function {\n")
	for _, dispatch := range scalarUnsignedOps {
		var goCase string
		var err error
		if isReadtestStatusControlFunction(dispatch.Function) {
			goCase, err = emitReadtestStatusControlGoCase(dispatch)
		} else {
			goCase, err = emitReadtestGoCase("scalar", "uint64", dispatch, false)
		}
		if err != nil {
			return nil, err
		}
		goCode.WriteString(goCase)
	}
	goCode.WriteString("\tdefault:\n\t\treturn 0, \"\", fmt.Errorf(\"unsupported generated readtest unsigned function %q\", function)\n\t}\n}\n\n")

	if containsReadtestFunction(decimalOps32, "bid32_round_integral_exact") {
		goCode.WriteString("func nativeReadtestBID32RoundIntegralExact(a uint32) uint32 {\n")
		goCode.WriteString("\tresult, _, _, err := nativeReadtestGeneratedBID32(\"bid32_round_integral_exact\", 0, []string{fmt.Sprintf(\"[%08x]\", a)})\n")
		goCode.WriteString("\tif err != nil {\n\t\tpanic(err)\n\t}\n")
		goCode.WriteString("\treturn result\n}\n\n")
	}
	if containsReadtestFunction(decimalOps64, "bid64_round_integral_exact") {
		goCode.WriteString("func nativeReadtestBID64RoundIntegralExact(a uint64) uint64 {\n")
		goCode.WriteString("\tresult, _, _, err := nativeReadtestGeneratedBID64(\"bid64_round_integral_exact\", 0, []string{fmt.Sprintf(\"[%016x]\", a)})\n")
		goCode.WriteString("\tif err != nil {\n\t\tpanic(err)\n\t}\n")
		goCode.WriteString("\treturn result\n}\n\n")
	}

	goCode.WriteString(readtestGeneratedCgoHelpers)

	return append(c.Bytes(), goCode.Bytes()...), nil
}

func generateReadtestStubDispatch() []byte {
	return []byte(genmarker.Line("testgen") + `
//go:build !cgo || !bid754_native

package bid754

import "fmt"

func nativeReadtestBID32FromString(input string, rounding int) (uint32, string) { return 0, "" }
func nativeReadtestBID64FromString(input string, rounding int) (uint64, string) { return 0, "" }
func nativeReadtestBID128FromString(input string, rounding int) ([16]byte, string) { return [16]byte{}, "" }
func nativeReadtestBID32ToString(a uint32) (string, string) { return "", "" }
func nativeReadtestBID64ToString(a uint64) (string, string) { return "", "" }
func nativeReadtestBID128ToString(a [16]byte) (string, string) { return "", "" }
func nativeReadtestGeneratedBID32(function string, rounding int, operands []string) (uint32, readtestSecondaryOutput, string, error) {
	return 0, readtestNoSecondaryOutput(), "", fmt.Errorf("native readtest disabled")
}
func nativeReadtestGeneratedBID64(function string, rounding int, operands []string) (uint64, readtestSecondaryOutput, string, error) {
	return 0, readtestNoSecondaryOutput(), "", fmt.Errorf("native readtest disabled")
}
func nativeReadtestGeneratedBID128(function string, rounding int, operands []string) ([16]byte, readtestSecondaryOutput, string, error) {
	return [16]byte{}, readtestNoSecondaryOutput(), "", fmt.Errorf("native readtest disabled")
}
func nativeReadtestGeneratedSigned(function string, rounding int, operands []string) (int64, string, error) {
	return 0, "", fmt.Errorf("native readtest disabled")
}
func nativeReadtestGeneratedBinary32(function string, rounding int, operands []string) (uint32, string, error) {
	return 0, "", fmt.Errorf("native readtest disabled")
}
func nativeReadtestGeneratedBinary64(function string, rounding int, operands []string) (uint64, string, error) {
	return 0, "", fmt.Errorf("native readtest disabled")
}
func nativeReadtestGeneratedBinary128(function string, rounding int, operands []string) ([16]byte, string, error) {
	return [16]byte{}, "", fmt.Errorf("native readtest disabled")
}
func nativeReadtestGeneratedUnsigned(function string, rounding int, operands []string) (uint64, string, error) {
	return 0, "", fmt.Errorf("native readtest disabled")
}
func nativeReadtestBID32RoundIntegralExact(a uint32) uint32 { return 0 }
func nativeReadtestBID64RoundIntegralExact(a uint64) uint64 { return 0 }
`)
}

// readtestGeneratedCgoHelpers holds the only readtest parse helper that is
// cgo-specific; every backend-independent parse/normalize/compare helper
// lives in the shared generated file (readtestGeneratedSharedHelpers).
const readtestGeneratedCgoHelpers = `
func parseGeneratedReadtestCString(input string) (*C.char, func()) {
	if strings.EqualFold(strings.TrimSpace(input), "NULL") {
		return nil, func() {}
	}
	cstr := C.CString(input)
	return cstr, func() { C.free(unsafe.Pointer(cstr)) }
}
`

// readtestGeneratedSharedHelpers is the backend-independent readtest helper
// surface emitted once as an untagged non-test file so the native cgo gate
// (dispatch and runner), the portable goport gate, and the generated FFI
// bit-compare support all share a single copy of the parse, normalization,
// and Intel readtest.c check_results comparison logic. Backend string
// conversions are injected through readtestStringBackend so each gate keeps
// exercising its own bid*_from_string implementation.
var readtestGeneratedSharedHelpers = genmarker.Line("testgen") + `

package bid754

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"

	bidgo "github.com/sky1core/bid754/bid754-go/internal/bidgo"
)

// readtestStringBackend carries the string conversions of the backend under
// test. The comparison helpers below mirror Intel readtest.c check_results,
// which round-trips decimal literals through the library's own
// bid*_from_string; injecting the functions keeps the native gate on the
// Intel C oracle and the goport gate on the Go mechanical port.
type readtestStringBackend struct {
	FromString32  func(input string, rounding int) (uint32, string)
	FromString64  func(input string, rounding int) (uint64, string)
	FromString128 func(input string, rounding int) ([16]byte, string)
}

const (
	readtestSecondaryNone   = "none"
	readtestSecondaryInt    = "int"
	readtestSecondaryDec32  = "dec32"
	readtestSecondaryDec64  = "dec64"
	readtestSecondaryDec128 = "dec128"
)

// readtestSecondaryOutput carries the extra in/out pointer result of Intel
// readtest functions such as bid*_frexp (int exponent) and bid*_modf
// (decimal integral part). OperandIndex names the readtest row operand that
// pins the expected secondary value, mirroring readtest.c get_ops which
// captures that operand into i1/R32_1/R64_1/R_1 before the call.
type readtestSecondaryOutput struct {
	Kind         string
	OperandIndex int
	Int          int64
	Bits32       uint32
	Bits64       uint64
	Bits128      [16]byte
}

func readtestNoSecondaryOutput() readtestSecondaryOutput {
	return readtestSecondaryOutput{Kind: readtestSecondaryNone, OperandIndex: -1}
}

// readtestSecondaryOutputEqual mirrors the readtest.c check_results secondary
// comparisons: i1 != i2 for the frexp int exponent and R32_1 != B32 /
// R64_1 != B64 / check128(R_1, B) for the modf integral part. The expected
// value is the row operand backing the Intel in/out pointer parameter, parsed
// like the upstream get_ops operand pass (ints via getop32i, decimals via
// getop32/64/128 under the forced round-to-nearest-even of get_ops).
func readtestSecondaryOutputEqual(sec readtestSecondaryOutput, operands []string) (bool, error) {
	if sec.Kind == readtestSecondaryNone {
		return true, nil
	}
	if sec.OperandIndex < 0 || sec.OperandIndex >= len(operands) {
		return false, fmt.Errorf("readtest secondary output operand index %d out of range (%d operands)", sec.OperandIndex, len(operands))
	}
	expected := operands[sec.OperandIndex]
	switch sec.Kind {
	case readtestSecondaryInt:
		want, err := parseReadtestInt(expected)
		if err != nil {
			return false, err
		}
		return want == sec.Int, nil
	case readtestSecondaryDec32:
		want, err := parseReadtestBits32(expected)
		if err != nil {
			return false, err
		}
		return want == sec.Bits32, nil
	case readtestSecondaryDec64:
		want, err := parseReadtestBits64(expected)
		if err != nil {
			return false, err
		}
		return want == sec.Bits64, nil
	case readtestSecondaryDec128:
		want, err := parseReadtestBits128(expected)
		if err != nil {
			return false, err
		}
		return want == sec.Bits128, nil
	default:
		return false, fmt.Errorf("unsupported readtest secondary output kind %q", sec.Kind)
	}
}

func readtestFormatBitWidth(format string) (int, error) {
	switch format {
	case "decimal32":
		return 32, nil
	case "decimal64":
		return 64, nil
	case "decimal128":
		return 128, nil
	default:
		return 0, fmt.Errorf("unsupported readtest format %q", format)
	}
}

// readtestValueBits converts a readtest value field into canonical bits.
// [hex] literals are used directly; decimal literals are converted with the
// backend bid*_from_string at the row rounding mode, mirroring readtest.c
// get_test, which parses the expected literal through getop32/64/128 after
// get_ops has restored the row rounding mode (readtest.c:1018,1032-1105).
// The second return is the conversion status of a decimal-literal parse;
// callers that mirror upstream paths where those flags are reset (the
// pre-call *pfpsf = fpsf_0) discard it.
func readtestValueBits(format, value string, rounding int, backend readtestStringBackend) (string, string, error) {
	if strings.HasPrefix(strings.TrimSpace(value), "[") {
		return value, "00", nil
	}
	switch format {
	case "decimal32":
		raw, status := backend.FromString32(value, rounding)
		return fmt.Sprintf("[%08x]", raw), status, nil
	case "decimal64":
		raw, status := backend.FromString64(value, rounding)
		return fmt.Sprintf("[%016x]", raw), status, nil
	case "decimal128":
		raw, status := backend.FromString128(value, rounding)
		return formatReadtestBits128(raw), status, nil
	default:
		return "", "", fmt.Errorf("unsupported readtest format %q", format)
	}
}

// readtestDecimalRowEqual mirrors readtest.c check_results for decimal-output
// rows whose expected field is a decimal literal: upstream parses the literal
// into R32/R64/R with the library's own from_string at the row rounding mode
// and compares the operation result bits exactly (R32 != Q32, check64,
// check128), so cohort members with different quantum stay distinct.
func readtestDecimalRowEqual(format, expected, gotBits string, rounding int, backend readtestStringBackend) (bool, error) {
	expectedBits, _, err := readtestValueBits(format, expected, rounding, backend)
	if err != nil {
		return false, err
	}
	width, err := readtestFormatBitWidth(format)
	if err != nil {
		return false, err
	}
	return normalizeReadtestBits(expectedBits, width) == normalizeReadtestBits(gotBits, width), nil
}

// readtestToStringRowEqual mirrors readtest.c check_results for the
// bid*_to_string rows: the expected literal and the produced string are both
// converted back with the backend bid*_from_string at the row rounding mode
// and compared as exact bits (readtest.c:1453-1489). The returned status is
// the produced-string conversion status; upstream accumulates those flags
// into *pfpsf before the expected_status comparison, so the caller folds it
// into the operation status with readtestCombineStatus. The expected-literal
// conversion status is discarded because upstream parses the expected literal
// before the *pfpsf = fpsf_0 reset that precedes the operation call.
func readtestToStringRowEqual(format, expected, got string, rounding int, backend readtestStringBackend) (bool, string, error) {
	expectedBits, _, err := readtestValueBits(format, expected, rounding, backend)
	if err != nil {
		return false, "", err
	}
	gotBits, roundTripStatus, err := readtestValueBits(format, got, rounding, backend)
	if err != nil {
		return false, "", err
	}
	width, err := readtestFormatBitWidth(format)
	if err != nil {
		return false, "", err
	}
	return normalizeReadtestBits(expectedBits, width) == normalizeReadtestBits(gotBits, width), roundTripStatus, nil
}

// readtestCombineStatus ORs two readtest status strings, mirroring the
// upstream *pfpsf accumulation across the operation call and the
// check_results bid*_from_string round-trip (readtest.c:1453-1457).
func readtestCombineStatus(a, b string) (string, error) {
	flagsA, err := strconv.ParseUint(normalizeReadtestStatus(a), 16, 32)
	if err != nil {
		return "", fmt.Errorf("parse readtest status %q: %w", a, err)
	}
	flagsB, err := strconv.ParseUint(normalizeReadtestStatus(b), 16, 32)
	if err != nil {
		return "", fmt.Errorf("parse readtest status %q: %w", b, err)
	}
	return formatReadtestStatus(uint32(flagsA | flagsB)), nil
}

// readtestQuietEqual implements the upstream readtest.c CMP_EQUALSTATUS value
// comparison (check_results): a row passes when the result bits match the
// expected bits exactly (check32/check64/check128) or when the library's own
// quiet_not_equal comparison reports the two values equal, so +0 and -0
// compare equal while NaNs still require an exact bit match. A decimal-literal
// expected value converts to bits through the backend bid*_from_string at the
// row rounding mode, mirroring the upstream getop32/64/128 parse of the
// expected field (get_test runs after get_ops restores the row rounding), and
// the quiet_not_equal call then receives those parsed bits like the upstream
// R32/Q32 values. Flags raised by the expected-literal parse and by the
// comparison call itself are discarded, mirroring the pre-call
// *pfpsf = fpsf_0 reset and the upstream tmp_pfpsf capture before
// bid*_quiet_not_equal runs; the caller still compares the operation status
// exactly. The conversions and the comparison run through the injected
// backend so each gate exercises its own implementation.
func readtestQuietEqual(format, expected, got string, rounding int, backend readtestStringBackend, signedDispatch func(function string, rounding int, operands []string) (int64, string, error)) (bool, error) {
	expectedBits, _, err := readtestValueBits(format, expected, rounding, backend)
	if err != nil {
		return false, err
	}
	gotBits, _, err := readtestValueBits(format, got, rounding, backend)
	if err != nil {
		return false, err
	}
	width, err := readtestFormatBitWidth(format)
	if err != nil {
		return false, err
	}
	if normalizeReadtestBits(expectedBits, width) == normalizeReadtestBits(gotBits, width) {
		return true, nil
	}
	var function string
	switch format {
	case "decimal32":
		function = "bid32_quiet_not_equal"
	case "decimal64":
		function = "bid64_quiet_not_equal"
	case "decimal128":
		function = "bid128_quiet_not_equal"
	default:
		return false, fmt.Errorf("unsupported read format %q", format)
	}
	notEqual, _, err := signedDispatch(function, 0, []string{expectedBits, gotBits})
	if err != nil {
		return false, err
	}
	return notEqual == 0, nil
}

func parseReadtestBits32(input string) (uint32, error) {
	if strings.HasPrefix(input, "[") {
		raw, err := parseReadtestHex(input, 32)
		return uint32(raw), err
	}
	// Intel readtest.c get_ops forces round-to-nearest-even while converting
	// decimal operands. Call the Go mechanical port directly: official operands
	// intentionally include excess precision and range cases, and public raw
	// wrappers deliberately have stronger NaN-payload handling than Intel.
	value, _ := bidgo.Bid32FromStringRaw(input, int(bidgo.RoundNearestEven))
	return value, nil
}

func parseReadtestBits64(input string) (uint64, error) {
	if strings.HasPrefix(input, "[") {
		// Intel getop64 uses sscanf("%016llx"). For a bracketed comma
		// form it consumes the first word and stops at the comma; do not
		// concatenate both words into an overflowing 128-bit token.
		trimmed := strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(input), "["), "]")
		if comma := strings.Index(trimmed, ","); comma >= 0 {
			trimmed = trimmed[:comma]
		}
		return strconv.ParseUint(strings.TrimSpace(trimmed), 16, 64)
	}
	value, _ := bidgo.Bid64FromString(input, int(bidgo.RoundNearestEven))
	return value, nil
}

func parseReadtestBits128(input string) ([16]byte, error) {
	var raw [16]byte
	if strings.HasPrefix(input, "[") {
		literal := strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(input), "["), "]")
		var hiText, loText string
		if strings.Contains(literal, ",") {
			parts := strings.SplitN(literal, ",", 2)
			hiText = strings.TrimSpace(parts[0])
			loText = strings.TrimSpace(parts[1])
		} else {
			compact := strings.ReplaceAll(literal, ",", "")
			if len(compact) > 32 {
				return raw, fmt.Errorf("readtest 128-bit hex literal %q exceeds 32 nybbles", input)
			}
			if len(compact) <= 16 {
				return raw, fmt.Errorf("readtest 128-bit hex literal %q does not contain high and low words", input)
			}
			hiText = compact[:16]
			loText = compact[16:]
		}
		if hiText == "" || loText == "" {
			return raw, fmt.Errorf("readtest 128-bit hex literal %q does not contain high and low words", input)
		}
		if len(hiText) > 16 || len(loText) > 16 {
			return raw, fmt.Errorf("readtest 128-bit hex literal %q exceeds 32 nybbles", input)
		}
		hi, err := strconv.ParseUint(hiText, 16, 64)
		if err != nil {
			return raw, err
		}
		lo, err := strconv.ParseUint(loText, 16, 64)
		if err != nil {
			return raw, err
		}
		binary.LittleEndian.PutUint64(raw[0:8], lo)
		binary.LittleEndian.PutUint64(raw[8:16], hi)
		return raw, nil
	}
	value, _ := bidgo.Bid128FromString(input, int(bidgo.RoundNearestEven))
	return decimal128BIDFromBidgo(value).ToBytes(), nil
}

func parseReadtestHex(input string, bits int) (uint64, error) {
	trimmed := strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(input), "["), "]")
	trimmed = strings.ReplaceAll(trimmed, ",", "")
	return strconv.ParseUint(trimmed, 16, bits)
}

func formatReadtestBits128(raw [16]byte) string {
	lo := binary.LittleEndian.Uint64(raw[0:8])
	hi := binary.LittleEndian.Uint64(raw[8:16])
	return fmt.Sprintf("[%016x%016x]", hi, lo)
}

func parseReadtestInt(input string) (int64, error) {
	prefix, base, err := readtestIntegerPrefix(input)
	if err != nil {
		return 0, err
	}
	if base == 16 {
		value, err := strconv.ParseUint(prefix, 16, 64)
		return int64(value), err
	}
	return strconv.ParseInt(prefix, 10, 64)
}

func parseReadtestUint(input string) (uint64, error) {
	prefix, base, err := readtestIntegerPrefix(input)
	if err != nil {
		return 0, err
	}
	if base == 16 {
		return strconv.ParseUint(prefix, 16, 64)
	}
	if strings.HasPrefix(prefix, "-") {
		magnitude, err := strconv.ParseUint(strings.TrimPrefix(prefix, "-"), 10, 64)
		if err != nil {
			return 0, err
		}
		return uint64(0) - magnitude, nil
	}
	return strconv.ParseUint(strings.TrimPrefix(prefix, "+"), 10, 64)
}

// readtestIntegerPrefix mirrors the lexical prefix consumed by the pinned
// readtest.c getop*i/getop*u scanf calls. A bracket starts a hexadecimal bit
// pattern. A decimal conversion accepts one optional sign and at least one
// digit, then stops at the first non-digit (so "1.0" is integer 1).
func readtestIntegerPrefix(input string) (string, int, error) {
	trimmed := strings.TrimSpace(input)
	if strings.HasPrefix(trimmed, "[") {
		end := 1
		for end < len(trimmed) {
			c := trimmed[end]
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
				break
			}
			end++
		}
		if end == 1 {
			return "", 0, fmt.Errorf("readtest integer %q has no hexadecimal digits", input)
		}
		return trimmed[1:end], 16, nil
	}

	end := 0
	if end < len(trimmed) && (trimmed[end] == '+' || trimmed[end] == '-') {
		end++
	}
	digits := end
	for end < len(trimmed) && trimmed[end] >= '0' && trimmed[end] <= '9' {
		end++
	}
	if end == digits {
		return "", 0, fmt.Errorf("readtest integer %q has no decimal digits", input)
	}
	return trimmed[:end], 10, nil
}

func formatReadtestStatus(flags uint32) string {
	return fmt.Sprintf("%02X", flags&0xFF)
}

// normalizeReadtestBits canonicalizes a readtest bit literal for comparison.
// For 128-bit values every parseable literal — the upstream getop128 comma
// form of any length and the sscanf-style greedy 16-nybble split — is
// reparsed and reformatted to the canonical 32-nybble form, so short comma
// literals such as [0,5] compare equal to full-width results. A literal of 16
// nybbles or fewer without a comma is not a valid 128-bit literal on either
// harness side (it fails the row at parse time), so it falls through to the
// generic zero-trim form. 32/64-bit literals keep the comma-strip and
// zero-trim canonical form.
func normalizeReadtestBits(input string, width int) string {
	trimmed := strings.ToLower(strings.TrimSpace(input))
	if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
		trimmed = strings.TrimSuffix(strings.TrimPrefix(trimmed, "["), "]")
	}
	if width == 128 && (strings.Contains(trimmed, ",") || len(trimmed) > 16) {
		if raw, err := parseReadtestBits128(input); err == nil {
			return formatReadtestBits128(raw)
		}
	}
	trimmed = strings.ReplaceAll(trimmed, ",", "")
	trimmed = strings.TrimLeft(trimmed, "0")
	if trimmed == "" {
		trimmed = "0"
	}
	return "[" + trimmed + "]"
}

func compareReadtestScalarOutput(outputType, expected, got string) bool {
	expected = strings.TrimSpace(strings.TrimPrefix(expected, "+"))
	got = strings.TrimSpace(strings.TrimPrefix(got, "+"))

	switch outputType {
	case "OP_BIN32":
		return normalizeReadtestBits(expected, 32) == normalizeReadtestBits(got, 32)
	case "OP_BIN64":
		return normalizeReadtestBits(expected, 64) == normalizeReadtestBits(got, 64)
	case "OP_BIN128":
		return normalizeReadtestBits(expected, 128) == normalizeReadtestBits(got, 128)
	case "OP_BID_UINT8":
		return normalizeReadtestUnsignedScalar(expected, 8) == normalizeReadtestUnsignedScalar(got, 8)
	case "OP_BID_UINT16":
		return normalizeReadtestUnsignedScalar(expected, 16) == normalizeReadtestUnsignedScalar(got, 16)
	case "OP_BID_UINT32":
		return normalizeReadtestUnsignedScalar(expected, 32) == normalizeReadtestUnsignedScalar(got, 32)
	case "OP_BID_UINT64":
		return normalizeReadtestUnsignedScalar(expected, 64) == normalizeReadtestUnsignedScalar(got, 64)
	case "OP_INT8":
		return normalizeReadtestSignedScalar(expected, 8) == normalizeReadtestSignedScalar(got, 8)
	case "OP_INT16":
		return normalizeReadtestSignedScalar(expected, 16) == normalizeReadtestSignedScalar(got, 16)
	case "OP_INT32":
		return normalizeReadtestSignedScalar(expected, 32) == normalizeReadtestSignedScalar(got, 32)
	case "OP_INT64", "OP_LINT":
		return normalizeReadtestSignedScalar(expected, 64) == normalizeReadtestSignedScalar(got, 64)
	default:
		return expected == got
	}
}

func normalizeReadtestUnsignedScalar(input string, bits int) string {
	value, err := parseReadtestUint(input)
	if err != nil {
		return input
	}
	switch bits {
	case 8:
		return strconv.FormatUint(uint64(uint8(value)), 10)
	case 16:
		return strconv.FormatUint(uint64(uint16(value)), 10)
	case 32:
		return strconv.FormatUint(uint64(uint32(value)), 10)
	case 64:
		return strconv.FormatUint(value, 10)
	default:
		return input
	}
}

func normalizeReadtestSignedScalar(input string, bits int) string {
	value, err := parseReadtestInt(input)
	if err != nil {
		return input
	}
	switch bits {
	case 8:
		return strconv.FormatInt(int64(int8(value)), 10)
	case 16:
		return strconv.FormatInt(int64(int16(value)), 10)
	case 32:
		return strconv.FormatInt(int64(int32(value)), 10)
	case 64:
		return strconv.FormatInt(value, 10)
	default:
		return input
	}
}

func isReadtestScalarOutput(output string) bool {
	switch output {
	case "OP_BIN32", "OP_BIN64", "OP_BIN128",
		"OP_BID_UINT8", "OP_BID_UINT16", "OP_BID_UINT32", "OP_BID_UINT64",
		"OP_INT8", "OP_INT16", "OP_INT32", "OP_INT64", "OP_LINT":
		return true
	default:
		return false
	}
}

func normalizeReadtestStatus(input string) string {
	trimmed := strings.ToUpper(strings.TrimSpace(input))
	trimmed = strings.TrimPrefix(trimmed, "0X")
	if trimmed == "" {
		return "00"
	}
	if len(trimmed)%2 != 0 {
		trimmed = "0" + trimmed
	}
	return trimmed
}
`

func containsReadtestFunction(values []readtestDispatchSpec, want string) bool {
	for _, value := range values {
		if value.Function == want {
			return true
		}
	}
	return false
}

func hasReadtestParam(values []readtestParamKind, want readtestParamKind) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func isReadtestStatusControlFunction(name string) bool {
	switch name {
	case "bid_testFlags",
		"bid_lowerFlags",
		"bid_signalException",
		"bid_saveFlags",
		"bid_restoreFlags",
		"bid_testSavedFlags",
		"bid_getDecimalRoundingDirection",
		"bid_setDecimalRoundingDirection":
		return true
	default:
		return false
	}
}

func isReadtestDecimalOutput(output string) bool {
	return output == "OP_DEC32" || output == "OP_DEC64" || output == "OP_DEC128"
}

func isReadtestUnsignedOutput(output string) bool {
	switch output {
	case "OP_BID_UINT8", "OP_BID_UINT16", "OP_BID_UINT32", "OP_BID_UINT64":
		return true
	default:
		return false
	}
}

func isReadtestBinary32Output(output string) bool {
	return output == "OP_BIN32"
}

func isReadtestBinary64Output(output string) bool {
	return output == "OP_BIN64"
}

func isReadtestBinary128Output(output string) bool {
	return output == "OP_BIN128"
}

func classifyReadtestParameters(params []string) ([]readtestParamKind, error) {
	kinds := make([]readtestParamKind, 0, len(params))
	for _, param := range params {
		normalized := strings.Join(strings.Fields(param), " ")
		switch {
		case strings.Contains(normalized, "char *ps"):
			kinds = append(kinds, readtestParamCharPs)
		case strings.Contains(normalized, "const char *") || strings.Contains(normalized, "const char*"):
			kinds = append(kinds, readtestParamCStr)
		case strings.Contains(normalized, "int*"):
			kinds = append(kinds, readtestParamPInt)
		case strings.Contains(normalized, "BID_UINT32 *") || strings.Contains(normalized, "BID_UINT32*"):
			kinds = append(kinds, readtestParamP32)
		case strings.Contains(normalized, "BID_UINT64 *") || strings.Contains(normalized, "BID_UINT64*"):
			kinds = append(kinds, readtestParamP64)
		case strings.Contains(normalized, "BID_UINT128 *") || strings.Contains(normalized, "BID_UINT128*"):
			kinds = append(kinds, readtestParamP128)
		case strings.Contains(normalized, "long int"):
			kinds = append(kinds, readtestParamLInt)
		case strings.Contains(normalized, "BID_SINT64"):
			kinds = append(kinds, readtestParamS64)
		case strings.Contains(normalized, "_IDEC_round"):
			kinds = append(kinds, readtestParamRound)
		case strings.Contains(normalized, "_IDEC_flags *") || strings.Contains(normalized, "_IDEC_flags*"):
			kinds = append(kinds, readtestParamFlags)
		case strings.Contains(normalized, "_IDEC_exceptionmasks"):
			kinds = append(kinds, readtestParamMasks)
		case strings.Contains(normalized, "_IDEC_exceptioninfo"):
			kinds = append(kinds, readtestParamInfo)
		case normalized == "unsigned int x" || normalized == "unsigned int":
			kinds = append(kinds, readtestParamUInt)
		case normalized == "int x" || normalized == "int n" || normalized == "int":
			kinds = append(kinds, readtestParamInt)
		case strings.Contains(normalized, "BID_UINT32"):
			kinds = append(kinds, readtestParamU32)
		case strings.Contains(normalized, "BID_UINT64"):
			kinds = append(kinds, readtestParamU64)
		case strings.Contains(normalized, "BID_UINT128"):
			kinds = append(kinds, readtestParamU128)
		default:
			return nil, fmt.Errorf("unsupported symbol parameter %q", param)
		}
	}
	return kinds, nil
}

func emitReadtestStatusControlCWrapper(dispatch readtestDispatchSpec) (string, error) {
	var buf bytes.Buffer
	switch dispatch.Function {
	case "bid_testFlags":
		buf.WriteString("static _IDEC_flags bid754_generated_readtest_bid_testFlags(_IDEC_flags flagsmask, _IDEC_flags initial_flags, _IDEC_flags* out_flags) {\n")
		buf.WriteString("\t_IDEC_flags flags = initial_flags & BID_FLAG_MASK;\n")
		buf.WriteString("\t_IDEC_flags result = bid_testFlags(flagsmask, &flags);\n")
		buf.WriteString("\tif (out_flags) { *out_flags = flags; }\n")
		buf.WriteString("\treturn result;\n")
		buf.WriteString("}\n")
	case "bid_lowerFlags":
		buf.WriteString("static _IDEC_flags bid754_generated_readtest_bid_lowerFlags(_IDEC_flags flagsmask, _IDEC_flags initial_flags, _IDEC_flags* out_flags) {\n")
		buf.WriteString("\t_IDEC_flags flags = initial_flags & BID_FLAG_MASK;\n")
		buf.WriteString("\tbid_lowerFlags(flagsmask, &flags);\n")
		buf.WriteString("\tif (out_flags) { *out_flags = flags; }\n")
		buf.WriteString("\treturn 0;\n")
		buf.WriteString("}\n")
	case "bid_signalException":
		buf.WriteString("static _IDEC_flags bid754_generated_readtest_bid_signalException(_IDEC_flags flagsmask, _IDEC_flags initial_flags, _IDEC_flags* out_flags) {\n")
		buf.WriteString("\t_IDEC_flags flags = initial_flags & BID_FLAG_MASK;\n")
		buf.WriteString("\tbid_signalException(flagsmask, &flags);\n")
		buf.WriteString("\tif (out_flags) { *out_flags = flags; }\n")
		buf.WriteString("\treturn 0;\n")
		buf.WriteString("}\n")
	case "bid_saveFlags":
		buf.WriteString("static _IDEC_flags bid754_generated_readtest_bid_saveFlags(_IDEC_flags flagsmask, _IDEC_flags initial_flags, _IDEC_flags* out_flags) {\n")
		buf.WriteString("\t_IDEC_flags flags = initial_flags & BID_FLAG_MASK;\n")
		buf.WriteString("\t_IDEC_flags result = bid_saveFlags(flagsmask, &flags);\n")
		buf.WriteString("\tif (out_flags) { *out_flags = flags; }\n")
		buf.WriteString("\treturn result;\n")
		buf.WriteString("}\n")
	case "bid_restoreFlags":
		buf.WriteString("static _IDEC_flags bid754_generated_readtest_bid_restoreFlags(_IDEC_flags flagsvalues, _IDEC_flags flagsmask, _IDEC_flags initial_flags, _IDEC_flags* out_flags) {\n")
		buf.WriteString("\t_IDEC_flags flags = initial_flags & BID_FLAG_MASK;\n")
		buf.WriteString("\tbid_restoreFlags(flagsvalues, flagsmask, &flags);\n")
		buf.WriteString("\tif (out_flags) { *out_flags = flags; }\n")
		buf.WriteString("\treturn 0;\n")
		buf.WriteString("}\n")
	case "bid_testSavedFlags":
		buf.WriteString("static _IDEC_flags bid754_generated_readtest_bid_testSavedFlags(_IDEC_flags savedflags, _IDEC_flags flagsmask, _IDEC_flags* out_flags) {\n")
		buf.WriteString("\t_IDEC_flags result = bid_testSavedFlags(savedflags, flagsmask);\n")
		buf.WriteString("\tif (out_flags) { *out_flags = 0; }\n")
		buf.WriteString("\treturn result;\n")
		buf.WriteString("}\n")
	case "bid_getDecimalRoundingDirection":
		buf.WriteString("static _IDEC_round bid754_generated_readtest_bid_getDecimalRoundingDirection(int rounding_mode, _IDEC_flags* out_flags) {\n")
		buf.WriteString("\t_IDEC_round result = bid_getDecimalRoundingDirection((_IDEC_round)rounding_mode);\n")
		buf.WriteString("\tif (out_flags) { *out_flags = 0; }\n")
		buf.WriteString("\treturn result;\n")
		buf.WriteString("}\n")
	case "bid_setDecimalRoundingDirection":
		buf.WriteString("static _IDEC_round bid754_generated_readtest_bid_setDecimalRoundingDirection(_IDEC_round requested_mode, int rounding_mode, _IDEC_flags* out_flags) {\n")
		buf.WriteString("\t_IDEC_round result = bid_setDecimalRoundingDirection(requested_mode, (_IDEC_round)rounding_mode);\n")
		buf.WriteString("\tif (out_flags) { *out_flags = 0; }\n")
		buf.WriteString("\treturn result;\n")
		buf.WriteString("}\n")
	default:
		return "", fmt.Errorf("unsupported status control readtest function %q", dispatch.Function)
	}
	return buf.String(), nil
}

func emitReadtestCWrapper(returnCType string, dispatch readtestDispatchSpec, withSecondary bool) (string, error) {
	params, callArgs, setup, secondary, err := readtestCWrapperParts(dispatch.ParamKinds)
	if err != nil {
		return "", fmt.Errorf("%s: %w", dispatch.Function, err)
	}
	if secondary != nil && !withSecondary {
		return "", fmt.Errorf("%s: pointer output parameter is only supported in the decimal dispatch groups", dispatch.Function)
	}
	returnCType = readtestCWrapperReturnType(returnCType, dispatch.Symbol.ReturnType)
	var buf bytes.Buffer
	if secondary != nil {
		params = append(params, secondary.CParam)
	}
	params = append(params, "_IDEC_flags* out_flags")
	buf.WriteString(fmt.Sprintf("static %s bid754_generated_readtest_%s(%s) {\n", returnCType, dispatch.Function, strings.Join(params, ", ")))
	for _, line := range setup {
		buf.WriteString("\t")
		buf.WriteString(line)
		buf.WriteString("\n")
	}
	buf.WriteString(fmt.Sprintf("\t%s result = %s(%s);\n", returnCType, dispatch.Function, strings.Join(callArgs, ", ")))
	if secondary != nil {
		buf.WriteString("\t")
		buf.WriteString(secondary.Writeback)
		buf.WriteString("\n")
	}
	if hasReadtestParam(dispatch.ParamKinds, readtestParamFlags) {
		buf.WriteString("\tif (out_flags) { *out_flags = flags; }\n")
	} else {
		buf.WriteString("\tif (out_flags) { *out_flags = 0; }\n")
	}
	buf.WriteString("\treturn result;\n")
	buf.WriteString("}\n")
	return buf.String(), nil
}

func readtestCWrapperReturnType(defaultType, symbolReturnType string) string {
	switch symbolReturnType {
	case "BID_UINT32", "BID_UINT64":
		return symbolReturnType
	default:
		return defaultType
	}
}

func emitReadtestCWrapper128(dispatch readtestDispatchSpec, withSecondary bool) (string, error) {
	params, callArgs, setup, secondary, err := readtestCWrapperParts(dispatch.ParamKinds)
	if err != nil {
		return "", fmt.Errorf("%s: %w", dispatch.Function, err)
	}
	if secondary != nil && !withSecondary {
		return "", fmt.Errorf("%s: pointer output parameter is only supported in the decimal dispatch groups", dispatch.Function)
	}
	var buf bytes.Buffer
	params = append([]string{"unsigned char out_result[16]"}, params...)
	if secondary != nil {
		params = append(params, secondary.CParam)
	}
	params = append(params, "_IDEC_flags* out_flags")
	buf.WriteString(fmt.Sprintf("static void bid754_generated_readtest_%s(%s) {\n", dispatch.Function, strings.Join(params, ", ")))
	for _, line := range setup {
		buf.WriteString("\t")
		buf.WriteString(line)
		buf.WriteString("\n")
	}
	buf.WriteString(fmt.Sprintf("\tBID_UINT128 result = %s(%s);\n", dispatch.Function, strings.Join(callArgs, ", ")))
	buf.WriteString("\tmemcpy(out_result, &result, 16);\n")
	if secondary != nil {
		buf.WriteString("\t")
		buf.WriteString(secondary.Writeback)
		buf.WriteString("\n")
	}
	if hasReadtestParam(dispatch.ParamKinds, readtestParamFlags) {
		buf.WriteString("\tif (out_flags) { *out_flags = flags; }\n")
	} else {
		buf.WriteString("\tif (out_flags) { *out_flags = 0; }\n")
	}
	buf.WriteString("}\n")
	return buf.String(), nil
}

// readtestCWrapperSecondary describes the single Intel in/out pointer
// parameter of a wrapped function (frexp exponent, modf integral part): the
// C wrapper exposes it as an extra out parameter so the runner can compare it
// like readtest.c check_results does (i1 != i2, R32_1 != B32, R64_1 != B64,
// check128(R_1, B)).
type readtestCWrapperSecondary struct {
	Kind         string // matches the shared readtestSecondary* constants
	OperandIndex int
	CParam       string
	Writeback    string
}

func setReadtestCWrapperSecondary(secondary **readtestCWrapperSecondary, value readtestCWrapperSecondary) error {
	if *secondary != nil {
		return fmt.Errorf("multiple pointer output parameters are not supported")
	}
	*secondary = &value
	return nil
}

func readtestCWrapperParts(kinds []readtestParamKind) ([]string, []string, []string, *readtestCWrapperSecondary, error) {
	params := []string{}
	callArgs := []string{}
	setup := []string{}
	var (
		valueIndex    int
		flagsDeclared bool
		secondary     *readtestCWrapperSecondary
	)
	for _, kind := range kinds {
		switch kind {
		case readtestParamU32:
			name := fmt.Sprintf("v%d", valueIndex)
			params = append(params, fmt.Sprintf("BID_UINT32 %s", name))
			callArgs = append(callArgs, name)
			valueIndex++
		case readtestParamU64:
			name := fmt.Sprintf("v%d", valueIndex)
			params = append(params, fmt.Sprintf("BID_UINT64 %s", name))
			callArgs = append(callArgs, name)
			valueIndex++
		case readtestParamU128:
			name := fmt.Sprintf("v%d", valueIndex)
			params = append(params, fmt.Sprintf("const unsigned char* %s", name))
			setup = append(setup, fmt.Sprintf("BID_UINT128 in_%s;", name))
			setup = append(setup, fmt.Sprintf("memcpy(&in_%s, %s, 16);", name, name))
			callArgs = append(callArgs, fmt.Sprintf("in_%s", name))
			valueIndex++
		case readtestParamP32:
			name := fmt.Sprintf("v%d", valueIndex)
			params = append(params, fmt.Sprintf("BID_UINT32 %s", name))
			setup = append(setup, fmt.Sprintf("BID_UINT32 out_%s = %s;", name, name))
			callArgs = append(callArgs, fmt.Sprintf("&out_%s", name))
			if err := setReadtestCWrapperSecondary(&secondary, readtestCWrapperSecondary{
				Kind:         "dec32",
				OperandIndex: valueIndex,
				CParam:       "BID_UINT32* out_second",
				Writeback:    fmt.Sprintf("if (out_second) { *out_second = out_%s; }", name),
			}); err != nil {
				return nil, nil, nil, nil, err
			}
			valueIndex++
		case readtestParamP64:
			name := fmt.Sprintf("v%d", valueIndex)
			params = append(params, fmt.Sprintf("BID_UINT64 %s", name))
			setup = append(setup, fmt.Sprintf("BID_UINT64 out_%s = %s;", name, name))
			callArgs = append(callArgs, fmt.Sprintf("&out_%s", name))
			if err := setReadtestCWrapperSecondary(&secondary, readtestCWrapperSecondary{
				Kind:         "dec64",
				OperandIndex: valueIndex,
				CParam:       "BID_UINT64* out_second",
				Writeback:    fmt.Sprintf("if (out_second) { *out_second = out_%s; }", name),
			}); err != nil {
				return nil, nil, nil, nil, err
			}
			valueIndex++
		case readtestParamP128:
			name := fmt.Sprintf("v%d", valueIndex)
			params = append(params, fmt.Sprintf("unsigned char* %s", name))
			setup = append(setup, fmt.Sprintf("BID_UINT128 out_%s;", name))
			setup = append(setup, fmt.Sprintf("memcpy(&out_%s, %s, 16);", name, name))
			callArgs = append(callArgs, fmt.Sprintf("&out_%s", name))
			if err := setReadtestCWrapperSecondary(&secondary, readtestCWrapperSecondary{
				Kind:         "dec128",
				OperandIndex: valueIndex,
				CParam:       "unsigned char* out_second",
				Writeback:    fmt.Sprintf("if (out_second) { memcpy(out_second, &out_%s, 16); }", name),
			}); err != nil {
				return nil, nil, nil, nil, err
			}
			valueIndex++
		case readtestParamPInt:
			name := fmt.Sprintf("v%d", valueIndex)
			params = append(params, fmt.Sprintf("int %s", name))
			// Intel readtest passes a zero-initialized int out for the frexp
			// exponent (get_test initializes i1 = i2 = 0 and calls with &i2);
			// the row operand only pins the expected value, so seeding the
			// out slot with it would let a code path that never writes the
			// output pass trivially. The decimal in/out pointers (P32/P64/
			// P128) keep the parsed-operand seed because Intel getop32/64/128
			// loads B32/B64/B from the operand before the call.
			setup = append(setup, fmt.Sprintf("int out_%s = 0;", name))
			setup = append(setup, fmt.Sprintf("(void)%s;", name))
			callArgs = append(callArgs, fmt.Sprintf("&out_%s", name))
			if err := setReadtestCWrapperSecondary(&secondary, readtestCWrapperSecondary{
				Kind:         "int",
				OperandIndex: valueIndex,
				CParam:       "int* out_second",
				Writeback:    fmt.Sprintf("if (out_second) { *out_second = out_%s; }", name),
			}); err != nil {
				return nil, nil, nil, nil, err
			}
			valueIndex++
		case readtestParamCStr:
			params = append(params, "const char* v0")
			callArgs = append(callArgs, "v0")
			valueIndex++
		case readtestParamInt:
			name := fmt.Sprintf("v%d", valueIndex)
			params = append(params, fmt.Sprintf("int %s", name))
			callArgs = append(callArgs, name)
			valueIndex++
		case readtestParamUInt:
			name := fmt.Sprintf("v%d", valueIndex)
			params = append(params, fmt.Sprintf("unsigned int %s", name))
			callArgs = append(callArgs, name)
			valueIndex++
		case readtestParamLInt:
			name := fmt.Sprintf("v%d", valueIndex)
			params = append(params, fmt.Sprintf("long %s", name))
			callArgs = append(callArgs, name)
			valueIndex++
		case readtestParamS64:
			name := fmt.Sprintf("v%d", valueIndex)
			params = append(params, fmt.Sprintf("long long %s", name))
			callArgs = append(callArgs, name)
			valueIndex++
		case readtestParamRound:
			params = append(params, "int rounding_mode")
			callArgs = append(callArgs, "(_IDEC_round)rounding_mode")
		case readtestParamFlags:
			if !flagsDeclared {
				setup = append(setup, "_IDEC_flags flags = 0;")
				flagsDeclared = true
			}
			callArgs = append(callArgs, "&flags")
		case readtestParamMasks, readtestParamInfo, readtestParamCharPs:
			return nil, nil, nil, nil, fmt.Errorf("unsupported wrapper parameter kind %q", kind)
		default:
			return nil, nil, nil, nil, fmt.Errorf("unsupported wrapper parameter kind %q", kind)
		}
	}
	return params, callArgs, setup, secondary, nil
}

func emitReadtestStatusControlGoCase(dispatch readtestDispatchSpec) (string, error) {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("\tcase %q:\n", dispatch.Function))
	buf.WriteString(fmt.Sprintf("\t\tif len(operands) != %d {\n", len(dispatch.InputTypes)))
	buf.WriteString(fmt.Sprintf("\t\t\treturn 0, \"\", fmt.Errorf(\"%s expects %d operands, got %%d\", len(operands))\n", dispatch.Function, len(dispatch.InputTypes)))
	buf.WriteString("\t\t}\n")
	for i := range dispatch.InputTypes {
		buf.WriteString(fmt.Sprintf("\t\targ%dRaw, err := parseReadtestUint(operands[%d])\n", i, i))
		buf.WriteString("\t\tif err != nil {\n")
		buf.WriteString("\t\t\treturn 0, \"\", err\n")
		buf.WriteString("\t\t}\n")
	}
	if dispatch.Function == "bid_getDecimalRoundingDirection" {
		buf.WriteString("\t\t_ = arg0Raw\n")
	}
	buf.WriteString("\t\tvar flags C._IDEC_flags\n")
	switch dispatch.Function {
	case "bid_testFlags",
		"bid_lowerFlags",
		"bid_signalException",
		"bid_saveFlags",
		"bid_testSavedFlags":
		buf.WriteString(fmt.Sprintf("\t\tresult := uint64(C.bid754_generated_readtest_%s(C._IDEC_flags(arg0Raw), C._IDEC_flags(arg1Raw), &flags))\n", dispatch.Function))
	case "bid_restoreFlags":
		buf.WriteString(fmt.Sprintf("\t\tresult := uint64(C.bid754_generated_readtest_%s(C._IDEC_flags(arg0Raw), C._IDEC_flags(arg1Raw), C._IDEC_flags(arg2Raw), &flags))\n", dispatch.Function))
	case "bid_getDecimalRoundingDirection":
		buf.WriteString(fmt.Sprintf("\t\tresult := uint64(C.bid754_generated_readtest_%s(C.int(rounding), &flags))\n", dispatch.Function))
	case "bid_setDecimalRoundingDirection":
		buf.WriteString(fmt.Sprintf("\t\tresult := uint64(C.bid754_generated_readtest_%s(C._IDEC_round(arg0Raw), C.int(rounding), &flags))\n", dispatch.Function))
	default:
		return "", fmt.Errorf("unsupported status control readtest function %q", dispatch.Function)
	}
	buf.WriteString("\t\treturn result, formatReadtestStatus(uint32(flags)), nil\n")
	return buf.String(), nil
}

// readtestGoSecondaryEmit maps a C wrapper secondary output to the cgo
// declaration, call argument, and readtestSecondaryOutput literal used by the
// generated Go dispatch case.
func readtestGoSecondaryEmit(secondary *readtestCWrapperSecondary) (decl, callArg, secLiteral string, err error) {
	switch secondary.Kind {
	case "int":
		return "var second C.int", "&second",
			fmt.Sprintf("readtestSecondaryOutput{Kind: readtestSecondaryInt, OperandIndex: %d, Int: int64(second)}", secondary.OperandIndex), nil
	case "dec32":
		return "var second C.BID_UINT32", "&second",
			fmt.Sprintf("readtestSecondaryOutput{Kind: readtestSecondaryDec32, OperandIndex: %d, Bits32: uint32(second)}", secondary.OperandIndex), nil
	case "dec64":
		return "var second C.BID_UINT64", "&second",
			fmt.Sprintf("readtestSecondaryOutput{Kind: readtestSecondaryDec64, OperandIndex: %d, Bits64: uint64(second)}", secondary.OperandIndex), nil
	case "dec128":
		return "var second [16]byte", "(*C.uchar)(unsafe.Pointer(&second[0]))",
			fmt.Sprintf("readtestSecondaryOutput{Kind: readtestSecondaryDec128, OperandIndex: %d, Bits128: second}", secondary.OperandIndex), nil
	default:
		return "", "", "", fmt.Errorf("unsupported secondary output kind %q", secondary.Kind)
	}
}

func emitReadtestGoCase(bits string, goType string, dispatch readtestDispatchSpec, withSecondary bool) (string, error) {
	var buf bytes.Buffer
	operandCount := len(dispatch.InputTypes)
	if operandCount == 0 {
		operandCount = readtestValueOperandCount(dispatch.ParamKinds)
	}
	zeroValue := "0"
	if bits == "128" {
		zeroValue = "[16]byte{}"
	}
	_, _, _, secondary, err := readtestCWrapperParts(dispatch.ParamKinds)
	if err != nil {
		return "", fmt.Errorf("%s: %w", dispatch.Function, err)
	}
	if secondary != nil && !withSecondary {
		return "", fmt.Errorf("%s: pointer output parameter is only supported in the decimal dispatch groups", dispatch.Function)
	}
	zeroReturn := zeroValue
	if withSecondary {
		zeroReturn = zeroValue + ", readtestNoSecondaryOutput()"
	}
	lines, callArgs, needsCleanup, err := readtestGoCallParts(bits, dispatch.InputTypes, dispatch.ParamKinds, zeroReturn)
	if err != nil {
		return "", fmt.Errorf("%s: %w", dispatch.Function, err)
	}
	buf.WriteString(fmt.Sprintf("\tcase %q:\n", dispatch.Function))
	buf.WriteString(fmt.Sprintf("\t\tif len(operands) != %d {\n", operandCount))
	buf.WriteString(fmt.Sprintf("\t\t\treturn %s, \"\", fmt.Errorf(\"%s expects %d operands, got %%d\", len(operands))\n", zeroReturn, dispatch.Function, operandCount))
	buf.WriteString("\t\t}\n")
	for _, line := range lines {
		buf.WriteString("\t\t")
		buf.WriteString(line)
		buf.WriteString("\n")
	}
	if needsCleanup {
		buf.WriteString("\t\tdefer cleanup()\n")
	}
	buf.WriteString("\t\tvar flags C._IDEC_flags\n")
	secExpr := ""
	if withSecondary {
		secExpr = "readtestNoSecondaryOutput()"
	}
	if secondary != nil {
		decl, callArg, secLiteral, err := readtestGoSecondaryEmit(secondary)
		if err != nil {
			return "", fmt.Errorf("%s: %w", dispatch.Function, err)
		}
		buf.WriteString("\t\t")
		buf.WriteString(decl)
		buf.WriteString("\n")
		callArgs = append(callArgs, callArg)
		secExpr = secLiteral
	}
	callArgs = append(callArgs, "&flags")
	returnTuple := "result"
	if withSecondary {
		returnTuple = "result, " + secExpr
	}
	if bits == "128" {
		buf.WriteString("\t\tvar result [16]byte\n")
		buf.WriteString(fmt.Sprintf("\t\tC.bid754_generated_readtest_%s((*C.uchar)(unsafe.Pointer(&result[0])), %s)\n", dispatch.Function, strings.Join(callArgs, ", ")))
		buf.WriteString(fmt.Sprintf("\t\treturn %s, formatReadtestStatus(uint32(flags)), nil\n", returnTuple))
	} else {
		buf.WriteString(fmt.Sprintf("\t\tresult := %s(C.bid754_generated_readtest_%s(%s))\n", goType, dispatch.Function, strings.Join(callArgs, ", ")))
		buf.WriteString(fmt.Sprintf("\t\treturn %s, formatReadtestStatus(uint32(flags)), nil\n", returnTuple))
	}
	return buf.String(), nil
}

func emitReadtestGoFloatCase(bits string, dispatch readtestDispatchSpec) (string, error) {
	var buf bytes.Buffer
	operandCount := len(dispatch.InputTypes)
	if operandCount == 0 {
		operandCount = readtestValueOperandCount(dispatch.ParamKinds)
	}
	lines, callArgs, needsCleanup, err := readtestGoCallParts("scalar", dispatch.InputTypes, dispatch.ParamKinds, "0")
	if err != nil {
		return "", fmt.Errorf("%s: %w", dispatch.Function, err)
	}
	buf.WriteString(fmt.Sprintf("\tcase %q:\n", dispatch.Function))
	buf.WriteString(fmt.Sprintf("\t\tif len(operands) != %d {\n", operandCount))
	buf.WriteString(fmt.Sprintf("\t\t\treturn 0, \"\", fmt.Errorf(\"%s expects %d operands, got %%d\", len(operands))\n", dispatch.Function, operandCount))
	buf.WriteString("\t\t}\n")
	for _, line := range lines {
		buf.WriteString("\t\t")
		buf.WriteString(line)
		buf.WriteString("\n")
	}
	if needsCleanup {
		buf.WriteString("\t\tdefer cleanup()\n")
	}
	buf.WriteString("\t\tvar flags C._IDEC_flags\n")
	callArgs = append(callArgs, "&flags")
	switch bits {
	case "32":
		buf.WriteString(fmt.Sprintf("\t\tresult := math.Float32bits(float32(C.bid754_generated_readtest_%s(%s)))\n", dispatch.Function, strings.Join(callArgs, ", ")))
	case "64":
		buf.WriteString(fmt.Sprintf("\t\tresult := math.Float64bits(float64(C.bid754_generated_readtest_%s(%s)))\n", dispatch.Function, strings.Join(callArgs, ", ")))
	default:
		return "", fmt.Errorf("%s: unsupported float bits %q", dispatch.Function, bits)
	}
	buf.WriteString("\t\treturn result, formatReadtestStatus(uint32(flags)), nil\n")
	return buf.String(), nil
}

func readtestValueOperandCount(kinds []readtestParamKind) int {
	count := 0
	for _, kind := range kinds {
		switch kind {
		case readtestParamU32, readtestParamU64, readtestParamU128, readtestParamP32, readtestParamP64, readtestParamP128, readtestParamPInt, readtestParamCStr, readtestParamInt, readtestParamUInt, readtestParamLInt, readtestParamS64:
			count++
		}
	}
	return count
}

func readtestGoOperandParser(kind readtestParamKind, inputType string) (parser string, caster string, err error) {
	switch inputType {
	case "OP_DEC32":
		return "parseReadtestBits32", "C.BID_UINT32", nil
	case "OP_DEC64":
		return "parseReadtestBits64", "C.BID_UINT64", nil
	case "OP_DEC128":
		return "parseReadtestBits128", "", nil
	case "OP_INT8", "OP_INT16", "OP_INT32":
		return "parseReadtestInt", "C.int", nil
	case "OP_INT64":
		if kind == readtestParamS64 {
			return "parseReadtestInt", "C.longlong", nil
		}
		return "parseReadtestInt", "C.longlong", nil
	case "OP_LINT":
		return "parseReadtestInt", "C.long", nil
	case "OP_BID_UINT8", "OP_BID_UINT16", "OP_BID_UINT32":
		switch kind {
		case readtestParamInt, readtestParamPInt:
			return "parseReadtestUint", "C.int", nil
		case readtestParamUInt:
			return "parseReadtestUint", "C.uint", nil
		default:
			return "parseReadtestUint", "C.BID_UINT32", nil
		}
	case "OP_BID_UINT64":
		switch kind {
		case readtestParamInt, readtestParamPInt:
			return "parseReadtestUint", "C.int", nil
		case readtestParamUInt:
			return "parseReadtestUint", "C.uint", nil
		default:
			return "parseReadtestUint", "C.BID_UINT64", nil
		}
	default:
		return "", "", fmt.Errorf("unsupported readtest input type %q", inputType)
	}
}

func readtestGoCallParts(bits string, inputTypes []string, kinds []readtestParamKind, zeroValue string) ([]string, []string, bool, error) {
	lines := []string{}
	args := []string{}
	operandIndex := 0
	needsCleanup := false
	for _, kind := range kinds {
		switch kind {
		case readtestParamU32, readtestParamU64, readtestParamU128, readtestParamP32, readtestParamP64, readtestParamP128, readtestParamInt, readtestParamUInt, readtestParamLInt, readtestParamS64, readtestParamPInt:
			if operandIndex >= len(inputTypes) {
				return nil, nil, false, fmt.Errorf("missing input type for operand %d", operandIndex)
			}
			name := fmt.Sprintf("arg%dRaw", operandIndex)
			parser, caster, err := readtestGoOperandParser(kind, inputTypes[operandIndex])
			if err != nil {
				return nil, nil, false, err
			}
			lines = append(lines, fmt.Sprintf("%s, err := %s(operands[%d])", name, parser, operandIndex))
			lines = append(lines, "if err != nil {")
			lines = append(lines, fmt.Sprintf("\treturn %s, \"\", err", zeroValue))
			lines = append(lines, "}")
			switch kind {
			case readtestParamU128, readtestParamP128:
				args = append(args, fmt.Sprintf("(*C.uchar)(unsafe.Pointer(&%s[0]))", name))
			default:
				args = append(args, fmt.Sprintf("%s(%s)", caster, name))
			}
			operandIndex++
		case readtestParamCStr:
			lines = append(lines, fmt.Sprintf("arg%d, cleanup := parseGeneratedReadtestCString(operands[%d])", operandIndex, operandIndex))
			args = append(args, fmt.Sprintf("arg%d", operandIndex))
			operandIndex++
			needsCleanup = true
		case readtestParamRound:
			args = append(args, "C.int(rounding)")
		case readtestParamFlags:
		default:
			return nil, nil, false, fmt.Errorf("unsupported Go call parameter kind %q", kind)
		}
	}
	return lines, args, needsCleanup, nil
}
