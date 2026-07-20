// Package benchcompare benchmarks bid754-go against shopspring/decimal, the
// most widely used Go decimal library, over parse, to-string, parts
// encode/decode, and the add/mul/div arithmetic core, at every product
// width: Decimal32, Decimal64, and Decimal128.
//
// This is comparative benchmark infrastructure, not a verification domain:
// the two libraries implement different arithmetic models, so rows compare
// cost on a shared per-width operand set, not result equality. Fairness
// notes:
//
//   - each width's operand literals are exactly representable at that BID
//     width (7/16/28 significant digits; the d128 set stays within 28 so
//     the same list also stays exact in the Rust module's rust_decimal
//     counterpart) and in shopspring (arbitrary precision);
//   - div semantics differ (bid754: IEEE 754-2019 round-to-nearest-even to
//     the width's precision; shopspring: DivisionPrecision-digit result) —
//     div rows compare cost only;
//   - shopspring stores (coefficient, exponent) natively, so its parts
//     accessors are near-free by construction, while the BID side must
//     decode the interchange encoding; the parts rows measure that
//     asymmetry rather than hiding it;
//   - every result lands in a package-level sink and allocations are
//     reported for every row.
package benchcompare

import (
	"encoding/binary"
	"math/big"
	"testing"

	shop "github.com/shopspring/decimal"
	codec "github.com/sky1core/bid754/bid754-codec-go"
	bid "github.com/sky1core/bid754/bid754-go"
)

// Per-width parse inputs: exactly representable at the width, shared with
// the same-width rows of benchcompare-rs (keep the two lists in sync by
// hand; each language's operand-contract test enforces its own side).
var (
	parseInputsD32 = []string{
		"0", "1", "42", "1.5", "-2.25", "1234.56", "0.001", "-9999.99",
		"1234567", "-999999.9", "3.14159", "-0.0001",
	}
	parseInputsD64 = []string{
		"0", "1", "42", "1.5", "-2.25", "12345.67", "0.001", "-999999.99",
		"123456789.123456", "9999999999999.99", "3.14159265358979", "-0.00000001",
	}
	// d128 list: <= 28 significant digits and value magnitudes within
	// [1e-8, 1e12], so the identical list stays exact and overflow-free in
	// rust_decimal (96-bit mantissa, scale <= 28) for the Rust module.
	parseInputsD128 = []string{
		"0", "1", "42", "1.5", "-2.25", "123456789012.3456789012345678",
		"0.001", "-999999999999.9999999999999999", "1234567890.123456789012345678",
		"3.141592653589793238462643383", "0.9876543210987654321098765432", "-0.00000001",
	}
)

type partsRow struct {
	mant int64
	exp  int32
}

var (
	partsRowsD32 = []partsRow{
		{0, 0}, {1, 0}, {42, 0}, {15, -1}, {-225, -2}, {123456, -2},
		{1, -3}, {-9999999, -2}, {1234567, -4}, {9999999, -1},
		{314159, -5}, {-1, -7},
	}
	partsRowsD64 = []partsRow{
		{0, 0}, {1, 0}, {42, 0}, {15, -1}, {-225, -2}, {1234567, -2},
		{1, -3}, {-99999999, -2}, {123456789123456, -6}, {999999999999999, -2},
		{314159265358979, -14}, {-1, -8},
	}
	partsRowsD128 = []partsRow{
		{0, 0}, {1, 0}, {42, 0}, {15, -1}, {-225, -2}, {1234567, -2},
		{1, -3}, {-99999999, -2}, {123456789012345678, -9},
		{999999999999999999, -12}, {314159265358979323, -17}, {-1, -8},
	}
)

var (
	sink32    bid.Decimal32BID
	sink64    bid.Decimal64BID
	sink128   bid.Decimal128BID
	sinkShop  shop.Decimal
	sinkStr   string
	sinkU32   uint32
	sinkU64   uint64
	sinkI64   int64
	sinkI32   int32
	sinkErr   error
	sinkComp  codec.Components
	sinkBig   *big.Int
	sinkFlags bid.ExceptionFlags
)

var (
	bid32Vals   []bid.Decimal32BID
	bid64Vals   []bid.Decimal64BID
	bid128Vals  []bid.Decimal128BID
	shop32Vals  []shop.Decimal
	shop64Vals  []shop.Decimal
	shop128Vals []shop.Decimal
	pairIdx     [][2]int
	divIdx      [][2]int
)

func mustShop(s string) shop.Decimal {
	v, err := shop.NewFromString(s)
	if err != nil {
		panic(err)
	}
	return v
}

func init() {
	for _, s := range parseInputsD32 {
		v, flags := bid.ParseDecimal32BIDRaw(s)
		if flags != 0 {
			panic("benchcompare operand contract: inexact d32 parse input " + s)
		}
		bid32Vals = append(bid32Vals, v)
		shop32Vals = append(shop32Vals, mustShop(s))
	}
	for _, s := range parseInputsD64 {
		v, flags := bid.ParseDecimal64BIDRaw(s)
		if flags != 0 {
			panic("benchcompare operand contract: inexact d64 parse input " + s)
		}
		bid64Vals = append(bid64Vals, v)
		shop64Vals = append(shop64Vals, mustShop(s))
	}
	for _, s := range parseInputsD128 {
		v, flags := bid.ParseDecimal128BIDRaw(s)
		if flags != 0 {
			panic("benchcompare operand contract: inexact d128 parse input " + s)
		}
		bid128Vals = append(bid128Vals, v)
		shop128Vals = append(shop128Vals, mustShop(s))
	}
	n := len(parseInputsD32)
	if len(parseInputsD64) != n || len(parseInputsD128) != n {
		panic("benchcompare operand contract: width lists must share one length")
	}
	for i := 0; i < n; i++ {
		pairIdx = append(pairIdx, [2]int{i, (i*7 + 3) % n})
	}
	// index 0 is the literal "0" in every width list; skip it as a divisor.
	for i := 0; i < n; i++ {
		j := (i*5 + 1) % n
		if j == 0 {
			j = 1
		}
		divIdx = append(divIdx, [2]int{i, j})
	}
}

func partsComponents(r partsRow) codec.Components {
	m := r.mant
	neg := m < 0
	if m < 0 {
		m = -m
	}
	kind := codec.Normal
	if m == 0 {
		kind = codec.Zero
	}
	return codec.Components{Sign: neg, Coefficient: big.NewInt(m), Exponent: r.exp, Kind: kind}
}

func d128Words(v bid.Decimal128BID) (lo, hi uint64) {
	return binary.LittleEndian.Uint64(v[0:8]), binary.LittleEndian.Uint64(v[8:16])
}

func BenchmarkParse(b *testing.B) {
	n := len(parseInputsD32)
	b.Run("d32_bid754", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			sink32, sinkFlags = bid.ParseDecimal32BIDRaw(parseInputsD32[i%n])
		}
	})
	b.Run("d32_shopspring", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			sinkShop, sinkErr = shop.NewFromString(parseInputsD32[i%n])
		}
	})
	b.Run("d64_bid754", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			sink64, sinkFlags = bid.ParseDecimal64BIDRaw(parseInputsD64[i%n])
		}
	})
	b.Run("d64_shopspring", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			sinkShop, sinkErr = shop.NewFromString(parseInputsD64[i%n])
		}
	})
	b.Run("d128_bid754", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			sink128, sinkFlags = bid.ParseDecimal128BIDRaw(parseInputsD128[i%n])
		}
	})
	b.Run("d128_shopspring", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			sinkShop, sinkErr = shop.NewFromString(parseInputsD128[i%n])
		}
	})
}

func BenchmarkToString(b *testing.B) {
	n := len(bid32Vals)
	b.Run("d32_bid754", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			sinkStr = bid32Vals[i%n].String()
		}
	})
	b.Run("d32_shopspring", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			sinkStr = shop32Vals[i%n].String()
		}
	})
	b.Run("d64_bid754", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			sinkStr = bid64Vals[i%n].String()
		}
	})
	b.Run("d64_shopspring", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			sinkStr = shop64Vals[i%n].String()
		}
	})
	b.Run("d128_bid754", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			sinkStr = bid128Vals[i%n].String()
		}
	})
	b.Run("d128_shopspring", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			sinkStr = shop128Vals[i%n].String()
		}
	})
}

// FromParts rows include the caller-side big.NewInt because a caller
// starting from (int64 mantissa, int32 scale) must build one; the shopspring
// row's internal big.Int wrap is the same story on its side.
func BenchmarkFromParts(b *testing.B) {
	b.Run("d32_bid754codec", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			r := partsRowsD32[i%len(partsRowsD32)]
			m := r.mant
			neg := m < 0
			if m < 0 {
				m = -m
			}
			kind := codec.Normal
			if m == 0 {
				kind = codec.Zero
			}
			v, err := codec.Encode32(codec.Components{
				Sign: neg, Coefficient: big.NewInt(m), Exponent: r.exp, Kind: kind,
			})
			sinkU32, sinkErr = v, err
		}
	})
	b.Run("d32_shopspring", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			r := partsRowsD32[i%len(partsRowsD32)]
			sinkShop = shop.New(r.mant, r.exp)
		}
	})
	b.Run("d64_bid754codec", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			r := partsRowsD64[i%len(partsRowsD64)]
			m := r.mant
			neg := m < 0
			if m < 0 {
				m = -m
			}
			kind := codec.Normal
			if m == 0 {
				kind = codec.Zero
			}
			v, err := codec.Encode64(codec.Components{
				Sign: neg, Coefficient: big.NewInt(m), Exponent: r.exp, Kind: kind,
			})
			sinkU64, sinkErr = v, err
		}
	})
	b.Run("d64_shopspring", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			r := partsRowsD64[i%len(partsRowsD64)]
			sinkShop = shop.New(r.mant, r.exp)
		}
	})
	b.Run("d128_bid754codec", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			r := partsRowsD128[i%len(partsRowsD128)]
			m := r.mant
			neg := m < 0
			if m < 0 {
				m = -m
			}
			kind := codec.Normal
			if m == 0 {
				kind = codec.Zero
			}
			lo, hi, err := codec.Encode128(codec.Components{
				Sign: neg, Coefficient: big.NewInt(m), Exponent: r.exp, Kind: kind,
			})
			sinkU64, sinkErr = lo^hi, err
		}
	})
	b.Run("d128_shopspring", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			r := partsRowsD128[i%len(partsRowsD128)]
			sinkShop = shop.New(r.mant, r.exp)
		}
	})
}

func BenchmarkToParts(b *testing.B) {
	n := len(bid32Vals)
	b.Run("d32_bid754codec", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			sinkComp = codec.Decode32(uint32(bid32Vals[i%n]))
		}
	})
	b.Run("d32_shopspring", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			v := shop32Vals[i%n]
			sinkI64 = v.CoefficientInt64()
			sinkI32 = v.Exponent()
		}
	})
	b.Run("d64_bid754codec", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			sinkComp = codec.Decode64(uint64(bid64Vals[i%n]))
		}
	})
	b.Run("d64_shopspring", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			v := shop64Vals[i%n]
			sinkI64 = v.CoefficientInt64()
			sinkI32 = v.Exponent()
		}
	})
	b.Run("d128_bid754codec", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			lo, hi := d128Words(bid128Vals[i%n])
			sinkComp = codec.Decode128(lo, hi)
		}
	})
	// d128 coefficients exceed int64, so the shopspring row uses
	// Coefficient() (a full big.Int copy) — the same full-coefficient
	// recovery the BID row performs via Decode128.
	b.Run("d128_shopspring", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			v := shop128Vals[i%n]
			sinkBig = v.Coefficient()
			sinkI32 = v.Exponent()
		}
	})
}

func BenchmarkAdd(b *testing.B) {
	n := len(pairIdx)
	b.Run("d32_bid754", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			p := pairIdx[i%n]
			sink32 = bid32Vals[p[0]].Add(bid32Vals[p[1]])
		}
	})
	b.Run("d32_shopspring", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			p := pairIdx[i%n]
			sinkShop = shop32Vals[p[0]].Add(shop32Vals[p[1]])
		}
	})
	b.Run("d64_bid754", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			p := pairIdx[i%n]
			sink64 = bid64Vals[p[0]].Add(bid64Vals[p[1]])
		}
	})
	b.Run("d64_shopspring", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			p := pairIdx[i%n]
			sinkShop = shop64Vals[p[0]].Add(shop64Vals[p[1]])
		}
	})
	b.Run("d128_bid754", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			p := pairIdx[i%n]
			sink128 = bid128Vals[p[0]].Add(bid128Vals[p[1]])
		}
	})
	b.Run("d128_shopspring", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			p := pairIdx[i%n]
			sinkShop = shop128Vals[p[0]].Add(shop128Vals[p[1]])
		}
	})
}

func BenchmarkMul(b *testing.B) {
	n := len(pairIdx)
	b.Run("d32_bid754", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			p := pairIdx[i%n]
			sink32 = bid32Vals[p[0]].Mul(bid32Vals[p[1]])
		}
	})
	b.Run("d32_shopspring", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			p := pairIdx[i%n]
			sinkShop = shop32Vals[p[0]].Mul(shop32Vals[p[1]])
		}
	})
	b.Run("d64_bid754", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			p := pairIdx[i%n]
			sink64 = bid64Vals[p[0]].Mul(bid64Vals[p[1]])
		}
	})
	b.Run("d64_shopspring", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			p := pairIdx[i%n]
			sinkShop = shop64Vals[p[0]].Mul(shop64Vals[p[1]])
		}
	})
	b.Run("d128_bid754", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			p := pairIdx[i%n]
			sink128 = bid128Vals[p[0]].Mul(bid128Vals[p[1]])
		}
	})
	b.Run("d128_shopspring", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			p := pairIdx[i%n]
			sinkShop = shop128Vals[p[0]].Mul(shop128Vals[p[1]])
		}
	})
}

func BenchmarkDiv(b *testing.B) {
	n := len(divIdx)
	b.Run("d32_bid754", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			p := divIdx[i%n]
			sink32 = bid32Vals[p[0]].Div(bid32Vals[p[1]])
		}
	})
	b.Run("d32_shopspring", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			p := divIdx[i%n]
			sinkShop = shop32Vals[p[0]].Div(shop32Vals[p[1]])
		}
	})
	b.Run("d64_bid754", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			p := divIdx[i%n]
			sink64 = bid64Vals[p[0]].Div(bid64Vals[p[1]])
		}
	})
	b.Run("d64_shopspring", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			p := divIdx[i%n]
			sinkShop = shop64Vals[p[0]].Div(shop64Vals[p[1]])
		}
	})
	b.Run("d128_bid754", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			p := divIdx[i%n]
			sink128 = bid128Vals[p[0]].Div(bid128Vals[p[1]])
		}
	})
	b.Run("d128_shopspring", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			p := divIdx[i%n]
			sinkShop = shop128Vals[p[0]].Div(shop128Vals[p[1]])
		}
	})
}
