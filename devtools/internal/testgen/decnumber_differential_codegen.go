package testgen

// decNumber third-oracle differential gate codegen.
//
// Generates the batch corpus, the independent BID triple codec, the 3-leg
// runner (pinned Intel BID C / Go mechanical port / pinned IBM decNumber
// 3.68), and the closed-world exclusion inventory for the
// `decnumber_differential` gate. decNumber is a divergence tripwire here,
// never a correctness definition (docs/SPEC.md Non-Goals): comparison is
// exact over the pre-declared agreement region, exclusions are
// generation-time class predicates counted in the generated inventory, and
// the runner carries no tolerance, no runtime heuristic, and no runtime
// skip other than the generation-time class predicates it replays.
//
// Corpus sources:
//  1. the Tier 1 arithmetic boundary sets, reused at function level and
//     filtered to the gate's operand contract (canonical-only: exclusion
//     class noncanonical_operand_no_gda_counterpart; NaN payload zero only:
//     exclusion class nan_payload_operand_zero_only),
//  2. the Tier 1 probe sets with NaN payloads normalized to zero,
//  3. a manifest-owned exact-product overflow class (1E<p±k> and ±Nmax/±Nmin
//     crossed through mul/fma; closes the corpus gap that hid the
//     Bid128Fma overflow defect),
//  4. deterministic seeded random triples generated in canonical triple
//     space (tier1 splitmix64 pattern), stream-hash pinned in the runner
//     constants and re-pinned by hand in devtools/verification_anchors.json.
//
// GUARDRAILS: this generator never reads or writes
// devtools/verification_anchors.json or devtools/verification_sentinels.json.
// Sentinel rows land only in the generated runner; the human pin flows
// through `cmd/testgen -print-decnumber-sentinel-anchors` stdout and a
// manual paste audited by TestVerificationAnchorsMatchGeneratedArtifacts.

import (
	"embed"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"

	"github.com/sky1core/bid754/devtools/internal/genmarker"
)

const (
	decnumberDiffSharedGeneratedPath     = "../bid754-go/generated_decnumber_differential_shared_test.go"
	decnumberDiffNativeShimGeneratedPath = "../bid754-go/generated_decnumber_differential_native.go"
	decnumberDiffRunnerGeneratedPath     = "../bid754-go/generated_decnumber_differential_native_test.go"
	decnumberDiffStubGeneratedPath       = "../bid754-go/generated_decnumber_differential_stub_test.go"

	decnumberDiffHashOffset = uint64(0xcbf29ce484222325)
	decnumberDiffHashPrime  = uint64(0x100000001b3)
)

//go:embed decnumber_differential_templates/*
var decnumberDiffTemplates embed.FS

// Exclusion class reason ids (design ledger L13/L11/L16/L7/L12). The ids are
// emitted verbatim into the generated inventory and the runner log lines.
const (
	decnumberDiffClassFmaZeroInfQnan = "fma_zero_inf_qnan_invalid_ieee_optional"
	decnumberDiffClassSqrtNonRN      = "sqrt_gda_fixed_half_even_rounding"
	decnumberDiffClassNonCanonical   = "noncanonical_operand_no_gda_counterpart"
	decnumberDiffClassNaNPayload     = "nan_payload_operand_zero_only"
	decnumberDiffClassRemainder      = "remainder_gda_division_impossible"
)

// decnumberDiffKind mirrors the generated runner's triple classes.
type decnumberDiffKind uint8

const (
	decnumberDiffFinite decnumberDiffKind = iota // includes zeros
	decnumberDiffInf
	decnumberDiffQNaN
	decnumberDiffSNaN
)

// decnumberDiffTriple is the generation-time twin of the runner's ddTriple.
type decnumberDiffTriple struct {
	kind  decnumberDiffKind
	sign  bool
	coeff *big.Int // finite only; may be zero
	exp   int32    // quantum exponent, finite only
}

type decnumberDiffWidth struct {
	bits     int
	label    string // anchors map entry name, one of the three width labels
	p        int
	minExp   int32
	maxExp   int32
	maxCoeff *big.Int
	seedBase uint64
}

func decnumberDiffPow10(k int) *big.Int {
	return new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(k)), nil)
}

var decnumberDiffWidths = [...]decnumberDiffWidth{
	{bits: 32, label: "decimal32", p: 7, minExp: -101, maxExp: 90,
		maxCoeff: big.NewInt(9999999), seedBase: 0xdecd1ff032000000},
	{bits: 64, label: "decimal64", p: 16, minExp: -398, maxExp: 369,
		maxCoeff: big.NewInt(9999999999999999), seedBase: 0xdecd1ff064000000},
	{bits: 128, label: "decimal128", p: 34, minExp: -6176, maxExp: 6111,
		maxCoeff: new(big.Int).Sub(decnumberDiffPow10(34), big.NewInt(1)), seedBase: 0xdecd1ff128000000},
}

// Random stream op tags XORed onto the width seed base. Order is the fixed
// execution and digest order of the runner's random phase.
var decnumberDiffRoundedOps = [...]string{"add", "sub", "mul", "div", "quantize"}

var decnumberDiffOpSeedTags = map[string]uint64{
	"add": 0xa1, "sub": 0xa2, "mul": 0xa3, "div": 0xa4,
	"quantize": 0xa5, "fma": 0xa6, "sqrt": 0xa7,
}

// ---- independent triple <-> components bridging over bid_codec_reference ----

func decnumberDiffFromComponents(c bidCodecRefComponents) decnumberDiffTriple {
	switch c.Kind {
	case bidCodecRefInfinity:
		return decnumberDiffTriple{kind: decnumberDiffInf, sign: c.Sign}
	case bidCodecRefQNaN:
		return decnumberDiffTriple{kind: decnumberDiffQNaN, sign: c.Sign}
	case bidCodecRefSNaN:
		return decnumberDiffTriple{kind: decnumberDiffSNaN, sign: c.Sign}
	case bidCodecRefZero:
		return decnumberDiffTriple{kind: decnumberDiffFinite, sign: c.Sign, coeff: new(big.Int), exp: c.Exponent}
	default:
		return decnumberDiffTriple{kind: decnumberDiffFinite, sign: c.Sign, coeff: new(big.Int).Set(c.Coefficient), exp: c.Exponent}
	}
}

func decnumberDiffToComponents(t decnumberDiffTriple) bidCodecRefComponents {
	switch t.kind {
	case decnumberDiffInf:
		return bidCodecRefComponents{Sign: t.sign, Kind: bidCodecRefInfinity}
	case decnumberDiffQNaN:
		return bidCodecRefComponents{Sign: t.sign, Kind: bidCodecRefQNaN}
	case decnumberDiffSNaN:
		return bidCodecRefComponents{Sign: t.sign, Kind: bidCodecRefSNaN}
	default:
		if t.coeff.Sign() == 0 {
			return bidCodecRefComponents{Sign: t.sign, Kind: bidCodecRefZero, Exponent: t.exp}
		}
		return bidCodecRefComponents{Sign: t.sign, Kind: bidCodecRefNormal, Coefficient: new(big.Int).Set(t.coeff), Exponent: t.exp}
	}
}

func decnumberDiffDecode(width decnumberDiffWidth, value bid128BidCodecValue) decnumberDiffTriple {
	switch width.bits {
	case 32:
		return decnumberDiffFromComponents(refDecode32(uint32(value.lo)))
	case 64:
		return decnumberDiffFromComponents(refDecode64(value.lo))
	default:
		return decnumberDiffFromComponents(refDecode128(value.lo, value.hi))
	}
}

func decnumberDiffEncode(width decnumberDiffWidth, t decnumberDiffTriple) (bid128BidCodecValue, error) {
	if t.kind == decnumberDiffFinite {
		if t.exp < width.minExp || t.exp > width.maxExp {
			return bid128BidCodecValue{}, fmt.Errorf("decnumber differential encode: exponent %d out of range for width %d", t.exp, width.bits)
		}
		if t.coeff.Sign() < 0 || t.coeff.Cmp(width.maxCoeff) > 0 {
			return bid128BidCodecValue{}, fmt.Errorf("decnumber differential encode: coefficient %s out of range for width %d", t.coeff, width.bits)
		}
	}
	c := decnumberDiffToComponents(t)
	switch width.bits {
	case 32:
		return bid128BidCodecValue{lo: uint64(refEncode32(c))}, nil
	case 64:
		return bid128BidCodecValue{lo: refEncode64(c)}, nil
	default:
		lo, hi := refEncode128(c)
		return bid128BidCodecValue{lo: lo, hi: hi}, nil
	}
}

// decnumberDiffCanonical reports whether the encoded value survives the
// independent decode->encode round trip unchanged (the closed canonicality
// test used to filter the reused Tier 1 corpora).
func decnumberDiffCanonical(width decnumberDiffWidth, value bid128BidCodecValue) bool {
	decoded := decnumberDiffDecode(width, value)
	// Payloaded NaNs never round-trip through the payloadless triple space;
	// canonicality of NaNs is checked against the zero-payload re-encoding of
	// the raw payload bits below.
	var c bidCodecRefComponents
	switch width.bits {
	case 32:
		c = refDecode32(uint32(value.lo))
	case 64:
		c = refDecode64(value.lo)
	default:
		c = refDecode128(value.lo, value.hi)
	}
	var reencoded bid128BidCodecValue
	switch width.bits {
	case 32:
		reencoded = bid128BidCodecValue{lo: uint64(refEncode32(c))}
	case 64:
		reencoded = bid128BidCodecValue{lo: refEncode64(c)}
	default:
		lo, hi := refEncode128(c)
		reencoded = bid128BidCodecValue{lo: lo, hi: hi}
	}
	_ = decoded
	return reencoded == value
}

func decnumberDiffIsNaN(width decnumberDiffWidth, value bid128BidCodecValue) (isNaN, payloadNonZero bool) {
	var c bidCodecRefComponents
	switch width.bits {
	case 32:
		c = refDecode32(uint32(value.lo))
	case 64:
		c = refDecode64(value.lo)
	default:
		c = refDecode128(value.lo, value.hi)
	}
	if c.Kind != bidCodecRefQNaN && c.Kind != bidCodecRefSNaN {
		return false, false
	}
	return true, c.Payload != nil && c.Payload.Sign() != 0
}

// ---- boundary corpus reuse + filter ----

type decnumberDiffBoundaryFilter struct {
	source                int
	included              []bid128BidCodecValue
	excludedNonCanonical  int
	excludedNaNPayload    int
	includedKindCoverage  map[string]int
	includedZeroCount     int
	includedFiniteNonZero int
}

// decnumberDiffTier1Boundary returns the reused Tier 1 boundary base set for
// one width (function reuse, no data copy). The arithmetic runners' extra
// exponent-cap extension (tier1_exponent_cap_boundary.go) is deliberately not
// part of this corpus; extending it here would repin every decNumber
// differential count and stream-hash anchor.
func decnumberDiffTier1Boundary(width decnumberDiffWidth) []bid128BidCodecValue {
	switch width.bits {
	case 32:
		values := tier1ArithmeticBoundary32Values()
		out := make([]bid128BidCodecValue, len(values))
		for i, v := range values {
			out[i] = bid128BidCodecValue{lo: uint64(v)}
		}
		return out
	case 64:
		values := bid64BidCodecEdgeValues()
		out := make([]bid128BidCodecValue, len(values))
		for i, v := range values {
			out[i] = bid128BidCodecValue{lo: v}
		}
		return out
	default:
		return bid128BidCodecEdgeValues()
	}
}

func decnumberDiffFilterBoundary(width decnumberDiffWidth) decnumberDiffBoundaryFilter {
	source := decnumberDiffTier1Boundary(width)
	filter := decnumberDiffBoundaryFilter{
		source:               len(source),
		includedKindCoverage: map[string]int{},
	}
	for _, value := range source {
		if !decnumberDiffCanonical(width, value) {
			filter.excludedNonCanonical++
			continue
		}
		if isNaN, payloaded := decnumberDiffIsNaN(width, value); isNaN && payloaded {
			filter.excludedNaNPayload++
			continue
		}
		filter.included = append(filter.included, value)
		t := decnumberDiffDecode(width, value)
		switch t.kind {
		case decnumberDiffInf:
			filter.includedKindCoverage["inf"]++
		case decnumberDiffQNaN:
			filter.includedKindCoverage["qnan"]++
		case decnumberDiffSNaN:
			filter.includedKindCoverage["snan"]++
		default:
			if t.coeff.Sign() == 0 {
				filter.includedZeroCount++
				filter.includedKindCoverage["zero"]++
			} else {
				filter.includedFiniteNonZero++
				filter.includedKindCoverage["finite"]++
			}
		}
	}
	return filter
}

// decnumberDiffProbes returns the reused Tier 1 probe set with NaN payloads
// normalized to zero and non-canonical probes excluded (the BID128 steered
// probe has no canonical counterpart).
func decnumberDiffProbes(width decnumberDiffWidth) (probes []bid128BidCodecValue, payloadZeroed, droppedNonCanonical int, err error) {
	var source []bid128BidCodecValue
	switch width.bits {
	case 32:
		source = make([]bid128BidCodecValue, len(tier1ArithmeticProbes32Values))
		for i, v := range tier1ArithmeticProbes32Values {
			source[i] = bid128BidCodecValue{lo: uint64(v)}
		}
	case 64:
		source = make([]bid128BidCodecValue, len(tier1ArithmeticProbes64Values))
		for i, v := range tier1ArithmeticProbes64Values {
			source[i] = bid128BidCodecValue{lo: v}
		}
	default:
		source = append(source, tier1ArithmeticProbes128Values...)
	}
	seen := map[bid128BidCodecValue]bool{}
	for _, value := range source {
		if isNaN, payloaded := decnumberDiffIsNaN(width, value); isNaN && payloaded {
			t := decnumberDiffDecode(width, value)
			normalized, encodeErr := decnumberDiffEncode(width, t)
			if encodeErr != nil {
				return nil, 0, 0, encodeErr
			}
			value = normalized
			payloadZeroed++
		}
		if !decnumberDiffCanonical(width, value) {
			droppedNonCanonical++
			continue
		}
		if seen[value] {
			return nil, 0, 0, fmt.Errorf("decnumber differential probes: width %d probe collision at %016x:%016x after NaN payload normalization", width.bits, value.hi, value.lo)
		}
		seen[value] = true
		probes = append(probes, value)
	}
	return probes, payloadZeroed, droppedNonCanonical, nil
}

// ---- exact-product overflow class (manifest-owned; design 6.1 item 3) ----

func decnumberDiffExactProductValues(width decnumberDiffWidth, offsets []int) ([]bid128BidCodecValue, error) {
	one := big.NewInt(1)
	minNorm := decnumberDiffPow10(width.p - 1)
	var triples []decnumberDiffTriple
	for _, offset := range offsets {
		exp := int32(width.p + offset)
		if exp < width.minExp || exp > width.maxExp {
			return nil, fmt.Errorf("decnumber differential exact-product exponent p%+d out of range for width %d", offset, width.bits)
		}
		triples = append(triples,
			decnumberDiffTriple{kind: decnumberDiffFinite, coeff: one, exp: exp},
			decnumberDiffTriple{kind: decnumberDiffFinite, sign: true, coeff: one, exp: exp},
		)
	}
	triples = append(triples,
		decnumberDiffTriple{kind: decnumberDiffFinite, coeff: width.maxCoeff, exp: width.maxExp},
		decnumberDiffTriple{kind: decnumberDiffFinite, sign: true, coeff: width.maxCoeff, exp: width.maxExp},
		decnumberDiffTriple{kind: decnumberDiffFinite, coeff: minNorm, exp: width.minExp},
		decnumberDiffTriple{kind: decnumberDiffFinite, sign: true, coeff: minNorm, exp: width.minExp},
	)
	values := make([]bid128BidCodecValue, 0, len(triples))
	seen := map[bid128BidCodecValue]bool{}
	for _, t := range triples {
		value, err := decnumberDiffEncode(width, t)
		if err != nil {
			return nil, err
		}
		if seen[value] {
			return nil, fmt.Errorf("decnumber differential exact-product corpus: duplicate value %016x:%016x", value.hi, value.lo)
		}
		seen[value] = true
		values = append(values, value)
	}
	return values, nil
}

func decnumberDiffExactProductZ(width decnumberDiffWidth) ([]bid128BidCodecValue, error) {
	one := big.NewInt(1)
	triples := []decnumberDiffTriple{
		{kind: decnumberDiffFinite, coeff: width.maxCoeff, exp: width.maxExp},
		{kind: decnumberDiffFinite, sign: true, coeff: width.maxCoeff, exp: width.maxExp},
		{kind: decnumberDiffFinite, coeff: new(big.Int), exp: 0},
		{kind: decnumberDiffFinite, coeff: one, exp: 0},
	}
	values := make([]bid128BidCodecValue, 0, len(triples))
	for _, t := range triples {
		value, err := decnumberDiffEncode(width, t)
		if err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	return values, nil
}

// ---- deterministic random triple stream (mirrors the runner template) ----

func decnumberDiffSplitMix64(v uint64) uint64 {
	v += 0x9e3779b97f4a7c15
	v = (v ^ (v >> 30)) * 0xbf58476d1ce4e5b9
	v = (v ^ (v >> 27)) * 0x94d049bb133111eb
	return v ^ (v >> 31)
}

func decnumberDiffRandomWord(seed, caseIndex, lane uint64) uint64 {
	return decnumberDiffSplitMix64(seed ^ caseIndex*0xd1342543de82ef95 ^ lane*0x9e3779b97f4a7c15)
}

// decnumberDiffRandomTriple mirrors the generated runner's random operand
// generator: canonical triple space, NaN payload always zero, quantum-spread
// zeros, partner-correlated exponents. Any drift between this replica and
// the emitted twin fails the anchored stream-hash contract at runtime.
func decnumberDiffRandomTriple(width decnumberDiffWidth, seed, caseIndex, lane uint64, partner *decnumberDiffTriple) decnumberDiffTriple {
	r := decnumberDiffRandomWord(seed, caseIndex, lane*16+0) % 1000
	sign := decnumberDiffRandomWord(seed, caseIndex, lane*16+1)&1 == 1
	expRange := uint64(width.maxExp - width.minExp + 1)
	switch {
	case r < 850: // finite nonzero
		digitCount := 1 + int(decnumberDiffRandomWord(seed, caseIndex, lane*16+2)%uint64(width.p))
		state := decnumberDiffRandomWord(seed, caseIndex, lane*16+3)
		digits := make([]byte, digitCount)
		digits[0] = byte('1' + state%9)
		for i := 1; i < digitCount; i++ {
			state = decnumberDiffSplitMix64(state)
			digits[i] = byte('0' + state%10)
		}
		coeff, ok := new(big.Int).SetString(string(digits), 10)
		if !ok {
			panic("decnumber differential random coefficient parse cannot fail")
		}
		var exp int32
		if partner != nil && partner.kind == decnumberDiffFinite && decnumberDiffRandomWord(seed, caseIndex, lane*16+5)&1 == 0 {
			span := int32(width.p + 4)
			delta := int32(decnumberDiffRandomWord(seed, caseIndex, lane*16+6)%uint64(2*span+1)) - span
			exp = partner.exp + delta
			if exp < width.minExp {
				exp = width.minExp
			}
			if exp > width.maxExp {
				exp = width.maxExp
			}
		} else {
			exp = width.minExp + int32(decnumberDiffRandomWord(seed, caseIndex, lane*16+4)%expRange)
		}
		return decnumberDiffTriple{kind: decnumberDiffFinite, sign: sign, coeff: coeff, exp: exp}
	case r < 910: // zero across the quantum range
		exp := width.minExp + int32(decnumberDiffRandomWord(seed, caseIndex, lane*16+4)%expRange)
		return decnumberDiffTriple{kind: decnumberDiffFinite, sign: sign, coeff: new(big.Int), exp: exp}
	case r < 950:
		return decnumberDiffTriple{kind: decnumberDiffInf, sign: sign}
	case r < 980:
		return decnumberDiffTriple{kind: decnumberDiffQNaN, sign: sign}
	default:
		return decnumberDiffTriple{kind: decnumberDiffSNaN, sign: sign}
	}
}

// ---- known-divergence rows (manifest-owned, hand-audited) ----

// decnumberDiffKnownRow is one parsed known-divergence pin.
type decnumberDiffKnownRow struct {
	spec              DecnumberDifferentialKnownDivergence
	op                string
	arity             int
	mode              int
	x, y, z           bid128BidCodecValue
	matchedStructured uint64
	matchedRandom     uint64
}

func decnumberDiffOpArity(op string) (int, error) {
	switch op {
	case "add", "sub", "mul", "div", "quantize":
		return 2, nil
	case "fma":
		return 3, nil
	case "sqrt":
		return 1, nil
	default:
		return 0, fmt.Errorf("decnumber differential: unknown operation %q", op)
	}
}

func decnumberDiffParseWidthText(width decnumberDiffWidth, text string) (bid128BidCodecValue, error) {
	var value bid128BidCodecValue
	switch width.bits {
	case 32, 64:
		if _, err := fmt.Sscanf(text, "%x", &value.lo); err != nil {
			return value, fmt.Errorf("operand %q: %w", text, err)
		}
	default:
		if _, err := fmt.Sscanf(text, "%x:%x", &value.hi, &value.lo); err != nil {
			return value, fmt.Errorf("128-bit operand %q is not <hi>:<lo>: %w", text, err)
		}
	}
	return value, nil
}

// decnumberDiffParseFFILegText parses a pinned Intel leg in the generated
// FFI shim's text form back into raw words + flags (128-bit bits are the
// little-endian byte hex image).
func decnumberDiffParseFFILegText(width decnumberDiffWidth, leg string) (bid128BidCodecValue, uint32, error) {
	bitsText, flagsText, ok := strings.Cut(leg, "/")
	if !ok {
		return bid128BidCodecValue{}, 0, fmt.Errorf("intel leg %q is not <bits>/<flags>", leg)
	}
	var flags uint32
	if _, err := fmt.Sscanf(flagsText, "%x", &flags); err != nil {
		return bid128BidCodecValue{}, 0, fmt.Errorf("intel leg flags %q: %w", flagsText, err)
	}
	var value bid128BidCodecValue
	switch width.bits {
	case 32, 64:
		if _, err := fmt.Sscanf(bitsText, "%x", &value.lo); err != nil {
			return value, 0, fmt.Errorf("intel leg bits %q: %w", bitsText, err)
		}
	default:
		if len(bitsText) != 32 {
			return value, 0, fmt.Errorf("intel leg 128-bit image %q is not 32 hex chars", bitsText)
		}
		for i := 0; i < 8; i++ {
			var b uint8
			if _, err := fmt.Sscanf(bitsText[2*i:2*i+2], "%x", &b); err != nil {
				return value, 0, fmt.Errorf("intel leg 128-bit image %q: %w", bitsText, err)
			}
			value.lo |= uint64(b) << (8 * i)
			if _, err := fmt.Sscanf(bitsText[16+2*i:16+2*i+2], "%x", &b); err != nil {
				return value, 0, fmt.Errorf("intel leg 128-bit image %q: %w", bitsText, err)
			}
			value.hi |= uint64(b) << (8 * i)
		}
	}
	return value, flags, nil
}

func decnumberDiffParseTripleKey(text string) (decnumberDiffTriple, error) {
	switch text {
	case "Inf":
		return decnumberDiffTriple{kind: decnumberDiffInf}, nil
	case "-Inf":
		return decnumberDiffTriple{kind: decnumberDiffInf, sign: true}, nil
	case "qNaN":
		return decnumberDiffTriple{kind: decnumberDiffQNaN}, nil
	case "sNaN":
		return decnumberDiffTriple{kind: decnumberDiffSNaN}, nil
	}
	sign := false
	rest := text
	if strings.HasPrefix(rest, "-") {
		sign = true
		rest = rest[1:]
	}
	coeffText, expText, ok := strings.Cut(rest, "E")
	if !ok {
		return decnumberDiffTriple{}, fmt.Errorf("triple key %q has no exponent", text)
	}
	coeff, okCoeff := new(big.Int).SetString(coeffText, 10)
	if !okCoeff || coeff.Sign() < 0 {
		return decnumberDiffTriple{}, fmt.Errorf("triple key %q has a bad coefficient", text)
	}
	var exp int32
	if _, err := fmt.Sscanf(expText, "%d", &exp); err != nil {
		return decnumberDiffTriple{}, fmt.Errorf("triple key %q has a bad exponent: %w", text, err)
	}
	return decnumberDiffTriple{kind: decnumberDiffFinite, sign: sign, coeff: coeff, exp: exp}, nil
}

func decnumberDiffTripleEqual(a, b decnumberDiffTriple) bool {
	if a.kind != b.kind {
		return false
	}
	switch a.kind {
	case decnumberDiffInf:
		return a.sign == b.sign
	case decnumberDiffQNaN, decnumberDiffSNaN:
		return true
	default:
		return a.sign == b.sign && a.exp == b.exp && a.coeff.Cmp(b.coeff) == 0
	}
}

const decnumberDiffFlags5MaskWord = uint32(0x3d)

func decnumberDiffParseKnownRows(width decnumberDiffWidth, spec DecnumberDifferentialSpec) ([]*decnumberDiffKnownRow, error) {
	var rows []*decnumberDiffKnownRow
	for i, entry := range spec.KnownDivergences {
		if entry.Width != width.label {
			continue
		}
		arity, err := decnumberDiffOpArity(entry.Operation)
		if err != nil {
			return nil, fmt.Errorf("known_divergences[%d]: %w", i, err)
		}
		if entry.Mode < 0 || entry.Mode > 4 {
			return nil, fmt.Errorf("known_divergences[%d]: mode %d out of range", i, entry.Mode)
		}
		if entry.Operation == "sqrt" && entry.Mode != 0 {
			return nil, fmt.Errorf("known_divergences[%d]: sqrt is generated at round-nearest-even only", i)
		}
		if entry.Classification != "decnumber_defect" {
			return nil, fmt.Errorf("known_divergences[%d]: classification %q is not a known-divergence bucket (a pinned-Intel IEEE violation goes through the IEEE-deviation procedure)", i, entry.Classification)
		}
		row := &decnumberDiffKnownRow{spec: entry, op: entry.Operation, arity: arity, mode: entry.Mode}
		if row.x, err = decnumberDiffParseWidthText(width, entry.X); err != nil {
			return nil, fmt.Errorf("known_divergences[%d]: x: %w", i, err)
		}
		if arity >= 2 {
			if entry.Y == "" {
				return nil, fmt.Errorf("known_divergences[%d]: %s requires y", i, entry.Operation)
			}
			if row.y, err = decnumberDiffParseWidthText(width, entry.Y); err != nil {
				return nil, fmt.Errorf("known_divergences[%d]: y: %w", i, err)
			}
		}
		if arity >= 3 {
			if entry.Z == "" {
				return nil, fmt.Errorf("known_divergences[%d]: %s requires z", i, entry.Operation)
			}
			if row.z, err = decnumberDiffParseWidthText(width, entry.Z); err != nil {
				return nil, fmt.Errorf("known_divergences[%d]: z: %w", i, err)
			}
		}
		for slot, value := range []bid128BidCodecValue{row.x, row.y, row.z}[:arity] {
			if !decnumberDiffCanonical(width, value) {
				return nil, fmt.Errorf("known_divergences[%d]: operand %d is non-canonical", i, slot)
			}
		}
		if arity == 3 && decnumberDiffFmaExcluded(
			decnumberDiffDecode(width, row.x), decnumberDiffDecode(width, row.y), decnumberDiffDecode(width, row.z)) {
			return nil, fmt.Errorf("known_divergences[%d]: tuple lies in the excluded L13 class and never executes", i)
		}
		intelBits, intelFlags, err := decnumberDiffParseFFILegText(width, entry.Intel)
		if err != nil {
			return nil, fmt.Errorf("known_divergences[%d]: %w", i, err)
		}
		dnKeyText, dnFlagsText, ok := strings.Cut(entry.Decnumber, "/")
		if !ok {
			return nil, fmt.Errorf("known_divergences[%d]: decnumber pin %q is not <triple>/<flags5>", i, entry.Decnumber)
		}
		dnTriple, err := decnumberDiffParseTripleKey(dnKeyText)
		if err != nil {
			return nil, fmt.Errorf("known_divergences[%d]: %w", i, err)
		}
		var dnFlags uint32
		if _, err := fmt.Sscanf(dnFlagsText, "%x", &dnFlags); err != nil {
			return nil, fmt.Errorf("known_divergences[%d]: decnumber flags %q: %w", i, dnFlagsText, err)
		}
		if dnFlags&^decnumberDiffFlags5MaskWord != 0 {
			return nil, fmt.Errorf("known_divergences[%d]: decnumber flags %02x leave the 5-flag surface", i, dnFlags)
		}
		intelTriple := decnumberDiffDecode(width, intelBits)
		if decnumberDiffTripleEqual(intelTriple, dnTriple) && intelFlags&decnumberDiffFlags5MaskWord == dnFlags {
			return nil, fmt.Errorf("known_divergences[%d]: pinned legs agree on the comparison surface; the row is stale, not a divergence", i)
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func (r *decnumberDiffKnownRow) matchesPair(op string, x, y bid128BidCodecValue) bool {
	return r.arity == 2 && r.op == op && r.x == x && r.y == y
}

func (r *decnumberDiffKnownRow) matchesTriple(x, y, z bid128BidCodecValue) bool {
	return r.arity == 3 && r.x == x && r.y == y && r.z == z
}

func (r *decnumberDiffKnownRow) matchesUnary(op string, x bid128BidCodecValue) bool {
	return r.arity == 1 && r.op == op && r.x == x
}

// decnumberDiffFmaExcluded is the generation-time class predicate for ledger
// L13 (fma with one product operand ±0 of any quantum, the other ±Inf, and a
// quiet NaN addend; IEEE 754-2019 7.2 leaves the invalid signal
// implementation-defined there, and pinned Intel C and decNumber measurably
// disagree). The runner replays the same predicate and both sides count.
func decnumberDiffFmaExcluded(x, y, z decnumberDiffTriple) bool {
	isZero := func(t decnumberDiffTriple) bool {
		return t.kind == decnumberDiffFinite && t.coeff.Sign() == 0
	}
	isInf := func(t decnumberDiffTriple) bool { return t.kind == decnumberDiffInf }
	if z.kind != decnumberDiffQNaN {
		return false
	}
	return (isZero(x) && isInf(y)) || (isInf(x) && isZero(y))
}

// ---- corpus digests (shared FNV mix, tier1 pattern) ----

func decnumberDiffMix(digest, word uint64) uint64 {
	return (digest ^ word) * decnumberDiffHashPrime
}

func decnumberDiffMixValue(digest uint64, value bid128BidCodecValue) uint64 {
	digest = decnumberDiffMix(digest, value.lo)
	return decnumberDiffMix(digest, value.hi)
}

// decnumberDiffVisitStructuredPairs replicates the runner's structured pair
// visitation: boundary x probes in both operand orders, then probes x probes,
// then the exact-product cross.
func decnumberDiffVisitStructuredPairs(boundary, probes, exactProduct []bid128BidCodecValue, visit func(x, y bid128BidCodecValue)) {
	for _, x := range boundary {
		for _, y := range probes {
			visit(x, y)
			visit(y, x)
		}
	}
	for _, x := range probes {
		for _, y := range probes {
			visit(x, y)
		}
	}
	for _, x := range exactProduct {
		for _, y := range exactProduct {
			visit(x, y)
		}
	}
}

// decnumberDiffVisitStructuredTriples replicates the runner's structured fma
// visitation: the boundary value rotated through all three operand slots
// against (i+j)%len(probes) companion pairs, then the full probes^3 cross,
// then the exact-product x exact-product x z cross.
func decnumberDiffVisitStructuredTriples(boundary, probes, exactProduct, exactProductZ []bid128BidCodecValue, visit func(x, y, z bid128BidCodecValue)) {
	for i, x := range boundary {
		for j, y := range probes {
			z := probes[(i+j)%len(probes)]
			visit(x, y, z)
			visit(y, z, x)
			visit(z, x, y)
		}
	}
	for _, x := range probes {
		for _, y := range probes {
			for _, z := range probes {
				visit(x, y, z)
			}
		}
	}
	for _, x := range exactProduct {
		for _, y := range exactProduct {
			for _, z := range exactProductZ {
				visit(x, y, z)
			}
		}
	}
}

func decnumberDiffStructuredStreamHash(width decnumberDiffWidth, boundary, probes, exactProduct, exactProductZ []bid128BidCodecValue) uint64 {
	digest := decnumberDiffMix(decnumberDiffHashOffset, uint64(width.bits))
	var visits uint64
	decnumberDiffVisitStructuredPairs(boundary, probes, exactProduct, func(x, y bid128BidCodecValue) {
		digest = decnumberDiffMixValue(digest, x)
		digest = decnumberDiffMixValue(digest, y)
		visits++
	})
	decnumberDiffVisitStructuredTriples(boundary, probes, exactProduct, exactProductZ, func(x, y, z bid128BidCodecValue) {
		digest = decnumberDiffMixValue(digest, x)
		digest = decnumberDiffMixValue(digest, y)
		digest = decnumberDiffMixValue(digest, z)
		visits++
	})
	for _, x := range boundary {
		digest = decnumberDiffMixValue(digest, x)
		visits++
	}
	return decnumberDiffMix(digest, visits)
}

type decnumberDiffRandomStats struct {
	streamHash        uint64
	fmaExcludedTriple uint64
	knownDivergence   uint64
}

func decnumberDiffRandomStream(width decnumberDiffWidth, pairsPerOp uint64, knownRows []*decnumberDiffKnownRow) (decnumberDiffRandomStats, error) {
	digest := decnumberDiffMix(decnumberDiffHashOffset, uint64(width.bits))
	var total uint64
	var stats decnumberDiffRandomStats
	mixTriple := func(t decnumberDiffTriple) (bid128BidCodecValue, error) {
		value, err := decnumberDiffEncode(width, t)
		if err != nil {
			return value, err
		}
		digest = decnumberDiffMixValue(digest, value)
		return value, nil
	}
	for _, op := range decnumberDiffRoundedOps {
		seed := width.seedBase ^ decnumberDiffOpSeedTags[op]
		for caseIndex := uint64(0); caseIndex < pairsPerOp; caseIndex++ {
			x := decnumberDiffRandomTriple(width, seed, caseIndex, 0, nil)
			y := decnumberDiffRandomTriple(width, seed, caseIndex, 1, &x)
			xBits, err := mixTriple(x)
			if err != nil {
				return stats, err
			}
			yBits, err := mixTriple(y)
			if err != nil {
				return stats, err
			}
			for _, row := range knownRows {
				if row.matchesPair(op, xBits, yBits) {
					row.matchedRandom++
					stats.knownDivergence++
				}
			}
			total++
		}
	}
	fmaSeed := width.seedBase ^ decnumberDiffOpSeedTags["fma"]
	for caseIndex := uint64(0); caseIndex < pairsPerOp; caseIndex++ {
		x := decnumberDiffRandomTriple(width, fmaSeed, caseIndex, 0, nil)
		y := decnumberDiffRandomTriple(width, fmaSeed, caseIndex, 1, &x)
		z := decnumberDiffRandomTriple(width, fmaSeed, caseIndex, 2, &x)
		xBits, err := mixTriple(x)
		if err != nil {
			return stats, err
		}
		yBits, err := mixTriple(y)
		if err != nil {
			return stats, err
		}
		zBits, err := mixTriple(z)
		if err != nil {
			return stats, err
		}
		if decnumberDiffFmaExcluded(x, y, z) {
			stats.fmaExcludedTriple++
		} else {
			for _, row := range knownRows {
				if row.matchesTriple(xBits, yBits, zBits) {
					row.matchedRandom++
					stats.knownDivergence++
				}
			}
		}
		total++
	}
	sqrtSeed := width.seedBase ^ decnumberDiffOpSeedTags["sqrt"]
	for caseIndex := uint64(0); caseIndex < pairsPerOp; caseIndex++ {
		x := decnumberDiffRandomTriple(width, sqrtSeed, caseIndex, 0, nil)
		xBits, err := mixTriple(x)
		if err != nil {
			return stats, err
		}
		for _, row := range knownRows {
			if row.matchesUnary("sqrt", xBits) {
				row.matchedRandom++
				stats.knownDivergence++
			}
		}
		total++
	}
	stats.streamHash = decnumberDiffMix(digest, total)
	return stats, nil
}

// ---- per-width generation summary ----

type decnumberDiffWidthPlan struct {
	width         decnumberDiffWidth
	boundary      decnumberDiffBoundaryFilter
	probes        []bid128BidCodecValue
	probesZeroed  int
	probesDropped int
	exactProduct  []bid128BidCodecValue
	exactProductZ []bid128BidCodecValue

	structuredPairs       uint64
	structuredFmaTriples  uint64
	structuredFmaExcluded uint64
	structuredSqrtCases   uint64
	structuredComparisons uint64
	structuredStreamHash  uint64

	randomPairsPerOp    uint64
	randomFmaExcluded   uint64
	randomComparisons   uint64
	randomStreamHash    uint64
	totalComparisons    uint64
	sqrtNonRNStructural uint64

	knownRows                 []*decnumberDiffKnownRow
	structuredKnownDivergence uint64
	randomKnownDivergence     uint64
}

func decnumberDiffBuildWidthPlan(width decnumberDiffWidth, spec DecnumberDifferentialSpec) (decnumberDiffWidthPlan, error) {
	plan := decnumberDiffWidthPlan{width: width}
	plan.boundary = decnumberDiffFilterBoundary(width)
	for _, class := range []string{"zero", "finite", "inf", "qnan", "snan"} {
		if plan.boundary.includedKindCoverage[class] == 0 {
			return plan, fmt.Errorf("decnumber differential width %d: filtered boundary corpus lost operand class %q", width.bits, class)
		}
	}
	probes, zeroed, dropped, err := decnumberDiffProbes(width)
	if err != nil {
		return plan, err
	}
	plan.probes, plan.probesZeroed, plan.probesDropped = probes, zeroed, dropped
	plan.exactProduct, err = decnumberDiffExactProductValues(width, spec.ExactProductExponentOffsets)
	if err != nil {
		return plan, err
	}
	plan.exactProductZ, err = decnumberDiffExactProductZ(width)
	if err != nil {
		return plan, err
	}

	plan.knownRows, err = decnumberDiffParseKnownRows(width, spec)
	if err != nil {
		return plan, err
	}

	decnumberDiffVisitStructuredPairs(plan.boundary.included, plan.probes, plan.exactProduct, func(x, y bid128BidCodecValue) {
		plan.structuredPairs++
		for _, row := range plan.knownRows {
			if row.arity != 2 {
				continue
			}
			// A pinned pair executes under every rounded operation; the row
			// redirects only its own operation/mode evaluation.
			if row.x == x && row.y == y {
				row.matchedStructured++
				plan.structuredKnownDivergence++
			}
		}
	})
	decnumberDiffVisitStructuredTriples(plan.boundary.included, plan.probes, plan.exactProduct, plan.exactProductZ, func(x, y, z bid128BidCodecValue) {
		plan.structuredFmaTriples++
		if decnumberDiffFmaExcluded(decnumberDiffDecode(width, x), decnumberDiffDecode(width, y), decnumberDiffDecode(width, z)) {
			plan.structuredFmaExcluded++
			return
		}
		for _, row := range plan.knownRows {
			if row.matchesTriple(x, y, z) {
				row.matchedStructured++
				plan.structuredKnownDivergence++
			}
		}
	})
	plan.structuredSqrtCases = uint64(len(plan.boundary.included))
	for _, value := range plan.boundary.included {
		for _, row := range plan.knownRows {
			if row.matchesUnary("sqrt", value) {
				row.matchedStructured++
				plan.structuredKnownDivergence++
			}
		}
	}
	plan.structuredStreamHash = decnumberDiffStructuredStreamHash(width, plan.boundary.included, plan.probes, plan.exactProduct, plan.exactProductZ)

	const modes = 5
	const roundedOps = 5
	plan.structuredComparisons = plan.structuredPairs*roundedOps*modes +
		(plan.structuredFmaTriples-plan.structuredFmaExcluded)*modes +
		plan.structuredSqrtCases

	plan.randomPairsPerOp = uint64(spec.RandomPairsPerOperation)
	randomStats, err := decnumberDiffRandomStream(width, plan.randomPairsPerOp, plan.knownRows)
	if err != nil {
		return plan, err
	}
	plan.randomFmaExcluded = randomStats.fmaExcludedTriple
	plan.randomStreamHash = randomStats.streamHash
	plan.randomKnownDivergence = randomStats.knownDivergence
	plan.randomComparisons = plan.randomPairsPerOp*roundedOps*modes +
		(plan.randomPairsPerOp-plan.randomFmaExcluded)*modes +
		plan.randomPairsPerOp
	plan.totalComparisons = plan.structuredComparisons + plan.randomComparisons
	plan.sqrtNonRNStructural = (plan.structuredSqrtCases + plan.randomPairsPerOp) * 4
	for _, row := range plan.knownRows {
		if row.matchedStructured+row.matchedRandom == 0 {
			return plan, fmt.Errorf("decnumber differential width %d: known-divergence row %s (%s) matches no executed case; the pin is stale",
				width.bits, row.spec.ReasonID, row.spec.X)
		}
	}
	return plan, nil
}

// ---- inventory ----

type decnumberDiffInventoryWidth struct {
	Width                        string `json:"width"`
	BoundarySource               int    `json:"boundary_source_values"`
	BoundaryIncluded             int    `json:"boundary_included_values"`
	BoundaryExcludedNonCanonical int    `json:"boundary_excluded_noncanonical_operand_no_gda_counterpart"`
	BoundaryExcludedNaNPayload   int    `json:"boundary_excluded_nan_payload_operand_zero_only"`
	ProbeValues                  int    `json:"probe_values"`
	ProbeNaNPayloadZeroed        int    `json:"probe_nan_payload_zeroed"`
	ProbeDroppedNonCanonical     int    `json:"probe_excluded_noncanonical_operand_no_gda_counterpart"`
	ExactProductValues           int    `json:"exact_product_values"`
	ExactProductAddends          int    `json:"exact_product_fma_addends"`
	StructuredPairs              uint64 `json:"structured_pairs"`
	StructuredFmaTriples         uint64 `json:"structured_fma_triples"`
	StructuredFmaExcluded        uint64 `json:"structured_fma_excluded_fma_zero_inf_qnan_invalid_ieee_optional"`
	StructuredSqrtCases          uint64 `json:"structured_sqrt_cases"`
	StructuredComparisons        uint64 `json:"structured_comparisons"`
	StructuredKnownDivergences   uint64 `json:"structured_known_divergences"`
	StructuredStreamHash         uint64 `json:"structured_stream_hash"`
	RandomPairsPerOperation      uint64 `json:"random_pairs_per_operation"`
	RandomFmaExcluded            uint64 `json:"random_fma_excluded_fma_zero_inf_qnan_invalid_ieee_optional"`
	RandomComparisons            uint64 `json:"random_comparisons"`
	RandomKnownDivergences       uint64 `json:"random_known_divergences"`
	RandomStreamHash             uint64 `json:"random_stream_hash"`
	TotalComparisons             uint64 `json:"total_comparisons"`
	SqrtNonRNNotGenerated        uint64 `json:"sqrt_non_rn_cases_not_generated_sqrt_gda_fixed_half_even_rounding"`
}

type decnumberDiffInventoryKnownDivergence struct {
	ReasonID       string `json:"reason_id"`
	Classification string `json:"classification"`
	Width          string `json:"width"`
	Operation      string `json:"operation"`
	Mode           int    `json:"mode"`
	X              string `json:"x"`
	Y              string `json:"y,omitempty"`
	Z              string `json:"z,omitempty"`
	Intel          string `json:"intel"`
	Decnumber      string `json:"decnumber"`
	Note           string `json:"note"`
	MatchedCases   uint64 `json:"matched_cases"`
}

type decnumberDiffInventoryExclusion struct {
	ID         string `json:"id"`
	Kind       string `json:"kind"` // generated_class_predicate | corpus_filter | structural
	Definition string `json:"definition"`
}

type decnumberDiffInventory struct {
	Domain           string                                  `json:"domain"`
	Oracle           string                                  `json:"oracle"`
	Status           string                                  `json:"status"`
	Operations       []string                                `json:"operations"`
	SqrtRoundingMode string                                  `json:"sqrt_rounding_mode"`
	RoundingModes    []string                                `json:"rounding_modes"`
	ExclusionClasses []decnumberDiffInventoryExclusion       `json:"exclusion_classes"`
	KnownDivergences []decnumberDiffInventoryKnownDivergence `json:"known_divergences"`
	Widths           []decnumberDiffInventoryWidth           `json:"widths"`
}

// ---- template rendering ----

func WriteDecnumberDifferentialOutputs(repoRoot string, manifest Manifest) error {
	files, err := GenerateDecnumberDifferentialOutputs(manifest)
	if err != nil {
		return err
	}
	for path, data := range files {
		fullPath := filepath.Join(repoRoot, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			return fmt.Errorf("mkdir %q: %w", filepath.Dir(fullPath), err)
		}
		if err := os.WriteFile(fullPath, data, 0o644); err != nil {
			return fmt.Errorf("write generated decNumber differential artifact %q: %w", fullPath, err)
		}
	}
	return nil
}

func GenerateDecnumberDifferentialOutputs(manifest Manifest) (map[string][]byte, error) {
	spec := manifest.DecnumberDifferential
	if spec == nil {
		return nil, fmt.Errorf("decnumber differential codegen: manifest block decnumber_differential is required")
	}

	plans := make([]decnumberDiffWidthPlan, 0, len(decnumberDiffWidths))
	for _, width := range decnumberDiffWidths {
		plan, err := decnumberDiffBuildWidthPlan(width, *spec)
		if err != nil {
			return nil, err
		}
		plans = append(plans, plan)
	}

	sentinelRows, err := GenerateDecnumberDifferentialSentinelRows()
	if err != nil {
		return nil, err
	}

	sharedTemplate, err := decnumberDiffTemplates.ReadFile("decnumber_differential_templates/go_decnumber_differential_shared_test.go.tmpl")
	if err != nil {
		return nil, fmt.Errorf("read decNumber differential shared template: %w", err)
	}
	shimTemplate, err := decnumberDiffTemplates.ReadFile("decnumber_differential_templates/go_decnumber_differential_native.go.tmpl")
	if err != nil {
		return nil, fmt.Errorf("read decNumber differential native shim template: %w", err)
	}
	runnerTemplate, err := decnumberDiffTemplates.ReadFile("decnumber_differential_templates/go_decnumber_differential_native_test.go.tmpl")
	if err != nil {
		return nil, fmt.Errorf("read decNumber differential runner template: %w", err)
	}
	stubTemplate, err := decnumberDiffTemplates.ReadFile("decnumber_differential_templates/go_decnumber_differential_stub_test.go.tmpl")
	if err != nil {
		return nil, fmt.Errorf("read decNumber differential stub template: %w", err)
	}

	replacements := []string{
		"@@DND_SENTINEL_COUNT@@", fmt.Sprint(len(sentinelRows)),
		"@@DND_SENTINEL_ROWS@@", decnumberDiffSentinelGoRowLiterals(sentinelRows),
		"@@DND_RANDOM_PAIRS_PER_OP@@", fmt.Sprint(spec.RandomPairsPerOperation),
	}
	for _, plan := range plans {
		suffix := fmt.Sprint(plan.width.bits)
		replacements = append(replacements,
			"@@DND_BOUNDARY"+suffix+"_COUNT@@", fmt.Sprint(len(plan.boundary.included)),
			"@@DND_PROBES"+suffix+"_COUNT@@", fmt.Sprint(len(plan.probes)),
			"@@DND_EXACT_PRODUCT"+suffix+"_COUNT@@", fmt.Sprint(len(plan.exactProduct)),
			"@@DND_EXACT_PRODUCT_Z"+suffix+"_COUNT@@", fmt.Sprint(len(plan.exactProductZ)),
			"@@DND_STRUCTURED_FMA_EXCLUDED"+suffix+"@@", fmt.Sprint(plan.structuredFmaExcluded),
			"@@DND_RANDOM_FMA_EXCLUDED"+suffix+"@@", fmt.Sprint(plan.randomFmaExcluded),
			"@@DND_STRUCTURED_COMPARISONS"+suffix+"@@", fmt.Sprint(plan.structuredComparisons),
			"@@DND_RANDOM_COMPARISONS"+suffix+"@@", fmt.Sprint(plan.randomComparisons),
			"@@DND_TOTAL_COMPARISONS"+suffix+"@@", fmt.Sprint(plan.totalComparisons),
			"@@DND_STRUCTURED_STREAM_HASH"+suffix+"@@", fmt.Sprint(plan.structuredStreamHash),
			"@@DND_RANDOM_STREAM_HASH"+suffix+"@@", fmt.Sprint(plan.randomStreamHash),
			"@@DND_SEED_BASE"+suffix+"@@", fmt.Sprintf("0x%016x", plan.width.seedBase),
			"@@DND_STRUCTURED_KNOWN_DIVERGENCES"+suffix+"@@", fmt.Sprint(plan.structuredKnownDivergence),
			"@@DND_RANDOM_KNOWN_DIVERGENCES"+suffix+"@@", fmt.Sprint(plan.randomKnownDivergence),
			"@@DND_KNOWN_DIVERGENCE_ROWS"+suffix+"@@", decnumberDiffKnownRowLiterals(plan.knownRows),
		)
		switch plan.width.bits {
		case 32:
			replacements = append(replacements,
				"@@DND_BOUNDARY32_VALUES@@", decnumberDiffUint32Literals(plan.boundary.included),
				"@@DND_PROBES32_VALUES@@", decnumberDiffUint32Literals(plan.probes),
				"@@DND_EXACT_PRODUCT32_VALUES@@", decnumberDiffUint32Literals(plan.exactProduct),
				"@@DND_EXACT_PRODUCT_Z32_VALUES@@", decnumberDiffUint32Literals(plan.exactProductZ),
			)
		case 64:
			replacements = append(replacements,
				"@@DND_BOUNDARY64_VALUES@@", decnumberDiffUint64Literals(plan.boundary.included),
				"@@DND_PROBES64_VALUES@@", decnumberDiffUint64Literals(plan.probes),
				"@@DND_EXACT_PRODUCT64_VALUES@@", decnumberDiffUint64Literals(plan.exactProduct),
				"@@DND_EXACT_PRODUCT_Z64_VALUES@@", decnumberDiffUint64Literals(plan.exactProductZ),
			)
		default:
			replacements = append(replacements,
				"@@DND_BOUNDARY128_VALUES@@", tier1ArithmeticUint128Literals(plan.boundary.included),
				"@@DND_PROBES128_VALUES@@", tier1ArithmeticUint128Literals(plan.probes),
				"@@DND_EXACT_PRODUCT128_VALUES@@", tier1ArithmeticUint128Literals(plan.exactProduct),
				"@@DND_EXACT_PRODUCT_Z128_VALUES@@", tier1ArithmeticUint128Literals(plan.exactProductZ),
			)
		}
	}
	replacer := strings.NewReplacer(replacements...)

	outputs := map[string][]byte{
		decnumberDiffSharedGeneratedPath:     []byte(genmarker.Line("testgen") + "\n" + replacer.Replace(string(sharedTemplate))),
		decnumberDiffNativeShimGeneratedPath: []byte(genmarker.Line("testgen") + "\n" + replacer.Replace(string(shimTemplate))),
		decnumberDiffRunnerGeneratedPath:     []byte(genmarker.Line("testgen") + "\n" + replacer.Replace(string(runnerTemplate))),
		decnumberDiffStubGeneratedPath:       []byte(genmarker.Line("testgen") + "\n" + replacer.Replace(string(stubTemplate))),
	}
	// The cgo preamble comment block must not pass through gofmt-based
	// formatting with the template tokens unresolved; format all Go outputs
	// after replacement (cgo files survive format.Source unchanged).
	formatted, err := formatGeneratedGoOutputs(outputs)
	if err != nil {
		return nil, err
	}

	inventory := decnumberDiffInventory{
		Domain: "decnumber_differential",
		Oracle: "IBM decNumber 3.68 (pinned by devtools/scripts/install_ibm_decnumber.sh)",
		Status: "additional generated differential gate; not a regular verification domain; decNumber is a divergence tripwire, not a correctness definition",
		Operations: []string{
			"add", "sub", "mul", "div", "quantize", "fma", "sqrt",
		},
		SqrtRoundingMode: "nearest_even_only",
		RoundingModes:    []string{"nearest_even", "toward_negative", "toward_positive", "toward_zero", "nearest_away"},
		ExclusionClasses: []decnumberDiffInventoryExclusion{
			{
				ID:         decnumberDiffClassFmaZeroInfQnan,
				Kind:       "generated_class_predicate",
				Definition: "fma where one product operand is a zero of any quantum and sign, the other is an infinity of any sign, and the addend is a quiet NaN; IEEE 754-2019 7.2 leaves the invalid signal implementation-defined and pinned Intel C and decNumber 3.68 measurably disagree",
			},
			{
				ID:         decnumberDiffClassSqrtNonRN,
				Kind:       "structural",
				Definition: "sqrt cases are generated for round-nearest-even only; GDA squareRoot rounds half-even regardless of context rounding, so non-RN sqrt has no decNumber counterpart and stays covered by the Intel-oracle domains only",
			},
			{
				ID:         decnumberDiffClassNonCanonical,
				Kind:       "corpus_filter",
				Definition: "non-canonical BID operand encodings have no GDA counterpart; they are excluded from the reused Tier 1 corpora and stay covered by the FFI and BID codec domains",
			},
			{
				ID:         decnumberDiffClassNaNPayload,
				Kind:       "corpus_filter",
				Definition: "v1 corpus NaN operands carry payload zero only (NaN payload propagation is an IEEE 'should'); payloaded boundary NaNs are excluded and probe NaN payloads are normalized to zero",
			},
			{
				ID:         decnumberDiffClassRemainder,
				Kind:       "structural",
				Definition: "remainder/fmod/remainderNear are not in the v1 operation set: GDA raises Division_impossible where IEEE remainder is always exact, the divergence class already documented for the decTest domain",
			},
		},
	}
	for _, plan := range plans {
		inventory.Widths = append(inventory.Widths, decnumberDiffInventoryWidth{
			Width:                        plan.width.label,
			BoundarySource:               plan.boundary.source,
			BoundaryIncluded:             len(plan.boundary.included),
			BoundaryExcludedNonCanonical: plan.boundary.excludedNonCanonical,
			BoundaryExcludedNaNPayload:   plan.boundary.excludedNaNPayload,
			ProbeValues:                  len(plan.probes),
			ProbeNaNPayloadZeroed:        plan.probesZeroed,
			ProbeDroppedNonCanonical:     plan.probesDropped,
			ExactProductValues:           len(plan.exactProduct),
			ExactProductAddends:          len(plan.exactProductZ),
			StructuredPairs:              plan.structuredPairs,
			StructuredFmaTriples:         plan.structuredFmaTriples,
			StructuredFmaExcluded:        plan.structuredFmaExcluded,
			StructuredSqrtCases:          plan.structuredSqrtCases,
			StructuredComparisons:        plan.structuredComparisons,
			StructuredKnownDivergences:   plan.structuredKnownDivergence,
			StructuredStreamHash:         plan.structuredStreamHash,
			RandomPairsPerOperation:      plan.randomPairsPerOp,
			RandomFmaExcluded:            plan.randomFmaExcluded,
			RandomComparisons:            plan.randomComparisons,
			RandomKnownDivergences:       plan.randomKnownDivergence,
			RandomStreamHash:             plan.randomStreamHash,
			TotalComparisons:             plan.totalComparisons,
			SqrtNonRNNotGenerated:        plan.sqrtNonRNStructural,
		})
		for _, row := range plan.knownRows {
			inventory.KnownDivergences = append(inventory.KnownDivergences, decnumberDiffInventoryKnownDivergence{
				ReasonID:       row.spec.ReasonID,
				Classification: row.spec.Classification,
				Width:          row.spec.Width,
				Operation:      row.spec.Operation,
				Mode:           row.spec.Mode,
				X:              row.spec.X,
				Y:              row.spec.Y,
				Z:              row.spec.Z,
				Intel:          row.spec.Intel,
				Decnumber:      row.spec.Decnumber,
				Note:           row.spec.Note,
				MatchedCases:   row.matchedStructured + row.matchedRandom,
			})
		}
	}
	inventoryJSON, err := json.MarshalIndent(inventory, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal decNumber differential inventory: %w", err)
	}
	formatted[filepath.Join(spec.InventoryOutput)] = append(inventoryJSON, '\n')
	return formatted, nil
}

// decnumberDiffKnownRowLiterals renders the runner's known-divergence table
// literal for one width.
func decnumberDiffKnownRowLiterals(rows []*decnumberDiffKnownRow) string {
	var out strings.Builder
	for _, row := range rows {
		fmt.Fprintf(&out,
			"\t{op: %q, mode: %d, xLo: 0x%016x, xHi: 0x%016x, yLo: 0x%016x, yHi: 0x%016x, zLo: 0x%016x, zHi: 0x%016x, intelLeg: %q, dnKey: %q, dnFlags5: 0x%02x, reason: %q},\n",
			row.op, row.mode, row.x.lo, row.x.hi, row.y.lo, row.y.hi, row.z.lo, row.z.hi,
			row.spec.Intel, decnumberDiffKnownRowDnKey(row), decnumberDiffKnownRowDnFlags(row), row.spec.ReasonID)
	}
	return out.String()
}

func decnumberDiffKnownRowDnKey(row *decnumberDiffKnownRow) string {
	key, _, _ := strings.Cut(row.spec.Decnumber, "/")
	return key
}

func decnumberDiffKnownRowDnFlags(row *decnumberDiffKnownRow) uint32 {
	_, flagsText, _ := strings.Cut(row.spec.Decnumber, "/")
	var flags uint32
	fmt.Sscanf(flagsText, "%x", &flags)
	return flags
}

func decnumberDiffUint32Literals(values []bid128BidCodecValue) string {
	var out strings.Builder
	for _, value := range values {
		fmt.Fprintf(&out, "\t0x%08x,\n", uint32(value.lo))
	}
	return out.String()
}

func decnumberDiffUint64Literals(values []bid128BidCodecValue) string {
	var out strings.Builder
	for _, value := range values {
		fmt.Fprintf(&out, "\t0x%016x,\n", value.lo)
	}
	return out.String()
}
