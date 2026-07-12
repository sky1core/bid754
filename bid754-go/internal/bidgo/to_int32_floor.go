package bidgo

import "math"

// Bid64ToInt32Floor is ported mechanically from Intel bid64_to_int32.c: bid64_to_int32_floor.
func Bid64ToInt32Floor(x uint64) (int32, uint32) {
	var res int32
	var x_sign uint64
	var x_exp uint64
	var exp int // unbiased exponent
	// Note: C1 represents x_significand (BID_UINT64)
	var tmp64 uint64
	var tmp1 uint64
	var x_nr_bits int
	var q, ind, shift int
	var C1 uint64
	var Cstar uint64 // C* represents up to 16 decimal digits ~ 54 bits
	var fstar BID_UINT128
	var P128 BID_UINT128
	var pfpsf uint32

	// check for NaN or Infinity
	if (x&MASK_NAN) == MASK_NAN || (x&MASK_INF) == MASK_INF {
		// set invalid flag
		pfpsf |= BID_INVALID_EXCEPTION
		// return Integer Indefinite
		res = -0x80000000
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
		res = 0x00000000
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

	if (q + exp) > 10 { // x >= 10^10 ~= 2^33.2... (cannot fit in 32 bits)
		// set invalid flag
		pfpsf |= BID_INVALID_EXCEPTION
		// return Integer Indefinite
		res = -0x80000000
		return res, pfpsf
	} else if (q + exp) == 10 { // x = c(0)c(1)...c(9).c(10)...c(q-1)
		// in this case 2^29.89... ~= 10^9 <= x < 10^10 ~= 2^33.2...
		// so x rounded to an integer may or may not fit in a signed 32-bit int
		// the cases that do not fit are identified here; the ones that fit
		// fall through and will be handled with other cases further,
		// under '1 <= q + exp <= 10'
		if x_sign != 0 { // if n < 0 and q + exp = 10
			// if n < -2^31 then n is too large
			// too large if c(0)c(1)...c(9).c(10)...c(q-1) > 2^31
			// <=> 0.c(0)c(1)...c(q-1) * 10^11 > 0x500000000, 1<=q<=16
			// <=> C * 10^(11-q) >= 0x500000000, 1<=q<=16
			if q <= 11 {
				// Note: C * 10^(11-q) has 10 or 11 digits; 0x500000000 has 11 digits
				tmp64 = C1 * bid_ten2k64[11-q] // C scaled up to 11-digit int
				// c(0)c(1)...c(9)c(10) or c(0)c(1)...c(q-1)0...0 (11 digits)
				if tmp64 > 0x500000000 {
					// set invalid flag
					pfpsf |= BID_INVALID_EXCEPTION
					// return Integer Indefinite
					res = -0x80000000
					return res, pfpsf
				}
			} else { // if (q > 11), i.e. 12 <= q <= 16 and so -15 <= exp <= -2
				// C * 10^(11-q) > 0x500000000 <=>
				// C > 0x500000000 * 10^(q-11) where 1 <= q - 11 <= 5
				// (scale 2^31+1 up)
				// Note: 0x500000000*10^(q-11) has q-1 or q digits, where q <= 16
				tmp64 = 0x500000000 * bid_ten2k64[q-11]
				if C1 > tmp64 {
					// set invalid flag
					pfpsf |= BID_INVALID_EXCEPTION
					// return Integer Indefinite
					res = -0x80000000
					return res, pfpsf
				}
			}
		} else { // if n > 0 and q + exp = 10
			// if n >= 2^31 then n is too large
			// too large if c(0)c(1)...c(9).c(10)...c(q-1) >= 2^31
			// <=> 0.c(0)c(1)...c(q-1) * 10^11 >= 0x500000000, 1<=q<=16
			// <=> C * 10^(11-q) >= 0x500000000, 1<=q<=16
			if q <= 11 {
				// Note: C * 10^(11-q) has 10 or 11 digits; 0x500000000 has 11 digits
				tmp64 = C1 * bid_ten2k64[11-q] // C scaled up to 11-digit int
				// c(0)c(1)...c(9)c(10) or c(0)c(1)...c(q-1)0...0 (11 digits)
				if tmp64 >= 0x500000000 {
					// set invalid flag
					pfpsf |= BID_INVALID_EXCEPTION
					// return Integer Indefinite
					res = -0x80000000
					return res, pfpsf
				}
			} else { // if (q > 11), i.e. 12 <= q <= 16 and so -15 <= exp <= -2
				// C * 10^(11-q) >= 0x500000000 <=>
				// C >= 0x500000000 * 10^(q-11) where 1 <= q - 11 <= 5
				// (scale 2^31-1 up)
				// Note: 0x500000000*10^(q-11) has q-1 or q digits, where q <= 16
				tmp64 = 0x500000000 * bid_ten2k64[q-11]
				if C1 >= tmp64 {
					// set invalid flag
					pfpsf |= BID_INVALID_EXCEPTION
					// return Integer Indefinite
					res = -0x80000000
					return res, pfpsf
				}
			}
		}
	}
	// n is not too large to be converted to int32: -2^31 <= n < 2^31
	// Note: some of the cases tested for above fall through to this point
	if (q + exp) <= 0 { // n = +/-0.[0...0]c(0)c(1)...c(q-1)
		// return -1 or 0
		if x_sign != 0 {
			res = -1
		} else {
			res = 0x00000000
		}
		return res, pfpsf
	} else { // if (1 <= q + exp <= 10, 1 <= q <= 16, -15 <= exp <= 9)
		// -2^31-1 < x <= -1 or 1 <= x < 2^31 so x can be rounded
		// to nearest to a 32-bit signed integer
		if exp < 0 { // 2 <= q <= 16, -15 <= exp <= -1, 1 <= q + exp <= 10
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
			// correct by Property 1)
			// n = C* * 10^(e+x)

			// shift right C* by Ex-64 = bid_shiftright128[ind]
			shift = bid_shiftright128[ind-1] // 0 <= shift <= 39
			Cstar = Cstar >> shift
			// determine inexactness of the rounding of C*
			// if (0 < f* < 10^(-x)) then
			//   the result is exact
			// else // if (f* > T*) then
			//   the result is inexact
			if (ind - 1) <= 2 {
				if fstar.w[0] > bid_ten2mk128trunc[ind-1].w[1] {
					// bid_ten2mk128trunc[ind -1].w[1] is identical to
					// bid_ten2mk128[ind -1].w[1]
					if x_sign != 0 { // negative and inexact
						Cstar++
					}
				}
			} else { // if 3 <= ind - 1 <= 14
				if fstar.w[1] != 0 || fstar.w[0] > bid_ten2mk128trunc[ind-1].w[1] {
					// bid_ten2mk128trunc[ind -1].w[1] is identical to
					// bid_ten2mk128[ind -1].w[1]
					if x_sign != 0 { // negative and inexact
						Cstar++
					}
				}
			}

			if x_sign != 0 {
				res = -int32(Cstar)
			} else {
				res = int32(Cstar)
			}
		} else if exp == 0 {
			// 1 <= q <= 10
			// res = +/-C (exact)
			if x_sign != 0 {
				res = -int32(C1)
			} else {
				res = int32(C1)
			}
		} else { // if (exp > 0) => 1 <= exp <= 9, 1 <= q < 9, 2 <= q + exp <= 10
			// res = +/-C * 10^exp (exact)
			if x_sign != 0 {
				res = -int32(C1 * bid_ten2k64[exp])
			} else {
				res = int32(C1 * bid_ten2k64[exp])
			}
		}
	}
	return res, pfpsf
}

// Bid64ToInt32Xfloor is ported mechanically from Intel bid64_to_int32.c: bid64_to_int32_xfloor.
func Bid64ToInt32Xfloor(x uint64) (int32, uint32) {
	var res int32
	var x_sign uint64
	var x_exp uint64
	var exp int // unbiased exponent
	// Note: C1 represents x_significand (BID_UINT64)
	var tmp64 uint64
	var tmp1 uint64
	var x_nr_bits int
	var q, ind, shift int
	var C1 uint64
	var Cstar uint64 // C* represents up to 16 decimal digits ~ 54 bits
	var fstar BID_UINT128
	var P128 BID_UINT128
	var pfpsf uint32

	// check for NaN or Infinity
	if (x&MASK_NAN) == MASK_NAN || (x&MASK_INF) == MASK_INF {
		// set invalid flag
		pfpsf |= BID_INVALID_EXCEPTION
		// return Integer Indefinite
		res = -0x80000000
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
		res = 0x00000000
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

	if (q + exp) > 10 { // x >= 10^10 ~= 2^33.2... (cannot fit in 32 bits)
		// set invalid flag
		pfpsf |= BID_INVALID_EXCEPTION
		// return Integer Indefinite
		res = -0x80000000
		return res, pfpsf
	} else if (q + exp) == 10 { // x = c(0)c(1)...c(9).c(10)...c(q-1)
		// in this case 2^29.89... ~= 10^9 <= x < 10^10 ~= 2^33.2...
		// so x rounded to an integer may or may not fit in a signed 32-bit int
		// the cases that do not fit are identified here; the ones that fit
		// fall through and will be handled with other cases further,
		// under '1 <= q + exp <= 10'
		if x_sign != 0 { // if n < 0 and q + exp = 10
			// if n < -2^31 then n is too large
			// too large if c(0)c(1)...c(9).c(10)...c(q-1) > 2^31
			// <=> 0.c(0)c(1)...c(q-1) * 10^11 > 0x500000000, 1<=q<=16
			// <=> C * 10^(11-q) >= 0x500000000, 1<=q<=16
			if q <= 11 {
				// Note: C * 10^(11-q) has 10 or 11 digits; 0x500000000 has 11 digits
				tmp64 = C1 * bid_ten2k64[11-q] // C scaled up to 11-digit int
				// c(0)c(1)...c(9)c(10) or c(0)c(1)...c(q-1)0...0 (11 digits)
				if tmp64 > 0x500000000 {
					// set invalid flag
					pfpsf |= BID_INVALID_EXCEPTION
					// return Integer Indefinite
					res = -0x80000000
					return res, pfpsf
				}
			} else { // if (q > 11), i.e. 12 <= q <= 16 and so -15 <= exp <= -2
				// C * 10^(11-q) > 0x500000000 <=>
				// C > 0x500000000 * 10^(q-11) where 1 <= q - 11 <= 5
				// (scale 2^31+1 up)
				// Note: 0x500000000*10^(q-11) has q-1 or q digits, where q <= 16
				tmp64 = 0x500000000 * bid_ten2k64[q-11]
				if C1 > tmp64 {
					// set invalid flag
					pfpsf |= BID_INVALID_EXCEPTION
					// return Integer Indefinite
					res = -0x80000000
					return res, pfpsf
				}
			}
		} else { // if n > 0 and q + exp = 10
			// if n >= 2^31 then n is too large
			// too large if c(0)c(1)...c(9).c(10)...c(q-1) >= 2^31
			// <=> 0.c(0)c(1)...c(q-1) * 10^11 >= 0x500000000, 1<=q<=16
			// <=> C * 10^(11-q) >= 0x500000000, 1<=q<=16
			if q <= 11 {
				// Note: C * 10^(11-q) has 10 or 11 digits; 0x500000000 has 11 digits
				tmp64 = C1 * bid_ten2k64[11-q] // C scaled up to 11-digit int
				// c(0)c(1)...c(9)c(10) or c(0)c(1)...c(q-1)0...0 (11 digits)
				if tmp64 >= 0x500000000 {
					// set invalid flag
					pfpsf |= BID_INVALID_EXCEPTION
					// return Integer Indefinite
					res = -0x80000000
					return res, pfpsf
				}
			} else { // if (q > 11), i.e. 12 <= q <= 16 and so -15 <= exp <= -2
				// C * 10^(11-q) >= 0x500000000 <=>
				// C >= 0x500000000 * 10^(q-11) where 1 <= q - 11 <= 5
				// (scale 2^31-1 up)
				// Note: 0x500000000*10^(q-11) has q-1 or q digits, where q <= 16
				tmp64 = 0x500000000 * bid_ten2k64[q-11]
				if C1 >= tmp64 {
					// set invalid flag
					pfpsf |= BID_INVALID_EXCEPTION
					// return Integer Indefinite
					res = -0x80000000
					return res, pfpsf
				}
			}
		}
	}
	// n is not too large to be converted to int32: -2^31 <= n < 2^31
	// Note: some of the cases tested for above fall through to this point
	if (q + exp) <= 0 { // n = +/-0.[0...0]c(0)c(1)...c(q-1)
		// set inexact flag
		pfpsf |= BID_INEXACT_EXCEPTION
		// return -1 or 0
		if x_sign != 0 {
			res = -1
		} else {
			res = 0x00000000
		}
		return res, pfpsf
	} else { // if (1 <= q + exp <= 10, 1 <= q <= 16, -15 <= exp <= 9)
		// -2^31-1 < x <= -1 or 1 <= x < 2^31 so x can be rounded
		// to nearest to a 32-bit signed integer
		if exp < 0 { // 2 <= q <= 16, -15 <= exp <= -1, 1 <= q + exp <= 10
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
			// correct by Property 1)
			// n = C* * 10^(e+x)

			// shift right C* by Ex-64 = bid_shiftright128[ind]
			shift = bid_shiftright128[ind-1] // 0 <= shift <= 39
			Cstar = Cstar >> shift
			// determine inexactness of the rounding of C*
			// if (0 < f* < 10^(-x)) then
			//   the result is exact
			// else // if (f* > T*) then
			//   the result is inexact
			if (ind - 1) <= 2 {
				if fstar.w[0] > bid_ten2mk128trunc[ind-1].w[1] {
					// bid_ten2mk128trunc[ind -1].w[1] is identical to
					// bid_ten2mk128[ind -1].w[1]
					if x_sign != 0 { // negative and inexact
						Cstar++
					}
					// set the inexact flag
					pfpsf |= BID_INEXACT_EXCEPTION
				}
			} else { // if 3 <= ind - 1 <= 14
				if fstar.w[1] != 0 || fstar.w[0] > bid_ten2mk128trunc[ind-1].w[1] {
					// bid_ten2mk128trunc[ind -1].w[1] is identical to
					// bid_ten2mk128[ind -1].w[1]
					if x_sign != 0 { // negative and inexact
						Cstar++
					}
					// set the inexact flag
					pfpsf |= BID_INEXACT_EXCEPTION
				}
			}

			if x_sign != 0 {
				res = -int32(Cstar)
			} else {
				res = int32(Cstar)
			}
		} else if exp == 0 {
			// 1 <= q <= 10
			// res = +/-C (exact)
			if x_sign != 0 {
				res = -int32(C1)
			} else {
				res = int32(C1)
			}
		} else { // if (exp > 0) => 1 <= exp <= 9, 1 <= q < 9, 2 <= q + exp <= 10
			// res = +/-C * 10^exp (exact)
			if x_sign != 0 {
				res = -int32(C1 * bid_ten2k64[exp])
			} else {
				res = int32(C1 * bid_ten2k64[exp])
			}
		}
	}
	return res, pfpsf
}
