// Ported from: Intel bid128_fma.c (helper functions)
// Mechanical translation - all logic preserved exactly.

package bidgo

import "math"

const expmax = 6111

// bid_add256 adds two 256-bit values.
func bid_add256(x, y BID_UINT256) BID_UINT256 {
	var z BID_UINT256
	z.w[0] = x.w[0] + y.w[0]
	if z.w[0] < x.w[0] {
		x.w[1]++
		if x.w[1] == 0 {
			x.w[2]++
			if x.w[2] == 0 {
				x.w[3]++
			}
		}
	}
	z.w[1] = x.w[1] + y.w[1]
	if z.w[1] < x.w[1] {
		x.w[2]++
		if x.w[2] == 0 {
			x.w[3]++
		}
	}
	z.w[2] = x.w[2] + y.w[2]
	if z.w[2] < x.w[2] {
		x.w[3]++
	}
	z.w[3] = x.w[3] + y.w[3]
	return z
}

// bid_sub256 subtracts y from x (assumes x >= y).
func bid_sub256(x, y BID_UINT256) BID_UINT256 {
	var z BID_UINT256
	z.w[0] = x.w[0] - y.w[0]
	if z.w[0] > x.w[0] {
		x.w[1]--
		if x.w[1] == 0xffffffffffffffff {
			x.w[2]--
			if x.w[2] == 0xffffffffffffffff {
				x.w[3]--
			}
		}
	}
	z.w[1] = x.w[1] - y.w[1]
	if z.w[1] > x.w[1] {
		x.w[2]--
		if x.w[2] == 0xffffffffffffffff {
			x.w[3]--
		}
	}
	z.w[2] = x.w[2] - y.w[2]
	if z.w[2] > x.w[2] {
		x.w[3]--
	}
	z.w[3] = x.w[3] - y.w[3]
	return z
}

// bid_rounding_correction applies rounding correction for bid128.
func bid_rounding_correction(rnd_mode int,
	is_inexact_lt_midpoint, is_inexact_gt_midpoint,
	is_midpoint_lt_even, is_midpoint_gt_even int,
	unbexp int, ptrres *BID_UINT128, ptrfpsf *uint32) {

	res := *ptrres
	var sign, exp uint64
	var C_hi, C_lo uint64

	if is_inexact_lt_midpoint != 0 || is_inexact_gt_midpoint != 0 ||
		is_midpoint_lt_even != 0 || is_midpoint_gt_even != 0 {
		*ptrfpsf |= BID_INEXACT_EXCEPTION
	}

	sign = res.w[1] & MASK_SIGN64
	exp = uint64(unbexp+6176) << 49
	C_hi = res.w[1] & MASK_COEFF128
	C_lo = res.w[0]

	if (sign == 0 && ((rnd_mode == BID_ROUNDING_UP && is_inexact_lt_midpoint != 0) ||
		((rnd_mode == BID_ROUNDING_TIES_AWAY || rnd_mode == BID_ROUNDING_UP) && is_midpoint_gt_even != 0))) ||
		(sign != 0 && ((rnd_mode == BID_ROUNDING_DOWN && is_inexact_lt_midpoint != 0) ||
			((rnd_mode == BID_ROUNDING_TIES_AWAY || rnd_mode == BID_ROUNDING_DOWN) && is_midpoint_gt_even != 0))) {
		C_lo = C_lo + 1
		if C_lo == 0 {
			C_hi = C_hi + 1
		}
		if C_hi == 0x0001ed09bead87c0 && C_lo == 0x378d8e6400000000 {
			C_hi = 0x0000314dc6448d93
			C_lo = 0x38c15b0a00000000
			unbexp = unbexp + 1
			exp = uint64(unbexp+6176) << 49
		}
	} else if (is_midpoint_lt_even != 0 || is_inexact_gt_midpoint != 0) &&
		((sign != 0 && (rnd_mode == BID_ROUNDING_UP || rnd_mode == BID_ROUNDING_TO_ZERO)) ||
			(sign == 0 && (rnd_mode == BID_ROUNDING_DOWN || rnd_mode == BID_ROUNDING_TO_ZERO))) {
		C_lo = C_lo - 1
		if C_lo == 0xffffffffffffffff {
			C_hi--
		}
		if C_hi == 0x0000314dc6448d93 && C_lo == 0x38c15b09ffffffff {
			if exp > 0 {
				C_hi = 0x0001ed09bead87c0
				C_lo = 0x378d8e63ffffffff
				unbexp = unbexp - 1
				exp = uint64(unbexp+6176) << 49
			} else {
				*ptrfpsf |= BID_UNDERFLOW_EXCEPTION
			}
		}
	}

	if unbexp > expmax {
		*ptrfpsf |= (BID_INEXACT_EXCEPTION | BID_OVERFLOW_EXCEPTION)
		exp = 0
		if sign == 0 {
			if rnd_mode == BID_ROUNDING_UP || rnd_mode == BID_ROUNDING_TIES_AWAY {
				C_hi = 0x7800000000000000
				C_lo = 0x0000000000000000
			} else {
				C_hi = 0x5fffed09bead87c0
				C_lo = 0x378d8e63ffffffff
			}
		} else {
			if rnd_mode == BID_ROUNDING_DOWN || rnd_mode == BID_ROUNDING_TIES_AWAY {
				C_hi = 0xf800000000000000
				C_lo = 0x0000000000000000
			} else {
				C_hi = 0xdfffed09bead87c0
				C_lo = 0x378d8e63ffffffff
			}
		}
	}

	res.w[1] = sign | exp | C_hi
	res.w[0] = C_lo
	*ptrres = res
}

// bid128_count_digits counts decimal digits in a 128-bit coefficient.
// Ported from the digit-counting pattern in bid128_fma.c.
func bid128_count_digits(C BID_UINT128) int {
	var x_nr_bits int
	if C.w[1] == 0 {
		if C.w[0] == 0 {
			return 0
		}
		if C.w[0] >= 0x0020000000000000 {
			tmp := math.Float64bits(float64(C.w[0] >> 32))
			x_nr_bits = 33 + int(((tmp>>52)&0x7ff)-0x3ff)
		} else {
			tmp := math.Float64bits(float64(C.w[0]))
			x_nr_bits = 1 + int(((tmp>>52)&0x7ff)-0x3ff)
		}
	} else {
		tmp := math.Float64bits(float64(C.w[1]))
		x_nr_bits = 65 + int(((tmp>>52)&0x7ff)-0x3ff)
	}
	q := int(bid_nr_digits[x_nr_bits-1].digits)
	if q == 0 {
		q = int(bid_nr_digits[x_nr_bits-1].digits1)
		if C.w[1] > bid_nr_digits[x_nr_bits-1].threshold_hi ||
			(C.w[1] == bid_nr_digits[x_nr_bits-1].threshold_hi &&
				C.w[0] >= bid_nr_digits[x_nr_bits-1].threshold_lo) {
			q++
		}
	}
	return q
}
