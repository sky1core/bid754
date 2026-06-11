// Ported from: Intel bid_from_int.c (bid32 section)
// Mechanical translation - all logic preserved exactly.

package bidgo

const BID32_SIG_MAX = 9999999 // 10^7 - 1

// Bid32FromInt32 is ported mechanically from bid_from_int.c: bid32_from_int32.
func Bid32FromInt32(x int32, rnd_mode int) (uint32, uint32) {
	var res uint32
	var res64 uint64
	var x_sign uint32
	var C uint32
	var q, ind uint32
	var incr_exp int
	var is_midpoint_lt_even, is_midpoint_gt_even int
	var is_inexact_lt_midpoint, is_inexact_gt_midpoint int
	var pfpsf uint32

	x_sign = uint32(x) & MASK_SIGN32
	if x_sign != 0 {
		C = ^uint32(x) + 1
	} else {
		C = uint32(x)
	}
	if C <= BID32_SIG_MAX {
		if C < 0x00800000 {
			res = x_sign | 0x32800000 | C
		} else {
			res = x_sign | 0x6ca00000 | (C & 0x001fffff)
		}
	} else {
		if C < 0x05f5e100 {
			q = 8
			ind = 1
		} else if C < 0x3b9aca00 {
			q = 9
			ind = 2
		} else {
			q = 10
			ind = 3
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

// Bid32FromUint32 is ported mechanically from bid_from_int.c: bid32_from_uint32.
func Bid32FromUint32(x uint32, rnd_mode int) (uint32, uint32) {
	var res uint32
	var res64 uint64
	var C uint32
	var q, ind uint32
	var incr_exp int
	var is_midpoint_lt_even, is_midpoint_gt_even int
	var is_inexact_lt_midpoint, is_inexact_gt_midpoint int
	var pfpsf uint32

	C = x
	if C <= BID32_SIG_MAX {
		if C < 0x00800000 {
			res = 0x32800000 | C
		} else {
			res = 0x6ca00000 | (C & 0x001fffff)
		}
	} else {
		if C < 0x05f5e100 {
			q = 8
			ind = 1
		} else if C < 0x3b9aca00 {
			q = 9
			ind = 2
		} else {
			q = 10
			ind = 3
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
