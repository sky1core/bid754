// Ported from: IntelRDFPMathLib20U4/LIBRARY/src/bid128_string.c
// Functions: bid128_to_string, bid128_from_string
// Version: Intel(R) Decimal Floating-Point Math Library 2.0 Update 4
//
// This file is a MECHANICAL LINE-BY-LINE translation of the Intel BID library to Go.
// All logic, magic numbers, variable names, and control flow are preserved exactly.
// DO NOT REFACTOR OR "IMPROVE" THIS CODE.

package bidgo

import "fmt"

const (
	MAX_FORMAT_DIGITS_128_str = 34
	MAX_STRING_DIGITS_128_str = 100
)

// l1_Split_MiDi_6 - __L1_Split_MiDi_6 macro from bid128_2_str_macros.h
func l1_Split_MiDi_6(X uint64, MiDi []uint32, ptr *int) {
	L1_Xhi_64 := ((X >> 28) * bid_Inv_Tento9) >> 33
	L1_Xlo_64 := X - L1_Xhi_64*uint64(bid_Tento9)
	if L1_Xlo_64 >= uint64(bid_Tento9) {
		L1_Xlo_64 -= uint64(bid_Tento9)
		L1_Xhi_64 += 1
	}
	L1_X_hi := uint32(L1_Xhi_64)
	L1_X_lo := uint32(L1_Xlo_64)
	l0_Split_MiDi_3(uint64(L1_X_hi), MiDi, ptr)
	l0_Split_MiDi_3(uint64(L1_X_lo), MiDi, ptr)
}

// handle_UF_128 and bid_get_BID128 are in bid128_div.go

// Bid128ToString - bid128_to_string
// Ported from: bid128_string.c lines 40-269
func Bid128ToString(x BID_UINT128) string {
	var str [128]byte
	var k uint = 0
	var d0, d123 uint
	var zero_digit uint = uint('0')
	var HI_18Dig, LO_18Dig, Tmp uint64
	var MiDi [12]uint32
	var midi_ind, k_lcv, length int

	// check for NaN or Infinity
	if (x.w[1] & MASK_SPECIAL128) == MASK_SPECIAL128 {
		// x is special
		if (x.w[1] & NAN_MASK64) == NAN_MASK64 { // x is NAN
			if (x.w[1] & SNAN_MASK64) == SNAN_MASK64 { // x is SNAN
				if int64(x.w[1]) < 0 {
					str[0] = '-'
				} else {
					str[0] = '+'
				}
				str[1] = 'S'
				str[2] = 'N'
				str[3] = 'a'
				str[4] = 'N'
				return string(str[:5])
			} else { // x is QNaN
				if int64(x.w[1]) < 0 {
					str[0] = '-'
				} else {
					str[0] = '+'
				}
				str[1] = 'N'
				str[2] = 'a'
				str[3] = 'N'
				return string(str[:4])
			}
		} else { // x is not a NaN, so it must be infinity
			if (x.w[1] & 0x8000000000000000) == 0 { // x is +inf
				str[0] = '+'
				str[1] = 'I'
				str[2] = 'n'
				str[3] = 'f'
				return string(str[:4])
			} else { // x is -inf
				str[0] = '-'
				str[1] = 'I'
				str[2] = 'n'
				str[3] = 'f'
				return string(str[:4])
			}
		}
	} else if ((x.w[1] & MASK_COEFF128) == 0) && (x.w[0] == 0) {
		// x is 0
		length = 0

		// determine if +/-
		if x.w[1]&0x8000000000000000 != 0 {
			str[length] = '-'
		} else {
			str[length] = '+'
		}
		length++
		str[length] = '0'
		length++
		str[length] = 'E'
		length++

		// extract the exponent and print
		exp := int((x.w[1]&MASK_EXP128)>>49) - 6176
		if exp > ((0x5ffe >> 1) - (6176)) {
			exp = int(((x.w[1]<<2)&MASK_EXP128)>>49) - 6176
		}
		if exp >= 0 {
			str[length] = '+'
			length++
			s := fmt.Sprintf("%d", exp)
			copy(str[length:], s)
			length += len(s)
		} else {
			s := fmt.Sprintf("%d", exp)
			copy(str[length:], s)
			length += len(s)
		}
		return string(str[:length])
	} else { // x is not special and is not zero
		// unpack x
		x_sign := x.w[1] & 0x8000000000000000 // 0 for positive, MASK_SIGN for negative
		x_exp := x.w[1] & MASK_EXP128         // biased and shifted left 49 bit positions
		if (x.w[1] & 0x6000000000000000) == 0x6000000000000000 {
			x_exp = (x.w[1] << 2) & MASK_EXP128 // biased and shifted left 49 bit positions
		}
		var C1 BID_UINT128
		C1.w[1] = x.w[1] & MASK_COEFF128
		C1.w[0] = x.w[0]
		exp := int(x_exp>>49) - 6176
		_ = x_sign

		// determine sign's representation as a char
		if x_sign != 0 {
			str[k] = '-' // negative number
		} else {
			str[k] = '+' // positive number
		}
		k++

		// determine coefficient's representation as a decimal string

		// if zero or non-canonical, set coefficient to '0'
		if (C1.w[1] > 0x0001ed09bead87c0) ||
			(C1.w[1] == 0x0001ed09bead87c0 &&
				(C1.w[0] > 0x378d8e63ffffffff)) ||
			((x.w[1] & 0x6000000000000000) == 0x6000000000000000) ||
			((C1.w[1] == 0) && (C1.w[0] == 0)) {
			str[k] = '0'
			k++
		} else {
			Tmp = C1.w[0] >> 59
			LO_18Dig = (C1.w[0] << 5) >> 5
			Tmp += (C1.w[1] << 5)
			HI_18Dig = 0
			k_lcv = 0

			for Tmp != 0 {
				midi_ind = int(Tmp & 0x000000000000003F)
				midi_ind <<= 1
				Tmp >>= 6
				HI_18Dig += mod10_18_tbl[k_lcv][midi_ind]
				midi_ind++
				LO_18Dig += mod10_18_tbl[k_lcv][midi_ind]
				k_lcv++
				l0_Normalize_10to18(&HI_18Dig, &LO_18Dig)
			}
			ptr := 0
			if HI_18Dig == 0 {
				l1_Split_MiDi_6_Lead(LO_18Dig, MiDi[:], &ptr)
			} else {
				l1_Split_MiDi_6_Lead(HI_18Dig, MiDi[:], &ptr)
				l1_Split_MiDi_6(LO_18Dig, MiDi[:], &ptr)
			}
			length = ptr
			c_ptr_start := int(k)
			c_ptr := c_ptr_start

			// now convert the MiDi into character strings
			l0_MiDi2Str_Lead(MiDi[0], str[:], &c_ptr)
			for k_lcv = 1; k_lcv < length; k_lcv++ {
				l0_MiDi2Str(MiDi[k_lcv], str[:], &c_ptr)
			}
			k = k + uint(c_ptr-c_ptr_start)
		}

		// print E and sign of exponent
		str[k] = 'E'
		k++
		if exp < 0 {
			exp = -exp
			str[k] = '-'
		} else {
			str[k] = '+'
		}
		k++

		// determine exponent's representation as a decimal string
		// d0 = exp / 1000;
		// Use Property 1
		d0 = uint((uint(exp) * 0x418a) >> 24) // 0x418a * 2^-24 = (10^(-3))RP,15
		d123 = uint(exp) - 1000*d0

		if d0 != 0 { // 1000 <= exp <= 6144 => 4 digits to return
			str[k] = byte(d0 + zero_digit) // ASCII for decimal digit d0
			k++
			str[k] = bid_midi_tbl[d123][0]
			k++
			str[k] = bid_midi_tbl[d123][1]
			k++
			str[k] = bid_midi_tbl[d123][2]
			k++
		} else { // 0 <= exp <= 999 => d0 = 0
			if d123 < 10 { // 0 <= exp <= 9 => 1 digit to return
				str[k] = byte(d123 + zero_digit) // ASCII
				k++
			} else if d123 < 100 { // 10 <= exp <= 99 => 2 digits to return
				str[k] = bid_midi_tbl[d123][1]
				k++
				str[k] = bid_midi_tbl[d123][2]
				k++
			} else { // 100 <= exp <= 999 => 3 digits to return
				str[k] = bid_midi_tbl[d123][0]
				k++
				str[k] = bid_midi_tbl[d123][1]
				k++
				str[k] = bid_midi_tbl[d123][2]
				k++
			}
		}
	}

	return string(str[:k])
}

// Bid128FromString - bid128_from_string
// Ported from: bid128_string.c lines 274-707
func Bid128FromString(str string, rnd_mode int) (res BID_UINT128, pfpsf uint32) {
	var CX BID_UINT128
	var sign_x, coeff_high, coeff_low, coeff2, coeff_l2, carry uint64
	var scale_high, right_radix_leading_zeros uint64
	var ndigits_before, ndigits_after, ndigits_total, dec_expon, sgn_exp int
	var i, d2, rdx_pt_enc int
	var set_inexact int
	var min_digits, sticky_bit int
	var buffer [MAX_STRING_DIGITS_128_str]byte
	var c byte

	// Convert string to byte slice with null terminator for C-like pointer semantics
	ps := []byte(str)
	ps = append(ps, 0)

	right_radix_leading_zeros = 0
	rdx_pt_enc = 0

	// eliminate leading white space
	for (ps[0] == ' ' || ps[0] == '\t') && ps[0] != 0 {
		ps = ps[1:]
	}

	// c gets first character
	c = ps[0]

	// if c is null or not equal to a (radix point, negative sign,
	// positive sign, or number) it might be SNaN, sNaN, Infinity
	if c == 0 ||
		(c != '.' && c != '-' && c != '+' &&
			(uint(c-'0') > 9)) {
		res.w[0] = 0
		// Infinity?
		if (tolower_macro(ps[0]) == 'i' && tolower_macro(ps[1]) == 'n' &&
			tolower_macro(ps[2]) == 'f') &&
			(ps[3] == 0 ||
				(tolower_macro(ps[3]) == 'i' &&
					tolower_macro(ps[4]) == 'n' &&
					tolower_macro(ps[5]) == 'i' &&
					tolower_macro(ps[6]) == 't' &&
					tolower_macro(ps[7]) == 'y' && ps[8] == 0)) {
			res.w[1] = 0x7800000000000000
			return
		}
		// return sNaN
		if tolower_macro(ps[0]) == 's' && tolower_macro(ps[1]) == 'n' &&
			tolower_macro(ps[2]) == 'a' && tolower_macro(ps[3]) == 'n' {
			res.w[1] = 0x7e00000000000000
			return
		}
		// return qNaN
		res.w[1] = 0x7c00000000000000
		return
	}

	// if +Inf, -Inf, +Infinity, or -Infinity (case insensitive check for inf)
	if (tolower_macro(ps[1]) == 'i' && tolower_macro(ps[2]) == 'n' &&
		tolower_macro(ps[3]) == 'f') && (ps[4] == 0 ||
		(tolower_macro(ps[4]) == 'i' && tolower_macro(ps[5]) == 'n' &&
			tolower_macro(ps[6]) == 'i' && tolower_macro(ps[7]) == 't' &&
			tolower_macro(ps[8]) == 'y' && ps[9] == 0)) {
		res.w[0] = 0

		if c == '+' {
			res.w[1] = 0x7800000000000000
		} else if c == '-' {
			res.w[1] = 0xf800000000000000
		} else {
			res.w[1] = 0x7c00000000000000
		}

		return
	}

	// if +sNaN, +SNaN, -sNaN, or -SNaN
	if tolower_macro(ps[1]) == 's' && tolower_macro(ps[2]) == 'n' &&
		tolower_macro(ps[3]) == 'a' && tolower_macro(ps[4]) == 'n' {
		res.w[0] = 0
		if c == '-' {
			res.w[1] = 0xfe00000000000000
		} else {
			res.w[1] = 0x7e00000000000000
		}
		return
	}

	// set up sign_x to be OR'ed with the upper word later
	if c == '-' {
		sign_x = 0x8000000000000000
	} else {
		sign_x = 0
	}

	// go to next character if leading sign
	if c == '-' || c == '+' {
		ps = ps[1:]
	}

	c = ps[0]

	// if c isn't a decimal point or a decimal digit, return NaN
	if c != '.' && (uint(c-'0') > 9) {
		res.w[1] = 0x7c00000000000000 | sign_x
		res.w[0] = 0
		return
	}
	if c == '.' {
		rdx_pt_enc = 1
		ps = ps[1:]
	}

	// detect zero (and eliminate/ignore leading zeros)
	if ps[0] == '0' {
		// if all numbers are zeros (with possibly 1 radix point, the number is zero
		for ps[0] == '0' {
			ps = ps[1:]

			// for numbers such as 0.0000000000000000000000000000000000001001,
			// we want to count the leading zeros
			if rdx_pt_enc != 0 {
				right_radix_leading_zeros++
			}
			// if this character is a radix point, make sure we haven't already
			// encountered one
			if ps[0] == '.' {
				if rdx_pt_enc == 0 {
					rdx_pt_enc = 1
					// if this is the first radix point, and the next character is NULL,
					// we have a zero
					if ps[1] == 0 {
						res.w[1] =
							(0x3040000000000000 -
								(right_radix_leading_zeros << 49)) | sign_x
						res.w[0] = 0
						return
					}
					ps = ps[1:]
				} else {
					// if 2 radix points, return NaN
					res.w[1] = 0x7c00000000000000 | sign_x
					res.w[0] = 0
					return
				}
			} else if ps[0] == 0 {
				if right_radix_leading_zeros > 6176 {
					right_radix_leading_zeros = 6176
				}
				res.w[1] =
					(0x3040000000000000 -
						(right_radix_leading_zeros << 49)) | sign_x
				res.w[0] = 0
				return
			}
		}
	}

	c = ps[0]

	// initialize local variables
	ndigits_before = 0
	ndigits_after = 0
	ndigits_total = 0
	sgn_exp = 0

	if rdx_pt_enc == 0 {
		// investigate string (before radix point)
		for uint(c-'0') <= 9 {
			if ndigits_before < MAX_FORMAT_DIGITS_128_str {
				buffer[ndigits_before] = c
			} else if ndigits_before < MAX_STRING_DIGITS_128_str {
				buffer[ndigits_before] = c
				if c > '0' {
					set_inexact = 1
				}
			} else if c > '0' {
				set_inexact = 1
				sticky_bit = 1
			}
			ps = ps[1:]
			c = ps[0]
			ndigits_before++
		}

		ndigits_total = ndigits_before
		if c == '.' {
			ps = ps[1:]
			c = ps[0]
			if c != 0 {
				// investigate string (after radix point)
				for uint(c-'0') <= 9 {
					if ndigits_total < MAX_FORMAT_DIGITS_128_str {
						buffer[ndigits_total] = c
					} else if ndigits_total < MAX_STRING_DIGITS_128_str {
						buffer[ndigits_total] = c
						if c > '0' {
							set_inexact = 1
						}
					} else if c > '0' {
						set_inexact = 1
						sticky_bit = 1
					}
					ps = ps[1:]
					c = ps[0]
					ndigits_total++
				}
				ndigits_after = ndigits_total - ndigits_before
			}
		}
	} else {
		// we encountered a radix point while detecting zeros
		c = ps[0]
		ndigits_total = 0
		// investigate string (after radix point)
		for uint(c-'0') <= 9 {
			if ndigits_total < MAX_FORMAT_DIGITS_128_str {
				buffer[ndigits_total] = c
			} else if ndigits_total < MAX_STRING_DIGITS_128_str {
				buffer[ndigits_total] = c
				if c > '0' {
					set_inexact = 1
				}
			} else if c > '0' {
				set_inexact = 1
				sticky_bit = 1
			}
			ps = ps[1:]
			c = ps[0]
			ndigits_total++
		}
		ndigits_after = ndigits_total - ndigits_before
	}

	// get exponent
	dec_expon = 0
	if c != 0 {
		if c != 'e' && c != 'E' {
			// return NaN
			res.w[1] = 0x7c00000000000000
			res.w[0] = 0
			return
		}
		ps = ps[1:]
		c = ps[0]

		if (uint(c-'0') > 9) &&
			((c != '+' && c != '-') || (uint(ps[1]-'0') > 9)) {
			// return NaN
			res.w[1] = 0x7c00000000000000
			res.w[0] = 0
			return
		}

		if c == '-' {
			sgn_exp = -1
			ps = ps[1:]
			c = ps[0]
		} else if c == '+' {
			ps = ps[1:]
			c = ps[0]
		}

		dec_expon = int(c - '0')
		i = 1
		ps = ps[1:]

		if dec_expon == 0 {
			for ps[0] == '0' {
				ps = ps[1:]
			}
		}
		c = ps[0] - '0'

		for uint(c) <= 9 && i < 7 {
			d2 = dec_expon + dec_expon
			dec_expon = (d2 << 2) + d2 + int(c)
			ps = ps[1:]
			c = ps[0] - '0'
			i++
		}
	}

	dec_expon = (dec_expon + sgn_exp) ^ sgn_exp

	if ndigits_total <= MAX_FORMAT_DIGITS_128_str {
		dec_expon +=
			EXPONENT_BIAS128 - ndigits_after -
				int(right_radix_leading_zeros)
		if dec_expon < 0 {
			res.w[1] = 0 | sign_x
			res.w[0] = 0
		}
		if ndigits_total == 0 {
			CX.w[0] = 0
			CX.w[1] = 0
		} else if ndigits_total <= 19 {
			coeff_high = uint64(buffer[0] - '0')
			for i = 1; i < ndigits_total; i++ {
				coeff2 = coeff_high + coeff_high
				coeff_high = (coeff2 << 2) + coeff2 + uint64(buffer[i]-'0')
			}
			CX.w[0] = coeff_high
			CX.w[1] = 0
		} else {
			coeff_high = uint64(buffer[0] - '0')
			for i = 1; i < ndigits_total-17; i++ {
				coeff2 = coeff_high + coeff_high
				coeff_high = (coeff2 << 2) + coeff2 + uint64(buffer[i]-'0')
			}
			coeff_low = uint64(buffer[i] - '0')
			i++
			for ; i < ndigits_total; i++ {
				coeff_l2 = coeff_low + coeff_low
				coeff_low = (coeff_l2 << 2) + coeff_l2 + uint64(buffer[i]-'0')
			}
			// now form the coefficient as coeff_high*10^17+coeff_low+carry
			scale_high = 100000000000000000
			CX = __mul_64x64_to_128_fast(coeff_high, scale_high)

			CX.w[0] += coeff_low
			if CX.w[0] < coeff_low {
				CX.w[1]++
			}
		}
		res = bid_get_BID128(sign_x, dec_expon, CX, rnd_mode, &pfpsf)
		return
	} else {
		// simply round using the digits that were read

		dec_expon +=
			ndigits_before + EXPONENT_BIAS128 -
				MAX_FORMAT_DIGITS_128_str - int(right_radix_leading_zeros)

		if dec_expon < 0 {
			res.w[1] = 0 | sign_x
			res.w[0] = 0
		}

		coeff_high = uint64(buffer[0] - '0')
		for i = 1; i < MAX_FORMAT_DIGITS_128_str-17; i++ {
			coeff2 = coeff_high + coeff_high
			coeff_high = (coeff2 << 2) + coeff2 + uint64(buffer[i]-'0')
		}
		coeff_low = uint64(buffer[i] - '0')
		i++
		for ; i < MAX_FORMAT_DIGITS_128_str; i++ {
			coeff_l2 = coeff_low + coeff_low
			coeff_low = (coeff_l2 << 2) + coeff_l2 + uint64(buffer[i]-'0')
		}
		switch rnd_mode {
		case BID_ROUNDING_TO_NEAREST:
			carry = uint64(uint32(int('4')-int(buffer[i]))) >> 31
			if (buffer[i] == '5' && (coeff_low&1) == 0 && sticky_bit == 0) || dec_expon < 0 {
				if dec_expon >= 0 {
					carry = 0
					i++
				}
				min_digits = ndigits_total
				if min_digits > MAX_STRING_DIGITS_128_str {
					min_digits = MAX_STRING_DIGITS_128_str
				}
				carry = uint64(sticky_bit)
				for ; carry == 0 && i < min_digits; i++ {
					if buffer[i] > '0' {
						carry = 1
						break
					}
				}
			}

		case BID_ROUNDING_DOWN:
			carry = 0
			if sign_x != 0 {
				min_digits = ndigits_total
				if min_digits > MAX_STRING_DIGITS_128_str {
					min_digits = MAX_STRING_DIGITS_128_str
				}
				carry = uint64(sticky_bit)
				for ; carry == 0 && i < min_digits; i++ {
					if buffer[i] > '0' {
						carry = 1
						break
					}
				}
			}
		case BID_ROUNDING_UP:
			carry = 0
			if sign_x == 0 {
				min_digits = ndigits_total
				if min_digits > MAX_STRING_DIGITS_128_str {
					min_digits = MAX_STRING_DIGITS_128_str
				}
				carry = uint64(sticky_bit)
				for ; carry == 0 && i < min_digits; i++ {
					if buffer[i] > '0' {
						carry = 1
						break
					}
				}
			}
		case BID_ROUNDING_TO_ZERO:
			carry = 0
		case BID_ROUNDING_TIES_AWAY:
			carry = uint64(uint32(int('4')-int(buffer[i]))) >> 31
			if dec_expon < 0 {
				min_digits = ndigits_total
				if min_digits > MAX_STRING_DIGITS_128_str {
					min_digits = MAX_STRING_DIGITS_128_str
				}
				carry = uint64(sticky_bit)
				for ; carry == 0 && i < min_digits; i++ {
					if buffer[i] > '0' {
						carry = 1
						break
					}
				}
			}
		default:
			carry = 0
		}
		// now form the coefficient as coeff_high*10^17+coeff_low+carry
		scale_high = 100000000000000000
		if dec_expon < 0 {
			if dec_expon > -MAX_FORMAT_DIGITS_128_str {
				scale_high = 1000000000000000000
				coeff_low = (coeff_low << 3) + (coeff_low << 1)
				dec_expon--
			}
			if dec_expon == -MAX_FORMAT_DIGITS_128_str &&
				coeff_high > 50000000000000000 {
				carry = 0
			}
		}

		CX = __mul_64x64_to_128_fast(coeff_high, scale_high)

		coeff_low += carry
		CX.w[0] += coeff_low
		if CX.w[0] < coeff_low {
			CX.w[1]++
		}

		if set_inexact != 0 {
			pfpsf |= BID_INEXACT_EXCEPTION
		}

		res = bid_get_BID128(sign_x, dec_expon, CX, rnd_mode, &pfpsf)
		return
	}
}
