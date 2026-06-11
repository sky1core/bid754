package bidgo

// bid64_next.c 기계적 포팅

// Bid64NextUp - Intel bid64_nextup 기계적 포팅
func Bid64NextUp(x uint64) (uint64, uint32) {
	var res uint64
	var x_sign uint64
	var x_exp uint64
	var q1, ind int
	var C1 uint64
	var pfpsf uint32
	var C BID_UINT128

	// check for NaNs and infinities
	if (x & MASK_NAN64) == MASK_NAN64 { // check for NaN
		if (x & 0x0003ffffffffffff) > 999999999999999 {
			x = x & 0xfe00000000000000 // clear G6-G12 and the payload bits
		} else {
			x = x & 0xfe03ffffffffffff // clear G6-G12
		}
		if (x & MASK_SNAN64) == MASK_SNAN64 { // SNaN
			// set invalid flag
			pfpsf |= BID_INVALID_EXCEPTION
			// return quiet (SNaN)
			res = x & 0xfdffffffffffffff
		} else { // QNaN
			res = x
		}
		return res, pfpsf
	} else if (x & MASK_INF64) == MASK_INF64 { // check for Infinity
		if (x & MASK_SIGN64) == 0 { // x is +inf
			res = 0x7800000000000000
		} else { // x is -inf
			res = 0xf7fb86f26fc0ffff // -MAXFP = -999...99 * 10^emax
		}
		return res, pfpsf
	}
	// unpack the argument
	x_sign = x & MASK_SIGN64 // 0 for positive, MASK_SIGN for negative
	// if steering bits are 11 (condition will be 0), then exponent is G[0:w+1] =>
	if (x & MASK_STEERING_BITS64) == MASK_STEERING_BITS64 {
		x_exp = (x & MASK_BINARY_EXPONENT2_64) >> 51 // biased
		C1 = (x & MASK_BINARY_SIG2_64) | MASK_BINARY_OR2_64
		if C1 > 9999999999999999 { // non-canonical
			x_exp = 0
			C1 = 0
		}
	} else {
		x_exp = (x & MASK_BINARY_EXPONENT1_64) >> 53 // biased
		C1 = x & MASK_BINARY_SIG1_64
	}

	// check for zeros (possibly from non-canonical values)
	if C1 == 0x0 { // x is 0
		res = 0x0000000000000001 // MINFP = 1 * 10^emin
	} else { // x is not special and is not zero
		if x == 0x77fb86f26fc0ffff {
			// x = +MAXFP = 999...99 * 10^emax
			res = 0x7800000000000000 // +inf
		} else if x == 0x8000000000000001 {
			// x = -MINFP = 1...99 * 10^emin
			res = 0x8000000000000000 // -0
		} else { // -MAXFP <= x <= -MINFP - 1 ulp OR MINFP <= x <= MAXFP - 1 ulp
			// can add/subtract 1 ulp to the significand

			C.w[0] = C1
			C.w[1] = 0
			q1 = __get_dec_digits64(C)
			// if q1 < P16 then pad the significand with zeros
			if q1 < 16 {
				if x_exp > uint64(16-q1) {
					ind = 16 - q1 // 1 <= ind <= P16 - 1
					// pad with P16 - q1 zeros, until exponent = emin
					// C1 = C1 * 10^ind
					C1 = C1 * bid_ten2k64[ind]
					x_exp = x_exp - uint64(ind)
				} else { // pad with zeros until the exponent reaches emin
					ind = int(x_exp)
					C1 = C1 * bid_ten2k64[ind]
					x_exp = 0
				}
			}
			if x_sign == 0 { // x > 0
				// add 1 ulp (add 1 to the significand)
				C1++
				if C1 == 0x002386f26fc10000 { // if  C1 = 10^16
					C1 = 0x00038d7ea4c68000 // C1 = 10^15
					x_exp++
				}
				// Ok, because MAXFP = 999...99 * 10^emax was caught already
			} else { // x < 0
				// subtract 1 ulp (subtract 1 from the significand)
				C1--
				if C1 == 0x00038d7ea4c67fff && x_exp != 0 { // if  C1 = 10^15 - 1
					C1 = 0x002386f26fc0ffff // C1 = 10^16 - 1
					x_exp--
				}
			}
			// assemble the result
			// if significand has 54 bits
			if (C1 & MASK_BINARY_OR2_64) != 0 {
				res = x_sign | (x_exp << 51) | MASK_STEERING_BITS64 | (C1 & MASK_BINARY_SIG2_64)
			} else { // significand fits in 53 bits
				res = x_sign | (x_exp << 53) | C1
			}
		}
	}
	return res, pfpsf
}

// Bid64NextDown - Intel bid64_nextdown 기계적 포팅
func Bid64NextDown(x uint64) (uint64, uint32) {
	var res uint64
	var x_sign uint64
	var x_exp uint64
	var q1, ind int
	var C1 uint64
	var pfpsf uint32
	var C BID_UINT128

	// check for NaNs and infinities
	if (x & MASK_NAN64) == MASK_NAN64 { // check for NaN
		if (x & 0x0003ffffffffffff) > 999999999999999 {
			x = x & 0xfe00000000000000 // clear G6-G12 and the payload bits
		} else {
			x = x & 0xfe03ffffffffffff // clear G6-G12
		}
		if (x & MASK_SNAN64) == MASK_SNAN64 { // SNaN
			// set invalid flag
			pfpsf |= BID_INVALID_EXCEPTION
			// return quiet (SNaN)
			res = x & 0xfdffffffffffffff
		} else { // QNaN
			res = x
		}
		return res, pfpsf
	} else if (x & MASK_INF64) == MASK_INF64 { // check for Infinity
		if (x & MASK_SIGN64) == MASK_SIGN64 { // x is -inf
			res = 0xf800000000000000
		} else { // x is +inf
			res = 0x77fb86f26fc0ffff // +MAXFP = +999...99 * 10^emax
		}
		return res, pfpsf
	}
	// unpack the argument
	x_sign = x & MASK_SIGN64 // 0 for positive, MASK_SIGN for negative
	// if steering bits are 11 (condition will be 0), then exponent is G[0:w+1] =>
	if (x & MASK_STEERING_BITS64) == MASK_STEERING_BITS64 {
		x_exp = (x & MASK_BINARY_EXPONENT2_64) >> 51 // biased
		C1 = (x & MASK_BINARY_SIG2_64) | MASK_BINARY_OR2_64
		if C1 > 9999999999999999 { // non-canonical
			x_exp = 0
			C1 = 0
		}
	} else {
		x_exp = (x & MASK_BINARY_EXPONENT1_64) >> 53 // biased
		C1 = x & MASK_BINARY_SIG1_64
	}

	// check for zeros (possibly from non-canonical values)
	if C1 == 0x0 { // x is 0
		res = 0x8000000000000001 // -MINFP = -1 * 10^emin
	} else { // x is not special and is not zero
		if x == 0xf7fb86f26fc0ffff {
			// x = -MAXFP = -999...99 * 10^emax
			res = 0xf800000000000000 // -inf
		} else if x == 0x0000000000000001 {
			// x = +MINFP = 1...99 * 10^emin
			res = 0x0000000000000000 // -0
		} else { // -MAXFP + 1ulp <= x <= -MINFP OR MINFP + 1 ulp <= x <= MAXFP
			// can add/subtract 1 ulp to the significand

			C.w[0] = C1
			C.w[1] = 0
			q1 = __get_dec_digits64(C)
			// if q1 < P16 then pad the significand with zeros
			if q1 < 16 {
				if x_exp > uint64(16-q1) {
					ind = 16 - q1 // 1 <= ind <= P16 - 1
					// pad with P16 - q1 zeros, until exponent = emin
					// C1 = C1 * 10^ind
					C1 = C1 * bid_ten2k64[ind]
					x_exp = x_exp - uint64(ind)
				} else { // pad with zeros until the exponent reaches emin
					ind = int(x_exp)
					C1 = C1 * bid_ten2k64[ind]
					x_exp = 0
				}
			}
			if x_sign != 0 { // x < 0
				// add 1 ulp (add 1 to the significand)
				C1++
				if C1 == 0x002386f26fc10000 { // if  C1 = 10^16
					C1 = 0x00038d7ea4c68000 // C1 = 10^15
					x_exp++
					// Ok, because -MAXFP = -999...99 * 10^emax was caught already
				}
			} else { // x > 0
				// subtract 1 ulp (subtract 1 from the significand)
				C1--
				if C1 == 0x00038d7ea4c67fff && x_exp != 0 { // if  C1 = 10^15 - 1
					C1 = 0x002386f26fc0ffff // C1 = 10^16 - 1
					x_exp--
				}
			}
			// assemble the result
			// if significand has 54 bits
			if (C1 & MASK_BINARY_OR2_64) != 0 {
				res = x_sign | (x_exp << 51) | MASK_STEERING_BITS64 | (C1 & MASK_BINARY_SIG2_64)
			} else { // significand fits in 53 bits
				res = x_sign | (x_exp << 53) | C1
			}
		}
	}
	return res, pfpsf
}

// Bid64NextAfter - Intel bid64_nextafter 기계적 포팅
func Bid64NextAfter(x, y uint64) (uint64, uint32) {
	var res uint64
	var tmp1, tmp2 uint64
	var tmp_fpsf uint32
	var pfpsf uint32
	var res1, res2 int

	// check for NaNs or infinities
	if ((x & MASK_INF64) == MASK_INF64) || ((y & MASK_INF64) == MASK_INF64) {
		// x is NaN or infinity or y is NaN or infinity

		if (x & MASK_NAN64) == MASK_NAN64 { // x is NAN
			if (x & 0x0003ffffffffffff) > 999999999999999 {
				x = x & 0xfe00000000000000 // clear G6-G12 and the payload bits
			} else {
				x = x & 0xfe03ffffffffffff // clear G6-G12
			}
			if (x & MASK_SNAN64) == MASK_SNAN64 { // x is SNAN
				// set invalid flag
				pfpsf |= BID_INVALID_EXCEPTION
				// return quiet (x)
				res = x & 0xfdffffffffffffff
			} else { // x is QNaN
				if (y & MASK_SNAN64) == MASK_SNAN64 { // y is SNAN
					// set invalid flag
					pfpsf |= BID_INVALID_EXCEPTION
				}
				// return x
				res = x
			}
			return res, pfpsf
		} else if (y & MASK_NAN64) == MASK_NAN64 { // y is NAN
			if (y & 0x0003ffffffffffff) > 999999999999999 {
				y = y & 0xfe00000000000000 // clear G6-G12 and the payload bits
			} else {
				y = y & 0xfe03ffffffffffff // clear G6-G12
			}
			if (y & MASK_SNAN64) == MASK_SNAN64 { // y is SNAN
				// set invalid flag
				pfpsf |= BID_INVALID_EXCEPTION
				// return quiet (y)
				res = y & 0xfdffffffffffffff
			} else { // y is QNaN
				// return y
				res = y
			}
			return res, pfpsf
		} else { // at least one is infinity
			if (x & MASK_INF64) == MASK_INF64 { // x = inf
				x = x & (MASK_SIGN64 | MASK_INF64)
			}
			if (y & MASK_INF64) == MASK_INF64 { // y = inf
				y = y & (MASK_SIGN64 | MASK_INF64)
			}
		}
	}
	// neither x nor y is NaN

	// if not infinity, check for non-canonical values x (treated as zero)
	if (x & MASK_INF64) != MASK_INF64 { // x != inf
		// unpack x
		if (x & MASK_STEERING_BITS64) == MASK_STEERING_BITS64 {
			// if the steering bits are 11 (condition will be 0), then
			// the exponent is G[0:w+1]
			if ((x & MASK_BINARY_SIG2_64) | MASK_BINARY_OR2_64) > 9999999999999999 {
				// non-canonical
				x = (x & MASK_SIGN64) | ((x & MASK_BINARY_EXPONENT2_64) << 2)
			}
		} else { // if ((x & MASK_STEERING_BITS) != MASK_STEERING_BITS) x is unch.
			// canonical
		}
	}
	// no need to check for non-canonical y

	// neither x nor y is NaN
	tmp_fpsf = pfpsf // save fpsf
	res1, _ = Bid64QuietEqual(x, y)
	res2, _ = Bid64QuietGreater(x, y)
	pfpsf = tmp_fpsf // restore fpsf
	if res1 != 0 {   // x = y
		// return x with the sign of y
		res = (y & 0x8000000000000000) | (x & 0x7fffffffffffffff)
	} else if res2 != 0 { // x > y
		res, tmp_fpsf = Bid64NextDown(x)
		pfpsf |= tmp_fpsf
	} else { // x < y
		res, tmp_fpsf = Bid64NextUp(x)
		pfpsf |= tmp_fpsf
	}
	// if the operand x is finite but the result is infinite, signal
	// overflow and inexact
	if ((x & MASK_INF64) != MASK_INF64) && ((res & MASK_INF64) == MASK_INF64) {
		// set the inexact flag
		pfpsf |= BID_INEXACT_EXCEPTION
		// set the overflow flag
		pfpsf |= BID_OVERFLOW_EXCEPTION
	}
	// if the result is in (-10^emin, 10^emin), and is different from the
	// operand x, signal underflow and inexact
	tmp1 = 0x00038d7ea4c68000 // +100...0[16] * 10^emin
	tmp2 = res & 0x7fffffffffffffff
	tmp_fpsf = pfpsf // save fpsf
	res1, _ = Bid64QuietGreater(tmp1, tmp2)
	res2, _ = Bid64QuietNotEqual(x, res)
	pfpsf = tmp_fpsf // restore fpsf
	if res1 != 0 && res2 != 0 {
		// set the inexact flag
		pfpsf |= BID_INEXACT_EXCEPTION
		// set the underflow flag
		pfpsf |= BID_UNDERFLOW_EXCEPTION
	}
	return res, pfpsf
}
