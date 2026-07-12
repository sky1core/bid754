//go:build cgo && bid754_native
// +build cgo,bid754_native

package bid754

import "testing"

var (
	alignedSink32     Decimal32BID
	alignedSink64     Decimal64BID
	alignedSink128    Decimal128BID
	alignedSinkString string
	alignedSinkErr    error
)

func BenchmarkAlignedBID32(b *testing.B) {
	requireNativeBenchmark(b)
	inputs := loadBenchmarkInputs(b)

	x := exactBenchmarkDecimal32(b, inputs.Decimal32.X)
	y := exactBenchmarkDecimal32(b, inputs.Decimal32.Y)

	b.Run("add", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			alignedSink32 = x.Add(y)
		}
	})
	b.Run("mul", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			alignedSink32 = x.Mul(y)
		}
	})
	b.Run("div", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			alignedSink32 = x.Div(y)
		}
	})
	b.Run("parse", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			alignedSink32, alignedSinkErr = NewDecimal32BIDDirect(inputs.Decimal32.X)
		}
	})
	b.Run("to_string", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			alignedSinkString = x.String()
		}
	})
}

func BenchmarkAlignedBID64(b *testing.B) {
	requireNativeBenchmark(b)
	inputs := loadBenchmarkInputs(b)

	x := exactBenchmarkDecimal64(b, inputs.Decimal64.X)
	y := exactBenchmarkDecimal64(b, inputs.Decimal64.Y)

	b.Run("add", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			alignedSink64 = x.Add(y)
		}
	})
	b.Run("mul", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			alignedSink64 = x.Mul(y)
		}
	})
	b.Run("div", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			alignedSink64 = x.Div(y)
		}
	})
	b.Run("parse", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			alignedSink64, alignedSinkErr = NewDecimal64BIDDirect(inputs.Decimal64.X)
		}
	})
	b.Run("to_string", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			alignedSinkString = x.String()
		}
	})
}

func BenchmarkAlignedBID128(b *testing.B) {
	requireNativeBenchmark(b)
	inputs := loadBenchmarkInputs(b)

	x := exactBenchmarkDecimal128(b, inputs.Decimal128.X)
	y := exactBenchmarkDecimal128(b, inputs.Decimal128.Y)

	b.Run("add", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			alignedSink128 = x.Add(y)
		}
	})
	b.Run("mul", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			alignedSink128 = x.Mul(y)
		}
	})
	b.Run("div", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			alignedSink128 = x.Div(y)
		}
	})
	b.Run("parse", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			alignedSink128, alignedSinkErr = NewDecimal128BIDDirect(inputs.Decimal128.X)
		}
	})
	b.Run("to_string", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			alignedSinkString = x.String()
		}
	})
}
