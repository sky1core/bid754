// Ported from: Intel bid32_sqrt.c
// Mechanical translation - all logic preserved exactly.

package bidgo

import "math"

// Bid32Sqrt is ported mechanically from bid32_sqrt.c: bid32_sqrt.
func Bid32Sqrt(x uint32, rnd_mode int) (uint32, uint32) {
	var CA, CT uint64
	var sign_x, coefficient_x uint32
	var Q, A10, QE, res uint32
	var dq, dqe float64
	var exponent_x, exponent_q, bin_expon_cx int
	var digits_x int
	var scale int
	var pfpsf uint32

	sign_x, exponent_x, coefficient_x, valid := unpack_BID32(x)
	if coefficient_x == 0 {
		valid = false
	}
	if !valid {
		if (x & INFINITY_MASK32) == INFINITY_MASK32 {
			res = coefficient_x
			if (coefficient_x & SSNAN_MASK32) == SINFINITY_MASK32 {
				res = NAN_MASK32
				pfpsf |= BID_INVALID_EXCEPTION
			}
			if (x & SNAN_MASK32) == SNAN_MASK32 {
				pfpsf |= BID_INVALID_EXCEPTION
			}
			return res & QUIET_MASK32, pfpsf
		}
		exponent_x = (exponent_x + DECIMAL_EXPONENT_BIAS_32) >> 1
		res = sign_x | (uint32(exponent_x) << 23)
		return res, pfpsf
	}
	if sign_x != 0 && coefficient_x != 0 {
		res = NAN_MASK32
		pfpsf |= BID_INVALID_EXCEPTION
		return res, pfpsf
	}

	tempx := math.Float32bits(float32(coefficient_x))
	bin_expon_cx = int((tempx>>23)&0xff) - 0x7f
	digits_x = bid_estimate_decimal_digits[bin_expon_cx]
	if uint64(coefficient_x) >= uint64(bid_power10_index_binexp[bin_expon_cx]) {
		digits_x++
	}

	A10 = coefficient_x
	if (exponent_x & 1) == 0 {
		A10 = (A10 << 2) + A10
		A10 += A10
	}

	dqe = math.Sqrt(float64(A10))
	QE = uint32(dqe)
	if QE*QE == A10 {
		res = very_fast_get_BID32(0, (exponent_x+DECIMAL_EXPONENT_BIAS_32)>>1, QE)
		return res, pfpsf
	}

	scale = 13 - digits_x
	exponent_q = exponent_x + DECIMAL_EXPONENT_BIAS_32 - scale
	scale += (exponent_q & 1)

	CT = bid_power10_table_128[scale].w[0]
	CA = uint64(coefficient_x) * CT

	dq = math.Sqrt(float64(CA))

	exponent_q = (exponent_q) >> 1

	pfpsf |= BID_INEXACT_EXCEPTION

	if (rnd_mode & 3) == 0 {
		Q = uint32(dq + 0.5)
	} else {
		Q = uint32(dq)
		if rnd_mode == BID_ROUNDING_UP {
			Q++
		}
	}

	res = fast_get_BID32(0, exponent_q, Q)
	return res, pfpsf
}
