package bidgo

func make_bid_round_const_table_128() [6][36]BID_UINT128 {
	var table [6][36]BID_UINT128

	for i := 1; i < 36; i++ {
		half := __shr_128(bid_power10_table_128[i], 1)
		up := __sub_128_64(bid_power10_table_128[i], 1)

		table[BID_ROUNDING_TO_NEAREST][i] = half
		table[BID_ROUNDING_UP][i] = up
		table[BID_ROUNDING_TIES_AWAY][i] = half
		table[BID_ROUNDING_NEAREST_DOWN][i] = half
	}

	return table
}

// bid_round_const_table_128 mirrors Intel's bid_round_const_table_128.
var bid_round_const_table_128 = make_bid_round_const_table_128()
