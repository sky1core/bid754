// Ported from: Intel bid128_nearbyintd.c
// Mechanical translation - all logic preserved exactly.

package bidgo

import "math"

// Bid128Nearbyint: bid128_nearbyint
func Bid128Nearbyint(x BID_UINT128, rnd_mode int) (BID_UINT128, uint32) {
	var res BID_UINT128
	var x_sign uint64
	var x_exp uint64
	var exp int
	var tmp64 uint64
	var x_nr_bits uint
	var q, ind, shift int
	var C1 BID_UINT128
	var fstar BID_UINT256
	var P256 BID_UINT256
	var pfpsf uint32

	// check for NaN or Infinity
	if (x.hi & MASK_SPECIAL128) == MASK_SPECIAL128 {
		// x is special
		if (x.hi & NAN_MASK64) == NAN_MASK64 { // x is NAN
			// check first for non-canonical NaN payload
			if ((x.hi & 0x00003fffffffffff) > 0x0000314dc6448d93) ||
				((x.hi&0x00003fffffffffff) == 0x0000314dc6448d93 && x.lo > 0x38c15b09ffffffff) {
				x.hi = x.hi & 0xffffc00000000000
				x.lo = 0x0
			}
			if (x.hi & SNAN_MASK64) == SNAN_MASK64 { // x is SNAN
				pfpsf |= BID_INVALID_EXCEPTION
				res.hi = x.hi & 0xfc003fffffffffff
				res.lo = x.lo
			} else { // x is QNaN
				res.hi = x.hi & 0xfc003fffffffffff
				res.lo = x.lo
			}
			return res, pfpsf
		} else { // x is not a NaN, so it must be infinity
			if (x.hi & MASK_SIGN64) == 0x0 { // x is +inf
				res.hi = 0x7800000000000000
				res.lo = 0x0000000000000000
			} else { // x is -inf
				res.hi = 0xf800000000000000
				res.lo = 0x0000000000000000
			}
			return res, pfpsf
		}
	}
	// unpack x
	x_sign = x.hi & MASK_SIGN64
	C1.hi = x.hi & MASK_COEFF128
	C1.lo = x.lo

	// check for non-canonical values (treated as zero)
	if (x.hi & 0x6000000000000000) == 0x6000000000000000 { // G0_G1=11
		x_exp = (x.hi << 2) & MASK_EXP128
		C1.hi = 0
		C1.lo = 0
	} else {
		x_exp = x.hi & MASK_EXP128
		if C1.hi > 0x0001ed09bead87c0 || (C1.hi == 0x0001ed09bead87c0 && C1.lo > 0x378d8e63ffffffff) {
			C1.hi = 0
			C1.lo = 0
		}
	}

	// test for input equal to zero
	if C1.hi == 0x0 && C1.lo == 0x0 {
		if x_exp <= (0x1820 << 49) {
			res.hi = (x.hi & 0x8000000000000000) | 0x3040000000000000
		} else {
			res.hi = x_sign | x_exp
		}
		res.lo = 0x0000000000000000
		return res, pfpsf
	}

	// x is not special and is not zero
	switch rnd_mode {
	case BID_ROUNDING_TO_NEAREST, BID_ROUNDING_TIES_AWAY:
		if x_exp <= 0x2ffa000000000000 { // exp <= -35
			res.hi = x_sign | 0x3040000000000000
			res.lo = 0x0000000000000000
			return res, pfpsf
		}
	case BID_ROUNDING_DOWN:
		if x_exp <= 0x2ffc000000000000 { // exp <= -34
			if x_sign != 0 {
				res.hi = 0xb040000000000000
				res.lo = 0x0000000000000001
			} else {
				res.hi = 0x3040000000000000
				res.lo = 0x0000000000000000
			}
			return res, pfpsf
		}
	case BID_ROUNDING_UP:
		if x_exp <= 0x2ffc000000000000 {
			if x_sign != 0 {
				res.hi = 0xb040000000000000
				res.lo = 0x0000000000000000
			} else {
				res.hi = 0x3040000000000000
				res.lo = 0x0000000000000001
			}
			return res, pfpsf
		}
	case BID_ROUNDING_TO_ZERO:
		if x_exp <= 0x2ffc000000000000 {
			res.hi = x_sign | 0x3040000000000000
			res.lo = 0x0000000000000000
			return res, pfpsf
		}
	}

	// q = nr. of decimal digits in x
	var tmp1 uint64
	if C1.hi == 0 {
		if C1.lo >= 0x0020000000000000 {
			tmp1 = math.Float64bits(float64(C1.lo >> 32))
			x_nr_bits = 33 + uint((uint32(tmp1>>52)&0x7ff)-0x3ff)
		} else {
			tmp1 = math.Float64bits(float64(C1.lo))
			x_nr_bits = 1 + uint((uint32(tmp1>>52)&0x7ff)-0x3ff)
		}
	} else {
		tmp1 = math.Float64bits(float64(C1.hi))
		x_nr_bits = 65 + uint((uint32(tmp1>>52)&0x7ff)-0x3ff)
	}

	q = int(bid_nr_digits[x_nr_bits-1].digits)
	if q == 0 {
		q = int(bid_nr_digits[x_nr_bits-1].digits1)
		if C1.hi > bid_nr_digits[x_nr_bits-1].threshold_hi ||
			(C1.hi == bid_nr_digits[x_nr_bits-1].threshold_hi &&
				C1.lo >= bid_nr_digits[x_nr_bits-1].threshold_lo) {
			q++
		}
	}
	exp = int(x_exp>>49) - 6176
	if exp >= 0 {
		// the argument is an integer already
		res.hi = x.hi
		res.lo = x.lo
		return res, pfpsf
	}

	// exp < 0
	switch rnd_mode {
	case BID_ROUNDING_TO_NEAREST:
		if q+exp >= 0 {
			ind = -exp
			tmp64 = C1.lo
			if ind <= 19 {
				C1.lo = C1.lo + bid_midpoint64[ind-1]
			} else {
				C1.lo = C1.lo + bid_midpoint128[ind-20].lo
				C1.hi = C1.hi + bid_midpoint128[ind-20].hi
			}
			if C1.lo < tmp64 {
				C1.hi++
			}
			P256 = __mul_128x128_to_256(C1, bid_ten2mk128[ind-1])

			if ind-1 <= 2 {
				res.hi = P256.w3
				res.lo = P256.w2
				fstar.w1 = P256.w1
				fstar.w0 = P256.w0
				if (res.lo&0x0000000000000001 != 0) &&
					((fstar.w1 < bid_ten2mk128[ind-1].hi) ||
						(fstar.w1 == bid_ten2mk128[ind-1].hi &&
							fstar.w0 < bid_ten2mk128[ind-1].lo)) {
					res.lo--
				}
			} else if ind-1 <= 21 {
				shift = bid_shiftright128[ind-1]
				res.hi = P256.w3 >> uint(shift)
				res.lo = (P256.w3 << uint(64-shift)) | (P256.w2 >> uint(shift))
				fstar.w2 = P256.w2 & bid_maskhigh128[ind-1]
				fstar.w1 = P256.w1
				fstar.w0 = P256.w0
				if (res.lo&0x0000000000000001 != 0) &&
					fstar.w2 == 0 && (fstar.w1 < bid_ten2mk128[ind-1].hi ||
					(fstar.w1 == bid_ten2mk128[ind-1].hi && fstar.w0 < bid_ten2mk128[ind-1].lo)) {
					res.lo--
				}
			} else {
				shift = bid_shiftright128[ind-1] - 64
				res.hi = 0
				res.lo = P256.w3 >> uint(shift)
				fstar.w3 = P256.w3 & bid_maskhigh128[ind-1]
				fstar.w2 = P256.w2
				fstar.w1 = P256.w1
				fstar.w0 = P256.w0
				if (res.lo&0x0000000000000001 != 0) &&
					fstar.w3 == 0 && fstar.w2 == 0 && (fstar.w1 < bid_ten2mk128[ind-1].hi ||
					(fstar.w1 == bid_ten2mk128[ind-1].hi && fstar.w0 < bid_ten2mk128[ind-1].lo)) {
					res.lo--
				}
			}
			res.hi = x_sign | 0x3040000000000000 | res.hi
			return res, pfpsf
		} else {
			res.hi = x_sign | 0x3040000000000000
			res.lo = 0x0000000000000000
			return res, pfpsf
		}

	case BID_ROUNDING_TIES_AWAY:
		if q+exp >= 0 {
			ind = -exp
			tmp64 = C1.lo
			if ind <= 19 {
				C1.lo = C1.lo + bid_midpoint64[ind-1]
			} else {
				C1.lo = C1.lo + bid_midpoint128[ind-20].lo
				C1.hi = C1.hi + bid_midpoint128[ind-20].hi
			}
			if C1.lo < tmp64 {
				C1.hi++
			}
			P256 = __mul_128x128_to_256(C1, bid_ten2mk128[ind-1])

			if ind-1 <= 2 {
				res.hi = P256.w3
				res.lo = P256.w2
			} else if ind-1 <= 21 {
				shift = bid_shiftright128[ind-1]
				res.hi = P256.w3 >> uint(shift)
				res.lo = (P256.w3 << uint(64-shift)) | (P256.w2 >> uint(shift))
			} else {
				shift = bid_shiftright128[ind-1] - 64
				res.hi = 0
				res.lo = P256.w3 >> uint(shift)
			}
			res.hi |= x_sign | 0x3040000000000000
			return res, pfpsf
		} else {
			res.hi = x_sign | 0x3040000000000000
			res.lo = 0x0000000000000000
			return res, pfpsf
		}

	case BID_ROUNDING_DOWN:
		if q+exp > 0 {
			ind = -exp
			P256 = __mul_128x128_to_256(C1, bid_ten2mk128[ind-1])
			if ind-1 <= 2 {
				res.hi = P256.w3
				res.lo = P256.w2
				if (P256.w1 > bid_ten2mk128[ind-1].hi) ||
					(P256.w1 == bid_ten2mk128[ind-1].hi && P256.w0 >= bid_ten2mk128[ind-1].lo) {
					if x_sign != 0 {
						res.lo++
						if res.lo == 0 {
							res.hi++
						}
					}
				}
			} else if ind-1 <= 21 {
				shift = bid_shiftright128[ind-1]
				res.hi = P256.w3 >> uint(shift)
				res.lo = (P256.w3 << uint(64-shift)) | (P256.w2 >> uint(shift))
				fstar.w2 = P256.w2 & bid_maskhigh128[ind-1]
				fstar.w1 = P256.w1
				fstar.w0 = P256.w0
				if fstar.w2 != 0 || fstar.w1 > bid_ten2mk128[ind-1].hi ||
					(fstar.w1 == bid_ten2mk128[ind-1].hi && fstar.w0 >= bid_ten2mk128[ind-1].lo) {
					if x_sign != 0 {
						res.lo++
						if res.lo == 0 {
							res.hi++
						}
					}
				}
			} else {
				shift = bid_shiftright128[ind-1] - 64
				res.hi = 0
				res.lo = P256.w3 >> uint(shift)
				fstar.w3 = P256.w3 & bid_maskhigh128[ind-1]
				fstar.w2 = P256.w2
				fstar.w1 = P256.w1
				fstar.w0 = P256.w0
				if fstar.w3 != 0 || fstar.w2 != 0 ||
					fstar.w1 > bid_ten2mk128[ind-1].hi ||
					(fstar.w1 == bid_ten2mk128[ind-1].hi && fstar.w0 >= bid_ten2mk128[ind-1].lo) {
					if x_sign != 0 {
						res.lo++
						if res.lo == 0 {
							res.hi++
						}
					}
				}
			}
			res.hi = x_sign | 0x3040000000000000 | res.hi
			return res, pfpsf
		} else {
			if x_sign != 0 {
				res.hi = 0xb040000000000000
				res.lo = 0x0000000000000001
			} else {
				res.hi = 0x3040000000000000
				res.lo = 0x0000000000000000
			}
			return res, pfpsf
		}

	case BID_ROUNDING_UP:
		if q+exp > 0 {
			ind = -exp
			P256 = __mul_128x128_to_256(C1, bid_ten2mk128[ind-1])
			if ind-1 <= 2 {
				res.hi = P256.w3
				res.lo = P256.w2
				if (P256.w1 > bid_ten2mk128[ind-1].hi) ||
					(P256.w1 == bid_ten2mk128[ind-1].hi && P256.w0 >= bid_ten2mk128[ind-1].lo) {
					if x_sign == 0 {
						res.lo++
						if res.lo == 0 {
							res.hi++
						}
					}
				}
			} else if ind-1 <= 21 {
				shift = bid_shiftright128[ind-1]
				res.hi = P256.w3 >> uint(shift)
				res.lo = (P256.w3 << uint(64-shift)) | (P256.w2 >> uint(shift))
				fstar.w2 = P256.w2 & bid_maskhigh128[ind-1]
				fstar.w1 = P256.w1
				fstar.w0 = P256.w0
				if fstar.w2 != 0 || fstar.w1 > bid_ten2mk128[ind-1].hi ||
					(fstar.w1 == bid_ten2mk128[ind-1].hi && fstar.w0 >= bid_ten2mk128[ind-1].lo) {
					if x_sign == 0 {
						res.lo++
						if res.lo == 0 {
							res.hi++
						}
					}
				}
			} else {
				shift = bid_shiftright128[ind-1] - 64
				res.hi = 0
				res.lo = P256.w3 >> uint(shift)
				fstar.w3 = P256.w3 & bid_maskhigh128[ind-1]
				fstar.w2 = P256.w2
				fstar.w1 = P256.w1
				fstar.w0 = P256.w0
				if fstar.w3 != 0 || fstar.w2 != 0 ||
					fstar.w1 > bid_ten2mk128[ind-1].hi ||
					(fstar.w1 == bid_ten2mk128[ind-1].hi && fstar.w0 >= bid_ten2mk128[ind-1].lo) {
					if x_sign == 0 {
						res.lo++
						if res.lo == 0 {
							res.hi++
						}
					}
				}
			}
			res.hi = x_sign | 0x3040000000000000 | res.hi
			return res, pfpsf
		} else {
			if x_sign != 0 {
				res.hi = 0xb040000000000000
				res.lo = 0x0000000000000000
			} else {
				res.hi = 0x3040000000000000
				res.lo = 0x0000000000000001
			}
			return res, pfpsf
		}

	case BID_ROUNDING_TO_ZERO:
		if q+exp > 0 {
			ind = -exp
			P256 = __mul_128x128_to_256(C1, bid_ten2mk128[ind-1])
			if ind-1 <= 2 {
				res.hi = P256.w3
				res.lo = P256.w2
			} else if ind-1 <= 21 {
				shift = bid_shiftright128[ind-1]
				res.hi = P256.w3 >> uint(shift)
				res.lo = (P256.w3 << uint(64-shift)) | (P256.w2 >> uint(shift))
			} else {
				shift = bid_shiftright128[ind-1] - 64
				res.hi = 0
				res.lo = P256.w3 >> uint(shift)
			}
			res.hi = x_sign | 0x3040000000000000 | res.hi
			return res, pfpsf
		} else {
			res.hi = x_sign | 0x3040000000000000
			res.lo = 0x0000000000000000
			return res, pfpsf
		}
	}

	return res, pfpsf
}
