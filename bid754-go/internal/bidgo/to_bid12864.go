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
			res.w[0] = coefficient_x & 0x0003ffffffffffff
			res.w[1], res.w[0] = bits.Mul64(res.w[0], bid_power10_table_128[18].w[0])
			res.w[1] |= coefficient_x & 0xfc00000000000000
			return res, pfpsf
		}
	}

	new_coeff.w[0] = coefficient_x
	new_coeff.w[1] = 0
	res.w[0] = new_coeff.w[0]
	res.w[1] = sign_x | (uint64(exponent_x+6176-398) << 49)
	return res, pfpsf
}
