package bidgo

import "testing"

func TestBid32NextAfterNormalBoundary(t *testing.T) {
	for _, sign := range []uint32{0, 0x80000000} {
		for _, coeff := range []uint32{999998, 999999, 1000000, 1000001, 1000002} {
			for _, outward := range []bool{false, true} {
				toward := (sign ^ 0x80000000) | 0x78000000
				wantCoeff := coeff - 1
				if outward {
					toward = sign | 0x78000000
					wantCoeff = coeff + 1
				}
				var wantFlags uint32
				if wantCoeff < 1000000 {
					wantFlags = BID_UNDERFLOW_EXCEPTION | BID_INEXACT_EXCEPTION
				}
				got, flags := Bid32NextAfter(sign|coeff, toward)
				if got != sign|wantCoeff || flags != wantFlags {
					t.Errorf("nextafter(%08x,%08x)=%08x/%x want %08x/%x", sign|coeff, toward, got, flags, sign|wantCoeff, wantFlags)
				}
			}
		}
	}
}

func TestBid32NextAfterEqualCohorts(t *testing.T) {
	for _, sign := range []uint32{0, 0x80000000} {
		for _, coeff := range []uint32{1, 7, 100001, 765432} {
			x := sign | 101<<23 | coeff
			y := sign | 100<<23 | coeff*10
			for _, pair := range [][2]uint32{{x, y}, {y, x}} {
				got, flags := Bid32NextAfter(pair[0], pair[1])
				if got != pair[0] || flags != 0 {
					t.Errorf("nextafter(%08x,%08x)=%08x/%x want %08x/0", pair[0], pair[1], got, flags, pair[0])
				}
			}
		}
	}
	for _, tc := range []struct{ x, y, want uint32 }{
		{0x32800000, 0x80000000, 0xb2800000},
		{0x78001234, 0x78005678, 0x78000000},
		{0xf8001234, 0xf8005678, 0xf8000000},
		{0x6cbfffff, 0x00000000, 0x32800000},
	} {
		got, flags := Bid32NextAfter(tc.x, tc.y)
		if got != tc.want || flags != 0 {
			t.Errorf("nextafter(%08x,%08x)=%08x/%x want %08x/0", tc.x, tc.y, got, flags, tc.want)
		}
	}
}
