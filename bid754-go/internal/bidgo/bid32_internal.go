// Ported from: IntelRDFPMathLib20U4/LIBRARY/src/bid_internal.h
// Version: Intel(R) Decimal Floating-Point Math Library 2.0 Update 4
//
// This file is a mechanical translation of the Intel BID library to Go.
// All logic, magic numbers, and table references are preserved exactly.

package bidgo

import "math/bits"

// BID32 constants from bid_internal.h
const (
	DECIMAL_MAX_EXPON_32     = 191
	DECIMAL_EXPONENT_BIAS_32 = 101
	MAX_FORMAT_DIGITS_32     = 7
	SPECIAL_ENCODING_MASK32  = 0x60000000
	SINFINITY_MASK32         = 0xf8000000
	INFINITY_MASK32          = 0x78000000
	LARGE_COEFF_MASK32       = 0x007fffff
	LARGE_COEFF_HIGH_BIT32   = 0x00800000
	SMALL_COEFF_MASK32       = 0x001fffff
	EXPONENT_MASK32          = 0xff
	LARGEST_BID32            = 0x77f8967f
	NAN_MASK32               = 0x7c000000
	SNAN_MASK32              = 0x7e000000
	SSNAN_MASK32             = 0xfc000000
	QUIET_MASK32             = 0xfdffffff

	// Aliases for Intel BID naming convention
	MASK_INF32               = 0x78000000
	MASK_SIGN32              = 0x80000000
	MASK_NAN32               = 0x7c000000
	MASK_SNAN32              = 0x7e000000
	MASK_STEERING_BITS32     = 0x60000000
	MASK_BINARY_EXPONENT1_32 = 0x7f800000
	MASK_BINARY_SIG1_32      = 0x007fffff
	MASK_BINARY_EXPONENT2_32 = 0x1fe00000
	MASK_BINARY_SIG2_32      = 0x001fffff
	MASK_BINARY_OR2_32       = 0x00800000
)

// very_fast_get_BID32 packs sign, exponent, and coefficient into BID32
// No overflow/underflow checking, no rounding
// Ported from Intel BID library bid_internal.h
func very_fast_get_BID32(sgn uint32, expon int, coeff uint32) uint32 {
	var r, mask uint32

	mask = 1 << 23

	// check whether coefficient fits in 10*2+3 bits
	if coeff < mask {
		r = uint32(expon)
		r <<= 23
		r |= (coeff | sgn)
		return r
	}
	// special format
	r = uint32(expon)
	r <<= 21
	r |= (sgn | SPECIAL_ENCODING_MASK32)
	// add coeff, without leading bits
	mask = (1 << 21) - 1
	coeff &= mask
	r |= coeff

	return r
}

// fast_get_BID32 packs sign, exponent, and coefficient into BID32
// With coefficient overflow handling
// Ported from Intel BID library bid_internal.h
func fast_get_BID32(sgn uint32, expon int, coeff uint32) uint32 {
	var r, mask uint32

	mask = 1 << 23

	if coeff > 9999999 {
		expon++
		coeff = 1000000
	}
	// check whether coefficient fits in 10*2+3 bits
	if coeff < mask {
		r = uint32(expon)
		r <<= 23
		r |= (coeff | sgn)
		return r
	}
	// special format
	r = uint32(expon)
	r <<= 21
	r |= (sgn | SPECIAL_ENCODING_MASK32)
	// add coeff, without leading bits
	mask = (1 << 21) - 1
	coeff &= mask
	r |= coeff

	return r
}

// get_BID32 packs with full overflow/underflow checking and rounding
// Ported from Intel BID library bid_internal.h
func get_BID32(sgn uint32, expon int, coeff uint64, rmode int) uint32 {
	var Q BID_UINT128
	var _C64, remainder_h uint64
	var r, mask uint32
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
				// note: get_BID32 does not have pfpsf parameter; flags not set
				if rmode == BID_ROUNDING_DOWN && sgn != 0 {
					return 0x80000001
				}
				if rmode == BID_ROUNDING_UP && sgn == 0 {
					return 1
				}
				// result is 0
				return sgn
			}
			// get digits to be shifted out
			if sgn != 0 && uint(rmode-1) < 2 {
				rmode = 3 - rmode
			}

			extra_digits = -expon
			coeff += bid_round_const_table[rmode][extra_digits]

			// get coeff*(2^M[extra_digits])/10^extra_digits
			Q = __mul_64x64_to_128(coeff, bid_reciprocals10_64[extra_digits])

			// now get P/10^extra_digits: shift Q_high right by M[extra_digits]-64
			amount = bid_short_recip_scale[extra_digits]

			_C64 = Q.w[1] >> uint(amount)

			if rmode == 0 { // BID_ROUNDING_TO_NEAREST
				if _C64&1 != 0 {
					// check whether fractional part of initial_P/10^extra_digits is exactly .5

					// get remainder
					amount2 = 64 - amount
					remainder_h = (^uint64(0)) >> uint(amount2)
					remainder_h = remainder_h & Q.w[1]

					if remainder_h == 0 && Q.w[0] < bid_reciprocals10_64[extra_digits] {
						_C64--
					}
				}
			}

			return sgn | uint32(_C64)
		}

		if coeff == 0 {
			if expon > DECIMAL_MAX_EXPON_32 {
				expon = DECIMAL_MAX_EXPON_32
			}
		}

		for coeff < 1000000 && expon >= 3*64 {
			expon--
			coeff = (coeff << 3) + (coeff << 1)
		}

		if expon > DECIMAL_MAX_EXPON_32 {
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
					r = 0x80000000 | LARGEST_BID32
				}
			}
			return r
		}
	}

	mask = 1 << 23

	// check whether coefficient fits in 10*2+3 bits
	if uint32(coeff) < mask {
		r = uint32(expon)
		r <<= 23
		r |= (uint32(coeff) | sgn)
		return r
	}
	// special format
	r = uint32(expon)
	r <<= 21
	r |= (sgn | SPECIAL_ENCODING_MASK32)
	// add coeff, without leading bits
	mask = (1 << 21) - 1
	coeff64 := uint32(coeff)
	coeff64 &= mask
	r |= coeff64

	return r
}

// get_BID32_flags packs with full overflow/underflow checking, rounding, and flags.
// Ported mechanically from Intel bid_internal.h:get_BID32.
func get_BID32_flags(sgn uint32, expon int, coeff uint64, rmode int, pfpsf *uint32) uint32 {
	var Q BID_UINT128
	var _C64, remainder_h, carry, Stemp uint64
	var r, mask uint32
	var extra_digits, amount, amount2 int
	var status uint32

	if coeff > 9999999 {
		expon++
		coeff = 1000000
	}
	if uint(expon) > DECIMAL_MAX_EXPON_32 {
		if expon < 0 {
			if expon+MAX_FORMAT_DIGITS_32 < 0 {
				*pfpsf |= BID_UNDERFLOW_EXCEPTION | BID_INEXACT_EXCEPTION
				if rmode == BID_ROUNDING_DOWN && sgn != 0 {
					return 0x80000001
				}
				if rmode == BID_ROUNDING_UP && sgn == 0 {
					return 1
				}
				return sgn
			}
			if sgn != 0 && uint(rmode-1) < 2 {
				rmode = 3 - rmode
			}

			extra_digits = -expon
			coeff += bid_round_const_table[rmode][extra_digits]
			Q = __mul_64x64_to_128(coeff, bid_reciprocals10_64[extra_digits])
			amount = bid_short_recip_scale[extra_digits]
			_C64 = Q.w[1] >> uint(amount)

			if rmode == BID_ROUNDING_TO_NEAREST && (_C64&1) != 0 {
				amount2 = 64 - amount
				remainder_h = ^uint64(0)
				remainder_h >>= uint(amount2)
				remainder_h &= Q.w[1]

				if remainder_h == 0 && Q.w[0] < bid_reciprocals10_64[extra_digits] {
					_C64--
				}
			}

			if (*pfpsf & BID_INEXACT_EXCEPTION) != 0 {
				*pfpsf |= BID_UNDERFLOW_EXCEPTION
			} else {
				status = BID_INEXACT_EXCEPTION
				remainder_h = Q.w[1] << uint(64-amount)

				switch rmode {
				case BID_ROUNDING_TO_NEAREST, BID_ROUNDING_TIES_AWAY:
					if remainder_h == 0x8000000000000000 &&
						Q.w[0] < bid_reciprocals10_64[extra_digits] {
						status = 0
					}
				case BID_ROUNDING_DOWN, BID_ROUNDING_TO_ZERO:
					if remainder_h == 0 &&
						Q.w[0] < bid_reciprocals10_64[extra_digits] {
						status = 0
					}
				default:
					Stemp, carry = bits.Add64(Q.w[0], bid_reciprocals10_64[extra_digits], 0)
					_ = Stemp
					if (remainder_h>>uint(64-amount))+carry >= (uint64(1) << uint(amount)) {
						status = 0
					}
				}
				if status != 0 {
					*pfpsf |= BID_UNDERFLOW_EXCEPTION | status
				}
			}

			return sgn | uint32(_C64)
		}

		if coeff == 0 {
			if expon > DECIMAL_MAX_EXPON_32 {
				expon = DECIMAL_MAX_EXPON_32
			}
		}
		for coeff < 1000000 && expon > DECIMAL_MAX_EXPON_32 {
			coeff = (coeff << 3) + (coeff << 1)
			expon--
		}
		if uint(expon) > DECIMAL_MAX_EXPON_32 {
			*pfpsf |= BID_OVERFLOW_EXCEPTION | BID_INEXACT_EXCEPTION
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
			return r
		}
	}

	mask = 1 << 23
	if uint32(coeff) < mask {
		r = uint32(expon)
		r <<= 23
		r |= (uint32(coeff) | sgn)
		return r
	}

	r = uint32(expon)
	r <<= 21
	r |= (sgn | SPECIAL_ENCODING_MASK32)
	mask = (1 << 21) - 1
	r |= (uint32(coeff) & mask)
	return r
}

// get_BID32_UF is called when underflow is known to occur
// Ported from Intel BID library bid_internal.h
func get_BID32_UF(sgn uint32, expon int, coeff uint64, R uint32, rmode int, pfpsf *uint32) uint32 {
	var Q BID_UINT128
	var _C64, remainder_h uint64
	var r, mask uint32
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
				*pfpsf |= BID_UNDERFLOW_EXCEPTION | BID_INEXACT_EXCEPTION
				if rmode == BID_ROUNDING_DOWN && sgn != 0 {
					return 0x80000001
				}
				if rmode == BID_ROUNDING_UP && sgn == 0 {
					return 1
				}
				// result is 0
				return sgn
			}
			// get digits to be shifted out
			if sgn != 0 && uint(rmode-1) < 2 {
				rmode = 3 - rmode
			}

			// 10*coeff
			coeff = (coeff << 3) + (coeff << 1)
			if R != 0 {
				coeff |= 1
			}

			extra_digits = 1 - expon
			coeff += bid_round_const_table[rmode][extra_digits]

			// get coeff*(2^M[extra_digits])/10^extra_digits
			Q = __mul_64x64_to_128(coeff, bid_reciprocals10_64[extra_digits])

			// now get P/10^extra_digits: shift Q_high right by M[extra_digits]-64
			amount = bid_short_recip_scale[extra_digits]

			_C64 = Q.w[1] >> uint(amount)

			if rmode == 0 { // BID_ROUNDING_TO_NEAREST
				if _C64&1 != 0 {
					// check whether fractional part of initial_P/10^extra_digits is exactly .5

					// get remainder
					amount2 = 64 - amount
					remainder_h = (^uint64(0)) >> uint(amount2)
					remainder_h = remainder_h & Q.w[1]

					if remainder_h == 0 && Q.w[0] < bid_reciprocals10_64[extra_digits] {
						_C64--
					}
				}
			}

			// set underflow/inexact flags
			if (*pfpsf & BID_INEXACT_EXCEPTION) != 0 {
				*pfpsf |= BID_UNDERFLOW_EXCEPTION
			} else {
				status := uint32(BID_INEXACT_EXCEPTION)
				remainder_h = Q.w[1] << uint(64-amount)
				switch rmode {
				case BID_ROUNDING_TO_NEAREST, BID_ROUNDING_TIES_AWAY:
					if remainder_h == 0x8000000000000000 &&
						Q.w[0] < bid_reciprocals10_64[extra_digits] {
						status = BID_EXACT_STATUS
					}
				case BID_ROUNDING_DOWN, BID_ROUNDING_TO_ZERO:
					if remainder_h == 0 && Q.w[0] < bid_reciprocals10_64[extra_digits] {
						status = BID_EXACT_STATUS
					}
				default:
					var carry uint64
					_, carry = __add_carry_out(Q.w[0], bid_reciprocals10_64[extra_digits])
					if (remainder_h>>uint(64-amount))+carry >= (uint64(1) << uint(amount)) {
						status = BID_EXACT_STATUS
					}
				}
				if status != BID_EXACT_STATUS {
					*pfpsf |= BID_UNDERFLOW_EXCEPTION | status
				}
			}

			return sgn | uint32(_C64)
		}

		if coeff == 0 {
			if expon > DECIMAL_MAX_EXPON_32 {
				expon = DECIMAL_MAX_EXPON_32
			}
		}

		for coeff < 1000000 && expon >= 3*64 {
			expon--
			coeff = (coeff << 3) + (coeff << 1)
		}

		if expon > DECIMAL_MAX_EXPON_32 {
			// overflow
			*pfpsf |= BID_OVERFLOW_EXCEPTION | BID_INEXACT_EXCEPTION
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
					r = 0x80000000 | LARGEST_BID32
				}
			}
			return r
		}
	}

	mask = 1 << 23

	// check whether coefficient fits in 10*2+3 bits
	if uint32(coeff) < mask {
		r = uint32(expon)
		r <<= 23
		r |= (uint32(coeff) | sgn)
		return r
	}
	// special format
	r = uint32(expon)
	r <<= 21
	r |= (sgn | SPECIAL_ENCODING_MASK32)
	// add coeff, without leading bits
	mask = (1 << 21) - 1
	coeff64 := uint32(coeff)
	coeff64 &= mask
	r |= coeff64

	return r
}

// unpack_BID32_intel unpacks a BID32 value into sign, exponent, and coefficient
// Returns valid=false if the value is NaN or Infinity
// Ported from Intel BID library bid_internal.h (same as existing unpack_BID32 but returns uint32 types)
func unpack_BID32_intel(x uint32) (sign uint32, exponent int, coefficient uint32, valid bool) {
	sign = x & 0x80000000

	if (x & SPECIAL_ENCODING_MASK32) == SPECIAL_ENCODING_MASK32 {
		// special encodings
		coefficient = (x & SMALL_COEFF_MASK32) | LARGE_COEFF_HIGH_BIT32

		if (x & INFINITY_MASK32) == INFINITY_MASK32 {
			exponent = 0
			coefficient = x & 0xfe0fffff
			if (x & 0x000fffff) >= 1000000 {
				coefficient = x & 0xfe000000
			}
			if (x & NAN_MASK32) == INFINITY_MASK32 {
				coefficient = x & SINFINITY_MASK32
			}
			return sign, exponent, coefficient, false // NaN or Infinity
		}
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
