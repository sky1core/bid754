// Decimal32 → integer conversions via bid64.
// Intel bid32_to_int32.c etc. are direct implementations,
// but we can safely convert bid32→bid64 first then use bid64_to_int* functions
// since bid32 values are a subset of bid64.

package bidgo

// === to_int32 ===

func Bid32ToInt32Rnint(x uint32) (int32, uint32) {
	x64, _ := Bid32ToBid64(x)
	return Bid64ToInt32Rnint(x64)
}

func Bid32ToInt32Xrnint(x uint32) (int32, uint32) {
	x64, _ := Bid32ToBid64(x)
	return Bid64ToInt32Xrnint(x64)
}

func Bid32ToInt32Rninta(x uint32) (int32, uint32) {
	x64, _ := Bid32ToBid64(x)
	return Bid64ToInt32Rninta(x64)
}

func Bid32ToInt32Xrninta(x uint32) (int32, uint32) {
	x64, _ := Bid32ToBid64(x)
	return Bid64ToInt32Xrninta(x64)
}

func Bid32ToInt32Int(x uint32) (int32, uint32) {
	x64, _ := Bid32ToBid64(x)
	return Bid64ToInt32Int(x64)
}

func Bid32ToInt32Xint(x uint32) (int32, uint32) {
	x64, _ := Bid32ToBid64(x)
	return Bid64ToInt32Xint(x64)
}

func Bid32ToInt32Floor(x uint32) (int32, uint32) {
	x64, _ := Bid32ToBid64(x)
	return Bid64ToInt32Floor(x64)
}

func Bid32ToInt32Xfloor(x uint32) (int32, uint32) {
	x64, _ := Bid32ToBid64(x)
	return Bid64ToInt32Xfloor(x64)
}

func Bid32ToInt32Ceil(x uint32) (int32, uint32) {
	x64, _ := Bid32ToBid64(x)
	return Bid64ToInt32Ceil(x64)
}

func Bid32ToInt32Xceil(x uint32) (int32, uint32) {
	x64, _ := Bid32ToBid64(x)
	return Bid64ToInt32Xceil(x64)
}

// === to_int64 ===

func Bid32ToInt64Rnint(x uint32) (int64, uint32) {
	x64, _ := Bid32ToBid64(x)
	return Bid64ToInt64Rnint(x64)
}

func Bid32ToInt64Xrnint(x uint32) (int64, uint32) {
	x64, _ := Bid32ToBid64(x)
	return Bid64ToInt64Xrnint(x64)
}

func Bid32ToInt64Rninta(x uint32) (int64, uint32) {
	x64, _ := Bid32ToBid64(x)
	return Bid64ToInt64Rninta(x64)
}

func Bid32ToInt64Xrninta(x uint32) (int64, uint32) {
	x64, _ := Bid32ToBid64(x)
	return Bid64ToInt64Xrninta(x64)
}

func Bid32ToInt64Int(x uint32) (int64, uint32) {
	x64, _ := Bid32ToBid64(x)
	return Bid64ToInt64Int(x64)
}

func Bid32ToInt64Xint(x uint32) (int64, uint32) {
	x64, _ := Bid32ToBid64(x)
	return Bid64ToInt64Xint(x64)
}

func Bid32ToInt64Floor(x uint32) (int64, uint32) {
	x64, _ := Bid32ToBid64(x)
	return Bid64ToInt64Floor(x64)
}

func Bid32ToInt64Xfloor(x uint32) (int64, uint32) {
	x64, _ := Bid32ToBid64(x)
	return Bid64ToInt64Xfloor(x64)
}

func Bid32ToInt64Ceil(x uint32) (int64, uint32) {
	x64, _ := Bid32ToBid64(x)
	return Bid64ToInt64Ceil(x64)
}

func Bid32ToInt64Xceil(x uint32) (int64, uint32) {
	x64, _ := Bid32ToBid64(x)
	return Bid64ToInt64Xceil(x64)
}

// === to_uint32 ===

func Bid32ToUint32Rnint(x uint32) (uint32, uint32) {
	x64, _ := Bid32ToBid64(x)
	return Bid64ToUint32Rnint(x64)
}

func Bid32ToUint32Xrnint(x uint32) (uint32, uint32) {
	x64, _ := Bid32ToBid64(x)
	return Bid64ToUint32Xrnint(x64)
}

func Bid32ToUint32Rninta(x uint32) (uint32, uint32) {
	x64, _ := Bid32ToBid64(x)
	return Bid64ToUint32Rninta(x64)
}

func Bid32ToUint32Xrninta(x uint32) (uint32, uint32) {
	x64, _ := Bid32ToBid64(x)
	return Bid64ToUint32Xrninta(x64)
}

func Bid32ToUint32Int(x uint32) (uint32, uint32) {
	x64, _ := Bid32ToBid64(x)
	return Bid64ToUint32Int(x64)
}

func Bid32ToUint32Xint(x uint32) (uint32, uint32) {
	x64, _ := Bid32ToBid64(x)
	return Bid64ToUint32Xint(x64)
}

func Bid32ToUint32Floor(x uint32) (uint32, uint32) {
	x64, _ := Bid32ToBid64(x)
	return Bid64ToUint32Floor(x64)
}

func Bid32ToUint32Xfloor(x uint32) (uint32, uint32) {
	x64, _ := Bid32ToBid64(x)
	return Bid64ToUint32Xfloor(x64)
}

func Bid32ToUint32Ceil(x uint32) (uint32, uint32) {
	x64, _ := Bid32ToBid64(x)
	return Bid64ToUint32Ceil(x64)
}

func Bid32ToUint32Xceil(x uint32) (uint32, uint32) {
	x64, _ := Bid32ToBid64(x)
	return Bid64ToUint32Xceil(x64)
}

// === to_uint64 ===

func Bid32ToUint64Rnint(x uint32) (uint64, uint32) {
	x64, _ := Bid32ToBid64(x)
	return Bid64ToUint64Rnint(x64)
}

func Bid32ToUint64Xrnint(x uint32) (uint64, uint32) {
	x64, _ := Bid32ToBid64(x)
	return Bid64ToUint64Xrnint(x64)
}

func Bid32ToUint64Rninta(x uint32) (uint64, uint32) {
	x64, _ := Bid32ToBid64(x)
	return Bid64ToUint64Rninta(x64)
}

func Bid32ToUint64Xrninta(x uint32) (uint64, uint32) {
	x64, _ := Bid32ToBid64(x)
	return Bid64ToUint64Xrninta(x64)
}

func Bid32ToUint64Int(x uint32) (uint64, uint32) {
	x64, _ := Bid32ToBid64(x)
	return Bid64ToUint64Int(x64)
}

func Bid32ToUint64Xint(x uint32) (uint64, uint32) {
	x64, _ := Bid32ToBid64(x)
	return Bid64ToUint64Xint(x64)
}

func Bid32ToUint64Floor(x uint32) (uint64, uint32) {
	x64, _ := Bid32ToBid64(x)
	return Bid64ToUint64Floor(x64)
}

func Bid32ToUint64Xfloor(x uint32) (uint64, uint32) {
	x64, _ := Bid32ToBid64(x)
	return Bid64ToUint64Xfloor(x64)
}

func Bid32ToUint64Ceil(x uint32) (uint64, uint32) {
	x64, _ := Bid32ToBid64(x)
	return Bid64ToUint64Ceil(x64)
}

func Bid32ToUint64Xceil(x uint32) (uint64, uint32) {
	x64, _ := Bid32ToBid64(x)
	return Bid64ToUint64Xceil(x64)
}

// === to_int8 (via int32 with range check) ===

func bid32_to_small_int(fn func(uint32) (int32, uint32), x uint32, sizeMask int32, invalidResult int8) (int8, uint32) {
	v, f := fn(x)
	if f&BID_INVALID_EXCEPTION != 0 {
		return invalidResult, f
	}
	sgnMask := v & sizeMask
	if sgnMask != 0 && sgnMask != sizeMask {
		return invalidResult, BID_INVALID_EXCEPTION
	}
	return int8(v), f
}

func bid32_to_small_int16(fn func(uint32) (int32, uint32), x uint32, sizeMask int32, invalidResult int16) (int16, uint32) {
	v, f := fn(x)
	if f&BID_INVALID_EXCEPTION != 0 {
		return invalidResult, f
	}
	sgnMask := v & sizeMask
	if sgnMask != 0 && sgnMask != sizeMask {
		return invalidResult, BID_INVALID_EXCEPTION
	}
	return int16(v), f
}

func bid32_to_small_uint(fn func(uint32) (uint32, uint32), x uint32, sizeMask uint32, invalidResult uint8) (uint8, uint32) {
	v, f := fn(x)
	if f&BID_INVALID_EXCEPTION != 0 {
		return invalidResult, f
	}
	if v&sizeMask != 0 {
		return invalidResult, BID_INVALID_EXCEPTION
	}
	return uint8(v), f
}

func bid32_to_small_uint16(fn func(uint32) (uint32, uint32), x uint32, sizeMask uint32, invalidResult uint16) (uint16, uint32) {
	v, f := fn(x)
	if f&BID_INVALID_EXCEPTION != 0 {
		return invalidResult, f
	}
	if v&sizeMask != 0 {
		return invalidResult, BID_INVALID_EXCEPTION
	}
	return uint16(v), f
}

func Bid32ToInt8Rnint(x uint32) (int8, uint32) {
	return bid32_to_small_int(Bid32ToInt32Rnint, x, -128, -128)
}
func Bid32ToInt8Xrnint(x uint32) (int8, uint32) {
	return bid32_to_small_int(Bid32ToInt32Xrnint, x, -128, -128)
}
func Bid32ToInt8Rninta(x uint32) (int8, uint32) {
	return bid32_to_small_int(Bid32ToInt32Rninta, x, -128, -128)
}
func Bid32ToInt8Xrninta(x uint32) (int8, uint32) {
	return bid32_to_small_int(Bid32ToInt32Xrninta, x, -128, -128)
}
func Bid32ToInt8Int(x uint32) (int8, uint32) {
	return bid32_to_small_int(Bid32ToInt32Int, x, -128, -128)
}
func Bid32ToInt8Xint(x uint32) (int8, uint32) {
	return bid32_to_small_int(Bid32ToInt32Xint, x, -128, -128)
}
func Bid32ToInt8Floor(x uint32) (int8, uint32) {
	return bid32_to_small_int(Bid32ToInt32Floor, x, -128, -128)
}
func Bid32ToInt8Xfloor(x uint32) (int8, uint32) {
	return bid32_to_small_int(Bid32ToInt32Xfloor, x, -128, -128)
}
func Bid32ToInt8Ceil(x uint32) (int8, uint32) {
	return bid32_to_small_int(Bid32ToInt32Ceil, x, -128, -128)
}
func Bid32ToInt8Xceil(x uint32) (int8, uint32) {
	return bid32_to_small_int(Bid32ToInt32Xceil, x, -128, -128)
}

// === to_int16 (via int32) ===

func Bid32ToInt16Rnint(x uint32) (int16, uint32) {
	return bid32_to_small_int16(Bid32ToInt32Rnint, x, -32768, -32768)
}
func Bid32ToInt16Xrnint(x uint32) (int16, uint32) {
	return bid32_to_small_int16(Bid32ToInt32Xrnint, x, -32768, -32768)
}
func Bid32ToInt16Rninta(x uint32) (int16, uint32) {
	return bid32_to_small_int16(Bid32ToInt32Rninta, x, -32768, -32768)
}
func Bid32ToInt16Xrninta(x uint32) (int16, uint32) {
	return bid32_to_small_int16(Bid32ToInt32Xrninta, x, -32768, -32768)
}
func Bid32ToInt16Int(x uint32) (int16, uint32) {
	return bid32_to_small_int16(Bid32ToInt32Int, x, -32768, -32768)
}
func Bid32ToInt16Xint(x uint32) (int16, uint32) {
	return bid32_to_small_int16(Bid32ToInt32Xint, x, -32768, -32768)
}
func Bid32ToInt16Floor(x uint32) (int16, uint32) {
	return bid32_to_small_int16(Bid32ToInt32Floor, x, -32768, -32768)
}
func Bid32ToInt16Xfloor(x uint32) (int16, uint32) {
	return bid32_to_small_int16(Bid32ToInt32Xfloor, x, -32768, -32768)
}
func Bid32ToInt16Ceil(x uint32) (int16, uint32) {
	return bid32_to_small_int16(Bid32ToInt32Ceil, x, -32768, -32768)
}
func Bid32ToInt16Xceil(x uint32) (int16, uint32) {
	return bid32_to_small_int16(Bid32ToInt32Xceil, x, -32768, -32768)
}

// === to_uint8 (via uint32) ===

func Bid32ToUint8Rnint(x uint32) (uint8, uint32) {
	return bid32_to_small_uint(Bid32ToUint32Rnint, x, 0xffffff00, 0x80)
}
func Bid32ToUint8Xrnint(x uint32) (uint8, uint32) {
	return bid32_to_small_uint(Bid32ToUint32Xrnint, x, 0xffffff00, 0x80)
}
func Bid32ToUint8Rninta(x uint32) (uint8, uint32) {
	return bid32_to_small_uint(Bid32ToUint32Rninta, x, 0xffffff00, 0x80)
}
func Bid32ToUint8Xrninta(x uint32) (uint8, uint32) {
	return bid32_to_small_uint(Bid32ToUint32Xrninta, x, 0xffffff00, 0x80)
}
func Bid32ToUint8Int(x uint32) (uint8, uint32) {
	return bid32_to_small_uint(Bid32ToUint32Int, x, 0xffffff00, 0x80)
}
func Bid32ToUint8Xint(x uint32) (uint8, uint32) {
	return bid32_to_small_uint(Bid32ToUint32Xint, x, 0xffffff00, 0x80)
}
func Bid32ToUint8Floor(x uint32) (uint8, uint32) {
	return bid32_to_small_uint(Bid32ToUint32Floor, x, 0xffffff00, 0x80)
}
func Bid32ToUint8Xfloor(x uint32) (uint8, uint32) {
	return bid32_to_small_uint(Bid32ToUint32Xfloor, x, 0xffffff00, 0x80)
}
func Bid32ToUint8Ceil(x uint32) (uint8, uint32) {
	return bid32_to_small_uint(Bid32ToUint32Ceil, x, 0xffffff00, 0x80)
}
func Bid32ToUint8Xceil(x uint32) (uint8, uint32) {
	return bid32_to_small_uint(Bid32ToUint32Xceil, x, 0xffffff00, 0x80)
}

// === to_uint16 (via uint32) ===

func Bid32ToUint16Rnint(x uint32) (uint16, uint32) {
	return bid32_to_small_uint16(Bid32ToUint32Rnint, x, 0xffff0000, 0x8000)
}
func Bid32ToUint16Xrnint(x uint32) (uint16, uint32) {
	return bid32_to_small_uint16(Bid32ToUint32Xrnint, x, 0xffff0000, 0x8000)
}
func Bid32ToUint16Rninta(x uint32) (uint16, uint32) {
	return bid32_to_small_uint16(Bid32ToUint32Rninta, x, 0xffff0000, 0x8000)
}
func Bid32ToUint16Xrninta(x uint32) (uint16, uint32) {
	return bid32_to_small_uint16(Bid32ToUint32Xrninta, x, 0xffff0000, 0x8000)
}
func Bid32ToUint16Int(x uint32) (uint16, uint32) {
	return bid32_to_small_uint16(Bid32ToUint32Int, x, 0xffff0000, 0x8000)
}
func Bid32ToUint16Xint(x uint32) (uint16, uint32) {
	return bid32_to_small_uint16(Bid32ToUint32Xint, x, 0xffff0000, 0x8000)
}
func Bid32ToUint16Floor(x uint32) (uint16, uint32) {
	return bid32_to_small_uint16(Bid32ToUint32Floor, x, 0xffff0000, 0x8000)
}
func Bid32ToUint16Xfloor(x uint32) (uint16, uint32) {
	return bid32_to_small_uint16(Bid32ToUint32Xfloor, x, 0xffff0000, 0x8000)
}
func Bid32ToUint16Ceil(x uint32) (uint16, uint32) {
	return bid32_to_small_uint16(Bid32ToUint32Ceil, x, 0xffff0000, 0x8000)
}
func Bid32ToUint16Xceil(x uint32) (uint16, uint32) {
	return bid32_to_small_uint16(Bid32ToUint32Xceil, x, 0xffff0000, 0x8000)
}

// === lrint/llrint/lround/llround ===

func Bid32Lrint(x uint32, rnd_mode int) (int64, uint32) {
	x64, _ := Bid32ToBid64(x)
	return Bid64Lrint(x64, rnd_mode)
}

func Bid32Llrint(x uint32, rnd_mode int) (int64, uint32) {
	x64, _ := Bid32ToBid64(x)
	return Bid64Llrint(x64, rnd_mode)
}

func Bid32Lround(x uint32) (int64, uint32) {
	x64, _ := Bid32ToBid64(x)
	return Bid64Lround(x64)
}

func Bid32Llround(x uint32) (int64, uint32) {
	x64, _ := Bid32ToBid64(x)
	return Bid64Llround(x64)
}

func Bid32FromInt64(x int64, rnd_mode int) (uint32, uint32) {
	var res uint32
	var res64 uint64
	var x_sign uint32
	var C uint64
	var q, ind uint32
	var incr_exp int
	var is_midpoint_lt_even, is_midpoint_gt_even int
	var is_inexact_lt_midpoint, is_inexact_gt_midpoint int
	var pfpsf uint32

	x_sign = uint32(uint64(x)>>32) & MASK_SIGN32
	if x_sign != 0 {
		C = ^uint64(x) + 1
	} else {
		C = uint64(x)
	}
	if C <= uint64(BID32_SIG_MAX) {
		if C < 0x00800000 {
			res = x_sign | 0x32800000 | uint32(C)
		} else {
			res = x_sign | 0x6ca00000 | (uint32(C) & 0x001fffff)
		}
	} else {
		if C < 100000000 {
			q = 8
			ind = 1
		} else if C < 1000000000 {
			q = 9
			ind = 2
		} else if C < 10000000000 {
			q = 10
			ind = 3
		} else if C < 100000000000 {
			q = 11
			ind = 4
		} else if C < 1000000000000 {
			q = 12
			ind = 5
		} else if C < 10000000000000 {
			q = 13
			ind = 6
		} else if C < 100000000000000 {
			q = 14
			ind = 7
		} else if C < 1000000000000000 {
			q = 15
			ind = 8
		} else if C < 10000000000000000 {
			q = 16
			ind = 9
		} else if C < 100000000000000000 {
			q = 17
			ind = 10
		} else if C < 1000000000000000000 {
			q = 18
			ind = 11
		} else {
			q = 19
			ind = 12
		}
		res64 = bid_round64_2_18(int(q), int(ind), uint64(C), &incr_exp,
			&is_midpoint_lt_even, &is_midpoint_gt_even,
			&is_inexact_lt_midpoint, &is_inexact_gt_midpoint)
		res = uint32(res64)
		if incr_exp != 0 {
			ind++
		}
		if is_inexact_lt_midpoint != 0 || is_inexact_gt_midpoint != 0 ||
			is_midpoint_lt_even != 0 || is_midpoint_gt_even != 0 {
			pfpsf |= BID_INEXACT_EXCEPTION
		}
		if rnd_mode != BID_ROUNDING_TO_NEAREST {
			if (x_sign == 0 &&
				((rnd_mode == BID_ROUNDING_UP && is_inexact_lt_midpoint != 0) ||
					((rnd_mode == BID_ROUNDING_TIES_AWAY || rnd_mode == BID_ROUNDING_UP) && is_midpoint_gt_even != 0))) ||
				(x_sign != 0 &&
					((rnd_mode == BID_ROUNDING_DOWN && is_inexact_lt_midpoint != 0) ||
						((rnd_mode == BID_ROUNDING_TIES_AWAY || rnd_mode == BID_ROUNDING_DOWN) && is_midpoint_gt_even != 0))) {
				res = res + 1
				if res == 10000000 {
					res = 1000000
					ind = ind + 1
				}
			} else if (is_midpoint_lt_even != 0 || is_inexact_gt_midpoint != 0) &&
				((x_sign != 0 && (rnd_mode == BID_ROUNDING_UP || rnd_mode == BID_ROUNDING_TO_ZERO)) ||
					(x_sign == 0 && (rnd_mode == BID_ROUNDING_DOWN || rnd_mode == BID_ROUNDING_TO_ZERO))) {
				res = res - 1
				if res == 999999 {
					res = 9999999
					ind = ind - 1
				}
			}
		}
		if res < 0x00800000 {
			res = x_sign | ((ind + 101) << 23) | res
		} else {
			res = x_sign | 0x60000000 | ((ind + 101) << 21) | (res & 0x001fffff)
		}
	}
	return res, pfpsf
}

func Bid32FromUint64(x uint64, rnd_mode int) (uint32, uint32) {
	var res uint32
	var res64 uint64
	var C uint64
	var q, ind uint32
	var incr_exp int
	var is_midpoint_lt_even, is_midpoint_gt_even int
	var is_inexact_lt_midpoint, is_inexact_gt_midpoint int
	var pfpsf uint32

	C = x
	if C <= uint64(BID32_SIG_MAX) {
		if C < 0x00800000 {
			res = 0x32800000 | uint32(C)
		} else {
			res = 0x6ca00000 | (uint32(C) & 0x001fffff)
		}
	} else {
		if C < 100000000 {
			q = 8
			ind = 1
		} else if C < 1000000000 {
			q = 9
			ind = 2
		} else if C < 10000000000 {
			q = 10
			ind = 3
		} else if C < 100000000000 {
			q = 11
			ind = 4
		} else if C < 1000000000000 {
			q = 12
			ind = 5
		} else if C < 10000000000000 {
			q = 13
			ind = 6
		} else if C < 100000000000000 {
			q = 14
			ind = 7
		} else if C < 1000000000000000 {
			q = 15
			ind = 8
		} else if C < 10000000000000000 {
			q = 16
			ind = 9
		} else if C < 100000000000000000 {
			q = 17
			ind = 10
		} else if C < 1000000000000000000 {
			q = 18
			ind = 11
		} else if C < 10000000000000000000 {
			q = 19
			ind = 12
		} else {
			q = 20
			ind = 13
		}
		if q <= 19 {
			res64 = bid_round64_2_18(int(q), int(ind), uint64(C), &incr_exp,
				&is_midpoint_lt_even, &is_midpoint_gt_even,
				&is_inexact_lt_midpoint, &is_inexact_gt_midpoint)
			res = uint32(res64)
		} else {
			var res128 BID_UINT128
			res128, incr_exp, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint = bid_round128_19_38(int(q), int(ind), BID_UINT128{lo: C, hi: 0})
			res = uint32(res128.lo)
		}
		if incr_exp != 0 {
			ind++
		}
		if is_inexact_lt_midpoint != 0 || is_inexact_gt_midpoint != 0 ||
			is_midpoint_lt_even != 0 || is_midpoint_gt_even != 0 {
			pfpsf |= BID_INEXACT_EXCEPTION
		}
		if rnd_mode != BID_ROUNDING_TO_NEAREST {
			if (rnd_mode == BID_ROUNDING_UP && is_inexact_lt_midpoint != 0) ||
				((rnd_mode == BID_ROUNDING_TIES_AWAY || rnd_mode == BID_ROUNDING_UP) && is_midpoint_gt_even != 0) {
				res = res + 1
				if res == 10000000 {
					res = 1000000
					ind = ind + 1
				}
			} else if (is_midpoint_lt_even != 0 || is_inexact_gt_midpoint != 0) &&
				(rnd_mode == BID_ROUNDING_DOWN || rnd_mode == BID_ROUNDING_TO_ZERO) {
				res = res - 1
				if res == 999999 {
					res = 9999999
					ind = ind - 1
				}
			}
		}
		if res < 0x00800000 {
			res = ((ind + 101) << 23) | res
		} else {
			res = 0x60000000 | ((ind + 101) << 21) | (res & 0x001fffff)
		}
	}
	return res, pfpsf
}
