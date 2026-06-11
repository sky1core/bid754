// Ported from: Intel bid64_to_bid128.c (bid128_to_bid64 section)
// and Intel bid32_to_bid128.c (bid128_to_bid32 section)
// Mechanical translation - all logic preserved exactly.

package bidgo

import "math"

// DECIMAL_EXPONENT_BIAS_128 matches C name. Same as EXPONENT_BIAS128 in bid128_internal.go.
const DECIMAL_EXPONENT_BIAS_128 = 6176

// Bid128ToBid64 converts a BID128 to BID64 and returns status flags.
// Ported mechanically from Intel bid64_to_bid128.c: bid128_to_bid64.
func Bid128ToBid64(x BID_UINT128, rnd_mode int) (uint64, uint32) {
	var CX, T128, TP128, Qh, Ql, Qh1, Stemp, Tmp, Tmp1, CX1 BID_UINT128
	var sign_x, carry, cy, res uint64
	var D int64
	var exponent_x, extra_digits, amount, bin_expon_cx int
	var rmode, status, uf_check uint32
	var pfpsf uint32

	// BID_SWAP128 is no-op on little-endian
	// unpack arguments, check for NaN or Infinity or 0
	sign_x, exponent_x, CX, valid := unpack_BID128_value(x)
	if !valid {
		if (x.w[1] << 1) >= 0xf000000000000000 {
			Tmp.w[1] = CX.w[1] & 0x00003fffffffffff
			Tmp.w[0] = CX.w[0]
			TP128 = bid_reciprocals10_128[18]
			Qh, Ql = __mul_128x128_full(Tmp, TP128)
			amount = bid_recip_scale[18]
			Tmp = __shr_128(Qh, uint(amount))
			res = (CX.w[1] & 0xfc00000000000000) | Tmp.w[0]
			if (x.w[1] & SNAN_MASK64) == SNAN_MASK64 { // sNaN
				pfpsf |= BID_INVALID_EXCEPTION
			}
			return res, pfpsf
		}
		exponent_x =
			exponent_x - DECIMAL_EXPONENT_BIAS_128 + DECIMAL_EXPONENT_BIAS
		if exponent_x < 0 {
			res = sign_x
			return res, pfpsf
		}
		if exponent_x > DECIMAL_MAX_EXPON_64 {
			exponent_x = DECIMAL_MAX_EXPON_64
		}
		res = sign_x | (uint64(exponent_x) << 53)
		return res, pfpsf
	}

	if CX.w[1] != 0 || (CX.w[0] >= 10000000000000000) {
		// find number of digits in coefficient
		// 2^64
		f64_i := uint32(0x5f800000)
		// fx ~ CX
		fx_d := noFmaMulAddF32(float32(CX.w[1]), math.Float32frombits(f64_i), float32(CX.w[0]))
		fx_i := math.Float32bits(fx_d)
		bin_expon_cx = int((fx_i>>23)&0xff) - 0x7f
		extra_digits = bid_estimate_decimal_digits[bin_expon_cx] - 16
		// scale = 38-estimate_decimal_digits[bin_expon_cx];
		D = int64(CX.w[1]) - int64(bid_power10_index_binexp_128[bin_expon_cx].w[1])
		if D > 0 ||
			(D == 0 &&
				CX.w[0] >= bid_power10_index_binexp_128[bin_expon_cx].w[0]) {
			extra_digits++
		}

		exponent_x += extra_digits

		rmode = uint32(rnd_mode)
		if sign_x != 0 && uint32(rmode-1) < 2 {
			rmode = 3 - rmode
		}

		if exponent_x < DECIMAL_EXPONENT_BIAS_128-DECIMAL_EXPONENT_BIAS {
			uf_check = 1
			if -extra_digits+exponent_x-DECIMAL_EXPONENT_BIAS_128+
				DECIMAL_EXPONENT_BIAS+35 >= 0 {
				if exponent_x ==
					DECIMAL_EXPONENT_BIAS_128-DECIMAL_EXPONENT_BIAS-1 {
					T128 = bid_round_const_table_128[rmode][extra_digits]
					CX1.w[0], carry = __add_carry_out(T128.w[0], CX.w[0])
					CX1.w[1] = CX.w[1] + T128.w[1] + carry
					// DECIMAL_TINY_DETECTION_AFTER_ROUNDING is 0, skip
				}
				extra_digits =
					extra_digits + DECIMAL_EXPONENT_BIAS_128 -
						DECIMAL_EXPONENT_BIAS - exponent_x
				exponent_x = DECIMAL_EXPONENT_BIAS_128 - DECIMAL_EXPONENT_BIAS
			} else {
				rmode = BID_ROUNDING_TO_ZERO
			}
		}

		T128 = bid_round_const_table_128[rmode][extra_digits]
		CX.w[0], carry = __add_carry_out(T128.w[0], CX.w[0])
		CX.w[1] = CX.w[1] + T128.w[1] + carry

		TP128 = bid_reciprocals10_128[extra_digits]
		Qh, Ql = __mul_128x128_full(CX, TP128)
		amount = bid_recip_scale[extra_digits]

		if amount >= 64 {
			CX.w[0] = Qh.w[1] >> uint(amount-64)
			CX.w[1] = 0
		} else {
			CX = __shr_128(Qh, uint(amount))
		}

		if rnd_mode == BID_ROUNDING_TO_NEAREST {
			if CX.w[0]&1 != 0 {
				// check whether fractional part of initial_P/10^ed1 is exactly .5

				// get remainder
				Qh1 = __shl_128_long(Qh, uint(128-amount))

				if Qh1.w[1] == 0 && Qh1.w[0] == 0 &&
					(Ql.w[1] < bid_reciprocals10_128[extra_digits].w[1] ||
						(Ql.w[1] == bid_reciprocals10_128[extra_digits].w[1] &&
							Ql.w[0] < bid_reciprocals10_128[extra_digits].w[0])) {
					CX.w[0]--
				}
			}
		}

		{
			status = BID_INEXACT_EXCEPTION
			// get remainder
			Qh1 = __shl_128_long(Qh, uint(128-amount))

			switch rmode {
			case BID_ROUNDING_TO_NEAREST, BID_ROUNDING_TIES_AWAY:
				// test whether fractional part is 0
				if Qh1.w[1] == 0x8000000000000000 && Qh1.w[0] == 0 &&
					(Ql.w[1] < bid_reciprocals10_128[extra_digits].w[1] ||
						(Ql.w[1] == bid_reciprocals10_128[extra_digits].w[1] &&
							Ql.w[0] < bid_reciprocals10_128[extra_digits].w[0])) {
					status = BID_EXACT_STATUS
				}
			case BID_ROUNDING_DOWN, BID_ROUNDING_TO_ZERO:
				if Qh1.w[1] == 0 && Qh1.w[0] == 0 &&
					(Ql.w[1] < bid_reciprocals10_128[extra_digits].w[1] ||
						(Ql.w[1] == bid_reciprocals10_128[extra_digits].w[1] &&
							Ql.w[0] < bid_reciprocals10_128[extra_digits].w[0])) {
					status = BID_EXACT_STATUS
				}
			default:
				// round up
				Stemp.w[0], cy = __add_carry_out(Ql.w[0],
					bid_reciprocals10_128[extra_digits].w[0])
				Stemp.w[1], carry = __add_carry_in_out(Ql.w[1],
					bid_reciprocals10_128[extra_digits].w[1], cy)
				Qh = __shr_128_long(Qh1, uint(128-amount))
				Tmp.w[0] = 1
				Tmp.w[1] = 0
				Tmp1 = __shl_128_long(Tmp, uint(amount))
				Qh.w[0] += carry
				if Qh.w[0] < carry {
					Qh.w[1]++
				}
				if __unsigned_compare_ge_128(Qh, Tmp1) {
					status = BID_EXACT_STATUS
				}
			}

			if status != BID_EXACT_STATUS {
				if uf_check != 0 {
					status |= BID_UNDERFLOW_EXCEPTION
				}
				pfpsf |= status
			}
		}

		_ = CX1
		_ = Stemp
	}

	res, flags := get_BID64_flags(sign_x,
		exponent_x-DECIMAL_EXPONENT_BIAS_128+DECIMAL_EXPONENT_BIAS,
		CX.w[0], rnd_mode)
	pfpsf |= flags
	return res, pfpsf
}

// Bid128ToBid32 converts a BID128 to BID32 and returns status flags.
// Ported mechanically from Intel bid32_to_bid128.c: bid128_to_bid32.
func Bid128ToBid32(x BID_UINT128, rnd_mode int) (uint32, uint32) {
	var CX, T128, TP128, Qh, Ql, Qh1, Stemp, Tmp, Tmp1, CX1 BID_UINT128
	var sign_x, carry, cy uint64
	var D int64
	var res uint32
	var exponent_x, extra_digits, amount, bin_expon_cx int
	var uf_check int
	var rmode, status uint32
	var pfpsf uint32

	// BID_SWAP128 is no-op on little-endian
	// unpack arguments, check for NaN or Infinity or 0
	sign_x, exponent_x, CX, valid := unpack_BID128_value(x)
	if !valid {
		if (x.w[1] & 0x7800000000000000) == 0x7800000000000000 {
			Tmp.w[1] = CX.w[1] & 0x00003fffffffffff
			Tmp.w[0] = CX.w[0]
			TP128 = bid_reciprocals10_128[27]
			Qh, Ql = __mul_128x128_full(Tmp, TP128)
			amount = bid_recip_scale[27] - 64
			res = uint32((CX.w[1]>>32)&0xfc000000) | uint32(Qh.w[1]>>uint(amount))
			if (x.w[1] & SNAN_MASK64) == SNAN_MASK64 { // sNaN
				pfpsf |= BID_INVALID_EXCEPTION
			}
			_ = Ql
			return res, pfpsf
		}
		// x is 0
		exponent_x =
			exponent_x - DECIMAL_EXPONENT_BIAS_128 + DECIMAL_EXPONENT_BIAS_32
		if exponent_x < 0 {
			exponent_x = 0
		}
		if exponent_x > DECIMAL_MAX_EXPON_32 {
			exponent_x = DECIMAL_MAX_EXPON_32
		}
		res = uint32(sign_x>>32) | uint32(exponent_x<<23)
		return res, pfpsf
	}

	if CX.w[1] != 0 || (CX.w[0] >= 10000000) {
		// find number of digits in coefficient
		// 2^64
		f64_i := uint32(0x5f800000)
		// fx ~ CX
		fx_d := noFmaMulAddF32(float32(CX.w[1]), math.Float32frombits(f64_i), float32(CX.w[0]))
		fx_i := math.Float32bits(fx_d)
		bin_expon_cx = int((fx_i>>23)&0xff) - 0x7f
		extra_digits = bid_estimate_decimal_digits[bin_expon_cx] - 7
		// scale = 38-estimate_decimal_digits[bin_expon_cx];
		D = int64(CX.w[1]) - int64(bid_power10_index_binexp_128[bin_expon_cx].w[1])
		if D > 0 ||
			(D == 0 &&
				CX.w[0] >= bid_power10_index_binexp_128[bin_expon_cx].w[0]) {
			extra_digits++
		}

		exponent_x += extra_digits

		rmode = uint32(rnd_mode)
		if sign_x != 0 && uint32(rmode-1) < 2 {
			rmode = 3 - rmode
		}

		if exponent_x <
			DECIMAL_EXPONENT_BIAS_128-DECIMAL_EXPONENT_BIAS_32 {
			uf_check = 1
			if -extra_digits+exponent_x-DECIMAL_EXPONENT_BIAS_128+
				DECIMAL_EXPONENT_BIAS_32+35 >= 0 {
				if exponent_x ==
					DECIMAL_EXPONENT_BIAS_128-DECIMAL_EXPONENT_BIAS_32-1 {
					T128 = bid_round_const_table_128[rmode][extra_digits]
					CX1.w[0], carry = __add_carry_out(T128.w[0], CX.w[0])
					CX1.w[1] = CX.w[1] + T128.w[1] + carry
					// DECIMAL_TINY_DETECTION_AFTER_ROUNDING is 0, skip
				}
				extra_digits =
					extra_digits + DECIMAL_EXPONENT_BIAS_128 -
						DECIMAL_EXPONENT_BIAS_32 - exponent_x
				exponent_x =
					DECIMAL_EXPONENT_BIAS_128 - DECIMAL_EXPONENT_BIAS_32
			} else {
				rmode = BID_ROUNDING_TO_ZERO
			}
		}

		T128 = bid_round_const_table_128[rmode][extra_digits]
		CX.w[0], carry = __add_carry_out(T128.w[0], CX.w[0])
		CX.w[1] = CX.w[1] + T128.w[1] + carry

		TP128 = bid_reciprocals10_128[extra_digits]
		Qh, Ql = __mul_128x128_full(CX, TP128)
		amount = bid_recip_scale[extra_digits]

		if amount >= 64 {
			CX.w[0] = Qh.w[1] >> uint(amount-64)
			CX.w[1] = 0
		} else {
			CX = __shr_128(Qh, uint(amount))
		}

		if rnd_mode == BID_ROUNDING_TO_NEAREST {
			if CX.w[0]&1 != 0 {
				// check whether fractional part of initial_P/10^ed1 is exactly .5

				// get remainder
				Qh1 = __shl_128_long(Qh, uint(128-amount))

				if Qh1.w[1] == 0 && Qh1.w[0] == 0 &&
					(Ql.w[1] < bid_reciprocals10_128[extra_digits].w[1] ||
						(Ql.w[1] == bid_reciprocals10_128[extra_digits].w[1] &&
							Ql.w[0] < bid_reciprocals10_128[extra_digits].w[0])) {
					CX.w[0]--
				}
			}
		}

		{
			status = BID_INEXACT_EXCEPTION
			// get remainder
			Qh1 = __shl_128_long(Qh, uint(128-amount))

			switch rmode {
			case BID_ROUNDING_TO_NEAREST, BID_ROUNDING_TIES_AWAY:
				// test whether fractional part is 0
				if Qh1.w[1] == 0x8000000000000000 && Qh1.w[0] == 0 &&
					(Ql.w[1] < bid_reciprocals10_128[extra_digits].w[1] ||
						(Ql.w[1] == bid_reciprocals10_128[extra_digits].w[1] &&
							Ql.w[0] < bid_reciprocals10_128[extra_digits].w[0])) {
					status = BID_EXACT_STATUS
				}
			case BID_ROUNDING_DOWN, BID_ROUNDING_TO_ZERO:
				if Qh1.w[1] == 0 && Qh1.w[0] == 0 &&
					(Ql.w[1] < bid_reciprocals10_128[extra_digits].w[1] ||
						(Ql.w[1] == bid_reciprocals10_128[extra_digits].w[1] &&
							Ql.w[0] < bid_reciprocals10_128[extra_digits].w[0])) {
					status = BID_EXACT_STATUS
				}
			default:
				// round up
				Stemp.w[0], cy = __add_carry_out(Ql.w[0],
					bid_reciprocals10_128[extra_digits].w[0])
				Stemp.w[1], carry = __add_carry_in_out(Ql.w[1],
					bid_reciprocals10_128[extra_digits].w[1], cy)
				Qh = __shr_128_long(Qh1, uint(128-amount))
				Tmp.w[0] = 1
				Tmp.w[1] = 0
				Tmp1 = __shl_128_long(Tmp, uint(amount))
				Qh.w[0] += carry
				if Qh.w[0] < carry {
					Qh.w[1]++
				}
				if __unsigned_compare_ge_128(Qh, Tmp1) {
					status = BID_EXACT_STATUS
				}
			}

			if status != BID_EXACT_STATUS {
				if uf_check != 0 {
					status |= BID_UNDERFLOW_EXCEPTION
				}
				pfpsf |= status
			}
		}

		_ = CX1
		_ = Stemp
	}

	res = get_BID32_flags(uint32(sign_x>>32),
		exponent_x-DECIMAL_EXPONENT_BIAS_128+
			DECIMAL_EXPONENT_BIAS_32, CX.w[0], rnd_mode, &pfpsf)
	return res, pfpsf
}
