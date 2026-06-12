package bid754

import "testing"

func TestDeterminePrecisionFromStringIgnoresExponentAndInsignificantZeros(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{input: "1e123456", want: 1},
		{input: "-000123.4500e-999", want: 5},
		{input: "0.0001000", want: 1},
		{input: "10000000", want: 1},
		{input: "12345678", want: 8},
		{input: "NaN123456789", want: 1},
		{input: "-Inf", want: 1},
	}

	for _, tc := range tests {
		if got := GetRequiredPrecision(tc.input); got != tc.want {
			t.Fatalf("GetRequiredPrecision(%q) = %d, want %d", tc.input, got, tc.want)
		}
	}
}

func TestNewDecimalWithFlagsSurfacesParseFlags(t *testing.T) {
	// Decimal32 holds 7 significant digits, so an 8-digit literal must round
	// during parsing and raise FlagInexact; an exact literal raises no flags;
	// an invalid literal returns an error.
	const overPrecise = "1.2345678" // 8 significant digits
	v32, f32, err := NewDecimal32WithFlags(overPrecise)
	if err != nil {
		t.Fatalf("NewDecimal32WithFlags(%s): %v", overPrecise, err)
	}
	if f32&FlagInexact == 0 {
		t.Fatalf("NewDecimal32WithFlags(%s) flags = %v, want FlagInexact set", overPrecise, f32)
	}
	if v32 != mustDecimal32BID(t, overPrecise) {
		t.Fatalf("NewDecimal32WithFlags(%s) value mismatch vs NewDecimal32BIDDirect", overPrecise)
	}

	if _, f64, err := NewDecimal64WithFlags("1"); err != nil || f64 != 0 {
		t.Fatalf("NewDecimal64WithFlags(1) = flags %v, err %v; want no flags, no error", f64, err)
	}

	if _, _, err := NewDecimal128WithFlags("not-a-number"); err == nil {
		t.Fatalf("NewDecimal128WithFlags(not-a-number) error = nil, want error")
	}
}

func TestParseNarrowestDecimalSelectsWidthFromCoefficientPolicy(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: "1e123", want: "Decimal32BID"},
		{input: "10000000", want: "Decimal32BID"},
		{input: "12345678", want: "Decimal64BID"},
		{input: "12345678901234567", want: "Decimal128BID"},
		{input: "Inf", want: "Decimal32BID"},
	}

	for _, tc := range tests {
		got, err := ParseNarrowestDecimal(tc.input)
		if err != nil {
			t.Fatalf("ParseNarrowestDecimal(%q): %v", tc.input, err)
		}
		if gotType := parsedDecimalType(got); gotType != tc.want {
			t.Fatalf("ParseNarrowestDecimal(%q) type = %s, want %s", tc.input, gotType, tc.want)
		}
	}
}

func TestParseNarrowestDecimalSelectsWidthForNaNPayloadPreservation(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: "NaN123456", want: "Decimal32BID"},
		{input: "NaN1234567", want: "Decimal64BID"},
		{input: "sNaN9999999999999999", want: "Decimal128BID"},
	}

	for _, tc := range tests {
		got, err := ParseNarrowestDecimal(tc.input)
		if err != nil {
			t.Fatalf("ParseNarrowestDecimal(%q): %v", tc.input, err)
		}
		if gotType := parsedDecimalType(got); gotType != tc.want {
			t.Fatalf("ParseNarrowestDecimal(%q) type = %s, want %s", tc.input, gotType, tc.want)
		}
	}
}

func TestPublicBIDConstantsMatchDocumentedLiterals(t *testing.T) {
	if got, want := Zero32BID.String(), mustDecimal32BID(t, "0").String(); got != want {
		t.Fatalf("Zero32BID = %q, want %q", got, want)
	}
	if got, want := Zero64BID.String(), mustDecimal64BID(t, "0").String(); got != want {
		t.Fatalf("Zero64BID = %q, want %q", got, want)
	}
	if got, want := Zero128BID.String(), mustDecimal128BID(t, "0").String(); got != want {
		t.Fatalf("Zero128BID = %q, want %q", got, want)
	}

	if got, want := One32BID.String(), mustDecimal32BID(t, "1").String(); got != want {
		t.Fatalf("One32BID = %q, want %q", got, want)
	}
	if got, want := One64BID.String(), mustDecimal64BID(t, "1").String(); got != want {
		t.Fatalf("One64BID = %q, want %q", got, want)
	}
	if got, want := One128BID.String(), mustDecimal128BID(t, "1").String(); got != want {
		t.Fatalf("One128BID = %q, want %q", got, want)
	}

	if got, want := Pi32BID.String(), mustDecimal32BID(t, "3.141593").String(); got != want {
		t.Fatalf("Pi32BID = %q, want %q", got, want)
	}
	if got, want := Pi64BID.String(), mustDecimal64BID(t, "3.141592653589793").String(); got != want {
		t.Fatalf("Pi64BID = %q, want %q", got, want)
	}
	if got, want := Pi128BID.String(), mustDecimal128BID(t, "3.141592653589793238462643383279503").String(); got != want {
		t.Fatalf("Pi128BID = %q, want %q", got, want)
	}

	if got, want := E32BID.String(), mustDecimal32BID(t, "2.718282").String(); got != want {
		t.Fatalf("E32BID = %q, want %q", got, want)
	}
	if got, want := E64BID.String(), mustDecimal64BID(t, "2.718281828459045").String(); got != want {
		t.Fatalf("E64BID = %q, want %q", got, want)
	}
	if got, want := E128BID.String(), mustDecimal128BID(t, "2.718281828459045235360287471352662").String(); got != want {
		t.Fatalf("E128BID = %q, want %q", got, want)
	}
}

func parsedDecimalType(v interface{}) string {
	switch v.(type) {
	case Decimal32BID:
		return "Decimal32BID"
	case Decimal64BID:
		return "Decimal64BID"
	case Decimal128BID:
		return "Decimal128BID"
	default:
		return "unknown"
	}
}
