// Ported from: Intel bid128_round_integral.c
// Mechanical translation - all logic preserved exactly.

package bidgo

import "math"

// Bid128RoundIntegralExact is ported mechanically from bid128_round_integral.c: bid128_round_integral_exact.
func Bid128RoundIntegralExact(x BID_UINT128, rnd_mode int, pfpsf *uint32) BID_UINT128 {
	var res BID_UINT128
	res.w[0] = 0xbaddbaddbaddbadd
	res.w[1] = 0xbaddbaddbaddbadd
	var x_sign uint64
	var x_exp uint64
	var exp int // unbiased exponent
	// Note: C1.w[1], C1.w[0] represent x_signif_hi, x_signif_lo (all are uint64)
	var tmp64 uint64
	var x_nr_bits uint32
	var q, ind, shift int
	var C1 BID_UINT128
	var fstar BID_UINT256
	var P256 BID_UINT256

	// check for NaN or Infinity
	if (x.w[1] & MASK_SPECIAL128) == MASK_SPECIAL128 {
		// x is special
		if (x.w[1] & NAN_MASK64) == NAN_MASK64 { // x is NAN
			// if x = NaN, then res = Q (x)
			// check first for non-canonical NaN payload
			if ((x.w[1] & 0x00003fffffffffff) > 0x0000314dc6448d93) ||
				(((x.w[1] & 0x00003fffffffffff) == 0x0000314dc6448d93) &&
					(x.w[0] > 0x38c15b09ffffffff)) {
				x.w[1] = x.w[1] & 0xffffc00000000000
				x.w[0] = 0x0
			}
			if (x.w[1] & SNAN_MASK64) == SNAN_MASK64 { // x is SNAN
				// set invalid flag
				*pfpsf |= BID_INVALID_EXCEPTION
				// return quiet (x)
				res.w[1] = x.w[1] & 0xfc003fffffffffff // clear out also G[6]-G[16]
				res.w[0] = x.w[0]
			} else { // x is QNaN
				// return x
				res.w[1] = x.w[1] & 0xfc003fffffffffff // clear out G[6]-G[16]
				res.w[0] = x.w[0]
			}
			return res
		} else { // x is not a NaN, so it must be infinity
			if (x.w[1] & MASK_SIGN_128) == 0x0 { // x is +inf
				// return +inf
				res.w[1] = 0x7800000000000000
				res.w[0] = 0x0000000000000000
			} else { // x is -inf
				// return -inf
				res.w[1] = 0xf800000000000000
				res.w[0] = 0x0000000000000000
			}
			return res
		}
	}
	// unpack x
	x_sign = x.w[1] & MASK_SIGN_128 // 0 for positive, MASK_SIGN for negative
	C1.w[1] = x.w[1] & MASK_COEFF_128
	C1.w[0] = x.w[0]

	// check for non-canonical values (treated as zero)
	if (x.w[1] & 0x6000000000000000) == 0x6000000000000000 { // G0_G1=11
		// non-canonical
		x_exp = (x.w[1] << 2) & MASK_EXP_128 // biased and shifted left 49 bits
		C1.w[1] = 0                          // significand high
		C1.w[0] = 0                          // significand low
	} else { // G0_G1 != 11
		x_exp = x.w[1] & MASK_EXP_128 // biased and shifted left 49 bits
		if C1.w[1] > 0x0001ed09bead87c0 ||
			(C1.w[1] == 0x0001ed09bead87c0 &&
				C1.w[0] > 0x378d8e63ffffffff) {
			// x is non-canonical if coefficient is larger than 10^34 -1
			C1.w[1] = 0
			C1.w[0] = 0
		} else { // canonical
			// ;
		}
	}

	// test for input equal to zero
	if (C1.w[1] == 0x0) && (C1.w[0] == 0x0) {
		// x is 0
		// return 0 preserving the sign bit and the preferred exponent
		// of MAX(Q(x), 0)
		if x_exp <= (0x1820 << 49) {
			res.w[1] = (x.w[1] & 0x8000000000000000) | 0x3040000000000000
		} else {
			res.w[1] = x_sign | x_exp
		}
		res.w[0] = 0x0000000000000000
		return res
	}
	// x is not special and is not zero

	switch rnd_mode {
	case BID_ROUNDING_TO_NEAREST, BID_ROUNDING_TIES_AWAY:
		// if (exp <= -(p+1)) return 0.0
		if x_exp <= 0x2ffa000000000000 { // 0x2ffa000000000000 == -35
			res.w[1] = x_sign | 0x3040000000000000
			res.w[0] = 0x0000000000000000
			*pfpsf |= BID_INEXACT_EXCEPTION
			return res
		}
	case BID_ROUNDING_DOWN:
		// if (exp <= -p) return -1.0 or +0.0
		if x_exp <= 0x2ffc000000000000 { // 0x2ffc000000000000 == -34
			if x_sign != 0 {
				// if negative, return negative 1, because we know coefficient
				// is non-zero (would have been caught above)
				res.w[1] = 0xb040000000000000
				res.w[0] = 0x0000000000000001
			} else {
				// if positive, return positive 0, because we know coefficient is
				// non-zero (would have been caught above)
				res.w[1] = 0x3040000000000000
				res.w[0] = 0x0000000000000000
			}
			*pfpsf |= BID_INEXACT_EXCEPTION
			return res
		}
	case BID_ROUNDING_UP:
		// if (exp <= -p) return -0.0 or +1.0
		if x_exp <= 0x2ffc000000000000 { // 0x2ffc000000000000 == -34
			if x_sign != 0 {
				// if negative, return negative 0, because we know the coefficient
				// is non-zero (would have been caught above)
				res.w[1] = 0xb040000000000000
				res.w[0] = 0x0000000000000000
			} else {
				// if positive, return positive 1, because we know coefficient is
				// non-zero (would have been caught above)
				res.w[1] = 0x3040000000000000
				res.w[0] = 0x0000000000000001
			}
			*pfpsf |= BID_INEXACT_EXCEPTION
			return res
		}
	case BID_ROUNDING_TO_ZERO:
		// if (exp <= -p) return -0.0 or +0.0
		if x_exp <= 0x2ffc000000000000 { // 0x2ffc000000000000 == -34
			res.w[1] = x_sign | 0x3040000000000000
			res.w[0] = 0x0000000000000000
			*pfpsf |= BID_INEXACT_EXCEPTION
			return res
		}
	default:
		// default added to avoid compiler warning
	}

	// q = nr. of decimal digits in x
	//  determine first the nr. of bits in x
	if C1.w[1] == 0 {
		if C1.w[0] >= 0x0020000000000000 { // x >= 2^53
			// split the 64-bit value in two 32-bit halves to avoid rounding errors
			tmp1 := math.Float64bits(float64(C1.w[0] >> 32)) // exact conversion
			x_nr_bits = 33 + uint32(((tmp1>>52)&0x7ff)-0x3ff)
		} else { // if x < 2^53
			tmp1 := math.Float64bits(float64(C1.w[0])) // exact conversion
			x_nr_bits = 1 + uint32(((tmp1>>52)&0x7ff)-0x3ff)
		}
	} else { // C1.w[1] != 0 => nr. bits = 64 + nr_bits (C1.w[1])
		tmp1 := math.Float64bits(float64(C1.w[1])) // exact conversion
		x_nr_bits = 65 + uint32(((tmp1>>52)&0x7ff)-0x3ff)
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
	if exp >= 0 { // -exp <= 0
		// the argument is an integer already
		res.w[1] = x.w[1]
		res.w[0] = x.w[0]
		return res
	}
	// exp < 0
	switch rnd_mode {
	case BID_ROUNDING_TO_NEAREST:
		if (q + exp) >= 0 { // exp < 0 and 1 <= -exp <= q
			// need to shift right -exp digits from the coefficient; exp will be 0
			ind = -exp // 1 <= ind <= 34; ind is a synonym for 'x'
			// chop off ind digits from the lower part of C1
			// C1 = C1 + 1/2 * 10^x where the result C1 fits in 127 bits
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
			// calculate C* and f*
			// C* is actually floor(C*) in this case
			// C* and f* need shifting and masking, as shown by
			// bid_shiftright128[] and bid_maskhigh128[]
			// 1 <= x <= 34
			// kx = 10^(-x) = bid_ten2mk128[ind - 1]
			// C* = (C1 + 1/2 * 10^x) * 10^(-x)
			// the approximation of 10^(-x) was rounded up to 118 bits
			P256 = __mul_128x128_to_256(C1, bid_ten2mk128[ind-1])
			// determine the value of res and fstar

			// determine inexactness of the rounding of C*
			// if (0 < f* - 1/2 < 10^(-x)) then
			//   the result is exact
			// else // if (f* - 1/2 > T*) then
			//   the result is inexact

			if ind-1 <= 2 { // 0 <= ind - 1 <= 2 => shift = 0
				res.w[1] = P256.w[3]
				res.w[0] = P256.w[2]
				fstar.w[1] = P256.w[1]
				fstar.w[0] = P256.w[0]
				// if 0 < fstar < 10^(-x), subtract 1 if odd (for rounding to even)
				if (res.w[0]&0x0000000000000001) != 0 && // is result odd, and from MP?
					((fstar.w[1] < (bid_ten2mk128[ind-1].w[1])) ||
						((fstar.w[1] == bid_ten2mk128[ind-1].w[1]) &&
							(fstar.w[0] < bid_ten2mk128[ind-1].w[0]))) {
					// subtract 1 to make even
					res.w[0]--
				}
				if fstar.w[1] > 0x8000000000000000 ||
					(fstar.w[1] == 0x8000000000000000 &&
						fstar.w[0] > 0x0) {
					// f* > 1/2 and the result may be exact
					tmp64 = fstar.w[1] - 0x8000000000000000 // f* - 1/2
					if tmp64 > bid_ten2mk128[ind-1].w[1] ||
						(tmp64 == bid_ten2mk128[ind-1].w[1] &&
							fstar.w[0] >= bid_ten2mk128[ind-1].w[0]) {
						// set the inexact flag
						*pfpsf |= BID_INEXACT_EXCEPTION
					} // else the result is exact
				} else { // the result is inexact; f2* <= 1/2
					// set the inexact flag
					*pfpsf |= BID_INEXACT_EXCEPTION
				}
			} else if ind-1 <= 21 { // 3 <= ind - 1 <= 21 => 3 <= shift <= 63
				shift = bid_shiftright128[ind-1] // 3 <= shift <= 63
				res.w[1] = (P256.w[3] >> uint(shift))
				res.w[0] = (P256.w[3] << uint(64-shift)) | (P256.w[2] >> uint(shift))
				fstar.w[2] = P256.w[2] & bid_maskhigh128[ind-1]
				fstar.w[1] = P256.w[1]
				fstar.w[0] = P256.w[0]
				if (res.w[0]&0x0000000000000001) != 0 && // is result odd, and from MP?
					fstar.w[2] == 0 && (fstar.w[1] < bid_ten2mk128[ind-1].w[1] ||
					(fstar.w[1] == bid_ten2mk128[ind-1].w[1] &&
						fstar.w[0] < bid_ten2mk128[ind-1].w[0])) {
					// subtract 1 to make even
					res.w[0]--
				}
				if fstar.w[2] > bid_onehalf128[ind-1] ||
					(fstar.w[2] == bid_onehalf128[ind-1] &&
						(fstar.w[1] != 0 || fstar.w[0] != 0)) {
					// f2* > 1/2 and the result may be exact
					// Calculate f2* - 1/2
					tmp64 = fstar.w[2] - bid_onehalf128[ind-1]
					if tmp64 != 0 || fstar.w[1] > bid_ten2mk128[ind-1].w[1] ||
						(fstar.w[1] == bid_ten2mk128[ind-1].w[1] &&
							fstar.w[0] >= bid_ten2mk128[ind-1].w[0]) {
						// set the inexact flag
						*pfpsf |= BID_INEXACT_EXCEPTION
					} // else the result is exact
				} else { // the result is inexact; f2* <= 1/2
					// set the inexact flag
					*pfpsf |= BID_INEXACT_EXCEPTION
				}
			} else { // 22 <= ind - 1 <= 33
				shift = bid_shiftright128[ind-1] - 64 // 2 <= shift <= 38
				res.w[1] = 0
				res.w[0] = P256.w[3] >> uint(shift)
				fstar.w[3] = P256.w[3] & bid_maskhigh128[ind-1]
				fstar.w[2] = P256.w[2]
				fstar.w[1] = P256.w[1]
				fstar.w[0] = P256.w[0]
				if (res.w[0]&0x0000000000000001) != 0 && // is result odd, and from MP?
					fstar.w[3] == 0 && fstar.w[2] == 0 &&
					(fstar.w[1] < bid_ten2mk128[ind-1].w[1] ||
						(fstar.w[1] == bid_ten2mk128[ind-1].w[1] &&
							fstar.w[0] < bid_ten2mk128[ind-1].w[0])) {
					// subtract 1 to make even
					res.w[0]--
				}
				if fstar.w[3] > bid_onehalf128[ind-1] ||
					(fstar.w[3] == bid_onehalf128[ind-1] &&
						(fstar.w[2] != 0 || fstar.w[1] != 0 || fstar.w[0] != 0)) {
					// f2* > 1/2 and the result may be exact
					// Calculate f2* - 1/2
					tmp64 = fstar.w[3] - bid_onehalf128[ind-1]
					if tmp64 != 0 || fstar.w[2] != 0 || fstar.w[1] > bid_ten2mk128[ind-1].w[1] ||
						(fstar.w[1] == bid_ten2mk128[ind-1].w[1] &&
							fstar.w[0] >= bid_ten2mk128[ind-1].w[0]) {
						// set the inexact flag
						*pfpsf |= BID_INEXACT_EXCEPTION
					} // else the result is exact
				} else { // the result is inexact; f2* <= 1/2
					// set the inexact flag
					*pfpsf |= BID_INEXACT_EXCEPTION
				}
			}
			res.w[1] = x_sign | 0x3040000000000000 | res.w[1]
			return res
		} else { // if ((q + exp) < 0) <=> q < -exp
			// the result is +0 or -0
			res.w[1] = x_sign | 0x3040000000000000
			res.w[0] = 0x0000000000000000
			*pfpsf |= BID_INEXACT_EXCEPTION
			return res
		}
	case BID_ROUNDING_TIES_AWAY:
		if (q + exp) >= 0 { // exp < 0 and 1 <= -exp <= q
			// need to shift right -exp digits from the coefficient; exp will be 0
			ind = -exp // 1 <= ind <= 34; ind is a synonym for 'x'
			// chop off ind digits from the lower part of C1
			// C1 = C1 + 1/2 * 10^x where the result C1 fits in 127 bits
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
			// calculate C* and f*
			P256 = __mul_128x128_to_256(C1, bid_ten2mk128[ind-1])

			if ind-1 <= 2 { // 0 <= ind - 1 <= 2 => shift = 0
				res.w[1] = P256.w[3]
				res.w[0] = P256.w[2]
				fstar.w[1] = P256.w[1]
				fstar.w[0] = P256.w[0]
				if fstar.w[1] > 0x8000000000000000 ||
					(fstar.w[1] == 0x8000000000000000 &&
						fstar.w[0] > 0x0) {
					// f* > 1/2 and the result may be exact
					tmp64 = fstar.w[1] - 0x8000000000000000 // f* - 1/2
					if tmp64 > bid_ten2mk128[ind-1].w[1] ||
						(tmp64 == bid_ten2mk128[ind-1].w[1] &&
							fstar.w[0] >= bid_ten2mk128[ind-1].w[0]) {
						// set the inexact flag
						*pfpsf |= BID_INEXACT_EXCEPTION
					} // else the result is exact
				} else { // the result is inexact; f2* <= 1/2
					// set the inexact flag
					*pfpsf |= BID_INEXACT_EXCEPTION
				}
			} else if ind-1 <= 21 { // 3 <= ind - 1 <= 21 => 3 <= shift <= 63
				shift = bid_shiftright128[ind-1] // 3 <= shift <= 63
				res.w[1] = (P256.w[3] >> uint(shift))
				res.w[0] = (P256.w[3] << uint(64-shift)) | (P256.w[2] >> uint(shift))
				fstar.w[2] = P256.w[2] & bid_maskhigh128[ind-1]
				fstar.w[1] = P256.w[1]
				fstar.w[0] = P256.w[0]
				if fstar.w[2] > bid_onehalf128[ind-1] ||
					(fstar.w[2] == bid_onehalf128[ind-1] &&
						(fstar.w[1] != 0 || fstar.w[0] != 0)) {
					// f2* > 1/2 and the result may be exact
					// Calculate f2* - 1/2
					tmp64 = fstar.w[2] - bid_onehalf128[ind-1]
					if tmp64 != 0 || fstar.w[1] > bid_ten2mk128[ind-1].w[1] ||
						(fstar.w[1] == bid_ten2mk128[ind-1].w[1] &&
							fstar.w[0] >= bid_ten2mk128[ind-1].w[0]) {
						// set the inexact flag
						*pfpsf |= BID_INEXACT_EXCEPTION
					} // else the result is exact
				} else { // the result is inexact; f2* <= 1/2
					// set the inexact flag
					*pfpsf |= BID_INEXACT_EXCEPTION
				}
			} else { // 22 <= ind - 1 <= 33
				shift = bid_shiftright128[ind-1] - 64 // 2 <= shift <= 38
				res.w[1] = 0
				res.w[0] = P256.w[3] >> uint(shift)
				fstar.w[3] = P256.w[3] & bid_maskhigh128[ind-1]
				fstar.w[2] = P256.w[2]
				fstar.w[1] = P256.w[1]
				fstar.w[0] = P256.w[0]
				if fstar.w[3] > bid_onehalf128[ind-1] ||
					(fstar.w[3] == bid_onehalf128[ind-1] &&
						(fstar.w[2] != 0 || fstar.w[1] != 0 || fstar.w[0] != 0)) {
					// f2* > 1/2 and the result may be exact
					// Calculate f2* - 1/2
					tmp64 = fstar.w[3] - bid_onehalf128[ind-1]
					if tmp64 != 0 || fstar.w[2] != 0 || fstar.w[1] > bid_ten2mk128[ind-1].w[1] ||
						(fstar.w[1] == bid_ten2mk128[ind-1].w[1] &&
							fstar.w[0] >= bid_ten2mk128[ind-1].w[0]) {
						// set the inexact flag
						*pfpsf |= BID_INEXACT_EXCEPTION
					} // else the result is exact
				} else { // the result is inexact; f2* <= 1/2
					// set the inexact flag
					*pfpsf |= BID_INEXACT_EXCEPTION
				}
			}
			// if the result was a midpoint, it was already rounded away from zero
			res.w[1] |= x_sign | 0x3040000000000000
			return res
		} else { // if ((q + exp) < 0) <=> q < -exp
			// the result is +0 or -0
			res.w[1] = x_sign | 0x3040000000000000
			res.w[0] = 0x0000000000000000
			*pfpsf |= BID_INEXACT_EXCEPTION
			return res
		}
	case BID_ROUNDING_DOWN:
		if (q + exp) > 0 { // exp < 0 and 1 <= -exp < q
			// need to shift right -exp digits from the coefficient; exp will be 0
			ind = -exp // 1 <= ind <= 34; ind is a synonym for 'x'
			P256 = __mul_128x128_to_256(C1, bid_ten2mk128[ind-1])
			if ind-1 <= 2 { // 0 <= ind - 1 <= 2 => shift = 0
				res.w[1] = P256.w[3]
				res.w[0] = P256.w[2]
				if (P256.w[1] > bid_ten2mk128[ind-1].w[1]) ||
					(P256.w[1] == bid_ten2mk128[ind-1].w[1] &&
						(P256.w[0] >= bid_ten2mk128[ind-1].w[0])) {
					*pfpsf |= BID_INEXACT_EXCEPTION
					// if positive, the truncated value is already the correct result
					if x_sign != 0 { // if negative
						res.w[0]++
						if res.w[0] == 0 {
							res.w[1]++
						}
					}
				}
			} else if ind-1 <= 21 { // 3 <= ind - 1 <= 21 => 3 <= shift <= 63
				shift = bid_shiftright128[ind-1] // 0 <= shift <= 102
				res.w[1] = (P256.w[3] >> uint(shift))
				res.w[0] = (P256.w[3] << uint(64-shift)) | (P256.w[2] >> uint(shift))
				fstar.w[2] = P256.w[2] & bid_maskhigh128[ind-1]
				fstar.w[1] = P256.w[1]
				fstar.w[0] = P256.w[0]
				if fstar.w[2] != 0 || fstar.w[1] > bid_ten2mk128[ind-1].w[1] ||
					(fstar.w[1] == bid_ten2mk128[ind-1].w[1] &&
						fstar.w[0] >= bid_ten2mk128[ind-1].w[0]) {
					*pfpsf |= BID_INEXACT_EXCEPTION
					// if positive, the truncated value is already the correct result
					if x_sign != 0 { // if negative
						res.w[0]++
						if res.w[0] == 0 {
							res.w[1]++
						}
					}
				}
			} else { // 22 <= ind - 1 <= 33
				shift = bid_shiftright128[ind-1] - 64 // 2 <= shift <= 38
				res.w[1] = 0
				res.w[0] = P256.w[3] >> uint(shift)
				fstar.w[3] = P256.w[3] & bid_maskhigh128[ind-1]
				fstar.w[2] = P256.w[2]
				fstar.w[1] = P256.w[1]
				fstar.w[0] = P256.w[0]
				if fstar.w[3] != 0 || fstar.w[2] != 0 ||
					fstar.w[1] > bid_ten2mk128[ind-1].w[1] ||
					(fstar.w[1] == bid_ten2mk128[ind-1].w[1] &&
						fstar.w[0] >= bid_ten2mk128[ind-1].w[0]) {
					*pfpsf |= BID_INEXACT_EXCEPTION
					// if positive, the truncated value is already the correct result
					if x_sign != 0 { // if negative
						res.w[0]++
						if res.w[0] == 0 {
							res.w[1]++
						}
					}
				}
			}
			res.w[1] = x_sign | 0x3040000000000000 | res.w[1]
			return res
		} else { // if exp < 0 and q + exp <= 0
			if x_sign != 0 { // negative rounds down to -1.0
				res.w[1] = 0xb040000000000000
				res.w[0] = 0x0000000000000001
			} else { // positive rounds down to +0.0
				res.w[1] = 0x3040000000000000
				res.w[0] = 0x0000000000000000
			}
			*pfpsf |= BID_INEXACT_EXCEPTION
			return res
		}
	case BID_ROUNDING_UP:
		if (q + exp) > 0 { // exp < 0 and 1 <= -exp < q
			// need to shift right -exp digits from the coefficient; exp will be 0
			ind = -exp // 1 <= ind <= 34; ind is a synonym for 'x'
			P256 = __mul_128x128_to_256(C1, bid_ten2mk128[ind-1])
			if ind-1 <= 2 { // 0 <= ind - 1 <= 2 => shift = 0
				res.w[1] = P256.w[3]
				res.w[0] = P256.w[2]
				if (P256.w[1] > bid_ten2mk128[ind-1].w[1]) ||
					(P256.w[1] == bid_ten2mk128[ind-1].w[1] &&
						(P256.w[0] >= bid_ten2mk128[ind-1].w[0])) {
					*pfpsf |= BID_INEXACT_EXCEPTION
					// if negative, the truncated value is already the correct result
					if x_sign == 0 { // if positive
						res.w[0]++
						if res.w[0] == 0 {
							res.w[1]++
						}
					}
				}
			} else if ind-1 <= 21 { // 3 <= ind - 1 <= 21 => 3 <= shift <= 63
				shift = bid_shiftright128[ind-1] // 3 <= shift <= 63
				res.w[1] = (P256.w[3] >> uint(shift))
				res.w[0] = (P256.w[3] << uint(64-shift)) | (P256.w[2] >> uint(shift))
				fstar.w[2] = P256.w[2] & bid_maskhigh128[ind-1]
				fstar.w[1] = P256.w[1]
				fstar.w[0] = P256.w[0]
				if fstar.w[2] != 0 || fstar.w[1] > bid_ten2mk128[ind-1].w[1] ||
					(fstar.w[1] == bid_ten2mk128[ind-1].w[1] &&
						fstar.w[0] >= bid_ten2mk128[ind-1].w[0]) {
					*pfpsf |= BID_INEXACT_EXCEPTION
					// if negative, the truncated value is already the correct result
					if x_sign == 0 { // if positive
						res.w[0]++
						if res.w[0] == 0 {
							res.w[1]++
						}
					}
				}
			} else { // 22 <= ind - 1 <= 33
				shift = bid_shiftright128[ind-1] - 64 // 2 <= shift <= 38
				res.w[1] = 0
				res.w[0] = P256.w[3] >> uint(shift)
				fstar.w[3] = P256.w[3] & bid_maskhigh128[ind-1]
				fstar.w[2] = P256.w[2]
				fstar.w[1] = P256.w[1]
				fstar.w[0] = P256.w[0]
				if fstar.w[3] != 0 || fstar.w[2] != 0 ||
					fstar.w[1] > bid_ten2mk128[ind-1].w[1] ||
					(fstar.w[1] == bid_ten2mk128[ind-1].w[1] &&
						fstar.w[0] >= bid_ten2mk128[ind-1].w[0]) {
					*pfpsf |= BID_INEXACT_EXCEPTION
					// if negative, the truncated value is already the correct result
					if x_sign == 0 { // if positive
						res.w[0]++
						if res.w[0] == 0 {
							res.w[1]++
						}
					}
				}
			}
			res.w[1] = x_sign | 0x3040000000000000 | res.w[1]
			return res
		} else { // if exp < 0 and q + exp <= 0
			if x_sign != 0 { // negative rounds up to -0.0
				res.w[1] = 0xb040000000000000
				res.w[0] = 0x0000000000000000
			} else { // positive rounds up to +1.0
				res.w[1] = 0x3040000000000000
				res.w[0] = 0x0000000000000001
			}
			*pfpsf |= BID_INEXACT_EXCEPTION
			return res
		}
	case BID_ROUNDING_TO_ZERO:
		if (q + exp) > 0 { // exp < 0 and 1 <= -exp < q
			// need to shift right -exp digits from the coefficient; exp will be 0
			ind = -exp // 1 <= ind <= 34; ind is a synonym for 'x'
			P256 = __mul_128x128_to_256(C1, bid_ten2mk128[ind-1])
			if ind-1 <= 2 { // 0 <= ind - 1 <= 2 => shift = 0
				res.w[1] = P256.w[3]
				res.w[0] = P256.w[2]
				if (P256.w[1] > bid_ten2mk128[ind-1].w[1]) ||
					(P256.w[1] == bid_ten2mk128[ind-1].w[1] &&
						(P256.w[0] >= bid_ten2mk128[ind-1].w[0])) {
					*pfpsf |= BID_INEXACT_EXCEPTION
				}
			} else if ind-1 <= 21 { // 3 <= ind - 1 <= 21 => 3 <= shift <= 63
				shift = bid_shiftright128[ind-1] // 3 <= shift <= 63
				res.w[1] = (P256.w[3] >> uint(shift))
				res.w[0] = (P256.w[3] << uint(64-shift)) | (P256.w[2] >> uint(shift))
				fstar.w[2] = P256.w[2] & bid_maskhigh128[ind-1]
				fstar.w[1] = P256.w[1]
				fstar.w[0] = P256.w[0]
				if fstar.w[2] != 0 || fstar.w[1] > bid_ten2mk128[ind-1].w[1] ||
					(fstar.w[1] == bid_ten2mk128[ind-1].w[1] &&
						fstar.w[0] >= bid_ten2mk128[ind-1].w[0]) {
					*pfpsf |= BID_INEXACT_EXCEPTION
				}
			} else { // 22 <= ind - 1 <= 33
				shift = bid_shiftright128[ind-1] - 64 // 2 <= shift <= 38
				res.w[1] = 0
				res.w[0] = P256.w[3] >> uint(shift)
				fstar.w[3] = P256.w[3] & bid_maskhigh128[ind-1]
				fstar.w[2] = P256.w[2]
				fstar.w[1] = P256.w[1]
				fstar.w[0] = P256.w[0]
				if fstar.w[3] != 0 || fstar.w[2] != 0 ||
					fstar.w[1] > bid_ten2mk128[ind-1].w[1] ||
					(fstar.w[1] == bid_ten2mk128[ind-1].w[1] &&
						fstar.w[0] >= bid_ten2mk128[ind-1].w[0]) {
					*pfpsf |= BID_INEXACT_EXCEPTION
				}
			}
			res.w[1] = x_sign | 0x3040000000000000 | res.w[1]
			return res
		} else { // if exp < 0 and q + exp <= 0 the result is +0 or -0
			res.w[1] = x_sign | 0x3040000000000000
			res.w[0] = 0x0000000000000000
			*pfpsf |= BID_INEXACT_EXCEPTION
			return res
		}
	default:
		// default added to avoid compiler warning
	}

	return res
}

// bid128_round_integral_nornd_common contains the common NaN/Inf/zero/unpack
// logic shared by all non-rounding-mode variants (nearest_even, nearest_away,
// negative, positive, zero). It returns (res, done, x_sign, x_exp, C1, q, exp).
// If done is true, the caller should return res immediately.
func bid128_round_integral_nornd_common(x BID_UINT128, pfpsf *uint32) (
	res BID_UINT128, done bool, x_sign, x_exp uint64, C1 BID_UINT128, q, exp int) {

	// check for NaN or Infinity
	if (x.w[1] & MASK_SPECIAL128) == MASK_SPECIAL128 {
		// x is special
		if (x.w[1] & NAN_MASK64) == NAN_MASK64 { // x is NAN
			// if x = NaN, then res = Q (x)
			// check first for non-canonical NaN payload
			if ((x.w[1] & 0x00003fffffffffff) > 0x0000314dc6448d93) ||
				(((x.w[1] & 0x00003fffffffffff) == 0x0000314dc6448d93) &&
					(x.w[0] > 0x38c15b09ffffffff)) {
				x.w[1] = x.w[1] & 0xffffc00000000000
				x.w[0] = 0x0
			}
			if (x.w[1] & SNAN_MASK64) == SNAN_MASK64 { // x is SNAN
				// set invalid flag
				*pfpsf |= BID_INVALID_EXCEPTION
				// return quiet (x)
				res.w[1] = x.w[1] & 0xfc003fffffffffff // clear out also G[6]-G[16]
				res.w[0] = x.w[0]
			} else { // x is QNaN
				// return x
				res.w[1] = x.w[1] & 0xfc003fffffffffff // clear out G[6]-G[16]
				res.w[0] = x.w[0]
			}
			return res, true, 0, 0, C1, 0, 0
		} else { // x is not a NaN, so it must be infinity
			if (x.w[1] & MASK_SIGN_128) == 0x0 { // x is +inf
				// return +inf
				res.w[1] = 0x7800000000000000
				res.w[0] = 0x0000000000000000
			} else { // x is -inf
				// return -inf
				res.w[1] = 0xf800000000000000
				res.w[0] = 0x0000000000000000
			}
			return res, true, 0, 0, C1, 0, 0
		}
	}
	// unpack x
	x_sign = x.w[1] & MASK_SIGN_128 // 0 for positive, MASK_SIGN for negative
	C1.w[1] = x.w[1] & MASK_COEFF_128
	C1.w[0] = x.w[0]

	// check for non-canonical values (treated as zero)
	if (x.w[1] & 0x6000000000000000) == 0x6000000000000000 { // G0_G1=11
		// non-canonical
		x_exp = (x.w[1] << 2) & MASK_EXP_128 // biased and shifted left 49 bits
		C1.w[1] = 0                          // significand high
		C1.w[0] = 0                          // significand low
	} else { // G0_G1 != 11
		x_exp = x.w[1] & MASK_EXP_128 // biased and shifted left 49 bits
		if C1.w[1] > 0x0001ed09bead87c0 ||
			(C1.w[1] == 0x0001ed09bead87c0 &&
				C1.w[0] > 0x378d8e63ffffffff) {
			// x is non-canonical if coefficient is larger than 10^34 -1
			C1.w[1] = 0
			C1.w[0] = 0
		} else { // canonical
			// ;
		}
	}

	// test for input equal to zero
	if (C1.w[1] == 0x0) && (C1.w[0] == 0x0) {
		// x is 0
		// return 0 preserving the sign bit and the preferred exponent
		// of MAX(Q(x), 0)
		if x_exp <= (0x1820 << 49) {
			res.w[1] = (x.w[1] & 0x8000000000000000) | 0x3040000000000000
		} else {
			res.w[1] = x_sign | x_exp
		}
		res.w[0] = 0x0000000000000000
		return res, true, x_sign, x_exp, C1, 0, 0
	}

	// q = nr. of decimal digits in x
	// determine first the nr. of bits in x
	var x_nr_bits uint32
	if C1.w[1] == 0 {
		if C1.w[0] >= 0x0020000000000000 { // x >= 2^53
			tmp1 := math.Float64bits(float64(C1.w[0] >> 32)) // exact conversion
			x_nr_bits = 33 + uint32(((tmp1>>52)&0x7ff)-0x3ff)
		} else { // if x < 2^53
			tmp1 := math.Float64bits(float64(C1.w[0])) // exact conversion
			x_nr_bits = 1 + uint32(((tmp1>>52)&0x7ff)-0x3ff)
		}
	} else { // C1.w[1] != 0 => nr. bits = 64 + nr_bits (C1.w[1])
		tmp1 := math.Float64bits(float64(C1.w[1])) // exact conversion
		x_nr_bits = 65 + uint32(((tmp1>>52)&0x7ff)-0x3ff)
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

	return res, false, x_sign, x_exp, C1, q, exp
}

// Bid128RoundIntegralNearestEven is ported mechanically from bid128_round_integral.c.
func Bid128RoundIntegralNearestEven(x BID_UINT128, pfpsf *uint32) BID_UINT128 {
	var res BID_UINT128
	var fstar BID_UINT256
	var P256 BID_UINT256

	res, done, x_sign, _, C1, q, exp := bid128_round_integral_nornd_common(x, pfpsf)
	if done {
		return res
	}

	// x is not special and is not zero

	// if (exp <= -(p+1)) return 0
	if x.w[1]&MASK_EXP_128 <= 0x2ffa000000000000 { // 0x2ffa000000000000 == -35
		res.w[1] = x_sign | 0x3040000000000000
		res.w[0] = 0x0000000000000000
		return res
	}

	if exp >= 0 { // -exp <= 0
		// the argument is an integer already
		res.w[1] = x.w[1]
		res.w[0] = x.w[0]
		return res
	} else if (q + exp) >= 0 { // exp < 0 and 1 <= -exp <= q
		// need to shift right -exp digits from the coefficient; the exp will be 0
		ind := -exp // 1 <= ind <= 34; ind is a synonym for 'x'
		// chop off ind digits from the lower part of C1
		// C1 = C1 + 1/2 * 10^x where the result C1 fits in 127 bits
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
		// determine the value of res and fstar
		if ind-1 <= 2 { // 0 <= ind - 1 <= 2 => shift = 0
			res.w[1] = P256.w[3]
			res.w[0] = P256.w[2]
			// if 0 < fstar < 10^(-x), subtract 1 if odd (for rounding to even)
			if (res.w[0]&0x0000000000000001) != 0 &&
				((P256.w[1] < (bid_ten2mk128[ind-1].w[1])) ||
					((P256.w[1] == bid_ten2mk128[ind-1].w[1]) &&
						(P256.w[0] < bid_ten2mk128[ind-1].w[0]))) {
				// subtract 1 to make even
				res.w[0]--
			}
		} else if ind-1 <= 21 { // 3 <= ind - 1 <= 21 => 3 <= shift <= 63
			shift := bid_shiftright128[ind-1] // 3 <= shift <= 63
			res.w[1] = (P256.w[3] >> uint(shift))
			res.w[0] = (P256.w[3] << uint(64-shift)) | (P256.w[2] >> uint(shift))
			fstar.w[2] = P256.w[2] & bid_maskhigh128[ind-1]
			fstar.w[1] = P256.w[1]
			fstar.w[0] = P256.w[0]
			if (res.w[0]&0x0000000000000001) != 0 &&
				fstar.w[2] == 0 && (fstar.w[1] < bid_ten2mk128[ind-1].w[1] ||
				(fstar.w[1] == bid_ten2mk128[ind-1].w[1] &&
					fstar.w[0] < bid_ten2mk128[ind-1].w[0])) {
				// subtract 1 to make even
				res.w[0]--
			}
		} else { // 22 <= ind - 1 <= 33
			shift := bid_shiftright128[ind-1] - 64 // 2 <= shift <= 38
			res.w[1] = 0
			res.w[0] = P256.w[3] >> uint(shift)
			fstar.w[3] = P256.w[3] & bid_maskhigh128[ind-1]
			fstar.w[2] = P256.w[2]
			fstar.w[1] = P256.w[1]
			fstar.w[0] = P256.w[0]
			if (res.w[0]&0x0000000000000001) != 0 &&
				fstar.w[3] == 0 && fstar.w[2] == 0 &&
				(fstar.w[1] < bid_ten2mk128[ind-1].w[1] ||
					(fstar.w[1] == bid_ten2mk128[ind-1].w[1] &&
						fstar.w[0] < bid_ten2mk128[ind-1].w[0])) {
				// subtract 1 to make even
				res.w[0]--
			}
		}
		res.w[1] = x_sign | 0x3040000000000000 | res.w[1]
		return res
	} else { // if ((q + exp) < 0) <=> q < -exp
		// the result is +0 or -0
		res.w[1] = x_sign | 0x3040000000000000
		res.w[0] = 0x0000000000000000
		return res
	}
}

// Bid128RoundIntegralNegative is ported mechanically from bid128_round_integral.c: bid128_round_integral_negative.
func Bid128RoundIntegralNegative(x BID_UINT128, pfpsf *uint32) BID_UINT128 {
	var res BID_UINT128
	var fstar BID_UINT256
	var P256 BID_UINT256

	res, done, x_sign, x_exp, C1, q, exp := bid128_round_integral_nornd_common(x, pfpsf)
	if done {
		return res
	}

	// x is not special and is not zero

	// if (exp <= -p) return -1.0 or +0.0
	if x_exp <= 0x2ffc000000000000 { // 0x2ffc000000000000 == -34
		if x_sign != 0 {
			// if negative, return negative 1, because we know the coefficient
			// is non-zero (would have been caught above)
			res.w[1] = 0xb040000000000000
			res.w[0] = 0x0000000000000001
		} else {
			// if positive, return positive 0, because we know coefficient is
			// non-zero (would have been caught above)
			res.w[1] = 0x3040000000000000
			res.w[0] = 0x0000000000000000
		}
		return res
	}

	if exp >= 0 { // -exp <= 0
		// the argument is an integer already
		res.w[1] = x.w[1]
		res.w[0] = x.w[0]
		return res
	} else if (q + exp) > 0 { // exp < 0 and 1 <= -exp < q
		// need to shift right -exp digits from the coefficient; the exp will be 0
		ind := -exp // 1 <= ind <= 34; ind is a synonym for 'x'
		P256 = __mul_128x128_to_256(C1, bid_ten2mk128[ind-1])
		if ind-1 <= 2 { // 0 <= ind - 1 <= 2 => shift = 0
			res.w[1] = P256.w[3]
			res.w[0] = P256.w[2]
			// if positive, the truncated value is already the correct result
			if x_sign != 0 { // if negative
				if (P256.w[1] > bid_ten2mk128[ind-1].w[1]) ||
					(P256.w[1] == bid_ten2mk128[ind-1].w[1] &&
						(P256.w[0] >= bid_ten2mk128[ind-1].w[0])) {
					res.w[0]++
					if res.w[0] == 0 {
						res.w[1]++
					}
				}
			}
		} else if ind-1 <= 21 { // 3 <= ind - 1 <= 21 => 3 <= shift <= 63
			shift := bid_shiftright128[ind-1] // 0 <= shift <= 102
			res.w[1] = (P256.w[3] >> uint(shift))
			res.w[0] = (P256.w[3] << uint(64-shift)) | (P256.w[2] >> uint(shift))
			// if positive, the truncated value is already the correct result
			if x_sign != 0 { // if negative
				fstar.w[2] = P256.w[2] & bid_maskhigh128[ind-1]
				fstar.w[1] = P256.w[1]
				fstar.w[0] = P256.w[0]
				if fstar.w[2] != 0 || fstar.w[1] > bid_ten2mk128[ind-1].w[1] ||
					(fstar.w[1] == bid_ten2mk128[ind-1].w[1] &&
						fstar.w[0] >= bid_ten2mk128[ind-1].w[0]) {
					res.w[0]++
					if res.w[0] == 0 {
						res.w[1]++
					}
				}
			}
		} else { // 22 <= ind - 1 <= 33
			shift := bid_shiftright128[ind-1] - 64 // 2 <= shift <= 38
			res.w[1] = 0
			res.w[0] = P256.w[3] >> uint(shift)
			// if positive, the truncated value is already the correct result
			if x_sign != 0 { // if negative
				fstar.w[3] = P256.w[3] & bid_maskhigh128[ind-1]
				fstar.w[2] = P256.w[2]
				fstar.w[1] = P256.w[1]
				fstar.w[0] = P256.w[0]
				if fstar.w[3] != 0 || fstar.w[2] != 0 ||
					fstar.w[1] > bid_ten2mk128[ind-1].w[1] ||
					(fstar.w[1] == bid_ten2mk128[ind-1].w[1] &&
						fstar.w[0] >= bid_ten2mk128[ind-1].w[0]) {
					res.w[0]++
					if res.w[0] == 0 {
						res.w[1]++
					}
				}
			}
		}
		res.w[1] = x_sign | 0x3040000000000000 | res.w[1]
		return res
	} else { // if exp < 0 and q + exp <= 0
		if x_sign != 0 { // negative rounds down to -1.0
			res.w[1] = 0xb040000000000000
			res.w[0] = 0x0000000000000001
		} else { // positive rounds down to +0.0
			res.w[1] = 0x3040000000000000
			res.w[0] = 0x0000000000000000
		}
		return res
	}
}

// Bid128RoundIntegralPositive is ported mechanically from bid128_round_integral.c: bid128_round_integral_positive.
func Bid128RoundIntegralPositive(x BID_UINT128, pfpsf *uint32) BID_UINT128 {
	var res BID_UINT128
	var fstar BID_UINT256
	var P256 BID_UINT256

	res, done, x_sign, x_exp, C1, q, exp := bid128_round_integral_nornd_common(x, pfpsf)
	if done {
		return res
	}

	// x is not special and is not zero

	// if (exp <= -p) return -0.0 or +1.0
	if x_exp <= 0x2ffc000000000000 { // 0x2ffc000000000000 == -34
		if x_sign != 0 {
			// if negative, return negative 0, because we know the coefficient
			// is non-zero (would have been caught above)
			res.w[1] = 0xb040000000000000
			res.w[0] = 0x0000000000000000
		} else {
			// if positive, return positive 1, because we know coefficient is
			// non-zero (would have been caught above)
			res.w[1] = 0x3040000000000000
			res.w[0] = 0x0000000000000001
		}
		return res
	}

	if exp >= 0 { // -exp <= 0
		// the argument is an integer already
		res.w[1] = x.w[1]
		res.w[0] = x.w[0]
		return res
	} else if (q + exp) > 0 { // exp < 0 and 1 <= -exp < q
		// need to shift right -exp digits from the coefficient; exp will be 0
		ind := -exp // 1 <= ind <= 34; ind is a synonym for 'x'
		P256 = __mul_128x128_to_256(C1, bid_ten2mk128[ind-1])
		if ind-1 <= 2 { // 0 <= ind - 1 <= 2 => shift = 0
			res.w[1] = P256.w[3]
			res.w[0] = P256.w[2]
			// if negative, the truncated value is already the correct result
			if x_sign == 0 { // if positive
				if (P256.w[1] > bid_ten2mk128[ind-1].w[1]) ||
					(P256.w[1] == bid_ten2mk128[ind-1].w[1] &&
						(P256.w[0] >= bid_ten2mk128[ind-1].w[0])) {
					res.w[0]++
					if res.w[0] == 0 {
						res.w[1]++
					}
				}
			}
		} else if ind-1 <= 21 { // 3 <= ind - 1 <= 21 => 3 <= shift <= 63
			shift := bid_shiftright128[ind-1] // 3 <= shift <= 63
			res.w[1] = (P256.w[3] >> uint(shift))
			res.w[0] = (P256.w[3] << uint(64-shift)) | (P256.w[2] >> uint(shift))
			// if negative, the truncated value is already the correct result
			if x_sign == 0 { // if positive
				fstar.w[2] = P256.w[2] & bid_maskhigh128[ind-1]
				fstar.w[1] = P256.w[1]
				fstar.w[0] = P256.w[0]
				if fstar.w[2] != 0 || fstar.w[1] > bid_ten2mk128[ind-1].w[1] ||
					(fstar.w[1] == bid_ten2mk128[ind-1].w[1] &&
						fstar.w[0] >= bid_ten2mk128[ind-1].w[0]) {
					res.w[0]++
					if res.w[0] == 0 {
						res.w[1]++
					}
				}
			}
		} else { // 22 <= ind - 1 <= 33
			shift := bid_shiftright128[ind-1] - 64 // 2 <= shift <= 38
			res.w[1] = 0
			res.w[0] = P256.w[3] >> uint(shift)
			// if negative, the truncated value is already the correct result
			if x_sign == 0 { // if positive
				fstar.w[3] = P256.w[3] & bid_maskhigh128[ind-1]
				fstar.w[2] = P256.w[2]
				fstar.w[1] = P256.w[1]
				fstar.w[0] = P256.w[0]
				if fstar.w[3] != 0 || fstar.w[2] != 0 ||
					fstar.w[1] > bid_ten2mk128[ind-1].w[1] ||
					(fstar.w[1] == bid_ten2mk128[ind-1].w[1] &&
						fstar.w[0] >= bid_ten2mk128[ind-1].w[0]) {
					res.w[0]++
					if res.w[0] == 0 {
						res.w[1]++
					}
				}
			}
		}
		res.w[1] = x_sign | 0x3040000000000000 | res.w[1]
		return res
	} else { // if exp < 0 and q + exp <= 0
		if x_sign != 0 { // negative rounds up to -0.0
			res.w[1] = 0xb040000000000000
			res.w[0] = 0x0000000000000000
		} else { // positive rounds up to +1.0
			res.w[1] = 0x3040000000000000
			res.w[0] = 0x0000000000000001
		}
		return res
	}
}

// Bid128RoundIntegralZero is ported mechanically from bid128_round_integral.c: bid128_round_integral_zero.
func Bid128RoundIntegralZero(x BID_UINT128, pfpsf *uint32) BID_UINT128 {
	var res BID_UINT128
	var P256 BID_UINT256

	res, done, x_sign, x_exp, C1, q, exp := bid128_round_integral_nornd_common(x, pfpsf)
	if done {
		return res
	}

	// x is not special and is not zero

	// if (exp <= -p) return -0.0 or +0.0
	if x_exp <= 0x2ffc000000000000 { // 0x2ffc000000000000 == -34
		res.w[1] = x_sign | 0x3040000000000000
		res.w[0] = 0x0000000000000000
		return res
	}

	if exp >= 0 { // -exp <= 0
		// the argument is an integer already
		res.w[1] = x.w[1]
		res.w[0] = x.w[0]
		return res
	} else if (q + exp) > 0 { // exp < 0 and 1 <= -exp < q
		// need to shift right -exp digits from the coefficient; the exp will be 0
		ind := -exp // 1 <= ind <= 34; ind is a synonym for 'x'
		P256 = __mul_128x128_to_256(C1, bid_ten2mk128[ind-1])
		if ind-1 <= 2 { // 0 <= ind - 1 <= 2 => shift = 0
			res.w[1] = P256.w[3]
			res.w[0] = P256.w[2]
		} else if ind-1 <= 21 { // 3 <= ind - 1 <= 21 => 3 <= shift <= 63
			shift := bid_shiftright128[ind-1] // 3 <= shift <= 63
			res.w[1] = (P256.w[3] >> uint(shift))
			res.w[0] = (P256.w[3] << uint(64-shift)) | (P256.w[2] >> uint(shift))
		} else { // 22 <= ind - 1 <= 33
			shift := bid_shiftright128[ind-1] - 64 // 2 <= shift <= 38
			res.w[1] = 0
			res.w[0] = P256.w[3] >> uint(shift)
		}
		res.w[1] = x_sign | 0x3040000000000000 | res.w[1]
		return res
	} else { // if exp < 0 and q + exp <= 0 the result is +0 or -0
		res.w[1] = x_sign | 0x3040000000000000
		res.w[0] = 0x0000000000000000
		return res
	}
}

// Bid128RoundIntegralNearestAway is ported mechanically from bid128_round_integral.c: bid128_round_integral_nearest_away.
func Bid128RoundIntegralNearestAway(x BID_UINT128, pfpsf *uint32) BID_UINT128 {
	var res BID_UINT128
	var P256 BID_UINT256

	res, done, x_sign, _, C1, q, exp := bid128_round_integral_nornd_common(x, pfpsf)
	if done {
		return res
	}

	// x is not special and is not zero

	// if (exp <= -(p+1)) return 0.0
	if x.w[1]&MASK_EXP_128 <= 0x2ffa000000000000 { // 0x2ffa000000000000 == -35
		res.w[1] = x_sign | 0x3040000000000000
		res.w[0] = 0x0000000000000000
		return res
	}

	if exp >= 0 { // -exp <= 0
		// the argument is an integer already
		res.w[1] = x.w[1]
		res.w[0] = x.w[0]
		return res
	} else if (q + exp) >= 0 { // exp < 0 and 1 <= -exp <= q
		// need to shift right -exp digits from the coefficient; the exp will be 0
		ind := -exp // 1 <= ind <= 34; ind is a synonym for 'x'
		// chop off ind digits from the lower part of C1
		// C1 = C1 + 1/2 * 10^x where the result C1 fits in 127 bits
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
		// shift right C* by Ex-128 = bid_shiftright128[ind]
		if ind-1 <= 2 { // 0 <= ind - 1 <= 2 => shift = 0
			res.w[1] = P256.w[3]
			res.w[0] = P256.w[2]
		} else if ind-1 <= 21 { // 3 <= ind - 1 <= 21 => 3 <= shift <= 63
			shift := bid_shiftright128[ind-1] // 3 <= shift <= 63
			res.w[0] = (P256.w[3] << uint(64-shift)) | (P256.w[2] >> uint(shift))
			res.w[1] = (P256.w[3] >> uint(shift))
		} else { // 22 <= ind - 1 <= 33
			shift := bid_shiftright128[ind-1] // 2 <= shift <= 38
			res.w[1] = 0
			res.w[0] = (P256.w[3] >> uint(shift-64)) // 2 <= shift - 64 <= 38
		}
		// if the result was a midpoint, it was already rounded away from zero
		res.w[1] |= x_sign | 0x3040000000000000
		return res
	} else { // if ((q + exp) < 0) <=> q < -exp
		// the result is +0 or -0
		res.w[1] = x_sign | 0x3040000000000000
		res.w[0] = 0x0000000000000000
		return res
	}
}
