// Hand-written correctness coverage for `__mul_256x256_to_512` and for its one
// production caller, `bid_round256_58_76`.
//
// Failure mode: the helper's body was once a hand-designed 128x128-block
// schoolbook decomposition instead of a transcription of Intel's
// `__mul_256x256_to_512` macro (devtools/third_party/intel_dfp/LIBRARY/src/
// bid_internal.h, lines ~569-593), and it dropped a carry out of the second
// 128-bit column. On a = 2^128+1, b = 2^256-2^128+1 it returned
// 0xffff...f * 2^256 instead of 2^384 + 1, and the same defect corrupted Cstar
// out of bid_round256_58_76 inside that routine's documented domain.
//
// Coverage layers, deliberately separate, all calling the production symbols
// directly (a local copy would keep passing after a regression):
//
//   - pinned helper witness with literal expected limbs;
//   - pinned caller witness with literal expected Cstar limbs;
//   - boundary-table oracle and fixed-seed random oracle against math/big.
package bidgo

import (
	"math/big"
	"math/rand"
	"strconv"
	"testing"
)

const mul256TestMaxU64 = uint64(0xffffffffffffffff)

// mul256BigFrom256 converts a BID_UINT256 to its exact integer value.
func mul256BigFrom256(v BID_UINT256) *big.Int {
	r := new(big.Int)
	for i, w := range [4]uint64{v.w0, v.w1, v.w2, v.w3} {
		term := new(big.Int).SetUint64(w)
		term.Lsh(term, uint(64*i))
		r.Or(r, term)
	}
	return r
}

// mul256BigTo512Limbs splits an exact integer into the eight little-endian
// 64-bit limbs of a BID_UINT512, and reports whether it fit in 512 bits.
func mul256BigTo512Limbs(v *big.Int) ([8]uint64, bool) {
	var out [8]uint64
	if v.Sign() < 0 || v.BitLen() > 512 {
		return out, false
	}
	tmp := new(big.Int)
	mask := new(big.Int).SetUint64(mul256TestMaxU64)
	for i := range out {
		tmp.Rsh(v, uint(64*i))
		tmp.And(tmp, mask)
		out[i] = tmp.Uint64()
	}
	return out, true
}

// mul256Limbs512 flattens a BID_UINT512 into little-endian limb order.
func mul256Limbs512(p BID_UINT512) [8]uint64 {
	return [8]uint64{p.w0, p.w1, p.w2, p.w3, p.w4, p.w5, p.w6, p.w7}
}

// mul256U256 builds a BID_UINT256 from little-endian limbs.
func mul256U256(w0, w1, w2, w3 uint64) BID_UINT256 {
	return BID_UINT256{w0: w0, w1: w1, w2: w2, w3: w3}
}

// TestMul256x256To512PinnedIntelWitness pins the exact input pair that
// separated the previous hand-designed body from Intel's macro, with literal
// expected limbs written out rather than computed, so this case cannot drift
// with the oracle or the random corpus.
func TestMul256x256To512PinnedIntelWitness(t *testing.T) {
	// a = 2^128 + 1
	a := mul256U256(1, 0, 1, 0)
	// b = 2^256 - 2^128 + 1
	b := mul256U256(1, 0, mul256TestMaxU64, mul256TestMaxU64)

	// a*b = (2^128 + 1) * (2^256 - 2^128 + 1) = 2^384 + 1 exactly.
	want := [8]uint64{1, 0, 0, 0, 0, 0, 1, 0}

	got := mul256Limbs512(__mul_256x256_to_512(a, b))
	if got != want {
		t.Fatalf("__mul_256x256_to_512(2^128+1, 2^256-2^128+1) limbs =\n  %#016x\nwant\n  %#016x",
			got, want)
	}

	// Independent confirmation that the literal above is the true product, so
	// a future editor cannot "fix" the pin to match a broken implementation.
	exact := new(big.Int).Mul(mul256BigFrom256(a), mul256BigFrom256(b))
	oracle, ok := mul256BigTo512Limbs(exact)
	if !ok {
		t.Fatalf("pinned witness product does not fit in 512 bits: %s", exact.String())
	}
	if oracle != want {
		t.Fatalf("pinned literal disagrees with math/big: literal %#016x, math/big %#016x", want, oracle)
	}
	if bl := exact.BitLen(); bl != 385 {
		t.Fatalf("pinned witness product bit length = %d, want 385 (2^384 + 1)", bl)
	}
}

// TestBidRound256_58_76PinnedMul256Witness pins a production-caller witness for
// the same defect: with the hand-designed helper body, bid_round256_58_76
// returned a wrong Cstar for this (q, x, C) triple, which is inside the
// routine's documented domain (58 <= q <= 76, 1 <= x <= q-1).
func TestBidRound256_58_76PinnedMul256Witness(t *testing.T) {
	const (
		q       = 58
		x       = 1
		decimal = "4483276967880711252201765259069979864743078621649663452830"
	)

	cBig, ok := new(big.Int).SetString(decimal, 10)
	if !ok {
		t.Fatalf("could not parse witness coefficient %q", decimal)
	}
	cLimbs, fits := mul256BigTo512Limbs(cBig)
	if !fits || cLimbs[4] != 0 || cLimbs[5] != 0 || cLimbs[6] != 0 || cLimbs[7] != 0 {
		t.Fatalf("witness coefficient does not fit in 256 bits: %s", decimal)
	}
	if got := len(decimal); got != q {
		t.Fatalf("witness coefficient has %d decimal digits, want q = %d", got, q)
	}
	c := mul256U256(cLimbs[0], cLimbs[1], cLimbs[2], cLimbs[3])

	// Literal expected Cstar limbs, pinned independently of the check below.
	wantCstar := [4]uint64{0x8dae6f47bf020a43, 0x196635fb712d5b61, 0x1248c2724fed0055, 0}

	cstar, incrExp, midLtEven, midGtEven, inexactLtMid, inexactGtMid := bid_round256_58_76(q, x, c)

	gotCstar := [4]uint64{cstar.w0, cstar.w1, cstar.w2, cstar.w3}
	if gotCstar != wantCstar {
		t.Errorf("bid_round256_58_76(%d, %d, C).Cstar limbs =\n  %#016x\nwant\n  %#016x",
			q, x, gotCstar, wantCstar)
	}

	// The witness decimal ends in 0 and x = 1, so removing one digit is exact
	// division by 10: Cstar is C/10, no midpoint or inexact flag may be raised,
	// and Cstar keeps q - x = 57 digits so there is no decimal carry either.
	if decimal[len(decimal)-1] != '0' {
		t.Fatalf("witness coefficient must end in a zero digit for exact division: %s", decimal)
	}
	quo := new(big.Int).Quo(cBig, big.NewInt(10))
	quoLimbs, fits := mul256BigTo512Limbs(quo)
	if !fits {
		t.Fatalf("C/10 does not fit in 512 bits: %s", quo.String())
	}
	if exactCstar := [4]uint64{quoLimbs[0], quoLimbs[1], quoLimbs[2], quoLimbs[3]}; exactCstar != wantCstar {
		t.Fatalf("pinned Cstar literal disagrees with C/10: literal %#016x, C/10 %#016x",
			wantCstar, exactCstar)
	}
	if incrExp != 0 {
		t.Errorf("incr_exp = %d, want 0", incrExp)
	}
	if midLtEven != 0 {
		t.Errorf("is_midpoint_lt_even = %d, want 0", midLtEven)
	}
	if midGtEven != 0 {
		t.Errorf("is_midpoint_gt_even = %d, want 0", midGtEven)
	}
	if inexactLtMid != 0 {
		t.Errorf("is_inexact_lt_midpoint = %d, want 0", inexactLtMid)
	}
	if inexactGtMid != 0 {
		t.Errorf("is_inexact_gt_midpoint = %d, want 0", inexactGtMid)
	}
}

// mul256BoundaryValues is a compact table of 256-bit operands chosen to hit the
// carry-sensitive shapes of the macro's three ripple chains: zero, one, the
// all-ones saturation value, single high bits at every limb boundary, and
// alternating bit patterns.
var mul256BoundaryValues = []BID_UINT256{
	mul256U256(0, 0, 0, 0),
	mul256U256(1, 0, 0, 0),
	mul256U256(2, 0, 0, 0),
	mul256U256(mul256TestMaxU64, mul256TestMaxU64, mul256TestMaxU64, mul256TestMaxU64),
	mul256U256(mul256TestMaxU64, 0, 0, 0),
	mul256U256(0, 1, 0, 0),
	mul256U256(0, 0, 1, 0),
	mul256U256(0, 0, 0, 1),
	mul256U256(1, 0, 1, 0),
	mul256U256(1, 0, mul256TestMaxU64, mul256TestMaxU64),
	mul256U256(0, 0, 0, 0x8000000000000000),
	mul256U256(0x8000000000000000, 0, 0, 0),
	mul256U256(0x8000000000000000, 0x8000000000000000, 0x8000000000000000, 0x8000000000000000),
	mul256U256(mul256TestMaxU64, mul256TestMaxU64, 0, 0),
	mul256U256(0, 0, mul256TestMaxU64, mul256TestMaxU64),
	mul256U256(mul256TestMaxU64, 0, mul256TestMaxU64, 0),
	mul256U256(0, mul256TestMaxU64, 0, mul256TestMaxU64),
	mul256U256(0xaaaaaaaaaaaaaaaa, 0x5555555555555555, 0xaaaaaaaaaaaaaaaa, 0x5555555555555555),
	mul256U256(0x5555555555555555, 0xaaaaaaaaaaaaaaaa, 0x5555555555555555, 0xaaaaaaaaaaaaaaaa),
	mul256U256(mul256TestMaxU64-1, mul256TestMaxU64, mul256TestMaxU64, mul256TestMaxU64),
	mul256U256(0, mul256TestMaxU64, mul256TestMaxU64, mul256TestMaxU64),
}

// mul256CheckAgainstOracle compares every one of the eight result limbs against
// the exact math/big product.
func mul256CheckAgainstOracle(t *testing.T, label string, a, b BID_UINT256) {
	t.Helper()
	exact := new(big.Int).Mul(mul256BigFrom256(a), mul256BigFrom256(b))
	want, ok := mul256BigTo512Limbs(exact)
	if !ok {
		t.Fatalf("%s: product exceeds 512 bits, which is impossible for 256x256", label)
	}
	got := mul256Limbs512(__mul_256x256_to_512(a, b))
	if got != want {
		t.Fatalf("%s: __mul_256x256_to_512(%#016x, %#016x) limbs =\n  %#016x\nwant\n  %#016x",
			label,
			[4]uint64{a.w0, a.w1, a.w2, a.w3},
			[4]uint64{b.w0, b.w1, b.w2, b.w3},
			got, want)
	}
}

// TestMul256x256To512MatchesBigIntOracleBoundaries cross-multiplies the whole
// boundary table in both operand orders.
func TestMul256x256To512MatchesBigIntOracleBoundaries(t *testing.T) {
	for i, a := range mul256BoundaryValues {
		for j, b := range mul256BoundaryValues {
			mul256CheckAgainstOracle(t, mul256Label("boundary", i, j), a, b)
		}
	}
}

func mul256Label(kind string, i, j int) string {
	return kind + "[" + strconv.Itoa(i) + "][" + strconv.Itoa(j) + "]"
}

// TestMul256x256To512MatchesBigIntOracleRandom sweeps a fixed-seed
// pseudorandom corpus. The seed is pinned so a failure reproduces exactly, and
// the operand generator mixes full-width draws with limb-sparse and
// limb-saturated shapes so carry chains are exercised, not just typical
// mid-range products.
func TestMul256x256To512MatchesBigIntOracleRandom(t *testing.T) {
	const iterations = 20000

	r := rand.New(rand.NewSource(0x6d756c32_35360001))
	for i := 0; i < iterations; i++ {
		a := mul256RandomOperand(r)
		b := mul256RandomOperand(r)
		mul256CheckAgainstOracle(t, mul256Label("random", i, 0), a, b)
	}

	// Also pair the random corpus against the boundary table, which finds
	// carry interactions that two independent random draws rarely produce.
	for i := 0; i < iterations/4; i++ {
		a := mul256RandomOperand(r)
		b := mul256BoundaryValues[r.Intn(len(mul256BoundaryValues))]
		mul256CheckAgainstOracle(t, mul256Label("random-boundary", i, 0), a, b)
		mul256CheckAgainstOracle(t, mul256Label("boundary-random", i, 0), b, a)
	}
}

// mul256RandomOperand draws one 256-bit operand, biased toward carry-relevant
// limb shapes.
func mul256RandomOperand(r *rand.Rand) BID_UINT256 {
	var w [4]uint64
	for i := range w {
		switch r.Intn(8) {
		case 0:
			w[i] = 0
		case 1:
			w[i] = mul256TestMaxU64
		case 2:
			w[i] = mul256TestMaxU64 - uint64(r.Intn(4))
		case 3:
			w[i] = uint64(r.Intn(4))
		default:
			w[i] = r.Uint64()
		}
	}
	return mul256U256(w[0], w[1], w[2], w[3])
}
