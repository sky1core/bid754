package bidgo

// Bid64Scalbn scales x by 10^n and returns status flags.
// Ported mechanically from Intel bid64_scalb.c.
func Bid64Scalbn(x uint64, n int, rndMode int) (uint64, uint32) {
	var sign_x, coefficient_x, res uint64
	var exp64 int64
	var exponent_x int
	var pfpsf uint32

	// unpack arguments, check for NaN or Infinity
	sign_x, exponent_x, coefficient_x, valid_x := unpack_BID64(x)
	if !valid_x {
		// x is Inf. or NaN or 0
		if (x & SNAN_MASK64) == SNAN_MASK64 {
			pfpsf |= BID_INVALID_EXCEPTION
		}
		if coefficient_x != 0 {
			res = coefficient_x & QUIET_MASK64
		} else {
			exp64 = int64(exponent_x) + int64(n)
			if exp64 < 0 {
				exp64 = 0
			}
			if exp64 > DECIMAL_MAX_EXPON_64 {
				exp64 = DECIMAL_MAX_EXPON_64
			}
			exponent_x = int(exp64)
			res = very_fast_get_BID64(sign_x, exponent_x, coefficient_x)
		}
		return res, pfpsf
	}

	exp64 = int64(exponent_x) + int64(n)
	exponent_x = int(exp64)

	if uint32(exponent_x) <= DECIMAL_MAX_EXPON_64 {
		res = very_fast_get_BID64(sign_x, exponent_x, coefficient_x)
		return res, pfpsf
	}
	// check for overflow
	if exp64 > DECIMAL_MAX_EXPON_64 {
		// try to normalize coefficient
		for coefficient_x < 1000000000000000 && exp64 > DECIMAL_MAX_EXPON_64 {
			// coefficient_x < 10^15, scale by 10
			coefficient_x = (coefficient_x << 1) + (coefficient_x << 3)
			exponent_x--
			exp64--
		}
		if exp64 <= DECIMAL_MAX_EXPON_64 {
			res = very_fast_get_BID64(sign_x, exponent_x, coefficient_x)
			return res, pfpsf
		}
		exponent_x = 0x7fffffff // overflow
	}
	// exponent < 0
	// the BID pack routine will round the coefficient
	res, pfpsf = get_BID64_flags(sign_x, exponent_x, coefficient_x, rndMode)
	return res, pfpsf
}

// Bid64Scalbln scales x by 10^n and returns status flags.
// Ported mechanically from Intel bid64_scalbl.c.
func Bid64Scalbln(x uint64, n int64, rndMode int) (uint64, uint32) {
	var res uint64
	var n1 int32

	n1 = int32(n)
	if int64(n1) < n {
		n1 = 0x7fffffff
	} else if int64(n1) > n {
		n1 = -0x7fffffff - 1
	}

	res, pfpsf := Bid64Scalbn(x, int(n1), rndMode)
	return res, pfpsf
}

// Bid64Ldexp scales x by 10^n and returns status flags.
// Ported mechanically from Intel bid64_ldexp.c.
func Bid64Ldexp(x uint64, n int, rndMode int) (uint64, uint32) {
	var sign_x, coefficient_x, res uint64
	var exp64 int64
	var exponent_x, rmode int
	var pfpsf uint32

	// unpack arguments, check for NaN or Infinity
	sign_x, exponent_x, coefficient_x, valid_x := unpack_BID64(x)
	if !valid_x {
		// x is Inf. or NaN or 0
		if (x & SNAN_MASK64) == SNAN_MASK64 {
			pfpsf |= BID_INVALID_EXCEPTION
		}
		if coefficient_x != 0 {
			res = coefficient_x & QUIET_MASK64
		} else {
			exp64 = int64(exponent_x) + int64(n)
			if exp64 < 0 {
				exp64 = 0
			}
			if exp64 > DECIMAL_MAX_EXPON_64 {
				exp64 = DECIMAL_MAX_EXPON_64
			}
			exponent_x = int(exp64)
			res = very_fast_get_BID64(sign_x, exponent_x, coefficient_x) // 0
		}
		return res, pfpsf
	}

	exp64 = int64(exponent_x) + int64(n)
	exponent_x = int(exp64)

	if uint32(exponent_x) <= DECIMAL_MAX_EXPON_64 {
		res = very_fast_get_BID64(sign_x, exponent_x, coefficient_x)
		return res, pfpsf
	}
	// check for overflow
	if exp64 > DECIMAL_MAX_EXPON_64 {
		// try to normalize coefficient
		for coefficient_x < 1000000000000000 && exp64 > DECIMAL_MAX_EXPON_64 {
			// coefficient_x < 10^15, scale by 10
			coefficient_x = (coefficient_x << 1) + (coefficient_x << 3)
			exponent_x--
			exp64--
		}
		if exp64 <= DECIMAL_MAX_EXPON_64 {
			res = very_fast_get_BID64(sign_x, exponent_x, coefficient_x)
			return res, pfpsf
		}
		exponent_x = 0x7fffffff // overflow
	}
	// exponent < 0
	// the BID pack routine will round the coefficient
	rmode = rndMode
	res, pfpsf = get_BID64_flags(sign_x, exponent_x, coefficient_x, rmode)
	return res, pfpsf
}
