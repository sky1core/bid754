package apiemit

import (
	"fmt"
	"sort"
	"strings"
)

// This file holds the width-parameterized shape-family emitters for the
// Decimal<w> wrapper surface. Each emitter is a thin mechanical template over
// a manifest-declared shape; none reproduces arithmetic.
//
// buildDecimalRs (rust_templates.go) collects the manifest emit rows for one
// width into the op slices below and calls these emitters in a fixed order
// for byte-reproducible output.

// decOp is one manifest-driven Decimal<w> wrapper: a Rust method name plus
// the generated port module/function it routes through. boolPort records
// whether that port function already returns Rust bool (Decimal32's
// bid32_exports.go convenience predicates) rather than the Intel-mirroring
// numeric type every other port function uses (compared via "!= 0"); it is
// populated from boolResultPorts when the op is built, never guessed. pfpsf
// records whether the port function uses the width-128-only
// pfpsf-output-parameter calling convention instead of the tuple-return
// convention; it is populated from portPfpsf, never guessed.
type decOp struct {
	method   string
	port     string
	module   string
	boolPort bool
	pfpsf    bool
}

type mixedDecOp struct {
	decOp
	left  widthSpec
	right widthSpec
}

// flagsCallStmt renders the Rust statement(s) that bind valueVar (the port
// call's raw result) and flagsVar (the raised-flags u32) from one
// crate::generated::* port call, dispatching on op.pfpsf: the
// pfpsf-output-parameter convention (a fixed subset of Decimal128 arithmetic
// functions, with one generation point for the fork) or the
// tuple-return convention every other operation (every 32/64 op, and the
// majority of Decimal128 ops -- verified per-function against the actual
// generated:: signatures, not assumed uniform) uses. args are the already
// width-converted port call argument expressions in order (via
// widthSpec.selfArg), NOT including a trailing rounding-mode literal (the
// caller appends that to args itself, same as before) or the pfpsf pointer
// (this appends that itself when op.pfpsf is set). flagsVar is the local
// binding name for the flags word: "raw" when the shape keeps it or "_raw"
// when the shape drops it (mirrors the existing tuple-destructure
// convention -- Rust's unused-variable lint only fires on a plain
// destructured binding that is never read afterward, not on a variable whose
// address is taken via `&mut`, so the pfpsf branch needs no such distinction
// and "raw"/"_raw" both work unchanged). valueVar is normally "bits", except
// emitCompareBoolFlagsOps's "truth" naming.
func flagsCallStmt(op decOp, args []string, valueVar, flagsVar string) string {
	call := "crate::generated::" + op.module + "::" + op.port
	if op.pfpsf {
		allArgs := append(append([]string{}, args...), "&mut "+flagsVar)
		return "let mut " + flagsVar + " = 0u32;\n        let " + valueVar + " = " + call + "(" + strings.Join(allArgs, ", ") + ");"
	}
	return "let (" + valueVar + ", " + flagsVar + ") = " + call + "(" + strings.Join(args, ", ") + ");"
}

// decConvOp is one ConvertToInt<N>/ConvertToUint<N>(Exact) wrapper: the Rust
// method name, the target Rust integer type, and whether it is the
// inexact-signaling ("Exact") family. The five RoundingMode-dispatch port
// function names are derived from the manifest's canonical bidgo_function.
// Width-independent: the target integer type is fixed by the shape
// (to_i8..to_u64), not by the decimal receiver's own width.
type decConvOp struct {
	method    string
	rustType  string
	canonical string
	exact     bool
}

func sortDecOps(ops []decOp) {
	sort.Slice(ops, func(i, j int) bool { return ops[i].method < ops[j].method })
}

// convShapeTypes is the closed table mapping each ConvertToInt<N>/
// ConvertToUint<N>(Exact) shape name to its target Rust integer type and
// whether it is the inexact-signaling ("Exact") family. Explicit and closed
// rather than derived from the shape string, per the no-implicit-inference
// project convention. Shared by every width: the shape name alone fixes the
// target integer type.
var convShapeTypes = map[string]struct {
	rustType string
	exact    bool
}{
	"to_i8": {"i8", false}, "to_i8_exact": {"i8", true},
	"to_i16": {"i16", false}, "to_i16_exact": {"i16", true},
	"to_i32": {"i32", false}, "to_i32_exact": {"i32", true},
	"to_i64": {"i64", false}, "to_i64_exact": {"i64", true},
	"to_u8": {"u8", false}, "to_u8_exact": {"u8", true},
	"to_u16": {"u16", false}, "to_u16_exact": {"u16", true},
	"to_u32": {"u32", false}, "to_u32_exact": {"u32", true},
	"to_u64": {"u64", false}, "to_u64_exact": {"u64", true},
}

// emitUnaryOps renders "unary" shape wrappers: (recv) Op() Decimal<w>, no
// flags, no rounding mode (Abs, Negate). neg is annotated against
// clippy::should_implement_trait: it is an inherent named operation (matches
// the Go .Negate method), and std::ops::Neg is outside the declared generated
// surface.
func emitUnaryOps(b *strings.Builder, ops []decOp, w widthSpec) {
	sortDecOps(ops)
	for _, op := range ops {
		allowAttr := ""
		if op.method == "neg" {
			allowAttr = "    #[allow(clippy::should_implement_trait)]\n"
		}
		call := fmt.Sprintf("crate::generated::%s::%s(%s)", op.module, op.port, w.selfArg("self"))
		fmt.Fprintf(b, `
%s    pub fn %s(self) -> %s {
        %s
    }
`, allowAttr, op.method, w.selfType, w.wrapResult(call))
	}
}

// emitPredicateOps renders "predicate" shape wrappers: (recv) Op() bool (the
// 9 IEEE isXxx predicates). Width-independent signature; per-op boolPort
// decides whether the port result is used directly (already bool) or
// compared "!= 0" (the Intel-mirroring numeric convention).
func emitPredicateOps(b *strings.Builder, ops []decOp, w widthSpec) {
	sortDecOps(ops)
	for _, op := range ops {
		call := fmt.Sprintf("crate::generated::%s::%s(%s)", op.module, op.port, w.selfArg("self"))
		if !op.boolPort {
			call += " != 0"
		}
		fmt.Fprintf(b, `
    pub fn %s(self) -> bool {
        %s
    }
`, op.method, call)
	}
}

// emitUnaryFlagsNoRoundOps renders unary ops returning (Decimal<w>,
// ExceptionFlags) whose port call takes no rounding-mode argument (LogB,
// NextMinus/NextPlus, the 5 non-exact RoundIntegral* variants).
func emitUnaryFlagsNoRoundOps(b *strings.Builder, ops []decOp, w widthSpec) {
	sortDecOps(ops)
	for _, op := range ops {
		stmt := flagsCallStmt(op, []string{w.selfArg("self")}, "bits", "raw")
		fmt.Fprintf(b, `
    pub fn %s(self) -> (%s, ExceptionFlags) {
        %s
        (%s, ExceptionFlags::from_bidgo(raw))
    }
`, op.method, w.selfType, stmt, w.wrapResult("bits"))
	}
}

// emitUnaryFlagsDefaultRoundOps renders unary ops returning (Decimal<w>,
// ExceptionFlags) whose port call takes a rounding mode; the wrapper always
// passes default round-nearest-even (Sqrt, RoundIntegralExactWithFlags).
func emitUnaryFlagsDefaultRoundOps(b *strings.Builder, ops []decOp, w widthSpec) {
	sortDecOps(ops)
	for _, op := range ops {
		stmt := flagsCallStmt(op, []string{w.selfArg("self"), "BIDGO_ROUND_NEAREST_EVEN"}, "bits", "raw")
		fmt.Fprintf(b, `
    pub fn %s(self) -> (%s, ExceptionFlags) {
        %s
        (%s, ExceptionFlags::from_bidgo(raw))
    }
`, op.method, w.selfType, stmt, w.wrapResult("bits"))
	}
}

// emitBinaryFlagsNoRoundOps renders 2-operand ops returning (Decimal<w>,
// ExceptionFlags) whose port call takes no rounding-mode argument (Fmod,
// MaxNum, MaxNumMag, MinNum, MinNumMag, Remainder).
func emitBinaryFlagsNoRoundOps(b *strings.Builder, ops []decOp, w widthSpec) {
	sortDecOps(ops)
	for _, op := range ops {
		stmt := flagsCallStmt(op, []string{w.selfArg("self"), w.selfArg("rhs")}, "bits", "raw")
		fmt.Fprintf(b, `
    pub fn %s(self, rhs: %s) -> (%s, ExceptionFlags) {
        %s
        (%s, ExceptionFlags::from_bidgo(raw))
    }
`, op.method, w.selfType, w.selfType, stmt, w.wrapResult("bits"))
	}
}

// emitCompareBoolFlagsOps renders 2-operand comparison predicates returning
// (bool, ExceptionFlags): the 12 direct Quiet* ops and the 8 direct
// Signaling* ops (SignalingEqual/NotEqual are composed elsewhere, since
// Intel has no signaling-equal entrypoint).
func emitCompareBoolFlagsOps(b *strings.Builder, ops []decOp, w widthSpec) {
	sortDecOps(ops)
	for _, op := range ops {
		stmt := flagsCallStmt(op, []string{w.selfArg("self"), w.selfArg("rhs")}, "truth", "raw")
		fmt.Fprintf(b, `
    pub fn %s(self, rhs: %s) -> (bool, ExceptionFlags) {
        %s
        (truth != 0, ExceptionFlags::from_bidgo(raw))
    }
`, op.method, w.selfType, stmt)
	}
}

// emitCopySignOp renders CopySign: 2 operands, no rounding mode, no flags.
func emitCopySignOp(b *strings.Builder, op decOp, w widthSpec) {
	call := fmt.Sprintf("crate::generated::%s::%s(%s, %s)", op.module, op.port, w.selfArg("self"), w.selfArg("sign_source"))
	fmt.Fprintf(b, `
    pub fn %s(self, sign_source: %s) -> %s {
        %s
    }
`, op.method, w.selfType, w.selfType, w.wrapResult(call))
}

// emitFMAOp renders FMA: 3 operands, default rounding, flags.
// modeMethodDocSentences pins the rustdoc first sentence of every
// explicit-rounding-mode wrapper method the mode shapes emit. These sentences
// ship as public docs.rs documentation, so each must state what the operation
// actually does and that it takes an explicit rounding mode; a mode-shape
// method missing here fails generation (modeMethodDocSentence) instead of
// emitting a generic fallback sentence that misdescribes the operation.
var modeMethodDocSentences = map[string]string{
	"add_with_mode":      "Adds two decimals with an explicit rounding mode, also returning the raised flags.",
	"sub_with_mode":      "Subtracts two decimals with an explicit rounding mode, also returning the raised flags.",
	"mul_with_mode":      "Multiplies two decimals with an explicit rounding mode, also returning the raised flags.",
	"div_with_mode":      "Divides two decimals with an explicit rounding mode, also returning the raised flags.",
	"quantize_with_mode": "Rescales self to the quantum (target exponent) of rhs with an explicit rounding mode, also returning the raised flags.",
	"sqrt_with_mode":     "Computes the square root of self with an explicit rounding mode, also returning the raised flags.",
	"fma_with_mode":      "Computes the fused multiply-add self * mul + add with a single rounding at an explicit rounding mode, also returning the raised flags.",
	"scaleb_with_mode":   "Scales self by ten to the integer exponent, rounding at the overflow/underflow range boundary with an explicit rounding mode, also returning the raised flags.",
}

var mixedModeMethodDocSentences = map[string]string{
	"Decimal64.add_dq_with_mode":  "Adds a Decimal64 left operand and Decimal128 right operand, rounding directly to Decimal64 with an explicit rounding mode and returning the raised flags.",
	"Decimal64.add_qd_with_mode":  "Adds a Decimal128 left operand and Decimal64 right operand, rounding directly to Decimal64 with an explicit rounding mode and returning the raised flags.",
	"Decimal64.add_qq_with_mode":  "Adds two Decimal128 operands, rounding directly to Decimal64 with an explicit rounding mode and returning the raised flags.",
	"Decimal64.sub_dq_with_mode":  "Subtracts a Decimal128 right operand from a Decimal64 left operand, rounding directly to Decimal64 with an explicit rounding mode and returning the raised flags.",
	"Decimal64.sub_qd_with_mode":  "Subtracts a Decimal64 right operand from a Decimal128 left operand, rounding directly to Decimal64 with an explicit rounding mode and returning the raised flags.",
	"Decimal64.sub_qq_with_mode":  "Subtracts two Decimal128 operands, rounding directly to Decimal64 with an explicit rounding mode and returning the raised flags.",
	"Decimal64.mul_dq_with_mode":  "Multiplies a Decimal64 left operand by a Decimal128 right operand, rounding directly to Decimal64 with an explicit rounding mode and returning the raised flags.",
	"Decimal64.mul_qd_with_mode":  "Multiplies a Decimal128 left operand by a Decimal64 right operand, rounding directly to Decimal64 with an explicit rounding mode and returning the raised flags.",
	"Decimal64.mul_qq_with_mode":  "Multiplies two Decimal128 operands, rounding directly to Decimal64 with an explicit rounding mode and returning the raised flags.",
	"Decimal64.div_dq_with_mode":  "Divides a Decimal64 left operand by a Decimal128 right operand, rounding directly to Decimal64 with an explicit rounding mode and returning the raised flags.",
	"Decimal64.div_qd_with_mode":  "Divides a Decimal128 left operand by a Decimal64 right operand, rounding directly to Decimal64 with an explicit rounding mode and returning the raised flags.",
	"Decimal64.div_qq_with_mode":  "Divides two Decimal128 operands, rounding directly to Decimal64 with an explicit rounding mode and returning the raised flags.",
	"Decimal128.add_dd_with_mode": "Adds two Decimal64 operands to produce Decimal128 with an explicit rounding mode and returns the raised flags.",
	"Decimal128.add_dq_with_mode": "Adds a Decimal64 left operand and Decimal128 right operand to produce Decimal128 with an explicit rounding mode and returns the raised flags.",
	"Decimal128.add_qd_with_mode": "Adds a Decimal128 left operand and Decimal64 right operand to produce Decimal128 with an explicit rounding mode and returns the raised flags.",
	"Decimal128.sub_dd_with_mode": "Subtracts two Decimal64 operands to produce Decimal128 with an explicit rounding mode and returns the raised flags.",
	"Decimal128.sub_dq_with_mode": "Subtracts a Decimal128 right operand from a Decimal64 left operand to produce Decimal128 with an explicit rounding mode and returns the raised flags.",
	"Decimal128.sub_qd_with_mode": "Subtracts a Decimal64 right operand from a Decimal128 left operand to produce Decimal128 with an explicit rounding mode and returns the raised flags.",
	"Decimal128.mul_dd_with_mode": "Multiplies two Decimal64 operands to produce an exact Decimal128 result and returns the raised flags.",
	"Decimal128.mul_dq_with_mode": "Multiplies a Decimal64 left operand by a Decimal128 right operand to produce Decimal128 with an explicit rounding mode and returns the raised flags.",
	"Decimal128.mul_qd_with_mode": "Multiplies a Decimal128 left operand by a Decimal64 right operand to produce Decimal128 with an explicit rounding mode and returns the raised flags.",
	"Decimal128.div_dd_with_mode": "Divides two Decimal64 operands to produce Decimal128 with an explicit rounding mode and returns the raised flags.",
	"Decimal128.div_dq_with_mode": "Divides a Decimal64 left operand by a Decimal128 right operand to produce Decimal128 with an explicit rounding mode and returns the raised flags.",
	"Decimal128.div_qd_with_mode": "Divides a Decimal128 left operand by a Decimal64 right operand to produce Decimal128 with an explicit rounding mode and returns the raised flags.",
}

func modeMethodDocSentence(method string) (string, error) {
	doc, ok := modeMethodDocSentences[method]
	if !ok {
		return "", fmt.Errorf("apiemit: no rustdoc sentence for mode-shape method %q; add an accurate one to modeMethodDocSentences (public docs must state the operation, not a generic fallback)", method)
	}
	return doc, nil
}

func mixedModeMethodDocSentence(owner, method string) (string, error) {
	key := owner + "." + method
	doc, ok := mixedModeMethodDocSentences[key]
	if !ok {
		return "", fmt.Errorf("apiemit: no rustdoc sentence for mixed mode-shape method %q; add an exact owner+method entry to mixedModeMethodDocSentences", key)
	}
	return doc, nil
}

// emitBinaryModeFlagsOps renders "binary_mode_flags" shape wrappers: (recv)
// OpWithMode(rhs, mode) (Decimal<w>, ExceptionFlags). This is the binsFlags
// (binary_with_flags) template with the port's rounding argument taken from an
// explicit RoundingMode parameter via super::types::to_bidgo_rounding (the same
// mode-conversion path emitToDecimal32Op/emitFromIntModeOp use) instead of the
// hardcoded BIDGO_ROUND_NEAREST_EVEN. No arithmetic is reproduced; the port
// call and its per-width flags convention (pfpsf vs tuple) come from
// flagsCallStmt, identical to every other flags-carrying shape.
func emitBinaryModeFlagsOps(b *strings.Builder, ops []decOp, w widthSpec) error {
	sortDecOps(ops)
	for _, op := range ops {
		doc, err := modeMethodDocSentence(op.method)
		if err != nil {
			return err
		}
		stmt := flagsCallStmt(op, []string{w.selfArg("self"), w.selfArg("rhs"), "super::types::to_bidgo_rounding(mode)"}, "bits", "raw")
		fmt.Fprintf(b, `
    /// %s
    pub fn %s(self, rhs: %s, mode: RoundingMode) -> (%s, ExceptionFlags) {
        %s
        (%s, ExceptionFlags::from_bidgo(raw))
    }
`, doc, op.method, w.selfType, w.selfType, stmt, w.wrapResult("bits"))
	}
	return nil
}

// emitMixedBinaryModeFlagsOps renders destination-owner associated functions
// for the Intel D/Q mixed-width families. They are not receiver methods: the
// first operand is not necessarily the result width. Both operands therefore
// use the public cross-module bit accessors, and the result is wrapped in the
// destination module exactly once.
func emitMixedBinaryModeFlagsOps(b *strings.Builder, ops []mixedDecOp, result widthSpec) error {
	sort.Slice(ops, func(i, j int) bool { return ops[i].method < ops[j].method })
	for _, op := range ops {
		doc, err := mixedModeMethodDocSentence(result.selfType, op.method)
		if err != nil {
			return err
		}
		stmt := flagsCallStmt(op.decOp, []string{op.left.crossModuleArg("left"), op.right.crossModuleArg("right"), "super::types::to_bidgo_rounding(mode)"}, "bits", "raw")
		fmt.Fprintf(b, `
    /// %s
    pub fn %s(left: %s, right: %s, mode: RoundingMode) -> (%s, ExceptionFlags) {
        %s
        (%s, ExceptionFlags::from_bidgo(raw))
    }
`, doc, op.method, op.left.selfType, op.right.selfType, result.selfType, stmt, result.wrapResult("bits"))
	}
	return nil
}

// emitUnaryModeFlagsOps renders "unary_mode_flags" shape wrappers: (recv)
// OpWithMode(mode) (Decimal<w>, ExceptionFlags) -- the receiver is the only
// operand and the port's rounding argument comes from the explicit
// RoundingMode via super::types::to_bidgo_rounding (Sqrt). Same mechanical
// pattern as emitBinaryModeFlagsOps at arity 1.
func emitUnaryModeFlagsOps(b *strings.Builder, ops []decOp, w widthSpec) error {
	sortDecOps(ops)
	for _, op := range ops {
		doc, err := modeMethodDocSentence(op.method)
		if err != nil {
			return err
		}
		stmt := flagsCallStmt(op, []string{w.selfArg("self"), "super::types::to_bidgo_rounding(mode)"}, "bits", "raw")
		fmt.Fprintf(b, `
    /// %s
    pub fn %s(self, mode: RoundingMode) -> (%s, ExceptionFlags) {
        %s
        (%s, ExceptionFlags::from_bidgo(raw))
    }
`, doc, op.method, w.selfType, stmt, w.wrapResult("bits"))
	}
	return nil
}

// emitTernaryModeFlagsOps renders "ternary_mode_flags" shape wrappers: (recv)
// FMAWithMode(mul, add, mode) (Decimal<w>, ExceptionFlags) -- emitFMAOp with
// the explicit-mode rounding argument instead of the hardcoded
// BIDGO_ROUND_NEAREST_EVEN.
func emitTernaryModeFlagsOps(b *strings.Builder, ops []decOp, w widthSpec) error {
	sortDecOps(ops)
	for _, op := range ops {
		doc, err := modeMethodDocSentence(op.method)
		if err != nil {
			return err
		}
		stmt := flagsCallStmt(op, []string{w.selfArg("self"), w.selfArg("mul"), w.selfArg("add"), "super::types::to_bidgo_rounding(mode)"}, "bits", "raw")
		fmt.Fprintf(b, `
    /// %s
    pub fn %s(self, mul: %s, add: %s, mode: RoundingMode) -> (%s, ExceptionFlags) {
        %s
        (%s, ExceptionFlags::from_bidgo(raw))
    }
`, doc, op.method, w.selfType, w.selfType, w.selfType, stmt, w.wrapResult("bits"))
	}
	return nil
}

// emitScalebModeOps renders "scaleb_mode" shape wrappers: (recv)
// ScaleBWithMode(exponent, mode) (Decimal<w>, ExceptionFlags). The supported
// Go platforms are LP64, so the public Rust exponent remains i64 all the way
// into the scalbln port entrypoint instead of narrowing Go's int domain.
func emitScalebModeOps(b *strings.Builder, ops []decOp, w widthSpec) error {
	sortDecOps(ops)
	for _, op := range ops {
		doc, err := modeMethodDocSentence(op.method)
		if err != nil {
			return err
		}
		stmt := flagsCallStmt(op, []string{w.selfArg("self"), "exponent", "super::types::to_bidgo_rounding(mode)"}, "bits", "raw")
		fmt.Fprintf(b, `
    /// %s
    pub fn %s(self, exponent: i64, mode: RoundingMode) -> (%s, ExceptionFlags) {
        %s
        (%s, ExceptionFlags::from_bidgo(raw))
    }
`, doc, op.method, w.selfType, stmt, w.wrapResult("bits"))
	}
	return nil
}

func emitFMAOp(b *strings.Builder, op decOp, w widthSpec) {
	stmt := flagsCallStmt(op, []string{w.selfArg("self"), w.selfArg("mul"), w.selfArg("add"), "BIDGO_ROUND_NEAREST_EVEN"}, "bits", "raw")
	fmt.Fprintf(b, `
    pub fn %s(self, mul: %s, add: %s) -> (%s, ExceptionFlags) {
        %s
        (%s, ExceptionFlags::from_bidgo(raw))
    }
`, op.method, w.selfType, w.selfType, w.selfType, stmt, w.wrapResult("bits"))
}

// emitSameQuantumOp renders SameQuantum: 2 operands, bool, no flags. Like
// emitPredicateOps, op.boolPort decides direct-return vs "!= 0".
func emitSameQuantumOp(b *strings.Builder, op decOp, w widthSpec) {
	call := fmt.Sprintf("crate::generated::%s::%s(%s, %s)", op.module, op.port, w.selfArg("self"), w.selfArg("rhs"))
	if !op.boolPort {
		call += " != 0"
	}
	fmt.Fprintf(b, `
    pub fn %s(self, rhs: %s) -> bool {
        %s
    }
`, op.method, w.selfType, call)
}

// emitClassOp renders Class(): maps the port's raw class index through the
// shared decimal_class_from_bid_class helper (types.rs). Width-independent
// (the port's class-index return is int/i64 for every width).
func emitClassOp(b *strings.Builder, op decOp, w widthSpec) {
	fmt.Fprintf(b, `
    pub fn %s(self) -> DecimalClass {
        super::types::decimal_class_from_bid_class(crate::generated::%s::%s(%s))
    }
`, op.method, op.module, op.port, w.selfArg("self"))
}

// emitTotalCmpOp renders CompareTotal/CompareTotalMag: composed from two
// directional port calls via the shared total_order_comparison helper
// (types.rs), surfaced as core::cmp::Ordering (f64::total_cmp idiom).
func emitTotalCmpOp(b *strings.Builder, op decOp, w widthSpec) {
	fmt.Fprintf(b, `
    pub fn %s(self, rhs: %s) -> core::cmp::Ordering {
        super::types::total_order_comparison(
            crate::generated::%s::%s(%s, %s),
            crate::generated::%s::%s(%s, %s),
        )
    }
`, op.method, w.selfType, op.module, op.port, w.selfArg("self"), w.selfArg("rhs"), op.module, op.port, w.selfArg("rhs"), w.selfArg("self"))
}

// emitSignOp renders Sign(): composed from IsZero/IsSigned (there is no
// single Intel sign entrypoint), matching decimal<w>BIDSignPort exactly.
// Width-independent signature; isZero/isSigned are each compared through
// their own boolPort (Decimal32's IsZero is bool-native; IsSigned is not).
func emitSignOp(b *strings.Builder, method string, isZero, isSigned decOp, w widthSpec) {
	zeroCall := fmt.Sprintf("crate::generated::%s::%s(%s)", isZero.module, isZero.port, w.selfArg("self"))
	if !isZero.boolPort {
		zeroCall += " != 0"
	}
	signedCall := fmt.Sprintf("crate::generated::%s::%s(%s)", isSigned.module, isSigned.port, w.selfArg("self"))
	if !isSigned.boolPort {
		signedCall += " != 0"
	}
	fmt.Fprintf(b, `
    pub fn %s(self) -> i32 {
        if %s {
            0
        } else if %s {
            -1
        } else {
            1
        }
    }
`, method, zeroCall, signedCall)
}

// emitRadixConstOp renders Radix() as an associated const: the IEEE 754-2019
// decimal radix is fixed at 10 for every conforming BID width, so this is a
// spec invariant rather than a runtime port call (mirrors the
// existing precedent of hardcoding the fixed Intel exception-flag bit values
// in ExceptionFlags::from_bidgo).
func emitRadixConstOp(b *strings.Builder, name string, w widthSpec) {
	fmt.Fprintf(b, `
    /// The radix of the %s format (always 10; IEEE 754-2019 clause
    /// 5.7.2 radix(x)). A fixed spec invariant, not a runtime port call.
    pub const %s: u32 = 10;
`, w.selfType, name)
}

// emitQuantizeDropOp renders Quantize(): a value-only wrapper over the
// flags-exposing Bid<w>Quantize port entrypoint, default rounding, flags
// dropped by the public signature (matches decimal<w>BIDQuantizePort).
func emitQuantizeDropOp(b *strings.Builder, op decOp, w widthSpec) {
	stmt := flagsCallStmt(op, []string{w.selfArg("self"), w.selfArg("rhs"), "BIDGO_ROUND_NEAREST_EVEN"}, "bits", "_raw")
	fmt.Fprintf(b, `
    pub fn %s(self, rhs: %s) -> %s {
        %s
        %s
    }
`, op.method, w.selfType, w.selfType, stmt, w.wrapResult("bits"))
}

// emitRoundIntegralExactDropOp renders RoundIntegralExact(): a value-only
// wrapper over the flags-exposing port entrypoint, default rounding, flags
// dropped (matches decimal<w>BIDRoundIntegralExactPort).
func emitRoundIntegralExactDropOp(b *strings.Builder, op decOp, w widthSpec) {
	stmt := flagsCallStmt(op, []string{w.selfArg("self"), "BIDGO_ROUND_NEAREST_EVEN"}, "bits", "_raw")
	fmt.Fprintf(b, `
    pub fn %s(self) -> %s {
        %s
        %s
    }
`, op.method, w.selfType, stmt, w.wrapResult("bits"))
}

// emitNextTowardOp renders NextToward(target): the directed-rounding target
// is always Decimal128-width (matches every Go width's NextToward), so the
// BID_UINT128 bytes conversion helper from types.rs is used, no unsafe. This
// is a *different* call than w.selfArg: target is always Decimal128-typed
// regardless of the receiver width w, and (for a Decimal64/Decimal32
// receiver) decimal128.rs's private tuple field is not visible here, so the
// public to_le_bytes() accessor is the only route -- see widthSpec.selfArg's
// doc comment for why that is not the same case as self.0.
func emitNextTowardOp(b *strings.Builder, op decOp, w widthSpec) {
	stmt := flagsCallStmt(op, []string{w.selfArg("self"), "super::types::bid_uint128_from_le_bytes(target.to_le_bytes())"}, "bits", "raw")
	fmt.Fprintf(b, `
    pub fn %s(self, target: Decimal128) -> (%s, ExceptionFlags) {
        %s
        (%s, ExceptionFlags::from_bidgo(raw))
    }
`, op.method, w.selfType, stmt, w.wrapResult("bits"))
}

// emitScaleBOp renders ScaleB(exponent) with the supported-platform LP64 int
// domain preserved as i64 through the public Rust API and port call.
func emitScaleBOp(b *strings.Builder, op decOp, w widthSpec) {
	stmt := flagsCallStmt(op, []string{w.selfArg("self"), "exponent", "BIDGO_ROUND_NEAREST_EVEN"}, "bits", "raw")
	fmt.Fprintf(b, `
    pub fn %s(self, exponent: i64) -> (%s, ExceptionFlags) {
        %s
        (%s, ExceptionFlags::from_bidgo(raw))
    }
`, op.method, w.selfType, stmt, w.wrapResult("bits"))
}

// emitSignalingEqOp renders SignalingEqual: composed from GE && LE (Intel has
// no signaling-equal entrypoint), flags OR'd (matches
// decimal<w>BIDSignalingEqualPort exactly).
func emitSignalingEqOp(b *strings.Builder, method string, ge, le decOp, w widthSpec) {
	geStmt := flagsCallStmt(ge, []string{w.selfArg("self"), w.selfArg("rhs")}, "ge", "ge_raw")
	leStmt := flagsCallStmt(le, []string{w.selfArg("self"), w.selfArg("rhs")}, "le", "le_raw")
	fmt.Fprintf(b, `
    pub fn %s(self, rhs: %s) -> (bool, ExceptionFlags) {
        %s
        %s
        (ge != 0 && le != 0, ExceptionFlags::from_bidgo(ge_raw | le_raw))
    }
`, method, w.selfType, geStmt, leStmt)
}

// emitSignalingNotEqOp renders SignalingNotEqual: negation of the sibling
// signaling_eq wrapper method, same flags (matches
// decimal<w>BIDSignalingNotEqualPort, which calls the Equal composition and
// negates).
func emitSignalingNotEqOp(b *strings.Builder, method, eqMethod string, w widthSpec) {
	fmt.Fprintf(b, `
    pub fn %s(self, rhs: %s) -> (bool, ExceptionFlags) {
        let (eq, flags) = self.%s(rhs);
        (!eq, flags)
    }
`, method, w.selfType, eqMethod)
}

// emitToBinary32Op renders ToBinary32(mode). The source decimal's own width
// never appears in the *signature* (only the target f32), but the port call
// argument does need width-w's own selfArg conversion, AND the port's return
// type itself forks by width: Bid64/Bid32ToBinary32 return the f32 BIT
// PATTERN as a u32 tuple element (wrapped via f32::from_bits, matching the Go
// math.Float32frombits), whereas Bid128ToBinary32 (a pfpsf function) returns
// a native f32 directly (the Go decimal128BIDToBinary32Port returns it
// unwrapped) -- so width 128 must NOT re-interpret it through from_bits.
func emitToBinary32Op(b *strings.Builder, op decOp, w widthSpec) {
	stmt := flagsCallStmt(op, []string{w.selfArg("self"), "super::types::to_bidgo_rounding(mode)"}, "bits", "raw")
	result := "f32::from_bits(bits)"
	if w.is128 {
		result = "bits"
	}
	fmt.Fprintf(b, `
    pub fn %s(self, mode: RoundingMode) -> (f32, ExceptionFlags) {
        %s
        (%s, ExceptionFlags::from_bidgo(raw))
    }
`, op.method, stmt, result)
}

// emitToBinary64Op renders ToBinary64(mode). Same per-width return-type fork
// as emitToBinary32Op (Bid128ToBinary64 returns a native f64 directly).
func emitToBinary64Op(b *strings.Builder, op decOp, w widthSpec) {
	stmt := flagsCallStmt(op, []string{w.selfArg("self"), "super::types::to_bidgo_rounding(mode)"}, "bits", "raw")
	result := "f64::from_bits(bits)"
	if w.is128 {
		result = "bits"
	}
	fmt.Fprintf(b, `
    pub fn %s(self, mode: RoundingMode) -> (f64, ExceptionFlags) {
        %s
        (%s, ExceptionFlags::from_bidgo(raw))
    }
`, op.method, stmt, result)
}

// emitToBinary128Op renders ToBinary128(mode): result BID_UINT128 converts to
// the public Binary128 byte representation via the shared helper, no unsafe.
func emitToBinary128Op(b *strings.Builder, op decOp, w widthSpec) {
	stmt := flagsCallStmt(op, []string{w.selfArg("self"), "super::types::to_bidgo_rounding(mode)"}, "bits", "raw")
	fmt.Fprintf(b, `
    pub fn %s(self, mode: RoundingMode) -> (Binary128, ExceptionFlags) {
        %s
        (
            Binary128::from_le_bytes(super::types::bid_uint128_to_le_bytes(bits)),
            ExceptionFlags::from_bidgo(raw),
        )
    }
`, op.method, stmt)
}

// emitToDecimal128Op renders ToDecimal128(): exact widening, no rounding
// mode; result BID_UINT128 converts to Decimal128 via the shared helper.
// Width-independent: the target is always Decimal128 regardless of source
// width (used by both Decimal64.ToDecimal128 and Decimal32.ToDecimal128).
func emitToDecimal128Op(b *strings.Builder, op decOp) {
	fmt.Fprintf(b, `
    pub fn %s(self) -> (Decimal128, ExceptionFlags) {
        let (bits, raw) = crate::generated::%s::%s(self.0);
        (
            Decimal128::from_le_bytes(super::types::bid_uint128_to_le_bytes(bits)),
            ExceptionFlags::from_bidgo(raw),
        )
    }
`, op.method, op.module, op.port)
}

// emitToDecimal64Op renders ToDecimal64(): exact widening from Decimal32, no
// rounding mode (structurally identical to emitToDecimal128Op, but the
// target is a plain u64-wrapped Decimal64 rather than a 16-byte Decimal128,
// so no BID_UINT128 byte-array conversion is needed). Decimal64 has no
// counterpart shape (it cannot widen to itself); only Decimal32 emits this.
func emitToDecimal64Op(b *strings.Builder, op decOp) {
	fmt.Fprintf(b, `
    pub fn %s(self) -> (Decimal64, ExceptionFlags) {
        let (bits, raw) = crate::generated::%s::%s(self.0);
        (Decimal64::from_bits(bits), ExceptionFlags::from_bidgo(raw))
    }
`, op.method, op.module, op.port)
}

// emitToDecimal32Op renders ToDecimal32(mode): narrowing conversion.
// Decimal64 and Decimal128 both narrow to Decimal32 with an explicit
// rounding mode (Decimal32 itself has no counterpart -- it cannot narrow to
// itself); the target is always Decimal32 regardless of the receiver width w
// (only the port call argument needs w's own selfArg conversion).
func emitToDecimal32Op(b *strings.Builder, op decOp, w widthSpec) {
	stmt := flagsCallStmt(op, []string{w.selfArg("self"), "super::types::to_bidgo_rounding(mode)"}, "bits", "raw")
	fmt.Fprintf(b, `
    pub fn %s(self, mode: RoundingMode) -> (Decimal32, ExceptionFlags) {
        %s
        (Decimal32::from_bits(bits), ExceptionFlags::from_bidgo(raw))
    }
`, op.method, stmt)
}

// emitToDecimal64ModeOp renders Decimal128's ToDecimal64(mode): unlike
// Decimal32's parameterless exact ToDecimal64 (emitToDecimal64Op -- BID32's 7
// digits always fit BID64's 16 exactly), Decimal128 narrowing to Decimal64
// is inexact in general, so the Go signature takes an explicit RoundingMode
// (types_bid_methods.go: `func (d Decimal128BID) ToDecimal64(mode
// RoundingMode) (Decimal64BID, ExceptionFlags)`), giving this its own shape
// ("to_decimal64_mode") structurally identical to emitToDecimal32Op's, only
// the target type differs. Decimal128-receiver only.
func emitToDecimal64ModeOp(b *strings.Builder, op decOp, w widthSpec) {
	stmt := flagsCallStmt(op, []string{w.selfArg("self"), "super::types::to_bidgo_rounding(mode)"}, "bits", "raw")
	fmt.Fprintf(b, `
    pub fn %s(self, mode: RoundingMode) -> (Decimal64, ExceptionFlags) {
        %s
        (Decimal64::from_bits(bits), ExceptionFlags::from_bidgo(raw))
    }
`, op.method, stmt)
}

// emitConvOps renders the 16 ConvertToInt<N>/ConvertToUint<N>(Exact)
// mode-dispatch wrappers: the closed RoundingMode enum selects one of 5 port
// functions via an exhaustive match (no invalid-mode arm is reachable, unlike
// the Go open-int switch's default case). The target integer type and dispatch pattern never reference
// the decimal receiver's own width, but the port call argument does need
// width-w's own selfArg conversion; every one of the 5 dispatch leaves is
// tuple-return for every width including 128 (verified against the actual
// generated:: signatures -- ConvertToInt<N> never uses the pfpsf convention),
// so this needs no pfpsf branch, unlike the flags-carrying shapes above.
func emitConvOps(b *strings.Builder, ops []decConvOp, w widthSpec) error {
	sort.Slice(ops, func(i, j int) bool { return ops[i].method < ops[j].method })
	for _, op := range ops {
		variants, err := intConvVariants(op.canonical, op.exact)
		if err != nil {
			return err
		}
		var arms [5]struct{ module, fn string }
		for i, name := range variants {
			p, ok := portPath[name]
			if !ok {
				return fmt.Errorf("apiemit: no port path for derived int-conversion bidgo_function %q (from canonical %q)", name, op.canonical)
			}
			arms[i] = struct{ module, fn string }{p.module, p.fn}
		}
		selfArg := w.selfArg("self")
		fmt.Fprintf(b, `
    pub fn %s(self, mode: RoundingMode) -> (%s, ExceptionFlags) {
        let (result, raw) = match mode {
            RoundingMode::NearestEven => crate::generated::%s::%s(%s),
            RoundingMode::NearestAway => crate::generated::%s::%s(%s),
            RoundingMode::TowardZero => crate::generated::%s::%s(%s),
            RoundingMode::TowardPositive => crate::generated::%s::%s(%s),
            RoundingMode::TowardNegative => crate::generated::%s::%s(%s),
        };
        (result, ExceptionFlags::from_bidgo(raw))
    }
`, op.method, op.rustType,
			arms[0].module, arms[0].fn, selfArg,
			arms[1].module, arms[1].fn, selfArg,
			arms[2].module, arms[2].fn, selfArg,
			arms[3].module, arms[3].fn, selfArg,
			arms[4].module, arms[4].fn, selfArg)
	}
	return nil
}

// intConvVariants derives the 5 RoundingMode-dispatch bidgo function names
// (nearest-even, nearest-away, toward-zero, toward-positive, toward-negative)
// from the manifest's canonical (nearest-even-family) bidgo_function name,
// using the fixed Intel/GDA suffix convention the census reason text
// documents (Rnint/Rninta/Int/Ceil/Floor, or Xrnint/Xrninta/Xint/Xceil/Xfloor
// for the inexact-signaling "Exact" family). bid754-go's
// decimal64BIDConvertToInt8Port (etc., and its Decimal32/Decimal128
// counterparts) already encode this exact dispatch table in their switch
// statements; this derives the sibling *names* for portPath lookup, which is
// not a second implementation of the dispatch, just repeating its naming
// pattern so every leaf still resolves through the hand-extended portPath
// table (strict if a name does not exist there). Width-independent: the
// canonical name already carries its own "Bid32"/"Bid64" prefix.
func intConvVariants(canonical string, exact bool) ([5]string, error) {
	var variants [5]string
	suffix := "Rnint"
	if exact {
		suffix = "Xrnint"
	}
	if !strings.HasSuffix(canonical, suffix) {
		return variants, fmt.Errorf("apiemit: bidgo_function %q does not end with the expected suffix %q for an int-conversion shape", canonical, suffix)
	}
	base := strings.TrimSuffix(canonical, suffix)
	if exact {
		variants = [5]string{base + "Xrnint", base + "Xrninta", base + "Xint", base + "Xceil", base + "Xfloor"}
	} else {
		variants = [5]string{base + "Rnint", base + "Rninta", base + "Int", base + "Ceil", base + "Floor"}
	}
	return variants, nil
}

// emitFromIntExactOrErrorOp renders NewDecimal<w>FromInt as a direct port
// call. Decimal32/64 integer ports return flags and reject Inexact through
// Result; Decimal128 represents every int64 exactly and its port returns only
// the value. paramType mirrors the Go signature (i32 for Decimal32, i64 for
// Decimal64/128).
func emitFromIntExactOrErrorOp(b *strings.Builder, method string, op decOp, w widthSpec, paramType string, portHasFlags bool) {
	if portHasFlags {
		fmt.Fprintf(b, `
    /// Converts an integer exactly into a %s. Returns an error when the
    /// integer is not exactly representable; use the explicit-mode integer
    /// constructor when a rounded result and flags are desired.
    pub fn %s(i: %s) -> Result<%s, InexactIntegerError> {
        let (bits, raw) = crate::generated::%s::%s(i, BIDGO_ROUND_NEAREST_EVEN);
        let flags = ExceptionFlags::from_bidgo(raw);
        if flags.contains(ExceptionFlags::INEXACT) {
            return Err(InexactIntegerError::new(i.to_string(), %q));
        }
        Ok(%s)
    }
`, w.selfType, method, paramType, w.selfType, op.module, op.port, w.selfType, w.wrapResult("bits"))
		return
	}

	fmt.Fprintf(b, `
    /// Converts an integer exactly into a %s. Every value of the source
    /// integer type is representable, so the error branch is unreachable.
    pub fn %s(i: %s) -> Result<%s, InexactIntegerError> {
        let bits = crate::generated::%s::%s(i);
        Ok(%s)
    }
`, w.selfType, method, paramType, w.selfType, op.module, op.port, w.wrapResult("bits"))
}

// emitFromIntModeOp renders NewDecimal<w>FromInt32/FromUint32/FromInt64/
// FromUint64(x, mode) (Decimal<w>BID, ExceptionFlags). It covers Decimal32's
// from_i32_mode/from_u32_mode (BID32 cannot represent every int32/uint32
// exactly, so Go gave those a RoundingMode + ExceptionFlags signature
// instead of Decimal64's exact impl From<i32>/From<u32>). paramType is the Go
// parameter's own integer type ("i32"/"u32"/"i64"/"u64").
func emitFromIntModeOp(b *strings.Builder, op decOp, w widthSpec, paramType string) {
	fmt.Fprintf(b, `
    pub fn %s(x: %s, mode: RoundingMode) -> (%s, ExceptionFlags) {
        let (bits, raw) = crate::generated::%s::%s(x, super::types::to_bidgo_rounding(mode));
        (%s(bits), ExceptionFlags::from_bidgo(raw))
    }
`, op.method, paramType, w.selfType, op.module, op.port, w.selfType)
}

// emitParseModeOp renders NewDecimal<w>WithMode: the parse_raw body with the
// caller's RoundingMode carried into the from-string port call (via
// super::types::to_bidgo_rounding) folded through parse_with_flags' Ok/Err
// decision (error iff the input string is rejected; flags are meaningless on
// Err and are not returned). Rust's closed RoundingMode enum makes the Go
// invalid-mode flag channel structurally unreachable here, so no rejection
// arm exists.
func emitParseModeOp(b *strings.Builder, method string, op decOp, w widthSpec) {
	fmt.Fprintf(b, `
    /// Parses a decimal string literal rounding excess precision with an
    /// explicit mode, returning the value and the exception flags raised
    /// while parsing on success. Mirrors the Go NewDecimal%sWithMode Ok/Err
    /// decision (error iff the input string is rejected, including a cohort
    /// the port would silently coerce with zero status); on Err the flags are
    /// meaningless (matching the Go signature's discarded-on-error contract),
    /// so they are not returned at all.
    pub fn %s(s: &str, mode: RoundingMode) -> Result<(%s, ExceptionFlags), ParseDecimalError> {
        let (value, flags) = if let Some(value) = parse_decimal%s_nan(s) {
            (value, ExceptionFlags::empty())
        } else if super::types::parse_bid_nan_literal(s).is_some() {
            (%s, ExceptionFlags::INVALID_OPERATION)
        } else {
            let (bits, raw) = crate::generated::%s::%s(s, super::types::to_bidgo_rounding(mode));
            if raw == 0 && bid_finite_literal_cohort_unrepresentable(s, %s, %s, %s) {
                (%s, ExceptionFlags::INVALID_OPERATION)
            } else {
                (%s, ExceptionFlags::from_bidgo(raw))
            }
        };
        if rejected_bid_string_input(s, result_is_nan(value.0), flags) {
            return Err(ParseDecimalError::new(s));
        }
        Ok((value, flags))
    }
`, w.digits, method, w.selfType, w.digits, w.canonicalQNaNResult(), op.module, op.port, w.minQuantum, w.maxQuantum, w.precision, w.canonicalQNaNResult(), w.wrapResult("bits"))
}

// emitParseWithFlagsOp renders NewDecimal<w>WithFlags: Result<(Self,
// ExceptionFlags), ParseDecimalError> folds Go's (D, flags, error) triple;
// flags are meaningless/zero on Err, matching the Go signature's
// discarded-on-error contract.
func emitParseWithFlagsOp(b *strings.Builder, method string, w widthSpec) {
	fmt.Fprintf(b, `
    /// Parses a decimal string literal, returning the value and the
    /// exception flags raised while parsing on success. Mirrors the Go
    /// NewDecimal%sWithFlags Ok/Err decision, including rejection of a cohort
    /// the port would silently coerce with zero status; on Err the flags are
    /// meaningless (matching the Go signature's discarded-on-error contract),
    /// so they are not returned at all.
    pub fn %s(s: &str) -> Result<(%s, ExceptionFlags), ParseDecimalError> {
        let (value, flags) = %s::parse_raw(s);
        if rejected_bid_string_input(s, result_is_nan(value.0), flags) {
            return Err(ParseDecimalError::new(s));
        }
        Ok((value, flags))
    }
`, w.digits, method, w.selfType, w.selfType)
}

// emitFromI32ExactImpl renders `impl From<i32> for Decimal<w>`. Decimal64
// and Decimal128 only: BID32's 7 significant digits cannot represent every
// int32 exactly, so Decimal32 has no such impl (see emitFromIntModeOp /
// from_i32_mode).
func emitFromI32ExactImpl(b *strings.Builder, op decOp, w widthSpec) {
	call := fmt.Sprintf("crate::generated::%s::%s(x)", op.module, op.port)
	fmt.Fprintf(b, `
impl From<i32> for %s {
    fn from(x: i32) -> %s {
        %s
    }
}
`, w.selfType, w.selfType, w.wrapResult(call))
}

// emitFromU32ExactImpl renders `impl From<u32> for Decimal<w>`. Decimal64
// and Decimal128 only; see emitFromI32ExactImpl.
func emitFromU32ExactImpl(b *strings.Builder, op decOp, w widthSpec) {
	call := fmt.Sprintf("crate::generated::%s::%s(x)", op.module, op.port)
	fmt.Fprintf(b, `
impl From<u32> for %s {
    fn from(x: u32) -> %s {
        %s
    }
}
`, w.selfType, w.selfType, w.wrapResult(call))
}

// emitFromI64ExactImpl renders `impl From<i64> for Decimal128`. Decimal128
// only: its 34 significant digits represent every int64 exactly, unlike
// Decimal64/Decimal32 (see emitFromIntModeOp / from_i64_mode, which those two
// widths use instead because they need a RoundingMode + ExceptionFlags
// signature).
func emitFromI64ExactImpl(b *strings.Builder, op decOp, w widthSpec) {
	call := fmt.Sprintf("crate::generated::%s::%s(x)", op.module, op.port)
	fmt.Fprintf(b, `
impl From<i64> for %s {
    fn from(x: i64) -> %s {
        %s
    }
}
`, w.selfType, w.selfType, w.wrapResult(call))
}

// emitFromU64ExactImpl renders `impl From<u64> for Decimal128`. Decimal128
// only; see emitFromI64ExactImpl.
func emitFromU64ExactImpl(b *strings.Builder, op decOp, w widthSpec) {
	call := fmt.Sprintf("crate::generated::%s::%s(x)", op.module, op.port)
	fmt.Fprintf(b, `
impl From<u64> for %s {
    fn from(x: u64) -> %s {
        %s
    }
}
`, w.selfType, w.selfType, w.wrapResult(call))
}

// emitPartialEqOrd renders the idiomatic quiet-semantics comparison traits,
// delegating to the already-emitted quiet comparison methods (flags dropped,
// matching the design's "trait는 플래그 버림" decision and the std f64
// precedent). eqMethod/ltMethod/gtMethod are the manifest-assigned surface
// names of QuietEqual/QuietLess/QuietGreater. These replace the removed
// #[derive(PartialEq, Eq, Hash)] on Decimal<w>: bit-equality is unsound for
// an IEEE decimal float (it makes NaN == NaN and -0 != +0), so the type is
// deliberately neither Eq nor Hash.
func emitPartialEqOrd(b *strings.Builder, eqMethod, ltMethod, gtMethod string, w widthSpec) {
	// The bit-identity accessor name is the public one for the width: the 32/64
	// value types expose to_bits(), the 128 type exposes to_le_bytes() (it has
	// no to_bits()). Naming the wrong one in the rustdoc would point users at a
	// method that does not exist for Decimal128.
	bitAccessor := "to_bits()"
	if w.is128 {
		bitAccessor = "to_le_bytes()"
	}
	fmt.Fprintf(b, `
impl PartialEq for %s {
    /// IEEE 754 quiet equality (clause 5.11), matching the std `+"`f64`"+` `+"`==`"+`:
    /// a NaN is unequal to everything including itself, and `+"`-0 == +0`"+`. This
    /// is numeric equality, NOT bit-equality; the raised exception flags are
    /// dropped (use `+"`%s`"+` for the flag-carrying form, or compare `+"`%s`"+`
    /// for bit-identity).
    fn eq(&self, other: &%s) -> bool {
        self.%s(*other).0
    }
}

impl PartialOrd for %s {
    /// IEEE 754 quiet ordering, matching the std `+"`f64`"+` `+"`partial_cmp`"+`:
    /// returns `+"`None`"+` when the operands are unordered (either is a NaN). For
    /// a total order over every bit pattern use `+"`total_cmp`"+`. Exception flags
    /// raised by the underlying quiet comparisons are dropped.
    fn partial_cmp(&self, other: &%s) -> Option<core::cmp::Ordering> {
        if self.%s(*other).0 {
            Some(core::cmp::Ordering::Less)
        } else if self.%s(*other).0 {
            Some(core::cmp::Ordering::Greater)
        } else if self.%s(*other).0 {
            Some(core::cmp::Ordering::Equal)
        } else {
            None
        }
    }
}
`, w.selfType, eqMethod, bitAccessor, w.selfType, eqMethod, w.selfType, w.selfType, ltMethod, gtMethod, eqMethod)
}

// Per-width NaN-literal helpers.
//
// These four emitters render the bottom-of-file free functions buildDecimalRs
// calls when the corresponding parse/parse_raw/display shapes are emitted.
// Each mirrors a specific bid754-go/types_bid_nan_payload.go function
// byte-for-byte (mask/limit values taken directly from that Go source, not
// re-derived), parameterized by widthSpec instead of duplicated per width.

// emitResultIsNaNFn renders the per-width NaN predicate helper used by
// parse's rejection check (mirrors the Go bidgo.Bid<w>IsNaN call inside
// newDecimal<w>BIDDirectPort exactly). isNaNOp is the already-resolved (and
// portPath-validated) decOp for this width's IsNaN census row, reused here
// rather than looked up a second time; its boolPort decides direct-return
// vs "!= 0" (Decimal32's IsNaN is bool-native; Decimal64's is not).
func emitResultIsNaNFn(b *strings.Builder, w widthSpec, isNaNOp decOp) {
	bitsType := w.bitsType
	arg := "bits"
	if w.is128 {
		bitsType = "[u8; 16]"
		arg = "super::types::bid_uint128_from_le_bytes(bits)"
	}
	call := fmt.Sprintf("crate::generated::%s::%s(%s)", isNaNOp.module, isNaNOp.port, arg)
	if !isNaNOp.boolPort {
		call += " != 0"
	}
	fmt.Fprintf(b, `
/// Reports whether a raw %s-bit BID pattern is a NaN, via the port predicate.
fn result_is_nan(bits: %s) -> bool {
    %s
}
`, w.digits, bitsType, call)
}

// emitInvalidBidStringInputFn renders the shared parse-rejection helpers.
// They mirror the Go rejectedBIDStringInput/invalidBIDStringInput/
// validBIDFiniteLiteral/parseBIDFiniteLiteral/
// unrepresentableBIDStringFlags helpers exactly: reject invalid-operation
// status on an error-returning parser, require complete finite syntax, detect
// an otherwise-silent cohort coercion, and make an error-only parse fail
// on inexact/range flags.
// The text is identical across every width's file, but stays private per
// wrapper file.
func emitInvalidBidStringInputFn(b *strings.Builder) {
	b.WriteString(`
/// Mirrors the Go public error-returning parse rejection rule: malformed input
/// a width-invalid NaN payload, and a silently coerced written cohort
/// (reported by parse_raw as INVALID_OPERATION) are rejected through
/// ParseDecimalError.
fn rejected_bid_string_input(
    input: &str,
    result_is_nan: bool,
    flags: ExceptionFlags,
) -> bool {
    invalid_bid_string_input(input, result_is_nan)
        || flags.contains(ExceptionFlags::INVALID_OPERATION)
}

/// Reports whether a finite conversion changed numeric precision or range.
/// Error-only parse cannot expose these flags and therefore rejects them.
fn unrepresentable_bid_string_flags(flags: ExceptionFlags) -> bool {
    !(flags
        & (ExceptionFlags::INEXACT | ExceptionFlags::UNDERFLOW | ExceptionFlags::OVERFLOW))
        .is_empty()
}

/// Rejects empty input, incomplete finite syntax, and a NaN result whose input
/// was not itself a NaN literal.
fn invalid_bid_string_input(input: &str, result_is_nan: bool) -> bool {
    if input.trim().is_empty() {
        return true;
    }
    if !result_is_nan {
        return !valid_bid_finite_literal(input);
    }
    super::types::parse_bid_nan_literal(input).is_none()
}

/// Mirrors Go validBIDFiniteLiteral by sharing the complete parser used for
/// silent-cohort detection.
fn valid_bid_finite_literal(input: &str) -> bool {
    parse_bid_finite_literal(input).is_some()
}

struct BidFiniteLiteral {
    quantum: Option<BigInt>,
    coefficient_digits: usize,
}

/// Parses the complete finite literal grammar and returns the written cohort:
/// quantum (explicit exponent minus fractional digit count) and coefficient
/// digit count after leading zeros. None rejects malformed syntax; a literal
/// with quantum None is an exact infinity spelling. BigInt keeps huge exponents
/// and huge fractional counts exact while they cancel.
fn parse_bid_finite_literal(input: &str) -> Option<BidFiniteLiteral> {
    let mut rest = input.trim_start_matches(|c| c == ' ' || c == '\t');
    if rest.starts_with('+') || rest.starts_with('-') {
        rest = &rest[1..];
    }
    if rest.eq_ignore_ascii_case("inf") || rest.eq_ignore_ascii_case("infinity") {
        return Some(BidFiniteLiteral { quantum: None, coefficient_digits: 0 });
    }
    let bytes = rest.as_bytes();
    let mut index = 0usize;
    let mut seen_digit = false;
    let mut seen_point = false;
    let mut fractional_digits = 0usize;
    let mut coefficient_digits = 0usize;
    while index < bytes.len() {
        match bytes[index] {
            b'0'..=b'9' => {
                seen_digit = true;
                if coefficient_digits > 0 || bytes[index] != b'0' {
                    coefficient_digits += 1;
                }
                if seen_point {
                    fractional_digits += 1;
                }
                index += 1;
            }
            b'.' if !seen_point => {
                seen_point = true;
                index += 1;
            }
            _ => break,
        }
    }
    if index == bytes.len() {
        return seen_digit.then(|| BidFiniteLiteral {
            quantum: Some(-BigInt::from(fractional_digits)),
            coefficient_digits,
        });
    }
    if !seen_digit || (bytes[index] != b'e' && bytes[index] != b'E') {
        return None;
    }
    index += 1;
    let mut exponent_negative = false;
    if index < bytes.len() && (bytes[index] == b'+' || bytes[index] == b'-') {
        exponent_negative = bytes[index] == b'-';
        index += 1;
    }
    let exponent_start = index;
    while index < bytes.len() && bytes[index].is_ascii_digit() {
        index += 1;
    }
    if index != bytes.len() || index == exponent_start {
        return None;
    }
    let mut exponent = BigInt::parse_bytes(&bytes[exponent_start..index], 10)?;
    if exponent_negative {
        exponent = -exponent;
    }
    Some(BidFiniteLiteral {
        quantum: Some(exponent - BigInt::from(fractional_digits)),
        coefficient_digits,
    })
}

/// Detects a numeric finite literal whose requested quantum or coefficient
/// cannot be encoded by this width. Callers additionally require a zero raw
/// status word, so explicitly flagged port rounding/range behavior remains
/// unchanged.
fn bid_finite_literal_cohort_unrepresentable(
    input: &str,
    min: i64,
    max: i64,
    precision: usize,
) -> bool {
    let Some(BidFiniteLiteral { quantum: Some(quantum), coefficient_digits }) =
        parse_bid_finite_literal(input)
    else {
        return false;
    };
    coefficient_digits > precision || quantum < BigInt::from(min) || quantum > BigInt::from(max)
}
`)
}

// emitParseDecimalNaNFn renders the per-width NaN-literal fast-path
// constructor (mirrors the Go parseDecimal<w>BIDNaN exactly): sign, then
// nan/qnan/snan case-insensitive, then optional decimal payload digits.
// Packs the sign/signaling/payload bits without any port call. The shared
// types.rs parse_uint_payload helper returns u64 regardless of width, so a
// non-u64 width narrows it via w.castFromU64() before OR-ing into bits
// (mirrors the Go parseDecimal32BIDNaN's explicit uint32(payload) cast).
func emitParseDecimalNaNFn(b *strings.Builder, w widthSpec) {
	fmt.Fprintf(b, `
/// Constructs a %s NaN directly from the public NaN-literal grammar (mirrors
/// the Go parseDecimal%sBIDNaN exactly): sign, then nan/qnan/snan
/// case-insensitive, then optional decimal payload digits. Packs the
/// sign/signaling/payload bits without any port call. Returns None for
/// non-NaN-literal input (including a payload at or above the canonical
/// %s limit); callers distinguish an oversized payload from non-NaN syntax
/// and reject it through their declared error/flag channel.
fn parse_decimal%s_nan(input: &str) -> Option<%s> {
    let lit = super::types::parse_bid_nan_literal(input)?;
    let payload = super::types::parse_uint_payload(&lit.payload, %s)?;
    let mut bits: %s = %s;
    if lit.signaling {
        bits = %s;
    }
    if lit.negative {
        bits |= %s;
    }
    bits |= payload%s;
    Some(%s(bits))
}
`, w.selfType, w.digits, w.canonicalLimitLabel, w.digits, w.selfType,
		w.parseMaxPayload, w.bitsType, w.nanBits, w.snanBits, w.signBit, w.castFromU64(), w.selfType)
}

// emitFormatDecimalNaNFn renders the per-width NaN display formatter
// (mirrors the Go formatDecimal<w>BIDNaN exactly), or None for a non-NaN
// value so the caller falls through to the port's to_string. The extracted
// payload is a bitsType value; a non-u64 width widens it via w.castToU64()
// before calling the shared u64-typed types.rs payload_string helper
// (mirrors the Go formatDecimal32BIDNaN's explicit uint64(payload) cast).
func emitFormatDecimalNaNFn(b *strings.Builder, w widthSpec) {
	fmt.Fprintf(b, `
/// Formats a %s NaN directly from its bits (mirrors the Go
/// formatDecimal%sBIDNaN exactly), or returns None for a non-NaN value so the
/// caller falls through to the port's to_string.
fn format_decimal%s_nan(bits: %s) -> Option<String> {
    if bits & %s != %s {
        return None;
    }
    let payload = bits & %s;
    Some(super::types::format_bid_nan(
        bits & %s != 0,
        bits & %s == %s,
        &super::types::payload_string(payload%s, %s),
    ))
}
`, w.selfType, w.digits, w.digits, w.bitsType,
		w.nanBits, w.nanBits,
		w.payloadMask,
		w.signBit,
		w.snanBits, w.snanBits,
		w.castToU64(), w.formatCanonicalLimit)
}

// Decimal128 NaN-literal helpers.
//
// Decimal128's NaN payload (up to 10^33-1, ~110 bits per the pinned Intel
// encoding) does not fit widthSpec's u64-based bitsType/nanBits/parseMax
// Payload/... fields the way emitParseDecimalNaNFn/emitFormatDecimalNaNFn
// use them for 32/64 (see the widthSpec doc comment), so these two emitters
// are a genuinely different (not width-parameterized) implementation, called
// instead of the shared pair when w.is128 is set. Both mirror
// bid754-go/types_bid_nan_payload.go's parseDecimal128BIDNaN/
// formatDecimal128BIDNaN exactly -- verified bit-for-bit against that Go
// source (zero/negative/signaling/payload-spanning-both-words/max-payload
// cases) using native u128 arithmetic split into the BID_UINT128 [lo, hi]
// word pair (bid_uint128_to_le_bytes/from_le_bytes, the same single
// generation point every other 128-wide byte conversion in this file uses),
// not a bignum type or dependency. The quiet/signaling/sign high-word
// constants are declared explicitly for Decimal128 even though IEEE BID uses
// the same bit positions in Decimal64. This prevents a width-64 generator
// setting from implicitly controlling width-128 output.

// emitParseDecimalNaN128Fn renders parse_decimal128_nan: the Decimal128
// sibling of emitParseDecimalNaNFn.
func emitParseDecimalNaN128Fn(b *strings.Builder) {
	fmt.Fprintf(b, `
/// Constructs a 128-bit NaN directly from the public NaN-literal grammar
/// (mirrors the Go parseDecimal128BIDNaN exactly): sign, then nan/qnan/snan
/// case-insensitive, then optional decimal payload digits. Packs the
/// sign/signaling/payload bits without any port call, via native u128
/// arithmetic split into the BID_UINT128 [lo, hi] word pair (no bignum type
/// needed: the payload is always well under u128::MAX). Returns None for
/// non-NaN-literal input (including a payload at or above the canonical
/// 10^33 limit); callers distinguish an oversized payload from non-NaN syntax
/// and reject it through their declared error/flag channel.
fn parse_decimal128_nan(input: &str) -> Option<Decimal128> {
    let lit = super::types::parse_bid_nan_literal(input)?;
    let payload = super::types::parse_u128_payload(&lit.payload, %s)?;
    let lo = payload as u64;
    let mut hi = (payload >> 64) as u64;
    hi |= %s;
    if lit.signaling {
        hi = (hi & !%s) | %s;
    }
    if lit.negative {
        hi |= %s;
    }
    Some(Decimal128(super::types::bid_uint128_to_le_bytes(
        crate::gen_types::BID_UINT128 { lo, hi },
    )))
}
`, decimal128NaNMaxPayload, decimal128QuietNaNHighBits, decimal128QuietNaNHighBits, decimal128SignalingNaNHighBits, decimal128SignHighBit)
}

// emitFormatDecimalNaN128Fn renders format_decimal128_nan: the Decimal128
// sibling of emitFormatDecimalNaNFn.
func emitFormatDecimalNaN128Fn(b *strings.Builder) {
	fmt.Fprintf(b, `
/// Formats a 128-bit NaN directly from its bytes (mirrors the Go
/// formatDecimal128BIDNaN exactly), or returns None for a non-NaN value so
/// the caller falls through to the port's to_string.
fn format_decimal128_nan(bytes: [u8; 16]) -> Option<String> {
    let v = super::types::bid_uint128_from_le_bytes(bytes);
    let hi = v.hi;
    if hi & %s != %s {
        return None;
    }
    let payload: u128 = ((hi & %s) as u128) << 64 | (v.lo as u128);
    Some(super::types::format_bid_nan(
        hi & %s != 0,
        hi & %s == %s,
        &super::types::payload_string_u128(payload, %s),
    ))
}
`, decimal128QuietNaNHighBits, decimal128QuietNaNHighBits, decimal128NaNPayloadMaskHi,
		decimal128SignHighBit, decimal128SignalingNaNHighBits, decimal128SignalingNaNHighBits, decimal128NaNCanonicalLimit)
}
