// Ported from: IntelRDFPMathLib20U4/LIBRARY/src/bid32_minmax.c
// Version: Intel(R) Decimal Floating-Point Math Library 2.0 Update 4

package bidgo

// bid_mult_factor32 for minmax operations
var bid_mult_factor32 = []uint32{
	1, 10, 100, 1000, 10000, 100000, 1000000,
}

// bid32_minnum_pure returns the smaller of two numbers
func bid32_minnum_pure(x, y uint32) uint32 {
	var res uint32
	var exp_x, exp_y int
	var sig_x, sig_y uint32
	var sig_n_prime uint64
	var x_is_zero, y_is_zero bool

	// check for non-canonical x
	if (x & MASK_NAN32) == MASK_NAN32 { // x is NaN
		x = x & 0xfe0fffff // clear G6-G10
		if (x & 0x000fffff) > 999999 {
			x = x & 0xfe000000 // clear G6-G10 and the payload bits
		}
	} else if (x & MASK_INF32) == MASK_INF32 { // check for Infinity
		x = x & (MASK_SIGN32 | MASK_INF32)
	} else { // x is not special
		// check for non-canonical values - treated as zero
		if (x & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
			// if the steering bits are 11, then the exponent is G[0:w+1]
			if ((x & MASK_BINARY_SIG2_32) | MASK_BINARY_OR2_32) > 9999999 {
				// non-canonical
				x = (x & MASK_SIGN32) | ((x & MASK_BINARY_EXPONENT2_32) << 2)
			} // else canonical
		} // else canonical
	}

	// check for non-canonical y
	if (y & MASK_NAN32) == MASK_NAN32 { // y is NaN
		y = y & 0xfe0fffff // clear G6-G10
		if (y & 0x000fffff) > 999999 {
			y = y & 0xfe000000 // clear G6-G10 and the payload bits
		}
	} else if (y & MASK_INF32) == MASK_INF32 { // check for Infinity
		y = y & (MASK_SIGN32 | MASK_INF32)
	} else { // y is not special
		// check for non-canonical values - treated as zero
		if (y & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
			// if the steering bits are 11, then the exponent is G[0:w+1]
			if ((y & MASK_BINARY_SIG2_32) | MASK_BINARY_OR2_32) > 9999999 {
				// non-canonical
				y = (y & MASK_SIGN32) | ((y & MASK_BINARY_EXPONENT2_32) << 2)
			} // else canonical
		} // else canonical
	}

	// NaN (CASE1)
	if (x & MASK_NAN32) == MASK_NAN32 { // x is NAN
		if (x & MASK_SNAN32) == MASK_SNAN32 { // x is SNaN
			// if x is SNAN, then return quiet (x)
			x = x & 0xfdffffff // quietize x
			res = x
		} else { // x is QNaN
			if (y & MASK_NAN32) == MASK_NAN32 { // y is NAN
				res = x
			} else {
				res = y
			}
		}
		return res
	} else if (y & MASK_NAN32) == MASK_NAN32 { // y is NaN, but x is not
		if (y & MASK_SNAN32) == MASK_SNAN32 {
			y = y & 0xfdffffff // quietize y
			res = y
		} else {
			// will return x (which is not NaN)
			res = x
		}
		return res
	}

	// SIMPLE (CASE2)
	// if all the bits are the same, these numbers are equal, return either number
	if x == y {
		return x
	}

	// INFINITY (CASE3)
	if (x & MASK_INF32) == MASK_INF32 {
		// if x is neg infinity, there is no way it is greater than y, return x
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			return x
		}
		// x is pos infinity, return y
		return y
	} else if (y & MASK_INF32) == MASK_INF32 {
		// x is finite, so if y is positive infinity, then x is less, return y
		//                 if y is negative infinity, then x is greater, return x
		if (y & MASK_SIGN32) == MASK_SIGN32 {
			return y
		}
		return x
	}

	// if steering bits are 11, then exponent is G[0:w+1]
	if (x & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
		exp_x = int((x & MASK_BINARY_EXPONENT2_32) >> 21)
		sig_x = (x & MASK_BINARY_SIG2_32) | MASK_BINARY_OR2_32
	} else {
		exp_x = int((x & MASK_BINARY_EXPONENT1_32) >> 23)
		sig_x = x & MASK_BINARY_SIG1_32
	}

	if (y & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
		exp_y = int((y & MASK_BINARY_EXPONENT2_32) >> 21)
		sig_y = (y & MASK_BINARY_SIG2_32) | MASK_BINARY_OR2_32
	} else {
		exp_y = int((y & MASK_BINARY_EXPONENT1_32) >> 23)
		sig_y = y & MASK_BINARY_SIG1_32
	}

	// ZERO (CASE4)
	if sig_x == 0 {
		x_is_zero = true
	}
	if sig_y == 0 {
		y_is_zero = true
	}

	if x_is_zero && y_is_zero {
		// if both numbers are zero, neither is greater => return either
		return y
	} else if x_is_zero {
		// if x is zero, it is greater if Y is negative
		if (y & MASK_SIGN32) == MASK_SIGN32 {
			return y
		}
		return x
	} else if y_is_zero {
		// if y is zero, X is greater if it is positive
		if (x & MASK_SIGN32) != MASK_SIGN32 {
			return y
		}
		return x
	}

	// OPPOSITE SIGN (CASE5)
	if ((x ^ y) & MASK_SIGN32) == MASK_SIGN32 {
		if (y & MASK_SIGN32) == MASK_SIGN32 {
			return y
		}
		return x
	}

	// REDUNDANT REPRESENTATIONS (CASE6)
	if sig_x > sig_y && exp_x >= exp_y {
		if (x & MASK_SIGN32) != MASK_SIGN32 {
			return y
		}
		return x
	}
	if sig_x < sig_y && exp_x <= exp_y {
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			return y
		}
		return x
	}

	// if exp_x is 6 greater than exp_y, no need for compensation
	if exp_x-exp_y > 6 {
		if (x & MASK_SIGN32) != MASK_SIGN32 {
			return y // difference cannot be >10^6
		}
		return x
	}

	// if exp_x is 6 less than exp_y, no need for compensation
	if exp_y-exp_x > 6 {
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			return y
		}
		return x
	}

	// if |exp_x - exp_y|< 6, it comes down to the compensated significand
	if exp_x > exp_y {
		// adjust the x significand upwards
		sig_n_prime = uint64(sig_x) * uint64(bid_mult_factor32[exp_x-exp_y])
		if sig_n_prime == uint64(sig_y) {
			return y
		}
		// return based on sign and comparison
		if (sig_n_prime > uint64(sig_y)) != ((x & MASK_SIGN32) == MASK_SIGN32) {
			return y
		}
		return x
	}

	// adjust the y significand upwards
	sig_n_prime = uint64(sig_y) * uint64(bid_mult_factor32[exp_y-exp_x])
	if sig_n_prime == uint64(sig_x) {
		return y
	}
	if (uint64(sig_x) > sig_n_prime) != ((x & MASK_SIGN32) == MASK_SIGN32) {
		return y
	}
	return x
}

// bid32_maxnum_pure returns the larger of two numbers
func bid32_maxnum_pure(x, y uint32) uint32 {
	var res uint32
	var exp_x, exp_y int
	var sig_x, sig_y uint32
	var sig_n_prime uint64
	var x_is_zero, y_is_zero bool

	// check for non-canonical x
	if (x & MASK_NAN32) == MASK_NAN32 { // x is NaN
		x = x & 0xfe0fffff // clear G6-G10
		if (x & 0x000fffff) > 999999 {
			x = x & 0xfe000000 // clear G6-G10 and the payload bits
		}
	} else if (x & MASK_INF32) == MASK_INF32 { // check for Infinity
		x = x & (MASK_SIGN32 | MASK_INF32)
	} else { // x is not special
		if (x & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
			if ((x & MASK_BINARY_SIG2_32) | MASK_BINARY_OR2_32) > 9999999 {
				x = (x & MASK_SIGN32) | ((x & MASK_BINARY_EXPONENT2_32) << 2)
			}
		}
	}

	// check for non-canonical y
	if (y & MASK_NAN32) == MASK_NAN32 { // y is NaN
		y = y & 0xfe0fffff
		if (y & 0x000fffff) > 999999 {
			y = y & 0xfe000000
		}
	} else if (y & MASK_INF32) == MASK_INF32 {
		y = y & (MASK_SIGN32 | MASK_INF32)
	} else {
		if (y & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
			if ((y & MASK_BINARY_SIG2_32) | MASK_BINARY_OR2_32) > 9999999 {
				y = (y & MASK_SIGN32) | ((y & MASK_BINARY_EXPONENT2_32) << 2)
			}
		}
	}

	// NaN (CASE1)
	if (x & MASK_NAN32) == MASK_NAN32 { // x is NAN
		if (x & MASK_SNAN32) == MASK_SNAN32 { // x is SNaN
			x = x & 0xfdffffff // quietize x
			res = x
		} else { // x is QNaN
			if (y & MASK_NAN32) == MASK_NAN32 { // y is NAN
				res = x
			} else {
				res = y
			}
		}
		return res
	} else if (y & MASK_NAN32) == MASK_NAN32 { // y is NaN, but x is not
		if (y & MASK_SNAN32) == MASK_SNAN32 {
			y = y & 0xfdffffff // quietize y
			res = y
		} else {
			res = x
		}
		return res
	}

	// SIMPLE (CASE2)
	if x == y {
		return x
	}

	// INFINITY (CASE3)
	if (x & MASK_INF32) == MASK_INF32 {
		if (x & MASK_SIGN32) == MASK_SIGN32 { // x = -infinity
			return y
		}
		return x // x = +infinity
	} else if (y & MASK_INF32) == MASK_INF32 {
		if (y & MASK_SIGN32) == MASK_SIGN32 {
			return x
		}
		return y
	}

	// Extract exponents and significands
	if (x & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
		exp_x = int((x & MASK_BINARY_EXPONENT2_32) >> 21)
		sig_x = (x & MASK_BINARY_SIG2_32) | MASK_BINARY_OR2_32
	} else {
		exp_x = int((x & MASK_BINARY_EXPONENT1_32) >> 23)
		sig_x = x & MASK_BINARY_SIG1_32
	}

	if (y & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
		exp_y = int((y & MASK_BINARY_EXPONENT2_32) >> 21)
		sig_y = (y & MASK_BINARY_SIG2_32) | MASK_BINARY_OR2_32
	} else {
		exp_y = int((y & MASK_BINARY_EXPONENT1_32) >> 23)
		sig_y = y & MASK_BINARY_SIG1_32
	}

	// ZERO (CASE4)
	if sig_x == 0 {
		x_is_zero = true
	}
	if sig_y == 0 {
		y_is_zero = true
	}

	if x_is_zero && y_is_zero {
		return y
	} else if x_is_zero {
		if (y & MASK_SIGN32) == MASK_SIGN32 {
			return x
		}
		return y
	} else if y_is_zero {
		if (x & MASK_SIGN32) != MASK_SIGN32 {
			return x
		}
		return y
	}

	// OPPOSITE SIGN (CASE5)
	if ((x ^ y) & MASK_SIGN32) == MASK_SIGN32 {
		if (y & MASK_SIGN32) == MASK_SIGN32 {
			return x
		}
		return y
	}

	// REDUNDANT REPRESENTATIONS (CASE6)
	if sig_x > sig_y && exp_x >= exp_y {
		if (x & MASK_SIGN32) != MASK_SIGN32 {
			return x
		}
		return y
	}
	if sig_x < sig_y && exp_x <= exp_y {
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			return x
		}
		return y
	}

	if exp_x-exp_y > 6 {
		if (x & MASK_SIGN32) != MASK_SIGN32 {
			return x
		}
		return y
	}
	if exp_y-exp_x > 6 {
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			return x
		}
		return y
	}

	if exp_x > exp_y {
		sig_n_prime = uint64(sig_x) * uint64(bid_mult_factor32[exp_x-exp_y])
		if sig_n_prime == uint64(sig_y) {
			return y
		}
		if (sig_n_prime > uint64(sig_y)) != ((x & MASK_SIGN32) == MASK_SIGN32) {
			return x
		}
		return y
	}

	sig_n_prime = uint64(sig_y) * uint64(bid_mult_factor32[exp_y-exp_x])
	if sig_n_prime == uint64(sig_x) {
		return y
	}
	if (uint64(sig_x) > sig_n_prime) != ((x & MASK_SIGN32) == MASK_SIGN32) {
		return x
	}
	return y
}

// bid32_minnum_mag_pure returns the number with smaller magnitude
func bid32_minnum_mag_pure(x, y uint32) uint32 {
	var exp_x, exp_y int
	var sig_x, sig_y uint32
	var sig_n_prime uint64

	// check for non-canonical x
	if (x & MASK_NAN32) == MASK_NAN32 {
		x = x & 0xfe0fffff
		if (x & 0x000fffff) > 999999 {
			x = x & 0xfe000000
		}
	} else if (x & MASK_INF32) == MASK_INF32 {
		x = x & (MASK_SIGN32 | MASK_INF32)
	} else {
		if (x & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
			if ((x & MASK_BINARY_SIG2_32) | MASK_BINARY_OR2_32) > 9999999 {
				x = (x & MASK_SIGN32) | ((x & MASK_BINARY_EXPONENT2_32) << 2)
			}
		}
	}

	// check for non-canonical y
	if (y & MASK_NAN32) == MASK_NAN32 {
		y = y & 0xfe0fffff
		if (y & 0x000fffff) > 999999 {
			y = y & 0xfe000000
		}
	} else if (y & MASK_INF32) == MASK_INF32 {
		y = y & (MASK_SIGN32 | MASK_INF32)
	} else {
		if (y & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
			if ((y & MASK_BINARY_SIG2_32) | MASK_BINARY_OR2_32) > 9999999 {
				y = (y & MASK_SIGN32) | ((y & MASK_BINARY_EXPONENT2_32) << 2)
			}
		}
	}

	// NaN (CASE1)
	if (x & MASK_NAN32) == MASK_NAN32 {
		if (x & MASK_SNAN32) == MASK_SNAN32 {
			return x & 0xfdffffff
		}
		if (y & MASK_NAN32) == MASK_NAN32 {
			return x
		}
		return y
	} else if (y & MASK_NAN32) == MASK_NAN32 {
		if (y & MASK_SNAN32) == MASK_SNAN32 {
			return y & 0xfdffffff
		}
		return x
	}

	// SIMPLE (CASE2)
	if x == y {
		return x
	}

	// INFINITY (CASE3)
	if (x & MASK_INF32) == MASK_INF32 {
		// x is infinity, return y unless both are infinity and x is negative
		if (x&MASK_SIGN32) == MASK_SIGN32 && (y&MASK_INF32) == MASK_INF32 {
			return x
		}
		return y
	} else if (y & MASK_INF32) == MASK_INF32 {
		return x
	}

	// Extract exponents and significands
	if (x & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
		exp_x = int((x & MASK_BINARY_EXPONENT2_32) >> 21)
		sig_x = (x & MASK_BINARY_SIG2_32) | MASK_BINARY_OR2_32
	} else {
		exp_x = int((x & MASK_BINARY_EXPONENT1_32) >> 23)
		sig_x = x & MASK_BINARY_SIG1_32
	}

	if (y & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
		exp_y = int((y & MASK_BINARY_EXPONENT2_32) >> 21)
		sig_y = (y & MASK_BINARY_SIG2_32) | MASK_BINARY_OR2_32
	} else {
		exp_y = int((y & MASK_BINARY_EXPONENT1_32) >> 23)
		sig_y = y & MASK_BINARY_SIG1_32
	}

	// ZERO (CASE4)
	if sig_x == 0 {
		return x
	}
	if sig_y == 0 {
		return y
	}

	// REDUNDANT REPRESENTATIONS (CASE6)
	if sig_x > sig_y && exp_x >= exp_y {
		return y
	}
	if sig_x < sig_y && exp_x <= exp_y {
		return x
	}

	if exp_x-exp_y > 6 {
		return y
	}
	if exp_y-exp_x > 6 {
		return x
	}

	if exp_x > exp_y {
		sig_n_prime = uint64(sig_x) * uint64(bid_mult_factor32[exp_x-exp_y])
		if sig_n_prime == uint64(sig_y) {
			if (y & MASK_SIGN32) == MASK_SIGN32 {
				return y
			}
			return x
		}
		if sig_n_prime > uint64(sig_y) {
			return y
		}
		return x
	}

	sig_n_prime = uint64(sig_y) * uint64(bid_mult_factor32[exp_y-exp_x])
	if sig_n_prime == uint64(sig_x) {
		if (y & MASK_SIGN32) == MASK_SIGN32 {
			return y
		}
		return x
	}
	if uint64(sig_x) > sig_n_prime {
		return y
	}
	return x
}

// bid32_maxnum_mag_pure returns the number with larger magnitude
func bid32_maxnum_mag_pure(x, y uint32) uint32 {
	var exp_x, exp_y int
	var sig_x, sig_y uint32
	var sig_n_prime uint64

	// check for non-canonical x
	if (x & MASK_NAN32) == MASK_NAN32 {
		x = x & 0xfe0fffff
		if (x & 0x000fffff) > 999999 {
			x = x & 0xfe000000
		}
	} else if (x & MASK_INF32) == MASK_INF32 {
		x = x & (MASK_SIGN32 | MASK_INF32)
	} else {
		if (x & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
			if ((x & MASK_BINARY_SIG2_32) | MASK_BINARY_OR2_32) > 9999999 {
				x = (x & MASK_SIGN32) | ((x & MASK_BINARY_EXPONENT2_32) << 2)
			}
		}
	}

	// check for non-canonical y
	if (y & MASK_NAN32) == MASK_NAN32 {
		y = y & 0xfe0fffff
		if (y & 0x000fffff) > 999999 {
			y = y & 0xfe000000
		}
	} else if (y & MASK_INF32) == MASK_INF32 {
		y = y & (MASK_SIGN32 | MASK_INF32)
	} else {
		if (y & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
			if ((y & MASK_BINARY_SIG2_32) | MASK_BINARY_OR2_32) > 9999999 {
				y = (y & MASK_SIGN32) | ((y & MASK_BINARY_EXPONENT2_32) << 2)
			}
		}
	}

	// NaN (CASE1)
	if (x & MASK_NAN32) == MASK_NAN32 {
		if (x & MASK_SNAN32) == MASK_SNAN32 {
			return x & 0xfdffffff
		}
		if (y & MASK_NAN32) == MASK_NAN32 {
			return x
		}
		return y
	} else if (y & MASK_NAN32) == MASK_NAN32 {
		if (y & MASK_SNAN32) == MASK_SNAN32 {
			return y & 0xfdffffff
		}
		return x
	}

	// SIMPLE (CASE2)
	if x == y {
		return x
	}

	// INFINITY (CASE3)
	if (x & MASK_INF32) == MASK_INF32 {
		// x is infinity, return x unless x is negative and y is also infinity
		if (x&MASK_SIGN32) == MASK_SIGN32 && (y&MASK_INF32) == MASK_INF32 {
			return y
		}
		return x
	} else if (y & MASK_INF32) == MASK_INF32 {
		return y
	}

	// Extract exponents and significands
	if (x & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
		exp_x = int((x & MASK_BINARY_EXPONENT2_32) >> 21)
		sig_x = (x & MASK_BINARY_SIG2_32) | MASK_BINARY_OR2_32
	} else {
		exp_x = int((x & MASK_BINARY_EXPONENT1_32) >> 23)
		sig_x = x & MASK_BINARY_SIG1_32
	}

	if (y & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
		exp_y = int((y & MASK_BINARY_EXPONENT2_32) >> 21)
		sig_y = (y & MASK_BINARY_SIG2_32) | MASK_BINARY_OR2_32
	} else {
		exp_y = int((y & MASK_BINARY_EXPONENT1_32) >> 23)
		sig_y = y & MASK_BINARY_SIG1_32
	}

	// ZERO (CASE4)
	if sig_x == 0 {
		return y
	}
	if sig_y == 0 {
		return x
	}

	// REDUNDANT REPRESENTATIONS (CASE6)
	if sig_x > sig_y && exp_x >= exp_y {
		return x
	}
	if sig_x < sig_y && exp_x <= exp_y {
		return y
	}

	if exp_x-exp_y > 6 {
		return x
	}
	if exp_y-exp_x > 6 {
		return y
	}

	if exp_x > exp_y {
		sig_n_prime = uint64(sig_x) * uint64(bid_mult_factor32[exp_x-exp_y])
		if sig_n_prime == uint64(sig_y) {
			if (y & MASK_SIGN32) == MASK_SIGN32 {
				return x
			}
			return y
		}
		if sig_n_prime > uint64(sig_y) {
			return x
		}
		return y
	}

	sig_n_prime = uint64(sig_y) * uint64(bid_mult_factor32[exp_y-exp_x])
	if sig_n_prime == uint64(sig_x) {
		if (y & MASK_SIGN32) == MASK_SIGN32 {
			return x
		}
		return y
	}
	if uint64(sig_x) > sig_n_prime {
		return x
	}
	return y
}

// bid32_sameQuantum_pure returns true if x and y have the same quantum (exponent)
func bid32_sameQuantum_pure(x, y uint32) bool {
	var exp_x, exp_y uint32

	// if both operands are NaN, return true; if just one is NaN, return false
	if (x&MASK_NAN32) == MASK_NAN32 || (y&MASK_NAN32) == MASK_NAN32 {
		return (x&MASK_NAN32) == MASK_NAN32 && (y&MASK_NAN32) == MASK_NAN32
	}

	// if both operands are INF, return true; if just one is INF, return false
	if (x&MASK_INF32) == MASK_INF32 || (y&MASK_INF32) == MASK_INF32 {
		return (x&MASK_INF32) == MASK_INF32 && (y&MASK_INF32) == MASK_INF32
	}

	// decode exponents for both numbers, and return true if they match
	if (x & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
		exp_x = (x & MASK_BINARY_EXPONENT2_32) >> 21
	} else {
		exp_x = (x & MASK_BINARY_EXPONENT1_32) >> 23
	}

	if (y & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
		exp_y = (y & MASK_BINARY_EXPONENT2_32) >> 21
	} else {
		exp_y = (y & MASK_BINARY_EXPONENT1_32) >> 23
	}

	return exp_x == exp_y
}

// bid32_quantum_pure returns the quantum (10^exponent) of x
func bid32_quantum_pure(x uint32) uint32 {
	var int_exp int

	// If x is infinite, the result is +Inf. If x is NaN, the result is NaN
	if (x & MASK_INF32) == MASK_INF32 {
		return x & ^uint32(MASK_SIGN32)
	}
	if (x & MASK_NAN32) == MASK_NAN32 {
		// Return quiet NaN with same payload but positive
		return x & 0x7fffffff
	}

	// Extract exponent
	// Note: C uses MASK_STEERING_BITS (64-bit 0x6000000000000000) with 32-bit x,
	// so the condition is always false. We replicate this behavior.
	int_exp = int((x>>23)&0xff) - 101

	// Form 10^exponent * 1
	// Result is 1 * 10^(int_exp + bias) where bias = 101
	// Using format: sign=0, exp=int_exp+101, sig=1
	return uint32((int_exp+101)<<23) + 1
}

// MinNum returns the smaller of a and b
func (a Decimal32Pure) MinNum(b Decimal32Pure) Decimal32Pure {
	return Decimal32Pure(bid32_minnum_pure(uint32(a), uint32(b)))
}

// MaxNum returns the larger of a and b
func (a Decimal32Pure) MaxNum(b Decimal32Pure) Decimal32Pure {
	return Decimal32Pure(bid32_maxnum_pure(uint32(a), uint32(b)))
}

// MinNumMag returns the number with smaller magnitude
func (a Decimal32Pure) MinNumMag(b Decimal32Pure) Decimal32Pure {
	return Decimal32Pure(bid32_minnum_mag_pure(uint32(a), uint32(b)))
}

// MaxNumMag returns the number with larger magnitude
func (a Decimal32Pure) MaxNumMag(b Decimal32Pure) Decimal32Pure {
	return Decimal32Pure(bid32_maxnum_mag_pure(uint32(a), uint32(b)))
}

// SameQuantum returns true if a and b have the same quantum (exponent)
func (a Decimal32Pure) SameQuantum(b Decimal32Pure) bool {
	return bid32_sameQuantum_pure(uint32(a), uint32(b))
}

// Quantum returns the quantum (10^exponent) of the number
func (d Decimal32Pure) Quantum() Decimal32Pure {
	return Decimal32Pure(bid32_quantum_pure(uint32(d)))
}
