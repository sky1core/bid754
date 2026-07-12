package testgen

import "testing"

func TestGeneratedDectestCaseSkipReasonIsEmptyWhenCaseIsNotSkipped(t *testing.T) {
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
			reason, ok := generatedDectestCaseSkipReason(parsedCase{
				Operation: "add",
				Precision: tc.precision,
			}, tc.testType)
			if ok || reason != "" {
				t.Fatalf("generatedDectestCaseSkipReason precision=%d type=%s = %q/%v, want empty/false", tc.precision, tc.testType, reason, ok)
			}
		})
	}
}

func TestGeneratedDectestMinMaxZeroTieSkipsOnlyPinnedIntelDivergences(t *testing.T) {
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
			reason, got := generatedDectestCaseSkipReason(parsedCase{
				Operation: tc.operation,
				Operands:  tc.operands,
				Result:    tc.result,
				Flags:     tc.flags,
				Precision: map[string]int{"decimal32": 7, "decimal64": 16, "decimal128": 34}[tc.testType],
			}, tc.testType)
			if got != tc.wantSkip {
				t.Fatalf("generatedDectestCaseSkipReason(%s %v -> %s, %s) = %q/%v, want skip %v", tc.operation, tc.operands, tc.result, tc.testType, reason, got, tc.wantSkip)
			}
			if got && reason != "minmax_zero_tie" {
				t.Fatalf("skip reason = %q, want minmax_zero_tie", reason)
			}
		})
	}
}

func TestGeneratedDectestRemainderNaNPrecedenceSkipsOnlyDivergentIdentities(t *testing.T) {
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
			reason, got := generatedDectestCaseSkipReason(parsedCase{
				Operation: tc.operation,
				Operands:  tc.operands,
				Result:    tc.result,
				Flags:     tc.flags,
			}, "decimal64")
			if got != tc.wantSkip {
				t.Fatalf("generatedDectestCaseSkipReason(%s %v -> %s) = %q/%v, want skip %v", tc.operation, tc.operands, tc.result, reason, got, tc.wantSkip)
			}
			if got && reason != tc.operation+"_nan_payload_precedence" {
				t.Fatalf("skip reason = %q, want %s_nan_payload_precedence", reason, tc.operation)
			}
		})
	}
}

func TestGeneratedDectestRequiresExactRemainderStatusGapShape(t *testing.T) {
	testCases := []struct {
		name       string
		tc         parsedCase
		wantReason string
	}{
		{
			name: "division impossible finite operands",
			tc: parsedCase{
				Operation: "remainder",
				Operands:  []string{"1E+6144", "1"},
				Result:    "NaN",
				Flags:     []string{"Division_impossible"},
			},
			wantReason: "remainder_division_impossible_status_gap",
		},
		{
			name: "clamped finite operands and result",
			tc: parsedCase{
				Operation: "remaindernear",
				Operands:  []string{"1E+6144", "1E+6143"},
				Result:    "0E+6111",
				Flags:     []string{"Clamped"},
			},
			wantReason: "remaindernear_clamped_status_gap",
		},
		{name: "division impossible extra condition", tc: parsedCase{Operation: "remainder", Operands: []string{"1", "1"}, Result: "NaN", Flags: []string{"Division_impossible", "Rounded"}}},
		{name: "clamped extra condition", tc: parsedCase{Operation: "remainder", Operands: []string{"1", "1"}, Result: "0", Flags: []string{"Clamped", "Rounded"}}},
		{name: "division impossible NaN operand", tc: parsedCase{Operation: "remainder", Operands: []string{"NaN3", "sNaN9"}, Result: "NaN9", Flags: []string{"Division_impossible"}}},
		{name: "clamped NaN operand", tc: parsedCase{Operation: "remainder", Operands: []string{"NaN3", "sNaN9"}, Result: "NaN9", Flags: []string{"Clamped"}}},
		{name: "division impossible malformed operand", tc: parsedCase{Operation: "remainder", Operands: []string{"1bogus", "1"}, Result: "NaN", Flags: []string{"Division_impossible"}}},
		{name: "division impossible finite result", tc: parsedCase{Operation: "remainder", Operands: []string{"1", "1"}, Result: "0", Flags: []string{"Division_impossible"}}},
		{name: "clamped NaN result", tc: parsedCase{Operation: "remainder", Operands: []string{"1", "1"}, Result: "NaN", Flags: []string{"Clamped"}}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reason, got := generatedDectestCaseSkipReason(tc.tc, "decimal128")
			wantSkip := tc.wantReason != ""
			if got != wantSkip || (got && reason != tc.wantReason) {
				t.Fatalf("generatedDectestCaseSkipReason(%+v) = %q/%v, want %q/%v", tc.tc, reason, got, tc.wantReason, wantSkip)
			}
		})
	}
}

func TestGeneratedDectestSkipsOnlyAuthoritativeFMANaNIdentityDivergences(t *testing.T) {
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
			reason, got := generatedDectestCaseSkipReason(parsedCase{
				Operation:    "fma",
				Operands:     tc.operands,
				Result:       tc.result,
				Flags:        tc.flags,
				RoundingMode: "half_even",
			}, "decimal64")
			if got != tc.wantSkip {
				t.Fatalf("generatedDectestCaseSkipReason(fma %v -> %s %v) = %q/%v, want skip %v", tc.operands, tc.result, tc.flags, reason, got, tc.wantSkip)
			}
			if got && reason != "fma_nan_payload_precedence" {
				t.Fatalf("skip reason = %q, want fma_nan_payload_precedence", reason)
			}
		})
	}
}

func TestGeneratedDectestRequiresExactFMAStatusGapShape(t *testing.T) {
	testCases := []struct {
		name       string
		tc         parsedCase
		wantReason string
	}{
		{name: "rounded only finite", tc: parsedCase{Operation: "fma", Operands: []string{"1.23456789", "1.00000000", "0e+384"}, Result: "1.234567890000000", Flags: []string{"Rounded"}, RoundingMode: "half_even"}, wantReason: "fma_rounded_only_status_gap"},
		{name: "subnormal rounded finite", tc: parsedCase{Operation: "fma", Operands: []string{"1.0E-394", "1e-4", "0e+384"}, Result: "1E-398", Flags: []string{"Subnormal", "Rounded"}, RoundingMode: "half_even"}, wantReason: "fma_rounded_only_status_gap"},
		{name: "clamped only finite", tc: parsedCase{Operation: "fma", Operands: []string{"100E+260", "0E+260", "0e+384"}, Result: "0E+369", Flags: []string{"Clamped"}, RoundingMode: "half_even"}, wantReason: "fma_clamped_status_gap"},
		{name: "rounded extra condition", tc: parsedCase{Operation: "fma", Operands: []string{"1", "1", "0"}, Result: "1", Flags: []string{"Rounded", "Invalid_operation"}, RoundingMode: "half_even"}},
		{name: "clamped extra condition", tc: parsedCase{Operation: "fma", Operands: []string{"1", "1", "0"}, Result: "1", Flags: []string{"Clamped", "Invalid_operation"}, RoundingMode: "half_even"}},
		{name: "rounded NaN operands", tc: parsedCase{Operation: "fma", Operands: []string{"NaN2", "NaN3", "NaN5"}, Result: "NaN2", Flags: []string{"Rounded"}, RoundingMode: "half_even"}},
		{name: "clamped NaN operands", tc: parsedCase{Operation: "fma", Operands: []string{"NaN2", "NaN3", "NaN5"}, Result: "NaN2", Flags: []string{"Clamped"}, RoundingMode: "half_even"}},
		{name: "rounded malformed operand", tc: parsedCase{Operation: "fma", Operands: []string{"1bogus", "1", "0"}, Result: "1", Flags: []string{"Rounded"}, RoundingMode: "half_even"}},
		{name: "clamped NaN result", tc: parsedCase{Operation: "fma", Operands: []string{"1", "1", "0"}, Result: "NaN", Flags: []string{"Clamped"}, RoundingMode: "half_even"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reason, got := generatedDectestCaseSkipReason(tc.tc, "decimal64")
			wantSkip := tc.wantReason != ""
			if got != wantSkip || (got && reason != tc.wantReason) {
				t.Fatalf("generatedDectestCaseSkipReason(%+v) = %q/%v, want %q/%v", tc.tc, reason, got, tc.wantReason, wantSkip)
			}
		})
	}
}

func TestGeneratedDectestRequiresExactScaleBStatusGapShape(t *testing.T) {
	testCases := []struct {
		name       string
		tc         parsedCase
		wantReason string
	}{
		{name: "rounded only finite", tc: parsedCase{Operation: "scaleb", Operands: []string{"1.0", "-1"}, Result: "0.10", Flags: []string{"Rounded"}}, wantReason: "scaleb_rounded_only_status_gap"},
		{name: "subnormal rounded finite", tc: parsedCase{Operation: "scaleb", Operands: []string{"1.000000000000000E-383", "-1"}, Result: "1.00000000000000E-384", Flags: []string{"Subnormal", "Rounded"}}, wantReason: "scaleb_rounded_only_status_gap"},
		{name: "clamped only finite", tc: parsedCase{Operation: "scaleb", Operands: []string{"1000E+369", "+1"}, Result: "1.0000E+373", Flags: []string{"Clamped"}}, wantReason: "scaleb_clamped_status_gap"},
		{name: "rounded extra condition", tc: parsedCase{Operation: "scaleb", Operands: []string{"1", "1"}, Result: "1E+1", Flags: []string{"Rounded", "Invalid_operation"}}},
		{name: "clamped extra condition", tc: parsedCase{Operation: "scaleb", Operands: []string{"1", "1"}, Result: "1E+1", Flags: []string{"Clamped", "Invalid_operation"}}},
		{name: "rounded NaN operand", tc: parsedCase{Operation: "scaleb", Operands: []string{"NaN", "1"}, Result: "NaN", Flags: []string{"Rounded"}}},
		{name: "clamped infinity operand", tc: parsedCase{Operation: "scaleb", Operands: []string{"Infinity", "1"}, Result: "Infinity", Flags: []string{"Clamped"}}},
		{name: "rounded malformed exponent operand", tc: parsedCase{Operation: "scaleb", Operands: []string{"1", "1bogus"}, Result: "1E+1", Flags: []string{"Rounded"}}},
		{name: "clamped NaN result", tc: parsedCase{Operation: "scaleb", Operands: []string{"1", "1"}, Result: "NaN", Flags: []string{"Clamped"}}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reason, got := generatedDectestCaseSkipReason(tc.tc, "decimal64")
			wantSkip := tc.wantReason != ""
			if got != wantSkip || (got && reason != tc.wantReason) {
				t.Fatalf("generatedDectestCaseSkipReason(%+v) = %q/%v, want %q/%v", tc.tc, reason, got, tc.wantReason, wantSkip)
			}
		})
	}
}

func TestGeneratedDectestNextTowardNaNPrecedenceSkipsOnlyDivergentIdentities(t *testing.T) {
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
			reason, got := generatedDectestCaseSkipReason(parsedCase{
				Operation: "nexttoward",
				Operands:  tc.operands,
				Result:    tc.result,
				Flags:     tc.flags,
			}, "decimal64")
			if got != tc.wantSkip {
				t.Fatalf("generatedDectestCaseSkipReason(nexttoward %v -> %s) = %q/%v, want skip %v", tc.operands, tc.result, reason, got, tc.wantSkip)
			}
			if got && reason != "nexttoward_nan_payload_precedence" {
				t.Fatalf("skip reason = %q, want nexttoward_nan_payload_precedence", reason)
			}
		})
	}
}

func TestGeneratedDectestMinMaxNaNPrecedenceSkipsOnlyDivergentIdentities(t *testing.T) {
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
			reason, got := generatedDectestCaseSkipReason(parsedCase{
				Operation: tc.operation,
				Operands:  tc.operands,
				Result:    tc.result,
				Flags:     tc.flags,
			}, "decimal64")
			if got != tc.wantSkip {
				t.Fatalf("generatedDectestCaseSkipReason(%s %v -> %s) = %q/%v, want skip %v", tc.operation, tc.operands, tc.result, reason, got, tc.wantSkip)
			}
			if got && reason != "minmax_nan_payload_precedence" {
				t.Fatalf("skip reason = %q, want minmax_nan_payload_precedence", reason)
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
			reason, got := generatedDectestCaseSkipReason(parsedCase{
				Operation: "min",
				Operands:  []string{"NaN1", "sNaN" + tc.payload},
				Result:    "NaN" + tc.payload,
				Flags:     []string{"Invalid_operation"},
			}, tc.testType)
			if got != tc.wantSkip {
				t.Fatalf("payload length %d for %s: reason=%q skip=%v, want %v", len(tc.payload), tc.testType, reason, got, tc.wantSkip)
			}
		})
	}
}
