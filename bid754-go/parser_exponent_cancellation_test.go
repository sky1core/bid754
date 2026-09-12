package bid754

import (
	"fmt"
	"strings"
	"testing"

	bidgo "github.com/sky1core/bid754/bid754-go/internal/bidgo"
)

type widthSpec struct {
	name      string
	precision int
	minQ      int
	maxQ      int
	bias      int
}

var (
	specD32  = widthSpec{"d32", 7, -101, 90, 101}
	specD64  = widthSpec{"d64", 16, -398, 369, 398}
	specD128 = widthSpec{"d128", 34, -6176, 6111, 6176}
	allSpecs = []widthSpec{specD32, specD64, specD128}
)

func refEncodeCohort(w widthSpec, coeff uint64, neg bool, q int) (hi, lo uint64) {
	switch w.name {
	case "d32":
		lo = (uint64(q+w.bias) << 23) | coeff
		if neg {
			lo |= 1 << 31
		}
	case "d64":
		lo = (uint64(q+w.bias) << 53) | coeff
		if neg {
			lo |= 1 << 63
		}
	case "d128":
		hi = uint64(q+w.bias) << 49
		lo = coeff
		if neg {
			hi |= 1 << 63
		}
	}
	return
}

func representable(w widthSpec, coeffDigits, q int) bool {
	return coeffDigits <= w.precision && q >= w.minQ && q <= w.maxQ
}

func rawParse(w widthSpec, s string, mode int) (hi, lo uint64, flags uint32) {
	switch w.name {
	case "d32":
		b, f := bidgo.Bid32FromStringRaw(s, mode)
		return 0, uint64(b), f
	case "d64":
		b, f := bidgo.Bid64FromString(s, mode)
		return 0, b, f
	default:
		r, f := bidgo.Bid128FromString(s, mode)
		rh, rl := bidgo.Bid128Words(r)
		return rh, rl, f
	}
}

func publicWithFlags(w widthSpec, s string) (hi, lo uint64, flags ExceptionFlags, err error) {
	switch w.name {
	case "d32":
		v, f, e := NewDecimal32WithFlags(s)
		return 0, uint64(v.ToUint32()), f, e
	case "d64":
		v, f, e := NewDecimal64WithFlags(s)
		return 0, v.ToUint64(), f, e
	default:
		v, f, e := NewDecimal128WithFlags(s)
		vh, vl := decimal128BIDWords(v)
		return vh, vl, f, e
	}
}

func publicWithMode(w widthSpec, s string, mode RoundingMode) (hi, lo uint64, flags ExceptionFlags, err error) {
	switch w.name {
	case "d32":
		v, f, e := NewDecimal32WithMode(s, mode)
		return 0, uint64(v.ToUint32()), f, e
	case "d64":
		v, f, e := NewDecimal64WithMode(s, mode)
		return 0, v.ToUint64(), f, e
	default:
		v, f, e := NewDecimal128WithMode(s, mode)
		vh, vl := decimal128BIDWords(v)
		return vh, vl, f, e
	}
}

func publicExact(w widthSpec, s string, direct bool) (hi, lo uint64, err error) {
	switch w.name {
	case "d32":
		parse := NewDecimal32
		if direct {
			parse = NewDecimal32BIDDirect
		}
		v, err := parse(s)
		return 0, uint64(v.ToUint32()), err
	case "d64":
		parse := NewDecimal64
		if direct {
			parse = NewDecimal64BIDDirect
		}
		v, err := parse(s)
		return 0, v.ToUint64(), err
	default:
		parse := NewDecimal128
		if direct {
			parse = NewDecimal128BIDDirect
		}
		v, err := parse(s)
		hi, lo := decimal128BIDWords(v)
		return hi, lo, err
	}
}

func publicReject(w widthSpec, s string) bool {
	_, _, err := publicExact(w, s, false)
	return err != nil
}

var rawRoundingModes = []struct {
	name string
	mode int
}{
	{"nearest_even", bidgo.BID_ROUNDING_TO_NEAREST},
	{"toward_negative", bidgo.BID_ROUNDING_DOWN},
	{"toward_positive", bidgo.BID_ROUNDING_UP},
	{"toward_zero", bidgo.BID_ROUNDING_TO_ZERO},
	{"ties_away", bidgo.BID_ROUNDING_TIES_AWAY},
}

var publicRoundingModes = []struct {
	name string
	mode RoundingMode
}{
	{"nearest_even", RoundNearestEven},
	{"nearest_away", RoundNearestAway},
	{"toward_zero", RoundTowardZero},
	{"toward_positive", RoundTowardPositive},
	{"toward_negative", RoundTowardNegative},
}

type expCase struct {
	id          string
	neg         bool
	coeff       uint64
	coeffDigits int
	q           int
	big         bool
	build       func() string
}

func repeatZeros(n int) string {
	if n <= 0 {
		return ""
	}
	return strings.Repeat("0", n)
}

func fractionLiteral(neg bool, F, lz, E int) string {
	var b strings.Builder
	b.Grow(F + lz + 24)
	if neg {
		b.WriteByte('-')
	}
	b.WriteString("0.")
	b.WriteString(repeatZeros(F - 1))
	b.WriteString("1e")
	if E < 0 {
		b.WriteByte('-')
		b.WriteString(repeatZeros(lz))
		fmt.Fprintf(&b, "%d", -E)
	} else {
		b.WriteString(repeatZeros(lz))
		fmt.Fprintf(&b, "%d", E)
	}
	return b.String()
}

func coefficientLiteral(neg bool, K, lz, E int) string {
	var b strings.Builder
	b.Grow(K + lz + 24)
	if neg {
		b.WriteByte('-')
	}
	b.WriteByte('1')
	b.WriteString(repeatZeros(K))
	b.WriteString("e")
	if E < 0 {
		b.WriteByte('-')
		b.WriteString(repeatZeros(lz))
		fmt.Fprintf(&b, "%d", -E)
	} else {
		b.WriteString(repeatZeros(lz))
		fmt.Fprintf(&b, "%d", E)
	}
	return b.String()
}

func fracCase(id string, neg bool, F, lz, E int, big bool) expCase {
	return expCase{
		id: id, neg: neg, coeff: 1, coeffDigits: 1, q: E - F, big: big,
		build: func() string { return fractionLiteral(neg, F, lz, E) },
	}
}

func coeffRejectCase(id string, neg bool, K, lz, E int, big bool) expCase {
	return expCase{
		id: id, neg: neg, coeff: 0, coeffDigits: K + 1, q: E, big: big,
		build: func() string { return coefficientLiteral(neg, K, lz, E) },
	}
}

func literalCase(id string, neg bool, lit string, coeff uint64, coeffDigits, q int) expCase {
	return expCase{
		id: id, neg: neg, coeff: coeff, coeffDigits: coeffDigits, q: q,
		build: func() string { return lit },
	}
}

func parserExponentCancellationCases() []expCase {
	return []expCase{
		fracCase("ctl_q0", false, 5, 0, 5, false),
		fracCase("ctl_q0_neg", true, 5, 0, 5, false),
		fracCase("ctl_q90_d32max", false, 5, 0, 95, false),
		fracCase("ctl_qm101_d32min", false, 105, 0, 4, false),
		fracCase("ctl_q369_d64max", false, 5, 0, 374, false),
		fracCase("ctl_qm398_d64min", false, 400, 0, 2, false),
		fracCase("ctl_q6111_d128max", false, 5, 0, 6116, false),
		fracCase("ctl_qm6176_d128min", false, 6178, 0, 2, false),
		literalCase("ctl_coeff100_q-2", false, "100e-2", 100, 3, -2),
		literalCase("ctl_coeff10000_q-4_neg", true, "-10000e-4", 10000, 5, -4),

		fracCase("rej_q91_d32over", false, 5, 0, 96, false),
		fracCase("rej_q6112_d128over", false, 5, 0, 6117, false),
		fracCase("rej_qm6177_d128under", false, 6179, 0, 2, false),
		coeffRejectCase("rej_coeff35", false, 34, 0, -2, false),
		coeffRejectCase("rej_coeff41_val1", false, 40, 0, -40, false),
		coeffRejectCase("rej_coeff41_val1_capexp", false, 40, 0, -40000000, true),

		fracCase("lz0_q0_7dig", false, 1000000, 0, 1000000, true),
		fracCase("lz1_q0_7dig_confirmed", false, 1000000, 1, 1000000, true),
		fracCase("lz2_q0_7dig", false, 1000000, 2, 1000000, true),
		fracCase("lz3_q0_7dig", false, 1000000, 3, 1000000, true),

		fracCase("d128max_lz1_q6111", false, 993889, 1, 1000000, true),
		fracCase("d128min_lz1_qm6176", false, 1006176, 1, 1000000, true),

		fracCase("big_n9999999", false, 9999999, 0, 9999999, true),
		fracCase("big_n10485759", false, 10485759, 0, 10485759, true),
		fracCase("lz1_q0_6dig", false, 999999, 1, 999999, true),
		fracCase("big_n10000000", false, 10000000, 0, 10000000, true),
		fracCase("big_n10485760", false, 10485760, 0, 10485760, true),
		fracCase("big_n10485761", false, 10485761, 0, 10485761, true),
		fracCase("big_n10485760_neg", true, 10485760, 1, 10485760, true),
		fracCase("lz1_q0_7dig_neg", true, 1000000, 1, 1000000, true),
		{id: "big_explicit_plus", coeff: 1, coeffDigits: 1, q: 0, big: true,
			build: func() string { return strings.Replace(fractionLiteral(false, 10485760, 1, 10485760), "e", "e+", 1) }},
	}
}

func TestParserExponentCancellationReferenceEncoder(t *testing.T) {
	if _, lo := refEncodeCohort(specD32, 1, false, 0); lo != uint64(One32BID().ToUint32()) {
		t.Errorf("refEncodeCohort d32 (1,0) = %#x, want One32 %#x", lo, One32BID().ToUint32())
	}
	if _, lo := refEncodeCohort(specD64, 1, false, 0); lo != One64BID().ToUint64() {
		t.Errorf("refEncodeCohort d64 (1,0) = %#x, want One64 %#x", lo, One64BID().ToUint64())
	}
	oneHi, oneLo := decimal128BIDWords(One128BID())
	if hi, lo := refEncodeCohort(specD128, 1, false, 0); hi != oneHi || lo != oneLo {
		t.Errorf("refEncodeCohort d128 (1,0) = %#x,%#x, want One128 %#x,%#x", hi, lo, oneHi, oneLo)
	}
	for _, tc := range []struct {
		w   widthSpec
		s   string
		q   int
		neg bool
	}{
		{specD64, "1e369", 369, false},
		{specD64, "1e-398", -398, false},
		{specD128, "-1e6111", 6111, true},
	} {
		wantHi, wantLo, _ := rawParse(tc.w, tc.s, bidgo.BID_ROUNDING_TO_NEAREST)
		gotHi, gotLo := refEncodeCohort(tc.w, 1, tc.neg, tc.q)
		if gotHi != wantHi || gotLo != wantLo {
			t.Errorf("%s refEncodeCohort(1,%v,%d)=%#x,%#x but short parse %q=%#x,%#x", tc.w.name, tc.neg, tc.q, gotHi, gotLo, tc.s, wantHi, wantLo)
		}
	}
}

func TestParserExponentCancellation(t *testing.T) {
	cache := map[string]string{}
	get := func(c expCase) string {
		s, ok := cache[c.id]
		if !ok {
			s = c.build()
			cache[c.id] = s
		}
		return s
	}

	for _, c := range parserExponentCancellationCases() {
		c := c
		if c.big && testing.Short() {
			continue
		}
		for _, w := range allSpecs {
			w := w
			t.Run(w.name+"/"+c.id, func(t *testing.T) {
				s := get(c)
				repr := representable(w, c.coeffDigits, c.q)
				if repr {
					wantHi, wantLo := refEncodeCohort(w, c.coeff, c.neg, c.q)
					for _, rm := range rawRoundingModes {
						hi, lo, fl := rawParse(w, s, rm.mode)
						if hi != wantHi || lo != wantLo || fl != 0 {
							t.Errorf("raw %s q=%d: got %#x,%#x fl=%#x; want exact cohort %#x,%#x fl=0", rm.name, c.q, hi, lo, fl, wantHi, wantLo)
						}
					}
					for _, pm := range publicRoundingModes {
						hi, lo, fl, err := publicWithMode(w, s, pm.mode)
						if err != nil {
							t.Errorf("public WithMode %s: unexpected error %v (representable cohort must be accepted)", pm.name, err)
							continue
						}
						if hi != wantHi || lo != wantLo {
							t.Errorf("public WithMode %s: got %#x,%#x want %#x,%#x", pm.name, hi, lo, wantHi, wantLo)
						}
						if fl != 0 {
							t.Errorf("public WithMode %s: flags %v, want none (exact)", pm.name, fl)
						}
					}

					for _, direct := range []bool{false, true} {
						hi, lo, err := publicExact(w, s, direct)
						if err != nil || hi != wantHi || lo != wantLo {
							t.Errorf("public exact direct=%v: got %#x,%#x err=%v; want exact cohort %#x,%#x", direct, hi, lo, err, wantHi, wantLo)
						}
					}
					hi, lo, fl, err := publicWithFlags(w, s)
					if err != nil || hi != wantHi || lo != wantLo || fl != 0 {
						t.Errorf("public WithFlags: got %#x,%#x flags=%v err=%v; want exact cohort %#x,%#x flags=0", hi, lo, fl, err, wantHi, wantLo)
					}
				} else {
					if !publicReject(w, s) {
						t.Errorf("error-only constructor accepted an unrepresentable written cohort (coeffDigits=%d, q=%d)", c.coeffDigits, c.q)
					}
				}
			})
		}
	}
}

func TestParserExponentCancellationSignSymmetry(t *testing.T) {
	for _, c := range parserExponentCancellationCases() {
		if c.big || c.neg {
			continue
		}
		pos := c.build()
		neg := "-" + pos
		for _, w := range allSpecs {
			ph, pl, pf := rawParse(w, pos, bidgo.BID_ROUNDING_TO_NEAREST)
			nh, nl, nf := rawParse(w, neg, bidgo.BID_ROUNDING_TO_NEAREST)
			var wantHi, wantLo uint64
			if w.name == "d128" {
				wantHi, wantLo = ph|(1<<63), pl
			} else if w.name == "d32" {
				wantHi, wantLo = ph, pl|(1<<31)
			} else {
				wantHi, wantLo = ph, pl|(1<<63)
			}
			if nh != wantHi || nl != wantLo || nf != pf {
				t.Errorf("%s/%s raw sign symmetry: + %#x,%#x fl=%#x ; - %#x,%#x fl=%#x", w.name, c.id, ph, pl, pf, nh, nl, nf)
			}
		}
	}
}
