//go:build cgo && bid754_native
// +build cgo,bid754_native

package bid754

import "testing"

func runIntelCBIDBench(b *testing.B, fn func(int)) {
	requireNativeBenchmark(b)
	inputs := loadBenchmarkInputs(b)
	if !nativeBenchCBIDInit(
		inputs.Decimal32.X,
		inputs.Decimal32.Y,
		inputs.Decimal32.Z,
		inputs.Decimal64.X,
		inputs.Decimal64.Y,
		inputs.Decimal64.Z,
		inputs.Decimal128.X,
		inputs.Decimal128.Y,
		inputs.Decimal128.Z,
	) {
		b.Fatal("Intel C benchmark inputs are not finite and exact")
	}
	b.ReportAllocs()
	b.ResetTimer()
	fn(b.N)
}

func TestIntelCBenchmarkInputsAreFiniteAndExact(t *testing.T) {
	inputs := loadBenchmarkInputs(t)
	if !nativeBenchCBIDInit(
		inputs.Decimal32.X,
		inputs.Decimal32.Y,
		inputs.Decimal32.Z,
		inputs.Decimal64.X,
		inputs.Decimal64.Y,
		inputs.Decimal64.Z,
		inputs.Decimal128.X,
		inputs.Decimal128.Y,
		inputs.Decimal128.Z,
	) {
		t.Fatal("Intel C benchmark inputs are not finite and exact")
	}
}

func BenchmarkIntelCBID32(b *testing.B) {
	b.Run("add", func(b *testing.B) { runIntelCBIDBench(b, nativeBenchCBID32Add) })
	b.Run("mul", func(b *testing.B) { runIntelCBIDBench(b, nativeBenchCBID32Mul) })
	b.Run("div", func(b *testing.B) { runIntelCBIDBench(b, nativeBenchCBID32Div) })
	b.Run("fma", func(b *testing.B) { runIntelCBIDBench(b, nativeBenchCBID32Fma) })
	b.Run("sqrt", func(b *testing.B) { runIntelCBIDBench(b, nativeBenchCBID32Sqrt) })
	b.Run("parse", func(b *testing.B) { runIntelCBIDBench(b, nativeBenchCBID32Parse) })
	b.Run("to_string", func(b *testing.B) { runIntelCBIDBench(b, nativeBenchCBID32ToString) })
}

func BenchmarkIntelCBID64(b *testing.B) {
	b.Run("add", func(b *testing.B) { runIntelCBIDBench(b, nativeBenchCBID64Add) })
	b.Run("mul", func(b *testing.B) { runIntelCBIDBench(b, nativeBenchCBID64Mul) })
	b.Run("div", func(b *testing.B) { runIntelCBIDBench(b, nativeBenchCBID64Div) })
	b.Run("fma", func(b *testing.B) { runIntelCBIDBench(b, nativeBenchCBID64Fma) })
	b.Run("sqrt", func(b *testing.B) { runIntelCBIDBench(b, nativeBenchCBID64Sqrt) })
	b.Run("parse", func(b *testing.B) { runIntelCBIDBench(b, nativeBenchCBID64Parse) })
	b.Run("to_string", func(b *testing.B) { runIntelCBIDBench(b, nativeBenchCBID64ToString) })
}

func BenchmarkIntelCBID128(b *testing.B) {
	b.Run("add", func(b *testing.B) { runIntelCBIDBench(b, nativeBenchCBID128Add) })
	b.Run("mul", func(b *testing.B) { runIntelCBIDBench(b, nativeBenchCBID128Mul) })
	b.Run("div", func(b *testing.B) { runIntelCBIDBench(b, nativeBenchCBID128Div) })
	b.Run("fma", func(b *testing.B) { runIntelCBIDBench(b, nativeBenchCBID128Fma) })
	b.Run("sqrt", func(b *testing.B) { runIntelCBIDBench(b, nativeBenchCBID128Sqrt) })
	b.Run("parse", func(b *testing.B) { runIntelCBIDBench(b, nativeBenchCBID128Parse) })
	b.Run("to_string", func(b *testing.B) { runIntelCBIDBench(b, nativeBenchCBID128ToString) })
}
