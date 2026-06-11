//go:build cgo && bid754_native
// +build cgo,bid754_native

package bid754

import "testing"

func runIntelCBIDBench(b *testing.B, fn func(int)) {
	requireNativeBenchmark(b)
	nativeBenchCBIDInit()
	b.ReportAllocs()
	b.ResetTimer()
	fn(b.N)
}

func BenchmarkIntelCBID32(b *testing.B) {
	b.Run("add", func(b *testing.B) { runIntelCBIDBench(b, nativeBenchCBID32Add) })
	b.Run("mul", func(b *testing.B) { runIntelCBIDBench(b, nativeBenchCBID32Mul) })
	b.Run("div", func(b *testing.B) { runIntelCBIDBench(b, nativeBenchCBID32Div) })
	b.Run("parse", func(b *testing.B) { runIntelCBIDBench(b, nativeBenchCBID32Parse) })
	b.Run("to_string", func(b *testing.B) { runIntelCBIDBench(b, nativeBenchCBID32ToString) })
}

func BenchmarkIntelCBID64(b *testing.B) {
	b.Run("add", func(b *testing.B) { runIntelCBIDBench(b, nativeBenchCBID64Add) })
	b.Run("mul", func(b *testing.B) { runIntelCBIDBench(b, nativeBenchCBID64Mul) })
	b.Run("div", func(b *testing.B) { runIntelCBIDBench(b, nativeBenchCBID64Div) })
	b.Run("parse", func(b *testing.B) { runIntelCBIDBench(b, nativeBenchCBID64Parse) })
	b.Run("to_string", func(b *testing.B) { runIntelCBIDBench(b, nativeBenchCBID64ToString) })
}

func BenchmarkIntelCBID128(b *testing.B) {
	b.Run("add", func(b *testing.B) { runIntelCBIDBench(b, nativeBenchCBID128Add) })
	b.Run("mul", func(b *testing.B) { runIntelCBIDBench(b, nativeBenchCBID128Mul) })
	b.Run("div", func(b *testing.B) { runIntelCBIDBench(b, nativeBenchCBID128Div) })
	b.Run("parse", func(b *testing.B) { runIntelCBIDBench(b, nativeBenchCBID128Parse) })
	b.Run("to_string", func(b *testing.B) { runIntelCBIDBench(b, nativeBenchCBID128ToString) })
}
