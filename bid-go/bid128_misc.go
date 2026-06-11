// Ported from: Intel bid128_noncomp.c (class), bid128_fdimd.c, bid128_llquantexpd.c,
//              bid128_quantumd.c, bid128_scalb.c, bid128_scalbl.c,
//              bid128_logb.c, bid128_logbd.c
// Mechanical translation - all logic preserved exactly.

package bidgo

import "math"

// Bid128Class returns the class of x.
// Ported from bid128_class.
func Bid128Class(x BID_UINT128) int {
	var res int
	var sig_x_prime256 BID_UINT256
	var sig_x_prime192 BID_UINT192
	var sig_x BID_UINT128
	var exp_x int

	if (x.w[1] & NAN_MASK64) == NAN_MASK64 {
		if (x.w[1] & SNAN_MASK64) == SNAN_MASK64 {
			res = signalingNaN
		} else {
			res = quietNaN
		}
		return res
	}
	if (x.w[1] & INFINITY_MASK64) == INFINITY_MASK64 {
		if (x.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = negativeInfinity
		} else {
			res = positiveInfinity
		}
		return res
	}
	// decode number into exponent and significand
	sig_x.w[1] = x.w[1] & 0x0001ffffffffffff
	sig_x.w[0] = x.w[0]
	// check for zero or non-canonical
	if (sig_x.w[1] > 0x0001ed09bead87c0) ||
		((sig_x.w[1] == 0x0001ed09bead87c0) &&
			(sig_x.w[0] > 0x378d8e63ffffffff)) ||
		((x.w[1] & 0x6000000000000000) == 0x6000000000000000) ||
		((sig_x.w[1] == 0) && (sig_x.w[0] == 0)) {
		if (x.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = negativeZero
		} else {
			res = positiveZero
		}
		return res
	}
	exp_x = int((x.w[1] >> 49) & 0x000000000003fff)
	// if exponent is less than -6176, the number may be subnormal
	if exp_x < 33 { // sig_x * 10^exp_x
		if exp_x > 19 {
			sig_x_prime256 = __mul_128x128_to_256(sig_x, bid_ten2k128[exp_x-20])
			// 10^33 = 0x0000314dc6448d93_38c15b0a00000000
			if (sig_x_prime256.w[3] == 0) && (sig_x_prime256.w[2] == 0) &&
				((sig_x_prime256.w[1] < 0x0000314dc6448d93) ||
					((sig_x_prime256.w[1] == 0x0000314dc6448d93) &&
						(sig_x_prime256.w[0] < 0x38c15b0a00000000))) {
				if (x.w[1] & MASK_SIGN64) == MASK_SIGN64 {
					res = negativeSubnormal
				} else {
					res = positiveSubnormal
				}
				return res
			}
		} else {
			sig_x_prime192 = __mul_64x128_to_192(bid_ten2k64[exp_x], sig_x)
			// 10^33 = 0x0000314dc6448d93_38c15b0a00000000
			if (sig_x_prime192.w[2] == 0) &&
				((sig_x_prime192.w[1] < 0x0000314dc6448d93) ||
					((sig_x_prime192.w[1] == 0x0000314dc6448d93) &&
						(sig_x_prime192.w[0] < 0x38c15b0a00000000))) {
				if (x.w[1] & MASK_SIGN64) == MASK_SIGN64 {
					res = negativeSubnormal
				} else {
					res = positiveSubnormal
				}
				return res
			}
		}
	}
	// otherwise, normal number, determine the sign
	if (x.w[1] & MASK_SIGN64) == MASK_SIGN64 {
		res = negativeNormal
	} else {
		res = positiveNormal
	}
	return res
}

// Bid128Llquantexp returns the quantum exponent of x as int64.
// Ported from bid128_llquantexp.
func Bid128Llquantexp(x BID_UINT128, pfpsf *uint32) int64 {
	var res int64

	if (x.w[1] & MASK_SPECIAL128) == MASK_SPECIAL128 {
		// set invalid flag
		*pfpsf |= BID_INVALID_EXCEPTION
		res = -1 << 63 // int64 minimum
		return res
	}
	if (x.w[1] & MASK_STEERING_BITS64) == MASK_STEERING_BITS64 {
		res = int64((x.w[1]>>47)&0x3fff) - 6176
	} else {
		res = int64((x.w[1]>>49)&0x3fff) - 6176
	}
	return res
}

// Bid128Quantexp returns the quantum exponent of x as int32.
// Ported from bid128_quantexp.
func Bid128Quantexp(x BID_UINT128, pfpsf *uint32) int32 {
	if (x.w[1] & MASK_SPECIAL128) == MASK_SPECIAL128 {
		*pfpsf |= BID_INVALID_EXCEPTION
		return -0x80000000
	}
	if (x.w[1] & MASK_STEERING_BITS64) == MASK_STEERING_BITS64 {
		return int32((x.w[1]>>47)&0x3fff) - 6176
	}
	return int32((x.w[1]>>49)&0x3fff) - 6176
}

// Bid128Quantum returns 10^exponent(x).
// Ported from bid128_quantum.
func Bid128Quantum(x BID_UINT128) BID_UINT128 {
	var res BID_UINT128
	var int_exp int

	// If x is infinite, the result is +Inf. If x is NaN, the result is NaN
	if (x.w[1] & MASK_ANY_INF_128) == INFINITY_MASK64 {
		res.w[1] = 0x7800000000000000
		res.w[0] = 0x0000000000000000
		return res
	} else if (x.w[1] & NAN_MASK64) == NAN_MASK64 {
		res.w[1] = x.w[1] & QUIET_MASK64
		res.w[0] = x.w[0]
		return res
	}

	// Extract exponent
	if (x.w[1] & MASK_STEERING_BITS64) == MASK_STEERING_BITS64 {
		int_exp = int((x.w[1]>>47)&0x3fff) - 6176
	} else {
		int_exp = int((x.w[1]>>49)&0x3fff) - 6176
	}

	// Form 10^new_exponent*1
	res.w[1] = (uint64(int64(int_exp)) << 49) + 0x3040000000000000
	res.w[0] = 0x0000000000000001

	return res
}

// Bid128Scalbn returns x * 10^n.
// Ported from bid128_scalbn.
func Bid128Scalbn(x BID_UINT128, n int, rnd_mode int, pfpsf *uint32) BID_UINT128 {
	var CX, CX2, CBID_X8, res BID_UINT128
	var exp64 int64
	var sign_x uint64
	var exponent_x int

	// unpack arguments, check for NaN or Infinity
	sign_x, exponent_x, CX, valid := unpack_BID128_value(x)
	if !valid {
		// x is Inf. or NaN or 0
		if (x.w[1] & SNAN_MASK64) == SNAN_MASK64 { // y is sNaN
			*pfpsf |= BID_INVALID_EXCEPTION
		}
		res.w[1] = CX.w[1] & QUIET_MASK64
		res.w[0] = CX.w[0]
		if CX.w[1] == 0 && CX.w[0] == 0 {
			exp64 = int64(exponent_x) + int64(n)
			if exp64 < 0 {
				exp64 = 0
			}
			if exp64 > DECIMAL_MAX_EXPON_128 {
				exp64 = DECIMAL_MAX_EXPON_128
			}
			exponent_x = int(exp64)
			res = very_fast_get_BID128(sign_x, exponent_x, CX)
		}
		return res
	}

	exp64 = int64(exponent_x) + int64(n)
	exponent_x = int(exp64)

	if uint32(exponent_x) <= uint32(DECIMAL_MAX_EXPON_128) {
		res = very_fast_get_BID128(sign_x, exponent_x, CX)
		return res
	}
	// check for overflow
	if exp64 > int64(DECIMAL_MAX_EXPON_128) {
		if CX.w[1] < 0x314dc6448d93 {
			// try to normalize coefficient
			for {
				CBID_X8.w[1] = (CX.w[1] << 3) | (CX.w[0] >> 61)
				CBID_X8.w[0] = CX.w[0] << 3
				CX2.w[1] = (CX.w[1] << 1) | (CX.w[0] >> 63)
				CX2.w[0] = CX.w[0] << 1
				CX = __add_128_128(CX2, CBID_X8)

				exponent_x--
				exp64--
				if !(CX.w[1] < 0x314dc6448d93 && exp64 > int64(DECIMAL_MAX_EXPON_128)) {
					break
				}
			}
		}
		if exp64 <= int64(DECIMAL_MAX_EXPON_128) {
			res = very_fast_get_BID128(sign_x, exponent_x, CX)
			return res
		} else {
			exponent_x = 0x7fffffff // overflow
		}
	}
	// exponent < 0
	// the BID pack routine will round the coefficient
	res = bid_get_BID128(sign_x, exponent_x, CX, rnd_mode, pfpsf)
	return res
}

// Bid128Scalbln returns x * 10^n (n is long int).
// Ported from bid128_scalbln.
func Bid128Scalbln(x BID_UINT128, n int64, rnd_mode int, pfpsf *uint32) BID_UINT128 {
	n1 := int32(n)
	n1 = func() int32 {
		if int64(n1) < n {
			return int32(0x7fffffff)
		}
		if int64(n1) > n {
			return int32(-0x80000000)
		}
		return n1
	}()
	return Bid128Scalbn(x, int(n1), rnd_mode, pfpsf)
}

// Bid128Ilogb returns the unbiased exponent of x as int.
// Ported from bid128_ilogb.
func Bid128Ilogb(x BID_UINT128, pfpsf *uint32) int {
	var CX BID_UINT128
	var D int64
	var exponent_x, bin_expon_cx, digits, res int

	_, exponent_x, CX, valid := unpack_BID128_value(x)
	if !valid {
		*pfpsf |= BID_INVALID_EXCEPTION
		if (x.w[1] & 0x7c00000000000000) == 0x7800000000000000 {
			res = 0x7fffffff
		} else {
			res = int(int32(-1 << 31)) // INT_MIN
		}
		return res
	}
	// find number of digits in coefficient
	// 2^64
	f64_i := uint32(0x5f800000)
	// fx ~ CX
	fx_d := noFmaMulAddF32(float32(CX.w[1]), math.Float32frombits(f64_i), float32(CX.w[0]))
	fx_i := math.Float32bits(fx_d)
	bin_expon_cx = int((fx_i>>23)&0xff) - 0x7f
	digits = bid_estimate_decimal_digits[bin_expon_cx]
	// scale = 38-estimate_decimal_digits[bin_expon_cx];
	D = int64(CX.w[1]) - int64(bid_power10_index_binexp_128[bin_expon_cx].w[1])
	if D > 0 || (D == 0 && CX.w[0] >= bid_power10_index_binexp_128[bin_expon_cx].w[0]) {
		digits++
	}

	exponent_x = exponent_x - EXPONENT_BIAS128 - 1 + digits

	return exponent_x
}

// Bid128Logb returns the exponent of x as a BID128 value.
// Ported from bid128_logb.
func Bid128Logb(x BID_UINT128, pfpsf *uint32) BID_UINT128 {
	var ires int
	var res, CX BID_UINT128

	_, _, CX, valid := unpack_BID128_value(x)
	if !valid {
		// test if x is NaN/Inf
		if (x.w[1] & 0x7800000000000000) == 0x7800000000000000 {
			if (x.w[1] & 0x7e00000000000000) == 0x7e00000000000000 { // sNaN
				*pfpsf |= BID_INVALID_EXCEPTION
			}
			res.w[1] = CX.w[1] & QUIET_MASK64
			res.w[0] = CX.w[0]
			if (x.w[1] & 0x7c00000000000000) == 0x7800000000000000 {
				res.w[1] = res.w[1] & 0x7fffffffffffffff
			}
			return res
		}

		// x is 0
		*pfpsf |= BID_ZERO_DIVIDE_EXCEPTION
		res.w[1] = 0xf800000000000000
		res.w[0] = 0
		return res
	}

	ires = Bid128Ilogb(x, pfpsf)
	if ires&0x80000000 != 0 {
		res.w[1] = 0xb040000000000000
		res.w[0] = uint64(-int64(int32(ires)))
	} else {
		res.w[1] = 0x3040000000000000
		res.w[0] = uint64(ires)
	}
	return res
}

// Bid128Fdim returns x - y if x > y, and +0 if x <= y.
// Ported from bid128_fdim.
func Bid128Fdim(x, y BID_UINT128, rnd_mode int, pfpsf *uint32) BID_UINT128 {
	var res BID_UINT128
	var tmp_fpsf uint32

	tmp_fpsf = *pfpsf // save fpsf
	cmpres, _ := Bid128QuietGreater(x, y)
	*pfpsf = tmp_fpsf // restore fpsf
	if ((x.w[1] & NAN_MASK64) != NAN_MASK64) && ((y.w[1] & NAN_MASK64) != NAN_MASK64) &&
		cmpres == 0 { // if x != NaN and y != NaN and x <= y return +0
		res.w[1] = 0x3040000000000000
		res.w[0] = 0x0000000000000000
		return res
	}

	// else if x = NaN or y = NaN or x > y return x - y
	res = Bid128Sub(x, y, rnd_mode, pfpsf)
	return res
}
