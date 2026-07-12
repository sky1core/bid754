package bid754

import (
	"encoding/binary"
	"math/big"
	"testing"
)

// The public NaN literal parser only accepts canonical payloads (BID32 < 10^6,
// BID64 < 10^15, BID128 < 10^33). String() must therefore never render a larger
// raw payload: doing so would break the String()->parse round-trip and make the
// three widths behave inconsistently. These tests pin that a noncanonical
// payload collapses to a bare NaN token on every width, that canonical payloads
// still render, and that the round-trip returns a canonical NaN.
//
// The mechanical port never produces a noncanonical NaN payload (arithmetic
// canonicalizes), so these encodings are reached only by constructing raw bits
// (e.g. via a codec decode of a noncanonical NaN); no regular verification
// domain exercises this public formatting path, which is why it is pinned here.

func TestNaNPayloadStringSuppressesNoncanonical32(t *testing.T) {
	cases := []struct {
		name string
		bits uint32
		want string
	}{
		{"canonical qNaN max", 0x7c000000 | 999999, "+NaN999999"},
		{"canonical sNaN", 0x7e000000 | 999999, "+SNaN999999"},
		{"noncanonical qNaN just over limit", 0x7c000000 | 1000000, "+NaN"},
		{"noncanonical qNaN mask max", 0x7c000000 | 0x000fffff, "+NaN"},
		{"noncanonical sNaN", 0x7e000000 | 1000000, "+SNaN"},
		{"noncanonical negative qNaN", 0xfc000000 | 1000000, "-NaN"},
		{"zero payload qNaN", 0x7c000000, "+NaN"},
	}
	for _, tc := range cases {
		d := Decimal32BID(tc.bits)
		if got := d.String(); got != tc.want {
			t.Errorf("%s: Decimal32BID(%#x).String() = %q, want %q", tc.name, tc.bits, got, tc.want)
		}
	}
}

func TestNaNPayloadStringSuppressesNoncanonical64(t *testing.T) {
	const canonicalLimit = uint64(1000000000000000) // 10^15
	cases := []struct {
		name string
		bits uint64
		want string
	}{
		{"canonical qNaN max", 0x7c00000000000000 | (canonicalLimit - 1), "+NaN999999999999999"},
		{"noncanonical qNaN at limit", 0x7c00000000000000 | canonicalLimit, "+NaN"},
		{"noncanonical qNaN mask max", 0x7c00000000000000 | 0x0003ffffffffffff, "+NaN"},
		{"noncanonical sNaN", 0x7e00000000000000 | canonicalLimit, "+SNaN"},
		{"zero payload qNaN", 0x7c00000000000000, "+NaN"},
	}
	for _, tc := range cases {
		d := Decimal64BID(tc.bits)
		if got := d.String(); got != tc.want {
			t.Errorf("%s: Decimal64BID(%#x).String() = %q, want %q", tc.name, tc.bits, got, tc.want)
		}
	}
}

// build128NaN assembles a raw Decimal128BID NaN from a payload magnitude. The
// project targets little-endian only (PLATFORM_SPEC), matching how the
// formatter reads the bytes; the low 64 payload bits go in bytes 0..7 and the
// combination field (payload high bits plus the NaN/sign tag) in bytes 8..15.
func build128NaN(payload *big.Int, signaling, negative bool) Decimal128BID {
	lo := new(big.Int).And(payload, new(big.Int).SetUint64(^uint64(0))).Uint64()
	hi := new(big.Int).Rsh(payload, 64).Uint64()
	hi |= 0x7c00000000000000
	if signaling {
		hi = (hi &^ uint64(0x7c00000000000000)) | 0x7e00000000000000
	}
	if negative {
		hi |= 0x8000000000000000
	}
	var d Decimal128BID
	binary.LittleEndian.PutUint64(d[0:8], lo)
	binary.LittleEndian.PutUint64(d[8:16], hi)
	return d
}

func TestNaNPayloadStringSuppressesNoncanonical128(t *testing.T) {
	limit := new(big.Int).Exp(big.NewInt(10), big.NewInt(33), nil) // 10^33
	canonical := new(big.Int).Sub(limit, big.NewInt(1))            // 10^33 - 1, max canonical

	if got := build128NaN(canonical, false, false).String(); got != "+NaN"+canonical.String() {
		t.Errorf("canonical Decimal128BID payload String() = %q, want +NaN%s", got, canonical.String())
	}
	if got := build128NaN(limit, false, false).String(); got != "+NaN" {
		t.Errorf("noncanonical Decimal128BID payload (10^33) String() = %q, want +NaN", got)
	}
	if got := build128NaN(limit, true, true).String(); got != "-SNaN" {
		t.Errorf("noncanonical Decimal128BID sNaN payload (10^33) String() = %q, want -SNaN", got)
	}
}

// A noncanonical payload that String() suppresses must round-trip back to a
// canonical (zero-payload) NaN of the same sign and signaling kind.
func TestNaNPayloadStringRoundTripDropsNoncanonical(t *testing.T) {
	t.Run("Decimal32", func(t *testing.T) {
		orig := Decimal32BID(0x7c000000 | 1000000)
		s := orig.String()
		if s != "+NaN" {
			t.Fatalf("noncanonical String() = %q, want +NaN", s)
		}
		reparsed, err := NewDecimal32BIDDirect(s)
		if err != nil {
			t.Fatalf("reparse %q: %v", s, err)
		}
		if !reparsed.IsNaN() {
			t.Fatalf("reparsed %q is not NaN", s)
		}
		if got := reparsed.String(); got != "+NaN" {
			t.Fatalf("reparsed String() = %q, want +NaN", got)
		}
		if uint32(reparsed)&0x000fffff != 0 {
			t.Fatalf("reparsed payload = %#x, want canonical zero payload", uint32(reparsed)&0x000fffff)
		}
	})

	t.Run("Decimal64", func(t *testing.T) {
		orig := Decimal64BID(0x7e00000000000000 | 1000000000000000)
		s := orig.String()
		if s != "+SNaN" {
			t.Fatalf("noncanonical String() = %q, want +SNaN", s)
		}
		reparsed, err := NewDecimal64BIDDirect(s)
		if err != nil {
			t.Fatalf("reparse %q: %v", s, err)
		}
		if !reparsed.IsNaN() {
			t.Fatalf("reparsed %q is not NaN", s)
		}
		if got := reparsed.String(); got != "+SNaN" {
			t.Fatalf("reparsed String() = %q, want +SNaN", got)
		}
		if uint64(reparsed)&0x0003ffffffffffff != 0 {
			t.Fatalf("reparsed payload = %#x, want canonical zero payload", uint64(reparsed)&0x0003ffffffffffff)
		}
	})
}
