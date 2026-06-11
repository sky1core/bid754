// Ported from: IntelRDFPMathLib20U4/LIBRARY/src/Bid64Div.c
// Version: Intel(R) Decimal Floating-Point Math Library 2.0 Update 4
//
// This file is a mechanical translation of the Intel BID library to Go.
// All logic, magic numbers, and table references are preserved exactly.

package bidgo

import "math"

// Bid64Div divides x by y
// Ported from Bid64Div in Bid64Div.c (line-by-line mechanical translation)
func Bid64Div(x, y uint64, rndMode int) uint64 {
	var CA BID_UINT128
	var sign_x, sign_y, coefficient_x, coefficient_y, A, B, Q, Q2, R, T, DU, res uint64
	var B2, B4, B5 uint64
	var valid_x, valid_y bool
	var exponent_x, exponent_y, bin_expon_cx int
	var diff_expon, ed1, ed2, bin_index int
	var rmode int
	var D int64
	var db float64

	sign_x, exponent_x, coefficient_x, valid_x = unpack_BID64(x)
	sign_y, exponent_y, coefficient_y, valid_y = unpack_BID64(y)

	// unpack arguments, check for NaN or Infinity
	if !valid_x {
		// x is Inf. or NaN

		// test if x is NaN
		if (x & NAN_MASK64) == NAN_MASK64 {
			return coefficient_x & QUIET_MASK64
		}
		// x is Infinity?
		if (x & INFINITY_MASK64) == INFINITY_MASK64 {
			// check if y is Inf or NaN
			if (y & INFINITY_MASK64) == INFINITY_MASK64 {
				// y==Inf, return NaN
				if (y & NAN_MASK64) == INFINITY_MASK64 { // Inf/Inf
					return NAN_MASK64
				}
			} else {
				// otherwise return +/-Inf
				return ((x ^ y) & 0x8000000000000000) | INFINITY_MASK64
			}
		}
		// x==0
		if ((y & INFINITY_MASK64) != INFINITY_MASK64) && coefficient_y == 0 {
			// y==0 , return NaN
			return NAN_MASK64
		}
		if (y & INFINITY_MASK64) != INFINITY_MASK64 {
			if (y & SPECIAL_ENCODING_MASK64) == SPECIAL_ENCODING_MASK64 {
				exponent_y = int((uint32(y>>51) & 0x3ff))
			} else {
				exponent_y = int((uint32(y>>53) & 0x3ff))
			}
			sign_y = y & 0x8000000000000000

			exponent_x = exponent_x - exponent_y + DECIMAL_EXPONENT_BIAS
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
			return coefficient_y & QUIET_MASK64
		}
		// y is Infinity?
		if (y & INFINITY_MASK64) == INFINITY_MASK64 {
			// return +/-0
			return (x ^ y) & 0x8000000000000000
		}
		// y is 0
		return (sign_x ^ sign_y) | INFINITY_MASK64
	}

	diff_expon = exponent_x - exponent_y + DECIMAL_EXPONENT_BIAS

	if coefficient_x < coefficient_y {
		// get number of decimal digits for c_x, c_y

		//--- get number of bits in the coefficients of x and y ---
		tempx := math.Float32bits(float32(coefficient_x))
		tempy := math.Float32bits(float32(coefficient_y))
		bin_index = int((tempy - tempx) >> 23)

		A = coefficient_x * bid_power10_index_binexp[bin_index]
		B = coefficient_y

		temp_b := float64(B)

		// compare A, B
		DU = (A - B) >> 63
		ed1 = 15 + int(DU)
		ed2 = bid_estimate_decimal_digits[bin_index] + ed1
		T = bid_power10_table_128[ed1].w[0]
		CA = __mul_64x64_to_128(A, T)

		Q = 0
		diff_expon = diff_expon - ed2

		// adjust double precision db, to ensure that later A/B - (int)(da/db) > -1
		if coefficient_y < 0x0020000000000000 {
			temp_b_bits := math.Float64bits(temp_b)
			temp_b_bits += 1
			db = math.Float64frombits(temp_b_bits)
		} else {
			db = float64(B + 2 + (B & 1))
		}

	} else {
		// get c_x/c_y

		// set last bit before conversion to DP
		A2 := coefficient_x | 1
		da := float64(A2)

		db = float64(coefficient_y)

		dq := da / db
		Q = uint64(dq)

		R = coefficient_x - coefficient_y*Q

		// will use to get number of dec. digits of Q
		tempq := math.Float64bits(dq)
		bin_expon_cx = int((tempq >> 52)) - 0x3ff

		// R<0 ?
		D = int64(R) >> 63
		Q += uint64(D)
		R += coefficient_y & uint64(D)

		// exact result ?
		if int64(R) <= 0 {
			// can have R==-1 for coeff_y==1
			res = get_BID64(sign_x^sign_y, diff_expon, Q+R, rndMode)
			return res
		}

		// get decimal digits of Q
		DU = bid_power10_index_binexp[bin_expon_cx] - Q - 1
		DU >>= 63

		ed2 = 16 - bid_estimate_decimal_digits[bin_expon_cx] - int(DU)

		T = bid_power10_table_128[ed2].w[0]
		CA = __mul_64x64_to_128(R, T)
		B = coefficient_y

		Q *= bid_power10_table_128[ed2].w[0]
		diff_expon -= ed2
	}

	if CA.w[1] == 0 {
		Q2 = CA.w[0] / B
		B2 = B + B
		B4 = B2 + B2
		R = CA.w[0] - Q2*B
		Q += Q2
	} else {
		// 2^64
		t_scale := math.Float64frombits(0x43f0000000000000)
		// convert CA to DP
		da_h := float64(CA.w[1])
		da_l := float64(CA.w[0])
		da := noFmaMulAddF64(da_h, t_scale, da_l)

		// quotient
		dq := da / db
		Q2 = uint64(dq)

		// get w[0] remainder
		R = CA.w[0] - Q2*B

		// R<0 ?
		D = int64(R) >> 63
		Q2 += uint64(D)
		R += B & uint64(D)

		// now R<6*B

		// quick divide

		// 4*B
		B2 = B + B
		B4 = B2 + B2

		R = R - B4
		// R<0 ?
		D = int64(R) >> 63
		// restore R if negative
		R += B4 & uint64(D)
		Q2 += (^uint64(D)) & 4

		R = R - B2
		// R<0 ?
		D = int64(R) >> 63
		// restore R if negative
		R += B2 & uint64(D)
		Q2 += (^uint64(D)) & 2

		R = R - B
		// R<0 ?
		D = int64(R) >> 63
		// restore R if negative
		R += B & uint64(D)
		Q2 += (^uint64(D)) & 1

		Q += Q2
	}

	// Trailing zero elimination (when R == 0, exact division)
	if R == 0 {
		// eliminate trailing zeros
		var nzeros int

		// check whether CX, CY are short
		if coefficient_x <= 1024 && coefficient_y <= 1024 {
			i := int(coefficient_y) - 1
			j := int(coefficient_x) - 1
			// difference in powers of 2 bid_factors for Y and X
			nzeros = ed2 - int(bid_factors[i][0]) + int(bid_factors[j][0])
			// difference in powers of 5 bid_factors
			d5 := ed2 - int(bid_factors[i][1]) + int(bid_factors[j][1])
			if d5 < nzeros {
				nzeros = d5
			}

			if nzeros > 0 {
				CT := __mul_64x64_to_128(Q, bid_reciprocals10_64[nzeros])
				// now get P/10^extra_digits: shift C64 right by M[extra_digits]-128
				amount := bid_short_recip_scale[nzeros]
				Q = CT.w[1] >> uint(amount)
			}

			diff_expon += nzeros
		} else {
			var tdigit [3]uint32
			tdigit[0] = uint32(Q & 0x3ffffff)
			tdigit[1] = 0
			QX := Q >> 26
			QX32 := uint32(QX)
			nzeros = 0

			for j := 0; QX32 != 0; j, QX32 = j+1, QX32>>7 {
				k := int(QX32 & 127)
				tdigit[0] += bid_convert_table[j][k][0]
				tdigit[1] += bid_convert_table[j][k][1]
				if tdigit[0] >= 100000000 {
					tdigit[0] -= 100000000
					tdigit[1]++
				}
			}

			digit := tdigit[0]
			if digit == 0 && tdigit[1] == 0 {
				nzeros += 16
			} else {
				if digit == 0 {
					nzeros += 8
					digit = tdigit[1]
				}
				// decompose digit
				PD := uint64(digit) * 0x068DB8BB
				digit_h := uint32(PD >> 40)
				digit_low := digit - digit_h*10000

				if digit_low == 0 {
					nzeros += 4
				} else {
					digit_h = digit_low
				}

				if (digit_h & 1) == 0 {
					nzeros += int(3 & (uint32(bid_packed_10000_zeros[digit_h>>3]) >> (digit_h & 7)))
				}
			}

			if nzeros > 0 {
				CT := __mul_64x64_to_128(Q, bid_reciprocals10_64[nzeros])
				// now get P/10^extra_digits: shift C64 right by M[extra_digits]-128
				amount := bid_short_recip_scale[nzeros]
				Q = CT.w[1] >> uint(amount)
			}
			diff_expon += nzeros
		}

		if diff_expon >= 0 {
			res = fast_get_BID64_check_OF(sign_x^sign_y, diff_expon, Q, rndMode)
			return res
		}
	}

	if diff_expon >= 0 {
		rmode = rndMode
		if (sign_x^sign_y) != 0 && uint(rmode-1) < 2 {
			rmode = 3 - rmode
		}

		switch rmode {
		case 0, BID_ROUNDING_TIES_AWAY: // round to nearest (half_even, half_up)
			// R*10
			R += R
			R = (R << 2) + R
			B5 = B4 + B
			// compare 10*R to 5*B
			R = B5 - R
			// correction for (R==0 && (Q&1))
			R -= (Q | (uint64(rmode) >> 2)) & 1
			// R<0 ? (NOTE: C uses ((BID_UINT64) R) >> 63, unsigned shift giving 0 or 1)
			Q += R >> 63
		case BID_ROUNDING_DOWN, BID_ROUNDING_TO_ZERO:
			// no adjustment needed
		default: // rounding up
			Q++
		}

		res = fast_get_BID64_check_OF(sign_x^sign_y, diff_expon, Q, rndMode)
		return res
	} else {
		// UF occurs
		rmode = rndMode
		res = get_BID64_UF(sign_x^sign_y, diff_expon, Q, R, rmode)
		return res
	}
}

// Bid64DivWithFlags divides x by y, returning result and status flags
// Ported from Intel bid64_div.c with flag tracking
func Bid64DivWithFlags(x, y uint64, rndMode int) (uint64, uint32) {
	var CA BID_UINT128
	var sign_x, sign_y, coefficient_x, coefficient_y, A, B, Q, Q2, R, T, DU, res uint64
	var B2, B4, B5 uint64
	var valid_x, valid_y bool
	var exponent_x, exponent_y, bin_expon_cx int
	var diff_expon, ed1, ed2, bin_index int
	var rmode int
	var D int64
	var db float64
	var pfpsf uint32

	sign_x, exponent_x, coefficient_x, valid_x = unpack_BID64(x)
	sign_y, exponent_y, coefficient_y, valid_y = unpack_BID64(y)

	// unpack arguments, check for NaN or Infinity
	if !valid_x {
		// check for SNaN in y first
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
			// check if y is Inf or NaN
			if (y & INFINITY_MASK64) == INFINITY_MASK64 {
				// y==Inf, return NaN
				if (y & NAN_MASK64) == INFINITY_MASK64 { // Inf/Inf
					pfpsf |= BID_INVALID_EXCEPTION
					return NAN_MASK64, pfpsf
				}
			} else {
				// otherwise return +/-Inf
				return ((x ^ y) & 0x8000000000000000) | INFINITY_MASK64, pfpsf
			}
		}
		// x==0
		if ((y & INFINITY_MASK64) != INFINITY_MASK64) && coefficient_y == 0 {
			// y==0, 0/0 return NaN
			pfpsf |= BID_INVALID_EXCEPTION
			return NAN_MASK64, pfpsf
		}
		if (y & INFINITY_MASK64) != INFINITY_MASK64 {
			if (y & SPECIAL_ENCODING_MASK64) == SPECIAL_ENCODING_MASK64 {
				exponent_y = int((uint32(y>>51) & 0x3ff))
			} else {
				exponent_y = int((uint32(y>>53) & 0x3ff))
			}
			sign_y = y & 0x8000000000000000

			exponent_x = exponent_x - exponent_y + DECIMAL_EXPONENT_BIAS
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
			// return +/-0
			return (x ^ y) & 0x8000000000000000, pfpsf
		}
		// y is 0 (division by zero)
		pfpsf |= BID_ZERO_DIVIDE_EXCEPTION
		return (sign_x ^ sign_y) | INFINITY_MASK64, pfpsf
	}

	diff_expon = exponent_x - exponent_y + DECIMAL_EXPONENT_BIAS

	if coefficient_x < coefficient_y {
		// get number of decimal digits for c_x, c_y

		//--- get number of bits in the coefficients of x and y ---
		tempx := math.Float32bits(float32(coefficient_x))
		tempy := math.Float32bits(float32(coefficient_y))
		bin_index = int((tempy - tempx) >> 23)

		A = coefficient_x * bid_power10_index_binexp[bin_index]
		B = coefficient_y

		temp_b := float64(B)

		// compare A, B
		DU = (A - B) >> 63
		ed1 = 15 + int(DU)
		ed2 = bid_estimate_decimal_digits[bin_index] + ed1
		T = bid_power10_table_128[ed1].w[0]
		CA = __mul_64x64_to_128(A, T)

		Q = 0
		diff_expon = diff_expon - ed2

		// adjust double precision db, to ensure that later A/B - (int)(da/db) > -1
		if coefficient_y < 0x0020000000000000 {
			temp_b_bits := math.Float64bits(temp_b)
			temp_b_bits += 1
			db = math.Float64frombits(temp_b_bits)
		} else {
			db = float64(B + 2 + (B & 1))
		}

	} else {
		// get c_x/c_y

		// set last bit before conversion to DP
		A2 := coefficient_x | 1
		da := float64(A2)

		db = float64(coefficient_y)

		dq := da / db
		Q = uint64(dq)

		R = coefficient_x - coefficient_y*Q

		// will use to get number of dec. digits of Q
		tempq := math.Float64bits(dq)
		bin_expon_cx = int((tempq >> 52)) - 0x3ff

		// R<0 ?
		D = int64(R) >> 63
		Q += uint64(D)
		R += coefficient_y & uint64(D)

		// exact result ?
		if int64(R) <= 0 {
			// can have R==-1 for coeff_y==1
			res, flags := get_BID64_flags(sign_x^sign_y, diff_expon, Q+R, rndMode)
			pfpsf |= flags
			return res, pfpsf
		}

		// get decimal digits of Q
		DU = bid_power10_index_binexp[bin_expon_cx] - Q - 1
		DU >>= 63

		ed2 = 16 - bid_estimate_decimal_digits[bin_expon_cx] - int(DU)

		T = bid_power10_table_128[ed2].w[0]
		CA = __mul_64x64_to_128(R, T)
		B = coefficient_y

		Q *= bid_power10_table_128[ed2].w[0]
		diff_expon -= ed2
	}

	if CA.w[1] == 0 {
		Q2 = CA.w[0] / B
		B2 = B + B
		B4 = B2 + B2
		R = CA.w[0] - Q2*B
		Q += Q2
	} else {
		// 2^64
		t_scale := math.Float64frombits(0x43f0000000000000)
		// convert CA to DP
		da_h := float64(CA.w[1])
		da_l := float64(CA.w[0])
		da := noFmaMulAddF64(da_h, t_scale, da_l)

		// quotient
		dq := da / db
		Q2 = uint64(dq)

		// get w[0] remainder
		R = CA.w[0] - Q2*B

		// R<0 ?
		D = int64(R) >> 63
		Q2 += uint64(D)
		R += B & uint64(D)

		// now R<6*B

		// quick divide

		// 4*B
		B2 = B + B
		B4 = B2 + B2

		R = R - B4
		// R<0 ?
		D = int64(R) >> 63
		// restore R if negative
		R += B4 & uint64(D)
		Q2 += (^uint64(D)) & 4

		R = R - B2
		// R<0 ?
		D = int64(R) >> 63
		// restore R if negative
		R += B2 & uint64(D)
		Q2 += (^uint64(D)) & 2

		R = R - B
		// R<0 ?
		D = int64(R) >> 63
		// restore R if negative
		R += B & uint64(D)
		Q2 += (^uint64(D)) & 1

		Q += Q2
	}

	// Set INEXACT if remainder is non-zero
	if R != 0 {
		pfpsf |= BID_INEXACT_EXCEPTION
	}

	// Trailing zero elimination (when R == 0, exact division)
	if R == 0 {
		// eliminate trailing zeros
		var nzeros int

		// check whether CX, CY are short
		if coefficient_x <= 1024 && coefficient_y <= 1024 {
			i := int(coefficient_y) - 1
			j := int(coefficient_x) - 1
			// difference in powers of 2 bid_factors for Y and X
			nzeros = ed2 - int(bid_factors[i][0]) + int(bid_factors[j][0])
			// difference in powers of 5 bid_factors
			d5 := ed2 - int(bid_factors[i][1]) + int(bid_factors[j][1])
			if d5 < nzeros {
				nzeros = d5
			}

			if nzeros > 0 {
				CT := __mul_64x64_to_128(Q, bid_reciprocals10_64[nzeros])
				// now get P/10^extra_digits: shift C64 right by M[extra_digits]-128
				amount := bid_short_recip_scale[nzeros]
				Q = CT.w[1] >> uint(amount)
			}

			diff_expon += nzeros
		} else {
			var tdigit [3]uint32
			tdigit[0] = uint32(Q & 0x3ffffff)
			tdigit[1] = 0
			QX := Q >> 26
			QX32 := uint32(QX)
			nzeros = 0

			for j := 0; QX32 != 0; j, QX32 = j+1, QX32>>7 {
				k := int(QX32 & 127)
				tdigit[0] += bid_convert_table[j][k][0]
				tdigit[1] += bid_convert_table[j][k][1]
				if tdigit[0] >= 100000000 {
					tdigit[0] -= 100000000
					tdigit[1]++
				}
			}

			digit := tdigit[0]
			if digit == 0 && tdigit[1] == 0 {
				nzeros += 16
			} else {
				if digit == 0 {
					nzeros += 8
					digit = tdigit[1]
				}
				// decompose digit
				PD := uint64(digit) * 0x068DB8BB
				digit_h := uint32(PD >> 40)
				digit_low := digit - digit_h*10000

				if digit_low == 0 {
					nzeros += 4
				} else {
					digit_h = digit_low
				}

				if (digit_h & 1) == 0 {
					nzeros += int(3 & (uint32(bid_packed_10000_zeros[digit_h>>3]) >> (digit_h & 7)))
				}
			}

			if nzeros > 0 {
				CT := __mul_64x64_to_128(Q, bid_reciprocals10_64[nzeros])
				// now get P/10^extra_digits: shift C64 right by M[extra_digits]-128
				amount := bid_short_recip_scale[nzeros]
				Q = CT.w[1] >> uint(amount)
			}
			diff_expon += nzeros
		}

		if diff_expon >= 0 {
			res, flags := fast_get_BID64_check_OF_flags(sign_x^sign_y, diff_expon, Q, rndMode)
			pfpsf |= flags
			return res, pfpsf
		}
	}

	if diff_expon >= 0 {
		rmode = rndMode
		if (sign_x^sign_y) != 0 && uint(rmode-1) < 2 {
			rmode = 3 - rmode
		}

		switch rmode {
		case 0, BID_ROUNDING_TIES_AWAY: // round to nearest (half_even, half_up)
			// R*10
			R += R
			R = (R << 2) + R
			B5 = B4 + B
			// compare 10*R to 5*B
			R = B5 - R
			// correction for (R==0 && (Q&1))
			R -= (Q | (uint64(rmode) >> 2)) & 1
			// R<0 ?
			Q += R >> 63
		case BID_ROUNDING_DOWN, BID_ROUNDING_TO_ZERO:
			// no adjustment needed
		default: // rounding up
			Q++
		}

		res, flags := fast_get_BID64_check_OF_flags(sign_x^sign_y, diff_expon, Q, rndMode)
		pfpsf |= flags
		return res, pfpsf
	} else {
		// UF occurs
		if diff_expon+16 < 0 {
			pfpsf |= BID_INEXACT_EXCEPTION
		}
		rmode = rndMode
		res = get_BID64_UF_withFlags(sign_x^sign_y, diff_expon, Q, R, rmode, &pfpsf)
		return res, pfpsf
	}
}
