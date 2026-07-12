package apiemit

import (
	"math/big"
	"strings"
	"testing"
)

func TestParseSimplePositiveDecimal(t *testing.T) {
	cases := []struct {
		lit      string
		coeff    string
		exponent int
	}{
		{"0", "0", 0},
		{"1", "1", 0},
		{"3.141593", "3141593", -6},
		{"3.141592653589793", "3141592653589793", -15},
		{"3.141592653589793238462643383279503", "3141592653589793238462643383279503", -33},
		{"2.718282", "2718282", -6},
		{"2.718281828459045", "2718281828459045", -15},
		{"2.718281828459045235360287471352662", "2718281828459045235360287471352662", -33},
	}
	for _, c := range cases {
		coeff, exponent, err := parseSimplePositiveDecimal(c.lit)
		if err != nil {
			t.Fatalf("parseSimplePositiveDecimal(%q): unexpected error: %v", c.lit, err)
		}
		if coeff.String() != c.coeff {
			t.Errorf("parseSimplePositiveDecimal(%q): coeff = %s, want %s", c.lit, coeff, c.coeff)
		}
		if exponent != c.exponent {
			t.Errorf("parseSimplePositiveDecimal(%q): exponent = %d, want %d", c.lit, exponent, c.exponent)
		}
	}
}

func TestParseSimplePositiveDecimalRejectsOutOfScopeInput(t *testing.T) {
	cases := []struct {
		lit    string
		errSub string
	}{
		{"", "empty literal"},
		{"-1", "does not support"},
		{"+1", "does not support"},
		{"1e5", "does not support"},
		{"NaN", "does not support"},
		{"1.2.3", "more than one"},
		{".", "no digits"},
	}
	for _, c := range cases {
		_, _, err := parseSimplePositiveDecimal(c.lit)
		if err == nil || !strings.Contains(err.Error(), c.errSub) {
			t.Errorf("parseSimplePositiveDecimal(%q): expected error containing %q, got %v", c.lit, c.errSub, err)
		}
	}
}

// TestConstBitsRustLiteralMatchesIndependentCodecReference pins the 12
// ZERO/ONE/PI/E literal -> BID bits mappings against values independently
// cross-checked (during development, via a throwaway probe test, not
// committed) against devtools/internal/testgen/bid_codec_reference.go's
// already-validated refEncode32/64/128 -- the reference codec backing the
// BID codec vector-verification domain. A regression in encodeBID32/64/128
// (a wrong bias, a wrong field-width shift, a wrong byte order) would move
// one of these literal hex values; this test is the permanent record of that
// cross-check so it need not be re-derived by hand again.
func TestConstBitsRustLiteralMatchesIndependentCodecReference(t *testing.T) {
	cases := []struct {
		literal string
		w       widthSpec
		want    string
	}{
		{"0", decimal32Width, "0x32800000"},
		{"1", decimal32Width, "0x32800001"},
		{"3.141593", decimal32Width, "0x2fafefd9"},
		{"2.718282", decimal32Width, "0x2fa97a4a"},

		{"0", decimal64Width, "0x31c0000000000000"},
		{"1", decimal64Width, "0x31c0000000000001"},
		{"3.141592653589793", decimal64Width, "0x2feb29430a256d21"},
		{"2.718281828459045", decimal64Width, "0x2fe9a8434ec8e225"},

		{"0", decimal128Width, "[0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x40, 0x30]"},
		{"1", decimal128Width, "[0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x40, 0x30]"},
		{"3.141592653589793238462643383279503", decimal128Width, "[0x8f, 0x9f, 0xf3, 0xe6, 0x64, 0x55, 0xbe, 0xba, 0xa7, 0x96, 0x57, 0x79, 0xe4, 0x9a, 0xfe, 0x2f]"},
		{"2.718281828459045235360287471352662", decimal128Width, "[0x56, 0xbb, 0x6a, 0xb2, 0xcc, 0x6a, 0x90, 0x4e, 0xde, 0xf4, 0x4b, 0x8a, 0x05, 0x86, 0xfe, 0x2f]"},
	}
	for _, c := range cases {
		got, err := constBitsRustLiteral(c.literal, c.w)
		if err != nil {
			t.Fatalf("constBitsRustLiteral(%q, %s): unexpected error: %v", c.literal, c.w.selfType, err)
		}
		if got != c.want {
			t.Errorf("constBitsRustLiteral(%q, %s) = %s, want %s", c.literal, c.w.selfType, got, c.want)
		}
	}
}

func TestEncodeBID32RejectsLongForm(t *testing.T) {
	// 2^23 = 8388608, the first coefficient the direct/short BID32 field
	// cannot hold. Below 10^7 (a valid finite value), so it is the steering-
	// form branch -- not the decimal-bound branch -- that must reject it.
	coeff := new(big.Int).Lsh(big.NewInt(1), 23)
	_, err := encodeBID32(coeff, 0)
	if err == nil || !strings.Contains(err.Error(), "steering/long form") {
		t.Fatalf("expected a steering/long-form rejection, got: %v", err)
	}
}

func TestEncodeBID64RejectsLongForm(t *testing.T) {
	coeff := new(big.Int).Lsh(big.NewInt(1), 53)
	_, err := encodeBID64(coeff, 0)
	if err == nil || !strings.Contains(err.Error(), "steering/long form") {
		t.Fatalf("expected a steering/long-form rejection, got: %v", err)
	}
}

// TestEncodeRejectsCoefficientAtDecimalBound pins the High-review fix: the
// first INVALID (too-large finite) coefficient per width -- exactly 10^7 /
// 10^16 / 10^34 -- must fail with a decimal-bound error rather than be accepted by a
// bit-length approximation. The BID128 10^34 case is the one a BitLen()>113
// check missed (10^34 < 2^113): it must be rejected here.
func TestEncodeRejectsCoefficientAtDecimalBound(t *testing.T) {
	if _, err := encodeBID32(pow10(7), 0); err == nil || !strings.Contains(err.Error(), "< 10^7") {
		t.Fatalf("encodeBID32(10^7): expected a decimal-bound rejection, got: %v", err)
	}
	if _, err := encodeBID64(pow10(16), 0); err == nil || !strings.Contains(err.Error(), "< 10^16") {
		t.Fatalf("encodeBID64(10^16): expected a decimal-bound rejection, got: %v", err)
	}
	if _, _, err := encodeBID128(pow10(34), 0); err == nil || !strings.Contains(err.Error(), "< 10^34") {
		t.Fatalf("encodeBID128(10^34): expected a decimal-bound rejection, got: %v", err)
	}
}

// TestEncodeRejectsBiasedExponentOnePastMax pins the Medium-review fix: a
// biased exponent one past the finite maximum (192 / 768 / 12288, i.e.
// unbiased 91 / 370 / 6112) must fail with a finite-range error, not encode
// against the wider raw field width (0xff / 0x3ff / 0x3fff).
func TestEncodeRejectsBiasedExponentOnePastMax(t *testing.T) {
	one := big.NewInt(1)
	if _, err := encodeBID32(one, (bidBiasedExpMax32+1)-bidBias32); err == nil || !strings.Contains(err.Error(), "finite range") {
		t.Fatalf("encodeBID32 biased %d: expected an exponent-range rejection, got: %v", bidBiasedExpMax32+1, err)
	}
	if _, err := encodeBID64(one, (bidBiasedExpMax64+1)-bidBias64); err == nil || !strings.Contains(err.Error(), "finite range") {
		t.Fatalf("encodeBID64 biased %d: expected an exponent-range rejection, got: %v", bidBiasedExpMax64+1, err)
	}
	if _, _, err := encodeBID128(one, (bidBiasedExpMax128+1)-bidBias128); err == nil || !strings.Contains(err.Error(), "finite range") {
		t.Fatalf("encodeBID128 biased %d: expected an exponent-range rejection, got: %v", bidBiasedExpMax128+1, err)
	}
}

// TestEncodeAcceptsMaxBiasedExponent checks the biased-exponent range from
// over-tightening: the finite maximum itself (biased 191 / 767 / 12287) is
// IN range and must still encode. Together with the one-past test above this
// pins the boundary from both sides.
func TestEncodeAcceptsMaxBiasedExponent(t *testing.T) {
	one := big.NewInt(1)
	if _, err := encodeBID32(one, bidBiasedExpMax32-bidBias32); err != nil {
		t.Fatalf("encodeBID32 at biased max %d: unexpected error: %v", bidBiasedExpMax32, err)
	}
	if _, err := encodeBID64(one, bidBiasedExpMax64-bidBias64); err != nil {
		t.Fatalf("encodeBID64 at biased max %d: unexpected error: %v", bidBiasedExpMax64, err)
	}
	if _, _, err := encodeBID128(one, bidBiasedExpMax128-bidBias128); err != nil {
		t.Fatalf("encodeBID128 at biased max %d: unexpected error: %v", bidBiasedExpMax128, err)
	}
}

func TestEncodeRejectsNegativeCoefficient(t *testing.T) {
	neg := big.NewInt(-1)
	if _, err := encodeBID32(neg, 0); err == nil || !strings.Contains(err.Error(), "negative coefficient") {
		t.Fatalf("encodeBID32: expected a negative-coefficient rejection, got: %v", err)
	}
	if _, err := encodeBID64(neg, 0); err == nil || !strings.Contains(err.Error(), "negative coefficient") {
		t.Fatalf("encodeBID64: expected a negative-coefficient rejection, got: %v", err)
	}
	if _, _, err := encodeBID128(neg, 0); err == nil || !strings.Contains(err.Error(), "negative coefficient") {
		t.Fatalf("encodeBID128: expected a negative-coefficient rejection, got: %v", err)
	}
}

func TestWidthSpecForOwner(t *testing.T) {
	if w, ok := widthSpecForOwner("Decimal64"); !ok || w.selfType != "Decimal64" {
		t.Fatalf("widthSpecForOwner(%q) = %+v, %v", "Decimal64", w, ok)
	}
	if _, ok := widthSpecForOwner("Decimal65"); ok {
		t.Fatalf("widthSpecForOwner(%q) unexpectedly ok", "Decimal65")
	}
}
