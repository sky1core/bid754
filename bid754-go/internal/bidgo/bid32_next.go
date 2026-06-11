// Ported from: Intel bid32_next.c
// Mechanical translation - all logic preserved exactly.

package bidgo

import "math"

const P7 = 7
const EXP_MIN32 = 0

// Bid32NextUp is ported mechanically from bid32_next.c: bid32_nextup.
func Bid32NextUp(x uint32) (uint32, uint32) {
	var res uint32
	var x_sign, x_exp uint32
	var x_nr_bits int
	var q1, ind int
	var C1 uint32
	var pfpsf uint32

	if (x & MASK_NAN32) == MASK_NAN32 {
		if (x & 0x000fffff) > 999999 {
			x = x & 0xfe000000
		} else {
			x = x & 0xfe0fffff
		}
		if (x & MASK_SNAN32) == MASK_SNAN32 {
			pfpsf |= BID_INVALID_EXCEPTION
			res = x & 0xfdffffff
		} else {
			res = x
		}
		return res, pfpsf
	} else if (x & MASK_INF32) == MASK_INF32 {
		if (x & 0x80000000) == 0 {
			res = 0x78000000
		} else {
			res = 0xf7f8967f
		}
		return res, pfpsf
	}
	x_sign = x & MASK_SIGN32
	if (x & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
		x_exp = (x & MASK_BINARY_EXPONENT2_32) >> 21
		C1 = (x & MASK_BINARY_SIG2_32) | MASK_BINARY_OR2_32
		if C1 > 9999999 {
			x_exp = 0
			C1 = 0
		}
	} else {
		x_exp = (x & MASK_BINARY_EXPONENT1_32) >> 23
		C1 = x & MASK_BINARY_SIG1_32
	}

	if C1 == 0 {
		res = 0x00000001
	} else {
		if x == 0x77f8967f {
			res = 0x78000000
		} else if x == 0x80000001 {
			res = 0x80000000
		} else {
			tmp1 := math.Float32bits(float32(C1))
			x_nr_bits = 1 + int((tmp1>>23)&0xff) - 0x7f
			q1 = int(bid_nr_digits[x_nr_bits-1].digits)
			if q1 == 0 {
				q1 = int(bid_nr_digits[x_nr_bits-1].digits1)
				if uint64(C1) >= bid_nr_digits[x_nr_bits-1].threshold_lo {
					q1++
				}
			}
			if q1 < P7 {
				if x_exp > uint32(P7-q1) {
					ind = P7 - q1
					C1 = C1 * uint32(bid_ten2k64[ind])
					x_exp = x_exp - uint32(ind)
				} else {
					ind = int(x_exp)
					C1 = C1 * uint32(bid_ten2k64[ind])
					x_exp = EXP_MIN32
				}
			}
			if x_sign == 0 {
				C1++
				if C1 == 0x989680 {
					C1 = 0x0f4240
					x_exp++
				}
			} else {
				C1--
				if C1 == 0x0f423f && x_exp != 0 {
					C1 = 0x98967f
					x_exp--
				}
			}
			if C1&MASK_BINARY_OR2_32 != 0 {
				res = x_sign | (x_exp << 21) | MASK_STEERING_BITS32 | (C1 & MASK_BINARY_SIG2_32)
			} else {
				res = x_sign | (x_exp << 23) | C1
			}
		}
	}
	return res, pfpsf
}

// Bid32NextDown is ported mechanically from bid32_next.c: bid32_nextdown.
// nextdown(x) = -nextup(-x)
func Bid32NextDown(x uint32) (uint32, uint32) {
	res, flags := Bid32NextUp(x ^ 0x80000000)
	return res ^ 0x80000000, flags
}

// Bid32NextAfter is ported mechanically from bid32_next.c: bid32_nextafter.
func Bid32NextAfter(x, y uint32) (uint32, uint32) {
	var res uint32
	var pfpsf uint32

	// NaN handling
	if (x & MASK_NAN32) == MASK_NAN32 {
		if (x & 0x000fffff) > 999999 {
			x = x & 0xfe000000
		} else {
			x = x & 0xfe0fffff
		}
		if (x & MASK_SNAN32) == MASK_SNAN32 {
			pfpsf |= BID_INVALID_EXCEPTION
			res = x & 0xfdffffff
		} else {
			res = x
		}
		if (y & MASK_SNAN32) == MASK_SNAN32 {
			pfpsf |= BID_INVALID_EXCEPTION
		}
		return res, pfpsf
	}
	if (y & MASK_NAN32) == MASK_NAN32 {
		if (y & 0x000fffff) > 999999 {
			y = y & 0xfe000000
		} else {
			y = y & 0xfe0fffff
		}
		if (y & MASK_SNAN32) == MASK_SNAN32 {
			pfpsf |= BID_INVALID_EXCEPTION
			res = y & 0xfdffffff
		} else {
			res = y
		}
		return res, pfpsf
	}

	// compare x and y using quiet comparison
	eqRes, _ := Bid32QuietEqual(x, y)
	if eqRes != 0 {
		res = y
		return res, pfpsf
	}

	lessRes, _ := Bid32QuietGreater(x, y)
	if lessRes != 0 {
		// x > y, next toward y = nextdown
		res, pfpsf = Bid32NextDown(x)
	} else {
		// x < y, next toward y = nextup
		res, pfpsf = Bid32NextUp(x)
	}

	// overflow/underflow checks
	if ((x & MASK_INF32) != MASK_INF32) && ((res & MASK_INF32) == MASK_INF32) {
		pfpsf |= BID_INEXACT_EXCEPTION | BID_OVERFLOW_EXCEPTION
	}
	tmp1 := uint32(0x00784000) // +1E-101 * 10^6
	tmp2 := res & 0x7fffffff
	gtRes, _ := Bid32QuietGreater(tmp1, tmp2)
	neRes, _ := Bid32QuietNotEqual(x, res)
	if gtRes != 0 && neRes != 0 {
		pfpsf |= BID_INEXACT_EXCEPTION | BID_UNDERFLOW_EXCEPTION
	}

	return res, pfpsf
}
