// Ported from: IntelRDFPMathLib20U4/LIBRARY/src/bid32_to_bid64.c
// (the bid32 packing helpers follow get_BID32 and very_fast_get_BID32 in
// bid_internal.h)
// Version: Intel(R) Decimal Floating-Point Math Library 2.0 Update 4
//
// This file is a mechanical translation of the Intel BID library to Go.
// All logic, magic numbers, and table references are preserved exactly.

package bidgo

import (
	"math"
)

func bid32GetNoFlags(sgn uint32, expon int, coeff uint64, rmode int) uint32 {
	if coeff > 9999999 {
		expon++
		coeff = 1000000
	}
	if uint(expon) > 191 {
		if expon < 0 {
			if expon+7 < 0 {
				if rmode == 1 && sgn != 0 {
					return 0x80000001
				}
				if rmode == 2 && sgn == 0 {
					return 0x00000001
				}
				return sgn
			}
			if sgn != 0 && uint(rmode-1) < 2 {
				rmode = 3 - rmode
			}
			extraDigits := -expon
			coeff += bid_round_const_table[rmode][extraDigits]
			Q := __mul_64x64_to_128(coeff, bid_reciprocals10_64[extraDigits])
			amount := bid_short_recip_scale[extraDigits]
			coeff = Q.w[1] >> amount
			if rmode == 0 && (coeff&1) != 0 {
				remainder_h := Q.w[1] & ((uint64(1) << amount) - 1)
				if remainder_h == 0 && Q.w[0] < bid_reciprocals10_64[extraDigits] {
					coeff--
				}
			}
			return sgn | uint32(coeff)
		}
		if coeff == 0 && expon > 191 {
			expon = 191
		}
		for coeff < 1000000 && expon >= 192 {
			expon--
			coeff = (coeff << 3) + (coeff << 1)
		}
		if expon > 191 {
			r := sgn | 0x78000000
			switch rmode {
			case 1:
				if sgn == 0 {
					r = 0x77f8967f
				}
			case 3:
				r = sgn | 0x77f8967f
			case 2:
				if sgn != 0 {
					r = 0xf7f8967f
				}
			}
			return r
		}
	}
	return very_fast_get_BID32(sgn, expon, uint32(coeff))
}

func bid32GetWithFlags(sgn uint32, expon int, coeff uint64, rmode int, flags uint32) (uint32, uint32) {
	var Q BID_UINT128
	var C64, remainder_h, carry, Stemp uint64
	var r, mask, status uint32
	var extra_digits, amount, amount2 int

	if coeff > 9999999 {
		expon++
		coeff = 1000000
	}
	// check for possible underflow/overflow
	if uint(expon) > DECIMAL_MAX_EXPON_32 {
		if expon < 0 {
			// underflow
			if expon+MAX_FORMAT_DIGITS_32 < 0 {
				flags |= BID_UNDERFLOW_EXCEPTION | BID_INEXACT_EXCEPTION
				if rmode == BID_ROUNDING_DOWN && sgn != 0 {
					return 0x80000001, flags
				}
				if rmode == BID_ROUNDING_UP && sgn == 0 {
					return 1, flags
				}
				// result is 0
				return sgn, flags
			}
			// get digits to be shifted out
			if sgn != 0 && uint(rmode-1) < 2 {
				rmode = 3 - rmode
			}

			extra_digits = -expon
			coeff += bid_round_const_table[rmode][extra_digits]

			// get coeff*(2^M[extra_digits])/10^extra_digits
			Q = __mul_64x64_to_128(coeff, bid_reciprocals10_64[extra_digits])

			// now get P/10^extra_digits: shift Q_high right by M[extra_digits]-128
			amount = bid_short_recip_scale[extra_digits]

			C64 = Q.w[1] >> uint(amount)

			if rmode == BID_ROUNDING_TO_NEAREST && (C64&1) != 0 {
				// check whether fractional part of initial_P/10^extra_digits is exactly .5
				amount2 = 64 - amount
				remainder_h = ^uint64(0)
				remainder_h >>= uint(amount2)
				remainder_h &= Q.w[1]

				if remainder_h == 0 && Q.w[0] < bid_reciprocals10_64[extra_digits] {
					C64--
				}
			}

			if (flags & BID_INEXACT_EXCEPTION) != 0 {
				flags |= BID_UNDERFLOW_EXCEPTION
			} else {
				status = BID_INEXACT_EXCEPTION
				// get remainder
				remainder_h = Q.w[1] << uint(64-amount)

				switch rmode {
				case BID_ROUNDING_TO_NEAREST, BID_ROUNDING_TIES_AWAY:
					// test whether fractional part is 0
					if remainder_h == 0x8000000000000000 &&
						(Q.w[0] < bid_reciprocals10_64[extra_digits]) {
						status = BID_EXACT_STATUS
					}
				case BID_ROUNDING_DOWN, BID_ROUNDING_TO_ZERO:
					if remainder_h == 0 && (Q.w[0] < bid_reciprocals10_64[extra_digits]) {
						status = BID_EXACT_STATUS
					}
				default:
					// round up
					Stemp, carry = __add_carry_out(Q.w[0], bid_reciprocals10_64[extra_digits])
					_ = Stemp
					if (remainder_h>>uint(64-amount))+carry >=
						(uint64(1) << uint(amount)) {
						status = BID_EXACT_STATUS
					}
				}

				if status != BID_EXACT_STATUS {
					flags |= BID_UNDERFLOW_EXCEPTION | status
				}
			}

			return sgn | uint32(C64), flags
		}

		if coeff == 0 && expon > DECIMAL_MAX_EXPON_32 {
			expon = DECIMAL_MAX_EXPON_32
		}
		for coeff < 1000000 && expon > DECIMAL_MAX_EXPON_32 {
			coeff = (coeff << 3) + (coeff << 1)
			expon--
		}
		if uint(expon) > DECIMAL_MAX_EXPON_32 {
			flags |= BID_OVERFLOW_EXCEPTION | BID_INEXACT_EXCEPTION
			// overflow
			r = sgn | INFINITY_MASK32
			switch rmode {
			case BID_ROUNDING_DOWN:
				if sgn == 0 {
					r = LARGEST_BID32
				}
			case BID_ROUNDING_TO_ZERO:
				r = sgn | LARGEST_BID32
			case BID_ROUNDING_UP:
				if sgn != 0 {
					r = sgn | LARGEST_BID32
				}
			}
			return r, flags
		}
	}

	mask = 1 << 23

	// check whether coefficient fits in DECIMAL_COEFF_FIT bits
	if coeff < uint64(mask) {
		r = uint32(expon)
		r <<= 23
		r |= (uint32(coeff) | sgn)
		return r, flags
	}
	// special format
	r = uint32(expon)
	r <<= 21
	r |= (sgn | SPECIAL_ENCODING_MASK32)
	// add coeff, without leading bits
	mask = (1 << 21) - 1
	r |= (uint32(coeff) & mask)

	return r, flags
}

// Bid64ToBid32 converts a BID64 to BID32 and returns status flags.
// Ported mechanically from Intel bid32_to_bid64.c: bid64_to_bid32.
func Bid64ToBid32(x uint64, rndMode int) (uint32, uint32) {
	var Q BID_UINT128
	var sign_x, coefficient_x, remainder_h, carry, Stemp uint64
	var res uint32
	var t64 uint64
	var exponent_x, bin_expon_cx, extra_digits, rmode, amount int
	var status uint32

	sign_x, exponent_x, coefficient_x, valid := unpack_BID64(x)
	// unpack arguments, check for NaN or Infinity, 0
	if !valid {
		if (x & 0x7800000000000000) == 0x7800000000000000 {
			t64 = coefficient_x & 0x0003ffffffffffff
			res = uint32(t64 / 1000000000)
			res |= uint32((coefficient_x >> 32) & 0xfc000000)
			if (x & SNAN_MASK64) == SNAN_MASK64 { // sNaN
				status |= BID_INVALID_EXCEPTION
			}
			return res, status
		}
		exponent_x =
			exponent_x - DECIMAL_EXPONENT_BIAS + DECIMAL_EXPONENT_BIAS_32
		if exponent_x < 0 {
			exponent_x = 0
		}
		if exponent_x > DECIMAL_MAX_EXPON_32 {
			exponent_x = DECIMAL_MAX_EXPON_32
		}
		res = uint32(sign_x>>32) | uint32(exponent_x<<23)
		return res, status
	}

	exponent_x =
		exponent_x - DECIMAL_EXPONENT_BIAS + DECIMAL_EXPONENT_BIAS_32

	// check number of digits
	if coefficient_x >= 10000000 {
		tempx := math.Float32bits(float32(coefficient_x))
		bin_expon_cx = int((tempx>>23)&0xff) - 0x7f
		extra_digits = bid_estimate_decimal_digits[bin_expon_cx] - 7
		// add test for range
		if coefficient_x >= bid_power10_index_binexp[bin_expon_cx] {
			extra_digits++
		}

		rmode = rndMode
		if sign_x != 0 && uint(rmode-1) < 2 {
			rmode = 3 - rmode
		}

		exponent_x += extra_digits
		if (exponent_x < 0) && (exponent_x+MAX_FORMAT_DIGITS_32 >= 0) {
			status = BID_UNDERFLOW_EXCEPTION
			extra_digits -= exponent_x
			exponent_x = 0
		}
		coefficient_x += bid_round_const_table[rmode][extra_digits]
		Q = __mul_64x64_to_128(coefficient_x, bid_reciprocals10_64[extra_digits])

		// now get P/10^extra_digits: shift Q_high right by M[extra_digits]-128
		amount = bid_short_recip_scale[extra_digits]

		coefficient_x = Q.w[1] >> amount

		if rmode == 0 { // BID_ROUNDING_TO_NEAREST
			if (coefficient_x & 1) != 0 {
				// check whether fractional part of initial_P/10^extra_digits
				// is exactly .5

				// get remainder
				remainder_h = Q.w[1] << (64 - amount)

				if remainder_h == 0 && (Q.w[0] < bid_reciprocals10_64[extra_digits]) {
					coefficient_x--
				}
			}
		}

		status |= BID_INEXACT_EXCEPTION
		// get remainder
		remainder_h = Q.w[1] << (64 - amount)

		switch rmode {
		case BID_ROUNDING_TO_NEAREST, BID_ROUNDING_TIES_AWAY:
			// test whether fractional part is 0
			if remainder_h == 0x8000000000000000 &&
				(Q.w[0] < bid_reciprocals10_64[extra_digits]) {
				status = BID_EXACT_STATUS
			}
		case BID_ROUNDING_DOWN, BID_ROUNDING_TO_ZERO:
			if remainder_h == 0 && (Q.w[0] < bid_reciprocals10_64[extra_digits]) {
				status = BID_EXACT_STATUS
			}
		default:
			// round up
			Stemp, carry = __add_carry_out(Q.w[0], bid_reciprocals10_64[extra_digits])
			_ = Stemp
			if (remainder_h>>uint(64-amount))+carry >=
				(uint64(1) << uint(amount)) {
				status = BID_EXACT_STATUS
			}
		}
	}

	res, status = bid32GetWithFlags(uint32(sign_x>>32), exponent_x, coefficient_x, rndMode, status)
	return res, status
}
