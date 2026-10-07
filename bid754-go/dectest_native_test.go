//go:build cgo && bid754_native

package bid754

import "testing"

func TestStatusToExceptionFlagsMapsNativeStatusBits(t *testing.T) {
	status := nativeStatusFlagBits.ConversionSyntax | nativeStatusFlagBits.Rounded | nativeStatusFlagBits.Clamped

	flags := statusToExceptionFlags(status)

	expected := FlagInvalidOperation | FlagRounded | FlagClamped
	if flags != expected {
		t.Fatalf("expected %s, got %s", expected.String(), flags.String())
	}
}

func TestNativeDectestConditionComparisonRejectsMutation(t *testing.T) {
	tc := decTestCase{Operation: "divide", Operands: []string{"00.00", "0.000"}, Result: "NaN", Flags: []string{"Division_undefined"}}
	exec := func(decTestCase, string) (decTestExecResult, error) {
		return decTestExecResult{Result: "NaN"}, nil
	}
	if err := runGeneratedDectestCase(tc, "decimal64", 2, exec, generatedDectestCompareDecimalResult, generatedDectestFlagCheckNative); err == nil {
		t.Fatal("missing native division undefined condition passed")
	}
	exec = func(decTestCase, string) (decTestExecResult, error) {
		return decTestExecResult{Result: "NaN", Conditions: decTestInvalidOperation}, nil
	}
	if err := runGeneratedDectestCase(tc, "decimal64", 2, exec, generatedDectestCompareDecimalResult, generatedDectestFlagCheckNative); err == nil {
		t.Fatal("different GDA invalid condition passed")
	}
	exec = func(decTestCase, string) (decTestExecResult, error) {
		return decTestExecResult{Result: "NaN", Conditions: decTestDivisionUndefined}, nil
	}
	if err := runGeneratedDectestCase(tc, "decimal64", 2, exec, generatedDectestCompareDecimalResult, generatedDectestFlagCheckNative); err != nil {
		t.Fatalf("matching native condition rejected: %v", err)
	}
	for _, condition := range []string{"Conversion_syntax", "Division_impossible", "Insufficient_storage", "Invalid_operation"} {
		t.Run(condition, func(t *testing.T) {
			mutated := tc
			mutated.Flags = []string{condition}
			if err := runGeneratedDectestCase(mutated, "decimal64", 2, exec, generatedDectestCompareDecimalResult, generatedDectestFlagCheckNative); err == nil {
				t.Fatal("different GDA invalid condition passed")
			}
		})
	}
}

func TestNativeDectestGetsDivisionUndefinedFromEngine(t *testing.T) {
	tc := decTestCase{Operation: "divide", Operands: []string{"00.00", "0.000"}, Precision: 16, MaxExponent: 384, MinExponent: -383, Clamp: 1}
	got, err := executeDecTestOperation(tc, "decimal64")
	if err != nil {
		t.Fatal(err)
	}
	if got.Conditions != decTestDivisionUndefined {
		t.Fatalf("native conditions = %d, want Division_undefined", got.Conditions)
	}
	if got.Flags != FlagInvalidOperation {
		t.Fatalf("native IEEE flags = %s, want invalid", got.Flags)
	}
}

func TestNativeDectestGetsClampedFromEngine(t *testing.T) {
	tc := decTestCase{Operation: "add", Operands: []string{"1E+384", "1E+384"}, Precision: 16, MaxExponent: 384, MinExponent: -383, Clamp: 1}
	got, err := executeDecTestOperation(tc, "decimal64")
	if err != nil {
		t.Fatal(err)
	}
	if got.Conditions != decTestClamped {
		t.Fatalf("native conditions = %d, want Clamped", got.Conditions)
	}
}

func TestExecuteDecTestOperationNativeSupportsCompareFamily(t *testing.T) {
	testCases := []struct {
		name     string
		tc       decTestCase
		testType string
		result   string
		flags    ExceptionFlags
	}{
		{
			name: "decimal64 compare finite",
			tc: decTestCase{
				Operation:    "compare",
				Operands:     []string{"70E-1", "7"},
				Precision:    16,
				RoundingMode: "half_even",
				MaxExponent:  384,
				MinExponent:  -383,
				Clamp:        1,
			},
			testType: "decimal64",
			result:   "0",
		},
		{
			name: "decimal64 comparesig quiet nan signals",
			tc: decTestCase{
				Operation:    "compareSig",
				Operands:     []string{"NaN8", "999"},
				Precision:    16,
				RoundingMode: "half_even",
				MaxExponent:  384,
				MinExponent:  -383,
				Clamp:        1,
			},
			testType: "decimal64",
			result:   "NaN8",
			flags:    FlagInvalidOperation,
		},
		{
			name: "general compare huge exponent",
			tc: decTestCase{
				Operation:    "compare",
				Operands:     []string{"9.99999999E+999999999", "-9.99999999E+999999999"},
				Precision:    9,
				RoundingMode: "half_up",
				MaxExponent:  999999999,
				MinExponent:  -999999999,
			},
			testType: "general",
			result:   "1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := executeDecTestOperation(tc.tc, tc.testType)
			if err != nil {
				t.Fatalf("executeDecTestOperation returned error: %v", err)
			}
			if !compareDecimalResults(tc.result, got.Result) {
				t.Fatalf("expected result %q, got %q", tc.result, got.Result)
			}
			if got.Flags != tc.flags {
				t.Fatalf("expected flags %s, got %s", tc.flags.String(), got.Flags.String())
			}
		})
	}
}

func TestExecuteDecTestOperationNativeSupportsQuantize(t *testing.T) {
	testCases := []struct {
		name     string
		tc       decTestCase
		testType string
		result   string
		flags    ExceptionFlags
	}{
		{
			name: "decimal64 quantize exact",
			tc: decTestCase{
				Operation:    "quantize",
				Operands:     []string{"2.17", "0.001"},
				Precision:    16,
				RoundingMode: "half_even",
				MaxExponent:  384,
				MinExponent:  -383,
				Clamp:        1,
			},
			testType: "decimal64",
			result:   "2.170",
		},
		{
			name: "general quantize rounded",
			tc: decTestCase{
				Operation:    "quantize",
				Operands:     []string{"2.17", "0.1"},
				Precision:    9,
				RoundingMode: "half_up",
				MaxExponent:  999,
				MinExponent:  -999,
			},
			testType: "general",
			result:   "2.2",
			flags:    FlagInexact | FlagRounded,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := executeDecTestOperation(tc.tc, tc.testType)
			if err != nil {
				t.Fatalf("executeDecTestOperation returned error: %v", err)
			}
			if !compareDecimalResults(tc.result, got.Result) {
				t.Fatalf("expected result %q, got %q", tc.result, got.Result)
			}
			if got.Flags != tc.flags {
				t.Fatalf("expected flags %s, got %s", tc.flags.String(), got.Flags.String())
			}
		})
	}
}

func TestExecuteDecTestReadOperationNativeSupportsToIntegralFamily(t *testing.T) {
	testCases := []struct {
		name     string
		tc       decTestCase
		testType string
		result   string
		flags    ExceptionFlags
	}{
		{
			name: "decimal64 tointegral suppresses inexact",
			tc: decTestCase{
				Operation:    "tointegral",
				Operands:     []string{"101.5"},
				RoundingMode: "half_up",
			},
			testType: "decimal64",
			result:   "102",
		},
		{
			name: "decimal64 tointegralx preserves rounded and inexact",
			tc: decTestCase{
				Operation:    "tointegralx",
				Operands:     []string{"1.0"},
				RoundingMode: "half_even",
			},
			testType: "decimal64",
			result:   "1",
			flags:    FlagRounded,
		},
		{
			name: "general tointegralx supports non-ieee rounding aliases",
			tc: decTestCase{
				Operation:    "tointegralx",
				Operands:     []string{"56.5"},
				RoundingMode: "half_down",
				Precision:    9,
				MaxExponent:  999,
				MinExponent:  -999,
			},
			testType: "general",
			result:   "56",
			flags:    FlagInexact | FlagRounded,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := executeDecTestReadOperation(tc.tc, tc.testType)
			if err != nil {
				t.Fatalf("executeDecTestReadOperation returned error: %v", err)
			}
			if !compareDecimalResults(tc.result, got.Result) {
				t.Fatalf("expected result %q, got %q", tc.result, got.Result)
			}
			if got.Flags != tc.flags {
				t.Fatalf("expected flags %s, got %s", tc.flags.String(), got.Flags.String())
			}
		})
	}
}
