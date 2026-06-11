// Ported from: IntelRDFPMathLib20U4/LIBRARY/src/bid64_string.c
// Function: bid64_from_string (lines 245-538)
// Version: Intel(R) Decimal Floating-Point Math Library 2.0 Update 4
//
// This file is a MECHANICAL LINE-BY-LINE translation of the Intel BID library to Go.
// All logic, magic numbers, variable names, and control flow are preserved exactly,
// with one documented IEEE-conformance deviation: the directed-rounding overflow on
// the no-exponent from_string path follows IEEE 754 (largest finite) instead of
// pinned Intel C (which ignores the rounding mode and returns Inf). See
// IEEE754_SPEC.md "pinned Intel BID C 대비 의도적 IEEE 편차".
// DO NOT REFACTOR OR "IMPROVE" THIS CODE beyond that documented deviation.

package bidgo

// tolower_macro from bid_internal.h line 94
// #define tolower_macro(x) (((unsigned char)((x)-'A')<=('Z'-'A'))?((x)-'A'+'a'):(x))
func tolower_macro(x byte) byte {
	if x >= 'A' && x <= 'Z' {
		return x + ('a' - 'A')
	}
	return x
}

// bid64_from_string - Mechanical port of Intel bid64_from_string
// Original: bid64_string.c lines 245-538
//
// This uses a byte slice approach to closely mirror C pointer semantics.
// ps is treated like a C char* pointer that can be incremented.
func bid64_from_string(str string, rnd_mode int) (res uint64, pfpsf uint32) {
	// Convert string to byte slice for C-like pointer semantics
	ps := []byte(str)
	// Add null terminator for C-like string handling
	ps = append(ps, 0)

	var sign_x, coefficient_x, rounded uint64
	var expon_x, sgn_expon, ndigits, add_expon, midpoint, rounded_up, dround int
	var dec_expon_scale, right_radix_leading_zeros, rdx_pt_enc int
	var c byte

	// line 272-274: eliminate leading whitespace
	// while (((*ps == ' ') || (*ps == '\t')) && (*ps))
	//     ps++;
	for (ps[0] == ' ' || ps[0] == '\t') && ps[0] != 0 {
		ps = ps[1:]
	}

	// line 277: get first non-whitespace character
	c = ps[0]

	// line 280: detect special cases (INF or NaN)
	// if (!c || (c != '.' && c != '-' && c != '+' && (c < '0' || c > '9')))
	if c == 0 || (c != '.' && c != '-' && c != '+' && (c < '0' || c > '9')) {
		// line 282-290: Infinity?
		if tolower_macro(ps[0]) == 'i' && tolower_macro(ps[1]) == 'n' &&
			tolower_macro(ps[2]) == 'f' && (ps[3] == 0 ||
			(tolower_macro(ps[3]) == 'i' &&
				tolower_macro(ps[4]) == 'n' && tolower_macro(ps[5]) == 'i' &&
				tolower_macro(ps[6]) == 't' && tolower_macro(ps[7]) == 'y' &&
				ps[8] == 0)) {
			res = 0x7800000000000000
			return
		}
		// line 292-296: return sNaN
		if tolower_macro(ps[0]) == 's' && tolower_macro(ps[1]) == 'n' &&
			tolower_macro(ps[2]) == 'a' && tolower_macro(ps[3]) == 'n' {
			res = 0x7e00000000000000
			return
		}
		// line 298-300: return qNaN
		res = 0x7c00000000000000
		return
	}

	// line 304-316: detect +INF or -INF
	if (tolower_macro(ps[1]) == 'i' && tolower_macro(ps[2]) == 'n' &&
		tolower_macro(ps[3]) == 'f') && (ps[4] == 0 ||
		(tolower_macro(ps[4]) == 'i' && tolower_macro(ps[5]) == 'n' &&
			tolower_macro(ps[6]) == 'i' && tolower_macro(ps[7]) == 't' &&
			tolower_macro(ps[8]) == 'y' && ps[9] == 0)) {
		if c == '+' {
			res = 0x7800000000000000
		} else if c == '-' {
			res = 0xf800000000000000
		} else {
			res = 0x7c00000000000000
		}
		return
	}

	// line 318-324: if +sNaN, +SNaN, -sNaN, or -SNaN
	if tolower_macro(ps[1]) == 's' && tolower_macro(ps[2]) == 'n' &&
		tolower_macro(ps[3]) == 'a' && tolower_macro(ps[4]) == 'n' {
		if c == '-' {
			res = 0xfe00000000000000
		} else {
			res = 0x7e00000000000000
		}
		return
	}

	// line 327-330: determine sign
	if c == '-' {
		sign_x = 0x8000000000000000
	} else {
		sign_x = 0
	}

	// line 333-336: get next character if leading +/- sign
	if c == '-' || c == '+' {
		ps = ps[1:]
		c = ps[0]
	}

	// line 338-342: if c isn't a decimal point or a decimal digit, return NaN
	if c != '.' && (c < '0' || c > '9') {
		res = 0x7c00000000000000 | sign_x
		return
	}

	// line 344
	rdx_pt_enc = 0

	// line 347-388: detect zero (and eliminate/ignore leading zeros)
	if ps[0] == '0' || ps[0] == '.' {
		// line 349-352
		if ps[0] == '.' {
			rdx_pt_enc = 1
			ps = ps[1:]
		}

		// line 355-387: while (*ps == '0')
		for ps[0] == '0' {
			ps = ps[1:]
			// line 359-361
			if rdx_pt_enc != 0 {
				right_radix_leading_zeros++
			}
			// line 364-386
			if ps[0] == '.' {
				if rdx_pt_enc == 0 {
					rdx_pt_enc = 1
					// line 369-373
					if ps[1] == 0 {
						res = (uint64(398-right_radix_leading_zeros) << 53) | sign_x
						return
					}
					ps = ps[1:]
				} else {
					// line 377-379: if 2 radix points, return NaN
					res = 0x7c00000000000000 | sign_x
					return
				}
			} else if ps[0] == 0 {
				// line 381-385
				res = (uint64(398-right_radix_leading_zeros) << 53) | sign_x
				return
			}
		}
	}

	// line 390
	c = ps[0]

	// line 392
	ndigits = 0

	// line 393-468: while ((c >= '0' && c <= '9') || c == '.')
	for (c >= '0' && c <= '9') || c == '.' {
		// line 394-404
		if c == '.' {
			if rdx_pt_enc != 0 {
				res = 0x7c00000000000000 | sign_x
				return
			}
			rdx_pt_enc = 1
			ps = ps[1:]
			c = ps[0]
			continue
		}

		// line 405
		dec_expon_scale += rdx_pt_enc

		// line 407
		ndigits++

		if ndigits <= 16 {
			// line 408-410
			coefficient_x = (coefficient_x << 1) + (coefficient_x << 3)
			coefficient_x += uint64(c - '0')
		} else if ndigits == 17 {
			// line 411-442: coefficient rounding
			// CRITICAL: Intel's switch has Duff's device structure
			// case DOWN/UP/TIES_AWAY are INSIDE the if block for TO_NEAREST
			// Only TO_NEAREST with condition=false runs the overflow check
			// TO_ZERO (3) has no case, so switch is skipped entirely

			doOverflowCheck := false

			switch rnd_mode {
			case BID_ROUNDING_TO_NEAREST: // 0
				// line 415
				if c == '5' && (coefficient_x&1) == 0 {
					midpoint = 1
				} else {
					midpoint = 0
				}
				// line 420-423
				if c > '5' || (c == '5' && (coefficient_x&1) != 0) {
					coefficient_x++
					rounded_up = 1
					// break - no overflow check
				} else {
					// condition false: will run overflow check
					doOverflowCheck = true
				}

			case BID_ROUNDING_DOWN: // 1
				// line 425-427
				if sign_x != 0 {
					if c > '0' {
						coefficient_x++
						rounded_up = 1
					} else {
						dround = 1
					}
				}
				// break - no overflow check

			case BID_ROUNDING_UP: // 2
				// line 428-430
				if sign_x == 0 {
					if c > '0' {
						coefficient_x++
						rounded_up = 1
					} else {
						dround = 1
					}
				}
				// break - no overflow check

			case BID_ROUNDING_TIES_AWAY: // 4
				// line 431-433
				if c >= '5' {
					coefficient_x++
					rounded_up = 1
				}
				// break - no overflow check

			default:
				// BID_ROUNDING_TO_ZERO (3) has no case in Intel code
				// switch is effectively skipped
			}

			// line 435-438: overflow check - only for TO_NEAREST with condition false
			if doOverflowCheck {
				if coefficient_x == 10000000000000000 {
					coefficient_x = 1000000000000000
					add_expon = 1
				}
			}

			// line 440-441
			if c > '0' {
				rounded = 1
			}
			// line 442
			add_expon += 1

		} else { // ndigits > 17
			// line 443-464
			add_expon++
			// line 445-448
			if midpoint != 0 && c > '0' {
				coefficient_x++
				midpoint = 0
				rounded_up = 1
			}
			// line 450-464
			if c > '0' {
				rounded = 1
				// line 453-463
				if dround != 0 {
					dround = 0
					coefficient_x++
					rounded_up = 1
					// line 459-462
					if coefficient_x == 10000000000000000 {
						coefficient_x = 1000000000000000
						add_expon++
					}
				}
			}
		}

		// line 466-467
		ps = ps[1:]
		c = ps[0]
	}

	// line 470
	add_expon -= (dec_expon_scale + right_radix_leading_zeros)

	// line 472-482
	if c == 0 {
		if rounded != 0 {
			pfpsf |= BID_INEXACT_EXCEPTION
		}
		res = fast_get_BID64_check_OF_withFlags(sign_x,
			add_expon+DECIMAL_EXPONENT_BIAS,
			coefficient_x, rnd_mode, &pfpsf)
		return
	}

	// line 484-488
	if c != 'E' && c != 'e' {
		res = 0x7c00000000000000 | sign_x
		return
	}

	// line 489-490
	ps = ps[1:]
	c = ps[0]

	// line 491
	if c == '-' {
		sgn_expon = 1
	} else {
		sgn_expon = 0
	}

	// line 492-495
	if c == '-' || c == '+' {
		ps = ps[1:]
		c = ps[0]
	}

	// line 496-500
	if c == 0 || c < '0' || c > '9' {
		res = 0x7c00000000000000 | sign_x
		return
	}

	// line 502-509
	for c >= '0' && c <= '9' {
		if expon_x < (1 << 20) {
			expon_x = (expon_x << 1) + (expon_x << 3)
			expon_x += int(c - '0')
		}
		ps = ps[1:]
		c = ps[0]
	}

	// line 511-515
	if c != 0 {
		res = 0x7c00000000000000 | sign_x
		return
	}

	// line 517-520
	if rounded != 0 {
		pfpsf |= BID_INEXACT_EXCEPTION
	}

	// line 522-523
	if sgn_expon != 0 {
		expon_x = -expon_x
	}

	// line 525
	expon_x += add_expon + DECIMAL_EXPONENT_BIAS

	// line 527-534
	if expon_x < 0 {
		if rounded_up != 0 {
			coefficient_x--
		}
		rnd_mode = 0
		res = get_BID64_UF_withFlags(sign_x, expon_x, coefficient_x, rounded, rnd_mode, &pfpsf)
		return
	}

	// line 536
	res = get_BID64_withFlags(sign_x, expon_x, coefficient_x, rnd_mode, &pfpsf)
	return
}

// fast_get_BID64_check_OF_withFlags - version with flag output
// Ported from: bid_internal.h fast_get_BID64_check_OF
func fast_get_BID64_check_OF_withFlags(sgn uint64, expon int, coeff uint64, rmode int, pfpsf *uint32) uint64 {
	var r, mask uint64

	if uint(expon) >= 3*256-1 {
		if expon == 3*256-1 && coeff == 10000000000000000 {
			expon = 3 * 256
			coeff = 1000000000000000
		}

		if uint(expon) >= 3*256 {
			for coeff < 1000000000000000 && expon >= 3*256 {
				expon--
				coeff = (coeff << 3) + (coeff << 1)
			}
			if expon > DECIMAL_MAX_EXPON_64 {
				*pfpsf |= BID_OVERFLOW_EXCEPTION | BID_INEXACT_EXCEPTION
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

	mask = 1
	mask <<= EXPONENT_SHIFT_SMALL64

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

// get_BID64_UF_withFlags - version with flag output
// This pack function is used when underflow is known to occur
// Ported from: bid_internal.h get_BID64_UF
func get_BID64_UF_withFlags(sgn uint64, expon int, coeff uint64, R uint64, rmode int, pfpsf *uint32) uint64 {
	var Q_low BID_UINT128
	var Stemp BID_UINT128
	var _C64, remainder_h, QH, carry, CY uint64
	var extra_digits, amount, amount2 int
	var status uint32

	// underflow
	if expon+MAX_FORMAT_DIGITS < 0 {
		*pfpsf |= BID_UNDERFLOW_EXCEPTION | BID_INEXACT_EXCEPTION
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
	}

	// Set status flags
	if (*pfpsf & BID_INEXACT_EXCEPTION) != 0 {
		*pfpsf |= BID_UNDERFLOW_EXCEPTION
	} else {
		status = BID_INEXACT_EXCEPTION
		// get remainder
		remainder_h = QH << uint(64-amount)

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
			Stemp.w[0], CY = __add_carry_out(Q_low.w[0], bid_reciprocals10_128[extra_digits].w[0])
			Stemp.w[1], carry = __add_carry_in_out(Q_low.w[1], bid_reciprocals10_128[extra_digits].w[1], CY)
			_ = Stemp
			if (remainder_h>>uint(64-amount))+carry >= (uint64(1) << uint(amount)) {
				status = BID_EXACT_STATUS
			}
		}

		if status != BID_EXACT_STATUS {
			*pfpsf |= BID_UNDERFLOW_EXCEPTION | status
		}
	}

	return sgn | _C64
}

// get_BID64_withFlags - version with flag output
// Ported from: bid_internal.h get_BID64
func get_BID64_withFlags(sgn uint64, expon int, coeff uint64, rmode int, pfpsf *uint32) uint64 {
	var Q_low BID_UINT128
	var Stemp BID_UINT128
	var QH, r, mask, _C64, remainder_h, CY, carry uint64
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
				*pfpsf |= BID_UNDERFLOW_EXCEPTION | BID_INEXACT_EXCEPTION
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
			}

			// Set status flags
			if (*pfpsf & BID_INEXACT_EXCEPTION) != 0 {
				*pfpsf |= BID_UNDERFLOW_EXCEPTION
			} else {
				status = BID_INEXACT_EXCEPTION
				// get remainder
				remainder_h = QH << uint(64-amount)

				switch rmode {
				case BID_ROUNDING_TO_NEAREST, BID_ROUNDING_TIES_AWAY:
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
					Stemp.w[0], CY = __add_carry_out(Q_low.w[0], bid_reciprocals10_128[extra_digits].w[0])
					Stemp.w[1], carry = __add_carry_in_out(Q_low.w[1], bid_reciprocals10_128[extra_digits].w[1], CY)
					_ = Stemp
					if (remainder_h>>uint(64-amount))+carry >= (uint64(1) << uint(amount)) {
						status = BID_EXACT_STATUS
					}
				}

				if status != BID_EXACT_STATUS {
					*pfpsf |= BID_UNDERFLOW_EXCEPTION | status
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
			*pfpsf |= BID_OVERFLOW_EXCEPTION | BID_INEXACT_EXCEPTION
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

	mask = 1
	mask <<= EXPONENT_SHIFT_SMALL64

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

// Bid64FromString is the public interface for bid64_from_string
// Returns (result, flags)
func Bid64FromString(s string, rndMode int) (uint64, uint32) {
	return bid64_from_string(s, rndMode)
}
