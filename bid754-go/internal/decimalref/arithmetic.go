package decimalref

import (
	"fmt"
	"math/big"
)

func exactValue(d Decimal) *big.Rat {
	n := new(big.Int).Set(d.Coeff)
	if d.Negative {
		n.Neg(n)
	}
	if d.Exp >= 0 {
		return new(big.Rat).SetInt(n.Mul(n, power10(d.Exp)))
	}
	return new(big.Rat).SetFrac(n, power10(-d.Exp))
}

func adjusted(x *big.Rat) int {
	n := new(big.Int).Abs(x.Num())
	d := x.Denom()
	e := len(n.String()) - len(d.String())
	if e >= 0 {
		if n.Cmp(new(big.Int).Mul(d, power10(e))) < 0 {
			e--
		}
	} else if new(big.Int).Mul(n, power10(-e)).Cmp(d) < 0 {
		e--
	}
	return e
}

func roundAt(x *big.Rat, exp int, negative bool, mode string) (*big.Int, string, bool) {
	n := new(big.Int).Abs(x.Num())
	d := new(big.Int).Set(x.Denom())
	if exp >= 0 {
		d.Mul(d, power10(exp))
	} else {
		n.Mul(n, power10(-exp))
	}
	q, rem := new(big.Int), new(big.Int)
	q.QuoRem(n, d, rem)
	if rem.Sign() == 0 {
		return q, "exact", false
	}
	cmp := new(big.Int).Lsh(rem, 1).Cmp(d)
	rounding := "tie"
	if cmp < 0 {
		rounding = "below"
	} else if cmp > 0 {
		rounding = "above"
	}
	increment := false
	switch mode {
	case "nearest_even":
		increment = cmp > 0 || (cmp == 0 && q.Bit(0) == 1)
	case "nearest_away":
		increment = cmp >= 0
	case "toward_positive":
		increment = !negative
	case "toward_negative":
		increment = negative
	}
	if increment {
		q.Add(q, big.NewInt(1))
	}
	return q, rounding, increment
}

func invalidResult() Result {
	return Result{Value: Decimal{Kind: "nan"}, Flags: flagInvalid, Rounding: "exact"}
}

func roundResult(p Parameters, x *big.Rat, preferred int, negative bool, mode string, divide, quantize bool) Result {
	r := Result{Rounding: "exact"}
	if x.Sign() == 0 {
		r.Value = Decimal{Kind: "finite", Negative: negative, Coeff: new(big.Int), Exp: max(p.MinExp, min(preferred, p.MaxExp))}
		return r
	}
	negative = x.Sign() < 0
	order := adjusted(x)
	exp := max(order-p.Precision+1, p.MinExp)
	if quantize {
		exp = preferred
	} else if !divide {
		exp = max(exp, min(preferred, p.MaxExp))
	}
	coeff, rounding, increment := roundAt(x, exp, negative, mode)
	r.Rounding = rounding
	r.Carry = increment && coeff.Cmp(power10(p.Precision)) == 0
	if quantize && coeff.Cmp(power10(p.Precision)) >= 0 {
		r.Value = Decimal{Kind: "nan"}
		r.Flags = flagInvalid
		return r
	}
	if rounding != "exact" {
		r.Flags |= flagInexact
		if !quantize && order < p.MinExp+p.Precision-1 {
			r.Flags |= flagUnderflow
		}
	}
	if r.Carry && !quantize {
		coeff.Quo(coeff, big.NewInt(10))
		exp++
	}
	if exp > p.MaxExp {
		r.Flags = flagOverflow | flagInexact
		toFinite := mode == "toward_zero" || (mode == "toward_positive" && negative) || (mode == "toward_negative" && !negative)
		if toFinite {
			r.Value = Decimal{Kind: "finite", Negative: negative, Coeff: new(big.Int).Sub(power10(p.Precision), big.NewInt(1)), Exp: p.MaxExp}
		} else {
			r.Value = Decimal{Kind: "infinity", Negative: negative}
		}
		return r
	}
	if divide && rounding == "exact" {
		limit := min(preferred, p.MaxExp)
		for exp < limit {
			q, rem := new(big.Int), new(big.Int)
			q.QuoRem(coeff, big.NewInt(10), rem)
			if rem.Sign() != 0 {
				break
			}
			coeff = q
			exp++
		}
	}
	r.Value = Decimal{Kind: "finite", Negative: negative, Coeff: coeff, Exp: exp}
	return r
}

func cancellation(a, b, sum *big.Rat, preferred int) int {
	if a.Sign() == 0 || b.Sign() == 0 || a.Sign() == b.Sign() {
		return 0
	}
	largest := max(adjusted(a), adjusted(b))
	if sum.Sign() == 0 {
		return max(0, largest-preferred+1)
	}
	return max(0, largest-adjusted(sum))
}

func Evaluate(c Case) (Result, error) {
	p, err := ParametersFor(c.Width)
	if err != nil {
		return Result{}, err
	}
	switch c.Mode {
	case "nearest_even", "nearest_away", "toward_zero", "toward_positive", "toward_negative":
	default:
		return Result{}, fmt.Errorf("unsupported rounding mode %q", c.Mode)
	}
	count := 2
	switch c.Op {
	case "add", "sub", "mul", "div", "quantize":
	case "fma":
		count = 3
	default:
		return Result{}, fmt.Errorf("unsupported operation %q", c.Op)
	}
	if len(c.Operands) != count {
		return Result{}, fmt.Errorf("%s needs %d operands", c.Op, count)
	}
	operands := make([]Decimal, count)
	for i, raw := range c.Operands {
		d, err := Decode(c.Width, raw)
		if err != nil {
			return Result{}, fmt.Errorf("operand %d: %w", i, err)
		}
		if d.Kind != "finite" {
			return Result{}, fmt.Errorf("operand %d: nonfinite input unsupported", i)
		}
		operands[i] = d
	}
	a, b := operands[0], operands[1]
	x, y := exactValue(a), exactValue(b)
	preferred, negative := a.Exp, a.Negative
	lost := 0
	switch c.Op {
	case "add", "sub", "fma":
		if c.Op == "sub" {
			y.Neg(y)
			b.Negative = !b.Negative
		}
		if c.Op == "fma" {
			x.Mul(x, y)
			a.Exp += b.Exp
			a.Negative = a.Negative != b.Negative
			b = operands[2]
			y = exactValue(b)
		}
		preferred = min(a.Exp, b.Exp)
		sum := new(big.Rat).Add(x, y)
		lost = cancellation(x, y, sum, preferred)
		negative = (a.Negative && b.Negative) || (a.Negative != b.Negative && c.Mode == "toward_negative")
		x = sum
	case "mul":
		x.Mul(x, y)
		preferred = a.Exp + b.Exp
		negative = a.Negative != b.Negative
	case "div":
		negative = a.Negative != b.Negative
		if y.Sign() == 0 {
			if x.Sign() == 0 {
				return invalidResult(), nil
			}
			return Result{Value: Decimal{Kind: "infinity", Negative: negative}, Flags: flagDivisionByZero, Rounding: "exact"}, nil
		}
		x.Quo(x, y)
		preferred = a.Exp - b.Exp
	case "quantize":
		preferred = b.Exp
	}
	r := roundResult(p, x, preferred, negative, c.Mode, c.Op == "div", c.Op == "quantize")
	r.Cancellation = lost
	return r, nil
}
