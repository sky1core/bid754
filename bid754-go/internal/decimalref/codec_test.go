package decimalref

import (
	"math/big"
	"strings"
	"testing"
)

func TestCodecKnownBits(t *testing.T) {
	for _, tc := range []struct {
		width int
		raw   string
		want  Decimal
	}{
		{32, "32800001", finite("1", 0)},
		{64, "31c0000000000001", finite("1", 0)},
		{128, "3040000000000000:0000000000000001", finite("1", 0)},
		{32, "00000001", finite("1", -101)},
		{64, "0000000000000001", finite("1", -398)},
		{128, "0000000000000000:0000000000000001", finite("1", -6176)},
		{32, "6cb89680", finite("0", 0)},
		{64, "6c7386f26fc10000", finite("0", 0)},
		{128, "6c10000000000000:0000000000000000", finite("0", 0)},
	} {
		d, err := Decode(tc.width, tc.raw)
		if err != nil {
			t.Fatal(err)
		}
		check(t, Result{Value: d, Rounding: "exact"}, tc.want, 0, "exact")
		canonical, err := Encode(tc.width, tc.want)
		if err != nil {
			t.Fatal(err)
		}
		if tc.want.Coeff.Sign() != 0 && canonical != tc.raw {
			t.Fatalf("got %s want %s", canonical, tc.raw)
		}
		if tc.want.Coeff.Sign() == 0 && Compare(tc.width, Result{Value: tc.want}, tc.raw, 0) == nil {
			t.Fatal("accepted noncanonical actual")
		}
	}
}

func TestCodecBoundariesAndClasses(t *testing.T) {
	for _, width := range []int{32, 64, 128} {
		p, _ := ParametersFor(width)
		for _, exp := range []int{p.MinExp, 0, p.MaxExp} {
			for _, coeff := range []*big.Int{big.NewInt(0), big.NewInt(1), new(big.Int).Sub(power10(p.Precision), big.NewInt(1))} {
				for _, neg := range []bool{false, true} {
					d := Decimal{Kind: "finite", Negative: neg, Coeff: coeff, Exp: exp}
					raw, err := Encode(width, d)
					if err != nil {
						t.Fatal(err)
					}
					decoded, err := Decode(width, raw)
					if err != nil {
						t.Fatal(err)
					}
					check(t, Result{Value: decoded, Rounding: "exact"}, d, 0, "exact")
					if err := Compare(width, Result{Value: d}, raw, 0); err != nil {
						t.Fatal(err)
					}
				}
			}
		}
		for _, tc := range []struct {
			d     Decimal
			class string
		}{
			{finite("1", p.MinExp), "subnormal"},
			{Decimal{Kind: "finite", Coeff: power10(p.Precision - 1), Exp: p.MinExp}, "normal"},
			{finite("0", 0), "positive_zero"},
			{Decimal{Kind: "finite", Negative: true, Coeff: big.NewInt(0)}, "negative_zero"},
			{Decimal{Kind: "infinity"}, "infinity"},
			{Decimal{Kind: "nan"}, "nan"},
		} {
			got, err := Classify(width, tc.d)
			if err != nil || got != tc.class {
				t.Fatalf("class=%s err=%v want=%s", got, err, tc.class)
			}
		}
	}
}

func TestCompareContract(t *testing.T) {
	want := Result{Value: finite("1", 0), Flags: flagInexact}
	raw, _ := Encode(32, finite("100", -2))
	if err := Compare(32, want, raw, flagInexact); err != nil {
		t.Fatal(err)
	}
	if Compare(32, want, raw, flagInexact|2) == nil {
		t.Fatal("masked raw flags")
	}
	negativeZero := Decimal{Kind: "finite", Negative: true, Coeff: big.NewInt(0)}
	raw, _ = Encode(32, negativeZero)
	if Compare(32, Result{Value: finite("0", 0)}, raw, 0) == nil {
		t.Fatal("ignored zero sign")
	}
	nan := Result{Value: Decimal{Kind: "nan"}, Flags: flagInvalid}
	for _, raw := range []string{"7c000000", "fc000001"} {
		if err := Compare(32, nan, raw, flagInvalid); err != nil {
			t.Fatal(err)
		}
	}
	for _, raw := range []string{"7e000000", "7c100000", "7c0f4240", "78000000", "00000000"} {
		if Compare(32, nan, raw, flagInvalid) == nil {
			t.Fatalf("accepted %s as canonical quiet NaN", raw)
		}
	}
}

func TestCompareQuantumCohort(t *testing.T) {
	for _, width := range []int{32, 64, 128} {
		for _, tc := range []struct {
			want, other Decimal
		}{
			{finite("1", 0), finite("100", -2)},
			{finite("50", -2), finite("5", -1)},
			{finite("0", 0), finite("0", -2)},
			{Decimal{Kind: "finite", Negative: true, Coeff: big.NewInt(0), Exp: 0}, Decimal{Kind: "finite", Negative: true, Coeff: big.NewInt(0), Exp: 3}},
		} {
			want := Result{Value: tc.want}
			same, err := Encode(width, tc.want)
			if err != nil {
				t.Fatal(err)
			}
			other, err := Encode(width, tc.other)
			if err != nil {
				t.Fatal(err)
			}
			if exactValue(tc.want).Cmp(exactValue(tc.other)) != 0 {
				t.Fatalf("test cohorts differ in value: %+v vs %+v", tc.want, tc.other)
			}
			if err := Compare(width, want, other, 0); err != nil {
				t.Fatalf("numeric Compare rejected equal value: %v", err)
			}
			if CompareQuantum(width, want, other, 0) == nil {
				t.Fatalf("CompareQuantum accepted %sE%d for want %sE%d", tc.other.Coeff, tc.other.Exp, tc.want.Coeff, tc.want.Exp)
			}
			if err := CompareQuantum(width, want, same, 0); err != nil {
				t.Fatalf("CompareQuantum rejected exact cohort: %v", err)
			}
		}
	}
	nan := Result{Value: Decimal{Kind: "nan"}, Flags: flagInvalid}
	if err := CompareQuantum(32, nan, "7c000000", flagInvalid); err != nil {
		t.Fatalf("CompareQuantum changed canonical quiet NaN judgment: %v", err)
	}
	if CompareQuantum(32, nan, "7e000000", flagInvalid) == nil {
		t.Fatal("CompareQuantum accepted signaling NaN")
	}
}

func TestRejectInputs(t *testing.T) {
	for _, raw := range []string{"", "0", "3280000A", " 32800001", "0x32800001", "3280000g", strings.Repeat("0", 100000)} {
		if _, err := Decode(32, raw); err == nil {
			t.Fatal("accepted malformed raw")
		}
	}
	if _, err := Decode(128, "3040000000000000_0000000000000001"); err == nil {
		t.Fatal("accepted separator")
	}
	for _, width := range []int{0, 16, 256} {
		if _, err := ParametersFor(width); err == nil {
			t.Fatal("accepted width")
		}
		if _, err := Decode(width, "00000000"); err == nil {
			t.Fatal("accepted width")
		}
		if _, err := Encode(width, finite("1", 0)); err == nil {
			t.Fatal("accepted width")
		}
	}
	for _, d := range []Decimal{finite("1", -102), finite("1", 91), finite("10000000", 0), {Kind: "finite"}, {Kind: "finite", Coeff: big.NewInt(-1)}, {Kind: "bogus"}, {Kind: "nan", Exp: 1}, {Kind: "infinity", Coeff: big.NewInt(1)}} {
		if _, err := Encode(32, d); err == nil {
			t.Fatalf("accepted %+v", d)
		}
		if _, err := Classify(32, d); err == nil {
			t.Fatalf("classified %+v", d)
		}
	}
	for _, c := range []Case{
		{Width: 16}, {Width: 32, Mode: "bad"}, {Width: 32, Mode: "nearest_even", Op: "sqrt"},
		{Width: 32, Mode: "nearest_even", Op: "add"},
		{Width: 32, Mode: "nearest_even", Op: "add", Operands: []string{"7c000000", "32800001"}},
		{Width: 32, Mode: "nearest_even", Op: "add", Operands: []string{"78000000", "32800001"}},
		{Width: 32, Mode: "nearest_even", Op: "add", Operands: []string{"bad", "32800001"}},
	} {
		if _, err := Evaluate(c); err == nil {
			t.Fatalf("accepted %+v", c)
		}
	}
}
