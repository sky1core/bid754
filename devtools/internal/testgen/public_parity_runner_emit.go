package testgen

import (
	"fmt"
	"sort"
	"strings"

	"github.com/sky1core/bid754/devtools/internal/genmarker"
)

// emitPublicParityFiles renders the parity dispatch and cases test files from
// the resolved units, the bit-literal corpus, and the pinned exception masks.
func emitPublicParityFiles(units []parityUnit, corpus publicParityCorpus, masks publicParityExceptionMasks) ([]byte, []byte, error) {
	dispatch, err := emitPublicParityDispatch(units, corpus, masks)
	if err != nil {
		return nil, nil, err
	}
	cases := emitPublicParityCases(units)
	return dispatch, cases, nil
}

func emitPublicParityDispatch(units []parityUnit, corpus publicParityCorpus, masks publicParityExceptionMasks) ([]byte, error) {
	var b strings.Builder
	b.WriteString(genmarker.Line("testgen") + "\n\n")
	b.WriteString("package bid754\n\n")
	b.WriteString("import (\n")
	b.WriteString("\t\"math\"\n")
	b.WriteString("\t\"testing\"\n")
	b.WriteString("\t\"unsafe\"\n\n")
	b.WriteString("\t\"github.com/sky1core/bid754/bid754-go/internal/bidgo\"\n")
	b.WriteString(")\n\n")

	emitPublicParityStaticHelpers(&b, masks)
	emitPublicParityCorpusTables(&b, corpus)

	for _, u := range units {
		if err := emitParityUnitFunc(&b, u); err != nil {
			return nil, fmt.Errorf("emit parity unit %q: %w", u.Symbol, err)
		}
	}

	b.WriteString("var publicParityUnits = []struct {\n\tname  string\n\tshape string\n\trun   func(*testing.T) int\n}{\n")
	for _, u := range units {
		fmt.Fprintf(&b, "\t{%q, %q, %s},\n", u.Symbol, shapeName(u.Shape), u.FuncName)
	}
	b.WriteString("}\n")
	return []byte(b.String()), nil
}

func emitPublicParityStaticHelpers(b *strings.Builder, masks publicParityExceptionMasks) {
	b.WriteString(`// publicParityToBidgo128 / publicParityFromBidgo128 reinterpret the
// fixed-width 128-bit byte pattern as the port operand type, matching the
// value types' 1:1 byte correspondence. This is a pure bit reinterpretation,
// not a flag or rounding conversion.
func publicParityToBidgo128(raw [16]byte) bidgo.BID_UINT128 {
	return *(*bidgo.BID_UINT128)(unsafe.Pointer(&raw))
}

func publicParityFromBidgo128(value bidgo.BID_UINT128) [16]byte {
	return *(*[16]byte)(unsafe.Pointer(&value))
}

`)
	fmt.Fprintf(b, `// mapPortFlagsForParity maps the raw Intel status word to the public
// ExceptionFlags using numeric literals extracted from the pinned Intel
// bid_functions.h (BID_*_EXCEPTION). The parity runner must not reuse the
// wrappers' own bidgoExceptionFlags converter: a shared converter bug would be
// invisible. The hand-written public_flag_rounding_anchor_test.go pins these
// bit values and the wrapper converter numerically.
func mapPortFlagsForParity(flags uint32) ExceptionFlags {
	var out ExceptionFlags
	if flags&0x%02x != 0 {
		out |= FlagInvalidOperation
	}
	if flags&0x%02x != 0 {
		out |= FlagDivisionByZero
	}
	if flags&0x%02x != 0 {
		out |= FlagOverflow
	}
	if flags&0x%02x != 0 {
		out |= FlagUnderflow
	}
	if flags&0x%02x != 0 {
		out |= FlagInexact
	}
	return out
}

`, masks.Invalid, masks.DivByZero, masks.Overflow, masks.Underflow, masks.Inexact)

	b.WriteString(`// publicParityModes pairs each public RoundingMode with its port-domain
// integer as independent numeric literals; a wrapper whose bidgoRoundingMode
// mapping drifts diverges in value against these.
var publicParityModes = []struct {
	pub  RoundingMode
	port int
}{
	{RoundNearestEven, 0},
	{RoundNearestAway, 4},
	{RoundTowardZero, 3},
	{RoundTowardPositive, 2},
	{RoundTowardNegative, 1},
}

// publicParityClassName maps the Intel class_t integer to the public
// DecimalClass spelling independently of the decimalClassFromBIDClass wrapper.
func publicParityClassName(c int) string {
	switch c {
	case 0:
		return "sNaN"
	case 1:
		return "NaN"
	case 2:
		return "-Infinity"
	case 3:
		return "-Normal"
	case 4:
		return "-Subnormal"
	case 5:
		return "-Zero"
	case 6:
		return "+Zero"
	case 7:
		return "+Subnormal"
	case 8:
		return "+Normal"
	case 9:
		return "+Infinity"
	default:
		return "NaN"
	}
}

`)
}

func emitPublicParityCorpusTables(b *strings.Builder, corpus publicParityCorpus) {
	b.WriteString("var publicParityCorpus32 = []uint32{\n")
	for _, v := range corpus.Bits32 {
		fmt.Fprintf(b, "\t0x%08x,\n", v)
	}
	b.WriteString("}\n\n")

	b.WriteString("var publicParityCorpus64 = []uint64{\n")
	for _, v := range corpus.Bits64 {
		fmt.Fprintf(b, "\t0x%016x,\n", v)
	}
	b.WriteString("}\n\n")

	b.WriteString("var publicParityCorpus128 = [][16]byte{\n")
	for _, raw := range corpus.Bits128 {
		b.WriteString("\t{")
		for i, by := range raw {
			if i > 0 {
				b.WriteString(", ")
			}
			fmt.Fprintf(b, "0x%02x", by)
		}
		b.WriteString("},\n")
	}
	b.WriteString("}\n\n")

	// Per-width operand index matrices, resolved at generation time from the
	// label-level direction matrix so every width's binary/ternary wrappers
	// see qNaN, sNaN, and noncanonical operands in both directions.
	emitPairTable := func(name string, pairs [][2]int) {
		fmt.Fprintf(b, "var %s = [][2]int{\n", name)
		for _, p := range pairs {
			fmt.Fprintf(b, "\t{%d, %d},\n", p[0], p[1])
		}
		b.WriteString("}\n\n")
	}
	emitPairTable("publicParityBinaryPairs32", corpus.Pairs32)
	emitPairTable("publicParityBinaryPairs64", corpus.Pairs64)
	emitPairTable("publicParityBinaryPairs128", corpus.Pairs128)

	emitTripleTable := func(name string, triples [][3]int) {
		fmt.Fprintf(b, "var %s = [][3]int{\n", name)
		for _, t := range triples {
			fmt.Fprintf(b, "\t{%d, %d, %d},\n", t[0], t[1], t[2])
		}
		b.WriteString("}\n\n")
	}
	emitTripleTable("publicParityTernaryTriples32", corpus.Triples32)
	emitTripleTable("publicParityTernaryTriples64", corpus.Triples64)
	emitTripleTable("publicParityTernaryTriples128", corpus.Triples128)

	b.WriteString("var publicParityScaleBExps = []int{")
	for i, e := range publicParityScaleBExps {
		if i > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(b, "%d", e)
	}
	b.WriteString("}\n\n")

	b.WriteString("var publicParityIntCorpus32 = []int32{")
	for i, v := range publicParityIntCorpus32 {
		if i > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(b, "%d", v)
	}
	b.WriteString("}\n\n")

	b.WriteString("var publicParityUintCorpus32 = []uint32{")
	for i, v := range publicParityUintCorpus32 {
		if i > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(b, "%d", v)
	}
	b.WriteString("}\n\n")

	b.WriteString("var publicParityIntCorpus64 = []int64{")
	for i, v := range publicParityIntCorpus64 {
		if i > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(b, "%d", v)
	}
	b.WriteString("}\n\n")

	b.WriteString("var publicParityUintCorpus64 = []uint64{")
	for i, v := range publicParityUintCorpus64 {
		if i > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(b, "%d", v)
	}
	b.WriteString("}\n\n")

	b.WriteString(`// publicParityStringCases carries the generation-time classification of each
// string input against the documented public parsing contract. nanMinWidth
// and cohortMinWidth are the smallest BID widths that can preserve the
// corresponding property; zero means no supported width. Rejected payloads
// and otherwise-silent cohort coercions use the error channel where available
// and canonical quiet NaN + FlagInvalidOperation on raw flag-only parsers.
var publicParityStringCases = []struct {
	input           string
	kind            string
	signaling       bool
	nanMinWidth     int
	cohortMinWidth  int
}{
`)
	for _, sc := range publicParityStringCorpusCases {
		fmt.Fprintf(b, "\t{%q, %q, %v, %d, %d},\n", sc.Input, sc.Kind, sc.Signaling, sc.NaNMinWidth, sc.CohortMinWidth)
	}
	b.WriteString("}\n\n")
}

// ---- width helpers ----

func pwCorpus(w int) string  { return fmt.Sprintf("publicParityCorpus%d", w) }
func pwPairs(w int) string   { return fmt.Sprintf("publicParityBinaryPairs%d", w) }
func pwTriples(w int) string { return fmt.Sprintf("publicParityTernaryTriples%d", w) }
func pwDecType(w int) string { return fmt.Sprintf("Decimal%dBID", w) }

// pubToPortArg converts a public decimal value expression into the port
// operand type for predicate checks on public results.
func pubToPortArg(w int, expr string) string {
	switch w {
	case 32:
		return "uint32(" + expr + ")"
	case 64:
		return "uint64(" + expr + ")"
	default:
		return "publicParityToBidgo128(" + expr + ".ToBytes())"
	}
}

func pwPublicVal(w int, elem string) string {
	return fmt.Sprintf("%s(%s)", pwDecType(w), elem)
}

func pwPortArg(w int, elem string) string {
	if w == 128 {
		return fmt.Sprintf("publicParityToBidgo128(%s)", elem)
	}
	return elem
}

// pubBitsExpr / portBitsExpr normalize both sides to a comparable token.
func pubBitsExpr(class, pv string) string {
	switch class {
	case "dec32":
		return "uint32(" + pv + ")"
	case "dec64":
		return "uint64(" + pv + ")"
	case "dec128", "bin128":
		return pv + ".ToBytes()"
	case "bool", "int":
		return pv
	case "f32":
		return "math.Float32bits(" + pv + ")"
	case "f64":
		return "math.Float64bits(" + pv + ")"
	case "intn":
		return "int64(" + pv + ")"
	case "uintn":
		return "uint64(" + pv + ")"
	default:
		return pv
	}
}

func portBitsExpr(class, pr, primary string) string {
	switch class {
	case "dec32", "dec64":
		return pr
	case "dec128", "bin128":
		return "publicParityFromBidgo128(" + pr + ")"
	case "bool":
		if primary == "bool" {
			return pr
		}
		return "(" + pr + " != 0)"
	case "int":
		return pr
	case "f32":
		if primary == "uint32" {
			return pr
		}
		return "math.Float32bits(" + pr + ")"
	case "f64":
		if primary == "uint64" {
			return pr
		}
		return "math.Float64bits(" + pr + ")"
	case "intn":
		return "int64(" + pr + ")"
	case "uintn":
		return "uint64(" + pr + ")"
	default:
		return pr
	}
}

// emitGenericPort writes the port invocation, returning the result variable
// ("pr") and the flags variable ("pf" or "0"). compareFlags controls whether
// the port's flags are captured for comparison.
func emitGenericPort(b *strings.Builder, indent string, plan parityPortPlan, argExprs []string, modeExpr string, compareFlags bool) string {
	args := append([]string{}, argExprs...)
	if plan.HasRounding {
		args = append(args, modeExpr)
	}
	switch plan.FlagsKind {
	case "pointer":
		fmt.Fprintf(b, "%svar prf uint32\n", indent)
		args = append(args, "&prf")
		fmt.Fprintf(b, "%spr := bidgo.%s(%s)\n", indent, plan.GoName, strings.Join(args, ", "))
		return "prf"
	case "result":
		if compareFlags {
			fmt.Fprintf(b, "%spr, prf := bidgo.%s(%s)\n", indent, plan.GoName, strings.Join(args, ", "))
			return "prf"
		}
		fmt.Fprintf(b, "%spr, _ := bidgo.%s(%s)\n", indent, plan.GoName, strings.Join(args, ", "))
		return "0"
	default:
		fmt.Fprintf(b, "%spr := bidgo.%s(%s)\n", indent, plan.GoName, strings.Join(args, ", "))
		return "0"
	}
}

func emitResultCheck(b *strings.Builder, indent, symbol, class, pv, pr, primary, ctxFmt, ctxArgs string) {
	pub := pubBitsExpr(class, pv)
	port := portBitsExpr(class, pr, primary)
	prefix := ""
	if ctxArgs != "" {
		prefix = ctxArgs + ", "
	}
	fmt.Fprintf(b, "%sif %s != %s {\n", indent, pub, port)
	fmt.Fprintf(b, "%s\tt.Errorf(\"public parity %s: %s: result mismatch public=%%v port=%%v\", %s%s, %s)\n", indent, symbol, ctxFmt, prefix, pub, port)
	fmt.Fprintf(b, "%s}\n", indent)
}

func emitFlagCheck(b *strings.Builder, indent, symbol, pubFlags, portFlags, ctxFmt, ctxArgs string) {
	prefix := ""
	if ctxArgs != "" {
		prefix = ctxArgs + ", "
	}
	fmt.Fprintf(b, "%sif %s != mapPortFlagsForParity(%s) {\n", indent, pubFlags, portFlags)
	fmt.Fprintf(b, "%s\tt.Errorf(\"public parity %s: %s: flag mismatch public=%%v port=%%v\", %s%s, mapPortFlagsForParity(%s))\n", indent, symbol, ctxFmt, prefix, pubFlags, portFlags)
	fmt.Fprintf(b, "%s}\n", indent)
}

func emitParityUnitFunc(b *strings.Builder, u parityUnit) error {
	fmt.Fprintf(b, "func %s(t *testing.T) int {\n", u.FuncName)
	fmt.Fprintf(b, "\tcount := 0\n")
	if err := emitParityUnitBody(b, u); err != nil {
		return err
	}
	fmt.Fprintf(b, "\treturn count\n}\n\n")
	return nil
}

// publicCallMethodNoFlags / WithFlags render the public method invocation.
func emitParityUnitBody(b *strings.Builder, u parityUnit) error {
	switch u.Shape {
	case shapeVMUnary:
		return emitVMUnary(b, u)
	case shapeVMBinary:
		return emitVMBinary(b, u)
	case shapeVMTernary:
		return emitVMTernary(b, u)
	case shapeVMScaleB:
		return emitVMScaleB(b, u)
	case shapeVMNextToward:
		return emitVMNextToward(b, u)
	case shapeVMModeUnary:
		return emitVMModeUnary(b, u)
	case shapeVMModeUnaryArith:
		return emitVMModeUnaryArith(b, u)
	case shapeVMModeBinary:
		return emitVMModeBinary(b, u)
	case shapeVMModeTernary:
		return emitVMModeTernary(b, u)
	case shapeVMModeScaleB:
		return emitVMModeScaleB(b, u)
	case shapeVMConvert:
		return emitVMConvert(b, u)
	case shapeVMNullary:
		return emitVMNullary(b, u)
	case shapeVMCompareTotal:
		return emitVMCompareTotal(b, u)
	case shapeVMSign:
		return emitVMSign(b, u)
	case shapeVMSignalingEqual, shapeVMSignalingNotEqual:
		return emitVMSignalingEqual(b, u)
	case shapeVMClass:
		return emitVMClass(b, u)
	case shapeVMString:
		return emitVMString(b, u)
	case shapeFuncIntCtor:
		return emitFuncIntCtor(b, u)
	case shapeFuncFromInt:
		return emitFuncFromInt(b, u)
	case shapeFuncString:
		return emitFuncString(b, u)
	case shapeFuncStringMode:
		return emitFuncStringMode(b, u)
	case shapeFuncContext:
		return emitFuncContext(b, u)
	case shapeFuncMixedModeBinary:
		return emitFuncMixedModeBinary(b, u)
	case shapeFuncMixedModeTernary:
		return emitFuncMixedModeTernary(b, u)
	case shapeFuncMixedModeUnary:
		return emitFuncMixedModeUnary(b, u)
	default:
		return fmt.Errorf("no emitter for shape %d", u.Shape)
	}
}

// emitFuncMixedModeBinary renders one Intel D/Q mixed-width public free
// function. The routing leg resolves the same label-level pair matrix at each
// operand's own width, so NaN/infinity/noncanonical directionality is retained
// across unlike representations. The discriminant leg encodes each decimal
// component at its input width and compares every rounding mode directly with
// the mapped mechanical-port function. A final case pins the public invalid-
// mode rejection contract without passing an out-of-domain mode to Intel.
func emitFuncMixedModeBinary(b *strings.Builder, u parityUnit) error {
	leftWidth, rightWidth := u.OperandWidths[0], u.OperandWidths[1]
	disc, err := mixedModeBinaryDiscriminantOperands(u.Operation, u.Width, u.OperandWidths)
	if err != nil {
		return err
	}

	// Shared routing-corpus leg. The label matrix has the same length and row
	// meaning at every width; take the left slot from the left-width row and
	// the right slot from the right-width row.
	fmt.Fprintf(b, "\tfor pairIndex := range %s {\n", pwPairs(leftWidth))
	fmt.Fprintf(b, "\t\tleftPair := %s[pairIndex]\n", pwPairs(leftWidth))
	fmt.Fprintf(b, "\t\trightPair := %s[pairIndex]\n", pwPairs(rightWidth))
	fmt.Fprintf(b, "\t\tleftBits := %s[leftPair[0]]\n", pwCorpus(leftWidth))
	fmt.Fprintf(b, "\t\trightBits := %s[rightPair[1]]\n", pwCorpus(rightWidth))
	fmt.Fprintf(b, "\t\tleft := %s\n", pwPublicVal(leftWidth, "leftBits"))
	fmt.Fprintf(b, "\t\tright := %s\n", pwPublicVal(rightWidth, "rightBits"))
	fmt.Fprintf(b, "\t\tfor _, mode := range publicParityModes {\n")
	fmt.Fprintf(b, "\t\t\tpv, pf := %s(left, right, mode.pub)\n", u.Func)
	pfPort := emitGenericPort(b, "\t\t\t", u.Port, []string{pwPortArg(leftWidth, "leftBits"), pwPortArg(rightWidth, "rightBits")}, "mode.port", true)
	emitResultCheck(b, "\t\t\t", u.Symbol, u.ResultClass, "pv", "pr", u.Port.PrimaryResult, "operands %v,%v mode %v", "leftBits, rightBits, mode.pub")
	emitFlagCheck(b, "\t\t\t", u.Symbol, "pf", pfPort, "operands %v,%v mode %v", "leftBits, rightBits, mode.pub")
	fmt.Fprintf(b, "\t\t\tcount++\n\t\t}\n\t}\n")

	if len(disc) > 0 {
		leftType, rightType := modeDiscGoType(leftWidth), modeDiscGoType(rightWidth)
		fmt.Fprintf(b, "\tdiscPairs := []struct {\n\t\tleft %s\n\t\tright %s\n\t}{\n", leftType, rightType)
		for _, pair := range disc {
			leftLit, err := modeDiscGoLiteral(leftWidth, pair[0])
			if err != nil {
				return fmt.Errorf("%s: left discriminant: %w", u.Symbol, err)
			}
			rightLit, err := modeDiscGoLiteral(rightWidth, pair[1])
			if err != nil {
				return fmt.Errorf("%s: right discriminant: %w", u.Symbol, err)
			}
			if leftWidth == 128 {
				leftLit = "[16]byte" + leftLit
			}
			if rightWidth == 128 {
				rightLit = "[16]byte" + rightLit
			}
			fmt.Fprintf(b, "\t\t{%s, %s},\n", leftLit, rightLit)
		}
		fmt.Fprintf(b, "\t}\n")
		fmt.Fprintf(b, "\tfor _, pair := range discPairs {\n")
		fmt.Fprintf(b, "\t\tleft := %s\n", pwPublicVal(leftWidth, "pair.left"))
		fmt.Fprintf(b, "\t\tright := %s\n", pwPublicVal(rightWidth, "pair.right"))
		fmt.Fprintf(b, "\t\tvar modeSeen [%d]%s\n", len(parityModeOrder), modeDiscGoType(u.Width))
		fmt.Fprintf(b, "\t\tfor mi, mode := range publicParityModes {\n")
		fmt.Fprintf(b, "\t\t\tpv, pf := %s(left, right, mode.pub)\n", u.Func)
		pfDisc := emitGenericPort(b, "\t\t\t", u.Port, []string{pwPortArg(leftWidth, "pair.left"), pwPortArg(rightWidth, "pair.right")}, "mode.port", true)
		emitResultCheck(b, "\t\t\t", u.Symbol, u.ResultClass, "pv", "pr", u.Port.PrimaryResult, "discriminant operands %v,%v mode %v", "pair.left, pair.right, mode.pub")
		emitFlagCheck(b, "\t\t\t", u.Symbol, "pf", pfDisc, "discriminant operands %v,%v mode %v", "pair.left, pair.right, mode.pub")
		fmt.Fprintf(b, "\t\t\tmodeSeen[mi] = %s\n", pubBitsExpr(u.ResultClass, "pv"))
		fmt.Fprintf(b, "\t\t\tcount++\n\t\t}\n")
		emitModeDiscAssertion(b, u.Symbol, "discriminant operands %v,%v", "pair.left, pair.right")
		fmt.Fprintf(b, "\t}\n")
	}

	// Invalid public mode: do not call Intel with an unsupported integer.
	fmt.Fprintf(b, "\tinvalidLeft := %s\n", pwPublicVal(leftWidth, pwCorpus(leftWidth)+"[0]"))
	fmt.Fprintf(b, "\tinvalidRight := %s\n", pwPublicVal(rightWidth, pwCorpus(rightWidth)+"[0]"))
	fmt.Fprintf(b, "\tinvalidValue, invalidFlags := %s(invalidLeft, invalidRight, RoundingMode(99))\n", u.Func)
	canonical := fmt.Sprintf("canonicalQNaN%dBID()", u.Width)
	fmt.Fprintf(b, "\tif %s != %s || invalidFlags != FlagInvalidOperation {\n", pubBitsExpr(u.ResultClass, "invalidValue"), pubBitsExpr(u.ResultClass, canonical))
	fmt.Fprintf(b, "\t\tt.Errorf(\"public parity %s: invalid rounding mode result=%%v flags=%%v, want canonical qNaN and FlagInvalidOperation\", %s, invalidFlags)\n", u.Symbol, pubBitsExpr(u.ResultClass, "invalidValue"))
	fmt.Fprintf(b, "\t}\n\tcount++\n")
	return nil
}

// emitFuncMixedModeTernary renders the Intel D/Q mixed-width FMA free
// functions. Each operand position resolves the shared label triple at its
// own width, then the result is compared with the exact mapped Go-port call.
func emitFuncMixedModeTernary(b *strings.Builder, u parityUnit) error {
	widths := u.TernaryOperandWidths
	disc, err := mixedModeTernaryDiscriminantOperands(u.Operation, u.Width, widths)
	if err != nil {
		return err
	}

	fmt.Fprintf(b, "\tfor tripleIndex := range %s {\n", pwTriples(widths[0]))
	fmt.Fprintf(b, "\t\taTriple := %s[tripleIndex]\n", pwTriples(widths[0]))
	fmt.Fprintf(b, "\t\tbTriple := %s[tripleIndex]\n", pwTriples(widths[1]))
	fmt.Fprintf(b, "\t\tcTriple := %s[tripleIndex]\n", pwTriples(widths[2]))
	fmt.Fprintf(b, "\t\taBits := %s[aTriple[0]]\n", pwCorpus(widths[0]))
	fmt.Fprintf(b, "\t\tbBits := %s[bTriple[1]]\n", pwCorpus(widths[1]))
	fmt.Fprintf(b, "\t\tcBits := %s[cTriple[2]]\n", pwCorpus(widths[2]))
	fmt.Fprintf(b, "\t\ta := %s\n", pwPublicVal(widths[0], "aBits"))
	fmt.Fprintf(b, "\t\tbv := %s\n", pwPublicVal(widths[1], "bBits"))
	fmt.Fprintf(b, "\t\tc := %s\n", pwPublicVal(widths[2], "cBits"))
	fmt.Fprintf(b, "\t\tfor _, mode := range publicParityModes {\n")
	fmt.Fprintf(b, "\t\t\tpv, pf := %s(a, bv, c, mode.pub)\n", u.Func)
	pfPort := emitGenericPort(b, "\t\t\t", u.Port, []string{pwPortArg(widths[0], "aBits"), pwPortArg(widths[1], "bBits"), pwPortArg(widths[2], "cBits")}, "mode.port", true)
	emitResultCheck(b, "\t\t\t", u.Symbol, u.ResultClass, "pv", "pr", u.Port.PrimaryResult, "operands %v,%v,%v mode %v", "aBits, bBits, cBits, mode.pub")
	emitFlagCheck(b, "\t\t\t", u.Symbol, "pf", pfPort, "operands %v,%v,%v mode %v", "aBits, bBits, cBits, mode.pub")
	fmt.Fprintf(b, "\t\t\tcount++\n\t\t}\n\t}\n")

	fmt.Fprintf(b, "\tdiscTriples := []struct {\n\t\ta %s\n\t\tb %s\n\t\tc %s\n\t}{\n", modeDiscGoType(widths[0]), modeDiscGoType(widths[1]), modeDiscGoType(widths[2]))
	for _, triple := range disc {
		literals := [3]string{}
		for i, operand := range triple {
			literal, err := modeDiscGoLiteral(widths[i], operand)
			if err != nil {
				return fmt.Errorf("%s: discriminant operand %d: %w", u.Symbol, i, err)
			}
			if widths[i] == 128 {
				literal = "[16]byte" + literal
			}
			literals[i] = literal
		}
		fmt.Fprintf(b, "\t\t{%s, %s, %s},\n", literals[0], literals[1], literals[2])
	}
	fmt.Fprintf(b, "\t}\n")
	fmt.Fprintf(b, "\tfor _, triple := range discTriples {\n")
	fmt.Fprintf(b, "\t\ta := %s\n", pwPublicVal(widths[0], "triple.a"))
	fmt.Fprintf(b, "\t\tbv := %s\n", pwPublicVal(widths[1], "triple.b"))
	fmt.Fprintf(b, "\t\tc := %s\n", pwPublicVal(widths[2], "triple.c"))
	fmt.Fprintf(b, "\t\tvar modeSeen [%d]%s\n", len(parityModeOrder), modeDiscGoType(u.Width))
	fmt.Fprintf(b, "\t\tfor mi, mode := range publicParityModes {\n")
	fmt.Fprintf(b, "\t\t\tpv, pf := %s(a, bv, c, mode.pub)\n", u.Func)
	pfDisc := emitGenericPort(b, "\t\t\t", u.Port, []string{pwPortArg(widths[0], "triple.a"), pwPortArg(widths[1], "triple.b"), pwPortArg(widths[2], "triple.c")}, "mode.port", true)
	emitResultCheck(b, "\t\t\t", u.Symbol, u.ResultClass, "pv", "pr", u.Port.PrimaryResult, "discriminant operands %v,%v,%v mode %v", "triple.a, triple.b, triple.c, mode.pub")
	emitFlagCheck(b, "\t\t\t", u.Symbol, "pf", pfDisc, "discriminant operands %v,%v,%v mode %v", "triple.a, triple.b, triple.c, mode.pub")
	fmt.Fprintf(b, "\t\t\tmodeSeen[mi] = %s\n", pubBitsExpr(u.ResultClass, "pv"))
	fmt.Fprintf(b, "\t\t\tcount++\n\t\t}\n")
	emitModeDiscAssertion(b, u.Symbol, "discriminant operands %v,%v,%v", "triple.a, triple.b, triple.c")
	fmt.Fprintf(b, "\t}\n")

	fmt.Fprintf(b, "\tinvalidA := %s\n", pwPublicVal(widths[0], pwCorpus(widths[0])+"[0]"))
	fmt.Fprintf(b, "\tinvalidB := %s\n", pwPublicVal(widths[1], pwCorpus(widths[1])+"[0]"))
	fmt.Fprintf(b, "\tinvalidC := %s\n", pwPublicVal(widths[2], pwCorpus(widths[2])+"[0]"))
	fmt.Fprintf(b, "\tinvalidValue, invalidFlags := %s(invalidA, invalidB, invalidC, RoundingMode(99))\n", u.Func)
	canonical := fmt.Sprintf("canonicalQNaN%dBID()", u.Width)
	fmt.Fprintf(b, "\tif %s != %s || invalidFlags != FlagInvalidOperation {\n", pubBitsExpr(u.ResultClass, "invalidValue"), pubBitsExpr(u.ResultClass, canonical))
	fmt.Fprintf(b, "\t\tt.Errorf(\"public parity %s: invalid rounding mode result=%%v flags=%%v, want canonical qNaN and FlagInvalidOperation\", %s, invalidFlags)\n", u.Symbol, pubBitsExpr(u.ResultClass, "invalidValue"))
	fmt.Fprintf(b, "\t}\n\tcount++\n")
	return nil
}

// emitFuncMixedModeUnary renders the two unlike-width sqrt free functions.
// The routing and discriminant operands are encoded at the input width while
// the mode-sensitivity collector uses the result width.
func emitFuncMixedModeUnary(b *strings.Builder, u parityUnit) error {
	operandWidth := u.UnaryOperandWidth
	disc, err := mixedModeUnaryDiscriminantOperands(u.Operation, u.Width, operandWidth)
	if err != nil {
		return err
	}

	fmt.Fprintf(b, "\tfor _, elem := range %s {\n", pwCorpus(operandWidth))
	fmt.Fprintf(b, "\t\toperand := %s\n", pwPublicVal(operandWidth, "elem"))
	fmt.Fprintf(b, "\t\tfor _, mode := range publicParityModes {\n")
	fmt.Fprintf(b, "\t\t\tpv, pf := %s(operand, mode.pub)\n", u.Func)
	pfPort := emitGenericPort(b, "\t\t\t", u.Port, []string{pwPortArg(operandWidth, "elem")}, "mode.port", true)
	emitResultCheck(b, "\t\t\t", u.Symbol, u.ResultClass, "pv", "pr", u.Port.PrimaryResult, "operand %v mode %v", "elem, mode.pub")
	emitFlagCheck(b, "\t\t\t", u.Symbol, "pf", pfPort, "operand %v mode %v", "elem, mode.pub")
	fmt.Fprintf(b, "\t\t\tcount++\n\t\t}\n\t}\n")

	fmt.Fprintf(b, "\tdiscVals := []%s{\n", modeDiscGoType(operandWidth))
	for _, operand := range disc {
		literal, err := modeDiscGoLiteral(operandWidth, operand)
		if err != nil {
			return fmt.Errorf("%s: discriminant operand: %w", u.Symbol, err)
		}
		fmt.Fprintf(b, "\t\t%s,\n", literal)
	}
	fmt.Fprintf(b, "\t}\n")
	fmt.Fprintf(b, "\tfor _, dv := range discVals {\n")
	fmt.Fprintf(b, "\t\toperand := %s\n", pwPublicVal(operandWidth, "dv"))
	fmt.Fprintf(b, "\t\tvar modeSeen [%d]%s\n", len(parityModeOrder), modeDiscGoType(u.Width))
	fmt.Fprintf(b, "\t\tfor mi, mode := range publicParityModes {\n")
	fmt.Fprintf(b, "\t\t\tpv, pf := %s(operand, mode.pub)\n", u.Func)
	pfDisc := emitGenericPort(b, "\t\t\t", u.Port, []string{pwPortArg(operandWidth, "dv")}, "mode.port", true)
	emitResultCheck(b, "\t\t\t", u.Symbol, u.ResultClass, "pv", "pr", u.Port.PrimaryResult, "discriminant operand %v mode %v", "dv, mode.pub")
	emitFlagCheck(b, "\t\t\t", u.Symbol, "pf", pfDisc, "discriminant operand %v mode %v", "dv, mode.pub")
	fmt.Fprintf(b, "\t\t\tmodeSeen[mi] = %s\n", pubBitsExpr(u.ResultClass, "pv"))
	fmt.Fprintf(b, "\t\t\tcount++\n\t\t}\n")
	emitModeDiscAssertion(b, u.Symbol, "discriminant operand %v", "dv")
	fmt.Fprintf(b, "\t}\n")

	fmt.Fprintf(b, "\tinvalidOperand := %s\n", pwPublicVal(operandWidth, pwCorpus(operandWidth)+"[0]"))
	fmt.Fprintf(b, "\tinvalidValue, invalidFlags := %s(invalidOperand, RoundingMode(99))\n", u.Func)
	canonical := fmt.Sprintf("canonicalQNaN%dBID()", u.Width)
	fmt.Fprintf(b, "\tif %s != %s || invalidFlags != FlagInvalidOperation {\n", pubBitsExpr(u.ResultClass, "invalidValue"), pubBitsExpr(u.ResultClass, canonical))
	fmt.Fprintf(b, "\t\tt.Errorf(\"public parity %s: invalid rounding mode result=%%v flags=%%v, want canonical qNaN and FlagInvalidOperation\", %s, invalidFlags)\n", u.Symbol, pubBitsExpr(u.ResultClass, "invalidValue"))
	fmt.Fprintf(b, "\t}\n\tcount++\n")
	return nil
}

func emitVMUnary(b *strings.Builder, u parityUnit) error {
	w := u.Width
	fmt.Fprintf(b, "\tfor _, elem := range %s {\n", pwCorpus(w))
	fmt.Fprintf(b, "\t\ta := %s\n", pwPublicVal(w, "elem"))
	if u.HasFlags {
		fmt.Fprintf(b, "\t\tpv, pf := a.%s()\n", u.Method)
	} else {
		fmt.Fprintf(b, "\t\tpv := a.%s()\n", u.Method)
	}
	pfPort := emitGenericPort(b, "\t\t", u.Port, []string{pwPortArg(w, "elem")}, "0", u.HasFlags)
	emitResultCheck(b, "\t\t", u.Symbol, u.ResultClass, "pv", "pr", u.Port.PrimaryResult, "operand %#x", "elem")
	if u.HasFlags {
		emitFlagCheck(b, "\t\t", u.Symbol, "pf", pfPort, "operand %#x", "elem")
	}
	fmt.Fprintf(b, "\t\tcount++\n\t}\n")
	return nil
}

func emitVMBinary(b *strings.Builder, u parityUnit) error {
	w := u.Width
	fmt.Fprintf(b, "\tfor _, pair := range %s {\n", pwPairs(w))
	fmt.Fprintf(b, "\t\tav := %s[pair[0]]\n", pwCorpus(w))
	fmt.Fprintf(b, "\t\tbv := %s[pair[1]]\n", pwCorpus(w))
	fmt.Fprintf(b, "\t\ta := %s\n", pwPublicVal(w, "av"))
	fmt.Fprintf(b, "\t\tb := %s\n", pwPublicVal(w, "bv"))
	if u.HasFlags {
		fmt.Fprintf(b, "\t\tpv, pf := a.%s(b)\n", u.Method)
	} else {
		fmt.Fprintf(b, "\t\tpv := a.%s(b)\n", u.Method)
	}
	pfPort := emitGenericPort(b, "\t\t", u.Port, []string{pwPortArg(w, "av"), pwPortArg(w, "bv")}, "0", u.HasFlags)
	emitResultCheck(b, "\t\t", u.Symbol, u.ResultClass, "pv", "pr", u.Port.PrimaryResult, "operands %#x,%#x", "av, bv")
	if u.HasFlags {
		emitFlagCheck(b, "\t\t", u.Symbol, "pf", pfPort, "operands %#x,%#x", "av, bv")
	}
	fmt.Fprintf(b, "\t\tcount++\n\t}\n")
	return nil
}

func emitVMTernary(b *strings.Builder, u parityUnit) error {
	w := u.Width
	fmt.Fprintf(b, "\tfor _, tri := range %s {\n", pwTriples(w))
	fmt.Fprintf(b, "\t\tav := %s[tri[0]]\n", pwCorpus(w))
	fmt.Fprintf(b, "\t\tbv := %s[tri[1]]\n", pwCorpus(w))
	fmt.Fprintf(b, "\t\tcv := %s[tri[2]]\n", pwCorpus(w))
	fmt.Fprintf(b, "\t\ta := %s\n", pwPublicVal(w, "av"))
	fmt.Fprintf(b, "\t\tb := %s\n", pwPublicVal(w, "bv"))
	fmt.Fprintf(b, "\t\tc := %s\n", pwPublicVal(w, "cv"))
	fmt.Fprintf(b, "\t\tpv, pf := a.%s(b, c)\n", u.Method)
	pfPort := emitGenericPort(b, "\t\t", u.Port, []string{pwPortArg(w, "av"), pwPortArg(w, "bv"), pwPortArg(w, "cv")}, "0", true)
	emitResultCheck(b, "\t\t", u.Symbol, u.ResultClass, "pv", "pr", u.Port.PrimaryResult, "operands %#x,%#x,%#x", "av, bv, cv")
	emitFlagCheck(b, "\t\t", u.Symbol, "pf", pfPort, "operands %#x,%#x,%#x", "av, bv, cv")
	fmt.Fprintf(b, "\t\tcount++\n\t}\n")
	return nil
}

// scaleBExpArg wraps the int corpus exponent to match the port entrypoint's
// exponent parameter type. The Intel bid<w>_scalbln long-int entrypoints take
// an int64 (they clamp the wider exponent into the int32 domain in the port),
// so the int exponent is widened; a historical bid<w>_scalbn int entrypoint
// would take it directly.
func scaleBExpArg(u parityUnit, expExpr string) string {
	if len(u.Port.ValueParams) >= 2 && u.Port.ValueParams[1] == "int64" {
		return "int64(" + expExpr + ")"
	}
	return expExpr
}

func emitVMScaleB(b *strings.Builder, u parityUnit) error {
	w := u.Width
	fmt.Fprintf(b, "\tfor _, elem := range %s {\n", pwCorpus(w))
	fmt.Fprintf(b, "\t\ta := %s\n", pwPublicVal(w, "elem"))
	fmt.Fprintf(b, "\t\tfor _, exp := range publicParityScaleBExps {\n")
	fmt.Fprintf(b, "\t\t\tpv, pf := a.ScaleB(exp)\n")
	pfPort := emitGenericPort(b, "\t\t\t", u.Port, []string{pwPortArg(w, "elem"), scaleBExpArg(u, "exp")}, "0", true)
	emitResultCheck(b, "\t\t\t", u.Symbol, u.ResultClass, "pv", "pr", u.Port.PrimaryResult, "operand %#x exp %d", "elem, exp")
	emitFlagCheck(b, "\t\t\t", u.Symbol, "pf", pfPort, "operand %#x exp %d", "elem, exp")
	fmt.Fprintf(b, "\t\t\tcount++\n\t\t}\n\t}\n")
	return nil
}

func emitVMNextToward(b *strings.Builder, u parityUnit) error {
	w := u.Width
	fmt.Fprintf(b, "\tfor _, elem := range %s {\n", pwCorpus(w))
	fmt.Fprintf(b, "\t\ta := %s\n", pwPublicVal(w, "elem"))
	fmt.Fprintf(b, "\t\tfor _, tb := range publicParityCorpus128[:%d] {\n", publicParityNextTowardTargets)
	fmt.Fprintf(b, "\t\t\ttarget := Decimal128BID(tb)\n")
	fmt.Fprintf(b, "\t\t\tpv, pf := a.NextToward(target)\n")
	pfPort := emitGenericPort(b, "\t\t\t", u.Port, []string{pwPortArg(w, "elem"), "publicParityToBidgo128(tb)"}, "0", true)
	emitResultCheck(b, "\t\t\t", u.Symbol, u.ResultClass, "pv", "pr", u.Port.PrimaryResult, "operand %#x target %x", "elem, tb")
	emitFlagCheck(b, "\t\t\t", u.Symbol, "pf", pfPort, "operand %#x target %x", "elem, tb")
	fmt.Fprintf(b, "\t\t\tcount++\n\t\t}\n\t}\n")
	return nil
}

func emitVMModeUnary(b *strings.Builder, u parityUnit) error {
	w := u.Width
	fmt.Fprintf(b, "\tfor _, elem := range %s {\n", pwCorpus(w))
	fmt.Fprintf(b, "\t\ta := %s\n", pwPublicVal(w, "elem"))
	fmt.Fprintf(b, "\t\tfor _, mode := range publicParityModes {\n")
	if u.HasFlags {
		fmt.Fprintf(b, "\t\t\tpv, pf := a.%s(mode.pub)\n", u.Method)
	} else {
		fmt.Fprintf(b, "\t\t\tpv := a.%s(mode.pub)\n", u.Method)
	}
	pfPort := emitGenericPort(b, "\t\t\t", u.Port, []string{pwPortArg(w, "elem")}, "mode.port", u.HasFlags)
	emitResultCheck(b, "\t\t\t", u.Symbol, u.ResultClass, "pv", "pr", u.Port.PrimaryResult, "operand %#x mode %v", "elem, mode.pub")
	if u.HasFlags {
		emitFlagCheck(b, "\t\t\t", u.Symbol, "pf", pfPort, "operand %#x mode %v", "elem, mode.pub")
	}
	fmt.Fprintf(b, "\t\t\tcount++\n\t\t}\n\t}\n")
	return nil
}

// emitVMModeBinary renders the {Add,Sub,Mul,Div}WithMode parity body: one
// same-width operand plus an explicit RoundingMode, always returning flags.
// Two operand legs run under every mode in publicParityModes against the
// independently mapped port call (mode.port):
//
//   - the shared parityLabelPairs corpus: NaN/infinity/sign/noncanonical
//     ROUTING probes. Most of these combinations are exact under every
//     rounding mode, so this leg alone cannot prove the wrapper forwards its
//     mode argument (measured: its Mul/Div combinations at widths 32/64 never
//     round);
//   - the per-operation mode-discriminant table (modeBinaryDiscriminants):
//     operand pairs constructed so the exact result does not fit the width's
//     precision, where the five rounding directions cannot all agree. The
//     generated body asserts per pair that the five wrapper results are not
//     all identical, so a wrapper that drops its mode argument fails both the
//     per-mode bit compare and this discrimination assertion, and a future
//     corpus edit that stops discriminating fails the gate instead of
//     silently making it vacuous.
func emitVMModeBinary(b *strings.Builder, u parityUnit) error {
	w := u.Width
	op := strings.TrimSuffix(u.Method, "WithMode")
	disc, err := modeBinaryDiscriminantOperands(op, w)
	if err != nil {
		return err
	}

	// Shared routing-corpus leg.
	fmt.Fprintf(b, "\tfor _, pair := range %s {\n", pwPairs(w))
	fmt.Fprintf(b, "\t\tav := %s[pair[0]]\n", pwCorpus(w))
	fmt.Fprintf(b, "\t\tbv := %s[pair[1]]\n", pwCorpus(w))
	fmt.Fprintf(b, "\t\ta := %s\n", pwPublicVal(w, "av"))
	fmt.Fprintf(b, "\t\tb := %s\n", pwPublicVal(w, "bv"))
	fmt.Fprintf(b, "\t\tfor _, mode := range publicParityModes {\n")
	fmt.Fprintf(b, "\t\t\tpv, pf := a.%s(b, mode.pub)\n", u.Method)
	pfPort := emitGenericPort(b, "\t\t\t", u.Port, []string{pwPortArg(w, "av"), pwPortArg(w, "bv")}, "mode.port", true)
	emitResultCheck(b, "\t\t\t", u.Symbol, u.ResultClass, "pv", "pr", u.Port.PrimaryResult, "operands %#x,%#x mode %v", "av, bv, mode.pub")
	emitFlagCheck(b, "\t\t\t", u.Symbol, "pf", pfPort, "operands %#x,%#x mode %v", "av, bv, mode.pub")
	fmt.Fprintf(b, "\t\t\tcount++\n\t\t}\n\t}\n")

	// Mode-discriminant leg: inexact operand pairs plus the per-pair
	// discrimination assertion.
	var discType string
	switch w {
	case 32:
		discType = "uint32"
	case 64:
		discType = "uint64"
	default:
		discType = "[16]byte"
	}
	fmt.Fprintf(b, "\tdiscPairs := [][2]%s{\n", discType)
	for _, dp := range disc {
		aLit, err := modeDiscGoLiteral(w, dp[0])
		if err != nil {
			return fmt.Errorf("%s: %w", u.Symbol, err)
		}
		bLit, err := modeDiscGoLiteral(w, dp[1])
		if err != nil {
			return fmt.Errorf("%s: %w", u.Symbol, err)
		}
		fmt.Fprintf(b, "\t\t{%s, %s},\n", aLit, bLit)
	}
	fmt.Fprintf(b, "\t}\n")
	fmt.Fprintf(b, "\tfor _, dp := range discPairs {\n")
	fmt.Fprintf(b, "\t\ta := %s\n", pwPublicVal(w, "dp[0]"))
	fmt.Fprintf(b, "\t\tb := %s\n", pwPublicVal(w, "dp[1]"))
	fmt.Fprintf(b, "\t\tvar modeSeen [%d]%s\n", len(parityModeOrder), discType)
	fmt.Fprintf(b, "\t\tfor mi, mode := range publicParityModes {\n")
	fmt.Fprintf(b, "\t\t\tpv, pf := a.%s(b, mode.pub)\n", u.Method)
	pfDisc := emitGenericPort(b, "\t\t\t", u.Port, []string{pwPortArg(w, "dp[0]"), pwPortArg(w, "dp[1]")}, "mode.port", true)
	emitResultCheck(b, "\t\t\t", u.Symbol, u.ResultClass, "pv", "pr", u.Port.PrimaryResult, "discriminant operands %#x,%#x mode %v", "dp[0], dp[1], mode.pub")
	emitFlagCheck(b, "\t\t\t", u.Symbol, "pf", pfDisc, "discriminant operands %#x,%#x mode %v", "dp[0], dp[1], mode.pub")
	fmt.Fprintf(b, "\t\t\tmodeSeen[mi] = %s\n", pubBitsExpr(u.ResultClass, "pv"))
	fmt.Fprintf(b, "\t\t\tcount++\n\t\t}\n")
	fmt.Fprintf(b, "\t\tmodeInsensitive := true\n")
	fmt.Fprintf(b, "\t\tfor mi := 1; mi < len(modeSeen); mi++ {\n")
	fmt.Fprintf(b, "\t\t\tif modeSeen[mi] != modeSeen[0] {\n\t\t\t\tmodeInsensitive = false\n\t\t\t\tbreak\n\t\t\t}\n\t\t}\n")
	fmt.Fprintf(b, "\t\tif modeInsensitive {\n")
	fmt.Fprintf(b, "\t\t\tt.Errorf(\"public parity %s: discriminant operands %%#x,%%#x: every rounding mode produced the same result; the mode-discriminant corpus entry no longer discriminates\", dp[0], dp[1])\n", u.Symbol)
	fmt.Fprintf(b, "\t\t}\n\t}\n")
	return nil
}

// modeDiscGoType / modeDiscSeenDecl give the Go corpus element type name for
// a width and the per-mode result collector declaration the discriminant
// assertions share.
func modeDiscGoType(width int) string {
	switch width {
	case 32:
		return "uint32"
	case 64:
		return "uint64"
	default:
		return "[16]byte"
	}
}

// emitModeDiscAssertion writes the shared not-all-identical check over the
// collected per-mode results, failing when a discriminant case stops
// discriminating (or a wrapper stops forwarding its mode).
func emitModeDiscAssertion(b *strings.Builder, symbol, ctxFmt, ctxArgs string) {
	fmt.Fprintf(b, "\t\tmodeInsensitive := true\n")
	fmt.Fprintf(b, "\t\tfor mi := 1; mi < len(modeSeen); mi++ {\n")
	fmt.Fprintf(b, "\t\t\tif modeSeen[mi] != modeSeen[0] {\n\t\t\t\tmodeInsensitive = false\n\t\t\t\tbreak\n\t\t\t}\n\t\t}\n")
	fmt.Fprintf(b, "\t\tif modeInsensitive {\n")
	fmt.Fprintf(b, "\t\t\tt.Errorf(\"public parity %s: %s: every rounding mode produced the same result; the mode-discriminant corpus entry no longer discriminates\", %s)\n", symbol, ctxFmt, ctxArgs)
	fmt.Fprintf(b, "\t\t}\n")
}

// emitVMModeUnaryArith renders unary explicit-mode parity bodies: the receiver
// is the only operand, always returning flags. Two legs run under every mode:
// the shared per-width corpus (routing probes), then the per-operation unary
// mode-discriminant table (irrational-result operands) with the per-case
// discrimination assertion, so a wrapper that drops its mode argument fails
// both the per-mode bit compare and the assertion, and a corpus edit that
// stops discriminating fails the gate.
func emitVMModeUnaryArith(b *strings.Builder, u parityUnit) error {
	w := u.Width
	op := strings.TrimSuffix(u.Method, "WithMode")
	disc, err := modeUnaryDiscriminantOperands(op, w)
	if err != nil {
		return err
	}

	// Shared routing-corpus leg.
	fmt.Fprintf(b, "\tfor _, elem := range %s {\n", pwCorpus(w))
	fmt.Fprintf(b, "\t\ta := %s\n", pwPublicVal(w, "elem"))
	fmt.Fprintf(b, "\t\tfor _, mode := range publicParityModes {\n")
	fmt.Fprintf(b, "\t\t\tpv, pf := a.%s(mode.pub)\n", u.Method)
	pfPort := emitGenericPort(b, "\t\t\t", u.Port, []string{pwPortArg(w, "elem")}, "mode.port", true)
	emitResultCheck(b, "\t\t\t", u.Symbol, u.ResultClass, "pv", "pr", u.Port.PrimaryResult, "operand %#x mode %v", "elem, mode.pub")
	emitFlagCheck(b, "\t\t\t", u.Symbol, "pf", pfPort, "operand %#x mode %v", "elem, mode.pub")
	fmt.Fprintf(b, "\t\t\tcount++\n\t\t}\n\t}\n")

	// Mode-discriminant leg.
	discType := modeDiscGoType(w)
	fmt.Fprintf(b, "\tdiscVals := []%s{\n", discType)
	for _, dv := range disc {
		lit, err := modeDiscGoLiteral(w, dv)
		if err != nil {
			return fmt.Errorf("%s: %w", u.Symbol, err)
		}
		fmt.Fprintf(b, "\t\t%s,\n", lit)
	}
	fmt.Fprintf(b, "\t}\n")
	fmt.Fprintf(b, "\tfor _, dv := range discVals {\n")
	fmt.Fprintf(b, "\t\ta := %s\n", pwPublicVal(w, "dv"))
	fmt.Fprintf(b, "\t\tvar modeSeen [%d]%s\n", len(parityModeOrder), discType)
	fmt.Fprintf(b, "\t\tfor mi, mode := range publicParityModes {\n")
	fmt.Fprintf(b, "\t\t\tpv, pf := a.%s(mode.pub)\n", u.Method)
	pfDisc := emitGenericPort(b, "\t\t\t", u.Port, []string{pwPortArg(w, "dv")}, "mode.port", true)
	emitResultCheck(b, "\t\t\t", u.Symbol, u.ResultClass, "pv", "pr", u.Port.PrimaryResult, "discriminant operand %#x mode %v", "dv, mode.pub")
	emitFlagCheck(b, "\t\t\t", u.Symbol, "pf", pfDisc, "discriminant operand %#x mode %v", "dv, mode.pub")
	fmt.Fprintf(b, "\t\t\tmodeSeen[mi] = %s\n", pubBitsExpr(u.ResultClass, "pv"))
	fmt.Fprintf(b, "\t\t\tcount++\n\t\t}\n")
	emitModeDiscAssertion(b, u.Symbol, "discriminant operand %#x", "dv")
	fmt.Fprintf(b, "\t}\n")
	return nil
}

// emitVMModeTernary renders the FMAWithMode parity body: two same-width
// operands plus a RoundingMode, always returning flags. Shared triples leg
// plus the ternary mode-discriminant leg with the per-case assertion (same
// two-leg structure as emitVMModeBinary at arity 3).
func emitVMModeTernary(b *strings.Builder, u parityUnit) error {
	w := u.Width
	op := strings.TrimSuffix(u.Method, "WithMode")
	disc, err := modeTernaryDiscriminantOperands(op, w)
	if err != nil {
		return err
	}

	// Shared routing-corpus leg.
	fmt.Fprintf(b, "\tfor _, tri := range %s {\n", pwTriples(w))
	fmt.Fprintf(b, "\t\tav := %s[tri[0]]\n", pwCorpus(w))
	fmt.Fprintf(b, "\t\tbv := %s[tri[1]]\n", pwCorpus(w))
	fmt.Fprintf(b, "\t\tcv := %s[tri[2]]\n", pwCorpus(w))
	fmt.Fprintf(b, "\t\ta := %s\n", pwPublicVal(w, "av"))
	fmt.Fprintf(b, "\t\tb := %s\n", pwPublicVal(w, "bv"))
	fmt.Fprintf(b, "\t\tc := %s\n", pwPublicVal(w, "cv"))
	fmt.Fprintf(b, "\t\tfor _, mode := range publicParityModes {\n")
	fmt.Fprintf(b, "\t\t\tpv, pf := a.%s(b, c, mode.pub)\n", u.Method)
	pfPort := emitGenericPort(b, "\t\t\t", u.Port, []string{pwPortArg(w, "av"), pwPortArg(w, "bv"), pwPortArg(w, "cv")}, "mode.port", true)
	emitResultCheck(b, "\t\t\t", u.Symbol, u.ResultClass, "pv", "pr", u.Port.PrimaryResult, "operands %#x,%#x,%#x mode %v", "av, bv, cv, mode.pub")
	emitFlagCheck(b, "\t\t\t", u.Symbol, "pf", pfPort, "operands %#x,%#x,%#x mode %v", "av, bv, cv, mode.pub")
	fmt.Fprintf(b, "\t\t\tcount++\n\t\t}\n\t}\n")

	// Mode-discriminant leg.
	discType := modeDiscGoType(w)
	fmt.Fprintf(b, "\tdiscTriples := [][3]%s{\n", discType)
	for _, dt := range disc {
		aLit, err := modeDiscGoLiteral(w, dt[0])
		if err != nil {
			return fmt.Errorf("%s: %w", u.Symbol, err)
		}
		bLit, err := modeDiscGoLiteral(w, dt[1])
		if err != nil {
			return fmt.Errorf("%s: %w", u.Symbol, err)
		}
		cLit, err := modeDiscGoLiteral(w, dt[2])
		if err != nil {
			return fmt.Errorf("%s: %w", u.Symbol, err)
		}
		fmt.Fprintf(b, "\t\t{%s, %s, %s},\n", aLit, bLit, cLit)
	}
	fmt.Fprintf(b, "\t}\n")
	fmt.Fprintf(b, "\tfor _, dt := range discTriples {\n")
	fmt.Fprintf(b, "\t\ta := %s\n", pwPublicVal(w, "dt[0]"))
	fmt.Fprintf(b, "\t\tb := %s\n", pwPublicVal(w, "dt[1]"))
	fmt.Fprintf(b, "\t\tc := %s\n", pwPublicVal(w, "dt[2]"))
	fmt.Fprintf(b, "\t\tvar modeSeen [%d]%s\n", len(parityModeOrder), discType)
	fmt.Fprintf(b, "\t\tfor mi, mode := range publicParityModes {\n")
	fmt.Fprintf(b, "\t\t\tpv, pf := a.%s(b, c, mode.pub)\n", u.Method)
	pfDisc := emitGenericPort(b, "\t\t\t", u.Port, []string{pwPortArg(w, "dt[0]"), pwPortArg(w, "dt[1]"), pwPortArg(w, "dt[2]")}, "mode.port", true)
	emitResultCheck(b, "\t\t\t", u.Symbol, u.ResultClass, "pv", "pr", u.Port.PrimaryResult, "discriminant operands %#x,%#x,%#x mode %v", "dt[0], dt[1], dt[2], mode.pub")
	emitFlagCheck(b, "\t\t\t", u.Symbol, "pf", pfDisc, "discriminant operands %#x,%#x,%#x mode %v", "dt[0], dt[1], dt[2], mode.pub")
	fmt.Fprintf(b, "\t\t\tmodeSeen[mi] = %s\n", pubBitsExpr(u.ResultClass, "pv"))
	fmt.Fprintf(b, "\t\t\tcount++\n\t\t}\n")
	emitModeDiscAssertion(b, u.Symbol, "discriminant operands %#x,%#x,%#x", "dt[0], dt[1], dt[2]")
	fmt.Fprintf(b, "\t}\n")
	return nil
}

// emitVMModeScaleB renders the ScaleBWithMode parity body: an int exponent
// plus a RoundingMode, always returning flags. Shared corpus x exponent leg
// plus the scaleB mode-discriminant leg (operand + boundary exponent tuples)
// with the per-case assertion. scaleB is exact inside the format range, so
// the shared leg alone cannot prove mode forwarding; the discriminant
// exponents sit at the overflow/underflow boundary where the five directions
// split.
func emitVMModeScaleB(b *strings.Builder, u parityUnit) error {
	w := u.Width
	op := strings.TrimSuffix(u.Method, "WithMode")
	disc, err := modeScaleBDiscriminantCases(op, w)
	if err != nil {
		return err
	}

	// Shared routing-corpus leg.
	fmt.Fprintf(b, "\tfor _, elem := range %s {\n", pwCorpus(w))
	fmt.Fprintf(b, "\t\ta := %s\n", pwPublicVal(w, "elem"))
	fmt.Fprintf(b, "\t\tfor _, exp := range publicParityScaleBExps {\n")
	fmt.Fprintf(b, "\t\t\tfor _, mode := range publicParityModes {\n")
	fmt.Fprintf(b, "\t\t\t\tpv, pf := a.%s(exp, mode.pub)\n", u.Method)
	pfPort := emitGenericPort(b, "\t\t\t\t", u.Port, []string{pwPortArg(w, "elem"), scaleBExpArg(u, "exp")}, "mode.port", true)
	emitResultCheck(b, "\t\t\t\t", u.Symbol, u.ResultClass, "pv", "pr", u.Port.PrimaryResult, "operand %#x exp %d mode %v", "elem, exp, mode.pub")
	emitFlagCheck(b, "\t\t\t\t", u.Symbol, "pf", pfPort, "operand %#x exp %d mode %v", "elem, exp, mode.pub")
	fmt.Fprintf(b, "\t\t\t\tcount++\n\t\t\t}\n\t\t}\n\t}\n")

	// Mode-discriminant leg.
	discType := modeDiscGoType(w)
	fmt.Fprintf(b, "\tdiscCases := []struct {\n\t\tv   %s\n\t\texp int\n\t}{\n", discType)
	for _, dc := range disc {
		lit, err := modeDiscGoLiteral(w, dc.Val)
		if err != nil {
			return fmt.Errorf("%s: %w", u.Symbol, err)
		}
		if w == 128 {
			// A struct-field composite literal cannot elide its [16]byte
			// type (elision is an array/slice/map element privilege).
			lit = "[16]byte" + lit
		}
		fmt.Fprintf(b, "\t\t{%s, %d},\n", lit, dc.Exp)
	}
	fmt.Fprintf(b, "\t}\n")
	fmt.Fprintf(b, "\tfor _, dc := range discCases {\n")
	fmt.Fprintf(b, "\t\ta := %s\n", pwPublicVal(w, "dc.v"))
	fmt.Fprintf(b, "\t\tvar modeSeen [%d]%s\n", len(parityModeOrder), discType)
	fmt.Fprintf(b, "\t\tfor mi, mode := range publicParityModes {\n")
	fmt.Fprintf(b, "\t\t\tpv, pf := a.%s(dc.exp, mode.pub)\n", u.Method)
	pfDisc := emitGenericPort(b, "\t\t\t", u.Port, []string{pwPortArg(w, "dc.v"), scaleBExpArg(u, "dc.exp")}, "mode.port", true)
	emitResultCheck(b, "\t\t\t", u.Symbol, u.ResultClass, "pv", "pr", u.Port.PrimaryResult, "discriminant operand %#x exp %d mode %v", "dc.v, dc.exp, mode.pub")
	emitFlagCheck(b, "\t\t\t", u.Symbol, "pf", pfDisc, "discriminant operand %#x exp %d mode %v", "dc.v, dc.exp, mode.pub")
	fmt.Fprintf(b, "\t\t\tmodeSeen[mi] = %s\n", pubBitsExpr(u.ResultClass, "pv"))
	fmt.Fprintf(b, "\t\t\tcount++\n\t\t}\n")
	emitModeDiscAssertion(b, u.Symbol, "discriminant operand %#x exp %d", "dc.v, dc.exp")
	fmt.Fprintf(b, "\t}\n")
	return nil
}

// modeDiscGoLiteral renders one encoded mode-discriminant operand as a Go
// literal of the width's corpus element type.
func modeDiscGoLiteral(width int, o modeDiscOperand) (string, error) {
	switch width {
	case 32:
		bits, err := encodeModeDiscOperand32(o)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("0x%08x", bits), nil
	case 64:
		bits, err := encodeModeDiscOperand64(o)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("0x%016x", bits), nil
	default:
		raw, err := encodeModeDiscOperand128(o)
		if err != nil {
			return "", err
		}
		parts := make([]string, len(raw))
		for i, by := range raw {
			parts[i] = fmt.Sprintf("0x%02x", by)
		}
		return "{" + strings.Join(parts, ", ") + "}", nil
	}
}

func emitVMConvert(b *strings.Builder, u parityUnit) error {
	w := u.Width
	exact := strings.HasSuffix(u.Method, "Exact")
	var base string
	var variants []string
	if exact {
		base = strings.TrimSuffix(u.BidgoFn, "Xrnint")
		variants = []string{"Xrnint", "Xrninta", "Xint", "Xceil", "Xfloor"}
	} else {
		base = strings.TrimSuffix(u.BidgoFn, "Rnint")
		variants = []string{"Rnint", "Rninta", "Int", "Ceil", "Floor"}
	}
	fmt.Fprintf(b, "\tfor _, elem := range %s {\n", pwCorpus(w))
	fmt.Fprintf(b, "\t\ta := %s\n", pwPublicVal(w, "elem"))
	fmt.Fprintf(b, "\t\tfor _, mode := range publicParityModes {\n")
	fmt.Fprintf(b, "\t\t\tpv, pf := a.%s(mode.pub)\n", u.Method)
	fmt.Fprintf(b, "\t\t\tvar pr %s\n\t\t\tvar prf uint32\n", u.Port.PrimaryResult)
	fmt.Fprintf(b, "\t\t\tswitch mode.pub {\n")
	for i, m := range parityModeOrder {
		fmt.Fprintf(b, "\t\t\tcase %s:\n\t\t\t\tpr, prf = bidgo.%s%s(%s)\n", m, base, variants[i], pwPortArg(w, "elem"))
	}
	fmt.Fprintf(b, "\t\t\t}\n")
	emitResultCheck(b, "\t\t\t", u.Symbol, u.ResultClass, "pv", "pr", u.Port.PrimaryResult, "operand %#x mode %v", "elem, mode.pub")
	emitFlagCheck(b, "\t\t\t", u.Symbol, "pf", "prf", "operand %#x mode %v", "elem, mode.pub")
	fmt.Fprintf(b, "\t\t\tcount++\n\t\t}\n\t}\n")
	return nil
}

func emitVMNullary(b *strings.Builder, u parityUnit) error {
	w := u.Width
	var zero string
	if w == 128 {
		zero = "Decimal128BID{}"
	} else {
		zero = fmt.Sprintf("%s(0)", pwDecType(w))
	}
	fmt.Fprintf(b, "\ta := %s\n", zero)
	fmt.Fprintf(b, "\tpv := a.%s()\n", u.Method)
	fmt.Fprintf(b, "\tpr := bidgo.%s()\n", u.BidgoFn)
	fmt.Fprintf(b, "\tif pv != pr {\n\t\tt.Errorf(\"public parity %s: result mismatch public=%%v port=%%v\", pv, pr)\n\t}\n", u.Symbol)
	fmt.Fprintf(b, "\tcount++\n")
	return nil
}

func emitVMCompareTotal(b *strings.Builder, u parityUnit) error {
	w := u.Width
	fmt.Fprintf(b, "\tfor _, pair := range %s {\n", pwPairs(w))
	fmt.Fprintf(b, "\t\tav := %s[pair[0]]\n", pwCorpus(w))
	fmt.Fprintf(b, "\t\tbv := %s[pair[1]]\n", pwCorpus(w))
	fmt.Fprintf(b, "\t\ta := %s\n", pwPublicVal(w, "av"))
	fmt.Fprintf(b, "\t\tb := %s\n", pwPublicVal(w, "bv"))
	fmt.Fprintf(b, "\t\tpv := a.%s(b)\n", u.Method)
	fmt.Fprintf(b, "\t\tfwd := bidgo.%s(%s, %s)\n", u.BidgoFn, pwPortArg(w, "av"), pwPortArg(w, "bv"))
	fmt.Fprintf(b, "\t\trev := bidgo.%s(%s, %s)\n", u.BidgoFn, pwPortArg(w, "bv"), pwPortArg(w, "av"))
	fmt.Fprintf(b, "\t\tvar want int\n")
	fmt.Fprintf(b, "\t\tswitch {\n\t\tcase fwd != 0 && rev != 0:\n\t\t\twant = 0\n\t\tcase fwd != 0:\n\t\t\twant = -1\n\t\tdefault:\n\t\t\twant = 1\n\t\t}\n")
	fmt.Fprintf(b, "\t\tif pv != want {\n\t\t\tt.Errorf(\"public parity %s: operands %%#x,%%#x: result mismatch public=%%v port=%%v\", av, bv, pv, want)\n\t\t}\n", u.Symbol)
	fmt.Fprintf(b, "\t\tcount++\n\t}\n")
	return nil
}

func emitVMSign(b *strings.Builder, u parityUnit) error {
	w := u.Width
	zeroTruth := portTruthExprForWidth("IsZero", w, pwPortArg(w, "elem"))
	signedTruth := portTruthExprForWidth("IsSigned", w, pwPortArg(w, "elem"))
	fmt.Fprintf(b, "\tfor _, elem := range %s {\n", pwCorpus(w))
	fmt.Fprintf(b, "\t\ta := %s\n", pwPublicVal(w, "elem"))
	fmt.Fprintf(b, "\t\tpv := a.Sign()\n")
	fmt.Fprintf(b, "\t\tvar want int\n")
	fmt.Fprintf(b, "\t\tswitch {\n\t\tcase %s:\n\t\t\twant = 0\n\t\tcase %s:\n\t\t\twant = -1\n\t\tdefault:\n\t\t\twant = 1\n\t\t}\n", zeroTruth, signedTruth)
	fmt.Fprintf(b, "\t\tif pv != want {\n\t\t\tt.Errorf(\"public parity %s: operand %%#x: result mismatch public=%%v port=%%v\", elem, pv, want)\n\t\t}\n", u.Symbol)
	fmt.Fprintf(b, "\t\tcount++\n\t}\n")
	return nil
}

func emitVMSignalingEqual(b *strings.Builder, u parityUnit) error {
	w := u.Width
	geFn := fmt.Sprintf("Bid%dSignalingGreaterEqual", w)
	leFn := fmt.Sprintf("Bid%dSignalingLessEqual", w)
	fmt.Fprintf(b, "\tfor _, pair := range %s {\n", pwPairs(w))
	fmt.Fprintf(b, "\t\tav := %s[pair[0]]\n", pwCorpus(w))
	fmt.Fprintf(b, "\t\tbv := %s[pair[1]]\n", pwCorpus(w))
	fmt.Fprintf(b, "\t\ta := %s\n", pwPublicVal(w, "av"))
	fmt.Fprintf(b, "\t\tb := %s\n", pwPublicVal(w, "bv"))
	fmt.Fprintf(b, "\t\tpv, pf := a.%s(b)\n", u.Method)
	fmt.Fprintf(b, "\t\tge, geF := bidgo.%s(%s, %s)\n", geFn, pwPortArg(w, "av"), pwPortArg(w, "bv"))
	fmt.Fprintf(b, "\t\tle, leF := bidgo.%s(%s, %s)\n", leFn, pwPortArg(w, "av"), pwPortArg(w, "bv"))
	fmt.Fprintf(b, "\t\teq := ge != 0 && le != 0\n")
	if u.Shape == shapeVMSignalingNotEqual {
		fmt.Fprintf(b, "\t\twant := !eq\n")
	} else {
		fmt.Fprintf(b, "\t\twant := eq\n")
	}
	fmt.Fprintf(b, "\t\tif pv != want {\n\t\t\tt.Errorf(\"public parity %s: operands %%#x,%%#x: result mismatch public=%%v port=%%v\", av, bv, pv, want)\n\t\t}\n", u.Symbol)
	fmt.Fprintf(b, "\t\tif pf != mapPortFlagsForParity(geF|leF) {\n\t\t\tt.Errorf(\"public parity %s: operands %%#x,%%#x: flag mismatch public=%%v port=%%v\", av, bv, pf, mapPortFlagsForParity(geF|leF))\n\t\t}\n", u.Symbol)
	fmt.Fprintf(b, "\t\tcount++\n\t}\n")
	return nil
}

func emitVMClass(b *strings.Builder, u parityUnit) error {
	w := u.Width
	fmt.Fprintf(b, "\tfor _, elem := range %s {\n", pwCorpus(w))
	fmt.Fprintf(b, "\t\ta := %s\n", pwPublicVal(w, "elem"))
	fmt.Fprintf(b, "\t\tpv := a.Class()\n")
	fmt.Fprintf(b, "\t\tpr := bidgo.%s(%s)\n", u.BidgoFn, pwPortArg(w, "elem"))
	fmt.Fprintf(b, "\t\tif string(pv) != publicParityClassName(pr) {\n\t\t\tt.Errorf(\"public parity %s: operand %%#x: result mismatch public=%%v port=%%v\", elem, string(pv), publicParityClassName(pr))\n\t\t}\n", u.Symbol)
	fmt.Fprintf(b, "\t\tcount++\n\t}\n")
	return nil
}

func emitVMString(b *strings.Builder, u parityUnit) error {
	w := u.Width
	nanTruth := portTruthExprForWidth("IsNaN", w, pwPortArg(w, "elem"))
	fmt.Fprintf(b, "\tfor _, elem := range %s {\n", pwCorpus(w))
	// NaN entries take the hand-built display branch (NaN payload tests own
	// them); they are skipped without counting so the pinned per-unit count
	// equals the number of executed comparisons.
	fmt.Fprintf(b, "\t\tif %s {\n\t\t\tcontinue\n\t\t}\n", nanTruth)
	fmt.Fprintf(b, "\t\ta := %s\n", pwPublicVal(w, "elem"))
	fmt.Fprintf(b, "\t\tpv := a.String()\n")
	fmt.Fprintf(b, "\t\tpr := bidgo.%s(%s)\n", u.BidgoFn, pwPortArg(w, "elem"))
	fmt.Fprintf(b, "\t\tif pv != pr {\n\t\t\tt.Errorf(\"public parity %s: operand %%#x: result mismatch public=%%q port=%%q\", elem, pv, pr)\n\t\t}\n", u.Symbol)
	fmt.Fprintf(b, "\t\tcount++\n\t}\n")
	return nil
}

func emitFuncIntCtor(b *strings.Builder, u parityUnit) error {
	corpus := intCorpusName(u.IntParam)
	fmt.Fprintf(b, "\tfor _, x := range %s {\n", corpus)
	indent := "\t\t"
	loopClose := "\t}\n"
	if u.HasMode {
		fmt.Fprintf(b, "\t\tfor _, mode := range publicParityModes {\n")
		indent = "\t\t\t"
		loopClose = "\t\t}\n\t}\n"
	}
	if u.HasFlags {
		if u.HasMode {
			fmt.Fprintf(b, "%spv, pf := %s(x, mode.pub)\n", indent, u.Func)
		} else {
			fmt.Fprintf(b, "%spv, pf := %s(x)\n", indent, u.Func)
		}
	} else {
		fmt.Fprintf(b, "%spv := %s(x)\n", indent, u.Func)
	}
	modeExpr := "0"
	if u.HasMode {
		modeExpr = "mode.port"
	}
	pfPort := emitGenericPort(b, indent, u.Port, []string{"x"}, modeExpr, u.HasFlags)
	ctxFmt := "operand %v"
	ctxArgs := "x"
	if u.HasMode {
		ctxFmt = "operand %v mode %v"
		ctxArgs = "x, mode.pub"
	}
	emitResultCheck(b, indent, u.Symbol, u.ResultClass, "pv", "pr", u.Port.PrimaryResult, ctxFmt, ctxArgs)
	if u.HasFlags {
		emitFlagCheck(b, indent, u.Symbol, "pf", pfPort, ctxFmt, ctxArgs)
	}
	fmt.Fprintf(b, "%scount++\n", indent)
	b.WriteString(loopClose)
	return nil
}

func emitFuncFromInt(b *strings.Builder, u parityUnit) error {
	corpus := intCorpusName(u.IntParam)
	fmt.Fprintf(b, "\tfor _, x := range %s {\n", corpus)
	fmt.Fprintf(b, "\t\tpv, err := %s(x)\n", u.Func)
	portArg := "int64(x)"
	if u.IntParam == "int32" && u.Port.GoName == "Bid32FromInt32" {
		portArg = "x"
	}
	pfPort := emitGenericPort(b, "\t\t", u.Port, []string{portArg}, "0", true)
	fmt.Fprintf(b, "\t\tshouldError := mapPortFlagsForParity(%s)&FlagInexact != 0\n", pfPort)
	fmt.Fprintf(b, "\t\tif shouldError {\n")
	fmt.Fprintf(b, "\t\t\tif err == nil {\n\t\t\t\tt.Errorf(\"public parity %s: operand %%v: expected exact-representation error, got %%v\", x, pv)\n\t\t\t}\n", u.Symbol)
	switch u.ResultClass {
	case "dec32":
		fmt.Fprintf(b, "\t\t\tif bits := uint32(pv); bits != 0x7c000000 {\n\t\t\t\tt.Errorf(\"public parity %s: operand %%v: error result bits = %%#x, want canonical qNaN 0x7c000000\", x, bits)\n\t\t\t}\n", u.Symbol)
	case "dec64":
		fmt.Fprintf(b, "\t\t\tif bits := uint64(pv); bits != 0x7c00000000000000 {\n\t\t\t\tt.Errorf(\"public parity %s: operand %%v: error result bits = %%#x, want canonical qNaN 0x7c00000000000000\", x, bits)\n\t\t\t}\n", u.Symbol)
	case "dec128":
		// Every int64 is exactly representable as Decimal128, so this branch is
		// unreachable for the current manifest. Keep the generic invalid-result
		// assertion so an unexpected future error cannot pass with a finite value.
		fmt.Fprintf(b, "\t\t\tif !pv.IsNaN() {\n\t\t\t\tt.Errorf(\"public parity %s: operand %%v: error result must be NaN, got %%v\", x, pv)\n\t\t\t}\n", u.Symbol)
	default:
		return fmt.Errorf("func_from_int %s has unsupported result class %q", u.Symbol, u.ResultClass)
	}
	fmt.Fprintf(b, "\t\t} else {\n")
	fmt.Fprintf(b, "\t\t\tif err != nil {\n\t\t\t\tt.Errorf(\"public parity %s: operand %%v: unexpected error %%v\", x, err)\n\t\t\t}\n", u.Symbol)
	emitResultCheck(b, "\t\t\t", u.Symbol, u.ResultClass, "pv", "pr", u.Port.PrimaryResult, "operand %v", "x")
	fmt.Fprintf(b, "\t\t}\n")
	fmt.Fprintf(b, "\t\tcount++\n\t}\n")
	return nil
}

func emitStringZeroResultCheck(b *strings.Builder, indent string, u parityUnit, checkFlags bool) error {
	switch u.ResultClass {
	case "dec32":
		fmt.Fprintf(b, "%sif pv != 0 {\n%s\tt.Errorf(\"public parity %s: input %%q: error result bits = %%#x, want zero\", sc.input, uint32(pv))\n%s}\n", indent, indent, u.Symbol, indent)
	case "dec64":
		fmt.Fprintf(b, "%sif pv != 0 {\n%s\tt.Errorf(\"public parity %s: input %%q: error result bits = %%#x, want zero\", sc.input, uint64(pv))\n%s}\n", indent, indent, u.Symbol, indent)
	case "dec128":
		fmt.Fprintf(b, "%sif pv != (Decimal128BID{}) {\n%s\tt.Errorf(\"public parity %s: input %%q: error result bits = %%x, want zero\", sc.input, pv.ToBytes())\n%s}\n", indent, indent, u.Symbol, indent)
	default:
		return fmt.Errorf("string error-result check %s has unsupported class %q", u.Symbol, u.ResultClass)
	}
	if checkFlags {
		fmt.Fprintf(b, "%sif pf != 0 {\n%s\tt.Errorf(\"public parity %s: input %%q: error flags = %%v, want zero\", sc.input, pf)\n%s}\n", indent, indent, u.Symbol, indent)
	}
	return nil
}

func emitRawRejectedStringCheck(b *strings.Builder, indent string, u parityUnit) error {
	switch u.ResultClass {
	case "dec32":
		fmt.Fprintf(b, "%sif bits := uint32(pv); bits != 0x7c000000 {\n%s\tt.Errorf(\"public parity %s: input %%q: rejected-input result bits = %%#x, want canonical qNaN 0x7c000000\", sc.input, bits)\n%s}\n", indent, indent, u.Symbol, indent)
	case "dec64":
		fmt.Fprintf(b, "%sif bits := uint64(pv); bits != 0x7c00000000000000 {\n%s\tt.Errorf(\"public parity %s: input %%q: rejected-input result bits = %%#x, want canonical qNaN 0x7c00000000000000\", sc.input, bits)\n%s}\n", indent, indent, u.Symbol, indent)
	case "dec128":
		fmt.Fprintf(b, "%sif bits := pv.ToBytes(); bits != ([16]byte{15: 0x7c}) {\n%s\tt.Errorf(\"public parity %s: input %%q: rejected-input result bits = %%x, want canonical qNaN\", sc.input, bits)\n%s}\n", indent, indent, u.Symbol, indent)
	default:
		return fmt.Errorf("raw rejected-string check %s has unsupported class %q", u.Symbol, u.ResultClass)
	}
	fmt.Fprintf(b, "%sif pf != FlagInvalidOperation {\n%s\tt.Errorf(\"public parity %s: input %%q: rejected-input flags = %%v, want FlagInvalidOperation\", sc.input, pf)\n%s}\n", indent, indent, u.Symbol, indent)
	return nil
}

// emitFuncString renders the shared syntax, exact-or-error, flag-returning,
// raw-port, and per-width NaN-payload contracts. Its expected flags come from
// the independent numeric mapPortFlagsForParity mapping, not production code.
func emitFuncString(b *strings.Builder, u parityUnit) error {
	w := u.Width
	pvArg := pubToPortArg(w, "pv")
	pvIsNaN := portTruthExprForWidth("IsNaN", w, pvArg)
	pvIsSignaling := fmt.Sprintf("bidgo.Bid%dIsSignaling(%s) != 0", w, pvArg)
	prIsNaN := portTruthExprForWidth("IsNaN", w, "pr")

	fmt.Fprintf(b, "\tfor _, sc := range publicParityStringCases {\n")
	switch u.StringKind {
	case "raw":
		fmt.Fprintf(b, "\t\tpv, pf := %s(sc.input)\n", u.Func)
		fmt.Fprintf(b, "\t\tif sc.kind == \"nan_literal\" {\n")
		fmt.Fprintf(b, "\t\t\tif sc.nanMinWidth != 0 && sc.nanMinWidth <= %d {\n", w)
		fmt.Fprintf(b, "\t\t\t\tif !(%s) {\n\t\t\t\t\tt.Errorf(\"public parity %s: input %%q: expected a NaN result from the NaN literal branch\", sc.input)\n\t\t\t\t}\n", pvIsNaN, u.Symbol)
		fmt.Fprintf(b, "\t\t\t\tif (%s) != sc.signaling {\n\t\t\t\t\tt.Errorf(\"public parity %s: input %%q: signaling bit mismatch public=%%v literal=%%v\", sc.input, %s, sc.signaling)\n\t\t\t\t}\n", pvIsSignaling, u.Symbol, pvIsSignaling)
		fmt.Fprintf(b, "\t\t\t\tif pf != 0 {\n\t\t\t\t\tt.Errorf(\"public parity %s: input %%q: representable NaN literal must raise no flags, got %%v\", sc.input, pf)\n\t\t\t\t}\n", u.Symbol)
		fmt.Fprintf(b, "\t\t\t} else {\n")
		if err := emitRawRejectedStringCheck(b, "\t\t\t\t", u); err != nil {
			return err
		}
		fmt.Fprintf(b, "\t\t\t}\n")
		fmt.Fprintf(b, "\t\t} else {\n")
		emitGenericPortNamed(b, "\t\t\t", u.Port, []string{"sc.input"}, "0", true, "pr", "prf")
		fmt.Fprintf(b, "\t\t\tsilentCohortCoercion := sc.kind == \"cohort_coercion\" && (sc.cohortMinWidth == 0 || sc.cohortMinWidth > %d) && prf == 0\n", w)
		fmt.Fprintf(b, "\t\t\tmalformedInput := sc.kind == \"blank\" || sc.kind == \"invalid_syntax\"\n")
		fmt.Fprintf(b, "\t\t\tif malformedInput || silentCohortCoercion {\n")
		if err := emitRawRejectedStringCheck(b, "\t\t\t\t", u); err != nil {
			return err
		}
		fmt.Fprintf(b, "\t\t\t} else {\n")
		emitResultCheck(b, "\t\t\t\t", u.Symbol, u.ResultClass, "pv", "pr", u.Port.PrimaryResult, "input %q", "sc.input")
		emitFlagCheck(b, "\t\t\t\t", u.Symbol, "pf", "prf", "input %q", "sc.input")
		fmt.Fprintf(b, "\t\t\t}\n")
		fmt.Fprintf(b, "\t\t}\n")
	case "direct", "withflags":
		if u.StringKind == "withflags" {
			fmt.Fprintf(b, "\t\tpv, pf, err := %s(sc.input)\n", u.Func)
		} else {
			fmt.Fprintf(b, "\t\tpv, err := %s(sc.input)\n", u.Func)
		}
		fmt.Fprintf(b, "\t\tswitch sc.kind {\n")
		fmt.Fprintf(b, "\t\tcase \"blank\", \"invalid_syntax\":\n")
		fmt.Fprintf(b, "\t\t\tif err == nil {\n\t\t\t\tt.Errorf(\"public parity %s: input %%q: rejected input kind %%s must error\", sc.input, sc.kind)\n\t\t\t}\n", u.Symbol)
		if err := emitStringZeroResultCheck(b, "\t\t\t", u, u.StringKind == "withflags"); err != nil {
			return err
		}
		fmt.Fprintf(b, "\t\tcase \"nan_literal\":\n")
		fmt.Fprintf(b, "\t\t\tif sc.nanMinWidth == 0 || sc.nanMinWidth > %d {\n", w)
		fmt.Fprintf(b, "\t\t\t\tif err == nil {\n\t\t\t\t\tt.Errorf(\"public parity %s: input %%q: unrepresentable NaN payload must error\", sc.input)\n\t\t\t\t}\n", u.Symbol)
		if err := emitStringZeroResultCheck(b, "\t\t\t\t", u, u.StringKind == "withflags"); err != nil {
			return err
		}
		fmt.Fprintf(b, "\t\t\t} else {\n")
		fmt.Fprintf(b, "\t\t\t\tif err != nil {\n\t\t\t\t\tt.Errorf(\"public parity %s: input %%q: unexpected error %%v\", sc.input, err)\n\t\t\t\t} else {\n", u.Symbol)
		fmt.Fprintf(b, "\t\t\t\t\tif !(%s) {\n\t\t\t\t\t\tt.Errorf(\"public parity %s: input %%q: expected a NaN result from the NaN literal branch\", sc.input)\n\t\t\t\t\t}\n", pvIsNaN, u.Symbol)
		fmt.Fprintf(b, "\t\t\t\t\tif (%s) != sc.signaling {\n\t\t\t\t\t\tt.Errorf(\"public parity %s: input %%q: signaling bit mismatch public=%%v literal=%%v\", sc.input, %s, sc.signaling)\n\t\t\t\t\t}\n", pvIsSignaling, u.Symbol, pvIsSignaling)
		if u.StringKind == "withflags" {
			fmt.Fprintf(b, "\t\t\t\t\tif pf != 0 {\n\t\t\t\t\t\tt.Errorf(\"public parity %s: input %%q: representable NaN literal must raise no flags, got %%v\", sc.input, pf)\n\t\t\t\t\t}\n", u.Symbol)
		}
		fmt.Fprintf(b, "\t\t\t\t}\n\t\t\t}\n")
		fmt.Fprintf(b, "\t\tdefault:\n")
		emitGenericPortNamed(b, "\t\t\t", u.Port, []string{"sc.input"}, "0", true, "pr", "prf")
		fmt.Fprintf(b, "\t\t\tportFlags := mapPortFlagsForParity(prf)\n")
		fmt.Fprintf(b, "\t\t\tsilentCohortCoercion := sc.kind == \"cohort_coercion\" && (sc.cohortMinWidth == 0 || sc.cohortMinWidth > %d) && prf == 0\n", w)
		fmt.Fprintf(b, "\t\t\tshouldError := (%s) || silentCohortCoercion\n", prIsNaN)
		if u.StringKind == "direct" {
			fmt.Fprintf(b, "\t\t\tshouldError = shouldError || portFlags&(FlagInvalidOperation|FlagOverflow|FlagUnderflow|FlagInexact) != 0\n")
		}
		fmt.Fprintf(b, "\t\t\tswitch {\n")
		fmt.Fprintf(b, "\t\t\tcase shouldError:\n")
		fmt.Fprintf(b, "\t\t\t\tif err == nil {\n\t\t\t\t\tt.Errorf(\"public parity %s: input %%q: expected an error for a NaN or unrepresentable port result\", sc.input)\n\t\t\t\t}\n", u.Symbol)
		if err := emitStringZeroResultCheck(b, "\t\t\t\t", u, u.StringKind == "withflags"); err != nil {
			return err
		}
		fmt.Fprintf(b, "\t\t\tcase err != nil:\n")
		fmt.Fprintf(b, "\t\t\t\tt.Errorf(\"public parity %s: input %%q: unexpected error %%v\", sc.input, err)\n", u.Symbol)
		fmt.Fprintf(b, "\t\t\tdefault:\n")
		emitResultCheck(b, "\t\t\t\t", u.Symbol, u.ResultClass, "pv", "pr", u.Port.PrimaryResult, "input %q", "sc.input")
		if u.StringKind == "withflags" {
			fmt.Fprintf(b, "\t\t\t\tif pf != portFlags {\n\t\t\t\t\tt.Errorf(\"public parity %s: input %%q: flag mismatch public=%%v port=%%v\", sc.input, pf, portFlags)\n\t\t\t\t}\n", u.Symbol)
		}
		fmt.Fprintf(b, "\t\t\t}\n")
		fmt.Fprintf(b, "\t\t}\n")
	default:
		return fmt.Errorf("unsupported string kind %q", u.StringKind)
	}
	fmt.Fprintf(b, "\t\tcount++\n\t}\n")
	return nil
}

// emitFuncStringMode renders the NewDecimal<w>WithMode parity body: the full
// string corpus under every mode (the same documented parsing contract as the
// withflags constructors -- malformed input and unrepresentable NaN payloads
// error, fitting NaN literals take the payload branch with zero flags, and
// complete finite inputs bit/flag-compare against the port at mode.port), then
// the parse mode-discriminant literals
// (decimal strings one digit past the width's precision) with the per-case
// not-all-identical assertion, so a constructor that drops its mode argument
// fails both the per-mode bit compare and the assertion.
func emitFuncStringMode(b *strings.Builder, u parityUnit) error {
	w := u.Width
	disc, err := modeParseDiscriminantStrings(w)
	if err != nil {
		return err
	}
	pvArg := pubToPortArg(w, "pv")
	pvIsNaN := portTruthExprForWidth("IsNaN", w, pvArg)
	pvIsSignaling := fmt.Sprintf("bidgo.Bid%dIsSignaling(%s) != 0", w, pvArg)
	prIsNaN := portTruthExprForWidth("IsNaN", w, "pr")

	// Shared string-corpus leg.
	fmt.Fprintf(b, "\tfor _, sc := range publicParityStringCases {\n")
	fmt.Fprintf(b, "\t\tfor _, mode := range publicParityModes {\n")
	fmt.Fprintf(b, "\t\t\tpv, pf, err := %s(sc.input, mode.pub)\n", u.Func)
	fmt.Fprintf(b, "\t\t\tswitch sc.kind {\n")
	fmt.Fprintf(b, "\t\t\tcase \"blank\", \"invalid_syntax\":\n")
	fmt.Fprintf(b, "\t\t\t\tif err == nil {\n\t\t\t\t\tt.Errorf(\"public parity %s: input %%q mode %%v: rejected input kind %%s must error\", sc.input, mode.pub, sc.kind)\n\t\t\t\t}\n", u.Symbol)
	if err := emitStringZeroResultCheck(b, "\t\t\t\t", u, true); err != nil {
		return err
	}
	fmt.Fprintf(b, "\t\t\tcase \"nan_literal\":\n")
	fmt.Fprintf(b, "\t\t\t\tif sc.nanMinWidth == 0 || sc.nanMinWidth > %d {\n", w)
	fmt.Fprintf(b, "\t\t\t\t\tif err == nil {\n\t\t\t\t\t\tt.Errorf(\"public parity %s: input %%q mode %%v: unrepresentable NaN payload must error\", sc.input, mode.pub)\n\t\t\t\t\t}\n", u.Symbol)
	if err := emitStringZeroResultCheck(b, "\t\t\t\t\t", u, true); err != nil {
		return err
	}
	fmt.Fprintf(b, "\t\t\t\t} else {\n")
	fmt.Fprintf(b, "\t\t\t\t\tif err != nil {\n\t\t\t\t\t\tt.Errorf(\"public parity %s: input %%q mode %%v: unexpected error %%v\", sc.input, mode.pub, err)\n\t\t\t\t\t} else {\n", u.Symbol)
	fmt.Fprintf(b, "\t\t\t\t\t\tif !(%s) {\n\t\t\t\t\t\t\tt.Errorf(\"public parity %s: input %%q mode %%v: expected a NaN result from the NaN literal branch\", sc.input, mode.pub)\n\t\t\t\t\t\t}\n", pvIsNaN, u.Symbol)
	fmt.Fprintf(b, "\t\t\t\t\t\tif (%s) != sc.signaling {\n\t\t\t\t\t\t\tt.Errorf(\"public parity %s: input %%q mode %%v: signaling bit mismatch public=%%v literal=%%v\", sc.input, mode.pub, %s, sc.signaling)\n\t\t\t\t\t\t}\n", pvIsSignaling, u.Symbol, pvIsSignaling)
	fmt.Fprintf(b, "\t\t\t\t\t\tif pf != 0 {\n\t\t\t\t\t\t\tt.Errorf(\"public parity %s: input %%q mode %%v: representable NaN literal must raise no flags, got %%v\", sc.input, mode.pub, pf)\n\t\t\t\t\t\t}\n", u.Symbol)
	fmt.Fprintf(b, "\t\t\t\t\t}\n\t\t\t\t}\n")
	fmt.Fprintf(b, "\t\t\tdefault:\n")
	emitGenericPortNamed(b, "\t\t\t\t", u.Port, []string{"sc.input"}, "mode.port", true, "pr", "prf")
	fmt.Fprintf(b, "\t\t\t\tsilentCohortCoercion := sc.kind == \"cohort_coercion\" && (sc.cohortMinWidth == 0 || sc.cohortMinWidth > %d) && prf == 0\n", w)
	fmt.Fprintf(b, "\t\t\t\tswitch {\n")
	fmt.Fprintf(b, "\t\t\t\tcase (%s) || silentCohortCoercion:\n", prIsNaN)
	fmt.Fprintf(b, "\t\t\t\t\tif err == nil {\n\t\t\t\t\t\tt.Errorf(\"public parity %s: input %%q mode %%v: rejected NaN or unrepresentable cohort result must error\", sc.input, mode.pub)\n\t\t\t\t\t}\n", u.Symbol)
	if err := emitStringZeroResultCheck(b, "\t\t\t\t\t", u, true); err != nil {
		return err
	}
	fmt.Fprintf(b, "\t\t\t\tcase err != nil:\n")
	fmt.Fprintf(b, "\t\t\t\t\tt.Errorf(\"public parity %s: input %%q mode %%v: unexpected error %%v\", sc.input, mode.pub, err)\n", u.Symbol)
	fmt.Fprintf(b, "\t\t\t\tdefault:\n")
	emitResultCheck(b, "\t\t\t\t\t", u.Symbol, u.ResultClass, "pv", "pr", u.Port.PrimaryResult, "input %q mode %v", "sc.input, mode.pub")
	emitFlagCheck(b, "\t\t\t\t\t", u.Symbol, "pf", "prf", "input %q mode %v", "sc.input, mode.pub")
	fmt.Fprintf(b, "\t\t\t\t}\n")
	fmt.Fprintf(b, "\t\t\t}\n")
	fmt.Fprintf(b, "\t\t\tcount++\n\t\t}\n\t}\n")

	// Mode-discriminant leg.
	discType := modeDiscGoType(w)
	fmt.Fprintf(b, "\tdiscInputs := []string{\n")
	for _, ds := range disc {
		fmt.Fprintf(b, "\t\t%q,\n", ds)
	}
	fmt.Fprintf(b, "\t}\n")
	fmt.Fprintf(b, "\tfor _, ds := range discInputs {\n")
	fmt.Fprintf(b, "\t\tvar modeSeen [%d]%s\n", len(parityModeOrder), discType)
	fmt.Fprintf(b, "\t\tfor mi, mode := range publicParityModes {\n")
	fmt.Fprintf(b, "\t\t\tpv, pf, err := %s(ds, mode.pub)\n", u.Func)
	fmt.Fprintf(b, "\t\t\tif err != nil {\n\t\t\t\tt.Errorf(\"public parity %s: discriminant input %%q mode %%v: unexpected error %%v\", ds, mode.pub, err)\n\t\t\t\tcontinue\n\t\t\t}\n", u.Symbol)
	emitGenericPortNamed(b, "\t\t\t", u.Port, []string{"ds"}, "mode.port", true, "pr", "prf")
	emitResultCheck(b, "\t\t\t", u.Symbol, u.ResultClass, "pv", "pr", u.Port.PrimaryResult, "discriminant input %q mode %v", "ds, mode.pub")
	emitFlagCheck(b, "\t\t\t", u.Symbol, "pf", "prf", "discriminant input %q mode %v", "ds, mode.pub")
	fmt.Fprintf(b, "\t\t\tmodeSeen[mi] = %s\n", pubBitsExpr(u.ResultClass, "pv"))
	fmt.Fprintf(b, "\t\t\tcount++\n\t\t}\n")
	emitModeDiscAssertion(b, u.Symbol, "discriminant input %q", "ds")
	fmt.Fprintf(b, "\t}\n")
	return nil
}

func emitFuncContext(b *strings.Builder, u parityUnit) error {
	w := u.Width
	fmt.Fprintf(b, "\tprevDefaultRounding := DefaultArithmeticContext().RoundingMode\n")
	fmt.Fprintf(b, "\tfor _, pair := range %s {\n", pwPairs(w))
	fmt.Fprintf(b, "\t\tav := %s[pair[0]]\n", pwCorpus(w))
	fmt.Fprintf(b, "\t\tbv := %s[pair[1]]\n", pwCorpus(w))
	fmt.Fprintf(b, "\t\ta := %s\n", pwPublicVal(w, "av"))
	fmt.Fprintf(b, "\t\tb := %s\n", pwPublicVal(w, "bv"))
	fmt.Fprintf(b, "\t\tfor _, mode := range publicParityModes {\n")
	fmt.Fprintf(b, "\t\t\tctx := &ArithmeticContext{RoundingMode: mode.pub}\n")
	fmt.Fprintf(b, "\t\t\tpv := %s(a, b, ctx)\n", u.Func)
	pfPort := emitGenericPort(b, "\t\t\t", u.Port, []string{pwPortArg(w, "av"), pwPortArg(w, "bv")}, "mode.port", true)
	emitResultCheck(b, "\t\t\t", u.Symbol, u.ResultClass, "pv", "pr", u.Port.PrimaryResult, "operands %#x,%#x mode %v", "av, bv, mode.pub")
	emitFlagCheck(b, "\t\t\t", u.Symbol, "ctx.Flags", pfPort, "operands %#x,%#x mode %v", "av, bv, mode.pub")
	fmt.Fprintf(b, "\t\t\tcount++\n\t\t}\n")
	// A nil ctx uses the process-global default rounding mode
	// (context_v2.go contextBIDRoundingMode -> defaultRoundingMode), so the
	// comparison must not depend on whatever the global happens to be: pin
	// the default to each mode, compare against the port at that mode, and
	// restore the previous default afterwards.
	fmt.Fprintf(b, "\t\tfor _, mode := range publicParityModes {\n")
	fmt.Fprintf(b, "\t\t\tSetDefaultRounding(mode.pub)\n")
	fmt.Fprintf(b, "\t\t\tpvNil := %s(a, b, nil)\n", u.Func)
	pfNil := emitGenericPortNamed(b, "\t\t\t", u.Port, []string{pwPortArg(w, "av"), pwPortArg(w, "bv")}, "mode.port", false, "prNil", "pfNil")
	_ = pfNil
	emitResultCheck(b, "\t\t\t", u.Symbol, u.ResultClass, "pvNil", "prNil", u.Port.PrimaryResult, "operands %#x,%#x nil-ctx default %v", "av, bv, mode.pub")
	fmt.Fprintf(b, "\t\t\tcount++\n\t\t}\n")
	fmt.Fprintf(b, "\t}\n")
	fmt.Fprintf(b, "\tSetDefaultRounding(prevDefaultRounding)\n")
	return nil
}

// emitGenericPortNamed is emitGenericPort with caller-chosen result/flag var
// names, used where two port invocations share one scope (context nil branch).
func emitGenericPortNamed(b *strings.Builder, indent string, plan parityPortPlan, argExprs []string, modeExpr string, compareFlags bool, resVar, flagVar string) string {
	args := append([]string{}, argExprs...)
	if plan.HasRounding {
		args = append(args, modeExpr)
	}
	switch plan.FlagsKind {
	case "pointer":
		fmt.Fprintf(b, "%svar %s uint32\n", indent, flagVar)
		args = append(args, "&"+flagVar)
		fmt.Fprintf(b, "%s%s := bidgo.%s(%s)\n", indent, resVar, plan.GoName, strings.Join(args, ", "))
		if compareFlags {
			return flagVar
		}
		fmt.Fprintf(b, "%s_ = %s\n", indent, flagVar)
		return "0"
	case "result":
		if compareFlags {
			fmt.Fprintf(b, "%s%s, %s := bidgo.%s(%s)\n", indent, resVar, flagVar, plan.GoName, strings.Join(args, ", "))
			return flagVar
		}
		fmt.Fprintf(b, "%s%s, _ := bidgo.%s(%s)\n", indent, resVar, plan.GoName, strings.Join(args, ", "))
		return "0"
	default:
		fmt.Fprintf(b, "%s%s := bidgo.%s(%s)\n", indent, resVar, plan.GoName, strings.Join(args, ", "))
		return "0"
	}
}

func intCorpusName(paramType string) string {
	switch paramType {
	case "int32":
		return "publicParityIntCorpus32"
	case "uint32":
		return "publicParityUintCorpus32"
	case "int64":
		return "publicParityIntCorpus64"
	case "uint64":
		return "publicParityUintCorpus64"
	default:
		return ""
	}
}

// portTruthExprForWidth renders a boolean expression for a port predicate that
// returns bool at width 32 and int elsewhere (matching the pinned bidgo
// surface), used by the composed Sign/String shapes.
func portTruthExprForWidth(op string, w int, arg string) string {
	call := fmt.Sprintf("bidgo.Bid%d%s(%s)", w, op, arg)
	// Bid32IsZero / Bid32IsNaN return bool; every other predicate/width returns int.
	if w == 32 && (op == "IsZero" || op == "IsNaN") {
		return call
	}
	return call + " != 0"
}

func emitPublicParityCases(units []parityUnit) []byte {
	total := 0
	byShape := map[string]int{}
	for _, u := range units {
		total += u.Cases
		byShape[shapeName(u.Shape)] += u.Cases
	}

	var b strings.Builder
	b.WriteString(genmarker.Line("testgen") + "\n\n")
	b.WriteString("package bid754\n\n")
	b.WriteString("import \"testing\"\n\n")
	b.WriteString(`// The public-API parity gate exercises every mapped public wrapper against an
// independent invocation of its pinned Go mechanical port function on a
// deterministic bit-literal corpus, comparing result bits and mapped exception
// flags. It is an architecture-contract gate (does the wrapper route through
// the port and preserve semantics?), not a fifth regular verification domain.
// Case counts are pinned here at generation time so a generator regression that
// shrinks the corpus cannot silently re-pin a smaller surface.
`)
	fmt.Fprintf(&b, "const (\n\texpectedPublicParityWrappers = %d\n\texpectedPublicParityCases    = %d\n)\n\n", len(units), total)

	b.WriteString("var expectedPublicParityCasesByShape = map[string]int{\n")
	shapeNames := make([]string, 0, len(byShape))
	for name := range byShape {
		shapeNames = append(shapeNames, name)
	}
	sort.Strings(shapeNames)
	for _, name := range shapeNames {
		fmt.Fprintf(&b, "\t%q: %d,\n", name, byShape[name])
	}
	b.WriteString("}\n\n")

	b.WriteString(`func TestGeneratedPublicAPIParity(t *testing.T) {
	if testing.Short() {
		t.Skip("public-API parity runs in non-short mode; it exercises the full corpus")
	}
	if len(publicParityUnits) != expectedPublicParityWrappers {
		t.Fatalf("expected %d parity wrappers, got %d", expectedPublicParityWrappers, len(publicParityUnits))
	}
	total := 0
	byShape := map[string]int{}
	for _, u := range publicParityUnits {
		n := u.run(t)
		if n == 0 {
			t.Errorf("public parity wrapper %q ran zero cases", u.name)
		}
		total += n
		byShape[u.shape] += n
	}
	if total != expectedPublicParityCases {
		t.Fatalf("expected %d parity cases, ran %d", expectedPublicParityCases, total)
	}
	for shape, want := range expectedPublicParityCasesByShape {
		if byShape[shape] != want {
			t.Errorf("public parity shape %q: expected %d cases, ran %d", shape, want, byShape[shape])
		}
	}
	if len(byShape) != len(expectedPublicParityCasesByShape) {
		t.Errorf("public parity: ran %d shapes, pinned %d", len(byShape), len(expectedPublicParityCasesByShape))
	}
}
`)
	return []byte(b.String())
}

func shapeName(s parityShape) string {
	switch s {
	case shapeVMUnary:
		return "vm_unary"
	case shapeVMBinary:
		return "vm_binary"
	case shapeVMTernary:
		return "vm_ternary"
	case shapeVMScaleB:
		return "vm_scaleb"
	case shapeVMNextToward:
		return "vm_nexttoward"
	case shapeVMModeUnary:
		return "vm_mode_unary"
	case shapeVMModeUnaryArith:
		return "vm_mode_unary_arith"
	case shapeVMModeBinary:
		return "vm_mode_binary"
	case shapeVMModeTernary:
		return "vm_mode_ternary"
	case shapeVMModeScaleB:
		return "vm_mode_scaleb"
	case shapeVMConvert:
		return "vm_convert"
	case shapeVMNullary:
		return "vm_nullary"
	case shapeVMCompareTotal:
		return "vm_compare_total"
	case shapeVMSign:
		return "vm_sign"
	case shapeVMSignalingEqual:
		return "vm_signaling_equal"
	case shapeVMSignalingNotEqual:
		return "vm_signaling_not_equal"
	case shapeVMClass:
		return "vm_class"
	case shapeVMString:
		return "vm_string"
	case shapeFuncIntCtor:
		return "func_int_ctor"
	case shapeFuncFromInt:
		return "func_from_int"
	case shapeFuncString:
		return "func_string"
	case shapeFuncStringMode:
		return "func_string_mode"
	case shapeFuncContext:
		return "func_context"
	case shapeFuncMixedModeBinary:
		return "func_mixed_mode_binary"
	case shapeFuncMixedModeTernary:
		return "func_mixed_mode_ternary"
	case shapeFuncMixedModeUnary:
		return "func_mixed_mode_unary"
	default:
		return "unknown"
	}
}
