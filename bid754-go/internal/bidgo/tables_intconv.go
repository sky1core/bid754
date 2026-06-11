package bidgo

// DEC_DIGITS is ported mechanically from Intel bid_internal.h.
type DEC_DIGITS struct {
	digits       uint32
	threshold_hi uint64
	threshold_lo uint64
	digits1      uint32
}

// bid_nr_digits is ported mechanically from Intel bid128.c.
var bid_nr_digits = [...]DEC_DIGITS{
	{digits: 1, threshold_hi: 0x0000000000000000, threshold_lo: 0x000000000000000a, digits1: 1},
	{digits: 1, threshold_hi: 0x0000000000000000, threshold_lo: 0x000000000000000a, digits1: 1},
	{digits: 1, threshold_hi: 0x0000000000000000, threshold_lo: 0x000000000000000a, digits1: 1},
	{digits: 0, threshold_hi: 0x0000000000000000, threshold_lo: 0x000000000000000a, digits1: 1},
	{digits: 2, threshold_hi: 0x0000000000000000, threshold_lo: 0x0000000000000064, digits1: 2},
	{digits: 2, threshold_hi: 0x0000000000000000, threshold_lo: 0x0000000000000064, digits1: 2},
	{digits: 0, threshold_hi: 0x0000000000000000, threshold_lo: 0x0000000000000064, digits1: 2},
	{digits: 3, threshold_hi: 0x0000000000000000, threshold_lo: 0x00000000000003e8, digits1: 3},
	{digits: 3, threshold_hi: 0x0000000000000000, threshold_lo: 0x00000000000003e8, digits1: 3},
	{digits: 0, threshold_hi: 0x0000000000000000, threshold_lo: 0x00000000000003e8, digits1: 3},
	{digits: 4, threshold_hi: 0x0000000000000000, threshold_lo: 0x0000000000002710, digits1: 4},
	{digits: 4, threshold_hi: 0x0000000000000000, threshold_lo: 0x0000000000002710, digits1: 4},
	{digits: 4, threshold_hi: 0x0000000000000000, threshold_lo: 0x0000000000002710, digits1: 4},
	{digits: 0, threshold_hi: 0x0000000000000000, threshold_lo: 0x0000000000002710, digits1: 4},
	{digits: 5, threshold_hi: 0x0000000000000000, threshold_lo: 0x00000000000186a0, digits1: 5},
	{digits: 5, threshold_hi: 0x0000000000000000, threshold_lo: 0x00000000000186a0, digits1: 5},
	{digits: 0, threshold_hi: 0x0000000000000000, threshold_lo: 0x00000000000186a0, digits1: 5},
	{digits: 6, threshold_hi: 0x0000000000000000, threshold_lo: 0x00000000000f4240, digits1: 6},
	{digits: 6, threshold_hi: 0x0000000000000000, threshold_lo: 0x00000000000f4240, digits1: 6},
	{digits: 0, threshold_hi: 0x0000000000000000, threshold_lo: 0x00000000000f4240, digits1: 6},
	{digits: 7, threshold_hi: 0x0000000000000000, threshold_lo: 0x0000000000989680, digits1: 7},
	{digits: 7, threshold_hi: 0x0000000000000000, threshold_lo: 0x0000000000989680, digits1: 7},
	{digits: 7, threshold_hi: 0x0000000000000000, threshold_lo: 0x0000000000989680, digits1: 7},
	{digits: 0, threshold_hi: 0x0000000000000000, threshold_lo: 0x0000000000989680, digits1: 7},
	{digits: 8, threshold_hi: 0x0000000000000000, threshold_lo: 0x0000000005f5e100, digits1: 8},
	{digits: 8, threshold_hi: 0x0000000000000000, threshold_lo: 0x0000000005f5e100, digits1: 8},
	{digits: 0, threshold_hi: 0x0000000000000000, threshold_lo: 0x0000000005f5e100, digits1: 8},
	{digits: 9, threshold_hi: 0x0000000000000000, threshold_lo: 0x000000003b9aca00, digits1: 9},
	{digits: 9, threshold_hi: 0x0000000000000000, threshold_lo: 0x000000003b9aca00, digits1: 9},
	{digits: 0, threshold_hi: 0x0000000000000000, threshold_lo: 0x000000003b9aca00, digits1: 9},
	{digits: 10, threshold_hi: 0x0000000000000000, threshold_lo: 0x00000002540be400, digits1: 10},
	{digits: 10, threshold_hi: 0x0000000000000000, threshold_lo: 0x00000002540be400, digits1: 10},
	{digits: 10, threshold_hi: 0x0000000000000000, threshold_lo: 0x00000002540be400, digits1: 10},
	{digits: 0, threshold_hi: 0x0000000000000000, threshold_lo: 0x00000002540be400, digits1: 10},
	{digits: 11, threshold_hi: 0x0000000000000000, threshold_lo: 0x000000174876e800, digits1: 11},
	{digits: 11, threshold_hi: 0x0000000000000000, threshold_lo: 0x000000174876e800, digits1: 11},
	{digits: 0, threshold_hi: 0x0000000000000000, threshold_lo: 0x000000174876e800, digits1: 11},
	{digits: 12, threshold_hi: 0x0000000000000000, threshold_lo: 0x000000e8d4a51000, digits1: 12},
	{digits: 12, threshold_hi: 0x0000000000000000, threshold_lo: 0x000000e8d4a51000, digits1: 12},
	{digits: 0, threshold_hi: 0x0000000000000000, threshold_lo: 0x000000e8d4a51000, digits1: 12},
	{digits: 13, threshold_hi: 0x0000000000000000, threshold_lo: 0x000009184e72a000, digits1: 13},
	{digits: 13, threshold_hi: 0x0000000000000000, threshold_lo: 0x000009184e72a000, digits1: 13},
	{digits: 13, threshold_hi: 0x0000000000000000, threshold_lo: 0x000009184e72a000, digits1: 13},
	{digits: 0, threshold_hi: 0x0000000000000000, threshold_lo: 0x000009184e72a000, digits1: 13},
	{digits: 14, threshold_hi: 0x0000000000000000, threshold_lo: 0x00005af3107a4000, digits1: 14},
	{digits: 14, threshold_hi: 0x0000000000000000, threshold_lo: 0x00005af3107a4000, digits1: 14},
	{digits: 0, threshold_hi: 0x0000000000000000, threshold_lo: 0x00005af3107a4000, digits1: 14},
	{digits: 15, threshold_hi: 0x0000000000000000, threshold_lo: 0x00038d7ea4c68000, digits1: 15},
	{digits: 15, threshold_hi: 0x0000000000000000, threshold_lo: 0x00038d7ea4c68000, digits1: 15},
	{digits: 0, threshold_hi: 0x0000000000000000, threshold_lo: 0x00038d7ea4c68000, digits1: 15},
	{digits: 16, threshold_hi: 0x0000000000000000, threshold_lo: 0x002386f26fc10000, digits1: 16},
	{digits: 16, threshold_hi: 0x0000000000000000, threshold_lo: 0x002386f26fc10000, digits1: 16},
	{digits: 16, threshold_hi: 0x0000000000000000, threshold_lo: 0x002386f26fc10000, digits1: 16},
	{digits: 0, threshold_hi: 0x0000000000000000, threshold_lo: 0x002386f26fc10000, digits1: 16},
	{digits: 17, threshold_hi: 0x0000000000000000, threshold_lo: 0x016345785d8a0000, digits1: 17},
	{digits: 17, threshold_hi: 0x0000000000000000, threshold_lo: 0x016345785d8a0000, digits1: 17},
	{digits: 0, threshold_hi: 0x0000000000000000, threshold_lo: 0x016345785d8a0000, digits1: 17},
	{digits: 18, threshold_hi: 0x0000000000000000, threshold_lo: 0x0de0b6b3a7640000, digits1: 18},
	{digits: 18, threshold_hi: 0x0000000000000000, threshold_lo: 0x0de0b6b3a7640000, digits1: 18},
	{digits: 0, threshold_hi: 0x0000000000000000, threshold_lo: 0x0de0b6b3a7640000, digits1: 18},
	{digits: 19, threshold_hi: 0x0000000000000000, threshold_lo: 0x8ac7230489e80000, digits1: 19},
	{digits: 19, threshold_hi: 0x0000000000000000, threshold_lo: 0x8ac7230489e80000, digits1: 19},
	{digits: 19, threshold_hi: 0x0000000000000000, threshold_lo: 0x8ac7230489e80000, digits1: 19},
	{digits: 0, threshold_hi: 0x0000000000000000, threshold_lo: 0x8ac7230489e80000, digits1: 19},
	{digits: 20, threshold_hi: 0x0000000000000005, threshold_lo: 0x6bc75e2d63100000, digits1: 20},
	{digits: 20, threshold_hi: 0x0000000000000005, threshold_lo: 0x6bc75e2d63100000, digits1: 20},
	{digits: 0, threshold_hi: 0x0000000000000005, threshold_lo: 0x6bc75e2d63100000, digits1: 20},
	{digits: 21, threshold_hi: 0x0000000000000036, threshold_lo: 0x35c9adc5dea00000, digits1: 21},
	{digits: 21, threshold_hi: 0x0000000000000036, threshold_lo: 0x35c9adc5dea00000, digits1: 21},
	{digits: 0, threshold_hi: 0x0000000000000036, threshold_lo: 0x35c9adc5dea00000, digits1: 21},
	{digits: 22, threshold_hi: 0x000000000000021e, threshold_lo: 0x19e0c9bab2400000, digits1: 22},
	{digits: 22, threshold_hi: 0x000000000000021e, threshold_lo: 0x19e0c9bab2400000, digits1: 22},
	{digits: 22, threshold_hi: 0x000000000000021e, threshold_lo: 0x19e0c9bab2400000, digits1: 22},
	{digits: 0, threshold_hi: 0x000000000000021e, threshold_lo: 0x19e0c9bab2400000, digits1: 22},
	{digits: 23, threshold_hi: 0x000000000000152d, threshold_lo: 0x02c7e14af6800000, digits1: 23},
	{digits: 23, threshold_hi: 0x000000000000152d, threshold_lo: 0x02c7e14af6800000, digits1: 23},
	{digits: 0, threshold_hi: 0x000000000000152d, threshold_lo: 0x02c7e14af6800000, digits1: 23},
	{digits: 24, threshold_hi: 0x000000000000d3c2, threshold_lo: 0x1bcecceda1000000, digits1: 24},
	{digits: 24, threshold_hi: 0x000000000000d3c2, threshold_lo: 0x1bcecceda1000000, digits1: 24},
	{digits: 0, threshold_hi: 0x000000000000d3c2, threshold_lo: 0x1bcecceda1000000, digits1: 24},
	{digits: 25, threshold_hi: 0x0000000000084595, threshold_lo: 0x161401484a000000, digits1: 25},
	{digits: 25, threshold_hi: 0x0000000000084595, threshold_lo: 0x161401484a000000, digits1: 25},
	{digits: 25, threshold_hi: 0x0000000000084595, threshold_lo: 0x161401484a000000, digits1: 25},
	{digits: 0, threshold_hi: 0x0000000000084595, threshold_lo: 0x161401484a000000, digits1: 25},
	{digits: 26, threshold_hi: 0x000000000052b7d2, threshold_lo: 0xdcc80cd2e4000000, digits1: 26},
	{digits: 26, threshold_hi: 0x000000000052b7d2, threshold_lo: 0xdcc80cd2e4000000, digits1: 26},
	{digits: 0, threshold_hi: 0x000000000052b7d2, threshold_lo: 0xdcc80cd2e4000000, digits1: 26},
	{digits: 27, threshold_hi: 0x00000000033b2e3c, threshold_lo: 0x9fd0803ce8000000, digits1: 27},
	{digits: 27, threshold_hi: 0x00000000033b2e3c, threshold_lo: 0x9fd0803ce8000000, digits1: 27},
	{digits: 0, threshold_hi: 0x00000000033b2e3c, threshold_lo: 0x9fd0803ce8000000, digits1: 27},
	{digits: 28, threshold_hi: 0x00000000204fce5e, threshold_lo: 0x3e25026110000000, digits1: 28},
	{digits: 28, threshold_hi: 0x00000000204fce5e, threshold_lo: 0x3e25026110000000, digits1: 28},
	{digits: 28, threshold_hi: 0x00000000204fce5e, threshold_lo: 0x3e25026110000000, digits1: 28},
	{digits: 0, threshold_hi: 0x00000000204fce5e, threshold_lo: 0x3e25026110000000, digits1: 28},
	{digits: 29, threshold_hi: 0x00000001431e0fae, threshold_lo: 0x6d7217caa0000000, digits1: 29},
	{digits: 29, threshold_hi: 0x00000001431e0fae, threshold_lo: 0x6d7217caa0000000, digits1: 29},
	{digits: 0, threshold_hi: 0x00000001431e0fae, threshold_lo: 0x6d7217caa0000000, digits1: 29},
	{digits: 30, threshold_hi: 0x0000000c9f2c9cd0, threshold_lo: 0x4674edea40000000, digits1: 30},
	{digits: 30, threshold_hi: 0x0000000c9f2c9cd0, threshold_lo: 0x4674edea40000000, digits1: 30},
	{digits: 0, threshold_hi: 0x0000000c9f2c9cd0, threshold_lo: 0x4674edea40000000, digits1: 30},
	{digits: 31, threshold_hi: 0x0000007e37be2022, threshold_lo: 0xc0914b2680000000, digits1: 31},
	{digits: 31, threshold_hi: 0x0000007e37be2022, threshold_lo: 0xc0914b2680000000, digits1: 31},
	{digits: 0, threshold_hi: 0x0000007e37be2022, threshold_lo: 0xc0914b2680000000, digits1: 31},
	{digits: 32, threshold_hi: 0x000004ee2d6d415b, threshold_lo: 0x85acef8100000000, digits1: 32},
	{digits: 32, threshold_hi: 0x000004ee2d6d415b, threshold_lo: 0x85acef8100000000, digits1: 32},
	{digits: 32, threshold_hi: 0x000004ee2d6d415b, threshold_lo: 0x85acef8100000000, digits1: 32},
	{digits: 0, threshold_hi: 0x000004ee2d6d415b, threshold_lo: 0x85acef8100000000, digits1: 32},
	{digits: 33, threshold_hi: 0x0000314dc6448d93, threshold_lo: 0x38c15b0a00000000, digits1: 33},
	{digits: 33, threshold_hi: 0x0000314dc6448d93, threshold_lo: 0x38c15b0a00000000, digits1: 33},
	{digits: 0, threshold_hi: 0x0000314dc6448d93, threshold_lo: 0x38c15b0a00000000, digits1: 33},
	{digits: 34, threshold_hi: 0x0001ed09bead87c0, threshold_lo: 0x378d8e6400000000, digits1: 34},
	{digits: 34, threshold_hi: 0x0001ed09bead87c0, threshold_lo: 0x378d8e6400000000, digits1: 34},
	{digits: 0, threshold_hi: 0x0001ed09bead87c0, threshold_lo: 0x378d8e6400000000, digits1: 34},
	{digits: 35, threshold_hi: 0x0013426172c74d82, threshold_lo: 0x2b878fe800000000, digits1: 35},
}

// bid_shiftright128 is ported mechanically from Intel bid128.c.
// Full 34-entry table (indices 0..33 for 128-bit operations).
var bid_shiftright128 = [34]int{
	0,   // 128 - 128
	0,   // 128 - 128
	0,   // 128 - 128
	3,   // 131 - 128
	6,   // 134 - 128
	9,   // 137 - 128
	13,  // 141 - 128
	16,  // 144 - 128
	19,  // 147 - 128
	23,  // 151 - 128
	26,  // 154 - 128
	29,  // 157 - 128
	33,  // 161 - 128
	36,  // 164 - 128
	39,  // 167 - 128
	43,  // 171 - 128
	46,  // 174 - 128
	49,  // 177 - 128
	53,  // 181 - 128
	56,  // 184 - 128
	59,  // 187 - 128
	63,  // 191 - 128
	66,  // 194 - 128
	69,  // 197 - 128
	73,  // 201 - 128
	76,  // 204 - 128
	79,  // 207 - 128
	83,  // 211 - 128
	86,  // 214 - 128
	89,  // 217 - 128
	92,  // 220 - 128
	96,  // 224 - 128
	99,  // 227 - 128
	102, // 230 - 128
}

// bid_maskhigh128 is ported mechanically from Intel bid128.c.
// Full 34-entry table (indices 0..33 for 128-bit operations).
var bid_maskhigh128 = [34]uint64{
	0x0000000000000000, //  0 = 128 - 128 bits
	0x0000000000000000, //  0 = 128 - 128 bits
	0x0000000000000000, //  0 = 128 - 128 bits
	0x0000000000000007, //  3 = 131 - 128 bits
	0x000000000000003f, //  6 = 134 - 128 bits
	0x00000000000001ff, //  9 = 137 - 128 bits
	0x0000000000001fff, // 13 = 141 - 128 bits
	0x000000000000ffff, // 16 = 144 - 128 bits
	0x000000000007ffff, // 19 = 147 - 128 bits
	0x00000000007fffff, // 23 = 151 - 128 bits
	0x0000000003ffffff, // 26 = 154 - 128 bits
	0x000000001fffffff, // 29 = 157 - 128 bits
	0x00000001ffffffff, // 33 = 161 - 128 bits
	0x0000000fffffffff, // 36 = 164 - 128 bits
	0x0000007fffffffff, // 39 = 167 - 128 bits
	0x000007ffffffffff, // 43 = 171 - 128 bits
	0x00003fffffffffff, // 46 = 174 - 128 bits
	0x0001ffffffffffff, // 49 = 177 - 128 bits
	0x001fffffffffffff, // 53 = 181 - 128 bits
	0x00ffffffffffffff, // 56 = 184 - 128 bits
	0x07ffffffffffffff, // 59 = 187 - 128 bits
	0x7fffffffffffffff, // 63 = 191 - 128 bits
	0x0000000000000003, //  2 = 194 - 192 bits
	0x000000000000001f, //  5 = 197 - 192 bits
	0x00000000000001ff, //  9 = 201 - 192 bits
	0x0000000000000fff, // 12 = 204 - 192 bits
	0x0000000000007fff, // 15 = 207 - 192 bits
	0x000000000007ffff, // 21 = 211 - 192 bits
	0x00000000003fffff, // 22 = 214 - 192 bits
	0x0000000001ffffff, // 25 = 217 - 192 bits
	0x000000000fffffff, // 28 = 220 - 192 bits
	0x00000000ffffffff, // 32 = 224 - 192 bits
	0x00000007ffffffff, // 35 = 227 - 192 bits
	0x0000003fffffffff, // 38 = 230 - 192 bits
}

// bid_onehalf128 is ported mechanically from Intel bid128.c.
// Full 34-entry table (indices 0..33 for 128-bit operations).
var bid_onehalf128 = [34]uint64{
	0x0000000000000000, //  0 bits
	0x0000000000000000, //  0 bits
	0x0000000000000000, //  0 bits
	0x0000000000000004, //  3 bits
	0x0000000000000020, //  6 bits
	0x0000000000000100, //  9 bits
	0x0000000000001000, // 13 bits
	0x0000000000008000, // 16 bits
	0x0000000000040000, // 19 bits
	0x0000000000400000, // 23 bits
	0x0000000002000000, // 26 bits
	0x0000000010000000, // 29 bits
	0x0000000100000000, // 33 bits
	0x0000000800000000, // 36 bits
	0x0000004000000000, // 39 bits
	0x0000040000000000, // 43 bits
	0x0000200000000000, // 46 bits
	0x0001000000000000, // 49 bits
	0x0010000000000000, // 53 bits
	0x0080000000000000, // 56 bits
	0x0400000000000000, // 59 bits
	0x4000000000000000, // 63 bits
	0x0000000000000002, // 66 bits
	0x0000000000000010, // 69 bits
	0x0000000000000100, // 73 bits
	0x0000000000000800, // 76 bits
	0x0000000000004000, // 79 bits
	0x0000000000040000, // 83 bits
	0x0000000000200000, // 86 bits
	0x0000000001000000, // 89 bits
	0x0000000008000000, // 92 bits
	0x0000000080000000, // 96 bits
	0x0000000400000000, // 99 bits
	0x0000002000000000, // 102 bits
}

// bid_ten2mk64 is ported mechanically from Intel bid128.c.
var bid_ten2mk64 = bid_ten2mk64_round64

// bid_ten2mk128 is ported mechanically from Intel bid128.c.
// bid_ten2mk128[k-1] = 10^(-k) * 2^exp(k), where 1 <= k <= 34 and
// exp(k) = bid_shiftright128[k-1] + 128 (rounded up to 118 bits).
var bid_ten2mk128 = [34]BID_UINT128{
	{w: [2]uint64{0x999999999999999a, 0x1999999999999999}}, //  10^(-1) * 2^128
	{w: [2]uint64{0x28f5c28f5c28f5c3, 0x028f5c28f5c28f5c}}, //  10^(-2) * 2^128
	{w: [2]uint64{0x9db22d0e56041894, 0x004189374bc6a7ef}}, //  10^(-3) * 2^128
	{w: [2]uint64{0x4af4f0d844d013aa, 0x00346dc5d6388659}}, //  10^(-4) * 2^131
	{w: [2]uint64{0x08c3f3e0370cdc88, 0x0029f16b11c6d1e1}}, //  10^(-5) * 2^134
	{w: [2]uint64{0x6d698fe69270b06d, 0x00218def416bdb1a}}, //  10^(-6) * 2^137
	{w: [2]uint64{0xaf0f4ca41d811a47, 0x0035afe535795e90}}, //  10^(-7) * 2^141
	{w: [2]uint64{0xbf3f70834acdaea0, 0x002af31dc4611873}}, //  10^(-8) * 2^144
	{w: [2]uint64{0x65cc5a02a23e254d, 0x00225c17d04dad29}}, //  10^(-9) * 2^147
	{w: [2]uint64{0x6fad5cd10396a214, 0x0036f9bfb3af7b75}}, // 10^(-10) * 2^151
	{w: [2]uint64{0xbfbde3da69454e76, 0x002bfaffc2f2c92a}}, // 10^(-11) * 2^154
	{w: [2]uint64{0x32fe4fe1edd10b92, 0x00232f33025bd422}}, // 10^(-12) * 2^157
	{w: [2]uint64{0x84ca19697c81ac1c, 0x00384b84d092ed03}}, // 10^(-13) * 2^161
	{w: [2]uint64{0x03d4e1213067bce4, 0x002d09370d425736}}, // 10^(-14) * 2^164
	{w: [2]uint64{0x3643e74dc052fd83, 0x0024075f3dceac2b}}, // 10^(-15) * 2^167
	{w: [2]uint64{0x56d30baf9a1e626b, 0x0039a5652fb11378}}, // 10^(-16) * 2^171
	{w: [2]uint64{0x12426fbfae7eb522, 0x002e1dea8c8da92d}}, // 10^(-17) * 2^174
	{w: [2]uint64{0x41cebfcc8b9890e8, 0x0024e4bba3a48757}}, // 10^(-18) * 2^177
	{w: [2]uint64{0x694acc7a78f41b0d, 0x003b07929f6da558}}, // 10^(-19) * 2^181
	{w: [2]uint64{0xbaa23d2ec729af3e, 0x002f394219248446}}, // 10^(-20) * 2^184
	{w: [2]uint64{0xfbb4fdbf05baf298, 0x0025c768141d369e}}, // 10^(-21) * 2^187
	{w: [2]uint64{0x2c54c931a2c4b759, 0x003c7240202ebdcb}}, // 10^(-22) * 2^191
	{w: [2]uint64{0x89dd6dc14f03c5e1, 0x00305b66802564a2}}, // 10^(-23) * 2^194
	{w: [2]uint64{0xd4b1249aa59c9e4e, 0x0026af8533511d4e}}, // 10^(-24) * 2^197
	{w: [2]uint64{0x544ea0f76f60fd49, 0x003de5a1ebb4fbb1}}, // 10^(-25) * 2^201
	{w: [2]uint64{0x76a54d92bf80caa1, 0x00318481895d9627}}, // 10^(-26) * 2^204
	{w: [2]uint64{0x921dd7a89933d54e, 0x00279d346de4781f}}, // 10^(-27) * 2^207
	{w: [2]uint64{0x8362f2a75b862215, 0x003f61ed7ca0c032}}, // 10^(-28) * 2^211
	{w: [2]uint64{0xcf825bb91604e811, 0x0032b4bdfd4d668e}}, // 10^(-29) * 2^214
	{w: [2]uint64{0x0c684960de6a5341, 0x00289097fdd7853f}}, // 10^(-30) * 2^217
	{w: [2]uint64{0x3d203ab3e521dc34, 0x002073accb12d0ff}}, // 10^(-31) * 2^220
	{w: [2]uint64{0x2e99f7863b696053, 0x0033ec47ab514e65}}, // 10^(-32) * 2^224
	{w: [2]uint64{0x587b2c6b62bab376, 0x002989d2ef743eb7}}, // 10^(-33) * 2^227
	{w: [2]uint64{0xad2f56bc4efbc2c5, 0x00213b0f25f69892}}, // 10^(-34) * 2^230
}

// bid_ten2mk128trunc is ported mechanically from Intel bid128.c.
var bid_ten2mk128trunc = [...]BID_UINT128{
	{w: [2]uint64{0x9999999999999999, 0x1999999999999999}},
	{w: [2]uint64{0x28f5c28f5c28f5c2, 0x028f5c28f5c28f5c}},
	{w: [2]uint64{0x9db22d0e56041893, 0x004189374bc6a7ef}},
	{w: [2]uint64{0x4af4f0d844d013a9, 0x00346dc5d6388659}},
	{w: [2]uint64{0x08c3f3e0370cdc87, 0x0029f16b11c6d1e1}},
	{w: [2]uint64{0x6d698fe69270b06c, 0x00218def416bdb1a}},
	{w: [2]uint64{0xaf0f4ca41d811a46, 0x0035afe535795e90}},
	{w: [2]uint64{0xbf3f70834acdae9f, 0x002af31dc4611873}},
	{w: [2]uint64{0x65cc5a02a23e254c, 0x00225c17d04dad29}},
	{w: [2]uint64{0x6fad5cd10396a213, 0x0036f9bfb3af7b75}},
	{w: [2]uint64{0xbfbde3da69454e75, 0x002bfaffc2f2c92a}},
	{w: [2]uint64{0x32fe4fe1edd10b91, 0x00232f33025bd422}},
	{w: [2]uint64{0x84ca19697c81ac1b, 0x00384b84d092ed03}},
	{w: [2]uint64{0x03d4e1213067bce3, 0x002d09370d425736}},
	{w: [2]uint64{0x3643e74dc052fd82, 0x0024075f3dceac2b}},
	{w: [2]uint64{0x56d30baf9a1e626a, 0x0039a5652fb11378}},
	{w: [2]uint64{0x12426fbfae7eb521, 0x002e1dea8c8da92d}},
	{w: [2]uint64{0x41cebfcc8b9890e7, 0x0024e4bba3a48757}},
	{w: [2]uint64{0x694acc7a78f41b0c, 0x003b07929f6da558}},
	{w: [2]uint64{0xbaa23d2ec729af3d, 0x002f394219248446}},
	{w: [2]uint64{0xfbb4fdbf05baf297, 0x0025c768141d369e}},
	{w: [2]uint64{0x2c54c931a2c4b758, 0x003c7240202ebdcb}},
	{w: [2]uint64{0x89dd6dc14f03c5e0, 0x00305b66802564a2}},
	{w: [2]uint64{0xd4b1249aa59c9e4d, 0x0026af8533511d4e}},
	{w: [2]uint64{0x544ea0f76f60fd48, 0x003de5a1ebb4fbb1}},
	{w: [2]uint64{0x76a54d92bf80caa0, 0x00318481895d9627}},
	{w: [2]uint64{0x921dd7a89933d54d, 0x00279d346de4781f}},
	{w: [2]uint64{0x8362f2a75b862214, 0x003f61ed7ca0c032}},
	{w: [2]uint64{0xcf825bb91604e810, 0x0032b4bdfd4d668e}},
	{w: [2]uint64{0x0c684960de6a5340, 0x00289097fdd7853f}},
	{w: [2]uint64{0x3d203ab3e521dc33, 0x002073accb12d0ff}},
	{w: [2]uint64{0x2e99f7863b696052, 0x0033ec47ab514e65}},
	{w: [2]uint64{0x587b2c6b62bab375, 0x002989d2ef743eb7}},
	{w: [2]uint64{0xad2f56bc4efbc2c4, 0x00213b0f25f69892}},
}
