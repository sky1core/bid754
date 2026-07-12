package bidgo

import "math"

// Bid64Quantize quantizes x to the exponent of y and returns status flags.
// Ported mechanically from Intel bid64_quantize.c.
func Bid64Quantize(x, y uint64, rndMode int) (uint64, uint32) {
	var CT BID_UINT128
	var sign_x, sign_y, coefficient_x, coefficient_y, remainder_h, C64 uint64
	var valid_x, valid_y bool
	var tmp, carry, res uint64
	var exponent_x, exponent_y, digits_x, extra_digits, amount, amount2 int
	var expon_diff, total_digits, bin_expon_cx int
	var rmode int
	var status uint32
	var pfpsf uint32

	sign_x, exponent_x, coefficient_x, valid_x = unpack_BID64(x)
	// unpack arguments, check for NaN or Infinity
	sign_y, exponent_y, coefficient_y, valid_y = unpack_BID64(y)
	if !valid_y {
		// Inf. or NaN or 0
		if (x & SNAN_MASK64) == SNAN_MASK64 {
			pfpsf |= BID_INVALID_EXCEPTION
		}

		// x=Inf, y=Inf?
		if (coefficient_x<<1) == 0xf000000000000000 &&
			(coefficient_y<<1) == 0xf000000000000000 {
			res = coefficient_x
			return res, pfpsf
		}
		// Inf or NaN?
		if (y & INFINITY_MASK64) == INFINITY_MASK64 {
			if ((y & SNAN_MASK64) == SNAN_MASK64) ||
				(((y & NAN_MASK64) == INFINITY_MASK64) &&
					((x & NAN_MASK64) < INFINITY_MASK64)) {
				pfpsf |= BID_INVALID_EXCEPTION
			}
			if (y & NAN_MASK64) != NAN_MASK64 {
				coefficient_y = 0
			}
			if (x & NAN_MASK64) != NAN_MASK64 {
				res = NAN_MASK64 | (coefficient_y & QUIET_MASK64)
				if ((y & NAN_MASK64) != NAN_MASK64) && ((x & NAN_MASK64) == INFINITY_MASK64) {
					res = x
				}
				return res, pfpsf
			}
		}
	}
	_ = sign_y
	// unpack arguments, check for NaN or Infinity
	if !valid_x {
		// x is Inf. or NaN or 0

		// Inf or NaN?
		if (x & INFINITY_MASK64) == INFINITY_MASK64 {
			if ((x & SNAN_MASK64) == SNAN_MASK64) || ((x & NAN_MASK64) == INFINITY_MASK64) {
				pfpsf |= BID_INVALID_EXCEPTION
			}
			if (x & NAN_MASK64) != NAN_MASK64 {
				coefficient_x = 0
			}
			res = NAN_MASK64 | (coefficient_x & QUIET_MASK64)
			return res, pfpsf
		}

		res = very_fast_get_BID64_small_mantissa(sign_x, exponent_y, 0)
		return res, pfpsf
	}
	// get number of decimal digits in coefficient_x
	tempx := math.Float32bits(float32(coefficient_x))
	bin_expon_cx = int((tempx>>23)&0xff) - 0x7f
	digits_x = bid_estimate_decimal_digits[bin_expon_cx]
	if coefficient_x >= bid_power10_table_128[digits_x].w[0] {
		digits_x++
	}

	expon_diff = exponent_x - exponent_y
	total_digits = digits_x + expon_diff

	// check range of scaled coefficient
	if uint32(total_digits+1) <= 17 {
		if expon_diff >= 0 {
			coefficient_x *= bid_power10_table_128[expon_diff].w[0]
			res = very_fast_get_BID64(sign_x, exponent_y, coefficient_x)
			return res, pfpsf
		}
		// must round off -expon_diff digits
		extra_digits = -expon_diff
		rmode = rndMode
		if sign_x != 0 && uint(rmode-1) < 2 {
			rmode = 3 - rmode
		}
		coefficient_x += bid_round_const_table[rmode][extra_digits]

		// get P*(2^M[extra_digits])/10^extra_digits
		CT = __mul_64x64_to_128(coefficient_x, bid_reciprocals10_64[extra_digits])

		// now get P/10^extra_digits: shift C64 right by M[extra_digits]-128
		amount = bid_short_recip_scale[extra_digits]
		C64 = CT.w[1] >> uint(amount)
		if rndMode == 0 {
			if C64&1 != 0 {
				// check whether fractional part of initial_P/10^extra_digits
				// is exactly .5
				// this is the same as fractional part of
				//   (initial_P + 0.5*10^extra_digits)/10^extra_digits is exactly zero

				// get remainder
				amount2 = 64 - amount
				remainder_h = 0
				remainder_h--
				remainder_h >>= uint(amount2)
				remainder_h = remainder_h & CT.w[1]

				// test whether fractional part is 0
				if remainder_h == 0 && (CT.w[0] < bid_reciprocals10_64[extra_digits]) {
					C64--
				}
			}
		}

		status = BID_INEXACT_EXCEPTION
		// get remainder
		remainder_h = CT.w[1] << (64 - uint(amount))
		switch rmode {
		case BID_ROUNDING_TO_NEAREST:
			fallthrough
		case BID_ROUNDING_TIES_AWAY:
			// test whether fractional part is 0
			if (remainder_h == 0x8000000000000000) &&
				(CT.w[0] < bid_reciprocals10_64[extra_digits]) {
				status = BID_EXACT_STATUS
			}
		case BID_ROUNDING_DOWN:
			fallthrough
		case BID_ROUNDING_TO_ZERO:
			if remainder_h == 0 && (CT.w[0] < bid_reciprocals10_64[extra_digits]) {
				status = BID_EXACT_STATUS
			}
		default:
			// round up
			tmp, carry = __add_carry_out(CT.w[0], bid_reciprocals10_64[extra_digits])
			_ = tmp
			if (remainder_h>>uint(64-amount))+carry >= (uint64(1) << uint(amount)) {
				status = BID_EXACT_STATUS
			}
		}
		pfpsf |= status

		res = very_fast_get_BID64_small_mantissa(sign_x, exponent_y, C64)
		return res, pfpsf
	}

	if total_digits < 0 {
		pfpsf |= BID_INEXACT_EXCEPTION
		C64 = 0
		rmode = rndMode
		if sign_x != 0 && uint(rmode-1) < 2 {
			rmode = 3 - rmode
		}
		if rmode == BID_ROUNDING_UP {
			C64 = 1
		}
		res = very_fast_get_BID64_small_mantissa(sign_x, exponent_y, C64)
		return res, pfpsf
	}
	// else more than 16 digits in coefficient
	pfpsf |= BID_INVALID_EXCEPTION
	res = NAN_MASK64
	return res, pfpsf
}
