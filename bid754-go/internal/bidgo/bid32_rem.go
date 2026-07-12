// Ported from: Intel bid32_rem.c
// Mechanical translation - all logic preserved exactly.

package bidgo

import "math"

// Bid32Rem is ported mechanically from bid32_rem.c: bid32_rem.
func Bid32Rem(x, y uint32) (uint32, uint32) {
	var CX, Q64, CYL uint64
	var CY, sign_x, sign_y, coefficient_x, coefficient_y, res uint32
	var Q, R, R2, T uint32
	var exponent_x, exponent_y, bin_expon, e_scale int
	var digits_x, diff_expon int
	var pfpsf uint32

	sign_y, exponent_y, coefficient_y, valid_y := unpack_BID32(y)
	sign_x, exponent_x, coefficient_x, valid_x := unpack_BID32(x)
	_ = sign_y
	if coefficient_x == 0 {
		valid_x = false
	}
	if coefficient_y == 0 {
		valid_y = false
	}

	if !valid_x {
		if (y & SNAN_MASK32) == SNAN_MASK32 {
			pfpsf |= BID_INVALID_EXCEPTION
		}
		if (x & 0x7c000000) == 0x7c000000 {
			if (x & SNAN_MASK32) == SNAN_MASK32 {
				pfpsf |= BID_INVALID_EXCEPTION
			}
			res = coefficient_x & QUIET_MASK32
			return res, pfpsf
		}
		if (x & 0x78000000) == 0x78000000 {
			if (y & NAN_MASK32) != NAN_MASK32 {
				pfpsf |= BID_INVALID_EXCEPTION
				res = 0x7c000000
				return res, pfpsf
			}
		}
		if ((y & 0x78000000) < 0x78000000) && coefficient_y != 0 {
			if (y & 0x60000000) == 0x60000000 {
				exponent_y = int((y >> 21) & 0xff)
			} else {
				exponent_y = int((y >> 23) & 0xff)
			}
			if exponent_y < exponent_x {
				exponent_x = exponent_y
			}
			res = (uint32(exponent_x) << 23) | sign_x
			return res, pfpsf
		}
	}
	if !valid_y {
		if (y & 0x7c000000) == 0x7c000000 {
			if (y & SNAN_MASK32) == SNAN_MASK32 {
				pfpsf |= BID_INVALID_EXCEPTION
			}
			res = coefficient_y & QUIET_MASK32
			return res, pfpsf
		}
		if (y & 0x78000000) == 0x78000000 {
			res = very_fast_get_BID32(sign_x, exponent_x, coefficient_x)
			return res, pfpsf
		}
		pfpsf |= BID_INVALID_EXCEPTION
		res = 0x7c000000
		return res, pfpsf
	}

	diff_expon = exponent_x - exponent_y
	if diff_expon <= 0 {
		diff_expon = -diff_expon

		if diff_expon > 7 {
			res = x
			return res, pfpsf
		}
		T = uint32(bid_power10_table_128[diff_expon].w[0])
		CYL = uint64(coefficient_y) * uint64(T)
		if CYL > uint64(coefficient_x<<1) {
			res = x
			return res, pfpsf
		}

		CY = uint32(CYL)
		Q = coefficient_x / CY
		R = coefficient_x - Q*CY

		R2 = R + R
		if R2 > CY || (R2 == CY && (Q&1) != 0) {
			R = CY - R
			sign_x ^= 0x80000000
		}

		res = very_fast_get_BID32(sign_x, exponent_x, R)
		return res, pfpsf
	}

	CX = uint64(coefficient_x)
	for diff_expon > 0 {
		tempx := math.Float32bits(float32(CX))
		bin_expon = int((tempx>>23)&0xff) - 0x7f
		digits_x = bid_estimate_decimal_digits[bin_expon]

		e_scale = 18 - digits_x
		if diff_expon >= e_scale {
			diff_expon -= e_scale
		} else {
			e_scale = diff_expon
			diff_expon = 0
		}

		CX *= bid_power10_table_128[e_scale].w[0]

		Q64 = CX / uint64(coefficient_y)
		CX -= Q64 * uint64(coefficient_y)

		if CX == 0 {
			res = very_fast_get_BID32(sign_x, exponent_y, 0)
			return res, pfpsf
		}
	}

	coefficient_x = uint32(CX)
	R2 = coefficient_x + coefficient_x
	if R2 > coefficient_y || (R2 == coefficient_y && (Q64&1) != 0) {
		coefficient_x = coefficient_y - coefficient_x
		sign_x ^= 0x80000000
	}

	res = very_fast_get_BID32(sign_x, exponent_y, coefficient_x)
	return res, pfpsf
}
