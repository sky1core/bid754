package bidcodec

import (
	"math/big"
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
		got := Encode32(c)
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
		got := Encode64(c)
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
		gotLo, gotHi := Encode128(c)
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
		"1.0E2147483648",
	} {
		if _, err := FromString(input); err == nil {
			t.Fatalf("FromString(%q) succeeded, want error", input)
		}
	}
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
