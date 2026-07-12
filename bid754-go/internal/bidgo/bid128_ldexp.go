// Ported from: Intel bid128_ldexp.c
// Mechanical translation - all logic preserved exactly.

package bidgo

// Use EXPONENT_BIAS128 and DECIMAL_MAX_EXPON_128 from bid128_internal.go

// Bid128Ldexp returns x * 10^n.
// Ported mechanically from Intel bid128_ldexp.c: bid128_ldexp.
func Bid128Ldexp(x BID_UINT128, n int, rnd_mode int) (BID_UINT128, uint32) {
	var CX, CX2, CBID_X8, res BID_UINT128
	var exp64 int64
	var sign_x uint64
	var exponent_x int
	var pfpsf uint32

	// unpack arguments, check for NaN or Infinity
	sign_x, exponent_x, CX, valid := unpack_BID128_value(x)
	if !valid {
		// x is Inf. or NaN or 0
		if (x.w[1] & SNAN_MASK64) == SNAN_MASK64 { // y is sNaN
			pfpsf |= BID_INVALID_EXCEPTION
		}
		res.w[1] = CX.w[1] & QUIET_MASK64
		res.w[0] = CX.w[0]
		if CX.w[1] == 0 {
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
		return res, pfpsf
	}

	exp64 = int64(exponent_x) + int64(n)
	exponent_x = int(exp64)

	if uint32(exponent_x) <= DECIMAL_MAX_EXPON_128 {
		res = very_fast_get_BID128(sign_x, exponent_x, CX)
		return res, pfpsf
	}
	// check for overflow
	if exp64 > DECIMAL_MAX_EXPON_128 {
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

				if !(CX.w[1] < 0x314dc6448d93 && exp64 > DECIMAL_MAX_EXPON_128) {
					break
				}
			}
		}
		if exp64 <= DECIMAL_MAX_EXPON_128 {
			res = very_fast_get_BID128(sign_x, exponent_x, CX)
			return res, pfpsf
		} else {
			exponent_x = 0x7fffffff // overflow
		}
	}
	// exponent < 0
	// the BID pack routine will round the coefficient
	rmode := rnd_mode
	res = bid_get_BID128(sign_x, exponent_x, CX, rmode, &pfpsf)
	return res, pfpsf
}
