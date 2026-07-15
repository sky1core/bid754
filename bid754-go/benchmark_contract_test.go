package bid754

import (
	"encoding/json"
	"os"
	"testing"
)

type benchmarkInputPair struct {
	X string `json:"x"`
	Y string `json:"y"`
	// Z is the third fma operand (result = x*y + z); sqrt reuses the
	// non-negative X operand and needs no dedicated input.
	Z string `json:"z"`
}

type benchmarkInputs struct {
	FormatVersion  int                `json:"format_version"`
	IntegerOperand int64              `json:"integer_operand"`
	ScaleExponent  int                `json:"scale_exponent"`
	Decimal32      benchmarkInputPair `json:"decimal32"`
	Decimal64      benchmarkInputPair `json:"decimal64"`
	Decimal128     benchmarkInputPair `json:"decimal128"`
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
	if inputs.FormatVersion != 2 {
		tb.Fatalf("benchmark input format_version = %d, want 2", inputs.FormatVersion)
	}
	if inputs.IntegerOperand == 0 {
		tb.Fatal("benchmark integer_operand must be non-zero")
	}
	if inputs.ScaleExponent == 0 {
		tb.Fatal("benchmark scale_exponent must be non-zero")
	}
	if !benchmarkScaleExponentFitsCInt(int64(inputs.ScaleExponent)) {
		tb.Fatalf("benchmark scale_exponent %d does not fit the Intel C int32 contract", inputs.ScaleExponent)
	}
	for _, item := range []struct {
		name string
		pair benchmarkInputPair
	}{
		{"decimal32", inputs.Decimal32},
		{"decimal64", inputs.Decimal64},
		{"decimal128", inputs.Decimal128},
	} {
		if item.pair.X == "" || item.pair.Y == "" || item.pair.Z == "" {
			tb.Fatalf("benchmark input %s requires non-empty x, y, and z", item.name)
		}
	}
	validateBenchmarkInputs(tb, inputs)
	return inputs
}

func benchmarkScaleExponentFitsCInt(exponent int64) bool {
	return exponent >= -1<<31 && exponent <= 1<<31-1
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
	loadBenchmarkInputs(t)
}

func TestBenchmarkScaleExponentFitsCInt(t *testing.T) {
	for _, tc := range []struct {
		name     string
		exponent int64
		want     bool
	}{
		{name: "min", exponent: -1 << 31, want: true},
		{name: "max", exponent: 1<<31 - 1, want: true},
		{name: "below_min", exponent: -1<<31 - 1, want: false},
		{name: "above_max", exponent: 1 << 31, want: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := benchmarkScaleExponentFitsCInt(tc.exponent); got != tc.want {
				t.Fatalf("benchmarkScaleExponentFitsCInt(%d) = %t, want %t", tc.exponent, got, tc.want)
			}
		})
	}
}

func validateBenchmarkInputs(tb testing.TB, inputs benchmarkInputs) {
	tb.Helper()
	exactBenchmarkDecimal32(tb, inputs.Decimal32.X)
	exactBenchmarkDecimal32(tb, inputs.Decimal32.Y)
	exactBenchmarkDecimal32(tb, inputs.Decimal32.Z)
	exactBenchmarkDecimal64(tb, inputs.Decimal64.X)
	exactBenchmarkDecimal64(tb, inputs.Decimal64.Y)
	exactBenchmarkDecimal64(tb, inputs.Decimal64.Z)
	exactBenchmarkDecimal128(tb, inputs.Decimal128.X)
	exactBenchmarkDecimal128(tb, inputs.Decimal128.Y)
	exactBenchmarkDecimal128(tb, inputs.Decimal128.Z)
	requireNonNegativeSqrtOperands(tb, inputs)
	requireNonZeroDivisionOperands(tb, inputs)
	requireExactIntegerOperand(tb, inputs)
	requireFiniteScaleBOperands(tb, inputs)
}

func requireExactIntegerOperand(tb testing.TB, inputs benchmarkInputs) {
	tb.Helper()

	d32, flags32 := NewDecimal32FromInt64(inputs.IntegerOperand, RoundNearestEven)
	if flags32 != 0 {
		tb.Fatalf("benchmark integer_operand %d is not exact as Decimal32: flags=%v", inputs.IntegerOperand, flags32)
	}
	got32, flags32 := d32.ConvertToInt64(RoundNearestEven)
	if got32 != inputs.IntegerOperand || flags32 != 0 {
		tb.Fatalf("Decimal32 benchmark integer round trip = (%d, %v), want (%d, 0)", got32, flags32, inputs.IntegerOperand)
	}

	d64, flags64 := NewDecimal64FromInt64(inputs.IntegerOperand, RoundNearestEven)
	if flags64 != 0 {
		tb.Fatalf("benchmark integer_operand %d is not exact as Decimal64: flags=%v", inputs.IntegerOperand, flags64)
	}
	got64, flags64 := d64.ConvertToInt64(RoundNearestEven)
	if got64 != inputs.IntegerOperand || flags64 != 0 {
		tb.Fatalf("Decimal64 benchmark integer round trip = (%d, %v), want (%d, 0)", got64, flags64, inputs.IntegerOperand)
	}

	d128 := NewDecimal128FromInt64(inputs.IntegerOperand)
	got128, flags128 := d128.ConvertToInt64(RoundNearestEven)
	if got128 != inputs.IntegerOperand || flags128 != 0 {
		tb.Fatalf("Decimal128 benchmark integer round trip = (%d, %v), want (%d, 0)", got128, flags128, inputs.IntegerOperand)
	}
}

func requireFiniteScaleBOperands(tb testing.TB, inputs benchmarkInputs) {
	tb.Helper()

	if got, flags := exactBenchmarkDecimal32(tb, inputs.Decimal32.X).ScaleB(inputs.ScaleExponent); flags != 0 || !got.IsFinite() {
		tb.Fatalf("Decimal32 benchmark scaleB input raised flags or became non-finite: value=%v flags=%v", got, flags)
	}
	if got, flags := exactBenchmarkDecimal64(tb, inputs.Decimal64.X).ScaleB(inputs.ScaleExponent); flags != 0 || !got.IsFinite() {
		tb.Fatalf("Decimal64 benchmark scaleB input raised flags or became non-finite: value=%v flags=%v", got, flags)
	}
	if got, flags := exactBenchmarkDecimal128(tb, inputs.Decimal128.X).ScaleB(inputs.ScaleExponent); flags != 0 || !got.IsFinite() {
		tb.Fatalf("Decimal128 benchmark scaleB input raised flags or became non-finite: value=%v flags=%v", got, flags)
	}
}

// requireNonNegativeSqrtOperands pins the sqrt-benchmark precondition: the
// sqrt rows reuse the X operands, so a negative X would silently turn every
// sqrt benchmark into a NaN/invalid path instead of a real square root. The
// check inspects the parsed value's sign and NaN class, not the raw text, so
// a disguised spelling (e.g. " -1") cannot slip past it.
func requireNonNegativeSqrtOperands(tb testing.TB, inputs benchmarkInputs) {
	tb.Helper()
	if x := exactBenchmarkDecimal32(tb, inputs.Decimal32.X); x.IsNaN() || x.IsSignMinus() {
		tb.Fatalf("benchmark input decimal32 x = %q parses negative or NaN; sqrt benchmarks reuse x and require a non-negative operand", inputs.Decimal32.X)
	}
	if x := exactBenchmarkDecimal64(tb, inputs.Decimal64.X); x.IsNaN() || x.IsSignMinus() {
		tb.Fatalf("benchmark input decimal64 x = %q parses negative or NaN; sqrt benchmarks reuse x and require a non-negative operand", inputs.Decimal64.X)
	}
	if x := exactBenchmarkDecimal128(tb, inputs.Decimal128.X); x.IsNaN() || x.IsSignMinus() {
		tb.Fatalf("benchmark input decimal128 x = %q parses negative or NaN; sqrt benchmarks reuse x and require a non-negative operand", inputs.Decimal128.X)
	}
}

func requireNonZeroDivisionOperands(tb testing.TB, inputs benchmarkInputs) {
	tb.Helper()
	for _, operand := range []struct {
		name  string
		value Decimal32BID
	}{
		{"decimal32.x", exactBenchmarkDecimal32(tb, inputs.Decimal32.X)},
		{"decimal32.y", exactBenchmarkDecimal32(tb, inputs.Decimal32.Y)},
	} {
		if operand.value.IsZero() {
			tb.Fatalf("benchmark input %s must be non-zero for division/remainder rows", operand.name)
		}
	}
	for _, operand := range []struct {
		name  string
		value Decimal64BID
	}{
		{"decimal64.x", exactBenchmarkDecimal64(tb, inputs.Decimal64.X)},
		{"decimal64.y", exactBenchmarkDecimal64(tb, inputs.Decimal64.Y)},
	} {
		if operand.value.IsZero() {
			tb.Fatalf("benchmark input %s must be non-zero for division/remainder and mixed rows", operand.name)
		}
	}
	for _, operand := range []struct {
		name  string
		value Decimal128BID
	}{
		{"decimal128.x", exactBenchmarkDecimal128(tb, inputs.Decimal128.X)},
		{"decimal128.y", exactBenchmarkDecimal128(tb, inputs.Decimal128.Y)},
	} {
		if operand.value.IsZero() {
			tb.Fatalf("benchmark input %s must be non-zero for division/remainder and mixed rows", operand.name)
		}
	}
}
