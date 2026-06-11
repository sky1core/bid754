package bidgo

const (
	bid64ToUint8SizeMaskUint32  = uint32(0xffffff00)
	bid64ToUint8InvalidRes      = uint32(0x80)
	bid64ToUint16SizeMaskUint32 = uint32(0xffff0000)
	bid64ToUint16InvalidRes     = uint32(0x8000)
)

// Bid64ToUint8Rnint is ported mechanically from Intel bid64_to_uint8.c: bid64_to_uint8_rnint.
func Bid64ToUint8Rnint(x uint64) (uint8, uint32) {
	var res uint32
	var pfpsf uint32
	var saved_fpsc uint32

	saved_fpsc = pfpsf
	res, pfpsf = Bid64ToUint32Rnint(x)
	if (res & bid64ToUint8SizeMaskUint32) != 0 {
		pfpsf = saved_fpsc | BID_INVALID_EXCEPTION
		res = bid64ToUint8InvalidRes
	}
	return uint8(res), pfpsf
}

// Bid64ToUint8Xrnint is ported mechanically from Intel bid64_to_uint8.c: bid64_to_uint8_xrnint.
func Bid64ToUint8Xrnint(x uint64) (uint8, uint32) {
	var res uint32
	var pfpsf uint32
	var saved_fpsc uint32

	saved_fpsc = pfpsf
	res, pfpsf = Bid64ToUint32Xrnint(x)
	if (res & bid64ToUint8SizeMaskUint32) != 0 {
		pfpsf = saved_fpsc | BID_INVALID_EXCEPTION
		res = bid64ToUint8InvalidRes
	}
	return uint8(res), pfpsf
}

// Bid64ToUint8Rninta is ported mechanically from Intel bid64_to_uint8.c: bid64_to_uint8_rninta.
func Bid64ToUint8Rninta(x uint64) (uint8, uint32) {
	var res uint32
	var pfpsf uint32
	var saved_fpsc uint32

	saved_fpsc = pfpsf
	res, pfpsf = Bid64ToUint32Rninta(x)
	if (res & bid64ToUint8SizeMaskUint32) != 0 {
		pfpsf = saved_fpsc | BID_INVALID_EXCEPTION
		res = bid64ToUint8InvalidRes
	}
	return uint8(res), pfpsf
}

// Bid64ToUint8Xrninta is ported mechanically from Intel bid64_to_uint8.c: bid64_to_uint8_xrninta.
func Bid64ToUint8Xrninta(x uint64) (uint8, uint32) {
	var res uint32
	var pfpsf uint32
	var saved_fpsc uint32

	saved_fpsc = pfpsf
	res, pfpsf = Bid64ToUint32Xrninta(x)
	if (res & bid64ToUint8SizeMaskUint32) != 0 {
		pfpsf = saved_fpsc | BID_INVALID_EXCEPTION
		res = bid64ToUint8InvalidRes
	}
	return uint8(res), pfpsf
}

// Bid64ToUint8Int is ported mechanically from Intel bid64_to_uint8.c: bid64_to_uint8_int.
func Bid64ToUint8Int(x uint64) (uint8, uint32) {
	var res uint32
	var pfpsf uint32
	var saved_fpsc uint32

	saved_fpsc = pfpsf
	res, pfpsf = Bid64ToUint32Int(x)
	if (res & bid64ToUint8SizeMaskUint32) != 0 {
		pfpsf = saved_fpsc | BID_INVALID_EXCEPTION
		res = bid64ToUint8InvalidRes
	}
	return uint8(res), pfpsf
}

// Bid64ToUint8Xint is ported mechanically from Intel bid64_to_uint8.c: bid64_to_uint8_xint.
func Bid64ToUint8Xint(x uint64) (uint8, uint32) {
	var res uint32
	var pfpsf uint32
	var saved_fpsc uint32

	saved_fpsc = pfpsf
	res, pfpsf = Bid64ToUint32Xint(x)
	if (res & bid64ToUint8SizeMaskUint32) != 0 {
		pfpsf = saved_fpsc | BID_INVALID_EXCEPTION
		res = bid64ToUint8InvalidRes
	}
	return uint8(res), pfpsf
}

// Bid64ToUint8Floor is ported mechanically from Intel bid64_to_uint8.c: bid64_to_uint8_floor.
func Bid64ToUint8Floor(x uint64) (uint8, uint32) {
	var res uint32
	var pfpsf uint32
	var saved_fpsc uint32

	saved_fpsc = pfpsf
	res, pfpsf = Bid64ToUint32Floor(x)
	if (res & bid64ToUint8SizeMaskUint32) != 0 {
		pfpsf = saved_fpsc | BID_INVALID_EXCEPTION
		res = bid64ToUint8InvalidRes
	}
	return uint8(res), pfpsf
}

// Bid64ToUint8Ceil is ported mechanically from Intel bid64_to_uint8.c: bid64_to_uint8_ceil.
func Bid64ToUint8Ceil(x uint64) (uint8, uint32) {
	var res uint32
	var pfpsf uint32
	var saved_fpsc uint32

	saved_fpsc = pfpsf
	res, pfpsf = Bid64ToUint32Ceil(x)
	if (res & bid64ToUint8SizeMaskUint32) != 0 {
		pfpsf = saved_fpsc | BID_INVALID_EXCEPTION
		res = bid64ToUint8InvalidRes
	}
	return uint8(res), pfpsf
}

// Bid64ToUint8Xfloor is ported mechanically from Intel bid64_to_uint8.c: bid64_to_uint8_xfloor.
func Bid64ToUint8Xfloor(x uint64) (uint8, uint32) {
	var res uint32
	var pfpsf uint32
	var saved_fpsc uint32

	saved_fpsc = pfpsf
	res, pfpsf = Bid64ToUint32Xfloor(x)
	if (res & bid64ToUint8SizeMaskUint32) != 0 {
		pfpsf = saved_fpsc | BID_INVALID_EXCEPTION
		res = bid64ToUint8InvalidRes
	}
	return uint8(res), pfpsf
}

// Bid64ToUint8Xceil is ported mechanically from Intel bid64_to_uint8.c: bid64_to_uint8_xceil.
func Bid64ToUint8Xceil(x uint64) (uint8, uint32) {
	var res uint32
	var pfpsf uint32
	var saved_fpsc uint32

	saved_fpsc = pfpsf
	res, pfpsf = Bid64ToUint32Xceil(x)
	if (res & bid64ToUint8SizeMaskUint32) != 0 {
		pfpsf = saved_fpsc | BID_INVALID_EXCEPTION
		res = bid64ToUint8InvalidRes
	}
	return uint8(res), pfpsf
}

// Bid64ToUint16Rnint is ported mechanically from Intel bid64_to_uint16.c: bid64_to_uint16_rnint.
func Bid64ToUint16Rnint(x uint64) (uint16, uint32) {
	var res uint32
	var pfpsf uint32
	var saved_fpsc uint32

	saved_fpsc = pfpsf
	res, pfpsf = Bid64ToUint32Rnint(x)
	if (res & bid64ToUint16SizeMaskUint32) != 0 {
		pfpsf = saved_fpsc | BID_INVALID_EXCEPTION
		res = bid64ToUint16InvalidRes
	}
	return uint16(res), pfpsf
}

// Bid64ToUint16Xrnint is ported mechanically from Intel bid64_to_uint16.c: bid64_to_uint16_xrnint.
func Bid64ToUint16Xrnint(x uint64) (uint16, uint32) {
	var res uint32
	var pfpsf uint32
	var saved_fpsc uint32

	saved_fpsc = pfpsf
	res, pfpsf = Bid64ToUint32Xrnint(x)
	if (res & bid64ToUint16SizeMaskUint32) != 0 {
		pfpsf = saved_fpsc | BID_INVALID_EXCEPTION
		res = bid64ToUint16InvalidRes
	}
	return uint16(res), pfpsf
}

// Bid64ToUint16Rninta is ported mechanically from Intel bid64_to_uint16.c: bid64_to_uint16_rninta.
func Bid64ToUint16Rninta(x uint64) (uint16, uint32) {
	var res uint32
	var pfpsf uint32
	var saved_fpsc uint32

	saved_fpsc = pfpsf
	res, pfpsf = Bid64ToUint32Rninta(x)
	if (res & bid64ToUint16SizeMaskUint32) != 0 {
		pfpsf = saved_fpsc | BID_INVALID_EXCEPTION
		res = bid64ToUint16InvalidRes
	}
	return uint16(res), pfpsf
}

// Bid64ToUint16Xrninta is ported mechanically from Intel bid64_to_uint16.c: bid64_to_uint16_xrninta.
func Bid64ToUint16Xrninta(x uint64) (uint16, uint32) {
	var res uint32
	var pfpsf uint32
	var saved_fpsc uint32

	saved_fpsc = pfpsf
	res, pfpsf = Bid64ToUint32Xrninta(x)
	if (res & bid64ToUint16SizeMaskUint32) != 0 {
		pfpsf = saved_fpsc | BID_INVALID_EXCEPTION
		res = bid64ToUint16InvalidRes
	}
	return uint16(res), pfpsf
}

// Bid64ToUint16Int is ported mechanically from Intel bid64_to_uint16.c: bid64_to_uint16_int.
func Bid64ToUint16Int(x uint64) (uint16, uint32) {
	var res uint32
	var pfpsf uint32
	var saved_fpsc uint32

	saved_fpsc = pfpsf
	res, pfpsf = Bid64ToUint32Int(x)
	if (res & bid64ToUint16SizeMaskUint32) != 0 {
		pfpsf = saved_fpsc | BID_INVALID_EXCEPTION
		res = bid64ToUint16InvalidRes
	}
	return uint16(res), pfpsf
}

// Bid64ToUint16Xint is ported mechanically from Intel bid64_to_uint16.c: bid64_to_uint16_xint.
func Bid64ToUint16Xint(x uint64) (uint16, uint32) {
	var res uint32
	var pfpsf uint32
	var saved_fpsc uint32

	saved_fpsc = pfpsf
	res, pfpsf = Bid64ToUint32Xint(x)
	if (res & bid64ToUint16SizeMaskUint32) != 0 {
		pfpsf = saved_fpsc | BID_INVALID_EXCEPTION
		res = bid64ToUint16InvalidRes
	}
	return uint16(res), pfpsf
}

// Bid64ToUint16Floor is ported mechanically from Intel bid64_to_uint16.c: bid64_to_uint16_floor.
func Bid64ToUint16Floor(x uint64) (uint16, uint32) {
	var res uint32
	var pfpsf uint32
	var saved_fpsc uint32

	saved_fpsc = pfpsf
	res, pfpsf = Bid64ToUint32Floor(x)
	if (res & bid64ToUint16SizeMaskUint32) != 0 {
		pfpsf = saved_fpsc | BID_INVALID_EXCEPTION
		res = bid64ToUint16InvalidRes
	}
	return uint16(res), pfpsf
}

// Bid64ToUint16Ceil is ported mechanically from Intel bid64_to_uint16.c: bid64_to_uint16_ceil.
func Bid64ToUint16Ceil(x uint64) (uint16, uint32) {
	var res uint32
	var pfpsf uint32
	var saved_fpsc uint32

	saved_fpsc = pfpsf
	res, pfpsf = Bid64ToUint32Ceil(x)
	if (res & bid64ToUint16SizeMaskUint32) != 0 {
		pfpsf = saved_fpsc | BID_INVALID_EXCEPTION
		res = bid64ToUint16InvalidRes
	}
	return uint16(res), pfpsf
}

// Bid64ToUint16Xfloor is ported mechanically from Intel bid64_to_uint16.c: bid64_to_uint16_xfloor.
func Bid64ToUint16Xfloor(x uint64) (uint16, uint32) {
	var res uint32
	var pfpsf uint32
	var saved_fpsc uint32

	saved_fpsc = pfpsf
	res, pfpsf = Bid64ToUint32Xfloor(x)
	if (res & bid64ToUint16SizeMaskUint32) != 0 {
		pfpsf = saved_fpsc | BID_INVALID_EXCEPTION
		res = bid64ToUint16InvalidRes
	}
	return uint16(res), pfpsf
}

// Bid64ToUint16Xceil is ported mechanically from Intel bid64_to_uint16.c: bid64_to_uint16_xceil.
func Bid64ToUint16Xceil(x uint64) (uint16, uint32) {
	var res uint32
	var pfpsf uint32
	var saved_fpsc uint32

	saved_fpsc = pfpsf
	res, pfpsf = Bid64ToUint32Xceil(x)
	if (res & bid64ToUint16SizeMaskUint32) != 0 {
		pfpsf = saved_fpsc | BID_INVALID_EXCEPTION
		res = bid64ToUint16InvalidRes
	}
	return uint16(res), pfpsf
}
