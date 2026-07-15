package bid754

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestDecTestFailureErrorNoFailures(t *testing.T) {
	err := decTestFailureError([]decTestSuiteTotals{
		{Name: "Decimal32", Passed: 10, Failed: 0, Skipped: 0},
		{Name: "Decimal64", Passed: 20, Failed: 0, Skipped: 1},
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestDecTestFailureErrorSummarizesFailedSuites(t *testing.T) {
	err := decTestFailureError([]decTestSuiteTotals{
		{Name: "Decimal32", Passed: 10, Failed: 0, Skipped: 0},
		{Name: "Decimal64", Passed: 20, Failed: 3, Skipped: 1},
		{Name: "General", Passed: 5, Failed: 7, Skipped: 2},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	msg := err.Error()
	for _, part := range []string{"10 total", "Decimal64=3", "General=7"} {
		if !strings.Contains(msg, part) {
			t.Fatalf("expected error message %q to contain %q", msg, part)
		}
	}
}

func TestCompareDecimalResultsNormalizesEquivalentRepresentations(t *testing.T) {
	testCases := []struct {
		name     string
		expected string
		actual   string
	}{
		{name: "quoted literal", expected: "'1.23'", actual: "1.23"},
		{name: "infinity alias", expected: "+Inf", actual: "Infinity"},
		{name: "nan case", expected: "nan", actual: "NaN"},
		{name: "negative nan", expected: "-nan", actual: "-NaN"},
		{name: "quiet nan payload alias", expected: "qNaN007", actual: "NaN7"},
		{name: "signaling nan payload", expected: "-sNaN0009", actual: "-sNaN9"},
		{name: "finite encoded with exponent", expected: "7.50", actual: "+750E-2"},
		{name: "signed zero encoded with exponent", expected: "-0.0000", actual: "-0E-4"},
		{name: "decimal64 max alternate exponent", expected: "9.999999999999999E+384", actual: "+9999999999999999E+369"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if !compareDecimalResults(tc.expected, tc.actual) {
				t.Fatalf("expected %q and %q to compare equal", tc.expected, tc.actual)
			}
		})
	}
}

func TestCompareDecimalResultsDistinguishesQuietAndSignalingNaN(t *testing.T) {
	if compareDecimalResults("sNaN7", "NaN7") {
		t.Fatal("expected signaling NaN and quiet NaN to compare different")
	}
}

func TestCompareDecimalResultsRejectsApproximateNumericMatches(t *testing.T) {
	testCases := []struct {
		name     string
		expected string
		actual   string
	}{
		{name: "different last digit", expected: "1.23456789012345", actual: "1.23456789012346"},
		{name: "scientific vs rounded integer", expected: "4.28135971E+11", actual: "428135971041"},
		{name: "zero sign differs", expected: "-0.0000", actual: "+0E-4"},
		{name: "nan payload differs", expected: "NaN7", actual: "NaN"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if compareDecimalResults(tc.expected, tc.actual) {
				t.Fatalf("expected %q and %q to compare different", tc.expected, tc.actual)
			}
		})
	}
}

func TestShouldSkipDecTestCaseAllowsCopyFamilyNaNPayloadEdges(t *testing.T) {
	testCases := []struct {
		name string
		tc   decTestCase
	}{
		{
			name: "copy payload result",
			tc: decTestCase{
				Operation: "copy",
				Result:    "NaN7",
				Precision: 16,
			},
		},
		{
			name: "copy sign finite result with rhs payload",
			tc: decTestCase{
				Operation: "copySign",
				Operands:  []string{"-720", "+NaN8"},
				Result:    "720",
				Precision: 16,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if shouldSkipDecTestCase(tc.tc, "decimal64") {
				t.Fatal("expected payload case to run")
			}
		})
	}
}

func TestCompareDecTestFlagsKeepsNativeGDAConditionsExact(t *testing.T) {
	expected := []string{"Rounded", "Subnormal", "Clamped", "Division_by_zero", "inexact"}
	actual := FlagInexact | FlagDivisionByZero

	if compareDecTestFlags(expected, actual) {
		t.Fatalf("expected native comparison to retain GDA conditions in %v", expected)
	}
	if !compareDecTestFlags(expected, actual|FlagRounded|FlagSubnormal|FlagClamped) {
		t.Fatalf("expected exact GDA condition match for %v", expected)
	}
}

func TestCompareDecTestBIDFiveFlagsProjectsOutGDAConditions(t *testing.T) {
	expected := []string{"Rounded", "Subnormal", "Clamped", "Division_by_zero", "inexact"}
	actual := FlagInexact | FlagDivisionByZero

	if !compareDecTestBIDFiveFlags(expected, actual) {
		t.Fatalf("expected Intel five-flag projection of %v to match %s", expected, actual.String())
	}
	if !compareDecTestBIDFiveFlags(expected, actual|FlagRounded|FlagSubnormal|FlagClamped) {
		t.Fatal("expected GDA-only actual conditions to remain outside the BID comparison mask")
	}
}

func TestCompareDecTestFlagsRejectsUnsupportedOrMismatchedFlags(t *testing.T) {
	if compareDecTestFlags([]string{"Rounded", "ImpossibleFlag"}, FlagRounded) {
		t.Fatal("expected unsupported decTest flag to fail comparison")
	}

	if compareDecTestFlags([]string{"Rounded"}, FlagRounded|FlagInexact) {
		t.Fatal("expected mismatched flag set to fail comparison")
	}
}

func TestCompareDecTestFlagsTreatsDivisionUndefinedAsInvalidOperation(t *testing.T) {
	if !compareDecTestFlags([]string{"Division_undefined"}, FlagInvalidOperation) {
		t.Fatal("expected Division_undefined to map to invalid operation")
	}
}

func TestExecutePortableCompareOperationFiniteCases(t *testing.T) {
	testCases := []struct {
		name   string
		tc     decTestCase
		result string
		flags  ExceptionFlags
	}{
		{
			name: "equal with different exponents",
			tc: decTestCase{
				Operation: "compare",
				Operands:  []string{"70E-1", "7.0"},
			},
			result: "0",
		},
		{
			name: "general huge exponent comparison",
			tc: decTestCase{
				Operation: "compare",
				Operands:  []string{"9.99999999E+999999999", "-9.99999999E+999999999"},
			},
			result: "1",
		},
		{
			name: "quiet nan compare propagates lhs",
			tc: decTestCase{
				Operation: "compare",
				Operands:  []string{"-NaN67", "NaN5"},
			},
			result: "-NaN67",
		},
		{
			name: "compareSig quiet nan signals",
			tc: decTestCase{
				Operation: "compareSig",
				Operands:  []string{"NaN8", "999"},
			},
			result: "NaN8",
			flags:  FlagInvalidOperation,
		},
		{
			name: "signaling nan wins propagation",
			tc: decTestCase{
				Operation: "compare",
				Operands:  []string{"NaN85", "sNaN83"},
			},
			result: "NaN83",
			flags:  FlagInvalidOperation,
		},
		{
			name: "conversion syntax maps to invalid",
			tc: decTestCase{
				Operation: "compare",
				Operands:  []string{"10", "#"},
			},
			result: "NaN",
			flags:  FlagInvalidOperation,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := executePortableCompareOperation(tc.tc)
			if err != nil {
				t.Fatalf("executePortableCompareOperation returned error: %v", err)
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

func TestRunDecTestCaseV2SupportsCompareOperations(t *testing.T) {
	testCases := []decTestCase{
		{
			Operation: "compare",
			Operands:  []string{"70E-1", "7"},
			Result:    "0",
			Precision: 16,
		},
		{
			Operation: "compareSig",
			Operands:  []string{"NaN8", "999"},
			Result:    "NaN8",
			Flags:     []string{"Invalid_operation"},
			Precision: 16,
		},
	}

	for _, tc := range testCases {
		if err := runDecTestCaseV2(tc, "decimal64"); err != nil {
			t.Fatalf("runDecTestCaseV2(%s) error: %v", tc.Operation, err)
		}
	}
}

func TestRunDecTestCaseV2QuantizeSupportMatchesBackend(t *testing.T) {
	tc := decTestCase{
		Operation:    "quantize",
		Operands:     []string{"2.17", "0.1"},
		Result:       "2.2",
		Flags:        []string{"Inexact", "Rounded"},
		Precision:    16,
		RoundingMode: "half_even",
		MaxExponent:  384,
		MinExponent:  -383,
		Clamp:        1,
	}

	err := runDecTestCaseV2(tc, "decimal64")
	if NativeBackendEnabled() {
		if err != nil {
			t.Fatalf("expected native quantize support, got %v", err)
		}
		return
	}

	if !errors.Is(err, errUnsupportedDecTestOperation) {
		t.Fatalf("expected portable path to report unsupported quantize, got %v", err)
	}
}

func TestExecuteOperationV2RoutesDecimal128Precision(t *testing.T) {
	got, err := executeOperationV2("add", "12345678901234567", "1", 17)
	if err != nil {
		t.Fatalf("executeOperationV2 returned error: %v", err)
	}
	if got != "12345678901234568" {
		t.Fatalf("executeOperationV2 result = %q, want %q", got, "12345678901234568")
	}
}

func TestLoadGeneratedDectestRunSuitesMatchesSharedSpec(t *testing.T) {
	runSuites, err := loadGeneratedDectestRunSuites()
	if err != nil {
		t.Fatalf("loadGeneratedDectestRunSuites returned error: %v", err)
	}

	spec, err := loadGeneratedTestSpec()
	if err != nil {
		t.Fatalf("loadGeneratedTestSpec returned error: %v", err)
	}

	if len(runSuites) != len(spec.DectestSuites) {
		t.Fatalf("generated run suite count = %d, want %d", len(runSuites), len(spec.DectestSuites))
	}

	for i, suite := range spec.DectestSuites {
		got := runSuites[i]
		if got.Name != suite.Name || got.Pattern != suite.Pattern || got.TestType != suite.TestType {
			t.Fatalf("run suite[%d] = %+v, want name=%q pattern=%q test_type=%q", i, got, suite.Name, suite.Pattern, suite.TestType)
		}
		if !reflect.DeepEqual(got.Files, suite.Files) {
			t.Fatalf("run suite[%d] files = %v, want %v", i, got.Files, suite.Files)
		}
		if !reflect.DeepEqual(got.IgnoredOperations, suite.IgnoredOperations) {
			t.Fatalf("run suite[%d] ignored operations = %v, want %v", i, got.IgnoredOperations, suite.IgnoredOperations)
		}
	}
}

func TestShouldSkipDectestIgnoredOperationNormalizesNames(t *testing.T) {
	ignored := []string{"apply", "compare_sig", "to_integral"}

	for _, operation := range []string{"apply", "compareSig", "tointegral"} {
		if !shouldSkipDectestIgnoredOperation(ignored, operation) {
			t.Fatalf("expected %q to be skipped by ignored operation list %v", operation, ignored)
		}
	}

	if shouldSkipDectestIgnoredOperation(ignored, "quantize") {
		t.Fatalf("did not expect %q to be skipped by ignored operation list %v", "quantize", ignored)
	}
}

func TestShouldSkipDecTestCaseSupportsDecimal128PrecisionBoundary(t *testing.T) {
	if shouldSkipDecTestCase(decTestCase{Precision: 34}, "decimal128") {
		t.Fatal("expected decimal128 precision 34 case to run")
	}
	if !shouldSkipDecTestCase(decTestCase{Precision: 35}, "decimal128") {
		t.Fatal("expected decimal128 precision 35 case to skip")
	}
}

func TestShouldSkipDecTestCaseSkipsDecimal128TaggedLiterals(t *testing.T) {
	tc := decTestCase{
		Precision: 34,
		Operands:  []string{"#20800000000000008000000000000000", "1"},
		Result:    "#22080000000000000000000000000001",
	}
	if !shouldSkipDecTestCase(tc, "decimal128") {
		t.Fatal("expected decimal128 tagged-literal case to skip")
	}
}

func TestDecTestSkipReasonReportsSpecificReasons(t *testing.T) {
	testCases := []struct {
		name     string
		ignored  []string
		tc       decTestCase
		testType string
		want     string
	}{
		{
			name:     "ignored operation",
			ignored:  []string{"apply"},
			tc:       decTestCase{Operation: "apply"},
			testType: "decimal64",
			want:     "ignored_operation_apply",
		},
		{
			name: "tagged literal",
			tc: decTestCase{
				Operands: []string{"#20800000000000008000000000000000", "1"},
				Result:   "#22080000000000000000000000000001",
			},
			testType: "decimal128",
			want:     "tagged_literal",
		},
		{
			name: "tagged to integral",
			tc: decTestCase{
				Operation: "tointegralx",
				Operands:  []string{"1.23E+384"},
				Result:    "#47fd300000000000",
			},
			testType: "decimal64",
			want:     "tagged_to_integral",
		},
		{
			name: "nexttoward nan payload",
			tc: decTestCase{
				Operation: "nexttoward",
				Operands:  []string{"NaN123", "sNaN"},
				Result:    "NaN",
				Flags:     []string{"Invalid_operation"},
			},
			testType: "decimal64",
			want:     "nexttoward_nan_payload_precedence",
		},
		{
			name: "minmax zero tie",
			tc: decTestCase{
				Operation: "max",
				Operands:  []string{"0", "-0"},
				Result:    "0",
			},
			testType: "decimal64",
			want:     "minmax_zero_tie",
		},
		{
			name: "minmax nan payload",
			tc: decTestCase{
				Operation: "min",
				Operands:  []string{"NaN95", "sNaN93"},
				Result:    "NaN93",
				Flags:     []string{"Invalid_operation"},
			},
			testType: "decimal128",
			want:     "minmax_nan_payload_precedence",
		},
		{
			name: "fma unsupported rounding",
			tc: decTestCase{
				Operation:    "fma",
				Operands:     []string{"1", "0", "0E-19"},
				RoundingMode: "up",
			},
			testType: "decimal64",
			want:     "fma_unsupported_rounding",
		},
		{
			name: "fma nan payload precedence divergent",
			tc: decTestCase{
				Operation:    "fma",
				Operands:     []string{"NaN2", "NaN3", "NaN5"},
				Result:       "NaN2",
				RoundingMode: "half_even",
			},
			testType: "decimal64",
			want:     "fma_nan_payload_precedence",
		},
		{
			name: "fma snan beats later quiet nan divergent",
			tc: decTestCase{
				Operation:    "fma",
				Operands:     []string{"1", "NaN16", "sNaN19"},
				Result:       "NaN19",
				Flags:        []string{"Invalid_operation"},
				RoundingMode: "half_even",
			},
			testType: "decimal64",
			want:     "fma_nan_payload_precedence",
		},
		{
			name: "fma quietized sign divergence",
			tc: decTestCase{
				Operation:    "fma",
				Operands:     []string{"-sNaN00", "NaN", "0e+384"},
				Result:       "-NaN",
				Flags:        []string{"Invalid_operation"},
				RoundingMode: "half_even",
			},
			testType: "decimal64",
			want:     "fma_nan_payload_precedence",
		},
		{
			name: "remainder division impossible",
			tc: decTestCase{
				Operation: "remainder",
				Operands:  []string{"1E+6144", "1"},
				Result:    "NaN",
				Flags:     []string{"Division_impossible"},
			},
			testType: "decimal64",
			want:     "remainder_gda_division_impossible_context_semantics",
		},
		{
			name: "remaindernear nan payload",
			tc: decTestCase{
				Operation: "remaindernear",
				Operands:  []string{"NaN3", "sNaN9"},
				Result:    "NaN9",
				Flags:     []string{"Invalid_operation"},
			},
			testType: "decimal64",
			want:     "remaindernear_nan_payload_precedence",
		},
		{
			name: "precision limit",
			tc: decTestCase{
				Operation: "add",
				Precision: 17,
			},
			testType: "decimal64",
			want:     "precision_over_decimal64",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got, ok := decTestSkipReason(tc.ignored, tc.tc, tc.testType)
			if !ok {
				t.Fatalf("decTestSkipReason did not skip case")
			}
			if got != tc.want {
				t.Fatalf("decTestSkipReason = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestDecTestCaseSkipReasonIsEmptyWhenCaseIsNotSkipped(t *testing.T) {
	for _, tc := range []struct {
		name      string
		testType  string
		precision int
	}{
		{name: "decimal32 below limit", testType: "decimal32", precision: 6},
		{name: "decimal32 at limit", testType: "decimal32", precision: 7},
		{name: "decimal64 below limit", testType: "decimal64", precision: 15},
		{name: "decimal64 at limit", testType: "decimal64", precision: 16},
		{name: "decimal128 below limit", testType: "decimal128", precision: 33},
		{name: "decimal128 at limit", testType: "decimal128", precision: 34},
		{name: "general below limit", testType: "general", precision: 33},
		{name: "general at limit", testType: "general", precision: 34},
	} {
		t.Run(tc.name, func(t *testing.T) {
			reason, ok := decTestCaseSkipReason(decTestCase{
				Operation: "add",
				Precision: tc.precision,
			}, tc.testType)
			if ok || reason != "" {
				t.Fatalf("decTestCaseSkipReason precision=%d type=%s = %q/%v, want empty/false", tc.precision, tc.testType, reason, ok)
			}
		})
	}
}

func TestDecTestSkipReasonRunsFMANaNCasesWithMatchingIdentity(t *testing.T) {
	testCases := []struct {
		name string
		tc   decTestCase
	}{
		{
			name: "matching nan identity across both propagation rules",
			tc: decTestCase{
				Operation:    "fma",
				Operands:     []string{"NaN5", "NaN5", "1"},
				RoundingMode: "half_even",
			},
		},
		{
			name: "single nan operand propagates identically",
			tc: decTestCase{
				Operation:    "fma",
				Operands:     []string{"1", "2", "NaN7"},
				RoundingMode: "half_even",
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if reason, ok := decTestSkipReason(nil, tc.tc, "decimal64"); ok {
				t.Fatalf("decTestSkipReason skipped runnable fma NaN case with reason %q", reason)
			}
		})
	}
}

func TestExecutePortableToIntegralOperationPreservesFlagDistinction(t *testing.T) {
	testCases := []struct {
		name   string
		tc     decTestCase
		result string
		flags  ExceptionFlags
	}{
		{
			name: "tointegral suppresses rounded and inexact",
			tc: decTestCase{
				Operation:    "tointegral",
				Operands:     []string{"1.1"},
				RoundingMode: "half_even",
			},
			result: "1",
		},
		{
			name: "tointegralx keeps rounded and inexact",
			tc: decTestCase{
				Operation:    "tointegralx",
				Operands:     []string{"1.1"},
				RoundingMode: "half_even",
			},
			result: "1",
			flags:  FlagInexact | FlagRounded,
		},
		{
			name: "tointegralx exact nonzero integral only sets rounded",
			tc: decTestCase{
				Operation:    "tointegralx",
				Operands:     []string{"1.0"},
				RoundingMode: "half_even",
			},
			result: "1",
			flags:  FlagRounded,
		},
		{
			name: "tointegralx exact zero keeps flags clear",
			tc: decTestCase{
				Operation:    "tointegralx",
				Operands:     []string{"-0.0"},
				RoundingMode: "half_even",
			},
			result: "-0",
		},
		{
			name: "tointegralx signals signaling nan",
			tc: decTestCase{
				Operation: "tointegralx",
				Operands:  []string{"-sNaN080"},
			},
			result: "-NaN80",
			flags:  FlagInvalidOperation,
		},
		{
			name: "half down tie rounds toward zero",
			tc: decTestCase{
				Operation:    "tointegralx",
				Operands:     []string{"56.5"},
				RoundingMode: "half_down",
			},
			result: "56",
			flags:  FlagInexact | FlagRounded,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := executePortableToIntegralOperation(tc.tc)
			if err != nil {
				t.Fatalf("executePortableToIntegralOperation returned error: %v", err)
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

func TestRunDecTestCaseV2SupportsToIntegralFamily(t *testing.T) {
	testCases := []decTestCase{
		{
			Operation:    "tointegral",
			Operands:     []string{"101.5"},
			Result:       "102",
			RoundingMode: "half_up",
			Precision:    16,
		},
		{
			Operation:    "tointegralx",
			Operands:     []string{"1.0"},
			Result:       "1",
			Flags:        []string{"Rounded"},
			RoundingMode: "half_even",
			Precision:    16,
		},
	}

	for _, tc := range testCases {
		if err := runDecTestCaseV2(tc, "decimal64"); err != nil {
			t.Fatalf("runDecTestCaseV2(%s) error: %v", tc.Operation, err)
		}
	}
}

func TestRunDecTestCaseV2SupportsCopyFamily(t *testing.T) {
	testCases := []decTestCase{
		{
			ID:        "copy64",
			Operation: "copy",
			Operands:  []string{"'-123'"},
			Result:    "-123",
		},
		{
			ID:        "copyabs64",
			Operation: "copyAbs",
			Operands:  []string{"'-123'"},
			Result:    "123",
		},
		{
			ID:        "copynegate64",
			Operation: "copyNegate",
			Operands:  []string{"'123'"},
			Result:    "-123",
		},
		{
			ID:        "copysign64",
			Operation: "copySign",
			Operands:  []string{"'123'", "'-1'"},
			Result:    "-123",
		},
		{
			ID:        "copy_payload64",
			Operation: "copy",
			Operands:  []string{"'NaN101'"},
			Result:    "NaN101",
		},
		{
			ID:        "copynegate_payload64",
			Operation: "copyNegate",
			Operands:  []string{"'sNaN13'"},
			Result:    "-sNaN13",
		},
	}

	for _, tc := range testCases {
		if err := runDecTestCaseV2(tc, "decimal64"); err != nil {
			t.Fatalf("runDecTestCaseV2(%s) error: %v", tc.Operation, err)
		}
	}
}

func TestRunDecTestCaseV2SupportsClassOperation(t *testing.T) {
	testCases := []decTestCase{
		{
			ID:        "class_zero64",
			Operation: "class",
			Operands:  []string{"'-0'"},
			Result:    "-Zero",
		},
		{
			ID:        "class_subnormal64",
			Operation: "class",
			Operands:  []string{"'1E-396'"},
			Result:    "+Subnormal",
		},
		{
			ID:        "class_snan64",
			Operation: "class",
			Operands:  []string{"'+sNaN123'"},
			Result:    "sNaN",
		},
	}

	for _, tc := range testCases {
		if err := runDecTestCaseV2(tc, "decimal64"); err != nil {
			t.Fatalf("runDecTestCaseV2(%s) error: %v", tc.ID, err)
		}
	}
}

func TestRunDecTestCaseV2SupportsSameQuantumOperation(t *testing.T) {
	testCases := []decTestCase{
		{
			ID:        "samequantum_true64",
			Operation: "samequantum",
			Operands:  []string{"'7E-3'", "'0E-3'"},
			Result:    "1",
		},
		{
			ID:        "samequantum_false64",
			Operation: "samequantum",
			Operands:  []string{"'7E-3'", "'7'"},
			Result:    "0",
		},
		{
			ID:        "samequantum_nan64",
			Operation: "samequantum",
			Operands:  []string{"'NaN3'", "'sNaN4'"},
			Result:    "1",
		},
	}

	for _, tc := range testCases {
		if err := runDecTestCaseV2(tc, "decimal64"); err != nil {
			t.Fatalf("runDecTestCaseV2(%s) error: %v", tc.ID, err)
		}
	}
}

func TestRunDecTestCaseV2SupportsNextTowardOperation(t *testing.T) {
	testCases := []decTestCase{
		{
			ID:        "nexttoward_up64",
			Operation: "nexttoward",
			Operands:  []string{"'1'", "'10'"},
			Result:    "1.000000000000001",
		},
		{
			ID:        "nexttoward_zero64",
			Operation: "nexttoward",
			Operands:  []string{"'0'", "'10'"},
			Result:    "1E-398",
			Flags:     []string{"Underflow", "Subnormal", "Inexact", "Rounded"},
		},
	}

	for _, tc := range testCases {
		if err := runDecTestCaseV2(tc, "decimal64"); err != nil {
			t.Fatalf("runDecTestCaseV2(%s) error: %v", tc.ID, err)
		}
	}
}

func TestRunDecTestCaseV2SupportsNextPlusMinusOperation(t *testing.T) {
	testCases := []decTestCase{
		{
			ID:        "nextplus64",
			Operation: "nextplus",
			Operands:  []string{"'1'"},
			Result:    "1.000000000000001",
		},
		{
			ID:        "nextminus64",
			Operation: "nextminus",
			Operands:  []string{"'1'"},
			Result:    "0.9999999999999999",
		},
		{
			ID:        "nextplus_snan64",
			Operation: "nextplus",
			Operands:  []string{"'sNaN88'"},
			Result:    "NaN88",
			Flags:     []string{"Invalid_operation"},
		},
		{
			ID:        "nextminus_invalid64",
			Operation: "nextminus",
			Operands:  []string{"#"},
			Result:    "NaN",
			Flags:     []string{"Invalid_operation"},
		},
		{
			ID:        "nextplus128",
			Operation: "nextplus",
			Operands:  []string{"'1'"},
			Result:    "1.000000000000000000000000000000001",
		},
	}

	for _, tc := range testCases {
		testType := "decimal64"
		if tc.ID == "nextplus128" {
			testType = "decimal128"
		}
		if err := runDecTestCaseV2(tc, testType); err != nil {
			t.Fatalf("runDecTestCaseV2(%s) error: %v", tc.ID, err)
		}
	}
}

func TestRunDecTestCaseV2RejectsOutOfScopeReduceOperation(t *testing.T) {
	tc := decTestCase{
		ID:        "reduce_out_of_scope",
		Operation: "reduce",
		Operands:  []string{"'120.00'"},
		Result:    "1.2E+2",
	}
	if err := runDecTestCaseV2(tc, "decimal64"); !errors.Is(err, errUnsupportedDecTestOperation) {
		t.Fatalf("runDecTestCaseV2(reduce) error = %v, want errUnsupportedDecTestOperation", err)
	}
}

func TestRunDecTestCaseV2SupportsMinMaxOperation(t *testing.T) {
	testCases := []decTestCase{
		{
			ID:        "min64",
			Operation: "min",
			Operands:  []string{"'-2'", "'1'"},
			Result:    "-2",
		},
		{
			ID:        "max64",
			Operation: "max",
			Operands:  []string{"'-2'", "'1'"},
			Result:    "1",
		},
		{
			ID:        "minmag64",
			Operation: "minmag",
			Operands:  []string{"'-2'", "'1'"},
			Result:    "1",
		},
		{
			ID:        "maxmag64",
			Operation: "maxmag",
			Operands:  []string{"'-2'", "'1'"},
			Result:    "-2",
		},
		{
			ID:        "min_subnormal64",
			Operation: "min",
			Operands:  []string{"'-0.1E-383'", "'0'"},
			Result:    "-1E-384",
			Flags:     []string{"Subnormal"},
		},
		{
			ID:        "max_snan64",
			Operation: "max",
			Operands:  []string{"'sNaN88'", "'1'"},
			Result:    "NaN88",
			Flags:     []string{"Invalid_operation"},
		},
		{
			ID:        "min_invalid64",
			Operation: "min",
			Operands:  []string{"#", "'1'"},
			Result:    "NaN",
			Flags:     []string{"Invalid_operation"},
		},
		{
			ID:        "min128",
			Operation: "min",
			Operands:  []string{"'-2'", "'1'"},
			Result:    "-2",
		},
	}

	for _, tc := range testCases {
		testType := "decimal64"
		if tc.ID == "min128" {
			testType = "decimal128"
		}
		if err := runDecTestCaseV2(tc, testType); err != nil {
			t.Fatalf("runDecTestCaseV2(%s) error: %v", tc.ID, err)
		}
	}
}

func TestRunDecTestCaseV2SupportsCompareTotalOperation(t *testing.T) {
	testCases := []decTestCase{
		{
			ID:        "comparetotal_equal_encoding64",
			Operation: "comparetotal",
			Operands:  []string{"7.0", "7"},
			Result:    "-1",
		},
		{
			ID:        "comparetotmag_negative_magnitude64",
			Operation: "comparetotmag",
			Operands:  []string{"-2", "1"},
			Result:    "1",
		},
		{
			ID:        "comparetotal_invalid64",
			Operation: "comparetotal",
			Operands:  []string{"10", "#"},
			Result:    "NaN",
			Flags:     []string{"Invalid_operation"},
		},
		{
			ID:        "comparetotal_nan128",
			Operation: "comparetotal",
			Operands:  []string{"-NaN41", "+NaN42"},
			Result:    "-1",
		},
	}

	for _, tc := range testCases {
		testType := "decimal64"
		if tc.ID == "comparetotal_nan128" {
			testType = "decimal128"
		}
		if err := runDecTestCaseV2(tc, testType); err != nil {
			t.Fatalf("runDecTestCaseV2(%s) error: %v", tc.ID, err)
		}
	}
}

func TestRunDecTestCaseV2SupportsUnaryOperation(t *testing.T) {
	testCases := []decTestCase{
		{
			ID:        "abs_snan64",
			Operation: "abs",
			Operands:  []string{"'-sNaN33'"},
			Result:    "-NaN33",
			Flags:     []string{"Invalid_operation"},
		},
		{
			ID:        "plus_negative_zero64",
			Operation: "plus",
			Operands:  []string{"'-0E+4'"},
			Result:    "0E+4",
		},
		{
			ID:        "minus_negative64",
			Operation: "minus",
			Operands:  []string{"'-7.50'"},
			Result:    "7.50",
		},
		{
			ID:        "plus_subnormal64",
			Operation: "plus",
			Operands:  []string{"'1E-398'"},
			Result:    "1E-398",
			Flags:     []string{"Subnormal"},
		},
		{
			ID:        "minus_invalid64",
			Operation: "minus",
			Operands:  []string{"#"},
			Result:    "NaN",
			Flags:     []string{"Invalid_operation"},
		},
		{
			ID:        "abs128",
			Operation: "abs",
			Operands:  []string{"'-1'"},
			Result:    "1",
		},
	}

	for _, tc := range testCases {
		testType := "decimal64"
		if tc.ID == "abs128" {
			testType = "decimal128"
		}
		if err := runDecTestCaseV2(tc, testType); err != nil {
			t.Fatalf("runDecTestCaseV2(%s) error: %v", tc.ID, err)
		}
	}
}

func TestRunDecTestCaseV2SupportsFMAOperation(t *testing.T) {
	testCases := []decTestCase{
		{
			ID:           "fma_exact64",
			Operation:    "fma",
			Operands:     []string{"2", "2", "3"},
			Result:       "7",
			RoundingMode: "half_even",
			Precision:    16,
			MaxExponent:  384,
			MinExponent:  -383,
			Clamp:        1,
		},
		{
			ID:           "fma_invalid64",
			Operation:    "fma",
			Operands:     []string{"Inf", "Inf", "-Inf"},
			Result:       "NaN",
			Flags:        []string{"Invalid_operation"},
			RoundingMode: "half_even",
			Precision:    16,
			MaxExponent:  384,
			MinExponent:  -383,
			Clamp:        1,
		},
		{
			ID:           "fma_clamped64",
			Operation:    "fma",
			Operands:     []string{"1e+384", "10", "-1e+384"},
			Result:       "9.000000000000000E+384",
			Flags:        []string{"Clamped"},
			RoundingMode: "half_even",
			Precision:    16,
			MaxExponent:  384,
			MinExponent:  -383,
			Clamp:        1,
		},
		{
			ID:           "fma_underflow64",
			Operation:    "fma",
			Operands:     []string{"1e-398", "0.1", "0"},
			Result:       "0E-398",
			Flags:        []string{"Underflow", "Subnormal", "Inexact", "Rounded", "Clamped"},
			RoundingMode: "half_even",
			Precision:    16,
			MaxExponent:  384,
			MinExponent:  -383,
			Clamp:        1,
		},
		{
			ID:           "fma_exact128",
			Operation:    "fma",
			Operands:     []string{"2", "2", "3"},
			Result:       "7",
			RoundingMode: "half_even",
			Precision:    34,
			MaxExponent:  6144,
			MinExponent:  -6143,
			Clamp:        1,
		},
	}

	for _, tc := range testCases {
		testType := "decimal64"
		if tc.ID == "fma_exact128" {
			testType = "decimal128"
		}
		if err := runDecTestCaseV2(tc, testType); err != nil {
			t.Fatalf("runDecTestCaseV2(%s) error: %v", tc.ID, err)
		}
	}
}

func TestRunDecTestCaseV2SupportsLogBOperation(t *testing.T) {
	testCases := []decTestCase{
		{
			ID:        "logb_hundred64",
			Operation: "logb",
			Operands:  []string{"100"},
			Result:    "2",
		},
		{
			ID:        "logb_zero64",
			Operation: "logb",
			Operands:  []string{"0"},
			Result:    "-Infinity",
			Flags:     []string{"Division_by_zero"},
		},
		{
			ID:        "logb_snan64",
			Operation: "logb",
			Operands:  []string{"sNaN123"},
			Result:    "NaN123",
			Flags:     []string{"Invalid_operation"},
		},
		{
			ID:        "logb_hundred128",
			Operation: "logb",
			Operands:  []string{"100"},
			Result:    "2",
		},
	}

	for _, tc := range testCases {
		testType := "decimal64"
		if tc.ID == "logb_hundred128" {
			testType = "decimal128"
		}
		if err := runDecTestCaseV2(tc, testType); err != nil {
			t.Fatalf("runDecTestCaseV2(%s) error: %v", tc.ID, err)
		}
	}
}

func TestRunDecTestCaseV2SupportsScaleBOperation(t *testing.T) {
	testCases := []decTestCase{
		{
			ID:          "scaleb_exact64",
			Operation:   "scaleb",
			Operands:    []string{"7.50", "2"},
			Result:      "750",
			Precision:   16,
			MaxExponent: 384,
			MinExponent: -383,
			Clamp:       1,
		},
		{
			ID:          "scaleb_invalid_rhs64",
			Operation:   "scaleb",
			Operands:    []string{"1.23", "1.00"},
			Result:      "NaN",
			Flags:       []string{"Invalid_operation"},
			Precision:   16,
			MaxExponent: 384,
			MinExponent: -383,
			Clamp:       1,
		},
		{
			ID:          "scaleb_rhs_snan64",
			Operation:   "scaleb",
			Operands:    []string{"4", "sNaN"},
			Result:      "NaN",
			Flags:       []string{"Invalid_operation"},
			Precision:   16,
			MaxExponent: 384,
			MinExponent: -383,
			Clamp:       1,
		},
		{
			ID:          "scaleb_exact128",
			Operation:   "scaleb",
			Operands:    []string{"7.50", "2"},
			Result:      "750",
			Precision:   34,
			MaxExponent: 6144,
			MinExponent: -6143,
			Clamp:       1,
		},
	}

	for _, tc := range testCases {
		testType := "decimal64"
		if tc.ID == "scaleb_exact128" {
			testType = "decimal128"
		}
		if err := runDecTestCaseV2(tc, testType); err != nil {
			t.Fatalf("runDecTestCaseV2(%s) error: %v", tc.ID, err)
		}
	}
}

func TestRunDecTestCaseV2SupportsRemainderNearOperation(t *testing.T) {
	testCases := []decTestCase{
		{
			ID:        "remaindernear_ties64",
			Operation: "remaindernear",
			Operands:  []string{"2", "3"},
			Result:    "-1",
		},
		{
			ID:        "remaindernear_division_undefined64",
			Operation: "remaindernear",
			Operands:  []string{"0", "0"},
			Result:    "NaN",
			Flags:     []string{"Division_undefined"},
		},
		{
			ID:        "remaindernear_ties128",
			Operation: "remaindernear",
			Operands:  []string{"2", "3"},
			Result:    "-1",
		},
	}

	for _, tc := range testCases {
		testType := "decimal64"
		if tc.ID == "remaindernear_ties128" {
			testType = "decimal128"
		}
		if err := runDecTestCaseV2(tc, testType); err != nil {
			t.Fatalf("runDecTestCaseV2(%s) error: %v", tc.ID, err)
		}
	}
}

func TestRunDecTestCaseV2SupportsRemainderOperation(t *testing.T) {
	testCases := []decTestCase{
		{
			ID:        "remainder64",
			Operation: "remainder",
			Operands:  []string{"2", "3"},
			Result:    "2",
		},
		{
			ID:        "remainder_division_undefined64",
			Operation: "remainder",
			Operands:  []string{"0", "0"},
			Result:    "NaN",
			Flags:     []string{"Division_undefined"},
		},
		{
			ID:        "remainder128",
			Operation: "remainder",
			Operands:  []string{"2", "3"},
			Result:    "2",
		},
	}

	for _, tc := range testCases {
		testType := "decimal64"
		if tc.ID == "remainder128" {
			testType = "decimal128"
		}
		if err := runDecTestCaseV2(tc, testType); err != nil {
			t.Fatalf("runDecTestCaseV2(%s) error: %v", tc.ID, err)
		}
	}
}

func TestShouldSkipDecTestCaseSkipsTaggedToIntegralSubsetEdges(t *testing.T) {
	if !shouldSkipDecTestCase(decTestCase{
		Operation: "tointegralx",
		Operands:  []string{"1.23E+384"},
		Result:    "#47fd300000000000",
	}, "decimal64") {
		t.Fatal("expected tagged tointegralx result to be skipped")
	}
}

func TestShouldSkipDecTestCaseSkipsMinMaxIntelDivergenceEdges(t *testing.T) {
	testCases := []struct {
		name      string
		testType  string
		operation string
		operands  []string
		result    string
		flags     []string
		wantSkip  bool
	}{
		{name: "decimal32 min rhs diverges", testType: "decimal32", operation: "min", operands: []string{"-0", "0"}, result: "-0", wantSkip: true},
		{name: "decimal32 min rhs agrees", testType: "decimal32", operation: "min", operands: []string{"0", "-0"}, result: "-0", wantSkip: false},
		{name: "decimal32 max rhs diverges", testType: "decimal32", operation: "max", operands: []string{"0", "-0"}, result: "0", wantSkip: true},
		{name: "decimal32 max rhs agrees", testType: "decimal32", operation: "max", operands: []string{"-0", "0"}, result: "0", wantSkip: false},
		{name: "decimal32 minmag lhs diverges", testType: "decimal32", operation: "minmag", operands: []string{"0", "-0"}, result: "-0", wantSkip: true},
		{name: "decimal32 minmag lhs agrees", testType: "decimal32", operation: "minmag", operands: []string{"-0", "0"}, result: "-0", wantSkip: false},
		{name: "decimal32 maxmag rhs diverges", testType: "decimal32", operation: "maxmag", operands: []string{"0", "-0"}, result: "0", wantSkip: true},
		{name: "decimal32 maxmag rhs agrees", testType: "decimal32", operation: "maxmag", operands: []string{"-0", "0"}, result: "0", wantSkip: false},
		{name: "decimal64 min rhs diverges", testType: "decimal64", operation: "min", operands: []string{"-0", "0"}, result: "-0", wantSkip: true},
		{name: "decimal64 min rhs agrees", testType: "decimal64", operation: "min", operands: []string{"0", "-0"}, result: "-0", wantSkip: false},
		{name: "decimal64 max rhs diverges", testType: "decimal64", operation: "max", operands: []string{"0", "-0"}, result: "0", wantSkip: true},
		{name: "decimal64 max rhs agrees", testType: "decimal64", operation: "max", operands: []string{"-0", "0"}, result: "0", wantSkip: false},
		{name: "decimal64 minmag lhs diverges", testType: "decimal64", operation: "minmag", operands: []string{"0", "-0"}, result: "-0", wantSkip: true},
		{name: "decimal64 minmag lhs agrees", testType: "decimal64", operation: "minmag", operands: []string{"-0", "0"}, result: "-0", wantSkip: false},
		{name: "decimal64 maxmag rhs diverges", testType: "decimal64", operation: "maxmag", operands: []string{"0", "-0"}, result: "0", wantSkip: true},
		{name: "decimal64 maxmag rhs agrees", testType: "decimal64", operation: "maxmag", operands: []string{"-0", "0"}, result: "0", wantSkip: false},
		{name: "decimal128 min lhs diverges", testType: "decimal128", operation: "min", operands: []string{"0", "-0"}, result: "-0", wantSkip: true},
		{name: "decimal128 min lhs agrees", testType: "decimal128", operation: "min", operands: []string{"-0", "0"}, result: "-0", wantSkip: false},
		{name: "decimal128 max lhs diverges", testType: "decimal128", operation: "max", operands: []string{"-0", "0"}, result: "0", wantSkip: true},
		{name: "decimal128 max lhs agrees", testType: "decimal128", operation: "max", operands: []string{"0", "-0"}, result: "0", wantSkip: false},
		{name: "decimal128 minmag lhs diverges", testType: "decimal128", operation: "minmag", operands: []string{"0", "-0"}, result: "-0", wantSkip: true},
		{name: "decimal128 minmag lhs agrees", testType: "decimal128", operation: "minmag", operands: []string{"-0", "0"}, result: "-0", wantSkip: false},
		{name: "decimal128 maxmag rhs diverges", testType: "decimal128", operation: "maxmag", operands: []string{"0", "-0"}, result: "0", wantSkip: true},
		{name: "decimal128 maxmag rhs agrees", testType: "decimal128", operation: "maxmag", operands: []string{"-0", "0"}, result: "0", wantSkip: false},
		{name: "malformed expected exponent", testType: "decimal64", operation: "max", operands: []string{"0", "-0"}, result: "0Ebogus", wantSkip: false},
		{name: "empty expected exponent", testType: "decimal64", operation: "max", operands: []string{"0", "-0"}, result: "0E", wantSkip: false},
		{name: "duplicate expected decimal point", testType: "decimal64", operation: "max", operands: []string{"0", "-0"}, result: "0..0", wantSkip: false},
		{name: "overflowing expected exponent", testType: "decimal64", operation: "max", operands: []string{"0", "-0"}, result: "0E999999999999999999999999999999999999", wantSkip: false},
		{name: "malformed operand exponent", testType: "decimal64", operation: "max", operands: []string{"0Ebogus", "-0"}, result: "0", wantSkip: false},
		{name: "wrong condition", testType: "decimal64", operation: "max", operands: []string{"0", "-0"}, result: "0", flags: []string{"Rounded"}, wantSkip: false},
		{name: "extra condition", testType: "decimal128", operation: "min", operands: []string{"0", "-0"}, result: "-0", flags: []string{"Clamped"}, wantSkip: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := shouldSkipDecTestCase(decTestCase{
				Operation: tc.operation,
				Operands:  tc.operands,
				Result:    tc.result,
				Flags:     tc.flags,
				Precision: map[string]int{"decimal32": 7, "decimal64": 16, "decimal128": 34}[tc.testType],
			}, tc.testType)
			if got != tc.wantSkip {
				t.Fatalf("shouldSkipDecTestCase(%s %v -> %s, %s) = %v, want %v", tc.operation, tc.operands, tc.result, tc.testType, got, tc.wantSkip)
			}
		})
	}

	if !shouldSkipDecTestCase(decTestCase{
		Operation: "min",
		Operands:  []string{"NaN95", "sNaN93"},
		Result:    "NaN93",
		Flags:     []string{"Invalid_operation"},
	}, "decimal128") {
		t.Fatal("expected quiet-NaN/signaling-NaN min/max precedence case to be skipped")
	}
	if shouldSkipDecTestCase(decTestCase{
		Operation: "min",
		Operands:  []string{"-2", "1"},
		Result:    "-2",
	}, "decimal64") {
		t.Fatal("did not expect normal min case to be skipped")
	}
}

func TestShouldSkipDecTestCaseSkipsOnlyDivergentRemainderNaNIdentities(t *testing.T) {
	testCases := []struct {
		name      string
		operation string
		operands  []string
		result    string
		flags     []string
		wantSkip  bool
	}{
		{name: "remainder matching identity", operation: "remainder", operands: []string{"NaN", "sNaN"}, result: "NaN", flags: []string{"Invalid_operation"}, wantSkip: false},
		{name: "remainder divergent payload", operation: "remainder", operands: []string{"NaN3", "sNaN9"}, result: "NaN9", flags: []string{"Invalid_operation"}, wantSkip: true},
		{name: "remaindernear matching identity", operation: "remaindernear", operands: []string{"NaN", "sNaN"}, result: "NaN", flags: []string{"Invalid_operation"}, wantSkip: false},
		{name: "remaindernear divergent payload", operation: "remaindernear", operands: []string{"NaN3", "sNaN9"}, result: "NaN9", flags: []string{"Invalid_operation"}, wantSkip: true},
		{name: "matching payload", operation: "remainder", operands: []string{"NaN9", "sNaN09"}, result: "NaN9", flags: []string{"Invalid_operation"}, wantSkip: false},
		{name: "divergent sign", operation: "remainder", operands: []string{"-NaN9", "sNaN9"}, result: "NaN9", flags: []string{"Invalid_operation"}, wantSkip: true},
		{name: "malformed left payload", operation: "remainder", operands: []string{"NaN3bogus", "sNaN9"}, result: "NaN9", flags: []string{"Invalid_operation"}, wantSkip: false},
		{name: "malformed right payload", operation: "remainder", operands: []string{"NaN3", "sNaN9bogus"}, result: "NaN9", flags: []string{"Invalid_operation"}, wantSkip: false},
		{name: "malformed expected payload", operation: "remainder", operands: []string{"NaN3", "sNaN9"}, result: "NaN9bogus", flags: []string{"Invalid_operation"}, wantSkip: false},
		{name: "qnan left operand", operation: "remainder", operands: []string{"qNaN3", "sNaN9"}, result: "NaN9", flags: []string{"Invalid_operation"}, wantSkip: false},
		{name: "qnan expected result", operation: "remainder", operands: []string{"NaN3", "sNaN9"}, result: "qNaN9", flags: []string{"Invalid_operation"}, wantSkip: false},
		{name: "overlong decimal64 payload", operation: "remainder", operands: []string{"NaN3", "sNaN1234567890123456"}, result: "NaN1234567890123456", flags: []string{"Invalid_operation"}, wantSkip: false},
		{name: "missing expected result", operation: "remainder", operands: []string{"NaN3", "sNaN9"}, flags: []string{"Invalid_operation"}, wantSkip: false},
		{name: "expected does not follow gda precedence", operation: "remainder", operands: []string{"NaN3", "sNaN9"}, result: "NaN3", flags: []string{"Invalid_operation"}, wantSkip: false},
		{name: "missing invalid condition", operation: "remainder", operands: []string{"NaN3", "sNaN9"}, result: "NaN9", wantSkip: false},
		{name: "wrong condition", operation: "remainder", operands: []string{"NaN3", "sNaN9"}, result: "NaN9", flags: []string{"Rounded"}, wantSkip: false},
		{name: "clamped condition", operation: "remainder", operands: []string{"NaN3", "sNaN9"}, result: "NaN9", flags: []string{"Clamped"}, wantSkip: false},
		{name: "division impossible condition", operation: "remainder", operands: []string{"NaN3", "sNaN9"}, result: "NaN9", flags: []string{"Division_impossible"}, wantSkip: false},
		{name: "extra clamped condition", operation: "remainder", operands: []string{"NaN3", "sNaN9"}, result: "NaN9", flags: []string{"Invalid_operation", "Clamped"}, wantSkip: false},
		{name: "extra division impossible condition", operation: "remainder", operands: []string{"NaN3", "sNaN9"}, result: "NaN9", flags: []string{"Invalid_operation", "Division_impossible"}, wantSkip: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := shouldSkipDecTestCase(decTestCase{
				Operation: tc.operation,
				Operands:  tc.operands,
				Result:    tc.result,
				Flags:     tc.flags,
			}, "decimal64")
			if got != tc.wantSkip {
				t.Fatalf("shouldSkipDecTestCase(%s %v -> %s) = %v, want %v", tc.operation, tc.operands, tc.result, got, tc.wantSkip)
			}
		})
	}
}

func TestShouldSkipDecTestCaseKeepsOnlyRemainderGDADivisionImpossibleValueDivergence(t *testing.T) {
	testCases := []struct {
		name       string
		tc         decTestCase
		wantReason string
	}{
		{
			name: "division impossible finite operands",
			tc: decTestCase{
				Operation: "remainder",
				Operands:  []string{"1E+6144", "1"},
				Result:    "NaN",
				Flags:     []string{"Division_impossible"},
			},
			wantReason: "remainder_gda_division_impossible_context_semantics",
		},
		{
			name: "clamped finite operands and result",
			tc: decTestCase{
				Operation: "remaindernear",
				Operands:  []string{"1E+6144", "1E+6143"},
				Result:    "0E+6111",
				Flags:     []string{"Clamped"},
			},
		},
		{name: "division impossible extra condition", tc: decTestCase{Operation: "remainder", Operands: []string{"1", "1"}, Result: "NaN", Flags: []string{"Division_impossible", "Rounded"}}},
		{name: "clamped extra condition", tc: decTestCase{Operation: "remainder", Operands: []string{"1", "1"}, Result: "0", Flags: []string{"Clamped", "Rounded"}}},
		{name: "division impossible NaN operand", tc: decTestCase{Operation: "remainder", Operands: []string{"NaN3", "sNaN9"}, Result: "NaN9", Flags: []string{"Division_impossible"}}},
		{name: "clamped NaN operand", tc: decTestCase{Operation: "remainder", Operands: []string{"NaN3", "sNaN9"}, Result: "NaN9", Flags: []string{"Clamped"}}},
		{name: "division impossible malformed operand", tc: decTestCase{Operation: "remainder", Operands: []string{"1bogus", "1"}, Result: "NaN", Flags: []string{"Division_impossible"}}},
		{name: "division impossible finite result", tc: decTestCase{Operation: "remainder", Operands: []string{"1", "1"}, Result: "0", Flags: []string{"Division_impossible"}}},
		{name: "clamped NaN result", tc: decTestCase{Operation: "remainder", Operands: []string{"1", "1"}, Result: "NaN", Flags: []string{"Clamped"}}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reason, got := decTestCaseSkipReason(tc.tc, "decimal128")
			wantSkip := tc.wantReason != ""
			if got != wantSkip || (got && reason != tc.wantReason) {
				t.Fatalf("decTestCaseSkipReason(%+v) = %q/%v, want %q/%v", tc.tc, reason, got, tc.wantReason, wantSkip)
			}
		})
	}
}

func TestShouldSkipDecTestCaseSkipsOnlyAuthoritativeFMANaNIdentityDivergences(t *testing.T) {
	testCases := []struct {
		name     string
		operands []string
		result   string
		flags    []string
		wantSkip bool
	}{
		{name: "quiet NaN identity divergence", operands: []string{"NaN2", "NaN3", "NaN5"}, result: "NaN2", wantSkip: true},
		{name: "signaling NaN identity divergence", operands: []string{"1", "NaN16", "sNaN19"}, result: "NaN19", flags: []string{"Invalid_operation"}, wantSkip: true},
		{name: "matching identity", operands: []string{"NaN2", "NaN02", "NaN5"}, result: "NaN2", wantSkip: false},
		{name: "maximum decimal64 payload", operands: []string{"NaN123456789012345", "NaN3", "0"}, result: "NaN123456789012345", wantSkip: true},
		{name: "missing expected result", operands: []string{"NaN2", "NaN3", "NaN5"}, wantSkip: false},
		{name: "expected does not follow GDA precedence", operands: []string{"NaN2", "NaN3", "NaN5"}, result: "NaN3", wantSkip: false},
		{name: "signaling expected result", operands: []string{"1", "NaN16", "sNaN19"}, result: "sNaN19", flags: []string{"Invalid_operation"}, wantSkip: false},
		{name: "qNaN expected result", operands: []string{"NaN2", "NaN3", "NaN5"}, result: "qNaN2", wantSkip: false},
		{name: "malformed expected result", operands: []string{"NaN2", "NaN3", "NaN5"}, result: "NaN2bogus", wantSkip: false},
		{name: "qNaN operand", operands: []string{"NaN2", "qNaN3", "NaN5"}, result: "NaN2", wantSkip: false},
		{name: "malformed operand", operands: []string{"NaN2", "NaN3bogus", "NaN5"}, result: "NaN2", wantSkip: false},
		{name: "overlong decimal64 operand payload", operands: []string{"NaN2", "NaN1234567890123456", "0"}, result: "NaN2", wantSkip: false},
		{name: "overlong decimal64 expected payload", operands: []string{"NaN2", "NaN3", "0"}, result: "NaN1234567890123456", wantSkip: false},
		{name: "quiet NaN with invalid condition", operands: []string{"NaN2", "NaN3", "NaN5"}, result: "NaN2", flags: []string{"Invalid_operation"}, wantSkip: false},
		{name: "signaling NaN missing invalid condition", operands: []string{"1", "NaN16", "sNaN19"}, result: "NaN19", wantSkip: false},
		{name: "signaling NaN wrong condition", operands: []string{"1", "NaN16", "sNaN19"}, result: "NaN19", flags: []string{"Rounded"}, wantSkip: false},
		{name: "signaling NaN extra condition", operands: []string{"1", "NaN16", "sNaN19"}, result: "NaN19", flags: []string{"Invalid_operation", "Clamped"}, wantSkip: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := shouldSkipDecTestCase(decTestCase{
				Operation:    "fma",
				Operands:     tc.operands,
				Result:       tc.result,
				Flags:        tc.flags,
				RoundingMode: "half_even",
			}, "decimal64")
			if got != tc.wantSkip {
				t.Fatalf("shouldSkipDecTestCase(fma %v -> %s %v) = %v, want %v", tc.operands, tc.result, tc.flags, got, tc.wantSkip)
			}
		})
	}
}

func TestShouldSkipDecTestCaseDoesNotSkipFMAForNonBIDStatusConditions(t *testing.T) {
	testCases := []struct {
		name       string
		tc         decTestCase
		wantReason string
	}{
		{name: "rounded only finite", tc: decTestCase{Operation: "fma", Operands: []string{"1.23456789", "1.00000000", "0e+384"}, Result: "1.234567890000000", Flags: []string{"Rounded"}, RoundingMode: "half_even"}},
		{name: "subnormal rounded finite", tc: decTestCase{Operation: "fma", Operands: []string{"1.0E-394", "1e-4", "0e+384"}, Result: "1E-398", Flags: []string{"Subnormal", "Rounded"}, RoundingMode: "half_even"}},
		{name: "clamped only finite", tc: decTestCase{Operation: "fma", Operands: []string{"100E+260", "0E+260", "0e+384"}, Result: "0E+369", Flags: []string{"Clamped"}, RoundingMode: "half_even"}},
		{name: "rounded extra condition", tc: decTestCase{Operation: "fma", Operands: []string{"1", "1", "0"}, Result: "1", Flags: []string{"Rounded", "Invalid_operation"}, RoundingMode: "half_even"}},
		{name: "clamped extra condition", tc: decTestCase{Operation: "fma", Operands: []string{"1", "1", "0"}, Result: "1", Flags: []string{"Clamped", "Invalid_operation"}, RoundingMode: "half_even"}},
		{name: "rounded NaN operands", tc: decTestCase{Operation: "fma", Operands: []string{"NaN2", "NaN3", "NaN5"}, Result: "NaN2", Flags: []string{"Rounded"}, RoundingMode: "half_even"}},
		{name: "clamped NaN operands", tc: decTestCase{Operation: "fma", Operands: []string{"NaN2", "NaN3", "NaN5"}, Result: "NaN2", Flags: []string{"Clamped"}, RoundingMode: "half_even"}},
		{name: "rounded malformed operand", tc: decTestCase{Operation: "fma", Operands: []string{"1bogus", "1", "0"}, Result: "1", Flags: []string{"Rounded"}, RoundingMode: "half_even"}},
		{name: "clamped NaN result", tc: decTestCase{Operation: "fma", Operands: []string{"1", "1", "0"}, Result: "NaN", Flags: []string{"Clamped"}, RoundingMode: "half_even"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reason, got := decTestCaseSkipReason(tc.tc, "decimal64")
			wantSkip := tc.wantReason != ""
			if got != wantSkip || (got && reason != tc.wantReason) {
				t.Fatalf("decTestCaseSkipReason(%+v) = %q/%v, want %q/%v", tc.tc, reason, got, tc.wantReason, wantSkip)
			}
		})
	}
}

func TestShouldSkipDecTestCaseDoesNotSkipScaleBForNonBIDStatusConditions(t *testing.T) {
	testCases := []struct {
		name       string
		tc         decTestCase
		wantReason string
	}{
		{name: "rounded only finite", tc: decTestCase{Operation: "scaleb", Operands: []string{"1.0", "-1"}, Result: "0.10", Flags: []string{"Rounded"}}},
		{name: "subnormal rounded finite", tc: decTestCase{Operation: "scaleb", Operands: []string{"1.000000000000000E-383", "-1"}, Result: "1.00000000000000E-384", Flags: []string{"Subnormal", "Rounded"}}},
		{name: "clamped only finite", tc: decTestCase{Operation: "scaleb", Operands: []string{"1000E+369", "+1"}, Result: "1.0000E+373", Flags: []string{"Clamped"}}},
		{name: "rounded extra condition", tc: decTestCase{Operation: "scaleb", Operands: []string{"1", "1"}, Result: "1E+1", Flags: []string{"Rounded", "Invalid_operation"}}},
		{name: "clamped extra condition", tc: decTestCase{Operation: "scaleb", Operands: []string{"1", "1"}, Result: "1E+1", Flags: []string{"Clamped", "Invalid_operation"}}},
		{name: "rounded NaN operand", tc: decTestCase{Operation: "scaleb", Operands: []string{"NaN", "1"}, Result: "NaN", Flags: []string{"Rounded"}}},
		{name: "clamped infinity operand", tc: decTestCase{Operation: "scaleb", Operands: []string{"Infinity", "1"}, Result: "Infinity", Flags: []string{"Clamped"}}},
		{name: "rounded malformed exponent operand", tc: decTestCase{Operation: "scaleb", Operands: []string{"1", "1bogus"}, Result: "1E+1", Flags: []string{"Rounded"}}},
		{name: "clamped NaN result", tc: decTestCase{Operation: "scaleb", Operands: []string{"1", "1"}, Result: "NaN", Flags: []string{"Clamped"}}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reason, got := decTestCaseSkipReason(tc.tc, "decimal64")
			wantSkip := tc.wantReason != ""
			if got != wantSkip || (got && reason != tc.wantReason) {
				t.Fatalf("decTestCaseSkipReason(%+v) = %q/%v, want %q/%v", tc.tc, reason, got, tc.wantReason, wantSkip)
			}
		})
	}
}

func TestShouldSkipDecTestCaseSkipsOnlyDivergentNextTowardNaNIdentities(t *testing.T) {
	testCases := []struct {
		name     string
		operands []string
		result   string
		flags    []string
		wantSkip bool
	}{
		{name: "left signaling identity agrees", operands: []string{"sNaN15", "sNaN18"}, result: "NaN15", flags: []string{"Invalid_operation"}, wantSkip: false},
		{name: "negative left signaling identity agrees", operands: []string{"-sNaN27", "sNaN29"}, result: "-NaN27", flags: []string{"Invalid_operation"}, wantSkip: false},
		{name: "quiet left payload diverges", operands: []string{"NaN16", "sNaN19"}, result: "NaN19", flags: []string{"Invalid_operation"}, wantSkip: true},
		{name: "quiet left signed payload diverges", operands: []string{"+NaN25", "+sNaN24"}, result: "NaN24", flags: []string{"Invalid_operation"}, wantSkip: true},
		{name: "quiet and signaling identities agree", operands: []string{"NaN19", "sNaN019"}, result: "NaN19", flags: []string{"Invalid_operation"}, wantSkip: false},
		{name: "quiet and signaling signs diverge", operands: []string{"-NaN19", "sNaN19"}, result: "NaN19", flags: []string{"Invalid_operation"}, wantSkip: true},
		{name: "malformed left payload", operands: []string{"NaN1bogus", "sNaN19"}, result: "NaN19", flags: []string{"Invalid_operation"}, wantSkip: false},
		{name: "qnan left operand", operands: []string{"qNaN16", "sNaN19"}, result: "NaN19", flags: []string{"Invalid_operation"}, wantSkip: false},
		{name: "overlong decimal64 payload", operands: []string{"NaN16", "sNaN1234567890123456"}, result: "NaN1234567890123456", flags: []string{"Invalid_operation"}, wantSkip: false},
		{name: "missing expected result", operands: []string{"NaN16", "sNaN19"}, flags: []string{"Invalid_operation"}, wantSkip: false},
		{name: "expected does not follow gda precedence", operands: []string{"NaN16", "sNaN19"}, result: "NaN16", flags: []string{"Invalid_operation"}, wantSkip: false},
		{name: "missing invalid condition", operands: []string{"NaN16", "sNaN19"}, result: "NaN19", wantSkip: false},
		{name: "wrong condition", operands: []string{"NaN16", "sNaN19"}, result: "NaN19", flags: []string{"Rounded"}, wantSkip: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := shouldSkipDecTestCase(decTestCase{
				Operation: "nexttoward",
				Operands:  tc.operands,
				Result:    tc.result,
				Flags:     tc.flags,
			}, "decimal64")
			if got != tc.wantSkip {
				t.Fatalf("shouldSkipDecTestCase(nexttoward %v -> %s) = %v, want %v", tc.operands, tc.result, got, tc.wantSkip)
			}
		})
	}
}

func TestShouldSkipDecTestCaseSkipsOnlyDivergentMinMaxNaNIdentities(t *testing.T) {
	testCases := []struct {
		name      string
		operation string
		operands  []string
		result    string
		flags     []string
		wantSkip  bool
	}{
		{name: "min matching identity", operation: "min", operands: []string{"NaN", "sNaN"}, result: "NaN", flags: []string{"Invalid_operation"}, wantSkip: false},
		{name: "min divergent payload", operation: "min", operands: []string{"NaN95", "sNaN93"}, result: "NaN93", flags: []string{"Invalid_operation"}, wantSkip: true},
		{name: "max matching identity", operation: "max", operands: []string{"NaN", "sNaN"}, result: "NaN", flags: []string{"Invalid_operation"}, wantSkip: false},
		{name: "max divergent payload", operation: "max", operands: []string{"NaN95", "sNaN93"}, result: "NaN93", flags: []string{"Invalid_operation"}, wantSkip: true},
		{name: "minmag matching identity", operation: "minmag", operands: []string{"-NaN00", "-sNaN"}, result: "-NaN", flags: []string{"Invalid_operation"}, wantSkip: false},
		{name: "minmag divergent sign", operation: "minmag", operands: []string{"NaN", "-sNaN"}, result: "-NaN", flags: []string{"Invalid_operation"}, wantSkip: true},
		{name: "maxmag matching identity", operation: "maxmag", operands: []string{"NaN", "sNaN0"}, result: "NaN00", flags: []string{"Invalid_operation"}, wantSkip: false},
		{name: "maxmag divergent payload", operation: "maxmag", operands: []string{"NaN95", "sNaN93"}, result: "NaN93", flags: []string{"Invalid_operation"}, wantSkip: true},
		{name: "decimal64 maximum payload", operation: "min", operands: []string{"NaN1", "sNaN123456789012345"}, result: "NaN123456789012345", flags: []string{"Invalid_operation"}, wantSkip: true},
		{name: "malformed left payload", operation: "min", operands: []string{"NaNbogus", "sNaN93"}, result: "NaN93", flags: []string{"Invalid_operation"}, wantSkip: false},
		{name: "malformed right payload", operation: "min", operands: []string{"NaN95", "sNaNbogus"}, result: "NaN93", flags: []string{"Invalid_operation"}, wantSkip: false},
		{name: "malformed expected payload", operation: "min", operands: []string{"NaN95", "sNaN93"}, result: "NaNbogus", flags: []string{"Invalid_operation"}, wantSkip: false},
		{name: "overlong decimal64 payload", operation: "min", operands: []string{"NaN1", "sNaN1234567890123456"}, result: "NaN1234567890123456", flags: []string{"Invalid_operation"}, wantSkip: false},
		{name: "qnan left operand", operation: "min", operands: []string{"qNaN95", "sNaN93"}, result: "NaN93", flags: []string{"Invalid_operation"}, wantSkip: false},
		{name: "qnan expected result", operation: "min", operands: []string{"NaN95", "sNaN93"}, result: "qNaN93", flags: []string{"Invalid_operation"}, wantSkip: false},
		{name: "missing expected result", operation: "min", operands: []string{"NaN95", "sNaN93"}, flags: []string{"Invalid_operation"}, wantSkip: false},
		{name: "expected does not follow gda precedence", operation: "min", operands: []string{"NaN95", "sNaN93"}, result: "NaN95", flags: []string{"Invalid_operation"}, wantSkip: false},
		{name: "missing invalid condition", operation: "min", operands: []string{"NaN95", "sNaN93"}, result: "NaN93", wantSkip: false},
		{name: "wrong condition", operation: "min", operands: []string{"NaN95", "sNaN93"}, result: "NaN93", flags: []string{"Rounded"}, wantSkip: false},
		{name: "unknown condition", operation: "min", operands: []string{"NaN95", "sNaN93"}, result: "NaN93", flags: []string{"Bogus_condition"}, wantSkip: false},
		{name: "extra condition", operation: "min", operands: []string{"NaN95", "sNaN93"}, result: "NaN93", flags: []string{"Invalid_operation", "Rounded"}, wantSkip: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := shouldSkipDecTestCase(decTestCase{
				Operation: tc.operation,
				Operands:  tc.operands,
				Result:    tc.result,
				Flags:     tc.flags,
			}, "decimal64")
			if got != tc.wantSkip {
				t.Fatalf("shouldSkipDecTestCase(%s %v -> %s) = %v, want %v", tc.operation, tc.operands, tc.result, got, tc.wantSkip)
			}
		})
	}

	for _, tc := range []struct {
		name     string
		testType string
		payload  string
		wantSkip bool
	}{
		{name: "decimal32 maximum payload", testType: "decimal32", payload: "123456", wantSkip: true},
		{name: "decimal32 overlong payload", testType: "decimal32", payload: "1234567", wantSkip: false},
		{name: "decimal128 maximum payload", testType: "decimal128", payload: "123456789012345678901234567890123", wantSkip: true},
		{name: "decimal128 overlong payload", testType: "decimal128", payload: "1234567890123456789012345678901234", wantSkip: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := shouldSkipDecTestCase(decTestCase{
				Operation: "min",
				Operands:  []string{"NaN1", "sNaN" + tc.payload},
				Result:    "NaN" + tc.payload,
				Flags:     []string{"Invalid_operation"},
			}, tc.testType)
			if got != tc.wantSkip {
				t.Fatalf("payload length %d for %s: got skip %v, want %v", len(tc.payload), tc.testType, got, tc.wantSkip)
			}
		})
	}
}

func TestShouldSkipDecTestCaseClassifiesFMASubsetEdges(t *testing.T) {
	testCases := []struct {
		caseData decTestCase
		wantSkip bool
	}{
		{
			caseData: decTestCase{
				ID:           "fma_nan_payload_precedence",
				Operation:    "fma",
				Operands:     []string{"NaN2", "NaN3", "NaN5"},
				Result:       "NaN2",
				RoundingMode: "half_even",
			},
			wantSkip: true,
		},
		{
			caseData: decTestCase{
				ID:           "fma_unsupported_rounding",
				Operation:    "fma",
				Operands:     []string{"1", "0", "0E-19"},
				Result:       "0E-19",
				RoundingMode: "up",
			},
			wantSkip: true,
		},
		{
			caseData: decTestCase{
				ID:           "fma_rounded_only_status",
				Operation:    "fma",
				Operands:     []string{"1.23456789", "1.00000000", "0e+384"},
				Result:       "1.234567890000000",
				Flags:        []string{"Rounded"},
				RoundingMode: "half_even",
			},
		},
		{
			caseData: decTestCase{
				ID:           "fma_clamped_only_status",
				Operation:    "fma",
				Operands:     []string{"100E+260", "0E+260", "0e+384"},
				Result:       "0E+369",
				Flags:        []string{"Clamped"},
				RoundingMode: "half_even",
			},
		},
	}

	for _, tc := range testCases {
		if got := shouldSkipDecTestCase(tc.caseData, "decimal64"); got != tc.wantSkip {
			t.Fatalf("shouldSkipDecTestCase(%s) = %v, want %v", tc.caseData.ID, got, tc.wantSkip)
		}
	}

	if shouldSkipDecTestCase(decTestCase{
		ID:           "fma_inexact",
		Operation:    "fma",
		Operands:     []string{"27583489.6645", "2582471078.04", "2593183.42371"},
		Result:       "7.123356429257970E+16",
		Flags:        []string{"Inexact", "Rounded"},
		RoundingMode: "half_even",
		Precision:    16,
	}, "decimal64") {
		t.Fatal("did not expect regular inexact fma case to be skipped")
	}
}

func TestShouldSkipDecTestCaseClassifiesRemainderNearSubsetEdges(t *testing.T) {
	testCases := []struct {
		caseData decTestCase
		wantSkip bool
	}{
		{
			caseData: decTestCase{
				ID:        "remaindernear_division_impossible",
				Operation: "remaindernear",
				Operands:  []string{"1", "0"},
				Result:    "NaN",
				Flags:     []string{"Division_impossible"},
			},
			wantSkip: true,
		},
		{
			caseData: decTestCase{
				ID:        "remaindernear_clamped_only_status",
				Operation: "remaindernear",
				Operands:  []string{"1E-383", "1E-383"},
				Result:    "0E-398",
				Flags:     []string{"Clamped"},
			},
		},
		{
			caseData: decTestCase{
				ID:        "remaindernear_nan_payload_precedence",
				Operation: "remaindernear",
				Operands:  []string{"NaN3", "sNaN9"},
				Result:    "NaN9",
				Flags:     []string{"Invalid_operation"},
			},
			wantSkip: true,
		},
	}

	for _, tc := range testCases {
		if got := shouldSkipDecTestCase(tc.caseData, "decimal64"); got != tc.wantSkip {
			t.Fatalf("shouldSkipDecTestCase(%s) = %v, want %v", tc.caseData.ID, got, tc.wantSkip)
		}
	}

	if shouldSkipDecTestCase(decTestCase{
		ID:        "remaindernear_ties",
		Operation: "remaindernear",
		Operands:  []string{"2", "3"},
		Result:    "-1",
	}, "decimal64") {
		t.Fatal("did not expect regular remaindernear case to be skipped")
	}
}

func TestShouldSkipDecTestCaseClassifiesRemainderSubsetEdges(t *testing.T) {
	testCases := []struct {
		caseData decTestCase
		wantSkip bool
	}{
		{
			caseData: decTestCase{
				ID:        "remainder_division_impossible",
				Operation: "remainder",
				Operands:  []string{"1", "0"},
				Result:    "NaN",
				Flags:     []string{"Division_impossible"},
			},
			wantSkip: true,
		},
		{
			caseData: decTestCase{
				ID:        "remainder_clamped_only_status",
				Operation: "remainder",
				Operands:  []string{"1E-383", "1E-383"},
				Result:    "0E-398",
				Flags:     []string{"Clamped"},
			},
		},
		{
			caseData: decTestCase{
				ID:        "remainder_nan_payload_precedence",
				Operation: "remainder",
				Operands:  []string{"NaN3", "sNaN9"},
				Result:    "NaN9",
				Flags:     []string{"Invalid_operation"},
			},
			wantSkip: true,
		},
	}

	for _, tc := range testCases {
		if got := shouldSkipDecTestCase(tc.caseData, "decimal64"); got != tc.wantSkip {
			t.Fatalf("shouldSkipDecTestCase(%s) = %v, want %v", tc.caseData.ID, got, tc.wantSkip)
		}
	}

	if shouldSkipDecTestCase(decTestCase{
		ID:        "remainder",
		Operation: "remainder",
		Operands:  []string{"2", "3"},
		Result:    "2",
	}, "decimal64") {
		t.Fatal("did not expect regular remainder case to be skipped")
	}
}

func TestParseDecTestFileIgnoresInlineCommentsAfterResult(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "inline-comment.decTest")
	content := "version: 2.62\n" +
		"extended: 1\n" +
		"dectest: add\n" +
		"precision: 16 -- exact numeric directive with a comment\n" +
		"maxexponent: 384 -- upper boundary\n" +
		"minexponent: -383\n" +
		"clamp: 1\n" +
		"rounding: half_up\n" +
		"comment001 multiply 1.20 0 -> 0.00 -- note: rhs is 0\n" +
		"comment002 add -0 -0 -> -0 -- note: IEEE 854 special case\n" +
		"comment003 add '--1' 0 -> NaN -- note: preserve quoted -- and ignore -> in the comment\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write dectest file: %v", err)
	}

	cases, err := parseDecTestFile(path)
	if err != nil {
		t.Fatalf("parseDecTestFile returned error: %v", err)
	}
	if len(cases) != 3 {
		t.Fatalf("expected 3 cases, got %d", len(cases))
	}

	if len(cases[0].Flags) != 0 {
		t.Fatalf("expected inline comment to be ignored for first case, got %v", cases[0].Flags)
	}
	if got := cases[0]; got.Precision != 16 || got.MaxExponent != 384 || got.MinExponent != -383 || got.Clamp != 1 || got.RoundingMode != "half_up" {
		t.Fatalf("parsed context = precision %d emax %d emin %d clamp %d rounding %q", got.Precision, got.MaxExponent, got.MinExponent, got.Clamp, got.RoundingMode)
	}
	if len(cases[1].Flags) != 0 {
		t.Fatalf("expected inline comment to be ignored for second case, got %v", cases[1].Flags)
	}
	if len(cases[2].Operands) != 2 || cases[2].Operands[0] != "'--1'" {
		t.Fatalf("expected quoted comment marker to remain an operand, got %v", cases[2].Operands)
	}
	if len(cases[2].Flags) != 0 {
		t.Fatalf("expected arrow in inline comment to be ignored, got flags %v", cases[2].Flags)
	}
}

func TestParseDecTestFileRejectsMalformedDirectivesAndCases(t *testing.T) {
	testCases := []struct {
		name    string
		content string
	}{
		{name: "invalid precision", content: "precision: nope\n"},
		{name: "trailing precision junk", content: "precision: 5junk\n"},
		{name: "unknown rounding", content: "rounding: nearestish\n"},
		{name: "invalid clamp", content: "clamp: 2\n"},
		{name: "invalid extended", content: "extended: 2\n"},
		{name: "directive with arrow", content: "precision: 16 -> 7\n"},
		{name: "unknown directive", content: "mystery: value\n"},
		{name: "unexpected content", content: "orphan text\n"},
		{name: "missing result", content: "bad001 add 1 2 ->\n"},
		{name: "multiple arrows", content: "bad002 add 1 2 -> 3 -> 4\n"},
		{name: "unterminated quote", content: "bad003 add '1 2 -> 3\n"},
		{name: "missing operands", content: "bad004 add -> 3\n"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "malformed.decTest")
			if err := os.WriteFile(path, []byte(tc.content), 0o644); err != nil {
				t.Fatalf("write dectest file: %v", err)
			}
			if _, err := parseDecTestFile(path); err == nil {
				t.Fatalf("parseDecTestFile accepted malformed input %q", tc.content)
			}
		})
	}
}

func TestRunDecTestCaseV2MarksUnsupportedOperations(t *testing.T) {
	err := runDecTestCaseV2(decTestCase{
		Operation: "apply",
		Operands:  []string{"1"},
		Result:    "1",
		Precision: 16,
	}, "decimal64")
	if !errors.Is(err, errUnsupportedDecTestOperation) {
		t.Fatalf("expected unsupported operation error, got %v", err)
	}
}

func TestShouldSkipDecTestFlagsPortableArithmeticFlagEdges(t *testing.T) {
	if NativeBackendEnabled() {
		t.Skip("portable-only expectation")
	}

	testCases := []decTestCase{
		{Operation: "divide", Flags: []string{"Division_undefined"}},
		{Operation: "add", Flags: []string{"Clamped"}},
	}

	for _, tc := range testCases {
		if !shouldSkipDecTestFlags(tc) {
			t.Fatalf("expected portable path to skip %v", tc.Flags)
		}
	}
}

func TestShouldSkipDecTestFlagsPortableIgnoresOtherCases(t *testing.T) {
	if NativeBackendEnabled() {
		t.Skip("portable-only expectation")
	}

	testCases := []decTestCase{
		{Operation: "toSci", Flags: []string{"Clamped"}},
		{Operation: "add", Flags: []string{"Rounded"}},
	}

	for _, tc := range testCases {
		if shouldSkipDecTestFlags(tc) {
			t.Fatalf("expected %q with flags %v to remain runnable", tc.Operation, tc.Flags)
		}
	}
}

func TestShouldSkipDecTestCaseSkipsUnsupportedTaggedGeneralLiterals(t *testing.T) {
	tc := decTestCase{
		Operation: "quantize",
		Operands:  []string{"64#8.666666666666000E+384", "128#1E+384"},
		Result:    "64#9E+384",
	}

	if !shouldSkipDecTestCase(tc, "general") {
		t.Fatal("expected tagged general decTest literal to be skipped")
	}

	if shouldSkipDecTestCase(tc, "decimal64") {
		t.Fatal("expected decimal64 test type to decide based on precision limits instead")
	}
}
