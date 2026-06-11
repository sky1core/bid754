package bidgo

import "math"

// bid_get_add128 is ported mechanically from bid_inline_add.h.
func bid_get_add128(sign_x uint64, exponent_x int, coefficient_x uint64,
	sign_y uint64, final_exponent_y int, CY BID_UINT128,
	extra_digits int, rounding_mode int, fpsc *uint32) uint64 {
	var CY_L, CX, FS, F, CT, ST, T2 BID_UINT128
	var CYh, CY0L, T, S, coefficient_y, remainder_y uint64
	var D int64
	var diff_dec_expon, extra_digits2, exponent_y int
	var extra_dx, diff_dec2, bin_expon_cx, digits_x, rmode int
	var status uint32

	// CY has more than 16 decimal digits
	_ = status
	exponent_y = final_exponent_y - extra_digits

	if exponent_x > exponent_y {
		// normalize x
		tempx := math.Float64bits(float64(coefficient_x))
		bin_expon_cx = int((tempx&MASK_BINARY_EXPONENT)>>52) - 0x3ff
		digits_x = bid_estimate_decimal_digits[bin_expon_cx]
		if coefficient_x >= bid_power10_table_128[digits_x].w[0] {
			digits_x++
		}

		extra_dx = 16 - digits_x
		coefficient_x *= bid_power10_table_128[extra_dx].w[0]
		if (sign_x^sign_y) != 0 && (coefficient_x == 1000000000000000) {
			extra_dx++
			coefficient_x = 10000000000000000
		}
		exponent_x -= extra_dx

		if exponent_x > exponent_y {
			// exponent_x > exponent_y
			diff_dec_expon = exponent_x - exponent_y

			if exponent_x <= final_exponent_y+1 {
				CX = __mul_64x64_to_128(coefficient_x,
					bid_power10_table_128[diff_dec_expon].w[0])

				if sign_x == sign_y {
					CT = __add_128_128(CY, CX)
					if exponent_x > final_exponent_y {
						extra_digits++
					}
					if __unsigned_compare_ge_128(CT, bid_power10_table_128[16+extra_digits]) {
						extra_digits++
					}
				} else {
					CT = __sub_128_128(CY, CX)
					if int64(CT.w[1]) < 0 {
						CT.w[0] = 0 - CT.w[0]
						CT.w[1] = 0 - CT.w[1]
						if CT.w[0] != 0 {
							CT.w[1]--
						}
						sign_y = sign_x
					} else if (CT.w[1] | CT.w[0]) == 0 {
						if rounding_mode != BID_ROUNDING_DOWN {
							sign_y = 0
						} else {
							sign_y = 0x8000000000000000
						}
					}
					if exponent_x+1 >= final_exponent_y {
						extra_digits = __get_dec_digits64(CT) - 16
						if extra_digits <= 0 {
							if CT.w[0] == 0 && rounding_mode == BID_ROUNDING_DOWN {
								sign_y = 0x8000000000000000
							}
							return get_BID64_withFlags(sign_y, exponent_y, CT.w[0],
								rounding_mode, fpsc)
						}
					} else if __unsigned_compare_gt_128(
						bid_power10_table_128[15+extra_digits], CT) {
						extra_digits--
					}
				}

				return __bid_full_round64(sign_y, exponent_y, CT, extra_digits,
					rounding_mode, fpsc)
			}
			// diff_dec2+extra_digits is the number of digits to eliminate from
			// argument CY
			diff_dec2 = exponent_x - final_exponent_y

			if diff_dec2 >= 17 {
				if (rounding_mode & 3) != 0 {
					switch rounding_mode {
					case BID_ROUNDING_UP:
						if sign_y == 0 {
							D = int64(sign_x ^ sign_y)
							D >>= 63
							D = D + D + 1
							coefficient_x += uint64(D)
						}
					case BID_ROUNDING_DOWN:
						if sign_y != 0 {
							D = int64(sign_x ^ sign_y)
							D >>= 63
							D = D + D + 1
							coefficient_x += uint64(D)
						}
					case BID_ROUNDING_TO_ZERO:
						if sign_y != sign_x {
							D = 0 - 1
							coefficient_x += uint64(D)
						}
					default:
					}
					if coefficient_x < 1000000000000000 {
						coefficient_x -= uint64(D)
						coefficient_x = uint64(D) + (coefficient_x << 1) + (coefficient_x << 3)
						exponent_x--
					}
				}
				if (CY.w[1] | CY.w[0]) != 0 {
					*fpsc |= BID_INEXACT_EXCEPTION
				}
				return get_BID64_withFlags(sign_x, exponent_x, coefficient_x,
					rounding_mode, fpsc)
			}
			// here exponent_x <= 16+final_exponent_y

			// truncate CY to 16 dec. digits
			CYh = __truncate(CY, extra_digits)

			// get remainder
			T = bid_power10_table_128[extra_digits].w[0]
			CY0L = __mul_64x64_to_64(CYh, T)

			remainder_y = CY.w[0] - CY0L

			// align coeff_x, CYh
			CX = __mul_64x64_to_128(coefficient_x,
				bid_power10_table_128[diff_dec2].w[0])

			if sign_x == sign_y {
				CT = __add_128_64(CX, CYh)
				if __unsigned_compare_ge_128(CT, bid_power10_table_128[16+diff_dec2]) {
					diff_dec2++
				}
			} else {
				if remainder_y != 0 {
					CYh++
				}
				CT = __sub_128_64(CX, CYh)
				if __unsigned_compare_gt_128(bid_power10_table_128[15+diff_dec2], CT) {
					diff_dec2--
				}
			}

			return __bid_full_round64_remainder(sign_x, final_exponent_y, CT,
				diff_dec2, remainder_y, rounding_mode, fpsc, 0)
		}
	}
	// Here (exponent_x <= exponent_y)
	{
		diff_dec_expon = exponent_y - exponent_x

		if diff_dec_expon > MAX_FORMAT_DIGITS {
			rmode = rounding_mode

			if (sign_x ^ sign_y) != 0 {
				if CY.w[0] == 0 {
					CY.w[1]--
				}
				CY.w[0]--
				if __unsigned_compare_gt_128(bid_power10_table_128[15+extra_digits], CY) {
					if (rmode & 3) != 0 {
						extra_digits--
						final_exponent_y--
					} else {
						CY.w[0] = 1000000000000000
						CY.w[1] = 0
						extra_digits = 0
					}
				}
			}
			CY = __scale128_10(CY)
			extra_digits++
			CY.w[0] |= 1

			return __bid_simple_round64_sticky(sign_y, final_exponent_y, CY,
				extra_digits, rmode, fpsc)
		}
		// apply sign to coeff_x
		sign_x ^= sign_y
		sign_x = uint64(int64(sign_x) >> 63)
		CX.w[0] = (coefficient_x + sign_x) ^ sign_x
		CX.w[1] = sign_x

		// check whether CY (rounded to 16 digits) and CX have
		// any digits in the same position
		diff_dec2 = final_exponent_y - exponent_x

		if diff_dec2 <= 17 {
			// align CY to 10^ex
			S = bid_power10_table_128[diff_dec_expon].w[0]
			CY_L = __mul_64x128_short(S, CY)

			ST = __add_128_128(CY_L, CX)
			extra_digits2 = __get_dec_digits64(ST) - 16
			return __bid_full_round64(sign_y, exponent_x, ST, extra_digits2,
				rounding_mode, fpsc)
		}
		// truncate CY to 16 dec. digits
		CYh = __truncate(CY, extra_digits)

		// get remainder
		T = bid_power10_table_128[extra_digits].w[0]
		CY0L = __mul_64x64_to_64(CYh, T)

		coefficient_y = CY.w[0] - CY0L
		// add rounding constant
		rmode = rounding_mode
		if sign_y != 0 && uint(rmode-1) < 2 {
			rmode = 3 - rmode
		}
		if (rmode & 3) == 0 {
			coefficient_y += bid_round_const_table[rmode][extra_digits]
		}
		// align coefficient_y, coefficient_x
		S = bid_power10_table_128[diff_dec_expon].w[0]
		F = __mul_64x64_to_128(coefficient_y, S)

		// fraction
		FS = __add_128_128(F, CX)

		if rmode == 0 {
			// rounding code, here RN_EVEN
			// 10^(extra_digits+diff_dec_expon)
			T2 = bid_power10_table_128[diff_dec_expon+extra_digits]
			if __unsigned_compare_gt_128(FS, T2) ||
				((CYh&1) != 0 && __test_equal_128(FS, T2)) {
				CYh++
				FS = __sub_128_128(FS, T2)
			}
		}
		if rmode == 4 {
			// rounding code, here RN_AWAY
			T2 = bid_power10_table_128[diff_dec_expon+extra_digits]
			if __unsigned_compare_ge_128(FS, T2) {
				CYh++
				FS = __sub_128_128(FS, T2)
			}
		}
		switch rmode {
		case BID_ROUNDING_DOWN, BID_ROUNDING_TO_ZERO:
			if int64(FS.w[1]) < 0 {
				CYh--
				if CYh < 1000000000000000 {
					CYh = 9999999999999999
					final_exponent_y--
				}
			} else {
				T2 = bid_power10_table_128[diff_dec_expon+extra_digits]
				if __unsigned_compare_ge_128(FS, T2) {
					CYh++
					FS = __sub_128_128(FS, T2)
				}
			}
		case BID_ROUNDING_UP:
			if int64(FS.w[1]) < 0 {
				break
			}
			T2 = bid_power10_table_128[diff_dec_expon+extra_digits]
			if __unsigned_compare_gt_128(FS, T2) {
				CYh += 2
				FS = __sub_128_128(FS, T2)
			} else if (FS.w[1] == T2.w[1]) && (FS.w[0] == T2.w[0]) {
				CYh++
				FS.w[1] = 0
				FS.w[0] = 0
			} else if (FS.w[1] | FS.w[0]) != 0 {
				CYh++
			}
		default:
		}

		status = BID_INEXACT_EXCEPTION
		if (rmode & 3) == 0 {
			// RN modes
			if (FS.w[1] == bid_round_const_table_128[0][diff_dec_expon+extra_digits].w[1]) &&
				(FS.w[0] == bid_round_const_table_128[0][diff_dec_expon+extra_digits].w[0]) {
				status = BID_EXACT_STATUS
			}
		} else if FS.w[1] == 0 && FS.w[0] == 0 {
			status = BID_EXACT_STATUS
		}

		*fpsc |= status

		return get_BID64_withFlags(sign_y, final_exponent_y, CYh, rounding_mode, fpsc)
	}
}
