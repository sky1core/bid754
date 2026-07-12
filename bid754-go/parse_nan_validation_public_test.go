package bid754

import "testing"

// rejectedNonLiteralNaNInputs are inputs the mechanical-port parser maps to
// NaN even though they do not spell a NaN literal. The public constructors
// must reject them with an error.
var rejectedNonLiteralNaNInputs = []string{
	"banana",
	"nanoseconds", // "nan" prefix, non-digit payload
	"NANO",        // "nan" prefix, non-digit payload
	"finance",     // contains "nan" substring only
	"12banana",
	"nan1x",    // payload "1x" is not all digits
	"nan(123)", // no parenthesized-payload grammar in the Intel port
	"'NaN5'",   // decTest quoting is not public-API syntax
	"\"NaN\"",  // decTest quoting is not public-API syntax
	"snanfoo",  // non-digit payload
	"qnan-1",   // signs are not allowed inside the payload
	"nan 1",    // whitespace inside the payload
	"  NaN  ",  // trailing whitespace (leading alone would be trimmed)
	" nan ",    // trailing whitespace (leading alone would be trimmed)
	"nan ",     // trailing whitespace
	"+ nan",    // whitespace after the sign: only LEADING whitespace is trimmed
	"1.2.3",    // malformed number (two radix points)
	"12x",      // malformed number (junk after digits)
	"--1",      // malformed sign
	"",         // empty
	"   ",      // whitespace only
	// Trailing-whitespace contrast probes for the numeric/infinity paths:
	// the NaN literal grammar must reject "nan " for the same reason the
	// port rejects these (whitespace is trimmed at the start only).
	"1 ",   // trailing whitespace after digits
	"1\t",  // trailing tab after digits
	"inf ", // trailing whitespace after infinity
	"+ 1",  // whitespace after the sign
	".",    // radix point without a significand digit
	"+.",   // signed radix point without a significand digit
	"-.",   // signed radix point without a significand digit
	".e1",  // exponent digits do not replace a significand digit
	"1e5x", // exponent text must consume the complete input
	// A syntactically valid NaN whose payload cannot be represented by any
	// supported BID width must not be accepted with its payload discarded.
	"NaN99999999999999999999999999999999999999",
}

// acceptedDecimalInputs are inputs every public constructor must accept.
var acceptedDecimalInputs = []string{
	// NaN literal variants recognized by parseBIDNaNLiteral.
	"NaN",
	"nan",
	"+NaN",
	"-nan",
	"SNaN",
	"-sNaN",
	"snan",
	"qNaN",
	"QNAN7",
	"NaN123",
	"snan42",
	"NaN000", // all-zero payload
	// Leading space/tab: trimmed by the NaN literal grammar exactly as the
	// Intel bid<w>_from_string port trims it for numeric/infinity input.
	" nan",
	"\tnan",
	"  -NaN7",
	// Infinity spellings.
	"Inf",
	"+inf",
	"-Infinity",
	"INFINITY",
	// Leading-whitespace contrast probes for the numeric/infinity port
	// paths: these must stay accepted so the NaN literal grammar's leading
	// trim cannot drift apart from the port again (the prior asymmetry).
	" 1",
	"\t1",
	" inf",
	" +1",
	// Ordinary numbers.
	"0",
	"1.5",
	"-0.00",
	"12E5",
	"-7e-3",
	".5",
}

func TestNewDecimalRejectsNonLiteralNaNInputs(t *testing.T) {
	for _, s := range rejectedNonLiteralNaNInputs {
		if _, err := NewDecimal32(s); err == nil {
			t.Errorf("NewDecimal32(%q): expected error, got nil", s)
		}
		if _, _, err := NewDecimal32WithFlags(s); err == nil {
			t.Errorf("NewDecimal32WithFlags(%q): expected error, got nil", s)
		}
		if _, _, err := NewDecimal32WithMode(s, RoundNearestEven); err == nil {
			t.Errorf("NewDecimal32WithMode(%q): expected error, got nil", s)
		}
		if _, err := NewDecimal64(s); err == nil {
			t.Errorf("NewDecimal64(%q): expected error, got nil", s)
		}
		if _, _, err := NewDecimal64WithFlags(s); err == nil {
			t.Errorf("NewDecimal64WithFlags(%q): expected error, got nil", s)
		}
		if _, _, err := NewDecimal64WithMode(s, RoundNearestEven); err == nil {
			t.Errorf("NewDecimal64WithMode(%q): expected error, got nil", s)
		}
		if _, err := NewDecimal128(s); err == nil {
			t.Errorf("NewDecimal128(%q): expected error, got nil", s)
		}
		if _, _, err := NewDecimal128WithFlags(s); err == nil {
			t.Errorf("NewDecimal128WithFlags(%q): expected error, got nil", s)
		}
		if _, _, err := NewDecimal128WithMode(s, RoundNearestEven); err == nil {
			t.Errorf("NewDecimal128WithMode(%q): expected error, got nil", s)
		}
	}
}

func TestNewDecimalAcceptsLiteralInputs(t *testing.T) {
	for _, s := range acceptedDecimalInputs {
		if _, err := NewDecimal32(s); err != nil {
			t.Errorf("NewDecimal32(%q): unexpected error: %v", s, err)
		}
		if _, _, err := NewDecimal32WithFlags(s); err != nil {
			t.Errorf("NewDecimal32WithFlags(%q): unexpected error: %v", s, err)
		}
		if _, _, err := NewDecimal32WithMode(s, RoundNearestEven); err != nil {
			t.Errorf("NewDecimal32WithMode(%q): unexpected error: %v", s, err)
		}
		if _, err := NewDecimal64(s); err != nil {
			t.Errorf("NewDecimal64(%q): unexpected error: %v", s, err)
		}
		if _, _, err := NewDecimal64WithFlags(s); err != nil {
			t.Errorf("NewDecimal64WithFlags(%q): unexpected error: %v", s, err)
		}
		if _, _, err := NewDecimal64WithMode(s, RoundNearestEven); err != nil {
			t.Errorf("NewDecimal64WithMode(%q): unexpected error: %v", s, err)
		}
		if _, err := NewDecimal128(s); err != nil {
			t.Errorf("NewDecimal128(%q): unexpected error: %v", s, err)
		}
		if _, _, err := NewDecimal128WithFlags(s); err != nil {
			t.Errorf("NewDecimal128WithFlags(%q): unexpected error: %v", s, err)
		}
		if _, _, err := NewDecimal128WithMode(s, RoundNearestEven); err != nil {
			t.Errorf("NewDecimal128WithMode(%q): unexpected error: %v", s, err)
		}
	}
}

func TestIsValidDecimalStringMatchesConstructorValidation(t *testing.T) {
	for _, s := range rejectedNonLiteralNaNInputs {
		if IsValidDecimalString(s) {
			t.Errorf("IsValidDecimalString(%q) = true, want false", s)
		}
	}
	for _, s := range acceptedDecimalInputs {
		if !IsValidDecimalString(s) {
			t.Errorf("IsValidDecimalString(%q) = false, want true", s)
		}
	}
}

// Intel's Decimal128 from_string stops after the valid exponent prefix in
// "1e5x". Every public parser must reject the complete malformed literal;
// ParseDecimal128BIDRaw uses its declared flag channel while the generated
// readtest path continues to verify the unchanged mechanical-port behavior.
func TestExponentTrailingJunkRejectedAcrossPublicParsers(t *testing.T) {
	if IsValidDecimalString("1e5x") {
		t.Error(`IsValidDecimalString("1e5x") = true, want false`)
	}

	got, gotFlags := ParseDecimal128BIDRaw("1e5x")
	if got != canonicalQNaN128BID() || gotFlags != FlagInvalidOperation {
		t.Errorf(`ParseDecimal128BIDRaw("1e5x") = (%s, %v), want canonical qNaN and InvalidOperation`, got.String(), gotFlags)
	}
}

func TestNaNLiteralParseResultsStayNaN(t *testing.T) {
	cases := []struct {
		input     string
		signaling bool
		str64     string
	}{
		{"NaN", false, "+NaN"},
		{"-nan", false, "-NaN"},
		{"SNaN", true, "+SNaN"},
		{"NaN123", false, "+NaN123"},
		{"snan42", true, "+SNaN42"},
	}
	for _, tc := range cases {
		d, err := NewDecimal64(tc.input)
		if err != nil {
			t.Errorf("NewDecimal64(%q): unexpected error: %v", tc.input, err)
			continue
		}
		if !d.IsNaN() {
			t.Errorf("NewDecimal64(%q): expected NaN result", tc.input)
		}
		if got := d.IsSignaling(); got != tc.signaling {
			t.Errorf("NewDecimal64(%q).IsSignaling() = %v, want %v", tc.input, got, tc.signaling)
		}
		if got := d.String(); got != tc.str64 {
			t.Errorf("NewDecimal64(%q).String() = %q, want %q", tc.input, got, tc.str64)
		}
	}
}
