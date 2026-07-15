// Package apiemit is the public-API surface emitter half of go2rs.
//
// go2rs' main converter is an AST-level Go->Rust translator over the bidgo
// mechanical-port implementation (bid754-go/internal/bidgo). This subpackage is
// the *routing/plumbing* counterpart: it does not translate or reproduce any
// arithmetic. It emits the idiomatic Rust public API (value types, exception
// flags, and thin wrapper methods) whose bodies only call into the already
// translated crate::generated::* port functions, convert types, and map flags.
//
// It is a subpackage of the single permitted generator (devtools/tools/go2rs),
// not a second generator: one `go run ./tools/go2rs` invocation runs both the
// implementation translation and this API emission, and every output carries
// the go2rs generated marker via devtools/internal/genmarker.
//
// Three strict input origins (mirroring the publicroute census pattern):
//
//  1. the public-API routing census
//     (devtools/generated/testspec/public_api_routing_inventory.json): the closed
//     set of mapped public symbols and the bidgo port function each routes
//     through;
//  2. the Go public signatures parsed from the bid754-go module-root package
//     (AST only, no type-checking), used to validate the emit rules;
//  3. the hand-written Rust surface manifest
//     (devtools/tools/registry/rust_api_manifest.json): which mapped symbols are
//     emitted now (with a shape template) and which are explicitly deferred.
//
// Exhaustive set: every mapped census symbol must be either emitted or listed in
// deferred; a mapped symbol missing from both, listed in both, or an
// emit/deferred entry that matches no mapped symbol, is a generation-time hard
// failure. This is how a widened Go public surface is forced into a Rust
// routing decision instead of being silently dropped.
//
// A separate exhaustive set covers the census's excluded_constant_accessor
// symbols: the 12 ZERO/ONE/PI/E value-returning Go functions, e.g.
// bid754-go/api_v2.go's Zero64BID). They get their own manifest section
// ("constants"/"excluded_constants") and closure check
// (resolveConstantsClosure), and are emitted as Rust associated consts
// (const_bits.go) rather than wrapper methods.
package apiemit

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	gotypes "go/types"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// Paths are resolved relative to the devtools module root (where go2rs runs).
const (
	inventoryRel        = "generated/testspec/public_api_routing_inventory.json"
	manifestRel         = "tools/registry/rust_api_manifest.json"
	goSrcRel            = "../bid754-go"
	apiOutRel           = "../bid754-rs/src/generated/api"
	libRsRel            = "../bid754-rs/src/lib.rs"
	surfaceInventoryRel = "generated/testspec/rust_api_surface_inventory.json"
)

// MarkerTool is the tool label used for the generated marker line on every
// apiemit output. It is passed through devtools/internal/genmarker rather than
// hardcoded so the marker coverage check cannot diverge from it.
const MarkerTool = "go2rs apiemit"

// inventorySymbol is one row of the public-API routing census.
type inventorySymbol struct {
	Symbol        string `json:"symbol"`
	Kind          string `json:"kind"`
	Status        string `json:"status"`
	BidgoFunction string `json:"bidgo_function,omitempty"`
	Reason        string `json:"reason,omitempty"`
}

type inventoryFile struct {
	Total    int               `json:"total"`
	Mapped   int               `json:"mapped"`
	Excluded int               `json:"excluded"`
	Symbols  []inventorySymbol `json:"symbols"`
}

// emitRule is one hand-written manifest emit entry: a mapped Go symbol mapped to
// a Rust surface via a named shape template. It carries no free Rust code.
type emitRule struct {
	GoSymbol      string `json:"go_symbol"`
	RustOwner     string `json:"rust_owner"`
	RustSurface   string `json:"rust_surface"`
	Shape         string `json:"shape"`
	BidgoFunction string `json:"bidgo_function"`
	Reason        string `json:"reason"`
}

type deferredBlock struct {
	Category string          `json:"category"`
	Reason   json.RawMessage `json:"reason"`
	Symbols  []string        `json:"symbols"`
}

// constantRule is one hand-written manifest "constants" entry: a Go
// excluded_constant_accessor census symbol mapped to a Rust associated const
// via a fixed decimal literal. This is a SEPARATE exhaustive set from
// emitRule/deferredBlock above (see resolveConstantsClosure), so constant
// accessors get their own manifest section and closure check instead of
// overloading the wrapper-oriented "emit"/"deferred" sections.
type constantRule struct {
	GoSymbol  string `json:"go_symbol"`
	RustOwner string `json:"rust_owner"`
	RustConst string `json:"rust_const"`
	Literal   string `json:"literal"`
	Reason    string `json:"reason"`
}

type manifestFile struct {
	Version  string        `json:"version"`
	Kind     string        `json:"kind"`
	Emit     []emitRule    `json:"emit"`
	Deferred deferredBlock `json:"deferred"`

	Constants         []constantRule `json:"constants"`
	ExcludedConstants deferredBlock  `json:"excluded_constants"`
}

// goSig is a parsed Go public signature keyed by census symbol id.
type goSig struct {
	Symbol      string
	Kind        string // "func" | "method" | "var"
	Recv        string
	RecvPointer bool
	Name        string
	TypeParams  []string
	Params      []string
	Results     []string
}

// sigForm is a form-level classifier for one Go parameter or result type. It is
// deliberately coarse (form, not full type): the shape templates only need to
// know "string / error / flags / a decimal value type", not the exact width.
type sigForm int

const (
	formString               sigForm = iota // the Go string type
	formError                               // the Go error type
	formFlags                               // the public ExceptionFlags type
	formValue                               // a decimal value type (Decimal{32,64,128}BID)
	formBool                                // the Go bool type
	formInt                                 // the Go int type
	formRoundingMode                        // the public RoundingMode type
	formDecimalClass                        // the public DecimalClass type
	formFloat32                             // the Go float32 type
	formFloat64                             // the Go float64 type
	formBinary128                           // the public Binary128 type
	formInt8                                // the Go int8 type
	formInt16                               // the Go int16 type
	formInt32                               // the Go int32 type
	formInt64                               // the Go int64 type
	formUint8                               // the Go uint8 type
	formUint16                              // the Go uint16 type
	formUint32                              // the Go uint32 type
	formUint64                              // the Go uint64 type
	formArithmeticContextPtr                // the Go *ArithmeticContext type
)

func (f sigForm) matches(typeName string) bool {
	switch f {
	case formString:
		return typeName == "string"
	case formError:
		return typeName == "error"
	case formFlags:
		return typeName == "ExceptionFlags"
	case formValue:
		return apiValueTypes[typeName]
	case formBool:
		return typeName == "bool"
	case formInt:
		return typeName == "int"
	case formRoundingMode:
		return typeName == "RoundingMode"
	case formDecimalClass:
		return typeName == "DecimalClass"
	case formFloat32:
		return typeName == "float32"
	case formFloat64:
		return typeName == "float64"
	case formBinary128:
		return typeName == "Binary128"
	case formInt8:
		return typeName == "int8"
	case formInt16:
		return typeName == "int16"
	case formInt32:
		return typeName == "int32"
	case formInt64:
		return typeName == "int64"
	case formUint8:
		return typeName == "uint8"
	case formUint16:
		return typeName == "uint16"
	case formUint32:
		return typeName == "uint32"
	case formUint64:
		return typeName == "uint64"
	case formArithmeticContextPtr:
		return typeName == "*ArithmeticContext"
	default:
		return false
	}
}

func (f sigForm) String() string {
	switch f {
	case formString:
		return "string"
	case formError:
		return "error"
	case formFlags:
		return "ExceptionFlags"
	case formValue:
		return "a decimal value type (Decimal{32,64,128}BID)"
	case formBool:
		return "bool"
	case formInt:
		return "int"
	case formRoundingMode:
		return "RoundingMode"
	case formDecimalClass:
		return "DecimalClass"
	case formFloat32:
		return "float32"
	case formFloat64:
		return "float64"
	case formBinary128:
		return "Binary128"
	case formInt8:
		return "int8"
	case formInt16:
		return "int16"
	case formInt32:
		return "int32"
	case formInt64:
		return "int64"
	case formUint8:
		return "uint8"
	case formUint16:
		return "uint16"
	case formUint32:
		return "uint32"
	case formUint64:
		return "uint64"
	case formArithmeticContextPtr:
		return "*ArithmeticContext"
	default:
		return "?"
	}
}

// apiValueTypes is the set of Go decimal value-type receivers/operands a shape
// may reference (mirrors testgen's publicValueTypeWidth).
var apiValueTypes = map[string]bool{
	"Decimal32BID":  true,
	"Decimal64BID":  true,
	"Decimal128BID": true,
}

// shapeSig declares the Go signature form a shape template assumes: whether the
// symbol is a value-type method or a free function, and the form of each
// parameter and result. resolveClosure validates every emitted symbol's actual
// Go AST signature against this, so a Go public surface that changes shape
// under a template fails generation instead of emitting a stale wrapper.
type shapeSig struct {
	method  bool // true = value-type method (receiver required); false = free func
	params  []sigForm
	results []sigForm
}

// shapeSigs is the single closed registry of emittable shapes and the Go
// signature form each assumes. A manifest emit rule naming a shape absent here
// fails as an unknown shape.
var shapeSigs = map[string]shapeSig{
	// string constructor: NewDecimal<w>(s) (Decimal<w>BID, error)
	"parse":      {method: false, params: []sigForm{formString}, results: []sigForm{formValue, formError}},
	"parse_fold": {method: false, params: []sigForm{formString}, results: []sigForm{formValue, formError}},
	// raw parser: ParseDecimal<w>BIDRaw(s) (Decimal<w>BID, ExceptionFlags)
	"parse_raw": {method: false, params: []sigForm{formString}, results: []sigForm{formValue, formFlags}},
	// flag-returning string constructor: NewDecimal<w>WithFlags(s) (Decimal<w>BID, ExceptionFlags, error)
	"parse_with_flags": {method: false, params: []sigForm{formString}, results: []sigForm{formValue, formFlags, formError}},
	// String() string
	"display": {method: true, params: nil, results: []sigForm{formString}},
	// binary op: (recv) Op(other Decimal<w>BID) Decimal<w>BID
	"binary": {method: true, params: []sigForm{formValue}, results: []sigForm{formValue}},
	// binary op with flags: (recv) OpWithFlags(other Decimal<w>BID) (Decimal<w>BID, ExceptionFlags)
	"binary_with_flags": {method: true, params: []sigForm{formValue}, results: []sigForm{formValue, formFlags}},
	// binary op with an explicit rounding mode and flags: (recv) OpWithMode(other Decimal<w>BID, mode RoundingMode) (Decimal<w>BID, ExceptionFlags)
	"binary_mode_flags": {method: true, params: []sigForm{formValue, formRoundingMode}, results: []sigForm{formValue, formFlags}},
	// Intel mixed-width arithmetic is exposed as a destination-owner associated
	// function in Rust, so the Go source is a free function with two explicitly
	// typed D/Q operands. The four shape names keep that operand order in the
	// manifest instead of inferring it from a surface-name convention.
	"mixed_binary_mode_flags_dd": {method: false, params: []sigForm{formValue, formValue, formRoundingMode}, results: []sigForm{formValue, formFlags}},
	"mixed_binary_mode_flags_dq": {method: false, params: []sigForm{formValue, formValue, formRoundingMode}, results: []sigForm{formValue, formFlags}},
	"mixed_binary_mode_flags_qd": {method: false, params: []sigForm{formValue, formValue, formRoundingMode}, results: []sigForm{formValue, formFlags}},
	"mixed_binary_mode_flags_qq": {method: false, params: []sigForm{formValue, formValue, formRoundingMode}, results: []sigForm{formValue, formFlags}},
	// unary op with an explicit rounding mode and flags: (recv) SqrtWithMode(mode RoundingMode) (Decimal<w>BID, ExceptionFlags)
	"unary_mode_flags": {method: true, params: []sigForm{formRoundingMode}, results: []sigForm{formValue, formFlags}},
	// ternary op with an explicit rounding mode and flags: (recv) FMAWithMode(mul, add Decimal<w>BID, mode RoundingMode) (Decimal<w>BID, ExceptionFlags)
	"ternary_mode_flags": {method: true, params: []sigForm{formValue, formValue, formRoundingMode}, results: []sigForm{formValue, formFlags}},
	// scaleB with an explicit rounding mode: (recv) ScaleBWithMode(exponent int, mode RoundingMode) (Decimal<w>BID, ExceptionFlags)
	"scaleb_mode": {method: true, params: []sigForm{formInt, formRoundingMode}, results: []sigForm{formValue, formFlags}},
	// explicit-rounding-mode string constructor: NewDecimal<w>WithMode(s, mode) (Decimal<w>BID, ExceptionFlags, error)
	"parse_mode": {method: false, params: []sigForm{formString, formRoundingMode}, results: []sigForm{formValue, formFlags, formError}},

	// Decimal64 family shapes (devtools/tools/registry/rust_api_manifest.json).

	// context arithmetic (free func): Op<w>BIDWithContext(a, b Decimal<w>BID, ctx *ArithmeticContext) Decimal<w>BID
	"context_binary_with_flags": {method: false, params: []sigForm{formValue, formValue, formArithmeticContextPtr}, results: []sigForm{formValue}},

	// unary, no flags, no rounding: (recv) Op() Decimal<w>BID  (Abs, Negate)
	"unary": {method: true, params: nil, results: []sigForm{formValue}},
	// Copy() Decimal<w>BID folds into #[derive(Copy)]; emits no wrapper (mirrors parse_fold)
	"copy_fold": {method: true, params: nil, results: []sigForm{formValue}},
	// CopySign(signSource Decimal<w>BID) Decimal<w>BID -- no rounding mode in the port call
	"copysign": {method: true, params: []sigForm{formValue}, results: []sigForm{formValue}},

	// Class() DecimalClass
	"class": {method: true, params: nil, results: []sigForm{formDecimalClass}},
	// CompareTotal(other)/CompareTotalMag(other) int -- Rust surface is core::cmp::Ordering
	"total_cmp":     {method: true, params: []sigForm{formValue}, results: []sigForm{formInt}},
	"total_cmp_mag": {method: true, params: []sigForm{formValue}, results: []sigForm{formInt}},
	// Radix() int -- Rust surface is an associated const (fixed IEEE 754 radix 10)
	"radix_const": {method: true, params: nil, results: []sigForm{formInt}},
	// Sign() int -- composed from IsZero/IsSigned
	"sign": {method: true, params: nil, results: []sigForm{formInt}},
	// SameQuantum(other) bool -- no flags
	"same_quantum": {method: true, params: []sigForm{formValue}, results: []sigForm{formBool}},

	// 9 IEEE predicates: IsZero/IsNaN/IsInf/IsNormal/IsFinite/IsSubnormal/IsSignaling/IsCanonical/IsSignMinus
	"predicate": {method: true, params: nil, results: []sigForm{formBool}},

	// FMA(mul, add Decimal<w>BID) (Decimal<w>BID, ExceptionFlags)
	"fma": {method: true, params: []sigForm{formValue, formValue}, results: []sigForm{formValue, formFlags}},
	// 2-operand ops with flags, no rounding mode: Fmod/MaxNum/MaxNumMag/MinNum/MinNumMag/Remainder
	"binary_flags_no_round": {method: true, params: []sigForm{formValue}, results: []sigForm{formValue, formFlags}},
	// Quantize(other) Decimal<w>BID -- value-only wrapper dropping the port's flags
	"binary_drop_flags": {method: true, params: []sigForm{formValue}, results: []sigForm{formValue}},

	// 1-operand ops with flags, no rounding mode: LogB/NextMinus/NextPlus/RoundIntegralNearestAway/
	// RoundIntegralNearestEven/RoundIntegralNegative/RoundIntegralPositive/RoundIntegralZero
	"unary_with_flags_no_round": {method: true, params: nil, results: []sigForm{formValue, formFlags}},
	// 1-operand ops with flags, default (nearest-even) rounding: Sqrt, RoundIntegralExactWithFlags
	"unary_with_flags_default_round": {method: true, params: nil, results: []sigForm{formValue, formFlags}},
	// RoundIntegralExact() Decimal<w>BID -- value-only wrapper dropping the port's flags, default rounding
	"unary_mode_drop_flags": {method: true, params: nil, results: []sigForm{formValue}},

	// NextToward(target Decimal128BID) (Decimal<w>BID, ExceptionFlags) -- target is always Decimal128BID
	"next_toward": {method: true, params: []sigForm{formValue}, results: []sigForm{formValue, formFlags}},
	// ScaleB(exponent int) (Decimal<w>BID, ExceptionFlags)
	"scaleb": {method: true, params: []sigForm{formInt}, results: []sigForm{formValue, formFlags}},

	// Quiet*/Signaling* (direct) comparisons: Op(other) (bool, ExceptionFlags)
	"compare_bool_flags": {method: true, params: []sigForm{formValue}, results: []sigForm{formBool, formFlags}},
	// SignalingEqual(other) (bool, ExceptionFlags) -- composed from GE && LE (no Intel signaling-equal entrypoint)
	"signaling_eq_compose": {method: true, params: []sigForm{formValue}, results: []sigForm{formBool, formFlags}},
	// SignalingNotEqual(other) (bool, ExceptionFlags) -- negation of signaling_eq_compose
	"signaling_not_eq_compose": {method: true, params: []sigForm{formValue}, results: []sigForm{formBool, formFlags}},

	// ToBinary32/64/128(mode) (fN/Binary128, ExceptionFlags)
	"to_binary32":  {method: true, params: []sigForm{formRoundingMode}, results: []sigForm{formFloat32, formFlags}},
	"to_binary64":  {method: true, params: []sigForm{formRoundingMode}, results: []sigForm{formFloat64, formFlags}},
	"to_binary128": {method: true, params: []sigForm{formRoundingMode}, results: []sigForm{formBinary128, formFlags}},
	// ToDecimal128() (Decimal128BID, ExceptionFlags) -- exact widening, no rounding mode
	"to_decimal128": {method: true, params: nil, results: []sigForm{formValue, formFlags}},
	// ToDecimal32(mode) (Decimal32BID, ExceptionFlags) -- narrowing
	"to_decimal32": {method: true, params: []sigForm{formRoundingMode}, results: []sigForm{formValue, formFlags}},

	// ConvertToInt<N>(mode)/ConvertToUint<N>(mode) (iN/uN, ExceptionFlags) -- mode dispatches across 5 port fns
	"to_i8":        {method: true, params: []sigForm{formRoundingMode}, results: []sigForm{formInt8, formFlags}},
	"to_i8_exact":  {method: true, params: []sigForm{formRoundingMode}, results: []sigForm{formInt8, formFlags}},
	"to_i16":       {method: true, params: []sigForm{formRoundingMode}, results: []sigForm{formInt16, formFlags}},
	"to_i16_exact": {method: true, params: []sigForm{formRoundingMode}, results: []sigForm{formInt16, formFlags}},
	"to_i32":       {method: true, params: []sigForm{formRoundingMode}, results: []sigForm{formInt32, formFlags}},
	"to_i32_exact": {method: true, params: []sigForm{formRoundingMode}, results: []sigForm{formInt32, formFlags}},
	"to_i64":       {method: true, params: []sigForm{formRoundingMode}, results: []sigForm{formInt64, formFlags}},
	"to_i64_exact": {method: true, params: []sigForm{formRoundingMode}, results: []sigForm{formInt64, formFlags}},
	"to_u8":        {method: true, params: []sigForm{formRoundingMode}, results: []sigForm{formUint8, formFlags}},
	"to_u8_exact":  {method: true, params: []sigForm{formRoundingMode}, results: []sigForm{formUint8, formFlags}},
	"to_u16":       {method: true, params: []sigForm{formRoundingMode}, results: []sigForm{formUint16, formFlags}},
	"to_u16_exact": {method: true, params: []sigForm{formRoundingMode}, results: []sigForm{formUint16, formFlags}},
	"to_u32":       {method: true, params: []sigForm{formRoundingMode}, results: []sigForm{formUint32, formFlags}},
	"to_u32_exact": {method: true, params: []sigForm{formRoundingMode}, results: []sigForm{formUint32, formFlags}},
	"to_u64":       {method: true, params: []sigForm{formRoundingMode}, results: []sigForm{formUint64, formFlags}},
	"to_u64_exact": {method: true, params: []sigForm{formRoundingMode}, results: []sigForm{formUint64, formFlags}},

	// integer constructors (free funcs)
	// NewDecimal64/128FromInt(i int64) (Decimal<w>BID, error) -- exact or explicit error.
	"from_i64_exact_or_error": {method: false, params: []sigForm{formInt64}, results: []sigForm{formValue, formError}},
	// NewDecimal64FromInt32(x int32) Decimal64BID -- exact, folds into impl From<i32>
	"from_i32_exact": {method: false, params: []sigForm{formInt32}, results: []sigForm{formValue}},
	// NewDecimal64FromUint32(x uint32) Decimal64BID -- exact, folds into impl From<u32>
	"from_u32_exact": {method: false, params: []sigForm{formUint32}, results: []sigForm{formValue}},
	// NewDecimal64FromInt64(x int64, mode) (Decimal64BID, ExceptionFlags)
	"from_i64_mode": {method: false, params: []sigForm{formInt64, formRoundingMode}, results: []sigForm{formValue, formFlags}},
	// NewDecimal64FromUint64(x uint64, mode) (Decimal64BID, ExceptionFlags)
	"from_u64_mode": {method: false, params: []sigForm{formUint64, formRoundingMode}, results: []sigForm{formValue, formFlags}},

	// Decimal32 family shapes not already covered by a Decimal64 shape.
	// Every other Decimal32 emit row reuses a Decimal64 shape name; the Go
	// signature *form* -- receiver/method-ness, parameter arity+form, result
	// arity+form -- is identical across widths for those, only the concrete
	// Rust value type differs, which the shape template now takes as a
	// width parameter rather than a hardcoded literal.)

	// NewDecimal32FromInt(i int32) (Decimal32BID, error) -- exact or explicit error.
	"from_i32_exact_or_error": {method: false, params: []sigForm{formInt32}, results: []sigForm{formValue, formError}},
	// NewDecimal32FromInt32(x int32, mode) (Decimal32BID, ExceptionFlags) --
	// unlike Decimal64FromInt32 (exact, no mode: every int32 fits BID64's 16
	// digits), BID32's 7 digits cannot represent every int32 exactly, so Go
	// gave this a RoundingMode + ExceptionFlags signature instead of folding
	// into impl From<i32>.
	"from_i32_mode": {method: false, params: []sigForm{formInt32, formRoundingMode}, results: []sigForm{formValue, formFlags}},
	// NewDecimal32FromUint32(x uint32, mode) (Decimal32BID, ExceptionFlags) -- see from_i32_mode.
	"from_u32_mode": {method: false, params: []sigForm{formUint32, formRoundingMode}, results: []sigForm{formValue, formFlags}},
	// ToDecimal64() (Decimal64BID, ExceptionFlags) -- exact widening from
	// Decimal32, no rounding mode (mirrors to_decimal128's shape form exactly;
	// Decimal64 has no counterpart since it cannot widen to itself).
	"to_decimal64": {method: true, params: nil, results: []sigForm{formValue, formFlags}},

	// Decimal128 family shapes not already covered by a Decimal64/Decimal32 shape.

	// NewDecimal128FromInt64(x int64) Decimal128BID -- exact, folds into impl
	// From<i64>. Decimal128-only: its 34 significant digits represent every
	// int64 exactly, unlike Decimal64/Decimal32 (see from_i64_mode, which
	// those two widths use instead).
	"from_i64_exact": {method: false, params: []sigForm{formInt64}, results: []sigForm{formValue}},
	// NewDecimal128FromUint64(x uint64) Decimal128BID -- exact, folds into
	// impl From<u64>. Decimal128-only; see from_i64_exact.
	"from_u64_exact": {method: false, params: []sigForm{formUint64}, results: []sigForm{formValue}},
	// Decimal128BID.ToDecimal64(mode) (Decimal64BID, ExceptionFlags) --
	// narrowing conversion with an explicit rounding mode: unlike Decimal32's
	// exact, parameterless ToDecimal64 (BID32's 7 digits always fit BID64's
	// 16 exactly), Decimal128 narrowing to Decimal64 is inexact in general,
	// so Go gave this its own RoundingMode-taking signature. Structurally
	// identical in sigForm terms to to_decimal32, only the target type
	// differs (handled by the emitter, not the shape signature).
	"to_decimal64_mode": {method: true, params: []sigForm{formRoundingMode}, results: []sigForm{formValue, formFlags}},
}

// validateShapeSig checks one emitted symbol's actual Go AST signature against
// the form its shape assumes (receiver, parameter arity+form, result
// arity+form), failing closed on any divergence.
func validateShapeSig(goSymbol, shape string, spec shapeSig, sig goSig) error {
	if spec.method {
		if sig.Kind != "method" {
			return fmt.Errorf("apiemit: emit go_symbol %q shape %q expects a value-type method but the Go symbol is a %s", goSymbol, shape, sig.Kind)
		}
		if !apiValueTypes[sig.Recv] {
			return fmt.Errorf("apiemit: emit go_symbol %q shape %q expects a decimal value-type receiver but the Go receiver is %q", goSymbol, shape, sig.Recv)
		}
	} else if sig.Kind != "func" {
		return fmt.Errorf("apiemit: emit go_symbol %q shape %q expects a free function but the Go symbol is a %s", goSymbol, shape, sig.Kind)
	}
	if len(sig.Params) != len(spec.params) {
		return fmt.Errorf("apiemit: emit go_symbol %q shape %q expects %d parameter(s) but the Go signature has %d %v", goSymbol, shape, len(spec.params), len(sig.Params), sig.Params)
	}
	for i, f := range spec.params {
		if !f.matches(sig.Params[i]) {
			return fmt.Errorf("apiemit: emit go_symbol %q shape %q expects parameter %d to be %s but the Go signature has %q", goSymbol, shape, i, f, sig.Params[i])
		}
	}
	if len(sig.Results) != len(spec.results) {
		return fmt.Errorf("apiemit: emit go_symbol %q shape %q expects %d result(s) but the Go signature has %d %v", goSymbol, shape, len(spec.results), len(sig.Results), sig.Results)
	}
	for i, f := range spec.results {
		if !f.matches(sig.Results[i]) {
			return fmt.Errorf("apiemit: emit go_symbol %q shape %q expects result %d to be %s but the Go signature has %q", goSymbol, shape, i, f, sig.Results[i])
		}
	}
	return nil
}

var mixedBinaryShapeOperands = map[string][2]string{
	"mixed_binary_mode_flags_dd": {"Decimal64BID", "Decimal64BID"},
	"mixed_binary_mode_flags_dq": {"Decimal64BID", "Decimal128BID"},
	"mixed_binary_mode_flags_qd": {"Decimal128BID", "Decimal64BID"},
	"mixed_binary_mode_flags_qq": {"Decimal128BID", "Decimal128BID"},
}

func validateMixedBinaryShape(r emitRule, sig goSig) error {
	wantOperands, mixed := mixedBinaryShapeOperands[r.Shape]
	if !mixed {
		return nil
	}
	if sig.Params[0] != wantOperands[0] || sig.Params[1] != wantOperands[1] {
		return fmt.Errorf("apiemit: emit go_symbol %q shape %q requires operands (%s, %s), got (%s, %s)", r.GoSymbol, r.Shape, wantOperands[0], wantOperands[1], sig.Params[0], sig.Params[1])
	}
	wantResult := r.RustOwner + "BID"
	if sig.Results[0] != wantResult {
		return fmt.Errorf("apiemit: emit go_symbol %q shape %q owner %q requires result %s, got %s", r.GoSymbol, r.Shape, r.RustOwner, wantResult, sig.Results[0])
	}
	code := strings.TrimPrefix(r.Shape, "mixed_binary_mode_flags_")
	digits := strings.TrimPrefix(r.RustOwner, "Decimal")
	publicSuffix := digits + strings.ToUpper(code) + "BIDWithMode"
	operation := ""
	for _, candidate := range []string{"Add", "Sub", "Mul", "Div"} {
		if r.GoSymbol == candidate+publicSuffix {
			operation = candidate
			break
		}
	}
	if operation == "" {
		return fmt.Errorf("apiemit: emit go_symbol %q shape %q owner %q must name its mixed operation and operand widths as <Add|Sub|Mul|Div>%s", r.GoSymbol, r.Shape, r.RustOwner, publicSuffix)
	}
	wantPort := "Bid" + digits + code + operation
	if r.BidgoFunction != wantPort {
		return fmt.Errorf("apiemit: emit go_symbol %q shape %q owner %q requires exact port %q, got %q", r.GoSymbol, r.Shape, r.RustOwner, wantPort, r.BidgoFunction)
	}
	return nil
}

// surfaceRow is one row of the emitted rust_api_surface_inventory.json snapshot.
type surfaceRow struct {
	GoSymbol      string `json:"go_symbol"`
	Status        string `json:"status"` // "emitted" | "deferred"
	RustOwner     string `json:"rust_owner,omitempty"`
	RustSurface   string `json:"rust_surface,omitempty"`
	Shape         string `json:"shape,omitempty"`
	BidgoFunction string `json:"bidgo_function,omitempty"`
}

type surfaceInventory struct {
	Version     string       `json:"version"`
	Kind        string       `json:"kind"`
	Source      string       `json:"source"`
	MappedTotal int          `json:"mapped_total"`
	Emitted     int          `json:"emitted"`
	Deferred    int          `json:"deferred"`
	Rows        []surfaceRow `json:"rows"`

	// Constants is the separate constants closure: the 12
	// excluded_constant_accessor census symbols (outside the
	// mapped public-symbol set above), each either emitted as a Rust associated const
	// or excluded with a reason (resolveConstantsClosure). Kept as distinct
	// top-level fields rather than folded into Rows/MappedTotal/Emitted so the
	// two closures' accounting never collide (a reader iterating Rows for
	// "status == emitted" must count wrapper rows separately from constant rows).
	ConstantsTotal    int                  `json:"constants_total"`
	ConstantsEmitted  int                  `json:"constants_emitted"`
	ConstantsExcluded int                  `json:"constants_excluded"`
	Constants         []constantSurfaceRow `json:"constants"`
}

// Run performs the full public-API emission. root is the devtools module root.
// It reports an error on any census/manifest/signature inconsistency before writing
// a single byte, so a half-emitted surface never reaches the tree.
func Run(root string) error {
	inventory, err := loadInventory(filepath.Join(root, inventoryRel))
	if err != nil {
		return err
	}
	manifest, err := loadManifest(filepath.Join(root, manifestRel))
	if err != nil {
		return err
	}
	sigs, err := loadGoPublicSigs(filepath.Join(root, goSrcRel))
	if err != nil {
		return err
	}
	goLiterals, err := loadGoConstantLiterals(filepath.Join(root, goSrcRel))
	if err != nil {
		return err
	}

	rows, err := resolveClosure(inventory, manifest, sigs)
	if err != nil {
		return err
	}
	constRows, err := resolveConstantsClosure(inventory, manifest, goLiterals)
	if err != nil {
		return err
	}

	if err := emitRustAPI(filepath.Join(root, apiOutRel), manifest); err != nil {
		return err
	}
	if err := emitLibRs(filepath.Join(root, libRsRel)); err != nil {
		return err
	}
	if err := emitSurfaceInventory(filepath.Join(root, surfaceInventoryRel), rows, constRows); err != nil {
		return err
	}
	return nil
}

func loadInventory(path string) (*inventoryFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("apiemit: read public-API routing inventory %q: %w", path, err)
	}
	var inventory inventoryFile
	if err := json.Unmarshal(data, &inventory); err != nil {
		return nil, fmt.Errorf("apiemit: parse public-API routing inventory %q: %w", path, err)
	}
	if len(inventory.Symbols) == 0 {
		return nil, fmt.Errorf("apiemit: public-API routing inventory %q has no symbols", path)
	}
	return &inventory, nil
}

func loadManifest(path string) (*manifestFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("apiemit: read rust api manifest %q: %w", path, err)
	}
	var m manifestFile
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("apiemit: parse rust api manifest %q: %w", path, err)
	}
	return &m, nil
}

// loadGoPublicSigs enumerates every exported function, method, and package
// variable in the non-test files of the module-root package bid754 via the Go
// AST (SkipObjectResolution: no type-checking needed for signature shape). Exact
// build-variant duplicates collapse, but a same-named declaration with a
// different kind or signature fails instead of letting file order hide one
// variant. This is the emitter's own parse; extracting a shared census/shape
// helper between go2rs and testgen is intentionally deferred.
func loadGoPublicSigs(dir string) (map[string]goSig, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("apiemit: read Go public source dir %q: %w", dir, err)
	}
	seen := map[string]goSig{}
	origins := map[string]string{}
	fset := token.NewFileSet()
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		path := filepath.Join(dir, name)
		file, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if err != nil {
			return nil, fmt.Errorf("apiemit: parse Go public source %q: %w", path, err)
		}
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				if !d.Name.IsExported() {
					continue
				}
				typeParams := typeParamList(d.Type.TypeParams)
				params := fieldListTypes(d.Type.Params)
				results := fieldListTypes(d.Type.Results)
				if d.Recv == nil {
					sig := goSig{Symbol: d.Name.Name, Kind: "func", Name: d.Name.Name, TypeParams: typeParams, Params: params, Results: results}
					if err := recordGoPublicSig(seen, origins, sig, path); err != nil {
						return nil, err
					}
					continue
				}
				recv, recvPointer := receiverInfo(d.Recv)
				if recv == "" {
					return nil, fmt.Errorf("apiemit: %q: cannot resolve receiver type for method %q", path, d.Name.Name)
				}
				id := recv + "." + d.Name.Name
				sig := goSig{Symbol: id, Kind: "method", Recv: recv, RecvPointer: recvPointer, Name: d.Name.Name, TypeParams: typeParams, Params: params, Results: results}
				if err := recordGoPublicSig(seen, origins, sig, path); err != nil {
					return nil, err
				}
			case *ast.GenDecl:
				if d.Tok != token.VAR {
					continue
				}
				for _, spec := range d.Specs {
					vs, ok := spec.(*ast.ValueSpec)
					if !ok {
						continue
					}
					for _, ident := range vs.Names {
						if ident.IsExported() {
							sig := goSig{Symbol: ident.Name, Kind: "var", Name: ident.Name}
							if err := recordGoPublicSig(seen, origins, sig, path); err != nil {
								return nil, err
							}
						}
					}
				}
			}
		}
	}
	if len(seen) == 0 {
		return nil, fmt.Errorf("apiemit: no exported Go public symbols found under %q", dir)
	}
	return seen, nil
}

func recordGoPublicSig(seen map[string]goSig, origins map[string]string, sig goSig, path string) error {
	if existing, ok := seen[sig.Symbol]; ok {
		if sameGoPublicSig(existing, sig) {
			return nil
		}
		return fmt.Errorf("apiemit: Go public symbol %s has conflicting declarations in %q and %q: %+v vs %+v", sig.Symbol, origins[sig.Symbol], path, existing, sig)
	}
	seen[sig.Symbol] = sig
	origins[sig.Symbol] = path
	return nil
}

func sameGoPublicSig(a, b goSig) bool {
	return a.Symbol == b.Symbol && a.Kind == b.Kind && a.Recv == b.Recv && a.RecvPointer == b.RecvPointer && a.Name == b.Name &&
		sameGoPublicTypeList(a.TypeParams, b.TypeParams) && sameGoPublicTypeList(a.Params, b.Params) && sameGoPublicTypeList(a.Results, b.Results)
}

func sameGoPublicTypeList(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

type goConstantAccessorLiteral struct {
	Literal string
	GoType  string
}

// constructorLiteralFuncs is the closed map from a Go constructor name to its
// exact public value type. loadGoConstantLiterals uses both fields so the
// accessor's Go width cannot silently diverge from the Rust manifest owner.
var constructorLiteralFuncs = map[string]string{
	"NewDecimal32BIDDirect":  "Decimal32BID",
	"NewDecimal64BIDDirect":  "Decimal64BID",
	"NewDecimal128BIDDirect": "Decimal128BID",
}

// loadGoConstantLiterals parses the bid754-go module-root package (AST only,
// mirrors loadGoPublicSigs) in two passes. The first pass extracts the decimal
// literal from each package-private backing value initialized as
// NewDecimal{32,64,128}BIDDirect("..."). The second maps an exported zero-arg
// accessor whose sole statement is `return backingName` to that literal. This
// lets resolveConstantsClosure cross-check rust_api_manifest.json against the
// actual Go accessor source without reintroducing an exported mutable var. A
// backing value or accessor with any other shape is absent from the returned
// map; the exhaustive resolver then fails instead of guessing.
func loadGoConstantLiterals(dir string) (map[string]goConstantAccessorLiteral, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("apiemit: read Go public source dir %q: %w", dir, err)
	}
	var files []*ast.File
	fset := token.NewFileSet()
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		path := filepath.Join(dir, name)
		file, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if err != nil {
			return nil, fmt.Errorf("apiemit: parse Go public source %q: %w", path, err)
		}
		files = append(files, file)
	}

	backingLiterals := map[string]goConstantAccessorLiteral{}
	for _, file := range files {
		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.VAR {
				continue
			}
			for _, spec := range gen.Specs {
				vs, ok := spec.(*ast.ValueSpec)
				if !ok || len(vs.Values) != 1 {
					continue
				}
				call, ok := vs.Values[0].(*ast.CallExpr)
				if !ok || len(call.Args) != 1 {
					continue
				}
				fnIdent, ok := call.Fun.(*ast.Ident)
				if !ok {
					continue
				}
				goType, ok := constructorLiteralFuncs[fnIdent.Name]
				if !ok {
					continue
				}
				lit, ok := call.Args[0].(*ast.BasicLit)
				if !ok || lit.Kind != token.STRING {
					continue
				}
				value, err := strconv.Unquote(lit.Value)
				if err != nil {
					return nil, fmt.Errorf("apiemit: %q: unquote %s literal argument %s: %w", fset.Position(lit.Pos()).Filename, fnIdent.Name, lit.Value, err)
				}
				if len(vs.Names) > 0 && vs.Names[0].Name != "_" {
					name := vs.Names[0].Name
					candidate := goConstantAccessorLiteral{Literal: value, GoType: goType}
					if existing, duplicate := backingLiterals[name]; duplicate && existing != candidate {
						return nil, fmt.Errorf("apiemit: constant backing value %s has conflicting build-variant definitions: %+v vs %+v", name, existing, candidate)
					}
					backingLiterals[name] = candidate
				}
			}
		}
	}

	out := map[string]goConstantAccessorLiteral{}
	for _, file := range files {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv != nil || !fn.Name.IsExported() || fn.Body == nil {
				continue
			}
			if fn.Type.Params != nil && len(fn.Type.Params.List) != 0 {
				continue
			}
			resultTypes := fieldListTypes(fn.Type.Results)
			if len(resultTypes) != 1 {
				continue
			}
			if len(fn.Body.List) != 1 {
				continue
			}
			ret, ok := fn.Body.List[0].(*ast.ReturnStmt)
			if !ok || len(ret.Results) != 1 {
				continue
			}
			backing, ok := ret.Results[0].(*ast.Ident)
			if !ok {
				continue
			}
			if literal, ok := backingLiterals[backing.Name]; ok {
				if fn.Type.TypeParams != nil && len(fn.Type.TypeParams.List) != 0 {
					return nil, fmt.Errorf("apiemit: constant accessor %s must not declare type parameters", fn.Name.Name)
				}
				for _, result := range fn.Type.Results.List {
					if len(result.Names) != 0 {
						return nil, fmt.Errorf("apiemit: constant accessor %s must use an unnamed result so return %s unambiguously refers to the package backing value", fn.Name.Name, backing.Name)
					}
				}
				if resultTypes[0] != literal.GoType {
					return nil, fmt.Errorf("apiemit: constant accessor %s returns %s but backing value %s is initialized as %s", fn.Name.Name, resultTypes[0], backing.Name, literal.GoType)
				}
				if existing, duplicate := out[fn.Name.Name]; duplicate && existing != literal {
					return nil, fmt.Errorf("apiemit: constant accessor %s has conflicting build-variant definitions: %+v vs %+v", fn.Name.Name, existing, literal)
				}
				out[fn.Name.Name] = literal
			}
		}
	}
	return out, nil
}

func receiverInfo(recv *ast.FieldList) (string, bool) {
	if recv == nil || len(recv.List) == 0 {
		return "", false
	}
	expr := recv.List[0].Type
	pointer := false
	if star, ok := expr.(*ast.StarExpr); ok {
		expr = star.X
		pointer = true
	}
	if ident, ok := expr.(*ast.Ident); ok {
		return ident.Name, pointer
	}
	return "", pointer
}

func fieldListTypes(fields *ast.FieldList) []string {
	if fields == nil {
		return nil
	}
	var out []string
	for _, field := range fields.List {
		typeName := exprString(field.Type)
		n := len(field.Names)
		if n == 0 {
			n = 1
		}
		for i := 0; i < n; i++ {
			out = append(out, typeName)
		}
	}
	return out
}

// typeParamList preserves each type parameter's declared name, order, and
// constraint. Constraint-only comparison cannot distinguish reordered
// declarations when later parameter types refer to those names.
func typeParamList(fields *ast.FieldList) []string {
	if fields == nil {
		return nil
	}
	var out []string
	for _, field := range fields.List {
		constraint := exprString(field.Type)
		if len(field.Names) == 0 {
			out = append(out, constraint)
			continue
		}
		for _, name := range field.Names {
			out = append(out, name.Name+" "+constraint)
		}
	}
	return out
}

func exprString(expr ast.Expr) string {
	return gotypes.ExprString(expr)
}

// resolveClosure enforces the exhaustive set over the census "mapped" set and
// returns the diffable surface rows. Order of checks mirrors the publicroute
// census: manifest hygiene, then exact cover of the mapped set.
func resolveClosure(inventory *inventoryFile, manifest *manifestFile, sigs map[string]goSig) ([]surfaceRow, error) {
	// mapped census symbols -> bidgo function.
	mapped := map[string]string{}
	for _, s := range inventory.Symbols {
		if s.Status == "mapped" {
			mapped[s.Symbol] = s.BidgoFunction
		}
	}
	if len(mapped) == 0 {
		return nil, fmt.Errorf("apiemit: census has no mapped symbols")
	}

	// Manifest hygiene: emit rules complete and unique; each emitted symbol is a
	// mapped census symbol whose bidgo_function agrees with the census, and whose
	// Go signature exists.
	emitByGo := map[string]emitRule{}
	for i, r := range manifest.Emit {
		if r.GoSymbol == "" || r.RustOwner == "" || r.RustSurface == "" || r.Shape == "" || r.BidgoFunction == "" || strings.TrimSpace(r.Reason) == "" {
			return nil, fmt.Errorf("apiemit: manifest emit[%d] requires go_symbol, rust_owner, rust_surface, shape, bidgo_function, and reason", i)
		}
		if _, dup := emitByGo[r.GoSymbol]; dup {
			return nil, fmt.Errorf("apiemit: manifest emit lists go_symbol %q twice", r.GoSymbol)
		}
		spec, shapeOK := shapeSigs[r.Shape]
		if !shapeOK {
			return nil, fmt.Errorf("apiemit: manifest emit[%d] go_symbol %q uses unknown shape %q", i, r.GoSymbol, r.Shape)
		}
		wantBidgo, ok := mapped[r.GoSymbol]
		if !ok {
			return nil, fmt.Errorf("apiemit: manifest emit go_symbol %q is not a mapped census symbol", r.GoSymbol)
		}
		if wantBidgo != r.BidgoFunction {
			return nil, fmt.Errorf("apiemit: manifest emit go_symbol %q bidgo_function %q disagrees with census %q", r.GoSymbol, r.BidgoFunction, wantBidgo)
		}
		sig, sigOK := sigs[r.GoSymbol]
		if !sigOK {
			return nil, fmt.Errorf("apiemit: manifest emit go_symbol %q has no Go public signature (stale manifest or renamed symbol)", r.GoSymbol)
		}
		// Verify the emitted symbol's actual Go signature still has the form the
		// shape template assumes (receiver, parameter arity/kind, result
		// arity/kind). This is what forces a manifest update when the Go public
		// surface changes shape underneath a shape template (e.g. Add gaining a
		// rounding parameter): without it, generation would keep emitting a stale
		// wrapper against a changed port-facing signature.
		if err := validateShapeSig(r.GoSymbol, r.Shape, spec, sig); err != nil {
			return nil, err
		}
		if err := validateMixedBinaryShape(r, sig); err != nil {
			return nil, err
		}
		emitByGo[r.GoSymbol] = r
	}

	// Deferred hygiene: category tag and each deferred symbol is a mapped census
	// symbol, listed once, and not also emitted.
	if manifest.Deferred.Category != "deferred" {
		return nil, fmt.Errorf("apiemit: manifest deferred.category must be %q, got %q", "deferred", manifest.Deferred.Category)
	}
	deferredSet := map[string]bool{}
	for _, sym := range manifest.Deferred.Symbols {
		if _, ok := mapped[sym]; !ok {
			return nil, fmt.Errorf("apiemit: manifest deferred symbol %q is not a mapped census symbol", sym)
		}
		if deferredSet[sym] {
			return nil, fmt.Errorf("apiemit: manifest deferred lists symbol %q twice", sym)
		}
		if _, both := emitByGo[sym]; both {
			return nil, fmt.Errorf("apiemit: manifest symbol %q is both emitted and deferred", sym)
		}
		deferredSet[sym] = true
	}

	// Exact cover: every mapped census symbol is emitted xor deferred.
	var uncovered []string
	for sym := range mapped {
		_, emitted := emitByGo[sym]
		_, deferred := deferredSet[sym]
		if !emitted && !deferred {
			uncovered = append(uncovered, sym)
		}
	}
	if len(uncovered) > 0 {
		sort.Strings(uncovered)
		return nil, fmt.Errorf("apiemit: incomplete mapping: %d mapped census symbol(s) are neither emitted nor deferred in %s; add an emit rule or a deferred entry (strict):\n  %s",
			len(uncovered), manifestRel, strings.Join(uncovered, "\n  "))
	}

	// Emit rules referencing a non-existent mapped symbol are already rejected
	// above; deferred entries referencing a non-mapped symbol too. The remaining
	// reverse direction (a mapped symbol missing from both) is the uncovered
	// check. Total accounting is asserted so a census/manifest drift in size
	// fails loudly.
	if len(emitByGo)+len(deferredSet) != len(mapped) {
		return nil, fmt.Errorf("apiemit: closure accounting mismatch: emitted %d + deferred %d != mapped %d", len(emitByGo), len(deferredSet), len(mapped))
	}

	// Build the diffable surface snapshot rows.
	var rows []surfaceRow
	for _, r := range manifest.Emit {
		rows = append(rows, surfaceRow{
			GoSymbol:      r.GoSymbol,
			Status:        "emitted",
			RustOwner:     r.RustOwner,
			RustSurface:   r.RustSurface,
			Shape:         r.Shape,
			BidgoFunction: r.BidgoFunction,
		})
	}
	for sym := range deferredSet {
		rows = append(rows, surfaceRow{
			GoSymbol:      sym,
			Status:        "deferred",
			BidgoFunction: mapped[sym],
		})
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].GoSymbol < rows[j].GoSymbol })
	return rows, nil
}

// constantSurfaceRow is one row of the "constants" section of the emitted
// rust_api_surface_inventory.json snapshot -- the constants-closure counterpart of
// surfaceRow. BitsHex is the independently-computed BID bit pattern
// (constBitsRustLiteral's output), recorded here as an independent reference value: the
// Rust public-API parity gate does not read it back (it re-parses Literal at
// cargo-test time instead -- see the const_bits.go / rust_public_parity_emit.go
// doc comments for why that independence matters).
type constantSurfaceRow struct {
	GoSymbol  string `json:"go_symbol"`
	Status    string `json:"status"` // "emitted" | "excluded"
	RustOwner string `json:"rust_owner,omitempty"`
	RustConst string `json:"rust_const,omitempty"`
	Literal   string `json:"literal,omitempty"`
	BitsHex   string `json:"bits_hex,omitempty"`
}

// resolveConstantsClosure enforces the SEPARATE exhaustive set over the
// census's excluded_constant_accessor symbols: every accessor must be covered by
// exactly one of manifest.Constants (emit) or manifest.ExcludedConstants (a
// reasoned exclusion), never both, never neither. It cross-validates each
// manifest literal against the Go accessor's package-private backing literal
// extracted by loadGoConstantLiterals, then computes the Rust BID bit pattern
// via const_bits.go's independent encoder before writing any file.
func resolveConstantsClosure(inventory *inventoryFile, manifest *manifestFile, goLiterals map[string]goConstantAccessorLiteral) ([]constantSurfaceRow, error) {
	constantSymbols := map[string]bool{}
	for _, s := range inventory.Symbols {
		if s.Status == "excluded_constant_accessor" {
			constantSymbols[s.Symbol] = true
		}
	}
	if len(constantSymbols) == 0 {
		return nil, fmt.Errorf("apiemit: census has no excluded_constant_accessor symbols")
	}

	byGoSymbol := map[string]constantRule{}
	for i, r := range manifest.Constants {
		if r.GoSymbol == "" || r.RustOwner == "" || r.RustConst == "" || r.Literal == "" || strings.TrimSpace(r.Reason) == "" {
			return nil, fmt.Errorf("apiemit: manifest constants[%d] requires go_symbol, rust_owner, rust_const, literal, and reason", i)
		}
		if _, dup := byGoSymbol[r.GoSymbol]; dup {
			return nil, fmt.Errorf("apiemit: manifest constants lists go_symbol %q twice", r.GoSymbol)
		}
		if !constantSymbols[r.GoSymbol] {
			return nil, fmt.Errorf("apiemit: manifest constants go_symbol %q is not an excluded_constant_accessor census symbol", r.GoSymbol)
		}
		width, ok := widthSpecForOwner(r.RustOwner)
		if !ok {
			return nil, fmt.Errorf("apiemit: manifest constants go_symbol %q has unsupported rust_owner %q (Decimal32, Decimal64, Decimal128 are the only owners with a width record today)", r.GoSymbol, r.RustOwner)
		}
		wantLit, ok := goLiterals[r.GoSymbol]
		if !ok {
			return nil, fmt.Errorf("apiemit: manifest constants go_symbol %q has no parsed bid754-go accessor literal (stale manifest, renamed symbol, or accessor no longer directly returns a constructor-backed package value)", r.GoSymbol)
		}
		wantGoType := width.selfType + "BID"
		if wantLit.GoType != wantGoType {
			return nil, fmt.Errorf("apiemit: manifest constants go_symbol %q rust_owner %q requires Go result %s, but the accessor returns %s", r.GoSymbol, r.RustOwner, wantGoType, wantLit.GoType)
		}
		if wantLit.Literal != r.Literal {
			return nil, fmt.Errorf("apiemit: manifest constants go_symbol %q literal %q disagrees with bid754-go source %q", r.GoSymbol, r.Literal, wantLit.Literal)
		}
		byGoSymbol[r.GoSymbol] = r
	}

	if manifest.ExcludedConstants.Category != "" && manifest.ExcludedConstants.Category != "excluded_constants" {
		return nil, fmt.Errorf("apiemit: manifest excluded_constants.category must be %q, got %q", "excluded_constants", manifest.ExcludedConstants.Category)
	}
	excludedSet := map[string]bool{}
	for _, sym := range manifest.ExcludedConstants.Symbols {
		if !constantSymbols[sym] {
			return nil, fmt.Errorf("apiemit: manifest excluded_constants symbol %q is not an excluded_constant_accessor census symbol", sym)
		}
		if excludedSet[sym] {
			return nil, fmt.Errorf("apiemit: manifest excluded_constants lists symbol %q twice", sym)
		}
		if _, both := byGoSymbol[sym]; both {
			return nil, fmt.Errorf("apiemit: manifest symbol %q is both a constants row and excluded_constants", sym)
		}
		excludedSet[sym] = true
	}

	var uncovered []string
	for sym := range constantSymbols {
		_, emitted := byGoSymbol[sym]
		_, excluded := excludedSet[sym]
		if !emitted && !excluded {
			uncovered = append(uncovered, sym)
		}
	}
	if len(uncovered) > 0 {
		sort.Strings(uncovered)
		return nil, fmt.Errorf("apiemit: constants incomplete mapping: %d excluded_constant_accessor census symbol(s) are neither a constants row nor excluded_constants in %s; add a constants row or an excluded_constants entry (strict):\n  %s",
			len(uncovered), manifestRel, strings.Join(uncovered, "\n  "))
	}
	if len(byGoSymbol)+len(excludedSet) != len(constantSymbols) {
		return nil, fmt.Errorf("apiemit: constants closure accounting mismatch: emitted %d + excluded %d != census %d", len(byGoSymbol), len(excludedSet), len(constantSymbols))
	}

	var rows []constantSurfaceRow
	for _, r := range manifest.Constants {
		w, _ := widthSpecForOwner(r.RustOwner) // already validated above
		bitsHex, err := constBitsRustLiteral(r.Literal, w)
		if err != nil {
			return nil, fmt.Errorf("apiemit: manifest constants go_symbol %q: %w", r.GoSymbol, err)
		}
		rows = append(rows, constantSurfaceRow{GoSymbol: r.GoSymbol, Status: "emitted", RustOwner: r.RustOwner, RustConst: r.RustConst, Literal: r.Literal, BitsHex: bitsHex})
	}
	for sym := range excludedSet {
		rows = append(rows, constantSurfaceRow{GoSymbol: sym, Status: "excluded"})
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].GoSymbol < rows[j].GoSymbol })
	return rows, nil
}

func emitSurfaceInventory(path string, rows []surfaceRow, constRows []constantSurfaceRow) error {
	emitted, deferred := 0, 0
	for _, r := range rows {
		switch r.Status {
		case "emitted":
			emitted++
		case "deferred":
			deferred++
		}
	}
	constEmitted, constExcluded := 0, 0
	for _, r := range constRows {
		switch r.Status {
		case "emitted":
			constEmitted++
		case "excluded":
			constExcluded++
		}
	}
	inventory := surfaceInventory{
		Version:           "1.0",
		Kind:              "rust_api_surface_inventory",
		Source:            "devtools/generated/testspec/public_api_routing_inventory.json (mapped + excluded_constant_accessor) + devtools/tools/registry/rust_api_manifest.json + bid754-go module-root signatures/literals",
		MappedTotal:       len(rows),
		Emitted:           emitted,
		Deferred:          deferred,
		Rows:              rows,
		ConstantsTotal:    len(constRows),
		ConstantsEmitted:  constEmitted,
		ConstantsExcluded: constExcluded,
		Constants:         constRows,
	}
	data, err := json.MarshalIndent(inventory, "", "  ")
	if err != nil {
		return fmt.Errorf("apiemit: marshal rust api surface inventory: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("apiemit: mkdir %q: %w", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("apiemit: write rust api surface inventory %q: %w", path, err)
	}
	return nil
}
