package bidgo

import "testing"

// Layer: Go mechanical port called directly on the shared exact-operand
// contract (../../testdata/benchmark_inputs.json). Decimal32 additionally has
// value-only *_pure rows because those public methods route to separate port
// bodies. FMA consumes z as x*y+z; Sqrt reuses the non-negative x operand.

var sink64 uint64
var sink32 uint32
var sink128 BID_UINT128
var sinkString string
var sinkFlags uint32
var sinkInt64 int64
var sinkInt int

func BenchmarkFairBID32(b *testing.B) {
	runFairBenchmarkRows(b, fairBID32BenchmarkRows(b))
}

func BenchmarkFairBID64(b *testing.B) {
	runFairBenchmarkRows(b, fairBID64BenchmarkRows(b))
}

func BenchmarkFairBID128(b *testing.B) {
	runFairBenchmarkRows(b, fairBID128BenchmarkRows(b))
}
