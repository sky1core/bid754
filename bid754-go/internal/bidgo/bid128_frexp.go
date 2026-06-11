// Ported from: Intel bid128_frexp.c
// Mechanical translation - all logic preserved exactly.

package bidgo

import "math"

// MASK_EXP_128 and MASK_COEFF_128 are defined in bid128_fma.go
const (
	MASK_EXP2_128 = 0x1fff800000000000 // MASK_EXP2 for 128-bit
)

// Bid128Frexp returns the value res, such that res has a magnitude in the
// interval [1/10, 1) or zero, and x = res*10^*exp. If x is zero, both parts
// of the result are zero. frexp does not raise any exceptions.
// Ported mechanically from Intel bid128_frexp.c.
func Bid128Frexp(x BID_UINT128) (BID_UINT128, int) {
	var res BID_UINT128
	var sig_x BID_UINT128
	var exp_x uint32
	var x_nr_bits int
	var q int

	if (x.w[1] & INFINITY_MASK64) >= INFINITY_MASK64 {
		// if NaN or infinity
		exp := 0
		res = x
		// the binary frexp quitetizes SNaNs, so do the same
		if (x.w[1] & SNAN_MASK64) == SNAN_MASK64 { // x is SNAN
			// return quiet (x)
			res.w[1] = x.w[1] & 0xfdffffffffffffff
		}
		return res, exp
	} else {
		// x is 0, non-canonical, normal, or subnormal
		// check for non-canonical values with 114 bit-significands; can be zero too
		if (x.w[1] & 0x6000000000000000) == 0x6000000000000000 {
			exp := 0
			exp_x = uint32((x.w[1] & MASK_EXP2_128) >> 47) // biased
			res.w[1] = (x.w[1] & 0x8000000000000000) | (uint64(exp_x) << 49)
			// zero of same sign
			res.w[0] = 0x0000000000000000
			return res, exp
		}
		// unpack x
		exp_x = uint32((x.w[1] & MASK_EXP_128) >> 49) // biased
		sig_x.w[1] = x.w[1] & MASK_COEFF_128
		sig_x.w[0] = x.w[0]
		// check for non-canonical values or zero
		if (sig_x.w[1] > 0x0001ed09bead87c0) ||
			(sig_x.w[1] == 0x0001ed09bead87c0 &&
				(sig_x.w[0] > 0x378d8e63ffffffff)) ||
			((sig_x.w[1] == 0x0) && (sig_x.w[0] == 0x0)) {
			exp := 0
			res.w[1] = (x.w[1] & 0x8000000000000000) | (uint64(exp_x) << 49)
			// zero of same sign
			res.w[0] = 0x0000000000000000
			return res, exp
		} else {
			// continue, x is neither zero nor non-canonical
		}
		// x is normal or subnormal, with exp_x=biased exponent & sig_x=coefficient
		// determine the number of decimal digits in sig_x, which fits in 113 bits
		// q = nr. of decimal digits in sig_x (1 <= q <= 34)
		//  determine first the nr. of bits in sig_x
		if sig_x.w[1] == 0 {
			if sig_x.w[0] >= 0x0020000000000000 { // z >= 2^53
				// split the 64-bit value in two 32-bit halves to avoid rounding errors
				if sig_x.w[0] >= 0x0000000100000000 { // z >= 2^32
					tmp_ui64 := math.Float64bits(float64(sig_x.w[0] >> 32)) // exact conversion
					x_nr_bits =
						32 + (int((uint32(tmp_ui64>>52))&0x7ff) - 0x3ff)
				} else { // z < 2^32
					tmp_ui64 := math.Float64bits(float64(sig_x.w[0])) // exact conversion
					x_nr_bits =
						(int((uint32(tmp_ui64>>52))&0x7ff) - 0x3ff)
				}
			} else { // if z < 2^53
				tmp_ui64 := math.Float64bits(float64(sig_x.w[0])) // exact conversion
				x_nr_bits = (int((uint32(tmp_ui64>>52))&0x7ff) - 0x3ff)
			}
		} else { // sig_x.w[1] != 0 => nr. bits = 65 + nr_bits (sig_x.w[1])
			tmp_ui64 := math.Float64bits(float64(sig_x.w[1])) // exact conversion
			x_nr_bits = 64 + (int((uint32(tmp_ui64>>52))&0x7ff) - 0x3ff)
		}
		q = int(bid_nr_digits[x_nr_bits].digits)
		if q == 0 {
			q = int(bid_nr_digits[x_nr_bits].digits1)
			if sig_x.w[1] > bid_nr_digits[x_nr_bits].threshold_hi ||
				(sig_x.w[1] == bid_nr_digits[x_nr_bits].threshold_hi &&
					sig_x.w[0] >= bid_nr_digits[x_nr_bits].threshold_lo) {
				q++
			}
		}
		// Do not add trailing zeros if q < 34; leave sig_x with q digits
		exp := int(exp_x) - 6176 + q
		// assemble the result; sig_x < 2^113 so it fits in 113 bits
		res.w[1] = (x.w[1] & 0x8001ffffffffffff) | (uint64(-q+6176) << 49)
		res.w[0] = x.w[0]
		// replace exponent
		return res, exp
	}
}
