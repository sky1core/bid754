// Ported from: IntelRDFPMathLib20U4/LIBRARY/src/bid64_add.c
// Version: Intel(R) Decimal Floating-Point Math Library 2.0 Update 4
//
// This file is a mechanical translation of the Intel BID library to Go.
// All logic, magic numbers, and table references are preserved exactly.

package bidgo

import "math"

// Bid64Sub subtracts y from x
// Ported from bid64_sub in bid64_add.c
func Bid64Sub(x, y uint64, rndMode int) uint64 {
	result, _ := Bid64SubWithFlags(x, y, rndMode)
	return result
}

// Bid64SubWithFlags subtracts y from x and returns flags
func Bid64SubWithFlags(x, y uint64, rndMode int) (uint64, uint32) {
	// check if y is not NaN
	if (y & NAN_MASK64) != NAN_MASK64 {
		y ^= 0x8000000000000000
	}
	return Bid64AddWithFlags(x, y, rndMode)
}

// Bid64Add adds x and y
// Ported from bid64_add in bid64_add.c
func Bid64Add(x, y uint64, rndMode int) uint64 {
	result, _ := Bid64AddWithFlags(x, y, rndMode)
	return result
}

// Bid64AddWithFlags adds x and y and returns (result, flags)
// Ported from bid64_add in bid64_add.c (line-by-line mechanical translation)
func Bid64AddWithFlags(x, y uint64, rndMode int) (uint64, uint32) {
	var CA, CT, CT_new BID_UINT128
	var sign_x, sign_y, coefficient_x, coefficient_y, C64_new uint64
	var valid_x, valid_y bool
	var res uint64
	var pfpsf uint32 // status flags
	var sign_a, sign_b, coefficient_a, coefficient_b, sign_s, sign_ab, rem_a uint64
	var saved_ca, saved_cb, C0_64, C64, remainder_h, T1, carry, tmp uint64
	var exponent_x, exponent_y, exponent_a, exponent_b, diff_dec_expon int
	var bin_expon_ca, extra_digits, amount, scale_k, scale_ca int
	var rmode int
	var status uint32

	sign_x, exponent_x, coefficient_x, valid_x = unpack_BID64(x)
	sign_y, exponent_y, coefficient_y, valid_y = unpack_BID64(y)

	// unpack arguments, check for NaN or Infinity
	if !valid_x {
		// x is Inf. or NaN

		// test if x is NaN
		if (x & NAN_MASK64) == NAN_MASK64 {
			// #ifdef BID_SET_STATUS_FLAGS
			if ((x & SNAN_MASK64) == SNAN_MASK64) || // sNaN
				((y & SNAN_MASK64) == SNAN_MASK64) {
				pfpsf |= BID_INVALID_EXCEPTION
			}
			// #endif
			res = coefficient_x & QUIET_MASK64
			return res, pfpsf
		}
		// x is Infinity?
		if (x & INFINITY_MASK64) == INFINITY_MASK64 {
			// check if y is Inf
			if (y & NAN_MASK64) == INFINITY_MASK64 {
				if sign_x == (y & 0x8000000000000000) {
					res = coefficient_x
					return res, pfpsf
				}
				// return NaN
				// #ifdef BID_SET_STATUS_FLAGS
				pfpsf |= BID_INVALID_EXCEPTION
				// #endif
				res = NAN_MASK64
				return res, pfpsf
			}
			// check if y is NaN
			if (y & NAN_MASK64) == NAN_MASK64 {
				res = coefficient_y & QUIET_MASK64
				// #ifdef BID_SET_STATUS_FLAGS
				if (y & SNAN_MASK64) == SNAN_MASK64 {
					pfpsf |= BID_INVALID_EXCEPTION
				}
				// #endif
				return res, pfpsf
			}
			// otherwise return +/-Inf
			res = coefficient_x
			return res, pfpsf
		}
		// x is 0
		if ((y & INFINITY_MASK64) != INFINITY_MASK64) && coefficient_y != 0 {
			if exponent_y <= exponent_x {
				res = y
				return res, pfpsf
			}
		}
	}

	if !valid_y {
		// y is Inf. or NaN?
		if (y & INFINITY_MASK64) == INFINITY_MASK64 {
			// #ifdef BID_SET_STATUS_FLAGS
			if (y & SNAN_MASK64) == SNAN_MASK64 { // sNaN
				pfpsf |= BID_INVALID_EXCEPTION
			}
			// #endif
			res = coefficient_y & QUIET_MASK64
			return res, pfpsf
		}
		// y is 0
		if coefficient_x == 0 { // x==0
			if exponent_x <= exponent_y {
				res = uint64(exponent_x) << 53
			} else {
				res = uint64(exponent_y) << 53
			}
			if sign_x == sign_y {
				res |= sign_x
			}
			// #ifndef IEEE_ROUND_NEAREST_TIES_AWAY
			// #ifndef IEEE_ROUND_NEAREST
			if rndMode == BID_ROUNDING_DOWN && sign_x != sign_y {
				res |= 0x8000000000000000
			}
			// #endif
			// #endif
			return res, pfpsf
		} else if exponent_y >= exponent_x {
			res = x
			return res, pfpsf
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

	// get binary coefficients of x and y
	// --- get number of bits in the coefficients of x and y ---
	// version 2 (original)
	// tempx.d = (double) coefficient_a;
	// bin_expon_ca = ((tempx.i & MASK_BINARY_EXPONENT) >> 52) - 0x3ff;
	tempx := math.Float64frombits(math.Float64bits(float64(coefficient_a)))
	bin_expon_ca = int((math.Float64bits(tempx)&MASK_BINARY_EXPONENT)>>52) - 0x3ff

	if diff_dec_expon > MAX_FORMAT_DIGITS {
		// normalize a to a 16-digit coefficient

		scale_ca = bid_estimate_decimal_digits[bin_expon_ca]
		if coefficient_a >= bid_power10_table_128[scale_ca].w[0] {
			scale_ca++
		}

		scale_k = 16 - scale_ca

		coefficient_a *= bid_power10_table_128[scale_k].w[0]

		diff_dec_expon -= scale_k
		exponent_a -= scale_k

		// get binary coefficients of x and y
		// --- get number of bits in the coefficients of x and y ---
		tempx = float64(coefficient_a)
		bin_expon_ca = int((math.Float64bits(tempx)&MASK_BINARY_EXPONENT)>>52) - 0x3ff

		if diff_dec_expon > MAX_FORMAT_DIGITS {
			// #ifdef BID_SET_STATUS_FLAGS
			if coefficient_b != 0 {
				pfpsf |= BID_INEXACT_EXCEPTION
			}
			// #endif

			// #ifndef IEEE_ROUND_NEAREST_TIES_AWAY
			// #ifndef IEEE_ROUND_NEAREST
			if (rndMode&3) != 0 && coefficient_b != 0 { // not BID_ROUNDING_TO_NEAREST
				switch rndMode {
				case BID_ROUNDING_DOWN:
					if sign_b != 0 {
						coefficient_a -= uint64((int64(sign_a) >> 63) | 1)
						if coefficient_a < 1000000000000000 {
							exponent_a--
							coefficient_a = 9999999999999999
						} else if coefficient_a >= 10000000000000000 {
							exponent_a++
							coefficient_a = 1000000000000000
						}
					}
				case BID_ROUNDING_UP:
					if sign_b == 0 {
						coefficient_a += uint64((int64(sign_a) >> 63) | 1)
						if coefficient_a < 1000000000000000 {
							exponent_a--
							coefficient_a = 9999999999999999
						} else if coefficient_a >= 10000000000000000 {
							exponent_a++
							coefficient_a = 1000000000000000
						}
					}
				default: // RZ
					if sign_a != sign_b {
						coefficient_a--
						if coefficient_a < 1000000000000000 {
							exponent_a--
							coefficient_a = 9999999999999999
						}
					}
				}
			} else {
				// #endif
				// #endif
				// check special case here
				if coefficient_a == 1000000000000000 &&
					diff_dec_expon == MAX_FORMAT_DIGITS+1 &&
					(sign_a^sign_b) != 0 &&
					coefficient_b > 5000000000000000 {
					coefficient_a = 9999999999999999
					exponent_a--
				}
			}

			res, flags := fast_get_BID64_check_OF_flags(sign_a, exponent_a, coefficient_a, rndMode)
			pfpsf |= flags
			return res, pfpsf
		}
	}

	// test whether coefficient_a*10^(exponent_a-exponent_b) may exceed 2^62
	if bin_expon_ca+bid_estimate_bin_expon[diff_dec_expon] < 60 {
		// coefficient_a*10^(exponent_a-exponent_b)<2^63

		// multiply by 10^(exponent_a-exponent_b)
		coefficient_a *= bid_power10_table_128[diff_dec_expon].w[0]

		// sign mask
		sign_b = uint64(int64(sign_b) >> 63)
		// apply sign to coeff. of b
		coefficient_b = (coefficient_b + sign_b) ^ sign_b

		// apply sign to coefficient a
		sign_a = uint64(int64(sign_a) >> 63)
		coefficient_a = (coefficient_a + sign_a) ^ sign_a

		coefficient_a += coefficient_b
		// get sign
		sign_s = uint64(int64(coefficient_a) >> 63)
		coefficient_a = (coefficient_a + sign_s) ^ sign_s
		sign_s &= 0x8000000000000000

		// coefficient_a < 10^16 ?
		if coefficient_a < bid_power10_table_128[MAX_FORMAT_DIGITS].w[0] {
			// #ifndef IEEE_ROUND_NEAREST_TIES_AWAY
			// #ifndef IEEE_ROUND_NEAREST
			if rndMode == BID_ROUNDING_DOWN && coefficient_a == 0 && sign_a != sign_b {
				sign_s = 0x8000000000000000
			}
			// #endif
			// #endif
			res = very_fast_get_BID64(sign_s, exponent_b, coefficient_a)
			return res, pfpsf
		}
		// otherwise rounding is necessary

		// already know coefficient_a<10^19
		// coefficient_a < 10^17 ?
		if coefficient_a < bid_power10_table_128[17].w[0] {
			extra_digits = 1
		} else if coefficient_a < bid_power10_table_128[18].w[0] {
			extra_digits = 2
		} else {
			extra_digits = 3
		}

		// #ifndef IEEE_ROUND_NEAREST_TIES_AWAY
		// #ifndef IEEE_ROUND_NEAREST
		rmode = rndMode
		if sign_s != 0 && uint(rmode-1) < 2 {
			rmode = 3 - rmode
		}
		// #else
		// rmode = 0
		// #endif
		// #endif

		coefficient_a += bid_round_const_table[rmode][extra_digits]

		// get P*(2^M[extra_digits])/10^extra_digits
		CT = __mul_64x64_to_128(coefficient_a, bid_reciprocals10_64[extra_digits])

		// now get P/10^extra_digits: shift C64 right by M[extra_digits]-128
		amount = bid_short_recip_scale[extra_digits]
		C64 = CT.w[1] >> uint(amount)

	} else {
		// coefficient_a*10^(exponent_a-exponent_b) is large
		sign_s = sign_a

		// #ifndef IEEE_ROUND_NEAREST_TIES_AWAY
		// #ifndef IEEE_ROUND_NEAREST
		rmode = rndMode
		if sign_s != 0 && uint(rmode-1) < 2 {
			rmode = 3 - rmode
		}
		// #else
		// rmode = 0
		// #endif
		// #endif

		// check whether we can take faster path
		scale_ca = bid_estimate_decimal_digits[bin_expon_ca]

		sign_ab = sign_a ^ sign_b
		sign_ab = uint64(int64(sign_ab) >> 63)

		// T1 = 10^(16-diff_dec_expon)
		T1 = bid_power10_table_128[16-diff_dec_expon].w[0]

		// get number of digits in coefficient_a
		if coefficient_a >= bid_power10_table_128[scale_ca].w[0] {
			scale_ca++
		}

		scale_k = 16 - scale_ca

		// addition
		saved_ca = coefficient_a - T1
		coefficient_a = uint64(int64(saved_ca) * int64(bid_power10_table_128[scale_k].w[0]))
		extra_digits = diff_dec_expon - scale_k

		// apply sign
		saved_cb = (coefficient_b + sign_ab) ^ sign_ab
		// add 10^16 and rounding constant
		coefficient_b = saved_cb + 10000000000000000 + bid_round_const_table[rmode][extra_digits]

		// get P*(2^M[extra_digits])/10^extra_digits
		CT = __mul_64x64_to_128(coefficient_b, bid_reciprocals10_64[extra_digits])

		// now get P/10^extra_digits: shift C64 right by M[extra_digits]-128
		amount = bid_short_recip_scale[extra_digits]
		C0_64 = CT.w[1] >> uint(amount)

		// result coefficient
		C64 = C0_64 + uint64(int64(coefficient_a))

		// filter out difficult (corner) cases
		// this test ensures the number of digits in coefficient_a does not change
		// after adding (the appropriately scaled and rounded) coefficient_b
		if uint64(C64-1000000000000000-1) > 9000000000000000-2 {
			if C64 >= 10000000000000000 {
				// result has more than 16 digits
				if scale_k == 0 {
					// must divide coeff_a by 10
					saved_ca = saved_ca + T1
					CA = __mul_64x64_to_128(saved_ca, 0x3333333333333334)
					// reciprocals10_64[1]
					coefficient_a = CA.w[1] >> 1
					rem_a = saved_ca - (coefficient_a << 3) - (coefficient_a << 1)
					coefficient_a = coefficient_a - T1

					saved_cb += rem_a * bid_power10_table_128[diff_dec_expon].w[0]
				} else {
					coefficient_a = uint64(int64(saved_ca-T1-(T1<<3)) * int64(bid_power10_table_128[scale_k-1].w[0]))
				}

				extra_digits++
				coefficient_b = saved_cb + 100000000000000000 + bid_round_const_table[rmode][extra_digits]

				// get P*(2^M[extra_digits])/10^extra_digits
				CT = __mul_64x64_to_128(coefficient_b, bid_reciprocals10_64[extra_digits])

				// now get P/10^extra_digits: shift C64 right by M[extra_digits]-128
				amount = bid_short_recip_scale[extra_digits]
				C0_64 = CT.w[1] >> uint(amount)

				// result coefficient
				C64 = C0_64 + uint64(int64(coefficient_a))
			} else if C64 <= 1000000000000000 {
				// less than 16 digits in result
				coefficient_a = uint64(int64(saved_ca) * int64(bid_power10_table_128[scale_k+1].w[0]))
				// extra_digits--
				exponent_b--
				coefficient_b = (saved_cb << 3) + (saved_cb << 1) + 100000000000000000 + bid_round_const_table[rmode][extra_digits]

				// get P*(2^M[extra_digits])/10^extra_digits
				CT_new = __mul_64x64_to_128(coefficient_b, bid_reciprocals10_64[extra_digits])

				// now get P/10^extra_digits: shift C64 right by M[extra_digits]-128
				amount = bid_short_recip_scale[extra_digits]
				C0_64 = CT_new.w[1] >> uint(amount)

				// result coefficient
				C64_new = C0_64 + uint64(int64(coefficient_a))
				if C64_new < 10000000000000000 {
					C64 = C64_new
					// #ifdef BID_SET_STATUS_FLAGS
					CT = CT_new
					// #endif
				} else {
					exponent_b++
				}
			}
		}
	}

	// #ifndef IEEE_ROUND_NEAREST_TIES_AWAY
	// #ifndef IEEE_ROUND_NEAREST
	if rmode == 0 { // BID_ROUNDING_TO_NEAREST
		// #endif
		if C64&1 != 0 {
			// check whether fractional part of initial_P/10^extra_digits is
			// exactly .5
			// this is the same as fractional part of
			//      (initial_P + 0.5*10^extra_digits)/10^extra_digits is exactly zero

			// get remainder
			remainder_h = CT.w[1] << (64 - uint(amount))

			// test whether fractional part is 0
			if remainder_h == 0 && CT.w[0] < bid_reciprocals10_64[extra_digits] {
				C64--
			}
		}
	}
	// #endif

	// #ifdef BID_SET_STATUS_FLAGS
	status = BID_INEXACT_EXCEPTION

	// get remainder
	remainder_h = CT.w[1] << (64 - uint(amount))

	switch rmode {
	case BID_ROUNDING_TO_NEAREST, BID_ROUNDING_TIES_AWAY:
		// test whether fractional part is 0
		if remainder_h == 0x8000000000000000 &&
			CT.w[0] < bid_reciprocals10_64[extra_digits] {
			status = BID_EXACT_STATUS
		}
	case BID_ROUNDING_DOWN, BID_ROUNDING_TO_ZERO:
		if remainder_h == 0 && CT.w[0] < bid_reciprocals10_64[extra_digits] {
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
	// #endif

	res, flags := fast_get_BID64_check_OF_flags(sign_s, exponent_b+extra_digits, C64, rndMode)
	pfpsf |= flags
	return res, pfpsf
}

// fast_get_BID64_check_OF_flags is like fast_get_BID64_check_OF but returns flags
// Ported from Intel bid_internal.h fast_get_BID64_check_OF (lines 1079-1151)
func fast_get_BID64_check_OF_flags(sgn uint64, expon int, coeff uint64, rmode int) (uint64, uint32) {
	var r uint64
	var flags uint32

	// 3 * 256 - 1 = 767 = DECIMAL_MAX_EXPON_64
	// 3 * 256 = 768
	if uint(expon) >= 3*256-1 {
		// Special case: expon == 767 and coeff == 10^16
		if expon == 3*256-1 && coeff == 10000000000000000 {
			expon = 3 * 256 // 768
			coeff = 1000000000000000
		}

		if uint(expon) >= 3*256 {
			// try to normalize coefficient
			for coeff < 1000000000000000 && expon >= 3*256 {
				expon--
				coeff = (coeff << 3) + (coeff << 1) // coeff * 10
			}

			if expon > DECIMAL_MAX_EXPON_64 {
				// overflow - set flags
				flags |= BID_OVERFLOW_EXCEPTION | BID_INEXACT_EXCEPTION

				// overflow result based on rounding mode
				r = sgn | INFINITY_MASK64
				switch rmode {
				case BID_ROUNDING_DOWN:
					if sgn == 0 {
						r = LARGEST_BID64
					}
				case BID_ROUNDING_TO_ZERO:
					r = sgn | LARGEST_BID64
				case BID_ROUNDING_UP:
					if sgn != 0 {
						r = SMALLEST_BID64
					}
				}
				return r, flags
			}
		}
	}

	mask := uint64(1) << EXPONENT_SHIFT_SMALL64

	// check whether coefficient fits in 10*5+3 bits
	if coeff < mask {
		r = uint64(expon)
		r <<= EXPONENT_SHIFT_SMALL64
		r |= (coeff | sgn)
		return r, flags
	}

	// special format
	// eliminate the case coeff==10^16 after rounding
	if coeff == 10000000000000000 {
		r = uint64(expon + 1)
		r <<= EXPONENT_SHIFT_SMALL64
		r |= (1000000000000000 | sgn)
		return r, flags
	}

	r = uint64(expon)
	r <<= EXPONENT_SHIFT_LARGE64
	r |= (sgn | SPECIAL_ENCODING_MASK64)
	// add coeff, without leading bits
	mask = (mask >> 2) - 1
	coeff &= mask
	r |= coeff

	return r, flags
}
