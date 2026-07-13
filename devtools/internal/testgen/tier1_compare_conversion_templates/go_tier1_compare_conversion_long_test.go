//go:build cgo && bid754_native && bid754_tier1_long

package bid754

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/sky1core/bid754/bid754-go/internal/bidgo"
)

const (
	tier1CompareConversionBoundary32Count  = uint64(@@TIER1_BOUNDARY32_COUNT@@)
	tier1CompareConversionBoundary64Count  = uint64(@@TIER1_BOUNDARY64_COUNT@@)
	tier1CompareConversionBoundary128Count = uint64(@@TIER1_BOUNDARY128_COUNT@@)
	tier1ConversionSemantic32Count          = uint64(@@TIER1_CONVERSION_SEMANTIC32_COUNT@@)
	tier1ConversionSemantic64Count          = uint64(@@TIER1_CONVERSION_SEMANTIC64_COUNT@@)
	tier1ConversionSemantic128Count         = uint64(@@TIER1_CONVERSION_SEMANTIC128_COUNT@@)
	tier1ConstructorInt32Count              = uint64(@@TIER1_CONSTRUCTOR_INT32_COUNT@@)
	tier1ConstructorUint32Count             = uint64(@@TIER1_CONSTRUCTOR_UINT32_COUNT@@)
	tier1ConstructorInt64Count              = uint64(@@TIER1_CONSTRUCTOR_INT64_COUNT@@)
	tier1ConstructorUint64Count             = uint64(@@TIER1_CONSTRUCTOR_UINT64_COUNT@@)
	tier1CompareConversionRoutingSentinelCount = uint64(@@TIER1_CC_SENTINEL_COUNT@@)

	tier1QuietPredicateCount        = uint64(6)
	tier1MinMaxOperationCount       = uint64(4)
	tier1CompareMinMaxOperationCount = tier1QuietPredicateCount + tier1MinMaxOperationCount
	tier1QuietSemanticRelationCount = uint64(16)
	tier1CompareProbeCount          = uint64(12)
	tier1CompareBoundaryPairs32 = tier1CompareConversionBoundary32Count*tier1CompareProbeCount*2 +
		tier1CompareProbeCount*tier1CompareProbeCount
	tier1CompareBoundaryPairs64 = tier1CompareConversionBoundary64Count*tier1CompareProbeCount*2 +
		tier1CompareProbeCount*tier1CompareProbeCount
	tier1CompareBoundaryPairs128 = tier1CompareConversionBoundary128Count*tier1CompareProbeCount*2 +
		tier1CompareProbeCount*tier1CompareProbeCount

	tier1CompareStructured32 = tier1CompareBoundaryPairs32 * tier1CompareMinMaxOperationCount
	tier1CompareStructured64 = tier1CompareBoundaryPairs64 * tier1CompareMinMaxOperationCount
	tier1CompareStructured128 = tier1CompareBoundaryPairs128 * tier1CompareMinMaxOperationCount
	tier1CompareRandomPairs32  = uint64(1) << 20
	tier1CompareRandomPairs64  = uint64(1) << 20
	tier1CompareRandomPairs128 = uint64(1) << 19
	tier1CompareRandomComparisons32 = tier1CompareRandomPairs32 * tier1CompareMinMaxOperationCount
	tier1CompareRandomComparisons64 = tier1CompareRandomPairs64 * tier1CompareMinMaxOperationCount
	tier1CompareRandomComparisons128 = tier1CompareRandomPairs128 * tier1CompareMinMaxOperationCount
	tier1CompareTotal32 = tier1CompareStructured32 + tier1CompareRandomComparisons32
	tier1CompareTotal64 = tier1CompareStructured64 + tier1CompareRandomComparisons64
	tier1CompareTotal128 = tier1CompareStructured128 + tier1CompareRandomComparisons128

	tier1ToIntegerOperationCount = uint64(80) // 8 targets * exact/non-exact * 5 modes.
	tier1ToIntegerStructured32 = (tier1CompareConversionBoundary32Count + tier1ConversionSemantic32Count) * tier1ToIntegerOperationCount
	tier1ToIntegerStructured64 = (tier1CompareConversionBoundary64Count + tier1ConversionSemantic64Count) * tier1ToIntegerOperationCount
	tier1ToIntegerStructured128 = (tier1CompareConversionBoundary128Count + tier1ConversionSemantic128Count) * tier1ToIntegerOperationCount
	tier1ToIntegerRandom32  = uint64(1) << 18
	tier1ToIntegerRandom64  = uint64(1) << 18
	tier1ToIntegerRandom128 = uint64(1) << 17
	tier1ToIntegerTotal32 = tier1ToIntegerStructured32 + tier1ToIntegerRandom32
	tier1ToIntegerTotal64 = tier1ToIntegerStructured64 + tier1ToIntegerRandom64
	tier1ToIntegerTotal128 = tier1ToIntegerStructured128 + tier1ToIntegerRandom128

	// One-way BID -> binary interchange conversions: 3 targets x 5 rounding
	// modes from every source width.
	tier1BinaryOperationsPerSource = uint64(15)
	tier1BinaryStructured32  = (tier1CompareConversionBoundary32Count + tier1ConversionSemantic32Count) * tier1BinaryOperationsPerSource
	tier1BinaryStructured64  = (tier1CompareConversionBoundary64Count + tier1ConversionSemantic64Count) * tier1BinaryOperationsPerSource
	tier1BinaryStructured128 = (tier1CompareConversionBoundary128Count + tier1ConversionSemantic128Count) * tier1BinaryOperationsPerSource
	tier1BinaryRandom32  = uint64(1) << 18
	tier1BinaryRandom64  = uint64(1) << 18
	tier1BinaryRandom128 = uint64(1) << 17

	tier1WidthOperationsFrom32  = uint64(2)
	tier1WidthOperationsFrom64  = uint64(6)
	tier1WidthOperationsFrom128 = uint64(10)
	tier1WidthStructured32 = (tier1CompareConversionBoundary32Count + tier1ConversionSemantic32Count) * tier1WidthOperationsFrom32
	tier1WidthStructured64 = (tier1CompareConversionBoundary64Count + tier1ConversionSemantic64Count) * tier1WidthOperationsFrom64
	tier1WidthStructured128 = (tier1CompareConversionBoundary128Count + tier1ConversionSemantic128Count) * tier1WidthOperationsFrom128
	tier1WidthRandom32  = uint64(1) << 18
	tier1WidthRandom64  = uint64(1) << 18
	tier1WidthRandom128 = uint64(1) << 17
	tier1WidthTotal32 = tier1WidthStructured32 + tier1WidthRandom32
	tier1WidthTotal64 = tier1WidthStructured64 + tier1WidthRandom64
	tier1WidthTotal128 = tier1WidthStructured128 + tier1WidthRandom128

	tier1ConstructorOperationsFromInt32  = uint64(7)
	tier1ConstructorOperationsFromUint32 = uint64(7)
	tier1ConstructorOperationsFromInt64  = uint64(11)
	tier1ConstructorOperationsFromUint64 = uint64(11)
	tier1ConstructorOperationCount       = uint64(36)
	tier1ConstructorStructured = tier1ConstructorInt32Count*tier1ConstructorOperationsFromInt32 +
		tier1ConstructorUint32Count*tier1ConstructorOperationsFromUint32 +
		tier1ConstructorInt64Count*tier1ConstructorOperationsFromInt64 +
		tier1ConstructorUint64Count*tier1ConstructorOperationsFromUint64
	tier1ConstructorRandom = uint64(1) << 20
	tier1ConstructorTotal  = tier1ConstructorStructured + tier1ConstructorRandom
	tier1ConstructorConvenienceChecks = tier1ConstructorInt32Count + 2*tier1ConstructorInt64Count

	tier1ConversionStructured = tier1ToIntegerStructured32 + tier1ToIntegerStructured64 + tier1ToIntegerStructured128 +
		tier1WidthStructured32 + tier1WidthStructured64 + tier1WidthStructured128 +
		tier1BinaryStructured32 + tier1BinaryStructured64 + tier1BinaryStructured128 + tier1ConstructorStructured
	tier1ConversionRandom = tier1ToIntegerRandom32 + tier1ToIntegerRandom64 + tier1ToIntegerRandom128 +
		tier1WidthRandom32 + tier1WidthRandom64 + tier1WidthRandom128 +
		tier1BinaryRandom32 + tier1BinaryRandom64 + tier1BinaryRandom128 + tier1ConstructorRandom
	tier1ConversionTotal = tier1ConversionStructured + tier1ConversionRandom

	tier1CompareRandomStreamHash32     = uint64(@@TIER1_COMPARE_RANDOM_STREAM_HASH32@@)
	tier1CompareRandomStreamHash64     = uint64(@@TIER1_COMPARE_RANDOM_STREAM_HASH64@@)
	tier1CompareRandomStreamHash128    = uint64(@@TIER1_COMPARE_RANDOM_STREAM_HASH128@@)
	tier1ToIntegerRandomStreamHash32   = uint64(@@TIER1_TO_INTEGER_RANDOM_STREAM_HASH32@@)
	tier1ToIntegerRandomStreamHash64   = uint64(@@TIER1_TO_INTEGER_RANDOM_STREAM_HASH64@@)
	tier1ToIntegerRandomStreamHash128  = uint64(@@TIER1_TO_INTEGER_RANDOM_STREAM_HASH128@@)
	tier1WidthRandomStreamHash32       = uint64(@@TIER1_WIDTH_RANDOM_STREAM_HASH32@@)
	tier1WidthRandomStreamHash64       = uint64(@@TIER1_WIDTH_RANDOM_STREAM_HASH64@@)
	tier1WidthRandomStreamHash128      = uint64(@@TIER1_WIDTH_RANDOM_STREAM_HASH128@@)
	tier1BinaryRandomStreamHash32      = uint64(@@TIER1_BINARY_RANDOM_STREAM_HASH32@@)
	tier1BinaryRandomStreamHash64      = uint64(@@TIER1_BINARY_RANDOM_STREAM_HASH64@@)
	tier1BinaryRandomStreamHash128     = uint64(@@TIER1_BINARY_RANDOM_STREAM_HASH128@@)
	tier1ConstructorRandomStreamHash   = uint64(@@TIER1_CONSTRUCTOR_RANDOM_STREAM_HASH@@)

	// The deterministic-random seeds are shared by every random differential
	// block and the generator-side stream hashes, so a seed edit that reaches
	// only one consumer fails the anchored stream-hash contract.
	tier1CompareRandomSeed32    = uint64(0xdec75432c04d5001)
	tier1CompareRandomSeed64    = uint64(0xdec75464c04d5001)
	tier1CompareRandomSeed128   = uint64(0xdec754c0c04d5001)
	tier1ToIntegerRandomSeed32  = uint64(0xdec75432c0a70001)
	tier1ToIntegerRandomSeed64  = uint64(0xdec75464c0a70001)
	tier1ToIntegerRandomSeed128 = uint64(0xdec754c0c0a70001)
	tier1WidthRandomSeed32      = uint64(0xdec75432c0de0001)
	tier1WidthRandomSeed64      = uint64(0xdec75464c0de0001)
	tier1WidthRandomSeed128     = uint64(0xdec754c0c0de0001)
	tier1BinaryRandomSeed32     = uint64(0xdec75432c0b10001)
	tier1BinaryRandomSeed64     = uint64(0xdec75464c0b10001)
	tier1BinaryRandomSeed128    = uint64(0xdec754c0c0b10001)
	tier1ConstructorRandomSeed  = uint64(0xdec754c0c0570001)
)

type tier1CompareConversionShard struct {
	count uint64
	index uint64
}

func tier1CompareConversionLoadShard(t *testing.T) tier1CompareConversionShard {
	t.Helper()
	countText := os.Getenv("BID754_TIER1_COMPARE_CONVERSION_SHARD_COUNT")
	indexText := os.Getenv("BID754_TIER1_COMPARE_CONVERSION_SHARD_INDEX")
	if countText == "" && indexText == "" {
		return tier1CompareConversionShard{count: 1}
	}
	if countText == "" || indexText == "" {
		t.Fatal("BID754_TIER1_COMPARE_CONVERSION_SHARD_COUNT and BID754_TIER1_COMPARE_CONVERSION_SHARD_INDEX must be set together")
	}
	count, err := strconv.ParseUint(countText, 10, 64)
	if err != nil || count == 0 {
		t.Fatalf("invalid BID754_TIER1_COMPARE_CONVERSION_SHARD_COUNT %q", countText)
	}
	index, err := strconv.ParseUint(indexText, 10, 64)
	if err != nil || index >= count {
		t.Fatalf("invalid BID754_TIER1_COMPARE_CONVERSION_SHARD_INDEX %q for shard count %d", indexText, count)
	}
	return tier1CompareConversionShard{count: count, index: index}
}

func (s tier1CompareConversionShard) owns(caseIndex uint64) bool {
	return caseIndex%s.count == s.index
}

func (s tier1CompareConversionShard) ownedCount(total uint64) uint64 {
	if total <= s.index {
		return 0
	}
	return 1 + (total-1-s.index)/s.count
}

var tier1QuietPredicates = [...]string{
	"quiet_equal",
	"quiet_not_equal",
	"quiet_less",
	"quiet_less_equal",
	"quiet_greater",
	"quiet_greater_equal",
}

var tier1MinMaxOperations = [...]string{"minnum", "maxnum", "minnum_mag", "maxnum_mag"}

// tier1Legs* compute every leg of one comparison from one decoded case. They
// are the single operand/dispatch fan-out shared by the differential checks
// and the routing sentinels, so a slot swap or a dispatch-row mislabel in
// this glue skews the differential's legs identically (an agreed-upon wrong
// answer the differential cannot see) while the pinned sentinel rows diverge
// and fail.

func tier1LegsQuiet32(t *testing.T, operation string, x, y uint32) (native int, nativeFlags uint32, port int, portFlags uint32, public int, publicFlags uint32, unknownFlags ExceptionFlags) {
	t.Helper()
	native, nativeFlags, port, portFlags = runGeneratedFFICase32IntBinary(operation, x, y)
	left, right := Decimal32BID(x), Decimal32BID(y)
	var got bool
	var gotFlags ExceptionFlags
	switch operation {
	case "quiet_equal":
		got, gotFlags = left.QuietEqual(right)
	case "quiet_not_equal":
		got, gotFlags = left.QuietNotEqual(right)
	case "quiet_less":
		got, gotFlags = left.QuietLess(right)
	case "quiet_less_equal":
		got, gotFlags = left.QuietLessEqual(right)
	case "quiet_greater":
		got, gotFlags = left.QuietGreater(right)
	case "quiet_greater_equal":
		got, gotFlags = left.QuietGreaterEqual(right)
	default:
		t.Fatalf("decimal32 unknown Tier 1 quiet predicate %q", operation)
	}
	publicFlags, unknownFlags = tier1ArithmeticPublicRawFlags(gotFlags)
	if got {
		public = 1
	}
	return
}

func tier1CheckQuiet32(t *testing.T, operation string, x, y uint32) {
	t.Helper()
	native, nativeFlags, port, portFlags, public, publicFlags, unknownFlags := tier1LegsQuiet32(t, operation, x, y)
	if unknownFlags != 0 || native != port || nativeFlags != portFlags || native != public || nativeFlags != publicFlags {
		t.Fatalf("decimal32 %s mismatch: x=%08x y=%08x C=%d/%08x port=%d/%08x public=%d/%08x unknown_public_flags=%s", operation, x, y, native, nativeFlags, port, portFlags, public, publicFlags, unknownFlags)
	}
}

func tier1LegsQuiet64(t *testing.T, operation string, x, y uint64) (native int, nativeFlags uint32, port int, portFlags uint32, public int, publicFlags uint32, unknownFlags ExceptionFlags) {
	t.Helper()
	native, nativeFlags, port, portFlags = runGeneratedFFICase64IntBinary(operation, x, y)
	left, right := Decimal64BID(x), Decimal64BID(y)
	var got bool
	var gotFlags ExceptionFlags
	switch operation {
	case "quiet_equal":
		got, gotFlags = left.QuietEqual(right)
	case "quiet_not_equal":
		got, gotFlags = left.QuietNotEqual(right)
	case "quiet_less":
		got, gotFlags = left.QuietLess(right)
	case "quiet_less_equal":
		got, gotFlags = left.QuietLessEqual(right)
	case "quiet_greater":
		got, gotFlags = left.QuietGreater(right)
	case "quiet_greater_equal":
		got, gotFlags = left.QuietGreaterEqual(right)
	default:
		t.Fatalf("decimal64 unknown Tier 1 quiet predicate %q", operation)
	}
	publicFlags, unknownFlags = tier1ArithmeticPublicRawFlags(gotFlags)
	if got {
		public = 1
	}
	return
}

func tier1CheckQuiet64(t *testing.T, operation string, x, y uint64) {
	t.Helper()
	native, nativeFlags, port, portFlags, public, publicFlags, unknownFlags := tier1LegsQuiet64(t, operation, x, y)
	if unknownFlags != 0 || native != port || nativeFlags != portFlags || native != public || nativeFlags != publicFlags {
		t.Fatalf("decimal64 %s mismatch: x=%016x y=%016x C=%d/%08x port=%d/%08x public=%d/%08x unknown_public_flags=%s", operation, x, y, native, nativeFlags, port, portFlags, public, publicFlags, unknownFlags)
	}
}

func tier1LegsQuiet128(t *testing.T, operation string, x, y Decimal128BID) (native int, nativeFlags uint32, port int, portFlags uint32, public int, publicFlags uint32, unknownFlags ExceptionFlags) {
	t.Helper()
	native, nativeFlags, port, portFlags = runGeneratedFFICase128IntBinary(operation, x, y)
	var got bool
	var gotFlags ExceptionFlags
	switch operation {
	case "quiet_equal":
		got, gotFlags = x.QuietEqual(y)
	case "quiet_not_equal":
		got, gotFlags = x.QuietNotEqual(y)
	case "quiet_less":
		got, gotFlags = x.QuietLess(y)
	case "quiet_less_equal":
		got, gotFlags = x.QuietLessEqual(y)
	case "quiet_greater":
		got, gotFlags = x.QuietGreater(y)
	case "quiet_greater_equal":
		got, gotFlags = x.QuietGreaterEqual(y)
	default:
		t.Fatalf("decimal128 unknown Tier 1 quiet predicate %q", operation)
	}
	publicFlags, unknownFlags = tier1ArithmeticPublicRawFlags(gotFlags)
	if got {
		public = 1
	}
	return
}

func tier1CheckQuiet128(t *testing.T, operation string, x, y Decimal128BID) {
	t.Helper()
	native, nativeFlags, port, portFlags, public, publicFlags, unknownFlags := tier1LegsQuiet128(t, operation, x, y)
	if unknownFlags != 0 || native != port || nativeFlags != portFlags || native != public || nativeFlags != publicFlags {
		t.Fatalf("decimal128 %s mismatch: x=%x y=%x C=%d/%08x port=%d/%08x public=%d/%08x unknown_public_flags=%s", operation, x, y, native, nativeFlags, port, portFlags, public, publicFlags, unknownFlags)
	}
}

func tier1LegsMinMax32(t *testing.T, operation string, x, y uint32) (native, port, public, pure string, unknownFlags ExceptionFlags) {
	t.Helper()
	native, port = runGeneratedFFICase32Binary("bid32_"+operation, x, y, 0)
	left, right := Decimal32BID(x), Decimal32BID(y)
	var got Decimal32BID
	var gotFlags ExceptionFlags
	var pureBits uint32
	if operation == "minnum" {
		got, gotFlags = left.MinNum(right)
		pureBits = bidgo.Bid32MinNum(x, y)
	} else if operation == "maxnum" {
		got, gotFlags = left.MaxNum(right)
		pureBits = bidgo.Bid32MaxNum(x, y)
	} else if operation == "minnum_mag" {
		got, gotFlags = left.MinNumMag(right)
		pureBits = bidgo.Bid32MinNumMag(x, y)
	} else if operation == "maxnum_mag" {
		got, gotFlags = left.MaxNumMag(right)
		pureBits = bidgo.Bid32MaxNumMag(x, y)
	} else {
		t.Fatalf("decimal32 unknown Tier 1 min/max operation %q", operation)
	}
	var publicFlags uint32
	publicFlags, unknownFlags = tier1ArithmeticPublicRawFlags(gotFlags)
	public = fmt.Sprintf("%08x/%08x", got.ToUint32(), publicFlags)
	pure = fmt.Sprintf("%08x", pureBits)
	return
}

func tier1CheckMinMax32(t *testing.T, operation string, x, y uint32) {
	t.Helper()
	native, port, public, pure, unknownFlags := tier1LegsMinMax32(t, operation, x, y)
	// The value-only port bodies (bid32_minnum_pure / bid32_maxnum_pure /
	// bid32_minnum_mag_pure / bid32_maxnum_mag_pure) are separate
	// implementations from the status-aware bodies, so their result bits are
	// compared against the native value directly.
	if unknownFlags != 0 || native != port || native != public || !strings.HasPrefix(native, pure+"/") {
		t.Fatalf("decimal32 %s mismatch: x=%08x y=%08x C=%s port=%s public=%s pure=%s unknown_public_flags=%s", operation, x, y, native, port, public, pure, unknownFlags)
	}
}

func tier1LegsMinMax64(t *testing.T, operation string, x, y uint64) (native, port, public string, unknownFlags ExceptionFlags) {
	t.Helper()
	native, port = runGeneratedFFICase64Binary("bid64_"+operation, x, y, 0)
	left, right := Decimal64BID(x), Decimal64BID(y)
	var got Decimal64BID
	var gotFlags ExceptionFlags
	if operation == "minnum" {
		got, gotFlags = left.MinNum(right)
	} else if operation == "maxnum" {
		got, gotFlags = left.MaxNum(right)
	} else if operation == "minnum_mag" {
		got, gotFlags = left.MinNumMag(right)
	} else if operation == "maxnum_mag" {
		got, gotFlags = left.MaxNumMag(right)
	} else {
		t.Fatalf("decimal64 unknown Tier 1 min/max operation %q", operation)
	}
	var publicFlags uint32
	publicFlags, unknownFlags = tier1ArithmeticPublicRawFlags(gotFlags)
	public = fmt.Sprintf("%016x/%08x", got.ToUint64(), publicFlags)
	return
}

func tier1CheckMinMax64(t *testing.T, operation string, x, y uint64) {
	t.Helper()
	native, port, public, unknownFlags := tier1LegsMinMax64(t, operation, x, y)
	if unknownFlags != 0 || native != port || native != public {
		t.Fatalf("decimal64 %s mismatch: x=%016x y=%016x C=%s port=%s public=%s unknown_public_flags=%s", operation, x, y, native, port, public, unknownFlags)
	}
}

func tier1LegsMinMax128(t *testing.T, operation string, x, y Decimal128BID) (native, port, public string, unknownFlags ExceptionFlags) {
	t.Helper()
	native, port = runGeneratedFFICase128Binary("bid128_"+operation, x, y, 0)
	var got Decimal128BID
	var gotFlags ExceptionFlags
	if operation == "minnum" {
		got, gotFlags = x.MinNum(y)
	} else if operation == "maxnum" {
		got, gotFlags = x.MaxNum(y)
	} else if operation == "minnum_mag" {
		got, gotFlags = x.MinNumMag(y)
	} else if operation == "maxnum_mag" {
		got, gotFlags = x.MaxNumMag(y)
	} else {
		t.Fatalf("decimal128 unknown Tier 1 min/max operation %q", operation)
	}
	var publicFlags uint32
	publicFlags, unknownFlags = tier1ArithmeticPublicRawFlags(gotFlags)
	public = fmt.Sprintf("%s/%08x", formatFFIUint128Bits(got), publicFlags)
	return
}

func tier1CheckMinMax128(t *testing.T, operation string, x, y Decimal128BID) {
	t.Helper()
	native, port, public, unknownFlags := tier1LegsMinMax128(t, operation, x, y)
	if unknownFlags != 0 || native != port || native != public {
		t.Fatalf("decimal128 %s mismatch: x=%x y=%x C=%s port=%s public=%s unknown_public_flags=%s", operation, x, y, native, port, public, unknownFlags)
	}
}

func tier1ExpectedQuiet(operation string, relation int) bool {
	switch operation {
	case "quiet_equal":
		return relation == 0
	case "quiet_not_equal":
		return relation != 0
	case "quiet_less":
		return relation == -1
	case "quiet_less_equal":
		return relation == -1 || relation == 0
	case "quiet_greater":
		return relation == 1
	case "quiet_greater_equal":
		return relation == 1 || relation == 0
	default:
		panic("unknown Tier 1 quiet predicate: " + operation)
	}
}

func tier1CheckQuietSemanticCase(t *testing.T, label string, relation int, wantFlags ExceptionFlags, call func(string) (bool, ExceptionFlags)) {
	t.Helper()
	for _, operation := range tier1QuietPredicates {
		got, gotFlags := call(operation)
		want := tier1ExpectedQuiet(operation, relation)
		if got != want || gotFlags != wantFlags {
			t.Fatalf("%s %s = %v/%s, want %v/%s", label, operation, got, gotFlags, want, wantFlags)
		}
	}
}

func TestTier1QuietComparisonSemanticMatrix(t *testing.T) {
	t.Run("decimal32", func(t *testing.T) {
		cases := []struct {
			name     string
			x        Decimal32BID
			y        Decimal32BID
			relation int // -1 less, 0 equal, +1 greater, 2 unordered.
			flags    ExceptionFlags
		}{
			{name: "equal finite", x: 0x32800001, y: 0x32800001, relation: 0},
			{name: "less finite", x: 0xb2800001, y: 0x32800001, relation: -1},
			{name: "greater finite", x: 0x32800001, y: 0xb2800001, relation: 1},
			{name: "signed zero", x: 0x00000000, y: 0x80000000, relation: 0},
			{name: "positive infinity", x: 0x78000000, y: 0x32800001, relation: 1},
			{name: "finite below infinity", x: 0x32800001, y: 0x78000000, relation: -1},
			{name: "quiet NaN left", x: 0x7c000001, y: 0x32800001, relation: 2},
			{name: "quiet NaN right", x: 0x32800001, y: 0x7c000001, relation: 2},
			{name: "signaling NaN left", x: 0x7e000001, y: 0x32800001, relation: 2, flags: FlagInvalidOperation},
			{name: "signaling NaN right", x: 0x32800001, y: 0x7e000001, relation: 2, flags: FlagInvalidOperation},
			{name: "negative infinity below finite", x: 0xf8000000, y: 0x32800001, relation: -1},
			{name: "finite above negative infinity", x: 0x32800001, y: 0xf8000000, relation: 1},
			{name: "negative infinity below positive infinity", x: 0xf8000000, y: 0x78000000, relation: -1},
			{name: "positive infinity above negative infinity", x: 0x78000000, y: 0xf8000000, relation: 1},
			{name: "positive infinity equal", x: 0x78000000, y: 0x78000000, relation: 0},
			{name: "negative infinity equal", x: 0xf8000000, y: 0xf8000000, relation: 0},
		}
		if len(cases) != int(tier1QuietSemanticRelationCount) {
			t.Fatalf("decimal32 semantic relation count=%d want=%d", len(cases), tier1QuietSemanticRelationCount)
		}
		for _, tc := range cases {
			tier1CheckQuietSemanticCase(t, "decimal32 "+tc.name, tc.relation, tc.flags, func(operation string) (bool, ExceptionFlags) {
				switch operation {
				case "quiet_equal":
					return tc.x.QuietEqual(tc.y)
				case "quiet_not_equal":
					return tc.x.QuietNotEqual(tc.y)
				case "quiet_less":
					return tc.x.QuietLess(tc.y)
				case "quiet_less_equal":
					return tc.x.QuietLessEqual(tc.y)
				case "quiet_greater":
					return tc.x.QuietGreater(tc.y)
				case "quiet_greater_equal":
					return tc.x.QuietGreaterEqual(tc.y)
				default:
					panic(operation)
				}
			})
		}
	})

	t.Run("decimal64", func(t *testing.T) {
		cases := []struct {
			name     string
			x        Decimal64BID
			y        Decimal64BID
			relation int
			flags    ExceptionFlags
		}{
			{name: "equal finite", x: 0x31c0000000000001, y: 0x31c0000000000001, relation: 0},
			{name: "less finite", x: 0xb1c0000000000001, y: 0x31c0000000000001, relation: -1},
			{name: "greater finite", x: 0x31c0000000000001, y: 0xb1c0000000000001, relation: 1},
			{name: "signed zero", x: 0x0000000000000000, y: 0x8000000000000000, relation: 0},
			{name: "positive infinity", x: 0x7800000000000000, y: 0x31c0000000000001, relation: 1},
			{name: "finite below infinity", x: 0x31c0000000000001, y: 0x7800000000000000, relation: -1},
			{name: "quiet NaN left", x: 0x7c00000000000001, y: 0x31c0000000000001, relation: 2},
			{name: "quiet NaN right", x: 0x31c0000000000001, y: 0x7c00000000000001, relation: 2},
			{name: "signaling NaN left", x: 0x7e00000000000001, y: 0x31c0000000000001, relation: 2, flags: FlagInvalidOperation},
			{name: "signaling NaN right", x: 0x31c0000000000001, y: 0x7e00000000000001, relation: 2, flags: FlagInvalidOperation},
			{name: "negative infinity below finite", x: 0xf800000000000000, y: 0x31c0000000000001, relation: -1},
			{name: "finite above negative infinity", x: 0x31c0000000000001, y: 0xf800000000000000, relation: 1},
			{name: "negative infinity below positive infinity", x: 0xf800000000000000, y: 0x7800000000000000, relation: -1},
			{name: "positive infinity above negative infinity", x: 0x7800000000000000, y: 0xf800000000000000, relation: 1},
			{name: "positive infinity equal", x: 0x7800000000000000, y: 0x7800000000000000, relation: 0},
			{name: "negative infinity equal", x: 0xf800000000000000, y: 0xf800000000000000, relation: 0},
		}
		if len(cases) != int(tier1QuietSemanticRelationCount) {
			t.Fatalf("decimal64 semantic relation count=%d want=%d", len(cases), tier1QuietSemanticRelationCount)
		}
		for _, tc := range cases {
			tier1CheckQuietSemanticCase(t, "decimal64 "+tc.name, tc.relation, tc.flags, func(operation string) (bool, ExceptionFlags) {
				switch operation {
				case "quiet_equal":
					return tc.x.QuietEqual(tc.y)
				case "quiet_not_equal":
					return tc.x.QuietNotEqual(tc.y)
				case "quiet_less":
					return tc.x.QuietLess(tc.y)
				case "quiet_less_equal":
					return tc.x.QuietLessEqual(tc.y)
				case "quiet_greater":
					return tc.x.QuietGreater(tc.y)
				case "quiet_greater_equal":
					return tc.x.QuietGreaterEqual(tc.y)
				default:
					panic(operation)
				}
			})
		}
	})

	t.Run("decimal128", func(t *testing.T) {
		value := func(lo, hi uint64) Decimal128BID {
			return tier1ArithmeticDecimal128(tier1Arithmetic128Words{lo: lo, hi: hi})
		}
		cases := []struct {
			name     string
			x        Decimal128BID
			y        Decimal128BID
			relation int
			flags    ExceptionFlags
		}{
			{name: "equal finite", x: value(1, 0x3040000000000000), y: value(1, 0x3040000000000000), relation: 0},
			{name: "less finite", x: value(1, 0xb040000000000000), y: value(1, 0x3040000000000000), relation: -1},
			{name: "greater finite", x: value(1, 0x3040000000000000), y: value(1, 0xb040000000000000), relation: 1},
			{name: "signed zero", x: value(0, 0), y: value(0, 0x8000000000000000), relation: 0},
			{name: "positive infinity", x: value(0, 0x7800000000000000), y: value(1, 0x3040000000000000), relation: 1},
			{name: "finite below infinity", x: value(1, 0x3040000000000000), y: value(0, 0x7800000000000000), relation: -1},
			{name: "quiet NaN left", x: value(1, 0x7c00000000000000), y: value(1, 0x3040000000000000), relation: 2},
			{name: "quiet NaN right", x: value(1, 0x3040000000000000), y: value(1, 0x7c00000000000000), relation: 2},
			{name: "signaling NaN left", x: value(1, 0x7e00000000000000), y: value(1, 0x3040000000000000), relation: 2, flags: FlagInvalidOperation},
			{name: "signaling NaN right", x: value(1, 0x3040000000000000), y: value(1, 0x7e00000000000000), relation: 2, flags: FlagInvalidOperation},
			{name: "negative infinity below finite", x: value(0, 0xf800000000000000), y: value(1, 0x3040000000000000), relation: -1},
			{name: "finite above negative infinity", x: value(1, 0x3040000000000000), y: value(0, 0xf800000000000000), relation: 1},
			{name: "negative infinity below positive infinity", x: value(0, 0xf800000000000000), y: value(0, 0x7800000000000000), relation: -1},
			{name: "positive infinity above negative infinity", x: value(0, 0x7800000000000000), y: value(0, 0xf800000000000000), relation: 1},
			{name: "positive infinity equal", x: value(0, 0x7800000000000000), y: value(0, 0x7800000000000000), relation: 0},
			{name: "negative infinity equal", x: value(0, 0xf800000000000000), y: value(0, 0xf800000000000000), relation: 0},
		}
		if len(cases) != int(tier1QuietSemanticRelationCount) {
			t.Fatalf("decimal128 semantic relation count=%d want=%d", len(cases), tier1QuietSemanticRelationCount)
		}
		for _, tc := range cases {
			tier1CheckQuietSemanticCase(t, "decimal128 "+tc.name, tc.relation, tc.flags, func(operation string) (bool, ExceptionFlags) {
				switch operation {
				case "quiet_equal":
					return tc.x.QuietEqual(tc.y)
				case "quiet_not_equal":
					return tc.x.QuietNotEqual(tc.y)
				case "quiet_less":
					return tc.x.QuietLess(tc.y)
				case "quiet_less_equal":
					return tc.x.QuietLessEqual(tc.y)
				case "quiet_greater":
					return tc.x.QuietGreater(tc.y)
				case "quiet_greater_equal":
					return tc.x.QuietGreaterEqual(tc.y)
				default:
					panic(operation)
				}
			})
		}
	})
}

func TestTier1ComparisonMinMaxStructuredNativeDifferential(t *testing.T) {
	requireNative(t)
	if len(tier1ArithmeticBoundary32) != int(tier1CompareConversionBoundary32Count) ||
		len(tier1ArithmeticBoundary64) != int(tier1CompareConversionBoundary64Count) ||
		len(tier1ArithmeticBoundary128) != int(tier1CompareConversionBoundary128Count) {
		t.Fatal("shared Tier 1 boundary inventory count mismatch")
	}
	shard := tier1CompareConversionLoadShard(t)

	t.Run("decimal32", func(t *testing.T) {
		var comparison uint64
		tier1ArithmeticVisitPairs32(func(x, y uint32) {
			for _, operation := range tier1QuietPredicates {
				if shard.owns(comparison) {
					tier1CheckQuiet32(t, operation, x, y)
				}
				comparison++
			}
			for _, operation := range tier1MinMaxOperations {
				if shard.owns(comparison) {
					tier1CheckMinMax32(t, operation, x, y)
				}
				comparison++
			}
		})
		if comparison != tier1CompareStructured32 {
			t.Fatalf("decimal32 structured compare/minmax comparisons=%d want=%d", comparison, tier1CompareStructured32)
		}
		t.Logf("decimal32 structured compare/minmax exact comparisons: %d/%d", shard.ownedCount(comparison), comparison)
	})

	t.Run("decimal64", func(t *testing.T) {
		var comparison uint64
		tier1ArithmeticVisitPairs64(func(x, y uint64) {
			for _, operation := range tier1QuietPredicates {
				if shard.owns(comparison) {
					tier1CheckQuiet64(t, operation, x, y)
				}
				comparison++
			}
			for _, operation := range tier1MinMaxOperations {
				if shard.owns(comparison) {
					tier1CheckMinMax64(t, operation, x, y)
				}
				comparison++
			}
		})
		if comparison != tier1CompareStructured64 {
			t.Fatalf("decimal64 structured compare/minmax comparisons=%d want=%d", comparison, tier1CompareStructured64)
		}
		t.Logf("decimal64 structured compare/minmax exact comparisons: %d/%d", shard.ownedCount(comparison), comparison)
	})

	t.Run("decimal128", func(t *testing.T) {
		var comparison uint64
		tier1ArithmeticVisitPairs128(func(xWords, yWords tier1Arithmetic128Words) {
			x := tier1ArithmeticDecimal128(xWords)
			y := tier1ArithmeticDecimal128(yWords)
			for _, operation := range tier1QuietPredicates {
				if shard.owns(comparison) {
					tier1CheckQuiet128(t, operation, x, y)
				}
				comparison++
			}
			for _, operation := range tier1MinMaxOperations {
				if shard.owns(comparison) {
					tier1CheckMinMax128(t, operation, x, y)
				}
				comparison++
			}
		})
		if comparison != tier1CompareStructured128 {
			t.Fatalf("decimal128 structured compare/minmax comparisons=%d want=%d", comparison, tier1CompareStructured128)
		}
		t.Logf("decimal128 structured compare/minmax exact comparisons: %d/%d", shard.ownedCount(comparison), comparison)
	})
}

func TestTier1ComparisonMinMaxDeterministicRandomNativeDifferential(t *testing.T) {
	requireNative(t)
	shard := tier1CompareConversionLoadShard(t)
	t.Run("decimal32", func(t *testing.T) {
		var comparison uint64
		var randomConsumed uint64
		randomDigest := tier1ArithmeticScaleTupleHashMix(tier1ArithmeticScaleTupleHashOffset, 32)
		for i := uint64(0); i < tier1CompareRandomPairs32; i++ {
			x := tier1ArithmeticRandomOperand32(tier1CompareRandomSeed32, i, 0)
			y := tier1ArithmeticRandomOperand32(tier1CompareRandomSeed32, i, 1)
			randomDigest = tier1ArithmeticMixOperand32(randomDigest, x)
			randomDigest = tier1ArithmeticMixOperand32(randomDigest, y)
			randomConsumed++
			for _, operation := range tier1QuietPredicates {
				if shard.owns(comparison) {
					tier1CheckQuiet32(t, operation, x, y)
				}
				comparison++
			}
			for _, operation := range tier1MinMaxOperations {
				if shard.owns(comparison) {
					tier1CheckMinMax32(t, operation, x, y)
				}
				comparison++
			}
		}
		if got := tier1ArithmeticScaleTupleHashMix(randomDigest, randomConsumed); got != tier1CompareRandomStreamHash32 {
			t.Fatalf("decimal32 random compare/minmax stream hash=%d want=%d (consumed operands diverge from the anchored stream)", got, tier1CompareRandomStreamHash32)
		}
		if comparison != tier1CompareRandomComparisons32 {
			t.Fatalf("decimal32 random compare/minmax comparisons=%d want=%d", comparison, tier1CompareRandomComparisons32)
		}
		t.Logf("decimal32 random compare/minmax exact comparisons: %d/%d", shard.ownedCount(comparison), comparison)
	})
	t.Run("decimal64", func(t *testing.T) {
		var comparison uint64
		var randomConsumed uint64
		randomDigest := tier1ArithmeticScaleTupleHashMix(tier1ArithmeticScaleTupleHashOffset, 64)
		for i := uint64(0); i < tier1CompareRandomPairs64; i++ {
			x := tier1ArithmeticRandomOperand64(tier1CompareRandomSeed64, i, 0)
			y := tier1ArithmeticRandomOperand64(tier1CompareRandomSeed64, i, 1)
			randomDigest = tier1ArithmeticMixOperand64(randomDigest, x)
			randomDigest = tier1ArithmeticMixOperand64(randomDigest, y)
			randomConsumed++
			for _, operation := range tier1QuietPredicates {
				if shard.owns(comparison) {
					tier1CheckQuiet64(t, operation, x, y)
				}
				comparison++
			}
			for _, operation := range tier1MinMaxOperations {
				if shard.owns(comparison) {
					tier1CheckMinMax64(t, operation, x, y)
				}
				comparison++
			}
		}
		if got := tier1ArithmeticScaleTupleHashMix(randomDigest, randomConsumed); got != tier1CompareRandomStreamHash64 {
			t.Fatalf("decimal64 random compare/minmax stream hash=%d want=%d (consumed operands diverge from the anchored stream)", got, tier1CompareRandomStreamHash64)
		}
		if comparison != tier1CompareRandomComparisons64 {
			t.Fatalf("decimal64 random compare/minmax comparisons=%d want=%d", comparison, tier1CompareRandomComparisons64)
		}
		t.Logf("decimal64 random compare/minmax exact comparisons: %d/%d", shard.ownedCount(comparison), comparison)
	})
	t.Run("decimal128", func(t *testing.T) {
		var comparison uint64
		var randomConsumed uint64
		randomDigest := tier1ArithmeticScaleTupleHashMix(tier1ArithmeticScaleTupleHashOffset, 128)
		for i := uint64(0); i < tier1CompareRandomPairs128; i++ {
			xWords := tier1ArithmeticRandomOperand128(tier1CompareRandomSeed128, i, 0)
			yWords := tier1ArithmeticRandomOperand128(tier1CompareRandomSeed128, i, 1)
			randomDigest = tier1ArithmeticMixOperand128(randomDigest, xWords)
			randomDigest = tier1ArithmeticMixOperand128(randomDigest, yWords)
			randomConsumed++
			x := tier1ArithmeticDecimal128(xWords)
			y := tier1ArithmeticDecimal128(yWords)
			for _, operation := range tier1QuietPredicates {
				if shard.owns(comparison) {
					tier1CheckQuiet128(t, operation, x, y)
				}
				comparison++
			}
			for _, operation := range tier1MinMaxOperations {
				if shard.owns(comparison) {
					tier1CheckMinMax128(t, operation, x, y)
				}
				comparison++
			}
		}
		if got := tier1ArithmeticScaleTupleHashMix(randomDigest, randomConsumed); got != tier1CompareRandomStreamHash128 {
			t.Fatalf("decimal128 random compare/minmax stream hash=%d want=%d (consumed operands diverge from the anchored stream)", got, tier1CompareRandomStreamHash128)
		}
		if comparison != tier1CompareRandomComparisons128 {
			t.Fatalf("decimal128 random compare/minmax comparisons=%d want=%d", comparison, tier1CompareRandomComparisons128)
		}
		t.Logf("decimal128 random compare/minmax exact comparisons: %d/%d", shard.ownedCount(comparison), comparison)
	})
}

type tier1ToIntegerOperation struct {
	kind     string
	exact    bool
	mode     tier1ArithmeticMode
	suffix   string
}

func tier1ToIntegerOperations() []tier1ToIntegerOperation {
	kinds := []string{"int8", "int16", "int32", "int64", "uint8", "uint16", "uint32", "uint64"}
	operations := make([]tier1ToIntegerOperation, 0, tier1ToIntegerOperationCount)
	for _, kind := range kinds {
		for _, exact := range []bool{false, true} {
			for _, mode := range tier1ArithmeticModes {
				suffix := ""
				switch mode.public {
				case RoundNearestEven:
					suffix = "rnint"
				case RoundNearestAway:
					suffix = "rninta"
				case RoundTowardZero:
					suffix = "int"
				case RoundTowardPositive:
					suffix = "ceil"
				case RoundTowardNegative:
					suffix = "floor"
				default:
					panic("unknown Tier 1 public rounding mode")
				}
				if exact {
					suffix = "x" + suffix
				}
				operations = append(operations, tier1ToIntegerOperation{
					kind: kind, exact: exact, mode: mode, suffix: suffix,
				})
			}
		}
	}
	return operations
}

func tier1FormatPublicInteger(value any, flags ExceptionFlags) (string, ExceptionFlags) {
	rawFlags, unknownFlags := tier1ArithmeticPublicRawFlags(flags)
	return fmt.Sprintf("%d/%s", value, formatReadtestStatus(rawFlags)), unknownFlags
}

func tier1PublicToInteger32(value Decimal32BID, operation tier1ToIntegerOperation) (string, ExceptionFlags) {
	if operation.exact {
		switch operation.kind {
		case "int8":
			result, flags := value.ConvertToInt8Exact(operation.mode.public)
			return tier1FormatPublicInteger(result, flags)
		case "int16":
			result, flags := value.ConvertToInt16Exact(operation.mode.public)
			return tier1FormatPublicInteger(result, flags)
		case "int32":
			result, flags := value.ConvertToInt32Exact(operation.mode.public)
			return tier1FormatPublicInteger(result, flags)
		case "int64":
			result, flags := value.ConvertToInt64Exact(operation.mode.public)
			return tier1FormatPublicInteger(result, flags)
		case "uint8":
			result, flags := value.ConvertToUint8Exact(operation.mode.public)
			return tier1FormatPublicInteger(result, flags)
		case "uint16":
			result, flags := value.ConvertToUint16Exact(operation.mode.public)
			return tier1FormatPublicInteger(result, flags)
		case "uint32":
			result, flags := value.ConvertToUint32Exact(operation.mode.public)
			return tier1FormatPublicInteger(result, flags)
		case "uint64":
			result, flags := value.ConvertToUint64Exact(operation.mode.public)
			return tier1FormatPublicInteger(result, flags)
		}
	} else {
		switch operation.kind {
		case "int8":
			result, flags := value.ConvertToInt8(operation.mode.public)
			return tier1FormatPublicInteger(result, flags)
		case "int16":
			result, flags := value.ConvertToInt16(operation.mode.public)
			return tier1FormatPublicInteger(result, flags)
		case "int32":
			result, flags := value.ConvertToInt32(operation.mode.public)
			return tier1FormatPublicInteger(result, flags)
		case "int64":
			result, flags := value.ConvertToInt64(operation.mode.public)
			return tier1FormatPublicInteger(result, flags)
		case "uint8":
			result, flags := value.ConvertToUint8(operation.mode.public)
			return tier1FormatPublicInteger(result, flags)
		case "uint16":
			result, flags := value.ConvertToUint16(operation.mode.public)
			return tier1FormatPublicInteger(result, flags)
		case "uint32":
			result, flags := value.ConvertToUint32(operation.mode.public)
			return tier1FormatPublicInteger(result, flags)
		case "uint64":
			result, flags := value.ConvertToUint64(operation.mode.public)
			return tier1FormatPublicInteger(result, flags)
		}
	}
	panic("unknown Decimal32 Tier 1 integer conversion kind: " + operation.kind)
}

func tier1PublicToInteger64(value Decimal64BID, operation tier1ToIntegerOperation) (string, ExceptionFlags) {
	if operation.exact {
		switch operation.kind {
		case "int8":
			result, flags := value.ConvertToInt8Exact(operation.mode.public)
			return tier1FormatPublicInteger(result, flags)
		case "int16":
			result, flags := value.ConvertToInt16Exact(operation.mode.public)
			return tier1FormatPublicInteger(result, flags)
		case "int32":
			result, flags := value.ConvertToInt32Exact(operation.mode.public)
			return tier1FormatPublicInteger(result, flags)
		case "int64":
			result, flags := value.ConvertToInt64Exact(operation.mode.public)
			return tier1FormatPublicInteger(result, flags)
		case "uint8":
			result, flags := value.ConvertToUint8Exact(operation.mode.public)
			return tier1FormatPublicInteger(result, flags)
		case "uint16":
			result, flags := value.ConvertToUint16Exact(operation.mode.public)
			return tier1FormatPublicInteger(result, flags)
		case "uint32":
			result, flags := value.ConvertToUint32Exact(operation.mode.public)
			return tier1FormatPublicInteger(result, flags)
		case "uint64":
			result, flags := value.ConvertToUint64Exact(operation.mode.public)
			return tier1FormatPublicInteger(result, flags)
		}
	} else {
		switch operation.kind {
		case "int8":
			result, flags := value.ConvertToInt8(operation.mode.public)
			return tier1FormatPublicInteger(result, flags)
		case "int16":
			result, flags := value.ConvertToInt16(operation.mode.public)
			return tier1FormatPublicInteger(result, flags)
		case "int32":
			result, flags := value.ConvertToInt32(operation.mode.public)
			return tier1FormatPublicInteger(result, flags)
		case "int64":
			result, flags := value.ConvertToInt64(operation.mode.public)
			return tier1FormatPublicInteger(result, flags)
		case "uint8":
			result, flags := value.ConvertToUint8(operation.mode.public)
			return tier1FormatPublicInteger(result, flags)
		case "uint16":
			result, flags := value.ConvertToUint16(operation.mode.public)
			return tier1FormatPublicInteger(result, flags)
		case "uint32":
			result, flags := value.ConvertToUint32(operation.mode.public)
			return tier1FormatPublicInteger(result, flags)
		case "uint64":
			result, flags := value.ConvertToUint64(operation.mode.public)
			return tier1FormatPublicInteger(result, flags)
		}
	}
	panic("unknown Decimal64 Tier 1 integer conversion kind: " + operation.kind)
}

func tier1PublicToInteger128(value Decimal128BID, operation tier1ToIntegerOperation) (string, ExceptionFlags) {
	if operation.exact {
		switch operation.kind {
		case "int8":
			result, flags := value.ConvertToInt8Exact(operation.mode.public)
			return tier1FormatPublicInteger(result, flags)
		case "int16":
			result, flags := value.ConvertToInt16Exact(operation.mode.public)
			return tier1FormatPublicInteger(result, flags)
		case "int32":
			result, flags := value.ConvertToInt32Exact(operation.mode.public)
			return tier1FormatPublicInteger(result, flags)
		case "int64":
			result, flags := value.ConvertToInt64Exact(operation.mode.public)
			return tier1FormatPublicInteger(result, flags)
		case "uint8":
			result, flags := value.ConvertToUint8Exact(operation.mode.public)
			return tier1FormatPublicInteger(result, flags)
		case "uint16":
			result, flags := value.ConvertToUint16Exact(operation.mode.public)
			return tier1FormatPublicInteger(result, flags)
		case "uint32":
			result, flags := value.ConvertToUint32Exact(operation.mode.public)
			return tier1FormatPublicInteger(result, flags)
		case "uint64":
			result, flags := value.ConvertToUint64Exact(operation.mode.public)
			return tier1FormatPublicInteger(result, flags)
		}
	} else {
		switch operation.kind {
		case "int8":
			result, flags := value.ConvertToInt8(operation.mode.public)
			return tier1FormatPublicInteger(result, flags)
		case "int16":
			result, flags := value.ConvertToInt16(operation.mode.public)
			return tier1FormatPublicInteger(result, flags)
		case "int32":
			result, flags := value.ConvertToInt32(operation.mode.public)
			return tier1FormatPublicInteger(result, flags)
		case "int64":
			result, flags := value.ConvertToInt64(operation.mode.public)
			return tier1FormatPublicInteger(result, flags)
		case "uint8":
			result, flags := value.ConvertToUint8(operation.mode.public)
			return tier1FormatPublicInteger(result, flags)
		case "uint16":
			result, flags := value.ConvertToUint16(operation.mode.public)
			return tier1FormatPublicInteger(result, flags)
		case "uint32":
			result, flags := value.ConvertToUint32(operation.mode.public)
			return tier1FormatPublicInteger(result, flags)
		case "uint64":
			result, flags := value.ConvertToUint64(operation.mode.public)
			return tier1FormatPublicInteger(result, flags)
		}
	}
	panic("unknown Decimal128 Tier 1 integer conversion kind: " + operation.kind)
}

// tier1LegsToInteger computes the Intel C and Go-port legs of one integer
// conversion from the shared function-name/operand fan-out. Together with the
// per-width tier1LegsToInteger* wrappers (which add the public leg) it is the
// glue the routing sentinels bind.
func tier1LegsToInteger(t *testing.T, width int, operand string, operation tier1ToIntegerOperation) (function, native, port string) {
	t.Helper()
	function = fmt.Sprintf("bid%d_to_%s_%s", width, operation.kind, operation.suffix)
	tc := generatedFFICase{
		Format:    fmt.Sprintf("decimal%d", width),
		Operation: fmt.Sprintf("to_%s_%s", operation.kind, operation.suffix),
		Function:  function,
		Operands:  []string{operand},
	}
	var err error
	native, err = runGeneratedFFICaseNativeBaseIntegerTo(tc)
	if err != nil {
		t.Fatalf("%s native integer conversion: %v", function, err)
	}
	port, err = runGeneratedFFICaseGoBaseIntegerTo(tc)
	if err != nil {
		t.Fatalf("%s Go-port integer conversion: %v", function, err)
	}
	return
}

func tier1CheckToInteger(t *testing.T, width int, operand string, operation tier1ToIntegerOperation, public string, unknownFlags ExceptionFlags) {
	t.Helper()
	function, native, port := tier1LegsToInteger(t, width, operand, operation)
	if unknownFlags != 0 || native != port || native != public {
		t.Fatalf("%s mismatch: operand=%s mode=%s exact=%v C=%s port=%s public=%s unknown_public_flags=%s", function, operand, operation.mode.name, operation.exact, native, port, public, unknownFlags)
	}
}

func tier1LegsToInteger32(t *testing.T, value uint32, operation tier1ToIntegerOperation) (native, port, public string, unknownFlags ExceptionFlags) {
	t.Helper()
	public, unknownFlags = tier1PublicToInteger32(Decimal32BID(value), operation)
	_, native, port = tier1LegsToInteger(t, 32, fmt.Sprintf("%08x", value), operation)
	return
}

func tier1LegsToInteger64(t *testing.T, value uint64, operation tier1ToIntegerOperation) (native, port, public string, unknownFlags ExceptionFlags) {
	t.Helper()
	public, unknownFlags = tier1PublicToInteger64(Decimal64BID(value), operation)
	_, native, port = tier1LegsToInteger(t, 64, fmt.Sprintf("%016x", value), operation)
	return
}

func tier1LegsToInteger128(t *testing.T, value Decimal128BID, operation tier1ToIntegerOperation) (native, port, public string, unknownFlags ExceptionFlags) {
	t.Helper()
	public, unknownFlags = tier1PublicToInteger128(value, operation)
	raw := value.ToBytes()
	operand := fmt.Sprintf("%016x%016x", binary.LittleEndian.Uint64(raw[8:16]), binary.LittleEndian.Uint64(raw[0:8]))
	_, native, port = tier1LegsToInteger(t, 128, operand, operation)
	return
}

func tier1CheckToInteger32(t *testing.T, value uint32, operation tier1ToIntegerOperation) {
	public, unknownFlags := tier1PublicToInteger32(Decimal32BID(value), operation)
	tier1CheckToInteger(t, 32, fmt.Sprintf("%08x", value), operation, public, unknownFlags)
}

func tier1CheckToInteger64(t *testing.T, value uint64, operation tier1ToIntegerOperation) {
	public, unknownFlags := tier1PublicToInteger64(Decimal64BID(value), operation)
	tier1CheckToInteger(t, 64, fmt.Sprintf("%016x", value), operation, public, unknownFlags)
}

func tier1CheckToInteger128(t *testing.T, value Decimal128BID, operation tier1ToIntegerOperation) {
	public, unknownFlags := tier1PublicToInteger128(value, operation)
	raw := value.ToBytes()
	operand := fmt.Sprintf("%016x%016x", binary.LittleEndian.Uint64(raw[8:16]), binary.LittleEndian.Uint64(raw[0:8]))
	tier1CheckToInteger(t, 128, operand, operation, public, unknownFlags)
}

type tier1BinaryOperation struct {
	source int
	dest   int
	mode   tier1ArithmeticMode
}

func tier1BinaryOperations(source int) []tier1BinaryOperation {
	var operations []tier1BinaryOperation
	for _, dest := range []int{32, 64, 128} {
		for _, mode := range tier1ArithmeticModes {
			operations = append(operations, tier1BinaryOperation{source: source, dest: dest, mode: mode})
		}
	}
	return operations
}

// tier1LegsBinaryConversion computes the Intel C and Go-port legs of one
// BID-to-binary conversion from the shared function-name/operand/mode
// fan-out; the routing sentinels bind it together with the per-width public
// dispatchers.
func tier1LegsBinaryConversion(t *testing.T, operation tier1BinaryOperation, operand string) (function, native, port string) {
	t.Helper()
	function = fmt.Sprintf("bid%d_to_binary%d", operation.source, operation.dest)
	tc := generatedFFICase{
		Format:    fmt.Sprintf("decimal%d", operation.source),
		Operation: fmt.Sprintf("to_binary%d", operation.dest),
		Function:  function,
		Rounding:  operation.mode.native,
		Operands:  []string{operand},
	}
	var err error
	native, port, err = runGeneratedFFICaseBinaryConversion(tc)
	if err != nil {
		t.Fatalf("%s native/port binary conversion: %v", function, err)
	}
	return
}

func tier1CheckBinaryConversion(t *testing.T, operation tier1BinaryOperation, operand string, public string, unknownFlags ExceptionFlags) {
	t.Helper()
	function, native, port := tier1LegsBinaryConversion(t, operation, operand)
	if unknownFlags != 0 || native != port || native != public {
		t.Fatalf("%s mismatch: operand=%s mode=%s C=%s port=%s public=%s unknown_public_flags=%s", function, operand, operation.mode.name, native, port, public, unknownFlags)
	}
}

func tier1PublicBinary32(source Decimal32BID, operation tier1BinaryOperation) (string, ExceptionFlags) {
	switch operation.dest {
	case 32:
		result, flags := source.ToBinary32(operation.mode.public)
		rawFlags, unknown := tier1ArithmeticPublicRawFlags(flags)
		return fmt.Sprintf("%08x/%08x", math.Float32bits(result), rawFlags), unknown
	case 64:
		result, flags := source.ToBinary64(operation.mode.public)
		rawFlags, unknown := tier1ArithmeticPublicRawFlags(flags)
		return fmt.Sprintf("%016x/%08x", math.Float64bits(result), rawFlags), unknown
	case 128:
		result, flags := source.ToBinary128(operation.mode.public)
		rawFlags, unknown := tier1ArithmeticPublicRawFlags(flags)
		return fmt.Sprintf("%s/%08x", formatFFIUint128Bits(Decimal128BID(result.ToBytes())), rawFlags), unknown
	}
	panic(fmt.Sprintf("unknown Decimal32 binary-conversion destination %d", operation.dest))
}

func tier1PublicBinary64(source Decimal64BID, operation tier1BinaryOperation) (string, ExceptionFlags) {
	switch operation.dest {
	case 32:
		result, flags := source.ToBinary32(operation.mode.public)
		rawFlags, unknown := tier1ArithmeticPublicRawFlags(flags)
		return fmt.Sprintf("%08x/%08x", math.Float32bits(result), rawFlags), unknown
	case 64:
		result, flags := source.ToBinary64(operation.mode.public)
		rawFlags, unknown := tier1ArithmeticPublicRawFlags(flags)
		return fmt.Sprintf("%016x/%08x", math.Float64bits(result), rawFlags), unknown
	case 128:
		result, flags := source.ToBinary128(operation.mode.public)
		rawFlags, unknown := tier1ArithmeticPublicRawFlags(flags)
		return fmt.Sprintf("%s/%08x", formatFFIUint128Bits(Decimal128BID(result.ToBytes())), rawFlags), unknown
	}
	panic(fmt.Sprintf("unknown Decimal64 binary-conversion destination %d", operation.dest))
}

func tier1PublicBinary128(source Decimal128BID, operation tier1BinaryOperation) (string, ExceptionFlags) {
	switch operation.dest {
	case 32:
		result, flags := source.ToBinary32(operation.mode.public)
		rawFlags, unknown := tier1ArithmeticPublicRawFlags(flags)
		return fmt.Sprintf("%08x/%08x", math.Float32bits(result), rawFlags), unknown
	case 64:
		result, flags := source.ToBinary64(operation.mode.public)
		rawFlags, unknown := tier1ArithmeticPublicRawFlags(flags)
		return fmt.Sprintf("%016x/%08x", math.Float64bits(result), rawFlags), unknown
	case 128:
		result, flags := source.ToBinary128(operation.mode.public)
		rawFlags, unknown := tier1ArithmeticPublicRawFlags(flags)
		return fmt.Sprintf("%s/%08x", formatFFIUint128Bits(Decimal128BID(result.ToBytes())), rawFlags), unknown
	}
	panic(fmt.Sprintf("unknown Decimal128 binary-conversion destination %d", operation.dest))
}

func tier1CheckBinary32(t *testing.T, value uint32, operation tier1BinaryOperation) {
	public, unknownFlags := tier1PublicBinary32(Decimal32BID(value), operation)
	tier1CheckBinaryConversion(t, operation, fmt.Sprintf("%08x", value), public, unknownFlags)
}

func tier1CheckBinary64(t *testing.T, value uint64, operation tier1BinaryOperation) {
	public, unknownFlags := tier1PublicBinary64(Decimal64BID(value), operation)
	tier1CheckBinaryConversion(t, operation, fmt.Sprintf("%016x", value), public, unknownFlags)
}

func tier1CheckBinary128(t *testing.T, value Decimal128BID, operation tier1BinaryOperation) {
	public, unknownFlags := tier1PublicBinary128(value, operation)
	tier1CheckBinaryConversion(t, operation, formatFFIUint128Bits(value), public, unknownFlags)
}

type tier1WidthOperation struct {
	source  int
	dest    int
	rounded bool
	mode    tier1ArithmeticMode
}

func tier1WidthOperations(source int) []tier1WidthOperation {
	var operations []tier1WidthOperation
	addRounded := func(dest int) {
		for _, mode := range tier1ArithmeticModes {
			operations = append(operations, tier1WidthOperation{source: source, dest: dest, rounded: true, mode: mode})
		}
	}
	switch source {
	case 32:
		operations = append(operations,
			tier1WidthOperation{source: 32, dest: 64},
			tier1WidthOperation{source: 32, dest: 128},
		)
	case 64:
		addRounded(32)
		operations = append(operations, tier1WidthOperation{source: 64, dest: 128})
	case 128:
		addRounded(32)
		addRounded(64)
	default:
		panic(fmt.Sprintf("unknown Tier 1 width-conversion source %d", source))
	}
	return operations
}

// tier1LegsWidthConversion computes the Intel C and Go-port legs of one BID
// width conversion from the shared function-name/operand/mode fan-out; the
// routing sentinels bind it together with the per-width public dispatchers.
func tier1LegsWidthConversion(t *testing.T, operation tier1WidthOperation, operand string) (function, native, port string) {
	t.Helper()
	function = fmt.Sprintf("bid%d_to_bid%d", operation.source, operation.dest)
	rounding := 0
	if operation.rounded {
		rounding = operation.mode.native
	}
	tc := generatedFFICase{
		Format:    fmt.Sprintf("decimal%d", operation.source),
		Operation: fmt.Sprintf("to_bid%d", operation.dest),
		Function:  function,
		Rounding:  rounding,
		Operands:  []string{operand},
	}
	var err error
	native, port, err = runGeneratedFFICaseBIDWidthConversion(tc)
	if err != nil {
		t.Fatalf("%s native/port width conversion: %v", function, err)
	}
	return
}

func tier1CheckWidthConversion(t *testing.T, operation tier1WidthOperation, operand string, public string, unknownFlags ExceptionFlags) {
	t.Helper()
	function, native, port := tier1LegsWidthConversion(t, operation, operand)
	if unknownFlags != 0 || native != port || native != public {
		t.Fatalf("%s mismatch: operand=%s mode=%s C=%s port=%s public=%s unknown_public_flags=%s", function, operand, operation.mode.name, native, port, public, unknownFlags)
	}
}

func tier1PublicWidth32(t *testing.T, value uint32, operation tier1WidthOperation) (string, ExceptionFlags) {
	t.Helper()
	source := Decimal32BID(value)
	switch operation.dest {
	case 64:
		result, flags := source.ToDecimal64()
		rawFlags, unknown := tier1ArithmeticPublicRawFlags(flags)
		return fmt.Sprintf("%016x/%08x", result.ToUint64(), rawFlags), unknown
	case 128:
		result, flags := source.ToDecimal128()
		rawFlags, unknown := tier1ArithmeticPublicRawFlags(flags)
		return fmt.Sprintf("%s/%08x", formatFFIUint128Bits(result), rawFlags), unknown
	default:
		t.Fatalf("unknown Decimal32 width-conversion destination %d", operation.dest)
		return "", 0
	}
}

func tier1PublicWidth64(t *testing.T, value uint64, operation tier1WidthOperation) (string, ExceptionFlags) {
	t.Helper()
	source := Decimal64BID(value)
	switch operation.dest {
	case 32:
		result, flags := source.ToDecimal32(operation.mode.public)
		rawFlags, unknown := tier1ArithmeticPublicRawFlags(flags)
		return fmt.Sprintf("%08x/%08x", result.ToUint32(), rawFlags), unknown
	case 128:
		result, flags := source.ToDecimal128()
		rawFlags, unknown := tier1ArithmeticPublicRawFlags(flags)
		return fmt.Sprintf("%s/%08x", formatFFIUint128Bits(result), rawFlags), unknown
	default:
		t.Fatalf("unknown Decimal64 width-conversion destination %d", operation.dest)
		return "", 0
	}
}

func tier1PublicWidth128(t *testing.T, value Decimal128BID, operation tier1WidthOperation) (string, ExceptionFlags) {
	t.Helper()
	switch operation.dest {
	case 32:
		result, flags := value.ToDecimal32(operation.mode.public)
		rawFlags, unknown := tier1ArithmeticPublicRawFlags(flags)
		return fmt.Sprintf("%08x/%08x", result.ToUint32(), rawFlags), unknown
	case 64:
		result, flags := value.ToDecimal64(operation.mode.public)
		rawFlags, unknown := tier1ArithmeticPublicRawFlags(flags)
		return fmt.Sprintf("%016x/%08x", result.ToUint64(), rawFlags), unknown
	default:
		t.Fatalf("unknown Decimal128 width-conversion destination %d", operation.dest)
		return "", 0
	}
}

func tier1CheckWidth32(t *testing.T, value uint32, operation tier1WidthOperation) {
	public, unknownFlags := tier1PublicWidth32(t, value, operation)
	tier1CheckWidthConversion(t, operation, fmt.Sprintf("%08x", value), public, unknownFlags)
}

func tier1CheckWidth64(t *testing.T, value uint64, operation tier1WidthOperation) {
	public, unknownFlags := tier1PublicWidth64(t, value, operation)
	tier1CheckWidthConversion(t, operation, fmt.Sprintf("%016x", value), public, unknownFlags)
}

func tier1CheckWidth128(t *testing.T, value Decimal128BID, operation tier1WidthOperation) {
	public, unknownFlags := tier1PublicWidth128(t, value, operation)
	tier1CheckWidthConversion(t, operation, formatFFIUint128Bits(value), public, unknownFlags)
}

type tier1ConstructorOperation struct {
	dest    int
	kind    string
	rounded bool
	mode    tier1ArithmeticMode
}

func tier1ConstructorOperations() []tier1ConstructorOperation {
	operations := make([]tier1ConstructorOperation, 0, tier1ConstructorOperationCount)
	for _, kind := range []string{"int32", "uint32", "int64", "uint64"} {
		for _, mode := range tier1ArithmeticModes {
			operations = append(operations, tier1ConstructorOperation{dest: 32, kind: kind, rounded: true, mode: mode})
		}
	}
	operations = append(operations,
		tier1ConstructorOperation{dest: 64, kind: "int32"},
		tier1ConstructorOperation{dest: 64, kind: "uint32"},
	)
	for _, kind := range []string{"int64", "uint64"} {
		for _, mode := range tier1ArithmeticModes {
			operations = append(operations, tier1ConstructorOperation{dest: 64, kind: kind, rounded: true, mode: mode})
		}
	}
	for _, kind := range []string{"int32", "uint32", "int64", "uint64"} {
		operations = append(operations, tier1ConstructorOperation{dest: 128, kind: kind})
	}
	return operations
}

func tier1ConstructorOperand(raw uint64, kind string) string {
	switch kind {
	case "int32":
		return strconv.FormatInt(int64(int32(raw)), 10)
	case "uint32":
		return strconv.FormatUint(uint64(uint32(raw)), 10)
	case "int64":
		return strconv.FormatInt(int64(raw), 10)
	case "uint64":
		return strconv.FormatUint(raw, 10)
	default:
		panic("unknown Tier 1 constructor kind: " + kind)
	}
}

func tier1PublicConstructor(operation tier1ConstructorOperation, raw uint64) (string, ExceptionFlags) {
	switch operation.dest {
	case 32:
		var result Decimal32BID
		var flags ExceptionFlags
		switch operation.kind {
		case "int32":
			result, flags = NewDecimal32FromInt32(int32(raw), operation.mode.public)
		case "uint32":
			result, flags = NewDecimal32FromUint32(uint32(raw), operation.mode.public)
		case "int64":
			result, flags = NewDecimal32FromInt64(int64(raw), operation.mode.public)
		case "uint64":
			result, flags = NewDecimal32FromUint64(raw, operation.mode.public)
		default:
			panic(operation.kind)
		}
		rawFlags, unknownFlags := tier1ArithmeticPublicRawFlags(flags)
		return fmt.Sprintf("%08x/%08x", result.ToUint32(), rawFlags), unknownFlags
	case 64:
		switch operation.kind {
		case "int32":
			return fmt.Sprintf("%016x", NewDecimal64FromInt32(int32(raw)).ToUint64()), 0
		case "uint32":
			return fmt.Sprintf("%016x", NewDecimal64FromUint32(uint32(raw)).ToUint64()), 0
		case "int64":
			result, flags := NewDecimal64FromInt64(int64(raw), operation.mode.public)
			rawFlags, unknownFlags := tier1ArithmeticPublicRawFlags(flags)
			return fmt.Sprintf("%016x/%08x", result.ToUint64(), rawFlags), unknownFlags
		case "uint64":
			result, flags := NewDecimal64FromUint64(raw, operation.mode.public)
			rawFlags, unknownFlags := tier1ArithmeticPublicRawFlags(flags)
			return fmt.Sprintf("%016x/%08x", result.ToUint64(), rawFlags), unknownFlags
		default:
			panic(operation.kind)
		}
	case 128:
		var result Decimal128BID
		switch operation.kind {
		case "int32":
			result = NewDecimal128FromInt32(int32(raw))
		case "uint32":
			result = NewDecimal128FromUint32(uint32(raw))
		case "int64":
			result = NewDecimal128FromInt64(int64(raw))
		case "uint64":
			result = NewDecimal128FromUint64(raw)
		default:
			panic(operation.kind)
		}
		return formatFFIUint128Bits(result), 0
	default:
		panic(fmt.Sprintf("unknown Tier 1 constructor destination %d", operation.dest))
	}
}

// tier1LegsConstructor computes every leg of one integer constructor from the
// shared register/kind-cast/mode fan-out; the routing sentinels bind it.
func tier1LegsConstructor(t *testing.T, operation tier1ConstructorOperation, raw uint64) (function, operand, native, port, public string, unknownFlags ExceptionFlags) {
	t.Helper()
	function = fmt.Sprintf("bid%d_from_%s", operation.dest, operation.kind)
	rounding := 0
	if operation.rounded {
		rounding = operation.mode.native
	}
	operand = tier1ConstructorOperand(raw, operation.kind)
	tc := generatedFFICase{
		Format:    fmt.Sprintf("decimal%d", operation.dest),
		Operation: "from_" + operation.kind,
		Function:  function,
		Rounding:  rounding,
		Operands:  []string{operand},
	}
	var err error
	native, port, err = runGeneratedFFICaseBaseIntegerFrom(tc)
	if err != nil {
		t.Fatalf("%s native/port constructor: %v", function, err)
	}
	public, unknownFlags = tier1PublicConstructor(operation, raw)
	return
}

func tier1CheckConstructor(t *testing.T, operation tier1ConstructorOperation, raw uint64) {
	t.Helper()
	function, operand, native, port, public, unknownFlags := tier1LegsConstructor(t, operation, raw)
	if unknownFlags != 0 || native != port || native != public {
		t.Fatalf("%s mismatch: operand=%s mode=%s C=%s port=%s public=%s unknown_public_flags=%s", function, operand, operation.mode.name, native, port, public, unknownFlags)
	}
}

func tier1CheckConstructorConvenience32(t *testing.T, value int32) {
	t.Helper()
	rounded, flags := NewDecimal32FromInt32(value, RoundNearestEven)
	got, err := NewDecimal32FromInt(value)
	if flags&FlagInexact != 0 {
		if err == nil || got != canonicalQNaN32BID() {
			t.Fatalf("NewDecimal32FromInt(%d) = %08x/%v, want canonical qNaN and error after inexact conversion", value, got.ToUint32(), err)
		}
		return
	}
	if flags != 0 || err != nil || got != rounded {
		t.Fatalf("NewDecimal32FromInt(%d) = %08x/%v, rounded=%08x flags=%s", value, got.ToUint32(), err, rounded.ToUint32(), flags)
	}
}

func tier1CheckConstructorConvenience64(t *testing.T, value int64) {
	t.Helper()
	rounded, flags := NewDecimal64FromInt64(value, RoundNearestEven)
	got, err := NewDecimal64FromInt(value)
	if flags&FlagInexact != 0 {
		if err == nil || got != canonicalQNaN64BID() {
			t.Fatalf("NewDecimal64FromInt(%d) = %016x/%v, want canonical qNaN and error after inexact conversion", value, got.ToUint64(), err)
		}
		return
	}
	if flags != 0 || err != nil || got != rounded {
		t.Fatalf("NewDecimal64FromInt(%d) = %016x/%v, rounded=%016x flags=%s", value, got.ToUint64(), err, rounded.ToUint64(), flags)
	}
}

func tier1CheckConstructorConvenience128(t *testing.T, value int64) {
	t.Helper()
	want := NewDecimal128FromInt64(value)
	got, err := NewDecimal128FromInt(value)
	if err != nil || got != want {
		t.Fatalf("NewDecimal128FromInt(%d) = %x/%v, want %x/nil", value, got, err, want)
	}
}

func TestTier1ConversionStructuredNativeDifferential(t *testing.T) {
	requireNative(t)
	if strconv.IntSize != 64 {
		t.Fatalf("Tier 1 conversion native oracle requires the guaranteed LP64 platform contract; int size=%d", strconv.IntSize)
	}
	if len(tier1ConversionSemantic32) != int(tier1ConversionSemantic32Count) ||
		len(tier1ConversionSemantic64) != int(tier1ConversionSemantic64Count) ||
		len(tier1ConversionSemantic128) != int(tier1ConversionSemantic128Count) ||
		len(tier1ConstructorInt32Inputs) != int(tier1ConstructorInt32Count) ||
		len(tier1ConstructorUint32Inputs) != int(tier1ConstructorUint32Count) ||
		len(tier1ConstructorInt64Inputs) != int(tier1ConstructorInt64Count) ||
		len(tier1ConstructorUint64Inputs) != int(tier1ConstructorUint64Count) {
		t.Fatal("generated Tier 1 conversion semantic inventory count mismatch")
	}
	toIntegerOperations := tier1ToIntegerOperations()
	if len(toIntegerOperations) != int(tier1ToIntegerOperationCount) {
		t.Fatalf("Tier 1 to-integer operation count=%d want=%d", len(toIntegerOperations), tier1ToIntegerOperationCount)
	}
	constructorOperations := tier1ConstructorOperations()
	if len(constructorOperations) != int(tier1ConstructorOperationCount) {
		t.Fatalf("Tier 1 constructor operation count=%d want=%d", len(constructorOperations), tier1ConstructorOperationCount)
	}
	if len(tier1WidthOperations(32)) != int(tier1WidthOperationsFrom32) ||
		len(tier1WidthOperations(64)) != int(tier1WidthOperationsFrom64) ||
		len(tier1WidthOperations(128)) != int(tier1WidthOperationsFrom128) {
		t.Fatal("Tier 1 width-conversion operation count mismatch")
	}
	if len(tier1BinaryOperations(32)) != int(tier1BinaryOperationsPerSource) ||
		len(tier1BinaryOperations(64)) != int(tier1BinaryOperationsPerSource) ||
		len(tier1BinaryOperations(128)) != int(tier1BinaryOperationsPerSource) {
		t.Fatal("Tier 1 binary-conversion operation count mismatch")
	}
	shard := tier1CompareConversionLoadShard(t)
	var structuredTotal uint64
	var structuredExecuted uint64
	var structuredDomains uint64

	t.Run("bid_to_integer_decimal32", func(t *testing.T) {
		var comparison uint64
		visit := func(value uint32) {
			for _, operation := range toIntegerOperations {
				if shard.owns(comparison) {
					tier1CheckToInteger32(t, value, operation)
				}
				comparison++
			}
		}
		for _, value := range tier1ArithmeticBoundary32 {
			visit(value)
		}
		for _, value := range tier1ConversionSemantic32 {
			visit(value)
		}
		if comparison != tier1ToIntegerStructured32 {
			t.Fatalf("Decimal32 structured BID-to-integer comparisons=%d want=%d", comparison, tier1ToIntegerStructured32)
		}
		structuredTotal += comparison
		structuredExecuted += shard.ownedCount(comparison)
		structuredDomains++
		t.Logf("Decimal32 structured BID-to-integer exact comparisons: %d/%d", shard.ownedCount(comparison), comparison)
	})

	t.Run("bid_to_integer_decimal64", func(t *testing.T) {
		var comparison uint64
		visit := func(value uint64) {
			for _, operation := range toIntegerOperations {
				if shard.owns(comparison) {
					tier1CheckToInteger64(t, value, operation)
				}
				comparison++
			}
		}
		for _, value := range tier1ArithmeticBoundary64 {
			visit(value)
		}
		for _, value := range tier1ConversionSemantic64 {
			visit(value)
		}
		if comparison != tier1ToIntegerStructured64 {
			t.Fatalf("Decimal64 structured BID-to-integer comparisons=%d want=%d", comparison, tier1ToIntegerStructured64)
		}
		structuredTotal += comparison
		structuredExecuted += shard.ownedCount(comparison)
		structuredDomains++
		t.Logf("Decimal64 structured BID-to-integer exact comparisons: %d/%d", shard.ownedCount(comparison), comparison)
	})

	t.Run("bid_to_integer_decimal128", func(t *testing.T) {
		var comparison uint64
		visit := func(value Decimal128BID) {
			for _, operation := range toIntegerOperations {
				if shard.owns(comparison) {
					tier1CheckToInteger128(t, value, operation)
				}
				comparison++
			}
		}
		for _, words := range tier1ArithmeticBoundary128 {
			visit(tier1ArithmeticDecimal128(words))
		}
		for _, words := range tier1ConversionSemantic128 {
			visit(tier1ArithmeticDecimal128(words))
		}
		if comparison != tier1ToIntegerStructured128 {
			t.Fatalf("Decimal128 structured BID-to-integer comparisons=%d want=%d", comparison, tier1ToIntegerStructured128)
		}
		structuredTotal += comparison
		structuredExecuted += shard.ownedCount(comparison)
		structuredDomains++
		t.Logf("Decimal128 structured BID-to-integer exact comparisons: %d/%d", shard.ownedCount(comparison), comparison)
	})

	t.Run("width_from_decimal32", func(t *testing.T) {
		operations := tier1WidthOperations(32)
		var comparison uint64
		visit := func(value uint32) {
			for _, operation := range operations {
				if shard.owns(comparison) {
					tier1CheckWidth32(t, value, operation)
				}
				comparison++
			}
		}
		for _, value := range tier1ArithmeticBoundary32 {
			visit(value)
		}
		for _, value := range tier1ConversionSemantic32 {
			visit(value)
		}
		if comparison != tier1WidthStructured32 {
			t.Fatalf("Decimal32-source structured width conversions=%d want=%d", comparison, tier1WidthStructured32)
		}
		structuredTotal += comparison
		structuredExecuted += shard.ownedCount(comparison)
		structuredDomains++
		t.Logf("Decimal32-source structured width exact comparisons: %d/%d", shard.ownedCount(comparison), comparison)
	})

	t.Run("width_from_decimal64", func(t *testing.T) {
		operations := tier1WidthOperations(64)
		var comparison uint64
		visit := func(value uint64) {
			for _, operation := range operations {
				if shard.owns(comparison) {
					tier1CheckWidth64(t, value, operation)
				}
				comparison++
			}
		}
		for _, value := range tier1ArithmeticBoundary64 {
			visit(value)
		}
		for _, value := range tier1ConversionSemantic64 {
			visit(value)
		}
		if comparison != tier1WidthStructured64 {
			t.Fatalf("Decimal64-source structured width conversions=%d want=%d", comparison, tier1WidthStructured64)
		}
		structuredTotal += comparison
		structuredExecuted += shard.ownedCount(comparison)
		structuredDomains++
		t.Logf("Decimal64-source structured width exact comparisons: %d/%d", shard.ownedCount(comparison), comparison)
	})

	t.Run("width_from_decimal128", func(t *testing.T) {
		operations := tier1WidthOperations(128)
		var comparison uint64
		visit := func(value Decimal128BID) {
			for _, operation := range operations {
				if shard.owns(comparison) {
					tier1CheckWidth128(t, value, operation)
				}
				comparison++
			}
		}
		for _, words := range tier1ArithmeticBoundary128 {
			visit(tier1ArithmeticDecimal128(words))
		}
		for _, words := range tier1ConversionSemantic128 {
			visit(tier1ArithmeticDecimal128(words))
		}
		if comparison != tier1WidthStructured128 {
			t.Fatalf("Decimal128-source structured width conversions=%d want=%d", comparison, tier1WidthStructured128)
		}
		structuredTotal += comparison
		structuredExecuted += shard.ownedCount(comparison)
		structuredDomains++
		t.Logf("Decimal128-source structured width exact comparisons: %d/%d", shard.ownedCount(comparison), comparison)
	})

	t.Run("binary_from_decimal32", func(t *testing.T) {
		operations := tier1BinaryOperations(32)
		var comparison uint64
		visit := func(value uint32) {
			for _, operation := range operations {
				if shard.owns(comparison) {
					tier1CheckBinary32(t, value, operation)
				}
				comparison++
			}
		}
		for _, value := range tier1ArithmeticBoundary32 {
			visit(value)
		}
		for _, value := range tier1ConversionSemantic32 {
			visit(value)
		}
		if comparison != tier1BinaryStructured32 {
			t.Fatalf("Decimal32-source structured binary conversions=%d want=%d", comparison, tier1BinaryStructured32)
		}
		structuredTotal += comparison
		structuredExecuted += shard.ownedCount(comparison)
		structuredDomains++
		t.Logf("Decimal32-source structured binary exact comparisons: %d/%d", shard.ownedCount(comparison), comparison)
	})

	t.Run("binary_from_decimal64", func(t *testing.T) {
		operations := tier1BinaryOperations(64)
		var comparison uint64
		visit := func(value uint64) {
			for _, operation := range operations {
				if shard.owns(comparison) {
					tier1CheckBinary64(t, value, operation)
				}
				comparison++
			}
		}
		for _, value := range tier1ArithmeticBoundary64 {
			visit(value)
		}
		for _, value := range tier1ConversionSemantic64 {
			visit(value)
		}
		if comparison != tier1BinaryStructured64 {
			t.Fatalf("Decimal64-source structured binary conversions=%d want=%d", comparison, tier1BinaryStructured64)
		}
		structuredTotal += comparison
		structuredExecuted += shard.ownedCount(comparison)
		structuredDomains++
		t.Logf("Decimal64-source structured binary exact comparisons: %d/%d", shard.ownedCount(comparison), comparison)
	})

	t.Run("binary_from_decimal128", func(t *testing.T) {
		operations := tier1BinaryOperations(128)
		var comparison uint64
		visit := func(value Decimal128BID) {
			for _, operation := range operations {
				if shard.owns(comparison) {
					tier1CheckBinary128(t, value, operation)
				}
				comparison++
			}
		}
		for _, words := range tier1ArithmeticBoundary128 {
			visit(tier1ArithmeticDecimal128(words))
		}
		for _, words := range tier1ConversionSemantic128 {
			visit(tier1ArithmeticDecimal128(words))
		}
		if comparison != tier1BinaryStructured128 {
			t.Fatalf("Decimal128-source structured binary conversions=%d want=%d", comparison, tier1BinaryStructured128)
		}
		structuredTotal += comparison
		structuredExecuted += shard.ownedCount(comparison)
		structuredDomains++
		t.Logf("Decimal128-source structured binary exact comparisons: %d/%d", shard.ownedCount(comparison), comparison)
	})

	t.Run("integer_constructors", func(t *testing.T) {
		var comparison uint64
		for _, operation := range constructorOperations {
			switch operation.kind {
			case "int32":
				for _, value := range tier1ConstructorInt32Inputs {
					if shard.owns(comparison) {
						tier1CheckConstructor(t, operation, uint64(uint32(value)))
					}
					comparison++
				}
			case "uint32":
				for _, value := range tier1ConstructorUint32Inputs {
					if shard.owns(comparison) {
						tier1CheckConstructor(t, operation, uint64(value))
					}
					comparison++
				}
			case "int64":
				for _, value := range tier1ConstructorInt64Inputs {
					if shard.owns(comparison) {
						tier1CheckConstructor(t, operation, uint64(value))
					}
					comparison++
				}
			case "uint64":
				for _, value := range tier1ConstructorUint64Inputs {
					if shard.owns(comparison) {
						tier1CheckConstructor(t, operation, value)
					}
					comparison++
				}
			default:
				t.Fatalf("unknown Tier 1 constructor kind %q", operation.kind)
			}
		}
		if comparison != tier1ConstructorStructured {
			t.Fatalf("structured integer constructors=%d want=%d", comparison, tier1ConstructorStructured)
		}
		structuredTotal += comparison
		structuredExecuted += shard.ownedCount(comparison)
		structuredDomains++
		t.Logf("structured integer-constructor exact comparisons: %d/%d", shard.ownedCount(comparison), comparison)
	})

	t.Run("exact_convenience_contract", func(t *testing.T) {
		for _, value := range tier1ConstructorInt32Inputs {
			tier1CheckConstructorConvenience32(t, value)
		}
		for _, value := range tier1ConstructorInt64Inputs {
			tier1CheckConstructorConvenience64(t, value)
			tier1CheckConstructorConvenience128(t, value)
		}
		executed := uint64(len(tier1ConstructorInt32Inputs) + 2*len(tier1ConstructorInt64Inputs))
		if executed != tier1ConstructorConvenienceChecks {
			t.Fatalf("constructor convenience checks=%d want=%d", executed, tier1ConstructorConvenienceChecks)
		}
		t.Logf("exact integer-constructor convenience checks: %d", executed)
	})

	if structuredDomains == 10 {
		if structuredTotal != tier1ConversionStructured {
			t.Fatalf("Tier 1 structured conversion comparisons=%d want=%d", structuredTotal, tier1ConversionStructured)
		}
		t.Logf("Tier 1 structured conversion exact comparisons: %d/%d", structuredExecuted, structuredTotal)
	} else {
		t.Logf("Tier 1 selected structured conversion exact comparisons: %d", structuredExecuted)
	}
}

func TestTier1ConversionDeterministicRandomNativeDifferential(t *testing.T) {
	requireNative(t)
	toIntegerOperations := tier1ToIntegerOperations()
	constructorOperations := tier1ConstructorOperations()
	shard := tier1CompareConversionLoadShard(t)
	var randomTotal uint64
	var randomExecuted uint64
	var randomDomains uint64

	t.Run("bid_to_integer_decimal32", func(t *testing.T) {
		var comparison uint64
		var randomConsumed uint64
		randomDigest := tier1ArithmeticScaleTupleHashMix(tier1ArithmeticScaleTupleHashOffset, 32)
		for i := uint64(0); i < tier1ToIntegerRandom32; i++ {
			value := tier1ArithmeticRandomOperand32(tier1ToIntegerRandomSeed32, i, 0)
			randomDigest = tier1ArithmeticMixOperand32(randomDigest, value)
			randomConsumed++
			if shard.owns(comparison) {
				tier1CheckToInteger32(t, value, toIntegerOperations[i%uint64(len(toIntegerOperations))])
			}
			comparison++
		}
		if got := tier1ArithmeticScaleTupleHashMix(randomDigest, randomConsumed); got != tier1ToIntegerRandomStreamHash32 {
			t.Fatalf("Decimal32 random BID-to-integer stream hash=%d want=%d (consumed operands diverge from the anchored stream)", got, tier1ToIntegerRandomStreamHash32)
		}
		randomTotal += comparison
		randomExecuted += shard.ownedCount(comparison)
		randomDomains++
		t.Logf("Decimal32 random BID-to-integer exact comparisons: %d/%d", shard.ownedCount(comparison), comparison)
	})
	t.Run("bid_to_integer_decimal64", func(t *testing.T) {
		var comparison uint64
		var randomConsumed uint64
		randomDigest := tier1ArithmeticScaleTupleHashMix(tier1ArithmeticScaleTupleHashOffset, 64)
		for i := uint64(0); i < tier1ToIntegerRandom64; i++ {
			value := tier1ArithmeticRandomOperand64(tier1ToIntegerRandomSeed64, i, 0)
			randomDigest = tier1ArithmeticMixOperand64(randomDigest, value)
			randomConsumed++
			if shard.owns(comparison) {
				tier1CheckToInteger64(t, value, toIntegerOperations[i%uint64(len(toIntegerOperations))])
			}
			comparison++
		}
		if got := tier1ArithmeticScaleTupleHashMix(randomDigest, randomConsumed); got != tier1ToIntegerRandomStreamHash64 {
			t.Fatalf("Decimal64 random BID-to-integer stream hash=%d want=%d (consumed operands diverge from the anchored stream)", got, tier1ToIntegerRandomStreamHash64)
		}
		randomTotal += comparison
		randomExecuted += shard.ownedCount(comparison)
		randomDomains++
		t.Logf("Decimal64 random BID-to-integer exact comparisons: %d/%d", shard.ownedCount(comparison), comparison)
	})
	t.Run("bid_to_integer_decimal128", func(t *testing.T) {
		var comparison uint64
		var randomConsumed uint64
		randomDigest := tier1ArithmeticScaleTupleHashMix(tier1ArithmeticScaleTupleHashOffset, 128)
		for i := uint64(0); i < tier1ToIntegerRandom128; i++ {
			words := tier1ArithmeticRandomOperand128(tier1ToIntegerRandomSeed128, i, 0)
			randomDigest = tier1ArithmeticMixOperand128(randomDigest, words)
			randomConsumed++
			if shard.owns(comparison) {
				tier1CheckToInteger128(t, tier1ArithmeticDecimal128(words), toIntegerOperations[i%uint64(len(toIntegerOperations))])
			}
			comparison++
		}
		if got := tier1ArithmeticScaleTupleHashMix(randomDigest, randomConsumed); got != tier1ToIntegerRandomStreamHash128 {
			t.Fatalf("Decimal128 random BID-to-integer stream hash=%d want=%d (consumed operands diverge from the anchored stream)", got, tier1ToIntegerRandomStreamHash128)
		}
		randomTotal += comparison
		randomExecuted += shard.ownedCount(comparison)
		randomDomains++
		t.Logf("Decimal128 random BID-to-integer exact comparisons: %d/%d", shard.ownedCount(comparison), comparison)
	})

	t.Run("width_from_decimal32", func(t *testing.T) {
		operations := tier1WidthOperations(32)
		var comparison uint64
		var randomConsumed uint64
		randomDigest := tier1ArithmeticScaleTupleHashMix(tier1ArithmeticScaleTupleHashOffset, 32)
		for i := uint64(0); i < tier1WidthRandom32; i++ {
			value := tier1ArithmeticRandomOperand32(tier1WidthRandomSeed32, i, 0)
			randomDigest = tier1ArithmeticMixOperand32(randomDigest, value)
			randomConsumed++
			if shard.owns(comparison) {
				tier1CheckWidth32(t, value, operations[i%uint64(len(operations))])
			}
			comparison++
		}
		if got := tier1ArithmeticScaleTupleHashMix(randomDigest, randomConsumed); got != tier1WidthRandomStreamHash32 {
			t.Fatalf("Decimal32-source random width stream hash=%d want=%d (consumed operands diverge from the anchored stream)", got, tier1WidthRandomStreamHash32)
		}
		randomTotal += comparison
		randomExecuted += shard.ownedCount(comparison)
		randomDomains++
		t.Logf("Decimal32-source random width exact comparisons: %d/%d", shard.ownedCount(comparison), comparison)
	})
	t.Run("width_from_decimal64", func(t *testing.T) {
		operations := tier1WidthOperations(64)
		var comparison uint64
		var randomConsumed uint64
		randomDigest := tier1ArithmeticScaleTupleHashMix(tier1ArithmeticScaleTupleHashOffset, 64)
		for i := uint64(0); i < tier1WidthRandom64; i++ {
			value := tier1ArithmeticRandomOperand64(tier1WidthRandomSeed64, i, 0)
			randomDigest = tier1ArithmeticMixOperand64(randomDigest, value)
			randomConsumed++
			if shard.owns(comparison) {
				tier1CheckWidth64(t, value, operations[i%uint64(len(operations))])
			}
			comparison++
		}
		if got := tier1ArithmeticScaleTupleHashMix(randomDigest, randomConsumed); got != tier1WidthRandomStreamHash64 {
			t.Fatalf("Decimal64-source random width stream hash=%d want=%d (consumed operands diverge from the anchored stream)", got, tier1WidthRandomStreamHash64)
		}
		randomTotal += comparison
		randomExecuted += shard.ownedCount(comparison)
		randomDomains++
		t.Logf("Decimal64-source random width exact comparisons: %d/%d", shard.ownedCount(comparison), comparison)
	})
	t.Run("width_from_decimal128", func(t *testing.T) {
		operations := tier1WidthOperations(128)
		var comparison uint64
		var randomConsumed uint64
		randomDigest := tier1ArithmeticScaleTupleHashMix(tier1ArithmeticScaleTupleHashOffset, 128)
		for i := uint64(0); i < tier1WidthRandom128; i++ {
			words := tier1ArithmeticRandomOperand128(tier1WidthRandomSeed128, i, 0)
			randomDigest = tier1ArithmeticMixOperand128(randomDigest, words)
			randomConsumed++
			if shard.owns(comparison) {
				tier1CheckWidth128(t, tier1ArithmeticDecimal128(words), operations[i%uint64(len(operations))])
			}
			comparison++
		}
		if got := tier1ArithmeticScaleTupleHashMix(randomDigest, randomConsumed); got != tier1WidthRandomStreamHash128 {
			t.Fatalf("Decimal128-source random width stream hash=%d want=%d (consumed operands diverge from the anchored stream)", got, tier1WidthRandomStreamHash128)
		}
		randomTotal += comparison
		randomExecuted += shard.ownedCount(comparison)
		randomDomains++
		t.Logf("Decimal128-source random width exact comparisons: %d/%d", shard.ownedCount(comparison), comparison)
	})

	t.Run("binary_from_decimal32", func(t *testing.T) {
		operations := tier1BinaryOperations(32)
		var comparison uint64
		var randomConsumed uint64
		randomDigest := tier1ArithmeticScaleTupleHashMix(tier1ArithmeticScaleTupleHashOffset, 32)
		for i := uint64(0); i < tier1BinaryRandom32; i++ {
			value := tier1ArithmeticRandomOperand32(tier1BinaryRandomSeed32, i, 0)
			randomDigest = tier1ArithmeticMixOperand32(randomDigest, value)
			randomConsumed++
			if shard.owns(comparison) {
				tier1CheckBinary32(t, value, operations[i%uint64(len(operations))])
			}
			comparison++
		}
		if got := tier1ArithmeticScaleTupleHashMix(randomDigest, randomConsumed); got != tier1BinaryRandomStreamHash32 {
			t.Fatalf("Decimal32-source random binary stream hash=%d want=%d (consumed operands diverge from the anchored stream)", got, tier1BinaryRandomStreamHash32)
		}
		randomTotal += comparison
		randomExecuted += shard.ownedCount(comparison)
		randomDomains++
		t.Logf("Decimal32-source random binary exact comparisons: %d/%d", shard.ownedCount(comparison), comparison)
	})
	t.Run("binary_from_decimal64", func(t *testing.T) {
		operations := tier1BinaryOperations(64)
		var comparison uint64
		var randomConsumed uint64
		randomDigest := tier1ArithmeticScaleTupleHashMix(tier1ArithmeticScaleTupleHashOffset, 64)
		for i := uint64(0); i < tier1BinaryRandom64; i++ {
			value := tier1ArithmeticRandomOperand64(tier1BinaryRandomSeed64, i, 0)
			randomDigest = tier1ArithmeticMixOperand64(randomDigest, value)
			randomConsumed++
			if shard.owns(comparison) {
				tier1CheckBinary64(t, value, operations[i%uint64(len(operations))])
			}
			comparison++
		}
		if got := tier1ArithmeticScaleTupleHashMix(randomDigest, randomConsumed); got != tier1BinaryRandomStreamHash64 {
			t.Fatalf("Decimal64-source random binary stream hash=%d want=%d (consumed operands diverge from the anchored stream)", got, tier1BinaryRandomStreamHash64)
		}
		randomTotal += comparison
		randomExecuted += shard.ownedCount(comparison)
		randomDomains++
		t.Logf("Decimal64-source random binary exact comparisons: %d/%d", shard.ownedCount(comparison), comparison)
	})
	t.Run("binary_from_decimal128", func(t *testing.T) {
		operations := tier1BinaryOperations(128)
		var comparison uint64
		var randomConsumed uint64
		randomDigest := tier1ArithmeticScaleTupleHashMix(tier1ArithmeticScaleTupleHashOffset, 128)
		for i := uint64(0); i < tier1BinaryRandom128; i++ {
			words := tier1ArithmeticRandomOperand128(tier1BinaryRandomSeed128, i, 0)
			randomDigest = tier1ArithmeticMixOperand128(randomDigest, words)
			randomConsumed++
			if shard.owns(comparison) {
				tier1CheckBinary128(t, tier1ArithmeticDecimal128(words), operations[i%uint64(len(operations))])
			}
			comparison++
		}
		if got := tier1ArithmeticScaleTupleHashMix(randomDigest, randomConsumed); got != tier1BinaryRandomStreamHash128 {
			t.Fatalf("Decimal128-source random binary stream hash=%d want=%d (consumed operands diverge from the anchored stream)", got, tier1BinaryRandomStreamHash128)
		}
		randomTotal += comparison
		randomExecuted += shard.ownedCount(comparison)
		randomDomains++
		t.Logf("Decimal128-source random binary exact comparisons: %d/%d", shard.ownedCount(comparison), comparison)
	})

	t.Run("integer_constructors", func(t *testing.T) {
		var comparison uint64
		var randomConsumed uint64
		randomDigest := tier1ArithmeticScaleTupleHashMix(tier1ArithmeticScaleTupleHashOffset, 0)
		for i := uint64(0); i < tier1ConstructorRandom; i++ {
			raw := tier1ArithmeticRandomOperand64(tier1ConstructorRandomSeed, i, 0)
			randomDigest = tier1ArithmeticMixOperand64(randomDigest, raw)
			randomConsumed++
			if shard.owns(comparison) {
				tier1CheckConstructor(t, constructorOperations[i%uint64(len(constructorOperations))], raw)
			}
			comparison++
		}
		if got := tier1ArithmeticScaleTupleHashMix(randomDigest, randomConsumed); got != tier1ConstructorRandomStreamHash {
			t.Fatalf("random integer-constructor stream hash=%d want=%d (consumed operands diverge from the anchored stream)", got, tier1ConstructorRandomStreamHash)
		}
		randomTotal += comparison
		randomExecuted += shard.ownedCount(comparison)
		randomDomains++
		t.Logf("random integer-constructor exact comparisons: %d/%d", shard.ownedCount(comparison), comparison)
	})

	if randomDomains == 10 {
		if randomTotal != tier1ConversionRandom {
			t.Fatalf("Tier 1 random conversion comparisons=%d want=%d", randomTotal, tier1ConversionRandom)
		}
		if tier1ConversionStructured+randomTotal != tier1ConversionTotal {
			t.Fatal("Tier 1 conversion total comparison constant is inconsistent")
		}
		t.Logf("Tier 1 deterministic random conversion exact comparisons: %d/%d", randomExecuted, randomTotal)
	} else {
		t.Logf("Tier 1 selected deterministic random conversion exact comparisons: %d", randomExecuted)
	}
}

// Routing sentinels: generator-selected known-answer rows that bind the
// compare/conversion runner glue (operand slots, dispatch-row labels
// including their embedded rounding modes) to values pinned outside the
// runtime; see the arithmetic runner's sentinel block for the axis-reading
// rules. Every dispatch-table row of every family carries at least one row.

// tier1CCSentinelCanonToInt rewrites one to-int leg from the runner-internal
// "<decimal>/<2-hex status>" form into the cross-language canonical
// "<64-bit two's-complement hex>/<8-hex flags>" form.
func tier1CCSentinelCanonToInt(t *testing.T, row, leg, text string, signed bool) string {
	t.Helper()
	slash := strings.IndexByte(text, '/')
	if slash < 0 {
		t.Fatalf("routing sentinel row [%s]: unexpected %s integer result %q", row, leg, text)
	}
	var register uint64
	if signed {
		value, err := strconv.ParseInt(text[:slash], 10, 64)
		if err != nil {
			t.Fatalf("routing sentinel row [%s]: undecodable %s signed integer %q: %v", row, leg, text, err)
		}
		register = uint64(value)
	} else {
		value, err := strconv.ParseUint(text[:slash], 10, 64)
		if err != nil {
			t.Fatalf("routing sentinel row [%s]: undecodable %s unsigned integer %q: %v", row, leg, text, err)
		}
		register = value
	}
	status, err := strconv.ParseUint(text[slash+1:], 16, 32)
	if err != nil {
		t.Fatalf("routing sentinel row [%s]: undecodable %s status %q: %v", row, leg, text, err)
	}
	return fmt.Sprintf("%016x/%08x", register, uint32(status))
}

// tier1CCSentinelCanon128NoFlags rewrites a flagless 128-bit leg
// ("<32 le-hex>") into the canonical "<hi16>:<lo16>" form.
func tier1CCSentinelCanon128NoFlags(t *testing.T, row, leg, ffi string) string {
	t.Helper()
	canon := tier1ArithmeticSentinelCanon128(t, row, leg, ffi+"/00000000")
	return strings.TrimSuffix(canon, "/00000000")
}

func tier1CCSentinelBool(value int, flags uint32) string {
	return fmt.Sprintf("%02d/%08x", value, flags)
}

func tier1CCSentinelQuiet(t *testing.T, row, widthLabel, operation, xText, yText, pinned string) {
	t.Helper()
	var native, port, public int
	var nativeFlags, portFlags, publicFlags uint32
	var unknownFlags ExceptionFlags
	switch widthLabel {
	case "d32":
		native, nativeFlags, port, portFlags, public, publicFlags, unknownFlags = tier1LegsQuiet32(t, operation,
			tier1ArithmeticSentinelHex32(t, row, xText), tier1ArithmeticSentinelHex32(t, row, yText))
	case "d64":
		native, nativeFlags, port, portFlags, public, publicFlags, unknownFlags = tier1LegsQuiet64(t, operation,
			tier1ArithmeticSentinelHex64(t, row, xText), tier1ArithmeticSentinelHex64(t, row, yText))
	case "d128":
		native, nativeFlags, port, portFlags, public, publicFlags, unknownFlags = tier1LegsQuiet128(t, operation,
			tier1ArithmeticDecimal128(tier1ArithmeticSentinelWords128(t, row, xText)),
			tier1ArithmeticDecimal128(tier1ArithmeticSentinelWords128(t, row, yText)))
	default:
		t.Fatalf("routing sentinel row [%s]: unknown width %q", row, widthLabel)
	}
	tier1ArithmeticSentinelUnknownFlags(t, row, unknownFlags)
	tier1ArithmeticSentinelAssert(t, row, "(none)", pinned,
		tier1CCSentinelBool(native, nativeFlags), tier1CCSentinelBool(port, portFlags), tier1CCSentinelBool(public, publicFlags))
}

func tier1CCSentinelMinMax(t *testing.T, row, widthLabel, operation, xText, yText, pinned string) {
	t.Helper()
	var native, port, public string
	var unknownFlags ExceptionFlags
	switch widthLabel {
	case "d32":
		native, port, public, _, unknownFlags = tier1LegsMinMax32(t, operation,
			tier1ArithmeticSentinelHex32(t, row, xText), tier1ArithmeticSentinelHex32(t, row, yText))
	case "d64":
		native, port, public, unknownFlags = tier1LegsMinMax64(t, operation,
			tier1ArithmeticSentinelHex64(t, row, xText), tier1ArithmeticSentinelHex64(t, row, yText))
	case "d128":
		nativeFFI, portFFI, publicFFI, unknown := tier1LegsMinMax128(t, operation,
			tier1ArithmeticDecimal128(tier1ArithmeticSentinelWords128(t, row, xText)),
			tier1ArithmeticDecimal128(tier1ArithmeticSentinelWords128(t, row, yText)))
		native = tier1ArithmeticSentinelCanon128(t, row, "C", nativeFFI)
		port = tier1ArithmeticSentinelCanon128(t, row, "port", portFFI)
		public = tier1ArithmeticSentinelCanon128(t, row, "public", publicFFI)
		unknownFlags = unknown
	default:
		t.Fatalf("routing sentinel row [%s]: unknown width %q", row, widthLabel)
	}
	tier1ArithmeticSentinelUnknownFlags(t, row, unknownFlags)
	tier1ArithmeticSentinelAssert(t, row, "(none)", pinned, native, port, public)
}

func tier1CCSentinelToIntOperation(t *testing.T, row, operation string) tier1ToIntegerOperation {
	t.Helper()
	for _, candidate := range tier1ToIntegerOperations() {
		if "to_"+candidate.kind+"_"+candidate.suffix == operation {
			return candidate
		}
	}
	t.Fatalf("routing sentinel row [%s]: operation %q is not in the runner to-integer table", row, operation)
	return tier1ToIntegerOperation{}
}

func tier1CCSentinelToInt(t *testing.T, row, widthLabel, operation, xText, pinned string) {
	t.Helper()
	op := tier1CCSentinelToIntOperation(t, row, operation)
	signed := strings.HasPrefix(op.kind, "int")
	var native, port, public string
	var unknownFlags ExceptionFlags
	switch widthLabel {
	case "d32":
		native, port, public, unknownFlags = tier1LegsToInteger32(t, tier1ArithmeticSentinelHex32(t, row, xText), op)
	case "d64":
		native, port, public, unknownFlags = tier1LegsToInteger64(t, tier1ArithmeticSentinelHex64(t, row, xText), op)
	case "d128":
		native, port, public, unknownFlags = tier1LegsToInteger128(t, tier1ArithmeticDecimal128(tier1ArithmeticSentinelWords128(t, row, xText)), op)
	default:
		t.Fatalf("routing sentinel row [%s]: unknown width %q", row, widthLabel)
	}
	tier1ArithmeticSentinelUnknownFlags(t, row, unknownFlags)
	tier1ArithmeticSentinelAssert(t, row, op.mode.name+"(suffix)", pinned,
		tier1CCSentinelCanonToInt(t, row, "C", native, signed),
		tier1CCSentinelCanonToInt(t, row, "port", port, signed),
		tier1CCSentinelCanonToInt(t, row, "public", public, signed))
}

func tier1CCSentinelWidthOperation(t *testing.T, row string, source int, operation string, haveMode bool, mode tier1ArithmeticMode) tier1WidthOperation {
	t.Helper()
	for _, candidate := range tier1WidthOperations(source) {
		if fmt.Sprintf("to_bid%d", candidate.dest) != operation {
			continue
		}
		if candidate.rounded != haveMode {
			continue
		}
		if candidate.rounded && candidate.mode.native != mode.native {
			continue
		}
		return candidate
	}
	t.Fatalf("routing sentinel row [%s]: operation %q (mode present %v) is not in the runner width table", row, operation, haveMode)
	return tier1WidthOperation{}
}

func tier1CCSentinelWidth(t *testing.T, row, widthLabel, operation, xText string, haveMode bool, mode tier1ArithmeticMode, pinned string) {
	t.Helper()
	var source int
	var operand string
	var public string
	var unknownFlags ExceptionFlags
	var op tier1WidthOperation
	switch widthLabel {
	case "d32":
		source = 32
		value := tier1ArithmeticSentinelHex32(t, row, xText)
		op = tier1CCSentinelWidthOperation(t, row, source, operation, haveMode, mode)
		operand = fmt.Sprintf("%08x", value)
		public, unknownFlags = tier1PublicWidth32(t, value, op)
	case "d64":
		source = 64
		value := tier1ArithmeticSentinelHex64(t, row, xText)
		op = tier1CCSentinelWidthOperation(t, row, source, operation, haveMode, mode)
		operand = fmt.Sprintf("%016x", value)
		public, unknownFlags = tier1PublicWidth64(t, value, op)
	case "d128":
		source = 128
		value := tier1ArithmeticDecimal128(tier1ArithmeticSentinelWords128(t, row, xText))
		op = tier1CCSentinelWidthOperation(t, row, source, operation, haveMode, mode)
		operand = formatFFIUint128Bits(value)
		public, unknownFlags = tier1PublicWidth128(t, value, op)
	default:
		t.Fatalf("routing sentinel row [%s]: unknown width %q", row, widthLabel)
	}
	tier1ArithmeticSentinelUnknownFlags(t, row, unknownFlags)
	_, native, port := tier1LegsWidthConversion(t, op, operand)
	if op.dest == 128 {
		native = tier1ArithmeticSentinelCanon128(t, row, "C", native)
		port = tier1ArithmeticSentinelCanon128(t, row, "port", port)
		public = tier1ArithmeticSentinelCanon128(t, row, "public", public)
	}
	modeLabel := "(none)"
	if haveMode {
		modeLabel = tier1ArithmeticSentinelModeLabel(mode)
	}
	tier1ArithmeticSentinelAssert(t, row, modeLabel, pinned, native, port, public)
}

func tier1CCSentinelBinaryOperation(t *testing.T, row string, source int, operation string, mode tier1ArithmeticMode) tier1BinaryOperation {
	t.Helper()
	for _, candidate := range tier1BinaryOperations(source) {
		if fmt.Sprintf("to_binary%d", candidate.dest) == operation && candidate.mode.native == mode.native {
			return candidate
		}
	}
	t.Fatalf("routing sentinel row [%s]: operation %q is not in the runner binary-conversion table", row, operation)
	return tier1BinaryOperation{}
}

func tier1CCSentinelBinary(t *testing.T, row, widthLabel, operation, xText string, mode tier1ArithmeticMode, pinned string) {
	t.Helper()
	var operand string
	var public string
	var unknownFlags ExceptionFlags
	var op tier1BinaryOperation
	switch widthLabel {
	case "d32":
		value := tier1ArithmeticSentinelHex32(t, row, xText)
		op = tier1CCSentinelBinaryOperation(t, row, 32, operation, mode)
		operand = fmt.Sprintf("%08x", value)
		public, unknownFlags = tier1PublicBinary32(Decimal32BID(value), op)
	case "d64":
		value := tier1ArithmeticSentinelHex64(t, row, xText)
		op = tier1CCSentinelBinaryOperation(t, row, 64, operation, mode)
		operand = fmt.Sprintf("%016x", value)
		public, unknownFlags = tier1PublicBinary64(Decimal64BID(value), op)
	case "d128":
		value := tier1ArithmeticDecimal128(tier1ArithmeticSentinelWords128(t, row, xText))
		op = tier1CCSentinelBinaryOperation(t, row, 128, operation, mode)
		operand = formatFFIUint128Bits(value)
		public, unknownFlags = tier1PublicBinary128(value, op)
	default:
		t.Fatalf("routing sentinel row [%s]: unknown width %q", row, widthLabel)
	}
	tier1ArithmeticSentinelUnknownFlags(t, row, unknownFlags)
	_, native, port := tier1LegsBinaryConversion(t, op, operand)
	if op.dest == 128 {
		native = tier1ArithmeticSentinelCanon128(t, row, "C", native)
		port = tier1ArithmeticSentinelCanon128(t, row, "port", port)
		public = tier1ArithmeticSentinelCanon128(t, row, "public", public)
	}
	tier1ArithmeticSentinelAssert(t, row, tier1ArithmeticSentinelModeLabel(mode), pinned, native, port, public)
}

func tier1CCSentinelConstructorOperation(t *testing.T, row string, dest int, operation string, haveMode bool, mode tier1ArithmeticMode) tier1ConstructorOperation {
	t.Helper()
	for _, candidate := range tier1ConstructorOperations() {
		if candidate.dest != dest || "from_"+candidate.kind != operation {
			continue
		}
		if candidate.rounded != haveMode {
			continue
		}
		if candidate.rounded && candidate.mode.native != mode.native {
			continue
		}
		return candidate
	}
	t.Fatalf("routing sentinel row [%s]: operation %q (mode present %v) is not in the runner constructor table", row, operation, haveMode)
	return tier1ConstructorOperation{}
}

// tier1CCSentinelConstructorRegister rebuilds the 64-bit register image from
// the row's kind-typed decimal input, mirroring the structured differential's
// convention (32-bit kinds zero-extend their low word).
func tier1CCSentinelConstructorRegister(t *testing.T, row, kind, text string) uint64 {
	t.Helper()
	switch kind {
	case "int32":
		value, err := strconv.ParseInt(text, 10, 32)
		if err != nil {
			t.Fatalf("routing sentinel row [%s]: bad int32 input %q: %v", row, text, err)
		}
		return uint64(uint32(int32(value)))
	case "uint32":
		value, err := strconv.ParseUint(text, 10, 32)
		if err != nil {
			t.Fatalf("routing sentinel row [%s]: bad uint32 input %q: %v", row, text, err)
		}
		return uint64(uint32(value))
	case "int64":
		value, err := strconv.ParseInt(text, 10, 64)
		if err != nil {
			t.Fatalf("routing sentinel row [%s]: bad int64 input %q: %v", row, text, err)
		}
		return uint64(value)
	case "uint64":
		value, err := strconv.ParseUint(text, 10, 64)
		if err != nil {
			t.Fatalf("routing sentinel row [%s]: bad uint64 input %q: %v", row, text, err)
		}
		return value
	default:
		t.Fatalf("routing sentinel row [%s]: unknown constructor kind %q", row, kind)
		return 0
	}
}

func tier1CCSentinelConstructor(t *testing.T, row, widthLabel, operation, iText string, haveMode bool, mode tier1ArithmeticMode, pinned string) {
	t.Helper()
	var dest int
	switch widthLabel {
	case "d32":
		dest = 32
	case "d64":
		dest = 64
	case "d128":
		dest = 128
	default:
		t.Fatalf("routing sentinel row [%s]: unknown width %q", row, widthLabel)
	}
	op := tier1CCSentinelConstructorOperation(t, row, dest, operation, haveMode, mode)
	raw := tier1CCSentinelConstructorRegister(t, row, op.kind, iText)
	_, _, native, port, public, unknownFlags := tier1LegsConstructor(t, op, raw)
	tier1ArithmeticSentinelUnknownFlags(t, row, unknownFlags)
	if dest == 128 {
		native = tier1CCSentinelCanon128NoFlags(t, row, "C", native)
		port = tier1CCSentinelCanon128NoFlags(t, row, "port", port)
		public = tier1CCSentinelCanon128NoFlags(t, row, "public", public)
	}
	modeLabel := "(none)"
	if haveMode {
		modeLabel = tier1ArithmeticSentinelModeLabel(mode)
	}
	tier1ArithmeticSentinelAssert(t, row, modeLabel, pinned, native, port, public)
}

func tier1CompareConversionCheckRoutingSentinelRow(t *testing.T, row string) {
	t.Helper()
	fields := strings.Split(row, " ")
	if len(fields) < 4 || fields[len(fields)-2] != "->" {
		t.Fatalf("routing sentinel row [%s]: malformed layout", row)
	}
	widthLabel, operation := fields[0], fields[1]
	pinned := fields[len(fields)-1]
	var haveX, haveY, haveI, haveMode bool
	var xText, yText, iText string
	var mode tier1ArithmeticMode
	for _, field := range fields[2 : len(fields)-2] {
		switch {
		case strings.HasPrefix(field, "x="):
			xText, haveX = field[2:], true
		case strings.HasPrefix(field, "y="):
			yText, haveY = field[2:], true
		case strings.HasPrefix(field, "i="):
			iText, haveI = field[2:], true
		case strings.HasPrefix(field, "m="):
			native, err := strconv.Atoi(field[2:])
			if err != nil {
				t.Fatalf("routing sentinel row [%s]: bad native mode %q: %v", row, field, err)
			}
			mode, haveMode = tier1ArithmeticSentinelModeByNative(t, row, native), true
		default:
			t.Fatalf("routing sentinel row [%s]: unknown field %q", row, field)
		}
	}
	requireShape := func(needX, needY, needI bool) {
		t.Helper()
		if haveX != needX || haveY != needY || haveI != needI {
			t.Fatalf("routing sentinel row [%s]: field shape does not match operation %q", row, operation)
		}
	}
	switch {
	case strings.HasPrefix(operation, "quiet_"):
		requireShape(true, true, false)
		if haveMode {
			t.Fatalf("routing sentinel row [%s]: quiet predicates carry no mode", row)
		}
		tier1CCSentinelQuiet(t, row, widthLabel, operation, xText, yText, pinned)
	case operation == "minnum" || operation == "maxnum" || operation == "minnum_mag" || operation == "maxnum_mag":
		requireShape(true, true, false)
		if haveMode {
			t.Fatalf("routing sentinel row [%s]: min/max operations carry no mode", row)
		}
		tier1CCSentinelMinMax(t, row, widthLabel, operation, xText, yText, pinned)
	case strings.HasPrefix(operation, "to_int") || strings.HasPrefix(operation, "to_uint"):
		requireShape(true, false, false)
		if haveMode {
			t.Fatalf("routing sentinel row [%s]: to-integer operations embed the mode in the operation name", row)
		}
		tier1CCSentinelToInt(t, row, widthLabel, operation, xText, pinned)
	case strings.HasPrefix(operation, "to_bid"):
		requireShape(true, false, false)
		tier1CCSentinelWidth(t, row, widthLabel, operation, xText, haveMode, mode, pinned)
	case strings.HasPrefix(operation, "to_binary"):
		requireShape(true, false, false)
		if !haveMode {
			t.Fatalf("routing sentinel row [%s]: binary conversions require a mode", row)
		}
		tier1CCSentinelBinary(t, row, widthLabel, operation, xText, mode, pinned)
	case strings.HasPrefix(operation, "from_"):
		requireShape(false, false, true)
		tier1CCSentinelConstructor(t, row, widthLabel, operation, iText, haveMode, mode, pinned)
	default:
		t.Fatalf("routing sentinel row [%s]: unknown operation %q", row, operation)
	}
}

// TestTier1CompareConversionRoutingSentinels runs every pinned sentinel row
// on every leg. It deliberately ignores the shard environment: the rows are
// few and every shard configuration (and the -full gate) must execute all of
// them.
func TestTier1CompareConversionRoutingSentinels(t *testing.T) {
	requireNative(t)
	if len(tier1CompareConversionRoutingSentinelRows) == 0 {
		t.Fatal("generated Tier 1 compare/conversion routing sentinel row set is empty")
	}
	if uint64(len(tier1CompareConversionRoutingSentinelRows)) != tier1CompareConversionRoutingSentinelCount {
		t.Fatalf("generated Tier 1 compare/conversion routing sentinel rows=%d want=%d", len(tier1CompareConversionRoutingSentinelRows), tier1CompareConversionRoutingSentinelCount)
	}
	for _, row := range tier1CompareConversionRoutingSentinelRows {
		tier1CompareConversionCheckRoutingSentinelRow(t, row)
	}
	t.Logf("Tier 1 compare/conversion routing sentinels: %d/%d", len(tier1CompareConversionRoutingSentinelRows), len(tier1CompareConversionRoutingSentinelRows))
}

// tier1CompareConversionRoutingSentinelRows is the canonical sentinel row
// set. The identical byte sequence is pinned by hand in
// devtools/verification_sentinels.json and emitted into the generated Rust
// runner; TestVerificationAnchorsMatchGeneratedArtifacts requires the three
// copies to match exactly.
var tier1CompareConversionRoutingSentinelRows = []string{
@@TIER1_CC_SENTINEL_ROWS@@
}

var tier1ConversionSemantic32 = []uint32{
@@TIER1_CONVERSION_SEMANTIC32_VALUES@@
}

var tier1ConversionSemantic64 = []uint64{
@@TIER1_CONVERSION_SEMANTIC64_VALUES@@
}

var tier1ConversionSemantic128 = []tier1Arithmetic128Words{
@@TIER1_CONVERSION_SEMANTIC128_VALUES@@
}

var tier1ConstructorInt32Inputs = []int32{
@@TIER1_CONSTRUCTOR_INT32_VALUES@@
}

var tier1ConstructorUint32Inputs = []uint32{
@@TIER1_CONSTRUCTOR_UINT32_VALUES@@
}

var tier1ConstructorInt64Inputs = []int64{
@@TIER1_CONSTRUCTOR_INT64_VALUES@@
}

var tier1ConstructorUint64Inputs = []uint64{
@@TIER1_CONSTRUCTOR_UINT64_VALUES@@
}
