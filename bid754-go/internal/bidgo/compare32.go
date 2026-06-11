package bidgo

// bid32_compare.c 기계적 포팅
// Intel BID 라이브러리의 64-bit 비교 함수들

// Bid32QuietEqual - Intel bid64_quiet_equal 기계적 포팅
func Bid32QuietEqual(x, y uint32) (int, uint32) {
	var res int
	var exp_x, exp_y, exp_t int
	var sig_x, sig_y, sig_t uint32
	var pfpsf uint32
	var x_is_zero, y_is_zero, non_canon_x, non_canon_y, lcv int

	// NaN (CASE1)
	// if either number is NAN, the comparison is unordered,
	// rather than equal : return 0
	if ((x & MASK_NAN32) == MASK_NAN32) || ((y & MASK_NAN32) == MASK_NAN32) {
		if (x&MASK_SNAN32) == MASK_SNAN32 || (y&MASK_SNAN32) == MASK_SNAN32 {
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
	if ((x & MASK_INF32) == MASK_INF32) && ((y & MASK_INF32) == MASK_INF32) {
		res = 0
		if ((x ^ y) & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// ONE INFINITY (CASE3')
	if ((x & MASK_INF32) == MASK_INF32) || ((y & MASK_INF32) == MASK_INF32) {
		res = 0
		return res, pfpsf
	}
	// if steering bits are 11 (condition will be 0), then exponent is G[0:w+1] =>
	if (x & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
		exp_x = int((x & MASK_BINARY_EXPONENT2_32) >> 21)
		sig_x = (x & MASK_BINARY_SIG2_32) | MASK_BINARY_OR2_32
		if sig_x > 9999999 {
			non_canon_x = 1
		} else {
			non_canon_x = 0
		}
	} else {
		exp_x = int((x & MASK_BINARY_EXPONENT1_32) >> 23)
		sig_x = x & MASK_BINARY_SIG1_32
		non_canon_x = 0
	}
	// if steering bits are 11 (condition will be 0), then exponent is G[0:w+1] =>
	if (y & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
		exp_y = int((y & MASK_BINARY_EXPONENT2_32) >> 21)
		sig_y = (y & MASK_BINARY_SIG2_32) | MASK_BINARY_OR2_32
		if sig_y > 9999999 {
			non_canon_y = 1
		} else {
			non_canon_y = 0
		}
	} else {
		exp_y = int((y & MASK_BINARY_EXPONENT1_32) >> 23)
		sig_y = y & MASK_BINARY_SIG1_32
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
	if ((x ^ y) & MASK_SIGN32) != 0 {
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
	if exp_y-exp_x > 6 {
		res = 0 // difference cannot be greater than 10^15
		return res, pfpsf
	}
	for lcv = 0; lcv < (exp_y - exp_x); lcv++ {
		// recalculate y's significand upwards
		sig_y = sig_y * 10
		if sig_y > 9999999 {
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

// Bid32QuietGreater - Intel bid64_quiet_greater 기계적 포팅
func Bid32QuietGreater(x, y uint32) (int, uint32) {
	var res int
	var exp_x, exp_y int
	var sig_x, sig_y uint32
	var sig_n_prime uint64
	var pfpsf uint32
	var x_is_zero, y_is_zero, non_canon_x, non_canon_y int

	// NaN (CASE1)
	// if either number is NAN, the comparison is unordered, rather than equal :
	// return 0
	if ((x & MASK_NAN32) == MASK_NAN32) || ((y & MASK_NAN32) == MASK_NAN32) {
		if (x&MASK_SNAN32) == MASK_SNAN32 || (y&MASK_SNAN32) == MASK_SNAN32 {
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
	if (x & MASK_INF32) == MASK_INF32 {
		// if x is neg infinity, there is no way it is greater than y, return 0
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			res = 0
			return res, pfpsf
		} else {
			// x is pos infinity, it is greater, unless y is positive
			// infinity => return y!=pos_infinity
			res = 0
			if ((y & MASK_INF32) != MASK_INF32) || ((y & MASK_SIGN32) == MASK_SIGN32) {
				res = 1
			}
			return res, pfpsf
		}
	} else if (y & MASK_INF32) == MASK_INF32 {
		// x is finite, so if y is positive infinity, then x is less, return 0
		//                 if y is negative infinity, then x is greater, return 1
		res = 0
		if (y & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// if steering bits are 11 (condition will be 0), then exponent is G[0:w+1] =>
	if (x & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
		exp_x = int((x & MASK_BINARY_EXPONENT2_32) >> 21)
		sig_x = (x & MASK_BINARY_SIG2_32) | MASK_BINARY_OR2_32
		if sig_x > 9999999 {
			non_canon_x = 1
		} else {
			non_canon_x = 0
		}
	} else {
		exp_x = int((x & MASK_BINARY_EXPONENT1_32) >> 23)
		sig_x = x & MASK_BINARY_SIG1_32
		non_canon_x = 0
	}
	// if steering bits are 11 (condition will be 0), then exponent is G[0:w+1] =>
	if (y & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
		exp_y = int((y & MASK_BINARY_EXPONENT2_32) >> 21)
		sig_y = (y & MASK_BINARY_SIG2_32) | MASK_BINARY_OR2_32
		if sig_y > 9999999 {
			non_canon_y = 1
		} else {
			non_canon_y = 0
		}
	} else {
		exp_y = int((y & MASK_BINARY_EXPONENT1_32) >> 23)
		sig_y = y & MASK_BINARY_SIG1_32
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
		if (y & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	} else if y_is_zero != 0 {
		// is y is zero, X is greater if it is positive
		res = 0
		if (x & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// OPPOSITE SIGN (CASE5)
	// now, if the sign bits differ, x is greater if y is negative
	if ((x ^ y) & MASK_SIGN32) == MASK_SIGN32 {
		res = 0
		if (y & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// REDUNDANT REPRESENTATIONS (CASE6)
	// if both components are either bigger or smaller,
	// it is clear what needs to be done
	if sig_x > sig_y && exp_x > exp_y {
		res = 0
		if (x & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	if sig_x < sig_y && exp_x < exp_y {
		res = 0
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// if exp_x is 15 greater than exp_y, no need for compensation
	if exp_x-exp_y > 6 { // difference cannot be greater than 10^15
		if (x & MASK_SIGN32) != 0 { // if both are negative
			res = 0
		} else { // if both are positive
			res = 1
		}
		return res, pfpsf
	}
	// if exp_x is 15 less than exp_y, no need for compensation
	if exp_y-exp_x > 6 {
		if (x & MASK_SIGN32) != 0 { // if both are negative
			res = 1
		} else { // if both are positive
			res = 0
		}
		return res, pfpsf
	}
	// if |exp_x - exp_y| < 15, it comes down to the compensated significand
	if exp_x > exp_y { // to simplify the loop below,
		// otherwise adjust the x significand upwards
		sig_n_prime = uint64(sig_x) * bid32_mult_factor[exp_x-exp_y]
		// if postitive, return whichever significand is larger (converse if neg.)
		if 0 == 0 && (sig_n_prime == uint64(sig_y)) {
			res = 0
			return res, pfpsf
		}
		res = 0
		if ((0 > 0) || sig_n_prime > uint64(sig_y)) != ((x & MASK_SIGN32) == MASK_SIGN32) {
			res = 1
		}
		return res, pfpsf
	}
	// adjust the y significand upwards
	sig_n_prime = uint64(sig_y) * bid32_mult_factor[exp_y-exp_x]
	// if postitive, return whichever significand is larger
	//     (converse if negative)
	if 0 == 0 && (sig_n_prime == uint64(sig_x)) {
		res = 0
		return res, pfpsf
	}
	res = 0
	if (uint64(sig_x) > sig_n_prime) != ((x & MASK_SIGN32) == MASK_SIGN32) {
		res = 1
	}
	return res, pfpsf
}

// Bid32QuietGreaterEqual - Intel bid64_quiet_greater_equal 기계적 포팅
func Bid32QuietGreaterEqual(x, y uint32) (int, uint32) {
	var res int
	var exp_x, exp_y int
	var sig_x, sig_y uint32
	var sig_n_prime uint64
	var pfpsf uint32
	var x_is_zero, y_is_zero, non_canon_x, non_canon_y int

	// NaN (CASE1)
	// if either number is NAN, the comparison is unordered : return 1
	if ((x & MASK_NAN32) == MASK_NAN32) || ((y & MASK_NAN32) == MASK_NAN32) {
		if (x&MASK_SNAN32) == MASK_SNAN32 || (y&MASK_SNAN32) == MASK_SNAN32 {
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
	if (x & MASK_INF32) == MASK_INF32 {
		// if x==neg_inf, { res = (y == neg_inf)?1:0; BID_RETURN (res) }
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			// x is -inf, so it is less than y unless y is -inf
			res = 0
			if ((y & MASK_INF32) == MASK_INF32) && (y&MASK_SIGN32) == MASK_SIGN32 {
				res = 1
			}
			return res, pfpsf
		} else { // x is pos_inf, no way for it to be less than y
			res = 1
			return res, pfpsf
		}
	} else if (y & MASK_INF32) == MASK_INF32 {
		// x is finite, so:
		//    if y is +inf, x<y
		//    if y is -inf, x>y
		res = 0
		if (y & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// if steering bits are 11 (condition will be 0), then exponent is G[0:w+1] =>
	if (x & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
		exp_x = int((x & MASK_BINARY_EXPONENT2_32) >> 21)
		sig_x = (x & MASK_BINARY_SIG2_32) | MASK_BINARY_OR2_32
		if sig_x > 9999999 {
			non_canon_x = 1
		} else {
			non_canon_x = 0
		}
	} else {
		exp_x = int((x & MASK_BINARY_EXPONENT1_32) >> 23)
		sig_x = x & MASK_BINARY_SIG1_32
		non_canon_x = 0
	}
	// if steering bits are 11 (condition will be 0), then exponent is G[0:w+1] =>
	if (y & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
		exp_y = int((y & MASK_BINARY_EXPONENT2_32) >> 21)
		sig_y = (y & MASK_BINARY_SIG2_32) | MASK_BINARY_OR2_32
		if sig_y > 9999999 {
			non_canon_y = 1
		} else {
			non_canon_y = 0
		}
	} else {
		exp_y = int((y & MASK_BINARY_EXPONENT1_32) >> 23)
		sig_y = y & MASK_BINARY_SIG1_32
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
		if (y & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	} else if y_is_zero != 0 {
		// if y is zero, X is less if it is negative
		res = 0
		if (x & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// OPPOSITE SIGN (CASE5)
	// now, if the sign bits differ, x is less than if y is positive
	if ((x ^ y) & MASK_SIGN32) == MASK_SIGN32 {
		res = 0
		if (y & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// REDUNDANT REPRESENTATIONS (CASE6)
	// if both components are either bigger or smaller
	if sig_x > sig_y && exp_x >= exp_y {
		res = 0
		if (x & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	if sig_x < sig_y && exp_x <= exp_y {
		res = 0
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// if exp_x is 15 greater than exp_y, no need for compensation
	if exp_x-exp_y > 6 {
		res = 0
		if (x & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		// difference cannot be greater than 10^15
		return res, pfpsf
	}
	// if exp_x is 15 less than exp_y, no need for compensation
	if exp_y-exp_x > 6 {
		res = 0
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// if |exp_x - exp_y| < 15, it comes down to the compensated significand
	if exp_x > exp_y { // to simplify the loop below,
		// otherwise adjust the x significand upwards
		sig_n_prime = uint64(sig_x) * bid32_mult_factor[exp_x-exp_y]
		// return 1 if values are equal
		if 0 == 0 && (sig_n_prime == uint64(sig_y)) {
			res = 1
			return res, pfpsf
		}
		// if postitive, return whichever significand abs is smaller
		// (converse if negative)
		res = 0
		if (sig_n_prime < uint64(sig_y)) != ((x & MASK_SIGN32) != MASK_SIGN32) {
			res = 1
		}
		return res, pfpsf
	}
	// adjust the y significand upwards
	sig_n_prime = uint64(sig_y) * bid32_mult_factor[exp_y-exp_x]
	// return 0 if values are equal
	if 0 == 0 && (sig_n_prime == uint64(sig_x)) {
		res = 1
		return res, pfpsf
	}
	// if positive, return whichever significand abs is smaller
	// (converse if negative)
	res = 0
	if (uint64(sig_x) < sig_n_prime) != ((x & MASK_SIGN32) != MASK_SIGN32) {
		res = 1
	}
	return res, pfpsf
}

// Bid32QuietGreaterUnordered - Intel bid64_quiet_greater_unordered 기계적 포팅
func Bid32QuietGreaterUnordered(x, y uint32) (int, uint32) {
	var res int
	var exp_x, exp_y int
	var sig_x, sig_y uint32
	var sig_n_prime uint64
	var pfpsf uint32
	var x_is_zero, y_is_zero, non_canon_x, non_canon_y int

	// NaN (CASE1)
	// if either number is NAN, the comparison is unordered, rather than equal :
	// return 0
	if ((x & MASK_NAN32) == MASK_NAN32) || ((y & MASK_NAN32) == MASK_NAN32) {
		if (x&MASK_SNAN32) == MASK_SNAN32 || (y&MASK_SNAN32) == MASK_SNAN32 {
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
	if (x & MASK_INF32) == MASK_INF32 {
		// if x is neg infinity, there is no way it is greater than y, return 0
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			res = 0
			return res, pfpsf
		} else {
			// x is pos infinity, it is greater, unless y is positive infinity =>
			// return y!=pos_infinity
			res = 0
			if ((y & MASK_INF32) != MASK_INF32) || ((y & MASK_SIGN32) == MASK_SIGN32) {
				res = 1
			}
			return res, pfpsf
		}
	} else if (y & MASK_INF32) == MASK_INF32 {
		// x is finite, so if y is positive infinity, then x is less, return 0
		//                 if y is negative infinity, then x is greater, return 1
		res = 0
		if (y & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// if steering bits are 11 (condition will be 0), then exponent is G[0:w+1] =>
	if (x & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
		exp_x = int((x & MASK_BINARY_EXPONENT2_32) >> 21)
		sig_x = (x & MASK_BINARY_SIG2_32) | MASK_BINARY_OR2_32
		if sig_x > 9999999 {
			non_canon_x = 1
		} else {
			non_canon_x = 0
		}
	} else {
		exp_x = int((x & MASK_BINARY_EXPONENT1_32) >> 23)
		sig_x = x & MASK_BINARY_SIG1_32
		non_canon_x = 0
	}
	// if steering bits are 11 (condition will be 0), then exponent is G[0:w+1] =>
	if (y & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
		exp_y = int((y & MASK_BINARY_EXPONENT2_32) >> 21)
		sig_y = (y & MASK_BINARY_SIG2_32) | MASK_BINARY_OR2_32
		if sig_y > 9999999 {
			non_canon_y = 1
		} else {
			non_canon_y = 0
		}
	} else {
		exp_y = int((y & MASK_BINARY_EXPONENT1_32) >> 23)
		sig_y = y & MASK_BINARY_SIG1_32
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
		if (y & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	} else if y_is_zero != 0 {
		// is y is zero, X is greater if it is positive
		res = 0
		if (x & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// OPPOSITE SIGN (CASE5)
	// now, if the sign bits differ, x is greater if y is negative
	if ((x ^ y) & MASK_SIGN32) == MASK_SIGN32 {
		res = 0
		if (y & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// REDUNDANT REPRESENTATIONS (CASE6)
	// if both components are either bigger or smaller
	if sig_x > sig_y && exp_x >= exp_y {
		res = 0
		if (x & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	if sig_x < sig_y && exp_x <= exp_y {
		res = 0
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// if exp_x is 15 greater than exp_y, no need for compensation
	if exp_x-exp_y > 6 {
		// difference cannot be greater than 10^15
		res = 0
		if (x & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// if exp_x is 15 less than exp_y, no need for compensation
	if exp_y-exp_x > 6 {
		res = 0
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// if |exp_x - exp_y| < 15, it comes down to the compensated significand
	if exp_x > exp_y { // to simplify the loop below,
		// otherwise adjust the x significand upwards
		sig_n_prime = uint64(sig_x) * bid32_mult_factor[exp_x-exp_y]
		// if postitive, return whichever significand is larger
		// (converse if negative)
		if 0 == 0 && (sig_n_prime == uint64(sig_y)) {
			res = 0
			return res, pfpsf
		}
		res = 0
		if ((0 > 0) || sig_n_prime > uint64(sig_y)) != ((x & MASK_SIGN32) == MASK_SIGN32) {
			res = 1
		}
		return res, pfpsf
	}
	// adjust the y significand upwards
	sig_n_prime = uint64(sig_y) * bid32_mult_factor[exp_y-exp_x]
	// if postitive, return whichever significand is larger (converse if negative)
	if 0 == 0 && (sig_n_prime == uint64(sig_x)) {
		res = 0
		return res, pfpsf
	}
	res = 0
	if (uint64(sig_x) > sig_n_prime) != ((x & MASK_SIGN32) == MASK_SIGN32) {
		res = 1
	}
	return res, pfpsf
}

// Bid32QuietLess - Intel bid64_quiet_less 기계적 포팅
func Bid32QuietLess(x, y uint32) (int, uint32) {
	var res int
	var exp_x, exp_y int
	var sig_x, sig_y uint32
	var sig_n_prime uint64
	var pfpsf uint32
	var x_is_zero, y_is_zero, non_canon_x, non_canon_y int

	// NaN (CASE1)
	// if either number is NAN, the comparison is unordered : return 0
	if ((x & MASK_NAN32) == MASK_NAN32) || ((y & MASK_NAN32) == MASK_NAN32) {
		if (x&MASK_SNAN32) == MASK_SNAN32 || (y&MASK_SNAN32) == MASK_SNAN32 {
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
	if (x & MASK_INF32) == MASK_INF32 {
		// if x==neg_inf, { res = (y == neg_inf)?0:1; BID_RETURN (res) }
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			// x is -inf, so it is less than y unless y is -inf
			res = 0
			if ((y & MASK_INF32) != MASK_INF32) || (y&MASK_SIGN32) != MASK_SIGN32 {
				res = 1
			}
			return res, pfpsf
		} else {
			// x is pos_inf, no way for it to be less than y
			res = 0
			return res, pfpsf
		}
	} else if (y & MASK_INF32) == MASK_INF32 {
		// x is finite, so:
		//    if y is +inf, x<y
		//    if y is -inf, x>y
		res = 0
		if (y & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// if steering bits are 11 (condition will be 0), then exponent is G[0:w+1] =>
	if (x & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
		exp_x = int((x & MASK_BINARY_EXPONENT2_32) >> 21)
		sig_x = (x & MASK_BINARY_SIG2_32) | MASK_BINARY_OR2_32
		if sig_x > 9999999 {
			non_canon_x = 1
		} else {
			non_canon_x = 0
		}
	} else {
		exp_x = int((x & MASK_BINARY_EXPONENT1_32) >> 23)
		sig_x = x & MASK_BINARY_SIG1_32
		non_canon_x = 0
	}
	// if steering bits are 11 (condition will be 0), then exponent is G[0:w+1] =>
	if (y & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
		exp_y = int((y & MASK_BINARY_EXPONENT2_32) >> 21)
		sig_y = (y & MASK_BINARY_SIG2_32) | MASK_BINARY_OR2_32
		if sig_y > 9999999 {
			non_canon_y = 1
		} else {
			non_canon_y = 0
		}
	} else {
		exp_y = int((y & MASK_BINARY_EXPONENT1_32) >> 23)
		sig_y = y & MASK_BINARY_SIG1_32
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
		if (y & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	} else if y_is_zero != 0 {
		// if y is zero, X is less if it is negative
		res = 0
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// OPPOSITE SIGN (CASE5)
	// now, if the sign bits differ, x is less than if y is positive
	if ((x ^ y) & MASK_SIGN32) == MASK_SIGN32 {
		res = 0
		if (y & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// REDUNDANT REPRESENTATIONS (CASE6)
	// if both components are either bigger or smaller,
	// it is clear what needs to be done
	if sig_x > sig_y && exp_x >= exp_y {
		res = 0
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	if sig_x < sig_y && exp_x <= exp_y {
		res = 0
		if (x & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// if exp_x is 15 greater than exp_y, no need for compensation
	if exp_x-exp_y > 6 {
		res = 0
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		// difference cannot be greater than 10^15
		return res, pfpsf
	}
	// if exp_x is 15 less than exp_y, no need for compensation
	if exp_y-exp_x > 6 {
		res = 0
		if (x & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// if |exp_x - exp_y| < 15, it comes down to the compensated significand
	if exp_x > exp_y { // to simplify the loop below,
		// otherwise adjust the x significand upwards
		sig_n_prime = uint64(sig_x) * bid32_mult_factor[exp_x-exp_y]
		// return 0 if values are equal
		if 0 == 0 && (sig_n_prime == uint64(sig_y)) {
			res = 0
			return res, pfpsf
		}
		// if postitive, return whichever significand abs is smaller
		// (converse if negative)
		res = 0
		if (sig_n_prime < uint64(sig_y)) != ((x & MASK_SIGN32) == MASK_SIGN32) {
			res = 1
		}
		return res, pfpsf
	}
	// adjust the y significand upwards
	sig_n_prime = uint64(sig_y) * bid32_mult_factor[exp_y-exp_x]
	// return 0 if values are equal
	if 0 == 0 && (sig_n_prime == uint64(sig_x)) {
		res = 0
		return res, pfpsf
	}
	// if positive, return whichever significand abs is smaller
	// (converse if negative)
	res = 0
	if (uint64(sig_x) < sig_n_prime) != ((x & MASK_SIGN32) == MASK_SIGN32) {
		res = 1
	}
	return res, pfpsf
}

// Bid32QuietLessEqual - Intel bid64_quiet_less_equal 기계적 포팅
func Bid32QuietLessEqual(x, y uint32) (int, uint32) {
	var res int
	var exp_x, exp_y int
	var sig_x, sig_y uint32
	var sig_n_prime uint64
	var pfpsf uint32
	var x_is_zero, y_is_zero, non_canon_x, non_canon_y int

	// NaN (CASE1)
	// if either number is NAN, the comparison is unordered, rather than equal :
	//     return 0
	if ((x & MASK_NAN32) == MASK_NAN32) || ((y & MASK_NAN32) == MASK_NAN32) {
		if (x&MASK_SNAN32) == MASK_SNAN32 || (y&MASK_SNAN32) == MASK_SNAN32 {
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
	if (x & MASK_INF32) == MASK_INF32 {
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			// if x is neg infinity, it must be lessthan or equal to y return 1
			res = 1
			return res, pfpsf
		} else {
			// x is pos infinity, it is greater, unless y is positive infinity =>
			// return y==pos_infinity
			res = 1
			if ((y & MASK_INF32) != MASK_INF32) || ((y & MASK_SIGN32) == MASK_SIGN32) {
				res = 0
			}
			return res, pfpsf
		}
	} else if (y & MASK_INF32) == MASK_INF32 {
		// x is finite, so if y is positive infinity, then x is less, return 1
		//                 if y is negative infinity, then x is greater, return 0
		res = 0
		if (y & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// if steering bits are 11 (condition will be 0), then exponent is G[0:w+1] =>
	if (x & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
		exp_x = int((x & MASK_BINARY_EXPONENT2_32) >> 21)
		sig_x = (x & MASK_BINARY_SIG2_32) | MASK_BINARY_OR2_32
		if sig_x > 9999999 {
			non_canon_x = 1
		} else {
			non_canon_x = 0
		}
	} else {
		exp_x = int((x & MASK_BINARY_EXPONENT1_32) >> 23)
		sig_x = x & MASK_BINARY_SIG1_32
		non_canon_x = 0
	}
	// if steering bits are 11 (condition will be 0), then exponent is G[0:w+1] =>
	if (y & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
		exp_y = int((y & MASK_BINARY_EXPONENT2_32) >> 21)
		sig_y = (y & MASK_BINARY_SIG2_32) | MASK_BINARY_OR2_32
		if sig_y > 9999999 {
			non_canon_y = 1
		} else {
			non_canon_y = 0
		}
	} else {
		exp_y = int((y & MASK_BINARY_EXPONENT1_32) >> 23)
		sig_y = y & MASK_BINARY_SIG1_32
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
		if (y & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	} else if y_is_zero != 0 {
		// if y is zero, X is less if it is negative
		res = 0
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// OPPOSITE SIGN (CASE5)
	// now, if the sign bits differ, x is less than if y is positive
	if ((x ^ y) & MASK_SIGN32) == MASK_SIGN32 {
		res = 0
		if (y & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// REDUNDANT REPRESENTATIONS (CASE6)
	// if both components are either bigger or smaller
	if sig_x > sig_y && exp_x >= exp_y {
		res = 0
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	if sig_x < sig_y && exp_x <= exp_y {
		res = 0
		if (x & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// if exp_x is 15 greater than exp_y, no need for compensation
	if exp_x-exp_y > 6 {
		res = 0
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		// difference cannot be greater than 10^15
		return res, pfpsf
	}
	// if exp_x is 15 less than exp_y, no need for compensation
	if exp_y-exp_x > 6 {
		res = 0
		if (x & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// if |exp_x - exp_y| < 15, it comes down to the compensated significand
	if exp_x > exp_y { // to simplify the loop below,
		// otherwise adjust the x significand upwards
		sig_n_prime = uint64(sig_x) * bid32_mult_factor[exp_x-exp_y]
		// return 1 if values are equal
		if 0 == 0 && (sig_n_prime == uint64(sig_y)) {
			res = 1
			return res, pfpsf
		}
		// if postitive, return whichever significand abs is smaller
		//     (converse if negative)
		res = 0
		if (sig_n_prime < uint64(sig_y)) != ((x & MASK_SIGN32) == MASK_SIGN32) {
			res = 1
		}
		return res, pfpsf
	}
	// adjust the y significand upwards
	sig_n_prime = uint64(sig_y) * bid32_mult_factor[exp_y-exp_x]
	// return 1 if values are equal
	if 0 == 0 && (sig_n_prime == uint64(sig_x)) {
		res = 1
		return res, pfpsf
	}
	// if positive, return whichever significand abs is smaller
	//     (converse if negative)
	res = 0
	if (uint64(sig_x) < sig_n_prime) != ((x & MASK_SIGN32) == MASK_SIGN32) {
		res = 1
	}
	return res, pfpsf
}

// Bid32QuietLessUnordered - Intel bid64_quiet_less_unordered 기계적 포팅
func Bid32QuietLessUnordered(x, y uint32) (int, uint32) {
	var res int
	var exp_x, exp_y int
	var sig_x, sig_y uint32
	var sig_n_prime uint64
	var pfpsf uint32
	var x_is_zero, y_is_zero, non_canon_x, non_canon_y int

	// NaN (CASE1)
	// if either number is NAN, the comparison is unordered : return 0
	if ((x & MASK_NAN32) == MASK_NAN32) || ((y & MASK_NAN32) == MASK_NAN32) {
		if (x&MASK_SNAN32) == MASK_SNAN32 || (y&MASK_SNAN32) == MASK_SNAN32 {
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
	if (x & MASK_INF32) == MASK_INF32 {
		// if x==neg_inf, { res = (y == neg_inf)?0:1; BID_RETURN (res) }
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			// x is -inf, so it is less than y unless y is -inf
			res = 0
			if ((y & MASK_INF32) != MASK_INF32) || (y&MASK_SIGN32) != MASK_SIGN32 {
				res = 1
			}
			return res, pfpsf
		} else {
			// x is pos_inf, no way for it to be less than y
			res = 0
			return res, pfpsf
		}
	} else if (y & MASK_INF32) == MASK_INF32 {
		// x is finite, so:
		//    if y is +inf, x<y
		//    if y is -inf, x>y
		res = 0
		if (y & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// if steering bits are 11 (condition will be 0), then exponent is G[0:w+1] =>
	if (x & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
		exp_x = int((x & MASK_BINARY_EXPONENT2_32) >> 21)
		sig_x = (x & MASK_BINARY_SIG2_32) | MASK_BINARY_OR2_32
		if sig_x > 9999999 {
			non_canon_x = 1
		} else {
			non_canon_x = 0
		}
	} else {
		exp_x = int((x & MASK_BINARY_EXPONENT1_32) >> 23)
		sig_x = x & MASK_BINARY_SIG1_32
		non_canon_x = 0
	}
	// if steering bits are 11 (condition will be 0), then exponent is G[0:w+1] =>
	if (y & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
		exp_y = int((y & MASK_BINARY_EXPONENT2_32) >> 21)
		sig_y = (y & MASK_BINARY_SIG2_32) | MASK_BINARY_OR2_32
		if sig_y > 9999999 {
			non_canon_y = 1
		} else {
			non_canon_y = 0
		}
	} else {
		exp_y = int((y & MASK_BINARY_EXPONENT1_32) >> 23)
		sig_y = y & MASK_BINARY_SIG1_32
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
		if (y & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	} else if y_is_zero != 0 {
		// if y is zero, X is less if it is negative
		res = 0
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// OPPOSITE SIGN (CASE5)
	// now, if the sign bits differ, x is less than if y is positive
	if ((x ^ y) & MASK_SIGN32) == MASK_SIGN32 {
		res = 0
		if (y & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// REDUNDANT REPRESENTATIONS (CASE6)
	// if both components are either bigger or smaller
	if sig_x > sig_y && exp_x >= exp_y {
		res = 0
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	if sig_x < sig_y && exp_x <= exp_y {
		res = 0
		if (x & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// if exp_x is 15 greater than exp_y, no need for compensation
	if exp_x-exp_y > 6 {
		res = 0
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		// difference cannot be greater than 10^15
		return res, pfpsf
	}
	// if exp_x is 15 less than exp_y, no need for compensation
	if exp_y-exp_x > 6 {
		res = 0
		if (x & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// if |exp_x - exp_y| < 15, it comes down to the compensated significand
	if exp_x > exp_y { // to simplify the loop below,
		// otherwise adjust the x significand upwards
		sig_n_prime = uint64(sig_x) * bid32_mult_factor[exp_x-exp_y]
		// return 0 if values are equal
		if 0 == 0 && (sig_n_prime == uint64(sig_y)) {
			res = 0
			return res, pfpsf
		}
		// if postitive, return whichever significand abs is smaller
		//     (converse if negative)
		res = 0
		if (sig_n_prime < uint64(sig_y)) != ((x & MASK_SIGN32) == MASK_SIGN32) {
			res = 1
		}
		return res, pfpsf
	}
	// adjust the y significand upwards
	sig_n_prime = uint64(sig_y) * bid32_mult_factor[exp_y-exp_x]
	// return 0 if values are equal
	if 0 == 0 && (sig_n_prime == uint64(sig_x)) {
		res = 0
		return res, pfpsf
	}
	// if positive, return whichever significand abs is smaller
	//     (converse if negative)
	res = 0
	if (uint64(sig_x) < sig_n_prime) != ((x & MASK_SIGN32) == MASK_SIGN32) {
		res = 1
	}
	return res, pfpsf
}

// Bid32QuietNotEqual - Intel bid64_quiet_not_equal 기계적 포팅
func Bid32QuietNotEqual(x, y uint32) (int, uint32) {
	var res int
	var exp_x, exp_y, exp_t int
	var sig_x, sig_y, sig_t uint32
	var pfpsf uint32
	var x_is_zero, y_is_zero, non_canon_x, non_canon_y, lcv int

	// NaN (CASE1)
	// if either number is NAN, the comparison is unordered,
	// rather than equal : return 1
	if ((x & MASK_NAN32) == MASK_NAN32) || ((y & MASK_NAN32) == MASK_NAN32) {
		if (x&MASK_SNAN32) == MASK_SNAN32 || (y&MASK_SNAN32) == MASK_SNAN32 {
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
	if ((x & MASK_INF32) == MASK_INF32) && ((y & MASK_INF32) == MASK_INF32) {
		res = 0
		if ((x ^ y) & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// ONE INFINITY (CASE3')
	if ((x & MASK_INF32) == MASK_INF32) || ((y & MASK_INF32) == MASK_INF32) {
		res = 1
		return res, pfpsf
	}
	// if steering bits are 11 (condition will be 0), then exponent is G[0:w+1] =>
	if (x & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
		exp_x = int((x & MASK_BINARY_EXPONENT2_32) >> 21)
		sig_x = (x & MASK_BINARY_SIG2_32) | MASK_BINARY_OR2_32
		if sig_x > 9999999 {
			non_canon_x = 1
		} else {
			non_canon_x = 0
		}
	} else {
		exp_x = int((x & MASK_BINARY_EXPONENT1_32) >> 23)
		sig_x = x & MASK_BINARY_SIG1_32
		non_canon_x = 0
	}

	// if steering bits are 11 (condition will be 0), then exponent is G[0:w+1] =>
	if (y & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
		exp_y = int((y & MASK_BINARY_EXPONENT2_32) >> 21)
		sig_y = (y & MASK_BINARY_SIG2_32) | MASK_BINARY_OR2_32
		if sig_y > 9999999 {
			non_canon_y = 1
		} else {
			non_canon_y = 0
		}
	} else {
		exp_y = int((y & MASK_BINARY_EXPONENT1_32) >> 23)
		sig_y = y & MASK_BINARY_SIG1_32
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
	if ((x ^ y) & MASK_SIGN32) != 0 {
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

	if exp_y-exp_x > 6 {
		res = 1
		return res, pfpsf
	}
	// difference cannot be greater than 10^16

	for lcv = 0; lcv < (exp_y - exp_x); lcv++ {
		// recalculate y's significand upwards
		sig_y = sig_y * 10
		if sig_y > 9999999 {
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

// Bid32QuietNotGreater - Intel bid64_quiet_not_greater 기계적 포팅
func Bid32QuietNotGreater(x, y uint32) (int, uint32) {
	var res int
	var exp_x, exp_y int
	var sig_x, sig_y uint32
	var sig_n_prime uint64
	var pfpsf uint32
	var x_is_zero, y_is_zero, non_canon_x, non_canon_y int

	// NaN (CASE1)
	// if either number is NAN, the comparison is unordered,
	//   rather than equal : return 0
	if ((x & MASK_NAN32) == MASK_NAN32) || ((y & MASK_NAN32) == MASK_NAN32) {
		if (x&MASK_SNAN32) == MASK_SNAN32 || (y&MASK_SNAN32) == MASK_SNAN32 {
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
	if (x & MASK_INF32) == MASK_INF32 {
		// if x is neg infinity, it must be lessthan or equal to y return 1
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
			return res, pfpsf
		}
		// x is pos infinity, it is greater, unless y is positive
		// infinity => return y==pos_infinity
		res = 1
		if ((y & MASK_INF32) != MASK_INF32) || ((y & MASK_SIGN32) == MASK_SIGN32) {
			res = 0
		}
		return res, pfpsf
	} else if (y & MASK_INF32) == MASK_INF32 {
		// x is finite, so if y is positive infinity, then x is less, return 1
		//                 if y is negative infinity, then x is greater, return 0
		res = 0
		if (y & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// if steering bits are 11 (condition will be 0), then exponent is G[0:w+1] =>
	if (x & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
		exp_x = int((x & MASK_BINARY_EXPONENT2_32) >> 21)
		sig_x = (x & MASK_BINARY_SIG2_32) | MASK_BINARY_OR2_32
		if sig_x > 9999999 {
			non_canon_x = 1
		} else {
			non_canon_x = 0
		}
	} else {
		exp_x = int((x & MASK_BINARY_EXPONENT1_32) >> 23)
		sig_x = x & MASK_BINARY_SIG1_32
		non_canon_x = 0
	}

	// if steering bits are 11 (condition will be 0), then exponent is G[0:w+1] =>
	if (y & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
		exp_y = int((y & MASK_BINARY_EXPONENT2_32) >> 21)
		sig_y = (y & MASK_BINARY_SIG2_32) | MASK_BINARY_OR2_32
		if sig_y > 9999999 {
			non_canon_y = 1
		} else {
			non_canon_y = 0
		}
	} else {
		exp_y = int((y & MASK_BINARY_EXPONENT1_32) >> 23)
		sig_y = y & MASK_BINARY_SIG1_32
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
		if (y & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// if y is zero, X is less if it is negative
	if y_is_zero != 0 {
		res = 0
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// OPPOSITE SIGN (CASE5)
	// now, if the sign bits differ, x is less than if y is positive
	if ((x ^ y) & MASK_SIGN32) == MASK_SIGN32 {
		res = 0
		if (y & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// REDUNDANT REPRESENTATIONS (CASE6)
	// if both components are either bigger or smaller
	if sig_x > sig_y && exp_x >= exp_y {
		res = 0
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	if sig_x < sig_y && exp_x <= exp_y {
		res = 0
		if (x & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// if exp_x is 15 greater than exp_y, no need for compensation
	if exp_x-exp_y > 6 {
		res = 0
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// difference cannot be greater than 10^15

	// if exp_x is 15 less than exp_y, no need for compensation
	if exp_y-exp_x > 6 {
		res = 0
		if (x & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// if |exp_x - exp_y| < 15, it comes down to the compensated significand
	if exp_x > exp_y { // to simplify the loop below,
		// otherwise adjust the x significand upwards
		sig_n_prime = uint64(sig_x) * bid32_mult_factor[exp_x-exp_y]

		// return 1 if values are equal
		if 0 == 0 && (sig_n_prime == uint64(sig_y)) {
			res = 1
			return res, pfpsf
		}
		// if postitive, return whichever significand abs is smaller
		//     (converse if negative)
		res = 0
		if (sig_n_prime < uint64(sig_y)) != ((x & MASK_SIGN32) == MASK_SIGN32) {
			res = 1
		}
		return res, pfpsf
	}
	// adjust the y significand upwards
	sig_n_prime = uint64(sig_y) * bid32_mult_factor[exp_y-exp_x]

	// return 1 if values are equal
	if 0 == 0 && (sig_n_prime == uint64(sig_x)) {
		res = 1
		return res, pfpsf
	}
	// if positive, return whichever significand abs is smaller
	//     (converse if negative)
	res = 0
	if (uint64(sig_x) < sig_n_prime) != ((x & MASK_SIGN32) == MASK_SIGN32) {
		res = 1
	}
	return res, pfpsf
}

// Bid32QuietNotLess - Intel bid64_quiet_not_less 기계적 포팅
func Bid32QuietNotLess(x, y uint32) (int, uint32) {
	var res int
	var exp_x, exp_y int
	var sig_x, sig_y uint32
	var sig_n_prime uint64
	var pfpsf uint32
	var x_is_zero, y_is_zero, non_canon_x, non_canon_y int

	// NaN (CASE1)
	// if either number is NAN, the comparison is unordered : return 1
	if ((x & MASK_NAN32) == MASK_NAN32) || ((y & MASK_NAN32) == MASK_NAN32) {
		if (x&MASK_SNAN32) == MASK_SNAN32 || (y&MASK_SNAN32) == MASK_SNAN32 {
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
	if (x & MASK_INF32) == MASK_INF32 {
		// if x==neg_inf, { res = (y == neg_inf)?1:0; BID_RETURN (res) }
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			// x is -inf, so it is less than y unless y is -inf
			res = 0
			if ((y & MASK_INF32) == MASK_INF32) && (y&MASK_SIGN32) == MASK_SIGN32 {
				res = 1
			}
			return res, pfpsf
		}
		// x is pos_inf, no way for it to be less than y
		res = 1
		return res, pfpsf
	} else if (y & MASK_INF32) == MASK_INF32 {
		// x is finite, so:
		//    if y is +inf, x<y
		//    if y is -inf, x>y
		res = 0
		if (y & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// if steering bits are 11 (condition will be 0), then exponent is G[0:w+1] =>
	if (x & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
		exp_x = int((x & MASK_BINARY_EXPONENT2_32) >> 21)
		sig_x = (x & MASK_BINARY_SIG2_32) | MASK_BINARY_OR2_32
		if sig_x > 9999999 {
			non_canon_x = 1
		} else {
			non_canon_x = 0
		}
	} else {
		exp_x = int((x & MASK_BINARY_EXPONENT1_32) >> 23)
		sig_x = x & MASK_BINARY_SIG1_32
		non_canon_x = 0
	}

	// if steering bits are 11 (condition will be 0), then exponent is G[0:w+1] =>
	if (y & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
		exp_y = int((y & MASK_BINARY_EXPONENT2_32) >> 21)
		sig_y = (y & MASK_BINARY_SIG2_32) | MASK_BINARY_OR2_32
		if sig_y > 9999999 {
			non_canon_y = 1
		} else {
			non_canon_y = 0
		}
	} else {
		exp_y = int((y & MASK_BINARY_EXPONENT1_32) >> 23)
		sig_y = y & MASK_BINARY_SIG1_32
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
		if (y & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// if y is zero, X is less if it is negative
	if y_is_zero != 0 {
		res = 0
		if (x & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// OPPOSITE SIGN (CASE5)
	// now, if the sign bits differ, x is less than if y is positive
	if ((x ^ y) & MASK_SIGN32) == MASK_SIGN32 {
		res = 0
		if (y & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// REDUNDANT REPRESENTATIONS (CASE6)
	// if both components are either bigger or smaller
	if sig_x > sig_y && exp_x >= exp_y {
		res = 0
		if (x & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	if sig_x < sig_y && exp_x <= exp_y {
		res = 0
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// if exp_x is 15 greater than exp_y, no need for compensation
	if exp_x-exp_y > 6 {
		res = 0
		if (x & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// difference cannot be greater than 10^15

	// if exp_x is 15 less than exp_y, no need for compensation
	if exp_y-exp_x > 6 {
		res = 0
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// if |exp_x - exp_y| < 15, it comes down to the compensated significand
	if exp_x > exp_y { // to simplify the loop below,

		// otherwise adjust the x significand upwards
		sig_n_prime = uint64(sig_x) * bid32_mult_factor[exp_x-exp_y]

		// return 0 if values are equal
		if 0 == 0 && (sig_n_prime == uint64(sig_y)) {
			res = 1
			return res, pfpsf
		}
		// if postitive, return whichever significand abs is smaller
		//     (converse if negative)
		res = 0
		if (sig_n_prime < uint64(sig_y)) != ((x & MASK_SIGN32) != MASK_SIGN32) {
			res = 1
		}
		return res, pfpsf
	}
	// adjust the y significand upwards
	sig_n_prime = uint64(sig_y) * bid32_mult_factor[exp_y-exp_x]

	// return 0 if values are equal
	if 0 == 0 && (sig_n_prime == uint64(sig_x)) {
		res = 1
		return res, pfpsf
	}
	// if positive, return whichever significand abs is smaller
	//     (converse if negative)
	res = 0
	if (uint64(sig_x) < sig_n_prime) != ((x & MASK_SIGN32) != MASK_SIGN32) {
		res = 1
	}
	return res, pfpsf
}

// Bid32QuietOrdered - Intel bid64_quiet_ordered 기계적 포팅
func Bid32QuietOrdered(x, y uint32) (int, uint32) {
	var res int
	var pfpsf uint32

	// NaN (CASE1)
	// if either number is NAN, the comparison is ordered, rather than equal : return 0
	if ((x & MASK_NAN32) == MASK_NAN32) || ((y & MASK_NAN32) == MASK_NAN32) {
		if (x&MASK_SNAN32) == MASK_SNAN32 || (y&MASK_SNAN32) == MASK_SNAN32 {
			pfpsf |= BID_INVALID_EXCEPTION // set exception if sNaN
		}
		res = 0
		return res, pfpsf
	}
	res = 1
	return res, pfpsf
}

// Bid32QuietUnordered - Intel bid64_quiet_unordered 기계적 포팅
func Bid32QuietUnordered(x, y uint32) (int, uint32) {
	var res int
	var pfpsf uint32

	// NaN (CASE1)
	// if either number is NAN, the comparison is unordered,
	//     rather than equal : return 0
	if ((x & MASK_NAN32) == MASK_NAN32) || ((y & MASK_NAN32) == MASK_NAN32) {
		if (x&MASK_SNAN32) == MASK_SNAN32 || (y&MASK_SNAN32) == MASK_SNAN32 {
			pfpsf |= BID_INVALID_EXCEPTION // set exception if sNaN
		}
		res = 1
		return res, pfpsf
	}
	res = 0
	return res, pfpsf
}

// Bid32SignalingGreater - Intel bid64_signaling_greater 기계적 포팅
func Bid32SignalingGreater(x, y uint32) (int, uint32) {
	var res int
	var exp_x, exp_y int
	var sig_x, sig_y uint32
	var sig_n_prime uint64
	var pfpsf uint32
	var x_is_zero, y_is_zero, non_canon_x, non_canon_y int

	// NaN (CASE1)
	// if either number is NAN, the comparison is unordered,
	//     rather than equal : return 0
	if ((x & MASK_NAN32) == MASK_NAN32) || ((y & MASK_NAN32) == MASK_NAN32) {
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
	if (x & MASK_INF32) == MASK_INF32 {
		// if x is neg infinity, there is no way it is greater than y, return 0
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			res = 0
			return res, pfpsf
		}
		// x is pos infinity, it is greater,
		// unless y is positive infinity => return y!=pos_infinity
		res = 0
		if ((y & MASK_INF32) != MASK_INF32) || ((y & MASK_SIGN32) == MASK_SIGN32) {
			res = 1
		}
		return res, pfpsf
	} else if (y & MASK_INF32) == MASK_INF32 {
		// x is finite, so if y is positive infinity, then x is less, return 0
		//                 if y is negative infinity, then x is greater, return 1
		res = 0
		if (y & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// if steering bits are 11 (condition will be 0), then exponent is G[0:w+1] =>
	if (x & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
		exp_x = int((x & MASK_BINARY_EXPONENT2_32) >> 21)
		sig_x = (x & MASK_BINARY_SIG2_32) | MASK_BINARY_OR2_32
		if sig_x > 9999999 {
			non_canon_x = 1
		} else {
			non_canon_x = 0
		}
	} else {
		exp_x = int((x & MASK_BINARY_EXPONENT1_32) >> 23)
		sig_x = x & MASK_BINARY_SIG1_32
		non_canon_x = 0
	}

	// if steering bits are 11 (condition will be 0), then exponent is G[0:w+1] =>
	if (y & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
		exp_y = int((y & MASK_BINARY_EXPONENT2_32) >> 21)
		sig_y = (y & MASK_BINARY_SIG2_32) | MASK_BINARY_OR2_32
		if sig_y > 9999999 {
			non_canon_y = 1
		} else {
			non_canon_y = 0
		}
	} else {
		exp_y = int((y & MASK_BINARY_EXPONENT1_32) >> 23)
		sig_y = y & MASK_BINARY_SIG1_32
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
		if (y & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// is y is zero, X is greater if it is positive
	if y_is_zero != 0 {
		res = 0
		if (x & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// OPPOSITE SIGN (CASE5)
	// now, if the sign bits differ, x is greater if y is negative
	if ((x ^ y) & MASK_SIGN32) == MASK_SIGN32 {
		res = 0
		if (y & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// REDUNDANT REPRESENTATIONS (CASE6)

	// if both components are either bigger or smaller
	if sig_x > sig_y && exp_x >= exp_y {
		res = 0
		if (x & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	if sig_x < sig_y && exp_x <= exp_y {
		res = 0
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// if exp_x is 15 greater than exp_y, no need for compensation
	if exp_x-exp_y > 6 {
		res = 0
		if (x & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// difference cannot be greater than 10^15

	// if exp_x is 15 less than exp_y, no need for compensation
	if exp_y-exp_x > 6 {
		res = 0
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// if |exp_x - exp_y| < 15, it comes down to the compensated significand
	if exp_x > exp_y { // to simplify the loop below,

		// otherwise adjust the x significand upwards
		sig_n_prime = uint64(sig_x) * bid32_mult_factor[exp_x-exp_y]

		// if postitive, return whichever significand is larger
		//     (converse if negative)
		if 0 == 0 && (sig_n_prime == uint64(sig_y)) {
			res = 0
			return res, pfpsf
		}

		res = 0
		if ((0 > 0) || sig_n_prime > uint64(sig_y)) != ((x & MASK_SIGN32) == MASK_SIGN32) {
			res = 1
		}
		return res, pfpsf
	}
	// adjust the y significand upwards
	sig_n_prime = uint64(sig_y) * bid32_mult_factor[exp_y-exp_x]

	// if postitive, return whichever significand is larger
	//     (converse if negative)
	if 0 == 0 && (sig_n_prime == uint64(sig_x)) {
		res = 0
		return res, pfpsf
	}
	res = 0
	if (uint64(sig_x) > sig_n_prime) != ((x & MASK_SIGN32) == MASK_SIGN32) {
		res = 1
	}
	return res, pfpsf
}

// Bid32SignalingGreaterEqual - Intel bid64_signaling_greater_equal 기계적 포팅
func Bid32SignalingGreaterEqual(x, y uint32) (int, uint32) {
	var res int
	var exp_x, exp_y int
	var sig_x, sig_y uint32
	var sig_n_prime uint64
	var pfpsf uint32
	var x_is_zero, y_is_zero, non_canon_x, non_canon_y int

	// NaN (CASE1)
	// if either number is NAN, the comparison is unordered : return 1
	if ((x & MASK_NAN32) == MASK_NAN32) || ((y & MASK_NAN32) == MASK_NAN32) {
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
	if (x & MASK_INF32) == MASK_INF32 {
		// if x==neg_inf, { res = (y == neg_inf)?1:0; BID_RETURN (res) }
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			// x is -inf, so it is less than y unless y is -inf
			res = 0
			if ((y & MASK_INF32) == MASK_INF32) && (y&MASK_SIGN32) == MASK_SIGN32 {
				res = 1
			}
			return res, pfpsf
		}
		// x is pos_inf, no way for it to be less than y
		res = 1
		return res, pfpsf
	} else if (y & MASK_INF32) == MASK_INF32 {
		// x is finite, so:
		//    if y is +inf, x<y
		//    if y is -inf, x>y
		res = 0
		if (y & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// if steering bits are 11 (condition will be 0), then exponent is G[0:w+1] =>
	if (x & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
		exp_x = int((x & MASK_BINARY_EXPONENT2_32) >> 21)
		sig_x = (x & MASK_BINARY_SIG2_32) | MASK_BINARY_OR2_32
		if sig_x > 9999999 {
			non_canon_x = 1
		} else {
			non_canon_x = 0
		}
	} else {
		exp_x = int((x & MASK_BINARY_EXPONENT1_32) >> 23)
		sig_x = x & MASK_BINARY_SIG1_32
		non_canon_x = 0
	}

	// if steering bits are 11 (condition will be 0), then exponent is G[0:w+1] =>
	if (y & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
		exp_y = int((y & MASK_BINARY_EXPONENT2_32) >> 21)
		sig_y = (y & MASK_BINARY_SIG2_32) | MASK_BINARY_OR2_32
		if sig_y > 9999999 {
			non_canon_y = 1
		} else {
			non_canon_y = 0
		}
	} else {
		exp_y = int((y & MASK_BINARY_EXPONENT1_32) >> 23)
		sig_y = y & MASK_BINARY_SIG1_32
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
		if (y & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// if y is zero, X is less if it is negative
	if y_is_zero != 0 {
		res = 0
		if (x & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// OPPOSITE SIGN (CASE5)
	// now, if the sign bits differ, x is less than if y is positive
	if ((x ^ y) & MASK_SIGN32) == MASK_SIGN32 {
		res = 0
		if (y & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// REDUNDANT REPRESENTATIONS (CASE6)
	// if both components are either bigger or smaller
	if sig_x > sig_y && exp_x >= exp_y {
		res = 0
		if (x & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	if sig_x < sig_y && exp_x <= exp_y {
		res = 0
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// if exp_x is 15 greater than exp_y, no need for compensation
	if exp_x-exp_y > 6 {
		res = 0
		if (x & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// difference cannot be greater than 10^15

	// if exp_x is 15 less than exp_y, no need for compensation
	if exp_y-exp_x > 6 {
		res = 0
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// if |exp_x - exp_y| < 15, it comes down to the compensated significand
	if exp_x > exp_y { // to simplify the loop below,

		// otherwise adjust the x significand upwards
		sig_n_prime = uint64(sig_x) * bid32_mult_factor[exp_x-exp_y]

		// return 1 if values are equal
		if 0 == 0 && (sig_n_prime == uint64(sig_y)) {
			res = 1
			return res, pfpsf
		}
		// if postitive, return whichever significand abs is smaller
		//     (converse if negative)
		res = 0
		if (sig_n_prime < uint64(sig_y)) != ((x & MASK_SIGN32) != MASK_SIGN32) {
			res = 1
		}
		return res, pfpsf
	}
	// adjust the y significand upwards
	sig_n_prime = uint64(sig_y) * bid32_mult_factor[exp_y-exp_x]

	// return 0 if values are equal
	if 0 == 0 && (sig_n_prime == uint64(sig_x)) {
		res = 1
		return res, pfpsf
	}
	// if positive, return whichever significand abs is smaller
	//     (converse if negative)
	res = 0
	if (uint64(sig_x) < sig_n_prime) != ((x & MASK_SIGN32) != MASK_SIGN32) {
		res = 1
	}
	return res, pfpsf
}

// Bid32SignalingGreaterUnordered - Intel bid64_signaling_greater_unordered 기계적 포팅
func Bid32SignalingGreaterUnordered(x, y uint32) (int, uint32) {
	var res int
	var exp_x, exp_y int
	var sig_x, sig_y uint32
	var sig_n_prime uint64
	var pfpsf uint32
	var x_is_zero, y_is_zero, non_canon_x, non_canon_y int

	// NaN (CASE1)
	// if either number is NAN, the comparison is unordered,
	// rather than equal : return 0
	if ((x & MASK_NAN32) == MASK_NAN32) || ((y & MASK_NAN32) == MASK_NAN32) {
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
	if (x & MASK_INF32) == MASK_INF32 {
		// if x is neg infinity, there is no way it is greater than y, return 0
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			res = 0
			return res, pfpsf
		}
		// x is pos infinity, it is greater,
		// unless y is positive infinity => return y!=pos_infinity
		res = 0
		if ((y & MASK_INF32) != MASK_INF32) || ((y & MASK_SIGN32) == MASK_SIGN32) {
			res = 1
		}
		return res, pfpsf
	} else if (y & MASK_INF32) == MASK_INF32 {
		// x is finite, so if y is positive infinity, then x is less, return 0
		//                 if y is negative infinity, then x is greater, return 1
		res = 0
		if (y & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// if steering bits are 11 (condition will be 0), then exponent is G[0:w+1] =>
	if (x & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
		exp_x = int((x & MASK_BINARY_EXPONENT2_32) >> 21)
		sig_x = (x & MASK_BINARY_SIG2_32) | MASK_BINARY_OR2_32
		if sig_x > 9999999 {
			non_canon_x = 1
		} else {
			non_canon_x = 0
		}
	} else {
		exp_x = int((x & MASK_BINARY_EXPONENT1_32) >> 23)
		sig_x = x & MASK_BINARY_SIG1_32
		non_canon_x = 0
	}

	// if steering bits are 11 (condition will be 0), then exponent is G[0:w+1] =>
	if (y & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
		exp_y = int((y & MASK_BINARY_EXPONENT2_32) >> 21)
		sig_y = (y & MASK_BINARY_SIG2_32) | MASK_BINARY_OR2_32
		if sig_y > 9999999 {
			non_canon_y = 1
		} else {
			non_canon_y = 0
		}
	} else {
		exp_y = int((y & MASK_BINARY_EXPONENT1_32) >> 23)
		sig_y = y & MASK_BINARY_SIG1_32
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
		if (y & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// is y is zero, X is greater if it is positive
	if y_is_zero != 0 {
		res = 0
		if (x & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// OPPOSITE SIGN (CASE5)
	// now, if the sign bits differ, x is greater if y is negative
	if ((x ^ y) & MASK_SIGN32) == MASK_SIGN32 {
		res = 0
		if (y & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// REDUNDANT REPRESENTATIONS (CASE6)

	// if both components are either bigger or smaller
	if sig_x > sig_y && exp_x >= exp_y {
		res = 0
		if (x & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	if sig_x < sig_y && exp_x <= exp_y {
		res = 0
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// if exp_x is 15 greater than exp_y, no need for compensation
	if exp_x-exp_y > 6 {
		res = 0
		if (x & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// difference cannot be greater than 10^15

	// if exp_x is 15 less than exp_y, no need for compensation
	if exp_y-exp_x > 6 {
		res = 0
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// if |exp_x - exp_y| < 15, it comes down to the compensated significand
	if exp_x > exp_y { // to simplify the loop below,

		// otherwise adjust the x significand upwards
		sig_n_prime = uint64(sig_x) * bid32_mult_factor[exp_x-exp_y]

		// if postitive, return whichever significand is larger
		//     (converse if negative)
		if 0 == 0 && (sig_n_prime == uint64(sig_y)) {
			res = 0
			return res, pfpsf
		}

		res = 0
		if ((0 > 0) || sig_n_prime > uint64(sig_y)) != ((x & MASK_SIGN32) == MASK_SIGN32) {
			res = 1
		}
		return res, pfpsf
	}
	// adjust the y significand upwards
	sig_n_prime = uint64(sig_y) * bid32_mult_factor[exp_y-exp_x]

	// if postitive, return whichever significand is larger
	//     (converse if negative)
	if 0 == 0 && (sig_n_prime == uint64(sig_x)) {
		res = 0
		return res, pfpsf
	}
	res = 0
	if (uint64(sig_x) > sig_n_prime) != ((x & MASK_SIGN32) == MASK_SIGN32) {
		res = 1
	}
	return res, pfpsf
}

// Bid32SignalingLessEqual - Intel bid64_signaling_less_equal 기계적 포팅
func Bid32SignalingLessEqual(x, y uint32) (int, uint32) {
	var res int
	var exp_x, exp_y int
	var sig_x, sig_y uint32
	var sig_n_prime uint64
	var pfpsf uint32
	var x_is_zero, y_is_zero, non_canon_x, non_canon_y int

	// NaN (CASE1)
	// if either number is NAN, the comparison is unordered,
	// rather than equal : return 0
	if ((x & MASK_NAN32) == MASK_NAN32) || ((y & MASK_NAN32) == MASK_NAN32) {
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
	if (x & MASK_INF32) == MASK_INF32 {
		// if x is neg infinity, it must be lessthan or equal to y return 1
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
			return res, pfpsf
		}
		// x is pos infinity, it is greater,
		// unless y is positive infinity => return y==pos_infinity
		res = 1
		if ((y & MASK_INF32) != MASK_INF32) || ((y & MASK_SIGN32) == MASK_SIGN32) {
			res = 0
		}
		return res, pfpsf
	} else if (y & MASK_INF32) == MASK_INF32 {
		// x is finite, so if y is positive infinity, then x is less, return 1
		//                 if y is negative infinity, then x is greater, return 0
		res = 0
		if (y & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// if steering bits are 11 (condition will be 0), then exponent is G[0:w+1] =>
	if (x & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
		exp_x = int((x & MASK_BINARY_EXPONENT2_32) >> 21)
		sig_x = (x & MASK_BINARY_SIG2_32) | MASK_BINARY_OR2_32
		if sig_x > 9999999 {
			non_canon_x = 1
		} else {
			non_canon_x = 0
		}
	} else {
		exp_x = int((x & MASK_BINARY_EXPONENT1_32) >> 23)
		sig_x = x & MASK_BINARY_SIG1_32
		non_canon_x = 0
	}

	// if steering bits are 11 (condition will be 0), then exponent is G[0:w+1] =>
	if (y & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
		exp_y = int((y & MASK_BINARY_EXPONENT2_32) >> 21)
		sig_y = (y & MASK_BINARY_SIG2_32) | MASK_BINARY_OR2_32
		if sig_y > 9999999 {
			non_canon_y = 1
		} else {
			non_canon_y = 0
		}
	} else {
		exp_y = int((y & MASK_BINARY_EXPONENT1_32) >> 23)
		sig_y = y & MASK_BINARY_SIG1_32
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
		if (y & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// if y is zero, X is less if it is negative
	if y_is_zero != 0 {
		res = 0
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// OPPOSITE SIGN (CASE5)
	// now, if the sign bits differ, x is less than if y is positive
	if ((x ^ y) & MASK_SIGN32) == MASK_SIGN32 {
		res = 0
		if (y & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// REDUNDANT REPRESENTATIONS (CASE6)
	// if both components are either bigger or smaller
	if sig_x > sig_y && exp_x >= exp_y {
		res = 0
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	if sig_x < sig_y && exp_x <= exp_y {
		res = 0
		if (x & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// if exp_x is 15 greater than exp_y, no need for compensation
	if exp_x-exp_y > 6 {
		res = 0
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// difference cannot be greater than 10^15

	// if exp_x is 15 less than exp_y, no need for compensation
	if exp_y-exp_x > 6 {
		res = 0
		if (x & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// if |exp_x - exp_y| < 15, it comes down to the compensated significand
	if exp_x > exp_y { // to simplify the loop below,

		// otherwise adjust the x significand upwards
		sig_n_prime = uint64(sig_x) * bid32_mult_factor[exp_x-exp_y]

		// return 1 if values are equal
		if 0 == 0 && (sig_n_prime == uint64(sig_y)) {
			res = 1
			return res, pfpsf
		}
		// if postitive, return whichever significand abs is smaller
		//     (converse if negative)
		res = 0
		if (sig_n_prime < uint64(sig_y)) != ((x & MASK_SIGN32) == MASK_SIGN32) {
			res = 1
		}
		return res, pfpsf
	}
	// adjust the y significand upwards
	sig_n_prime = uint64(sig_y) * bid32_mult_factor[exp_y-exp_x]

	// return 1 if values are equal
	if 0 == 0 && (sig_n_prime == uint64(sig_x)) {
		res = 1
		return res, pfpsf
	}
	// if positive, return whichever significand abs is smaller
	//     (converse if negative)
	res = 0
	if (uint64(sig_x) < sig_n_prime) != ((x & MASK_SIGN32) == MASK_SIGN32) {
		res = 1
	}
	return res, pfpsf
}

// Bid32SignalingLessUnordered - Intel bid64_signaling_less_unordered 기계적 포팅
func Bid32SignalingLessUnordered(x, y uint32) (int, uint32) {
	var res int
	var exp_x, exp_y int
	var sig_x, sig_y uint32
	var sig_n_prime uint64
	var pfpsf uint32
	var x_is_zero, y_is_zero, non_canon_x, non_canon_y int

	// NaN (CASE1)
	// if either number is NAN, the comparison is unordered : return 0
	if ((x & MASK_NAN32) == MASK_NAN32) || ((y & MASK_NAN32) == MASK_NAN32) {
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
	if (x & MASK_INF32) == MASK_INF32 {
		// if x==neg_inf, { res = (y == neg_inf)?0:1; BID_RETURN (res) }
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			// x is -inf, so it is less than y unless y is -inf
			res = 0
			if ((y & MASK_INF32) != MASK_INF32) || (y&MASK_SIGN32) != MASK_SIGN32 {
				res = 1
			}
			return res, pfpsf
		}
		// x is pos_inf, no way for it to be less than y
		res = 0
		return res, pfpsf
	} else if (y & MASK_INF32) == MASK_INF32 {
		// x is finite, so:
		//    if y is +inf, x<y
		//    if y is -inf, x>y
		res = 0
		if (y & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// if steering bits are 11 (condition will be 0), then exponent is G[0:w+1] =>
	if (x & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
		exp_x = int((x & MASK_BINARY_EXPONENT2_32) >> 21)
		sig_x = (x & MASK_BINARY_SIG2_32) | MASK_BINARY_OR2_32
		if sig_x > 9999999 {
			non_canon_x = 1
		} else {
			non_canon_x = 0
		}
	} else {
		exp_x = int((x & MASK_BINARY_EXPONENT1_32) >> 23)
		sig_x = x & MASK_BINARY_SIG1_32
		non_canon_x = 0
	}

	// if steering bits are 11 (condition will be 0), then exponent is G[0:w+1] =>
	if (y & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
		exp_y = int((y & MASK_BINARY_EXPONENT2_32) >> 21)
		sig_y = (y & MASK_BINARY_SIG2_32) | MASK_BINARY_OR2_32
		if sig_y > 9999999 {
			non_canon_y = 1
		} else {
			non_canon_y = 0
		}
	} else {
		exp_y = int((y & MASK_BINARY_EXPONENT1_32) >> 23)
		sig_y = y & MASK_BINARY_SIG1_32
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
		if (y & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// if y is zero, X is less if it is negative
	if y_is_zero != 0 {
		res = 0
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// OPPOSITE SIGN (CASE5)
	// now, if the sign bits differ, x is less than if y is positive
	if ((x ^ y) & MASK_SIGN32) == MASK_SIGN32 {
		res = 0
		if (y & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// REDUNDANT REPRESENTATIONS (CASE6)
	// if both components are either bigger or smaller
	if sig_x > sig_y && exp_x >= exp_y {
		res = 0
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	if sig_x < sig_y && exp_x <= exp_y {
		res = 0
		if (x & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// if exp_x is 15 greater than exp_y, no need for compensation
	if exp_x-exp_y > 6 {
		res = 0
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// difference cannot be greater than 10^15

	// if exp_x is 15 less than exp_y, no need for compensation
	if exp_y-exp_x > 6 {
		res = 0
		if (x & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// if |exp_x - exp_y| < 15, it comes down to the compensated significand
	if exp_x > exp_y { // to simplify the loop below,

		// otherwise adjust the x significand upwards
		sig_n_prime = uint64(sig_x) * bid32_mult_factor[exp_x-exp_y]

		// return 0 if values are equal
		if 0 == 0 && (sig_n_prime == uint64(sig_y)) {
			res = 0
			return res, pfpsf
		}
		// if postitive, return whichever significand abs is smaller
		//     (converse if negative)
		res = 0
		if (sig_n_prime < uint64(sig_y)) != ((x & MASK_SIGN32) == MASK_SIGN32) {
			res = 1
		}
		return res, pfpsf
	}
	// adjust the y significand upwards
	sig_n_prime = uint64(sig_y) * bid32_mult_factor[exp_y-exp_x]

	// return 0 if values are equal
	if 0 == 0 && (sig_n_prime == uint64(sig_x)) {
		res = 0
		return res, pfpsf
	}
	// if positive, return whichever significand abs is smaller
	//     (converse if negative)
	res = 0
	if (uint64(sig_x) < sig_n_prime) != ((x & MASK_SIGN32) == MASK_SIGN32) {
		res = 1
	}
	return res, pfpsf
}

// Bid32SignalingNotGreater - Intel bid64_signaling_not_greater 기계적 포팅
func Bid32SignalingNotGreater(x, y uint32) (int, uint32) {
	var res int
	var exp_x, exp_y int
	var sig_x, sig_y uint32
	var sig_n_prime uint64
	var pfpsf uint32
	var x_is_zero, y_is_zero, non_canon_x, non_canon_y int

	// NaN (CASE1)
	// if either number is NAN, the comparison is unordered,
	// rather than equal : return 0
	if ((x & MASK_NAN32) == MASK_NAN32) || ((y & MASK_NAN32) == MASK_NAN32) {
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
	if (x & MASK_INF32) == MASK_INF32 {
		// if x is neg infinity, it must be lessthan or equal to y return 1
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
			return res, pfpsf
		}
		// x is pos infinity, it is greater,
		// unless y is positive infinity => return y==pos_infinity
		res = 1
		if ((y & MASK_INF32) != MASK_INF32) || ((y & MASK_SIGN32) == MASK_SIGN32) {
			res = 0
		}
		return res, pfpsf
	} else if (y & MASK_INF32) == MASK_INF32 {
		// x is finite, so if y is positive infinity, then x is less, return 1
		//                 if y is negative infinity, then x is greater, return 0
		res = 0
		if (y & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// if steering bits are 11 (condition will be 0), then exponent is G[0:w+1] =>
	if (x & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
		exp_x = int((x & MASK_BINARY_EXPONENT2_32) >> 21)
		sig_x = (x & MASK_BINARY_SIG2_32) | MASK_BINARY_OR2_32
		if sig_x > 9999999 {
			non_canon_x = 1
		} else {
			non_canon_x = 0
		}
	} else {
		exp_x = int((x & MASK_BINARY_EXPONENT1_32) >> 23)
		sig_x = x & MASK_BINARY_SIG1_32
		non_canon_x = 0
	}

	// if steering bits are 11 (condition will be 0), then exponent is G[0:w+1] =>
	if (y & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
		exp_y = int((y & MASK_BINARY_EXPONENT2_32) >> 21)
		sig_y = (y & MASK_BINARY_SIG2_32) | MASK_BINARY_OR2_32
		if sig_y > 9999999 {
			non_canon_y = 1
		} else {
			non_canon_y = 0
		}
	} else {
		exp_y = int((y & MASK_BINARY_EXPONENT1_32) >> 23)
		sig_y = y & MASK_BINARY_SIG1_32
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
		if (y & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// if y is zero, X is less if it is negative
	if y_is_zero != 0 {
		res = 0
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// OPPOSITE SIGN (CASE5)
	// now, if the sign bits differ, x is less than if y is positive
	if ((x ^ y) & MASK_SIGN32) == MASK_SIGN32 {
		res = 0
		if (y & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// REDUNDANT REPRESENTATIONS (CASE6)
	// if both components are either bigger or smaller
	if sig_x > sig_y && exp_x >= exp_y {
		res = 0
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	if sig_x < sig_y && exp_x <= exp_y {
		res = 0
		if (x & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// if exp_x is 15 greater than exp_y, no need for compensation
	if exp_x-exp_y > 6 {
		res = 0
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// difference cannot be greater than 10^15

	// if exp_x is 15 less than exp_y, no need for compensation
	if exp_y-exp_x > 6 {
		res = 0
		if (x & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// if |exp_x - exp_y| < 15, it comes down to the compensated significand
	if exp_x > exp_y { // to simplify the loop below,

		// otherwise adjust the x significand upwards
		sig_n_prime = uint64(sig_x) * bid32_mult_factor[exp_x-exp_y]

		// return 1 if values are equal
		if 0 == 0 && (sig_n_prime == uint64(sig_y)) {
			res = 1
			return res, pfpsf
		}
		// if postitive, return whichever significand abs is smaller
		//     (converse if negative)
		res = 0
		if (sig_n_prime < uint64(sig_y)) != ((x & MASK_SIGN32) == MASK_SIGN32) {
			res = 1
		}
		return res, pfpsf
	}
	// adjust the y significand upwards
	sig_n_prime = uint64(sig_y) * bid32_mult_factor[exp_y-exp_x]

	// return 1 if values are equal
	if 0 == 0 && (sig_n_prime == uint64(sig_x)) {
		res = 1
		return res, pfpsf
	}
	// if positive, return whichever significand abs is smaller
	//     (converse if negative)
	res = 0
	if (uint64(sig_x) < sig_n_prime) != ((x & MASK_SIGN32) == MASK_SIGN32) {
		res = 1
	}
	return res, pfpsf
}

// Bid32SignalingNotLess - Intel bid64_signaling_not_less 기계적 포팅
func Bid32SignalingNotLess(x, y uint32) (int, uint32) {
	var res int
	var exp_x, exp_y int
	var sig_x, sig_y uint32
	var sig_n_prime uint64
	var pfpsf uint32
	var x_is_zero, y_is_zero, non_canon_x, non_canon_y int

	// NaN (CASE1)
	// if either number is NAN, the comparison is unordered : return 1
	if ((x & MASK_NAN32) == MASK_NAN32) || ((y & MASK_NAN32) == MASK_NAN32) {
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
	if (x & MASK_INF32) == MASK_INF32 {
		// if x==neg_inf, { res = (y == neg_inf)?1:0; BID_RETURN (res) }
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			// x is -inf, so it is less than y unless y is -inf
			res = 0
			if ((y & MASK_INF32) == MASK_INF32) && (y&MASK_SIGN32) == MASK_SIGN32 {
				res = 1
			}
			return res, pfpsf
		}
		// x is pos_inf, no way for it to be less than y
		res = 1
		return res, pfpsf
	} else if (y & MASK_INF32) == MASK_INF32 {
		// x is finite, so:
		//    if y is +inf, x<y
		//    if y is -inf, x>y
		res = 0
		if (y & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// if steering bits are 11 (condition will be 0), then exponent is G[0:w+1] =>
	if (x & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
		exp_x = int((x & MASK_BINARY_EXPONENT2_32) >> 21)
		sig_x = (x & MASK_BINARY_SIG2_32) | MASK_BINARY_OR2_32
		if sig_x > 9999999 {
			non_canon_x = 1
		} else {
			non_canon_x = 0
		}
	} else {
		exp_x = int((x & MASK_BINARY_EXPONENT1_32) >> 23)
		sig_x = x & MASK_BINARY_SIG1_32
		non_canon_x = 0
	}

	// if steering bits are 11 (condition will be 0), then exponent is G[0:w+1] =>
	if (y & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
		exp_y = int((y & MASK_BINARY_EXPONENT2_32) >> 21)
		sig_y = (y & MASK_BINARY_SIG2_32) | MASK_BINARY_OR2_32
		if sig_y > 9999999 {
			non_canon_y = 1
		} else {
			non_canon_y = 0
		}
	} else {
		exp_y = int((y & MASK_BINARY_EXPONENT1_32) >> 23)
		sig_y = y & MASK_BINARY_SIG1_32
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
		if (y & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// if y is zero, X is less if it is negative
	if y_is_zero != 0 {
		res = 0
		if (x & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// OPPOSITE SIGN (CASE5)
	// now, if the sign bits differ, x is less than if y is positive
	if ((x ^ y) & MASK_SIGN32) == MASK_SIGN32 {
		res = 0
		if (y & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// REDUNDANT REPRESENTATIONS (CASE6)
	// if both components are either bigger or smaller
	if sig_x > sig_y && exp_x >= exp_y {
		res = 0
		if (x & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	if sig_x < sig_y && exp_x <= exp_y {
		res = 0
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// if exp_x is 15 greater than exp_y, no need for compensation
	if exp_x-exp_y > 6 {
		res = 0
		if (x & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// difference cannot be greater than 10^15

	// if exp_x is 15 less than exp_y, no need for compensation
	if exp_y-exp_x > 6 {
		res = 0
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	// if |exp_x - exp_y| < 15, it comes down to the compensated significand
	if exp_x > exp_y { // to simplify the loop below,

		// otherwise adjust the x significand upwards
		sig_n_prime = uint64(sig_x) * bid32_mult_factor[exp_x-exp_y]

		// return 0 if values are equal
		if 0 == 0 && (sig_n_prime == uint64(sig_y)) {
			res = 1
			return res, pfpsf
		}
		// if postitive, return whichever significand abs is smaller
		//     (converse if negative)
		res = 0
		if (sig_n_prime < uint64(sig_y)) != ((x & MASK_SIGN32) != MASK_SIGN32) {
			res = 1
		}
		return res, pfpsf
	}
	// adjust the y significand upwards
	sig_n_prime = uint64(sig_y) * bid32_mult_factor[exp_y-exp_x]

	// return 0 if values are equal
	if 0 == 0 && (sig_n_prime == uint64(sig_x)) {
		res = 1
		return res, pfpsf
	}
	// if positive, return whichever significand abs is smaller
	//     (converse if negative)
	res = 0
	if (uint64(sig_x) < sig_n_prime) != ((x & MASK_SIGN32) != MASK_SIGN32) {
		res = 1
	}
	return res, pfpsf
}

// Bid32SignalingLess - Intel bid32_signaling_less 기계적 포팅
func Bid32SignalingLess(x, y uint32) (int, uint32) {
	var res int
	var exp_x, exp_y int
	var sig_x, sig_y uint32
	var sig_n_prime uint64
	var x_is_zero, y_is_zero, non_canon_x, non_canon_y int
	var pfpsf uint32

	if ((x & MASK_NAN32) == MASK_NAN32) || ((y & MASK_NAN32) == MASK_NAN32) {
		pfpsf |= BID_INVALID_EXCEPTION
		res = 0
		return res, pfpsf
	}
	if x == y {
		res = 0
		return res, pfpsf
	}
	if (x & MASK_INF32) == MASK_INF32 {
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			res = 0
			if ((y & MASK_INF32) != MASK_INF32) || (y&MASK_SIGN32) != MASK_SIGN32 {
				res = 1
			}
			return res, pfpsf
		}
		res = 0
		return res, pfpsf
	} else if (y & MASK_INF32) == MASK_INF32 {
		res = 0
		if (y & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	if (x & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
		exp_x = int((x & MASK_BINARY_EXPONENT2_32) >> 21)
		sig_x = (x & MASK_BINARY_SIG2_32) | MASK_BINARY_OR2_32
		if sig_x > 9999999 {
			non_canon_x = 1
		}
	} else {
		exp_x = int((x & MASK_BINARY_EXPONENT1_32) >> 23)
		sig_x = x & MASK_BINARY_SIG1_32
	}
	if (y & MASK_STEERING_BITS32) == MASK_STEERING_BITS32 {
		exp_y = int((y & MASK_BINARY_EXPONENT2_32) >> 21)
		sig_y = (y & MASK_BINARY_SIG2_32) | MASK_BINARY_OR2_32
		if sig_y > 9999999 {
			non_canon_y = 1
		}
	} else {
		exp_y = int((y & MASK_BINARY_EXPONENT1_32) >> 23)
		sig_y = y & MASK_BINARY_SIG1_32
	}
	if non_canon_x != 0 || sig_x == 0 {
		x_is_zero = 1
	}
	if non_canon_y != 0 || sig_y == 0 {
		y_is_zero = 1
	}
	if x_is_zero != 0 && y_is_zero != 0 {
		res = 0
		return res, pfpsf
	}
	if x_is_zero != 0 {
		res = 0
		if (y & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	if y_is_zero != 0 {
		res = 0
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	if ((x ^ y) & MASK_SIGN32) == MASK_SIGN32 {
		res = 0
		if (y & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	if sig_x > sig_y && exp_x >= exp_y {
		res = 0
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	if sig_x < sig_y && exp_x <= exp_y {
		res = 0
		if (x & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	if exp_x-exp_y > 6 {
		res = 0
		if (x & MASK_SIGN32) == MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	if exp_y-exp_x > 6 {
		res = 0
		if (x & MASK_SIGN32) != MASK_SIGN32 {
			res = 1
		}
		return res, pfpsf
	}
	if exp_x > exp_y {
		sig_n_prime = uint64(sig_x) * bid32_mult_factor[exp_x-exp_y]
		if sig_n_prime == uint64(sig_y) {
			res = 0
			return res, pfpsf
		}
		res = 0
		if (sig_n_prime < uint64(sig_y)) != ((x & MASK_SIGN32) == MASK_SIGN32) {
			res = 1
		}
		return res, pfpsf
	}
	sig_n_prime = uint64(sig_y) * bid32_mult_factor[exp_y-exp_x]
	if sig_n_prime == uint64(sig_x) {
		res = 0
		return res, pfpsf
	}
	res = 0
	if (uint64(sig_x) < sig_n_prime) != ((x & MASK_SIGN32) == MASK_SIGN32) {
		res = 1
	}
	return res, pfpsf
}
