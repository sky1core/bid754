// Ported from: Intel bid128_to_int32.c, bid128_to_int64.c,
//              bid128_to_uint32.c, bid128_to_uint64.c,
//              bid128_to_int8.c, bid128_to_int16.c,
//              bid128_to_uint8.c, bid128_to_uint16.c,
//              bid128_llrintd.c, bid128_lrintd.c,
//              bid128_llround.c, bid128_lround.c
// Mechanical translation - all logic preserved exactly.

package bidgo

import "math"

// ============================================================================
// Common helpers for bid128 -> integer conversions
// ============================================================================

// bid128_unpack_for_int unpacks a BID128 for integer conversion.
// Returns (x_sign, x_exp, C1, pfpsf, special).
// If special is true, pfpsf and returned int indicate the result should be returned immediately.
func bid128_unpack_for_int(x BID_UINT128) (x_sign uint64, x_exp uint64, C1 BID_UINT128, is_special bool) {
	x_sign = x.w[1] & MASK_SIGN64
	x_exp = x.w[1] & MASK_EXP128
	C1.w[1] = x.w[1] & MASK_COEFF128
	C1.w[0] = x.w[0]
	is_special = (x.w[1] & MASK_SPECIAL128) == MASK_SPECIAL128
	return
}

// bid128_is_nan_for_int checks if x is NaN for integer conversion purposes.
func bid128_is_nan_for_int(x BID_UINT128) bool {
	return (x.w[1] & NAN_MASK64) == NAN_MASK64
}

// bid128_is_snan_for_int checks if x is SNaN for integer conversion purposes.
func bid128_is_snan_for_int(x BID_UINT128) bool {
	return (x.w[1] & SNAN_MASK64) == SNAN_MASK64
}

// bid128_is_noncanonical checks for non-canonical values.
func bid128_is_noncanonical(C1 BID_UINT128, x BID_UINT128) bool {
	return (C1.w[1] > 0x0001ed09bead87c0) ||
		(C1.w[1] == 0x0001ed09bead87c0 && C1.w[0] > 0x378d8e63ffffffff) ||
		((x.w[1] & 0x6000000000000000) == 0x6000000000000000)
}

// bid128_nr_digits computes q = nr. of decimal digits in 128-bit coefficient C1.
func bid128_nr_digits(C1 BID_UINT128) (q int, x_nr_bits uint) {
	var tmp1 uint64
	if C1.w[1] == 0 {
		if C1.w[0] >= 0x0020000000000000 { // x >= 2^53
			tmp1 = math.Float64bits(float64(C1.w[0] >> 32))
			x_nr_bits = 33 + uint((uint32(tmp1>>52)&0x7ff)-0x3ff)
		} else {
			tmp1 = math.Float64bits(float64(C1.w[0]))
			x_nr_bits = 1 + uint((uint32(tmp1>>52)&0x7ff)-0x3ff)
		}
	} else {
		tmp1 = math.Float64bits(float64(C1.w[1]))
		x_nr_bits = 65 + uint((uint32(tmp1>>52)&0x7ff)-0x3ff)
	}
	q = int(bid_nr_digits[x_nr_bits-1].digits)
	if q == 0 {
		q = int(bid_nr_digits[x_nr_bits-1].digits1)
		if C1.w[1] > bid_nr_digits[x_nr_bits-1].threshold_hi ||
			(C1.w[1] == bid_nr_digits[x_nr_bits-1].threshold_hi &&
				C1.w[0] >= bid_nr_digits[x_nr_bits-1].threshold_lo) {
			q++
		}
	}
	return
}

// bid128_check_overflow_int32 checks if the boundary condition for (q+exp)==10
// overflow is met. Returns true if invalid.
// cmp_neg_gt: comparison for negative side (>, >=)
// cmp_pos_ge: comparison for positive side (>=, >)
func bid128_check_overflow_10(C1 BID_UINT128, x_sign uint64, q int,
	neg_limit uint64, neg_cmp_ge bool,
	pos_limit uint64, pos_cmp_ge bool) (bool, uint32) {
	var C BID_UINT128
	var tmp64 uint64
	var pfpsf uint32

	if x_sign != 0 { // if n < 0 and q + exp = 10
		if q <= 11 {
			tmp64 = C1.w[0] * bid_ten2k64[11-q]
			if neg_cmp_ge {
				if tmp64 >= neg_limit {
					pfpsf |= BID_INVALID_EXCEPTION
					return true, pfpsf
				}
			} else {
				if tmp64 > neg_limit {
					pfpsf |= BID_INVALID_EXCEPTION
					return true, pfpsf
				}
			}
		} else {
			tmp64 = neg_limit
			if q-11 <= 19 {
				C = __mul_64x64_to_128(tmp64, bid_ten2k64[q-11])
			} else {
				C = __mul_128x64_to_128(tmp64, bid_ten2k128[q-31])
			}
			if neg_cmp_ge {
				if C1.w[1] > C.w[1] || (C1.w[1] == C.w[1] && C1.w[0] >= C.w[0]) {
					pfpsf |= BID_INVALID_EXCEPTION
					return true, pfpsf
				}
			} else {
				if C1.w[1] > C.w[1] || (C1.w[1] == C.w[1] && C1.w[0] > C.w[0]) {
					pfpsf |= BID_INVALID_EXCEPTION
					return true, pfpsf
				}
			}
		}
	} else { // if n > 0
		if q <= 11 {
			tmp64 = C1.w[0] * bid_ten2k64[11-q]
			if pos_cmp_ge {
				if tmp64 >= pos_limit {
					pfpsf |= BID_INVALID_EXCEPTION
					return true, pfpsf
				}
			} else {
				if tmp64 > pos_limit {
					pfpsf |= BID_INVALID_EXCEPTION
					return true, pfpsf
				}
			}
		} else {
			tmp64 = pos_limit
			if q-11 <= 19 {
				C = __mul_64x64_to_128(tmp64, bid_ten2k64[q-11])
			} else {
				C = __mul_128x64_to_128(tmp64, bid_ten2k128[q-31])
			}
			if pos_cmp_ge {
				if C1.w[1] > C.w[1] || (C1.w[1] == C.w[1] && C1.w[0] >= C.w[0]) {
					pfpsf |= BID_INVALID_EXCEPTION
					return true, pfpsf
				}
			} else {
				if C1.w[1] > C.w[1] || (C1.w[1] == C.w[1] && C1.w[0] > C.w[0]) {
					pfpsf |= BID_INVALID_EXCEPTION
					return true, pfpsf
				}
			}
		}
	}
	return false, 0
}

// bid128_round_rnint_common rounds C1 to integer using round-to-nearest-even.
// Returns (Cstar_w0, is_midpoint bool, fstar BID_UINT256, P256 BID_UINT256)
func bid128_round_rnint_common(C1 BID_UINT128, ind int) (Cstar_w0 uint64) {
	var Cstar BID_UINT128
	var fstar BID_UINT256
	var P256 BID_UINT256

	// chop off ind digits from the lower part of C1
	// C1 = C1 + 1/2 * 10^ind where the result C1 fits in 127 bits
	tmp64 := C1.w[0]
	if ind <= 19 {
		C1.w[0] = C1.w[0] + bid_midpoint64[ind-1]
	} else {
		C1.w[0] = C1.w[0] + bid_midpoint128[ind-20].w[0]
		C1.w[1] = C1.w[1] + bid_midpoint128[ind-20].w[1]
	}
	if C1.w[0] < tmp64 {
		C1.w[1]++
	}
	// calculate C* and f*
	P256 = __mul_128x128_to_256(C1, bid_ten2mk128[ind-1])
	if ind-1 <= 21 {
		Cstar.w[1] = P256.w[3]
		Cstar.w[0] = P256.w[2]
		fstar.w[3] = 0
		fstar.w[2] = P256.w[2] & bid_maskhigh128[ind-1]
		fstar.w[1] = P256.w[1]
		fstar.w[0] = P256.w[0]
	} else {
		Cstar.w[1] = 0
		Cstar.w[0] = P256.w[3]
		fstar.w[3] = P256.w[3] & bid_maskhigh128[ind-1]
		fstar.w[2] = P256.w[2]
		fstar.w[1] = P256.w[1]
		fstar.w[0] = P256.w[0]
	}

	// shift right C* by Ex-128 = bid_shiftright128[ind]
	shift := bid_shiftright128[ind-1]
	if ind-1 <= 21 {
		Cstar.w[0] = (Cstar.w[0] >> uint(shift)) | (Cstar.w[1] << uint(64-shift))
	} else {
		Cstar.w[0] = Cstar.w[0] >> uint(shift-64)
	}
	// check for midpoints
	if (fstar.w[3] == 0) && (fstar.w[2] == 0) &&
		(fstar.w[1] != 0 || fstar.w[0] != 0) &&
		(fstar.w[1] < bid_ten2mk128trunc[ind-1].w[1] ||
			(fstar.w[1] == bid_ten2mk128trunc[ind-1].w[1] &&
				fstar.w[0] <= bid_ten2mk128trunc[ind-1].w[0])) {
		// the result is a midpoint; round to nearest
		if Cstar.w[0]&0x01 != 0 {
			Cstar.w[0]--
		}
	}
	return Cstar.w[0]
}

// bid128_round_floor_ceil_int_common rounds C1 with direction-aware rounding.
// mode: 0=floor, 1=ceil, 2=int(truncate)
func bid128_round_floor_ceil_int_common(C1 BID_UINT128, ind int, x_sign uint64, mode int) (Cstar_w0 uint64) {
	var Cstar BID_UINT128
	var fstar BID_UINT256
	var P256 BID_UINT256
	var is_inexact_lt_midpoint int
	var is_inexact_gt_midpoint int
	var is_midpoint_lt_even int
	var is_midpoint_gt_even int
	_ = is_midpoint_gt_even

	// chop off ind digits from the lower part of C1
	tmp64 := C1.w[0]
	if ind <= 19 {
		C1.w[0] = C1.w[0] + bid_midpoint64[ind-1]
	} else {
		C1.w[0] = C1.w[0] + bid_midpoint128[ind-20].w[0]
		C1.w[1] = C1.w[1] + bid_midpoint128[ind-20].w[1]
	}
	if C1.w[0] < tmp64 {
		C1.w[1]++
	}
	P256 = __mul_128x128_to_256(C1, bid_ten2mk128[ind-1])
	if ind-1 <= 21 {
		Cstar.w[1] = P256.w[3]
		Cstar.w[0] = P256.w[2]
		fstar.w[3] = 0
		fstar.w[2] = P256.w[2] & bid_maskhigh128[ind-1]
		fstar.w[1] = P256.w[1]
		fstar.w[0] = P256.w[0]
	} else {
		Cstar.w[1] = 0
		Cstar.w[0] = P256.w[3]
		fstar.w[3] = P256.w[3] & bid_maskhigh128[ind-1]
		fstar.w[2] = P256.w[2]
		fstar.w[1] = P256.w[1]
		fstar.w[0] = P256.w[0]
	}

	shift := bid_shiftright128[ind-1]
	if ind-1 <= 21 {
		Cstar.w[0] = (Cstar.w[0] >> uint(shift)) | (Cstar.w[1] << uint(64-shift))
	} else {
		Cstar.w[0] = Cstar.w[0] >> uint(shift-64)
	}

	// determine inexactness
	var tmp64A uint64
	if ind-1 <= 2 {
		if fstar.w[1] > 0x8000000000000000 || (fstar.w[1] == 0x8000000000000000 && fstar.w[0] > 0x0) {
			tmp64 = fstar.w[1] - 0x8000000000000000
			if tmp64 > bid_ten2mk128trunc[ind-1].w[1] ||
				(tmp64 == bid_ten2mk128trunc[ind-1].w[1] &&
					fstar.w[0] >= bid_ten2mk128trunc[ind-1].w[0]) {
				is_inexact_lt_midpoint = 1
			}
		} else {
			is_inexact_gt_midpoint = 1
		}
	} else if ind-1 <= 21 {
		if fstar.w[3] > 0x0 ||
			(fstar.w[3] == 0x0 && fstar.w[2] > bid_onehalf128[ind-1]) ||
			(fstar.w[3] == 0x0 && fstar.w[2] == bid_onehalf128[ind-1] &&
				(fstar.w[1] != 0 || fstar.w[0] != 0)) {
			tmp64 = fstar.w[2] - bid_onehalf128[ind-1]
			tmp64A = fstar.w[3]
			if tmp64 > fstar.w[2] {
				tmp64A--
			}
			if tmp64A != 0 || tmp64 != 0 ||
				fstar.w[1] > bid_ten2mk128trunc[ind-1].w[1] ||
				(fstar.w[1] == bid_ten2mk128trunc[ind-1].w[1] &&
					fstar.w[0] > bid_ten2mk128trunc[ind-1].w[0]) {
				is_inexact_lt_midpoint = 1
			}
		} else {
			is_inexact_gt_midpoint = 1
		}
	} else {
		if fstar.w[3] > bid_onehalf128[ind-1] ||
			(fstar.w[3] == bid_onehalf128[ind-1] &&
				(fstar.w[2] != 0 || fstar.w[1] != 0 || fstar.w[0] != 0)) {
			tmp64 = fstar.w[3] - bid_onehalf128[ind-1]
			if tmp64 != 0 || fstar.w[2] != 0 ||
				fstar.w[1] > bid_ten2mk128trunc[ind-1].w[1] ||
				(fstar.w[1] == bid_ten2mk128trunc[ind-1].w[1] &&
					fstar.w[0] > bid_ten2mk128trunc[ind-1].w[0]) {
				is_inexact_lt_midpoint = 1
			}
		} else {
			is_inexact_gt_midpoint = 1
		}
	}

	// check for midpoints
	if (fstar.w[3] == 0) && (fstar.w[2] == 0) &&
		(fstar.w[1] != 0 || fstar.w[0] != 0) &&
		(fstar.w[1] < bid_ten2mk128trunc[ind-1].w[1] ||
			(fstar.w[1] == bid_ten2mk128trunc[ind-1].w[1] &&
				fstar.w[0] <= bid_ten2mk128trunc[ind-1].w[0])) {
		if Cstar.w[0]&0x01 != 0 {
			Cstar.w[0]--
			is_midpoint_gt_even = 1
			is_inexact_lt_midpoint = 0
			is_inexact_gt_midpoint = 0
		} else {
			is_midpoint_lt_even = 1
			is_inexact_lt_midpoint = 0
			is_inexact_gt_midpoint = 0
		}
	}

	switch mode {
	case 0: // floor (RM)
		if x_sign != 0 && (is_midpoint_gt_even != 0 || is_inexact_lt_midpoint != 0) {
			Cstar.w[0] = Cstar.w[0] + 1
		} else if x_sign == 0 && (is_midpoint_lt_even != 0 || is_inexact_gt_midpoint != 0) {
			Cstar.w[0] = Cstar.w[0] - 1
		}
	case 1: // ceil (RP)
		if x_sign != 0 && (is_midpoint_lt_even != 0 || is_inexact_gt_midpoint != 0) {
			Cstar.w[0] = Cstar.w[0] - 1
		} else if x_sign == 0 && (is_midpoint_gt_even != 0 || is_inexact_lt_midpoint != 0) {
			Cstar.w[0] = Cstar.w[0] + 1
		}
	case 2: // int/truncate (RZ)
		if is_midpoint_lt_even != 0 || is_inexact_gt_midpoint != 0 {
			Cstar.w[0] = Cstar.w[0] - 1
		}
	}
	return Cstar.w[0]
}

func bid128_trunc_inexact_common(C1 BID_UINT128, ind int) (Cstar_w0 uint64, inexact bool) {
	var Cstar BID_UINT128
	var fstar BID_UINT256
	var P256 BID_UINT256

	P256 = __mul_128x128_to_256(C1, bid_ten2mk128[ind-1])
	if ind-1 <= 21 {
		Cstar.w[1] = P256.w[3]
		Cstar.w[0] = P256.w[2]
		fstar.w[3] = 0
		fstar.w[2] = P256.w[2] & bid_maskhigh128[ind-1]
		fstar.w[1] = P256.w[1]
		fstar.w[0] = P256.w[0]
	} else {
		Cstar.w[1] = 0
		Cstar.w[0] = P256.w[3]
		fstar.w[3] = P256.w[3] & bid_maskhigh128[ind-1]
		fstar.w[2] = P256.w[2]
		fstar.w[1] = P256.w[1]
		fstar.w[0] = P256.w[0]
	}

	shift := bid_shiftright128[ind-1]
	if ind-1 <= 21 {
		Cstar.w[0] = (Cstar.w[0] >> uint(shift)) | (Cstar.w[1] << uint(64-shift))
	} else {
		Cstar.w[0] = Cstar.w[0] >> uint(shift-64)
	}

	if ind-1 <= 2 {
		inexact = fstar.w[1] > bid_ten2mk128trunc[ind-1].w[1] ||
			(fstar.w[1] == bid_ten2mk128trunc[ind-1].w[1] &&
				fstar.w[0] > bid_ten2mk128trunc[ind-1].w[0])
	} else if ind-1 <= 21 {
		inexact = fstar.w[2] != 0 ||
			fstar.w[1] > bid_ten2mk128trunc[ind-1].w[1] ||
			(fstar.w[1] == bid_ten2mk128trunc[ind-1].w[1] &&
				fstar.w[0] > bid_ten2mk128trunc[ind-1].w[0])
	} else {
		inexact = fstar.w[3] != 0 || fstar.w[2] != 0 ||
			fstar.w[1] > bid_ten2mk128trunc[ind-1].w[1] ||
			(fstar.w[1] == bid_ten2mk128trunc[ind-1].w[1] &&
				fstar.w[0] > bid_ten2mk128trunc[ind-1].w[0])
	}

	return Cstar.w[0], inexact
}

func bid128_round_trunc_mode_common(C1 BID_UINT128, ind int, x_sign uint64, mode int, setInexact bool) (Cstar_w0 uint64, pfpsf uint32) {
	Cstar_w0, inexact := bid128_trunc_inexact_common(C1, ind)
	switch mode {
	case 0: // floor
		if x_sign != 0 && inexact {
			Cstar_w0++
		}
	case 1: // ceil
		if x_sign == 0 && inexact {
			Cstar_w0++
		}
	case 2: // truncate
	}
	if setInexact && inexact {
		pfpsf |= BID_INEXACT_EXCEPTION
	}
	return Cstar_w0, pfpsf
}

// bid128_round_xrnint_common rounds with INEXACT flag detection for xrnint.
func bid128_round_xrnint_common(C1 BID_UINT128, ind int) (Cstar_w0 uint64, pfpsf uint32) {
	var Cstar BID_UINT128
	var fstar BID_UINT256
	var P256 BID_UINT256

	tmp64 := C1.w[0]
	if ind <= 19 {
		C1.w[0] = C1.w[0] + bid_midpoint64[ind-1]
	} else {
		C1.w[0] = C1.w[0] + bid_midpoint128[ind-20].w[0]
		C1.w[1] = C1.w[1] + bid_midpoint128[ind-20].w[1]
	}
	if C1.w[0] < tmp64 {
		C1.w[1]++
	}
	P256 = __mul_128x128_to_256(C1, bid_ten2mk128[ind-1])
	if ind-1 <= 21 {
		Cstar.w[1] = P256.w[3]
		Cstar.w[0] = P256.w[2]
		fstar.w[3] = 0
		fstar.w[2] = P256.w[2] & bid_maskhigh128[ind-1]
		fstar.w[1] = P256.w[1]
		fstar.w[0] = P256.w[0]
	} else {
		Cstar.w[1] = 0
		Cstar.w[0] = P256.w[3]
		fstar.w[3] = P256.w[3] & bid_maskhigh128[ind-1]
		fstar.w[2] = P256.w[2]
		fstar.w[1] = P256.w[1]
		fstar.w[0] = P256.w[0]
	}

	shift := bid_shiftright128[ind-1]
	if ind-1 <= 21 {
		Cstar.w[0] = (Cstar.w[0] >> uint(shift)) | (Cstar.w[1] << uint(64-shift))
	} else {
		Cstar.w[0] = Cstar.w[0] >> uint(shift-64)
	}

	// determine inexactness
	var tmp64A uint64
	if ind-1 <= 2 {
		if fstar.w[1] > 0x8000000000000000 || (fstar.w[1] == 0x8000000000000000 && fstar.w[0] > 0x0) {
			tmp64 = fstar.w[1] - 0x8000000000000000
			if tmp64 > bid_ten2mk128trunc[ind-1].w[1] ||
				(tmp64 == bid_ten2mk128trunc[ind-1].w[1] &&
					fstar.w[0] >= bid_ten2mk128trunc[ind-1].w[0]) {
				pfpsf |= BID_INEXACT_EXCEPTION
			}
		} else {
			pfpsf |= BID_INEXACT_EXCEPTION
		}
	} else if ind-1 <= 21 {
		if fstar.w[3] > 0x0 ||
			(fstar.w[3] == 0x0 && fstar.w[2] > bid_onehalf128[ind-1]) ||
			(fstar.w[3] == 0x0 && fstar.w[2] == bid_onehalf128[ind-1] &&
				(fstar.w[1] != 0 || fstar.w[0] != 0)) {
			tmp64 = fstar.w[2] - bid_onehalf128[ind-1]
			tmp64A = fstar.w[3]
			if tmp64 > fstar.w[2] {
				tmp64A--
			}
			if tmp64A != 0 || tmp64 != 0 ||
				fstar.w[1] > bid_ten2mk128trunc[ind-1].w[1] ||
				(fstar.w[1] == bid_ten2mk128trunc[ind-1].w[1] &&
					fstar.w[0] > bid_ten2mk128trunc[ind-1].w[0]) {
				pfpsf |= BID_INEXACT_EXCEPTION
			}
		} else {
			pfpsf |= BID_INEXACT_EXCEPTION
		}
	} else {
		if fstar.w[3] > bid_onehalf128[ind-1] ||
			(fstar.w[3] == bid_onehalf128[ind-1] &&
				(fstar.w[2] != 0 || fstar.w[1] != 0 || fstar.w[0] != 0)) {
			tmp64 = fstar.w[3] - bid_onehalf128[ind-1]
			if tmp64 != 0 || fstar.w[2] != 0 ||
				fstar.w[1] > bid_ten2mk128trunc[ind-1].w[1] ||
				(fstar.w[1] == bid_ten2mk128trunc[ind-1].w[1] &&
					fstar.w[0] > bid_ten2mk128trunc[ind-1].w[0]) {
				pfpsf |= BID_INEXACT_EXCEPTION
			}
		} else {
			pfpsf |= BID_INEXACT_EXCEPTION
		}
	}

	// check for midpoints
	if (fstar.w[3] == 0) && (fstar.w[2] == 0) &&
		(fstar.w[1] != 0 || fstar.w[0] != 0) &&
		(fstar.w[1] < bid_ten2mk128trunc[ind-1].w[1] ||
			(fstar.w[1] == bid_ten2mk128trunc[ind-1].w[1] &&
				fstar.w[0] <= bid_ten2mk128trunc[ind-1].w[0])) {
		if Cstar.w[0]&0x01 != 0 {
			Cstar.w[0]--
		}
	}
	return Cstar.w[0], pfpsf
}

// bid128_round_xfloor_xceil_xint_common rounds with INEXACT flag for xfloor/xceil/xint.
// mode: 0=xfloor, 1=xceil, 2=xint
func bid128_round_xfloor_xceil_xint_common(C1 BID_UINT128, ind int, x_sign uint64, mode int) (Cstar_w0 uint64, pfpsf uint32) {
	var Cstar BID_UINT128
	var fstar BID_UINT256
	var P256 BID_UINT256
	var is_inexact_lt_midpoint int
	var is_inexact_gt_midpoint int
	var is_midpoint_lt_even int
	var is_midpoint_gt_even int
	_ = is_midpoint_gt_even

	tmp64 := C1.w[0]
	if ind <= 19 {
		C1.w[0] = C1.w[0] + bid_midpoint64[ind-1]
	} else {
		C1.w[0] = C1.w[0] + bid_midpoint128[ind-20].w[0]
		C1.w[1] = C1.w[1] + bid_midpoint128[ind-20].w[1]
	}
	if C1.w[0] < tmp64 {
		C1.w[1]++
	}
	P256 = __mul_128x128_to_256(C1, bid_ten2mk128[ind-1])
	if ind-1 <= 21 {
		Cstar.w[1] = P256.w[3]
		Cstar.w[0] = P256.w[2]
		fstar.w[3] = 0
		fstar.w[2] = P256.w[2] & bid_maskhigh128[ind-1]
		fstar.w[1] = P256.w[1]
		fstar.w[0] = P256.w[0]
	} else {
		Cstar.w[1] = 0
		Cstar.w[0] = P256.w[3]
		fstar.w[3] = P256.w[3] & bid_maskhigh128[ind-1]
		fstar.w[2] = P256.w[2]
		fstar.w[1] = P256.w[1]
		fstar.w[0] = P256.w[0]
	}

	shift := bid_shiftright128[ind-1]
	if ind-1 <= 21 {
		Cstar.w[0] = (Cstar.w[0] >> uint(shift)) | (Cstar.w[1] << uint(64-shift))
	} else {
		Cstar.w[0] = Cstar.w[0] >> uint(shift-64)
	}

	// determine inexactness
	var tmp64A uint64
	if ind-1 <= 2 {
		if fstar.w[1] > 0x8000000000000000 || (fstar.w[1] == 0x8000000000000000 && fstar.w[0] > 0x0) {
			tmp64 = fstar.w[1] - 0x8000000000000000
			if tmp64 > bid_ten2mk128trunc[ind-1].w[1] ||
				(tmp64 == bid_ten2mk128trunc[ind-1].w[1] &&
					fstar.w[0] >= bid_ten2mk128trunc[ind-1].w[0]) {
				is_inexact_lt_midpoint = 1
			}
		} else {
			is_inexact_gt_midpoint = 1
		}
	} else if ind-1 <= 21 {
		if fstar.w[3] > 0x0 ||
			(fstar.w[3] == 0x0 && fstar.w[2] > bid_onehalf128[ind-1]) ||
			(fstar.w[3] == 0x0 && fstar.w[2] == bid_onehalf128[ind-1] &&
				(fstar.w[1] != 0 || fstar.w[0] != 0)) {
			tmp64 = fstar.w[2] - bid_onehalf128[ind-1]
			tmp64A = fstar.w[3]
			if tmp64 > fstar.w[2] {
				tmp64A--
			}
			if tmp64A != 0 || tmp64 != 0 ||
				fstar.w[1] > bid_ten2mk128trunc[ind-1].w[1] ||
				(fstar.w[1] == bid_ten2mk128trunc[ind-1].w[1] &&
					fstar.w[0] > bid_ten2mk128trunc[ind-1].w[0]) {
				is_inexact_lt_midpoint = 1
			}
		} else {
			is_inexact_gt_midpoint = 1
		}
	} else {
		if fstar.w[3] > bid_onehalf128[ind-1] ||
			(fstar.w[3] == bid_onehalf128[ind-1] &&
				(fstar.w[2] != 0 || fstar.w[1] != 0 || fstar.w[0] != 0)) {
			tmp64 = fstar.w[3] - bid_onehalf128[ind-1]
			if tmp64 != 0 || fstar.w[2] != 0 ||
				fstar.w[1] > bid_ten2mk128trunc[ind-1].w[1] ||
				(fstar.w[1] == bid_ten2mk128trunc[ind-1].w[1] &&
					fstar.w[0] > bid_ten2mk128trunc[ind-1].w[0]) {
				is_inexact_lt_midpoint = 1
			}
		} else {
			is_inexact_gt_midpoint = 1
		}
	}

	// check for midpoints
	if (fstar.w[3] == 0) && (fstar.w[2] == 0) &&
		(fstar.w[1] != 0 || fstar.w[0] != 0) &&
		(fstar.w[1] < bid_ten2mk128trunc[ind-1].w[1] ||
			(fstar.w[1] == bid_ten2mk128trunc[ind-1].w[1] &&
				fstar.w[0] <= bid_ten2mk128trunc[ind-1].w[0])) {
		if Cstar.w[0]&0x01 != 0 {
			Cstar.w[0]--
			is_midpoint_gt_even = 1
			is_inexact_lt_midpoint = 0
			is_inexact_gt_midpoint = 0
		} else {
			is_midpoint_lt_even = 1
			is_inexact_lt_midpoint = 0
			is_inexact_gt_midpoint = 0
		}
	}

	// set inexact if not exact
	if is_inexact_lt_midpoint != 0 || is_inexact_gt_midpoint != 0 || is_midpoint_lt_even != 0 || is_midpoint_gt_even != 0 {
		pfpsf |= BID_INEXACT_EXCEPTION
	}

	switch mode {
	case 0: // xfloor (RM)
		if x_sign != 0 && (is_midpoint_gt_even != 0 || is_inexact_lt_midpoint != 0) {
			Cstar.w[0] = Cstar.w[0] + 1
		} else if x_sign == 0 && (is_midpoint_lt_even != 0 || is_inexact_gt_midpoint != 0) {
			Cstar.w[0] = Cstar.w[0] - 1
		}
	case 1: // xceil (RP)
		if x_sign != 0 && (is_midpoint_lt_even != 0 || is_inexact_gt_midpoint != 0) {
			Cstar.w[0] = Cstar.w[0] - 1
		} else if x_sign == 0 && (is_midpoint_gt_even != 0 || is_inexact_lt_midpoint != 0) {
			Cstar.w[0] = Cstar.w[0] + 1
		}
	case 2: // xint (RZ)
		if is_midpoint_lt_even != 0 || is_inexact_gt_midpoint != 0 {
			Cstar.w[0] = Cstar.w[0] - 1
		}
	}
	return Cstar.w[0], pfpsf
}

// bid128_round_rninta_common rounds C1 to integer using round-to-nearest-away.
func bid128_round_rninta_common(C1 BID_UINT128, ind int) (Cstar_w0 uint64) {
	var Cstar BID_UINT128
	var P256 BID_UINT256

	tmp64 := C1.w[0]
	if ind <= 19 {
		C1.w[0] = C1.w[0] + bid_midpoint64[ind-1]
	} else {
		C1.w[0] = C1.w[0] + bid_midpoint128[ind-20].w[0]
		C1.w[1] = C1.w[1] + bid_midpoint128[ind-20].w[1]
	}
	if C1.w[0] < tmp64 {
		C1.w[1]++
	}
	P256 = __mul_128x128_to_256(C1, bid_ten2mk128[ind-1])
	if ind-1 <= 21 {
		Cstar.w[1] = P256.w[3]
		Cstar.w[0] = P256.w[2]
	} else {
		Cstar.w[1] = 0
		Cstar.w[0] = P256.w[3]
	}

	shift := bid_shiftright128[ind-1]
	if ind-1 <= 21 {
		Cstar.w[0] = (Cstar.w[0] >> uint(shift)) | (Cstar.w[1] << uint(64-shift))
	} else {
		Cstar.w[0] = Cstar.w[0] >> uint(shift-64)
	}
	// no midpoint correction needed for rninta (ties round away from zero)
	return Cstar.w[0]
}

// bid128_round_xrninta_common rounds C1 with INEXACT flag detection for xrninta.
func bid128_round_xrninta_common(C1 BID_UINT128, ind int) (Cstar_w0 uint64, pfpsf uint32) {
	var Cstar BID_UINT128
	var fstar BID_UINT256
	var P256 BID_UINT256

	tmp64 := C1.w[0]
	if ind <= 19 {
		C1.w[0] = C1.w[0] + bid_midpoint64[ind-1]
	} else {
		C1.w[0] = C1.w[0] + bid_midpoint128[ind-20].w[0]
		C1.w[1] = C1.w[1] + bid_midpoint128[ind-20].w[1]
	}
	if C1.w[0] < tmp64 {
		C1.w[1]++
	}
	P256 = __mul_128x128_to_256(C1, bid_ten2mk128[ind-1])
	if ind-1 <= 21 {
		Cstar.w[1] = P256.w[3]
		Cstar.w[0] = P256.w[2]
		fstar.w[3] = 0
		fstar.w[2] = P256.w[2] & bid_maskhigh128[ind-1]
		fstar.w[1] = P256.w[1]
		fstar.w[0] = P256.w[0]
	} else {
		Cstar.w[1] = 0
		Cstar.w[0] = P256.w[3]
		fstar.w[3] = P256.w[3] & bid_maskhigh128[ind-1]
		fstar.w[2] = P256.w[2]
		fstar.w[1] = P256.w[1]
		fstar.w[0] = P256.w[0]
	}

	shift := bid_shiftright128[ind-1]
	if ind-1 <= 21 {
		Cstar.w[0] = (Cstar.w[0] >> uint(shift)) | (Cstar.w[1] << uint(64-shift))
	} else {
		Cstar.w[0] = Cstar.w[0] >> uint(shift-64)
	}

	// determine inexactness for xrninta
	var tmp64A uint64
	if ind-1 <= 2 {
		if fstar.w[1] > 0x8000000000000000 || (fstar.w[1] == 0x8000000000000000 && fstar.w[0] > 0x0) {
			tmp64 = fstar.w[1] - 0x8000000000000000
			if tmp64 > bid_ten2mk128trunc[ind-1].w[1] ||
				(tmp64 == bid_ten2mk128trunc[ind-1].w[1] &&
					fstar.w[0] >= bid_ten2mk128trunc[ind-1].w[0]) {
				pfpsf |= BID_INEXACT_EXCEPTION
			}
		} else {
			pfpsf |= BID_INEXACT_EXCEPTION
		}
	} else if ind-1 <= 21 {
		if fstar.w[3] > 0x0 ||
			(fstar.w[3] == 0x0 && fstar.w[2] > bid_onehalf128[ind-1]) ||
			(fstar.w[3] == 0x0 && fstar.w[2] == bid_onehalf128[ind-1] &&
				(fstar.w[1] != 0 || fstar.w[0] != 0)) {
			tmp64 = fstar.w[2] - bid_onehalf128[ind-1]
			tmp64A = fstar.w[3]
			if tmp64 > fstar.w[2] {
				tmp64A--
			}
			if tmp64A != 0 || tmp64 != 0 ||
				fstar.w[1] > bid_ten2mk128trunc[ind-1].w[1] ||
				(fstar.w[1] == bid_ten2mk128trunc[ind-1].w[1] &&
					fstar.w[0] > bid_ten2mk128trunc[ind-1].w[0]) {
				pfpsf |= BID_INEXACT_EXCEPTION
			}
		} else {
			pfpsf |= BID_INEXACT_EXCEPTION
		}
	} else {
		if fstar.w[3] > bid_onehalf128[ind-1] ||
			(fstar.w[3] == bid_onehalf128[ind-1] &&
				(fstar.w[2] != 0 || fstar.w[1] != 0 || fstar.w[0] != 0)) {
			tmp64 = fstar.w[3] - bid_onehalf128[ind-1]
			if tmp64 != 0 || fstar.w[2] != 0 ||
				fstar.w[1] > bid_ten2mk128trunc[ind-1].w[1] ||
				(fstar.w[1] == bid_ten2mk128trunc[ind-1].w[1] &&
					fstar.w[0] > bid_ten2mk128trunc[ind-1].w[0]) {
				pfpsf |= BID_INEXACT_EXCEPTION
			}
		} else {
			pfpsf |= BID_INEXACT_EXCEPTION
		}
	}

	// no midpoint correction for rninta
	return Cstar.w[0], pfpsf
}

// ============================================================================
// BID128 -> int32 conversions
// ============================================================================

// Bid128ToInt32Rnint: bid128_to_int32_rnint
func Bid128ToInt32Rnint(x BID_UINT128) (int32, uint32) {
	var res int32
	var pfpsf uint32

	x_sign, x_exp, C1, is_special := bid128_unpack_for_int(x)

	if is_special {
		pfpsf |= BID_INVALID_EXCEPTION
		return int32(-0x80000000), pfpsf
	}
	if bid128_is_noncanonical(C1, x) {
		return 0, 0
	}
	if C1.w[1] == 0 && C1.w[0] == 0 {
		return 0, 0
	}

	q, _ := bid128_nr_digits(C1)
	exp := int(x_exp>>49) - 6176

	if q+exp > 10 {
		pfpsf |= BID_INVALID_EXCEPTION
		return int32(-0x80000000), pfpsf
	} else if q+exp == 10 {
		invalid, f := bid128_check_overflow_10(C1, x_sign, q,
			0x500000005, false, // neg: > 0x500000005
			0x4fffffffb, true) // pos: >= 0x4fffffffb
		if invalid {
			return int32(-0x80000000), f
		}
	}

	if q+exp < 0 {
		return 0, 0
	} else if q+exp == 0 {
		ind := q - 1
		if ind <= 18 {
			if C1.w[1] == 0 && C1.w[0] <= bid_midpoint64[ind] {
				res = 0
			} else if x_sign != 0 {
				res = -1
			} else {
				res = 1
			}
		} else {
			if C1.w[1] < bid_midpoint128[ind-19].w[1] ||
				(C1.w[1] == bid_midpoint128[ind-19].w[1] &&
					C1.w[0] <= bid_midpoint128[ind-19].w[0]) {
				res = 0
			} else if x_sign != 0 {
				res = -1
			} else {
				res = 1
			}
		}
	} else {
		if exp < 0 {
			ind := -exp
			Cstar_w0 := bid128_round_rnint_common(C1, ind)
			if x_sign != 0 {
				res = -int32(Cstar_w0)
			} else {
				res = int32(Cstar_w0)
			}
		} else if exp == 0 {
			if x_sign != 0 {
				res = -int32(C1.w[0])
			} else {
				res = int32(C1.w[0])
			}
		} else {
			if x_sign != 0 {
				res = -int32(C1.w[0] * bid_ten2k64[exp])
			} else {
				res = int32(C1.w[0] * bid_ten2k64[exp])
			}
		}
	}
	return res, pfpsf
}

// Bid128ToInt32Xrnint: bid128_to_int32_xrnint
func Bid128ToInt32Xrnint(x BID_UINT128) (int32, uint32) {
	var res int32
	var pfpsf uint32

	x_sign, x_exp, C1, is_special := bid128_unpack_for_int(x)

	if is_special {
		pfpsf |= BID_INVALID_EXCEPTION
		return int32(-0x80000000), pfpsf
	}
	if bid128_is_noncanonical(C1, x) {
		return 0, 0
	}
	if C1.w[1] == 0 && C1.w[0] == 0 {
		return 0, 0
	}

	q, _ := bid128_nr_digits(C1)
	exp := int(x_exp>>49) - 6176

	if q+exp > 10 {
		pfpsf |= BID_INVALID_EXCEPTION
		return int32(-0x80000000), pfpsf
	} else if q+exp == 10 {
		invalid, f := bid128_check_overflow_10(C1, x_sign, q,
			0x500000005, false, 0x4fffffffb, true)
		if invalid {
			return int32(-0x80000000), f
		}
	}

	if q+exp < 0 {
		pfpsf |= BID_INEXACT_EXCEPTION
		return 0, pfpsf
	} else if q+exp == 0 {
		ind := q - 1
		if ind <= 18 {
			if C1.w[1] == 0 && C1.w[0] <= bid_midpoint64[ind] {
				res = 0
			} else if x_sign != 0 {
				res = -1
			} else {
				res = 1
			}
		} else {
			if C1.w[1] < bid_midpoint128[ind-19].w[1] ||
				(C1.w[1] == bid_midpoint128[ind-19].w[1] &&
					C1.w[0] <= bid_midpoint128[ind-19].w[0]) {
				res = 0
			} else if x_sign != 0 {
				res = -1
			} else {
				res = 1
			}
		}
		pfpsf |= BID_INEXACT_EXCEPTION
	} else {
		if exp < 0 {
			ind := -exp
			Cstar_w0, f := bid128_round_xrnint_common(C1, ind)
			pfpsf |= f
			if x_sign != 0 {
				res = -int32(Cstar_w0)
			} else {
				res = int32(Cstar_w0)
			}
		} else if exp == 0 {
			if x_sign != 0 {
				res = -int32(C1.w[0])
			} else {
				res = int32(C1.w[0])
			}
		} else {
			if x_sign != 0 {
				res = -int32(C1.w[0] * bid_ten2k64[exp])
			} else {
				res = int32(C1.w[0] * bid_ten2k64[exp])
			}
		}
	}
	return res, pfpsf
}

// Bid128ToInt32Floor: bid128_to_int32_floor
func Bid128ToInt32Floor(x BID_UINT128) (int32, uint32) {
	var res int32
	var pfpsf uint32

	x_sign, x_exp, C1, is_special := bid128_unpack_for_int(x)

	if is_special {
		pfpsf |= BID_INVALID_EXCEPTION
		return int32(-0x80000000), pfpsf
	}
	if bid128_is_noncanonical(C1, x) {
		return 0, 0
	}
	if C1.w[1] == 0 && C1.w[0] == 0 {
		return 0, 0
	}

	q, _ := bid128_nr_digits(C1)
	exp := int(x_exp>>49) - 6176

	if q+exp > 10 {
		pfpsf |= BID_INVALID_EXCEPTION
		return int32(-0x80000000), pfpsf
	} else if q+exp == 10 {
		invalid, f := bid128_check_overflow_10(C1, x_sign, q,
			0x500000000, false, // neg: > 0x500000000
			0x500000000, true) // pos: >= 0x500000000
		if invalid {
			return int32(-0x80000000), f
		}
	}

	if q+exp <= 0 {
		if x_sign != 0 {
			return -1, 0
		}
		return 0, 0
	} else {
		if exp < 0 {
			ind := -exp
			Cstar_w0 := bid128_round_floor_ceil_int_common(C1, ind, x_sign, 0)
			if x_sign != 0 {
				res = -int32(Cstar_w0)
			} else {
				res = int32(Cstar_w0)
			}
		} else if exp == 0 {
			if x_sign != 0 {
				res = -int32(C1.w[0])
			} else {
				res = int32(C1.w[0])
			}
		} else {
			if x_sign != 0 {
				res = -int32(C1.w[0] * bid_ten2k64[exp])
			} else {
				res = int32(C1.w[0] * bid_ten2k64[exp])
			}
		}
	}
	return res, pfpsf
}

// Bid128ToInt32Xfloor: bid128_to_int32_xfloor
func Bid128ToInt32Xfloor(x BID_UINT128) (int32, uint32) {
	var res int32
	var pfpsf uint32

	x_sign, x_exp, C1, is_special := bid128_unpack_for_int(x)

	if is_special {
		pfpsf |= BID_INVALID_EXCEPTION
		return int32(-0x80000000), pfpsf
	}
	if bid128_is_noncanonical(C1, x) {
		return 0, 0
	}
	if C1.w[1] == 0 && C1.w[0] == 0 {
		return 0, 0
	}

	q, _ := bid128_nr_digits(C1)
	exp := int(x_exp>>49) - 6176

	if q+exp > 10 {
		pfpsf |= BID_INVALID_EXCEPTION
		return int32(-0x80000000), pfpsf
	} else if q+exp == 10 {
		invalid, f := bid128_check_overflow_10(C1, x_sign, q,
			0x500000000, false, 0x500000000, true)
		if invalid {
			return int32(-0x80000000), f
		}
	}

	if q+exp <= 0 {
		if x_sign != 0 {
			pfpsf |= BID_INEXACT_EXCEPTION
			return -1, pfpsf
		}
		pfpsf |= BID_INEXACT_EXCEPTION
		return 0, pfpsf
	} else {
		if exp < 0 {
			ind := -exp
			Cstar_w0, f := bid128_round_xfloor_xceil_xint_common(C1, ind, x_sign, 0)
			pfpsf |= f
			if x_sign != 0 {
				res = -int32(Cstar_w0)
			} else {
				res = int32(Cstar_w0)
			}
		} else if exp == 0 {
			if x_sign != 0 {
				res = -int32(C1.w[0])
			} else {
				res = int32(C1.w[0])
			}
		} else {
			if x_sign != 0 {
				res = -int32(C1.w[0] * bid_ten2k64[exp])
			} else {
				res = int32(C1.w[0] * bid_ten2k64[exp])
			}
		}
	}
	return res, pfpsf
}

// Bid128ToInt32Ceil: bid128_to_int32_ceil
func Bid128ToInt32Ceil(x BID_UINT128) (int32, uint32) {
	var res int32
	var pfpsf uint32

	x_sign, x_exp, C1, is_special := bid128_unpack_for_int(x)

	if is_special {
		pfpsf |= BID_INVALID_EXCEPTION
		return int32(-0x80000000), pfpsf
	}
	if bid128_is_noncanonical(C1, x) {
		return 0, 0
	}
	if C1.w[1] == 0 && C1.w[0] == 0 {
		return 0, 0
	}

	q, _ := bid128_nr_digits(C1)
	exp := int(x_exp>>49) - 6176

	if q+exp > 10 {
		pfpsf |= BID_INVALID_EXCEPTION
		return int32(-0x80000000), pfpsf
	} else if q+exp == 10 {
		invalid, f := bid128_check_overflow_10(C1, x_sign, q,
			0x50000000a, true, // neg: >= 0x50000000a
			0x4fffffff6, false) // pos: > 0x4fffffff6
		if invalid {
			return int32(-0x80000000), f
		}
	}

	if q+exp <= 0 {
		if x_sign != 0 {
			return 0, 0
		}
		return 1, 0
	} else {
		if exp < 0 {
			ind := -exp
			Cstar_w0 := bid128_round_floor_ceil_int_common(C1, ind, x_sign, 1)
			if x_sign != 0 {
				res = -int32(Cstar_w0)
			} else {
				res = int32(Cstar_w0)
			}
		} else if exp == 0 {
			if x_sign != 0 {
				res = -int32(C1.w[0])
			} else {
				res = int32(C1.w[0])
			}
		} else {
			if x_sign != 0 {
				res = -int32(C1.w[0] * bid_ten2k64[exp])
			} else {
				res = int32(C1.w[0] * bid_ten2k64[exp])
			}
		}
	}
	return res, pfpsf
}

// Bid128ToInt32Xceil: bid128_to_int32_xceil
func Bid128ToInt32Xceil(x BID_UINT128) (int32, uint32) {
	var res int32
	var pfpsf uint32

	x_sign, x_exp, C1, is_special := bid128_unpack_for_int(x)

	if is_special {
		pfpsf |= BID_INVALID_EXCEPTION
		return int32(-0x80000000), pfpsf
	}
	if bid128_is_noncanonical(C1, x) {
		return 0, 0
	}
	if C1.w[1] == 0 && C1.w[0] == 0 {
		return 0, 0
	}

	q, _ := bid128_nr_digits(C1)
	exp := int(x_exp>>49) - 6176

	if q+exp > 10 {
		pfpsf |= BID_INVALID_EXCEPTION
		return int32(-0x80000000), pfpsf
	} else if q+exp == 10 {
		invalid, f := bid128_check_overflow_10(C1, x_sign, q,
			0x50000000a, true, 0x4fffffff6, false)
		if invalid {
			return int32(-0x80000000), f
		}
	}

	if q+exp <= 0 {
		if x_sign != 0 {
			pfpsf |= BID_INEXACT_EXCEPTION
			return 0, pfpsf
		}
		pfpsf |= BID_INEXACT_EXCEPTION
		return 1, pfpsf
	} else {
		if exp < 0 {
			ind := -exp
			Cstar_w0, f := bid128_round_xfloor_xceil_xint_common(C1, ind, x_sign, 1)
			pfpsf |= f
			if x_sign != 0 {
				res = -int32(Cstar_w0)
			} else {
				res = int32(Cstar_w0)
			}
		} else if exp == 0 {
			if x_sign != 0 {
				res = -int32(C1.w[0])
			} else {
				res = int32(C1.w[0])
			}
		} else {
			if x_sign != 0 {
				res = -int32(C1.w[0] * bid_ten2k64[exp])
			} else {
				res = int32(C1.w[0] * bid_ten2k64[exp])
			}
		}
	}
	return res, pfpsf
}

// Bid128ToInt32Int: bid128_to_int32_int (truncate)
func Bid128ToInt32Int(x BID_UINT128) (int32, uint32) {
	var res int32
	var pfpsf uint32

	x_sign, x_exp, C1, is_special := bid128_unpack_for_int(x)

	if is_special {
		pfpsf |= BID_INVALID_EXCEPTION
		return int32(-0x80000000), pfpsf
	}
	if bid128_is_noncanonical(C1, x) {
		return 0, 0
	}
	if C1.w[1] == 0 && C1.w[0] == 0 {
		return 0, 0
	}

	q, _ := bid128_nr_digits(C1)
	exp := int(x_exp>>49) - 6176

	if q+exp > 10 {
		pfpsf |= BID_INVALID_EXCEPTION
		return int32(-0x80000000), pfpsf
	} else if q+exp == 10 {
		invalid, f := bid128_check_overflow_10(C1, x_sign, q,
			0x50000000a, true, // neg: >= 0x50000000a
			0x500000000, true) // pos: >= 0x500000000
		if invalid {
			return int32(-0x80000000), f
		}
	}

	if q+exp <= 0 {
		return 0, 0
	} else {
		if exp < 0 {
			ind := -exp
			Cstar_w0 := bid128_round_floor_ceil_int_common(C1, ind, x_sign, 2)
			if x_sign != 0 {
				res = -int32(Cstar_w0)
			} else {
				res = int32(Cstar_w0)
			}
		} else if exp == 0 {
			if x_sign != 0 {
				res = -int32(C1.w[0])
			} else {
				res = int32(C1.w[0])
			}
		} else {
			if x_sign != 0 {
				res = -int32(C1.w[0] * bid_ten2k64[exp])
			} else {
				res = int32(C1.w[0] * bid_ten2k64[exp])
			}
		}
	}
	return res, pfpsf
}

// Bid128ToInt32Xint: bid128_to_int32_xint
func Bid128ToInt32Xint(x BID_UINT128) (int32, uint32) {
	var res int32
	var pfpsf uint32

	x_sign, x_exp, C1, is_special := bid128_unpack_for_int(x)

	if is_special {
		pfpsf |= BID_INVALID_EXCEPTION
		return int32(-0x80000000), pfpsf
	}
	if bid128_is_noncanonical(C1, x) {
		return 0, 0
	}
	if C1.w[1] == 0 && C1.w[0] == 0 {
		return 0, 0
	}

	q, _ := bid128_nr_digits(C1)
	exp := int(x_exp>>49) - 6176

	if q+exp > 10 {
		pfpsf |= BID_INVALID_EXCEPTION
		return int32(-0x80000000), pfpsf
	} else if q+exp == 10 {
		invalid, f := bid128_check_overflow_10(C1, x_sign, q,
			0x50000000a, true, 0x500000000, true)
		if invalid {
			return int32(-0x80000000), f
		}
	}

	if q+exp <= 0 {
		pfpsf |= BID_INEXACT_EXCEPTION
		return 0, pfpsf
	} else {
		if exp < 0 {
			ind := -exp
			Cstar_w0, f := bid128_round_xfloor_xceil_xint_common(C1, ind, x_sign, 2)
			pfpsf |= f
			if x_sign != 0 {
				res = -int32(Cstar_w0)
			} else {
				res = int32(Cstar_w0)
			}
		} else if exp == 0 {
			if x_sign != 0 {
				res = -int32(C1.w[0])
			} else {
				res = int32(C1.w[0])
			}
		} else {
			if x_sign != 0 {
				res = -int32(C1.w[0] * bid_ten2k64[exp])
			} else {
				res = int32(C1.w[0] * bid_ten2k64[exp])
			}
		}
	}
	return res, pfpsf
}

// Bid128ToInt32Rninta: bid128_to_int32_rninta
func Bid128ToInt32Rninta(x BID_UINT128) (int32, uint32) {
	var res int32
	var pfpsf uint32

	x_sign, x_exp, C1, is_special := bid128_unpack_for_int(x)

	if is_special {
		pfpsf |= BID_INVALID_EXCEPTION
		return int32(-0x80000000), pfpsf
	}
	if bid128_is_noncanonical(C1, x) {
		return 0, 0
	}
	if C1.w[1] == 0 && C1.w[0] == 0 {
		return 0, 0
	}

	q, _ := bid128_nr_digits(C1)
	exp := int(x_exp>>49) - 6176

	if q+exp > 10 {
		pfpsf |= BID_INVALID_EXCEPTION
		return int32(-0x80000000), pfpsf
	} else if q+exp == 10 {
		// rninta: neg >= 0x500000005, pos >= 0x4fffffffb
		invalid, f := bid128_check_overflow_10(C1, x_sign, q,
			0x500000005, true, 0x4fffffffb, true)
		if invalid {
			return int32(-0x80000000), f
		}
	}

	if q+exp < 0 {
		return 0, 0
	} else if q+exp == 0 {
		ind := q - 1
		if ind <= 18 {
			if C1.w[1] == 0 && C1.w[0] < bid_midpoint64[ind] {
				res = 0
			} else if x_sign != 0 {
				res = -1
			} else {
				res = 1
			}
		} else {
			if C1.w[1] < bid_midpoint128[ind-19].w[1] ||
				(C1.w[1] == bid_midpoint128[ind-19].w[1] &&
					C1.w[0] < bid_midpoint128[ind-19].w[0]) {
				res = 0
			} else if x_sign != 0 {
				res = -1
			} else {
				res = 1
			}
		}
	} else {
		if exp < 0 {
			ind := -exp
			Cstar_w0 := bid128_round_rninta_common(C1, ind)
			if x_sign != 0 {
				res = -int32(Cstar_w0)
			} else {
				res = int32(Cstar_w0)
			}
		} else if exp == 0 {
			if x_sign != 0 {
				res = -int32(C1.w[0])
			} else {
				res = int32(C1.w[0])
			}
		} else {
			if x_sign != 0 {
				res = -int32(C1.w[0] * bid_ten2k64[exp])
			} else {
				res = int32(C1.w[0] * bid_ten2k64[exp])
			}
		}
	}
	return res, pfpsf
}

// Bid128ToInt32Xrninta: bid128_to_int32_xrninta
func Bid128ToInt32Xrninta(x BID_UINT128) (int32, uint32) {
	var res int32
	var pfpsf uint32

	x_sign, x_exp, C1, is_special := bid128_unpack_for_int(x)

	if is_special {
		pfpsf |= BID_INVALID_EXCEPTION
		return int32(-0x80000000), pfpsf
	}
	if bid128_is_noncanonical(C1, x) {
		return 0, 0
	}
	if C1.w[1] == 0 && C1.w[0] == 0 {
		return 0, 0
	}

	q, _ := bid128_nr_digits(C1)
	exp := int(x_exp>>49) - 6176

	if q+exp > 10 {
		pfpsf |= BID_INVALID_EXCEPTION
		return int32(-0x80000000), pfpsf
	} else if q+exp == 10 {
		invalid, f := bid128_check_overflow_10(C1, x_sign, q,
			0x500000005, true, 0x4fffffffb, true)
		if invalid {
			return int32(-0x80000000), f
		}
	}

	if q+exp < 0 {
		pfpsf |= BID_INEXACT_EXCEPTION
		return 0, pfpsf
	} else if q+exp == 0 {
		ind := q - 1
		if ind <= 18 {
			if C1.w[1] == 0 && C1.w[0] < bid_midpoint64[ind] {
				res = 0
			} else if x_sign != 0 {
				res = -1
			} else {
				res = 1
			}
		} else {
			if C1.w[1] < bid_midpoint128[ind-19].w[1] ||
				(C1.w[1] == bid_midpoint128[ind-19].w[1] &&
					C1.w[0] < bid_midpoint128[ind-19].w[0]) {
				res = 0
			} else if x_sign != 0 {
				res = -1
			} else {
				res = 1
			}
		}
		pfpsf |= BID_INEXACT_EXCEPTION
	} else {
		if exp < 0 {
			ind := -exp
			Cstar_w0, f := bid128_round_xrninta_common(C1, ind)
			pfpsf |= f
			if x_sign != 0 {
				res = -int32(Cstar_w0)
			} else {
				res = int32(Cstar_w0)
			}
		} else if exp == 0 {
			if x_sign != 0 {
				res = -int32(C1.w[0])
			} else {
				res = int32(C1.w[0])
			}
		} else {
			if x_sign != 0 {
				res = -int32(C1.w[0] * bid_ten2k64[exp])
			} else {
				res = int32(C1.w[0] * bid_ten2k64[exp])
			}
		}
	}
	return res, pfpsf
}

// ============================================================================
// BID128 -> int64 conversions
// ============================================================================

// bid128_check_overflow_19 checks overflow for int64 (q+exp == 19).
func bid128_check_overflow_19(x BID_UINT128, x_sign uint64, C1 BID_UINT128, q int,
	neg_hi, neg_lo uint64, neg_cmp_ge bool,
	pos_hi, pos_lo uint64, pos_cmp_ge bool) (bool, uint32, BID_UINT128) {
	var C BID_UINT128
	var pfpsf uint32

	if x_sign != 0 {
		C.w[1] = neg_hi
		C.w[0] = neg_lo
		if q <= 19 {
			C1 = __mul_64x64_to_128(C1.w[0], bid_ten2k64[20-q])
		} else if q == 20 {
			// C1 * 10^0 = C1
		} else {
			C = __mul_128x64_to_128(bid_ten2k64[q-20], C)
		}
		if neg_cmp_ge {
			if C1.w[1] > C.w[1] || (C1.w[1] == C.w[1] && C1.w[0] >= C.w[0]) {
				pfpsf |= BID_INVALID_EXCEPTION
				return true, pfpsf, C1
			}
		} else {
			if C1.w[1] > C.w[1] || (C1.w[1] == C.w[1] && C1.w[0] > C.w[0]) {
				pfpsf |= BID_INVALID_EXCEPTION
				return true, pfpsf, C1
			}
		}
	} else {
		C.w[1] = pos_hi
		C.w[0] = pos_lo
		if q <= 19 {
			C1 = __mul_64x64_to_128(C1.w[0], bid_ten2k64[20-q])
		} else if q == 20 {
			// C1 * 10^0 = C1
		} else {
			C = __mul_128x64_to_128(bid_ten2k64[q-20], C)
		}
		if pos_cmp_ge {
			if C1.w[1] > C.w[1] || (C1.w[1] == C.w[1] && C1.w[0] >= C.w[0]) {
				pfpsf |= BID_INVALID_EXCEPTION
				return true, pfpsf, C1
			}
		} else {
			if C1.w[1] > C.w[1] || (C1.w[1] == C.w[1] && C1.w[0] > C.w[0]) {
				pfpsf |= BID_INVALID_EXCEPTION
				return true, pfpsf, C1
			}
		}
	}
	return false, 0, C1
}

func bid128_cmp_128(a, b BID_UINT128) int {
	if a.w[1] < b.w[1] {
		return -1
	}
	if a.w[1] > b.w[1] {
		return 1
	}
	if a.w[0] < b.w[0] {
		return -1
	}
	if a.w[0] > b.w[0] {
		return 1
	}
	return 0
}

func bid128_check_overflow_20(C1 BID_UINT128, x_sign uint64, q int,
	neg_hi, neg_lo uint64, neg_cmp_ge bool,
	pos_hi, pos_lo uint64, pos_cmp_ge bool) (bool, uint32) {
	var scaled BID_UINT128
	var limit BID_UINT128
	var cmp int
	var pfpsf uint32

	limit.w[1] = pos_hi
	limit.w[0] = pos_lo
	if x_sign != 0 {
		limit.w[1] = neg_hi
		limit.w[0] = neg_lo
	}

	if q < 21 {
		if q == 1 {
			scaled = __mul_128x64_to_128(C1.w[0], bid_ten2k128[0])
		} else if q <= 19 {
			scaled = __mul_64x64_to_128(C1.w[0], bid_ten2k64[21-q])
		} else {
			scaled = __mul_128x64_to_128(bid_ten2k64[1], C1)
		}
		cmp = bid128_cmp_128(scaled, limit)
	} else if q == 21 {
		cmp = bid128_cmp_128(C1, limit)
	} else {
		limit = __mul_128x64_to_128(bid_ten2k64[q-21], limit)
		cmp = bid128_cmp_128(C1, limit)
	}

	if x_sign != 0 {
		if cmp > 0 || (neg_cmp_ge && cmp == 0) {
			pfpsf |= BID_INVALID_EXCEPTION
			return true, pfpsf
		}
	} else {
		if cmp > 0 || (pos_cmp_ge && cmp == 0) {
			pfpsf |= BID_INVALID_EXCEPTION
			return true, pfpsf
		}
	}
	return false, 0
}

func bid128_int64_core(x BID_UINT128, roundMode int, setInexact bool,
	maxQE int, invalidRes int64,
	negHi, negLo uint64, negCmpGe bool,
	posHi, posLo uint64, posCmpGe bool,
	qpExpEqHandler func(x_sign uint64, C1 BID_UINT128, q, exp int) (int64, uint32),
	qpExpLeqHandler func(x_sign uint64, C1 BID_UINT128, q, exp int) (int64, uint32),
) (int64, uint32) {
	var pfpsf uint32

	x_sign, x_exp, C1, is_special := bid128_unpack_for_int(x)

	if is_special {
		pfpsf |= BID_INVALID_EXCEPTION
		return invalidRes, pfpsf
	}
	if bid128_is_noncanonical(C1, x) {
		return 0, 0
	}
	if C1.w[1] == 0 && C1.w[0] == 0 {
		return 0, 0
	}

	q, _ := bid128_nr_digits(C1)
	exp := int(x_exp>>49) - 6176

	if q+exp > maxQE {
		pfpsf |= BID_INVALID_EXCEPTION
		return invalidRes, pfpsf
	} else if q+exp == maxQE {
		invalid, f, newC1 := bid128_check_overflow_19(x, x_sign, C1, q,
			negHi, negLo, negCmpGe,
			posHi, posLo, posCmpGe)
		if invalid {
			return invalidRes, f
		}
		C1 = newC1
		// Restore C1 which may have been modified above
		C1.w[1] = x.w[1] & MASK_COEFF128
		C1.w[0] = x.w[0]
	}

	return qpExpLeqHandler(x_sign, C1, q, exp)
}

// Bid128ToInt64Rnint: bid128_to_int64_rnint
func Bid128ToInt64Rnint(x BID_UINT128) (int64, uint32) {
	var res int64
	var pfpsf uint32

	x_sign, x_exp, C1, is_special := bid128_unpack_for_int(x)

	if is_special {
		pfpsf |= BID_INVALID_EXCEPTION
		return -0x8000000000000000, pfpsf
	}
	if bid128_is_noncanonical(C1, x) {
		return 0, 0
	}
	if C1.w[1] == 0 && C1.w[0] == 0 {
		return 0, 0
	}

	q, _ := bid128_nr_digits(C1)
	exp := int(x_exp>>49) - 6176

	if q+exp > 19 {
		pfpsf |= BID_INVALID_EXCEPTION
		return -0x8000000000000000, pfpsf
	} else if q+exp == 19 {
		invalid, f, _ := bid128_check_overflow_19(x, x_sign, C1, q,
			0x0000000000000005, 0x0000000000000005, false, // neg: > {5,5}
			0x0000000000000004, 0xfffffffffffffffb, true) // pos: >= {4, 0xfffffffffffffffb}
		if invalid {
			return -0x8000000000000000, f
		}
		// Restore C1
		C1.w[1] = x.w[1] & MASK_COEFF128
		C1.w[0] = x.w[0]
	}

	if q+exp < 0 {
		return 0, 0
	} else if q+exp == 0 {
		ind := q - 1
		if ind <= 18 {
			if C1.w[1] == 0 && C1.w[0] <= bid_midpoint64[ind] {
				res = 0
			} else if x_sign != 0 {
				res = -1
			} else {
				res = 1
			}
		} else {
			if C1.w[1] < bid_midpoint128[ind-19].w[1] ||
				(C1.w[1] == bid_midpoint128[ind-19].w[1] &&
					C1.w[0] <= bid_midpoint128[ind-19].w[0]) {
				res = 0
			} else if x_sign != 0 {
				res = -1
			} else {
				res = 1
			}
		}
	} else {
		if exp < 0 {
			ind := -exp
			Cstar_w0 := bid128_round_rnint_common(C1, ind)
			if x_sign != 0 {
				res = -int64(Cstar_w0)
			} else {
				res = int64(Cstar_w0)
			}
		} else if exp == 0 {
			if x_sign != 0 {
				res = -int64(C1.w[0])
			} else {
				res = int64(C1.w[0])
			}
		} else {
			if x_sign != 0 {
				res = -int64(C1.w[0] * bid_ten2k64[exp])
			} else {
				res = int64(C1.w[0] * bid_ten2k64[exp])
			}
		}
	}
	return res, pfpsf
}

// Bid128ToInt64Xrnint: bid128_to_int64_xrnint
func Bid128ToInt64Xrnint(x BID_UINT128) (int64, uint32) {
	var res int64
	var pfpsf uint32

	x_sign, x_exp, C1, is_special := bid128_unpack_for_int(x)

	if is_special {
		pfpsf |= BID_INVALID_EXCEPTION
		return -0x8000000000000000, pfpsf
	}
	if bid128_is_noncanonical(C1, x) {
		return 0, 0
	}
	if C1.w[1] == 0 && C1.w[0] == 0 {
		return 0, 0
	}

	q, _ := bid128_nr_digits(C1)
	exp := int(x_exp>>49) - 6176

	if q+exp > 19 {
		pfpsf |= BID_INVALID_EXCEPTION
		return -0x8000000000000000, pfpsf
	} else if q+exp == 19 {
		invalid, f, _ := bid128_check_overflow_19(x, x_sign, C1, q,
			0x0000000000000005, 0x0000000000000005, false,
			0x0000000000000004, 0xfffffffffffffffb, true)
		if invalid {
			return -0x8000000000000000, f
		}
		C1.w[1] = x.w[1] & MASK_COEFF128
		C1.w[0] = x.w[0]
	}

	if q+exp < 0 {
		pfpsf |= BID_INEXACT_EXCEPTION
		return 0, pfpsf
	} else if q+exp == 0 {
		ind := q - 1
		if ind <= 18 {
			if C1.w[1] == 0 && C1.w[0] <= bid_midpoint64[ind] {
				res = 0
			} else if x_sign != 0 {
				res = -1
			} else {
				res = 1
			}
		} else {
			if C1.w[1] < bid_midpoint128[ind-19].w[1] ||
				(C1.w[1] == bid_midpoint128[ind-19].w[1] &&
					C1.w[0] <= bid_midpoint128[ind-19].w[0]) {
				res = 0
			} else if x_sign != 0 {
				res = -1
			} else {
				res = 1
			}
		}
		pfpsf |= BID_INEXACT_EXCEPTION
	} else {
		if exp < 0 {
			ind := -exp
			Cstar_w0, f := bid128_round_xrnint_common(C1, ind)
			pfpsf |= f
			if x_sign != 0 {
				res = -int64(Cstar_w0)
			} else {
				res = int64(Cstar_w0)
			}
		} else if exp == 0 {
			if x_sign != 0 {
				res = -int64(C1.w[0])
			} else {
				res = int64(C1.w[0])
			}
		} else {
			if x_sign != 0 {
				res = -int64(C1.w[0] * bid_ten2k64[exp])
			} else {
				res = int64(C1.w[0] * bid_ten2k64[exp])
			}
		}
	}
	return res, pfpsf
}

// Bid128ToInt64Floor: bid128_to_int64_floor
func Bid128ToInt64Floor(x BID_UINT128) (int64, uint32) {
	var res int64
	var pfpsf uint32

	x_sign, x_exp, C1, is_special := bid128_unpack_for_int(x)

	if is_special {
		pfpsf |= BID_INVALID_EXCEPTION
		return -0x8000000000000000, pfpsf
	}
	if bid128_is_noncanonical(C1, x) {
		return 0, 0
	}
	if C1.w[1] == 0 && C1.w[0] == 0 {
		return 0, 0
	}

	q, _ := bid128_nr_digits(C1)
	exp := int(x_exp>>49) - 6176

	if q+exp > 19 {
		pfpsf |= BID_INVALID_EXCEPTION
		return -0x8000000000000000, pfpsf
	} else if q+exp == 19 {
		// floor: neg > {5,0}, pos >= {4, 0xfffffffffffffffe}
		invalid, f, _ := bid128_check_overflow_19(x, x_sign, C1, q,
			0x0000000000000005, 0x0000000000000000, false,
			0x0000000000000005, 0x0000000000000000, true)
		if invalid {
			return -0x8000000000000000, f
		}
		C1.w[1] = x.w[1] & MASK_COEFF128
		C1.w[0] = x.w[0]
	}

	if q+exp <= 0 {
		if x_sign != 0 {
			return -1, 0
		}
		return 0, 0
	} else {
		if exp < 0 {
			ind := -exp
			Cstar_w0, _ := bid128_round_trunc_mode_common(C1, ind, x_sign, 0, false)
			if x_sign != 0 {
				res = -int64(Cstar_w0)
			} else {
				res = int64(Cstar_w0)
			}
		} else if exp == 0 {
			if x_sign != 0 {
				res = -int64(C1.w[0])
			} else {
				res = int64(C1.w[0])
			}
		} else {
			if x_sign != 0 {
				res = -int64(C1.w[0] * bid_ten2k64[exp])
			} else {
				res = int64(C1.w[0] * bid_ten2k64[exp])
			}
		}
	}
	return res, pfpsf
}

// Bid128ToInt64Xfloor: bid128_to_int64_xfloor
func Bid128ToInt64Xfloor(x BID_UINT128) (int64, uint32) {
	var res int64
	var pfpsf uint32

	x_sign, x_exp, C1, is_special := bid128_unpack_for_int(x)

	if is_special {
		pfpsf |= BID_INVALID_EXCEPTION
		return -0x8000000000000000, pfpsf
	}
	if bid128_is_noncanonical(C1, x) {
		return 0, 0
	}
	if C1.w[1] == 0 && C1.w[0] == 0 {
		return 0, 0
	}

	q, _ := bid128_nr_digits(C1)
	exp := int(x_exp>>49) - 6176

	if q+exp > 19 {
		pfpsf |= BID_INVALID_EXCEPTION
		return -0x8000000000000000, pfpsf
	} else if q+exp == 19 {
		invalid, f, _ := bid128_check_overflow_19(x, x_sign, C1, q,
			0x0000000000000005, 0x0000000000000000, false,
			0x0000000000000005, 0x0000000000000000, true)
		if invalid {
			return -0x8000000000000000, f
		}
		C1.w[1] = x.w[1] & MASK_COEFF128
		C1.w[0] = x.w[0]
	}

	if q+exp <= 0 {
		pfpsf |= BID_INEXACT_EXCEPTION
		if x_sign != 0 {
			return -1, pfpsf
		}
		return 0, pfpsf
	} else {
		if exp < 0 {
			ind := -exp
			Cstar_w0, f := bid128_round_trunc_mode_common(C1, ind, x_sign, 0, true)
			pfpsf |= f
			if x_sign != 0 {
				res = -int64(Cstar_w0)
			} else {
				res = int64(Cstar_w0)
			}
		} else if exp == 0 {
			if x_sign != 0 {
				res = -int64(C1.w[0])
			} else {
				res = int64(C1.w[0])
			}
		} else {
			if x_sign != 0 {
				res = -int64(C1.w[0] * bid_ten2k64[exp])
			} else {
				res = int64(C1.w[0] * bid_ten2k64[exp])
			}
		}
	}
	return res, pfpsf
}

// Bid128ToInt64Ceil: bid128_to_int64_ceil
func Bid128ToInt64Ceil(x BID_UINT128) (int64, uint32) {
	var res int64
	var pfpsf uint32

	x_sign, x_exp, C1, is_special := bid128_unpack_for_int(x)
	if is_special {
		pfpsf |= BID_INVALID_EXCEPTION
		return -0x8000000000000000, pfpsf
	}
	if bid128_is_noncanonical(C1, x) {
		return 0, 0
	}
	if C1.w[1] == 0 && C1.w[0] == 0 {
		return 0, 0
	}

	q, _ := bid128_nr_digits(C1)
	exp := int(x_exp>>49) - 6176

	if q+exp > 19 {
		pfpsf |= BID_INVALID_EXCEPTION
		return -0x8000000000000000, pfpsf
	} else if q+exp == 19 {
		// ceil: neg >= {5, 2}, pos > {4, 0xfffffffffffffffa}
		invalid, f, _ := bid128_check_overflow_19(x, x_sign, C1, q,
			0x0000000000000005, 0x000000000000000a, true,
			0x0000000000000004, 0xfffffffffffffff6, false)
		if invalid {
			return -0x8000000000000000, f
		}
		C1.w[1] = x.w[1] & MASK_COEFF128
		C1.w[0] = x.w[0]
	}

	if q+exp <= 0 {
		if x_sign != 0 {
			return 0, 0
		}
		return 1, 0
	} else {
		if exp < 0 {
			ind := -exp
			Cstar_w0, _ := bid128_round_trunc_mode_common(C1, ind, x_sign, 1, false)
			if x_sign != 0 {
				res = -int64(Cstar_w0)
			} else {
				res = int64(Cstar_w0)
			}
		} else if exp == 0 {
			if x_sign != 0 {
				res = -int64(C1.w[0])
			} else {
				res = int64(C1.w[0])
			}
		} else {
			if x_sign != 0 {
				res = -int64(C1.w[0] * bid_ten2k64[exp])
			} else {
				res = int64(C1.w[0] * bid_ten2k64[exp])
			}
		}
	}
	return res, pfpsf
}

// Bid128ToInt64Xceil: bid128_to_int64_xceil
func Bid128ToInt64Xceil(x BID_UINT128) (int64, uint32) {
	var res int64
	var pfpsf uint32

	x_sign, x_exp, C1, is_special := bid128_unpack_for_int(x)
	if is_special {
		pfpsf |= BID_INVALID_EXCEPTION
		return -0x8000000000000000, pfpsf
	}
	if bid128_is_noncanonical(C1, x) {
		return 0, 0
	}
	if C1.w[1] == 0 && C1.w[0] == 0 {
		return 0, 0
	}

	q, _ := bid128_nr_digits(C1)
	exp := int(x_exp>>49) - 6176

	if q+exp > 19 {
		pfpsf |= BID_INVALID_EXCEPTION
		return -0x8000000000000000, pfpsf
	} else if q+exp == 19 {
		invalid, f, _ := bid128_check_overflow_19(x, x_sign, C1, q,
			0x0000000000000005, 0x000000000000000a, true,
			0x0000000000000004, 0xfffffffffffffff6, false)
		if invalid {
			return -0x8000000000000000, f
		}
		C1.w[1] = x.w[1] & MASK_COEFF128
		C1.w[0] = x.w[0]
	}

	if q+exp <= 0 {
		pfpsf |= BID_INEXACT_EXCEPTION
		if x_sign != 0 {
			return 0, pfpsf
		}
		return 1, pfpsf
	} else {
		if exp < 0 {
			ind := -exp
			Cstar_w0, f := bid128_round_trunc_mode_common(C1, ind, x_sign, 1, true)
			pfpsf |= f
			if x_sign != 0 {
				res = -int64(Cstar_w0)
			} else {
				res = int64(Cstar_w0)
			}
		} else if exp == 0 {
			if x_sign != 0 {
				res = -int64(C1.w[0])
			} else {
				res = int64(C1.w[0])
			}
		} else {
			if x_sign != 0 {
				res = -int64(C1.w[0] * bid_ten2k64[exp])
			} else {
				res = int64(C1.w[0] * bid_ten2k64[exp])
			}
		}
	}
	return res, pfpsf
}

// Bid128ToInt64Int: bid128_to_int64_int (truncate)
func Bid128ToInt64Int(x BID_UINT128) (int64, uint32) {
	var res int64
	var pfpsf uint32

	x_sign, x_exp, C1, is_special := bid128_unpack_for_int(x)
	if is_special {
		pfpsf |= BID_INVALID_EXCEPTION
		return -0x8000000000000000, pfpsf
	}
	if bid128_is_noncanonical(C1, x) {
		return 0, 0
	}
	if C1.w[1] == 0 && C1.w[0] == 0 {
		return 0, 0
	}

	q, _ := bid128_nr_digits(C1)
	exp := int(x_exp>>49) - 6176

	if q+exp > 19 {
		pfpsf |= BID_INVALID_EXCEPTION
		return -0x8000000000000000, pfpsf
	} else if q+exp == 19 {
		// int: neg >= {5, 2}, pos >= {4, 0xfffffffffffffffe}
		invalid, f, _ := bid128_check_overflow_19(x, x_sign, C1, q,
			0x0000000000000005, 0x000000000000000a, true,
			0x0000000000000005, 0x0000000000000000, true)
		if invalid {
			return -0x8000000000000000, f
		}
		C1.w[1] = x.w[1] & MASK_COEFF128
		C1.w[0] = x.w[0]
	}

	if q+exp <= 0 {
		return 0, 0
	} else {
		if exp < 0 {
			ind := -exp
			Cstar_w0, _ := bid128_round_trunc_mode_common(C1, ind, x_sign, 2, false)
			if x_sign != 0 {
				res = -int64(Cstar_w0)
			} else {
				res = int64(Cstar_w0)
			}
		} else if exp == 0 {
			if x_sign != 0 {
				res = -int64(C1.w[0])
			} else {
				res = int64(C1.w[0])
			}
		} else {
			if x_sign != 0 {
				res = -int64(C1.w[0] * bid_ten2k64[exp])
			} else {
				res = int64(C1.w[0] * bid_ten2k64[exp])
			}
		}
	}
	return res, pfpsf
}

// Bid128ToInt64Xint: bid128_to_int64_xint
func Bid128ToInt64Xint(x BID_UINT128) (int64, uint32) {
	var res int64
	var pfpsf uint32

	x_sign, x_exp, C1, is_special := bid128_unpack_for_int(x)
	if is_special {
		pfpsf |= BID_INVALID_EXCEPTION
		return -0x8000000000000000, pfpsf
	}
	if bid128_is_noncanonical(C1, x) {
		return 0, 0
	}
	if C1.w[1] == 0 && C1.w[0] == 0 {
		return 0, 0
	}

	q, _ := bid128_nr_digits(C1)
	exp := int(x_exp>>49) - 6176

	if q+exp > 19 {
		pfpsf |= BID_INVALID_EXCEPTION
		return -0x8000000000000000, pfpsf
	} else if q+exp == 19 {
		invalid, f, _ := bid128_check_overflow_19(x, x_sign, C1, q,
			0x0000000000000005, 0x000000000000000a, true,
			0x0000000000000005, 0x0000000000000000, true)
		if invalid {
			return -0x8000000000000000, f
		}
		C1.w[1] = x.w[1] & MASK_COEFF128
		C1.w[0] = x.w[0]
	}

	if q+exp <= 0 {
		pfpsf |= BID_INEXACT_EXCEPTION
		return 0, pfpsf
	} else {
		if exp < 0 {
			ind := -exp
			Cstar_w0, f := bid128_round_trunc_mode_common(C1, ind, x_sign, 2, true)
			pfpsf |= f
			if x_sign != 0 {
				res = -int64(Cstar_w0)
			} else {
				res = int64(Cstar_w0)
			}
		} else if exp == 0 {
			if x_sign != 0 {
				res = -int64(C1.w[0])
			} else {
				res = int64(C1.w[0])
			}
		} else {
			if x_sign != 0 {
				res = -int64(C1.w[0] * bid_ten2k64[exp])
			} else {
				res = int64(C1.w[0] * bid_ten2k64[exp])
			}
		}
	}
	return res, pfpsf
}

// Bid128ToInt64Rninta: bid128_to_int64_rninta
func Bid128ToInt64Rninta(x BID_UINT128) (int64, uint32) {
	var res int64
	var pfpsf uint32

	x_sign, x_exp, C1, is_special := bid128_unpack_for_int(x)
	if is_special {
		pfpsf |= BID_INVALID_EXCEPTION
		return -0x8000000000000000, pfpsf
	}
	if bid128_is_noncanonical(C1, x) {
		return 0, 0
	}
	if C1.w[1] == 0 && C1.w[0] == 0 {
		return 0, 0
	}

	q, _ := bid128_nr_digits(C1)
	exp := int(x_exp>>49) - 6176

	if q+exp > 19 {
		pfpsf |= BID_INVALID_EXCEPTION
		return -0x8000000000000000, pfpsf
	} else if q+exp == 19 {
		invalid, f, _ := bid128_check_overflow_19(x, x_sign, C1, q,
			0x0000000000000005, 0x0000000000000005, true,
			0x0000000000000004, 0xfffffffffffffffb, true)
		if invalid {
			return -0x8000000000000000, f
		}
		C1.w[1] = x.w[1] & MASK_COEFF128
		C1.w[0] = x.w[0]
	}

	if q+exp < 0 {
		return 0, 0
	} else if q+exp == 0 {
		ind := q - 1
		if ind <= 18 {
			if C1.w[1] == 0 && C1.w[0] < bid_midpoint64[ind] {
				res = 0
			} else if x_sign != 0 {
				res = -1
			} else {
				res = 1
			}
		} else {
			if C1.w[1] < bid_midpoint128[ind-19].w[1] ||
				(C1.w[1] == bid_midpoint128[ind-19].w[1] &&
					C1.w[0] < bid_midpoint128[ind-19].w[0]) {
				res = 0
			} else if x_sign != 0 {
				res = -1
			} else {
				res = 1
			}
		}
	} else {
		if exp < 0 {
			ind := -exp
			Cstar_w0 := bid128_round_rninta_common(C1, ind)
			if x_sign != 0 {
				res = -int64(Cstar_w0)
			} else {
				res = int64(Cstar_w0)
			}
		} else if exp == 0 {
			if x_sign != 0 {
				res = -int64(C1.w[0])
			} else {
				res = int64(C1.w[0])
			}
		} else {
			if x_sign != 0 {
				res = -int64(C1.w[0] * bid_ten2k64[exp])
			} else {
				res = int64(C1.w[0] * bid_ten2k64[exp])
			}
		}
	}
	return res, pfpsf
}

// Bid128ToInt64Xrninta: bid128_to_int64_xrninta
func Bid128ToInt64Xrninta(x BID_UINT128) (int64, uint32) {
	var res int64
	var pfpsf uint32

	x_sign, x_exp, C1, is_special := bid128_unpack_for_int(x)
	if is_special {
		pfpsf |= BID_INVALID_EXCEPTION
		return -0x8000000000000000, pfpsf
	}
	if bid128_is_noncanonical(C1, x) {
		return 0, 0
	}
	if C1.w[1] == 0 && C1.w[0] == 0 {
		return 0, 0
	}

	q, _ := bid128_nr_digits(C1)
	exp := int(x_exp>>49) - 6176

	if q+exp > 19 {
		pfpsf |= BID_INVALID_EXCEPTION
		return -0x8000000000000000, pfpsf
	} else if q+exp == 19 {
		invalid, f, _ := bid128_check_overflow_19(x, x_sign, C1, q,
			0x0000000000000005, 0x0000000000000005, true,
			0x0000000000000004, 0xfffffffffffffffb, true)
		if invalid {
			return -0x8000000000000000, f
		}
		C1.w[1] = x.w[1] & MASK_COEFF128
		C1.w[0] = x.w[0]
	}

	if q+exp < 0 {
		pfpsf |= BID_INEXACT_EXCEPTION
		return 0, pfpsf
	} else if q+exp == 0 {
		ind := q - 1
		if ind <= 18 {
			if C1.w[1] == 0 && C1.w[0] < bid_midpoint64[ind] {
				res = 0
			} else if x_sign != 0 {
				res = -1
			} else {
				res = 1
			}
		} else {
			if C1.w[1] < bid_midpoint128[ind-19].w[1] ||
				(C1.w[1] == bid_midpoint128[ind-19].w[1] &&
					C1.w[0] < bid_midpoint128[ind-19].w[0]) {
				res = 0
			} else if x_sign != 0 {
				res = -1
			} else {
				res = 1
			}
		}
		pfpsf |= BID_INEXACT_EXCEPTION
	} else {
		if exp < 0 {
			ind := -exp
			Cstar_w0, f := bid128_round_xrninta_common(C1, ind)
			pfpsf |= f
			if x_sign != 0 {
				res = -int64(Cstar_w0)
			} else {
				res = int64(Cstar_w0)
			}
		} else if exp == 0 {
			if x_sign != 0 {
				res = -int64(C1.w[0])
			} else {
				res = int64(C1.w[0])
			}
		} else {
			if x_sign != 0 {
				res = -int64(C1.w[0] * bid_ten2k64[exp])
			} else {
				res = int64(C1.w[0] * bid_ten2k64[exp])
			}
		}
	}
	return res, pfpsf
}

// ============================================================================
// BID128 -> uint32 conversions
// ============================================================================

// bid128_uint_rnint_qpexp0 handles (q+exp)==0 for unsigned rnint.
func bid128_uint_rnint_qpexp0(C1 BID_UINT128, x_sign uint64, q int) (uint32, uint32) {
	ind := q - 1
	if ind <= 18 {
		if C1.w[1] == 0 && C1.w[0] <= bid_midpoint64[ind] {
			return 0, 0
		} else if x_sign == 0 {
			return 1, 0
		} else {
			return 0x80000000, BID_INVALID_EXCEPTION
		}
	} else {
		if C1.w[1] < bid_midpoint128[ind-19].w[1] ||
			(C1.w[1] == bid_midpoint128[ind-19].w[1] &&
				C1.w[0] <= bid_midpoint128[ind-19].w[0]) {
			return 0, 0
		} else if x_sign == 0 {
			return 1, 0
		} else {
			return 0x80000000, BID_INVALID_EXCEPTION
		}
	}
}

// Bid128ToUint32Rnint: bid128_to_uint32_rnint
func Bid128ToUint32Rnint(x BID_UINT128) (uint32, uint32) {
	var pfpsf uint32

	x_sign, x_exp, C1, is_special := bid128_unpack_for_int(x)
	if is_special {
		pfpsf |= BID_INVALID_EXCEPTION
		return 0x80000000, pfpsf
	}
	if bid128_is_noncanonical(C1, x) {
		return 0, 0
	}
	if C1.w[1] == 0 && C1.w[0] == 0 {
		return 0, 0
	}

	q, _ := bid128_nr_digits(C1)
	exp := int(x_exp>>49) - 6176

	if q+exp > 10 {
		pfpsf |= BID_INVALID_EXCEPTION
		return 0x80000000, pfpsf
	} else if q+exp == 10 {
		invalid, f := bid128_check_overflow_10(C1, x_sign, q,
			0x05, false, // neg: > 0x05
			0x9fffffffb, true) // pos: >= 0x9fffffffb
		if invalid {
			return 0x80000000, f
		}
	}

	if q+exp < 0 {
		return 0, 0
	} else if q+exp == 0 {
		return bid128_uint_rnint_qpexp0(C1, x_sign, q)
	} else {
		if x_sign != 0 { // x <= -1
			pfpsf |= BID_INVALID_EXCEPTION
			return 0x80000000, pfpsf
		}
		if exp < 0 {
			ind := -exp
			Cstar_w0 := bid128_round_rnint_common(C1, ind)
			return uint32(Cstar_w0), pfpsf
		} else if exp == 0 {
			return uint32(C1.w[0]), 0
		} else {
			return uint32(C1.w[0] * bid_ten2k64[exp]), 0
		}
	}
}

// Bid128ToUint32Xrnint: bid128_to_uint32_xrnint
func Bid128ToUint32Xrnint(x BID_UINT128) (uint32, uint32) {
	var pfpsf uint32

	x_sign, x_exp, C1, is_special := bid128_unpack_for_int(x)
	if is_special {
		pfpsf |= BID_INVALID_EXCEPTION
		return 0x80000000, pfpsf
	}
	if bid128_is_noncanonical(C1, x) {
		return 0, 0
	}
	if C1.w[1] == 0 && C1.w[0] == 0 {
		return 0, 0
	}

	q, _ := bid128_nr_digits(C1)
	exp := int(x_exp>>49) - 6176

	if q+exp > 10 {
		pfpsf |= BID_INVALID_EXCEPTION
		return 0x80000000, pfpsf
	} else if q+exp == 10 {
		invalid, f := bid128_check_overflow_10(C1, x_sign, q,
			0x05, false, 0x9fffffffb, true)
		if invalid {
			return 0x80000000, f
		}
	}

	if q+exp < 0 {
		pfpsf |= BID_INEXACT_EXCEPTION
		return 0, pfpsf
	} else if q+exp == 0 {
		r, f := bid128_uint_rnint_qpexp0(C1, x_sign, q)
		if f != 0 {
			return r, f
		}
		return r, BID_INEXACT_EXCEPTION
	} else {
		if x_sign != 0 {
			pfpsf |= BID_INVALID_EXCEPTION
			return 0x80000000, pfpsf
		}
		if exp < 0 {
			ind := -exp
			Cstar_w0, f := bid128_round_xrnint_common(C1, ind)
			pfpsf |= f
			return uint32(Cstar_w0), pfpsf
		} else if exp == 0 {
			return uint32(C1.w[0]), 0
		} else {
			return uint32(C1.w[0] * bid_ten2k64[exp]), 0
		}
	}
}

// Bid128ToUint32Floor: bid128_to_uint32_floor
func Bid128ToUint32Floor(x BID_UINT128) (uint32, uint32) {
	var pfpsf uint32

	x_sign, x_exp, C1, is_special := bid128_unpack_for_int(x)
	if is_special {
		pfpsf |= BID_INVALID_EXCEPTION
		return 0x80000000, pfpsf
	}
	if bid128_is_noncanonical(C1, x) {
		return 0, 0
	}
	if C1.w[1] == 0 && C1.w[0] == 0 {
		return 0, 0
	}

	q, _ := bid128_nr_digits(C1)
	exp := int(x_exp>>49) - 6176

	if q+exp > 10 {
		pfpsf |= BID_INVALID_EXCEPTION
		return 0x80000000, pfpsf
	} else if q+exp == 10 {
		invalid, f := bid128_check_overflow_10(C1, x_sign, q,
			0x0a, true, // neg: any negative with q+exp=10 -> invalid
			0xa00000000, true) // pos: >= 0xa00000000
		if invalid {
			return 0x80000000, f
		}
	}

	if q+exp <= 0 {
		if x_sign != 0 {
			pfpsf |= BID_INVALID_EXCEPTION
			return 0x80000000, pfpsf
		}
		return 0, 0
	} else {
		if x_sign != 0 {
			pfpsf |= BID_INVALID_EXCEPTION
			return 0x80000000, pfpsf
		}
		if exp < 0 {
			ind := -exp
			Cstar_w0 := bid128_round_floor_ceil_int_common(C1, ind, x_sign, 0)
			return uint32(Cstar_w0), pfpsf
		} else if exp == 0 {
			return uint32(C1.w[0]), 0
		} else {
			return uint32(C1.w[0] * bid_ten2k64[exp]), 0
		}
	}
}

// Bid128ToUint32Xfloor: bid128_to_uint32_xfloor
func Bid128ToUint32Xfloor(x BID_UINT128) (uint32, uint32) {
	var pfpsf uint32

	x_sign, x_exp, C1, is_special := bid128_unpack_for_int(x)
	if is_special {
		pfpsf |= BID_INVALID_EXCEPTION
		return 0x80000000, pfpsf
	}
	if bid128_is_noncanonical(C1, x) {
		return 0, 0
	}
	if C1.w[1] == 0 && C1.w[0] == 0 {
		return 0, 0
	}

	q, _ := bid128_nr_digits(C1)
	exp := int(x_exp>>49) - 6176

	if q+exp > 10 {
		pfpsf |= BID_INVALID_EXCEPTION
		return 0x80000000, pfpsf
	} else if q+exp == 10 {
		invalid, f := bid128_check_overflow_10(C1, x_sign, q,
			0x0a, true, 0xa00000000, true)
		if invalid {
			return 0x80000000, f
		}
	}

	if q+exp <= 0 {
		if x_sign != 0 {
			pfpsf |= BID_INVALID_EXCEPTION
			return 0x80000000, pfpsf
		}
		pfpsf |= BID_INEXACT_EXCEPTION
		return 0, pfpsf
	} else {
		if x_sign != 0 {
			pfpsf |= BID_INVALID_EXCEPTION
			return 0x80000000, pfpsf
		}
		if exp < 0 {
			ind := -exp
			Cstar_w0, f := bid128_round_xfloor_xceil_xint_common(C1, ind, x_sign, 0)
			pfpsf |= f
			return uint32(Cstar_w0), pfpsf
		} else if exp == 0 {
			return uint32(C1.w[0]), 0
		} else {
			return uint32(C1.w[0] * bid_ten2k64[exp]), 0
		}
	}
}

// Bid128ToUint32Ceil: bid128_to_uint32_ceil
func Bid128ToUint32Ceil(x BID_UINT128) (uint32, uint32) {
	var pfpsf uint32

	x_sign, x_exp, C1, is_special := bid128_unpack_for_int(x)
	if is_special {
		pfpsf |= BID_INVALID_EXCEPTION
		return 0x80000000, pfpsf
	}
	if bid128_is_noncanonical(C1, x) {
		return 0, 0
	}
	if C1.w[1] == 0 && C1.w[0] == 0 {
		return 0, 0
	}

	q, _ := bid128_nr_digits(C1)
	exp := int(x_exp>>49) - 6176

	if q+exp > 10 {
		pfpsf |= BID_INVALID_EXCEPTION
		return 0x80000000, pfpsf
	} else if q+exp == 10 {
		invalid, f := bid128_check_overflow_10(C1, x_sign, q,
			0x0a, true, // neg: any negative is invalid for uint
			0x9fffffff6, false) // pos: > 0x9fffffff6
		if invalid {
			return 0x80000000, f
		}
	}

	if q+exp <= 0 {
		if x_sign != 0 {
			return 0, 0
		}
		return 1, 0
	} else {
		if x_sign != 0 {
			pfpsf |= BID_INVALID_EXCEPTION
			return 0x80000000, pfpsf
		}
		if exp < 0 {
			ind := -exp
			Cstar_w0 := bid128_round_floor_ceil_int_common(C1, ind, x_sign, 1)
			return uint32(Cstar_w0), pfpsf
		} else if exp == 0 {
			return uint32(C1.w[0]), 0
		} else {
			return uint32(C1.w[0] * bid_ten2k64[exp]), 0
		}
	}
}

// Bid128ToUint32Xceil: bid128_to_uint32_xceil
func Bid128ToUint32Xceil(x BID_UINT128) (uint32, uint32) {
	var pfpsf uint32

	x_sign, x_exp, C1, is_special := bid128_unpack_for_int(x)
	if is_special {
		pfpsf |= BID_INVALID_EXCEPTION
		return 0x80000000, pfpsf
	}
	if bid128_is_noncanonical(C1, x) {
		return 0, 0
	}
	if C1.w[1] == 0 && C1.w[0] == 0 {
		return 0, 0
	}

	q, _ := bid128_nr_digits(C1)
	exp := int(x_exp>>49) - 6176

	if q+exp > 10 {
		pfpsf |= BID_INVALID_EXCEPTION
		return 0x80000000, pfpsf
	} else if q+exp == 10 {
		invalid, f := bid128_check_overflow_10(C1, x_sign, q,
			0x0a, true, 0x9fffffff6, false)
		if invalid {
			return 0x80000000, f
		}
	}

	if q+exp <= 0 {
		if x_sign != 0 {
			pfpsf |= BID_INEXACT_EXCEPTION
			return 0, pfpsf
		}
		pfpsf |= BID_INEXACT_EXCEPTION
		return 1, pfpsf
	} else {
		if x_sign != 0 {
			pfpsf |= BID_INVALID_EXCEPTION
			return 0x80000000, pfpsf
		}
		if exp < 0 {
			ind := -exp
			Cstar_w0, f := bid128_round_xfloor_xceil_xint_common(C1, ind, x_sign, 1)
			pfpsf |= f
			return uint32(Cstar_w0), pfpsf
		} else if exp == 0 {
			return uint32(C1.w[0]), 0
		} else {
			return uint32(C1.w[0] * bid_ten2k64[exp]), 0
		}
	}
}

// Bid128ToUint32Int: bid128_to_uint32_int
func Bid128ToUint32Int(x BID_UINT128) (uint32, uint32) {
	var pfpsf uint32

	x_sign, x_exp, C1, is_special := bid128_unpack_for_int(x)
	if is_special {
		pfpsf |= BID_INVALID_EXCEPTION
		return 0x80000000, pfpsf
	}
	if bid128_is_noncanonical(C1, x) {
		return 0, 0
	}
	if C1.w[1] == 0 && C1.w[0] == 0 {
		return 0, 0
	}

	q, _ := bid128_nr_digits(C1)
	exp := int(x_exp>>49) - 6176

	if q+exp > 10 {
		pfpsf |= BID_INVALID_EXCEPTION
		return 0x80000000, pfpsf
	} else if q+exp == 10 {
		invalid, f := bid128_check_overflow_10(C1, x_sign, q,
			0x0a, true, 0xa00000000, true)
		if invalid {
			return 0x80000000, f
		}
	}

	if q+exp <= 0 {
		if x_sign != 0 {
			return 0, 0
		}
		return 0, 0
	} else {
		if x_sign != 0 {
			pfpsf |= BID_INVALID_EXCEPTION
			return 0x80000000, pfpsf
		}
		if exp < 0 {
			ind := -exp
			Cstar_w0 := bid128_round_floor_ceil_int_common(C1, ind, x_sign, 2)
			return uint32(Cstar_w0), pfpsf
		} else if exp == 0 {
			return uint32(C1.w[0]), 0
		} else {
			return uint32(C1.w[0] * bid_ten2k64[exp]), 0
		}
	}
}

// Bid128ToUint32Xint: bid128_to_uint32_xint
func Bid128ToUint32Xint(x BID_UINT128) (uint32, uint32) {
	var pfpsf uint32

	x_sign, x_exp, C1, is_special := bid128_unpack_for_int(x)
	if is_special {
		pfpsf |= BID_INVALID_EXCEPTION
		return 0x80000000, pfpsf
	}
	if bid128_is_noncanonical(C1, x) {
		return 0, 0
	}
	if C1.w[1] == 0 && C1.w[0] == 0 {
		return 0, 0
	}

	q, _ := bid128_nr_digits(C1)
	exp := int(x_exp>>49) - 6176

	if q+exp > 10 {
		pfpsf |= BID_INVALID_EXCEPTION
		return 0x80000000, pfpsf
	} else if q+exp == 10 {
		invalid, f := bid128_check_overflow_10(C1, x_sign, q,
			0x0a, true, 0xa00000000, true)
		if invalid {
			return 0x80000000, f
		}
	}

	if q+exp <= 0 {
		if x_sign != 0 {
			pfpsf |= BID_INEXACT_EXCEPTION
			return 0, pfpsf
		}
		pfpsf |= BID_INEXACT_EXCEPTION
		return 0, pfpsf
	} else {
		if x_sign != 0 {
			pfpsf |= BID_INVALID_EXCEPTION
			return 0x80000000, pfpsf
		}
		if exp < 0 {
			ind := -exp
			Cstar_w0, f := bid128_round_xfloor_xceil_xint_common(C1, ind, x_sign, 2)
			pfpsf |= f
			return uint32(Cstar_w0), pfpsf
		} else if exp == 0 {
			return uint32(C1.w[0]), 0
		} else {
			return uint32(C1.w[0] * bid_ten2k64[exp]), 0
		}
	}
}

// Bid128ToUint32Rninta: bid128_to_uint32_rninta
func Bid128ToUint32Rninta(x BID_UINT128) (uint32, uint32) {
	var pfpsf uint32

	x_sign, x_exp, C1, is_special := bid128_unpack_for_int(x)
	if is_special {
		pfpsf |= BID_INVALID_EXCEPTION
		return 0x80000000, pfpsf
	}
	if bid128_is_noncanonical(C1, x) {
		return 0, 0
	}
	if C1.w[1] == 0 && C1.w[0] == 0 {
		return 0, 0
	}

	q, _ := bid128_nr_digits(C1)
	exp := int(x_exp>>49) - 6176

	if q+exp > 10 {
		pfpsf |= BID_INVALID_EXCEPTION
		return 0x80000000, pfpsf
	} else if q+exp == 10 {
		invalid, f := bid128_check_overflow_10(C1, x_sign, q,
			0x05, true, 0x9fffffffb, true)
		if invalid {
			return 0x80000000, f
		}
	}

	if q+exp < 0 {
		return 0, 0
	} else if q+exp == 0 {
		ind := q - 1
		if ind <= 18 {
			if C1.w[1] == 0 && C1.w[0] < bid_midpoint64[ind] {
				return 0, 0
			} else if x_sign == 0 {
				return 1, 0
			} else {
				return 0x80000000, BID_INVALID_EXCEPTION
			}
		} else {
			if C1.w[1] < bid_midpoint128[ind-19].w[1] ||
				(C1.w[1] == bid_midpoint128[ind-19].w[1] &&
					C1.w[0] < bid_midpoint128[ind-19].w[0]) {
				return 0, 0
			} else if x_sign == 0 {
				return 1, 0
			} else {
				return 0x80000000, BID_INVALID_EXCEPTION
			}
		}
	} else {
		if x_sign != 0 {
			pfpsf |= BID_INVALID_EXCEPTION
			return 0x80000000, pfpsf
		}
		if exp < 0 {
			ind := -exp
			Cstar_w0 := bid128_round_rninta_common(C1, ind)
			return uint32(Cstar_w0), pfpsf
		} else if exp == 0 {
			return uint32(C1.w[0]), 0
		} else {
			return uint32(C1.w[0] * bid_ten2k64[exp]), 0
		}
	}
}

// Bid128ToUint32Xrninta: bid128_to_uint32_xrninta
func Bid128ToUint32Xrninta(x BID_UINT128) (uint32, uint32) {
	var pfpsf uint32

	x_sign, x_exp, C1, is_special := bid128_unpack_for_int(x)
	if is_special {
		pfpsf |= BID_INVALID_EXCEPTION
		return 0x80000000, pfpsf
	}
	if bid128_is_noncanonical(C1, x) {
		return 0, 0
	}
	if C1.w[1] == 0 && C1.w[0] == 0 {
		return 0, 0
	}

	q, _ := bid128_nr_digits(C1)
	exp := int(x_exp>>49) - 6176

	if q+exp > 10 {
		pfpsf |= BID_INVALID_EXCEPTION
		return 0x80000000, pfpsf
	} else if q+exp == 10 {
		invalid, f := bid128_check_overflow_10(C1, x_sign, q,
			0x05, true, 0x9fffffffb, true)
		if invalid {
			return 0x80000000, f
		}
	}

	if q+exp < 0 {
		pfpsf |= BID_INEXACT_EXCEPTION
		return 0, pfpsf
	} else if q+exp == 0 {
		ind := q - 1
		if ind <= 18 {
			if C1.w[1] == 0 && C1.w[0] < bid_midpoint64[ind] {
				return 0, BID_INEXACT_EXCEPTION
			} else if x_sign == 0 {
				return 1, BID_INEXACT_EXCEPTION
			} else {
				return 0x80000000, BID_INVALID_EXCEPTION
			}
		} else {
			if C1.w[1] < bid_midpoint128[ind-19].w[1] ||
				(C1.w[1] == bid_midpoint128[ind-19].w[1] &&
					C1.w[0] < bid_midpoint128[ind-19].w[0]) {
				return 0, BID_INEXACT_EXCEPTION
			} else if x_sign == 0 {
				return 1, BID_INEXACT_EXCEPTION
			} else {
				return 0x80000000, BID_INVALID_EXCEPTION
			}
		}
	} else {
		if x_sign != 0 {
			pfpsf |= BID_INVALID_EXCEPTION
			return 0x80000000, pfpsf
		}
		if exp < 0 {
			ind := -exp
			Cstar_w0, f := bid128_round_xrninta_common(C1, ind)
			pfpsf |= f
			return uint32(Cstar_w0), pfpsf
		} else if exp == 0 {
			return uint32(C1.w[0]), 0
		} else {
			return uint32(C1.w[0] * bid_ten2k64[exp]), 0
		}
	}
}

// ============================================================================
// BID128 -> uint64 conversions
// These follow the same pattern but with uint64 return and different limits.
// For uint64, q+exp > 20 is overflow, q+exp == 20 is boundary check.
// ============================================================================

// Bid128ToUint64Rnint: bid128_to_uint64_rnint
func Bid128ToUint64Rnint(x BID_UINT128) (uint64, uint32) {
	return bid128_to_uint64_core(x, 0, false)
}

func Bid128ToUint64Xrnint(x BID_UINT128) (uint64, uint32) {
	return bid128_to_uint64_core(x, 0, true)
}

func Bid128ToUint64Floor(x BID_UINT128) (uint64, uint32) {
	return bid128_to_uint64_core(x, 1, false)
}

func Bid128ToUint64Xfloor(x BID_UINT128) (uint64, uint32) {
	return bid128_to_uint64_core(x, 1, true)
}

func Bid128ToUint64Ceil(x BID_UINT128) (uint64, uint32) {
	return bid128_to_uint64_core(x, 2, false)
}

func Bid128ToUint64Xceil(x BID_UINT128) (uint64, uint32) {
	return bid128_to_uint64_core(x, 2, true)
}

func Bid128ToUint64Int(x BID_UINT128) (uint64, uint32) {
	return bid128_to_uint64_core(x, 3, false)
}

func Bid128ToUint64Xint(x BID_UINT128) (uint64, uint32) {
	return bid128_to_uint64_core(x, 3, true)
}

func Bid128ToUint64Rninta(x BID_UINT128) (uint64, uint32) {
	return bid128_to_uint64_core(x, 4, false)
}

func Bid128ToUint64Xrninta(x BID_UINT128) (uint64, uint32) {
	return bid128_to_uint64_core(x, 4, true)
}

func bid128_uint_midpoint_cmp(C1 BID_UINT128, q int) int {
	ind := q - 1
	if ind <= 18 {
		if C1.w[1] == 0 {
			if C1.w[0] < bid_midpoint64[ind] {
				return -1
			}
			if C1.w[0] > bid_midpoint64[ind] {
				return 1
			}
			return 0
		}
		return 1
	}
	if C1.w[1] < bid_midpoint128[ind-19].w[1] {
		return -1
	}
	if C1.w[1] > bid_midpoint128[ind-19].w[1] {
		return 1
	}
	if C1.w[0] < bid_midpoint128[ind-19].w[0] {
		return -1
	}
	if C1.w[0] > bid_midpoint128[ind-19].w[0] {
		return 1
	}
	return 0
}

// mode: 0=rnint, 1=floor, 2=ceil, 3=int, 4=rninta
func bid128_to_uint64_core(x BID_UINT128, mode int, setInexact bool) (uint64, uint32) {
	var pfpsf uint32

	x_sign, x_exp, C1, is_special := bid128_unpack_for_int(x)
	if is_special {
		pfpsf |= BID_INVALID_EXCEPTION
		return 0x8000000000000000, pfpsf
	}
	if bid128_is_noncanonical(C1, x) {
		return 0, 0
	}
	if C1.w[1] == 0 && C1.w[0] == 0 {
		return 0, 0
	}

	q, _ := bid128_nr_digits(C1)
	exp := int(x_exp>>49) - 6176

	if q+exp > 20 {
		pfpsf |= BID_INVALID_EXCEPTION
		return 0x8000000000000000, pfpsf
	} else if q+exp == 20 {
		var invalid bool
		switch mode {
		case 0:
			invalid, pfpsf = bid128_check_overflow_20(C1, x_sign, q,
				0x0000000000000000, 0x0000000000000005, false,
				0x0000000000000009, 0xfffffffffffffffb, true)
		case 1:
			invalid, pfpsf = bid128_check_overflow_20(C1, x_sign, q,
				0x0000000000000000, 0x000000000000000a, true,
				0x000000000000000a, 0x0000000000000000, true)
		case 2:
			invalid, pfpsf = bid128_check_overflow_20(C1, x_sign, q,
				0x0000000000000000, 0x000000000000000a, true,
				0x0000000000000009, 0xfffffffffffffff6, false)
		case 3:
			invalid, pfpsf = bid128_check_overflow_20(C1, x_sign, q,
				0x0000000000000000, 0x000000000000000a, true,
				0x000000000000000a, 0x0000000000000000, true)
		default:
			invalid, pfpsf = bid128_check_overflow_20(C1, x_sign, q,
				0x0000000000000000, 0x0000000000000005, true,
				0x0000000000000009, 0xfffffffffffffffb, true)
		}
		if invalid {
			return 0x8000000000000000, pfpsf
		}
	}

	if q+exp < 0 {
		switch mode {
		case 0, 4:
			if setInexact {
				pfpsf |= BID_INEXACT_EXCEPTION
			}
			return 0, pfpsf
		case 1:
			if x_sign != 0 {
				pfpsf |= BID_INVALID_EXCEPTION
				return 0x8000000000000000, pfpsf
			}
			if setInexact {
				pfpsf |= BID_INEXACT_EXCEPTION
			}
			return 0, pfpsf
		case 2:
			if x_sign != 0 {
				if setInexact {
					pfpsf |= BID_INEXACT_EXCEPTION
				}
				return 0, pfpsf
			}
			if setInexact {
				pfpsf |= BID_INEXACT_EXCEPTION
			}
			return 1, pfpsf
		case 3:
			if setInexact {
				pfpsf |= BID_INEXACT_EXCEPTION
			}
			return 0, pfpsf
		}
	}

	if q+exp == 0 {
		cmp := bid128_uint_midpoint_cmp(C1, q)
		switch mode {
		case 0:
			if cmp <= 0 {
				if setInexact {
					pfpsf |= BID_INEXACT_EXCEPTION
				}
				return 0, pfpsf
			}
			if x_sign == 0 {
				if setInexact {
					pfpsf |= BID_INEXACT_EXCEPTION
				}
				return 1, pfpsf
			}
			pfpsf |= BID_INVALID_EXCEPTION
			return 0x8000000000000000, pfpsf
		case 1:
			if x_sign != 0 {
				pfpsf |= BID_INVALID_EXCEPTION
				return 0x8000000000000000, pfpsf
			}
			if setInexact {
				pfpsf |= BID_INEXACT_EXCEPTION
			}
			return 0, pfpsf
		case 2:
			if x_sign != 0 {
				if setInexact {
					pfpsf |= BID_INEXACT_EXCEPTION
				}
				return 0, pfpsf
			}
			if setInexact {
				pfpsf |= BID_INEXACT_EXCEPTION
			}
			return 1, pfpsf
		case 3:
			if setInexact {
				pfpsf |= BID_INEXACT_EXCEPTION
			}
			return 0, pfpsf
		default:
			if cmp < 0 {
				if setInexact {
					pfpsf |= BID_INEXACT_EXCEPTION
				}
				return 0, pfpsf
			}
			if x_sign == 0 {
				if setInexact {
					pfpsf |= BID_INEXACT_EXCEPTION
				}
				return 1, pfpsf
			}
			pfpsf |= BID_INVALID_EXCEPTION
			return 0x8000000000000000, pfpsf
		}
	}

	if x_sign != 0 {
		pfpsf |= BID_INVALID_EXCEPTION
		return 0x8000000000000000, pfpsf
	}

	if exp < 0 {
		ind := -exp
		switch mode {
		case 0:
			if setInexact {
				res, f := bid128_round_xrnint_common(C1, ind)
				return res, f
			}
			return bid128_round_rnint_common(C1, ind), 0
		case 1:
			res, f := bid128_round_trunc_mode_common(C1, ind, 0, 0, setInexact)
			return res, f
		case 2:
			res, f := bid128_round_trunc_mode_common(C1, ind, 0, 1, setInexact)
			return res, f
		case 3:
			res, f := bid128_round_trunc_mode_common(C1, ind, 0, 2, setInexact)
			return res, f
		default:
			if setInexact {
				res, f := bid128_round_xrninta_common(C1, ind)
				return res, f
			}
			return bid128_round_rninta_common(C1, ind), 0
		}
	} else if exp == 0 {
		return C1.w[0], 0
	}
	return C1.w[0] * bid_ten2k64[exp], 0
}

// ============================================================================
// BID128 -> small integer conversions (via int32/uint32)
// Ported from Intel BID_TO_SMALL_INT_CVT_FUNCTION / BID_TO_SMALL_BID_UINT_CVT_FUNCTION macros.
// ============================================================================

// === int8 (via int32 with SIZE_MASK=0xffffff80, INVALID_RESULT=0x80) ===

func bid128_to_small_int(fn func(BID_UINT128) (int32, uint32), x BID_UINT128, sizeMask int32, invalidResult int8) (int8, uint32) {
	v, f := fn(x)
	if f&BID_INVALID_EXCEPTION != 0 {
		return invalidResult, f
	}
	if v&sizeMask != 0 && v&sizeMask != sizeMask {
		return invalidResult, BID_INVALID_EXCEPTION
	}
	return int8(v), f
}

func bid128_to_small_uint(fn func(BID_UINT128) (uint32, uint32), x BID_UINT128, sizeMask uint32, invalidResult uint8) (uint8, uint32) {
	v, f := fn(x)
	if f&BID_INVALID_EXCEPTION != 0 {
		return invalidResult, f
	}
	if v&sizeMask != 0 {
		return invalidResult, BID_INVALID_EXCEPTION
	}
	return uint8(v), f
}

func bid128_to_small_int16(fn func(BID_UINT128) (int32, uint32), x BID_UINT128, sizeMask int32, invalidResult int16) (int16, uint32) {
	v, f := fn(x)
	if f&BID_INVALID_EXCEPTION != 0 {
		return invalidResult, f
	}
	if v&sizeMask != 0 && v&sizeMask != sizeMask {
		return invalidResult, BID_INVALID_EXCEPTION
	}
	return int16(v), f
}

func bid128_to_small_uint16(fn func(BID_UINT128) (uint32, uint32), x BID_UINT128, sizeMask uint32, invalidResult uint16) (uint16, uint32) {
	v, f := fn(x)
	if f&BID_INVALID_EXCEPTION != 0 {
		return invalidResult, f
	}
	if v&sizeMask != 0 {
		return invalidResult, BID_INVALID_EXCEPTION
	}
	return uint16(v), f
}

// int8 conversions
func Bid128ToInt8Rnint(x BID_UINT128) (int8, uint32) {
	return bid128_to_small_int(Bid128ToInt32Rnint, x, -128, -128)
}
func Bid128ToInt8Xrnint(x BID_UINT128) (int8, uint32) {
	return bid128_to_small_int(Bid128ToInt32Xrnint, x, -128, -128)
}
func Bid128ToInt8Rninta(x BID_UINT128) (int8, uint32) {
	return bid128_to_small_int(Bid128ToInt32Rninta, x, -128, -128)
}
func Bid128ToInt8Xrninta(x BID_UINT128) (int8, uint32) {
	return bid128_to_small_int(Bid128ToInt32Xrninta, x, -128, -128)
}
func Bid128ToInt8Int(x BID_UINT128) (int8, uint32) {
	return bid128_to_small_int(Bid128ToInt32Int, x, -128, -128)
}
func Bid128ToInt8Xint(x BID_UINT128) (int8, uint32) {
	return bid128_to_small_int(Bid128ToInt32Xint, x, -128, -128)
}
func Bid128ToInt8Floor(x BID_UINT128) (int8, uint32) {
	return bid128_to_small_int(Bid128ToInt32Floor, x, -128, -128)
}
func Bid128ToInt8Xfloor(x BID_UINT128) (int8, uint32) {
	return bid128_to_small_int(Bid128ToInt32Xfloor, x, -128, -128)
}
func Bid128ToInt8Ceil(x BID_UINT128) (int8, uint32) {
	return bid128_to_small_int(Bid128ToInt32Ceil, x, -128, -128)
}
func Bid128ToInt8Xceil(x BID_UINT128) (int8, uint32) {
	return bid128_to_small_int(Bid128ToInt32Xceil, x, -128, -128)
}

// int16 conversions
func Bid128ToInt16Rnint(x BID_UINT128) (int16, uint32) {
	return bid128_to_small_int16(Bid128ToInt32Rnint, x, -32768, -32768)
}
func Bid128ToInt16Xrnint(x BID_UINT128) (int16, uint32) {
	return bid128_to_small_int16(Bid128ToInt32Xrnint, x, -32768, -32768)
}
func Bid128ToInt16Rninta(x BID_UINT128) (int16, uint32) {
	return bid128_to_small_int16(Bid128ToInt32Rninta, x, -32768, -32768)
}
func Bid128ToInt16Xrninta(x BID_UINT128) (int16, uint32) {
	return bid128_to_small_int16(Bid128ToInt32Xrninta, x, -32768, -32768)
}
func Bid128ToInt16Int(x BID_UINT128) (int16, uint32) {
	return bid128_to_small_int16(Bid128ToInt32Int, x, -32768, -32768)
}
func Bid128ToInt16Xint(x BID_UINT128) (int16, uint32) {
	return bid128_to_small_int16(Bid128ToInt32Xint, x, -32768, -32768)
}
func Bid128ToInt16Floor(x BID_UINT128) (int16, uint32) {
	return bid128_to_small_int16(Bid128ToInt32Floor, x, -32768, -32768)
}
func Bid128ToInt16Xfloor(x BID_UINT128) (int16, uint32) {
	return bid128_to_small_int16(Bid128ToInt32Xfloor, x, -32768, -32768)
}
func Bid128ToInt16Ceil(x BID_UINT128) (int16, uint32) {
	return bid128_to_small_int16(Bid128ToInt32Ceil, x, -32768, -32768)
}
func Bid128ToInt16Xceil(x BID_UINT128) (int16, uint32) {
	return bid128_to_small_int16(Bid128ToInt32Xceil, x, -32768, -32768)
}

// uint8 conversions
func Bid128ToUint8Rnint(x BID_UINT128) (uint8, uint32) {
	return bid128_to_small_uint(Bid128ToUint32Rnint, x, 0xffffff00, 0x80)
}
func Bid128ToUint8Xrnint(x BID_UINT128) (uint8, uint32) {
	return bid128_to_small_uint(Bid128ToUint32Xrnint, x, 0xffffff00, 0x80)
}
func Bid128ToUint8Rninta(x BID_UINT128) (uint8, uint32) {
	return bid128_to_small_uint(Bid128ToUint32Rninta, x, 0xffffff00, 0x80)
}
func Bid128ToUint8Xrninta(x BID_UINT128) (uint8, uint32) {
	return bid128_to_small_uint(Bid128ToUint32Xrninta, x, 0xffffff00, 0x80)
}
func Bid128ToUint8Int(x BID_UINT128) (uint8, uint32) {
	return bid128_to_small_uint(Bid128ToUint32Int, x, 0xffffff00, 0x80)
}
func Bid128ToUint8Xint(x BID_UINT128) (uint8, uint32) {
	return bid128_to_small_uint(Bid128ToUint32Xint, x, 0xffffff00, 0x80)
}
func Bid128ToUint8Floor(x BID_UINT128) (uint8, uint32) {
	return bid128_to_small_uint(Bid128ToUint32Floor, x, 0xffffff00, 0x80)
}
func Bid128ToUint8Xfloor(x BID_UINT128) (uint8, uint32) {
	return bid128_to_small_uint(Bid128ToUint32Xfloor, x, 0xffffff00, 0x80)
}
func Bid128ToUint8Ceil(x BID_UINT128) (uint8, uint32) {
	return bid128_to_small_uint(Bid128ToUint32Ceil, x, 0xffffff00, 0x80)
}
func Bid128ToUint8Xceil(x BID_UINT128) (uint8, uint32) {
	return bid128_to_small_uint(Bid128ToUint32Xceil, x, 0xffffff00, 0x80)
}

// uint16 conversions
func Bid128ToUint16Rnint(x BID_UINT128) (uint16, uint32) {
	return bid128_to_small_uint16(Bid128ToUint32Rnint, x, 0xffff0000, 0x8000)
}
func Bid128ToUint16Xrnint(x BID_UINT128) (uint16, uint32) {
	return bid128_to_small_uint16(Bid128ToUint32Xrnint, x, 0xffff0000, 0x8000)
}
func Bid128ToUint16Rninta(x BID_UINT128) (uint16, uint32) {
	return bid128_to_small_uint16(Bid128ToUint32Rninta, x, 0xffff0000, 0x8000)
}
func Bid128ToUint16Xrninta(x BID_UINT128) (uint16, uint32) {
	return bid128_to_small_uint16(Bid128ToUint32Xrninta, x, 0xffff0000, 0x8000)
}
func Bid128ToUint16Int(x BID_UINT128) (uint16, uint32) {
	return bid128_to_small_uint16(Bid128ToUint32Int, x, 0xffff0000, 0x8000)
}
func Bid128ToUint16Xint(x BID_UINT128) (uint16, uint32) {
	return bid128_to_small_uint16(Bid128ToUint32Xint, x, 0xffff0000, 0x8000)
}
func Bid128ToUint16Floor(x BID_UINT128) (uint16, uint32) {
	return bid128_to_small_uint16(Bid128ToUint32Floor, x, 0xffff0000, 0x8000)
}
func Bid128ToUint16Xfloor(x BID_UINT128) (uint16, uint32) {
	return bid128_to_small_uint16(Bid128ToUint32Xfloor, x, 0xffff0000, 0x8000)
}
func Bid128ToUint16Ceil(x BID_UINT128) (uint16, uint32) {
	return bid128_to_small_uint16(Bid128ToUint32Ceil, x, 0xffff0000, 0x8000)
}
func Bid128ToUint16Xceil(x BID_UINT128) (uint16, uint32) {
	return bid128_to_small_uint16(Bid128ToUint32Xceil, x, 0xffff0000, 0x8000)
}

// ============================================================================
// llrint, lrint, llround, lround (wrappers)
// ============================================================================

// Bid128Llrint: bid128_llrint - round to int64 per rounding mode
func Bid128Llrint(x BID_UINT128, rnd_mode int) (int64, uint32) {
	switch rnd_mode {
	case BID_ROUNDING_TO_NEAREST:
		return Bid128ToInt64Xrnint(x)
	case BID_ROUNDING_TIES_AWAY:
		return Bid128ToInt64Xrninta(x)
	case BID_ROUNDING_DOWN:
		return Bid128ToInt64Xfloor(x)
	case BID_ROUNDING_UP:
		return Bid128ToInt64Xceil(x)
	default: // BID_ROUNDING_TO_ZERO
		return Bid128ToInt64Xint(x)
	}
}

// Bid128Lrint: bid128_lrint - same as llrint on 64-bit platform
func Bid128Lrint(x BID_UINT128, rnd_mode int) (int64, uint32) {
	return Bid128Llrint(x, rnd_mode)
}

// Bid128Llround: bid128_llround - round to int64 using rninta
func Bid128Llround(x BID_UINT128) (int64, uint32) {
	return Bid128ToInt64Rninta(x)
}

// Bid128Lround: bid128_lround - same as llround on 64-bit platform
func Bid128Lround(x BID_UINT128) (int64, uint32) {
	return Bid128Llround(x)
}
