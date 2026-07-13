package testgen

// Tier 1 routing-sentinel codegen.
//
// The Tier 1 long differential runners compare port/public/Rust results
// against pinned Intel C bit-for-bit, but a glue bug applied identically to
// every leg (operand slot swap, rounding-mode miswiring, dispatch-row
// mislabeling) produces an agreed-upon wrong answer that the differential
// cannot see. This file selects a small, deterministic set of known-answer
// sentinel rows per (operation, width): each row pins the expected
// (result bits, raw flags) computed at generation time through the public
// bid754-go API (publicroute proves that surface routes through the Go
// mechanical port; the build is cgo-free). At runtime both generated runners
// require pinned == Intel C == port/public for every row, so a routing bug
// on any leg — including one introduced at pin time — diverges from live
// Intel C on the first full run instead of passing silently.
//
// Selection is a deterministic greedy walk over hand-declared candidate
// tables (fixed declaration order, pure oracle), so regeneration is
// byte-identical and `make verify-generated` enforces it. The selector
// mechanically verifies three sensitivity properties before emitting rows:
//
//	S1 (slot):     for every unordered operand-slot pair of a multi-operand
//	               operation, at least one adopted tuple changes its result
//	               when that pair is swapped (commutative operations satisfy
//	               this through distinct-qNaN-payload tuples, measured — not
//	               assumed — via the oracle). scaleb's typed integer slot
//	               cannot be swapped, so its S1 analogue requires the
//	               exponent sign to be visible (n vs -n).
//	S2 (mode):     for every unordered rounding-mode pair of a mode-taking
//	               operation, at least one adopted tuple produces different
//	               results under the two modes.
//	S3 (dispatch): for every same-signature operation pair inside a dispatch
//	               family, the adopted tuples of each operation distinguish
//	               it from every other operation in the family.
//
// A requirement that no candidate can satisfy fails generation immediately
// (no partial output, no silent fallback) unless it is waived by an explicit
// reasoned entry in tier1SentinelExceptions; every exception entry must be
// genuinely unsatisfiable by the full candidate pool at every width — a
// separable excepted requirement fails generation as a stale waiver
// (closed-world, GUARDRAILS allowlist principle).
//
// GUARDRAILS: this generator never reads or writes
// devtools/verification_anchors.json or devtools/verification_sentinels.json.
// The rows land only in the two generated runners; the human pin flows
// through `cmd/testgen -print-sentinel-anchors` stdout (see
// Tier1SentinelAnchorProposal) and a manual paste audited by
// TestVerificationAnchorsMatchGeneratedArtifacts.

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/bits"
	"strings"

	bid754 "github.com/sky1core/bid754/bid754-go"
)

// tier1SentinelWidths is the canonical width iteration order (row order:
// width ascending, then family declaration order, then operation declaration
// order, then tuple adoption order, then native mode table order).
var tier1SentinelWidths = [...]int{32, 64, 128}

func tier1SentinelWidthLabel(width int) string {
	return fmt.Sprintf("d%d", width)
}

// tier1SentinelMode pairs one native Intel rounding-mode integer with its
// public bid754-go RoundingMode and the runner mode-table name. This table is
// the codegen-owned truth for the native<->public correspondence; the
// generated runners carry the same pairing, and any divergence between the
// two glues fails the runtime pinned==C==port/public comparison (false-fail
// direction, never false-pass).
type tier1SentinelMode struct {
	name   string
	native int
	public bid754.RoundingMode
}

var tier1SentinelModes = [...]tier1SentinelMode{
	{name: "nearest_even", native: 0, public: bid754.RoundNearestEven},
	{name: "nearest_away", native: 4, public: bid754.RoundNearestAway},
	{name: "toward_zero", native: 3, public: bid754.RoundTowardZero},
	{name: "toward_positive", native: 2, public: bid754.RoundTowardPositive},
	{name: "toward_negative", native: 1, public: bid754.RoundTowardNegative},
}

// tier1SentinelModeResults carries one candidate's oracle result per mode in
// tier1SentinelModes order.
type tier1SentinelModeResults [len(tier1SentinelModes)]string

// tier1SentinelTupleCap bounds the adopted tuples per (operation, width) so a
// candidate-table mistake cannot explode the pinned row set; hitting the cap
// fails generation and the candidate table must be improved by hand.
const tier1SentinelTupleCap = 6

// tier1SentinelOperand is one operand of a sentinel candidate: either decimal
// components encoded through the width's small BID form (shared with the
// public-parity mode-discriminant corpus) or raw width bits for values the
// small form cannot express (NaN payloads).
type tier1SentinelOperand struct {
	raw  bool
	bits bid128BidCodecValue
	dec  modeDiscOperand
}

func sentinelDec(o modeDiscOperand) tier1SentinelOperand {
	return tier1SentinelOperand{dec: o}
}

func sentinelRaw32(bits uint32) tier1SentinelOperand {
	return tier1SentinelOperand{raw: true, bits: bid128BidCodecValue{lo: uint64(bits)}}
}

func sentinelRaw64(bits uint64) tier1SentinelOperand {
	return tier1SentinelOperand{raw: true, bits: bid128BidCodecValue{lo: bits}}
}

func sentinelRaw128(lo, hi uint64) tier1SentinelOperand {
	return tier1SentinelOperand{raw: true, bits: bid128BidCodecValue{lo: lo, hi: hi}}
}

// mdoStr builds a modeDiscOperand whose coefficient does not fit a uint64
// literal, from its decimal digit string (width-128 sentinel candidates). It
// panics on a malformed table literal: candidate tables are compile-time
// constants and a bad digit is a programming error, not an input condition.
func mdoStr(digits string, exp int) modeDiscOperand {
	if digits == "" {
		panic("tier1 sentinel: empty coefficient digit string")
	}
	var hi, lo uint64
	for _, ch := range digits {
		if ch < '0' || ch > '9' {
			panic(fmt.Sprintf("tier1 sentinel: coefficient digit string %q has a non-digit", digits))
		}
		hiCarry, hiTimes10 := bits.Mul64(hi, 10)
		loCarry, loTimes10 := bits.Mul64(lo, 10)
		newLo, addCarry := bits.Add64(loTimes10, uint64(ch-'0'), 0)
		newHi, hiOverflow := bits.Add64(hiTimes10, loCarry+addCarry, 0)
		if hiCarry != 0 || hiOverflow != 0 {
			panic(fmt.Sprintf("tier1 sentinel: coefficient digit string %q overflows 128 bits", digits))
		}
		hi, lo = newHi, newLo
	}
	return modeDiscOperand{CoeffHi: hi, CoeffLo: lo, Exp: exp}
}

// encode materializes the operand at one width, failing on any component
// outside the width's representable candidate range.
func (o tier1SentinelOperand) encode(width int) (bid128BidCodecValue, error) {
	if o.raw {
		switch width {
		case 32:
			if o.bits.hi != 0 || o.bits.lo > 0xffffffff {
				return bid128BidCodecValue{}, fmt.Errorf("tier1 sentinel raw operand %016x:%016x does not fit width 32", o.bits.hi, o.bits.lo)
			}
		case 64:
			if o.bits.hi != 0 {
				return bid128BidCodecValue{}, fmt.Errorf("tier1 sentinel raw operand %016x:%016x does not fit width 64", o.bits.hi, o.bits.lo)
			}
		case 128:
		default:
			return bid128BidCodecValue{}, fmt.Errorf("unsupported tier1 sentinel width %d", width)
		}
		return o.bits, nil
	}
	switch width {
	case 32:
		encoded, err := encodeModeDiscOperand32(o.dec)
		if err != nil {
			return bid128BidCodecValue{}, err
		}
		return bid128BidCodecValue{lo: uint64(encoded)}, nil
	case 64:
		encoded, err := encodeModeDiscOperand64(o.dec)
		if err != nil {
			return bid128BidCodecValue{}, err
		}
		return bid128BidCodecValue{lo: encoded}, nil
	case 128:
		encoded, err := encodeModeDiscOperand128(o.dec)
		if err != nil {
			return bid128BidCodecValue{}, err
		}
		return bid128BidCodecValue{
			lo: binary.LittleEndian.Uint64(encoded[0:8]),
			hi: binary.LittleEndian.Uint64(encoded[8:16]),
		}, nil
	default:
		return bid128BidCodecValue{}, fmt.Errorf("unsupported tier1 sentinel width %d", width)
	}
}

// tier1SentinelQNaNPayloadPair returns two quiet NaNs with distinct small
// payloads at the given width. Intel BID propagates the first NaN operand, so
// these pairs are the slot-swap discriminators for value-commutative
// operations; the selector measures the payload survival through the oracle
// instead of assuming it.
func tier1SentinelQNaNPayloadPair(width int) (tier1SentinelOperand, tier1SentinelOperand) {
	switch width {
	case 32:
		return sentinelRaw32(0x7c000001), sentinelRaw32(0x7c000002)
	case 64:
		return sentinelRaw64(0x7c00000000000001), sentinelRaw64(0x7c00000000000002)
	case 128:
		return sentinelRaw128(0x0000000000000001, 0x7c00000000000000),
			sentinelRaw128(0x0000000000000002, 0x7c00000000000000)
	default:
		panic(fmt.Sprintf("unsupported tier1 sentinel width %d", width))
	}
}

// ---------------------------------------------------------------------------
// Candidate tables (arithmetic domain)
// ---------------------------------------------------------------------------

type tier1SentinelBinaryTuple struct {
	x, y tier1SentinelOperand
}

type tier1SentinelTernaryTuple struct {
	x, y, z tier1SentinelOperand
}

type tier1SentinelScaleTuple struct {
	x tier1SentinelOperand
	n int64
}

// tier1SentinelTieCoefficient is the per-width odd coefficient whose ×5
// product and ÷2 quotient land exactly half-way between two representables
// with an even keep digit, so half-even and half-away must diverge.
func tier1SentinelTieCoefficient(width int) modeDiscOperand {
	switch width {
	case 32:
		return mdo(2469129, 0)
	case 64:
		return mdo(2469135802469129, 0)
	case 128:
		return mdoStr("2469135802469135802469135802469129", 0)
	default:
		panic(fmt.Sprintf("unsupported tier1 sentinel width %d", width))
	}
}

// tier1SentinelRoundedSupplements extends the reused public-parity
// mode-discriminant corpus with sentinel-specific candidates. Per operation:
//
//   - add/mul (value-commutative): a distinct-qNaN-payload pair is the only
//     way to make a slot swap visible (S1);
//   - sub: the discriminant corpus never separates half-even from half-away
//     (its only tie sits at a decade boundary where both round up) and never
//     drops a below-half fraction on a positive difference, so the
//     {nearest_even, nearest_away} and {nearest_away, toward_positive} pairs
//     need the …8.5 tie and the below-half candidates;
//   - mul: the corpus carries no exact half-way tie and no negative product,
//     so {nearest_even, nearest_away} and {toward_zero, toward_negative}
//     need the tie and negative-product candidates;
//   - div: same two gaps — an exactly-representable half quotient and a
//     negative non-terminating quotient close them.
//
// Candidate order matters (greedy adoption order is part of the canonical
// row order); primaries from the discriminant corpus always precede these.
var tier1SentinelRoundedSupplements = map[string]map[int][]tier1SentinelBinaryTuple{
	"add": {
		32:  {tier1SentinelPayloadBinary(32)},
		64:  {tier1SentinelPayloadBinary(64)},
		128: {tier1SentinelPayloadBinary(128)},
	},
	"sub": {
		32: {
			{x: sentinelDec(mdo(2469129, 0)), y: sentinelDec(mdo(5, -1))}, // 2469128.5: exact tie, keep digit even
			{x: sentinelDec(mdo(2, 0)), y: sentinelDec(mdo(9, -7))},       // 1.9999991: positive below-half drop
		},
		64: {
			{x: sentinelDec(mdo(2469135802469129, 0)), y: sentinelDec(mdo(5, -1))},
			{x: sentinelDec(mdo(2, 0)), y: sentinelDec(mdo(9, -16))},
		},
		128: {
			{x: sentinelDec(mdoStr("2469135802469135802469135802469129", 0)), y: sentinelDec(mdo(5, -1))},
			{x: sentinelDec(mdo(2, 0)), y: sentinelDec(mdo(9, -34))},
		},
	},
	"mul": {
		32: {
			{x: sentinelDec(tier1SentinelTieCoefficient(32)), y: sentinelDec(mdo(5, 0))}, // 12345645: exact tie, keep digit even
			{x: sentinelDec(mdoNeg(3333334, 0)), y: sentinelDec(mdo(3, 0))},              // -10000002: negative inexact product
			tier1SentinelPayloadBinary(32),
		},
		64: {
			{x: sentinelDec(tier1SentinelTieCoefficient(64)), y: sentinelDec(mdo(5, 0))},
			{x: sentinelDec(mdoNeg(3333333333333334, 0)), y: sentinelDec(mdo(3, 0))},
			tier1SentinelPayloadBinary(64),
		},
		128: {
			{x: sentinelDec(tier1SentinelTieCoefficient(128)), y: sentinelDec(mdo(5, 0))},
			{x: sentinelDec(mdoNeg(9999999999999999999, 0)), y: sentinelDec(mdo(9999999999999999999, 0))},
			tier1SentinelPayloadBinary(128),
		},
	},
	"div": {
		32: {
			{x: sentinelDec(tier1SentinelTieCoefficient(32)), y: sentinelDec(mdo(2, 0))}, // 1234564.5: exact tie, keep digit even
			{x: sentinelDec(mdoNeg(1, 0)), y: sentinelDec(mdo(3, 0))},                    // -0.333…: negative non-terminating
		},
		64: {
			{x: sentinelDec(tier1SentinelTieCoefficient(64)), y: sentinelDec(mdo(2, 0))},
			{x: sentinelDec(mdoNeg(1, 0)), y: sentinelDec(mdo(3, 0))},
		},
		128: {
			{x: sentinelDec(tier1SentinelTieCoefficient(128)), y: sentinelDec(mdo(2, 0))},
			{x: sentinelDec(mdoNeg(1, 0)), y: sentinelDec(mdo(3, 0))},
		},
	},
}

func tier1SentinelPayloadBinary(width int) tier1SentinelBinaryTuple {
	x, y := tier1SentinelQNaNPayloadPair(width)
	return tier1SentinelBinaryTuple{x: x, y: y}
}

// tier1SentinelUnroundedCandidates drive the modeless remainder/fmod family:
// (5,3) alone separates remainder (-1) from fmod (2) and is slot-asymmetric;
// the remaining entries are sign mirrors kept as fallback candidates.
var tier1SentinelUnroundedCandidates = []tier1SentinelBinaryTuple{
	{x: sentinelDec(mdo(5, 0)), y: sentinelDec(mdo(3, 0))},
	{x: sentinelDec(mdo(7, 0)), y: sentinelDec(mdo(4, 0))},
	{x: sentinelDec(mdoNeg(5, 0)), y: sentinelDec(mdo(3, 0))},
	{x: sentinelDec(mdo(5, 0)), y: sentinelDec(mdoNeg(3, 0))},
}

// tier1SentinelFmaSupplements: the ternary discriminant corpus never drops an
// above-half fraction on a positive fused result, so {nearest, toward_zero}
// needs the first candidate; and every corpus entry is x↔y-commutative (fused
// product), so the distinct-qNaN payload pair in the x/y slots is the S1
// discriminator for that pair (finite z keeps the z slot inert).
func tier1SentinelFmaSupplements(width int) []tier1SentinelTernaryTuple {
	var aboveHalf tier1SentinelTernaryTuple
	switch width {
	case 32:
		aboveHalf = tier1SentinelTernaryTuple{
			x: sentinelDec(mdo(2222223, 0)), y: sentinelDec(mdo(9, 0)), z: sentinelDec(mdo(0, 0)), // 20000007
		}
	case 64:
		aboveHalf = tier1SentinelTernaryTuple{
			x: sentinelDec(mdo(2222222222222223, 0)), y: sentinelDec(mdo(9, 0)), z: sentinelDec(mdo(0, 0)),
		}
	case 128:
		aboveHalf = tier1SentinelTernaryTuple{
			x: sentinelDec(mdo(3333333333333333333, 0)), y: sentinelDec(mdo(3333333333333333333, 0)), z: sentinelDec(mdo(0, 0)),
		}
	default:
		panic(fmt.Sprintf("unsupported tier1 sentinel width %d", width))
	}
	x, y := tier1SentinelQNaNPayloadPair(width)
	return []tier1SentinelTernaryTuple{
		aboveHalf,
		{x: x, y: y, z: sentinelDec(mdo(1, 0))},
	}
}

// tier1SentinelSqrtSupplements: the √10 operand drops an above-half fraction
// at width 128 (the √2/√3/√5/√7 primaries all drop below-half there), keeping
// the {nearest, toward_zero/toward_negative} pairs separable at every width.
var tier1SentinelSqrtSupplements = []tier1SentinelOperand{
	sentinelDec(mdo(10, 0)),
}

// tier1SentinelExceptions waives requirements that are structurally
// unsatisfiable, each with a written reason. Only rounding-mode pairs are
// waivable. Closed world: an entry whose requirement the full candidate pool
// can actually separate (at any width), or that matches no live requirement,
// fails generation.
type tier1SentinelException struct {
	op     string
	key    string // requirement key, e.g. "mode:0,4"
	reason string
}

var tier1SentinelExceptions = []tier1SentinelException{
	{
		op:  "sqrt",
		key: "mode:0,4",
		reason: "sqrt cannot produce an exact half-way result at any width: a tie needs a " +
			"(precision+1)-digit exact root ending in 5, whose square carries about twice the " +
			"digits and can never fit a valid input coefficient, so nearest_even and " +
			"nearest_away always agree",
	},
	{
		op:  "sqrt",
		key: "mode:3,1",
		reason: "sqrt results are never negative (a negative input yields the identical invalid " +
			"NaN in every mode, and -0 is exact), so toward_zero and toward_negative always agree",
	},
}

// ---------------------------------------------------------------------------
// Oracle (public bid754-go API → canonical "<bits>/<rawflags>" strings)
// ---------------------------------------------------------------------------

// tier1SentinelRawFlags mirrors the runners' tier1ArithmeticPublicRawFlags /
// public_raw_flags mapping onto the Intel raw flag bits. Any public flag
// outside the five Intel-visible bits fails generation.
func tier1SentinelRawFlags(flags bid754.ExceptionFlags) (uint32, error) {
	var raw uint32
	if flags&bid754.FlagInvalidOperation != 0 {
		raw |= 0x01
	}
	if flags&bid754.FlagDivisionByZero != 0 {
		raw |= 0x04
	}
	if flags&bid754.FlagOverflow != 0 {
		raw |= 0x08
	}
	if flags&bid754.FlagUnderflow != 0 {
		raw |= 0x10
	}
	if flags&bid754.FlagInexact != 0 {
		raw |= 0x20
	}
	known := bid754.FlagInvalidOperation | bid754.FlagDivisionByZero |
		bid754.FlagOverflow | bid754.FlagUnderflow | bid754.FlagInexact
	if unknown := flags &^ known; unknown != 0 {
		return 0, fmt.Errorf("tier1 sentinel oracle produced flags outside the Intel raw set: %s", unknown)
	}
	return raw, nil
}

func tier1SentinelDecimal128(v bid128BidCodecValue) bid754.Decimal128BID {
	var out bid754.Decimal128BID
	binary.LittleEndian.PutUint64(out[0:8], v.lo)
	binary.LittleEndian.PutUint64(out[8:16], v.hi)
	return out
}

func tier1SentinelValueText(width int, v bid128BidCodecValue) string {
	switch width {
	case 32:
		return fmt.Sprintf("%08x", uint32(v.lo))
	case 64:
		return fmt.Sprintf("%016x", v.lo)
	case 128:
		return fmt.Sprintf("%016x:%016x", v.hi, v.lo)
	default:
		panic(fmt.Sprintf("unsupported tier1 sentinel width %d", width))
	}
}

func tier1SentinelResult32(value bid754.Decimal32BID, flags bid754.ExceptionFlags) (string, error) {
	raw, err := tier1SentinelRawFlags(flags)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%08x/%08x", value.ToUint32(), raw), nil
}

func tier1SentinelResult64(value bid754.Decimal64BID, flags bid754.ExceptionFlags) (string, error) {
	raw, err := tier1SentinelRawFlags(flags)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%016x/%08x", value.ToUint64(), raw), nil
}

func tier1SentinelResult128(value bid754.Decimal128BID, flags bid754.ExceptionFlags) (string, error) {
	raw, err := tier1SentinelRawFlags(flags)
	if err != nil {
		return "", err
	}
	valueBytes := value.ToBytes()
	lo := binary.LittleEndian.Uint64(valueBytes[0:8])
	hi := binary.LittleEndian.Uint64(valueBytes[8:16])
	return fmt.Sprintf("%016x:%016x/%08x", hi, lo, raw), nil
}

func tier1SentinelEvalRounded(width int, op string, x, y bid128BidCodecValue, mode tier1SentinelMode) (string, error) {
	switch width {
	case 32:
		left, right := bid754.Decimal32BID(uint32(x.lo)), bid754.Decimal32BID(uint32(y.lo))
		var value bid754.Decimal32BID
		var flags bid754.ExceptionFlags
		switch op {
		case "add":
			value, flags = left.AddWithMode(right, mode.public)
		case "sub":
			value, flags = left.SubWithMode(right, mode.public)
		case "mul":
			value, flags = left.MulWithMode(right, mode.public)
		case "div":
			value, flags = left.DivWithMode(right, mode.public)
		case "quantize":
			value, flags = left.QuantizeWithMode(right, mode.public)
		default:
			return "", fmt.Errorf("unknown tier1 sentinel rounded operation %q", op)
		}
		return tier1SentinelResult32(value, flags)
	case 64:
		left, right := bid754.Decimal64BID(x.lo), bid754.Decimal64BID(y.lo)
		var value bid754.Decimal64BID
		var flags bid754.ExceptionFlags
		switch op {
		case "add":
			value, flags = left.AddWithMode(right, mode.public)
		case "sub":
			value, flags = left.SubWithMode(right, mode.public)
		case "mul":
			value, flags = left.MulWithMode(right, mode.public)
		case "div":
			value, flags = left.DivWithMode(right, mode.public)
		case "quantize":
			value, flags = left.QuantizeWithMode(right, mode.public)
		default:
			return "", fmt.Errorf("unknown tier1 sentinel rounded operation %q", op)
		}
		return tier1SentinelResult64(value, flags)
	case 128:
		left, right := tier1SentinelDecimal128(x), tier1SentinelDecimal128(y)
		var value bid754.Decimal128BID
		var flags bid754.ExceptionFlags
		switch op {
		case "add":
			value, flags = left.AddWithMode(right, mode.public)
		case "sub":
			value, flags = left.SubWithMode(right, mode.public)
		case "mul":
			value, flags = left.MulWithMode(right, mode.public)
		case "div":
			value, flags = left.DivWithMode(right, mode.public)
		case "quantize":
			value, flags = left.QuantizeWithMode(right, mode.public)
		default:
			return "", fmt.Errorf("unknown tier1 sentinel rounded operation %q", op)
		}
		return tier1SentinelResult128(value, flags)
	default:
		return "", fmt.Errorf("unsupported tier1 sentinel width %d", width)
	}
}

func tier1SentinelEvalUnrounded(width int, op string, x, y bid128BidCodecValue) (string, error) {
	switch width {
	case 32:
		left, right := bid754.Decimal32BID(uint32(x.lo)), bid754.Decimal32BID(uint32(y.lo))
		var value bid754.Decimal32BID
		var flags bid754.ExceptionFlags
		switch op {
		case "remainder":
			value, flags = left.Remainder(right)
		case "fmod":
			value, flags = left.Fmod(right)
		default:
			return "", fmt.Errorf("unknown tier1 sentinel unrounded operation %q", op)
		}
		return tier1SentinelResult32(value, flags)
	case 64:
		left, right := bid754.Decimal64BID(x.lo), bid754.Decimal64BID(y.lo)
		var value bid754.Decimal64BID
		var flags bid754.ExceptionFlags
		switch op {
		case "remainder":
			value, flags = left.Remainder(right)
		case "fmod":
			value, flags = left.Fmod(right)
		default:
			return "", fmt.Errorf("unknown tier1 sentinel unrounded operation %q", op)
		}
		return tier1SentinelResult64(value, flags)
	case 128:
		left, right := tier1SentinelDecimal128(x), tier1SentinelDecimal128(y)
		var value bid754.Decimal128BID
		var flags bid754.ExceptionFlags
		switch op {
		case "remainder":
			value, flags = left.Remainder(right)
		case "fmod":
			value, flags = left.Fmod(right)
		default:
			return "", fmt.Errorf("unknown tier1 sentinel unrounded operation %q", op)
		}
		return tier1SentinelResult128(value, flags)
	default:
		return "", fmt.Errorf("unsupported tier1 sentinel width %d", width)
	}
}

func tier1SentinelEvalFma(width int, x, y, z bid128BidCodecValue, mode tier1SentinelMode) (string, error) {
	switch width {
	case 32:
		value, flags := bid754.Decimal32BID(uint32(x.lo)).FMAWithMode(bid754.Decimal32BID(uint32(y.lo)), bid754.Decimal32BID(uint32(z.lo)), mode.public)
		return tier1SentinelResult32(value, flags)
	case 64:
		value, flags := bid754.Decimal64BID(x.lo).FMAWithMode(bid754.Decimal64BID(y.lo), bid754.Decimal64BID(z.lo), mode.public)
		return tier1SentinelResult64(value, flags)
	case 128:
		value, flags := tier1SentinelDecimal128(x).FMAWithMode(tier1SentinelDecimal128(y), tier1SentinelDecimal128(z), mode.public)
		return tier1SentinelResult128(value, flags)
	default:
		return "", fmt.Errorf("unsupported tier1 sentinel width %d", width)
	}
}

func tier1SentinelEvalSqrt(width int, x bid128BidCodecValue, mode tier1SentinelMode) (string, error) {
	switch width {
	case 32:
		value, flags := bid754.Decimal32BID(uint32(x.lo)).SqrtWithMode(mode.public)
		return tier1SentinelResult32(value, flags)
	case 64:
		value, flags := bid754.Decimal64BID(x.lo).SqrtWithMode(mode.public)
		return tier1SentinelResult64(value, flags)
	case 128:
		value, flags := tier1SentinelDecimal128(x).SqrtWithMode(mode.public)
		return tier1SentinelResult128(value, flags)
	default:
		return "", fmt.Errorf("unsupported tier1 sentinel width %d", width)
	}
}

func tier1SentinelEvalScale(width int, x bid128BidCodecValue, n int64, mode tier1SentinelMode) (string, error) {
	switch width {
	case 32:
		value, flags := bid754.Decimal32BID(uint32(x.lo)).ScaleBWithMode(int(n), mode.public)
		return tier1SentinelResult32(value, flags)
	case 64:
		value, flags := bid754.Decimal64BID(x.lo).ScaleBWithMode(int(n), mode.public)
		return tier1SentinelResult64(value, flags)
	case 128:
		value, flags := tier1SentinelDecimal128(x).ScaleBWithMode(int(n), mode.public)
		return tier1SentinelResult128(value, flags)
	default:
		return "", fmt.Errorf("unsupported tier1 sentinel width %d", width)
	}
}

// ---------------------------------------------------------------------------
// Requirement tracking, exceptions, and greedy selection
// ---------------------------------------------------------------------------

// tier1SentinelRequirements is one operation's unmet sensitivity-requirement
// set, keyed by stable requirement keys:
//
//	"slot:<a>,<b>"     operand slots a,b swap-distinguishable (S1)
//	"mode:<m1>,<m2>"   native rounding modes m1,m2 separated (S2)
//	"dispatch:<op>"    this operation separated from family sibling op (S3)
//	"sign:n"           scaleb integer-exponent sign visible (n vs -n)
type tier1SentinelRequirements struct {
	unmet map[string]bool
	order []string
}

func newTier1SentinelRequirements(keys []string) *tier1SentinelRequirements {
	unmet := make(map[string]bool, len(keys))
	for _, key := range keys {
		if unmet[key] {
			panic(fmt.Sprintf("tier1 sentinel: duplicate requirement key %q", key))
		}
		unmet[key] = true
	}
	return &tier1SentinelRequirements{unmet: unmet, order: append([]string(nil), keys...)}
}

func (r *tier1SentinelRequirements) drop(key string) {
	delete(r.unmet, key)
	for i, existing := range r.order {
		if existing == key {
			r.order = append(r.order[:i], r.order[i+1:]...)
			return
		}
	}
}

func (r *tier1SentinelRequirements) satisfy(key string) bool {
	if r.unmet[key] {
		delete(r.unmet, key)
		return true
	}
	return false
}

func (r *tier1SentinelRequirements) remaining() []string {
	out := make([]string, 0, len(r.unmet))
	for _, key := range r.order {
		if r.unmet[key] {
			out = append(out, key)
		}
	}
	return out
}

func tier1SentinelModePairKey(i, j int) string {
	return fmt.Sprintf("mode:%d,%d", tier1SentinelModes[i].native, tier1SentinelModes[j].native)
}

func tier1SentinelModePairKeys() []string {
	keys := make([]string, 0, 10)
	for i := 0; i < len(tier1SentinelModes); i++ {
		for j := i + 1; j < len(tier1SentinelModes); j++ {
			keys = append(keys, tier1SentinelModePairKey(i, j))
		}
	}
	return keys
}

// tier1SentinelSatisfyModePairs marks every unmet mode-pair requirement the
// per-mode result set separates and returns the number newly satisfied.
func tier1SentinelSatisfyModePairs(reqs *tier1SentinelRequirements, results tier1SentinelModeResults) int {
	gain := 0
	for i := 0; i < len(tier1SentinelModes); i++ {
		for j := i + 1; j < len(tier1SentinelModes); j++ {
			key := tier1SentinelModePairKey(i, j)
			if reqs.unmet[key] && results[i] != results[j] && reqs.satisfy(key) {
				gain++
			}
		}
	}
	return gain
}

// tier1SentinelApplyExceptions removes waived requirement keys for the
// operation before selection and returns the removed keys so the caller can
// verify (against the full candidate pool) that every waiver was genuinely
// unsatisfiable. Only mode-pair requirements are waivable.
func tier1SentinelApplyExceptions(op string, reqs *tier1SentinelRequirements) ([]string, error) {
	removed := []string{}
	for _, exception := range tier1SentinelExceptions {
		if exception.op != op {
			continue
		}
		if !strings.HasPrefix(exception.key, "mode:") {
			return nil, fmt.Errorf("tier1 sentinel exception (op %q, %s): only mode-pair requirements are waivable", exception.op, exception.key)
		}
		if !reqs.unmet[exception.key] {
			continue
		}
		reqs.drop(exception.key)
		removed = append(removed, exception.key)
	}
	return removed, nil
}

// tier1SentinelExceptionUse records, per exception table entry, at how many
// (op, width) selection runs the waiver was genuinely needed. Closed world:
// GenerateTier1ArithmeticSentinelRows fails when an entry was needed at
// fewer than all widths (stale, mis-keyed, or wrongly-scoped waiver).
type tier1SentinelExceptionUse map[int]int

func (u tier1SentinelExceptionUse) requireAllUsed(widthsPerOp int) error {
	for i, exception := range tier1SentinelExceptions {
		if u[i] != widthsPerOp {
			return fmt.Errorf("tier1 sentinel exception %d (op %q, %s) was needed at %d/%d widths; remove or rescope the stale waiver",
				i, exception.op, exception.key, u[i], widthsPerOp)
		}
	}
	return nil
}

// tier1SentinelVerifyModeWaivers checks each waived mode-pair against the
// FULL candidate pool's per-mode results: a waiver the pool could actually
// satisfy is stale and fails generation; a genuinely needed waiver is
// recorded in the closed-world use census.
func tier1SentinelVerifyModeWaivers(op string, width int, waived []string, poolResults []tier1SentinelModeResults, use tier1SentinelExceptionUse) error {
	for _, key := range waived {
		probe := newTier1SentinelRequirements([]string{key})
		for _, results := range poolResults {
			tier1SentinelSatisfyModePairs(probe, results)
			if len(probe.unmet) == 0 {
				break
			}
		}
		if len(probe.unmet) == 0 {
			return fmt.Errorf("tier1 sentinel exception for %s %s %s is stale: the candidate pool can separate the waived mode pair",
				tier1SentinelWidthLabel(width), op, key)
		}
		for i, exception := range tier1SentinelExceptions {
			if exception.op == op && exception.key == key {
				use[i]++
			}
		}
	}
	return nil
}

func tier1SentinelSelectionFailure(width int, op string, remaining []string) error {
	return fmt.Errorf("tier1 sentinel selection failed for %s %s: candidate pool cannot satisfy %s; extend the sentinel candidate tables (no fallback, no partial output)",
		tier1SentinelWidthLabel(width), op, strings.Join(remaining, " "))
}

func tier1SentinelCapFailure(width int, op string, adopted, cap int) error {
	return fmt.Errorf("tier1 sentinel selection for %s %s adopted %d tuples, above the per-operation cap %d; improve the candidate table instead of raising the cap",
		tier1SentinelWidthLabel(width), op, adopted, cap)
}

func tier1SentinelRequireNoWaivers(width int, op string, waived []string) error {
	if len(waived) > 0 {
		return fmt.Errorf("tier1 sentinel exception declared for %s %s but that family defines no waivable requirement", tier1SentinelWidthLabel(width), op)
	}
	return nil
}

var tier1SentinelRoundedOps = [...]string{"add", "sub", "mul", "div", "quantize"}
var tier1SentinelUnroundedOps = [...]string{"remainder", "fmod"}

func tier1SentinelRoundedTitle(op string) string {
	switch op {
	case "add":
		return "Add"
	case "sub":
		return "Sub"
	case "mul":
		return "Mul"
	case "div":
		return "Div"
	case "quantize":
		return "Quantize"
	default:
		panic("unknown tier1 sentinel rounded operation: " + op)
	}
}

func tier1SentinelRoundedCandidates(op string, width int) ([]tier1SentinelBinaryTuple, error) {
	pairs, err := modeBinaryDiscriminantOperands(tier1SentinelRoundedTitle(op), width)
	if err != nil {
		return nil, fmt.Errorf("tier1 sentinel rounded candidates: %w", err)
	}
	out := make([]tier1SentinelBinaryTuple, 0, len(pairs)+3)
	for _, pair := range pairs {
		out = append(out, tier1SentinelBinaryTuple{x: sentinelDec(pair[0]), y: sentinelDec(pair[1])})
	}
	out = append(out, tier1SentinelRoundedSupplements[op][width]...)
	return out, nil
}

// tier1SentinelSelectRounded runs the greedy selection for one rounded binary
// operation at one width and returns the adopted tuples in adoption order.
func tier1SentinelSelectRounded(op string, width int, exceptions tier1SentinelExceptionUse) ([]tier1SentinelBinaryTuple, error) {
	keys := []string{"slot:x,y"}
	keys = append(keys, tier1SentinelModePairKeys()...)
	for _, other := range tier1SentinelRoundedOps {
		if other != op {
			keys = append(keys, "dispatch:"+other)
		}
	}
	reqs := newTier1SentinelRequirements(keys)
	waived, err := tier1SentinelApplyExceptions(op, reqs)
	if err != nil {
		return nil, err
	}

	candidates, err := tier1SentinelRoundedCandidates(op, width)
	if err != nil {
		return nil, err
	}
	evalAll := func(evalOp string, x, y bid128BidCodecValue) (tier1SentinelModeResults, error) {
		var out tier1SentinelModeResults
		for i, mode := range tier1SentinelModes {
			result, err := tier1SentinelEvalRounded(width, evalOp, x, y, mode)
			if err != nil {
				return out, err
			}
			out[i] = result
		}
		return out, nil
	}
	adopted := []tier1SentinelBinaryTuple{}
	poolResults := []tier1SentinelModeResults{}
	for _, cand := range candidates {
		x, err := cand.x.encode(width)
		if err != nil {
			return nil, err
		}
		y, err := cand.y.encode(width)
		if err != nil {
			return nil, err
		}
		results, err := evalAll(op, x, y)
		if err != nil {
			return nil, err
		}
		poolResults = append(poolResults, results)
		if len(reqs.unmet) == 0 {
			continue
		}
		gain := tier1SentinelSatisfyModePairs(reqs, results)
		if reqs.unmet["slot:x,y"] {
			swapped, err := evalAll(op, y, x)
			if err != nil {
				return nil, err
			}
			if swapped != results && reqs.satisfy("slot:x,y") {
				gain++
			}
		}
		for _, other := range tier1SentinelRoundedOps {
			if other == op {
				continue
			}
			key := "dispatch:" + other
			if !reqs.unmet[key] {
				continue
			}
			otherResults, err := evalAll(other, x, y)
			if err != nil {
				return nil, err
			}
			if otherResults != results && reqs.satisfy(key) {
				gain++
			}
		}
		if gain > 0 {
			adopted = append(adopted, cand)
		}
	}
	if remaining := reqs.remaining(); len(remaining) > 0 {
		return nil, tier1SentinelSelectionFailure(width, op, remaining)
	}
	if len(adopted) > tier1SentinelTupleCap {
		return nil, tier1SentinelCapFailure(width, op, len(adopted), tier1SentinelTupleCap)
	}
	if err := tier1SentinelVerifyModeWaivers(op, width, waived, poolResults, exceptions); err != nil {
		return nil, err
	}
	return adopted, nil
}

// tier1SentinelSelectUnrounded covers the modeless remainder/fmod family:
// S1 (slot swap) plus S3 against the sibling operation.
func tier1SentinelSelectUnrounded(op string, width int, exceptions tier1SentinelExceptionUse) ([]tier1SentinelBinaryTuple, error) {
	_ = exceptions
	keys := []string{"slot:x,y"}
	for _, other := range tier1SentinelUnroundedOps {
		if other != op {
			keys = append(keys, "dispatch:"+other)
		}
	}
	reqs := newTier1SentinelRequirements(keys)
	waived, err := tier1SentinelApplyExceptions(op, reqs)
	if err != nil {
		return nil, err
	}
	if err := tier1SentinelRequireNoWaivers(width, op, waived); err != nil {
		return nil, err
	}
	adopted := []tier1SentinelBinaryTuple{}
	for _, cand := range tier1SentinelUnroundedCandidates {
		if len(reqs.unmet) == 0 {
			break
		}
		x, err := cand.x.encode(width)
		if err != nil {
			return nil, err
		}
		y, err := cand.y.encode(width)
		if err != nil {
			return nil, err
		}
		result, err := tier1SentinelEvalUnrounded(width, op, x, y)
		if err != nil {
			return nil, err
		}
		gain := 0
		if reqs.unmet["slot:x,y"] {
			swapped, err := tier1SentinelEvalUnrounded(width, op, y, x)
			if err != nil {
				return nil, err
			}
			if swapped != result && reqs.satisfy("slot:x,y") {
				gain++
			}
		}
		for _, other := range tier1SentinelUnroundedOps {
			if other == op {
				continue
			}
			key := "dispatch:" + other
			if !reqs.unmet[key] {
				continue
			}
			otherResult, err := tier1SentinelEvalUnrounded(width, other, x, y)
			if err != nil {
				return nil, err
			}
			if otherResult != result && reqs.satisfy(key) {
				gain++
			}
		}
		if gain > 0 {
			adopted = append(adopted, cand)
		}
	}
	if remaining := reqs.remaining(); len(remaining) > 0 {
		return nil, tier1SentinelSelectionFailure(width, op, remaining)
	}
	if len(adopted) > tier1SentinelTupleCap {
		return nil, tier1SentinelCapFailure(width, op, len(adopted), tier1SentinelTupleCap)
	}
	return adopted, nil
}

// tier1SentinelSelectFma covers the ternary fma family: three slot pairs
// (x↔y through the payload tuple — the fused product is value-commutative),
// the ten mode pairs, and no S3 (single-operation family).
func tier1SentinelSelectFma(width int, exceptions tier1SentinelExceptionUse) ([]tier1SentinelTernaryTuple, error) {
	_ = exceptions
	const op = "fma"
	keys := []string{"slot:x,y", "slot:y,z", "slot:x,z"}
	keys = append(keys, tier1SentinelModePairKeys()...)
	reqs := newTier1SentinelRequirements(keys)
	waived, err := tier1SentinelApplyExceptions(op, reqs)
	if err != nil {
		return nil, err
	}
	if err := tier1SentinelRequireNoWaivers(width, op, waived); err != nil {
		return nil, err
	}

	primaries, err := modeTernaryDiscriminantOperands("FMA", width)
	if err != nil {
		return nil, fmt.Errorf("tier1 sentinel fma candidates: %w", err)
	}
	candidates := make([]tier1SentinelTernaryTuple, 0, len(primaries)+2)
	for _, triple := range primaries {
		candidates = append(candidates, tier1SentinelTernaryTuple{
			x: sentinelDec(triple[0]), y: sentinelDec(triple[1]), z: sentinelDec(triple[2]),
		})
	}
	candidates = append(candidates, tier1SentinelFmaSupplements(width)...)

	evalAll := func(x, y, z bid128BidCodecValue) (tier1SentinelModeResults, error) {
		var out tier1SentinelModeResults
		for i, mode := range tier1SentinelModes {
			result, err := tier1SentinelEvalFma(width, x, y, z, mode)
			if err != nil {
				return out, err
			}
			out[i] = result
		}
		return out, nil
	}
	adopted := []tier1SentinelTernaryTuple{}
	for _, cand := range candidates {
		if len(reqs.unmet) == 0 {
			break
		}
		x, err := cand.x.encode(width)
		if err != nil {
			return nil, err
		}
		y, err := cand.y.encode(width)
		if err != nil {
			return nil, err
		}
		z, err := cand.z.encode(width)
		if err != nil {
			return nil, err
		}
		results, err := evalAll(x, y, z)
		if err != nil {
			return nil, err
		}
		gain := tier1SentinelSatisfyModePairs(reqs, results)
		for _, slot := range []struct {
			key     string
			a, b, c bid128BidCodecValue
		}{
			{key: "slot:x,y", a: y, b: x, c: z},
			{key: "slot:y,z", a: x, b: z, c: y},
			{key: "slot:x,z", a: z, b: y, c: x},
		} {
			if !reqs.unmet[slot.key] {
				continue
			}
			swapped, err := evalAll(slot.a, slot.b, slot.c)
			if err != nil {
				return nil, err
			}
			if swapped != results && reqs.satisfy(slot.key) {
				gain++
			}
		}
		if gain > 0 {
			adopted = append(adopted, cand)
		}
	}
	if remaining := reqs.remaining(); len(remaining) > 0 {
		return nil, tier1SentinelSelectionFailure(width, op, remaining)
	}
	if len(adopted) > tier1SentinelTupleCap {
		return nil, tier1SentinelCapFailure(width, op, len(adopted), tier1SentinelTupleCap)
	}
	return adopted, nil
}

// tier1SentinelSelectSqrt covers the unary sqrt family: mode pairs only
// (single slot, single-operation family). {nearest_even, nearest_away} and
// {toward_zero, toward_negative} are waived by the reasoned exceptions above
// — sqrt cannot tie and never yields a rounded negative result.
func tier1SentinelSelectSqrt(width int, exceptions tier1SentinelExceptionUse) ([]tier1SentinelOperand, error) {
	const op = "sqrt"
	reqs := newTier1SentinelRequirements(tier1SentinelModePairKeys())
	waived, err := tier1SentinelApplyExceptions(op, reqs)
	if err != nil {
		return nil, err
	}

	primaries, err := modeUnaryDiscriminantOperands("Sqrt", width)
	if err != nil {
		return nil, fmt.Errorf("tier1 sentinel sqrt candidates: %w", err)
	}
	candidates := make([]tier1SentinelOperand, 0, len(primaries)+len(tier1SentinelSqrtSupplements))
	for _, operand := range primaries {
		candidates = append(candidates, sentinelDec(operand))
	}
	candidates = append(candidates, tier1SentinelSqrtSupplements...)

	adopted := []tier1SentinelOperand{}
	poolResults := []tier1SentinelModeResults{}
	for _, cand := range candidates {
		x, err := cand.encode(width)
		if err != nil {
			return nil, err
		}
		var results tier1SentinelModeResults
		for i, mode := range tier1SentinelModes {
			result, err := tier1SentinelEvalSqrt(width, x, mode)
			if err != nil {
				return nil, err
			}
			results[i] = result
		}
		poolResults = append(poolResults, results)
		if len(reqs.unmet) == 0 {
			continue
		}
		if tier1SentinelSatisfyModePairs(reqs, results) > 0 {
			adopted = append(adopted, cand)
		}
	}
	if remaining := reqs.remaining(); len(remaining) > 0 {
		return nil, tier1SentinelSelectionFailure(width, op, remaining)
	}
	if len(adopted) > tier1SentinelTupleCap {
		return nil, tier1SentinelCapFailure(width, op, len(adopted), tier1SentinelTupleCap)
	}
	if err := tier1SentinelVerifyModeWaivers(op, width, waived, poolResults, exceptions); err != nil {
		return nil, err
	}
	return adopted, nil
}

// tier1SentinelSelectScale covers scaleb: the integer exponent slot cannot be
// swapped with the decimal slot (type-distinct), so its S1 analogue is the
// exponent-sign requirement (n vs -n visible), plus the ten mode pairs.
func tier1SentinelSelectScale(width int, exceptions tier1SentinelExceptionUse) ([]tier1SentinelScaleTuple, error) {
	_ = exceptions
	const op = "scaleb"
	keys := []string{"sign:n"}
	keys = append(keys, tier1SentinelModePairKeys()...)
	reqs := newTier1SentinelRequirements(keys)
	waived, err := tier1SentinelApplyExceptions(op, reqs)
	if err != nil {
		return nil, err
	}
	if err := tier1SentinelRequireNoWaivers(width, op, waived); err != nil {
		return nil, err
	}

	primaries, err := modeScaleBDiscriminantCases("ScaleB", width)
	if err != nil {
		return nil, fmt.Errorf("tier1 sentinel scaleb candidates: %w", err)
	}
	candidates := make([]tier1SentinelScaleTuple, 0, len(primaries))
	for _, tc := range primaries {
		candidates = append(candidates, tier1SentinelScaleTuple{x: sentinelDec(tc.Val), n: int64(tc.Exp)})
	}

	adopted := []tier1SentinelScaleTuple{}
	for _, cand := range candidates {
		if len(reqs.unmet) == 0 {
			break
		}
		x, err := cand.x.encode(width)
		if err != nil {
			return nil, err
		}
		var results tier1SentinelModeResults
		for i, mode := range tier1SentinelModes {
			result, err := tier1SentinelEvalScale(width, x, cand.n, mode)
			if err != nil {
				return nil, err
			}
			results[i] = result
		}
		gain := tier1SentinelSatisfyModePairs(reqs, results)
		if reqs.unmet["sign:n"] && cand.n != 0 {
			signVisible := false
			for i, mode := range tier1SentinelModes {
				negated, err := tier1SentinelEvalScale(width, x, -cand.n, mode)
				if err != nil {
					return nil, err
				}
				if negated != results[i] {
					signVisible = true
					break
				}
			}
			if signVisible && reqs.satisfy("sign:n") {
				gain++
			}
		}
		if gain > 0 {
			adopted = append(adopted, cand)
		}
	}
	if remaining := reqs.remaining(); len(remaining) > 0 {
		return nil, tier1SentinelSelectionFailure(width, op, remaining)
	}
	if len(adopted) > tier1SentinelTupleCap {
		return nil, tier1SentinelCapFailure(width, op, len(adopted), tier1SentinelTupleCap)
	}
	return adopted, nil
}

// ---------------------------------------------------------------------------
// Row emission
// ---------------------------------------------------------------------------

// tier1SentinelRow is one canonical sentinel row plus its human-audit
// interpretation for -print-sentinel-anchors.
type tier1SentinelRow struct {
	text    string
	comment string
}

// tier1SentinelAssertRowText enforces the canonical row charter at emit time:
// printable ASCII, no double quote, no backslash, single-space separation
// (the same string must embed verbatim into Go, Rust, and JSON literals).
func tier1SentinelAssertRowText(text string) error {
	for _, ch := range text {
		if ch < 0x20 || ch > 0x7e || ch == '"' || ch == '\\' {
			return fmt.Errorf("tier1 sentinel row %q contains a character outside the canonical ASCII charter", text)
		}
	}
	if strings.Contains(text, "  ") || strings.HasPrefix(text, " ") || strings.HasSuffix(text, " ") {
		return fmt.Errorf("tier1 sentinel row %q violates single-space field separation", text)
	}
	return nil
}

type tier1SentinelFlagName struct {
	bit  uint32
	name string
}

var tier1SentinelFlagNames = [...]tier1SentinelFlagName{
	{bit: 0x01, name: "invalid"},
	{bit: 0x04, name: "divzero"},
	{bit: 0x08, name: "overflow"},
	{bit: 0x10, name: "underflow"},
	{bit: 0x20, name: "inexact"},
}

func tier1SentinelFlagsComment(result string) string {
	slash := strings.LastIndexByte(result, '/')
	if slash < 0 {
		return "?"
	}
	var raw uint32
	if _, err := fmt.Sscanf(result[slash+1:], "%08x", &raw); err != nil {
		return "?"
	}
	if raw == 0 {
		return "none"
	}
	names := []string{}
	for _, flag := range tier1SentinelFlagNames {
		if raw&flag.bit != 0 {
			names = append(names, flag.name)
		}
	}
	return strings.Join(names, "|")
}

func tier1SentinelDecimalComment(width int, v bid128BidCodecValue) string {
	switch width {
	case 32:
		return bid754.Decimal32BID(uint32(v.lo)).String()
	case 64:
		return bid754.Decimal64BID(v.lo).String()
	case 128:
		return tier1SentinelDecimal128(v).String()
	default:
		return "?"
	}
}

func tier1SentinelResultComment(width int, result string) string {
	slash := strings.LastIndexByte(result, '/')
	if slash < 0 {
		return "?"
	}
	bitsText := result[:slash]
	switch width {
	case 32:
		var value uint32
		if _, err := fmt.Sscanf(bitsText, "%08x", &value); err != nil {
			return "?"
		}
		return bid754.Decimal32BID(value).String()
	case 64:
		var value uint64
		if _, err := fmt.Sscanf(bitsText, "%016x", &value); err != nil {
			return "?"
		}
		return bid754.Decimal64BID(value).String()
	case 128:
		var hi, lo uint64
		if _, err := fmt.Sscanf(bitsText, "%016x:%016x", &hi, &lo); err != nil {
			return "?"
		}
		return tier1SentinelDecimal128(bid128BidCodecValue{lo: lo, hi: hi}).String()
	default:
		return "?"
	}
}

func tier1SentinelAppendRow(rows []tier1SentinelRow, text, comment string) ([]tier1SentinelRow, error) {
	if err := tier1SentinelAssertRowText(text); err != nil {
		return nil, err
	}
	return append(rows, tier1SentinelRow{text: text, comment: comment}), nil
}

// GenerateTier1ArithmeticSentinelRows selects and serializes the arithmetic
// routing-sentinel rows in canonical order: width ascending → family
// declaration order (rounded, unrounded, fma, sqrt, scaleb) → operation
// declaration order → tuple adoption order → native mode table order.
func GenerateTier1ArithmeticSentinelRows() ([]tier1SentinelRow, error) {
	exceptionUse := tier1SentinelExceptionUse{}
	rows := []tier1SentinelRow{}
	var err error
	for _, width := range tier1SentinelWidths {
		label := tier1SentinelWidthLabel(width)
		for _, op := range tier1SentinelRoundedOps {
			adopted, selErr := tier1SentinelSelectRounded(op, width, exceptionUse)
			if selErr != nil {
				return nil, selErr
			}
			for _, tuple := range adopted {
				x, encErr := tuple.x.encode(width)
				if encErr != nil {
					return nil, encErr
				}
				y, encErr := tuple.y.encode(width)
				if encErr != nil {
					return nil, encErr
				}
				for _, mode := range tier1SentinelModes {
					result, evalErr := tier1SentinelEvalRounded(width, op, x, y, mode)
					if evalErr != nil {
						return nil, evalErr
					}
					text := fmt.Sprintf("%s %s x=%s y=%s m=%d -> %s",
						label, op, tier1SentinelValueText(width, x), tier1SentinelValueText(width, y), mode.native, result)
					comment := fmt.Sprintf("%s %s %s , %s [%s] = %s , flags %s",
						label, op, tier1SentinelDecimalComment(width, x), tier1SentinelDecimalComment(width, y),
						mode.name, tier1SentinelResultComment(width, result), tier1SentinelFlagsComment(result))
					if rows, err = tier1SentinelAppendRow(rows, text, comment); err != nil {
						return nil, err
					}
				}
			}
		}
		for _, op := range tier1SentinelUnroundedOps {
			adopted, selErr := tier1SentinelSelectUnrounded(op, width, exceptionUse)
			if selErr != nil {
				return nil, selErr
			}
			for _, tuple := range adopted {
				x, encErr := tuple.x.encode(width)
				if encErr != nil {
					return nil, encErr
				}
				y, encErr := tuple.y.encode(width)
				if encErr != nil {
					return nil, encErr
				}
				result, evalErr := tier1SentinelEvalUnrounded(width, op, x, y)
				if evalErr != nil {
					return nil, evalErr
				}
				text := fmt.Sprintf("%s %s x=%s y=%s -> %s",
					label, op, tier1SentinelValueText(width, x), tier1SentinelValueText(width, y), result)
				comment := fmt.Sprintf("%s %s %s , %s = %s , flags %s",
					label, op, tier1SentinelDecimalComment(width, x), tier1SentinelDecimalComment(width, y),
					tier1SentinelResultComment(width, result), tier1SentinelFlagsComment(result))
				if rows, err = tier1SentinelAppendRow(rows, text, comment); err != nil {
					return nil, err
				}
			}
		}
		fmaAdopted, selErr := tier1SentinelSelectFma(width, exceptionUse)
		if selErr != nil {
			return nil, selErr
		}
		for _, tuple := range fmaAdopted {
			x, encErr := tuple.x.encode(width)
			if encErr != nil {
				return nil, encErr
			}
			y, encErr := tuple.y.encode(width)
			if encErr != nil {
				return nil, encErr
			}
			z, encErr := tuple.z.encode(width)
			if encErr != nil {
				return nil, encErr
			}
			for _, mode := range tier1SentinelModes {
				result, evalErr := tier1SentinelEvalFma(width, x, y, z, mode)
				if evalErr != nil {
					return nil, evalErr
				}
				text := fmt.Sprintf("%s fma x=%s y=%s z=%s m=%d -> %s",
					label, tier1SentinelValueText(width, x), tier1SentinelValueText(width, y), tier1SentinelValueText(width, z), mode.native, result)
				comment := fmt.Sprintf("%s fma %s * %s + %s [%s] = %s , flags %s",
					label, tier1SentinelDecimalComment(width, x), tier1SentinelDecimalComment(width, y), tier1SentinelDecimalComment(width, z),
					mode.name, tier1SentinelResultComment(width, result), tier1SentinelFlagsComment(result))
				if rows, err = tier1SentinelAppendRow(rows, text, comment); err != nil {
					return nil, err
				}
			}
		}
		sqrtAdopted, selErr := tier1SentinelSelectSqrt(width, exceptionUse)
		if selErr != nil {
			return nil, selErr
		}
		for _, operand := range sqrtAdopted {
			x, encErr := operand.encode(width)
			if encErr != nil {
				return nil, encErr
			}
			for _, mode := range tier1SentinelModes {
				result, evalErr := tier1SentinelEvalSqrt(width, x, mode)
				if evalErr != nil {
					return nil, evalErr
				}
				text := fmt.Sprintf("%s sqrt x=%s m=%d -> %s",
					label, tier1SentinelValueText(width, x), mode.native, result)
				comment := fmt.Sprintf("%s sqrt %s [%s] = %s , flags %s",
					label, tier1SentinelDecimalComment(width, x), mode.name,
					tier1SentinelResultComment(width, result), tier1SentinelFlagsComment(result))
				if rows, err = tier1SentinelAppendRow(rows, text, comment); err != nil {
					return nil, err
				}
			}
		}
		scaleAdopted, selErr := tier1SentinelSelectScale(width, exceptionUse)
		if selErr != nil {
			return nil, selErr
		}
		for _, tuple := range scaleAdopted {
			x, encErr := tuple.x.encode(width)
			if encErr != nil {
				return nil, encErr
			}
			for _, mode := range tier1SentinelModes {
				result, evalErr := tier1SentinelEvalScale(width, x, tuple.n, mode)
				if evalErr != nil {
					return nil, evalErr
				}
				text := fmt.Sprintf("%s scaleb x=%s n=%d m=%d -> %s",
					label, tier1SentinelValueText(width, x), tuple.n, mode.native, result)
				comment := fmt.Sprintf("%s scaleb %s scaled by 10^%d [%s] = %s , flags %s",
					label, tier1SentinelDecimalComment(width, x), tuple.n, mode.name,
					tier1SentinelResultComment(width, result), tier1SentinelFlagsComment(result))
				if rows, err = tier1SentinelAppendRow(rows, text, comment); err != nil {
					return nil, err
				}
			}
		}
	}
	if err := exceptionUse.requireAllUsed(len(tier1SentinelWidths)); err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("tier1 sentinel selection produced zero rows")
	}
	seen := map[string]bool{}
	for _, row := range rows {
		if seen[row.text] {
			return nil, fmt.Errorf("tier1 sentinel selection produced duplicate row %q", row.text)
		}
		seen[row.text] = true
	}
	return rows, nil
}

func tier1SentinelRowTexts(rows []tier1SentinelRow) []string {
	texts := make([]string, len(rows))
	for i, row := range rows {
		texts[i] = row.text
	}
	return texts
}

// tier1SentinelGoRowLiterals renders the canonical rows as one-per-line Go
// string-slice elements (the anchor test re-extracts them via AST and
// requires byte equality with the hand-pinned JSON rows).
func tier1SentinelGoRowLiterals(rows []tier1SentinelRow) string {
	var out strings.Builder
	for _, row := range rows {
		fmt.Fprintf(&out, "\t%q,\n", row.text)
	}
	return out.String()
}

// tier1SentinelRustRowLiterals renders the same rows as one-per-line Rust
// &str array elements.
func tier1SentinelRustRowLiterals(rows []tier1SentinelRow) string {
	var out strings.Builder
	for _, row := range rows {
		fmt.Fprintf(&out, "    %q,\n", row.text)
	}
	return out.String()
}

// Tier1SentinelAnchorProposal renders the stdout block for
// `cmd/testgen -print-sentinel-anchors`: the proposed
// devtools/verification_sentinels.json row arrays plus a per-row decimal
// interpretation for the human audit. It writes no file and reads no anchor
// (GUARDRAILS: generators never touch the pin files).
func Tier1SentinelAnchorProposal() (string, error) {
	rows, err := GenerateTier1ArithmeticSentinelRows()
	if err != nil {
		return "", err
	}
	ccRows, err := GenerateTier1CompareConversionSentinelRows()
	if err != nil {
		return "", err
	}
	proposal := struct {
		Arithmetic        []string `json:"tier1_arithmetic_long_routing_sentinel_rows"`
		CompareConversion []string `json:"tier1_compare_conversion_long_routing_sentinel_rows"`
	}{Arithmetic: tier1SentinelRowTexts(rows), CompareConversion: tier1SentinelRowTexts(ccRows)}
	var encoded strings.Builder
	encoder := json.NewEncoder(&encoded)
	encoder.SetEscapeHTML(false) // keep the "->" row arrow literal for the human audit
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(proposal); err != nil {
		return "", err
	}
	var out strings.Builder
	out.WriteString("Proposed devtools/verification_sentinels.json rows (audit each row, then paste manually — no generator writes that file):\n")
	out.WriteString(encoded.String())
	out.WriteString("\nRow interpretations (audit aid, not pinned data):\n")
	for _, row := range rows {
		fmt.Fprintf(&out, "# %s\n", row.comment)
	}
	for _, row := range ccRows {
		fmt.Fprintf(&out, "# %s\n", row.comment)
	}
	return out.String(), nil
}
