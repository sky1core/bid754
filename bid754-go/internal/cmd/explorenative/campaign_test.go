package main

import (
	"github.com/sky1core/bid754/bid754-go/internal/decimalprobe"
	"strings"
	"testing"

	"github.com/sky1core/bid754/bid754-go/internal/decimalref"
)

const (
	rawOne = "32800001" // Decimal32 1E0
	rawTwo = "32800002" // Decimal32 2E0
)

func sampleLine(extra string) string {
	return `{"type":"sample","sample":{"version":1,"family":"uniform-finite","case":{"width":32,"op":"add","mode":"nearest_even","operands":["` +
		rawOne + `","` + rawTwo + `"]}}` + extra + `}`
}

func TestDecodeReplayLineAcceptsSample(t *testing.T) {
	rec, err := decodeReplayLine([]byte(sampleLine("")))
	if err != nil {
		t.Fatalf("valid sample rejected: %v", err)
	}
	samples, err := rec.replaySamples()
	if err != nil || len(samples) != 1 {
		t.Fatalf("replaySamples = %d %v", len(samples), err)
	}
	if err := samples[0].validate(); err != nil {
		t.Fatalf("valid sample failed validate: %v", err)
	}
}

func TestDecodeReplayLineRejectsUnknownFields(t *testing.T) {
	cases := map[string]string{
		"unknown top-level":   `{"type":"sample","bogus":1,"sample":{"version":1,"family":"uniform-finite","case":{"width":32,"op":"add","mode":"nearest_even","operands":["` + rawOne + `","` + rawTwo + `"]}}}`,
		"unknown nested":      `{"type":"sample","sample":{"version":1,"family":"uniform-finite","case":{"width":32,"op":"add","mode":"nearest_even","operands":["` + rawOne + `","` + rawTwo + `"],"extra":9}}}`,
		"unknown record type": `{"type":"counters","generated":3}`,
		"missing type":        `{"sample":{"version":1}}`,
		"truncated":           `{"type":"sample","sample":`,
	}
	for name, line := range cases {
		if _, err := decodeReplayLine([]byte(line)); err == nil {
			t.Errorf("%s: expected rejection, got nil", name)
		}
	}
}

func TestReplaySamplesRejectsAmbiguous(t *testing.T) {
	sampleWithOriginal := sampleLine(`,"original":{"version":1,"family":"uniform-finite","case":{"width":32,"op":"add","mode":"nearest_even","operands":["` + rawOne + `","` + rawTwo + `"]}}`)
	rec, err := decodeReplayLine([]byte(sampleWithOriginal))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if _, err := rec.replaySamples(); err == nil {
		t.Error("sample record carrying original must be rejected")
	}

	findingNoOriginal := `{"type":"finding","shrunk":{"version":1,"family":"uniform-finite","case":{"width":32,"op":"add","mode":"nearest_even","operands":["` + rawOne + `","` + rawTwo + `"]}}}`
	rec, err = decodeReplayLine([]byte(findingNoOriginal))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if _, err := rec.replaySamples(); err == nil {
		t.Error("finding record without original must be rejected")
	}
}

func TestFindingReplaysOriginalAndShrunk(t *testing.T) {
	line := `{"type":"finding","campaign":"relations","op":"add","width":32,"mode":"nearest_even",` +
		`"original":{"version":1,"family":"uniform-finite","case":{"width":32,"op":"add","mode":"nearest_even","operands":["` + rawOne + `","` + rawTwo + `"]}},` +
		`"shrunk":{"version":1,"family":"uniform-finite","case":{"width":32,"op":"add","mode":"nearest_even","operands":["` + rawOne + `","` + rawOne + `"]}}}`
	rec, err := decodeReplayLine([]byte(line))
	if err != nil {
		t.Fatalf("finding rejected: %v", err)
	}
	samples, err := rec.replaySamples()
	if err != nil || len(samples) != 2 {
		t.Fatalf("replaySamples = %d %v", len(samples), err)
	}
	for i, s := range samples {
		if err := s.validate(); err != nil {
			t.Fatalf("sample %d invalid: %v", i, err)
		}
	}
}

func TestSampleValidateRejectsUnsupported(t *testing.T) {
	mk := func(op, mode string, ops []string) *sampleRec {
		return &sampleRec{Version: 1, Family: "uniform-finite", Case: decimalref.Case{Width: 32, Op: op, Mode: mode, Operands: ops}}
	}
	tests := map[string]*sampleRec{
		"sqrt op":       mk("sqrt", "nearest_even", []string{rawOne}),
		"bad mode":      mk("add", "toward_odd", []string{rawOne, rawTwo}),
		"wrong arity":   mk("add", "nearest_even", []string{rawOne}),
		"nonfinite":     mk("add", "nearest_even", []string{"78000000", rawTwo}),
		"malformed raw": mk("add", "nearest_even", []string{"zz", rawTwo}),
	}
	for name, s := range tests {
		if err := s.validate(); err == nil {
			t.Errorf("%s: expected validate error, got nil", name)
		}
	}
	if err := mk("fma", "nearest_even", []string{rawOne, rawTwo, rawOne}).validate(); err != nil {
		t.Errorf("valid fma sample rejected: %v", err)
	}
}

func TestRefLegDiffIsolatesValueAndFlags(t *testing.T) {
	res, err := decimalref.Evaluate(decimalref.Case{Width: 32, Op: "add", Mode: "nearest_even", Operands: []string{rawOne, rawTwo}})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := decimalref.Encode(32, res.Value) // canonical 3E0
	if err != nil {
		t.Fatal(err)
	}
	if v, f := refLegDiff(32, res, raw, res.Flags); v || f {
		t.Fatalf("matching leg reported divergence: value=%v flags=%v", v, f)
	}
	if v, f := refLegDiff(32, res, raw, res.Flags|flagInexact); v || !f {
		t.Fatalf("flag-only divergence misclassified: value=%v flags=%v", v, f)
	}
	if v, f := refLegDiff(32, res, rawTwo, res.Flags); !v || f {
		t.Fatalf("value-only divergence misclassified: value=%v flags=%v", v, f)
	}
}

func TestCounterReconcile(t *testing.T) {
	ops := make([]decimalref.Decimal, 2)
	for i, raw := range []string{rawOne, rawTwo} {
		d, err := decimalref.Decode(32, raw)
		if err != nil {
			t.Fatal(err)
		}
		ops[i] = d
	}
	res, err := decimalref.Evaluate(decimalref.Case{Width: 32, Op: "add", Mode: "nearest_even", Operands: []string{rawOne, rawTwo}})
	if err != nil {
		t.Fatal(err)
	}

	cs := newCounterSet("add")
	cs.Generated = 3
	cs.Reached = 3
	cs.OracleError = 1
	for i := 0; i < 2; i++ {
		if err := cs.addOracle(32, res, ops); err != nil {
			t.Fatal(err)
		}
	}
	if err := cs.reconcile(); err != nil {
		t.Fatalf("reconcile should pass: %v", err)
	}

	cs.OracleError = 2 // now reached(3) != completed(2)+error(2)
	if err := cs.reconcile(); err == nil || !strings.Contains(err.Error(), "reached") {
		t.Fatalf("dropped-count reconcile must fail, got %v", err)
	}
}

func TestReplayContractNegatives(t *testing.T) {
	valid := sampleLine("")
	for name, line := range map[string]string{
		"version":              strings.Replace(valid, `"version":1`, `"version":999`, 1),
		"empty family":         strings.Replace(valid, `"uniform-finite"`, `""`, 1),
		"false relation":       strings.Replace(valid, `"uniform-finite"`, `"add_below"`, 1),
		"duplicate field":      strings.Replace(valid, `"version":1`, `"version":999,"version":1`, 1),
		"case alias field":     strings.Replace(valid, `"version":1`, `"version":999,"Version":1`, 1),
		"unicode alias field":  strings.Replace(valid, `"version":1`, `"version":999,"verſion":1`, 1),
		"noncanonical":         strings.Replace(valid, rawOne, "3280001", 1),
		"trailing":             valid + `{}`,
		"truncated":            valid[:len(valid)-1],
		"empty":                " \n",
		"conflicting metadata": strings.Replace(valid, `"type":"sample"`, `"type":"sample","op":"sub"`, 1),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := readReplay(strings.NewReader(line)); err == nil {
				t.Fatal("accepted invalid replay")
			}
		})
	}
	if got, err := readReplay(strings.NewReader(valid + "\n" + valid)); err != nil || len(got) != 2 {
		t.Fatalf("valid replay: %d %v", len(got), err)
	}
}

func TestRefLegUnknownFlagsAreOnlyFlagDifference(t *testing.T) {
	r, err := decimalref.Evaluate(decimalref.Case{Width: 32, Op: "add", Mode: "nearest_even", Operands: []string{rawOne, rawTwo}})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := decimalref.Encode(32, r.Value)
	if err != nil {
		t.Fatal(err)
	}
	for _, flags := range []uint32{2, 0x80000000, 0xffffffff} {
		if value, flag := refLegDiff(32, r, raw, flags); value || !flag {
			t.Fatalf("flags %08x: value=%v flag=%v", flags, value, flag)
		}
	}
}

func TestSameClassesRequiresExactSet(t *testing.T) {
	if !sameClasses([]string{"a", "b"}, []string{"b", "a"}) || sameClasses([]string{"a", "b"}, []string{"a"}) || sameClasses([]string{"a"}, []string{"b"}) {
		t.Fatal("incorrect discrepancy class matching")
	}
}

func TestRelationOpsFilter(t *testing.T) {
	ops, families, err := referenceTargets(campaignRel, "sub")
	if err != nil {
		t.Fatal(err)
	}
	if len(ops) != 1 || len(families) == 0 {
		t.Fatalf("empty target selection: %v %v", ops, families)
	}
	for _, fam := range families {
		op, err := decimalprobe.FamilyOperation(fam)
		if err != nil || op != "sub" {
			t.Fatalf("unselected operation: %s %s %v", fam, op, err)
		}
	}
	for _, text := range []string{"sqrt", "add,sqrt", "add,add", "", "add,"} {
		if _, _, err := referenceTargets(campaignRel, text); err == nil {
			t.Fatalf("accepted %q", text)
		}
	}
}
