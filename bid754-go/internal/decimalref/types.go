package decimalref

import (
	"fmt"
	"math/big"
)

type Decimal struct {
	Kind     string
	Negative bool
	Coeff    *big.Int
	Exp      int
}

type Case struct {
	Width    int      `json:"width"`
	Op       string   `json:"op"`
	Mode     string   `json:"mode"`
	Operands []string `json:"operands"`
}

type Result struct {
	Value        Decimal
	Flags        uint32
	Rounding     string
	Cancellation int
	Carry        bool
}

type Parameters struct {
	Precision int
	MinExp    int
	MaxExp    int
}

const (
	flagInvalid        uint32 = 0x01
	flagDivisionByZero uint32 = 0x04
	flagOverflow       uint32 = 0x08
	flagUnderflow      uint32 = 0x10
	flagInexact        uint32 = 0x20
)

func ParametersFor(width int) (Parameters, error) {
	switch width {
	case 32:
		return Parameters{7, -101, 90}, nil
	case 64:
		return Parameters{16, -398, 369}, nil
	case 128:
		return Parameters{34, -6176, 6111}, nil
	default:
		return Parameters{}, fmt.Errorf("unsupported decimal width %d", width)
	}
}

func power10(n int) *big.Int {
	return new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(n)), nil)
}

func validate(p Parameters, d Decimal) error {
	switch d.Kind {
	case "finite":
		if d.Exp < p.MinExp || d.Exp > p.MaxExp {
			return fmt.Errorf("exponent %d outside format", d.Exp)
		}
		if d.Coeff == nil || d.Coeff.Sign() < 0 || d.Coeff.Cmp(power10(p.Precision)) >= 0 {
			return fmt.Errorf("invalid finite coefficient")
		}
	case "infinity", "nan":
		if d.Exp != 0 || (d.Coeff != nil && d.Coeff.Sign() != 0) {
			return fmt.Errorf("nonfinite value has finite fields")
		}
	default:
		return fmt.Errorf("unsupported decimal kind %q", d.Kind)
	}
	return nil
}

func Classify(width int, d Decimal) (string, error) {
	p, err := ParametersFor(width)
	if err != nil {
		return "", err
	}
	if err := validate(p, d); err != nil {
		return "", err
	}
	if d.Kind != "finite" {
		return d.Kind, nil
	}
	if d.Coeff.Sign() == 0 {
		if d.Negative {
			return "negative_zero", nil
		}
		return "positive_zero", nil
	}
	if d.Exp+len(d.Coeff.String())-1 < p.MinExp+p.Precision-1 {
		return "subnormal", nil
	}
	return "normal", nil
}
