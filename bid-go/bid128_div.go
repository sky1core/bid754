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
	{w: [2]uint64{0x000000000000000a, 0x0000000000000000}}, // 0
	{w: [2]uint64{0x000000000000000a, 0x0000000000000000}}, // 1
	{w: [2]uint64{0x000000000000000a, 0x0000000000000000}}, // 2
	{w: [2]uint64{0x000000000000000a, 0x0000000000000000}}, // 3
	{w: [2]uint64{0x0000000000000064, 0x0000000000000000}}, // 4
	{w: [2]uint64{0x0000000000000064, 0x0000000000000000}}, // 5
	{w: [2]uint64{0x0000000000000064, 0x0000000000000000}}, // 6
	{w: [2]uint64{0x00000000000003e8, 0x0000000000000000}}, // 7
	{w: [2]uint64{0x00000000000003e8, 0x0000000000000000}}, // 8
	{w: [2]uint64{0x00000000000003e8, 0x0000000000000000}}, // 9
	{w: [2]uint64{0x0000000000002710, 0x0000000000000000}}, // 10
	{w: [2]uint64{0x0000000000002710, 0x0000000000000000}}, // 11
	{w: [2]uint64{0x0000000000002710, 0x0000000000000000}}, // 12
	{w: [2]uint64{0x0000000000002710, 0x0000000000000000}}, // 13
	{w: [2]uint64{0x00000000000186a0, 0x0000000000000000}}, // 14
	{w: [2]uint64{0x00000000000186a0, 0x0000000000000000}}, // 15
	{w: [2]uint64{0x00000000000186a0, 0x0000000000000000}}, // 16
	{w: [2]uint64{0x00000000000f4240, 0x0000000000000000}}, // 17
	{w: [2]uint64{0x00000000000f4240, 0x0000000000000000}}, // 18
	{w: [2]uint64{0x00000000000f4240, 0x0000000000000000}}, // 19
	{w: [2]uint64{0x0000000000989680, 0x0000000000000000}}, // 20
	{w: [2]uint64{0x0000000000989680, 0x0000000000000000}}, // 21
	{w: [2]uint64{0x0000000000989680, 0x0000000000000000}}, // 22
	{w: [2]uint64{0x0000000000989680, 0x0000000000000000}}, // 23
	{w: [2]uint64{0x0000000005f5e100, 0x0000000000000000}}, // 24
	{w: [2]uint64{0x0000000005f5e100, 0x0000000000000000}}, // 25
	{w: [2]uint64{0x0000000005f5e100, 0x0000000000000000}}, // 26
	{w: [2]uint64{0x000000003b9aca00, 0x0000000000000000}}, // 27
	{w: [2]uint64{0x000000003b9aca00, 0x0000000000000000}}, // 28
	{w: [2]uint64{0x000000003b9aca00, 0x0000000000000000}}, // 29
	{w: [2]uint64{0x00000002540be400, 0x0000000000000000}}, // 30
	{w: [2]uint64{0x00000002540be400, 0x0000000000000000}}, // 31
	{w: [2]uint64{0x00000002540be400, 0x0000000000000000}}, // 32
	{w: [2]uint64{0x00000002540be400, 0x0000000000000000}}, // 33
	{w: [2]uint64{0x000000174876e800, 0x0000000000000000}}, // 34
	{w: [2]uint64{0x000000174876e800, 0x0000000000000000}}, // 35
	{w: [2]uint64{0x000000174876e800, 0x0000000000000000}}, // 36
	{w: [2]uint64{0x000000e8d4a51000, 0x0000000000000000}}, // 37
	{w: [2]uint64{0x000000e8d4a51000, 0x0000000000000000}}, // 38
	{w: [2]uint64{0x000000e8d4a51000, 0x0000000000000000}}, // 39
	{w: [2]uint64{0x000009184e72a000, 0x0000000000000000}}, // 40
	{w: [2]uint64{0x000009184e72a000, 0x0000000000000000}}, // 41
	{w: [2]uint64{0x000009184e72a000, 0x0000000000000000}}, // 42
	{w: [2]uint64{0x000009184e72a000, 0x0000000000000000}}, // 43
	{w: [2]uint64{0x00005af3107a4000, 0x0000000000000000}}, // 44
	{w: [2]uint64{0x00005af3107a4000, 0x0000000000000000}}, // 45
	{w: [2]uint64{0x00005af3107a4000, 0x0000000000000000}}, // 46
	{w: [2]uint64{0x00038d7ea4c68000, 0x0000000000000000}}, // 47
	{w: [2]uint64{0x00038d7ea4c68000, 0x0000000000000000}}, // 48
	{w: [2]uint64{0x00038d7ea4c68000, 0x0000000000000000}}, // 49
	{w: [2]uint64{0x002386f26fc10000, 0x0000000000000000}}, // 50
	{w: [2]uint64{0x002386f26fc10000, 0x0000000000000000}}, // 51
	{w: [2]uint64{0x002386f26fc10000, 0x0000000000000000}}, // 52
	{w: [2]uint64{0x002386f26fc10000, 0x0000000000000000}}, // 53
	{w: [2]uint64{0x016345785d8a0000, 0x0000000000000000}}, // 54
	{w: [2]uint64{0x016345785d8a0000, 0x0000000000000000}}, // 55
	{w: [2]uint64{0x016345785d8a0000, 0x0000000000000000}}, // 56
	{w: [2]uint64{0x0de0b6b3a7640000, 0x0000000000000000}}, // 57
	{w: [2]uint64{0x0de0b6b3a7640000, 0x0000000000000000}}, // 58
	{w: [2]uint64{0x0de0b6b3a7640000, 0x0000000000000000}}, // 59
	{w: [2]uint64{0x8ac7230489e80000, 0x0000000000000000}}, // 60
	{w: [2]uint64{0x8ac7230489e80000, 0x0000000000000000}}, // 61
	{w: [2]uint64{0x8ac7230489e80000, 0x0000000000000000}}, // 62
	{w: [2]uint64{0x8ac7230489e80000, 0x0000000000000000}}, // 63
	{w: [2]uint64{0x6bc75e2d63100000, 0x0000000000000005}}, // 64: 10^20
	{w: [2]uint64{0x6bc75e2d63100000, 0x0000000000000005}}, // 65
	{w: [2]uint64{0x6bc75e2d63100000, 0x0000000000000005}}, // 66
	{w: [2]uint64{0x35c9adc5dea00000, 0x0000000000000036}}, // 67: 10^21
	{w: [2]uint64{0x35c9adc5dea00000, 0x0000000000000036}}, // 68
	{w: [2]uint64{0x35c9adc5dea00000, 0x0000000000000036}}, // 69
	{w: [2]uint64{0x19e0c9bab2400000, 0x000000000000021e}}, // 70: 10^22
	{w: [2]uint64{0x19e0c9bab2400000, 0x000000000000021e}}, // 71
	{w: [2]uint64{0x19e0c9bab2400000, 0x000000000000021e}}, // 72
	{w: [2]uint64{0x19e0c9bab2400000, 0x000000000000021e}}, // 73
	{w: [2]uint64{0x02c7e14af6800000, 0x000000000000152d}}, // 74: 10^23
	{w: [2]uint64{0x02c7e14af6800000, 0x000000000000152d}}, // 75
	{w: [2]uint64{0x02c7e14af6800000, 0x000000000000152d}}, // 76
	{w: [2]uint64{0x1bcecceda1000000, 0x000000000000d3c2}}, // 77: 10^24
	{w: [2]uint64{0x1bcecceda1000000, 0x000000000000d3c2}}, // 78
	{w: [2]uint64{0x1bcecceda1000000, 0x000000000000d3c2}}, // 79
	{w: [2]uint64{0x161401484a000000, 0x0000000000084595}}, // 80: 10^25
	{w: [2]uint64{0x161401484a000000, 0x0000000000084595}}, // 81
	{w: [2]uint64{0x161401484a000000, 0x0000000000084595}}, // 82
	{w: [2]uint64{0x161401484a000000, 0x0000000000084595}}, // 83
	{w: [2]uint64{0xdcc80cd2e4000000, 0x000000000052b7d2}}, // 84: 10^26
	{w: [2]uint64{0xdcc80cd2e4000000, 0x000000000052b7d2}}, // 85
	{w: [2]uint64{0xdcc80cd2e4000000, 0x000000000052b7d2}}, // 86
	{w: [2]uint64{0x9fd0803ce8000000, 0x00000000033b2e3c}}, // 87: 10^27
	{w: [2]uint64{0x9fd0803ce8000000, 0x00000000033b2e3c}}, // 88
	{w: [2]uint64{0x9fd0803ce8000000, 0x00000000033b2e3c}}, // 89
	{w: [2]uint64{0x3e25026110000000, 0x00000000204fce5e}}, // 90: 10^28
	{w: [2]uint64{0x3e25026110000000, 0x00000000204fce5e}}, // 91
	{w: [2]uint64{0x3e25026110000000, 0x00000000204fce5e}}, // 92
	{w: [2]uint64{0x3e25026110000000, 0x00000000204fce5e}}, // 93
	{w: [2]uint64{0x6d7217caa0000000, 0x00000001431e0fae}}, // 94: 10^29
	{w: [2]uint64{0x6d7217caa0000000, 0x00000001431e0fae}}, // 95
	{w: [2]uint64{0x6d7217caa0000000, 0x00000001431e0fae}}, // 96
	{w: [2]uint64{0x4674edea40000000, 0x0000000c9f2c9cd0}}, // 97: 10^30
	{w: [2]uint64{0x4674edea40000000, 0x0000000c9f2c9cd0}}, // 98
	{w: [2]uint64{0x4674edea40000000, 0x0000000c9f2c9cd0}}, // 99
	{w: [2]uint64{0xc0914b2680000000, 0x0000007e37be2022}}, // 100: 10^31
	{w: [2]uint64{0xc0914b2680000000, 0x0000007e37be2022}}, // 101
	{w: [2]uint64{0xc0914b2680000000, 0x0000007e37be2022}}, // 102
	{w: [2]uint64{0x85acef8100000000, 0x000004ee2d6d415b}}, // 103: 10^32
	{w: [2]uint64{0x85acef8100000000, 0x000004ee2d6d415b}}, // 104
	{w: [2]uint64{0x85acef8100000000, 0x000004ee2d6d415b}}, // 105
	{w: [2]uint64{0x85acef8100000000, 0x000004ee2d6d415b}}, // 106
	{w: [2]uint64{0x38c15b0a00000000, 0x0000314dc6448d93}}, // 107: 10^33
	{w: [2]uint64{0x38c15b0a00000000, 0x0000314dc6448d93}}, // 108
	{w: [2]uint64{0x38c15b0a00000000, 0x0000314dc6448d93}}, // 109: entry 112 in C
	{w: [2]uint64{0x378d8e6400000000, 0x0001ed09bead87c0}}, // 110: 10^34
	{w: [2]uint64{0x378d8e6400000000, 0x0001ed09bead87c0}}, // 111
	{w: [2]uint64{0x378d8e6400000000, 0x0001ed09bead87c0}}, // 112
	{w: [2]uint64{0x2b878fe800000000, 0x0013426172c74d82}}, // 113: 10^35
	{w: [2]uint64{0x2b878fe800000000, 0x0013426172c74d82}}, // 114
	{w: [2]uint64{0x2b878fe800000000, 0x0013426172c74d82}}, // 115
	{w: [2]uint64{0x2b878fe800000000, 0x0013426172c74d82}}, // 116
	{w: [2]uint64{0xb34b9f1000000000, 0x00c097ce7bc90715}}, // 117: 10^36
	{w: [2]uint64{0x00f436a000000000, 0x0785ee10d5da46d9}}, // 118: 10^37
	{w: [2]uint64{0x00f436a000000000, 0x0785ee10d5da46d9}}, // 119
	{w: [2]uint64{0x00f436a000000000, 0x0785ee10d5da46d9}}, // 120
	{w: [2]uint64{0x098a224000000000, 0x4b3b4ca85a86c47a}}, // 121: 10^38
	{w: [2]uint64{0x098a224000000000, 0x4b3b4ca85a86c47a}}, // 122
	{w: [2]uint64{0x098a224000000000, 0x4b3b4ca85a86c47a}}, // 123
	{w: [2]uint64{0x098a224000000000, 0x4b3b4ca85a86c47a}}, // 124
}

// __mul_128x128_low multiplies two 128-bit values, returning only the low 128 bits.
// Ported from bid_internal.h
func __mul_128x128_low(A, B BID_UINT128) BID_UINT128 {
	var Ql BID_UINT128
	ALBL := __mul_64x64_to_128(A.w[0], B.w[0])
	QM64 := B.w[0]*A.w[1] + A.w[0]*B.w[1]

	Ql.w[0] = ALBL.w[0]
	Ql.w[1] = QM64 + ALBL.w[1]
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

	if CX0.w[1] == 0 && CY.w[1] == 0 {
		CQ.w[0] = CX0.w[0] / CY.w[0]
		CQ.w[1] = 0
		CR.w[1] = 0
		CR.w[0] = CX0.w[0] - CQ.w[0]*CY.w[0]
		return
	}

	CX := CX0

	// 2^64
	t64 := math.Float64frombits(0x43f0000000000000)
	lx := noFmaMulAddF64(float64(CX.w[1]), t64, float64(CX.w[0]))
	ly := noFmaMulAddF64(float64(CY.w[1]), t64, float64(CY.w[0]))
	lq := lx / ly

	CY36.w[1] = CY.w[0] >> (64 - 36)
	CY36.w[0] = CY.w[0] << 36

	CQ.w[1] = 0
	CQ.w[0] = 0

	// Q >= 2^100 ?
	if CY.w[1] == 0 && CY36.w[1] == 0 && (CX.w[1] >= CY36.w[0]) {
		// then Q >= 2^100

		// 2^(-60)*CX/CY
		d60 := math.Float64frombits(0x3c30000000000000)
		lq *= d60
		Q = uint64(lq) - 4

		// Q*CY
		A2 = __mul_64x64_to_128(Q, CY.w[0])

		// A2 <<= 60
		A2.w[1] = (A2.w[1] << 60) | (A2.w[0] >> (64 - 60))
		A2.w[0] <<= 60

		CX = __sub_128_128(CX, A2)

		lx = noFmaMulAddF64(float64(CX.w[1]), t64, float64(CX.w[0]))
		lq = lx / ly

		CQ.w[1] = Q >> (64 - 60)
		CQ.w[0] = Q << 60
	}

	CY51.w[1] = (CY.w[1] << 51) | (CY.w[0] >> (64 - 51))
	CY51.w[0] = CY.w[0] << 51

	if CY.w[1] < (1<<(64-51)) && __unsigned_compare_gt_128(CX, CY51) {
		// Q > 2^51

		// 2^(-49)*CX/CY
		d49 := math.Float64frombits(0x3ce0000000000000)
		lq *= d49

		Q = uint64(lq) - 1

		// Q*CY
		A2 = __mul_64x64_to_128(Q, CY.w[0])
		A2.w[1] += Q * CY.w[1]

		// A2 <<= 49
		A2.w[1] = (A2.w[1] << 49) | (A2.w[0] >> (64 - 49))
		A2.w[0] <<= 49

		CX = __sub_128_128(CX, A2)

		CQT.w[1] = Q >> (64 - 49)
		CQT.w[0] = Q << 49
		CQ = __add_128_128(CQ, CQT)

		lx = noFmaMulAddF64(float64(CX.w[1]), t64, float64(CX.w[0]))
		lq = lx / ly
	}

	Q = uint64(lq)

	A2 = __mul_64x64_to_128(Q, CY.w[0])
	A2.w[1] += Q * CY.w[1]

	CX = __sub_128_128(CX, A2)
	if int64(CX.w[1]) < 0 {
		Q--
		CX.w[0] += CY.w[0]
		if CX.w[0] < CY.w[0] {
			CX.w[1]++
		}
		CX.w[1] += CY.w[1]
		if int64(CX.w[1]) < 0 {
			Q--
			CX.w[0] += CY.w[0]
			if CX.w[0] < CY.w[0] {
				CX.w[1]++
			}
			CX.w[1] += CY.w[1]
		}
	} else if __unsigned_compare_ge_128(CX, CY) {
		Q++
		CX = __sub_128_128(CX, CY)
	}

	CQ = __add_128_64(CQ, Q)

	CR.w[1] = CX.w[1]
	CR.w[0] = CX.w[0]
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
	CA4.w[3] = pCA4.w[3]
	CA4.w[2] = pCA4.w[2]
	CA4.w[1] = pCA4.w[1]
	CA4.w[0] = pCA4.w[0]
	CQ.w[1] = pCQ.w[1]
	CQ.w[0] = pCQ.w[0]

	// 2^64
	t64 := math.Float64frombits(0x43f0000000000000)
	d128 := t64 * t64
	d192 := d128 * t64
	lx := noFmaMulAddF64(float64(CA4.w[3]), d192,
		noFmaMulAddF64(float64(CA4.w[2]), d128,
			noFmaMulAddF64(float64(CA4.w[1]), t64, float64(CA4.w[0]))))
	ly := noFmaMulAddF64(float64(CY.w[1]), t64, float64(CY.w[0]))
	lq := lx / ly

	var CY36_2, CY36_1, CY36_0 uint64
	CY36_2 = CY.w[1] >> (64 - 36)
	CY36_1 = (CY.w[1] << 36) | (CY.w[0] >> (64 - 36))
	CY36_0 = CY.w[0] << 36

	// Q >= 2^100 ?
	if CA4.w[3] > CY36_2 ||
		(CA4.w[3] == CY36_2 &&
			(CA4.w[2] > CY36_1 ||
				(CA4.w[2] == CY36_1 && CA4.w[1] >= CY36_0))) {
		// 2^(-60)*CA4/CY
		d60 := math.Float64frombits(0x3c30000000000000)
		lq *= d60
		Q = uint64(lq) - 4

		// Q*CY
		tmp192 := __mul_64x128_to_192(Q, CY)

		// CA2 <<= 60
		CA2[2] = (tmp192.w[2] << 60) | (tmp192.w[1] >> (64 - 60))
		CA2[1] = (tmp192.w[1] << 60) | (tmp192.w[0] >> (64 - 60))
		CA2[0] = tmp192.w[0] << 60

		// CA4 -= CA2
		CA4.w[0], carry64 = __sub_borrow_out(CA4.w[0], CA2[0])
		CA4.w[1], carry64 = __sub_borrow_in_out(CA4.w[1], CA2[1], carry64)
		CA4.w[2] = CA4.w[2] - CA2[2] - carry64

		lx = noFmaMulAddF64(float64(CA4.w[2]), d128,
			noFmaMulAddF64(float64(CA4.w[1]), t64, float64(CA4.w[0])))
		lq = lx / ly

		CQT.w[1] = Q >> (64 - 60)
		CQT.w[0] = Q << 60
		CQ = __add_128_128(CQ, CQT)
	}

	var CY51_2, CY51_1, CY51_0 uint64
	CY51_2 = CY.w[1] >> (64 - 51)
	CY51_1 = (CY.w[1] << 51) | (CY.w[0] >> (64 - 51))
	CY51_0 = CY.w[0] << 51

	// compare CA4 with CY51 (as 192-bit values)
	ca4_128 := BID_UINT128{w: [2]uint64{CA4.w[0], CA4.w[1]}}
	cy51_128 := BID_UINT128{w: [2]uint64{CY51_0, CY51_1}}

	if CA4.w[2] > CY51_2 || ((CA4.w[2] == CY51_2) &&
		__unsigned_compare_gt_128(ca4_128, cy51_128)) {
		// Q > 2^51

		// 2^(-49)*CA4/CY
		d49 := math.Float64frombits(0x3ce0000000000000)
		lq *= d49

		Q = uint64(lq) - 1

		// Q*CY
		A2 = __mul_64x64_to_128(Q, CY.w[0])
		A2h = __mul_64x64_to_128(Q, CY.w[1])
		A2.w[1] += A2h.w[0]
		if A2.w[1] < A2h.w[0] {
			A2h.w[1]++
		}

		// A2 <<= 49
		CA2[2] = (A2h.w[1] << 49) | (A2.w[1] >> (64 - 49))
		CA2[1] = (A2.w[1] << 49) | (A2.w[0] >> (64 - 49))
		CA2[0] = A2.w[0] << 49

		CA4.w[0], carry64 = __sub_borrow_out(CA4.w[0], CA2[0])
		CA4.w[1], carry64 = __sub_borrow_in_out(CA4.w[1], CA2[1], carry64)
		CA4.w[2] = CA4.w[2] - CA2[2] - carry64

		CQT.w[1] = Q >> (64 - 49)
		CQT.w[0] = Q << 49
		CQ = __add_128_128(CQ, CQT)

		lx = noFmaMulAddF64(float64(CA4.w[2]), d128,
			noFmaMulAddF64(float64(CA4.w[1]), t64, float64(CA4.w[0])))
		lq = lx / ly
	}

	Q = uint64(lq)
	A2 = __mul_64x64_to_128(Q, CY.w[0])
	A2.w[1] += Q * CY.w[1]

	// __sub_128_128(CA4, CA4, A2) - using CA4.w[0:1] as 128-bit
	tmpCA := BID_UINT128{w: [2]uint64{CA4.w[0], CA4.w[1]}}
	tmpCA = __sub_128_128(tmpCA, A2)
	CA4.w[0] = tmpCA.w[0]
	CA4.w[1] = tmpCA.w[1]

	if int64(CA4.w[1]) < 0 {
		Q--
		CA4.w[0] += CY.w[0]
		if CA4.w[0] < CY.w[0] {
			CA4.w[1]++
		}
		CA4.w[1] += CY.w[1]
		if int64(CA4.w[1]) < 0 {
			Q--
			CA4.w[0] += CY.w[0]
			if CA4.w[0] < CY.w[0] {
				CA4.w[1]++
			}
			CA4.w[1] += CY.w[1]
		}
	} else if CA4.w[1] > CY.w[1] || (CA4.w[1] == CY.w[1] && CA4.w[0] >= CY.w[0]) {
		Q++
		tmpCA2 := BID_UINT128{w: [2]uint64{CA4.w[0], CA4.w[1]}}
		tmpCA2 = __sub_128_128(tmpCA2, CY)
		CA4.w[0] = tmpCA2.w[0]
		CA4.w[1] = tmpCA2.w[1]
	}

	CQ = __add_128_64(CQ, Q)

	pCQ.w[1] = CQ.w[1]
	pCQ.w[0] = CQ.w[0]
	pCA4.w[1] = CA4.w[1]
	pCA4.w[0] = CA4.w[0]
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
		res.w[1] = sgn
		res.w[0] = 0
		if (sgn != 0 && prounding_mode == BID_ROUNDING_DOWN) ||
			(sgn == 0 && prounding_mode == BID_ROUNDING_UP) {
			res.w[0] = 1
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
	CQ.w[0], carry = __add_carry_out(T128.w[0], CQ.w[0])
	CQ.w[1] = CQ.w[1] + T128.w[1] + carry

	TP128 = bid_reciprocals10_128[ed2]
	Qh, Ql = __mul_128x128_full(CQ, TP128)
	amount = bid_recip_scale[ed2]

	if amount >= 64 {
		CQ.w[0] = Qh.w[1] >> uint(amount-64)
		CQ.w[1] = 0
	} else {
		CQ = __shr_128(Qh, uint(amount))
	}

	expon = 0

	if prounding_mode == BID_ROUNDING_TO_NEAREST {
		if CQ.w[0]&1 != 0 {
			// check whether fractional part of initial_P/10^ed1 is exactly .5

			// get remainder
			Qh1 = __shl_128_long(Qh, uint(128-amount))

			if Qh1.w[1] == 0 && Qh1.w[0] == 0 &&
				(Ql.w[1] < bid_reciprocals10_128[ed2].w[1] ||
					(Ql.w[1] == bid_reciprocals10_128[ed2].w[1] &&
						Ql.w[0] < bid_reciprocals10_128[ed2].w[0])) {
				CQ.w[0]--
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
			if Qh1.w[1] == 0x8000000000000000 && Qh1.w[0] == 0 &&
				(Ql.w[1] < bid_reciprocals10_128[ed2].w[1] ||
					(Ql.w[1] == bid_reciprocals10_128[ed2].w[1] &&
						Ql.w[0] < bid_reciprocals10_128[ed2].w[0])) {
				status = BID_EXACT_STATUS
			}
		case BID_ROUNDING_DOWN, BID_ROUNDING_TO_ZERO:
			if Qh1.w[1] == 0 && Qh1.w[0] == 0 &&
				(Ql.w[1] < bid_reciprocals10_128[ed2].w[1] ||
					(Ql.w[1] == bid_reciprocals10_128[ed2].w[1] &&
						Ql.w[0] < bid_reciprocals10_128[ed2].w[0])) {
				status = BID_EXACT_STATUS
			}
		default:
			// round up
			Stemp.w[0], CY = __add_carry_out(Ql.w[0], bid_reciprocals10_128[ed2].w[0])
			Stemp.w[1], carry = __add_carry_in_out(Ql.w[1], bid_reciprocals10_128[ed2].w[1], CY)
			_ = Stemp
			Qh = __shr_128_long(Qh1, uint(128-amount))
			Tmp.w[0] = 1
			Tmp.w[1] = 0
			Tmp1 = __shl_128_long(Tmp, uint(amount))
			Qh.w[0] += carry
			if Qh.w[0] < carry {
				Qh.w[1]++
			}
			if __unsigned_compare_ge_128(Qh, Tmp1) {
				status = BID_EXACT_STATUS
			}
		}

		if status != BID_EXACT_STATUS {
			*fpsc |= BID_UNDERFLOW_EXCEPTION | status
		}
	}

	res.w[1] = sgn | CQ.w[1]
	res.w[0] = CQ.w[0]

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
		res.w[1] = sgn
		res.w[0] = 0
		if (sgn != 0 && prounding_mode == BID_ROUNDING_DOWN) ||
			(sgn == 0 && prounding_mode == BID_ROUNDING_UP) {
			res.w[0] = 1
		}
		return res
	}
	// CQ *= 10
	CQ2.w[1] = (CQ.w[1] << 1) | (CQ.w[0] >> 63)
	CQ2.w[0] = CQ.w[0] << 1
	CQ8.w[1] = (CQ.w[1] << 3) | (CQ.w[0] >> 61)
	CQ8.w[0] = CQ.w[0] << 3
	CQ = __add_128_128(CQ2, CQ8)

	// add remainder
	if R != 0 {
		CQ.w[0] |= 1
	}

	ed2 = 1 - expon
	// add rounding constant to CQ
	rmode = uint(prounding_mode)
	if sgn != 0 && uint(rmode-1) < 2 {
		rmode = 3 - rmode
	}
	T128 = bid_round_const_table_128[rmode][ed2]
	CQ.w[0], carry = __add_carry_out(T128.w[0], CQ.w[0])
	CQ.w[1] = CQ.w[1] + T128.w[1] + carry

	TP128 = bid_reciprocals10_128[ed2]
	Qh, Ql = __mul_128x128_full(CQ, TP128)
	amount = bid_recip_scale[ed2]

	if amount >= 64 {
		CQ.w[0] = Qh.w[1] >> uint(amount-64)
		CQ.w[1] = 0
	} else {
		CQ = __shr_128(Qh, uint(amount))
	}

	expon = 0

	if prounding_mode == BID_ROUNDING_TO_NEAREST {
		if CQ.w[0]&1 != 0 {
			// check whether fractional part of initial_P/10^ed1 is exactly .5

			// get remainder
			Qh1 = __shl_128_long(Qh, uint(128-amount))

			if Qh1.w[1] == 0 && Qh1.w[0] == 0 &&
				(Ql.w[1] < bid_reciprocals10_128[ed2].w[1] ||
					(Ql.w[1] == bid_reciprocals10_128[ed2].w[1] &&
						Ql.w[0] < bid_reciprocals10_128[ed2].w[0])) {
				CQ.w[0]--
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
			if Qh1.w[1] == 0x8000000000000000 && Qh1.w[0] == 0 &&
				(Ql.w[1] < bid_reciprocals10_128[ed2].w[1] ||
					(Ql.w[1] == bid_reciprocals10_128[ed2].w[1] &&
						Ql.w[0] < bid_reciprocals10_128[ed2].w[0])) {
				status = BID_EXACT_STATUS
			}
		case BID_ROUNDING_DOWN, BID_ROUNDING_TO_ZERO:
			if Qh1.w[1] == 0 && Qh1.w[0] == 0 &&
				(Ql.w[1] < bid_reciprocals10_128[ed2].w[1] ||
					(Ql.w[1] == bid_reciprocals10_128[ed2].w[1] &&
						Ql.w[0] < bid_reciprocals10_128[ed2].w[0])) {
				status = BID_EXACT_STATUS
			}
		default:
			// round up
			Stemp.w[0], CY = __add_carry_out(Ql.w[0], bid_reciprocals10_128[ed2].w[0])
			Stemp.w[1], carry = __add_carry_in_out(Ql.w[1], bid_reciprocals10_128[ed2].w[1], CY)
			_ = Stemp
			Qh = __shr_128_long(Qh1, uint(128-amount))
			Tmp.w[0] = 1
			Tmp.w[1] = 0
			Tmp1 = __shl_128_long(Tmp, uint(amount))
			Qh.w[0] += carry
			if Qh.w[0] < carry {
				Qh.w[1]++
			}
			if __unsigned_compare_ge_128(Qh, Tmp1) {
				status = BID_EXACT_STATUS
			}
		}

		if status != BID_EXACT_STATUS {
			*fpsc |= BID_UNDERFLOW_EXCEPTION | status
		}
	}

	res.w[1] = sgn | CQ.w[1]
	res.w[0] = CQ.w[0]

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
	if coeff.w[1] == 0x0001ed09bead87c0 && coeff.w[0] == 0x378d8e6400000000 {
		expon++
		// set coefficient to 10^33
		coeff.w[1] = 0x0000314dc6448d93
		coeff.w[0] = 0x38c15b0a00000000
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
				coeff.w[1] =
					(coeff.w[1] << 3) + (coeff.w[1] << 1) + (coeff.w[0] >> 61) +
						(coeff.w[0] >> 63)
				tmp2 = coeff.w[0] << 3
				coeff.w[0] = (coeff.w[0] << 1) + tmp2
				if coeff.w[0] < tmp2 {
					coeff.w[1]++
				}

				expon--
			}
		}
		if expon > DECIMAL_MAX_EXPON_128 {
			if coeff.w[1] == 0 && coeff.w[0] == 0 {
				res.w[1] = sgn | (uint64(DECIMAL_MAX_EXPON_128) << 49)
				res.w[0] = 0
				return res
			}
			// OF
			*fpsc |= BID_OVERFLOW_EXCEPTION | BID_INEXACT_EXCEPTION
			if prounding_mode == BID_ROUNDING_TO_ZERO ||
				(sgn != 0 && prounding_mode == BID_ROUNDING_UP) ||
				(sgn == 0 && prounding_mode == BID_ROUNDING_DOWN) {
				res.w[1] = sgn | LARGEST_BID128_HIGH
				res.w[0] = LARGEST_BID128_LOW
			} else {
				res.w[1] = sgn | INFINITY_MASK64
				res.w[0] = 0
			}
			return res
		}
	}

	res.w[0] = coeff.w[0]
	tmp = uint64(expon)
	tmp <<= 49
	res.w[1] = sgn | tmp | coeff.w[1]

	return res
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
		if (x.w[1] & 0x7c00000000000000) == 0x7c00000000000000 {
			if (x.w[1]&0x7e00000000000000) == 0x7e00000000000000 || // sNaN
				(y.w[1]&0x7e00000000000000) == 0x7e00000000000000 {
				pfpsf |= BID_INVALID_EXCEPTION
			}
			res.w[1] = (CX.w[1]) & QUIET_MASK64
			res.w[0] = CX.w[0]
			return res, pfpsf
		}
		// x is Infinity?
		if (x.w[1] & 0x7800000000000000) == 0x7800000000000000 {
			// check if y is Inf.
			if (y.w[1] & 0x7c00000000000000) == 0x7800000000000000 {
				// return NaN
				pfpsf |= BID_INVALID_EXCEPTION
				res.w[1] = 0x7c00000000000000
				res.w[0] = 0
				return res, pfpsf
			}
			// y is NaN?
			if (y.w[1] & 0x7c00000000000000) != 0x7c00000000000000 {
				// return +/-Inf
				res.w[1] = ((x.w[1] ^ y.w[1]) & 0x8000000000000000) |
					0x7800000000000000
				res.w[0] = 0
				return res, pfpsf
			}
		}
		// x is 0
		if (y.w[1] & 0x7800000000000000) < 0x7800000000000000 {
			if CY.w[0] == 0 && (CY.w[1]&0x0001ffffffffffff) == 0 {
				pfpsf |= BID_INVALID_EXCEPTION
				// x=y=0, return NaN
				res.w[1] = 0x7c00000000000000
				res.w[0] = 0
				return res, pfpsf
			}
			// return 0
			res.w[1] = (x.w[1] ^ y.w[1]) & 0x8000000000000000
			exponent_x = exponent_x - exponent_y + EXPONENT_BIAS128
			if exponent_x > DECIMAL_MAX_EXPON_128 {
				exponent_x = DECIMAL_MAX_EXPON_128
			} else if exponent_x < 0 {
				exponent_x = 0
			}
			res.w[1] |= uint64(exponent_x) << 49
			res.w[0] = 0
			return res, pfpsf
		}
	}
	if !valid_y {
		// y is Inf. or NaN

		// test if y is NaN
		if (y.w[1] & 0x7c00000000000000) == 0x7c00000000000000 {
			if (y.w[1] & 0x7e00000000000000) == 0x7e00000000000000 { // sNaN
				pfpsf |= BID_INVALID_EXCEPTION
			}
			res.w[1] = CY.w[1] & QUIET_MASK64
			res.w[0] = CY.w[0]
			return res, pfpsf
		}
		// y is Infinity?
		if (y.w[1] & 0x7800000000000000) == 0x7800000000000000 {
			// return +/-0
			res.w[1] = sign_x ^ sign_y
			res.w[0] = 0
			return res, pfpsf
		}
		// y is 0, return +/-Inf
		pfpsf |= BID_ZERO_DIVIDE_EXCEPTION
		res.w[1] =
			((x.w[1] ^ y.w[1]) & 0x8000000000000000) | 0x7800000000000000
		res.w[0] = 0
		return res, pfpsf
	}

	diff_expon = exponent_x - exponent_y + EXPONENT_BIAS128

	if __unsigned_compare_gt_128(CY, CX) {
		// CX < CY

		// 2^64
		f64_i := uint32(0x5f800000)
		f64_d := math.Float32frombits(f64_i)

		// fx ~ CX,   fy ~ CY
		fx_d := noFmaMulAddF32(float32(CX.w[1]), f64_d, float32(CX.w[0]))
		fy_d := noFmaMulAddF32(float32(CY.w[1]), f64_d, float32(CY.w[0]))
		fx_i := math.Float32bits(fx_d)
		fy_i := math.Float32bits(fy_d)
		// expon_cy - expon_cx
		bin_index = int((fy_i - fx_i) >> 23)

		if CX.w[1] != 0 {
			T = bid_power10_index_binexp_128[bin_index].w[0]
			CA = __mul_64x128_short(T, CX)
		} else {
			T128 = bid_power10_index_binexp_128[bin_index]
			CA = __mul_64x128_short(CX.w[0], T128)
		}

		ed2 = 33
		if __unsigned_compare_gt_128(CY, CA) {
			ed2++
		}

		T128 = bid_power10_table_128[ed2]
		CA4 = __mul_128x128_to_256(CA, T128)

		ed2 += bid_estimate_decimal_digits[bin_index]
		CQ.w[0] = 0
		CQ.w[1] = 0
		diff_expon = diff_expon - ed2

	} else {
		// get CQ = CX/CY
		CQ, CR = bid___div_128_by_128(CX, CY)

		if CR.w[1] == 0 && CR.w[0] == 0 {
			res = bid_get_BID128(sign_x^sign_y, diff_expon, CQ, rnd_mode, &pfpsf)
			return res, pfpsf
		}
		// get number of decimal digits in CQ
		// 2^64
		f64_i := uint32(0x5f800000)
		f64_d := math.Float32frombits(f64_i)
		fx_d := noFmaMulAddF32(float32(CQ.w[1]), f64_d, float32(CQ.w[0]))
		fx_i := math.Float32bits(fx_d)
		// binary expon. of CQ
		bin_expon = int((fx_i - 0x3f800000) >> 23)

		digits_q = bid_estimate_decimal_digits[bin_expon]
		TP128.w[0] = bid_power10_index_binexp_128[bin_expon].w[0]
		TP128.w[1] = bid_power10_index_binexp_128[bin_expon].w[1]
		if __unsigned_compare_ge_128(CQ, TP128) {
			digits_q++
		}

		ed2 = 34 - digits_q
		T128.w[0] = bid_power10_table_128[ed2].w[0]
		T128.w[1] = bid_power10_table_128[ed2].w[1]
		CA4 = __mul_128x128_to_256(CR, T128)
		diff_expon = diff_expon - ed2
		CQ = __mul_128x128_low(CQ, T128)

	}

	bid___div_256_by_128(&CQ, &CA4, CY)

	if CA4.w[0] != 0 || CA4.w[1] != 0 {
		// set status flags
		pfpsf |= BID_INEXACT_EXCEPTION
	} else {
		// check whether result is exact
		// check whether CX, CY are short
		if CX.w[1] == 0 && CY.w[1] == 0 && (CX.w[0] <= 1024) && (CY.w[0] <= 1024) {
			i = int(CY.w[0]) - 1
			j = int(CX.w[0]) - 1
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
			T128.w[0] = 0x44909befeb9fad49
			T128.w[1] = 0x000b877aa3236a4b
			P256 = __mul_128x128_to_256(CQ, T128)
			//amount = bid_recip_scale[17];
			Q_high = (P256.w[2] >> 44) | (P256.w[3] << (64 - 44))
			Q_low = CQ.w[0] - Q_high*100000000000000000

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
					CQ.w[0] = CQ.w[1] >> uint(amount)
				} else {
					CQ.w[0] = Q_high
				}
				CQ.w[1] = 0

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
			CA4r.w[1] = (CA4.w[1] + CA4.w[1]) | (CA4.w[0] >> 63)
			CA4r.w[0] = CA4.w[0] + CA4.w[0]
			CA4r.w[0], carry64 = __sub_borrow_out(CA4r.w[0], CY.w[0])
			CA4r.w[1] = CA4r.w[1] - CY.w[1] - carry64
			if (CA4r.w[1] | CA4r.w[0]) != 0 {
				D = 1
			} else {
				D = 0
			}
			carry64 = uint64(1+int64(CA4r.w[1])>>63) & ((CQ.w[0]) | D)
			CQ.w[0] += carry64
			if CQ.w[0] < carry64 {
				CQ.w[1]++
			}
		case BID_ROUNDING_TIES_AWAY:
			// rounding
			// 2*CA4 - CY
			CA4r.w[1] = (CA4.w[1] + CA4.w[1]) | (CA4.w[0] >> 63)
			CA4r.w[0] = CA4.w[0] + CA4.w[0]
			CA4r.w[0], carry64 = __sub_borrow_out(CA4r.w[0], CY.w[0])
			CA4r.w[1] = CA4r.w[1] - CY.w[1] - carry64
			if (CA4r.w[1] | CA4r.w[0]) != 0 {
				D = 0
			} else {
				D = 1
			}
			carry64 = uint64(1+int64(CA4r.w[1])>>63) | D
			CQ.w[0] += carry64
			if CQ.w[0] < carry64 {
				CQ.w[1]++
			}
		case BID_ROUNDING_DOWN, BID_ROUNDING_TO_ZERO:
			// do nothing
		default: // rounding up
			CQ.w[0]++
			if CQ.w[0] == 0 {
				CQ.w[1]++
			}
		}

	} else {
		if CA4.w[0] != 0 || CA4.w[1] != 0 {
			// set status flags
			pfpsf |= BID_INEXACT_EXCEPTION
		}

		res = bid_handle_UF_128_rem(sign_x^sign_y, diff_expon, CQ,
			CA4.w[1]|CA4.w[0], rnd_mode, &pfpsf)
		return res, pfpsf
	}

	res = bid_get_BID128(sign_x^sign_y, diff_expon, CQ, rnd_mode, &pfpsf)
	return res, pfpsf
}
