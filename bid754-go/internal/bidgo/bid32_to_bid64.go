// Ported from: Intel bid32_to_bid64.c
// Mechanical translation - all logic preserved exactly.

package bidgo

// Bid32ToBid64 is ported mechanically from bid32_to_bid64.c: bid32_to_bid64.
func Bid32ToBid64(x uint32) (uint64, uint32) {
	var res uint64
	var sign_x, coefficient_x uint32
	var exponent_x int
	var pfpsf uint32

	sign_x, exponent_x, coefficient_x, valid := unpack_BID32(x)
	if !valid {
		if (x & 0x78000000) == 0x78000000 {
			if (x & 0x7e000000) == 0x7e000000 {
				pfpsf |= BID_INVALID_EXCEPTION
			}
			res = uint64(coefficient_x & 0x000fffff)
			res *= 1000000000
			res |= ((uint64(coefficient_x)) << 32) & 0xfc00000000000000
			return res, pfpsf
		}
	}

	res = very_fast_get_BID64_small_mantissa(uint64(sign_x)<<32,
		exponent_x+DECIMAL_EXPONENT_BIAS-DECIMAL_EXPONENT_BIAS_32,
		uint64(coefficient_x))
	return res, pfpsf
}
