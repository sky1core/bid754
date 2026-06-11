// Ported from: Intel bid32_logb.c, bid32_logbd.c
// Mechanical translation - all logic preserved exactly.

package bidgo

import "math"

// Bid32ILogb is ported mechanically from bid32_logb.c: bid32_ilogb.
func Bid32ILogb(x uint32) (int, uint32) {
	var sign_x, coefficient_x uint32
	var bin_expon_cx, digits, exponent_x int
	var pfpsf uint32

	sign_x, exponent_x, coefficient_x, valid := unpack_BID32(x)
	_ = sign_x
	if !valid {
		pfpsf |= BID_INVALID_EXCEPTION
		var res int
		if (x & 0x7c000000) == 0x78000000 {
			res = 0x7fffffff
		} else {
			res = -0x80000000
		}
		return res, pfpsf
	}
	if coefficient_x >= 1000000 {
		digits = 7
	} else {
		dx := math.Float32bits(float32(coefficient_x))
		bin_expon_cx = int(dx>>23) - 127
		digits = bid_estimate_decimal_digits[bin_expon_cx]
		if uint64(coefficient_x) >= bid_power10_table_128[digits].w[0] {
			digits++
		}
	}
	exponent_x = exponent_x - DECIMAL_EXPONENT_BIAS_32 + digits - 1
	return exponent_x, pfpsf
}

// Bid32Logb is ported mechanically from bid32_logbd.c: bid32_logb.
func Bid32Logb(x uint32) (uint32, uint32) {
	var sign_x, coefficient_x uint32
	var exponent_x int
	var res uint32
	var pfpsf uint32

	sign_x, exponent_x, coefficient_x, valid := unpack_BID32(x)
	_, _ = sign_x, exponent_x

	if !valid {
		if (x & 0x78000000) == 0x78000000 {
			if (x & 0x7e000000) == 0x7e000000 {
				pfpsf |= BID_INVALID_EXCEPTION
			}
			res = coefficient_x & QUIET_MASK32
			if (x & 0x7c000000) == 0x78000000 {
				res &= 0x7fffffff
			}
			return res, pfpsf
		}
		pfpsf |= BID_ZERO_DIVIDE_EXCEPTION
		res = 0xf8000000
		return res, pfpsf
	}

	ires, iflags := Bid32ILogb(x)
	pfpsf |= iflags
	if ires < 0 {
		res = 0xb2800000 | uint32(-ires)
	} else {
		res = 0x32800000 | uint32(ires)
	}
	return res, pfpsf
}
