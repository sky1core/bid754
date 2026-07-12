package bidgo

import "testing"

// Layer: Go mechanical port called directly (status-aware *WithFlags bodies),
// on the shared exact-operand contract (../../testdata/benchmark_inputs.json).
// The bid32 add_pure/mul_pure/div_pure rows measure the separate value-only
// port bodies that the value-only public methods route to; 64/128 have no
// separate pure bodies, so no pure rows exist there. See
// aligned_benchmark_test.go for the full bench-name -> layer matrix.

var sink64 uint64
var sink32 uint32
var sink128 BID_UINT128
var sinkString string
var sinkFlags uint32

func BenchmarkFairBID32(b *testing.B) {
	inputs := loadBenchmarkInputs(b)
	x := exactBenchmarkDecimal32(b, inputs.Decimal32.X)
	y := exactBenchmarkDecimal32(b, inputs.Decimal32.Y)

	b.Run("add", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			sink32, sinkFlags = Bid32AddWithFlags(x, y, 0)
		}
	})
	b.Run("mul", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			sink32, sinkFlags = Bid32MulWithFlags(x, y, 0)
		}
	})
	b.Run("div", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			sink32, sinkFlags = Bid32DivWithFlags(x, y, 0)
		}
	})
	b.Run("add_pure", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			sink32 = Bid32Add(x, y, 0)
		}
	})
	b.Run("mul_pure", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			sink32 = Bid32Mul(x, y, 0)
		}
	})
	b.Run("div_pure", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			sink32 = Bid32Div(x, y, 0)
		}
	})
	b.Run("parse", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			sink32, sinkFlags = Bid32FromStringRaw(inputs.Decimal32.X, 0)
		}
	})
	b.Run("to_string", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			sinkString = Bid32ToStringRaw(x)
		}
	})
}

func BenchmarkFairBID64(b *testing.B) {
	inputs := loadBenchmarkInputs(b)
	x := exactBenchmarkDecimal64(b, inputs.Decimal64.X)
	y := exactBenchmarkDecimal64(b, inputs.Decimal64.Y)

	b.Run("add", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			sink64, sinkFlags = Bid64AddWithFlags(x, y, 0)
		}
	})
	b.Run("mul", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			sink64, sinkFlags = Bid64MulWithFlags(x, y, 0)
		}
	})
	b.Run("div", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			sink64, sinkFlags = Bid64DivWithFlags(x, y, 0)
		}
	})
	b.Run("parse", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			sink64, sinkFlags = Bid64FromString(inputs.Decimal64.X, 0)
		}
	})
	b.Run("to_string", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			sinkString = Bid64ToString(x)
		}
	})
}

func BenchmarkFairBID128(b *testing.B) {
	inputs := loadBenchmarkInputs(b)
	x := exactBenchmarkDecimal128(b, inputs.Decimal128.X)
	y := exactBenchmarkDecimal128(b, inputs.Decimal128.Y)

	b.Run("add", func(b *testing.B) {
		b.ReportAllocs()
		var pfpsf uint32
		for i := 0; i < b.N; i++ {
			pfpsf = 0
			sink128 = Bid128Add(x, y, 0, &pfpsf)
			sinkFlags = pfpsf
		}
	})
	b.Run("mul", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			sink128, sinkFlags = Bid128Mul(x, y, 0)
		}
	})
	b.Run("div", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			sink128, sinkFlags = Bid128Div(x, y, 0)
		}
	})
	b.Run("parse", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			sink128, sinkFlags = Bid128FromString(inputs.Decimal128.X, 0)
		}
	})
	b.Run("to_string", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			sinkString = Bid128ToString(x)
		}
	})
}
