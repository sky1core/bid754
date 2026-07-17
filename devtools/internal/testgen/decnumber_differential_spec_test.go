package testgen

// Generator-side unit tests for the decNumber differential gate codegen:
// corpus filters, exclusion predicates, stream determinism, and
// known-divergence row validation. Pure functions only — the pin-time
// sentinel oracle path is exercised by the full generation run and the
// anchors test, not here.

import (
	"math/big"
	"strings"
	"testing"
)

func TestDecnumberDiffBoundaryFilterClosedWorld(t *testing.T) {
	for _, width := range decnumberDiffWidths {
		filter := decnumberDiffFilterBoundary(width)
		if got := len(filter.included) + filter.excludedNonCanonical + filter.excludedNaNPayload; got != filter.source {
			t.Errorf("width %d: boundary filter is not closed-world: %d included + %d noncanonical + %d nan-payload != %d source",
				width.bits, len(filter.included), filter.excludedNonCanonical, filter.excludedNaNPayload, filter.source)
		}
		for _, class := range []string{"zero", "finite", "inf", "qnan", "snan"} {
			if filter.includedKindCoverage[class] == 0 {
				t.Errorf("width %d: filtered boundary corpus lost operand class %q", width.bits, class)
			}
		}
		seen := map[bid128BidCodecValue]bool{}
		for _, value := range filter.included {
			if !decnumberDiffCanonical(width, value) {
				t.Errorf("width %d: included boundary value %016x:%016x is not canonical", width.bits, value.hi, value.lo)
			}
			if isNaN, payloaded := decnumberDiffIsNaN(width, value); isNaN && payloaded {
				t.Errorf("width %d: included boundary value %016x:%016x carries a NaN payload", width.bits, value.hi, value.lo)
			}
			if seen[value] {
				t.Errorf("width %d: boundary value %016x:%016x is duplicated", width.bits, value.hi, value.lo)
			}
			seen[value] = true
		}
	}
}

func TestDecnumberDiffProbeNormalization(t *testing.T) {
	wantCounts := map[int]int{32: 12, 64: 12, 128: 11}
	for _, width := range decnumberDiffWidths {
		probes, zeroed, dropped, err := decnumberDiffProbes(width)
		if err != nil {
			t.Fatalf("width %d probes: %v", width.bits, err)
		}
		if len(probes) != wantCounts[width.bits] {
			t.Errorf("width %d: %d probes, want %d", width.bits, len(probes), wantCounts[width.bits])
		}
		if zeroed != 2 {
			t.Errorf("width %d: %d NaN probes payload-zeroed, want 2 (the Tier 1 payload-1 qNaN and sNaN probes)", width.bits, zeroed)
		}
		wantDropped := 0
		if width.bits == 128 {
			wantDropped = 1 // the BID128 steered probe has no canonical counterpart
		}
		if dropped != wantDropped {
			t.Errorf("width %d: %d probes dropped as non-canonical, want %d", width.bits, dropped, wantDropped)
		}
		for _, value := range probes {
			if !decnumberDiffCanonical(width, value) {
				t.Errorf("width %d: probe %016x:%016x is not canonical", width.bits, value.hi, value.lo)
			}
			if isNaN, payloaded := decnumberDiffIsNaN(width, value); isNaN && payloaded {
				t.Errorf("width %d: probe %016x:%016x carries a NaN payload", width.bits, value.hi, value.lo)
			}
		}
	}
}

func TestDecnumberDiffFmaExclusionPredicate(t *testing.T) {
	zeroPos := decnumberDiffTriple{kind: decnumberDiffFinite, coeff: new(big.Int), exp: 5}
	zeroNeg := decnumberDiffTriple{kind: decnumberDiffFinite, sign: true, coeff: new(big.Int), exp: -7}
	infPos := decnumberDiffTriple{kind: decnumberDiffInf}
	infNeg := decnumberDiffTriple{kind: decnumberDiffInf, sign: true}
	qnan := decnumberDiffTriple{kind: decnumberDiffQNaN}
	snan := decnumberDiffTriple{kind: decnumberDiffSNaN}
	one := decnumberDiffTriple{kind: decnumberDiffFinite, coeff: big.NewInt(1), exp: 0}

	excluded := [][3]decnumberDiffTriple{
		{zeroPos, infPos, qnan},
		{zeroNeg, infNeg, qnan},
		{infPos, zeroPos, qnan},
		{infNeg, zeroNeg, qnan},
	}
	for i, tuple := range excluded {
		if !decnumberDiffFmaExcluded(tuple[0], tuple[1], tuple[2]) {
			t.Errorf("excluded[%d]: L13 predicate must match zero x inf with a quiet NaN addend of any quantum/sign", i)
		}
	}
	included := [][3]decnumberDiffTriple{
		{zeroPos, infPos, snan}, // sNaN addend: both sides signal invalid, measured agreement
		{zeroPos, infPos, one},  // finite addend: both sides signal invalid, measured agreement
		{zeroPos, zeroPos, qnan},
		{infPos, infPos, qnan},
		{one, infPos, qnan},
		{qnan, infPos, qnan}, // NaN product operand quiets on both sides without invalid
	}
	for i, tuple := range included {
		if decnumberDiffFmaExcluded(tuple[0], tuple[1], tuple[2]) {
			t.Errorf("included[%d]: L13 predicate must not widen beyond the measured implementation-defined class", i)
		}
	}
}

func TestDecnumberDiffExactProductCorpus(t *testing.T) {
	offsets := []int{-1, 0, 1}
	for _, width := range decnumberDiffWidths {
		values, err := decnumberDiffExactProductValues(width, offsets)
		if err != nil {
			t.Fatalf("width %d exact-product corpus: %v", width.bits, err)
		}
		if len(values) != 2*len(offsets)+4 {
			t.Fatalf("width %d: %d exact-product values, want %d", width.bits, len(values), 2*len(offsets)+4)
		}
		onePowP, err := decnumberDiffEncode(width, decnumberDiffTriple{kind: decnumberDiffFinite, coeff: big.NewInt(1), exp: int32(width.p)})
		if err != nil {
			t.Fatalf("width %d: encode 1E<p>: %v", width.bits, err)
		}
		found := false
		for _, value := range values {
			if value == onePowP {
				found = true
			}
			if !decnumberDiffCanonical(width, value) {
				t.Errorf("width %d: exact-product value %016x:%016x is not canonical", width.bits, value.hi, value.lo)
			}
		}
		if !found {
			t.Errorf("width %d: exact-product corpus is missing 1E<p>, the D2-class trigger scale value", width.bits)
		}
	}
	// The Bid128Fma overflow defect trigger (1E34) must be a member for
	// decimal128 so the D2-class input executes on every full run.
	width128 := decnumberDiffWidths[2]
	values, err := decnumberDiffExactProductValues(width128, offsets)
	if err != nil {
		t.Fatalf("decimal128 exact-product corpus: %v", err)
	}
	d2Trigger := bid128BidCodecValue{lo: 0x0000000000000001, hi: 0x3084000000000000}
	found := false
	for _, value := range values {
		if value == d2Trigger {
			found = true
		}
	}
	if !found {
		t.Errorf("decimal128 exact-product corpus is missing the 1E34 D2 trigger value %016x:%016x", d2Trigger.hi, d2Trigger.lo)
	}
}

func TestDecnumberDiffRandomStreamDeterminism(t *testing.T) {
	width := decnumberDiffWidths[0]
	first, err := decnumberDiffRandomStream(width, 512, nil)
	if err != nil {
		t.Fatalf("random stream: %v", err)
	}
	second, err := decnumberDiffRandomStream(width, 512, nil)
	if err != nil {
		t.Fatalf("random stream: %v", err)
	}
	if first.streamHash != second.streamHash {
		t.Errorf("random stream digest is not deterministic: %d vs %d", first.streamHash, second.streamHash)
	}
	if first.streamHash == 0 {
		t.Errorf("random stream digest is zero; the mix is not folding operand words")
	}
}

func TestDecnumberDiffKnownRowValidation(t *testing.T) {
	width32 := decnumberDiffWidths[0]
	valid := DecnumberDifferentialKnownDivergence{
		ReasonID:       "test_reason",
		Classification: "decnumber_defect",
		Width:          "decimal32",
		Operation:      "fma",
		Mode:           0,
		X:              "3e8265b3",
		Y:              "1600cdcf",
		Z:              "a180005c",
		Intel:          "23fe4df8/00000020",
		Decnumber:      "8277497E-30/00000020",
		Note:           "test row",
	}
	rows, err := decnumberDiffParseKnownRows(width32, DecnumberDifferentialSpec{KnownDivergences: []DecnumberDifferentialKnownDivergence{valid}})
	if err != nil {
		t.Fatalf("valid known-divergence row rejected: %v", err)
	}
	if len(rows) != 1 || rows[0].arity != 3 {
		t.Fatalf("valid known-divergence row parsed incorrectly: %+v", rows)
	}

	agree := valid
	agree.Decnumber = "8277496E-30/00000020" // equals the decoded Intel pin: not a divergence
	if _, err := decnumberDiffParseKnownRows(width32, DecnumberDifferentialSpec{KnownDivergences: []DecnumberDifferentialKnownDivergence{agree}}); err == nil {
		t.Errorf("a known-divergence row whose pinned legs agree must be rejected as stale")
	} else if !strings.Contains(err.Error(), "stale") {
		t.Errorf("agreeing row rejection carries the wrong reason: %v", err)
	}

	excluded := valid
	excluded.X = "32800000" // +0E0
	excluded.Y = "78000000" // +Inf
	excluded.Z = "7c000000" // qNaN
	excluded.Intel = "7c000000/00000000"
	excluded.Decnumber = "qNaN/01"
	if _, err := decnumberDiffParseKnownRows(width32, DecnumberDifferentialSpec{KnownDivergences: []DecnumberDifferentialKnownDivergence{excluded}}); err == nil {
		t.Errorf("a known-divergence row inside the excluded L13 class must be rejected (it never executes)")
	}

	nonCanonical := valid
	nonCanonical.X = "7c800000" // non-canonical NaN encoding
	if _, err := decnumberDiffParseKnownRows(width32, DecnumberDifferentialSpec{KnownDivergences: []DecnumberDifferentialKnownDivergence{nonCanonical}}); err == nil {
		t.Errorf("a known-divergence row with a non-canonical operand must be rejected")
	}

	wrongClass := valid
	wrongClass.Classification = "intel_defect"
	if _, err := decnumberDiffParseKnownRows(width32, DecnumberDifferentialSpec{KnownDivergences: []DecnumberDifferentialKnownDivergence{wrongClass}}); err == nil {
		t.Errorf("a known-divergence row with a non-decnumber_defect classification must be rejected (IEEE deviations take the documented separate procedure)")
	}
}

func TestDecnumberDiffKnownRowStaleWhenUnmatched(t *testing.T) {
	spec := DecnumberDifferentialSpec{
		InventoryOutput:             "unused.json",
		RandomPairsPerOperation:     64, // far too small to reach the pinned random triple
		ExactProductExponentOffsets: []int{-1, 0, 1},
		KnownDivergences: []DecnumberDifferentialKnownDivergence{{
			ReasonID:       "test_reason",
			Classification: "decnumber_defect",
			Width:          "decimal32",
			Operation:      "fma",
			Mode:           0,
			X:              "3e8265b3",
			Y:              "1600cdcf",
			Z:              "a180005c",
			Intel:          "23fe4df8/00000020",
			Decnumber:      "8277497E-30/00000020",
			Note:           "test row",
		}},
	}
	if _, err := decnumberDiffBuildWidthPlan(decnumberDiffWidths[0], spec); err == nil {
		t.Errorf("a known-divergence row matching no executed case must fail generation as stale")
	} else if !strings.Contains(err.Error(), "stale") {
		t.Errorf("unmatched row rejection carries the wrong reason: %v", err)
	}
}
