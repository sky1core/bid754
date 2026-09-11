package bid754

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/sky1core/bid754/bid754-go/internal/decimalref"
	"github.com/sky1core/bid754/bid754-go/internal/tier1ref"
)

type propOut struct {
	kind  string
	dec   decimalref.Decimal
	text  string
	width int
	flags uint32
}

func propGather(t *testing.T, c tier1ref.Case) map[string]propOut {
	t.Helper()
	want, err := tier1ref.Evaluate(c)
	if err != nil {
		t.Fatalf("reference %+v: %v", c, err)
	}
	obs, err := tier1PublicPaths(c)
	if err != nil {
		t.Fatalf("paths %+v: %v", c, err)
	}
	out := map[string]propOut{}
	ref := propOut{kind: want.Kind, width: want.Width, flags: want.Flags}
	if want.Kind == "decimal" {
		ref.dec = want.Decimal
	} else {
		ref.text = want.Value
	}
	out["ref"] = ref
	paths := map[string]bool{}
	for _, path := range tier1ExpectedPaths(c, "go") {
		paths[path] = true
	}
	for _, o := range obs {
		if !paths[o.Path] {
			t.Fatalf("unexpected or duplicate path %s for %+v", o.Path, c)
		}
		delete(paths, o.Path)
		if !o.HasFlags {
			t.Fatalf("%s omitted flags for %+v", o.Path, c)
		}
		if _, dup := out[o.Path]; dup {
			t.Fatalf("duplicate path %s for %+v", o.Path, c)
		}
		p := propOut{kind: o.Kind, width: o.Width, flags: o.Flags}
		if o.Kind == "decimal" {
			d, err := decimalref.Decode(o.Width, o.Value)
			if err != nil {
				t.Fatalf("%s decode %q: %v", o.Path, o.Value, err)
			}
			p.dec = d
		} else {
			p.text = o.Value
		}
		out[o.Path] = p
	}
	if len(paths) != 0 {
		t.Fatalf("missing paths %v for %+v", paths, c)
	}
	return out
}

func propDecCheck(a, b propOut, negate bool) error {
	if a.kind != "decimal" || b.kind != "decimal" {
		return fmt.Errorf("kind %s/%s not decimal", a.kind, b.kind)
	}
	if a.width != b.width {
		return fmt.Errorf("width %d != %d", a.width, b.width)
	}
	if a.flags != b.flags {
		return fmt.Errorf("flags %#x != %#x", a.flags, b.flags)
	}
	if a.dec.Kind != b.dec.Kind {
		return fmt.Errorf("class %s != %s", a.dec.Kind, b.dec.Kind)
	}
	switch a.dec.Kind {
	case "finite":
		na := a.dec.Negative != negate
		if na != b.dec.Negative {
			return fmt.Errorf("sign %v != %v", na, b.dec.Negative)
		}
		if a.dec.Coeff.Cmp(b.dec.Coeff) != 0 {
			return fmt.Errorf("coeff %s != %s", a.dec.Coeff, b.dec.Coeff)
		}
		if a.dec.Exp != b.dec.Exp {
			return fmt.Errorf("quantum %d != %d", a.dec.Exp, b.dec.Exp)
		}
	case "infinity":
		na := a.dec.Negative != negate
		if na != b.dec.Negative {
			return fmt.Errorf("infinity sign %v != %v", na, b.dec.Negative)
		}
	case "nan":
	}
	return nil
}

func propRelate(t *testing.T, label string, mA, mB map[string]propOut, check func(a, b propOut) error) int {
	t.Helper()
	if len(mA) != len(mB) {
		t.Fatalf("%s: path counts %d != %d", label, len(mA), len(mB))
	}
	count := 0
	for path, a := range mA {
		b, ok := mB[path]
		if !ok {
			t.Fatalf("%s: path %s absent in paired case", label, path)
		}
		if err := check(a, b); err != nil {
			t.Fatalf("%s [%s]: %v", label, path, err)
		}
		count++
	}
	return count
}

func propFinite(t *testing.T, width int, negative bool, coeff int64, exp int) string {
	t.Helper()
	raw, err := decimalref.Encode(width, decimalref.Decimal{Kind: "finite", Negative: negative, Coeff: big.NewInt(coeff), Exp: exp})
	if err != nil {
		t.Fatalf("encode %d/%d/%dE%d: %v", width, boolInt(negative), coeff, exp, err)
	}
	return raw
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func propReencode(t *testing.T, width int, d decimalref.Decimal) string {
	t.Helper()
	raw, err := decimalref.Encode(width, d)
	if err != nil {
		t.Fatalf("re-encode %d %+v: %v", width, d, err)
	}
	return raw
}

var propWidths = []int{32, 64, 128}

func propLawRemainderSigns(t *testing.T) int {
	pairs := []struct {
		cx int64
		ex int
		cy int64
		ey int
	}{
		{7, 0, 2, 0}, {75, -1, 2, 0}, {123, -2, 7, -1}, {6, 0, 2, 0}, {0, 0, 3, 0},
	}
	comparisons := 0
	for _, width := range propWidths {
		for _, op := range []string{"rem", "fmod"} {
			for _, p := range pairs {
				base := tier1ref.Case{Width: width, Op: op, Mode: "nearest_even",
					Operands: []string{propFinite(t, width, false, p.cx, p.ex), propFinite(t, width, false, p.cy, p.ey)}}
				flipDivisor := base
				flipDivisor.Operands = []string{base.Operands[0], propFinite(t, width, true, p.cy, p.ey)}
				flipDividend := base
				flipDividend.Operands = []string{propFinite(t, width, true, p.cx, p.ex), base.Operands[1]}

				mBase := propGather(t, base)
				comparisons += propRelate(t, fmt.Sprintf("divisor-sign %d/%s/%+v", width, op, p), mBase, propGather(t, flipDivisor),
					func(a, b propOut) error { return propDecCheck(a, b, false) })
				comparisons += propRelate(t, fmt.Sprintf("dividend-sign %d/%s/%+v", width, op, p), mBase, propGather(t, flipDividend),
					func(a, b propOut) error { return propDecCheck(a, b, true) })
			}
		}
	}
	return comparisons
}

func propLawScaleBReversible(t *testing.T) int {
	comparisons := 0
	for _, width := range propWidths {
		p, _ := decimalref.ParametersFor(width)
		for _, coeff := range []int64{1, 123, 987654} {
			for _, exp := range []int{0, 5, -7} {
				for _, n := range []int{1, 4, -3} {
					fwd := tier1ref.Case{Width: width, Op: "scaleb", Mode: "nearest_even",
						Operands: []string{propFinite(t, width, false, coeff, exp)}, Param: fmt.Sprint(n)}
					mFwd := propGather(t, fwd)
					for path, o := range mFwd {
						if o.flags != 0 || o.dec.Kind != "finite" || o.dec.Negative ||
							o.dec.Coeff.Cmp(big.NewInt(coeff)) != 0 || o.dec.Exp != exp+n {
							t.Fatalf("scaleB shift %d/%dE%d n=%d [%s]: %+v", width, coeff, exp, n, path, o)
						}
						comparisons++
					}
					back := tier1ref.Case{Width: width, Op: "scaleb", Mode: "nearest_even",
						Operands: []string{propReencode(t, width, mFwd["ref"].dec)}, Param: fmt.Sprint(-n)}
					for path, o := range propGather(t, back) {
						if o.flags != 0 || o.dec.Kind != "finite" || o.dec.Negative ||
							o.dec.Coeff.Cmp(big.NewInt(coeff)) != 0 || o.dec.Exp != exp {
							t.Fatalf("scaleB round trip %d/%dE%d n=%d [%s]: %+v", width, coeff, exp, n, path, o)
						}
						comparisons++
					}
				}
			}
		}
		for _, exp := range []int{0, p.MinExp + 3, p.MaxExp - 3} {
			for _, n := range []int64{1, -1, 9223372036854775807, -9223372036854775808} {
				c := tier1ref.Case{Width: width, Op: "scaleb", Mode: "nearest_even",
					Operands: []string{propFinite(t, width, false, 0, exp)}, Param: fmt.Sprint(n)}
				sum := new(big.Int).Add(big.NewInt(int64(exp)), big.NewInt(n))
				wantExp := clampExp(sum, p.MinExp, p.MaxExp)
				for path, o := range propGather(t, c) {
					if o.flags != 0 || o.dec.Kind != "finite" || o.dec.Coeff.Sign() != 0 || o.dec.Exp != wantExp {
						t.Fatalf("scaleB zero quantum %d exp=%d n=%d [%s]: %+v want exp %d", width, exp, n, path, o, wantExp)
					}
					comparisons++
				}
			}
		}
	}
	return comparisons
}

func clampExp(v *big.Int, lo, hi int) int {
	if v.Cmp(big.NewInt(int64(lo))) < 0 {
		return lo
	}
	if v.Cmp(big.NewInt(int64(hi))) > 0 {
		return hi
	}
	return int(v.Int64())
}

func propLawWidthRoundTrip(t *testing.T) int {
	comparisons := 0
	type conv struct{ from, to int }
	for _, cv := range []conv{{32, 64}, {32, 128}, {64, 128}} {
		for _, coeff := range []int64{1, 12300, 987654} {
			for _, exp := range []int{-4, 0, 7} {
				for _, negative := range []bool{false, true} {
					widen := tier1ref.Case{Width: cv.from, Op: "convert", Mode: "nearest_even",
						Operands: []string{propFinite(t, cv.from, negative, coeff, exp)}, Target: cv.to}
					mWiden := propGather(t, widen)
					for path, o := range mWiden {
						if o.flags != 0 || o.dec.Kind != "finite" || o.dec.Negative != negative ||
							o.dec.Coeff.Cmp(big.NewInt(coeff)) != 0 || o.dec.Exp != exp {
							t.Fatalf("widen %d->%d %dE%d neg=%v [%s]: %+v", cv.from, cv.to, coeff, exp, negative, path, o)
						}
						comparisons++
					}
					narrow := tier1ref.Case{Width: cv.to, Op: "convert", Mode: "nearest_even",
						Operands: []string{propReencode(t, cv.to, mWiden["ref"].dec)}, Target: cv.from}
					for path, o := range propGather(t, narrow) {
						if o.kind != "decimal" || o.flags != 0 || o.dec.Kind != "finite" || o.width != cv.from ||
							o.dec.Negative != negative || o.dec.Coeff.Cmp(big.NewInt(coeff)) != 0 || o.dec.Exp != exp {
							t.Fatalf("narrow-back %d->%d->%d %dE%d neg=%v [%s]: %+v", cv.from, cv.to, cv.from, coeff, exp, negative, path, o)
						}
						comparisons++
					}
				}
			}
		}
	}
	return comparisons
}

func propLawQuietSymmetry(t *testing.T) int {
	comparisons := 0
	pairs := []struct {
		ca int64
		ea int
		cb int64
		eb int
	}{
		{2, 0, 1, 0}, {1, 0, 2, 0}, {1, 0, 100, -2}, {3, 0, 2, 0}, {5, -1, 5, -1},
	}
	boolEqual := func(a, b propOut) error {
		if a.kind != "bool" || b.kind != "bool" {
			return fmt.Errorf("kind %s/%s not bool", a.kind, b.kind)
		}
		if a.flags != 0 || b.flags != 0 {
			return fmt.Errorf("flags %#x/%#x", a.flags, b.flags)
		}
		if a.text != b.text {
			return fmt.Errorf("value %s != %s", a.text, b.text)
		}
		return nil
	}
	boolComplement := func(a, b propOut) error {
		if a.kind != "bool" || b.kind != "bool" {
			return fmt.Errorf("kind %s/%s not bool", a.kind, b.kind)
		}
		if a.flags != 0 || b.flags != 0 {
			return fmt.Errorf("flags %#x/%#x", a.flags, b.flags)
		}
		if a.text == b.text {
			return fmt.Errorf("equal/not-equal share value %s", a.text)
		}
		return nil
	}
	for _, width := range propWidths {
		for _, p := range pairs {
			aRaw := propFinite(t, width, false, p.ca, p.ea)
			bRaw := propFinite(t, width, false, p.cb, p.eb)
			mk := func(op string, x, y string) tier1ref.Case {
				return tier1ref.Case{Width: width, Op: op, Mode: "nearest_even", Operands: []string{x, y}}
			}
			comparisons += propRelate(t, fmt.Sprintf("swap-lt/gt %d/%+v", width, p),
				propGather(t, mk("quiet_less", aRaw, bRaw)), propGather(t, mk("quiet_greater", bRaw, aRaw)), boolEqual)
			comparisons += propRelate(t, fmt.Sprintf("swap-le/ge %d/%+v", width, p),
				propGather(t, mk("quiet_less_equal", aRaw, bRaw)), propGather(t, mk("quiet_greater_equal", bRaw, aRaw)), boolEqual)
			comparisons += propRelate(t, fmt.Sprintf("complement-eq %d/%+v", width, p),
				propGather(t, mk("quiet_equal", aRaw, bRaw)), propGather(t, mk("quiet_not_equal", aRaw, bRaw)), boolComplement)
		}
		nan, err := decimalref.Encode(width, decimalref.Decimal{Kind: "nan"})
		if err != nil {
			t.Fatal(err)
		}
		finite := propFinite(t, width, false, 1, 0)
		for _, tv := range []struct {
			op   string
			want string
		}{
			{"quiet_ordered", "false"}, {"quiet_unordered", "true"}, {"quiet_equal", "false"},
			{"quiet_not_equal", "true"}, {"quiet_less", "false"}, {"quiet_greater", "false"},
		} {
			for _, slot := range [][2]string{{nan, finite}, {finite, nan}} {
				m := propGather(t, tier1ref.Case{Width: width, Op: tv.op, Mode: "nearest_even", Operands: []string{slot[0], slot[1]}})
				for path, o := range m {
					if o.kind != "bool" || o.flags != 0 || o.text != tv.want {
						t.Fatalf("NaN %d/%s [%s]: %+v want %s", width, tv.op, path, o, tv.want)
					}
					comparisons++
				}
			}
		}
	}
	return comparisons
}

func propLawSpecialFlags(t *testing.T) int {
	comparisons := 0
	specialDecimal := func(label string, c tier1ref.Case, class string, flags uint32) {
		for path, o := range propGather(t, c) {
			if o.kind != "decimal" || o.dec.Kind != class || o.flags != flags {
				t.Fatalf("%s [%s]: %+v want %s flags %#x", label, path, o, class, flags)
			}
			comparisons++
		}
	}
	for _, width := range propWidths {
		p, _ := decimalref.ParametersFor(width)
		zero := propFinite(t, width, false, 0, 0)
		one := propFinite(t, width, false, 1, 0)
		inf, err := decimalref.Encode(width, decimalref.Decimal{Kind: "infinity"})
		if err != nil {
			t.Fatal(err)
		}
		for _, op := range []string{"rem", "fmod"} {
			specialDecimal(fmt.Sprintf("%s-by-zero %d", op, width),
				tier1ref.Case{Width: width, Op: op, Mode: "nearest_even", Operands: []string{one, zero}}, "nan", tier1ref.Invalid)
			specialDecimal(fmt.Sprintf("%s-inf-dividend %d", op, width),
				tier1ref.Case{Width: width, Op: op, Mode: "nearest_even", Operands: []string{inf, one}}, "nan", tier1ref.Invalid)
		}
		maxFinite := propReencode(t, width, decimalref.Decimal{Kind: "finite",
			Coeff: new(big.Int).Sub(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(p.Precision)), nil), big.NewInt(1)), Exp: p.MaxExp})
		specialDecimal(fmt.Sprintf("scaleb-overflow %d", width),
			tier1ref.Case{Width: width, Op: "scaleb", Mode: "nearest_even", Operands: []string{maxFinite}, Param: "9223372036854775807"},
			"infinity", tier1ref.Overflow|tier1ref.Inexact)

		nan, err := decimalref.Encode(width, decimalref.Decimal{Kind: "nan"})
		if err != nil {
			t.Fatal(err)
		}
		c := tier1ref.Case{Width: width, Op: "to_int", Mode: "nearest_even", Operands: []string{nan}, Target: 32}
		for path, o := range propGather(t, c) {
			if o.kind != "integer" || o.flags != tier1ref.Invalid || o.text != "-2147483648" {
				t.Fatalf("to_int(NaN) %d [%s]: %+v", width, path, o)
			}
			comparisons++
		}
	}
	return comparisons
}

func TestTier1MetamorphicProperties(t *testing.T) {
	counts := map[string]int{
		"remainder-signs":   propLawRemainderSigns(t),
		"scaleb-reversible": propLawScaleBReversible(t),
		"width-roundtrip":   propLawWidthRoundTrip(t),
		"quiet-symmetry":    propLawQuietSymmetry(t),
		"special-flags":     propLawSpecialFlags(t),
	}
	want := map[string]int{
		"remainder-signs":   180,
		"scaleb-reversible": 792,
		"width-roundtrip":   324,
		"quiet-symmetry":    243,
		"special-flags":     57,
	}
	total := 0
	for law, n := range counts {
		total += n
		if n != want[law] {
			t.Fatalf("law %s comparisons=%d, pinned %d", law, n, want[law])
		}
	}
	if total != 1596 {
		t.Fatalf("total comparisons=%d, pinned 1596", total)
	}
	t.Logf("TIER1-PROPERTIES comparisons=%d laws=%v", total, counts)
}
