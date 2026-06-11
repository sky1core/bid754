// Ported from: IntelRDFPMathLib20U4/LIBRARY/src/Bid64Mul.c
// Version: Intel(R) Decimal Floating-Point Math Library 2.0 Update 4
//
// This file is a mechanical translation of the Intel BID library to Go.
// All logic, magic numbers, and table references are preserved exactly.

package bidgo

import (
	"math"
	"math/bits"
)

// Bid64Mul multiplies x and y
// Ported from Bid64Mul in Bid64Mul.c (line-by-line mechanical translation)
func Bid64Mul(x, y uint64, rndMode int) uint64 {
	var P, C128, Q_high, Q_low BID_UINT128
	var sign_x, sign_y, coefficient_x, coefficient_y uint64
	var C64, remainder_h, res uint64
	var valid_x, valid_y bool
	var extra_digits, exponent_x, exponent_y, bin_expon_cx, bin_expon_cy int
	var bin_expon_product int
	var rmode, digits_p, bp, amount, final_exponent, round_up int

	sign_x, exponent_x, coefficient_x, valid_x = unpack_BID64(x)
	sign_y, exponent_y, coefficient_y, valid_y = unpack_BID64(y)

	// unpack arguments, check for NaN or Infinity
	if !valid_x {
		if (y & SNAN_MASK64) == SNAN_MASK64 {
			// C sets BID_INVALID_EXCEPTION here; flagless API preserves the branch structure only.
		}
		// x is Inf. or NaN

		// test if x is NaN
		if (x & NAN_MASK64) == NAN_MASK64 {
			if (x & SNAN_MASK64) == SNAN_MASK64 {
				// C sets BID_INVALID_EXCEPTION here; flagless API preserves the branch structure only.
			}
			return coefficient_x & QUIET_MASK64
		}
		// x is Infinity?
		if (x & INFINITY_MASK64) == INFINITY_MASK64 {
			// check if y is 0
			if ((y & INFINITY_MASK64) != INFINITY_MASK64) && coefficient_y == 0 {
				// y==0, return NaN
				return NAN_MASK64
			}
			// check if y is NaN
			if (y & NAN_MASK64) == NAN_MASK64 {
				// y==NaN, return NaN
				return coefficient_y & QUIET_MASK64
			}
			// otherwise return +/-Inf
			return ((x ^ y) & 0x8000000000000000) | INFINITY_MASK64
		}
		// x is 0
		if (y & INFINITY_MASK64) != INFINITY_MASK64 {
			if (y & SPECIAL_ENCODING_MASK64) == SPECIAL_ENCODING_MASK64 {
				exponent_y = int((uint32(y>>51) & 0x3ff))
			} else {
				exponent_y = int((uint32(y>>53) & 0x3ff))
			}
			sign_y = y & 0x8000000000000000

			exponent_x += exponent_y - DECIMAL_EXPONENT_BIAS
			if exponent_x > DECIMAL_MAX_EXPON_64 {
				exponent_x = DECIMAL_MAX_EXPON_64
			} else if exponent_x < 0 {
				exponent_x = 0
			}
			return (sign_x ^ sign_y) | (uint64(exponent_x) << 53)
		}
	}

	if !valid_y {
		// y is Inf. or NaN

		// test if y is NaN
		if (y & NAN_MASK64) == NAN_MASK64 {
			if (y & SNAN_MASK64) == SNAN_MASK64 {
				// C sets BID_INVALID_EXCEPTION here; flagless API preserves the branch structure only.
			}
			return coefficient_y & QUIET_MASK64
		}
		// y is Infinity?
		if (y & INFINITY_MASK64) == INFINITY_MASK64 {
			// check if x is 0
			if coefficient_x == 0 {
				// x==0, return NaN
				return NAN_MASK64
			}
			// otherwise return +/-Inf
			return ((x ^ y) & 0x8000000000000000) | INFINITY_MASK64
		}
		// y is 0
		exponent_x += exponent_y - DECIMAL_EXPONENT_BIAS
		if exponent_x > DECIMAL_MAX_EXPON_64 {
			exponent_x = DECIMAL_MAX_EXPON_64
		} else if exponent_x < 0 {
			exponent_x = 0
		}
		return (sign_x ^ sign_y) | (uint64(exponent_x) << 53)
	}

	// --- get number of bits in the coefficients of x and y ---
	// version 2 (original)
	tempx := math.Float64bits(float64(coefficient_x))
	bin_expon_cx = int((tempx & MASK_BINARY_EXPONENT) >> 52)
	tempy := math.Float64bits(float64(coefficient_y))
	bin_expon_cy = int((tempy & MASK_BINARY_EXPONENT) >> 52)

	// magnitude estimate for coefficient_x*coefficient_y is
	//        2^(unbiased_bin_expon_cx + unbiased_bin_expon_cx)
	bin_expon_product = bin_expon_cx + bin_expon_cy

	// check if coefficient_x*coefficient_y<2^(10*k+3)
	// equivalent to unbiased_bin_expon_cx + unbiased_bin_expon_cx < 10*k+1
	if bin_expon_product < UPPER_EXPON_LIMIT+2*BINARY_EXPONENT_BIAS {
		// easy multiply
		C64 = coefficient_x * coefficient_y

		res = get_BID64_small_mantissa(sign_x^sign_y,
			exponent_x+exponent_y-DECIMAL_EXPONENT_BIAS, C64, rndMode)
		return res
	}

	// get 128-bit product: coefficient_x*coefficient_y
	P = __mul_64x64_to_128(coefficient_x, coefficient_y)

	// tighten binary range of P: leading bit is 2^bp
	// unbiased_bin_expon_product <= bp <= unbiased_bin_expon_product+1
	bin_expon_product -= 2 * BINARY_EXPONENT_BIAS

	bp = __tight_bin_range_128(P, bin_expon_product)

	// get number of decimal digits in the product
	digits_p = bid_estimate_decimal_digits[bp]
	if !__unsigned_compare_gt_128(bid_power10_table_128[digits_p], P) {
		digits_p++ // if bid_power10_table_128[digits_p] <= P
	}

	// determine number of decimal digits to be rounded out
	extra_digits = digits_p - MAX_FORMAT_DIGITS
	final_exponent = exponent_x + exponent_y + extra_digits - DECIMAL_EXPONENT_BIAS

	rmode = rndMode
	if (sign_x^sign_y) != 0 && uint(rmode-1) < 2 {
		rmode = 3 - rmode
	}

	round_up = 0
	if uint(final_exponent) >= 3*256 {
		if final_exponent < 0 {
			// underflow
			if final_exponent+16 < 0 {
				res = sign_x ^ sign_y
				if rmode == BID_ROUNDING_UP {
					res |= 1
				}
				return res
			}

			extra_digits -= final_exponent
			final_exponent = 0

			if extra_digits > 17 {
				Q_high, Q_low = __mul_128x128_full(P, bid_reciprocals10_128[16])

				amount = bid_recip_scale[16]
				P = __shr_128(Q_high, uint(amount))

				// get sticky bits
				amount2 := 64 - amount
				remainder_h = (^uint64(0)) >> uint(amount2)
				remainder_h = remainder_h & Q_high.w[0]

				extra_digits -= 16
				if remainder_h != 0 || (Q_low.w[1] > bid_reciprocals10_128[16].w[1] ||
					(Q_low.w[1] == bid_reciprocals10_128[16].w[1] &&
						Q_low.w[0] >= bid_reciprocals10_128[16].w[0])) {
					round_up = 1
					P.w[0] = (P.w[0] << 3) + (P.w[0] << 1)
					P.w[0] |= 1
					extra_digits++
				}
			}
		} else {
			res = fast_get_BID64_check_OF(sign_x^sign_y, final_exponent,
				1000000000000000, rndMode)
			return res
		}
	}

	if extra_digits > 0 {
		// will divide by 10^(digits_p - 16)

		// add a constant to P, depending on rounding mode
		// 0.5*10^(digits_p - 16) for round-to-nearest
		P = __add_128_64(P, bid_round_const_table[rmode][extra_digits])

		// get P*(2^M[extra_digits])/10^extra_digits
		Q_high, Q_low = __mul_128x128_full(P, bid_reciprocals10_128[extra_digits])

		// now get P/10^extra_digits: shift Q_high right by M[extra_digits]-128
		amount = bid_recip_scale[extra_digits]
		C128 = __shr_128(Q_high, uint(amount))

		C64 = C128.w[0]

		if rmode == 0 { // BID_ROUNDING_TO_NEAREST
			if (C64&1) != 0 && round_up == 0 {
				// check whether fractional part of initial_P/10^extra_digits
				// is exactly .5

				// get remainder
				remainder_h = Q_high.w[0] << (64 - uint(amount))

				// test whether fractional part is 0
				if remainder_h == 0 &&
					(Q_low.w[1] < bid_reciprocals10_128[extra_digits].w[1] ||
						(Q_low.w[1] == bid_reciprocals10_128[extra_digits].w[1] &&
							Q_low.w[0] < bid_reciprocals10_128[extra_digits].w[0])) {
					C64--
				}
			}
		}

		// convert to BID and return
		res = fast_get_BID64_check_OF(sign_x^sign_y, final_exponent, C64, rndMode)
		return res
	}

	// go to convert_format and exit
	C64 = P.w[0]
	res = get_BID64(sign_x^sign_y,
		exponent_x+exponent_y-DECIMAL_EXPONENT_BIAS, C64, rndMode)
	return res
}

// Bid64MulWithFlags multiplies x and y, returning result and status flags
// Ported from Intel bid64_mul.c with flag tracking
func Bid64MulWithFlags(x, y uint64, rndMode int) (uint64, uint32) {
	var P, C128, Q_high, Q_low BID_UINT128
	var sign_x, sign_y, coefficient_x, coefficient_y uint64
	var C64, remainder_h, res uint64
	var valid_x, valid_y bool
	var extra_digits, exponent_x, exponent_y, bin_expon_cx, bin_expon_cy int
	var bin_expon_product int
	var rmode, digits_p, bp, amount, final_exponent, round_up int
	var pfpsf uint32
	var uf_status uint32

	sign_x, exponent_x, coefficient_x, valid_x = unpack_BID64(x)
	sign_y, exponent_y, coefficient_y, valid_y = unpack_BID64(y)

	// unpack arguments, check for NaN or Infinity
	if !valid_x {
		// check for SNaN in y first (Intel order)
		if (y & SNAN_MASK64) == SNAN_MASK64 {
			pfpsf |= BID_INVALID_EXCEPTION
		}

		// x is Inf. or NaN
		// test if x is NaN
		if (x & NAN_MASK64) == NAN_MASK64 {
			if (x & SNAN_MASK64) == SNAN_MASK64 {
				pfpsf |= BID_INVALID_EXCEPTION
			}
			return coefficient_x & QUIET_MASK64, pfpsf
		}
		// x is Infinity?
		if (x & INFINITY_MASK64) == INFINITY_MASK64 {
			// check if y is 0
			if ((y & INFINITY_MASK64) != INFINITY_MASK64) && coefficient_y == 0 {
				pfpsf |= BID_INVALID_EXCEPTION
				// y==0, return NaN
				return NAN_MASK64, pfpsf
			}
			// check if y is NaN
			if (y & NAN_MASK64) == NAN_MASK64 {
				// y==NaN, return NaN
				return coefficient_y & QUIET_MASK64, pfpsf
			}
			// otherwise return +/-Inf
			return ((x ^ y) & 0x8000000000000000) | INFINITY_MASK64, pfpsf
		}
		// x is 0
		if (y & INFINITY_MASK64) != INFINITY_MASK64 {
			if (y & SPECIAL_ENCODING_MASK64) == SPECIAL_ENCODING_MASK64 {
				exponent_y = int((uint32(y>>51) & 0x3ff))
			} else {
				exponent_y = int((uint32(y>>53) & 0x3ff))
			}
			sign_y = y & 0x8000000000000000

			exponent_x += exponent_y - DECIMAL_EXPONENT_BIAS
			if exponent_x > DECIMAL_MAX_EXPON_64 {
				exponent_x = DECIMAL_MAX_EXPON_64
			} else if exponent_x < 0 {
				exponent_x = 0
			}
			return (sign_x ^ sign_y) | (uint64(exponent_x) << 53), pfpsf
		}
	}

	if !valid_y {
		// y is Inf. or NaN

		// test if y is NaN
		if (y & NAN_MASK64) == NAN_MASK64 {
			if (y & SNAN_MASK64) == SNAN_MASK64 {
				pfpsf |= BID_INVALID_EXCEPTION
			}
			return coefficient_y & QUIET_MASK64, pfpsf
		}
		// y is Infinity?
		if (y & INFINITY_MASK64) == INFINITY_MASK64 {
			// check if x is 0
			if coefficient_x == 0 {
				pfpsf |= BID_INVALID_EXCEPTION
				// x==0, return NaN
				return NAN_MASK64, pfpsf
			}
			// otherwise return +/-Inf
			return ((x ^ y) & 0x8000000000000000) | INFINITY_MASK64, pfpsf
		}
		// y is 0
		exponent_x += exponent_y - DECIMAL_EXPONENT_BIAS
		if exponent_x > DECIMAL_MAX_EXPON_64 {
			exponent_x = DECIMAL_MAX_EXPON_64
		} else if exponent_x < 0 {
			exponent_x = 0
		}
		return (sign_x ^ sign_y) | (uint64(exponent_x) << 53), pfpsf
	}

	// --- get number of bits in the coefficients of x and y ---
	tempx := math.Float64bits(float64(coefficient_x))
	bin_expon_cx = int((tempx & MASK_BINARY_EXPONENT) >> 52)
	tempy := math.Float64bits(float64(coefficient_y))
	bin_expon_cy = int((tempy & MASK_BINARY_EXPONENT) >> 52)

	bin_expon_product = bin_expon_cx + bin_expon_cy

	if bin_expon_product < UPPER_EXPON_LIMIT+2*BINARY_EXPONENT_BIAS {
		// easy multiply
		C64 = coefficient_x * coefficient_y
		res = get_BID64_small_mantissa_flags(sign_x^sign_y,
			exponent_x+exponent_y-DECIMAL_EXPONENT_BIAS, C64, rndMode, &pfpsf)
		return res, pfpsf
	}

	// get 128-bit product: coefficient_x*coefficient_y
	P = __mul_64x64_to_128(coefficient_x, coefficient_y)

	bin_expon_product -= 2 * BINARY_EXPONENT_BIAS
	bp = __tight_bin_range_128(P, bin_expon_product)

	digits_p = bid_estimate_decimal_digits[bp]
	if !__unsigned_compare_gt_128(bid_power10_table_128[digits_p], P) {
		digits_p++
	}

	extra_digits = digits_p - MAX_FORMAT_DIGITS
	final_exponent = exponent_x + exponent_y + extra_digits - DECIMAL_EXPONENT_BIAS

	rmode = rndMode
	if (sign_x^sign_y) != 0 && uint(rmode-1) < 2 {
		rmode = 3 - rmode
	}

	round_up = 0
	if uint(final_exponent) >= 3*256 {
		if final_exponent < 0 {
			// underflow
			if final_exponent+16 < 0 {
				res = sign_x ^ sign_y
				pfpsf |= BID_UNDERFLOW_EXCEPTION | BID_INEXACT_EXCEPTION
				if rmode == BID_ROUNDING_UP {
					res |= 1
				}
				return res, pfpsf
			}

			uf_status = BID_UNDERFLOW_EXCEPTION
			extra_digits -= final_exponent
			final_exponent = 0

			if extra_digits > 17 {
				Q_high, Q_low = __mul_128x128_full(P, bid_reciprocals10_128[16])

				amount = bid_recip_scale[16]
				P = __shr_128(Q_high, uint(amount))

				amount2 := 64 - amount
				remainder_h = (^uint64(0)) >> uint(amount2)
				remainder_h = remainder_h & Q_high.w[0]

				extra_digits -= 16
				if remainder_h != 0 || (Q_low.w[1] > bid_reciprocals10_128[16].w[1] ||
					(Q_low.w[1] == bid_reciprocals10_128[16].w[1] &&
						Q_low.w[0] >= bid_reciprocals10_128[16].w[0])) {
					round_up = 1
					pfpsf |= BID_UNDERFLOW_EXCEPTION | BID_INEXACT_EXCEPTION
					P.w[0] = (P.w[0] << 3) + (P.w[0] << 1)
					P.w[0] |= 1
					extra_digits++
				}
			}
		} else {
			// overflow case
			res, flags := fast_get_BID64_check_OF_flags(sign_x^sign_y, final_exponent,
				1000000000000000, rndMode)
			pfpsf |= flags
			return res, pfpsf
		}
	}

	if extra_digits > 0 {
		P = __add_128_64(P, bid_round_const_table[rmode][extra_digits])
		Q_high, Q_low = __mul_128x128_full(P, bid_reciprocals10_128[extra_digits])
		amount = bid_recip_scale[extra_digits]
		C128 = __shr_128(Q_high, uint(amount))
		C64 = C128.w[0]

		if rmode == 0 { // BID_ROUNDING_TO_NEAREST
			if (C64&1) != 0 && round_up == 0 {
				remainder_h = Q_high.w[0] << (64 - uint(amount))
				if remainder_h == 0 &&
					(Q_low.w[1] < bid_reciprocals10_128[extra_digits].w[1] ||
						(Q_low.w[1] == bid_reciprocals10_128[extra_digits].w[1] &&
							Q_low.w[0] < bid_reciprocals10_128[extra_digits].w[0])) {
					C64--
				}
			}
		}

		// Intel status flag logic (bid64_mul.c lines 312-349)
		// Start with INEXACT, then check if actually exact
		status := uint32(BID_INEXACT_EXCEPTION) | uf_status
		remainder_h = Q_high.w[0] << (64 - uint(amount))

		switch rmode {
		case BID_ROUNDING_TO_NEAREST, BID_ROUNDING_TIES_AWAY:
			// test whether fractional part is exactly 0.5 (which rounds to even)
			if remainder_h == 0x8000000000000000 &&
				(Q_low.w[1] < bid_reciprocals10_128[extra_digits].w[1] ||
					(Q_low.w[1] == bid_reciprocals10_128[extra_digits].w[1] &&
						Q_low.w[0] < bid_reciprocals10_128[extra_digits].w[0])) {
				status = 0 // BID_EXACT_STATUS
			}
		case BID_ROUNDING_DOWN, BID_ROUNDING_TO_ZERO:
			// test whether fractional part is 0
			if remainder_h == 0 &&
				(Q_low.w[1] < bid_reciprocals10_128[extra_digits].w[1] ||
					(Q_low.w[1] == bid_reciprocals10_128[extra_digits].w[1] &&
						Q_low.w[0] < bid_reciprocals10_128[extra_digits].w[0])) {
				status = 0 // BID_EXACT_STATUS
			}
		default:
			// BID_ROUNDING_UP: check if adding reciprocal causes overflow
			var CY, carry uint64
			Stemp_w0, CY := bits.Add64(Q_low.w[0], bid_reciprocals10_128[extra_digits].w[0], 0)
			_, carry = bits.Add64(Q_low.w[1], bid_reciprocals10_128[extra_digits].w[1], CY)
			_ = Stemp_w0
			if (remainder_h>>(64-uint(amount)))+carry >= (uint64(1) << uint(amount)) {
				status = 0 // BID_EXACT_STATUS
			}
		}
		pfpsf |= status

		res, flags := fast_get_BID64_check_OF_flags(sign_x^sign_y, final_exponent, C64, rndMode)
		pfpsf |= flags
		return res, pfpsf
	}

	// go to convert_format and exit
	C64 = P.w[0]
	res = get_BID64(sign_x^sign_y,
		exponent_x+exponent_y-DECIMAL_EXPONENT_BIAS, C64, rndMode)
	return res, pfpsf
}
