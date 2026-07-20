// Package benchcompare benchmarks bid754-go against shopspring/decimal, the
// most widely used Go decimal library, over parse, to-string, parts
// encode/decode, and the add/mul/div arithmetic core.
//
// This is comparative benchmark infrastructure, not a verification domain:
// the two libraries implement different arithmetic models, so rows compare
// cost on a shared operand set, not result equality. Fairness notes:
//
//   - operand literals are exactly representable in Decimal64 (<= 16
//     significant digits) and in shopspring (arbitrary precision);
//   - div semantics differ (bid754: IEEE 754-2019 round-to-nearest-even to
//     16 digits; shopspring: DivisionPrecision-digit result) — div rows
//     compare cost only;
//   - shopspring stores (coefficient, exponent) natively, so its parts
//     accessors are near-free by construction, while the BID side must
//     decode the interchange encoding; the parts rows measure that
//     asymmetry rather than hiding it;
//   - every result lands in a package-level sink and allocations are
//     reported for every row.
package benchcompare

import (
	"math/big"
	"testing"

	shop "github.com/shopspring/decimal"
	codec "github.com/sky1core/bid754/bid754-codec-go"
	bid "github.com/sky1core/bid754/bid754-go"
)

var parseInputs = []string{
	"0", "1", "42", "1.5", "-2.25", "12345.67", "0.001", "-999999.99",
	"123456789.123456", "9999999999999.99", "3.14159265358979", "-0.00000001",
}

type partsRow struct {
	mant int64
	exp  int32
}

var partsRows = []partsRow{
	{0, 0}, {1, 0}, {42, 0}, {15, -1}, {-225, -2}, {1234567, -2},
	{1, -3}, {-99999999, -2}, {123456789123456, -6}, {999999999999999, -2},
	{314159265358979, -14}, {-1, -8},
}

var (
	sinkBid   bid.Decimal64BID
	sinkShop  shop.Decimal
	sinkStr   string
	sinkU64   uint64
	sinkI64   int64
	sinkI32   int32
	sinkErr   error
	sinkComp  codec.Components
	sinkFlags bid.ExceptionFlags
	bidVals   []bid.Decimal64BID
	shopVals  []shop.Decimal
	bidPairs  [][2]bid.Decimal64BID
	shopPairs [][2]shop.Decimal
	bidDivPr  [][2]bid.Decimal64BID
	shopDivPr [][2]shop.Decimal
)

func init() {
	for _, s := range parseInputs {
		bv, flags := bid.ParseDecimal64BIDRaw(s)
		if flags != 0 {
			panic("benchcompare operand contract: inexact parse input " + s)
		}
		sv, err := shop.NewFromString(s)
		if err != nil {
			panic(err)
		}
		bidVals = append(bidVals, bv)
		shopVals = append(shopVals, sv)
	}
	n := len(bidVals)
	for i := 0; i < n; i++ {
		j := (i*7 + 3) % n
		bidPairs = append(bidPairs, [2]bid.Decimal64BID{bidVals[i], bidVals[j]})
		shopPairs = append(shopPairs, [2]shop.Decimal{shopVals[i], shopVals[j]})
	}
	for i := 0; i < n; i++ {
		j := (i*5 + 1) % n
		if shopVals[j].IsZero() {
			j = (j + 1) % n
		}
		bidDivPr = append(bidDivPr, [2]bid.Decimal64BID{bidVals[i], bidVals[j]})
		shopDivPr = append(shopDivPr, [2]shop.Decimal{shopVals[i], shopVals[j]})
	}
}

func BenchmarkParse_bid754(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		v, f := bid.ParseDecimal64BIDRaw(parseInputs[i%len(parseInputs)])
		sinkBid, sinkFlags = v, f
	}
}

func BenchmarkParse_shopspring(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		v, err := shop.NewFromString(parseInputs[i%len(parseInputs)])
		sinkShop, sinkErr = v, err
	}
}

func BenchmarkToString_bid754(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sinkStr = bidVals[i%len(bidVals)].String()
	}
}

func BenchmarkToString_shopspring(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sinkStr = shopVals[i%len(shopVals)].String()
	}
}

// BenchmarkFromParts_bid754codec includes the caller-side big.NewInt because a
// caller starting from (int64 mantissa, int32 scale) must build one; the
// shopspring row's internal big.Int wrap is the same story on its side.
func BenchmarkFromParts_bid754codec(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		k := i % len(partsRows)
		m := partsRows[k].mant
		neg := m < 0
		if m < 0 {
			m = -m
		}
		kind := codec.Normal
		if m == 0 {
			kind = codec.Zero
		}
		v, err := codec.Encode64(codec.Components{
			Sign: neg, Coefficient: big.NewInt(m), Exponent: partsRows[k].exp, Kind: kind,
		})
		sinkU64, sinkErr = v, err
	}
}

func BenchmarkFromParts_shopspring(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		k := i % len(partsRows)
		sinkShop = shop.New(partsRows[k].mant, partsRows[k].exp)
	}
}

func BenchmarkToParts_bid754codec(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sinkComp = codec.Decode64(uint64(bidVals[i%len(bidVals)]))
	}
}

func BenchmarkToParts_shopspring(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		v := shopVals[i%len(shopVals)]
		sinkI64 = v.CoefficientInt64()
		sinkI32 = v.Exponent()
	}
}

func BenchmarkAdd_bid754(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		p := bidPairs[i%len(bidPairs)]
		sinkBid = p[0].Add(p[1])
	}
}

func BenchmarkAdd_shopspring(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		p := shopPairs[i%len(shopPairs)]
		sinkShop = p[0].Add(p[1])
	}
}

func BenchmarkMul_bid754(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		p := bidPairs[i%len(bidPairs)]
		sinkBid = p[0].Mul(p[1])
	}
}

func BenchmarkMul_shopspring(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		p := shopPairs[i%len(shopPairs)]
		sinkShop = p[0].Mul(p[1])
	}
}

func BenchmarkDiv_bid754(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		p := bidDivPr[i%len(bidDivPr)]
		sinkBid = p[0].Div(p[1])
	}
}

func BenchmarkDiv_shopspring(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		p := shopDivPr[i%len(shopDivPr)]
		sinkShop = p[0].Div(p[1])
	}
}
