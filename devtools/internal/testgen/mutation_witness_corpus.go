package testgen

import (
	"fmt"
	"math/bits"
)

// Mutation-audit witness corpus (batch C-1).
//
// The 2026-07-18 mutation audit (devtools/cmd/mutgate, seed 754, per-file 20,
// arithmetic-core preset, audited at commit dbb2fa1; the bidgo sources are
// byte-identical through this tree) measured 37 mutants that survived the
// regular generated verification chain (goport readtest, goport decTest,
// public-API parity, native readtest/decTest/FFI subset, decNumber
// differential) but were killed by the Tier 1 arithmetic long Structured leg
// or by the hand-written flagless-variant equivalence gate. Each such mutant
// is a measured detection gap of the regular chain, and this file promotes
// one distinguishing input row per mutant into a regular generated domain:
//
//   - mutants on flags-exposing port paths -> extra C-FFI exact bit-compare
//     cases (ffiMutationWitnessCases), compared against pinned Intel C;
//   - mutants on the separately ported flagless ("pure") port bodies -> the
//     generated flagless-sibling equivalence leg of the public parity gate
//     (parityFlaglessSiblingTargets / parityFlaglessWitnessRows), compared
//     against the oracle-verified *WithFlags sibling.
//
// MutantID is the mutgate site ID (file:offset:category:variant) at the audit
// commit. These tables are generator-owned closed worlds: a witness row whose
// FFI function or flagless target does not resolve fails generation. No entry
// here is read from or written to devtools/verification_anchors.json; count
// changes are re-pinned there by hand.

// ffiMutationWitnessCase is one extra generated FFI bit-compare case pinned by
// a mutation-audit witness. Operands use the FFI operand string encoding
// (8/16 hex digits for BID32/BID64, 32-hex little-endian byte image for
// BID128).
type ffiMutationWitnessCase struct {
	Function string
	Rounding int
	Operands []string
	MutantID string
	Reason   string
}

// ffiMutationWitnessCases holds the witness rows for mutants on flags-exposing
// port paths. The distinguishing operands were harvested by re-applying each
// mutant and recording the first failing Tier 1 Structured row (for the two
// rows marked "semantic literal", the failing row was a Tier 1 semantic
// decimal literal, re-encoded here as BID128 bits; for the row marked
// "panic probe", the mutant kills by runtime panic and the operands come from
// the panicking call frame).
var ffiMutationWitnessCases = []ffiMutationWitnessCase{
	{Function: "bid64_fma", Rounding: 1, Operands: []string{"b1a000000000000b", "f7fb86f26fc0ffff", "f7fb86f26fc0ffff"},
		MutantID: "add128_inline.go:2112:cmp:==->!=", Reason: "add128_inline.go:72 bid_get_add128 (64-bit fma wide-add helper)"},
	{Function: "bid64_add", Rounding: 0, Operands: []string{"6b7386f26fc0ffff", "31c0000000000001"},
		MutantID: "add64.go:6627:const:+1", Reason: "add64.go:236 Bid64AddWithFlags"},
	{Function: "bid128_fma", Rounding: 0, Operands: []string{"01000000000000000000000000003eb0", "01000000000000000000000000000000", "01000000000000000000000000000000"},
		MutantID: "bid128_fma_body.go:65360:negcond:negate", Reason: "bid128_fma_body.go:2087 bid_add_and_round; semantic literal row fma(-1E-1, +1E-6176, +1E-6176) re-encoded as bits"},
	{Function: "bid128_quantize", Rounding: 0, Operands: []string{"000000000a5bc138938d44c64d31fe2f", "01000000000000000000000000004030"},
		MutantID: "bid128_quantize.go:4563:aor:-->+", Reason: "bid128_quantize.go:161 Bid128Quantize; semantic literal row +1000000000000000000000000000000000E-33 quantize +1E+0 re-encoded as bits"},
	{Function: "bid128_fma", Rounding: 0, Operands: []string{"000000000a5bc138938d44c64d31faaf", "ffffffff638e8d37c087adbe09edff5f", "ffffffff638e8d37c087adbe09edff5f"},
		MutantID: "bid128_fma_body.go:31590:aor:-->+", Reason: "bid128_fma_body.go:990 bid_fma_cases_2_to_6; panic probe (index out of range without the pinned Intel subtraction)"},

	// Batch C-2 detection-gap witnesses. The 2026-07-19 re-adjudication of the
	// audit's residual survivors ran each mutant through the full regular
	// chain (goport readtest, goport decTest, public-API parity, native
	// readtest/decTest/FFI, decNumber differential, Tier 1 Structured and
	// DeterministicRandom legs) and then searched for a distinguishing input at
	// a public port entrypoint. These eleven mutants passed every gate and
	// still produced a wrong result or wrong status flags on the operands
	// below, so each row is a measured gap of the regular chain rather than a
	// synthetic case: the mutant is provably non-equivalent and provably
	// undetected before this row existed.
	//
	// Five of them live on the Intel mixed-format entrypoints. Four
	// (Bid64qqMul, Bid64qqDiv x3) were unreachable from any FFI domain at all
	// until this change registered the mixed-format arithmetic family, which is
	// why the audit measured that family as the thinnest covered surface. The
	// rows are pinned here rather than left to the seeded corpus because a
	// seeded corpus that happens to cover them today can stop covering them
	// after any generator reseed.
	{Function: "bid128_fma", Rounding: 0, Operands: []string{"0200000000000000000000000000ea57", "0500000000000000000000000000d68d", "8c7072812f7a5e1faa0600000000c2a7"},
		MutantID: "bid128_fma.go:7523:cmp:<-><=", Reason: "bid128_fma.go:290 bid128_ext_fma"},
	{Function: "bid128_fma", Rounding: 4, Operands: []string{"435d000000000000000000000000f451", "501b32752491c0b3df69311ec306b1bc", "501b32752491c0b3df69311ec306b13c"},
		MutantID: "bid128_round.go:5917:const:+1", Reason: "bid128_round.go:192 bid_round192_39_57 (reached through Bid128Fma)"},
	{Function: "bid128_round_integral_exact", Rounding: 4, Operands: []string{"5a000000000000000000000000003cb0"},
		MutantID: "bid128_round_integral.go:10891:cmp:>=->>", Reason: "bid128_round_integral.go:323 Bid128RoundIntegralExact"},
	{Function: "bid128_round_integral_exact", Rounding: 3, Operands: []string{"f9bfb800000000000000000000003230"},
		MutantID: "bid128_round_integral.go:5547:negcond:negate", Reason: "bid128_round_integral.go:174 Bid128RoundIntegralExact"},
	{Function: "bid128d_sqrt", Rounding: 3, Operands: []string{"360000c465b09a48"},
		MutantID: "bid128_sqrt.go:15195:cmp:==->!=", Reason: "bid128_sqrt.go:628 Bid128dSqrt"},
	{Function: "bid64_fma", Rounding: 0, Operands: []string{"afe7230489e80000", "1aa0000000000009", "1aa0000000000009"},
		MutantID: "inline_round64.go:3686:negcond:negate", Reason: "inline_round64.go:132 __bid_full_round64 (reached through Bid64Fma)"},
	{Function: "bid64q_sqrt", Rounding: 1, Operands: []string{"0010a5d4e8000000000000000000fe2f"},
		MutantID: "sqrt64.go:6178:const:-1", Reason: "sqrt64.go:244 Bid64qSqrt"},
	{Function: "bid64qq_mul", Rounding: 2, Operands: []string{"6ab04059b01dc720acd5b35169f2153b", "1c478271bc15ea0000000000000010a6"},
		MutantID: "bid128_mul.go:1461:const:+1", Reason: "bid128_mul.go:38 Bid64qqMul"},
	{Function: "bid64qq_div", Rounding: 2, Operands: []string{"3800000000000000000000000000fedf", "0500000000000000000000000000b2df"},
		MutantID: "div64.go:22362:aor:+->-", Reason: "div64.go:864 Bid64qqDiv"},
	{Function: "bid64qq_div", Rounding: 0, Operands: []string{"c1ae3ad16ee21bfdf5300f2b6451e35a", "ae3281672beb14000000000000009cdb"},
		MutantID: "div64.go:24267:negcond:negate", Reason: "div64.go:936 Bid64qqDiv"},
	{Function: "bid64qq_div", Rounding: 4, Operands: []string{"47b99912a426f80b2c000000000066c4", "48b99912a426f80b2c000000000066c4"},
		MutantID: "div64.go:24735:aor:+=->-=", Reason: "div64.go:953 Bid64qqDiv"},

	// Batch C-7 coverage-guided witnesses. The audit's residual "gate never
	// executed this line" group was re-measured at this tree against the union
	// of the portable chain, the native FFI/readtest/decTest chain, the
	// decNumber differential and the full Tier 1 long legs: the six mutants
	// below sit on lines that no gate corpus executes at all, so no gate could
	// distinguish them regardless of comparison strength. A directed search at
	// the public port entrypoints (Bid64Fma, Bid64qqqFma, Bid128Fma) found the
	// operands below, which both execute the line and make the mutated port
	// return a different result or different status flags than the pinned
	// port. Each row therefore converts an unexecuted line into an executed
	// one and pins a distinguishing input for it.
	//
	// The operand shapes are the ones the lines require, not arbitrary
	// samples: the three bid64_fma rows drive the 64-bit wide-add helper
	// (bid_get_add128) and its rounding tail with an addend whose exponent
	// sits just under the product's, the bid64qqq_fma row lands a Decimal64
	// result with an unbiased exponent above emax but few enough digits to be
	// scaled back in (the scale-down arm), and the bid128_fma row cancels the
	// product against z at the 10^33 decade boundary so the subnormal
	// half-ulp classification runs, and the last bid128_fma row forces the
	// 256-bit midpoint-addition double carry described above it.
	{Function: "bid64_fma", Rounding: 1, Operands: []string{"b1a05af3107a3ffd", "818000174876e802", "00078f1a37d5699b"},
		MutantID: "add128_inline.go:7046:aor:inc->dec", Reason: "add128_inline.go:263 bid_get_add128 RD/RZ rounding-up correction (reached through Bid64Fma)"},
	{Function: "bid64_fma", Rounding: 2, Operands: []string{"39438d7ea4c67fff", "22c38d7ea4c68001", "29000002540be400"},
		MutantID: "add128_inline.go:7375:const:+1", Reason: "add128_inline.go:277 bid_get_add128 RU exact-fraction clear (reached through Bid64Fma); mutant changes the status flags only"},
	{Function: "bid64_fma", Rounding: 0, Operands: []string{"5621c6bf52634000", "8bdce92c535a7821", "32a0000089035b8b"},
		MutantID: "inline_round64.go:2607:aor:dec->inc", Reason: "inline_round64.go:97 __bid_full_round64 RN-even tie correction (reached through Bid64Fma -> bid_get_add128)"},
	{Function: "bid64qqq_fma", Rounding: 0, Operands: []string{"6400000000000000000000000000deb1", "0a000000000000000000000000009031", "64000000000000000000000000002eb3"},
		MutantID: "bid128_fma.go:23977:aor:*=->+=", Reason: "bid128_fma.go:834 bid64qqqFmaCore unbexp>369 scale-down arm"},
	{Function: "bid128_fma", Rounding: 1, Operands: []string{"000000000a5bc138938d44c64d31f6a5", "0100000000000000000000000000060a", "000000000a5bc138938d44c64d310000"},
		MutantID: "bid128_fma_body.go:24677:const:+1", Reason: "bid128_fma_body.go:792 bid_fma_case1ppB_psign_ne_zsign e3==emin lt_half_ulp arm"},

	// The sixth batch C-7 row closes the same kind of unexecuted-line gap
	// inside bid_round256_58_76's 38 <= ind <= 57 midpoint192 arm. Its target
	// is the innermost double-carry tail of the w1 addition: reaching it needs
	// the w1 addition to overflow while C.w2 is already all ones, so that
	// C.w2++ wraps to zero and the w3 increment runs. The operands make the
	// exact product 2^112 * (2^80-1) = 2^192 - 2^112, whose 256-bit image is
	// exactly {w3:0, w2:ffffffffffffffff, w1:ffff000000000000, w0:0} - the only
	// C.w2 value that lets the wrap happen - and whose 58 significant digits
	// with z = 1E72 drive bid_fma_cases_2_to_6 into
	// bid_round256_58_76(q=58, x=39), i.e. ind=38, the first midpoint192 index.
	// bid_midpoint192[0].w1 = 0x78287f49c4a1d662 then overflows w1. Under the
	// mutant the carry decrements w3 to 2^64-1 instead of raising it to 1, so
	// the 512-bit product against bid_Kx256[38] and the rounded coefficient
	// both change and the exact bit compare against pinned Intel C diverges.
	{Function: "bid128_fma", Rounding: 0, Operands: []string{"00000000000000000000000000004130", "ffffffffffffffffffff000000004030", "0100000000000000000000000000d030"},
		MutantID: "bid128_round.go:11702:aor:inc->dec", Reason: "bid128_round.go:389 bid_round256_58_76 ind=38 midpoint192 w1-carry w2-wrap tail (reached through Bid128Fma -> bid_fma_cases_2_to_6)"},
}

// ffiMutationWitnessIndex groups the witness rows by FFI function for
// consumption inside buildFFICases; ffiVerifyWitnessConsumption fails closed
// when a row's function is absent from the generated FFI suite.
func ffiMutationWitnessIndex() map[string][]ffiMutationWitnessCase {
	index := make(map[string][]ffiMutationWitnessCase)
	for _, w := range ffiMutationWitnessCases {
		index[w.Function] = append(index[w.Function], w)
	}
	return index
}

func ffiVerifyWitnessConsumption(consumed map[string]bool) error {
	for _, w := range ffiMutationWitnessCases {
		if !consumed[w.Function] {
			return fmt.Errorf("ffi mutation witness %s targets function %q, which the generated FFI suite does not exercise; witness rows must stay inside the generated suite", w.MutantID, w.Function)
		}
	}
	return nil
}

// ---- BID32 divide bid_factors32 exactness sweep ----
//
// The audit also measured two surviving mutants inside the package-level
// bid_factors32 table (bid32_div.go:553/564, const +-1): the table feeds the
// exact-quotient trailing-zero elimination of BID32 divide
// (bid32_div.go:154-158 pure / 396-400 WithFlags), and no generated corpus
// consumed the mutated rows. This sweep adds, for every one of the 1024 table
// rows (coefficient values 1..1024), two exact-division bid32_div cases whose
// trailing-zero count reads that row with sign-sensitive weight:
//
//	n2 = ed2 - factors2(y) + factors2(x), n5 likewise for factors of 5,
//	strip = max(min(n2, n5), 0)
//
// A case with n2 < n5 (and n2 >= 1) changes its strip count when factors2 of
// the swept row moves by +-1 in either direction, and symmetrically for
// n5 < n2 and factors5. The generator pins Delta2 = v2(x)-v2(y) and
// Delta5 = v5(x)-v5(y) so that Delta2 < Delta5 (F2 case) or Delta5 < Delta2
// (F5 case); the runtime scaling ed2 >= 2 then keeps the binding count >= 1.
// A changed strip count either changes the result quantum (under-strip) or
// divides a nonzero digit away (over-strip), so the Intel-C-vs-port exact
// bit compare diverges. Rows are swept in both index slots; construction
// avoids the exact-first-division early return (x >= y is only emitted when
// y does not divide x).
//
// Both operands are encoded at the bias exponent (field 101), so results stay
// far from the overflow/underflow boundaries and the exact path is
// rounding-mode independent (mode 0 is used for every sweep case).
const bid32FactorsTableRows = 1024

type bid32DivFactorsSweepCase struct {
	X, Y uint32 // coefficient values (1..1024)
}

func bid32DivFactors32SweepCases() ([]bid32DivFactorsSweepCase, error) {
	v2 := func(v uint32) int { return bits.TrailingZeros32(v) }
	v5 := func(v uint32) int {
		n := 0
		for v%5 == 0 {
			v /= 5
			n++
		}
		return n
	}
	var out []bid32DivFactorsSweepCase
	for v := uint32(1); v <= bid32FactorsTableRows; v++ {
		a, b := v2(v), v5(v)

		// F2 case: row v is read with weight +1 on its factors-of-2 column
		// (x slot) or -1 (y slot), with Delta2 < Delta5.
		var f2 bid32DivFactorsSweepCase
		switch {
		case v == 1024:
			// y slot: 2^11 does not fit the 1..1024 coefficient range, so
			// sweep the row from the y side with x = 32 (Delta2 = 5-10 = -5,
			// Delta5 = 0).
			f2 = bid32DivFactorsSweepCase{X: 32, Y: 1024}
		case v%2 == 1:
			// x slot, y = 2: Delta2 = -1, Delta5 = b >= 0.
			f2 = bid32DivFactorsSweepCase{X: v, Y: 2}
		default:
			// x slot, y = 2^(a+1) (does not divide v, so no early exact
			// return): Delta2 = -1, Delta5 = b >= 0.
			f2 = bid32DivFactorsSweepCase{X: v, Y: 1 << (a + 1)}
		}

		// F5 case: symmetric for the factors-of-5 column.
		var f5 bid32DivFactorsSweepCase
		switch {
		case v == 625:
			// y slot: 5^5 exceeds 1024, so sweep from the y side with x = 25
			// (Delta5 = 2-4 = -2, Delta2 = 0).
			f5 = bid32DivFactorsSweepCase{X: 25, Y: 625}
		default:
			// x slot, y = 5^(b+1) (b <= 3 for every other v <= 1024):
			// Delta5 = -1, Delta2 = a >= 0.
			f5 = bid32DivFactorsSweepCase{X: v, Y: pow5(b + 1)}
		}

		for _, c := range []bid32DivFactorsSweepCase{f2, f5} {
			if c.X < 1 || c.X > bid32FactorsTableRows || c.Y < 1 || c.Y > bid32FactorsTableRows {
				return nil, fmt.Errorf("bid_factors32 sweep row %d: operands (%d, %d) leave the short-coefficient range", v, c.X, c.Y)
			}
			if c.X >= c.Y && c.X%c.Y == 0 {
				return nil, fmt.Errorf("bid_factors32 sweep row %d: operands (%d, %d) take the exact-first-division early return and never read the table", v, c.X, c.Y)
			}
			if c.X != v && c.Y != v {
				return nil, fmt.Errorf("bid_factors32 sweep row %d: case (%d, %d) does not read the swept row", v, c.X, c.Y)
			}
		}
		d2 := func(c bid32DivFactorsSweepCase) int { return v2(c.X) - v2(c.Y) }
		d5 := func(c bid32DivFactorsSweepCase) int { return v5(c.X) - v5(c.Y) }
		if d2(f2) >= d5(f2) {
			return nil, fmt.Errorf("bid_factors32 sweep row %d: F2 case (%d, %d) has Delta2 %d >= Delta5 %d and cannot bind the factors-of-2 column", v, f2.X, f2.Y, d2(f2), d5(f2))
		}
		if d5(f5) >= d2(f5) {
			return nil, fmt.Errorf("bid_factors32 sweep row %d: F5 case (%d, %d) has Delta5 %d >= Delta2 %d and cannot bind the factors-of-5 column", v, f5.X, f5.Y, d5(f5), d2(f5))
		}
		out = append(out, f2, f5)
	}
	if len(out) != 2*bid32FactorsTableRows {
		return nil, fmt.Errorf("bid_factors32 sweep emitted %d cases, want %d", len(out), 2*bid32FactorsTableRows)
	}
	return out, nil
}

func pow5(n int) uint32 {
	v := uint32(1)
	for i := 0; i < n; i++ {
		v *= 5
	}
	return v
}

// encodeBid32SweepOperand encodes a sweep coefficient at the BID32 bias
// exponent (small form, exponent field 101).
func encodeBid32SweepOperand(coeff uint32) string {
	return fmt.Sprintf("%08x", uint32(101)<<23|coeff)
}

// ---- generated flagless-sibling equivalence leg (public parity gate) ----

// parityFlaglessSiblingTarget pins one separately ported flagless port
// entrypoint against its oracle-verified *WithFlags sibling. These six
// flagless bodies are full ported implementations (not thin wrappers), are
// what the public value-only wrappers route through, and have no oracle of
// their own in the regular generated chain: the audit measured 32 surviving
// mutants inside them. The generated leg asserts bit-exact value agreement
// flagless(x, y, mode) == value(WithFlags(x, y, mode)) over the parity
// corpus crossed both ways, a seeded pseudo-random supplement, and the pinned
// witness rows below. The hand-written in-package gate
// (bid754-go/internal/bidgo/flagless_variant_equivalence_test.go) stays the
// independent architecture anchor; this generated leg puts the same contract
// inside a regular generated domain with pinned counts.
type parityFlaglessSiblingTarget struct {
	Width     int
	Flagless  string
	WithFlags string
	Reason    string
}

var parityFlaglessSiblingTargets = []parityFlaglessSiblingTarget{
	{32, "Bid32Add", "Bid32AddWithFlags", "separately ported flagless BID32 add body (bid32_add_pure)"},
	{32, "Bid32Sub", "Bid32SubWithFlags", "separately ported flagless BID32 sub body (routes through bid32_add_pure)"},
	{32, "Bid32Mul", "Bid32MulWithFlags", "separately ported flagless BID32 mul body (bid32_mul_pure)"},
	{32, "Bid32Div", "Bid32DivWithFlags", "separately ported flagless BID32 div body (bid32_div_pure)"},
	{64, "Bid64Mul", "Bid64MulWithFlags", "separately ported flagless BID64 mul body (mul64.go Bid64Mul)"},
	{64, "Bid64Div", "Bid64DivWithFlags", "separately ported flagless BID64 div body (div64.go Bid64Div)"},
}

// Deterministic pseudo-random supplement sizes and seeds for the generated
// flagless-sibling leg (matching the density that killed the audited pure-path
// mutants in the hand-written gate).
const (
	parityFlaglessRandomPairs32 = 1 << 20
	parityFlaglessRandomPairs64 = 4096
	parityFlaglessSeed32        = 0x754c1f32a95
	parityFlaglessSeed64        = 0x754c1f64a95
	parityFlaglessModeCount     = 5
)

// parityFlaglessWitnessRow is one pinned distinguishing input for a
// mutation-audit witness mutant on a flagless port body. X and Y carry the
// operand bits at the target's width.
type parityFlaglessWitnessRow struct {
	Target   string
	X, Y     uint64
	Mode     int
	MutantID string
}

var parityFlaglessWitnessRows = []parityFlaglessWitnessRow{
	{Target: "Bid32Add", X: 0x78000000, Y: 0x78000000, Mode: 0, MutantID: "bid32_add.go:1309:negcond:negate"},             // bid32_add.go:42 bid32_add_pure
	{Target: "Bid32Add", X: 0x78000000, Y: 0x00000000, Mode: 0, MutantID: "bid32_add.go:1477:bit:&->|"},                   // bid32_add.go:51 bid32_add_pure
	{Target: "Bid32Add", X: 0x32800000, Y: 0xb2800000, Mode: 0, MutantID: "bid32_add.go:1717:const:+1"},                   // bid32_add.go:60 bid32_add_pure
	{Target: "Bid32Add", X: 0x32800000, Y: 0x00000000, Mode: 0, MutantID: "bid32_add.go:1987:negcond:negate"},             // bid32_add.go:75 bid32_add_pure
	{Target: "Bid32Add", X: 0x32800001, Y: 0x32800001, Mode: 0, MutantID: "bid32_add.go:3309:aor:+->-"},                   // bid32_add.go:131 bid32_add_pure
	{Target: "Bid32Add", X: 0x0ddca526, Y: 0x0d56c291, Mode: 0, MutantID: "bid32_add.go:4371:aor:*->+"},                   // bid32_add.go:179 bid32_add_pure
	{Target: "Bid32Add", X: 0x32800000, Y: 0x77f8967f, Mode: 1, MutantID: "bid32_add.go:4422:cmp:==->!="},                 // bid32_add.go:181 bid32_add_pure
	{Target: "Bid32Div", X: 0x32800000, Y: 0x32800000, Mode: 0, MutantID: "bid32_div.go:1138:negcond:negate"},             // bid32_div.go:39 bid32_div_pure
	{Target: "Bid32Div", X: 0x32800000, Y: 0x32800000, Mode: 0, MutantID: "bid32_div.go:1491:negcond:negate"},             // bid32_div.go:52 bid32_div_pure
	{Target: "Bid32Div", X: 0x43421f4e, Y: 0x2885003c, Mode: 0, MutantID: "bid32_div.go:3539:aor:-->+"},                   // bid32_div.go:128 bid32_div_pure
	{Target: "Bid32Div", X: 0x77f8967f, Y: 0x3280007b, Mode: 0, MutantID: "bid32_div.go:3570:aor:-->+"},                   // bid32_div.go:131 bid32_div_pure
	{Target: "Bid32Div", X: 0x32800001, Y: 0x3280007b, Mode: 0, MutantID: "bid32_div.go:3870:aor:*->+"},                   // bid32_div.go:144 bid32_div_pure
	{Target: "Bid32Div", X: 0xb2db13aa, Y: 0x608d5802, Mode: 0, MutantID: "bid32_div.go:5615:aor:-->+"},                   // bid32_div.go:217 bid32_div_pure
	{Target: "Bid32Mul", X: 0x00000000, Y: 0x32800000, Mode: 0, MutantID: "bid32_mul.go:1030:bit:&->|"},                   // bid32_mul.go:34 bid32_mul_pure
	{Target: "Bid32Mul", X: 0x78000000, Y: 0x32800000, Mode: 0, MutantID: "bid32_mul.go:1121:cmp:!=->=="},                 // bid32_mul.go:36 bid32_mul_pure
	{Target: "Bid32Mul", X: 0x78000000, Y: 0x32800001, Mode: 0, MutantID: "bid32_mul.go:1264:cmp:==->!="},                 // bid32_mul.go:41 bid32_mul_pure
	{Target: "Bid32Mul", X: 0x3280007b, Y: 0x77f8967f, Mode: 0, MutantID: "bid32_mul.go:3066:aor:+=->-="},                 // bid32_mul.go:112 bid32_mul_pure
	{Target: "Bid32Mul", X: 0xacb27555, Y: 0x07000005, Mode: 0, MutantID: "bid32_mul.go:3696:aor:-->+"},                   // bid32_mul.go:138 bid32_mul_pure
	{Target: "Bid32Mul", X: 0x6cb8967f, Y: 0x6cb8967f, Mode: 2, MutantID: "bid32_mul.go:3744:negcond:negate"},             // bid32_mul.go:140 bid32_mul_pure
	{Target: "Bid64Div", X: 0x0000000000000000, Y: 0x0000000000000000, Mode: 0, MutantID: "div64.go:1512:cmp:==->!="},     // div64.go:49 Bid64Div
	{Target: "Bid64Div", X: 0x0000000000000000, Y: 0x0000000000000001, Mode: 0, MutantID: "div64.go:1983:negcond:negate"}, // div64.go:64 Bid64Div
	{Target: "Bid64Div", X: 0x31c0000000000001, Y: 0x77fb86f26fc0ffff, Mode: 0, MutantID: "div64.go:2767:bit:>>-><<"},     // div64.go:95 Bid64Div (panic probe)
	{Target: "Bid64Div", X: 0x0000000000000001, Y: 0x0000000000000009, Mode: 0, MutantID: "div64.go:4329:aor:+->-"},       // div64.go:168 Bid64Div
	{Target: "Bid64Div", X: 0x6000000000000000, Y: 0x31c000000000000a, Mode: 0, MutantID: "div64.go:6196:const:+1"},       // div64.go:258 Bid64Div
	{Target: "Bid64Mul", X: 0x7800000000000000, Y: 0x0000000000000000, Mode: 0, MutantID: "mul64.go:1493:negcond:negate"}, // mul64.go:45 Bid64Mul
	{Target: "Bid64Mul", X: 0x7800000000000000, Y: 0x0000000000000000, Mode: 0, MutantID: "mul64.go:1516:cmp:!=->=="},     // mul64.go:45 Bid64Mul
	{Target: "Bid64Mul", X: 0x7800000000000000, Y: 0x0000000000000000, Mode: 0, MutantID: "mul64.go:1556:const:+1"},       // mul64.go:45 Bid64Mul
	{Target: "Bid64Mul", X: 0x0000000000000000, Y: 0x7800000000000000, Mode: 0, MutantID: "mul64.go:2748:cmp:==->!="},     // mul64.go:89 Bid64Mul
	{Target: "Bid64Mul", X: 0x31c0000000000001, Y: 0x0000000000000000, Mode: 0, MutantID: "mul64.go:2924:aor:+=->-="},     // mul64.go:97 Bid64Mul
	{Target: "Bid64Mul", X: 0x2de38d7ea4c68000, Y: 0x6000000000000000, Mode: 0, MutantID: "mul64.go:5435:const:-1"},       // mul64.go:178 Bid64Mul
	{Target: "Bid64Mul", X: 0x31c000000000000a, Y: 0x6000000000000000, Mode: 0, MutantID: "mul64.go:5859:const:+1"},       // mul64.go:195 Bid64Mul
	{Target: "Bid64Mul", X: 0x000000000000000a, Y: 0x77fb86f26fc0ffff, Mode: 0, MutantID: "mul64.go:6598:aor:-->+"},       // mul64.go:217 Bid64Mul

	// Batch C-7 residual rows. Both mutants sit on the round-to-nearest tail of
	// a flagless body, on a line the leg already executes, and both need an
	// operand whose discarded remainder is *adjacent* to the exact halfway
	// point rather than at it - which no corpus row happened to hit.
	//
	// div64.go:320 is Intel's `R -= (Q | (rmode >> 2)) & 1` ties-away
	// correction: only rmode == BID_ROUNDING_TIES_AWAY reaches it with a
	// nonzero shift result, so the mutated shift count changes the correction
	// to `Q & 1`, and the two disagree exactly on an even quotient at an exact
	// midpoint. 2000000000000001E+0 / 2E+0 = 1000000000000000.5 is that
	// midpoint with an even truncated quotient; ties-away must deliver
	// ...0001 and the mutant truncates to ...0000.
	//
	// mul64.go:221 is the 128-bit `Q_low < reciprocals10_128[k]` exactness test
	// of the round-half-to-even correction. Widening its high-word compare to
	// `<=` only matters when Q_low's high word equals the reciprocal's, which
	// happens exactly when the pre-rounding remainder is one unit above the
	// halfway point (Q_low = k*d + reciprocal). 2E+0 * 5000000000000003E+0 =
	// 10000000000000006 puts the discarded digit at 6 with an odd rounded
	// coefficient, so the mutant wrongly applies the tie-to-even decrement.
	{Target: "Bid64Div", X: 0x31c71afd498d0001, Y: 0x31c0000000000002, Mode: 4, MutantID: "div64.go:7610:const:-1"},  // div64.go:320 Bid64Div
	{Target: "Bid64Mul", X: 0x31c0000000000002, Y: 0x31d1c37937e08003, Mode: 0, MutantID: "mul64.go:6698:cmp:<-><="}, // mul64.go:221 Bid64Mul
}

// validateParityFlaglessSiblingTables enforces the closed world of the
// flagless-sibling leg against the loaded bidgo port signatures: every target
// must resolve to a real flagless/WithFlags pair with the expected binary
// arithmetic shapes at its width, and every witness row must reference a
// declared target with operands that fit the width.
func validateParityFlaglessSiblingTables(sigs map[string]bidgoFuncSig) error {
	widthUint := map[int]string{32: "uint32", 64: "uint64"}
	byName := map[string]parityFlaglessSiblingTarget{}
	for _, target := range parityFlaglessSiblingTargets {
		if _, dup := byName[target.Flagless]; dup {
			return fmt.Errorf("flagless sibling target %q is declared twice", target.Flagless)
		}
		if target.Reason == "" {
			return fmt.Errorf("flagless sibling target %q needs a concrete reason", target.Flagless)
		}
		uintType, ok := widthUint[target.Width]
		if !ok {
			return fmt.Errorf("flagless sibling target %q has unsupported width %d", target.Flagless, target.Width)
		}
		flagless, ok := sigs[target.Flagless]
		if !ok {
			return fmt.Errorf("flagless sibling target %q does not exist in the bidgo port", target.Flagless)
		}
		withFlags, ok := sigs[target.WithFlags]
		if !ok {
			return fmt.Errorf("flagless sibling target %q names missing WithFlags sibling %q", target.Flagless, target.WithFlags)
		}
		if err := checkFlaglessShape(flagless, uintType, false); err != nil {
			return fmt.Errorf("flagless sibling target %q: %w", target.Flagless, err)
		}
		if err := checkFlaglessShape(withFlags, uintType, true); err != nil {
			return fmt.Errorf("flagless sibling target %q sibling %q: %w", target.Flagless, target.WithFlags, err)
		}
		byName[target.Flagless] = target
	}
	for _, row := range parityFlaglessWitnessRows {
		target, ok := byName[row.Target]
		if !ok {
			return fmt.Errorf("flagless witness row %s references undeclared target %q", row.MutantID, row.Target)
		}
		if target.Width == 32 && (row.X > 0xffffffff || row.Y > 0xffffffff) {
			return fmt.Errorf("flagless witness row %s has operands beyond width 32", row.MutantID)
		}
		if row.Mode < 0 || row.Mode >= parityFlaglessModeCount {
			return fmt.Errorf("flagless witness row %s has out-of-domain rounding mode %d", row.MutantID, row.Mode)
		}
		if row.MutantID == "" {
			return fmt.Errorf("flagless witness row for target %q needs its mutant ID", row.Target)
		}
	}
	return nil
}

func checkFlaglessShape(sig bidgoFuncSig, uintType string, wantFlags bool) error {
	if len(sig.Params) != 3 || sig.Params[0].Type != uintType || sig.Params[1].Type != uintType || sig.Params[2].Type != "int" {
		return fmt.Errorf("signature params %v do not match (%s, %s, int)", sig.Params, uintType, uintType)
	}
	if wantFlags {
		if len(sig.Results) != 2 || sig.Results[0].Type != uintType || sig.Results[1].Type != "uint32" {
			return fmt.Errorf("signature results %v do not match (%s, uint32)", sig.Results, uintType)
		}
		return nil
	}
	if len(sig.Results) != 1 || sig.Results[0].Type != uintType {
		return fmt.Errorf("signature results %v do not match (%s)", sig.Results, uintType)
	}
	return nil
}

// parityFlaglessSiblingCaseCount is the pinned case count of the generated
// flagless-sibling leg: per target, the width corpus crossed both ways plus
// the seeded random pairs, each under all five rounding modes, plus one case
// per pinned witness row.
func parityFlaglessSiblingCaseCount() int {
	total := 0
	for _, target := range parityFlaglessSiblingTargets {
		pairs := publicParityCorpusLen * publicParityCorpusLen
		if target.Width == 32 {
			pairs += parityFlaglessRandomPairs32
		} else {
			pairs += parityFlaglessRandomPairs64
		}
		total += pairs * parityFlaglessModeCount
	}
	return total + len(parityFlaglessWitnessRows)
}
