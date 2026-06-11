package bidgo

import "math/big"

func bid128NaNToBid64(hi, lo uint64) (uint64, uint32) {
	payloadHi := hi & 0x00003fffffffffff
	payloadLo := lo
	t33hi := uint64(0x0000314dc6448d93)
	t33lo := uint64(0x38c15b09ffffffff)
	if payloadHi > t33hi || (payloadHi == t33hi && payloadLo > t33lo) {
		payloadHi = 0
		payloadLo = 0
	}
	payload := bid128CoeffBig(payloadHi, payloadLo)
	payload.Quo(payload, big.NewInt(1000000000000000000))
	return (hi & 0xfc00000000000000) | payload.Uint64(), func() uint32 {
		if (hi & 0x7e00000000000000) == 0x7e00000000000000 {
			return BID_INVALID_EXCEPTION
		}
		return 0
	}()
}

func bid64CanonicalizeNonCanonicalFinite(x uint64) uint64 {
	if (x & MASK_INF64) == MASK_INF64 {
		return x
	}
	if (x & MASK_STEERING_BITS64) == MASK_STEERING_BITS64 {
		if ((x & MASK_BINARY_SIG2_64) | MASK_BINARY_OR2_64) > 9999999999999999 {
			return (x & MASK_SIGN64) | ((x & MASK_BINARY_EXPONENT2_64) << 2)
		}
	}
	return x
}

func bid64DecodeForCompare(x uint64) (sign uint64, exp int, coeff *big.Int, isZero bool) {
	sign, exp, c := bid64UnpackFiniteForRoundLocal(x)
	coeff = new(big.Int).SetUint64(c)
	return sign, exp, coeff, c == 0
}

func bid64CompareToBid128(x uint64, y BID_UINT128) int {
	x = bid64CanonicalizeNonCanonicalFinite(x)
	xSign, xExp, xCoeff, xZero := bid64DecodeForCompare(x)
	yd := bid128Decode(y.w[1], y.w[0])
	if Bid64IsInf(x) != 0 {
		if yd.isInf {
			if xSign == yd.sign {
				return 0
			}
			if xSign != 0 {
				return -1
			}
			return 1
		}
		if xSign != 0 {
			return -1
		}
		return 1
	}
	if yd.isInf {
		if yd.sign != 0 {
			return 1
		}
		return -1
	}
	if xZero && yd.isZero {
		return 0
	}
	if xSign != yd.sign {
		if xZero && yd.isZero {
			return 0
		}
		if xSign != 0 {
			return -1
		}
		return 1
	}
	xc := new(big.Int).Set(xCoeff)
	yc := new(big.Int).Set(yd.coeff)
	if xExp > yd.exp {
		xc.Mul(xc, bid128Pow10Big(xExp-yd.exp))
	} else if yd.exp > xExp {
		yc.Mul(yc, bid128Pow10Big(yd.exp-xExp))
	}
	cmp := xc.Cmp(yc)
	if xSign != 0 {
		cmp = -cmp
	}
	return cmp
}

// Bid64NextToward is ported mechanically from Intel bid64_nexttowardd.c: bid64_nexttoward.
func Bid64NextToward(x uint64, y BID_UINT128) (uint64, uint32) {
	var res uint64
	var tmp1, tmp2 uint64
	var pfpsf uint32
	var res1, res2 int

	yd := bid128Decode(y.w[1], y.w[0])

	// check for NaNs or infinities
	if (x & MASK_NAN) == MASK_NAN { // x is NAN
		if (x & 0x0003ffffffffffff) > 999999999999999 {
			x = x & 0xfe00000000000000 // clear G6-G12 and the payload bits
		} else {
			x = x & 0xfe03ffffffffffff // clear G6-G12
		}
		if (x & MASK_SNAN64) == MASK_SNAN64 { // x is SNAN
			// set invalid flag
			pfpsf |= BID_INVALID_EXCEPTION
			// return quiet (x)
			res = x & 0xfdffffffffffffff
		} else { // x is QNaN
			if yd.isSNaN { // y is SNAN
				// set invalid flag
				pfpsf |= BID_INVALID_EXCEPTION
			}
			// return x
			res = x
		}
		return res, pfpsf
	} else if yd.isNaN { // y is NAN then res = Q (y)
		res, pfpsf = bid128NaNToBid64(y.w[1], y.w[0])
		return res, pfpsf
	} else { // at least one is infinity
		if (x & MASK_INF) == MASK_INF { // x = inf
			x = x & (MASK_SIGN | MASK_INF)
		}
	}
	// neither x nor y is NaN

	// if not infinity, check for non-canonical values x (treated as zero)
	if (x & MASK_INF) != MASK_INF { // x != inf
		x = bid64CanonicalizeNonCanonicalFinite(x)
	}
	// no need to check for non-canonical y

	// neither x nor y is NaN
	res2 = bid64CompareToBid128(x, y)
	if res2 == 0 { // x = y
		// return x with the sign of y
		res = (y.w[1] & MASK_SIGN) | (x & 0x7fffffffffffffff)
	} else if res2 > 0 { // x > y
		res, _ = Bid64NextDown(x)
	} else { // x < y
		res, _ = Bid64NextUp(x)
	}
	// if the operand x is finite but the result is infinite, signal
	// overflow and inexact
	if ((x & MASK_INF) != MASK_INF) && ((res & MASK_INF) == MASK_INF) {
		// set the inexact flag
		pfpsf |= BID_INEXACT_EXCEPTION
		// set the overflow flag
		pfpsf |= BID_OVERFLOW_EXCEPTION
	}
	// if the result is in (-10^emin, 10^emin), and is different from the
	// operand x, signal underflow and inexact
	tmp1 = 0x00038d7ea4c68000 // +100...0[16] * 10^emin
	tmp2 = res & 0x7fffffffffffffff
	res1, _ = Bid64QuietGreater(tmp1, tmp2)
	res2, _ = Bid64QuietNotEqual(x, res)
	if res1 != 0 && res2 != 0 {
		// set the inexact flag
		pfpsf |= BID_INEXACT_EXCEPTION
		// set the underflow flag
		pfpsf |= BID_UNDERFLOW_EXCEPTION
	}
	return res, pfpsf
}
