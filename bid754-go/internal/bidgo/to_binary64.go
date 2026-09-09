// Ported from: IntelRDFPMathLib20U4/LIBRARY/src/bid_binarydecimal.c
// Version: Intel(R) Decimal Floating-Point Math Library 2.0 Update 4

package bidgo

func unpack_bid64_binarydecimal(x uint64) (s int, e int, k int, c uint64, isZero bool, isInf bool, isNaN bool, nanPayloadHi uint64, isSNaN bool) {
	s = int(x >> 63)
	if (x & (3 << 61)) == (3 << 61) {
		if (x & (0xf << 59)) == (0xf << 59) {
			if (x & (0x1f << 58)) != (0x1f << 58) {
				isInf = true
				return
			}
			isSNaN = (x & (1 << 57)) != 0
			isNaN = true
			if (x & 0x3ffffffffffff) <= 999999999999999 {
				nanPayloadHi = x << 14
			}
			return
		}
		e = int((x>>51)&((1<<10)-1)) - 398
		c = (1 << 53) + (x & ((1 << 51) - 1))
		if c > 9999999999999999 {
			isZero = true
		}
		return
	}
	e = int((x>>53)&((1<<10)-1)) - 398
	c = x & ((1 << 53) - 1)
	if c == 0 {
		isZero = true
		return
	}
	k = clz64_nz(c) - 10
	c <<= uint(k)
	return
}

func Bid64ToBinary32(x uint64, rnd_mode int) (uint32, uint32) {
	if rnd_mode < BID_ROUNDING_TO_NEAREST || rnd_mode > BID_ROUNDING_TIES_AWAY {
		return 0x7fc00000, BID_INVALID_EXCEPTION
	}
	var flags uint32
	var c_prov uint64
	var c BID_UINT128
	var m_min BID_UINT128
	var e_out int
	var r BID_UINT256
	var z BID_UINT384

	s, e, k, coeff, isZero, isInf, isNaN, nanPayloadHi, isSNaN :=
		unpack_bid64_binarydecimal(x)

	if isZero {
		return (uint32(s) << 31) + (uint32(0) << 23) + uint32(0), flags
	}
	if isInf {
		return (uint32(s) << 31) + (uint32(255) << 23) + uint32(0), flags
	}
	if isNaN {
		if isSNaN {
			flags |= BID_INVALID_EXCEPTION
		}
		return (uint32(s) << 31) + (uint32(255) << 23) + uint32((nanPayloadHi>>42)+(1<<22)), flags
	}

	c.lo = coeff
	c.hi, c.lo = sll128_short(c.hi, c.lo, 59)
	k += 59

	if e >= 39 {
		flags |= BID_OVERFLOW_EXCEPTION | BID_INEXACT_EXCEPTION
		if (rnd_mode == BID_ROUNDING_TO_ZERO) ||
			(rnd_mode == boolToRndMode(s != 0)) {
			return (uint32(s) << 31) + (uint32(254) << 23) + uint32((1<<23)-1), flags
		}
		return (uint32(s) << 31) + (uint32(255) << 23) + uint32(0), flags
	}

	if e <= -80 {
		e = -80
	}

	m_min = bid_breakpoints_binary32[e+80]
	e_out = bid_exponents_binary32[e+80] - k

	if le128(c.hi, c.lo, m_min.hi, m_min.lo) {
		r = bid_multipliers1_binary32[e+80]
	} else {
		r = bid_multipliers2_binary32[e+80]
		e_out = e_out + 1
	}

	z = __mul_128x256_to_384(c, r)

	if e_out < 1 {
		d := 1 - e_out
		if d > 26 {
			d = 26
		}
		e_out = 1
		z.w5, z.w4, z.w3, z.w2 = srl256_short(z.w5, z.w4, z.w3, z.w2, uint(d))
	}
	c_prov = z.w5

	rbIdx := (rnd_mode << 2) + ((s & 1) << 1) + int(c_prov&1)
	if lt128(
		bid_roundbound_128[rbIdx].hi,
		bid_roundbound_128[rbIdx].lo,
		z.w4, z.w3) {
		c_prov = c_prov + 1
		if c_prov == (1 << 24) {
			c_prov = 1 << 23
			e_out = e_out + 1
		} else if (c_prov == (1 << 23)) && (e_out == 1) {
			if (((rnd_mode & 3) == 0) && (z.w4 < (3 << 62))) ||
				((rnd_mode+int(s&1) == 2) && (z.w4 < (1 << 63))) {
				flags |= BID_UNDERFLOW_EXCEPTION
			}
		}
	}

	if e_out >= 255 {
		flags |= BID_OVERFLOW_EXCEPTION | BID_INEXACT_EXCEPTION
		if (rnd_mode == BID_ROUNDING_TO_ZERO) ||
			(rnd_mode == boolToRndMode(s != 0)) {
			return (uint32(s) << 31) + (uint32(254) << 23) + uint32((1<<23)-1), flags
		}
		return (uint32(s) << 31) + (uint32(255) << 23) + uint32(0), flags
	}

	if c_prov < (1 << 23) {
		e_out = 0
	} else {
		c_prov = c_prov & ((1 << 23) - 1)
	}

	if (z.w4 != 0) || (z.w3 != 0) {
		flags |= BID_INEXACT_EXCEPTION
		if e_out == 0 {
			flags |= BID_UNDERFLOW_EXCEPTION
		}
	}

	return (uint32(s) << 31) + (uint32(e_out) << 23) + uint32(c_prov), flags
}

func Bid64ToBinary64(x uint64, rnd_mode int) (uint64, uint32) {
	if rnd_mode < BID_ROUNDING_TO_NEAREST || rnd_mode > BID_ROUNDING_TIES_AWAY {
		return 0x7ff8000000000000, BID_INVALID_EXCEPTION
	}
	var flags uint32
	var c_prov uint64
	var c BID_UINT128
	var m_min BID_UINT128
	var e_out int
	var r BID_UINT256
	var z BID_UINT384

	s, e, k, coeff, isZero, isInf, isNaN, nanPayloadHi, isSNaN :=
		unpack_bid64_binarydecimal(x)

	if isZero {
		return (uint64(s) << 63) + (uint64(0) << 52) + uint64(0), flags
	}
	if isInf {
		return (uint64(s) << 63) + (uint64(2047) << 52) + uint64(0), flags
	}
	if isNaN {
		if isSNaN {
			flags |= BID_INVALID_EXCEPTION
		}
		return (uint64(s) << 63) + (uint64(2047) << 52) + uint64((nanPayloadHi>>13)+(1<<51)), flags
	}

	c.hi = coeff << 1
	c.lo = 0
	k += 59

	if e >= 309 {
		flags |= BID_OVERFLOW_EXCEPTION | BID_INEXACT_EXCEPTION
		if (rnd_mode == BID_ROUNDING_TO_ZERO) ||
			(rnd_mode == boolToRndMode(s != 0)) {
			return (uint64(s) << 63) + (uint64(2046) << 52) + uint64((1<<52)-1), flags
		}
		return (uint64(s) << 63) + (uint64(2047) << 52) + uint64(0), flags
	}

	if e <= -358 {
		e = -358
	}

	m_min = bid_breakpoints_binary64[e+358]
	e_out = bid_exponents_binary64[e+358] - k

	if le128(c.hi, c.lo, m_min.hi, m_min.lo) {
		r = bid_multipliers1_binary64[e+358]
	} else {
		r = bid_multipliers2_binary64[e+358]
		e_out = e_out + 1
	}

	product := __mul_64x256_to_320(c.hi, r)
	z = BID_UINT384{w0: 0, w1: product.w0, w2: product.w1, w3: product.w2, w4: product.w3, w5: product.w4}

	if e_out < 1 {
		d := 1 - e_out
		if d > 55 {
			d = 55
		}
		e_out = 1
		z.w5, z.w4, z.w3, z.w2 = srl256_short(z.w5, z.w4, z.w3, z.w2, uint(d))
	}
	c_prov = z.w5

	rbIdx := (rnd_mode << 2) + ((s & 1) << 1) + int(c_prov&1)
	if lt128(
		bid_roundbound_128[rbIdx].hi,
		bid_roundbound_128[rbIdx].lo,
		z.w4, z.w3) {
		c_prov = c_prov + 1
		if c_prov == (1 << 53) {
			c_prov = 1 << 52
			e_out = e_out + 1
		} else if (c_prov == (1 << 52)) && (e_out == 1) {
			if (((rnd_mode & 3) == 0) && (z.w4 < (3 << 62))) ||
				((rnd_mode+int(s&1) == 2) && (z.w4 < (1 << 63))) {
				flags |= BID_UNDERFLOW_EXCEPTION
			}
		}
	}

	if e_out >= 2047 {
		flags |= BID_OVERFLOW_EXCEPTION | BID_INEXACT_EXCEPTION
		if (rnd_mode == BID_ROUNDING_TO_ZERO) ||
			(rnd_mode == boolToRndMode(s != 0)) {
			return (uint64(s) << 63) + (uint64(2046) << 52) + uint64((1<<52)-1), flags
		}
		return (uint64(s) << 63) + (uint64(2047) << 52) + uint64(0), flags
	}

	if c_prov < (1 << 52) {
		e_out = 0
	} else {
		c_prov = c_prov & ((1 << 52) - 1)
	}

	if (z.w4 != 0) || (z.w3 != 0) {
		flags |= BID_INEXACT_EXCEPTION
		if e_out == 0 {
			flags |= BID_UNDERFLOW_EXCEPTION
		}
	}

	return (uint64(s) << 63) + (uint64(e_out) << 52) + uint64(c_prov), flags
}

func Bid64ToBinary128(x uint64, rnd_mode int) (BID_UINT128, uint32) {
	if rnd_mode < BID_ROUNDING_TO_NEAREST || rnd_mode > BID_ROUNDING_TIES_AWAY {
		return BID_UINT128{hi: 0x7fff800000000000}, BID_INVALID_EXCEPTION
	}
	var flags uint32
	s, e, k, coeff, isZero, isInf, isNaN, nanPayloadHi, isSNaN := unpack_bid64_binarydecimal(x)
	if isZero {
		return BID_UINT128{hi: uint64(s) << 63}, flags
	}
	if isInf {
		return BID_UINT128{hi: (uint64(s) << 63) + (32767 << 48)}, flags
	}
	if isNaN {
		if isSNaN {
			flags |= BID_INVALID_EXCEPTION
		}
		return BID_UINT128{hi: (uint64(s) << 63) + (32767 << 48) + (nanPayloadHi >> 17) + (1 << 47), lo: nanPayloadHi << 47}, flags
	}
	c := BID_UINT128{lo: coeff}
	c.hi, c.lo = sll128_short(c.hi, c.lo, 61)
	k += 59
	m_min := bid_breakpoints_binary128[e+5000]
	e_out := bid_exponents_binary128[e+5000] - k
	var r BID_UINT256
	if le128(c.hi, c.lo, m_min.hi, m_min.lo) {
		r = bid_multipliers1_binary128[e+5000]
	} else {
		r = bid_multipliers2_binary128[e+5000]
		e_out++
	}
	z := __mul_128x256_to_384(c, r)
	c_prov_hi := z.w5
	c_prov_lo := z.w4
	rbIdx := (rnd_mode << 2) + ((s & 1) << 1) + int(c_prov_lo&1)
	if lt128(bid_roundbound_128[rbIdx].hi, bid_roundbound_128[rbIdx].lo, z.w3, z.w2) {
		c_prov_lo++
		if c_prov_lo == 0 {
			c_prov_hi++
		}
	}
	c_prov_hi &= (1 << 48) - 1
	if z.w3 != 0 || z.w2 != 0 {
		flags |= BID_INEXACT_EXCEPTION
	}
	return BID_UINT128{hi: (uint64(s) << 63) + (uint64(e_out) << 48) + c_prov_hi, lo: c_prov_lo}, flags
}

func Bid128ToBinary128(x BID_UINT128, rnd_mode int) (BID_UINT128, uint32) {
	if rnd_mode < BID_ROUNDING_TO_NEAREST || rnd_mode > BID_ROUNDING_TIES_AWAY {
		return BID_UINT128{hi: 0x7fff800000000000}, BID_INVALID_EXCEPTION
	}
	var flags uint32
	s, e, k, c, isZero, isInf, isNaN, nanPayloadHi, nanPayloadLo, isSNaN := unpack_bid128_binarydecimal(x)
	if isZero {
		return BID_UINT128{hi: uint64(s) << 63}, flags
	}
	if isInf {
		return BID_UINT128{hi: (uint64(s) << 63) + (32767 << 48)}, flags
	}
	if isNaN {
		if isSNaN {
			flags |= BID_INVALID_EXCEPTION
		}
		return BID_UINT128{hi: (uint64(s) << 63) + (32767 << 48) + (nanPayloadHi >> 17) + (1 << 47), lo: (nanPayloadLo >> 17) + (nanPayloadHi << 47)}, flags
	}
	c.hi, c.lo = sll128_short(c.hi, c.lo, 2)
	if e >= 4933 {
		flags |= BID_OVERFLOW_EXCEPTION | BID_INEXACT_EXCEPTION
		if rnd_mode == BID_ROUNDING_TO_ZERO || rnd_mode == boolToRndMode(s != 0) {
			return BID_UINT128{hi: (uint64(s) << 63) + (32766 << 48) + ((1 << 48) - 1), lo: 0xffffffffffffffff}, flags
		}
		return BID_UINT128{hi: (uint64(s) << 63) + (32767 << 48)}, flags
	}
	if e <= -5000 {
		e = -5000
	}
	m_min := bid_breakpoints_binary128[e+5000]
	e_out := bid_exponents_binary128[e+5000] - k
	var r BID_UINT256
	if le128(c.hi, c.lo, m_min.hi, m_min.lo) {
		r = bid_multipliers1_binary128[e+5000]
	} else {
		r = bid_multipliers2_binary128[e+5000]
		e_out++
	}
	z := __mul_128x256_to_384(c, r)
	if e_out < 1 {
		d := 1 - e_out
		if d > 115 {
			d = 115
		}
		if d >= 64 {
			d -= 64
			z.w2, z.w3, z.w4, z.w5 = z.w3, z.w4, z.w5, 0
		}
		e_out = 1
		if d > 0 {
			z.w5, z.w4, z.w3, z.w2 = srl256_short(z.w5, z.w4, z.w3, z.w2, uint(d))
		}
	}
	c_prov_hi := z.w5
	c_prov_lo := z.w4
	rbIdx := (rnd_mode << 2) + ((s & 1) << 1) + int(c_prov_lo&1)
	if lt128(bid_roundbound_128[rbIdx].hi, bid_roundbound_128[rbIdx].lo, z.w3, z.w2) {
		c_prov_lo++
		if c_prov_lo == 0 {
			c_prov_hi++
			if c_prov_hi == 1<<49 {
				c_prov_hi = 1 << 48
				e_out++
			} else if c_prov_hi == 1<<48 && e_out == 1 {
				if rnd_mode+(s&1) == 2 && z.w3 < 1<<63 {
					flags |= BID_UNDERFLOW_EXCEPTION
				}
			}
		}
	}
	if e_out >= 32767 {
		flags |= BID_OVERFLOW_EXCEPTION | BID_INEXACT_EXCEPTION
		if rnd_mode == BID_ROUNDING_TO_ZERO || rnd_mode == boolToRndMode(s != 0) {
			return BID_UINT128{hi: (uint64(s) << 63) + (32766 << 48) + ((1 << 48) - 1), lo: 0xffffffffffffffff}, flags
		}
		return BID_UINT128{hi: (uint64(s) << 63) + (32767 << 48)}, flags
	}
	if c_prov_hi < 1<<48 {
		e_out = 0
	} else {
		c_prov_hi &= (1 << 48) - 1
	}
	if z.w3 != 0 || z.w2 != 0 {
		flags |= BID_INEXACT_EXCEPTION
		if e_out == 0 {
			flags |= BID_UNDERFLOW_EXCEPTION
		}
	}
	return BID_UINT128{hi: (uint64(s) << 63) + (uint64(e_out) << 48) + c_prov_hi, lo: c_prov_lo}, flags
}
