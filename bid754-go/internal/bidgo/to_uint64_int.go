package bidgo

import "math"

// Bid64ToUint64Int is ported mechanically from Intel bid64_to_uint64.c: bid64_to_uint64_int.
func Bid64ToUint64Int(x uint64) (uint64, uint32) {
	var res uint64
	var x_sign uint64
	var x_exp uint64
	var exp int // unbiased exponent
	// Note: C1 represents x_significand (BID_UINT64)
	var tmp1 uint64
	var x_nr_bits int
	var q, ind, shift int
	var C1 uint64
	var C BID_UINT128
	var Cstar uint64 // C* represents up to 16 decimal digits ~ 54 bits
	var P128 BID_UINT128
	var pfpsf uint32

	// check for NaN or Infinity
	if (x&MASK_NAN) == MASK_NAN || (x&MASK_INF) == MASK_INF {
		// set invalid flag
		pfpsf |= BID_INVALID_EXCEPTION
		// return Integer Indefinite
		res = 0x8000000000000000
		return res, pfpsf
	}
	// unpack x
	x_sign = x & MASK_SIGN // 0 for positive, MASK_SIGN for negative
	// if steering bits are 11 (condition will be 0), then exponent is G[0:w+1] =>
	if (x & MASK_STEERING_BITS) == MASK_STEERING_BITS {
		x_exp = (x & MASK_BINARY_EXPONENT2) >> 51 // biased
		C1 = (x & MASK_BINARY_SIG2) | MASK_BINARY_OR2
		if C1 > 9999999999999999 { // non-canonical
			x_exp = 0
			C1 = 0
		}
	} else {
		x_exp = (x & MASK_BINARY_EXPONENT1) >> 53 // biased
		C1 = x & MASK_BINARY_SIG1
	}

	// check for zeros (possibly from non-canonical values)
	if C1 == 0x0 {
		// x is 0
		res = 0x0000000000000000
		return res, pfpsf
	}
	// x is not special and is not zero

	// q = nr. of decimal digits in x (1 <= q <= 54)
	// determine first the nr. of bits in x
	if C1 >= 0x0020000000000000 { // x >= 2^53
		// split the 64-bit value in two 32-bit halves to avoid rounding errors
		tmp1 = math.Float64bits(float64(C1 >> 32)) // exact conversion
		x_nr_bits = 33 + int(((tmp1>>52)&0x7ff)-0x3ff)
	} else { // if x < 2^53
		tmp1 = math.Float64bits(float64(C1)) // exact conversion
		x_nr_bits = 1 + int(((tmp1>>52)&0x7ff)-0x3ff)
	}
	q = int(bid_nr_digits[x_nr_bits-1].digits)
	if q == 0 {
		q = int(bid_nr_digits[x_nr_bits-1].digits1)
		if C1 >= bid_nr_digits[x_nr_bits-1].threshold_lo {
			q++
		}
	}
	exp = int(x_exp) - 398 // unbiased exponent

	if (q + exp) > 20 { // x >= 10^20 ~= 2^66.45... (cannot fit in 64 bits)
		// set invalid flag
		pfpsf |= BID_INVALID_EXCEPTION
		// return Integer Indefinite
		res = 0x8000000000000000
		return res, pfpsf
	} else if (q + exp) == 20 { // x = c(0)c(1)...c(19).c(20)...c(q-1)
		// in this case 2^63.11... ~= 10^19 <= x < 10^20 ~= 2^66.43...
		// so x rounded to an integer may or may not fit in an unsigned 64-bit int
		// the cases that do not fit are identified here; the ones that fit
		// fall through and will be handled with other cases further,
		// under '1 <= q + exp <= 20'
		if x_sign != 0 { // if n < 0 and q + exp = 20 then x is much less than -1
			// => set invalid flag
			pfpsf |= BID_INVALID_EXCEPTION
			// return Integer Indefinite
			res = 0x8000000000000000
			return res, pfpsf
		} else { // if n > 0 and q + exp = 20
			// if n >= 2^64 then n is too large
			// <=> c(0)c(1)...c(19).c(20)...c(q-1) >= 2^64
			// <=> 0.c(0)c(1)...c(19)c(20)...c(q-1) * 10^20 >= 2^64
			// <=> 0.c(0)c(1)...c(19)c(20)...c(q-1) * 10^21 >= 5*(2^65)
			// <=> C * 10^(21-q) >= 0xa0000000000000000, 1<=q<=16
			if q == 1 {
				// C * 10^20 >= 0xa0000000000000000
				C = __mul_128x64_to_128(C1, bid_ten2k128[0]) // 10^20 * C
				if C.w[1] >= 0x0a {
					// actually C.w[1] == 0x0a && C.w[0] >= 0x0000000000000000ull) {
					// set invalid flag
					pfpsf |= BID_INVALID_EXCEPTION
					// return Integer Indefinite
					res = 0x8000000000000000
					return res, pfpsf
				}
				// else cases that can be rounded to a 64-bit int fall through
				// to '1 <= q + exp <= 20'
			} else { // if (2 <= q <= 16) => 5 <= 21 - q <= 19
				// Note: C * 10^(21-q) has 20 or 21 digits; 0xa0000000000000000
				// has 21 digits
				C = __mul_64x64_to_128(C1, bid_ten2k64[21-q])
				if C.w[1] >= 0x0a {
					// actually C.w[1] == 0x0a && C.w[0] >= 0x0000000000000000ull) {
					// set invalid flag
					pfpsf |= BID_INVALID_EXCEPTION
					// return Integer Indefinite
					res = 0x8000000000000000
					return res, pfpsf
				}
				// else cases that can be rounded to a 64-bit int fall through
				// to '1 <= q + exp <= 20'
			}
		}
	}
	// n is not too large to be converted to int64 if -1 < n < 2^64
	// Note: some of the cases tested for above fall through to this point
	if (q + exp) <= 0 { // n = +/-0.[0...0]c(0)c(1)...c(q-1)
		// return 0
		res = 0x0000000000000000
		return res, pfpsf
	} else { // if (1 <= q + exp <= 20, 1 <= q <= 16, -15 <= exp <= 19)
		// x <= -1 or 1 <= x < 2^64 so if positive x can be rounded
		// to nearest to a 64-bit unsigned signed integer
		if x_sign != 0 { // x <= -1
			// set invalid flag
			pfpsf |= BID_INVALID_EXCEPTION
			// return Integer Indefinite
			res = 0x8000000000000000
			return res, pfpsf
		}
		// 1 <= x < 2^64 so x can be rounded
		// to nearest to a 64-bit unsigned integer
		if exp < 0 { // 2 <= q <= 16, -15 <= exp <= -1, 1 <= q + exp <= 20
			ind = -exp // 1 <= ind <= 15; ind is a synonym for 'x'
			// chop off ind digits from the lower part of C1
			// C1 fits in 64 bits
			// calculate C* and f*
			// C* is actually floor(C*) in this case
			// C* and f* need shifting and masking, as shown by
			// bid_shiftright128[] and bid_maskhigh128[]
			// 1 <= x <= 15
			// kx = 10^(-x) = bid_ten2mk64[ind - 1]
			// C* = C1 * 10^(-x)
			// the approximation of 10^(-x) was rounded up to 54 bits
			P128 = __mul_64x64_to_128(C1, bid_ten2mk64[ind-1])
			Cstar = P128.w[1]
			// the top Ex bits of 10^(-x) are T* = bid_ten2mk128trunc[ind].w[0], e.g.
			// if x=1, T*=bid_ten2mk128trunc[0].w[0]=0x1999999999999999
			// C* = floor(C*) (logical right shift; C has p decimal digits,
			//     correct by Property 1)
			// n = C* * 10^(e+x)

			// shift right C* by Ex-64 = bid_shiftright128[ind]
			shift = bid_shiftright128[ind-1] // 0 <= shift <= 39
			Cstar = Cstar >> shift
			res = Cstar // the result is positive
		} else if exp == 0 {
			// 1 <= q <= 10
			// res = +C (exact)
			res = C1 // the result is positive
		} else { // if (exp > 0) => 1 <= exp <= 9, 1 <= q < 9, 2 <= q + exp <= 10
			// res = +C * 10^exp (exact)
			res = C1 * bid_ten2k64[exp] // the result is positive
		}
	}
	return res, pfpsf
}

// Bid64ToUint64Xint is ported mechanically from Intel bid64_to_uint64.c: bid64_to_uint64_xint.
func Bid64ToUint64Xint(x uint64) (uint64, uint32) {
	var res uint64
	var x_sign uint64
	var x_exp uint64
	var exp int // unbiased exponent
	// Note: C1 represents x_significand (BID_UINT64)
	var tmp1 uint64
	var x_nr_bits int
	var q, ind, shift int
	var C1 uint64
	var C BID_UINT128
	var Cstar uint64 // C* represents up to 16 decimal digits ~ 54 bits
	var fstar BID_UINT128
	var P128 BID_UINT128
	var pfpsf uint32

	// check for NaN or Infinity
	if (x&MASK_NAN) == MASK_NAN || (x&MASK_INF) == MASK_INF {
		// set invalid flag
		pfpsf |= BID_INVALID_EXCEPTION
		// return Integer Indefinite
		res = 0x8000000000000000
		return res, pfpsf
	}
	// unpack x
	x_sign = x & MASK_SIGN // 0 for positive, MASK_SIGN for negative
	// if steering bits are 11 (condition will be 0), then exponent is G[0:w+1] =>
	if (x & MASK_STEERING_BITS) == MASK_STEERING_BITS {
		x_exp = (x & MASK_BINARY_EXPONENT2) >> 51 // biased
		C1 = (x & MASK_BINARY_SIG2) | MASK_BINARY_OR2
		if C1 > 9999999999999999 { // non-canonical
			x_exp = 0
			C1 = 0
		}
	} else {
		x_exp = (x & MASK_BINARY_EXPONENT1) >> 53 // biased
		C1 = x & MASK_BINARY_SIG1
	}

	// check for zeros (possibly from non-canonical values)
	if C1 == 0x0 {
		// x is 0
		res = 0x0000000000000000
		return res, pfpsf
	}
	// x is not special and is not zero

	// q = nr. of decimal digits in x (1 <= q <= 54)
	// determine first the nr. of bits in x
	if C1 >= 0x0020000000000000 { // x >= 2^53
		// split the 64-bit value in two 32-bit halves to avoid rounding errors
		tmp1 = math.Float64bits(float64(C1 >> 32)) // exact conversion
		x_nr_bits = 33 + int(((tmp1>>52)&0x7ff)-0x3ff)
	} else { // if x < 2^53
		tmp1 = math.Float64bits(float64(C1)) // exact conversion
		x_nr_bits = 1 + int(((tmp1>>52)&0x7ff)-0x3ff)
	}
	q = int(bid_nr_digits[x_nr_bits-1].digits)
	if q == 0 {
		q = int(bid_nr_digits[x_nr_bits-1].digits1)
		if C1 >= bid_nr_digits[x_nr_bits-1].threshold_lo {
			q++
		}
	}
	exp = int(x_exp) - 398 // unbiased exponent

	if (q + exp) > 20 { // x >= 10^20 ~= 2^66.45... (cannot fit in 64 bits)
		// set invalid flag
		pfpsf |= BID_INVALID_EXCEPTION
		// return Integer Indefinite
		res = 0x8000000000000000
		return res, pfpsf
	} else if (q + exp) == 20 { // x = c(0)c(1)...c(19).c(20)...c(q-1)
		// in this case 2^63.11... ~= 10^19 <= x < 10^20 ~= 2^66.43...
		// so x rounded to an integer may or may not fit in an unsigned 64-bit int
		// the cases that do not fit are identified here; the ones that fit
		// fall through and will be handled with other cases further,
		// under '1 <= q + exp <= 20'
		if x_sign != 0 { // if n < 0 and q + exp = 20 then x is much less than -1
			// => set invalid flag
			pfpsf |= BID_INVALID_EXCEPTION
			// return Integer Indefinite
			res = 0x8000000000000000
			return res, pfpsf
		} else { // if n > 0 and q + exp = 20
			// if n >= 2^64 then n is too large
			// <=> c(0)c(1)...c(19).c(20)...c(q-1) >= 2^64
			// <=> 0.c(0)c(1)...c(19)c(20)...c(q-1) * 10^20 >= 2^64
			// <=> 0.c(0)c(1)...c(19)c(20)...c(q-1) * 10^21 >= 5*(2^65)
			// <=> C * 10^(21-q) >= 0xa0000000000000000, 1<=q<=16
			if q == 1 {
				// C * 10^20 >= 0xa0000000000000000
				C = __mul_128x64_to_128(C1, bid_ten2k128[0]) // 10^20 * C
				if C.w[1] >= 0x0a {
					// actually C.w[1] == 0x0a && C.w[0] >= 0x0000000000000000ull) {
					// set invalid flag
					pfpsf |= BID_INVALID_EXCEPTION
					// return Integer Indefinite
					res = 0x8000000000000000
					return res, pfpsf
				}
				// else cases that can be rounded to a 64-bit int fall through
				// to '1 <= q + exp <= 20'
			} else { // if (2 <= q <= 16) => 5 <= 21 - q <= 19
				// Note: C * 10^(21-q) has 20 or 21 digits; 0xa0000000000000000
				// has 21 digits
				C = __mul_64x64_to_128(C1, bid_ten2k64[21-q])
				if C.w[1] >= 0x0a {
					// actually C.w[1] == 0x0a && C.w[0] >= 0x0000000000000000ull) {
					// set invalid flag
					pfpsf |= BID_INVALID_EXCEPTION
					// return Integer Indefinite
					res = 0x8000000000000000
					return res, pfpsf
				}
				// else cases that can be rounded to a 64-bit int fall through
				// to '1 <= q + exp <= 20'
			}
		}
	}
	// n is not too large to be converted to int64 if -1 < n < 2^64
	// Note: some of the cases tested for above fall through to this point
	if (q + exp) <= 0 { // n = +/-0.[0...0]c(0)c(1)...c(q-1)
		// set inexact flag
		pfpsf |= BID_INEXACT_EXCEPTION
		// return 0
		res = 0x0000000000000000
		return res, pfpsf
	} else { // if (1 <= q + exp <= 20, 1 <= q <= 16, -15 <= exp <= 19)
		// x <= -1 or 1 <= x < 2^64 so if positive x can be rounded
		// to nearest to a 64-bit unsigned signed integer
		if x_sign != 0 { // x <= -1
			// set invalid flag
			pfpsf |= BID_INVALID_EXCEPTION
			// return Integer Indefinite
			res = 0x8000000000000000
			return res, pfpsf
		}
		// 1 <= x < 2^64 so x can be rounded
		// to nearest to a 64-bit unsigned integer
		if exp < 0 { // 2 <= q <= 16, -15 <= exp <= -1, 1 <= q + exp <= 20
			ind = -exp // 1 <= ind <= 15; ind is a synonym for 'x'
			// chop off ind digits from the lower part of C1
			// C1 fits in 64 bits
			// calculate C* and f*
			// C* is actually floor(C*) in this case
			// C* and f* need shifting and masking, as shown by
			// bid_shiftright128[] and bid_maskhigh128[]
			// 1 <= x <= 15
			// kx = 10^(-x) = bid_ten2mk64[ind - 1]
			// C* = C1 * 10^(-x)
			// the approximation of 10^(-x) was rounded up to 54 bits
			P128 = __mul_64x64_to_128(C1, bid_ten2mk64[ind-1])
			Cstar = P128.w[1]
			fstar.w[1] = P128.w[1] & bid_maskhigh128[ind-1]
			fstar.w[0] = P128.w[0]
			// the top Ex bits of 10^(-x) are T* = bid_ten2mk128trunc[ind].w[0], e.g.
			// if x=1, T*=bid_ten2mk128trunc[0].w[0]=0x1999999999999999
			// C* = floor(C*) (logical right shift; C has p decimal digits,
			//     correct by Property 1)
			// n = C* * 10^(e+x)

			// shift right C* by Ex-64 = bid_shiftright128[ind]
			shift = bid_shiftright128[ind-1] // 0 <= shift <= 39
			Cstar = Cstar >> shift
			// determine inexactness of the rounding of C*
			// if (0 < f* < 10^(-x)) then
			//   the result is exact
			// else // if (f* > T*) then
			//   the result is inexact
			if ind-1 <= 2 { // fstar.w[1] is 0
				if fstar.w[0] > bid_ten2mk128trunc[ind-1].w[1] {
					// bid_ten2mk128trunc[ind -1].w[1] is identical to
					// bid_ten2mk128[ind -1].w[1]
					// set the inexact flag
					pfpsf |= BID_INEXACT_EXCEPTION
				}
			} else { // if 3 <= ind - 1 <= 14
				if fstar.w[1] != 0 || fstar.w[0] > bid_ten2mk128trunc[ind-1].w[1] {
					// bid_ten2mk128trunc[ind -1].w[1] is identical to
					// bid_ten2mk128[ind -1].w[1]
					// set the inexact flag
					pfpsf |= BID_INEXACT_EXCEPTION
				}
			}

			res = Cstar // the result is positive
		} else if exp == 0 {
			// 1 <= q <= 10
			// res = +C (exact)
			res = C1 // the result is positive
		} else { // if (exp > 0) => 1 <= exp <= 9, 1 <= q < 9, 2 <= q + exp <= 10
			// res = +C * 10^exp (exact)
			res = C1 * bid_ten2k64[exp] // the result is positive
		}
	}
	return res, pfpsf
}
