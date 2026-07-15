package bidgo

import "testing"

func TestBid32IsZeroClassifiesCanonicalAndNonCanonicalZero(t *testing.T) {
	cases := []struct {
		name string
		bits uint32
		want bool
	}{
		{name: "positive canonical zero", bits: uint32(DECIMAL_EXPONENT_BIAS_32) << 23, want: true},
		{name: "negative canonical zero", bits: MASK_SIGN32, want: true},
		{name: "noncanonical finite zero", bits: SPECIAL_ENCODING_MASK32 | uint32(DECIMAL_EXPONENT_BIAS_32)<<21 | SMALL_COEFF_MASK32, want: true},
		{name: "one", bits: uint32(DECIMAL_EXPONENT_BIAS_32)<<23 | 1, want: false},
		{name: "infinity", bits: INFINITY_MASK32, want: false},
		{name: "nan", bits: NAN_MASK32, want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Bid32IsZero(tc.bits); got != tc.want {
				t.Fatalf("Bid32IsZero(%08x) = %v, want %v", tc.bits, got, tc.want)
			}
		})
	}
}
