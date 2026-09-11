package tier1ref

import (
	"math/big"
	"strings"
	"testing"

	"github.com/sky1core/bid754/bid754-go/internal/decimalref"
)

func specialRaw(t *testing.T, width int, kind string, negative, snan bool) string {
	t.Helper()
	raw, err := decimalref.Encode(width, decimalref.Decimal{Kind: kind, Negative: negative})
	if err != nil {
		t.Fatal(err)
	}
	if snan {
		bits, _ := new(big.Int).SetString(strings.ReplaceAll(raw, ":", ""), 16)
		bits.SetBit(bits, width-7, 1)
		raw = bits.Text(16)
		if width == 128 {
			raw = raw[:16] + ":" + raw[16:]
		}
	}
	return raw
}

func TestSpecialArithmeticAndIntegerResults(t *testing.T) {
	for _, width := range []int{32, 64, 128} {
		for _, mode := range modes {
			for _, op := range []string{"rem", "fmod", "minnum", "maxnum", "scaleb", "convert"} {
				for _, snan := range []bool{false, true} {
					for _, negative := range []bool{false, true} {
						c := makeCase(t, width, op, mode, number("1", -2))
						if op == "scaleb" {
							c.Param = "9223372036854775807"
						} else if op == "convert" {
							c.Target = 32
							if width == 32 {
								c.Target = 128
							}
						} else {
							c.Operands = append(c.Operands, c.Operands[0])
						}
						for slot := range c.Operands {
							trial := c
							trial.Operands = append([]string(nil), c.Operands...)
							trial.Operands[slot] = specialRaw(t, width, "nan", negative, snan)
							flags := uint32(0)
							if snan {
								flags = 1
							}
							want := decimalref.Decimal{Kind: "nan"}
							if !snan && (op == "minnum" || op == "maxnum") {
								want = number("1", -2)
							}
							checkDecimal(t, trial, want, flags)
						}
						if len(c.Operands) == 2 {
							c.Operands[0], c.Operands[1] = specialRaw(t, width, "nan", false, false), specialRaw(t, width, "nan", negative, snan)
							flags := uint32(0)
							if snan {
								flags = 1
							}
							checkDecimal(t, c, decimalref.Decimal{Kind: "nan"}, flags)
						}
					}
				}
			}
			for _, op := range []string{"rem", "fmod"} {
				for _, negative := range []bool{false, true} {
					inf := decimalref.Decimal{Kind: "infinity", Negative: negative}
					checkDecimal(t, makeCase(t, width, op, mode, inf, number("1", 0)), decimalref.Decimal{Kind: "nan"}, 1)
					checkDecimal(t, makeCase(t, width, op, mode, inf, inf), decimalref.Decimal{Kind: "nan"}, 1)
					checkDecimal(t, makeCase(t, width, op, mode, number("-0", -3), inf), number("-0", -3), 0)
					checkDecimal(t, makeCase(t, width, op, mode, number("17", -3), inf), number("17", -3), 0)
					checkDecimal(t, makeCase(t, width, op, mode, number("1", 0), number("0", 0)), decimalref.Decimal{Kind: "nan"}, 1)
					checkDecimal(t, makeCase(t, width, op, mode, number("0", 0), number("0", 0)), decimalref.Decimal{Kind: "nan"}, 1)
				}
			}
			for _, negative := range []bool{false, true} {
				inf := decimalref.Decimal{Kind: "infinity", Negative: negative}
				c := makeCase(t, width, "scaleb", mode, inf)
				c.Param = "-9223372036854775808"
				checkDecimal(t, c, inf, 0)
				for _, target := range []int{32, 64, 128} {
					if target == width {
						continue
					}
					c = makeCase(t, width, "convert", mode, inf)
					c.Target = target
					checkDecimal(t, c, inf, 0)
				}
			}
			for _, raw := range []string{specialRaw(t, width, "nan", false, false), specialRaw(t, width, "nan", true, true), specialRaw(t, width, "infinity", false, false), specialRaw(t, width, "infinity", true, false)} {
				for _, row := range []struct {
					bits             int
					signed, unsigned string
				}{{8, "-128", "128"}, {16, "-32768", "32768"}, {32, "-2147483648", "2147483648"}, {64, "-9223372036854775808", "9223372036854775808"}} {
					for _, op := range []string{"to_int", "to_int_exact", "to_uint", "to_uint_exact"} {
						want := row.signed
						if strings.HasPrefix(op, "to_uint") {
							want = row.unsigned
						}
						checkInteger(t, Case{Width: width, Op: op, Mode: mode, Operands: []string{raw}, Target: row.bits}, want, 1)
					}
				}
			}
		}
	}
}

func TestQuietPredicatesTruthTable(t *testing.T) {
	for _, row := range []struct {
		op                              string
		less, equal, greater, unordered bool
	}{
		{"quiet_equal", false, true, false, false}, {"quiet_not_equal", true, false, true, true},
		{"quiet_greater", false, false, true, false}, {"quiet_greater_equal", false, true, true, false},
		{"quiet_greater_unordered", false, false, true, true}, {"quiet_less", true, false, false, false},
		{"quiet_less_equal", true, true, false, false}, {"quiet_less_unordered", true, false, false, true},
		{"quiet_not_greater", true, true, false, true}, {"quiet_not_less", false, true, true, true},
		{"quiet_ordered", true, true, true, false}, {"quiet_unordered", false, false, false, true},
	} {
		for _, width := range []int{32, 64, 128} {
			for _, mode := range modes {
				for _, pair := range []struct {
					a, b decimalref.Decimal
					want bool
				}{
					{number("-2", 0), number("1", 0), row.less}, {number("1", 0), number("100", -2), row.equal},
					{number("3", 0), number("2", 0), row.greater}, {number("-0", 3), number("0", -2), row.equal},
					{decimalref.Decimal{Kind: "infinity", Negative: true}, number("1", 0), row.less},
					{decimalref.Decimal{Kind: "infinity"}, number("1", 0), row.greater},
					{decimalref.Decimal{Kind: "infinity"}, decimalref.Decimal{Kind: "infinity"}, row.equal},
				} {
					c := makeCase(t, width, row.op, mode, pair.a, pair.b)
					r := evaluate(t, c)
					want := "false"
					if pair.want {
						want = "true"
					}
					if r.Kind != "bool" || r.Width != 0 || r.Flags != 0 || r.Value != want {
						t.Fatalf("%+v: %+v want=%s", c, r, want)
					}
				}
				for _, snan := range []bool{false, true} {
					for slot := 0; slot < 2; slot++ {
						c := makeCase(t, width, row.op, mode, number("1", 0), number("2", 0))
						c.Operands[slot] = specialRaw(t, width, "nan", true, snan)
						r := evaluate(t, c)
						want := "false"
						if row.unordered {
							want = "true"
						}
						flags := uint32(0)
						if snan {
							flags = 1
						}
						if r.Value != want || r.Flags != flags {
							t.Fatalf("%+v: %+v want=%s flags=%d", c, r, want, flags)
						}
					}
				}
			}
		}
	}
}

func TestMinMaxAndNoncanonicalInputs(t *testing.T) {
	for _, width := range []int{32, 64, 128} {
		for _, mode := range modes {
			for _, op := range []string{"minnum", "maxnum"} {
				want := number("-2", 0)
				if op == "maxnum" {
					want = number("1", 0)
				}
				checkDecimal(t, makeCase(t, width, op, mode, number("-2", 0), number("1", 0)), want, 0)
				wantInf := decimalref.Decimal{Kind: "infinity", Negative: op == "minnum"}
				checkDecimal(t, makeCase(t, width, op, mode, decimalref.Decimal{Kind: "infinity"}, decimalref.Decimal{Kind: "infinity", Negative: true}), wantInf, 0)
				c := makeCase(t, width, op, mode, number("-0", 3), number("0", -2))
				r := evaluate(t, c)
				if !r.AnyZeroSign || r.Quantum {
					t.Fatalf("invented zero tie guarantee: %+v", r)
				}
				for _, raw := range c.Operands {
					if err := Compare(c, r, Observation{Kind: "decimal", Width: width, Value: raw, HasFlags: true}); err != nil {
						t.Fatal(err)
					}
				}
				c = makeCase(t, width, op, mode, number("1", 0), number("100", -2))
				r = evaluate(t, c)
				if r.Quantum {
					t.Fatal("invented equivalent cohort tie guarantee")
				}
				for _, raw := range c.Operands {
					if err := Compare(c, r, Observation{Kind: "decimal", Width: width, Value: raw, HasFlags: true}); err != nil {
						t.Fatal(err)
					}
				}
			}
			for _, raw := range map[int][]string{32: {"6cb89680", "ecb89680"}, 64: {"6c7386f26fc10000", "ec7386f26fc10000"}, 128: {"6c10000000000000:0000000000000000", "ec10000000000000:0000000000000000", "3041ed09bead87c0:378d8e6400000000"}}[width] {
				negative := raw[0] == 'e'
				prefix := ""
				if negative {
					prefix = "-"
				}
				c := makeCase(t, width, "rem", mode, number("0", 0), number("2", -2))
				c.Operands[0] = raw
				checkDecimal(t, c, number(prefix+"0", -2), 0)
				c.Op = "minnum"
				checkDecimal(t, c, number(prefix+"0", 0), 0)
			}
		}
	}
}
