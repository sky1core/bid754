package testgen

import "testing"

// TestGeneratedDectestGoportFlagExemptReason pins the goport Phase 2
// flag-exemption classifier's signature and its strict behavior with
// synthetic cases. The classifier must classify exactly the measured
// divergence shape — a tosci/toeng zero result low-clamped on parse with
// clamped-only Conditions — and must never classify a case that carries an
// unrecognized condition token (exemption waives only the flag comparison,
// never the token validation an executed case performs before exemption).
func TestGeneratedDectestGoportFlagExemptReason(t *testing.T) {
	exemptReason := "from_string_zero_low_clamp_divergence"
	cases := []struct {
		name   string
		tc     parsedCase
		exempt bool
	}{
		{
			name: "tosci low-side zero clamp is exempt",
			tc: parsedCase{
				Operation: "toSci", Operands: []string{"0e-10000"},
				Result: "0E-398", Flags: []string{"Clamped"},
			},
			exempt: true,
		},
		{
			name: "toeng negative low-side zero clamp is exempt",
			tc: parsedCase{
				Operation: "toEng", Operands: []string{"-0e-10000"},
				Result: "-0E-6176", Flags: []string{"Clamped"},
			},
			exempt: true,
		},
		{
			name: "high-side zero clamp is NOT exempt (measured flag-quiet on both sides)",
			tc: parsedCase{
				Operation: "toSci", Operands: []string{"0e+10000"},
				Result: "0E+369", Flags: []string{"Clamped"},
			},
			exempt: false,
		},
		{
			name: "arithmetic zero clamp is NOT exempt (divide, measured flag-quiet)",
			tc: parsedCase{
				Operation: "divide", Operands: []string{"0E-390", "1000E+13"},
				Result: "0E-398", Flags: []string{"Clamped"},
			},
			exempt: false,
		},
		{
			name: "arithmetic zero clamp is NOT exempt (multiply, measured flag-quiet)",
			tc: parsedCase{
				Operation: "multiply", Operands: []string{"0E-260", "1000E-260"},
				Result: "0E-398", Flags: []string{"Clamped"},
			},
			exempt: false,
		},
		{
			name: "five-surface condition alongside clamped is NOT exempt",
			tc: parsedCase{
				Operation: "toSci", Operands: []string{"0e-10000"},
				Result: "0E-398", Flags: []string{"Clamped", "Underflow"},
			},
			exempt: false,
		},
		{
			name: "insufficient-storage invalid condition alongside clamped is NOT exempt",
			tc: parsedCase{
				Operation: "toSci", Operands: []string{"0e-10000"},
				Result: "0E-398", Flags: []string{"Clamped", "Insufficient_storage"},
			},
			exempt: false,
		},
		{
			name: "conversion-syntax invalid condition alongside clamped is NOT exempt",
			tc: parsedCase{
				Operation: "toSci", Operands: []string{"0e-10000"},
				Result: "0E-398", Flags: []string{"Clamped", "Conversion_syntax"},
			},
			exempt: false,
		},
		{
			name: "rounded condition alongside clamped is NOT exempt",
			tc: parsedCase{
				Operation: "toSci", Operands: []string{"0e-10000"},
				Result: "0E-398", Flags: []string{"Clamped", "Rounded"},
			},
			exempt: false,
		},
		{
			name: "subnormal condition alongside clamped is NOT exempt",
			tc: parsedCase{
				Operation: "toSci", Operands: []string{"0e-10000"},
				Result: "0E-398", Flags: []string{"Clamped", "Subnormal"},
			},
			exempt: false,
		},
		{
			name: "non-zero result is NOT exempt",
			tc: parsedCase{
				Operation: "toSci", Operands: []string{"1e-10000"},
				Result: "1E-398", Flags: []string{"Clamped"},
			},
			exempt: false,
		},
		{
			name: "strict: unrecognized condition token is NEVER exempt",
			tc: parsedCase{
				Operation: "toSci", Operands: []string{"0e-10000"},
				Result: "0E-398", Flags: []string{"Clamped", "Totally_bogus_condition"},
			},
			exempt: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reason, ok := generatedDectestGoportFlagExemptReason(tc.tc)
			if ok != tc.exempt {
				t.Fatalf("exempt = %v (reason %q), want %v", ok, reason, tc.exempt)
			}
			if ok && reason != exemptReason {
				t.Fatalf("reason = %q, want %q", reason, exemptReason)
			}
		})
	}
}

// TestGeneratedDectestGoportRecognizedConditions pins the strict token
// gate itself: the recognized set must match the generated runner's
// parseDecTestFlags mapping (bid754-go/dectest_driver.go), and anything
// outside it must refuse classification.
func TestGeneratedDectestGoportRecognizedConditions(t *testing.T) {
	recognized := []string{
		"", "none", "no_flags", "Inexact", "Underflow", "Overflow", "Division_by_zero",
		"Invalid_operation", "Division_undefined", "Division_impossible",
		"Insufficient_storage", "Conversion_syntax", "Subnormal", "Rounded", "Clamped",
	}
	for _, token := range recognized {
		if !generatedDectestGoportRecognizedConditions([]string{token}) {
			t.Errorf("token %q should be recognized (parseDecTestFlags lockstep)", token)
		}
	}
	for _, token := range []string{"Lost_digits", "bogus", "under flowx"} {
		if generatedDectestGoportRecognizedConditions([]string{token}) {
			t.Errorf("token %q must not be recognized", token)
		}
	}
	if generatedDectestGoportRecognizedConditions([]string{"Clamped", "bogus"}) {
		t.Error("a recognized token must not mask an unrecognized one in the same case")
	}
}
