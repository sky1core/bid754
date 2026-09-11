package bid754

import (
	"strings"
	"testing"

	"github.com/sky1core/bid754/bid754-go/internal/tier1ref"
)

func TestTier1OracleStrength(t *testing.T) {
	cases := []tier1ref.Case{
		{Width: 32, Op: "rem", Mode: "nearest_even", Operands: []string{"32800007", "32800002"}},
		{Width: 32, Op: "quiet_less", Mode: "nearest_even", Operands: []string{"32800007", "32800002"}},
		{Width: 32, Op: "to_int_exact", Mode: "nearest_even", Operands: []string{"3200000f"}, Target: 8},
	}
	for _, c := range cases {
		want, err := tier1ref.Evaluate(c)
		if err != nil {
			t.Fatal(err)
		}
		obs, err := tier1PublicPaths(c)
		if err != nil {
			t.Fatal(err)
		}
		for _, o := range obs {
			if err := tier1ref.Compare(c, want, o); err != nil {
				t.Fatalf("%+v: %v", c, err)
			}
		}
		all := append([]tier1ref.Observation{}, obs...)
		for _, o := range obs {
			o.Path = strings.Replace(o.Path, "go/", "rust/", 1)
			all = append(all, o)
		}
		java := tier1JavaResult{Status: "ok", Kind: want.Kind, Value: want.Value}
		if want.Kind == "decimal" {
			java.Value = want.Decimal.Coeff.String()
			java.Exp = want.Decimal.Exp
			if want.Decimal.Negative {
				java.Value = "-" + java.Value
			}
		}
		if err := tier1Check(c, want, all, java); err != nil {
			t.Fatal(err)
		}
		for i := range all {
			bad := append([]tier1ref.Observation{}, all...)
			bad[i].Flags ^= 0x01
			if tier1Check(c, want, bad, java) == nil {
				t.Fatal("changed flag accepted")
			}
			bad = append([]tier1ref.Observation{}, all...)
			bad[i].HasFlags = false
			bad[i].Flags = 0
			if tier1Check(c, want, bad, java) == nil {
				t.Fatal("missing flags accepted")
			}
			bad = append([]tier1ref.Observation{}, all...)
			bad = append(bad[:i], bad[i+1:]...)
			if tier1Check(c, want, bad, java) == nil {
				t.Fatal("missing path accepted")
			}
			bad = append([]tier1ref.Observation{}, all...)
			switch bad[i].Kind {
			case "bool":
				bad[i].Value = "true"
			case "integer":
				bad[i].Value = "3"
			case "decimal":
				bad[i].Value = "32800001"
			}
			if tier1Check(c, want, bad, java) == nil {
				t.Fatal("changed result accepted")
			}
		}
		badJava := java
		badJava.Status = "excluded"
		badJava.Reason = "invalid-integer"
		if tier1Check(c, want, all, badJava) == nil {
			t.Fatal("unsupported Java exclusion accepted")
		}
	}
}

func TestTier1JavaProtocol(t *testing.T) {
	for _, line := range []string{"1\tok\tbool\ttrue\n", "1\tok\tinteger\t-9223372036854775808\n", "1\tok\tdecimal\t-100\t-3\n", "1\texcluded\texponent-range\n"} {
		if _, err := tier1DecodeJava([]byte(line)); err != nil {
			t.Fatal(err)
		}
	}
	for _, line := range []string{"1\tok\tbool\t1\n", "1\tok\tinteger\t01\n", "1\tok\tinteger\t-0\n", "1\tok\tdecimal\t1\t9999999\n", "1\texcluded\tunknown\n", "1\tok\tinteger\t1\textra\n"} {
		if _, err := tier1DecodeJava([]byte(line)); err == nil {
			t.Fatalf("accepted %q", line)
		}
	}
}
