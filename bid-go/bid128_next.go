// Ported from: Intel bid128_next.c + bid128_nexttowardd.c
// Mechanical translation - all logic preserved exactly.

package bidgo

import "math"

// bid128 next constants from bid_internal.h
const (
	EXP_P1     = 0x0002000000000000
	EXP_MAX_P1 = 0x6000000000000000
	EXP_MIN    = 0x0000000000000000
)

// Bid128NextUp - Intel bid128_nextup 기계적 포팅
func Bid128NextUp(x BID_UINT128) (BID_UINT128, uint32) {
	var res BID_UINT128
	var x_sign uint64
	var x_exp uint64
	var exp int
	var x_nr_bits int
	var q1, ind int
	var C1 BID_UINT128 // C1.w[1], C1.w[0] represent x_signif_hi, x_signif_lo
	var pfpsf uint32

	// unpack the argument
	x_sign = x.w[1] & MASK_SIGN_128 // 0 for positive, MASK_SIGN for negative
	C1.w[1] = x.w[1] & MASK_COEFF_128
	C1.w[0] = x.w[0]

	// check for NaN or Infinity
	if (x.w[1] & MASK_INF_128) == MASK_INF_128 {
		// x is special
		if (x.w[1] & MASK_NAN_128) == MASK_NAN_128 { // x is NAN
			// if x = NaN, then res = Q (x)
			// check first for non-canonical NaN payload
			if ((x.w[1] & 0x00003fffffffffff) > 0x0000314dc6448d93) ||
				(((x.w[1] & 0x00003fffffffffff) == 0x0000314dc6448d93) &&
					(x.w[0] > 0x38c15b09ffffffff)) {
				x.w[1] = x.w[1] & 0xffffc00000000000
				x.w[0] = 0x0
			}
			if (x.w[1] & MASK_SNAN_128) == MASK_SNAN_128 { // x is SNAN
				// set invalid flag
				pfpsf |= BID_INVALID_EXCEPTION
				// return quiet (x)
				res.w[1] = x.w[1] & 0xfc003fffffffffff // clear out also G[6]-G[16]
				res.w[0] = x.w[0]
			} else { // x is QNaN
				// return x
				res.w[1] = x.w[1] & 0xfc003fffffffffff // clear out G[6]-G[16]
				res.w[0] = x.w[0]
			}
		} else { // x is not NaN, so it must be infinity
			if x_sign == 0 { // x is +inf
				res.w[1] = 0x7800000000000000 // +inf
				res.w[0] = 0x0000000000000000
			} else { // x is -inf
				res.w[1] = 0xdfffed09bead87c0 // -MAXFP = -999...99 * 10^emax
				res.w[0] = 0x378d8e63ffffffff
			}
		}
		return res, pfpsf
	}
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

	if (C1.w[1] == 0x0) && (C1.w[0] == 0x0) {
		// x is +/-0
		res.w[1] = 0x0000000000000000 // +1 * 10^emin
		res.w[0] = 0x0000000000000001
	} else { // x is not special and is not zero
		if x.w[1] == 0x5fffed09bead87c0 &&
			x.w[0] == 0x378d8e63ffffffff {
			// x = +MAXFP = 999...99 * 10^emax
			res.w[1] = 0x7800000000000000 // +inf
			res.w[0] = 0x0000000000000000
		} else if x.w[1] == 0x8000000000000000 &&
			x.w[0] == 0x0000000000000001 {
			// x = -MINFP = 1...99 * 10^emin
			res.w[1] = 0x8000000000000000 // -0
			res.w[0] = 0x0000000000000000
		} else { // -MAXFP <= x <= -MINFP - 1 ulp OR MINFP <= x <= MAXFP - 1 ulp
			// can add/subtract 1 ulp to the significand

			// Note: we could check here if x >= 10^34 to speed up the case q1 = 34
			// q1 = nr. of decimal digits in x
			// determine first the nr. of bits in x
			if C1.w[1] == 0 {
				if C1.w[0] >= 0x0020000000000000 { // x >= 2^53
					// split the 64-bit value in two 32-bit halves to avoid rnd errors
					if C1.w[0] >= 0x0000000100000000 { // x >= 2^32
						tmp1 := math.Float64bits(float64(C1.w[0] >> 32)) // exact conversion
						x_nr_bits =
							33 + int(((uint32(tmp1>>52))&0x7ff)-
								0x3ff)
					} else { // x < 2^32
						tmp1 := math.Float64bits(float64(C1.w[0])) // exact conversion
						x_nr_bits =
							1 + int(((uint32(tmp1>>52))&0x7ff)-
								0x3ff)
					}
				} else { // if x < 2^53
					tmp1 := math.Float64bits(float64(C1.w[0])) // exact conversion
					x_nr_bits =
						1 + int(((uint32(tmp1>>52))&0x7ff)-0x3ff)
				}
			} else { // C1.w[1] != 0 => nr. bits = 64 + nr_bits (C1.w[1])
				tmp1 := math.Float64bits(float64(C1.w[1])) // exact conversion
				x_nr_bits =
					65 + int(((uint32(tmp1>>52))&0x7ff)-0x3ff)
			}
			q1 = int(bid_nr_digits[x_nr_bits-1].digits)
			if q1 == 0 {
				q1 = int(bid_nr_digits[x_nr_bits-1].digits1)
				if C1.w[1] > bid_nr_digits[x_nr_bits-1].threshold_hi ||
					(C1.w[1] == bid_nr_digits[x_nr_bits-1].threshold_hi &&
						C1.w[0] >= bid_nr_digits[x_nr_bits-1].threshold_lo) {
					q1++
				}
			}
			// if q1 < P34 then pad the significand with zeros
			if q1 < P34 {
				exp = int(x_exp>>49) - 6176
				if exp+6176 > P34-q1 {
					ind = P34 - q1 // 1 <= ind <= P34 - 1
					// pad with P34 - q1 zeros, until exponent = emin
					// C1 = C1 * 10^ind
					if q1 <= 19 { // 64-bit C1
						if ind <= 19 { // 64-bit 10^ind and 64-bit C1
							C1 = __mul_64x64_to_128(C1.w[0], bid_ten2k64[ind])
						} else { // 128-bit 10^ind and 64-bit C1
							C1 = __mul_128x64_to_128(C1.w[0], bid_ten2k128[ind-20])
						}
					} else { // C1 is (most likely) 128-bit
						if ind <= 14 { // 64-bit 10^ind and 128-bit C1 (most likely)
							C1 = __mul_128x64_to_128(bid_ten2k64[ind], C1)
						} else if ind <= 19 { // 64-bit 10^ind and 64-bit C1 (q1 <= 19)
							C1 = __mul_64x64_to_128(C1.w[0], bid_ten2k64[ind])
						} else { // 128-bit 10^ind and 64-bit C1 (C1 must be 64-bit)
							C1 = __mul_128x64_to_128(C1.w[0], bid_ten2k128[ind-20])
						}
					}
					x_exp = x_exp - (uint64(ind) << 49)
				} else { // pad with zeros until the exponent reaches emin
					ind = exp + 6176
					// C1 = C1 * 10^ind
					if ind <= 19 { // 1 <= P34 - q1 <= 19 <=> 15 <= q1 <= 33
						if q1 <= 19 { // 64-bit C1, 64-bit 10^ind
							C1 = __mul_64x64_to_128(C1.w[0], bid_ten2k64[ind])
						} else { // 20 <= q1 <= 33 => 128-bit C1, 64-bit 10^ind
							C1 = __mul_128x64_to_128(bid_ten2k64[ind], C1)
						}
					} else { // if 20 <= P34 - q1 <= 33 <=> 1 <= q1 <= 14 =>
						// 64-bit C1, 128-bit 10^ind
						C1 = __mul_128x64_to_128(C1.w[0], bid_ten2k128[ind-20])
					}
					x_exp = EXP_MIN
				}
			}
			if x_sign == 0 { // x > 0
				// add 1 ulp (add 1 to the significand)
				C1.w[0]++
				if C1.w[0] == 0 {
					C1.w[1]++
				}
				if C1.w[1] == 0x0001ed09bead87c0 && C1.w[0] == 0x378d8e6400000000 { // if  C1 = 10^34
					C1.w[1] = 0x0000314dc6448d93 // C1 = 10^33
					C1.w[0] = 0x38c15b0a00000000
					x_exp = x_exp + EXP_P1
				}
			} else { // x < 0
				// subtract 1 ulp (subtract 1 from the significand)
				C1.w[0]--
				if C1.w[0] == 0xffffffffffffffff {
					C1.w[1]--
				}
				if x_exp != 0 && C1.w[1] == 0x0000314dc6448d93 && C1.w[0] == 0x38c15b09ffffffff { // if  C1 = 10^33 - 1
					C1.w[1] = 0x0001ed09bead87c0 // C1 = 10^34 - 1
					C1.w[0] = 0x378d8e63ffffffff
					x_exp = x_exp - EXP_P1
				}
			}
			// assemble the result
			res.w[1] = x_sign | x_exp | C1.w[1]
			res.w[0] = C1.w[0]
		} // end -MAXFP <= x <= -MINFP - 1 ulp OR MINFP <= x <= MAXFP - 1 ulp
	} // end x is not special and is not zero
	return res, pfpsf
}

// Bid128NextDown - Intel bid128_nextdown 기계적 포팅
func Bid128NextDown(x BID_UINT128) (BID_UINT128, uint32) {
	var res BID_UINT128
	var x_sign uint64
	var x_exp uint64
	var exp int
	var x_nr_bits int
	var q1, ind int
	var C1 BID_UINT128 // C1.w[1], C1.w[0] represent x_signif_hi, x_signif_lo
	var pfpsf uint32

	// unpack the argument
	x_sign = x.w[1] & MASK_SIGN_128 // 0 for positive, MASK_SIGN for negative
	C1.w[1] = x.w[1] & MASK_COEFF_128
	C1.w[0] = x.w[0]

	// check for NaN or Infinity
	if (x.w[1] & MASK_INF_128) == MASK_INF_128 {
		// x is special
		if (x.w[1] & MASK_NAN_128) == MASK_NAN_128 { // x is NAN
			// if x = NaN, then res = Q (x)
			// check first for non-canonical NaN payload
			if ((x.w[1] & 0x00003fffffffffff) > 0x0000314dc6448d93) ||
				(((x.w[1] & 0x00003fffffffffff) == 0x0000314dc6448d93) &&
					(x.w[0] > 0x38c15b09ffffffff)) {
				x.w[1] = x.w[1] & 0xffffc00000000000
				x.w[0] = 0x0
			}
			if (x.w[1] & MASK_SNAN_128) == MASK_SNAN_128 { // x is SNAN
				// set invalid flag
				pfpsf |= BID_INVALID_EXCEPTION
				// return quiet (x)
				res.w[1] = x.w[1] & 0xfc003fffffffffff // clear out also G[6]-G[16]
				res.w[0] = x.w[0]
			} else { // x is QNaN
				// return x
				res.w[1] = x.w[1] & 0xfc003fffffffffff // clear out G[6]-G[16]
				res.w[0] = x.w[0]
			}
		} else { // x is not NaN, so it must be infinity
			if x_sign == 0 { // x is +inf
				res.w[1] = 0x5fffed09bead87c0 // +MAXFP = +999...99 * 10^emax
				res.w[0] = 0x378d8e63ffffffff
			} else { // x is -inf
				res.w[1] = 0xf800000000000000 // -inf
				res.w[0] = 0x0000000000000000
			}
		}
		return res, pfpsf
	}
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

	if (C1.w[1] == 0x0) && (C1.w[0] == 0x0) {
		// x is +/-0
		res.w[1] = 0x8000000000000000 // -1 * 10^emin
		res.w[0] = 0x0000000000000001
	} else { // x is not special and is not zero
		if x.w[1] == 0xdfffed09bead87c0 &&
			x.w[0] == 0x378d8e63ffffffff {
			// x = -MAXFP = -999...99 * 10^emax
			res.w[1] = 0xf800000000000000 // -inf
			res.w[0] = 0x0000000000000000
		} else if x.w[1] == 0x0 && x.w[0] == 0x0000000000000001 { // +MINFP
			res.w[1] = 0x0000000000000000 // +0
			res.w[0] = 0x0000000000000000
		} else { // -MAXFP <= x <= -MINFP - 1 ulp OR MINFP <= x <= MAXFP - 1 ulp
			// can add/subtract 1 ulp to the significand

			// Note: we could check here if x >= 10^34 to speed up the case q1 = 34
			// q1 = nr. of decimal digits in x
			// determine first the nr. of bits in x
			if C1.w[1] == 0 {
				if C1.w[0] >= 0x0020000000000000 { // x >= 2^53
					// split the 64-bit value in two 32-bit halves to avoid rnd errors
					if C1.w[0] >= 0x0000000100000000 { // x >= 2^32
						tmp1 := math.Float64bits(float64(C1.w[0] >> 32)) // exact conversion
						x_nr_bits =
							33 + int(((uint32(tmp1>>52))&0x7ff)-
								0x3ff)
					} else { // x < 2^32
						tmp1 := math.Float64bits(float64(C1.w[0])) // exact conversion
						x_nr_bits =
							1 + int(((uint32(tmp1>>52))&0x7ff)-
								0x3ff)
					}
				} else { // if x < 2^53
					tmp1 := math.Float64bits(float64(C1.w[0])) // exact conversion
					x_nr_bits =
						1 + int(((uint32(tmp1>>52))&0x7ff)-0x3ff)
				}
			} else { // C1.w[1] != 0 => nr. bits = 64 + nr_bits (C1.w[1])
				tmp1 := math.Float64bits(float64(C1.w[1])) // exact conversion
				x_nr_bits =
					65 + int(((uint32(tmp1>>52))&0x7ff)-0x3ff)
			}
			q1 = int(bid_nr_digits[x_nr_bits-1].digits)
			if q1 == 0 {
				q1 = int(bid_nr_digits[x_nr_bits-1].digits1)
				if C1.w[1] > bid_nr_digits[x_nr_bits-1].threshold_hi ||
					(C1.w[1] == bid_nr_digits[x_nr_bits-1].threshold_hi &&
						C1.w[0] >= bid_nr_digits[x_nr_bits-1].threshold_lo) {
					q1++
				}
			}
			// if q1 < P then pad the significand with zeros
			if q1 < P34 {
				exp = int(x_exp>>49) - 6176
				if exp+6176 > P34-q1 {
					ind = P34 - q1 // 1 <= ind <= P34 - 1
					// pad with P34 - q1 zeros, until exponent = emin
					// C1 = C1 * 10^ind
					if q1 <= 19 { // 64-bit C1
						if ind <= 19 { // 64-bit 10^ind and 64-bit C1
							C1 = __mul_64x64_to_128(C1.w[0], bid_ten2k64[ind])
						} else { // 128-bit 10^ind and 64-bit C1
							C1 = __mul_128x64_to_128(C1.w[0], bid_ten2k128[ind-20])
						}
					} else { // C1 is (most likely) 128-bit
						if ind <= 14 { // 64-bit 10^ind and 128-bit C1 (most likely)
							C1 = __mul_128x64_to_128(bid_ten2k64[ind], C1)
						} else if ind <= 19 { // 64-bit 10^ind and 64-bit C1 (q1 <= 19)
							C1 = __mul_64x64_to_128(C1.w[0], bid_ten2k64[ind])
						} else { // 128-bit 10^ind and 64-bit C1 (C1 must be 64-bit)
							C1 = __mul_128x64_to_128(C1.w[0], bid_ten2k128[ind-20])
						}
					}
					x_exp = x_exp - (uint64(ind) << 49)
				} else { // pad with zeros until the exponent reaches emin
					ind = exp + 6176
					// C1 = C1 * 10^ind
					if ind <= 19 { // 1 <= P34 - q1 <= 19 <=> 15 <= q1 <= 33
						if q1 <= 19 { // 64-bit C1, 64-bit 10^ind
							C1 = __mul_64x64_to_128(C1.w[0], bid_ten2k64[ind])
						} else { // 20 <= q1 <= 33 => 128-bit C1, 64-bit 10^ind
							C1 = __mul_128x64_to_128(bid_ten2k64[ind], C1)
						}
					} else { // if 20 <= P34 - q1 <= 33 <=> 1 <= q1 <= 14 =>
						// 64-bit C1, 128-bit 10^ind
						C1 = __mul_128x64_to_128(C1.w[0], bid_ten2k128[ind-20])
					}
					x_exp = EXP_MIN
				}
			}
			if x_sign != 0 { // x < 0
				// add 1 ulp (add 1 to the significand)
				C1.w[0]++
				if C1.w[0] == 0 {
					C1.w[1]++
				}
				if C1.w[1] == 0x0001ed09bead87c0 && C1.w[0] == 0x378d8e6400000000 { // if  C1 = 10^34
					C1.w[1] = 0x0000314dc6448d93 // C1 = 10^33
					C1.w[0] = 0x38c15b0a00000000
					x_exp = x_exp + EXP_P1
				}
			} else { // x > 0
				// subtract 1 ulp (subtract 1 from the significand)
				C1.w[0]--
				if C1.w[0] == 0xffffffffffffffff {
					C1.w[1]--
				}
				if x_exp != 0 && C1.w[1] == 0x0000314dc6448d93 && C1.w[0] == 0x38c15b09ffffffff { // if  C1 = 10^33 - 1
					C1.w[1] = 0x0001ed09bead87c0 // C1 = 10^34 - 1
					C1.w[0] = 0x378d8e63ffffffff
					x_exp = x_exp - EXP_P1
				}
			}
			// assemble the result
			res.w[1] = x_sign | x_exp | C1.w[1]
			res.w[0] = C1.w[0]
		} // end -MAXFP <= x <= -MINFP - 1 ulp OR MINFP <= x <= MAXFP - 1 ulp
	} // end x is not special and is not zero
	return res, pfpsf
}

// Bid128NextAfter - Intel bid128_nextafter 기계적 포팅
func Bid128NextAfter(x, y BID_UINT128) (BID_UINT128, uint32) {
	var xnswp BID_UINT128 = x
	var ynswp BID_UINT128 = y
	var res BID_UINT128
	var tmp1, tmp2, tmp3 BID_UINT128
	var tmp_fpsf uint32 // dummy fpsf for calls to comparison functions
	var res1, res2 int
	var x_exp uint64
	var pfpsf uint32

	_ = xnswp
	_ = ynswp
	// check for NaNs
	if ((x.w[1] & MASK_INF_128) == MASK_INF_128) ||
		((y.w[1] & MASK_INF_128) == MASK_INF_128) {
		// x is special or y is special
		if (x.w[1] & MASK_NAN_128) == MASK_NAN_128 { // x is NAN
			// if x = NaN, then res = Q (x)
			// check first for non-canonical NaN payload
			if ((x.w[1] & 0x00003fffffffffff) > 0x0000314dc6448d93) ||
				(((x.w[1] & 0x00003fffffffffff) == 0x0000314dc6448d93) &&
					(x.w[0] > 0x38c15b09ffffffff)) {
				x.w[1] = x.w[1] & 0xffffc00000000000
				x.w[0] = 0x0
			}
			if (x.w[1] & MASK_SNAN_128) == MASK_SNAN_128 { // x is SNAN
				// set invalid flag
				pfpsf |= BID_INVALID_EXCEPTION
				// return quiet (x)
				res.w[1] = x.w[1] & 0xfc003fffffffffff // clear out also G[6]-G[16]
				res.w[0] = x.w[0]
			} else { // x is QNaN
				// return x
				res.w[1] = x.w[1] & 0xfc003fffffffffff // clear out G[6]-G[16]
				res.w[0] = x.w[0]
				if (y.w[1] & MASK_SNAN_128) == MASK_SNAN_128 { // y is SNAN
					// set invalid flag
					pfpsf |= BID_INVALID_EXCEPTION
				}
			}
			return res, pfpsf
		} else if (y.w[1] & MASK_NAN_128) == MASK_NAN_128 { // y is NAN
			// if x = NaN, then res = Q (x)
			// check first for non-canonical NaN payload
			if ((y.w[1] & 0x00003fffffffffff) > 0x0000314dc6448d93) ||
				(((y.w[1] & 0x00003fffffffffff) == 0x0000314dc6448d93) &&
					(y.w[0] > 0x38c15b09ffffffff)) {
				y.w[1] = y.w[1] & 0xffffc00000000000
				y.w[0] = 0x0
			}
			if (y.w[1] & MASK_SNAN_128) == MASK_SNAN_128 { // y is SNAN
				// set invalid flag
				pfpsf |= BID_INVALID_EXCEPTION
				// return quiet (x)
				res.w[1] = y.w[1] & 0xfc003fffffffffff // clear out also G[6]-G[16]
				res.w[0] = y.w[0]
			} else { // x is QNaN
				// return x
				res.w[1] = y.w[1] & 0xfc003fffffffffff // clear out G[6]-G[16]
				res.w[0] = y.w[0]
			}
			return res, pfpsf
		} else { // at least one is infinity
			if (x.w[1] & MASK_ANY_INF_128) == MASK_INF_128 { // x = inf
				x.w[1] = x.w[1] & (MASK_SIGN_128 | MASK_INF_128)
				x.w[0] = 0x0
			}
			if (y.w[1] & MASK_ANY_INF_128) == MASK_INF_128 { // y = inf
				y.w[1] = y.w[1] & (MASK_SIGN_128 | MASK_INF_128)
				y.w[0] = 0x0
			}
		}
	}
	// neither x nor y is NaN

	// if not infinity, check for non-canonical values x (treated as zero)
	if (x.w[1] & MASK_ANY_INF_128) != MASK_INF_128 { // x != inf
		if (x.w[1] & 0x6000000000000000) == 0x6000000000000000 { // G0_G1=11
			// non-canonical
			x_exp = (x.w[1] << 2) & MASK_EXP_128 // biased and shifted left 49 bits
			x.w[1] = (x.w[1] & MASK_SIGN_128) | x_exp
			x.w[0] = 0x0
		} else { // G0_G1 != 11
			x_exp = x.w[1] & MASK_EXP_128 // biased and shifted left 49 bits
			if (x.w[1]&MASK_COEFF_128) > 0x0001ed09bead87c0 ||
				((x.w[1]&MASK_COEFF_128) == 0x0001ed09bead87c0 &&
					x.w[0] > 0x378d8e63ffffffff) {
				// x is non-canonical if coefficient is larger than 10^34 -1
				x.w[1] = (x.w[1] & MASK_SIGN_128) | x_exp
				x.w[0] = 0x0
			} else { // canonical
				// ;
			}
		}
	}
	// no need to check for non-canonical y

	// neither x nor y is NaN
	tmp_fpsf = pfpsf // save fpsf
	xnswp = x
	ynswp = y
	res1, _ = Bid128QuietEqual(xnswp, ynswp)
	res2, _ = Bid128QuietGreater(xnswp, ynswp)
	pfpsf = tmp_fpsf // restore fpsf

	if res1 != 0 { // x = y
		// return x with the sign of y
		res.w[1] =
			(x.w[1] & 0x7fffffffffffffff) | (y.w[1] & 0x8000000000000000)
		res.w[0] = x.w[0]
	} else if res2 != 0 { // x > y
		res, tmp_fpsf = Bid128NextDown(xnswp)
		pfpsf |= tmp_fpsf
	} else { // x < y
		res, tmp_fpsf = Bid128NextUp(xnswp)
		pfpsf |= tmp_fpsf
	}
	// if the operand x is finite but the result is infinite, signal
	// overflow and inexact
	if ((x.w[1] & MASK_INF_128) != MASK_INF_128) &&
		((res.w[1] & MASK_INF_128) == MASK_INF_128) {
		// set the inexact flag
		pfpsf |= BID_INEXACT_EXCEPTION
		// set the overflow flag
		pfpsf |= BID_OVERFLOW_EXCEPTION
	}
	// if the result is in (-10^emin, 10^emin), and is different from the
	// operand x, signal underflow and inexact
	tmp1.w[1] = 0x0000314dc6448d93 // BID_HIGH_128W
	tmp1.w[0] = 0x38c15b0a00000000 // BID_LOW_128W  +100...0[34] * 10^emin
	tmp2.w[1] = res.w[1] & 0x7fffffffffffffff
	tmp2.w[0] = res.w[0]
	tmp3.w[1] = res.w[1]
	tmp3.w[0] = res.w[0]
	tmp_fpsf = pfpsf // save fpsf
	res1, _ = Bid128QuietGreater(tmp1, tmp2)
	res2, _ = Bid128QuietNotEqual(xnswp, tmp3)
	pfpsf = tmp_fpsf // restore fpsf
	if res1 != 0 && res2 != 0 {
		// set the inexact flag
		pfpsf |= BID_INEXACT_EXCEPTION
		// set the underflow flag
		pfpsf |= BID_UNDERFLOW_EXCEPTION
	}
	return res, pfpsf
}

// Bid128NextToward - Intel bid128_nexttoward 기계적 포팅
// Note: same as bid128_nextafter
func Bid128NextToward(x, y BID_UINT128) (BID_UINT128, uint32) {
	res, pfpsf := Bid128NextAfter(x, y)
	return res, pfpsf
}
