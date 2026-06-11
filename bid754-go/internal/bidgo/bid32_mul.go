// Ported from: IntelRDFPMathLib20U4/LIBRARY/src/bid32_mul.c
// Version: Intel(R) Decimal Floating-Point Math Library 2.0 Update 4
//
// This file is a mechanical translation of the Intel BID library to Go.
// All logic, magic numbers, and table references are preserved exactly.

package bidgo

import (
	"math"
)

// bid32_mul_pure performs BID32 multiplication
// Ported mechanically from Intel bid32_mul.c
func bid32_mul_pure(x, y uint32, rndMode int) uint32 {
	var Tmp BID_UINT128
	var P, Q, R uint64
	var sign_x, sign_y, coefficient_x, coefficient_y, res uint32
	var valid_x, valid_y bool
	var exponent_x, exponent_y, bin_expon_p, amount, n_digits, extra_digits, rmode int

	sign_x, exponent_x, coefficient_x, valid_x = unpack_BID32_add(x)
	sign_y, exponent_y, coefficient_y, valid_y = unpack_BID32_add(y)

	// unpack arguments, check for NaN or Infinity
	if !valid_x {
		// x is Inf. or NaN

		// test if x is NaN
		if (x & NAN_MASK32) == NAN_MASK32 {
			return coefficient_x & QUIET_MASK32
		}
		// x is Infinity?
		if (x & INFINITY_MASK32) == INFINITY_MASK32 {
			// check if y is 0
			if ((y & INFINITY_MASK32) != INFINITY_MASK32) && coefficient_y == 0 {
				// y==0 , return NaN
				return NAN_MASK32
			}
			// check if y is NaN
			if (y & NAN_MASK32) == NAN_MASK32 {
				// y==NaN , return NaN
				return coefficient_y & QUIET_MASK32
			}
			// otherwise return +/-Inf
			return ((x ^ y) & 0x80000000) | INFINITY_MASK32
		}
		// x is 0
		if (y & INFINITY_MASK32) != INFINITY_MASK32 {
			if (y & SPECIAL_ENCODING_MASK32) == SPECIAL_ENCODING_MASK32 {
				exponent_y = int((y >> 21) & 0xff)
			} else {
				exponent_y = int((y >> 23) & 0xff)
			}
			sign_y = y & 0x80000000

			exponent_x += exponent_y - DECIMAL_EXPONENT_BIAS_32
			if exponent_x > DECIMAL_MAX_EXPON_32 {
				exponent_x = DECIMAL_MAX_EXPON_32
			} else if exponent_x < 0 {
				exponent_x = 0
			}
			return (sign_x ^ sign_y) | (uint32(exponent_x) << 23)
		}
	}
	if !valid_y {
		// y is Inf. or NaN

		// test if y is NaN
		if (y & NAN_MASK32) == NAN_MASK32 {
			return coefficient_y & QUIET_MASK32
		}
		// y is Infinity?
		if (y & INFINITY_MASK32) == INFINITY_MASK32 {
			// check if x is 0
			if coefficient_x == 0 {
				// x==0, return NaN
				return NAN_MASK32
			}
			// otherwise return +/-Inf
			return ((x ^ y) & 0x80000000) | INFINITY_MASK32
		}
		// y is 0
		exponent_x += exponent_y - DECIMAL_EXPONENT_BIAS_32
		if exponent_x > DECIMAL_MAX_EXPON_32 {
			exponent_x = DECIMAL_MAX_EXPON_32
		} else if exponent_x < 0 {
			exponent_x = 0
		}
		return (sign_x ^ sign_y) | (uint32(exponent_x) << 23)
	}

	P = uint64(coefficient_x) * uint64(coefficient_y)

	//--- get number of bits in C64 ---
	// version 2 (original)
	tempx := float64(P)
	bin_expon_p = int((math.Float64bits(tempx)&MASK_BINARY_EXPONENT)>>52) - 0x3ff
	n_digits = bid_estimate_decimal_digits[bin_expon_p]
	if P >= bid_power10_table_128[n_digits].w[0] {
		n_digits++
	}

	exponent_x += exponent_y - DECIMAL_EXPONENT_BIAS_32

	if n_digits <= 7 {
		extra_digits = 0
	} else {
		extra_digits = n_digits - 7
	}

	exponent_x += extra_digits

	if extra_digits == 0 {
		res = get_BID32(sign_x^sign_y, exponent_x, P, rndMode)
		return res
	}

	rmode = rndMode
	if (sign_x^sign_y) != 0 && uint(rmode-1) < 2 {
		rmode = 3 - rmode
	}

	if exponent_x < 0 {
		rmode = 3 // RZ
	}

	// add a constant to P, depending on rounding mode
	// 0.5*10^(digits_p - 16) for round-to-nearest
	P += bid_round_const_table[rmode][extra_digits]
	Tmp = __mul_64x64_to_128(P, bid_reciprocals10_64[extra_digits])

	// now get P/10^extra_digits: shift Q_high right by M[extra_digits]-64
	amount = bid_short_recip_scale[extra_digits]
	Q = Tmp.w[1] >> uint(amount)

	// remainder
	R = P - Q*bid_power10_table_128[extra_digits].w[0]

	if rmode == 0 { // BID_ROUNDING_TO_NEAREST
		if R == 0 {
			Q &= 0xfffffffe
		}
	}

	// DECIMAL_TINY_DETECTION_AFTER_ROUNDING
	if exponent_x == -1 && Q == 9999999 && rndMode != BID_ROUNDING_TO_ZERO {
		rmode = rndMode
		if (sign_x^sign_y) != 0 && uint(rmode-1) < 2 {
			rmode = 3 - rmode
		}

		if (R != 0 && rmode == BID_ROUNDING_UP) ||
			((rmode&3) == 0 && R+R >= bid_power10_table_128[extra_digits].w[0]) {
			res = very_fast_get_BID32(sign_x^sign_y, 0, 1000000)
			return res
		}
	}

	var uf_pfpsf uint32
	res = get_BID32_UF(sign_x^sign_y, exponent_x, Q, uint32(R), rndMode, &uf_pfpsf)

	return res
}
