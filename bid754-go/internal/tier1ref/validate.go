package tier1ref

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/sky1core/bid754/bid754-go/internal/decimalref"
)

func canonicalInteger(text string, maxDigits int) (*big.Int, error) {
	digits := text
	if strings.HasPrefix(digits, "-") {
		digits = digits[1:]
	}
	if len(digits) == 0 || len(digits) > maxDigits || (len(digits) > 1 && digits[0] == '0') || text == "-0" {
		return nil, fmt.Errorf("invalid canonical integer")
	}
	for _, ch := range digits {
		if ch < '0' || ch > '9' {
			return nil, fmt.Errorf("invalid canonical integer")
		}
	}
	n, ok := new(big.Int).SetString(text, 10)
	if !ok {
		return nil, fmt.Errorf("invalid canonical integer")
	}
	return n, nil
}

func integerBounds(bits int, unsigned bool) (*big.Int, *big.Int) {
	limit := new(big.Int).Lsh(big.NewInt(1), uint(bits))
	if unsigned {
		return new(big.Int), new(big.Int).Sub(limit, big.NewInt(1))
	}
	limit.Rsh(limit, 1)
	return new(big.Int).Neg(limit), new(big.Int).Sub(limit, big.NewInt(1))
}

func Validate(c Case) error {
	if _, err := decimalref.ParametersFor(c.Width); err != nil {
		return err
	}
	switch c.Mode {
	case "nearest_even", "nearest_away", "toward_zero", "toward_positive", "toward_negative":
	default:
		return fmt.Errorf("unsupported rounding mode")
	}
	arity, hasParam, hasTarget := 2, false, false
	switch c.Op {
	case "rem", "fmod", "minnum", "maxnum", "quiet_equal", "quiet_not_equal", "quiet_greater", "quiet_greater_equal", "quiet_greater_unordered", "quiet_less", "quiet_less_equal", "quiet_less_unordered", "quiet_not_greater", "quiet_not_less", "quiet_ordered", "quiet_unordered":
	case "scaleb":
		arity, hasParam = 1, true
		n, err := canonicalInteger(c.Param, 19)
		if err != nil || !n.IsInt64() {
			return fmt.Errorf("scaleb parameter must be a canonical signed 64-bit integer")
		}
	case "convert":
		arity, hasTarget = 1, true
		if _, err := decimalref.ParametersFor(c.Target); err != nil || c.Target == c.Width {
			return fmt.Errorf("conversion target must be a distinct decimal width")
		}
	case "from_int", "from_uint":
		arity, hasParam, hasTarget = 0, true, true
		if c.Target != 32 && c.Target != 64 {
			return fmt.Errorf("unsupported input integer width")
		}
		n, err := canonicalInteger(c.Param, 20)
		if err != nil {
			return err
		}
		lo, hi := integerBounds(c.Target, c.Op == "from_uint")
		if n.Cmp(lo) < 0 || n.Cmp(hi) > 0 {
			return fmt.Errorf("input integer outside declared width")
		}
	case "to_int", "to_uint", "to_int_exact", "to_uint_exact":
		arity, hasTarget = 1, true
		if c.Target != 8 && c.Target != 16 && c.Target != 32 && c.Target != 64 {
			return fmt.Errorf("unsupported output integer width")
		}
	default:
		return fmt.Errorf("unsupported operation")
	}
	if !hasParam && c.Param != "" {
		return fmt.Errorf("unexpected parameter")
	}
	if !hasTarget && c.Target != 0 {
		return fmt.Errorf("unexpected target")
	}
	if len(c.Operands) != arity {
		return fmt.Errorf("operation requires %d operands", arity)
	}
	for i, raw := range c.Operands {
		if _, err := decimalref.Decode(c.Width, raw); err != nil {
			return fmt.Errorf("operand %d: %w", i, err)
		}
	}
	return nil
}
