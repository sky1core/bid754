package bid754

import (
	"math/big"
	"testing"
)

func TestParserNegativeExponentCancellation(t *testing.T) {
	for _, zeros := range []int{1000000, 10485760} {
		for _, leading := range []int{0, 1} {
			positive := coefficientLiteral(false, zeros, leading, -zeros)
			for _, neg := range []bool{false, true} {
				input := positive
				if neg {
					input = "-" + positive
				}
				for _, width := range allSpecs {
					coefficient := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(width.precision-1)), nil)
					quantum := 1 - width.precision
					var hi, lo uint64
					if width.name == "d128" {
						lo = coefficient.Uint64()
						hi = new(big.Int).Rsh(coefficient, 64).Uint64() | uint64(quantum+width.bias)<<49
						if neg {
							hi |= 1 << 63
						}
					} else {
						hi, lo = refEncodeCohort(width, coefficient.Uint64(), neg, quantum)
					}
					for _, mode := range rawRoundingModes {
						gotHi, gotLo, flags := rawParse(width, input, mode.mode)
						if gotHi != hi || gotLo != lo || flags != 0 {
							t.Errorf("%s zeros=%d leading=%d negative=%v mode=%s: got %#x,%#x flags=%#x want exact cohort %#x,%#x flags=0", width.name, zeros, leading, neg, mode.name, gotHi, gotLo, flags, hi, lo)
						}
					}
					if !publicReject(width, input) {
						t.Errorf("%s accepted %d coefficient digits", width.name, zeros+1)
					}
					if _, _, _, err := publicWithFlags(width, input); err == nil {
						t.Errorf("%s WithFlags silently coerced %d coefficient digits", width.name, zeros+1)
					}
					for _, mode := range publicRoundingModes {
						if _, _, _, err := publicWithMode(width, input, mode.mode); err == nil {
							t.Errorf("%s WithMode %s silently coerced %d coefficient digits", width.name, mode.name, zeros+1)
						}
					}
				}
			}
		}
	}
}
