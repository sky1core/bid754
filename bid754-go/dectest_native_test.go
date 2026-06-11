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

func TestShouldAddClampedFlagDetectsZeroClampBoundary(t *testing.T) {
	tc := decTestCase{
		Operation:   "add",
		Precision:   16,
		MinExponent: -383,
		Clamp:       1,
	}

	if !shouldAddClampedFlag(tc, "0E-398", 0) {
		t.Fatal("expected zero result at clamp boundary to add clamped flag")
	}
}

func TestShouldAddClampedFlagDetectsFiniteClampWithoutRounding(t *testing.T) {
	tc := decTestCase{
		Operation:   "add",
		Precision:   16,
		MaxExponent: 384,
		Clamp:       1,
	}

	if !shouldAddClampedFlag(tc, "1.2300E+374", 0) {
		t.Fatal("expected trailing-zero finite result beyond adjusted max exponent to add clamped flag")
	}

	if shouldAddClampedFlag(tc, "1.2300E+374", FlagRounded) {
		t.Fatal("expected rounded result to not add heuristic clamped flag")
	}
}

func TestShouldAddClampedFlagDetectsDecimal64BoundaryClampCases(t *testing.T) {
	tc := decTestCase{
		Operation:   "add",
		Precision:   16,
		MaxExponent: 384,
		Clamp:       1,
	}

	testCases := []struct {
		name   string
		result string
	}{
		{name: "ddadd380", result: "2.000000000000000E+384"},
		{name: "ddadd381", result: "2.00000000000E+380"},
		{name: "ddadd382", result: "2.0000000E+376"},
		{name: "ddadd383", result: "2.000E+372"},
		{name: "dddiv274", result: "9.000000000000000E+384"},
		{name: "dddiv275", result: "9.900000000000000E+384"},
		{name: "dddiv276", result: "9.990000000000000E+384"},
		{name: "dddiv277", result: "9.999999999999900E+384"},
	}

	for _, tcResult := range testCases {
		t.Run(tcResult.name, func(t *testing.T) {
			if !shouldAddClampedFlag(tc, tcResult.result, 0) {
				t.Fatalf("expected %s result %q to add clamped flag", tcResult.name, tcResult.result)
			}
		})
	}
}

func TestShouldAddClampedFlagDoesNotHeuristicallyClampReadOperations(t *testing.T) {
	tc := decTestCase{
		Operation:   "toSci",
		Precision:   7,
		MaxExponent: 96,
		Clamp:       1,
	}

	if shouldAddClampedFlag(tc, "1.0E+91", 0) {
		t.Fatal("expected read operations to rely on native status instead of arithmetic clamped heuristic")
	}
}

func TestShouldAddInvalidOperationFlagNormalizesOperationName(t *testing.T) {
	tc := decTestCase{
		Operation: "Divide",
		Operands:  []string{"0", "0"},
	}

	if !shouldAddInvalidOperationFlag(tc) {
		t.Fatal("expected divide alias with 0/0 operands to add invalid operation")
	}
}

func TestApplyDecTestFlagHeuristicsAddsExpectedFlags(t *testing.T) {
	tc := decTestCase{
		Operation:   "divide",
		Operands:    []string{"0", "0"},
		Precision:   16,
		MinExponent: -383,
		Clamp:       1,
	}

	flags := applyDecTestFlagHeuristics(tc, "0E-398", 0)

	expected := FlagClamped | FlagInvalidOperation
	if flags != expected {
		t.Fatalf("expected %s, got %s", expected.String(), flags.String())
	}
}

func TestApplyDecTestFlagHeuristicsSuppressesToIntegralSubnormal(t *testing.T) {
	flags := applyDecTestFlagHeuristics(decTestCase{Operation: "tointegralx"}, "0", FlagSubnormal|FlagInexact|FlagRounded)

	expected := FlagInexact | FlagRounded
	if flags != expected {
		t.Fatalf("expected %s, got %s", expected.String(), flags.String())
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
				Operation: "compare",
				Operands:  []string{"70E-1", "7"},
			},
			testType: "decimal64",
			result:   "0",
		},
		{
			name: "decimal64 comparesig quiet nan signals",
			tc: decTestCase{
				Operation: "compareSig",
				Operands:  []string{"NaN8", "999"},
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
