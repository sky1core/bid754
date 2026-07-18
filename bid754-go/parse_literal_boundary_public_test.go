package bid754

import (
	"strings"
	"testing"
)

// This file pins the public string-parse acceptance/rejection boundary of all
// five per-width entrypoints (NewDecimal*BIDDirect, NewDecimal*WithFlags,
// NewDecimal*WithMode, ParseDecimal*BIDRaw) as observed bit+flag results, so
// any plumbing rewrite of the parse path (P-1 de-allocation) can be proven
// decision-equivalent at the extreme boundaries: quantum clamp thresholds ±1,
// coefficient digit-count limits, exponent digit-string magnitudes across the
// int32/int64/uint64 representability edges, huge fractional digit counts
// canceling huge exponents, inexact rounding ties, sign/zero/subnormal edges,
// NaN-literal payload limits, and syntax/whitespace error edges. The pinned
// values were recorded from the pre-rewrite implementation and must never
// change without an explicit public-contract decision.

type parseBoundaryRow struct {
	name  string
	input string
}

func parseBoundaryRows() []parseBoundaryRow {
	rows := []parseBoundaryRow{
		// zeros, signs, and zero-cohort quantum clamp thresholds +-1
		{name: "0", input: "0"},
		{name: "-0", input: "-0"},
		{name: "+0.00", input: "+0.00"},
		{name: "0e0", input: "0e0"},
		{name: "0e90", input: "0e90"},
		{name: "0e91", input: "0e91"},
		{name: "0e-101", input: "0e-101"},
		{name: "0e-102", input: "0e-102"},
		{name: "0e369", input: "0e369"},
		{name: "0e370", input: "0e370"},
		{name: "0e-398", input: "0e-398"},
		{name: "0e-399", input: "0e-399"},
		{name: "0e6111", input: "0e6111"},
		{name: "0e6112", input: "0e6112"},
		{name: "-0e-6176", input: "-0e-6176"},
		{name: "0e-6177", input: "0e-6177"},
		// coefficient digit-count boundaries (leading zeros are free,
		// trailing zeros consume precision)
		{name: "9999999", input: "9999999"},
		{name: "99999999", input: "99999999"},
		{name: "0009999999", input: "0009999999"},
		{name: "10000000", input: "10000000"},
		{name: "1000000.0", input: "1000000.0"},
		{name: "9999999999999999", input: "9999999999999999"},
		{name: "99999999999999999", input: "99999999999999999"},
		{name: "10000000000000000", input: "10000000000000000"},
		{name: "34x9", input: strings.Repeat("9", 34)},
		{name: "35x9", input: strings.Repeat("9", 35)},
		{name: "1_33x0", input: "1" + strings.Repeat("0", 33)},
		// nonzero finite exponent clamp thresholds +-1
		{name: "1e90", input: "1e90"},
		{name: "1e91", input: "1e91"},
		{name: "9.999999e96", input: "9.999999e96"},
		{name: "1e97", input: "1e97"},
		{name: "1e-101", input: "1e-101"},
		{name: "1e-102", input: "1e-102"},
		{name: "0.000001e-95", input: "0.000001e-95"},
		{name: "9.999999999999999e384", input: "9.999999999999999e384"},
		{name: "1e385", input: "1e385"},
		{name: "1e-398", input: "1e-398"},
		{name: "1e-399", input: "1e-399"},
		{name: "d128_max_normal", input: "9." + strings.Repeat("9", 33) + "e6144"},
		{name: "1e6145", input: "1e6145"},
		{name: "1e-6176", input: "1e-6176"},
		{name: "1e-6177", input: "1e-6177"},
		// exponent digit-string magnitudes across int32/int64/uint64 edges
		{name: "1e2147483647", input: "1e2147483647"},
		{name: "1e2147483648", input: "1e2147483648"},
		{name: "1e-2147483648", input: "1e-2147483648"},
		{name: "1e-2147483649", input: "1e-2147483649"},
		{name: "1e_int64max", input: "1e9223372036854775807"},
		{name: "1e_int64max+1", input: "1e9223372036854775808"},
		{name: "1e_int64min", input: "1e-9223372036854775808"},
		{name: "1e_int64min-1", input: "1e-9223372036854775809"},
		{name: "1e_uint64max", input: "1e18446744073709551615"},
		{name: "1e_uint64max+1", input: "1e18446744073709551616"},
		{name: "1e_-uint64max", input: "1e-18446744073709551615"},
		{name: "1e_-uint64max-1", input: "1e-18446744073709551616"},
		{name: "1e_30nines", input: "1e" + strings.Repeat("9", 30)},
		{name: "1e_-30nines", input: "1e-" + strings.Repeat("9", 30)},
		{name: "1e_zeropad90", input: "1e00000000000000000000090"},
		// zero coefficient with an exponent at the int64/uint64 edges: the
		// port parses these with exact status, so the written-cohort quantum
		// arithmetic (not port flags) makes the accept/reject decision
		{name: "0e_int64max", input: "0e9223372036854775807"},
		{name: "0e_int64max+1", input: "0e9223372036854775808"},
		{name: "0e_int64min", input: "0e-9223372036854775808"},
		{name: "0e_int64min-1", input: "0e-9223372036854775809"},
		{name: "0e_uint64max+1", input: "0e18446744073709551616"},
		// huge fractional digit count canceling a huge exponent
		{name: "frac_cancel_in_range", input: "0." + strings.Repeat("0", 299) + "1e250"},
		{name: "frac_cancel_at_int64max", input: "0." + strings.Repeat("0", 199) + "1e9223372036854775807"},
		{name: "frac_cancel_past_int64", input: "0." + strings.Repeat("0", 199) + "1e9223372036854775808"},
		{name: "frac_cancel_past_uint64", input: "0." + strings.Repeat("0", 199) + "1e18446744073709551616"},
		{name: "frac_cancel_zero_coeff", input: "0." + strings.Repeat("0", 199) + "e9223372036854775808"},
		// inexact rounding ties at each width's precision limit
		{name: "1.0000005", input: "1.0000005"},
		{name: "-1.0000005", input: "-1.0000005"},
		{name: "9999999.5", input: "9999999.5"},
		{name: "d64_tie", input: "1.0000000000000005"},
		{name: "d128_tie", input: "1." + strings.Repeat("0", 33) + "5"},
		// infinities
		{name: "inf", input: "inf"},
		{name: "Infinity", input: "Infinity"},
		{name: "-inf", input: "-inf"},
		{name: "+INF", input: "+INF"},
		{name: "sp_inf", input: " inf"},
		{name: "inf_sp", input: "inf "},
		{name: "infin", input: "infin"},
		{name: "-infinity", input: "-infinity"},
		// NaN literals and payload width limits
		{name: "nan", input: "nan"},
		{name: "NaN", input: "NaN"},
		{name: "-NaN", input: "-NaN"},
		{name: "+qnan", input: "+qnan"},
		{name: "SNaN", input: "SNaN"},
		{name: "snan123", input: "snan123"},
		{name: "-nan999999", input: "-nan999999"},
		{name: "nan1000000", input: "nan1000000"},
		{name: "nan_d64_payload_max", input: "nan999999999999999"},
		{name: "nan_d64_payload_over", input: "nan1000000000000000"},
		{name: "nan_d128_payload_over", input: "nan" + strings.Repeat("9", 33)},
		{name: "nan_paren", input: "nan(123)"},
		{name: "nan_sp_1", input: "nan 1"},
		{name: "sp_nan", input: " nan"},
		{name: "nan_sp", input: "nan "},
		{name: "long_s_nan", input: "ſnan"},
		// syntax and whitespace error edges
		{name: "empty", input: ""},
		{name: "space", input: " "},
		{name: "tab", input: "\t"},
		{name: "dot", input: "."},
		{name: "dot_e1", input: ".e1"},
		{name: "1e_bare", input: "1e"},
		{name: "1e_plus", input: "1e+"},
		{name: "1e_minus", input: "1e-"},
		{name: "e5", input: "e5"},
		{name: "plus", input: "+"},
		{name: "minus", input: "-"},
		{name: "double_minus", input: "--1"},
		{name: "1..2", input: "1..2"},
		{name: "1.2.3", input: "1.2.3"},
		{name: "1_sp_e5", input: "1 e5"},
		{name: "1e5_sp", input: "1e5 "},
		{name: "1.0_sp", input: "1.0 "},
		{name: "sp_1.5", input: " 1.5"},
		{name: "tab_1.5", input: "\t1.5"},
		{name: "+.5", input: "+.5"},
		{name: "comma", input: "1,5"},
		{name: "hex", input: "0x10"},
		{name: "arabic_digits", input: "١٥"},
		{name: "1e+5", input: "1e+5"},
		{name: "1E5", input: "1E5"},
	}
	return rows
}

// parseBoundaryPin freezes the observed public parse results of one input:
// raw parse bits+flags per width plus the error-channel accept decisions of
// the Direct and WithFlags constructors. WithMode under RoundNearestEven is
// asserted structurally equal to WithFlags, so it needs no separate pin.
type parseBoundaryPin struct {
	raw32       uint32
	rawFlags32  ExceptionFlags
	direct32    bool
	withFlags32 bool

	raw64       uint64
	rawFlags64  ExceptionFlags
	direct64    bool
	withFlags64 bool

	raw128hi     uint64
	raw128lo     uint64
	rawFlags128  ExceptionFlags
	direct128    bool
	withFlags128 bool
}

func TestPublicParseBoundaryDecisionsPinned(t *testing.T) {
	rows := parseBoundaryRows()
	if len(rows) != len(parseBoundaryPins) {
		t.Fatalf("row/pin count mismatch: %d rows, %d pins", len(rows), len(parseBoundaryPins))
	}
	for _, row := range rows {
		row := row
		pin, ok := parseBoundaryPins[row.name]
		if !ok {
			t.Fatalf("missing pin for row %q", row.name)
		}
		t.Run(row.name, func(t *testing.T) {
			assertParseBoundary32(t, row.input, pin)
			assertParseBoundary64(t, row.input, pin)
			assertParseBoundary128(t, row.input, pin)
		})
	}
}

func assertParseBoundary32(t *testing.T, input string, pin parseBoundaryPin) {
	t.Helper()
	raw, rawFlags := ParseDecimal32BIDRaw(input)
	if uint32(raw) != pin.raw32 || rawFlags != pin.rawFlags32 {
		t.Errorf("ParseDecimal32BIDRaw(%q) = (%#08x, %v), pinned (%#08x, %v)",
			input, uint32(raw), rawFlags, pin.raw32, pin.rawFlags32)
	}
	v, err := NewDecimal32BIDDirect(input)
	if (err == nil) != pin.direct32 {
		t.Errorf("NewDecimal32BIDDirect(%q) accept = %v, pinned %v (err=%v)", input, err == nil, pin.direct32, err)
	}
	if err == nil && v != raw {
		t.Errorf("NewDecimal32BIDDirect(%q) = %#08x, want raw parse result %#08x", input, uint32(v), uint32(raw))
	}
	if err != nil && v != 0 {
		t.Errorf("NewDecimal32BIDDirect(%q) rejected with nonzero value %#08x", input, uint32(v))
	}
	w, wf, werr := NewDecimal32WithFlags(input)
	if (werr == nil) != pin.withFlags32 {
		t.Errorf("NewDecimal32WithFlags(%q) accept = %v, pinned %v (err=%v)", input, werr == nil, pin.withFlags32, werr)
	}
	if werr == nil && (w != raw || wf != rawFlags) {
		t.Errorf("NewDecimal32WithFlags(%q) = (%#08x, %v), want raw parse result (%#08x, %v)",
			input, uint32(w), wf, uint32(raw), rawFlags)
	}
	if werr != nil && (w != 0 || wf != 0) {
		t.Errorf("NewDecimal32WithFlags(%q) rejected with nonzero (%#08x, %v)", input, uint32(w), wf)
	}
	m, mf, merr := NewDecimal32WithMode(input, RoundNearestEven)
	if m != w || mf != wf || (merr == nil) != (werr == nil) {
		t.Errorf("NewDecimal32WithMode(%q, RoundNearestEven) = (%#08x, %v, %v), want WithFlags result (%#08x, %v, %v)",
			input, uint32(m), mf, merr, uint32(w), wf, werr)
	}
	if pin.direct32 && !pin.withFlags32 {
		t.Errorf("pin inconsistency for %q: direct accepts but WithFlags rejects", input)
	}
}

func assertParseBoundary64(t *testing.T, input string, pin parseBoundaryPin) {
	t.Helper()
	raw, rawFlags := ParseDecimal64BIDRaw(input)
	if uint64(raw) != pin.raw64 || rawFlags != pin.rawFlags64 {
		t.Errorf("ParseDecimal64BIDRaw(%q) = (%#016x, %v), pinned (%#016x, %v)",
			input, uint64(raw), rawFlags, pin.raw64, pin.rawFlags64)
	}
	v, err := NewDecimal64BIDDirect(input)
	if (err == nil) != pin.direct64 {
		t.Errorf("NewDecimal64BIDDirect(%q) accept = %v, pinned %v (err=%v)", input, err == nil, pin.direct64, err)
	}
	if err == nil && v != raw {
		t.Errorf("NewDecimal64BIDDirect(%q) = %#016x, want raw parse result %#016x", input, uint64(v), uint64(raw))
	}
	if err != nil && v != 0 {
		t.Errorf("NewDecimal64BIDDirect(%q) rejected with nonzero value %#016x", input, uint64(v))
	}
	w, wf, werr := NewDecimal64WithFlags(input)
	if (werr == nil) != pin.withFlags64 {
		t.Errorf("NewDecimal64WithFlags(%q) accept = %v, pinned %v (err=%v)", input, werr == nil, pin.withFlags64, werr)
	}
	if werr == nil && (w != raw || wf != rawFlags) {
		t.Errorf("NewDecimal64WithFlags(%q) = (%#016x, %v), want raw parse result (%#016x, %v)",
			input, uint64(w), wf, uint64(raw), rawFlags)
	}
	if werr != nil && (w != 0 || wf != 0) {
		t.Errorf("NewDecimal64WithFlags(%q) rejected with nonzero (%#016x, %v)", input, uint64(w), wf)
	}
	m, mf, merr := NewDecimal64WithMode(input, RoundNearestEven)
	if m != w || mf != wf || (merr == nil) != (werr == nil) {
		t.Errorf("NewDecimal64WithMode(%q, RoundNearestEven) = (%#016x, %v, %v), want WithFlags result (%#016x, %v, %v)",
			input, uint64(m), mf, merr, uint64(w), wf, werr)
	}
	if pin.direct64 && !pin.withFlags64 {
		t.Errorf("pin inconsistency for %q: direct accepts but WithFlags rejects", input)
	}
}

func assertParseBoundary128(t *testing.T, input string, pin parseBoundaryPin) {
	t.Helper()
	raw, rawFlags := ParseDecimal128BIDRaw(input)
	hi, lo := decimal128BIDWords(raw)
	if hi != pin.raw128hi || lo != pin.raw128lo || rawFlags != pin.rawFlags128 {
		t.Errorf("ParseDecimal128BIDRaw(%q) = (%#016x_%016x, %v), pinned (%#016x_%016x, %v)",
			input, hi, lo, rawFlags, pin.raw128hi, pin.raw128lo, pin.rawFlags128)
	}
	v, err := NewDecimal128BIDDirect(input)
	if (err == nil) != pin.direct128 {
		t.Errorf("NewDecimal128BIDDirect(%q) accept = %v, pinned %v (err=%v)", input, err == nil, pin.direct128, err)
	}
	if err == nil && v != raw {
		t.Errorf("NewDecimal128BIDDirect(%q) = %x, want raw parse result %x", input, v.ToBytes(), raw.ToBytes())
	}
	if err != nil && v != (Decimal128BID{}) {
		t.Errorf("NewDecimal128BIDDirect(%q) rejected with nonzero value %x", input, v.ToBytes())
	}
	w, wf, werr := NewDecimal128WithFlags(input)
	if (werr == nil) != pin.withFlags128 {
		t.Errorf("NewDecimal128WithFlags(%q) accept = %v, pinned %v (err=%v)", input, werr == nil, pin.withFlags128, werr)
	}
	if werr == nil && (w != raw || wf != rawFlags) {
		t.Errorf("NewDecimal128WithFlags(%q) = (%x, %v), want raw parse result (%x, %v)",
			input, w.ToBytes(), wf, raw.ToBytes(), rawFlags)
	}
	if werr != nil && (w != (Decimal128BID{}) || wf != 0) {
		t.Errorf("NewDecimal128WithFlags(%q) rejected with nonzero (%x, %v)", input, w.ToBytes(), wf)
	}
	m, mf, merr := NewDecimal128WithMode(input, RoundNearestEven)
	if m != w || mf != wf || (merr == nil) != (werr == nil) {
		t.Errorf("NewDecimal128WithMode(%q, RoundNearestEven) = (%x, %v, %v), want WithFlags result (%x, %v, %v)",
			input, m.ToBytes(), mf, merr, w.ToBytes(), wf, werr)
	}
	if pin.direct128 && !pin.withFlags128 {
		t.Errorf("pin inconsistency for %q: direct accepts but WithFlags rejects", input)
	}
}

// Rounding-mode sensitivity: ties and overflow/underflow rows where the carried
// mode changes the parsed value, pinned per explicit mode so a mode-threading
// regression in the WithMode plumbing cannot hide behind the default mode.
type parseBoundaryModeRow struct {
	name  string
	input string
	mode  RoundingMode
}

type parseBoundaryModeKey struct {
	name string
	mode RoundingMode
}

type parseBoundaryModePin struct {
	v32    uint32
	f32    ExceptionFlags
	ok32   bool
	v64    uint64
	f64    ExceptionFlags
	ok64   bool
	v128hi uint64
	v128lo uint64
	f128   ExceptionFlags
	ok128  bool
}

func parseBoundaryModeRows() []parseBoundaryModeRow {
	d128Tie := "1." + strings.Repeat("0", 33) + "5"
	return []parseBoundaryModeRow{
		{name: "1.0000005", input: "1.0000005", mode: RoundTowardZero},
		{name: "1.0000005", input: "1.0000005", mode: RoundNearestAway},
		{name: "-1.0000005", input: "-1.0000005", mode: RoundTowardNegative},
		{name: "-1.0000005", input: "-1.0000005", mode: RoundNearestAway},
		{name: "9999999.5", input: "9999999.5", mode: RoundTowardZero},
		{name: "9999999.5", input: "9999999.5", mode: RoundNearestAway},
		{name: "d64_tie", input: "1.0000000000000005", mode: RoundTowardZero},
		{name: "d64_tie", input: "1.0000000000000005", mode: RoundNearestAway},
		{name: "d128_tie", input: d128Tie, mode: RoundTowardZero},
		{name: "d128_tie", input: d128Tie, mode: RoundNearestAway},
		{name: "1e97", input: "1e97", mode: RoundTowardZero},
		{name: "1e97", input: "1e97", mode: RoundTowardNegative},
		{name: "1e-102", input: "1e-102", mode: RoundTowardPositive},
		{name: "1e-102", input: "1e-102", mode: RoundTowardZero},
		{name: "1e385", input: "1e385", mode: RoundTowardZero},
		{name: "1e6145", input: "1e6145", mode: RoundTowardZero},
	}
}

func TestPublicParseBoundaryModeDecisionsPinned(t *testing.T) {
	rows := parseBoundaryModeRows()
	if len(rows) != len(parseBoundaryModePins) {
		t.Fatalf("row/pin count mismatch: %d rows, %d pins", len(rows), len(parseBoundaryModePins))
	}
	for _, row := range rows {
		row := row
		pin, ok := parseBoundaryModePins[parseBoundaryModeKey{row.name, row.mode}]
		if !ok {
			t.Fatalf("missing mode pin for row %q mode %v", row.name, row.mode)
		}
		t.Run(row.name+"/"+row.mode.String(), func(t *testing.T) {
			v32, f32, e32 := NewDecimal32WithMode(row.input, row.mode)
			if uint32(v32) != pin.v32 || f32 != pin.f32 || (e32 == nil) != pin.ok32 {
				t.Errorf("NewDecimal32WithMode(%q, %v) = (%#08x, %v, %v), pinned (%#08x, %v, ok=%v)",
					row.input, row.mode, uint32(v32), f32, e32, pin.v32, pin.f32, pin.ok32)
			}
			v64, f64, e64 := NewDecimal64WithMode(row.input, row.mode)
			if uint64(v64) != pin.v64 || f64 != pin.f64 || (e64 == nil) != pin.ok64 {
				t.Errorf("NewDecimal64WithMode(%q, %v) = (%#016x, %v, %v), pinned (%#016x, %v, ok=%v)",
					row.input, row.mode, uint64(v64), f64, e64, pin.v64, pin.f64, pin.ok64)
			}
			v128, f128, e128 := NewDecimal128WithMode(row.input, row.mode)
			hi, lo := decimal128BIDWords(v128)
			if hi != pin.v128hi || lo != pin.v128lo || f128 != pin.f128 || (e128 == nil) != pin.ok128 {
				t.Errorf("NewDecimal128WithMode(%q, %v) = (%#016x_%016x, %v, %v), pinned (%#016x_%016x, %v, ok=%v)",
					row.input, row.mode, hi, lo, f128, e128, pin.v128hi, pin.v128lo, pin.f128, pin.ok128)
			}
		})
	}
}

// Invalid rounding-mode boundary: an accepted string with an undefined mode
// reports canonical quiet NaN + FlagInvalidOperation with a nil error, while a
// rejected string keeps the ordinary (zero value, zero flags, error) result.
func TestPublicParseBoundaryInvalidRoundingMode(t *testing.T) {
	const badMode = RoundingMode(99)
	if v, f, err := NewDecimal32WithMode("1.5", badMode); v != canonicalQNaN32BID() || f != FlagInvalidOperation || err != nil {
		t.Errorf("NewDecimal32WithMode(1.5, invalid) = (%#08x, %v, %v), want canonical qNaN + FlagInvalidOperation + nil", uint32(v), f, err)
	}
	if v, f, err := NewDecimal32WithMode("junk", badMode); v != 0 || f != 0 || err == nil {
		t.Errorf("NewDecimal32WithMode(junk, invalid) = (%#08x, %v, %v), want zero + zero flags + error", uint32(v), f, err)
	}
	if v, f, err := NewDecimal64WithMode("1.5", badMode); v != canonicalQNaN64BID() || f != FlagInvalidOperation || err != nil {
		t.Errorf("NewDecimal64WithMode(1.5, invalid) = (%#016x, %v, %v), want canonical qNaN + FlagInvalidOperation + nil", uint64(v), f, err)
	}
	if v, f, err := NewDecimal64WithMode("junk", badMode); v != 0 || f != 0 || err == nil {
		t.Errorf("NewDecimal64WithMode(junk, invalid) = (%#016x, %v, %v), want zero + zero flags + error", uint64(v), f, err)
	}
	if v, f, err := NewDecimal128WithMode("1.5", badMode); v != canonicalQNaN128BID() || f != FlagInvalidOperation || err != nil {
		t.Errorf("NewDecimal128WithMode(1.5, invalid) = (%x, %v, %v), want canonical qNaN + FlagInvalidOperation + nil", v.ToBytes(), f, err)
	}
	if v, f, err := NewDecimal128WithMode("junk", badMode); v != (Decimal128BID{}) || f != 0 || err == nil {
		t.Errorf("NewDecimal128WithMode(junk, invalid) = (%x, %v, %v), want zero + zero flags + error", v.ToBytes(), f, err)
	}
}

var parseBoundaryPins = map[string]parseBoundaryPin{
	"0": {
		raw32: 0x32800000, rawFlags32: 0x0, direct32: true, withFlags32: true,
		raw64: 0x31c0000000000000, rawFlags64: 0x0, direct64: true, withFlags64: true,
		raw128hi: 0x3040000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"-0": {
		raw32: 0xb2800000, rawFlags32: 0x0, direct32: true, withFlags32: true,
		raw64: 0xb1c0000000000000, rawFlags64: 0x0, direct64: true, withFlags64: true,
		raw128hi: 0xb040000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"+0.00": {
		raw32: 0x31800000, rawFlags32: 0x0, direct32: true, withFlags32: true,
		raw64: 0x3180000000000000, rawFlags64: 0x0, direct64: true, withFlags64: true,
		raw128hi: 0x303c000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"0e0": {
		raw32: 0x32800000, rawFlags32: 0x0, direct32: true, withFlags32: true,
		raw64: 0x31c0000000000000, rawFlags64: 0x0, direct64: true, withFlags64: true,
		raw128hi: 0x3040000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"0e90": {
		raw32: 0x5f800000, rawFlags32: 0x0, direct32: true, withFlags32: true,
		raw64: 0x3d00000000000000, rawFlags64: 0x0, direct64: true, withFlags64: true,
		raw128hi: 0x30f4000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"0e91": {
		raw32: 0x7c000000, rawFlags32: 0x10, direct32: false, withFlags32: false,
		raw64: 0x3d20000000000000, rawFlags64: 0x0, direct64: true, withFlags64: true,
		raw128hi: 0x30f6000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"0e-101": {
		raw32: 0x00000000, rawFlags32: 0x0, direct32: true, withFlags32: true,
		raw64: 0x2520000000000000, rawFlags64: 0x0, direct64: true, withFlags64: true,
		raw128hi: 0x2f76000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"0e-102": {
		raw32: 0x7c000000, rawFlags32: 0x10, direct32: false, withFlags32: false,
		raw64: 0x2500000000000000, rawFlags64: 0x0, direct64: true, withFlags64: true,
		raw128hi: 0x2f74000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"0e369": {
		raw32: 0x7c000000, rawFlags32: 0x10, direct32: false, withFlags32: false,
		raw64: 0x5fe0000000000000, rawFlags64: 0x0, direct64: true, withFlags64: true,
		raw128hi: 0x3322000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"0e370": {
		raw32: 0x7c000000, rawFlags32: 0x10, direct32: false, withFlags32: false,
		raw64: 0x7c00000000000000, rawFlags64: 0x10, direct64: false, withFlags64: false,
		raw128hi: 0x3324000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"0e-398": {
		raw32: 0x00000000, rawFlags32: 0x3, direct32: false, withFlags32: true,
		raw64: 0x0000000000000000, rawFlags64: 0x0, direct64: true, withFlags64: true,
		raw128hi: 0x2d24000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"0e-399": {
		raw32: 0x00000000, rawFlags32: 0x3, direct32: false, withFlags32: true,
		raw64: 0x7c00000000000000, rawFlags64: 0x10, direct64: false, withFlags64: false,
		raw128hi: 0x2d22000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"0e6111": {
		raw32: 0x7c000000, rawFlags32: 0x10, direct32: false, withFlags32: false,
		raw64: 0x7c00000000000000, rawFlags64: 0x10, direct64: false, withFlags64: false,
		raw128hi: 0x5ffe000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"0e6112": {
		raw32: 0x7c000000, rawFlags32: 0x10, direct32: false, withFlags32: false,
		raw64: 0x7c00000000000000, rawFlags64: 0x10, direct64: false, withFlags64: false,
		raw128hi: 0x7c00000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x10, direct128: false, withFlags128: false,
	},
	"-0e-6176": {
		raw32: 0x80000000, rawFlags32: 0x3, direct32: false, withFlags32: true,
		raw64: 0x8000000000000000, rawFlags64: 0x3, direct64: false, withFlags64: true,
		raw128hi: 0x8000000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"0e-6177": {
		raw32: 0x00000000, rawFlags32: 0x3, direct32: false, withFlags32: true,
		raw64: 0x0000000000000000, rawFlags64: 0x3, direct64: false, withFlags64: true,
		raw128hi: 0x7c00000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x10, direct128: false, withFlags128: false,
	},
	"9999999": {
		raw32: 0x6cb8967f, rawFlags32: 0x0, direct32: true, withFlags32: true,
		raw64: 0x31c000000098967f, rawFlags64: 0x0, direct64: true, withFlags64: true,
		raw128hi: 0x3040000000000000, raw128lo: 0x000000000098967f, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"99999999": {
		raw32: 0x338f4240, rawFlags32: 0x1, direct32: false, withFlags32: true,
		raw64: 0x31c0000005f5e0ff, rawFlags64: 0x0, direct64: true, withFlags64: true,
		raw128hi: 0x3040000000000000, raw128lo: 0x0000000005f5e0ff, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"0009999999": {
		raw32: 0x6cb8967f, rawFlags32: 0x0, direct32: true, withFlags32: true,
		raw64: 0x31c000000098967f, rawFlags64: 0x0, direct64: true, withFlags64: true,
		raw128hi: 0x3040000000000000, raw128lo: 0x000000000098967f, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"10000000": {
		raw32: 0x7c000000, rawFlags32: 0x10, direct32: false, withFlags32: false,
		raw64: 0x31c0000000989680, rawFlags64: 0x0, direct64: true, withFlags64: true,
		raw128hi: 0x3040000000000000, raw128lo: 0x0000000000989680, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"1000000.0": {
		raw32: 0x7c000000, rawFlags32: 0x10, direct32: false, withFlags32: false,
		raw64: 0x31a0000000989680, rawFlags64: 0x0, direct64: true, withFlags64: true,
		raw128hi: 0x303e000000000000, raw128lo: 0x0000000000989680, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"9999999999999999": {
		raw32: 0x378f4240, rawFlags32: 0x1, direct32: false, withFlags32: true,
		raw64: 0x6c7386f26fc0ffff, rawFlags64: 0x0, direct64: true, withFlags64: true,
		raw128hi: 0x3040000000000000, raw128lo: 0x002386f26fc0ffff, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"99999999999999999": {
		raw32: 0x380f4240, rawFlags32: 0x1, direct32: false, withFlags32: true,
		raw64: 0x32038d7ea4c68000, rawFlags64: 0x1, direct64: false, withFlags64: true,
		raw128hi: 0x3040000000000000, raw128lo: 0x016345785d89ffff, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"10000000000000000": {
		raw32: 0x7c000000, rawFlags32: 0x10, direct32: false, withFlags32: false,
		raw64: 0x7c00000000000000, rawFlags64: 0x10, direct64: false, withFlags64: false,
		raw128hi: 0x3040000000000000, raw128lo: 0x002386f26fc10000, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"34x9": {
		raw32: 0x408f4240, rawFlags32: 0x1, direct32: false, withFlags32: true,
		raw64: 0x34238d7ea4c68000, rawFlags64: 0x1, direct64: false, withFlags64: true,
		raw128hi: 0x3041ed09bead87c0, raw128lo: 0x378d8e63ffffffff, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"35x9": {
		raw32: 0x410f4240, rawFlags32: 0x1, direct32: false, withFlags32: true,
		raw64: 0x34438d7ea4c68000, rawFlags64: 0x1, direct64: false, withFlags64: true,
		raw128hi: 0x3044314dc6448d93, raw128lo: 0x38c15b0a00000000, rawFlags128: 0x1, direct128: false, withFlags128: true,
	},
	"1_33x0": {
		raw32: 0x7c000000, rawFlags32: 0x10, direct32: false, withFlags32: false,
		raw64: 0x7c00000000000000, rawFlags64: 0x10, direct64: false, withFlags64: false,
		raw128hi: 0x3040314dc6448d93, raw128lo: 0x38c15b0a00000000, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"1e90": {
		raw32: 0x5f800001, rawFlags32: 0x0, direct32: true, withFlags32: true,
		raw64: 0x3d00000000000001, rawFlags64: 0x0, direct64: true, withFlags64: true,
		raw128hi: 0x30f4000000000000, raw128lo: 0x0000000000000001, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"1e91": {
		raw32: 0x7c000000, rawFlags32: 0x10, direct32: false, withFlags32: false,
		raw64: 0x3d20000000000001, rawFlags64: 0x0, direct64: true, withFlags64: true,
		raw128hi: 0x30f6000000000000, raw128lo: 0x0000000000000001, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"9.999999e96": {
		raw32: 0x77f8967f, rawFlags32: 0x0, direct32: true, withFlags32: true,
		raw64: 0x3d0000000098967f, rawFlags64: 0x0, direct64: true, withFlags64: true,
		raw128hi: 0x30f4000000000000, raw128lo: 0x000000000098967f, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"1e97": {
		raw32: 0x78000000, rawFlags32: 0x5, direct32: false, withFlags32: true,
		raw64: 0x3de0000000000001, rawFlags64: 0x0, direct64: true, withFlags64: true,
		raw128hi: 0x3102000000000000, raw128lo: 0x0000000000000001, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"1e-101": {
		raw32: 0x00000001, rawFlags32: 0x0, direct32: true, withFlags32: true,
		raw64: 0x2520000000000001, rawFlags64: 0x0, direct64: true, withFlags64: true,
		raw128hi: 0x2f76000000000000, raw128lo: 0x0000000000000001, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"1e-102": {
		raw32: 0x00000000, rawFlags32: 0x3, direct32: false, withFlags32: true,
		raw64: 0x2500000000000001, rawFlags64: 0x0, direct64: true, withFlags64: true,
		raw128hi: 0x2f74000000000000, raw128lo: 0x0000000000000001, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"0.000001e-95": {
		raw32: 0x00000001, rawFlags32: 0x0, direct32: true, withFlags32: true,
		raw64: 0x2520000000000001, rawFlags64: 0x0, direct64: true, withFlags64: true,
		raw128hi: 0x2f76000000000000, raw128lo: 0x0000000000000001, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"9.999999999999999e384": {
		raw32: 0x78000000, rawFlags32: 0x5, direct32: false, withFlags32: true,
		raw64: 0x77fb86f26fc0ffff, rawFlags64: 0x0, direct64: true, withFlags64: true,
		raw128hi: 0x3322000000000000, raw128lo: 0x002386f26fc0ffff, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"1e385": {
		raw32: 0x78000000, rawFlags32: 0x5, direct32: false, withFlags32: true,
		raw64: 0x7800000000000000, rawFlags64: 0x5, direct64: false, withFlags64: true,
		raw128hi: 0x3342000000000000, raw128lo: 0x0000000000000001, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"1e-398": {
		raw32: 0x00000000, rawFlags32: 0x3, direct32: false, withFlags32: true,
		raw64: 0x0000000000000001, rawFlags64: 0x0, direct64: true, withFlags64: true,
		raw128hi: 0x2d24000000000000, raw128lo: 0x0000000000000001, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"1e-399": {
		raw32: 0x00000000, rawFlags32: 0x3, direct32: false, withFlags32: true,
		raw64: 0x0000000000000000, rawFlags64: 0x3, direct64: false, withFlags64: true,
		raw128hi: 0x2d22000000000000, raw128lo: 0x0000000000000001, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"d128_max_normal": {
		raw32: 0x78000000, rawFlags32: 0x5, direct32: false, withFlags32: true,
		raw64: 0x7800000000000000, rawFlags64: 0x5, direct64: false, withFlags64: true,
		raw128hi: 0x5fffed09bead87c0, raw128lo: 0x378d8e63ffffffff, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"1e6145": {
		raw32: 0x78000000, rawFlags32: 0x5, direct32: false, withFlags32: true,
		raw64: 0x7800000000000000, rawFlags64: 0x5, direct64: false, withFlags64: true,
		raw128hi: 0x7800000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x5, direct128: false, withFlags128: true,
	},
	"1e-6176": {
		raw32: 0x00000000, rawFlags32: 0x3, direct32: false, withFlags32: true,
		raw64: 0x0000000000000000, rawFlags64: 0x3, direct64: false, withFlags64: true,
		raw128hi: 0x0000000000000000, raw128lo: 0x0000000000000001, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"1e-6177": {
		raw32: 0x00000000, rawFlags32: 0x3, direct32: false, withFlags32: true,
		raw64: 0x0000000000000000, rawFlags64: 0x3, direct64: false, withFlags64: true,
		raw128hi: 0x0000000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x3, direct128: false, withFlags128: true,
	},
	"1e2147483647": {
		raw32: 0x78000000, rawFlags32: 0x5, direct32: false, withFlags32: true,
		raw64: 0x7800000000000000, rawFlags64: 0x5, direct64: false, withFlags64: true,
		raw128hi: 0x7800000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x5, direct128: false, withFlags128: true,
	},
	"1e2147483648": {
		raw32: 0x78000000, rawFlags32: 0x5, direct32: false, withFlags32: true,
		raw64: 0x7800000000000000, rawFlags64: 0x5, direct64: false, withFlags64: true,
		raw128hi: 0x7800000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x5, direct128: false, withFlags128: true,
	},
	"1e-2147483648": {
		raw32: 0x00000000, rawFlags32: 0x3, direct32: false, withFlags32: true,
		raw64: 0x0000000000000000, rawFlags64: 0x3, direct64: false, withFlags64: true,
		raw128hi: 0x0000000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x3, direct128: false, withFlags128: true,
	},
	"1e-2147483649": {
		raw32: 0x00000000, rawFlags32: 0x3, direct32: false, withFlags32: true,
		raw64: 0x0000000000000000, rawFlags64: 0x3, direct64: false, withFlags64: true,
		raw128hi: 0x0000000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x3, direct128: false, withFlags128: true,
	},
	"1e_int64max": {
		raw32: 0x78000000, rawFlags32: 0x5, direct32: false, withFlags32: true,
		raw64: 0x7800000000000000, rawFlags64: 0x5, direct64: false, withFlags64: true,
		raw128hi: 0x7800000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x5, direct128: false, withFlags128: true,
	},
	"1e_int64max+1": {
		raw32: 0x78000000, rawFlags32: 0x5, direct32: false, withFlags32: true,
		raw64: 0x7800000000000000, rawFlags64: 0x5, direct64: false, withFlags64: true,
		raw128hi: 0x7800000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x5, direct128: false, withFlags128: true,
	},
	"1e_int64min": {
		raw32: 0x00000000, rawFlags32: 0x3, direct32: false, withFlags32: true,
		raw64: 0x0000000000000000, rawFlags64: 0x3, direct64: false, withFlags64: true,
		raw128hi: 0x0000000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x3, direct128: false, withFlags128: true,
	},
	"1e_int64min-1": {
		raw32: 0x00000000, rawFlags32: 0x3, direct32: false, withFlags32: true,
		raw64: 0x0000000000000000, rawFlags64: 0x3, direct64: false, withFlags64: true,
		raw128hi: 0x0000000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x3, direct128: false, withFlags128: true,
	},
	"1e_uint64max": {
		raw32: 0x78000000, rawFlags32: 0x5, direct32: false, withFlags32: true,
		raw64: 0x7800000000000000, rawFlags64: 0x5, direct64: false, withFlags64: true,
		raw128hi: 0x7800000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x5, direct128: false, withFlags128: true,
	},
	"1e_uint64max+1": {
		raw32: 0x78000000, rawFlags32: 0x5, direct32: false, withFlags32: true,
		raw64: 0x7800000000000000, rawFlags64: 0x5, direct64: false, withFlags64: true,
		raw128hi: 0x7800000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x5, direct128: false, withFlags128: true,
	},
	"1e_-uint64max": {
		raw32: 0x00000000, rawFlags32: 0x3, direct32: false, withFlags32: true,
		raw64: 0x0000000000000000, rawFlags64: 0x3, direct64: false, withFlags64: true,
		raw128hi: 0x0000000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x3, direct128: false, withFlags128: true,
	},
	"1e_-uint64max-1": {
		raw32: 0x00000000, rawFlags32: 0x3, direct32: false, withFlags32: true,
		raw64: 0x0000000000000000, rawFlags64: 0x3, direct64: false, withFlags64: true,
		raw128hi: 0x0000000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x3, direct128: false, withFlags128: true,
	},
	"1e_30nines": {
		raw32: 0x78000000, rawFlags32: 0x5, direct32: false, withFlags32: true,
		raw64: 0x7800000000000000, rawFlags64: 0x5, direct64: false, withFlags64: true,
		raw128hi: 0x7800000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x5, direct128: false, withFlags128: true,
	},
	"1e_-30nines": {
		raw32: 0x00000000, rawFlags32: 0x3, direct32: false, withFlags32: true,
		raw64: 0x0000000000000000, rawFlags64: 0x3, direct64: false, withFlags64: true,
		raw128hi: 0x0000000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x3, direct128: false, withFlags128: true,
	},
	"1e_zeropad90": {
		raw32: 0x5f800001, rawFlags32: 0x0, direct32: true, withFlags32: true,
		raw64: 0x3d00000000000001, rawFlags64: 0x0, direct64: true, withFlags64: true,
		raw128hi: 0x30f4000000000000, raw128lo: 0x0000000000000001, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"0e_int64max": {
		raw32: 0x7c000000, rawFlags32: 0x10, direct32: false, withFlags32: false,
		raw64: 0x7c00000000000000, rawFlags64: 0x10, direct64: false, withFlags64: false,
		raw128hi: 0x7c00000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x10, direct128: false, withFlags128: false,
	},
	"0e_int64max+1": {
		raw32: 0x7c000000, rawFlags32: 0x10, direct32: false, withFlags32: false,
		raw64: 0x7c00000000000000, rawFlags64: 0x10, direct64: false, withFlags64: false,
		raw128hi: 0x7c00000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x10, direct128: false, withFlags128: false,
	},
	"0e_int64min": {
		raw32: 0x00000000, rawFlags32: 0x3, direct32: false, withFlags32: true,
		raw64: 0x0000000000000000, rawFlags64: 0x3, direct64: false, withFlags64: true,
		raw128hi: 0x0000000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x3, direct128: false, withFlags128: true,
	},
	"0e_int64min-1": {
		raw32: 0x00000000, rawFlags32: 0x3, direct32: false, withFlags32: true,
		raw64: 0x0000000000000000, rawFlags64: 0x3, direct64: false, withFlags64: true,
		raw128hi: 0x0000000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x3, direct128: false, withFlags128: true,
	},
	"0e_uint64max+1": {
		raw32: 0x7c000000, rawFlags32: 0x10, direct32: false, withFlags32: false,
		raw64: 0x7c00000000000000, rawFlags64: 0x10, direct64: false, withFlags64: false,
		raw128hi: 0x7c00000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x10, direct128: false, withFlags128: false,
	},
	"frac_cancel_in_range": {
		raw32: 0x19800001, rawFlags32: 0x0, direct32: true, withFlags32: true,
		raw64: 0x2b80000000000001, rawFlags64: 0x0, direct64: true, withFlags64: true,
		raw128hi: 0x2fdc000000000000, raw128lo: 0x0000000000000001, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"frac_cancel_at_int64max": {
		raw32: 0x78000000, rawFlags32: 0x5, direct32: false, withFlags32: true,
		raw64: 0x7800000000000000, rawFlags64: 0x5, direct64: false, withFlags64: true,
		raw128hi: 0x7800000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x5, direct128: false, withFlags128: true,
	},
	"frac_cancel_past_int64": {
		raw32: 0x78000000, rawFlags32: 0x5, direct32: false, withFlags32: true,
		raw64: 0x7800000000000000, rawFlags64: 0x5, direct64: false, withFlags64: true,
		raw128hi: 0x7800000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x5, direct128: false, withFlags128: true,
	},
	"frac_cancel_past_uint64": {
		raw32: 0x78000000, rawFlags32: 0x5, direct32: false, withFlags32: true,
		raw64: 0x7800000000000000, rawFlags64: 0x5, direct64: false, withFlags64: true,
		raw128hi: 0x7800000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x5, direct128: false, withFlags128: true,
	},
	"frac_cancel_zero_coeff": {
		raw32: 0x7c000000, rawFlags32: 0x10, direct32: false, withFlags32: false,
		raw64: 0x7c00000000000000, rawFlags64: 0x10, direct64: false, withFlags64: false,
		raw128hi: 0x7c00000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x10, direct128: false, withFlags128: false,
	},
	"1.0000005": {
		raw32: 0x2f8f4240, rawFlags32: 0x1, direct32: false, withFlags32: true,
		raw64: 0x30e0000000989685, rawFlags64: 0x0, direct64: true, withFlags64: true,
		raw128hi: 0x3032000000000000, raw128lo: 0x0000000000989685, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"-1.0000005": {
		raw32: 0xaf8f4240, rawFlags32: 0x1, direct32: false, withFlags32: true,
		raw64: 0xb0e0000000989685, rawFlags64: 0x0, direct64: true, withFlags64: true,
		raw128hi: 0xb032000000000000, raw128lo: 0x0000000000989685, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"9999999.5": {
		raw32: 0x330f4240, rawFlags32: 0x1, direct32: false, withFlags32: true,
		raw64: 0x31a0000005f5e0fb, rawFlags64: 0x0, direct64: true, withFlags64: true,
		raw128hi: 0x303e000000000000, raw128lo: 0x0000000005f5e0fb, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"d64_tie": {
		raw32: 0x2f8f4240, rawFlags32: 0x1, direct32: false, withFlags32: true,
		raw64: 0x2fe38d7ea4c68000, rawFlags64: 0x1, direct64: false, withFlags64: true,
		raw128hi: 0x3020000000000000, raw128lo: 0x002386f26fc10005, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"d128_tie": {
		raw32: 0x2f8f4240, rawFlags32: 0x1, direct32: false, withFlags32: true,
		raw64: 0x2fe38d7ea4c68000, rawFlags64: 0x1, direct64: false, withFlags64: true,
		raw128hi: 0x2ffe314dc6448d93, raw128lo: 0x38c15b0a00000000, rawFlags128: 0x1, direct128: false, withFlags128: true,
	},
	"inf": {
		raw32: 0x78000000, rawFlags32: 0x0, direct32: true, withFlags32: true,
		raw64: 0x7800000000000000, rawFlags64: 0x0, direct64: true, withFlags64: true,
		raw128hi: 0x7800000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"Infinity": {
		raw32: 0x78000000, rawFlags32: 0x0, direct32: true, withFlags32: true,
		raw64: 0x7800000000000000, rawFlags64: 0x0, direct64: true, withFlags64: true,
		raw128hi: 0x7800000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"-inf": {
		raw32: 0xf8000000, rawFlags32: 0x0, direct32: true, withFlags32: true,
		raw64: 0xf800000000000000, rawFlags64: 0x0, direct64: true, withFlags64: true,
		raw128hi: 0xf800000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"+INF": {
		raw32: 0x78000000, rawFlags32: 0x0, direct32: true, withFlags32: true,
		raw64: 0x7800000000000000, rawFlags64: 0x0, direct64: true, withFlags64: true,
		raw128hi: 0x7800000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"sp_inf": {
		raw32: 0x78000000, rawFlags32: 0x0, direct32: true, withFlags32: true,
		raw64: 0x7800000000000000, rawFlags64: 0x0, direct64: true, withFlags64: true,
		raw128hi: 0x7800000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"inf_sp": {
		raw32: 0x7c000000, rawFlags32: 0x10, direct32: false, withFlags32: false,
		raw64: 0x7c00000000000000, rawFlags64: 0x10, direct64: false, withFlags64: false,
		raw128hi: 0x7c00000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x10, direct128: false, withFlags128: false,
	},
	"infin": {
		raw32: 0x7c000000, rawFlags32: 0x10, direct32: false, withFlags32: false,
		raw64: 0x7c00000000000000, rawFlags64: 0x10, direct64: false, withFlags64: false,
		raw128hi: 0x7c00000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x10, direct128: false, withFlags128: false,
	},
	"-infinity": {
		raw32: 0xf8000000, rawFlags32: 0x0, direct32: true, withFlags32: true,
		raw64: 0xf800000000000000, rawFlags64: 0x0, direct64: true, withFlags64: true,
		raw128hi: 0xf800000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"nan": {
		raw32: 0x7c000000, rawFlags32: 0x0, direct32: true, withFlags32: true,
		raw64: 0x7c00000000000000, rawFlags64: 0x0, direct64: true, withFlags64: true,
		raw128hi: 0x7c00000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"NaN": {
		raw32: 0x7c000000, rawFlags32: 0x0, direct32: true, withFlags32: true,
		raw64: 0x7c00000000000000, rawFlags64: 0x0, direct64: true, withFlags64: true,
		raw128hi: 0x7c00000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"-NaN": {
		raw32: 0xfc000000, rawFlags32: 0x0, direct32: true, withFlags32: true,
		raw64: 0xfc00000000000000, rawFlags64: 0x0, direct64: true, withFlags64: true,
		raw128hi: 0xfc00000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"+qnan": {
		raw32: 0x7c000000, rawFlags32: 0x0, direct32: true, withFlags32: true,
		raw64: 0x7c00000000000000, rawFlags64: 0x0, direct64: true, withFlags64: true,
		raw128hi: 0x7c00000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"SNaN": {
		raw32: 0x7e000000, rawFlags32: 0x0, direct32: true, withFlags32: true,
		raw64: 0x7e00000000000000, rawFlags64: 0x0, direct64: true, withFlags64: true,
		raw128hi: 0x7e00000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"snan123": {
		raw32: 0x7e00007b, rawFlags32: 0x0, direct32: true, withFlags32: true,
		raw64: 0x7e0000000000007b, rawFlags64: 0x0, direct64: true, withFlags64: true,
		raw128hi: 0x7e00000000000000, raw128lo: 0x000000000000007b, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"-nan999999": {
		raw32: 0xfc0f423f, rawFlags32: 0x0, direct32: true, withFlags32: true,
		raw64: 0xfc000000000f423f, rawFlags64: 0x0, direct64: true, withFlags64: true,
		raw128hi: 0xfc00000000000000, raw128lo: 0x00000000000f423f, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"nan1000000": {
		raw32: 0x7c000000, rawFlags32: 0x10, direct32: false, withFlags32: false,
		raw64: 0x7c000000000f4240, rawFlags64: 0x0, direct64: true, withFlags64: true,
		raw128hi: 0x7c00000000000000, raw128lo: 0x00000000000f4240, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"nan_d64_payload_max": {
		raw32: 0x7c000000, rawFlags32: 0x10, direct32: false, withFlags32: false,
		raw64: 0x7c038d7ea4c67fff, rawFlags64: 0x0, direct64: true, withFlags64: true,
		raw128hi: 0x7c00000000000000, raw128lo: 0x00038d7ea4c67fff, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"nan_d64_payload_over": {
		raw32: 0x7c000000, rawFlags32: 0x10, direct32: false, withFlags32: false,
		raw64: 0x7c00000000000000, rawFlags64: 0x10, direct64: false, withFlags64: false,
		raw128hi: 0x7c00000000000000, raw128lo: 0x00038d7ea4c68000, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"nan_d128_payload_over": {
		raw32: 0x7c000000, rawFlags32: 0x10, direct32: false, withFlags32: false,
		raw64: 0x7c00000000000000, rawFlags64: 0x10, direct64: false, withFlags64: false,
		raw128hi: 0x7c00314dc6448d93, raw128lo: 0x38c15b09ffffffff, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"nan_paren": {
		raw32: 0x7c000000, rawFlags32: 0x10, direct32: false, withFlags32: false,
		raw64: 0x7c00000000000000, rawFlags64: 0x10, direct64: false, withFlags64: false,
		raw128hi: 0x7c00000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x10, direct128: false, withFlags128: false,
	},
	"nan_sp_1": {
		raw32: 0x7c000000, rawFlags32: 0x10, direct32: false, withFlags32: false,
		raw64: 0x7c00000000000000, rawFlags64: 0x10, direct64: false, withFlags64: false,
		raw128hi: 0x7c00000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x10, direct128: false, withFlags128: false,
	},
	"sp_nan": {
		raw32: 0x7c000000, rawFlags32: 0x0, direct32: true, withFlags32: true,
		raw64: 0x7c00000000000000, rawFlags64: 0x0, direct64: true, withFlags64: true,
		raw128hi: 0x7c00000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"nan_sp": {
		raw32: 0x7c000000, rawFlags32: 0x10, direct32: false, withFlags32: false,
		raw64: 0x7c00000000000000, rawFlags64: 0x10, direct64: false, withFlags64: false,
		raw128hi: 0x7c00000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x10, direct128: false, withFlags128: false,
	},
	"long_s_nan": {
		raw32: 0x7c000000, rawFlags32: 0x10, direct32: false, withFlags32: false,
		raw64: 0x7c00000000000000, rawFlags64: 0x10, direct64: false, withFlags64: false,
		raw128hi: 0x7c00000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x10, direct128: false, withFlags128: false,
	},
	"empty": {
		raw32: 0x7c000000, rawFlags32: 0x10, direct32: false, withFlags32: false,
		raw64: 0x7c00000000000000, rawFlags64: 0x10, direct64: false, withFlags64: false,
		raw128hi: 0x7c00000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x10, direct128: false, withFlags128: false,
	},
	"space": {
		raw32: 0x7c000000, rawFlags32: 0x10, direct32: false, withFlags32: false,
		raw64: 0x7c00000000000000, rawFlags64: 0x10, direct64: false, withFlags64: false,
		raw128hi: 0x7c00000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x10, direct128: false, withFlags128: false,
	},
	"tab": {
		raw32: 0x7c000000, rawFlags32: 0x10, direct32: false, withFlags32: false,
		raw64: 0x7c00000000000000, rawFlags64: 0x10, direct64: false, withFlags64: false,
		raw128hi: 0x7c00000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x10, direct128: false, withFlags128: false,
	},
	"dot": {
		raw32: 0x7c000000, rawFlags32: 0x10, direct32: false, withFlags32: false,
		raw64: 0x7c00000000000000, rawFlags64: 0x10, direct64: false, withFlags64: false,
		raw128hi: 0x7c00000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x10, direct128: false, withFlags128: false,
	},
	"dot_e1": {
		raw32: 0x7c000000, rawFlags32: 0x10, direct32: false, withFlags32: false,
		raw64: 0x7c00000000000000, rawFlags64: 0x10, direct64: false, withFlags64: false,
		raw128hi: 0x7c00000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x10, direct128: false, withFlags128: false,
	},
	"1e_bare": {
		raw32: 0x7c000000, rawFlags32: 0x10, direct32: false, withFlags32: false,
		raw64: 0x7c00000000000000, rawFlags64: 0x10, direct64: false, withFlags64: false,
		raw128hi: 0x7c00000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x10, direct128: false, withFlags128: false,
	},
	"1e_plus": {
		raw32: 0x7c000000, rawFlags32: 0x10, direct32: false, withFlags32: false,
		raw64: 0x7c00000000000000, rawFlags64: 0x10, direct64: false, withFlags64: false,
		raw128hi: 0x7c00000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x10, direct128: false, withFlags128: false,
	},
	"1e_minus": {
		raw32: 0x7c000000, rawFlags32: 0x10, direct32: false, withFlags32: false,
		raw64: 0x7c00000000000000, rawFlags64: 0x10, direct64: false, withFlags64: false,
		raw128hi: 0x7c00000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x10, direct128: false, withFlags128: false,
	},
	"e5": {
		raw32: 0x7c000000, rawFlags32: 0x10, direct32: false, withFlags32: false,
		raw64: 0x7c00000000000000, rawFlags64: 0x10, direct64: false, withFlags64: false,
		raw128hi: 0x7c00000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x10, direct128: false, withFlags128: false,
	},
	"plus": {
		raw32: 0x7c000000, rawFlags32: 0x10, direct32: false, withFlags32: false,
		raw64: 0x7c00000000000000, rawFlags64: 0x10, direct64: false, withFlags64: false,
		raw128hi: 0x7c00000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x10, direct128: false, withFlags128: false,
	},
	"minus": {
		raw32: 0x7c000000, rawFlags32: 0x10, direct32: false, withFlags32: false,
		raw64: 0x7c00000000000000, rawFlags64: 0x10, direct64: false, withFlags64: false,
		raw128hi: 0x7c00000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x10, direct128: false, withFlags128: false,
	},
	"double_minus": {
		raw32: 0x7c000000, rawFlags32: 0x10, direct32: false, withFlags32: false,
		raw64: 0x7c00000000000000, rawFlags64: 0x10, direct64: false, withFlags64: false,
		raw128hi: 0x7c00000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x10, direct128: false, withFlags128: false,
	},
	"1..2": {
		raw32: 0x7c000000, rawFlags32: 0x10, direct32: false, withFlags32: false,
		raw64: 0x7c00000000000000, rawFlags64: 0x10, direct64: false, withFlags64: false,
		raw128hi: 0x7c00000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x10, direct128: false, withFlags128: false,
	},
	"1.2.3": {
		raw32: 0x7c000000, rawFlags32: 0x10, direct32: false, withFlags32: false,
		raw64: 0x7c00000000000000, rawFlags64: 0x10, direct64: false, withFlags64: false,
		raw128hi: 0x7c00000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x10, direct128: false, withFlags128: false,
	},
	"1_sp_e5": {
		raw32: 0x7c000000, rawFlags32: 0x10, direct32: false, withFlags32: false,
		raw64: 0x7c00000000000000, rawFlags64: 0x10, direct64: false, withFlags64: false,
		raw128hi: 0x7c00000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x10, direct128: false, withFlags128: false,
	},
	"1e5_sp": {
		raw32: 0x7c000000, rawFlags32: 0x10, direct32: false, withFlags32: false,
		raw64: 0x7c00000000000000, rawFlags64: 0x10, direct64: false, withFlags64: false,
		raw128hi: 0x7c00000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x10, direct128: false, withFlags128: false,
	},
	"1.0_sp": {
		raw32: 0x7c000000, rawFlags32: 0x10, direct32: false, withFlags32: false,
		raw64: 0x7c00000000000000, rawFlags64: 0x10, direct64: false, withFlags64: false,
		raw128hi: 0x7c00000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x10, direct128: false, withFlags128: false,
	},
	"sp_1.5": {
		raw32: 0x3200000f, rawFlags32: 0x0, direct32: true, withFlags32: true,
		raw64: 0x31a000000000000f, rawFlags64: 0x0, direct64: true, withFlags64: true,
		raw128hi: 0x303e000000000000, raw128lo: 0x000000000000000f, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"tab_1.5": {
		raw32: 0x3200000f, rawFlags32: 0x0, direct32: true, withFlags32: true,
		raw64: 0x31a000000000000f, rawFlags64: 0x0, direct64: true, withFlags64: true,
		raw128hi: 0x303e000000000000, raw128lo: 0x000000000000000f, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"+.5": {
		raw32: 0x32000005, rawFlags32: 0x0, direct32: true, withFlags32: true,
		raw64: 0x31a0000000000005, rawFlags64: 0x0, direct64: true, withFlags64: true,
		raw128hi: 0x303e000000000000, raw128lo: 0x0000000000000005, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"comma": {
		raw32: 0x7c000000, rawFlags32: 0x10, direct32: false, withFlags32: false,
		raw64: 0x7c00000000000000, rawFlags64: 0x10, direct64: false, withFlags64: false,
		raw128hi: 0x7c00000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x10, direct128: false, withFlags128: false,
	},
	"hex": {
		raw32: 0x7c000000, rawFlags32: 0x10, direct32: false, withFlags32: false,
		raw64: 0x7c00000000000000, rawFlags64: 0x10, direct64: false, withFlags64: false,
		raw128hi: 0x7c00000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x10, direct128: false, withFlags128: false,
	},
	"arabic_digits": {
		raw32: 0x7c000000, rawFlags32: 0x10, direct32: false, withFlags32: false,
		raw64: 0x7c00000000000000, rawFlags64: 0x10, direct64: false, withFlags64: false,
		raw128hi: 0x7c00000000000000, raw128lo: 0x0000000000000000, rawFlags128: 0x10, direct128: false, withFlags128: false,
	},
	"1e+5": {
		raw32: 0x35000001, rawFlags32: 0x0, direct32: true, withFlags32: true,
		raw64: 0x3260000000000001, rawFlags64: 0x0, direct64: true, withFlags64: true,
		raw128hi: 0x304a000000000000, raw128lo: 0x0000000000000001, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
	"1E5": {
		raw32: 0x35000001, rawFlags32: 0x0, direct32: true, withFlags32: true,
		raw64: 0x3260000000000001, rawFlags64: 0x0, direct64: true, withFlags64: true,
		raw128hi: 0x304a000000000000, raw128lo: 0x0000000000000001, rawFlags128: 0x0, direct128: true, withFlags128: true,
	},
}

var parseBoundaryModePins = map[parseBoundaryModeKey]parseBoundaryModePin{
	parseBoundaryModeKey{"1.0000005", RoundTowardZero}: {
		v32: 0x2f8f4240, f32: 0x1, ok32: true,
		v64: 0x30e0000000989685, f64: 0x0, ok64: true,
		v128hi: 0x3032000000000000, v128lo: 0x0000000000989685, f128: 0x0, ok128: true,
	},
	parseBoundaryModeKey{"1.0000005", RoundNearestAway}: {
		v32: 0x2f8f4241, f32: 0x1, ok32: true,
		v64: 0x30e0000000989685, f64: 0x0, ok64: true,
		v128hi: 0x3032000000000000, v128lo: 0x0000000000989685, f128: 0x0, ok128: true,
	},
	parseBoundaryModeKey{"-1.0000005", RoundTowardNegative}: {
		v32: 0xaf8f4241, f32: 0x1, ok32: true,
		v64: 0xb0e0000000989685, f64: 0x0, ok64: true,
		v128hi: 0xb032000000000000, v128lo: 0x0000000000989685, f128: 0x0, ok128: true,
	},
	parseBoundaryModeKey{"-1.0000005", RoundNearestAway}: {
		v32: 0xaf8f4241, f32: 0x1, ok32: true,
		v64: 0xb0e0000000989685, f64: 0x0, ok64: true,
		v128hi: 0xb032000000000000, v128lo: 0x0000000000989685, f128: 0x0, ok128: true,
	},
	parseBoundaryModeKey{"9999999.5", RoundTowardZero}: {
		v32: 0x6cb8967f, f32: 0x1, ok32: true,
		v64: 0x31a0000005f5e0fb, f64: 0x0, ok64: true,
		v128hi: 0x303e000000000000, v128lo: 0x0000000005f5e0fb, f128: 0x0, ok128: true,
	},
	parseBoundaryModeKey{"9999999.5", RoundNearestAway}: {
		v32: 0x330f4240, f32: 0x1, ok32: true,
		v64: 0x31a0000005f5e0fb, f64: 0x0, ok64: true,
		v128hi: 0x303e000000000000, v128lo: 0x0000000005f5e0fb, f128: 0x0, ok128: true,
	},
	parseBoundaryModeKey{"d64_tie", RoundTowardZero}: {
		v32: 0x2f8f4240, f32: 0x1, ok32: true,
		v64: 0x2fe38d7ea4c68000, f64: 0x1, ok64: true,
		v128hi: 0x3020000000000000, v128lo: 0x002386f26fc10005, f128: 0x0, ok128: true,
	},
	parseBoundaryModeKey{"d64_tie", RoundNearestAway}: {
		v32: 0x2f8f4240, f32: 0x1, ok32: true,
		v64: 0x2fe38d7ea4c68001, f64: 0x1, ok64: true,
		v128hi: 0x3020000000000000, v128lo: 0x002386f26fc10005, f128: 0x0, ok128: true,
	},
	parseBoundaryModeKey{"d128_tie", RoundTowardZero}: {
		v32: 0x2f8f4240, f32: 0x1, ok32: true,
		v64: 0x2fe38d7ea4c68000, f64: 0x1, ok64: true,
		v128hi: 0x2ffe314dc6448d93, v128lo: 0x38c15b0a00000000, f128: 0x1, ok128: true,
	},
	parseBoundaryModeKey{"d128_tie", RoundNearestAway}: {
		v32: 0x2f8f4240, f32: 0x1, ok32: true,
		v64: 0x2fe38d7ea4c68000, f64: 0x1, ok64: true,
		v128hi: 0x2ffe314dc6448d93, v128lo: 0x38c15b0a00000001, f128: 0x1, ok128: true,
	},
	parseBoundaryModeKey{"1e97", RoundTowardZero}: {
		v32: 0x77f8967f, f32: 0x5, ok32: true,
		v64: 0x3de0000000000001, f64: 0x0, ok64: true,
		v128hi: 0x3102000000000000, v128lo: 0x0000000000000001, f128: 0x0, ok128: true,
	},
	parseBoundaryModeKey{"1e97", RoundTowardNegative}: {
		v32: 0x77f8967f, f32: 0x5, ok32: true,
		v64: 0x3de0000000000001, f64: 0x0, ok64: true,
		v128hi: 0x3102000000000000, v128lo: 0x0000000000000001, f128: 0x0, ok128: true,
	},
	parseBoundaryModeKey{"1e-102", RoundTowardPositive}: {
		v32: 0x00000000, f32: 0x3, ok32: true,
		v64: 0x2500000000000001, f64: 0x0, ok64: true,
		v128hi: 0x2f74000000000000, v128lo: 0x0000000000000001, f128: 0x0, ok128: true,
	},
	parseBoundaryModeKey{"1e-102", RoundTowardZero}: {
		v32: 0x00000000, f32: 0x3, ok32: true,
		v64: 0x2500000000000001, f64: 0x0, ok64: true,
		v128hi: 0x2f74000000000000, v128lo: 0x0000000000000001, f128: 0x0, ok128: true,
	},
	parseBoundaryModeKey{"1e385", RoundTowardZero}: {
		v32: 0x77f8967f, f32: 0x5, ok32: true,
		v64: 0x77fb86f26fc0ffff, f64: 0x5, ok64: true,
		v128hi: 0x3342000000000000, v128lo: 0x0000000000000001, f128: 0x0, ok128: true,
	},
	parseBoundaryModeKey{"1e6145", RoundTowardZero}: {
		v32: 0x77f8967f, f32: 0x5, ok32: true,
		v64: 0x77fb86f26fc0ffff, f64: 0x5, ok64: true,
		v128hi: 0x5fffed09bead87c0, v128lo: 0x378d8e63ffffffff, f128: 0x5, ok128: true,
	},
}
