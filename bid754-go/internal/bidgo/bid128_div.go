// Ported from: Intel bid128_div.c, bid_div_macros.h, bid_internal.h
// Mechanical translation - all logic preserved exactly.

package bidgo

import "math"

// Constants needed for bid128_div
const (
	LARGEST_BID128_HIGH = 0x5fffed09bead87c0
	LARGEST_BID128_LOW  = 0x378d8e63ffffffff
)

// bid_power10_index_binexp_128 contains powers of 10 indexed by binary exponent (128-bit version)
// Ported from bid_decimal_data.c
var bid_power10_index_binexp_128 = []BID_UINT128{
	{lo: 0x000000000000000a, hi: 0x0000000000000000}, // 0
	{lo: 0x000000000000000a, hi: 0x0000000000000000}, // 1
	{lo: 0x000000000000000a, hi: 0x0000000000000000}, // 2
	{lo: 0x000000000000000a, hi: 0x0000000000000000}, // 3
	{lo: 0x0000000000000064, hi: 0x0000000000000000}, // 4
	{lo: 0x0000000000000064, hi: 0x0000000000000000}, // 5
	{lo: 0x0000000000000064, hi: 0x0000000000000000}, // 6
	{lo: 0x00000000000003e8, hi: 0x0000000000000000}, // 7
	{lo: 0x00000000000003e8, hi: 0x0000000000000000}, // 8
	{lo: 0x00000000000003e8, hi: 0x0000000000000000}, // 9
	{lo: 0x0000000000002710, hi: 0x0000000000000000}, // 10
	{lo: 0x0000000000002710, hi: 0x0000000000000000}, // 11
	{lo: 0x0000000000002710, hi: 0x0000000000000000}, // 12
	{lo: 0x0000000000002710, hi: 0x0000000000000000}, // 13
	{lo: 0x00000000000186a0, hi: 0x0000000000000000}, // 14
	{lo: 0x00000000000186a0, hi: 0x0000000000000000}, // 15
	{lo: 0x00000000000186a0, hi: 0x0000000000000000}, // 16
	{lo: 0x00000000000f4240, hi: 0x0000000000000000}, // 17
	{lo: 0x00000000000f4240, hi: 0x0000000000000000}, // 18
	{lo: 0x00000000000f4240, hi: 0x0000000000000000}, // 19
	{lo: 0x0000000000989680, hi: 0x0000000000000000}, // 20
	{lo: 0x0000000000989680, hi: 0x0000000000000000}, // 21
	{lo: 0x0000000000989680, hi: 0x0000000000000000}, // 22
	{lo: 0x0000000000989680, hi: 0x0000000000000000}, // 23
	{lo: 0x0000000005f5e100, hi: 0x0000000000000000}, // 24
	{lo: 0x0000000005f5e100, hi: 0x0000000000000000}, // 25
	{lo: 0x0000000005f5e100, hi: 0x0000000000000000}, // 26
	{lo: 0x000000003b9aca00, hi: 0x0000000000000000}, // 27
	{lo: 0x000000003b9aca00, hi: 0x0000000000000000}, // 28
	{lo: 0x000000003b9aca00, hi: 0x0000000000000000}, // 29
	{lo: 0x00000002540be400, hi: 0x0000000000000000}, // 30
	{lo: 0x00000002540be400, hi: 0x0000000000000000}, // 31
	{lo: 0x00000002540be400, hi: 0x0000000000000000}, // 32
	{lo: 0x00000002540be400, hi: 0x0000000000000000}, // 33
	{lo: 0x000000174876e800, hi: 0x0000000000000000}, // 34
	{lo: 0x000000174876e800, hi: 0x0000000000000000}, // 35
	{lo: 0x000000174876e800, hi: 0x0000000000000000}, // 36
	{lo: 0x000000e8d4a51000, hi: 0x0000000000000000}, // 37
	{lo: 0x000000e8d4a51000, hi: 0x0000000000000000}, // 38
	{lo: 0x000000e8d4a51000, hi: 0x0000000000000000}, // 39
	{lo: 0x000009184e72a000, hi: 0x0000000000000000}, // 40
	{lo: 0x000009184e72a000, hi: 0x0000000000000000}, // 41
	{lo: 0x000009184e72a000, hi: 0x0000000000000000}, // 42
	{lo: 0x000009184e72a000, hi: 0x0000000000000000}, // 43
	{lo: 0x00005af3107a4000, hi: 0x0000000000000000}, // 44
	{lo: 0x00005af3107a4000, hi: 0x0000000000000000}, // 45
	{lo: 0x00005af3107a4000, hi: 0x0000000000000000}, // 46
	{lo: 0x00038d7ea4c68000, hi: 0x0000000000000000}, // 47
	{lo: 0x00038d7ea4c68000, hi: 0x0000000000000000}, // 48
	{lo: 0x00038d7ea4c68000, hi: 0x0000000000000000}, // 49
	{lo: 0x002386f26fc10000, hi: 0x0000000000000000}, // 50
	{lo: 0x002386f26fc10000, hi: 0x0000000000000000}, // 51
	{lo: 0x002386f26fc10000, hi: 0x0000000000000000}, // 52
	{lo: 0x002386f26fc10000, hi: 0x0000000000000000}, // 53
	{lo: 0x016345785d8a0000, hi: 0x0000000000000000}, // 54
	{lo: 0x016345785d8a0000, hi: 0x0000000000000000}, // 55
	{lo: 0x016345785d8a0000, hi: 0x0000000000000000}, // 56
	{lo: 0x0de0b6b3a7640000, hi: 0x0000000000000000}, // 57
	{lo: 0x0de0b6b3a7640000, hi: 0x0000000000000000}, // 58
	{lo: 0x0de0b6b3a7640000, hi: 0x0000000000000000}, // 59
	{lo: 0x8ac7230489e80000, hi: 0x0000000000000000}, // 60
	{lo: 0x8ac7230489e80000, hi: 0x0000000000000000}, // 61
	{lo: 0x8ac7230489e80000, hi: 0x0000000000000000}, // 62
	{lo: 0x8ac7230489e80000, hi: 0x0000000000000000}, // 63
	{lo: 0x6bc75e2d63100000, hi: 0x0000000000000005}, // 64: 10^20
	{lo: 0x6bc75e2d63100000, hi: 0x0000000000000005}, // 65
	{lo: 0x6bc75e2d63100000, hi: 0x0000000000000005}, // 66
	{lo: 0x35c9adc5dea00000, hi: 0x0000000000000036}, // 67: 10^21
	{lo: 0x35c9adc5dea00000, hi: 0x0000000000000036}, // 68
	{lo: 0x35c9adc5dea00000, hi: 0x0000000000000036}, // 69
	{lo: 0x19e0c9bab2400000, hi: 0x000000000000021e}, // 70: 10^22
	{lo: 0x19e0c9bab2400000, hi: 0x000000000000021e}, // 71
	{lo: 0x19e0c9bab2400000, hi: 0x000000000000021e}, // 72
	{lo: 0x19e0c9bab2400000, hi: 0x000000000000021e}, // 73
	{lo: 0x02c7e14af6800000, hi: 0x000000000000152d}, // 74: 10^23
	{lo: 0x02c7e14af6800000, hi: 0x000000000000152d}, // 75
	{lo: 0x02c7e14af6800000, hi: 0x000000000000152d}, // 76
	{lo: 0x1bcecceda1000000, hi: 0x000000000000d3c2}, // 77: 10^24
	{lo: 0x1bcecceda1000000, hi: 0x000000000000d3c2}, // 78
	{lo: 0x1bcecceda1000000, hi: 0x000000000000d3c2}, // 79
	{lo: 0x161401484a000000, hi: 0x0000000000084595}, // 80: 10^25
	{lo: 0x161401484a000000, hi: 0x0000000000084595}, // 81
	{lo: 0x161401484a000000, hi: 0x0000000000084595}, // 82
	{lo: 0x161401484a000000, hi: 0x0000000000084595}, // 83
	{lo: 0xdcc80cd2e4000000, hi: 0x000000000052b7d2}, // 84: 10^26
	{lo: 0xdcc80cd2e4000000, hi: 0x000000000052b7d2}, // 85
	{lo: 0xdcc80cd2e4000000, hi: 0x000000000052b7d2}, // 86
	{lo: 0x9fd0803ce8000000, hi: 0x00000000033b2e3c}, // 87: 10^27
	{lo: 0x9fd0803ce8000000, hi: 0x00000000033b2e3c}, // 88
	{lo: 0x9fd0803ce8000000, hi: 0x00000000033b2e3c}, // 89
	{lo: 0x3e25026110000000, hi: 0x00000000204fce5e}, // 90: 10^28
	{lo: 0x3e25026110000000, hi: 0x00000000204fce5e}, // 91
	{lo: 0x3e25026110000000, hi: 0x00000000204fce5e}, // 92
	{lo: 0x3e25026110000000, hi: 0x00000000204fce5e}, // 93
	{lo: 0x6d7217caa0000000, hi: 0x00000001431e0fae}, // 94: 10^29
	{lo: 0x6d7217caa0000000, hi: 0x00000001431e0fae}, // 95
	{lo: 0x6d7217caa0000000, hi: 0x00000001431e0fae}, // 96
	{lo: 0x4674edea40000000, hi: 0x0000000c9f2c9cd0}, // 97: 10^30
	{lo: 0x4674edea40000000, hi: 0x0000000c9f2c9cd0}, // 98
	{lo: 0x4674edea40000000, hi: 0x0000000c9f2c9cd0}, // 99
	{lo: 0xc0914b2680000000, hi: 0x0000007e37be2022}, // 100: 10^31
	{lo: 0xc0914b2680000000, hi: 0x0000007e37be2022}, // 101
	{lo: 0xc0914b2680000000, hi: 0x0000007e37be2022}, // 102
	{lo: 0x85acef8100000000, hi: 0x000004ee2d6d415b}, // 103: 10^32
	{lo: 0x85acef8100000000, hi: 0x000004ee2d6d415b}, // 104
	{lo: 0x85acef8100000000, hi: 0x000004ee2d6d415b}, // 105
	{lo: 0x85acef8100000000, hi: 0x000004ee2d6d415b}, // 106
	{lo: 0x38c15b0a00000000, hi: 0x0000314dc6448d93}, // 107: 10^33
	{lo: 0x38c15b0a00000000, hi: 0x0000314dc6448d93}, // 108
	{lo: 0x38c15b0a00000000, hi: 0x0000314dc6448d93}, // 109: entry 112 in C
	{lo: 0x378d8e6400000000, hi: 0x0001ed09bead87c0}, // 110: 10^34
	{lo: 0x378d8e6400000000, hi: 0x0001ed09bead87c0}, // 111
	{lo: 0x378d8e6400000000, hi: 0x0001ed09bead87c0}, // 112
	{lo: 0x2b878fe800000000, hi: 0x0013426172c74d82}, // 113: 10^35
	{lo: 0x2b878fe800000000, hi: 0x0013426172c74d82}, // 114
	{lo: 0x2b878fe800000000, hi: 0x0013426172c74d82}, // 115
	{lo: 0x2b878fe800000000, hi: 0x0013426172c74d82}, // 116
	{lo: 0xb34b9f1000000000, hi: 0x00c097ce7bc90715}, // 117: 10^36
	{lo: 0x00f436a000000000, hi: 0x0785ee10d5da46d9}, // 118: 10^37
	{lo: 0x00f436a000000000, hi: 0x0785ee10d5da46d9}, // 119
	{lo: 0x00f436a000000000, hi: 0x0785ee10d5da46d9}, // 120
	{lo: 0x098a224000000000, hi: 0x4b3b4ca85a86c47a}, // 121: 10^38
	{lo: 0x098a224000000000, hi: 0x4b3b4ca85a86c47a}, // 122
	{lo: 0x098a224000000000, hi: 0x4b3b4ca85a86c47a}, // 123
	{lo: 0x098a224000000000, hi: 0x4b3b4ca85a86c47a}, // 124
}

// __mul_128x128_low multiplies two 128-bit values, returning only the low 128 bits.
// Ported from bid_internal.h
func __mul_128x128_low(A, B BID_UINT128) BID_UINT128 {
	var Ql BID_UINT128
	ALBL := __mul_64x64_to_128(A.lo, B.lo)
	QM64 := B.lo*A.hi + A.lo*B.hi

	Ql.lo = ALBL.lo
	Ql.hi = QM64 + ALBL.hi
	return Ql
}

// __sub_borrow_in_out subtracts with borrow in and returns borrow out
func __sub_borrow_in_out(x, y, ci uint64) (s uint64, co uint64) {
	x1 := x - ci
	if x1 > x {
		co = 1
	}
	s = x1 - y
	if s > x1 {
		co = 1
	}
	return
}

// bid___div_128_by_128 divides 128-bit CX by 128-bit CY.
// Returns quotient CQ and remainder CR.
// Ported from bid_div_macros.h (non-DOUBLE_EXTENDED path)
func bid___div_128_by_128(CX0, CY BID_UINT128) (CQ, CR BID_UINT128) {
	var CY36, CY51, A2, CQT BID_UINT128
	var Q uint64

	if CX0.hi == 0 && CY.hi == 0 {
		CQ.lo = CX0.lo / CY.lo
		CQ.hi = 0
		CR.hi = 0
		CR.lo = CX0.lo - CQ.lo*CY.lo
		return
	}

	CX := CX0

	// 2^64
	t64 := math.Float64frombits(0x43f0000000000000)
	lx := noFmaMulAddF64(float64(CX.hi), t64, float64(CX.lo))
	ly := noFmaMulAddF64(float64(CY.hi), t64, float64(CY.lo))
	lq := lx / ly

	CY36.hi = CY.lo >> (64 - 36)
	CY36.lo = CY.lo << 36

	CQ.hi = 0
	CQ.lo = 0

	// Q >= 2^100 ?
	if CY.hi == 0 && CY36.hi == 0 && (CX.hi >= CY36.lo) {
		// then Q >= 2^100

		// 2^(-60)*CX/CY
		d60 := math.Float64frombits(0x3c30000000000000)
		lq *= d60
		Q = uint64(lq) - 4

		// Q*CY
		A2 = __mul_64x64_to_128(Q, CY.lo)

		// A2 <<= 60
		A2.hi = (A2.hi << 60) | (A2.lo >> (64 - 60))
		A2.lo <<= 60

		CX = __sub_128_128(CX, A2)

		lx = noFmaMulAddF64(float64(CX.hi), t64, float64(CX.lo))
		lq = lx / ly

		CQ.hi = Q >> (64 - 60)
		CQ.lo = Q << 60
	}

	CY51.hi = (CY.hi << 51) | (CY.lo >> (64 - 51))
	CY51.lo = CY.lo << 51

	if CY.hi < (1<<(64-51)) && __unsigned_compare_gt_128(CX, CY51) {
		// Q > 2^51

		// 2^(-49)*CX/CY
		d49 := math.Float64frombits(0x3ce0000000000000)
		lq *= d49

		Q = uint64(lq) - 1

		// Q*CY
		A2 = __mul_64x64_to_128(Q, CY.lo)
		A2.hi += Q * CY.hi

		// A2 <<= 49
		A2.hi = (A2.hi << 49) | (A2.lo >> (64 - 49))
		A2.lo <<= 49

		CX = __sub_128_128(CX, A2)

		CQT.hi = Q >> (64 - 49)
		CQT.lo = Q << 49
		CQ = __add_128_128(CQ, CQT)

		lx = noFmaMulAddF64(float64(CX.hi), t64, float64(CX.lo))
		lq = lx / ly
	}

	Q = uint64(lq)

	A2 = __mul_64x64_to_128(Q, CY.lo)
	A2.hi += Q * CY.hi

	CX = __sub_128_128(CX, A2)
	if int64(CX.hi) < 0 {
		Q--
		CX.lo += CY.lo
		if CX.lo < CY.lo {
			CX.hi++
		}
		CX.hi += CY.hi
		if int64(CX.hi) < 0 {
			Q--
			CX.lo += CY.lo
			if CX.lo < CY.lo {
				CX.hi++
			}
			CX.hi += CY.hi
		}
	} else if __unsigned_compare_ge_128(CX, CY) {
		Q++
		CX = __sub_128_128(CX, CY)
	}

	CQ = __add_128_64(CQ, Q)

	CR.hi = CX.hi
	CR.lo = CX.lo
	return
}

// bid___div_256_by_128 divides 256-bit CA4 by 128-bit CY.
// CQ is initial quotient (accumulated), CA4 is modified to hold remainder.
// Ported from bid_div_macros.h (non-DOUBLE_EXTENDED path)
func bid___div_256_by_128(pCQ *BID_UINT128, pCA4 *BID_UINT256, CY BID_UINT128) {
	var CA4 BID_UINT256
	var CA2 [3]uint64 // 192-bit
	var CQ, A2, A2h, CQT BID_UINT128
	var Q, carry64 uint64

	// the quotient is assumed to be at most 113 bits,
	// as needed by BID128 divide routines

	// initial dividend
	CA4.w3 = pCA4.w3
	CA4.w2 = pCA4.w2
	CA4.w1 = pCA4.w1
	CA4.w0 = pCA4.w0
	CQ.hi = pCQ.hi
	CQ.lo = pCQ.lo

	// 2^64
	t64 := math.Float64frombits(0x43f0000000000000)
	d128 := t64 * t64
	d192 := d128 * t64
	lx := noFmaMulAddF64(float64(CA4.w3), d192,
		noFmaMulAddF64(float64(CA4.w2), d128,
			noFmaMulAddF64(float64(CA4.w1), t64, float64(CA4.w0))))
	ly := noFmaMulAddF64(float64(CY.hi), t64, float64(CY.lo))
	lq := lx / ly

	var CY36_2, CY36_1, CY36_0 uint64
	CY36_2 = CY.hi >> (64 - 36)
	CY36_1 = (CY.hi << 36) | (CY.lo >> (64 - 36))
	CY36_0 = CY.lo << 36

	// Q >= 2^100 ?
	if CA4.w3 > CY36_2 ||
		(CA4.w3 == CY36_2 &&
			(CA4.w2 > CY36_1 ||
				(CA4.w2 == CY36_1 && CA4.w1 >= CY36_0))) {
		// 2^(-60)*CA4/CY
		d60 := math.Float64frombits(0x3c30000000000000)
		lq *= d60
		Q = uint64(lq) - 4

		// Q*CY
		tmp192 := __mul_64x128_to_192(Q, CY)

		// CA2 <<= 60
		CA2[2] = (tmp192.w2 << 60) | (tmp192.w1 >> (64 - 60))
		CA2[1] = (tmp192.w1 << 60) | (tmp192.w0 >> (64 - 60))
		CA2[0] = tmp192.w0 << 60

		// CA4 -= CA2
		CA4.w0, carry64 = __sub_borrow_out(CA4.w0, CA2[0])
		CA4.w1, carry64 = __sub_borrow_in_out(CA4.w1, CA2[1], carry64)
		CA4.w2 = CA4.w2 - CA2[2] - carry64

		lx = noFmaMulAddF64(float64(CA4.w2), d128,
			noFmaMulAddF64(float64(CA4.w1), t64, float64(CA4.w0)))
		lq = lx / ly

		CQT.hi = Q >> (64 - 60)
		CQT.lo = Q << 60
		CQ = __add_128_128(CQ, CQT)
	}

	var CY51_2, CY51_1, CY51_0 uint64
	CY51_2 = CY.hi >> (64 - 51)
	CY51_1 = (CY.hi << 51) | (CY.lo >> (64 - 51))
	CY51_0 = CY.lo << 51

	// compare CA4 with CY51 (as 192-bit values)
	ca4_128 := BID_UINT128{lo: CA4.w0, hi: CA4.w1}
	cy51_128 := BID_UINT128{lo: CY51_0, hi: CY51_1}

	if CA4.w2 > CY51_2 || ((CA4.w2 == CY51_2) &&
		__unsigned_compare_gt_128(ca4_128, cy51_128)) {
		// Q > 2^51

		// 2^(-49)*CA4/CY
		d49 := math.Float64frombits(0x3ce0000000000000)
		lq *= d49

		Q = uint64(lq) - 1

		// Q*CY
		A2 = __mul_64x64_to_128(Q, CY.lo)
		A2h = __mul_64x64_to_128(Q, CY.hi)
		A2.hi += A2h.lo
		if A2.hi < A2h.lo {
			A2h.hi++
		}

		// A2 <<= 49
		CA2[2] = (A2h.hi << 49) | (A2.hi >> (64 - 49))
		CA2[1] = (A2.hi << 49) | (A2.lo >> (64 - 49))
		CA2[0] = A2.lo << 49

		CA4.w0, carry64 = __sub_borrow_out(CA4.w0, CA2[0])
		CA4.w1, carry64 = __sub_borrow_in_out(CA4.w1, CA2[1], carry64)
		CA4.w2 = CA4.w2 - CA2[2] - carry64

		CQT.hi = Q >> (64 - 49)
		CQT.lo = Q << 49
		CQ = __add_128_128(CQ, CQT)

		lx = noFmaMulAddF64(float64(CA4.w2), d128,
			noFmaMulAddF64(float64(CA4.w1), t64, float64(CA4.w0)))
		lq = lx / ly
	}

	Q = uint64(lq)
	A2 = __mul_64x64_to_128(Q, CY.lo)
	A2.hi += Q * CY.hi

	// __sub_128_128(CA4, CA4, A2) - using CA4.w[0:1] as 128-bit
	tmpCA := BID_UINT128{lo: CA4.w0, hi: CA4.w1}
	tmpCA = __sub_128_128(tmpCA, A2)
	CA4.w0 = tmpCA.lo
	CA4.w1 = tmpCA.hi

	if int64(CA4.w1) < 0 {
		Q--
		CA4.w0 += CY.lo
		if CA4.w0 < CY.lo {
			CA4.w1++
		}
		CA4.w1 += CY.hi
		if int64(CA4.w1) < 0 {
			Q--
			CA4.w0 += CY.lo
			if CA4.w0 < CY.lo {
				CA4.w1++
			}
			CA4.w1 += CY.hi
		}
	} else if CA4.w1 > CY.hi || (CA4.w1 == CY.hi && CA4.w0 >= CY.lo) {
		Q++
		tmpCA2 := BID_UINT128{lo: CA4.w0, hi: CA4.w1}
		tmpCA2 = __sub_128_128(tmpCA2, CY)
		CA4.w0 = tmpCA2.lo
		CA4.w1 = tmpCA2.hi
	}

	CQ = __add_128_64(CQ, Q)

	pCQ.hi = CQ.hi
	pCQ.lo = CQ.lo
	pCA4.w1 = CA4.w1
	pCA4.w0 = CA4.w0
}

// handle_UF_128 handles BID128 underflow (without remainder).
// Ported from bid_internal.h
func handle_UF_128(sgn uint64, expon int, CQ BID_UINT128,
	prounding_mode int, fpsc *uint32) BID_UINT128 {
	var res BID_UINT128
	var T128, TP128, Qh, Ql, Qh1, Stemp, Tmp, Tmp1 BID_UINT128
	var carry, CY uint64
	var ed2, amount int
	var rmode uint
	var status uint32

	// UF occurs
	if expon+MAX_FORMAT_DIGITS_128 < 0 {
		*fpsc |= BID_UNDERFLOW_EXCEPTION | BID_INEXACT_EXCEPTION
		res.hi = sgn
		res.lo = 0
		if (sgn != 0 && prounding_mode == BID_ROUNDING_DOWN) ||
			(sgn == 0 && prounding_mode == BID_ROUNDING_UP) {
			res.lo = 1
		}
		return res
	}

	ed2 = 0 - expon
	// add rounding constant to CQ
	rmode = uint(prounding_mode)
	if sgn != 0 && uint(rmode-1) < 2 {
		rmode = 3 - rmode
	}

	T128 = bid_round_const_table_128[rmode][ed2]
	CQ.lo, carry = __add_carry_out(T128.lo, CQ.lo)
	CQ.hi = CQ.hi + T128.hi + carry

	TP128 = bid_reciprocals10_128[ed2]
	Qh, Ql = __mul_128x128_full(CQ, TP128)
	amount = bid_recip_scale[ed2]

	if amount >= 64 {
		CQ.lo = Qh.hi >> uint(amount-64)
		CQ.hi = 0
	} else {
		CQ = __shr_128(Qh, uint(amount))
	}

	expon = 0

	if prounding_mode == BID_ROUNDING_TO_NEAREST {
		if CQ.lo&1 != 0 {
			// check whether fractional part of initial_P/10^ed1 is exactly .5

			// get remainder
			Qh1 = __shl_128_long(Qh, uint(128-amount))

			if Qh1.hi == 0 && Qh1.lo == 0 &&
				(Ql.hi < bid_reciprocals10_128[ed2].hi ||
					(Ql.hi == bid_reciprocals10_128[ed2].hi &&
						Ql.lo < bid_reciprocals10_128[ed2].lo)) {
				CQ.lo--
			}
		}
	}

	if (*fpsc & BID_INEXACT_EXCEPTION) != 0 {
		*fpsc |= BID_UNDERFLOW_EXCEPTION
	} else {
		status = BID_INEXACT_EXCEPTION
		// get remainder
		Qh1 = __shl_128_long(Qh, uint(128-amount))

		switch rmode {
		case BID_ROUNDING_TO_NEAREST, BID_ROUNDING_TIES_AWAY:
			// test whether fractional part is 0
			if Qh1.hi == 0x8000000000000000 && Qh1.lo == 0 &&
				(Ql.hi < bid_reciprocals10_128[ed2].hi ||
					(Ql.hi == bid_reciprocals10_128[ed2].hi &&
						Ql.lo < bid_reciprocals10_128[ed2].lo)) {
				status = BID_EXACT_STATUS
			}
		case BID_ROUNDING_DOWN, BID_ROUNDING_TO_ZERO:
			if Qh1.hi == 0 && Qh1.lo == 0 &&
				(Ql.hi < bid_reciprocals10_128[ed2].hi ||
					(Ql.hi == bid_reciprocals10_128[ed2].hi &&
						Ql.lo < bid_reciprocals10_128[ed2].lo)) {
				status = BID_EXACT_STATUS
			}
		default:
			// round up
			Stemp.lo, CY = __add_carry_out(Ql.lo, bid_reciprocals10_128[ed2].lo)
			Stemp.hi, carry = __add_carry_in_out(Ql.hi, bid_reciprocals10_128[ed2].hi, CY)
			_ = Stemp
			Qh = __shr_128_long(Qh1, uint(128-amount))
			Tmp.lo = 1
			Tmp.hi = 0
			Tmp1 = __shl_128_long(Tmp, uint(amount))
			Qh.lo += carry
			if Qh.lo < carry {
				Qh.hi++
			}
			if __unsigned_compare_ge_128(Qh, Tmp1) {
				status = BID_EXACT_STATUS
			}
		}

		if status != BID_EXACT_STATUS {
			*fpsc |= BID_UNDERFLOW_EXCEPTION | status
		}
	}

	res.hi = sgn | CQ.hi
	res.lo = CQ.lo

	return res
}

// bid_handle_UF_128_rem handles BID128 underflow with remainder.
// Ported from bid_internal.h
func bid_handle_UF_128_rem(sgn uint64, expon int, CQ BID_UINT128,
	R uint64, prounding_mode int, fpsc *uint32) BID_UINT128 {
	var res BID_UINT128
	var T128, TP128, Qh, Ql, Qh1, Stemp, Tmp, Tmp1, CQ2, CQ8 BID_UINT128
	var carry, CY uint64
	var ed2, amount int
	var rmode uint
	var status uint32

	// UF occurs
	if expon+MAX_FORMAT_DIGITS_128 < 0 {
		*fpsc |= BID_UNDERFLOW_EXCEPTION | BID_INEXACT_EXCEPTION
		res.hi = sgn
		res.lo = 0
		if (sgn != 0 && prounding_mode == BID_ROUNDING_DOWN) ||
			(sgn == 0 && prounding_mode == BID_ROUNDING_UP) {
			res.lo = 1
		}
		return res
	}
	// CQ *= 10
	CQ2.hi = (CQ.hi << 1) | (CQ.lo >> 63)
	CQ2.lo = CQ.lo << 1
	CQ8.hi = (CQ.hi << 3) | (CQ.lo >> 61)
	CQ8.lo = CQ.lo << 3
	CQ = __add_128_128(CQ2, CQ8)

	// add remainder
	if R != 0 {
		CQ.lo |= 1
	}

	ed2 = 1 - expon
	// add rounding constant to CQ
	rmode = uint(prounding_mode)
	if sgn != 0 && uint(rmode-1) < 2 {
		rmode = 3 - rmode
	}
	T128 = bid_round_const_table_128[rmode][ed2]
	CQ.lo, carry = __add_carry_out(T128.lo, CQ.lo)
	CQ.hi = CQ.hi + T128.hi + carry

	TP128 = bid_reciprocals10_128[ed2]
	Qh, Ql = __mul_128x128_full(CQ, TP128)
	amount = bid_recip_scale[ed2]

	if amount >= 64 {
		CQ.lo = Qh.hi >> uint(amount-64)
		CQ.hi = 0
	} else {
		CQ = __shr_128(Qh, uint(amount))
	}

	expon = 0

	if prounding_mode == BID_ROUNDING_TO_NEAREST {
		if CQ.lo&1 != 0 {
			// check whether fractional part of initial_P/10^ed1 is exactly .5

			// get remainder
			Qh1 = __shl_128_long(Qh, uint(128-amount))

			if Qh1.hi == 0 && Qh1.lo == 0 &&
				(Ql.hi < bid_reciprocals10_128[ed2].hi ||
					(Ql.hi == bid_reciprocals10_128[ed2].hi &&
						Ql.lo < bid_reciprocals10_128[ed2].lo)) {
				CQ.lo--
			}
		}
	}

	if (*fpsc & BID_INEXACT_EXCEPTION) != 0 {
		*fpsc |= BID_UNDERFLOW_EXCEPTION
	} else {
		status = BID_INEXACT_EXCEPTION
		// get remainder
		Qh1 = __shl_128_long(Qh, uint(128-amount))

		switch rmode {
		case BID_ROUNDING_TO_NEAREST, BID_ROUNDING_TIES_AWAY:
			// test whether fractional part is 0
			if Qh1.hi == 0x8000000000000000 && Qh1.lo == 0 &&
				(Ql.hi < bid_reciprocals10_128[ed2].hi ||
					(Ql.hi == bid_reciprocals10_128[ed2].hi &&
						Ql.lo < bid_reciprocals10_128[ed2].lo)) {
				status = BID_EXACT_STATUS
			}
		case BID_ROUNDING_DOWN, BID_ROUNDING_TO_ZERO:
			if Qh1.hi == 0 && Qh1.lo == 0 &&
				(Ql.hi < bid_reciprocals10_128[ed2].hi ||
					(Ql.hi == bid_reciprocals10_128[ed2].hi &&
						Ql.lo < bid_reciprocals10_128[ed2].lo)) {
				status = BID_EXACT_STATUS
			}
		default:
			// round up
			Stemp.lo, CY = __add_carry_out(Ql.lo, bid_reciprocals10_128[ed2].lo)
			Stemp.hi, carry = __add_carry_in_out(Ql.hi, bid_reciprocals10_128[ed2].hi, CY)
			_ = Stemp
			Qh = __shr_128_long(Qh1, uint(128-amount))
			Tmp.lo = 1
			Tmp.hi = 0
			Tmp1 = __shl_128_long(Tmp, uint(amount))
			Qh.lo += carry
			if Qh.lo < carry {
				Qh.hi++
			}
			if __unsigned_compare_ge_128(Qh, Tmp1) {
				status = BID_EXACT_STATUS
			}
		}

		if status != BID_EXACT_STATUS {
			*fpsc |= BID_UNDERFLOW_EXCEPTION | status
		}
	}

	res.hi = sgn | CQ.hi
	res.lo = CQ.lo

	return res
}

// bid_get_BID128 packs sign, exponent, and coefficient into BID128
// with full overflow/underflow checking and rounding.
// Ported from bid_internal.h
func bid_get_BID128(sgn uint64, expon int, coeff BID_UINT128,
	prounding_mode int, fpsc *uint32) BID_UINT128 {
	var res BID_UINT128
	var T BID_UINT128
	var tmp, tmp2 uint64

	// coeff==10^34?
	if coeff.hi == 0x0001ed09bead87c0 && coeff.lo == 0x378d8e6400000000 {
		expon++
		// set coefficient to 10^33
		coeff.hi = 0x0000314dc6448d93
		coeff.lo = 0x38c15b0a00000000
	}
	// check OF, UF
	if expon < 0 || expon > DECIMAL_MAX_EXPON_128 {
		// check UF
		if expon < 0 {
			return handle_UF_128(sgn, expon, coeff, prounding_mode, fpsc)
		}

		if expon-MAX_FORMAT_DIGITS_128 <= DECIMAL_MAX_EXPON_128 {
			T = bid_power10_table_128[MAX_FORMAT_DIGITS_128-1]
			for __unsigned_compare_gt_128(T, coeff) && expon > DECIMAL_MAX_EXPON_128 {
				coeff.hi =
					(coeff.hi << 3) + (coeff.hi << 1) + (coeff.lo >> 61) +
						(coeff.lo >> 63)
				tmp2 = coeff.lo << 3
				coeff.lo = (coeff.lo << 1) + tmp2
				if coeff.lo < tmp2 {
					coeff.hi++
				}

				expon--
			}
		}
		if expon > DECIMAL_MAX_EXPON_128 {
			if coeff.hi == 0 && coeff.lo == 0 {
				res.hi = sgn | (uint64(DECIMAL_MAX_EXPON_128) << 49)
				res.lo = 0
				return res
			}
			// OF
			*fpsc |= BID_OVERFLOW_EXCEPTION | BID_INEXACT_EXCEPTION
			if prounding_mode == BID_ROUNDING_TO_ZERO ||
				(sgn != 0 && prounding_mode == BID_ROUNDING_UP) ||
				(sgn == 0 && prounding_mode == BID_ROUNDING_DOWN) {
				res.hi = sgn | LARGEST_BID128_HIGH
				res.lo = LARGEST_BID128_LOW
			} else {
				res.hi = sgn | INFINITY_MASK64
				res.lo = 0
			}
			return res
		}
	}

	res.lo = coeff.lo
	tmp = uint64(expon)
	tmp <<= 49
	res.hi = sgn | tmp | coeff.hi

	return res
}

// Bid128ddDiv is ported from bid128_div.c: bid128dd_div. BID64 operands are
// widened exactly before the existing BID128 division core.
func Bid128ddDiv(x, y uint64, rnd_mode int) (BID_UINT128, uint32) {
	x1, flagsX := Bid64ToBid128(x)
	y1, flagsY := Bid64ToBid128(y)
	res, opFlags := Bid128Div(x1, y1, rnd_mode)
	return res, flagsX | flagsY | opFlags
}

// Bid128dqDiv is ported from bid128_div.c: bid128dq_div.
func Bid128dqDiv(x uint64, y BID_UINT128, rnd_mode int) (BID_UINT128, uint32) {
	x1, flags := Bid64ToBid128(x)
	res, opFlags := Bid128Div(x1, y, rnd_mode)
	return res, flags | opFlags
}

// Bid128qdDiv is ported from bid128_div.c: bid128qd_div.
func Bid128qdDiv(x BID_UINT128, y uint64, rnd_mode int) (BID_UINT128, uint32) {
	y1, flags := Bid64ToBid128(y)
	res, opFlags := Bid128Div(x, y1, rnd_mode)
	return res, flags | opFlags
}

// Bid128Div divides x by y (BID128).
// Ported from bid128_div in bid128_div.c (line-by-line mechanical translation)
func Bid128Div(x, y BID_UINT128, rnd_mode int) (BID_UINT128, uint32) {
	var CA4, CA4r, P256 BID_UINT256
	var CX, CY, T128, CQ, CR, CA, TP128, Qh, res BID_UINT128
	var sign_x, sign_y, T, carry64, D, Q_high, Q_low, QX, PD uint64
	var valid_y bool
	var QX32, digit, digit_h, digit_low uint32
	var tdigit [3]uint32
	var exponent_x, exponent_y, bin_index, bin_expon, diff_expon, ed2,
		digits_q, amount int
	var nzeros, i, j, k, d5 int
	var rmode uint
	var pfpsf uint32

	sign_y, exponent_y, CY, valid_y = unpack_BID128_value(y)
	_ = valid_y

	// unpack arguments, check for NaN or Infinity
	sign_x_raw, exponent_x_raw, CX_raw, valid_x := unpack_BID128_value(x)
	sign_x = sign_x_raw
	exponent_x = exponent_x_raw
	CX = CX_raw

	if !valid_x {
		// test if x is NaN
		if (x.hi & 0x7c00000000000000) == 0x7c00000000000000 {
			if (x.hi&0x7e00000000000000) == 0x7e00000000000000 || // sNaN
				(y.hi&0x7e00000000000000) == 0x7e00000000000000 {
				pfpsf |= BID_INVALID_EXCEPTION
			}
			res.hi = (CX.hi) & QUIET_MASK64
			res.lo = CX.lo
			return res, pfpsf
		}
		// x is Infinity?
		if (x.hi & 0x7800000000000000) == 0x7800000000000000 {
			// check if y is Inf.
			if (y.hi & 0x7c00000000000000) == 0x7800000000000000 {
				// return NaN
				pfpsf |= BID_INVALID_EXCEPTION
				res.hi = 0x7c00000000000000
				res.lo = 0
				return res, pfpsf
			}
			// y is NaN?
			if (y.hi & 0x7c00000000000000) != 0x7c00000000000000 {
				// return +/-Inf
				res.hi = ((x.hi ^ y.hi) & 0x8000000000000000) |
					0x7800000000000000
				res.lo = 0
				return res, pfpsf
			}
		}
		// x is 0
		if (y.hi & 0x7800000000000000) < 0x7800000000000000 {
			if CY.lo == 0 && (CY.hi&0x0001ffffffffffff) == 0 {
				pfpsf |= BID_INVALID_EXCEPTION
				// x=y=0, return NaN
				res.hi = 0x7c00000000000000
				res.lo = 0
				return res, pfpsf
			}
			// return 0
			res.hi = (x.hi ^ y.hi) & 0x8000000000000000
			exponent_x = exponent_x - exponent_y + EXPONENT_BIAS128
			if exponent_x > DECIMAL_MAX_EXPON_128 {
				exponent_x = DECIMAL_MAX_EXPON_128
			} else if exponent_x < 0 {
				exponent_x = 0
			}
			res.hi |= uint64(exponent_x) << 49
			res.lo = 0
			return res, pfpsf
		}
	}
	if !valid_y {
		// y is Inf. or NaN

		// test if y is NaN
		if (y.hi & 0x7c00000000000000) == 0x7c00000000000000 {
			if (y.hi & 0x7e00000000000000) == 0x7e00000000000000 { // sNaN
				pfpsf |= BID_INVALID_EXCEPTION
			}
			res.hi = CY.hi & QUIET_MASK64
			res.lo = CY.lo
			return res, pfpsf
		}
		// y is Infinity?
		if (y.hi & 0x7800000000000000) == 0x7800000000000000 {
			// return +/-0
			res.hi = sign_x ^ sign_y
			res.lo = 0
			return res, pfpsf
		}
		// y is 0, return +/-Inf
		pfpsf |= BID_ZERO_DIVIDE_EXCEPTION
		res.hi =
			((x.hi ^ y.hi) & 0x8000000000000000) | 0x7800000000000000
		res.lo = 0
		return res, pfpsf
	}

	diff_expon = exponent_x - exponent_y + EXPONENT_BIAS128

	if __unsigned_compare_gt_128(CY, CX) {
		// CX < CY

		// 2^64
		f64_i := uint32(0x5f800000)
		f64_d := math.Float32frombits(f64_i)

		// fx ~ CX,   fy ~ CY
		fx_d := noFmaMulAddF32(float32(CX.hi), f64_d, float32(CX.lo))
		fy_d := noFmaMulAddF32(float32(CY.hi), f64_d, float32(CY.lo))
		fx_i := math.Float32bits(fx_d)
		fy_i := math.Float32bits(fy_d)
		// expon_cy - expon_cx
		bin_index = int((fy_i - fx_i) >> 23)

		if CX.hi != 0 {
			T = bid_power10_index_binexp_128[bin_index].lo
			CA = __mul_64x128_short(T, CX)
		} else {
			T128 = bid_power10_index_binexp_128[bin_index]
			CA = __mul_64x128_short(CX.lo, T128)
		}

		ed2 = 33
		if __unsigned_compare_gt_128(CY, CA) {
			ed2++
		}

		T128 = bid_power10_table_128[ed2]
		CA4 = __mul_128x128_to_256(CA, T128)

		ed2 += bid_estimate_decimal_digits[bin_index]
		CQ.lo = 0
		CQ.hi = 0
		diff_expon = diff_expon - ed2

	} else {
		// get CQ = CX/CY
		CQ, CR = bid___div_128_by_128(CX, CY)

		if CR.hi == 0 && CR.lo == 0 {
			res = bid_get_BID128(sign_x^sign_y, diff_expon, CQ, rnd_mode, &pfpsf)
			return res, pfpsf
		}
		// get number of decimal digits in CQ
		// 2^64
		f64_i := uint32(0x5f800000)
		f64_d := math.Float32frombits(f64_i)
		fx_d := noFmaMulAddF32(float32(CQ.hi), f64_d, float32(CQ.lo))
		fx_i := math.Float32bits(fx_d)
		// binary expon. of CQ
		bin_expon = int((fx_i - 0x3f800000) >> 23)

		digits_q = bid_estimate_decimal_digits[bin_expon]
		TP128.lo = bid_power10_index_binexp_128[bin_expon].lo
		TP128.hi = bid_power10_index_binexp_128[bin_expon].hi
		if __unsigned_compare_ge_128(CQ, TP128) {
			digits_q++
		}

		ed2 = 34 - digits_q
		T128.lo = bid_power10_table_128[ed2].lo
		T128.hi = bid_power10_table_128[ed2].hi
		CA4 = __mul_128x128_to_256(CR, T128)
		diff_expon = diff_expon - ed2
		CQ = __mul_128x128_low(CQ, T128)

	}

	bid___div_256_by_128(&CQ, &CA4, CY)

	if CA4.w0 != 0 || CA4.w1 != 0 {
		// set status flags
		pfpsf |= BID_INEXACT_EXCEPTION
	} else {
		// check whether result is exact
		// check whether CX, CY are short
		if CX.hi == 0 && CY.hi == 0 && (CX.lo <= 1024) && (CY.lo <= 1024) {
			i = int(CY.lo) - 1
			j = int(CX.lo) - 1
			// difference in powers of 2 bid_factors for Y and X
			nzeros = ed2 - int(bid_factors[i][0]) + int(bid_factors[j][0])
			// difference in powers of 5 bid_factors
			d5 = ed2 - int(bid_factors[i][1]) + int(bid_factors[j][1])
			if d5 < nzeros {
				nzeros = d5
			}
			// get P*(2^M[extra_digits])/10^extra_digits
			Qh, _ = __mul_128x128_full(CQ, bid_reciprocals10_128[nzeros])

			// now get P/10^extra_digits: shift Q_high right by M[extra_digits]-128
			amount = bid_recip_scale[nzeros]
			CQ = __shr_128_long(Qh, uint(amount))

			diff_expon += nzeros
		} else {
			// decompose Q as Qh*10^17 + Ql
			//T128 = bid_reciprocals10_128[17];
			T128.lo = 0x44909befeb9fad49
			T128.hi = 0x000b877aa3236a4b
			P256 = __mul_128x128_to_256(CQ, T128)
			//amount = bid_recip_scale[17];
			Q_high = (P256.w2 >> 44) | (P256.w3 << (64 - 44))
			Q_low = CQ.lo - Q_high*100000000000000000

			if Q_low == 0 {
				diff_expon += 17

				tdigit[0] = uint32(Q_high) & 0x3ffffff
				tdigit[1] = 0
				QX = Q_high >> 26
				QX32 = uint32(QX)
				nzeros = 0

				for j = 0; QX32 != 0; j, QX32 = j+1, QX32>>7 {
					k = int(QX32 & 127)
					tdigit[0] += bid_convert_table[j][k][0]
					tdigit[1] += bid_convert_table[j][k][1]
					if tdigit[0] >= 100000000 {
						tdigit[0] -= 100000000
						tdigit[1]++
					}
				}

				if tdigit[1] >= 100000000 {
					tdigit[1] -= 100000000
					if tdigit[1] >= 100000000 {
						tdigit[1] -= 100000000
					}
				}

				digit = tdigit[0]
				if digit == 0 && tdigit[1] == 0 {
					nzeros += 16
				} else {
					if digit == 0 {
						nzeros += 8
						digit = tdigit[1]
					}
					// decompose digit
					PD = uint64(digit) * 0x068DB8BB
					digit_h = uint32(PD >> 40)
					digit_low = digit - digit_h*10000

					if digit_low == 0 {
						nzeros += 4
					} else {
						digit_h = digit_low
					}

					if (digit_h & 1) == 0 {
						nzeros +=
							int(3 & uint32(bid_packed_10000_zeros[digit_h>>3]>>
								(digit_h&7)))
					}
				}

				if nzeros != 0 {
					CQ = __mul_64x64_to_128(Q_high, bid_reciprocals10_64[nzeros])

					// now get P/10^extra_digits: shift C64 right by M[extra_digits]-64
					amount = bid_short_recip_scale[nzeros]
					CQ.lo = CQ.hi >> uint(amount)
				} else {
					CQ.lo = Q_high
				}
				CQ.hi = 0

				diff_expon += nzeros
			} else {
				tdigit[0] = uint32(Q_low) & 0x3ffffff
				tdigit[1] = 0
				QX = Q_low >> 26
				QX32 = uint32(QX)
				nzeros = 0

				for j = 0; QX32 != 0; j, QX32 = j+1, QX32>>7 {
					k = int(QX32 & 127)
					tdigit[0] += bid_convert_table[j][k][0]
					tdigit[1] += bid_convert_table[j][k][1]
					if tdigit[0] >= 100000000 {
						tdigit[0] -= 100000000
						tdigit[1]++
					}
				}

				if tdigit[1] >= 100000000 {
					tdigit[1] -= 100000000
					if tdigit[1] >= 100000000 {
						tdigit[1] -= 100000000
					}
				}

				digit = tdigit[0]
				if digit == 0 && tdigit[1] == 0 {
					nzeros += 16
				} else {
					if digit == 0 {
						nzeros += 8
						digit = tdigit[1]
					}
					// decompose digit
					PD = uint64(digit) * 0x068DB8BB
					digit_h = uint32(PD >> 40)
					digit_low = digit - digit_h*10000

					if digit_low == 0 {
						nzeros += 4
					} else {
						digit_h = digit_low
					}

					if (digit_h & 1) == 0 {
						nzeros +=
							int(3 & uint32(bid_packed_10000_zeros[digit_h>>3]>>
								(digit_h&7)))
					}
				}

				if nzeros != 0 {
					// get P*(2^M[extra_digits])/10^extra_digits
					Qh, _ = __mul_128x128_full(CQ, bid_reciprocals10_128[nzeros])

					//now get P/10^extra_digits: shift Q_high right by M[extra_digits]-128
					amount = bid_recip_scale[nzeros]
					CQ = __shr_128(Qh, uint(amount))
				}
				diff_expon += nzeros

			}
		}
		res = bid_get_BID128(sign_x^sign_y, diff_expon, CQ, rnd_mode, &pfpsf)
		return res, pfpsf
	}

	if diff_expon >= 0 {
		rmode = uint(rnd_mode)
		if (sign_x^sign_y) != 0 && uint(rmode-1) < 2 {
			rmode = 3 - rmode
		}
		switch rmode {
		case BID_ROUNDING_TO_NEAREST: // round to nearest code
			// rounding
			// 2*CA4 - CY
			CA4r.w1 = (CA4.w1 + CA4.w1) | (CA4.w0 >> 63)
			CA4r.w0 = CA4.w0 + CA4.w0
			CA4r.w0, carry64 = __sub_borrow_out(CA4r.w0, CY.lo)
			CA4r.w1 = CA4r.w1 - CY.hi - carry64
			if (CA4r.w1 | CA4r.w0) != 0 {
				D = 1
			} else {
				D = 0
			}
			carry64 = uint64(1+int64(CA4r.w1)>>63) & ((CQ.lo) | D)
			CQ.lo += carry64
			if CQ.lo < carry64 {
				CQ.hi++
			}
		case BID_ROUNDING_TIES_AWAY:
			// rounding
			// 2*CA4 - CY
			CA4r.w1 = (CA4.w1 + CA4.w1) | (CA4.w0 >> 63)
			CA4r.w0 = CA4.w0 + CA4.w0
			CA4r.w0, carry64 = __sub_borrow_out(CA4r.w0, CY.lo)
			CA4r.w1 = CA4r.w1 - CY.hi - carry64
			if (CA4r.w1 | CA4r.w0) != 0 {
				D = 0
			} else {
				D = 1
			}
			carry64 = uint64(1+int64(CA4r.w1)>>63) | D
			CQ.lo += carry64
			if CQ.lo < carry64 {
				CQ.hi++
			}
		case BID_ROUNDING_DOWN, BID_ROUNDING_TO_ZERO:
			// do nothing
		default: // rounding up
			CQ.lo++
			if CQ.lo == 0 {
				CQ.hi++
			}
		}

	} else {
		if CA4.w0 != 0 || CA4.w1 != 0 {
			// set status flags
			pfpsf |= BID_INEXACT_EXCEPTION
		}

		res = bid_handle_UF_128_rem(sign_x^sign_y, diff_expon, CQ,
			CA4.w1|CA4.w0, rnd_mode, &pfpsf)
		return res, pfpsf
	}

	res = bid_get_BID128(sign_x^sign_y, diff_expon, CQ, rnd_mode, &pfpsf)
	return res, pfpsf
}
