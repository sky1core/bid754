// Ported from: Intel bid128_fma.c: bid128_ext_fma
// Mechanical translation - all logic preserved exactly.
// This is the core FMA engine for Decimal128.
// Work in progress - NaN/Inf/Zero handling complete, arithmetic body pending.

package bidgo

const P34 = 34

// MASK aliases for bid128 (same values as 64-bit, operating on w[1])
const (
	MASK_NAN_128     = 0x7c00000000000000
	MASK_SNAN_128    = 0x7e00000000000000
	MASK_ANY_INF_128 = 0x7c00000000000000
	MASK_INF_128     = 0x7800000000000000
	MASK_SIGN_128    = 0x8000000000000000
	MASK_COEFF_128   = 0x0001ffffffffffff
	MASK_EXP_128     = 0x7ffe000000000000
)

// bid128_ext_fma is the extended FMA returning midpoint/inexact info.
func bid128_ext_fma(x, y, z BID_UINT128, rnd_mode int, pfpsf *uint32) (
	res BID_UINT128,
	is_midpoint_lt_even, is_midpoint_gt_even,
	is_inexact_lt_midpoint, is_inexact_gt_midpoint int) {

	res = BID_UINT128{w: [2]uint64{0xbaddbaddbaddbadd, 0xbaddbaddbaddbadd}}
	var x_sign, y_sign, z_sign, p_sign uint64
	var x_exp, y_exp, z_exp, p_exp uint64
	var true_p_exp int
	var C1, C2, C3 BID_UINT128

	// NaN handling
	if (y.w[1] & MASK_NAN_128) == MASK_NAN_128 {
		if ((y.w[1] & 0x00003fffffffffff) > 0x0000314dc6448d93) ||
			((y.w[1]&0x00003fffffffffff) == 0x0000314dc6448d93 && y.w[0] > 0x38c15b09ffffffff) {
			y.w[1] = y.w[1] & 0xffffc00000000000
			y.w[0] = 0
		}
		if (y.w[1] & MASK_SNAN_128) == MASK_SNAN_128 {
			*pfpsf |= BID_INVALID_EXCEPTION
			res.w[1] = y.w[1] & 0xfc003fffffffffff
			res.w[0] = y.w[0]
		} else {
			res.w[1] = y.w[1] & 0xfc003fffffffffff
			res.w[0] = y.w[0]
			if (z.w[1]&MASK_SNAN_128) == MASK_SNAN_128 || (x.w[1]&MASK_SNAN_128) == MASK_SNAN_128 {
				*pfpsf |= BID_INVALID_EXCEPTION
			}
		}
		return
	} else if (z.w[1] & MASK_NAN_128) == MASK_NAN_128 {
		if ((z.w[1] & 0x00003fffffffffff) > 0x0000314dc6448d93) ||
			((z.w[1]&0x00003fffffffffff) == 0x0000314dc6448d93 && z.w[0] > 0x38c15b09ffffffff) {
			z.w[1] = z.w[1] & 0xffffc00000000000
			z.w[0] = 0
		}
		if (z.w[1] & MASK_SNAN_128) == MASK_SNAN_128 {
			*pfpsf |= BID_INVALID_EXCEPTION
			res.w[1] = z.w[1] & 0xfc003fffffffffff
			res.w[0] = z.w[0]
		} else {
			res.w[1] = z.w[1] & 0xfc003fffffffffff
			res.w[0] = z.w[0]
			if (x.w[1] & MASK_SNAN_128) == MASK_SNAN_128 {
				*pfpsf |= BID_INVALID_EXCEPTION
			}
		}
		return
	} else if (x.w[1] & MASK_NAN_128) == MASK_NAN_128 {
		if ((x.w[1] & 0x00003fffffffffff) > 0x0000314dc6448d93) ||
			((x.w[1]&0x00003fffffffffff) == 0x0000314dc6448d93 && x.w[0] > 0x38c15b09ffffffff) {
			x.w[1] = x.w[1] & 0xffffc00000000000
			x.w[0] = 0
		}
		if (x.w[1] & MASK_SNAN_128) == MASK_SNAN_128 {
			*pfpsf |= BID_INVALID_EXCEPTION
			res.w[1] = x.w[1] & 0xfc003fffffffffff
			res.w[0] = x.w[0]
		} else {
			res.w[1] = x.w[1] & 0xfc003fffffffffff
			res.w[0] = x.w[0]
		}
		return
	}

	// Unpack x, y, z - check for non-canonical values
	x_sign = x.w[1] & MASK_SIGN_128
	C1.w[1] = x.w[1] & MASK_COEFF_128
	C1.w[0] = x.w[0]
	if (x.w[1] & MASK_ANY_INF_128) != MASK_INF_128 {
		if (x.w[1] & 0x6000000000000000) == 0x6000000000000000 {
			x_exp = (x.w[1] << 2) & MASK_EXP_128
			C1.w[1] = 0
			C1.w[0] = 0
		} else {
			x_exp = x.w[1] & MASK_EXP_128
			if C1.w[1] > 0x0001ed09bead87c0 || (C1.w[1] == 0x0001ed09bead87c0 && C1.w[0] > 0x378d8e63ffffffff) {
				C1.w[1] = 0
				C1.w[0] = 0
			}
		}
	}
	y_sign = y.w[1] & MASK_SIGN_128
	C2.w[1] = y.w[1] & MASK_COEFF_128
	C2.w[0] = y.w[0]
	if (y.w[1] & MASK_ANY_INF_128) != MASK_INF_128 {
		if (y.w[1] & 0x6000000000000000) == 0x6000000000000000 {
			y_exp = (y.w[1] << 2) & MASK_EXP_128
			C2.w[1] = 0
			C2.w[0] = 0
		} else {
			y_exp = y.w[1] & MASK_EXP_128
			if C2.w[1] > 0x0001ed09bead87c0 || (C2.w[1] == 0x0001ed09bead87c0 && C2.w[0] > 0x378d8e63ffffffff) {
				C2.w[1] = 0
				C2.w[0] = 0
			}
		}
	}
	z_sign = z.w[1] & MASK_SIGN_128
	C3.w[1] = z.w[1] & MASK_COEFF_128
	C3.w[0] = z.w[0]
	if (z.w[1] & MASK_ANY_INF_128) != MASK_INF_128 {
		if (z.w[1] & 0x6000000000000000) == 0x6000000000000000 {
			z_exp = (z.w[1] << 2) & MASK_EXP_128
			C3.w[1] = 0
			C3.w[0] = 0
		} else {
			z_exp = z.w[1] & MASK_EXP_128
			if C3.w[1] > 0x0001ed09bead87c0 || (C3.w[1] == 0x0001ed09bead87c0 && C3.w[0] > 0x378d8e63ffffffff) {
				C3.w[1] = 0
				C3.w[0] = 0
			}
		}
	}

	p_sign = x_sign ^ y_sign

	// Infinity handling
	if (x.w[1] & MASK_ANY_INF_128) == MASK_INF_128 { // x = inf
		if (y.w[1] & MASK_ANY_INF_128) == MASK_INF_128 { // y = inf
			if (z.w[1] & MASK_ANY_INF_128) == MASK_INF_128 { // z = inf
				if p_sign == z_sign {
					res.w[1] = z_sign | MASK_INF_128
					res.w[0] = 0
				} else {
					res.w[1] = 0x7c00000000000000
					res.w[0] = 0
					*pfpsf |= BID_INVALID_EXCEPTION
				}
			} else {
				res.w[1] = p_sign | MASK_INF_128
				res.w[0] = 0
			}
		} else if C2.w[1] != 0 || C2.w[0] != 0 { // y = f
			if (z.w[1] & MASK_ANY_INF_128) == MASK_INF_128 {
				if p_sign == z_sign {
					res.w[1] = z_sign | MASK_INF_128
					res.w[0] = 0
				} else {
					res.w[1] = 0x7c00000000000000
					res.w[0] = 0
					*pfpsf |= BID_INVALID_EXCEPTION
				}
			} else {
				res.w[1] = p_sign | MASK_INF_128
				res.w[0] = 0
			}
		} else { // y = 0
			res.w[1] = 0x7c00000000000000
			res.w[0] = 0
			*pfpsf |= BID_INVALID_EXCEPTION
		}
		return
	} else if (y.w[1] & MASK_ANY_INF_128) == MASK_INF_128 { // y = inf
		if (z.w[1] & MASK_ANY_INF_128) == MASK_INF_128 {
			if (p_sign != z_sign) || (C1.w[1] == 0 && C1.w[0] == 0) {
				res.w[1] = 0x7c00000000000000
				res.w[0] = 0
				*pfpsf |= BID_INVALID_EXCEPTION
			} else {
				res.w[1] = z_sign | MASK_INF_128
				res.w[0] = 0
			}
		} else if C1.w[1] == 0 && C1.w[0] == 0 {
			res.w[1] = 0x7c00000000000000
			res.w[0] = 0
			*pfpsf |= BID_INVALID_EXCEPTION
		} else {
			res.w[1] = p_sign | MASK_INF_128
			res.w[0] = 0
		}
		return
	} else if (z.w[1] & MASK_ANY_INF_128) == MASK_INF_128 { // z = inf
		res.w[1] = z_sign | MASK_INF_128
		res.w[0] = 0
		return
	}

	// Compute p_exp
	true_p_exp = int(x_exp>>49) - 6176 + int(y_exp>>49) - 6176
	if true_p_exp < -6176 {
		p_exp = 0
	} else {
		p_exp = uint64(true_p_exp+6176) << 49
	}

	// (x=0 or y=0) and z=0
	if ((C1.w[1] == 0 && C1.w[0] == 0) || (C2.w[1] == 0 && C2.w[0] == 0)) && C3.w[1] == 0 && C3.w[0] == 0 {
		if p_exp < z_exp {
			res.w[1] = p_exp
		} else {
			res.w[1] = z_exp
		}
		if p_sign == z_sign {
			res.w[1] |= z_sign
			res.w[0] = 0
		} else {
			if rnd_mode == BID_ROUNDING_DOWN {
				res.w[1] |= MASK_SIGN_128
				res.w[0] = 0
			} else {
				res.w[0] = 0
			}
		}
		return
	}

	// Count decimal digits in C1, C2, C3
	var q1, q2, q3, q4 int
	var e1, e2, e3, e4 int
	var C4 BID_UINT256
	var scale, ind int
	if C1.w[1] != 0 || C1.w[0] != 0 {
		q1 = bid128_count_digits(C1)
	}
	if C2.w[1] != 0 || C2.w[0] != 0 {
		q2 = bid128_count_digits(C2)
	}
	if C3.w[1] != 0 || C3.w[0] != 0 {
		q3 = bid128_count_digits(C3)
	}

	// x = 0 or y = 0, z = f (non-zero)
	if (C1.w[1] == 0 && C1.w[0] == 0) || (C2.w[1] == 0 && C2.w[0] == 0) {
		p34 := P34
		if z_exp <= p_exp {
			res.w[1] = z_sign | (z_exp & MASK_EXP_128) | C3.w[1]
			res.w[0] = C3.w[0]
		} else {
			scale = p34 - q3
			ind = int((z_exp - p_exp) >> 49)
			if ind < scale {
				scale = ind
			}
			if scale == 0 {
				res.w[1] = z.w[1]
				res.w[0] = z.w[0]
			} else if q3 <= 19 {
				if scale <= 19 {
					res = __mul_64x64_to_128(C3.w[0], bid_ten2k64[scale])
				} else {
					_, res = __mul_64x128_full(C3.w[0], bid_ten2k128[scale-20])
				}
			} else {
				_, res = __mul_64x128_full(bid_ten2k64[scale], C3)
			}
			z_exp = z_exp - (uint64(scale) << 49)
			res.w[1] = z_sign | (z_exp & MASK_EXP_128) | res.w[1]
		}
		return
	}

	e1 = int(x_exp>>49) - 6176
	e2 = int(y_exp>>49) - 6176
	e3 = int(z_exp>>49) - 6176
	e4 = e1 + e2

	// Calculate C4 = C1 * C2
	C4.w[3] = 0
	C4.w[2] = 0
	C4.w[1] = 0
	C4.w[0] = 0

	if q1+q2 <= 19 {
		C4.w[0] = C1.w[0] * C2.w[0]
		if C4.w[0] < bid_ten2k64[q1+q2-1] {
			q4 = q1 + q2 - 1
		} else {
			q4 = q1 + q2
		}
	} else if q1+q2 == 20 {
		tmp128 := __mul_64x64_to_128(C1.w[0], C2.w[0])
		C4.w[0] = tmp128.w[0]
		C4.w[1] = tmp128.w[1]
		if C4.w[1] == 0 && C4.w[0] < bid_ten2k64[19] {
			q4 = 19
		} else {
			q4 = 20
		}
	} else if q1+q2 <= 38 {
		var tmp128 BID_UINT128
		if q1 <= 19 {
			_, tmp128 = __mul_64x128_full(C1.w[0], C2)
		} else {
			_, tmp128 = __mul_64x128_full(C2.w[0], C1)
		}
		C4.w[0] = tmp128.w[0]
		C4.w[1] = tmp128.w[1]
		if C4.w[1] < bid_ten2k128[q1+q2-21].w[1] ||
			(C4.w[1] == bid_ten2k128[q1+q2-21].w[1] && C4.w[0] < bid_ten2k128[q1+q2-21].w[0]) {
			q4 = q1 + q2 - 1
		} else {
			q4 = q1 + q2
		}
	} else if q1+q2 == 39 {
		C4 = __mul_128x128_to_256(C1, C2)
		if C4.w[2] == 0 && (C4.w[1] < bid_ten2k128[18].w[1] ||
			(C4.w[1] == bid_ten2k128[18].w[1] && C4.w[0] < bid_ten2k128[18].w[0])) {
			q4 = 38
		} else {
			q4 = 39
		}
	} else if q1+q2 <= 57 { // 40 <= q1+q2 <= 57
		C4 = __mul_128x128_to_256(C1, C2)
		if C4.w[2] < bid_ten2k256[q1+q2-40].w[2] ||
			(C4.w[2] == bid_ten2k256[q1+q2-40].w[2] &&
				(C4.w[1] < bid_ten2k256[q1+q2-40].w[1] ||
					(C4.w[1] == bid_ten2k256[q1+q2-40].w[1] &&
						C4.w[0] < bid_ten2k256[q1+q2-40].w[0]))) {
			q4 = q1 + q2 - 1
		} else {
			q4 = q1 + q2
		}
	} else if q1+q2 == 58 {
		C4 = __mul_128x128_to_256(C1, C2)
		if C4.w[3] == 0 && (C4.w[2] < bid_ten2k256[18].w[2] ||
			(C4.w[2] == bid_ten2k256[18].w[2] &&
				(C4.w[1] < bid_ten2k256[18].w[1] ||
					(C4.w[1] == bid_ten2k256[18].w[1] &&
						C4.w[0] < bid_ten2k256[18].w[0])))) {
			q4 = 57
		} else {
			q4 = 58
		}
	} else { // 59 <= q1+q2 <= 68
		C4 = __mul_128x128_to_256(C1, C2)
		if C4.w[3] < bid_ten2k256[q1+q2-40].w[3] ||
			(C4.w[3] == bid_ten2k256[q1+q2-40].w[3] &&
				(C4.w[2] < bid_ten2k256[q1+q2-40].w[2] ||
					(C4.w[2] == bid_ten2k256[q1+q2-40].w[2] &&
						(C4.w[1] < bid_ten2k256[q1+q2-40].w[1] ||
							(C4.w[1] == bid_ten2k256[q1+q2-40].w[1] &&
								C4.w[0] < bid_ten2k256[q1+q2-40].w[0]))))) {
			q4 = q1 + q2 - 1
		} else {
			q4 = q1 + q2
		}
	}

	var save_fpsf uint32
	var is_midpoint_lt_even0, is_midpoint_gt_even0 int
	var is_inexact_lt_midpoint0, is_inexact_gt_midpoint0 int
	var incr_exp int
	var lt_half_ulp, eq_half_ulp int
	var is_tiny int
	var R64 uint64
	var P128 BID_UINT128
	var P192, R192 BID_UINT192
	var R256 BID_UINT256
	var x0 int
	var p34 = P34

	if C3.w[1] == 0x0 && C3.w[0] == 0x0 { // x = f, y = f, z = 0
		save_fpsf = *pfpsf
		*pfpsf = 0

		if q4 > p34 {
			// truncate C4 to p34 digits into res
			x0 = q4 - p34
			if q4 <= 38 {
				P128.w[1] = C4.w[1]
				P128.w[0] = C4.w[0]
				res, incr_exp, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint = bid_round128_19_38(q4, x0, P128)
			} else if q4 <= 57 {
				P192.w[2] = C4.w[2]
				P192.w[1] = C4.w[1]
				P192.w[0] = C4.w[0]
				R192, incr_exp, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint = bid_round192_39_57(q4, x0, P192)
				res.w[0] = R192.w[0]
				res.w[1] = R192.w[1]
			} else { // if q4 <= 68
				R256, incr_exp, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint = bid_round256_58_76(q4, x0, C4)
				res.w[0] = R256.w[0]
				res.w[1] = R256.w[1]
			}
			e4 = e4 + x0
			q4 = p34
			if incr_exp != 0 {
				e4 = e4 + 1
				// DECIMAL_TINY_DETECTION_AFTER_ROUNDING = 0 => before rounding
				if q4+e4 == expmin128+p34 {
					*pfpsf |= (BID_INEXACT_EXCEPTION | BID_UNDERFLOW_EXCEPTION)
				}
			}
		} else { // if q4 <= p34
			if (q4+e4 <= p34+expmax) && (e4 > expmax) {
				scale = e4 - expmax
				if q4 <= 19 {
					if scale <= 19 {
						res = __mul_64x64_to_128(C4.w[0], bid_ten2k64[scale])
					} else {
						res = __mul_64x128_to_128(C4.w[0], bid_ten2k128[scale-20])
					}
				} else {
					res = __mul_64x128_to_128(bid_ten2k64[scale], BID_UINT128{w: [2]uint64{C4.w[0], C4.w[1]}})
				}
				e4 = e4 - scale
				q4 = q4 + scale
			} else {
				res.w[1] = C4.w[1]
				res.w[0] = C4.w[0]
			}
		}

		// check for overflow
		if q4+e4 > p34+expmax {
			if rnd_mode == BID_ROUNDING_TO_NEAREST {
				res.w[1] = p_sign | 0x7800000000000000
				res.w[0] = 0x0000000000000000
				*pfpsf |= (BID_INEXACT_EXCEPTION | BID_OVERFLOW_EXCEPTION)
			} else {
				res.w[1] = p_sign | res.w[1]
				bid_rounding_correction(rnd_mode,
					is_inexact_lt_midpoint,
					is_inexact_gt_midpoint,
					is_midpoint_lt_even, is_midpoint_gt_even,
					e4, &res, pfpsf)
			}
			*pfpsf |= save_fpsf
			return
		}

		// check for underflow
		if q4+e4 < expmin128+p34 {
			is_tiny = 1
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
				if x0 < q4 {
					if q4 <= 18 {
						R64 = bid_round64_2_18(q4, x0, res.w[0], &incr_exp,
							&is_midpoint_lt_even, &is_midpoint_gt_even,
							&is_inexact_lt_midpoint,
							&is_inexact_gt_midpoint)
						if incr_exp != 0 {
							R64 = bid_ten2k64[q4-x0]
						}
						res.w[0] = R64
					} else { // q4 <= 34
						P128.w[1] = res.w[1]
						P128.w[0] = res.w[0]
						res, incr_exp, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint = bid_round128_19_38(q4, x0, P128)
						if incr_exp != 0 {
							if q4-x0 <= 19 {
								res.w[0] = bid_ten2k64[q4-x0]
							} else {
								res.w[0] = bid_ten2k128[q4-x0-20].w[0]
								res.w[1] = bid_ten2k128[q4-x0-20].w[1]
							}
						}
					}
					e4 = e4 + x0
				} else if x0 == q4 {
					if q4 <= 19 {
						if res.w[0] < bid_midpoint64[q4-1] {
							lt_half_ulp = 1
							is_inexact_lt_midpoint = 1
						} else if res.w[0] == bid_midpoint64[q4-1] {
							eq_half_ulp = 1
							is_midpoint_gt_even = 1
						} else {
							is_inexact_gt_midpoint = 1
						}
					} else { // q4 <= 34
						if res.w[1] < bid_midpoint128[q4-20].w[1] ||
							(res.w[1] == bid_midpoint128[q4-20].w[1] &&
								res.w[0] < bid_midpoint128[q4-20].w[0]) {
							lt_half_ulp = 1
							is_inexact_lt_midpoint = 1
						} else if res.w[1] == bid_midpoint128[q4-20].w[1] &&
							res.w[0] == bid_midpoint128[q4-20].w[0] {
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
					e4 = expmin128
				} else { // x0 > q4
					res.w[1] = 0
					res.w[0] = 0
					e4 = expmin128
					is_inexact_lt_midpoint = 1
				}
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
				} else {
					// leave as is
				}
			} else { // if e4 >= emin then q4 < P and the result is tiny and exact
				if e3 < e4 {
					scale = p34 - q4
					ind = e4 - e3
					if ind < scale {
						scale = ind
					}
					if scale == 0 {
						// res and e4 are unchanged
					} else if q4 <= 19 {
						if scale <= 19 {
							res = __mul_64x64_to_128(res.w[0], bid_ten2k64[scale])
						} else {
							res = __mul_64x128_to_128(res.w[0], bid_ten2k128[scale-20])
						}
					} else {
						res = __mul_64x128_to_128(bid_ten2k64[scale], res)
					}
					e4 = e4 - scale
				}
			}

			// check for inexact result
			if is_inexact_lt_midpoint != 0 || is_inexact_gt_midpoint != 0 ||
				is_midpoint_lt_even != 0 || is_midpoint_gt_even != 0 {
				*pfpsf |= BID_INEXACT_EXCEPTION
				*pfpsf |= BID_UNDERFLOW_EXCEPTION
			}
			res.w[1] = p_sign | (uint64(e4+6176) << 49) | res.w[1]
			if rnd_mode != BID_ROUNDING_TO_NEAREST {
				bid_rounding_correction(rnd_mode,
					is_inexact_lt_midpoint,
					is_inexact_gt_midpoint,
					is_midpoint_lt_even, is_midpoint_gt_even,
					e4, &res, pfpsf)
			}
			*pfpsf |= save_fpsf
			return
		}

		// no overflow, and no underflow for rounding to nearest
		res.w[1] = p_sign | (uint64(e4+6176) << 49) | res.w[1]
		if rnd_mode != BID_ROUNDING_TO_NEAREST {
			bid_rounding_correction(rnd_mode,
				is_inexact_lt_midpoint,
				is_inexact_gt_midpoint,
				is_midpoint_lt_even, is_midpoint_gt_even,
				e4, &res, pfpsf)
			if e4 == expmin128 {
				if (res.w[1]&MASK_COEFF128) < 0x0000314dc6448d93 ||
					((res.w[1]&MASK_COEFF128) == 0x0000314dc6448d93 &&
						res.w[0] < 0x38c15b0a00000000) {
					is_tiny = 1
				}
			}
		}

		if is_inexact_lt_midpoint != 0 || is_inexact_gt_midpoint != 0 ||
			is_midpoint_lt_even != 0 || is_midpoint_gt_even != 0 {
			*pfpsf |= BID_INEXACT_EXCEPTION
			if is_tiny != 0 {
				*pfpsf |= BID_UNDERFLOW_EXCEPTION
			}
		}

		if (*pfpsf & BID_INEXACT_EXCEPTION) == 0 { // x * y is exact
			p_exp = res.w[1] & MASK_EXP_128
			if z_exp < p_exp {
				C3.w[1] = res.w[1] & MASK_COEFF128
				C3.w[0] = res.w[0]
				scale = p34 - q4
				ind = int((p_exp - z_exp) >> 49)
				if ind < scale {
					scale = ind
				}
				p_exp = p_exp - (uint64(scale) << 49)
				if scale == 0 {
					// leave res unchanged
				} else if q4 <= 19 {
					if scale <= 19 {
						res = __mul_64x64_to_128(C3.w[0], bid_ten2k64[scale])
					} else {
						res = __mul_64x128_to_128(C3.w[0], bid_ten2k128[scale-20])
					}
					res.w[1] = p_sign | (p_exp & MASK_EXP_128) | res.w[1]
				} else {
					res = __mul_64x128_to_128(bid_ten2k64[scale], C3)
					res.w[1] = p_sign | (p_exp & MASK_EXP_128) | res.w[1]
				}
			}
		}
		*pfpsf |= save_fpsf
		return
	} // else we have f * f + f

	// continue with x = f, y = f, z = f
	delta := q3 + e3 - q4 - e4

	// The main body calls bid_fma_main_body which handles delta >= 0 and delta < 0
	bid_fma_main_body(p34, &res, &is_midpoint_lt_even, &is_midpoint_gt_even,
		&is_inexact_lt_midpoint, &is_inexact_gt_midpoint,
		p_sign, z_sign, &z_exp, &p_exp,
		q3, q4, &e3, &e4, delta,
		&C3, C4, rnd_mode, pfpsf)
	return
}

// Bid128Fma computes x*y+z for Decimal128.
// Ported from bid128_fma.c: bid128_fma which delegates to bid128_ext_fma.
func Bid128Fma(x, y, z BID_UINT128, rnd_mode int) (BID_UINT128, uint32) {
	var pfpsf uint32
	res, _, _, _, _ := bid128_ext_fma(x, y, z, rnd_mode, &pfpsf)
	return res, pfpsf
}
