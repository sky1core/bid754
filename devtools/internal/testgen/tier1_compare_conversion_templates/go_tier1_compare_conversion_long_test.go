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
		tier1WidthStructured32 + tier1WidthStructured64 + tier1WidthStructured128 + tier1ConstructorStructured
	tier1ConversionRandom = tier1ToIntegerRandom32 + tier1ToIntegerRandom64 + tier1ToIntegerRandom128 +
		tier1WidthRandom32 + tier1WidthRandom64 + tier1WidthRandom128 + tier1ConstructorRandom
	tier1ConversionTotal = tier1ConversionStructured + tier1ConversionRandom
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

func tier1CheckQuiet32(t *testing.T, operation string, x, y uint32) {
	t.Helper()
	native, nativeFlags, port, portFlags := runGeneratedFFICase32IntBinary(operation, x, y)
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
	publicFlags, unknownFlags := tier1ArithmeticPublicRawFlags(gotFlags)
	public := 0
	if got {
		public = 1
	}
	if unknownFlags != 0 || native != port || nativeFlags != portFlags || native != public || nativeFlags != publicFlags {
		t.Fatalf("decimal32 %s mismatch: x=%08x y=%08x C=%d/%08x port=%d/%08x public=%d/%08x unknown_public_flags=%s", operation, x, y, native, nativeFlags, port, portFlags, public, publicFlags, unknownFlags)
	}
}

func tier1CheckQuiet64(t *testing.T, operation string, x, y uint64) {
	t.Helper()
	native, nativeFlags, port, portFlags := runGeneratedFFICase64IntBinary(operation, x, y)
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
	publicFlags, unknownFlags := tier1ArithmeticPublicRawFlags(gotFlags)
	public := 0
	if got {
		public = 1
	}
	if unknownFlags != 0 || native != port || nativeFlags != portFlags || native != public || nativeFlags != publicFlags {
		t.Fatalf("decimal64 %s mismatch: x=%016x y=%016x C=%d/%08x port=%d/%08x public=%d/%08x unknown_public_flags=%s", operation, x, y, native, nativeFlags, port, portFlags, public, publicFlags, unknownFlags)
	}
}

func tier1CheckQuiet128(t *testing.T, operation string, x, y Decimal128BID) {
	t.Helper()
	native, nativeFlags, port, portFlags := runGeneratedFFICase128IntBinary(operation, x, y)
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
	publicFlags, unknownFlags := tier1ArithmeticPublicRawFlags(gotFlags)
	public := 0
	if got {
		public = 1
	}
	if unknownFlags != 0 || native != port || nativeFlags != portFlags || native != public || nativeFlags != publicFlags {
		t.Fatalf("decimal128 %s mismatch: x=%x y=%x C=%d/%08x port=%d/%08x public=%d/%08x unknown_public_flags=%s", operation, x, y, native, nativeFlags, port, portFlags, public, publicFlags, unknownFlags)
	}
}

func tier1CheckMinMax32(t *testing.T, operation string, x, y uint32) {
	t.Helper()
	native, port := runGeneratedFFICase32Binary("bid32_"+operation, x, y, 0)
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
	publicFlags, unknownFlags := tier1ArithmeticPublicRawFlags(gotFlags)
	public := fmt.Sprintf("%08x/%08x", got.ToUint32(), publicFlags)
	// The value-only port bodies (bid32_minnum_pure / bid32_maxnum_pure /
	// bid32_minnum_mag_pure / bid32_maxnum_mag_pure) are separate
	// implementations from the status-aware bodies, so their result bits are
	// compared against the native value directly.
	pure := fmt.Sprintf("%08x", pureBits)
	if unknownFlags != 0 || native != port || native != public || !strings.HasPrefix(native, pure+"/") {
		t.Fatalf("decimal32 %s mismatch: x=%08x y=%08x C=%s port=%s public=%s pure=%s unknown_public_flags=%s", operation, x, y, native, port, public, pure, unknownFlags)
	}
}

func tier1CheckMinMax64(t *testing.T, operation string, x, y uint64) {
	t.Helper()
	native, port := runGeneratedFFICase64Binary("bid64_"+operation, x, y, 0)
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
	publicFlags, unknownFlags := tier1ArithmeticPublicRawFlags(gotFlags)
	public := fmt.Sprintf("%016x/%08x", got.ToUint64(), publicFlags)
	if unknownFlags != 0 || native != port || native != public {
		t.Fatalf("decimal64 %s mismatch: x=%016x y=%016x C=%s port=%s public=%s unknown_public_flags=%s", operation, x, y, native, port, public, unknownFlags)
	}
}

func tier1CheckMinMax128(t *testing.T, operation string, x, y Decimal128BID) {
	t.Helper()
	native, port := runGeneratedFFICase128Binary("bid128_"+operation, x, y, 0)
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
	publicFlags, unknownFlags := tier1ArithmeticPublicRawFlags(gotFlags)
	public := fmt.Sprintf("%s/%08x", formatFFIUint128Bits(got), publicFlags)
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
		for i := uint64(0); i < tier1CompareRandomPairs32; i++ {
			x := uint32(tier1ArithmeticRandomWord(0xdec75432c04d5001, i, 0))
			y := uint32(tier1ArithmeticRandomWord(0xdec75432c04d5001, i, 1))
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
		if comparison != tier1CompareRandomComparisons32 {
			t.Fatalf("decimal32 random compare/minmax comparisons=%d want=%d", comparison, tier1CompareRandomComparisons32)
		}
		t.Logf("decimal32 random compare/minmax exact comparisons: %d/%d", shard.ownedCount(comparison), comparison)
	})
	t.Run("decimal64", func(t *testing.T) {
		var comparison uint64
		for i := uint64(0); i < tier1CompareRandomPairs64; i++ {
			x := tier1ArithmeticRandomWord(0xdec75464c04d5001, i, 0)
			y := tier1ArithmeticRandomWord(0xdec75464c04d5001, i, 1)
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
		if comparison != tier1CompareRandomComparisons64 {
			t.Fatalf("decimal64 random compare/minmax comparisons=%d want=%d", comparison, tier1CompareRandomComparisons64)
		}
		t.Logf("decimal64 random compare/minmax exact comparisons: %d/%d", shard.ownedCount(comparison), comparison)
	})
	t.Run("decimal128", func(t *testing.T) {
		var comparison uint64
		for i := uint64(0); i < tier1CompareRandomPairs128; i++ {
			x := tier1ArithmeticDecimal128(tier1Arithmetic128Words{
				lo: tier1ArithmeticRandomWord(0xdec754c0c04d5001, i, 0),
				hi: tier1ArithmeticRandomWord(0xdec754c0c04d5001, i, 1),
			})
			y := tier1ArithmeticDecimal128(tier1Arithmetic128Words{
				lo: tier1ArithmeticRandomWord(0xdec754c0c04d5001, i, 2),
				hi: tier1ArithmeticRandomWord(0xdec754c0c04d5001, i, 3),
			})
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

func tier1CheckToInteger(t *testing.T, width int, operand string, operation tier1ToIntegerOperation, public string, unknownFlags ExceptionFlags) {
	t.Helper()
	function := fmt.Sprintf("bid%d_to_%s_%s", width, operation.kind, operation.suffix)
	tc := generatedFFICase{
		Format: fmt.Sprintf("decimal%d", width),
		Operation: fmt.Sprintf("to_%s_%s", operation.kind, operation.suffix),
		Function: function,
		Operands: []string{operand},
	}
	native, err := runGeneratedFFICaseNativeBaseIntegerTo(tc)
	if err != nil {
		t.Fatalf("%s native integer conversion: %v", function, err)
	}
	port, err := runGeneratedFFICaseGoBaseIntegerTo(tc)
	if err != nil {
		t.Fatalf("%s Go-port integer conversion: %v", function, err)
	}
	if unknownFlags != 0 || native != port || native != public {
		t.Fatalf("%s mismatch: operand=%s mode=%s exact=%v C=%s port=%s public=%s unknown_public_flags=%s", function, operand, operation.mode.name, operation.exact, native, port, public, unknownFlags)
	}
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

func tier1CheckWidthConversion(t *testing.T, operation tier1WidthOperation, operand string, public string, unknownFlags ExceptionFlags) {
	t.Helper()
	function := fmt.Sprintf("bid%d_to_bid%d", operation.source, operation.dest)
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
	native, port, err := runGeneratedFFICaseBIDWidthConversion(tc)
	if err != nil {
		t.Fatalf("%s native/port width conversion: %v", function, err)
	}
	if unknownFlags != 0 || native != port || native != public {
		t.Fatalf("%s mismatch: operand=%s mode=%s C=%s port=%s public=%s unknown_public_flags=%s", function, operand, operation.mode.name, native, port, public, unknownFlags)
	}
}

func tier1CheckWidth32(t *testing.T, value uint32, operation tier1WidthOperation) {
	source := Decimal32BID(value)
	var public string
	var unknownFlags ExceptionFlags
	switch operation.dest {
	case 64:
		result, flags := source.ToDecimal64()
		rawFlags, unknown := tier1ArithmeticPublicRawFlags(flags)
		public, unknownFlags = fmt.Sprintf("%016x/%08x", result.ToUint64(), rawFlags), unknown
	case 128:
		result, flags := source.ToDecimal128()
		rawFlags, unknown := tier1ArithmeticPublicRawFlags(flags)
		public, unknownFlags = fmt.Sprintf("%s/%08x", formatFFIUint128Bits(result), rawFlags), unknown
	default:
		t.Fatalf("unknown Decimal32 width-conversion destination %d", operation.dest)
	}
	tier1CheckWidthConversion(t, operation, fmt.Sprintf("%08x", value), public, unknownFlags)
}

func tier1CheckWidth64(t *testing.T, value uint64, operation tier1WidthOperation) {
	source := Decimal64BID(value)
	var public string
	var unknownFlags ExceptionFlags
	switch operation.dest {
	case 32:
		result, flags := source.ToDecimal32(operation.mode.public)
		rawFlags, unknown := tier1ArithmeticPublicRawFlags(flags)
		public, unknownFlags = fmt.Sprintf("%08x/%08x", result.ToUint32(), rawFlags), unknown
	case 128:
		result, flags := source.ToDecimal128()
		rawFlags, unknown := tier1ArithmeticPublicRawFlags(flags)
		public, unknownFlags = fmt.Sprintf("%s/%08x", formatFFIUint128Bits(result), rawFlags), unknown
	default:
		t.Fatalf("unknown Decimal64 width-conversion destination %d", operation.dest)
	}
	tier1CheckWidthConversion(t, operation, fmt.Sprintf("%016x", value), public, unknownFlags)
}

func tier1CheckWidth128(t *testing.T, value Decimal128BID, operation tier1WidthOperation) {
	var public string
	var unknownFlags ExceptionFlags
	switch operation.dest {
	case 32:
		result, flags := value.ToDecimal32(operation.mode.public)
		rawFlags, unknown := tier1ArithmeticPublicRawFlags(flags)
		public, unknownFlags = fmt.Sprintf("%08x/%08x", result.ToUint32(), rawFlags), unknown
	case 64:
		result, flags := value.ToDecimal64(operation.mode.public)
		rawFlags, unknown := tier1ArithmeticPublicRawFlags(flags)
		public, unknownFlags = fmt.Sprintf("%016x/%08x", result.ToUint64(), rawFlags), unknown
	default:
		t.Fatalf("unknown Decimal128 width-conversion destination %d", operation.dest)
	}
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

func tier1CheckConstructor(t *testing.T, operation tier1ConstructorOperation, raw uint64) {
	t.Helper()
	function := fmt.Sprintf("bid%d_from_%s", operation.dest, operation.kind)
	rounding := 0
	if operation.rounded {
		rounding = operation.mode.native
	}
	tc := generatedFFICase{
		Format:    fmt.Sprintf("decimal%d", operation.dest),
		Operation: "from_" + operation.kind,
		Function:  function,
		Rounding:  rounding,
		Operands:  []string{tier1ConstructorOperand(raw, operation.kind)},
	}
	native, port, err := runGeneratedFFICaseBaseIntegerFrom(tc)
	if err != nil {
		t.Fatalf("%s native/port constructor: %v", function, err)
	}
	public, unknownFlags := tier1PublicConstructor(operation, raw)
	if unknownFlags != 0 || native != port || native != public {
		t.Fatalf("%s mismatch: operand=%s mode=%s C=%s port=%s public=%s unknown_public_flags=%s", function, tc.Operands[0], operation.mode.name, native, port, public, unknownFlags)
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

	if structuredDomains == 7 {
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
		for i := uint64(0); i < tier1ToIntegerRandom32; i++ {
			if shard.owns(comparison) {
				value := uint32(tier1ArithmeticRandomWord(0xdec75432c0a70001, i, 0))
				tier1CheckToInteger32(t, value, toIntegerOperations[i%uint64(len(toIntegerOperations))])
			}
			comparison++
		}
		randomTotal += comparison
		randomExecuted += shard.ownedCount(comparison)
		randomDomains++
		t.Logf("Decimal32 random BID-to-integer exact comparisons: %d/%d", shard.ownedCount(comparison), comparison)
	})
	t.Run("bid_to_integer_decimal64", func(t *testing.T) {
		var comparison uint64
		for i := uint64(0); i < tier1ToIntegerRandom64; i++ {
			if shard.owns(comparison) {
				value := tier1ArithmeticRandomWord(0xdec75464c0a70001, i, 0)
				tier1CheckToInteger64(t, value, toIntegerOperations[i%uint64(len(toIntegerOperations))])
			}
			comparison++
		}
		randomTotal += comparison
		randomExecuted += shard.ownedCount(comparison)
		randomDomains++
		t.Logf("Decimal64 random BID-to-integer exact comparisons: %d/%d", shard.ownedCount(comparison), comparison)
	})
	t.Run("bid_to_integer_decimal128", func(t *testing.T) {
		var comparison uint64
		for i := uint64(0); i < tier1ToIntegerRandom128; i++ {
			if shard.owns(comparison) {
				value := tier1ArithmeticDecimal128(tier1Arithmetic128Words{
					lo: tier1ArithmeticRandomWord(0xdec754c0c0a70001, i, 0),
					hi: tier1ArithmeticRandomWord(0xdec754c0c0a70001, i, 1),
				})
				tier1CheckToInteger128(t, value, toIntegerOperations[i%uint64(len(toIntegerOperations))])
			}
			comparison++
		}
		randomTotal += comparison
		randomExecuted += shard.ownedCount(comparison)
		randomDomains++
		t.Logf("Decimal128 random BID-to-integer exact comparisons: %d/%d", shard.ownedCount(comparison), comparison)
	})

	t.Run("width_from_decimal32", func(t *testing.T) {
		operations := tier1WidthOperations(32)
		var comparison uint64
		for i := uint64(0); i < tier1WidthRandom32; i++ {
			if shard.owns(comparison) {
				value := uint32(tier1ArithmeticRandomWord(0xdec75432c0de0001, i, 0))
				tier1CheckWidth32(t, value, operations[i%uint64(len(operations))])
			}
			comparison++
		}
		randomTotal += comparison
		randomExecuted += shard.ownedCount(comparison)
		randomDomains++
		t.Logf("Decimal32-source random width exact comparisons: %d/%d", shard.ownedCount(comparison), comparison)
	})
	t.Run("width_from_decimal64", func(t *testing.T) {
		operations := tier1WidthOperations(64)
		var comparison uint64
		for i := uint64(0); i < tier1WidthRandom64; i++ {
			if shard.owns(comparison) {
				value := tier1ArithmeticRandomWord(0xdec75464c0de0001, i, 0)
				tier1CheckWidth64(t, value, operations[i%uint64(len(operations))])
			}
			comparison++
		}
		randomTotal += comparison
		randomExecuted += shard.ownedCount(comparison)
		randomDomains++
		t.Logf("Decimal64-source random width exact comparisons: %d/%d", shard.ownedCount(comparison), comparison)
	})
	t.Run("width_from_decimal128", func(t *testing.T) {
		operations := tier1WidthOperations(128)
		var comparison uint64
		for i := uint64(0); i < tier1WidthRandom128; i++ {
			if shard.owns(comparison) {
				value := tier1ArithmeticDecimal128(tier1Arithmetic128Words{
					lo: tier1ArithmeticRandomWord(0xdec754c0c0de0001, i, 0),
					hi: tier1ArithmeticRandomWord(0xdec754c0c0de0001, i, 1),
				})
				tier1CheckWidth128(t, value, operations[i%uint64(len(operations))])
			}
			comparison++
		}
		randomTotal += comparison
		randomExecuted += shard.ownedCount(comparison)
		randomDomains++
		t.Logf("Decimal128-source random width exact comparisons: %d/%d", shard.ownedCount(comparison), comparison)
	})

	t.Run("integer_constructors", func(t *testing.T) {
		var comparison uint64
		for i := uint64(0); i < tier1ConstructorRandom; i++ {
			if shard.owns(comparison) {
				operation := constructorOperations[i%uint64(len(constructorOperations))]
				raw := tier1ArithmeticRandomWord(0xdec754c0c0570001, i, 0)
				tier1CheckConstructor(t, operation, raw)
			}
			comparison++
		}
		randomTotal += comparison
		randomExecuted += shard.ownedCount(comparison)
		randomDomains++
		t.Logf("random integer-constructor exact comparisons: %d/%d", shard.ownedCount(comparison), comparison)
	})

	if randomDomains == 7 {
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
