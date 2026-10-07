package bid754

import (
	"strconv"
	"strings"
	"testing"

	"github.com/sky1core/bid754/bid754-go/internal/bidgo"
)

func TestPublicParserDirectedTinyExponent(t *testing.T) {
	modes := []RoundingMode{RoundNearestEven, RoundNearestAway, RoundTowardZero, RoundTowardPositive, RoundTowardNegative}
	for _, width := range []struct {
		name string
		min  int
	}{
		{"d32", -101},
		{"d64", -398},
	} {
		for _, sign := range []string{"", "-"} {
			for _, exponent := range []int{width.min, width.min - 1, width.min - 2} {
				for _, mode := range modes {
					t.Run(width.name+"/"+sign+"1e"+strconv.Itoa(exponent)+"/"+mode.String(), func(t *testing.T) {
						wantCoeff := uint64(0)
						wantFlags := FlagUnderflow | FlagInexact
						if exponent == width.min {
							wantCoeff, wantFlags = 1, 0
						} else if (sign == "" && mode == RoundTowardPositive) || (sign == "-" && mode == RoundTowardNegative) {
							wantCoeff = 1
						}
						input := sign + "1e" + strconv.Itoa(exponent)
						if width.name == "d32" {
							got, flags, err := NewDecimal32WithMode(input, mode)
							want := uint32(wantCoeff)
							if sign == "-" {
								want |= 0x80000000
							}
							if err != nil || got.ToUint32() != want || flags != wantFlags {
								t.Errorf("input=%q mode=%s: got=%08x flags=%v err=%v; want=%08x flags=%v", input, mode, got.ToUint32(), flags, err, want, wantFlags)
							}
						} else {
							got, flags, err := NewDecimal64WithMode(input, mode)
							want := wantCoeff
							if sign == "-" {
								want |= 0x8000000000000000
							}
							if err != nil || got.ToUint64() != want || flags != wantFlags {
								t.Errorf("input=%q mode=%s: got=%016x flags=%v err=%v; want=%016x flags=%v", input, mode, got.ToUint64(), flags, err, want, wantFlags)
							}
						}
					})
				}
			}
		}
	}
}

func TestPublicParserD32SingleRoundingAtUnderflow(t *testing.T) {
	modes := []RoundingMode{RoundNearestEven, RoundNearestAway, RoundTowardZero, RoundTowardPositive, RoundTowardNegative}
	for _, digits := range []string{"14999998", "14999999", "15000000", "15000001"} {
		for _, sign := range []string{"", "-"} {
			for _, mode := range modes {
				wantCoeff := uint32(1)
				if digits >= "15000000" && (mode == RoundNearestAway || mode == RoundNearestEven) {
					wantCoeff = 2
				}
				if mode == RoundTowardPositive && sign == "" || mode == RoundTowardNegative && sign == "-" {
					wantCoeff = 2
				}
				if sign == "-" {
					wantCoeff |= 0x80000000
				}
				for _, input := range []string{sign + "0." + strings.Repeat("0", 100) + digits, sign + digits + "e-108"} {
					t.Run(input+"/"+mode.String(), func(t *testing.T) {
						got, flags, err := NewDecimal32WithMode(input, mode)
						if err != nil || got.ToUint32() != wantCoeff || flags != FlagUnderflow|FlagInexact {
							t.Errorf("got=%08x flags=%v err=%v; want=%08x flags=%v", got.ToUint32(), flags, err, wantCoeff, FlagUnderflow|FlagInexact)
						}
					})
				}
			}
		}
	}
}

func TestPublicParserD64LongFractionStaysTiny(t *testing.T) {
	modes := []RoundingMode{RoundNearestEven, RoundNearestAway, RoundTowardZero, RoundTowardPositive, RoundTowardNegative}
	for _, zeros := range []int{893, 894, 895} {
		for _, sign := range []string{"", "-"} {
			for _, mode := range modes {
				input := sign + "0." + strings.Repeat("0", zeros) + "12345678901234567"
				got, flags, err := NewDecimal64WithMode(input, mode)
				want := uint64(0)
				if mode == RoundTowardPositive && sign == "" || mode == RoundTowardNegative && sign == "-" {
					want = 1
				}
				if sign == "-" {
					want |= 0x8000000000000000
				}
				if err != nil || got.ToUint64() != want || flags != FlagUnderflow|FlagInexact {
					t.Errorf("zeros=%d sign=%q mode=%s: got=%016x flags=%v err=%v; want=%016x flags=%v", zeros, sign, mode, got.ToUint64(), flags, err, want, FlagUnderflow|FlagInexact)
				}
			}
		}
	}
}

func TestPublicParserZeroCohortIgnoresRawStatus(t *testing.T) {
	modes := []RoundingMode{RoundNearestEven, RoundNearestAway, RoundTowardZero, RoundTowardPositive, RoundTowardNegative}
	for _, mode := range modes {
		for _, sign := range []string{"", "-"} {
			for _, exponent := range []int{-6176, -6177, -6210, -6211, -6212} {
				input := sign + "0e" + strconv.Itoa(exponent)
				got, flags, err := NewDecimal128WithMode(input, mode)
				if exponent < -6176 {
					if err == nil || got != (Decimal128BID{}) || flags != 0 {
						t.Errorf("input=%q mode=%s: got=%x flags=%v err=%v; want rejection", input, mode, got.ToBytes(), flags, err)
					}
				} else if err != nil || flags != 0 || decimal128BIDWordsIsNotZero(got, sign == "-") {
					t.Errorf("input=%q mode=%s: got=%x flags=%v err=%v; want signed exact zero", input, mode, got.ToBytes(), flags, err)
				}
			}
		}
	}
	for _, input := range []string{"0e-6211", "-0e-6211"} {
		got, flags := bidgo.Bid128FromString(input, bidgo.BID_ROUNDING_UP)
		hi, lo := bidgo.Bid128Words(got)
		wantHi := uint64(0)
		if strings.HasPrefix(input, "-") {
			wantHi = 0x8000000000000000
		}
		if flags != 0 || hi != wantHi || lo != 0 {
			t.Errorf("port %q: got=%016x:%016x flags=%08x; want signed exact zero", input, hi, lo, flags)
		}
		public, publicFlags := ParseDecimal128BIDRaw(input)
		if public != canonicalQNaN128BID() || publicFlags != FlagInvalidOperation {
			t.Errorf("public raw %q: got=%x flags=%v; want invalid NaN", input, public.ToBytes(), publicFlags)
		}
	}
	for _, input := range []string{"0e-398", "-0e-398", "0e-399", "-0e-399"} {
		got, flags, err := NewDecimal32WithMode(input, RoundTowardPositive)
		if err == nil || got != (Decimal32BID{}) || flags != 0 {
			t.Errorf("D32 input=%q: got=%08x flags=%v err=%v; want rejection", input, got.ToUint32(), flags, err)
		}
	}
}

func decimal128BIDWordsIsNotZero(value Decimal128BID, negative bool) bool {
	hi, lo := decimal128BIDWords(value)
	if negative {
		return hi != 0x8000000000000000 || lo != 0
	}
	return hi != 0 || lo != 0
}
