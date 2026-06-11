// Ported from: Intel bid128_modf.c
// Mechanical translation - all logic preserved exactly.

package bidgo

// Bid128Modf splits x into integer and fractional parts.
// Returns fractional part as result, integer part via pint.
// Ported mechanically from Intel bid128_modf.c: bid128_modf.
func Bid128Modf(x BID_UINT128) (BID_UINT128, BID_UINT128, uint32) {
	var res, xi BID_UINT128
	var pfpsf uint32

	xi = Bid128RoundIntegralZero(x, &pfpsf)

	// check for Infinity
	if (x.w[1] & 0x7c00000000000000) == 0x7800000000000000 {
		res.w[1] = (x.w[1] & 0x8000000000000000) | 0x5ffe000000000000
		res.w[0] = 0
	} else {
		res = Bid128Sub(x, xi, BID_ROUNDING_TO_NEAREST, &pfpsf)
	}

	xi.w[1] |= (x.w[1] & 0x8000000000000000)
	res.w[1] |= (x.w[1] & 0x8000000000000000)

	return res, xi, pfpsf
}
