//go:build cgo && bid754_native && bid754_tier1_long

package bid754

import (
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/sky1core/bid754/bid754-go/internal/bidgo"
)

const (
	tier1ArithmeticBoundary32Count  = uint64(@@TIER1_BOUNDARY32_COUNT@@)
	tier1ArithmeticBoundary64Count  = uint64(@@TIER1_BOUNDARY64_COUNT@@)
	tier1ArithmeticBoundary128Count = uint64(@@TIER1_BOUNDARY128_COUNT@@)
	tier1ArithmeticSemanticRounded32Count  = uint64(@@TIER1_SEMANTIC_ROUNDED32_COUNT@@)
	tier1ArithmeticSemanticRounded64Count  = uint64(@@TIER1_SEMANTIC_ROUNDED64_COUNT@@)
	tier1ArithmeticSemanticRounded128Count = uint64(@@TIER1_SEMANTIC_ROUNDED128_COUNT@@)
	tier1ArithmeticSemanticScale32Count    = uint64(@@TIER1_SEMANTIC_SCALE32_COUNT@@)
	tier1ArithmeticSemanticScale64Count    = uint64(@@TIER1_SEMANTIC_SCALE64_COUNT@@)
	tier1ArithmeticSemanticScale128Count   = uint64(@@TIER1_SEMANTIC_SCALE128_COUNT@@)
	tier1ArithmeticSemanticRemainder32Count  = uint64(@@TIER1_SEMANTIC_REMAINDER32_COUNT@@)
	tier1ArithmeticSemanticRemainder64Count  = uint64(@@TIER1_SEMANTIC_REMAINDER64_COUNT@@)
	tier1ArithmeticSemanticRemainder128Count = uint64(@@TIER1_SEMANTIC_REMAINDER128_COUNT@@)
	tier1ArithmeticSemanticFma32Count  = uint64(@@TIER1_SEMANTIC_FMA32_COUNT@@)
	tier1ArithmeticSemanticFma64Count  = uint64(@@TIER1_SEMANTIC_FMA64_COUNT@@)
	tier1ArithmeticSemanticFma128Count = uint64(@@TIER1_SEMANTIC_FMA128_COUNT@@)
	tier1ArithmeticSemanticSqrt32Count  = uint64(@@TIER1_SEMANTIC_SQRT32_COUNT@@)
	tier1ArithmeticSemanticSqrt64Count  = uint64(@@TIER1_SEMANTIC_SQRT64_COUNT@@)
	tier1ArithmeticSemanticSqrt128Count = uint64(@@TIER1_SEMANTIC_SQRT128_COUNT@@)
	tier1ArithmeticProbeCount       = uint64(12)
	tier1ArithmeticRoundingModes    = uint64(5)
	tier1ArithmeticRoundedOps       = uint64(5)
	tier1ArithmeticUnroundedOps     = uint64(2)
	tier1ArithmeticFmaSlotArrangements = uint64(3)
	tier1ArithmeticRandomOps        = uint64(10)
	tier1ArithmeticScaleExponents                = uint64(25)
	tier1ArithmeticScaleFiniteTransitionLimit32  = uint64(@@TIER1_SCALE_FINITE_TRANSITION_LIMIT32@@)
	tier1ArithmeticScaleFiniteTransitionLimit64  = uint64(@@TIER1_SCALE_FINITE_TRANSITION_LIMIT64@@)
	tier1ArithmeticScaleFiniteTransitionLimit128 = uint64(@@TIER1_SCALE_FINITE_TRANSITION_LIMIT128@@)
	tier1ArithmeticScaleRandomStrata             = uint64(@@TIER1_SCALE_RANDOM_STRATA@@)
	tier1ArithmeticScaleModeCross                = uint64(@@TIER1_SCALE_MODE_CROSS@@)
	tier1ArithmeticScaleModeCrossGroups32        = uint64(@@TIER1_SCALE_MODE_CROSS_GROUPS32@@)
	tier1ArithmeticScaleModeCrossGroups64        = uint64(@@TIER1_SCALE_MODE_CROSS_GROUPS64@@)
	tier1ArithmeticScaleModeCrossGroups128       = uint64(@@TIER1_SCALE_MODE_CROSS_GROUPS128@@)
	tier1ArithmeticScaleTupleHash32              = uint64(@@TIER1_SCALE_TUPLE_HASH32@@)
	tier1ArithmeticScaleTupleHash64              = uint64(@@TIER1_SCALE_TUPLE_HASH64@@)
	tier1ArithmeticScaleTupleHash128             = uint64(@@TIER1_SCALE_TUPLE_HASH128@@)
	tier1ArithmeticPairStreamHash32              = uint64(@@TIER1_PAIR_STREAM_HASH32@@)
	tier1ArithmeticPairStreamHash64              = uint64(@@TIER1_PAIR_STREAM_HASH64@@)
	tier1ArithmeticPairStreamHash128             = uint64(@@TIER1_PAIR_STREAM_HASH128@@)
	tier1ArithmeticFmaTripleStreamHash32         = uint64(@@TIER1_FMA_TRIPLE_STREAM_HASH32@@)
	tier1ArithmeticFmaTripleStreamHash64         = uint64(@@TIER1_FMA_TRIPLE_STREAM_HASH64@@)
	tier1ArithmeticFmaTripleStreamHash128        = uint64(@@TIER1_FMA_TRIPLE_STREAM_HASH128@@)
	tier1ArithmeticRandomStreamHash32            = uint64(@@TIER1_RANDOM_STREAM_HASH32@@)
	tier1ArithmeticRandomStreamHash64            = uint64(@@TIER1_RANDOM_STREAM_HASH64@@)
	tier1ArithmeticRandomStreamHash128           = uint64(@@TIER1_RANDOM_STREAM_HASH128@@)
	tier1ArithmeticScaleTupleHashOffset          = uint64(0xcbf29ce484222325)
	tier1ArithmeticScaleTupleHashPrime           = uint64(0x100000001b3)

	// The deterministic-random seeds are shared by the differential blocks and
	// the corpus stream contract, so a seed edit that reaches only one of them
	// is impossible by construction.
	tier1ArithmeticRandomRoundedSeedBase32   = uint64(0xdec7543200000000)
	tier1ArithmeticRandomUnroundedSeedBase32 = uint64(0xdec7543210000000)
	tier1ArithmeticRandomFmaSeed32           = uint64(0xdec7543220000000)
	tier1ArithmeticRandomSqrtSeed32          = uint64(0xdec7543230000000)
	tier1ArithmeticRandomRoundedSeedBase64   = uint64(0xdec7546400000000)
	tier1ArithmeticRandomUnroundedSeedBase64 = uint64(0xdec7546410000000)
	tier1ArithmeticRandomFmaSeed64           = uint64(0xdec7546420000000)
	tier1ArithmeticRandomSqrtSeed64          = uint64(0xdec7546430000000)
	tier1ArithmeticRandomRoundedSeedBase128   = uint64(0xdec7541200000000)
	tier1ArithmeticRandomUnroundedSeedBase128 = uint64(0xdec7541210000000)
	tier1ArithmeticRandomFmaSeed128           = uint64(0xdec7541220000000)
	tier1ArithmeticRandomSqrtSeed128          = uint64(0xdec7541230000000)

	tier1ArithmeticBoundaryPairs32 = tier1ArithmeticBoundary32Count*tier1ArithmeticProbeCount*2 +
		tier1ArithmeticProbeCount*tier1ArithmeticProbeCount
	tier1ArithmeticBoundaryPairs64 = tier1ArithmeticBoundary64Count*tier1ArithmeticProbeCount*2 +
		tier1ArithmeticProbeCount*tier1ArithmeticProbeCount
	tier1ArithmeticBoundaryPairs128 = tier1ArithmeticBoundary128Count*tier1ArithmeticProbeCount*2 +
		tier1ArithmeticProbeCount*tier1ArithmeticProbeCount

	tier1ArithmeticFmaTriples32 = tier1ArithmeticBoundary32Count*tier1ArithmeticProbeCount*tier1ArithmeticFmaSlotArrangements +
		tier1ArithmeticProbeCount*tier1ArithmeticProbeCount*tier1ArithmeticProbeCount
	tier1ArithmeticFmaTriples64 = tier1ArithmeticBoundary64Count*tier1ArithmeticProbeCount*tier1ArithmeticFmaSlotArrangements +
		tier1ArithmeticProbeCount*tier1ArithmeticProbeCount*tier1ArithmeticProbeCount
	tier1ArithmeticFmaTriples128 = tier1ArithmeticBoundary128Count*tier1ArithmeticProbeCount*tier1ArithmeticFmaSlotArrangements +
		tier1ArithmeticProbeCount*tier1ArithmeticProbeCount*tier1ArithmeticProbeCount

	tier1ArithmeticStructuredComparisons32 = tier1ArithmeticBoundaryPairs32*
		(tier1ArithmeticRoundedOps*tier1ArithmeticRoundingModes+tier1ArithmeticUnroundedOps) +
		(tier1ArithmeticFmaTriples32+tier1ArithmeticSemanticFma32Count)*tier1ArithmeticRoundingModes +
		(tier1ArithmeticBoundary32Count+tier1ArithmeticSemanticSqrt32Count)*tier1ArithmeticRoundingModes +
		tier1ArithmeticBoundary32Count*tier1ArithmeticScaleExponents*tier1ArithmeticRoundingModes +
		(tier1ArithmeticSemanticRounded32Count+tier1ArithmeticSemanticScale32Count)*tier1ArithmeticRoundingModes +
		tier1ArithmeticSemanticRemainder32Count*tier1ArithmeticUnroundedOps
	tier1ArithmeticStructuredComparisons64 = tier1ArithmeticBoundaryPairs64*
		(tier1ArithmeticRoundedOps*tier1ArithmeticRoundingModes+tier1ArithmeticUnroundedOps) +
		(tier1ArithmeticFmaTriples64+tier1ArithmeticSemanticFma64Count)*tier1ArithmeticRoundingModes +
		(tier1ArithmeticBoundary64Count+tier1ArithmeticSemanticSqrt64Count)*tier1ArithmeticRoundingModes +
		tier1ArithmeticBoundary64Count*tier1ArithmeticScaleExponents*tier1ArithmeticRoundingModes +
		(tier1ArithmeticSemanticRounded64Count+tier1ArithmeticSemanticScale64Count)*tier1ArithmeticRoundingModes +
		tier1ArithmeticSemanticRemainder64Count*tier1ArithmeticUnroundedOps
	tier1ArithmeticStructuredComparisons128 = tier1ArithmeticBoundaryPairs128*
		(tier1ArithmeticRoundedOps*tier1ArithmeticRoundingModes+tier1ArithmeticUnroundedOps) +
		(tier1ArithmeticFmaTriples128+tier1ArithmeticSemanticFma128Count)*tier1ArithmeticRoundingModes +
		(tier1ArithmeticBoundary128Count+tier1ArithmeticSemanticSqrt128Count)*tier1ArithmeticRoundingModes +
		tier1ArithmeticBoundary128Count*tier1ArithmeticScaleExponents*tier1ArithmeticRoundingModes +
		(tier1ArithmeticSemanticRounded128Count+tier1ArithmeticSemanticScale128Count)*tier1ArithmeticRoundingModes +
		tier1ArithmeticSemanticRemainder128Count*tier1ArithmeticUnroundedOps

	tier1ArithmeticRandomCasesPerOp32  = uint64(1) << 20
	tier1ArithmeticRandomCasesPerOp64  = uint64(1) << 20
	tier1ArithmeticRandomCasesPerOp128 = uint64(1) << 19
	tier1ArithmeticRandomComparisons32 = tier1ArithmeticRandomCasesPerOp32 * tier1ArithmeticRandomOps
	tier1ArithmeticRandomComparisons64 = tier1ArithmeticRandomCasesPerOp64 * tier1ArithmeticRandomOps
	tier1ArithmeticRandomComparisons128 = tier1ArithmeticRandomCasesPerOp128 *
		tier1ArithmeticRandomOps

	tier1ArithmeticTotalComparisons32 = tier1ArithmeticStructuredComparisons32 +
		tier1ArithmeticRandomComparisons32
	tier1ArithmeticTotalComparisons64 = tier1ArithmeticStructuredComparisons64 +
		tier1ArithmeticRandomComparisons64
	tier1ArithmeticTotalComparisons128 = tier1ArithmeticStructuredComparisons128 +
		tier1ArithmeticRandomComparisons128
)

type tier1ArithmeticMode struct {
	name   string
	public RoundingMode
	native int
}

var tier1ArithmeticModes = [...]tier1ArithmeticMode{
	{name: "nearest_even", public: RoundNearestEven, native: 0},
	{name: "nearest_away", public: RoundNearestAway, native: 4},
	{name: "toward_zero", public: RoundTowardZero, native: 3},
	{name: "toward_positive", public: RoundTowardPositive, native: 2},
	{name: "toward_negative", public: RoundTowardNegative, native: 1},
}

var tier1ArithmeticRoundedOperations = [...]string{"add", "sub", "mul", "div", "quantize"}
var tier1ArithmeticUnroundedOperations = [...]string{"remainder", "fmod"}

var tier1ArithmeticScaleExponentValues = [...]int64{
	-9223372036854775808,
	-2147483649,
	-2147483648,
	-6177,
	-6176,
	-1000,
	-399,
	-398,
	-102,
	-101,
	-1,
	0,
	1,
	90,
	91,
	369,
	370,
	398,
	399,
	1000,
	6176,
	6177,
	2147483647,
	2147483648,
	9223372036854775807,
}

type tier1ArithmeticShard struct {
	count uint64
	index uint64
}

func tier1ArithmeticLoadShard(t *testing.T) tier1ArithmeticShard {
	t.Helper()
	countText := os.Getenv("BID754_TIER1_ARITH_SHARD_COUNT")
	indexText := os.Getenv("BID754_TIER1_ARITH_SHARD_INDEX")
	if countText == "" && indexText == "" {
		return tier1ArithmeticShard{count: 1}
	}
	if countText == "" || indexText == "" {
		t.Fatal("BID754_TIER1_ARITH_SHARD_COUNT and BID754_TIER1_ARITH_SHARD_INDEX must be set together")
	}
	count, err := strconv.ParseUint(countText, 10, 64)
	if err != nil || count == 0 {
		t.Fatalf("invalid BID754_TIER1_ARITH_SHARD_COUNT %q", countText)
	}
	index, err := strconv.ParseUint(indexText, 10, 64)
	if err != nil || index >= count {
		t.Fatalf("invalid BID754_TIER1_ARITH_SHARD_INDEX %q for shard count %d", indexText, count)
	}
	return tier1ArithmeticShard{count: count, index: index}
}

func (s tier1ArithmeticShard) owns(caseIndex uint64) bool {
	return caseIndex%s.count == s.index
}

func (s tier1ArithmeticShard) ownedCount(total uint64) uint64 {
	if total <= s.index {
		return 0
	}
	return 1 + (total-1-s.index)/s.count
}

type tier1Arithmetic128Words struct {
	lo uint64
	hi uint64
}

type tier1ArithmeticSemanticRounded32 struct {
	operation string
	x         uint32
	y         uint32
}

type tier1ArithmeticSemanticRounded64 struct {
	operation string
	x         uint64
	y         uint64
}

type tier1ArithmeticSemanticRounded128 struct {
	operation string
	x         tier1Arithmetic128Words
	y         tier1Arithmetic128Words
}

type tier1ArithmeticSemanticScale32 struct {
	x        uint32
	exponent int64
}

type tier1ArithmeticSemanticScale64 struct {
	x        uint64
	exponent int64
}

type tier1ArithmeticSemanticScale128 struct {
	x        tier1Arithmetic128Words
	exponent int64
}

type tier1ArithmeticPair32 struct {
	x uint32
	y uint32
}

type tier1ArithmeticPair64 struct {
	x uint64
	y uint64
}

type tier1ArithmeticPair128 struct {
	x tier1Arithmetic128Words
	y tier1Arithmetic128Words
}

type tier1ArithmeticTriple32 struct {
	x uint32
	y uint32
	z uint32
}

type tier1ArithmeticTriple64 struct {
	x uint64
	y uint64
	z uint64
}

type tier1ArithmeticTriple128 struct {
	x tier1Arithmetic128Words
	y tier1Arithmetic128Words
	z tier1Arithmetic128Words
}

func tier1ArithmeticDecimal128(words tier1Arithmetic128Words) Decimal128BID {
	var value Decimal128BID
	binary.LittleEndian.PutUint64(value[0:8], words.lo)
	binary.LittleEndian.PutUint64(value[8:16], words.hi)
	return value
}

func tier1ArithmeticPublicRawFlags(flags ExceptionFlags) (uint32, ExceptionFlags) {
	var raw uint32
	if flags&FlagInvalidOperation != 0 {
		raw |= 0x01
	}
	if flags&FlagDivisionByZero != 0 {
		raw |= 0x04
	}
	if flags&FlagOverflow != 0 {
		raw |= 0x08
	}
	if flags&FlagUnderflow != 0 {
		raw |= 0x10
	}
	if flags&FlagInexact != 0 {
		raw |= 0x20
	}
	known := FlagInvalidOperation | FlagDivisionByZero | FlagOverflow | FlagUnderflow | FlagInexact
	return raw, flags &^ known
}

func tier1ArithmeticCheckRounded32(t *testing.T, operation string, x, y uint32, mode tier1ArithmeticMode) {
	native, port := runGeneratedFFICase32Binary("bid32_"+operation, x, y, mode.native)
	left, right := Decimal32BID(x), Decimal32BID(y)
	var got Decimal32BID
	var gotFlags ExceptionFlags
	pureChecked := true
	var pure uint32
	switch operation {
	case "add":
		got, gotFlags = left.AddWithMode(right, mode.public)
		pure = bidgo.Bid32Add(x, y, mode.native)
	case "sub":
		got, gotFlags = left.SubWithMode(right, mode.public)
		pure = bidgo.Bid32Sub(x, y, mode.native)
	case "mul":
		got, gotFlags = left.MulWithMode(right, mode.public)
		pure = bidgo.Bid32Mul(x, y, mode.native)
	case "div":
		got, gotFlags = left.DivWithMode(right, mode.public)
		pure = bidgo.Bid32Div(x, y, mode.native)
	case "quantize":
		got, gotFlags = left.QuantizeWithMode(right, mode.public)
		pureChecked = false
	default:
		t.Fatalf("decimal32 unknown rounded Tier 1 operation %q", operation)
	}
	publicRawFlags, unknownPublicFlags := tier1ArithmeticPublicRawFlags(gotFlags)
	public := fmt.Sprintf("%08x/%08x", got.ToUint32(), publicRawFlags)
	if unknownPublicFlags != 0 || native != port || native != public {
		t.Fatalf("decimal32 %s mismatch: x=%08x y=%08x mode=%s C=%s port=%s public=%s unknown_public_flags=%s", operation, x, y, mode.name, native, port, public, unknownPublicFlags)
	}
	if pureChecked && !strings.HasPrefix(native, fmt.Sprintf("%08x/", pure)) {
		t.Fatalf("decimal32 %s value-only port mismatch: x=%08x y=%08x mode=%s C=%s pure=%08x", operation, x, y, mode.name, native, pure)
	}
}

func tier1ArithmeticCheckFma32(t *testing.T, x, y, z uint32, mode tier1ArithmeticMode) {
	native, port := runGeneratedFFICase32Ternary("bid32_fma", x, y, z, mode.native)
	got, gotFlags := Decimal32BID(x).FMAWithMode(Decimal32BID(y), Decimal32BID(z), mode.public)
	publicRawFlags, unknownPublicFlags := tier1ArithmeticPublicRawFlags(gotFlags)
	public := fmt.Sprintf("%08x/%08x", got.ToUint32(), publicRawFlags)
	if unknownPublicFlags != 0 || native != port || native != public {
		t.Fatalf("decimal32 fma mismatch: x=%08x y=%08x z=%08x mode=%s C=%s port=%s public=%s unknown_public_flags=%s", x, y, z, mode.name, native, port, public, unknownPublicFlags)
	}
}

func tier1ArithmeticCheckSqrt32(t *testing.T, x uint32, mode tier1ArithmeticMode) {
	native, port := runGeneratedFFICase32Unary("bid32_sqrt", x, mode.native)
	got, gotFlags := Decimal32BID(x).SqrtWithMode(mode.public)
	publicRawFlags, unknownPublicFlags := tier1ArithmeticPublicRawFlags(gotFlags)
	public := fmt.Sprintf("%08x/%08x", got.ToUint32(), publicRawFlags)
	if unknownPublicFlags != 0 || native != port || native != public {
		t.Fatalf("decimal32 sqrt mismatch: x=%08x mode=%s C=%s port=%s public=%s unknown_public_flags=%s", x, mode.name, native, port, public, unknownPublicFlags)
	}
}

func tier1ArithmeticCheckUnrounded32(t *testing.T, operation string, x, y uint32) {
	function := "bid32_rem"
	if operation == "fmod" {
		function = "bid32_fmod"
	} else if operation != "remainder" {
		t.Fatalf("decimal32 unknown unrounded Tier 1 operation %q", operation)
	}
	native, port := runGeneratedFFICase32Binary(function, x, y, 0)
	left, right := Decimal32BID(x), Decimal32BID(y)
	var got Decimal32BID
	var gotFlags ExceptionFlags
	if operation == "remainder" {
		got, gotFlags = left.Remainder(right)
	} else {
		got, gotFlags = left.Fmod(right)
	}
	publicRawFlags, unknownPublicFlags := tier1ArithmeticPublicRawFlags(gotFlags)
	public := fmt.Sprintf("%08x/%08x", got.ToUint32(), publicRawFlags)
	if unknownPublicFlags != 0 || native != port || native != public {
		t.Fatalf("decimal32 %s mismatch: x=%08x y=%08x C=%s port=%s public=%s unknown_public_flags=%s", operation, x, y, native, port, public, unknownPublicFlags)
	}
}

func tier1ArithmeticCheckScale32(t *testing.T, x uint32, exponent int64, mode tier1ArithmeticMode) {
	native, nativeFlags, port, portFlags := runGeneratedFFICase32DecimalFlagInt("scalbln", x, int(exponent), mode.native)
	got, gotFlags := Decimal32BID(x).ScaleBWithMode(int(exponent), mode.public)
	publicRawFlags, unknownPublicFlags := tier1ArithmeticPublicRawFlags(gotFlags)
	if unknownPublicFlags != 0 || native != port || nativeFlags != portFlags || got.ToUint32() != native || publicRawFlags != nativeFlags {
		t.Fatalf("decimal32 scaleB mismatch: x=%08x exponent=%d mode=%s C=%08x/%08x port=%08x/%08x public=%08x/%08x unknown_public_flags=%s", x, exponent, mode.name, native, nativeFlags, port, portFlags, got.ToUint32(), publicRawFlags, unknownPublicFlags)
	}
}

func tier1ArithmeticCheckRounded64(t *testing.T, operation string, x, y uint64, mode tier1ArithmeticMode) {
	native, port := runGeneratedFFICase64Binary("bid64_"+operation, x, y, mode.native)
	left, right := Decimal64BID(x), Decimal64BID(y)
	var got Decimal64BID
	var gotFlags ExceptionFlags
	pureChecked := true
	var pure uint64
	switch operation {
	case "add":
		got, gotFlags = left.AddWithMode(right, mode.public)
		pure = bidgo.Bid64Add(x, y, mode.native)
	case "sub":
		got, gotFlags = left.SubWithMode(right, mode.public)
		pure = bidgo.Bid64Sub(x, y, mode.native)
	case "mul":
		got, gotFlags = left.MulWithMode(right, mode.public)
		pure = bidgo.Bid64Mul(x, y, mode.native)
	case "div":
		got, gotFlags = left.DivWithMode(right, mode.public)
		pure = bidgo.Bid64Div(x, y, mode.native)
	case "quantize":
		got, gotFlags = left.QuantizeWithMode(right, mode.public)
		pureChecked = false
	default:
		t.Fatalf("decimal64 unknown rounded Tier 1 operation %q", operation)
	}
	publicRawFlags, unknownPublicFlags := tier1ArithmeticPublicRawFlags(gotFlags)
	public := fmt.Sprintf("%016x/%08x", got.ToUint64(), publicRawFlags)
	if unknownPublicFlags != 0 || native != port || native != public {
		t.Fatalf("decimal64 %s mismatch: x=%016x y=%016x mode=%s C=%s port=%s public=%s unknown_public_flags=%s", operation, x, y, mode.name, native, port, public, unknownPublicFlags)
	}
	if pureChecked && !strings.HasPrefix(native, fmt.Sprintf("%016x/", pure)) {
		t.Fatalf("decimal64 %s value-only port mismatch: x=%016x y=%016x mode=%s C=%s pure=%016x", operation, x, y, mode.name, native, pure)
	}
}

func tier1ArithmeticCheckFma64(t *testing.T, x, y, z uint64, mode tier1ArithmeticMode) {
	native, port := runGeneratedFFICase64Ternary("bid64_fma", x, y, z, mode.native)
	got, gotFlags := Decimal64BID(x).FMAWithMode(Decimal64BID(y), Decimal64BID(z), mode.public)
	publicRawFlags, unknownPublicFlags := tier1ArithmeticPublicRawFlags(gotFlags)
	public := fmt.Sprintf("%016x/%08x", got.ToUint64(), publicRawFlags)
	if unknownPublicFlags != 0 || native != port || native != public {
		t.Fatalf("decimal64 fma mismatch: x=%016x y=%016x z=%016x mode=%s C=%s port=%s public=%s unknown_public_flags=%s", x, y, z, mode.name, native, port, public, unknownPublicFlags)
	}
}

func tier1ArithmeticCheckSqrt64(t *testing.T, x uint64, mode tier1ArithmeticMode) {
	native, port := runGeneratedFFICase64Unary("bid64_sqrt", x, mode.native)
	got, gotFlags := Decimal64BID(x).SqrtWithMode(mode.public)
	publicRawFlags, unknownPublicFlags := tier1ArithmeticPublicRawFlags(gotFlags)
	public := fmt.Sprintf("%016x/%08x", got.ToUint64(), publicRawFlags)
	if unknownPublicFlags != 0 || native != port || native != public {
		t.Fatalf("decimal64 sqrt mismatch: x=%016x mode=%s C=%s port=%s public=%s unknown_public_flags=%s", x, mode.name, native, port, public, unknownPublicFlags)
	}
}

func tier1ArithmeticCheckUnrounded64(t *testing.T, operation string, x, y uint64) {
	function := "bid64_rem"
	if operation == "fmod" {
		function = "bid64_fmod"
	} else if operation != "remainder" {
		t.Fatalf("decimal64 unknown unrounded Tier 1 operation %q", operation)
	}
	native, port := runGeneratedFFICase64Binary(function, x, y, 0)
	left, right := Decimal64BID(x), Decimal64BID(y)
	var got Decimal64BID
	var gotFlags ExceptionFlags
	if operation == "remainder" {
		got, gotFlags = left.Remainder(right)
	} else {
		got, gotFlags = left.Fmod(right)
	}
	publicRawFlags, unknownPublicFlags := tier1ArithmeticPublicRawFlags(gotFlags)
	public := fmt.Sprintf("%016x/%08x", got.ToUint64(), publicRawFlags)
	if unknownPublicFlags != 0 || native != port || native != public {
		t.Fatalf("decimal64 %s mismatch: x=%016x y=%016x C=%s port=%s public=%s unknown_public_flags=%s", operation, x, y, native, port, public, unknownPublicFlags)
	}
}

func tier1ArithmeticCheckScale64(t *testing.T, x uint64, exponent int64, mode tier1ArithmeticMode) {
	native, nativeFlags, port, portFlags := runGeneratedFFICase64DecimalFlagInt("scalbln", x, int(exponent), mode.native)
	got, gotFlags := Decimal64BID(x).ScaleBWithMode(int(exponent), mode.public)
	publicRawFlags, unknownPublicFlags := tier1ArithmeticPublicRawFlags(gotFlags)
	if unknownPublicFlags != 0 || native != port || nativeFlags != portFlags || got.ToUint64() != native || publicRawFlags != nativeFlags {
		t.Fatalf("decimal64 scaleB mismatch: x=%016x exponent=%d mode=%s C=%016x/%08x port=%016x/%08x public=%016x/%08x unknown_public_flags=%s", x, exponent, mode.name, native, nativeFlags, port, portFlags, got.ToUint64(), publicRawFlags, unknownPublicFlags)
	}
}

func tier1ArithmeticCheckRounded128(t *testing.T, operation string, x, y Decimal128BID, mode tier1ArithmeticMode) {
	native, port := runGeneratedFFICase128Binary("bid128_"+operation, x, y, mode.native)
	var got Decimal128BID
	var gotFlags ExceptionFlags
	switch operation {
	case "add":
		got, gotFlags = x.AddWithMode(y, mode.public)
	case "sub":
		got, gotFlags = x.SubWithMode(y, mode.public)
	case "mul":
		got, gotFlags = x.MulWithMode(y, mode.public)
	case "div":
		got, gotFlags = x.DivWithMode(y, mode.public)
	case "quantize":
		got, gotFlags = x.QuantizeWithMode(y, mode.public)
	default:
		t.Fatalf("decimal128 unknown rounded Tier 1 operation %q", operation)
	}
	publicRawFlags, unknownPublicFlags := tier1ArithmeticPublicRawFlags(gotFlags)
	public := fmt.Sprintf("%s/%08x", formatFFIUint128Bits(got), publicRawFlags)
	if unknownPublicFlags != 0 || native != port || native != public {
		t.Fatalf("decimal128 %s mismatch: x=%x y=%x mode=%s C=%s port=%s public=%s unknown_public_flags=%s", operation, x, y, mode.name, native, port, public, unknownPublicFlags)
	}
}

func tier1ArithmeticCheckFma128(t *testing.T, x, y, z Decimal128BID, mode tier1ArithmeticMode) {
	native, port := runGeneratedFFICase128Ternary("bid128_fma", x, y, z, mode.native)
	got, gotFlags := x.FMAWithMode(y, z, mode.public)
	publicRawFlags, unknownPublicFlags := tier1ArithmeticPublicRawFlags(gotFlags)
	public := fmt.Sprintf("%s/%08x", formatFFIUint128Bits(got), publicRawFlags)
	if unknownPublicFlags != 0 || native != port || native != public {
		t.Fatalf("decimal128 fma mismatch: x=%x y=%x z=%x mode=%s C=%s port=%s public=%s unknown_public_flags=%s", x, y, z, mode.name, native, port, public, unknownPublicFlags)
	}
}

func tier1ArithmeticCheckSqrt128(t *testing.T, x Decimal128BID, mode tier1ArithmeticMode) {
	native, port := runGeneratedFFICase128Unary("bid128_sqrt", x, mode.native)
	got, gotFlags := x.SqrtWithMode(mode.public)
	publicRawFlags, unknownPublicFlags := tier1ArithmeticPublicRawFlags(gotFlags)
	public := fmt.Sprintf("%s/%08x", formatFFIUint128Bits(got), publicRawFlags)
	if unknownPublicFlags != 0 || native != port || native != public {
		t.Fatalf("decimal128 sqrt mismatch: x=%x mode=%s C=%s port=%s public=%s unknown_public_flags=%s", x, mode.name, native, port, public, unknownPublicFlags)
	}
}

func tier1ArithmeticCheckUnrounded128(t *testing.T, operation string, x, y Decimal128BID) {
	function := "bid128_rem"
	if operation == "fmod" {
		function = "bid128_fmod"
	} else if operation != "remainder" {
		t.Fatalf("decimal128 unknown unrounded Tier 1 operation %q", operation)
	}
	native, port := runGeneratedFFICase128Binary(function, x, y, 0)
	var got Decimal128BID
	var gotFlags ExceptionFlags
	if operation == "remainder" {
		got, gotFlags = x.Remainder(y)
	} else {
		got, gotFlags = x.Fmod(y)
	}
	publicRawFlags, unknownPublicFlags := tier1ArithmeticPublicRawFlags(gotFlags)
	public := fmt.Sprintf("%s/%08x", formatFFIUint128Bits(got), publicRawFlags)
	if unknownPublicFlags != 0 || native != port || native != public {
		t.Fatalf("decimal128 %s mismatch: x=%x y=%x C=%s port=%s public=%s unknown_public_flags=%s", operation, x, y, native, port, public, unknownPublicFlags)
	}
}

func tier1ArithmeticCheckScale128(t *testing.T, x Decimal128BID, exponent int64, mode tier1ArithmeticMode) {
	native, nativeFlags, port, portFlags := runGeneratedFFICase128DecimalFlagInt("scalbln", x, int(exponent), mode.native)
	got, gotFlags := x.ScaleBWithMode(int(exponent), mode.public)
	publicRawFlags, unknownPublicFlags := tier1ArithmeticPublicRawFlags(gotFlags)
	if unknownPublicFlags != 0 || native != port || nativeFlags != portFlags || got != native || publicRawFlags != nativeFlags {
		t.Fatalf("decimal128 scaleB mismatch: x=%x exponent=%d mode=%s C=%x/%08x port=%x/%08x public=%x/%08x unknown_public_flags=%s", x, exponent, mode.name, native, nativeFlags, port, portFlags, got, publicRawFlags, unknownPublicFlags)
	}
}

func tier1ArithmeticVisitPairs32(visit func(uint32, uint32)) {
	for _, x := range tier1ArithmeticBoundary32 {
		for _, y := range tier1ArithmeticProbes32 {
			visit(x, y)
			visit(y, x)
		}
	}
	for _, x := range tier1ArithmeticProbes32 {
		for _, y := range tier1ArithmeticProbes32 {
			visit(x, y)
		}
	}
}

func tier1ArithmeticVisitPairs64(visit func(uint64, uint64)) {
	for _, x := range tier1ArithmeticBoundary64 {
		for _, y := range tier1ArithmeticProbes64 {
			visit(x, y)
			visit(y, x)
		}
	}
	for _, x := range tier1ArithmeticProbes64 {
		for _, y := range tier1ArithmeticProbes64 {
			visit(x, y)
		}
	}
}

func tier1ArithmeticVisitPairs128(visit func(tier1Arithmetic128Words, tier1Arithmetic128Words)) {
	for _, x := range tier1ArithmeticBoundary128 {
		for _, y := range tier1ArithmeticProbes128 {
			visit(x, y)
			visit(y, x)
		}
	}
	for _, x := range tier1ArithmeticProbes128 {
		for _, y := range tier1ArithmeticProbes128 {
			visit(x, y)
		}
	}
}

func tier1ArithmeticVisitTriples32(visit func(uint32, uint32, uint32)) {
	for i, x := range tier1ArithmeticBoundary32 {
		for j, y := range tier1ArithmeticProbes32 {
			z := tier1ArithmeticProbes32[(i+j)%len(tier1ArithmeticProbes32)]
			visit(x, y, z)
			visit(y, z, x)
			visit(z, x, y)
		}
	}
	for _, x := range tier1ArithmeticProbes32 {
		for _, y := range tier1ArithmeticProbes32 {
			for _, z := range tier1ArithmeticProbes32 {
				visit(x, y, z)
			}
		}
	}
}

func tier1ArithmeticVisitTriples64(visit func(uint64, uint64, uint64)) {
	for i, x := range tier1ArithmeticBoundary64 {
		for j, y := range tier1ArithmeticProbes64 {
			z := tier1ArithmeticProbes64[(i+j)%len(tier1ArithmeticProbes64)]
			visit(x, y, z)
			visit(y, z, x)
			visit(z, x, y)
		}
	}
	for _, x := range tier1ArithmeticProbes64 {
		for _, y := range tier1ArithmeticProbes64 {
			for _, z := range tier1ArithmeticProbes64 {
				visit(x, y, z)
			}
		}
	}
}

func tier1ArithmeticVisitTriples128(visit func(tier1Arithmetic128Words, tier1Arithmetic128Words, tier1Arithmetic128Words)) {
	for i, x := range tier1ArithmeticBoundary128 {
		for j, y := range tier1ArithmeticProbes128 {
			z := tier1ArithmeticProbes128[(i+j)%len(tier1ArithmeticProbes128)]
			visit(x, y, z)
			visit(y, z, x)
			visit(z, x, y)
		}
	}
	for _, x := range tier1ArithmeticProbes128 {
		for _, y := range tier1ArithmeticProbes128 {
			for _, z := range tier1ArithmeticProbes128 {
				visit(x, y, z)
			}
		}
	}
}

func TestTier1ArithmeticStructuredNativeDifferential(t *testing.T) {
	requireNative(t)
	if strconv.IntSize != 64 {
		t.Fatalf("Tier 1 arithmetic native oracle requires the guaranteed LP64 platform contract; int size=%d", strconv.IntSize)
	}
	if len(tier1ArithmeticBoundary32) != int(tier1ArithmeticBoundary32Count) ||
		len(tier1ArithmeticBoundary64) != int(tier1ArithmeticBoundary64Count) ||
		len(tier1ArithmeticBoundary128) != int(tier1ArithmeticBoundary128Count) {
		t.Fatalf("generated Tier 1 boundary inventory count mismatch: d32=%d d64=%d d128=%d", len(tier1ArithmeticBoundary32), len(tier1ArithmeticBoundary64), len(tier1ArithmeticBoundary128))
	}
	if len(tier1ArithmeticSemanticRounded32Cases) != int(tier1ArithmeticSemanticRounded32Count) ||
		len(tier1ArithmeticSemanticRounded64Cases) != int(tier1ArithmeticSemanticRounded64Count) ||
		len(tier1ArithmeticSemanticRounded128Cases) != int(tier1ArithmeticSemanticRounded128Count) ||
		len(tier1ArithmeticSemanticScale32Cases) != int(tier1ArithmeticSemanticScale32Count) ||
		len(tier1ArithmeticSemanticScale64Cases) != int(tier1ArithmeticSemanticScale64Count) ||
		len(tier1ArithmeticSemanticScale128Cases) != int(tier1ArithmeticSemanticScale128Count) ||
		len(tier1ArithmeticSemanticRemainder32Pairs) != int(tier1ArithmeticSemanticRemainder32Count) ||
		len(tier1ArithmeticSemanticRemainder64Pairs) != int(tier1ArithmeticSemanticRemainder64Count) ||
		len(tier1ArithmeticSemanticRemainder128Pairs) != int(tier1ArithmeticSemanticRemainder128Count) ||
		len(tier1ArithmeticSemanticFma32Cases) != int(tier1ArithmeticSemanticFma32Count) ||
		len(tier1ArithmeticSemanticFma64Cases) != int(tier1ArithmeticSemanticFma64Count) ||
		len(tier1ArithmeticSemanticFma128Cases) != int(tier1ArithmeticSemanticFma128Count) ||
		len(tier1ArithmeticSemanticSqrt32Cases) != int(tier1ArithmeticSemanticSqrt32Count) ||
		len(tier1ArithmeticSemanticSqrt64Cases) != int(tier1ArithmeticSemanticSqrt64Count) ||
		len(tier1ArithmeticSemanticSqrt128Cases) != int(tier1ArithmeticSemanticSqrt128Count) {
		t.Fatal("generated Tier 1 semantic rounding-discriminant inventory count mismatch")
	}
	if len(tier1ArithmeticProbes32) != int(tier1ArithmeticProbeCount) ||
		len(tier1ArithmeticProbes64) != int(tier1ArithmeticProbeCount) ||
		len(tier1ArithmeticProbes128) != int(tier1ArithmeticProbeCount) ||
		len(tier1ArithmeticScaleExponentValues) != int(tier1ArithmeticScaleExponents) {
		t.Fatal("generated Tier 1 probe or scale-exponent inventory count mismatch")
	}
	shard := tier1ArithmeticLoadShard(t)

	t.Run("decimal32", func(t *testing.T) {
		var caseIndex uint64
		for _, operation := range tier1ArithmeticRoundedOperations {
			tier1ArithmeticVisitPairs32(func(x, y uint32) {
				for _, mode := range tier1ArithmeticModes {
					if shard.owns(caseIndex) {
						tier1ArithmeticCheckRounded32(t, operation, x, y, mode)
					}
					caseIndex++
				}
			})
		}
		for _, tc := range tier1ArithmeticSemanticRounded32Cases {
			for _, mode := range tier1ArithmeticModes {
				if shard.owns(caseIndex) {
					tier1ArithmeticCheckRounded32(t, tc.operation, tc.x, tc.y, mode)
				}
				caseIndex++
			}
		}
		tier1ArithmeticVisitTriples32(func(x, y, z uint32) {
			for _, mode := range tier1ArithmeticModes {
				if shard.owns(caseIndex) {
					tier1ArithmeticCheckFma32(t, x, y, z, mode)
				}
				caseIndex++
			}
		})
		for _, tc := range tier1ArithmeticSemanticFma32Cases {
			for _, mode := range tier1ArithmeticModes {
				if shard.owns(caseIndex) {
					tier1ArithmeticCheckFma32(t, tc.x, tc.y, tc.z, mode)
				}
				caseIndex++
			}
		}
		for _, operation := range tier1ArithmeticUnroundedOperations {
			tier1ArithmeticVisitPairs32(func(x, y uint32) {
				if shard.owns(caseIndex) {
					tier1ArithmeticCheckUnrounded32(t, operation, x, y)
				}
				caseIndex++
			})
		}
		for _, tc := range tier1ArithmeticSemanticRemainder32Pairs {
			for _, operation := range tier1ArithmeticUnroundedOperations {
				if shard.owns(caseIndex) {
					tier1ArithmeticCheckUnrounded32(t, operation, tc.x, tc.y)
				}
				caseIndex++
			}
			if shard.count == 1 {
				remainder, _ := Decimal32BID(tc.x).Remainder(Decimal32BID(tc.y))
				fmod, _ := Decimal32BID(tc.x).Fmod(Decimal32BID(tc.y))
				if remainder == fmod {
					t.Fatalf("decimal32 remainder/fmod semantic discriminator collapsed: x=%08x y=%08x result=%08x", tc.x, tc.y, remainder.ToUint32())
				}
			}
		}
		for _, x := range tier1ArithmeticBoundary32 {
			for _, mode := range tier1ArithmeticModes {
				if shard.owns(caseIndex) {
					tier1ArithmeticCheckSqrt32(t, x, mode)
				}
				caseIndex++
			}
		}
		for _, x := range tier1ArithmeticSemanticSqrt32Cases {
			for _, mode := range tier1ArithmeticModes {
				if shard.owns(caseIndex) {
					tier1ArithmeticCheckSqrt32(t, x, mode)
				}
				caseIndex++
			}
		}
		for _, x := range tier1ArithmeticBoundary32 {
			for _, exponent := range tier1ArithmeticScaleExponentValues {
				for _, mode := range tier1ArithmeticModes {
					if shard.owns(caseIndex) {
						tier1ArithmeticCheckScale32(t, x, exponent, mode)
					}
					caseIndex++
				}
			}
		}
		for _, tc := range tier1ArithmeticSemanticScale32Cases {
			for _, mode := range tier1ArithmeticModes {
				if shard.owns(caseIndex) {
					tier1ArithmeticCheckScale32(t, tc.x, tc.exponent, mode)
				}
				caseIndex++
			}
		}
		if caseIndex != tier1ArithmeticStructuredComparisons32 {
			t.Fatalf("decimal32 structured comparisons=%d want=%d", caseIndex, tier1ArithmeticStructuredComparisons32)
		}
		t.Logf("decimal32 structured exact comparisons: %d/%d", shard.ownedCount(caseIndex), caseIndex)
	})

	t.Run("decimal64", func(t *testing.T) {
		var caseIndex uint64
		for _, operation := range tier1ArithmeticRoundedOperations {
			tier1ArithmeticVisitPairs64(func(x, y uint64) {
				for _, mode := range tier1ArithmeticModes {
					if shard.owns(caseIndex) {
						tier1ArithmeticCheckRounded64(t, operation, x, y, mode)
					}
					caseIndex++
				}
			})
		}
		for _, tc := range tier1ArithmeticSemanticRounded64Cases {
			for _, mode := range tier1ArithmeticModes {
				if shard.owns(caseIndex) {
					tier1ArithmeticCheckRounded64(t, tc.operation, tc.x, tc.y, mode)
				}
				caseIndex++
			}
		}
		tier1ArithmeticVisitTriples64(func(x, y, z uint64) {
			for _, mode := range tier1ArithmeticModes {
				if shard.owns(caseIndex) {
					tier1ArithmeticCheckFma64(t, x, y, z, mode)
				}
				caseIndex++
			}
		})
		for _, tc := range tier1ArithmeticSemanticFma64Cases {
			for _, mode := range tier1ArithmeticModes {
				if shard.owns(caseIndex) {
					tier1ArithmeticCheckFma64(t, tc.x, tc.y, tc.z, mode)
				}
				caseIndex++
			}
		}
		for _, operation := range tier1ArithmeticUnroundedOperations {
			tier1ArithmeticVisitPairs64(func(x, y uint64) {
				if shard.owns(caseIndex) {
					tier1ArithmeticCheckUnrounded64(t, operation, x, y)
				}
				caseIndex++
			})
		}
		for _, tc := range tier1ArithmeticSemanticRemainder64Pairs {
			for _, operation := range tier1ArithmeticUnroundedOperations {
				if shard.owns(caseIndex) {
					tier1ArithmeticCheckUnrounded64(t, operation, tc.x, tc.y)
				}
				caseIndex++
			}
			if shard.count == 1 {
				remainder, _ := Decimal64BID(tc.x).Remainder(Decimal64BID(tc.y))
				fmod, _ := Decimal64BID(tc.x).Fmod(Decimal64BID(tc.y))
				if remainder == fmod {
					t.Fatalf("decimal64 remainder/fmod semantic discriminator collapsed: x=%016x y=%016x result=%016x", tc.x, tc.y, remainder.ToUint64())
				}
			}
		}
		for _, x := range tier1ArithmeticBoundary64 {
			for _, mode := range tier1ArithmeticModes {
				if shard.owns(caseIndex) {
					tier1ArithmeticCheckSqrt64(t, x, mode)
				}
				caseIndex++
			}
		}
		for _, x := range tier1ArithmeticSemanticSqrt64Cases {
			for _, mode := range tier1ArithmeticModes {
				if shard.owns(caseIndex) {
					tier1ArithmeticCheckSqrt64(t, x, mode)
				}
				caseIndex++
			}
		}
		for _, x := range tier1ArithmeticBoundary64 {
			for _, exponent := range tier1ArithmeticScaleExponentValues {
				for _, mode := range tier1ArithmeticModes {
					if shard.owns(caseIndex) {
						tier1ArithmeticCheckScale64(t, x, exponent, mode)
					}
					caseIndex++
				}
			}
		}
		for _, tc := range tier1ArithmeticSemanticScale64Cases {
			for _, mode := range tier1ArithmeticModes {
				if shard.owns(caseIndex) {
					tier1ArithmeticCheckScale64(t, tc.x, tc.exponent, mode)
				}
				caseIndex++
			}
		}
		if caseIndex != tier1ArithmeticStructuredComparisons64 {
			t.Fatalf("decimal64 structured comparisons=%d want=%d", caseIndex, tier1ArithmeticStructuredComparisons64)
		}
		t.Logf("decimal64 structured exact comparisons: %d/%d", shard.ownedCount(caseIndex), caseIndex)
	})

	t.Run("decimal128", func(t *testing.T) {
		var caseIndex uint64
		for _, operation := range tier1ArithmeticRoundedOperations {
			tier1ArithmeticVisitPairs128(func(x, y tier1Arithmetic128Words) {
				for _, mode := range tier1ArithmeticModes {
					if shard.owns(caseIndex) {
						tier1ArithmeticCheckRounded128(t, operation, tier1ArithmeticDecimal128(x), tier1ArithmeticDecimal128(y), mode)
					}
					caseIndex++
				}
			})
		}
		for _, tc := range tier1ArithmeticSemanticRounded128Cases {
			for _, mode := range tier1ArithmeticModes {
				if shard.owns(caseIndex) {
					tier1ArithmeticCheckRounded128(t, tc.operation, tier1ArithmeticDecimal128(tc.x), tier1ArithmeticDecimal128(tc.y), mode)
				}
				caseIndex++
			}
		}
		tier1ArithmeticVisitTriples128(func(x, y, z tier1Arithmetic128Words) {
			for _, mode := range tier1ArithmeticModes {
				if shard.owns(caseIndex) {
					tier1ArithmeticCheckFma128(t, tier1ArithmeticDecimal128(x), tier1ArithmeticDecimal128(y), tier1ArithmeticDecimal128(z), mode)
				}
				caseIndex++
			}
		})
		for _, tc := range tier1ArithmeticSemanticFma128Cases {
			for _, mode := range tier1ArithmeticModes {
				if shard.owns(caseIndex) {
					tier1ArithmeticCheckFma128(t, tier1ArithmeticDecimal128(tc.x), tier1ArithmeticDecimal128(tc.y), tier1ArithmeticDecimal128(tc.z), mode)
				}
				caseIndex++
			}
		}
		for _, operation := range tier1ArithmeticUnroundedOperations {
			tier1ArithmeticVisitPairs128(func(x, y tier1Arithmetic128Words) {
				if shard.owns(caseIndex) {
					tier1ArithmeticCheckUnrounded128(t, operation, tier1ArithmeticDecimal128(x), tier1ArithmeticDecimal128(y))
				}
				caseIndex++
			})
		}
		for _, tc := range tier1ArithmeticSemanticRemainder128Pairs {
			x := tier1ArithmeticDecimal128(tc.x)
			y := tier1ArithmeticDecimal128(tc.y)
			for _, operation := range tier1ArithmeticUnroundedOperations {
				if shard.owns(caseIndex) {
					tier1ArithmeticCheckUnrounded128(t, operation, x, y)
				}
				caseIndex++
			}
			if shard.count == 1 {
				remainder, _ := x.Remainder(y)
				fmod, _ := x.Fmod(y)
				if remainder == fmod {
					t.Fatalf("decimal128 remainder/fmod semantic discriminator collapsed: x=%x y=%x result=%x", x, y, remainder)
				}
			}
		}
		for _, words := range tier1ArithmeticBoundary128 {
			x := tier1ArithmeticDecimal128(words)
			for _, mode := range tier1ArithmeticModes {
				if shard.owns(caseIndex) {
					tier1ArithmeticCheckSqrt128(t, x, mode)
				}
				caseIndex++
			}
		}
		for _, words := range tier1ArithmeticSemanticSqrt128Cases {
			x := tier1ArithmeticDecimal128(words)
			for _, mode := range tier1ArithmeticModes {
				if shard.owns(caseIndex) {
					tier1ArithmeticCheckSqrt128(t, x, mode)
				}
				caseIndex++
			}
		}
		for _, words := range tier1ArithmeticBoundary128 {
			x := tier1ArithmeticDecimal128(words)
			for _, exponent := range tier1ArithmeticScaleExponentValues {
				for _, mode := range tier1ArithmeticModes {
					if shard.owns(caseIndex) {
						tier1ArithmeticCheckScale128(t, x, exponent, mode)
					}
					caseIndex++
				}
			}
		}
		for _, tc := range tier1ArithmeticSemanticScale128Cases {
			for _, mode := range tier1ArithmeticModes {
				if shard.owns(caseIndex) {
					tier1ArithmeticCheckScale128(t, tier1ArithmeticDecimal128(tc.x), tc.exponent, mode)
				}
				caseIndex++
			}
		}
		if caseIndex != tier1ArithmeticStructuredComparisons128 {
			t.Fatalf("decimal128 structured comparisons=%d want=%d", caseIndex, tier1ArithmeticStructuredComparisons128)
		}
		t.Logf("decimal128 structured exact comparisons: %d/%d", shard.ownedCount(caseIndex), caseIndex)
	})
}

func tier1ArithmeticSplitMix64(value uint64) uint64 {
	value += 0x9e3779b97f4a7c15
	value = (value ^ (value >> 30)) * 0xbf58476d1ce4e5b9
	value = (value ^ (value >> 27)) * 0x94d049bb133111eb
	return value ^ (value >> 31)
}

func tier1ArithmeticRandomWord(seed, caseIndex, lane uint64) uint64 {
	return tier1ArithmeticSplitMix64(seed ^ caseIndex*0xd1342543de82ef95 ^ lane*0x9e3779b97f4a7c15)
}

// tier1ArithmeticRandomOperand* derive the deterministic-random operands for
// one logical operand slot. The differential blocks and the corpus stream
// contract both consume these helpers, so a lane or truncation drift cannot
// reach only one of them.
func tier1ArithmeticRandomOperand32(seed, caseIndex, lane uint64) uint32 {
	return uint32(tier1ArithmeticRandomWord(seed, caseIndex, lane))
}

func tier1ArithmeticRandomOperand64(seed, caseIndex, lane uint64) uint64 {
	return tier1ArithmeticRandomWord(seed, caseIndex, lane)
}

func tier1ArithmeticRandomOperand128(seed, caseIndex, lane uint64) tier1Arithmetic128Words {
	return tier1Arithmetic128Words{
		lo: tier1ArithmeticRandomWord(seed, caseIndex, lane*2),
		hi: tier1ArithmeticRandomWord(seed, caseIndex, lane*2+1),
	}
}

func tier1ArithmeticScaleModeCrossGroups(transitionLimit int64) uint64 {
	return uint64(transitionLimit*2 + 1)
}

func tier1ArithmeticScaleModeCrossCase(caseIndex uint64, transitionLimit int64) bool {
	if caseIndex%tier1ArithmeticScaleRandomStrata != 1 {
		return false
	}
	slot := caseIndex / tier1ArithmeticScaleRandomStrata
	return slot < tier1ArithmeticScaleModeCrossGroups(transitionLimit)*tier1ArithmeticScaleModeCross
}

func tier1ArithmeticScaleModeCrossGroup(caseIndex uint64) uint64 {
	return (caseIndex / tier1ArithmeticScaleRandomStrata) / tier1ArithmeticScaleModeCross
}

func tier1ArithmeticRandomScaleExponent(seed, caseIndex, lane uint64, transitionLimit int64) int64 {
	switch caseIndex % tier1ArithmeticScaleRandomStrata {
	case 0:
		return int64(tier1ArithmeticRandomWord(seed, caseIndex, lane))
	case 1:
		if tier1ArithmeticScaleModeCrossCase(caseIndex, transitionLimit) {
			return int64(tier1ArithmeticScaleModeCrossGroup(caseIndex)) - transitionLimit
		}
		return int64(tier1ArithmeticRandomWord(seed, caseIndex, lane)%uint64(transitionLimit*2+1)) - transitionLimit
	case 2:
		return int64(tier1ArithmeticRandomWord(seed, caseIndex, lane)%uint64(transitionLimit*4+1)) - transitionLimit*2
	default:
		return int64(tier1ArithmeticRandomWord(seed, caseIndex, lane)%uint64(transitionLimit*2+1)) - transitionLimit
	}
}

func tier1ArithmeticRandomScaleOperandIndex(caseIndex uint64, transitionLimit int64) uint64 {
	if tier1ArithmeticScaleModeCrossCase(caseIndex, transitionLimit) {
		return tier1ArithmeticScaleModeCrossGroup(caseIndex)
	}
	return caseIndex
}

func tier1ArithmeticScaleModeCrossSign(group uint64) uint64 {
	return group & 1
}

func tier1ArithmeticRandomScaleOperand32(seed, caseIndex uint64, transitionLimit int64) uint32 {
	if tier1ArithmeticScaleModeCrossCase(caseIndex, transitionLimit) {
		group := tier1ArithmeticScaleModeCrossGroup(caseIndex)
		exponent := int64(group) - transitionLimit
		sign := uint32(tier1ArithmeticScaleModeCrossSign(group) << 31)
		switch {
		case exponent > 0:
			return sign | 0x00000001
		case exponent < 0:
			return sign | 0x77f8967f
		default:
			return sign | 0x32800001
		}
	}
	raw := uint32(tier1ArithmeticRandomWord(seed, tier1ArithmeticRandomScaleOperandIndex(caseIndex, transitionLimit), 0))
	if caseIndex%tier1ArithmeticScaleRandomStrata != 0 {
		raw = raw&^0x20000000 | 1
	}
	return raw
}

func tier1ArithmeticRandomScaleOperand64(seed, caseIndex uint64, transitionLimit int64) uint64 {
	if tier1ArithmeticScaleModeCrossCase(caseIndex, transitionLimit) {
		group := tier1ArithmeticScaleModeCrossGroup(caseIndex)
		exponent := int64(group) - transitionLimit
		sign := tier1ArithmeticScaleModeCrossSign(group) << 63
		switch {
		case exponent > 0:
			return sign | 0x0000000000000001
		case exponent < 0:
			return sign | 0x77fb86f26fc0ffff
		default:
			return sign | 0x31c0000000000001
		}
	}
	raw := tier1ArithmeticRandomWord(seed, tier1ArithmeticRandomScaleOperandIndex(caseIndex, transitionLimit), 0)
	if caseIndex%tier1ArithmeticScaleRandomStrata != 0 {
		raw = raw&^0x2000000000000000 | 1
	}
	return raw
}

func tier1ArithmeticRandomScaleOperand128(seed, caseIndex uint64, transitionLimit int64) tier1Arithmetic128Words {
	if tier1ArithmeticScaleModeCrossCase(caseIndex, transitionLimit) {
		group := tier1ArithmeticScaleModeCrossGroup(caseIndex)
		exponent := int64(group) - transitionLimit
		sign := tier1ArithmeticScaleModeCrossSign(group) << 63
		switch {
		case exponent > 0:
			return tier1Arithmetic128Words{lo: 0x0000000000000001, hi: sign}
		case exponent < 0:
			return tier1Arithmetic128Words{lo: 0x378d8e63ffffffff, hi: sign | 0x5fffed09bead87c0}
		default:
			return tier1Arithmetic128Words{lo: 0x0000000000000001, hi: sign | 0x3040000000000000}
		}
	}
	operandIndex := tier1ArithmeticRandomScaleOperandIndex(caseIndex, transitionLimit)
	raw := tier1Arithmetic128Words{
		lo: tier1ArithmeticRandomWord(seed, operandIndex, 0),
		hi: tier1ArithmeticRandomWord(seed, operandIndex, 1),
	}
	if caseIndex%tier1ArithmeticScaleRandomStrata != 0 {
		raw.lo |= 1
		raw.hi &^= 0x2001000000000000
	}
	return raw
}

func tier1ArithmeticScaleOperand32IsCanonicalNonzeroFinite(raw uint32) bool {
	if raw&0x60000000 == 0x60000000 {
		if raw&0x78000000 == 0x78000000 {
			return false
		}
		coefficient := raw&0x001fffff | 0x00800000
		return coefficient < 10000000
	}
	return raw&0x007fffff != 0
}

func tier1ArithmeticScaleOperand64IsCanonicalNonzeroFinite(raw uint64) bool {
	if raw&0x6000000000000000 == 0x6000000000000000 {
		if raw&0x7800000000000000 == 0x7800000000000000 {
			return false
		}
		coefficient := raw&0x0007ffffffffffff | 0x0020000000000000
		return coefficient < 10000000000000000
	}
	return raw&0x001fffffffffffff != 0
}

func tier1ArithmeticScaleOperand128IsCanonicalNonzeroFinite(raw tier1Arithmetic128Words) bool {
	if raw.hi&0x7800000000000000 >= 0x6000000000000000 {
		return false
	}
	coefficientHi := raw.hi & 0x0001ffffffffffff
	if coefficientHi == 0 && raw.lo == 0 {
		return false
	}
	return coefficientHi < 0x0001ed09bead87c0 ||
		coefficientHi == 0x0001ed09bead87c0 && raw.lo <= 0x378d8e63ffffffff
}

func tier1ArithmeticScaleTupleHashMix(digest, word uint64) uint64 {
	return (digest ^ word) * tier1ArithmeticScaleTupleHashPrime
}

func tier1ArithmeticScaleOperand(bits int, seed, caseIndex uint64, transitionLimit int64) (lo, hi uint64, finite bool) {
	switch bits {
	case 32:
		raw := tier1ArithmeticRandomScaleOperand32(seed, caseIndex, transitionLimit)
		return uint64(raw), 0, tier1ArithmeticScaleOperand32IsCanonicalNonzeroFinite(raw)
	case 64:
		raw := tier1ArithmeticRandomScaleOperand64(seed, caseIndex, transitionLimit)
		return raw, 0, tier1ArithmeticScaleOperand64IsCanonicalNonzeroFinite(raw)
	case 128:
		raw := tier1ArithmeticRandomScaleOperand128(seed, caseIndex, transitionLimit)
		return raw.lo, raw.hi, tier1ArithmeticScaleOperand128IsCanonicalNonzeroFinite(raw)
	default:
		panic(fmt.Sprintf("unsupported Tier 1 ScaleB corpus width %d", bits))
	}
}

func tier1ArithmeticAssertScaleCorpusContract(t *testing.T, bits int, seed, lane uint64, transitionLimit int64, cases, wantGroups, wantHash uint64) {
	t.Helper()
	if got := tier1ArithmeticScaleModeCrossGroups(transitionLimit); got != wantGroups {
		t.Fatalf("decimal%d ScaleB mode-cross groups=%d want=%d", bits, got, wantGroups)
	}
	seen := make([]bool, wantGroups*tier1ArithmeticScaleModeCross)
	modeSets := make([]uint8, wantGroups)
	operandLo := make([]uint64, wantGroups)
	operandHi := make([]uint64, wantGroups)
	operandSeen := make([]bool, wantGroups)
	digest := tier1ArithmeticScaleTupleHashMix(tier1ArithmeticScaleTupleHashOffset, uint64(bits))
	var modeCrossCases uint64
	for i := uint64(0); i < cases; i++ {
		lo, hi, finite := tier1ArithmeticScaleOperand(bits, seed, i, transitionLimit)
		exponent := tier1ArithmeticRandomScaleExponent(seed, i, lane, transitionLimit)
		mode := tier1ArithmeticModes[i%uint64(len(tier1ArithmeticModes))]
		digest = tier1ArithmeticScaleTupleHashMix(digest, i)
		digest = tier1ArithmeticScaleTupleHashMix(digest, lo)
		digest = tier1ArithmeticScaleTupleHashMix(digest, hi)
		digest = tier1ArithmeticScaleTupleHashMix(digest, uint64(exponent))
		digest = tier1ArithmeticScaleTupleHashMix(digest, uint64(mode.native))

		switch i % tier1ArithmeticScaleRandomStrata {
		case 0:
			want := int64(tier1ArithmeticRandomWord(seed, i, lane))
			if exponent != want {
				t.Fatalf("decimal%d ScaleB full-domain exponent=%d want=%d at case %d", bits, exponent, want, i)
			}
		case 1:
			if !tier1ArithmeticScaleModeCrossCase(i, transitionLimit) {
				if exponent < -transitionLimit || exponent > transitionLimit || !finite {
					t.Fatalf("decimal%d ScaleB surplus in-range case %d exponent=%d finite=%t", bits, i, exponent, finite)
				}
				continue
			}
			modeCrossCases++
			group := tier1ArithmeticScaleModeCrossGroup(i)
			wantExponent := int64(group) - transitionLimit
			if exponent != wantExponent {
				t.Fatalf("decimal%d ScaleB mode-cross exponent=%d want=%d at case %d", bits, exponent, wantExponent, i)
			}
			if got := tier1ArithmeticRandomScaleOperandIndex(i, transitionLimit); got != group {
				t.Fatalf("decimal%d ScaleB mode-cross operand index=%d want=%d at case %d", bits, got, group, i)
			}
			if !finite {
				t.Fatalf("decimal%d ScaleB mode-cross operand is not canonical nonzero finite at case %d: hi=%016x lo=%016x", bits, i, hi, lo)
			}
			if operandSeen[group] {
				if lo != operandLo[group] || hi != operandHi[group] {
					t.Fatalf("decimal%d ScaleB mode-cross operand drift in group %d: got=%016x:%016x want=%016x:%016x", bits, group, hi, lo, operandHi[group], operandLo[group])
				}
			} else {
				operandLo[group], operandHi[group], operandSeen[group] = lo, hi, true
			}
			modeIndex := i % tier1ArithmeticScaleModeCross
			modeSets[group] |= 1 << modeIndex
			seen[group*tier1ArithmeticScaleModeCross+modeIndex] = true
		case 2:
			if exponent < -transitionLimit*2 || exponent > transitionLimit*2 || !finite {
				t.Fatalf("decimal%d ScaleB transition-window case %d exponent=%d finite=%t", bits, i, exponent, finite)
			}
		case 3:
			if exponent < -transitionLimit || exponent > transitionLimit || !finite {
				t.Fatalf("decimal%d ScaleB in-range case %d exponent=%d finite=%t", bits, i, exponent, finite)
			}
		}
	}
	digest = tier1ArithmeticScaleTupleHashMix(digest, cases)
	if digest != wantHash {
		t.Fatalf("decimal%d ScaleB tuple stream hash=%d want=%d", bits, digest, wantHash)
	}
	if want := wantGroups * tier1ArithmeticScaleModeCross; modeCrossCases != want {
		t.Fatalf("decimal%d ScaleB complete mode-cross cases=%d want=%d", bits, modeCrossCases, want)
	}
	missing := 0
	for _, covered := range seen {
		if !covered {
			missing++
		}
	}
	if missing != 0 {
		t.Fatalf("decimal%d ScaleB random corpus misses %d/%d finite-transition exponent×rounding-mode cells", bits, missing, len(seen))
	}
	for group, modes := range modeSets {
		if modes != 0x1f {
			t.Fatalf("decimal%d ScaleB group %d native-mode index set=%05b want=11111", bits, group, modes)
		}
	}
}

// tier1ArithmeticAssertCorpusStreamContracts recomputes the pair, fma-triple,
// and non-ScaleB random operand streams through the exact visit functions and
// seed/lane/mode derivations the differentials consume, and compares the
// digests against the generator-anchored constants. Both language runners
// carry this contract, so an iteration-order, probe, seed, lane, truncation,
// or mode-rotation drift in either template fails here instead of silently
// splitting the cross-language corpus.
func tier1ArithmeticAssertCorpusStreamContracts(t *testing.T) {
	t.Helper()
	mix := tier1ArithmeticScaleTupleHashMix

	pairDigest32 := mix(tier1ArithmeticScaleTupleHashOffset, 32)
	var pairVisits32 uint64
	tier1ArithmeticVisitPairs32(func(x, y uint32) {
		pairDigest32 = mix(pairDigest32, uint64(x))
		pairDigest32 = mix(pairDigest32, 0)
		pairDigest32 = mix(pairDigest32, uint64(y))
		pairDigest32 = mix(pairDigest32, 0)
		pairVisits32++
	})
	if got := mix(pairDigest32, pairVisits32); got != tier1ArithmeticPairStreamHash32 {
		t.Fatalf("decimal32 pair stream hash=%d want=%d", got, tier1ArithmeticPairStreamHash32)
	}
	pairDigest64 := mix(tier1ArithmeticScaleTupleHashOffset, 64)
	var pairVisits64 uint64
	tier1ArithmeticVisitPairs64(func(x, y uint64) {
		pairDigest64 = mix(pairDigest64, x)
		pairDigest64 = mix(pairDigest64, 0)
		pairDigest64 = mix(pairDigest64, y)
		pairDigest64 = mix(pairDigest64, 0)
		pairVisits64++
	})
	if got := mix(pairDigest64, pairVisits64); got != tier1ArithmeticPairStreamHash64 {
		t.Fatalf("decimal64 pair stream hash=%d want=%d", got, tier1ArithmeticPairStreamHash64)
	}
	pairDigest128 := mix(tier1ArithmeticScaleTupleHashOffset, 128)
	var pairVisits128 uint64
	tier1ArithmeticVisitPairs128(func(x, y tier1Arithmetic128Words) {
		pairDigest128 = mix(pairDigest128, x.lo)
		pairDigest128 = mix(pairDigest128, x.hi)
		pairDigest128 = mix(pairDigest128, y.lo)
		pairDigest128 = mix(pairDigest128, y.hi)
		pairVisits128++
	})
	if got := mix(pairDigest128, pairVisits128); got != tier1ArithmeticPairStreamHash128 {
		t.Fatalf("decimal128 pair stream hash=%d want=%d", got, tier1ArithmeticPairStreamHash128)
	}

	tripleDigest32 := mix(tier1ArithmeticScaleTupleHashOffset, 32)
	var tripleVisits32 uint64
	tier1ArithmeticVisitTriples32(func(x, y, z uint32) {
		tripleDigest32 = mix(tripleDigest32, uint64(x))
		tripleDigest32 = mix(tripleDigest32, 0)
		tripleDigest32 = mix(tripleDigest32, uint64(y))
		tripleDigest32 = mix(tripleDigest32, 0)
		tripleDigest32 = mix(tripleDigest32, uint64(z))
		tripleDigest32 = mix(tripleDigest32, 0)
		tripleVisits32++
	})
	if got := mix(tripleDigest32, tripleVisits32); got != tier1ArithmeticFmaTripleStreamHash32 {
		t.Fatalf("decimal32 fma triple stream hash=%d want=%d", got, tier1ArithmeticFmaTripleStreamHash32)
	}
	tripleDigest64 := mix(tier1ArithmeticScaleTupleHashOffset, 64)
	var tripleVisits64 uint64
	tier1ArithmeticVisitTriples64(func(x, y, z uint64) {
		tripleDigest64 = mix(tripleDigest64, x)
		tripleDigest64 = mix(tripleDigest64, 0)
		tripleDigest64 = mix(tripleDigest64, y)
		tripleDigest64 = mix(tripleDigest64, 0)
		tripleDigest64 = mix(tripleDigest64, z)
		tripleDigest64 = mix(tripleDigest64, 0)
		tripleVisits64++
	})
	if got := mix(tripleDigest64, tripleVisits64); got != tier1ArithmeticFmaTripleStreamHash64 {
		t.Fatalf("decimal64 fma triple stream hash=%d want=%d", got, tier1ArithmeticFmaTripleStreamHash64)
	}
	tripleDigest128 := mix(tier1ArithmeticScaleTupleHashOffset, 128)
	var tripleVisits128 uint64
	tier1ArithmeticVisitTriples128(func(x, y, z tier1Arithmetic128Words) {
		tripleDigest128 = mix(tripleDigest128, x.lo)
		tripleDigest128 = mix(tripleDigest128, x.hi)
		tripleDigest128 = mix(tripleDigest128, y.lo)
		tripleDigest128 = mix(tripleDigest128, y.hi)
		tripleDigest128 = mix(tripleDigest128, z.lo)
		tripleDigest128 = mix(tripleDigest128, z.hi)
		tripleVisits128++
	})
	if got := mix(tripleDigest128, tripleVisits128); got != tier1ArithmeticFmaTripleStreamHash128 {
		t.Fatalf("decimal128 fma triple stream hash=%d want=%d", got, tier1ArithmeticFmaTripleStreamHash128)
	}

	randomStream := func(bits int, roundedBase, unroundedBase, fmaSeed, sqrtSeed, cases uint64) uint64 {
		digest := mix(tier1ArithmeticScaleTupleHashOffset, uint64(bits))
		mixOperand := func(seed, caseIndex, lane uint64) {
			switch bits {
			case 32:
				digest = mix(digest, uint64(tier1ArithmeticRandomOperand32(seed, caseIndex, lane)))
				digest = mix(digest, 0)
			case 64:
				digest = mix(digest, tier1ArithmeticRandomOperand64(seed, caseIndex, lane))
				digest = mix(digest, 0)
			default:
				words := tier1ArithmeticRandomOperand128(seed, caseIndex, lane)
				digest = mix(digest, words.lo)
				digest = mix(digest, words.hi)
			}
		}
		var total uint64
		for opIndex := uint64(0); opIndex < tier1ArithmeticRoundedOps; opIndex++ {
			seed := roundedBase ^ opIndex
			for i := uint64(0); i < cases; i++ {
				mixOperand(seed, i, 0)
				mixOperand(seed, i, 1)
				digest = mix(digest, uint64(tier1ArithmeticModes[i%uint64(len(tier1ArithmeticModes))].native))
				total++
			}
		}
		for opIndex := uint64(0); opIndex < tier1ArithmeticUnroundedOps; opIndex++ {
			seed := unroundedBase ^ opIndex
			for i := uint64(0); i < cases; i++ {
				mixOperand(seed, i, 0)
				mixOperand(seed, i, 1)
				total++
			}
		}
		for i := uint64(0); i < cases; i++ {
			mixOperand(fmaSeed, i, 0)
			mixOperand(fmaSeed, i, 1)
			mixOperand(fmaSeed, i, 2)
			digest = mix(digest, uint64(tier1ArithmeticModes[i%uint64(len(tier1ArithmeticModes))].native))
			total++
		}
		for i := uint64(0); i < cases; i++ {
			mixOperand(sqrtSeed, i, 0)
			digest = mix(digest, uint64(tier1ArithmeticModes[i%uint64(len(tier1ArithmeticModes))].native))
			total++
		}
		return mix(digest, total)
	}
	if got := randomStream(32, tier1ArithmeticRandomRoundedSeedBase32, tier1ArithmeticRandomUnroundedSeedBase32, tier1ArithmeticRandomFmaSeed32, tier1ArithmeticRandomSqrtSeed32, tier1ArithmeticRandomCasesPerOp32); got != tier1ArithmeticRandomStreamHash32 {
		t.Fatalf("decimal32 random stream hash=%d want=%d", got, tier1ArithmeticRandomStreamHash32)
	}
	if got := randomStream(64, tier1ArithmeticRandomRoundedSeedBase64, tier1ArithmeticRandomUnroundedSeedBase64, tier1ArithmeticRandomFmaSeed64, tier1ArithmeticRandomSqrtSeed64, tier1ArithmeticRandomCasesPerOp64); got != tier1ArithmeticRandomStreamHash64 {
		t.Fatalf("decimal64 random stream hash=%d want=%d", got, tier1ArithmeticRandomStreamHash64)
	}
	if got := randomStream(128, tier1ArithmeticRandomRoundedSeedBase128, tier1ArithmeticRandomUnroundedSeedBase128, tier1ArithmeticRandomFmaSeed128, tier1ArithmeticRandomSqrtSeed128, tier1ArithmeticRandomCasesPerOp128); got != tier1ArithmeticRandomStreamHash128 {
		t.Fatalf("decimal128 random stream hash=%d want=%d", got, tier1ArithmeticRandomStreamHash128)
	}
}

func TestTier1ArithmeticCorpusContract(t *testing.T) {
	if tier1ArithmeticScaleModeCross != uint64(len(tier1ArithmeticModes)) {
		t.Fatalf("ScaleB mode cross=%d rounding modes=%d", tier1ArithmeticScaleModeCross, len(tier1ArithmeticModes))
	}
	tests := []struct {
		seed, caseIndex, lane uint64
		want                  uint64
	}{
		{0xdec7543200000000, 0, 0, @@TIER1_RANDOM_SAMPLE0@@},
		{0xdec7546400000004, (uint64(1) << 20) - 1, 1, @@TIER1_RANDOM_SAMPLE1@@},
		{0xdec7541253414c45, (uint64(1) << 19) - 1, 2, @@TIER1_RANDOM_SAMPLE2@@},
	}
	for _, tc := range tests {
		if got := tier1ArithmeticRandomWord(tc.seed, tc.caseIndex, tc.lane); got != tc.want {
			t.Fatalf("Tier 1 corpus PRNG drift: seed=%016x case=%d lane=%d got=%016x want=%016x", tc.seed, tc.caseIndex, tc.lane, got, tc.want)
		}
	}
	tier1ArithmeticAssertScaleCorpusContract(t, 32, 0xdec7543253414c45, 1, int64(tier1ArithmeticScaleFiniteTransitionLimit32), tier1ArithmeticRandomCasesPerOp32, tier1ArithmeticScaleModeCrossGroups32, tier1ArithmeticScaleTupleHash32)
	tier1ArithmeticAssertScaleCorpusContract(t, 64, 0xdec7546453414c45, 1, int64(tier1ArithmeticScaleFiniteTransitionLimit64), tier1ArithmeticRandomCasesPerOp64, tier1ArithmeticScaleModeCrossGroups64, tier1ArithmeticScaleTupleHash64)
	tier1ArithmeticAssertScaleCorpusContract(t, 128, 0xdec7541253414c45, 2, int64(tier1ArithmeticScaleFiniteTransitionLimit128), tier1ArithmeticRandomCasesPerOp128, tier1ArithmeticScaleModeCrossGroups128, tier1ArithmeticScaleTupleHash128)
	tier1ArithmeticAssertCorpusStreamContracts(t)
}

func TestTier1ArithmeticDeterministicRandomNativeDifferential(t *testing.T) {
	requireNative(t)
	if strconv.IntSize != 64 {
		t.Fatalf("Tier 1 arithmetic native oracle requires the guaranteed LP64 platform contract; int size=%d", strconv.IntSize)
	}
	shard := tier1ArithmeticLoadShard(t)

	t.Run("decimal32", func(t *testing.T) {
		var comparison uint64
		for operationIndex, operation := range tier1ArithmeticRoundedOperations {
			seed := tier1ArithmeticRandomRoundedSeedBase32 ^ uint64(operationIndex)
			for i := uint64(0); i < tier1ArithmeticRandomCasesPerOp32; i++ {
				if shard.owns(comparison) {
					x := tier1ArithmeticRandomOperand32(seed, i, 0)
					y := tier1ArithmeticRandomOperand32(seed, i, 1)
					tier1ArithmeticCheckRounded32(t, operation, x, y, tier1ArithmeticModes[i%uint64(len(tier1ArithmeticModes))])
				}
				comparison++
			}
		}
		for operationIndex, operation := range tier1ArithmeticUnroundedOperations {
			seed := tier1ArithmeticRandomUnroundedSeedBase32 ^ uint64(operationIndex)
			for i := uint64(0); i < tier1ArithmeticRandomCasesPerOp32; i++ {
				if shard.owns(comparison) {
					x := tier1ArithmeticRandomOperand32(seed, i, 0)
					y := tier1ArithmeticRandomOperand32(seed, i, 1)
					tier1ArithmeticCheckUnrounded32(t, operation, x, y)
				}
				comparison++
			}
		}
		seed := uint64(0xdec7543253414c45)
		for i := uint64(0); i < tier1ArithmeticRandomCasesPerOp32; i++ {
			if shard.owns(comparison) {
				x := tier1ArithmeticRandomScaleOperand32(seed, i, int64(tier1ArithmeticScaleFiniteTransitionLimit32))
				exponent := tier1ArithmeticRandomScaleExponent(seed, i, 1, int64(tier1ArithmeticScaleFiniteTransitionLimit32))
				tier1ArithmeticCheckScale32(t, x, exponent, tier1ArithmeticModes[i%uint64(len(tier1ArithmeticModes))])
			}
			comparison++
		}
		seed = tier1ArithmeticRandomFmaSeed32
		for i := uint64(0); i < tier1ArithmeticRandomCasesPerOp32; i++ {
			if shard.owns(comparison) {
				x := tier1ArithmeticRandomOperand32(seed, i, 0)
				y := tier1ArithmeticRandomOperand32(seed, i, 1)
				z := tier1ArithmeticRandomOperand32(seed, i, 2)
				tier1ArithmeticCheckFma32(t, x, y, z, tier1ArithmeticModes[i%uint64(len(tier1ArithmeticModes))])
			}
			comparison++
		}
		seed = tier1ArithmeticRandomSqrtSeed32
		for i := uint64(0); i < tier1ArithmeticRandomCasesPerOp32; i++ {
			if shard.owns(comparison) {
				tier1ArithmeticCheckSqrt32(t, tier1ArithmeticRandomOperand32(seed, i, 0), tier1ArithmeticModes[i%uint64(len(tier1ArithmeticModes))])
			}
			comparison++
		}
		if comparison != tier1ArithmeticRandomComparisons32 {
			t.Fatalf("decimal32 random comparisons=%d want=%d", comparison, tier1ArithmeticRandomComparisons32)
		}
		t.Logf("decimal32 deterministic random exact comparisons: %d/%d", shard.ownedCount(comparison), comparison)
	})

	t.Run("decimal64", func(t *testing.T) {
		var comparison uint64
		for operationIndex, operation := range tier1ArithmeticRoundedOperations {
			seed := tier1ArithmeticRandomRoundedSeedBase64 ^ uint64(operationIndex)
			for i := uint64(0); i < tier1ArithmeticRandomCasesPerOp64; i++ {
				if shard.owns(comparison) {
					x := tier1ArithmeticRandomOperand64(seed, i, 0)
					y := tier1ArithmeticRandomOperand64(seed, i, 1)
					tier1ArithmeticCheckRounded64(t, operation, x, y, tier1ArithmeticModes[i%uint64(len(tier1ArithmeticModes))])
				}
				comparison++
			}
		}
		for operationIndex, operation := range tier1ArithmeticUnroundedOperations {
			seed := tier1ArithmeticRandomUnroundedSeedBase64 ^ uint64(operationIndex)
			for i := uint64(0); i < tier1ArithmeticRandomCasesPerOp64; i++ {
				if shard.owns(comparison) {
					x := tier1ArithmeticRandomOperand64(seed, i, 0)
					y := tier1ArithmeticRandomOperand64(seed, i, 1)
					tier1ArithmeticCheckUnrounded64(t, operation, x, y)
				}
				comparison++
			}
		}
		seed := uint64(0xdec7546453414c45)
		for i := uint64(0); i < tier1ArithmeticRandomCasesPerOp64; i++ {
			if shard.owns(comparison) {
				x := tier1ArithmeticRandomScaleOperand64(seed, i, int64(tier1ArithmeticScaleFiniteTransitionLimit64))
				exponent := tier1ArithmeticRandomScaleExponent(seed, i, 1, int64(tier1ArithmeticScaleFiniteTransitionLimit64))
				tier1ArithmeticCheckScale64(t, x, exponent, tier1ArithmeticModes[i%uint64(len(tier1ArithmeticModes))])
			}
			comparison++
		}
		seed = tier1ArithmeticRandomFmaSeed64
		for i := uint64(0); i < tier1ArithmeticRandomCasesPerOp64; i++ {
			if shard.owns(comparison) {
				x := tier1ArithmeticRandomOperand64(seed, i, 0)
				y := tier1ArithmeticRandomOperand64(seed, i, 1)
				z := tier1ArithmeticRandomOperand64(seed, i, 2)
				tier1ArithmeticCheckFma64(t, x, y, z, tier1ArithmeticModes[i%uint64(len(tier1ArithmeticModes))])
			}
			comparison++
		}
		seed = tier1ArithmeticRandomSqrtSeed64
		for i := uint64(0); i < tier1ArithmeticRandomCasesPerOp64; i++ {
			if shard.owns(comparison) {
				tier1ArithmeticCheckSqrt64(t, tier1ArithmeticRandomOperand64(seed, i, 0), tier1ArithmeticModes[i%uint64(len(tier1ArithmeticModes))])
			}
			comparison++
		}
		if comparison != tier1ArithmeticRandomComparisons64 {
			t.Fatalf("decimal64 random comparisons=%d want=%d", comparison, tier1ArithmeticRandomComparisons64)
		}
		t.Logf("decimal64 deterministic random exact comparisons: %d/%d", shard.ownedCount(comparison), comparison)
	})

	t.Run("decimal128", func(t *testing.T) {
		var comparison uint64
		for operationIndex, operation := range tier1ArithmeticRoundedOperations {
			seed := tier1ArithmeticRandomRoundedSeedBase128 ^ uint64(operationIndex)
			for i := uint64(0); i < tier1ArithmeticRandomCasesPerOp128; i++ {
				if shard.owns(comparison) {
					x := tier1ArithmeticDecimal128(tier1ArithmeticRandomOperand128(seed, i, 0))
					y := tier1ArithmeticDecimal128(tier1ArithmeticRandomOperand128(seed, i, 1))
					tier1ArithmeticCheckRounded128(t, operation, x, y, tier1ArithmeticModes[i%uint64(len(tier1ArithmeticModes))])
				}
				comparison++
			}
		}
		for operationIndex, operation := range tier1ArithmeticUnroundedOperations {
			seed := tier1ArithmeticRandomUnroundedSeedBase128 ^ uint64(operationIndex)
			for i := uint64(0); i < tier1ArithmeticRandomCasesPerOp128; i++ {
				if shard.owns(comparison) {
					x := tier1ArithmeticDecimal128(tier1ArithmeticRandomOperand128(seed, i, 0))
					y := tier1ArithmeticDecimal128(tier1ArithmeticRandomOperand128(seed, i, 1))
					tier1ArithmeticCheckUnrounded128(t, operation, x, y)
				}
				comparison++
			}
		}
		seed := uint64(0xdec7541253414c45)
		for i := uint64(0); i < tier1ArithmeticRandomCasesPerOp128; i++ {
			if shard.owns(comparison) {
				x := tier1ArithmeticDecimal128(tier1ArithmeticRandomScaleOperand128(seed, i, int64(tier1ArithmeticScaleFiniteTransitionLimit128)))
				exponent := tier1ArithmeticRandomScaleExponent(seed, i, 2, int64(tier1ArithmeticScaleFiniteTransitionLimit128))
				tier1ArithmeticCheckScale128(t, x, exponent, tier1ArithmeticModes[i%uint64(len(tier1ArithmeticModes))])
			}
			comparison++
		}
		seed = tier1ArithmeticRandomFmaSeed128
		for i := uint64(0); i < tier1ArithmeticRandomCasesPerOp128; i++ {
			if shard.owns(comparison) {
				x := tier1ArithmeticDecimal128(tier1ArithmeticRandomOperand128(seed, i, 0))
				y := tier1ArithmeticDecimal128(tier1ArithmeticRandomOperand128(seed, i, 1))
				z := tier1ArithmeticDecimal128(tier1ArithmeticRandomOperand128(seed, i, 2))
				tier1ArithmeticCheckFma128(t, x, y, z, tier1ArithmeticModes[i%uint64(len(tier1ArithmeticModes))])
			}
			comparison++
		}
		seed = tier1ArithmeticRandomSqrtSeed128
		for i := uint64(0); i < tier1ArithmeticRandomCasesPerOp128; i++ {
			if shard.owns(comparison) {
				x := tier1ArithmeticDecimal128(tier1ArithmeticRandomOperand128(seed, i, 0))
				tier1ArithmeticCheckSqrt128(t, x, tier1ArithmeticModes[i%uint64(len(tier1ArithmeticModes))])
			}
			comparison++
		}
		if comparison != tier1ArithmeticRandomComparisons128 {
			t.Fatalf("decimal128 random comparisons=%d want=%d", comparison, tier1ArithmeticRandomComparisons128)
		}
		t.Logf("decimal128 deterministic random exact comparisons: %d/%d", shard.ownedCount(comparison), comparison)
	})
}

var tier1ArithmeticProbes32 = [...]uint32{
@@TIER1_PROBES32_VALUES@@
}

var tier1ArithmeticProbes64 = [...]uint64{
@@TIER1_PROBES64_VALUES@@
}

var tier1ArithmeticProbes128 = [...]tier1Arithmetic128Words{
@@TIER1_PROBES128_VALUES@@
}

var tier1ArithmeticSemanticRounded32Cases = []tier1ArithmeticSemanticRounded32{
@@TIER1_SEMANTIC_ROUNDED32_VALUES@@
}

var tier1ArithmeticSemanticRounded64Cases = []tier1ArithmeticSemanticRounded64{
@@TIER1_SEMANTIC_ROUNDED64_VALUES@@
}

var tier1ArithmeticSemanticRounded128Cases = []tier1ArithmeticSemanticRounded128{
@@TIER1_SEMANTIC_ROUNDED128_VALUES@@
}

var tier1ArithmeticSemanticScale32Cases = []tier1ArithmeticSemanticScale32{
@@TIER1_SEMANTIC_SCALE32_VALUES@@
}

var tier1ArithmeticSemanticScale64Cases = []tier1ArithmeticSemanticScale64{
@@TIER1_SEMANTIC_SCALE64_VALUES@@
}

var tier1ArithmeticSemanticScale128Cases = []tier1ArithmeticSemanticScale128{
@@TIER1_SEMANTIC_SCALE128_VALUES@@
}

var tier1ArithmeticSemanticFma32Cases = []tier1ArithmeticTriple32{
@@TIER1_SEMANTIC_FMA32_VALUES@@
}

var tier1ArithmeticSemanticFma64Cases = []tier1ArithmeticTriple64{
@@TIER1_SEMANTIC_FMA64_VALUES@@
}

var tier1ArithmeticSemanticFma128Cases = []tier1ArithmeticTriple128{
@@TIER1_SEMANTIC_FMA128_VALUES@@
}

var tier1ArithmeticSemanticSqrt32Cases = []uint32{
@@TIER1_SEMANTIC_SQRT32_VALUES@@
}

var tier1ArithmeticSemanticSqrt64Cases = []uint64{
@@TIER1_SEMANTIC_SQRT64_VALUES@@
}

var tier1ArithmeticSemanticSqrt128Cases = []tier1Arithmetic128Words{
@@TIER1_SEMANTIC_SQRT128_VALUES@@
}

var tier1ArithmeticSemanticRemainder32Pairs = []tier1ArithmeticPair32{
@@TIER1_SEMANTIC_REMAINDER32_VALUES@@
}

var tier1ArithmeticSemanticRemainder64Pairs = []tier1ArithmeticPair64{
@@TIER1_SEMANTIC_REMAINDER64_VALUES@@
}

var tier1ArithmeticSemanticRemainder128Pairs = []tier1ArithmeticPair128{
@@TIER1_SEMANTIC_REMAINDER128_VALUES@@
}

var tier1ArithmeticBoundary32 = []uint32{
@@TIER1_BOUNDARY32_VALUES@@
}

var tier1ArithmeticBoundary64 = []uint64{
@@TIER1_BOUNDARY64_VALUES@@
}

var tier1ArithmeticBoundary128 = []tier1Arithmetic128Words{
@@TIER1_BOUNDARY128_VALUES@@
}
