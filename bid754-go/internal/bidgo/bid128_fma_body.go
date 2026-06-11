// Ported from: Intel bid128_fma.c lines 1653-3637
// Mechanical translation - all logic preserved exactly.
// This file contains the main body of bid128_ext_fma (delta >= 0 and delta < 0 cases)
// and the bid_add_and_round helper function.

package bidgo

// bid_fma_main_body handles the f*f+f case of bid128_ext_fma.
// It processes both delta >= 0 and delta < 0 cases.
func bid_fma_main_body(
	p34 int,
	res *BID_UINT128,
	ptr_is_midpoint_lt_even, ptr_is_midpoint_gt_even,
	ptr_is_inexact_lt_midpoint, ptr_is_inexact_gt_midpoint *int,
	p_sign, z_sign uint64,
	z_exp_ptr, p_exp_ptr *uint64,
	q3, q4 int,
	e3_ptr, e4_ptr *int,
	delta int,
	C3 *BID_UINT128, C4 BID_UINT256,
	rnd_mode int, pfpsf *uint32) {

	z_exp := *z_exp_ptr
	p_exp := *p_exp_ptr
	e3 := *e3_ptr
	e4 := *e4_ptr

	var is_midpoint_lt_even, is_midpoint_gt_even int
	var is_inexact_lt_midpoint, is_inexact_gt_midpoint int

	if delta >= 0 {
		bid_fma_delta_ge_zero(p34, res,
			&is_midpoint_lt_even, &is_midpoint_gt_even,
			&is_inexact_lt_midpoint, &is_inexact_gt_midpoint,
			p_sign, z_sign, &z_exp, &p_exp,
			q3, q4, &e3, &e4, delta,
			C3, C4, rnd_mode, pfpsf)
	} else { // delta < 0
		bid_fma_delta_lt_zero(p34, res,
			&is_midpoint_lt_even, &is_midpoint_gt_even,
			&is_inexact_lt_midpoint, &is_inexact_gt_midpoint,
			p_sign, z_sign, &z_exp, &p_exp,
			q3, q4, &e3, &e4, delta,
			C3, C4, rnd_mode, pfpsf)
	}

	*ptr_is_midpoint_lt_even = is_midpoint_lt_even
	*ptr_is_midpoint_gt_even = is_midpoint_gt_even
	*ptr_is_inexact_lt_midpoint = is_inexact_lt_midpoint
	*ptr_is_inexact_gt_midpoint = is_inexact_gt_midpoint
	*z_exp_ptr = z_exp
	*p_exp_ptr = p_exp
	*e3_ptr = e3
	*e4_ptr = e4
}

// bid_fma_delta_ge_zero handles the delta >= 0 case.
// This corresponds to C lines 1655-2965.
func bid_fma_delta_ge_zero(
	p34 int, res *BID_UINT128,
	ptr_is_midpoint_lt_even, ptr_is_midpoint_gt_even,
	ptr_is_inexact_lt_midpoint, ptr_is_inexact_gt_midpoint *int,
	p_sign, z_sign uint64,
	z_exp_ptr, p_exp_ptr *uint64,
	q3, q4 int,
	e3_ptr, e4_ptr *int,
	delta int,
	C3 *BID_UINT128, C4 BID_UINT256,
	rnd_mode int, pfpsf *uint32) {

	z_exp := *z_exp_ptr
	p_exp := *p_exp_ptr
	e3 := *e3_ptr
	e4 := *e4_ptr

	var is_midpoint_lt_even, is_midpoint_gt_even int
	var is_inexact_lt_midpoint, is_inexact_gt_midpoint int
	var is_midpoint_lt_even0, is_midpoint_gt_even0 int
	var is_inexact_lt_midpoint0, is_inexact_gt_midpoint0 int
	var incr_exp int
	var lt_half_ulp, eq_half_ulp, gt_half_ulp int
	var is_tiny int
	var R64 uint64
	var P128, R128 BID_UINT128
	var P192, R192 BID_UINT192
	var R256 BID_UINT256
	var scale, ind, x0 int
	var tmp_sign uint64
	_, _, _, _, _, _ = is_midpoint_lt_even0, is_midpoint_gt_even0, is_inexact_lt_midpoint0, is_inexact_gt_midpoint0, is_tiny, x0
	_, _, _ = R128, R192, R256
	_, _, _, _, _ = R64, P128, P192, ind, tmp_sign

	if p34 <= delta-1 || // Case (1')
		(p34 == delta && e3+6176 < p34-q3) { // Case (1''A)

		// check for overflow
		if (q3+e3) > (p34+expmax) && p34 <= delta-1 {
			if rnd_mode == BID_ROUNDING_TO_NEAREST {
				res.w[1] = z_sign | 0x7800000000000000
				res.w[0] = 0x0000000000000000
				*pfpsf |= (BID_INEXACT_EXCEPTION | BID_OVERFLOW_EXCEPTION)
			} else {
				if p_sign == z_sign {
					is_inexact_lt_midpoint = 1
				} else {
					is_inexact_gt_midpoint = 1
				}
				scale = p34 - q3
				if scale == 0 {
					res.w[1] = z_sign | C3.w[1]
					res.w[0] = C3.w[0]
				} else {
					if q3 <= 19 {
						if scale <= 19 {
							*res = __mul_64x64_to_128(C3.w[0], bid_ten2k64[scale])
						} else {
							*res = __mul_64x128_to_128(C3.w[0], bid_ten2k128[scale-20])
						}
					} else {
						*res = __mul_64x128_to_128(bid_ten2k64[scale], *C3)
					}
				}
				e3 = e3 - scale
				res.w[1] = z_sign | res.w[1]
				bid_rounding_correction(rnd_mode,
					is_inexact_lt_midpoint,
					is_inexact_gt_midpoint,
					is_midpoint_lt_even, is_midpoint_gt_even,
					e3, res, pfpsf)
			}
			*ptr_is_midpoint_lt_even = is_midpoint_lt_even
			*ptr_is_midpoint_gt_even = is_midpoint_gt_even
			*ptr_is_inexact_lt_midpoint = is_inexact_lt_midpoint
			*ptr_is_inexact_gt_midpoint = is_inexact_gt_midpoint
			goto done
		}

		// res = z
		if q3 < p34 {
			scale = p34 - q3
			ind = e3 + 6176
			if ind < scale {
				scale = ind
			}
			if scale == 0 {
				res.w[1] = C3.w[1]
				res.w[0] = C3.w[0]
			} else if q3 <= 19 {
				if scale <= 19 {
					*res = __mul_64x64_to_128(C3.w[0], bid_ten2k64[scale])
				} else {
					*res = __mul_64x128_to_128(C3.w[0], bid_ten2k128[scale-20])
				}
			} else {
				*res = __mul_64x128_to_128(bid_ten2k64[scale], *C3)
			}
			z_exp = z_exp - (uint64(scale) << 49)
			e3 = e3 - scale
			res.w[1] = z_sign | (z_exp & MASK_EXP_128) | res.w[1]
			if scale+q3 < p34 {
				*pfpsf |= BID_UNDERFLOW_EXCEPTION
			}
		} else { // q3 = p34
			scale = 0
			res.w[1] = z_sign | (uint64(e3+6176) << 49) | C3.w[1]
			res.w[0] = C3.w[0]
		}

		// avoid double rounding errors
		if (p_sign != z_sign) && (delta == (q3 + scale + 1)) {
			if (q3 <= 19 && C3.w[0] != bid_ten2k64[q3-1]) ||
				(q3 == 20 && (C3.w[1] != 0 || C3.w[0] != bid_ten2k64[19])) ||
				(q3 >= 21 && (C3.w[1] != bid_ten2k128[q3-21].w[1] ||
					C3.w[0] != bid_ten2k128[q3-21].w[0])) {
				is_inexact_gt_midpoint = 1
			} else { // C3 * 10^scale = 10^(q3+scale-1)
				if q4 == 1 {
					R64 = C4.w[0]
				} else {
					if q4 <= 18 {
						R64 = bid_round64_2_18(q4, q4-1, C4.w[0], &incr_exp,
							&is_midpoint_lt_even, &is_midpoint_gt_even,
							&is_inexact_lt_midpoint,
							&is_inexact_gt_midpoint)
					} else if q4 <= 38 {
						P128.w[1] = C4.w[1]
						P128.w[0] = C4.w[0]
						R128, incr_exp, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint = bid_round128_19_38(q4, q4-1, P128)
						R64 = R128.w[0]
					} else if q4 <= 57 {
						P192.w[2] = C4.w[2]
						P192.w[1] = C4.w[1]
						P192.w[0] = C4.w[0]
						R192, incr_exp, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint = bid_round192_39_57(q4, q4-1, P192)
						R64 = R192.w[0]
					} else {
						R256, incr_exp, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint = bid_round256_58_76(q4, q4-1, C4)
						R64 = R256.w[0]
					}
					if incr_exp != 0 {
						R64 = 10
					}
				}
				if R64 == 5 && is_inexact_lt_midpoint == 0 && is_inexact_gt_midpoint == 0 &&
					is_midpoint_lt_even == 0 && is_midpoint_gt_even == 0 {
					is_inexact_lt_midpoint = 0
					is_inexact_gt_midpoint = 0
					is_midpoint_lt_even = 1
					is_midpoint_gt_even = 0
				} else if (e3 == expmin128) ||
					R64 < 5 || (R64 == 5 && is_inexact_gt_midpoint != 0) {
					is_inexact_lt_midpoint = 0
					is_inexact_gt_midpoint = 1
					is_midpoint_lt_even = 0
					is_midpoint_gt_even = 0
				} else {
					is_inexact_lt_midpoint = 1
					is_inexact_gt_midpoint = 0
					is_midpoint_lt_even = 0
					is_midpoint_gt_even = 0
					if (q3 + scale) <= 19 {
						res.w[1] = 0
						res.w[0] = bid_ten2k64[q3+scale]
					} else {
						res.w[1] = bid_ten2k128[q3+scale-20].w[1]
						res.w[0] = bid_ten2k128[q3+scale-20].w[0]
					}
					res.w[0] = res.w[0] - 1
					z_exp = z_exp - EXP_P1_128
					e3 = e3 - 1
					res.w[1] = z_sign | (uint64(e3+6176) << 49) | res.w[1]
				}
				if e3 == expmin128 {
					// DECIMAL_TINY_DETECTION_AFTER_ROUNDING = 0
					*pfpsf |= BID_UNDERFLOW_EXCEPTION
				}
			}
			*pfpsf |= BID_INEXACT_EXCEPTION
		} else {
			if p_sign == z_sign {
				is_inexact_lt_midpoint = 1
			} else {
				is_inexact_gt_midpoint = 1
			}
			*pfpsf |= BID_INEXACT_EXCEPTION
		}

		// underflow detection (DECIMAL_TINY_DETECTION_AFTER_ROUNDING = 0)
		if (e3 == expmin128 && (q3+scale) < p34) ||
			(e3 == expmin128 && (q3+scale) == p34 &&
				(res.w[1]&MASK_COEFF128) == 0x0000314dc6448d93 &&
				res.w[0] == 0x38c15b0a00000000 &&
				z_sign != p_sign) {
			*pfpsf |= BID_UNDERFLOW_EXCEPTION
		}
		if rnd_mode != BID_ROUNDING_TO_NEAREST {
			bid_rounding_correction(rnd_mode,
				is_inexact_lt_midpoint,
				is_inexact_gt_midpoint,
				is_midpoint_lt_even, is_midpoint_gt_even,
				e3, res, pfpsf)
		}
		*ptr_is_midpoint_lt_even = is_midpoint_lt_even
		*ptr_is_midpoint_gt_even = is_midpoint_gt_even
		*ptr_is_inexact_lt_midpoint = is_inexact_lt_midpoint
		*ptr_is_inexact_gt_midpoint = is_inexact_gt_midpoint
		goto done

	} else if p34 == delta { // Case (1''B)

		// scale C3 to p34 digits
		scale = p34 - q3
		if scale == 0 {
			res.w[1] = C3.w[1]
			res.w[0] = C3.w[0]
		} else if q3 <= 19 {
			if scale <= 19 {
				*res = __mul_64x64_to_128(C3.w[0], bid_ten2k64[scale])
			} else {
				*res = __mul_64x128_to_128(C3.w[0], bid_ten2k128[scale-20])
			}
		} else {
			*res = __mul_64x128_to_128(bid_ten2k64[scale], *C3)
		}
		z_exp = z_exp - (uint64(scale) << 49)
		e3 = e3 - scale

		// determine whether x*y is less than, equal to, or greater than 1/2 ulp(z)
		lt_half_ulp = 0
		eq_half_ulp = 0
		gt_half_ulp = 0
		if q4 <= 19 {
			if C4.w[0] < bid_midpoint64[q4-1] {
				lt_half_ulp = 1
			} else if C4.w[0] == bid_midpoint64[q4-1] {
				eq_half_ulp = 1
			} else {
				gt_half_ulp = 1
			}
		} else if q4 <= 38 {
			if C4.w[2] == 0 && (C4.w[1] < bid_midpoint128[q4-20].w[1] ||
				(C4.w[1] == bid_midpoint128[q4-20].w[1] &&
					C4.w[0] < bid_midpoint128[q4-20].w[0])) {
				lt_half_ulp = 1
			} else if C4.w[2] == 0 && C4.w[1] == bid_midpoint128[q4-20].w[1] &&
				C4.w[0] == bid_midpoint128[q4-20].w[0] {
				eq_half_ulp = 1
			} else {
				gt_half_ulp = 1
			}
		} else if q4 <= 58 {
			if C4.w[3] == 0 && (C4.w[2] < bid_midpoint192[q4-39].w[2] ||
				(C4.w[2] == bid_midpoint192[q4-39].w[2] &&
					C4.w[1] < bid_midpoint192[q4-39].w[1]) ||
				(C4.w[2] == bid_midpoint192[q4-39].w[2] &&
					C4.w[1] == bid_midpoint192[q4-39].w[1] &&
					C4.w[0] < bid_midpoint192[q4-39].w[0])) {
				lt_half_ulp = 1
			} else if C4.w[3] == 0 && C4.w[2] == bid_midpoint192[q4-39].w[2] &&
				C4.w[1] == bid_midpoint192[q4-39].w[1] &&
				C4.w[0] == bid_midpoint192[q4-39].w[0] {
				eq_half_ulp = 1
			} else {
				gt_half_ulp = 1
			}
		} else {
			if C4.w[3] < bid_midpoint256[q4-59].w[3] ||
				(C4.w[3] == bid_midpoint256[q4-59].w[3] &&
					C4.w[2] < bid_midpoint256[q4-59].w[2]) ||
				(C4.w[3] == bid_midpoint256[q4-59].w[3] &&
					C4.w[2] == bid_midpoint256[q4-59].w[2] &&
					C4.w[1] < bid_midpoint256[q4-59].w[1]) ||
				(C4.w[3] == bid_midpoint256[q4-59].w[3] &&
					C4.w[2] == bid_midpoint256[q4-59].w[2] &&
					C4.w[1] == bid_midpoint256[q4-59].w[1] &&
					C4.w[0] < bid_midpoint256[q4-59].w[0]) {
				lt_half_ulp = 1
			} else if C4.w[3] == bid_midpoint256[q4-59].w[3] &&
				C4.w[2] == bid_midpoint256[q4-59].w[2] &&
				C4.w[1] == bid_midpoint256[q4-59].w[1] &&
				C4.w[0] == bid_midpoint256[q4-59].w[0] {
				eq_half_ulp = 1
			} else {
				gt_half_ulp = 1
			}
		}

		if p_sign == z_sign {
			if lt_half_ulp != 0 {
				res.w[1] = z_sign | (z_exp & MASK_EXP_128) | res.w[1]
				is_inexact_lt_midpoint = 1
			} else if (eq_half_ulp != 0 && (res.w[0]&0x01) != 0) || gt_half_ulp != 0 {
				res.w[0]++
				if res.w[0] == 0x0 {
					res.w[1]++
				}
				if (res.w[1]&MASK_COEFF128) == 0x0001ed09bead87c0 &&
					res.w[0] == 0x378d8e6400000000 {
					e3 = e3 + 1
					z_exp = (uint64(e3+6176) << 49) & MASK_EXP_128
					res.w[1] = 0x0000314dc6448d93
					res.w[0] = 0x38c15b0a00000000
				}
				res.w[1] = z_sign | (z_exp & MASK_EXP_128) | res.w[1]
				if eq_half_ulp != 0 {
					is_midpoint_lt_even = 1
				} else {
					is_inexact_gt_midpoint = 1
				}
			} else { // eq_half_ulp && !(res.w[0] & 0x01)
				res.w[1] = z_sign | (z_exp & MASK_EXP_128) | res.w[1]
				is_midpoint_gt_even = 1
			}
			*pfpsf |= BID_INEXACT_EXCEPTION
			if e3 > expmax && rnd_mode == BID_ROUNDING_TO_NEAREST {
				res.w[1] = z_sign | 0x7800000000000000
				res.w[0] = 0x0000000000000000
				*pfpsf |= (BID_INEXACT_EXCEPTION | BID_OVERFLOW_EXCEPTION)
				*ptr_is_midpoint_lt_even = is_midpoint_lt_even
				*ptr_is_midpoint_gt_even = is_midpoint_gt_even
				*ptr_is_inexact_lt_midpoint = is_inexact_lt_midpoint
				*ptr_is_inexact_gt_midpoint = is_inexact_gt_midpoint
				goto done
			}
			if rnd_mode != BID_ROUNDING_TO_NEAREST {
				bid_rounding_correction(rnd_mode,
					is_inexact_lt_midpoint,
					is_inexact_gt_midpoint,
					is_midpoint_lt_even, is_midpoint_gt_even,
					e3, res, pfpsf)
				z_exp = res.w[1] & MASK_EXP_128
			}
		} else { // p_sign != z_sign
			// This is the complex case with C3*10^scale = 10^33 as a special case.
			// For brevity, delegate to bid_fma_case1ppB_psign_ne_zsign
			bid_fma_case1ppB_psign_ne_zsign(p34, res,
				&is_midpoint_lt_even, &is_midpoint_gt_even,
				&is_inexact_lt_midpoint, &is_inexact_gt_midpoint,
				p_sign, z_sign, &z_exp,
				q3, q4, &e3, scale,
				C3, C4,
				lt_half_ulp, eq_half_ulp, gt_half_ulp,
				rnd_mode, pfpsf)
		}

		res.w[1] = z_sign | (z_exp & MASK_EXP_128) | (res.w[1] & MASK_COEFF128)
		*ptr_is_midpoint_lt_even = is_midpoint_lt_even
		*ptr_is_midpoint_gt_even = is_midpoint_gt_even
		*ptr_is_inexact_lt_midpoint = is_inexact_lt_midpoint
		*ptr_is_inexact_gt_midpoint = is_inexact_gt_midpoint
		goto done

	} else if ((q3 <= delta && delta < p34 && p34 < delta+q4) ||
		(q3 <= delta && delta+q4 <= p34) ||
		(delta < q3 && p34 < delta+q4) ||
		(delta < q3 && q3 <= delta+q4 && delta+q4 <= p34) ||
		(delta+q4 < q3)) &&
		!(delta <= 1 && p_sign != z_sign) { // Cases (2)-(6)

		// delegate to bid_fma_cases_2_to_6
		bid_fma_cases_2_to_6(p34, res,
			&is_midpoint_lt_even, &is_midpoint_gt_even,
			&is_inexact_lt_midpoint, &is_inexact_gt_midpoint,
			p_sign, z_sign, &z_exp,
			q3, q4, &e3, &e4, delta,
			C3, C4, rnd_mode, pfpsf)
		*ptr_is_midpoint_lt_even = is_midpoint_lt_even
		*ptr_is_midpoint_gt_even = is_midpoint_gt_even
		*ptr_is_inexact_lt_midpoint = is_inexact_lt_midpoint
		*ptr_is_inexact_gt_midpoint = is_inexact_gt_midpoint
		goto done

	} else {
		// delta <= 1 and signs are opposite; Cases (2)-(6) with massive cancellation
		if delta+q4 < q3 { // from Case (6)
			P128.w[1] = C3.w[1]
			P128.w[0] = C3.w[0]
			C3.w[1] = C4.w[1]
			C3.w[0] = C4.w[0]
			C4.w[1] = P128.w[1]
			C4.w[0] = P128.w[0]
			ind = q3
			q3 = q4
			q4 = ind
			ind = e3
			e3 = e4
			e4 = ind
			tmp_sign = z_sign
			z_sign = p_sign
			p_sign = tmp_sign
		} else { // from Cases (2), (3), (4), (5)
			delta = -delta
		}
		bid_add_and_round(q3, q4, e4, delta, p34, z_sign, p_sign, *C3, C4,
			rnd_mode, &is_midpoint_lt_even,
			&is_midpoint_gt_even, &is_inexact_lt_midpoint,
			&is_inexact_gt_midpoint, pfpsf, res)
		*ptr_is_midpoint_lt_even = is_midpoint_lt_even
		*ptr_is_midpoint_gt_even = is_midpoint_gt_even
		*ptr_is_inexact_lt_midpoint = is_inexact_lt_midpoint
		*ptr_is_inexact_gt_midpoint = is_inexact_gt_midpoint
		goto done
	}

done:
	*z_exp_ptr = z_exp
	*p_exp_ptr = p_exp
	*e3_ptr = e3
	*e4_ptr = e4
}

// bid_fma_delta_lt_zero handles delta < 0 cases (Cases 7-18).
// Ported from bid128_fma.c lines 2967-3630.
func bid_fma_delta_lt_zero(
	p34 int, res *BID_UINT128,
	ptr_is_midpoint_lt_even, ptr_is_midpoint_gt_even,
	ptr_is_inexact_lt_midpoint, ptr_is_inexact_gt_midpoint *int,
	p_sign, z_sign uint64,
	z_exp_ptr, p_exp_ptr *uint64,
	q3, q4 int,
	e3_ptr, e4_ptr *int,
	delta int,
	C3 *BID_UINT128, C4 BID_UINT256,
	rnd_mode int, pfpsf *uint32) {

	z_exp := *z_exp_ptr
	p_exp := *p_exp_ptr
	e3 := *e3_ptr
	e4 := *e4_ptr

	var is_midpoint_lt_even, is_midpoint_gt_even int
	var is_inexact_lt_midpoint, is_inexact_gt_midpoint int
	var is_midpoint_lt_even0, is_midpoint_gt_even0 int
	var is_inexact_lt_midpoint0, is_inexact_gt_midpoint0 int
	var incr_exp int
	var lsb int
	var lt_half_ulp, eq_half_ulp int
	var is_tiny int
	var R64 uint64
	var P128, R128 BID_UINT128
	var P192, R192 BID_UINT192
	var R256 BID_UINT256
	var scale, ind, x0 int
	var tmp_sign uint64

	delta = -delta

	if p34 < q4 && q4 <= delta { // Case (7)
		bid_fma_case7(p34, res,
			&is_midpoint_lt_even, &is_midpoint_gt_even,
			&is_inexact_lt_midpoint, &is_inexact_gt_midpoint,
			p_sign, z_sign, q3, q4, &e4, delta,
			C3, C4, rnd_mode, pfpsf)

	} else if (q4 <= p34 && p34 <= delta) || // Case (8)
		(q4 <= delta && delta < p34 && p34 < delta+q3) || // Case (9)
		(q4 <= delta && delta+q3 <= p34) || // Case (10)
		(delta < q4 && q4 <= p34 && p34 < delta+q3) || // Case (13)
		(delta < q4 && q4 <= delta+q3 && delta+q3 <= p34) || // Case (14)
		(delta+q3 < q4 && q4 <= p34) { // Case (18)

		// swap (C3, C4), (q3, q4), (e3, e4), (z_sign, p_sign), (z_exp, p_exp)
		// and go to delta_ge_zero
		P128.w[1] = C3.w[1]
		P128.w[0] = C3.w[0]
		C3.w[1] = C4.w[1]
		C3.w[0] = C4.w[0]
		C4.w[1] = P128.w[1]
		C4.w[0] = P128.w[0]
		ind = q3
		q3 = q4
		q4 = ind
		ind = e3
		e3 = e4
		e4 = ind
		tmp_sign = z_sign
		z_sign = p_sign
		p_sign = tmp_sign
		tmp64 := z_exp
		z_exp = p_exp
		p_exp = tmp64
		delta = q3 + e3 - q4 - e4
		bid_fma_delta_ge_zero(p34, res,
			&is_midpoint_lt_even, &is_midpoint_gt_even,
			&is_inexact_lt_midpoint, &is_inexact_gt_midpoint,
			p_sign, z_sign, &z_exp, &p_exp,
			q3, q4, &e3, &e4, delta,
			C3, C4, rnd_mode, pfpsf)

	} else if (p34 <= delta && delta < q4 && q4 < delta+q3) || // Case (11)
		(delta < p34 && p34 < q4 && q4 < delta+q3) { // Case (12)

		bid_fma_cases_11_12(p34, res,
			&is_midpoint_lt_even, &is_midpoint_gt_even,
			&is_inexact_lt_midpoint, &is_inexact_gt_midpoint,
			p_sign, z_sign,
			q3, q4, &e3, &e4, delta,
			C3, C4, rnd_mode, pfpsf)

	} else if (p34 <= delta && delta+q3 <= q4) || // Case (15)
		(delta < p34 && p34 < delta+q3 && delta+q3 <= q4) || // Case (16)
		(delta+q3 <= p34 && p34 < q4) { // Case (17)

		bid_add_and_round(q3, q4, e4, delta, p34, z_sign, p_sign, *C3, C4,
			rnd_mode, &is_midpoint_lt_even,
			&is_midpoint_gt_even, &is_inexact_lt_midpoint,
			&is_inexact_gt_midpoint, pfpsf, res)
	} else {
		// fallthrough
	}

	_ = R64
	_ = P128
	_ = R128
	_ = P192
	_ = R192
	_ = R256
	_ = scale
	_ = ind
	_ = x0
	_ = incr_exp
	_ = lsb
	_ = lt_half_ulp
	_ = eq_half_ulp
	_ = is_tiny
	_ = is_midpoint_lt_even0
	_ = is_midpoint_gt_even0
	_ = is_inexact_lt_midpoint0
	_ = is_inexact_gt_midpoint0
	_ = tmp_sign

	*ptr_is_midpoint_lt_even = is_midpoint_lt_even
	*ptr_is_midpoint_gt_even = is_midpoint_gt_even
	*ptr_is_inexact_lt_midpoint = is_inexact_lt_midpoint
	*ptr_is_inexact_gt_midpoint = is_inexact_gt_midpoint
	*z_exp_ptr = z_exp
	*p_exp_ptr = p_exp
	*e3_ptr = e3
	*e4_ptr = e4
}

// bid_fma_case1ppB_psign_ne_zsign handles Case (1”B) when p_sign != z_sign.
// Ported from the corresponding Case (1”B) branch in Intel bid128_fma.c.
func bid_fma_case1ppB_psign_ne_zsign(
	p34 int, res *BID_UINT128,
	ptr_is_midpoint_lt_even, ptr_is_midpoint_gt_even,
	ptr_is_inexact_lt_midpoint, ptr_is_inexact_gt_midpoint *int,
	p_sign, z_sign uint64, z_exp_ptr *uint64,
	q3, q4 int, e3_ptr *int, scale int,
	C3 *BID_UINT128, C4 BID_UINT256,
	lt_half_ulp, eq_half_ulp, gt_half_ulp int,
	rnd_mode int, pfpsf *uint32) {

	z_exp := *z_exp_ptr
	e3 := *e3_ptr
	var is_midpoint_lt_even, is_midpoint_gt_even int
	var is_inexact_lt_midpoint, is_inexact_gt_midpoint int
	var incr_exp int
	var R64 uint64
	var R128 BID_UINT128
	var P128 BID_UINT128
	var P192 BID_UINT192
	var R192 BID_UINT192
	var R256 BID_UINT256

	// consider two cases: C3*10^scale != 10^33, and C3*10^scale = 10^33
	if res.w[1] != 0x0000314dc6448d93 || res.w[0] != 0x38c15b0a00000000 { // != 10^33
		if lt_half_ulp != 0 {
			res.w[1] = z_sign | (z_exp & MASK_EXP_128) | res.w[1]
			is_inexact_gt_midpoint = 1
		} else if (eq_half_ulp != 0 && (res.w[0]&0x01) != 0) || gt_half_ulp != 0 {
			res.w[0]--
			if res.w[0] == 0xffffffffffffffff {
				res.w[1]--
			}
			res.w[1] = z_sign | (z_exp & MASK_EXP_128) | res.w[1]
			if eq_half_ulp != 0 {
				is_midpoint_gt_even = 1
			} else {
				is_inexact_lt_midpoint = 1
			}
		} else { // eq_half_ulp && !(res.w[0] & 0x01)
			res.w[1] = z_sign | (z_exp & MASK_EXP_128) | res.w[1]
			is_midpoint_lt_even = 1
		}
		if e3 > expmax {
			if rnd_mode == BID_ROUNDING_TO_NEAREST {
				res.w[1] = z_sign | 0x7800000000000000
				res.w[0] = 0x0000000000000000
				*pfpsf |= (BID_INEXACT_EXCEPTION | BID_OVERFLOW_EXCEPTION)
			} else {
				bid_rounding_correction(rnd_mode,
					is_inexact_lt_midpoint,
					is_inexact_gt_midpoint,
					is_midpoint_lt_even,
					is_midpoint_gt_even, e3, res, pfpsf)
			}
			*ptr_is_midpoint_lt_even = is_midpoint_lt_even
			*ptr_is_midpoint_gt_even = is_midpoint_gt_even
			*ptr_is_inexact_lt_midpoint = is_inexact_lt_midpoint
			*ptr_is_inexact_gt_midpoint = is_inexact_gt_midpoint
			*z_exp_ptr = z_exp
			*e3_ptr = e3
			return
		}
		*pfpsf |= BID_INEXACT_EXCEPTION
		if rnd_mode != BID_ROUNDING_TO_NEAREST {
			bid_rounding_correction(rnd_mode,
				is_inexact_lt_midpoint,
				is_inexact_gt_midpoint,
				is_midpoint_lt_even,
				is_midpoint_gt_even, e3, res, pfpsf)
		}
		z_exp = res.w[1] & MASK_EXP_128
	} else { // C3 * 10^scale = 10^33
		e3 = int(z_exp>>49) - 6176
		if e3 > expmin128 {
			if q4 == 1 {
				res.w[1] = 0x0001ed09bead87c0
				res.w[0] = 0x378d8e6400000000 - C4.w[0]
				z_exp = z_exp - EXP_P1_128
				e3 = e3 - 1
				res.w[1] = z_sign | (z_exp & MASK_EXP_128) | res.w[1]
			} else {
				if q4 <= 18 {
					R64 = bid_round64_2_18(q4, q4-1, C4.w[0], &incr_exp,
						&is_midpoint_lt_even,
						&is_midpoint_gt_even,
						&is_inexact_lt_midpoint,
						&is_inexact_gt_midpoint)
				} else if q4 <= 38 {
					P128.w[1] = C4.w[1]
					P128.w[0] = C4.w[0]
					R128, incr_exp, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint = bid_round128_19_38(q4, q4-1, P128)
					R64 = R128.w[0]
				} else if q4 <= 57 {
					P192.w[2] = C4.w[2]
					P192.w[1] = C4.w[1]
					P192.w[0] = C4.w[0]
					R192, incr_exp, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint = bid_round192_39_57(q4, q4-1, P192)
					R64 = R192.w[0]
				} else {
					R256, incr_exp, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint = bid_round256_58_76(q4, q4-1, C4)
					R64 = R256.w[0]
				}
				if is_midpoint_lt_even == 0 && is_midpoint_gt_even == 0 &&
					is_inexact_lt_midpoint == 0 && is_inexact_gt_midpoint == 0 {
					z_exp = z_exp - EXP_P1_128
					e3 = e3 - 1
					res.w[1] = z_sign | (z_exp & MASK_EXP_128) | 0x0001ed09bead87c0
					res.w[0] = 0x378d8e6400000000 - R64
				} else {
					if incr_exp != 0 {
						R64 = 10
					}
					res.w[1] = 0x0001ed09bead87c0
					res.w[0] = 0x378d8e6400000000 - R64
					z_exp = z_exp - EXP_P1_128
					e3 = e3 - 1
					if is_inexact_lt_midpoint != 0 {
						is_inexact_lt_midpoint = 0
						is_inexact_gt_midpoint = 1
					} else if is_inexact_gt_midpoint != 0 {
						is_inexact_gt_midpoint = 0
						is_inexact_lt_midpoint = 1
					} else if is_midpoint_lt_even != 0 {
						is_midpoint_lt_even = 0
						is_midpoint_gt_even = 1
					} else if is_midpoint_gt_even != 0 {
						is_midpoint_gt_even = 0
						is_midpoint_lt_even = 1
					}
					if e3 > expmax {
						if rnd_mode == BID_ROUNDING_TO_NEAREST {
							res.w[1] = z_sign | 0x7800000000000000
							res.w[0] = 0x0000000000000000
							*pfpsf |= (BID_INEXACT_EXCEPTION | BID_OVERFLOW_EXCEPTION)
						} else {
							bid_rounding_correction(rnd_mode,
								is_inexact_lt_midpoint,
								is_inexact_gt_midpoint,
								is_midpoint_lt_even,
								is_midpoint_gt_even, e3, res, pfpsf)
						}
						*ptr_is_midpoint_lt_even = is_midpoint_lt_even
						*ptr_is_midpoint_gt_even = is_midpoint_gt_even
						*ptr_is_inexact_lt_midpoint = is_inexact_lt_midpoint
						*ptr_is_inexact_gt_midpoint = is_inexact_gt_midpoint
						*z_exp_ptr = z_exp
						*e3_ptr = e3
						return
					}
					*pfpsf |= BID_INEXACT_EXCEPTION
					res.w[1] = z_sign | (uint64(e3+6176) << 49) | res.w[1]
					if rnd_mode != BID_ROUNDING_TO_NEAREST {
						bid_rounding_correction(rnd_mode,
							is_inexact_lt_midpoint,
							is_inexact_gt_midpoint,
							is_midpoint_lt_even,
							is_midpoint_gt_even, e3, res, pfpsf)
					}
					z_exp = res.w[1] & MASK_EXP_128
				}
			}
		} else { // e3 = emin
			if gt_half_ulp != 0 {
				res.w[1] = 0x0000314dc6448d93
				res.w[0] = 0x38c15b09ffffffff
			} else {
				res.w[1] = 0x0000314dc6448d93
				res.w[0] = 0x38c15b0a00000000
			}
			res.w[1] = z_sign | (z_exp & MASK_EXP_128) | res.w[1]
			*pfpsf |= BID_UNDERFLOW_EXCEPTION

			if eq_half_ulp != 0 {
				is_midpoint_lt_even = 1
			} else if lt_half_ulp != 0 {
				is_inexact_gt_midpoint = 1
			} else {
				is_inexact_lt_midpoint = 1
			}

			if rnd_mode != BID_ROUNDING_TO_NEAREST {
				bid_rounding_correction(rnd_mode,
					is_inexact_lt_midpoint,
					is_inexact_gt_midpoint,
					is_midpoint_lt_even,
					is_midpoint_gt_even, e3, res, pfpsf)
				z_exp = res.w[1] & MASK_EXP_128
			}
		}
		if is_inexact_lt_midpoint != 0 || is_inexact_gt_midpoint != 0 ||
			is_midpoint_lt_even != 0 || is_midpoint_gt_even != 0 {
			*pfpsf |= BID_INEXACT_EXCEPTION
		}
	}

	*ptr_is_midpoint_lt_even = is_midpoint_lt_even
	*ptr_is_midpoint_gt_even = is_midpoint_gt_even
	*ptr_is_inexact_lt_midpoint = is_inexact_lt_midpoint
	*ptr_is_inexact_gt_midpoint = is_inexact_gt_midpoint
	*z_exp_ptr = z_exp
	*e3_ptr = e3
}

// The following helpers cover the remaining complex FMA case groups from
// Intel bid128_fma.c; they are part of the mechanical port, not placeholders.

func bid_fma_cases_2_to_6(
	p34 int, res *BID_UINT128,
	ptr_is_midpoint_lt_even, ptr_is_midpoint_gt_even,
	ptr_is_inexact_lt_midpoint, ptr_is_inexact_gt_midpoint *int,
	p_sign, z_sign uint64, z_exp_ptr *uint64,
	q3, q4 int, e3_ptr, e4_ptr *int, delta int,
	C3 *BID_UINT128, C4 BID_UINT256,
	rnd_mode int, pfpsf *uint32) {
	// Ported from bid128_fma.c lines 2332-2915
	// Cases (2)-(6): the result has the sign of z

	e3 := *e3_ptr
	var scale, x0 int
	var is_midpoint_lt_even, is_midpoint_gt_even int
	var is_inexact_lt_midpoint, is_inexact_gt_midpoint int
	var is_midpoint_lt_even0, is_midpoint_gt_even0 int
	var is_inexact_lt_midpoint0, is_inexact_gt_midpoint0 int
	var incr_exp int
	var is_tiny int
	var R128 BID_UINT128
	var P128 BID_UINT128
	var R64 uint64
	var P192 BID_UINT192
	var R192 BID_UINT192
	var R256 BID_UINT256
	var lsb uint64
	var tmp64 uint64
	var ind int

	if (q3 <= delta && delta < p34 && p34 < delta+q4) || // Case (2)
		(delta < q3 && p34 < delta+q4) { // Case (4)
		// round first the sum x * y + z with unbounded exponent
		// scale C3 up by scale = p34 - q3, 1 <= scale <= p34-1,
		// 1 <= scale <= 33
		// calculate res = C3 * 10^scale
		scale = p34 - q3
		x0 = delta + q4 - p34
	} else if delta+q4 < q3 { // Case (6)
		// make Case (6) look like Case (3) or Case (5) with scale = 0
		// by scaling up C4 by 10^(q3 - delta - q4)
		scale = q3 - delta - q4 // 1 <= scale <= 33
		if q4 <= 19 {           // 1 <= scale <= 19; C4 fits in 64 bits
			if scale <= 19 { // 10^scale fits in 64 bits
				// 64 x 64 C4.w[0] * bid_ten2k64[scale]
				P128 = __mul_64x64_to_128(C4.w[0], bid_ten2k64[scale])
			} else { // 10^scale fits in 128 bits
				// 64 x 128 C4.w[0] * bid_ten2k128[scale - 20]
				P128 = __mul_128x64_to_128(C4.w[0], bid_ten2k128[scale-20])
			}
		} else { // C4 fits in 128 bits, but 10^scale must fit in 64 bits
			// 64 x 128 bid_ten2k64[scale] * C4
			var C4_128 BID_UINT128
			C4_128.w[0] = C4.w[0]
			C4_128.w[1] = C4.w[1]
			P128 = __mul_128x64_to_128(bid_ten2k64[scale], C4_128)
		}
		C4.w[0] = P128.w[0]
		C4.w[1] = P128.w[1]
		// e4 does not need adjustment, as it is not used from this point on
		scale = 0
		x0 = 0
		// now Case (6) looks like Case (3) or Case (5) with scale = 0
	} else { // if Case (3) or Case (5)
		// calculate first the sum x * y + z with unbounded exponent (exact)
		// scale C3 up by scale = delta + q4 - q3, 1 <= scale <= p34-1,
		// 1 <= scale <= 33
		// calculate res = C3 * 10^scale
		scale = delta + q4 - q3
		x0 = 0
		// Note: the comments which follow refer [mainly] to Case (2)]
	}

	// case2_repeat:
	for {
		if scale == 0 { // this could happen e.g. if we return to case2_repeat
			// or in Case (4)
			res.w[1] = C3.w[1]
			res.w[0] = C3.w[0]
		} else if q3 <= 19 { // 1 <= scale <= 19; z fits in 64 bits
			if scale <= 19 { // 10^scale fits in 64 bits
				// 64 x 64 C3.w[0] * bid_ten2k64[scale]
				*res = __mul_64x64_to_128(C3.w[0], bid_ten2k64[scale])
			} else { // 10^scale fits in 128 bits
				// 64 x 128 C3.w[0] * bid_ten2k128[scale - 20]
				*res = __mul_128x64_to_128(C3.w[0], bid_ten2k128[scale-20])
			}
		} else { // z fits in 128 bits, but 10^scale must fit in 64 bits
			// 64 x 128 bid_ten2k64[scale] * C3
			*res = __mul_128x64_to_128(bid_ten2k64[scale], *C3)
		}
		// e3 is already calculated
		e3 = e3 - scale
		// now res = C3 * 10^scale and e3 = e3 - scale

		// round C4 to nearest to q4 - x0 digits, where x0 = delta + q4 - p34
		if x0 == 0 { // this could happen only if we return to case2_repeat, or
			// for Case (3) or Case (6)
			R128.w[1] = C4.w[1]
			R128.w[0] = C4.w[0]
		} else if q4 <= 18 {
			// 2 <= q4 <= 18, max(1, q3+q4-p34) <= x0 <= q4 - 1, 1 <= x0 <= 17
			R64 = bid_round64_2_18(q4, x0, C4.w[0], &incr_exp,
				&is_midpoint_lt_even, &is_midpoint_gt_even,
				&is_inexact_lt_midpoint, &is_inexact_gt_midpoint)
			if incr_exp != 0 {
				// R64 = 10^(q4-x0), 1 <= q4 - x0 <= q4 - 1, 1 <= q4 - x0 <= 17
				R64 = bid_ten2k64[q4-x0]
			}
			R128.w[1] = 0
			R128.w[0] = R64
		} else if q4 <= 38 {
			// 19 <= q4 <= 38, max(1, q3+q4-p34) <= x0 <= q4 - 1, 1 <= x0 <= 37
			P128.w[1] = C4.w[1]
			P128.w[0] = C4.w[0]
			R128, incr_exp, is_midpoint_lt_even, is_midpoint_gt_even,
				is_inexact_lt_midpoint, is_inexact_gt_midpoint =
				bid_round128_19_38(q4, x0, P128)
			if incr_exp != 0 {
				// R128 = 10^(q4-x0), 1 <= q4 - x0 <= q4 - 1, 1 <= q4 - x0 <= 37
				if q4-x0 <= 19 { // 1 <= q4 - x0 <= 19
					R128.w[0] = bid_ten2k64[q4-x0]
					// R128.w[1] stays 0
				} else { // 20 <= q4 - x0 <= 37
					R128.w[0] = bid_ten2k128[q4-x0-20].w[0]
					R128.w[1] = bid_ten2k128[q4-x0-20].w[1]
				}
			}
		} else if q4 <= 57 {
			// 38 <= q4 <= 57, max(1, q3+q4-p34) <= x0 <= q4 - 1, 5 <= x0 <= 56
			P192.w[2] = C4.w[2]
			P192.w[1] = C4.w[1]
			P192.w[0] = C4.w[0]
			R192, incr_exp, is_midpoint_lt_even, is_midpoint_gt_even,
				is_inexact_lt_midpoint, is_inexact_gt_midpoint =
				bid_round192_39_57(q4, x0, P192)
			// R192.w[2] is always 0
			if incr_exp != 0 {
				// R192 = 10^(q4-x0), 1 <= q4 - x0 <= q4 - 5, 1 <= q4 - x0 <= 52
				if q4-x0 <= 19 { // 1 <= q4 - x0 <= 19
					R192.w[0] = bid_ten2k64[q4-x0]
					// R192.w[1] stays 0
					// R192.w[2] stays 0
				} else { // 20 <= q4 - x0 <= 33
					R192.w[0] = bid_ten2k128[q4-x0-20].w[0]
					R192.w[1] = bid_ten2k128[q4-x0-20].w[1]
					// R192.w[2] stays 0
				}
			}
			R128.w[1] = R192.w[1]
			R128.w[0] = R192.w[0]
		} else {
			// 58 <= q4 <= 68, max(1, q3+q4-p34) <= x0 <= q4 - 1, 25 <= x0 <= 67
			R256, incr_exp, is_midpoint_lt_even, is_midpoint_gt_even,
				is_inexact_lt_midpoint, is_inexact_gt_midpoint =
				bid_round256_58_76(q4, x0, C4)
			// R256.w[3] and R256.w[2] are always 0
			if incr_exp != 0 {
				// R256 = 10^(q4-x0), 1 <= q4 - x0 <= q4 - 25, 1 <= q4 - x0 <= 43
				if q4-x0 <= 19 { // 1 <= q4 - x0 <= 19
					R256.w[0] = bid_ten2k64[q4-x0]
					// R256.w[1] stays 0
					// R256.w[2] stays 0
					// R256.w[3] stays 0
				} else { // 20 <= q4 - x0 <= 33
					R256.w[0] = bid_ten2k128[q4-x0-20].w[0]
					R256.w[1] = bid_ten2k128[q4-x0-20].w[1]
					// R256.w[2] stays 0
					// R256.w[3] stays 0
				}
			}
			R128.w[1] = R256.w[1]
			R128.w[0] = R256.w[0]
		}
		// now add C3 * 10^scale in res and the signed top (q4-x0) digits of C4,
		// rounded to nearest, which were copied into R128
		if z_sign == p_sign {
			lsb = res.w[0] & 0x01 // lsb of C3 * 10^scale
			// the sum can result in [up to] p34 or p34 + 1 digits
			res.w[0] = res.w[0] + R128.w[0]
			res.w[1] = res.w[1] + R128.w[1]
			if res.w[0] < R128.w[0] {
				res.w[1]++ // carry
			}
			// if res > 10^34 - 1 need to increase x0 and decrease scale by 1
			if res.w[1] > 0x0001ed09bead87c0 ||
				(res.w[1] == 0x0001ed09bead87c0 &&
					res.w[0] > 0x378d8e63ffffffff) {
				// avoid double rounding error
				is_inexact_lt_midpoint0 = is_inexact_lt_midpoint
				is_inexact_gt_midpoint0 = is_inexact_gt_midpoint
				is_midpoint_lt_even0 = is_midpoint_lt_even
				is_midpoint_gt_even0 = is_midpoint_gt_even
				is_inexact_lt_midpoint = 0
				is_inexact_gt_midpoint = 0
				is_midpoint_lt_even = 0
				is_midpoint_gt_even = 0
				P128.w[1] = res.w[1]
				P128.w[0] = res.w[0]
				*res, incr_exp, is_midpoint_lt_even, is_midpoint_gt_even,
					is_inexact_lt_midpoint, is_inexact_gt_midpoint =
					bid_round128_19_38(35, 1, P128)
				// incr_exp is 0 with certainty in this case
				_ = incr_exp
				// avoid a double rounding error
				if (is_inexact_gt_midpoint0 != 0 || is_midpoint_lt_even0 != 0) &&
					is_midpoint_lt_even != 0 { // double rounding error upward
					// res = res - 1
					res.w[0]--
					if res.w[0] == 0xffffffffffffffff {
						res.w[1]--
					}
					is_midpoint_lt_even = 0
					is_inexact_lt_midpoint = 1
				} else if (is_inexact_lt_midpoint0 != 0 || is_midpoint_gt_even0 != 0) &&
					is_midpoint_gt_even != 0 { // double rounding error downward
					// res = res + 1
					res.w[0]++
					if res.w[0] == 0 {
						res.w[1]++
					}
					is_midpoint_gt_even = 0
					is_inexact_gt_midpoint = 1
				} else if is_midpoint_lt_even == 0 && is_midpoint_gt_even == 0 &&
					is_inexact_lt_midpoint == 0 &&
					is_inexact_gt_midpoint == 0 {
					// if this second rounding was exact the result may still be
					// inexact because of the first rounding
					if is_inexact_gt_midpoint0 != 0 || is_midpoint_lt_even0 != 0 {
						is_inexact_gt_midpoint = 1
					}
					if is_inexact_lt_midpoint0 != 0 || is_midpoint_gt_even0 != 0 {
						is_inexact_lt_midpoint = 1
					}
				} else if is_midpoint_gt_even != 0 &&
					(is_inexact_gt_midpoint0 != 0 ||
						is_midpoint_lt_even0 != 0) {
					// pulled up to a midpoint
					is_inexact_lt_midpoint = 1
					is_inexact_gt_midpoint = 0
					is_midpoint_lt_even = 0
					is_midpoint_gt_even = 0
				} else if is_midpoint_lt_even != 0 &&
					(is_inexact_lt_midpoint0 != 0 ||
						is_midpoint_gt_even0 != 0) {
					// pulled down to a midpoint
					is_inexact_lt_midpoint = 0
					is_inexact_gt_midpoint = 1
					is_midpoint_lt_even = 0
					is_midpoint_gt_even = 0
				} else {
					// ;
				}
				// adjust exponent
				e3 = e3 + 1
				if is_midpoint_lt_even == 0 && is_midpoint_gt_even == 0 &&
					is_inexact_lt_midpoint == 0 && is_inexact_gt_midpoint == 0 {
					if is_midpoint_lt_even0 != 0 || is_midpoint_gt_even0 != 0 ||
						is_inexact_lt_midpoint0 != 0 || is_inexact_gt_midpoint0 != 0 {
						is_inexact_lt_midpoint = 1
					}
				}
			} else {
				// this is the result rounded with unbounded exponent, unless a
				// correction is needed
				res.w[1] = res.w[1] & MASK_COEFF128
				if lsb == 1 {
					if is_midpoint_gt_even != 0 {
						// res = res + 1
						is_midpoint_gt_even = 0
						is_midpoint_lt_even = 1
						res.w[0]++
						if res.w[0] == 0x0 {
							res.w[1]++
						}
						// check for rounding overflow
						if res.w[1] == 0x0001ed09bead87c0 &&
							res.w[0] == 0x378d8e6400000000 {
							// res = 10^34 => rounding overflow
							res.w[1] = 0x0000314dc6448d93
							res.w[0] = 0x38c15b0a00000000 // 10^33
							e3++
						}
					} else if is_midpoint_lt_even != 0 {
						// res = res - 1
						is_midpoint_lt_even = 0
						is_midpoint_gt_even = 1
						res.w[0]--
						if res.w[0] == 0xffffffffffffffff {
							res.w[1]--
						}
						// if the result is pure zero, the sign depends on the rounding
						// mode (x*y and z had opposite signs)
						if res.w[1] == 0x0 && res.w[0] == 0x0 {
							if rnd_mode != BID_ROUNDING_DOWN {
								z_sign = 0x0000000000000000
							} else {
								z_sign = 0x8000000000000000
							}
							// the exponent is max (e3, expmin)
							res.w[1] = 0x0
							res.w[0] = 0x0
							*ptr_is_midpoint_lt_even = is_midpoint_lt_even
							*ptr_is_midpoint_gt_even = is_midpoint_gt_even
							*ptr_is_inexact_lt_midpoint = is_inexact_lt_midpoint
							*ptr_is_inexact_gt_midpoint = is_inexact_gt_midpoint
							// BID_SWAP128 (res); BID_RETURN (res)
							*e3_ptr = e3
							return
						}
					} else {
						// ;
					}
				}
			}
		} else { // if (z_sign != p_sign)
			lsb = res.w[0] & 0x01 // lsb of C3 * 10^scale; R128 contains rounded C4
			// used to swap rounding indicators if p_sign != z_sign
			// the sum can result in [up to] p34 or p34 - 1 digits
			tmp64 = res.w[0]
			res.w[0] = res.w[0] - R128.w[0]
			res.w[1] = res.w[1] - R128.w[1]
			if res.w[0] > tmp64 {
				res.w[1]-- // borrow
			}
			// if res < 10^33 and exp > expmin need to decrease x0 and
			// increase scale by 1
			if e3 > expmin128 && ((res.w[1] < 0x0000314dc6448d93 ||
				(res.w[1] == 0x0000314dc6448d93 &&
					res.w[0] < 0x38c15b0a00000000)) ||
				((is_inexact_lt_midpoint|is_midpoint_gt_even) != 0 &&
					res.w[1] == 0x0000314dc6448d93 &&
					res.w[0] == 0x38c15b0a00000000)) &&
				x0 >= 1 {
				x0 = x0 - 1
				// first restore e3, otherwise it will be too small
				e3 = e3 + scale
				scale = scale + 1
				is_inexact_lt_midpoint = 0
				is_inexact_gt_midpoint = 0
				is_midpoint_lt_even = 0
				is_midpoint_gt_even = 0
				incr_exp = 0
				continue // goto case2_repeat
			}
			// else this is the result rounded with unbounded exponent;
			// because the result has opposite sign to that of C4 which was
			// rounded, need to change the rounding indicators
			if is_inexact_lt_midpoint != 0 {
				is_inexact_lt_midpoint = 0
				is_inexact_gt_midpoint = 1
			} else if is_inexact_gt_midpoint != 0 {
				is_inexact_gt_midpoint = 0
				is_inexact_lt_midpoint = 1
			} else if lsb == 0 {
				if is_midpoint_lt_even != 0 {
					is_midpoint_lt_even = 0
					is_midpoint_gt_even = 1
				} else if is_midpoint_gt_even != 0 {
					is_midpoint_gt_even = 0
					is_midpoint_lt_even = 1
				} else {
					// ;
				}
			} else if lsb == 1 {
				if is_midpoint_lt_even != 0 {
					// res = res + 1
					res.w[0]++
					if res.w[0] == 0x0 {
						res.w[1]++
					}
					// check for rounding overflow
					if res.w[1] == 0x0001ed09bead87c0 &&
						res.w[0] == 0x378d8e6400000000 {
						// res = 10^34 => rounding overflow
						res.w[1] = 0x0000314dc6448d93
						res.w[0] = 0x38c15b0a00000000 // 10^33
						e3++
					}
				} else if is_midpoint_gt_even != 0 {
					// res = res - 1
					res.w[0]--
					if res.w[0] == 0xffffffffffffffff {
						res.w[1]--
					}
					// if the result is pure zero, the sign depends on the rounding
					// mode (x*y and z had opposite signs)
					if res.w[1] == 0x0 && res.w[0] == 0x0 {
						if rnd_mode != BID_ROUNDING_DOWN {
							z_sign = 0x0000000000000000
						} else {
							z_sign = 0x8000000000000000
						}
						// the exponent is max (e3, expmin)
						res.w[1] = 0x0
						res.w[0] = 0x0
						*ptr_is_midpoint_lt_even = is_midpoint_lt_even
						*ptr_is_midpoint_gt_even = is_midpoint_gt_even
						*ptr_is_inexact_lt_midpoint = is_inexact_lt_midpoint
						*ptr_is_inexact_gt_midpoint = is_inexact_gt_midpoint
						// BID_SWAP128 (res); BID_RETURN (res)
						*e3_ptr = e3
						return
					}
				} else {
					// ;
				}
			} else {
				// ;
			}
		}
		// check for underflow
		if e3 == expmin128 { // and if significand < 10^33 => result is tiny
			if (res.w[1]&MASK_COEFF128) < 0x0000314dc6448d93 ||
				((res.w[1]&MASK_COEFF128) == 0x0000314dc6448d93 &&
					res.w[0] < 0x38c15b0a00000000) {
				is_tiny = 1
			}
			// DECIMAL_TINY_DETECTION_AFTER_ROUNDING = 0
			if ((res.w[1] & 0x7fffffffffffffff) == 0x0000314dc6448d93) &&
				(res.w[0] == 0x38c15b0a00000000) && // 10^33*10^-6176
				(z_sign != p_sign) {
				is_tiny = 1
			}
		} else if e3 < expmin128 {
			// the result is tiny, so we must truncate more of res
			is_tiny = 1
			x0 = expmin128 - e3
			is_inexact_lt_midpoint0 = is_inexact_lt_midpoint
			is_inexact_gt_midpoint0 = is_inexact_gt_midpoint
			is_midpoint_lt_even0 = is_midpoint_lt_even
			is_midpoint_gt_even0 = is_midpoint_gt_even
			is_inexact_lt_midpoint = 0
			is_inexact_gt_midpoint = 0
			is_midpoint_lt_even = 0
			is_midpoint_gt_even = 0
			// determine the number of decimal digits in res
			if res.w[1] == 0x0 {
				// between 1 and 19 digits
				for ind = 1; ind <= 19; ind++ {
					if res.w[0] < bid_ten2k64[ind] {
						break
					}
				}
				// ind digits
			} else if res.w[1] < bid_ten2k128[0].w[1] ||
				(res.w[1] == bid_ten2k128[0].w[1] &&
					res.w[0] < bid_ten2k128[0].w[0]) {
				// 20 digits
				ind = 20
			} else { // between 21 and 38 digits
				for ind = 1; ind <= 18; ind++ {
					if res.w[1] < bid_ten2k128[ind].w[1] ||
						(res.w[1] == bid_ten2k128[ind].w[1] &&
							res.w[0] < bid_ten2k128[ind].w[0]) {
						break
					}
				}
				// ind + 20 digits
				ind = ind + 20
			}

			if x0 == ind { // the result before rounding is 0.9... * 10^emin
				res.w[1] = 0x0
				res.w[0] = 0x1
				is_inexact_gt_midpoint = 1
			} else if ind <= 18 { // check that 2 <= ind
				// 2 <= ind <= 18, 1 <= x0 <= 17
				R64 = bid_round64_2_18(ind, x0, res.w[0], &incr_exp,
					&is_midpoint_lt_even, &is_midpoint_gt_even,
					&is_inexact_lt_midpoint,
					&is_inexact_gt_midpoint)
				if incr_exp != 0 {
					// R64 = 10^(ind-x0), 1 <= ind - x0 <= ind - 1, 1 <= ind - x0 <= 17
					R64 = bid_ten2k64[ind-x0]
				}
				res.w[1] = 0
				res.w[0] = R64
			} else if ind <= 38 {
				// 19 <= ind <= 38
				P128.w[1] = res.w[1]
				P128.w[0] = res.w[0]
				*res, incr_exp, is_midpoint_lt_even, is_midpoint_gt_even,
					is_inexact_lt_midpoint, is_inexact_gt_midpoint =
					bid_round128_19_38(ind, x0, P128)
				if incr_exp != 0 {
					// R128 = 10^(ind-x0), 1 <= ind - x0 <= ind - 1, 1 <= ind - x0 <= 37
					if ind-x0 <= 19 { // 1 <= ind - x0 <= 19
						res.w[0] = bid_ten2k64[ind-x0]
						// res.w[1] stays 0
					} else { // 20 <= ind - x0 <= 37
						res.w[0] = bid_ten2k128[ind-x0-20].w[0]
						res.w[1] = bid_ten2k128[ind-x0-20].w[1]
					}
				}
			}
			// avoid a double rounding error
			if (is_inexact_gt_midpoint0 != 0 || is_midpoint_lt_even0 != 0) &&
				is_midpoint_lt_even != 0 { // double rounding error upward
				// res = res - 1
				res.w[0]--
				if res.w[0] == 0xffffffffffffffff {
					res.w[1]--
				}
				is_midpoint_lt_even = 0
				is_inexact_lt_midpoint = 1
			} else if (is_inexact_lt_midpoint0 != 0 || is_midpoint_gt_even0 != 0) &&
				is_midpoint_gt_even != 0 { // double rounding error downward
				// res = res + 1
				res.w[0]++
				if res.w[0] == 0 {
					res.w[1]++
				}
				is_midpoint_gt_even = 0
				is_inexact_gt_midpoint = 1
			} else if is_midpoint_lt_even == 0 && is_midpoint_gt_even == 0 &&
				is_inexact_lt_midpoint == 0 && is_inexact_gt_midpoint == 0 {
				// if this second rounding was exact the result may still be
				// inexact because of the first rounding
				if is_inexact_gt_midpoint0 != 0 || is_midpoint_lt_even0 != 0 {
					is_inexact_gt_midpoint = 1
				}
				if is_inexact_lt_midpoint0 != 0 || is_midpoint_gt_even0 != 0 {
					is_inexact_lt_midpoint = 1
				}
			} else if is_midpoint_gt_even != 0 &&
				(is_inexact_gt_midpoint0 != 0 || is_midpoint_lt_even0 != 0) {
				// pulled up to a midpoint
				is_inexact_lt_midpoint = 1
				is_inexact_gt_midpoint = 0
				is_midpoint_lt_even = 0
				is_midpoint_gt_even = 0
			} else if is_midpoint_lt_even != 0 &&
				(is_inexact_lt_midpoint0 != 0 || is_midpoint_gt_even0 != 0) {
				// pulled down to a midpoint
				is_inexact_lt_midpoint = 0
				is_inexact_gt_midpoint = 1
				is_midpoint_lt_even = 0
				is_midpoint_gt_even = 0
			} else {
				// ;
			}
			// adjust exponent
			e3 = e3 + x0
			if is_midpoint_lt_even == 0 && is_midpoint_gt_even == 0 &&
				is_inexact_lt_midpoint == 0 && is_inexact_gt_midpoint == 0 {
				if is_midpoint_lt_even0 != 0 || is_midpoint_gt_even0 != 0 ||
					is_inexact_lt_midpoint0 != 0 || is_inexact_gt_midpoint0 != 0 {
					is_inexact_lt_midpoint = 1
				}
			}
		} else {
			// ; // not underflow
		}
		// check for inexact result
		if is_inexact_lt_midpoint != 0 || is_inexact_gt_midpoint != 0 ||
			is_midpoint_lt_even != 0 || is_midpoint_gt_even != 0 {
			// set the inexact flag
			*pfpsf |= BID_INEXACT_EXCEPTION
			if is_tiny != 0 {
				*pfpsf |= BID_UNDERFLOW_EXCEPTION
			}
		}
		// now check for significand = 10^34 (may have resulted from going
		// back to case2_repeat)
		if res.w[1] == 0x0001ed09bead87c0 &&
			res.w[0] == 0x378d8e6400000000 { // if res = 10^34
			res.w[1] = 0x0000314dc6448d93 // res = 10^33
			res.w[0] = 0x38c15b0a00000000
			e3 = e3 + 1
		}
		res.w[1] = z_sign | (uint64(e3+6176) << 49) | res.w[1]
		// check for overflow
		if rnd_mode == BID_ROUNDING_TO_NEAREST && e3 > expmax {
			res.w[1] = z_sign | 0x7800000000000000 // +/-inf
			res.w[0] = 0x0000000000000000
			*pfpsf |= (BID_INEXACT_EXCEPTION | BID_OVERFLOW_EXCEPTION)
		}
		if rnd_mode != BID_ROUNDING_TO_NEAREST {
			bid_rounding_correction(rnd_mode,
				is_inexact_lt_midpoint,
				is_inexact_gt_midpoint,
				is_midpoint_lt_even, is_midpoint_gt_even,
				e3, res, pfpsf)
		}
		*ptr_is_midpoint_lt_even = is_midpoint_lt_even
		*ptr_is_midpoint_gt_even = is_midpoint_gt_even
		*ptr_is_inexact_lt_midpoint = is_inexact_lt_midpoint
		*ptr_is_inexact_gt_midpoint = is_inexact_gt_midpoint
		// BID_SWAP128 (res); BID_RETURN (res)
		*e3_ptr = e3
		break // normal exit from case2_repeat loop
	} // end for (case2_repeat loop)
}

func bid_fma_case7(
	p34 int, res *BID_UINT128,
	ptr_is_midpoint_lt_even, ptr_is_midpoint_gt_even,
	ptr_is_inexact_lt_midpoint, ptr_is_inexact_gt_midpoint *int,
	p_sign, z_sign uint64,
	q3, q4 int, e4_ptr *int, delta int,
	C3 *BID_UINT128, C4 BID_UINT256,
	rnd_mode int, pfpsf *uint32) {

	e4 := *e4_ptr
	var is_midpoint_lt_even, is_midpoint_gt_even int
	var is_inexact_lt_midpoint, is_inexact_gt_midpoint int
	var incr_exp int
	var P128 BID_UINT128
	var P192 BID_UINT192
	var R192 BID_UINT192
	var R256 BID_UINT256
	var x0 int

	// truncate C4 to p34 digits into res
	x0 = q4 - p34
	if q4 <= 38 {
		P128.w[1] = C4.w[1]
		P128.w[0] = C4.w[0]
		*res, incr_exp, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint = bid_round128_19_38(q4, x0, P128)
	} else if q4 <= 57 {
		P192.w[2] = C4.w[2]
		P192.w[1] = C4.w[1]
		P192.w[0] = C4.w[0]
		R192, incr_exp, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint = bid_round192_39_57(q4, x0, P192)
		res.w[0] = R192.w[0]
		res.w[1] = R192.w[1]
	} else {
		R256, incr_exp, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint = bid_round256_58_76(q4, x0, C4)
		res.w[0] = R256.w[0]
		res.w[1] = R256.w[1]
	}
	e4 = e4 + x0
	if incr_exp != 0 {
		e4 = e4 + 1
	}

	if is_midpoint_lt_even == 0 && is_midpoint_gt_even == 0 &&
		is_inexact_lt_midpoint == 0 && is_inexact_gt_midpoint == 0 {
		if p_sign == z_sign {
			is_inexact_lt_midpoint = 1
		} else {
			if res.w[1] != 0x0000314dc6448d93 || res.w[0] != 0x38c15b0a00000000 {
				is_inexact_gt_midpoint = 1
			} else {
				// res = 10^33 and exact: special case
				if delta > p34+1 {
					is_inexact_gt_midpoint = 1
				} else { // delta == p34 + 1
					if q3 <= 19 {
						if C3.w[0] < bid_midpoint64[q3-1] {
							is_inexact_gt_midpoint = 1
						} else if C3.w[0] == bid_midpoint64[q3-1] {
							is_midpoint_lt_even = 1
						} else {
							res.w[1] = 0x0001ed09bead87c0
							res.w[0] = 0x378d8e63ffffffff
							e4 = e4 - 1
							is_inexact_lt_midpoint = 1
						}
					} else {
						if C3.w[1] < bid_midpoint128[q3-20].w[1] ||
							(C3.w[1] == bid_midpoint128[q3-20].w[1] &&
								C3.w[0] < bid_midpoint128[q3-20].w[0]) {
							is_inexact_gt_midpoint = 1
						} else if C3.w[1] == bid_midpoint128[q3-20].w[1] &&
							C3.w[0] == bid_midpoint128[q3-20].w[0] {
							is_midpoint_lt_even = 1
						} else {
							res.w[1] = 0x0001ed09bead87c0
							res.w[0] = 0x378d8e63ffffffff
							e4 = e4 - 1
							is_inexact_lt_midpoint = 1
						}
					}
				}
			}
		}
	} else if is_midpoint_lt_even != 0 {
		if z_sign != p_sign {
			res.w[0] = res.w[0] - 1
			if res.w[0] == 0xffffffffffffffff {
				res.w[1]--
			}
			if res.w[1] == 0x0000314dc6448d93 && res.w[0] == 0x38c15b09ffffffff {
				res.w[1] = 0x0001ed09bead87c0
				res.w[0] = 0x378d8e63ffffffff
				e4 = e4 - 1
			}
			is_midpoint_lt_even = 0
			is_inexact_lt_midpoint = 1
		} else {
			is_midpoint_lt_even = 0
			is_inexact_gt_midpoint = 1
		}
	} else if is_midpoint_gt_even != 0 {
		if z_sign == p_sign {
			res.w[0] = res.w[0] + 1
			if res.w[0] == 0x0000000000000000 {
				res.w[1]++
			}
			is_midpoint_gt_even = 0
			is_inexact_gt_midpoint = 1
		} else {
			is_midpoint_gt_even = 0
			is_inexact_lt_midpoint = 1
		}
	}

	// check for overflow
	if rnd_mode == BID_ROUNDING_TO_NEAREST && e4 > expmax {
		res.w[1] = p_sign | 0x7800000000000000
		res.w[0] = 0x0000000000000000
		*pfpsf |= (BID_OVERFLOW_EXCEPTION | BID_INEXACT_EXCEPTION)
	} else {
		p_exp := uint64(e4+6176) << 49
		res.w[1] = p_sign | (p_exp & MASK_EXP_128) | res.w[1]
	}
	if rnd_mode != BID_ROUNDING_TO_NEAREST {
		bid_rounding_correction(rnd_mode,
			is_inexact_lt_midpoint,
			is_inexact_gt_midpoint,
			is_midpoint_lt_even, is_midpoint_gt_even,
			e4, res, pfpsf)
	}
	if is_inexact_lt_midpoint != 0 || is_inexact_gt_midpoint != 0 ||
		is_midpoint_lt_even != 0 || is_midpoint_gt_even != 0 {
		*pfpsf |= BID_INEXACT_EXCEPTION
	}

	*ptr_is_midpoint_lt_even = is_midpoint_lt_even
	*ptr_is_midpoint_gt_even = is_midpoint_gt_even
	*ptr_is_inexact_lt_midpoint = is_inexact_lt_midpoint
	*ptr_is_inexact_gt_midpoint = is_inexact_gt_midpoint
	*e4_ptr = e4
}

func bid_fma_cases_11_12(
	p34 int, res *BID_UINT128,
	ptr_is_midpoint_lt_even, ptr_is_midpoint_gt_even,
	ptr_is_inexact_lt_midpoint, ptr_is_inexact_gt_midpoint *int,
	p_sign, z_sign uint64,
	q3, q4 int, e3_ptr, e4_ptr *int, delta int,
	C3 *BID_UINT128, C4 BID_UINT256,
	rnd_mode int, pfpsf *uint32) {

	// Cases (11) and (12) from bid128_fma.c lines 3157-3607
	e3 := *e3_ptr
	e4 := *e4_ptr
	var is_midpoint_lt_even, is_midpoint_gt_even int
	var is_inexact_lt_midpoint, is_inexact_gt_midpoint int
	var is_inexact_lt_midpoint0, is_inexact_gt_midpoint0 int
	var is_midpoint_lt_even0, is_midpoint_gt_even0 int
	var R64 uint64
	var R128, P128, P192_128 BID_UINT128
	var R192 BID_UINT192
	var R256 BID_UINT256
	var incr_exp int
	var ind, x0 int
	var lsb uint64
	var is_tiny int
	var lt_half_ulp, eq_half_ulp, gt_half_ulp int

	expmin128 := -6176

	// round C3 to nearest to q3 - x0 digits, where x0 = e4 - e3
	x0 = e4 - e3
	if q3 <= 18 {
		R64 = bid_round64_2_18(q3, x0, C3.w[0], &incr_exp,
			&is_midpoint_lt_even, &is_midpoint_gt_even,
			&is_inexact_lt_midpoint, &is_inexact_gt_midpoint)
		C3.w[0] = R64
	} else if q3 <= 38 {
		R128, incr_exp, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint = bid_round128_19_38(q3, x0, *C3)
		C3.w[1] = R128.w[1]
		C3.w[0] = R128.w[0]
	}
	if incr_exp != 0 {
		P128.w[1] = C3.w[1]
		P128.w[0] = C3.w[0]
		*C3 = __mul_64x128_to_128(bid_ten2k64[1], P128)
	}
	e3 = e3 + x0 // this is e4

	// now add/subtract the 256-bit C4 and the new (and shorter) 128-bit C3
	R256.w[3] = 0
	R256.w[2] = 0
	R256.w[1] = C3.w[1]
	R256.w[0] = C3.w[0]
	if p_sign == z_sign {
		R256 = bid_add256(C4, R256)
	} else {
		R256 = bid_sub256(C4, R256)
		lsb = C4.w[0] & 0x01
		if is_inexact_lt_midpoint != 0 {
			is_inexact_lt_midpoint = 0
			is_inexact_gt_midpoint = 1
		} else if is_inexact_gt_midpoint != 0 {
			is_inexact_gt_midpoint = 0
			is_inexact_lt_midpoint = 1
		} else if lsb == 0 {
			if is_midpoint_lt_even != 0 {
				is_midpoint_lt_even = 0
				is_midpoint_gt_even = 1
			} else if is_midpoint_gt_even != 0 {
				is_midpoint_gt_even = 0
				is_midpoint_lt_even = 1
			}
		} else if lsb == 1 {
			if is_midpoint_lt_even != 0 {
				R256.w[0]++
				if R256.w[0] == 0 {
					R256.w[1]++
					if R256.w[1] == 0 {
						R256.w[2]++
						if R256.w[2] == 0 {
							R256.w[3]++
						}
					}
				}
			} else if is_midpoint_gt_even != 0 {
				R256.w[0]--
				if R256.w[0] == 0xffffffffffffffff {
					R256.w[1]--
					if R256.w[1] == 0xffffffffffffffff {
						R256.w[2]--
						if R256.w[2] == 0xffffffffffffffff {
							R256.w[3]--
						}
					}
				}
			}
		}
	}

	ind = bid_bid_nr_digits256(R256)

	if ind < p34 {
		// do nothing
	} else if ind == p34 {
		res.w[1] = R256.w[1]
		res.w[0] = R256.w[0]
	} else { // ind > p34
		x0 = ind - p34
		is_inexact_lt_midpoint0 = is_inexact_lt_midpoint
		is_inexact_gt_midpoint0 = is_inexact_gt_midpoint
		is_midpoint_lt_even0 = is_midpoint_lt_even
		is_midpoint_gt_even0 = is_midpoint_gt_even
		is_inexact_lt_midpoint = 0
		is_inexact_gt_midpoint = 0
		is_midpoint_lt_even = 0
		is_midpoint_gt_even = 0

		if ind <= 38 {
			P128.w[1] = R256.w[1]
			P128.w[0] = R256.w[0]
			R128, incr_exp, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint = bid_round128_19_38(ind, x0, P128)
		} else if ind <= 57 {
			var P192 BID_UINT192
			P192.w[2] = R256.w[2]
			P192.w[1] = R256.w[1]
			P192.w[0] = R256.w[0]
			R192, incr_exp, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint = bid_round192_39_57(ind, x0, P192)
			R128.w[1] = R192.w[1]
			R128.w[0] = R192.w[0]
		} else {
			R256, incr_exp, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint = bid_round256_58_76(ind, x0, R256)
			R128.w[1] = R256.w[1]
			R128.w[0] = R256.w[0]
		}
		e4 = e4 + x0 + incr_exp
		res.w[1] = R128.w[1]
		res.w[0] = R128.w[0]

		// avoid double rounding error
		if (is_inexact_gt_midpoint0 != 0 || is_midpoint_lt_even0 != 0) &&
			is_midpoint_lt_even != 0 {
			res.w[0]--
			if res.w[0] == 0xffffffffffffffff {
				res.w[1]--
			}
			is_midpoint_lt_even = 0
			is_inexact_lt_midpoint = 1
			if res.w[1] == 0x0000314dc6448d93 && res.w[0] == 0x38c15b09ffffffff {
				res.w[1] = 0x0001ed09bead87c0
				res.w[0] = 0x378d8e63ffffffff
				e4--
			}
		} else if (is_inexact_lt_midpoint0 != 0 || is_midpoint_gt_even0 != 0) &&
			is_midpoint_gt_even != 0 {
			res.w[0]++
			if res.w[0] == 0 {
				res.w[1]++
			}
			is_midpoint_gt_even = 0
			is_inexact_gt_midpoint = 1
		} else if is_midpoint_lt_even == 0 && is_midpoint_gt_even == 0 &&
			is_inexact_lt_midpoint == 0 && is_inexact_gt_midpoint == 0 {
			if is_inexact_gt_midpoint0 != 0 || is_midpoint_lt_even0 != 0 {
				is_inexact_gt_midpoint = 1
			}
			if is_inexact_lt_midpoint0 != 0 || is_midpoint_gt_even0 != 0 {
				is_inexact_lt_midpoint = 1
			}
		} else if is_midpoint_gt_even != 0 &&
			(is_inexact_gt_midpoint0 != 0 || is_midpoint_lt_even0 != 0) {
			is_inexact_lt_midpoint = 1
			is_inexact_gt_midpoint = 0
			is_midpoint_lt_even = 0
			is_midpoint_gt_even = 0
		} else if is_midpoint_lt_even != 0 &&
			(is_inexact_lt_midpoint0 != 0 || is_midpoint_gt_even0 != 0) {
			is_inexact_lt_midpoint = 0
			is_inexact_gt_midpoint = 1
			is_midpoint_lt_even = 0
			is_midpoint_gt_even = 0
		}
	}

	// determine tininess
	if rnd_mode == BID_ROUNDING_TO_NEAREST {
		if e4 < expmin128 {
			is_tiny = 1
		}
	} else {
		P128.w[1] = p_sign | 0x3040000000000000 | res.w[1]
		P128.w[0] = res.w[0]
		bid_rounding_correction(rnd_mode,
			is_inexact_lt_midpoint, is_inexact_gt_midpoint,
			is_midpoint_lt_even, is_midpoint_gt_even,
			0, &P128, pfpsf)
		scale := int((P128.w[1]&MASK_EXP128)>>49) - 6176
		if e4+scale < expmin128 {
			is_tiny = 1
		}
	}

	res.w[1] = p_sign | (uint64(e4+6176) << 49) | res.w[1]

	ind = p34

	// check for overflow if RN
	if rnd_mode == BID_ROUNDING_TO_NEAREST && (ind+e4) > (p34+expmax) {
		res.w[1] = p_sign | 0x7800000000000000
		res.w[0] = 0x0000000000000000
		*pfpsf |= BID_INEXACT_EXCEPTION | BID_OVERFLOW_EXCEPTION
		*ptr_is_midpoint_lt_even = is_midpoint_lt_even
		*ptr_is_midpoint_gt_even = is_midpoint_gt_even
		*ptr_is_inexact_lt_midpoint = is_inexact_lt_midpoint
		*ptr_is_inexact_gt_midpoint = is_inexact_gt_midpoint
		return
	}

	if e4 < expmin128 {
		x0 = expmin128 - e4
		is_inexact_lt_midpoint0 = is_inexact_lt_midpoint
		is_inexact_gt_midpoint0 = is_inexact_gt_midpoint
		is_midpoint_lt_even0 = is_midpoint_lt_even
		is_midpoint_gt_even0 = is_midpoint_gt_even
		is_inexact_lt_midpoint = 0
		is_inexact_gt_midpoint = 0
		is_midpoint_lt_even = 0
		is_midpoint_gt_even = 0

		if x0 > ind {
			is_inexact_lt_midpoint = 1
			res.w[1] = p_sign | 0x0000000000000000
			res.w[0] = 0x0000000000000000
			e4 = expmin128
		} else if x0 == ind {
			R128.w[1] = res.w[1] & MASK_COEFF128
			R128.w[0] = res.w[0]
			if ind <= 19 {
				if R128.w[0] < bid_midpoint64[ind-1] {
					lt_half_ulp = 1
					is_inexact_lt_midpoint = 1
				} else if R128.w[0] == bid_midpoint64[ind-1] {
					eq_half_ulp = 1
					is_midpoint_gt_even = 1
				} else {
					gt_half_ulp = 1
					is_inexact_gt_midpoint = 1
				}
			} else {
				if R128.w[1] < bid_midpoint128[ind-20].w[1] ||
					(R128.w[1] == bid_midpoint128[ind-20].w[1] &&
						R128.w[0] < bid_midpoint128[ind-20].w[0]) {
					lt_half_ulp = 1
					is_inexact_lt_midpoint = 1
				} else if R128.w[1] == bid_midpoint128[ind-20].w[1] &&
					R128.w[0] == bid_midpoint128[ind-20].w[0] {
					eq_half_ulp = 1
					is_midpoint_gt_even = 1
				} else {
					gt_half_ulp = 1
					is_inexact_gt_midpoint = 1
				}
			}
			_ = lt_half_ulp
			_ = eq_half_ulp
			if lt_half_ulp != 0 || eq_half_ulp != 0 {
				res.w[1] = 0x0000000000000000
				res.w[0] = 0x0000000000000000
			} else { // gt_half_ulp
				res.w[1] = 0x0000000000000000
				res.w[0] = 0x0000000000000001
			}
			_ = gt_half_ulp
			res.w[1] = p_sign | res.w[1]
			e4 = expmin128
		} else { // 1 <= x0 <= ind - 1 <= 33
			if ind <= 18 {
				R64 = bid_round64_2_18(ind, x0, res.w[0], &incr_exp,
					&is_midpoint_lt_even, &is_midpoint_gt_even,
					&is_inexact_lt_midpoint, &is_inexact_gt_midpoint)
				res.w[1] = 0x0
				res.w[0] = R64
			} else if ind <= 38 {
				P128.w[1] = res.w[1] & MASK_COEFF128
				P128.w[0] = res.w[0]
				P192_128, incr_exp, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint = bid_round128_19_38(ind, x0, P128)
				res.w[1] = P192_128.w[1]
				res.w[0] = P192_128.w[0]
			}
			e4 = e4 + x0
			if incr_exp != 0 {
				P128.w[1] = res.w[1] & MASK_COEFF128
				P128.w[0] = res.w[0]
				*res = __mul_64x128_to_128(bid_ten2k64[1], P128)
			}
			res.w[1] = p_sign | (uint64(e4+6176) << 49) | (res.w[1] & MASK_COEFF128)
			// avoid double rounding error
			if (is_inexact_gt_midpoint0 != 0 || is_midpoint_lt_even0 != 0) &&
				is_midpoint_lt_even != 0 {
				res.w[0]--
				if res.w[0] == 0xffffffffffffffff {
					res.w[1]--
				}
				is_midpoint_lt_even = 0
				is_inexact_lt_midpoint = 1
			} else if (is_inexact_lt_midpoint0 != 0 || is_midpoint_gt_even0 != 0) &&
				is_midpoint_gt_even != 0 {
				res.w[0]++
				if res.w[0] == 0 {
					res.w[1]++
				}
				is_midpoint_gt_even = 0
				is_inexact_gt_midpoint = 1
			} else if is_midpoint_lt_even == 0 && is_midpoint_gt_even == 0 &&
				is_inexact_lt_midpoint == 0 && is_inexact_gt_midpoint == 0 {
				if is_inexact_gt_midpoint0 != 0 || is_midpoint_lt_even0 != 0 {
					is_inexact_gt_midpoint = 1
				}
				if is_inexact_lt_midpoint0 != 0 || is_midpoint_gt_even0 != 0 {
					is_inexact_lt_midpoint = 1
				}
			} else if is_midpoint_gt_even != 0 &&
				(is_inexact_gt_midpoint0 != 0 || is_midpoint_lt_even0 != 0) {
				is_inexact_lt_midpoint = 1
				is_inexact_gt_midpoint = 0
				is_midpoint_lt_even = 0
				is_midpoint_gt_even = 0
			} else if is_midpoint_lt_even != 0 &&
				(is_inexact_lt_midpoint0 != 0 || is_midpoint_gt_even0 != 0) {
				is_inexact_lt_midpoint = 0
				is_inexact_gt_midpoint = 1
				is_midpoint_lt_even = 0
				is_midpoint_gt_even = 0
			}
		}
	}

	// apply correction if not rounding to nearest
	if rnd_mode != BID_ROUNDING_TO_NEAREST {
		bid_rounding_correction(rnd_mode,
			is_inexact_lt_midpoint, is_inexact_gt_midpoint,
			is_midpoint_lt_even, is_midpoint_gt_even,
			e4, res, pfpsf)
	}
	// correction for tininess detection before rounding (DECIMAL_TINY_DETECTION_AFTER_ROUNDING=0)
	if ((res.w[1]&0x7fffffffffffffff) == 0x0000314dc6448d93 &&
		res.w[0] == 0x38c15b0a00000000) &&
		(((rnd_mode == BID_ROUNDING_TO_NEAREST || rnd_mode == BID_ROUNDING_TIES_AWAY) &&
			(is_midpoint_lt_even != 0 || is_inexact_gt_midpoint != 0)) ||
			((((rnd_mode == BID_ROUNDING_UP) && (res.w[1]&MASK_SIGN64 == 0)) ||
				((rnd_mode == BID_ROUNDING_DOWN) && (res.w[1]&MASK_SIGN64 != 0))) &&
				(is_midpoint_lt_even != 0 || is_midpoint_gt_even != 0 ||
					is_inexact_lt_midpoint != 0 || is_inexact_gt_midpoint != 0))) {
		is_tiny = 1
	}
	if is_midpoint_lt_even != 0 || is_midpoint_gt_even != 0 ||
		is_inexact_lt_midpoint != 0 || is_inexact_gt_midpoint != 0 {
		*pfpsf |= BID_INEXACT_EXCEPTION
		if is_tiny != 0 {
			*pfpsf |= BID_UNDERFLOW_EXCEPTION
		}
	}

	*ptr_is_midpoint_lt_even = is_midpoint_lt_even
	*ptr_is_midpoint_gt_even = is_midpoint_gt_even
	*ptr_is_inexact_lt_midpoint = is_inexact_lt_midpoint
	*ptr_is_inexact_gt_midpoint = is_inexact_gt_midpoint
	*e4_ptr = e4
}

// bid_add_and_round adds/subtracts C4 and C3 * 10^scale, then rounds.
// Ported from bid128_fma.c lines 258-645.
func bid_add_and_round(q3, q4, e4, delta, p34 int,
	z_sign, p_sign uint64,
	C3 BID_UINT128, C4 BID_UINT256,
	rnd_mode int,
	ptr_is_midpoint_lt_even, ptr_is_midpoint_gt_even,
	ptr_is_inexact_lt_midpoint, ptr_is_inexact_gt_midpoint *int,
	ptrfpsf *uint32, ptrres *BID_UINT128) {

	var scale, x0, ind int
	var R64 uint64
	var P128, R128 BID_UINT128
	var P192, R192 BID_UINT192
	var R256 BID_UINT256
	var is_midpoint_lt_even, is_midpoint_gt_even int
	var is_inexact_lt_midpoint, is_inexact_gt_midpoint int
	var is_midpoint_lt_even0, is_midpoint_gt_even0 int
	var is_inexact_lt_midpoint0, is_inexact_gt_midpoint0 int
	var incr_exp int
	var is_tiny int
	var lt_half_ulp, eq_half_ulp int
	res := *ptrres

	// scale C3 up by 10^(q4-delta-q3)
	scale = q4 - delta - q3

	// calculate C3 * 10^scale in R256
	if scale == 0 {
		R256.w[3] = 0x0
		R256.w[2] = 0x0
		R256.w[1] = C3.w[1]
		R256.w[0] = C3.w[0]
	} else if scale <= 19 {
		P128.w[1] = 0
		P128.w[0] = bid_ten2k64[scale]
		R256 = __mul_128x128_to_256(P128, C3)
	} else if scale <= 38 {
		R256 = __mul_128x128_to_256(bid_ten2k128[scale-20], C3)
	} else if scale <= 57 {
		R128 = __mul_64x128_to_128(bid_ten2k64[scale-38], C3)
		R256 = __mul_128x128_to_256(R128, bid_ten2k128[18])
	} else { // 58 <= scale <= 66
		R128 = __mul_64x128_to_128(C3.w[0], bid_ten2k128[scale-58])
		R256 = __mul_128x128_to_256(R128, bid_ten2k128[18])
	}

	if p_sign == z_sign { // R256 = C4 + R256
		R256 = bid_add256(C4, R256)
	} else { // R256 = C4 - R256
		if R256.w[3] > C4.w[3] || (R256.w[3] == C4.w[3] && R256.w[2] > C4.w[2]) ||
			(R256.w[3] == C4.w[3] && R256.w[2] == C4.w[2] && R256.w[1] > C4.w[1]) ||
			(R256.w[3] == C4.w[3] && R256.w[2] == C4.w[2] && R256.w[1] == C4.w[1] &&
				R256.w[0] >= C4.w[0]) { // C3 * 10^scale >= C4
			R256 = bid_sub256(R256, C4)
			p_sign = z_sign
		} else {
			R256 = bid_sub256(C4, R256)
		}
		// if the result is pure zero
		if R256.w[3] == 0x0 && R256.w[2] == 0x0 &&
			R256.w[1] == 0x0 && R256.w[0] == 0x0 {
			if rnd_mode != BID_ROUNDING_DOWN {
				p_sign = 0x0000000000000000
			} else {
				p_sign = 0x8000000000000000
			}
			if e4 < -6176 {
				e4 = expmin128
			}
			res.w[1] = p_sign | (uint64(e4+6176) << 49)
			res.w[0] = 0x0
			*ptrres = res
			return
		}
	}

	// determine the number of decimal digits in R256
	ind = bid_bid_nr_digits256(R256)

	if ind <= p34 {
		if ind+e4 < p34+expmin128 {
			is_tiny = 1
		}
		res.w[1] = p_sign | (uint64(e4+6176) << 49) | R256.w[1]
		res.w[0] = R256.w[0]
	} else { // ind > p34
		x0 = ind - p34
		if ind <= 38 {
			P128.w[1] = R256.w[1]
			P128.w[0] = R256.w[0]
			R128, incr_exp, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint = bid_round128_19_38(ind, x0, P128)
		} else if ind <= 57 {
			P192.w[2] = R256.w[2]
			P192.w[1] = R256.w[1]
			P192.w[0] = R256.w[0]
			R192, incr_exp, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint = bid_round192_39_57(ind, x0, P192)
			R128.w[1] = R192.w[1]
			R128.w[0] = R192.w[0]
		} else { // ind <= 68
			R256, incr_exp, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint = bid_round256_58_76(ind, x0, R256)
			R128.w[1] = R256.w[1]
			R128.w[0] = R256.w[0]
		}
		// DECIMAL_TINY_DETECTION_AFTER_ROUNDING = 0
		if e4+x0 < expmin128 {
			is_tiny = 1
		}
		e4 = e4 + x0 + incr_exp
		if rnd_mode == BID_ROUNDING_TO_NEAREST {
			// nothing extra for RN
		} else {
			P128.w[1] = p_sign | 0x3040000000000000 | R128.w[1]
			P128.w[0] = R128.w[0]
			bid_rounding_correction(rnd_mode,
				is_inexact_lt_midpoint,
				is_inexact_gt_midpoint, is_midpoint_lt_even,
				is_midpoint_gt_even, 0, &P128, ptrfpsf)
			scale = int((P128.w[1]&MASK_EXP_128)>>49) - 6176
			// DECIMAL_TINY_DETECTION_AFTER_ROUNDING = 0 handled above
		}
		ind = p34
		res.w[1] = p_sign | (uint64(e4+6176) << 49) | R128.w[1]
		res.w[0] = R128.w[0]
	}

	// check for overflow if RN
	if rnd_mode == BID_ROUNDING_TO_NEAREST && (ind+e4) > (p34+expmax) {
		res.w[1] = p_sign | 0x7800000000000000
		res.w[0] = 0x0000000000000000
		*ptrres = res
		*ptrfpsf |= (BID_INEXACT_EXCEPTION | BID_OVERFLOW_EXCEPTION)
		return
	}

	if e4 < expmin128 {
		x0 = expmin128 - e4
		is_inexact_lt_midpoint0 = is_inexact_lt_midpoint
		is_inexact_gt_midpoint0 = is_inexact_gt_midpoint
		is_midpoint_lt_even0 = is_midpoint_lt_even
		is_midpoint_gt_even0 = is_midpoint_gt_even
		is_inexact_lt_midpoint = 0
		is_inexact_gt_midpoint = 0
		is_midpoint_lt_even = 0
		is_midpoint_gt_even = 0

		if x0 > ind {
			is_inexact_lt_midpoint = 1
			res.w[1] = p_sign | 0x0000000000000000
			res.w[0] = 0x0000000000000000
			e4 = expmin128
		} else if x0 == ind {
			R128.w[1] = res.w[1] & MASK_COEFF128
			R128.w[0] = res.w[0]
			if ind <= 19 {
				if R128.w[0] < bid_midpoint64[ind-1] {
					lt_half_ulp = 1
					is_inexact_lt_midpoint = 1
				} else if R128.w[0] == bid_midpoint64[ind-1] {
					eq_half_ulp = 1
					is_midpoint_gt_even = 1
				} else {
					is_inexact_gt_midpoint = 1
				}
			} else {
				if R128.w[1] < bid_midpoint128[ind-20].w[1] ||
					(R128.w[1] == bid_midpoint128[ind-20].w[1] &&
						R128.w[0] < bid_midpoint128[ind-20].w[0]) {
					lt_half_ulp = 1
					is_inexact_lt_midpoint = 1
				} else if R128.w[1] == bid_midpoint128[ind-20].w[1] &&
					R128.w[0] == bid_midpoint128[ind-20].w[0] {
					eq_half_ulp = 1
					is_midpoint_gt_even = 1
				} else {
					is_inexact_gt_midpoint = 1
				}
			}
			if lt_half_ulp != 0 || eq_half_ulp != 0 {
				res.w[1] = 0x0000000000000000
				res.w[0] = 0x0000000000000000
			} else {
				res.w[1] = 0x0000000000000000
				res.w[0] = 0x0000000000000001
			}
			res.w[1] = p_sign | res.w[1]
			e4 = expmin128
		} else { // 1 <= x0 <= ind - 1
			if ind <= 18 {
				R64 = bid_round64_2_18(ind, x0, res.w[0], &incr_exp,
					&is_midpoint_lt_even, &is_midpoint_gt_even,
					&is_inexact_lt_midpoint, &is_inexact_gt_midpoint)
				res.w[1] = 0x0
				res.w[0] = R64
			} else if ind <= 38 {
				P128.w[1] = res.w[1] & MASK_COEFF128
				P128.w[0] = res.w[0]
				res, incr_exp, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint = bid_round128_19_38(ind, x0, P128)
			}
			e4 = e4 + x0
			if incr_exp != 0 {
				P128.w[1] = res.w[1] & MASK_COEFF128
				P128.w[0] = res.w[0]
				res = __mul_64x128_to_128(bid_ten2k64[1], P128)
			}
			res.w[1] = p_sign | (uint64(e4+6176) << 49) | (res.w[1] & MASK_COEFF128)

			// avoid double rounding error
			if (is_inexact_gt_midpoint0 != 0 || is_midpoint_lt_even0 != 0) &&
				is_midpoint_lt_even != 0 {
				res.w[0]--
				if res.w[0] == 0xffffffffffffffff {
					res.w[1]--
				}
				is_midpoint_lt_even = 0
				is_inexact_lt_midpoint = 1
			} else if (is_inexact_lt_midpoint0 != 0 || is_midpoint_gt_even0 != 0) &&
				is_midpoint_gt_even != 0 {
				res.w[0]++
				if res.w[0] == 0 {
					res.w[1]++
				}
				is_midpoint_gt_even = 0
				is_inexact_gt_midpoint = 1
			} else if is_midpoint_lt_even == 0 && is_midpoint_gt_even == 0 &&
				is_inexact_lt_midpoint == 0 && is_inexact_gt_midpoint == 0 {
				if is_inexact_gt_midpoint0 != 0 || is_midpoint_lt_even0 != 0 {
					is_inexact_gt_midpoint = 1
				}
				if is_inexact_lt_midpoint0 != 0 || is_midpoint_gt_even0 != 0 {
					is_inexact_lt_midpoint = 1
				}
			} else if is_midpoint_gt_even != 0 &&
				(is_inexact_gt_midpoint0 != 0 || is_midpoint_lt_even0 != 0) {
				is_inexact_lt_midpoint = 1
				is_inexact_gt_midpoint = 0
				is_midpoint_lt_even = 0
				is_midpoint_gt_even = 0
			} else if is_midpoint_lt_even != 0 &&
				(is_inexact_lt_midpoint0 != 0 || is_midpoint_gt_even0 != 0) {
				is_inexact_lt_midpoint = 0
				is_inexact_gt_midpoint = 1
				is_midpoint_lt_even = 0
				is_midpoint_gt_even = 0
			}
		}
	}

	// apply correction if not rounding to nearest
	if rnd_mode != BID_ROUNDING_TO_NEAREST {
		bid_rounding_correction(rnd_mode,
			is_inexact_lt_midpoint, is_inexact_gt_midpoint,
			is_midpoint_lt_even, is_midpoint_gt_even,
			e4, &res, ptrfpsf)
	}
	if is_midpoint_lt_even != 0 || is_midpoint_gt_even != 0 ||
		is_inexact_lt_midpoint != 0 || is_inexact_gt_midpoint != 0 {
		*ptrfpsf |= BID_INEXACT_EXCEPTION
		if is_tiny != 0 {
			*ptrfpsf |= BID_UNDERFLOW_EXCEPTION
		}
	}

	*ptr_is_midpoint_lt_even = is_midpoint_lt_even
	*ptr_is_midpoint_gt_even = is_midpoint_gt_even
	*ptr_is_inexact_lt_midpoint = is_inexact_lt_midpoint
	*ptr_is_inexact_gt_midpoint = is_inexact_gt_midpoint
	*ptrres = res
}

// bid_bid_nr_digits256 determines the number of decimal digits in R256.
// Ported from bid128_fma.c lines 196-253.
func bid_bid_nr_digits256(R256 BID_UINT256) int {
	var ind int
	if R256.w[3] == 0x0 && R256.w[2] == 0x0 && R256.w[1] == 0x0 {
		for ind = 1; ind <= 19; ind++ {
			if R256.w[0] < bid_ten2k64[ind] {
				break
			}
		}
	} else if R256.w[3] == 0x0 && R256.w[2] == 0x0 &&
		(R256.w[1] < bid_ten2k128[0].w[1] ||
			(R256.w[1] == bid_ten2k128[0].w[1] && R256.w[0] < bid_ten2k128[0].w[0])) {
		ind = 20
	} else if R256.w[3] == 0x0 && R256.w[2] == 0x0 {
		for ind = 1; ind <= 18; ind++ {
			if R256.w[1] < bid_ten2k128[ind].w[1] ||
				(R256.w[1] == bid_ten2k128[ind].w[1] && R256.w[0] < bid_ten2k128[ind].w[0]) {
				break
			}
		}
		ind = ind + 20
	} else if R256.w[3] == 0x0 &&
		(R256.w[2] < bid_ten2k256[0].w[2] ||
			(R256.w[2] == bid_ten2k256[0].w[2] && R256.w[1] < bid_ten2k256[0].w[1]) ||
			(R256.w[2] == bid_ten2k256[0].w[2] && R256.w[1] == bid_ten2k256[0].w[1] &&
				R256.w[0] < bid_ten2k256[0].w[0])) {
		ind = 39
	} else {
		for ind = 1; ind <= 29; ind++ {
			if R256.w[3] < bid_ten2k256[ind].w[3] ||
				(R256.w[3] == bid_ten2k256[ind].w[3] && R256.w[2] < bid_ten2k256[ind].w[2]) ||
				(R256.w[3] == bid_ten2k256[ind].w[3] && R256.w[2] == bid_ten2k256[ind].w[2] &&
					R256.w[1] < bid_ten2k256[ind].w[1]) ||
				(R256.w[3] == bid_ten2k256[ind].w[3] && R256.w[2] == bid_ten2k256[ind].w[2] &&
					R256.w[1] == bid_ten2k256[ind].w[1] && R256.w[0] < bid_ten2k256[ind].w[0]) {
				break
			}
		}
		ind = ind + 39
	}
	return ind
}
