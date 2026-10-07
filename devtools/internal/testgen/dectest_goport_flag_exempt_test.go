package testgen

import "testing"

func TestGeneratedDectestGoportFlagExemptReason(t *testing.T) {
	for _, tc := range []parsedCase{
		{Operation: "toSci", Operands: []string{"0e-10000"}, Result: "0E-398", Flags: []string{"Clamped"}},
		{Operation: "toEng", Operands: []string{"-0e-10000"}, Result: "-0E-6176", Flags: []string{"Clamped"}},
		{Operation: "toSci", Operands: []string{"0e-10000"}, Result: "0E-398", Flags: []string{"Clamped", "Unknown_condition"}},
	} {
		if reason, ok := generatedDectestGoportFlagExemptReason(tc); ok || reason != "" {
			t.Errorf("%s %v: exemption = %q/%v, want empty/false", tc.Operation, tc.Flags, reason, ok)
		}
	}
}
