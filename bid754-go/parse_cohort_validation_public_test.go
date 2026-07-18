package bid754

import (
	"strings"
	"testing"
)

func TestStringParsersRejectSilentCohortCoercion(t *testing.T) {
	for _, input := range []string{"1e91", "10e-102", "0e91", "0e-102", "10000000", "1000000.0"} {
		assertRejectedCohortCoercion32(t, input)
	}
	for _, input := range []string{"1e370", "10e-399", "0e370", "0e-399", "10000000000000000", "1000000000000000.0"} {
		assertRejectedCohortCoercion64(t, input)
	}
	for _, input := range []string{"1e6112", "10e-6177", "0e6112", "0e-6177", "10000000000000000000000000000000000", "1000000000000000000000000000000000.0"} {
		assertRejectedCohortCoercion128(t, input)
	}
}

func TestStringParsersAcceptRepresentableCohortBoundary(t *testing.T) {
	for _, input := range []string{
		"1" + strings.Repeat("0", decimal32Precision-2) + ".0",
		"0001" + strings.Repeat("0", decimal32Precision-2) + ".0",
		strings.Repeat("0", decimal32Precision+1) + ".0",
	} {
		assertAcceptedCohort32(t, input)
	}
	for _, input := range []string{
		"1" + strings.Repeat("0", decimal64Precision-2) + ".0",
		"0001" + strings.Repeat("0", decimal64Precision-2) + ".0",
		strings.Repeat("0", decimal64Precision+1) + ".0",
	} {
		assertAcceptedCohort64(t, input)
	}
	for _, input := range []string{
		"1" + strings.Repeat("0", decimal128Precision-2) + ".0",
		"0001" + strings.Repeat("0", decimal128Precision-2) + ".0",
		strings.Repeat("0", decimal128Precision+1) + ".0",
	} {
		assertAcceptedCohort128(t, input)
	}
}

func TestParseBIDFiniteLiteralCohort(t *testing.T) {
	tests := []struct {
		input       string
		wantQuantum int64
		wantOutside bool
		wantDigits  int
		infinity    bool
		ok          bool
	}{
		{input: "0001000.00", wantQuantum: -2, wantDigits: 6, ok: true},
		{input: "-1000000.0", wantQuantum: -1, wantDigits: 8, ok: true},
		{input: "0.000", wantQuantum: -3, wantDigits: 0, ok: true},
		{input: ".00100e+4", wantQuantum: -1, wantDigits: 3, ok: true},
		{input: "1.5e2", wantQuantum: 1, wantDigits: 2, ok: true},
		{input: "12e-3", wantQuantum: -3, wantDigits: 2, ok: true},
		{input: "1e00000000000000000000090", wantQuantum: 90, wantDigits: 1, ok: true},
		// exact int64/uint64 edges of quantum = +-exponent - fractionalDigits
		{input: "1e9223372036854775807", wantQuantum: 9223372036854775807, wantDigits: 1, ok: true},
		{input: "1e9223372036854775808", wantOutside: true, wantDigits: 1, ok: true},
		{input: ".0001e9223372036854775811", wantQuantum: 9223372036854775807, wantDigits: 1, ok: true},
		{input: ".0001e9223372036854775812", wantOutside: true, wantDigits: 1, ok: true},
		{input: "1e-9223372036854775808", wantQuantum: -9223372036854775808, wantDigits: 1, ok: true},
		{input: "1e-9223372036854775809", wantOutside: true, wantDigits: 1, ok: true},
		{input: "1.0e-9223372036854775807", wantQuantum: -9223372036854775808, wantDigits: 2, ok: true},
		{input: "1.0e-9223372036854775808", wantOutside: true, wantDigits: 2, ok: true},
		{input: "1e18446744073709551615", wantOutside: true, wantDigits: 1, ok: true},
		{input: "1e18446744073709551616", wantOutside: true, wantDigits: 1, ok: true},
		{input: "1e-18446744073709551616", wantOutside: true, wantDigits: 1, ok: true},
		{input: "Infinity", infinity: true, ok: true},
		{input: "1e"},
		{input: "1.0 "},
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got, ok := parseBIDFiniteLiteral(tc.input)
			if ok != tc.ok {
				t.Fatalf("parseBIDFiniteLiteral(%q) ok = %v, want %v", tc.input, ok, tc.ok)
			}
			if !ok {
				return
			}
			if tc.infinity {
				if !got.infinite || got.coefficientDigits != 0 {
					t.Fatalf("parseBIDFiniteLiteral(%q) = %+v, want infinity cohort", tc.input, got)
				}
				return
			}
			if got.infinite || got.quantumOutsideInt64 != tc.wantOutside || got.coefficientDigits != tc.wantDigits {
				t.Fatalf("parseBIDFiniteLiteral(%q) = %+v, want (outside=%v, digits=%d)", tc.input, got, tc.wantOutside, tc.wantDigits)
			}
			if !tc.wantOutside && got.quantum != tc.wantQuantum {
				t.Fatalf("parseBIDFiniteLiteral(%q) quantum = %d, want %d", tc.input, got.quantum, tc.wantQuantum)
			}
		})
	}
}

func assertAcceptedCohort32(t *testing.T, input string) {
	t.Helper()
	want, wantFlags, rawStatus := parseDecimal32BIDPortModeWithRawStatus(input, defaultBIDRoundingMode)
	if rawStatus != 0 || wantFlags != 0 {
		t.Fatalf("mechanical Decimal32 parse of representable cohort %q raised flags=%v raw=%#x", input, wantFlags, rawStatus)
	}
	if got, err := NewDecimal32(input); err != nil || got != want {
		t.Errorf("NewDecimal32(%q) = (%#x, %v), want (%#x, nil)", input, uint32(got), err, uint32(want))
	}
	if got, flags, err := NewDecimal32WithFlags(input); err != nil || got != want || flags != 0 {
		t.Errorf("NewDecimal32WithFlags(%q) = (%#x, %v, %v), want (%#x, 0, nil)", input, uint32(got), flags, err, uint32(want))
	}
	if got, flags, err := NewDecimal32WithMode(input, RoundNearestEven); err != nil || got != want || flags != 0 {
		t.Errorf("NewDecimal32WithMode(%q) = (%#x, %v, %v), want (%#x, 0, nil)", input, uint32(got), flags, err, uint32(want))
	}
	if got, flags := ParseDecimal32BIDRaw(input); got != want || flags != 0 {
		t.Errorf("ParseDecimal32BIDRaw(%q) = (%#x, %v), want (%#x, 0)", input, uint32(got), flags, uint32(want))
	}
}

func assertAcceptedCohort64(t *testing.T, input string) {
	t.Helper()
	want, wantFlags, rawStatus := parseDecimal64BIDPortModeWithRawStatus(input, defaultBIDRoundingMode)
	if rawStatus != 0 || wantFlags != 0 {
		t.Fatalf("mechanical Decimal64 parse of representable cohort %q raised flags=%v raw=%#x", input, wantFlags, rawStatus)
	}
	if got, err := NewDecimal64(input); err != nil || got != want {
		t.Errorf("NewDecimal64(%q) = (%#x, %v), want (%#x, nil)", input, uint64(got), err, uint64(want))
	}
	if got, flags, err := NewDecimal64WithFlags(input); err != nil || got != want || flags != 0 {
		t.Errorf("NewDecimal64WithFlags(%q) = (%#x, %v, %v), want (%#x, 0, nil)", input, uint64(got), flags, err, uint64(want))
	}
	if got, flags, err := NewDecimal64WithMode(input, RoundNearestEven); err != nil || got != want || flags != 0 {
		t.Errorf("NewDecimal64WithMode(%q) = (%#x, %v, %v), want (%#x, 0, nil)", input, uint64(got), flags, err, uint64(want))
	}
	if got, flags := ParseDecimal64BIDRaw(input); got != want || flags != 0 {
		t.Errorf("ParseDecimal64BIDRaw(%q) = (%#x, %v), want (%#x, 0)", input, uint64(got), flags, uint64(want))
	}
}

func assertAcceptedCohort128(t *testing.T, input string) {
	t.Helper()
	want, wantFlags, rawStatus := parseDecimal128BIDPortModeWithRawStatus(input, defaultBIDRoundingMode)
	if rawStatus != 0 || wantFlags != 0 {
		t.Fatalf("mechanical Decimal128 parse of representable cohort %q raised flags=%v raw=%#x", input, wantFlags, rawStatus)
	}
	if got, err := NewDecimal128(input); err != nil || got != want {
		t.Errorf("NewDecimal128(%q) = (%x, %v), want (%x, nil)", input, got.ToBytes(), err, want.ToBytes())
	}
	if got, flags, err := NewDecimal128WithFlags(input); err != nil || got != want || flags != 0 {
		t.Errorf("NewDecimal128WithFlags(%q) = (%x, %v, %v), want (%x, 0, nil)", input, got.ToBytes(), flags, err, want.ToBytes())
	}
	if got, flags, err := NewDecimal128WithMode(input, RoundNearestEven); err != nil || got != want || flags != 0 {
		t.Errorf("NewDecimal128WithMode(%q) = (%x, %v, %v), want (%x, 0, nil)", input, got.ToBytes(), flags, err, want.ToBytes())
	}
	if got, flags := ParseDecimal128BIDRaw(input); got != want || flags != 0 {
		t.Errorf("ParseDecimal128BIDRaw(%q) = (%x, %v), want (%x, 0)", input, got.ToBytes(), flags, want.ToBytes())
	}
}

func assertRejectedCohortCoercion32(t *testing.T, input string) {
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

func assertRejectedCohortCoercion64(t *testing.T, input string) {
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

func assertRejectedCohortCoercion128(t *testing.T, input string) {
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
