package bidgo

// bid64_noncomp.c 기계적 포팅
// Intel BID 라이브러리의 비계산 함수들

// 64-bit BID masks - Intel bid_internal.h에서 가져옴
const (
	MASK_SIGN64              uint64 = 0x8000000000000000
	MASK_INF64               uint64 = 0x7800000000000000
	MASK_NAN64               uint64 = 0x7c00000000000000
	MASK_SNAN64              uint64 = 0x7e00000000000000
	MASK_STEERING_BITS64     uint64 = 0x6000000000000000
	MASK_BINARY_EXPONENT1_64 uint64 = 0x7fe0000000000000
	MASK_BINARY_EXPONENT2_64 uint64 = 0x1ff8000000000000
	MASK_BINARY_SIG1_64      uint64 = 0x001fffffffffffff
	MASK_BINARY_SIG2_64      uint64 = 0x0007ffffffffffff
	MASK_BINARY_OR2_64       uint64 = 0x0020000000000000
)

// bid_mult_factor for subnormal detection
var bid_mult_factor = [16]uint64{
	1, 10, 100, 1000,
	10000, 100000, 1000000, 10000000,
	100000000, 1000000000, 10000000000, 100000000000,
	1000000000000, 10000000000000,
	100000000000000, 1000000000000000,
}

// Bid64IsSigned - returns 1 if x is negative, 0 otherwise
// Intel bid64_isSigned 기계적 포팅
func Bid64IsSigned(x uint64) int {
	if (x & MASK_SIGN64) == MASK_SIGN64 {
		return 1
	}
	return 0
}

// Bid64IsNaN - returns 1 if x is NaN, 0 otherwise
// Intel bid64_isNaN 기계적 포팅
func Bid64IsNaN(x uint64) int {
	if (x & MASK_NAN64) == MASK_NAN64 {
		return 1
	}
	return 0
}

// Bid64IsFinite - returns 1 if x is finite (not Inf or NaN), 0 otherwise
// Intel bid64_isFinite 기계적 포팅
func Bid64IsFinite(x uint64) int {
	if (x & MASK_INF64) != MASK_INF64 {
		return 1
	}
	return 0
}

// Bid64IsInf - returns 1 if x is infinity, 0 otherwise
// Intel bid64_isInf 기계적 포팅
func Bid64IsInf(x uint64) int {
	if (x&MASK_INF64) == MASK_INF64 && (x&MASK_NAN64) != MASK_NAN64 {
		return 1
	}
	return 0
}

// Bid64IsSignaling - returns 1 if x is signaling NaN, 0 otherwise
// Intel bid64_isSignaling 기계적 포팅
func Bid64IsSignaling(x uint64) int {
	if (x & MASK_SNAN64) == MASK_SNAN64 {
		return 1
	}
	return 0
}

// Bid64IsCanonical - returns 1 if x is canonical, 0 otherwise
// Intel bid64_isCanonical 기계적 포팅
func Bid64IsCanonical(x uint64) int {
	if (x & MASK_NAN64) == MASK_NAN64 { // NaN
		if x&0x01fc000000000000 != 0 {
			return 0
		} else if (x & 0x0003ffffffffffff) > 999999999999999 { // payload
			return 0
		} else {
			return 1
		}
	} else if (x & MASK_INF64) == MASK_INF64 {
		if x&0x03ffffffffffffff != 0 {
			return 0
		} else {
			return 1
		}
	} else if (x & MASK_STEERING_BITS64) == MASK_STEERING_BITS64 { // 54-bit coeff.
		if ((x & MASK_BINARY_SIG2_64) | MASK_BINARY_OR2_64) <= 9999999999999999 {
			return 1
		}
		return 0
	} else { // 53-bit coeff.
		return 1
	}
}

// Bid64IsZero - returns 1 if x is zero, 0 otherwise
// Intel bid64_isZero 기계적 포팅
func Bid64IsZero(x uint64) int {
	// if infinity or nan, return 0
	if (x & MASK_INF64) == MASK_INF64 {
		return 0
	} else if (x & MASK_STEERING_BITS64) == MASK_STEERING_BITS64 {
		// if steering bits are 11, then exponent is G[0:w+1]
		// sig_x = (x & MASK_BINARY_SIG2) | MASK_BINARY_OR2
		if ((x & MASK_BINARY_SIG2_64) | MASK_BINARY_OR2_64) > 9999999999999999 {
			return 1
		}
		return 0
	} else {
		if (x & MASK_BINARY_SIG1_64) == 0 {
			return 1
		}
		return 0
	}
}

// Bid64IsNormal - returns 1 if x is normal (not zero, NaN, subnormal, or infinity)
// Intel bid64_isNormal 기계적 포팅
func Bid64IsNormal(x uint64) int {
	var sig_x_prime BID_UINT128
	var sig_x uint64
	var exp_x uint32

	if (x & MASK_INF64) == MASK_INF64 { // x is either INF or NaN
		return 0
	}

	// decode number into exponent and significand
	if (x & MASK_STEERING_BITS64) == MASK_STEERING_BITS64 {
		sig_x = (x & MASK_BINARY_SIG2_64) | MASK_BINARY_OR2_64
		// check for zero or non-canonical
		if sig_x > 9999999999999999 || sig_x == 0 {
			return 0 // zero or non-canonical
		}
		exp_x = uint32((x & MASK_BINARY_EXPONENT2_64) >> 51)
	} else {
		sig_x = x & MASK_BINARY_SIG1_64
		if sig_x == 0 {
			return 0 // zero
		}
		exp_x = uint32((x & MASK_BINARY_EXPONENT1_64) >> 53)
	}

	// if exponent is less than -383, the number may be subnormal
	// if (exp_x - 398 = -383) the number may be subnormal
	if exp_x < 15 {
		sig_x_prime = __mul_64x64_to_128(sig_x, bid_mult_factor[exp_x])
		if sig_x_prime.w[1] == 0 && sig_x_prime.w[0] < 1000000000000000 {
			return 0 // subnormal
		}
		return 1 // normal
	}
	return 1 // normal
}

// Bid64IsSubnormal - returns 1 if x is subnormal, 0 otherwise
// Intel bid64_isSubnormal 기계적 포팅
func Bid64IsSubnormal(x uint64) int {
	var sig_x_prime BID_UINT128
	var sig_x uint64
	var exp_x uint32

	if (x & MASK_INF64) == MASK_INF64 { // x is either INF or NaN
		return 0
	}

	// decode number into exponent and significand
	if (x & MASK_STEERING_BITS64) == MASK_STEERING_BITS64 {
		sig_x = (x & MASK_BINARY_SIG2_64) | MASK_BINARY_OR2_64
		// check for zero or non-canonical
		if sig_x > 9999999999999999 || sig_x == 0 {
			return 0 // zero or non-canonical
		}
		exp_x = uint32((x & MASK_BINARY_EXPONENT2_64) >> 51)
	} else {
		sig_x = x & MASK_BINARY_SIG1_64
		if sig_x == 0 {
			return 0 // zero
		}
		exp_x = uint32((x & MASK_BINARY_EXPONENT1_64) >> 53)
	}

	// if exponent is less than -383, the number may be subnormal
	if exp_x < 15 {
		sig_x_prime = __mul_64x64_to_128(sig_x, bid_mult_factor[exp_x])
		if sig_x_prime.w[1] == 0 && sig_x_prime.w[0] < 1000000000000000 {
			return 1 // subnormal
		}
		return 0 // normal
	}
	return 0 // normal
}

// Bid64Copy - copies x to result
// Intel bid64_copy 기계적 포팅
func Bid64Copy(x uint64) uint64 {
	return x
}

// Bid64Negate - reverses the sign of x
// Intel bid64_negate 기계적 포팅
func Bid64Negate(x uint64) uint64 {
	return x ^ MASK_SIGN64
}

// Bid64Abs - returns absolute value of x
// Intel bid64_abs 기계적 포팅
func Bid64Abs(x uint64) uint64 {
	return x & ^MASK_SIGN64
}

// Bid64CopySign - copies x with sign of y
// Intel bid64_copySign 기계적 포팅
func Bid64CopySign(x, y uint64) uint64 {
	return (x & ^MASK_SIGN64) | (y & MASK_SIGN64)
}

// Bid64SameQuantum - returns 1 if x and y have the same quantum (exponent)
// Intel bid64_sameQuantum 기계적 포팅
func Bid64SameQuantum(x, y uint64) int {
	var exp_x, exp_y uint32

	// if both operands are NaN, return true; if just one is NaN, return false
	if (x&MASK_NAN64) == MASK_NAN64 || (y&MASK_NAN64) == MASK_NAN64 {
		if (x&MASK_NAN64) == MASK_NAN64 && (y&MASK_NAN64) == MASK_NAN64 {
			return 1
		}
		return 0
	}
	// if both operands are INF, return true; if just one is INF, return false
	if (x&MASK_INF64) == MASK_INF64 || (y&MASK_INF64) == MASK_INF64 {
		if (x&MASK_INF64) == MASK_INF64 && (y&MASK_INF64) == MASK_INF64 {
			return 1
		}
		return 0
	}
	// decode exponents for both numbers, and return true if they match
	if (x & MASK_STEERING_BITS64) == MASK_STEERING_BITS64 {
		exp_x = uint32((x & MASK_BINARY_EXPONENT2_64) >> 51)
	} else {
		exp_x = uint32((x & MASK_BINARY_EXPONENT1_64) >> 53)
	}
	if (y & MASK_STEERING_BITS64) == MASK_STEERING_BITS64 {
		exp_y = uint32((y & MASK_BINARY_EXPONENT2_64) >> 51)
	} else {
		exp_y = uint32((y & MASK_BINARY_EXPONENT1_64) >> 53)
	}
	if exp_x == exp_y {
		return 1
	}
	return 0
}

// Bid64Radix - returns 10 (the radix of decimal floating-point)
// Intel bid64_radix 기계적 포팅
func Bid64Radix() int {
	return 10
}

// class_t constants for Bid64Class
const (
	signalingNaN      = 0
	quietNaN          = 1
	negativeInfinity  = 2
	negativeNormal    = 3
	negativeSubnormal = 4
	negativeZero      = 5
	positiveZero      = 6
	positiveSubnormal = 7
	positiveNormal    = 8
	positiveInfinity  = 9
)

// Bid64Class - returns the class of x
// Intel bid64_class 기계적 포팅
func Bid64Class(x uint64) int {
	var sig_x_prime BID_UINT128
	var sig_x uint64
	var exp_x int

	if (x & MASK_NAN64) == MASK_NAN64 {
		// is the NaN signaling?
		if (x & MASK_SNAN64) == MASK_SNAN64 {
			return signalingNaN
		}
		// if NaN and not signaling, must be quietNaN
		return quietNaN
	} else if (x & MASK_INF64) == MASK_INF64 {
		// is the Infinity negative?
		if (x & MASK_SIGN64) == MASK_SIGN64 {
			return negativeInfinity
		}
		// otherwise, must be positive infinity
		return positiveInfinity
	} else if (x & MASK_STEERING_BITS64) == MASK_STEERING_BITS64 {
		// decode number into exponent and significand
		sig_x = (x & MASK_BINARY_SIG2_64) | MASK_BINARY_OR2_64
		// check for zero or non-canonical
		if sig_x > 9999999999999999 || sig_x == 0 {
			if (x & MASK_SIGN64) == MASK_SIGN64 {
				return negativeZero
			}
			return positiveZero
		}
		exp_x = int((x & MASK_BINARY_EXPONENT2_64) >> 51)
	} else {
		sig_x = x & MASK_BINARY_SIG1_64
		if sig_x == 0 {
			if (x & MASK_SIGN64) == MASK_SIGN64 {
				return negativeZero
			}
			return positiveZero
		}
		exp_x = int((x & MASK_BINARY_EXPONENT1_64) >> 53)
	}

	// if exponent is less than -383, number may be subnormal
	if exp_x < 15 { // sig_x * 10^exp_x
		sig_x_prime = __mul_64x64_to_128(sig_x, bid_mult_factor[exp_x])
		if sig_x_prime.w[1] == 0 && sig_x_prime.w[0] < 1000000000000000 {
			if (x & MASK_SIGN64) == MASK_SIGN64 {
				return negativeSubnormal
			}
			return positiveSubnormal
		}
	}
	// otherwise, normal number, determine the sign
	if (x & MASK_SIGN64) == MASK_SIGN64 {
		return negativeNormal
	}
	return positiveNormal
}

// Bid64Inf - returns positive infinity
func Bid64Inf() uint64 {
	return 0x7800000000000000
}

// Bid64NaN - returns quiet NaN
func Bid64NaN() uint64 {
	return 0x7c00000000000000
}

// Bid64TotalOrder - returns 1 if x <= y in total order, 0 otherwise
// Intel bid64_totalOrder 기계적 포팅
func Bid64TotalOrder(x, y uint64) int {
	var res int
	var exp_x, exp_y int
	var sig_x, sig_y, pyld_y, pyld_x uint64
	var sig_n_prime BID_UINT128
	var x_is_zero, y_is_zero int

	// NaN (CASE1)
	if (x & MASK_NAN64) == MASK_NAN64 {
		// if x is -NaN
		if (x & MASK_SIGN64) == MASK_SIGN64 {
			// return true, unless y is -NaN also
			if (y&MASK_NAN64) != MASK_NAN64 || (y&MASK_SIGN64) != MASK_SIGN64 {
				res = 1
				return res
			}

			// if y and x are both -NaN
			xIsSNaN := (x & MASK_SNAN64) == MASK_SNAN64
			yIsSNaN := (y & MASK_SNAN64) == MASK_SNAN64
			if xIsSNaN == yIsSNaN {
				// compare payloads: for -NaN, larger payload is less
				pyld_y = y & 0x0003ffffffffffff
				pyld_x = x & 0x0003ffffffffffff
				if pyld_y > 999999999999999 || pyld_y == 0 {
					res = 1
					return res
				}
				if pyld_x > 999999999999999 || pyld_x == 0 {
					res = 0
					return res
				}
				if pyld_x >= pyld_y {
					res = 1
				} else {
					res = 0
				}
				return res
			}

			// totalOrder(-qNaN, -sNaN) == 1
			if (y & MASK_SNAN64) == MASK_SNAN64 {
				res = 1
			} else {
				res = 0
			}
			return res
		}

		// x is +NaN
		if (y&MASK_NAN64) != MASK_NAN64 || (y&MASK_SIGN64) == MASK_SIGN64 {
			res = 0
			return res
		}

		// x and y are both +NaN
		xIsSNaN := (x & MASK_SNAN64) == MASK_SNAN64
		yIsSNaN := (y & MASK_SNAN64) == MASK_SNAN64
		if xIsSNaN == yIsSNaN {
			// compare payloads: for +NaN, smaller payload is less
			pyld_y = y & 0x0003ffffffffffff
			pyld_x = x & 0x0003ffffffffffff
			if pyld_x > 999999999999999 || pyld_x == 0 {
				res = 1
				return res
			}
			if pyld_y > 999999999999999 || pyld_y == 0 {
				res = 0
				return res
			}
			if pyld_x <= pyld_y {
				res = 1
			} else {
				res = 0
			}
			return res
		}

		// totalOrder(+sNaN, +qNaN) == 1
		if (x & MASK_SNAN64) == MASK_SNAN64 {
			res = 1
		} else {
			res = 0
		}
		return res
	} else if (y & MASK_NAN64) == MASK_NAN64 {
		if (y & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		} else {
			res = 0
		}
		return res
	}

	// SIMPLE (CASE2)
	if x == y {
		res = 1
		return res
	}

	// OPPOSITE SIGNS (CASE3)
	if ((x & MASK_SIGN64) == MASK_SIGN64) != ((y & MASK_SIGN64) == MASK_SIGN64) {
		if (x & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		} else {
			res = 0
		}
		return res
	}

	// INFINITY (CASE4)
	if (x & MASK_INF64) == MASK_INF64 {
		if (x & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
			return res
		}
		if (y & MASK_INF64) == MASK_INF64 {
			res = 1
		} else {
			res = 0
		}
		return res
	} else if (y & MASK_INF64) == MASK_INF64 {
		if (y & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		} else {
			res = 0
		}
		return res
	}

	// Decode x
	if (x & MASK_STEERING_BITS64) == MASK_STEERING_BITS64 {
		exp_x = int((x & MASK_BINARY_EXPONENT2_64) >> 51)
		sig_x = (x & MASK_BINARY_SIG2_64) | MASK_BINARY_OR2_64
		if sig_x > 9999999999999999 || sig_x == 0 {
			x_is_zero = 1
		}
	} else {
		exp_x = int((x & MASK_BINARY_EXPONENT1_64) >> 53)
		sig_x = x & MASK_BINARY_SIG1_64
		if sig_x == 0 {
			x_is_zero = 1
		}
	}

	// Decode y
	if (y & MASK_STEERING_BITS64) == MASK_STEERING_BITS64 {
		exp_y = int((y & MASK_BINARY_EXPONENT2_64) >> 51)
		sig_y = (y & MASK_BINARY_SIG2_64) | MASK_BINARY_OR2_64
		if sig_y > 9999999999999999 || sig_y == 0 {
			y_is_zero = 1
		}
	} else {
		exp_y = int((y & MASK_BINARY_EXPONENT1_64) >> 53)
		sig_y = y & MASK_BINARY_SIG1_64
		if sig_y == 0 {
			y_is_zero = 1
		}
	}

	// ZERO (CASE5)
	if x_is_zero != 0 && y_is_zero != 0 {
		if ((x & MASK_SIGN64) == MASK_SIGN64) == ((y & MASK_SIGN64) == MASK_SIGN64) {
			if exp_x == exp_y {
				res = 1
				return res
			}
			if (exp_x <= exp_y) != ((x & MASK_SIGN64) == MASK_SIGN64) {
				res = 1
			} else {
				res = 0
			}
			return res
		}
		if (x & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		} else {
			res = 0
		}
		return res
	}
	if x_is_zero != 0 {
		if (y & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		} else {
			res = 0
		}
		return res
	}
	if y_is_zero != 0 {
		if (x & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		} else {
			res = 0
		}
		return res
	}

	// REDUNDANT REPRESENTATIONS (CASE6)
	if sig_x > sig_y && exp_x >= exp_y {
		if (x & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		} else {
			res = 0
		}
		return res
	}
	if sig_x < sig_y && exp_x <= exp_y {
		if (x & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		} else {
			res = 0
		}
		return res
	}

	if exp_x-exp_y > 15 {
		if (x & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		} else {
			res = 0
		}
		return res
	}
	if exp_y-exp_x > 15 {
		if (x & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		} else {
			res = 0
		}
		return res
	}

	if exp_x > exp_y {
		sig_n_prime = __mul_64x64_to_128(sig_x, bid_mult_factor[exp_x-exp_y])
		if sig_n_prime.w[1] == 0 && sig_n_prime.w[0] == sig_y {
			if (exp_x <= exp_y) != ((x & MASK_SIGN64) == MASK_SIGN64) {
				res = 1
			} else {
				res = 0
			}
			return res
		}
		cond := (sig_n_prime.w[1] == 0) && (sig_n_prime.w[0] < sig_y)
		sign := (x & MASK_SIGN64) == MASK_SIGN64
		if cond != sign {
			res = 1
		} else {
			res = 0
		}
		return res
	}

	sig_n_prime = __mul_64x64_to_128(sig_y, bid_mult_factor[exp_y-exp_x])
	if sig_n_prime.w[1] == 0 && sig_n_prime.w[0] == sig_x {
		if (exp_x <= exp_y) != ((x & MASK_SIGN64) == MASK_SIGN64) {
			res = 1
		} else {
			res = 0
		}
		return res
	}
	cond := (sig_n_prime.w[1] > 0) || (sig_x < sig_n_prime.w[0])
	sign := (x & MASK_SIGN64) == MASK_SIGN64
	if cond != sign {
		res = 1
	} else {
		res = 0
	}
	return res
}

// Bid64TotalOrderMag - returns 1 if |x| <= |y| in total order, 0 otherwise
// Intel bid64_totalOrderMag 기계적 포팅
func Bid64TotalOrderMag(x, y uint64) int {
	var res int
	var exp_x, exp_y int
	var sig_x, sig_y, pyld_y, pyld_x uint64
	var sig_n_prime BID_UINT128
	var x_is_zero, y_is_zero int

	// NaN (CASE1)
	if (x & MASK_NAN64) == MASK_NAN64 {
		// return false, unless y is NaN also
		if (y & MASK_NAN64) != MASK_NAN64 {
			res = 0
			return res
		}

		// both NaN: compare signaling class and payload
		xIsSNaN := (x & MASK_SNAN64) == MASK_SNAN64
		yIsSNaN := (y & MASK_SNAN64) == MASK_SNAN64
		if xIsSNaN == yIsSNaN {
			pyld_y = y & 0x0003ffffffffffff
			pyld_x = x & 0x0003ffffffffffff
			if pyld_x > 999999999999999 || pyld_x == 0 {
				res = 1
				return res
			}
			if pyld_y > 999999999999999 || pyld_y == 0 {
				res = 0
				return res
			}
			if pyld_x <= pyld_y {
				res = 1
			} else {
				res = 0
			}
			return res
		}

		if (x & MASK_SNAN64) == MASK_SNAN64 {
			res = 1
		} else {
			res = 0
		}
		return res
	} else if (y & MASK_NAN64) == MASK_NAN64 {
		res = 1
		return res
	}

	// SIMPLE (CASE2)
	if (x & ^MASK_SIGN64) == (y & ^MASK_SIGN64) {
		res = 1
		return res
	}

	// INFINITY (CASE3)
	if (x & MASK_INF64) == MASK_INF64 {
		if (y & MASK_INF64) == MASK_INF64 {
			res = 1
		} else {
			res = 0
		}
		return res
	} else if (y & MASK_INF64) == MASK_INF64 {
		res = 1
		return res
	}

	// Decode x
	if (x & MASK_STEERING_BITS64) == MASK_STEERING_BITS64 {
		exp_x = int((x & MASK_BINARY_EXPONENT2_64) >> 51)
		sig_x = (x & MASK_BINARY_SIG2_64) | MASK_BINARY_OR2_64
		if sig_x > 9999999999999999 || sig_x == 0 {
			x_is_zero = 1
		}
	} else {
		exp_x = int((x & MASK_BINARY_EXPONENT1_64) >> 53)
		sig_x = (x & MASK_BINARY_SIG1_64)
		if sig_x == 0 {
			x_is_zero = 1
		}
	}

	// Decode y
	if (y & MASK_STEERING_BITS64) == MASK_STEERING_BITS64 {
		exp_y = int((y & MASK_BINARY_EXPONENT2_64) >> 51)
		sig_y = (y & MASK_BINARY_SIG2_64) | MASK_BINARY_OR2_64
		if sig_y > 9999999999999999 || sig_y == 0 {
			y_is_zero = 1
		}
	} else {
		exp_y = int((y & MASK_BINARY_EXPONENT1_64) >> 53)
		sig_y = (y & MASK_BINARY_SIG1_64)
		if sig_y == 0 {
			y_is_zero = 1
		}
	}

	// ZERO (CASE5)
	if x_is_zero != 0 && y_is_zero != 0 {
		if exp_x <= exp_y {
			res = 1
		} else {
			res = 0
		}
		return res
	}
	if x_is_zero != 0 {
		res = 1
		return res
	}
	if y_is_zero != 0 {
		res = 0
		return res
	}

	// REDUNDANT REPRESENTATIONS (CASE6)
	if sig_x > sig_y && exp_x >= exp_y {
		res = 0
		return res
	}
	if sig_x < sig_y && exp_x <= exp_y {
		res = 1
		return res
	}

	if exp_x-exp_y > 15 {
		res = 0
		return res
	}
	if exp_y-exp_x > 15 {
		res = 1
		return res
	}

	if exp_x > exp_y {
		sig_n_prime = __mul_64x64_to_128(sig_x, bid_mult_factor[exp_x-exp_y])
		if sig_n_prime.w[1] == 0 && (sig_n_prime.w[0] == sig_y) {
			res = 0 // exp_x > exp_y
			return res
		}
		if (sig_n_prime.w[1] == 0) && sig_n_prime.w[0] < sig_y {
			res = 1
		} else {
			res = 0
		}
		return res
	}

	sig_n_prime = __mul_64x64_to_128(sig_y, bid_mult_factor[exp_y-exp_x])
	if sig_n_prime.w[1] == 0 && (sig_n_prime.w[0] == sig_x) {
		res = 1 // exp_x <= exp_y
		return res
	}
	if (sig_n_prime.w[1] > 0) || (sig_x < sig_n_prime.w[0]) {
		res = 1
	} else {
		res = 0
	}
	return res
}

// Bid64Quantum - returns the quantum of x (10^exp)
// Intel bid64_quantum 기계적 포팅 (bid64_quantumd.c)
// Exceptions: none
func Bid64Quantum(x uint64) uint64 {
	// If x is infinite, return +Inf
	if (x & MASK_INF64) == MASK_INF64 {
		return x & ^MASK_SIGN64
	}
	// If x is NaN, return quiet NaN (clear sign and signaling bit)
	if (x & MASK_NAN64) == MASK_NAN64 {
		// QUIET_MASK64 = 0x7dffffffffffffffull clears sign and signaling bit
		return x & 0x7dffffffffffffff
	}

	// Extract exponent
	var intExp int
	if (x & MASK_STEERING_BITS64) == MASK_STEERING_BITS64 {
		intExp = int((x>>51)&0x3ff) - 398
	} else {
		intExp = int((x>>53)&0x3ff) - 398
	}

	// Return 10^exp = 1 * 10^exp
	// 0x31c0000000000001 is the BID64 encoding of +1E0 with bias
	// We add intExp (already unbiased) to the biased base
	return (uint64(intExp) << 53) + 0x31c0000000000001
}

// Bid64Quantexp - returns the exponent of x as int32
// Intel bid64_quantexp 기계적 포팅 (bid64_quantexpd.c)
// Exceptions: INVALID if x is Inf or NaN
func Bid64Quantexp(x uint64) (int32, uint32) {
	// If Inf or NaN: set INVALID and return MIN_INT32
	if ((x & MASK_INF64) == MASK_INF64) || ((x & MASK_NAN64) == MASK_NAN64) {
		return -2147483648, BID_INVALID_EXCEPTION // 0x80000000
	}
	// Extract exponent
	if (x & MASK_STEERING_BITS64) == MASK_STEERING_BITS64 {
		return int32((x>>51)&0x3ff) - 398, 0
	}
	return int32((x>>53)&0x3ff) - 398, 0
}

// Bid64LLQuantexp - returns the exponent of x as int64
// Intel bid64_llquantexp 기계적 포팅 (bid64_llquantexpd.c)
// Exceptions: INVALID if x is Inf or NaN
func Bid64LLQuantexp(x uint64) (int64, uint32) {
	// If Inf or NaN: set INVALID and return MIN_INT64
	if ((x & MASK_INF64) == MASK_INF64) || ((x & MASK_NAN64) == MASK_NAN64) {
		return -9223372036854775808, BID_INVALID_EXCEPTION // 0x8000000000000000
	}
	// Extract exponent
	if (x & MASK_STEERING_BITS64) == MASK_STEERING_BITS64 {
		return int64((x>>51)&0x3ff) - 398, 0
	}
	return int64((x>>53)&0x3ff) - 398, 0
}

// Bid64SignalingLess - returns 1 if x < y, 0 otherwise
// Intel bid64_signaling_less 기계적 포팅 (bid64_compare.c)
// Exceptions: INVALID if either operand is NaN
func Bid64SignalingLess(x, y uint64) (int, uint32) {
	var res int
	var exp_x, exp_y int
	var sig_x, sig_y uint64
	var sig_n_prime BID_UINT128
	var pfpsf uint32
	var x_is_zero, y_is_zero, non_canon_x, non_canon_y int

	// NaN (CASE1)
	if ((x & MASK_NAN64) == MASK_NAN64) || ((y & MASK_NAN64) == MASK_NAN64) {
		pfpsf |= BID_INVALID_EXCEPTION
		res = 0
		return res, pfpsf
	}

	// SIMPLE (CASE2)
	if x == y {
		res = 0
		return res, pfpsf
	}

	// INFINITY (CASE3)
	if (x & MASK_INF64) == MASK_INF64 {
		// if x==neg_inf, { res = (y == neg_inf)?0:1; BID_RETURN (res) }
		if (x & MASK_SIGN64) == MASK_SIGN64 {
			// x is -inf, so it is less than y unless y is -inf
			res = 0
			if ((y & MASK_INF64) != MASK_INF64) || ((y & MASK_SIGN64) != MASK_SIGN64) {
				res = 1
			}
			return res, pfpsf
		}
		// x is pos_inf, no way for it to be less than y
		res = 0
		return res, pfpsf
	} else if (y & MASK_INF64) == MASK_INF64 {
		// x is finite, so:
		//    if y is +inf, x<y
		//    if y is -inf, x>y
		res = 0
		if (y & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}

	// if steering bits are 11 (condition will be 0), then exponent is G[0:w+1] =>
	if (x & MASK_STEERING_BITS64) == MASK_STEERING_BITS64 {
		exp_x = int((x & MASK_BINARY_EXPONENT2_64) >> 51)
		sig_x = (x & MASK_BINARY_SIG2_64) | MASK_BINARY_OR2_64
		if sig_x > 9999999999999999 {
			non_canon_x = 1
		} else {
			non_canon_x = 0
		}
	} else {
		exp_x = int((x & MASK_BINARY_EXPONENT1_64) >> 53)
		sig_x = x & MASK_BINARY_SIG1_64
		non_canon_x = 0
	}

	// if steering bits are 11 (condition will be 0), then exponent is G[0:w+1] =>
	if (y & MASK_STEERING_BITS64) == MASK_STEERING_BITS64 {
		exp_y = int((y & MASK_BINARY_EXPONENT2_64) >> 51)
		sig_y = (y & MASK_BINARY_SIG2_64) | MASK_BINARY_OR2_64
		if sig_y > 9999999999999999 {
			non_canon_y = 1
		} else {
			non_canon_y = 0
		}
	} else {
		exp_y = int((y & MASK_BINARY_EXPONENT1_64) >> 53)
		sig_y = y & MASK_BINARY_SIG1_64
		non_canon_y = 0
	}

	// ZERO (CASE4)
	if non_canon_x != 0 || sig_x == 0 {
		x_is_zero = 1
	}
	if non_canon_y != 0 || sig_y == 0 {
		y_is_zero = 1
	}

	if x_is_zero != 0 && y_is_zero != 0 {
		res = 0
		return res, pfpsf
	} else if x_is_zero != 0 {
		res = 0
		if (y & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	} else if y_is_zero != 0 {
		res = 0
		if (x & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}

	// OPPOSITE SIGN (CASE5)
	if ((x ^ y) & MASK_SIGN64) == MASK_SIGN64 {
		res = 0
		if (y & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}

	// REDUNDANT REPRESENTATIONS (CASE6)
	if sig_x > sig_y && exp_x >= exp_y {
		res = 0
		if (x & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	if sig_x < sig_y && exp_x <= exp_y {
		res = 0
		if (x & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}

	// Large exponent difference
	if exp_x-exp_y > 15 {
		res = 0
		if (x & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	if exp_y-exp_x > 15 {
		res = 0
		if (x & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}

	// if |exp_x - exp_y| < 15, it comes down to the compensated significand
	if exp_x > exp_y {
		// otherwise adjust the x significand upwards
		sig_n_prime = __mul_64x64_to_128(sig_x, bid_mult_factor[exp_x-exp_y])

		// return 0 if values are equal
		if sig_n_prime.w[1] == 0 && (sig_n_prime.w[0] == sig_y) {
			res = 0
			return res, pfpsf
		}
		// if positive, return whichever significand abs is smaller
		//     (converse if negative)
		res = 0
		if ((sig_n_prime.w[1] == 0) && sig_n_prime.w[0] < sig_y) != ((x & MASK_SIGN64) == MASK_SIGN64) {
			res = 1
		}
		return res, pfpsf
	}

	// adjust the y significand upwards
	sig_n_prime = __mul_64x64_to_128(sig_y, bid_mult_factor[exp_y-exp_x])

	// return 0 if values are equal
	if sig_n_prime.w[1] == 0 && (sig_n_prime.w[0] == sig_x) {
		res = 0
		return res, pfpsf
	}

	// if positive, return whichever significand abs is smaller
	//     (converse if negative)
	res = 0
	if ((sig_n_prime.w[1] > 0) || (sig_x < sig_n_prime.w[0])) != ((x & MASK_SIGN64) == MASK_SIGN64) {
		res = 1
	}
	return res, pfpsf
}
