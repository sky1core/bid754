package bidgo

import "testing"

// Mixed-width Tier 1 operand mapping is shared across every measured layer:
// dq = decimal64.x op decimal128.y, qd = decimal128.x op decimal64.y,
// qq = decimal128.x op decimal128.y, and dd = decimal64.x op decimal64.y.

func BenchmarkFairMixedBID64(b *testing.B) {
	runFairBenchmarkRows(b, fairMixedBID64BenchmarkRows(b))
}

func BenchmarkFairMixedBID128(b *testing.B) {
	runFairBenchmarkRows(b, fairMixedBID128BenchmarkRows(b))
}
