package bidgo

// Bid64Frexp returns res and exp such that x = res * 10^exp.
// Ported mechanically from Intel bid64_frexp.c.
func Bid64Frexp(x uint64) (uint64, int) {
	var res uint64
	var sig_x, exp_x uint64
	var C BID_UINT128
	var q int

	if (x&MASK_NAN64) == MASK_NAN64 || (x&MASK_INF64) == MASK_INF64 {
		// if NaN or infinity
		res = x
		// the binary frexp quietetizes SNaNs, so do the same
		if (x & MASK_SNAN64) == MASK_SNAN64 { // x is SNAN
			// return quiet (x)
			res = x & 0xfdffffffffffffff
		}
		return res, 0
	} else {
		// x is 0, non-canonical, normal, or subnormal
		// unpack x
		// if steering bits are 11 (condition will be 0), then exponent is G[0:w+1]
		if (x & MASK_STEERING_BITS64) == MASK_STEERING_BITS64 {
			sig_x = (x & MASK_BINARY_SIG2_64) | MASK_BINARY_OR2_64
			exp_x = (x & MASK_BINARY_EXPONENT2_64) >> 51 // biased
			if sig_x > 9999999999999999 || sig_x == 0 {  // non-canonical or zero
				res = (x & 0x8000000000000000) | (exp_x << 53) // zero of same sign
				return res, 0
			}
		} else {
			sig_x = x & MASK_BINARY_SIG1_64
			exp_x = (x & MASK_BINARY_EXPONENT1_64) >> 53 // biased
			if sig_x == 0x0 {
				res = x // same zero
				return res, 0
			}
		}
		// x is normal or subnormal, with exp_x=biased exponent & sig_x=coefficient
		// determine the number of decimal digits in sig_x, which fits in 54 bits
		// q = nr. of decimal digits in sig_x (1 <= q <= 16)
		if sig_x >= 0x0020000000000000 { // x >= 2^53
			q = 16
		} else { // if x < 2^53
			C.w[0] = sig_x
			C.w[1] = 0
			q = __get_dec_digits64(C)
		}
		// Do not add trailing zeros if q < 16; leave sig_x with q digits
		if sig_x < 0x0020000000000000 { // sig_x < 2^53 (fits in 53 bits)
			res = (x & 0x801fffffffffffff) | (uint64(-q+398) << 53) // replace exp.
		} else { // sig_x fits in 54 bits, but not in 53
			res = (x & 0xe007ffffffffffff) | (uint64(-q+398) << 51) // replace exp.
		}
		return res, int(exp_x) - 398 + q
	}
}
