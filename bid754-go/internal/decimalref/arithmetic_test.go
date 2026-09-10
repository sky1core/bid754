package decimalref

import (
	"fmt"
	"math/big"
	"testing"
)

func finite(coeff string, exp int) Decimal {
	n, ok := new(big.Int).SetString(coeff, 10)
	if !ok {
		panic("invalid test coefficient")
	}
	negative := n.Sign() < 0
	return Decimal{Kind: "finite", Negative: negative, Coeff: n.Abs(n), Exp: exp}
}

func eval(t *testing.T, width int, op, mode string, operands ...Decimal) Result {
	t.Helper()
	c := Case{Width: width, Op: op, Mode: mode}
	for _, d := range operands {
		raw, err := Encode(width, d)
		if err != nil {
			t.Fatal(err)
		}
		c.Operands = append(c.Operands, raw)
	}
	r, err := Evaluate(c)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func check(t *testing.T, r Result, want Decimal, flags uint32, rounding string) {
	t.Helper()
	if r.Value.Kind != want.Kind || r.Value.Negative != want.Negative || r.Flags != flags || r.Rounding != rounding {
		t.Fatalf("got %+v; want value=%+v flags=%#x rounding=%s", r, want, flags, rounding)
	}
	if want.Kind == "finite" && (r.Value.Exp != want.Exp || r.Value.Coeff.Cmp(want.Coeff) != 0) {
		t.Fatalf("got %sE%d; want %sE%d", r.Value.Coeff, r.Value.Exp, want.Coeff, want.Exp)
	}
}

func TestExactArithmetic(t *testing.T) {
	for _, width := range []int{32, 64, 128} {
		t.Run(fmt.Sprint(width), func(t *testing.T) {
			check(t, eval(t, width, "add", "nearest_even", finite("12", 0), finite("700", -2)), finite("1900", -2), 0, "exact")
			check(t, eval(t, width, "sub", "nearest_even", finite("13", -1), finite("107", -2)), finite("23", -2), 0, "exact")
			check(t, eval(t, width, "mul", "nearest_even", finite("123", -2), finite("20", -1)), finite("2460", -3), 0, "exact")
			check(t, eval(t, width, "div", "nearest_even", finite("1", 0), finite("-8", 0)), finite("-125", -3), 0, "exact")
			check(t, eval(t, width, "div", "nearest_even", finite("100", -2), finite("2", 0)), finite("50", -2), 0, "exact")
			check(t, eval(t, width, "quantize", "nearest_even", finite("217", -2), finite("1", -1)), finite("22", -1), flagInexact, "above")
		})
	}
}

func TestRoundingModesAndDivision(t *testing.T) {
	for _, width := range []int{32, 64, 128} {
		p, _ := ParametersFor(width)
		base := power10(p.Precision - 1)
		for _, negative := range []bool{false, true} {
			for _, mode := range []string{"nearest_even", "nearest_away", "toward_zero", "toward_positive", "toward_negative"} {
				t.Run(fmt.Sprintf("%d/%t/%s", width, negative, mode), func(t *testing.T) {
					a := Decimal{Kind: "finite", Negative: negative, Coeff: base, Exp: 0}
					b := Decimal{Kind: "finite", Negative: negative, Coeff: big.NewInt(5), Exp: -1}
					want := new(big.Int).Set(base)
					if mode == "nearest_away" || (mode == "toward_positive" && !negative) || (mode == "toward_negative" && negative) {
						want.Add(want, big.NewInt(1))
					}
					check(t, eval(t, width, "add", mode, a, b), Decimal{Kind: "finite", Negative: negative, Coeff: want, Exp: 0}, flagInexact, "tie")
					third := new(big.Int).Quo(new(big.Int).Sub(power10(p.Precision), big.NewInt(1)), big.NewInt(3))
					if (mode == "toward_positive" && !negative) || (mode == "toward_negative" && negative) {
						third.Add(third, big.NewInt(1))
					}
					one := finite("1", 0)
					one.Negative = negative
					check(t, eval(t, width, "div", mode, one, finite("3", 0)), Decimal{Kind: "finite", Negative: negative, Coeff: third, Exp: -p.Precision}, flagInexact, "below")
				})
			}
		}
	}
}

func TestBoundaries(t *testing.T) {
	for _, width := range []int{32, 64, 128} {
		t.Run(fmt.Sprint(width), func(t *testing.T) {
			p, _ := ParametersFor(width)
			maxCoeff := new(big.Int).Sub(power10(p.Precision), big.NewInt(1))
			largest := Decimal{Kind: "finite", Coeff: maxCoeff, Exp: p.MaxExp}
			for _, negative := range []bool{false, true} {
				largest.Negative = negative
				for _, mode := range []string{"nearest_even", "nearest_away", "toward_zero", "toward_positive", "toward_negative"} {
					want := Decimal{Kind: "infinity", Negative: negative}
					if mode == "toward_zero" || (mode == "toward_positive" && negative) || (mode == "toward_negative" && !negative) {
						want = largest
					}
					check(t, eval(t, width, "mul", mode, largest, finite("10", 0)), want, flagOverflow|flagInexact, "exact")
				}
			}
			check(t, eval(t, width, "div", "nearest_even", finite("1", p.MinExp), finite("2", 0)), finite("0", p.MinExp), flagUnderflow|flagInexact, "tie")
			check(t, eval(t, width, "div", "toward_positive", finite("1", p.MinExp), finite("2", 0)), finite("1", p.MinExp), flagUnderflow|flagInexact, "tie")
			check(t, eval(t, width, "mul", "nearest_even", finite("1", p.MinExp), finite("1", 0)), finite("1", p.MinExp), 0, "exact")
			nearNormal := new(big.Int).Sub(power10(p.Precision), big.NewInt(5))
			check(t, eval(t, width, "div", "nearest_even", Decimal{Kind: "finite", Coeff: nearNormal, Exp: p.MinExp}, finite("10", 0)), Decimal{Kind: "finite", Coeff: power10(p.Precision - 1), Exp: p.MinExp}, flagUnderflow|flagInexact, "tie")
			check(t, eval(t, width, "quantize", "nearest_even", finite("15", p.MinExp), finite("1", p.MinExp+1)), finite("2", p.MinExp+1), flagInexact, "tie")
			check(t, eval(t, width, "quantize", "nearest_even", finite("1", p.MaxExp), finite("1", p.MinExp)), Decimal{Kind: "nan"}, flagInvalid, "exact")
			carry := eval(t, width, "add", "nearest_even", Decimal{Kind: "finite", Coeff: maxCoeff, Exp: 0}, finite("5", -1))
			check(t, carry, Decimal{Kind: "finite", Coeff: power10(p.Precision - 1), Exp: 1}, flagInexact, "tie")
			if !carry.Carry {
				t.Fatal("missing rounding carry")
			}
			odd := Decimal{Kind: "finite", Coeff: new(big.Int).Add(power10(p.Precision-1), big.NewInt(1)), Exp: 0}
			check(t, eval(t, width, "add", "nearest_even", odd, finite("5", -1)), Decimal{Kind: "finite", Coeff: new(big.Int).Add(power10(p.Precision-1), big.NewInt(2)), Exp: 0}, flagInexact, "tie")
			negativeTiny := finite("-1", p.MinExp)
			negativeZero := finite("0", p.MinExp)
			negativeZero.Negative = true
			check(t, eval(t, width, "div", "toward_positive", negativeTiny, finite("3", 0)), negativeZero, flagUnderflow|flagInexact, "below")
			check(t, eval(t, width, "div", "toward_negative", negativeTiny, finite("3", 0)), negativeTiny, flagUnderflow|flagInexact, "below")
		})
	}
}

func TestSignedZeroAndInvalidDivision(t *testing.T) {
	for _, width := range []int{32, 64, 128} {
		for _, mode := range []string{"nearest_even", "nearest_away", "toward_zero", "toward_positive", "toward_negative"} {
			for _, an := range []bool{false, true} {
				for _, bn := range []bool{false, true} {
					a, b := finite("0", -2), finite("0", 0)
					a.Negative, b.Negative = an, bn
					want := finite("0", -2)
					want.Negative = an && bn || an != bn && mode == "toward_negative"
					check(t, eval(t, width, "add", mode, a, b), want, 0, "exact")
					want.Negative = an != bn
					check(t, eval(t, width, "mul", mode, a, b), want, 0, "exact")
					check(t, eval(t, width, "div", mode, a, b), Decimal{Kind: "nan"}, flagInvalid, "exact")
					a.Coeff.SetInt64(1)
					check(t, eval(t, width, "div", mode, a, b), Decimal{Kind: "infinity", Negative: an != bn}, flagDivisionByZero, "exact")
				}
			}
			zero := finite("0", -1)
			zero.Negative = mode == "toward_negative"
			r := eval(t, width, "sub", mode, finite("13", -1), finite("13", -1))
			check(t, r, zero, 0, "exact")
			if r.Cancellation != 2 {
				t.Fatalf("zero cancellation: %d", r.Cancellation)
			}
		}
	}
}

func TestFMAOneRounding(t *testing.T) {
	for _, width := range []int{32, 64, 128} {
		p, _ := ParametersFor(width)
		e := p.Precision - 1
		a := Decimal{Kind: "finite", Coeff: new(big.Int).Add(power10(e), big.NewInt(1)), Exp: -e}
		b := Decimal{Kind: "finite", Coeff: new(big.Int).Sub(power10(e), big.NewInt(1)), Exp: -e}
		r := eval(t, width, "fma", "nearest_even", a, b, finite("-1", 0))
		check(t, r, finite("-1", -2*e), 0, "exact")
		if r.Cancellation != 2*e {
			t.Fatalf("cancellation got %d want %d", r.Cancellation, 2*e)
		}
		product := eval(t, width, "mul", "nearest_even", a, b)
		separate := eval(t, width, "add", "nearest_even", product.Value, finite("-1", 0))
		if separate.Value.Coeff.Sign() != 0 {
			t.Fatal("test must distinguish separate rounding")
		}
		large := finite("1", p.MaxExp)
		tiny := finite("1", p.MinExp)
		r = eval(t, width, "fma", "nearest_even", large, large, tiny)
		if r.Value.Kind != "infinity" || r.Flags != flagOverflow|flagInexact {
			t.Fatalf("extreme fma: %+v", r)
		}
		cancelLarge := Decimal{Kind: "finite", Negative: true, Coeff: new(big.Int).Sub(power10(p.Precision), big.NewInt(1)), Exp: p.MaxExp}
		check(t, eval(t, width, "fma", "nearest_even", large, finite("1", p.Precision), cancelLarge), large, 0, "exact")
	}
}
