// Ported from: Intel bid128_mul.c
// Mechanical translation - all logic preserved exactly.
// Note: bid128_mul delegates to bid128_fma(y, x, 0) for non-zero cases.

package bidgo

// bid128 uses the same mask values as bid64 (they operate on w[1] which is 64-bit)
// MASK_NAN64 = 0x7c00000000000000
// MASK_SIGN6464 = 0x8000000000000000
// MASK_INF64 = 0x7800000000000000 (same as INFINITY_MASK64)
// MASK_COEFF128128 = 0x0001ffffffffffff
// MASK_EXP128128 = 0x7ffe000000000000

// Bid128Mul is ported mechanically from bid128_mul.c: bid128_mul.
func Bid128Mul(x, y BID_UINT128, rnd_mode int) (BID_UINT128, uint32) {
	var res BID_UINT128
	var x_sign, y_sign, p_sign uint64
	var x_exp, y_exp, p_exp uint64
	var true_p_exp int
	var C1, C2 BID_UINT128
	var pfpsf uint32

	// skip cases where at least one operand is NaN or infinity
	if !(((x.w[1] & NAN_MASK64) == NAN_MASK64) ||
		((y.w[1] & NAN_MASK64) == NAN_MASK64) ||
		((x.w[1] & NAN_MASK64) == INFINITY_MASK64) ||
		((y.w[1] & NAN_MASK64) == INFINITY_MASK64)) {

		x_sign = x.w[1] & MASK_SIGN64
		C1.w[1] = x.w[1] & MASK_COEFF128
		C1.w[0] = x.w[0]
		if (x.w[1] & 0x6000000000000000) == 0x6000000000000000 {
			x_exp = (x.w[1] << 2) & MASK_EXP128
			C1.w[1] = 0
			C1.w[0] = 0
		} else {
			x_exp = x.w[1] & MASK_EXP128
			if C1.w[1] > 0x0001ed09bead87c0 ||
				(C1.w[1] == 0x0001ed09bead87c0 && C1.w[0] > 0x378d8e63ffffffff) {
				C1.w[1] = 0
				C1.w[0] = 0
			}
		}
		y_sign = y.w[1] & MASK_SIGN64
		C2.w[1] = y.w[1] & MASK_COEFF128
		C2.w[0] = y.w[0]
		if (y.w[1] & 0x6000000000000000) == 0x6000000000000000 {
			y_exp = (y.w[1] << 2) & MASK_EXP128
			C2.w[1] = 0
			C2.w[0] = 0
		} else {
			y_exp = y.w[1] & MASK_EXP128
			if C2.w[1] > 0x0001ed09bead87c0 ||
				(C2.w[1] == 0x0001ed09bead87c0 && C2.w[0] > 0x378d8e63ffffffff) {
				C2.w[1] = 0
				C2.w[0] = 0
			}
		}
		p_sign = x_sign ^ y_sign

		true_p_exp = int(x_exp>>49) - 6176 + int(y_exp>>49) - 6176
		if true_p_exp < -6176 {
			p_exp = 0
		} else if true_p_exp > 6111 {
			p_exp = uint64(6111+6176) << 49
		} else {
			p_exp = uint64(true_p_exp+6176) << 49
		}

		if (C1.w[1] == 0 && C1.w[0] == 0) || (C2.w[1] == 0 && C2.w[0] == 0) {
			res.w[1] = p_sign | p_exp
			res.w[0] = 0
			return res, pfpsf
		}
	}

	// For non-zero, non-special cases: delegate to bid128_fma(y, x, 0)
	z := BID_UINT128{w: [2]uint64{0x0000000000000000, 0x5ffe000000000000}}
	return Bid128Fma(y, x, z, rnd_mode)
}

// Bid128Fma is in bid128_fma.go
