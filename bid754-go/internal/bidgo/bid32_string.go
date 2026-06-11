// Ported from: Intel bid32_string.c
// Mechanical translation of the Intel BID library to Go, with one documented
// IEEE-conformance deviation: the directed-rounding overflow on the no-exponent
// from_string path follows IEEE 754 (largest finite) instead of pinned Intel C
// (which ignores the rounding mode and returns Inf). See IEEE754_SPEC.md
// "pinned Intel BID C 대비 의도적 IEEE 편차". All other logic is preserved exactly.

package bidgo

import (
	"strconv"
	"strings"
)

// Bid32ToStringRaw is ported mechanically from bid32_string.c: bid32_to_string.
func Bid32ToStringRaw(x uint32) string {
	var CT uint64
	var d, istart, istart0 int
	var sign_x, coefficient_x uint32
	var exponent_x int
	var ps [64]byte

	sign_x, exponent_x, coefficient_x, valid := unpack_BID32(x)
	if !valid {
		if sign_x != 0 {
			ps[0] = '-'
		} else {
			ps[0] = '+'
		}
		if (x & NAN_MASK32) == NAN_MASK32 {
			if (x & SNAN_MASK32) == SNAN_MASK32 {
				return string(ps[:1]) + "SNaN"
			}
			return string(ps[:1]) + "NaN"
		}
		if (x & INFINITY_MASK32) == INFINITY_MASK32 {
			return string(ps[:1]) + "Inf"
		}
		istart = 1
		ps[istart] = '0'
		istart++
	} else {
		if sign_x != 0 {
			ps[0] = '-'
		} else {
			ps[0] = '+'
		}
		istart = 1
		if coefficient_x >= 1000000 {
			CT = uint64(coefficient_x) * 0x431BDE83
			CT >>= 32
			d = int(CT >> (50 - 32))
			ps[istart] = byte(d) + '0'
			istart++

			coefficient_x -= uint32(d) * 1000000

			CT = uint64(coefficient_x) * 0x20C49BA6
			CT >>= 32
			d = int(CT >> (39 - 32))
			ps[istart] = bid_midi_tbl[d][0]
			istart++
			ps[istart] = bid_midi_tbl[d][1]
			istart++
			ps[istart] = bid_midi_tbl[d][2]
			istart++

			d = int(coefficient_x) - d*1000

			ps[istart] = bid_midi_tbl[d][0]
			istart++
			ps[istart] = bid_midi_tbl[d][1]
			istart++
			ps[istart] = bid_midi_tbl[d][2]
			istart++
		} else if coefficient_x >= 1000 {
			CT = uint64(coefficient_x) * 0x20C49BA6
			CT >>= 32
			d = int(CT >> (39 - 32))

			istart0 = istart
			ps[istart] = bid_midi_tbl[d][0]
			if ps[istart] != '0' {
				istart++
			}
			ps[istart] = bid_midi_tbl[d][1]
			if ps[istart] != '0' || istart != istart0 {
				istart++
			}
			ps[istart] = bid_midi_tbl[d][2]
			istart++

			d = int(coefficient_x) - d*1000

			ps[istart] = bid_midi_tbl[d][0]
			istart++
			ps[istart] = bid_midi_tbl[d][1]
			istart++
			ps[istart] = bid_midi_tbl[d][2]
			istart++
		} else {
			d = int(coefficient_x)

			istart0 = istart
			ps[istart] = bid_midi_tbl[d][0]
			if ps[istart] != '0' {
				istart++
			}
			ps[istart] = bid_midi_tbl[d][1]
			if ps[istart] != '0' || istart != istart0 {
				istart++
			}
			ps[istart] = bid_midi_tbl[d][2]
			istart++
		}
	}

	if !valid {
		ps[istart] = 'E'
		istart++

		exponent_x -= DECIMAL_EXPONENT_BIAS_32
		if exponent_x < 0 {
			ps[istart] = '-'
			istart++
			exponent_x = -exponent_x
		} else {
			ps[istart] = '+'
			istart++
		}

		istart0 = istart
		ps[istart] = bid_midi_tbl[exponent_x][0]
		if ps[istart] != '0' {
			istart++
		}
		ps[istart] = bid_midi_tbl[exponent_x][1]
		if ps[istart] != '0' || istart != istart0 {
			istart++
		}
		ps[istart] = bid_midi_tbl[exponent_x][2]
		istart++

		return string(ps[:istart])
	}

	digits := string(ps[1:istart])
	adjustedExp := exponent_x - DECIMAL_EXPONENT_BIAS_32 + len(digits) - 1
	out := string(ps[:1]) + digits[:1]
	if len(digits) > 1 {
		out += "." + digits[1:]
	}
	if adjustedExp != 0 {
		out += "e" + strconv.Itoa(adjustedExp)
	}
	return out
}

// Bid32FromStringRaw is ported mechanically from bid32_string.c: bid32_from_string.
func Bid32FromStringRaw(ps string, rnd_mode int) (uint32, uint32) {
	var sign_x, coefficient_x, rounded uint64
	var expon_x, sgn_expon, ndigits, add_expon int
	var midpoint, rounded_up, dround int
	var dec_expon_scale, right_radix_leading_zeros, rdx_pt_enc int
	var pfpsf uint32
	var res uint64

	s := strings.TrimLeft(ps, " \t")
	if len(s) == 0 {
		return 0x7c000000, 0
	}

	c := s[0]
	idx := 0

	// detect special cases
	sl := strings.ToLower(s)
	if c != '.' && c != '-' && c != '+' && (c < '0' || c > '9') {
		if sl == "inf" || sl == "infinity" {
			return 0x78000000, 0
		}
		if strings.HasPrefix(sl, "snan") {
			return 0x7e000000, 0
		}
		return 0x7c000000, 0
	}

	// detect +/-INF, +/-sNaN
	if len(s) > 1 {
		sl1 := strings.ToLower(s[1:])
		if sl1 == "inf" || sl1 == "infinity" {
			if c == '+' {
				return 0x78000000, 0
			} else if c == '-' {
				return 0xf8000000, 0
			}
			return 0x7c000000, 0
		}
		if strings.HasPrefix(sl1, "snan") {
			if c == '-' {
				return 0xfe000000, 0
			}
			return 0x7e000000, 0
		}
		// +NaN or -NaN
		if sl1 == "nan" {
			if c == '-' {
				return 0xfc000000, 0
			}
			return 0x7c000000, 0
		}
	}

	if c == '-' {
		sign_x = 0x80000000
	} else {
		sign_x = 0
	}

	if c == '-' || c == '+' {
		idx++
		if idx >= len(s) {
			return uint32(0x7c000000 | sign_x), 0
		}
		c = s[idx]
	}

	if c != '.' && (c < '0' || c > '9') {
		return uint32(0x7c000000 | sign_x), 0
	}

	rdx_pt_enc = 0

	// detect zero and eliminate leading zeros
	if idx < len(s) && (s[idx] == '0' || s[idx] == '.') {
		if s[idx] == '.' {
			rdx_pt_enc = 1
			idx++
		}
		for idx < len(s) && s[idx] == '0' {
			idx++
			if rdx_pt_enc != 0 {
				right_radix_leading_zeros++
			}
			if idx < len(s) && s[idx] == '.' {
				if rdx_pt_enc == 0 {
					rdx_pt_enc = 1
					if idx+1 >= len(s) {
						right_radix_leading_zeros = DECIMAL_EXPONENT_BIAS_32 - right_radix_leading_zeros
						if right_radix_leading_zeros < 0 {
							right_radix_leading_zeros = 0
						}
						res = (uint64(right_radix_leading_zeros) << 23) | sign_x
						return uint32(res), 0
					}
					idx++
				} else {
					return uint32(0x7c000000 | sign_x), 0
				}
			} else if idx >= len(s) {
				right_radix_leading_zeros = DECIMAL_EXPONENT_BIAS_32 - right_radix_leading_zeros
				if right_radix_leading_zeros < 0 {
					right_radix_leading_zeros = 0
				}
				res = (uint64(right_radix_leading_zeros) << 23) | sign_x
				return uint32(res), 0
			}
		}
	}

	if idx >= len(s) {
		right_radix_leading_zeros = DECIMAL_EXPONENT_BIAS_32 - right_radix_leading_zeros
		if right_radix_leading_zeros < 0 {
			right_radix_leading_zeros = 0
		}
		res = (uint64(right_radix_leading_zeros) << 23) | sign_x
		return uint32(res), 0
	}

	c = s[idx]
	ndigits = 0
	for idx < len(s) && ((c >= '0' && c <= '9') || c == '.') {
		if c == '.' {
			if rdx_pt_enc != 0 {
				return uint32(0x7c000000 | sign_x), 0
			}
			rdx_pt_enc = 1
			idx++
			if idx < len(s) {
				c = s[idx]
			}
			continue
		}
		dec_expon_scale += rdx_pt_enc

		ndigits++
		if ndigits <= 7 {
			coefficient_x = (coefficient_x << 1) + (coefficient_x << 3)
			coefficient_x += uint64(c - '0')
		} else if ndigits == 8 {
			switch rnd_mode {
			case BID_ROUNDING_TO_NEAREST:
				if c == '5' && (coefficient_x&1) == 0 {
					midpoint = 1
				}
				if c > '5' || (c == '5' && (coefficient_x&1) != 0) {
					coefficient_x++
					rounded_up = 1
				}
			case BID_ROUNDING_DOWN:
				if sign_x != 0 {
					if c > '0' {
						coefficient_x++
						rounded_up = 1
					} else {
						dround = 1
					}
				}
			case BID_ROUNDING_UP:
				if sign_x == 0 {
					if c > '0' {
						coefficient_x++
						rounded_up = 1
					} else {
						dround = 1
					}
				}
			case BID_ROUNDING_TIES_AWAY:
				if c >= '5' {
					coefficient_x++
					rounded_up = 1
				}
			}
			if coefficient_x == 10000000 {
				coefficient_x = 1000000
				add_expon = 1
			}
			if c > '0' {
				rounded = 1
			}
			add_expon += 1
		} else {
			add_expon++
			if midpoint != 0 && c > '0' {
				coefficient_x++
				midpoint = 0
				rounded_up = 1
			}
			if c > '0' {
				rounded = 1
				if dround != 0 {
					dround = 0
					coefficient_x++
					rounded_up = 1
					if coefficient_x == 10000000 {
						coefficient_x = 1000000
						add_expon++
					}
				}
			}
		}
		idx++
		if idx < len(s) {
			c = s[idx]
		} else {
			c = 0
		}
	}

	add_expon -= (dec_expon_scale + right_radix_leading_zeros)

	if idx >= len(s) {
		if rounded != 0 {
			pfpsf |= BID_INEXACT_EXCEPTION
		}
		res = uint64(get_BID32_flags(uint32(sign_x), add_expon+DECIMAL_EXPONENT_BIAS_32, coefficient_x, rnd_mode, &pfpsf))
		return uint32(res), pfpsf
	}

	c = s[idx]
	if c != 'E' && c != 'e' {
		return uint32(0x7c000000 | sign_x), 0
	}
	idx++
	if idx >= len(s) {
		return uint32(0x7c000000 | sign_x), 0
	}
	c = s[idx]
	if c == '-' {
		sgn_expon = 1
	}
	if c == '-' || c == '+' {
		idx++
		if idx >= len(s) {
			return uint32(0x7c000000 | sign_x), 0
		}
		c = s[idx]
	}
	if c < '0' || c > '9' {
		return uint32(0x7c000000 | sign_x), 0
	}

	for idx < len(s) && s[idx] >= '0' && s[idx] <= '9' {
		if expon_x < (1 << 20) {
			expon_x = (expon_x << 1) + (expon_x << 3)
			expon_x += int(s[idx] - '0')
		}
		idx++
	}

	if idx < len(s) {
		return uint32(0x7c000000 | sign_x), 0
	}

	if rounded != 0 {
		pfpsf |= BID_INEXACT_EXCEPTION
	}

	if sgn_expon != 0 {
		expon_x = -expon_x
	}

	expon_x += add_expon + DECIMAL_EXPONENT_BIAS_32

	if expon_x < 0 {
		if rounded_up != 0 {
			coefficient_x--
		}
		rnd_mode = 0
		res = uint64(get_BID32_UF(uint32(sign_x), expon_x, coefficient_x, uint32(rounded), rnd_mode, &pfpsf))
		return uint32(res), pfpsf
	}
	res = uint64(get_BID32_flags(uint32(sign_x), expon_x, coefficient_x, rnd_mode, &pfpsf))
	return uint32(res), pfpsf
}
