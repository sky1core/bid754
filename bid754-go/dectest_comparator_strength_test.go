package bid754

import "testing"

func TestGeneratedDectestAdapterRejectsUnexpectedFlagsAndConditions(t *testing.T) {
	for _, op := range []string{"class", "samequantum", "copy", "copyabs", "copynegate", "copysign"} {
		t.Run(op, func(t *testing.T) {
			flags := []string{}
			if err := runGeneratedDectestCase(decTestCase{Operation: op, Result: "1", Flags: flags}, "decimal64", -1,
				func(decTestCase, string) (decTestExecResult, error) {
					return decTestExecResult{Result: "1", Flags: FlagInvalidOperation}, nil
				}, generatedDectestCompareTokenResult, generatedDectestFlagCheckBIDFive); err == nil {
				t.Fatal("unexpected invalid flag passed")
			}
			if err := runGeneratedDectestCase(decTestCase{Operation: op, Result: "1", Flags: []string{"Unrecognized_condition"}}, "decimal64", -1,
				func(decTestCase, string) (decTestExecResult, error) {
					return decTestExecResult{Result: "1"}, nil
				}, generatedDectestCompareTokenResult, generatedDectestFlagCheckBIDFive); err == nil {
				t.Fatal("unknown expected condition passed")
			}
		})
	}
}

func TestDectestCopyPreservesOperandParseFlags(t *testing.T) {
	for _, tc := range []decTestCase{
		{Operation: "copy", Operands: []string{"1e1000"}},
		{Operation: "copysign", Operands: []string{"1", "1e1000"}},
	} {
		got, err := executeDecTestCopyOperation(tc, "decimal64")
		if err != nil {
			t.Fatal(err)
		}
		if got.Flags&FlagOverflow == 0 {
			t.Fatalf("%s lost operand overflow flag: %s", tc.Operation, got.Flags)
		}
	}
}

func TestGeneratedDectestZeroLowExponentFlagsAreCompared(t *testing.T) {
	for _, width := range []struct {
		testType                            string
		result                              string
		precision, maxExponent, minExponent int
	}{
		{"decimal32", "0E-101", 7, 96, -95},
		{"decimal64", "0E-398", 16, 384, -383},
		{"decimal128", "0E-6176", 34, 6144, -6143},
	} {
		t.Run(width.testType, func(t *testing.T) {
			tc := decTestCase{
				Operation: "toSci", Operands: []string{"0e-10000"}, Result: width.result,
				Flags: []string{"Clamped"}, RoundingMode: "half_even", Precision: width.precision,
				MaxExponent: width.maxExponent, MinExponent: width.minExponent,
			}
			got, flags, err := runDectestGoportCase(tc, width.testType)
			if err != nil {
				t.Fatal(err)
			}
			if !compareDecimalResults(tc.Result, got) || flags != 0 {
				t.Fatalf("actual port result = %q flags=%s, want zero and no flags", got, flags)
			}
			if reason, ok := dectestGoportFlagExemptReason(tc); ok {
				t.Fatalf("unexpected exemption %q", reason)
			}
			for _, injected := range []ExceptionFlags{FlagUnderflow, FlagInexact, FlagUnderflow | FlagInexact} {
				err := runGeneratedDectestCase(tc, width.testType, 1,
					func(decTestCase, string) (decTestExecResult, error) {
						return decTestExecResult{Result: got, Flags: injected}, nil
					}, generatedDectestCompareDecimalResult, generatedDectestFlagCheckBIDFive)
				if err == nil {
					t.Fatalf("generated comparator accepted injected flags %s", injected)
				}
			}
			tc.Flags = []string{"Clamped", "Unknown_condition"}
			if _, ok := dectestGoportExpectedFlags(tc); ok {
				t.Fatal("portable comparator accepted unknown condition")
			}
		})
	}
}
