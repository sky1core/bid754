package main

import "testing"

// TestSplitMix64ReferenceVector pins the generator to Vigna's published
// SplitMix64 outputs for initial state 0, so a silent PRNG edit cannot
// change every recorded seed's meaning unnoticed.
func TestSplitMix64ReferenceVector(t *testing.T) {
	rng := &splitMix64{state: 0}
	want := []uint64{0xe220a8397b1dcdaf, 0x6e789e6aa1b965f4, 0x06c45d188009454f}
	for i, w := range want {
		if got := rng.next(); got != w {
			t.Fatalf("splitmix64 output %d = %016x, want %016x", i, got, w)
		}
	}
}

// TestEncodeBIDKnownPatterns pins the encoders against independently known
// BID bit patterns (including the operands of the pinned Bid128Fma overflow
// regression in internal/bidgo).
func TestEncodeBIDKnownPatterns(t *testing.T) {
	tests := []struct {
		name  string
		w     widthParams
		sign  uint64
		exp   int64 // biased
		coeff words
		want  words
	}{
		{"d32_one", width32, 0, 101, u128FromUint64(1), words{lo: 0x32800001}},
		{"d32_nmax", width32, 0, 191, u128FromUint64(9999999), words{lo: 0x77f8967f}},
		{"d32_neg_zero", width32, 1, 101, u128FromUint64(0), words{lo: 0xb2800000}},
		{"d64_one", width64, 0, 398, u128FromUint64(1), words{lo: 0x31c0000000000001}},
		{"d64_nmax", width64, 0, 767, u128FromUint64(9999999999999999),
			words{lo: 0x77fb86f26fc0ffff}},
		{"d128_one_e34", width128, 0, 34 + 6176, u128FromUint64(1),
			words{hi: 0x3084000000000000, lo: 1}},
		{"d128_nmax", width128, 0, 12287, u128Sub64(pow10u128(34), 1),
			words{hi: 0x5fffed09bead87c0, lo: 0x378d8e63ffffffff}},
		{"d128_neg_nmax", width128, 1, 12287, u128Sub64(pow10u128(34), 1),
			words{hi: 0xdfffed09bead87c0, lo: 0x378d8e63ffffffff}},
	}
	for _, tc := range tests {
		got, ok := encodeBID(tc.w, tc.sign, tc.exp, tc.coeff)
		if !ok || got != tc.want {
			t.Errorf("%s: encodeBID = %016x:%016x ok=%v, want %016x:%016x",
				tc.name, got.hi, got.lo, ok, tc.want.hi, tc.want.lo)
		}
	}
}

// TestEncodeBIDRejectsOutOfRange keeps the encoders explicit about
// unencodable inputs instead of silently masking them.
func TestEncodeBIDRejectsOutOfRange(t *testing.T) {
	if _, ok := encodeBID(width32, 0, 192, u128FromUint64(1)); ok {
		t.Error("d32 exponent 192 must not encode")
	}
	if _, ok := encodeBID(width32, 0, 0, u128FromUint64(1<<23+1<<21)); ok {
		t.Error("d32 coefficient beyond the special-form range must not encode")
	}
	if _, ok := encodeBID(width64, 0, 0, u128FromUint64(1<<53+1<<51)); ok {
		t.Error("d64 coefficient beyond the special-form range must not encode")
	}
	if _, ok := encodeBID(width128, 0, 0, words{hi: 1 << 49}); ok {
		t.Error("d128 coefficient with hi >= 2^49 must not encode")
	}
}

// TestBuildPoolEntriesEncodable checks that every finite pool entry
// materialized and that its stored parameters re-encode to its stored bits
// (the invariant exponent correlation relies on).
func TestBuildPoolEntriesEncodable(t *testing.T) {
	for _, w := range []widthParams{width32, width64, width128} {
		pool := buildPool(w)
		if len(pool) < 50 {
			t.Fatalf("d%d pool has only %d entries", w.width, len(pool))
		}
		for i, e := range pool {
			if !e.finite {
				continue
			}
			bits, ok := encodeBID(w, 0, e.exp, e.coeff)
			if !ok || bits != e.bits {
				t.Fatalf("d%d pool[%d] does not re-encode: %v", w.width, i, e)
			}
		}
	}
}

// TestDrawTupleDeterministic locks stream reproducibility: the same seed
// must reproduce the same tuples, and distinct seeds must diverge.
func TestDrawTupleDeterministic(t *testing.T) {
	for _, w := range []widthParams{width32, width64, width128} {
		pool := buildPool(w)
		for _, op := range allOps {
			a := &splitMix64{state: targetSeed(754, w, op)}
			b := &splitMix64{state: targetSeed(754, w, op)}
			c := &splitMix64{state: targetSeed(755, w, op)}
			diverged := false
			for i := 0; i < 200; i++ {
				ta := drawTuple(a, w, op, 0.5, pool)
				tb := drawTuple(b, w, op, 0.5, pool)
				tc := drawTuple(c, w, op, 0.5, pool)
				if ta != tb {
					t.Fatalf("d%d/%s case %d not reproducible", w.width, op, i)
				}
				if ta != tc {
					diverged = true
				}
			}
			if !diverged {
				t.Fatalf("d%d/%s: seeds 754 and 755 drew identical streams", w.width, op)
			}
		}
	}
}

// TestDrawTupleZeroBiasIsRawRandom checks that -bias 0 never consumes the
// pool (pure uniform exploration stays available).
func TestDrawTupleZeroBiasIsRawRandom(t *testing.T) {
	pool := buildPool(width32)
	rng := &splitMix64{state: 1}
	seen := map[uint64]bool{}
	for i := 0; i < 500; i++ {
		tup := drawTuple(rng, width32, "add", 0, pool)
		seen[tup[0].lo] = true
	}
	if len(seen) < 400 {
		t.Fatalf("bias=0 draw looks non-uniform: %d distinct of 500", len(seen))
	}
}
