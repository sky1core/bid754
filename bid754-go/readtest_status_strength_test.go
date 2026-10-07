package bid754

import (
	"fmt"
	"testing"

	"github.com/sky1core/bid754/bid754-go/internal/testspec"
)

func TestReadtestScaleBInitialStatus(t *testing.T) {
	checked := 0
	for _, function := range []string{"bid128_scalbn", "bid128_scalbln"} {
		for _, sign := range []uint64{0, 0x8000000000000000} {
			for mode := 0; mode < 5; mode++ {
				for initial := uint32(0); initial <= 0x3d; initial++ {
					if initial & ^uint32(0x3d) != 0 {
						continue
					}
					tc := testspec.GeneratedReadCase{
						Function: function, Format: "decimal128", Kind: "binary_op", Rounding: mode,
						InitialStatus: initial,
						Operands:      []string{fmt.Sprintf("[%016x000000000000000a]", sign), "-1"},
					}
					got, _, flags, err := goportReadCaseOperationBits(tc)
					want := fmt.Sprintf("[%016x0000000000000001]", sign)
					if err != nil || got != want || flags != fmt.Sprintf("%02X", initial) {
						t.Fatalf("%s sign=%x mode=%d initial=%02x: got %s/%s/%v, want %s/%02x", function, sign, mode, initial, got, flags, err, want, initial)
					}
					checked++
				}
			}
		}
	}
	if checked != 640 {
		t.Fatalf("status comparisons=%d, want 640", checked)
	}
	for _, tc := range []testspec.GeneratedReadCase{
		{Function: "bid128_scalbn", Format: "decimal128", InitialStatus: 2, Operands: []string{"[0000000000000000000000000000000a]", "-1"}},
		{Function: "bid128_scalbln", Format: "decimal128", InitialStatus: 0x40, Operands: []string{"[0000000000000000000000000000000a]", "-1"}},
		{Function: "bid128_fma", Format: "decimal128", InitialStatus: 0x20, Operands: []string{"[00000000000000000000000000000000]", "[00000000000000000000000000000000]", "[00000000000000000000000000000000]"}},
	} {
		if _, _, _, err := goportReadCaseOperationBits(tc); err == nil {
			t.Fatalf("accepted unsupported initial status: %+v", tc)
		}
	}
	t.Logf("ScaleB initial-status comparisons=%d", checked)
}
