// Ported from: IntelRDFPMathLib20U4/LIBRARY/src/bid32_add.c
// Version: Intel(R) Decimal Floating-Point Math Library 2.0 Update 4
//
// This file is a mechanical translation of the Intel BID library to Go.
// All logic, magic numbers, and table references are preserved exactly.

package bidgo

import (
	"math"
)

// bid32_add_pure performs BID32 addition
// Ported mechanically from Intel bid32_add.c
func bid32_add_pure(x, y uint32, rndMode int) uint32 {
	var Tmp BID_UINT128
	var S int64
	var sign_ab int64
	var SU, CB, P, Q, R uint64
	var sign_x, sign_y, coefficient_x, coefficient_y, res uint32
	var sign_a, sign_b, coefficient_a, coefficient_b uint32
	var valid_x, valid_y bool
	var exponent_x, exponent_y, bin_expon, amount, n_digits, extra_digits, rmode int
	var exponent_a, exponent_b, scale_ca, diff_dec_expon, d2 int

	sign_x, exponent_x, coefficient_x, valid_x = unpack_BID32_add(x)
	sign_y, exponent_y, coefficient_y, valid_y = unpack_BID32_add(y)

	// unpack arguments, check for NaN or Infinity
	if !valid_x {
		// x is Inf. or NaN

		// test if x is NaN
		if (x & NAN_MASK32) == NAN_MASK32 {
			res = coefficient_x & QUIET_MASK32
			return res
		}
		// x is Infinity?
		if (x & INFINITY_MASK32) == INFINITY_MASK32 {
			// check if y is Inf
			if (y & NAN_MASK32) == INFINITY_MASK32 {
				if sign_x == (y & 0x80000000) {
					res = coefficient_x
					return res
				}
				// return NaN
				res = NAN_MASK32
				return res
			}
			// check if y is NaN
			if (y & NAN_MASK32) == NAN_MASK32 {
				res = coefficient_y & QUIET_MASK32
				return res
			}
			// otherwise return +/-Inf
			res = coefficient_x
			return res
		}
		// x is 0
		if ((y & INFINITY_MASK32) != INFINITY_MASK32) && coefficient_y != 0 {
			if exponent_y <= exponent_x {
				res = y
				return res
			}
		}
	}
	if !valid_y {
		// y is Inf. or NaN?
		if (y & INFINITY_MASK32) == INFINITY_MASK32 {
			res = coefficient_y & QUIET_MASK32
			return res
		}
		// y is 0
		if coefficient_x == 0 { // x==0
			if exponent_x <= exponent_y {
				res = uint32(exponent_x) << 23
			} else {
				res = uint32(exponent_y) << 23
			}
			if sign_x == sign_y {
				res |= sign_x
			}
			if rndMode == BID_ROUNDING_DOWN && sign_x != sign_y {
				res |= 0x80000000
			}
			return res
		} else if exponent_y >= exponent_x {
			res = x
			return res
		}
	}

	// sort arguments by exponent
	if exponent_x < exponent_y {
		sign_a = sign_y
		exponent_a = exponent_y
		coefficient_a = coefficient_y
		sign_b = sign_x
		exponent_b = exponent_x
		coefficient_b = coefficient_x
	} else {
		sign_a = sign_x
		exponent_a = exponent_x
		coefficient_a = coefficient_x
		sign_b = sign_y
		exponent_b = exponent_y
		coefficient_b = coefficient_y
	}

	// exponent difference
	diff_dec_expon = exponent_a - exponent_b

	if diff_dec_expon > MAX_FORMAT_DIGITS_32 {
		tempx := float64(coefficient_a)
		bin_expon = int((math.Float64bits(tempx)&MASK_BINARY_EXPONENT)>>52) - 0x3ff

		scale_ca = bid_estimate_decimal_digits[bin_expon]

		d2 = 16 - scale_ca
		if diff_dec_expon > d2 {
			diff_dec_expon = d2
			exponent_b = exponent_a - diff_dec_expon
		}
	}

	sign_ab = int64(sign_a^sign_b) << 32
	sign_ab = sign_ab >> 63
	CB = uint64(int64(coefficient_b)+sign_ab) ^ uint64(sign_ab)

	SU = uint64(coefficient_a) * bid_power10_table_128[diff_dec_expon].w[0]
	S = int64(SU) + int64(CB)

	if S < 0 {
		sign_a ^= 0x80000000
		S = -S
	}
	P = uint64(S)

	if P == 0 {
		sign_a = 0
		if rndMode == BID_ROUNDING_DOWN {
			sign_a = 0x80000000
		}
		if coefficient_a == 0 {
			sign_a = sign_x
		}
		n_digits = 0
	} else {
		tempx := float64(P)
		bin_expon = int((math.Float64bits(tempx)&MASK_BINARY_EXPONENT)>>52) - 0x3ff
		n_digits = bid_estimate_decimal_digits[bin_expon]
		if P >= bid_power10_table_128[n_digits].w[0] {
			n_digits++
		}
	}

	if n_digits <= MAX_FORMAT_DIGITS_32 {
		res = get_BID32(sign_a, exponent_b, P, rndMode)
		return res
	}

	extra_digits = n_digits - 7

	rmode = rndMode
	if sign_a != 0 && uint(rmode-1) < 2 {
		rmode = 3 - rmode
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

	res = get_BID32(sign_a, exponent_b+extra_digits, Q, rndMode)

	return res
}

// unpack_BID32_add unpacks a BID32 value for addition
// Returns sign, exponent, coefficient, and valid flag
func unpack_BID32_add(x uint32) (sign uint32, exponent int, coefficient uint32, valid bool) {
	sign = x & 0x80000000

	if (x & SPECIAL_ENCODING_MASK32) == SPECIAL_ENCODING_MASK32 {
		// special encodings
		if (x & INFINITY_MASK32) == INFINITY_MASK32 {
			coefficient = x & 0xfe0fffff
			if (x & 0x000fffff) >= 1000000 {
				coefficient = x & 0xfe000000
			}
			if (x & NAN_MASK32) == INFINITY_MASK32 {
				coefficient = x & SINFINITY_MASK32
			}
			return sign, 0, coefficient, false // NaN or Infinity
		}
		// coefficient
		coefficient = (x & SMALL_COEFF_MASK32) | LARGE_COEFF_HIGH_BIT32
		// check for non-canonical values
		if coefficient >= 10000000 {
			coefficient = 0
		}
		// get exponent
		exponent = int((x >> 21) & EXPONENT_MASK32)
		return sign, exponent, coefficient, coefficient != 0
	}
	// exponent
	exponent = int((x >> 23) & EXPONENT_MASK32)
	// coefficient
	coefficient = x & LARGE_COEFF_MASK32

	return sign, exponent, coefficient, coefficient != 0
}

// bid32_sub_pure performs BID32 subtraction
// Ported mechanically from Intel bid32_sub.c
func bid32_sub_pure(x, y uint32, rndMode int) uint32 {
	// negate y if it's not NaN
	if (y & NAN_MASK32) != NAN_MASK32 {
		y ^= 0x80000000
	}

	return bid32_add_pure(x, y, rndMode)
}
