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

func TestCompareDecTestFlagsNormalizesAliasesAndOrder(t *testing.T) {
	expected := []string{"Rounded", "Division_by_zero", "inexact"}
	actual := FlagInexact | FlagRounded | FlagDivisionByZero

	if !compareDecTestFlags(expected, actual) {
		t.Fatalf("expected flags %v to match %s", expected, actual.String())
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
			},
			testType: "decimal64",
			want:     "nexttoward_nan_payload_precedence",
		},
		{
			name: "minmax zero tie",
			tc: decTestCase{
				Operation: "max",
				Operands:  []string{"0", "-0"},
			},
			testType: "decimal64",
			want:     "minmax_zero_tie",
		},
		{
			name: "minmax nan payload",
			tc: decTestCase{
				Operation: "min",
				Operands:  []string{"NaN95", "sNaN93"},
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
				RoundingMode: "half_even",
			},
			testType: "decimal64",
			want:     "fma_nan_payload_precedence",
		},
		{
			name: "fma rounded only",
			tc: decTestCase{
				Operation:    "fma",
				Operands:     []string{"1.23456789", "1.00000000", "0e+384"},
				Flags:        []string{"Rounded"},
				RoundingMode: "half_even",
			},
			testType: "decimal64",
			want:     "fma_rounded_only_status_gap",
		},
		{
			name: "fma clamped only",
			tc: decTestCase{
				Operation:    "fma",
				Operands:     []string{"100E+260", "0E+260", "0e+384"},
				Flags:        []string{"Clamped"},
				RoundingMode: "half_even",
			},
			testType: "decimal64",
			want:     "fma_clamped_status_gap",
		},
		{
			name: "scaleb rounded only",
			tc: decTestCase{
				Operation: "scaleb",
				Flags:     []string{"Rounded"},
			},
			testType: "decimal64",
			want:     "scaleb_rounded_only_status_gap",
		},
		{
			name: "scaleb clamped only",
			tc: decTestCase{
				Operation: "scaleb",
				Flags:     []string{"Clamped"},
			},
			testType: "decimal64",
			want:     "scaleb_clamped_status_gap",
		},
		{
			name: "remainder division impossible",
			tc: decTestCase{
				Operation: "remainder",
				Flags:     []string{"Division_impossible"},
			},
			testType: "decimal64",
			want:     "remainder_division_impossible_status_gap",
		},
		{
			name: "remaindernear nan payload",
			tc: decTestCase{
				Operation: "remaindernear",
				Operands:  []string{"NaN3", "sNaN9"},
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

func TestRunDecTestCaseV2SupportsReduceOperation(t *testing.T) {
	testCases := []decTestCase{
		{
			ID:        "reduce_normal64",
			Operation: "reduce",
			Operands:  []string{"'120.00'"},
			Result:    "1.2E+2",
		},
		{
			ID:        "reduce_subnormal64",
			Operation: "reduce",
			Operands:  []string{"'2.000E-395'"},
			Result:    "2E-395",
			Flags:     []string{"Subnormal"},
		},
		{
			ID:        "reduce_snan64",
			Operation: "reduce",
			Operands:  []string{"'sNaN010'"},
			Result:    "NaN10",
			Flags:     []string{"Invalid_operation"},
		},
		{
			ID:        "reduce_invalid64",
			Operation: "reduce",
			Operands:  []string{"#"},
			Result:    "NaN",
			Flags:     []string{"Invalid_operation"},
		},
	}

	for _, tc := range testCases {
		if err := runDecTestCaseV2(tc, "decimal64"); err != nil {
			t.Fatalf("runDecTestCaseV2(%s) error: %v", tc.ID, err)
		}
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
	if !shouldSkipDecTestCase(decTestCase{
		Operation: "max",
		Operands:  []string{"0", "-0"},
		Result:    "0",
	}, "decimal64") {
		t.Fatal("expected zero-vs-zero min/max tie case to be skipped")
	}
	if !shouldSkipDecTestCase(decTestCase{
		Operation: "min",
		Operands:  []string{"NaN95", "sNaN93"},
		Result:    "NaN93",
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

func TestShouldSkipDecTestCaseSkipsFMASubsetEdges(t *testing.T) {
	testCases := []decTestCase{
		{
			ID:           "fma_nan_payload_precedence",
			Operation:    "fma",
			Operands:     []string{"NaN2", "NaN3", "NaN5"},
			Result:       "NaN2",
			RoundingMode: "half_even",
		},
		{
			ID:           "fma_unsupported_rounding",
			Operation:    "fma",
			Operands:     []string{"1", "0", "0E-19"},
			Result:       "0E-19",
			RoundingMode: "up",
		},
		{
			ID:           "fma_rounded_only_status",
			Operation:    "fma",
			Operands:     []string{"1.23456789", "1.00000000", "0e+384"},
			Result:       "1.234567890000000",
			Flags:        []string{"Rounded"},
			RoundingMode: "half_even",
		},
		{
			ID:           "fma_clamped_only_status",
			Operation:    "fma",
			Operands:     []string{"100E+260", "0E+260", "0e+384"},
			Result:       "0E+369",
			Flags:        []string{"Clamped"},
			RoundingMode: "half_even",
		},
	}

	for _, tc := range testCases {
		if !shouldSkipDecTestCase(tc, "decimal64") {
			t.Fatalf("expected %s to be skipped", tc.ID)
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

func TestShouldSkipDecTestCaseSkipsRemainderNearSubsetEdges(t *testing.T) {
	testCases := []decTestCase{
		{
			ID:        "remaindernear_division_impossible",
			Operation: "remaindernear",
			Operands:  []string{"1", "0"},
			Result:    "NaN",
			Flags:     []string{"Division_impossible"},
		},
		{
			ID:        "remaindernear_clamped_only_status",
			Operation: "remaindernear",
			Operands:  []string{"1E-383", "1E-383"},
			Result:    "0E-398",
			Flags:     []string{"Clamped"},
		},
		{
			ID:        "remaindernear_nan_payload_precedence",
			Operation: "remaindernear",
			Operands:  []string{"NaN3", "sNaN9"},
			Result:    "NaN9",
		},
	}

	for _, tc := range testCases {
		if !shouldSkipDecTestCase(tc, "decimal64") {
			t.Fatalf("expected %s to be skipped", tc.ID)
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

func TestShouldSkipDecTestCaseSkipsRemainderSubsetEdges(t *testing.T) {
	testCases := []decTestCase{
		{
			ID:        "remainder_division_impossible",
			Operation: "remainder",
			Operands:  []string{"1", "0"},
			Result:    "NaN",
			Flags:     []string{"Division_impossible"},
		},
		{
			ID:        "remainder_clamped_only_status",
			Operation: "remainder",
			Operands:  []string{"1E-383", "1E-383"},
			Result:    "0E-398",
			Flags:     []string{"Clamped"},
		},
		{
			ID:        "remainder_nan_payload_precedence",
			Operation: "remainder",
			Operands:  []string{"NaN3", "sNaN9"},
			Result:    "NaN9",
		},
	}

	for _, tc := range testCases {
		if !shouldSkipDecTestCase(tc, "decimal64") {
			t.Fatalf("expected %s to be skipped", tc.ID)
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
	content := "precision: 16\n" +
		"comment001 multiply 1.20 0 -> 0.00 -- rhs is 0\n" +
		"comment002 add -0 -0 -> -0 -- IEEE 854 special case\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write dectest file: %v", err)
	}

	cases, err := parseDecTestFile(path)
	if err != nil {
		t.Fatalf("parseDecTestFile returned error: %v", err)
	}
	if len(cases) != 2 {
		t.Fatalf("expected 2 cases, got %d", len(cases))
	}

	if len(cases[0].Flags) != 0 {
		t.Fatalf("expected inline comment to be ignored for first case, got %v", cases[0].Flags)
	}
	if len(cases[1].Flags) != 0 {
		t.Fatalf("expected inline comment to be ignored for second case, got %v", cases[1].Flags)
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
