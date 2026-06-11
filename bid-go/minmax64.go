package bidgo

// bid64_minmax.c 기계적 포팅
// Intel BID 라이브러리의 min/max 함수들

// bid_mult_factor for minmax (local copy, same as noncomp64.go)
var bid_mult_factor_minmax = [16]uint64{
	1, 10, 100, 1000,
	10000, 100000, 1000000, 10000000,
	100000000, 1000000000, 10000000000, 100000000000,
	1000000000000, 10000000000000,
	100000000000000, 1000000000000000,
}

// Bid64MinNum - returns the minimum of two numbers
// Intel bid64_minnum 기계적 포팅
func Bid64MinNum(x, y uint64) (uint64, uint32) {
	var res uint64
	var exp_x, exp_y int
	var sig_x, sig_y uint64
	var sig_n_prime BID_UINT128
	var x_is_zero, y_is_zero int
	var flags uint32 = 0

	// check for non-canonical x
	if (x & MASK_NAN64) == MASK_NAN64 { // x is NaN
		x = x & 0xfe03ffffffffffff // clear G6-G12
		if (x & 0x0003ffffffffffff) > 999999999999999 {
			x = x & 0xfe00000000000000 // clear G6-G12 and the payload bits
		}
	} else if (x & MASK_INF64) == MASK_INF64 { // check for Infinity
		x = x & (MASK_SIGN64 | MASK_INF64)
	} else { // x is not special
		// check for non-canonical values - treated as zero
		if (x & MASK_STEERING_BITS64) == MASK_STEERING_BITS64 {
			// if the steering bits are 11, then the exponent is G[0:w+1]
			if ((x & MASK_BINARY_SIG2_64) | MASK_BINARY_OR2_64) > 9999999999999999 {
				// non-canonical
				x = (x & MASK_SIGN64) | ((x & MASK_BINARY_EXPONENT2_64) << 2)
			} // else canonical
		} // else canonical
	}

	// check for non-canonical y
	if (y & MASK_NAN64) == MASK_NAN64 { // y is NaN
		y = y & 0xfe03ffffffffffff // clear G6-G12
		if (y & 0x0003ffffffffffff) > 999999999999999 {
			y = y & 0xfe00000000000000 // clear G6-G12 and the payload bits
		}
	} else if (y & MASK_INF64) == MASK_INF64 { // check for Infinity
		y = y & (MASK_SIGN64 | MASK_INF64)
	} else { // y is not special
		// check for non-canonical values - treated as zero
		if (y & MASK_STEERING_BITS64) == MASK_STEERING_BITS64 {
			// if the steering bits are 11, then the exponent is G[0:w+1]
			if ((y & MASK_BINARY_SIG2_64) | MASK_BINARY_OR2_64) > 9999999999999999 {
				// non-canonical
				y = (y & MASK_SIGN64) | ((y & MASK_BINARY_EXPONENT2_64) << 2)
			} // else canonical
		} // else canonical
	}

	// NaN (CASE1)
	if (x & MASK_NAN64) == MASK_NAN64 { // x is NAN
		if (x & MASK_SNAN64) == MASK_SNAN64 { // x is SNaN
			// if x is SNAN, then return quiet (x)
			flags |= BID_INVALID_EXCEPTION // set exception if SNaN
			x = x & 0xfdffffffffffffff     // quietize x
			res = x
		} else { // x is QNaN
			if (y & MASK_NAN64) == MASK_NAN64 { // y is NAN
				if (y & MASK_SNAN64) == MASK_SNAN64 { // y is SNAN
					flags |= BID_INVALID_EXCEPTION // set invalid flag
				}
				res = x
			} else {
				res = y
			}
		}
		return res, flags
	} else if (y & MASK_NAN64) == MASK_NAN64 { // y is NaN, but x is not
		if (y & MASK_SNAN64) == MASK_SNAN64 {
			flags |= BID_INVALID_EXCEPTION // set exception if SNaN
			y = y & 0xfdffffffffffffff     // quietize y
			res = y
		} else {
			// will return x (which is not NaN)
			res = x
		}
		return res, flags
	}

	// SIMPLE (CASE2)
	// if all the bits are the same, these numbers are equal, return either number
	if x == y {
		res = x
		return res, flags
	}

	// INFINITY (CASE3)
	if (x & MASK_INF64) == MASK_INF64 {
		// if x is neg infinity, there is no way it is greater than y, return x
		if (x & MASK_SIGN64) == MASK_SIGN64 {
			res = x
			return res, flags
		}
		// x is pos infinity, return y
		res = y
		return res, flags
	} else if (y & MASK_INF64) == MASK_INF64 {
		// x is finite, so if y is positive infinity, then x is less, return x
		//                 if y is negative infinity, then x is greater, return y
		if (y & MASK_SIGN64) == MASK_SIGN64 {
			res = y
		} else {
			res = x
		}
		return res, flags
	}

	// if steering bits are 11 (condition will be 0), then exponent is G[0:w+1] =>
	if (x & MASK_STEERING_BITS64) == MASK_STEERING_BITS64 {
		exp_x = int((x & MASK_BINARY_EXPONENT2_64) >> 51)
		sig_x = (x & MASK_BINARY_SIG2_64) | MASK_BINARY_OR2_64
	} else {
		exp_x = int((x & MASK_BINARY_EXPONENT1_64) >> 53)
		sig_x = x & MASK_BINARY_SIG1_64
	}

	// if steering bits are 11 (condition will be 0), then exponent is G[0:w+1] =>
	if (y & MASK_STEERING_BITS64) == MASK_STEERING_BITS64 {
		exp_y = int((y & MASK_BINARY_EXPONENT2_64) >> 51)
		sig_y = (y & MASK_BINARY_SIG2_64) | MASK_BINARY_OR2_64
	} else {
		exp_y = int((y & MASK_BINARY_EXPONENT1_64) >> 53)
		sig_y = y & MASK_BINARY_SIG1_64
	}

	// ZERO (CASE4)
	if sig_x == 0 {
		x_is_zero = 1
	}
	if sig_y == 0 {
		y_is_zero = 1
	}

	if x_is_zero != 0 && y_is_zero != 0 {
		// if both numbers are zero, neither is greater => return either
		res = y
		return res, flags
	} else if x_is_zero != 0 {
		// is x is zero, it is greater if Y is negative
		if (y & MASK_SIGN64) == MASK_SIGN64 {
			res = y
		} else {
			res = x
		}
		return res, flags
	} else if y_is_zero != 0 {
		// is y is zero, X is greater if it is positive
		if (x & MASK_SIGN64) != MASK_SIGN64 {
			res = y
		} else {
			res = x
		}
		return res, flags
	}

	// OPPOSITE SIGN (CASE5)
	// now, if the sign bits differ, x is greater if y is negative
	if ((x ^ y) & MASK_SIGN64) == MASK_SIGN64 {
		if (y & MASK_SIGN64) == MASK_SIGN64 {
			res = y
		} else {
			res = x
		}
		return res, flags
	}

	// REDUNDANT REPRESENTATIONS (CASE6)
	// if both components are either bigger or smaller,
	// it is clear what needs to be done
	if sig_x > sig_y && exp_x >= exp_y {
		if (x & MASK_SIGN64) != MASK_SIGN64 {
			res = y
		} else {
			res = x
		}
		return res, flags
	}
	if sig_x < sig_y && exp_x <= exp_y {
		if (x & MASK_SIGN64) == MASK_SIGN64 {
			res = y
		} else {
			res = x
		}
		return res, flags
	}

	// if exp_x is 15 greater than exp_y, no need for compensation
	if exp_x-exp_y > 15 {
		if (x & MASK_SIGN64) != MASK_SIGN64 {
			res = y
		} else {
			res = x
		}
		return res, flags
	}

	// if exp_x is 15 less than exp_y, no need for compensation
	if exp_y-exp_x > 15 {
		if (x & MASK_SIGN64) == MASK_SIGN64 {
			res = y
		} else {
			res = x
		}
		return res, flags
	}

	// if |exp_x - exp_y| < 15, it comes down to the compensated significand
	if exp_x > exp_y { // to simplify the loop below,
		// otherwise adjust the x significand upwards
		sig_n_prime = __mul_64x64_to_128(sig_x, bid_mult_factor_minmax[exp_x-exp_y])
		// if positive, return whichever significand is larger
		// (converse if negative)
		if sig_n_prime.w[1] == 0 && sig_n_prime.w[0] == sig_y {
			res = y
			return res, flags
		}

		cond := (sig_n_prime.w[1] > 0) || (sig_n_prime.w[0] > sig_y)
		sign := (x & MASK_SIGN64) == MASK_SIGN64
		if cond != sign {
			res = y
		} else {
			res = x
		}
		return res, flags
	}

	// adjust the y significand upwards
	sig_n_prime = __mul_64x64_to_128(sig_y, bid_mult_factor_minmax[exp_y-exp_x])

	// if positive, return whichever significand is larger (converse if negative)
	if sig_n_prime.w[1] == 0 && sig_n_prime.w[0] == sig_x {
		res = y
		return res, flags
	}

	cond := (sig_n_prime.w[1] == 0) && (sig_x > sig_n_prime.w[0])
	sign := (x & MASK_SIGN64) == MASK_SIGN64
	if cond != sign {
		res = y
	} else {
		res = x
	}
	return res, flags
}

// Bid64MinNumMag - returns the number with smaller magnitude
// Intel bid64_minnum_mag 기계적 포팅
func Bid64MinNumMag(x, y uint64) (uint64, uint32) {
	var res uint64
	var exp_x, exp_y int
	var sig_x, sig_y uint64
	var sig_n_prime BID_UINT128
	var flags uint32 = 0

	// check for non-canonical x
	if (x & MASK_NAN64) == MASK_NAN64 { // x is NaN
		x = x & 0xfe03ffffffffffff // clear G6-G12
		if (x & 0x0003ffffffffffff) > 999999999999999 {
			x = x & 0xfe00000000000000 // clear G6-G12 and the payload bits
		}
	} else if (x & MASK_INF64) == MASK_INF64 { // check for Infinity
		x = x & (MASK_SIGN64 | MASK_INF64)
	} else { // x is not special
		// check for non-canonical values - treated as zero
		if (x & MASK_STEERING_BITS64) == MASK_STEERING_BITS64 {
			if ((x & MASK_BINARY_SIG2_64) | MASK_BINARY_OR2_64) > 9999999999999999 {
				x = (x & MASK_SIGN64) | ((x & MASK_BINARY_EXPONENT2_64) << 2)
			}
		}
	}

	// check for non-canonical y
	if (y & MASK_NAN64) == MASK_NAN64 { // y is NaN
		y = y & 0xfe03ffffffffffff // clear G6-G12
		if (y & 0x0003ffffffffffff) > 999999999999999 {
			y = y & 0xfe00000000000000 // clear G6-G12 and the payload bits
		}
	} else if (y & MASK_INF64) == MASK_INF64 { // check for Infinity
		y = y & (MASK_SIGN64 | MASK_INF64)
	} else { // y is not special
		// check for non-canonical values - treated as zero
		if (y & MASK_STEERING_BITS64) == MASK_STEERING_BITS64 {
			if ((y & MASK_BINARY_SIG2_64) | MASK_BINARY_OR2_64) > 9999999999999999 {
				y = (y & MASK_SIGN64) | ((y & MASK_BINARY_EXPONENT2_64) << 2)
			}
		}
	}

	// NaN (CASE1)
	if (x & MASK_NAN64) == MASK_NAN64 { // x is NAN
		if (x & MASK_SNAN64) == MASK_SNAN64 { // x is SNaN
			flags |= BID_INVALID_EXCEPTION
			x = x & 0xfdffffffffffffff // quietize x
			res = x
		} else { // x is QNaN
			if (y & MASK_NAN64) == MASK_NAN64 { // y is NAN
				if (y & MASK_SNAN64) == MASK_SNAN64 { // y is SNAN
					flags |= BID_INVALID_EXCEPTION
				}
				res = x
			} else {
				res = y
			}
		}
		return res, flags
	} else if (y & MASK_NAN64) == MASK_NAN64 { // y is NaN, but x is not
		if (y & MASK_SNAN64) == MASK_SNAN64 {
			flags |= BID_INVALID_EXCEPTION
			y = y & 0xfdffffffffffffff // quietize y
			res = y
		} else {
			res = x
		}
		return res, flags
	}

	// SIMPLE (CASE2)
	if x == y {
		res = x
		return res, flags
	}

	// INFINITY (CASE3)
	if (x & MASK_INF64) == MASK_INF64 {
		// x is infinity, its magnitude is greater than or equal to y
		// return x only if y is infinity and x is negative
		if (x&MASK_SIGN64) == MASK_SIGN64 && (y&MASK_INF64) == MASK_INF64 {
			res = x
		} else {
			res = y
		}
		return res, flags
	} else if (y & MASK_INF64) == MASK_INF64 {
		// y is infinity, then it must be greater in magnitude, return x
		res = x
		return res, flags
	}

	// decode x
	if (x & MASK_STEERING_BITS64) == MASK_STEERING_BITS64 {
		exp_x = int((x & MASK_BINARY_EXPONENT2_64) >> 51)
		sig_x = (x & MASK_BINARY_SIG2_64) | MASK_BINARY_OR2_64
	} else {
		exp_x = int((x & MASK_BINARY_EXPONENT1_64) >> 53)
		sig_x = x & MASK_BINARY_SIG1_64
	}

	// decode y
	if (y & MASK_STEERING_BITS64) == MASK_STEERING_BITS64 {
		exp_y = int((y & MASK_BINARY_EXPONENT2_64) >> 51)
		sig_y = (y & MASK_BINARY_SIG2_64) | MASK_BINARY_OR2_64
	} else {
		exp_y = int((y & MASK_BINARY_EXPONENT1_64) >> 53)
		sig_y = y & MASK_BINARY_SIG1_64
	}

	// ZERO (CASE4)
	if sig_x == 0 {
		res = x // x_is_zero, its magnitude must be smaller than y
		return res, flags
	}
	if sig_y == 0 {
		res = y // y_is_zero, its magnitude must be smaller than x
		return res, flags
	}

	// REDUNDANT REPRESENTATIONS (CASE6)
	if sig_x > sig_y && exp_x >= exp_y {
		res = y
		return res, flags
	}
	if sig_x < sig_y && exp_x <= exp_y {
		res = x
		return res, flags
	}

	// if exp_x is 15 greater than exp_y, no need for compensation
	if exp_x-exp_y > 15 {
		res = y // difference cannot be greater than 10^15
		return res, flags
	}

	// if exp_x is 15 less than exp_y, no need for compensation
	if exp_y-exp_x > 15 {
		res = x
		return res, flags
	}

	// if |exp_x - exp_y| < 15, it comes down to the compensated significand
	if exp_x > exp_y {
		sig_n_prime = __mul_64x64_to_128(sig_x, bid_mult_factor_minmax[exp_x-exp_y])
		if sig_n_prime.w[1] == 0 && sig_n_prime.w[0] == sig_y {
			// two numbers are equal, return minNum(x,y)
			if (y & MASK_SIGN64) == MASK_SIGN64 {
				res = y
			} else {
				res = x
			}
			return res, flags
		}
		// if compensated_x is greater than y, return y, otherwise return x
		if (sig_n_prime.w[1] != 0) || sig_n_prime.w[0] > sig_y {
			res = y
		} else {
			res = x
		}
		return res, flags
	}

	// exp_y must be greater than exp_x, thus adjust the y significand upwards
	sig_n_prime = __mul_64x64_to_128(sig_y, bid_mult_factor_minmax[exp_y-exp_x])

	if sig_n_prime.w[1] == 0 && sig_n_prime.w[0] == sig_x {
		if (y & MASK_SIGN64) == MASK_SIGN64 {
			res = y
		} else {
			res = x
		}
		return res, flags
	}

	if (sig_n_prime.w[1] == 0) && (sig_x > sig_n_prime.w[0]) {
		res = y
	} else {
		res = x
	}
	return res, flags
}

// Bid64MaxNum - returns the maximum of two numbers
// Intel bid64_maxnum 기계적 포팅
func Bid64MaxNum(x, y uint64) (uint64, uint32) {
	var res uint64
	var exp_x, exp_y int
	var sig_x, sig_y uint64
	var sig_n_prime BID_UINT128
	var x_is_zero, y_is_zero int
	var flags uint32 = 0

	// check for non-canonical x
	if (x & MASK_NAN64) == MASK_NAN64 { // x is NaN
		x = x & 0xfe03ffffffffffff // clear G6-G12
		if (x & 0x0003ffffffffffff) > 999999999999999 {
			x = x & 0xfe00000000000000 // clear G6-G12 and the payload bits
		}
	} else if (x & MASK_INF64) == MASK_INF64 { // check for Infinity
		x = x & (MASK_SIGN64 | MASK_INF64)
	} else { // x is not special
		if (x & MASK_STEERING_BITS64) == MASK_STEERING_BITS64 {
			if ((x & MASK_BINARY_SIG2_64) | MASK_BINARY_OR2_64) > 9999999999999999 {
				x = (x & MASK_SIGN64) | ((x & MASK_BINARY_EXPONENT2_64) << 2)
			}
		}
	}

	// check for non-canonical y
	if (y & MASK_NAN64) == MASK_NAN64 { // y is NaN
		y = y & 0xfe03ffffffffffff // clear G6-G12
		if (y & 0x0003ffffffffffff) > 999999999999999 {
			y = y & 0xfe00000000000000 // clear G6-G12 and the payload bits
		}
	} else if (y & MASK_INF64) == MASK_INF64 { // check for Infinity
		y = y & (MASK_SIGN64 | MASK_INF64)
	} else { // y is not special
		if (y & MASK_STEERING_BITS64) == MASK_STEERING_BITS64 {
			if ((y & MASK_BINARY_SIG2_64) | MASK_BINARY_OR2_64) > 9999999999999999 {
				y = (y & MASK_SIGN64) | ((y & MASK_BINARY_EXPONENT2_64) << 2)
			}
		}
	}

	// NaN (CASE1)
	if (x & MASK_NAN64) == MASK_NAN64 { // x is NAN
		if (x & MASK_SNAN64) == MASK_SNAN64 { // x is SNaN
			flags |= BID_INVALID_EXCEPTION
			x = x & 0xfdffffffffffffff // quietize x
			res = x
		} else { // x is QNaN
			if (y & MASK_NAN64) == MASK_NAN64 { // y is NAN
				if (y & MASK_SNAN64) == MASK_SNAN64 { // y is SNAN
					flags |= BID_INVALID_EXCEPTION
				}
				res = x
			} else {
				res = y
			}
		}
		return res, flags
	} else if (y & MASK_NAN64) == MASK_NAN64 { // y is NaN, but x is not
		if (y & MASK_SNAN64) == MASK_SNAN64 {
			flags |= BID_INVALID_EXCEPTION
			y = y & 0xfdffffffffffffff // quietize y
			res = y
		} else {
			res = x
		}
		return res, flags
	}

	// SIMPLE (CASE2)
	if x == y {
		res = x
		return res, flags
	}

	// INFINITY (CASE3)
	if (x & MASK_INF64) == MASK_INF64 { // x = +/-infinity
		if (x & MASK_SIGN64) == MASK_SIGN64 { // x = -infinity
			res = y
		} else { // x = +infinity
			res = x
		}
		return res, flags
	} else if (y & MASK_INF64) == MASK_INF64 {
		if (y & MASK_SIGN64) == MASK_SIGN64 {
			res = x
		} else {
			res = y
		}
		return res, flags
	}

	// decode x
	if (x & MASK_STEERING_BITS64) == MASK_STEERING_BITS64 {
		exp_x = int((x & MASK_BINARY_EXPONENT2_64) >> 51)
		sig_x = (x & MASK_BINARY_SIG2_64) | MASK_BINARY_OR2_64
	} else {
		exp_x = int((x & MASK_BINARY_EXPONENT1_64) >> 53)
		sig_x = x & MASK_BINARY_SIG1_64
	}

	// decode y
	if (y & MASK_STEERING_BITS64) == MASK_STEERING_BITS64 {
		exp_y = int((y & MASK_BINARY_EXPONENT2_64) >> 51)
		sig_y = (y & MASK_BINARY_SIG2_64) | MASK_BINARY_OR2_64
	} else {
		exp_y = int((y & MASK_BINARY_EXPONENT1_64) >> 53)
		sig_y = y & MASK_BINARY_SIG1_64
	}

	// ZERO (CASE4)
	if sig_x == 0 {
		x_is_zero = 1
	}
	if sig_y == 0 {
		y_is_zero = 1
	}

	if x_is_zero != 0 && y_is_zero != 0 {
		res = y
		return res, flags
	} else if x_is_zero != 0 {
		if (y & MASK_SIGN64) == MASK_SIGN64 {
			res = x
		} else {
			res = y
		}
		return res, flags
	} else if y_is_zero != 0 {
		if (x & MASK_SIGN64) != MASK_SIGN64 {
			res = x
		} else {
			res = y
		}
		return res, flags
	}

	// OPPOSITE SIGN (CASE5)
	if ((x ^ y) & MASK_SIGN64) == MASK_SIGN64 {
		if (y & MASK_SIGN64) == MASK_SIGN64 {
			res = x
		} else {
			res = y
		}
		return res, flags
	}

	// REDUNDANT REPRESENTATIONS (CASE6)
	if sig_x > sig_y && exp_x >= exp_y {
		if (x & MASK_SIGN64) != MASK_SIGN64 {
			res = x
		} else {
			res = y
		}
		return res, flags
	}
	if sig_x < sig_y && exp_x <= exp_y {
		if (x & MASK_SIGN64) == MASK_SIGN64 {
			res = x
		} else {
			res = y
		}
		return res, flags
	}

	// if exp_x is 15 greater than exp_y, no need for compensation
	if exp_x-exp_y > 15 {
		if (x & MASK_SIGN64) != MASK_SIGN64 {
			res = x
		} else {
			res = y
		}
		return res, flags
	}

	// if exp_x is 15 less than exp_y, no need for compensation
	if exp_y-exp_x > 15 {
		if (x & MASK_SIGN64) == MASK_SIGN64 {
			res = x
		} else {
			res = y
		}
		return res, flags
	}

	// if |exp_x - exp_y| < 15, it comes down to the compensated significand
	if exp_x > exp_y {
		sig_n_prime = __mul_64x64_to_128(sig_x, bid_mult_factor_minmax[exp_x-exp_y])
		if sig_n_prime.w[1] == 0 && sig_n_prime.w[0] == sig_y {
			res = y
			return res, flags
		}

		cond := (sig_n_prime.w[1] > 0) || (sig_n_prime.w[0] > sig_y)
		sign := (x & MASK_SIGN64) == MASK_SIGN64
		if cond != sign {
			res = x
		} else {
			res = y
		}
		return res, flags
	}

	// adjust the y significand upwards
	sig_n_prime = __mul_64x64_to_128(sig_y, bid_mult_factor_minmax[exp_y-exp_x])

	if sig_n_prime.w[1] == 0 && sig_n_prime.w[0] == sig_x {
		res = y
		return res, flags
	}

	cond := (sig_n_prime.w[1] == 0) && (sig_x > sig_n_prime.w[0])
	sign := (x & MASK_SIGN64) == MASK_SIGN64
	if cond != sign {
		res = x
	} else {
		res = y
	}
	return res, flags
}

// Bid64MaxNumMag - returns the number with larger magnitude
// Intel bid64_maxnum_mag 기계적 포팅
func Bid64MaxNumMag(x, y uint64) (uint64, uint32) {
	var res uint64
	var exp_x, exp_y int
	var sig_x, sig_y uint64
	var sig_n_prime BID_UINT128
	var flags uint32 = 0

	// check for non-canonical x
	if (x & MASK_NAN64) == MASK_NAN64 { // x is NaN
		x = x & 0xfe03ffffffffffff // clear G6-G12
		if (x & 0x0003ffffffffffff) > 999999999999999 {
			x = x & 0xfe00000000000000 // clear G6-G12 and the payload bits
		}
	} else if (x & MASK_INF64) == MASK_INF64 { // check for Infinity
		x = x & (MASK_SIGN64 | MASK_INF64)
	} else { // x is not special
		if (x & MASK_STEERING_BITS64) == MASK_STEERING_BITS64 {
			if ((x & MASK_BINARY_SIG2_64) | MASK_BINARY_OR2_64) > 9999999999999999 {
				x = (x & MASK_SIGN64) | ((x & MASK_BINARY_EXPONENT2_64) << 2)
			}
		}
	}

	// check for non-canonical y
	if (y & MASK_NAN64) == MASK_NAN64 { // y is NaN
		y = y & 0xfe03ffffffffffff // clear G6-G12
		if (y & 0x0003ffffffffffff) > 999999999999999 {
			y = y & 0xfe00000000000000 // clear G6-G12 and the payload bits
		}
	} else if (y & MASK_INF64) == MASK_INF64 { // check for Infinity
		y = y & (MASK_SIGN64 | MASK_INF64)
	} else { // y is not special
		if (y & MASK_STEERING_BITS64) == MASK_STEERING_BITS64 {
			if ((y & MASK_BINARY_SIG2_64) | MASK_BINARY_OR2_64) > 9999999999999999 {
				y = (y & MASK_SIGN64) | ((y & MASK_BINARY_EXPONENT2_64) << 2)
			}
		}
	}

	// NaN (CASE1)
	if (x & MASK_NAN64) == MASK_NAN64 { // x is NAN
		if (x & MASK_SNAN64) == MASK_SNAN64 { // x is SNaN
			flags |= BID_INVALID_EXCEPTION
			x = x & 0xfdffffffffffffff // quietize x
			res = x
		} else { // x is QNaN
			if (y & MASK_NAN64) == MASK_NAN64 { // y is NAN
				if (y & MASK_SNAN64) == MASK_SNAN64 { // y is SNAN
					flags |= BID_INVALID_EXCEPTION
				}
				res = x
			} else {
				res = y
			}
		}
		return res, flags
	} else if (y & MASK_NAN64) == MASK_NAN64 { // y is NaN, but x is not
		if (y & MASK_SNAN64) == MASK_SNAN64 {
			flags |= BID_INVALID_EXCEPTION
			y = y & 0xfdffffffffffffff // quietize y
			res = y
		} else {
			res = x
		}
		return res, flags
	}

	// SIMPLE (CASE2)
	if x == y {
		res = x
		return res, flags
	}

	// INFINITY (CASE3)
	if (x & MASK_INF64) == MASK_INF64 {
		// x is infinity, its magnitude is greater than or equal to y
		// return y as long as x isn't negative infinity
		if (x&MASK_SIGN64) == MASK_SIGN64 && (y&MASK_INF64) == MASK_INF64 {
			res = y
		} else {
			res = x
		}
		return res, flags
	} else if (y & MASK_INF64) == MASK_INF64 {
		// y is infinity, then it must be greater in magnitude
		res = y
		return res, flags
	}

	// decode x
	if (x & MASK_STEERING_BITS64) == MASK_STEERING_BITS64 {
		exp_x = int((x & MASK_BINARY_EXPONENT2_64) >> 51)
		sig_x = (x & MASK_BINARY_SIG2_64) | MASK_BINARY_OR2_64
	} else {
		exp_x = int((x & MASK_BINARY_EXPONENT1_64) >> 53)
		sig_x = x & MASK_BINARY_SIG1_64
	}

	// decode y
	if (y & MASK_STEERING_BITS64) == MASK_STEERING_BITS64 {
		exp_y = int((y & MASK_BINARY_EXPONENT2_64) >> 51)
		sig_y = (y & MASK_BINARY_SIG2_64) | MASK_BINARY_OR2_64
	} else {
		exp_y = int((y & MASK_BINARY_EXPONENT1_64) >> 53)
		sig_y = y & MASK_BINARY_SIG1_64
	}

	// ZERO (CASE4)
	if sig_x == 0 {
		res = y // x_is_zero, its magnitude must be smaller than y
		return res, flags
	}
	if sig_y == 0 {
		res = x // y_is_zero, its magnitude must be smaller than x
		return res, flags
	}

	// REDUNDANT REPRESENTATIONS (CASE6)
	if sig_x > sig_y && exp_x >= exp_y {
		res = x
		return res, flags
	}
	if sig_x < sig_y && exp_x <= exp_y {
		res = y
		return res, flags
	}

	// if exp_x is 15 greater than exp_y, no need for compensation
	if exp_x-exp_y > 15 {
		res = x // difference cannot be greater than 10^15
		return res, flags
	}

	// if exp_x is 15 less than exp_y, no need for compensation
	if exp_y-exp_x > 15 {
		res = y
		return res, flags
	}

	// if |exp_x - exp_y| < 15, it comes down to the compensated significand
	if exp_x > exp_y {
		sig_n_prime = __mul_64x64_to_128(sig_x, bid_mult_factor_minmax[exp_x-exp_y])
		if sig_n_prime.w[1] == 0 && sig_n_prime.w[0] == sig_y {
			// two numbers are equal, return maxNum(x,y)
			if (y & MASK_SIGN64) == MASK_SIGN64 {
				res = x
			} else {
				res = y
			}
			return res, flags
		}
		// if compensated_x is greater than y return x, otherwise return y
		if (sig_n_prime.w[1] != 0) || sig_n_prime.w[0] > sig_y {
			res = x
		} else {
			res = y
		}
		return res, flags
	}

	// exp_y must be greater than exp_x, thus adjust the y significand upwards
	sig_n_prime = __mul_64x64_to_128(sig_y, bid_mult_factor_minmax[exp_y-exp_x])

	if sig_n_prime.w[1] == 0 && sig_n_prime.w[0] == sig_x {
		if (y & MASK_SIGN64) == MASK_SIGN64 {
			res = x
		} else {
			res = y
		}
		return res, flags
	}

	if (sig_n_prime.w[1] == 0) && (sig_x > sig_n_prime.w[0]) {
		res = x
	} else {
		res = y
	}
	return res, flags
}
