// Ported from: Intel bid32_scalb.c
// Mechanical translation - all logic preserved exactly.

package bidgo

// Bid32Scalbn is ported mechanically from bid32_scalb.c: bid32_scalbn.
func Bid32Scalbn(x uint32, n int, rnd_mode int) (uint32, uint32) {
	var sign_x, coefficient_x, res uint32
	var exp64 int64
	var exponent_x, rmode int
	var pfpsf uint32

	sign_x, exponent_x, coefficient_x, valid := unpack_BID32(x)
	if !valid {
		if (x & SNAN_MASK32) == SNAN_MASK32 {
			pfpsf |= BID_INVALID_EXCEPTION
		}
		if coefficient_x != 0 {
			res = coefficient_x & QUIET_MASK32
		} else {
			exp64 = int64(exponent_x) + int64(n)
			if exp64 < 0 {
				exp64 = 0
			}
			if exp64 > int64(DECIMAL_MAX_EXPON_32) {
				exp64 = int64(DECIMAL_MAX_EXPON_32)
			}
			exponent_x = int(exp64)
			res = very_fast_get_BID32(sign_x, exponent_x, coefficient_x)
		}
		return res, pfpsf
	}

	exp64 = int64(exponent_x) + int64(n)
	exponent_x = int(exp64)

	if uint32(exponent_x) <= DECIMAL_MAX_EXPON_32 {
		res = very_fast_get_BID32(sign_x, exponent_x, coefficient_x)
		return res, pfpsf
	}
	if exp64 > int64(DECIMAL_MAX_EXPON_32) {
		for coefficient_x < 1000000 && exp64 > int64(DECIMAL_MAX_EXPON_32) {
			coefficient_x = (coefficient_x << 1) + (coefficient_x << 3)
			exponent_x--
			exp64--
		}
		if exp64 <= int64(DECIMAL_MAX_EXPON_32) {
			res = very_fast_get_BID32(sign_x, exponent_x, coefficient_x)
			return res, pfpsf
		} else {
			exponent_x = 0x7fffffff
		}
	}
	rmode = rnd_mode
	res = get_BID32(sign_x, exponent_x, uint64(coefficient_x), rmode)
	return res, pfpsf
}
