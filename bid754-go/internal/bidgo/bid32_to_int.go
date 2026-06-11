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

// === from_int64/uint64 (via bid64) ===

func Bid32FromInt64(x int64, rnd_mode int) (uint32, uint32) {
	r64, f1 := Bid64FromInt64(x, rnd_mode)
	r32, f2 := Bid64ToBid32(r64, rnd_mode)
	return r32, f1 | f2
}

func Bid32FromUint64(x uint64, rnd_mode int) (uint32, uint32) {
	r64, f1 := Bid64FromUint64(x, rnd_mode)
	r32, f2 := Bid64ToBid32(r64, rnd_mode)
	return r32, f1 | f2
}
