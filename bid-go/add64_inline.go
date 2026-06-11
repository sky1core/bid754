package bidgo

import "math"

// bid_get_add64 is ported mechanically from bid_inline_add.h.
func bid_get_add64(sign_x uint64, exponent_x int, coefficient_x uint64,
	sign_y uint64, exponent_y int, coefficient_y uint64,
	rounding_mode int, fpsc *uint32) uint64 {
	var CA, CT, CT_new BID_UINT128
	var sign_a, sign_b, coefficient_a, coefficient_b, sign_s, sign_ab, rem_a uint64
	var saved_ca, saved_cb, C0_64, C64, remainder_h, T1, carry, tmp, C64_new uint64
	var exponent_a, exponent_b, diff_dec_expon int
	var bin_expon_ca, extra_digits, amount, scale_k, scale_ca int
	var rmode int
	var status uint32

	// sort arguments by exponent
	if exponent_x <= exponent_y {
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

	// --- get number of bits in the coefficients of x and y ---
	tempx := math.Float64bits(float64(coefficient_a))
	bin_expon_ca = int((tempx&MASK_BINARY_EXPONENT)>>52) - 0x3ff

	if coefficient_a == 0 {
		return get_BID64_withFlags(sign_b, exponent_b, coefficient_b, rounding_mode, fpsc)
	}
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

		// --- get number of bits in the coefficients of x and y ---
		tempx = math.Float64bits(float64(coefficient_a))
		bin_expon_ca = int((tempx&MASK_BINARY_EXPONENT)>>52) - 0x3ff

		if diff_dec_expon > MAX_FORMAT_DIGITS {
			if coefficient_b != 0 {
				*fpsc |= BID_INEXACT_EXCEPTION
			}

			if (rounding_mode&3) != 0 && coefficient_b != 0 {
				switch rounding_mode {
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
			} else if (coefficient_a == 1000000000000000) &&
				diff_dec_expon == MAX_FORMAT_DIGITS+1 &&
				(sign_a^sign_b) != 0 &&
				(coefficient_b > 5000000000000000) {
				coefficient_a = 9999999999999999
				exponent_a--
			}

			return get_BID64_withFlags(sign_a, exponent_a, coefficient_a, rounding_mode, fpsc)
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
			if rounding_mode == BID_ROUNDING_DOWN && coefficient_a == 0 && sign_a != sign_b {
				sign_s = 0x8000000000000000
			}
			return get_BID64_withFlags(sign_s, exponent_b, coefficient_a, rounding_mode, fpsc)
		}
		// otherwise rounding is necessary

		// already know coefficient_a<10^19
		if coefficient_a < bid_power10_table_128[17].w[0] {
			extra_digits = 1
		} else if coefficient_a < bid_power10_table_128[18].w[0] {
			extra_digits = 2
		} else {
			extra_digits = 3
		}

		rmode = rounding_mode
		if sign_s != 0 && uint(rmode-1) < 2 {
			rmode = 3 - rmode
		}
		coefficient_a += bid_round_const_table[rmode][extra_digits]

		// get P*(2^M[extra_digits])/10^extra_digits
		CT = __mul_64x64_to_128(coefficient_a, bid_reciprocals10_64[extra_digits])

		// now get P/10^extra_digits: shift C64 right by M[extra_digits]-128
		amount = bid_short_recip_scale[extra_digits]
		C64 = CT.w[1] >> uint(amount)

	} else {
		// coefficient_a*10^(exponent_a-exponent_b) is large
		sign_s = sign_a

		rmode = rounding_mode
		if sign_s != 0 && uint(rmode-1) < 2 {
			rmode = 3 - rmode
		}

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
		C64 = C0_64 + coefficient_a
		// filter out difficult (corner) cases
		if uint64(C64-1000000000000000-1) > 9000000000000000-2 {
			if C64 >= 10000000000000000 {
				// result has more than 16 digits
				if scale_k == 0 {
					// must divide coeff_a by 10
					saved_ca = saved_ca + T1
					CA = __mul_64x64_to_128(saved_ca, 0x3333333333333334)
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
				C64 = C0_64 + coefficient_a
			} else if C64 <= 1000000000000000 {
				// less than 16 digits in result
				coefficient_a = uint64(int64(saved_ca) * int64(bid_power10_table_128[scale_k+1].w[0]))
				exponent_b--
				coefficient_b = (saved_cb << 3) + (saved_cb << 1) + 100000000000000000 + bid_round_const_table[rmode][extra_digits]

				// get P*(2^M[extra_digits])/10^extra_digits
				CT_new = __mul_64x64_to_128(coefficient_b, bid_reciprocals10_64[extra_digits])

				// now get P/10^extra_digits: shift C64 right by M[extra_digits]-128
				amount = bid_short_recip_scale[extra_digits]
				C0_64 = CT_new.w[1] >> uint(amount)

				// result coefficient
				C64_new = C0_64 + coefficient_a
				if C64_new < 10000000000000000 {
					C64 = C64_new
					CT = CT_new
				} else {
					exponent_b++
				}
			}
		}
	}

	if rmode == 0 {
		if (C64 & 1) != 0 {
			// get remainder
			remainder_h = CT.w[1] << uint(64-amount)

			// test whether fractional part is 0
			if remainder_h == 0 && (CT.w[0] < bid_reciprocals10_64[extra_digits]) {
				C64--
			}
		}
	}

	status = BID_INEXACT_EXCEPTION

	// get remainder
	remainder_h = CT.w[1] << uint(64-amount)

	switch rmode {
	case BID_ROUNDING_TO_NEAREST, BID_ROUNDING_TIES_AWAY:
		if (remainder_h == 0x8000000000000000) &&
			(CT.w[0] < bid_reciprocals10_64[extra_digits]) {
			status = BID_EXACT_STATUS
		}
	case BID_ROUNDING_DOWN, BID_ROUNDING_TO_ZERO:
		if remainder_h == 0 && (CT.w[0] < bid_reciprocals10_64[extra_digits]) {
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
	*fpsc |= status

	return get_BID64_withFlags(sign_s, exponent_b+extra_digits, C64, rounding_mode, fpsc)
}
