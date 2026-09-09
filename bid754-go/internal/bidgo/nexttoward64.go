// Ported from: IntelRDFPMathLib20U4/LIBRARY/src/bid64_nexttowardd.c
// (the BID128 NaN narrowing follows bid128_to_bid64 in bid64_to_bid128.c)
// Version: Intel(R) Decimal Floating-Point Math Library 2.0 Update 4

package bidgo

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

// Bid64NextToward is ported mechanically from Intel bid64_nexttowardd.c: bid64_nexttoward.
func Bid64NextToward(x uint64, y BID_UINT128) (uint64, uint32) {
	var res uint64
	var tmp1, tmp2 uint64
	var pfpsf uint32
	var res1, res2 int

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
			if (y.hi & MASK_SNAN64) == MASK_SNAN64 { // y is SNAN
				// set invalid flag
				pfpsf |= BID_INVALID_EXCEPTION
			}
			// return x
			res = x
		}
		return res, pfpsf
	} else if (y.hi & MASK_NAN) == MASK_NAN { // y is NAN then res = Q (y)
		if (y.hi & MASK_SNAN64) == MASK_SNAN64 {
			pfpsf |= BID_INVALID_EXCEPTION
		}
		if (y.hi&0x00003fffffffffff) > 0x0000314dc6448d93 ||
			((y.hi&0x00003fffffffffff) == 0x0000314dc6448d93 && y.lo > 0x38c15b09ffffffff) {
			y.hi &= 0xffffc00000000000
			y.lo = 0
		}
		y.hi &= 0xfc003fffffffffff
		res, _ = Bid128ToBid64(y, BID_ROUNDING_TO_NEAREST)
		return res, pfpsf
	} else { // at least one is infinity
		if (x & MASK_INF) == MASK_INF { // x = inf
			x = x & (MASK_SIGN | MASK_INF)
		}
		if (y.hi & MASK_INF) == MASK_INF {
			y.hi &= MASK_SIGN | MASK_INF
			y.lo = 0
		}
	}
	// neither x nor y is NaN

	// if not infinity, check for non-canonical values x (treated as zero)
	if (x & MASK_INF) != MASK_INF { // x != inf
		x = bid64CanonicalizeNonCanonicalFinite(x)
	}
	// no need to check for non-canonical y

	// neither x nor y is NaN
	x128, _ := Bid64ToBid128(x)
	res1, _ = Bid128QuietEqual(x128, y)
	res2, _ = Bid128QuietGreater(x128, y)
	if res1 != 0 { // x = y
		// return x with the sign of y
		res = (y.hi & MASK_SIGN) | (x & 0x7fffffffffffffff)
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
