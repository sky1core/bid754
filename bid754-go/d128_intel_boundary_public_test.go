package bid754

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/sky1core/bid754/bid754-go/internal/decimalref"
	"github.com/sky1core/bid754/bid754-go/internal/tier1ref"
)

func d128BoundaryInput(t *testing.T, coeff *big.Int, exp int, negative bool) (string, Decimal128BID) {
	t.Helper()
	raw, err := decimalref.Encode(128, decimalref.Decimal{Kind: "finite", Negative: negative, Coeff: coeff, Exp: exp})
	if err != nil {
		t.Fatal(err)
	}
	values, err := finiteParse128([]string{raw})
	if err != nil {
		t.Fatal(err)
	}
	return raw, values[0].Dec
}

func TestD128ScaleBThresholdBoundaryPublic(t *testing.T) {
	threshold := new(big.Int).Exp(big.NewInt(10), big.NewInt(33), nil)
	limb := new(big.Int).Lsh(new(big.Int).Rsh(new(big.Int).Set(threshold), 64), 64)
	coeffs := []*big.Int{
		new(big.Int).Sub(threshold, big.NewInt(1)),
		new(big.Int).Set(threshold),
		new(big.Int).Add(threshold, big.NewInt(1)),
		new(big.Int).Sub(limb, big.NewInt(1)),
		new(big.Int).Set(limb),
		new(big.Int).Add(limb, big.NewInt(1)),
	}
	for _, coeff := range coeffs {
		for _, expAndShift := range [][2]int{{6110, 1}, {6110, 2}, {6111, 1}} {
			for _, negative := range []bool{false, true} {
				raw, input := d128BoundaryInput(t, coeff, expAndShift[0], negative)
				for _, modeName := range []string{"nearest_even", "nearest_away", "toward_zero", "toward_positive", "toward_negative"} {
					mode, _ := finitePublicMode(modeName)
					c := tier1ref.Case{Width: 128, Op: "scaleb", Mode: modeName, Operands: []string{raw}, Param: fmt.Sprint(expAndShift[1])}
					want, err := tier1ref.Evaluate(c)
					if err != nil {
						t.Fatal(err)
					}
					got, flags := input.ScaleBWithMode(expAndShift[1], mode)
					mapped, err := finiteMapPublicFlags(flags)
					if err != nil {
						t.Fatal(err)
					}
					obs := tier1ref.Observation{Path: "go/public", Kind: "decimal", Width: 128, Value: finiteDec128Hex(got), Flags: mapped, HasFlags: true}
					if err := tier1ref.Compare(c, want, obs); err != nil {
						t.Errorf("coeff=%s exp=%d shift=%d negative=%t mode=%s got=%s flags=%#x: %v", coeff, expAndShift[0], expAndShift[1], negative, modeName, obs.Value, obs.Flags, err)
					}
					if modeName == "nearest_even" {
						defaultGot, defaultFlags := input.ScaleB(expAndShift[1])
						defaultMapped, err := finiteMapPublicFlags(defaultFlags)
						if err != nil {
							t.Fatal(err)
						}
						obs.Path, obs.Value, obs.Flags = "go/public/default", finiteDec128Hex(defaultGot), defaultMapped
						if err := tier1ref.Compare(c, want, obs); err != nil {
							t.Errorf("default ScaleB coeff=%s exp=%d shift=%d negative=%t: %v", coeff, expAndShift[0], expAndShift[1], negative, err)
						}
					}
				}
			}
		}
	}
}

func TestD128FMASubtractionOverflowBoundaryPublic(t *testing.T) {
	one := big.NewInt(1)
	for _, xCohort := range [][2]int64{{1, 35}, {10, 34}} {
		for _, exp := range []int{6110, 6111} {
			for _, subtract := range []int64{4, 5, 6, 7, 8, 9, 10, 11} {
				for _, negative := range []bool{false, true} {
					xRaw, x := d128BoundaryInput(t, big.NewInt(xCohort[0]), int(xCohort[1]), false)
					yRaw, y := d128BoundaryInput(t, one, exp, negative)
					zRaw, z := d128BoundaryInput(t, big.NewInt(subtract), exp, !negative)
					for _, modeName := range []string{"nearest_even", "nearest_away", "toward_zero", "toward_positive", "toward_negative"} {
						mode, _ := finitePublicMode(modeName)
						c := decimalref.Case{Width: 128, Op: "fma", Mode: modeName, Operands: []string{xRaw, yRaw, zRaw}}
						want, err := decimalref.Evaluate(c)
						if err != nil {
							t.Fatal(err)
						}
						got, flags := x.FMAWithMode(y, z, mode)
						mapped, err := finiteMapPublicFlags(flags)
						if err != nil {
							t.Fatal(err)
						}
						if err := decimalref.Compare(128, want, finiteDec128Hex(got), mapped); err != nil {
							t.Errorf("x=%dE%d exp=%d subtract=%d negative=%t mode=%s got=%s flags=%#x: %v", xCohort[0], xCohort[1], exp, subtract, negative, modeName, finiteDec128Hex(got), mapped, err)
						}
						if modeName == "nearest_even" {
							defaultGot, defaultFlags := x.FMA(y, z)
							defaultMapped, err := finiteMapPublicFlags(defaultFlags)
							if err != nil {
								t.Fatal(err)
							}
							if err := decimalref.Compare(128, want, finiteDec128Hex(defaultGot), defaultMapped); err != nil {
								t.Errorf("default FMA x=%dE%d exp=%d subtract=%d negative=%t: %v", xCohort[0], xCohort[1], exp, subtract, negative, err)
							}
						}
					}
				}
			}
		}
	}
}
