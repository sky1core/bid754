package bid754

import (
	"encoding/json"
	"os"
	"testing"
)

type benchmarkInputPair struct {
	X string `json:"x"`
	Y string `json:"y"`
}

type benchmarkInputs struct {
	FormatVersion int                `json:"format_version"`
	Decimal32     benchmarkInputPair `json:"decimal32"`
	Decimal64     benchmarkInputPair `json:"decimal64"`
	Decimal128    benchmarkInputPair `json:"decimal128"`
}

func loadBenchmarkInputs(tb testing.TB) benchmarkInputs {
	tb.Helper()
	data, err := os.ReadFile("testdata/benchmark_inputs.json")
	if err != nil {
		tb.Fatalf("read benchmark input contract: %v", err)
	}
	var inputs benchmarkInputs
	if err := json.Unmarshal(data, &inputs); err != nil {
		tb.Fatalf("parse benchmark input contract: %v", err)
	}
	if inputs.FormatVersion != 1 {
		tb.Fatalf("benchmark input format_version = %d, want 1", inputs.FormatVersion)
	}
	for _, item := range []struct {
		name string
		pair benchmarkInputPair
	}{
		{"decimal32", inputs.Decimal32},
		{"decimal64", inputs.Decimal64},
		{"decimal128", inputs.Decimal128},
	} {
		if item.pair.X == "" || item.pair.Y == "" {
			tb.Fatalf("benchmark input %s requires non-empty x and y", item.name)
		}
	}
	return inputs
}

func exactBenchmarkDecimal32(tb testing.TB, input string) Decimal32BID {
	tb.Helper()
	value, err := NewDecimal32BIDDirect(input)
	if err != nil || !value.IsFinite() {
		tb.Fatalf("Decimal32 benchmark input %q is not finite and exact: value=%v error=%v", input, value, err)
	}
	return value
}

func exactBenchmarkDecimal64(tb testing.TB, input string) Decimal64BID {
	tb.Helper()
	value, err := NewDecimal64BIDDirect(input)
	if err != nil || !value.IsFinite() {
		tb.Fatalf("Decimal64 benchmark input %q is not finite and exact: value=%v error=%v", input, value, err)
	}
	return value
}

func exactBenchmarkDecimal128(tb testing.TB, input string) Decimal128BID {
	tb.Helper()
	value, err := NewDecimal128BIDDirect(input)
	if err != nil || !value.IsFinite() {
		tb.Fatalf("Decimal128 benchmark input %q is not finite and exact: value=%v error=%v", input, value, err)
	}
	return value
}

func TestBenchmarkInputsAreFiniteAndExact(t *testing.T) {
	inputs := loadBenchmarkInputs(t)
	exactBenchmarkDecimal32(t, inputs.Decimal32.X)
	exactBenchmarkDecimal32(t, inputs.Decimal32.Y)
	exactBenchmarkDecimal64(t, inputs.Decimal64.X)
	exactBenchmarkDecimal64(t, inputs.Decimal64.Y)
	exactBenchmarkDecimal128(t, inputs.Decimal128.X)
	exactBenchmarkDecimal128(t, inputs.Decimal128.Y)
}
