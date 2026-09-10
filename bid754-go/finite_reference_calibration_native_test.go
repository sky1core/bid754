//go:build cgo && bid754_native && bid754_decnumber_diff

package bid754

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/sky1core/bid754/bid754-go/internal/decimalprobe"
	"github.com/sky1core/bid754/bid754-go/internal/decimalref"
)

func TestFiniteReferenceDecnumberCalibration(t *testing.T) {
	anchors := loadFiniteReferenceAnchors(t)
	if version := decnumberDiffCVersion(); version != "decNumber 3.68" {
		t.Fatalf("unrecognized reference engine: %q", version)
	}
	ops := append([]decnumberDiffOp(nil), decnumberDiffRoundedOpsTable[:]...)
	ops = append(ops, decnumberDiffFmaOp)
	checked := map[int]int{}
	check := func(s decimalprobe.Sample) {
		t.Helper()
		want, err := decimalprobe.Validate(s)
		if err != nil {
			t.Fatal(err)
		}
		var op decnumberDiffOp
		found := false
		for _, candidate := range ops {
			if candidate.name == s.Case.Op {
				op, found = candidate, true
				break
			}
		}
		if !found {
			t.Fatalf("missing decNumber operation %q", s.Case.Op)
		}
		var text [3]string
		for i, raw := range s.Case.Operands {
			d, err := decimalref.Decode(s.Case.Width, raw)
			if err != nil {
				t.Fatal(err)
			}
			sign := ""
			if d.Negative {
				sign = "-"
			}
			text[i] = fmt.Sprintf("%s%sE%d", sign, d.Coeff, d.Exp)
		}
		modes := map[string]int{"nearest_even": 0, "toward_negative": 1, "toward_positive": 2, "toward_zero": 3, "nearest_away": 4}
		mode, ok := modes[s.Case.Mode]
		if !ok {
			t.Fatal("unknown mode")
		}
		dn := decnumberDiffRunDn(op.opCode, s.Case.Width, mode, text[0], text[1], text[2], op.arity)
		flags, unknown := decnumberDiffProjectStatus(dn.rawStatus)
		if dn.rc != 0 || dn.parseStatus != 0 || unknown != 0 {
			t.Fatalf("decNumber execution failed: rc=%d parse=%x unknown=%x sample=%+v", dn.rc, dn.parseStatus, unknown, s)
		}
		value := decimalref.Decimal{Negative: dn.triple.sign}
		switch dn.triple.kind {
		case decnumberDiffKindFinite:
			value.Kind, value.Coeff, value.Exp = "finite", dn.triple.coeff, int(dn.triple.exp)
		case decnumberDiffKindInf:
			value.Kind = "infinity"
		case decnumberDiffKindQNaN:
			value.Kind = "nan"
		default:
			t.Fatalf("unexpected decNumber result class %v", dn.triple.kind)
		}
		raw, err := decimalref.Encode(s.Case.Width, value)
		if err != nil {
			t.Fatalf("decNumber result cannot be encoded: %v sample=%+v", err, s)
		}
		if err := decimalref.CompareQuantum(s.Case.Width, want, raw, flags); err != nil {
			t.Fatalf("reference/decNumber disagreement: %v sample=%+v decNumber=%s/%x reference=%+v", err, s, dn.triple.key(), flags, want)
		}
		checked[s.Case.Width]++
	}
	for _, width := range []int{32, 64, 128} {
		for _, mode := range finiteModes {
			for _, family := range decimalprobe.Families() {
				for _, negative := range []bool{false, true} {
					s, err := decimalprobe.Generate(family, width, mode, 0, 754, 40, negative)
					if err != nil {
						t.Fatal(err)
					}
					check(s)
				}
			}
		}
		for _, seed := range []int64{9102026, 1112027, 314159} {
			rng := rand.New(rand.NewSource(seed))
			for _, op := range ops {
				for _, mode := range finiteModes {
					for i := 0; i < 4; i++ {
						s, err := decimalprobe.Uniform(op.name, width, mode, rng.Uint64(), rng.Uint64(), int32(rng.Uint32()), rng.Intn(2) == 1)
						if err != nil {
							t.Fatal(err)
						}
						check(s)
					}
				}
			}
		}
		if checked[width] != anchors.Decnumber[width] {
			t.Fatalf("decimal%d calibrated %d cases, want %d", width, checked[width], anchors.Decnumber[width])
		}
		t.Logf("FINITE-REFERENCE-CALIBRATION width=%d cases=%d excluded=0", width, checked[width])
	}
}
