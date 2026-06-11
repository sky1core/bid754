package bidgo

import "math"

// Bid64Fma is ported mechanically from Intel bid64_fma.c.
func Bid64Fma(x, y, z uint64, rndMode int) (uint64, uint32) {
	var P, CT, CZ BID_UINT128
	var sign_x, sign_y, coefficient_x, coefficient_y, sign_z, coefficient_z uint64
	var C64, remainder_y, res uint64
	var CYh, CY0L, T uint64
	var valid_x, valid_y, valid_z bool
	var extra_digits, exponent_x, exponent_y, bin_expon_cx, bin_expon_cy int
	var bin_expon_product int
	var digits_p, bp, final_exponent, exponent_z, digits_z, ez, ey, scale_z int
	var uf_status uint32
	var pfpsf uint32

	sign_x, exponent_x, coefficient_x, valid_x = unpack_BID64(x)
	sign_y, exponent_y, coefficient_y, valid_y = unpack_BID64(y)
	sign_z, exponent_z, coefficient_z, valid_z = unpack_BID64(z)

	// unpack arguments, check for NaN, Infinity, or 0
	if !valid_x || !valid_y || !valid_z {
		if (y & NAN_MASK64) == NAN_MASK64 {
			// check first for non-canonical NaN payload
			y = y & 0xfe03ffffffffffff // clear G6-G12
			if (y & 0x0003ffffffffffff) > 999999999999999 {
				y = y & 0xfe00000000000000 // clear G6-G12 and the payload bits
			}
			if (y & SNAN_MASK64) == SNAN_MASK64 {
				pfpsf |= BID_INVALID_EXCEPTION
				res = y & 0xfdffffffffffffff
			} else {
				res = y
				if (z&SNAN_MASK64) == SNAN_MASK64 || (x&SNAN_MASK64) == SNAN_MASK64 {
					pfpsf |= BID_INVALID_EXCEPTION
				}
			}
			return res, pfpsf
		} else if (z & NAN_MASK64) == NAN_MASK64 {
			// check first for non-canonical NaN payload
			z = z & 0xfe03ffffffffffff // clear G6-G12
			if (z & 0x0003ffffffffffff) > 999999999999999 {
				z = z & 0xfe00000000000000 // clear G6-G12 and the payload bits
			}
			if (z & SNAN_MASK64) == SNAN_MASK64 {
				pfpsf |= BID_INVALID_EXCEPTION
				res = z & 0xfdffffffffffffff
			} else {
				res = z
				if (x & SNAN_MASK64) == SNAN_MASK64 {
					pfpsf |= BID_INVALID_EXCEPTION
				}
			}
			return res, pfpsf
		} else if (x & NAN_MASK64) == NAN_MASK64 {
			// check first for non-canonical NaN payload
			x = x & 0xfe03ffffffffffff // clear G6-G12
			if (x & 0x0003ffffffffffff) > 999999999999999 {
				x = x & 0xfe00000000000000 // clear G6-G12 and the payload bits
			}
			if (x & SNAN_MASK64) == SNAN_MASK64 {
				pfpsf |= BID_INVALID_EXCEPTION
				res = x & 0xfdffffffffffffff
			} else {
				res = x
			}
			return res, pfpsf
		}

		if !valid_x {
			// x is Inf. or 0
			if (x & INFINITY_MASK64) == INFINITY_MASK64 {
				// check if y is 0
				if coefficient_y == 0 {
					if (z & SNAN_MASK64) != NAN_MASK64 {
						pfpsf |= BID_INVALID_EXCEPTION
					}
					return NAN_MASK64, pfpsf
				}
				// test if z is Inf of opposite sign
				if ((z & NAN_MASK64) == INFINITY_MASK64) && (((x^y)^z)&0x8000000000000000) != 0 {
					pfpsf |= BID_INVALID_EXCEPTION
					return NAN_MASK64, pfpsf
				}
				// otherwise return +/-Inf
				return ((x ^ y) & 0x8000000000000000) | INFINITY_MASK64, pfpsf
			}
			// x is 0
			if ((y & INFINITY_MASK64) != INFINITY_MASK64) && ((z & INFINITY_MASK64) != INFINITY_MASK64) {
				if coefficient_z != 0 {
					exponent_y = exponent_x - DECIMAL_EXPONENT_BIAS + exponent_y
					sign_z = z & 0x8000000000000000
					if exponent_y >= exponent_z {
						return z, pfpsf
					}
					res = add_zero64(exponent_y, sign_z, exponent_z, coefficient_z, &rndMode, &pfpsf)
					return res, pfpsf
				}
			}
		}
		if !valid_y {
			// y is Inf. or 0
			if (y & INFINITY_MASK64) == INFINITY_MASK64 {
				// check if x is 0
				if coefficient_x == 0 {
					pfpsf |= BID_INVALID_EXCEPTION
					return NAN_MASK64, pfpsf
				}
				// test if z is Inf of opposite sign
				if ((z & NAN_MASK64) == INFINITY_MASK64) && (((x^y)^z)&0x8000000000000000) != 0 {
					pfpsf |= BID_INVALID_EXCEPTION
					return NAN_MASK64, pfpsf
				}
				return ((x ^ y) & 0x8000000000000000) | INFINITY_MASK64, pfpsf
			}
			// y is 0
			if (z & INFINITY_MASK64) != INFINITY_MASK64 {
				if coefficient_z != 0 {
					exponent_y += exponent_x - DECIMAL_EXPONENT_BIAS
					sign_z = z & 0x8000000000000000
					if exponent_y >= exponent_z {
						return z, pfpsf
					}
					res = add_zero64(exponent_y, sign_z, exponent_z, coefficient_z, &rndMode, &pfpsf)
					return res, pfpsf
				}
			}
		}

		if !valid_z {
			// z is Inf. or 0
			if (z & INFINITY_MASK64) == INFINITY_MASK64 {
				return coefficient_z & QUIET_MASK64, pfpsf
			}
			// z is 0, return x*y
			if coefficient_x == 0 || coefficient_y == 0 {
				exponent_x += exponent_y - DECIMAL_EXPONENT_BIAS
				if exponent_x > DECIMAL_MAX_EXPON_64 {
					exponent_x = DECIMAL_MAX_EXPON_64
				} else if exponent_x < 0 {
					exponent_x = 0
				}
				if exponent_x <= exponent_z {
					res = uint64(exponent_x) << 53
				} else {
					res = uint64(exponent_z) << 53
				}
				if (sign_x ^ sign_y) == sign_z {
					res |= sign_z
				} else if rndMode == BID_ROUNDING_DOWN {
					res |= 0x8000000000000000
				}
				return res, pfpsf
			}
		}
	}

	// --- get number of bits in the coefficients of x and y ---
	tempx := math.Float64bits(float64(coefficient_x))
	bin_expon_cx = int((tempx & MASK_BINARY_EXPONENT) >> 52)

	tempy := math.Float64bits(float64(coefficient_y))
	bin_expon_cy = int((tempy & MASK_BINARY_EXPONENT) >> 52)

	// magnitude estimate for coefficient_x*coefficient_y is 2^(...)
	bin_expon_product = bin_expon_cx + bin_expon_cy

	if bin_expon_product < UPPER_EXPON_LIMIT+2*BINARY_EXPONENT_BIAS {
		// easy multiply
		C64 = coefficient_x * coefficient_y
		final_exponent = exponent_x + exponent_y - DECIMAL_EXPONENT_BIAS
		if (final_exponent > 0) || (coefficient_z == 0) {
			res = bid_get_add64(sign_x^sign_y,
				final_exponent, C64, sign_z, exponent_z, coefficient_z, rndMode, &pfpsf)
			return res, pfpsf
		}
		P.w[0] = C64
		P.w[1] = 0
		extra_digits = 0
	} else {
		if coefficient_z == 0 {
			res, flags := Bid64MulWithFlags(x, y, rndMode)
			pfpsf |= flags
			return res, pfpsf
		}
		// get 128-bit product: coefficient_x*coefficient_y
		P = __mul_64x64_to_128(coefficient_x, coefficient_y)

		// tighten binary range of P: leading bit is 2^bp
		bin_expon_product -= 2 * BINARY_EXPONENT_BIAS
		bp = __tight_bin_range_128(P, bin_expon_product)

		// get number of decimal digits in the product
		digits_p = bid_estimate_decimal_digits[bp]
		if !__unsigned_compare_gt_128(bid_power10_table_128[digits_p], P) {
			digits_p++
		}

		// determine number of decimal digits to be rounded out
		extra_digits = digits_p - MAX_FORMAT_DIGITS
		final_exponent = exponent_x + exponent_y + extra_digits - DECIMAL_EXPONENT_BIAS
	}

	if uint(final_exponent) >= 3*256 {
		if final_exponent < 0 {
			// --- get number of bits in the coefficients of z ---
			tempx = math.Float64bits(float64(coefficient_z))
			bin_expon_cx = int((tempx&MASK_BINARY_EXPONENT)>>52) - 0x3ff
			digits_z = bid_estimate_decimal_digits[bin_expon_cx]
			if coefficient_z >= bid_power10_table_128[digits_z].w[0] {
				digits_z++
			}
			// underflow
			if (final_exponent+16 < 0) || (exponent_z+digits_z > 33+final_exponent) {
				res = BID_normalize(sign_z, exponent_z, coefficient_z,
					sign_x^sign_y, 1, rndMode, &pfpsf)
				return res, pfpsf
			}

			ez = exponent_z + digits_z - 16
			if ez < 0 {
				ez = 0
			}
			scale_z = exponent_z - ez
			coefficient_z *= bid_power10_table_128[scale_z].w[0]
			ey = final_exponent - extra_digits
			extra_digits = ez - ey

			if extra_digits > 17 {
				CYh = __truncate(P, 16)
				// get remainder
				T = bid_power10_table_128[16].w[0]
				CY0L = __mul_64x64_to_64(CYh, T)
				remainder_y = P.w[0] - CY0L

				extra_digits -= 16
				P.w[0] = CYh
				P.w[1] = 0
			} else {
				remainder_y = 0
			}

			// align coeff_x, CYh
			CZ = __mul_64x64_to_128(coefficient_z,
				bid_power10_table_128[extra_digits].w[0])

			if sign_z == (sign_y ^ sign_x) {
				CT = __add_128_128(CZ, P)
				if __unsigned_compare_ge_128(CT, bid_power10_table_128[16+extra_digits]) {
					extra_digits++
					ez++
				}
			} else {
				if remainder_y != 0 && (__unsigned_compare_ge_128(CZ, P)) {
					P.w[0]++
					if P.w[0] == 0 {
						P.w[1]++
					}
				}
				CT = __sub_128_128(CZ, P)
				if int64(CT.w[1]) < 0 {
					sign_z = sign_y ^ sign_x
					CT.w[0] = 0 - CT.w[0]
					CT.w[1] = 0 - CT.w[1]
					if CT.w[0] != 0 {
						CT.w[1]--
					}
				} else if (CT.w[1] | CT.w[0]) == 0 {
					if rndMode != BID_ROUNDING_DOWN {
						sign_z = 0
					} else {
						sign_z = 0x8000000000000000
					}
				}
				if ez != 0 && __unsigned_compare_gt_128(bid_power10_table_128[15+extra_digits], CT) {
					extra_digits--
					ez--
				}
			}

			uf_status = 0
			if (ez == 0) && __unsigned_compare_gt_128(bid_power10_table_128[extra_digits+15], CT) {
				uf_status = BID_UNDERFLOW_EXCEPTION
			}
			res = __bid_full_round64_remainder(sign_z, ez-extra_digits, CT,
				extra_digits, remainder_y, rndMode, &pfpsf, uf_status)
			return res, pfpsf

		} else {
			if (sign_z == (sign_x ^ sign_y)) || (final_exponent > 3*256+15) {
				res, flags := fast_get_BID64_check_OF_flags(sign_x^sign_y, final_exponent,
					1000000000000000, rndMode)
				pfpsf |= flags
				return res, pfpsf
			}
		}
	}

	if extra_digits > 0 {
		res = bid_get_add128(sign_z, exponent_z, coefficient_z, sign_x^sign_y,
			final_exponent, P, extra_digits, rndMode, &pfpsf)
		return res, pfpsf
	}
	// go to convert_format and exit
	C64 = __low_64(P)
	res = bid_get_add64(sign_x^sign_y,
		exponent_x+exponent_y-DECIMAL_EXPONENT_BIAS, C64,
		sign_z, exponent_z, coefficient_z,
		rndMode, &pfpsf)
	return res, pfpsf
}
