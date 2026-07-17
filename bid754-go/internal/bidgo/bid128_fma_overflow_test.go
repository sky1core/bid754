package bidgo

import "testing"

// TestBid128FmaCase1ppBOverflowEncoding locks the fix for a transcription
// defect in the Case (1”B) overflow path of the bid128_ext_fma port.
//
// In Intel bid128_fma.c the Case (1”B) p_sign != z_sign overflow exits
// (lines 2117-2136 and 2245-2264, e3 > expmax) finish with
// "BID_SWAP128 (res); BID_RETURN (res)" and therefore never execute the
// final Case (1”B) reassembly "res.w[1] = z_sign | (z_exp & MASK_EXP) |
// res.w[1]" (line 2324). The Go port extracted that branch into
// bid_fma_case1ppB_psign_ne_zsign, whose early returns used to fall back
// into the caller's unconditional reassembly, overwriting the +/-Inf or
// +/-Nmax overflow result with the stale unclamped biased exponent 12321
// (6145 > emax 6144) and producing non-canonical bits such as
// 0x6042000000000000:0 (RN) and 0x6043ed09bead87c0:378d8e63ffffffff (RD).
//
// Inputs reach that path with x = 1E+34, y = +/-Nmax, z = -/+Nmax: the
// exact product (10^34-1)E+6145 forces the delta < 0 Case (8) swap into
// Case (1”B) with e3 = 6145 and opposite signs. Expected bits and flags
// below were pinned directly from pinned Intel C bid128_fma (all five
// rounding modes agree with IEEE 754-2019 overflow rules and decNumber).
func TestBid128FmaCase1ppBOverflowEncoding(t *testing.T) {
	oneE34 := BID_UINT128{hi: 0x3084000000000000, lo: 0x0000000000000001}
	posNmax := BID_UINT128{hi: 0x5fffed09bead87c0, lo: 0x378d8e63ffffffff}
	negNmax := BID_UINT128{hi: 0xdfffed09bead87c0, lo: 0x378d8e63ffffffff}
	posInf := BID_UINT128{hi: 0x7800000000000000, lo: 0}
	negInf := BID_UINT128{hi: 0xf800000000000000, lo: 0}
	const wantFlags = BID_OVERFLOW_EXCEPTION | BID_INEXACT_EXCEPTION

	tests := []struct {
		name    string
		x, y, z BID_UINT128
		rnd     int
		want    BID_UINT128
	}{
		{name: "pos_nearest_even", x: oneE34, y: posNmax, z: negNmax, rnd: BID_ROUNDING_TO_NEAREST, want: posInf},
		{name: "pos_down", x: oneE34, y: posNmax, z: negNmax, rnd: BID_ROUNDING_DOWN, want: posNmax},
		{name: "pos_up", x: oneE34, y: posNmax, z: negNmax, rnd: BID_ROUNDING_UP, want: posInf},
		{name: "pos_to_zero", x: oneE34, y: posNmax, z: negNmax, rnd: BID_ROUNDING_TO_ZERO, want: posNmax},
		{name: "pos_ties_away", x: oneE34, y: posNmax, z: negNmax, rnd: BID_ROUNDING_TIES_AWAY, want: posInf},
		{name: "neg_nearest_even", x: oneE34, y: negNmax, z: posNmax, rnd: BID_ROUNDING_TO_NEAREST, want: negInf},
		{name: "neg_down", x: oneE34, y: negNmax, z: posNmax, rnd: BID_ROUNDING_DOWN, want: negInf},
		{name: "neg_up", x: oneE34, y: negNmax, z: posNmax, rnd: BID_ROUNDING_UP, want: negNmax},
		{name: "neg_to_zero", x: oneE34, y: negNmax, z: posNmax, rnd: BID_ROUNDING_TO_ZERO, want: negNmax},
		{name: "neg_ties_away", x: oneE34, y: negNmax, z: posNmax, rnd: BID_ROUNDING_TIES_AWAY, want: negInf},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got, flags := Bid128Fma(tc.x, tc.y, tc.z, tc.rnd)
			if got != tc.want || flags != wantFlags {
				t.Fatalf("Bid128Fma = %016x:%016x/%02x, want %016x:%016x/%02x",
					got.hi, got.lo, flags, tc.want.hi, tc.want.lo, wantFlags)
			}
		})
	}
}
