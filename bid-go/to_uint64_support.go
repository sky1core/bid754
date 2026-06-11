package bidgo

// bid_ten2k128 is ported mechanically from Intel bid128.c.
var bid_ten2k128 = bid_power10_table_128[20:]

// __mul_128x64_to_128 is ported mechanically from Intel bid_internal.h.
func __mul_128x64_to_128(a uint64, b BID_UINT128) BID_UINT128 {
	return __mul_64x128_short(a, b)
}
