package bidgo

import (
	"fmt"
	"testing"
)

func TestRawParserExactZeroAtUnderflow(t *testing.T) {
	for _, width := range []struct{ bits, min, precision int }{{32, -101, 7}, {64, -398, 16}, {128, -6176, 34}} {
		for _, exp := range []int{width.min, width.min - 1, width.min - width.precision, width.min - width.precision - 1, -10000} {
			for _, neg := range []bool{false, true} {
				sign := ""
				if neg {
					sign = "-"
				}
				input := fmt.Sprintf("%s0e%d", sign, exp)
				for mode := 0; mode < 5; mode++ {
					var hi, lo uint64
					var flags uint32
					switch width.bits {
					case 32:
						raw, f := Bid32FromStringRaw(input, mode)
						lo = uint64(raw)
						flags = f
					case 64:
						lo, flags = Bid64FromString(input, mode)
					case 128:
						raw, f := Bid128FromString(input, mode)
						hi, lo = Bid128Words(raw)
						flags = f
					}
					wantLo, wantHi := uint64(0), uint64(0)
					if neg {
						switch width.bits {
						case 32:
							wantLo = 1 << 31
						case 64:
							wantLo = 1 << 63
						case 128:
							wantHi = 1 << 63
						}
					}
					if lo != wantLo || hi != wantHi || flags != 0 {
						t.Errorf("d%d %s mode=%d got=%016x:%016x flags=%x want=%016x:%016x flags=0", width.bits, input, mode, hi, lo, flags, wantHi, wantLo)
					}
				}
			}
		}
	}
}
