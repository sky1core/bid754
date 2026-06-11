package bidcodec

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

// ToString converts Components to IEEE 754 string representation.
// Examples: "+1234567E+3", "-INF", "+NaN"
func ToString(c Components) string {
	prefix := "+"
	if c.Sign {
		prefix = "-"
	}
	switch c.Kind {
	case Infinity:
		return prefix + "Inf"
	case QNaN:
		if c.Payload != 0 {
			return fmt.Sprintf("%sNaN%d", prefix, c.Payload)
		}
		return prefix + "NaN"
	case SNaN:
		if c.Payload != 0 {
			return fmt.Sprintf("%sSNaN%d", prefix, c.Payload)
		}
		return prefix + "SNaN"
	case Zero:
		if c.Exponent == 0 {
			return prefix + "0"
		}
		return fmt.Sprintf("%s0E%+d", prefix, c.Exponent)
	}
	// Normal
	digits := c.Coefficient.String()
	exp := int(c.Exponent) + len(digits) - 1
	if len(digits) == 1 {
		return fmt.Sprintf("%s%sE%+d", prefix, digits, exp)
	}
	return fmt.Sprintf("%s%s.%sE%+d", prefix, digits[:1], digits[1:], exp)
}

// FromString parses an IEEE 754 string into Components.
// Supports: "123.45", "+1.23E+5", "-INF", "NaN", "SNaN123"
func FromString(s string) (Components, error) {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return Components{}, fmt.Errorf("empty string")
	}

	sign := false
	if s[0] == '+' {
		s = s[1:]
	} else if s[0] == '-' {
		sign = true
		s = s[1:]
	}

	upper := strings.ToUpper(s)
	if upper == "INF" || upper == "INFINITY" {
		return Components{Sign: sign, Kind: Infinity}, nil
	}
	if strings.HasPrefix(upper, "SNAN") {
		payload := uint64(0)
		if len(s) > 4 {
			p, err := parseUint64Payload(s[4:])
			if err != nil {
				return Components{}, err
			}
			payload = p
		}
		return Components{Sign: sign, Kind: SNaN, Payload: payload}, nil
	}
	if strings.HasPrefix(upper, "NAN") {
		payload := uint64(0)
		if len(s) > 3 {
			p, err := parseUint64Payload(s[3:])
			if err != nil {
				return Components{}, err
			}
			payload = p
		}
		return Components{Sign: sign, Kind: QNaN, Payload: payload}, nil
	}

	// Parse number: digits, decimal point, exponent
	var digits []byte
	expAdjust := 0
	foundDot := false
	i := 0
	for i < len(s) && s[i] != 'E' && s[i] != 'e' {
		if s[i] == '.' {
			if foundDot {
				return Components{}, fmt.Errorf("multiple decimal points")
			}
			foundDot = true
		} else if s[i] >= '0' && s[i] <= '9' {
			digits = append(digits, s[i])
			if foundDot {
				expAdjust--
			}
		} else {
			return Components{}, fmt.Errorf("unexpected character: %c", s[i])
		}
		i++
	}

	expPart := int64(0)
	if i < len(s) && (s[i] == 'E' || s[i] == 'e') {
		i++
		expStr := s[i:]
		n, err := strconv.ParseInt(expStr, 10, 32)
		if err != nil {
			return Components{}, fmt.Errorf("invalid exponent: %s", expStr)
		}
		expPart = n
	}

	if len(digits) == 0 {
		return Components{}, fmt.Errorf("no digits")
	}

	// Remove leading zeros
	start := 0
	for start < len(digits)-1 && digits[start] == '0' {
		start++
	}
	digits = digits[start:]

	coeff, ok := new(big.Int).SetString(string(digits), 10)
	if !ok {
		return Components{}, fmt.Errorf("invalid coefficient: %s", string(digits))
	}

	exponent64 := expPart + int64(expAdjust)
	if exponent64 < -2147483648 || exponent64 > 2147483647 {
		return Components{}, fmt.Errorf("exponent out of int32 range: %d", exponent64)
	}
	exponent := int32(exponent64)

	if coeff.Sign() == 0 {
		return Components{Sign: sign, Exponent: exponent, Kind: Zero}, nil
	}
	return Components{
		Sign:        sign,
		Coefficient: coeff,
		Exponent:    exponent,
		Kind:        Normal,
	}, nil
}

func parseUint64Payload(s string) (uint64, error) {
	if s == "" {
		return 0, fmt.Errorf("empty NaN payload")
	}
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid NaN payload: %s", s)
	}
	return v, nil
}
