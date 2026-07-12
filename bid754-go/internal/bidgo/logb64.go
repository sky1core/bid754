package bidgo

import "math"

// bid64_logb.c / bid64_logbd.c 기계적 포팅

// Bid64ILogb - Intel bid64_ilogb 기계적 포팅
func Bid64ILogb(x uint64) (int, uint32) {
	var sign_x, coefficient_x uint64
	var exponent_x, bin_expon_cx, digits, res int
	var pfpsf uint32
	var valid_x bool

	// unpack arguments, check for NaN or Infinity
	sign_x, exponent_x, coefficient_x, valid_x = unpack_BID64(x)
	_ = sign_x
	if !valid_x {
		// x is Inf. or NaN
		pfpsf |= BID_INVALID_EXCEPTION
		if (x & 0x7c00000000000000) == 0x7800000000000000 {
			res = 0x7fffffff
		} else {
			res = -2147483648
		}
		return res, pfpsf
	}
	// find number of digits in coefficient
	if coefficient_x >= 1000000000000000 {
		digits = 16
	} else {
		dx := math.Float64bits(float64(coefficient_x)) // exact conversion
		bin_expon_cx = int(dx>>52) - 1023
		digits = bid_estimate_decimal_digits[bin_expon_cx]
		if coefficient_x >= bid_power10_table_128[digits].w[0] {
			digits++
		}
	}
	exponent_x = exponent_x - DECIMAL_EXPONENT_BIAS + digits - 1

	return exponent_x, pfpsf
}

// Bid64Logb - Intel bid64_logb 기계적 포팅
func Bid64Logb(x uint64) (uint64, uint32) {
	var ires, exponent_x int
	var sign_x, coefficient_x uint64
	var valid_x bool
	var res uint64
	var pfpsf uint32

	sign_x, exponent_x, coefficient_x, valid_x = unpack_BID64(x)
	_ = sign_x
	_ = exponent_x

	if !valid_x {
		// test if x is NaN/Inf
		if (x & 0x7800000000000000) == 0x7800000000000000 {
			if (x & 0x7e00000000000000) == 0x7e00000000000000 { // sNaN
				pfpsf |= BID_INVALID_EXCEPTION
			}
			res = coefficient_x & QUIET_MASK64
			if (x & 0x7c00000000000000) == 0x7800000000000000 {
				res &= 0x7fffffffffffffff
			}
			return res, pfpsf
		}
		// x is 0
		pfpsf |= BID_ZERO_DIVIDE_EXCEPTION
		res = 0xf800000000000000
		return res, pfpsf
	}

	ires, _ = Bid64ILogb(x)
	if ires < 0 {
		res = 0xb1c0000000000000 | uint64(-ires)
	} else {
		res = 0x31c0000000000000 | uint64(ires)
	}
	return res, pfpsf
}
