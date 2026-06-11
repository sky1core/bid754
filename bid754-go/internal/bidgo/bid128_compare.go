// Ported from: Intel bid128_compare.c and bid128_noncomp.c (totalOrder, totalOrderMag)
// Mechanical translation - all logic preserved exactly.

package bidgo

// Bid128QuietEqual - Intel bid128_quiet_equal 기계적 포팅
func Bid128QuietEqual(x, y BID_UINT128) (int, uint32) {
	var res int
	var exp_x, exp_y, exp_t int
	var sig_x, sig_y, sig_t BID_UINT128
	var sig_n_prime192 BID_UINT192
	var sig_n_prime256 BID_UINT256
	var pfpsf uint32
	var x_is_zero, y_is_zero int
	var non_canon_x, non_canon_y int

	// NaN (CASE1)
	// if either number is NAN, the comparison is unordered,
	// rather than equal : return 0
	if ((x.w[1] & NAN_MASK64) == NAN_MASK64) ||
		((y.w[1] & NAN_MASK64) == NAN_MASK64) {
		if (x.w[1]&SNAN_MASK64) == SNAN_MASK64 ||
			(y.w[1]&SNAN_MASK64) == SNAN_MASK64 {
			pfpsf |= BID_INVALID_EXCEPTION
		}
		res = 0
		return res, pfpsf
	}
	// SIMPLE (CASE2)
	// if all the bits are the same, these numbers are equivalent.
	if x.w[0] == y.w[0] && x.w[1] == y.w[1] {
		res = 1
		return res, pfpsf
	}
	// INFINITY (CASE3)
	if (x.w[1] & INFINITY_MASK64) == INFINITY_MASK64 {
		if (y.w[1] & INFINITY_MASK64) == INFINITY_MASK64 {
			res = 0
			if ((x.w[1] ^ y.w[1]) & MASK_SIGN64) != MASK_SIGN64 {
				res = 1
			}
			return res, pfpsf
		} else {
			res = 0
			return res, pfpsf
		}
	}
	if (y.w[1] & INFINITY_MASK64) == INFINITY_MASK64 {
		res = 0
		return res, pfpsf
	}
	// CONVERT X
	sig_x.w[1] = x.w[1] & 0x0001ffffffffffff
	sig_x.w[0] = x.w[0]
	exp_x = int((x.w[1] >> 49) & 0x000000000003fff)

	// CHECK IF X IS CANONICAL
	if (sig_x.w[1] > 0x0001ed09bead87c0) ||
		((sig_x.w[1] == 0x0001ed09bead87c0) &&
			(sig_x.w[0] > 0x378d8e63ffffffff)) ||
		((x.w[1] & 0x6000000000000000) == 0x6000000000000000) {
		non_canon_x = 1
	} else {
		non_canon_x = 0
	}

	// CONVERT Y
	exp_y = int((y.w[1] >> 49) & 0x0000000000003fff)
	sig_y.w[1] = y.w[1] & 0x0001ffffffffffff
	sig_y.w[0] = y.w[0]

	// CHECK IF Y IS CANONICAL
	if (sig_y.w[1] > 0x0001ed09bead87c0) ||
		((sig_y.w[1] == 0x0001ed09bead87c0) &&
			(sig_y.w[0] > 0x378d8e63ffffffff)) ||
		((y.w[1] & 0x6000000000000000) == 0x6000000000000000) {
		non_canon_y = 1
	} else {
		non_canon_y = 0
	}

	// some properties:
	//    (+ZERO == -ZERO) => therefore ignore the sign
	//    (ZERO x 10^A == ZERO x 10^B) for any valid A, B => therefore
	//    ignore the exponent field
	//    (Any non-canonical # is considered 0)
	if non_canon_x != 0 || ((sig_x.w[1] == 0) && (sig_x.w[0] == 0)) {
		x_is_zero = 1
	}
	if non_canon_y != 0 || ((sig_y.w[1] == 0) && (sig_y.w[0] == 0)) {
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
	if (x.w[1]^y.w[1])&MASK_SIGN64 != 0 {
		res = 0
		return res, pfpsf
	}
	// REDUNDANT REPRESENTATIONS (CASE6)
	if exp_x > exp_y { // to simplify the loop below,
		exp_t = exp_x // put the larger exp in y,
		exp_x = exp_y
		exp_y = exp_t
		sig_t.w[1] = sig_x.w[1] // and the smaller exp in x
		sig_x.w[1] = sig_y.w[1]
		sig_y.w[1] = sig_t.w[1]
		sig_t.w[0] = sig_x.w[0] // and the smaller exp in x
		sig_x.w[0] = sig_y.w[0]
		sig_y.w[0] = sig_t.w[0]
	}

	if exp_y-exp_x > 33 {
		res = 0
		return res, pfpsf
	} // difference cannot be greater than 10^33

	if exp_y-exp_x > 19 {
		// recalculate y's significand upwards
		sig_n_prime256 = __mul_128x128_to_256(sig_y,
			bid_ten2k128[exp_y-exp_x-20])
		res = 0
		if (sig_n_prime256.w[3] == 0) && (sig_n_prime256.w[2] == 0) &&
			(sig_n_prime256.w[1] == sig_x.w[1]) &&
			(sig_n_prime256.w[0] == sig_x.w[0]) {
			res = 1
		}
		return res, pfpsf
	}
	// recalculate y's significand upwards
	sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[exp_y-exp_x], sig_y)
	res = 0
	if (sig_n_prime192.w[2] == 0) &&
		(sig_n_prime192.w[1] == sig_x.w[1]) &&
		(sig_n_prime192.w[0] == sig_x.w[0]) {
		res = 1
	}
	return res, pfpsf
}

// Bid128QuietGreater - Intel bid128_quiet_greater 기계적 포팅
func Bid128QuietGreater(x, y BID_UINT128) (int, uint32) {
	var res int
	var exp_x, exp_y int
	var diff int
	var sig_x, sig_y BID_UINT128
	var sig_n_prime192 BID_UINT192
	var sig_n_prime256 BID_UINT256
	var pfpsf uint32
	var x_is_zero, y_is_zero, non_canon_x, non_canon_y int

	// NaN (CASE1)
	if ((x.w[1] & NAN_MASK64) == NAN_MASK64) ||
		((y.w[1] & NAN_MASK64) == NAN_MASK64) {
		if (x.w[1]&SNAN_MASK64) == SNAN_MASK64 ||
			(y.w[1]&SNAN_MASK64) == SNAN_MASK64 {
			pfpsf |= BID_INVALID_EXCEPTION
		}
		res = 0
		return res, pfpsf
	}
	// SIMPLE (CASE2)
	if x.w[0] == y.w[0] && x.w[1] == y.w[1] {
		res = 0
		return res, pfpsf
	}
	// INFINITY (CASE3)
	if (x.w[1] & INFINITY_MASK64) == INFINITY_MASK64 {
		if (x.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = 0
			return res, pfpsf
		} else {
			res = 0
			if ((y.w[1] & INFINITY_MASK64) != INFINITY_MASK64) ||
				((y.w[1] & MASK_SIGN64) == MASK_SIGN64) {
				res = 1
			}
			return res, pfpsf
		}
	} else if (y.w[1] & INFINITY_MASK64) == INFINITY_MASK64 {
		res = 0
		if (y.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// CONVERT X
	sig_x.w[1] = x.w[1] & 0x0001ffffffffffff
	sig_x.w[0] = x.w[0]
	exp_x = int((x.w[1] >> 49) & 0x000000000003fff)

	// CHECK IF X IS CANONICAL
	if (sig_x.w[1] > 0x0001ed09bead87c0) ||
		((sig_x.w[1] == 0x0001ed09bead87c0) &&
			(sig_x.w[0] > 0x378d8e63ffffffff)) ||
		((x.w[1] & 0x6000000000000000) == 0x6000000000000000) {
		non_canon_x = 1
	} else {
		non_canon_x = 0
	}

	// CONVERT Y
	exp_y = int((y.w[1] >> 49) & 0x0000000000003fff)
	sig_y.w[1] = y.w[1] & 0x0001ffffffffffff
	sig_y.w[0] = y.w[0]

	// CHECK IF Y IS CANONICAL
	if (sig_y.w[1] > 0x0001ed09bead87c0) ||
		((sig_y.w[1] == 0x0001ed09bead87c0) &&
			(sig_y.w[0] > 0x378d8e63ffffffff)) ||
		((y.w[1] & 0x6000000000000000) == 0x6000000000000000) {
		non_canon_y = 1
	} else {
		non_canon_y = 0
	}

	// ZERO (CASE4)
	if non_canon_x != 0 || ((sig_x.w[1] == 0) && (sig_x.w[0] == 0)) {
		x_is_zero = 1
	}
	if non_canon_y != 0 || ((sig_y.w[1] == 0) && (sig_y.w[0] == 0)) {
		y_is_zero = 1
	}
	if x_is_zero != 0 && y_is_zero != 0 {
		res = 0
		return res, pfpsf
	} else if x_is_zero != 0 {
		res = 0
		if (y.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	} else if y_is_zero != 0 {
		res = 0
		if (x.w[1] & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// OPPOSITE SIGN (CASE5)
	if ((x.w[1] ^ y.w[1]) & MASK_SIGN64) == MASK_SIGN64 {
		res = 0
		if (y.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// REDUNDANT REPRESENTATIONS (CASE6)
	if exp_y == exp_x {
		res = 0
		if ((sig_x.w[1] > sig_y.w[1]) ||
			(sig_x.w[1] == sig_y.w[1] && sig_x.w[0] >= sig_y.w[0])) !=
			((x.w[1] & MASK_SIGN64) == MASK_SIGN64) {
			res = 1
		}
		return res, pfpsf
	}
	if (sig_x.w[1] > sig_y.w[1] ||
		(sig_x.w[1] == sig_y.w[1] && sig_x.w[0] > sig_y.w[0])) &&
		exp_x >= exp_y {
		res = 0
		if (x.w[1] & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	if (sig_x.w[1] < sig_y.w[1] ||
		(sig_x.w[1] == sig_y.w[1] && sig_x.w[0] < sig_y.w[0])) &&
		exp_x <= exp_y {
		res = 0
		if (x.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}

	diff = exp_x - exp_y

	if diff > 0 {
		if diff > 33 {
			res = 0
			if (x.w[1] & MASK_SIGN64) != MASK_SIGN64 {
				res = 1
			}
			return res, pfpsf
		}

		if diff > 19 {
			sig_n_prime256 = __mul_128x128_to_256(sig_x, bid_ten2k128[diff-20])
			if sig_n_prime256.w[3] == 0 && (sig_n_prime256.w[2] == 0) &&
				sig_n_prime256.w[1] == sig_y.w[1] &&
				(sig_n_prime256.w[0] == sig_y.w[0]) {
				res = 0
				return res, pfpsf
			}
			res = 0
			if (((sig_n_prime256.w[3] > 0) || sig_n_prime256.w[2] > 0) ||
				(sig_n_prime256.w[1] > sig_y.w[1]) ||
				(sig_n_prime256.w[1] == sig_y.w[1] &&
					sig_n_prime256.w[0] > sig_y.w[0])) !=
				((y.w[1] & MASK_SIGN64) == MASK_SIGN64) {
				res = 1
			}
			return res, pfpsf
		}
		sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[diff], sig_x)
		if (sig_n_prime192.w[2] == 0) && sig_n_prime192.w[1] == sig_y.w[1] &&
			(sig_n_prime192.w[0] == sig_y.w[0]) {
			res = 0
			return res, pfpsf
		}
		res = 0
		if ((sig_n_prime192.w[2] > 0) ||
			(sig_n_prime192.w[1] > sig_y.w[1]) ||
			(sig_n_prime192.w[1] == sig_y.w[1] &&
				sig_n_prime192.w[0] > sig_y.w[0])) !=
			((y.w[1] & MASK_SIGN64) == MASK_SIGN64) {
			res = 1
		}
		return res, pfpsf
	}

	diff = exp_y - exp_x

	if diff > 33 {
		res = 0
		if (x.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}

	if diff > 19 {
		sig_n_prime256 = __mul_128x128_to_256(sig_y, bid_ten2k128[diff-20])
		if sig_n_prime256.w[3] == 0 && (sig_n_prime256.w[2] == 0) &&
			sig_n_prime256.w[1] == sig_x.w[1] &&
			(sig_n_prime256.w[0] == sig_x.w[0]) {
			res = 0
			return res, pfpsf
		}
		res = 0
		if (sig_n_prime256.w[3] != 0 || sig_n_prime256.w[2] != 0 ||
			(sig_n_prime256.w[1] > sig_x.w[1] ||
				(sig_n_prime256.w[1] == sig_x.w[1] &&
					sig_n_prime256.w[0] > sig_x.w[0]))) !=
			((x.w[1] & MASK_SIGN64) != MASK_SIGN64) {
			res = 1
		}
		return res, pfpsf
	}
	sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[diff], sig_y)
	if (sig_n_prime192.w[2] == 0) && sig_n_prime192.w[1] == sig_x.w[1] &&
		(sig_n_prime192.w[0] == sig_x.w[0]) {
		res = 0
		return res, pfpsf
	}
	res = 0
	if (sig_n_prime192.w[2] != 0 ||
		(sig_n_prime192.w[1] > sig_x.w[1] ||
			(sig_n_prime192.w[1] == sig_x.w[1] &&
				sig_n_prime192.w[0] > sig_x.w[0]))) !=
		((y.w[1] & MASK_SIGN64) != MASK_SIGN64) {
		res = 1
	}
	return res, pfpsf
}

// Bid128QuietGreaterEqual - Intel bid128_quiet_greater_equal 기계적 포팅
func Bid128QuietGreaterEqual(x, y BID_UINT128) (int, uint32) {
	var res int
	var exp_x, exp_y int
	var diff int
	var sig_x, sig_y BID_UINT128
	var sig_n_prime192 BID_UINT192
	var sig_n_prime256 BID_UINT256
	var pfpsf uint32
	var x_is_zero, y_is_zero, non_canon_x, non_canon_y int

	// NaN (CASE1)
	if ((x.w[1] & NAN_MASK64) == NAN_MASK64) ||
		((y.w[1] & NAN_MASK64) == NAN_MASK64) {
		if (x.w[1]&SNAN_MASK64) == SNAN_MASK64 ||
			(y.w[1]&SNAN_MASK64) == SNAN_MASK64 {
			pfpsf |= BID_INVALID_EXCEPTION
		}
		res = 0
		return res, pfpsf
	}
	// SIMPLE (CASE2)
	if x.w[0] == y.w[0] && x.w[1] == y.w[1] {
		res = 1
		return res, pfpsf
	}
	// INFINITY (CASE3)
	if (x.w[1] & INFINITY_MASK64) == INFINITY_MASK64 {
		if (x.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = 0
			if ((y.w[1] & INFINITY_MASK64) == INFINITY_MASK64) &&
				(y.w[1]&MASK_SIGN64) == MASK_SIGN64 {
				res = 1
			}
			return res, pfpsf
		} else {
			res = 1
			return res, pfpsf
		}
	} else if (y.w[1] & INFINITY_MASK64) == INFINITY_MASK64 {
		res = 0
		if (y.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// CONVERT X
	sig_x.w[1] = x.w[1] & 0x0001ffffffffffff
	sig_x.w[0] = x.w[0]
	exp_x = int((x.w[1] >> 49) & 0x000000000003fff)

	if (sig_x.w[1] > 0x0001ed09bead87c0) ||
		((sig_x.w[1] == 0x0001ed09bead87c0) &&
			(sig_x.w[0] > 0x378d8e63ffffffff)) ||
		((x.w[1] & 0x6000000000000000) == 0x6000000000000000) {
		non_canon_x = 1
	} else {
		non_canon_x = 0
	}

	exp_y = int((y.w[1] >> 49) & 0x0000000000003fff)
	sig_y.w[1] = y.w[1] & 0x0001ffffffffffff
	sig_y.w[0] = y.w[0]

	if (sig_y.w[1] > 0x0001ed09bead87c0) ||
		((sig_y.w[1] == 0x0001ed09bead87c0) &&
			(sig_y.w[0] > 0x378d8e63ffffffff)) ||
		((y.w[1] & 0x6000000000000000) == 0x6000000000000000) {
		non_canon_y = 1
	} else {
		non_canon_y = 0
	}

	if non_canon_x != 0 || ((sig_x.w[1] == 0) && (sig_x.w[0] == 0)) {
		x_is_zero = 1
	}
	if non_canon_y != 0 || ((sig_y.w[1] == 0) && (sig_y.w[0] == 0)) {
		y_is_zero = 1
	}
	if x_is_zero != 0 && y_is_zero != 0 {
		res = 1
		return res, pfpsf
	} else if x_is_zero != 0 {
		res = 0
		if (y.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	} else if y_is_zero != 0 {
		res = 0
		if (x.w[1] & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	if ((x.w[1] ^ y.w[1]) & MASK_SIGN64) == MASK_SIGN64 {
		res = 0
		if (y.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	if exp_y == exp_x {
		res = 0
		if ((sig_x.w[1] > sig_y.w[1]) ||
			(sig_x.w[1] == sig_y.w[1] && sig_x.w[0] >= sig_y.w[0])) !=
			((x.w[1] & MASK_SIGN64) == MASK_SIGN64) {
			res = 1
		}
		return res, pfpsf
	}
	if sig_x.w[1] >= sig_y.w[1] && sig_x.w[0] >= sig_y.w[0] &&
		exp_x > exp_y {
		res = 0
		if (x.w[1] & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	if sig_x.w[1] <= sig_y.w[1] && sig_x.w[0] <= sig_y.w[0] &&
		exp_x < exp_y {
		res = 0
		if (x.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}

	diff = exp_x - exp_y

	if diff > 0 {
		if diff > 33 {
			res = 0
			if (x.w[1] & MASK_SIGN64) != MASK_SIGN64 {
				res = 1
			}
			return res, pfpsf
		}
		if diff > 19 {
			sig_n_prime256 = __mul_128x128_to_256(sig_x, bid_ten2k128[diff-20])
			if sig_n_prime256.w[3] == 0 && (sig_n_prime256.w[2] == 0) &&
				sig_n_prime256.w[1] == sig_y.w[1] &&
				(sig_n_prime256.w[0] == sig_y.w[0]) {
				res = 1
				return res, pfpsf
			}
			res = 0
			if (((sig_n_prime256.w[3] > 0) || sig_n_prime256.w[2] > 0) ||
				(sig_n_prime256.w[1] > sig_y.w[1]) ||
				(sig_n_prime256.w[1] == sig_y.w[1] &&
					sig_n_prime256.w[0] > sig_y.w[0])) !=
				((y.w[1] & MASK_SIGN64) == MASK_SIGN64) {
				res = 1
			}
			return res, pfpsf
		}
		sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[diff], sig_x)
		if (sig_n_prime192.w[2] == 0) && sig_n_prime192.w[1] == sig_y.w[1] &&
			(sig_n_prime192.w[0] == sig_y.w[0]) {
			res = 1
			return res, pfpsf
		}
		res = 0
		if ((sig_n_prime192.w[2] > 0) ||
			(sig_n_prime192.w[1] > sig_y.w[1]) ||
			(sig_n_prime192.w[1] == sig_y.w[1] &&
				sig_n_prime192.w[0] > sig_y.w[0])) !=
			((y.w[1] & MASK_SIGN64) == MASK_SIGN64) {
			res = 1
		}
		return res, pfpsf
	}

	diff = exp_y - exp_x

	if diff > 33 {
		res = 0
		if (x.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}

	if diff > 19 {
		sig_n_prime256 = __mul_128x128_to_256(sig_y, bid_ten2k128[diff-20])
		if sig_n_prime256.w[3] == 0 && (sig_n_prime256.w[2] == 0) &&
			sig_n_prime256.w[1] == sig_x.w[1] &&
			(sig_n_prime256.w[0] == sig_x.w[0]) {
			res = 1
			return res, pfpsf
		}
		res = 0
		if (sig_n_prime256.w[3] == 0 && sig_n_prime256.w[2] == 0 &&
			(sig_n_prime256.w[1] < sig_x.w[1] ||
				(sig_n_prime256.w[1] == sig_x.w[1] &&
					sig_n_prime256.w[0] < sig_x.w[0]))) !=
			((x.w[1] & MASK_SIGN64) == MASK_SIGN64) {
			res = 1
		}
		return res, pfpsf
	}
	sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[diff], sig_y)
	if (sig_n_prime192.w[2] == 0) && sig_n_prime192.w[1] == sig_x.w[1] &&
		(sig_n_prime192.w[0] == sig_x.w[0]) {
		res = 1
		return res, pfpsf
	}
	res = 0
	if (sig_n_prime192.w[2] == 0 &&
		(sig_n_prime192.w[1] < sig_x.w[1] ||
			(sig_n_prime192.w[1] == sig_x.w[1] &&
				sig_n_prime192.w[0] < sig_x.w[0]))) !=
		((y.w[1] & MASK_SIGN64) == MASK_SIGN64) {
		res = 1
	}
	return res, pfpsf
}

// Bid128QuietGreaterUnordered - Intel bid128_quiet_greater_unordered 기계적 포팅
func Bid128QuietGreaterUnordered(x, y BID_UINT128) (int, uint32) {
	var res int
	var exp_x, exp_y int
	var diff int
	var sig_x, sig_y BID_UINT128
	var sig_n_prime192 BID_UINT192
	var sig_n_prime256 BID_UINT256
	var pfpsf uint32
	var x_is_zero, y_is_zero, non_canon_x, non_canon_y int

	// NaN (CASE1)
	if ((x.w[1] & NAN_MASK64) == NAN_MASK64) ||
		((y.w[1] & NAN_MASK64) == NAN_MASK64) {
		if (x.w[1]&SNAN_MASK64) == SNAN_MASK64 ||
			(y.w[1]&SNAN_MASK64) == SNAN_MASK64 {
			pfpsf |= BID_INVALID_EXCEPTION
		}
		res = 1
		return res, pfpsf
	}
	// SIMPLE (CASE2)
	if x.w[0] == y.w[0] && x.w[1] == y.w[1] {
		res = 0
		return res, pfpsf
	}
	// INFINITY (CASE3)
	if (x.w[1] & INFINITY_MASK64) == INFINITY_MASK64 {
		if (x.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = 0
			return res, pfpsf
		} else {
			res = 0
			if ((y.w[1] & INFINITY_MASK64) != INFINITY_MASK64) ||
				((y.w[1] & MASK_SIGN64) == MASK_SIGN64) {
				res = 1
			}
			return res, pfpsf
		}
	} else if (y.w[1] & INFINITY_MASK64) == INFINITY_MASK64 {
		res = 0
		if (y.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// CONVERT X
	sig_x.w[1] = x.w[1] & 0x0001ffffffffffff
	sig_x.w[0] = x.w[0]
	exp_x = int((x.w[1] >> 49) & 0x000000000003fff)

	if (sig_x.w[1] > 0x0001ed09bead87c0) ||
		((sig_x.w[1] == 0x0001ed09bead87c0) &&
			(sig_x.w[0] > 0x378d8e63ffffffff)) ||
		((x.w[1] & 0x6000000000000000) == 0x6000000000000000) {
		non_canon_x = 1
	} else {
		non_canon_x = 0
	}

	exp_y = int((y.w[1] >> 49) & 0x0000000000003fff)
	sig_y.w[1] = y.w[1] & 0x0001ffffffffffff
	sig_y.w[0] = y.w[0]

	if (sig_y.w[1] > 0x0001ed09bead87c0) ||
		((sig_y.w[1] == 0x0001ed09bead87c0) &&
			(sig_y.w[0] > 0x378d8e63ffffffff)) ||
		((y.w[1] & 0x6000000000000000) == 0x6000000000000000) {
		non_canon_y = 1
	} else {
		non_canon_y = 0
	}

	if non_canon_x != 0 || ((sig_x.w[1] == 0) && (sig_x.w[0] == 0)) {
		x_is_zero = 1
	}
	if non_canon_y != 0 || ((sig_y.w[1] == 0) && (sig_y.w[0] == 0)) {
		y_is_zero = 1
	}
	if x_is_zero != 0 && y_is_zero != 0 {
		res = 0
		return res, pfpsf
	} else if x_is_zero != 0 {
		res = 0
		if (y.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	} else if y_is_zero != 0 {
		res = 0
		if (x.w[1] & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	if ((x.w[1] ^ y.w[1]) & MASK_SIGN64) == MASK_SIGN64 {
		res = 0
		if (y.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// REDUNDANT REPRESENTATIONS (CASE6)
	// This is the same as quiet_greater from CASE6 onward
	if exp_y == exp_x {
		res = 0
		if ((sig_x.w[1] > sig_y.w[1]) ||
			(sig_x.w[1] == sig_y.w[1] && sig_x.w[0] >= sig_y.w[0])) !=
			((x.w[1] & MASK_SIGN64) == MASK_SIGN64) {
			res = 1
		}
		return res, pfpsf
	}
	if sig_x.w[1] >= sig_y.w[1] && sig_x.w[0] >= sig_y.w[0] &&
		exp_x > exp_y {
		res = 0
		if (x.w[1] & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	if sig_x.w[1] <= sig_y.w[1] && sig_x.w[0] <= sig_y.w[0] &&
		exp_x < exp_y {
		res = 0
		if (x.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}

	diff = exp_x - exp_y

	if diff > 0 {
		if diff > 33 {
			res = 0
			if (x.w[1] & MASK_SIGN64) != MASK_SIGN64 {
				res = 1
			}
			return res, pfpsf
		}
		if diff > 19 {
			sig_n_prime256 = __mul_128x128_to_256(sig_x, bid_ten2k128[diff-20])
			if sig_n_prime256.w[3] == 0 && (sig_n_prime256.w[2] == 0) &&
				sig_n_prime256.w[1] == sig_y.w[1] &&
				(sig_n_prime256.w[0] == sig_y.w[0]) {
				res = 0
				return res, pfpsf
			}
			res = 0
			if (((sig_n_prime256.w[3] > 0) || sig_n_prime256.w[2] > 0) ||
				(sig_n_prime256.w[1] > sig_y.w[1]) ||
				(sig_n_prime256.w[1] == sig_y.w[1] &&
					sig_n_prime256.w[0] > sig_y.w[0])) !=
				((y.w[1] & MASK_SIGN64) == MASK_SIGN64) {
				res = 1
			}
			return res, pfpsf
		}
		sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[diff], sig_x)
		if (sig_n_prime192.w[2] == 0) && sig_n_prime192.w[1] == sig_y.w[1] &&
			(sig_n_prime192.w[0] == sig_y.w[0]) {
			res = 0
			return res, pfpsf
		}
		res = 0
		if ((sig_n_prime192.w[2] > 0) ||
			(sig_n_prime192.w[1] > sig_y.w[1]) ||
			(sig_n_prime192.w[1] == sig_y.w[1] &&
				sig_n_prime192.w[0] > sig_y.w[0])) !=
			((y.w[1] & MASK_SIGN64) == MASK_SIGN64) {
			res = 1
		}
		return res, pfpsf
	}

	diff = exp_y - exp_x

	if diff > 33 {
		res = 0
		if (x.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}

	if diff > 19 {
		sig_n_prime256 = __mul_128x128_to_256(sig_y, bid_ten2k128[diff-20])
		if sig_n_prime256.w[3] == 0 && (sig_n_prime256.w[2] == 0) &&
			sig_n_prime256.w[1] == sig_x.w[1] &&
			(sig_n_prime256.w[0] == sig_x.w[0]) {
			res = 0
			return res, pfpsf
		}
		res = 0
		if (sig_n_prime256.w[3] == 0 && sig_n_prime256.w[2] == 0 &&
			(sig_n_prime256.w[1] < sig_x.w[1] ||
				(sig_n_prime256.w[1] == sig_x.w[1] &&
					sig_n_prime256.w[0] < sig_x.w[0]))) !=
			((x.w[1] & MASK_SIGN64) == MASK_SIGN64) {
			res = 1
		}
		return res, pfpsf
	}
	sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[diff], sig_y)
	if (sig_n_prime192.w[2] == 0) && sig_n_prime192.w[1] == sig_x.w[1] &&
		(sig_n_prime192.w[0] == sig_x.w[0]) {
		res = 0
		return res, pfpsf
	}
	res = 0
	if (sig_n_prime192.w[2] == 0 &&
		(sig_n_prime192.w[1] < sig_x.w[1] ||
			(sig_n_prime192.w[1] == sig_x.w[1] &&
				sig_n_prime192.w[0] < sig_x.w[0]))) !=
		((y.w[1] & MASK_SIGN64) == MASK_SIGN64) {
		res = 1
	}
	return res, pfpsf
}

// Bid128QuietLess - Intel bid128_quiet_less 기계적 포팅
func Bid128QuietLess(x, y BID_UINT128) (int, uint32) {
	var res int
	var exp_x, exp_y int
	var diff int
	var sig_x, sig_y BID_UINT128
	var sig_n_prime192 BID_UINT192
	var sig_n_prime256 BID_UINT256
	var pfpsf uint32
	var x_is_zero, y_is_zero, non_canon_x, non_canon_y int

	// NaN (CASE1)
	if ((x.w[1] & NAN_MASK64) == NAN_MASK64) ||
		((y.w[1] & NAN_MASK64) == NAN_MASK64) {
		if (x.w[1]&SNAN_MASK64) == SNAN_MASK64 ||
			(y.w[1]&SNAN_MASK64) == SNAN_MASK64 {
			pfpsf |= BID_INVALID_EXCEPTION
		}
		res = 0
		return res, pfpsf
	}
	// SIMPLE (CASE2)
	if x.w[0] == y.w[0] && x.w[1] == y.w[1] {
		res = 0
		return res, pfpsf
	}
	// INFINITY (CASE3)
	if (x.w[1] & INFINITY_MASK64) == INFINITY_MASK64 {
		if (x.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = 0
			if ((y.w[1] & INFINITY_MASK64) != INFINITY_MASK64) ||
				(y.w[1]&MASK_SIGN64) != MASK_SIGN64 {
				res = 1
			}
			return res, pfpsf
		} else {
			res = 0
			return res, pfpsf
		}
	} else if (y.w[1] & INFINITY_MASK64) == INFINITY_MASK64 {
		res = 0
		if (y.w[1] & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// CONVERT X
	sig_x.w[1] = x.w[1] & 0x0001ffffffffffff
	sig_x.w[0] = x.w[0]
	exp_x = int((x.w[1] >> 49) & 0x000000000003fff)

	if (sig_x.w[1] > 0x0001ed09bead87c0) ||
		((sig_x.w[1] == 0x0001ed09bead87c0) &&
			(sig_x.w[0] > 0x378d8e63ffffffff)) ||
		((x.w[1] & 0x6000000000000000) == 0x6000000000000000) {
		non_canon_x = 1
	} else {
		non_canon_x = 0
	}

	exp_y = int((y.w[1] >> 49) & 0x0000000000003fff)
	sig_y.w[1] = y.w[1] & 0x0001ffffffffffff
	sig_y.w[0] = y.w[0]

	if (sig_y.w[1] > 0x0001ed09bead87c0) ||
		((sig_y.w[1] == 0x0001ed09bead87c0) &&
			(sig_y.w[0] > 0x378d8e63ffffffff)) ||
		((y.w[1] & 0x6000000000000000) == 0x6000000000000000) {
		non_canon_y = 1
	} else {
		non_canon_y = 0
	}

	if non_canon_x != 0 || ((sig_x.w[1] == 0) && (sig_x.w[0] == 0)) {
		x_is_zero = 1
	}
	if non_canon_y != 0 || ((sig_y.w[1] == 0) && (sig_y.w[0] == 0)) {
		y_is_zero = 1
	}
	if x_is_zero != 0 && y_is_zero != 0 {
		res = 0
		return res, pfpsf
	} else if x_is_zero != 0 {
		res = 0
		if (y.w[1] & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	} else if y_is_zero != 0 {
		res = 0
		if (x.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	if ((x.w[1] ^ y.w[1]) & MASK_SIGN64) == MASK_SIGN64 {
		res = 0
		if (y.w[1] & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	if exp_y == exp_x {
		res = 0
		if ((sig_x.w[1] > sig_y.w[1]) ||
			(sig_x.w[1] == sig_y.w[1] && sig_x.w[0] >= sig_y.w[0])) !=
			((x.w[1] & MASK_SIGN64) != MASK_SIGN64) {
			res = 1
		}
		return res, pfpsf
	}
	if (sig_x.w[1] > sig_y.w[1] ||
		(sig_x.w[1] == sig_y.w[1] && sig_x.w[0] > sig_y.w[0])) &&
		exp_x >= exp_y {
		res = 0
		if (x.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	if (sig_x.w[1] < sig_y.w[1] ||
		(sig_x.w[1] == sig_y.w[1] && sig_x.w[0] < sig_y.w[0])) &&
		exp_x <= exp_y {
		res = 0
		if (x.w[1] & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}

	diff = exp_x - exp_y

	if diff > 0 {
		if diff > 33 {
			res = 0
			if (x.w[1] & MASK_SIGN64) == MASK_SIGN64 {
				res = 1
			}
			return res, pfpsf
		}
		if diff > 19 {
			sig_n_prime256 = __mul_128x128_to_256(sig_x, bid_ten2k128[diff-20])
			if sig_n_prime256.w[3] == 0 && (sig_n_prime256.w[2] == 0) &&
				sig_n_prime256.w[1] == sig_y.w[1] &&
				(sig_n_prime256.w[0] == sig_y.w[0]) {
				res = 0
				return res, pfpsf
			}
			res = 0
			if (((sig_n_prime256.w[3] > 0) || sig_n_prime256.w[2] > 0) ||
				(sig_n_prime256.w[1] > sig_y.w[1]) ||
				(sig_n_prime256.w[1] == sig_y.w[1] &&
					sig_n_prime256.w[0] > sig_y.w[0])) !=
				((y.w[1] & MASK_SIGN64) != MASK_SIGN64) {
				res = 1
			}
			return res, pfpsf
		}
		sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[diff], sig_x)
		if (sig_n_prime192.w[2] == 0) && sig_n_prime192.w[1] == sig_y.w[1] &&
			(sig_n_prime192.w[0] == sig_y.w[0]) {
			res = 0
			return res, pfpsf
		}
		res = 0
		if ((sig_n_prime192.w[2] > 0) ||
			(sig_n_prime192.w[1] > sig_y.w[1]) ||
			(sig_n_prime192.w[1] == sig_y.w[1] &&
				sig_n_prime192.w[0] > sig_y.w[0])) !=
			((y.w[1] & MASK_SIGN64) != MASK_SIGN64) {
			res = 1
		}
		return res, pfpsf
	}

	diff = exp_y - exp_x

	if diff > 33 {
		res = 0
		if (x.w[1] & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}

	if diff > 19 {
		sig_n_prime256 = __mul_128x128_to_256(sig_y, bid_ten2k128[diff-20])
		if sig_n_prime256.w[3] == 0 && (sig_n_prime256.w[2] == 0) &&
			sig_n_prime256.w[1] == sig_x.w[1] &&
			(sig_n_prime256.w[0] == sig_x.w[0]) {
			res = 0
			return res, pfpsf
		}
		res = 0
		if (sig_n_prime256.w[3] != 0 || sig_n_prime256.w[2] != 0 ||
			(sig_n_prime256.w[1] > sig_x.w[1] ||
				(sig_n_prime256.w[1] == sig_x.w[1] &&
					sig_n_prime256.w[0] > sig_x.w[0]))) !=
			((x.w[1] & MASK_SIGN64) == MASK_SIGN64) {
			res = 1
		}
		return res, pfpsf
	}
	sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[diff], sig_y)
	if (sig_n_prime192.w[2] == 0) && sig_n_prime192.w[1] == sig_x.w[1] &&
		(sig_n_prime192.w[0] == sig_x.w[0]) {
		res = 0
		return res, pfpsf
	}
	res = 0
	if (sig_n_prime192.w[2] != 0 ||
		(sig_n_prime192.w[1] > sig_x.w[1] ||
			(sig_n_prime192.w[1] == sig_x.w[1] &&
				sig_n_prime192.w[0] > sig_x.w[0]))) !=
		((y.w[1] & MASK_SIGN64) == MASK_SIGN64) {
		res = 1
	}
	return res, pfpsf
}

// Bid128QuietLessEqual - Intel bid128_quiet_less_equal 기계적 포팅
func Bid128QuietLessEqual(x, y BID_UINT128) (int, uint32) {
	var res int
	var exp_x, exp_y int
	var diff int
	var sig_x, sig_y BID_UINT128
	var sig_n_prime192 BID_UINT192
	var sig_n_prime256 BID_UINT256
	var pfpsf uint32
	var x_is_zero, y_is_zero, non_canon_x, non_canon_y int

	// NaN (CASE1)
	if ((x.w[1] & NAN_MASK64) == NAN_MASK64) ||
		((y.w[1] & NAN_MASK64) == NAN_MASK64) {
		if (x.w[1]&SNAN_MASK64) == SNAN_MASK64 ||
			(y.w[1]&SNAN_MASK64) == SNAN_MASK64 {
			pfpsf |= BID_INVALID_EXCEPTION
		}
		res = 0
		return res, pfpsf
	}
	// SIMPLE (CASE2)
	if x.w[0] == y.w[0] && x.w[1] == y.w[1] {
		res = 1
		return res, pfpsf
	}
	// INFINITY (CASE3)
	if (x.w[1] & INFINITY_MASK64) == INFINITY_MASK64 {
		if (x.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
			return res, pfpsf
		} else {
			res = 0
			if ((y.w[1] & INFINITY_MASK64) == INFINITY_MASK64) &&
				((y.w[1] & MASK_SIGN64) != MASK_SIGN64) {
				res = 1
			}
			return res, pfpsf
		}
	} else if (y.w[1] & INFINITY_MASK64) == INFINITY_MASK64 {
		res = 0
		if (y.w[1] & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	// CONVERT X
	sig_x.w[1] = x.w[1] & 0x0001ffffffffffff
	sig_x.w[0] = x.w[0]
	exp_x = int((x.w[1] >> 49) & 0x000000000003fff)

	if (sig_x.w[1] > 0x0001ed09bead87c0) ||
		((sig_x.w[1] == 0x0001ed09bead87c0) &&
			(sig_x.w[0] > 0x378d8e63ffffffff)) ||
		((x.w[1] & 0x6000000000000000) == 0x6000000000000000) {
		non_canon_x = 1
	} else {
		non_canon_x = 0
	}

	exp_y = int((y.w[1] >> 49) & 0x0000000000003fff)
	sig_y.w[1] = y.w[1] & 0x0001ffffffffffff
	sig_y.w[0] = y.w[0]

	if (sig_y.w[1] > 0x0001ed09bead87c0) ||
		((sig_y.w[1] == 0x0001ed09bead87c0) &&
			(sig_y.w[0] > 0x378d8e63ffffffff)) ||
		((y.w[1] & 0x6000000000000000) == 0x6000000000000000) {
		non_canon_y = 1
	} else {
		non_canon_y = 0
	}

	if non_canon_x != 0 || ((sig_x.w[1] == 0) && (sig_x.w[0] == 0)) {
		x_is_zero = 1
	}
	if non_canon_y != 0 || ((sig_y.w[1] == 0) && (sig_y.w[0] == 0)) {
		y_is_zero = 1
	}
	if x_is_zero != 0 && y_is_zero != 0 {
		res = 1
		return res, pfpsf
	} else if x_is_zero != 0 {
		res = 0
		if (y.w[1] & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	} else if y_is_zero != 0 {
		res = 0
		if (x.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	if ((x.w[1] ^ y.w[1]) & MASK_SIGN64) == MASK_SIGN64 {
		res = 0
		if (y.w[1] & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	if exp_y == exp_x {
		res = 0
		if ((sig_x.w[1] > sig_y.w[1]) ||
			(sig_x.w[1] == sig_y.w[1] && sig_x.w[0] >= sig_y.w[0])) !=
			((x.w[1] & MASK_SIGN64) != MASK_SIGN64) {
			res = 1
		}
		return res, pfpsf
	}
	if (sig_x.w[1] > sig_y.w[1] ||
		(sig_x.w[1] == sig_y.w[1] && sig_x.w[0] > sig_y.w[0])) &&
		exp_x >= exp_y {
		res = 0
		if (x.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	if (sig_x.w[1] < sig_y.w[1] ||
		(sig_x.w[1] == sig_y.w[1] && sig_x.w[0] < sig_y.w[0])) &&
		exp_x <= exp_y {
		res = 0
		if (x.w[1] & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}

	diff = exp_x - exp_y

	if diff > 0 {
		if diff > 33 {
			res = 0
			if (x.w[1] & MASK_SIGN64) == MASK_SIGN64 {
				res = 1
			}
			return res, pfpsf
		}
		if diff > 19 {
			sig_n_prime256 = __mul_128x128_to_256(sig_x, bid_ten2k128[diff-20])
			if sig_n_prime256.w[3] == 0 && (sig_n_prime256.w[2] == 0) &&
				sig_n_prime256.w[1] == sig_y.w[1] &&
				(sig_n_prime256.w[0] == sig_y.w[0]) {
				res = 1
				return res, pfpsf
			}
			res = 0
			if (((sig_n_prime256.w[3] > 0) || sig_n_prime256.w[2] > 0) ||
				(sig_n_prime256.w[1] > sig_y.w[1]) ||
				(sig_n_prime256.w[1] == sig_y.w[1] &&
					sig_n_prime256.w[0] > sig_y.w[0])) !=
				((y.w[1] & MASK_SIGN64) != MASK_SIGN64) {
				res = 1
			}
			return res, pfpsf
		}
		sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[diff], sig_x)
		if (sig_n_prime192.w[2] == 0) && sig_n_prime192.w[1] == sig_y.w[1] &&
			(sig_n_prime192.w[0] == sig_y.w[0]) {
			res = 1
			return res, pfpsf
		}
		res = 0
		if ((sig_n_prime192.w[2] > 0) ||
			(sig_n_prime192.w[1] > sig_y.w[1]) ||
			(sig_n_prime192.w[1] == sig_y.w[1] &&
				sig_n_prime192.w[0] > sig_y.w[0])) !=
			((y.w[1] & MASK_SIGN64) != MASK_SIGN64) {
			res = 1
		}
		return res, pfpsf
	}

	diff = exp_y - exp_x

	if diff > 33 {
		res = 0
		if (x.w[1] & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}

	if diff > 19 {
		sig_n_prime256 = __mul_128x128_to_256(sig_y, bid_ten2k128[diff-20])
		if sig_n_prime256.w[3] == 0 && (sig_n_prime256.w[2] == 0) &&
			sig_n_prime256.w[1] == sig_x.w[1] &&
			(sig_n_prime256.w[0] == sig_x.w[0]) {
			res = 1
			return res, pfpsf
		}
		res = 0
		if (sig_n_prime256.w[3] != 0 || sig_n_prime256.w[2] != 0 ||
			(sig_n_prime256.w[1] > sig_x.w[1] ||
				(sig_n_prime256.w[1] == sig_x.w[1] &&
					sig_n_prime256.w[0] > sig_x.w[0]))) !=
			((x.w[1] & MASK_SIGN64) == MASK_SIGN64) {
			res = 1
		}
		return res, pfpsf
	}
	sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[diff], sig_y)
	if (sig_n_prime192.w[2] == 0) && sig_n_prime192.w[1] == sig_x.w[1] &&
		(sig_n_prime192.w[0] == sig_x.w[0]) {
		res = 1
		return res, pfpsf
	}
	res = 0
	if (sig_n_prime192.w[2] != 0 ||
		(sig_n_prime192.w[1] > sig_x.w[1] ||
			(sig_n_prime192.w[1] == sig_x.w[1] &&
				sig_n_prime192.w[0] > sig_x.w[0]))) !=
		((y.w[1] & MASK_SIGN64) == MASK_SIGN64) {
		res = 1
	}
	return res, pfpsf
}

// Bid128QuietLessUnordered - Intel bid128_quiet_less_unordered 기계적 포팅
func Bid128QuietLessUnordered(x, y BID_UINT128) (int, uint32) {
	// Same as quiet_less but NaN returns 1
	var res int
	var exp_x, exp_y int
	var diff int
	var sig_x, sig_y BID_UINT128
	var sig_n_prime192 BID_UINT192
	var sig_n_prime256 BID_UINT256
	var pfpsf uint32
	var x_is_zero, y_is_zero, non_canon_x, non_canon_y int

	if ((x.w[1] & NAN_MASK64) == NAN_MASK64) ||
		((y.w[1] & NAN_MASK64) == NAN_MASK64) {
		if (x.w[1]&SNAN_MASK64) == SNAN_MASK64 ||
			(y.w[1]&SNAN_MASK64) == SNAN_MASK64 {
			pfpsf |= BID_INVALID_EXCEPTION
		}
		res = 1
		return res, pfpsf
	}
	if x.w[0] == y.w[0] && x.w[1] == y.w[1] {
		res = 0
		return res, pfpsf
	}
	if (x.w[1] & INFINITY_MASK64) == INFINITY_MASK64 {
		if (x.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = 0
			if ((y.w[1] & INFINITY_MASK64) != INFINITY_MASK64) ||
				(y.w[1]&MASK_SIGN64) != MASK_SIGN64 {
				res = 1
			}
			return res, pfpsf
		} else {
			res = 0
			return res, pfpsf
		}
	} else if (y.w[1] & INFINITY_MASK64) == INFINITY_MASK64 {
		res = 0
		if (y.w[1] & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	sig_x.w[1] = x.w[1] & 0x0001ffffffffffff
	sig_x.w[0] = x.w[0]
	exp_x = int((x.w[1] >> 49) & 0x000000000003fff)

	if (sig_x.w[1] > 0x0001ed09bead87c0) ||
		((sig_x.w[1] == 0x0001ed09bead87c0) &&
			(sig_x.w[0] > 0x378d8e63ffffffff)) ||
		((x.w[1] & 0x6000000000000000) == 0x6000000000000000) {
		non_canon_x = 1
	} else {
		non_canon_x = 0
	}

	exp_y = int((y.w[1] >> 49) & 0x0000000000003fff)
	sig_y.w[1] = y.w[1] & 0x0001ffffffffffff
	sig_y.w[0] = y.w[0]

	if (sig_y.w[1] > 0x0001ed09bead87c0) ||
		((sig_y.w[1] == 0x0001ed09bead87c0) &&
			(sig_y.w[0] > 0x378d8e63ffffffff)) ||
		((y.w[1] & 0x6000000000000000) == 0x6000000000000000) {
		non_canon_y = 1
	} else {
		non_canon_y = 0
	}

	if non_canon_x != 0 || ((sig_x.w[1] == 0) && (sig_x.w[0] == 0)) {
		x_is_zero = 1
	}
	if non_canon_y != 0 || ((sig_y.w[1] == 0) && (sig_y.w[0] == 0)) {
		y_is_zero = 1
	}
	if x_is_zero != 0 && y_is_zero != 0 {
		res = 0
		return res, pfpsf
	} else if x_is_zero != 0 {
		res = 0
		if (y.w[1] & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	} else if y_is_zero != 0 {
		res = 0
		if (x.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	if ((x.w[1] ^ y.w[1]) & MASK_SIGN64) == MASK_SIGN64 {
		res = 0
		if (y.w[1] & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	if exp_y == exp_x {
		res = 0
		if ((sig_x.w[1] > sig_y.w[1]) ||
			(sig_x.w[1] == sig_y.w[1] && sig_x.w[0] >= sig_y.w[0])) !=
			((x.w[1] & MASK_SIGN64) != MASK_SIGN64) {
			res = 1
		}
		return res, pfpsf
	}
	if (sig_x.w[1] > sig_y.w[1] ||
		(sig_x.w[1] == sig_y.w[1] && sig_x.w[0] > sig_y.w[0])) &&
		exp_x >= exp_y {
		res = 0
		if (x.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	if (sig_x.w[1] < sig_y.w[1] ||
		(sig_x.w[1] == sig_y.w[1] && sig_x.w[0] < sig_y.w[0])) &&
		exp_x <= exp_y {
		res = 0
		if (x.w[1] & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}

	diff = exp_x - exp_y
	if diff > 0 {
		if diff > 33 {
			res = 0
			if (x.w[1] & MASK_SIGN64) == MASK_SIGN64 {
				res = 1
			}
			return res, pfpsf
		}
		if diff > 19 {
			sig_n_prime256 = __mul_128x128_to_256(sig_x, bid_ten2k128[diff-20])
			if sig_n_prime256.w[3] == 0 && (sig_n_prime256.w[2] == 0) &&
				sig_n_prime256.w[1] == sig_y.w[1] &&
				(sig_n_prime256.w[0] == sig_y.w[0]) {
				res = 0
				return res, pfpsf
			}
			res = 0
			if (((sig_n_prime256.w[3] > 0) || sig_n_prime256.w[2] > 0) ||
				(sig_n_prime256.w[1] > sig_y.w[1]) ||
				(sig_n_prime256.w[1] == sig_y.w[1] &&
					sig_n_prime256.w[0] > sig_y.w[0])) !=
				((y.w[1] & MASK_SIGN64) != MASK_SIGN64) {
				res = 1
			}
			return res, pfpsf
		}
		sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[diff], sig_x)
		if (sig_n_prime192.w[2] == 0) && sig_n_prime192.w[1] == sig_y.w[1] &&
			(sig_n_prime192.w[0] == sig_y.w[0]) {
			res = 0
			return res, pfpsf
		}
		res = 0
		if ((sig_n_prime192.w[2] > 0) ||
			(sig_n_prime192.w[1] > sig_y.w[1]) ||
			(sig_n_prime192.w[1] == sig_y.w[1] &&
				sig_n_prime192.w[0] > sig_y.w[0])) !=
			((y.w[1] & MASK_SIGN64) != MASK_SIGN64) {
			res = 1
		}
		return res, pfpsf
	}

	diff = exp_y - exp_x
	if diff > 33 {
		res = 0
		if (x.w[1] & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	if diff > 19 {
		sig_n_prime256 = __mul_128x128_to_256(sig_y, bid_ten2k128[diff-20])
		if sig_n_prime256.w[3] == 0 && (sig_n_prime256.w[2] == 0) &&
			sig_n_prime256.w[1] == sig_x.w[1] &&
			(sig_n_prime256.w[0] == sig_x.w[0]) {
			res = 0
			return res, pfpsf
		}
		res = 0
		if (sig_n_prime256.w[3] != 0 || sig_n_prime256.w[2] != 0 ||
			(sig_n_prime256.w[1] > sig_x.w[1] ||
				(sig_n_prime256.w[1] == sig_x.w[1] &&
					sig_n_prime256.w[0] > sig_x.w[0]))) !=
			((x.w[1] & MASK_SIGN64) == MASK_SIGN64) {
			res = 1
		}
		return res, pfpsf
	}
	sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[diff], sig_y)
	if (sig_n_prime192.w[2] == 0) && sig_n_prime192.w[1] == sig_x.w[1] &&
		(sig_n_prime192.w[0] == sig_x.w[0]) {
		res = 0
		return res, pfpsf
	}
	res = 0
	if (sig_n_prime192.w[2] != 0 ||
		(sig_n_prime192.w[1] > sig_x.w[1] ||
			(sig_n_prime192.w[1] == sig_x.w[1] &&
				sig_n_prime192.w[0] > sig_x.w[0]))) !=
		((y.w[1] & MASK_SIGN64) == MASK_SIGN64) {
		res = 1
	}
	return res, pfpsf
}

// Bid128QuietNotEqual - Intel bid128_quiet_not_equal 기계적 포팅
func Bid128QuietNotEqual(x, y BID_UINT128) (int, uint32) {
	var res int
	var exp_x, exp_y, exp_t int
	var sig_x, sig_y, sig_t BID_UINT128
	var sig_n_prime192 BID_UINT192
	var sig_n_prime256 BID_UINT256
	var pfpsf uint32
	var x_is_zero, y_is_zero int
	var non_canon_x, non_canon_y int

	// NaN (CASE1)
	if ((x.w[1] & NAN_MASK64) == NAN_MASK64) ||
		((y.w[1] & NAN_MASK64) == NAN_MASK64) {
		if (x.w[1]&SNAN_MASK64) == SNAN_MASK64 ||
			(y.w[1]&SNAN_MASK64) == SNAN_MASK64 {
			pfpsf |= BID_INVALID_EXCEPTION
		}
		res = 1
		return res, pfpsf
	}
	// SIMPLE (CASE2)
	if x.w[0] == y.w[0] && x.w[1] == y.w[1] {
		res = 0
		return res, pfpsf
	}
	// INFINITY (CASE3)
	if (x.w[1] & INFINITY_MASK64) == INFINITY_MASK64 {
		if (y.w[1] & INFINITY_MASK64) == INFINITY_MASK64 {
			res = 0
			if ((x.w[1] ^ y.w[1]) & MASK_SIGN64) == MASK_SIGN64 {
				res = 1
			}
			return res, pfpsf
		} else {
			res = 1
			return res, pfpsf
		}
	}
	if (y.w[1] & INFINITY_MASK64) == INFINITY_MASK64 {
		res = 1
		return res, pfpsf
	}
	// CONVERT X
	sig_x.w[1] = x.w[1] & 0x0001ffffffffffff
	sig_x.w[0] = x.w[0]
	exp_x = int((x.w[1] >> 49) & 0x000000000003fff)

	if (sig_x.w[1] > 0x0001ed09bead87c0) ||
		((sig_x.w[1] == 0x0001ed09bead87c0) &&
			(sig_x.w[0] > 0x378d8e63ffffffff)) ||
		((x.w[1] & 0x6000000000000000) == 0x6000000000000000) {
		non_canon_x = 1
	} else {
		non_canon_x = 0
	}

	exp_y = int((y.w[1] >> 49) & 0x0000000000003fff)
	sig_y.w[1] = y.w[1] & 0x0001ffffffffffff
	sig_y.w[0] = y.w[0]

	if (sig_y.w[1] > 0x0001ed09bead87c0) ||
		((sig_y.w[1] == 0x0001ed09bead87c0) &&
			(sig_y.w[0] > 0x378d8e63ffffffff)) ||
		((y.w[1] & 0x6000000000000000) == 0x6000000000000000) {
		non_canon_y = 1
	} else {
		non_canon_y = 0
	}

	if non_canon_x != 0 || ((sig_x.w[1] == 0) && (sig_x.w[0] == 0)) {
		x_is_zero = 1
	}
	if non_canon_y != 0 || ((sig_y.w[1] == 0) && (sig_y.w[0] == 0)) {
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
	if (x.w[1]^y.w[1])&MASK_SIGN64 != 0 {
		res = 1
		return res, pfpsf
	}
	// REDUNDANT REPRESENTATIONS (CASE6)
	if exp_x > exp_y {
		exp_t = exp_x
		exp_x = exp_y
		exp_y = exp_t
		sig_t.w[1] = sig_x.w[1]
		sig_x.w[1] = sig_y.w[1]
		sig_y.w[1] = sig_t.w[1]
		sig_t.w[0] = sig_x.w[0]
		sig_x.w[0] = sig_y.w[0]
		sig_y.w[0] = sig_t.w[0]
	}

	if exp_y-exp_x > 33 {
		res = 1
		return res, pfpsf
	}

	if exp_y-exp_x > 19 {
		sig_n_prime256 = __mul_128x128_to_256(sig_y,
			bid_ten2k128[exp_y-exp_x-20])
		res = 0
		if (sig_n_prime256.w[3] != 0) || (sig_n_prime256.w[2] != 0) ||
			(sig_n_prime256.w[1] != sig_x.w[1]) ||
			(sig_n_prime256.w[0] != sig_x.w[0]) {
			res = 1
		}
		return res, pfpsf
	}
	sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[exp_y-exp_x], sig_y)
	res = 0
	if (sig_n_prime192.w[2] != 0) ||
		(sig_n_prime192.w[1] != sig_x.w[1]) ||
		(sig_n_prime192.w[0] != sig_x.w[0]) {
		res = 1
	}
	return res, pfpsf
}

// Bid128QuietNotGreater - Intel bid128_quiet_not_greater 기계적 포팅
// This is quiet_less_equal with NaN returning 1
func Bid128QuietNotGreater(x, y BID_UINT128) (int, uint32) {
	var res int
	var exp_x, exp_y int
	var diff int
	var sig_x, sig_y BID_UINT128
	var sig_n_prime192 BID_UINT192
	var sig_n_prime256 BID_UINT256
	var pfpsf uint32
	var x_is_zero, y_is_zero, non_canon_x, non_canon_y int

	if ((x.w[1] & NAN_MASK64) == NAN_MASK64) ||
		((y.w[1] & NAN_MASK64) == NAN_MASK64) {
		if (x.w[1]&SNAN_MASK64) == SNAN_MASK64 ||
			(y.w[1]&SNAN_MASK64) == SNAN_MASK64 {
			pfpsf |= BID_INVALID_EXCEPTION
		}
		res = 1
		return res, pfpsf
	}
	if x.w[0] == y.w[0] && x.w[1] == y.w[1] {
		res = 1
		return res, pfpsf
	}
	if (x.w[1] & INFINITY_MASK64) == INFINITY_MASK64 {
		if (x.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
			return res, pfpsf
		} else {
			res = 0
			if ((y.w[1] & INFINITY_MASK64) == INFINITY_MASK64) &&
				((y.w[1] & MASK_SIGN64) != MASK_SIGN64) {
				res = 1
			}
			return res, pfpsf
		}
	} else if (y.w[1] & INFINITY_MASK64) == INFINITY_MASK64 {
		res = 0
		if (y.w[1] & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	sig_x.w[1] = x.w[1] & 0x0001ffffffffffff
	sig_x.w[0] = x.w[0]
	exp_x = int((x.w[1] >> 49) & 0x000000000003fff)

	if (sig_x.w[1] > 0x0001ed09bead87c0) ||
		((sig_x.w[1] == 0x0001ed09bead87c0) &&
			(sig_x.w[0] > 0x378d8e63ffffffff)) ||
		((x.w[1] & 0x6000000000000000) == 0x6000000000000000) {
		non_canon_x = 1
	} else {
		non_canon_x = 0
	}

	exp_y = int((y.w[1] >> 49) & 0x0000000000003fff)
	sig_y.w[1] = y.w[1] & 0x0001ffffffffffff
	sig_y.w[0] = y.w[0]

	if (sig_y.w[1] > 0x0001ed09bead87c0) ||
		((sig_y.w[1] == 0x0001ed09bead87c0) &&
			(sig_y.w[0] > 0x378d8e63ffffffff)) ||
		((y.w[1] & 0x6000000000000000) == 0x6000000000000000) {
		non_canon_y = 1
	} else {
		non_canon_y = 0
	}

	if non_canon_x != 0 || ((sig_x.w[1] == 0) && (sig_x.w[0] == 0)) {
		x_is_zero = 1
	}
	if non_canon_y != 0 || ((sig_y.w[1] == 0) && (sig_y.w[0] == 0)) {
		y_is_zero = 1
	}
	if x_is_zero != 0 && y_is_zero != 0 {
		res = 1
		return res, pfpsf
	} else if x_is_zero != 0 {
		res = 0
		if (y.w[1] & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	} else if y_is_zero != 0 {
		res = 0
		if (x.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	if ((x.w[1] ^ y.w[1]) & MASK_SIGN64) == MASK_SIGN64 {
		res = 0
		if (y.w[1] & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	if exp_y == exp_x {
		res = 0
		if ((sig_x.w[1] > sig_y.w[1]) ||
			(sig_x.w[1] == sig_y.w[1] && sig_x.w[0] >= sig_y.w[0])) !=
			((x.w[1] & MASK_SIGN64) != MASK_SIGN64) {
			res = 1
		}
		return res, pfpsf
	}
	if (sig_x.w[1] > sig_y.w[1] ||
		(sig_x.w[1] == sig_y.w[1] && sig_x.w[0] > sig_y.w[0])) &&
		exp_x >= exp_y {
		res = 0
		if (x.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	if (sig_x.w[1] < sig_y.w[1] ||
		(sig_x.w[1] == sig_y.w[1] && sig_x.w[0] < sig_y.w[0])) &&
		exp_x <= exp_y {
		res = 0
		if (x.w[1] & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}

	diff = exp_x - exp_y
	if diff > 0 {
		if diff > 33 {
			res = 0
			if (x.w[1] & MASK_SIGN64) == MASK_SIGN64 {
				res = 1
			}
			return res, pfpsf
		}
		if diff > 19 {
			sig_n_prime256 = __mul_128x128_to_256(sig_x, bid_ten2k128[diff-20])
			if sig_n_prime256.w[3] == 0 && (sig_n_prime256.w[2] == 0) &&
				sig_n_prime256.w[1] == sig_y.w[1] &&
				(sig_n_prime256.w[0] == sig_y.w[0]) {
				res = 1
				return res, pfpsf
			}
			res = 0
			if (((sig_n_prime256.w[3] > 0) || sig_n_prime256.w[2] > 0) ||
				(sig_n_prime256.w[1] > sig_y.w[1]) ||
				(sig_n_prime256.w[1] == sig_y.w[1] &&
					sig_n_prime256.w[0] > sig_y.w[0])) !=
				((y.w[1] & MASK_SIGN64) != MASK_SIGN64) {
				res = 1
			}
			return res, pfpsf
		}
		sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[diff], sig_x)
		if (sig_n_prime192.w[2] == 0) && sig_n_prime192.w[1] == sig_y.w[1] &&
			(sig_n_prime192.w[0] == sig_y.w[0]) {
			res = 1
			return res, pfpsf
		}
		res = 0
		if ((sig_n_prime192.w[2] > 0) ||
			(sig_n_prime192.w[1] > sig_y.w[1]) ||
			(sig_n_prime192.w[1] == sig_y.w[1] &&
				sig_n_prime192.w[0] > sig_y.w[0])) !=
			((y.w[1] & MASK_SIGN64) != MASK_SIGN64) {
			res = 1
		}
		return res, pfpsf
	}

	diff = exp_y - exp_x
	if diff > 33 {
		res = 0
		if (x.w[1] & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	if diff > 19 {
		sig_n_prime256 = __mul_128x128_to_256(sig_y, bid_ten2k128[diff-20])
		if sig_n_prime256.w[3] == 0 && (sig_n_prime256.w[2] == 0) &&
			sig_n_prime256.w[1] == sig_x.w[1] &&
			(sig_n_prime256.w[0] == sig_x.w[0]) {
			res = 1
			return res, pfpsf
		}
		res = 0
		if (sig_n_prime256.w[3] != 0 || sig_n_prime256.w[2] != 0 ||
			(sig_n_prime256.w[1] > sig_x.w[1] ||
				(sig_n_prime256.w[1] == sig_x.w[1] &&
					sig_n_prime256.w[0] > sig_x.w[0]))) !=
			((x.w[1] & MASK_SIGN64) == MASK_SIGN64) {
			res = 1
		}
		return res, pfpsf
	}
	sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[diff], sig_y)
	if (sig_n_prime192.w[2] == 0) && sig_n_prime192.w[1] == sig_x.w[1] &&
		(sig_n_prime192.w[0] == sig_x.w[0]) {
		res = 1
		return res, pfpsf
	}
	res = 0
	if (sig_n_prime192.w[2] != 0 ||
		(sig_n_prime192.w[1] > sig_x.w[1] ||
			(sig_n_prime192.w[1] == sig_x.w[1] &&
				sig_n_prime192.w[0] > sig_x.w[0]))) !=
		((y.w[1] & MASK_SIGN64) == MASK_SIGN64) {
		res = 1
	}
	return res, pfpsf
}

// Bid128QuietNotLess - Intel bid128_quiet_not_less 기계적 포팅
// This is quiet_greater_equal with NaN returning 1
func Bid128QuietNotLess(x, y BID_UINT128) (int, uint32) {
	var res int
	var exp_x, exp_y int
	var diff int
	var sig_x, sig_y BID_UINT128
	var sig_n_prime192 BID_UINT192
	var sig_n_prime256 BID_UINT256
	var pfpsf uint32
	var x_is_zero, y_is_zero, non_canon_x, non_canon_y int

	if ((x.w[1] & NAN_MASK64) == NAN_MASK64) ||
		((y.w[1] & NAN_MASK64) == NAN_MASK64) {
		if (x.w[1]&SNAN_MASK64) == SNAN_MASK64 ||
			(y.w[1]&SNAN_MASK64) == SNAN_MASK64 {
			pfpsf |= BID_INVALID_EXCEPTION
		}
		res = 1
		return res, pfpsf
	}
	if x.w[0] == y.w[0] && x.w[1] == y.w[1] {
		res = 1
		return res, pfpsf
	}
	if (x.w[1] & INFINITY_MASK64) == INFINITY_MASK64 {
		if (x.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = 0
			if ((y.w[1] & INFINITY_MASK64) == INFINITY_MASK64) &&
				(y.w[1]&MASK_SIGN64) == MASK_SIGN64 {
				res = 1
			}
			return res, pfpsf
		} else {
			res = 1
			return res, pfpsf
		}
	} else if (y.w[1] & INFINITY_MASK64) == INFINITY_MASK64 {
		res = 0
		if (y.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	sig_x.w[1] = x.w[1] & 0x0001ffffffffffff
	sig_x.w[0] = x.w[0]
	exp_x = int((x.w[1] >> 49) & 0x000000000003fff)

	if (sig_x.w[1] > 0x0001ed09bead87c0) ||
		((sig_x.w[1] == 0x0001ed09bead87c0) &&
			(sig_x.w[0] > 0x378d8e63ffffffff)) ||
		((x.w[1] & 0x6000000000000000) == 0x6000000000000000) {
		non_canon_x = 1
	} else {
		non_canon_x = 0
	}

	exp_y = int((y.w[1] >> 49) & 0x0000000000003fff)
	sig_y.w[1] = y.w[1] & 0x0001ffffffffffff
	sig_y.w[0] = y.w[0]

	if (sig_y.w[1] > 0x0001ed09bead87c0) ||
		((sig_y.w[1] == 0x0001ed09bead87c0) &&
			(sig_y.w[0] > 0x378d8e63ffffffff)) ||
		((y.w[1] & 0x6000000000000000) == 0x6000000000000000) {
		non_canon_y = 1
	} else {
		non_canon_y = 0
	}

	if non_canon_x != 0 || ((sig_x.w[1] == 0) && (sig_x.w[0] == 0)) {
		x_is_zero = 1
	}
	if non_canon_y != 0 || ((sig_y.w[1] == 0) && (sig_y.w[0] == 0)) {
		y_is_zero = 1
	}
	if x_is_zero != 0 && y_is_zero != 0 {
		res = 1
		return res, pfpsf
	} else if x_is_zero != 0 {
		res = 0
		if (y.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	} else if y_is_zero != 0 {
		res = 0
		if (x.w[1] & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	if ((x.w[1] ^ y.w[1]) & MASK_SIGN64) == MASK_SIGN64 {
		res = 0
		if (y.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	if exp_y == exp_x {
		res = 0
		if ((sig_x.w[1] > sig_y.w[1]) ||
			(sig_x.w[1] == sig_y.w[1] && sig_x.w[0] >= sig_y.w[0])) !=
			((x.w[1] & MASK_SIGN64) == MASK_SIGN64) {
			res = 1
		}
		return res, pfpsf
	}
	if sig_x.w[1] >= sig_y.w[1] && sig_x.w[0] >= sig_y.w[0] &&
		exp_x > exp_y {
		res = 0
		if (x.w[1] & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	if sig_x.w[1] <= sig_y.w[1] && sig_x.w[0] <= sig_y.w[0] &&
		exp_x < exp_y {
		res = 0
		if (x.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}

	diff = exp_x - exp_y
	if diff > 0 {
		if diff > 33 {
			res = 0
			if (x.w[1] & MASK_SIGN64) != MASK_SIGN64 {
				res = 1
			}
			return res, pfpsf
		}
		if diff > 19 {
			sig_n_prime256 = __mul_128x128_to_256(sig_x, bid_ten2k128[diff-20])
			if sig_n_prime256.w[3] == 0 && (sig_n_prime256.w[2] == 0) &&
				sig_n_prime256.w[1] == sig_y.w[1] &&
				(sig_n_prime256.w[0] == sig_y.w[0]) {
				res = 1
				return res, pfpsf
			}
			res = 0
			if (((sig_n_prime256.w[3] > 0) || sig_n_prime256.w[2] > 0) ||
				(sig_n_prime256.w[1] > sig_y.w[1]) ||
				(sig_n_prime256.w[1] == sig_y.w[1] &&
					sig_n_prime256.w[0] > sig_y.w[0])) !=
				((y.w[1] & MASK_SIGN64) == MASK_SIGN64) {
				res = 1
			}
			return res, pfpsf
		}
		sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[diff], sig_x)
		if (sig_n_prime192.w[2] == 0) && sig_n_prime192.w[1] == sig_y.w[1] &&
			(sig_n_prime192.w[0] == sig_y.w[0]) {
			res = 1
			return res, pfpsf
		}
		res = 0
		if ((sig_n_prime192.w[2] > 0) ||
			(sig_n_prime192.w[1] > sig_y.w[1]) ||
			(sig_n_prime192.w[1] == sig_y.w[1] &&
				sig_n_prime192.w[0] > sig_y.w[0])) !=
			((y.w[1] & MASK_SIGN64) == MASK_SIGN64) {
			res = 1
		}
		return res, pfpsf
	}

	diff = exp_y - exp_x
	if diff > 33 {
		res = 0
		if (x.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	if diff > 19 {
		sig_n_prime256 = __mul_128x128_to_256(sig_y, bid_ten2k128[diff-20])
		if sig_n_prime256.w[3] == 0 && (sig_n_prime256.w[2] == 0) &&
			sig_n_prime256.w[1] == sig_x.w[1] &&
			(sig_n_prime256.w[0] == sig_x.w[0]) {
			res = 1
			return res, pfpsf
		}
		res = 0
		if (sig_n_prime256.w[3] == 0 && sig_n_prime256.w[2] == 0 &&
			(sig_n_prime256.w[1] < sig_x.w[1] ||
				(sig_n_prime256.w[1] == sig_x.w[1] &&
					sig_n_prime256.w[0] < sig_x.w[0]))) !=
			((x.w[1] & MASK_SIGN64) == MASK_SIGN64) {
			res = 1
		}
		return res, pfpsf
	}
	sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[diff], sig_y)
	if (sig_n_prime192.w[2] == 0) && sig_n_prime192.w[1] == sig_x.w[1] &&
		(sig_n_prime192.w[0] == sig_x.w[0]) {
		res = 1
		return res, pfpsf
	}
	res = 0
	if (sig_n_prime192.w[2] == 0 &&
		(sig_n_prime192.w[1] < sig_x.w[1] ||
			(sig_n_prime192.w[1] == sig_x.w[1] &&
				sig_n_prime192.w[0] < sig_x.w[0]))) !=
		((y.w[1] & MASK_SIGN64) == MASK_SIGN64) {
		res = 1
	}
	return res, pfpsf
}

// Bid128QuietOrdered - Intel bid128_quiet_ordered 기계적 포팅
func Bid128QuietOrdered(x, y BID_UINT128) (int, uint32) {
	var res int
	var pfpsf uint32

	if ((x.w[1] & NAN_MASK64) == NAN_MASK64) ||
		((y.w[1] & NAN_MASK64) == NAN_MASK64) {
		if (x.w[1]&SNAN_MASK64) == SNAN_MASK64 ||
			(y.w[1]&SNAN_MASK64) == SNAN_MASK64 {
			pfpsf |= BID_INVALID_EXCEPTION
		}
		res = 0
		return res, pfpsf
	}
	res = 1
	return res, pfpsf
}

// Bid128QuietUnordered - Intel bid128_quiet_unordered 기계적 포팅
func Bid128QuietUnordered(x, y BID_UINT128) (int, uint32) {
	var res int
	var pfpsf uint32

	if ((x.w[1] & NAN_MASK64) == NAN_MASK64) ||
		((y.w[1] & NAN_MASK64) == NAN_MASK64) {
		if (x.w[1]&SNAN_MASK64) == SNAN_MASK64 ||
			(y.w[1]&SNAN_MASK64) == SNAN_MASK64 {
			pfpsf |= BID_INVALID_EXCEPTION
		}
		res = 1
		return res, pfpsf
	}
	res = 0
	return res, pfpsf
}

// Bid128SignalingGreater - Intel bid128_signaling_greater 기계적 포팅
func Bid128SignalingGreater(x, y BID_UINT128) (int, uint32) {
	var res int
	var exp_x, exp_y int
	var diff int
	var sig_x, sig_y BID_UINT128
	var sig_n_prime192 BID_UINT192
	var sig_n_prime256 BID_UINT256
	var pfpsf uint32
	var x_is_zero, y_is_zero, non_canon_x, non_canon_y int

	if ((x.w[1] & NAN_MASK64) == NAN_MASK64) ||
		((y.w[1] & NAN_MASK64) == NAN_MASK64) {
		pfpsf |= BID_INVALID_EXCEPTION
		res = 0
		return res, pfpsf
	}
	if x.w[0] == y.w[0] && x.w[1] == y.w[1] {
		res = 0
		return res, pfpsf
	}
	if (x.w[1] & INFINITY_MASK64) == INFINITY_MASK64 {
		if (x.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = 0
			return res, pfpsf
		} else {
			res = 0
			if ((y.w[1] & INFINITY_MASK64) != INFINITY_MASK64) ||
				((y.w[1] & MASK_SIGN64) == MASK_SIGN64) {
				res = 1
			}
			return res, pfpsf
		}
	} else if (y.w[1] & INFINITY_MASK64) == INFINITY_MASK64 {
		res = 0
		if (y.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	sig_x.w[1] = x.w[1] & 0x0001ffffffffffff
	sig_x.w[0] = x.w[0]
	exp_x = int((x.w[1] >> 49) & 0x000000000003fff)

	if (sig_x.w[1] > 0x0001ed09bead87c0) ||
		((sig_x.w[1] == 0x0001ed09bead87c0) &&
			(sig_x.w[0] > 0x378d8e63ffffffff)) ||
		((x.w[1] & 0x6000000000000000) == 0x6000000000000000) {
		non_canon_x = 1
	} else {
		non_canon_x = 0
	}

	exp_y = int((y.w[1] >> 49) & 0x0000000000003fff)
	sig_y.w[1] = y.w[1] & 0x0001ffffffffffff
	sig_y.w[0] = y.w[0]

	if (sig_y.w[1] > 0x0001ed09bead87c0) ||
		((sig_y.w[1] == 0x0001ed09bead87c0) &&
			(sig_y.w[0] > 0x378d8e63ffffffff)) ||
		((y.w[1] & 0x6000000000000000) == 0x6000000000000000) {
		non_canon_y = 1
	} else {
		non_canon_y = 0
	}

	if non_canon_x != 0 || ((sig_x.w[1] == 0) && (sig_x.w[0] == 0)) {
		x_is_zero = 1
	}
	if non_canon_y != 0 || ((sig_y.w[1] == 0) && (sig_y.w[0] == 0)) {
		y_is_zero = 1
	}
	if x_is_zero != 0 && y_is_zero != 0 {
		res = 0
		return res, pfpsf
	} else if x_is_zero != 0 {
		res = 0
		if (y.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	} else if y_is_zero != 0 {
		res = 0
		if (x.w[1] & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	if ((x.w[1] ^ y.w[1]) & MASK_SIGN64) == MASK_SIGN64 {
		res = 0
		if (y.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	if exp_y == exp_x {
		res = 0
		if ((sig_x.w[1] > sig_y.w[1]) ||
			(sig_x.w[1] == sig_y.w[1] && sig_x.w[0] >= sig_y.w[0])) !=
			((x.w[1] & MASK_SIGN64) == MASK_SIGN64) {
			res = 1
		}
		return res, pfpsf
	}
	if (sig_x.w[1] > sig_y.w[1] ||
		(sig_x.w[1] == sig_y.w[1] && sig_x.w[0] > sig_y.w[0])) &&
		exp_x >= exp_y {
		res = 0
		if (x.w[1] & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	if (sig_x.w[1] < sig_y.w[1] ||
		(sig_x.w[1] == sig_y.w[1] && sig_x.w[0] < sig_y.w[0])) &&
		exp_x <= exp_y {
		res = 0
		if (x.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}

	diff = exp_x - exp_y
	if diff > 0 {
		if diff > 33 {
			res = 0
			if (x.w[1] & MASK_SIGN64) != MASK_SIGN64 {
				res = 1
			}
			return res, pfpsf
		}
		if diff > 19 {
			sig_n_prime256 = __mul_128x128_to_256(sig_x, bid_ten2k128[diff-20])
			if sig_n_prime256.w[3] == 0 && (sig_n_prime256.w[2] == 0) &&
				sig_n_prime256.w[1] == sig_y.w[1] &&
				(sig_n_prime256.w[0] == sig_y.w[0]) {
				res = 0
				return res, pfpsf
			}
			res = 0
			if (((sig_n_prime256.w[3] > 0) || sig_n_prime256.w[2] > 0) ||
				(sig_n_prime256.w[1] > sig_y.w[1]) ||
				(sig_n_prime256.w[1] == sig_y.w[1] &&
					sig_n_prime256.w[0] > sig_y.w[0])) !=
				((y.w[1] & MASK_SIGN64) == MASK_SIGN64) {
				res = 1
			}
			return res, pfpsf
		}
		sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[diff], sig_x)
		if (sig_n_prime192.w[2] == 0) && sig_n_prime192.w[1] == sig_y.w[1] &&
			(sig_n_prime192.w[0] == sig_y.w[0]) {
			res = 0
			return res, pfpsf
		}
		res = 0
		if ((sig_n_prime192.w[2] > 0) ||
			(sig_n_prime192.w[1] > sig_y.w[1]) ||
			(sig_n_prime192.w[1] == sig_y.w[1] &&
				sig_n_prime192.w[0] > sig_y.w[0])) !=
			((y.w[1] & MASK_SIGN64) == MASK_SIGN64) {
			res = 1
		}
		return res, pfpsf
	}

	diff = exp_y - exp_x
	if diff > 33 {
		res = 0
		if (x.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	if diff > 19 {
		sig_n_prime256 = __mul_128x128_to_256(sig_y, bid_ten2k128[diff-20])
		if sig_n_prime256.w[3] == 0 && (sig_n_prime256.w[2] == 0) &&
			sig_n_prime256.w[1] == sig_x.w[1] &&
			(sig_n_prime256.w[0] == sig_x.w[0]) {
			res = 0
			return res, pfpsf
		}
		res = 0
		if (sig_n_prime256.w[3] != 0 || sig_n_prime256.w[2] != 0 ||
			(sig_n_prime256.w[1] > sig_x.w[1] ||
				(sig_n_prime256.w[1] == sig_x.w[1] &&
					sig_n_prime256.w[0] > sig_x.w[0]))) !=
			((x.w[1] & MASK_SIGN64) != MASK_SIGN64) {
			res = 1
		}
		return res, pfpsf
	}
	sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[diff], sig_y)
	if (sig_n_prime192.w[2] == 0) && sig_n_prime192.w[1] == sig_x.w[1] &&
		(sig_n_prime192.w[0] == sig_x.w[0]) {
		res = 0
		return res, pfpsf
	}
	res = 0
	if (sig_n_prime192.w[2] != 0 ||
		(sig_n_prime192.w[1] > sig_x.w[1] ||
			(sig_n_prime192.w[1] == sig_x.w[1] &&
				sig_n_prime192.w[0] > sig_x.w[0]))) !=
		((y.w[1] & MASK_SIGN64) != MASK_SIGN64) {
		res = 1
	}
	return res, pfpsf
}

// Bid128SignalingGreaterEqual - Intel bid128_signaling_greater_equal 기계적 포팅
func Bid128SignalingGreaterEqual(x, y BID_UINT128) (int, uint32) {
	// Same as quiet_greater_equal but NaN always sets INVALID
	res, pfpsf := Bid128QuietGreaterEqual(x, y)
	if ((x.w[1] & NAN_MASK64) == NAN_MASK64) ||
		((y.w[1] & NAN_MASK64) == NAN_MASK64) {
		pfpsf |= BID_INVALID_EXCEPTION
		return 0, pfpsf
	}
	return res, pfpsf
}

// Bid128SignalingGreaterUnordered - Intel bid128_signaling_greater_unordered 기계적 포팅
func Bid128SignalingGreaterUnordered(x, y BID_UINT128) (int, uint32) {
	// Same as quiet_greater_unordered but NaN always sets INVALID
	res, pfpsf := Bid128QuietGreaterUnordered(x, y)
	if ((x.w[1] & NAN_MASK64) == NAN_MASK64) ||
		((y.w[1] & NAN_MASK64) == NAN_MASK64) {
		pfpsf |= BID_INVALID_EXCEPTION
		return 1, pfpsf
	}
	return res, pfpsf
}

// Bid128SignalingLess - Intel bid128_signaling_less 기계적 포팅
func Bid128SignalingLess(x, y BID_UINT128) (int, uint32) {
	// Same as quiet_less but NaN always sets INVALID
	res, pfpsf := Bid128QuietLess(x, y)
	if ((x.w[1] & NAN_MASK64) == NAN_MASK64) ||
		((y.w[1] & NAN_MASK64) == NAN_MASK64) {
		pfpsf |= BID_INVALID_EXCEPTION
		return 0, pfpsf
	}
	return res, pfpsf
}

// Bid128SignalingLessEqual - Intel bid128_signaling_less_equal 기계적 포팅
func Bid128SignalingLessEqual(x, y BID_UINT128) (int, uint32) {
	// Same as quiet_less_equal but NaN always sets INVALID
	res, pfpsf := Bid128QuietLessEqual(x, y)
	if ((x.w[1] & NAN_MASK64) == NAN_MASK64) ||
		((y.w[1] & NAN_MASK64) == NAN_MASK64) {
		pfpsf |= BID_INVALID_EXCEPTION
		return 0, pfpsf
	}
	return res, pfpsf
}

// Bid128SignalingLessUnordered - Intel bid128_signaling_less_unordered 기계적 포팅
func Bid128SignalingLessUnordered(x, y BID_UINT128) (int, uint32) {
	// Same as quiet_less_unordered but NaN always sets INVALID
	res, pfpsf := Bid128QuietLessUnordered(x, y)
	if ((x.w[1] & NAN_MASK64) == NAN_MASK64) ||
		((y.w[1] & NAN_MASK64) == NAN_MASK64) {
		pfpsf |= BID_INVALID_EXCEPTION
		return 1, pfpsf
	}
	return res, pfpsf
}

// Bid128SignalingNotGreater - Intel bid128_signaling_not_greater 기계적 포팅
func Bid128SignalingNotGreater(x, y BID_UINT128) (int, uint32) {
	// Same as quiet_not_greater but NaN always sets INVALID (not just sNaN)
	var res int
	var exp_x, exp_y int
	var diff int
	var sig_x, sig_y BID_UINT128
	var sig_n_prime192 BID_UINT192
	var sig_n_prime256 BID_UINT256
	var pfpsf uint32
	var x_is_zero, y_is_zero, non_canon_x, non_canon_y int

	if ((x.w[1] & NAN_MASK64) == NAN_MASK64) ||
		((y.w[1] & NAN_MASK64) == NAN_MASK64) {
		pfpsf |= BID_INVALID_EXCEPTION
		res = 1
		return res, pfpsf
	}
	if x.w[0] == y.w[0] && x.w[1] == y.w[1] {
		res = 1
		return res, pfpsf
	}
	if (x.w[1] & INFINITY_MASK64) == INFINITY_MASK64 {
		if (x.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
			return res, pfpsf
		} else {
			res = 0
			if ((y.w[1] & INFINITY_MASK64) == INFINITY_MASK64) &&
				((y.w[1] & MASK_SIGN64) != MASK_SIGN64) {
				res = 1
			}
			return res, pfpsf
		}
	} else if (y.w[1] & INFINITY_MASK64) == INFINITY_MASK64 {
		res = 0
		if (y.w[1] & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	sig_x.w[1] = x.w[1] & 0x0001ffffffffffff
	sig_x.w[0] = x.w[0]
	exp_x = int((x.w[1] >> 49) & 0x000000000003fff)

	if (sig_x.w[1] > 0x0001ed09bead87c0) ||
		((sig_x.w[1] == 0x0001ed09bead87c0) &&
			(sig_x.w[0] > 0x378d8e63ffffffff)) ||
		((x.w[1] & 0x6000000000000000) == 0x6000000000000000) {
		non_canon_x = 1
	} else {
		non_canon_x = 0
	}

	exp_y = int((y.w[1] >> 49) & 0x0000000000003fff)
	sig_y.w[1] = y.w[1] & 0x0001ffffffffffff
	sig_y.w[0] = y.w[0]

	if (sig_y.w[1] > 0x0001ed09bead87c0) ||
		((sig_y.w[1] == 0x0001ed09bead87c0) &&
			(sig_y.w[0] > 0x378d8e63ffffffff)) ||
		((y.w[1] & 0x6000000000000000) == 0x6000000000000000) {
		non_canon_y = 1
	} else {
		non_canon_y = 0
	}

	if non_canon_x != 0 || ((sig_x.w[1] == 0) && (sig_x.w[0] == 0)) {
		x_is_zero = 1
	}
	if non_canon_y != 0 || ((sig_y.w[1] == 0) && (sig_y.w[0] == 0)) {
		y_is_zero = 1
	}
	if x_is_zero != 0 && y_is_zero != 0 {
		res = 1
		return res, pfpsf
	} else if x_is_zero != 0 {
		res = 0
		if (y.w[1] & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	} else if y_is_zero != 0 {
		res = 0
		if (x.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	if ((x.w[1] ^ y.w[1]) & MASK_SIGN64) == MASK_SIGN64 {
		res = 0
		if (y.w[1] & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	if exp_y == exp_x {
		res = 0
		if ((sig_x.w[1] > sig_y.w[1]) ||
			(sig_x.w[1] == sig_y.w[1] && sig_x.w[0] >= sig_y.w[0])) !=
			((x.w[1] & MASK_SIGN64) != MASK_SIGN64) {
			res = 1
		}
		return res, pfpsf
	}
	if (sig_x.w[1] > sig_y.w[1] ||
		(sig_x.w[1] == sig_y.w[1] && sig_x.w[0] > sig_y.w[0])) &&
		exp_x >= exp_y {
		res = 0
		if (x.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	if (sig_x.w[1] < sig_y.w[1] ||
		(sig_x.w[1] == sig_y.w[1] && sig_x.w[0] < sig_y.w[0])) &&
		exp_x <= exp_y {
		res = 0
		if (x.w[1] & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}

	diff = exp_x - exp_y
	if diff > 0 {
		if diff > 33 {
			res = 0
			if (x.w[1] & MASK_SIGN64) == MASK_SIGN64 {
				res = 1
			}
			return res, pfpsf
		}
		if diff > 19 {
			sig_n_prime256 = __mul_128x128_to_256(sig_x, bid_ten2k128[diff-20])
			if sig_n_prime256.w[3] == 0 && (sig_n_prime256.w[2] == 0) &&
				sig_n_prime256.w[1] == sig_y.w[1] &&
				(sig_n_prime256.w[0] == sig_y.w[0]) {
				res = 1
				return res, pfpsf
			}
			res = 0
			if (((sig_n_prime256.w[3] > 0) || sig_n_prime256.w[2] > 0) ||
				(sig_n_prime256.w[1] > sig_y.w[1]) ||
				(sig_n_prime256.w[1] == sig_y.w[1] &&
					sig_n_prime256.w[0] > sig_y.w[0])) !=
				((y.w[1] & MASK_SIGN64) != MASK_SIGN64) {
				res = 1
			}
			return res, pfpsf
		}
		sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[diff], sig_x)
		if (sig_n_prime192.w[2] == 0) && sig_n_prime192.w[1] == sig_y.w[1] &&
			(sig_n_prime192.w[0] == sig_y.w[0]) {
			res = 1
			return res, pfpsf
		}
		res = 0
		if ((sig_n_prime192.w[2] > 0) ||
			(sig_n_prime192.w[1] > sig_y.w[1]) ||
			(sig_n_prime192.w[1] == sig_y.w[1] &&
				sig_n_prime192.w[0] > sig_y.w[0])) !=
			((y.w[1] & MASK_SIGN64) != MASK_SIGN64) {
			res = 1
		}
		return res, pfpsf
	}

	diff = exp_y - exp_x
	if diff > 33 {
		res = 0
		if (x.w[1] & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	if diff > 19 {
		sig_n_prime256 = __mul_128x128_to_256(sig_y, bid_ten2k128[diff-20])
		if sig_n_prime256.w[3] == 0 && (sig_n_prime256.w[2] == 0) &&
			sig_n_prime256.w[1] == sig_x.w[1] &&
			(sig_n_prime256.w[0] == sig_x.w[0]) {
			res = 1
			return res, pfpsf
		}
		res = 0
		if (sig_n_prime256.w[3] != 0 || sig_n_prime256.w[2] != 0 ||
			(sig_n_prime256.w[1] > sig_x.w[1] ||
				(sig_n_prime256.w[1] == sig_x.w[1] &&
					sig_n_prime256.w[0] > sig_x.w[0]))) !=
			((x.w[1] & MASK_SIGN64) == MASK_SIGN64) {
			res = 1
		}
		return res, pfpsf
	}
	sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[diff], sig_y)
	if (sig_n_prime192.w[2] == 0) && sig_n_prime192.w[1] == sig_x.w[1] &&
		(sig_n_prime192.w[0] == sig_x.w[0]) {
		res = 1
		return res, pfpsf
	}
	res = 0
	if (sig_n_prime192.w[2] != 0 ||
		(sig_n_prime192.w[1] > sig_x.w[1] ||
			(sig_n_prime192.w[1] == sig_x.w[1] &&
				sig_n_prime192.w[0] > sig_x.w[0]))) !=
		((y.w[1] & MASK_SIGN64) == MASK_SIGN64) {
		res = 1
	}
	return res, pfpsf
}

// Bid128SignalingNotLess - Intel bid128_signaling_not_less 기계적 포팅
func Bid128SignalingNotLess(x, y BID_UINT128) (int, uint32) {
	// Same as quiet_not_less but NaN always sets INVALID (not just sNaN)
	var res int
	var exp_x, exp_y int
	var diff int
	var sig_x, sig_y BID_UINT128
	var sig_n_prime192 BID_UINT192
	var sig_n_prime256 BID_UINT256
	var pfpsf uint32
	var x_is_zero, y_is_zero, non_canon_x, non_canon_y int

	if ((x.w[1] & NAN_MASK64) == NAN_MASK64) ||
		((y.w[1] & NAN_MASK64) == NAN_MASK64) {
		pfpsf |= BID_INVALID_EXCEPTION
		res = 1
		return res, pfpsf
	}
	if x.w[0] == y.w[0] && x.w[1] == y.w[1] {
		res = 1
		return res, pfpsf
	}
	if (x.w[1] & INFINITY_MASK64) == INFINITY_MASK64 {
		if (x.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = 0
			if ((y.w[1] & INFINITY_MASK64) == INFINITY_MASK64) &&
				(y.w[1]&MASK_SIGN64) == MASK_SIGN64 {
				res = 1
			}
			return res, pfpsf
		} else {
			res = 1
			return res, pfpsf
		}
	} else if (y.w[1] & INFINITY_MASK64) == INFINITY_MASK64 {
		res = 0
		if (y.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	sig_x.w[1] = x.w[1] & 0x0001ffffffffffff
	sig_x.w[0] = x.w[0]
	exp_x = int((x.w[1] >> 49) & 0x000000000003fff)

	if (sig_x.w[1] > 0x0001ed09bead87c0) ||
		((sig_x.w[1] == 0x0001ed09bead87c0) &&
			(sig_x.w[0] > 0x378d8e63ffffffff)) ||
		((x.w[1] & 0x6000000000000000) == 0x6000000000000000) {
		non_canon_x = 1
	} else {
		non_canon_x = 0
	}

	exp_y = int((y.w[1] >> 49) & 0x0000000000003fff)
	sig_y.w[1] = y.w[1] & 0x0001ffffffffffff
	sig_y.w[0] = y.w[0]

	if (sig_y.w[1] > 0x0001ed09bead87c0) ||
		((sig_y.w[1] == 0x0001ed09bead87c0) &&
			(sig_y.w[0] > 0x378d8e63ffffffff)) ||
		((y.w[1] & 0x6000000000000000) == 0x6000000000000000) {
		non_canon_y = 1
	} else {
		non_canon_y = 0
	}

	if non_canon_x != 0 || ((sig_x.w[1] == 0) && (sig_x.w[0] == 0)) {
		x_is_zero = 1
	}
	if non_canon_y != 0 || ((sig_y.w[1] == 0) && (sig_y.w[0] == 0)) {
		y_is_zero = 1
	}
	if x_is_zero != 0 && y_is_zero != 0 {
		res = 1
		return res, pfpsf
	} else if x_is_zero != 0 {
		res = 0
		if (y.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	} else if y_is_zero != 0 {
		res = 0
		if (x.w[1] & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	if ((x.w[1] ^ y.w[1]) & MASK_SIGN64) == MASK_SIGN64 {
		res = 0
		if (y.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	if exp_y == exp_x {
		res = 0
		if ((sig_x.w[1] > sig_y.w[1]) ||
			(sig_x.w[1] == sig_y.w[1] && sig_x.w[0] >= sig_y.w[0])) !=
			((x.w[1] & MASK_SIGN64) == MASK_SIGN64) {
			res = 1
		}
		return res, pfpsf
	}
	if sig_x.w[1] >= sig_y.w[1] && sig_x.w[0] >= sig_y.w[0] &&
		exp_x > exp_y {
		res = 0
		if (x.w[1] & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	if sig_x.w[1] <= sig_y.w[1] && sig_x.w[0] <= sig_y.w[0] &&
		exp_x < exp_y {
		res = 0
		if (x.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}

	diff = exp_x - exp_y
	if diff > 0 {
		if diff > 33 {
			res = 0
			if (x.w[1] & MASK_SIGN64) != MASK_SIGN64 {
				res = 1
			}
			return res, pfpsf
		}
		if diff > 19 {
			sig_n_prime256 = __mul_128x128_to_256(sig_x, bid_ten2k128[diff-20])
			if sig_n_prime256.w[3] == 0 && (sig_n_prime256.w[2] == 0) &&
				sig_n_prime256.w[1] == sig_y.w[1] &&
				(sig_n_prime256.w[0] == sig_y.w[0]) {
				res = 1
				return res, pfpsf
			}
			res = 0
			if (((sig_n_prime256.w[3] > 0) || sig_n_prime256.w[2] > 0) ||
				(sig_n_prime256.w[1] > sig_y.w[1]) ||
				(sig_n_prime256.w[1] == sig_y.w[1] &&
					sig_n_prime256.w[0] > sig_y.w[0])) !=
				((y.w[1] & MASK_SIGN64) == MASK_SIGN64) {
				res = 1
			}
			return res, pfpsf
		}
		sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[diff], sig_x)
		if (sig_n_prime192.w[2] == 0) && sig_n_prime192.w[1] == sig_y.w[1] &&
			(sig_n_prime192.w[0] == sig_y.w[0]) {
			res = 1
			return res, pfpsf
		}
		res = 0
		if ((sig_n_prime192.w[2] > 0) ||
			(sig_n_prime192.w[1] > sig_y.w[1]) ||
			(sig_n_prime192.w[1] == sig_y.w[1] &&
				sig_n_prime192.w[0] > sig_y.w[0])) !=
			((y.w[1] & MASK_SIGN64) == MASK_SIGN64) {
			res = 1
		}
		return res, pfpsf
	}

	diff = exp_y - exp_x
	if diff > 33 {
		res = 0
		if (x.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res, pfpsf
	}
	if diff > 19 {
		sig_n_prime256 = __mul_128x128_to_256(sig_y, bid_ten2k128[diff-20])
		if sig_n_prime256.w[3] == 0 && (sig_n_prime256.w[2] == 0) &&
			sig_n_prime256.w[1] == sig_x.w[1] &&
			(sig_n_prime256.w[0] == sig_x.w[0]) {
			res = 1
			return res, pfpsf
		}
		res = 0
		if (sig_n_prime256.w[3] == 0 && sig_n_prime256.w[2] == 0 &&
			(sig_n_prime256.w[1] < sig_x.w[1] ||
				(sig_n_prime256.w[1] == sig_x.w[1] &&
					sig_n_prime256.w[0] < sig_x.w[0]))) !=
			((x.w[1] & MASK_SIGN64) == MASK_SIGN64) {
			res = 1
		}
		return res, pfpsf
	}
	sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[diff], sig_y)
	if (sig_n_prime192.w[2] == 0) && sig_n_prime192.w[1] == sig_x.w[1] &&
		(sig_n_prime192.w[0] == sig_x.w[0]) {
		res = 1
		return res, pfpsf
	}
	res = 0
	if (sig_n_prime192.w[2] == 0 &&
		(sig_n_prime192.w[1] < sig_x.w[1] ||
			(sig_n_prime192.w[1] == sig_x.w[1] &&
				sig_n_prime192.w[0] < sig_x.w[0]))) !=
		((y.w[1] & MASK_SIGN64) == MASK_SIGN64) {
		res = 1
	}
	return res, pfpsf
}

// Bid128TotalOrder - Intel bid128_totalOrder 기계적 포팅
func Bid128TotalOrder(x, y BID_UINT128) int {
	var res int
	var exp_x, exp_y int
	var sig_x, sig_y, pyld_y, pyld_x BID_UINT128
	var sig_n_prime192 BID_UINT192
	var sig_n_prime256 BID_UINT256
	var x_is_zero, y_is_zero int

	// NaN (CASE 1)
	if (x.w[1] & NAN_MASK64) == NAN_MASK64 {
		if (x.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			// x is -NaN
			if (y.w[1]&NAN_MASK64) != NAN_MASK64 ||
				(y.w[1]&MASK_SIGN64) != MASK_SIGN64 {
				res = 1
				return res
			} else {
				// both -NaN
				pyld_x.w[1] = x.w[1] & 0x00003fffffffffff
				pyld_x.w[0] = x.w[0]
				pyld_y.w[1] = y.w[1] & 0x00003fffffffffff
				pyld_y.w[0] = y.w[0]
				if (pyld_x.w[1] > 0x0000314dc6448d93) ||
					((pyld_x.w[1] == 0x0000314dc6448d93) &&
						(pyld_x.w[0] > 0x38c15b09ffffffff)) {
					pyld_x.w[1] = 0
					pyld_x.w[0] = 0
				}
				if (pyld_y.w[1] > 0x0000314dc6448d93) ||
					((pyld_y.w[1] == 0x0000314dc6448d93) &&
						(pyld_y.w[0] > 0x38c15b09ffffffff)) {
					pyld_y.w[1] = 0
					pyld_y.w[0] = 0
				}
				if !(((y.w[1] & SNAN_MASK64) == SNAN_MASK64) !=
					((x.w[1] & SNAN_MASK64) == SNAN_MASK64)) {
					// both SNaN or both QNaN
					if (pyld_x.w[1] > pyld_y.w[1]) ||
						((pyld_x.w[1] == pyld_y.w[1]) &&
							(pyld_x.w[0] >= pyld_y.w[0])) {
						res = 1
					} else {
						res = 0
					}
					return res
				} else {
					res = 0
					if (y.w[1] & SNAN_MASK64) == SNAN_MASK64 {
						res = 1
					}
					return res
				}
			}
		} else {
			// x is +NaN
			if (y.w[1]&NAN_MASK64) != NAN_MASK64 ||
				(y.w[1]&MASK_SIGN64) == MASK_SIGN64 {
				res = 0
				return res
			} else {
				// both +NaN
				pyld_x.w[1] = x.w[1] & 0x00003fffffffffff
				pyld_x.w[0] = x.w[0]
				pyld_y.w[1] = y.w[1] & 0x00003fffffffffff
				pyld_y.w[0] = y.w[0]
				if (pyld_x.w[1] > 0x0000314dc6448d93) ||
					((pyld_x.w[1] == 0x0000314dc6448d93) &&
						(pyld_x.w[0] > 0x38c15b09ffffffff)) {
					pyld_x.w[1] = 0
					pyld_x.w[0] = 0
				}
				if (pyld_y.w[1] > 0x0000314dc6448d93) ||
					((pyld_y.w[1] == 0x0000314dc6448d93) &&
						(pyld_y.w[0] > 0x38c15b09ffffffff)) {
					pyld_y.w[1] = 0
					pyld_y.w[0] = 0
				}
				if !(((y.w[1] & SNAN_MASK64) == SNAN_MASK64) !=
					((x.w[1] & SNAN_MASK64) == SNAN_MASK64)) {
					// both SNaN or both QNaN
					if (pyld_x.w[1] < pyld_y.w[1]) ||
						((pyld_x.w[1] == pyld_y.w[1]) &&
							(pyld_x.w[0] <= pyld_y.w[0])) {
						res = 1
					} else {
						res = 0
					}
					return res
				} else {
					res = 0
					if (x.w[1] & SNAN_MASK64) == SNAN_MASK64 {
						res = 1
					}
					return res
				}
			}
		}
	} else if (y.w[1] & NAN_MASK64) == NAN_MASK64 {
		res = 0
		if (y.w[1] & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res
	}
	// SIMPLE (CASE 2)
	if (x.w[1] == y.w[1]) && (x.w[0] == y.w[0]) {
		res = 1
		return res
	}
	// OPPOSITE SIGNS (CASE 3)
	if ((x.w[1] & MASK_SIGN64) == MASK_SIGN64) != ((y.w[1] & MASK_SIGN64) == MASK_SIGN64) {
		res = 0
		if (x.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res
	}
	// INFINITY (CASE 4)
	if (x.w[1] & INFINITY_MASK64) == INFINITY_MASK64 {
		if (x.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
			return res
		} else {
			res = 0
			if (y.w[1] & INFINITY_MASK64) == INFINITY_MASK64 {
				res = 1
			}
			return res
		}
	} else if (y.w[1] & INFINITY_MASK64) == INFINITY_MASK64 {
		res = 0
		if (y.w[1] & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res
	}
	// CONVERT x
	sig_x.w[1] = x.w[1] & 0x0001ffffffffffff
	sig_x.w[0] = x.w[0]
	exp_x = int((x.w[1] >> 49) & 0x000000000003fff)

	// CHECK IF x IS CANONICAL
	if (((sig_x.w[1] > 0x0001ed09bead87c0) ||
		((sig_x.w[1] == 0x0001ed09bead87c0) &&
			(sig_x.w[0] > 0x378d8e63ffffffff))) &&
		((x.w[1] & 0x6000000000000000) != 0x6000000000000000)) ||
		((x.w[1] & 0x6000000000000000) == 0x6000000000000000) ||
		((sig_x.w[1] == 0) && (sig_x.w[0] == 0)) {
		x_is_zero = 1
		if (x.w[1] & 0x6000000000000000) == 0x6000000000000000 {
			exp_x = int((x.w[1] >> 47) & 0x000000000003fff)
		}
	}
	// CONVERT y
	exp_y = int((y.w[1] >> 49) & 0x0000000000003fff)
	sig_y.w[1] = y.w[1] & 0x0001ffffffffffff
	sig_y.w[0] = y.w[0]

	// CHECK IF y IS CANONICAL
	if (((sig_y.w[1] > 0x0001ed09bead87c0) ||
		((sig_y.w[1] == 0x0001ed09bead87c0) &&
			(sig_y.w[0] > 0x378d8e63ffffffff))) &&
		((y.w[1] & 0x6000000000000000) != 0x6000000000000000)) ||
		((y.w[1] & 0x6000000000000000) == 0x6000000000000000) ||
		((sig_y.w[1] == 0) && (sig_y.w[0] == 0)) {
		y_is_zero = 1
		if (y.w[1] & 0x6000000000000000) == 0x6000000000000000 {
			exp_y = int((y.w[1] >> 47) & 0x000000000003fff)
		}
	}
	// ZERO (CASE 5)
	if x_is_zero != 0 && y_is_zero != 0 {
		if exp_x == exp_y {
			res = 1
			return res
		}
		res = 0
		if (exp_x <= exp_y) != ((x.w[1] & MASK_SIGN64) == MASK_SIGN64) {
			res = 1
		}
		return res
	}
	if x_is_zero != 0 {
		res = 0
		if (y.w[1] & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res
	}
	if y_is_zero != 0 {
		res = 0
		if (x.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res
	}
	// REDUNDANT REPRESENTATIONS (CASE 6)
	if ((sig_x.w[1] > sig_y.w[1]) ||
		(sig_x.w[1] == sig_y.w[1] && sig_x.w[0] > sig_y.w[0])) &&
		exp_x >= exp_y {
		res = 0
		if (x.w[1] & MASK_SIGN64) == MASK_SIGN64 {
			res = 1
		}
		return res
	}
	if ((sig_x.w[1] < sig_y.w[1]) ||
		(sig_x.w[1] == sig_y.w[1] && sig_x.w[0] < sig_y.w[0])) &&
		exp_x <= exp_y {
		res = 0
		if (x.w[1] & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res
	}
	if exp_x > exp_y {
		if exp_x-exp_y > 33 {
			res = 0
			if (x.w[1] & MASK_SIGN64) == MASK_SIGN64 {
				res = 1
			}
			return res
		}
		if exp_x-exp_y > 19 {
			sig_n_prime256 = __mul_128x128_to_256(sig_x,
				bid_ten2k128[exp_x-exp_y-20])
			if (sig_n_prime256.w[3] == 0) && (sig_n_prime256.w[2] == 0) &&
				(sig_n_prime256.w[1] == sig_y.w[1]) &&
				(sig_n_prime256.w[0] == sig_y.w[0]) {
				res = 0
				if (exp_x <= exp_y) != ((x.w[1] & MASK_SIGN64) == MASK_SIGN64) {
					res = 1
				}
				return res
			}
			res = 0
			if ((sig_n_prime256.w[3] == 0) && (sig_n_prime256.w[2] == 0) &&
				((sig_n_prime256.w[1] < sig_y.w[1]) ||
					(sig_n_prime256.w[1] == sig_y.w[1] &&
						sig_n_prime256.w[0] < sig_y.w[0]))) !=
				((x.w[1] & MASK_SIGN64) == MASK_SIGN64) {
				res = 1
			}
			return res
		}
		sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[exp_x-exp_y], sig_x)
		if (sig_n_prime192.w[2] == 0) && sig_n_prime192.w[1] == sig_y.w[1] &&
			(sig_n_prime192.w[0] == sig_y.w[0]) {
			res = 0
			if (exp_x <= exp_y) != ((x.w[1] & MASK_SIGN64) == MASK_SIGN64) {
				res = 1
			}
			return res
		}
		res = 0
		if ((sig_n_prime192.w[2] == 0) &&
			((sig_n_prime192.w[1] < sig_y.w[1]) ||
				(sig_n_prime192.w[1] == sig_y.w[1] &&
					sig_n_prime192.w[0] < sig_y.w[0]))) !=
			((x.w[1] & MASK_SIGN64) == MASK_SIGN64) {
			res = 1
		}
		return res
	}
	// exp_y > exp_x
	if exp_y-exp_x > 33 {
		res = 0
		if (x.w[1] & MASK_SIGN64) != MASK_SIGN64 {
			res = 1
		}
		return res
	}
	if exp_y-exp_x > 19 {
		sig_n_prime256 = __mul_128x128_to_256(sig_y,
			bid_ten2k128[exp_y-exp_x-20])
		if (sig_n_prime256.w[3] == 0) && (sig_n_prime256.w[2] == 0) &&
			(sig_n_prime256.w[1] == sig_x.w[1]) &&
			(sig_n_prime256.w[0] == sig_x.w[0]) {
			res = 0
			if (exp_x <= exp_y) != ((x.w[1] & MASK_SIGN64) == MASK_SIGN64) {
				res = 1
			}
			return res
		}
		res = 0
		if ((sig_n_prime256.w[3] != 0) ||
			(sig_n_prime256.w[2] != 0) ||
			(sig_n_prime256.w[1] > sig_x.w[1]) ||
			(sig_n_prime256.w[1] == sig_x.w[1] &&
				sig_n_prime256.w[0] > sig_x.w[0])) !=
			((x.w[1] & MASK_SIGN64) == MASK_SIGN64) {
			res = 1
		}
		return res
	}
	sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[exp_y-exp_x], sig_y)
	if (sig_n_prime192.w[2] == 0) && (sig_n_prime192.w[1] == sig_x.w[1]) &&
		(sig_n_prime192.w[0] == sig_x.w[0]) {
		res = 0
		if (exp_x <= exp_y) != ((x.w[1] & MASK_SIGN64) == MASK_SIGN64) {
			res = 1
		}
		return res
	}
	res = 0
	if ((sig_n_prime192.w[2] != 0) ||
		(sig_n_prime192.w[1] > sig_x.w[1]) ||
		(sig_n_prime192.w[1] == sig_x.w[1] &&
			sig_n_prime192.w[0] > sig_x.w[0])) !=
		((x.w[1] & MASK_SIGN64) == MASK_SIGN64) {
		res = 1
	}
	return res
}

// Bid128TotalOrderMag - Intel bid128_totalOrderMag 기계적 포팅
func Bid128TotalOrderMag(x, y BID_UINT128) int {
	var res int
	var exp_x, exp_y int
	var sig_x, sig_y, pyld_y, pyld_x BID_UINT128
	var sig_n_prime192 BID_UINT192
	var sig_n_prime256 BID_UINT256
	var x_is_zero, y_is_zero int

	// clear sign bits
	x.w[1] = x.w[1] & 0x7fffffffffffffff
	y.w[1] = y.w[1] & 0x7fffffffffffffff

	// NaN (CASE 1)
	if (x.w[1] & NAN_MASK64) == NAN_MASK64 {
		if (y.w[1] & NAN_MASK64) != NAN_MASK64 {
			res = 0
			return res
		} else {
			// both +NaN
			pyld_x.w[1] = x.w[1] & 0x00003fffffffffff
			pyld_x.w[0] = x.w[0]
			pyld_y.w[1] = y.w[1] & 0x00003fffffffffff
			pyld_y.w[0] = y.w[0]
			if (pyld_x.w[1] > 0x0000314dc6448d93) ||
				((pyld_x.w[1] == 0x0000314dc6448d93) &&
					(pyld_x.w[0] > 0x38c15b09ffffffff)) {
				pyld_x.w[1] = 0
				pyld_x.w[0] = 0
			}
			if (pyld_y.w[1] > 0x0000314dc6448d93) ||
				((pyld_y.w[1] == 0x0000314dc6448d93) &&
					(pyld_y.w[0] > 0x38c15b09ffffffff)) {
				pyld_y.w[1] = 0
				pyld_y.w[0] = 0
			}
			if !(((y.w[1] & SNAN_MASK64) == SNAN_MASK64) !=
				((x.w[1] & SNAN_MASK64) == SNAN_MASK64)) {
				if (pyld_x.w[1] < pyld_y.w[1]) ||
					((pyld_x.w[1] == pyld_y.w[1]) &&
						(pyld_x.w[0] <= pyld_y.w[0])) {
					res = 1
				} else {
					res = 0
				}
				return res
			} else {
				res = 0
				if (x.w[1] & SNAN_MASK64) == SNAN_MASK64 {
					res = 1
				}
				return res
			}
		}
	} else if (y.w[1] & NAN_MASK64) == NAN_MASK64 {
		res = 1
		return res
	}
	// SIMPLE (CASE 2)
	if (x.w[1] == y.w[1]) && (x.w[0] == y.w[0]) {
		res = 1
		return res
	}
	// INFINITY (CASE 3)
	if (x.w[1] & INFINITY_MASK64) == INFINITY_MASK64 {
		res = 0
		if (y.w[1] & INFINITY_MASK64) == INFINITY_MASK64 {
			res = 1
		}
		return res
	} else if (y.w[1] & INFINITY_MASK64) == INFINITY_MASK64 {
		res = 1
		return res
	}

	// CONVERT x
	sig_x.w[1] = x.w[1] & 0x0001ffffffffffff
	sig_x.w[0] = x.w[0]
	exp_x = int((x.w[1] >> 49) & 0x000000000003fff)

	if (((sig_x.w[1] > 0x0001ed09bead87c0) ||
		((sig_x.w[1] == 0x0001ed09bead87c0) &&
			(sig_x.w[0] > 0x378d8e63ffffffff))) &&
		((x.w[1] & 0x6000000000000000) != 0x6000000000000000)) ||
		((x.w[1] & 0x6000000000000000) == 0x6000000000000000) ||
		((sig_x.w[1] == 0) && (sig_x.w[0] == 0)) {
		x_is_zero = 1
		if (x.w[1] & 0x6000000000000000) == 0x6000000000000000 {
			exp_x = int((x.w[1] >> 47) & 0x000000000003fff)
		}
	}
	// CONVERT y
	exp_y = int((y.w[1] >> 49) & 0x0000000000003fff)
	sig_y.w[1] = y.w[1] & 0x0001ffffffffffff
	sig_y.w[0] = y.w[0]

	if (((sig_y.w[1] > 0x0001ed09bead87c0) ||
		((sig_y.w[1] == 0x0001ed09bead87c0) &&
			(sig_y.w[0] > 0x378d8e63ffffffff))) &&
		((y.w[1] & 0x6000000000000000) != 0x6000000000000000)) ||
		((y.w[1] & 0x6000000000000000) == 0x6000000000000000) ||
		((sig_y.w[1] == 0) && (sig_y.w[0] == 0)) {
		y_is_zero = 1
		if (y.w[1] & 0x6000000000000000) == 0x6000000000000000 {
			exp_y = int((y.w[1] >> 47) & 0x000000000003fff)
		}
	}
	// ZERO (CASE 4)
	if x_is_zero != 0 && y_is_zero != 0 {
		if exp_x == exp_y {
			res = 1
			return res
		}
		res = 0
		if exp_x <= exp_y {
			res = 1
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
	// REDUNDANT REPRESENTATIONS (CASE 5)
	if ((sig_x.w[1] > sig_y.w[1]) ||
		(sig_x.w[1] == sig_y.w[1] && sig_x.w[0] > sig_y.w[0])) &&
		exp_x >= exp_y {
		res = 0
		return res
	}
	if ((sig_x.w[1] < sig_y.w[1]) ||
		(sig_x.w[1] == sig_y.w[1] && sig_x.w[0] < sig_y.w[0])) &&
		exp_x <= exp_y {
		res = 1
		return res
	}
	if exp_x > exp_y {
		if exp_x-exp_y > 33 {
			res = 0
			return res
		}
		if exp_x-exp_y > 19 {
			sig_n_prime256 = __mul_128x128_to_256(sig_x,
				bid_ten2k128[exp_x-exp_y-20])
			if (sig_n_prime256.w[3] == 0) && (sig_n_prime256.w[2] == 0) &&
				(sig_n_prime256.w[1] == sig_y.w[1]) &&
				(sig_n_prime256.w[0] == sig_y.w[0]) {
				res = 0
				return res
			}
			res = 0
			if (sig_n_prime256.w[3] == 0) && (sig_n_prime256.w[2] == 0) &&
				((sig_n_prime256.w[1] < sig_y.w[1]) ||
					(sig_n_prime256.w[1] == sig_y.w[1] &&
						sig_n_prime256.w[0] < sig_y.w[0])) {
				res = 1
			}
			return res
		}
		sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[exp_x-exp_y], sig_x)
		if (sig_n_prime192.w[2] == 0) && sig_n_prime192.w[1] == sig_y.w[1] &&
			(sig_n_prime192.w[0] == sig_y.w[0]) {
			res = 0
			return res
		}
		res = 0
		if (sig_n_prime192.w[2] == 0) &&
			((sig_n_prime192.w[1] < sig_y.w[1]) ||
				(sig_n_prime192.w[1] == sig_y.w[1] &&
					sig_n_prime192.w[0] < sig_y.w[0])) {
			res = 1
		}
		return res
	}
	// exp_y > exp_x
	if exp_y-exp_x > 33 {
		res = 1
		return res
	}
	if exp_y-exp_x > 19 {
		sig_n_prime256 = __mul_128x128_to_256(sig_y,
			bid_ten2k128[exp_y-exp_x-20])
		if (sig_n_prime256.w[3] == 0) && (sig_n_prime256.w[2] == 0) &&
			(sig_n_prime256.w[1] == sig_x.w[1]) &&
			(sig_n_prime256.w[0] == sig_x.w[0]) {
			res = 1
			return res
		}
		res = 0
		if (sig_n_prime256.w[3] != 0) ||
			(sig_n_prime256.w[2] != 0) ||
			(sig_n_prime256.w[1] > sig_x.w[1]) ||
			(sig_n_prime256.w[1] == sig_x.w[1] &&
				sig_n_prime256.w[0] > sig_x.w[0]) {
			res = 1
		}
		return res
	}
	sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[exp_y-exp_x], sig_y)
	if (sig_n_prime192.w[2] == 0) && (sig_n_prime192.w[1] == sig_x.w[1]) &&
		(sig_n_prime192.w[0] == sig_x.w[0]) {
		res = 1
		return res
	}
	res = 0
	if (sig_n_prime192.w[2] != 0) ||
		(sig_n_prime192.w[1] > sig_x.w[1]) ||
		(sig_n_prime192.w[1] == sig_x.w[1] &&
			sig_n_prime192.w[0] > sig_x.w[0]) {
		res = 1
	}
	return res
}
