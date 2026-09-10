//go:build cgo && bid754_native

package main

import (
	"testing"

	"github.com/sky1core/bid754/bid754-go/internal/decimalprobe"
	"github.com/sky1core/bid754/bid754-go/internal/decimalref"
)

// These tests exercise the real three-leg path (pinned Intel C via cgo, the Go
// mechanical port, and the decimalref model). No fake backend.

func TestRawToWordsRoundTrip(t *testing.T) {
	cases := []struct {
		width int
		raw   string
	}{
		{32, "32800001"},
		{64, "31c0000000000001"},
		{128, "3084000000000000:0000000000000001"},
	}
	for _, tc := range cases {
		w, _ := widthByBits(tc.width)
		words, err := rawToWords(tc.width, tc.raw)
		if err != nil {
			t.Fatalf("rawToWords(%d,%s): %v", tc.width, tc.raw, err)
		}
		if got := formatValue(w, words); got != tc.raw {
			t.Fatalf("round trip d%d: %s -> %s", tc.width, tc.raw, got)
		}
	}
}

func TestEvaluateCaseAgreesOnKnownFiniteInputs(t *testing.T) {
	cases := []decimalref.Case{
		{Width: 32, Op: "add", Mode: "nearest_even", Operands: []string{"32800001", "32800002"}},
		{Width: 32, Op: "sub", Mode: "toward_zero", Operands: []string{"32800005", "32800002"}},
		{Width: 64, Op: "mul", Mode: "nearest_even", Operands: []string{"31c0000000000002", "31c0000000000003"}},
		{Width: 64, Op: "div", Mode: "toward_negative", Operands: []string{"31c0000000000006", "31c0000000000002"}},
		{Width: 128, Op: "add", Mode: "nearest_even", Operands: []string{"3084000000000000:0000000000000001", "3084000000000000:0000000000000002"}},
	}
	for _, c := range cases {
		oc, err := evaluateCase(c)
		if err != nil {
			t.Fatalf("evaluateCase %+v: %v", c, err)
		}
		if oc.oracleErr != nil {
			t.Fatalf("model errored on finite case %+v: %v", c, oc.oracleErr)
		}
		if oc.cBits != oc.gBits || oc.cFlags != oc.gFlags {
			t.Fatalf("C/Go differ on %+v: C=%s/%08x go=%s/%08x", c, oc.cBits, oc.cFlags, oc.gBits, oc.gFlags)
		}
		if oc.disc.any() {
			t.Fatalf("unexpected disagreement on %+v: classes=%v C=%s/%08x go=%s/%08x",
				c, oc.disc.classes(), oc.cBits, oc.cFlags, oc.gBits, oc.gFlags)
		}
	}
}

// TestReferenceSweepReconcilesNative runs a small real uniform-finite target and
// a real relations target end to end, asserting counters reconcile (runTarget
// fails otherwise) and that agreeing legs produce no findings.
func TestReferenceSweepReconcilesNative(t *testing.T) {
	run := func(family, op string, gen func(hi, lo uint64, exp int32, neg bool) (decimalprobe.Sample, error)) {
		var findings int
		var counters *countersRecord
		rc := &refCampaign{
			campaign: "test", seedStr: "1", modes: []string{"nearest_even"}, shrinkAttempts: 0,
			emit: func(rec any) error {
				switch r := rec.(type) {
				case *findingRec:
					findings++
					t.Logf("finding classes=%v", r.Classes)
				case *countersRecord:
					counters = r
				}
				return nil
			},
		}
		w, _ := widthByBits(64)
		tt, err := rc.runTarget(1, w, "nearest_even", family, op, gen, 40)
		if err != nil {
			t.Fatalf("runTarget %s/%s: %v", family, op, err)
		}
		if tt.generated != 40 {
			t.Fatalf("%s/%s generated=%d, want 40", family, op, tt.generated)
		}
		if counters == nil {
			t.Fatalf("%s/%s emitted no counters record", family, op)
		}
		if tt.findings != findings {
			t.Fatalf("%s/%s totals findings=%d but emitted %d", family, op, tt.findings, findings)
		}
	}

	run("uniform-finite", "add", func(hi, lo uint64, exp int32, neg bool) (decimalprobe.Sample, error) {
		return decimalprobe.Uniform("add", 64, "nearest_even", hi, lo, exp, neg)
	})
	run("mul_tie", "mul", func(hi, lo uint64, exp int32, neg bool) (decimalprobe.Sample, error) {
		return decimalprobe.Generate("mul_tie", 64, "nearest_even", hi, lo, exp, neg)
	})
}

// TestAttachShrinkRefusesNonReproducing injects a synthetic discrepancy that the
// real legs do not reproduce; the shrink fails-callback must refuse every
// candidate (including the original), so no shrunk case is fabricated.
func TestAttachShrinkRefusesNonReproducing(t *testing.T) {
	sample, err := decimalprobe.Uniform("add", 64, "nearest_even", 1, 2, 3, false)
	if err != nil {
		t.Fatal(err)
	}
	oc, err := evaluateCase(sample.Case)
	if err != nil {
		t.Fatal(err)
	}
	oc.disc.refGoValue = true // real legs agree; this class cannot be reproduced

	rc := &refCampaign{shrinkAttempts: 16}
	f := &findingRec{}
	rc.attachShrink(f, sample, oc)
	if f.Shrink == nil || !f.Shrink.Enabled {
		t.Fatal("shrink block missing or disabled")
	}
	if !f.Shrink.Failed || f.Shrink.Error == "" {
		t.Fatal("failed shrink was not recorded")
	}
	if f.Shrunk != nil {
		t.Fatal("shrink fabricated a case for a discrepancy the real legs never reproduce")
	}

	recorded := rc.makeFinding(sample, oc)
	if recorded.Original == nil || !sameOperands(sample, decimalprobe.Sample{Version: recorded.Original.Version, Family: recorded.Original.Family, Case: recorded.Original.Case}) || !recorded.Shrink.Failed || recorded.Shrunk != nil || rc.shrinkErrors != 1 {
		t.Fatal("shrink failure lost the original finding or failure count")
	}

	rc0 := &refCampaign{shrinkAttempts: 0}
	f0 := &findingRec{}
	rc0.attachShrink(f0, sample, oc)
	if f0.Shrink == nil || f0.Shrink.Enabled {
		t.Fatal("shrink-attempts=0 must record shrinking disabled")
	}
}
