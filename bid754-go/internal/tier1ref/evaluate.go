package tier1ref

import (
	"math/big"
	"strconv"
	"strings"

	"github.com/sky1core/bid754/bid754-go/internal/decimalref"
)

func pow10(n int) *big.Int { return new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(n)), nil) }

func decimalResult(width int, d decimalref.Decimal, flags uint32) Result {
	return Result{Kind: "decimal", Width: width, Decimal: d, Flags: flags, Quantum: true}
}

func nanResult(width int, flags uint32) Result {
	return decimalResult(width, decimalref.Decimal{Kind: "nan"}, flags)
}

func signaling(width int, raw string) bool {
	bits, _ := new(big.Int).SetString(strings.ReplaceAll(raw, ":", ""), 16)
	return bits.Bit(width-7) != 0
}

func increment(q, rem, divisor *big.Int, negative bool, mode string) bool {
	if rem.Sign() == 0 {
		return false
	}
	switch mode {
	case "nearest_even", "nearest_away":
		cmp := new(big.Int).Lsh(rem, 1).Cmp(divisor)
		return cmp > 0 || cmp == 0 && (mode == "nearest_away" || q.Bit(0) != 0)
	case "toward_positive":
		return !negative
	case "toward_negative":
		return negative
	default:
		return false
	}
}

func overflowResult(width int, negative bool, mode string) Result {
	p, _ := decimalref.ParametersFor(width)
	d := decimalref.Decimal{Kind: "infinity", Negative: negative}
	if mode == "toward_zero" || mode == "toward_positive" && negative || mode == "toward_negative" && !negative {
		d = decimalref.Decimal{Kind: "finite", Negative: negative, Coeff: new(big.Int).Sub(pow10(p.Precision), big.NewInt(1)), Exp: p.MaxExp}
	}
	return decimalResult(width, d, Overflow|Inexact)
}

func roundedDecimal(width int, coeff *big.Int, exponent *big.Int, negative bool, mode string) Result {
	p, _ := decimalref.ParametersFor(width)
	n := new(big.Int).Set(coeff)
	if n.Sign() == 0 {
		e := p.MinExp
		if exponent.Cmp(big.NewInt(int64(p.MaxExp))) > 0 {
			e = p.MaxExp
		} else if exponent.Cmp(big.NewInt(int64(p.MinExp))) >= 0 {
			e = int(exponent.Int64())
		}
		return decimalResult(width, decimalref.Decimal{Kind: "finite", Negative: negative, Coeff: n, Exp: e}, 0)
	}
	digits := len(n.String())
	if exponent.Cmp(big.NewInt(int64(p.MaxExp+p.Precision))) > 0 {
		return overflowResult(width, negative, mode)
	}
	if exponent.Cmp(big.NewInt(int64(p.MinExp-digits))) < 0 {
		if mode == "toward_positive" && !negative || mode == "toward_negative" && negative {
			n.SetInt64(1)
		} else {
			n.SetInt64(0)
		}
		return decimalResult(width, decimalref.Decimal{Kind: "finite", Negative: negative, Coeff: n, Exp: p.MinExp}, Underflow|Inexact)
	}
	e := int(exponent.Int64())
	tiny := e+digits-1 < p.MinExp+p.Precision-1
	shift := max(0, digits-p.Precision, p.MinExp-e)
	flags := uint32(0)
	if shift > 0 {
		divisor := pow10(shift)
		q, rem := new(big.Int), new(big.Int)
		q.QuoRem(n, divisor, rem)
		if rem.Sign() != 0 {
			flags = Inexact
			if tiny {
				flags |= Underflow
			}
		}
		if increment(q, rem, divisor, negative, mode) {
			q.Add(q, big.NewInt(1))
		}
		n, e = q, e+shift
	}
	if n.Cmp(pow10(p.Precision)) >= 0 {
		n.Quo(n, big.NewInt(10))
		e++
	}
	if e > p.MaxExp {
		available := p.Precision - len(n.String())
		if e-p.MaxExp > available {
			return overflowResult(width, negative, mode)
		}
		n.Mul(n, pow10(e-p.MaxExp))
		e = p.MaxExp
	}
	return decimalResult(width, decimalref.Decimal{Kind: "finite", Negative: negative, Coeff: n, Exp: e}, flags)
}

func signedCoefficient(d decimalref.Decimal, exponent int) *big.Int {
	n := new(big.Int).Mul(d.Coeff, pow10(d.Exp-exponent))
	if d.Negative {
		n.Neg(n)
	}
	return n
}

func compareNumbers(a, b decimalref.Decimal) int {
	if a.Kind == "infinity" {
		if b.Kind == "infinity" && a.Negative == b.Negative {
			return 0
		}
		if a.Negative {
			return -1
		}
		return 1
	}
	if b.Kind == "infinity" {
		if b.Negative {
			return 1
		}
		return -1
	}
	e := min(a.Exp, b.Exp)
	return signedCoefficient(a, e).Cmp(signedCoefficient(b, e))
}

func predicate(op string, order int, unordered bool) bool {
	switch op {
	case "quiet_equal":
		return !unordered && order == 0
	case "quiet_not_equal":
		return unordered || order != 0
	case "quiet_greater":
		return !unordered && order > 0
	case "quiet_greater_equal":
		return !unordered && order >= 0
	case "quiet_greater_unordered":
		return unordered || order > 0
	case "quiet_less":
		return !unordered && order < 0
	case "quiet_less_equal":
		return !unordered && order <= 0
	case "quiet_less_unordered":
		return unordered || order < 0
	case "quiet_not_greater":
		return unordered || order <= 0
	case "quiet_not_less":
		return unordered || order >= 0
	case "quiet_ordered":
		return !unordered
	default:
		return unordered
	}
}

func toInteger(c Case, d decimalref.Decimal) Result {
	unsigned := strings.HasPrefix(c.Op, "to_uint")
	r := Result{Kind: "integer", Width: c.Target}
	invalid := func() Result {
		n := new(big.Int).Lsh(big.NewInt(1), uint(c.Target-1))
		if !unsigned {
			n.Neg(n)
		}
		r.Value, r.Flags = n.String(), Invalid
		return r
	}
	if d.Kind != "finite" {
		return invalid()
	}
	n := new(big.Int).Set(d.Coeff)
	inexact := false
	if d.Exp >= 0 {
		n.Mul(n, pow10(d.Exp))
	} else {
		divisor := pow10(-d.Exp)
		q, rem := new(big.Int), new(big.Int)
		q.QuoRem(n, divisor, rem)
		inexact = rem.Sign() != 0
		if increment(q, rem, divisor, d.Negative, c.Mode) {
			q.Add(q, big.NewInt(1))
		}
		n = q
	}
	if d.Negative {
		n.Neg(n)
	}
	lo, hi := integerBounds(c.Target, unsigned)
	if n.Cmp(lo) < 0 || n.Cmp(hi) > 0 {
		return invalid()
	}
	if inexact && strings.HasSuffix(c.Op, "_exact") {
		r.Flags = Inexact
	}
	r.Value = n.String()
	return r
}

func Evaluate(c Case) (Result, error) {
	if err := Validate(c); err != nil {
		return Result{}, err
	}
	width := c.Width
	if c.Op == "convert" {
		width = c.Target
	}
	if c.Op == "from_int" || c.Op == "from_uint" {
		n, _ := canonicalInteger(c.Param, 20)
		negative := n.Sign() < 0
		return roundedDecimal(width, n.Abs(n), new(big.Int), negative, c.Mode), nil
	}
	operands := make([]decimalref.Decimal, len(c.Operands))
	flags, hasNaN := uint32(0), false
	for i, raw := range c.Operands {
		operands[i], _ = decimalref.Decode(c.Width, raw)
		if operands[i].Kind == "nan" {
			hasNaN = true
			if signaling(c.Width, raw) {
				flags = Invalid
			}
		}
	}
	a := operands[0]
	if strings.HasPrefix(c.Op, "to_") {
		return toInteger(c, a), nil
	}
	if strings.HasPrefix(c.Op, "quiet_") {
		order := 0
		if !hasNaN {
			order = compareNumbers(a, operands[1])
		}
		return Result{Kind: "bool", Value: strconv.FormatBool(predicate(c.Op, order, hasNaN)), Flags: flags}, nil
	}
	if c.Op == "minnum" || c.Op == "maxnum" {
		b := operands[1]
		if hasNaN {
			if flags != 0 || a.Kind == "nan" && b.Kind == "nan" {
				return nanResult(width, flags), nil
			}
			if a.Kind == "nan" {
				return decimalResult(width, b, 0), nil
			}
			return decimalResult(width, a, 0), nil
		}
		order := compareNumbers(a, b)
		selected := a
		if c.Op == "minnum" && order > 0 || c.Op == "maxnum" && order < 0 {
			selected = b
		}
		r := decimalResult(width, selected, 0)
		if order == 0 {
			r.Quantum = a.Exp == b.Exp
			r.AnyZeroSign = a.Kind == "finite" && a.Coeff.Sign() == 0 && a.Negative != b.Negative
		}
		return r, nil
	}
	if hasNaN {
		return nanResult(width, flags), nil
	}
	switch c.Op {
	case "convert", "scaleb":
		if a.Kind == "infinity" {
			return decimalResult(width, a, 0), nil
		}
		exponent := big.NewInt(int64(a.Exp))
		if c.Op == "scaleb" {
			shift, _ := canonicalInteger(c.Param, 19)
			exponent.Add(exponent, shift)
		}
		return roundedDecimal(width, a.Coeff, exponent, a.Negative, c.Mode), nil
	default:
		b := operands[1]
		if a.Kind == "infinity" || b.Kind == "finite" && b.Coeff.Sign() == 0 {
			return nanResult(width, Invalid), nil
		}
		if b.Kind == "infinity" {
			return decimalResult(width, a, 0), nil
		}
		e := min(a.Exp, b.Exp)
		x, y := signedCoefficient(a, e), signedCoefficient(b, e)
		q, rem := new(big.Int), new(big.Int)
		q.QuoRem(x, y, rem)
		if c.Op == "rem" {
			absQ, absRem, absY := new(big.Int).Abs(q), new(big.Int).Abs(rem), new(big.Int).Abs(y)
			if increment(absQ, absRem, absY, false, "nearest_even") {
				if x.Sign()*y.Sign() < 0 {
					q.Sub(q, big.NewInt(1))
				} else {
					q.Add(q, big.NewInt(1))
				}
				rem.Sub(x, new(big.Int).Mul(q, y))
			}
		}
		negative := rem.Sign() < 0 || rem.Sign() == 0 && a.Negative
		return decimalResult(width, decimalref.Decimal{Kind: "finite", Negative: negative, Coeff: rem.Abs(rem), Exp: e}, 0), nil
	}
}
