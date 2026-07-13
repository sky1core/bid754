// Ported from: IntelRDFPMathLib20U4/LIBRARY/src/bid_round.c
// Mechanical translation of Intel BID rounding functions for 128/192/256-bit.
// All logic, magic numbers, and table references are preserved exactly.

package bidgo

// bid_round128_19_38 rounds a number C with q decimal digits (19 <= q <= 38)
// to q - x digits (1 <= x <= 37).
// Returns: Cstar, incr_exp, is_midpoint_lt_even, is_midpoint_gt_even,
// is_midpoint_lt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint
func bid_round128_19_38(q int, x int, C BID_UINT128) (Cstar BID_UINT128, incr_exp int, is_midpoint_lt_even int, is_midpoint_gt_even int, is_inexact_lt_midpoint int, is_inexact_gt_midpoint int) {

	var P256 BID_UINT256
	var fstar BID_UINT256
	var tmp64 uint64
	var shift int
	var ind int

	// round a number C with q decimal digits, 19 <= q <= 38
	// to q - x digits, 1 <= x <= 37
	ind = x - 1    // 0 <= ind <= 36
	if ind <= 18 { // if 0 <= ind <= 18
		tmp64 = C.lo
		C.lo = C.lo + bid_midpoint64[ind]
		if C.lo < tmp64 {
			C.hi++
		}
	} else { // if 19 <= ind <= 37
		tmp64 = C.lo
		C.lo = C.lo + bid_midpoint128[ind-19].lo
		if C.lo < tmp64 {
			C.hi++
		}
		C.hi = C.hi + bid_midpoint128[ind-19].hi
	}
	// kx ~= 10^(-x), kx = bid_Kx128[ind] * 2^(-Ex), 0 <= ind <= 36
	// P256 = (C + 1/2 * 10^x) * kx * 2^Ex = (C + 1/2 * 10^x) * Kx
	P256 = __mul_128x128_to_256(C, bid_Kx128[ind])
	// calculate C* = floor (P256) and f*
	// Cstar = P256 >> Ex
	// fstar = low Ex bits of P256
	shift = int(bid_Ex128m128[ind]) // in [2, 63]
	if ind <= 18 {                  // if 0 <= ind <= 18
		Cstar.lo = (P256.w2 >> uint(shift)) | (P256.w3 << uint(64-shift))
		Cstar.hi = (P256.w3 >> uint(shift))
		fstar.w0 = P256.w0
		fstar.w1 = P256.w1
		fstar.w2 = P256.w2 & bid_mask128[ind]
		fstar.w3 = 0x0
	} else { // if 19 <= ind <= 37
		Cstar.lo = P256.w3 >> uint(shift)
		Cstar.hi = 0x0
		fstar.w0 = P256.w0
		fstar.w1 = P256.w1
		fstar.w2 = P256.w2
		fstar.w3 = P256.w3 & bid_mask128[ind]
	}

	// determine inexactness of the rounding of C*
	if ind <= 18 { // if 0 <= ind <= 18
		if fstar.w2 > bid_half128[ind] ||
			(fstar.w2 == bid_half128[ind] && (fstar.w1 != 0 || fstar.w0 != 0)) {
			// f* > 1/2 and the result may be exact
			tmp64 = fstar.w2 - bid_half128[ind]
			if tmp64 != 0 || fstar.w1 > bid_ten2mxtrunc128[ind].hi || (fstar.w1 == bid_ten2mxtrunc128[ind].hi && fstar.w0 > bid_ten2mxtrunc128[ind].lo) { // f* - 1/2 > 10^(-x)
				is_inexact_lt_midpoint = 1
			} // else the result is exact
		} else { // the result is inexact; f2* <= 1/2
			is_inexact_gt_midpoint = 1
		}
	} else { // if 19 <= ind <= 37
		if fstar.w3 > bid_half128[ind] || (fstar.w3 == bid_half128[ind] &&
			(fstar.w2 != 0 || fstar.w1 != 0 || fstar.w0 != 0)) {
			// f* > 1/2 and the result may be exact
			tmp64 = fstar.w3 - bid_half128[ind]
			if tmp64 != 0 || fstar.w2 != 0 || fstar.w1 > bid_ten2mxtrunc128[ind].hi || (fstar.w1 == bid_ten2mxtrunc128[ind].hi && fstar.w0 > bid_ten2mxtrunc128[ind].lo) { // f* - 1/2 > 10^(-x)
				is_inexact_lt_midpoint = 1
			} // else the result is exact
		} else { // the result is inexact; f2* <= 1/2
			is_inexact_gt_midpoint = 1
		}
	}
	// check for midpoints
	if fstar.w3 == 0 && fstar.w2 == 0 &&
		(fstar.w1 < bid_ten2mxtrunc128[ind].hi ||
			(fstar.w1 == bid_ten2mxtrunc128[ind].hi &&
				fstar.w0 <= bid_ten2mxtrunc128[ind].lo)) {
		// the result is a midpoint
		if Cstar.lo&0x01 != 0 { // Cstar is odd; MP in [EVEN, ODD]
			Cstar.lo-- // Cstar is now even
			if Cstar.lo == 0xffffffffffffffff {
				Cstar.hi--
			}
			is_midpoint_gt_even = 1
			is_inexact_lt_midpoint = 0
			is_inexact_gt_midpoint = 0
		} else { // else MP in [ODD, EVEN]
			is_midpoint_lt_even = 1
			is_inexact_lt_midpoint = 0
			is_inexact_gt_midpoint = 0
		}
	}
	// check for rounding overflow
	ind = q - x // 1 <= ind <= q - 1
	if ind <= 19 {
		if Cstar.hi == 0x0 && Cstar.lo == bid_ten2k64[ind] {
			Cstar.lo = bid_ten2k64[ind-1]
			incr_exp = 1
		} else {
			incr_exp = 0
		}
	} else if ind == 20 {
		if Cstar.hi == bid_ten2k128[0].hi &&
			Cstar.lo == bid_ten2k128[0].lo {
			Cstar.lo = bid_ten2k64[19]
			Cstar.hi = 0x0
			incr_exp = 1
		} else {
			incr_exp = 0
		}
	} else { // if 21 <= ind <= 37
		if Cstar.hi == bid_ten2k128[ind-20].hi &&
			Cstar.lo == bid_ten2k128[ind-20].lo {
			Cstar.lo = bid_ten2k128[ind-21].lo
			Cstar.hi = bid_ten2k128[ind-21].hi
			incr_exp = 1
		} else {
			incr_exp = 0
		}
	}
	return
}

// bid_round192_39_57 rounds a number C with q decimal digits (39 <= q <= 57)
// to q - x digits (1 <= x <= 56).
func bid_round192_39_57(q int, x int, C BID_UINT192) (Cstar BID_UINT192, incr_exp int, is_midpoint_lt_even int, is_midpoint_gt_even int, is_inexact_lt_midpoint int, is_inexact_gt_midpoint int) {

	var P384 BID_UINT384
	var fstar BID_UINT384
	var tmp64 uint64
	var shift int
	var ind int

	ind = x - 1    // 0 <= ind <= 55
	if ind <= 18 { // if 0 <= ind <= 18
		tmp64 = C.w0
		C.w0 = C.w0 + bid_midpoint64[ind]
		if C.w0 < tmp64 {
			C.w1++
			if C.w1 == 0x0 {
				C.w2++
			}
		}
	} else if ind <= 37 { // if 19 <= ind <= 37
		tmp64 = C.w0
		C.w0 = C.w0 + bid_midpoint128[ind-19].lo
		if C.w0 < tmp64 {
			C.w1++
			if C.w1 == 0x0 {
				C.w2++
			}
		}
		tmp64 = C.w1
		C.w1 = C.w1 + bid_midpoint128[ind-19].hi
		if C.w1 < tmp64 {
			C.w2++
		}
	} else { // if 38 <= ind <= 57 (actually ind <= 55)
		tmp64 = C.w0
		C.w0 = C.w0 + bid_midpoint192[ind-38].w0
		if C.w0 < tmp64 {
			C.w1++
			if C.w1 == 0x0 {
				C.w2++
			}
		}
		tmp64 = C.w1
		C.w1 = C.w1 + bid_midpoint192[ind-38].w1
		if C.w1 < tmp64 {
			C.w2++
		}
		C.w2 = C.w2 + bid_midpoint192[ind-38].w2
	}
	// P384 = (C + 1/2 * 10^x) * Kx
	P384 = __mul_192x192_to_384(C, bid_Kx192[ind])
	shift = int(bid_Ex192m192[ind])
	if ind <= 18 { // if 0 <= ind <= 18
		Cstar.w2 = (P384.w5 >> uint(shift))
		Cstar.w1 = (P384.w5 << uint(64-shift)) | (P384.w4 >> uint(shift))
		Cstar.w0 = (P384.w4 << uint(64-shift)) | (P384.w3 >> uint(shift))
		fstar.w5 = 0x0
		fstar.w4 = 0x0
		fstar.w3 = P384.w3 & bid_mask192[ind]
		fstar.w2 = P384.w2
		fstar.w1 = P384.w1
		fstar.w0 = P384.w0
	} else if ind <= 37 { // if 19 <= ind <= 37
		Cstar.w2 = 0x0
		Cstar.w1 = P384.w5 >> uint(shift)
		Cstar.w0 = (P384.w5 << uint(64-shift)) | (P384.w4 >> uint(shift))
		fstar.w5 = 0x0
		fstar.w4 = P384.w4 & bid_mask192[ind]
		fstar.w3 = P384.w3
		fstar.w2 = P384.w2
		fstar.w1 = P384.w1
		fstar.w0 = P384.w0
	} else { // if 38 <= ind <= 57
		Cstar.w2 = 0x0
		Cstar.w1 = 0x0
		Cstar.w0 = P384.w5 >> uint(shift)
		fstar.w5 = P384.w5 & bid_mask192[ind]
		fstar.w4 = P384.w4
		fstar.w3 = P384.w3
		fstar.w2 = P384.w2
		fstar.w1 = P384.w1
		fstar.w0 = P384.w0
	}

	// determine inexactness
	if ind <= 18 { // if 0 <= ind <= 18
		if fstar.w3 > bid_half192[ind] || (fstar.w3 == bid_half192[ind] &&
			(fstar.w2 != 0 || fstar.w1 != 0 || fstar.w0 != 0)) {
			tmp64 = fstar.w3 - bid_half192[ind]
			if tmp64 != 0 || fstar.w2 > bid_ten2mxtrunc192[ind].w2 || (fstar.w2 == bid_ten2mxtrunc192[ind].w2 && fstar.w1 > bid_ten2mxtrunc192[ind].w1) || (fstar.w2 == bid_ten2mxtrunc192[ind].w2 && fstar.w1 == bid_ten2mxtrunc192[ind].w1 && fstar.w0 > bid_ten2mxtrunc192[ind].w0) {
				is_inexact_lt_midpoint = 1
			}
		} else {
			is_inexact_gt_midpoint = 1
		}
	} else if ind <= 37 { // if 19 <= ind <= 37
		if fstar.w4 > bid_half192[ind] || (fstar.w4 == bid_half192[ind] &&
			(fstar.w3 != 0 || fstar.w2 != 0 || fstar.w1 != 0 || fstar.w0 != 0)) {
			tmp64 = fstar.w4 - bid_half192[ind]
			if tmp64 != 0 || fstar.w3 != 0 || fstar.w2 > bid_ten2mxtrunc192[ind].w2 || (fstar.w2 == bid_ten2mxtrunc192[ind].w2 && fstar.w1 > bid_ten2mxtrunc192[ind].w1) || (fstar.w2 == bid_ten2mxtrunc192[ind].w2 && fstar.w1 == bid_ten2mxtrunc192[ind].w1 && fstar.w0 > bid_ten2mxtrunc192[ind].w0) {
				is_inexact_lt_midpoint = 1
			}
		} else {
			is_inexact_gt_midpoint = 1
		}
	} else { // if 38 <= ind <= 55
		if fstar.w5 > bid_half192[ind] || (fstar.w5 == bid_half192[ind] &&
			(fstar.w4 != 0 || fstar.w3 != 0 || fstar.w2 != 0 || fstar.w1 != 0 || fstar.w0 != 0)) {
			tmp64 = fstar.w5 - bid_half192[ind]
			if tmp64 != 0 || fstar.w4 != 0 || fstar.w3 != 0 || fstar.w2 > bid_ten2mxtrunc192[ind].w2 || (fstar.w2 == bid_ten2mxtrunc192[ind].w2 && fstar.w1 > bid_ten2mxtrunc192[ind].w1) || (fstar.w2 == bid_ten2mxtrunc192[ind].w2 && fstar.w1 == bid_ten2mxtrunc192[ind].w1 && fstar.w0 > bid_ten2mxtrunc192[ind].w0) {
				is_inexact_lt_midpoint = 1
			}
		} else {
			is_inexact_gt_midpoint = 1
		}
	}
	// check for midpoints
	if fstar.w5 == 0 && fstar.w4 == 0 && fstar.w3 == 0 &&
		(fstar.w2 < bid_ten2mxtrunc192[ind].w2 ||
			(fstar.w2 == bid_ten2mxtrunc192[ind].w2 &&
				fstar.w1 < bid_ten2mxtrunc192[ind].w1) ||
			(fstar.w2 == bid_ten2mxtrunc192[ind].w2 &&
				fstar.w1 == bid_ten2mxtrunc192[ind].w1 &&
				fstar.w0 <= bid_ten2mxtrunc192[ind].w0)) {
		if Cstar.w0&0x01 != 0 { // Cstar is odd
			Cstar.w0--
			if Cstar.w0 == 0xffffffffffffffff {
				Cstar.w1--
				if Cstar.w1 == 0xffffffffffffffff {
					Cstar.w2--
				}
			}
			is_midpoint_gt_even = 1
			is_inexact_lt_midpoint = 0
			is_inexact_gt_midpoint = 0
		} else {
			is_midpoint_lt_even = 1
			is_inexact_lt_midpoint = 0
			is_inexact_gt_midpoint = 0
		}
	}
	// check for rounding overflow
	ind = q - x
	if ind <= 19 {
		if Cstar.w2 == 0x0 && Cstar.w1 == 0x0 &&
			Cstar.w0 == bid_ten2k64[ind] {
			Cstar.w0 = bid_ten2k64[ind-1]
			incr_exp = 1
		} else {
			incr_exp = 0
		}
	} else if ind == 20 {
		if Cstar.w2 == 0x0 && Cstar.w1 == bid_ten2k128[0].hi &&
			Cstar.w0 == bid_ten2k128[0].lo {
			Cstar.w0 = bid_ten2k64[19]
			Cstar.w1 = 0x0
			incr_exp = 1
		} else {
			incr_exp = 0
		}
	} else if ind <= 38 { // if 21 <= ind <= 38
		if Cstar.w2 == 0x0 && Cstar.w1 == bid_ten2k128[ind-20].hi &&
			Cstar.w0 == bid_ten2k128[ind-20].lo {
			Cstar.w0 = bid_ten2k128[ind-21].lo
			Cstar.w1 = bid_ten2k128[ind-21].hi
			incr_exp = 1
		} else {
			incr_exp = 0
		}
	} else if ind == 39 {
		if Cstar.w2 == bid_ten2k256[0].w2 && Cstar.w1 == bid_ten2k256[0].w1 &&
			Cstar.w0 == bid_ten2k256[0].w0 {
			Cstar.w0 = bid_ten2k128[18].lo
			Cstar.w1 = bid_ten2k128[18].hi
			Cstar.w2 = 0x0
			incr_exp = 1
		} else {
			incr_exp = 0
		}
	} else { // if 40 <= ind <= 56
		if Cstar.w2 == bid_ten2k256[ind-39].w2 &&
			Cstar.w1 == bid_ten2k256[ind-39].w1 &&
			Cstar.w0 == bid_ten2k256[ind-39].w0 {
			Cstar.w0 = bid_ten2k256[ind-40].w0
			Cstar.w1 = bid_ten2k256[ind-40].w1
			Cstar.w2 = bid_ten2k256[ind-40].w2
			incr_exp = 1
		} else {
			incr_exp = 0
		}
	}
	return
}

// bid_round256_58_76 rounds a number C with q decimal digits (58 <= q <= 76)
// to q - x digits (1 <= x <= 75).
func bid_round256_58_76(q int, x int, C BID_UINT256) (Cstar BID_UINT256, incr_exp int, is_midpoint_lt_even int, is_midpoint_gt_even int, is_inexact_lt_midpoint int, is_inexact_gt_midpoint int) {

	var P512 BID_UINT512
	var fstar BID_UINT512
	var tmp64 uint64
	var shift int
	var ind int

	ind = x - 1    // 0 <= ind <= 74
	if ind <= 18 { // if 0 <= ind <= 18
		tmp64 = C.w0
		C.w0 = C.w0 + bid_midpoint64[ind]
		if C.w0 < tmp64 {
			C.w1++
			if C.w1 == 0x0 {
				C.w2++
				if C.w2 == 0x0 {
					C.w3++
				}
			}
		}
	} else if ind <= 37 { // if 19 <= ind <= 37
		tmp64 = C.w0
		C.w0 = C.w0 + bid_midpoint128[ind-19].lo
		if C.w0 < tmp64 {
			C.w1++
			if C.w1 == 0x0 {
				C.w2++
				if C.w2 == 0x0 {
					C.w3++
				}
			}
		}
		tmp64 = C.w1
		C.w1 = C.w1 + bid_midpoint128[ind-19].hi
		if C.w1 < tmp64 {
			C.w2++
			if C.w2 == 0x0 {
				C.w3++
			}
		}
	} else if ind <= 57 { // if 38 <= ind <= 57
		tmp64 = C.w0
		C.w0 = C.w0 + bid_midpoint192[ind-38].w0
		if C.w0 < tmp64 {
			C.w1++
			if C.w1 == 0x0 {
				C.w2++
				if C.w2 == 0x0 {
					C.w3++
				}
			}
		}
		tmp64 = C.w1
		C.w1 = C.w1 + bid_midpoint192[ind-38].w1
		if C.w1 < tmp64 {
			C.w2++
			if C.w2 == 0x0 {
				C.w3++
			}
		}
		tmp64 = C.w2
		C.w2 = C.w2 + bid_midpoint192[ind-38].w2
		if C.w2 < tmp64 {
			C.w3++
		}
	} else { // if 58 <= ind <= 76 (actually 58 <= ind <= 74)
		tmp64 = C.w0
		C.w0 = C.w0 + bid_midpoint256[ind-58].w0
		if C.w0 < tmp64 {
			C.w1++
			if C.w1 == 0x0 {
				C.w2++
				if C.w2 == 0x0 {
					C.w3++
				}
			}
		}
		tmp64 = C.w1
		C.w1 = C.w1 + bid_midpoint256[ind-58].w1
		if C.w1 < tmp64 {
			C.w2++
			if C.w2 == 0x0 {
				C.w3++
			}
		}
		tmp64 = C.w2
		C.w2 = C.w2 + bid_midpoint256[ind-58].w2
		if C.w2 < tmp64 {
			C.w3++
		}
		C.w3 = C.w3 + bid_midpoint256[ind-58].w3
	}
	// P512 = (C + 1/2 * 10^x) * Kx
	P512 = __mul_256x256_to_512(C, bid_Kx256[ind])
	shift = int(bid_Ex256m256[ind])
	if ind <= 18 { // if 0 <= ind <= 18
		Cstar.w3 = (P512.w7 >> uint(shift))
		Cstar.w2 = (P512.w7 << uint(64-shift)) | (P512.w6 >> uint(shift))
		Cstar.w1 = (P512.w6 << uint(64-shift)) | (P512.w5 >> uint(shift))
		Cstar.w0 = (P512.w5 << uint(64-shift)) | (P512.w4 >> uint(shift))
		fstar.w7 = 0x0
		fstar.w6 = 0x0
		fstar.w5 = 0x0
		fstar.w4 = P512.w4 & bid_mask256[ind]
		fstar.w3 = P512.w3
		fstar.w2 = P512.w2
		fstar.w1 = P512.w1
		fstar.w0 = P512.w0
	} else if ind <= 37 { // if 19 <= ind <= 37
		Cstar.w3 = 0x0
		Cstar.w2 = P512.w7 >> uint(shift)
		Cstar.w1 = (P512.w7 << uint(64-shift)) | (P512.w6 >> uint(shift))
		Cstar.w0 = (P512.w6 << uint(64-shift)) | (P512.w5 >> uint(shift))
		fstar.w7 = 0x0
		fstar.w6 = 0x0
		fstar.w5 = P512.w5 & bid_mask256[ind]
		fstar.w4 = P512.w4
		fstar.w3 = P512.w3
		fstar.w2 = P512.w2
		fstar.w1 = P512.w1
		fstar.w0 = P512.w0
	} else if ind <= 56 { // if 38 <= ind <= 56
		Cstar.w3 = 0x0
		Cstar.w2 = 0x0
		Cstar.w1 = P512.w7 >> uint(shift)
		Cstar.w0 = (P512.w7 << uint(64-shift)) | (P512.w6 >> uint(shift))
		fstar.w7 = 0x0
		fstar.w6 = P512.w6 & bid_mask256[ind]
		fstar.w5 = P512.w5
		fstar.w4 = P512.w4
		fstar.w3 = P512.w3
		fstar.w2 = P512.w2
		fstar.w1 = P512.w1
		fstar.w0 = P512.w0
	} else if ind == 57 {
		Cstar.w3 = 0x0
		Cstar.w2 = 0x0
		Cstar.w1 = 0x0
		Cstar.w0 = P512.w7
		fstar.w7 = 0x0
		fstar.w6 = P512.w6
		fstar.w5 = P512.w5
		fstar.w4 = P512.w4
		fstar.w3 = P512.w3
		fstar.w2 = P512.w2
		fstar.w1 = P512.w1
		fstar.w0 = P512.w0
	} else { // if 58 <= ind <= 74
		Cstar.w3 = 0x0
		Cstar.w2 = 0x0
		Cstar.w1 = 0x0
		Cstar.w0 = P512.w7 >> uint(shift)
		fstar.w7 = P512.w7 & bid_mask256[ind]
		fstar.w6 = P512.w6
		fstar.w5 = P512.w5
		fstar.w4 = P512.w4
		fstar.w3 = P512.w3
		fstar.w2 = P512.w2
		fstar.w1 = P512.w1
		fstar.w0 = P512.w0
	}

	// determine inexactness
	if ind <= 18 { // if 0 <= ind <= 18
		if fstar.w4 > bid_half256[ind] || (fstar.w4 == bid_half256[ind] &&
			(fstar.w3 != 0 || fstar.w2 != 0 || fstar.w1 != 0 || fstar.w0 != 0)) {
			tmp64 = fstar.w4 - bid_half256[ind]
			if tmp64 != 0 || fstar.w3 > bid_ten2mxtrunc256[ind].w2 || (fstar.w3 == bid_ten2mxtrunc256[ind].w3 && fstar.w2 > bid_ten2mxtrunc256[ind].w2) || (fstar.w3 == bid_ten2mxtrunc256[ind].w3 && fstar.w2 == bid_ten2mxtrunc256[ind].w2 && fstar.w1 > bid_ten2mxtrunc256[ind].w1) || (fstar.w3 == bid_ten2mxtrunc256[ind].w3 && fstar.w2 == bid_ten2mxtrunc256[ind].w2 && fstar.w1 == bid_ten2mxtrunc256[ind].w1 && fstar.w0 > bid_ten2mxtrunc256[ind].w0) {
				is_inexact_lt_midpoint = 1
			}
		} else {
			is_inexact_gt_midpoint = 1
		}
	} else if ind <= 37 { // if 19 <= ind <= 37
		if fstar.w5 > bid_half256[ind] || (fstar.w5 == bid_half256[ind] &&
			(fstar.w4 != 0 || fstar.w3 != 0 || fstar.w2 != 0 || fstar.w1 != 0 || fstar.w0 != 0)) {
			tmp64 = fstar.w5 - bid_half256[ind]
			if tmp64 != 0 || fstar.w4 != 0 || fstar.w3 > bid_ten2mxtrunc256[ind].w3 || (fstar.w3 == bid_ten2mxtrunc256[ind].w3 && fstar.w2 > bid_ten2mxtrunc256[ind].w2) || (fstar.w3 == bid_ten2mxtrunc256[ind].w3 && fstar.w2 == bid_ten2mxtrunc256[ind].w2 && fstar.w1 > bid_ten2mxtrunc256[ind].w1) || (fstar.w3 == bid_ten2mxtrunc256[ind].w3 && fstar.w2 == bid_ten2mxtrunc256[ind].w2 && fstar.w1 == bid_ten2mxtrunc256[ind].w1 && fstar.w0 > bid_ten2mxtrunc256[ind].w0) {
				is_inexact_lt_midpoint = 1
			}
		} else {
			is_inexact_gt_midpoint = 1
		}
	} else if ind <= 57 { // if 38 <= ind <= 57
		if fstar.w6 > bid_half256[ind] || (fstar.w6 == bid_half256[ind] &&
			(fstar.w5 != 0 || fstar.w4 != 0 || fstar.w3 != 0 || fstar.w2 != 0 || fstar.w1 != 0 || fstar.w0 != 0)) {
			tmp64 = fstar.w6 - bid_half256[ind]
			if tmp64 != 0 || fstar.w5 != 0 || fstar.w4 != 0 || fstar.w3 > bid_ten2mxtrunc256[ind].w3 || (fstar.w3 == bid_ten2mxtrunc256[ind].w3 && fstar.w2 > bid_ten2mxtrunc256[ind].w2) || (fstar.w3 == bid_ten2mxtrunc256[ind].w3 && fstar.w2 == bid_ten2mxtrunc256[ind].w2 && fstar.w1 > bid_ten2mxtrunc256[ind].w1) || (fstar.w3 == bid_ten2mxtrunc256[ind].w3 && fstar.w2 == bid_ten2mxtrunc256[ind].w2 && fstar.w1 == bid_ten2mxtrunc256[ind].w1 && fstar.w0 > bid_ten2mxtrunc256[ind].w0) {
				is_inexact_lt_midpoint = 1
			}
		} else {
			is_inexact_gt_midpoint = 1
		}
	} else { // if 58 <= ind <= 74
		if fstar.w7 > bid_half256[ind] || (fstar.w7 == bid_half256[ind] &&
			(fstar.w6 != 0 || fstar.w5 != 0 || fstar.w4 != 0 || fstar.w3 != 0 || fstar.w2 != 0 || fstar.w1 != 0 || fstar.w0 != 0)) {
			tmp64 = fstar.w7 - bid_half256[ind]
			if tmp64 != 0 || fstar.w6 != 0 || fstar.w5 != 0 || fstar.w4 != 0 || fstar.w3 > bid_ten2mxtrunc256[ind].w3 || (fstar.w3 == bid_ten2mxtrunc256[ind].w3 && fstar.w2 > bid_ten2mxtrunc256[ind].w2) || (fstar.w3 == bid_ten2mxtrunc256[ind].w3 && fstar.w2 == bid_ten2mxtrunc256[ind].w2 && fstar.w1 > bid_ten2mxtrunc256[ind].w1) || (fstar.w3 == bid_ten2mxtrunc256[ind].w3 && fstar.w2 == bid_ten2mxtrunc256[ind].w2 && fstar.w1 == bid_ten2mxtrunc256[ind].w1 && fstar.w0 > bid_ten2mxtrunc256[ind].w0) {
				is_inexact_lt_midpoint = 1
			}
		} else {
			is_inexact_gt_midpoint = 1
		}
	}
	// check for midpoints
	if fstar.w7 == 0 && fstar.w6 == 0 &&
		fstar.w5 == 0 && fstar.w4 == 0 &&
		(fstar.w3 < bid_ten2mxtrunc256[ind].w3 ||
			(fstar.w3 == bid_ten2mxtrunc256[ind].w3 &&
				fstar.w2 < bid_ten2mxtrunc256[ind].w2) ||
			(fstar.w3 == bid_ten2mxtrunc256[ind].w3 &&
				fstar.w2 == bid_ten2mxtrunc256[ind].w2 &&
				fstar.w1 < bid_ten2mxtrunc256[ind].w1) ||
			(fstar.w3 == bid_ten2mxtrunc256[ind].w3 &&
				fstar.w2 == bid_ten2mxtrunc256[ind].w2 &&
				fstar.w1 == bid_ten2mxtrunc256[ind].w1 &&
				fstar.w0 <= bid_ten2mxtrunc256[ind].w0)) {
		if Cstar.w0&0x01 != 0 { // Cstar is odd
			Cstar.w0--
			if Cstar.w0 == 0xffffffffffffffff {
				Cstar.w1--
				if Cstar.w1 == 0xffffffffffffffff {
					Cstar.w2--
					if Cstar.w2 == 0xffffffffffffffff {
						Cstar.w3--
					}
				}
			}
			is_midpoint_gt_even = 1
			is_inexact_lt_midpoint = 0
			is_inexact_gt_midpoint = 0
		} else {
			is_midpoint_lt_even = 1
			is_inexact_lt_midpoint = 0
			is_inexact_gt_midpoint = 0
		}
	}
	// check for rounding overflow
	ind = q - x
	if ind <= 19 {
		if Cstar.w3 == 0x0 && Cstar.w2 == 0x0 &&
			Cstar.w1 == 0x0 && Cstar.w0 == bid_ten2k64[ind] {
			Cstar.w0 = bid_ten2k64[ind-1]
			incr_exp = 1
		} else {
			incr_exp = 0
		}
	} else if ind == 20 {
		if Cstar.w3 == 0x0 && Cstar.w2 == 0x0 &&
			Cstar.w1 == bid_ten2k128[0].hi &&
			Cstar.w0 == bid_ten2k128[0].lo {
			Cstar.w0 = bid_ten2k64[19]
			Cstar.w1 = 0x0
			incr_exp = 1
		} else {
			incr_exp = 0
		}
	} else if ind <= 38 { // if 21 <= ind <= 38
		if Cstar.w3 == 0x0 && Cstar.w2 == 0x0 &&
			Cstar.w1 == bid_ten2k128[ind-20].hi &&
			Cstar.w0 == bid_ten2k128[ind-20].lo {
			Cstar.w0 = bid_ten2k128[ind-21].lo
			Cstar.w1 = bid_ten2k128[ind-21].hi
			incr_exp = 1
		} else {
			incr_exp = 0
		}
	} else if ind == 39 {
		if Cstar.w3 == 0x0 && Cstar.w2 == bid_ten2k256[0].w2 &&
			Cstar.w1 == bid_ten2k256[0].w1 &&
			Cstar.w0 == bid_ten2k256[0].w0 {
			Cstar.w0 = bid_ten2k128[18].lo
			Cstar.w1 = bid_ten2k128[18].hi
			Cstar.w2 = 0x0
			incr_exp = 1
		} else {
			incr_exp = 0
		}
	} else if ind <= 57 { // if 40 <= ind <= 57
		if Cstar.w3 == 0x0 && Cstar.w2 == bid_ten2k256[ind-39].w2 &&
			Cstar.w1 == bid_ten2k256[ind-39].w1 &&
			Cstar.w0 == bid_ten2k256[ind-39].w0 {
			Cstar.w0 = bid_ten2k256[ind-40].w0
			Cstar.w1 = bid_ten2k256[ind-40].w1
			Cstar.w2 = bid_ten2k256[ind-40].w2
			incr_exp = 1
		} else {
			incr_exp = 0
		}
	} else { // if 58 <= ind <= 77 (actually 58 <= ind <= 74)
		if Cstar.w3 == bid_ten2k256[ind-39].w3 &&
			Cstar.w2 == bid_ten2k256[ind-39].w2 &&
			Cstar.w1 == bid_ten2k256[ind-39].w1 &&
			Cstar.w0 == bid_ten2k256[ind-39].w0 {
			Cstar.w0 = bid_ten2k256[ind-40].w0
			Cstar.w1 = bid_ten2k256[ind-40].w1
			Cstar.w2 = bid_ten2k256[ind-40].w2
			Cstar.w3 = bid_ten2k256[ind-40].w3
			incr_exp = 1
		} else {
			incr_exp = 0
		}
	}
	return
}
