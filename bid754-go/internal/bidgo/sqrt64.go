package bidgo

import "math"

// Bid64Sqrt is ported mechanically from bid64_sqrt.c.
func Bid64Sqrt(x uint64, rndMode int) (uint64, uint32) {
	var CA BID_UINT128
	var sign_x, coefficient_x uint64
	var Q, Q2, A10, C4, R, R2, QE, res uint64
	var D int64
	var t_scale uint64
	var da, dq, da_h, da_l, dqe float64
	var exponent_x, exponent_q, bin_expon_cx int
	var digits_x int
	var scale int
	var pfpsf uint32

	var valid bool
	sign_x, exponent_x, coefficient_x, valid = unpack_BID64(x)
	if !valid {
		if (x & INFINITY_MASK64) == INFINITY_MASK64 {
			res = coefficient_x
			if (coefficient_x & SSNAN_MASK64) == SINFINITY_MASK64 {
				res = NAN_MASK64
				pfpsf |= BID_INVALID_EXCEPTION
			}
			if (x & SNAN_MASK64) == SNAN_MASK64 {
				pfpsf |= BID_INVALID_EXCEPTION
			}
			return res & QUIET_MASK64, pfpsf
		}
		exponent_x = (exponent_x + DECIMAL_EXPONENT_BIAS) >> 1
		res = sign_x | (uint64(exponent_x) << 53)
		return res, pfpsf
	}
	if sign_x != 0 && coefficient_x != 0 {
		res = NAN_MASK64
		pfpsf |= BID_INVALID_EXCEPTION
		return res, pfpsf
	}
	t_scale = 0x43f0000000000000
	bin_expon_cx = int((math.Float32bits(float32(coefficient_x))>>23)&0xff) - 0x7f
	digits_x = bid_estimate_decimal_digits[bin_expon_cx]
	if coefficient_x >= bid_power10_index_binexp[bin_expon_cx] {
		digits_x++
	}

	A10 = coefficient_x
	if (exponent_x & 1) != 0 {
		A10 = (A10 << 2) + A10
		A10 += A10
	}

	dqe = math.Sqrt(float64(A10))
	QE = uint64(dqe)
	if QE*QE == A10 {
		res = very_fast_get_BID64(0, (exponent_x+DECIMAL_EXPONENT_BIAS)>>1, QE)
		return res, pfpsf
	}
	scale = 31 - digits_x
	exponent_q = exponent_x - scale
	scale += (exponent_q & 1)

	CT := bid_power10_table_128[scale]
	CA = __mul_64x128_short(coefficient_x, CT)

	da_h = float64(CA.hi)
	da_l = float64(CA.lo)
	da = noFmaMulAddF64(da_h, math.Float64frombits(t_scale), da_l)

	dq = math.Sqrt(da)
	Q = uint64(dq)

	R = uint64(int64(CA.lo-Q*Q) >> 63)
	D = int64(R + R + 1)

	exponent_q = (exponent_q + DECIMAL_EXPONENT_BIAS) >> 1

	pfpsf |= BID_INEXACT_EXCEPTION

	if (rndMode & 3) == 0 {
		Q2 = Q + Q + uint64(D)
		C4 = CA.lo << 2
		R2 = uint64(int64(Q2*Q2-C4) >> 63)
		Q += uint64(D) & (R ^ R2)
	} else {
		C4 = CA.lo
		Q += uint64(D)
		if int64(Q*Q-C4) > 0 {
			Q--
		}
		if rndMode == BID_ROUNDING_UP {
			Q++
		}
	}

	res = fast_get_BID64(0, exponent_q, Q)
	return res, pfpsf
}

// Bid64qSqrt computes the Decimal64 square root of a BID128 value.
// Ported mechanically from Intel bid64_sqrt.c: bid64q_sqrt.
func Bid64qSqrt(x BID_UINT128, rndMode int) (uint64, uint32) {
	var M256, C4, C8 BID_UINT128
	var CX, CX2, A10, S2, T128, CS, CSM, CS2, C256, CS1 BID_UINT128
	var mulFactor2Long, QH, Tmp, TP128, Qh BID_UINT128
	var signX, carry, B10, res, mulFactor, mulFactor2, CS0 uint64
	var D int64
	var exponentX, binExponentCX int
	var digits, scale, exponentQ, amount, extraDigits int
	var done bool
	exact := true
	var pfpsf uint32

	// Unpack arguments, check for NaN or Infinity.
	signX, exponentX, CX, valid := unpack_BID128_value(x)
	if !valid {
		res = CX.hi
		// NaN?
		if (x.hi & 0x7c00000000000000) == 0x7c00000000000000 {
			if (x.hi & 0x7e00000000000000) == 0x7e00000000000000 { // sNaN
				pfpsf |= BID_INVALID_EXCEPTION
			}
			Tmp.hi = CX.hi & 0x00003fffffffffff
			Tmp.lo = CX.lo
			TP128 = bid_reciprocals10_128[18]
			Qh, _ = __mul_128x128_full(Tmp, TP128)
			amount = bid_recip_scale[18]
			Tmp = __shr_128(Qh, uint(amount))
			res = (CX.hi & 0xfc00000000000000) | Tmp.lo
			return res, pfpsf
		}
		// Infinity?
		if (x.hi & 0x7800000000000000) == 0x7800000000000000 {
			if signX != 0 {
				// -Inf, return NaN.
				res = 0x7c00000000000000
				pfpsf |= BID_INVALID_EXCEPTION
			}
			return res, pfpsf
		}
		// x is zero otherwise.
		exponentX = ((exponentX - DECIMAL_EXPONENT_BIAS_128) >> 1) +
			DECIMAL_EXPONENT_BIAS
		if exponentX < 0 {
			exponentX = 0
		}
		if exponentX > DECIMAL_MAX_EXPON_64 {
			exponentX = DECIMAL_MAX_EXPON_64
		}
		res, flags := get_BID64_flags(signX, exponentX, 0, rndMode)
		pfpsf |= flags
		return res, pfpsf
	}
	if signX != 0 {
		res = 0x7c00000000000000
		pfpsf |= BID_INVALID_EXCEPTION
		return res, pfpsf
	}

	// 2^64 and fx ~ CX.
	f64 := math.Float32frombits(0x5f800000)
	fx := noFmaMulAddF32(float32(CX.hi), f64, float32(CX.lo))
	binExponentCX = int((math.Float32bits(fx)>>23)&0xff) - 0x7f
	digits = bid_estimate_decimal_digits[binExponentCX]

	A10 = CX
	if (exponentX & 1) != 0 {
		A10.hi = (CX.hi << 3) | (CX.lo >> 61)
		A10.lo = CX.lo << 3
		CX2.hi = (CX.hi << 1) | (CX.lo >> 63)
		CX2.lo = CX.lo << 1
		A10 = __add_128_128(A10, CX2)
	}

	C256 = A10
	CS.lo = short_sqrt128(A10)
	CS.hi = 0
	mulFactor = 0
	// Check for an exact result.
	if CS.lo < 10000000000000000 {
		if CS.lo*CS.lo == A10.lo {
			S2 = __mul_64x64_to_128_fast(CS.lo, CS.lo)
			if S2.hi == A10.hi { // S2.lo == A10.lo was checked above.
				res, flags := get_BID64_flags(0,
					((exponentX-DECIMAL_EXPONENT_BIAS_128)>>1)+
						DECIMAL_EXPONENT_BIAS,
					CS.lo, rndMode)
				pfpsf |= flags
				return res, pfpsf
			}
		}
		if CS.lo >= 1000000000000000 {
			done = true
			exponentQ = exponentX
			C256 = A10
		}
		pfpsf |= BID_INEXACT_EXCEPTION
		exact = false
	} else {
		B10 = 0x3333333333333334
		CS2 = __mul_64x64_to_128(CS.lo, B10)
		CS0 = CS2.hi >> 1
		if CS.lo != (CS0<<3)+(CS0<<1) {
			pfpsf |= BID_INEXACT_EXCEPTION
			exact = false
		}
		done = true
		CS.lo = CS0
		exponentQ = exponentX + 2
		mulFactor = 10
		mulFactor2 = 100
		if CS.lo >= 10000000000000000 {
			CS2 = __mul_64x64_to_128(CS.lo, B10)
			CS0 = CS2.hi >> 1
			if CS.lo != (CS0<<3)+(CS0<<1) {
				pfpsf |= BID_INEXACT_EXCEPTION
				exact = false
			}
			exponentQ += 2
			CS.lo = CS0
			mulFactor = 100
			mulFactor2 = 10000
		}
		if exact {
			CS0 = CS.lo * mulFactor
			CS1 = __mul_64x64_to_128_fast(CS0, CS0)
			if CS1.lo != A10.lo || CS1.hi != A10.hi {
				pfpsf |= BID_INEXACT_EXCEPTION
				exact = false
			}
		}
	}

	if !done {
		// Get the number of digits in CX.
		D = int64(CX.hi) - int64(bid_power10_index_binexp_128[binExponentCX].hi)
		if D > 0 ||
			(D == 0 && CX.lo >= bid_power10_index_binexp_128[binExponentCX].lo) {
			digits++
		}

		// Scale the coefficient so the result has Decimal64 precision.
		scale = 31 - digits
		exponentQ = exponentX - scale
		scale += exponentQ & 1 // The exponent bias is even.

		T128 = bid_power10_table_128[scale]
		C256 = __mul_128x128_low(CX, T128)
		CS.lo = short_sqrt128(C256)
	}

	exponentQ = ((exponentQ - DECIMAL_EXPONENT_BIAS_128) >> 1) +
		DECIMAL_EXPONENT_BIAS
	if exponentQ < 0 && exponentQ+MAX_FORMAT_DIGITS >= 0 {
		extraDigits = -exponentQ
		exponentQ = 0

		// Get coeff*(2^M[extraDigits])/10^extraDigits.
		QH = __mul_64x64_to_128(CS.lo, bid_reciprocals10_64[extraDigits])
		amount = bid_short_recip_scale[extraDigits]
		CS0 = QH.hi >> uint(amount)

		if exact {
			if CS.lo != CS0*bid_power10_table_128[extraDigits].lo {
				exact = false
			}
		}
		if !exact {
			pfpsf |= BID_UNDERFLOW_EXCEPTION | BID_INEXACT_EXCEPTION
		}

		CS.lo = CS0
		if mulFactor == 0 {
			mulFactor = 1
		}
		mulFactor *= bid_power10_table_128[extraDigits].lo
		mulFactor2Long = __mul_64x64_to_128(mulFactor, mulFactor)
		if mulFactor2Long.hi != 0 {
			mulFactor2 = 0
		} else {
			// Local lowering from Intel's general 64x128 path: when the square
			// fits in 64 bits, the 64x64 product has the same low-128 result.
			mulFactor2 = mulFactor2Long.lo
		}
	}

	// 4*C256.
	C4.hi = (C256.hi << 2) | (C256.lo >> 62)
	C4.lo = C256.lo << 2

	if (rndMode & 3) == 0 {
		// Compare to midpoints.
		CSM.lo = (CS.lo + CS.lo) | 1
		if mulFactor != 0 {
			CSM.lo *= mulFactor
		}
		M256 = __mul_64x64_to_128(CSM.lo, CSM.lo)

		if C4.hi > M256.hi || (C4.hi == M256.hi && C4.lo > M256.lo) {
			CS.lo++
		} else {
			C8.lo = CS.lo << 3
			C8.hi = 0
			if mulFactor != 0 {
				if mulFactor2 != 0 {
					C8 = __mul_64x64_to_128(C8.lo, mulFactor2)
				} else {
					C8 = __mul_64x128_to_128(C8.lo, mulFactor2Long)
				}
			}
			M256.lo, carry = __sub_borrow_out(M256.lo, C8.lo)
			M256.hi = M256.hi - C8.hi - carry

			if M256.hi > C4.hi || (M256.hi == C4.hi && M256.lo > C4.lo) {
				if CS.lo != 0 {
					CS.lo--
				}
			}
		}
	} else {
		CS.lo++
		CSM.lo = CS.lo
		C8.lo = CSM.lo << 1
		if mulFactor != 0 {
			CSM.lo *= mulFactor
		}
		M256 = __mul_64x64_to_128(CSM.lo, CSM.lo)
		C8.hi = 0
		if mulFactor != 0 {
			if mulFactor2 != 0 {
				C8 = __mul_64x64_to_128(C8.lo, mulFactor2)
			} else {
				C8 = __mul_64x128_to_128(C8.lo, mulFactor2Long)
			}
		}

		if M256.hi > C256.hi || (M256.hi == C256.hi && M256.lo > C256.lo) {
			M256.lo, carry = __sub_borrow_out(M256.lo, C8.lo)
			M256.hi = M256.hi - carry - C8.hi
			M256.lo++
			if M256.lo == 0 {
				M256.hi++
			}

			if (M256.hi > C256.hi ||
				(M256.hi == C256.hi && M256.lo > C256.lo)) && CS.lo > 1 {
				CS.lo--

				if CS.lo > 1 {
					M256.lo, carry = __sub_borrow_out(M256.lo, C8.lo)
					M256.hi = M256.hi - carry - C8.hi
					M256.lo++
					if M256.lo == 0 {
						M256.hi++
					}

					if M256.hi > C256.hi ||
						(M256.hi == C256.hi && M256.lo > C256.lo) {
						CS.lo--
					}
				}
			}
		} else {
			CS.lo++
		}
		// Round up only for an inexact result.
		if rndMode != BID_ROUNDING_UP || exact {
			if CS.lo != 0 {
				CS.lo--
			}
		}
	}

	res, flags := get_BID64_flags(0, exponentQ, CS.lo, rndMode)
	pfpsf |= flags
	return res, pfpsf
}
