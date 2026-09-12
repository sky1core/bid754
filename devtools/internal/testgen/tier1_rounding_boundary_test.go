package testgen

import (
	"math/big"
	"slices"
	"testing"
)

func TestTier1NativeRoundingBoundaryConstructors(t *testing.T) {
	i32, u32 := tier1ConstructorInt32Inputs(), tier1ConstructorUint32Inputs()
	i64, u64 := tier1ConstructorInt64Inputs(), tier1ConstructorUint64Inputs()
	if len(i32) != 293 || len(u32) != 151 || len(i64) != 1295 || len(u64) != 737 {
		t.Fatalf("constructor counts=%d/%d/%d/%d want293/151/1295/737", len(i32), len(u32), len(i64), len(u64))
	}
	for _, n := range []int64{12345665000000001, 12345664999999999, 12345675000000001, 12345674999999999} {
		if !slices.Contains(i64, n) || !slices.Contains(i64, -n) || !slices.Contains(u64, uint64(n)) {
			t.Errorf("missing signed/unsigned witness %d", n)
		}
	}
	for _, n := range []int32{-2147483648, -2147483647, 2147483646, 2147483647} {
		if !slices.Contains(i32, n) {
			t.Errorf("missing int32 limit %d", n)
		}
	}
	for _, n := range []uint32{0, 1, 4294967294, 4294967295} {
		if !slices.Contains(u32, n) {
			t.Errorf("missing uint32 limit %d", n)
		}
	}
	for _, n := range []int64{-9223372036854775808, -9223372036854775807, 9223372036854775806, 9223372036854775807} {
		if !slices.Contains(i64, n) {
			t.Errorf("missing int64 limit %d", n)
		}
	}
	for _, n := range []uint64{0, 1, 18446744073709551614, 18446744073709551615} {
		if !slices.Contains(u64, n) {
			t.Errorf("missing uint64 limit %d", n)
		}
	}
	for _, precision := range []int{7, 16} {
		for shift := 1; shift <= 19-precision; shift++ {
			scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(shift)), nil)
			q := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(precision-1)), nil)
			for parity := int64(0); parity < 2; parity++ {
				mid := new(big.Int).Mul(new(big.Int).Add(q, big.NewInt(parity)), scale)
				mid.Add(mid, new(big.Int).Quo(scale, big.NewInt(2)))
				for delta := int64(-2); delta <= 2; delta++ {
					n := new(big.Int).Add(mid, big.NewInt(delta)).Int64()
					if !slices.Contains(i64, n) || !slices.Contains(i64, -n) || !slices.Contains(u64, uint64(n)) {
						t.Errorf("missing p=%d shift=%d parity=%d delta=%d", precision, shift, parity, delta)
					}
				}
			}
		}
	}
	t.Logf("native constructor counts int32=%d uint32=%d int64=%d uint64=%d", len(i32), len(u32), len(i64), len(u64))
}

func TestTier1NativeRoundingBoundaryNarrowing(t *testing.T) {
	v32, v64, v128, err := tier1ConversionSemanticInputs()
	if err != nil {
		t.Fatal(err)
	}
	if len(v64) != 1744 || len(v128) != 4267 {
		t.Fatalf("semantic counts=%d/%d want1744/4267", len(v64), len(v128))
	}
	for _, source := range []int{16, 34} {
		for _, spec := range tier1NativeNarrowingBoundaries(source) {
			if source == 16 {
				v, err := tier1EncodeSemantic64(spec)
				if err != nil || !slices.Contains(v64, v) && !slices.Contains(tier1SharedLongBoundary64Values(), v) {
					t.Fatalf("missing native decimal64 boundary %+v: %v", spec, err)
				}
			} else {
				v, err := tier1EncodeSemantic128(spec)
				if err != nil || !slices.Contains(v128, v) && !slices.Contains(tier1SharedLongBoundary128Values(), v) {
					t.Fatalf("missing native decimal128 boundary %+v: %v", spec, err)
				}
			}
		}
	}
	for _, coeff := range []string{"12345665000000001", "12345664999999999", "12345675000000001", "12345674999999999"} {
		for _, negative := range []bool{false, true} {
			v, err := tier1EncodeSemantic128(tier1DecimalInputSpec{coefficient: coeff, negative: negative})
			if err != nil || !slices.Contains(v128, v) {
				t.Errorf("missing multi-rounding width boundary %s negative=%v: %v", coeff, negative, err)
			}
		}
	}
	t.Logf("native narrowing specs decimal64=%d decimal128=%d semantic64=%d semantic128=%d", len(tier1NativeNarrowingBoundaries(16)), len(tier1NativeNarrowingBoundaries(34)), len(v64), len(v128))
	counts := tier1CompareConversionCountsFor(
		uint64(len(tier1SharedLongBoundary32Values())), uint64(len(tier1SharedLongBoundary64Values())), uint64(len(tier1SharedLongBoundary128Values())),
		uint64(len(v32)), uint64(len(v64)), uint64(len(v128)),
		uint64(len(tier1ConstructorInt32Inputs())), uint64(len(tier1ConstructorUint32Inputs())), uint64(len(tier1ConstructorInt64Inputs())), uint64(len(tier1ConstructorUint64Inputs())),
	)
	t.Logf("native regenerated counts=%+v", counts)
}
