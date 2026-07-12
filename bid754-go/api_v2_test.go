package bid754

import "testing"

func TestGetRequiredPrecisionIgnoresExponentAndInsignificantZeros(t *testing.T) {
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
		got, err := GetRequiredPrecision(tc.input)
		if err != nil {
			t.Fatalf("GetRequiredPrecision(%q): %v", tc.input, err)
		}
		if got != tc.want {
			t.Fatalf("GetRequiredPrecision(%q) = %d, want %d", tc.input, got, tc.want)
		}
	}
}

func TestGetRequiredPrecisionRejectsInvalidInput(t *testing.T) {
	for _, input := range []string{"", ".", "1e", "not-a-decimal", "NaNpayload"} {
		got, err := GetRequiredPrecision(input)
		if err == nil {
			t.Errorf("GetRequiredPrecision(%q) = %d, nil; want an error", input, got)
		}
		if got != 0 {
			t.Errorf("GetRequiredPrecision(%q) precision = %d, want zero on error", input, got)
		}
	}
}

func TestFromIntConvenienceConstructorsRejectInexactValues(t *testing.T) {
	for _, input := range []int32{12345678, -12345678, 2147483647, -2147483648} {
		got, err := NewDecimal32FromInt(input)
		if err == nil {
			t.Errorf("NewDecimal32FromInt(%d) = %s, nil; want an error", input, got.String())
		}
		if bits := uint32(got); bits != 0x7c000000 {
			t.Errorf("NewDecimal32FromInt(%d) error result bits = %#x, want canonical qNaN 0x7c000000", input, bits)
		}
	}

	for _, input := range []int64{12345678901234567, -12345678901234567, 9223372036854775807, -9223372036854775808} {
		got, err := NewDecimal64FromInt(input)
		if err == nil {
			t.Errorf("NewDecimal64FromInt(%d) = %s, nil; want an error", input, got.String())
		}
		if bits := uint64(got); bits != 0x7c00000000000000 {
			t.Errorf("NewDecimal64FromInt(%d) error result bits = %#x, want canonical qNaN 0x7c00000000000000", input, bits)
		}
	}
}

func TestFromIntConvenienceConstructorsAcceptExactValues(t *testing.T) {
	for _, input := range []int32{0, 1234567, -1234567, 10000000, -10000000} {
		if _, err := NewDecimal32FromInt(input); err != nil {
			t.Errorf("NewDecimal32FromInt(%d): unexpected error: %v", input, err)
		}
	}

	for _, input := range []int64{0, 1234567890123456, -1234567890123456, 10000000000000000, -10000000000000000} {
		if _, err := NewDecimal64FromInt(input); err != nil {
			t.Errorf("NewDecimal64FromInt(%d): unexpected error: %v", input, err)
		}
	}

	for _, input := range []int64{0, 9223372036854775807, -9223372036854775808} {
		if _, err := NewDecimal128FromInt(input); err != nil {
			t.Errorf("NewDecimal128FromInt(%d): unexpected error: %v", input, err)
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
	raw32, rawFlags32 := ParseDecimal32BIDRaw(overPrecise)
	if v32 != raw32 || f32 != rawFlags32 {
		t.Fatalf("NewDecimal32WithFlags(%s) = (%s, %v), want raw result (%s, %v)", overPrecise, v32.String(), f32, raw32.String(), rawFlags32)
	}

	if _, f64, err := NewDecimal64WithFlags("1"); err != nil || f64 != 0 {
		t.Fatalf("NewDecimal64WithFlags(1) = flags %v, err %v; want no flags, no error", f64, err)
	}

	if _, _, err := NewDecimal128WithFlags("not-a-number"); err == nil {
		t.Fatalf("NewDecimal128WithFlags(not-a-number) error = nil, want error")
	}
}

func TestErrorOnlyStringConstructorsRejectLossyFiniteInputs(t *testing.T) {
	t.Run("decimal32", func(t *testing.T) {
		for _, tc := range []struct {
			name      string
			input     string
			wantFlags ExceptionFlags
		}{
			{name: "precision", input: "1.2345678", wantFlags: FlagInexact},
			{name: "overflow", input: "1e97", wantFlags: FlagOverflow | FlagInexact},
			{name: "underflow", input: "1e-102", wantFlags: FlagUnderflow | FlagInexact},
		} {
			t.Run(tc.name, func(t *testing.T) {
				assertLossyDecimal32String(t, tc.input, tc.wantFlags)
			})
		}
	})

	t.Run("decimal64", func(t *testing.T) {
		for _, tc := range []struct {
			name      string
			input     string
			wantFlags ExceptionFlags
		}{
			{name: "precision", input: "1.2345678901234567", wantFlags: FlagInexact},
			{name: "overflow", input: "1e385", wantFlags: FlagOverflow | FlagInexact},
			{name: "underflow", input: "1e-399", wantFlags: FlagUnderflow | FlagInexact},
		} {
			t.Run(tc.name, func(t *testing.T) {
				assertLossyDecimal64String(t, tc.input, tc.wantFlags)
			})
		}
	})

	t.Run("decimal128", func(t *testing.T) {
		for _, tc := range []struct {
			name      string
			input     string
			wantFlags ExceptionFlags
		}{
			{name: "precision", input: "1.2345678901234567890123456789012345", wantFlags: FlagInexact},
			{name: "overflow", input: "1e6145", wantFlags: FlagOverflow | FlagInexact},
			{name: "underflow", input: "1e-6177", wantFlags: FlagUnderflow | FlagInexact},
		} {
			t.Run(tc.name, func(t *testing.T) {
				assertLossyDecimal128String(t, tc.input, tc.wantFlags)
			})
		}
	})
}

func assertLossyDecimal32String(t *testing.T, input string, wantFlags ExceptionFlags) {
	t.Helper()
	if got, err := NewDecimal32(input); err == nil || got != 0 {
		t.Errorf("NewDecimal32(%q) = (%#x, %v), want zero and error", input, uint32(got), err)
	}
	if got, err := NewDecimal32BIDDirect(input); err == nil || got != 0 {
		t.Errorf("NewDecimal32BIDDirect(%q) = (%#x, %v), want zero and error", input, uint32(got), err)
	}
	raw, rawFlags := ParseDecimal32BIDRaw(input)
	if rawFlags&wantFlags != wantFlags {
		t.Fatalf("ParseDecimal32BIDRaw(%q) flags = %v, want at least %v", input, rawFlags, wantFlags)
	}
	withFlags, flags, err := NewDecimal32WithFlags(input)
	if err != nil || withFlags != raw || flags != rawFlags {
		t.Errorf("NewDecimal32WithFlags(%q) = (%s, %v, %v), want raw result (%s, %v, nil)", input, withFlags.String(), flags, err, raw.String(), rawFlags)
	}
	withMode, modeFlags, err := NewDecimal32WithMode(input, RoundNearestEven)
	if err != nil || withMode != raw || modeFlags != rawFlags {
		t.Errorf("NewDecimal32WithMode(%q) = (%s, %v, %v), want raw result (%s, %v, nil)", input, withMode.String(), modeFlags, err, raw.String(), rawFlags)
	}
}

func assertLossyDecimal64String(t *testing.T, input string, wantFlags ExceptionFlags) {
	t.Helper()
	if got, err := NewDecimal64(input); err == nil || got != 0 {
		t.Errorf("NewDecimal64(%q) = (%#x, %v), want zero and error", input, uint64(got), err)
	}
	if got, err := NewDecimal64BIDDirect(input); err == nil || got != 0 {
		t.Errorf("NewDecimal64BIDDirect(%q) = (%#x, %v), want zero and error", input, uint64(got), err)
	}
	raw, rawFlags := ParseDecimal64BIDRaw(input)
	if rawFlags&wantFlags != wantFlags {
		t.Fatalf("ParseDecimal64BIDRaw(%q) flags = %v, want at least %v", input, rawFlags, wantFlags)
	}
	withFlags, flags, err := NewDecimal64WithFlags(input)
	if err != nil || withFlags != raw || flags != rawFlags {
		t.Errorf("NewDecimal64WithFlags(%q) = (%s, %v, %v), want raw result (%s, %v, nil)", input, withFlags.String(), flags, err, raw.String(), rawFlags)
	}
	withMode, modeFlags, err := NewDecimal64WithMode(input, RoundNearestEven)
	if err != nil || withMode != raw || modeFlags != rawFlags {
		t.Errorf("NewDecimal64WithMode(%q) = (%s, %v, %v), want raw result (%s, %v, nil)", input, withMode.String(), modeFlags, err, raw.String(), rawFlags)
	}
}

func assertLossyDecimal128String(t *testing.T, input string, wantFlags ExceptionFlags) {
	t.Helper()
	if got, err := NewDecimal128(input); err == nil || got != (Decimal128BID{}) {
		t.Errorf("NewDecimal128(%q) = (%x, %v), want zero and error", input, got.ToBytes(), err)
	}
	if got, err := NewDecimal128BIDDirect(input); err == nil || got != (Decimal128BID{}) {
		t.Errorf("NewDecimal128BIDDirect(%q) = (%x, %v), want zero and error", input, got.ToBytes(), err)
	}
	raw, rawFlags := ParseDecimal128BIDRaw(input)
	if rawFlags&wantFlags != wantFlags {
		t.Fatalf("ParseDecimal128BIDRaw(%q) flags = %v, want at least %v", input, rawFlags, wantFlags)
	}
	withFlags, flags, err := NewDecimal128WithFlags(input)
	if err != nil || withFlags != raw || flags != rawFlags {
		t.Errorf("NewDecimal128WithFlags(%q) = (%s, %v, %v), want raw result (%s, %v, nil)", input, withFlags.String(), flags, err, raw.String(), rawFlags)
	}
	withMode, modeFlags, err := NewDecimal128WithMode(input, RoundNearestEven)
	if err != nil || withMode != raw || modeFlags != rawFlags {
		t.Errorf("NewDecimal128WithMode(%q) = (%s, %v, %v), want raw result (%s, %v, nil)", input, withMode.String(), modeFlags, err, raw.String(), rawFlags)
	}
}

func TestNaNPayloadRangeNeverDropsPayloadSilently(t *testing.T) {
	for _, input := range []string{
		"NaN1000000",
		"NaN1000000000000000",
		"NaN1000000000000000000000000000000000",
	} {
		assertRejectedNaNPayload32(t, input)
	}

	if got, err := NewDecimal64("NaN1000000"); err != nil || got.String() != "+NaN1000000" {
		t.Errorf("NewDecimal64(NaN1000000) = (%s, %v), want preserved payload", got.String(), err)
	}
	for _, input := range []string{
		"NaN1000000000000000",
		"NaN1000000000000000000000000000000000",
	} {
		assertRejectedNaNPayload64(t, input)
	}

	if got, err := NewDecimal128("NaN1000000000000000"); err != nil || got.String() != "+NaN1000000000000000" {
		t.Errorf("NewDecimal128(NaN1000000000000000) = (%s, %v), want preserved payload", got.String(), err)
	}
	assertRejectedNaNPayload128(t, "NaN1000000000000000000000000000000000")
}

func assertRejectedNaNPayload32(t *testing.T, input string) {
	t.Helper()
	if got, err := NewDecimal32(input); err == nil || got != 0 {
		t.Errorf("NewDecimal32(%q) = (%#x, %v), want zero and error", input, uint32(got), err)
	}
	if got, flags, err := NewDecimal32WithFlags(input); err == nil || got != 0 || flags != 0 {
		t.Errorf("NewDecimal32WithFlags(%q) = (%#x, %v, %v), want zero, zero flags, error", input, uint32(got), flags, err)
	}
	if got, flags, err := NewDecimal32WithMode(input, RoundNearestEven); err == nil || got != 0 || flags != 0 {
		t.Errorf("NewDecimal32WithMode(%q) = (%#x, %v, %v), want zero, zero flags, error", input, uint32(got), flags, err)
	}
	got, flags := ParseDecimal32BIDRaw(input)
	if uint32(got) != 0x7c000000 || flags != FlagInvalidOperation {
		t.Errorf("ParseDecimal32BIDRaw(%q) = (%#x, %v), want canonical qNaN and FlagInvalidOperation", input, uint32(got), flags)
	}
}

func assertRejectedNaNPayload64(t *testing.T, input string) {
	t.Helper()
	if got, err := NewDecimal64(input); err == nil || got != 0 {
		t.Errorf("NewDecimal64(%q) = (%#x, %v), want zero and error", input, uint64(got), err)
	}
	if got, flags, err := NewDecimal64WithFlags(input); err == nil || got != 0 || flags != 0 {
		t.Errorf("NewDecimal64WithFlags(%q) = (%#x, %v, %v), want zero, zero flags, error", input, uint64(got), flags, err)
	}
	if got, flags, err := NewDecimal64WithMode(input, RoundNearestEven); err == nil || got != 0 || flags != 0 {
		t.Errorf("NewDecimal64WithMode(%q) = (%#x, %v, %v), want zero, zero flags, error", input, uint64(got), flags, err)
	}
	got, flags := ParseDecimal64BIDRaw(input)
	if uint64(got) != 0x7c00000000000000 || flags != FlagInvalidOperation {
		t.Errorf("ParseDecimal64BIDRaw(%q) = (%#x, %v), want canonical qNaN and FlagInvalidOperation", input, uint64(got), flags)
	}
}

func assertRejectedNaNPayload128(t *testing.T, input string) {
	t.Helper()
	if got, err := NewDecimal128(input); err == nil || got != (Decimal128BID{}) {
		t.Errorf("NewDecimal128(%q) = (%x, %v), want zero and error", input, got.ToBytes(), err)
	}
	if got, flags, err := NewDecimal128WithFlags(input); err == nil || got != (Decimal128BID{}) || flags != 0 {
		t.Errorf("NewDecimal128WithFlags(%q) = (%x, %v, %v), want zero, zero flags, error", input, got.ToBytes(), flags, err)
	}
	if got, flags, err := NewDecimal128WithMode(input, RoundNearestEven); err == nil || got != (Decimal128BID{}) || flags != 0 {
		t.Errorf("NewDecimal128WithMode(%q) = (%x, %v, %v), want zero, zero flags, error", input, got.ToBytes(), flags, err)
	}
	got, flags := ParseDecimal128BIDRaw(input)
	if got != canonicalQNaN128BID() || flags != FlagInvalidOperation {
		t.Errorf("ParseDecimal128BIDRaw(%q) = (%x, %v), want canonical qNaN and FlagInvalidOperation", input, got.ToBytes(), flags)
	}
}

func TestPublicBIDConstantsMatchDocumentedLiterals(t *testing.T) {
	if got, want := Zero32BID().String(), mustDecimal32BID(t, "0").String(); got != want {
		t.Fatalf("Zero32BID = %q, want %q", got, want)
	}
	if got, want := Zero64BID().String(), mustDecimal64BID(t, "0").String(); got != want {
		t.Fatalf("Zero64BID = %q, want %q", got, want)
	}
	if got, want := Zero128BID().String(), mustDecimal128BID(t, "0").String(); got != want {
		t.Fatalf("Zero128BID = %q, want %q", got, want)
	}

	if got, want := One32BID().String(), mustDecimal32BID(t, "1").String(); got != want {
		t.Fatalf("One32BID = %q, want %q", got, want)
	}
	if got, want := One64BID().String(), mustDecimal64BID(t, "1").String(); got != want {
		t.Fatalf("One64BID = %q, want %q", got, want)
	}
	if got, want := One128BID().String(), mustDecimal128BID(t, "1").String(); got != want {
		t.Fatalf("One128BID = %q, want %q", got, want)
	}

	if got, want := Pi32BID().String(), mustDecimal32BID(t, "3.141593").String(); got != want {
		t.Fatalf("Pi32BID = %q, want %q", got, want)
	}
	if got, want := Pi64BID().String(), mustDecimal64BID(t, "3.141592653589793").String(); got != want {
		t.Fatalf("Pi64BID = %q, want %q", got, want)
	}
	if got, want := Pi128BID().String(), mustDecimal128BID(t, "3.141592653589793238462643383279503").String(); got != want {
		t.Fatalf("Pi128BID = %q, want %q", got, want)
	}

	if got, want := E32BID().String(), mustDecimal32BID(t, "2.718282").String(); got != want {
		t.Fatalf("E32BID = %q, want %q", got, want)
	}
	if got, want := E64BID().String(), mustDecimal64BID(t, "2.718281828459045").String(); got != want {
		t.Fatalf("E64BID = %q, want %q", got, want)
	}
	if got, want := E128BID().String(), mustDecimal128BID(t, "2.718281828459045235360287471352662").String(); got != want {
		t.Fatalf("E128BID = %q, want %q", got, want)
	}
}
