package bidgo

import "testing"

func TestBid32MinMaxFamilyReturnsIntelStatusFlagsDirectly(t *testing.T) {
	type operation struct {
		name  string
		value func(uint32, uint32) uint32
		call  func(uint32, uint32) (uint32, uint32)
	}
	operations := []operation{
		{name: "minnum", value: bid32_minnum_pure, call: Bid32MinNumWithFlags},
		{name: "maxnum", value: bid32_maxnum_pure, call: Bid32MaxNumWithFlags},
		{name: "minnum_mag", value: bid32_minnum_mag_pure, call: Bid32MinNumMagWithFlags},
		{name: "maxnum_mag", value: bid32_maxnum_mag_pure, call: Bid32MaxNumMagWithFlags},
	}
	tests := []struct {
		name      string
		x         uint32
		y         uint32
		wantFlags uint32
	}{
		{name: "finite", x: 0x3200000a, y: 0xb200000a, wantFlags: 0},
		{name: "quiet_nan_and_finite", x: 0x7c000123, y: 0x3200000a, wantFlags: 0},
		{name: "signaling_nan_left", x: 0x7e000123, y: 0x3200000a, wantFlags: BID_INVALID_EXCEPTION},
		{name: "signaling_nan_right_after_quiet_nan", x: 0x7c000123, y: 0x7e000456, wantFlags: BID_INVALID_EXCEPTION},
	}

	for _, op := range operations {
		op := op
		t.Run(op.name, func(t *testing.T) {
			for _, tc := range tests {
				tc := tc
				t.Run(tc.name, func(t *testing.T) {
					got, flags := op.call(tc.x, tc.y)
					want := op.value(tc.x, tc.y)
					if got != want || flags != tc.wantFlags {
						t.Fatalf("result/status = %08x/%02x, want %08x/%02x", got, flags, want, tc.wantFlags)
					}
				})
			}
		})
	}
}
