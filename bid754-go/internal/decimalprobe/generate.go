package decimalprobe

import (
	"fmt"
	"math/big"
	"math/bits"

	"github.com/sky1core/bid754/bid754-go/internal/decimalref"
)

type Sample struct {
	Version int             `json:"version"`
	Family  string          `json:"family"`
	Case    decimalref.Case `json:"case"`
}

type family struct {
	name, operation, rounding string
}

var families = []family{
	{"add_below", "add", "below"}, {"add_tie", "add", "tie"}, {"add_above", "add", "above"},
	{"carry_boundary", "add", "tie"}, {"cancellation", "sub", "exact"},
	{"mul_below", "mul", "below"}, {"mul_tie", "mul", "tie"}, {"mul_above", "mul", "above"},
	{"div_below", "div", "below"}, {"div_tie", "div", "tie"}, {"div_above", "div", "above"},
	{"quantize_below", "quantize", "below"}, {"quantize_tie", "quantize", "tie"}, {"quantize_above", "quantize", "above"},
	{"underflow_below", "mul", "below"}, {"underflow_tie", "mul", "tie"}, {"underflow_above", "mul", "above"},
	{"overflow", "mul", ""}, {"fma_cancellation", "fma", "exact"}, {"exact", "add", "exact"},
}

func Families() []string {
	out := make([]string, len(families))
	for i, f := range families {
		out[i] = f.name
	}
	return out
}

func findFamily(name string) (family, error) {
	for _, f := range families {
		if f.name == name {
			return f, nil
		}
	}
	return family{}, fmt.Errorf("unknown relation family %q", name)
}

func FamilyOperation(name string) (string, error) {
	f, err := findFamily(name)
	return f.operation, err
}

func pow10(n int) *big.Int {
	return new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(n)), nil)
}

func entropy(hi, lo uint64) *big.Int {
	n := new(big.Int).Lsh(new(big.Int).SetUint64(hi), 64)
	return n.Or(n, new(big.Int).SetUint64(lo))
}

func position(raw int32, lo, hi int) int {
	span := int64(hi - lo + 1)
	return lo + int((int64(raw)%span+span)%span)
}

func decimal(coeff *big.Int, exponent int, negative bool) decimalref.Decimal {
	return decimalref.Decimal{Kind: "finite", Negative: negative, Coeff: new(big.Int).Set(coeff), Exp: exponent}
}

func makeSample(name, op string, width int, mode string, operands ...decimalref.Decimal) (Sample, error) {
	s := Sample{Version: 1, Family: name, Case: decimalref.Case{Width: width, Op: op, Mode: mode}}
	for _, operand := range operands {
		raw, err := decimalref.Encode(width, operand)
		if err != nil {
			return Sample{}, err
		}
		s.Case.Operands = append(s.Case.Operands, raw)
	}
	if _, err := Validate(s); err != nil {
		return Sample{}, err
	}
	return s, nil
}

func Generate(name string, width int, mode string, hi, lo uint64, exp int32, negative bool) (Sample, error) {
	f, err := findFamily(name)
	if err != nil {
		return Sample{}, err
	}
	p, err := decimalref.ParametersFor(width)
	if err != nil {
		return Sample{}, err
	}
	base := pow10(p.Precision - 1)
	limit := pow10(p.Precision)
	maximum := new(big.Int).Sub(limit, big.NewInt(1))
	random := entropy(hi, lo)
	full := new(big.Int).Add(base, new(big.Int).Mod(random, new(big.Int).Sub(limit, base)))
	q := position(exp, p.MinExp, p.MaxExp)
	var operands []decimalref.Decimal
	switch name {
	case "add_below", "add_tie", "add_above", "carry_boundary":
		q = position(exp, p.MinExp+p.Precision, p.MaxExp)
		residue := new(big.Int).Mul(big.NewInt(5), base)
		if f.rounding == "below" {
			residue.Sub(residue, big.NewInt(1))
		} else if f.rounding == "above" {
			residue.Add(residue, big.NewInt(1))
		}
		if name == "carry_boundary" {
			full = maximum
		}
		operands = []decimalref.Decimal{decimal(full, q, negative), decimal(residue, q-p.Precision, negative)}
	case "cancellation":
		operands = []decimalref.Decimal{decimal(full, q, negative), decimal(new(big.Int).Sub(full, big.NewInt(1)), q, negative)}
	case "mul_below", "mul_tie", "mul_above":
		q = position(exp, p.MinExp, p.MaxExp-1)
		start := new(big.Int).Mul(base, big.NewInt(2))
		span := new(big.Int).Quo(new(big.Int).Sub(limit, start), big.NewInt(10))
		coefficient := new(big.Int).Mul(new(big.Int).Mod(random, span), big.NewInt(10))
		coefficient.Add(coefficient, start)
		tail := int64(5)
		if f.rounding == "below" {
			tail = 7
		} else if f.rounding == "above" {
			tail = 3
		}
		coefficient.Add(coefficient, big.NewInt(tail))
		operands = []decimalref.Decimal{decimal(coefficient, q, negative), decimal(big.NewInt(9), 0, false)}
	case "div_below", "div_tie", "div_above":
		denominator, remainder := int64(4), int64(1)
		if f.rounding == "tie" {
			denominator = 2
		} else if f.rounding == "above" {
			remainder = 3
		}
		start := new(big.Int).Mul(base, big.NewInt(denominator))
		span := new(big.Int).Quo(new(big.Int).Sub(limit, start), big.NewInt(denominator))
		coefficient := new(big.Int).Mul(new(big.Int).Mod(random, span), big.NewInt(denominator))
		coefficient.Add(coefficient, start).Add(coefficient, big.NewInt(remainder))
		operands = []decimalref.Decimal{decimal(coefficient, q, negative), decimal(big.NewInt(denominator), 0, false)}
	case "quantize_below", "quantize_tie", "quantize_above":
		q = position(exp, p.MinExp+1, p.MaxExp)
		coefficient := new(big.Int).Quo(full, big.NewInt(10))
		coefficient.Mul(coefficient, big.NewInt(10))
		tail := int64(5)
		if f.rounding == "below" {
			tail = 4
		} else if f.rounding == "above" {
			tail = 6
		}
		coefficient.Add(coefficient, big.NewInt(tail))
		operands = []decimalref.Decimal{decimal(coefficient, q-1, negative), decimal(big.NewInt(1), q, false)}
	case "underflow_below", "underflow_tie", "underflow_above":
		coefficient := new(big.Int).Mul(big.NewInt(5), base)
		if f.rounding == "below" {
			coefficient.Sub(coefficient, big.NewInt(1))
		} else if f.rounding == "above" {
			coefficient.Add(coefficient, big.NewInt(1))
		}
		operands = []decimalref.Decimal{decimal(big.NewInt(1), p.MinExp, negative), decimal(coefficient, -p.Precision, false)}
	case "overflow":
		operands = []decimalref.Decimal{decimal(maximum, p.MaxExp, negative), decimal(big.NewInt(10), 0, false)}
	case "fma_cancellation":
		q = position(exp, p.MinExp+2*(p.Precision-1), p.MaxExp)
		span := new(big.Int).Sub(pow10((p.Precision-2)/2), big.NewInt(1))
		delta := new(big.Int).Add(new(big.Int).Mod(random, span), big.NewInt(1))
		x := new(big.Int).Add(base, delta)
		z := new(big.Int).Add(base, new(big.Int).Lsh(delta, 1))
		operands = []decimalref.Decimal{
			decimal(x, -(p.Precision - 1), negative),
			decimal(x, q-(p.Precision-1), false),
			decimal(z, q-(p.Precision-1), !negative),
		}
	case "exact":
		x := new(big.Int).Mod(random, base)
		operands = []decimalref.Decimal{decimal(x, q, negative), decimal(big.NewInt(1), q, negative)}
	default:
		return Sample{}, fmt.Errorf("family %q has no generator", name)
	}
	return makeSample(name, f.operation, width, mode, operands...)
}

func Uniform(op string, width int, mode string, hi, lo uint64, exp int32, negative bool) (Sample, error) {
	p, err := decimalref.ParametersFor(width)
	if err != nil {
		return Sample{}, err
	}
	arity := 2
	switch op {
	case "add", "sub", "mul", "div", "quantize":
	case "fma":
		arity = 3
	default:
		return Sample{}, fmt.Errorf("unsupported finite operation %q", op)
	}
	operands := make([]decimalref.Decimal, arity)
	for i := range operands {
		coefficient := entropy(bits.RotateLeft64(hi, i*21), bits.RotateLeft64(lo, -i*17))
		coefficient.Mod(coefficient, pow10(p.Precision))
		q := position(int32(uint32(exp)+uint32(lo>>uint(i*16))), p.MinExp, p.MaxExp)
		sign := negative
		if i != 0 {
			sign = hi&(1<<uint(i-1)) != 0
		}
		operands[i] = decimal(coefficient, q, sign)
	}
	return makeSample("uniform-finite", op, width, mode, operands...)
}

func Validate(s Sample) (decimalref.Result, error) {
	if s.Version != 1 {
		return decimalref.Result{}, fmt.Errorf("unsupported sample version %d", s.Version)
	}
	p, err := decimalref.ParametersFor(s.Case.Width)
	if err != nil {
		return decimalref.Result{}, err
	}
	for _, raw := range s.Case.Operands {
		d, err := decimalref.Decode(s.Case.Width, raw)
		if err != nil {
			return decimalref.Result{}, err
		}
		if d.Kind != "finite" {
			return decimalref.Result{}, fmt.Errorf("finite campaign received %s operand", d.Kind)
		}
		encoded, err := decimalref.Encode(s.Case.Width, d)
		if err != nil || encoded != raw {
			return decimalref.Result{}, fmt.Errorf("finite campaign requires canonical raw operand %q", raw)
		}
	}
	r, err := decimalref.Evaluate(s.Case)
	if err != nil {
		return decimalref.Result{}, err
	}
	if s.Family == "uniform-finite" || s.Family == "boundary-finite" {
		return r, nil
	}
	f, err := findFamily(s.Family)
	if err != nil {
		return decimalref.Result{}, err
	}
	if s.Case.Op != f.operation || (f.rounding != "" && r.Rounding != f.rounding) {
		return decimalref.Result{}, fmt.Errorf("relation %s not satisfied: operation=%s rounding=%s", s.Family, s.Case.Op, r.Rounding)
	}
	switch s.Family {
	case "carry_boundary":
		x, _ := decimalref.Decode(s.Case.Width, s.Case.Operands[0])
		y, _ := decimalref.Decode(s.Case.Width, s.Case.Operands[1])
		if x.Coeff.Cmp(new(big.Int).Sub(pow10(p.Precision), big.NewInt(1))) != 0 || x.Negative != y.Negative || y.Coeff.Sign() == 0 {
			return decimalref.Result{}, fmt.Errorf("carry boundary requires full nines coefficient")
		}
		truncatedCase := s.Case
		truncatedCase.Mode = "toward_zero"
		truncated, err := decimalref.Evaluate(truncatedCase)
		if err != nil {
			return decimalref.Result{}, err
		}
		if truncated.Value.Kind != "finite" || truncated.Value.Coeff.Cmp(x.Coeff) != 0 || truncated.Value.Exp != x.Exp {
			return decimalref.Result{}, fmt.Errorf("carry boundary requires a midpoint above full nines at the operand quantum")
		}
	case "cancellation":
		if r.Cancellation < p.Precision-2 || r.Flags != 0 {
			return decimalref.Result{}, fmt.Errorf("deep exact cancellation relation not satisfied")
		}
	case "underflow_below", "underflow_tie", "underflow_above":
		if r.Flags&0x30 != 0x30 {
			return decimalref.Result{}, fmt.Errorf("tiny inexact relation not satisfied")
		}
	case "overflow":
		if r.Flags&0x28 != 0x28 {
			return decimalref.Result{}, fmt.Errorf("overflow relation not satisfied")
		}
	case "fma_cancellation":
		if r.Flags != 0 || r.Value.Kind != "finite" || r.Value.Coeff.Sign() == 0 || r.Cancellation < p.Precision {
			return decimalref.Result{}, fmt.Errorf("fused cancellation relation not satisfied")
		}
		productCase := s.Case
		productCase.Op, productCase.Operands = "mul", s.Case.Operands[:2]
		product, err := decimalref.Evaluate(productCase)
		if err != nil {
			return decimalref.Result{}, err
		}
		productBits, err := decimalref.Encode(s.Case.Width, product.Value)
		if err != nil {
			return decimalref.Result{}, err
		}
		unfusedCase := s.Case
		unfusedCase.Op, unfusedCase.Operands = "add", []string{productBits, s.Case.Operands[2]}
		unfused, err := decimalref.Evaluate(unfusedCase)
		if err != nil {
			return decimalref.Result{}, err
		}
		unfusedBits, err := decimalref.Encode(s.Case.Width, unfused.Value)
		if err != nil {
			return decimalref.Result{}, err
		}
		if decimalref.Compare(s.Case.Width, r, unfusedBits, product.Flags|unfused.Flags) == nil {
			return decimalref.Result{}, fmt.Errorf("fused cancellation does not distinguish intermediate rounding")
		}
	case "exact":
		if r.Flags != 0 {
			return decimalref.Result{}, fmt.Errorf("exact relation raised flags %#x", r.Flags)
		}
	}
	return r, nil
}
