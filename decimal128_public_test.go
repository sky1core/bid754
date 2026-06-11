package bid754

import "testing"

func TestDecimal128BIDPublicSurface(t *testing.T) {
	parse := func(input string) Decimal128BID {
		t.Helper()
		value, err := NewDecimal128BIDDirect(input)
		if err != nil {
			t.Fatalf("NewDecimal128BIDDirect(%q): %v", input, err)
		}
		return value
	}

	if got := parse("1.5").Add(parse("2.25")).String(); got != "+375E-2" {
		t.Fatalf("Add raw result = %q, want %q", got, "+375E-2")
	}
	if got := parse("10").Sub(parse("3")).String(); got != "+7E+0" {
		t.Fatalf("Sub raw result = %q, want %q", got, "+7E+0")
	}
	if got := parse("2").Mul(parse("3")).String(); got != "+6E+0" {
		t.Fatalf("Mul raw result = %q, want %q", got, "+6E+0")
	}
	if got := parse("7.5").Div(parse("2.5")).String(); got != "+3E+0" {
		t.Fatalf("Div raw result = %q, want %q", got, "+3E+0")
	}
	if got, want := parse("2.17").Quantize(parse("0.001")), parse("2.170"); got != want {
		t.Fatalf("Quantize result bits = %x, want %x", got.ToBytes(), want.ToBytes())
	}
	if got := parse("2.17").Quantize(parse("0.001")).String(); got != "+2170E-3" {
		t.Fatalf("Quantize raw string = %q, want %q", got, "+2170E-3")
	}
	if got := parse("1.5").RoundIntegralExact().String(); got != "+2E+0" {
		t.Fatalf("RoundIntegralExact raw result = %q, want %q", got, "+2E+0")
	}
	if got := parse("2.17").Quantize(parse("0.001")).PrettyString(); got != "2.17" {
		t.Fatalf("Quantize pretty string = %q, want %q", got, "2.17")
	}
}

func TestDecimal128BIDPublicHelpers(t *testing.T) {
	parse := func(input string) Decimal128BID {
		t.Helper()
		value, err := NewDecimal128BIDDirect(input)
		if err != nil {
			t.Fatalf("NewDecimal128BIDDirect(%q): %v", input, err)
		}
		return value
	}

	if got := AddSlice128BID([]Decimal128BID{parse("1"), parse("2"), parse("3")}).String(); got != "+6E+0" {
		t.Fatalf("AddSlice128BID raw result = %q, want %q", got, "+6E+0")
	}

	ctx := NewArithmeticContext()
	if got := Add128BIDWithContext(parse("1.25"), parse("0.75"), ctx).String(); got != "+200E-2" {
		t.Fatalf("Add128BIDWithContext raw result = %q, want %q", got, "+200E-2")
	}
}

func TestDecimalBIDRawParseAndConvenienceParse(t *testing.T) {
	if got, flags := ParseDecimal128BIDRaw("bogus"); !got.IsNaN() || flags != 0 {
		t.Fatalf("ParseDecimal128BIDRaw returned (%q, %v), want NaN with zero flags from raw port", got.String(), flags)
	}

	if _, err := NewDecimal128BIDDirect("bogus"); err == nil {
		t.Fatal("NewDecimal128BIDDirect should reject invalid input")
	}
}

func TestDecimalBIDNaNPayloadStringRoundTrip(t *testing.T) {
	d32, err := NewDecimal32BIDDirect("-sNaN999999")
	if err != nil {
		t.Fatalf("NewDecimal32BIDDirect payload NaN: %v", err)
	}
	if got := d32.String(); got != "-SNaN999999" {
		t.Fatalf("Decimal32BID payload string = %q, want -SNaN999999", got)
	}

	d64, err := NewDecimal64BIDDirect("NaN101")
	if err != nil {
		t.Fatalf("NewDecimal64BIDDirect payload NaN: %v", err)
	}
	if got := d64.Copy().String(); got != "+NaN101" {
		t.Fatalf("Decimal64BID copy payload string = %q, want +NaN101", got)
	}

	d128, err := NewDecimal128BIDDirect("-NaN123456789012345678901234567890123")
	if err != nil {
		t.Fatalf("NewDecimal128BIDDirect payload NaN: %v", err)
	}
	if got := d128.Abs().String(); got != "+NaN123456789012345678901234567890123" {
		t.Fatalf("Decimal128BID abs payload string = %q, want +NaN123456789012345678901234567890123", got)
	}
}
