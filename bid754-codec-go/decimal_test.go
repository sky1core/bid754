package bidcodec

import (
	"fmt"
	"math/big"
	"strings"
	"testing"
)

func TestDecode32Basic(t *testing.T) {
	tests := []struct {
		name string
		v    uint32
		want Components
	}{
		{"zero", 0x32800000, Components{Kind: Zero, Exponent: 0}},
		{"neg_zero", 0xb2800000, Components{Sign: true, Kind: Zero, Exponent: 0}},
		{"one", 0x32800001, Components{Kind: Normal, Coefficient: big.NewInt(1), Exponent: 0}},
		{"neg_one", 0xb2800001, Components{Sign: true, Kind: Normal, Coefficient: big.NewInt(1), Exponent: 0}},
		{"inf", 0x78000000, Components{Kind: Infinity}},
		{"neg_inf", 0xf8000000, Components{Sign: true, Kind: Infinity}},
		{"qnan", 0x7c000000, Components{Kind: QNaN}},
		{"snan", 0x7e000000, Components{Kind: SNaN}},
		{"9999999", 0x77f8967f, Components{Kind: Normal, Coefficient: big.NewInt(9999999), Exponent: 90}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Decode32(tt.v)
			if got.Sign != tt.want.Sign || got.Kind != tt.want.Kind || got.Exponent != tt.want.Exponent {
				t.Errorf("Decode32(0x%08x) sign/kind/exp = %v/%v/%d, want %v/%v/%d",
					tt.v, got.Sign, got.Kind, got.Exponent, tt.want.Sign, tt.want.Kind, tt.want.Exponent)
			}
			if tt.want.Coefficient != nil && (got.Coefficient == nil || got.Coefficient.Cmp(tt.want.Coefficient) != 0) {
				t.Errorf("Decode32(0x%08x) coeff = %v, want %v", tt.v, got.Coefficient, tt.want.Coefficient)
			}
		})
	}
}

func TestRoundtrip32(t *testing.T) {
	values := []uint32{
		0x32800000, // +0
		0xb2800000, // -0
		0x32800001, // +1
		0x32800064, // +100
		0x77f8967f, // 9999999 * 10^90 (special encoding)
		0x78000000, // +inf
		0xf8000000, // -inf
		0x7c000000, // NaN
		0x7e000000, // sNaN
	}
	for _, v := range values {
		c := Decode32(v)
		got, err := Encode32(c)
		if err != nil {
			t.Fatalf("Encode32 roundtrip 0x%08x: unexpected error: %v", v, err)
		}
		if got != v {
			t.Errorf("roundtrip 0x%08x: got 0x%08x", v, got)
		}
	}
}

func TestDecode64Basic(t *testing.T) {
	tests := []struct {
		name string
		v    uint64
		want Components
	}{
		{"zero", 0x31c0000000000000, Components{Kind: Zero, Exponent: 0}},
		{"one", 0x31c0000000000001, Components{Kind: Normal, Coefficient: big.NewInt(1), Exponent: 0}},
		{"inf", 0x7800000000000000, Components{Kind: Infinity}},
		{"qnan", 0x7c00000000000000, Components{Kind: QNaN}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Decode64(tt.v)
			if got.Kind != tt.want.Kind || got.Exponent != tt.want.Exponent {
				t.Errorf("Decode64(0x%016x) kind/exp = %v/%d, want %v/%d",
					tt.v, got.Kind, got.Exponent, tt.want.Kind, tt.want.Exponent)
			}
		})
	}
}

func TestRoundtrip64(t *testing.T) {
	values := []uint64{
		0x31c0000000000000, // +0
		0xb1c0000000000000, // -0
		0x31c0000000000001, // +1
		0x7800000000000000, // +inf
		0x7c00000000000000, // NaN
		0x7e00000000000000, // sNaN
	}
	for _, v := range values {
		c := Decode64(v)
		got, err := Encode64(c)
		if err != nil {
			t.Fatalf("Encode64 roundtrip 0x%016x: unexpected error: %v", v, err)
		}
		if got != v {
			t.Errorf("roundtrip 0x%016x: got 0x%016x", v, got)
		}
	}
}

func TestDecode128Basic(t *testing.T) {
	// +1E+0: biased exp=6176, coeff=1
	lo := uint64(0x0000000000000001)
	hi := uint64(6176) << 49

	c := Decode128(lo, hi)
	if c.Kind != Normal || c.Exponent != 0 || c.Coefficient.Cmp(big.NewInt(1)) != 0 || c.Sign {
		t.Errorf("Decode128(+1) = %+v", c)
	}
}

func TestRoundtrip128(t *testing.T) {
	cases := [][2]uint64{
		{0, uint64(6176) << 49},                // +0
		{0, bid128SignMask | uint64(6176)<<49}, // -0
		{1, uint64(6176) << 49},                // +1
		{0, 0x7800000000000000},                // +inf
		{0, 0x7c00000000000000},                // NaN
	}
	for _, tc := range cases {
		c := Decode128(tc[0], tc[1])
		gotLo, gotHi, err := Encode128(c)
		if err != nil {
			t.Fatalf("Encode128 roundtrip %016x_%016x: unexpected error: %v", tc[1], tc[0], err)
		}
		if gotLo != tc[0] || gotHi != tc[1] {
			t.Errorf("roundtrip %016x_%016x: got %016x_%016x", tc[1], tc[0], gotHi, gotLo)
		}
	}
}

func TestDecodeBytesExactLength(t *testing.T) {
	tests := []struct {
		name string
		fn   func() error
	}{
		{"Decode32Bytes short", func() error { _, err := Decode32Bytes(make([]byte, 3)); return err }},
		{"Decode32Bytes long", func() error { _, err := Decode32Bytes(make([]byte, 5)); return err }},
		{"Decode64Bytes short", func() error { _, err := Decode64Bytes(make([]byte, 7)); return err }},
		{"Decode64Bytes long", func() error { _, err := Decode64Bytes(make([]byte, 9)); return err }},
		{"Decode128Bytes short", func() error { _, err := Decode128Bytes(make([]byte, 15)); return err }},
		{"Decode128Bytes long", func() error { _, err := Decode128Bytes(make([]byte, 17)); return err }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.fn(); err == nil {
				t.Fatalf("expected error")
			}
		})
	}
}

func TestFromStringRejectsMalformedInputs(t *testing.T) {
	for _, input := range []string{
		"",
		"NaNabc",
		"SNaN-1",
		"1.2.3",
		"1E",
		"1Eabc",
		"1E2147483648",
		// Strict-ASCII grammar divergence cases (previously accepted by some
		// language ports through lenient stdlib parsers).
		"NaN+5", // payload leading sign
		"1E１",   // fullwidth Unicode digit in exponent
		"1E1_0", // underscore digit group in exponent
		"１２３",   // fullwidth Unicode digits in coefficient
		"1E 5",  // embedded ASCII whitespace in exponent
		" 1",    // leading NBSP (Unicode whitespace) before the token
	} {
		if _, err := FromString(input); err == nil {
			t.Fatalf("FromString(%q) succeeded, want error", input)
		}
	}
}

func TestFromStringAcceptsValidInputs(t *testing.T) {
	for _, input := range []string{
		"1.5",
		"+1.23E+5",
		"-inf",
		"NaN123",
		"1.",
		".5",
		"007",
	} {
		if _, err := FromString(input); err != nil {
			t.Fatalf("FromString(%q) failed: %v", input, err)
		}
	}
}

// TestFromStringSchemaCoefficientCap pins the schema-wide coefficient cap: the
// parsed value (leading zeros removed) must not exceed 10^34-1, the largest
// coefficient any supported BID width can hold. The cap is value-based, not
// digit-count-based, and identical in all six language packages.
func TestFromStringSchemaCoefficientCap(t *testing.T) {
	// 34 nines = 10^34-1, the schema max: accepted with the exact value.
	c, err := FromString(strings.Repeat("9", 34))
	if err != nil {
		t.Fatalf("FromString(34 nines) failed: %v", err)
	}
	if c.Kind != Normal || c.Coefficient.Cmp(bigStr(t, "9999999999999999999999999999999999")) != 0 {
		t.Fatalf("FromString(34 nines) = %+v, want coefficient 10^34-1", c)
	}
	// 35 nines: value above the schema cap, rejected.
	if _, err := FromString(strings.Repeat("9", 35)); err == nil {
		t.Fatal("FromString(35 nines) succeeded, want error")
	}
	// 41 digits but value 1: accepted, because the cap applies to the parsed
	// value after leading-zero removal, not to the digit count.
	c, err = FromString(strings.Repeat("0", 40) + "1")
	if err != nil {
		t.Fatalf("FromString(40 zeros + 1) failed: %v", err)
	}
	if c.Kind != Normal || c.Coefficient.Cmp(big.NewInt(1)) != 0 {
		t.Fatalf("FromString(40 zeros + 1) = %+v, want coefficient 1", c)
	}
}

// TestFromStringExponentInt32Bounds pins the exponent rule: the exponent
// literal must be below the shared exact-integer bound 2^53 in magnitude
// (identical in every language consumer), and only the fraction-adjusted
// FINAL exponent must fit a signed 32-bit integer — so every ToString
// rendering (adjusted exponent up to int32 max + 33) reparses (round-trip
// closure).
func TestFromStringExponentInt32Bounds(t *testing.T) {
	// Literal exceeds int32 but the fraction-adjusted final exponent fits:
	// accepted (this is exactly the shape ToString renders for near-max
	// exponents, e.g. FromString("10E2147483647") -> "+1.0E+2147483648").
	c, err := FromString("0.001E2147483649")
	if err != nil {
		t.Fatalf(`FromString("0.001E2147483649") failed: %v`, err)
	}
	if c.Exponent != 2147483646 {
		t.Fatalf(`FromString("0.001E2147483649") exponent = %d, want 2147483646`, c.Exponent)
	}
	c, err = FromString("1.0E2147483648")
	if err != nil {
		t.Fatalf(`FromString("1.0E2147483648") failed: %v`, err)
	}
	if c.Exponent != 2147483647 {
		t.Fatalf(`FromString("1.0E2147483648") exponent = %d, want 2147483647`, c.Exponent)
	}
	// Literal and fraction-adjusted value both fit: accepted.
	c, err = FromString("1.5E2147483647")
	if err != nil {
		t.Fatalf(`FromString("1.5E2147483647") failed: %v`, err)
	}
	if c.Exponent != 2147483646 {
		t.Fatalf(`FromString("1.5E2147483647") exponent = %d, want 2147483646`, c.Exponent)
	}
	// Fraction-adjusted final exponent leaves int32 (one past either edge):
	// rejected.
	if _, err := FromString("1.0E-2147483648"); err == nil {
		t.Fatal(`FromString("1.0E-2147483648") succeeded, want error`)
	}
	if _, err := FromString("1.0E+2147483649"); err == nil {
		t.Fatal(`FromString("1.0E+2147483649") succeeded, want error`)
	}
	// Literal at/beyond the shared 2^53 exact-integer bound: rejected at the
	// literal step (same error channel), both at the exact edge and far past
	// int64.
	if _, err := FromString("1E9007199254740992"); err == nil {
		t.Fatal(`FromString("1E9007199254740992") succeeded, want error`)
	}
	if _, err := FromString("1E-9007199254740992"); err == nil {
		t.Fatal(`FromString("1E-9007199254740992") succeeded, want error`)
	}
	if _, err := FromString("1E" + strings.Repeat("9", 25)); err == nil {
		t.Fatal(`FromString(25-nine exponent literal) succeeded, want error`)
	}
	// Round-trip closure at the rendered edge: parse(render(x)) succeeds and is
	// a fixed point.
	first, err := FromString("10E2147483647")
	if err != nil {
		t.Fatalf(`FromString("10E2147483647") failed: %v`, err)
	}
	rendered, err := ToString(first)
	if err != nil {
		t.Fatalf("ToString parsed boundary: %v", err)
	}
	if rendered != "+1.0E+2147483648" {
		t.Fatalf(`ToString(FromString("10E2147483647")) = %q, want "+1.0E+2147483648"`, rendered)
	}
	again, err := FromString(rendered)
	if err != nil {
		t.Fatalf("FromString(%q) failed: %v", rendered, err)
	}
	if got, err := ToString(again); err != nil || got != rendered {
		t.Fatalf("round-trip not a fixed point: %q -> %q", rendered, got)
	}
}

// bigStr builds a *big.Int from a decimal string, failing the test on error.
func bigStr(t *testing.T, s string) *big.Int {
	t.Helper()
	v, ok := new(big.Int).SetString(s, 10)
	if !ok {
		t.Fatalf("bad big.Int literal %q", s)
	}
	return v
}

// TestEncodeBoundaries checks that each field's exact upper/lower bound encodes
// successfully while one step past it is rejected, and that malformed
// Components (negative/nil coefficient, unknown kind, previously-crashing
// oversized coefficient) are rejected rather than truncated or panicking.
func TestEncodeBoundaries(t *testing.T) {
	pow2_128 := new(big.Int).Lsh(big.NewInt(1), 128) // 2^128, the historic FillBytes crash input

	// --- coefficient upper bound (Normal) ---
	t.Run("bid32 coefficient max ok", func(t *testing.T) {
		if _, err := Encode32(Components{Kind: Normal, Coefficient: big.NewInt(9999999)}); err != nil {
			t.Fatalf("coefficient 9999999 rejected: %v", err)
		}
	})
	t.Run("bid32 coefficient over", func(t *testing.T) {
		if _, err := Encode32(Components{Kind: Normal, Coefficient: big.NewInt(10000000)}); err == nil {
			t.Fatal("coefficient 10000000 accepted, want reject")
		}
	})
	t.Run("bid64 coefficient max ok", func(t *testing.T) {
		if _, err := Encode64(Components{Kind: Normal, Coefficient: bigStr(t, "9999999999999999")}); err != nil {
			t.Fatalf("coefficient 9999999999999999 rejected: %v", err)
		}
	})
	t.Run("bid64 coefficient over", func(t *testing.T) {
		if _, err := Encode64(Components{Kind: Normal, Coefficient: bigStr(t, "10000000000000000")}); err == nil {
			t.Fatal("coefficient 10000000000000000 accepted, want reject")
		}
	})
	t.Run("bid128 coefficient max ok", func(t *testing.T) {
		if _, _, err := Encode128(Components{Kind: Normal, Coefficient: bigStr(t, "9999999999999999999999999999999999")}); err != nil {
			t.Fatalf("coefficient 10^34-1 rejected: %v", err)
		}
	})
	t.Run("bid128 coefficient over", func(t *testing.T) {
		if _, _, err := Encode128(Components{Kind: Normal, Coefficient: bigStr(t, "10000000000000000000000000000000000")}); err == nil {
			t.Fatal("coefficient 10^34 accepted, want reject")
		}
	})
	t.Run("bid128 coefficient 2^128 (historic crash) rejected", func(t *testing.T) {
		if _, _, err := Encode128(Components{Kind: Normal, Coefficient: pow2_128}); err == nil {
			t.Fatal("coefficient 2^128 accepted, want reject")
		}
	})

	// --- exponent range (Zero/Normal) ---
	expCases := []struct {
		name string
		enc  func(exp int32) error
		lo   int32
		hi   int32
	}{
		{"bid32", func(exp int32) error {
			_, err := Encode32(Components{Kind: Normal, Coefficient: big.NewInt(1), Exponent: exp})
			return err
		}, -101, 90},
		{"bid64", func(exp int32) error {
			_, err := Encode64(Components{Kind: Normal, Coefficient: big.NewInt(1), Exponent: exp})
			return err
		}, -398, 369},
		{"bid128", func(exp int32) error {
			_, _, err := Encode128(Components{Kind: Normal, Coefficient: big.NewInt(1), Exponent: exp})
			return err
		}, -6176, 6111},
	}
	for _, ec := range expCases {
		ec := ec
		t.Run(ec.name+" exponent bounds", func(t *testing.T) {
			if err := ec.enc(ec.hi); err != nil {
				t.Errorf("exponent %d rejected: %v", ec.hi, err)
			}
			if err := ec.enc(ec.lo); err != nil {
				t.Errorf("exponent %d rejected: %v", ec.lo, err)
			}
			if err := ec.enc(ec.hi + 1); err == nil {
				t.Errorf("exponent %d accepted, want reject", ec.hi+1)
			}
			if err := ec.enc(ec.lo - 1); err == nil {
				t.Errorf("exponent %d accepted, want reject", ec.lo-1)
			}
		})
	}

	// --- NaN payload limit, per width (10^6 / 10^15 / 10^33) ---
	t.Run("bid32 payload max ok", func(t *testing.T) {
		if _, err := Encode32(Components{Kind: QNaN, Payload: big.NewInt(999999)}); err != nil {
			t.Fatalf("payload 999999 rejected: %v", err)
		}
	})
	t.Run("bid32 payload over", func(t *testing.T) {
		if _, err := Encode32(Components{Kind: SNaN, Payload: big.NewInt(1000000)}); err == nil {
			t.Fatal("payload 1000000 accepted, want reject")
		}
	})
	t.Run("bid64 payload max ok", func(t *testing.T) {
		if _, err := Encode64(Components{Kind: QNaN, Payload: big.NewInt(999999999999999)}); err != nil {
			t.Fatalf("payload 999999999999999 rejected: %v", err)
		}
	})
	t.Run("bid64 payload over", func(t *testing.T) {
		if _, err := Encode64(Components{Kind: SNaN, Payload: big.NewInt(1000000000000000)}); err == nil {
			t.Fatal("payload 1000000000000000 accepted, want reject")
		}
	})
	t.Run("bid128 payload max ok", func(t *testing.T) {
		// 10^33-1, the widest canonical BID128 NaN payload (uses high bits).
		if _, _, err := Encode128(Components{Kind: QNaN, Payload: bigStr(t, "999999999999999999999999999999999")}); err != nil {
			t.Fatalf("payload 10^33-1 rejected: %v", err)
		}
	})
	t.Run("bid128 payload over", func(t *testing.T) {
		if _, _, err := Encode128(Components{Kind: SNaN, Payload: bigStr(t, "1000000000000000000000000000000000")}); err == nil {
			t.Fatal("payload 10^33 accepted, want reject")
		}
	})

	// --- field-domain violations (Go big.Int / Kind can express these) ---
	t.Run("negative coefficient rejected", func(t *testing.T) {
		if _, err := Encode32(Components{Kind: Normal, Coefficient: big.NewInt(-1)}); err == nil {
			t.Fatal("negative coefficient accepted, want reject")
		}
		if _, err := Encode64(Components{Kind: Normal, Coefficient: big.NewInt(-1)}); err == nil {
			t.Fatal("negative coefficient accepted, want reject")
		}
		if _, _, err := Encode128(Components{Kind: Normal, Coefficient: big.NewInt(-1)}); err == nil {
			t.Fatal("negative coefficient accepted, want reject")
		}
	})
	t.Run("negative payload rejected", func(t *testing.T) {
		// Payload is now a signed *big.Int, so a negative NaN payload is
		// constructible in Go and must be rejected in every width.
		if _, err := Encode32(Components{Kind: QNaN, Payload: big.NewInt(-1)}); err == nil {
			t.Fatal("bid32 negative payload accepted, want reject")
		}
		if _, err := Encode64(Components{Kind: QNaN, Payload: big.NewInt(-1)}); err == nil {
			t.Fatal("bid64 negative payload accepted, want reject")
		}
		if _, _, err := Encode128(Components{Kind: SNaN, Payload: big.NewInt(-1)}); err == nil {
			t.Fatal("bid128 negative payload accepted, want reject")
		}
	})
	t.Run("normal with nil coefficient rejected", func(t *testing.T) {
		if _, err := Encode32(Components{Kind: Normal, Coefficient: nil}); err == nil {
			t.Fatal("nil coefficient accepted, want reject")
		}
		if _, err := Encode64(Components{Kind: Normal, Coefficient: nil}); err == nil {
			t.Fatal("nil coefficient accepted, want reject")
		}
		if _, _, err := Encode128(Components{Kind: Normal, Coefficient: nil}); err == nil {
			t.Fatal("nil coefficient accepted, want reject")
		}
	})
	t.Run("unrecognized kind rejected", func(t *testing.T) {
		if _, err := Encode32(Components{Kind: Kind(99)}); err == nil {
			t.Fatal("Kind(99) accepted, want reject")
		}
		if _, err := Encode64(Components{Kind: Kind(99)}); err == nil {
			t.Fatal("Kind(99) accepted, want reject")
		}
		if _, _, err := Encode128(Components{Kind: Kind(99)}); err == nil {
			t.Fatal("Kind(99) accepted, want reject")
		}
	})

	// --- byte encoders share the same reject contract ---
	t.Run("byte encoders reject too", func(t *testing.T) {
		if _, err := Encode32Bytes(Components{Kind: Normal, Coefficient: big.NewInt(10000000)}); err == nil {
			t.Fatal("Encode32Bytes accepted over-max coefficient, want reject")
		}
		if _, err := Encode64Bytes(Components{Kind: Normal, Coefficient: bigStr(t, "10000000000000000")}); err == nil {
			t.Fatal("Encode64Bytes accepted over-max coefficient, want reject")
		}
		if _, err := Encode128Bytes(Components{Kind: Normal, Coefficient: pow2_128}); err == nil {
			t.Fatal("Encode128Bytes accepted 2^128 coefficient, want reject")
		}
	})
}

func TestComponentsToString(t *testing.T) {
	// +123.45 = 12345 * 10^-2
	c := Components{
		Kind:        Normal,
		Coefficient: big.NewInt(12345),
		Exponent:    -2,
	}
	// coefficient=12345, exponent=-2 -> 123.45
	if c.Coefficient.Int64() != 12345 || c.Exponent != -2 {
		t.Errorf("unexpected: %+v", c)
	}
}

// nanBits128 builds the (lo, hi) BID128 words for a NaN carrying the given
// payload (up to 110 bits), signaling flag, and sign. It packs the payload the
// same way Encode128 does but without the canonical-limit check, so tests can
// construct non-canonical (>= 10^33) inputs directly.
func nanBits128(payload *big.Int, signaling, sign bool) (lo, hi uint64) {
	lo = new(big.Int).And(payload, new(big.Int).SetUint64(^uint64(0))).Uint64()
	hi = new(big.Int).Rsh(payload, 64).Uint64() & 0x00003fffffffffff
	if signaling {
		hi |= 0x7e00000000000000
	} else {
		hi |= 0x7c00000000000000
	}
	if sign {
		hi |= bid128SignMask
	}
	return lo, hi
}

// TestBID128FullPayloadRoundtrip exercises the full 110-bit BID128 NaN payload
// (values above 2^64 that populate the high payload word) through the bit
// round-trip (Components -> bits -> Components -> bits) and the string
// round-trip (Components -> ToString -> FromString -> bits).
func TestBID128FullPayloadRoundtrip(t *testing.T) {
	payloads := []string{
		"18446744073709551616",              // 2^64: first value needing the high word
		"999999999999999999999999999999999", // 10^33-1: widest canonical payload
	}
	for _, ps := range payloads {
		for _, kind := range []Kind{QNaN, SNaN} {
			for _, sign := range []bool{false, true} {
				p := bigStr(t, ps)
				t.Run(fmt.Sprintf("payload=%s kind=%v sign=%v", ps, kind, sign), func(t *testing.T) {
					in := Components{Sign: sign, Kind: kind, Payload: p}

					lo, hi, err := Encode128(in)
					if err != nil {
						t.Fatalf("Encode128(%s): %v", ps, err)
					}
					// The high payload word must actually carry bits, proving we
					// are not silently dropping the payload above 2^64.
					if hi&0x00003fffffffffff == 0 {
						t.Fatalf("payload %s did not populate the high payload word (hi=%016x)", ps, hi)
					}

					// bits -> Components -> bits
					got := Decode128(lo, hi)
					if got.Kind != kind {
						t.Fatalf("Decode128 kind = %v, want %v", got.Kind, kind)
					}
					if got.Payload == nil || got.Payload.Cmp(p) != 0 {
						t.Fatalf("Decode128 payload = %v, want %s", got.Payload, ps)
					}
					lo2, hi2, err := Encode128(got)
					if err != nil {
						t.Fatalf("re-Encode128: %v", err)
					}
					if lo2 != lo || hi2 != hi {
						t.Fatalf("bit round-trip: got %016x_%016x, want %016x_%016x", hi2, lo2, hi, lo)
					}

					// Components -> string -> Components -> bits
					s, err := ToString(in)
					if err != nil {
						t.Fatalf("ToString: %v", err)
					}
					parsed, err := FromString(s)
					if err != nil {
						t.Fatalf("FromString(%q): %v", s, err)
					}
					lo3, hi3, err := Encode128(parsed)
					if err != nil {
						t.Fatalf("Encode128(FromString(%q)): %v", s, err)
					}
					if lo3 != lo || hi3 != hi {
						t.Fatalf("string round-trip via %q: got %016x_%016x, want %016x_%016x", s, hi3, lo3, hi, lo)
					}
				})
			}
		}
	}
}

// TestBID128PayloadBoundaryFromString pins the FromString NaN payload boundary
// at the schema-wide 10^33 limit: 10^33-1 parses, 10^33 is rejected (for both
// NaN and SNaN).
func TestBID128PayloadBoundaryFromString(t *testing.T) {
	ok, err := FromString("NaN999999999999999999999999999999999") // 10^33-1
	if err != nil {
		t.Fatalf("FromString(10^33-1 payload) rejected: %v", err)
	}
	if ok.Kind != QNaN || ok.Payload == nil || ok.Payload.Cmp(bigStr(t, "999999999999999999999999999999999")) != 0 {
		t.Fatalf("FromString(10^33-1 payload) = %+v", ok)
	}
	if _, err := FromString("NaN1000000000000000000000000000000000"); err == nil { // 10^33
		t.Fatal("FromString(NaN 10^33 payload) accepted, want reject")
	}
	if _, err := FromString("SNaN1000000000000000000000000000000000"); err == nil { // 10^33
		t.Fatal("FromString(SNaN 10^33 payload) accepted, want reject")
	}
}

// TestBID128NoncanonicalPayloadDecodesToZero verifies that a BID128 NaN whose
// encoded payload bits are at or above 10^33 (non-canonical) decodes to payload
// 0, the same normalization boundary the encode reject contract enforces.
func TestBID128NoncanonicalPayloadDecodesToZero(t *testing.T) {
	t.Run("payload=10^33", func(t *testing.T) {
		lo, hi := nanBits128(bigStr(t, "1000000000000000000000000000000000"), false, false)
		assertNaNPayloadZero(t, Decode128(lo, hi))
	})
	t.Run("all-payload-bits-set", func(t *testing.T) {
		// lo = 2^64-1, hi = NaN mask | full 46-bit payload field => payload 2^110-1.
		assertNaNPayloadZero(t, Decode128(0xffffffffffffffff, 0x7c003fffffffffff))
	})
}

func assertNaNPayloadZero(t *testing.T, got Components) {
	t.Helper()
	if got.Kind != QNaN {
		t.Fatalf("kind = %v, want QNaN", got.Kind)
	}
	if got.Payload == nil || got.Payload.Sign() != 0 {
		t.Fatalf("non-canonical payload decoded to %v, want 0", got.Payload)
	}
}

// TestNilPayloadEncodesAsZeroPayload documents that a nil Payload on a NaN is
// the default zero payload: it encodes identically to an explicit big.NewInt(0),
// in every width.
func TestNilPayloadEncodesAsZeroPayload(t *testing.T) {
	t.Run("bid32", func(t *testing.T) {
		nilBits, err := Encode32(Components{Kind: QNaN, Payload: nil})
		if err != nil {
			t.Fatalf("nil payload rejected: %v", err)
		}
		zeroBits, err := Encode32(Components{Kind: QNaN, Payload: big.NewInt(0)})
		if err != nil {
			t.Fatalf("zero payload rejected: %v", err)
		}
		if nilBits != zeroBits {
			t.Fatalf("nil payload = %08x, zero payload = %08x", nilBits, zeroBits)
		}
		if nilBits != 0x7c000000 {
			t.Fatalf("bare QNaN = %08x, want 7c000000", nilBits)
		}
	})
	t.Run("bid64", func(t *testing.T) {
		nilBits, err := Encode64(Components{Kind: QNaN, Payload: nil})
		if err != nil {
			t.Fatalf("nil payload rejected: %v", err)
		}
		zeroBits, err := Encode64(Components{Kind: QNaN, Payload: big.NewInt(0)})
		if err != nil {
			t.Fatalf("zero payload rejected: %v", err)
		}
		if nilBits != zeroBits {
			t.Fatalf("nil payload = %016x, zero payload = %016x", nilBits, zeroBits)
		}
	})
	t.Run("bid128", func(t *testing.T) {
		nilLo, nilHi, err := Encode128(Components{Kind: SNaN, Payload: nil})
		if err != nil {
			t.Fatalf("nil payload rejected: %v", err)
		}
		zeroLo, zeroHi, err := Encode128(Components{Kind: SNaN, Payload: big.NewInt(0)})
		if err != nil {
			t.Fatalf("zero payload rejected: %v", err)
		}
		if nilLo != zeroLo || nilHi != zeroHi {
			t.Fatalf("nil payload = %016x_%016x, zero payload = %016x_%016x", nilHi, nilLo, zeroHi, zeroLo)
		}
	})
}
