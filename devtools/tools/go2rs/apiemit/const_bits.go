package apiemit

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"strings"
)

// This file computes the compile-time BID interchange bit patterns for the 12
// ZERO/ONE/PI/E convenience values. It does not call the generated Rust parser,
// bidgo, or Intel BID C. The independent derivation keeps the Rust public-API
// parity check (`const bits == parse(documented literal) bits`) meaningful.
//
// Deliberately narrow scope: only a non-negative coefficient in the BID
// interchange format's direct/"short" field form (IEEE 754-2008 clause 3.5.2)
// is implemented -- no sign, no NaN/Infinity, no the "steering"/long form a
// coefficient whose value is at or above the direct field's 2^N capacity would
// need. None of the 12 ZERO/ONE/PI/E literals (leading digit 0, 1, 2, or 3,
// well under any width's steering threshold -- verified against
// bid_codec_reference.go's refEncode32/64/128 for all 12 literals during
// development) ever exercises that path; encodeBID32/64/128 reject unresolved input
// (generation-time error, not a silently wrong bit pattern) if a future
// literal ever would, forcing this encoder to be extended deliberately rather
// than mis-encoding.

// parseSimplePositiveDecimal parses a plain fixed-point decimal literal with
// no sign, no exponent suffix, and no NaN/Infinity spelling -- e.g. "0", "1",
// "3.141592653589793" -- into its coefficient (all digits with the decimal
// point removed, as an arbitrary-precision non-negative integer) and quantum
// exponent (-(number of digits after the decimal point)). This mirrors the
// IEEE 754 from-string contract for a literal with no exponent suffix: the
// quantum exponent is exactly the position of the decimal point as written,
// with no trailing-zero normalization. It is deliberately not a general
// decimal-string parser -- anything outside this narrow shape (sign, exponent
// notation, NaN/Infinity, empty string, more than one '.') is a strict
// generation-time error rather than a best-effort guess.
func parseSimplePositiveDecimal(lit string) (*big.Int, int, error) {
	if lit == "" {
		return nil, 0, fmt.Errorf("empty literal")
	}
	dot := -1
	for i, r := range lit {
		switch {
		case r >= '0' && r <= '9':
			continue
		case r == '.':
			if dot >= 0 {
				return nil, 0, fmt.Errorf("literal %q has more than one '.'", lit)
			}
			dot = i
		default:
			return nil, 0, fmt.Errorf("literal %q contains %q, which this encoder does not support (only a plain non-negative fixed-point digit string is accepted -- no sign, exponent suffix, or NaN/Infinity spelling)", lit, string(r))
		}
	}
	intPart, fracPart := lit, ""
	if dot >= 0 {
		intPart, fracPart = lit[:dot], lit[dot+1:]
	}
	digits := intPart + fracPart
	if digits == "" {
		return nil, 0, fmt.Errorf("literal %q has no digits", lit)
	}
	coeff, ok := new(big.Int).SetString(digits, 10)
	if !ok {
		return nil, 0, fmt.Errorf("literal %q: %q is not a valid base-10 integer", lit, digits)
	}
	return coeff, -len(fracPart), nil
}

// Fixed IEEE 754-2008 clause 3.5.2 BID interchange-format parameters this
// encoder needs, width-keyed. bidShortFieldBits32/64 are the direct/short-form
// coefficient field widths: a coefficient at or above 2^N needs the
// "steering"/long form this encoder does not implement (see the file doc
// comment). BID128 has no such branch here: every legal (<= 34-digit, i.e. <
// 10^34) BID128 coefficient fits the 113-bit direct field (2^113 > 10^34), so
// encodeBID128 always uses the direct form -- verified against
// bid_codec_reference.go's refEncode128, whose Normal-kind case has no
// magnitude branch either.
//
// The finite-value bounds below are the EXACT decimal bounds a canonical
// finite BID<w> value must satisfy, matching bid754-codec-go/decimal.go's
// encoder bounds (bid32MaxCoeff=10^7-1 / bid64MaxCoeff=10^16-1 / ten34; the
// unbiased exponent ranges [-101,90] / [-398,369] / [-6176,6111], i.e. biased
// [0,191] / [0,767] / [0,12287]) and bid_codec_reference.go's clamp bounds.
// They are checked with exact big.Int/integer comparisons, NOT a BitLen()
// approximation: 10^34 (an invalid BID128 coefficient) is below 2^113, so a
// bit-length bound would silently pass it and emit a noncanonical/invalid
// pattern. An out-of-range coefficient or exponent is a hard error (never a
// silent truncate/clamp/mask), so this reuse encoder cannot emit a wrong bit
// pattern for an out-of-scope literal -- the project's no-silent-failure
// commitment applied to the constant-encoding path.
const (
	bidBias32  = 101
	bidBias64  = 398
	bidBias128 = 6176

	// Direct/"short"-form significand field widths. A coefficient at or above
	// 2^N needs the steering/long form this encoder does not implement. For
	// BID32/BID64 this field (2^23 / 2^53) is NARROWER than the decimal
	// coefficient bound (10^7 / 10^16), so a coefficient in [2^23,10^7) /
	// [2^53,10^16) is a valid finite value this encoder still must reject
	// (rather than overflow the field and alter the exponent bits). For
	// BID128 the decimal bound 10^34-1 is below 2^113, so every valid BID128
	// coefficient fits the 113-bit direct field and the decimal bound alone
	// is sufficient (encodeBID128 has no separate field-capacity branch).
	bidShortFieldBits32 = 23
	bidShortFieldBits64 = 53

	// Biased exponent maxima (finite range upper bound); minimum is 0.
	bidBiasedExpMax32  = 191
	bidBiasedExpMax64  = 767
	bidBiasedExpMax128 = 12287
)

// pow10 returns 10^n as a big.Int (n small, non-negative). Used for the exact
// per-width decimal coefficient bounds, which exceed int64 for BID128 (10^34).
func pow10(n int) *big.Int {
	return new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(n)), nil)
}

// Exclusive decimal coefficient bounds: a finite BID<w> coefficient is
// strictly below 10^7 / 10^16 / 10^34. A coefficient at or above the bound is
// not a representable finite value in that width at all.
var (
	bid32CoeffBound  = pow10(7)
	bid64CoeffBound  = pow10(16)
	bid128CoeffBound = pow10(34)
)

func encodeBID32(coeff *big.Int, exponent int) (uint32, error) {
	if coeff.Sign() < 0 {
		return 0, fmt.Errorf("negative coefficient %s not supported", coeff)
	}
	if coeff.Cmp(bid32CoeffBound) >= 0 {
		return 0, fmt.Errorf("coefficient %s is not a valid finite BID32 coefficient (must be < 10^7)", coeff)
	}
	if coeff.BitLen() > bidShortFieldBits32 {
		return 0, fmt.Errorf("coefficient %s needs the BID32 steering/long form, which this encoder does not implement (extend it before widening scope to a literal this large)", coeff)
	}
	biased := exponent + bidBias32
	if biased < 0 || biased > bidBiasedExpMax32 {
		return 0, fmt.Errorf("biased exponent %d out of the BID32 finite range [0,%d]", biased, bidBiasedExpMax32)
	}
	return (uint32(biased) << bidShortFieldBits32) | uint32(coeff.Uint64()), nil
}

func encodeBID64(coeff *big.Int, exponent int) (uint64, error) {
	if coeff.Sign() < 0 {
		return 0, fmt.Errorf("negative coefficient %s not supported", coeff)
	}
	if coeff.Cmp(bid64CoeffBound) >= 0 {
		return 0, fmt.Errorf("coefficient %s is not a valid finite BID64 coefficient (must be < 10^16)", coeff)
	}
	if coeff.BitLen() > bidShortFieldBits64 {
		return 0, fmt.Errorf("coefficient %s needs the BID64 steering/long form, which this encoder does not implement (extend it before widening scope to a literal this large)", coeff)
	}
	biased := exponent + bidBias64
	if biased < 0 || biased > bidBiasedExpMax64 {
		return 0, fmt.Errorf("biased exponent %d out of the BID64 finite range [0,%d]", biased, bidBiasedExpMax64)
	}
	return (uint64(biased) << bidShortFieldBits64) | coeff.Uint64(), nil
}

// encodeBID128 packs coeff/exponent into the BID128 interchange bit pattern's
// (lo, hi) 64-bit words, in the same word convention as
// crate::gen_types::BID_UINT128{w:[lo,hi]} and, via bid_uint128_to_le_bytes,
// Decimal128's public little-endian byte contract.
func encodeBID128(coeff *big.Int, exponent int) (lo, hi uint64, err error) {
	if coeff.Sign() < 0 {
		return 0, 0, fmt.Errorf("negative coefficient %s not supported", coeff)
	}
	if coeff.Cmp(bid128CoeffBound) >= 0 {
		return 0, 0, fmt.Errorf("coefficient %s is not a valid finite BID128 coefficient (must be < 10^34)", coeff)
	}
	biased := exponent + bidBias128
	if biased < 0 || biased > bidBiasedExpMax128 {
		return 0, 0, fmt.Errorf("biased exponent %d out of the BID128 finite range [0,%d]", biased, bidBiasedExpMax128)
	}
	// coeff < 10^34 < 2^113, so the top 15 bits of the 128-bit big-endian
	// fill are zero and hiCoeff below is < 2^49: the 49-bit mask is a no-op
	// on every valid coefficient (never a silent truncation -- an oversized
	// coefficient was already rejected above).
	var full [16]byte
	coeff.FillBytes(full[:])
	hiCoeff := binary.BigEndian.Uint64(full[0:8])
	lo = binary.BigEndian.Uint64(full[8:16])
	hi = (uint64(biased) << 49) | (hiCoeff & 0x0001ffffffffffff)
	return lo, hi, nil
}

// constBitsRustLiteral computes width w's packed BID bits for literal and
// renders them as the Rust literal expression the constant template embeds: a
// hex u32/u64 literal for width 32/64 (passed to Self::from_bits), or a
// 16-entry byte-array literal for width 128 (passed to Self::from_le_bytes,
// little-endian per Decimal128's public byte contract).
func constBitsRustLiteral(literal string, w widthSpec) (string, error) {
	coeff, exponent, err := parseSimplePositiveDecimal(literal)
	if err != nil {
		return "", fmt.Errorf("literal %q: %w", literal, err)
	}
	switch w.digits {
	case "32":
		bits, err := encodeBID32(coeff, exponent)
		if err != nil {
			return "", fmt.Errorf("literal %q at width 32: %w", literal, err)
		}
		return fmt.Sprintf("0x%08x", bits), nil
	case "64":
		bits, err := encodeBID64(coeff, exponent)
		if err != nil {
			return "", fmt.Errorf("literal %q at width 64: %w", literal, err)
		}
		return fmt.Sprintf("0x%016x", bits), nil
	case "128":
		lo, hi, err := encodeBID128(coeff, exponent)
		if err != nil {
			return "", fmt.Errorf("literal %q at width 128: %w", literal, err)
		}
		var bytes [16]byte
		binary.LittleEndian.PutUint64(bytes[0:8], lo)
		binary.LittleEndian.PutUint64(bytes[8:16], hi)
		parts := make([]string, 16)
		for i, bb := range bytes {
			parts[i] = fmt.Sprintf("0x%02x", bb)
		}
		return "[" + strings.Join(parts, ", ") + "]", nil
	default:
		return "", fmt.Errorf("unsupported width digits %q", w.digits)
	}
}

// constEmitOrder is the fixed, deterministic rendering order buildDecimalRs
// sorts a width's constants rows into before calling emitConstOps: Go source
// declaration order (bid754-go/api_v2.go declares Zero/One/Pi/E in that
// order), not manifest JSON array order, so the generated diff stays stable
// regardless of how the manifest rows happen to be listed.
var constEmitOrder = map[string]int{
	"ZERO": 0,
	"ONE":  1,
	"PI":   2,
	"E":    3,
}

// widthSpecForOwner resolves a constants-manifest row's rust_owner to its
// widthSpec record, the same three-owner closed set buildDecimalRs's three
// call sites (buildDecimal32Rs/buildDecimal64Rs/buildDecimal128Rs) filter
// manifest.Constants against by selfType. Kept as the single source of truth
// both resolveConstantsClosure (apiemit.go) and this strict check share,
// so the two can never drift into disagreeing about which owners are valid.
func widthSpecForOwner(rustOwner string) (widthSpec, bool) {
	switch rustOwner {
	case decimal32Width.selfType:
		return decimal32Width, true
	case decimal64Width.selfType:
		return decimal64Width, true
	case decimal128Width.selfType:
		return decimal128Width, true
	default:
		return widthSpec{}, false
	}
}

// constDocBlurb is the closed set of rustdoc lead sentences for the four known
// constant kinds the 12 manifest constants rows use (ZERO/ONE/PI/E). A
// rust_const value outside this set fails generation instead of emitting an
// undocumented constant.
var constDocBlurb = map[string]string{
	"ZERO": "The value zero (`+0`, exponent 0).",
	"ONE":  "The value one (coefficient 1, exponent 0).",
	"PI":   "The mathematical constant π (pi), rounded to the widest coefficient this format can represent.",
	"E":    "The mathematical constant e (Euler's number), rounded to the widest coefficient this format can represent.",
}

// emitConstOps renders width w's constants-manifest rows (ZERO/ONE/PI/E,
// already filtered and ordered by the caller) as compile-time associated
// consts, placed right after RADIX in buildDecimalRs's template (mirrors
// emitRadixConstOp's fixed-spec-invariant placement. The bit pattern comes
// from constBitsRustLiteral's independent generation-time encoder above,
// never a port call; the literal itself is kept in the rustdoc alongside a
// pointer to the Rust public-API parity gate, which cross-checks
// const.to_bits()/to_le_bytes() == a fresh parse(literal) at cargo-test time
// (rust_public_parity_emit.go's emitRustConstParityTests).
func emitConstOps(b *strings.Builder, rows []constantRule, w widthSpec) error {
	accessor := "to_bits()"
	if w.digits == "128" {
		accessor = "to_le_bytes()"
	}
	for _, r := range rows {
		blurb, ok := constDocBlurb[r.RustConst]
		if !ok {
			return fmt.Errorf("apiemit: constants go_symbol %q: unrecognized rust_const %q (extend constDocBlurb before adding a new constant kind)", r.GoSymbol, r.RustConst)
		}
		bitsLit, err := constBitsRustLiteral(r.Literal, w)
		if err != nil {
			return fmt.Errorf("apiemit: constants go_symbol %q: %w", r.GoSymbol, err)
		}
		ctor := w.selfType + "::from_bits(" + bitsLit + ")"
		if w.digits == "128" {
			ctor = w.selfType + "::from_le_bytes(" + bitsLit + ")"
		}
		fmt.Fprintf(b, `
    /// %s Literal: %q.
    ///
    /// Computed at generation time by an independent BID interchange-format
    /// encoder (devtools/tools/go2rs/apiemit/const_bits.go), not the
    /// generated from_string parser; the Rust public-API parity gate
    /// cross-checks this constant's %s against a fresh runtime parse of the
    /// same documented literal at cargo-test time.
    pub const %s: %s = %s;
`, blurb, r.Literal, accessor, r.RustConst, w.selfType, ctor)
	}
	return nil
}
