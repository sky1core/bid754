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
		tmp64 = C.w[0]
		C.w[0] = C.w[0] + bid_midpoint64[ind]
		if C.w[0] < tmp64 {
			C.w[1]++
		}
	} else { // if 19 <= ind <= 37
		tmp64 = C.w[0]
		C.w[0] = C.w[0] + bid_midpoint128[ind-19].w[0]
		if C.w[0] < tmp64 {
			C.w[1]++
		}
		C.w[1] = C.w[1] + bid_midpoint128[ind-19].w[1]
	}
	// kx ~= 10^(-x), kx = bid_Kx128[ind] * 2^(-Ex), 0 <= ind <= 36
	// P256 = (C + 1/2 * 10^x) * kx * 2^Ex = (C + 1/2 * 10^x) * Kx
	P256 = __mul_128x128_to_256(C, bid_Kx128[ind])
	// calculate C* = floor (P256) and f*
	// Cstar = P256 >> Ex
	// fstar = low Ex bits of P256
	shift = int(bid_Ex128m128[ind]) // in [2, 63]
	if ind <= 18 {                  // if 0 <= ind <= 18
		Cstar.w[0] = (P256.w[2] >> uint(shift)) | (P256.w[3] << uint(64-shift))
		Cstar.w[1] = (P256.w[3] >> uint(shift))
		fstar.w[0] = P256.w[0]
		fstar.w[1] = P256.w[1]
		fstar.w[2] = P256.w[2] & bid_mask128[ind]
		fstar.w[3] = 0x0
	} else { // if 19 <= ind <= 37
		Cstar.w[0] = P256.w[3] >> uint(shift)
		Cstar.w[1] = 0x0
		fstar.w[0] = P256.w[0]
		fstar.w[1] = P256.w[1]
		fstar.w[2] = P256.w[2]
		fstar.w[3] = P256.w[3] & bid_mask128[ind]
	}

	// determine inexactness of the rounding of C*
	if ind <= 18 { // if 0 <= ind <= 18
		if fstar.w[2] > bid_half128[ind] ||
			(fstar.w[2] == bid_half128[ind] && (fstar.w[1] != 0 || fstar.w[0] != 0)) {
			// f* > 1/2 and the result may be exact
			tmp64 = fstar.w[2] - bid_half128[ind]
			if tmp64 != 0 || fstar.w[1] > bid_ten2mxtrunc128[ind].w[1] || (fstar.w[1] == bid_ten2mxtrunc128[ind].w[1] && fstar.w[0] > bid_ten2mxtrunc128[ind].w[0]) { // f* - 1/2 > 10^(-x)
				is_inexact_lt_midpoint = 1
			} // else the result is exact
		} else { // the result is inexact; f2* <= 1/2
			is_inexact_gt_midpoint = 1
		}
	} else { // if 19 <= ind <= 37
		if fstar.w[3] > bid_half128[ind] || (fstar.w[3] == bid_half128[ind] &&
			(fstar.w[2] != 0 || fstar.w[1] != 0 || fstar.w[0] != 0)) {
			// f* > 1/2 and the result may be exact
			tmp64 = fstar.w[3] - bid_half128[ind]
			if tmp64 != 0 || fstar.w[2] != 0 || fstar.w[1] > bid_ten2mxtrunc128[ind].w[1] || (fstar.w[1] == bid_ten2mxtrunc128[ind].w[1] && fstar.w[0] > bid_ten2mxtrunc128[ind].w[0]) { // f* - 1/2 > 10^(-x)
				is_inexact_lt_midpoint = 1
			} // else the result is exact
		} else { // the result is inexact; f2* <= 1/2
			is_inexact_gt_midpoint = 1
		}
	}
	// check for midpoints
	if fstar.w[3] == 0 && fstar.w[2] == 0 &&
		(fstar.w[1] < bid_ten2mxtrunc128[ind].w[1] ||
			(fstar.w[1] == bid_ten2mxtrunc128[ind].w[1] &&
				fstar.w[0] <= bid_ten2mxtrunc128[ind].w[0])) {
		// the result is a midpoint
		if Cstar.w[0]&0x01 != 0 { // Cstar is odd; MP in [EVEN, ODD]
			Cstar.w[0]-- // Cstar is now even
			if Cstar.w[0] == 0xffffffffffffffff {
				Cstar.w[1]--
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
		if Cstar.w[1] == 0x0 && Cstar.w[0] == bid_ten2k64[ind] {
			Cstar.w[0] = bid_ten2k64[ind-1]
			incr_exp = 1
		} else {
			incr_exp = 0
		}
	} else if ind == 20 {
		if Cstar.w[1] == bid_ten2k128[0].w[1] &&
			Cstar.w[0] == bid_ten2k128[0].w[0] {
			Cstar.w[0] = bid_ten2k64[19]
			Cstar.w[1] = 0x0
			incr_exp = 1
		} else {
			incr_exp = 0
		}
	} else { // if 21 <= ind <= 37
		if Cstar.w[1] == bid_ten2k128[ind-20].w[1] &&
			Cstar.w[0] == bid_ten2k128[ind-20].w[0] {
			Cstar.w[0] = bid_ten2k128[ind-21].w[0]
			Cstar.w[1] = bid_ten2k128[ind-21].w[1]
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
		tmp64 = C.w[0]
		C.w[0] = C.w[0] + bid_midpoint64[ind]
		if C.w[0] < tmp64 {
			C.w[1]++
			if C.w[1] == 0x0 {
				C.w[2]++
			}
		}
	} else if ind <= 37 { // if 19 <= ind <= 37
		tmp64 = C.w[0]
		C.w[0] = C.w[0] + bid_midpoint128[ind-19].w[0]
		if C.w[0] < tmp64 {
			C.w[1]++
			if C.w[1] == 0x0 {
				C.w[2]++
			}
		}
		tmp64 = C.w[1]
		C.w[1] = C.w[1] + bid_midpoint128[ind-19].w[1]
		if C.w[1] < tmp64 {
			C.w[2]++
		}
	} else { // if 38 <= ind <= 57 (actually ind <= 55)
		tmp64 = C.w[0]
		C.w[0] = C.w[0] + bid_midpoint192[ind-38].w[0]
		if C.w[0] < tmp64 {
			C.w[1]++
			if C.w[1] == 0x0 {
				C.w[2]++
			}
		}
		tmp64 = C.w[1]
		C.w[1] = C.w[1] + bid_midpoint192[ind-38].w[1]
		if C.w[1] < tmp64 {
			C.w[2]++
		}
		C.w[2] = C.w[2] + bid_midpoint192[ind-38].w[2]
	}
	// P384 = (C + 1/2 * 10^x) * Kx
	P384 = __mul_192x192_to_384(C, bid_Kx192[ind])
	shift = int(bid_Ex192m192[ind])
	if ind <= 18 { // if 0 <= ind <= 18
		Cstar.w[2] = (P384.w[5] >> uint(shift))
		Cstar.w[1] = (P384.w[5] << uint(64-shift)) | (P384.w[4] >> uint(shift))
		Cstar.w[0] = (P384.w[4] << uint(64-shift)) | (P384.w[3] >> uint(shift))
		fstar.w[5] = 0x0
		fstar.w[4] = 0x0
		fstar.w[3] = P384.w[3] & bid_mask192[ind]
		fstar.w[2] = P384.w[2]
		fstar.w[1] = P384.w[1]
		fstar.w[0] = P384.w[0]
	} else if ind <= 37 { // if 19 <= ind <= 37
		Cstar.w[2] = 0x0
		Cstar.w[1] = P384.w[5] >> uint(shift)
		Cstar.w[0] = (P384.w[5] << uint(64-shift)) | (P384.w[4] >> uint(shift))
		fstar.w[5] = 0x0
		fstar.w[4] = P384.w[4] & bid_mask192[ind]
		fstar.w[3] = P384.w[3]
		fstar.w[2] = P384.w[2]
		fstar.w[1] = P384.w[1]
		fstar.w[0] = P384.w[0]
	} else { // if 38 <= ind <= 57
		Cstar.w[2] = 0x0
		Cstar.w[1] = 0x0
		Cstar.w[0] = P384.w[5] >> uint(shift)
		fstar.w[5] = P384.w[5] & bid_mask192[ind]
		fstar.w[4] = P384.w[4]
		fstar.w[3] = P384.w[3]
		fstar.w[2] = P384.w[2]
		fstar.w[1] = P384.w[1]
		fstar.w[0] = P384.w[0]
	}

	// determine inexactness
	if ind <= 18 { // if 0 <= ind <= 18
		if fstar.w[3] > bid_half192[ind] || (fstar.w[3] == bid_half192[ind] &&
			(fstar.w[2] != 0 || fstar.w[1] != 0 || fstar.w[0] != 0)) {
			tmp64 = fstar.w[3] - bid_half192[ind]
			if tmp64 != 0 || fstar.w[2] > bid_ten2mxtrunc192[ind].w[2] || (fstar.w[2] == bid_ten2mxtrunc192[ind].w[2] && fstar.w[1] > bid_ten2mxtrunc192[ind].w[1]) || (fstar.w[2] == bid_ten2mxtrunc192[ind].w[2] && fstar.w[1] == bid_ten2mxtrunc192[ind].w[1] && fstar.w[0] > bid_ten2mxtrunc192[ind].w[0]) {
				is_inexact_lt_midpoint = 1
			}
		} else {
			is_inexact_gt_midpoint = 1
		}
	} else if ind <= 37 { // if 19 <= ind <= 37
		if fstar.w[4] > bid_half192[ind] || (fstar.w[4] == bid_half192[ind] &&
			(fstar.w[3] != 0 || fstar.w[2] != 0 || fstar.w[1] != 0 || fstar.w[0] != 0)) {
			tmp64 = fstar.w[4] - bid_half192[ind]
			if tmp64 != 0 || fstar.w[3] != 0 || fstar.w[2] > bid_ten2mxtrunc192[ind].w[2] || (fstar.w[2] == bid_ten2mxtrunc192[ind].w[2] && fstar.w[1] > bid_ten2mxtrunc192[ind].w[1]) || (fstar.w[2] == bid_ten2mxtrunc192[ind].w[2] && fstar.w[1] == bid_ten2mxtrunc192[ind].w[1] && fstar.w[0] > bid_ten2mxtrunc192[ind].w[0]) {
				is_inexact_lt_midpoint = 1
			}
		} else {
			is_inexact_gt_midpoint = 1
		}
	} else { // if 38 <= ind <= 55
		if fstar.w[5] > bid_half192[ind] || (fstar.w[5] == bid_half192[ind] &&
			(fstar.w[4] != 0 || fstar.w[3] != 0 || fstar.w[2] != 0 || fstar.w[1] != 0 || fstar.w[0] != 0)) {
			tmp64 = fstar.w[5] - bid_half192[ind]
			if tmp64 != 0 || fstar.w[4] != 0 || fstar.w[3] != 0 || fstar.w[2] > bid_ten2mxtrunc192[ind].w[2] || (fstar.w[2] == bid_ten2mxtrunc192[ind].w[2] && fstar.w[1] > bid_ten2mxtrunc192[ind].w[1]) || (fstar.w[2] == bid_ten2mxtrunc192[ind].w[2] && fstar.w[1] == bid_ten2mxtrunc192[ind].w[1] && fstar.w[0] > bid_ten2mxtrunc192[ind].w[0]) {
				is_inexact_lt_midpoint = 1
			}
		} else {
			is_inexact_gt_midpoint = 1
		}
	}
	// check for midpoints
	if fstar.w[5] == 0 && fstar.w[4] == 0 && fstar.w[3] == 0 &&
		(fstar.w[2] < bid_ten2mxtrunc192[ind].w[2] ||
			(fstar.w[2] == bid_ten2mxtrunc192[ind].w[2] &&
				fstar.w[1] < bid_ten2mxtrunc192[ind].w[1]) ||
			(fstar.w[2] == bid_ten2mxtrunc192[ind].w[2] &&
				fstar.w[1] == bid_ten2mxtrunc192[ind].w[1] &&
				fstar.w[0] <= bid_ten2mxtrunc192[ind].w[0])) {
		if Cstar.w[0]&0x01 != 0 { // Cstar is odd
			Cstar.w[0]--
			if Cstar.w[0] == 0xffffffffffffffff {
				Cstar.w[1]--
				if Cstar.w[1] == 0xffffffffffffffff {
					Cstar.w[2]--
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
		if Cstar.w[2] == 0x0 && Cstar.w[1] == 0x0 &&
			Cstar.w[0] == bid_ten2k64[ind] {
			Cstar.w[0] = bid_ten2k64[ind-1]
			incr_exp = 1
		} else {
			incr_exp = 0
		}
	} else if ind == 20 {
		if Cstar.w[2] == 0x0 && Cstar.w[1] == bid_ten2k128[0].w[1] &&
			Cstar.w[0] == bid_ten2k128[0].w[0] {
			Cstar.w[0] = bid_ten2k64[19]
			Cstar.w[1] = 0x0
			incr_exp = 1
		} else {
			incr_exp = 0
		}
	} else if ind <= 38 { // if 21 <= ind <= 38
		if Cstar.w[2] == 0x0 && Cstar.w[1] == bid_ten2k128[ind-20].w[1] &&
			Cstar.w[0] == bid_ten2k128[ind-20].w[0] {
			Cstar.w[0] = bid_ten2k128[ind-21].w[0]
			Cstar.w[1] = bid_ten2k128[ind-21].w[1]
			incr_exp = 1
		} else {
			incr_exp = 0
		}
	} else if ind == 39 {
		if Cstar.w[2] == bid_ten2k256[0].w[2] && Cstar.w[1] == bid_ten2k256[0].w[1] &&
			Cstar.w[0] == bid_ten2k256[0].w[0] {
			Cstar.w[0] = bid_ten2k128[18].w[0]
			Cstar.w[1] = bid_ten2k128[18].w[1]
			Cstar.w[2] = 0x0
			incr_exp = 1
		} else {
			incr_exp = 0
		}
	} else { // if 40 <= ind <= 56
		if Cstar.w[2] == bid_ten2k256[ind-39].w[2] &&
			Cstar.w[1] == bid_ten2k256[ind-39].w[1] &&
			Cstar.w[0] == bid_ten2k256[ind-39].w[0] {
			Cstar.w[0] = bid_ten2k256[ind-40].w[0]
			Cstar.w[1] = bid_ten2k256[ind-40].w[1]
			Cstar.w[2] = bid_ten2k256[ind-40].w[2]
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
		tmp64 = C.w[0]
		C.w[0] = C.w[0] + bid_midpoint64[ind]
		if C.w[0] < tmp64 {
			C.w[1]++
			if C.w[1] == 0x0 {
				C.w[2]++
				if C.w[2] == 0x0 {
					C.w[3]++
				}
			}
		}
	} else if ind <= 37 { // if 19 <= ind <= 37
		tmp64 = C.w[0]
		C.w[0] = C.w[0] + bid_midpoint128[ind-19].w[0]
		if C.w[0] < tmp64 {
			C.w[1]++
			if C.w[1] == 0x0 {
				C.w[2]++
				if C.w[2] == 0x0 {
					C.w[3]++
				}
			}
		}
		tmp64 = C.w[1]
		C.w[1] = C.w[1] + bid_midpoint128[ind-19].w[1]
		if C.w[1] < tmp64 {
			C.w[2]++
			if C.w[2] == 0x0 {
				C.w[3]++
			}
		}
	} else if ind <= 57 { // if 38 <= ind <= 57
		tmp64 = C.w[0]
		C.w[0] = C.w[0] + bid_midpoint192[ind-38].w[0]
		if C.w[0] < tmp64 {
			C.w[1]++
			if C.w[1] == 0x0 {
				C.w[2]++
				if C.w[2] == 0x0 {
					C.w[3]++
				}
			}
		}
		tmp64 = C.w[1]
		C.w[1] = C.w[1] + bid_midpoint192[ind-38].w[1]
		if C.w[1] < tmp64 {
			C.w[2]++
			if C.w[2] == 0x0 {
				C.w[3]++
			}
		}
		tmp64 = C.w[2]
		C.w[2] = C.w[2] + bid_midpoint192[ind-38].w[2]
		if C.w[2] < tmp64 {
			C.w[3]++
		}
	} else { // if 58 <= ind <= 76 (actually 58 <= ind <= 74)
		tmp64 = C.w[0]
		C.w[0] = C.w[0] + bid_midpoint256[ind-58].w[0]
		if C.w[0] < tmp64 {
			C.w[1]++
			if C.w[1] == 0x0 {
				C.w[2]++
				if C.w[2] == 0x0 {
					C.w[3]++
				}
			}
		}
		tmp64 = C.w[1]
		C.w[1] = C.w[1] + bid_midpoint256[ind-58].w[1]
		if C.w[1] < tmp64 {
			C.w[2]++
			if C.w[2] == 0x0 {
				C.w[3]++
			}
		}
		tmp64 = C.w[2]
		C.w[2] = C.w[2] + bid_midpoint256[ind-58].w[2]
		if C.w[2] < tmp64 {
			C.w[3]++
		}
		C.w[3] = C.w[3] + bid_midpoint256[ind-58].w[3]
	}
	// P512 = (C + 1/2 * 10^x) * Kx
	P512 = __mul_256x256_to_512(C, bid_Kx256[ind])
	shift = int(bid_Ex256m256[ind])
	if ind <= 18 { // if 0 <= ind <= 18
		Cstar.w[3] = (P512.w[7] >> uint(shift))
		Cstar.w[2] = (P512.w[7] << uint(64-shift)) | (P512.w[6] >> uint(shift))
		Cstar.w[1] = (P512.w[6] << uint(64-shift)) | (P512.w[5] >> uint(shift))
		Cstar.w[0] = (P512.w[5] << uint(64-shift)) | (P512.w[4] >> uint(shift))
		fstar.w[7] = 0x0
		fstar.w[6] = 0x0
		fstar.w[5] = 0x0
		fstar.w[4] = P512.w[4] & bid_mask256[ind]
		fstar.w[3] = P512.w[3]
		fstar.w[2] = P512.w[2]
		fstar.w[1] = P512.w[1]
		fstar.w[0] = P512.w[0]
	} else if ind <= 37 { // if 19 <= ind <= 37
		Cstar.w[3] = 0x0
		Cstar.w[2] = P512.w[7] >> uint(shift)
		Cstar.w[1] = (P512.w[7] << uint(64-shift)) | (P512.w[6] >> uint(shift))
		Cstar.w[0] = (P512.w[6] << uint(64-shift)) | (P512.w[5] >> uint(shift))
		fstar.w[7] = 0x0
		fstar.w[6] = 0x0
		fstar.w[5] = P512.w[5] & bid_mask256[ind]
		fstar.w[4] = P512.w[4]
		fstar.w[3] = P512.w[3]
		fstar.w[2] = P512.w[2]
		fstar.w[1] = P512.w[1]
		fstar.w[0] = P512.w[0]
	} else if ind <= 56 { // if 38 <= ind <= 56
		Cstar.w[3] = 0x0
		Cstar.w[2] = 0x0
		Cstar.w[1] = P512.w[7] >> uint(shift)
		Cstar.w[0] = (P512.w[7] << uint(64-shift)) | (P512.w[6] >> uint(shift))
		fstar.w[7] = 0x0
		fstar.w[6] = P512.w[6] & bid_mask256[ind]
		fstar.w[5] = P512.w[5]
		fstar.w[4] = P512.w[4]
		fstar.w[3] = P512.w[3]
		fstar.w[2] = P512.w[2]
		fstar.w[1] = P512.w[1]
		fstar.w[0] = P512.w[0]
	} else if ind == 57 {
		Cstar.w[3] = 0x0
		Cstar.w[2] = 0x0
		Cstar.w[1] = 0x0
		Cstar.w[0] = P512.w[7]
		fstar.w[7] = 0x0
		fstar.w[6] = P512.w[6]
		fstar.w[5] = P512.w[5]
		fstar.w[4] = P512.w[4]
		fstar.w[3] = P512.w[3]
		fstar.w[2] = P512.w[2]
		fstar.w[1] = P512.w[1]
		fstar.w[0] = P512.w[0]
	} else { // if 58 <= ind <= 74
		Cstar.w[3] = 0x0
		Cstar.w[2] = 0x0
		Cstar.w[1] = 0x0
		Cstar.w[0] = P512.w[7] >> uint(shift)
		fstar.w[7] = P512.w[7] & bid_mask256[ind]
		fstar.w[6] = P512.w[6]
		fstar.w[5] = P512.w[5]
		fstar.w[4] = P512.w[4]
		fstar.w[3] = P512.w[3]
		fstar.w[2] = P512.w[2]
		fstar.w[1] = P512.w[1]
		fstar.w[0] = P512.w[0]
	}

	// determine inexactness
	if ind <= 18 { // if 0 <= ind <= 18
		if fstar.w[4] > bid_half256[ind] || (fstar.w[4] == bid_half256[ind] &&
			(fstar.w[3] != 0 || fstar.w[2] != 0 || fstar.w[1] != 0 || fstar.w[0] != 0)) {
			tmp64 = fstar.w[4] - bid_half256[ind]
			if tmp64 != 0 || fstar.w[3] > bid_ten2mxtrunc256[ind].w[2] || (fstar.w[3] == bid_ten2mxtrunc256[ind].w[3] && fstar.w[2] > bid_ten2mxtrunc256[ind].w[2]) || (fstar.w[3] == bid_ten2mxtrunc256[ind].w[3] && fstar.w[2] == bid_ten2mxtrunc256[ind].w[2] && fstar.w[1] > bid_ten2mxtrunc256[ind].w[1]) || (fstar.w[3] == bid_ten2mxtrunc256[ind].w[3] && fstar.w[2] == bid_ten2mxtrunc256[ind].w[2] && fstar.w[1] == bid_ten2mxtrunc256[ind].w[1] && fstar.w[0] > bid_ten2mxtrunc256[ind].w[0]) {
				is_inexact_lt_midpoint = 1
			}
		} else {
			is_inexact_gt_midpoint = 1
		}
	} else if ind <= 37 { // if 19 <= ind <= 37
		if fstar.w[5] > bid_half256[ind] || (fstar.w[5] == bid_half256[ind] &&
			(fstar.w[4] != 0 || fstar.w[3] != 0 || fstar.w[2] != 0 || fstar.w[1] != 0 || fstar.w[0] != 0)) {
			tmp64 = fstar.w[5] - bid_half256[ind]
			if tmp64 != 0 || fstar.w[4] != 0 || fstar.w[3] > bid_ten2mxtrunc256[ind].w[3] || (fstar.w[3] == bid_ten2mxtrunc256[ind].w[3] && fstar.w[2] > bid_ten2mxtrunc256[ind].w[2]) || (fstar.w[3] == bid_ten2mxtrunc256[ind].w[3] && fstar.w[2] == bid_ten2mxtrunc256[ind].w[2] && fstar.w[1] > bid_ten2mxtrunc256[ind].w[1]) || (fstar.w[3] == bid_ten2mxtrunc256[ind].w[3] && fstar.w[2] == bid_ten2mxtrunc256[ind].w[2] && fstar.w[1] == bid_ten2mxtrunc256[ind].w[1] && fstar.w[0] > bid_ten2mxtrunc256[ind].w[0]) {
				is_inexact_lt_midpoint = 1
			}
		} else {
			is_inexact_gt_midpoint = 1
		}
	} else if ind <= 57 { // if 38 <= ind <= 57
		if fstar.w[6] > bid_half256[ind] || (fstar.w[6] == bid_half256[ind] &&
			(fstar.w[5] != 0 || fstar.w[4] != 0 || fstar.w[3] != 0 || fstar.w[2] != 0 || fstar.w[1] != 0 || fstar.w[0] != 0)) {
			tmp64 = fstar.w[6] - bid_half256[ind]
			if tmp64 != 0 || fstar.w[5] != 0 || fstar.w[4] != 0 || fstar.w[3] > bid_ten2mxtrunc256[ind].w[3] || (fstar.w[3] == bid_ten2mxtrunc256[ind].w[3] && fstar.w[2] > bid_ten2mxtrunc256[ind].w[2]) || (fstar.w[3] == bid_ten2mxtrunc256[ind].w[3] && fstar.w[2] == bid_ten2mxtrunc256[ind].w[2] && fstar.w[1] > bid_ten2mxtrunc256[ind].w[1]) || (fstar.w[3] == bid_ten2mxtrunc256[ind].w[3] && fstar.w[2] == bid_ten2mxtrunc256[ind].w[2] && fstar.w[1] == bid_ten2mxtrunc256[ind].w[1] && fstar.w[0] > bid_ten2mxtrunc256[ind].w[0]) {
				is_inexact_lt_midpoint = 1
			}
		} else {
			is_inexact_gt_midpoint = 1
		}
	} else { // if 58 <= ind <= 74
		if fstar.w[7] > bid_half256[ind] || (fstar.w[7] == bid_half256[ind] &&
			(fstar.w[6] != 0 || fstar.w[5] != 0 || fstar.w[4] != 0 || fstar.w[3] != 0 || fstar.w[2] != 0 || fstar.w[1] != 0 || fstar.w[0] != 0)) {
			tmp64 = fstar.w[7] - bid_half256[ind]
			if tmp64 != 0 || fstar.w[6] != 0 || fstar.w[5] != 0 || fstar.w[4] != 0 || fstar.w[3] > bid_ten2mxtrunc256[ind].w[3] || (fstar.w[3] == bid_ten2mxtrunc256[ind].w[3] && fstar.w[2] > bid_ten2mxtrunc256[ind].w[2]) || (fstar.w[3] == bid_ten2mxtrunc256[ind].w[3] && fstar.w[2] == bid_ten2mxtrunc256[ind].w[2] && fstar.w[1] > bid_ten2mxtrunc256[ind].w[1]) || (fstar.w[3] == bid_ten2mxtrunc256[ind].w[3] && fstar.w[2] == bid_ten2mxtrunc256[ind].w[2] && fstar.w[1] == bid_ten2mxtrunc256[ind].w[1] && fstar.w[0] > bid_ten2mxtrunc256[ind].w[0]) {
				is_inexact_lt_midpoint = 1
			}
		} else {
			is_inexact_gt_midpoint = 1
		}
	}
	// check for midpoints
	if fstar.w[7] == 0 && fstar.w[6] == 0 &&
		fstar.w[5] == 0 && fstar.w[4] == 0 &&
		(fstar.w[3] < bid_ten2mxtrunc256[ind].w[3] ||
			(fstar.w[3] == bid_ten2mxtrunc256[ind].w[3] &&
				fstar.w[2] < bid_ten2mxtrunc256[ind].w[2]) ||
			(fstar.w[3] == bid_ten2mxtrunc256[ind].w[3] &&
				fstar.w[2] == bid_ten2mxtrunc256[ind].w[2] &&
				fstar.w[1] < bid_ten2mxtrunc256[ind].w[1]) ||
			(fstar.w[3] == bid_ten2mxtrunc256[ind].w[3] &&
				fstar.w[2] == bid_ten2mxtrunc256[ind].w[2] &&
				fstar.w[1] == bid_ten2mxtrunc256[ind].w[1] &&
				fstar.w[0] <= bid_ten2mxtrunc256[ind].w[0])) {
		if Cstar.w[0]&0x01 != 0 { // Cstar is odd
			Cstar.w[0]--
			if Cstar.w[0] == 0xffffffffffffffff {
				Cstar.w[1]--
				if Cstar.w[1] == 0xffffffffffffffff {
					Cstar.w[2]--
					if Cstar.w[2] == 0xffffffffffffffff {
						Cstar.w[3]--
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
		if Cstar.w[3] == 0x0 && Cstar.w[2] == 0x0 &&
			Cstar.w[1] == 0x0 && Cstar.w[0] == bid_ten2k64[ind] {
			Cstar.w[0] = bid_ten2k64[ind-1]
			incr_exp = 1
		} else {
			incr_exp = 0
		}
	} else if ind == 20 {
		if Cstar.w[3] == 0x0 && Cstar.w[2] == 0x0 &&
			Cstar.w[1] == bid_ten2k128[0].w[1] &&
			Cstar.w[0] == bid_ten2k128[0].w[0] {
			Cstar.w[0] = bid_ten2k64[19]
			Cstar.w[1] = 0x0
			incr_exp = 1
		} else {
			incr_exp = 0
		}
	} else if ind <= 38 { // if 21 <= ind <= 38
		if Cstar.w[3] == 0x0 && Cstar.w[2] == 0x0 &&
			Cstar.w[1] == bid_ten2k128[ind-20].w[1] &&
			Cstar.w[0] == bid_ten2k128[ind-20].w[0] {
			Cstar.w[0] = bid_ten2k128[ind-21].w[0]
			Cstar.w[1] = bid_ten2k128[ind-21].w[1]
			incr_exp = 1
		} else {
			incr_exp = 0
		}
	} else if ind == 39 {
		if Cstar.w[3] == 0x0 && Cstar.w[2] == bid_ten2k256[0].w[2] &&
			Cstar.w[1] == bid_ten2k256[0].w[1] &&
			Cstar.w[0] == bid_ten2k256[0].w[0] {
			Cstar.w[0] = bid_ten2k128[18].w[0]
			Cstar.w[1] = bid_ten2k128[18].w[1]
			Cstar.w[2] = 0x0
			incr_exp = 1
		} else {
			incr_exp = 0
		}
	} else if ind <= 57 { // if 40 <= ind <= 57
		if Cstar.w[3] == 0x0 && Cstar.w[2] == bid_ten2k256[ind-39].w[2] &&
			Cstar.w[1] == bid_ten2k256[ind-39].w[1] &&
			Cstar.w[0] == bid_ten2k256[ind-39].w[0] {
			Cstar.w[0] = bid_ten2k256[ind-40].w[0]
			Cstar.w[1] = bid_ten2k256[ind-40].w[1]
			Cstar.w[2] = bid_ten2k256[ind-40].w[2]
			incr_exp = 1
		} else {
			incr_exp = 0
		}
	} else { // if 58 <= ind <= 77 (actually 58 <= ind <= 74)
		if Cstar.w[3] == bid_ten2k256[ind-39].w[3] &&
			Cstar.w[2] == bid_ten2k256[ind-39].w[2] &&
			Cstar.w[1] == bid_ten2k256[ind-39].w[1] &&
			Cstar.w[0] == bid_ten2k256[ind-39].w[0] {
			Cstar.w[0] = bid_ten2k256[ind-40].w[0]
			Cstar.w[1] = bid_ten2k256[ind-40].w[1]
			Cstar.w[2] = bid_ten2k256[ind-40].w[2]
			Cstar.w[3] = bid_ten2k256[ind-40].w[3]
			incr_exp = 1
		} else {
			incr_exp = 0
		}
	}
	return
}
