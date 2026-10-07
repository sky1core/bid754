package bid754

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"sync"
	"testing"

	"github.com/sky1core/bid754/bid754-go/internal/decimalref"
	"github.com/sky1core/bid754/bid754-go/internal/tier1ref"
)

var tier1ScaleBBoundaryCases = sync.OnceValues(func() ([]tier1ref.Case, error) {
	var out []tier1ref.Case
	for _, width := range []int{32, 64, 128} {
		p, err := decimalref.ParametersFor(width)
		if err != nil {
			return nil, err
		}
		threshold := tier1BoundaryPower(p.Precision - 1)
		coeffs := []*big.Int{big.NewInt(0), big.NewInt(1), big.NewInt(5), new(big.Int).Sub(tier1BoundaryPower(p.Precision), big.NewInt(1))}
		for delta := int64(-2); delta <= 2; delta++ {
			coeffs = append(coeffs, new(big.Int).Add(threshold, big.NewInt(delta)))
		}
		if width == 128 {
			limb := new(big.Int).Rsh(new(big.Int).Set(threshold), 64)
			limb.Lsh(limb, 64)
			for delta := int64(-1); delta <= 1; delta++ {
				coeffs = append(coeffs, new(big.Int).Add(limb, big.NewInt(delta)))
			}
		}
		for _, coeff := range coeffs {
			for _, edge := range []struct {
				exponents []int
				shifts    []int
			}{
				{[]int{p.MaxExp - 1, p.MaxExp}, []int{0, 1, 2, p.Precision, p.Precision + 1}},
				{[]int{p.MinExp, p.MinExp + 1}, []int{-2, -1, 0, 1}},
			} {
				for _, exp := range edge.exponents {
					for _, shift := range edge.shifts {
						for _, negative := range []bool{false, true} {
							raw, err := tier1Finite(width, coeff, exp, negative)
							if err != nil {
								return nil, err
							}
							for _, mode := range finiteModes {
								out = append(out, tier1ref.Case{Width: width, Op: "scaleb", Mode: mode, Operands: []string{raw}, Param: strconv.Itoa(shift)})
							}
						}
					}
				}
			}
		}
	}
	return out, nil
})

func TestTier1ScaleBBoundaryGo(t *testing.T) {
	cases, err := tier1ScaleBBoundaryCases()
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range cases {
		want, err := tier1ref.Evaluate(c)
		if err != nil {
			t.Fatal(err)
		}
		obs, err := tier1PublicPaths(c)
		if err != nil {
			t.Fatal(err)
		}
		for _, o := range obs {
			if err := tier1ref.Compare(c, want, o); err != nil {
				t.Fatalf("case=%+v path=%s: %v", c, o.Path, err)
			}
		}
	}
	t.Logf("TIER1-SCALEB-BOUNDARY cases=%d", len(cases))
}

func TestTier1ScaleBBoundaryBigDecimal(t *testing.T) {
	configured, err := tier1Configured()
	if err != nil {
		t.Fatal(err)
	}
	if !configured {
		t.Skip("requires Tier1 Java and generated Rust comparison processes")
	}
	s, err := tier1Start()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := s.close(); err != nil {
			t.Error(err)
		}
	})
	cases, err := tier1ScaleBBoundaryCases()
	if err != nil {
		t.Fatal(err)
	}
	checked := 0
	checkedCells := map[string]int{}
	witnesses := map[string]bool{}
	for _, c := range cases {
		java, _ := tier1RunCase(t, s, c)
		if java.Status == "ok" {
			checked++
			checkedCells[fmt.Sprintf("%d/%s", c.Width, c.Mode)]++
			if c.Width == 128 && c.Operands[0] == "5ffe314dc6448d93:38c15b09ffffffff" && c.Param == "1" {
				witnesses[c.Mode] = true
			}
		}
	}
	if len(checkedCells) != 15 || len(witnesses) != 5 {
		t.Fatalf("Java checked cells=%d normalization modes=%d", len(checkedCells), len(witnesses))
	}
	t.Logf("TIER1-SCALEB-BIGDECIMAL cases=%d checked=%d", len(cases), checked)
}

func TestTier1ScaleBBoundaryFuzzSeeds(t *testing.T) {
	cases, err := tier1ScaleBBoundaryCases()
	if err != nil {
		t.Fatal(err)
	}
	counts := map[string]int{}
	witness := false
	for i, c := range cases {
		data := make([]byte, tier1InputSize)
		data[4] = 0x40
		binary.LittleEndian.PutUint64(data[5:13], uint64(i))
		for mode, name := range finiteModes {
			data[2] = byte(mode)
			got, err := tier1Sample(data)
			want := c
			want.Mode = name
			if err != nil || !reflect.DeepEqual(got, want) {
				t.Fatalf("seed=%d mode=%s: got=%+v want=%+v err=%v", i, name, got, want, err)
			}
		}
		counts[fmt.Sprintf("%d/%s", c.Width, c.Mode)]++
		if c.Width == 128 && c.Operands[0] == "5ffe314dc6448d93:38c15b09ffffffff" && c.Param == "1" {
			witness = true
		}
	}
	if len(cases) != 5400 || len(counts) != 15 || !witness {
		t.Fatalf("cases=%d cells=%d normalization witness=%t", len(cases), len(counts), witness)
	}
}
