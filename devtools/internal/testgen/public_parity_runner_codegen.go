package testgen

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// The public-API parity runner consumes the static routing inventory from
// public_parity_codegen.go. It exercises those mappings on a
// deterministic bit-literal corpus, comparing the public wrapper's result bits
// and exception flags against an independent invocation of the pinned port
// function. A wrong port routing, swapped argument order, or drifted
// flag/rounding conversion in a wrapper diverges in value here.
//
// The flag- and rounding-conversion comparison is emitted with numeric
// literals (mapPortFlagsForParity, publicParityModes) instead of reusing the
// bidgoExceptionFlags/bidgoRoundingMode converters the wrappers themselves use;
// a converter bug that both sides shared would be invisible. The hand-written
// anchor bid754-go/public_flag_rounding_anchor_test.go pins those converters
// numerically outside the generation path.
const (
	publicParityDispatchPath = "../bid754-go/generated_public_parity_dispatch_test.go"
	publicParityCasesPath    = "../bid754-go/generated_public_parity_cases_test.go"
	publicParityHeaderPath   = "third_party/intel_dfp/LIBRARY/src/bid_functions.h"
)

// publicParityCorpusLen is the fixed length of each per-width value corpus.
const publicParityCorpusLen = 24

// Category labels for the per-width corpus classification. The binary and
// ternary operand matrices are defined over these labels and resolved into
// per-width index tables, so every width's matrix provably includes the
// finite/zero/infinity/qNaN/sNaN/noncanonical directions regardless of where
// those bit patterns sit in that width's edge list.
const (
	catZeroPos      = "zero_pos"
	catZeroNeg      = "zero_neg"
	catFinite       = "finite"
	catInfPos       = "inf_pos"
	catInfNeg       = "inf_neg"
	catQNaN         = "qnan"
	catSNaN         = "snan"
	catNoncanonical = "noncanonical"
)

// parityRequiredCats are the categories every width corpus must contain; a
// missing category is backfilled from parityCategoryBackfill and, failing
// that, is a generation-time hard failure.
var parityRequiredCats = []string{
	catZeroPos, catZeroNeg, catFinite, catInfPos, catInfNeg, catQNaN, catSNaN, catNoncanonical,
}

type paritySel struct {
	Cat string
	Nth int
}

// parityLabelPairs is the width-independent binary direction matrix (§7.3):
// finite mixes, signed-zero pairs, finite-vs-infinity both ways, qNaN and
// sNaN both ways, sNaN-vs-qNaN, and noncanonical-vs-finite both ways.
var parityLabelPairs = [][2]paritySel{
	{{catFinite, 0}, {catFinite, 1}},
	{{catFinite, 1}, {catFinite, 0}},
	{{catFinite, 2}, {catFinite, 3}},
	{{catFinite, 4}, {catFinite, 0}},
	{{catZeroPos, 0}, {catZeroNeg, 0}},
	{{catZeroNeg, 0}, {catZeroPos, 0}},
	{{catZeroPos, 0}, {catFinite, 0}},
	{{catFinite, 0}, {catZeroNeg, 0}},
	{{catFinite, 0}, {catInfPos, 0}},
	{{catInfPos, 0}, {catFinite, 0}},
	{{catFinite, 0}, {catInfNeg, 0}},
	{{catInfNeg, 0}, {catFinite, 0}},
	{{catInfPos, 0}, {catInfNeg, 0}},
	{{catFinite, 0}, {catQNaN, 0}},
	{{catQNaN, 0}, {catFinite, 0}},
	{{catQNaN, 0}, {catQNaN, 1}},
	{{catFinite, 0}, {catSNaN, 0}},
	{{catSNaN, 0}, {catFinite, 0}},
	{{catSNaN, 0}, {catQNaN, 0}},
	{{catQNaN, 0}, {catSNaN, 0}},
	{{catNoncanonical, 0}, {catFinite, 0}},
	{{catFinite, 0}, {catNoncanonical, 0}},
	{{catNoncanonical, 0}, {catZeroPos, 0}},
	{{catInfPos, 0}, {catSNaN, 0}},
}

// parityLabelTriples is the width-independent FMA direction matrix: every
// operand position sees a qNaN, an sNaN, and a noncanonical value.
var parityLabelTriples = [][3]paritySel{
	{{catFinite, 0}, {catFinite, 1}, {catFinite, 2}},
	{{catZeroPos, 0}, {catFinite, 0}, {catFinite, 1}},
	{{catInfPos, 0}, {catFinite, 0}, {catFinite, 1}},
	{{catFinite, 0}, {catInfNeg, 0}, {catFinite, 1}},
	{{catQNaN, 0}, {catFinite, 0}, {catFinite, 1}},
	{{catFinite, 0}, {catQNaN, 0}, {catFinite, 1}},
	{{catFinite, 0}, {catFinite, 1}, {catQNaN, 0}},
	{{catSNaN, 0}, {catFinite, 0}, {catFinite, 1}},
	{{catFinite, 0}, {catSNaN, 0}, {catFinite, 1}},
	{{catFinite, 0}, {catFinite, 1}, {catSNaN, 0}},
	{{catNoncanonical, 0}, {catFinite, 0}, {catFinite, 1}},
	{{catFinite, 0}, {catNoncanonical, 0}, {catFinite, 1}},
	{{catFinite, 0}, {catFinite, 1}, {catNoncanonical, 0}},
	{{catFinite, 2}, {catFinite, 3}, {catZeroNeg, 0}},
}

// parityCategoryBackfill holds explicit per-width bit patterns appended to the
// corpus when the FFI edge list lacks a required category (the 64-bit edge
// list has no noncanonical value: its steering-11 entries all stay within the
// 10^16-1 coefficient limit).
var (
	parityBackfill32 = map[string]uint32{
		catSNaN:         0x7e000000,
		catNoncanonical: 0x6cb89680, // steering-11, coefficient 10000000 > 9999999
	}
	parityBackfill64 = map[string]uint64{
		catSNaN:         0x7e00000000000000,
		catNoncanonical: 0x6007ffffffffffff, // steering-11, coefficient 2^53+2^51-1 > 10^16-1
	}
	parityBackfill128 = map[string][2]uint64{ // hi, lo
		catSNaN:         {0x7e00000000000000, 0},
		catNoncanonical: {0x6000000000000000, 0}, // steering-11 is always noncanonical in BID128
	}
)

// classifyParity32/64/128 classify a corpus bit pattern by the BID encoding
// rules of the pinned Intel format (sign/steering/combination bits and the
// per-width coefficient limits).
func classifyParity32(v uint32) string {
	switch {
	case v&0x7e000000 == 0x7e000000:
		return catSNaN
	case v&0x7c000000 == 0x7c000000:
		return catQNaN
	case v&0x78000000 == 0x78000000:
		if v&0x80000000 != 0 {
			return catInfNeg
		}
		return catInfPos
	case v&0x60000000 == 0x60000000:
		sig := uint32(0x00800000) | (v & 0x001fffff)
		if sig > 9999999 {
			return catNoncanonical
		}
		return catFinite
	default:
		if v&0x007fffff == 0 {
			if v&0x80000000 != 0 {
				return catZeroNeg
			}
			return catZeroPos
		}
		return catFinite
	}
}

func classifyParity64(v uint64) string {
	switch {
	case v&0x7e00000000000000 == 0x7e00000000000000:
		return catSNaN
	case v&0x7c00000000000000 == 0x7c00000000000000:
		return catQNaN
	case v&0x7800000000000000 == 0x7800000000000000:
		if v&0x8000000000000000 != 0 {
			return catInfNeg
		}
		return catInfPos
	case v&0x6000000000000000 == 0x6000000000000000:
		sig := uint64(1)<<53 | (v & 0x0007ffffffffffff)
		if sig > 9999999999999999 {
			return catNoncanonical
		}
		return catFinite
	default:
		if v&0x001fffffffffffff == 0 {
			if v&0x8000000000000000 != 0 {
				return catZeroNeg
			}
			return catZeroPos
		}
		return catFinite
	}
}

func classifyParity128(raw [16]byte) string {
	lo := binary.LittleEndian.Uint64(raw[0:8])
	hi := binary.LittleEndian.Uint64(raw[8:16])
	switch {
	case hi&0x7e00000000000000 == 0x7e00000000000000:
		return catSNaN
	case hi&0x7c00000000000000 == 0x7c00000000000000:
		return catQNaN
	case hi&0x7800000000000000 == 0x7800000000000000:
		if hi&0x8000000000000000 != 0 {
			return catInfNeg
		}
		return catInfPos
	case hi&0x6000000000000000 == 0x6000000000000000:
		// BID128 has no steering-11 finite form; these bit patterns are
		// always noncanonical.
		return catNoncanonical
	default:
		coeffHi := hi & 0x0001ffffffffffff
		// 10^34-1 = 0x1ed09bead87c0_378d8e63ffffffff
		const maxCoeffHi = uint64(0x0001ed09bead87c0)
		const maxCoeffLo = uint64(0x378d8e63ffffffff)
		if coeffHi > maxCoeffHi || (coeffHi == maxCoeffHi && lo > maxCoeffLo) {
			return catNoncanonical
		}
		if coeffHi == 0 && lo == 0 {
			if hi&0x8000000000000000 != 0 {
				return catZeroNeg
			}
			return catZeroPos
		}
		return catFinite
	}
}

// publicParityScaleBExps are the LP64 integer exponents exercised for ScaleB.
// The large values prove that every public language path preserves Go's
// supported-platform 64-bit int domain instead of silently narrowing it to
// int32 before the scalbln port entrypoint.
var publicParityScaleBExps = []int{
	0, 1, -1, 5, -5, 90, -90, 369, -369, 6100, -6180,
	1 << 40, -(1 << 40),
}

// publicParityIntCorpus32 / *Uint32 / *Int64 / *Uint64 feed the integer
// constructors, hitting sign edges, precision boundaries, and type limits.
var publicParityIntCorpus32 = []int32{
	0, 1, -1, 2, -2, 9, -9, 10, -10, 99, -99, 9999999, -9999999,
	10000000, -10000000, 123456789, -123456789, 999999999, -999999999,
	2147483647, -2147483648,
}
var publicParityUintCorpus32 = []uint32{
	0, 1, 2, 9, 10, 99, 100, 9999999, 10000000, 99999999, 100000000,
	123456789, 999999999, 1000000000, 2147483647, 2147483648, 4294967295,
}
var publicParityIntCorpus64 = []int64{
	0, 1, -1, 2, -2, 9, -9, 9999999999999999, -9999999999999999,
	10000000000000000, -10000000000000000, 99999999999999995, -99999999999999995,
	1234567890123456789, -1234567890123456789, 9223372036854775807, -9223372036854775808,
}
var publicParityUintCorpus64 = []uint64{
	0, 1, 2, 9, 9999999999999999, 10000000000000000, 99999999999999995,
	1000000000000000000, 9223372036854775807, 9223372036854775808,
	9999999999999999995, 18446744073709551615,
}

// parityStringCase is one string-shape corpus entry with its generation-time
// classification against the documented public parsing contract
// (types_bidgo_runtime.go invalidBIDStringInput + types_bid_nan_payload.go):
//
//   - kind "port": a complete finite/infinity literal. Raw and flag-returning
//     parsers compare bits+flags against the port; error-only constructors
//     additionally reject any inexact/overflow/underflow status.
//   - kind "cohort_coercion": a complete finite literal chosen so a narrower
//     width silently changes its written quantum or coefficient with a zero
//     port status. CohortMinWidth is the first width that can encode both
//     (zero means none). The silent case must use the public invalid-input
//     channel; widths that contain the cohort still compare to the port.
//   - kind "nan_literal": a syntactically valid NaN spelling. NaNMinWidth is
//     the smallest BID width that can preserve its payload (zero means none).
//     A fitting payload is intercepted by the hand-built payload branch; a
//     non-fitting payload errors on Result/error surfaces and returns canonical
//     quiet NaN + InvalidOperation from raw flag-returning parsers.
//   - kind "blank": empty/whitespace input. Error-returning constructors must
//     error; public raw parsers return canonical quiet NaN + InvalidOperation.
//   - kind "invalid_syntax": malformed finite/NaN text. Error-returning
//     constructors reject it; public raw parsers return canonical quiet NaN +
//     InvalidOperation instead of exposing Intel's silent fall-through.
type parityStringCase struct {
	Input          string
	Kind           string // "port" | "cohort_coercion" | "nan_literal" | "blank" | "invalid_syntax"
	Signaling      bool   // nan_literal only: sNaN spelling
	NaNMinWidth    int    // nan_literal only: minimum fitting width; zero means none
	CohortMinWidth int    // cohort_coercion only: minimum fitting width; zero means none
}

var publicParityStringCorpusCases = []parityStringCase{
	{Input: "0", Kind: "port"},
	{Input: "-0", Kind: "port"},
	{Input: "1", Kind: "port"},
	{Input: "1000000.0", Kind: "cohort_coercion", CohortMinWidth: 64},
	{Input: "1.2345678901234567890123456789012345", Kind: "port"},
	{Input: "1e91", Kind: "cohort_coercion", CohortMinWidth: 64},
	{Input: "3.14159265", Kind: "port"},
	{Input: "1e6145", Kind: "port"},
	{Input: "0e-102", Kind: "cohort_coercion", CohortMinWidth: 64},
	{Input: "1e370", Kind: "cohort_coercion", CohortMinWidth: 128},
	{Input: "0e-399", Kind: "cohort_coercion", CohortMinWidth: 128},
	{Input: "9999999999999999999999999999999999", Kind: "port"},
	{Input: "1e6112", Kind: "cohort_coercion"},
	{Input: "0e-6177", Kind: "cohort_coercion"},
	{Input: "1e100", Kind: "port"},
	{Input: "1e-100", Kind: "port"},
	{Input: "12345678901234567890.12345", Kind: "port"},
	{Input: "Infinity", Kind: "port"},
	{Input: "-Infinity", Kind: "port"},
	{Input: "inf", Kind: "port"},
	{Input: "1000000000000000.0", Kind: "cohort_coercion", CohortMinWidth: 128},
	{Input: "1000000000000000000000000000000000.0", Kind: "cohort_coercion"},
	{Input: "-0.1", Kind: "port"},
	{Input: "1e-6177", Kind: "port"},
	// Malformed inputs: error-returning constructors reject complete syntax,
	// while public raw parsers use canonical qNaN + InvalidOperation.
	{Input: "abc", Kind: "invalid_syntax"},
	{Input: "1.2.3", Kind: "invalid_syntax"},
	// NaN whitespace boundaries: only leading ASCII space/tab is part of the
	// public grammar. Rust's str::trim must never broaden this to newlines or
	// trailing whitespace.
	{Input: "\nnan", Kind: "invalid_syntax"},
	{Input: "1e", Kind: "invalid_syntax"},
	{Input: "1e5x", Kind: "invalid_syntax"},
	{Input: "nan(123)", Kind: "invalid_syntax"},
	{Input: "nan ", Kind: "invalid_syntax"},
	// The port maps these digitless numeric spellings to finite signed zero;
	// every public parser rejects them through its language-specific channel.
	{Input: ".", Kind: "invalid_syntax"},
	{Input: "+.", Kind: "invalid_syntax"},
	{Input: "-.", Kind: "invalid_syntax"},
	{Input: ".e1", Kind: "invalid_syntax"},
	// NaN grammar with payload beyond every width limit.
	{Input: "nan9999999999999999999999999999999999", Kind: "nan_literal"},
	{Input: "snan9999999999999999999999999999999999", Kind: "nan_literal", Signaling: true},
	// blank inputs
	{Input: "", Kind: "blank"},
	{Input: "   ", Kind: "blank"},
	// NaN literals within every width's payload limit
	{Input: "nan", Kind: "nan_literal", NaNMinWidth: 32},
	{Input: "-nan", Kind: "nan_literal", NaNMinWidth: 32},
	{Input: "NaN1000000", Kind: "nan_literal", NaNMinWidth: 64},
	{Input: "qnan", Kind: "nan_literal", NaNMinWidth: 32},
	{Input: "SNaN1000000000000000", Kind: "nan_literal", Signaling: true, NaNMinWidth: 128},
	{Input: "snan", Kind: "nan_literal", Signaling: true, NaNMinWidth: 32},
	{Input: " \t-snan", Kind: "nan_literal", Signaling: true, NaNMinWidth: 32},
	{Input: "NaN123", Kind: "nan_literal", NaNMinWidth: 32},
	{Input: "SNaN42", Kind: "nan_literal", Signaling: true, NaNMinWidth: 32},
	{Input: "qnan999999", Kind: "nan_literal", NaNMinWidth: 32},
}

var publicParityFiniteLiteralPattern = regexp.MustCompile(`^[+-]?(?:(\d+)(?:\.(\d*))?|\.(\d+))(?:[eE]([+-]?\d+))?$`)

// publicParityFiniteCohort independently decodes a numeric finite literal's
// written quantum and coefficient digit count. It deliberately does not reuse
// production parsing code or its constants: this is the generation-time oracle
// that checks the corpus metadata before either language runner is emitted.
type parityFiniteCohort struct {
	Quantum           *big.Int
	CoefficientDigits int
}

func publicParityFiniteCohort(input string) (parityFiniteCohort, bool) {
	match := publicParityFiniteLiteralPattern.FindStringSubmatch(strings.TrimLeft(input, " \t"))
	if match == nil {
		return parityFiniteCohort{}, false
	}
	fractionalDigits := len(match[2])
	coefficientText := match[1] + match[2]
	if match[3] != "" {
		fractionalDigits = len(match[3])
		coefficientText = match[3]
	}
	exponent := new(big.Int)
	if match[4] != "" {
		var ok bool
		exponent, ok = exponent.SetString(match[4], 10)
		if !ok {
			return parityFiniteCohort{}, false
		}
	}
	return parityFiniteCohort{
		Quantum:           exponent.Sub(exponent, new(big.Int).SetUint64(uint64(fractionalDigits))),
		CoefficientDigits: len(strings.TrimLeft(coefficientText, "0")),
	}, true
}

// validatePublicParityStringCorpus prevents the expected per-width NaN and
// cohort acceptance tables from being self-asserted by typo. The independent
// decoders recompute the minimum fitting width before either language runner
// is emitted.
func validatePublicParityStringCorpus() error {
	seen := make(map[string]bool, len(publicParityStringCorpusCases))
	widths := []struct {
		bits int
		spec parityWidth
	}{
		{bits: 32, spec: decimal32ParityWidth},
		{bits: 64, spec: decimal64ParityWidth},
		{bits: 128, spec: decimal128ParityWidth},
	}
	for _, sc := range publicParityStringCorpusCases {
		if seen[sc.Input] {
			return fmt.Errorf("public parity string corpus contains duplicate input %q", sc.Input)
		}
		seen[sc.Input] = true
		switch sc.Kind {
		case "port", "blank":
			if sc.NaNMinWidth != 0 || sc.CohortMinWidth != 0 || sc.Signaling {
				return fmt.Errorf("public parity string corpus input %q kind %q carries metadata for another kind", sc.Input, sc.Kind)
			}
		case "invalid_syntax":
			if sc.NaNMinWidth != 0 || sc.CohortMinWidth != 0 || sc.Signaling {
				return fmt.Errorf("public parity string corpus input %q kind %q carries metadata for another kind", sc.Input, sc.Kind)
			}
			for _, width := range widths {
				if _, ok := parityNaNExpectation(sc.Input, width.spec); ok {
					return fmt.Errorf("public parity input %q is classified invalid_syntax but independent NaN grammar accepts it at width %d", sc.Input, width.bits)
				}
			}
		case "nan_literal":
			if sc.CohortMinWidth != 0 {
				return fmt.Errorf("public parity NaN input %q carries cohort-only metadata", sc.Input)
			}
			actualMin := 0
			for _, width := range widths {
				expectation, ok := parityNaNExpectation(sc.Input, width.spec)
				if !ok {
					continue
				}
				actualMin = width.bits
				if expectation.Signaling != sc.Signaling {
					return fmt.Errorf("public parity NaN input %q signaling=%v, independent grammar says %v", sc.Input, sc.Signaling, expectation.Signaling)
				}
				break
			}
			if actualMin != sc.NaNMinWidth {
				return fmt.Errorf("public parity NaN input %q minimum width=%d, independent range decoder says %d", sc.Input, sc.NaNMinWidth, actualMin)
			}
		case "cohort_coercion":
			if sc.NaNMinWidth != 0 || sc.Signaling {
				return fmt.Errorf("public parity cohort input %q carries NaN-only metadata", sc.Input)
			}
			cohort, ok := publicParityFiniteCohort(sc.Input)
			if !ok {
				return fmt.Errorf("public parity cohort input %q is not accepted by the independent finite grammar", sc.Input)
			}
			actualMin := 0
			for _, width := range widths {
				if cohort.Quantum.Cmp(big.NewInt(width.spec.minQuantum)) >= 0 &&
					cohort.Quantum.Cmp(big.NewInt(width.spec.maxQuantum)) <= 0 &&
					cohort.CoefficientDigits <= width.spec.precision {
					actualMin = width.bits
					break
				}
			}
			if actualMin != sc.CohortMinWidth {
				return fmt.Errorf("public parity cohort input %q minimum width=%d, independent decoder says %d", sc.Input, sc.CohortMinWidth, actualMin)
			}
		default:
			return fmt.Errorf("public parity string corpus input %q has unsupported kind %q", sc.Input, sc.Kind)
		}
	}
	return nil
}

// ---- mode-discriminant operand corpus (shared by the Go and Rust mode-shape
// emitters) ----
//
// The shared parityLabelPairs corpus probes NaN/infinity/sign/noncanonical
// ROUTING directions; most of its combinations are exact under every rounding
// mode (a NaN or infinity operand never rounds, noncanonical operands decode
// as zero, and the small-integer finite entries multiply/divide exactly), so
// on its own it cannot prove that a mode-taking wrapper actually forwards its
// RoundingMode argument — a wrapper hardcoding nearest-even would match the
// port at every mode for such pairs. Measured before this table existed, the
// shared pairs held zero mode-discriminating Mul/Div combinations at widths
// 32 and 64.
//
// This table supplies, per operation and width, operand tuples whose exact
// (unrounded) result does not fit the width's coefficient precision. Any such
// inexact result sits strictly between two representable neighbors, so the
// five IEEE rounding directions cannot all pick the same neighbor, and a
// correct wrapper MUST produce at least two distinct results across the mode
// cycle. The generated runners assert exactly that per tuple, which keeps the
// gate non-vacuous by construction: a mode-dropping wrapper and a future
// corpus edit that stops discriminating both fail the gate.
//
// Constructions per operation (last digit of every exact result is nonzero,
// so inexactness is arithmetic fact, not measurement):
//   - Add: a half-ULP tie (1 + 5E-p splits NearestEven from NearestAway and
//     the directed modes), sub-ULP increments (1 + 1E-p / 9E-p split the
//     directed modes from nearest), and a negative-operand mirror (splits
//     TowardNegative);
//   - Sub: just-below-one differences (1 - 1E-(p+1)), the tie at the
//     precision boundary (1 - 5E-(p+1)), a two-ULP variant, and a negative
//     mirror;
//   - Mul: products exactly one digit past the precision with a nonzero
//     dropped digit;
//   - Div: nonterminating quotients (thirds and sevenths).
//
// modeDiscOperand is one operand of a mode-discriminating parity case, given
// as decimal components: value = (-1)^Neg * coeff * 10^Exp, with the
// coefficient split into 64-bit halves so width-128 cases could carry
// >19-digit coefficients. Operands are encoded into the width's BID
// small-coefficient form at generation time (encodeModeDiscOperand32/64/128),
// which reports an error on any component outside the small form's range instead
// of emitting bits that would silently decode as a different value.
type modeDiscOperand struct {
	Neg     bool
	CoeffHi uint64 // coefficient bits 64..112 (width 128 only; 0 for 32/64)
	CoeffLo uint64 // coefficient bits 0..63
	Exp     int    // decimal exponent
}

func mdo(coeff uint64, exp int) modeDiscOperand {
	return modeDiscOperand{CoeffLo: coeff, Exp: exp}
}

func mdoNeg(coeff uint64, exp int) modeDiscOperand {
	return modeDiscOperand{Neg: true, CoeffLo: coeff, Exp: exp}
}

// modeBinaryDiscriminants is the discriminant corpus for the binary
// mode-shape wrappers ({Add,Sub,Mul,Div}WithMode), keyed by the operation
// name (the public method name with the WithMode suffix stripped) and width.
var modeBinaryDiscriminants = map[string]map[int][][2]modeDiscOperand{
	"Add": {
		32: {
			{mdo(1, 0), mdo(5, -7)},       // 1 + 5e-7: exact half-ULP tie (NE vs NA and the directed modes)
			{mdo(1, 0), mdo(1, -7)},       // 0.1 ULP above 1 (TowardPositive splits from nearest)
			{mdo(1, 0), mdo(9, -7)},       // 0.9 ULP above 1 (TowardZero/TowardNegative split from nearest)
			{mdoNeg(1, 0), mdoNeg(1, -7)}, // negative mirror (TowardNegative splits)
		},
		64: {
			{mdo(1, 0), mdo(5, -16)},
			{mdo(1, 0), mdo(1, -16)},
			{mdo(1, 0), mdo(9, -16)},
			{mdoNeg(1, 0), mdoNeg(1, -16)},
		},
		128: {
			{mdo(1, 0), mdo(5, -34)},
			{mdo(1, 0), mdo(1, -34)},
			{mdo(1, 0), mdo(9, -34)},
			{mdoNeg(1, 0), mdoNeg(1, -34)},
		},
	},
	"Sub": {
		32: {
			{mdo(1, 0), mdo(1, -8)},    // 1 - 1e-8 = 0.99999999: rounds to 1 except toward zero/negative
			{mdo(1, 0), mdo(5, -8)},    // tie between 0.9999999 and 1
			{mdo(2, 0), mdo(1, -7)},    // 1.9999999
			{mdoNeg(1, 0), mdo(1, -8)}, // -1.00000001 (TowardNegative splits)
		},
		64: {
			{mdo(1, 0), mdo(1, -17)},
			{mdo(1, 0), mdo(5, -17)},
			{mdo(2, 0), mdo(1, -16)},
			{mdoNeg(1, 0), mdo(1, -17)},
		},
		128: {
			{mdo(1, 0), mdo(1, -35)},
			{mdo(1, 0), mdo(5, -35)},
			{mdo(2, 0), mdo(1, -34)},
			{mdoNeg(1, 0), mdo(1, -35)},
		},
	},
	"Mul": {
		32: {
			{mdo(3333334, 0), mdo(3, 0)}, // 10000002: 8 digits, dropped digit 2
			{mdo(2222223, 0), mdo(9, 0)}, // 20000007
			{mdo(1234567, 0), mdo(9, 0)}, // 11111103
			{mdo(3333333, 0), mdo(7, 0)}, // 23333331
		},
		64: {
			{mdo(3333333333333334, 0), mdo(3, 0)},  // 10000000000000002: 17 digits
			{mdo(2222222222222223, 0), mdo(9, 0)},  // 20000000000000007
			{mdo(3333333333333333, 0), mdo(7, 0)},  // 23333333333333331
			{mdo(1111111111111111, 0), mdo(77, 0)}, // 85555555555555547
		},
		128: {
			{mdo(999999999999999999, 0), mdo(99999999999999999, 0)},    // (1e18-1)(1e17-1): 35 digits, ends in 1
			{mdo(9999999999999999999, 0), mdo(9999999999999999999, 0)}, // (1e19-1)^2: 38 digits, ends in 1
			{mdo(3333333333333333333, 0), mdo(3333333333333333333, 0)}, // 38 digits, ends in 9
			{mdo(1234567890123456789, 0), mdo(9876543210987654321, 0)}, // 38 digits, ends in 9
		},
	},
	"Div": {
		32: {
			{mdo(1, 0), mdo(3, 0)}, // 0.333... nonterminating
			{mdo(2, 0), mdo(3, 0)}, // 0.666...
			{mdo(1, 0), mdo(7, 0)}, // 0.142857...
			{mdo(5, 0), mdo(7, 0)}, // 0.714285...
		},
		64: {
			{mdo(1, 0), mdo(3, 0)},
			{mdo(2, 0), mdo(3, 0)},
			{mdo(1, 0), mdo(7, 0)},
			{mdo(5, 0), mdo(7, 0)},
		},
		128: {
			{mdo(1, 0), mdo(3, 0)},
			{mdo(2, 0), mdo(3, 0)},
			{mdo(1, 0), mdo(7, 0)},
			{mdo(5, 0), mdo(7, 0)},
		},
	},
	// Quantize's second operand is the quantum template (the target exponent),
	// so its discriminants are value+quantum pairs whose dropped fraction sits
	// at the half-way point of the target quantum: currency half-up/half-even/
	// directed rounding all become visible when a x.5 value is quantized to
	// integer digits. Every entry drops the nonzero digit 5, so the five
	// directions cannot all agree (ties split NearestEven from NearestAway and
	// the directed modes; the negative mirror splits TowardNegative). The same
	// components are exact operands at every width, so one table serves all
	// three.
	"Quantize": {
		32: {
			{mdo(25, -1), mdo(1, 0)},    // 2.5 -> E0 quantum: NE=2, NA=3, TP=3, TZ=TN=2
			{mdo(35, -1), mdo(1, 0)},    // 3.5 -> NE=NA=TP=4, TZ=TN=3
			{mdoNeg(25, -1), mdo(1, 0)}, // -2.5 -> NE=-2, NA=-3, TN=-3, TZ=TP=-2
			{mdo(15, -1), mdo(1, 0)},    // 1.5 -> NE=NA=TP=2, TZ=TN=1
		},
		64: {
			{mdo(25, -1), mdo(1, 0)},
			{mdo(35, -1), mdo(1, 0)},
			{mdoNeg(25, -1), mdo(1, 0)},
			{mdo(15, -1), mdo(1, 0)},
		},
		128: {
			{mdo(25, -1), mdo(1, 0)},
			{mdo(35, -1), mdo(1, 0)},
			{mdoNeg(25, -1), mdo(1, 0)},
			{mdo(15, -1), mdo(1, 0)},
		},
	},
}

// modeDiscMinPairs is the generation-time floor for discriminant tables: a
// mode-shape operation with fewer entries at any width fails generation, so
// the discriminating corpus cannot silently shrink below usefulness.
const modeDiscMinPairs = 3

// modeBinaryDiscriminantOperands returns the discriminant operand pairs for
// one binary mode-shape operation at one width, failing closed when the
// operation has no table or fewer than modeDiscMinPairs entries.
func modeBinaryDiscriminantOperands(op string, width int) ([][2]modeDiscOperand, error) {
	byWidth, ok := modeBinaryDiscriminants[op]
	if !ok {
		return nil, fmt.Errorf("public parity: no mode-discriminant table for operation %q; every mode-taking binary wrapper needs discriminating operands (add a modeBinaryDiscriminants entry)", op)
	}
	pairs, ok := byWidth[width]
	if !ok {
		return nil, fmt.Errorf("public parity: mode-discriminant table for operation %q has no width-%d entries", op, width)
	}
	if len(pairs) < modeDiscMinPairs {
		return nil, fmt.Errorf("public parity: mode-discriminant table for operation %q width %d has %d entries; at least %d are required", op, width, len(pairs), modeDiscMinPairs)
	}
	return pairs, nil
}

// mixedModeBinaryDiscriminantOperands returns mode-sensitive inputs for the
// Intel D/Q mixed-width arithmetic families. The result precision selects the
// rounding boundary, while each operand is later encoded at its own declared
// width. Decimal128 DD multiplication is the deliberate empty exception:
// multiplying two 16-digit Decimal64 coefficients produces at most 32 digits
// and their exponent ranges also remain inside Decimal128, so every finite DD
// product is exact and no input can distinguish rounding modes. Its routing
// corpus is still compared against the port under all five modes.
func mixedModeBinaryDiscriminantOperands(op string, resultWidth int, operandWidths [2]int) ([][2]modeDiscOperand, error) {
	if resultWidth == 128 && op == "Mul" && operandWidths == [2]int{64, 64} {
		return nil, nil
	}
	if resultWidth == 128 && op == "Mul" {
		// One 16-digit Decimal64 coefficient times one 19-digit Decimal128
		// coefficient yields a 35-digit exact product, one digit beyond
		// Decimal128 precision. Every product ends in a nonzero digit so the
		// discarded digit is provably significant.
		pairs := [][2]modeDiscOperand{
			{mdo(8_999_999_999_999_999, 0), mdo(9_999_999_999_999_999_999, 0)},
			{mdo(8_888_888_888_888_889, 0), mdo(9_999_999_999_999_999_999, 0)},
			{mdo(7_777_777_777_777_777, 0), mdo(9_999_999_999_999_999_999, 0)},
			{mdo(1_234_567_890_123_457, 0), mdo(9_876_543_210_987_654_321, 0)},
		}
		switch operandWidths {
		case [2]int{64, 128}:
			return pairs, nil
		case [2]int{128, 64}:
			for i := range pairs {
				pairs[i][0], pairs[i][1] = pairs[i][1], pairs[i][0]
			}
			return pairs, nil
		default:
			return nil, fmt.Errorf("public parity: unsupported Decimal128 mixed multiplication operand widths %v", operandWidths)
		}
	}
	pairs, err := modeBinaryDiscriminantOperands(op, resultWidth)
	if err != nil {
		return nil, err
	}
	for i, pair := range pairs {
		for j, operand := range pair {
			if _, err := modeDiscGoLiteral(operandWidths[j], operand); err != nil {
				return nil, fmt.Errorf("public parity: mixed %s result width %d pair %d operand %d cannot be encoded at width %d: %w", op, resultWidth, i, j, operandWidths[j], err)
			}
		}
	}
	return pairs, nil
}

// modeUnaryDiscriminants is the discriminant corpus for the unary mode-shape
// wrappers (SqrtWithMode): single operands whose result is irrational, so
// every finite precision rounds and the five directions cannot all agree.
// The same integer components are exact operands at every width.
var modeUnaryDiscriminants = map[string]map[int][]modeDiscOperand{
	"Sqrt": {
		32:  {mdo(2, 0), mdo(3, 0), mdo(5, 0), mdo(7, 0)},
		64:  {mdo(2, 0), mdo(3, 0), mdo(5, 0), mdo(7, 0)},
		128: {mdo(2, 0), mdo(3, 0), mdo(5, 0), mdo(7, 0)},
	},
}

// modeTernaryDiscriminants is the discriminant corpus for the ternary
// mode-shape wrappers (FMAWithMode, operands (d, mul, add) for d*mul + add):
// the fused product-sum is constructed inexact -- a half-ULP tie, a product
// one digit past the precision (the single fused rounding), a sub-ULP
// increment, and a negative mirror.
var modeTernaryDiscriminants = map[string]map[int][][3]modeDiscOperand{
	"FMA": {
		32: {
			{mdo(1, 0), mdo(1, 0), mdo(5, -7)},       // 1*1 + 5e-7: half-ULP tie
			{mdo(3333334, 0), mdo(3, 0), mdo(0, 0)},  // 10000002: 8 digits through the fused rounding
			{mdo(1, 0), mdo(1, 0), mdo(1, -7)},       // 0.1 ULP above 1
			{mdoNeg(1, 0), mdo(1, 0), mdoNeg(1, -7)}, // negative mirror
		},
		64: {
			{mdo(1, 0), mdo(1, 0), mdo(5, -16)},
			{mdo(3333333333333334, 0), mdo(3, 0), mdo(0, 0)},
			{mdo(1, 0), mdo(1, 0), mdo(1, -16)},
			{mdoNeg(1, 0), mdo(1, 0), mdoNeg(1, -16)},
		},
		128: {
			{mdo(1, 0), mdo(1, 0), mdo(5, -34)},
			{mdo(9999999999999999999, 0), mdo(9999999999999999999, 0), mdo(0, 0)}, // (1e19-1)^2: 38 digits
			{mdo(1, 0), mdo(1, 0), mdo(1, -34)},
			{mdoNeg(1, 0), mdo(1, 0), mdoNeg(1, -34)},
		},
	},
}

// modeScaleBDiscCase is one ScaleBWithMode discriminant: the operand plus the
// integer scale exponent (scaleB is exact inside the format's range, so its
// discriminants sit at the overflow/underflow boundary where the result
// rounds: overflow picks infinity or max-finite by direction, and an
// underflow below the minimum subnormal rounds between zero and that
// subnormal, including an exact half tie).
type modeScaleBDiscCase struct {
	Val modeDiscOperand
	Exp int
}

var modeScaleBDiscriminants = map[string]map[int][]modeScaleBDiscCase{
	"ScaleB": {
		32: {
			{mdo(1234567, 0), 91},    // 1.234567e96 * 10 -> overflow(+): inf vs max-finite by direction
			{mdoNeg(1234567, 0), 91}, // overflow(-)
			{mdo(1234567, 0), -107},  // 1.234567e-101 -> subnormal rounding at e_tiny
			{mdo(5, -1), -101},       // 0.5e-101: exact half of the minimum subnormal (tie)
		},
		64: {
			{mdo(1234567890123456, 0), 370},
			{mdoNeg(1234567890123456, 0), 370},
			{mdo(1234567890123456, 0), -414},
			{mdo(5, -1), -398},
		},
		128: {
			{mdo(1234567890123456789, 0), 6127},
			{mdoNeg(1234567890123456789, 0), 6127},
			{mdo(1234567890123456789, 0), -6195},
			{mdo(5, -1), -6176},
		},
	},
}

// modeUnaryDiscriminantOperands / modeTernaryDiscriminantOperands /
// modeScaleBDiscriminantCases mirror modeBinaryDiscriminantOperands's
// strict lookup (missing operation, missing width, or a table below
// modeDiscMinPairs entries fails generation) for the other mode-shape
// arities.
func modeUnaryDiscriminantOperands(op string, width int) ([]modeDiscOperand, error) {
	return modeDiscriminantsFor("unary", modeUnaryDiscriminants, op, width)
}

func modeTernaryDiscriminantOperands(op string, width int) ([][3]modeDiscOperand, error) {
	return modeDiscriminantsFor("ternary", modeTernaryDiscriminants, op, width)
}

func modeScaleBDiscriminantCases(op string, width int) ([]modeScaleBDiscCase, error) {
	return modeDiscriminantsFor("scaleb", modeScaleBDiscriminants, op, width)
}

// modeParseDiscriminants is the discriminant corpus for the explicit-mode
// string constructors (NewDecimal<w>WithMode), keyed "Parse": decimal
// literals carrying one significant digit past the width's precision, so the
// parse must round and the five directions cannot all agree (half-ULP ties
// split NearestEven from NearestAway; the sub-ULP and negative entries split
// the directed modes; the all-nines tie rounds across the decade boundary).
var modeParseDiscriminants = map[string]map[int][]string{
	"Parse": {
		32: {
			"1.0000005",  // 8 digits: half-ULP tie above 1
			"1.0000001",  // 0.1 ULP above 1
			"9.9999995",  // tie at the decade boundary (9.999999 vs 10.00000)
			"-1.0000001", // negative mirror (TowardNegative splits)
		},
		64: {
			"1.0000000000000005",
			"1.0000000000000001",
			"9.9999999999999995",
			"-1.0000000000000001",
		},
		128: {
			"1.0000000000000000000000000000000005",
			"1.0000000000000000000000000000000001",
			"9.9999999999999999999999999999999995",
			"-1.0000000000000000000000000000000001",
		},
	},
}

func modeParseDiscriminantStrings(width int) ([]string, error) {
	return modeDiscriminantsFor("parse", modeParseDiscriminants, "Parse", width)
}

func modeDiscriminantsFor[T any](kind string, tables map[string]map[int][]T, op string, width int) ([]T, error) {
	byWidth, ok := tables[op]
	if !ok {
		return nil, fmt.Errorf("public parity: no %s mode-discriminant table for operation %q; every mode-taking wrapper needs discriminating operands", kind, op)
	}
	cases, ok := byWidth[width]
	if !ok {
		return nil, fmt.Errorf("public parity: %s mode-discriminant table for operation %q has no width-%d entries", kind, op, width)
	}
	if len(cases) < modeDiscMinPairs {
		return nil, fmt.Errorf("public parity: %s mode-discriminant table for operation %q width %d has %d entries; at least %d are required", kind, op, width, len(cases), modeDiscMinPairs)
	}
	return cases, nil
}

// encodeModeDiscOperand32 encodes decimal components into the BID32
// small-coefficient form: sign(1) | exponent(8, bias 101) | coefficient(23).
// The exponent field must stay below 0xC0 so the two steering bits are not
// '11' (which would select the large-coefficient form).
func encodeModeDiscOperand32(o modeDiscOperand) (uint32, error) {
	if o.CoeffHi != 0 || o.CoeffLo >= 1<<23 {
		return 0, fmt.Errorf("mode-discriminant operand coefficient %d:%d does not fit the BID32 small form (< 2^23)", o.CoeffHi, o.CoeffLo)
	}
	field := o.Exp + 101
	if field < 0 || field >= 0xC0 {
		return 0, fmt.Errorf("mode-discriminant operand exponent %d is outside the BID32 small-form range", o.Exp)
	}
	bits := uint32(field)<<23 | uint32(o.CoeffLo)
	if o.Neg {
		bits |= 1 << 31
	}
	return bits, nil
}

// encodeModeDiscOperand64 encodes the BID64 small form: sign(1) |
// exponent(10, bias 398) | coefficient(53); exponent field below 0x300.
func encodeModeDiscOperand64(o modeDiscOperand) (uint64, error) {
	if o.CoeffHi != 0 || o.CoeffLo >= 1<<53 {
		return 0, fmt.Errorf("mode-discriminant operand coefficient %d:%d does not fit the BID64 small form (< 2^53)", o.CoeffHi, o.CoeffLo)
	}
	field := o.Exp + 398
	if field < 0 || field >= 0x300 {
		return 0, fmt.Errorf("mode-discriminant operand exponent %d is outside the BID64 small-form range", o.Exp)
	}
	bits := uint64(field)<<53 | o.CoeffLo
	if o.Neg {
		bits |= 1 << 63
	}
	return bits, nil
}

// encodeModeDiscOperand128 encodes the BID128 small form into the 16-byte
// little-endian layout the value types use (low word then high word): high
// word = sign(1) | exponent(14, bias 6176) | coefficient bits 64..112;
// exponent field below 0x3000.
func encodeModeDiscOperand128(o modeDiscOperand) ([16]byte, error) {
	var out [16]byte
	if o.CoeffHi >= 1<<49 {
		return out, fmt.Errorf("mode-discriminant operand coefficient high word %d does not fit the BID128 small form (< 2^49)", o.CoeffHi)
	}
	field := o.Exp + 6176
	if field < 0 || field >= 0x3000 {
		return out, fmt.Errorf("mode-discriminant operand exponent %d is outside the BID128 small-form range", o.Exp)
	}
	hi := uint64(field)<<49 | o.CoeffHi
	if o.Neg {
		hi |= 1 << 63
	}
	binary.LittleEndian.PutUint64(out[0:8], o.CoeffLo)
	binary.LittleEndian.PutUint64(out[8:16], hi)
	return out, nil
}

// publicParityExceptionMasks maps the five public exception flags to the pinned
// Intel BID_*_EXCEPTION bit values (loaded from bid_functions.h). Only these
// five bits are surfaced by the wrappers' bidgoExceptionFlags converter.
type publicParityExceptionMasks struct {
	Invalid   uint32
	DivByZero uint32
	Overflow  uint32
	Underflow uint32
	Inexact   uint32
}

// loadBidExceptionMasks resolves BID_INVALID_EXCEPTION / BID_ZERO_DIVIDE_EXCEPTION
// / BID_OVERFLOW_EXCEPTION / BID_UNDERFLOW_EXCEPTION / BID_INEXACT_EXCEPTION from
// the pinned Intel header, following the one indirection through the DEC_FE_*
// aliases. The numeric result is emitted as literals into mapPortFlagsForParity
// so the parity comparator never reuses the wrappers' own converter.
func loadBidExceptionMasks(path string) (publicParityExceptionMasks, error) {
	var masks publicParityExceptionMasks
	data, err := os.ReadFile(path)
	if err != nil {
		return masks, fmt.Errorf("read intel header %q: %w", path, err)
	}
	defRe := regexp.MustCompile(`(?m)^#define\s+(\w+)\s+(.+?)\s*$`)
	defs := map[string]string{}
	for _, m := range defRe.FindAllStringSubmatch(string(data), -1) {
		defs[m[1]] = strings.TrimSpace(m[2])
	}
	resolve := func(name string) (uint32, error) {
		rhs, ok := defs[name]
		if !ok {
			return 0, fmt.Errorf("intel header: %s is not defined", name)
		}
		// One indirection: BID_*_EXCEPTION -> DEC_FE_* -> 0xNN.
		if alias, ok := defs[rhs]; ok {
			rhs = alias
		}
		v, err := strconv.ParseUint(strings.TrimPrefix(strings.TrimSpace(rhs), "0x"), 16, 32)
		if err != nil {
			return 0, fmt.Errorf("intel header: %s resolves to non-hex %q: %w", name, rhs, err)
		}
		return uint32(v), nil
	}
	for _, spec := range []struct {
		name string
		dst  *uint32
	}{
		{"BID_INVALID_EXCEPTION", &masks.Invalid},
		{"BID_ZERO_DIVIDE_EXCEPTION", &masks.DivByZero},
		{"BID_OVERFLOW_EXCEPTION", &masks.Overflow},
		{"BID_UNDERFLOW_EXCEPTION", &masks.Underflow},
		{"BID_INEXACT_EXCEPTION", &masks.Inexact},
	} {
		v, err := resolve(spec.name)
		if err != nil {
			return masks, err
		}
		*spec.dst = v
	}
	return masks, nil
}

// publicParityCorpus holds the generated bit-literal corpora for the three
// widths (codec edges, category backfill, deterministic pseudo-random fill),
// the per-width binary/ternary operand index tables resolved from the
// label-level direction matrices, and the per-width non-NaN counts that pin
// the vm_string case counts.
type publicParityCorpus struct {
	Bits32  []uint32
	Bits64  []uint64
	Bits128 [][16]byte

	Pairs32  [][2]int
	Pairs64  [][2]int
	Pairs128 [][2]int

	Triples32  [][3]int
	Triples64  [][3]int
	Triples128 [][3]int

	NonNaN32  int
	NonNaN64  int
	NonNaN128 int
}

func (c publicParityCorpus) pairs(width int) [][2]int {
	switch width {
	case 32:
		return c.Pairs32
	case 64:
		return c.Pairs64
	default:
		return c.Pairs128
	}
}

func (c publicParityCorpus) triples(width int) [][3]int {
	switch width {
	case 32:
		return c.Triples32
	case 64:
		return c.Triples64
	default:
		return c.Triples128
	}
}

func (c publicParityCorpus) nonNaN(width int) int {
	switch width {
	case 32:
		return c.NonNaN32
	case 64:
		return c.NonNaN64
	default:
		return c.NonNaN128
	}
}

// parityCatIndex maps each category to its corpus indices in order.
func parityCatIndexOf(cats []string) map[string][]int {
	byCat := map[string][]int{}
	for i, cat := range cats {
		byCat[cat] = append(byCat[cat], i)
	}
	return byCat
}

func resolveParitySel(byCat map[string][]int, sel paritySel, width int) (int, error) {
	indices := byCat[sel.Cat]
	if len(indices) == 0 {
		return 0, fmt.Errorf("public parity corpus (width %d): required category %q is absent", width, sel.Cat)
	}
	return indices[sel.Nth%len(indices)], nil
}

func resolveParityMatrices(cats []string, width int) ([][2]int, [][3]int, error) {
	byCat := parityCatIndexOf(cats)
	for _, cat := range parityRequiredCats {
		if len(byCat[cat]) == 0 {
			return nil, nil, fmt.Errorf("public parity corpus (width %d): required category %q is absent even after backfill", width, cat)
		}
	}
	pairs := make([][2]int, 0, len(parityLabelPairs))
	for _, lp := range parityLabelPairs {
		a, err := resolveParitySel(byCat, lp[0], width)
		if err != nil {
			return nil, nil, err
		}
		b, err := resolveParitySel(byCat, lp[1], width)
		if err != nil {
			return nil, nil, err
		}
		pairs = append(pairs, [2]int{a, b})
	}
	triples := make([][3]int, 0, len(parityLabelTriples))
	for _, lt := range parityLabelTriples {
		a, err := resolveParitySel(byCat, lt[0], width)
		if err != nil {
			return nil, nil, err
		}
		b, err := resolveParitySel(byCat, lt[1], width)
		if err != nil {
			return nil, nil, err
		}
		c, err := resolveParitySel(byCat, lt[2], width)
		if err != nil {
			return nil, nil, err
		}
		triples = append(triples, [3]int{a, b, c})
	}
	return pairs, triples, nil
}

func countParityNonNaN(cats []string) int {
	n := 0
	for _, cat := range cats {
		if cat != catQNaN && cat != catSNaN {
			n++
		}
	}
	return n
}

func buildPublicParityCorpus() (publicParityCorpus, error) {
	var c publicParityCorpus

	// 32-bit: codec edges, backfill for any missing required category, then
	// deterministic pseudo-random fill up to the fixed corpus length.
	edges32 := ffiEdgeValues(32)
	for _, v := range edges32 {
		c.Bits32 = append(c.Bits32, uint32(v))
	}
	have32 := map[string]bool{}
	for _, v := range c.Bits32 {
		have32[classifyParity32(v)] = true
	}
	for _, cat := range parityRequiredCats {
		if !have32[cat] {
			fill, ok := parityBackfill32[cat]
			if !ok {
				return c, fmt.Errorf("public parity corpus (width 32): no backfill defined for missing category %q", cat)
			}
			c.Bits32 = append(c.Bits32, fill)
			have32[cat] = true
		}
	}
	gen32 := newDeterministicFFIGenerator(0x9e3779b97f4a7c15, "public-parity-32", 32)
	for len(c.Bits32) < publicParityCorpusLen {
		c.Bits32 = append(c.Bits32, uint32(gen32.next()))
	}
	c.Bits32 = c.Bits32[:publicParityCorpusLen]

	edges64 := ffiEdgeValues(64)
	c.Bits64 = append(c.Bits64, edges64...)
	have64 := map[string]bool{}
	for _, v := range c.Bits64 {
		have64[classifyParity64(v)] = true
	}
	for _, cat := range parityRequiredCats {
		if !have64[cat] {
			fill, ok := parityBackfill64[cat]
			if !ok {
				return c, fmt.Errorf("public parity corpus (width 64): no backfill defined for missing category %q", cat)
			}
			c.Bits64 = append(c.Bits64, fill)
			have64[cat] = true
		}
	}
	gen64 := newDeterministicFFIGenerator(0x9e3779b97f4a7c15, "public-parity-64", 64)
	for len(c.Bits64) < publicParityCorpusLen {
		c.Bits64 = append(c.Bits64, gen64.next())
	}
	c.Bits64 = c.Bits64[:publicParityCorpusLen]

	edges128 := ffiWideEdgeValues(128)
	append128 := func(hi, lo uint64) {
		var raw [16]byte
		binary.LittleEndian.PutUint64(raw[0:8], lo)
		binary.LittleEndian.PutUint64(raw[8:16], hi)
		c.Bits128 = append(c.Bits128, raw)
	}
	for _, e := range edges128 {
		append128(e[0], e[1])
	}
	have128 := map[string]bool{}
	for _, raw := range c.Bits128 {
		have128[classifyParity128(raw)] = true
	}
	for _, cat := range parityRequiredCats {
		if !have128[cat] {
			fill, ok := parityBackfill128[cat]
			if !ok {
				return c, fmt.Errorf("public parity corpus (width 128): no backfill defined for missing category %q", cat)
			}
			append128(fill[0], fill[1])
			have128[cat] = true
		}
	}
	gen128 := newDeterministicFFIGenerator(0x9e3779b97f4a7c15, "public-parity-128", 128)
	for len(c.Bits128) < publicParityCorpusLen {
		append128(gen128.next(), gen128.next())
	}
	c.Bits128 = c.Bits128[:publicParityCorpusLen]

	cats32 := make([]string, len(c.Bits32))
	for i, v := range c.Bits32 {
		cats32[i] = classifyParity32(v)
	}
	cats64 := make([]string, len(c.Bits64))
	for i, v := range c.Bits64 {
		cats64[i] = classifyParity64(v)
	}
	cats128 := make([]string, len(c.Bits128))
	for i, raw := range c.Bits128 {
		cats128[i] = classifyParity128(raw)
	}

	var err error
	if c.Pairs32, c.Triples32, err = resolveParityMatrices(cats32, 32); err != nil {
		return c, err
	}
	if c.Pairs64, c.Triples64, err = resolveParityMatrices(cats64, 64); err != nil {
		return c, err
	}
	if c.Pairs128, c.Triples128, err = resolveParityMatrices(cats128, 128); err != nil {
		return c, err
	}
	c.NonNaN32 = countParityNonNaN(cats32)
	c.NonNaN64 = countParityNonNaN(cats64)
	c.NonNaN128 = countParityNonNaN(cats128)
	return c, nil
}

// GeneratePublicParityRunnerOutputs returns the generated parity dispatch and
// cases test files. It is invoked by GeneratePublicParityOutputs after the
// census inventory resolves every mapped symbol's port function.
func GeneratePublicParityRunnerOutputs(repoRoot string, symbols []publicAPISymbol, sigs map[string]bidgoFuncSig, inventory publicAPIInventory) (map[string][]byte, error) {
	if err := validatePublicParityStringCorpus(); err != nil {
		return nil, err
	}
	masks, err := loadBidExceptionMasks(fpJoin(repoRoot, publicParityHeaderPath))
	if err != nil {
		return nil, err
	}
	corpus, err := buildPublicParityCorpus()
	if err != nil {
		return nil, err
	}

	symByID := map[string]publicAPISymbol{}
	for _, s := range symbols {
		symByID[s.Symbol] = s
	}

	var units []parityUnit
	for _, row := range inventory.Symbols {
		if row.Status != "mapped" {
			continue
		}
		sym, ok := symByID[row.Symbol]
		if !ok {
			return nil, fmt.Errorf("public parity runner: mapped symbol %q missing from public source census", row.Symbol)
		}
		unit, err := resolveParityUnit(sym, row.BidgoFunction, sigs, corpus)
		if err != nil {
			return nil, err
		}
		units = append(units, unit)
	}
	sort.Slice(units, func(i, j int) bool { return units[i].Symbol < units[j].Symbol })

	dispatch, cases, err := emitPublicParityFiles(units, corpus, masks)
	if err != nil {
		return nil, err
	}
	formatted, err := formatGeneratedGoOutputs(map[string][]byte{
		publicParityDispatchPath: dispatch,
		publicParityCasesPath:    cases,
	})
	if err != nil {
		return nil, err
	}
	return formatted, nil
}

func fpJoin(parts ...string) string {
	return strings.Join(parts, "/")
}

// parityShape enumerates the concrete comparison template a mapped symbol uses.
type parityShape int

const (
	shapeVMUnary             parityShape = iota // value method, no operand args, generic port
	shapeVMBinary                               // value method, one same-width operand
	shapeVMTernary                              // FMA
	shapeVMScaleB                               // ScaleB(int)
	shapeVMNextToward                           // NextToward(Decimal128BID)
	shapeVMModeUnary                            // ToBinary*/ToDecimal* with a RoundingMode
	shapeVMModeUnaryArith                       // SqrtWithMode: no operand args, a RoundingMode, same-width result
	shapeVMModeBinary                           // {Add,Sub,Mul,Div,Quantize}WithMode: one same-width operand + a RoundingMode
	shapeVMModeTernary                          // FMAWithMode: two same-width operands + a RoundingMode
	shapeVMModeScaleB                           // ScaleBWithMode: int exponent + a RoundingMode
	shapeVMConvert                              // ConvertTo{Int,Uint}N[Exact] (mode-dispatched port)
	shapeVMNullary                              // Radix
	shapeVMCompareTotal                         // CompareTotal / CompareTotalMag (composed)
	shapeVMSign                                 // Sign (composed)
	shapeVMSignalingEqual                       // SignalingEqual (composed)
	shapeVMSignalingNotEqual                    // SignalingNotEqual (composed)
	shapeVMClass                                // Class (int -> class string)
	shapeVMString                               // String (non-NaN only)
	shapeFuncIntCtor                            // NewDecimal*From{Int,Uint}{32,64}
	shapeFuncFromInt                            // NewDecimal*FromInt convenience
	shapeFuncString                             // string constructors / raw parsers
	shapeFuncStringMode                         // NewDecimal*WithMode(s, mode) explicit-mode string constructors
	shapeFuncContext                            // Add*BIDWithContext
	shapeFuncMixedModeBinary                    // Intel mixed-width {Add,Sub,Mul,Div} free functions
)

type parityUnit struct {
	Symbol        string
	FuncName      string
	Shape         parityShape
	Width         int    // receiver / target width
	Method        string // public method name (value methods)
	Func          string // public function name (funcs)
	BidgoFn       string // resolved port function (verification)
	Port          parityPortPlan
	ResultClass   string // dec32/dec64/dec128/bool/int/class/string/f32/f64/bin128/intn/uintn
	HasFlags      bool
	HasMode       bool
	HasErr        bool   // string/from-int funcs return an error
	IntParam      string // int32/uint32/int64/uint64 for scalar ctors
	StringKind    string // "raw" | "direct" | "withflags" for shapeFuncString
	OperandWidths [2]int // mixed-width free-function operand widths; zero for every other shape
	Operation     string // Add/Sub/Mul/Div for mixed-width mode discrimination
	Cases         int    // pinned per-unit case count
}

type parityPortPlan struct {
	GoName        string
	ValueParams   []string // non-rounding, non-flag param types in order
	HasRounding   bool
	FlagsKind     string // none/pointer/result
	PrimaryResult string
}

func planPort(sig bidgoFuncSig) parityPortPlan {
	p := parityPortPlan{GoName: sig.Name, FlagsKind: "none"}
	for _, prm := range sig.Params {
		switch {
		case prm.Type == "*uint32":
			p.FlagsKind = "pointer"
		case isGoportRoundingParam(prm):
			p.HasRounding = true
		default:
			p.ValueParams = append(p.ValueParams, prm.Type)
		}
	}
	if p.FlagsKind != "pointer" && len(sig.Results) >= 2 && sig.Results[len(sig.Results)-1].Type == "uint32" {
		p.FlagsKind = "result"
	}
	if len(sig.Results) > 0 {
		p.PrimaryResult = sig.Results[0].Type
	}
	return p
}

func widthOfRecv(recv string) int {
	switch recv {
	case "Decimal32BID":
		return 32
	case "Decimal64BID":
		return 64
	case "Decimal128BID":
		return 128
	default:
		return 0
	}
}

var newDecimalWidthRe = regexp.MustCompile(`^NewDecimal(32|64|128)`)
var parseDecimalWidthRe = regexp.MustCompile(`^ParseDecimal(32|64|128)BIDRaw$`)
var contextWidthRe = regexp.MustCompile(`^Add(32|64|128)BIDWithContext$`)
var mixedArithmeticRe = regexp.MustCompile(`^(Add|Sub|Mul|Div)(64|128)(DD|DQ|QD|QQ)BIDWithMode$`)

func symbolContainsFlags(results []string) bool {
	for _, r := range results {
		if r == "ExceptionFlags" {
			return true
		}
	}
	return false
}

func symbolContains(list []string, want string) bool {
	for _, v := range list {
		if v == want {
			return true
		}
	}
	return false
}

// resolveParityUnit classifies one mapped symbol into a parity shape and pins
// the port call plan and case count.
func resolveParityUnit(sym publicAPISymbol, bidgoFn string, sigs map[string]bidgoFuncSig, corpus publicParityCorpus) (parityUnit, error) {
	u := parityUnit{
		Symbol:   sym.Symbol,
		FuncName: "publicParity_" + strings.ReplaceAll(sym.Symbol, ".", "_"),
		BidgoFn:  bidgoFn,
		HasFlags: symbolContainsFlags(sym.Results),
		HasMode:  symbolContains(sym.Params, "RoundingMode"),
	}
	if sig, ok := sigs[bidgoFn]; ok {
		u.Port = planPort(sig)
	}

	if sym.Kind == "method" {
		u.Width = widthOfRecv(sym.Recv)
		u.Method = sym.Name
		if u.Width == 0 {
			return u, fmt.Errorf("public parity: method %q has non-value receiver", sym.Symbol)
		}
		return resolveValueMethodUnit(u, sym, corpus)
	}

	// Package functions.
	u.Func = sym.Name
	switch {
	case mixedArithmeticRe.MatchString(sym.Name):
		m := mixedArithmeticRe.FindStringSubmatch(sym.Name)
		u.Operation = m[1]
		u.Width, _ = strconv.Atoi(m[2])
		u.Shape = shapeFuncMixedModeBinary
		u.ResultClass = decClass(u.Width)
		if err := validateMixedArithmeticSignature(sym, u.Width, m[3]); err != nil {
			return u, err
		}
		u.OperandWidths = [2]int{widthOfRecv(sym.Params[0]), widthOfRecv(sym.Params[1])}
		disc, err := mixedModeBinaryDiscriminantOperands(u.Operation, u.Width, u.OperandWidths)
		if err != nil {
			return u, err
		}
		// Each shared routing pair and discriminant pair runs under all five
		// rounding modes. One final case pins invalid-mode rejection through
		// the public flag channel (Intel itself only accepts the five modes).
		u.Cases = (len(parityLabelPairs)+len(disc))*len(parityModeOrder) + 1
		return u, nil
	case contextWidthRe.MatchString(sym.Name):
		m := contextWidthRe.FindStringSubmatch(sym.Name)
		u.Width, _ = strconv.Atoi(m[1])
		u.Shape = shapeFuncContext
		u.ResultClass = decClass(u.Width)
		// per pair: 5 explicit-context modes + 5 nil-context runs (one per
		// pinned global default mode)
		u.Cases = len(parityLabelPairs) * len(parityModeOrder) * 2
		return u, nil
	case parseDecimalWidthRe.MatchString(sym.Name):
		m := parseDecimalWidthRe.FindStringSubmatch(sym.Name)
		u.Width, _ = strconv.Atoi(m[1])
		u.Shape = shapeFuncString
		u.ResultClass = decClass(u.Width)
		u.StringKind = "raw"
		u.Cases = len(publicParityStringCorpusCases)
		return u, nil
	case strings.HasSuffix(sym.Name, "FromInt") && newDecimalWidthRe.MatchString(sym.Name):
		m := newDecimalWidthRe.FindStringSubmatch(sym.Name)
		u.Width, _ = strconv.Atoi(m[1])
		u.Shape = shapeFuncFromInt
		u.ResultClass = decClass(u.Width)
		u.IntParam = sym.Params[0] // int32 or int64
		u.HasErr = true
		u.Cases = intCorpusLen(u.IntParam)
		return u, nil
	case newDecimalWidthRe.MatchString(sym.Name) && isScalarCtor(sym.Name):
		m := newDecimalWidthRe.FindStringSubmatch(sym.Name)
		u.Width, _ = strconv.Atoi(m[1])
		u.Shape = shapeFuncIntCtor
		u.ResultClass = decClass(u.Width)
		u.IntParam = sym.Params[0]
		u.Cases = intCorpusLen(u.IntParam)
		if u.HasMode {
			u.Cases *= len(parityModeOrder)
		}
		return u, nil
	case newDecimalWidthRe.MatchString(sym.Name) && strings.HasSuffix(sym.Name, "WithMode"):
		// Explicit-mode string constructors: NewDecimal*WithMode(s, mode).
		m := newDecimalWidthRe.FindStringSubmatch(sym.Name)
		u.Width, _ = strconv.Atoi(m[1])
		u.Shape = shapeFuncStringMode
		u.ResultClass = decClass(u.Width)
		u.HasErr = true
		discP, err := modeParseDiscriminantStrings(u.Width)
		if err != nil {
			return u, err
		}
		u.Cases = (len(publicParityStringCorpusCases) + len(discP)) * len(parityModeOrder)
		return u, nil
	case newDecimalWidthRe.MatchString(sym.Name):
		// String constructors: NewDecimal*, NewDecimal*WithFlags, NewDecimal*BIDDirect.
		m := newDecimalWidthRe.FindStringSubmatch(sym.Name)
		u.Width, _ = strconv.Atoi(m[1])
		u.Shape = shapeFuncString
		u.ResultClass = decClass(u.Width)
		u.HasErr = symbolContains(sym.Results, "error")
		if !u.HasErr {
			return u, fmt.Errorf("public parity: string constructor %q does not return an error", sym.Symbol)
		}
		if u.HasFlags {
			u.StringKind = "withflags"
		} else {
			u.StringKind = "direct"
		}
		u.Cases = len(publicParityStringCorpusCases)
		return u, nil
	default:
		return u, fmt.Errorf("public parity: unclassified mapped function %q", sym.Symbol)
	}
}

func validateMixedArithmeticSignature(sym publicAPISymbol, resultWidth int, operandCode string) error {
	if len(sym.Params) != 3 || len(sym.Results) != 2 {
		return fmt.Errorf("public parity: mixed arithmetic function %q has signature params=%v results=%v, want two decimal operands plus RoundingMode and (Decimal%dBID, ExceptionFlags)", sym.Symbol, sym.Params, sym.Results, resultWidth)
	}
	wantOperandWidths := [2]int{}
	for i, code := range operandCode {
		switch code {
		case 'D':
			wantOperandWidths[i] = 64
		case 'Q':
			wantOperandWidths[i] = 128
		default:
			return fmt.Errorf("public parity: mixed arithmetic function %q has unsupported operand code %q", sym.Symbol, operandCode)
		}
	}
	for i, want := range wantOperandWidths {
		if got := widthOfRecv(sym.Params[i]); got != want {
			return fmt.Errorf("public parity: mixed arithmetic function %q operand %d is %q, want Decimal%dBID from suffix %s", sym.Symbol, i, sym.Params[i], want, operandCode)
		}
	}
	if sym.Params[2] != "RoundingMode" || sym.Results[0] != fmt.Sprintf("Decimal%dBID", resultWidth) || sym.Results[1] != "ExceptionFlags" {
		return fmt.Errorf("public parity: mixed arithmetic function %q has signature params=%v results=%v inconsistent with its name", sym.Symbol, sym.Params, sym.Results)
	}
	return nil
}

var parityModeOrder = []string{"RoundNearestEven", "RoundNearestAway", "RoundTowardZero", "RoundTowardPositive", "RoundTowardNegative"}

func decClass(width int) string {
	switch width {
	case 32:
		return "dec32"
	case 64:
		return "dec64"
	default:
		return "dec128"
	}
}

func isScalarCtor(name string) bool {
	for _, suffix := range []string{"FromInt32", "FromInt64", "FromUint32", "FromUint64"} {
		if strings.HasSuffix(name, suffix) {
			return true
		}
	}
	return false
}

func intCorpusLen(paramType string) int {
	switch paramType {
	case "int32":
		return len(publicParityIntCorpus32)
	case "uint32":
		return len(publicParityUintCorpus32)
	case "int64":
		return len(publicParityIntCorpus64)
	case "uint64":
		return len(publicParityUintCorpus64)
	default:
		return 0
	}
}

func resolveValueMethodUnit(u parityUnit, sym publicAPISymbol, corpus publicParityCorpus) (parityUnit, error) {
	m := sym.Name
	// Composed / special-cased methods first.
	switch m {
	case "CompareTotal", "CompareTotalMag":
		u.Shape = shapeVMCompareTotal
		u.Cases = len(parityLabelPairs)
		return u, nil
	case "Sign":
		u.Shape = shapeVMSign
		u.Cases = publicParityCorpusLen
		return u, nil
	case "SignalingEqual":
		u.Shape = shapeVMSignalingEqual
		u.Cases = len(parityLabelPairs)
		return u, nil
	case "SignalingNotEqual":
		u.Shape = shapeVMSignalingNotEqual
		u.Cases = len(parityLabelPairs)
		return u, nil
	case "Class":
		u.Shape = shapeVMClass
		u.Cases = publicParityCorpusLen
		return u, nil
	case "String":
		// NaN corpus entries are display plumbing (hand-written NaN payload
		// tests own them) and are skipped without counting, so the pinned
		// count is the per-width non-NaN corpus size.
		u.Shape = shapeVMString
		u.Cases = corpus.nonNaN(u.Width)
		return u, nil
	case "Radix":
		u.Shape = shapeVMNullary
		u.Cases = 1
		return u, nil
	}
	if strings.HasPrefix(m, "ConvertTo") {
		u.Shape = shapeVMConvert
		u.ResultClass = convertResultClass(sym.Results[0])
		u.Cases = publicParityCorpusLen * len(parityModeOrder)
		return u, nil
	}

	// Determine the primary result class from the first result.
	u.ResultClass = classForResult(sym.Results[0])
	if u.ResultClass == "" {
		return u, fmt.Errorf("public parity: method %q has unsupported result %q", sym.Symbol, sym.Results[0])
	}

	// Operand shape from parameters (excluding a trailing mode).
	valueParams := nonModeParams(sym.Params)
	switch {
	case len(sym.Params) == 1 && sym.Params[0] == "RoundingMode":
		if sym.Results[0] == sym.Recv {
			// Same-width mode-taking arithmetic (SqrtWithMode) carries a
			// required discriminant table; the format conversions
			// (ToBinary*/ToDecimal*) stay on the plain mode-unary shape.
			u.Shape = shapeVMModeUnaryArith
			disc, err := modeUnaryDiscriminantOperands(strings.TrimSuffix(sym.Name, "WithMode"), u.Width)
			if err != nil {
				return u, err
			}
			u.Cases = (publicParityCorpusLen + len(disc)) * len(parityModeOrder)
		} else {
			u.Shape = shapeVMModeUnary
			u.Cases = publicParityCorpusLen * len(parityModeOrder)
		}
	case len(valueParams) == 0:
		u.Shape = shapeVMUnary
		u.Cases = publicParityCorpusLen
	case len(valueParams) == 1 && valueParams[0] == sym.Recv && u.HasMode:
		u.Shape = shapeVMModeBinary
		disc, err := modeBinaryDiscriminantOperands(strings.TrimSuffix(sym.Name, "WithMode"), u.Width)
		if err != nil {
			return u, err
		}
		u.Cases = (len(parityLabelPairs) + len(disc)) * len(parityModeOrder)
	case len(valueParams) == 1 && valueParams[0] == sym.Recv:
		u.Shape = shapeVMBinary
		u.Cases = len(parityLabelPairs)
	case len(valueParams) == 2 && valueParams[0] == sym.Recv && valueParams[1] == sym.Recv && u.HasMode:
		u.Shape = shapeVMModeTernary
		discT, err := modeTernaryDiscriminantOperands(strings.TrimSuffix(sym.Name, "WithMode"), u.Width)
		if err != nil {
			return u, err
		}
		u.Cases = (len(parityLabelTriples) + len(discT)) * len(parityModeOrder)
	case len(valueParams) == 2 && valueParams[0] == sym.Recv && valueParams[1] == sym.Recv:
		u.Shape = shapeVMTernary
		u.Cases = len(parityLabelTriples)
	case len(valueParams) == 1 && valueParams[0] == "int" && u.HasMode:
		u.Shape = shapeVMModeScaleB
		discS, err := modeScaleBDiscriminantCases(strings.TrimSuffix(sym.Name, "WithMode"), u.Width)
		if err != nil {
			return u, err
		}
		u.Cases = (publicParityCorpusLen*len(publicParityScaleBExps) + len(discS)) * len(parityModeOrder)
	case len(valueParams) == 1 && valueParams[0] == "int":
		u.Shape = shapeVMScaleB
		u.Cases = publicParityCorpusLen * len(publicParityScaleBExps)
	case len(valueParams) == 1 && valueParams[0] == "Decimal128BID":
		u.Shape = shapeVMNextToward
		u.Cases = publicParityCorpusLen * publicParityNextTowardTargets
	default:
		return u, fmt.Errorf("public parity: method %q has unsupported parameter shape %v", sym.Symbol, sym.Params)
	}
	return u, nil
}

const publicParityNextTowardTargets = 5

func nonModeParams(params []string) []string {
	var out []string
	for _, p := range params {
		if p == "RoundingMode" {
			continue
		}
		out = append(out, p)
	}
	return out
}

func classForResult(result string) string {
	switch result {
	case "Decimal32BID":
		return "dec32"
	case "Decimal64BID":
		return "dec64"
	case "Decimal128BID":
		return "dec128"
	case "bool":
		return "bool"
	case "int":
		return "int"
	case "float32":
		return "f32"
	case "float64":
		return "f64"
	case "Binary128":
		return "bin128"
	default:
		return ""
	}
}

func convertResultClass(result string) string {
	switch result {
	case "int8", "int16", "int32", "int64":
		return "intn"
	case "uint8", "uint16", "uint32", "uint64":
		return "uintn"
	default:
		return ""
	}
}
