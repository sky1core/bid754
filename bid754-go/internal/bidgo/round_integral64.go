// Ported from: IntelRDFPMathLib20U4/LIBRARY/src/bid64_round_integral.c
// (Bid64NearbyInt follows bid64_nearbyintd.c; the local rounding tables are
// copied from the tables in bid128.c)
// Version: Intel(R) Decimal Floating-Point Math Library 2.0 Update 4
//
// Derived from the Intel BID library: per-function control flow, rounding
// tables, magic constants, and comments follow the C source; the digit
// count uses bid_estimate_decimal_digits and bid_power10_table_128 instead
// of the C bid_nr_digits structure.

package bidgo

import "math"

var bid_shiftright128_round64 = [22]int{
	0,
	0,
	0,
	3,
	6,
	9,
	13,
	16,
	19,
	23,
	26,
	29,
	33,
	36,
	39,
	43,
	46,
	49,
	53,
	56,
	59,
	63,
}

var bid_maskhigh128_round64 = [22]uint64{
	0x0000000000000000,
	0x0000000000000000,
	0x0000000000000000,
	0x0000000000000007,
	0x000000000000003f,
	0x00000000000001ff,
	0x0000000000001fff,
	0x000000000000ffff,
	0x000000000007ffff,
	0x00000000007fffff,
	0x0000000003ffffff,
	0x000000001fffffff,
	0x00000001ffffffff,
	0x0000000fffffffff,
	0x0000007fffffffff,
	0x000007ffffffffff,
	0x00003fffffffffff,
	0x0001ffffffffffff,
	0x001fffffffffffff,
	0x00ffffffffffffff,
	0x07ffffffffffffff,
	0x7fffffffffffffff,
}

var bid_onehalf128_round64 = [22]uint64{
	0x0000000000000000,
	0x0000000000000000,
	0x0000000000000000,
	0x0000000000000004,
	0x0000000000000020,
	0x0000000000000100,
	0x0000000000001000,
	0x0000000000008000,
	0x0000000000040000,
	0x0000000000400000,
	0x0000000002000000,
	0x0000000010000000,
	0x0000000100000000,
	0x0000000800000000,
	0x0000004000000000,
	0x0000040000000000,
	0x0000200000000000,
	0x0001000000000000,
	0x0010000000000000,
	0x0080000000000000,
	0x0400000000000000,
	0x4000000000000000,
}

var bid_ten2mk64_round64 = [16]uint64{
	0x199999999999999a,
	0x028f5c28f5c28f5d,
	0x004189374bc6a7f0,
	0x00346dc5d638865a,
	0x0029f16b11c6d1e2,
	0x00218def416bdb1b,
	0x0035afe535795e91,
	0x002af31dc4611874,
	0x00225c17d04dad2a,
	0x0036f9bfb3af7b76,
	0x002bfaffc2f2c92b,
	0x00232f33025bd423,
	0x00384b84d092ed04,
	0x002d09370d425737,
	0x0024075f3dceac2c,
	0x0039a5652fb11379,
}

// Bid64RoundIntegralExact is ported mechanically from bid64_round_integral.c.
func Bid64RoundIntegralExact(x uint64, rndMode int) (uint64, uint32) {
	var res uint64 = 0xbaddbaddbaddbadd
	var x_sign uint64
	var x_nr_bits int
	var q, ind, shift int
	var C1 uint64
	var fstar BID_UINT128
	var P128 BID_UINT128
	var pfpsf uint32
	var exp int
	var tmp1 uint64

	x_sign = x & MASK_SIGN64 // 0 for positive, MASK_SIGN for negative

	// check for NaNs and infinities
	if (x & MASK_NAN64) == MASK_NAN64 { // check for NaN
		if (x & 0x0003ffffffffffff) > 999999999999999 {
			x = x & 0xfe00000000000000 // clear G6-G12 and the payload bits
		} else {
			x = x & 0xfe03ffffffffffff // clear G6-G12
		}
		if (x & MASK_SNAN64) == MASK_SNAN64 { // SNaN
			// set invalid flag
			pfpsf |= BID_INVALID_EXCEPTION
			// return quiet (SNaN)
			res = x & 0xfdffffffffffffff
		} else { // QNaN
			res = x
		}
		return res, pfpsf
	} else if (x & MASK_INF64) == MASK_INF64 { // check for Infinity
		res = x_sign | 0x7800000000000000
		return res, pfpsf
	}
	// unpack x
	if (x & MASK_STEERING_BITS64) == MASK_STEERING_BITS64 {
		// if the steering bits are 11 (condition will be 0), then
		// the exponent is G[0:w+1]
		exp = int((x&MASK_BINARY_EXPONENT2_64)>>51) - 398
		C1 = (x & MASK_BINARY_SIG2_64) | MASK_BINARY_OR2_64
		if C1 > 9999999999999999 { // non-canonical
			C1 = 0
		}
	} else { // if ((x & MASK_STEERING_BITS) != MASK_STEERING_BITS)
		exp = int((x&MASK_BINARY_EXPONENT1_64)>>53) - 398
		C1 = (x & MASK_BINARY_SIG1_64)
	}

	// if x is 0 or non-canonical return 0 preserving the sign bit and
	// the preferred exponent of MAX(Q(x), 0)
	if C1 == 0 {
		if exp < 0 {
			exp = 0
		}
		res = x_sign | ((uint64(exp) + 398) << 53)
		return res, pfpsf
	}
	// x is a finite non-zero number (not 0, non-canonical, or special)

	switch rndMode {
	case BID_ROUNDING_TO_NEAREST, BID_ROUNDING_TIES_AWAY:
		// return 0 if (exp <= -(p+1))
		if exp <= -17 {
			res = x_sign | 0x31c0000000000000
			pfpsf |= BID_INEXACT_EXCEPTION
			return res, pfpsf
		}
	case BID_ROUNDING_DOWN:
		// return 0 if (exp <= -p)
		if exp <= -16 {
			if x_sign != 0 {
				res = 0xb1c0000000000001
			} else {
				res = 0x31c0000000000000
			}
			pfpsf |= BID_INEXACT_EXCEPTION
			return res, pfpsf
		}
	case BID_ROUNDING_UP:
		// return 0 if (exp <= -p)
		if exp <= -16 {
			if x_sign != 0 {
				res = 0xb1c0000000000000
			} else {
				res = 0x31c0000000000001
			}
			pfpsf |= BID_INEXACT_EXCEPTION
			return res, pfpsf
		}
	case BID_ROUNDING_TO_ZERO:
		// return 0 if (exp <= -p)
		if exp <= -16 {
			res = x_sign | 0x31c0000000000000
			pfpsf |= BID_INEXACT_EXCEPTION
			return res, pfpsf
		}
	default:
		break
	}

	// q = nr. of decimal digits in x (1 <= q <= 54)
	//  determine first the nr. of bits in x
	if C1 >= 0x0020000000000000 { // x >= 2^53
		q = 16
	} else { // if x < 2^53
		tmp1 = math.Float64bits(float64(C1)) // exact conversion
		x_nr_bits = 1 + int(((tmp1>>52)&0x7ff)-0x3ff)
		q = bid_estimate_decimal_digits[x_nr_bits-1]
		if C1 >= bid_power10_table_128[q].w[0] {
			q++
		}
	}

	if exp >= 0 { // -exp <= 0
		// the argument is an integer already
		res = x
		return res, pfpsf
	}

	switch rndMode {
	case BID_ROUNDING_TO_NEAREST:
		if (q + exp) >= 0 { // exp < 0 and 1 <= -exp <= q
			// need to shift right -exp digits from the coefficient; exp will be 0
			ind = -exp // 1 <= ind <= 16; ind is a synonym for 'x'
			// chop off ind digits from the lower part of C1
			// C1 = C1 + 1/2 * 10^x where the result C1 fits in 64 bits
			// FOR ROUND_TO_NEAREST, WE ADD 1/2 ULP(y) then truncate
			C1 = C1 + bid_midpoint64[ind-1]
			// calculate C* and f*
			// C* is actually floor(C*) in this case
			// C* and f* need shifting and masking, as shown by
			// bid_shiftright128[] and bid_maskhigh128[]
			// 1 <= x <= 16
			// kx = 10^(-x) = bid_ten2mk64[ind - 1]
			// C* = (C1 + 1/2 * 10^x) * 10^(-x)
			// the approximation of 10^(-x) was rounded up to 64 bits
			P128 = __mul_64x64_to_128(C1, bid_ten2mk64_round64[ind-1])

			if (ind - 1) <= 2 { // 0 <= ind - 1 <= 2 => shift = 0
				res = P128.w[1]
				fstar.w[1] = 0
				fstar.w[0] = P128.w[0]
			} else if (ind - 1) <= 21 { // 3 <= ind - 1 <= 21 => 3 <= shift <= 63
				shift = bid_shiftright128_round64[ind-1] // 3 <= shift <= 63
				res = (P128.w[1] >> shift)
				fstar.w[1] = P128.w[1] & bid_maskhigh128_round64[ind-1]
				fstar.w[0] = P128.w[0]
			}
			// if (0 < f* < 10^(-x)) then the result is a midpoint
			// since round_to_even, subtract 1 if current result is odd
			if (res&0x0000000000000001) != 0 && (fstar.w[1] == 0) &&
				(fstar.w[0] < bid_ten2mk64_round64[ind-1]) {
				res--
			}
			// determine inexactness of the rounding of C*
			// if (0 < f* - 1/2 < 10^(-x)) then
			//   the result is exact
			// else // if (f* - 1/2 > T*) then
			//   the result is inexact
			if (ind - 1) <= 2 {
				if fstar.w[0] > 0x8000000000000000 {
					// f* > 1/2 and the result may be exact
					// fstar.w[0] - 0x8000000000000000ull is f* - 1/2
					if (fstar.w[0] - 0x8000000000000000) > bid_ten2mk64_round64[ind-1] {
						// set the inexact flag
						pfpsf |= BID_INEXACT_EXCEPTION
					} // else the result is exact
				} else { // the result is inexact; f2* <= 1/2
					// set the inexact flag
					pfpsf |= BID_INEXACT_EXCEPTION
				}
			} else { // if 3 <= ind - 1 <= 21
				if fstar.w[1] > bid_onehalf128_round64[ind-1] ||
					(fstar.w[1] == bid_onehalf128_round64[ind-1] && fstar.w[0] != 0) {
					// f2* > 1/2 and the result may be exact
					// Calculate f2* - 1/2
					if fstar.w[1] > bid_onehalf128_round64[ind-1] ||
						fstar.w[0] > bid_ten2mk64_round64[ind-1] {
						// set the inexact flag
						pfpsf |= BID_INEXACT_EXCEPTION
					} // else the result is exact
				} else { // the result is inexact; f2* <= 1/2
					// set the inexact flag
					pfpsf |= BID_INEXACT_EXCEPTION
				}
			}
			// set exponent to zero as it was negative before.
			res = x_sign | 0x31c0000000000000 | res
			return res, pfpsf
		} else { // if exp < 0 and q + exp < 0
			// the result is +0 or -0
			res = x_sign | 0x31c0000000000000
			pfpsf |= BID_INEXACT_EXCEPTION
			return res, pfpsf
		}
	case BID_ROUNDING_TIES_AWAY:
		if (q + exp) >= 0 { // exp < 0 and 1 <= -exp <= q
			ind = -exp
			C1 = C1 + bid_midpoint64[ind-1]
			P128 = __mul_64x64_to_128(C1, bid_ten2mk64_round64[ind-1])

			if (ind - 1) <= 2 {
				res = P128.w[1]
				fstar.w[1] = 0
				fstar.w[0] = P128.w[0]
			} else if (ind - 1) <= 21 {
				shift = bid_shiftright128_round64[ind-1]
				res = (P128.w[1] >> shift)
				fstar.w[1] = P128.w[1] & bid_maskhigh128_round64[ind-1]
				fstar.w[0] = P128.w[0]
			}
			// midpoints are already rounded correctly
			// determine inexactness of the rounding of C*
			if (ind - 1) <= 2 {
				if fstar.w[0] > 0x8000000000000000 {
					if (fstar.w[0] - 0x8000000000000000) > bid_ten2mk64_round64[ind-1] {
						pfpsf |= BID_INEXACT_EXCEPTION
					}
				} else {
					pfpsf |= BID_INEXACT_EXCEPTION
				}
			} else {
				if fstar.w[1] > bid_onehalf128_round64[ind-1] ||
					(fstar.w[1] == bid_onehalf128_round64[ind-1] && fstar.w[0] != 0) {
					if fstar.w[1] > bid_onehalf128_round64[ind-1] ||
						fstar.w[0] > bid_ten2mk64_round64[ind-1] {
						pfpsf |= BID_INEXACT_EXCEPTION
					}
				} else {
					pfpsf |= BID_INEXACT_EXCEPTION
				}
			}
			res = x_sign | 0x31c0000000000000 | res
			return res, pfpsf
		} else {
			res = x_sign | 0x31c0000000000000
			pfpsf |= BID_INEXACT_EXCEPTION
			return res, pfpsf
		}
	case BID_ROUNDING_DOWN:
		if (q + exp) > 0 { // exp < 0 and 1 <= -exp < q
			ind = -exp
			P128 = __mul_64x64_to_128(C1, bid_ten2mk64_round64[ind-1])

			if (ind - 1) <= 2 {
				res = P128.w[1]
				fstar.w[1] = 0
				fstar.w[0] = P128.w[0]
			} else if (ind - 1) <= 21 {
				shift = bid_shiftright128_round64[ind-1]
				res = (P128.w[1] >> shift)
				fstar.w[1] = P128.w[1] & bid_maskhigh128_round64[ind-1]
				fstar.w[0] = P128.w[0]
			}
			if (fstar.w[1] != 0) || (fstar.w[0] >= bid_ten2mk64_round64[ind-1]) {
				if x_sign != 0 {
					res++
				}
				pfpsf |= BID_INEXACT_EXCEPTION
			}
			res = x_sign | 0x31c0000000000000 | res
			return res, pfpsf
		} else {
			if x_sign != 0 {
				res = 0xb1c0000000000001
			} else {
				res = 0x31c0000000000000
			}
			pfpsf |= BID_INEXACT_EXCEPTION
			return res, pfpsf
		}
	case BID_ROUNDING_UP:
		if (q + exp) > 0 { // exp < 0 and 1 <= -exp < q
			ind = -exp
			P128 = __mul_64x64_to_128(C1, bid_ten2mk64_round64[ind-1])

			if (ind - 1) <= 2 {
				res = P128.w[1]
				fstar.w[1] = 0
				fstar.w[0] = P128.w[0]
			} else if (ind - 1) <= 21 {
				shift = bid_shiftright128_round64[ind-1]
				res = (P128.w[1] >> shift)
				fstar.w[1] = P128.w[1] & bid_maskhigh128_round64[ind-1]
				fstar.w[0] = P128.w[0]
			}
			if (fstar.w[1] != 0) || (fstar.w[0] >= bid_ten2mk64_round64[ind-1]) {
				if x_sign == 0 {
					res++
				}
				pfpsf |= BID_INEXACT_EXCEPTION
			}
			res = x_sign | 0x31c0000000000000 | res
			return res, pfpsf
		} else {
			if x_sign != 0 {
				res = 0xb1c0000000000000
			} else {
				res = 0x31c0000000000001
			}
			pfpsf |= BID_INEXACT_EXCEPTION
			return res, pfpsf
		}
	case BID_ROUNDING_TO_ZERO:
		if (q + exp) >= 0 { // exp < 0 and 1 <= -exp <= q
			ind = -exp
			P128 = __mul_64x64_to_128(C1, bid_ten2mk64_round64[ind-1])

			if (ind - 1) <= 2 {
				res = P128.w[1]
				fstar.w[1] = 0
				fstar.w[0] = P128.w[0]
			} else if (ind - 1) <= 21 {
				shift = bid_shiftright128_round64[ind-1]
				res = (P128.w[1] >> shift)
				fstar.w[1] = P128.w[1] & bid_maskhigh128_round64[ind-1]
				fstar.w[0] = P128.w[0]
			}
			if (fstar.w[1] != 0) || (fstar.w[0] >= bid_ten2mk64_round64[ind-1]) {
				pfpsf |= BID_INEXACT_EXCEPTION
			}
			res = x_sign | 0x31c0000000000000 | res
			return res, pfpsf
		} else {
			res = x_sign | 0x31c0000000000000
			pfpsf |= BID_INEXACT_EXCEPTION
			return res, pfpsf
		}
	default:
		break
	}
	return res, pfpsf
}

// Bid64NearbyInt is ported mechanically from bid64_nearbyintd.c.
func Bid64NearbyInt(x uint64, rndMode int) (uint64, uint32) {
	var res uint64 = 0xbaddbaddbaddbadd
	var x_sign uint64
	var x_nr_bits int
	var q, ind, shift int
	var C1 uint64
	var fstar BID_UINT128
	var P128 BID_UINT128
	var pfpsf uint32
	var exp int
	var tmp1 uint64

	x_sign = x & MASK_SIGN64 // 0 for positive, MASK_SIGN for negative

	// check for NaNs and infinities
	if (x & MASK_NAN64) == MASK_NAN64 { // check for NaN
		if (x & 0x0003ffffffffffff) > 999999999999999 {
			x = x & 0xfe00000000000000 // clear G6-G12 and the payload bits
		} else {
			x = x & 0xfe03ffffffffffff // clear G6-G12
		}
		if (x & MASK_SNAN64) == MASK_SNAN64 { // SNaN
			// set invalid flag
			pfpsf |= BID_INVALID_EXCEPTION
			// return quiet (SNaN)
			res = x & 0xfdffffffffffffff
		} else { // QNaN
			res = x
		}
		return res, pfpsf
	} else if (x & MASK_INF64) == MASK_INF64 { // check for Infinity
		res = x_sign | 0x7800000000000000
		return res, pfpsf
	}
	// unpack x
	if (x & MASK_STEERING_BITS64) == MASK_STEERING_BITS64 {
		exp = int((x&MASK_BINARY_EXPONENT2_64)>>51) - 398
		C1 = (x & MASK_BINARY_SIG2_64) | MASK_BINARY_OR2_64
		if C1 > 9999999999999999 {
			C1 = 0
		}
	} else {
		exp = int((x&MASK_BINARY_EXPONENT1_64)>>53) - 398
		C1 = (x & MASK_BINARY_SIG1_64)
	}

	if C1 == 0 {
		if exp < 0 {
			exp = 0
		}
		res = x_sign | ((uint64(exp) + 398) << 53)
		return res, pfpsf
	}

	switch rndMode {
	case BID_ROUNDING_TO_NEAREST, BID_ROUNDING_TIES_AWAY:
		if exp <= -17 {
			res = x_sign | 0x31c0000000000000
			return res, pfpsf
		}
	case BID_ROUNDING_DOWN:
		if exp <= -16 {
			if x_sign != 0 {
				res = 0xb1c0000000000001
			} else {
				res = 0x31c0000000000000
			}
			return res, pfpsf
		}
	case BID_ROUNDING_UP:
		if exp <= -16 {
			if x_sign != 0 {
				res = 0xb1c0000000000000
			} else {
				res = 0x31c0000000000001
			}
			return res, pfpsf
		}
	case BID_ROUNDING_TO_ZERO:
		if exp <= -16 {
			res = x_sign | 0x31c0000000000000
			return res, pfpsf
		}
	default:
		break
	}

	if C1 >= 0x0020000000000000 {
		q = 16
	} else {
		tmp1 = math.Float64bits(float64(C1))
		x_nr_bits = 1 + int(((tmp1>>52)&0x7ff)-0x3ff)
		q = bid_estimate_decimal_digits[x_nr_bits-1]
		if C1 >= bid_power10_table_128[q].w[0] {
			q++
		}
	}

	if exp >= 0 {
		res = x
		return res, pfpsf
	}

	switch rndMode {
	case BID_ROUNDING_TO_NEAREST:
		if (q + exp) >= 0 {
			ind = -exp
			C1 = C1 + bid_midpoint64[ind-1]
			P128 = __mul_64x64_to_128(C1, bid_ten2mk64_round64[ind-1])

			if (ind - 1) <= 2 {
				res = P128.w[1]
				fstar.w[1] = 0
				fstar.w[0] = P128.w[0]
			} else if (ind - 1) <= 21 {
				shift = bid_shiftright128_round64[ind-1]
				res = (P128.w[1] >> shift)
				fstar.w[1] = P128.w[1] & bid_maskhigh128_round64[ind-1]
				fstar.w[0] = P128.w[0]
			}
			if (res&0x0000000000000001) != 0 && (fstar.w[1] == 0) &&
				(fstar.w[0] < bid_ten2mk64_round64[ind-1]) {
				res--
			}
			res = x_sign | 0x31c0000000000000 | res
			return res, pfpsf
		} else {
			res = x_sign | 0x31c0000000000000
			return res, pfpsf
		}
	case BID_ROUNDING_TIES_AWAY:
		if (q + exp) >= 0 {
			ind = -exp
			C1 = C1 + bid_midpoint64[ind-1]
			P128 = __mul_64x64_to_128(C1, bid_ten2mk64_round64[ind-1])

			if (ind - 1) <= 2 {
				res = P128.w[1]
			} else if (ind - 1) <= 21 {
				shift = bid_shiftright128_round64[ind-1]
				res = (P128.w[1] >> shift)
			}
			res = x_sign | 0x31c0000000000000 | res
			return res, pfpsf
		} else {
			res = x_sign | 0x31c0000000000000
			return res, pfpsf
		}
	case BID_ROUNDING_DOWN:
		if (q + exp) > 0 {
			ind = -exp
			P128 = __mul_64x64_to_128(C1, bid_ten2mk64_round64[ind-1])

			if (ind - 1) <= 2 {
				res = P128.w[1]
				fstar.w[1] = 0
				fstar.w[0] = P128.w[0]
			} else if (ind - 1) <= 21 {
				shift = bid_shiftright128_round64[ind-1]
				res = (P128.w[1] >> shift)
				fstar.w[1] = P128.w[1] & bid_maskhigh128_round64[ind-1]
				fstar.w[0] = P128.w[0]
			}
			if (fstar.w[1] != 0) || (fstar.w[0] >= bid_ten2mk64_round64[ind-1]) {
				if x_sign != 0 {
					res++
				}
			}
			res = x_sign | 0x31c0000000000000 | res
			return res, pfpsf
		} else {
			if x_sign != 0 {
				res = 0xb1c0000000000001
			} else {
				res = 0x31c0000000000000
			}
			return res, pfpsf
		}
	case BID_ROUNDING_UP:
		if (q + exp) > 0 {
			ind = -exp
			P128 = __mul_64x64_to_128(C1, bid_ten2mk64_round64[ind-1])

			if (ind - 1) <= 2 {
				res = P128.w[1]
				fstar.w[1] = 0
				fstar.w[0] = P128.w[0]
			} else if (ind - 1) <= 21 {
				shift = bid_shiftright128_round64[ind-1]
				res = (P128.w[1] >> shift)
				fstar.w[1] = P128.w[1] & bid_maskhigh128_round64[ind-1]
				fstar.w[0] = P128.w[0]
			}
			if (fstar.w[1] != 0) || (fstar.w[0] >= bid_ten2mk64_round64[ind-1]) {
				if x_sign == 0 {
					res++
				}
			}
			res = x_sign | 0x31c0000000000000 | res
			return res, pfpsf
		} else {
			if x_sign != 0 {
				res = 0xb1c0000000000000
			} else {
				res = 0x31c0000000000001
			}
			return res, pfpsf
		}
	case BID_ROUNDING_TO_ZERO:
		if (q + exp) >= 0 {
			ind = -exp
			P128 = __mul_64x64_to_128(C1, bid_ten2mk64_round64[ind-1])

			if (ind - 1) <= 2 {
				res = P128.w[1]
			} else if (ind - 1) <= 21 {
				shift = bid_shiftright128_round64[ind-1]
				res = (P128.w[1] >> shift)
			}
			res = x_sign | 0x31c0000000000000 | res
			return res, pfpsf
		} else {
			res = x_sign | 0x31c0000000000000
			return res, pfpsf
		}
	default:
		break
	}
	return res, pfpsf
}

// Bid64RoundIntegralNearestEven is ported mechanically from bid64_round_integral.c.
func Bid64RoundIntegralNearestEven(x uint64) (uint64, uint32) {
	var res uint64 = 0xbaddbaddbaddbadd
	var x_sign uint64
	var x_nr_bits int
	var q, ind, shift int
	var C1 uint64
	var fstar BID_UINT128
	var P128 BID_UINT128
	var pfpsf uint32
	var exp int
	var tmp1 uint64

	x_sign = x & MASK_SIGN64 // 0 for positive, MASK_SIGN for negative

	// check for NaNs and infinities
	if (x & MASK_NAN64) == MASK_NAN64 { // check for NaN
		if (x & 0x0003ffffffffffff) > 999999999999999 {
			x = x & 0xfe00000000000000 // clear G6-G12 and the payload bits
		} else {
			x = x & 0xfe03ffffffffffff // clear G6-G12
		}
		if (x & MASK_SNAN64) == MASK_SNAN64 { // SNaN
			pfpsf |= BID_INVALID_EXCEPTION
			res = x & 0xfdffffffffffffff
		} else {
			res = x
		}
		return res, pfpsf
	} else if (x & MASK_INF64) == MASK_INF64 {
		res = x_sign | 0x7800000000000000
		return res, pfpsf
	}
	if (x & MASK_STEERING_BITS64) == MASK_STEERING_BITS64 {
		exp = int((x&MASK_BINARY_EXPONENT2_64)>>51) - 398
		C1 = (x & MASK_BINARY_SIG2_64) | MASK_BINARY_OR2_64
		if C1 > 9999999999999999 {
			C1 = 0
		}
	} else {
		exp = int((x&MASK_BINARY_EXPONENT1_64)>>53) - 398
		C1 = (x & MASK_BINARY_SIG1_64)
	}

	if C1 == 0 {
		if exp < 0 {
			exp = 0
		}
		res = x_sign | ((uint64(exp) + 398) << 53)
		return res, pfpsf
	}

	if exp <= -17 {
		res = x_sign | 0x31c0000000000000
		return res, pfpsf
	}
	if C1 >= 0x0020000000000000 {
		q = 16
	} else {
		tmp1 = math.Float64bits(float64(C1))
		x_nr_bits = 1 + int(((tmp1>>52)&0x7ff)-0x3ff)
		q = bid_estimate_decimal_digits[x_nr_bits-1]
		if C1 >= bid_power10_table_128[q].w[0] {
			q++
		}
	}

	if exp >= 0 {
		res = x
		return res, pfpsf
	} else if (q + exp) >= 0 {
		ind = -exp
		C1 = C1 + bid_midpoint64[ind-1]
		P128 = __mul_64x64_to_128(C1, bid_ten2mk64_round64[ind-1])

		if (ind - 1) <= 2 {
			res = P128.w[1]
			fstar.w[1] = 0
			fstar.w[0] = P128.w[0]
		} else if (ind - 1) <= 21 {
			shift = bid_shiftright128_round64[ind-1]
			res = (P128.w[1] >> shift)
			fstar.w[1] = P128.w[1] & bid_maskhigh128_round64[ind-1]
			fstar.w[0] = P128.w[0]
		}
		if (res&0x0000000000000001) != 0 && (fstar.w[1] == 0) &&
			(fstar.w[0] < bid_ten2mk64_round64[ind-1]) {
			res--
		}
		res = x_sign | 0x31c0000000000000 | res
		return res, pfpsf
	} else {
		res = x_sign | 0x31c0000000000000
		return res, pfpsf
	}
}

// Bid64RoundIntegralNegative is ported mechanically from bid64_round_integral.c.
func Bid64RoundIntegralNegative(x uint64) (uint64, uint32) {
	var res uint64 = 0xbaddbaddbaddbadd
	var x_sign uint64
	var x_nr_bits int
	var q, ind, shift int
	var C1 uint64
	var fstar BID_UINT128
	var P128 BID_UINT128
	var pfpsf uint32
	var exp int
	var tmp1 uint64

	x_sign = x & MASK_SIGN64

	if (x & MASK_NAN64) == MASK_NAN64 {
		if (x & 0x0003ffffffffffff) > 999999999999999 {
			x = x & 0xfe00000000000000
		} else {
			x = x & 0xfe03ffffffffffff
		}
		if (x & MASK_SNAN64) == MASK_SNAN64 {
			pfpsf |= BID_INVALID_EXCEPTION
			res = x & 0xfdffffffffffffff
		} else {
			res = x
		}
		return res, pfpsf
	} else if (x & MASK_INF64) == MASK_INF64 {
		res = x_sign | 0x7800000000000000
		return res, pfpsf
	}
	if (x & MASK_STEERING_BITS64) == MASK_STEERING_BITS64 {
		exp = int((x&MASK_BINARY_EXPONENT2_64)>>51) - 398
		C1 = (x & MASK_BINARY_SIG2_64) | MASK_BINARY_OR2_64
		if C1 > 9999999999999999 {
			C1 = 0
		}
	} else {
		exp = int((x&MASK_BINARY_EXPONENT1_64)>>53) - 398
		C1 = (x & MASK_BINARY_SIG1_64)
	}

	if C1 == 0 {
		if exp < 0 {
			exp = 0
		}
		res = x_sign | ((uint64(exp) + 398) << 53)
		return res, pfpsf
	}

	if exp <= -16 {
		if x_sign != 0 {
			res = 0xb1c0000000000001
		} else {
			res = 0x31c0000000000000
		}
		return res, pfpsf
	}
	if C1 >= 0x0020000000000000 {
		q = 16
	} else {
		tmp1 = math.Float64bits(float64(C1))
		x_nr_bits = 1 + int(((tmp1>>52)&0x7ff)-0x3ff)
		q = bid_estimate_decimal_digits[x_nr_bits-1]
		if C1 >= bid_power10_table_128[q].w[0] {
			q++
		}
	}

	if exp >= 0 {
		res = x
		return res, pfpsf
	} else if (q + exp) > 0 {
		ind = -exp
		P128 = __mul_64x64_to_128(C1, bid_ten2mk64_round64[ind-1])

		if (ind - 1) <= 2 {
			res = P128.w[1]
			fstar.w[1] = 0
			fstar.w[0] = P128.w[0]
		} else if (ind - 1) <= 21 {
			shift = bid_shiftright128_round64[ind-1]
			res = (P128.w[1] >> shift)
			fstar.w[1] = P128.w[1] & bid_maskhigh128_round64[ind-1]
			fstar.w[0] = P128.w[0]
		}
		if x_sign != 0 && ((fstar.w[1] != 0) || (fstar.w[0] >= bid_ten2mk64_round64[ind-1])) {
			res++
		}
		res = x_sign | 0x31c0000000000000 | res
		return res, pfpsf
	} else {
		if x_sign != 0 {
			res = 0xb1c0000000000001
		} else {
			res = 0x31c0000000000000
		}
		return res, pfpsf
	}
}

// Bid64RoundIntegralPositive is ported mechanically from bid64_round_integral.c.
func Bid64RoundIntegralPositive(x uint64) (uint64, uint32) {
	var res uint64 = 0xbaddbaddbaddbadd
	var x_sign uint64
	var x_nr_bits int
	var q, ind, shift int
	var C1 uint64
	var fstar BID_UINT128
	var P128 BID_UINT128
	var pfpsf uint32
	var exp int
	var tmp1 uint64

	x_sign = x & MASK_SIGN64

	if (x & MASK_NAN64) == MASK_NAN64 {
		if (x & 0x0003ffffffffffff) > 999999999999999 {
			x = x & 0xfe00000000000000
		} else {
			x = x & 0xfe03ffffffffffff
		}
		if (x & MASK_SNAN64) == MASK_SNAN64 {
			pfpsf |= BID_INVALID_EXCEPTION
			res = x & 0xfdffffffffffffff
		} else {
			res = x
		}
		return res, pfpsf
	} else if (x & MASK_INF64) == MASK_INF64 {
		res = x_sign | 0x7800000000000000
		return res, pfpsf
	}
	if (x & MASK_STEERING_BITS64) == MASK_STEERING_BITS64 {
		exp = int((x&MASK_BINARY_EXPONENT2_64)>>51) - 398
		C1 = (x & MASK_BINARY_SIG2_64) | MASK_BINARY_OR2_64
		if C1 > 9999999999999999 {
			C1 = 0
		}
	} else {
		exp = int((x&MASK_BINARY_EXPONENT1_64)>>53) - 398
		C1 = (x & MASK_BINARY_SIG1_64)
	}

	if C1 == 0 {
		if exp < 0 {
			exp = 0
		}
		res = x_sign | ((uint64(exp) + 398) << 53)
		return res, pfpsf
	}

	if exp <= -16 {
		if x_sign != 0 {
			res = 0xb1c0000000000000
		} else {
			res = 0x31c0000000000001
		}
		return res, pfpsf
	}
	if C1 >= 0x0020000000000000 {
		q = 16
	} else {
		tmp1 = math.Float64bits(float64(C1))
		x_nr_bits = 1 + int(((tmp1>>52)&0x7ff)-0x3ff)
		q = bid_estimate_decimal_digits[x_nr_bits-1]
		if C1 >= bid_power10_table_128[q].w[0] {
			q++
		}
	}

	if exp >= 0 {
		res = x
		return res, pfpsf
	} else if (q + exp) > 0 {
		ind = -exp
		P128 = __mul_64x64_to_128(C1, bid_ten2mk64_round64[ind-1])

		if (ind - 1) <= 2 {
			res = P128.w[1]
			fstar.w[1] = 0
			fstar.w[0] = P128.w[0]
		} else if (ind - 1) <= 21 {
			shift = bid_shiftright128_round64[ind-1]
			res = (P128.w[1] >> shift)
			fstar.w[1] = P128.w[1] & bid_maskhigh128_round64[ind-1]
			fstar.w[0] = P128.w[0]
		}
		if x_sign == 0 && ((fstar.w[1] != 0) || (fstar.w[0] >= bid_ten2mk64_round64[ind-1])) {
			res++
		}
		res = x_sign | 0x31c0000000000000 | res
		return res, pfpsf
	} else {
		if x_sign != 0 {
			res = 0xb1c0000000000000
		} else {
			res = 0x31c0000000000001
		}
		return res, pfpsf
	}
}

// Bid64RoundIntegralZero is ported mechanically from bid64_round_integral.c: bid64_round_integral_zero.
func Bid64RoundIntegralZero(x uint64) (uint64, uint32) {
	var res uint64 = 0xbaddbaddbaddbadd
	var x_sign uint64
	var x_nr_bits int
	var q, ind, shift int
	var C1 uint64
	var P128 BID_UINT128
	var pfpsf uint32
	var exp int
	var tmp1 uint64

	x_sign = x & MASK_SIGN64

	if (x & MASK_NAN64) == MASK_NAN64 {
		if (x & 0x0003ffffffffffff) > 999999999999999 {
			x = x & 0xfe00000000000000
		} else {
			x = x & 0xfe03ffffffffffff
		}
		if (x & MASK_SNAN64) == MASK_SNAN64 {
			pfpsf |= BID_INVALID_EXCEPTION
			res = x & 0xfdffffffffffffff
		} else {
			res = x
		}
		return res, pfpsf
	} else if (x & MASK_INF64) == MASK_INF64 {
		res = x_sign | 0x7800000000000000
		return res, pfpsf
	}
	if (x & MASK_STEERING_BITS64) == MASK_STEERING_BITS64 {
		exp = int((x&MASK_BINARY_EXPONENT2_64)>>51) - 398
		C1 = (x & MASK_BINARY_SIG2_64) | MASK_BINARY_OR2_64
		if C1 > 9999999999999999 {
			C1 = 0
		}
	} else {
		exp = int((x&MASK_BINARY_EXPONENT1_64)>>53) - 398
		C1 = (x & MASK_BINARY_SIG1_64)
	}

	if C1 == 0 {
		if exp < 0 {
			exp = 0
		}
		res = x_sign | ((uint64(exp) + 398) << 53)
		return res, pfpsf
	}

	if exp <= -16 {
		res = x_sign | 0x31c0000000000000
		return res, pfpsf
	}
	if C1 >= 0x0020000000000000 {
		q = 16
	} else {
		tmp1 = math.Float64bits(float64(C1))
		x_nr_bits = 1 + int(((tmp1>>52)&0x7ff)-0x3ff)
		q = bid_estimate_decimal_digits[x_nr_bits-1]
		if C1 >= bid_power10_table_128[q].w[0] {
			q++
		}
	}

	if exp >= 0 {
		res = x
		return res, pfpsf
	} else if (q + exp) >= 0 {
		ind = -exp
		P128 = __mul_64x64_to_128(C1, bid_ten2mk64_round64[ind-1])

		if (ind - 1) <= 2 {
			res = P128.w[1]
		} else if (ind - 1) <= 21 {
			shift = bid_shiftright128_round64[ind-1]
			res = (P128.w[1] >> shift)
		}
		res = x_sign | 0x31c0000000000000 | res
		return res, pfpsf
	} else {
		res = x_sign | 0x31c0000000000000
		return res, pfpsf
	}
}

// Bid64RoundIntegralNearestAway is ported mechanically from bid64_round_integral.c.
func Bid64RoundIntegralNearestAway(x uint64) (uint64, uint32) {
	var res uint64 = 0xbaddbaddbaddbadd
	var x_sign uint64
	var x_nr_bits int
	var q, ind, shift int
	var C1 uint64
	var P128 BID_UINT128
	var pfpsf uint32
	var exp int
	var tmp1 uint64

	x_sign = x & MASK_SIGN64

	if (x & MASK_NAN64) == MASK_NAN64 {
		if (x & 0x0003ffffffffffff) > 999999999999999 {
			x = x & 0xfe00000000000000
		} else {
			x = x & 0xfe03ffffffffffff
		}
		if (x & MASK_SNAN64) == MASK_SNAN64 {
			pfpsf |= BID_INVALID_EXCEPTION
			res = x & 0xfdffffffffffffff
		} else {
			res = x
		}
		return res, pfpsf
	} else if (x & MASK_INF64) == MASK_INF64 {
		res = x_sign | 0x7800000000000000
		return res, pfpsf
	}
	if (x & MASK_STEERING_BITS64) == MASK_STEERING_BITS64 {
		exp = int((x&MASK_BINARY_EXPONENT2_64)>>51) - 398
		C1 = (x & MASK_BINARY_SIG2_64) | MASK_BINARY_OR2_64
		if C1 > 9999999999999999 {
			C1 = 0
		}
	} else {
		exp = int((x&MASK_BINARY_EXPONENT1_64)>>53) - 398
		C1 = (x & MASK_BINARY_SIG1_64)
	}

	if C1 == 0 {
		if exp < 0 {
			exp = 0
		}
		res = x_sign | ((uint64(exp) + 398) << 53)
		return res, pfpsf
	}

	if exp <= -17 {
		res = x_sign | 0x31c0000000000000
		return res, pfpsf
	}
	if C1 >= 0x0020000000000000 {
		q = 16
	} else {
		tmp1 = math.Float64bits(float64(C1))
		x_nr_bits = 1 + int(((tmp1>>52)&0x7ff)-0x3ff)
		q = bid_estimate_decimal_digits[x_nr_bits-1]
		if C1 >= bid_power10_table_128[q].w[0] {
			q++
		}
	}

	if exp >= 0 {
		res = x
		return res, pfpsf
	} else if (q + exp) >= 0 {
		ind = -exp
		C1 = C1 + bid_midpoint64[ind-1]
		P128 = __mul_64x64_to_128(C1, bid_ten2mk64_round64[ind-1])

		if (ind - 1) <= 2 {
			res = P128.w[1]
		} else if (ind - 1) <= 21 {
			shift = bid_shiftright128_round64[ind-1]
			res = (P128.w[1] >> shift)
		}
		res = x_sign | 0x31c0000000000000 | res
		return res, pfpsf
	} else {
		res = x_sign | 0x31c0000000000000
		return res, pfpsf
	}
}
