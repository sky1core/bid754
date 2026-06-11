// Ported from: Intel bid128_minmax.c
// Mechanical translation - all logic preserved exactly.

package bidgo

// Bid128Minnum returns the minimum of two numbers (NaN-favoring-number semantics).
// Ported from bid128_minnum.
func Bid128Minnum(x, y BID_UINT128, pfpsf *uint32) BID_UINT128 {
	var res BID_UINT128
	var exp_x, exp_y int
	var diff int
	var sig_x, sig_y BID_UINT128
	var sig_n_prime192 BID_UINT192
	var sig_n_prime256 BID_UINT256
	var x_is_zero, y_is_zero byte

	// check for non-canonical x
	if (x.w[1] & NAN_MASK64) == NAN_MASK64 { // x is NAN
		x.w[1] = x.w[1] & 0xfe003fffffffffff // clear out G[6]-G[16]
		// check for non-canonical NaN payload
		if ((x.w[1] & 0x00003fffffffffff) > 0x0000314dc6448d93) ||
			(((x.w[1] & 0x00003fffffffffff) == 0x0000314dc6448d93) &&
				(x.w[0] > 0x38c15b09ffffffff)) {
			x.w[1] = x.w[1] & 0xffffc00000000000
			x.w[0] = 0x0
		}
	} else if (x.w[1] & MASK_ANY_INF_128) == INFINITY_MASK64 { // x = inf
		x.w[1] = x.w[1] & (MASK_SIGN64 | INFINITY_MASK64)
		x.w[0] = 0x0
	} else { // x is not special
		// check for non-canonical values - treated as zero
		if (x.w[1] & MASK_STEERING_BITS64) == MASK_STEERING_BITS64 { // G0_G1=11
			// non-canonical
			x.w[1] = (x.w[1] & MASK_SIGN64) | ((x.w[1] << 2) & MASK_EXP128)
			x.w[0] = 0x0
		} else { // G0_G1 != 11
			if (x.w[1]&MASK_COEFF128) > 0x0001ed09bead87c0 ||
				((x.w[1]&MASK_COEFF128) == 0x0001ed09bead87c0 &&
					x.w[0] > 0x378d8e63ffffffff) {
				// x is non-canonical if coefficient is larger than 10^34 -1
				x.w[1] = (x.w[1] & MASK_SIGN64) | (x.w[1] & MASK_EXP128)
				x.w[0] = 0x0
			} else { // canonical
			}
		}
	}
	// check for non-canonical y
	if (y.w[1] & NAN_MASK64) == NAN_MASK64 { // y is NAN
		y.w[1] = y.w[1] & 0xfe003fffffffffff // clear out G[6]-G[16]
		// check for non-canonical NaN payload
		if ((y.w[1] & 0x00003fffffffffff) > 0x0000314dc6448d93) ||
			(((y.w[1] & 0x00003fffffffffff) == 0x0000314dc6448d93) &&
				(y.w[0] > 0x38c15b09ffffffff)) {
			y.w[1] = y.w[1] & 0xffffc00000000000
			y.w[0] = 0x0
		}
	} else if (y.w[1] & MASK_ANY_INF_128) == INFINITY_MASK64 { // y = inf
		y.w[1] = y.w[1] & (MASK_SIGN64 | INFINITY_MASK64)
		y.w[0] = 0x0
	} else { // y is not special
		// check for non-canonical values - treated as zero
		if (y.w[1] & MASK_STEERING_BITS64) == MASK_STEERING_BITS64 { // G0_G1=11
			// non-canonical
			y.w[1] = (y.w[1] & MASK_SIGN64) | ((y.w[1] << 2) & MASK_EXP128)
			y.w[0] = 0x0
		} else { // G0_G1 != 11
			if (y.w[1]&MASK_COEFF128) > 0x0001ed09bead87c0 ||
				((y.w[1]&MASK_COEFF128) == 0x0001ed09bead87c0 &&
					y.w[0] > 0x378d8e63ffffffff) {
				// y is non-canonical if coefficient is larger than 10^34 -1
				y.w[1] = (y.w[1] & MASK_SIGN64) | (y.w[1] & MASK_EXP128)
				y.w[0] = 0x0
			} else { // canonical
			}
		}
	}

	// NaN (CASE1)
	if (x.w[1] & NAN_MASK64) == NAN_MASK64 { // x is NAN
		if (x.w[1] & SNAN_MASK64) == SNAN_MASK64 { // x is SNaN
			// if x is SNAN, then return quiet (x)
			*pfpsf |= BID_INVALID_EXCEPTION // set exception if SNaN
			x.w[1] = x.w[1] & QUIET_MASK64
			res = x
		} else { // x is QNaN
			if (y.w[1] & NAN_MASK64) == NAN_MASK64 { // y is NAN
				if (y.w[1] & SNAN_MASK64) == SNAN_MASK64 { // y is SNAN
					*pfpsf |= BID_INVALID_EXCEPTION // set invalid flag
				}
				res = x
			} else {
				res = y
			}
		}
		return res
	} else if (y.w[1] & NAN_MASK64) == NAN_MASK64 { // y is NaN, but x is not
		if (y.w[1] & SNAN_MASK64) == SNAN_MASK64 {
			*pfpsf |= BID_INVALID_EXCEPTION // set exception if SNaN
			y.w[1] = y.w[1] & QUIET_MASK64
			res = y
		} else {
			// will return x (which is not NaN)
			res = x
		}
		return res
	}
	// SIMPLE (CASE2)
	// if all the bits are the same, these numbers are equal (not Greater).
	if x.w[0] == y.w[0] && x.w[1] == y.w[1] {
		res = x
		return res
	}
	// INFINITY (CASE3)
	if (x.w[1] & INFINITY_MASK64) == INFINITY_MASK64 {
		// if x is neg infinity, there is no way it is greater than y, return 0
		if (x.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = x
		} else {
			res = y
		}
		return res
	} else if (y.w[1] & INFINITY_MASK64) == INFINITY_MASK64 {
		// x is finite, so if y is positive infinity, then x is less, return 0
		//                 if y is negative infinity, then x is greater, return 1
		if (y.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = y
		} else {
			res = x
		}
		return res
	}
	// CONVERT X
	sig_x.w[1] = x.w[1] & 0x0001ffffffffffff
	sig_x.w[0] = x.w[0]
	exp_x = int((x.w[1] >> 49) & 0x000000000003fff)

	// CONVERT Y
	exp_y = int((y.w[1] >> 49) & 0x0000000000003fff)
	sig_y.w[1] = y.w[1] & 0x0001ffffffffffff
	sig_y.w[0] = y.w[0]

	// ZERO (CASE4)
	if (sig_x.w[1] == 0) && (sig_x.w[0] == 0) {
		x_is_zero = 1
	}
	if (sig_y.w[1] == 0) && (sig_y.w[0] == 0) {
		y_is_zero = 1
	}

	if x_is_zero != 0 && y_is_zero != 0 {
		// if both numbers are zero, neither is greater => return either number
		res = x
		return res
	} else if x_is_zero != 0 {
		// is x is zero, it is greater if Y is negative
		if (y.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = y
		} else {
			res = x
		}
		return res
	} else if y_is_zero != 0 {
		// is y is zero, X is greater if it is positive
		if (x.w[1] & MASK_SIGN64) != MASK_SIGN64 {
			res = y
		} else {
			res = x
		}
		return res
	}
	// OPPOSITE SIGN (CASE5)
	// now, if the sign bits differ, x is greater if y is negative
	if ((x.w[1] ^ y.w[1]) & MASK_SIGN64) == MASK_SIGN64 {
		if (y.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = y
		} else {
			res = x
		}
		return res
	}
	// REDUNDANT REPRESENTATIONS (CASE6)
	// if exponents are the same, then we have a simple comparison of
	//    the significands
	if exp_y == exp_x {
		if ((sig_x.w[1] > sig_y.w[1]) ||
			(sig_x.w[1] == sig_y.w[1] && sig_x.w[0] >= sig_y.w[0])) != ((x.w[1] & MASK_SIGN64) == MASK_SIGN64) {
			res = y
		} else {
			res = x
		}
		return res
	}
	// if both components are either bigger or smaller, it is clear what
	//    needs to be done
	if sig_x.w[1] >= sig_y.w[1] && sig_x.w[0] >= sig_y.w[0] && exp_x > exp_y {
		if (x.w[1] & MASK_SIGN64) != MASK_SIGN64 {
			res = y
		} else {
			res = x
		}
		return res
	}
	if sig_x.w[1] <= sig_y.w[1] && sig_x.w[0] <= sig_y.w[0] && exp_x < exp_y {
		if (x.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = y
		} else {
			res = x
		}
		return res
	}

	diff = exp_x - exp_y

	// if |exp_x - exp_y| < 33, it comes down to the compensated significand
	if diff > 0 { // to simplify the loop below,
		// if exp_x is 33 greater than exp_y, no need for compensation
		if diff > 33 {
			// difference cannot be greater than 10^33
			if (x.w[1] & MASK_SIGN64) != MASK_SIGN64 {
				res = y
			} else {
				res = x
			}
			return res
		}
		if diff > 19 { //128 by 128 bit multiply -> 256 bits
			sig_n_prime256 = __mul_128x128_to_256(sig_x, bid_ten2k128[diff-20])
			// if postitive, return whichever significand is larger
			// (converse if negative)
			if (((sig_n_prime256.w[3] > 0) || sig_n_prime256.w[2] > 0) ||
				(sig_n_prime256.w[1] > sig_y.w[1]) ||
				(sig_n_prime256.w[1] == sig_y.w[1] &&
					sig_n_prime256.w[0] > sig_y.w[0])) != ((y.w[1] & MASK_SIGN64) == MASK_SIGN64) {
				res = y
			} else {
				res = x
			}
			return res
		}
		sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[diff], sig_x)
		// if postitive, return whichever significand is larger
		// (converse if negative)
		if ((sig_n_prime192.w[2] > 0) || (sig_n_prime192.w[1] > sig_y.w[1]) ||
			(sig_n_prime192.w[1] == sig_y.w[1] &&
				sig_n_prime192.w[0] > sig_y.w[0])) != ((y.w[1] & MASK_SIGN64) == MASK_SIGN64) {
			res = y
		} else {
			res = x
		}
		return res
	}
	diff = exp_y - exp_x
	// if exp_x is 33 less than exp_y, no need for compensation
	if diff > 33 {
		if (x.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = y
		} else {
			res = x
		}
		return res
	}
	if diff > 19 { //128 by 128 bit multiply -> 256 bits
		// adjust the y significand upwards
		sig_n_prime256 = __mul_128x128_to_256(sig_y, bid_ten2k128[diff-20])
		// if postitive, return whichever significand is larger
		// (converse if negative)
		if (sig_n_prime256.w[3] != 0 || sig_n_prime256.w[2] != 0 ||
			(sig_n_prime256.w[1] > sig_x.w[1] ||
				(sig_n_prime256.w[1] == sig_x.w[1] &&
					sig_n_prime256.w[0] > sig_x.w[0]))) != ((x.w[1] & MASK_SIGN64) == MASK_SIGN64) {
			res = x
		} else {
			res = y
		}
		return res
	}
	// adjust the y significand upwards
	sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[diff], sig_y)
	// if postitive, return whichever significand is larger (converse if negative)
	if (sig_n_prime192.w[2] != 0 ||
		(sig_n_prime192.w[1] > sig_x.w[1] ||
			(sig_n_prime192.w[1] == sig_x.w[1] &&
				sig_n_prime192.w[0] > sig_x.w[0]))) != ((y.w[1] & MASK_SIGN64) == MASK_SIGN64) {
		res = x
	} else {
		res = y
	}
	return res
}

// Bid128MinnumMag returns the operand with the smaller magnitude.
// Ported from bid128_minnum_mag.
func Bid128MinnumMag(x, y BID_UINT128, pfpsf *uint32) BID_UINT128 {
	var res BID_UINT128
	var exp_x, exp_y int
	var diff int
	var sig_x, sig_y BID_UINT128
	var sig_n_prime192 BID_UINT192
	var sig_n_prime256 BID_UINT256

	// check for non-canonical x
	if (x.w[1] & NAN_MASK64) == NAN_MASK64 { // x is NAN
		x.w[1] = x.w[1] & 0xfe003fffffffffff // clear out G[6]-G[16]
		if ((x.w[1] & 0x00003fffffffffff) > 0x0000314dc6448d93) ||
			(((x.w[1] & 0x00003fffffffffff) == 0x0000314dc6448d93) &&
				(x.w[0] > 0x38c15b09ffffffff)) {
			x.w[1] = x.w[1] & 0xffffc00000000000
			x.w[0] = 0x0
		}
	} else if (x.w[1] & MASK_ANY_INF_128) == INFINITY_MASK64 { // x = inf
		x.w[1] = x.w[1] & (MASK_SIGN64 | INFINITY_MASK64)
		x.w[0] = 0x0
	} else { // x is not special
		if (x.w[1] & MASK_STEERING_BITS64) == MASK_STEERING_BITS64 { // G0_G1=11
			x.w[1] = (x.w[1] & MASK_SIGN64) | ((x.w[1] << 2) & MASK_EXP128)
			x.w[0] = 0x0
		} else { // G0_G1 != 11
			if (x.w[1]&MASK_COEFF128) > 0x0001ed09bead87c0 ||
				((x.w[1]&MASK_COEFF128) == 0x0001ed09bead87c0 &&
					x.w[0] > 0x378d8e63ffffffff) {
				x.w[1] = (x.w[1] & MASK_SIGN64) | (x.w[1] & MASK_EXP128)
				x.w[0] = 0x0
			} else { // canonical
			}
		}
	}
	// check for non-canonical y
	if (y.w[1] & NAN_MASK64) == NAN_MASK64 { // y is NAN
		y.w[1] = y.w[1] & 0xfe003fffffffffff // clear out G[6]-G[16]
		if ((y.w[1] & 0x00003fffffffffff) > 0x0000314dc6448d93) ||
			(((y.w[1] & 0x00003fffffffffff) == 0x0000314dc6448d93) &&
				(y.w[0] > 0x38c15b09ffffffff)) {
			y.w[1] = y.w[1] & 0xffffc00000000000
			y.w[0] = 0x0
		}
	} else if (y.w[1] & MASK_ANY_INF_128) == INFINITY_MASK64 { // y = inf
		y.w[1] = y.w[1] & (MASK_SIGN64 | INFINITY_MASK64)
		y.w[0] = 0x0
	} else { // y is not special
		if (y.w[1] & MASK_STEERING_BITS64) == MASK_STEERING_BITS64 { // G0_G1=11
			y.w[1] = (y.w[1] & MASK_SIGN64) | ((y.w[1] << 2) & MASK_EXP128)
			y.w[0] = 0x0
		} else { // G0_G1 != 11
			if (y.w[1]&MASK_COEFF128) > 0x0001ed09bead87c0 ||
				((y.w[1]&MASK_COEFF128) == 0x0001ed09bead87c0 &&
					y.w[0] > 0x378d8e63ffffffff) {
				y.w[1] = (y.w[1] & MASK_SIGN64) | (y.w[1] & MASK_EXP128)
				y.w[0] = 0x0
			} else { // canonical
			}
		}
	}

	// NaN (CASE1)
	if (x.w[1] & NAN_MASK64) == NAN_MASK64 { // x is NAN
		if (x.w[1] & SNAN_MASK64) == SNAN_MASK64 { // x is SNaN
			*pfpsf |= BID_INVALID_EXCEPTION
			x.w[1] = x.w[1] & QUIET_MASK64
			res = x
		} else { // x is QNaN
			if (y.w[1] & NAN_MASK64) == NAN_MASK64 { // y is NAN
				if (y.w[1] & SNAN_MASK64) == SNAN_MASK64 { // y is SNAN
					*pfpsf |= BID_INVALID_EXCEPTION
				}
				res = x
			} else {
				res = y
			}
		}
		return res
	} else if (y.w[1] & NAN_MASK64) == NAN_MASK64 { // y is NaN, but x is not
		if (y.w[1] & SNAN_MASK64) == SNAN_MASK64 {
			*pfpsf |= BID_INVALID_EXCEPTION
			y.w[1] = y.w[1] & QUIET_MASK64
			res = y
		} else {
			res = x
		}
		return res
	}
	// SIMPLE (CASE2)
	if x.w[0] == y.w[0] && x.w[1] == y.w[1] {
		res = y
		return res
	}
	// INFINITY (CASE3)
	if (x.w[1] & INFINITY_MASK64) == INFINITY_MASK64 {
		// if x infinity, it has maximum magnitude.
		// Check if magnitudes are equal. If x is negative, return it.
		if (x.w[1]&MASK_SIGN64) == MASK_SIGN64 && (y.w[1]&INFINITY_MASK64) == INFINITY_MASK64 {
			res = x
		} else {
			res = y
		}
		return res
	} else if (y.w[1] & INFINITY_MASK64) == INFINITY_MASK64 {
		// x is finite, so if y is infinity, then x is less in magnitude
		res = x
		return res
	}
	// CONVERT X
	sig_x.w[1] = x.w[1] & 0x0001ffffffffffff
	sig_x.w[0] = x.w[0]
	exp_x = int((x.w[1] >> 49) & 0x000000000003fff)

	// CONVERT Y
	exp_y = int((y.w[1] >> 49) & 0x0000000000003fff)
	sig_y.w[1] = y.w[1] & 0x0001ffffffffffff
	sig_y.w[0] = y.w[0]

	// ZERO (CASE4)
	if (sig_x.w[1] == 0) && (sig_x.w[0] == 0) {
		res = x
		return res
	}
	if (sig_y.w[1] == 0) && (sig_y.w[0] == 0) {
		res = y
		return res
	}
	// REDUNDANT REPRESENTATIONS (CASE6)
	// check if exponents are the same and significands are the same
	if exp_y == exp_x && sig_x.w[1] == sig_y.w[1] && sig_x.w[0] == sig_y.w[0] {
		if x.w[1]&0x8000000000000000 != 0 { // x is negative
			res = x
			return res
		} else {
			res = y
			return res
		}
	} else if ((sig_x.w[1] > sig_y.w[1] || (sig_x.w[1] == sig_y.w[1] && sig_x.w[0] > sig_y.w[0])) && exp_x == exp_y) ||
		((sig_x.w[1] > sig_y.w[1] || (sig_x.w[1] == sig_y.w[1] && sig_x.w[0] >= sig_y.w[0])) && exp_x > exp_y) {
		// if both components are either bigger or smaller, it is clear what
		// needs to be done; also if the magnitudes are equal
		res = y
		return res
	} else if ((sig_y.w[1] > sig_x.w[1] || (sig_y.w[1] == sig_x.w[1] && sig_y.w[0] > sig_x.w[0])) && exp_y == exp_x) ||
		((sig_y.w[1] > sig_x.w[1] || (sig_y.w[1] == sig_x.w[1] && sig_y.w[0] >= sig_x.w[0])) && exp_y > exp_x) {
		res = x
		return res
	} else {
		// continue
	}
	diff = exp_x - exp_y
	// if |exp_x - exp_y| < 33, it comes down to the compensated significand
	if diff > 0 { // to simplify the loop below,
		// if exp_x is 33 greater than exp_y, no need for compensation
		if diff > 33 {
			res = y // difference cannot be greater than 10^33
			return res
		}
		if diff > 19 { //128 by 128 bit multiply -> 256 bits
			sig_n_prime256 = __mul_128x128_to_256(sig_x, bid_ten2k128[diff-20])
			if sig_n_prime256.w[3] == 0 && (sig_n_prime256.w[2] == 0) &&
				sig_n_prime256.w[1] == sig_y.w[1] &&
				(sig_n_prime256.w[0] == sig_y.w[0]) {
				if (y.w[1] & MASK_SIGN64) == MASK_SIGN64 {
					res = y
				} else {
					res = x
				}
				return res
			}
			if ((sig_n_prime256.w[3] > 0) || sig_n_prime256.w[2] > 0) ||
				(sig_n_prime256.w[1] > sig_y.w[1]) ||
				(sig_n_prime256.w[1] == sig_y.w[1] && sig_n_prime256.w[0] > sig_y.w[0]) {
				res = y
			} else {
				res = x
			}
			return res
		}
		sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[diff], sig_x)
		if (sig_n_prime192.w[2] == 0) && sig_n_prime192.w[1] == sig_y.w[1] &&
			(sig_n_prime192.w[0] == sig_y.w[0]) {
			// if = in magnitude, return +, (if possible)
			if (y.w[1] & MASK_SIGN64) == MASK_SIGN64 {
				res = y
			} else {
				res = x
			}
			return res
		}
		if (sig_n_prime192.w[2] > 0) ||
			(sig_n_prime192.w[1] > sig_y.w[1]) ||
			(sig_n_prime192.w[1] == sig_y.w[1] && sig_n_prime192.w[0] > sig_y.w[0]) {
			res = y
		} else {
			res = x
		}
		return res
	}
	diff = exp_y - exp_x
	// if exp_x is 33 less than exp_y, no need for compensation
	if diff > 33 {
		res = x
		return res
	}
	if diff > 19 { //128 by 128 bit multiply -> 256 bits
		// adjust the y significand upwards
		sig_n_prime256 = __mul_128x128_to_256(sig_y, bid_ten2k128[diff-20])
		if sig_n_prime256.w[3] == 0 && (sig_n_prime256.w[2] == 0) &&
			sig_n_prime256.w[1] == sig_x.w[1] &&
			(sig_n_prime256.w[0] == sig_x.w[0]) {
			// if = in magnitude, return +, (if possible)
			if (y.w[1] & MASK_SIGN64) == MASK_SIGN64 {
				res = y
			} else {
				res = x
			}
			return res
		}
		if sig_n_prime256.w[3] == 0 && sig_n_prime256.w[2] == 0 &&
			(sig_n_prime256.w[1] < sig_x.w[1] ||
				(sig_n_prime256.w[1] == sig_x.w[1] && sig_n_prime256.w[0] < sig_x.w[0])) {
			res = y
		} else {
			res = x
		}
		return res
	}
	// adjust the y significand upwards
	sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[diff], sig_y)
	if (sig_n_prime192.w[2] == 0) && sig_n_prime192.w[1] == sig_x.w[1] &&
		(sig_n_prime192.w[0] == sig_x.w[0]) {
		// if = in magnitude, return +, if possible)
		if (y.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = y
		} else {
			res = x
		}
		return res
	}
	if sig_n_prime192.w[2] == 0 &&
		(sig_n_prime192.w[1] < sig_x.w[1] ||
			(sig_n_prime192.w[1] == sig_x.w[1] && sig_n_prime192.w[0] < sig_x.w[0])) {
		res = y
	} else {
		res = x
	}
	return res
}

// Bid128Maxnum returns the maximum of two numbers.
// Ported from bid128_maxnum.
func Bid128Maxnum(x, y BID_UINT128, pfpsf *uint32) BID_UINT128 {
	var res BID_UINT128
	var exp_x, exp_y int
	var diff int
	var sig_x, sig_y BID_UINT128
	var sig_n_prime192 BID_UINT192
	var sig_n_prime256 BID_UINT256
	var x_is_zero, y_is_zero byte

	// check for non-canonical x
	if (x.w[1] & NAN_MASK64) == NAN_MASK64 { // x is NAN
		x.w[1] = x.w[1] & 0xfe003fffffffffff
		if ((x.w[1] & 0x00003fffffffffff) > 0x0000314dc6448d93) ||
			(((x.w[1] & 0x00003fffffffffff) == 0x0000314dc6448d93) &&
				(x.w[0] > 0x38c15b09ffffffff)) {
			x.w[1] = x.w[1] & 0xffffc00000000000
			x.w[0] = 0x0
		}
	} else if (x.w[1] & MASK_ANY_INF_128) == INFINITY_MASK64 {
		x.w[1] = x.w[1] & (MASK_SIGN64 | INFINITY_MASK64)
		x.w[0] = 0x0
	} else {
		if (x.w[1] & MASK_STEERING_BITS64) == MASK_STEERING_BITS64 {
			x.w[1] = (x.w[1] & MASK_SIGN64) | ((x.w[1] << 2) & MASK_EXP128)
			x.w[0] = 0x0
		} else {
			if (x.w[1]&MASK_COEFF128) > 0x0001ed09bead87c0 ||
				((x.w[1]&MASK_COEFF128) == 0x0001ed09bead87c0 && x.w[0] > 0x378d8e63ffffffff) {
				x.w[1] = (x.w[1] & MASK_SIGN64) | (x.w[1] & MASK_EXP128)
				x.w[0] = 0x0
			}
		}
	}
	// check for non-canonical y
	if (y.w[1] & NAN_MASK64) == NAN_MASK64 {
		y.w[1] = y.w[1] & 0xfe003fffffffffff
		if ((y.w[1] & 0x00003fffffffffff) > 0x0000314dc6448d93) ||
			(((y.w[1] & 0x00003fffffffffff) == 0x0000314dc6448d93) &&
				(y.w[0] > 0x38c15b09ffffffff)) {
			y.w[1] = y.w[1] & 0xffffc00000000000
			y.w[0] = 0x0
		}
	} else if (y.w[1] & MASK_ANY_INF_128) == INFINITY_MASK64 {
		y.w[1] = y.w[1] & (MASK_SIGN64 | INFINITY_MASK64)
		y.w[0] = 0x0
	} else {
		if (y.w[1] & MASK_STEERING_BITS64) == MASK_STEERING_BITS64 {
			y.w[1] = (y.w[1] & MASK_SIGN64) | ((y.w[1] << 2) & MASK_EXP128)
			y.w[0] = 0x0
		} else {
			if (y.w[1]&MASK_COEFF128) > 0x0001ed09bead87c0 ||
				((y.w[1]&MASK_COEFF128) == 0x0001ed09bead87c0 && y.w[0] > 0x378d8e63ffffffff) {
				y.w[1] = (y.w[1] & MASK_SIGN64) | (y.w[1] & MASK_EXP128)
				y.w[0] = 0x0
			}
		}
	}

	// NaN (CASE1)
	if (x.w[1] & NAN_MASK64) == NAN_MASK64 {
		if (x.w[1] & SNAN_MASK64) == SNAN_MASK64 {
			*pfpsf |= BID_INVALID_EXCEPTION
			x.w[1] = x.w[1] & QUIET_MASK64
			res = x
		} else {
			if (y.w[1] & NAN_MASK64) == NAN_MASK64 {
				if (y.w[1] & SNAN_MASK64) == SNAN_MASK64 {
					*pfpsf |= BID_INVALID_EXCEPTION
				}
				res = x
			} else {
				res = y
			}
		}
		return res
	} else if (y.w[1] & NAN_MASK64) == NAN_MASK64 {
		if (y.w[1] & SNAN_MASK64) == SNAN_MASK64 {
			*pfpsf |= BID_INVALID_EXCEPTION
			y.w[1] = y.w[1] & QUIET_MASK64
			res = y
		} else {
			res = x
		}
		return res
	}
	// SIMPLE (CASE2)
	if x.w[0] == y.w[0] && x.w[1] == y.w[1] {
		res = x
		return res
	}
	// INFINITY (CASE3)
	if (x.w[1] & INFINITY_MASK64) == INFINITY_MASK64 {
		if (x.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = y
		} else {
			res = x
		}
		return res
	} else if (y.w[1] & INFINITY_MASK64) == INFINITY_MASK64 {
		if (y.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = x
		} else {
			res = y
		}
		return res
	}
	// CONVERT X
	sig_x.w[1] = x.w[1] & 0x0001ffffffffffff
	sig_x.w[0] = x.w[0]
	exp_x = int((x.w[1] >> 49) & 0x000000000003fff)

	// CONVERT Y
	exp_y = int((y.w[1] >> 49) & 0x0000000000003fff)
	sig_y.w[1] = y.w[1] & 0x0001ffffffffffff
	sig_y.w[0] = y.w[0]

	// ZERO (CASE4)
	if (sig_x.w[1] == 0) && (sig_x.w[0] == 0) {
		x_is_zero = 1
	}
	if (sig_y.w[1] == 0) && (sig_y.w[0] == 0) {
		y_is_zero = 1
	}

	if x_is_zero != 0 && y_is_zero != 0 {
		res = x
		return res
	} else if x_is_zero != 0 {
		if (y.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = x
		} else {
			res = y
		}
		return res
	} else if y_is_zero != 0 {
		if (x.w[1] & MASK_SIGN64) != MASK_SIGN64 {
			res = x
		} else {
			res = y
		}
		return res
	}
	// OPPOSITE SIGN (CASE5)
	if ((x.w[1] ^ y.w[1]) & MASK_SIGN64) == MASK_SIGN64 {
		if (y.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = x
		} else {
			res = y
		}
		return res
	}
	// REDUNDANT REPRESENTATIONS (CASE6)
	if exp_y == exp_x {
		if ((sig_x.w[1] > sig_y.w[1]) || (sig_x.w[1] == sig_y.w[1] &&
			sig_x.w[0] >= sig_y.w[0])) != ((x.w[1] & MASK_SIGN64) == MASK_SIGN64) {
			res = x
		} else {
			res = y
		}
		return res
	}
	// if both components are either bigger or smaller
	if (sig_x.w[1] > sig_y.w[1] || (sig_x.w[1] == sig_y.w[1] && sig_x.w[0] > sig_y.w[0])) && exp_x >= exp_y {
		if (x.w[1] & MASK_SIGN64) != MASK_SIGN64 {
			res = x
		} else {
			res = y
		}
		return res
	}
	if (sig_x.w[1] < sig_y.w[1] || (sig_x.w[1] == sig_y.w[1] && sig_x.w[0] < sig_y.w[0])) && exp_x <= exp_y {
		if (x.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = x
		} else {
			res = y
		}
		return res
	}
	diff = exp_x - exp_y
	if diff > 0 {
		if diff > 33 {
			if (x.w[1] & MASK_SIGN64) != MASK_SIGN64 {
				res = x
			} else {
				res = y
			}
			return res
		}
		if diff > 19 {
			sig_n_prime256 = __mul_128x128_to_256(sig_x, bid_ten2k128[diff-20])
			if (((sig_n_prime256.w[3] > 0) || sig_n_prime256.w[2] > 0) ||
				(sig_n_prime256.w[1] > sig_y.w[1]) ||
				(sig_n_prime256.w[1] == sig_y.w[1] && sig_n_prime256.w[0] > sig_y.w[0])) != ((y.w[1] & MASK_SIGN64) == MASK_SIGN64) {
				res = x
			} else {
				res = y
			}
			return res
		}
		sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[diff], sig_x)
		if ((sig_n_prime192.w[2] > 0) || (sig_n_prime192.w[1] > sig_y.w[1]) ||
			(sig_n_prime192.w[1] == sig_y.w[1] && sig_n_prime192.w[0] > sig_y.w[0])) != ((y.w[1] & MASK_SIGN64) == MASK_SIGN64) {
			res = x
		} else {
			res = y
		}
		return res
	}
	diff = exp_y - exp_x
	if diff > 33 {
		if (x.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = x
		} else {
			res = y
		}
		return res
	}
	if diff > 19 {
		sig_n_prime256 = __mul_128x128_to_256(sig_y, bid_ten2k128[diff-20])
		if (sig_n_prime256.w[3] != 0 || sig_n_prime256.w[2] != 0 ||
			(sig_n_prime256.w[1] > sig_x.w[1] ||
				(sig_n_prime256.w[1] == sig_x.w[1] && sig_n_prime256.w[0] > sig_x.w[0]))) != ((x.w[1] & MASK_SIGN64) != MASK_SIGN64) {
			res = x
		} else {
			res = y
		}
		return res
	}
	sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[diff], sig_y)
	if (sig_n_prime192.w[2] != 0 ||
		(sig_n_prime192.w[1] > sig_x.w[1] ||
			(sig_n_prime192.w[1] == sig_x.w[1] && sig_n_prime192.w[0] > sig_x.w[0]))) != ((y.w[1] & MASK_SIGN64) != MASK_SIGN64) {
		res = x
	} else {
		res = y
	}
	return res
}

// Bid128MaxnumMag returns the operand with the larger magnitude.
// Ported from bid128_maxnum_mag.
func Bid128MaxnumMag(x, y BID_UINT128, pfpsf *uint32) BID_UINT128 {
	var res BID_UINT128
	var exp_x, exp_y int
	var diff int
	var sig_x, sig_y BID_UINT128
	var sig_n_prime192 BID_UINT192
	var sig_n_prime256 BID_UINT256

	// check for non-canonical x
	if (x.w[1] & NAN_MASK64) == NAN_MASK64 { // x is NAN
		x.w[1] = x.w[1] & 0xfe003fffffffffff
		if ((x.w[1] & 0x00003fffffffffff) > 0x0000314dc6448d93) ||
			(((x.w[1] & 0x00003fffffffffff) == 0x0000314dc6448d93) &&
				(x.w[0] > 0x38c15b09ffffffff)) {
			x.w[1] = x.w[1] & 0xffffc00000000000
			x.w[0] = 0x0
		}
	} else if (x.w[1] & MASK_ANY_INF_128) == INFINITY_MASK64 {
		x.w[1] = x.w[1] & (MASK_SIGN64 | INFINITY_MASK64)
		x.w[0] = 0x0
	} else {
		if (x.w[1] & MASK_STEERING_BITS64) == MASK_STEERING_BITS64 {
			x.w[1] = (x.w[1] & MASK_SIGN64) | ((x.w[1] << 2) & MASK_EXP128)
			x.w[0] = 0x0
		} else {
			if (x.w[1]&MASK_COEFF128) > 0x0001ed09bead87c0 ||
				((x.w[1]&MASK_COEFF128) == 0x0001ed09bead87c0 && x.w[0] > 0x378d8e63ffffffff) {
				x.w[1] = (x.w[1] & MASK_SIGN64) | (x.w[1] & MASK_EXP128)
				x.w[0] = 0x0
			}
		}
	}
	// check for non-canonical y
	if (y.w[1] & NAN_MASK64) == NAN_MASK64 {
		y.w[1] = y.w[1] & 0xfe003fffffffffff
		if ((y.w[1] & 0x00003fffffffffff) > 0x0000314dc6448d93) ||
			(((y.w[1] & 0x00003fffffffffff) == 0x0000314dc6448d93) &&
				(y.w[0] > 0x38c15b09ffffffff)) {
			y.w[1] = y.w[1] & 0xffffc00000000000
			y.w[0] = 0x0
		}
	} else if (y.w[1] & MASK_ANY_INF_128) == INFINITY_MASK64 {
		y.w[1] = y.w[1] & (MASK_SIGN64 | INFINITY_MASK64)
		y.w[0] = 0x0
	} else {
		if (y.w[1] & MASK_STEERING_BITS64) == MASK_STEERING_BITS64 {
			y.w[1] = (y.w[1] & MASK_SIGN64) | ((y.w[1] << 2) & MASK_EXP128)
			y.w[0] = 0x0
		} else {
			if (y.w[1]&MASK_COEFF128) > 0x0001ed09bead87c0 ||
				((y.w[1]&MASK_COEFF128) == 0x0001ed09bead87c0 && y.w[0] > 0x378d8e63ffffffff) {
				y.w[1] = (y.w[1] & MASK_SIGN64) | (y.w[1] & MASK_EXP128)
				y.w[0] = 0x0
			}
		}
	}

	// NaN (CASE1)
	if (x.w[1] & NAN_MASK64) == NAN_MASK64 {
		if (x.w[1] & SNAN_MASK64) == SNAN_MASK64 {
			*pfpsf |= BID_INVALID_EXCEPTION
			x.w[1] = x.w[1] & QUIET_MASK64
			res = x
		} else {
			if (y.w[1] & NAN_MASK64) == NAN_MASK64 {
				if (y.w[1] & SNAN_MASK64) == SNAN_MASK64 {
					*pfpsf |= BID_INVALID_EXCEPTION
				}
				res = x
			} else {
				res = y
			}
		}
		return res
	} else if (y.w[1] & NAN_MASK64) == NAN_MASK64 {
		if (y.w[1] & SNAN_MASK64) == SNAN_MASK64 {
			*pfpsf |= BID_INVALID_EXCEPTION
			y.w[1] = y.w[1] & QUIET_MASK64
			res = y
		} else {
			res = x
		}
		return res
	}
	// SIMPLE (CASE2)
	if x.w[0] == y.w[0] && x.w[1] == y.w[1] {
		res = y
		return res
	}
	// INFINITY (CASE3)
	if (x.w[1] & INFINITY_MASK64) == INFINITY_MASK64 {
		// if x infinity, it has maximum magnitude
		if (x.w[1]&MASK_SIGN64) == MASK_SIGN64 && (y.w[1]&INFINITY_MASK64) == INFINITY_MASK64 {
			res = y
		} else {
			res = x
		}
		return res
	} else if (y.w[1] & INFINITY_MASK64) == INFINITY_MASK64 {
		res = y
		return res
	}
	// CONVERT X
	sig_x.w[1] = x.w[1] & 0x0001ffffffffffff
	sig_x.w[0] = x.w[0]
	exp_x = int((x.w[1] >> 49) & 0x000000000003fff)

	// CONVERT Y
	exp_y = int((y.w[1] >> 49) & 0x0000000000003fff)
	sig_y.w[1] = y.w[1] & 0x0001ffffffffffff
	sig_y.w[0] = y.w[0]

	// ZERO (CASE4)
	if (sig_x.w[1] == 0) && (sig_x.w[0] == 0) {
		res = y
		return res
	}
	if (sig_y.w[1] == 0) && (sig_y.w[0] == 0) {
		res = x
		return res
	}
	// REDUNDANT REPRESENTATIONS (CASE6)
	if exp_y == exp_x && sig_x.w[1] == sig_y.w[1] && sig_x.w[0] == sig_y.w[0] {
		if x.w[1]&0x8000000000000000 != 0 { // x is negative
			res = y
			return res
		} else {
			res = x
			return res
		}
	} else if ((sig_x.w[1] > sig_y.w[1] || (sig_x.w[1] == sig_y.w[1] && sig_x.w[0] > sig_y.w[0])) && exp_x == exp_y) ||
		((sig_x.w[1] > sig_y.w[1] || (sig_x.w[1] == sig_y.w[1] && sig_x.w[0] >= sig_y.w[0])) && exp_x > exp_y) {
		res = x
		return res
	} else if ((sig_y.w[1] > sig_x.w[1] || (sig_y.w[1] == sig_x.w[1] && sig_y.w[0] > sig_x.w[0])) && exp_y == exp_x) ||
		((sig_y.w[1] > sig_x.w[1] || (sig_y.w[1] == sig_x.w[1] && sig_y.w[0] >= sig_x.w[0])) && exp_y > exp_x) {
		res = y
		return res
	} else {
		// continue
	}
	diff = exp_x - exp_y
	if diff > 0 {
		if diff > 33 {
			res = x
			return res
		}
		if diff > 19 {
			sig_n_prime256 = __mul_128x128_to_256(sig_x, bid_ten2k128[diff-20])
			if sig_n_prime256.w[3] == 0 && (sig_n_prime256.w[2] == 0) &&
				sig_n_prime256.w[1] == sig_y.w[1] &&
				(sig_n_prime256.w[0] == sig_y.w[0]) {
				if (y.w[1] & MASK_SIGN64) == MASK_SIGN64 {
					res = x
				} else {
					res = y
				}
				return res
			}
			if ((sig_n_prime256.w[3] > 0) || sig_n_prime256.w[2] > 0) ||
				(sig_n_prime256.w[1] > sig_y.w[1]) ||
				(sig_n_prime256.w[1] == sig_y.w[1] && sig_n_prime256.w[0] > sig_y.w[0]) {
				res = x
			} else {
				res = y
			}
			return res
		}
		sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[diff], sig_x)
		if (sig_n_prime192.w[2] == 0) && sig_n_prime192.w[1] == sig_y.w[1] &&
			(sig_n_prime192.w[0] == sig_y.w[0]) {
			if (y.w[1] & MASK_SIGN64) == MASK_SIGN64 {
				res = x
			} else {
				res = y
			}
			return res
		}
		if (sig_n_prime192.w[2] > 0) ||
			(sig_n_prime192.w[1] > sig_y.w[1]) ||
			(sig_n_prime192.w[1] == sig_y.w[1] && sig_n_prime192.w[0] > sig_y.w[0]) {
			res = x
		} else {
			res = y
		}
		return res
	}
	diff = exp_y - exp_x
	if diff > 33 {
		res = y
		return res
	}
	if diff > 19 {
		sig_n_prime256 = __mul_128x128_to_256(sig_y, bid_ten2k128[diff-20])
		if sig_n_prime256.w[3] == 0 && (sig_n_prime256.w[2] == 0) &&
			sig_n_prime256.w[1] == sig_x.w[1] &&
			(sig_n_prime256.w[0] == sig_x.w[0]) {
			if (y.w[1] & MASK_SIGN64) == MASK_SIGN64 {
				res = x
			} else {
				res = y
			}
			return res
		}
		if sig_n_prime256.w[3] == 0 && sig_n_prime256.w[2] == 0 &&
			(sig_n_prime256.w[1] < sig_x.w[1] ||
				(sig_n_prime256.w[1] == sig_x.w[1] && sig_n_prime256.w[0] < sig_x.w[0])) {
			res = x
		} else {
			res = y
		}
		return res
	}
	sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[diff], sig_y)
	if (sig_n_prime192.w[2] == 0) && sig_n_prime192.w[1] == sig_x.w[1] &&
		(sig_n_prime192.w[0] == sig_x.w[0]) {
		if (y.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = x
		} else {
			res = y
		}
		return res
	}
	if sig_n_prime192.w[2] == 0 &&
		(sig_n_prime192.w[1] < sig_x.w[1] ||
			(sig_n_prime192.w[1] == sig_x.w[1] && sig_n_prime192.w[0] < sig_x.w[0])) {
		res = x
	} else {
		res = y
	}
	return res
}
