// Ported from: Intel bid_internal.h (bid128 section)
// Mechanical translation - all logic preserved exactly.

package bidgo

// BID128 constants from bid_internal.h
const (
	EXPONENT_MASK128      = 0x3fff
	LARGE_COEFF_MASK128   = 0x0001ffffffffffff
	SMALL_COEFF_MASK128   = 0x0001ffffffffffff
	EXPONENT_BIAS128      = 6176
	MAX_FORMAT_DIGITS_128 = 34
	DECIMAL_MAX_EXPON_128 = 12287

	// 128-bit specific masks (same as 64-bit but named for clarity)
	MASK_EXP128     = 0x7ffe000000000000
	MASK_SPECIAL128 = 0x7800000000000000
	MASK_COEFF128   = 0x0001ffffffffffff
)

// unpack_BID128 unpacks a BID128 value into sign, exponent, and coefficient.
// Returns 0 if special (NaN/Inf/zero/non-canonical), non-zero otherwise.
func unpack_BID128(x BID_UINT128) (sign uint64, exponent int, coefficient BID_UINT128, valid uint64) {
	sign = x.hi & 0x8000000000000000

	// special encodings
	if (x.hi & INFINITY_MASK64) >= SPECIAL_ENCODING_MASK64 {
		if (x.hi & INFINITY_MASK64) < INFINITY_MASK64 {
			// non-canonical input
			coefficient.lo = 0
			coefficient.hi = 0
			ex := x.hi >> 47
			exponent = int(ex) & EXPONENT_MASK128
			return sign, exponent, coefficient, 0
		}
		// NaN or Infinity
		T33 := bid_power10_table_128[33]
		coeff := BID_UINT128{lo: x.lo, hi: x.hi & LARGE_COEFF_MASK128}
		coefficient.lo = x.lo
		coefficient.hi = x.hi
		if __unsigned_compare_ge_128(coeff, T33) {
			coefficient.hi &= ^uint64(LARGE_COEFF_MASK128)
			coefficient.lo = 0
		}
		exponent = 0
		return sign, exponent, coefficient, 0
	}

	coeff := BID_UINT128{lo: x.lo, hi: x.hi & SMALL_COEFF_MASK128}

	T34 := bid_power10_table_128[34]
	if __unsigned_compare_ge_128(coeff, T34) {
		coeff.lo = 0
		coeff.hi = 0
	}

	coefficient.lo = coeff.lo
	coefficient.hi = coeff.hi

	ex := x.hi >> 49
	exponent = int(ex) & EXPONENT_MASK128

	return sign, exponent, coefficient, coeff.lo | coeff.hi
}

// very_fast_get_BID128 packs sign, exponent, and coefficient into BID128.
func very_fast_get_BID128(sgn uint64, expon int, coeff BID_UINT128) BID_UINT128 {
	var res BID_UINT128
	res.lo = coeff.lo
	res.hi = sgn | (uint64(expon) << 49) | coeff.hi
	return res
}

// unpack_BID128_value unpacks a BID128 value into sign, exponent, and coefficient.
// Returns false if special (NaN/Inf/zero/non-canonical), true otherwise.
// Ported mechanically from Intel bid_internal.h: unpack_BID128_value.
// Note: different from unpack_BID128 in NaN/Inf coefficient handling.
func unpack_BID128_value(x BID_UINT128) (sign_x uint64, exponent_x int, coefficient_x BID_UINT128, valid bool) {
	sign_x = x.hi & 0x8000000000000000

	// special encodings
	if (x.hi & INFINITY_MASK64) >= SPECIAL_ENCODING_MASK64 {
		if (x.hi & INFINITY_MASK64) < INFINITY_MASK64 {
			// non-canonical input
			coefficient_x.lo = 0
			coefficient_x.hi = 0
			ex := x.hi >> 47
			exponent_x = int(ex) & EXPONENT_MASK128
			return sign_x, exponent_x, coefficient_x, false
		}
		// 10^33
		T33 := bid_power10_table_128[33]

		coefficient_x.lo = x.lo
		coefficient_x.hi = x.hi & 0x00003fffffffffff
		if __unsigned_compare_ge_128(coefficient_x, T33) { // non-canonical
			coefficient_x.hi = x.hi & 0xfe00000000000000
			coefficient_x.lo = 0
		} else {
			coefficient_x.hi = x.hi & 0xfe003fffffffffff
		}
		if (x.hi & NAN_MASK64) == INFINITY_MASK64 {
			coefficient_x.lo = 0
			coefficient_x.hi = x.hi & SINFINITY_MASK64
		}
		exponent_x = 0
		return sign_x, exponent_x, coefficient_x, false // NaN or Infinity
	}

	coeff := BID_UINT128{lo: x.lo, hi: x.hi & SMALL_COEFF_MASK128}

	// 10^34
	T34 := bid_power10_table_128[34]
	// check for non-canonical values
	if __unsigned_compare_ge_128(coeff, T34) {
		coeff.lo = 0
		coeff.hi = 0
	}

	coefficient_x.lo = coeff.lo
	coefficient_x.hi = coeff.hi

	ex := x.hi >> 49
	exponent_x = int(ex) & EXPONENT_MASK128

	return sign_x, exponent_x, coefficient_x, (coeff.lo | coeff.hi) != 0
}

// Bid128IsNaN returns 1 if x is NaN.
func Bid128IsNaN(x BID_UINT128) int {
	if (x.hi & NAN_MASK64) == NAN_MASK64 {
		return 1
	}
	return 0
}

// Bid128IsInf returns 1 if x is infinity.
func Bid128IsInf(x BID_UINT128) int {
	if ((x.hi & INFINITY_MASK64) == INFINITY_MASK64) && ((x.hi & NAN_MASK64) != NAN_MASK64) {
		return 1
	}
	return 0
}

// Bid128IsZero returns 1 if x is zero.
func Bid128IsZero(x BID_UINT128) int {
	_, _, coeff, valid := unpack_BID128(x)
	if valid == 0 && Bid128IsNaN(x) == 0 && Bid128IsInf(x) == 0 {
		return 1
	}
	if coeff.lo == 0 && coeff.hi == 0 {
		return 1
	}
	return 0
}
