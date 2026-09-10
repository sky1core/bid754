package decimalprobe

import (
	"fmt"
	"math/big"
	"reflect"

	"github.com/sky1core/bid754/bid754-go/internal/decimalref"
)

type ShrinkStats struct {
	Attempts  int  `json:"attempts"`
	Accepted  int  `json:"accepted"`
	Exhausted bool `json:"exhausted"`
}

type semantics struct {
	Rounding       string
	ResultClass    string
	ResultNegative bool
	Flags          uint32
	Carry          bool
	Signs          []bool
	InputClasses   []string
	TieParity      uint
}

func semanticSignature(s Sample) (semantics, error) {
	r, err := Validate(s)
	if err != nil {
		return semantics{}, err
	}
	class, err := decimalref.Classify(s.Case.Width, r.Value)
	if err != nil {
		return semantics{}, err
	}
	out := semantics{Rounding: r.Rounding, ResultClass: class, ResultNegative: r.Value.Negative, Flags: r.Flags, Carry: r.Carry}
	for _, raw := range s.Case.Operands {
		d, err := decimalref.Decode(s.Case.Width, raw)
		if err != nil {
			return semantics{}, err
		}
		class, err := decimalref.Classify(s.Case.Width, d)
		if err != nil {
			return semantics{}, err
		}
		out.Signs = append(out.Signs, d.Negative)
		out.InputClasses = append(out.InputClasses, class)
	}
	if r.Rounding == "tie" {
		truncated := s.Case
		truncated.Mode = "toward_zero"
		zero, err := decimalref.Evaluate(truncated)
		if err != nil {
			return semantics{}, err
		}
		if zero.Value.Kind == "finite" {
			out.TieParity = zero.Value.Coeff.Bit(0)
		}
	}
	return out, nil
}

type size struct {
	exponents int64
	coeff     *big.Int
}

func sampleSize(s Sample) size {
	out := size{coeff: new(big.Int)}
	for _, raw := range s.Case.Operands {
		d, _ := decimalref.Decode(s.Case.Width, raw)
		q := int64(d.Exp)
		if q < 0 {
			q = -q
		}
		out.exponents += q
		out.coeff.Add(out.coeff, d.Coeff)
	}
	return out
}

func smaller(a, b Sample) bool {
	x, y := sampleSize(a), sampleSize(b)
	if x.exponents != y.exponents {
		return x.exponents < y.exponents
	}
	return x.coeff.Cmp(y.coeff) < 0
}

func candidates(s Sample) []Sample {
	values := make([]decimalref.Decimal, len(s.Case.Operands))
	for i, raw := range s.Case.Operands {
		values[i], _ = decimalref.Decode(s.Case.Width, raw)
	}
	var out []Sample
	seen := make(map[string]bool)
	appendValues := func(vs []decimalref.Decimal) {
		candidate := s
		candidate.Case.Operands = make([]string, len(vs))
		key := ""
		for i, d := range vs {
			raw, err := decimalref.Encode(s.Case.Width, d)
			if err != nil {
				return
			}
			candidate.Case.Operands[i] = raw
			key += raw + "/"
		}
		if !seen[key] && smaller(candidate, s) {
			seen[key] = true
			out = append(out, candidate)
		}
	}
	for _, shift := range []int{values[0].Exp, values[0].Exp / 2} {
		if shift == 0 {
			continue
		}
		vs := append([]decimalref.Decimal(nil), values...)
		switch s.Case.Op {
		case "add", "sub", "quantize":
			for i := range vs {
				vs[i].Exp -= shift
			}
		case "mul", "div":
			vs[0].Exp -= shift
		case "fma":
			vs[0].Exp -= shift
			vs[2].Exp -= shift
		}
		appendValues(vs)
	}
	for i, value := range values {
		for _, exponent := range []int{0, value.Exp / 2} {
			vs := append([]decimalref.Decimal(nil), values...)
			vs[i].Exp = exponent
			appendValues(vs)
		}
		coefficients := []*big.Int{big.NewInt(0), big.NewInt(1), big.NewInt(2), big.NewInt(5), big.NewInt(9),
			new(big.Int).Quo(value.Coeff, big.NewInt(10)), new(big.Int).Quo(value.Coeff, big.NewInt(2))}
		for k := 1; k < len(value.Coeff.String()); k++ {
			for _, delta := range []int64{-1, 0, 1, 2, 5} {
				coefficients = append(coefficients, new(big.Int).Add(pow10(k), big.NewInt(delta)))
			}
		}
		text := value.Coeff.String()
		for j := range text {
			if text[j] == '0' {
				continue
			}
			bytes := []byte(text)
			bytes[j] = '0'
			coefficient, ok := new(big.Int).SetString(string(bytes), 10)
			if ok {
				coefficients = append(coefficients, coefficient)
			}
		}
		for _, coefficient := range coefficients {
			if coefficient.Sign() < 0 || coefficient.Cmp(value.Coeff) >= 0 {
				continue
			}
			vs := append([]decimalref.Decimal(nil), values...)
			vs[i].Coeff = coefficient
			appendValues(vs)
		}
	}
	return out
}

func Shrink(s Sample, maxAttempts int, fails func(Sample) (bool, error)) (Sample, ShrinkStats, error) {
	var stats ShrinkStats
	if maxAttempts <= 0 || fails == nil {
		return Sample{}, stats, fmt.Errorf("shrinking requires a positive attempt budget and a failure predicate")
	}
	signature, err := semanticSignature(s)
	if err != nil {
		return Sample{}, stats, err
	}
	failed, err := fails(s)
	stats.Attempts++
	if err != nil {
		return Sample{}, stats, err
	}
	if !failed {
		return Sample{}, stats, fmt.Errorf("initial sample does not reproduce the failure")
	}
	current := s
	for {
		changed := false
		for _, candidate := range candidates(current) {
			candidateSignature, err := semanticSignature(candidate)
			if err != nil || !reflect.DeepEqual(signature, candidateSignature) {
				continue
			}
			if stats.Attempts >= maxAttempts {
				stats.Exhausted = true
				return current, stats, nil
			}
			failed, err := fails(candidate)
			stats.Attempts++
			if err != nil {
				return Sample{}, stats, err
			}
			if failed {
				current = candidate
				stats.Accepted++
				changed = true
				break
			}
		}
		if !changed {
			return current, stats, nil
		}
	}
}
