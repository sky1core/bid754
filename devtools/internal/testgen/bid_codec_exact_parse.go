package testgen

import (
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"strings"
)

type bidCodecExactParseExpectation struct {
	Input    string
	Expected string
	Width    int
	Lo, Hi   uint64
}

var bidCodecExpectedFinite = regexp.MustCompile(`^([1-9][0-9]*)(?:\.([0-9]+))?E([+-][0-9]+)$`)
var bidCodecExpectedZero = regexp.MustCompile(`^0(?:E([+-][0-9]+))?$`)
var bidCodecExpectedNaN = regexp.MustCompile(`^(S?NaN)([0-9]*)$`)

func bidCodecParseExactExpected(expected string) (bidCodecRefComponents, error) {
	var c bidCodecRefComponents
	if len(expected) < 2 || (expected[0] != '+' && expected[0] != '-') {
		return c, fmt.Errorf("expected %q lacks an explicit sign", expected)
	}
	c.Sign = expected[0] == '-'
	body := expected[1:]
	switch body {
	case "Inf":
		c.Kind = bidCodecRefInfinity
	case "0":
		c.Kind = bidCodecRefZero
	default:
		if m := bidCodecExpectedNaN.FindStringSubmatch(body); m != nil {
			c.Kind = bidCodecRefQNaN
			if m[1] == "SNaN" {
				c.Kind = bidCodecRefSNaN
			}
			c.Payload = new(big.Int)
			if m[2] != "" {
				c.Payload.SetString(m[2], 10)
			}
		} else if m := bidCodecExpectedZero.FindStringSubmatch(body); m != nil {
			c.Kind = bidCodecRefZero
			exp, err := strconv.ParseInt(m[1], 10, 32)
			if err != nil || m[1] == "" {
				return c, fmt.Errorf("expected %q has invalid zero exponent", expected)
			}
			c.Exponent = int32(exp)
		} else if m := bidCodecExpectedFinite.FindStringSubmatch(body); m != nil {
			adjusted, err := strconv.ParseInt(m[3], 10, 64)
			if err != nil {
				return c, fmt.Errorf("expected %q has invalid adjusted exponent: %w", expected, err)
			}
			exp := adjusted - int64(len(m[1])+len(m[2])-1)
			if exp < -1<<31 || exp > 1<<31-1 {
				return c, fmt.Errorf("expected %q has out-of-range coefficient exponent", expected)
			}
			c.Kind = bidCodecRefNormal
			c.Exponent = int32(exp)
			c.Coefficient, _ = new(big.Int).SetString(m[1]+m[2], 10)
		} else {
			return c, fmt.Errorf("expected %q is outside normalized exact grammar", expected)
		}
	}
	if bidCodecDecimalString(c) != expected {
		return c, fmt.Errorf("expected %q is not a canonical rendering", expected)
	}
	return c, nil
}

func bidCodecExactExpectedBits(expected string, width int) (uint64, uint64, error) {
	c, err := bidCodecParseExactExpected(expected)
	if err != nil {
		return 0, 0, err
	}
	var precision int
	var minExp, maxExp int32
	var maxPayload *big.Int
	switch width {
	case 32:
		precision, minExp, maxExp = 7, -101, 90
		maxPayload = big.NewInt(1000000)
	case 64:
		precision, minExp, maxExp = 16, -398, 369
		maxPayload = big.NewInt(1000000000000000)
	case 128:
		precision, minExp, maxExp = 34, -6176, 6111
		maxPayload = bid128RefTen33
	default:
		return 0, 0, fmt.Errorf("unsupported exact width %d", width)
	}
	if c.Kind == bidCodecRefNormal || c.Kind == bidCodecRefZero {
		if c.Exponent < minExp || c.Exponent > maxExp {
			return 0, 0, fmt.Errorf("expected %q d%d exponent out of range", expected, width)
		}
		if c.Kind == bidCodecRefNormal && len(c.Coefficient.String()) > precision {
			return 0, 0, fmt.Errorf("expected %q d%d coefficient out of range", expected, width)
		}
	}
	if (c.Kind == bidCodecRefQNaN || c.Kind == bidCodecRefSNaN) && c.Payload.Cmp(maxPayload) >= 0 {
		return 0, 0, fmt.Errorf("expected %q d%d payload out of range", expected, width)
	}
	switch width {
	case 32:
		return uint64(refEncode32(c)), 0, nil
	case 64:
		return refEncode64(c), 0, nil
	default:
		lo, hi := refEncode128(c)
		return lo, hi, nil
	}
}

func bidCodecExactParseExpectations() ([]bidCodecExactParseExpectation, error) {
	var rows []bidCodecExactParseExpectation
	for _, sv := range bidCodecGoFullStringVectorRecords() {
		classes := bidCodecGoFullStringVectorClasses[sv.Input]
		for i, width := range []int{32, 64, 128} {
			if classes[i] != "exact" {
				continue
			}
			lo, hi, err := bidCodecExactExpectedBits(sv.Expected, width)
			if err != nil {
				return nil, fmt.Errorf("string_vectors input %q: %w", sv.Input, err)
			}
			rows = append(rows, bidCodecExactParseExpectation{Input: sv.Input, Expected: sv.Expected, Width: width, Lo: lo, Hi: hi})
		}
	}
	return rows, nil
}

func bidCodecExactParseExpectationLiterals(rows []bidCodecExactParseExpectation, rust bool) string {
	var b strings.Builder
	for _, row := range rows {
		if rust {
			fmt.Fprintf(&b, "\n    ExactParseExpectation { input: %s, expected: %s, width: %d, lo: 0x%x, hi: 0x%x },", rustStringLiteral(row.Input), rustStringLiteral(row.Expected), row.Width, row.Lo, row.Hi)
		} else {
			fmt.Fprintf(&b, "\n\t{%q, %q, %d, 0x%x, 0x%x},", row.Input, row.Expected, row.Width, row.Lo, row.Hi)
		}
	}
	b.WriteByte('\n')
	return b.String()
}
