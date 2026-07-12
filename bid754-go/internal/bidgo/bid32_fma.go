// Ported from: Intel bid32_fma.c
// Mechanical translation - all logic preserved exactly.

package bidgo

import "math"

// add_zero32 is ported from bid32_fma.c inline helper.
func add_zero32(exponent_y int, sign_z uint32, exponent_z int,
	coefficient_z uint32, prounding_mode *int, fpsc *uint32) uint32 {
	var bin_expon, scale_k, scale_cz int
	var diff_expon int

	diff_expon = exponent_z - exponent_y

	tempx := math.Float64bits(float64(coefficient_z))
	bin_expon = int(((tempx & MASK_BINARY_EXPONENT) >> 52)) - 0x3ff
	scale_cz = bid_estimate_decimal_digits[bin_expon]
	if uint64(coefficient_z) >= bid_power10_table_128[scale_cz].w[0] {
		scale_cz++
	}

	scale_k = 7 - scale_cz
	if diff_expon < scale_k {
		scale_k = diff_expon
	}
	coefficient_z *= uint32(bid_power10_table_128[scale_k].w[0])

	return get_BID32(sign_z, exponent_z-scale_k, uint64(coefficient_z), *prounding_mode)
}

// Bid32Fma is ported mechanically from bid32_fma.c: bid32_fma.
func Bid32Fma(x, y, z uint32, rnd_mode int) (uint32, uint32) {
	var P, Tmp, CB, Q_high, Q_low, Stemp, C128 BID_UINT128
	var P0, C64, remainder_h, rem_l, carry, CY, coefficient_a, coefficient_b, sign_ab uint64
	var sign_x, sign_y, coefficient_x, coefficient_y, sign_z, coefficient_z, R uint32
	var sign_a, sign_b, res uint32
	var extra_digits, exponent_x, exponent_y, exponent_z, bin_expon, rmode int
	var inexact int
	var n_digits, amount, exponent_a, exponent_b, diff_dec_expon, d2, scale_ca int
	var status, pfpsf uint32

	sign_x, exponent_x, coefficient_x, valid_x := unpack_BID32(x)
	sign_y, exponent_y, coefficient_y, valid_y := unpack_BID32(y)
	sign_z, exponent_z, coefficient_z, valid_z := unpack_BID32(z)
	if coefficient_x == 0 {
		valid_x = false
	}
	if coefficient_y == 0 {
		valid_y = false
	}
	if coefficient_z == 0 {
		valid_z = false
	}

	if !valid_x || !valid_y || !valid_z {
		if (y & NAN_MASK32) == NAN_MASK32 {
			if ((x & SNAN_MASK32) == SNAN_MASK32) ||
				((y & SNAN_MASK32) == SNAN_MASK32) || ((z & SNAN_MASK32) == SNAN_MASK32) {
				pfpsf |= BID_INVALID_EXCEPTION
			}
			res = coefficient_y & QUIET_MASK32
			return res, pfpsf
		}
		if (z & NAN_MASK32) == NAN_MASK32 {
			if ((x & SNAN_MASK32) == SNAN_MASK32) ||
				((z & SNAN_MASK32) == SNAN_MASK32) {
				pfpsf |= BID_INVALID_EXCEPTION
			}
			res = coefficient_z & QUIET_MASK32
			return res, pfpsf
		}
		if (x & NAN_MASK32) == NAN_MASK32 {
			if (x & SNAN_MASK32) == SNAN_MASK32 {
				pfpsf |= BID_INVALID_EXCEPTION
			}
			res = coefficient_x & QUIET_MASK32
			return res, pfpsf
		}

		if !valid_x {
			if (x & 0x78000000) == 0x78000000 {
				if coefficient_y == 0 {
					if (z & 0x7e000000) != 0x7c000000 {
						pfpsf |= BID_INVALID_EXCEPTION
					}
					return 0x7c000000, pfpsf
				}
				if ((z & 0x7c000000) == 0x78000000) && (((x^y)^z)&0x80000000) != 0 {
					pfpsf |= BID_INVALID_EXCEPTION
					return 0x7c000000, pfpsf
				}
				return ((x ^ y) & 0x80000000) | 0x78000000, pfpsf
			}
			if ((y & 0x78000000) != 0x78000000) && ((z & 0x78000000) != 0x78000000) {
				if coefficient_z != 0 {
					exponent_y = exponent_x - DECIMAL_EXPONENT_BIAS_32 + exponent_y
					sign_z = z & 0x80000000
					if exponent_y >= exponent_z {
						return z, pfpsf
					}
					res = add_zero32(exponent_y, sign_z, exponent_z, coefficient_z, &rnd_mode, &pfpsf)
					return res, pfpsf
				}
			}
		}
		if !valid_y {
			if (y & 0x78000000) == 0x78000000 {
				if coefficient_x == 0 {
					pfpsf |= BID_INVALID_EXCEPTION
					return 0x7c000000, pfpsf
				}
				if ((z & 0x7c000000) == 0x78000000) && (((x^y)^z)&0x80000000) != 0 {
					pfpsf |= BID_INVALID_EXCEPTION
					return 0x7c000000, pfpsf
				}
				return ((x ^ y) & 0x80000000) | 0x78000000, pfpsf
			}
			if (z & 0x78000000) != 0x78000000 {
				if coefficient_z != 0 {
					exponent_y += exponent_x - DECIMAL_EXPONENT_BIAS_32
					sign_z = z & 0x80000000
					if exponent_y >= exponent_z {
						return z, pfpsf
					}
					res = add_zero32(exponent_y, sign_z, exponent_z, coefficient_z, &rnd_mode, &pfpsf)
					return res, pfpsf
				}
			}
		}

		if !valid_z {
			if (z & 0x78000000) == 0x78000000 {
				res = coefficient_z & QUIET_MASK32
				return res, pfpsf
			}
			if coefficient_x == 0 || coefficient_y == 0 {
				exponent_x += exponent_y - DECIMAL_EXPONENT_BIAS_32
				if exponent_x > DECIMAL_MAX_EXPON_32 {
					exponent_x = DECIMAL_MAX_EXPON_32
				} else if exponent_x < 0 {
					exponent_x = 0
				}
				if exponent_x <= exponent_z {
					res = uint32(exponent_x) << 23
				} else {
					res = uint32(exponent_z) << 23
				}
				if (sign_x ^ sign_y) == sign_z {
					res |= sign_z
				} else if rnd_mode == BID_ROUNDING_DOWN {
					res |= 0x80000000
				}
				return res, pfpsf
			}
			d2 = exponent_x + exponent_y - DECIMAL_EXPONENT_BIAS_32
			if exponent_z > d2 {
				exponent_z = d2
			}
		}
	}

	P0 = uint64(coefficient_x) * uint64(coefficient_y)
	exponent_x += exponent_y - DECIMAL_EXPONENT_BIAS_32

	if exponent_x < exponent_z {
		sign_a = sign_z
		exponent_a = exponent_z
		coefficient_a = uint64(coefficient_z)
		sign_b = sign_x ^ sign_y
		exponent_b = exponent_x
		coefficient_b = P0
	} else {
		sign_a = sign_x ^ sign_y
		exponent_a = exponent_x
		coefficient_a = P0
		sign_b = sign_z
		exponent_b = exponent_z
		coefficient_b = uint64(coefficient_z)
	}

	diff_dec_expon = exponent_a - exponent_b

	if diff_dec_expon > 17 {
		tempx := math.Float64bits(float64(coefficient_a))
		bin_expon = int(((tempx & MASK_BINARY_EXPONENT) >> 52)) - 0x3ff
		scale_ca = bid_estimate_decimal_digits[bin_expon]

		d2 = 31 - scale_ca
		if diff_dec_expon > d2 {
			diff_dec_expon = d2
			exponent_b = exponent_a - diff_dec_expon
		}
		if coefficient_b != 0 {
			inexact = 1
		}
	}

	sign_ab = uint64(int64(int32(sign_a^sign_b)) << 32)
	sign_ab = uint64(int64(sign_ab) >> 63)
	CB.w[0] = (coefficient_b + sign_ab) ^ sign_ab
	CB.w[1] = uint64(int64(CB.w[0]) >> 63)

	_, Tmp = __mul_64x128_full(coefficient_a, bid_power10_table_128[diff_dec_expon])
	P = __add_128_128(Tmp, CB)
	if int64(P.w[1]) < 0 {
		sign_a ^= 0x80000000
		P.w[1] = 0 - P.w[1]
		if P.w[0] != 0 {
			P.w[1]--
		}
		P.w[0] = 0 - P.w[0]
	}

	if P.w[1] != 0 {
		tempx := math.Float64bits(float64(P.w[1]))
		bin_expon = int(((tempx & MASK_BINARY_EXPONENT) >> 52)) - 0x3ff + 64
		n_digits = bid_estimate_decimal_digits[bin_expon]
		if __unsigned_compare_ge_128(P, bid_power10_table_128[n_digits]) {
			n_digits++
		}
	} else {
		if P.w[0] != 0 {
			tempx := math.Float64bits(float64(P.w[0]))
			bin_expon = int(((tempx & MASK_BINARY_EXPONENT) >> 52)) - 0x3ff
			n_digits = bid_estimate_decimal_digits[bin_expon]
			if P.w[0] >= bid_power10_table_128[n_digits].w[0] {
				n_digits++
			}
		} else {
			sign_a = 0
			if rnd_mode == BID_ROUNDING_DOWN {
				sign_a = 0x80000000
			}
			if coefficient_a == 0 {
				sign_a = sign_x
			}
			n_digits = 0
		}
	}

	if n_digits <= MAX_FORMAT_DIGITS_32 {
		res = get_BID32_UF(sign_a, exponent_b, uint64(uint32(P.w[0])), 0, rnd_mode, &pfpsf)
		return res, pfpsf
	}

	extra_digits = n_digits - 7

	rmode = rnd_mode
	if sign_a != 0 && uint(rmode-1) < 2 {
		rmode = 3 - rmode
	}

	if exponent_b+extra_digits < 0 {
		rmode = 3 // RZ
	}

	if extra_digits <= 18 {
		P = __add_128_64(P, bid_round_const_table[rmode][extra_digits])
	} else {
		Stemp = __mul_64x64_to_128(bid_round_const_table[rmode][18], bid_power10_table_128[extra_digits-18].w[0])
		P = __add_128_128(P, Stemp)
		if rmode == BID_ROUNDING_UP {
			P = __add_128_64(P, bid_round_const_table[rmode][extra_digits-18])
		}
	}

	Q_high, Q_low = __mul_128x128_full(P, bid_reciprocals10_128[extra_digits])
	amount = bid_recip_scale[extra_digits]
	C128 = __shr_128_long(Q_high, uint(amount))

	C64 = __low_64(C128)

	if rmode == 0 {
		if (C64 & 1) != 0 {
			rem_l = Q_high.w[0]
			if amount < 64 {
				remainder_h = Q_high.w[0] << uint(64-amount)
				rem_l = 0
			} else {
				remainder_h = Q_high.w[1] << uint(128-amount)
			}

			if (remainder_h|rem_l) == 0 &&
				(Q_low.w[1] < bid_reciprocals10_128[extra_digits].w[1] ||
					(Q_low.w[1] == bid_reciprocals10_128[extra_digits].w[1] &&
						Q_low.w[0] < bid_reciprocals10_128[extra_digits].w[0])) {
				C64--
			}
		}
	}

	status = BID_INEXACT_EXCEPTION

	rem_l = Q_high.w[0]
	if amount < 64 {
		remainder_h = Q_high.w[0] << uint(64-amount)
		rem_l = 0
	} else {
		remainder_h = Q_high.w[1] << uint(128-amount)
	}

	switch rmode {
	case BID_ROUNDING_TO_NEAREST, BID_ROUNDING_TIES_AWAY:
		if (remainder_h == 0x8000000000000000 && rem_l == 0) &&
			(Q_low.w[1] < bid_reciprocals10_128[extra_digits].w[1] ||
				(Q_low.w[1] == bid_reciprocals10_128[extra_digits].w[1] &&
					Q_low.w[0] < bid_reciprocals10_128[extra_digits].w[0])) {
			status = BID_EXACT_STATUS
		}
	case BID_ROUNDING_DOWN, BID_ROUNDING_TO_ZERO:
		if (remainder_h|rem_l) == 0 &&
			(Q_low.w[1] < bid_reciprocals10_128[extra_digits].w[1] ||
				(Q_low.w[1] == bid_reciprocals10_128[extra_digits].w[1] &&
					Q_low.w[0] < bid_reciprocals10_128[extra_digits].w[0])) {
			status = BID_EXACT_STATUS
		}
	default:
		Stemp.w[0], CY = __add_carry_out(Q_low.w[0], bid_reciprocals10_128[extra_digits].w[0])
		Stemp.w[1], carry = __add_carry_in_out(Q_low.w[1], bid_reciprocals10_128[extra_digits].w[1], CY)
		_ = Stemp
		if amount < 64 {
			if (remainder_h>>uint(64-amount))+carry >= (uint64(1) << uint(amount)) {
				if inexact == 0 {
					status = BID_EXACT_STATUS
				}
			}
		} else {
			rem_l += carry
			remainder_h >>= uint(128 - amount)
			if carry != 0 && rem_l == 0 {
				remainder_h++
			}
			if remainder_h >= (uint64(1)<<uint(amount-64)) && inexact == 0 {
				status = BID_EXACT_STATUS
			}
		}
	}

	pfpsf |= status

	R = 0
	if status != BID_EXACT_STATUS {
		R = 1
	}

	// DECIMAL_TINY_DETECTION_AFTER_ROUNDING is 0 in our build,
	// so this block is skipped (matching bid64_fma behavior)

	res = get_BID32_UF(sign_a, exponent_b+extra_digits, uint64(uint32(C64)), uint32(R), rnd_mode, &pfpsf)
	return res, pfpsf
}
