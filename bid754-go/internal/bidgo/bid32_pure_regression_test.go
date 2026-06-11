package bidgo

import "testing"

func TestDecimal32PureIsZeroClassifiesCanonicalAndNonCanonicalZero(t *testing.T) {
	cases := []struct {
		name string
		bits uint32
		want bool
	}{
		{name: "positive canonical zero", bits: encodeBID32(0, bid32ExponentBias, 0), want: true},
		{name: "negative canonical zero", bits: encodeBID32(1, 0, 0), want: true},
		{name: "noncanonical finite zero", bits: bid32SpecialEncodingMask | (uint32(bid32ExponentBias) << bid32LargeExpShift) | bid32SmallCoeffMask, want: true},
		{name: "one", bits: encodeBID32(0, bid32ExponentBias, 1), want: false},
		{name: "infinity", bits: bid32InfinityMask, want: false},
		{name: "nan", bits: bid32NaNMask, want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Decimal32Pure(tc.bits).IsZero(); got != tc.want {
				t.Fatalf("Decimal32Pure(%08x).IsZero() = %v, want %v", tc.bits, got, tc.want)
			}
		})
	}
}

func TestFromDecimal64MatchesMechanicalBid64ToBid32(t *testing.T) {
	inputs := []uint64{
		encodeBID64(0, bid64ExponentBias, 0),
		encodeBID64(1, bid64ExponentBias+5, 0),
		encodeBID64(0, bid64ExponentBias, 123456789),
		bid64NaNMask | 123,
		bid64InfinityMask,
	}
	modes := []RoundingMode{
		RoundNearestEven,
		RoundTowardNegative,
		RoundTowardPositive,
		RoundTowardZero,
		RoundNearestAway,
	}
	for _, input := range inputs {
		for _, mode := range modes {
			want, _ := Bid64ToBid32(input, roundingModeToBID(mode))
			got := uint32(fromDecimal64(Decimal64Pure(input), mode))
			if got != want {
				t.Fatalf("fromDecimal64(%016x, %d) = %08x, want %08x", input, mode, got, want)
			}
		}
	}
}
