package bidgo

// Intel BID rounding tables - 기계적 포팅
// Source: IntelRDFPMathLib20U4/LIBRARY/src/bid128.c

// bid_midpoint64 - 1/2 * 10^x for x = 1 to 19
// Used in bid_round64_2_18 to add 0.5 ULP before truncation
var bid_midpoint64 = [19]uint64{
	0x0000000000000005, // 1/2 * 10^1 = 5 * 10^0
	0x0000000000000032, // 1/2 * 10^2 = 5 * 10^1
	0x00000000000001f4, // 1/2 * 10^3 = 5 * 10^2
	0x0000000000001388, // 1/2 * 10^4 = 5 * 10^3
	0x000000000000c350, // 1/2 * 10^5 = 5 * 10^4
	0x000000000007a120, // 1/2 * 10^6 = 5 * 10^5
	0x00000000004c4b40, // 1/2 * 10^7 = 5 * 10^6
	0x0000000002faf080, // 1/2 * 10^8 = 5 * 10^7
	0x000000001dcd6500, // 1/2 * 10^9 = 5 * 10^8
	0x000000012a05f200, // 1/2 * 10^10 = 5 * 10^9
	0x0000000ba43b7400, // 1/2 * 10^11 = 5 * 10^10
	0x000000746a528800, // 1/2 * 10^12 = 5 * 10^11
	0x0000048c27395000, // 1/2 * 10^13 = 5 * 10^12
	0x00002d79883d2000, // 1/2 * 10^14 = 5 * 10^13
	0x0001c6bf52634000, // 1/2 * 10^15 = 5 * 10^14
	0x0011c37937e08000, // 1/2 * 10^16 = 5 * 10^15
	0x00b1a2bc2ec50000, // 1/2 * 10^17 = 5 * 10^16
	0x06f05b59d3b20000, // 1/2 * 10^18 = 5 * 10^17
	0x4563918244f40000, // 1/2 * 10^19 = 5 * 10^18
}

// bid_Kx64 - 10^(-x) approximation, rounded up to 64 bits
// Kx from 10^(-x) ~= Kx * 2^(-Ex); Kx rounded up to 64 bits, 1 <= x <= 17
var bid_Kx64 = [17]uint64{
	0xcccccccccccccccd, // 10^-1 ~= cccccccccccccccd * 2^-67
	0xa3d70a3d70a3d70b, // 10^-2 ~= a3d70a3d70a3d70b * 2^-70
	0x83126e978d4fdf3c, // 10^-3 ~= 83126e978d4fdf3c * 2^-73
	0xd1b71758e219652c, // 10^-4 ~= d1b71758e219652c * 2^-77
	0xa7c5ac471b478424, // 10^-5 ~= a7c5ac471b478424 * 2^-80
	0x8637bd05af6c69b6, // 10^-6 ~= 8637bd05af6c69b6 * 2^-83
	0xd6bf94d5e57a42bd, // 10^-7 ~= d6bf94d5e57a42bd * 2^-87
	0xabcc77118461cefd, // 10^-8 ~= abcc77118461cefd * 2^-90
	0x89705f4136b4a598, // 10^-9 ~= 89705f4136b4a598 * 2^-93
	0xdbe6fecebdedd5bf, // 10^-10 ~= dbe6fecebdedd5bf * 2^-97
	0xafebff0bcb24aaff, // 10^-11 ~= afebff0bcb24aaff * 2^-100
	0x8cbccc096f5088cc, // 10^-12 ~= 8cbccc096f5088cc * 2^-103
	0xe12e13424bb40e14, // 10^-13 ~= e12e13424bb40e14 * 2^-107
	0xb424dc35095cd810, // 10^-14 ~= b424dc35095cd810 * 2^-110
	0x901d7cf73ab0acda, // 10^-15 ~= 901d7cf73ab0acda * 2^-113
	0xe69594bec44de15c, // 10^-16 ~= e69594bec44de15c * 2^-117
	0xb877aa3236a4b44a, // 10^-17 ~= b877aa3236a4b44a * 2^-120
}

// bid_Ex64m64 - Ex-64 from 10^(-x) ~= Kx * 2^(-Ex)
// Kx rounded up to 64 bits, 1 <= x <= 17
var bid_Ex64m64 = [17]uint32{
	3,  // 67 - 64, Ex = 67
	6,  // 70 - 64, Ex = 70
	9,  // 73 - 64, Ex = 73
	13, // 77 - 64, Ex = 77
	16, // 80 - 64, Ex = 80
	19, // 83 - 64, Ex = 83
	23, // 87 - 64, Ex = 87
	26, // 90 - 64, Ex = 90
	29, // 93 - 64, Ex = 93
	33, // 97 - 64, Ex = 97
	36, // 100 - 64, Ex = 100
	39, // 103 - 64, Ex = 103
	43, // 107 - 64, Ex = 107
	46, // 110 - 64, Ex = 110
	49, // 113 - 64, Ex = 113
	53, // 117 - 64, Ex = 117
	56, // 120 - 64, Ex = 120
}

// bid_half64 - Values of 1/2 in the right position to be compared with the fraction
// from C * kx, 1 <= x <= 17; the fraction consists of the low Ex bits in C * kx
// (these values are aligned with the high 64 bits of the fraction)
var bid_half64 = [17]uint64{
	0x0000000000000004, // half / 2^64 = 4
	0x0000000000000020, // half / 2^64 = 20
	0x0000000000000100, // half / 2^64 = 100
	0x0000000000001000, // half / 2^64 = 1000
	0x0000000000008000, // half / 2^64 = 8000
	0x0000000000040000, // half / 2^64 = 40000
	0x0000000000400000, // half / 2^64 = 400000
	0x0000000002000000, // half / 2^64 = 2000000
	0x0000000010000000, // half / 2^64 = 10000000
	0x0000000100000000, // half / 2^64 = 100000000
	0x0000000800000000, // half / 2^64 = 800000000
	0x0000004000000000, // half / 2^64 = 4000000000
	0x0000040000000000, // half / 2^64 = 40000000000
	0x0000200000000000, // half / 2^64 = 200000000000
	0x0001000000000000, // half / 2^64 = 1000000000000
	0x0010000000000000, // half / 2^64 = 10000000000000
	0x0080000000000000, // half / 2^64 = 80000000000000
}

// bid_mask64 - Values of mask in the right position to obtain the high Ex - 64 bits
// of the fraction from C * kx, 1 <= x <= 17; the fraction consists of
// the low Ex bits in C * kx
var bid_mask64 = [17]uint64{
	0x0000000000000007, // mask / 2^64
	0x000000000000003f, // mask / 2^64
	0x00000000000001ff, // mask / 2^64
	0x0000000000001fff, // mask / 2^64
	0x000000000000ffff, // mask / 2^64
	0x000000000007ffff, // mask / 2^64
	0x00000000007fffff, // mask / 2^64
	0x0000000003ffffff, // mask / 2^64
	0x000000001fffffff, // mask / 2^64
	0x00000001ffffffff, // mask / 2^64
	0x0000000fffffffff, // mask / 2^64
	0x0000007fffffffff, // mask / 2^64
	0x000007ffffffffff, // mask / 2^64
	0x00003fffffffffff, // mask / 2^64
	0x0001ffffffffffff, // mask / 2^64
	0x001fffffffffffff, // mask / 2^64
	0x00ffffffffffffff, // mask / 2^64
}

// bid_ten2mxtrunc64 - Values of 10^(-x) truncated to Ex bits beyond the binary point,
// and in the right position to be compared with the fraction from C * kx,
// 1 <= x <= 17; the fraction consists of the low Ex bits in C * kx
// (these values are aligned with the low 64 bits of the fraction)
var bid_ten2mxtrunc64 = [17]uint64{
	0xcccccccccccccccc, // (ten2mx >> 64) = cccccccccccccccc
	0xa3d70a3d70a3d70a, // (ten2mx >> 64) = a3d70a3d70a3d70a
	0x83126e978d4fdf3b, // (ten2mx >> 64) = 83126e978d4fdf3b
	0xd1b71758e219652b, // (ten2mx >> 64) = d1b71758e219652b
	0xa7c5ac471b478423, // (ten2mx >> 64) = a7c5ac471b478423
	0x8637bd05af6c69b5, // (ten2mx >> 64) = 8637bd05af6c69b5
	0xd6bf94d5e57a42bc, // (ten2mx >> 64) = d6bf94d5e57a42bc
	0xabcc77118461cefc, // (ten2mx >> 64) = abcc77118461cefc
	0x89705f4136b4a597, // (ten2mx >> 64) = 89705f4136b4a597
	0xdbe6fecebdedd5be, // (ten2mx >> 64) = dbe6fecebdedd5be
	0xafebff0bcb24aafe, // (ten2mx >> 64) = afebff0bcb24aafe
	0x8cbccc096f5088cb, // (ten2mx >> 64) = 8cbccc096f5088cb
	0xe12e13424bb40e13, // (ten2mx >> 64) = e12e13424bb40e13
	0xb424dc35095cd80f, // (ten2mx >> 64) = b424dc35095cd80f
	0x901d7cf73ab0acd9, // (ten2mx >> 64) = 901d7cf73ab0acd9
	0xe69594bec44de15b, // (ten2mx >> 64) = e69594bec44de15b
	0xb877aa3236a4b449, // (ten2mx >> 64) = b877aa3236a4b449
}

// bid_ten2k64 - Powers of 10: 10^0 to 10^19
var bid_ten2k64 = [20]uint64{
	0x0000000000000001, // 10^0
	0x000000000000000a, // 10^1
	0x0000000000000064, // 10^2
	0x00000000000003e8, // 10^3
	0x0000000000002710, // 10^4
	0x00000000000186a0, // 10^5
	0x00000000000f4240, // 10^6
	0x0000000000989680, // 10^7
	0x0000000005f5e100, // 10^8
	0x000000003b9aca00, // 10^9
	0x00000002540be400, // 10^10
	0x000000174876e800, // 10^11
	0x000000e8d4a51000, // 10^12
	0x000009184e72a000, // 10^13
	0x00005af3107a4000, // 10^14
	0x00038d7ea4c68000, // 10^15
	0x002386f26fc10000, // 10^16
	0x016345785d8a0000, // 10^17
	0x0de0b6b3a7640000, // 10^18
	0x8ac7230489e80000, // 10^19 (20 digits)
}
