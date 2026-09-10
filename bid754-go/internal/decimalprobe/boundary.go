package decimalprobe

import (
	"fmt"
	"math/big"

	"github.com/sky1core/bid754/bid754-go/internal/decimalref"
)

func Boundary(op string, width int, mode string, index int) (Sample, error) {
	p, err := decimalref.ParametersFor(width)
	if err != nil {
		return Sample{}, err
	}
	if index < 0 || index >= 20 {
		return Sample{}, fmt.Errorf("boundary index must be in [0,20)")
	}
	negative := index >= 10
	zero, one, ten := big.NewInt(0), big.NewInt(1), big.NewInt(10)
	maximum := new(big.Int).Sub(pow10(p.Precision), one)
	var a, b decimalref.Decimal
	switch index % 10 {
	case 0:
		a, b = decimal(zero, p.MinExp, negative), decimal(zero, p.MaxExp, !negative)
	case 1:
		a, b = decimal(zero, p.MaxExp, negative), decimal(one, 1, !negative)
	case 2:
		a, b = decimal(ten, -1, negative), decimal(one, 0, negative)
	case 3:
		a, b = decimal(maximum, 0, negative), decimal(one, 1, negative)
	case 4:
		a, b = decimal(pow10(p.Precision-1), p.MinExp, negative), decimal(one, -1, false)
	case 5:
		a, b = decimal(maximum, p.MaxExp, negative), decimal(one, 0, false)
	case 6:
		a, b = decimal(one, p.MinExp, negative), decimal(one, -1, false)
	case 7:
		a, b = decimal(maximum, 0, negative), decimal(maximum, 0, !negative)
	case 8, 9:
		a, b = decimal(one, 0, negative), decimal(zero, p.MinExp, index%10 == 9)
	}
	operands := []decimalref.Decimal{a, b}
	switch op {
	case "add", "sub", "mul", "div", "quantize":
	case "fma":
		operands = append(operands, decimal(zero, p.MinExp, negative))
	default:
		return Sample{}, fmt.Errorf("unsupported boundary operation %q", op)
	}
	return makeSample("boundary-finite", op, width, mode, operands...)
}
