// Ported from: Intel bid128_noncomp.c
// Mechanical translation - all logic preserved exactly.

package bidgo

import "math"

// Bid128IsSigned returns 1 if x has sign bit set.
func Bid128IsSigned(x BID_UINT128) int {
	if (x.hi & MASK_SIGN64) == MASK_SIGN64 {
		return 1
	}
	return 0
}

// Bid128IsNormal returns 1 if x is a normal number.
func Bid128IsNormal(x BID_UINT128) int {
	if (x.hi & MASK_SPECIAL128) == MASK_SPECIAL128 {
		return 0
	}
	C1_hi := x.hi & MASK_COEFF128
	C1_lo := x.lo
	if C1_hi == 0 && C1_lo == 0 {
		return 0
	}
	if ((C1_hi > 0x0001ed09bead87c0 || (C1_hi == 0x0001ed09bead87c0 && C1_lo > 0x378d8e63ffffffff)) &&
		((x.hi & 0x6000000000000000) != 0x6000000000000000)) ||
		((x.hi & 0x6000000000000000) == 0x6000000000000000) {
		return 0
	}

	var x_nr_bits int
	if C1_hi == 0 {
		if C1_lo >= 0x0020000000000000 {
			tmp1 := math.Float64bits(float64(C1_lo >> 32))
			x_nr_bits = 33 + int(((tmp1>>52)&0x7ff)-0x3ff)
		} else {
			tmp1 := math.Float64bits(float64(C1_lo))
			x_nr_bits = 1 + int(((tmp1>>52)&0x7ff)-0x3ff)
		}
	} else {
		tmp1 := math.Float64bits(float64(C1_hi))
		x_nr_bits = 65 + int(((tmp1>>52)&0x7ff)-0x3ff)
	}
	q := int(bid_nr_digits[x_nr_bits-1].digits)
	if q == 0 {
		q = int(bid_nr_digits[x_nr_bits-1].digits1)
		if C1_hi > bid_nr_digits[x_nr_bits-1].threshold_hi ||
			(C1_hi == bid_nr_digits[x_nr_bits-1].threshold_hi &&
				C1_lo >= bid_nr_digits[x_nr_bits-1].threshold_lo) {
			q++
		}
	}
	x_exp := x.hi & MASK_EXP128
	exp := int(x_exp>>49) - 6176
	if exp+q <= -6143 {
		return 0
	}
	return 1
}

// Bid128IsSubnormal returns 1 if x is a subnormal number.
func Bid128IsSubnormal(x BID_UINT128) int {
	if (x.hi & MASK_SPECIAL128) == MASK_SPECIAL128 {
		return 0
	}
	C1_hi := x.hi & MASK_COEFF128
	C1_lo := x.lo
	if C1_hi == 0 && C1_lo == 0 {
		return 0
	}
	if ((C1_hi > 0x0001ed09bead87c0 || (C1_hi == 0x0001ed09bead87c0 && C1_lo > 0x378d8e63ffffffff)) &&
		((x.hi & 0x6000000000000000) != 0x6000000000000000)) ||
		((x.hi & 0x6000000000000000) == 0x6000000000000000) {
		return 0
	}

	var x_nr_bits int
	if C1_hi == 0 {
		if C1_lo >= 0x0020000000000000 {
			tmp1 := math.Float64bits(float64(C1_lo >> 32))
			x_nr_bits = 33 + int(((tmp1>>52)&0x7ff)-0x3ff)
		} else {
			tmp1 := math.Float64bits(float64(C1_lo))
			x_nr_bits = 1 + int(((tmp1>>52)&0x7ff)-0x3ff)
		}
	} else {
		tmp1 := math.Float64bits(float64(C1_hi))
		x_nr_bits = 65 + int(((tmp1>>52)&0x7ff)-0x3ff)
	}
	q := int(bid_nr_digits[x_nr_bits-1].digits)
	if q == 0 {
		q = int(bid_nr_digits[x_nr_bits-1].digits1)
		if C1_hi > bid_nr_digits[x_nr_bits-1].threshold_hi ||
			(C1_hi == bid_nr_digits[x_nr_bits-1].threshold_hi &&
				C1_lo >= bid_nr_digits[x_nr_bits-1].threshold_lo) {
			q++
		}
	}
	x_exp := x.hi & MASK_EXP128
	exp := int(x_exp>>49) - 6176
	if exp+q <= -6143 {
		return 1
	}
	return 0
}

// Bid128IsFinite returns 1 if x is finite.
func Bid128IsFinite(x BID_UINT128) int {
	if (x.hi & MASK_INF64) != MASK_INF64 {
		return 1
	}
	return 0
}

// Bid128IsSignaling returns 1 if x is a signaling NaN.
func Bid128IsSignaling(x BID_UINT128) int {
	if (x.hi & SNAN_MASK64) == SNAN_MASK64 {
		return 1
	}
	return 0
}

// Bid128IsCanonical returns 1 if x is canonical.
func Bid128IsCanonical(x BID_UINT128) int {
	if (x.hi & NAN_MASK64) == NAN_MASK64 { // NaN
		if (x.hi & 0x01ffc00000000000) != 0 {
			return 0
		}
		sig_x_hi := x.hi & 0x00003fffffffffff // 46 bits
		sig_x_lo := x.lo                      // 64 bits
		// payload must be < 10^33 = 0x0000314dc6448d93_38c15b0a00000000
		if sig_x_hi < 0x0000314dc6448d93 ||
			(sig_x_hi == 0x0000314dc6448d93 && sig_x_lo < 0x38c15b0a00000000) {
			return 1
		}
		return 0
	}
	if (x.hi & INFINITY_MASK64) == INFINITY_MASK64 { // infinity
		if (x.hi&0x03ffffffffffffff) != 0 || x.lo != 0 {
			return 0
		}
		return 1
	}
	// not NaN or infinity; extract significand to ensure it is canonical
	sig_x_hi := x.hi & 0x0001ffffffffffff
	sig_x_lo := x.lo
	if (sig_x_hi > 0x0001ed09bead87c0) ||
		(sig_x_hi == 0x0001ed09bead87c0 && sig_x_lo > 0x378d8e63ffffffff) ||
		((x.hi & 0x6000000000000000) == 0x6000000000000000) {
		return 0
	}
	return 1
}

// Bid128Copy returns a copy of x.
func Bid128Copy(x BID_UINT128) BID_UINT128 {
	return x
}

// Bid128Negate returns -x.
func Bid128Negate(x BID_UINT128) BID_UINT128 {
	return BID_UINT128{lo: x.lo, hi: x.hi ^ MASK_SIGN64}
}

// Bid128Abs returns |x|.
func Bid128Abs(x BID_UINT128) BID_UINT128 {
	return BID_UINT128{lo: x.lo, hi: x.hi & 0x7fffffffffffffff}
}

// Bid128CopySign returns x with the sign of y.
func Bid128CopySign(x, y BID_UINT128) BID_UINT128 {
	return BID_UINT128{lo: x.lo, hi: (x.hi & 0x7fffffffffffffff) | (y.hi & MASK_SIGN64)}
}

// Bid128Radix returns 10.
func Bid128Radix() int {
	return 10
}

// Bid128Inf returns +Infinity.
// Ported from bid128_noncomp.c: bid128_inf.
func Bid128Inf() BID_UINT128 {
	return BID_UINT128{lo: 0x0000000000000000, hi: 0x7800000000000000}
}

// Bid128NaN returns +QNaN with optional payload from tagp string.
// Ported from bid128_noncomp.c: bid128_nan.
func Bid128NaN(tagp string) BID_UINT128 {
	res := BID_UINT128{lo: 0x0000000000000000, hi: 0x7c00000000000000}
	if tagp == "" {
		return res
	}
	x, _ := Bid128FromString(tagp, BID_ROUNDING_TO_NEAREST)
	x.hi = x.hi & 0x00003fffffffffff
	res.hi = res.hi | x.hi
	res.lo = x.lo
	return res
}

// Bid128SameQuantum returns 1 if x and y have the same quantum.
func Bid128SameQuantum(x, y BID_UINT128) int {
	if (x.hi&MASK_SPECIAL128) == MASK_SPECIAL128 || (y.hi&MASK_SPECIAL128) == MASK_SPECIAL128 {
		if (x.hi&MASK_SPECIAL128) == MASK_SPECIAL128 && (y.hi&MASK_SPECIAL128) == MASK_SPECIAL128 {
			if (x.hi&NAN_MASK64) == NAN_MASK64 || (y.hi&NAN_MASK64) == NAN_MASK64 {
				if (x.hi&NAN_MASK64) == NAN_MASK64 && (y.hi&NAN_MASK64) == NAN_MASK64 {
					return 1
				}
				return 0
			}
			return 1 // both infinity
		}
		return 0
	}
	var exp_x, exp_y uint64
	if (x.hi & 0x6000000000000000) == 0x6000000000000000 {
		exp_x = (x.hi << 2) & MASK_EXP128
	} else {
		exp_x = x.hi & MASK_EXP128
	}
	if (y.hi & 0x6000000000000000) == 0x6000000000000000 {
		exp_y = (y.hi << 2) & MASK_EXP128
	} else {
		exp_y = y.hi & MASK_EXP128
	}
	if exp_x == exp_y {
		return 1
	}
	return 0
}
