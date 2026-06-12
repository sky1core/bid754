// Ported from: IntelRDFPMathLib20U4/LIBRARY/src/bid64_modf.c
// (the local unpack helper follows the unpack block in bid64_round_integral.c)
// Version: Intel(R) Decimal Floating-Point Math Library 2.0 Update 4
//
// This file is a mechanical translation of the Intel BID library to Go.
// All logic, magic numbers, and table references are preserved exactly.

package bidgo

func bid64UnpackFiniteForRoundLocal(x uint64) (uint64, int, uint64) {
	xSign := x & 0x8000000000000000
	if (x & MASK_STEERING_BITS64) == MASK_STEERING_BITS64 {
		exp := int((x&MASK_BINARY_EXPONENT2_64)>>51) - 398
		coeff := (x & MASK_BINARY_SIG2_64) | MASK_BINARY_OR2_64
		if coeff > 9999999999999999 {
			coeff = 0
		}
		return xSign, exp, coeff
	}
	exp := int((x&MASK_BINARY_EXPONENT1_64)>>53) - 398
	coeff := x & MASK_BINARY_SIG1_64
	return xSign, exp, coeff
}

// Bid64Modf splits x into fractional and integral parts and returns status flags.
func Bid64Modf(x uint64) (uint64, uint64, uint32) {
	xi, flags := Bid64RoundIntegralZero(x)
	var res uint64

	if (x & 0x7c00000000000000) == 0x7800000000000000 {
		res = (x & 0x8000000000000000) | 0x5fe0000000000000
	} else {
		r, subFlags := Bid64SubWithFlags(x, xi, 0)
		res = r
		flags |= subFlags
	}

	iptr := xi | (x & 0x8000000000000000)
	res |= x & 0x8000000000000000

	return res, iptr, flags
}
