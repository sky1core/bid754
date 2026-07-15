//go:build cgo && bid754_native
// +build cgo,bid754_native

package bid754

import "testing"

// Layer: public Go API (module root), measured on the shared exact-operand
// contract (testdata/benchmark_inputs.json). Matrix mapping:
//
//	BenchmarkAlignedBID*  -> public Go API
//	BenchmarkFairBID*     -> Go mechanical port
//	BenchmarkIntelCBID*   -> Intel C direct (cgo-amortized)
//	bid*/op (Criterion)   -> generated Rust
//
// bid32 asymmetry: value-only Add/Sub/Mul/Div use separate pure-port bodies,
// while the *WithFlags rows route to the status-aware bodies measured by
// BenchmarkFairBID32. For 64/128 the value-only methods wrap the status-aware
// bodies. FMA and Sqrt always return flags. FMA consumes z as x*y+z; Sqrt
// reuses the non-negative x operand.

var (
	alignedSink32     Decimal32BID
	alignedSink64     Decimal64BID
	alignedSink128    Decimal128BID
	alignedSinkString string
	alignedSinkInt64  int64
	alignedSinkBool   bool
	alignedSinkErr    error
	alignedSinkFlags  ExceptionFlags
)

func BenchmarkAlignedBID32(b *testing.B) {
	requireNativeBenchmark(b)
	runAlignedBenchmarkRows(b, alignedBID32BenchmarkRows(b))
}

func BenchmarkAlignedBID64(b *testing.B) {
	requireNativeBenchmark(b)
	runAlignedBenchmarkRows(b, alignedBID64BenchmarkRows(b))
}

func BenchmarkAlignedBID128(b *testing.B) {
	requireNativeBenchmark(b)
	runAlignedBenchmarkRows(b, alignedBID128BenchmarkRows(b))
}
