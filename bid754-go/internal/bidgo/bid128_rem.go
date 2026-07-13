// Ported from: Intel bid128_rem.c and bid128_fmod.c
// Mechanical translation - all logic preserved exactly.

package bidgo

import "math"

// Bid128Rem computes the IEEE 754 remainder of x/y.
// Ported mechanically from bid128_rem.c: bid128_rem.
func Bid128Rem(x, y BID_UINT128) (BID_UINT128, uint32) {
	var P256 BID_UINT256
	var CX, CY, CX2, CQ, CR, T, CXS, P128, res BID_UINT128
	var sign_x, sign_y uint64
	var valid_y bool
	var D int64
	var exponent_x, exponent_y, diff_expon, bin_expon_cx, scale, scale0 int
	var pfpsf uint32

	// unpack arguments, check for NaN or Infinity

	sign_y, exponent_y, CY, valid_y = unpack_BID128_value(y)
	_ = sign_y

	var valid_x bool
	sign_x, exponent_x, CX, valid_x = unpack_BID128_value(x)

	if !valid_x {
		if (y.hi & SNAN_MASK64) == SNAN_MASK64 { // y is sNaN
			pfpsf |= BID_INVALID_EXCEPTION
		}
		// test if x is NaN
		if (x.hi & 0x7c00000000000000) == 0x7c00000000000000 {
			if (x.hi & SNAN_MASK64) == SNAN_MASK64 { // sNaN
				pfpsf |= BID_INVALID_EXCEPTION
			}
			res.hi = CX.hi & QUIET_MASK64
			res.lo = CX.lo
			return res, pfpsf
		}
		// x is Infinity?
		if (x.hi & 0x7800000000000000) == 0x7800000000000000 {
			// check if y is Inf.
			if (y.hi & 0x7c00000000000000) != 0x7c00000000000000 {
				// return NaN
				pfpsf |= BID_INVALID_EXCEPTION
				res.hi = 0x7c00000000000000
				res.lo = 0
				return res, pfpsf
			}
		}
		// x is 0
		if (CY.hi == 0) && (CY.lo == 0) {
			// set status flags
			pfpsf |= BID_INVALID_EXCEPTION
			// x=y=0, return NaN
			res.hi = 0x7c00000000000000
			res.lo = 0
			return res, pfpsf
		}
		if valid_y || ((y.hi & NAN_MASK64) == INFINITY_MASK64) {
			// return 0
			if (exponent_x > exponent_y) &&
				((y.hi & NAN_MASK64) != INFINITY_MASK64) {
				exponent_x = exponent_y
			}

			res.hi = sign_x | (uint64(exponent_x) << 49)
			res.lo = 0
			return res, pfpsf
		}
	}
	if !valid_y {
		// y is Inf. or NaN

		// test if y is NaN
		if (y.hi & 0x7c00000000000000) == 0x7c00000000000000 {
			if (y.hi & SNAN_MASK64) == SNAN_MASK64 { // y is sNaN
				pfpsf |= BID_INVALID_EXCEPTION
			}
			res.hi = CY.hi & QUIET_MASK64
			res.lo = CY.lo
			return res, pfpsf
		}
		// y is Infinity?
		if (y.hi & 0x7800000000000000) == 0x7800000000000000 {
			// return x
			res.hi = x.hi
			res.lo = x.lo
			return res, pfpsf
		}
		// y is 0
		// set status flags
		pfpsf |= BID_INVALID_EXCEPTION
		res.hi = 0x7c00000000000000
		res.lo = 0
		return res, pfpsf
	}

	diff_expon = exponent_x - exponent_y

	if diff_expon <= 0 {
		diff_expon = -diff_expon

		if diff_expon > 34 {
			// |x|<|y| in this case
			res = x
			return res, pfpsf
		}
		// set exponent of y to exponent_x, scale coefficient_y
		T = bid_power10_table_128[diff_expon]
		P256 = __mul_128x128_to_256(CY, T)

		if P256.w[2] != 0 || P256.w[3] != 0 {
			// |x|<|y| in this case
			res = x
			return res, pfpsf
		}

		CX2.hi = (CX.hi << 1) | (CX.lo >> 63)
		CX2.lo = CX.lo << 1
		P256_128 := BID_UINT128{lo: P256.w[0], hi: P256.w[1]}
		if __unsigned_compare_ge_128(P256_128, CX2) {
			// |x|<|y| in this case
			res = x
			return res, pfpsf
		}

		P128.lo = P256.w[0]
		P128.hi = P256.w[1]
		CQ, CR = bid___div_128_by_128(CX, P128)

		CX2.hi = (CR.hi << 1) | (CR.lo >> 63)
		CX2.lo = CR.lo << 1
		if __unsigned_compare_gt_128(CX2, P256_128) ||
			(CX2.hi == P256.w[1] && CX2.lo == P256.w[0] &&
				(CQ.lo&1) != 0) {
			CR = __sub_128_128(P256_128, CR)
			sign_x ^= 0x8000000000000000
		}

		res = very_fast_get_BID128(sign_x, exponent_x, CR)
		return res, pfpsf
	}
	// 2^64
	f64_d := math.Float32frombits(0x5f800000)

	scale0 = 38
	if CY.hi == 0 {
		scale0 = 34
	}

	for diff_expon > 0 {
		// get number of digits in CX and scale=38-digits
		// fx ~ CX
		fx_d := noFmaMulAddF32(float32(CX.hi), f64_d, float32(CX.lo))
		fx_i := math.Float32bits(fx_d)
		bin_expon_cx = int((fx_i>>23)&0xff) - 0x7f
		scale = scale0 - bid_estimate_decimal_digits[bin_expon_cx]
		// scale = 38-estimate_decimal_digits[bin_expon_cx];
		D = int64(CX.hi) - int64(bid_power10_index_binexp_128[bin_expon_cx].hi)
		if D > 0 ||
			(D == 0 && CX.lo >= bid_power10_index_binexp_128[bin_expon_cx].lo) {
			scale--
		}

		if diff_expon >= scale {
			diff_expon -= scale
		} else {
			scale = diff_expon
			diff_expon = 0
		}

		T = bid_power10_table_128[scale]
		CXS = __mul_128x128_low(CX, T)

		CQ, CX = bid___div_128_by_128(CXS, CY)

		// check for remainder == 0
		if CX.hi == 0 && CX.lo == 0 {
			res = very_fast_get_BID128(sign_x, exponent_y, CX)
			return res, pfpsf
		}
	}

	CX2.hi = (CX.hi << 1) | (CX.lo >> 63)
	CX2.lo = CX.lo << 1
	if __unsigned_compare_gt_128(CX2, CY) ||
		(CX2.hi == CY.hi && CX2.lo == CY.lo && (CQ.lo&1) != 0) {
		CX = __sub_128_128(CY, CX)
		sign_x ^= 0x8000000000000000
	}

	res = very_fast_get_BID128(sign_x, exponent_y, CX)
	return res, pfpsf
}

// Bid128Fmod computes the fmod of x/y.
// Ported mechanically from bid128_fmod.c: bid128_fmod.
func Bid128Fmod(x, y BID_UINT128) (BID_UINT128, uint32) {
	var P256 BID_UINT256
	var CX, CY, CQ, CR, T, CXS, P128, res BID_UINT128
	var sign_x, sign_y uint64
	var valid_y bool
	var D int64
	var exponent_x, exponent_y, diff_expon, bin_expon_cx, scale, scale0 int
	var pfpsf uint32

	// unpack arguments, check for NaN or Infinity

	sign_y, exponent_y, CY, valid_y = unpack_BID128_value(y)
	_ = sign_y

	var valid_x bool
	sign_x, exponent_x, CX, valid_x = unpack_BID128_value(x)

	if !valid_x {
		if (y.hi & SNAN_MASK64) == SNAN_MASK64 { // y is sNaN
			pfpsf |= BID_INVALID_EXCEPTION
		}
		// test if x is NaN
		if (x.hi & 0x7c00000000000000) == 0x7c00000000000000 {
			if (x.hi & SNAN_MASK64) == SNAN_MASK64 { // sNaN
				pfpsf |= BID_INVALID_EXCEPTION
			}
			res.hi = CX.hi & QUIET_MASK64
			res.lo = CX.lo
			return res, pfpsf
		}
		// x is Infinity?
		if (x.hi & 0x7800000000000000) == 0x7800000000000000 {
			// check if y is Inf.
			if (y.hi & 0x7c00000000000000) != 0x7c00000000000000 {
				// return NaN
				pfpsf |= BID_INVALID_EXCEPTION
				res.hi = 0x7c00000000000000
				res.lo = 0
				return res, pfpsf
			}
		}
		// x is 0
		if (CY.hi == 0) && (CY.lo == 0) {
			// set status flags
			pfpsf |= BID_INVALID_EXCEPTION
			// x=y=0, return NaN
			res.hi = 0x7c00000000000000
			res.lo = 0
			return res, pfpsf
		}
		if valid_y || ((y.hi & NAN_MASK64) == INFINITY_MASK64) {
			// return 0
			if (exponent_x > exponent_y) &&
				((y.hi & NAN_MASK64) != INFINITY_MASK64) {
				exponent_x = exponent_y
			}

			res.hi = sign_x | (uint64(exponent_x) << 49)
			res.lo = 0
			return res, pfpsf
		}
	}
	if !valid_y {
		// y is Inf. or NaN

		// test if y is NaN
		if (y.hi & 0x7c00000000000000) == 0x7c00000000000000 {
			if (y.hi & SNAN_MASK64) == SNAN_MASK64 { // y is sNaN
				pfpsf |= BID_INVALID_EXCEPTION
			}
			res.hi = CY.hi & QUIET_MASK64
			res.lo = CY.lo
			return res, pfpsf
		}
		// y is Infinity?
		if (y.hi & 0x7800000000000000) == 0x7800000000000000 {
			// return x
			res.hi = x.hi
			res.lo = x.lo
			return res, pfpsf
		}
		// y is 0
		// set status flags
		pfpsf |= BID_INVALID_EXCEPTION
		res.hi = 0x7c00000000000000
		res.lo = 0
		return res, pfpsf
	}

	diff_expon = exponent_x - exponent_y

	if diff_expon <= 0 {
		diff_expon = -diff_expon

		if diff_expon > 34 {
			// |x|<|y| in this case
			res = x
			return res, pfpsf
		}
		// set exponent of y to exponent_x, scale coefficient_y
		T = bid_power10_table_128[diff_expon]
		P256 = __mul_128x128_to_256(CY, T)

		if P256.w[2] != 0 || P256.w[3] != 0 {
			// |x|<|y| in this case
			res = x
			return res, pfpsf
		}

		if __unsigned_compare_gt_128(BID_UINT128{lo: P256.w[0], hi: P256.w[1]}, CX) {
			// |x|<|y| in this case
			res = x
			return res, pfpsf
		}

		P128.lo = P256.w[0]
		P128.hi = P256.w[1]
		_, CR = bid___div_128_by_128(CX, P128)

		res = very_fast_get_BID128(sign_x, exponent_x, CR)
		return res, pfpsf
	}
	// 2^64
	f64_d := math.Float32frombits(0x5f800000)

	scale0 = 38
	if CY.hi == 0 {
		scale0 = 34
	}

	for diff_expon > 0 {
		// get number of digits in CX and scale=38-digits
		// fx ~ CX
		fx_d := noFmaMulAddF32(float32(CX.hi), f64_d, float32(CX.lo))
		fx_i := math.Float32bits(fx_d)
		bin_expon_cx = int((fx_i>>23)&0xff) - 0x7f
		scale = scale0 - bid_estimate_decimal_digits[bin_expon_cx]
		// scale = 38-estimate_decimal_digits[bin_expon_cx];
		D = int64(CX.hi) - int64(bid_power10_index_binexp_128[bin_expon_cx].hi)
		if D > 0 ||
			(D == 0 && CX.lo >= bid_power10_index_binexp_128[bin_expon_cx].lo) {
			scale--
		}

		if diff_expon >= scale {
			diff_expon -= scale
		} else {
			scale = diff_expon
			diff_expon = 0
		}

		T = bid_power10_table_128[scale]
		CXS = __mul_128x128_low(CX, T)

		CQ, CX = bid___div_128_by_128(CXS, CY)
		_ = CQ

		// check for remainder == 0
		if CX.hi == 0 && CX.lo == 0 {
			res = very_fast_get_BID128(sign_x, exponent_y, CX)
			return res, pfpsf
		}
	}

	res = very_fast_get_BID128(sign_x, exponent_y, CX)
	return res, pfpsf
}
