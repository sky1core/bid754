// Hand-written unreachability certificate for the high-word carry arms of the
// Decimal128 round-to-integral family. It lives outside every generation path
// and must stay hand-written.
//
// Why this gate exists: bid128_round_integral.go:475 (Bid128RoundIntegralExact)
// and bid128_nearbyint.go:358 (Bid128Nearbyint) carry the `res.hi++` half of
// Intel's uniform
//
//	res.w0++; if (res.w0 == 0x0) res.w1++;
//
// increment. Intel writes that pair identically in all three shift arms of the
// routine, but in the widest arm - selected by `22 <= ind-1 <= 33`, i.e. at
// least 23 fractional decimal digits - the truncated integral part cannot come
// anywhere near 2^64, so the carry can never fire. The 2026-07 mutation audit
// measured both `res.hi++` lines as executed by no verification corpus, and a
// directed 465M-call search at the public entrypoints could not reach them
// either; this test is the reason why, expressed as a bound that a future
// table or lowering change would break.
//
// What it pins, from the port's own tables rather than from a restatement of
// the algorithm:
//
//	res.lo = P256.w3 >> (bid_shiftright128[ind-1] - 64), where
//	P256 = C1 * bid_ten2mk128[ind-1]
//
// equals floor(C1 / 10^ind) exactly for every canonical Decimal128
// coefficient C1 < 10^34 and every ind in the arm's range, and that quotient
// is at most floor((10^34-1) / 10^23) = 10^11 - 1. Since res.lo is bounded by
// 10^11-1 < 2^64-1, `res.lo++` cannot wrap and the `res.hi++` line is
// unreachable. Both routines compute res.lo with the same expression and the
// same tables, so one bound covers both call sites.
//
// If a future change makes the arm's quotient reach 2^64-1, this test fails
// and the two lines stop being dead code - which is exactly when the mutation
// audit's classification of them must be revisited.
package bidgo

import (
	"math/bits"
	"math/rand"
	"testing"
)

// roundIntegralWideArmIndices is the ind range of the arm that carries the two
// lines: the routines select it with `22 <= ind-1 <= 33` and reach it only
// when q+exp > 0 with ind = -exp, so ind <= 33 in practice. 34 is included so
// the bound also covers the last table slot the arm could index.
func roundIntegralWideArmIndices() []int {
	var out []int
	for ind := 23; ind <= 34; ind++ {
		out = append(out, ind)
	}
	return out
}

// maxCanonicalCoefficient128 is 10^34 - 1, the largest canonical Decimal128
// coefficient (the same constant the port uses in bid128_add.go).
var maxCanonicalCoefficient128 = BID_UINT128{lo: 0x378d8e63ffffffff, hi: 0x0001ed09bead87c0}

func TestRoundIntegralWideArmQuotientCannotReachHighWordCarry(t *testing.T) {
	coefficients := roundIntegralCertificateCoefficients()
	for _, ind := range roundIntegralWideArmIndices() {
		shift := bid_shiftright128[ind-1] - 64
		if shift < 0 {
			t.Fatalf("ind=%d: shift %d is negative; the arm's lowering assumes a >= 64-bit shift", ind, shift)
		}
		// The arm's quotient is monotone in C1, so the maximum over the whole
		// canonical coefficient range is attained at 10^34-1.
		maxQuotient := roundIntegralWideArmQuotient(maxCanonicalCoefficient128, ind)
		if maxQuotient == ^uint64(0) {
			t.Fatalf("ind=%d: maximal quotient is 2^64-1; res.lo++ would wrap and the res.hi++ carry arms stop being dead", ind)
		}
		for _, c := range coefficients {
			got := roundIntegralWideArmQuotient(c, ind)
			wantHi, wantLo := quotientBy10Pow(c, ind)
			if wantHi != 0 {
				t.Fatalf("ind=%d C1=%016x%016x: floor(C1/10^ind) does not fit in 64 bits (hi=%016x)", ind, c.hi, c.lo, wantHi)
			}
			if got != wantLo {
				t.Fatalf("ind=%d C1=%016x%016x: arm quotient %d, want floor(C1/10^%d) = %d",
					ind, c.hi, c.lo, got, ind, wantLo)
			}
			if got > maxQuotient {
				t.Fatalf("ind=%d C1=%016x%016x: quotient %d exceeds the maximal-coefficient quotient %d; the monotonicity the bound relies on is broken",
					ind, c.hi, c.lo, got, maxQuotient)
			}
		}
	}
}

// roundIntegralWideArmQuotient reproduces the arm's own lowering: the same
// 128x128->256 product against bid_ten2mk128 and the same
// bid_shiftright128-driven extraction that bid128_round_integral.go:461 and
// bid128_nearbyint.go:347 use to fill res.lo.
func roundIntegralWideArmQuotient(c BID_UINT128, ind int) uint64 {
	p := __mul_128x128_to_256(c, bid_ten2mk128[ind-1])
	return p.w3 >> uint(bid_shiftright128[ind-1]-64)
}

// quotientBy10Pow divides a 128-bit value by 10^n with exact 128-bit
// arithmetic, independently of the reciprocal tables the port uses.
func quotientBy10Pow(c BID_UINT128, n int) (hi, lo uint64) {
	hi, lo = c.hi, c.lo
	for i := 0; i < n; i++ {
		q1, r := bits.Div64(0, hi, 10)
		q0, _ := bits.Div64(r, lo, 10)
		hi, lo = q1, q0
	}
	return hi, lo
}

// roundIntegralCertificateCoefficients returns the canonical Decimal128
// coefficients the certificate is checked on: every decade boundary and its
// neighbours, the maximal coefficient, and a seeded pseudo-random sweep.
func roundIntegralCertificateCoefficients() []BID_UINT128 {
	out := []BID_UINT128{
		{lo: 0, hi: 0},
		{lo: 1, hi: 0},
		maxCanonicalCoefficient128,
	}
	pow := BID_UINT128{lo: 1, hi: 0}
	for d := 0; d <= 33; d++ {
		out = append(out, pow)
		if pow.lo != 0 || pow.hi != 0 {
			out = append(out, sub128Small(pow, 1))
		}
		out = append(out, add128Small(pow, 1))
		pow = mul128Small(pow, 10)
	}
	r := rand.New(rand.NewSource(7540754))
	for i := 0; i < 4096; i++ {
		c := BID_UINT128{lo: r.Uint64(), hi: r.Uint64() & 0x0001ffffffffffff}
		for compare128(c, maxCanonicalCoefficient128) > 0 {
			// The 49-bit coefficient field reaches past 10^34-1; fold the
			// out-of-range draws back into the canonical range instead of
			// dropping them.
			c.hi >>= 1
		}
		out = append(out, c)
	}
	return out
}

func add128Small(c BID_UINT128, v uint64) BID_UINT128 {
	lo := c.lo + v
	hi := c.hi
	if lo < c.lo {
		hi++
	}
	return BID_UINT128{lo: lo, hi: hi}
}

func sub128Small(c BID_UINT128, v uint64) BID_UINT128 {
	lo := c.lo - v
	hi := c.hi
	if c.lo < v {
		hi--
	}
	return BID_UINT128{lo: lo, hi: hi}
}

func mul128Small(c BID_UINT128, m uint64) BID_UINT128 {
	hi, lo := bits.Mul64(c.lo, m)
	return BID_UINT128{lo: lo, hi: c.hi*m + hi}
}

func compare128(a, b BID_UINT128) int {
	switch {
	case a.hi != b.hi:
		if a.hi < b.hi {
			return -1
		}
		return 1
	case a.lo != b.lo:
		if a.lo < b.lo {
			return -1
		}
		return 1
	}
	return 0
}
