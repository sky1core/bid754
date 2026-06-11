package bidgo

// bid64_compare.c 기계적 포팅
// Intel BID 라이브러리의 64-bit 비교 함수들

// Bid64QuietEqual - Intel bid64_quiet_equal 기계적 포팅
func Bid64QuietEqual(x, y uint64) (int, uint32) {
	var res int
	var exp_x, exp_y, exp_t int
	var sig_x, sig_y, sig_t uint64
	var pfpsf uint32
	var x_is_zero, y_is_zero, non_canon_x, non_canon_y, lcv int

	// NaN (CASE1)
	// if either number is NAN, the comparison is unordered,
	// rather than equal : return 0
	if ((x & MASK_NAN64) == MASK_NAN64) || ((y & MASK_NAN64) == MASK_NAN64) {
		if (x&MASK_SNAN64) == MASK_SNAN64 || (y&MASK_SNAN64) == MASK_SNAN64 {
			pfpsf |= BID_INVALID_EXCEPTION // set exception if sNaN
		}
		res = 0
		return res, pfpsf
	}
	// SIMPLE (CASE2)
	// if all the bits are the same, these numbers are equivalent.
	if x == y {
		res = 1
		return res, pfpsf
	}
	// INFINITY (CASE3)
	if ((x & MASK_INF64) == MASK_INF64) && ((y & MASK_INF64) == MASK_INF64) {
		res = 0
		if ((x ^ y) & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// ONE INFINITY (CASE3')
	if ((x & MASK_INF64) == MASK_INF64) || ((y & MASK_INF64) == MASK_INF64) {
		res = 0
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
	// some properties:
	// (+ZERO==-ZERO) => therefore ignore the sign
	//    (ZERO x 10^A == ZERO x 10^B) for any valid A, B =>
	//    therefore ignore the exponent field
	//    (Any non-canonical # is considered 0)
	if non_canon_x != 0 || sig_x == 0 {
		x_is_zero = 1
	}
	if non_canon_y != 0 || sig_y == 0 {
		y_is_zero = 1
	}
	if x_is_zero != 0 && y_is_zero != 0 {
		res = 1
		return res, pfpsf
	} else if (x_is_zero != 0 && y_is_zero == 0) || (x_is_zero == 0 && y_is_zero != 0) {
		res = 0
		return res, pfpsf
	}
	// OPPOSITE SIGN (CASE5)
	// now, if the sign bits differ => not equal : return 0
	if ((x ^ y) & MASK_SIGN64) != 0 {
		res = 0
		return res, pfpsf
	}
	// REDUNDANT REPRESENTATIONS (CASE6)
	if exp_x > exp_y { // to simplify the loop below,
		exp_t = exp_x // put the larger exp in y,
		exp_x = exp_y
		exp_y = exp_t
		sig_t = sig_x // and the smaller exp in x
		sig_x = sig_y
		sig_y = sig_t
	}
	if exp_y-exp_x > 15 {
		res = 0 // difference cannot be greater than 10^15
		return res, pfpsf
	}
	for lcv = 0; lcv < (exp_y - exp_x); lcv++ {
		// recalculate y's significand upwards
		sig_y = sig_y * 10
		if sig_y > 9999999999999999 {
			res = 0
			return res, pfpsf
		}
	}
	res = 0
	if sig_y == sig_x {
		res = 1
	}
	return res, pfpsf
}

// Bid64QuietGreater - Intel bid64_quiet_greater 기계적 포팅
func Bid64QuietGreater(x, y uint64) (int, uint32) {
	var res int
	var exp_x, exp_y int
	var sig_x, sig_y uint64
	var sig_n_prime BID_UINT128
	var pfpsf uint32
	var x_is_zero, y_is_zero, non_canon_x, non_canon_y int

	// NaN (CASE1)
	// if either number is NAN, the comparison is unordered, rather than equal :
	// return 0
	if ((x & MASK_NAN64) == MASK_NAN64) || ((y & MASK_NAN64) == MASK_NAN64) {
		if (x&MASK_SNAN64) == MASK_SNAN64 || (y&MASK_SNAN64) == MASK_SNAN64 {
			pfpsf |= BID_INVALID_EXCEPTION // set exception if sNaN
		}
		res = 0
		return res, pfpsf
	}
	// SIMPLE (CASE2)
	// if all the bits are the same, these numbers are equal (not Greater).
	if x == y {
		res = 0
		return res, pfpsf
	}
	// INFINITY (CASE3)
	if (x & MASK_INF64) == MASK_INF64 {
		// if x is neg infinity, there is no way it is greater than y, return 0
		if (x & MASK_SIGN64) == MASK_SIGN64 {
			res = 0
			return res, pfpsf
		} else {
			// x is pos infinity, it is greater, unless y is positive
			// infinity => return y!=pos_infinity
			res = 0
			if ((y & MASK_INF64) != MASK_INF64) || ((y & MASK_SIGN64) == MASK_SIGN64) {
				res = 1
			}
			return res, pfpsf
		}
	} else if (y & MASK_INF64) == MASK_INF64 {
		// x is finite, so if y is positive infinity, then x is less, return 0
		//                 if y is negative infinity, then x is greater, return 1
		res = 0
		if (y & MASK_SIGN64) == MASK_SIGN64 {
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
	// some properties:
	//(+ZERO==-ZERO) => therefore ignore the sign, and neither number is greater
	//(ZERO x 10^A == ZERO x 10^B) for any valid A, B => therefore ignore the
	// exponent field
	// (Any non-canonical # is considered 0)
	if non_canon_x != 0 || sig_x == 0 {
		x_is_zero = 1
	}
	if non_canon_y != 0 || sig_y == 0 {
		y_is_zero = 1
	}
	// if both numbers are zero, neither is greater => return NOTGREATERTHAN
	if x_is_zero != 0 && y_is_zero != 0 {
		res = 0
		return res, pfpsf
	} else if x_is_zero != 0 {
		// is x is zero, it is greater if Y is negative
		res = 0
		if (y & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	} else if y_is_zero != 0 {
		// is y is zero, X is greater if it is positive
		res = 0
		if (x & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// OPPOSITE SIGN (CASE5)
	// now, if the sign bits differ, x is greater if y is negative
	if ((x ^ y) & MASK_SIGN64) == MASK_SIGN64 {
		res = 0
		if (y & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// REDUNDANT REPRESENTATIONS (CASE6)
	// if both components are either bigger or smaller,
	// it is clear what needs to be done
	if sig_x > sig_y && exp_x > exp_y {
		res = 0
		if (x & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	if sig_x < sig_y && exp_x < exp_y {
		res = 0
		if (x & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// if exp_x is 15 greater than exp_y, no need for compensation
	if exp_x-exp_y > 15 { // difference cannot be greater than 10^15
		if (x & MASK_SIGN64) != 0 { // if both are negative
			res = 0
		} else { // if both are positive
			res = 1
		}
		return res, pfpsf
	}
	// if exp_x is 15 less than exp_y, no need for compensation
	if exp_y-exp_x > 15 {
		if (x & MASK_SIGN64) != 0 { // if both are negative
			res = 1
		} else { // if both are positive
			res = 0
		}
		return res, pfpsf
	}
	// if |exp_x - exp_y| < 15, it comes down to the compensated significand
	if exp_x > exp_y { // to simplify the loop below,
		// otherwise adjust the x significand upwards
		sig_n_prime = __mul_64x64_to_128(sig_x, bid_mult_factor[exp_x-exp_y])
		// if postitive, return whichever significand is larger (converse if neg.)
		if sig_n_prime.w[1] == 0 && (sig_n_prime.w[0] == sig_y) {
			res = 0
			return res, pfpsf
		}
		res = 0
		if ((sig_n_prime.w[1] > 0) || sig_n_prime.w[0] > sig_y) != ((x & MASK_SIGN64) == MASK_SIGN64) {
			res = 1
		}
		return res, pfpsf
	}
	// adjust the y significand upwards
	sig_n_prime = __mul_64x64_to_128(sig_y, bid_mult_factor[exp_y-exp_x])
	// if postitive, return whichever significand is larger
	//     (converse if negative)
	if sig_n_prime.w[1] == 0 && (sig_n_prime.w[0] == sig_x) {
		res = 0
		return res, pfpsf
	}
	res = 0
	if ((sig_n_prime.w[1] == 0) && (sig_x > sig_n_prime.w[0])) != ((x & MASK_SIGN64) == MASK_SIGN64) {
		res = 1
	}
	return res, pfpsf
}

// Bid64QuietGreaterEqual - Intel bid64_quiet_greater_equal 기계적 포팅
func Bid64QuietGreaterEqual(x, y uint64) (int, uint32) {
	var res int
	var exp_x, exp_y int
	var sig_x, sig_y uint64
	var sig_n_prime BID_UINT128
	var pfpsf uint32
	var x_is_zero, y_is_zero, non_canon_x, non_canon_y int

	// NaN (CASE1)
	// if either number is NAN, the comparison is unordered : return 1
	if ((x & MASK_NAN64) == MASK_NAN64) || ((y & MASK_NAN64) == MASK_NAN64) {
		if (x&MASK_SNAN64) == MASK_SNAN64 || (y&MASK_SNAN64) == MASK_SNAN64 {
			pfpsf |= BID_INVALID_EXCEPTION // set exception if sNaN
		}
		res = 0
		return res, pfpsf
	}
	// SIMPLE (CASE2)
	// if all the bits are the same, these numbers are equal.
	if x == y {
		res = 1
		return res, pfpsf
	}
	// INFINITY (CASE3)
	if (x & MASK_INF64) == MASK_INF64 {
		// if x==neg_inf, { res = (y == neg_inf)?1:0; BID_RETURN (res) }
		if (x & MASK_SIGN64) == MASK_SIGN64 {
			// x is -inf, so it is less than y unless y is -inf
			res = 0
			if ((y & MASK_INF64) == MASK_INF64) && (y&MASK_SIGN64) == MASK_SIGN64 {
				res = 1
			}
			return res, pfpsf
		} else { // x is pos_inf, no way for it to be less than y
			res = 1
			return res, pfpsf
		}
	} else if (y & MASK_INF64) == MASK_INF64 {
		// x is finite, so:
		//    if y is +inf, x<y
		//    if y is -inf, x>y
		res = 0
		if (y & MASK_SIGN64) == MASK_SIGN64 {
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
	// some properties:
	// (+ZERO==-ZERO) => therefore ignore the sign, and neither number is greater
	// (ZERO x 10^A == ZERO x 10^B) for any valid A, B =>
	//   therefore ignore the exponent field
	//  (Any non-canonical # is considered 0)
	if non_canon_x != 0 || sig_x == 0 {
		x_is_zero = 1
	}
	if non_canon_y != 0 || sig_y == 0 {
		y_is_zero = 1
	}
	if x_is_zero != 0 && y_is_zero != 0 {
		// if both numbers are zero, they are equal
		res = 1
		return res, pfpsf
	} else if x_is_zero != 0 {
		// if x is zero, it is lessthan if Y is positive
		res = 0
		if (y & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	} else if y_is_zero != 0 {
		// if y is zero, X is less if it is negative
		res = 0
		if (x & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// OPPOSITE SIGN (CASE5)
	// now, if the sign bits differ, x is less than if y is positive
	if ((x ^ y) & MASK_SIGN64) == MASK_SIGN64 {
		res = 0
		if (y & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// REDUNDANT REPRESENTATIONS (CASE6)
	// if both components are either bigger or smaller
	if sig_x > sig_y && exp_x >= exp_y {
		res = 0
		if (x & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	if sig_x < sig_y && exp_x <= exp_y {
		res = 0
		if (x & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// if exp_x is 15 greater than exp_y, no need for compensation
	if exp_x-exp_y > 15 {
		res = 0
		if (x & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		// difference cannot be greater than 10^15
		return res, pfpsf
	}
	// if exp_x is 15 less than exp_y, no need for compensation
	if exp_y-exp_x > 15 {
		res = 0
		if (x & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// if |exp_x - exp_y| < 15, it comes down to the compensated significand
	if exp_x > exp_y { // to simplify the loop below,
		// otherwise adjust the x significand upwards
		sig_n_prime = __mul_64x64_to_128(sig_x, bid_mult_factor[exp_x-exp_y])
		// return 1 if values are equal
		if sig_n_prime.w[1] == 0 && (sig_n_prime.w[0] == sig_y) {
			res = 1
			return res, pfpsf
		}
		// if postitive, return whichever significand abs is smaller
		// (converse if negative)
		res = 0
		if ((sig_n_prime.w[1] == 0) && sig_n_prime.w[0] < sig_y) != ((x & MASK_SIGN64) != MASK_SIGN64) {
			res = 1
		}
		return res, pfpsf
	}
	// adjust the y significand upwards
	sig_n_prime = __mul_64x64_to_128(sig_y, bid_mult_factor[exp_y-exp_x])
	// return 0 if values are equal
	if sig_n_prime.w[1] == 0 && (sig_n_prime.w[0] == sig_x) {
		res = 1
		return res, pfpsf
	}
	// if positive, return whichever significand abs is smaller
	// (converse if negative)
	res = 0
	if ((sig_n_prime.w[1] > 0) || (sig_x < sig_n_prime.w[0])) != ((x & MASK_SIGN64) != MASK_SIGN64) {
		res = 1
	}
	return res, pfpsf
}

// Bid64QuietGreaterUnordered - Intel bid64_quiet_greater_unordered 기계적 포팅
func Bid64QuietGreaterUnordered(x, y uint64) (int, uint32) {
	var res int
	var exp_x, exp_y int
	var sig_x, sig_y uint64
	var sig_n_prime BID_UINT128
	var pfpsf uint32
	var x_is_zero, y_is_zero, non_canon_x, non_canon_y int

	// NaN (CASE1)
	// if either number is NAN, the comparison is unordered, rather than equal :
	// return 0
	if ((x & MASK_NAN64) == MASK_NAN64) || ((y & MASK_NAN64) == MASK_NAN64) {
		if (x&MASK_SNAN64) == MASK_SNAN64 || (y&MASK_SNAN64) == MASK_SNAN64 {
			pfpsf |= BID_INVALID_EXCEPTION // set exception if sNaN
		}
		res = 1
		return res, pfpsf
	}
	// SIMPLE (CASE2)
	// if all the bits are the same, these numbers are equal (not Greater).
	if x == y {
		res = 0
		return res, pfpsf
	}
	// INFINITY (CASE3)
	if (x & MASK_INF64) == MASK_INF64 {
		// if x is neg infinity, there is no way it is greater than y, return 0
		if (x & MASK_SIGN64) == MASK_SIGN64 {
			res = 0
			return res, pfpsf
		} else {
			// x is pos infinity, it is greater, unless y is positive infinity =>
			// return y!=pos_infinity
			res = 0
			if ((y & MASK_INF64) != MASK_INF64) || ((y & MASK_SIGN64) == MASK_SIGN64) {
				res = 1
			}
			return res, pfpsf
		}
	} else if (y & MASK_INF64) == MASK_INF64 {
		// x is finite, so if y is positive infinity, then x is less, return 0
		//                 if y is negative infinity, then x is greater, return 1
		res = 0
		if (y & MASK_SIGN64) == MASK_SIGN64 {
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
	// some properties:
	// (+ZERO==-ZERO) => therefore ignore the sign, and neither number is greater
	// (ZERO x 10^A == ZERO x 10^B) for any valid A, B =>
	// therefore ignore the exponent field
	//    (Any non-canonical # is considered 0)
	if non_canon_x != 0 || sig_x == 0 {
		x_is_zero = 1
	}
	if non_canon_y != 0 || sig_y == 0 {
		y_is_zero = 1
	}
	// if both numbers are zero, neither is greater => return NOTGREATERTHAN
	if x_is_zero != 0 && y_is_zero != 0 {
		res = 0
		return res, pfpsf
	} else if x_is_zero != 0 {
		// is x is zero, it is greater if Y is negative
		res = 0
		if (y & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	} else if y_is_zero != 0 {
		// is y is zero, X is greater if it is positive
		res = 0
		if (x & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// OPPOSITE SIGN (CASE5)
	// now, if the sign bits differ, x is greater if y is negative
	if ((x ^ y) & MASK_SIGN64) == MASK_SIGN64 {
		res = 0
		if (y & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// REDUNDANT REPRESENTATIONS (CASE6)
	// if both components are either bigger or smaller
	if sig_x > sig_y && exp_x >= exp_y {
		res = 0
		if (x & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	if sig_x < sig_y && exp_x <= exp_y {
		res = 0
		if (x & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// if exp_x is 15 greater than exp_y, no need for compensation
	if exp_x-exp_y > 15 {
		// difference cannot be greater than 10^15
		res = 0
		if (x & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// if exp_x is 15 less than exp_y, no need for compensation
	if exp_y-exp_x > 15 {
		res = 0
		if (x & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// if |exp_x - exp_y| < 15, it comes down to the compensated significand
	if exp_x > exp_y { // to simplify the loop below,
		// otherwise adjust the x significand upwards
		sig_n_prime = __mul_64x64_to_128(sig_x, bid_mult_factor[exp_x-exp_y])
		// if postitive, return whichever significand is larger
		// (converse if negative)
		if sig_n_prime.w[1] == 0 && (sig_n_prime.w[0] == sig_y) {
			res = 0
			return res, pfpsf
		}
		res = 0
		if ((sig_n_prime.w[1] > 0) || sig_n_prime.w[0] > sig_y) != ((x & MASK_SIGN64) == MASK_SIGN64) {
			res = 1
		}
		return res, pfpsf
	}
	// adjust the y significand upwards
	sig_n_prime = __mul_64x64_to_128(sig_y, bid_mult_factor[exp_y-exp_x])
	// if postitive, return whichever significand is larger (converse if negative)
	if sig_n_prime.w[1] == 0 && (sig_n_prime.w[0] == sig_x) {
		res = 0
		return res, pfpsf
	}
	res = 0
	if ((sig_n_prime.w[1] == 0) && (sig_x > sig_n_prime.w[0])) != ((x & MASK_SIGN64) == MASK_SIGN64) {
		res = 1
	}
	return res, pfpsf
}

// Bid64QuietLess - Intel bid64_quiet_less 기계적 포팅
func Bid64QuietLess(x, y uint64) (int, uint32) {
	var res int
	var exp_x, exp_y int
	var sig_x, sig_y uint64
	var sig_n_prime BID_UINT128
	var pfpsf uint32
	var x_is_zero, y_is_zero, non_canon_x, non_canon_y int

	// NaN (CASE1)
	// if either number is NAN, the comparison is unordered : return 0
	if ((x & MASK_NAN64) == MASK_NAN64) || ((y & MASK_NAN64) == MASK_NAN64) {
		if (x&MASK_SNAN64) == MASK_SNAN64 || (y&MASK_SNAN64) == MASK_SNAN64 {
			pfpsf |= BID_INVALID_EXCEPTION // set exception if sNaN
		}
		res = 0
		return res, pfpsf
	}
	// SIMPLE (CASE2)
	// if all the bits are the same, these numbers are equal.
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
			if ((y & MASK_INF64) != MASK_INF64) || (y&MASK_SIGN64) != MASK_SIGN64 {
				res = 1
			}
			return res, pfpsf
		} else {
			// x is pos_inf, no way for it to be less than y
			res = 0
			return res, pfpsf
		}
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
	// some properties:
	// (+ZERO==-ZERO) => therefore ignore the sign, and neither number is greater
	// (ZERO x 10^A == ZERO x 10^B) for any valid A, B =>
	//  therefore ignore the exponent field
	//    (Any non-canonical # is considered 0)
	if non_canon_x != 0 || sig_x == 0 {
		x_is_zero = 1
	}
	if non_canon_y != 0 || sig_y == 0 {
		y_is_zero = 1
	}
	if x_is_zero != 0 && y_is_zero != 0 {
		// if both numbers are zero, they are equal
		res = 0
		return res, pfpsf
	} else if x_is_zero != 0 {
		// if x is zero, it is lessthan if Y is positive
		res = 0
		if (y & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	} else if y_is_zero != 0 {
		// if y is zero, X is less if it is negative
		res = 0
		if (x & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// OPPOSITE SIGN (CASE5)
	// now, if the sign bits differ, x is less than if y is positive
	if ((x ^ y) & MASK_SIGN64) == MASK_SIGN64 {
		res = 0
		if (y & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// REDUNDANT REPRESENTATIONS (CASE6)
	// if both components are either bigger or smaller,
	// it is clear what needs to be done
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
	// if exp_x is 15 greater than exp_y, no need for compensation
	if exp_x-exp_y > 15 {
		res = 0
		if (x & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		// difference cannot be greater than 10^15
		return res, pfpsf
	}
	// if exp_x is 15 less than exp_y, no need for compensation
	if exp_y-exp_x > 15 {
		res = 0
		if (x & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// if |exp_x - exp_y| < 15, it comes down to the compensated significand
	if exp_x > exp_y { // to simplify the loop below,
		// otherwise adjust the x significand upwards
		sig_n_prime = __mul_64x64_to_128(sig_x, bid_mult_factor[exp_x-exp_y])
		// return 0 if values are equal
		if sig_n_prime.w[1] == 0 && (sig_n_prime.w[0] == sig_y) {
			res = 0
			return res, pfpsf
		}
		// if postitive, return whichever significand abs is smaller
		// (converse if negative)
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
	// (converse if negative)
	res = 0
	if ((sig_n_prime.w[1] > 0) || (sig_x < sig_n_prime.w[0])) != ((x & MASK_SIGN64) == MASK_SIGN64) {
		res = 1
	}
	return res, pfpsf
}

// Bid64QuietLessEqual - Intel bid64_quiet_less_equal 기계적 포팅
func Bid64QuietLessEqual(x, y uint64) (int, uint32) {
	var res int
	var exp_x, exp_y int
	var sig_x, sig_y uint64
	var sig_n_prime BID_UINT128
	var pfpsf uint32
	var x_is_zero, y_is_zero, non_canon_x, non_canon_y int

	// NaN (CASE1)
	// if either number is NAN, the comparison is unordered, rather than equal :
	//     return 0
	if ((x & MASK_NAN64) == MASK_NAN64) || ((y & MASK_NAN64) == MASK_NAN64) {
		if (x&MASK_SNAN64) == MASK_SNAN64 || (y&MASK_SNAN64) == MASK_SNAN64 {
			pfpsf |= BID_INVALID_EXCEPTION // set exception if sNaN
		}
		res = 0
		return res, pfpsf
	}
	// SIMPLE (CASE2)
	// if all the bits are the same, these numbers are equal (LESSEQUAL).
	if x == y {
		res = 1
		return res, pfpsf
	}
	// INFINITY (CASE3)
	if (x & MASK_INF64) == MASK_INF64 {
		if (x & MASK_SIGN64) == MASK_SIGN64 {
			// if x is neg infinity, it must be lessthan or equal to y return 1
			res = 1
			return res, pfpsf
		} else {
			// x is pos infinity, it is greater, unless y is positive infinity =>
			// return y==pos_infinity
			res = 1
			if ((y & MASK_INF64) != MASK_INF64) || ((y & MASK_SIGN64) == MASK_SIGN64) {
				res = 0
			}
			return res, pfpsf
		}
	} else if (y & MASK_INF64) == MASK_INF64 {
		// x is finite, so if y is positive infinity, then x is less, return 1
		//                 if y is negative infinity, then x is greater, return 0
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
	// some properties:
	// (+ZERO==-ZERO) => therefore ignore the sign, and neither number is greater
	// (ZERO x 10^A == ZERO x 10^B) for any valid A, B =>
	//     therefore ignore the exponent field
	//    (Any non-canonical # is considered 0)
	if non_canon_x != 0 || sig_x == 0 {
		x_is_zero = 1
	}
	if non_canon_y != 0 || sig_y == 0 {
		y_is_zero = 1
	}
	if x_is_zero != 0 && y_is_zero != 0 {
		// if both numbers are zero, they are equal -> return 1
		res = 1
		return res, pfpsf
	} else if x_is_zero != 0 {
		// if x is zero, it is lessthan if Y is positive
		res = 0
		if (y & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	} else if y_is_zero != 0 {
		// if y is zero, X is less if it is negative
		res = 0
		if (x & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// OPPOSITE SIGN (CASE5)
	// now, if the sign bits differ, x is less than if y is positive
	if ((x ^ y) & MASK_SIGN64) == MASK_SIGN64 {
		res = 0
		if (y & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// REDUNDANT REPRESENTATIONS (CASE6)
	// if both components are either bigger or smaller
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
	// if exp_x is 15 greater than exp_y, no need for compensation
	if exp_x-exp_y > 15 {
		res = 0
		if (x & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		// difference cannot be greater than 10^15
		return res, pfpsf
	}
	// if exp_x is 15 less than exp_y, no need for compensation
	if exp_y-exp_x > 15 {
		res = 0
		if (x & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// if |exp_x - exp_y| < 15, it comes down to the compensated significand
	if exp_x > exp_y { // to simplify the loop below,
		// otherwise adjust the x significand upwards
		sig_n_prime = __mul_64x64_to_128(sig_x, bid_mult_factor[exp_x-exp_y])
		// return 1 if values are equal
		if sig_n_prime.w[1] == 0 && (sig_n_prime.w[0] == sig_y) {
			res = 1
			return res, pfpsf
		}
		// if postitive, return whichever significand abs is smaller
		//     (converse if negative)
		res = 0
		if ((sig_n_prime.w[1] == 0) && sig_n_prime.w[0] < sig_y) != ((x & MASK_SIGN64) == MASK_SIGN64) {
			res = 1
		}
		return res, pfpsf
	}
	// adjust the y significand upwards
	sig_n_prime = __mul_64x64_to_128(sig_y, bid_mult_factor[exp_y-exp_x])
	// return 1 if values are equal
	if sig_n_prime.w[1] == 0 && (sig_n_prime.w[0] == sig_x) {
		res = 1
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

// Bid64QuietLessUnordered - Intel bid64_quiet_less_unordered 기계적 포팅
func Bid64QuietLessUnordered(x, y uint64) (int, uint32) {
	var res int
	var exp_x, exp_y int
	var sig_x, sig_y uint64
	var sig_n_prime BID_UINT128
	var pfpsf uint32
	var x_is_zero, y_is_zero, non_canon_x, non_canon_y int

	// NaN (CASE1)
	// if either number is NAN, the comparison is unordered : return 0
	if ((x & MASK_NAN64) == MASK_NAN64) || ((y & MASK_NAN64) == MASK_NAN64) {
		if (x&MASK_SNAN64) == MASK_SNAN64 || (y&MASK_SNAN64) == MASK_SNAN64 {
			pfpsf |= BID_INVALID_EXCEPTION // set exception if sNaN
		}
		res = 1
		return res, pfpsf
	}
	// SIMPLE (CASE2)
	// if all the bits are the same, these numbers are equal.
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
			if ((y & MASK_INF64) != MASK_INF64) || (y&MASK_SIGN64) != MASK_SIGN64 {
				res = 1
			}
			return res, pfpsf
		} else {
			// x is pos_inf, no way for it to be less than y
			res = 0
			return res, pfpsf
		}
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
	// some properties:
	// (+ZERO==-ZERO) => therefore ignore the sign, and neither number is greater
	// (ZERO x 10^A == ZERO x 10^B) for any valid A, B =>
	//     therefore ignore the exponent field
	//    (Any non-canonical # is considered 0)
	if non_canon_x != 0 || sig_x == 0 {
		x_is_zero = 1
	}
	if non_canon_y != 0 || sig_y == 0 {
		y_is_zero = 1
	}
	if x_is_zero != 0 && y_is_zero != 0 {
		// if both numbers are zero, they are equal
		res = 0
		return res, pfpsf
	} else if x_is_zero != 0 {
		// if x is zero, it is lessthan if Y is positive
		res = 0
		if (y & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	} else if y_is_zero != 0 {
		// if y is zero, X is less if it is negative
		res = 0
		if (x & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// OPPOSITE SIGN (CASE5)
	// now, if the sign bits differ, x is less than if y is positive
	if ((x ^ y) & MASK_SIGN64) == MASK_SIGN64 {
		res = 0
		if (y & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// REDUNDANT REPRESENTATIONS (CASE6)
	// if both components are either bigger or smaller
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
	// if exp_x is 15 greater than exp_y, no need for compensation
	if exp_x-exp_y > 15 {
		res = 0
		if (x & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		// difference cannot be greater than 10^15
		return res, pfpsf
	}
	// if exp_x is 15 less than exp_y, no need for compensation
	if exp_y-exp_x > 15 {
		res = 0
		if (x & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// if |exp_x - exp_y| < 15, it comes down to the compensated significand
	if exp_x > exp_y { // to simplify the loop below,
		// otherwise adjust the x significand upwards
		sig_n_prime = __mul_64x64_to_128(sig_x, bid_mult_factor[exp_x-exp_y])
		// return 0 if values are equal
		if sig_n_prime.w[1] == 0 && (sig_n_prime.w[0] == sig_y) {
			res = 0
			return res, pfpsf
		}
		// if postitive, return whichever significand abs is smaller
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

// Bid64QuietNotEqual - Intel bid64_quiet_not_equal 기계적 포팅
func Bid64QuietNotEqual(x, y uint64) (int, uint32) {
	var res int
	var exp_x, exp_y, exp_t int
	var sig_x, sig_y, sig_t uint64
	var pfpsf uint32
	var x_is_zero, y_is_zero, non_canon_x, non_canon_y, lcv int

	// NaN (CASE1)
	// if either number is NAN, the comparison is unordered,
	// rather than equal : return 1
	if ((x & MASK_NAN64) == MASK_NAN64) || ((y & MASK_NAN64) == MASK_NAN64) {
		if (x&MASK_SNAN64) == MASK_SNAN64 || (y&MASK_SNAN64) == MASK_SNAN64 {
			pfpsf |= BID_INVALID_EXCEPTION // set exception if sNaN
		}
		res = 1
		return res, pfpsf
	}
	// SIMPLE (CASE2)
	// if all the bits are the same, these numbers are equivalent.
	if x == y {
		res = 0
		return res, pfpsf
	}
	// INFINITY (CASE3)
	if ((x & MASK_INF64) == MASK_INF64) && ((y & MASK_INF64) == MASK_INF64) {
		res = 0
		if ((x ^ y) & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// ONE INFINITY (CASE3')
	if ((x & MASK_INF64) == MASK_INF64) || ((y & MASK_INF64) == MASK_INF64) {
		res = 1
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
	// some properties:
	// (+ZERO==-ZERO) => therefore ignore the sign
	//    (ZERO x 10^A == ZERO x 10^B) for any valid A, B =>
	//        therefore ignore the exponent field
	//    (Any non-canonical # is considered 0)
	if non_canon_x != 0 || sig_x == 0 {
		x_is_zero = 1
	}
	if non_canon_y != 0 || sig_y == 0 {
		y_is_zero = 1
	}

	if x_is_zero != 0 && y_is_zero != 0 {
		res = 0
		return res, pfpsf
	} else if (x_is_zero != 0 && y_is_zero == 0) || (x_is_zero == 0 && y_is_zero != 0) {
		res = 1
		return res, pfpsf
	}
	// OPPOSITE SIGN (CASE5)
	// now, if the sign bits differ => not equal : return 1
	if ((x ^ y) & MASK_SIGN64) != 0 {
		res = 1
		return res, pfpsf
	}
	// REDUNDANT REPRESENTATIONS (CASE6)
	if exp_x > exp_y { // to simplify the loop below,
		exp_t = exp_x // put the larger exp in y,
		exp_x = exp_y
		exp_y = exp_t
		sig_t = sig_x // and the smaller exp in x
		sig_x = sig_y
		sig_y = sig_t
	}

	if exp_y-exp_x > 15 {
		res = 1
		return res, pfpsf
	}
	// difference cannot be greater than 10^16

	for lcv = 0; lcv < (exp_y - exp_x); lcv++ {
		// recalculate y's significand upwards
		sig_y = sig_y * 10
		if sig_y > 9999999999999999 {
			res = 1
			return res, pfpsf
		}
	}

	res = 0
	if sig_y != sig_x {
		res = 1
	}
	return res, pfpsf
}

// Bid64QuietNotGreater - Intel bid64_quiet_not_greater 기계적 포팅
func Bid64QuietNotGreater(x, y uint64) (int, uint32) {
	var res int
	var exp_x, exp_y int
	var sig_x, sig_y uint64
	var sig_n_prime BID_UINT128
	var pfpsf uint32
	var x_is_zero, y_is_zero, non_canon_x, non_canon_y int

	// NaN (CASE1)
	// if either number is NAN, the comparison is unordered,
	//   rather than equal : return 0
	if ((x & MASK_NAN64) == MASK_NAN64) || ((y & MASK_NAN64) == MASK_NAN64) {
		if (x&MASK_SNAN64) == MASK_SNAN64 || (y&MASK_SNAN64) == MASK_SNAN64 {
			pfpsf |= BID_INVALID_EXCEPTION // set exception if sNaN
		}
		res = 1
		return res, pfpsf
	}
	// SIMPLE (CASE2)
	// if all the bits are the same, these numbers are equal (LESSEQUAL).
	if x == y {
		res = 1
		return res, pfpsf
	}
	// INFINITY (CASE3)
	if (x & MASK_INF64) == MASK_INF64 {
		// if x is neg infinity, it must be lessthan or equal to y return 1
		if (x & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
			return res, pfpsf
		}
		// x is pos infinity, it is greater, unless y is positive
		// infinity => return y==pos_infinity
		res = 1
		if ((y & MASK_INF64) != MASK_INF64) || ((y & MASK_SIGN64) == MASK_SIGN64) {
			res = 0
		}
		return res, pfpsf
	} else if (y & MASK_INF64) == MASK_INF64 {
		// x is finite, so if y is positive infinity, then x is less, return 1
		//                 if y is negative infinity, then x is greater, return 0
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
	// some properties:
	// (+ZERO==-ZERO) => therefore ignore the sign, and neither
	//         number is greater
	//    (ZERO x 10^A == ZERO x 10^B) for any valid A, B =>
	//         therefore ignore the exponent field
	//    (Any non-canonical # is considered 0)
	if non_canon_x != 0 || sig_x == 0 {
		x_is_zero = 1
	}
	if non_canon_y != 0 || sig_y == 0 {
		y_is_zero = 1
	}
	// if both numbers are zero, they are equal -> return 1
	if x_is_zero != 0 && y_is_zero != 0 {
		res = 1
		return res, pfpsf
	}
	// if x is zero, it is lessthan if Y is positive
	if x_is_zero != 0 {
		res = 0
		if (y & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// if y is zero, X is less if it is negative
	if y_is_zero != 0 {
		res = 0
		if (x & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// OPPOSITE SIGN (CASE5)
	// now, if the sign bits differ, x is less than if y is positive
	if ((x ^ y) & MASK_SIGN64) == MASK_SIGN64 {
		res = 0
		if (y & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// REDUNDANT REPRESENTATIONS (CASE6)
	// if both components are either bigger or smaller
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
	// if exp_x is 15 greater than exp_y, no need for compensation
	if exp_x-exp_y > 15 {
		res = 0
		if (x & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// difference cannot be greater than 10^15

	// if exp_x is 15 less than exp_y, no need for compensation
	if exp_y-exp_x > 15 {
		res = 0
		if (x & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// if |exp_x - exp_y| < 15, it comes down to the compensated significand
	if exp_x > exp_y { // to simplify the loop below,
		// otherwise adjust the x significand upwards
		sig_n_prime = __mul_64x64_to_128(sig_x, bid_mult_factor[exp_x-exp_y])

		// return 1 if values are equal
		if sig_n_prime.w[1] == 0 && (sig_n_prime.w[0] == sig_y) {
			res = 1
			return res, pfpsf
		}
		// if postitive, return whichever significand abs is smaller
		//     (converse if negative)
		res = 0
		if ((sig_n_prime.w[1] == 0) && sig_n_prime.w[0] < sig_y) != ((x & MASK_SIGN64) == MASK_SIGN64) {
			res = 1
		}
		return res, pfpsf
	}
	// adjust the y significand upwards
	sig_n_prime = __mul_64x64_to_128(sig_y, bid_mult_factor[exp_y-exp_x])

	// return 1 if values are equal
	if sig_n_prime.w[1] == 0 && (sig_n_prime.w[0] == sig_x) {
		res = 1
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

// Bid64QuietNotLess - Intel bid64_quiet_not_less 기계적 포팅
func Bid64QuietNotLess(x, y uint64) (int, uint32) {
	var res int
	var exp_x, exp_y int
	var sig_x, sig_y uint64
	var sig_n_prime BID_UINT128
	var pfpsf uint32
	var x_is_zero, y_is_zero, non_canon_x, non_canon_y int

	// NaN (CASE1)
	// if either number is NAN, the comparison is unordered : return 1
	if ((x & MASK_NAN64) == MASK_NAN64) || ((y & MASK_NAN64) == MASK_NAN64) {
		if (x&MASK_SNAN64) == MASK_SNAN64 || (y&MASK_SNAN64) == MASK_SNAN64 {
			pfpsf |= BID_INVALID_EXCEPTION // set exception if sNaN
		}
		res = 1
		return res, pfpsf
	}
	// SIMPLE (CASE2)
	// if all the bits are the same, these numbers are equal.
	if x == y {
		res = 1
		return res, pfpsf
	}
	// INFINITY (CASE3)
	if (x & MASK_INF64) == MASK_INF64 {
		// if x==neg_inf, { res = (y == neg_inf)?1:0; BID_RETURN (res) }
		if (x & MASK_SIGN64) == MASK_SIGN64 {
			// x is -inf, so it is less than y unless y is -inf
			res = 0
			if ((y & MASK_INF64) == MASK_INF64) && (y&MASK_SIGN64) == MASK_SIGN64 {
				res = 1
			}
			return res, pfpsf
		}
		// x is pos_inf, no way for it to be less than y
		res = 1
		return res, pfpsf
	} else if (y & MASK_INF64) == MASK_INF64 {
		// x is finite, so:
		//    if y is +inf, x<y
		//    if y is -inf, x>y
		res = 0
		if (y & MASK_SIGN64) == MASK_SIGN64 {
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
	// some properties:
	// (+ZERO==-ZERO) => therefore ignore the sign, and neither
	//        number is greater
	//    (ZERO x 10^A == ZERO x 10^B) for any valid A, B =>
	//        therefore ignore the exponent field
	//    (Any non-canonical # is considered 0)
	if non_canon_x != 0 || sig_x == 0 {
		x_is_zero = 1
	}
	if non_canon_y != 0 || sig_y == 0 {
		y_is_zero = 1
	}
	// if both numbers are zero, they are equal
	if x_is_zero != 0 && y_is_zero != 0 {
		res = 1
		return res, pfpsf
	}
	// if x is zero, it is lessthan if Y is positive
	if x_is_zero != 0 {
		res = 0
		if (y & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// if y is zero, X is less if it is negative
	if y_is_zero != 0 {
		res = 0
		if (x & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// OPPOSITE SIGN (CASE5)
	// now, if the sign bits differ, x is less than if y is positive
	if ((x ^ y) & MASK_SIGN64) == MASK_SIGN64 {
		res = 0
		if (y & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// REDUNDANT REPRESENTATIONS (CASE6)
	// if both components are either bigger or smaller
	if sig_x > sig_y && exp_x >= exp_y {
		res = 0
		if (x & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	if sig_x < sig_y && exp_x <= exp_y {
		res = 0
		if (x & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// if exp_x is 15 greater than exp_y, no need for compensation
	if exp_x-exp_y > 15 {
		res = 0
		if (x & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// difference cannot be greater than 10^15

	// if exp_x is 15 less than exp_y, no need for compensation
	if exp_y-exp_x > 15 {
		res = 0
		if (x & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// if |exp_x - exp_y| < 15, it comes down to the compensated significand
	if exp_x > exp_y { // to simplify the loop below,

		// otherwise adjust the x significand upwards
		sig_n_prime = __mul_64x64_to_128(sig_x, bid_mult_factor[exp_x-exp_y])

		// return 0 if values are equal
		if sig_n_prime.w[1] == 0 && (sig_n_prime.w[0] == sig_y) {
			res = 1
			return res, pfpsf
		}
		// if postitive, return whichever significand abs is smaller
		//     (converse if negative)
		res = 0
		if ((sig_n_prime.w[1] == 0) && sig_n_prime.w[0] < sig_y) != ((x & MASK_SIGN64) != MASK_SIGN64) {
			res = 1
		}
		return res, pfpsf
	}
	// adjust the y significand upwards
	sig_n_prime = __mul_64x64_to_128(sig_y, bid_mult_factor[exp_y-exp_x])

	// return 0 if values are equal
	if sig_n_prime.w[1] == 0 && (sig_n_prime.w[0] == sig_x) {
		res = 1
		return res, pfpsf
	}
	// if positive, return whichever significand abs is smaller
	//     (converse if negative)
	res = 0
	if ((sig_n_prime.w[1] > 0) || (sig_x < sig_n_prime.w[0])) != ((x & MASK_SIGN64) != MASK_SIGN64) {
		res = 1
	}
	return res, pfpsf
}

// Bid64QuietOrdered - Intel bid64_quiet_ordered 기계적 포팅
func Bid64QuietOrdered(x, y uint64) (int, uint32) {
	var res int
	var pfpsf uint32

	// NaN (CASE1)
	// if either number is NAN, the comparison is ordered, rather than equal : return 0
	if ((x & MASK_NAN64) == MASK_NAN64) || ((y & MASK_NAN64) == MASK_NAN64) {
		if (x&MASK_SNAN64) == MASK_SNAN64 || (y&MASK_SNAN64) == MASK_SNAN64 {
			pfpsf |= BID_INVALID_EXCEPTION // set exception if sNaN
		}
		res = 0
		return res, pfpsf
	}
	res = 1
	return res, pfpsf
}

// Bid64QuietUnordered - Intel bid64_quiet_unordered 기계적 포팅
func Bid64QuietUnordered(x, y uint64) (int, uint32) {
	var res int
	var pfpsf uint32

	// NaN (CASE1)
	// if either number is NAN, the comparison is unordered,
	//     rather than equal : return 0
	if ((x & MASK_NAN64) == MASK_NAN64) || ((y & MASK_NAN64) == MASK_NAN64) {
		if (x&MASK_SNAN64) == MASK_SNAN64 || (y&MASK_SNAN64) == MASK_SNAN64 {
			pfpsf |= BID_INVALID_EXCEPTION // set exception if sNaN
		}
		res = 1
		return res, pfpsf
	}
	res = 0
	return res, pfpsf
}

// Bid64SignalingGreater - Intel bid64_signaling_greater 기계적 포팅
func Bid64SignalingGreater(x, y uint64) (int, uint32) {
	var res int
	var exp_x, exp_y int
	var sig_x, sig_y uint64
	var sig_n_prime BID_UINT128
	var pfpsf uint32
	var x_is_zero, y_is_zero, non_canon_x, non_canon_y int

	// NaN (CASE1)
	// if either number is NAN, the comparison is unordered,
	//     rather than equal : return 0
	if ((x & MASK_NAN64) == MASK_NAN64) || ((y & MASK_NAN64) == MASK_NAN64) {
		pfpsf |= BID_INVALID_EXCEPTION // set invalid exception if NaN
		res = 0
		return res, pfpsf
	}
	// SIMPLE (CASE2)
	// if all the bits are the same, these numbers are equal (not Greater).
	if x == y {
		res = 0
		return res, pfpsf
	}
	// INFINITY (CASE3)
	if (x & MASK_INF64) == MASK_INF64 {
		// if x is neg infinity, there is no way it is greater than y, return 0
		if (x & MASK_SIGN64) == MASK_SIGN64 {
			res = 0
			return res, pfpsf
		}
		// x is pos infinity, it is greater,
		// unless y is positive infinity => return y!=pos_infinity
		res = 0
		if ((y & MASK_INF64) != MASK_INF64) || ((y & MASK_SIGN64) == MASK_SIGN64) {
			res = 1
		}
		return res, pfpsf
	} else if (y & MASK_INF64) == MASK_INF64 {
		// x is finite, so if y is positive infinity, then x is less, return 0
		//                 if y is negative infinity, then x is greater, return 1
		res = 0
		if (y & MASK_SIGN64) == MASK_SIGN64 {
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
	// some properties:
	// (+ZERO==-ZERO) => therefore ignore the sign, and neither number is greater
	//    (ZERO x 10^A == ZERO x 10^B) for any valid A, B =>
	//      therefore ignore the exponent field
	//    (Any non-canonical # is considered 0)
	if non_canon_x != 0 || sig_x == 0 {
		x_is_zero = 1
	}
	if non_canon_y != 0 || sig_y == 0 {
		y_is_zero = 1
	}
	// if both numbers are zero, neither is greater => return NOTGREATERTHAN
	if x_is_zero != 0 && y_is_zero != 0 {
		res = 0
		return res, pfpsf
	}
	// is x is zero, it is greater if Y is negative
	if x_is_zero != 0 {
		res = 0
		if (y & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// is y is zero, X is greater if it is positive
	if y_is_zero != 0 {
		res = 0
		if (x & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// OPPOSITE SIGN (CASE5)
	// now, if the sign bits differ, x is greater if y is negative
	if ((x ^ y) & MASK_SIGN64) == MASK_SIGN64 {
		res = 0
		if (y & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// REDUNDANT REPRESENTATIONS (CASE6)

	// if both components are either bigger or smaller
	if sig_x > sig_y && exp_x >= exp_y {
		res = 0
		if (x & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	if sig_x < sig_y && exp_x <= exp_y {
		res = 0
		if (x & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// if exp_x is 15 greater than exp_y, no need for compensation
	if exp_x-exp_y > 15 {
		res = 0
		if (x & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// difference cannot be greater than 10^15

	// if exp_x is 15 less than exp_y, no need for compensation
	if exp_y-exp_x > 15 {
		res = 0
		if (x & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// if |exp_x - exp_y| < 15, it comes down to the compensated significand
	if exp_x > exp_y { // to simplify the loop below,

		// otherwise adjust the x significand upwards
		sig_n_prime = __mul_64x64_to_128(sig_x, bid_mult_factor[exp_x-exp_y])

		// if postitive, return whichever significand is larger
		//     (converse if negative)
		if sig_n_prime.w[1] == 0 && (sig_n_prime.w[0] == sig_y) {
			res = 0
			return res, pfpsf
		}

		res = 0
		if ((sig_n_prime.w[1] > 0) || sig_n_prime.w[0] > sig_y) != ((x & MASK_SIGN64) == MASK_SIGN64) {
			res = 1
		}
		return res, pfpsf
	}
	// adjust the y significand upwards
	sig_n_prime = __mul_64x64_to_128(sig_y, bid_mult_factor[exp_y-exp_x])

	// if postitive, return whichever significand is larger
	//     (converse if negative)
	if sig_n_prime.w[1] == 0 && (sig_n_prime.w[0] == sig_x) {
		res = 0
		return res, pfpsf
	}
	res = 0
	if ((sig_n_prime.w[1] == 0) && (sig_x > sig_n_prime.w[0])) != ((x & MASK_SIGN64) == MASK_SIGN64) {
		res = 1
	}
	return res, pfpsf
}

// Bid64SignalingGreaterEqual - Intel bid64_signaling_greater_equal 기계적 포팅
func Bid64SignalingGreaterEqual(x, y uint64) (int, uint32) {
	var res int
	var exp_x, exp_y int
	var sig_x, sig_y uint64
	var sig_n_prime BID_UINT128
	var pfpsf uint32
	var x_is_zero, y_is_zero, non_canon_x, non_canon_y int

	// NaN (CASE1)
	// if either number is NAN, the comparison is unordered : return 1
	if ((x & MASK_NAN64) == MASK_NAN64) || ((y & MASK_NAN64) == MASK_NAN64) {
		pfpsf |= BID_INVALID_EXCEPTION // set invalid exception if NaN
		res = 0
		return res, pfpsf
	}
	// SIMPLE (CASE2)
	// if all the bits are the same, these numbers are equal.
	if x == y {
		res = 1
		return res, pfpsf
	}
	// INFINITY (CASE3)
	if (x & MASK_INF64) == MASK_INF64 {
		// if x==neg_inf, { res = (y == neg_inf)?1:0; BID_RETURN (res) }
		if (x & MASK_SIGN64) == MASK_SIGN64 {
			// x is -inf, so it is less than y unless y is -inf
			res = 0
			if ((y & MASK_INF64) == MASK_INF64) && (y&MASK_SIGN64) == MASK_SIGN64 {
				res = 1
			}
			return res, pfpsf
		}
		// x is pos_inf, no way for it to be less than y
		res = 1
		return res, pfpsf
	} else if (y & MASK_INF64) == MASK_INF64 {
		// x is finite, so:
		//    if y is +inf, x<y
		//    if y is -inf, x>y
		res = 0
		if (y & MASK_SIGN64) == MASK_SIGN64 {
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
	// some properties:
	// (+ZERO==-ZERO) => therefore ignore the sign, and neither number is greater
	//    (ZERO x 10^A == ZERO x 10^B) for any valid A, B =>
	//      therefore ignore the exponent field
	//    (Any non-canonical # is considered 0)
	if non_canon_x != 0 || sig_x == 0 {
		x_is_zero = 1
	}
	if non_canon_y != 0 || sig_y == 0 {
		y_is_zero = 1
	}
	// if both numbers are zero, they are equal
	if x_is_zero != 0 && y_is_zero != 0 {
		res = 1
		return res, pfpsf
	}
	// if x is zero, it is lessthan if Y is positive
	if x_is_zero != 0 {
		res = 0
		if (y & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// if y is zero, X is less if it is negative
	if y_is_zero != 0 {
		res = 0
		if (x & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// OPPOSITE SIGN (CASE5)
	// now, if the sign bits differ, x is less than if y is positive
	if ((x ^ y) & MASK_SIGN64) == MASK_SIGN64 {
		res = 0
		if (y & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// REDUNDANT REPRESENTATIONS (CASE6)
	// if both components are either bigger or smaller
	if sig_x > sig_y && exp_x >= exp_y {
		res = 0
		if (x & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	if sig_x < sig_y && exp_x <= exp_y {
		res = 0
		if (x & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// if exp_x is 15 greater than exp_y, no need for compensation
	if exp_x-exp_y > 15 {
		res = 0
		if (x & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// difference cannot be greater than 10^15

	// if exp_x is 15 less than exp_y, no need for compensation
	if exp_y-exp_x > 15 {
		res = 0
		if (x & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// if |exp_x - exp_y| < 15, it comes down to the compensated significand
	if exp_x > exp_y { // to simplify the loop below,

		// otherwise adjust the x significand upwards
		sig_n_prime = __mul_64x64_to_128(sig_x, bid_mult_factor[exp_x-exp_y])

		// return 1 if values are equal
		if sig_n_prime.w[1] == 0 && (sig_n_prime.w[0] == sig_y) {
			res = 1
			return res, pfpsf
		}
		// if postitive, return whichever significand abs is smaller
		//     (converse if negative)
		res = 0
		if ((sig_n_prime.w[1] == 0) && sig_n_prime.w[0] < sig_y) != ((x & MASK_SIGN64) != MASK_SIGN64) {
			res = 1
		}
		return res, pfpsf
	}
	// adjust the y significand upwards
	sig_n_prime = __mul_64x64_to_128(sig_y, bid_mult_factor[exp_y-exp_x])

	// return 0 if values are equal
	if sig_n_prime.w[1] == 0 && (sig_n_prime.w[0] == sig_x) {
		res = 1
		return res, pfpsf
	}
	// if positive, return whichever significand abs is smaller
	//     (converse if negative)
	res = 0
	if ((sig_n_prime.w[1] > 0) || (sig_x < sig_n_prime.w[0])) != ((x & MASK_SIGN64) != MASK_SIGN64) {
		res = 1
	}
	return res, pfpsf
}

// Bid64SignalingGreaterUnordered - Intel bid64_signaling_greater_unordered 기계적 포팅
func Bid64SignalingGreaterUnordered(x, y uint64) (int, uint32) {
	var res int
	var exp_x, exp_y int
	var sig_x, sig_y uint64
	var sig_n_prime BID_UINT128
	var pfpsf uint32
	var x_is_zero, y_is_zero, non_canon_x, non_canon_y int

	// NaN (CASE1)
	// if either number is NAN, the comparison is unordered,
	// rather than equal : return 0
	if ((x & MASK_NAN64) == MASK_NAN64) || ((y & MASK_NAN64) == MASK_NAN64) {
		pfpsf |= BID_INVALID_EXCEPTION // set invalid exception if NaN
		res = 1
		return res, pfpsf
	}
	// SIMPLE (CASE2)
	// if all the bits are the same, these numbers are equal (not Greater).
	if x == y {
		res = 0
		return res, pfpsf
	}
	// INFINITY (CASE3)
	if (x & MASK_INF64) == MASK_INF64 {
		// if x is neg infinity, there is no way it is greater than y, return 0
		if (x & MASK_SIGN64) == MASK_SIGN64 {
			res = 0
			return res, pfpsf
		}
		// x is pos infinity, it is greater,
		// unless y is positive infinity => return y!=pos_infinity
		res = 0
		if ((y & MASK_INF64) != MASK_INF64) || ((y & MASK_SIGN64) == MASK_SIGN64) {
			res = 1
		}
		return res, pfpsf
	} else if (y & MASK_INF64) == MASK_INF64 {
		// x is finite, so if y is positive infinity, then x is less, return 0
		//                 if y is negative infinity, then x is greater, return 1
		res = 0
		if (y & MASK_SIGN64) == MASK_SIGN64 {
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
	// some properties:
	// (+ZERO==-ZERO) => therefore ignore the sign, and neither number is greater
	//    (ZERO x 10^A == ZERO x 10^B) for any valid A, B =>
	//      therefore ignore the exponent field
	//    (Any non-canonical # is considered 0)
	if non_canon_x != 0 || sig_x == 0 {
		x_is_zero = 1
	}
	if non_canon_y != 0 || sig_y == 0 {
		y_is_zero = 1
	}
	// if both numbers are zero, neither is greater => return NOTGREATERTHAN
	if x_is_zero != 0 && y_is_zero != 0 {
		res = 0
		return res, pfpsf
	}
	// is x is zero, it is greater if Y is negative
	if x_is_zero != 0 {
		res = 0
		if (y & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// is y is zero, X is greater if it is positive
	if y_is_zero != 0 {
		res = 0
		if (x & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// OPPOSITE SIGN (CASE5)
	// now, if the sign bits differ, x is greater if y is negative
	if ((x ^ y) & MASK_SIGN64) == MASK_SIGN64 {
		res = 0
		if (y & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// REDUNDANT REPRESENTATIONS (CASE6)

	// if both components are either bigger or smaller
	if sig_x > sig_y && exp_x >= exp_y {
		res = 0
		if (x & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	if sig_x < sig_y && exp_x <= exp_y {
		res = 0
		if (x & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// if exp_x is 15 greater than exp_y, no need for compensation
	if exp_x-exp_y > 15 {
		res = 0
		if (x & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// difference cannot be greater than 10^15

	// if exp_x is 15 less than exp_y, no need for compensation
	if exp_y-exp_x > 15 {
		res = 0
		if (x & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// if |exp_x - exp_y| < 15, it comes down to the compensated significand
	if exp_x > exp_y { // to simplify the loop below,

		// otherwise adjust the x significand upwards
		sig_n_prime = __mul_64x64_to_128(sig_x, bid_mult_factor[exp_x-exp_y])

		// if postitive, return whichever significand is larger
		//     (converse if negative)
		if sig_n_prime.w[1] == 0 && (sig_n_prime.w[0] == sig_y) {
			res = 0
			return res, pfpsf
		}

		res = 0
		if ((sig_n_prime.w[1] > 0) || sig_n_prime.w[0] > sig_y) != ((x & MASK_SIGN64) == MASK_SIGN64) {
			res = 1
		}
		return res, pfpsf
	}
	// adjust the y significand upwards
	sig_n_prime = __mul_64x64_to_128(sig_y, bid_mult_factor[exp_y-exp_x])

	// if postitive, return whichever significand is larger
	//     (converse if negative)
	if sig_n_prime.w[1] == 0 && (sig_n_prime.w[0] == sig_x) {
		res = 0
		return res, pfpsf
	}
	res = 0
	if ((sig_n_prime.w[1] == 0) && (sig_x > sig_n_prime.w[0])) != ((x & MASK_SIGN64) == MASK_SIGN64) {
		res = 1
	}
	return res, pfpsf
}

// Bid64SignalingLessEqual - Intel bid64_signaling_less_equal 기계적 포팅
func Bid64SignalingLessEqual(x, y uint64) (int, uint32) {
	var res int
	var exp_x, exp_y int
	var sig_x, sig_y uint64
	var sig_n_prime BID_UINT128
	var pfpsf uint32
	var x_is_zero, y_is_zero, non_canon_x, non_canon_y int

	// NaN (CASE1)
	// if either number is NAN, the comparison is unordered,
	// rather than equal : return 0
	if ((x & MASK_NAN64) == MASK_NAN64) || ((y & MASK_NAN64) == MASK_NAN64) {
		pfpsf |= BID_INVALID_EXCEPTION // set invalid exception if NaN
		res = 0
		return res, pfpsf
	}
	// SIMPLE (CASE2)
	// if all the bits are the same, these numbers are equal (LESSEQUAL).
	if x == y {
		res = 1
		return res, pfpsf
	}
	// INFINITY (CASE3)
	if (x & MASK_INF64) == MASK_INF64 {
		// if x is neg infinity, it must be lessthan or equal to y return 1
		if (x & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
			return res, pfpsf
		}
		// x is pos infinity, it is greater,
		// unless y is positive infinity => return y==pos_infinity
		res = 1
		if ((y & MASK_INF64) != MASK_INF64) || ((y & MASK_SIGN64) == MASK_SIGN64) {
			res = 0
		}
		return res, pfpsf
	} else if (y & MASK_INF64) == MASK_INF64 {
		// x is finite, so if y is positive infinity, then x is less, return 1
		//                 if y is negative infinity, then x is greater, return 0
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
	// some properties:
	// (+ZERO==-ZERO) => therefore ignore the sign, and neither number is greater
	//    (ZERO x 10^A == ZERO x 10^B) for any valid A, B =>
	//      therefore ignore the exponent field
	//    (Any non-canonical # is considered 0)
	if non_canon_x != 0 || sig_x == 0 {
		x_is_zero = 1
	}
	if non_canon_y != 0 || sig_y == 0 {
		y_is_zero = 1
	}
	// if both numbers are zero, they are equal -> return 1
	if x_is_zero != 0 && y_is_zero != 0 {
		res = 1
		return res, pfpsf
	}
	// if x is zero, it is lessthan if Y is positive
	if x_is_zero != 0 {
		res = 0
		if (y & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// if y is zero, X is less if it is negative
	if y_is_zero != 0 {
		res = 0
		if (x & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// OPPOSITE SIGN (CASE5)
	// now, if the sign bits differ, x is less than if y is positive
	if ((x ^ y) & MASK_SIGN64) == MASK_SIGN64 {
		res = 0
		if (y & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// REDUNDANT REPRESENTATIONS (CASE6)
	// if both components are either bigger or smaller
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
	// if exp_x is 15 greater than exp_y, no need for compensation
	if exp_x-exp_y > 15 {
		res = 0
		if (x & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// difference cannot be greater than 10^15

	// if exp_x is 15 less than exp_y, no need for compensation
	if exp_y-exp_x > 15 {
		res = 0
		if (x & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// if |exp_x - exp_y| < 15, it comes down to the compensated significand
	if exp_x > exp_y { // to simplify the loop below,

		// otherwise adjust the x significand upwards
		sig_n_prime = __mul_64x64_to_128(sig_x, bid_mult_factor[exp_x-exp_y])

		// return 1 if values are equal
		if sig_n_prime.w[1] == 0 && (sig_n_prime.w[0] == sig_y) {
			res = 1
			return res, pfpsf
		}
		// if postitive, return whichever significand abs is smaller
		//     (converse if negative)
		res = 0
		if ((sig_n_prime.w[1] == 0) && sig_n_prime.w[0] < sig_y) != ((x & MASK_SIGN64) == MASK_SIGN64) {
			res = 1
		}
		return res, pfpsf
	}
	// adjust the y significand upwards
	sig_n_prime = __mul_64x64_to_128(sig_y, bid_mult_factor[exp_y-exp_x])

	// return 1 if values are equal
	if sig_n_prime.w[1] == 0 && (sig_n_prime.w[0] == sig_x) {
		res = 1
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

// Bid64SignalingLessUnordered - Intel bid64_signaling_less_unordered 기계적 포팅
func Bid64SignalingLessUnordered(x, y uint64) (int, uint32) {
	var res int
	var exp_x, exp_y int
	var sig_x, sig_y uint64
	var sig_n_prime BID_UINT128
	var pfpsf uint32
	var x_is_zero, y_is_zero, non_canon_x, non_canon_y int

	// NaN (CASE1)
	// if either number is NAN, the comparison is unordered : return 0
	if ((x & MASK_NAN64) == MASK_NAN64) || ((y & MASK_NAN64) == MASK_NAN64) {
		pfpsf |= BID_INVALID_EXCEPTION // set invalid exception if NaN
		res = 1
		return res, pfpsf
	}
	// SIMPLE (CASE2)
	// if all the bits are the same, these numbers are equal.
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
			if ((y & MASK_INF64) != MASK_INF64) || (y&MASK_SIGN64) != MASK_SIGN64 {
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
	// some properties:
	// (+ZERO==-ZERO) => therefore ignore the sign, and neither number is greater
	//    (ZERO x 10^A == ZERO x 10^B) for any valid A, B =>
	//      therefore ignore the exponent field
	//    (Any non-canonical # is considered 0)
	if non_canon_x != 0 || sig_x == 0 {
		x_is_zero = 1
	}
	if non_canon_y != 0 || sig_y == 0 {
		y_is_zero = 1
	}
	// if both numbers are zero, they are equal
	if x_is_zero != 0 && y_is_zero != 0 {
		res = 0
		return res, pfpsf
	}
	// if x is zero, it is lessthan if Y is positive
	if x_is_zero != 0 {
		res = 0
		if (y & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// if y is zero, X is less if it is negative
	if y_is_zero != 0 {
		res = 0
		if (x & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// OPPOSITE SIGN (CASE5)
	// now, if the sign bits differ, x is less than if y is positive
	if ((x ^ y) & MASK_SIGN64) == MASK_SIGN64 {
		res = 0
		if (y & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// REDUNDANT REPRESENTATIONS (CASE6)
	// if both components are either bigger or smaller
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
	// if exp_x is 15 greater than exp_y, no need for compensation
	if exp_x-exp_y > 15 {
		res = 0
		if (x & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// difference cannot be greater than 10^15

	// if exp_x is 15 less than exp_y, no need for compensation
	if exp_y-exp_x > 15 {
		res = 0
		if (x & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// if |exp_x - exp_y| < 15, it comes down to the compensated significand
	if exp_x > exp_y { // to simplify the loop below,

		// otherwise adjust the x significand upwards
		sig_n_prime = __mul_64x64_to_128(sig_x, bid_mult_factor[exp_x-exp_y])

		// return 0 if values are equal
		if sig_n_prime.w[1] == 0 && (sig_n_prime.w[0] == sig_y) {
			res = 0
			return res, pfpsf
		}
		// if postitive, return whichever significand abs is smaller
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

// Bid64SignalingNotGreater - Intel bid64_signaling_not_greater 기계적 포팅
func Bid64SignalingNotGreater(x, y uint64) (int, uint32) {
	var res int
	var exp_x, exp_y int
	var sig_x, sig_y uint64
	var sig_n_prime BID_UINT128
	var pfpsf uint32
	var x_is_zero, y_is_zero, non_canon_x, non_canon_y int

	// NaN (CASE1)
	// if either number is NAN, the comparison is unordered,
	// rather than equal : return 0
	if ((x & MASK_NAN64) == MASK_NAN64) || ((y & MASK_NAN64) == MASK_NAN64) {
		pfpsf |= BID_INVALID_EXCEPTION // set invalid exception if NaN
		res = 1
		return res, pfpsf
	}
	// SIMPLE (CASE2)
	// if all the bits are the same, these numbers are equal (LESSEQUAL).
	if x == y {
		res = 1
		return res, pfpsf
	}
	// INFINITY (CASE3)
	if (x & MASK_INF64) == MASK_INF64 {
		// if x is neg infinity, it must be lessthan or equal to y return 1
		if (x & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
			return res, pfpsf
		}
		// x is pos infinity, it is greater,
		// unless y is positive infinity => return y==pos_infinity
		res = 1
		if ((y & MASK_INF64) != MASK_INF64) || ((y & MASK_SIGN64) == MASK_SIGN64) {
			res = 0
		}
		return res, pfpsf
	} else if (y & MASK_INF64) == MASK_INF64 {
		// x is finite, so if y is positive infinity, then x is less, return 1
		//                 if y is negative infinity, then x is greater, return 0
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
	// some properties:
	// (+ZERO==-ZERO) => therefore ignore the sign, and neither number is greater
	//    (ZERO x 10^A == ZERO x 10^B) for any valid A, B =>
	//      therefore ignore the exponent field
	//    (Any non-canonical # is considered 0)
	if non_canon_x != 0 || sig_x == 0 {
		x_is_zero = 1
	}
	if non_canon_y != 0 || sig_y == 0 {
		y_is_zero = 1
	}
	// if both numbers are zero, they are equal -> return 1
	if x_is_zero != 0 && y_is_zero != 0 {
		res = 1
		return res, pfpsf
	}
	// if x is zero, it is lessthan if Y is positive
	if x_is_zero != 0 {
		res = 0
		if (y & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// if y is zero, X is less if it is negative
	if y_is_zero != 0 {
		res = 0
		if (x & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// OPPOSITE SIGN (CASE5)
	// now, if the sign bits differ, x is less than if y is positive
	if ((x ^ y) & MASK_SIGN64) == MASK_SIGN64 {
		res = 0
		if (y & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// REDUNDANT REPRESENTATIONS (CASE6)
	// if both components are either bigger or smaller
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
	// if exp_x is 15 greater than exp_y, no need for compensation
	if exp_x-exp_y > 15 {
		res = 0
		if (x & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// difference cannot be greater than 10^15

	// if exp_x is 15 less than exp_y, no need for compensation
	if exp_y-exp_x > 15 {
		res = 0
		if (x & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// if |exp_x - exp_y| < 15, it comes down to the compensated significand
	if exp_x > exp_y { // to simplify the loop below,

		// otherwise adjust the x significand upwards
		sig_n_prime = __mul_64x64_to_128(sig_x, bid_mult_factor[exp_x-exp_y])

		// return 1 if values are equal
		if sig_n_prime.w[1] == 0 && (sig_n_prime.w[0] == sig_y) {
			res = 1
			return res, pfpsf
		}
		// if postitive, return whichever significand abs is smaller
		//     (converse if negative)
		res = 0
		if ((sig_n_prime.w[1] == 0) && sig_n_prime.w[0] < sig_y) != ((x & MASK_SIGN64) == MASK_SIGN64) {
			res = 1
		}
		return res, pfpsf
	}
	// adjust the y significand upwards
	sig_n_prime = __mul_64x64_to_128(sig_y, bid_mult_factor[exp_y-exp_x])

	// return 1 if values are equal
	if sig_n_prime.w[1] == 0 && (sig_n_prime.w[0] == sig_x) {
		res = 1
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

// Bid64SignalingNotLess - Intel bid64_signaling_not_less 기계적 포팅
func Bid64SignalingNotLess(x, y uint64) (int, uint32) {
	var res int
	var exp_x, exp_y int
	var sig_x, sig_y uint64
	var sig_n_prime BID_UINT128
	var pfpsf uint32
	var x_is_zero, y_is_zero, non_canon_x, non_canon_y int

	// NaN (CASE1)
	// if either number is NAN, the comparison is unordered : return 1
	if ((x & MASK_NAN64) == MASK_NAN64) || ((y & MASK_NAN64) == MASK_NAN64) {
		pfpsf |= BID_INVALID_EXCEPTION // set invalid exception if NaN
		res = 1
		return res, pfpsf
	}
	// SIMPLE (CASE2)
	// if all the bits are the same, these numbers are equal.
	if x == y {
		res = 1
		return res, pfpsf
	}
	// INFINITY (CASE3)
	if (x & MASK_INF64) == MASK_INF64 {
		// if x==neg_inf, { res = (y == neg_inf)?1:0; BID_RETURN (res) }
		if (x & MASK_SIGN64) == MASK_SIGN64 {
			// x is -inf, so it is less than y unless y is -inf
			res = 0
			if ((y & MASK_INF64) == MASK_INF64) && (y&MASK_SIGN64) == MASK_SIGN64 {
				res = 1
			}
			return res, pfpsf
		}
		// x is pos_inf, no way for it to be less than y
		res = 1
		return res, pfpsf
	} else if (y & MASK_INF64) == MASK_INF64 {
		// x is finite, so:
		//    if y is +inf, x<y
		//    if y is -inf, x>y
		res = 0
		if (y & MASK_SIGN64) == MASK_SIGN64 {
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
	// some properties:
	// (+ZERO==-ZERO) => therefore ignore the sign, and neither number is greater
	//    (ZERO x 10^A == ZERO x 10^B) for any valid A, B =>
	//      therefore ignore the exponent field
	//    (Any non-canonical # is considered 0)
	if non_canon_x != 0 || sig_x == 0 {
		x_is_zero = 1
	}
	if non_canon_y != 0 || sig_y == 0 {
		y_is_zero = 1
	}
	// if both numbers are zero, they are equal
	if x_is_zero != 0 && y_is_zero != 0 {
		res = 1
		return res, pfpsf
	}
	// if x is zero, it is lessthan if Y is positive
	if x_is_zero != 0 {
		res = 0
		if (y & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// if y is zero, X is less if it is negative
	if y_is_zero != 0 {
		res = 0
		if (x & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// OPPOSITE SIGN (CASE5)
	// now, if the sign bits differ, x is less than if y is positive
	if ((x ^ y) & MASK_SIGN64) == MASK_SIGN64 {
		res = 0
		if (y & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// REDUNDANT REPRESENTATIONS (CASE6)
	// if both components are either bigger or smaller
	if sig_x > sig_y && exp_x >= exp_y {
		res = 0
		if (x & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	if sig_x < sig_y && exp_x <= exp_y {
		res = 0
		if (x & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// if exp_x is 15 greater than exp_y, no need for compensation
	if exp_x-exp_y > 15 {
		res = 0
		if (x & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// difference cannot be greater than 10^15

	// if exp_x is 15 less than exp_y, no need for compensation
	if exp_y-exp_x > 15 {
		res = 0
		if (x & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// if |exp_x - exp_y| < 15, it comes down to the compensated significand
	if exp_x > exp_y { // to simplify the loop below,

		// otherwise adjust the x significand upwards
		sig_n_prime = __mul_64x64_to_128(sig_x, bid_mult_factor[exp_x-exp_y])

		// return 0 if values are equal
		if sig_n_prime.w[1] == 0 && (sig_n_prime.w[0] == sig_y) {
			res = 1
			return res, pfpsf
		}
		// if postitive, return whichever significand abs is smaller
		//     (converse if negative)
		res = 0
		if ((sig_n_prime.w[1] == 0) && sig_n_prime.w[0] < sig_y) != ((x & MASK_SIGN64) != MASK_SIGN64) {
			res = 1
		}
		return res, pfpsf
	}
	// adjust the y significand upwards
	sig_n_prime = __mul_64x64_to_128(sig_y, bid_mult_factor[exp_y-exp_x])

	// return 0 if values are equal
	if sig_n_prime.w[1] == 0 && (sig_n_prime.w[0] == sig_x) {
		res = 1
		return res, pfpsf
	}
	// if positive, return whichever significand abs is smaller
	//     (converse if negative)
	res = 0
	if ((sig_n_prime.w[1] > 0) || (sig_x < sig_n_prime.w[0])) != ((x & MASK_SIGN64) != MASK_SIGN64) {
		res = 1
	}
	return res, pfpsf
}
