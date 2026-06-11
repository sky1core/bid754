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
	Cstar = P128.w[1] >> shift
	fstar.w[1] = P128.w[1] & bid_mask64[ind]
	fstar.w[0] = P128.w[0]
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
	if fstar.w[1] > bid_half64[ind] ||
		(fstar.w[1] == bid_half64[ind] && fstar.w[0] != 0) {
		// f* > 1/2 and the result may be exact
		// Calculate f* - 1/2
		tmp64 = fstar.w[1] - bid_half64[ind]
		if tmp64 != 0 || fstar.w[0] > bid_ten2mxtrunc64[ind] { // f* - 1/2 > 10^(-x)
			*is_inexact_lt_midpoint = 1
		} // else the result is exact
	} else { // the result is inexact; f2* <= 1/2
		*is_inexact_gt_midpoint = 1
	}
	// check for midpoints (could do this before determining inexactness)
	if fstar.w[1] == 0 && fstar.w[0] <= bid_ten2mxtrunc64[ind] {
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

// bid_round128_19_38_for64 - 128비트 라운딩 (64비트 입력용)
// Intel bid_round128_19_38의 uint64 입력 경로(1 <= x <= 19) 기계적 포팅
// Source: third_party/intel_dfp/src/bid_round.c
var bid_Kx128_for64 = [19]BID_UINT128{
	{w: [2]uint64{0xcccccccccccccccd, 0xcccccccccccccccc}},
	{w: [2]uint64{0x3d70a3d70a3d70a4, 0xa3d70a3d70a3d70a}},
	{w: [2]uint64{0x645a1cac083126ea, 0x83126e978d4fdf3b}},
	{w: [2]uint64{0xd3c36113404ea4a9, 0xd1b71758e219652b}},
	{w: [2]uint64{0x0fcf80dc33721d54, 0xa7c5ac471b478423}},
	{w: [2]uint64{0xa63f9a49c2c1b110, 0x8637bd05af6c69b5}},
	{w: [2]uint64{0x3d32907604691b4d, 0xd6bf94d5e57a42bc}},
	{w: [2]uint64{0xfdc20d2b36ba7c3e, 0xabcc77118461cefc}},
	{w: [2]uint64{0x31680a88f8953031, 0x89705f4136b4a597}},
	{w: [2]uint64{0xb573440e5a884d1c, 0xdbe6fecebdedd5be}},
	{w: [2]uint64{0xf78f69a51539d749, 0xafebff0bcb24aafe}},
	{w: [2]uint64{0xf93f87b7442e45d4, 0x8cbccc096f5088cb}},
	{w: [2]uint64{0x2865a5f206b06fba, 0xe12e13424bb40e13}},
	{w: [2]uint64{0x538484c19ef38c95, 0xb424dc35095cd80f}},
	{w: [2]uint64{0x0f9d37014bf60a11, 0x901d7cf73ab0acd9}},
	{w: [2]uint64{0x4c2ebe687989a9b4, 0xe69594bec44de15b}},
	{w: [2]uint64{0x09befeb9fad487c3, 0xb877aa3236a4b449}},
	{w: [2]uint64{0x3aff322e62439fd0, 0x9392ee8e921d5d07}},
	{w: [2]uint64{0x2b31e9e3d06c32e6, 0xec1e4a7db69561a5}},
}

var bid_Ex128m128_for64 = [19]uint32{
	3, 6, 9, 13, 16, 19, 23, 26, 29, 33, 36, 39, 43, 46, 49, 53, 56, 59, 63,
}

var bid_half128_for64 = [19]uint64{
	0x0000000000000004, 0x0000000000000020, 0x0000000000000100,
	0x0000000000001000, 0x0000000000008000, 0x0000000000040000,
	0x0000000000400000, 0x0000000002000000, 0x0000000010000000,
	0x0000000100000000, 0x0000000800000000, 0x0000004000000000,
	0x0000040000000000, 0x0000200000000000, 0x0001000000000000,
	0x0010000000000000, 0x0080000000000000, 0x0400000000000000,
	0x4000000000000000,
}

var bid_mask128_for64 = [19]uint64{
	0x0000000000000007, 0x000000000000003f, 0x00000000000001ff,
	0x0000000000001fff, 0x000000000000ffff, 0x000000000007ffff,
	0x00000000007fffff, 0x0000000003ffffff, 0x000000001fffffff,
	0x00000001ffffffff, 0x0000000fffffffff, 0x0000007fffffffff,
	0x000007ffffffffff, 0x00003fffffffffff, 0x0001ffffffffffff,
	0x001fffffffffffff, 0x00ffffffffffffff, 0x07ffffffffffffff,
	0x7fffffffffffffff,
}

var bid_ten2mxtrunc128_for64 = [19]BID_UINT128{
	{w: [2]uint64{0xcccccccccccccccc, 0xcccccccccccccccc}},
	{w: [2]uint64{0x3d70a3d70a3d70a3, 0xa3d70a3d70a3d70a}},
	{w: [2]uint64{0x645a1cac083126e9, 0x83126e978d4fdf3b}},
	{w: [2]uint64{0xd3c36113404ea4a8, 0xd1b71758e219652b}},
	{w: [2]uint64{0x0fcf80dc33721d53, 0xa7c5ac471b478423}},
	{w: [2]uint64{0xa63f9a49c2c1b10f, 0x8637bd05af6c69b5}},
	{w: [2]uint64{0x3d32907604691b4c, 0xd6bf94d5e57a42bc}},
	{w: [2]uint64{0xfdc20d2b36ba7c3d, 0xabcc77118461cefc}},
	{w: [2]uint64{0x31680a88f8953030, 0x89705f4136b4a597}},
	{w: [2]uint64{0xb573440e5a884d1b, 0xdbe6fecebdedd5be}},
	{w: [2]uint64{0xf78f69a51539d748, 0xafebff0bcb24aafe}},
	{w: [2]uint64{0xf93f87b7442e45d3, 0x8cbccc096f5088cb}},
	{w: [2]uint64{0x2865a5f206b06fb9, 0xe12e13424bb40e13}},
	{w: [2]uint64{0x538484c19ef38c94, 0xb424dc35095cd80f}},
	{w: [2]uint64{0x0f9d37014bf60a10, 0x901d7cf73ab0acd9}},
	{w: [2]uint64{0x4c2ebe687989a9b3, 0xe69594bec44de15b}},
	{w: [2]uint64{0x09befeb9fad487c2, 0xb877aa3236a4b449}},
	{w: [2]uint64{0x3aff322e62439fcf, 0x9392ee8e921d5d07}},
	{w: [2]uint64{0x2b31e9e3d06c32e5, 0xec1e4a7db69561a5}},
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
	C128.w[0] = C
	C128.w[1] = 0
	tmp64 = C128.w[0]
	C128.w[0] = C128.w[0] + bid_midpoint64[ind]
	if C128.w[0] < tmp64 {
		C128.w[1]++
	}

	// P256 = (C + 1/2 * 10^x) * Kx
	P256 = __mul_128x128_to_256(C128, bid_Kx128_for64[ind])

	// Cstar = P256 >> Ex, fstar = low Ex bits
	shift = int(bid_Ex128m128_for64[ind])
	Cstar.w[0] = (P256.w[2] >> shift) | (P256.w[3] << (64 - shift))
	Cstar.w[1] = P256.w[3] >> shift
	fstar.w[0] = P256.w[0]
	fstar.w[1] = P256.w[1]
	fstar.w[2] = P256.w[2] & bid_mask128_for64[ind]
	fstar.w[3] = 0

	// determine inexactness
	if fstar.w[2] > bid_half128_for64[ind] ||
		(fstar.w[2] == bid_half128_for64[ind] && (fstar.w[1] != 0 || fstar.w[0] != 0)) {
		tmp64 = fstar.w[2] - bid_half128_for64[ind]
		if tmp64 != 0 ||
			fstar.w[1] > bid_ten2mxtrunc128_for64[ind].w[1] ||
			(fstar.w[1] == bid_ten2mxtrunc128_for64[ind].w[1] &&
				fstar.w[0] > bid_ten2mxtrunc128_for64[ind].w[0]) {
			*is_inexact_lt_midpoint = 1
		}
	} else {
		*is_inexact_gt_midpoint = 1
	}

	// check for midpoints
	if fstar.w[3] == 0 && fstar.w[2] == 0 &&
		(fstar.w[1] < bid_ten2mxtrunc128_for64[ind].w[1] ||
			(fstar.w[1] == bid_ten2mxtrunc128_for64[ind].w[1] &&
				fstar.w[0] <= bid_ten2mxtrunc128_for64[ind].w[0])) {
		if Cstar.w[0]&0x01 != 0 {
			Cstar.w[0]--
			if Cstar.w[0] == 0xffffffffffffffff {
				Cstar.w[1]--
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
		if Cstar.w[1] == 0x0 && Cstar.w[0] == bid_ten2k64[ind] {
			Cstar.w[0] = bid_ten2k64[ind-1]
			*incr_exp = 1
		} else {
			*incr_exp = 0
		}
	} else if ind == 20 {
		if Cstar.w[1] == 0x0000000000000005 &&
			Cstar.w[0] == 0x6bc75e2d63100000 {
			Cstar.w[0] = bid_ten2k64[19]
			Cstar.w[1] = 0x0
			*incr_exp = 1
		} else {
			*incr_exp = 0
		}
	} else {
		*incr_exp = 0
	}
	return Cstar.w[0]
}
