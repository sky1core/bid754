package bidgo

import "math"

// Bid64Rem returns the IEEE 754 remainder of x/y and the status flags.
// Ported mechanically from Intel bid64_rem.c.
func Bid64Rem(x, y uint64) (uint64, uint32) {
	var CY BID_UINT128
	var sign_x, sign_y, coefficient_x, coefficient_y, res uint64
	var Q, R, R2, T uint64
	var valid_y, valid_x bool
	var exponent_x, exponent_y, bin_expon, e_scale int
	var digits_x, diff_expon int
	var pfpsf uint32

	sign_y, exponent_y, coefficient_y, valid_y = unpack_BID64(y)
	sign_x, exponent_x, coefficient_x, valid_x = unpack_BID64(x)
	_ = sign_y

	// unpack arguments, check for NaN or Infinity
	if !valid_x {
		// x is Inf. or NaN or 0
		if (y & SNAN_MASK64) == SNAN_MASK64 {
			pfpsf |= BID_INVALID_EXCEPTION
		}

		// test if x is NaN
		if (x & NAN_MASK64) == NAN_MASK64 {
			if (x & SNAN_MASK64) == SNAN_MASK64 {
				pfpsf |= BID_INVALID_EXCEPTION
			}
			res = coefficient_x & QUIET_MASK64
			return res, pfpsf
		}
		// x is Infinity?
		if (x & INFINITY_MASK64) == INFINITY_MASK64 {
			if (y & NAN_MASK64) != NAN_MASK64 {
				pfpsf |= BID_INVALID_EXCEPTION
				res = NAN_MASK64
				return res, pfpsf
			}
		}
		// x is 0
		// return x if y != 0
		if ((y & INFINITY_MASK64) < INFINITY_MASK64) && coefficient_y != 0 {
			if (y & SPECIAL_ENCODING_MASK64) == SPECIAL_ENCODING_MASK64 {
				exponent_y = int((y >> 51) & 0x3ff)
			} else {
				exponent_y = int((y >> 53) & 0x3ff)
			}

			if exponent_y < exponent_x {
				exponent_x = exponent_y
			}

			x = uint64(exponent_x)
			x <<= 53

			res = x | sign_x
			return res, pfpsf
		}
	}
	if !valid_y {
		// y is Inf. or NaN

		// test if y is NaN
		if (y & NAN_MASK64) == NAN_MASK64 {
			if (y & SNAN_MASK64) == SNAN_MASK64 {
				pfpsf |= BID_INVALID_EXCEPTION
			}
			res = coefficient_y & QUIET_MASK64
			return res, pfpsf
		}
		// y is Infinity?
		if (y & INFINITY_MASK64) == INFINITY_MASK64 {
			res = very_fast_get_BID64(sign_x, exponent_x, coefficient_x)
			return res, pfpsf
		}
		// y is 0, return NaN
		pfpsf |= BID_INVALID_EXCEPTION
		res = NAN_MASK64
		return res, pfpsf
	}

	diff_expon = exponent_x - exponent_y
	if diff_expon <= 0 {
		diff_expon = -diff_expon

		if diff_expon > 16 {
			// |x|<|y| in this case
			res = x
			return res, pfpsf
		}
		// set exponent of y to exponent_x, scale coefficient_y
		T = bid_power10_table_128[diff_expon].w[0]
		CY = __mul_64x64_to_128(coefficient_y, T)

		if CY.w[1] != 0 || CY.w[0] > (coefficient_x<<1) {
			res = x
			return res, pfpsf
		}

		Q = coefficient_x / CY.w[0]
		R = coefficient_x - Q*CY.w[0]

		R2 = R + R
		if R2 > CY.w[0] || (R2 == CY.w[0] && (Q&1) != 0) {
			R = CY.w[0] - R
			sign_x ^= 0x8000000000000000
		}

		res = very_fast_get_BID64(sign_x, exponent_x, R)
		return res, pfpsf
	}

	for diff_expon > 0 {
		// get number of digits in coeff_x
		tempx := math.Float32bits(float32(coefficient_x))
		bin_expon = int((tempx>>23)&0xff) - 0x7f
		digits_x = bid_estimate_decimal_digits[bin_expon]
		// will not use this test, dividend will have 18 or 19 digits
		//if(coefficient_x >= bid_power10_table_128[digits_x].w[0])
		//      digits_x++

		e_scale = 18 - digits_x
		if diff_expon >= e_scale {
			diff_expon -= e_scale
		} else {
			e_scale = diff_expon
			diff_expon = 0
		}

		// scale dividend to 18 or 19 digits
		coefficient_x *= bid_power10_table_128[e_scale].w[0]

		// quotient
		Q = coefficient_x / coefficient_y
		// remainder
		coefficient_x -= Q * coefficient_y

		// check for remainder == 0
		if coefficient_x == 0 {
			res = very_fast_get_BID64_small_mantissa(sign_x, exponent_y, 0)
			return res, pfpsf
		}
	}

	R2 = coefficient_x + coefficient_x
	if R2 > coefficient_y || (R2 == coefficient_y && (Q&1) != 0) {
		coefficient_x = coefficient_y - coefficient_x
		sign_x ^= 0x8000000000000000
	}

	res = very_fast_get_BID64(sign_x, exponent_y, coefficient_x)
	return res, pfpsf
}
