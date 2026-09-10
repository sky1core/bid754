package decimalref

import (
	"fmt"
	"math/big"
	"strings"
)

func lowBits(n int) *big.Int {
	return new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(n)), big.NewInt(1))
}

func field(bits *big.Int, start, count int) *big.Int {
	return new(big.Int).And(new(big.Int).Rsh(bits, uint(start)), lowBits(count))
}

func coefficientBits(width int) int {
	switch width {
	case 32:
		return 23
	case 64:
		return 53
	default:
		return 113
	}
}

func parseRaw(width int, raw string) (*big.Int, error) {
	if _, err := ParametersFor(width); err != nil {
		return nil, err
	}
	n := width / 4
	if width == 128 {
		n++
	}
	if len(raw) != n {
		return nil, fmt.Errorf("raw length must be %d", n)
	}
	for i := 0; i < n; i++ {
		if width == 128 && i == 16 {
			if raw[i] != ':' {
				return nil, fmt.Errorf("missing hi:lo separator")
			}
			continue
		}
		if !((raw[i] >= '0' && raw[i] <= '9') || (raw[i] >= 'a' && raw[i] <= 'f')) {
			return nil, fmt.Errorf("raw must be padded lowercase hex")
		}
	}
	bits, ok := new(big.Int).SetString(strings.ReplaceAll(raw, ":", ""), 16)
	if !ok {
		return nil, fmt.Errorf("invalid raw hex")
	}
	return bits, nil
}

func Decode(width int, raw string) (Decimal, error) {
	bits, err := parseRaw(width, raw)
	if err != nil {
		return Decimal{}, err
	}
	p, _ := ParametersFor(width)
	d := Decimal{Kind: "finite", Negative: bits.Bit(width-1) != 0, Coeff: new(big.Int)}
	combination := field(bits, width-6, 5).Int64()
	if combination == 30 {
		d.Kind = "infinity"
		return d, nil
	}
	if combination == 31 {
		d.Kind = "nan"
		return d, nil
	}
	cb := coefficientBits(width)
	shift := cb
	if field(bits, width-3, 2).Int64() == 3 {
		shift -= 2
		d.Coeff.Set(field(bits, 0, shift))
		d.Coeff.SetBit(d.Coeff, cb, 1)
	} else {
		d.Coeff.Set(field(bits, 0, cb))
	}
	d.Exp = int(field(bits, shift, width-cb-1).Int64()) + p.MinExp
	if d.Coeff.Cmp(power10(p.Precision)) >= 0 {
		d.Coeff.SetInt64(0)
	}
	return d, nil
}

func Encode(width int, d Decimal) (string, error) {
	p, err := ParametersFor(width)
	if err != nil {
		return "", err
	}
	if err := validate(p, d); err != nil {
		return "", err
	}
	bits := new(big.Int)
	switch d.Kind {
	case "infinity":
		bits.Lsh(big.NewInt(30), uint(width-6))
	case "nan":
		bits.Lsh(big.NewInt(31), uint(width-6))
	case "finite":
		cb := coefficientBits(width)
		shift := cb
		bits.Set(d.Coeff)
		if d.Coeff.BitLen() > cb {
			shift -= 2
			bits.And(bits, lowBits(shift))
			bits.Or(bits, new(big.Int).Lsh(big.NewInt(3), uint(width-3)))
		}
		bits.Or(bits, new(big.Int).Lsh(big.NewInt(int64(d.Exp-p.MinExp)), uint(shift)))
	}
	if d.Negative {
		bits.SetBit(bits, width-1, 1)
	}
	raw := fmt.Sprintf("%0*x", width/4, bits)
	if width == 128 {
		raw = raw[:16] + ":" + raw[16:]
	}
	return raw, nil
}

func Compare(width int, want Result, actualBits string, actualFlags uint32) error {
	p, err := ParametersFor(width)
	if err != nil {
		return err
	}
	const ieeeFlags = flagInvalid | flagDivisionByZero | flagOverflow | flagUnderflow | flagInexact
	if want.Flags & ^ieeeFlags != 0 || actualFlags & ^ieeeFlags != 0 {
		return fmt.Errorf("flags outside IEEE five-flag word: got %#x, want %#x", actualFlags, want.Flags)
	}
	if err := validate(p, want.Value); err != nil {
		return fmt.Errorf("expected value: %w", err)
	}
	actual, err := Decode(width, actualBits)
	if err != nil {
		return err
	}
	bits, _ := parseRaw(width, actualBits)
	if actual.Kind == "nan" {
		payloadBits := coefficientBits(width) - 3
		if bits.Bit(width-7) != 0 {
			return fmt.Errorf("actual result is signaling NaN")
		}
		if field(bits, payloadBits, width-7-payloadBits).Sign() != 0 || field(bits, 0, payloadBits).Cmp(power10(p.Precision-1)) >= 0 {
			return fmt.Errorf("noncanonical actual NaN")
		}
	} else {
		canonical, err := Encode(width, actual)
		if err != nil {
			return err
		}
		if canonical != actualBits {
			return fmt.Errorf("noncanonical actual result %s", actualBits)
		}
	}
	if actualFlags != want.Flags {
		return fmt.Errorf("flags: got %#x, want %#x", actualFlags, want.Flags)
	}
	if actual.Kind != want.Value.Kind {
		return fmt.Errorf("class: got %s, want %s", actual.Kind, want.Value.Kind)
	}
	if actual.Kind == "nan" {
		return nil
	}
	if actual.Negative != want.Value.Negative {
		return fmt.Errorf("sign: got negative=%t, want %t", actual.Negative, want.Value.Negative)
	}
	if actual.Kind == "finite" && exactValue(actual).Cmp(exactValue(want.Value)) != 0 {
		return fmt.Errorf("finite value differs: got %sE%d, want %sE%d", actual.Coeff, actual.Exp, want.Value.Coeff, want.Value.Exp)
	}
	return nil
}

func CompareQuantum(width int, want Result, actualBits string, actualFlags uint32) error {
	if err := Compare(width, want, actualBits, actualFlags); err != nil {
		return err
	}
	if want.Value.Kind != "finite" {
		return nil
	}
	actual, err := Decode(width, actualBits)
	if err != nil {
		return err
	}
	if actual.Exp != want.Value.Exp || actual.Coeff.Cmp(want.Value.Coeff) != 0 {
		return fmt.Errorf("finite cohort differs: got %sE%d, want %sE%d", actual.Coeff, actual.Exp, want.Value.Coeff, want.Value.Exp)
	}
	return nil
}
