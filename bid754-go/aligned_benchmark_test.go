//go:build cgo && bid754_native
// +build cgo,bid754_native

package bid754

import "testing"

// Layer: public Go API (module root), measured on the shared exact-operand
// contract (testdata/benchmark_inputs.json). Matrix mapping:
//   BenchmarkAlignedBID*  -> public Go API        (this file)
//   BenchmarkFairBID*     -> Go mechanical port   (internal/bidgo/bench_fair_test.go)
//   BenchmarkIntelCBID*   -> Intel C direct       (c_bid_benchmark_test.go, cgo-amortized)
//   bid*/op (Criterion)   -> generated Rust       (bid754-rs/benches/core.rs)
//
// bid32 asymmetry: the value-only public methods (Add/Mul/Div) route to the
// separate pure port bodies, while the *WithFlags rows route to the
// status-aware port bodies that BenchmarkFairBID32 measures. Compare
// add<->add_pure and add_with_flags<->FairBID32/add; do not read
// AlignedBID32/add as a wrapper over FairBID32/add. For 64/128 the value-only
// methods wrap the status-aware bodies, so both rows share one implementation.

var (
	alignedSink32     Decimal32BID
	alignedSink64     Decimal64BID
	alignedSink128    Decimal128BID
	alignedSinkString string
	alignedSinkErr    error
	alignedSinkFlags  ExceptionFlags
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
	b.Run("add_with_flags", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			alignedSink32, alignedSinkFlags = x.AddWithFlags(y)
		}
	})
	b.Run("mul_with_flags", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			alignedSink32, alignedSinkFlags = x.MulWithFlags(y)
		}
	})
	b.Run("div_with_flags", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			alignedSink32, alignedSinkFlags = x.DivWithFlags(y)
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
	b.Run("add_with_flags", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			alignedSink64, alignedSinkFlags = x.AddWithFlags(y)
		}
	})
	b.Run("mul_with_flags", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			alignedSink64, alignedSinkFlags = x.MulWithFlags(y)
		}
	})
	b.Run("div_with_flags", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			alignedSink64, alignedSinkFlags = x.DivWithFlags(y)
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
	b.Run("add_with_flags", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			alignedSink128, alignedSinkFlags = x.AddWithFlags(y)
		}
	})
	b.Run("mul_with_flags", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			alignedSink128, alignedSinkFlags = x.MulWithFlags(y)
		}
	})
	b.Run("div_with_flags", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			alignedSink128, alignedSinkFlags = x.DivWithFlags(y)
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
