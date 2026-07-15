//go:build cgo && bid754_native
// +build cgo,bid754_native

package bid754

import "testing"

// Mixed-width Tier 1 operand mapping is shared across every measured layer:
// dq = decimal64.x op decimal128.y, qd = decimal128.x op decimal64.y,
// qq = decimal128.x op decimal128.y, and dd = decimal64.x op decimal64.y.

func BenchmarkAlignedMixedBID64(b *testing.B) {
	requireNativeBenchmark(b)
	runAlignedBenchmarkRows(b, alignedMixedBID64BenchmarkRows(b))
}

func BenchmarkAlignedMixedBID128(b *testing.B) {
	requireNativeBenchmark(b)
	runAlignedBenchmarkRows(b, alignedMixedBID128BenchmarkRows(b))
}
