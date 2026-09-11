package tier1ref

import (
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/sky1core/bid754/bid754-go/internal/decimalref"
)

var modes = []string{"nearest_even", "nearest_away", "toward_zero", "toward_positive", "toward_negative"}

func number(coeff string, exponent int) decimalref.Decimal {
	n, ok := new(big.Int).SetString(coeff, 10)
	if !ok {
		panic("bad test integer")
	}
	return decimalref.Decimal{Kind: "finite", Negative: strings.HasPrefix(coeff, "-"), Coeff: n.Abs(n), Exp: exponent}
}

func makeCase(t *testing.T, width int, op, mode string, values ...decimalref.Decimal) Case {
	t.Helper()
	c := Case{Width: width, Op: op, Mode: mode}
	for _, d := range values {
		raw, err := decimalref.Encode(width, d)
		if err != nil {
			t.Fatal(err)
		}
		c.Operands = append(c.Operands, raw)
	}
	return c
}

func evaluate(t *testing.T, c Case) Result {
	t.Helper()
	r, err := Evaluate(c)
	if err != nil {
		t.Fatalf("%+v: %v", c, err)
	}
	return r
}

func checkDecimal(t *testing.T, c Case, want decimalref.Decimal, flags uint32) {
	t.Helper()
	r := evaluate(t, c)
	width := c.Width
	if c.Op == "convert" {
		width = c.Target
	}
	if r.Kind != "decimal" || r.Width != width || r.Flags != flags || r.Decimal.Kind != want.Kind || r.Decimal.Negative != want.Negative {
		t.Fatalf("%+v: got %+v, want %+v flags=%#x", c, r, want, flags)
	}
	if want.Kind == "finite" && (r.Decimal.Exp != want.Exp || r.Decimal.Coeff.Cmp(want.Coeff) != 0) {
		t.Fatalf("%+v: got %sE%d, want %sE%d", c, r.Decimal.Coeff, r.Decimal.Exp, want.Coeff, want.Exp)
	}
	bits, err := decimalref.Encode(width, want)
	if err != nil {
		t.Fatal(err)
	}
	if err := Compare(c, r, Observation{Kind: "decimal", Width: width, Value: bits, HasFlags: true, Flags: flags}); err != nil {
		t.Fatal(err)
	}
}

func checkInteger(t *testing.T, c Case, value string, flags uint32) {
	t.Helper()
	r := evaluate(t, c)
	if r.Kind != "integer" || r.Width != c.Target || r.Value != value || r.Flags != flags {
		t.Fatalf("%+v: got %+v, want integer=%s flags=%#x", c, r, value, flags)
	}
}

func TestRemainderWitnesses(t *testing.T) {
	for _, width := range []int{32, 64, 128} {
		for _, mode := range modes {
			for _, row := range []struct{ x, y, rem, fmod string }{
				{"7", "2", "-1", "1"}, {"5", "2", "1", "1"}, {"3", "2", "-1", "1"},
				{"-7", "2", "1", "-1"}, {"7", "-2", "-1", "1"}, {"-7", "-2", "1", "-1"},
				{"6", "2", "0", "0"}, {"-6", "2", "-0", "-0"},
			} {
				checkDecimal(t, makeCase(t, width, "rem", mode, number(row.x, 0), number(row.y, 0)), number(row.rem, 0), 0)
				checkDecimal(t, makeCase(t, width, "fmod", mode, number(row.x, 0), number(row.y, 0)), number(row.fmod, 0), 0)
			}
			checkDecimal(t, makeCase(t, width, "rem", mode, number("75", -1), number("2", 0)), number("-5", -1), 0)
			checkDecimal(t, makeCase(t, width, "fmod", mode, number("75", -1), number("2", 0)), number("15", -1), 0)
			for _, op := range []string{"rem", "fmod"} {
				checkDecimal(t, makeCase(t, width, op, mode, number("-0", 3), number("2", -2)), number("-0", -2), 0)
				p, _ := decimalref.ParametersFor(width)
				checkDecimal(t, makeCase(t, width, op, mode, number("1", p.MaxExp), number("3", p.MinExp)), number("1", p.MinExp), 0)
			}
		}
	}
}

func TestScaleWitnesses(t *testing.T) {
	for _, width := range []int{32, 64, 128} {
		p, _ := decimalref.ParametersFor(width)
		for _, mode := range modes {
			for _, negative := range []bool{false, true} {
				prefix := ""
				if negative {
					prefix = "-"
				}
				for _, shift := range []string{"1", "9223372036854775807"} {
					c := makeCase(t, width, "scaleb", mode, number(prefix+strings.Repeat("9", p.Precision), p.MaxExp))
					c.Param = shift
					want := decimalref.Decimal{Kind: "infinity", Negative: negative}
					if mode == "toward_zero" || mode == "toward_positive" && negative || mode == "toward_negative" && !negative {
						want = number(prefix+strings.Repeat("9", p.Precision), p.MaxExp)
					}
					checkDecimal(t, c, want, 0x28)
				}
				for _, shift := range []string{"-1", "-9223372036854775808"} {
					c := makeCase(t, width, "scaleb", mode, number(prefix+"1", p.MinExp))
					c.Param = shift
					want := "0"
					if mode == "toward_positive" && !negative || mode == "toward_negative" && negative {
						want = "1"
					}
					checkDecimal(t, c, number(prefix+want, p.MinExp), 0x30)
				}
				for _, row := range []struct {
					coeff      string
					even, away string
				}{{"5", "0", "1"}, {"15", "2", "2"}, {"25", "2", "3"}} {
					c := makeCase(t, width, "scaleb", mode, number(prefix+row.coeff, p.MinExp))
					c.Param = "-1"
					want := row.even
					switch mode {
					case "nearest_away":
						want = row.away
					case "toward_zero", "toward_positive", "toward_negative":
						want = string(row.coeff[0:1])
						if len(row.coeff) == 1 {
							want = "0"
						}
						if mode == "toward_positive" && !negative || mode == "toward_negative" && negative {
							n, _ := new(big.Int).SetString(want, 10)
							want = n.Add(n, big.NewInt(1)).String()
						}
					}
					checkDecimal(t, c, number(prefix+want, p.MinExp), 0x30)
				}
				for _, row := range []struct {
					shift    string
					exponent int
				}{{"9223372036854775807", p.MaxExp}, {"-9223372036854775808", p.MinExp}} {
					c := makeCase(t, width, "scaleb", mode, number(prefix+"0", 0))
					c.Param = row.shift
					checkDecimal(t, c, number(prefix+"0", row.exponent), 0)
				}
			}
			c := makeCase(t, width, "scaleb", mode, number("10", p.MinExp))
			c.Param = "-1"
			checkDecimal(t, c, number("1", p.MinExp), 0)
			c = makeCase(t, width, "scaleb", mode, number("1", p.MaxExp))
			c.Param = "1"
			checkDecimal(t, c, number("10", p.MaxExp), 0)
		}
		c := makeCase(t, width, "scaleb", "nearest_even", number(strings.Repeat("9", p.Precision-1)+"5", p.MinExp))
		c.Param = "-1"
		checkDecimal(t, c, number("1"+strings.Repeat("0", p.Precision-1), p.MinExp), 0x30)
	}
}

func TestWidthConversionWitnesses(t *testing.T) {
	for _, source := range []int{32, 64, 128} {
		for _, target := range []int{32, 64, 128} {
			if source == target {
				continue
			}
			sp, _ := decimalref.ParametersFor(source)
			tp, _ := decimalref.ParametersFor(target)
			for _, mode := range modes {
				c := makeCase(t, source, "convert", mode, number("12300", -4))
				c.Target = target
				checkDecimal(t, c, number("12300", -4), 0)
				for _, negative := range []bool{false, true} {
					prefix := ""
					if negative {
						prefix = "-"
					}
					for _, exp := range []int{sp.MinExp, sp.MaxExp} {
						c = makeCase(t, source, "convert", mode, number(prefix+"0", exp))
						c.Target = target
						checkDecimal(t, c, number(prefix+"0", max(tp.MinExp, min(tp.MaxExp, exp))), 0)
					}
					if source < target {
						c = makeCase(t, source, "convert", mode, number(prefix+"1", sp.MinExp))
						c.Target = target
						checkDecimal(t, c, number(prefix+"1", sp.MinExp), 0)
						continue
					}
					c = makeCase(t, source, "convert", mode, number(prefix+"1"+strings.Repeat("0", tp.Precision-1)+"5", -1))
					c.Target = target
					last := "0"
					if mode == "nearest_away" || mode == "toward_positive" && !negative || mode == "toward_negative" && negative {
						last = "1"
					}
					checkDecimal(t, c, number(prefix+"1"+strings.Repeat("0", tp.Precision-2)+last, 0), 0x20)
					c = makeCase(t, source, "convert", mode, number(prefix+"1", tp.MinExp-1))
					c.Target = target
					last = "0"
					if mode == "toward_positive" && !negative || mode == "toward_negative" && negative {
						last = "1"
					}
					checkDecimal(t, c, number(prefix+last, tp.MinExp), 0x30)
					c = makeCase(t, source, "convert", mode, number(prefix+"1", tp.MaxExp+tp.Precision))
					c.Target = target
					want := decimalref.Decimal{Kind: "infinity", Negative: negative}
					if mode == "toward_zero" || mode == "toward_positive" && negative || mode == "toward_negative" && !negative {
						want = number(prefix+strings.Repeat("9", tp.Precision), tp.MaxExp)
					}
					checkDecimal(t, c, want, 0x28)
				}
			}
		}
	}
}

func TestIntegerConversionWitnesses(t *testing.T) {
	for _, width := range []int{32, 64, 128} {
		for _, mode := range modes {
			for _, bits := range []int{8, 16, 32, 64} {
				for _, op := range []string{"to_int", "to_uint", "to_int_exact", "to_uint_exact"} {
					c := makeCase(t, width, op, mode, number("125", -1))
					c.Target = bits
					want := "12"
					if mode == "nearest_away" || mode == "toward_positive" {
						want = "13"
					}
					flags := uint32(0)
					if strings.HasSuffix(op, "_exact") {
						flags = 0x20
					}
					checkInteger(t, c, want, flags)
					c.Operands = makeCase(t, width, op, mode, number("-0", -2)).Operands
					checkInteger(t, c, "0", 0)
				}
			}
		}
	}
	for _, row := range []struct {
		bits                                      int
		low, high, unsignedHigh, unsignedSentinel string
	}{
		{8, "-128", "127", "255", "128"}, {16, "-32768", "32767", "65535", "32768"},
		{32, "-2147483648", "2147483647", "4294967295", "2147483648"},
		{64, "-9223372036854775808", "9223372036854775807", "18446744073709551615", "9223372036854775808"},
	} {
		for _, mode := range modes {
			for _, exact := range []bool{false, true} {
				op := "to_int"
				if exact {
					op += "_exact"
				}
				for _, value := range []string{row.low, row.high} {
					c := makeCase(t, 128, op, mode, number(value, 0))
					c.Target = row.bits
					checkInteger(t, c, value, 0)
				}
				c := makeCase(t, 128, op, mode, number(row.high+"5", -1))
				c.Target = row.bits
				if mode == "toward_zero" || mode == "toward_negative" {
					flags := uint32(0)
					if exact {
						flags = 0x20
					}
					checkInteger(t, c, row.high, flags)
				} else {
					checkInteger(t, c, row.low, 0x01)
				}
				c.Operands = makeCase(t, 128, op, mode, number(row.low+"5", -1)).Operands
				if mode == "nearest_away" || mode == "toward_negative" {
					checkInteger(t, c, row.low, 0x01)
				} else {
					flags := uint32(0)
					if exact {
						flags = 0x20
					}
					checkInteger(t, c, row.low, flags)
				}
				op = "to_uint"
				if exact {
					op += "_exact"
				}
				c = makeCase(t, 128, op, mode, number(row.unsignedHigh, 0))
				c.Target = row.bits
				checkInteger(t, c, row.unsignedHigh, 0)
				c.Operands = makeCase(t, 128, op, mode, number(row.unsignedHigh+"5", -1)).Operands
				if mode == "toward_zero" || mode == "toward_negative" {
					flags := uint32(0)
					if exact {
						flags = 0x20
					}
					checkInteger(t, c, row.unsignedHigh, flags)
				} else {
					checkInteger(t, c, row.unsignedSentinel, 0x01)
				}
				c.Operands = makeCase(t, 128, op, mode, number("-5", -1)).Operands
				if mode == "nearest_away" || mode == "toward_negative" {
					checkInteger(t, c, row.unsignedSentinel, 0x01)
				} else {
					flags := uint32(0)
					if exact {
						flags = 0x20
					}
					checkInteger(t, c, "0", flags)
				}
				c.Operands = makeCase(t, 128, op, mode, number("1", 6111)).Operands
				checkInteger(t, c, row.unsignedSentinel, 0x01)
			}
		}
	}
}

func TestFromIntegerWitnesses(t *testing.T) {
	for _, row := range []struct {
		op         string
		bits       int
		value, c32 string
		e32        int
		c64        string
		e64        int
	}{
		{"from_int", 32, "2147483647", "2147483", 3, "2147483647", 0},
		{"from_int", 32, "-2147483648", "2147483", 3, "2147483648", 0},
		{"from_uint", 32, "4294967295", "4294967", 3, "4294967295", 0},
		{"from_int", 64, "9223372036854775807", "9223372", 12, "9223372036854775", 3},
		{"from_int", 64, "-9223372036854775808", "9223372", 12, "9223372036854775", 3},
		{"from_uint", 64, "18446744073709551615", "1844674", 13, "1844674407370955", 4},
	} {
		for _, width := range []int{32, 64, 128} {
			for _, mode := range modes {
				c := Case{Width: width, Op: row.op, Mode: mode, Target: row.bits, Param: row.value}
				negative := strings.HasPrefix(row.value, "-")
				coeff, exp := strings.TrimPrefix(row.value, "-"), 0
				if width == 32 {
					coeff, exp = row.c32, row.e32
				}
				if width == 64 {
					coeff, exp = row.c64, row.e64
				}
				flags := uint32(0)
				if exp != 0 {
					flags = 0x20
					up := mode == "toward_positive" && !negative || mode == "toward_negative" && negative
					if mode == "nearest_even" || mode == "nearest_away" {
						up = (width == 32 && row.op == "from_int" && row.bits == 32) || (width == 64 && row.op == "from_int" && row.bits == 64)
					}
					if up {
						n, _ := new(big.Int).SetString(coeff, 10)
						coeff = n.Add(n, big.NewInt(1)).String()
					}
				}
				if negative {
					coeff = "-" + coeff
				}
				checkDecimal(t, c, number(coeff, exp), flags)
			}
		}
	}
}

func TestNoMutableResultSharing(t *testing.T) {
	c := makeCase(t, 32, "rem", "nearest_even", number("7", 0), number("2", 0))
	r := evaluate(t, c)
	r.Decimal.Coeff.SetInt64(99)
	checkDecimal(t, c, number("-1", 0), 0)
	if fmt.Sprint(c.Operands) != "[32800007 32800002]" {
		t.Fatal("input changed")
	}
}
