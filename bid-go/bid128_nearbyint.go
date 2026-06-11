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
	if (x.w[1] & MASK_SPECIAL128) == MASK_SPECIAL128 {
		// x is special
		if (x.w[1] & NAN_MASK64) == NAN_MASK64 { // x is NAN
			// check first for non-canonical NaN payload
			if ((x.w[1] & 0x00003fffffffffff) > 0x0000314dc6448d93) ||
				((x.w[1]&0x00003fffffffffff) == 0x0000314dc6448d93 && x.w[0] > 0x38c15b09ffffffff) {
				x.w[1] = x.w[1] & 0xffffc00000000000
				x.w[0] = 0x0
			}
			if (x.w[1] & SNAN_MASK64) == SNAN_MASK64 { // x is SNAN
				pfpsf |= BID_INVALID_EXCEPTION
				res.w[1] = x.w[1] & 0xfc003fffffffffff
				res.w[0] = x.w[0]
			} else { // x is QNaN
				res.w[1] = x.w[1] & 0xfc003fffffffffff
				res.w[0] = x.w[0]
			}
			return res, pfpsf
		} else { // x is not a NaN, so it must be infinity
			if (x.w[1] & MASK_SIGN64) == 0x0 { // x is +inf
				res.w[1] = 0x7800000000000000
				res.w[0] = 0x0000000000000000
			} else { // x is -inf
				res.w[1] = 0xf800000000000000
				res.w[0] = 0x0000000000000000
			}
			return res, pfpsf
		}
	}
	// unpack x
	x_sign = x.w[1] & MASK_SIGN64
	C1.w[1] = x.w[1] & MASK_COEFF128
	C1.w[0] = x.w[0]

	// check for non-canonical values (treated as zero)
	if (x.w[1] & 0x6000000000000000) == 0x6000000000000000 { // G0_G1=11
		x_exp = (x.w[1] << 2) & MASK_EXP128
		C1.w[1] = 0
		C1.w[0] = 0
	} else {
		x_exp = x.w[1] & MASK_EXP128
		if C1.w[1] > 0x0001ed09bead87c0 || (C1.w[1] == 0x0001ed09bead87c0 && C1.w[0] > 0x378d8e63ffffffff) {
			C1.w[1] = 0
			C1.w[0] = 0
		}
	}

	// test for input equal to zero
	if C1.w[1] == 0x0 && C1.w[0] == 0x0 {
		if x_exp <= (0x1820 << 49) {
			res.w[1] = (x.w[1] & 0x8000000000000000) | 0x3040000000000000
		} else {
			res.w[1] = x_sign | x_exp
		}
		res.w[0] = 0x0000000000000000
		return res, pfpsf
	}

	// x is not special and is not zero
	switch rnd_mode {
	case BID_ROUNDING_TO_NEAREST, BID_ROUNDING_TIES_AWAY:
		if x_exp <= 0x2ffa000000000000 { // exp <= -35
			res.w[1] = x_sign | 0x3040000000000000
			res.w[0] = 0x0000000000000000
			return res, pfpsf
		}
	case BID_ROUNDING_DOWN:
		if x_exp <= 0x2ffc000000000000 { // exp <= -34
			if x_sign != 0 {
				res.w[1] = 0xb040000000000000
				res.w[0] = 0x0000000000000001
			} else {
				res.w[1] = 0x3040000000000000
				res.w[0] = 0x0000000000000000
			}
			return res, pfpsf
		}
	case BID_ROUNDING_UP:
		if x_exp <= 0x2ffc000000000000 {
			if x_sign != 0 {
				res.w[1] = 0xb040000000000000
				res.w[0] = 0x0000000000000000
			} else {
				res.w[1] = 0x3040000000000000
				res.w[0] = 0x0000000000000001
			}
			return res, pfpsf
		}
	case BID_ROUNDING_TO_ZERO:
		if x_exp <= 0x2ffc000000000000 {
			res.w[1] = x_sign | 0x3040000000000000
			res.w[0] = 0x0000000000000000
			return res, pfpsf
		}
	}

	// q = nr. of decimal digits in x
	var tmp1 uint64
	if C1.w[1] == 0 {
		if C1.w[0] >= 0x0020000000000000 {
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
	exp = int(x_exp>>49) - 6176
	if exp >= 0 {
		// the argument is an integer already
		res.w[1] = x.w[1]
		res.w[0] = x.w[0]
		return res, pfpsf
	}

	// exp < 0
	switch rnd_mode {
	case BID_ROUNDING_TO_NEAREST:
		if q+exp >= 0 {
			ind = -exp
			tmp64 = C1.w[0]
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

			if ind-1 <= 2 {
				res.w[1] = P256.w[3]
				res.w[0] = P256.w[2]
				fstar.w[1] = P256.w[1]
				fstar.w[0] = P256.w[0]
				if (res.w[0]&0x0000000000000001 != 0) &&
					((fstar.w[1] < bid_ten2mk128[ind-1].w[1]) ||
						(fstar.w[1] == bid_ten2mk128[ind-1].w[1] &&
							fstar.w[0] < bid_ten2mk128[ind-1].w[0])) {
					res.w[0]--
				}
			} else if ind-1 <= 21 {
				shift = bid_shiftright128[ind-1]
				res.w[1] = P256.w[3] >> uint(shift)
				res.w[0] = (P256.w[3] << uint(64-shift)) | (P256.w[2] >> uint(shift))
				fstar.w[2] = P256.w[2] & bid_maskhigh128[ind-1]
				fstar.w[1] = P256.w[1]
				fstar.w[0] = P256.w[0]
				if (res.w[0]&0x0000000000000001 != 0) &&
					fstar.w[2] == 0 && (fstar.w[1] < bid_ten2mk128[ind-1].w[1] ||
					(fstar.w[1] == bid_ten2mk128[ind-1].w[1] && fstar.w[0] < bid_ten2mk128[ind-1].w[0])) {
					res.w[0]--
				}
			} else {
				shift = bid_shiftright128[ind-1] - 64
				res.w[1] = 0
				res.w[0] = P256.w[3] >> uint(shift)
				fstar.w[3] = P256.w[3] & bid_maskhigh128[ind-1]
				fstar.w[2] = P256.w[2]
				fstar.w[1] = P256.w[1]
				fstar.w[0] = P256.w[0]
				if (res.w[0]&0x0000000000000001 != 0) &&
					fstar.w[3] == 0 && fstar.w[2] == 0 && (fstar.w[1] < bid_ten2mk128[ind-1].w[1] ||
					(fstar.w[1] == bid_ten2mk128[ind-1].w[1] && fstar.w[0] < bid_ten2mk128[ind-1].w[0])) {
					res.w[0]--
				}
			}
			res.w[1] = x_sign | 0x3040000000000000 | res.w[1]
			return res, pfpsf
		} else {
			res.w[1] = x_sign | 0x3040000000000000
			res.w[0] = 0x0000000000000000
			return res, pfpsf
		}

	case BID_ROUNDING_TIES_AWAY:
		if q+exp >= 0 {
			ind = -exp
			tmp64 = C1.w[0]
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

			if ind-1 <= 2 {
				res.w[1] = P256.w[3]
				res.w[0] = P256.w[2]
			} else if ind-1 <= 21 {
				shift = bid_shiftright128[ind-1]
				res.w[1] = P256.w[3] >> uint(shift)
				res.w[0] = (P256.w[3] << uint(64-shift)) | (P256.w[2] >> uint(shift))
			} else {
				shift = bid_shiftright128[ind-1] - 64
				res.w[1] = 0
				res.w[0] = P256.w[3] >> uint(shift)
			}
			res.w[1] |= x_sign | 0x3040000000000000
			return res, pfpsf
		} else {
			res.w[1] = x_sign | 0x3040000000000000
			res.w[0] = 0x0000000000000000
			return res, pfpsf
		}

	case BID_ROUNDING_DOWN:
		if q+exp > 0 {
			ind = -exp
			P256 = __mul_128x128_to_256(C1, bid_ten2mk128[ind-1])
			if ind-1 <= 2 {
				res.w[1] = P256.w[3]
				res.w[0] = P256.w[2]
				if (P256.w[1] > bid_ten2mk128[ind-1].w[1]) ||
					(P256.w[1] == bid_ten2mk128[ind-1].w[1] && P256.w[0] >= bid_ten2mk128[ind-1].w[0]) {
					if x_sign != 0 {
						res.w[0]++
						if res.w[0] == 0 {
							res.w[1]++
						}
					}
				}
			} else if ind-1 <= 21 {
				shift = bid_shiftright128[ind-1]
				res.w[1] = P256.w[3] >> uint(shift)
				res.w[0] = (P256.w[3] << uint(64-shift)) | (P256.w[2] >> uint(shift))
				fstar.w[2] = P256.w[2] & bid_maskhigh128[ind-1]
				fstar.w[1] = P256.w[1]
				fstar.w[0] = P256.w[0]
				if fstar.w[2] != 0 || fstar.w[1] > bid_ten2mk128[ind-1].w[1] ||
					(fstar.w[1] == bid_ten2mk128[ind-1].w[1] && fstar.w[0] >= bid_ten2mk128[ind-1].w[0]) {
					if x_sign != 0 {
						res.w[0]++
						if res.w[0] == 0 {
							res.w[1]++
						}
					}
				}
			} else {
				shift = bid_shiftright128[ind-1] - 64
				res.w[1] = 0
				res.w[0] = P256.w[3] >> uint(shift)
				fstar.w[3] = P256.w[3] & bid_maskhigh128[ind-1]
				fstar.w[2] = P256.w[2]
				fstar.w[1] = P256.w[1]
				fstar.w[0] = P256.w[0]
				if fstar.w[3] != 0 || fstar.w[2] != 0 ||
					fstar.w[1] > bid_ten2mk128[ind-1].w[1] ||
					(fstar.w[1] == bid_ten2mk128[ind-1].w[1] && fstar.w[0] >= bid_ten2mk128[ind-1].w[0]) {
					if x_sign != 0 {
						res.w[0]++
						if res.w[0] == 0 {
							res.w[1]++
						}
					}
				}
			}
			res.w[1] = x_sign | 0x3040000000000000 | res.w[1]
			return res, pfpsf
		} else {
			if x_sign != 0 {
				res.w[1] = 0xb040000000000000
				res.w[0] = 0x0000000000000001
			} else {
				res.w[1] = 0x3040000000000000
				res.w[0] = 0x0000000000000000
			}
			return res, pfpsf
		}

	case BID_ROUNDING_UP:
		if q+exp > 0 {
			ind = -exp
			P256 = __mul_128x128_to_256(C1, bid_ten2mk128[ind-1])
			if ind-1 <= 2 {
				res.w[1] = P256.w[3]
				res.w[0] = P256.w[2]
				if (P256.w[1] > bid_ten2mk128[ind-1].w[1]) ||
					(P256.w[1] == bid_ten2mk128[ind-1].w[1] && P256.w[0] >= bid_ten2mk128[ind-1].w[0]) {
					if x_sign == 0 {
						res.w[0]++
						if res.w[0] == 0 {
							res.w[1]++
						}
					}
				}
			} else if ind-1 <= 21 {
				shift = bid_shiftright128[ind-1]
				res.w[1] = P256.w[3] >> uint(shift)
				res.w[0] = (P256.w[3] << uint(64-shift)) | (P256.w[2] >> uint(shift))
				fstar.w[2] = P256.w[2] & bid_maskhigh128[ind-1]
				fstar.w[1] = P256.w[1]
				fstar.w[0] = P256.w[0]
				if fstar.w[2] != 0 || fstar.w[1] > bid_ten2mk128[ind-1].w[1] ||
					(fstar.w[1] == bid_ten2mk128[ind-1].w[1] && fstar.w[0] >= bid_ten2mk128[ind-1].w[0]) {
					if x_sign == 0 {
						res.w[0]++
						if res.w[0] == 0 {
							res.w[1]++
						}
					}
				}
			} else {
				shift = bid_shiftright128[ind-1] - 64
				res.w[1] = 0
				res.w[0] = P256.w[3] >> uint(shift)
				fstar.w[3] = P256.w[3] & bid_maskhigh128[ind-1]
				fstar.w[2] = P256.w[2]
				fstar.w[1] = P256.w[1]
				fstar.w[0] = P256.w[0]
				if fstar.w[3] != 0 || fstar.w[2] != 0 ||
					fstar.w[1] > bid_ten2mk128[ind-1].w[1] ||
					(fstar.w[1] == bid_ten2mk128[ind-1].w[1] && fstar.w[0] >= bid_ten2mk128[ind-1].w[0]) {
					if x_sign == 0 {
						res.w[0]++
						if res.w[0] == 0 {
							res.w[1]++
						}
					}
				}
			}
			res.w[1] = x_sign | 0x3040000000000000 | res.w[1]
			return res, pfpsf
		} else {
			if x_sign != 0 {
				res.w[1] = 0xb040000000000000
				res.w[0] = 0x0000000000000000
			} else {
				res.w[1] = 0x3040000000000000
				res.w[0] = 0x0000000000000001
			}
			return res, pfpsf
		}

	case BID_ROUNDING_TO_ZERO:
		if q+exp > 0 {
			ind = -exp
			P256 = __mul_128x128_to_256(C1, bid_ten2mk128[ind-1])
			if ind-1 <= 2 {
				res.w[1] = P256.w[3]
				res.w[0] = P256.w[2]
			} else if ind-1 <= 21 {
				shift = bid_shiftright128[ind-1]
				res.w[1] = P256.w[3] >> uint(shift)
				res.w[0] = (P256.w[3] << uint(64-shift)) | (P256.w[2] >> uint(shift))
			} else {
				shift = bid_shiftright128[ind-1] - 64
				res.w[1] = 0
				res.w[0] = P256.w[3] >> uint(shift)
			}
			res.w[1] = x_sign | 0x3040000000000000 | res.w[1]
			return res, pfpsf
		} else {
			res.w[1] = x_sign | 0x3040000000000000
			res.w[0] = 0x0000000000000000
			return res, pfpsf
		}
	}

	return res, pfpsf
}
