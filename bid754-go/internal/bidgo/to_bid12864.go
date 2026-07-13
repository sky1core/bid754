package bidgo

import "math/bits"

// Bid64ToBid128 converts a BID64 to BID128 and returns status flags.
// Ported mechanically from Intel bid64_to_bid128.c.
func Bid64ToBid128(x uint64) (BID_UINT128, uint32) {
	var new_coeff, res BID_UINT128
	var sign_x uint64
	var exponent_x int
	var coefficient_x uint64
	var pfpsf uint32

	sign_x, exponent_x, coefficient_x, valid := unpack_BID64(x)
	if !valid {
		if (x << 1) >= 0xf000000000000000 {
			if (x & SNAN_MASK64) == SNAN_MASK64 { // sNaN
				pfpsf |= BID_INVALID_EXCEPTION
			}
			res.lo = coefficient_x & 0x0003ffffffffffff
			res.hi, res.lo = bits.Mul64(res.lo, bid_power10_table_128[18].lo)
			res.hi |= coefficient_x & 0xfc00000000000000
			return res, pfpsf
		}
	}

	new_coeff.lo = coefficient_x
	new_coeff.hi = 0
	res.lo = new_coeff.lo
	res.hi = sign_x | (uint64(exponent_x+6176-398) << 49)
	return res, pfpsf
}
