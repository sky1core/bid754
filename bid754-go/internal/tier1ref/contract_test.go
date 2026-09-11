package tier1ref

import (
	"strings"
	"testing"

	"github.com/sky1core/bid754/bid754-go/internal/decimalref"
)

func TestValidationRejectsMalformedCases(t *testing.T) {
	base := makeCase(t, 32, "rem", "nearest_even", number("7", 0), number("2", 0))
	for _, mutate := range []func(*Case){
		func(c *Case) { c.Width = 16 }, func(c *Case) { c.Mode = "" }, func(c *Case) { c.Mode = "nearest" },
		func(c *Case) { c.Op = "add" }, func(c *Case) { c.Op = "quiet_unknown" },
		func(c *Case) { c.Param = "0" }, func(c *Case) { c.Target = 32 },
		func(c *Case) { c.Operands = c.Operands[:1] }, func(c *Case) { c.Operands = append(c.Operands, "32800003") },
		func(c *Case) { c.Operands[0] = "3280000A" }, func(c *Case) { c.Operands[0] = "0x32800007" },
		func(c *Case) { c.Operands[0] = strings.Repeat("1", 100000) },
		func(c *Case) {
			c.Width = 128
			c.Operands = []string{"3040000000000000_0000000000000001", "3040000000000000:0000000000000001"}
		},
	} {
		c := base
		c.Operands = append([]string(nil), base.Operands...)
		mutate(&c)
		if err := Validate(c); err == nil {
			t.Fatalf("validated %+v", c)
		}
		if _, err := Evaluate(c); err == nil {
			t.Fatal("Evaluate bypasses Validate")
		}
	}
	for _, param := range []string{"", "-0", "+1", "01", "-01", " 1", "1 ", "1e3", "9223372036854775808", "-9223372036854775809", strings.Repeat("9", 100000)} {
		c := makeCase(t, 32, "scaleb", "nearest_even", number("1", 0))
		c.Param = param
		if err := Validate(c); err == nil {
			t.Fatalf("accepted scaleb parameter len=%d", len(param))
		}
	}
	for _, row := range []struct {
		op     string
		target int
		param  string
	}{
		{"from_int", 32, "2147483648"}, {"from_int", 32, "-2147483649"}, {"from_int", 64, "9223372036854775808"},
		{"from_int", 64, "-9223372036854775809"}, {"from_uint", 32, "4294967296"}, {"from_uint", 64, "18446744073709551616"},
		{"from_uint", 64, "-1"}, {"from_int", 8, "1"}, {"from_uint", 128, "1"}, {"from_int", 32, "+1"},
	} {
		if err := Validate(Case{Width: 32, Op: row.op, Mode: "nearest_even", Target: row.target, Param: row.param}); err == nil {
			t.Fatalf("accepted %+v", row)
		}
	}
	for _, op := range []string{"convert", "to_int", "to_int_exact", "to_uint", "to_uint_exact"} {
		for _, target := range []int{0, 7, 31, 65, 256} {
			c := makeCase(t, 32, op, "nearest_even", number("1", 0))
			c.Target = target
			if err := Validate(c); err == nil {
				t.Fatalf("accepted %+v", c)
			}
		}
	}
	c := makeCase(t, 32, "convert", "nearest_even", number("1", 0))
	c.Target = 32
	if err := Validate(c); err == nil {
		t.Fatal("accepted same-width convert")
	}
}

func TestComparatorDetectsValueFlagsSignAndQuantum(t *testing.T) {
	c := makeCase(t, 32, "scaleb", "nearest_even", number("15", -101))
	c.Param = "-1"
	r := evaluate(t, c)
	raw, _ := decimalref.Encode(32, number("2", -101))
	baseline := Observation{Path: "test", Kind: "decimal", Width: 32, Value: raw, Flags: 0x30, HasFlags: true}
	if err := Compare(c, r, baseline); err != nil {
		t.Fatal(err)
	}
	for _, mutate := range []func(*Observation){
		func(o *Observation) { o.Flags = 0 }, func(o *Observation) { o.Flags = 0x20 }, func(o *Observation) { o.Flags |= 2 },
		func(o *Observation) { o.HasFlags = false }, func(o *Observation) { o.Width = 64 }, func(o *Observation) { o.Kind = "integer" },
		func(o *Observation) { o.Value = "00000001" }, func(o *Observation) { o.Value = "80000002" },
		func(o *Observation) { o.Value = "7c000000" }, func(o *Observation) { o.Value = "0000000A" },
	} {
		o := baseline
		mutate(&o)
		if Compare(c, r, o) == nil {
			t.Fatalf("accepted corrupted observation %+v", o)
		}
	}
	baseline.HasFlags = false
	baseline.Flags = 0
	if err := Compare(c, r, baseline); err != nil {
		t.Fatalf("value-only observation: %v", err)
	}
	c = makeCase(t, 32, "scaleb", "nearest_even", number("1", 0))
	c.Param = "0"
	r = evaluate(t, c)
	raw, _ = decimalref.Encode(32, number("10", -1))
	if Compare(c, r, Observation{Kind: "decimal", Width: 32, Value: raw}) == nil {
		t.Fatal("ignored prescribed quantum")
	}
	c = makeCase(t, 32, "scaleb", "nearest_even", number("-0", 0))
	c.Param = "0"
	r = evaluate(t, c)
	for _, raw := range []string{"32800000", "ecb89680"} {
		if Compare(c, r, Observation{Kind: "decimal", Width: 32, Value: raw}) == nil {
			t.Fatal("ignored zero sign or noncanonical result")
		}
	}
	c = makeCase(t, 32, "minnum", "nearest_even", decimalref.Decimal{Kind: "infinity"}, decimalref.Decimal{Kind: "infinity"})
	r = evaluate(t, c)
	for _, raw := range []string{"78000001", "f8000000"} {
		if Compare(c, r, Observation{Kind: "decimal", Width: 32, Value: raw}) == nil {
			t.Fatal("ignored infinity sign/canonicality")
		}
	}
}

func TestComparatorNaNGuarantee(t *testing.T) {
	for _, width := range []int{32, 64, 128} {
		c := makeCase(t, width, "rem", "nearest_even", number("1", 0), number("0", 0))
		r := evaluate(t, c)
		for _, raw := range map[int][]string{32: {"7c000001", "fc000009"}, 64: {"7c00000000000001", "fc00000000000009"}, 128: {"7c00000000000000:0000000000000001", "fc00000000000000:0000000000000009"}}[width] {
			if err := Compare(c, r, Observation{Kind: "decimal", Width: width, Value: raw, HasFlags: true, Flags: 1}); err != nil {
				t.Fatalf("invented NaN sign/payload guarantee: %v", err)
			}
		}
		for _, raw := range map[int][]string{32: {"7e000000", "7c100000", "7c0f4240"}, 64: {"7e00000000000000", "7c04000000000000", "7c038d7ea4c68000"}, 128: {"7e00000000000000:0000000000000000", "7c00400000000000:0000000000000000", "7c00314dc6448d93:38c15b0a00000000"}}[width] {
			if Compare(c, r, Observation{Kind: "decimal", Width: width, Value: raw, HasFlags: true, Flags: 1}) == nil {
				t.Fatalf("accepted noncanonical/signaling NaN %s", raw)
			}
		}
	}
}

func TestComparatorIntegerAndBool(t *testing.T) {
	c := makeCase(t, 32, "to_int", "nearest_even", number("12", 0))
	c.Target = 8
	r := evaluate(t, c)
	for _, text := range []string{"012", "+12", "12.0", "-0", "13", "128", strings.Repeat("1", 100000)} {
		if Compare(c, r, Observation{Kind: "integer", Width: 8, Value: text}) == nil {
			t.Fatal("accepted incorrect integer")
		}
	}
	if err := Compare(c, r, Observation{Kind: "integer", Width: 8, Value: "12"}); err != nil {
		t.Fatal(err)
	}
	c = makeCase(t, 32, "quiet_equal", "nearest_even", number("1", 0), number("10", -1))
	r = evaluate(t, c)
	for _, text := range []string{"1", "True", "false", " true"} {
		if Compare(c, r, Observation{Kind: "bool", Value: text}) == nil {
			t.Fatal("accepted incorrect bool")
		}
	}
	if err := Compare(c, r, Observation{Kind: "bool", Value: "true"}); err != nil {
		t.Fatal(err)
	}
}

func TestComparatorMinMaxOperandSelection(t *testing.T) {
	for _, op := range []string{"minnum", "maxnum"} {
		for _, values := range []struct {
			inputs    []string
			forbidden string
		}{
			{[]string{"32800001", "31800064"}, "3200000a"},
			{[]string{"32800000", "b1800000"}, "31800000"},
		} {
			c := Case{Width: 32, Op: op, Mode: "nearest_even", Operands: values.inputs}
			want, err := Evaluate(c)
			if err != nil {
				t.Fatal(err)
			}
			for _, raw := range values.inputs {
				o := Observation{Kind: "decimal", Width: 32, Value: raw, HasFlags: true}
				if err := Compare(c, want, o); err != nil {
					t.Fatal(err)
				}
			}
			bad := Observation{Kind: "decimal", Width: 32, Value: values.forbidden, HasFlags: true}
			if Compare(c, want, bad) == nil {
				t.Fatalf("%s accepted non-operand cohort %s", op, bad.Value)
			}
		}
	}
}
