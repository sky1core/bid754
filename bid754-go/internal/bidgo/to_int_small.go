package bidgo

const (
	bid64ToInt8SizeMaskInt32  = int32(-128)
	bid64ToInt8InvalidRes     = 0x80
	bid64ToInt16SizeMaskInt32 = int32(-32768)
	bid64ToInt16InvalidRes    = 0x8000
)

// Bid64ToInt8Rnint is ported mechanically from Intel bid64_to_int8.c: bid64_to_int8_rnint.
func Bid64ToInt8Rnint(x uint64) (int8, uint32) {
	var res int32
	var sgn_mask int32
	var pfpsf uint32
	var saved_fpsc uint32

	saved_fpsc = pfpsf
	res, pfpsf = Bid64ToInt32Rnint(x)
	sgn_mask = res & bid64ToInt8SizeMaskInt32
	if sgn_mask != 0 && sgn_mask != bid64ToInt8SizeMaskInt32 {
		pfpsf = saved_fpsc | BID_INVALID_EXCEPTION
		res = bid64ToInt8InvalidRes
	}
	return int8(res), pfpsf
}

// Bid64ToInt8Xrnint is ported mechanically from Intel bid64_to_int8.c: bid64_to_int8_xrnint.
func Bid64ToInt8Xrnint(x uint64) (int8, uint32) {
	var res int32
	var sgn_mask int32
	var pfpsf uint32
	var saved_fpsc uint32

	saved_fpsc = pfpsf
	res, pfpsf = Bid64ToInt32Xrnint(x)
	sgn_mask = res & bid64ToInt8SizeMaskInt32
	if sgn_mask != 0 && sgn_mask != bid64ToInt8SizeMaskInt32 {
		pfpsf = saved_fpsc | BID_INVALID_EXCEPTION
		res = bid64ToInt8InvalidRes
	}
	return int8(res), pfpsf
}

// Bid64ToInt16Rnint is ported mechanically from Intel bid64_to_int16.c: bid64_to_int16_rnint.
func Bid64ToInt16Rnint(x uint64) (int16, uint32) {
	var res int32
	var sgn_mask int32
	var pfpsf uint32
	var saved_fpsc uint32

	saved_fpsc = pfpsf
	res, pfpsf = Bid64ToInt32Rnint(x)
	sgn_mask = res & bid64ToInt16SizeMaskInt32
	if sgn_mask != 0 && sgn_mask != bid64ToInt16SizeMaskInt32 {
		pfpsf = saved_fpsc | BID_INVALID_EXCEPTION
		res = bid64ToInt16InvalidRes
	}
	return int16(res), pfpsf
}

// Bid64ToInt16Xrnint is ported mechanically from Intel bid64_to_int16.c: bid64_to_int16_xrnint.
func Bid64ToInt16Xrnint(x uint64) (int16, uint32) {
	var res int32
	var sgn_mask int32
	var pfpsf uint32
	var saved_fpsc uint32

	saved_fpsc = pfpsf
	res, pfpsf = Bid64ToInt32Xrnint(x)
	sgn_mask = res & bid64ToInt16SizeMaskInt32
	if sgn_mask != 0 && sgn_mask != bid64ToInt16SizeMaskInt32 {
		pfpsf = saved_fpsc | BID_INVALID_EXCEPTION
		res = bid64ToInt16InvalidRes
	}
	return int16(res), pfpsf
}

// Bid64ToInt8Rninta is ported mechanically from Intel bid64_to_int8.c: bid64_to_int8_rninta.
func Bid64ToInt8Rninta(x uint64) (int8, uint32) {
	var res int32
	var sgn_mask int32
	var pfpsf uint32
	var saved_fpsc uint32

	saved_fpsc = pfpsf
	res, pfpsf = Bid64ToInt32Rninta(x)
	sgn_mask = res & bid64ToInt8SizeMaskInt32
	if sgn_mask != 0 && sgn_mask != bid64ToInt8SizeMaskInt32 {
		pfpsf = saved_fpsc | BID_INVALID_EXCEPTION
		res = bid64ToInt8InvalidRes
	}
	return int8(res), pfpsf
}

// Bid64ToInt8Xrninta is ported mechanically from Intel bid64_to_int8.c: bid64_to_int8_xrninta.
func Bid64ToInt8Xrninta(x uint64) (int8, uint32) {
	var res int32
	var sgn_mask int32
	var pfpsf uint32
	var saved_fpsc uint32

	saved_fpsc = pfpsf
	res, pfpsf = Bid64ToInt32Xrninta(x)
	sgn_mask = res & bid64ToInt8SizeMaskInt32
	if sgn_mask != 0 && sgn_mask != bid64ToInt8SizeMaskInt32 {
		pfpsf = saved_fpsc | BID_INVALID_EXCEPTION
		res = bid64ToInt8InvalidRes
	}
	return int8(res), pfpsf
}

// Bid64ToInt16Rninta is ported mechanically from Intel bid64_to_int16.c: bid64_to_int16_rninta.
func Bid64ToInt16Rninta(x uint64) (int16, uint32) {
	var res int32
	var sgn_mask int32
	var pfpsf uint32
	var saved_fpsc uint32

	saved_fpsc = pfpsf
	res, pfpsf = Bid64ToInt32Rninta(x)
	sgn_mask = res & bid64ToInt16SizeMaskInt32
	if sgn_mask != 0 && sgn_mask != bid64ToInt16SizeMaskInt32 {
		pfpsf = saved_fpsc | BID_INVALID_EXCEPTION
		res = bid64ToInt16InvalidRes
	}
	return int16(res), pfpsf
}

// Bid64ToInt16Xrninta is ported mechanically from Intel bid64_to_int16.c: bid64_to_int16_xrninta.
func Bid64ToInt16Xrninta(x uint64) (int16, uint32) {
	var res int32
	var sgn_mask int32
	var pfpsf uint32
	var saved_fpsc uint32

	saved_fpsc = pfpsf
	res, pfpsf = Bid64ToInt32Xrninta(x)
	sgn_mask = res & bid64ToInt16SizeMaskInt32
	if sgn_mask != 0 && sgn_mask != bid64ToInt16SizeMaskInt32 {
		pfpsf = saved_fpsc | BID_INVALID_EXCEPTION
		res = bid64ToInt16InvalidRes
	}
	return int16(res), pfpsf
}

// Bid64ToInt8Int is ported mechanically from Intel bid64_to_int8.c: bid64_to_int8_int.
func Bid64ToInt8Int(x uint64) (int8, uint32) {
	var res int32
	var sgn_mask int32
	var pfpsf uint32
	var saved_fpsc uint32

	saved_fpsc = pfpsf
	res, pfpsf = Bid64ToInt32Int(x)
	sgn_mask = res & bid64ToInt8SizeMaskInt32
	if sgn_mask != 0 && sgn_mask != bid64ToInt8SizeMaskInt32 {
		pfpsf = saved_fpsc | BID_INVALID_EXCEPTION
		res = bid64ToInt8InvalidRes
	}
	return int8(res), pfpsf
}

// Bid64ToInt8Xint is ported mechanically from Intel bid64_to_int8.c: bid64_to_int8_xint.
func Bid64ToInt8Xint(x uint64) (int8, uint32) {
	var res int32
	var sgn_mask int32
	var pfpsf uint32
	var saved_fpsc uint32

	saved_fpsc = pfpsf
	res, pfpsf = Bid64ToInt32Xint(x)
	sgn_mask = res & bid64ToInt8SizeMaskInt32
	if sgn_mask != 0 && sgn_mask != bid64ToInt8SizeMaskInt32 {
		pfpsf = saved_fpsc | BID_INVALID_EXCEPTION
		res = bid64ToInt8InvalidRes
	}
	return int8(res), pfpsf
}

// Bid64ToInt16Int is ported mechanically from Intel bid64_to_int16.c: bid64_to_int16_int.
func Bid64ToInt16Int(x uint64) (int16, uint32) {
	var res int32
	var sgn_mask int32
	var pfpsf uint32
	var saved_fpsc uint32

	saved_fpsc = pfpsf
	res, pfpsf = Bid64ToInt32Int(x)
	sgn_mask = res & bid64ToInt16SizeMaskInt32
	if sgn_mask != 0 && sgn_mask != bid64ToInt16SizeMaskInt32 {
		pfpsf = saved_fpsc | BID_INVALID_EXCEPTION
		res = bid64ToInt16InvalidRes
	}
	return int16(res), pfpsf
}

// Bid64ToInt16Xint is ported mechanically from Intel bid64_to_int16.c: bid64_to_int16_xint.
func Bid64ToInt16Xint(x uint64) (int16, uint32) {
	var res int32
	var sgn_mask int32
	var pfpsf uint32
	var saved_fpsc uint32

	saved_fpsc = pfpsf
	res, pfpsf = Bid64ToInt32Xint(x)
	sgn_mask = res & bid64ToInt16SizeMaskInt32
	if sgn_mask != 0 && sgn_mask != bid64ToInt16SizeMaskInt32 {
		pfpsf = saved_fpsc | BID_INVALID_EXCEPTION
		res = bid64ToInt16InvalidRes
	}
	return int16(res), pfpsf
}

// Bid64ToInt8Floor is ported mechanically from Intel bid64_to_int8.c: bid64_to_int8_floor.
func Bid64ToInt8Floor(x uint64) (int8, uint32) {
	var res int32
	var sgn_mask int32
	var pfpsf uint32
	var saved_fpsc uint32

	saved_fpsc = pfpsf
	res, pfpsf = Bid64ToInt32Floor(x)
	sgn_mask = res & bid64ToInt8SizeMaskInt32
	if sgn_mask != 0 && sgn_mask != bid64ToInt8SizeMaskInt32 {
		pfpsf = saved_fpsc | BID_INVALID_EXCEPTION
		res = bid64ToInt8InvalidRes
	}
	return int8(res), pfpsf
}

// Bid64ToInt8Xfloor is ported mechanically from Intel bid64_to_int8.c: bid64_to_int8_xfloor.
func Bid64ToInt8Xfloor(x uint64) (int8, uint32) {
	var res int32
	var sgn_mask int32
	var pfpsf uint32
	var saved_fpsc uint32

	saved_fpsc = pfpsf
	res, pfpsf = Bid64ToInt32Xfloor(x)
	sgn_mask = res & bid64ToInt8SizeMaskInt32
	if sgn_mask != 0 && sgn_mask != bid64ToInt8SizeMaskInt32 {
		pfpsf = saved_fpsc | BID_INVALID_EXCEPTION
		res = bid64ToInt8InvalidRes
	}
	return int8(res), pfpsf
}

// Bid64ToInt16Floor is ported mechanically from Intel bid64_to_int16.c: bid64_to_int16_floor.
func Bid64ToInt16Floor(x uint64) (int16, uint32) {
	var res int32
	var sgn_mask int32
	var pfpsf uint32
	var saved_fpsc uint32

	saved_fpsc = pfpsf
	res, pfpsf = Bid64ToInt32Floor(x)
	sgn_mask = res & bid64ToInt16SizeMaskInt32
	if sgn_mask != 0 && sgn_mask != bid64ToInt16SizeMaskInt32 {
		pfpsf = saved_fpsc | BID_INVALID_EXCEPTION
		res = bid64ToInt16InvalidRes
	}
	return int16(res), pfpsf
}

// Bid64ToInt16Xfloor is ported mechanically from Intel bid64_to_int16.c: bid64_to_int16_xfloor.
func Bid64ToInt16Xfloor(x uint64) (int16, uint32) {
	var res int32
	var sgn_mask int32
	var pfpsf uint32
	var saved_fpsc uint32

	saved_fpsc = pfpsf
	res, pfpsf = Bid64ToInt32Xfloor(x)
	sgn_mask = res & bid64ToInt16SizeMaskInt32
	if sgn_mask != 0 && sgn_mask != bid64ToInt16SizeMaskInt32 {
		pfpsf = saved_fpsc | BID_INVALID_EXCEPTION
		res = bid64ToInt16InvalidRes
	}
	return int16(res), pfpsf
}

// Bid64ToInt8Ceil is ported mechanically from Intel bid64_to_int8.c: bid64_to_int8_ceil.
func Bid64ToInt8Ceil(x uint64) (int8, uint32) {
	var res int32
	var sgn_mask int32
	var pfpsf uint32
	var saved_fpsc uint32

	saved_fpsc = pfpsf
	res, pfpsf = Bid64ToInt32Ceil(x)
	sgn_mask = res & bid64ToInt8SizeMaskInt32
	if sgn_mask != 0 && sgn_mask != bid64ToInt8SizeMaskInt32 {
		pfpsf = saved_fpsc | BID_INVALID_EXCEPTION
		res = bid64ToInt8InvalidRes
	}
	return int8(res), pfpsf
}

// Bid64ToInt8Xceil is ported mechanically from Intel bid64_to_int8.c: bid64_to_int8_xceil.
func Bid64ToInt8Xceil(x uint64) (int8, uint32) {
	var res int32
	var sgn_mask int32
	var pfpsf uint32
	var saved_fpsc uint32

	saved_fpsc = pfpsf
	res, pfpsf = Bid64ToInt32Xceil(x)
	sgn_mask = res & bid64ToInt8SizeMaskInt32
	if sgn_mask != 0 && sgn_mask != bid64ToInt8SizeMaskInt32 {
		pfpsf = saved_fpsc | BID_INVALID_EXCEPTION
		res = bid64ToInt8InvalidRes
	}
	return int8(res), pfpsf
}

// Bid64ToInt16Ceil is ported mechanically from Intel bid64_to_int16.c: bid64_to_int16_ceil.
func Bid64ToInt16Ceil(x uint64) (int16, uint32) {
	var res int32
	var sgn_mask int32
	var pfpsf uint32
	var saved_fpsc uint32

	saved_fpsc = pfpsf
	res, pfpsf = Bid64ToInt32Ceil(x)
	sgn_mask = res & bid64ToInt16SizeMaskInt32
	if sgn_mask != 0 && sgn_mask != bid64ToInt16SizeMaskInt32 {
		pfpsf = saved_fpsc | BID_INVALID_EXCEPTION
		res = bid64ToInt16InvalidRes
	}
	return int16(res), pfpsf
}

// Bid64ToInt16Xceil is ported mechanically from Intel bid64_to_int16.c: bid64_to_int16_xceil.
func Bid64ToInt16Xceil(x uint64) (int16, uint32) {
	var res int32
	var sgn_mask int32
	var pfpsf uint32
	var saved_fpsc uint32

	saved_fpsc = pfpsf
	res, pfpsf = Bid64ToInt32Xceil(x)
	sgn_mask = res & bid64ToInt16SizeMaskInt32
	if sgn_mask != 0 && sgn_mask != bid64ToInt16SizeMaskInt32 {
		pfpsf = saved_fpsc | BID_INVALID_EXCEPTION
		res = bid64ToInt16InvalidRes
	}
	return int16(res), pfpsf
}
