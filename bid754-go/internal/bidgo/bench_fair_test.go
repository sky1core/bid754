package bidgo_test

import (
	"testing"

	"github.com/sky1core/bid754/bid754-go/internal/benchrows"
)

// Layer: Go mechanical port called directly on the shared exact-operand
// contract (../../testdata/benchmark_inputs.json). Decimal32 additionally has
// value-only *_pure rows because those public methods route to separate port
// bodies. FMA consumes z as x*y+z; Sqrt reuses the non-negative x operand.
//
// The row table, sinks, and timing-loop bodies live in
// internal/benchrows so the native benchmark preflight can execute the same
// rows from the module root and exact-compare them against the Intel C
// benchmark leg; this file only registers the unchanged benchmark names.

func preparedFairBenchmarkInputs(b *testing.B) benchrows.Prepared {
	b.Helper()
	inputs, err := benchrows.LoadInputs("../../testdata/benchmark_inputs.json")
	if err != nil {
		b.Fatal(err)
	}
	prepared, err := benchrows.Prepare(inputs)
	if err != nil {
		b.Fatal(err)
	}
	return prepared
}

func runFairBenchmarkRows(b *testing.B, rows []benchrows.Row) {
	b.Helper()
	for _, row := range rows {
		row := row
		b.Run(row.Name, func(b *testing.B) {
			b.ReportAllocs()
			row.Run(b.N)
		})
	}
}

func BenchmarkFairBID32(b *testing.B) {
	runFairBenchmarkRows(b, benchrows.FairBID32Rows(preparedFairBenchmarkInputs(b)))
}

func BenchmarkFairBID64(b *testing.B) {
	runFairBenchmarkRows(b, benchrows.FairBID64Rows(preparedFairBenchmarkInputs(b)))
}

func BenchmarkFairBID128(b *testing.B) {
	runFairBenchmarkRows(b, benchrows.FairBID128Rows(preparedFairBenchmarkInputs(b)))
}
