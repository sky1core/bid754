// Ported from: Intel bid128_quantize.c
// Mechanical translation - all logic preserved exactly.

package bidgo

import "math"

// __mul_128x128_low is in bid128_div.go

// Bid128Quantize is ported mechanically from bid128_quantize.c: bid128_quantize.
func Bid128Quantize(x, y BID_UINT128, rnd_mode int) (BID_UINT128, uint32) {
	var CT BID_UINT256
	var CX, CY, T, CX2, CR, Stemp, res, REM_H, C2N BID_UINT128
	var sign_x, remainder_h, carry, CY64 uint64
	var valid_x bool
	var exponent_x, exponent_y, digits_x, extra_digits, amount int
	var expon_diff, total_digits, bin_expon_cx, rmode, status int
	var pfpsf uint32

	sign_x, exponent_x, CX, valid_x = unpack_BID128_value(x)

	// unpack arguments, check for NaN or Infinity
	_, exponent_y, CY, valid_y := unpack_BID128_value(y)
	if !valid_y {
		// y is Inf. or NaN
		if (x.w[1] & SNAN_MASK64) == SNAN_MASK64 { // y is sNaN
			pfpsf |= BID_INVALID_EXCEPTION
		}

		// test if y is NaN
		if (y.w[1] & 0x7c00000000000000) == 0x7c00000000000000 {
			if (y.w[1] & 0x7e00000000000000) == 0x7e00000000000000 {
				// set status flags
				pfpsf |= BID_INVALID_EXCEPTION
			}
			if (x.w[1] & 0x7c00000000000000) != 0x7c00000000000000 {
				res.w[1] = CY.w[1] & QUIET_MASK64
				res.w[0] = CY.w[0]
			} else {
				res.w[1] = CX.w[1] & QUIET_MASK64
				res.w[0] = CX.w[0]
			}
			return res, pfpsf
		}
		// y is Infinity?
		if (y.w[1] & 0x7800000000000000) == 0x7800000000000000 {
			// check if x is not Inf.
			if (x.w[1] & 0x7c00000000000000) < 0x7800000000000000 {
				// return NaN
				// set status flags
				pfpsf |= BID_INVALID_EXCEPTION
				res.w[1] = 0x7c00000000000000
				res.w[0] = 0
				return res, pfpsf
			} else if (x.w[1] & 0x7c00000000000000) <= 0x7800000000000000 {
				res.w[1] = CX.w[1] & QUIET_MASK64
				res.w[0] = CX.w[0]
				return res, pfpsf
			}
		}

	}

	if !valid_x {
		// test if x is NaN or Inf
		if (x.w[1] & 0x7c00000000000000) == 0x7800000000000000 {
			// set status flags
			pfpsf |= BID_INVALID_EXCEPTION
			res.w[1] = 0x7c00000000000000
			res.w[0] = 0
			return res, pfpsf
		} else if (x.w[1] & 0x7c00000000000000) == 0x7c00000000000000 {
			if (x.w[1] & 0x7e00000000000000) == 0x7e00000000000000 {
				// set status flags
				pfpsf |= BID_INVALID_EXCEPTION
			}
			res.w[1] = CX.w[1] & QUIET_MASK64
			res.w[0] = CX.w[0]
			return res, pfpsf
		}
		if CX.w[1] == 0 && CX.w[0] == 0 {
			res = very_fast_get_BID128(sign_x, exponent_y, CX)
			return res, pfpsf
		}
	}
	// get number of decimal digits in coefficient_x
	if CX.w[1] != 0 {
		tempx := math.Float32bits(float32(CX.w[1]))
		bin_expon_cx = int((tempx>>23)&0xff) - 0x7f + 64
	} else {
		tempx := math.Float32bits(float32(CX.w[0]))
		bin_expon_cx = int((tempx>>23)&0xff) - 0x7f
	}

	digits_x = int(bid_estimate_decimal_digits[bin_expon_cx])
	if CX.w[1] > bid_power10_table_128[digits_x].w[1] ||
		(CX.w[1] == bid_power10_table_128[digits_x].w[1] &&
			CX.w[0] >= bid_power10_table_128[digits_x].w[0]) {
		digits_x++
	}

	expon_diff = exponent_x - exponent_y
	total_digits = digits_x + expon_diff

	if uint32(total_digits) <= 34 {
		if expon_diff >= 0 {
			T = bid_power10_table_128[expon_diff]
			CX2 = __mul_128x128_low(T, CX)
			res = very_fast_get_BID128(sign_x, exponent_y, CX2)
			return res, pfpsf
		}
		rmode = rnd_mode
		if sign_x != 0 && uint(rmode-1) < 2 {
			rmode = 3 - rmode
		}
		// must round off -expon_diff digits
		extra_digits = -expon_diff
		CX = __add_128_128(CX, bid_round_const_table_128[rmode][extra_digits])

		// get P*(2^M[extra_digits])/10^extra_digits
		CT = __mul_128x128_to_256(CX, bid_reciprocals10_128[extra_digits])

		// now get P/10^extra_digits: shift C64 right by M[extra_digits]-128
		amount = int(bid_recip_scale[extra_digits])
		CX2.w[0] = CT.w[2]
		CX2.w[1] = CT.w[3]
		if amount >= 64 {
			CR.w[1] = 0
			CR.w[0] = CX2.w[1] >> uint(amount-64)
		} else {
			CR = __shr_128(CX2, uint(amount))
		}

		if rnd_mode == 0 {
			if CR.w[0]&1 != 0 {
				// check whether fractional part of initial_P/10^extra_digits is
				// exactly .5 this is the same as fractional part of
				// (initial_P + 0.5*10^extra_digits)/10^extra_digits is exactly zero

				// get remainder
				if amount >= 64 {
					remainder_h = CX2.w[0] | (CX2.w[1] << uint(128-amount))
				} else {
					remainder_h = CX2.w[0] << uint(64-amount)
				}

				// test whether fractional part is 0
				if remainder_h == 0 &&
					(CT.w[1] < bid_reciprocals10_128[extra_digits].w[1] ||
						(CT.w[1] == bid_reciprocals10_128[extra_digits].w[1] &&
							CT.w[0] < bid_reciprocals10_128[extra_digits].w[0])) {
					CR.w[0]--
				}
			}
		}

		status = int(BID_INEXACT_EXCEPTION)

		// get remainder
		if amount >= 64 {
			REM_H.w[1] = CX2.w[1] << uint(128-amount)
			REM_H.w[0] = CX2.w[0]
		} else {
			REM_H.w[1] = CX2.w[0] << uint(64-amount)
			REM_H.w[0] = 0
		}

		switch rmode {
		case BID_ROUNDING_TO_NEAREST, BID_ROUNDING_TIES_AWAY:
			// test whether fractional part is 0
			if REM_H.w[1] == 0x8000000000000000 && REM_H.w[0] == 0 &&
				(CT.w[1] < bid_reciprocals10_128[extra_digits].w[1] ||
					(CT.w[1] == bid_reciprocals10_128[extra_digits].w[1] &&
						CT.w[0] < bid_reciprocals10_128[extra_digits].w[0])) {
				status = int(BID_EXACT_STATUS)
			}
		case BID_ROUNDING_DOWN, BID_ROUNDING_TO_ZERO:
			if (REM_H.w[1]|REM_H.w[0]) == 0 &&
				(CT.w[1] < bid_reciprocals10_128[extra_digits].w[1] ||
					(CT.w[1] == bid_reciprocals10_128[extra_digits].w[1] &&
						CT.w[0] < bid_reciprocals10_128[extra_digits].w[0])) {
				status = int(BID_EXACT_STATUS)
			}
		default:
			// round up
			Stemp.w[0], CY64 = __add_carry_out(CT.w[0],
				bid_reciprocals10_128[extra_digits].w[0])
			Stemp.w[1], carry = __add_carry_in_out(CT.w[1],
				bid_reciprocals10_128[extra_digits].w[1], CY64)
			if amount < 64 {
				C2N.w[1] = 0
				C2N.w[0] = uint64(1) << uint(amount)
				REM_H.w[0] = REM_H.w[1] >> uint(64-amount)
				REM_H.w[1] = 0
			} else {
				C2N.w[1] = uint64(1) << uint(amount-64)
				C2N.w[0] = 0
				REM_H.w[1] >>= uint(128 - amount)
			}
			REM_H.w[0] += carry
			if REM_H.w[0] < carry {
				REM_H.w[1]++
			}
			if __unsigned_compare_ge_128(REM_H, C2N) {
				status = int(BID_EXACT_STATUS)
			}
		}

		pfpsf |= uint32(status)

		res = very_fast_get_BID128(sign_x, exponent_y, CR)
		return res, pfpsf
	}
	if total_digits < 0 {
		CR.w[1] = 0
		CR.w[0] = 0
		rmode = rnd_mode
		if sign_x != 0 && uint(rmode-1) < 2 {
			rmode = 3 - rmode
		}
		if rmode == BID_ROUNDING_UP {
			CR.w[0] = 1
		}
		pfpsf |= BID_INEXACT_EXCEPTION
		res = very_fast_get_BID128(sign_x, exponent_y, CR)
		return res, pfpsf
	}
	// else  more than 34 digits in coefficient
	pfpsf |= BID_INVALID_EXCEPTION
	res.w[1] = 0x7c00000000000000
	res.w[0] = 0
	return res, pfpsf
}
