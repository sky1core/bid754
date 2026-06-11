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

	da_h = float64(CA.w[1])
	da_l = float64(CA.w[0])
	da = noFmaMulAddF64(da_h, math.Float64frombits(t_scale), da_l)

	dq = math.Sqrt(da)
	Q = uint64(dq)

	R = uint64(int64(CA.w[0]-Q*Q) >> 63)
	D = int64(R + R + 1)

	exponent_q = (exponent_q + DECIMAL_EXPONENT_BIAS) >> 1

	pfpsf |= BID_INEXACT_EXCEPTION

	if (rndMode & 3) == 0 {
		Q2 = Q + Q + uint64(D)
		C4 = CA.w[0] << 2
		R2 = uint64(int64(Q2*Q2-C4) >> 63)
		Q += uint64(D) & (R ^ R2)
	} else {
		C4 = CA.w[0]
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
