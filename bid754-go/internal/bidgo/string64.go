package bidgo

// bid64_to_string - Intel bid64_string.c에서 기계적 포팅
// MiDi (Millennial Digits) 알고리즘 사용

import "math"

const (
	tostring_MAX_FORMAT_DIGITS     = 16
	tostring_DECIMAL_EXPONENT_BIAS = 398
)

// __L0_Normalize_10to18 - 매크로 포팅
func l0_Normalize_10to18(X_hi, X_lo *uint64) {
	L0_tmp := *X_lo + bid_Twoto60_m_10to18
	if L0_tmp&bid_Twoto60 != 0 {
		*X_hi = *X_hi + 1
		*X_lo = (L0_tmp << 4) >> 4
	}
}

// __L0_Split_MiDi_2 - 매크로 포팅
// C의 *((ptr)++) 연속 기록 2회를, 커서 값 1회 로드 후 상수 오프셋 기록과
// 커서 1회 전진(순 효과 +2 동일)으로 국소 표현만 변경 (기록 값/순서 동일).
func l0_Split_MiDi_2(X uint32, MiDi []uint32, ptr *int) {
	L0_head := X >> 10
	L0_tail := (X & 0x03FF) + (L0_head << 5) - (L0_head << 3)
	L0_tmp := L0_tail >> 10
	L0_head += L0_tmp
	L0_tail = (L0_tail & 0x03FF) + (L0_tmp << 5) - (L0_tmp << 3)
	if L0_tail > 999 {
		L0_tail -= 1000
		L0_head += 1
	}
	p := *ptr
	MiDi[p] = L0_head
	MiDi[p+1] = L0_tail
	*ptr = p + 2
}

// __L0_Split_MiDi_3 - 매크로 포팅
// C의 *((ptr)++) 연속 기록 3회를, 커서 값 1회 로드 후 상수 오프셋 기록과
// 커서 1회 전진(순 효과 +3 동일)으로 국소 표현만 변경 (기록 값/순서 동일).
func l0_Split_MiDi_3(X uint64, MiDi []uint32, ptr *int) {
	L0_X := uint32(X)
	L0_head := ((L0_X >> 17) * 34359) >> 18
	L0_X -= L0_head * 1000000
	if L0_X >= 1000000 {
		L0_X -= 1000000
		L0_head += 1
	}
	L0_mid := L0_X >> 10
	L0_tail := (L0_X & 0x03FF) + (L0_mid << 5) - (L0_mid << 3)
	L0_tmp := L0_tail >> 10
	L0_mid += L0_tmp
	L0_tail = (L0_tail & 0x3FF) + (L0_tmp << 5) - (L0_tmp << 3)
	if L0_tail > 999 {
		L0_tail -= 1000
		L0_mid += 1
	}
	p := *ptr
	MiDi[p] = L0_head
	MiDi[p+1] = L0_mid
	MiDi[p+2] = L0_tail
	*ptr = p + 3
}

// __L1_Split_MiDi_6_Lead - 매크로 포팅
func l1_Split_MiDi_6_Lead(X uint64, MiDi []uint32, ptr *int) {
	if X >= uint64(bid_Tento9) {
		L1_Xhi_64 := ((X >> 28) * bid_Inv_Tento9) >> 33
		L1_Xlo_64 := X - L1_Xhi_64*uint64(bid_Tento9)
		if L1_Xlo_64 >= uint64(bid_Tento9) {
			L1_Xlo_64 -= uint64(bid_Tento9)
			L1_Xhi_64 += 1
		}
		L1_X_hi := uint32(L1_Xhi_64)
		L1_X_lo := uint32(L1_Xlo_64)
		if L1_X_hi >= bid_Tento6 {
			l0_Split_MiDi_3(uint64(L1_X_hi), MiDi, ptr)
			l0_Split_MiDi_3(uint64(L1_X_lo), MiDi, ptr)
		} else if L1_X_hi >= bid_Tento3 {
			l0_Split_MiDi_2(L1_X_hi, MiDi, ptr)
			l0_Split_MiDi_3(uint64(L1_X_lo), MiDi, ptr)
		} else {
			MiDi[*ptr] = L1_X_hi
			*ptr++
			l0_Split_MiDi_3(uint64(L1_X_lo), MiDi, ptr)
		}
	} else {
		L1_X_lo := uint32(X)
		if L1_X_lo >= bid_Tento6 {
			l0_Split_MiDi_3(uint64(L1_X_lo), MiDi, ptr)
		} else if L1_X_lo >= bid_Tento3 {
			l0_Split_MiDi_2(L1_X_lo, MiDi, ptr)
		} else {
			MiDi[*ptr] = L1_X_lo
			*ptr++
		}
	}
}

// __L0_MiDi2Str - 매크로 포팅 (3자리 출력)
// C의 *((c_ptr)++) 연속 기록 3회를, 커서 값 1회 로드 후 상수 오프셋 기록과
// 커서 1회 전진(순 효과 +3 동일)으로 국소 표현만 변경 (기록 값/순서 동일).
func l0_MiDi2Str(X uint32, ps []byte, c_ptr *int) {
	src := bid_midi_tbl[X]
	p := *c_ptr
	ps[p] = src[0]
	ps[p+1] = src[1]
	ps[p+2] = src[2]
	*c_ptr = p + 3
}

// __L0_MiDi2Str_Lead - 매크로 포팅 (선행 0 제거)
// 각 분기의 *((c_ptr)++) 연속 기록을, 커서 값 1회 로드 후 상수 오프셋 기록과
// 분기별 기록 폭만큼의 커서 1회 전진으로 국소 표현만 변경 (기록 값/순서 동일).
func l0_MiDi2Str_Lead(X uint32, ps []byte, c_ptr *int) {
	src := bid_midi_tbl[X]
	p := *c_ptr
	if X >= 100 {
		ps[p] = src[0]
		ps[p+1] = src[1]
		ps[p+2] = src[2]
		*c_ptr = p + 3
	} else if X >= 10 {
		ps[p] = src[1]
		ps[p+1] = src[2]
		*c_ptr = p + 2
	} else {
		ps[p] = src[2]
		*c_ptr = p + 1
	}
}

// Bid64ToString - bid64_to_string의 Go 구현
// Intel bid64_string.c lines 41-242 기계적 포팅
func Bid64ToString(x uint64) string {
	var ps [64]byte // 충분한 버퍼
	var istart int

	// unpack arguments, check for NaN or Infinity
	sign_x, exponent_x, coefficient_x, valid := unpack_BID64(x)
	if !valid {
		// x is Inf. or NaN or 0

		// Inf or NaN?
		if (x & 0x7800000000000000) == 0x7800000000000000 {
			if (x & 0x7c00000000000000) == 0x7c00000000000000 {
				if sign_x != 0 {
					ps[0] = '-'
				} else {
					ps[0] = '+'
				}
				ps[1] = 'S'
				j := 2
				if (x & SNAN_MASK64) != SNAN_MASK64 {
					j = 1
				}
				ps[j] = 'N'
				j++
				ps[j] = 'a'
				j++
				ps[j] = 'N'
				j++
				return string(ps[:j])
			}
			// x is Inf
			if sign_x != 0 {
				ps[0] = '-'
			} else {
				ps[0] = '+'
			}
			ps[1] = 'I'
			ps[2] = 'n'
			ps[3] = 'f'
			return string(ps[:4])
		}
		// 0
		istart = 1
		if sign_x != 0 {
			ps[0] = '-'
		} else {
			ps[0] = '+'
		}

		ps[istart] = '0'
		istart++
		ps[istart] = 'E'
		istart++

		exponent_x -= 398
		if exponent_x < 0 {
			ps[istart] = '-'
			istart++
			exponent_x = -exponent_x
		} else {
			ps[istart] = '+'
			istart++
		}

		if exponent_x != 0 {
			// get decimal digits in exponent_x
			tempx := float32(exponent_x)
			bin_expon_cx := int((math.Float32bits(tempx)>>23)&0xff) - 0x7f
			digits_x := bid_estimate_decimal_digits[bin_expon_cx]
			if uint64(exponent_x) >= bid_power10_table_128[digits_x].lo {
				digits_x++
			}

			j := istart + digits_x - 1
			istart = j + 1

			// 2^32/10
			ER10 := uint64(0x1999999a)

			exp := exponent_x
			for exp > 9 {
				D := uint64(exp) * ER10
				D >>= 32
				exp = exp - int(D<<1) - int(D<<3)
				ps[j] = '0' + byte(exp)
				j--
				exp = int(D)
			}
			ps[j] = '0' + byte(exp)
		} else {
			ps[istart] = '0'
			istart++
		}

		return string(ps[:istart])
	}

	// convert expon, coeff to ASCII
	exponent_x -= tostring_DECIMAL_EXPONENT_BIAS

	istart = 1
	if sign_x != 0 {
		ps[0] = '-'
	} else {
		ps[0] = '+'
	}

	// if zero or non-canonical, set coefficient to '0'
	if coefficient_x > 9999999999999999 || coefficient_x == 0 {
		ps[istart] = '0'
		istart++
	} else {
		// MiDi algorithm
		var MiDi [12]uint32
		ptr := 0

		Tmp := coefficient_x >> 59
		LO_18Dig := (coefficient_x << 5) >> 5
		HI_18Dig := uint64(0)
		k_lcv := 0

		for Tmp != 0 {
			midi_ind := int(Tmp & 0x000000000000003F)
			midi_ind <<= 1
			Tmp >>= 6
			HI_18Dig += mod10_18_tbl[k_lcv][midi_ind]
			midi_ind++
			LO_18Dig += mod10_18_tbl[k_lcv][midi_ind]
			k_lcv++
			l0_Normalize_10to18(&HI_18Dig, &LO_18Dig)
		}

		l1_Split_MiDi_6_Lead(LO_18Dig, MiDi[:], &ptr)
		length := ptr

		c_ptr := istart

		// now convert the MiDi into character strings
		l0_MiDi2Str_Lead(MiDi[0], ps[:], &c_ptr)
		for k := 1; k < length; k++ {
			l0_MiDi2Str(MiDi[k], ps[:], &c_ptr)
		}
		istart = c_ptr
	}

	ps[istart] = 'E'
	istart++

	if exponent_x < 0 {
		ps[istart] = '-'
		istart++
		exponent_x = -exponent_x
	} else {
		ps[istart] = '+'
		istart++
	}

	if exponent_x != 0 {
		// get decimal digits in exponent_x
		tempx := float32(exponent_x)
		bin_expon_cx := int((math.Float32bits(tempx)>>23)&0xff) - 0x7f
		digits_x := bid_estimate_decimal_digits[bin_expon_cx]
		if uint64(exponent_x) >= bid_power10_table_128[digits_x].lo {
			digits_x++
		}

		j := istart + digits_x - 1
		istart = j + 1

		// 2^32/10
		ER10 := uint64(0x1999999a)

		exp := exponent_x
		for exp > 9 {
			D := uint64(exp) * ER10
			D >>= 32
			exp = exp - int(D<<1) - int(D<<3)
			ps[j] = '0' + byte(exp)
			j--
			exp = int(D)
		}
		ps[j] = '0' + byte(exp)
	} else {
		ps[istart] = '0'
		istart++
	}

	return string(ps[:istart])
}
