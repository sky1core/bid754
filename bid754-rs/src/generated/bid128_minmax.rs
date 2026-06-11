// Auto-generated from bid128_minmax.go by go2rs. Do not edit.

use super::prelude::*;

pub fn bid128_minnum(mut x: BID_UINT128, mut y: BID_UINT128, pfpsf: &mut u32) -> BID_UINT128 {
    let mut res: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut diff: i64 = 0;
    let mut sig_x: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut sig_y: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut sig_n_prime192: BID_UINT192 = BID_UINT192 { w: [0, 0, 0] };
    let mut sig_n_prime256: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut x_is_zero: u8 = 0;
    let mut y_is_zero: u8 = 0;
    if ((x.w[1] & 0x7c00000000000000) == 0x7c00000000000000) {
        x.w[1] = (x.w[1] & 0xfe003fffffffffff);
        if ((((x.w[1] & 0x00003fffffffffff) > 0x0000314dc6448d93)) || (((((x.w[1] & 0x00003fffffffffff) == 0x0000314dc6448d93)) && (x.w[0] > 0x38c15b09ffffffff)))) {
            x.w[1] = (x.w[1] & 0xffffc00000000000);
            x.w[0] = 0x0;
        }
    } else if ((x.w[1] & 0x7c00000000000000) == 0x7800000000000000) {
        x.w[1] = (x.w[1] & (0x8000000000000000 | 0x7800000000000000));
        x.w[0] = 0x0;
    } else {
        if ((x.w[1] & 0x6000000000000000) == 0x6000000000000000) {
            x.w[1] = ((x.w[1] & 0x8000000000000000) | ((((go_checked_shl_u64(x.w[1], go_shift_count_u64((2) as u64)))) & 0x7ffe000000000000)));
            x.w[0] = 0x0;
        } else {
            if (((x.w[1] & 0x1ffffffffffff) > 0x0001ed09bead87c0) || ((((x.w[1] & 0x1ffffffffffff) == 0x0001ed09bead87c0) && (x.w[0] > 0x378d8e63ffffffff)))) {
                x.w[1] = ((x.w[1] & 0x8000000000000000) | (x.w[1] & 0x7ffe000000000000));
                x.w[0] = 0x0;
            } else {
            }
        }
    }
    if ((y.w[1] & 0x7c00000000000000) == 0x7c00000000000000) {
        y.w[1] = (y.w[1] & 0xfe003fffffffffff);
        if ((((y.w[1] & 0x00003fffffffffff) > 0x0000314dc6448d93)) || (((((y.w[1] & 0x00003fffffffffff) == 0x0000314dc6448d93)) && (y.w[0] > 0x38c15b09ffffffff)))) {
            y.w[1] = (y.w[1] & 0xffffc00000000000);
            y.w[0] = 0x0;
        }
    } else if ((y.w[1] & 0x7c00000000000000) == 0x7800000000000000) {
        y.w[1] = (y.w[1] & (0x8000000000000000 | 0x7800000000000000));
        y.w[0] = 0x0;
    } else {
        if ((y.w[1] & 0x6000000000000000) == 0x6000000000000000) {
            y.w[1] = ((y.w[1] & 0x8000000000000000) | ((((go_checked_shl_u64(y.w[1], go_shift_count_u64((2) as u64)))) & 0x7ffe000000000000)));
            y.w[0] = 0x0;
        } else {
            if (((y.w[1] & 0x1ffffffffffff) > 0x0001ed09bead87c0) || ((((y.w[1] & 0x1ffffffffffff) == 0x0001ed09bead87c0) && (y.w[0] > 0x378d8e63ffffffff)))) {
                y.w[1] = ((y.w[1] & 0x8000000000000000) | (y.w[1] & 0x7ffe000000000000));
                y.w[0] = 0x0;
            } else {
            }
        }
    }
    if ((x.w[1] & 0x7c00000000000000) == 0x7c00000000000000) {
        if ((x.w[1] & 0x7e00000000000000) == 0x7e00000000000000) {
            (*pfpsf) |= 1;
            x.w[1] = (x.w[1] & 0xfdffffffffffffff);
            res = x;
        } else {
            if ((y.w[1] & 0x7c00000000000000) == 0x7c00000000000000) {
                if ((y.w[1] & 0x7e00000000000000) == 0x7e00000000000000) {
                    (*pfpsf) |= 1;
                }
                res = x;
            } else {
                res = y;
            }
        }
        return res;
    } else if ((y.w[1] & 0x7c00000000000000) == 0x7c00000000000000) {
        if ((y.w[1] & 0x7e00000000000000) == 0x7e00000000000000) {
            (*pfpsf) |= 1;
            y.w[1] = (y.w[1] & 0xfdffffffffffffff);
            res = y;
        } else {
            res = x;
        }
        return res;
    }
    if ((x.w[0] == y.w[0]) && (x.w[1] == y.w[1])) {
        res = x;
        return res;
    }
    if ((x.w[1] & 0x7800000000000000) == 0x7800000000000000) {
        if ((x.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = x;
        } else {
            res = y;
        }
        return res;
    } else if ((y.w[1] & 0x7800000000000000) == 0x7800000000000000) {
        if ((y.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = y;
        } else {
            res = x;
        }
        return res;
    }
    sig_x.w[1] = (x.w[1] & 0x0001ffffffffffff);
    sig_x.w[0] = x.w[0];
    exp_x = ((((go_checked_shr_u64(x.w[1], go_shift_count_u64((49) as u64)))) & 0x000000000003fff) as i64);
    exp_y = ((((go_checked_shr_u64(y.w[1], go_shift_count_u64((49) as u64)))) & 0x0000000000003fff) as i64);
    sig_y.w[1] = (y.w[1] & 0x0001ffffffffffff);
    sig_y.w[0] = y.w[0];
    if ((sig_x.w[1] == 0) && (sig_x.w[0] == 0)) {
        x_is_zero = 1;
    }
    if ((sig_y.w[1] == 0) && (sig_y.w[0] == 0)) {
        y_is_zero = 1;
    }
    if ((x_is_zero != 0) && (y_is_zero != 0)) {
        res = x;
        return res;
    } else if (x_is_zero != 0) {
        if ((y.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = y;
        } else {
            res = x;
        }
        return res;
    } else if (y_is_zero != 0) {
        if ((x.w[1] & 0x8000000000000000) != 0x8000000000000000) {
            res = y;
        } else {
            res = x;
        }
        return res;
    }
    if ((((x.w[1] ^ y.w[1]) & 0x8000000000000000)) == 0x8000000000000000) {
        if ((y.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = y;
        } else {
            res = x;
        }
        return res;
    }
    if (exp_y == exp_x) {
        if ((((sig_x.w[1] > sig_y.w[1]) || (((sig_x.w[1] == sig_y.w[1]) && (sig_x.w[0] >= sig_y.w[0]))))) != (((x.w[1] & 0x8000000000000000) == 0x8000000000000000))) {
            res = y;
        } else {
            res = x;
        }
        return res;
    }
    if (((sig_x.w[1] >= sig_y.w[1]) && (sig_x.w[0] >= sig_y.w[0])) && (exp_x > exp_y)) {
        if ((x.w[1] & 0x8000000000000000) != 0x8000000000000000) {
            res = y;
        } else {
            res = x;
        }
        return res;
    }
    if (((sig_x.w[1] <= sig_y.w[1]) && (sig_x.w[0] <= sig_y.w[0])) && (exp_x < exp_y)) {
        if ((x.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = y;
        } else {
            res = x;
        }
        return res;
    }
    diff = (exp_x.wrapping_sub(exp_y));
    if (diff > 0) {
        if (diff > 33) {
            if ((x.w[1] & 0x8000000000000000) != 0x8000000000000000) {
                res = y;
            } else {
                res = x;
            }
            return res;
        }
        if (diff > 19) {
            sig_n_prime256 = __mul_128x128_to_256(sig_x, bid_ten2k128[(diff.wrapping_sub(20)) as usize]);
            if (((((((sig_n_prime256.w[3] > 0) || (sig_n_prime256.w[2] > 0))) || (sig_n_prime256.w[1] > sig_y.w[1])) || (((sig_n_prime256.w[1] == sig_y.w[1]) && (sig_n_prime256.w[0] > sig_y.w[0]))))) != (((y.w[1] & 0x8000000000000000) == 0x8000000000000000))) {
                res = y;
            } else {
                res = x;
            }
            return res;
        }
        sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[diff as usize], sig_x);
        if (((((sig_n_prime192.w[2] > 0) || (sig_n_prime192.w[1] > sig_y.w[1])) || (((sig_n_prime192.w[1] == sig_y.w[1]) && (sig_n_prime192.w[0] > sig_y.w[0]))))) != (((y.w[1] & 0x8000000000000000) == 0x8000000000000000))) {
            res = y;
        } else {
            res = x;
        }
        return res;
    }
    diff = (exp_y.wrapping_sub(exp_x));
    if (diff > 33) {
        if ((x.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = y;
        } else {
            res = x;
        }
        return res;
    }
    if (diff > 19) {
        sig_n_prime256 = __mul_128x128_to_256(sig_y, bid_ten2k128[(diff.wrapping_sub(20)) as usize]);
        if (((((sig_n_prime256.w[3] != 0) || (sig_n_prime256.w[2] != 0)) || (((sig_n_prime256.w[1] > sig_x.w[1]) || (((sig_n_prime256.w[1] == sig_x.w[1]) && (sig_n_prime256.w[0] > sig_x.w[0]))))))) != (((x.w[1] & 0x8000000000000000) == 0x8000000000000000))) {
            res = x;
        } else {
            res = y;
        }
        return res;
    }
    sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[diff as usize], sig_y);
    if ((((sig_n_prime192.w[2] != 0) || (((sig_n_prime192.w[1] > sig_x.w[1]) || (((sig_n_prime192.w[1] == sig_x.w[1]) && (sig_n_prime192.w[0] > sig_x.w[0]))))))) != (((y.w[1] & 0x8000000000000000) == 0x8000000000000000))) {
        res = x;
    } else {
        res = y;
    }
    return res;
}

pub fn bid128_minnum_mag(mut x: BID_UINT128, mut y: BID_UINT128, pfpsf: &mut u32) -> BID_UINT128 {
    let mut res: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut diff: i64 = 0;
    let mut sig_x: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut sig_y: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut sig_n_prime192: BID_UINT192 = BID_UINT192 { w: [0, 0, 0] };
    let mut sig_n_prime256: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    if ((x.w[1] & 0x7c00000000000000) == 0x7c00000000000000) {
        x.w[1] = (x.w[1] & 0xfe003fffffffffff);
        if ((((x.w[1] & 0x00003fffffffffff) > 0x0000314dc6448d93)) || (((((x.w[1] & 0x00003fffffffffff) == 0x0000314dc6448d93)) && (x.w[0] > 0x38c15b09ffffffff)))) {
            x.w[1] = (x.w[1] & 0xffffc00000000000);
            x.w[0] = 0x0;
        }
    } else if ((x.w[1] & 0x7c00000000000000) == 0x7800000000000000) {
        x.w[1] = (x.w[1] & (0x8000000000000000 | 0x7800000000000000));
        x.w[0] = 0x0;
    } else {
        if ((x.w[1] & 0x6000000000000000) == 0x6000000000000000) {
            x.w[1] = ((x.w[1] & 0x8000000000000000) | ((((go_checked_shl_u64(x.w[1], go_shift_count_u64((2) as u64)))) & 0x7ffe000000000000)));
            x.w[0] = 0x0;
        } else {
            if (((x.w[1] & 0x1ffffffffffff) > 0x0001ed09bead87c0) || ((((x.w[1] & 0x1ffffffffffff) == 0x0001ed09bead87c0) && (x.w[0] > 0x378d8e63ffffffff)))) {
                x.w[1] = ((x.w[1] & 0x8000000000000000) | (x.w[1] & 0x7ffe000000000000));
                x.w[0] = 0x0;
            } else {
            }
        }
    }
    if ((y.w[1] & 0x7c00000000000000) == 0x7c00000000000000) {
        y.w[1] = (y.w[1] & 0xfe003fffffffffff);
        if ((((y.w[1] & 0x00003fffffffffff) > 0x0000314dc6448d93)) || (((((y.w[1] & 0x00003fffffffffff) == 0x0000314dc6448d93)) && (y.w[0] > 0x38c15b09ffffffff)))) {
            y.w[1] = (y.w[1] & 0xffffc00000000000);
            y.w[0] = 0x0;
        }
    } else if ((y.w[1] & 0x7c00000000000000) == 0x7800000000000000) {
        y.w[1] = (y.w[1] & (0x8000000000000000 | 0x7800000000000000));
        y.w[0] = 0x0;
    } else {
        if ((y.w[1] & 0x6000000000000000) == 0x6000000000000000) {
            y.w[1] = ((y.w[1] & 0x8000000000000000) | ((((go_checked_shl_u64(y.w[1], go_shift_count_u64((2) as u64)))) & 0x7ffe000000000000)));
            y.w[0] = 0x0;
        } else {
            if (((y.w[1] & 0x1ffffffffffff) > 0x0001ed09bead87c0) || ((((y.w[1] & 0x1ffffffffffff) == 0x0001ed09bead87c0) && (y.w[0] > 0x378d8e63ffffffff)))) {
                y.w[1] = ((y.w[1] & 0x8000000000000000) | (y.w[1] & 0x7ffe000000000000));
                y.w[0] = 0x0;
            } else {
            }
        }
    }
    if ((x.w[1] & 0x7c00000000000000) == 0x7c00000000000000) {
        if ((x.w[1] & 0x7e00000000000000) == 0x7e00000000000000) {
            (*pfpsf) |= 1;
            x.w[1] = (x.w[1] & 0xfdffffffffffffff);
            res = x;
        } else {
            if ((y.w[1] & 0x7c00000000000000) == 0x7c00000000000000) {
                if ((y.w[1] & 0x7e00000000000000) == 0x7e00000000000000) {
                    (*pfpsf) |= 1;
                }
                res = x;
            } else {
                res = y;
            }
        }
        return res;
    } else if ((y.w[1] & 0x7c00000000000000) == 0x7c00000000000000) {
        if ((y.w[1] & 0x7e00000000000000) == 0x7e00000000000000) {
            (*pfpsf) |= 1;
            y.w[1] = (y.w[1] & 0xfdffffffffffffff);
            res = y;
        } else {
            res = x;
        }
        return res;
    }
    if ((x.w[0] == y.w[0]) && (x.w[1] == y.w[1])) {
        res = y;
        return res;
    }
    if ((x.w[1] & 0x7800000000000000) == 0x7800000000000000) {
        if (((x.w[1] & 0x8000000000000000) == 0x8000000000000000) && ((y.w[1] & 0x7800000000000000) == 0x7800000000000000)) {
            res = x;
        } else {
            res = y;
        }
        return res;
    } else if ((y.w[1] & 0x7800000000000000) == 0x7800000000000000) {
        res = x;
        return res;
    }
    sig_x.w[1] = (x.w[1] & 0x0001ffffffffffff);
    sig_x.w[0] = x.w[0];
    exp_x = ((((go_checked_shr_u64(x.w[1], go_shift_count_u64((49) as u64)))) & 0x000000000003fff) as i64);
    exp_y = ((((go_checked_shr_u64(y.w[1], go_shift_count_u64((49) as u64)))) & 0x0000000000003fff) as i64);
    sig_y.w[1] = (y.w[1] & 0x0001ffffffffffff);
    sig_y.w[0] = y.w[0];
    if ((sig_x.w[1] == 0) && (sig_x.w[0] == 0)) {
        res = x;
        return res;
    }
    if ((sig_y.w[1] == 0) && (sig_y.w[0] == 0)) {
        res = y;
        return res;
    }
    if (((exp_y == exp_x) && (sig_x.w[1] == sig_y.w[1])) && (sig_x.w[0] == sig_y.w[0])) {
        if ((x.w[1] & 0x8000000000000000) != 0) {
            res = x;
            return res;
        } else {
            res = y;
            return res;
        }
    } else if ((((((sig_x.w[1] > sig_y.w[1]) || (((sig_x.w[1] == sig_y.w[1]) && (sig_x.w[0] > sig_y.w[0]))))) && (exp_x == exp_y))) || (((((sig_x.w[1] > sig_y.w[1]) || (((sig_x.w[1] == sig_y.w[1]) && (sig_x.w[0] >= sig_y.w[0]))))) && (exp_x > exp_y)))) {
        res = y;
        return res;
    } else if ((((((sig_y.w[1] > sig_x.w[1]) || (((sig_y.w[1] == sig_x.w[1]) && (sig_y.w[0] > sig_x.w[0]))))) && (exp_y == exp_x))) || (((((sig_y.w[1] > sig_x.w[1]) || (((sig_y.w[1] == sig_x.w[1]) && (sig_y.w[0] >= sig_x.w[0]))))) && (exp_y > exp_x)))) {
        res = x;
        return res;
    } else {
    }
    diff = (exp_x.wrapping_sub(exp_y));
    if (diff > 0) {
        if (diff > 33) {
            res = y;
            return res;
        }
        if (diff > 19) {
            sig_n_prime256 = __mul_128x128_to_256(sig_x, bid_ten2k128[(diff.wrapping_sub(20)) as usize]);
            if ((((sig_n_prime256.w[3] == 0) && (sig_n_prime256.w[2] == 0)) && (sig_n_prime256.w[1] == sig_y.w[1])) && (sig_n_prime256.w[0] == sig_y.w[0])) {
                if ((y.w[1] & 0x8000000000000000) == 0x8000000000000000) {
                    res = y;
                } else {
                    res = x;
                }
                return res;
            }
            if (((((sig_n_prime256.w[3] > 0) || (sig_n_prime256.w[2] > 0))) || (sig_n_prime256.w[1] > sig_y.w[1])) || (((sig_n_prime256.w[1] == sig_y.w[1]) && (sig_n_prime256.w[0] > sig_y.w[0])))) {
                res = y;
            } else {
                res = x;
            }
            return res;
        }
        sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[diff as usize], sig_x);
        if (((sig_n_prime192.w[2] == 0) && (sig_n_prime192.w[1] == sig_y.w[1])) && (sig_n_prime192.w[0] == sig_y.w[0])) {
            if ((y.w[1] & 0x8000000000000000) == 0x8000000000000000) {
                res = y;
            } else {
                res = x;
            }
            return res;
        }
        if (((sig_n_prime192.w[2] > 0) || (sig_n_prime192.w[1] > sig_y.w[1])) || (((sig_n_prime192.w[1] == sig_y.w[1]) && (sig_n_prime192.w[0] > sig_y.w[0])))) {
            res = y;
        } else {
            res = x;
        }
        return res;
    }
    diff = (exp_y.wrapping_sub(exp_x));
    if (diff > 33) {
        res = x;
        return res;
    }
    if (diff > 19) {
        sig_n_prime256 = __mul_128x128_to_256(sig_y, bid_ten2k128[(diff.wrapping_sub(20)) as usize]);
        if ((((sig_n_prime256.w[3] == 0) && (sig_n_prime256.w[2] == 0)) && (sig_n_prime256.w[1] == sig_x.w[1])) && (sig_n_prime256.w[0] == sig_x.w[0])) {
            if ((y.w[1] & 0x8000000000000000) == 0x8000000000000000) {
                res = y;
            } else {
                res = x;
            }
            return res;
        }
        if (((sig_n_prime256.w[3] == 0) && (sig_n_prime256.w[2] == 0)) && (((sig_n_prime256.w[1] < sig_x.w[1]) || (((sig_n_prime256.w[1] == sig_x.w[1]) && (sig_n_prime256.w[0] < sig_x.w[0])))))) {
            res = y;
        } else {
            res = x;
        }
        return res;
    }
    sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[diff as usize], sig_y);
    if (((sig_n_prime192.w[2] == 0) && (sig_n_prime192.w[1] == sig_x.w[1])) && (sig_n_prime192.w[0] == sig_x.w[0])) {
        if ((y.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = y;
        } else {
            res = x;
        }
        return res;
    }
    if ((sig_n_prime192.w[2] == 0) && (((sig_n_prime192.w[1] < sig_x.w[1]) || (((sig_n_prime192.w[1] == sig_x.w[1]) && (sig_n_prime192.w[0] < sig_x.w[0])))))) {
        res = y;
    } else {
        res = x;
    }
    return res;
}

pub fn bid128_maxnum(mut x: BID_UINT128, mut y: BID_UINT128, pfpsf: &mut u32) -> BID_UINT128 {
    let mut res: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut diff: i64 = 0;
    let mut sig_x: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut sig_y: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut sig_n_prime192: BID_UINT192 = BID_UINT192 { w: [0, 0, 0] };
    let mut sig_n_prime256: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut x_is_zero: u8 = 0;
    let mut y_is_zero: u8 = 0;
    if ((x.w[1] & 0x7c00000000000000) == 0x7c00000000000000) {
        x.w[1] = (x.w[1] & 0xfe003fffffffffff);
        if ((((x.w[1] & 0x00003fffffffffff) > 0x0000314dc6448d93)) || (((((x.w[1] & 0x00003fffffffffff) == 0x0000314dc6448d93)) && (x.w[0] > 0x38c15b09ffffffff)))) {
            x.w[1] = (x.w[1] & 0xffffc00000000000);
            x.w[0] = 0x0;
        }
    } else if ((x.w[1] & 0x7c00000000000000) == 0x7800000000000000) {
        x.w[1] = (x.w[1] & (0x8000000000000000 | 0x7800000000000000));
        x.w[0] = 0x0;
    } else {
        if ((x.w[1] & 0x6000000000000000) == 0x6000000000000000) {
            x.w[1] = ((x.w[1] & 0x8000000000000000) | ((((go_checked_shl_u64(x.w[1], go_shift_count_u64((2) as u64)))) & 0x7ffe000000000000)));
            x.w[0] = 0x0;
        } else {
            if (((x.w[1] & 0x1ffffffffffff) > 0x0001ed09bead87c0) || ((((x.w[1] & 0x1ffffffffffff) == 0x0001ed09bead87c0) && (x.w[0] > 0x378d8e63ffffffff)))) {
                x.w[1] = ((x.w[1] & 0x8000000000000000) | (x.w[1] & 0x7ffe000000000000));
                x.w[0] = 0x0;
            }
        }
    }
    if ((y.w[1] & 0x7c00000000000000) == 0x7c00000000000000) {
        y.w[1] = (y.w[1] & 0xfe003fffffffffff);
        if ((((y.w[1] & 0x00003fffffffffff) > 0x0000314dc6448d93)) || (((((y.w[1] & 0x00003fffffffffff) == 0x0000314dc6448d93)) && (y.w[0] > 0x38c15b09ffffffff)))) {
            y.w[1] = (y.w[1] & 0xffffc00000000000);
            y.w[0] = 0x0;
        }
    } else if ((y.w[1] & 0x7c00000000000000) == 0x7800000000000000) {
        y.w[1] = (y.w[1] & (0x8000000000000000 | 0x7800000000000000));
        y.w[0] = 0x0;
    } else {
        if ((y.w[1] & 0x6000000000000000) == 0x6000000000000000) {
            y.w[1] = ((y.w[1] & 0x8000000000000000) | ((((go_checked_shl_u64(y.w[1], go_shift_count_u64((2) as u64)))) & 0x7ffe000000000000)));
            y.w[0] = 0x0;
        } else {
            if (((y.w[1] & 0x1ffffffffffff) > 0x0001ed09bead87c0) || ((((y.w[1] & 0x1ffffffffffff) == 0x0001ed09bead87c0) && (y.w[0] > 0x378d8e63ffffffff)))) {
                y.w[1] = ((y.w[1] & 0x8000000000000000) | (y.w[1] & 0x7ffe000000000000));
                y.w[0] = 0x0;
            }
        }
    }
    if ((x.w[1] & 0x7c00000000000000) == 0x7c00000000000000) {
        if ((x.w[1] & 0x7e00000000000000) == 0x7e00000000000000) {
            (*pfpsf) |= 1;
            x.w[1] = (x.w[1] & 0xfdffffffffffffff);
            res = x;
        } else {
            if ((y.w[1] & 0x7c00000000000000) == 0x7c00000000000000) {
                if ((y.w[1] & 0x7e00000000000000) == 0x7e00000000000000) {
                    (*pfpsf) |= 1;
                }
                res = x;
            } else {
                res = y;
            }
        }
        return res;
    } else if ((y.w[1] & 0x7c00000000000000) == 0x7c00000000000000) {
        if ((y.w[1] & 0x7e00000000000000) == 0x7e00000000000000) {
            (*pfpsf) |= 1;
            y.w[1] = (y.w[1] & 0xfdffffffffffffff);
            res = y;
        } else {
            res = x;
        }
        return res;
    }
    if ((x.w[0] == y.w[0]) && (x.w[1] == y.w[1])) {
        res = x;
        return res;
    }
    if ((x.w[1] & 0x7800000000000000) == 0x7800000000000000) {
        if ((x.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = y;
        } else {
            res = x;
        }
        return res;
    } else if ((y.w[1] & 0x7800000000000000) == 0x7800000000000000) {
        if ((y.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = x;
        } else {
            res = y;
        }
        return res;
    }
    sig_x.w[1] = (x.w[1] & 0x0001ffffffffffff);
    sig_x.w[0] = x.w[0];
    exp_x = ((((go_checked_shr_u64(x.w[1], go_shift_count_u64((49) as u64)))) & 0x000000000003fff) as i64);
    exp_y = ((((go_checked_shr_u64(y.w[1], go_shift_count_u64((49) as u64)))) & 0x0000000000003fff) as i64);
    sig_y.w[1] = (y.w[1] & 0x0001ffffffffffff);
    sig_y.w[0] = y.w[0];
    if ((sig_x.w[1] == 0) && (sig_x.w[0] == 0)) {
        x_is_zero = 1;
    }
    if ((sig_y.w[1] == 0) && (sig_y.w[0] == 0)) {
        y_is_zero = 1;
    }
    if ((x_is_zero != 0) && (y_is_zero != 0)) {
        res = x;
        return res;
    } else if (x_is_zero != 0) {
        if ((y.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = x;
        } else {
            res = y;
        }
        return res;
    } else if (y_is_zero != 0) {
        if ((x.w[1] & 0x8000000000000000) != 0x8000000000000000) {
            res = x;
        } else {
            res = y;
        }
        return res;
    }
    if ((((x.w[1] ^ y.w[1]) & 0x8000000000000000)) == 0x8000000000000000) {
        if ((y.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = x;
        } else {
            res = y;
        }
        return res;
    }
    if (exp_y == exp_x) {
        if ((((sig_x.w[1] > sig_y.w[1]) || (((sig_x.w[1] == sig_y.w[1]) && (sig_x.w[0] >= sig_y.w[0]))))) != (((x.w[1] & 0x8000000000000000) == 0x8000000000000000))) {
            res = x;
        } else {
            res = y;
        }
        return res;
    }
    if ((((sig_x.w[1] > sig_y.w[1]) || (((sig_x.w[1] == sig_y.w[1]) && (sig_x.w[0] > sig_y.w[0]))))) && (exp_x >= exp_y)) {
        if ((x.w[1] & 0x8000000000000000) != 0x8000000000000000) {
            res = x;
        } else {
            res = y;
        }
        return res;
    }
    if ((((sig_x.w[1] < sig_y.w[1]) || (((sig_x.w[1] == sig_y.w[1]) && (sig_x.w[0] < sig_y.w[0]))))) && (exp_x <= exp_y)) {
        if ((x.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = x;
        } else {
            res = y;
        }
        return res;
    }
    diff = (exp_x.wrapping_sub(exp_y));
    if (diff > 0) {
        if (diff > 33) {
            if ((x.w[1] & 0x8000000000000000) != 0x8000000000000000) {
                res = x;
            } else {
                res = y;
            }
            return res;
        }
        if (diff > 19) {
            sig_n_prime256 = __mul_128x128_to_256(sig_x, bid_ten2k128[(diff.wrapping_sub(20)) as usize]);
            if (((((((sig_n_prime256.w[3] > 0) || (sig_n_prime256.w[2] > 0))) || (sig_n_prime256.w[1] > sig_y.w[1])) || (((sig_n_prime256.w[1] == sig_y.w[1]) && (sig_n_prime256.w[0] > sig_y.w[0]))))) != (((y.w[1] & 0x8000000000000000) == 0x8000000000000000))) {
                res = x;
            } else {
                res = y;
            }
            return res;
        }
        sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[diff as usize], sig_x);
        if (((((sig_n_prime192.w[2] > 0) || (sig_n_prime192.w[1] > sig_y.w[1])) || (((sig_n_prime192.w[1] == sig_y.w[1]) && (sig_n_prime192.w[0] > sig_y.w[0]))))) != (((y.w[1] & 0x8000000000000000) == 0x8000000000000000))) {
            res = x;
        } else {
            res = y;
        }
        return res;
    }
    diff = (exp_y.wrapping_sub(exp_x));
    if (diff > 33) {
        if ((x.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = x;
        } else {
            res = y;
        }
        return res;
    }
    if (diff > 19) {
        sig_n_prime256 = __mul_128x128_to_256(sig_y, bid_ten2k128[(diff.wrapping_sub(20)) as usize]);
        if (((((sig_n_prime256.w[3] != 0) || (sig_n_prime256.w[2] != 0)) || (((sig_n_prime256.w[1] > sig_x.w[1]) || (((sig_n_prime256.w[1] == sig_x.w[1]) && (sig_n_prime256.w[0] > sig_x.w[0]))))))) != (((x.w[1] & 0x8000000000000000) != 0x8000000000000000))) {
            res = x;
        } else {
            res = y;
        }
        return res;
    }
    sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[diff as usize], sig_y);
    if ((((sig_n_prime192.w[2] != 0) || (((sig_n_prime192.w[1] > sig_x.w[1]) || (((sig_n_prime192.w[1] == sig_x.w[1]) && (sig_n_prime192.w[0] > sig_x.w[0]))))))) != (((y.w[1] & 0x8000000000000000) != 0x8000000000000000))) {
        res = x;
    } else {
        res = y;
    }
    return res;
}

pub fn bid128_maxnum_mag(mut x: BID_UINT128, mut y: BID_UINT128, pfpsf: &mut u32) -> BID_UINT128 {
    let mut res: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut diff: i64 = 0;
    let mut sig_x: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut sig_y: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut sig_n_prime192: BID_UINT192 = BID_UINT192 { w: [0, 0, 0] };
    let mut sig_n_prime256: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    if ((x.w[1] & 0x7c00000000000000) == 0x7c00000000000000) {
        x.w[1] = (x.w[1] & 0xfe003fffffffffff);
        if ((((x.w[1] & 0x00003fffffffffff) > 0x0000314dc6448d93)) || (((((x.w[1] & 0x00003fffffffffff) == 0x0000314dc6448d93)) && (x.w[0] > 0x38c15b09ffffffff)))) {
            x.w[1] = (x.w[1] & 0xffffc00000000000);
            x.w[0] = 0x0;
        }
    } else if ((x.w[1] & 0x7c00000000000000) == 0x7800000000000000) {
        x.w[1] = (x.w[1] & (0x8000000000000000 | 0x7800000000000000));
        x.w[0] = 0x0;
    } else {
        if ((x.w[1] & 0x6000000000000000) == 0x6000000000000000) {
            x.w[1] = ((x.w[1] & 0x8000000000000000) | ((((go_checked_shl_u64(x.w[1], go_shift_count_u64((2) as u64)))) & 0x7ffe000000000000)));
            x.w[0] = 0x0;
        } else {
            if (((x.w[1] & 0x1ffffffffffff) > 0x0001ed09bead87c0) || ((((x.w[1] & 0x1ffffffffffff) == 0x0001ed09bead87c0) && (x.w[0] > 0x378d8e63ffffffff)))) {
                x.w[1] = ((x.w[1] & 0x8000000000000000) | (x.w[1] & 0x7ffe000000000000));
                x.w[0] = 0x0;
            }
        }
    }
    if ((y.w[1] & 0x7c00000000000000) == 0x7c00000000000000) {
        y.w[1] = (y.w[1] & 0xfe003fffffffffff);
        if ((((y.w[1] & 0x00003fffffffffff) > 0x0000314dc6448d93)) || (((((y.w[1] & 0x00003fffffffffff) == 0x0000314dc6448d93)) && (y.w[0] > 0x38c15b09ffffffff)))) {
            y.w[1] = (y.w[1] & 0xffffc00000000000);
            y.w[0] = 0x0;
        }
    } else if ((y.w[1] & 0x7c00000000000000) == 0x7800000000000000) {
        y.w[1] = (y.w[1] & (0x8000000000000000 | 0x7800000000000000));
        y.w[0] = 0x0;
    } else {
        if ((y.w[1] & 0x6000000000000000) == 0x6000000000000000) {
            y.w[1] = ((y.w[1] & 0x8000000000000000) | ((((go_checked_shl_u64(y.w[1], go_shift_count_u64((2) as u64)))) & 0x7ffe000000000000)));
            y.w[0] = 0x0;
        } else {
            if (((y.w[1] & 0x1ffffffffffff) > 0x0001ed09bead87c0) || ((((y.w[1] & 0x1ffffffffffff) == 0x0001ed09bead87c0) && (y.w[0] > 0x378d8e63ffffffff)))) {
                y.w[1] = ((y.w[1] & 0x8000000000000000) | (y.w[1] & 0x7ffe000000000000));
                y.w[0] = 0x0;
            }
        }
    }
    if ((x.w[1] & 0x7c00000000000000) == 0x7c00000000000000) {
        if ((x.w[1] & 0x7e00000000000000) == 0x7e00000000000000) {
            (*pfpsf) |= 1;
            x.w[1] = (x.w[1] & 0xfdffffffffffffff);
            res = x;
        } else {
            if ((y.w[1] & 0x7c00000000000000) == 0x7c00000000000000) {
                if ((y.w[1] & 0x7e00000000000000) == 0x7e00000000000000) {
                    (*pfpsf) |= 1;
                }
                res = x;
            } else {
                res = y;
            }
        }
        return res;
    } else if ((y.w[1] & 0x7c00000000000000) == 0x7c00000000000000) {
        if ((y.w[1] & 0x7e00000000000000) == 0x7e00000000000000) {
            (*pfpsf) |= 1;
            y.w[1] = (y.w[1] & 0xfdffffffffffffff);
            res = y;
        } else {
            res = x;
        }
        return res;
    }
    if ((x.w[0] == y.w[0]) && (x.w[1] == y.w[1])) {
        res = y;
        return res;
    }
    if ((x.w[1] & 0x7800000000000000) == 0x7800000000000000) {
        if (((x.w[1] & 0x8000000000000000) == 0x8000000000000000) && ((y.w[1] & 0x7800000000000000) == 0x7800000000000000)) {
            res = y;
        } else {
            res = x;
        }
        return res;
    } else if ((y.w[1] & 0x7800000000000000) == 0x7800000000000000) {
        res = y;
        return res;
    }
    sig_x.w[1] = (x.w[1] & 0x0001ffffffffffff);
    sig_x.w[0] = x.w[0];
    exp_x = ((((go_checked_shr_u64(x.w[1], go_shift_count_u64((49) as u64)))) & 0x000000000003fff) as i64);
    exp_y = ((((go_checked_shr_u64(y.w[1], go_shift_count_u64((49) as u64)))) & 0x0000000000003fff) as i64);
    sig_y.w[1] = (y.w[1] & 0x0001ffffffffffff);
    sig_y.w[0] = y.w[0];
    if ((sig_x.w[1] == 0) && (sig_x.w[0] == 0)) {
        res = y;
        return res;
    }
    if ((sig_y.w[1] == 0) && (sig_y.w[0] == 0)) {
        res = x;
        return res;
    }
    if (((exp_y == exp_x) && (sig_x.w[1] == sig_y.w[1])) && (sig_x.w[0] == sig_y.w[0])) {
        if ((x.w[1] & 0x8000000000000000) != 0) {
            res = y;
            return res;
        } else {
            res = x;
            return res;
        }
    } else if ((((((sig_x.w[1] > sig_y.w[1]) || (((sig_x.w[1] == sig_y.w[1]) && (sig_x.w[0] > sig_y.w[0]))))) && (exp_x == exp_y))) || (((((sig_x.w[1] > sig_y.w[1]) || (((sig_x.w[1] == sig_y.w[1]) && (sig_x.w[0] >= sig_y.w[0]))))) && (exp_x > exp_y)))) {
        res = x;
        return res;
    } else if ((((((sig_y.w[1] > sig_x.w[1]) || (((sig_y.w[1] == sig_x.w[1]) && (sig_y.w[0] > sig_x.w[0]))))) && (exp_y == exp_x))) || (((((sig_y.w[1] > sig_x.w[1]) || (((sig_y.w[1] == sig_x.w[1]) && (sig_y.w[0] >= sig_x.w[0]))))) && (exp_y > exp_x)))) {
        res = y;
        return res;
    } else {
    }
    diff = (exp_x.wrapping_sub(exp_y));
    if (diff > 0) {
        if (diff > 33) {
            res = x;
            return res;
        }
        if (diff > 19) {
            sig_n_prime256 = __mul_128x128_to_256(sig_x, bid_ten2k128[(diff.wrapping_sub(20)) as usize]);
            if ((((sig_n_prime256.w[3] == 0) && (sig_n_prime256.w[2] == 0)) && (sig_n_prime256.w[1] == sig_y.w[1])) && (sig_n_prime256.w[0] == sig_y.w[0])) {
                if ((y.w[1] & 0x8000000000000000) == 0x8000000000000000) {
                    res = x;
                } else {
                    res = y;
                }
                return res;
            }
            if (((((sig_n_prime256.w[3] > 0) || (sig_n_prime256.w[2] > 0))) || (sig_n_prime256.w[1] > sig_y.w[1])) || (((sig_n_prime256.w[1] == sig_y.w[1]) && (sig_n_prime256.w[0] > sig_y.w[0])))) {
                res = x;
            } else {
                res = y;
            }
            return res;
        }
        sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[diff as usize], sig_x);
        if (((sig_n_prime192.w[2] == 0) && (sig_n_prime192.w[1] == sig_y.w[1])) && (sig_n_prime192.w[0] == sig_y.w[0])) {
            if ((y.w[1] & 0x8000000000000000) == 0x8000000000000000) {
                res = x;
            } else {
                res = y;
            }
            return res;
        }
        if (((sig_n_prime192.w[2] > 0) || (sig_n_prime192.w[1] > sig_y.w[1])) || (((sig_n_prime192.w[1] == sig_y.w[1]) && (sig_n_prime192.w[0] > sig_y.w[0])))) {
            res = x;
        } else {
            res = y;
        }
        return res;
    }
    diff = (exp_y.wrapping_sub(exp_x));
    if (diff > 33) {
        res = y;
        return res;
    }
    if (diff > 19) {
        sig_n_prime256 = __mul_128x128_to_256(sig_y, bid_ten2k128[(diff.wrapping_sub(20)) as usize]);
        if ((((sig_n_prime256.w[3] == 0) && (sig_n_prime256.w[2] == 0)) && (sig_n_prime256.w[1] == sig_x.w[1])) && (sig_n_prime256.w[0] == sig_x.w[0])) {
            if ((y.w[1] & 0x8000000000000000) == 0x8000000000000000) {
                res = x;
            } else {
                res = y;
            }
            return res;
        }
        if (((sig_n_prime256.w[3] == 0) && (sig_n_prime256.w[2] == 0)) && (((sig_n_prime256.w[1] < sig_x.w[1]) || (((sig_n_prime256.w[1] == sig_x.w[1]) && (sig_n_prime256.w[0] < sig_x.w[0])))))) {
            res = x;
        } else {
            res = y;
        }
        return res;
    }
    sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[diff as usize], sig_y);
    if (((sig_n_prime192.w[2] == 0) && (sig_n_prime192.w[1] == sig_x.w[1])) && (sig_n_prime192.w[0] == sig_x.w[0])) {
        if ((y.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = x;
        } else {
            res = y;
        }
        return res;
    }
    if ((sig_n_prime192.w[2] == 0) && (((sig_n_prime192.w[1] < sig_x.w[1]) || (((sig_n_prime192.w[1] == sig_x.w[1]) && (sig_n_prime192.w[0] < sig_x.w[0])))))) {
        res = x;
    } else {
        res = y;
    }
    return res;
}
