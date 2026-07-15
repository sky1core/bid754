package bidgo

import "testing"

func TestBid32ScaleFamilyReturnsIntelStatusFlagsDirectly(t *testing.T) {
	tests := []struct {
		name      string
		x         uint32
		n         int
		want      uint32
		wantFlags uint32
	}{
		{
			name:      "exact",
			x:         0x00000001,
			n:         1,
			want:      0x00800001,
			wantFlags: 0,
		},
		{
			name:      "underflow_inexact",
			x:         0x00000001,
			n:         -1,
			want:      0x00000000,
			wantFlags: BID_UNDERFLOW_EXCEPTION | BID_INEXACT_EXCEPTION,
		},
		{
			name:      "overflow_inexact",
			x:         0x77f8967f,
			n:         1,
			want:      0x78000000,
			wantFlags: BID_OVERFLOW_EXCEPTION | BID_INEXACT_EXCEPTION,
		},
		{
			name:      "signaling_nan",
			x:         0x7e000000,
			n:         0,
			want:      0x7c000000,
			wantFlags: BID_INVALID_EXCEPTION,
		},
	}

	operations := []struct {
		name string
		call func(uint32, int) (uint32, uint32)
	}{
		{name: "scalbn", call: func(x uint32, n int) (uint32, uint32) {
			return Bid32Scalbn(x, n, BID_ROUNDING_TO_NEAREST)
		}},
		{name: "scalbln", call: func(x uint32, n int) (uint32, uint32) {
			return Bid32Scalbln(x, int64(n), BID_ROUNDING_TO_NEAREST)
		}},
		{name: "ldexp", call: func(x uint32, n int) (uint32, uint32) {
			return Bid32Ldexp(x, n, BID_ROUNDING_TO_NEAREST)
		}},
	}

	for _, op := range operations {
		op := op
		t.Run(op.name, func(t *testing.T) {
			for _, tc := range tests {
				tc := tc
				t.Run(tc.name, func(t *testing.T) {
					got, flags := op.call(tc.x, tc.n)
					if got != tc.want || flags != tc.wantFlags {
						t.Fatalf("result/status = %08x/%02x, want %08x/%02x", got, flags, tc.want, tc.wantFlags)
					}
				})
			}
		})
	}
}
