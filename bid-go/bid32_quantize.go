// Ported from: Intel bid32_quantize.c
// Mechanical translation - all logic preserved exactly.

package bidgo

import "math"

// Bid32Quantize is ported mechanically from bid32_quantize.c: bid32_quantize.
func Bid32Quantize(x, y uint32, rnd_mode int) (uint32, uint32) {
	var CT uint64
	var sign_x, coefficient_x, coefficient_y, remainder_h, C64, CT0 uint32
	var carry, res uint32
	var exponent_x, exponent_y, digits_x, extra_digits, amount, amount2 int
	var expon_diff, total_digits, bin_expon_cx int
	var rmode int
	var status, pfpsf uint32
	var valid_x bool

	sign_x, exponent_x, coefficient_x, valid_x = unpack_BID32(x)
	_, exponent_y, coefficient_y, valid_y := unpack_BID32(y)
	if coefficient_x == 0 {
		valid_x = false
	}
	if coefficient_y == 0 {
		valid_y = false
	}

	if !valid_y {
		if (x & SNAN_MASK32) == SNAN_MASK32 {
			pfpsf |= BID_INVALID_EXCEPTION
		}
		if ((coefficient_x << 1) == 0xf0000000) && ((coefficient_y << 1) == 0xf0000000) {
			res = coefficient_x
			return res, pfpsf
		}
		if (y & 0x78000000) == 0x78000000 {
			if ((y & SNAN_MASK32) == SNAN_MASK32) ||
				(((y & NAN_MASK32) == INFINITY_MASK32) && ((x & NAN_MASK32) < INFINITY_MASK32)) {
				pfpsf |= BID_INVALID_EXCEPTION
			}
			if (y & NAN_MASK32) != NAN_MASK32 {
				coefficient_y = 0
			}
			if (x & NAN_MASK32) != NAN_MASK32 {
				res = 0x7c000000 | (coefficient_y & QUIET_MASK32)
				if ((y & NAN_MASK32) != NAN_MASK32) && ((x & NAN_MASK32) == INFINITY_MASK32) {
					res = x
				}
				return res, pfpsf
			}
		}
	}
	if !valid_x {
		if (x & INFINITY_MASK32) == INFINITY_MASK32 {
			if ((x & SNAN_MASK32) == SNAN_MASK32) || ((x & NAN_MASK32) == INFINITY_MASK32) {
				pfpsf |= BID_INVALID_EXCEPTION
			}
			if (x & NAN_MASK32) != NAN_MASK32 {
				coefficient_x = 0
			}
			res = NAN_MASK32 | (coefficient_x & QUIET_MASK32)
			return res, pfpsf
		}
		res = very_fast_get_BID32(sign_x, exponent_y, 0)
		return res, pfpsf
	}

	tempx := math.Float32bits(float32(coefficient_x))
	bin_expon_cx = int((tempx>>23)&0xff) - 0x7f
	digits_x = bid_estimate_decimal_digits[bin_expon_cx]
	if uint64(coefficient_x) >= bid_power10_table_128[digits_x].w[0] {
		digits_x++
	}

	expon_diff = exponent_x - exponent_y
	total_digits = digits_x + expon_diff

	if uint32(total_digits+1) <= 8 {
		if expon_diff >= 0 {
			coefficient_x *= uint32(bid_power10_table_128[expon_diff].w[0])
			res = very_fast_get_BID32(sign_x, exponent_y, coefficient_x)
			return res, pfpsf
		}
		extra_digits = -expon_diff
		rmode = rnd_mode
		if sign_x != 0 && uint(rmode-1) < 2 {
			rmode = 3 - rmode
		}
		coefficient_x += uint32(bid_round_const_table[rmode][extra_digits])

		CT = uint64(coefficient_x) * bid_bid_reciprocals10_32[extra_digits]

		amount = bid_bid_bid_recip_scale32[extra_digits]
		CT0 = uint32(CT >> 32)
		C64 = CT0 >> uint(amount)

		if rnd_mode == 0 {
			if C64&1 != 0 {
				amount2 = 32 - amount
				remainder_h = 0
				remainder_h--
				remainder_h >>= uint(amount2)
				remainder_h = remainder_h & CT0

				if remainder_h == 0 && (uint32(CT) < uint32(bid_bid_reciprocals10_32[extra_digits])) {
					C64--
				}
			}
		}

		status = BID_INEXACT_EXCEPTION
		remainder_h = CT0 << uint(32-amount)
		switch rmode {
		case BID_ROUNDING_TO_NEAREST, BID_ROUNDING_TIES_AWAY:
			if (remainder_h == 0x80000000) && (uint32(CT) < uint32(bid_bid_reciprocals10_32[extra_digits])) {
				status = BID_EXACT_STATUS
			}
		case BID_ROUNDING_DOWN, BID_ROUNDING_TO_ZERO:
			if remainder_h == 0 && (uint32(CT) < uint32(bid_bid_reciprocals10_32[extra_digits])) {
				status = BID_EXACT_STATUS
			}
		default:
			if uint32(CT)+uint32(bid_bid_reciprocals10_32[extra_digits]) < uint32(CT) {
				carry = 1
			} else {
				carry = 0
			}
			if (remainder_h>>uint(32-amount))+carry >= (uint32(1) << uint(amount)) {
				status = BID_EXACT_STATUS
			}
		}
		pfpsf |= status

		res = very_fast_get_BID32(sign_x, exponent_y, C64)
		return res, pfpsf
	}

	if total_digits < 0 {
		pfpsf |= BID_INEXACT_EXCEPTION
		C64 = 0
		rmode = rnd_mode
		if sign_x != 0 && uint(rmode-1) < 2 {
			rmode = 3 - rmode
		}
		if rmode == BID_ROUNDING_UP {
			C64 = 1
		}
		res = very_fast_get_BID32(sign_x, exponent_y, C64)
		return res, pfpsf
	}

	pfpsf |= BID_INVALID_EXCEPTION
	res = 0x7c000000
	return res, pfpsf
}
