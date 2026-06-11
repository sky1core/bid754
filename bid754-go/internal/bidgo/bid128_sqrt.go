// Ported from: Intel bid128_sqrt.c + bid_sqrt_macros.h
// Mechanical translation - all logic preserved exactly.
// Uses non-DOUBLE_EXTENDED path (double-based sqrt helpers).

package bidgo

import "math"

// short_sqrt128 computes approximate integer square root of a 128-bit value.
// Ported from bid_sqrt_macros.h (non-DOUBLE_EXTENDED_ON path).
func short_sqrt128(A10 BID_UINT128) uint64 {
	var ARS BID_UINT256
	var S BID_UINT192
	var AE0 BID_UINT256
	var AE BID_UINT192

	var MY, ES, CY uint64

	// 2^64
	l64 := math.Float64frombits(0x43f0000000000000)
	lx := noFmaMulAddF64(float64(A10.w[1]), l64, float64(A10.w[0]))
	ly_d := 1.0 / math.Sqrt(lx)
	ly_i := math.Float64bits(ly_d)

	MY = (ly_i & 0x000fffffffffffff) | 0x0010000000000000
	ey := int(0x3ff - (ly_i >> 52))

	// A10*RS^2
	ARS0 := __mul_64x128_to_192(MY, A10)
	ARS = __mul_64x192_to_256(MY, ARS0)

	// shr by 2*ey+40, to get a 64-bit value
	k := (ey << 1) + 104 - 64
	if k >= 128 {
		if k > 128 {
			ES = (ARS.w[2] >> uint(k-128)) | (ARS.w[3] << uint(192-k))
		} else {
			ES = ARS.w[2]
		}
	} else {
		if k >= 64 {
			ARS.w[0] = ARS.w[1]
			ARS.w[1] = ARS.w[2]
			k -= 64
		}
		if k != 0 {
			ARS_128 := __shr_128(BID_UINT128{w: [2]uint64{ARS.w[0], ARS.w[1]}}, uint(k))
			ARS.w[0] = ARS_128.w[0]
			ARS.w[1] = ARS_128.w[1]
		}
		ES = ARS.w[0]
	}

	ES = uint64(int64(ES) >> 1)

	if int64(ES) < 0 {
		ES = -ES

		// A*RS*eps (scaled by 2^64)
		AE0 = __mul_64x192_to_256(ES, ARS0)

		AE.w[0] = AE0.w[1]
		AE.w[1] = AE0.w[2]
		AE.w[2] = AE0.w[3]

		S.w[0], CY = __add_carry_out(ARS0.w[0], AE.w[0])
		S.w[1], CY = __add_carry_in_out(ARS0.w[1], AE.w[1], CY)
		S.w[2] = ARS0.w[2] + AE.w[2] + CY
	} else {
		// A*RS*eps (scaled by 2^64)
		AE0 = __mul_64x192_to_256(ES, ARS0)

		AE.w[0] = AE0.w[1]
		AE.w[1] = AE0.w[2]
		AE.w[2] = AE0.w[3]

		S.w[0], CY = __sub_borrow_out(ARS0.w[0], AE.w[0])
		S.w[1], CY = __sub_borrow_in_out(ARS0.w[1], AE.w[1], CY)
		S.w[2] = ARS0.w[2] - AE.w[2] - CY
	}

	k = ey + 51

	if k >= 64 {
		if k >= 128 {
			S.w[0] = S.w[2]
			S.w[1] = 0
			k -= 128
		} else {
			S.w[0] = S.w[1]
			S.w[1] = S.w[2]
		}
		k -= 64
	}
	if k != 0 {
		S_128 := __shr_128(BID_UINT128{w: [2]uint64{S.w[0], S.w[1]}}, uint(k))
		S.w[0] = S_128.w[0]
		S.w[1] = S_128.w[1]
	}

	return (S.w[0] + 1) >> 1
}

// bid_long_sqrt128 computes the approximate 128-bit integer square root of a 256-bit value.
// Ported from bid_sqrt_macros.h (non-DOUBLE_EXTENDED_ON path).
func bid_long_sqrt128(C256 BID_UINT256) BID_UINT128 {
	var S BID_UINT256
	var ES, ARS1, ES2 BID_UINT128
	var ARS00 BID_UINT256
	var AE, AE2 BID_UINT256
	var CY, MY, ES32 uint64

	// 2^64
	l64 := math.Float64frombits(0x43f0000000000000)

	l128 := l64 * l64
	lx := float64(C256.w[3]) * l64 * l128
	l2 := float64(C256.w[2]) * l128
	lx = lx + l2
	l1 := float64(C256.w[1]) * l64
	lx = lx + l1
	l0 := float64(C256.w[0])
	lx = lx + l0
	// sqrt(C256)
	ly_d := 1.0 / math.Sqrt(lx)
	ly_i := math.Float64bits(ly_d)

	MY = (ly_i & 0x000fffffffffffff) | 0x0010000000000000
	ey := int(0x3ff - (ly_i >> 52))

	// A10*RS^2, scaled by 2^(2*ey+104)
	ARS0 := __mul_64x256_to_320(MY, C256)
	ARS := __mul_64x320_to_384(MY, ARS0)

	// shr by k=(2*ey+104)-128-192
	k := (ey << 1) + 104 - 128 - 192
	k2 := 64 - k
	ES.w[0] = (ARS.w[3] >> uint(k+1)) | (ARS.w[4] << uint(k2-1))
	ES.w[1] = (ARS.w[4] >> uint(k)) | (ARS.w[5] << uint(k2))
	ES.w[1] = uint64(int64(ES.w[1]) >> 1)

	// A*RS >> 192 (for error term computation)
	ARS1.w[0] = ARS0.w[3]
	ARS1.w[1] = ARS0.w[4]

	// A*RS>>64
	ARS00.w[0] = ARS0.w[1]
	ARS00.w[1] = ARS0.w[2]
	ARS00.w[2] = ARS0.w[3]
	ARS00.w[3] = ARS0.w[4]

	if int64(ES.w[1]) < 0 {
		ES.w[0] = -ES.w[0]
		ES.w[1] = -ES.w[1]
		if ES.w[0] != 0 {
			ES.w[1]--
		}

		// A*RS*eps
		AE = __mul_128x128_to_256(ES, ARS1)

		S.w[0], CY = __add_carry_out(ARS00.w[0], AE.w[0])
		S.w[1], CY = __add_carry_in_out(ARS00.w[1], AE.w[1], CY)
		S.w[2], CY = __add_carry_in_out(ARS00.w[2], AE.w[2], CY)
		S.w[3] = ARS00.w[3] + AE.w[3] + CY
	} else {
		// A*RS*eps
		AE = __mul_128x128_to_256(ES, ARS1)

		S.w[0], CY = __sub_borrow_out(ARS00.w[0], AE.w[0])
		S.w[1], CY = __sub_borrow_in_out(ARS00.w[1], AE.w[1], CY)
		S.w[2], CY = __sub_borrow_in_out(ARS00.w[2], AE.w[2], CY)
		S.w[3] = ARS00.w[3] - AE.w[3] - CY
	}

	// 3/2*eps^2, scaled by 2^128
	ES32 = ES.w[1] + (ES.w[1] >> 1)
	ES2 = __mul_64x64_to_128(ES32, ES.w[1])
	// A*RS*3/2*eps^2
	AE2 = __mul_128x128_to_256(ES2, ARS1)

	// result, scaled by 2^(ey+52-64)
	S.w[0], CY = __add_carry_out(S.w[0], AE2.w[0])
	S.w[1], CY = __add_carry_in_out(S.w[1], AE2.w[1], CY)
	S.w[2], CY = __add_carry_in_out(S.w[2], AE2.w[2], CY)
	S.w[3] = S.w[3] + AE2.w[3] + CY

	// k in (0, 64)
	k = ey + 51 - 128
	k2 = 64 - k
	S.w[0] = (S.w[1] >> uint(k)) | (S.w[2] << uint(k2))
	S.w[1] = (S.w[2] >> uint(k)) | (S.w[3] << uint(k2))

	// round to nearest
	S.w[0]++
	if S.w[0] == 0 {
		S.w[1]++
	}

	var CS BID_UINT128
	CS.w[0] = (S.w[1] << 63) | (S.w[0] >> 1)
	CS.w[1] = S.w[1] >> 1

	return CS
}

// Bid128Sqrt computes the square root of a BID128 value.
// Ported from Intel bid128_sqrt.c: bid128_sqrt.
func Bid128Sqrt(x BID_UINT128, rnd_mode int) (BID_UINT128, uint32) {
	var M256, C256, C4, C8 BID_UINT256
	var CX, CX1, CX2, A10, S2, T128, TP128, CS, CSM, res BID_UINT128
	var sign_x, Carry uint64
	var D int64
	var exponent_x, bin_expon_cx int
	var digits, scale, exponent_q int
	var pfpsf uint32

	// unpack arguments, check for NaN or Infinity
	sign_x, exponent_x, CX, validBool := unpack_BID128_value(x)
	if !validBool {
		res.w[1] = CX.w[1]
		res.w[0] = CX.w[0]
		// NaN ?
		if (x.w[1] & 0x7c00000000000000) == 0x7c00000000000000 {
			if (x.w[1] & 0x7e00000000000000) == 0x7e00000000000000 { // sNaN
				pfpsf |= BID_INVALID_EXCEPTION
			}
			res.w[1] = CX.w[1] & QUIET_MASK64
			return res, pfpsf
		}
		// x is Infinity?
		if (x.w[1] & 0x7800000000000000) == 0x7800000000000000 {
			res.w[1] = CX.w[1]
			if sign_x != 0 {
				// -Inf, return NaN
				res.w[1] = 0x7c00000000000000
				pfpsf |= BID_INVALID_EXCEPTION
			}
			return res, pfpsf
		}
		// x is 0 otherwise
		res.w[1] = sign_x |
			((uint64(exponent_x+EXPONENT_BIAS128) >> 1) << 49)
		res.w[0] = 0
		return res, pfpsf
	}
	if sign_x != 0 {
		res.w[1] = 0x7c00000000000000
		res.w[0] = 0
		pfpsf |= BID_INVALID_EXCEPTION
		return res, pfpsf
	}

	// 2^64
	f64_i := uint32(0x5f800000)
	f64_d := math.Float32frombits(f64_i)

	// fx ~ CX
	fx_d := noFmaMulAddF32(float32(CX.w[1]), f64_d, float32(CX.w[0]))
	fx_i := math.Float32bits(fx_d)
	bin_expon_cx = int((fx_i>>23)&0xff) - 0x7f
	digits = bid_estimate_decimal_digits[bin_expon_cx]

	A10 = CX
	if (exponent_x & 1) != 0 {
		A10.w[1] = (CX.w[1] << 3) | (CX.w[0] >> 61)
		A10.w[0] = CX.w[0] << 3
		CX2.w[1] = (CX.w[1] << 1) | (CX.w[0] >> 63)
		CX2.w[0] = CX.w[0] << 1
		A10 = __add_128_128(A10, CX2)
	}

	CS.w[0] = short_sqrt128(A10)
	CS.w[1] = 0
	// check for exact result
	if CS.w[0]*CS.w[0] == A10.w[0] {
		S2 = __mul_64x64_to_128_fast(CS.w[0], CS.w[0])
		if S2.w[1] == A10.w[1] { // && S2.w[0]==A10.w[0]
			res = very_fast_get_BID128(0,
				(exponent_x+EXPONENT_BIAS128)>>1, CS)
			return res, pfpsf
		}
	}
	// get number of digits in CX
	D = int64(CX.w[1]) - int64(bid_power10_index_binexp_128[bin_expon_cx].w[1])
	if D > 0 ||
		(D == 0 && CX.w[0] >= bid_power10_index_binexp_128[bin_expon_cx].w[0]) {
		digits++
	}

	// if exponent is odd, scale coefficient by 10
	scale = 67 - digits
	exponent_q = exponent_x - scale
	scale += (exponent_q & 1) // exp. bias is even

	if scale > 38 {
		T128 = bid_power10_table_128[scale-37]
		CX1 = __mul_128x128_low(CX, T128)

		TP128 = bid_power10_table_128[37]
		C256 = __mul_128x128_to_256(CX1, TP128)
	} else {
		T128 = bid_power10_table_128[scale]
		C256 = __mul_128x128_to_256(CX, T128)
	}

	// 4*C256
	C4.w[3] = (C256.w[3] << 2) | (C256.w[2] >> 62)
	C4.w[2] = (C256.w[2] << 2) | (C256.w[1] >> 62)
	C4.w[1] = (C256.w[1] << 2) | (C256.w[0] >> 62)
	C4.w[0] = C256.w[0] << 2

	CS = bid_long_sqrt128(C256)

	if (rnd_mode & 3) == 0 {
		// compare to midpoints
		CSM.w[1] = (CS.w[1] << 1) | (CS.w[0] >> 63)
		CSM.w[0] = (CS.w[0] + CS.w[0]) | 1
		// CSM^2
		M256 = __sqr128_to_256(CSM)

		if C4.w[3] > M256.w[3] ||
			(C4.w[3] == M256.w[3] &&
				(C4.w[2] > M256.w[2] ||
					(C4.w[2] == M256.w[2] &&
						(C4.w[1] > M256.w[1] ||
							(C4.w[1] == M256.w[1] &&
								C4.w[0] > M256.w[0]))))) {
			// round up
			CS.w[0]++
			if CS.w[0] == 0 {
				CS.w[1]++
			}
		} else {
			C8.w[1] = (CS.w[1] << 3) | (CS.w[0] >> 61)
			C8.w[0] = CS.w[0] << 3
			// M256 - 8*CSM
			M256.w[0], Carry = __sub_borrow_out(M256.w[0], C8.w[0])
			M256.w[1], Carry = __sub_borrow_in_out(M256.w[1], C8.w[1], Carry)
			M256.w[2], Carry = __sub_borrow_in_out(M256.w[2], 0, Carry)
			M256.w[3] = M256.w[3] - Carry

			// if CSM' > C256, round up
			if M256.w[3] > C4.w[3] ||
				(M256.w[3] == C4.w[3] &&
					(M256.w[2] > C4.w[2] ||
						(M256.w[2] == C4.w[2] &&
							(M256.w[1] > C4.w[1] ||
								(M256.w[1] == C4.w[1] &&
									M256.w[0] > C4.w[0]))))) {
				// round down
				if CS.w[0] == 0 {
					CS.w[1]--
				}
				CS.w[0]--
			}
		}
	} else {
		M256 = __sqr128_to_256(CS)
		C8.w[1] = (CS.w[1] << 1) | (CS.w[0] >> 63)
		C8.w[0] = CS.w[0] << 1
		if M256.w[3] > C256.w[3] ||
			(M256.w[3] == C256.w[3] &&
				(M256.w[2] > C256.w[2] ||
					(M256.w[2] == C256.w[2] &&
						(M256.w[1] > C256.w[1] ||
							(M256.w[1] == C256.w[1] &&
								M256.w[0] > C256.w[0]))))) {
			M256.w[0], Carry = __sub_borrow_out(M256.w[0], C8.w[0])
			M256.w[1], Carry = __sub_borrow_in_out(M256.w[1], C8.w[1], Carry)
			M256.w[2], Carry = __sub_borrow_in_out(M256.w[2], 0, Carry)
			M256.w[3] = M256.w[3] - Carry
			M256.w[0]++
			if M256.w[0] == 0 {
				M256.w[1]++
				if M256.w[1] == 0 {
					M256.w[2]++
					if M256.w[2] == 0 {
						M256.w[3]++
					}
				}
			}

			if CS.w[0] == 0 {
				CS.w[1]--
			}
			CS.w[0]--

			if M256.w[3] > C256.w[3] ||
				(M256.w[3] == C256.w[3] &&
					(M256.w[2] > C256.w[2] ||
						(M256.w[2] == C256.w[2] &&
							(M256.w[1] > C256.w[1] ||
								(M256.w[1] == C256.w[1] &&
									M256.w[0] > C256.w[0]))))) {

				if CS.w[0] == 0 {
					CS.w[1]--
				}
				CS.w[0]--
			}
		} else {
			M256.w[0], Carry = __add_carry_out(M256.w[0], C8.w[0])
			M256.w[1], Carry = __add_carry_in_out(M256.w[1], C8.w[1], Carry)
			M256.w[2], Carry = __add_carry_in_out(M256.w[2], 0, Carry)
			M256.w[3] = M256.w[3] + Carry
			M256.w[0]++
			if M256.w[0] == 0 {
				M256.w[1]++
				if M256.w[1] == 0 {
					M256.w[2]++
					if M256.w[2] == 0 {
						M256.w[3]++
					}
				}
			}
			if M256.w[3] < C256.w[3] ||
				(M256.w[3] == C256.w[3] &&
					(M256.w[2] < C256.w[2] ||
						(M256.w[2] == C256.w[2] &&
							(M256.w[1] < C256.w[1] ||
								(M256.w[1] == C256.w[1] &&
									M256.w[0] <= C256.w[0]))))) {

				CS.w[0]++
				if CS.w[0] == 0 {
					CS.w[1]++
				}
			}
		}
		// RU?
		if rnd_mode == BID_ROUNDING_UP {
			CS.w[0]++
			if CS.w[0] == 0 {
				CS.w[1]++
			}
		}
	}

	pfpsf |= BID_INEXACT_EXCEPTION
	res = bid_get_BID128_fast(0, (exponent_q+EXPONENT_BIAS128)>>1, CS)
	return res, pfpsf
}
