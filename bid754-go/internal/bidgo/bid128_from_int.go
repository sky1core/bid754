// Ported from: Intel bid_from_int.c (bid128 section)
// Mechanical translation - all logic preserved exactly.

package bidgo

// Bid128FromInt32 converts int32 to BID128.
// Ported mechanically from Intel bid_from_int.c: bid128_from_int32.
func Bid128FromInt32(x int32) BID_UINT128 {
	var res BID_UINT128

	// if integer is negative, use the absolute value
	if (uint32(x) & SIGNMASK32) == SIGNMASK32 {
		res.w[1] = 0xb040000000000000
		res.w[0] = uint64(^uint32(x) + 1) // 2's complement of x
	} else {
		res.w[1] = 0x3040000000000000
		res.w[0] = uint64(uint32(x))
	}
	return res
}

// Bid128FromUint32 converts uint32 to BID128.
// Ported mechanically from Intel bid_from_int.c: bid128_from_uint32.
func Bid128FromUint32(x uint32) BID_UINT128 {
	var res BID_UINT128

	res.w[1] = 0x3040000000000000
	res.w[0] = uint64(x)
	return res
}

// Bid128FromInt64 converts int64 to BID128.
// Ported mechanically from Intel bid_from_int.c: bid128_from_int64.
func Bid128FromInt64(x int64) BID_UINT128 {
	var res BID_UINT128

	// if integer is negative, use the absolute value
	if (uint64(x) & SIGNMASK64) == SIGNMASK64 {
		res.w[1] = 0xb040000000000000
		res.w[0] = ^uint64(x) + 1 // 2's complement of x
	} else {
		res.w[1] = 0x3040000000000000
		res.w[0] = uint64(x)
	}
	return res
}

// Bid128FromUint64 converts uint64 to BID128.
// Ported mechanically from Intel bid_from_int.c: bid128_from_uint64.
func Bid128FromUint64(x uint64) BID_UINT128 {
	var res BID_UINT128

	res.w[1] = 0x3040000000000000
	res.w[0] = x
	return res
}
