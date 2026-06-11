package bidgo

import "math"

func __low_64(q BID_UINT128) uint64 {
	return q.w[0]
}

func __mul_64x64_to_64(cx, cy uint64) uint64 {
	return cx * cy
}

func __mul_64x128_short(a uint64, b BID_UINT128) BID_UINT128 {
	var ql BID_UINT128
	var ALBH_L uint64

	ALBH_L = __mul_64x64_to_64(a, b.w[1])
	ql = __mul_64x64_to_128(a, b.w[0])
	ql.w[1] += ALBH_L
	return ql
}

func __scale128_10(tmp BID_UINT128) BID_UINT128 {
	var tmp2, tmp8 BID_UINT128

	tmp2.w[1] = (tmp.w[1] << 1) | (tmp.w[0] >> 63)
	tmp2.w[0] = tmp.w[0] << 1
	tmp8.w[1] = (tmp.w[1] << 3) | (tmp.w[0] >> 61)
	tmp8.w[0] = tmp.w[0] << 3
	return __add_128_128(tmp2, tmp8)
}

// __bid_simple_round64_sticky is ported mechanically from bid_inline_add.h.
func __bid_simple_round64_sticky(sign uint64, exponent int, P BID_UINT128,
	extra_digits int, rounding_mode int, fpsc *uint32) uint64 {
	var Q_high, Q_low, C128 BID_UINT128
	var C64 uint64
	var amount, rmode int

	rmode = rounding_mode
	if sign != 0 && uint(rmode-1) < 2 {
		rmode = 3 - rmode
	}
	P = __add_128_64(P, bid_round_const_table[rmode][extra_digits])

	Q_high, Q_low = __mul_128x128_full(P, bid_reciprocals10_128[extra_digits])
	_ = Q_low
	amount = bid_recip_scale[extra_digits]
	C128 = __shr_128(Q_high, uint(amount))
	C64 = __low_64(C128)

	*fpsc |= BID_INEXACT_EXCEPTION

	return get_BID64_withFlags(sign, exponent, C64, rounding_mode, fpsc)
}

// __bid_full_round64 is ported mechanically from bid_inline_add.h.
func __bid_full_round64(sign uint64, exponent int, P BID_UINT128,
	extra_digits int, rounding_mode int, fpsc *uint32) uint64 {
	var Q_high, Q_low, C128, Stemp BID_UINT128
	var remainder_h, C64, carry, CY uint64
	var amount, amount2, rmode int
	var status uint32

	if exponent < 0 {
		if exponent >= -16 && (extra_digits+exponent < 0) {
			extra_digits = -exponent
			status = BID_UNDERFLOW_EXCEPTION
		}
	}

	if extra_digits > 0 {
		exponent += extra_digits
		rmode = rounding_mode
		if sign != 0 && uint(rmode-1) < 2 {
			rmode = 3 - rmode
		}
		P = __add_128_128(P, bid_round_const_table_128[rmode][extra_digits])

		Q_high, Q_low = __mul_128x128_full(P, bid_reciprocals10_128[extra_digits])
		amount = bid_recip_scale[extra_digits]
		C128 = __shr_128_long(Q_high, uint(amount))
		C64 = __low_64(C128)

		if rmode == 0 {
			if (C64 & 1) != 0 {
				amount2 = 64 - amount
				remainder_h = 0
				remainder_h--
				remainder_h >>= uint(amount2)
				remainder_h = remainder_h & Q_high.w[0]

				if remainder_h == 0 &&
					(Q_low.w[1] < bid_reciprocals10_128[extra_digits].w[1] ||
						(Q_low.w[1] == bid_reciprocals10_128[extra_digits].w[1] &&
							Q_low.w[0] < bid_reciprocals10_128[extra_digits].w[0])) {
					C64--
				}
			}
		}

		status |= BID_INEXACT_EXCEPTION
		remainder_h = Q_high.w[0] << uint(64-amount)

		switch rmode {
		case BID_ROUNDING_TO_NEAREST, BID_ROUNDING_TIES_AWAY:
			if remainder_h == 0x8000000000000000 &&
				(Q_low.w[1] < bid_reciprocals10_128[extra_digits].w[1] ||
					(Q_low.w[1] == bid_reciprocals10_128[extra_digits].w[1] &&
						Q_low.w[0] < bid_reciprocals10_128[extra_digits].w[0])) {
				status = BID_EXACT_STATUS
			}
		case BID_ROUNDING_DOWN, BID_ROUNDING_TO_ZERO:
			if remainder_h == 0 &&
				(Q_low.w[1] < bid_reciprocals10_128[extra_digits].w[1] ||
					(Q_low.w[1] == bid_reciprocals10_128[extra_digits].w[1] &&
						Q_low.w[0] < bid_reciprocals10_128[extra_digits].w[0])) {
				status = BID_EXACT_STATUS
			}
		default:
			Stemp.w[0], CY = __add_carry_out(Q_low.w[0], bid_reciprocals10_128[extra_digits].w[0])
			Stemp.w[1], carry = __add_carry_in_out(Q_low.w[1], bid_reciprocals10_128[extra_digits].w[1], CY)
			_ = Stemp
			if (remainder_h>>uint(64-amount))+carry >= (uint64(1) << uint(amount)) {
				status = BID_EXACT_STATUS
			}
		}

		*fpsc |= status
	} else {
		C64 = P.w[0]
		if C64 == 0 {
			sign = 0
			if rounding_mode == BID_ROUNDING_DOWN {
				sign = 0x8000000000000000
			}
		}
	}
	return get_BID64_withFlags(sign, exponent, C64, rounding_mode, fpsc)
}

// __bid_full_round64_remainder is ported mechanically from bid_inline_add.h.
func __bid_full_round64_remainder(sign uint64, exponent int, P BID_UINT128,
	extra_digits int, remainder_P uint64, rounding_mode int, fpsc *uint32,
	uf_status uint32) uint64 {
	var Q_high, Q_low, C128, Stemp BID_UINT128
	var remainder_h, C64, carry, CY uint64
	var amount, amount2, rmode int
	status := uf_status

	rmode = rounding_mode
	if sign != 0 && uint(rmode-1) < 2 {
		rmode = 3 - rmode
	}
	if rmode == BID_ROUNDING_UP && remainder_P != 0 {
		P.w[0]++
		if P.w[0] == 0 {
			P.w[1]++
		}
	}

	if extra_digits != 0 {
		P = __add_128_64(P, bid_round_const_table[rmode][extra_digits])

		Q_high, Q_low = __mul_128x128_full(P, bid_reciprocals10_128[extra_digits])
		amount = bid_recip_scale[extra_digits]
		C128 = __shr_128(Q_high, uint(amount))
		C64 = __low_64(C128)

		if rmode == 0 {
			if remainder_P == 0 && (C64&1) != 0 {
				amount2 = 64 - amount
				remainder_h = 0
				remainder_h--
				remainder_h >>= uint(amount2)
				remainder_h = remainder_h & Q_high.w[0]

				if remainder_h == 0 &&
					(Q_low.w[1] < bid_reciprocals10_128[extra_digits].w[1] ||
						(Q_low.w[1] == bid_reciprocals10_128[extra_digits].w[1] &&
							Q_low.w[0] < bid_reciprocals10_128[extra_digits].w[0])) {
					C64--
				}
			}
		}

		status |= BID_INEXACT_EXCEPTION

		if remainder_P == 0 {
			remainder_h = Q_high.w[0] << uint(64-amount)

			switch rmode {
			case BID_ROUNDING_TO_NEAREST, BID_ROUNDING_TIES_AWAY:
				if remainder_h == 0x8000000000000000 &&
					(Q_low.w[1] < bid_reciprocals10_128[extra_digits].w[1] ||
						(Q_low.w[1] == bid_reciprocals10_128[extra_digits].w[1] &&
							Q_low.w[0] < bid_reciprocals10_128[extra_digits].w[0])) {
					status = BID_EXACT_STATUS
				}
			case BID_ROUNDING_DOWN, BID_ROUNDING_TO_ZERO:
				if remainder_h == 0 &&
					(Q_low.w[1] < bid_reciprocals10_128[extra_digits].w[1] ||
						(Q_low.w[1] == bid_reciprocals10_128[extra_digits].w[1] &&
							Q_low.w[0] < bid_reciprocals10_128[extra_digits].w[0])) {
					status = BID_EXACT_STATUS
				}
			default:
				Stemp.w[0], CY = __add_carry_out(Q_low.w[0], bid_reciprocals10_128[extra_digits].w[0])
				Stemp.w[1], carry = __add_carry_in_out(Q_low.w[1], bid_reciprocals10_128[extra_digits].w[1], CY)
				_ = Stemp
				if (remainder_h>>uint(64-amount))+carry >= (uint64(1) << uint(amount)) {
					status = BID_EXACT_STATUS
				}
			}
		}
		*fpsc |= status
	} else {
		C64 = P.w[0]
		if remainder_P != 0 {
			*fpsc |= uf_status | BID_INEXACT_EXCEPTION
		}
	}

	return get_BID64_withFlags(sign, exponent+extra_digits, C64, rounding_mode, fpsc)
}

// __truncate is ported mechanically from bid_inline_add.h.
func __truncate(P BID_UINT128, extra_digits int) uint64 {
	var Q_high, Q_low, C128 BID_UINT128
	var C64 uint64
	var amount int

	Q_high, Q_low = __mul_128x128_full(P, bid_reciprocals10_128[extra_digits])
	_ = Q_low
	amount = bid_recip_scale[extra_digits]
	C128 = __shr_128(Q_high, uint(amount))
	C64 = __low_64(C128)
	return C64
}

// __get_dec_digits64 is ported mechanically from bid_inline_add.h.
func __get_dec_digits64(X BID_UINT128) int {
	var digits_x, bin_expon_cx int

	if X.w[1] == 0 {
		if X.w[0] == 0 {
			return 0
		}
		tempx := math.Float64bits(float64(X.w[0]))
		bin_expon_cx = int((tempx&MASK_BINARY_EXPONENT)>>52) - 0x3ff
		digits_x = bid_estimate_decimal_digits[bin_expon_cx]
		if X.w[0] >= bid_power10_table_128[digits_x].w[0] {
			digits_x++
		}
		return digits_x
	}
	tempx := math.Float64bits(float64(X.w[1]))
	bin_expon_cx = int((tempx&MASK_BINARY_EXPONENT)>>52) - 0x3ff
	digits_x = bid_estimate_decimal_digits[bin_expon_cx+64]
	if __unsigned_compare_ge_128(X, bid_power10_table_128[digits_x]) {
		digits_x++
	}

	return digits_x
}

// BID_normalize is ported mechanically from bid_inline_add.h.
func BID_normalize(sign_z uint64, exponent_z int, coefficient_z uint64,
	round_dir uint64, round_flag int, rounding_mode int, fpsc *uint32) uint64 {
	var D int64
	var digits_z, bin_expon, scale, rmode int

	rmode = rounding_mode
	if sign_z != 0 && uint(rmode-1) < 2 {
		rmode = 3 - rmode
	}

	tempx := math.Float64bits(float64(coefficient_z))
	bin_expon = int((tempx&MASK_BINARY_EXPONENT)>>52) - 0x3ff
	digits_z = bid_estimate_decimal_digits[bin_expon]
	if coefficient_z >= bid_power10_table_128[digits_z].w[0] {
		digits_z++
	}

	scale = 16 - digits_z
	exponent_z -= scale
	if exponent_z < 0 {
		scale += exponent_z
		exponent_z = 0
	}
	coefficient_z *= bid_power10_table_128[scale].w[0]

	if round_flag != 0 {
		*fpsc |= BID_INEXACT_EXCEPTION
		if coefficient_z < 1000000000000000 {
			*fpsc |= BID_UNDERFLOW_EXCEPTION
		} else if coefficient_z == 1000000000000000 && exponent_z == 0 &&
			(int64(round_dir^sign_z) < 0) && round_flag != 0 {
			*fpsc |= BID_UNDERFLOW_EXCEPTION
		}
	}

	if round_flag != 0 && (rmode&3) != 0 {
		D = int64(round_dir ^ sign_z)

		if rmode == BID_ROUNDING_UP {
			if D >= 0 {
				coefficient_z++
			}
		} else {
			if D < 0 {
				coefficient_z--
			}
			if coefficient_z < 1000000000000000 && exponent_z != 0 {
				coefficient_z = 9999999999999999
				exponent_z--
			}
		}
	}

	return get_BID64_withFlags(sign_z, exponent_z, coefficient_z, rounding_mode, fpsc)
}

// add_zero64 is ported mechanically from bid_inline_add.h.
func add_zero64(exponent_y int, sign_z uint64, exponent_z int,
	coefficient_z uint64, prounding_mode *int, fpsc *uint32) uint64 {
	var bin_expon, scale_k, scale_cz int
	var diff_expon int

	diff_expon = exponent_z - exponent_y

	tempx := math.Float64bits(float64(coefficient_z))
	bin_expon = int((tempx&MASK_BINARY_EXPONENT)>>52) - 0x3ff
	scale_cz = bid_estimate_decimal_digits[bin_expon]
	if coefficient_z >= bid_power10_table_128[scale_cz].w[0] {
		scale_cz++
	}

	scale_k = 16 - scale_cz
	if diff_expon < scale_k {
		scale_k = diff_expon
	}
	coefficient_z *= bid_power10_table_128[scale_k].w[0]

	return get_BID64_withFlags(sign_z, exponent_z-scale_k, coefficient_z, *prounding_mode, fpsc)
}
