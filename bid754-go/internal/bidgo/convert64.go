package bidgo

// 정수-Decimal64 변환 함수
// Intel bid_from_int.c에서 기계적 포팅

const (
	SIGNMASK32    = 0x80000000
	SIGNMASK64    = 0x8000000000000000
	BID64_SIG_MAX = 9999999999999999 // 10^16 - 1
)

// Bid64FromInt32 - int32를 Decimal64로 변환
// Intel bid64_from_int32 기계적 포팅
func Bid64FromInt32(x int32) uint64 {
	var res uint64

	// if integer is negative, put the absolute value
	// in the lowest 32bits of the result
	if (uint32(x) & SIGNMASK32) == SIGNMASK32 {
		// negative int32
		x = ^x + 1 // 2's complement of x
		res = uint64(uint32(x)) | 0xb1c0000000000000
		// (exp << 53)) = biased exp. is 0
	} else { // positive int32
		res = uint64(x) | 0x31c0000000000000 // (exp << 53)) = biased exp. is 0
	}
	return res
}

// Bid64FromUint32 - uint32를 Decimal64로 변환
// Intel bid64_from_uint32 기계적 포팅
func Bid64FromUint32(x uint32) uint64 {
	res := uint64(x) | 0x31c0000000000000 // (exp << 53)) = biased exp. is 0
	return res
}

// Bid64FromInt64 - int64를 Decimal64로 변환
// Intel bid64_from_int64 기계적 포팅
func Bid64FromInt64(x int64, rndMode int) (uint64, uint32) {
	var res uint64
	var pfpsf uint32
	var C uint64
	var incr_exp int
	var is_midpoint_lt_even, is_midpoint_gt_even int
	var is_inexact_lt_midpoint, is_inexact_gt_midpoint int

	x_sign := uint64(x) & 0x8000000000000000
	// if the integer is negative, use the absolute value
	if x_sign != 0 {
		C = ^uint64(x) + 1
	} else {
		C = uint64(x)
	}

	if C <= BID64_SIG_MAX { // |C| <= 10^16-1 and the result is exact
		if C < 0x0020000000000000 { // C < 2^53
			res = x_sign | 0x31c0000000000000 | C
		} else { // C >= 2^53
			res = x_sign | 0x6c70000000000000 | (C & 0x0007ffffffffffff)
		}
	} else { // |C| >= 10^16 and the result may be inexact
		// the smallest |C| is 10^16 which has 17 decimal digits
		// the largest |C| is 0x8000000000000000 = 9223372036854775808 w/ 19 digits
		var q, ind uint32
		if C < 0x16345785d8a0000 { // x < 10^17
			q = 17
			ind = 1 // number of digits to remove for q = 17
		} else if C < 0xde0b6b3a7640000 { // C < 10^18
			q = 18
			ind = 2 // number of digits to remove for q = 18
		} else { // C < 10^19
			q = 19
			ind = 3 // number of digits to remove for q = 19
		}
		// overflow and underflow are not possible
		res = bid_round64_2_18(int(q), int(ind), C, &incr_exp,
			&is_midpoint_lt_even, &is_midpoint_gt_even,
			&is_inexact_lt_midpoint, &is_inexact_gt_midpoint)
		if incr_exp != 0 {
			ind++
		}
		// set the inexact flag
		if is_inexact_lt_midpoint != 0 || is_inexact_gt_midpoint != 0 ||
			is_midpoint_lt_even != 0 || is_midpoint_gt_even != 0 {
			pfpsf |= BID_INEXACT_EXCEPTION
		}
		// general correction from RN to RA, RM, RP, RZ; result uses ind for exp
		if rndMode != BID_ROUNDING_TO_NEAREST {
			if (x_sign == 0 &&
				((rndMode == BID_ROUNDING_UP && is_inexact_lt_midpoint != 0) ||
					((rndMode == BID_ROUNDING_TIES_AWAY || rndMode == BID_ROUNDING_UP) && is_midpoint_gt_even != 0))) ||
				(x_sign != 0 &&
					((rndMode == BID_ROUNDING_DOWN && is_inexact_lt_midpoint != 0) ||
						((rndMode == BID_ROUNDING_TIES_AWAY || rndMode == BID_ROUNDING_DOWN) && is_midpoint_gt_even != 0))) {
				res = res + 1
				if res == 0x002386f26fc10000 { // res = 10^16 => rounding overflow
					res = 0x00038d7ea4c68000 // 10^15
					ind = ind + 1
				}
			} else if (is_midpoint_lt_even != 0 || is_inexact_gt_midpoint != 0) &&
				((x_sign != 0 && (rndMode == BID_ROUNDING_UP || rndMode == BID_ROUNDING_TO_ZERO)) ||
					(x_sign == 0 && (rndMode == BID_ROUNDING_DOWN || rndMode == BID_ROUNDING_TO_ZERO))) {
				res = res - 1
				// check if we crossed into the lower decade
				if res == 0x00038d7ea4c67fff { // 10^15 - 1
					res = 0x002386f26fc0ffff // 10^16 - 1
					ind = ind - 1
				}
			}
			// else: exact, the result is already correct
		}
		if res < 0x0020000000000000 { // res < 2^53
			res = x_sign | (uint64(ind+398) << 53) | res
		} else { // res >= 2^53
			res = x_sign | 0x6000000000000000 | (uint64(ind+398) << 51) | (res & 0x0007ffffffffffff)
		}
	}
	return res, pfpsf
}

// Bid64FromUint64 - uint64를 Decimal64로 변환
// Intel bid64_from_uint64 기계적 포팅
func Bid64FromUint64(x uint64, rndMode int) (uint64, uint32) {
	var res uint64
	var pfpsf uint32
	var incr_exp int
	var is_midpoint_lt_even, is_midpoint_gt_even int
	var is_inexact_lt_midpoint, is_inexact_gt_midpoint int

	if x <= BID64_SIG_MAX { // x <= 10^16-1 and the result is exact
		if x < 0x0020000000000000 { // x < 2^53
			res = 0x31c0000000000000 | x
		} else { // x >= 2^53
			res = 0x6c70000000000000 | (x & 0x0007ffffffffffff)
		}
	} else { // x >= 10^16 and the result may be inexact
		// the smallest x is 10^16 which has 17 decimal digits
		// the largest x is 0xffffffffffffffff = 18446744073709551615 w/ 20 digits
		var q, ind uint32
		if x < 0x16345785d8a0000 { // x < 10^17
			q = 17
			ind = 1 // number of digits to remove for q = 17
		} else if x < 0xde0b6b3a7640000 { // x < 10^18
			q = 18
			ind = 2 // number of digits to remove for q = 18
		} else if x < 0x8ac7230489e80000 { // x < 10^19
			q = 19
			ind = 3 // number of digits to remove for q = 19
		} else { // x < 10^20
			q = 20
			ind = 4 // number of digits to remove for q = 20
		}
		// overflow and underflow are not possible
		if q <= 19 {
			res = bid_round64_2_18(int(q), int(ind), x, &incr_exp,
				&is_midpoint_lt_even, &is_midpoint_gt_even,
				&is_inexact_lt_midpoint, &is_inexact_gt_midpoint)
		} else { // q = 20
			res = bid_round128_19_38_for64(int(q), int(ind), x, &incr_exp,
				&is_midpoint_lt_even, &is_midpoint_gt_even,
				&is_inexact_lt_midpoint, &is_inexact_gt_midpoint)
		}
		if incr_exp != 0 {
			ind++
		}
		// set the inexact flag
		if is_inexact_lt_midpoint != 0 || is_inexact_gt_midpoint != 0 ||
			is_midpoint_lt_even != 0 || is_midpoint_gt_even != 0 {
			pfpsf |= BID_INEXACT_EXCEPTION
		}
		// general correction from RN to RA, RM, RP, RZ; result uses ind for exp
		if rndMode != BID_ROUNDING_TO_NEAREST {
			if (rndMode == BID_ROUNDING_UP && is_inexact_lt_midpoint != 0) ||
				((rndMode == BID_ROUNDING_TIES_AWAY || rndMode == BID_ROUNDING_UP) && is_midpoint_gt_even != 0) {
				res = res + 1
				if res == 0x002386f26fc10000 { // res = 10^16 => rounding overflow
					res = 0x00038d7ea4c68000 // 10^15
					ind = ind + 1
				}
			} else if (is_midpoint_lt_even != 0 || is_inexact_gt_midpoint != 0) &&
				(rndMode == BID_ROUNDING_DOWN || rndMode == BID_ROUNDING_TO_ZERO) {
				res = res - 1
				// check if we crossed into the lower decade
				if res == 0x00038d7ea4c67fff { // 10^15 - 1
					res = 0x002386f26fc0ffff // 10^16 - 1
					ind = ind - 1
				}
			}
			// else: exact, the result is already correct
		}
		if res < 0x0020000000000000 { // res < 2^53
			res = (uint64(ind+398) << 53) | res
		} else { // res >= 2^53
			res = 0x6000000000000000 | (uint64(ind+398) << 51) | (res & 0x0007ffffffffffff)
		}
	}
	return res, pfpsf
}

// bid_round64_2_18 - 라운딩 함수 (2-18 자릿수)
// Intel bid_round.c lines 116-215 기계적 포팅
// round a number C with q decimal digits, 2 <= q <= 18
// to q - x digits, 1 <= x <= 17
func bid_round64_2_18(q, x int, C uint64, incr_exp *int,
	is_midpoint_lt_even, is_midpoint_gt_even,
	is_inexact_lt_midpoint, is_inexact_gt_midpoint *int) uint64 {

	var P128 BID_UINT128
	var fstar BID_UINT128
	var Cstar uint64
	var tmp64 uint64
	var shift int
	var ind int

	// Note:
	//    In round128_2_18() positive numbers with 2 <= q <= 18 will be
	//    rounded to nearest only for 1 <= x <= 3:
	//     x = 1 or x = 2 when q = 17
	//     x = 2 or x = 3 when q = 18
	// However, for generality and possible uses outside the frame of IEEE 754
	// this implementation works for 1 <= x <= q - 1

	// assume *ptr_is_midpoint_lt_even, *ptr_is_midpoint_gt_even,
	// *ptr_is_inexact_lt_midpoint, and *ptr_is_inexact_gt_midpoint are
	// initialized to 0 by the caller

	// round a number C with q decimal digits, 2 <= q <= 18
	// to q - x digits, 1 <= x <= 17
	// C = C + 1/2 * 10^x where the result C fits in 64 bits
	// (because the largest value is 999999999999999999 + 50000000000000000 =
	// 0x0e92596fd628ffff, which fits in 60 bits)
	ind = x - 1 // 0 <= ind <= 16
	C = C + bid_midpoint64[ind]
	// kx ~= 10^(-x), kx = bid_Kx64[ind] * 2^(-Ex), 0 <= ind <= 16
	// P128 = (C + 1/2 * 10^x) * kx * 2^Ex = (C + 1/2 * 10^x) * Kx
	// the approximation kx of 10^(-x) was rounded up to 64 bits
	P128 = __mul_64x64_to_128(C, bid_Kx64[ind])
	// calculate C* = floor (P128) and f*
	// Cstar = P128 >> Ex
	// fstar = low Ex bits of P128
	shift = int(bid_Ex64m64[ind]) // in [3, 56]
	Cstar = P128.hi >> shift
	fstar.hi = P128.hi & bid_mask64[ind]
	fstar.lo = P128.lo
	// the top Ex bits of 10^(-x) are T* = bid_ten2mxtrunc64[ind], e.g.
	// if x=1, T*=bid_ten2mxtrunc64[0]=0xcccccccccccccccc
	// if (0 < f* < 10^(-x)) then the result is a midpoint
	//   if floor(C*) is even then C* = floor(C*) - logical right
	//       shift; C* has q - x decimal digits, correct by Prop. 1)
	//   else if floor(C*) is odd C* = floor(C*)-1 (logical right
	//       shift; C* has q - x decimal digits, correct by Pr. 1)
	// else
	//   C* = floor(C*) (logical right shift; C has q - x decimal digits,
	//       correct by Property 1)
	// in the caling function n = C* * 10^(e+x)

	// determine inexactness of the rounding of C*
	// if (0 < f* - 1/2 < 10^(-x)) then
	//   the result is exact
	// else // if (f* - 1/2 > T*) then
	//   the result is inexact
	if fstar.hi > bid_half64[ind] ||
		(fstar.hi == bid_half64[ind] && fstar.lo != 0) {
		// f* > 1/2 and the result may be exact
		// Calculate f* - 1/2
		tmp64 = fstar.hi - bid_half64[ind]
		if tmp64 != 0 || fstar.lo > bid_ten2mxtrunc64[ind] { // f* - 1/2 > 10^(-x)
			*is_inexact_lt_midpoint = 1
		} // else the result is exact
	} else { // the result is inexact; f2* <= 1/2
		*is_inexact_gt_midpoint = 1
	}
	// check for midpoints (could do this before determining inexactness)
	if fstar.hi == 0 && fstar.lo <= bid_ten2mxtrunc64[ind] {
		// the result is a midpoint
		if Cstar&0x01 != 0 { // Cstar is odd; MP in [EVEN, ODD]
			// if floor(C*) is odd C = floor(C*) - 1; the result may be 0
			Cstar-- // Cstar is now even
			*is_midpoint_gt_even = 1
			*is_inexact_lt_midpoint = 0
			*is_inexact_gt_midpoint = 0
		} else { // else MP in [ODD, EVEN]
			*is_midpoint_lt_even = 1
			*is_inexact_lt_midpoint = 0
			*is_inexact_gt_midpoint = 0
		}
	}
	// check for rounding overflow, which occurs if Cstar = 10^(q-x)
	ind = q - x                    // 1 <= ind <= q - 1
	if Cstar == bid_ten2k64[ind] { // if  Cstar = 10^(q-x)
		Cstar = bid_ten2k64[ind-1] // Cstar = 10^(q-x-1)
		*incr_exp = 1
	} else { // 10^33 <= Cstar <= 10^34 - 1
		*incr_exp = 0
	}
	return Cstar
}

func bid_round128_19_38_for64(q, x int, C uint64, incr_exp *int,
	is_midpoint_lt_even, is_midpoint_gt_even,
	is_inexact_lt_midpoint, is_inexact_gt_midpoint *int) uint64 {
	var P256 BID_UINT256
	var fstar BID_UINT256
	var Cstar BID_UINT128
	var C128 BID_UINT128
	var tmp64 uint64
	var shift int
	var ind int

	*incr_exp = 0
	*is_midpoint_lt_even = 0
	*is_midpoint_gt_even = 0
	*is_inexact_lt_midpoint = 0
	*is_inexact_gt_midpoint = 0

	// for bid64_from_uint64 q=20, x=4 경로에서 호출됨
	ind = x - 1 // 0 <= ind <= 18
	if ind < 0 || ind > 18 {
		return 0
	}

	// C = C + 1/2 * 10^x
	C128.lo = C
	C128.hi = 0
	tmp64 = C128.lo
	C128.lo = C128.lo + bid_midpoint64[ind]
	if C128.lo < tmp64 {
		C128.hi++
	}

	// P256 = (C + 1/2 * 10^x) * Kx
	P256 = __mul_128x128_to_256(C128, bid_Kx128_for64[ind])

	// Cstar = P256 >> Ex, fstar = low Ex bits
	shift = int(bid_Ex128m128_for64[ind])
	Cstar.lo = (P256.w2 >> shift) | (P256.w3 << (64 - shift))
	Cstar.hi = P256.w3 >> shift
	fstar.w0 = P256.w0
	fstar.w1 = P256.w1
	fstar.w2 = P256.w2 & bid_mask128_for64[ind]
	fstar.w3 = 0

	// determine inexactness
	if fstar.w2 > bid_half128_for64[ind] ||
		(fstar.w2 == bid_half128_for64[ind] && (fstar.w1 != 0 || fstar.w0 != 0)) {
		tmp64 = fstar.w2 - bid_half128_for64[ind]
		if tmp64 != 0 ||
			fstar.w1 > bid_ten2mxtrunc128_for64[ind].hi ||
			(fstar.w1 == bid_ten2mxtrunc128_for64[ind].hi &&
				fstar.w0 > bid_ten2mxtrunc128_for64[ind].lo) {
			*is_inexact_lt_midpoint = 1
		}
	} else {
		*is_inexact_gt_midpoint = 1
	}

	// check for midpoints
	if fstar.w3 == 0 && fstar.w2 == 0 &&
		(fstar.w1 < bid_ten2mxtrunc128_for64[ind].hi ||
			(fstar.w1 == bid_ten2mxtrunc128_for64[ind].hi &&
				fstar.w0 <= bid_ten2mxtrunc128_for64[ind].lo)) {
		if Cstar.lo&0x01 != 0 {
			Cstar.lo--
			if Cstar.lo == 0xffffffffffffffff {
				Cstar.hi--
			}
			*is_midpoint_gt_even = 1
			*is_inexact_lt_midpoint = 0
			*is_inexact_gt_midpoint = 0
		} else {
			*is_midpoint_lt_even = 1
			*is_inexact_lt_midpoint = 0
			*is_inexact_gt_midpoint = 0
		}
	}

	// check for rounding overflow: Cstar = 10^(q-x)
	ind = q - x
	if ind <= 19 {
		if Cstar.hi == 0x0 && Cstar.lo == bid_ten2k64[ind] {
			Cstar.lo = bid_ten2k64[ind-1]
			*incr_exp = 1
		} else {
			*incr_exp = 0
		}
	} else if ind == 20 {
		if Cstar.hi == 0x0000000000000005 &&
			Cstar.lo == 0x6bc75e2d63100000 {
			Cstar.lo = bid_ten2k64[19]
			Cstar.hi = 0x0
			*incr_exp = 1
		} else {
			*incr_exp = 0
		}
	} else {
		*incr_exp = 0
	}
	return Cstar.lo
}
