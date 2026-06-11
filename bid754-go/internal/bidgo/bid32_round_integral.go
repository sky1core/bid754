// Ported from: Intel bid32_round_integral.c
// Mechanical translation - Intel implements bid32 round_integral via bid64.

package bidgo

// bid32_to_bid64 converts BID32 to BID64 (lossless widening).
func bid32_to_bid64_local(x uint32) uint64 {
	sign, exp, coeff, valid := unpack_BID32(x)
	if !valid {
		if (x & NAN_MASK32) == NAN_MASK32 {
			// NaN
			payload := uint64(x & 0x000fffff)
			if (x & SNAN_MASK32) == SNAN_MASK32 {
				return uint64(sign)<<32 | 0x7e00000000000000 | payload
			}
			return uint64(sign)<<32 | 0x7c00000000000000 | payload
		}
		if (x & INFINITY_MASK32) == INFINITY_MASK32 {
			return uint64(sign)<<32 | 0x7800000000000000
		}
		// zero
		exp64 := exp + (DECIMAL_EXPONENT_BIAS - DECIMAL_EXPONENT_BIAS_32)
		return uint64(sign)<<32 | (uint64(exp64) << 53)
	}
	exp64 := exp + (DECIMAL_EXPONENT_BIAS - DECIMAL_EXPONENT_BIAS_32)
	return uint64(sign)<<32 | (uint64(exp64) << 53) | uint64(coeff)
}

// bid64_to_bid32_local converts BID64 to BID32 (with rounding).
func bid64_to_bid32_local(x uint64) uint32 {
	r, _ := Bid64ToBid32(x, 0)
	return r
}

// Bid32RoundIntegralExact is ported mechanically from bid32_round_integral.c.
func Bid32RoundIntegralExact(x uint32, rndMode int) (uint32, uint32) {
	x64, flags0 := Bid32ToBid64(x)
	res64, flags1 := Bid64RoundIntegralExact(x64, rndMode)
	res, flags2 := Bid64ToBid32(res64, 0)
	return res, flags0 | flags1 | flags2
}

// Bid32RoundIntegralNearestEven is ported mechanically from bid32_round_integral.c.
func Bid32RoundIntegralNearestEven(x uint32) (uint32, uint32) {
	x64, flags0 := Bid32ToBid64(x)
	res64, flags1 := Bid64RoundIntegralNearestEven(x64)
	res, flags2 := Bid64ToBid32(res64, 0)
	return res, flags0 | flags1 | flags2
}

// Bid32RoundIntegralNegative is ported mechanically from bid32_round_integral.c.
func Bid32RoundIntegralNegative(x uint32) (uint32, uint32) {
	x64, flags0 := Bid32ToBid64(x)
	res64, flags1 := Bid64RoundIntegralNegative(x64)
	res, flags2 := Bid64ToBid32(res64, 0)
	return res, flags0 | flags1 | flags2
}

// Bid32RoundIntegralPositive is ported mechanically from bid32_round_integral.c.
func Bid32RoundIntegralPositive(x uint32) (uint32, uint32) {
	x64, flags0 := Bid32ToBid64(x)
	res64, flags1 := Bid64RoundIntegralPositive(x64)
	res, flags2 := Bid64ToBid32(res64, 0)
	return res, flags0 | flags1 | flags2
}

// Bid32RoundIntegralZero is ported mechanically from bid32_round_integral.c.
func Bid32RoundIntegralZero(x uint32) (uint32, uint32) {
	x64, flags0 := Bid32ToBid64(x)
	res64, flags1 := Bid64RoundIntegralZero(x64)
	res, flags2 := Bid64ToBid32(res64, 0)
	return res, flags0 | flags1 | flags2
}

// Bid32RoundIntegralNearestAway is ported mechanically from bid32_round_integral.c.
func Bid32RoundIntegralNearestAway(x uint32) (uint32, uint32) {
	x64, flags0 := Bid32ToBid64(x)
	res64, flags1 := Bid64RoundIntegralNearestAway(x64)
	res, flags2 := Bid64ToBid32(res64, 0)
	return res, flags0 | flags1 | flags2
}
