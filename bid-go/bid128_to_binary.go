// Ported from: Intel bid_binarydecimal.c
// Mechanical translation - all logic preserved exactly.

package bidgo

import (
	"math"
	"math/bits"
)

// clz64_nz counts leading zeros in a 64-bit word (undefined for 0 input).
// Ported from bid_binarydecimal.c: clz64_nz macro.
func clz64_nz(n uint64) int {
	return bits.LeadingZeros64(n)
}

// clz64 counts leading zeros in a 64-bit word.
// Ported from bid_binarydecimal.c: clz64 macro.
func clz64(n uint64) int {
	if n == 0 {
		return 64
	}
	return clz64_nz(n)
}

// clz128_nz counts leading zeros in a 128-bit word (undefined for 0 input).
// Ported from bid_binarydecimal.c: clz128_nz macro.
func clz128_nz(n_hi, n_lo uint64) int {
	if n_hi == 0 {
		return 64 + clz64_nz(n_lo)
	}
	return clz64_nz(n_hi)
}

// sll128_short shifts a 128-bit value left by c bits (0 < c < 64).
// Ported from bid_binarydecimal.c: sll128_short macro.
func sll128_short(hi, lo uint64, c uint) (uint64, uint64) {
	hi = (hi << c) + (lo >> (64 - c))
	lo = lo << c
	return hi, lo
}

// sll128 shifts a 128-bit value left by c bits.
// Ported from bid_binarydecimal.c: sll128 macro.
func sll128(hi, lo uint64, c uint) (uint64, uint64) {
	if c == 0 {
		return hi, lo
	}
	if c >= 64 {
		return lo << (c - 64), 0
	}
	return sll128_short(hi, lo, c)
}

// srl256_short shifts a 256-bit value right by c bits (0 < c < 64).
// Ported from bid_binarydecimal.c: srl256_short macro.
func srl256_short(x3, x2, x1, x0 uint64, c uint) (uint64, uint64, uint64, uint64) {
	x0 = (x1 << (64 - c)) + (x0 >> c)
	x1 = (x2 << (64 - c)) + (x1 >> c)
	x2 = (x3 << (64 - c)) + (x2 >> c)
	x3 = x3 >> c
	return x3, x2, x1, x0
}

// lt128 compares "<" two 2-part unsigned integers.
// Ported from bid_binarydecimal.c: lt128 macro.
func lt128(x_hi, x_lo, y_hi, y_lo uint64) bool {
	return (x_hi < y_hi) || ((x_hi == y_hi) && (x_lo < y_lo))
}

// le128 compares "<=" two 2-part unsigned integers.
// Ported from bid_binarydecimal.c: le128 macro.
func le128(x_hi, x_lo, y_hi, y_lo uint64) bool {
	return (x_hi < y_hi) || ((x_hi == y_hi) && (x_lo <= y_lo))
}

// __mul_128x256_to_384 multiplies 128-bit by 256-bit to get 384-bit result.
// Ported from bid_binarydecimal.c: __mul_128x256_to_384 macro.
func __mul_128x256_to_384(A BID_UINT128, B BID_UINT256) BID_UINT384 {
	var P BID_UINT384
	var CY uint64

	P0 := __mul_64x256_to_320(A.w[0], B)
	P1 := __mul_64x256_to_320(A.w[1], B)
	P.w[0] = P0.w[0]
	P.w[1], CY = __add_carry_out(P1.w[0], P0.w[1])
	P.w[2], CY = __add_carry_in_out(P1.w[1], P0.w[2], CY)
	P.w[3], CY = __add_carry_in_out(P1.w[2], P0.w[3], CY)
	P.w[4], CY = __add_carry_in_out(P1.w[3], P0.w[4], CY)
	P.w[5] = P1.w[4] + CY
	return P
}

// unpack_bid128_binarydecimal unpacks a BID128 for binary conversion.
// Returns s (sign in LSB), e (exponent), k (normalization shift), c (normalized coefficient).
// isZero, isInf, isNaN indicate special values.
// nanPayloadHi, nanPayloadLo are the NaN payload (shifted).
//
// Ported from bid_binarydecimal.c: unpack_bid128 macro (lines 685-712).
func unpack_bid128_binarydecimal(x BID_UINT128) (s int, e int, k int, c BID_UINT128, isZero bool, isInf bool, isNaN bool, nanPayloadHi uint64, nanPayloadLo uint64, isSNaN bool) {
	s = int(x.w[1] >> 63)

	if (x.w[1] & (3 << 61)) == (3 << 61) {
		if (x.w[1] & (0xF << 59)) == (0xF << 59) {
			if (x.w[1] & (0x1F << 58)) != (0x1F << 58) {
				isInf = true
				return
			}
			if (x.w[1] & (1 << 57)) != 0 {
				isSNaN = true
			}
			isNaN = true
			if lt128(54210108624275, 4089650035136921599,
				x.w[1]&0x3FFFFFFFFFFF, x.w[0]) {
				nanPayloadHi = 0
				nanPayloadLo = 0
			} else {
				nanPayloadHi = (x.w[1] << 18) + (x.w[0] >> 46)
				nanPayloadLo = x.w[0] << 18
			}
			return
		}
		// non-canonical with 11 prefix but not special -> zero
		isZero = true
		return
	}

	e = int((x.w[1]>>49)&((1<<14)-1)) - 6176
	c.w[1] = x.w[1] & ((1 << 49) - 1)
	c.w[0] = x.w[0]
	if lt128(542101086242752, 4003012203950112767, c.w[1], c.w[0]) {
		c.w[1] = 0
		c.w[0] = 0
	}
	if (c.w[1] == 0) && (c.w[0] == 0) {
		isZero = true
		return
	}
	k = clz128_nz(c.w[1], c.w[0]) - 15
	c.w[1], c.w[0] = sll128(c.w[1], c.w[0], uint(k))
	return
}

// return_binary32_pack packs sign, exponent, and coefficient into a float32 bits.
// Ported from bid_binarydecimal.c: return_binary32 macro.
func return_binary32_pack(s int, e int, c uint64) float32 {
	bits := (uint32(s) << 31) + (uint32(e) << 23) + uint32(c)
	return math.Float32frombits(bits)
}

// return_binary64_pack packs sign, exponent, and coefficient into a float64 bits.
// Ported from bid_binarydecimal.c: return_binary64 macro.
func return_binary64_pack(s int, e int, c uint64) float64 {
	bits := (uint64(s) << 63) + (uint64(e) << 52) + c
	return math.Float64frombits(bits)
}

// Bid128ToBinary32 converts BID128 to float32.
// Ported from bid_binarydecimal.c: bid128_to_binary32 (lines 144224-144343).
func Bid128ToBinary32(x BID_UINT128, rnd_mode int, pfpsf *uint32) float32 {
	var c_prov uint64
	var c BID_UINT128
	var m_min BID_UINT128
	var e_out int
	var r BID_UINT256
	var z BID_UINT384

	s, e, k, c, isZero, isInf, isNaN, nanPayloadHi, nanPayloadLo, isSNaN :=
		unpack_bid128_binarydecimal(x)

	if isZero {
		return return_binary32_pack(s, 0, 0)
	}
	if isInf {
		return return_binary32_pack(s, 255, 0)
	}
	if isNaN {
		if isSNaN {
			*pfpsf |= BID_INVALID_EXCEPTION
		}
		_ = nanPayloadLo
		return return_binary32_pack(s, 255, (nanPayloadHi>>42)+(1<<22))
	}

	// Check for "trivial" overflow
	// e >= ceil(128 * log_10(2)) = 39
	if e >= 39 {
		*pfpsf |= BID_OVERFLOW_EXCEPTION | BID_INEXACT_EXCEPTION
		// return_binary32_ovf(s)
		if (rnd_mode == BID_ROUNDING_TO_ZERO) ||
			(rnd_mode == boolToRndMode(s != 0)) {
			return return_binary32_pack(s, 254, (1<<23)-1)
		}
		return return_binary32_pack(s, 255, 0)
	}

	// Check for "trivial" underflow
	if e <= -80 {
		e = -80
	}

	// Look up the breakpoint and approximate exponent
	m_min = bid_breakpoints_binary32[e+80]
	e_out = bid_exponents_binary32[e+80] - k

	// Choose provisional exponent and reciprocal multiplier based on breakpoint
	if le128(c.w[1], c.w[0], m_min.w[1], m_min.w[0]) {
		r = bid_multipliers1_binary32[e+80]
	} else {
		r = bid_multipliers2_binary32[e+80]
		e_out = e_out + 1
	}

	// Do the reciprocal multiplication
	z = __mul_128x256_to_384(c, r)

	// Check for exponent underflow and compensate by shifting the product
	if e_out < 1 {
		d := 1 - e_out
		if d > 26 {
			d = 26
		}
		e_out = 1
		z.w[5], z.w[4], z.w[3], z.w[2] = srl256_short(z.w[5], z.w[4], z.w[3], z.w[2], uint(d))
	}
	c_prov = z.w[5]

	// Round using round-sticky words
	// If we spill into the next binade, correct
	// Flag underflow where it may be needed even for |result| = SNN
	rbIdx := (rnd_mode << 2) + ((s & 1) << 1) + int(c_prov&1)
	if lt128(
		bid_roundbound_128[rbIdx].w[1],
		bid_roundbound_128[rbIdx].w[0],
		z.w[4], z.w[3]) {
		c_prov = c_prov + 1
		if c_prov == (1 << 24) {
			c_prov = 1 << 23
			e_out = e_out + 1
		} else if (c_prov == (1 << 23)) && (e_out == 1) {
			// BINARY_TINY_DETECTION_AFTER_ROUNDING
			if (((rnd_mode & 3) == 0) && (z.w[4] < (3 << 62))) ||
				((rnd_mode+int(s&1) == 2) && (z.w[4] < (1 << 63))) {
				*pfpsf |= BID_UNDERFLOW_EXCEPTION
			}
		}
	}

	// Check for overflow
	if e_out >= 255 {
		*pfpsf |= BID_OVERFLOW_EXCEPTION | BID_INEXACT_EXCEPTION
		// return_binary32_ovf(s)
		if (rnd_mode == BID_ROUNDING_TO_ZERO) ||
			(rnd_mode == boolToRndMode(s != 0)) {
			return return_binary32_pack(s, 254, (1<<23)-1)
		}
		return return_binary32_pack(s, 255, 0)
	}

	// Modify exponent for a tiny result, otherwise lop the implicit bit
	if c_prov < (1 << 23) {
		e_out = 0
	} else {
		c_prov = c_prov & ((1 << 23) - 1)
	}

	// Set the inexact and underflow flag as appropriate
	if (z.w[4] != 0) || (z.w[3] != 0) {
		*pfpsf |= BID_INEXACT_EXCEPTION
		if e_out == 0 {
			*pfpsf |= BID_UNDERFLOW_EXCEPTION
		}
	}

	// Package up the result as a binary floating-point number
	return return_binary32_pack(s, e_out, c_prov)
}

// Bid128ToBinary64 converts BID128 to float64.
// Ported from bid_binarydecimal.c: bid128_to_binary64 (lines 144591-144708).
func Bid128ToBinary64(x BID_UINT128, rnd_mode int, pfpsf *uint32) float64 {
	var c_prov uint64
	var c BID_UINT128
	var m_min BID_UINT128
	var e_out int
	var r BID_UINT256
	var z BID_UINT384

	s, e, k, c, isZero, isInf, isNaN, nanPayloadHi, nanPayloadLo, isSNaN :=
		unpack_bid128_binarydecimal(x)

	if isZero {
		return return_binary64_pack(s, 0, 0)
	}
	if isInf {
		return return_binary64_pack(s, 2047, 0)
	}
	if isNaN {
		if isSNaN {
			*pfpsf |= BID_INVALID_EXCEPTION
		}
		_ = nanPayloadLo
		return return_binary64_pack(s, 2047, (nanPayloadHi>>13)+(1<<51))
	}

	// Shift 6 more places left ready for reciprocal multiplication
	c.w[1], c.w[0] = sll128_short(c.w[1], c.w[0], 6)

	// Check for "trivial" overflow
	// e >= ceil(1024 * log_10(2)) = ceil(308.25) = 309
	if e >= 309 {
		*pfpsf |= BID_OVERFLOW_EXCEPTION | BID_INEXACT_EXCEPTION
		// return_binary64_ovf(s)
		if (rnd_mode == BID_ROUNDING_TO_ZERO) ||
			(rnd_mode == boolToRndMode(s != 0)) {
			return return_binary64_pack(s, 2046, (1<<52)-1)
		}
		return return_binary64_pack(s, 2047, 0)
	}

	// Check for "trivial" underflow
	if e <= -358 {
		e = -358
	}

	// Look up the breakpoint and approximate exponent
	m_min = bid_breakpoints_binary64[e+358]
	e_out = bid_exponents_binary64[e+358] - k

	// Choose provisional exponent and reciprocal multiplier based on breakpoint
	if le128(c.w[1], c.w[0], m_min.w[1], m_min.w[0]) {
		r = bid_multipliers1_binary64[e+358]
	} else {
		r = bid_multipliers2_binary64[e+358]
		e_out = e_out + 1
	}

	// Do the reciprocal multiplication
	z = __mul_128x256_to_384(c, r)

	// Check for exponent underflow and compensate by shifting the product
	if e_out < 1 {
		d := 1 - e_out
		if d > 55 {
			d = 55
		}
		e_out = 1
		z.w[5], z.w[4], z.w[3], z.w[2] = srl256_short(z.w[5], z.w[4], z.w[3], z.w[2], uint(d))
	}
	c_prov = z.w[5]

	// Round using round-sticky words
	// If we spill into the next binade, correct
	// Flag underflow where it may be needed even for |result| = SNN
	rbIdx := (rnd_mode << 2) + ((s & 1) << 1) + int(c_prov&1)
	if lt128(
		bid_roundbound_128[rbIdx].w[1],
		bid_roundbound_128[rbIdx].w[0],
		z.w[4], z.w[3]) {
		c_prov = c_prov + 1
		if c_prov == (1 << 53) {
			c_prov = 1 << 52
			e_out = e_out + 1
		} else if (c_prov == (1 << 52)) && (e_out == 1) {
			// BINARY_TINY_DETECTION_AFTER_ROUNDING
			if (((rnd_mode & 3) == 0) && (z.w[4] < (3 << 62))) ||
				((rnd_mode+int(s&1) == 2) && (z.w[4] < (1 << 63))) {
				*pfpsf |= BID_UNDERFLOW_EXCEPTION
			}
		}
	}

	// Check for overflow
	if e_out >= 2047 {
		*pfpsf |= BID_OVERFLOW_EXCEPTION | BID_INEXACT_EXCEPTION
		// return_binary64_ovf(s)
		if (rnd_mode == BID_ROUNDING_TO_ZERO) ||
			(rnd_mode == boolToRndMode(s != 0)) {
			return return_binary64_pack(s, 2046, (1<<52)-1)
		}
		return return_binary64_pack(s, 2047, 0)
	}

	// Modify exponent for a tiny result, otherwise lop the implicit bit
	if c_prov < (1 << 52) {
		e_out = 0
	} else {
		c_prov = c_prov & ((1 << 52) - 1)
	}

	// Set the inexact and underflow flag as appropriate
	if (z.w[4] != 0) || (z.w[3] != 0) {
		*pfpsf |= BID_INEXACT_EXCEPTION
		if e_out == 0 {
			*pfpsf |= BID_UNDERFLOW_EXCEPTION
		}
	}

	// Package up the result as a binary floating-point number
	return return_binary64_pack(s, e_out, c_prov)
}

// boolToRndMode maps overflow direction:
// s != 0 (negative) → BID_ROUNDING_UP, s == 0 (positive) → BID_ROUNDING_DOWN
// This matches the C macro: rnd_mode == ((s!=0) ? BID_ROUNDING_UP : BID_ROUNDING_DOWN)
func boolToRndMode(neg bool) int {
	if neg {
		return BID_ROUNDING_UP
	}
	return BID_ROUNDING_DOWN
}
