// Ported from: IntelRDFPMathLib20U4/LIBRARY/src/bid_internal.h
// Version: Intel(R) Decimal Floating-Point Math Library 2.0 Update 4
//
// This file is a mechanical translation of the Intel BID library to Go.
// All logic, magic numbers, and table references are preserved exactly.

package bidgo

import (
	"math"
	"math/bits"
)

// BID_UINT128 represents a 128-bit unsigned integer
// w[0] is low 64 bits, w[1] is high 64 bits
type BID_UINT128 struct {
	w [2]uint64
}

// BID_UINT192 represents a 192-bit unsigned integer
type BID_UINT192 struct {
	w [3]uint64
}

// BID_UINT256 represents a 256-bit unsigned integer
type BID_UINT256 struct {
	w [4]uint64
}

// BID_UINT320 represents a 320-bit unsigned integer
type BID_UINT320 struct {
	w [5]uint64
}

// BID_UINT384 represents a 384-bit unsigned integer
type BID_UINT384 struct {
	w [6]uint64
}

// BID_UINT512 represents a 512-bit unsigned integer
type BID_UINT512 struct {
	w [8]uint64
}

// Constants from bid_internal.h
const (
	DECIMAL_MAX_EXPON_64    = 767
	DECIMAL_EXPONENT_BIAS   = 398
	MAX_FORMAT_DIGITS       = 16
	SPECIAL_ENCODING_MASK64 = 0x6000000000000000
	INFINITY_MASK64         = 0x7800000000000000
	SINFINITY_MASK64        = 0xf800000000000000
	SSNAN_MASK64            = 0xfc00000000000000
	NAN_MASK64              = 0x7c00000000000000
	SNAN_MASK64             = 0x7e00000000000000
	QUIET_MASK64            = 0xfdffffffffffffff
	LARGE_COEFF_MASK64      = 0x0007ffffffffffff
	LARGE_COEFF_HIGH_BIT64  = 0x0020000000000000
	SMALL_COEFF_MASK64      = 0x001fffffffffffff
	EXPONENT_MASK64         = 0x3ff
	EXPONENT_SHIFT_LARGE64  = 51
	EXPONENT_SHIFT_SMALL64  = 53
	LARGEST_BID64           = 0x77fb86f26fc0ffff
	SMALLEST_BID64          = 0xf7fb86f26fc0ffff
	MASK_BINARY_EXPONENT    = 0x7ff0000000000000
	BINARY_EXPONENT_BIAS    = 0x3ff
)

// Rounding modes
// IEEE 754-2019 표준 모드 (0-4): Intel BID 라이브러리와 동일
// 비표준 모드 (5+): decTest(IBM decNumber) 호환을 위해 추가
const (
	BID_ROUNDING_TO_NEAREST   = 0 // IEEE: roundTiesToEven (half_even)
	BID_ROUNDING_DOWN         = 1 // IEEE: roundTowardNegative (floor)
	BID_ROUNDING_UP           = 2 // IEEE: roundTowardPositive (ceiling)
	BID_ROUNDING_TO_ZERO      = 3 // IEEE: roundTowardZero (truncate)
	BID_ROUNDING_TIES_AWAY    = 4 // IEEE: roundTiesToAway (half_up)
	BID_ROUNDING_NEAREST_DOWN = 5 // 비표준: half_down (ties toward zero) - decTest 호환용
)

// Exception flags
const (
	BID_INEXACT_EXCEPTION     = 0x20
	BID_UNDERFLOW_EXCEPTION   = 0x10
	BID_OVERFLOW_EXCEPTION    = 0x08
	BID_ZERO_DIVIDE_EXCEPTION = 0x04
	BID_INVALID_EXCEPTION     = 0x01
	BID_EXACT_STATUS          = 0x00
)

// __shr_128 performs right shift on 128-bit value
func __shr_128(a BID_UINT128, k uint) BID_UINT128 {
	var q BID_UINT128
	q.w[0] = a.w[0] >> k
	q.w[0] |= a.w[1] << (64 - k)
	q.w[1] = a.w[1] >> k
	return q
}

// __shr_128_long performs right shift on 128-bit value (handles k >= 64)
func __shr_128_long(a BID_UINT128, k uint) BID_UINT128 {
	var q BID_UINT128
	if k < 64 {
		q.w[0] = a.w[0] >> k
		q.w[0] |= a.w[1] << (64 - k)
		q.w[1] = a.w[1] >> k
	} else {
		q.w[0] = a.w[1] >> (k - 64)
		q.w[1] = 0
	}
	return q
}

// __shl_128_long performs left shift on 128-bit value (handles k >= 64)
func __shl_128_long(a BID_UINT128, k uint) BID_UINT128 {
	var q BID_UINT128
	if k < 64 {
		q.w[1] = a.w[1] << k
		q.w[1] |= a.w[0] >> (64 - k)
		q.w[0] = a.w[0] << k
	} else {
		q.w[1] = a.w[0] << (k - 64)
		q.w[0] = 0
	}
	return q
}

// __unsigned_compare_gt_128 returns true if A > B
func __unsigned_compare_gt_128(a, b BID_UINT128) bool {
	return (a.w[1] > b.w[1]) || ((a.w[1] == b.w[1]) && (a.w[0] > b.w[0]))
}

// __unsigned_compare_ge_128 returns true if A >= B
func __unsigned_compare_ge_128(a, b BID_UINT128) bool {
	return (a.w[1] > b.w[1]) || ((a.w[1] == b.w[1]) && (a.w[0] >= b.w[0]))
}

// __test_equal_128 returns true if A == B
func __test_equal_128(a, b BID_UINT128) bool {
	return (a.w[1] == b.w[1]) && (a.w[0] == b.w[0])
}

// __add_128_64 adds 64-bit value to 128-bit
func __add_128_64(a BID_UINT128, b uint64) BID_UINT128 {
	var r BID_UINT128
	r64h := a.w[1]
	r.w[0] = b + a.w[0]
	if r.w[0] < b {
		r64h++
	}
	r.w[1] = r64h
	return r
}

// __sub_128_64 subtracts 64-bit value from 128-bit
func __sub_128_64(a BID_UINT128, b uint64) BID_UINT128 {
	var r BID_UINT128
	r64h := a.w[1]
	if a.w[0] < b {
		r64h--
	}
	r.w[1] = r64h
	r.w[0] = a.w[0] - b
	return r
}

// __add_128_128 adds two 128-bit values
func __add_128_128(a, b BID_UINT128) BID_UINT128 {
	var q BID_UINT128
	q.w[1] = a.w[1] + b.w[1]
	q.w[0] = b.w[0] + a.w[0]
	if q.w[0] < b.w[0] {
		q.w[1]++
	}
	return q
}

// __sub_128_128 subtracts two 128-bit values
func __sub_128_128(a, b BID_UINT128) BID_UINT128 {
	var q BID_UINT128
	q.w[1] = a.w[1] - b.w[1]
	q.w[0] = a.w[0] - b.w[0]
	if a.w[0] < b.w[0] {
		q.w[1]--
	}
	return q
}

// __add_carry_out adds two 64-bit values and returns carry
func __add_carry_out(x, y uint64) (s uint64, cy uint64) {
	s = x + y
	if s < x {
		cy = 1
	}
	return
}

// __add_carry_in_out adds two 64-bit values with carry in and returns carry out
func __add_carry_in_out(x, y, ci uint64) (s uint64, cy uint64) {
	x1 := x + ci
	s = x1 + y
	if (s < x1) || (x1 < ci) {
		cy = 1
	}
	return
}

// __sub_borrow_out subtracts two 64-bit values and returns borrow
func __sub_borrow_out(x, y uint64) (s uint64, cy uint64) {
	s = x - y
	if s > x {
		cy = 1
	}
	return
}

// __mul_64x64_to_128 multiplies two 64-bit values to get 128-bit result
func __mul_64x64_to_128(cx, cy uint64) BID_UINT128 {
	hi, lo := bits.Mul64(cx, cy)
	return BID_UINT128{w: [2]uint64{lo, hi}}
}

// __mul_64x64_to_128_fast is the same as __mul_64x64_to_128 for CX, CY < 2^61
func __mul_64x64_to_128_fast(cx, cy uint64) BID_UINT128 {
	return __mul_64x64_to_128(cx, cy)
}

// __mul_64x128_full multiplies 64-bit by 128-bit, returns 64-bit high and 128-bit low
func __mul_64x128_full(a uint64, b BID_UINT128) (ph uint64, ql BID_UINT128) {
	albh := __mul_64x64_to_128(a, b.w[1])
	albl := __mul_64x64_to_128(a, b.w[0])

	ql.w[0] = albl.w[0]
	qm2 := __add_128_64(albh, albl.w[1])
	ql.w[1] = qm2.w[0]
	ph = qm2.w[1]
	return
}

// __mul_64x128_to_192 multiplies 64-bit by 128-bit to get 192-bit result
func __mul_64x128_to_192(a uint64, b BID_UINT128) BID_UINT192 {
	var q BID_UINT192
	albh := __mul_64x64_to_128(a, b.w[1])
	albl := __mul_64x64_to_128(a, b.w[0])

	q.w[0] = albl.w[0]
	qm2 := __add_128_64(albh, albl.w[1])
	q.w[1] = qm2.w[0]
	q.w[2] = qm2.w[1]
	return q
}

// __mul_128x128_to_256 multiplies two 128-bit values to get 256-bit result
func __mul_128x128_to_256(a, b BID_UINT128) BID_UINT256 {
	var p256 BID_UINT256
	var cy1, cy2 uint64

	phl, qll := __mul_64x128_full(a.w[0], b)
	phh, qlh := __mul_64x128_full(a.w[1], b)

	p256.w[0] = qll.w[0]
	p256.w[1], cy1 = __add_carry_out(qlh.w[0], qll.w[1])
	p256.w[2], cy2 = __add_carry_in_out(qlh.w[1], phl, cy1)
	p256.w[3] = phh + cy2
	return p256
}

// unpack_BID64 unpacks a BID64 value into sign, exponent, and coefficient
// Returns coefficient (0 if NaN/Inf)
func unpack_BID64(x uint64) (sign uint64, exponent int, coefficient uint64, valid bool) {
	sign = x & 0x8000000000000000

	if (x & SPECIAL_ENCODING_MASK64) == SPECIAL_ENCODING_MASK64 {
		// special encodings
		coefficient = (x & LARGE_COEFF_MASK64) | LARGE_COEFF_HIGH_BIT64

		if (x & INFINITY_MASK64) == INFINITY_MASK64 {
			exponent = 0
			coefficient = x & 0xfe03ffffffffffff
			if (x & 0x0003ffffffffffff) >= 1000000000000000 {
				coefficient = x & 0xfe00000000000000
			}
			if (x & NAN_MASK64) == INFINITY_MASK64 {
				coefficient = x & SINFINITY_MASK64
			}
			return sign, exponent, coefficient, false // NaN or Infinity
		}
		// check for non-canonical values
		if coefficient >= 10000000000000000 {
			coefficient = 0
		}
		// get exponent
		tmp := x >> EXPONENT_SHIFT_LARGE64
		exponent = int(tmp & EXPONENT_MASK64)
		return sign, exponent, coefficient, coefficient != 0
	}
	// exponent
	tmp := x >> EXPONENT_SHIFT_SMALL64
	exponent = int(tmp & EXPONENT_MASK64)
	// coefficient
	coefficient = x & SMALL_COEFF_MASK64

	return sign, exponent, coefficient, coefficient != 0
}

// very_fast_get_BID64 packs sign, exponent, and coefficient into BID64
// No overflow/underflow checking, no rounding
func very_fast_get_BID64(sgn uint64, expon int, coeff uint64) uint64 {
	var r uint64
	mask := uint64(1) << EXPONENT_SHIFT_SMALL64

	// check whether coefficient fits in 10*5+3 bits
	if coeff < mask {
		r = uint64(expon)
		r <<= EXPONENT_SHIFT_SMALL64
		r |= (coeff | sgn)
		return r
	}
	// special format

	// eliminate the case coeff==10^16 after rounding
	if coeff == 10000000000000000 {
		r = uint64(expon + 1)
		r <<= EXPONENT_SHIFT_SMALL64
		r |= (1000000000000000 | sgn)
		return r
	}

	r = uint64(expon)
	r <<= EXPONENT_SHIFT_LARGE64
	r |= (sgn | SPECIAL_ENCODING_MASK64)
	// add coeff, without leading bits
	mask = (mask >> 2) - 1
	coeff &= mask
	r |= coeff

	return r
}

// fast_get_BID64 packs sign, exponent, and coefficient into BID64
// No overflow/underflow checking
func fast_get_BID64(sgn uint64, expon int, coeff uint64) uint64 {
	return very_fast_get_BID64(sgn, expon, coeff)
}

// fast_get_BID64_check_OF packs with overflow checking
// Ported from Intel BID library bid_internal.h
func fast_get_BID64_check_OF(sgn uint64, expon int, coeff uint64, rmode int) uint64 {
	var r uint64

	// 3 * 256 - 1 = 767 = DECIMAL_MAX_EXPON_64
	// 3 * 256 = 768
	if uint(expon) >= 3*256-1 {
		// Special case: expon == 767 and coeff == 10^16
		if expon == 3*256-1 && coeff == 10000000000000000 {
			expon = 3 * 256 // 768
			coeff = 1000000000000000
		}

		if uint(expon) >= 3*256 {
			// try to normalize coefficient
			for coeff < 1000000000000000 && expon >= 3*256 {
				expon--
				coeff = (coeff << 3) + (coeff << 1) // coeff * 10
			}

			if expon > DECIMAL_MAX_EXPON_64 {
				// overflow
				r = sgn | INFINITY_MASK64
				switch rmode {
				case BID_ROUNDING_DOWN:
					if sgn == 0 {
						r = LARGEST_BID64
					}
				case BID_ROUNDING_TO_ZERO:
					r = sgn | LARGEST_BID64
				case BID_ROUNDING_UP:
					if sgn != 0 {
						r = SMALLEST_BID64
					}
				}
				return r
			}
		}
	}

	mask := uint64(1) << EXPONENT_SHIFT_SMALL64

	// check whether coefficient fits in 10*5+3 bits
	if coeff < mask {
		r = uint64(expon)
		r <<= EXPONENT_SHIFT_SMALL64
		r |= (coeff | sgn)
		return r
	}

	// special format
	// eliminate the case coeff==10^16 after rounding
	if coeff == 10000000000000000 {
		r = uint64(expon + 1)
		r <<= EXPONENT_SHIFT_SMALL64
		r |= (1000000000000000 | sgn)
		return r
	}

	r = uint64(expon)
	r <<= EXPONENT_SHIFT_LARGE64
	r |= (sgn | SPECIAL_ENCODING_MASK64)
	// add coeff, without leading bits
	mask = (mask >> 2) - 1
	coeff &= mask
	r |= coeff

	return r
}

// get_BID64 packs with full overflow/underflow checking and rounding
func get_BID64(sgn uint64, expon int, coeff uint64, rmode int) uint64 {
	var r uint64

	if coeff > 9999999999999999 {
		expon++
		coeff = 1000000000000000
	}

	// check for possible underflow/overflow
	if uint(expon) >= 3*256 {
		if expon < 0 {
			// underflow
			if expon+MAX_FORMAT_DIGITS < 0 {
				if rmode == BID_ROUNDING_DOWN && sgn != 0 {
					return 0x8000000000000001
				}
				if rmode == BID_ROUNDING_UP && sgn == 0 {
					return 1
				}
				// result is 0
				return sgn
			}

			if sgn != 0 && (rmode == BID_ROUNDING_DOWN || rmode == BID_ROUNDING_UP) {
				if rmode == BID_ROUNDING_DOWN {
					rmode = BID_ROUNDING_UP
				} else {
					rmode = BID_ROUNDING_DOWN
				}
			}

			// get digits to be shifted out
			extraDigits := -expon
			coeff += bid_round_const_table[rmode][extraDigits]

			// get coeff*(2^M[extra_digits])/10^extra_digits
			qh, qLow := __mul_64x128_full(coeff, bid_reciprocals10_128[extraDigits])

			// now get P/10^extra_digits: shift Q_high right by M[extra_digits]-128
			amount := bid_recip_scale[extraDigits]
			c64 := qh >> uint(amount)

			if rmode == BID_ROUNDING_TO_NEAREST {
				if c64&1 != 0 {
					// check whether fractional part of initial_P/10^extra_digits is exactly .5
					// get remainder
					amount2 := 64 - amount
					remainderH := (^uint64(0)) >> uint(amount2)
					remainderH = remainderH & qh

					if remainderH == 0 &&
						(qLow.w[1] < bid_reciprocals10_128[extraDigits].w[1] ||
							(qLow.w[1] == bid_reciprocals10_128[extraDigits].w[1] &&
								qLow.w[0] < bid_reciprocals10_128[extraDigits].w[0])) {
						c64--
					}
				}
			} else if rmode == BID_ROUNDING_NEAREST_DOWN {
				// half_down: ties toward zero (always round down on tie)
				// check whether fractional part is exactly .5
				amount2 := 64 - amount
				remainderH := (^uint64(0)) >> uint(amount2)
				remainderH = remainderH & qh

				if remainderH == 0 &&
					(qLow.w[1] < bid_reciprocals10_128[extraDigits].w[1] ||
						(qLow.w[1] == bid_reciprocals10_128[extraDigits].w[1] &&
							qLow.w[0] < bid_reciprocals10_128[extraDigits].w[0])) {
					c64-- // tie → toward zero
				}
			}

			return sgn | c64
		}

		if coeff == 0 {
			if expon > DECIMAL_MAX_EXPON_64 {
				expon = DECIMAL_MAX_EXPON_64
			}
		}

		for coeff < 1000000000000000 && expon >= 3*256 {
			expon--
			coeff = (coeff << 3) + (coeff << 1)
		}

		if expon > DECIMAL_MAX_EXPON_64 {
			// overflow
			r = sgn | INFINITY_MASK64
			switch rmode {
			case BID_ROUNDING_DOWN:
				if sgn == 0 {
					r = LARGEST_BID64
				}
			case BID_ROUNDING_TO_ZERO:
				r = sgn | LARGEST_BID64
			case BID_ROUNDING_UP:
				if sgn != 0 {
					r = SMALLEST_BID64
				}
			}
			return r
		}
	}

	return very_fast_get_BID64(sgn, expon, coeff)
}

// getBinaryExponent extracts binary exponent from float64 representation
func getBinaryExponent(coefficient uint64) int {
	f := float64(coefficient)
	bits := math.Float64bits(f)
	return int((bits&MASK_BINARY_EXPONENT)>>52) - 0x3ff
}

// Additional constants for Bid64Mul
const (
	UPPER_EXPON_LIMIT = 51
)

// __tight_bin_range_128 tightens binary range of 128-bit value
// Ported from Intel BID library bid_internal.h
func __tight_bin_range_128(P BID_UINT128, binExpon int) int {
	var M uint64 = 1
	bp := binExpon

	if bp < 63 {
		M <<= uint(bp + 1)
		if P.w[0] >= M {
			bp++
		}
	} else if bp > 64 {
		M <<= uint(bp + 1 - 64)
		if P.w[1] > M || (P.w[1] == M && P.w[0] != 0) {
			bp++
		}
	} else if P.w[1] != 0 {
		bp++
	}
	return bp
}

// __mul_128x128_full multiplies two 128-bit values to get 256-bit result as two 128-bit parts
// Returns Qh (high 128 bits), Ql (low 128 bits)
// Ported from Intel BID library bid_internal.h
func __mul_128x128_full(A, B BID_UINT128) (Qh, Ql BID_UINT128) {
	ALBH := __mul_64x64_to_128(A.w[0], B.w[1])
	AHBL := __mul_64x64_to_128(B.w[0], A.w[1])
	ALBL := __mul_64x64_to_128(A.w[0], B.w[0])
	AHBH := __mul_64x64_to_128(A.w[1], B.w[1])

	QM := __add_128_128(ALBH, AHBL)
	Ql.w[0] = ALBL.w[0]
	QM2 := __add_128_64(QM, ALBL.w[1])
	Qh = __add_128_64(AHBH, QM2.w[1])
	Ql.w[1] = QM2.w[0]
	return
}

// very_fast_get_BID64_small_mantissa packs sign, exponent, and coefficient into BID64
// For small mantissas that don't require special encoding
// Ported from Intel BID library bid_internal.h
func very_fast_get_BID64_small_mantissa(sgn uint64, expon int, coeff uint64) uint64 {
	r := uint64(expon)
	r <<= EXPONENT_SHIFT_SMALL64
	r |= (coeff | sgn)
	return r
}

// get_BID64_small_mantissa packs with underflow/overflow handling for small mantissas
// Ported from Intel BID library bid_internal.h
func get_BID64_small_mantissa(sgn uint64, expon int, coeff uint64, rmode int) uint64 {
	var Q_low BID_UINT128
	var _C64, remainder_h, QH uint64
	var extra_digits, amount, amount2 int

	// check for possible underflow/overflow
	if uint(expon) >= 3*256 {
		if expon < 0 {
			// underflow
			if expon+MAX_FORMAT_DIGITS < 0 {
				if rmode == BID_ROUNDING_DOWN && sgn != 0 {
					return 0x8000000000000001
				}
				if rmode == BID_ROUNDING_UP && sgn == 0 {
					return 1
				}
				// result is 0
				return sgn
			}

			if sgn != 0 && uint(rmode-1) < 2 {
				rmode = 3 - rmode
			}

			// get digits to be shifted out
			extra_digits = -expon
			C128w0 := coeff + bid_round_const_table[rmode][extra_digits]

			// get coeff*(2^M[extra_digits])/10^extra_digits
			QH, Q_low = __mul_64x128_full(C128w0, bid_reciprocals10_128[extra_digits])

			// now get P/10^extra_digits: shift Q_high right by M[extra_digits]-128
			amount = bid_recip_scale[extra_digits]
			_C64 = QH >> uint(amount)

			if rmode == 0 { // BID_ROUNDING_TO_NEAREST
				if _C64&1 != 0 {
					// check whether fractional part is exactly .5
					amount2 = 64 - amount
					remainder_h = (^uint64(0)) >> uint(amount2)
					remainder_h = remainder_h & QH

					if remainder_h == 0 &&
						(Q_low.w[1] < bid_reciprocals10_128[extra_digits].w[1] ||
							(Q_low.w[1] == bid_reciprocals10_128[extra_digits].w[1] &&
								Q_low.w[0] < bid_reciprocals10_128[extra_digits].w[0])) {
						_C64--
					}
				}
			} else if rmode == BID_ROUNDING_NEAREST_DOWN {
				// half_down: ties toward zero
				amount2 = 64 - amount
				remainder_h = (^uint64(0)) >> uint(amount2)
				remainder_h = remainder_h & QH

				if remainder_h == 0 &&
					(Q_low.w[1] < bid_reciprocals10_128[extra_digits].w[1] ||
						(Q_low.w[1] == bid_reciprocals10_128[extra_digits].w[1] &&
							Q_low.w[0] < bid_reciprocals10_128[extra_digits].w[0])) {
					_C64-- // tie → toward zero
				}
			}

			return sgn | _C64
		}

		if coeff == 0 {
			if expon > DECIMAL_MAX_EXPON_64 {
				expon = DECIMAL_MAX_EXPON_64
			}
		}

		for coeff < 1000000000000000 && expon >= 3*256 {
			expon--
			coeff = (coeff << 3) + (coeff << 1)
		}

		if expon > DECIMAL_MAX_EXPON_64 {
			// overflow
			r := sgn | INFINITY_MASK64
			switch rmode {
			case BID_ROUNDING_DOWN:
				if sgn == 0 {
					r = LARGEST_BID64
				}
			case BID_ROUNDING_TO_ZERO:
				r = sgn | LARGEST_BID64
			case BID_ROUNDING_UP:
				if sgn != 0 {
					r = SMALLEST_BID64
				}
			}
			return r
		}
	}

	return very_fast_get_BID64_small_mantissa(sgn, expon, coeff)
}

// get_BID64_small_mantissa_flags packs with underflow/overflow handling for small mantissas
// and updates status flags in-place.
// Ported from Intel BID library bid_internal.h get_BID64_small_mantissa
func get_BID64_small_mantissa_flags(sgn uint64, expon int, coeff uint64, rmode int, fpsc *uint32) uint64 {
	var C128, Q_low, Stemp BID_UINT128
	var r, mask, _C64, remainder_h, QH, carry, CY uint64
	var extra_digits, amount, amount2 int
	var status uint32

	// check for possible underflow/overflow
	if uint(expon) >= 3*256 {
		if expon < 0 {
			// underflow
			if expon+MAX_FORMAT_DIGITS < 0 {
				*fpsc |= BID_UNDERFLOW_EXCEPTION | BID_INEXACT_EXCEPTION
				if rmode == BID_ROUNDING_DOWN && sgn != 0 {
					return 0x8000000000000001
				}
				if rmode == BID_ROUNDING_UP && sgn == 0 {
					return 1
				}
				// result is 0
				return sgn
			}

			if sgn != 0 && uint(rmode-1) < 2 {
				rmode = 3 - rmode
			}

			// get digits to be shifted out
			extra_digits = -expon
			C128.w[0] = coeff + bid_round_const_table[rmode][extra_digits]

			// get coeff*(2^M[extra_digits])/10^extra_digits
			QH, Q_low = __mul_64x128_full(C128.w[0], bid_reciprocals10_128[extra_digits])

			// now get P/10^extra_digits: shift Q_high right by M[extra_digits]-128
			amount = bid_recip_scale[extra_digits]

			_C64 = QH >> uint(amount)

			if rmode == 0 {
				if (_C64 & 1) != 0 {
					// check whether fractional part of initial_P/10^extra_digits is exactly .5

					// get remainder
					amount2 = 64 - amount
					remainder_h = 0
					remainder_h--
					remainder_h >>= uint(amount2)
					remainder_h = remainder_h & QH

					if remainder_h == 0 &&
						(Q_low.w[1] < bid_reciprocals10_128[extra_digits].w[1] ||
							(Q_low.w[1] == bid_reciprocals10_128[extra_digits].w[1] &&
								Q_low.w[0] < bid_reciprocals10_128[extra_digits].w[0])) {
						_C64--
					}
				}
			}

			if (*fpsc & BID_INEXACT_EXCEPTION) != 0 {
				*fpsc |= BID_UNDERFLOW_EXCEPTION
			} else {
				status = BID_INEXACT_EXCEPTION
				// get remainder
				remainder_h = QH << (64 - uint(amount))

				switch rmode {
				case BID_ROUNDING_TO_NEAREST:
					fallthrough
				case BID_ROUNDING_TIES_AWAY:
					// test whether fractional part is 0
					if remainder_h == 0x8000000000000000 &&
						(Q_low.w[1] < bid_reciprocals10_128[extra_digits].w[1] ||
							(Q_low.w[1] == bid_reciprocals10_128[extra_digits].w[1] &&
								Q_low.w[0] < bid_reciprocals10_128[extra_digits].w[0])) {
						status = BID_EXACT_STATUS
					}
				case BID_ROUNDING_DOWN:
					fallthrough
				case BID_ROUNDING_TO_ZERO:
					if remainder_h == 0 &&
						(Q_low.w[1] < bid_reciprocals10_128[extra_digits].w[1] ||
							(Q_low.w[1] == bid_reciprocals10_128[extra_digits].w[1] &&
								Q_low.w[0] < bid_reciprocals10_128[extra_digits].w[0])) {
						status = BID_EXACT_STATUS
					}
				default:
					// round up
					Stemp.w[0], CY = __add_carry_out(Q_low.w[0], bid_reciprocals10_128[extra_digits].w[0])
					Stemp.w[1], carry = __add_carry_in_out(Q_low.w[1], bid_reciprocals10_128[extra_digits].w[1], CY)
					if (remainder_h>>(64-uint(amount)))+carry >= (uint64(1) << uint(amount)) {
						status = BID_EXACT_STATUS
					}
				}

				if status != BID_EXACT_STATUS {
					*fpsc |= BID_UNDERFLOW_EXCEPTION | status
				}
			}

			return sgn | _C64
		}

		for coeff < 1000000000000000 && expon >= 3*256 {
			expon--
			coeff = (coeff << 3) + (coeff << 1)
		}
		if expon > DECIMAL_MAX_EXPON_64 {
			*fpsc |= BID_OVERFLOW_EXCEPTION | BID_INEXACT_EXCEPTION
			// overflow
			r = sgn | INFINITY_MASK64
			switch rmode {
			case BID_ROUNDING_DOWN:
				if sgn == 0 {
					r = LARGEST_BID64
				}
			case BID_ROUNDING_TO_ZERO:
				r = sgn | LARGEST_BID64
			case BID_ROUNDING_UP:
				// round up
				if sgn != 0 {
					r = SMALLEST_BID64
				}
			}
			return r
		} else {
			mask = 1
			mask <<= EXPONENT_SHIFT_SMALL64
			if coeff >= mask {
				r = uint64(expon)
				r <<= EXPONENT_SHIFT_LARGE64
				r |= sgn | SPECIAL_ENCODING_MASK64
				// add coeff, without leading bits
				mask = (mask >> 2) - 1
				coeff &= mask
				r |= coeff
				return r
			}
		}
	}

	r = uint64(expon)
	r <<= EXPONENT_SHIFT_SMALL64
	r |= coeff | sgn

	return r
}

// roundingModeToBID converts RoundingMode to BID rounding mode
func roundingModeToBID(mode RoundingMode) int {
	switch mode {
	case RoundNearestEven:
		return BID_ROUNDING_TO_NEAREST
	case RoundNearestAway:
		return BID_ROUNDING_TIES_AWAY
	case RoundTowardZero:
		return BID_ROUNDING_TO_ZERO
	case RoundTowardPositive:
		return BID_ROUNDING_UP
	case RoundTowardNegative:
		return BID_ROUNDING_DOWN
	case RoundNearestDown:
		return BID_ROUNDING_NEAREST_DOWN
	default:
		return BID_ROUNDING_TO_NEAREST
	}
}

// get_BID64_UF is called when underflow is known to occur
// Ported from: Intel BID library bid_internal.h
func get_BID64_UF(sgn uint64, expon int, coeff uint64, R uint64, rmode int) uint64 {
	var Q_low BID_UINT128
	var _C64, remainder_h, QH uint64
	var extra_digits, amount, amount2 int

	// underflow
	if expon+MAX_FORMAT_DIGITS < 0 {
		if rmode == BID_ROUNDING_DOWN && sgn != 0 {
			return 0x8000000000000001
		}
		if rmode == BID_ROUNDING_UP && sgn == 0 {
			return 1
		}
		// result is 0
		return sgn
	}

	// 10*coeff
	coeff = (coeff << 3) + (coeff << 1)
	if sgn != 0 && uint(rmode-1) < 2 {
		rmode = 3 - rmode
	}
	if R != 0 {
		coeff |= 1
	}

	// get digits to be shifted out
	extra_digits = 1 - expon
	C128w0 := coeff + bid_round_const_table[rmode][extra_digits]

	// get coeff*(2^M[extra_digits])/10^extra_digits
	QH, Q_low = __mul_64x128_full(C128w0, bid_reciprocals10_128[extra_digits])

	// now get P/10^extra_digits: shift Q_high right by M[extra_digits]-128
	amount = bid_recip_scale[extra_digits]
	_C64 = QH >> uint(amount)

	if rmode == 0 { // BID_ROUNDING_TO_NEAREST
		if _C64&1 != 0 {
			// check whether fractional part is exactly .5
			amount2 = 64 - amount
			remainder_h = (^uint64(0)) >> uint(amount2)
			remainder_h = remainder_h & QH

			if remainder_h == 0 &&
				(Q_low.w[1] < bid_reciprocals10_128[extra_digits].w[1] ||
					(Q_low.w[1] == bid_reciprocals10_128[extra_digits].w[1] &&
						Q_low.w[0] < bid_reciprocals10_128[extra_digits].w[0])) {
				_C64--
			}
		}
	} else if rmode == BID_ROUNDING_NEAREST_DOWN {
		// half_down: ties toward zero
		amount2 = 64 - amount
		remainder_h = (^uint64(0)) >> uint(amount2)
		remainder_h = remainder_h & QH

		if remainder_h == 0 &&
			(Q_low.w[1] < bid_reciprocals10_128[extra_digits].w[1] ||
				(Q_low.w[1] == bid_reciprocals10_128[extra_digits].w[1] &&
					Q_low.w[0] < bid_reciprocals10_128[extra_digits].w[0])) {
			_C64-- // tie → toward zero
		}
	}

	return sgn | _C64
}

// get_BID64_flags packs with full overflow/underflow checking and rounding
// Returns result and status flags
// Ported from: Intel BID library bid_internal.h get_BID64()
func get_BID64_flags(sgn uint64, expon int, coeff uint64, rmode int) (uint64, uint32) {
	var Q_low BID_UINT128
	var QH, r, _C64, remainder_h, CY, carry uint64
	var extra_digits, amount, amount2 int
	var status uint32

	if coeff > 9999999999999999 {
		expon++
		coeff = 1000000000000000
	}

	// check for possible underflow/overflow
	if uint(expon) >= 3*256 {
		if expon < 0 {
			// underflow
			if expon+MAX_FORMAT_DIGITS < 0 {
				status = BID_UNDERFLOW_EXCEPTION | BID_INEXACT_EXCEPTION
				if rmode == BID_ROUNDING_DOWN && sgn != 0 {
					return 0x8000000000000001, status
				}
				if rmode == BID_ROUNDING_UP && sgn == 0 {
					return 1, status
				}
				// result is 0
				return sgn, status
			}

			if sgn != 0 && uint(rmode-1) < 2 {
				rmode = 3 - rmode
			}

			// get digits to be shifted out
			extra_digits = -expon
			coeff += bid_round_const_table[rmode][extra_digits]

			// get coeff*(2^M[extra_digits])/10^extra_digits
			QH, Q_low = __mul_64x128_full(coeff, bid_reciprocals10_128[extra_digits])

			// now get P/10^extra_digits: shift Q_high right by M[extra_digits]-128
			amount = bid_recip_scale[extra_digits]
			_C64 = QH >> uint(amount)

			if rmode == 0 { // BID_ROUNDING_TO_NEAREST
				if _C64&1 != 0 {
					// check whether fractional part of initial_P/10^extra_digits is exactly .5
					// get remainder
					amount2 = 64 - amount
					remainder_h = (^uint64(0)) >> uint(amount2)
					remainder_h = remainder_h & QH

					if remainder_h == 0 &&
						(Q_low.w[1] < bid_reciprocals10_128[extra_digits].w[1] ||
							(Q_low.w[1] == bid_reciprocals10_128[extra_digits].w[1] &&
								Q_low.w[0] < bid_reciprocals10_128[extra_digits].w[0])) {
						_C64--
					}
				}
			}

			// Status flag determination for underflow
			status = BID_INEXACT_EXCEPTION

			// get remainder
			remainder_h = QH << (64 - uint(amount))

			switch rmode {
			case BID_ROUNDING_TO_NEAREST, BID_ROUNDING_TIES_AWAY:
				// test whether fractional part is 0
				if remainder_h == 0x8000000000000000 &&
					(Q_low.w[1] < bid_reciprocals10_128[extra_digits].w[1] ||
						(Q_low.w[1] == bid_reciprocals10_128[extra_digits].w[1] &&
							Q_low.w[0] < bid_reciprocals10_128[extra_digits].w[0])) {
					status = BID_EXACT_STATUS
				}
			case BID_ROUNDING_DOWN, BID_ROUNDING_TO_ZERO:
				if remainder_h == 0 &&
					(Q_low.w[1] < bid_reciprocals10_128[extra_digits].w[1] ||
						(Q_low.w[1] == bid_reciprocals10_128[extra_digits].w[1] &&
							Q_low.w[0] < bid_reciprocals10_128[extra_digits].w[0])) {
					status = BID_EXACT_STATUS
				}
			default:
				// round up
				var Stemp_w0 uint64
				Stemp_w0, CY = __add_carry_out(Q_low.w[0], bid_reciprocals10_128[extra_digits].w[0])
				_, carry = __add_carry_in_out(Q_low.w[1], bid_reciprocals10_128[extra_digits].w[1], CY)
				_ = Stemp_w0
				if (remainder_h>>(64-uint(amount)))+carry >= (uint64(1) << uint(amount)) {
					status = BID_EXACT_STATUS
				}
			}

			if status != BID_EXACT_STATUS {
				status = BID_UNDERFLOW_EXCEPTION | status
			}

			return sgn | _C64, status
		}

		if coeff == 0 {
			if expon > DECIMAL_MAX_EXPON_64 {
				expon = DECIMAL_MAX_EXPON_64
			}
		}

		for coeff < 1000000000000000 && expon >= 3*256 {
			expon--
			coeff = (coeff << 3) + (coeff << 1)
		}

		if expon > DECIMAL_MAX_EXPON_64 {
			status = BID_OVERFLOW_EXCEPTION | BID_INEXACT_EXCEPTION
			// overflow
			r = sgn | INFINITY_MASK64
			switch rmode {
			case BID_ROUNDING_DOWN:
				if sgn == 0 {
					r = LARGEST_BID64
				}
			case BID_ROUNDING_TO_ZERO:
				r = sgn | LARGEST_BID64
			case BID_ROUNDING_UP:
				if sgn != 0 {
					r = SMALLEST_BID64
				}
			}
			return r, status
		}
	}

	return very_fast_get_BID64(sgn, expon, coeff), 0
}

// __mul_192x192_to_384 multiplies two 192-bit values to get 384-bit result.
// Ported from bid_internal.h __mul_192x192_to_384 macro.
func __mul_192x192_to_384(a, b BID_UINT192) BID_UINT384 {
	var p BID_UINT384
	var cy uint64

	p00 := __mul_64x64_to_128(a.w[0], b.w[0])
	p01 := __mul_64x64_to_128(a.w[0], b.w[1])
	p02 := __mul_64x64_to_128(a.w[0], b.w[2])
	p10 := __mul_64x64_to_128(a.w[1], b.w[0])
	p11 := __mul_64x64_to_128(a.w[1], b.w[1])
	p12 := __mul_64x64_to_128(a.w[1], b.w[2])
	p20 := __mul_64x64_to_128(a.w[2], b.w[0])
	p21 := __mul_64x64_to_128(a.w[2], b.w[1])
	p22 := __mul_64x64_to_128(a.w[2], b.w[2])

	p.w[0] = p00.w[0]

	// w[1] = p00.w[1] + p01.w[0] + p10.w[0]
	p.w[1] = p00.w[1] + p01.w[0]
	cy = 0
	if p.w[1] < p00.w[1] {
		cy++
	}
	tmp := p.w[1]
	p.w[1] += p10.w[0]
	if p.w[1] < tmp {
		cy++
	}

	// w[2] = cy + p01.w[1] + p02.w[0] + p10.w[1] + p11.w[0] + p20.w[0]
	p.w[2] = cy + p01.w[1]
	cy = 0
	if p.w[2] < p01.w[1] {
		cy++
	}
	tmp = p.w[2]
	p.w[2] += p02.w[0]
	if p.w[2] < tmp {
		cy++
	}
	tmp = p.w[2]
	p.w[2] += p10.w[1]
	if p.w[2] < tmp {
		cy++
	}
	tmp = p.w[2]
	p.w[2] += p11.w[0]
	if p.w[2] < tmp {
		cy++
	}
	tmp = p.w[2]
	p.w[2] += p20.w[0]
	if p.w[2] < tmp {
		cy++
	}

	// w[3] = cy + p02.w[1] + p11.w[1] + p12.w[0] + p20.w[1] + p21.w[0]
	p.w[3] = cy + p02.w[1]
	cy = 0
	if p.w[3] < p02.w[1] {
		cy++
	}
	tmp = p.w[3]
	p.w[3] += p11.w[1]
	if p.w[3] < tmp {
		cy++
	}
	tmp = p.w[3]
	p.w[3] += p12.w[0]
	if p.w[3] < tmp {
		cy++
	}
	tmp = p.w[3]
	p.w[3] += p20.w[1]
	if p.w[3] < tmp {
		cy++
	}
	tmp = p.w[3]
	p.w[3] += p21.w[0]
	if p.w[3] < tmp {
		cy++
	}

	// w[4] = cy + p12.w[1] + p21.w[1] + p22.w[0]
	p.w[4] = cy + p12.w[1]
	cy = 0
	if p.w[4] < p12.w[1] {
		cy++
	}
	tmp = p.w[4]
	p.w[4] += p21.w[1]
	if p.w[4] < tmp {
		cy++
	}
	tmp = p.w[4]
	p.w[4] += p22.w[0]
	if p.w[4] < tmp {
		cy++
	}

	p.w[5] = cy + p22.w[1]

	return p
}

// __mul_256x256_to_512 multiplies two 256-bit values to get 512-bit result.
func __mul_256x256_to_512(a, b BID_UINT256) BID_UINT512 {
	// Split into 128-bit halves and use schoolbook multiplication
	// a = aH * 2^128 + aL, b = bH * 2^128 + bL
	// a*b = aH*bH*2^256 + (aH*bL + aL*bH)*2^128 + aL*bL
	var p BID_UINT512
	aL := BID_UINT128{w: [2]uint64{a.w[0], a.w[1]}}
	aH := BID_UINT128{w: [2]uint64{a.w[2], a.w[3]}}
	bL := BID_UINT128{w: [2]uint64{b.w[0], b.w[1]}}
	bH := BID_UINT128{w: [2]uint64{b.w[2], b.w[3]}}

	p0 := __mul_128x128_to_256(aL, bL) // aL * bL
	p1 := __mul_128x128_to_256(aH, bL) // aH * bL
	p2 := __mul_128x128_to_256(aL, bH) // aL * bH
	p3 := __mul_128x128_to_256(aH, bH) // aH * bH

	// p = p0 + (p1+p2)<<128 + p3<<256
	p.w[0] = p0.w[0]
	p.w[1] = p0.w[1]

	var cy uint64
	// Add p1 shifted by 128 bits
	p.w[2] = p0.w[2] + p1.w[0]
	cy = 0
	if p.w[2] < p0.w[2] {
		cy = 1
	}

	p.w[3] = p0.w[3] + p1.w[1] + cy
	cy = 0
	if p.w[3] < p1.w[1] || (cy == 0 && p.w[3] < p0.w[3]) {
		cy = 1
	}

	c4 := p1.w[2] + cy
	cy = 0
	if c4 < p1.w[2] {
		cy = 1
	}
	c5 := p1.w[3] + cy

	// Add p2 shifted by 128 bits
	tmp := p.w[2]
	p.w[2] += p2.w[0]
	cy = 0
	if p.w[2] < tmp {
		cy = 1
	}

	tmp = p.w[3]
	p.w[3] += p2.w[1] + cy
	cy = 0
	if p.w[3] < tmp || (p.w[3] == tmp && cy > 0) {
		cy = 1
	}

	tmp = c4
	c4 += p2.w[2] + cy
	cy = 0
	if c4 < tmp || (c4 == tmp && cy > 0) {
		cy = 1
	}

	c5 += p2.w[3] + cy

	// Add p3 shifted by 256 bits
	p.w[4] = c4 + p3.w[0]
	cy = 0
	if p.w[4] < c4 {
		cy = 1
	}

	p.w[5] = c5 + p3.w[1] + cy
	cy = 0
	if p.w[5] < c5 || (p.w[5] == c5 && cy > 0) {
		cy = 1
	}

	p.w[6] = p3.w[2] + cy
	cy = 0
	if p.w[6] < p3.w[2] {
		cy = 1
	}

	p.w[7] = p3.w[3] + cy

	return p
}

// __mul_64x128_to_128 multiplies a 64-bit value by a 128-bit value
// and returns only the low 128 bits.
// Ported from bid_internal.h __mul_128x64_to_128 and __mul_64x128_to_128 macros.
func __mul_64x128_to_128(a uint64, b BID_UINT128) BID_UINT128 {
	_, ql := __mul_64x128_full(a, b)
	return ql
}

// __mul_64x192_to_256 multiplies 64-bit by 192-bit to get 256-bit result.
// Ported from Intel BID library bid_internal.h
func __mul_64x192_to_256(A uint64, B BID_UINT192) BID_UINT256 {
	var P BID_UINT256
	var c uint64
	lP0 := __mul_64x64_to_128(A, B.w[0])
	lP1 := __mul_64x64_to_128(A, B.w[1])
	lP2 := __mul_64x64_to_128(A, B.w[2])
	P.w[0] = lP0.w[0]
	P.w[1], c = __add_carry_out(lP1.w[0], lP0.w[1])
	P.w[2], c = __add_carry_in_out(lP2.w[0], lP1.w[1], c)
	P.w[3] = lP2.w[1] + c
	return P
}

// __mul_64x256_to_320 multiplies 64-bit by 256-bit to get 320-bit result.
// Ported from Intel BID library bid_internal.h
func __mul_64x256_to_320(A uint64, B BID_UINT256) BID_UINT320 {
	var P BID_UINT320
	var c uint64
	lP0 := __mul_64x64_to_128(A, B.w[0])
	lP1 := __mul_64x64_to_128(A, B.w[1])
	lP2 := __mul_64x64_to_128(A, B.w[2])
	lP3 := __mul_64x64_to_128(A, B.w[3])
	P.w[0] = lP0.w[0]
	P.w[1], c = __add_carry_out(lP1.w[0], lP0.w[1])
	P.w[2], c = __add_carry_in_out(lP2.w[0], lP1.w[1], c)
	P.w[3], c = __add_carry_in_out(lP3.w[0], lP2.w[1], c)
	P.w[4] = lP3.w[1] + c
	return P
}

// __mul_64x320_to_384 multiplies 64-bit by 320-bit to get 384-bit result.
// Ported from Intel BID library bid_internal.h
func __mul_64x320_to_384(A uint64, B BID_UINT320) BID_UINT384 {
	var P BID_UINT384
	var c uint64
	lP0 := __mul_64x64_to_128(A, B.w[0])
	lP1 := __mul_64x64_to_128(A, B.w[1])
	lP2 := __mul_64x64_to_128(A, B.w[2])
	lP3 := __mul_64x64_to_128(A, B.w[3])
	lP4 := __mul_64x64_to_128(A, B.w[4])
	P.w[0] = lP0.w[0]
	P.w[1], c = __add_carry_out(lP1.w[0], lP0.w[1])
	P.w[2], c = __add_carry_in_out(lP2.w[0], lP1.w[1], c)
	P.w[3], c = __add_carry_in_out(lP3.w[0], lP2.w[1], c)
	P.w[4], c = __add_carry_in_out(lP4.w[0], lP3.w[1], c)
	P.w[5] = lP4.w[1] + c
	return P
}

// __sqr128_to_256 squares a 128-bit number to get a 256-bit result.
// Ported from Intel BID library bid_internal.h
func __sqr128_to_256(A BID_UINT128) BID_UINT256 {
	var P256 BID_UINT256
	var c1, c2 uint64
	Qhh := __mul_64x64_to_128(A.w[1], A.w[1])
	Qlh := __mul_64x64_to_128(A.w[0], A.w[1])
	Qhh.w[1] += (Qlh.w[1] >> 63)
	Qlh.w[1] = (Qlh.w[1] + Qlh.w[1]) | (Qlh.w[0] >> 63)
	Qlh.w[0] += Qlh.w[0]
	Qll := __mul_64x64_to_128(A.w[0], A.w[0])

	P256.w[1], c1 = __add_carry_out(Qlh.w[0], Qll.w[1])
	P256.w[0] = Qll.w[0]
	P256.w[2], c2 = __add_carry_in_out(Qlh.w[1], Qhh.w[0], c1)
	P256.w[3] = Qhh.w[1] + c2
	return P256
}

// bid_get_BID128_fast packs BID128 without overflow/underflow checking.
// Ported from Intel BID library bid_internal.h
func bid_get_BID128_fast(sgn uint64, expon int, coeff BID_UINT128) BID_UINT128 {
	var res BID_UINT128
	if coeff.w[1] == 0x0001ed09bead87c0 && coeff.w[0] == 0x378d8e6400000000 {
		expon++
		coeff.w[1] = 0x0000314dc6448d93
		coeff.w[0] = 0x38c15b0a00000000
	}
	res.w[0] = coeff.w[0]
	tmp := uint64(expon)
	tmp <<= 49
	res.w[1] = sgn | tmp | coeff.w[1]
	return res
}

// noFmaMulAddF64 computes a*b + c without hardware FMA.
// Go's ARM64 compiler fuses a*b+c into FMADDD, which produces different
// rounding than separate MUL+ADD (as Intel C expects). This forces
// the multiply result through Float64bits to prevent fusion.
func noFmaMulAddF64(a, b, c float64) float64 {
	return math.Float64frombits(math.Float64bits(a*b)) + c
}

// noFmaMulAddF32 computes float32(a)*b + float32(c) without hardware FMA.
func noFmaMulAddF32(a float32, b float32, c float32) float32 {
	return math.Float32frombits(math.Float32bits(a*b)) + c
}
