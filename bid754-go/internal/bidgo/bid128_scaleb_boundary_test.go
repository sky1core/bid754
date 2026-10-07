package bidgo

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"testing"

	"github.com/sky1core/bid754/bid754-go/internal/decimalref"
	"github.com/sky1core/bid754/bid754-go/internal/tier1ref"
)

func TestBid128ScalbnLdexpThresholdBoundary(t *testing.T) {
	threshold := new(big.Int).Exp(big.NewInt(10), big.NewInt(33), nil)
	limb := new(big.Int).Lsh(new(big.Int).Rsh(new(big.Int).Set(threshold), 64), 64)
	modes := []struct {
		name string
		port int
	}{
		{"nearest_even", BID_ROUNDING_TO_NEAREST},
		{"nearest_away", BID_ROUNDING_TIES_AWAY},
		{"toward_zero", BID_ROUNDING_TO_ZERO},
		{"toward_positive", BID_ROUNDING_UP},
		{"toward_negative", BID_ROUNDING_DOWN},
	}
	for _, coeff := range []*big.Int{
		new(big.Int).Sub(threshold, big.NewInt(1)), threshold, new(big.Int).Add(threshold, big.NewInt(1)),
		new(big.Int).Sub(limb, big.NewInt(1)), limb, new(big.Int).Add(limb, big.NewInt(1)),
	} {
		for _, negative := range []bool{false, true} {
			raw, err := decimalref.Encode(128, decimalref.Decimal{Kind: "finite", Negative: negative, Coeff: coeff, Exp: 6111})
			if err != nil {
				t.Fatal(err)
			}
			parts := strings.Split(raw, ":")
			hi, err := strconv.ParseUint(parts[0], 16, 64)
			if err != nil {
				t.Fatal(err)
			}
			lo, err := strconv.ParseUint(parts[1], 16, 64)
			if err != nil {
				t.Fatal(err)
			}
			x := Bid128FromWords(hi, lo)
			for _, mode := range modes {
				c := tier1ref.Case{Width: 128, Op: "scaleb", Mode: mode.name, Operands: []string{raw}, Param: "1"}
				want, err := tier1ref.Evaluate(c)
				if err != nil {
					t.Fatal(err)
				}
				var flags uint32
				scalbn := Bid128Scalbn(x, 1, mode.port, &flags)
				ldexp, ldexpFlags := Bid128Ldexp(x, 1, mode.port)
				for _, result := range []struct {
					name  string
					value BID_UINT128
					flags uint32
				}{{"scalbn", scalbn, flags}, {"ldexp", ldexp, ldexpFlags}} {
					hi, lo := Bid128Words(result.value)
					obs := tier1ref.Observation{Path: result.name, Kind: "decimal", Width: 128, Value: fmt.Sprintf("%016x:%016x", hi, lo), Flags: result.flags, HasFlags: true}
					if err := tier1ref.Compare(c, want, obs); err != nil {
						t.Errorf("%s coeff=%s negative=%t mode=%s got=%s flags=%#x: %v", result.name, coeff, negative, mode.name, obs.Value, obs.Flags, err)
					}
				}
			}
		}
	}
}
