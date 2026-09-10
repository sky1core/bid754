package bidcodec_test

import (
	"fmt"
	"math/big"
	"strings"
	"testing"

	codec "github.com/sky1core/bid754/bid754-codec-go"
)

func TestParserResourceContract(t *testing.T) {
	cases := checkResourceRejects(t) + checkResourceLeadingZeros(t)
	for _, width := range []struct {
		bits, precision, minExp, maxExp int
	}{
		{32, 7, -101, 90},
		{64, 16, -398, 369},
		{128, 34, -6176, 6111},
	} {
		for _, digits := range []int{width.precision - 1, width.precision, width.precision + 1} {
			for _, exponent := range []int{width.minExp - 1, width.minExp, width.minExp + 1, width.maxExp - 1, width.maxExp, width.maxExp + 1} {
				for _, sign := range []string{"+", "-"} {
					input := fmt.Sprintf("%s%sE%+d", sign, strings.Repeat("9", digits), exponent)
					c, err := codec.FromString(input)
					cases++
					if digits > 34 {
						if err == nil {
							t.Fatalf("accepted schema overflow: %s", input)
						}
						continue
					}
					if err != nil {
						t.Fatalf("shared schema rejected %s: %v", input, err)
					}
					if c.Coefficient.String() != strings.Repeat("9", digits) || c.Exponent != int32(exponent) || c.Sign != (sign == "-") || c.Kind != codec.Normal {
						t.Fatalf("changed parsed components: %s => %+v", input, c)
					}
					wantOK := digits <= width.precision && exponent >= width.minExp && exponent <= width.maxExp
					checkParserEncode(t, width.bits, c, wantOK)
				}
			}
		}
		for _, exponent := range []int{width.minExp - 1, width.minExp, width.maxExp, width.maxExp + 1} {
			c, err := codec.FromString(fmt.Sprintf("-0E%d", exponent))
			if err != nil || c.Kind != codec.Zero || !c.Sign || c.Exponent != int32(exponent) {
				t.Fatalf("signed zero exponent=%d: %+v %v", exponent, c, err)
			}
			checkParserEncode(t, width.bits, c, exponent >= width.minExp && exponent <= width.maxExp)
			cases++
		}
		for _, digits := range []int{width.precision - 1, width.precision} {
			for _, prefix := range []string{"NaN", "-sNaN"} {
				c, err := codec.FromString(prefix + strings.Repeat("9", digits))
				if digits > 33 {
					if err == nil {
						t.Fatal("accepted payload schema overflow")
					}
				} else {
					if err != nil {
						t.Fatal(err)
					}
					checkParserEncode(t, width.bits, c, digits < width.precision)
				}
				cases++
			}
		}
	}
	for _, tc := range []struct {
		input string
		exp   int32
		ok    bool
	}{
		{"1E2147483647", 2147483647, true},
		{"1E-2147483648", -2147483648, true},
		{"1E2147483648", 0, false},
		{"1E-2147483649", 0, false},
		{"0.1E2147483648", 2147483647, true},
		{"0.1E-2147483648", 0, false},
		{"1E9007199254740991", 0, false},
		{"1E9007199254740992", 0, false},
		{"1E-9007199254740991", 0, false},
		{"1E-9007199254740992", 0, false},
	} {
		c, err := codec.FromString(tc.input)
		if (err == nil) != tc.ok || (tc.ok && (c.Exponent != tc.exp || c.Coefficient.Cmp(big.NewInt(1)) != 0)) {
			t.Fatalf("%s: %+v %v", tc.input, c, err)
		}
		cases++
	}
	fmt.Printf("PARSER-RESOURCE language=go cases=%d\n", cases)
}

func checkParserEncode(t *testing.T, width int, c codec.Components, wantOK bool) {
	t.Helper()
	var decoded codec.Components
	var err error
	switch width {
	case 32:
		var bits uint32
		bits, err = codec.Encode32(c)
		decoded = codec.Decode32(bits)
	case 64:
		var bits uint64
		bits, err = codec.Encode64(c)
		decoded = codec.Decode64(bits)
	case 128:
		var lo, hi uint64
		lo, hi, err = codec.Encode128(c)
		decoded = codec.Decode128(lo, hi)
	default:
		t.Fatalf("unsupported width: %d", width)
	}
	if (err == nil) != wantOK {
		t.Fatalf("Encode%d %+v: want success=%v, got %v", width, c, wantOK, err)
	}
	if wantOK {
		got, gotErr := codec.ToString(decoded)
		want, wantErr := codec.ToString(c)
		if gotErr != nil || wantErr != nil || got != want {
			t.Fatalf("Encode%d changed components: got %s (%v), want %s (%v)", width, got, gotErr, want, wantErr)
		}
	}
}
