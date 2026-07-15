package bidgo

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
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
	data, err := os.ReadFile("../../testdata/benchmark_inputs.json")
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

func exactBenchmarkDecimal32(tb testing.TB, input string) uint32 {
	tb.Helper()
	value, err := parseExactBenchmarkDecimal32(input)
	if err != nil {
		tb.Fatal(err)
	}
	return value
}

func parseExactBenchmarkDecimal32(input string) (uint32, error) {
	value, flags := Bid32FromStringRaw(input, 0)
	if flags != 0 || Bid32IsFinite(value) == 0 {
		return 0, fmt.Errorf("Decimal32 benchmark input %q is not finite and exact: flags=%#x bits=%#x", input, flags, value)
	}
	if err := requireBenchmarkCohortPreserved("Decimal32", input, Bid32ToStringRaw(value)); err != nil {
		return 0, err
	}
	return value, nil
}

func exactBenchmarkDecimal64(tb testing.TB, input string) uint64 {
	tb.Helper()
	value, err := parseExactBenchmarkDecimal64(input)
	if err != nil {
		tb.Fatal(err)
	}
	return value
}

func parseExactBenchmarkDecimal64(input string) (uint64, error) {
	value, flags := Bid64FromString(input, 0)
	if flags != 0 || Bid64IsFinite(value) == 0 {
		return 0, fmt.Errorf("Decimal64 benchmark input %q is not finite and exact: flags=%#x bits=%#x", input, flags, value)
	}
	if err := requireBenchmarkCohortPreserved("Decimal64", input, Bid64ToString(value)); err != nil {
		return 0, err
	}
	return value, nil
}

func exactBenchmarkDecimal128(tb testing.TB, input string) BID_UINT128 {
	tb.Helper()
	value, err := parseExactBenchmarkDecimal128(input)
	if err != nil {
		tb.Fatal(err)
	}
	return value
}

func parseExactBenchmarkDecimal128(input string) (BID_UINT128, error) {
	value, flags := Bid128FromString(input, 0)
	if flags != 0 || Bid128IsFinite(value) == 0 {
		return BID_UINT128{}, fmt.Errorf("Decimal128 benchmark input %q is not finite and exact: flags=%#x bits=%#x/%#x", input, flags, value.hi, value.lo)
	}
	if err := requireBenchmarkCohortPreserved("Decimal128", input, Bid128ToString(value)); err != nil {
		return BID_UINT128{}, err
	}
	return value, nil
}

func requireBenchmarkCohortPreserved(width, input, actual string) error {
	want, err := benchmarkInputCohort(input)
	if err != nil {
		return fmt.Errorf("%s benchmark input %q has no exact decimal cohort: %w", width, input, err)
	}
	if actual != want {
		return fmt.Errorf("%s benchmark input %q does not preserve the requested cohort: got %s, want %s", width, input, actual, want)
	}
	return nil
}

// benchmarkInputCohort renders a finite benchmark literal as the exact
// coefficient/exponent pair requested by its spelling. Trailing coefficient
// zeros are intentionally retained because they select a different cohort.
func benchmarkInputCohort(input string) (string, error) {
	s := strings.TrimLeft(input, " \t")
	if s == "" {
		return "", fmt.Errorf("empty decimal literal")
	}

	sign := byte('+')
	position := 0
	if s[position] == '+' || s[position] == '-' {
		sign = s[position]
		position++
		if position == len(s) {
			return "", fmt.Errorf("missing coefficient")
		}
	}

	coefficient := make([]byte, 0, len(s)-position)
	fractionalDigits := int64(0)
	radixSeen := false
	digitSeen := false
	exponentPosition := len(s)
	for ; position < len(s); position++ {
		character := s[position]
		switch {
		case character >= '0' && character <= '9':
			coefficient = append(coefficient, character)
			digitSeen = true
			if radixSeen {
				fractionalDigits++
			}
		case character == '.' && !radixSeen:
			radixSeen = true
		case (character == 'e' || character == 'E') && digitSeen:
			exponentPosition = position
			position = len(s)
		default:
			return "", fmt.Errorf("invalid character %q", character)
		}
	}
	if !digitSeen {
		return "", fmt.Errorf("missing coefficient digits")
	}

	explicitExponent := int64(0)
	if exponentPosition < len(s) {
		exponentText := s[exponentPosition+1:]
		if exponentText == "" {
			return "", fmt.Errorf("missing exponent digits")
		}
		digitStart := 0
		if exponentText[0] == '+' || exponentText[0] == '-' {
			digitStart = 1
		}
		if digitStart == len(exponentText) {
			return "", fmt.Errorf("missing exponent digits")
		}
		for _, character := range exponentText[digitStart:] {
			if character < '0' || character > '9' {
				return "", fmt.Errorf("invalid exponent character %q", character)
			}
		}
		parsed, err := strconv.ParseInt(exponentText, 10, 64)
		if err != nil {
			return "", fmt.Errorf("exponent out of range: %w", err)
		}
		explicitExponent = parsed
	}

	const minInt64 = -1 << 63
	if explicitExponent < minInt64+fractionalDigits {
		return "", fmt.Errorf("cohort exponent out of range")
	}
	exponent := explicitExponent - fractionalDigits
	canonicalCoefficient := strings.TrimLeft(string(coefficient), "0")
	if canonicalCoefficient == "" {
		canonicalCoefficient = "0"
	}
	return fmt.Sprintf("%c%sE%+d", sign, canonicalCoefficient, exponent), nil
}

func TestBenchmarkExactOperandsRejectCohortCoercion(t *testing.T) {
	for _, tc := range []struct {
		name  string
		parse func(string) error
		input string
	}{
		{
			name: "decimal32_precision_plus_one_trailing_zero",
			parse: func(input string) error {
				_, err := parseExactBenchmarkDecimal32(input)
				return err
			},
			input: "1000000.0",
		},
		{
			name: "decimal64_precision_plus_one_trailing_zero",
			parse: func(input string) error {
				_, err := parseExactBenchmarkDecimal64(input)
				return err
			},
			input: "1000000000000000.0",
		},
		{
			name: "decimal128_precision_plus_one_trailing_zero",
			parse: func(input string) error {
				_, err := parseExactBenchmarkDecimal128(input)
				return err
			},
			input: "1000000000000000000000000000000000.0",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.parse(tc.input); err == nil {
				t.Fatalf("benchmark input %q accepted after exact numeric parsing silently changed its cohort", tc.input)
			}
		})
	}
}

func TestBenchmarkExactOperandsAcceptCohortBoundaries(t *testing.T) {
	for _, tc := range []struct {
		name  string
		parse func(string) error
		input string
	}{
		{
			name: "decimal32_minimum_quantum",
			parse: func(input string) error {
				_, err := parseExactBenchmarkDecimal32(input)
				return err
			},
			input: "1E-101",
		},
		{
			name: "decimal32_maximum_quantum_and_precision",
			parse: func(input string) error {
				_, err := parseExactBenchmarkDecimal32(input)
				return err
			},
			input: "9999999E+90",
		},
		{
			name: "decimal64_minimum_quantum",
			parse: func(input string) error {
				_, err := parseExactBenchmarkDecimal64(input)
				return err
			},
			input: "1E-398",
		},
		{
			name: "decimal64_maximum_quantum_and_precision",
			parse: func(input string) error {
				_, err := parseExactBenchmarkDecimal64(input)
				return err
			},
			input: "9999999999999999E+369",
		},
		{
			name: "decimal128_minimum_quantum",
			parse: func(input string) error {
				_, err := parseExactBenchmarkDecimal128(input)
				return err
			},
			input: "1E-6176",
		},
		{
			name: "decimal128_maximum_quantum_and_precision",
			parse: func(input string) error {
				_, err := parseExactBenchmarkDecimal128(input)
				return err
			},
			input: "9999999999999999999999999999999999E+6111",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.parse(tc.input); err != nil {
				t.Fatalf("benchmark input %q rejected at a valid cohort boundary: %v", tc.input, err)
			}
		})
	}
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

	d32, flags32 := Bid32FromInt64(inputs.IntegerOperand, 0)
	if flags32 != 0 {
		tb.Fatalf("benchmark integer_operand %d is not exact as Decimal32: flags=%#x", inputs.IntegerOperand, flags32)
	}
	got32, flags32 := Bid32ToInt64Rnint(d32)
	if got32 != inputs.IntegerOperand || flags32 != 0 {
		tb.Fatalf("Decimal32 benchmark integer round trip = (%d, %#x), want (%d, 0)", got32, flags32, inputs.IntegerOperand)
	}

	d64, flags64 := Bid64FromInt64(inputs.IntegerOperand, 0)
	if flags64 != 0 {
		tb.Fatalf("benchmark integer_operand %d is not exact as Decimal64: flags=%#x", inputs.IntegerOperand, flags64)
	}
	got64, flags64 := Bid64ToInt64Rnint(d64)
	if got64 != inputs.IntegerOperand || flags64 != 0 {
		tb.Fatalf("Decimal64 benchmark integer round trip = (%d, %#x), want (%d, 0)", got64, flags64, inputs.IntegerOperand)
	}

	d128 := Bid128FromInt64(inputs.IntegerOperand)
	got128, flags128 := Bid128ToInt64Rnint(d128)
	if got128 != inputs.IntegerOperand || flags128 != 0 {
		tb.Fatalf("Decimal128 benchmark integer round trip = (%d, %#x), want (%d, 0)", got128, flags128, inputs.IntegerOperand)
	}
}

func requireFiniteScaleBOperands(tb testing.TB, inputs benchmarkInputs) {
	tb.Helper()

	if got, flags := Bid32Scalbn(exactBenchmarkDecimal32(tb, inputs.Decimal32.X), inputs.ScaleExponent, 0); flags != 0 || Bid32IsFinite(got) == 0 {
		tb.Fatalf("Decimal32 benchmark scaleB input raised flags or became non-finite: bits=%#x flags=%#x", got, flags)
	}
	if got, flags := Bid64Scalbn(exactBenchmarkDecimal64(tb, inputs.Decimal64.X), inputs.ScaleExponent, 0); flags != 0 || Bid64IsFinite(got) == 0 {
		tb.Fatalf("Decimal64 benchmark scaleB input raised flags or became non-finite: bits=%#x flags=%#x", got, flags)
	}
	x128 := exactBenchmarkDecimal128(tb, inputs.Decimal128.X)
	var flags128 uint32
	got128 := Bid128Scalbn(x128, inputs.ScaleExponent, 0, &flags128)
	if flags128 != 0 || Bid128IsFinite(got128) == 0 {
		tb.Fatalf("Decimal128 benchmark scaleB input raised flags or became non-finite: bits=%#x/%#x flags=%#x", got128.hi, got128.lo, flags128)
	}
}

// requireNonNegativeSqrtOperands pins the sqrt-benchmark precondition: the
// sqrt rows reuse the X operands, so a negative X would silently turn every
// sqrt benchmark into a NaN/invalid path instead of a real square root. The
// check inspects the parsed value's sign and NaN class, not the raw text, so
// a disguised spelling (e.g. " -1") cannot slip past it.
func requireNonNegativeSqrtOperands(tb testing.TB, inputs benchmarkInputs) {
	tb.Helper()
	if x := exactBenchmarkDecimal32(tb, inputs.Decimal32.X); Bid32IsNaN(x) || Bid32IsSigned(x) != 0 {
		tb.Fatalf("benchmark input decimal32 x = %q parses negative or NaN; sqrt benchmarks reuse x and require a non-negative operand", inputs.Decimal32.X)
	}
	if x := exactBenchmarkDecimal64(tb, inputs.Decimal64.X); Bid64IsNaN(x) != 0 || Bid64IsSigned(x) != 0 {
		tb.Fatalf("benchmark input decimal64 x = %q parses negative or NaN; sqrt benchmarks reuse x and require a non-negative operand", inputs.Decimal64.X)
	}
	if x := exactBenchmarkDecimal128(tb, inputs.Decimal128.X); Bid128IsNaN(x) != 0 || Bid128IsSigned(x) != 0 {
		tb.Fatalf("benchmark input decimal128 x = %q parses negative or NaN; sqrt benchmarks reuse x and require a non-negative operand", inputs.Decimal128.X)
	}
}

func requireNonZeroDivisionOperands(tb testing.TB, inputs benchmarkInputs) {
	tb.Helper()
	if x := exactBenchmarkDecimal32(tb, inputs.Decimal32.X); Bid32IsZero(x) {
		tb.Fatal("benchmark input decimal32.x must be non-zero for remainder rows")
	}
	if y := exactBenchmarkDecimal32(tb, inputs.Decimal32.Y); Bid32IsZero(y) {
		tb.Fatal("benchmark input decimal32.y must be non-zero for division rows")
	}
	if x := exactBenchmarkDecimal64(tb, inputs.Decimal64.X); Bid64IsZero(x) != 0 {
		tb.Fatal("benchmark input decimal64.x must be non-zero for remainder and mixed rows")
	}
	if y := exactBenchmarkDecimal64(tb, inputs.Decimal64.Y); Bid64IsZero(y) != 0 {
		tb.Fatal("benchmark input decimal64.y must be non-zero for division and mixed rows")
	}
	if x := exactBenchmarkDecimal128(tb, inputs.Decimal128.X); Bid128IsZero(x) != 0 {
		tb.Fatal("benchmark input decimal128.x must be non-zero for remainder and mixed rows")
	}
	if y := exactBenchmarkDecimal128(tb, inputs.Decimal128.Y); Bid128IsZero(y) != 0 {
		tb.Fatal("benchmark input decimal128.y must be non-zero for division and mixed rows")
	}
}
