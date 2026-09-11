package tier1ref

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/sky1core/bid754/bid754-go/internal/decimalref"
)

func Compare(c Case, want Result, actual Observation) error {
	if err := Validate(c); err != nil {
		return err
	}
	if want.Flags & ^uint32(0x3d) != 0 || actual.Flags & ^uint32(0x3d) != 0 {
		return fmt.Errorf("flags outside IEEE mask")
	}
	if actual.HasFlags {
		if actual.Flags != want.Flags {
			return fmt.Errorf("flags: got %#x, want %#x", actual.Flags, want.Flags)
		}
	} else if actual.Flags != 0 {
		return fmt.Errorf("flags supplied without HasFlags")
	}
	if actual.Kind != want.Kind || actual.Width != want.Width {
		return fmt.Errorf("result kind/width: got %s/%d, want %s/%d", actual.Kind, actual.Width, want.Kind, want.Width)
	}
	switch want.Kind {
	case "integer":
		if _, err := canonicalInteger(actual.Value, 20); err != nil {
			return err
		}
		if actual.Value != want.Value {
			return fmt.Errorf("integer: got %s, want %s", actual.Value, want.Value)
		}
	case "bool":
		if actual.Value != "true" && actual.Value != "false" {
			return fmt.Errorf("invalid bool result")
		}
		if actual.Value != want.Value {
			return fmt.Errorf("bool: got %s, want %s", actual.Value, want.Value)
		}
	case "decimal":
		if _, err := decimalref.Encode(want.Width, want.Decimal); err != nil {
			return fmt.Errorf("expected decimal: %w", err)
		}
		d, err := decimalref.Decode(actual.Width, actual.Value)
		if err != nil {
			return err
		}
		if d.Kind != want.Decimal.Kind {
			return fmt.Errorf("decimal class: got %s, want %s", d.Kind, want.Decimal.Kind)
		}
		if d.Kind == "nan" {
			if signaling(actual.Width, actual.Value) {
				return fmt.Errorf("signaling NaN result")
			}
			p, _ := decimalref.ParametersFor(actual.Width)
			payloadBits := map[int]int{32: 20, 64: 50, 128: 110}[actual.Width]
			bits, _ := new(big.Int).SetString(strings.ReplaceAll(actual.Value, ":", ""), 16)
			bits.SetBit(bits, actual.Width-1, 0)
			bits.And(bits, new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(actual.Width-6)), big.NewInt(1)))
			if bits.BitLen() > payloadBits || bits.Cmp(pow10(p.Precision-1)) >= 0 {
				return fmt.Errorf("noncanonical NaN result")
			}
			return nil
		}
		canonical, err := decimalref.Encode(actual.Width, d)
		if err != nil {
			return err
		}
		if canonical != actual.Value {
			return fmt.Errorf("noncanonical decimal result")
		}
		if (c.Op == "minnum" || c.Op == "maxnum") && d.Kind == "finite" {
			selectedOperand := false
			for _, raw := range c.Operands {
				operand, _ := decimalref.Decode(c.Width, raw)
				encoded, err := decimalref.Encode(c.Width, operand)
				if err == nil && encoded == actual.Value {
					selectedOperand = true
				}
			}
			if !selectedOperand {
				return fmt.Errorf("min/max result is neither canonical input operand")
			}
		}
		if d.Negative != want.Decimal.Negative && !(want.AnyZeroSign && d.Kind == "finite" && d.Coeff.Sign() == 0 && want.Decimal.Coeff.Sign() == 0) {
			return fmt.Errorf("decimal sign differs")
		}
		if compareNumbers(d, want.Decimal) != 0 {
			return fmt.Errorf("decimal numeric value differs")
		}
		if want.Quantum && d.Kind == "finite" && d.Exp != want.Decimal.Exp {
			return fmt.Errorf("quantum: got %d, want %d", d.Exp, want.Decimal.Exp)
		}
	default:
		return fmt.Errorf("unsupported result kind")
	}
	return nil
}
