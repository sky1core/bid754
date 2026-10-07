// Ported from: IntelRDFPMathLib20U4/LIBRARY/src/bid32_div.c
// Version: Intel(R) Decimal Floating-Point Math Library 2.0 Update 4
//
// This file is a mechanical translation of the Intel BID library to Go.
// All logic, magic numbers, and table references are preserved exactly.

package bidgo

import (
	"math"
)

// bid32_div_pure performs BID32 division
// Ported mechanically from Intel bid32_div.c
func bid32_div_pure(x, y uint32, rndMode int) uint32 {
	var CA uint64
	var sign_x, sign_y, coefficient_x, coefficient_y, A, B uint32
	var Q, Q2, B2, B4, B5, R, T, DU, res uint32
	var valid_x, valid_y bool
	var D uint32
	var exponent_x, exponent_y, bin_expon_cx int
	var diff_expon, ed1, ed2, bin_index int
	var rmode, amount int
	var nzeros, i, j, d5 int
	var digit_h, digit_low uint32

	sign_x, exponent_x, coefficient_x, valid_x = unpack_BID32_add(x)
	sign_y, exponent_y, coefficient_y, valid_y = unpack_BID32_add(y)

	// unpack arguments, check for NaN or Infinity
	if !valid_x {
		// x is Inf. or NaN

		// test if x is NaN
		if (x & NAN_MASK32) == NAN_MASK32 {
			return coefficient_x & QUIET_MASK32
		}
		// x is Infinity?
		if (x & INFINITY_MASK32) == INFINITY_MASK32 {
			// check if y is Inf or NaN
			if (y & INFINITY_MASK32) == INFINITY_MASK32 {
				// y==Inf, return NaN
				if (y & NAN_MASK32) == INFINITY_MASK32 { // Inf/Inf
					return NAN_MASK32
				}
			} else {
				// otherwise return +/-Inf
				return ((x ^ y) & 0x80000000) | INFINITY_MASK32
			}
		}
		// x==0
		if ((y & INFINITY_MASK32) != INFINITY_MASK32) && coefficient_y == 0 {
			// y==0 , return NaN
			return NAN_MASK32
		}
		if (y & INFINITY_MASK32) != INFINITY_MASK32 {
			if (y & SPECIAL_ENCODING_MASK32) == SPECIAL_ENCODING_MASK32 {
				exponent_y = int((y >> 21) & 0xff)
			} else {
				exponent_y = int((y >> 23) & 0xff)
			}
			sign_y = y & 0x80000000

			exponent_x = exponent_x - exponent_y + DECIMAL_EXPONENT_BIAS_32
			if exponent_x > DECIMAL_MAX_EXPON_32 {
				exponent_x = DECIMAL_MAX_EXPON_32
			} else if exponent_x < 0 {
				exponent_x = 0
			}
			return (sign_x ^ sign_y) | (uint32(exponent_x) << 23)
		}

	}
	if !valid_y {
		// y is Inf. or NaN

		// test if y is NaN
		if (y & NAN_MASK32) == NAN_MASK32 {
			return coefficient_y & QUIET_MASK32
		}
		// y is Infinity?
		if (y & INFINITY_MASK32) == INFINITY_MASK32 {
			// return +/-0
			return (x ^ y) & 0x80000000
		}
		// y is 0
		return (sign_x ^ sign_y) | INFINITY_MASK32
	}

	diff_expon = exponent_x - exponent_y + DECIMAL_EXPONENT_BIAS_32

	if coefficient_x < coefficient_y {
		// get number of decimal digits for c_x, c_y

		//--- get number of bits in the coefficients of x and y ---
		tempx := float32(coefficient_x)
		tempy := float32(coefficient_y)
		bin_index = int((math.Float32bits(tempy) - math.Float32bits(tempx)) >> 23)

		A = coefficient_x * uint32(bid_power10_index_binexp[bin_index])
		B = coefficient_y

		// compare A, B
		DU = (A - B) >> 31
		ed1 = 6 + int(DU)
		ed2 = bid_estimate_decimal_digits[bin_index] + ed1
		T = uint32(bid_power10_table_128[ed1].lo)
		CA = uint64(A) * uint64(T)

		Q = 0
		diff_expon = diff_expon - ed2

	} else {
		// get c_x/c_y
		Q = coefficient_x / coefficient_y
		R = coefficient_x - coefficient_y*Q

		// will use to get number of dec. digits of Q
		tempq := float32(Q)
		bin_expon_cx = int((math.Float32bits(tempq) >> 23)) - 0x7f

		// exact result ?
		if R == 0 {
			res = get_BID32(sign_x^sign_y, diff_expon, uint64(Q), rndMode)
			return res
		}
		// get decimal digits of Q
		DU = uint32(bid_power10_index_binexp[bin_expon_cx]) - Q - 1
		DU >>= 31

		ed2 = 7 - bid_estimate_decimal_digits[bin_expon_cx] - int(DU)

		T = uint32(bid_power10_table_128[ed2].lo)
		CA = uint64(R) * uint64(T)
		B = coefficient_y

		Q *= uint32(bid_power10_table_128[ed2].lo)
		diff_expon -= ed2
	}

	Q2 = uint32(CA / uint64(B))
	B2 = B + B
	B4 = B2 + B2
	R = uint32(CA - uint64(Q2)*uint64(B))
	Q += Q2

	if R == 0 {
		// eliminate trailing zeros
		// check whether CX, CY are short
		if coefficient_x <= 1024 && coefficient_y <= 1024 {
			i = int(coefficient_y - 1)
			j = int(coefficient_x - 1)
			// difference in powers of 2 factors for Y and X
			nzeros = ed2 - bid_factors32[i][0] + bid_factors32[j][0]
			// difference in powers of 5 factors
			d5 = ed2 - bid_factors32[i][1] + bid_factors32[j][1]
			if d5 < nzeros {
				nzeros = d5
			}

			if nzeros > 0 {
				CT := uint64(Q) * bid_bid_reciprocals10_32[nzeros]
				CT >>= 32

				// now get P/10^extra_digits: shift C64 right by M[extra_digits]-128
				amount = bid_bid_bid_recip_scale32[nzeros]
				Q = uint32(CT >> uint(amount))

				diff_expon += nzeros
			}
		} else {
			nzeros = 0

			// decompose digit
			PD := uint64(Q) * 0x068DB8BB
			digit_h = uint32(PD >> 40)
			digit_low = Q - digit_h*10000

			if digit_low == 0 {
				nzeros += 4
			} else {
				digit_h = digit_low
			}

			if (digit_h & 1) == 0 {
				nzeros += int(3 & (bid_packed_10000_zeros[digit_h>>3] >> (digit_h & 7)))
			}

			if nzeros > 0 {
				CT := uint64(Q) * bid_bid_reciprocals10_32[nzeros]
				CT >>= 32

				// now get P/10^extra_digits: shift C64 right by M[extra_digits]-128
				amount = bid_bid_bid_recip_scale32[nzeros]
				Q = uint32(CT >> uint(amount))
			}
			diff_expon += nzeros
		}
		if diff_expon >= 0 {
			res = get_BID32(sign_x^sign_y, diff_expon, uint64(Q), rndMode)
			return res
		}
	}

	if diff_expon >= 0 {
		rmode = rndMode
		if (sign_x^sign_y) != 0 && uint(rmode-1) < 2 {
			rmode = 3 - rmode
		}
		switch rmode {
		case 0, BID_ROUNDING_TIES_AWAY: // round to nearest code
			// R*10
			R += R
			R = (R << 2) + R
			B5 = B4 + B
			// compare 10*R to 5*B
			R = B5 - R
			// C: R -= ((Q | (rmode >> 2)) & 1);
			R -= ((Q | uint32(rmode>>2)) & 1)
			// R<0 ?
			D = R >> 31
			Q += D
		case BID_ROUNDING_DOWN, BID_ROUNDING_TO_ZERO:
			// nothing
		default: // rounding up
			Q++
		}

		res = get_BID32(sign_x^sign_y, diff_expon, uint64(Q), rndMode)
		return res
	} else {
		// UF occurs
		rmode = rndMode
		var uf_pfpsf uint32
		res = get_BID32_UF(sign_x^sign_y, diff_expon, uint64(Q), R, rmode, &uf_pfpsf)
		return res
	}
}

// bid32_div_core performs the Intel BID32 division and returns the status
// raised by that same computation.
func bid32_div_core(x, y uint32, rndMode int) (uint32, uint32) {
	var CA uint64
	var sign_x, sign_y, coefficient_x, coefficient_y, A, B uint32
	var Q, Q2, B2, B4, B5, R, T, DU, res, flags uint32
	var valid_x, valid_y bool
	var D uint32
	var exponent_x, exponent_y, bin_expon_cx int
	var diff_expon, ed1, ed2, bin_index int
	var rmode, amount int
	var nzeros, i, j, d5 int
	var digit_h, digit_low uint32

	sign_x, exponent_x, coefficient_x, valid_x = unpack_BID32_add(x)
	sign_y, exponent_y, coefficient_y, valid_y = unpack_BID32_add(y)

	// unpack arguments, check for NaN or Infinity
	if !valid_x {
		if (y & SNAN_MASK32) == SNAN_MASK32 {
			flags |= BID_INVALID_EXCEPTION
		}
		// x is Inf. or NaN

		// test if x is NaN
		if (x & NAN_MASK32) == NAN_MASK32 {
			if (x & SNAN_MASK32) == SNAN_MASK32 {
				flags |= BID_INVALID_EXCEPTION
			}
			return coefficient_x & QUIET_MASK32, flags
		}
		// x is Infinity?
		if (x & INFINITY_MASK32) == INFINITY_MASK32 {
			// check if y is Inf or NaN
			if (y & INFINITY_MASK32) == INFINITY_MASK32 {
				// y==Inf, return NaN
				if (y & NAN_MASK32) == INFINITY_MASK32 { // Inf/Inf
					flags |= BID_INVALID_EXCEPTION
					return NAN_MASK32, flags
				}
			} else {
				// otherwise return +/-Inf
				return ((x ^ y) & 0x80000000) | INFINITY_MASK32, flags
			}
		}
		// x==0
		if ((y & INFINITY_MASK32) != INFINITY_MASK32) && coefficient_y == 0 {
			// y==0 , return NaN
			flags |= BID_INVALID_EXCEPTION
			return NAN_MASK32, flags
		}
		if (y & INFINITY_MASK32) != INFINITY_MASK32 {
			if (y & SPECIAL_ENCODING_MASK32) == SPECIAL_ENCODING_MASK32 {
				exponent_y = int((y >> 21) & 0xff)
			} else {
				exponent_y = int((y >> 23) & 0xff)
			}
			sign_y = y & 0x80000000

			exponent_x = exponent_x - exponent_y + DECIMAL_EXPONENT_BIAS_32
			if exponent_x > DECIMAL_MAX_EXPON_32 {
				exponent_x = DECIMAL_MAX_EXPON_32
			} else if exponent_x < 0 {
				exponent_x = 0
			}
			return (sign_x ^ sign_y) | (uint32(exponent_x) << 23), flags
		}

	}
	if !valid_y {
		// y is Inf. or NaN

		// test if y is NaN
		if (y & NAN_MASK32) == NAN_MASK32 {
			if (y & SNAN_MASK32) == SNAN_MASK32 {
				flags |= BID_INVALID_EXCEPTION
			}
			return coefficient_y & QUIET_MASK32, flags
		}
		// y is Infinity?
		if (y & INFINITY_MASK32) == INFINITY_MASK32 {
			// return +/-0
			return (x ^ y) & 0x80000000, flags
		}
		// y is 0
		flags |= BID_ZERO_DIVIDE_EXCEPTION
		return (sign_x ^ sign_y) | INFINITY_MASK32, flags
	}

	diff_expon = exponent_x - exponent_y + DECIMAL_EXPONENT_BIAS_32

	if coefficient_x < coefficient_y {
		// get number of decimal digits for c_x, c_y

		//--- get number of bits in the coefficients of x and y ---
		tempx := float32(coefficient_x)
		tempy := float32(coefficient_y)
		bin_index = int((math.Float32bits(tempy) - math.Float32bits(tempx)) >> 23)

		A = coefficient_x * uint32(bid_power10_index_binexp[bin_index])
		B = coefficient_y

		// compare A, B
		DU = (A - B) >> 31
		ed1 = 6 + int(DU)
		ed2 = bid_estimate_decimal_digits[bin_index] + ed1
		T = uint32(bid_power10_table_128[ed1].lo)
		CA = uint64(A) * uint64(T)

		Q = 0
		diff_expon = diff_expon - ed2

	} else {
		// get c_x/c_y
		Q = coefficient_x / coefficient_y
		R = coefficient_x - coefficient_y*Q

		// will use to get number of dec. digits of Q
		tempq := float32(Q)
		bin_expon_cx = int((math.Float32bits(tempq) >> 23)) - 0x7f

		// exact result ?
		if R == 0 {
			res = get_BID32_flags(sign_x^sign_y, diff_expon, uint64(Q), rndMode, &flags)
			return res, flags
		}
		// get decimal digits of Q
		DU = uint32(bid_power10_index_binexp[bin_expon_cx]) - Q - 1
		DU >>= 31

		ed2 = 7 - bid_estimate_decimal_digits[bin_expon_cx] - int(DU)

		T = uint32(bid_power10_table_128[ed2].lo)
		CA = uint64(R) * uint64(T)
		B = coefficient_y

		Q *= uint32(bid_power10_table_128[ed2].lo)
		diff_expon -= ed2
	}

	Q2 = uint32(CA / uint64(B))
	B2 = B + B
	B4 = B2 + B2
	R = uint32(CA - uint64(Q2)*uint64(B))
	Q += Q2
	if R != 0 {
		flags |= BID_INEXACT_EXCEPTION
	}

	if R == 0 {
		// eliminate trailing zeros
		// check whether CX, CY are short
		if coefficient_x <= 1024 && coefficient_y <= 1024 {
			i = int(coefficient_y - 1)
			j = int(coefficient_x - 1)
			// difference in powers of 2 factors for Y and X
			nzeros = ed2 - bid_factors32[i][0] + bid_factors32[j][0]
			// difference in powers of 5 factors
			d5 = ed2 - bid_factors32[i][1] + bid_factors32[j][1]
			if d5 < nzeros {
				nzeros = d5
			}

			if nzeros > 0 {
				CT := uint64(Q) * bid_bid_reciprocals10_32[nzeros]
				CT >>= 32

				// now get P/10^extra_digits: shift C64 right by M[extra_digits]-128
				amount = bid_bid_bid_recip_scale32[nzeros]
				Q = uint32(CT >> uint(amount))

				diff_expon += nzeros
			}
		} else {
			nzeros = 0

			// decompose digit
			PD := uint64(Q) * 0x068DB8BB
			digit_h = uint32(PD >> 40)
			digit_low = Q - digit_h*10000

			if digit_low == 0 {
				nzeros += 4
			} else {
				digit_h = digit_low
			}

			if (digit_h & 1) == 0 {
				nzeros += int(3 & (bid_packed_10000_zeros[digit_h>>3] >> (digit_h & 7)))
			}

			if nzeros > 0 {
				CT := uint64(Q) * bid_bid_reciprocals10_32[nzeros]
				CT >>= 32

				// now get P/10^extra_digits: shift C64 right by M[extra_digits]-128
				amount = bid_bid_bid_recip_scale32[nzeros]
				Q = uint32(CT >> uint(amount))
			}
			diff_expon += nzeros
		}
		if diff_expon >= 0 {
			res = get_BID32_flags(sign_x^sign_y, diff_expon, uint64(Q), rndMode, &flags)
			return res, flags
		}
	}

	if diff_expon >= 0 {
		rmode = rndMode
		if (sign_x^sign_y) != 0 && uint(rmode-1) < 2 {
			rmode = 3 - rmode
		}
		switch rmode {
		case 0, BID_ROUNDING_TIES_AWAY: // round to nearest code
			// R*10
			R += R
			R = (R << 2) + R
			B5 = B4 + B
			// compare 10*R to 5*B
			R = B5 - R
			// C: R -= ((Q | (rmode >> 2)) & 1);
			R -= ((Q | uint32(rmode>>2)) & 1)
			// R<0 ?
			D = R >> 31
			Q += D
		case BID_ROUNDING_DOWN, BID_ROUNDING_TO_ZERO:
			// nothing
		default: // rounding up
			Q++
		}

		res = get_BID32_flags(sign_x^sign_y, diff_expon, uint64(Q), rndMode, &flags)
		return res, flags
	} else {
		// UF occurs
		if diff_expon+7 < 0 {
			flags |= BID_INEXACT_EXCEPTION
		}
		rmode = rndMode
		res = get_BID32_UF(sign_x^sign_y, diff_expon, uint64(Q), R, rmode, &flags)
		return res, flags
	}
}
