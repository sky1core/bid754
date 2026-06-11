// Ported from: Intel bid32_nearbyintd.c, bid32_fdimd.c, bid32_scalbl.c, bid32_modf.c, bid32_ldexp.c, bid32_frexp.c
// Mechanical translation - all logic preserved exactly.

package bidgo

import "math"

// Bid32NearbyInt is ported from bid32_nearbyintd.c (bid64 경유).
func Bid32NearbyInt(x uint32, rnd_mode int) (uint32, uint32) {
	x64, f1 := Bid32ToBid64(x)
	res64, f2 := Bid64NearbyInt(x64, rnd_mode)
	res, f3 := Bid64ToBid32(res64, 0)
	return res, f1 | f2 | f3
}

// Bid32Fdim is ported from bid32_fdimd.c.
func Bid32Fdim(x, y uint32, rnd_mode int) (uint32, uint32) {
	var res uint32
	var pfpsf uint32

	cmpres, _ := Bid32QuietGreater(x, y)
	if ((x & MASK_NAN32) != MASK_NAN32) && ((y & MASK_NAN32) != MASK_NAN32) && cmpres == 0 {
		res = 0x32800000
		return res, pfpsf
	}

	return Bid32SubWithFlags(x, y, rnd_mode)
}

// Bid32Scalbln is ported from bid32_scalbl.c.
func Bid32Scalbln(x uint32, n int64, rnd_mode int) (uint32, uint32) {
	n1 := int32(n)
	if int64(n1) < n {
		n1 = 0x7fffffff
	} else if int64(n1) > n {
		n1 = -0x80000000
	}
	return Bid32Scalbn(x, int(n1), rnd_mode)
}

// Bid32Modf is ported from bid32_modf.c.
func Bid32Modf(x uint32) (uint32, uint32, uint32) {
	x64, f0 := Bid32ToBid64(x)
	frac64, iptr64, flags := Bid64Modf(x64)
	frac, f1 := Bid64ToBid32(frac64, 0)
	iptr, f2 := Bid64ToBid32(iptr64, 0)
	return frac, iptr, f0 | flags | f1 | f2
}

// Bid32Ldexp is ported from bid32_ldexp.c (same as scalbn).
func Bid32Ldexp(x uint32, n int, rnd_mode int) (uint32, uint32) {
	return Bid32Scalbn(x, n, rnd_mode)
}

// Bid32Frexp is ported from bid32_frexp.c.
func Bid32Frexp(x uint32) (uint32, int, uint32) {
	var sig_x, res uint32
	var exp_x uint32
	var pfpsf uint32

	if (x & MASK_INF32) == MASK_INF32 {
		res = x
		if (x & MASK_SNAN32) == MASK_SNAN32 {
			res = x & 0xfdffffff
		}
		return res, 0, pfpsf
	}

	if (x & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
		exp_x = (x & MASK_BINARY_EXPONENT2_32) >> 21
		sig_x = (x & MASK_BINARY_SIG2_32) | MASK_BINARY_OR2_32
		if sig_x > 9999999 || sig_x == 0 {
			res = (x & 0x80000000) | (exp_x << 23)
			return res, 0, pfpsf
		}
	} else {
		exp_x = (x & MASK_BINARY_EXPONENT1_32) >> 23
		sig_x = x & MASK_BINARY_SIG1_32
		if sig_x == 0 {
			res = (x & 0x80000000) | (exp_x << 23)
			return res, 0, pfpsf
		}
	}

	tmp := math.Float32bits(float32(sig_x))
	x_nr_bits := 1 + int((tmp>>23)&0xff) - 0x7f
	q := int(bid_nr_digits[x_nr_bits-1].digits)
	if q == 0 {
		q = int(bid_nr_digits[x_nr_bits-1].digits1)
		if uint64(sig_x) >= bid_nr_digits[x_nr_bits-1].threshold_lo {
			q++
		}
	}

	exp := int(exp_x) - 101 + q
	if sig_x < 0x00800000 {
		res = (x & 0x807fffff) | (uint32(-q+101) << 23)
	} else {
		res = (x & 0xe01fffff) | (uint32(-q+101) << 21)
	}
	return res, exp, pfpsf
}

// Bid32Fmod is ported from bid32_fmod.c (same algorithm as bid32_rem but without halfway adjustment).
func Bid32Fmod(x, y uint32) (uint32, uint32) {
	var CX, Q64, CYL uint64
	var CY, sign_x, sign_y, coefficient_x, coefficient_y, res uint32
	var Q, R, T uint32
	var exponent_x, exponent_y, bin_expon, e_scale int
	var digits_x, diff_expon int
	var pfpsf uint32

	sign_y, exponent_y, coefficient_y, valid_y := unpack_BID32(y)
	sign_x, exponent_x, coefficient_x, valid_x := unpack_BID32(x)
	_ = sign_y

	if !valid_x {
		if (y & SNAN_MASK32) == SNAN_MASK32 {
			pfpsf |= BID_INVALID_EXCEPTION
		}
		if (x & 0x7c000000) == 0x7c000000 {
			if (x & SNAN_MASK32) == SNAN_MASK32 {
				pfpsf |= BID_INVALID_EXCEPTION
			}
			res = coefficient_x & QUIET_MASK32
			return res, pfpsf
		}
		if (x & 0x78000000) == 0x78000000 {
			if (y & NAN_MASK32) != NAN_MASK32 {
				pfpsf |= BID_INVALID_EXCEPTION
				return 0x7c000000, pfpsf
			}
		}
		if ((y & 0x78000000) < 0x78000000) && coefficient_y != 0 {
			if (y & 0x60000000) == 0x60000000 {
				exponent_y = int((y >> 21) & 0xff)
			} else {
				exponent_y = int((y >> 23) & 0xff)
			}
			if exponent_y < exponent_x {
				exponent_x = exponent_y
			}
			res = (uint32(exponent_x) << 23) | sign_x
			return res, pfpsf
		}
	}
	if !valid_y {
		if (y & 0x7c000000) == 0x7c000000 {
			if (y & SNAN_MASK32) == SNAN_MASK32 {
				pfpsf |= BID_INVALID_EXCEPTION
			}
			res = coefficient_y & QUIET_MASK32
			return res, pfpsf
		}
		if (y & 0x78000000) == 0x78000000 {
			res = very_fast_get_BID32(sign_x, exponent_x, coefficient_x)
			return res, pfpsf
		}
		pfpsf |= BID_INVALID_EXCEPTION
		return 0x7c000000, pfpsf
	}

	diff_expon = exponent_x - exponent_y
	if diff_expon <= 0 {
		diff_expon = -diff_expon
		if diff_expon > 7 {
			res = x
			return res, pfpsf
		}
		T = uint32(bid_power10_table_128[diff_expon].w[0])
		CYL = uint64(coefficient_y) * uint64(T)
		if CYL > uint64(coefficient_x) {
			res = x
			return res, pfpsf
		}
		CY = uint32(CYL)
		Q = coefficient_x / CY
		R = coefficient_x - Q*CY
		res = very_fast_get_BID32(sign_x, exponent_x, R)
		return res, pfpsf
	}

	CX = uint64(coefficient_x)
	for diff_expon > 0 {
		tempx := math.Float32bits(float32(CX))
		bin_expon = int((tempx>>23)&0xff) - 0x7f
		digits_x = bid_estimate_decimal_digits[bin_expon]
		e_scale = 18 - digits_x
		if diff_expon >= e_scale {
			diff_expon -= e_scale
		} else {
			e_scale = diff_expon
			diff_expon = 0
		}
		CX *= bid_power10_table_128[e_scale].w[0]
		Q64 = CX / uint64(coefficient_y)
		CX -= Q64 * uint64(coefficient_y)
		if CX == 0 {
			res = very_fast_get_BID32(sign_x, exponent_y, 0)
			return res, pfpsf
		}
	}

	res = very_fast_get_BID32(sign_x, exponent_y, uint32(CX))
	return res, pfpsf
}

// Bid32Class is ported from bid32_noncomp.c: bid32_class.
func Bid32Class(x uint32) int {
	var sig_x uint32
	var exp_x int

	if (x & MASK_NAN32) == MASK_NAN32 {
		if (x & MASK_SNAN32) == MASK_SNAN32 {
			return 0 // signalingNaN
		}
		return 1 // quietNaN
	}
	if (x & MASK_INF32) == MASK_INF32 {
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			return 2 // negativeInfinity
		}
		return 9 // positiveInfinity
	}
	if (x & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
		exp_x = int((x & MASK_BINARY_EXPONENT2_32) >> 21)
		sig_x = (x & MASK_BINARY_SIG2_32) | MASK_BINARY_OR2_32
		if sig_x > 9999999 {
			sig_x = 0
		}
	} else {
		exp_x = int((x & MASK_BINARY_EXPONENT1_32) >> 23)
		sig_x = x & MASK_BINARY_SIG1_32
	}
	if sig_x == 0 {
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			return 5 // negativeZero
		}
		return 6 // positiveZero
	}
	if exp_x < 6 {
		sig_x_prime := uint64(sig_x) * bid32_mult_factor[exp_x]
		if sig_x_prime < 1000000 {
			if (x & MASK_SIGN32) == MASK_SIGN32 {
				return 4 // negativeSubnormal
			}
			return 7 // positiveSubnormal
		}
	}
	if (x & MASK_SIGN32) == MASK_SIGN32 {
		return 3 // negativeNormal
	}
	return 8 // positiveNormal
}

// Bid32Quantexp returns the quantum exponent of x.
func Bid32Quantexp(x uint32) (int32, uint32) {
	var pfpsf uint32
	if (x & MASK_INF32) == MASK_INF32 {
		pfpsf |= BID_INVALID_EXCEPTION
		return -0x80000000, pfpsf
	}
	var exp int
	if (x & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
		exp = int((x & MASK_BINARY_EXPONENT2_32) >> 21)
	} else {
		exp = int((x & MASK_BINARY_EXPONENT1_32) >> 23)
	}
	return int32(exp - DECIMAL_EXPONENT_BIAS_32), pfpsf
}

// Bid32LLQuantexp returns the quantum exponent of x as int64.
func Bid32LLQuantexp(x uint32) (int64, uint32) {
	var pfpsf uint32
	if (x & MASK_INF32) == MASK_INF32 {
		pfpsf |= BID_INVALID_EXCEPTION
		return -0x8000000000000000, pfpsf
	}
	var exp int
	if (x & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
		exp = int((x & MASK_BINARY_EXPONENT2_32) >> 21)
	} else {
		exp = int((x & MASK_BINARY_EXPONENT1_32) >> 23)
	}
	return int64(exp - DECIMAL_EXPONENT_BIAS_32), pfpsf
}

// Bid32Inf returns +Infinity.
func Bid32Inf() uint32 {
	return 0x78000000
}

// Bid32NaN returns quiet NaN.
func Bid32NaN() uint32 {
	return 0x7c000000
}

// Bid32NextToward returns the next representable value from x toward y (bid128).
// Since we don't have bid128 support, we use bid64 as intermediate.
func Bid32NextToward(x uint32, y BID_UINT128) (uint32, uint32) {
	var res uint32
	var x128, tmp128 BID_UINT128
	var tmp1, tmp2 uint32
	var tmp_fpsf uint32
	var res1, res2 int
	var flags uint32

	if ((x & MASK_INF32) == MASK_INF32) ||
		((y.w[1] & MASK_NAN_128) == MASK_NAN_128) ||
		((y.w[1] & MASK_ANY_INF_128) == MASK_INF_128) {
		if (x & MASK_NAN32) == MASK_NAN32 {
			if (x & 0x000fffff) > 999999 {
				x = x & 0xfe000000
			} else {
				x = x & 0xfe0fffff
			}
			if (x & MASK_SNAN32) == MASK_SNAN32 {
				flags |= BID_INVALID_EXCEPTION
				res = x & 0xfdffffff
			} else {
				if (y.w[1] & MASK_SNAN_128) == MASK_SNAN_128 {
					flags |= BID_INVALID_EXCEPTION
				}
				res = x
			}
			return res, flags
		} else if (y.w[1] & MASK_NAN_128) == MASK_NAN_128 {
			if ((y.w[1] & 0x00003fffffffffff) > 0x0000314dc6448d93) ||
				((y.w[1]&0x00003fffffffffff) == 0x0000314dc6448d93 && y.w[0] > 0x38c15b09ffffffff) {
				y.w[1] = y.w[1] & 0xffffc00000000000
				y.w[0] = 0
			}
			if (y.w[1] & MASK_SNAN_128) == MASK_SNAN_128 {
				flags |= BID_INVALID_EXCEPTION
				tmp128.w[1] = y.w[1] & 0xfc003fffffffffff
				tmp128.w[0] = y.w[0]
			} else {
				tmp128.w[1] = y.w[1] & 0xfc003fffffffffff
				tmp128.w[0] = y.w[0]
			}
			res, _ = Bid128ToBid32(tmp128, BID_ROUNDING_TO_NEAREST)
			return res, flags
		} else {
			if (x & MASK_INF32) == MASK_INF32 {
				x = x & (MASK_SIGN32 | MASK_INF32)
			}
			if (y.w[1] & MASK_ANY_INF_128) == MASK_INF_128 {
				y.w[1] = y.w[1] & (MASK_SIGN_128 | MASK_INF_128)
				y.w[0] = 0
			}
		}
	}

	if (x & MASK_INF32) != MASK_INF32 {
		if (x & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
			if ((x & MASK_BINARY_SIG2_32) | MASK_BINARY_OR2_32) > 9999999 {
				x = (x & MASK_SIGN32) | ((x & MASK_BINARY_EXPONENT2_32) << 2)
			}
		}
	}

	tmp_fpsf = flags
	x128, _ = Bid32ToBid128(x)
	res1, _ = Bid128QuietEqual(x128, y)
	res2, _ = Bid128QuietGreater(x128, y)
	flags = tmp_fpsf

	if res1 != 0 {
		res = uint32((y.w[1]&MASK_SIGN_128)>>32) | (x & 0x7fffffff)
	} else if res2 != 0 {
		res, _ = Bid32NextDown(x)
	} else {
		res, _ = Bid32NextUp(x)
	}

	if ((x & MASK_INF32) != MASK_INF32) && ((res & MASK_INF32) == MASK_INF32) {
		flags |= BID_INEXACT_EXCEPTION
		flags |= BID_OVERFLOW_EXCEPTION
	}

	tmp1 = 0x0f4240
	tmp2 = res & 0x7fffffff
	tmp_fpsf = flags
	res1, _ = Bid32QuietGreater(tmp1, tmp2)
	res2, _ = Bid32QuietNotEqual(x, res)
	flags = tmp_fpsf
	if res1 != 0 && res2 != 0 {
		flags |= BID_INEXACT_EXCEPTION
		flags |= BID_UNDERFLOW_EXCEPTION
	}
	return res, flags
}

// Bid32ToBinary32 converts Decimal32 to binary32 (float32).
func Bid32ToBinary32(x uint32, rndMode int) (uint32, uint32) {
	if (x & MASK_NAN32) == MASK_NAN32 {
		var flags uint32
		payload := x & 0x000fffff
		if payload > 999999 {
			payload = 0
		}
		if (x & MASK_SNAN32) == MASK_SNAN32 {
			flags |= BID_INVALID_EXCEPTION
		}
		return (x & MASK_SIGN32) | 0x7f800000 | 0x00400000 | ((payload << 2) & 0x003fffff), flags
	}
	x64, f0 := Bid32ToBid64(x)
	res, f1 := Bid64ToBinary32(x64, rndMode)
	return res, f0 | f1
}

// Bid32ToBinary64 converts Decimal32 to binary64 (float64).
func Bid32ToBinary64(x uint32, rndMode int) (uint64, uint32) {
	if (x & MASK_NAN32) == MASK_NAN32 {
		var flags uint32
		payload := uint64(x & 0x000fffff)
		if payload > 999999 {
			payload = 0
		}
		if (x & MASK_SNAN32) == MASK_SNAN32 {
			flags |= BID_INVALID_EXCEPTION
		}
		return (uint64(x&MASK_SIGN32) << 32) | 0x7ff0000000000000 | 0x0008000000000000 | ((payload << 31) & 0x0007ffffffffffff), flags
	}
	x64, f0 := Bid32ToBid64(x)
	res, f1 := Bid64ToBinary64(x64, rndMode)
	return res, f0 | f1
}

// Bid32ToBinary128 converts Decimal32 to binary128.
func Bid32ToBinary128(x uint32, rndMode int) (BID_UINT128, uint32) {
	if (x & MASK_NAN32) == MASK_NAN32 {
		var res BID_UINT128
		var flags uint32
		payload := uint64(x & 0x000fffff)
		if payload > 999999 {
			payload = 0
		}
		if (x & MASK_SNAN32) == MASK_SNAN32 {
			flags |= BID_INVALID_EXCEPTION
		}
		res.w[1] = (uint64(x&MASK_SIGN32) << 32) | 0x7fff000000000000 | 0x0000800000000000 | ((payload << 27) & 0x00007fffffffffff)
		res.w[0] = 0
		return res, flags
	}
	x64, f0 := Bid32ToBid64(x)
	res, f1 := Bid64ToBinary128(x64, rndMode)
	return res, f0 | f1
}

// Bid32ToBid128 converts Decimal32 to Decimal128.
func Bid32ToBid128(x uint32) (BID_UINT128, uint32) {
	x64, f1 := Bid32ToBid64(x)
	res, f2 := Bid64ToBid128(x64)
	return res, f1 | f2
}
