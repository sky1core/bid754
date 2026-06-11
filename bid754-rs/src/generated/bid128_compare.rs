// Auto-generated from bid128_compare.go by go2rs. Do not edit.

use super::prelude::*;

pub fn bid128_quiet_equal(mut x: BID_UINT128, mut y: BID_UINT128) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut exp_t: i64 = 0;
    let mut sig_x: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut sig_y: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut sig_t: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut sig_n_prime192: BID_UINT192 = BID_UINT192 { w: [0, 0, 0] };
    let mut sig_n_prime256: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut pfpsf: u32 = 0;
    let mut x_is_zero: i64 = 0;
    let mut y_is_zero: i64 = 0;
    let mut non_canon_x: i64 = 0;
    let mut non_canon_y: i64 = 0;
    if ((((x.w[1] & 0x7c00000000000000) == 0x7c00000000000000)) || (((y.w[1] & 0x7c00000000000000) == 0x7c00000000000000))) {
        if (((x.w[1] & 0x7e00000000000000) == 0x7e00000000000000) || ((y.w[1] & 0x7e00000000000000) == 0x7e00000000000000)) {
            pfpsf |= 1;
        }
        res = 0;
        return (res, pfpsf);
    }
    if ((x.w[0] == y.w[0]) && (x.w[1] == y.w[1])) {
        res = 1;
        return (res, pfpsf);
    }
    if ((x.w[1] & 0x7800000000000000) == 0x7800000000000000) {
        if ((y.w[1] & 0x7800000000000000) == 0x7800000000000000) {
            res = 0;
            if ((((x.w[1] ^ y.w[1]) & 0x8000000000000000)) != 0x8000000000000000) {
                res = 1;
            }
            return (res, pfpsf);
        } else {
            res = 0;
            return (res, pfpsf);
        }
    }
    if ((y.w[1] & 0x7800000000000000) == 0x7800000000000000) {
        res = 0;
        return (res, pfpsf);
    }
    sig_x.w[1] = (x.w[1] & 0x0001ffffffffffff);
    sig_x.w[0] = x.w[0];
    exp_x = ((((go_checked_shr_u64(x.w[1], go_shift_count_u64((49) as u64)))) & 0x000000000003fff) as i64);
    if (((sig_x.w[1] > 0x0001ed09bead87c0) || (((sig_x.w[1] == 0x0001ed09bead87c0) && (sig_x.w[0] > 0x378d8e63ffffffff)))) || (((x.w[1] & 0x6000000000000000) == 0x6000000000000000))) {
        non_canon_x = 1;
    } else {
        non_canon_x = 0;
    }
    exp_y = ((((go_checked_shr_u64(y.w[1], go_shift_count_u64((49) as u64)))) & 0x0000000000003fff) as i64);
    sig_y.w[1] = (y.w[1] & 0x0001ffffffffffff);
    sig_y.w[0] = y.w[0];
    if (((sig_y.w[1] > 0x0001ed09bead87c0) || (((sig_y.w[1] == 0x0001ed09bead87c0) && (sig_y.w[0] > 0x378d8e63ffffffff)))) || (((y.w[1] & 0x6000000000000000) == 0x6000000000000000))) {
        non_canon_y = 1;
    } else {
        non_canon_y = 0;
    }
    if ((non_canon_x != 0) || (((sig_x.w[1] == 0) && (sig_x.w[0] == 0)))) {
        x_is_zero = 1;
    }
    if ((non_canon_y != 0) || (((sig_y.w[1] == 0) && (sig_y.w[0] == 0)))) {
        y_is_zero = 1;
    }
    if ((x_is_zero != 0) && (y_is_zero != 0)) {
        res = 1;
        return (res, pfpsf);
    } else if ((((x_is_zero != 0) && (y_is_zero == 0))) || (((x_is_zero == 0) && (y_is_zero != 0)))) {
        res = 0;
        return (res, pfpsf);
    }
    if (((x.w[1] ^ y.w[1]) & 0x8000000000000000) != 0) {
        res = 0;
        return (res, pfpsf);
    }
    if (exp_x > exp_y) {
        exp_t = exp_x;
        exp_x = exp_y;
        exp_y = exp_t;
        sig_t.w[1] = sig_x.w[1];
        sig_x.w[1] = sig_y.w[1];
        sig_y.w[1] = sig_t.w[1];
        sig_t.w[0] = sig_x.w[0];
        sig_x.w[0] = sig_y.w[0];
        sig_y.w[0] = sig_t.w[0];
    }
    if ((exp_y.wrapping_sub(exp_x)) > 33) {
        res = 0;
        return (res, pfpsf);
    }
    if ((exp_y.wrapping_sub(exp_x)) > 19) {
        sig_n_prime256 = __mul_128x128_to_256(sig_y, bid_ten2k128[((exp_y.wrapping_sub(exp_x)).wrapping_sub(20)) as usize]);
        res = 0;
        if ((((sig_n_prime256.w[3] == 0) && (sig_n_prime256.w[2] == 0)) && (sig_n_prime256.w[1] == sig_x.w[1])) && (sig_n_prime256.w[0] == sig_x.w[0])) {
            res = 1;
        }
        return (res, pfpsf);
    }
    sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[(exp_y.wrapping_sub(exp_x)) as usize], sig_y);
    res = 0;
    if (((sig_n_prime192.w[2] == 0) && (sig_n_prime192.w[1] == sig_x.w[1])) && (sig_n_prime192.w[0] == sig_x.w[0])) {
        res = 1;
    }
    return (res, pfpsf);
}

pub fn bid128_quiet_greater(mut x: BID_UINT128, mut y: BID_UINT128) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut diff: i64 = 0;
    let mut sig_x: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut sig_y: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut sig_n_prime192: BID_UINT192 = BID_UINT192 { w: [0, 0, 0] };
    let mut sig_n_prime256: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut pfpsf: u32 = 0;
    let mut x_is_zero: i64 = 0;
    let mut y_is_zero: i64 = 0;
    let mut non_canon_x: i64 = 0;
    let mut non_canon_y: i64 = 0;
    if ((((x.w[1] & 0x7c00000000000000) == 0x7c00000000000000)) || (((y.w[1] & 0x7c00000000000000) == 0x7c00000000000000))) {
        if (((x.w[1] & 0x7e00000000000000) == 0x7e00000000000000) || ((y.w[1] & 0x7e00000000000000) == 0x7e00000000000000)) {
            pfpsf |= 1;
        }
        res = 0;
        return (res, pfpsf);
    }
    if ((x.w[0] == y.w[0]) && (x.w[1] == y.w[1])) {
        res = 0;
        return (res, pfpsf);
    }
    if ((x.w[1] & 0x7800000000000000) == 0x7800000000000000) {
        if ((x.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = 0;
            return (res, pfpsf);
        } else {
            res = 0;
            if ((((y.w[1] & 0x7800000000000000) != 0x7800000000000000)) || (((y.w[1] & 0x8000000000000000) == 0x8000000000000000))) {
                res = 1;
            }
            return (res, pfpsf);
        }
    } else if ((y.w[1] & 0x7800000000000000) == 0x7800000000000000) {
        res = 0;
        if ((y.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    sig_x.w[1] = (x.w[1] & 0x0001ffffffffffff);
    sig_x.w[0] = x.w[0];
    exp_x = ((((go_checked_shr_u64(x.w[1], go_shift_count_u64((49) as u64)))) & 0x000000000003fff) as i64);
    if (((sig_x.w[1] > 0x0001ed09bead87c0) || (((sig_x.w[1] == 0x0001ed09bead87c0) && (sig_x.w[0] > 0x378d8e63ffffffff)))) || (((x.w[1] & 0x6000000000000000) == 0x6000000000000000))) {
        non_canon_x = 1;
    } else {
        non_canon_x = 0;
    }
    exp_y = ((((go_checked_shr_u64(y.w[1], go_shift_count_u64((49) as u64)))) & 0x0000000000003fff) as i64);
    sig_y.w[1] = (y.w[1] & 0x0001ffffffffffff);
    sig_y.w[0] = y.w[0];
    if (((sig_y.w[1] > 0x0001ed09bead87c0) || (((sig_y.w[1] == 0x0001ed09bead87c0) && (sig_y.w[0] > 0x378d8e63ffffffff)))) || (((y.w[1] & 0x6000000000000000) == 0x6000000000000000))) {
        non_canon_y = 1;
    } else {
        non_canon_y = 0;
    }
    if ((non_canon_x != 0) || (((sig_x.w[1] == 0) && (sig_x.w[0] == 0)))) {
        x_is_zero = 1;
    }
    if ((non_canon_y != 0) || (((sig_y.w[1] == 0) && (sig_y.w[0] == 0)))) {
        y_is_zero = 1;
    }
    if ((x_is_zero != 0) && (y_is_zero != 0)) {
        res = 0;
        return (res, pfpsf);
    } else if (x_is_zero != 0) {
        res = 0;
        if ((y.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    } else if (y_is_zero != 0) {
        res = 0;
        if ((x.w[1] & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((((x.w[1] ^ y.w[1]) & 0x8000000000000000)) == 0x8000000000000000) {
        res = 0;
        if ((y.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (exp_y == exp_x) {
        res = 0;
        if ((((sig_x.w[1] > sig_y.w[1]) || (((sig_x.w[1] == sig_y.w[1]) && (sig_x.w[0] >= sig_y.w[0]))))) != (((x.w[1] & 0x8000000000000000) == 0x8000000000000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((((sig_x.w[1] > sig_y.w[1]) || (((sig_x.w[1] == sig_y.w[1]) && (sig_x.w[0] > sig_y.w[0]))))) && (exp_x >= exp_y)) {
        res = 0;
        if ((x.w[1] & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((((sig_x.w[1] < sig_y.w[1]) || (((sig_x.w[1] == sig_y.w[1]) && (sig_x.w[0] < sig_y.w[0]))))) && (exp_x <= exp_y)) {
        res = 0;
        if ((x.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    diff = (exp_x.wrapping_sub(exp_y));
    if (diff > 0) {
        if (diff > 33) {
            res = 0;
            if ((x.w[1] & 0x8000000000000000) != 0x8000000000000000) {
                res = 1;
            }
            return (res, pfpsf);
        }
        if (diff > 19) {
            sig_n_prime256 = __mul_128x128_to_256(sig_x, bid_ten2k128[(diff.wrapping_sub(20)) as usize]);
            if ((((sig_n_prime256.w[3] == 0) && (sig_n_prime256.w[2] == 0)) && (sig_n_prime256.w[1] == sig_y.w[1])) && (sig_n_prime256.w[0] == sig_y.w[0])) {
                res = 0;
                return (res, pfpsf);
            }
            res = 0;
            if (((((((sig_n_prime256.w[3] > 0) || (sig_n_prime256.w[2] > 0))) || (sig_n_prime256.w[1] > sig_y.w[1])) || (((sig_n_prime256.w[1] == sig_y.w[1]) && (sig_n_prime256.w[0] > sig_y.w[0]))))) != (((y.w[1] & 0x8000000000000000) == 0x8000000000000000))) {
                res = 1;
            }
            return (res, pfpsf);
        }
        sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[diff as usize], sig_x);
        if (((sig_n_prime192.w[2] == 0) && (sig_n_prime192.w[1] == sig_y.w[1])) && (sig_n_prime192.w[0] == sig_y.w[0])) {
            res = 0;
            return (res, pfpsf);
        }
        res = 0;
        if (((((sig_n_prime192.w[2] > 0) || (sig_n_prime192.w[1] > sig_y.w[1])) || (((sig_n_prime192.w[1] == sig_y.w[1]) && (sig_n_prime192.w[0] > sig_y.w[0]))))) != (((y.w[1] & 0x8000000000000000) == 0x8000000000000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    diff = (exp_y.wrapping_sub(exp_x));
    if (diff > 33) {
        res = 0;
        if ((x.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (diff > 19) {
        sig_n_prime256 = __mul_128x128_to_256(sig_y, bid_ten2k128[(diff.wrapping_sub(20)) as usize]);
        if ((((sig_n_prime256.w[3] == 0) && (sig_n_prime256.w[2] == 0)) && (sig_n_prime256.w[1] == sig_x.w[1])) && (sig_n_prime256.w[0] == sig_x.w[0])) {
            res = 0;
            return (res, pfpsf);
        }
        res = 0;
        if (((((sig_n_prime256.w[3] != 0) || (sig_n_prime256.w[2] != 0)) || (((sig_n_prime256.w[1] > sig_x.w[1]) || (((sig_n_prime256.w[1] == sig_x.w[1]) && (sig_n_prime256.w[0] > sig_x.w[0]))))))) != (((x.w[1] & 0x8000000000000000) != 0x8000000000000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[diff as usize], sig_y);
    if (((sig_n_prime192.w[2] == 0) && (sig_n_prime192.w[1] == sig_x.w[1])) && (sig_n_prime192.w[0] == sig_x.w[0])) {
        res = 0;
        return (res, pfpsf);
    }
    res = 0;
    if ((((sig_n_prime192.w[2] != 0) || (((sig_n_prime192.w[1] > sig_x.w[1]) || (((sig_n_prime192.w[1] == sig_x.w[1]) && (sig_n_prime192.w[0] > sig_x.w[0]))))))) != (((y.w[1] & 0x8000000000000000) != 0x8000000000000000))) {
        res = 1;
    }
    return (res, pfpsf);
}

pub fn bid128_quiet_greater_equal(mut x: BID_UINT128, mut y: BID_UINT128) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut diff: i64 = 0;
    let mut sig_x: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut sig_y: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut sig_n_prime192: BID_UINT192 = BID_UINT192 { w: [0, 0, 0] };
    let mut sig_n_prime256: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut pfpsf: u32 = 0;
    let mut x_is_zero: i64 = 0;
    let mut y_is_zero: i64 = 0;
    let mut non_canon_x: i64 = 0;
    let mut non_canon_y: i64 = 0;
    if ((((x.w[1] & 0x7c00000000000000) == 0x7c00000000000000)) || (((y.w[1] & 0x7c00000000000000) == 0x7c00000000000000))) {
        if (((x.w[1] & 0x7e00000000000000) == 0x7e00000000000000) || ((y.w[1] & 0x7e00000000000000) == 0x7e00000000000000)) {
            pfpsf |= 1;
        }
        res = 0;
        return (res, pfpsf);
    }
    if ((x.w[0] == y.w[0]) && (x.w[1] == y.w[1])) {
        res = 1;
        return (res, pfpsf);
    }
    if ((x.w[1] & 0x7800000000000000) == 0x7800000000000000) {
        if ((x.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = 0;
            if ((((y.w[1] & 0x7800000000000000) == 0x7800000000000000)) && ((y.w[1] & 0x8000000000000000) == 0x8000000000000000)) {
                res = 1;
            }
            return (res, pfpsf);
        } else {
            res = 1;
            return (res, pfpsf);
        }
    } else if ((y.w[1] & 0x7800000000000000) == 0x7800000000000000) {
        res = 0;
        if ((y.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    sig_x.w[1] = (x.w[1] & 0x0001ffffffffffff);
    sig_x.w[0] = x.w[0];
    exp_x = ((((go_checked_shr_u64(x.w[1], go_shift_count_u64((49) as u64)))) & 0x000000000003fff) as i64);
    if (((sig_x.w[1] > 0x0001ed09bead87c0) || (((sig_x.w[1] == 0x0001ed09bead87c0) && (sig_x.w[0] > 0x378d8e63ffffffff)))) || (((x.w[1] & 0x6000000000000000) == 0x6000000000000000))) {
        non_canon_x = 1;
    } else {
        non_canon_x = 0;
    }
    exp_y = ((((go_checked_shr_u64(y.w[1], go_shift_count_u64((49) as u64)))) & 0x0000000000003fff) as i64);
    sig_y.w[1] = (y.w[1] & 0x0001ffffffffffff);
    sig_y.w[0] = y.w[0];
    if (((sig_y.w[1] > 0x0001ed09bead87c0) || (((sig_y.w[1] == 0x0001ed09bead87c0) && (sig_y.w[0] > 0x378d8e63ffffffff)))) || (((y.w[1] & 0x6000000000000000) == 0x6000000000000000))) {
        non_canon_y = 1;
    } else {
        non_canon_y = 0;
    }
    if ((non_canon_x != 0) || (((sig_x.w[1] == 0) && (sig_x.w[0] == 0)))) {
        x_is_zero = 1;
    }
    if ((non_canon_y != 0) || (((sig_y.w[1] == 0) && (sig_y.w[0] == 0)))) {
        y_is_zero = 1;
    }
    if ((x_is_zero != 0) && (y_is_zero != 0)) {
        res = 1;
        return (res, pfpsf);
    } else if (x_is_zero != 0) {
        res = 0;
        if ((y.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    } else if (y_is_zero != 0) {
        res = 0;
        if ((x.w[1] & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((((x.w[1] ^ y.w[1]) & 0x8000000000000000)) == 0x8000000000000000) {
        res = 0;
        if ((y.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (exp_y == exp_x) {
        res = 0;
        if ((((sig_x.w[1] > sig_y.w[1]) || (((sig_x.w[1] == sig_y.w[1]) && (sig_x.w[0] >= sig_y.w[0]))))) != (((x.w[1] & 0x8000000000000000) == 0x8000000000000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (((sig_x.w[1] >= sig_y.w[1]) && (sig_x.w[0] >= sig_y.w[0])) && (exp_x > exp_y)) {
        res = 0;
        if ((x.w[1] & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (((sig_x.w[1] <= sig_y.w[1]) && (sig_x.w[0] <= sig_y.w[0])) && (exp_x < exp_y)) {
        res = 0;
        if ((x.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    diff = (exp_x.wrapping_sub(exp_y));
    if (diff > 0) {
        if (diff > 33) {
            res = 0;
            if ((x.w[1] & 0x8000000000000000) != 0x8000000000000000) {
                res = 1;
            }
            return (res, pfpsf);
        }
        if (diff > 19) {
            sig_n_prime256 = __mul_128x128_to_256(sig_x, bid_ten2k128[(diff.wrapping_sub(20)) as usize]);
            if ((((sig_n_prime256.w[3] == 0) && (sig_n_prime256.w[2] == 0)) && (sig_n_prime256.w[1] == sig_y.w[1])) && (sig_n_prime256.w[0] == sig_y.w[0])) {
                res = 1;
                return (res, pfpsf);
            }
            res = 0;
            if (((((((sig_n_prime256.w[3] > 0) || (sig_n_prime256.w[2] > 0))) || (sig_n_prime256.w[1] > sig_y.w[1])) || (((sig_n_prime256.w[1] == sig_y.w[1]) && (sig_n_prime256.w[0] > sig_y.w[0]))))) != (((y.w[1] & 0x8000000000000000) == 0x8000000000000000))) {
                res = 1;
            }
            return (res, pfpsf);
        }
        sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[diff as usize], sig_x);
        if (((sig_n_prime192.w[2] == 0) && (sig_n_prime192.w[1] == sig_y.w[1])) && (sig_n_prime192.w[0] == sig_y.w[0])) {
            res = 1;
            return (res, pfpsf);
        }
        res = 0;
        if (((((sig_n_prime192.w[2] > 0) || (sig_n_prime192.w[1] > sig_y.w[1])) || (((sig_n_prime192.w[1] == sig_y.w[1]) && (sig_n_prime192.w[0] > sig_y.w[0]))))) != (((y.w[1] & 0x8000000000000000) == 0x8000000000000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    diff = (exp_y.wrapping_sub(exp_x));
    if (diff > 33) {
        res = 0;
        if ((x.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (diff > 19) {
        sig_n_prime256 = __mul_128x128_to_256(sig_y, bid_ten2k128[(diff.wrapping_sub(20)) as usize]);
        if ((((sig_n_prime256.w[3] == 0) && (sig_n_prime256.w[2] == 0)) && (sig_n_prime256.w[1] == sig_x.w[1])) && (sig_n_prime256.w[0] == sig_x.w[0])) {
            res = 1;
            return (res, pfpsf);
        }
        res = 0;
        if (((((sig_n_prime256.w[3] == 0) && (sig_n_prime256.w[2] == 0)) && (((sig_n_prime256.w[1] < sig_x.w[1]) || (((sig_n_prime256.w[1] == sig_x.w[1]) && (sig_n_prime256.w[0] < sig_x.w[0]))))))) != (((x.w[1] & 0x8000000000000000) == 0x8000000000000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[diff as usize], sig_y);
    if (((sig_n_prime192.w[2] == 0) && (sig_n_prime192.w[1] == sig_x.w[1])) && (sig_n_prime192.w[0] == sig_x.w[0])) {
        res = 1;
        return (res, pfpsf);
    }
    res = 0;
    if ((((sig_n_prime192.w[2] == 0) && (((sig_n_prime192.w[1] < sig_x.w[1]) || (((sig_n_prime192.w[1] == sig_x.w[1]) && (sig_n_prime192.w[0] < sig_x.w[0]))))))) != (((y.w[1] & 0x8000000000000000) == 0x8000000000000000))) {
        res = 1;
    }
    return (res, pfpsf);
}

pub fn bid128_quiet_greater_unordered(mut x: BID_UINT128, mut y: BID_UINT128) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut diff: i64 = 0;
    let mut sig_x: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut sig_y: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut sig_n_prime192: BID_UINT192 = BID_UINT192 { w: [0, 0, 0] };
    let mut sig_n_prime256: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut pfpsf: u32 = 0;
    let mut x_is_zero: i64 = 0;
    let mut y_is_zero: i64 = 0;
    let mut non_canon_x: i64 = 0;
    let mut non_canon_y: i64 = 0;
    if ((((x.w[1] & 0x7c00000000000000) == 0x7c00000000000000)) || (((y.w[1] & 0x7c00000000000000) == 0x7c00000000000000))) {
        if (((x.w[1] & 0x7e00000000000000) == 0x7e00000000000000) || ((y.w[1] & 0x7e00000000000000) == 0x7e00000000000000)) {
            pfpsf |= 1;
        }
        res = 1;
        return (res, pfpsf);
    }
    if ((x.w[0] == y.w[0]) && (x.w[1] == y.w[1])) {
        res = 0;
        return (res, pfpsf);
    }
    if ((x.w[1] & 0x7800000000000000) == 0x7800000000000000) {
        if ((x.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = 0;
            return (res, pfpsf);
        } else {
            res = 0;
            if ((((y.w[1] & 0x7800000000000000) != 0x7800000000000000)) || (((y.w[1] & 0x8000000000000000) == 0x8000000000000000))) {
                res = 1;
            }
            return (res, pfpsf);
        }
    } else if ((y.w[1] & 0x7800000000000000) == 0x7800000000000000) {
        res = 0;
        if ((y.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    sig_x.w[1] = (x.w[1] & 0x0001ffffffffffff);
    sig_x.w[0] = x.w[0];
    exp_x = ((((go_checked_shr_u64(x.w[1], go_shift_count_u64((49) as u64)))) & 0x000000000003fff) as i64);
    if (((sig_x.w[1] > 0x0001ed09bead87c0) || (((sig_x.w[1] == 0x0001ed09bead87c0) && (sig_x.w[0] > 0x378d8e63ffffffff)))) || (((x.w[1] & 0x6000000000000000) == 0x6000000000000000))) {
        non_canon_x = 1;
    } else {
        non_canon_x = 0;
    }
    exp_y = ((((go_checked_shr_u64(y.w[1], go_shift_count_u64((49) as u64)))) & 0x0000000000003fff) as i64);
    sig_y.w[1] = (y.w[1] & 0x0001ffffffffffff);
    sig_y.w[0] = y.w[0];
    if (((sig_y.w[1] > 0x0001ed09bead87c0) || (((sig_y.w[1] == 0x0001ed09bead87c0) && (sig_y.w[0] > 0x378d8e63ffffffff)))) || (((y.w[1] & 0x6000000000000000) == 0x6000000000000000))) {
        non_canon_y = 1;
    } else {
        non_canon_y = 0;
    }
    if ((non_canon_x != 0) || (((sig_x.w[1] == 0) && (sig_x.w[0] == 0)))) {
        x_is_zero = 1;
    }
    if ((non_canon_y != 0) || (((sig_y.w[1] == 0) && (sig_y.w[0] == 0)))) {
        y_is_zero = 1;
    }
    if ((x_is_zero != 0) && (y_is_zero != 0)) {
        res = 0;
        return (res, pfpsf);
    } else if (x_is_zero != 0) {
        res = 0;
        if ((y.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    } else if (y_is_zero != 0) {
        res = 0;
        if ((x.w[1] & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((((x.w[1] ^ y.w[1]) & 0x8000000000000000)) == 0x8000000000000000) {
        res = 0;
        if ((y.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (exp_y == exp_x) {
        res = 0;
        if ((((sig_x.w[1] > sig_y.w[1]) || (((sig_x.w[1] == sig_y.w[1]) && (sig_x.w[0] >= sig_y.w[0]))))) != (((x.w[1] & 0x8000000000000000) == 0x8000000000000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (((sig_x.w[1] >= sig_y.w[1]) && (sig_x.w[0] >= sig_y.w[0])) && (exp_x > exp_y)) {
        res = 0;
        if ((x.w[1] & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (((sig_x.w[1] <= sig_y.w[1]) && (sig_x.w[0] <= sig_y.w[0])) && (exp_x < exp_y)) {
        res = 0;
        if ((x.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    diff = (exp_x.wrapping_sub(exp_y));
    if (diff > 0) {
        if (diff > 33) {
            res = 0;
            if ((x.w[1] & 0x8000000000000000) != 0x8000000000000000) {
                res = 1;
            }
            return (res, pfpsf);
        }
        if (diff > 19) {
            sig_n_prime256 = __mul_128x128_to_256(sig_x, bid_ten2k128[(diff.wrapping_sub(20)) as usize]);
            if ((((sig_n_prime256.w[3] == 0) && (sig_n_prime256.w[2] == 0)) && (sig_n_prime256.w[1] == sig_y.w[1])) && (sig_n_prime256.w[0] == sig_y.w[0])) {
                res = 0;
                return (res, pfpsf);
            }
            res = 0;
            if (((((((sig_n_prime256.w[3] > 0) || (sig_n_prime256.w[2] > 0))) || (sig_n_prime256.w[1] > sig_y.w[1])) || (((sig_n_prime256.w[1] == sig_y.w[1]) && (sig_n_prime256.w[0] > sig_y.w[0]))))) != (((y.w[1] & 0x8000000000000000) == 0x8000000000000000))) {
                res = 1;
            }
            return (res, pfpsf);
        }
        sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[diff as usize], sig_x);
        if (((sig_n_prime192.w[2] == 0) && (sig_n_prime192.w[1] == sig_y.w[1])) && (sig_n_prime192.w[0] == sig_y.w[0])) {
            res = 0;
            return (res, pfpsf);
        }
        res = 0;
        if (((((sig_n_prime192.w[2] > 0) || (sig_n_prime192.w[1] > sig_y.w[1])) || (((sig_n_prime192.w[1] == sig_y.w[1]) && (sig_n_prime192.w[0] > sig_y.w[0]))))) != (((y.w[1] & 0x8000000000000000) == 0x8000000000000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    diff = (exp_y.wrapping_sub(exp_x));
    if (diff > 33) {
        res = 0;
        if ((x.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (diff > 19) {
        sig_n_prime256 = __mul_128x128_to_256(sig_y, bid_ten2k128[(diff.wrapping_sub(20)) as usize]);
        if ((((sig_n_prime256.w[3] == 0) && (sig_n_prime256.w[2] == 0)) && (sig_n_prime256.w[1] == sig_x.w[1])) && (sig_n_prime256.w[0] == sig_x.w[0])) {
            res = 0;
            return (res, pfpsf);
        }
        res = 0;
        if (((((sig_n_prime256.w[3] == 0) && (sig_n_prime256.w[2] == 0)) && (((sig_n_prime256.w[1] < sig_x.w[1]) || (((sig_n_prime256.w[1] == sig_x.w[1]) && (sig_n_prime256.w[0] < sig_x.w[0]))))))) != (((x.w[1] & 0x8000000000000000) == 0x8000000000000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[diff as usize], sig_y);
    if (((sig_n_prime192.w[2] == 0) && (sig_n_prime192.w[1] == sig_x.w[1])) && (sig_n_prime192.w[0] == sig_x.w[0])) {
        res = 0;
        return (res, pfpsf);
    }
    res = 0;
    if ((((sig_n_prime192.w[2] == 0) && (((sig_n_prime192.w[1] < sig_x.w[1]) || (((sig_n_prime192.w[1] == sig_x.w[1]) && (sig_n_prime192.w[0] < sig_x.w[0]))))))) != (((y.w[1] & 0x8000000000000000) == 0x8000000000000000))) {
        res = 1;
    }
    return (res, pfpsf);
}

pub fn bid128_quiet_less(mut x: BID_UINT128, mut y: BID_UINT128) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut diff: i64 = 0;
    let mut sig_x: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut sig_y: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut sig_n_prime192: BID_UINT192 = BID_UINT192 { w: [0, 0, 0] };
    let mut sig_n_prime256: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut pfpsf: u32 = 0;
    let mut x_is_zero: i64 = 0;
    let mut y_is_zero: i64 = 0;
    let mut non_canon_x: i64 = 0;
    let mut non_canon_y: i64 = 0;
    if ((((x.w[1] & 0x7c00000000000000) == 0x7c00000000000000)) || (((y.w[1] & 0x7c00000000000000) == 0x7c00000000000000))) {
        if (((x.w[1] & 0x7e00000000000000) == 0x7e00000000000000) || ((y.w[1] & 0x7e00000000000000) == 0x7e00000000000000)) {
            pfpsf |= 1;
        }
        res = 0;
        return (res, pfpsf);
    }
    if ((x.w[0] == y.w[0]) && (x.w[1] == y.w[1])) {
        res = 0;
        return (res, pfpsf);
    }
    if ((x.w[1] & 0x7800000000000000) == 0x7800000000000000) {
        if ((x.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = 0;
            if ((((y.w[1] & 0x7800000000000000) != 0x7800000000000000)) || ((y.w[1] & 0x8000000000000000) != 0x8000000000000000)) {
                res = 1;
            }
            return (res, pfpsf);
        } else {
            res = 0;
            return (res, pfpsf);
        }
    } else if ((y.w[1] & 0x7800000000000000) == 0x7800000000000000) {
        res = 0;
        if ((y.w[1] & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    sig_x.w[1] = (x.w[1] & 0x0001ffffffffffff);
    sig_x.w[0] = x.w[0];
    exp_x = ((((go_checked_shr_u64(x.w[1], go_shift_count_u64((49) as u64)))) & 0x000000000003fff) as i64);
    if (((sig_x.w[1] > 0x0001ed09bead87c0) || (((sig_x.w[1] == 0x0001ed09bead87c0) && (sig_x.w[0] > 0x378d8e63ffffffff)))) || (((x.w[1] & 0x6000000000000000) == 0x6000000000000000))) {
        non_canon_x = 1;
    } else {
        non_canon_x = 0;
    }
    exp_y = ((((go_checked_shr_u64(y.w[1], go_shift_count_u64((49) as u64)))) & 0x0000000000003fff) as i64);
    sig_y.w[1] = (y.w[1] & 0x0001ffffffffffff);
    sig_y.w[0] = y.w[0];
    if (((sig_y.w[1] > 0x0001ed09bead87c0) || (((sig_y.w[1] == 0x0001ed09bead87c0) && (sig_y.w[0] > 0x378d8e63ffffffff)))) || (((y.w[1] & 0x6000000000000000) == 0x6000000000000000))) {
        non_canon_y = 1;
    } else {
        non_canon_y = 0;
    }
    if ((non_canon_x != 0) || (((sig_x.w[1] == 0) && (sig_x.w[0] == 0)))) {
        x_is_zero = 1;
    }
    if ((non_canon_y != 0) || (((sig_y.w[1] == 0) && (sig_y.w[0] == 0)))) {
        y_is_zero = 1;
    }
    if ((x_is_zero != 0) && (y_is_zero != 0)) {
        res = 0;
        return (res, pfpsf);
    } else if (x_is_zero != 0) {
        res = 0;
        if ((y.w[1] & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    } else if (y_is_zero != 0) {
        res = 0;
        if ((x.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((((x.w[1] ^ y.w[1]) & 0x8000000000000000)) == 0x8000000000000000) {
        res = 0;
        if ((y.w[1] & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (exp_y == exp_x) {
        res = 0;
        if ((((sig_x.w[1] > sig_y.w[1]) || (((sig_x.w[1] == sig_y.w[1]) && (sig_x.w[0] >= sig_y.w[0]))))) != (((x.w[1] & 0x8000000000000000) != 0x8000000000000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((((sig_x.w[1] > sig_y.w[1]) || (((sig_x.w[1] == sig_y.w[1]) && (sig_x.w[0] > sig_y.w[0]))))) && (exp_x >= exp_y)) {
        res = 0;
        if ((x.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((((sig_x.w[1] < sig_y.w[1]) || (((sig_x.w[1] == sig_y.w[1]) && (sig_x.w[0] < sig_y.w[0]))))) && (exp_x <= exp_y)) {
        res = 0;
        if ((x.w[1] & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    diff = (exp_x.wrapping_sub(exp_y));
    if (diff > 0) {
        if (diff > 33) {
            res = 0;
            if ((x.w[1] & 0x8000000000000000) == 0x8000000000000000) {
                res = 1;
            }
            return (res, pfpsf);
        }
        if (diff > 19) {
            sig_n_prime256 = __mul_128x128_to_256(sig_x, bid_ten2k128[(diff.wrapping_sub(20)) as usize]);
            if ((((sig_n_prime256.w[3] == 0) && (sig_n_prime256.w[2] == 0)) && (sig_n_prime256.w[1] == sig_y.w[1])) && (sig_n_prime256.w[0] == sig_y.w[0])) {
                res = 0;
                return (res, pfpsf);
            }
            res = 0;
            if (((((((sig_n_prime256.w[3] > 0) || (sig_n_prime256.w[2] > 0))) || (sig_n_prime256.w[1] > sig_y.w[1])) || (((sig_n_prime256.w[1] == sig_y.w[1]) && (sig_n_prime256.w[0] > sig_y.w[0]))))) != (((y.w[1] & 0x8000000000000000) != 0x8000000000000000))) {
                res = 1;
            }
            return (res, pfpsf);
        }
        sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[diff as usize], sig_x);
        if (((sig_n_prime192.w[2] == 0) && (sig_n_prime192.w[1] == sig_y.w[1])) && (sig_n_prime192.w[0] == sig_y.w[0])) {
            res = 0;
            return (res, pfpsf);
        }
        res = 0;
        if (((((sig_n_prime192.w[2] > 0) || (sig_n_prime192.w[1] > sig_y.w[1])) || (((sig_n_prime192.w[1] == sig_y.w[1]) && (sig_n_prime192.w[0] > sig_y.w[0]))))) != (((y.w[1] & 0x8000000000000000) != 0x8000000000000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    diff = (exp_y.wrapping_sub(exp_x));
    if (diff > 33) {
        res = 0;
        if ((x.w[1] & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (diff > 19) {
        sig_n_prime256 = __mul_128x128_to_256(sig_y, bid_ten2k128[(diff.wrapping_sub(20)) as usize]);
        if ((((sig_n_prime256.w[3] == 0) && (sig_n_prime256.w[2] == 0)) && (sig_n_prime256.w[1] == sig_x.w[1])) && (sig_n_prime256.w[0] == sig_x.w[0])) {
            res = 0;
            return (res, pfpsf);
        }
        res = 0;
        if (((((sig_n_prime256.w[3] != 0) || (sig_n_prime256.w[2] != 0)) || (((sig_n_prime256.w[1] > sig_x.w[1]) || (((sig_n_prime256.w[1] == sig_x.w[1]) && (sig_n_prime256.w[0] > sig_x.w[0]))))))) != (((x.w[1] & 0x8000000000000000) == 0x8000000000000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[diff as usize], sig_y);
    if (((sig_n_prime192.w[2] == 0) && (sig_n_prime192.w[1] == sig_x.w[1])) && (sig_n_prime192.w[0] == sig_x.w[0])) {
        res = 0;
        return (res, pfpsf);
    }
    res = 0;
    if ((((sig_n_prime192.w[2] != 0) || (((sig_n_prime192.w[1] > sig_x.w[1]) || (((sig_n_prime192.w[1] == sig_x.w[1]) && (sig_n_prime192.w[0] > sig_x.w[0]))))))) != (((y.w[1] & 0x8000000000000000) == 0x8000000000000000))) {
        res = 1;
    }
    return (res, pfpsf);
}

pub fn bid128_quiet_less_equal(mut x: BID_UINT128, mut y: BID_UINT128) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut diff: i64 = 0;
    let mut sig_x: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut sig_y: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut sig_n_prime192: BID_UINT192 = BID_UINT192 { w: [0, 0, 0] };
    let mut sig_n_prime256: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut pfpsf: u32 = 0;
    let mut x_is_zero: i64 = 0;
    let mut y_is_zero: i64 = 0;
    let mut non_canon_x: i64 = 0;
    let mut non_canon_y: i64 = 0;
    if ((((x.w[1] & 0x7c00000000000000) == 0x7c00000000000000)) || (((y.w[1] & 0x7c00000000000000) == 0x7c00000000000000))) {
        if (((x.w[1] & 0x7e00000000000000) == 0x7e00000000000000) || ((y.w[1] & 0x7e00000000000000) == 0x7e00000000000000)) {
            pfpsf |= 1;
        }
        res = 0;
        return (res, pfpsf);
    }
    if ((x.w[0] == y.w[0]) && (x.w[1] == y.w[1])) {
        res = 1;
        return (res, pfpsf);
    }
    if ((x.w[1] & 0x7800000000000000) == 0x7800000000000000) {
        if ((x.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
            return (res, pfpsf);
        } else {
            res = 0;
            if ((((y.w[1] & 0x7800000000000000) == 0x7800000000000000)) && (((y.w[1] & 0x8000000000000000) != 0x8000000000000000))) {
                res = 1;
            }
            return (res, pfpsf);
        }
    } else if ((y.w[1] & 0x7800000000000000) == 0x7800000000000000) {
        res = 0;
        if ((y.w[1] & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    sig_x.w[1] = (x.w[1] & 0x0001ffffffffffff);
    sig_x.w[0] = x.w[0];
    exp_x = ((((go_checked_shr_u64(x.w[1], go_shift_count_u64((49) as u64)))) & 0x000000000003fff) as i64);
    if (((sig_x.w[1] > 0x0001ed09bead87c0) || (((sig_x.w[1] == 0x0001ed09bead87c0) && (sig_x.w[0] > 0x378d8e63ffffffff)))) || (((x.w[1] & 0x6000000000000000) == 0x6000000000000000))) {
        non_canon_x = 1;
    } else {
        non_canon_x = 0;
    }
    exp_y = ((((go_checked_shr_u64(y.w[1], go_shift_count_u64((49) as u64)))) & 0x0000000000003fff) as i64);
    sig_y.w[1] = (y.w[1] & 0x0001ffffffffffff);
    sig_y.w[0] = y.w[0];
    if (((sig_y.w[1] > 0x0001ed09bead87c0) || (((sig_y.w[1] == 0x0001ed09bead87c0) && (sig_y.w[0] > 0x378d8e63ffffffff)))) || (((y.w[1] & 0x6000000000000000) == 0x6000000000000000))) {
        non_canon_y = 1;
    } else {
        non_canon_y = 0;
    }
    if ((non_canon_x != 0) || (((sig_x.w[1] == 0) && (sig_x.w[0] == 0)))) {
        x_is_zero = 1;
    }
    if ((non_canon_y != 0) || (((sig_y.w[1] == 0) && (sig_y.w[0] == 0)))) {
        y_is_zero = 1;
    }
    if ((x_is_zero != 0) && (y_is_zero != 0)) {
        res = 1;
        return (res, pfpsf);
    } else if (x_is_zero != 0) {
        res = 0;
        if ((y.w[1] & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    } else if (y_is_zero != 0) {
        res = 0;
        if ((x.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((((x.w[1] ^ y.w[1]) & 0x8000000000000000)) == 0x8000000000000000) {
        res = 0;
        if ((y.w[1] & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (exp_y == exp_x) {
        res = 0;
        if ((((sig_x.w[1] > sig_y.w[1]) || (((sig_x.w[1] == sig_y.w[1]) && (sig_x.w[0] >= sig_y.w[0]))))) != (((x.w[1] & 0x8000000000000000) != 0x8000000000000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((((sig_x.w[1] > sig_y.w[1]) || (((sig_x.w[1] == sig_y.w[1]) && (sig_x.w[0] > sig_y.w[0]))))) && (exp_x >= exp_y)) {
        res = 0;
        if ((x.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((((sig_x.w[1] < sig_y.w[1]) || (((sig_x.w[1] == sig_y.w[1]) && (sig_x.w[0] < sig_y.w[0]))))) && (exp_x <= exp_y)) {
        res = 0;
        if ((x.w[1] & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    diff = (exp_x.wrapping_sub(exp_y));
    if (diff > 0) {
        if (diff > 33) {
            res = 0;
            if ((x.w[1] & 0x8000000000000000) == 0x8000000000000000) {
                res = 1;
            }
            return (res, pfpsf);
        }
        if (diff > 19) {
            sig_n_prime256 = __mul_128x128_to_256(sig_x, bid_ten2k128[(diff.wrapping_sub(20)) as usize]);
            if ((((sig_n_prime256.w[3] == 0) && (sig_n_prime256.w[2] == 0)) && (sig_n_prime256.w[1] == sig_y.w[1])) && (sig_n_prime256.w[0] == sig_y.w[0])) {
                res = 1;
                return (res, pfpsf);
            }
            res = 0;
            if (((((((sig_n_prime256.w[3] > 0) || (sig_n_prime256.w[2] > 0))) || (sig_n_prime256.w[1] > sig_y.w[1])) || (((sig_n_prime256.w[1] == sig_y.w[1]) && (sig_n_prime256.w[0] > sig_y.w[0]))))) != (((y.w[1] & 0x8000000000000000) != 0x8000000000000000))) {
                res = 1;
            }
            return (res, pfpsf);
        }
        sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[diff as usize], sig_x);
        if (((sig_n_prime192.w[2] == 0) && (sig_n_prime192.w[1] == sig_y.w[1])) && (sig_n_prime192.w[0] == sig_y.w[0])) {
            res = 1;
            return (res, pfpsf);
        }
        res = 0;
        if (((((sig_n_prime192.w[2] > 0) || (sig_n_prime192.w[1] > sig_y.w[1])) || (((sig_n_prime192.w[1] == sig_y.w[1]) && (sig_n_prime192.w[0] > sig_y.w[0]))))) != (((y.w[1] & 0x8000000000000000) != 0x8000000000000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    diff = (exp_y.wrapping_sub(exp_x));
    if (diff > 33) {
        res = 0;
        if ((x.w[1] & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (diff > 19) {
        sig_n_prime256 = __mul_128x128_to_256(sig_y, bid_ten2k128[(diff.wrapping_sub(20)) as usize]);
        if ((((sig_n_prime256.w[3] == 0) && (sig_n_prime256.w[2] == 0)) && (sig_n_prime256.w[1] == sig_x.w[1])) && (sig_n_prime256.w[0] == sig_x.w[0])) {
            res = 1;
            return (res, pfpsf);
        }
        res = 0;
        if (((((sig_n_prime256.w[3] != 0) || (sig_n_prime256.w[2] != 0)) || (((sig_n_prime256.w[1] > sig_x.w[1]) || (((sig_n_prime256.w[1] == sig_x.w[1]) && (sig_n_prime256.w[0] > sig_x.w[0]))))))) != (((x.w[1] & 0x8000000000000000) == 0x8000000000000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[diff as usize], sig_y);
    if (((sig_n_prime192.w[2] == 0) && (sig_n_prime192.w[1] == sig_x.w[1])) && (sig_n_prime192.w[0] == sig_x.w[0])) {
        res = 1;
        return (res, pfpsf);
    }
    res = 0;
    if ((((sig_n_prime192.w[2] != 0) || (((sig_n_prime192.w[1] > sig_x.w[1]) || (((sig_n_prime192.w[1] == sig_x.w[1]) && (sig_n_prime192.w[0] > sig_x.w[0]))))))) != (((y.w[1] & 0x8000000000000000) == 0x8000000000000000))) {
        res = 1;
    }
    return (res, pfpsf);
}

pub fn bid128_quiet_less_unordered(mut x: BID_UINT128, mut y: BID_UINT128) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut diff: i64 = 0;
    let mut sig_x: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut sig_y: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut sig_n_prime192: BID_UINT192 = BID_UINT192 { w: [0, 0, 0] };
    let mut sig_n_prime256: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut pfpsf: u32 = 0;
    let mut x_is_zero: i64 = 0;
    let mut y_is_zero: i64 = 0;
    let mut non_canon_x: i64 = 0;
    let mut non_canon_y: i64 = 0;
    if ((((x.w[1] & 0x7c00000000000000) == 0x7c00000000000000)) || (((y.w[1] & 0x7c00000000000000) == 0x7c00000000000000))) {
        if (((x.w[1] & 0x7e00000000000000) == 0x7e00000000000000) || ((y.w[1] & 0x7e00000000000000) == 0x7e00000000000000)) {
            pfpsf |= 1;
        }
        res = 1;
        return (res, pfpsf);
    }
    if ((x.w[0] == y.w[0]) && (x.w[1] == y.w[1])) {
        res = 0;
        return (res, pfpsf);
    }
    if ((x.w[1] & 0x7800000000000000) == 0x7800000000000000) {
        if ((x.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = 0;
            if ((((y.w[1] & 0x7800000000000000) != 0x7800000000000000)) || ((y.w[1] & 0x8000000000000000) != 0x8000000000000000)) {
                res = 1;
            }
            return (res, pfpsf);
        } else {
            res = 0;
            return (res, pfpsf);
        }
    } else if ((y.w[1] & 0x7800000000000000) == 0x7800000000000000) {
        res = 0;
        if ((y.w[1] & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    sig_x.w[1] = (x.w[1] & 0x0001ffffffffffff);
    sig_x.w[0] = x.w[0];
    exp_x = ((((go_checked_shr_u64(x.w[1], go_shift_count_u64((49) as u64)))) & 0x000000000003fff) as i64);
    if (((sig_x.w[1] > 0x0001ed09bead87c0) || (((sig_x.w[1] == 0x0001ed09bead87c0) && (sig_x.w[0] > 0x378d8e63ffffffff)))) || (((x.w[1] & 0x6000000000000000) == 0x6000000000000000))) {
        non_canon_x = 1;
    } else {
        non_canon_x = 0;
    }
    exp_y = ((((go_checked_shr_u64(y.w[1], go_shift_count_u64((49) as u64)))) & 0x0000000000003fff) as i64);
    sig_y.w[1] = (y.w[1] & 0x0001ffffffffffff);
    sig_y.w[0] = y.w[0];
    if (((sig_y.w[1] > 0x0001ed09bead87c0) || (((sig_y.w[1] == 0x0001ed09bead87c0) && (sig_y.w[0] > 0x378d8e63ffffffff)))) || (((y.w[1] & 0x6000000000000000) == 0x6000000000000000))) {
        non_canon_y = 1;
    } else {
        non_canon_y = 0;
    }
    if ((non_canon_x != 0) || (((sig_x.w[1] == 0) && (sig_x.w[0] == 0)))) {
        x_is_zero = 1;
    }
    if ((non_canon_y != 0) || (((sig_y.w[1] == 0) && (sig_y.w[0] == 0)))) {
        y_is_zero = 1;
    }
    if ((x_is_zero != 0) && (y_is_zero != 0)) {
        res = 0;
        return (res, pfpsf);
    } else if (x_is_zero != 0) {
        res = 0;
        if ((y.w[1] & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    } else if (y_is_zero != 0) {
        res = 0;
        if ((x.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((((x.w[1] ^ y.w[1]) & 0x8000000000000000)) == 0x8000000000000000) {
        res = 0;
        if ((y.w[1] & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (exp_y == exp_x) {
        res = 0;
        if ((((sig_x.w[1] > sig_y.w[1]) || (((sig_x.w[1] == sig_y.w[1]) && (sig_x.w[0] >= sig_y.w[0]))))) != (((x.w[1] & 0x8000000000000000) != 0x8000000000000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((((sig_x.w[1] > sig_y.w[1]) || (((sig_x.w[1] == sig_y.w[1]) && (sig_x.w[0] > sig_y.w[0]))))) && (exp_x >= exp_y)) {
        res = 0;
        if ((x.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((((sig_x.w[1] < sig_y.w[1]) || (((sig_x.w[1] == sig_y.w[1]) && (sig_x.w[0] < sig_y.w[0]))))) && (exp_x <= exp_y)) {
        res = 0;
        if ((x.w[1] & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    diff = (exp_x.wrapping_sub(exp_y));
    if (diff > 0) {
        if (diff > 33) {
            res = 0;
            if ((x.w[1] & 0x8000000000000000) == 0x8000000000000000) {
                res = 1;
            }
            return (res, pfpsf);
        }
        if (diff > 19) {
            sig_n_prime256 = __mul_128x128_to_256(sig_x, bid_ten2k128[(diff.wrapping_sub(20)) as usize]);
            if ((((sig_n_prime256.w[3] == 0) && (sig_n_prime256.w[2] == 0)) && (sig_n_prime256.w[1] == sig_y.w[1])) && (sig_n_prime256.w[0] == sig_y.w[0])) {
                res = 0;
                return (res, pfpsf);
            }
            res = 0;
            if (((((((sig_n_prime256.w[3] > 0) || (sig_n_prime256.w[2] > 0))) || (sig_n_prime256.w[1] > sig_y.w[1])) || (((sig_n_prime256.w[1] == sig_y.w[1]) && (sig_n_prime256.w[0] > sig_y.w[0]))))) != (((y.w[1] & 0x8000000000000000) != 0x8000000000000000))) {
                res = 1;
            }
            return (res, pfpsf);
        }
        sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[diff as usize], sig_x);
        if (((sig_n_prime192.w[2] == 0) && (sig_n_prime192.w[1] == sig_y.w[1])) && (sig_n_prime192.w[0] == sig_y.w[0])) {
            res = 0;
            return (res, pfpsf);
        }
        res = 0;
        if (((((sig_n_prime192.w[2] > 0) || (sig_n_prime192.w[1] > sig_y.w[1])) || (((sig_n_prime192.w[1] == sig_y.w[1]) && (sig_n_prime192.w[0] > sig_y.w[0]))))) != (((y.w[1] & 0x8000000000000000) != 0x8000000000000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    diff = (exp_y.wrapping_sub(exp_x));
    if (diff > 33) {
        res = 0;
        if ((x.w[1] & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (diff > 19) {
        sig_n_prime256 = __mul_128x128_to_256(sig_y, bid_ten2k128[(diff.wrapping_sub(20)) as usize]);
        if ((((sig_n_prime256.w[3] == 0) && (sig_n_prime256.w[2] == 0)) && (sig_n_prime256.w[1] == sig_x.w[1])) && (sig_n_prime256.w[0] == sig_x.w[0])) {
            res = 0;
            return (res, pfpsf);
        }
        res = 0;
        if (((((sig_n_prime256.w[3] != 0) || (sig_n_prime256.w[2] != 0)) || (((sig_n_prime256.w[1] > sig_x.w[1]) || (((sig_n_prime256.w[1] == sig_x.w[1]) && (sig_n_prime256.w[0] > sig_x.w[0]))))))) != (((x.w[1] & 0x8000000000000000) == 0x8000000000000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[diff as usize], sig_y);
    if (((sig_n_prime192.w[2] == 0) && (sig_n_prime192.w[1] == sig_x.w[1])) && (sig_n_prime192.w[0] == sig_x.w[0])) {
        res = 0;
        return (res, pfpsf);
    }
    res = 0;
    if ((((sig_n_prime192.w[2] != 0) || (((sig_n_prime192.w[1] > sig_x.w[1]) || (((sig_n_prime192.w[1] == sig_x.w[1]) && (sig_n_prime192.w[0] > sig_x.w[0]))))))) != (((y.w[1] & 0x8000000000000000) == 0x8000000000000000))) {
        res = 1;
    }
    return (res, pfpsf);
}

pub fn bid128_quiet_not_equal(mut x: BID_UINT128, mut y: BID_UINT128) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut exp_t: i64 = 0;
    let mut sig_x: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut sig_y: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut sig_t: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut sig_n_prime192: BID_UINT192 = BID_UINT192 { w: [0, 0, 0] };
    let mut sig_n_prime256: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut pfpsf: u32 = 0;
    let mut x_is_zero: i64 = 0;
    let mut y_is_zero: i64 = 0;
    let mut non_canon_x: i64 = 0;
    let mut non_canon_y: i64 = 0;
    if ((((x.w[1] & 0x7c00000000000000) == 0x7c00000000000000)) || (((y.w[1] & 0x7c00000000000000) == 0x7c00000000000000))) {
        if (((x.w[1] & 0x7e00000000000000) == 0x7e00000000000000) || ((y.w[1] & 0x7e00000000000000) == 0x7e00000000000000)) {
            pfpsf |= 1;
        }
        res = 1;
        return (res, pfpsf);
    }
    if ((x.w[0] == y.w[0]) && (x.w[1] == y.w[1])) {
        res = 0;
        return (res, pfpsf);
    }
    if ((x.w[1] & 0x7800000000000000) == 0x7800000000000000) {
        if ((y.w[1] & 0x7800000000000000) == 0x7800000000000000) {
            res = 0;
            if ((((x.w[1] ^ y.w[1]) & 0x8000000000000000)) == 0x8000000000000000) {
                res = 1;
            }
            return (res, pfpsf);
        } else {
            res = 1;
            return (res, pfpsf);
        }
    }
    if ((y.w[1] & 0x7800000000000000) == 0x7800000000000000) {
        res = 1;
        return (res, pfpsf);
    }
    sig_x.w[1] = (x.w[1] & 0x0001ffffffffffff);
    sig_x.w[0] = x.w[0];
    exp_x = ((((go_checked_shr_u64(x.w[1], go_shift_count_u64((49) as u64)))) & 0x000000000003fff) as i64);
    if (((sig_x.w[1] > 0x0001ed09bead87c0) || (((sig_x.w[1] == 0x0001ed09bead87c0) && (sig_x.w[0] > 0x378d8e63ffffffff)))) || (((x.w[1] & 0x6000000000000000) == 0x6000000000000000))) {
        non_canon_x = 1;
    } else {
        non_canon_x = 0;
    }
    exp_y = ((((go_checked_shr_u64(y.w[1], go_shift_count_u64((49) as u64)))) & 0x0000000000003fff) as i64);
    sig_y.w[1] = (y.w[1] & 0x0001ffffffffffff);
    sig_y.w[0] = y.w[0];
    if (((sig_y.w[1] > 0x0001ed09bead87c0) || (((sig_y.w[1] == 0x0001ed09bead87c0) && (sig_y.w[0] > 0x378d8e63ffffffff)))) || (((y.w[1] & 0x6000000000000000) == 0x6000000000000000))) {
        non_canon_y = 1;
    } else {
        non_canon_y = 0;
    }
    if ((non_canon_x != 0) || (((sig_x.w[1] == 0) && (sig_x.w[0] == 0)))) {
        x_is_zero = 1;
    }
    if ((non_canon_y != 0) || (((sig_y.w[1] == 0) && (sig_y.w[0] == 0)))) {
        y_is_zero = 1;
    }
    if ((x_is_zero != 0) && (y_is_zero != 0)) {
        res = 0;
        return (res, pfpsf);
    } else if ((((x_is_zero != 0) && (y_is_zero == 0))) || (((x_is_zero == 0) && (y_is_zero != 0)))) {
        res = 1;
        return (res, pfpsf);
    }
    if (((x.w[1] ^ y.w[1]) & 0x8000000000000000) != 0) {
        res = 1;
        return (res, pfpsf);
    }
    if (exp_x > exp_y) {
        exp_t = exp_x;
        exp_x = exp_y;
        exp_y = exp_t;
        sig_t.w[1] = sig_x.w[1];
        sig_x.w[1] = sig_y.w[1];
        sig_y.w[1] = sig_t.w[1];
        sig_t.w[0] = sig_x.w[0];
        sig_x.w[0] = sig_y.w[0];
        sig_y.w[0] = sig_t.w[0];
    }
    if ((exp_y.wrapping_sub(exp_x)) > 33) {
        res = 1;
        return (res, pfpsf);
    }
    if ((exp_y.wrapping_sub(exp_x)) > 19) {
        sig_n_prime256 = __mul_128x128_to_256(sig_y, bid_ten2k128[((exp_y.wrapping_sub(exp_x)).wrapping_sub(20)) as usize]);
        res = 0;
        if ((((sig_n_prime256.w[3] != 0) || (sig_n_prime256.w[2] != 0)) || (sig_n_prime256.w[1] != sig_x.w[1])) || (sig_n_prime256.w[0] != sig_x.w[0])) {
            res = 1;
        }
        return (res, pfpsf);
    }
    sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[(exp_y.wrapping_sub(exp_x)) as usize], sig_y);
    res = 0;
    if (((sig_n_prime192.w[2] != 0) || (sig_n_prime192.w[1] != sig_x.w[1])) || (sig_n_prime192.w[0] != sig_x.w[0])) {
        res = 1;
    }
    return (res, pfpsf);
}

pub fn bid128_quiet_not_greater(mut x: BID_UINT128, mut y: BID_UINT128) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut diff: i64 = 0;
    let mut sig_x: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut sig_y: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut sig_n_prime192: BID_UINT192 = BID_UINT192 { w: [0, 0, 0] };
    let mut sig_n_prime256: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut pfpsf: u32 = 0;
    let mut x_is_zero: i64 = 0;
    let mut y_is_zero: i64 = 0;
    let mut non_canon_x: i64 = 0;
    let mut non_canon_y: i64 = 0;
    if ((((x.w[1] & 0x7c00000000000000) == 0x7c00000000000000)) || (((y.w[1] & 0x7c00000000000000) == 0x7c00000000000000))) {
        if (((x.w[1] & 0x7e00000000000000) == 0x7e00000000000000) || ((y.w[1] & 0x7e00000000000000) == 0x7e00000000000000)) {
            pfpsf |= 1;
        }
        res = 1;
        return (res, pfpsf);
    }
    if ((x.w[0] == y.w[0]) && (x.w[1] == y.w[1])) {
        res = 1;
        return (res, pfpsf);
    }
    if ((x.w[1] & 0x7800000000000000) == 0x7800000000000000) {
        if ((x.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
            return (res, pfpsf);
        } else {
            res = 0;
            if ((((y.w[1] & 0x7800000000000000) == 0x7800000000000000)) && (((y.w[1] & 0x8000000000000000) != 0x8000000000000000))) {
                res = 1;
            }
            return (res, pfpsf);
        }
    } else if ((y.w[1] & 0x7800000000000000) == 0x7800000000000000) {
        res = 0;
        if ((y.w[1] & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    sig_x.w[1] = (x.w[1] & 0x0001ffffffffffff);
    sig_x.w[0] = x.w[0];
    exp_x = ((((go_checked_shr_u64(x.w[1], go_shift_count_u64((49) as u64)))) & 0x000000000003fff) as i64);
    if (((sig_x.w[1] > 0x0001ed09bead87c0) || (((sig_x.w[1] == 0x0001ed09bead87c0) && (sig_x.w[0] > 0x378d8e63ffffffff)))) || (((x.w[1] & 0x6000000000000000) == 0x6000000000000000))) {
        non_canon_x = 1;
    } else {
        non_canon_x = 0;
    }
    exp_y = ((((go_checked_shr_u64(y.w[1], go_shift_count_u64((49) as u64)))) & 0x0000000000003fff) as i64);
    sig_y.w[1] = (y.w[1] & 0x0001ffffffffffff);
    sig_y.w[0] = y.w[0];
    if (((sig_y.w[1] > 0x0001ed09bead87c0) || (((sig_y.w[1] == 0x0001ed09bead87c0) && (sig_y.w[0] > 0x378d8e63ffffffff)))) || (((y.w[1] & 0x6000000000000000) == 0x6000000000000000))) {
        non_canon_y = 1;
    } else {
        non_canon_y = 0;
    }
    if ((non_canon_x != 0) || (((sig_x.w[1] == 0) && (sig_x.w[0] == 0)))) {
        x_is_zero = 1;
    }
    if ((non_canon_y != 0) || (((sig_y.w[1] == 0) && (sig_y.w[0] == 0)))) {
        y_is_zero = 1;
    }
    if ((x_is_zero != 0) && (y_is_zero != 0)) {
        res = 1;
        return (res, pfpsf);
    } else if (x_is_zero != 0) {
        res = 0;
        if ((y.w[1] & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    } else if (y_is_zero != 0) {
        res = 0;
        if ((x.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((((x.w[1] ^ y.w[1]) & 0x8000000000000000)) == 0x8000000000000000) {
        res = 0;
        if ((y.w[1] & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (exp_y == exp_x) {
        res = 0;
        if ((((sig_x.w[1] > sig_y.w[1]) || (((sig_x.w[1] == sig_y.w[1]) && (sig_x.w[0] >= sig_y.w[0]))))) != (((x.w[1] & 0x8000000000000000) != 0x8000000000000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((((sig_x.w[1] > sig_y.w[1]) || (((sig_x.w[1] == sig_y.w[1]) && (sig_x.w[0] > sig_y.w[0]))))) && (exp_x >= exp_y)) {
        res = 0;
        if ((x.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((((sig_x.w[1] < sig_y.w[1]) || (((sig_x.w[1] == sig_y.w[1]) && (sig_x.w[0] < sig_y.w[0]))))) && (exp_x <= exp_y)) {
        res = 0;
        if ((x.w[1] & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    diff = (exp_x.wrapping_sub(exp_y));
    if (diff > 0) {
        if (diff > 33) {
            res = 0;
            if ((x.w[1] & 0x8000000000000000) == 0x8000000000000000) {
                res = 1;
            }
            return (res, pfpsf);
        }
        if (diff > 19) {
            sig_n_prime256 = __mul_128x128_to_256(sig_x, bid_ten2k128[(diff.wrapping_sub(20)) as usize]);
            if ((((sig_n_prime256.w[3] == 0) && (sig_n_prime256.w[2] == 0)) && (sig_n_prime256.w[1] == sig_y.w[1])) && (sig_n_prime256.w[0] == sig_y.w[0])) {
                res = 1;
                return (res, pfpsf);
            }
            res = 0;
            if (((((((sig_n_prime256.w[3] > 0) || (sig_n_prime256.w[2] > 0))) || (sig_n_prime256.w[1] > sig_y.w[1])) || (((sig_n_prime256.w[1] == sig_y.w[1]) && (sig_n_prime256.w[0] > sig_y.w[0]))))) != (((y.w[1] & 0x8000000000000000) != 0x8000000000000000))) {
                res = 1;
            }
            return (res, pfpsf);
        }
        sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[diff as usize], sig_x);
        if (((sig_n_prime192.w[2] == 0) && (sig_n_prime192.w[1] == sig_y.w[1])) && (sig_n_prime192.w[0] == sig_y.w[0])) {
            res = 1;
            return (res, pfpsf);
        }
        res = 0;
        if (((((sig_n_prime192.w[2] > 0) || (sig_n_prime192.w[1] > sig_y.w[1])) || (((sig_n_prime192.w[1] == sig_y.w[1]) && (sig_n_prime192.w[0] > sig_y.w[0]))))) != (((y.w[1] & 0x8000000000000000) != 0x8000000000000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    diff = (exp_y.wrapping_sub(exp_x));
    if (diff > 33) {
        res = 0;
        if ((x.w[1] & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (diff > 19) {
        sig_n_prime256 = __mul_128x128_to_256(sig_y, bid_ten2k128[(diff.wrapping_sub(20)) as usize]);
        if ((((sig_n_prime256.w[3] == 0) && (sig_n_prime256.w[2] == 0)) && (sig_n_prime256.w[1] == sig_x.w[1])) && (sig_n_prime256.w[0] == sig_x.w[0])) {
            res = 1;
            return (res, pfpsf);
        }
        res = 0;
        if (((((sig_n_prime256.w[3] != 0) || (sig_n_prime256.w[2] != 0)) || (((sig_n_prime256.w[1] > sig_x.w[1]) || (((sig_n_prime256.w[1] == sig_x.w[1]) && (sig_n_prime256.w[0] > sig_x.w[0]))))))) != (((x.w[1] & 0x8000000000000000) == 0x8000000000000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[diff as usize], sig_y);
    if (((sig_n_prime192.w[2] == 0) && (sig_n_prime192.w[1] == sig_x.w[1])) && (sig_n_prime192.w[0] == sig_x.w[0])) {
        res = 1;
        return (res, pfpsf);
    }
    res = 0;
    if ((((sig_n_prime192.w[2] != 0) || (((sig_n_prime192.w[1] > sig_x.w[1]) || (((sig_n_prime192.w[1] == sig_x.w[1]) && (sig_n_prime192.w[0] > sig_x.w[0]))))))) != (((y.w[1] & 0x8000000000000000) == 0x8000000000000000))) {
        res = 1;
    }
    return (res, pfpsf);
}

pub fn bid128_quiet_not_less(mut x: BID_UINT128, mut y: BID_UINT128) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut diff: i64 = 0;
    let mut sig_x: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut sig_y: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut sig_n_prime192: BID_UINT192 = BID_UINT192 { w: [0, 0, 0] };
    let mut sig_n_prime256: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut pfpsf: u32 = 0;
    let mut x_is_zero: i64 = 0;
    let mut y_is_zero: i64 = 0;
    let mut non_canon_x: i64 = 0;
    let mut non_canon_y: i64 = 0;
    if ((((x.w[1] & 0x7c00000000000000) == 0x7c00000000000000)) || (((y.w[1] & 0x7c00000000000000) == 0x7c00000000000000))) {
        if (((x.w[1] & 0x7e00000000000000) == 0x7e00000000000000) || ((y.w[1] & 0x7e00000000000000) == 0x7e00000000000000)) {
            pfpsf |= 1;
        }
        res = 1;
        return (res, pfpsf);
    }
    if ((x.w[0] == y.w[0]) && (x.w[1] == y.w[1])) {
        res = 1;
        return (res, pfpsf);
    }
    if ((x.w[1] & 0x7800000000000000) == 0x7800000000000000) {
        if ((x.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = 0;
            if ((((y.w[1] & 0x7800000000000000) == 0x7800000000000000)) && ((y.w[1] & 0x8000000000000000) == 0x8000000000000000)) {
                res = 1;
            }
            return (res, pfpsf);
        } else {
            res = 1;
            return (res, pfpsf);
        }
    } else if ((y.w[1] & 0x7800000000000000) == 0x7800000000000000) {
        res = 0;
        if ((y.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    sig_x.w[1] = (x.w[1] & 0x0001ffffffffffff);
    sig_x.w[0] = x.w[0];
    exp_x = ((((go_checked_shr_u64(x.w[1], go_shift_count_u64((49) as u64)))) & 0x000000000003fff) as i64);
    if (((sig_x.w[1] > 0x0001ed09bead87c0) || (((sig_x.w[1] == 0x0001ed09bead87c0) && (sig_x.w[0] > 0x378d8e63ffffffff)))) || (((x.w[1] & 0x6000000000000000) == 0x6000000000000000))) {
        non_canon_x = 1;
    } else {
        non_canon_x = 0;
    }
    exp_y = ((((go_checked_shr_u64(y.w[1], go_shift_count_u64((49) as u64)))) & 0x0000000000003fff) as i64);
    sig_y.w[1] = (y.w[1] & 0x0001ffffffffffff);
    sig_y.w[0] = y.w[0];
    if (((sig_y.w[1] > 0x0001ed09bead87c0) || (((sig_y.w[1] == 0x0001ed09bead87c0) && (sig_y.w[0] > 0x378d8e63ffffffff)))) || (((y.w[1] & 0x6000000000000000) == 0x6000000000000000))) {
        non_canon_y = 1;
    } else {
        non_canon_y = 0;
    }
    if ((non_canon_x != 0) || (((sig_x.w[1] == 0) && (sig_x.w[0] == 0)))) {
        x_is_zero = 1;
    }
    if ((non_canon_y != 0) || (((sig_y.w[1] == 0) && (sig_y.w[0] == 0)))) {
        y_is_zero = 1;
    }
    if ((x_is_zero != 0) && (y_is_zero != 0)) {
        res = 1;
        return (res, pfpsf);
    } else if (x_is_zero != 0) {
        res = 0;
        if ((y.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    } else if (y_is_zero != 0) {
        res = 0;
        if ((x.w[1] & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((((x.w[1] ^ y.w[1]) & 0x8000000000000000)) == 0x8000000000000000) {
        res = 0;
        if ((y.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (exp_y == exp_x) {
        res = 0;
        if ((((sig_x.w[1] > sig_y.w[1]) || (((sig_x.w[1] == sig_y.w[1]) && (sig_x.w[0] >= sig_y.w[0]))))) != (((x.w[1] & 0x8000000000000000) == 0x8000000000000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (((sig_x.w[1] >= sig_y.w[1]) && (sig_x.w[0] >= sig_y.w[0])) && (exp_x > exp_y)) {
        res = 0;
        if ((x.w[1] & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (((sig_x.w[1] <= sig_y.w[1]) && (sig_x.w[0] <= sig_y.w[0])) && (exp_x < exp_y)) {
        res = 0;
        if ((x.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    diff = (exp_x.wrapping_sub(exp_y));
    if (diff > 0) {
        if (diff > 33) {
            res = 0;
            if ((x.w[1] & 0x8000000000000000) != 0x8000000000000000) {
                res = 1;
            }
            return (res, pfpsf);
        }
        if (diff > 19) {
            sig_n_prime256 = __mul_128x128_to_256(sig_x, bid_ten2k128[(diff.wrapping_sub(20)) as usize]);
            if ((((sig_n_prime256.w[3] == 0) && (sig_n_prime256.w[2] == 0)) && (sig_n_prime256.w[1] == sig_y.w[1])) && (sig_n_prime256.w[0] == sig_y.w[0])) {
                res = 1;
                return (res, pfpsf);
            }
            res = 0;
            if (((((((sig_n_prime256.w[3] > 0) || (sig_n_prime256.w[2] > 0))) || (sig_n_prime256.w[1] > sig_y.w[1])) || (((sig_n_prime256.w[1] == sig_y.w[1]) && (sig_n_prime256.w[0] > sig_y.w[0]))))) != (((y.w[1] & 0x8000000000000000) == 0x8000000000000000))) {
                res = 1;
            }
            return (res, pfpsf);
        }
        sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[diff as usize], sig_x);
        if (((sig_n_prime192.w[2] == 0) && (sig_n_prime192.w[1] == sig_y.w[1])) && (sig_n_prime192.w[0] == sig_y.w[0])) {
            res = 1;
            return (res, pfpsf);
        }
        res = 0;
        if (((((sig_n_prime192.w[2] > 0) || (sig_n_prime192.w[1] > sig_y.w[1])) || (((sig_n_prime192.w[1] == sig_y.w[1]) && (sig_n_prime192.w[0] > sig_y.w[0]))))) != (((y.w[1] & 0x8000000000000000) == 0x8000000000000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    diff = (exp_y.wrapping_sub(exp_x));
    if (diff > 33) {
        res = 0;
        if ((x.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (diff > 19) {
        sig_n_prime256 = __mul_128x128_to_256(sig_y, bid_ten2k128[(diff.wrapping_sub(20)) as usize]);
        if ((((sig_n_prime256.w[3] == 0) && (sig_n_prime256.w[2] == 0)) && (sig_n_prime256.w[1] == sig_x.w[1])) && (sig_n_prime256.w[0] == sig_x.w[0])) {
            res = 1;
            return (res, pfpsf);
        }
        res = 0;
        if (((((sig_n_prime256.w[3] == 0) && (sig_n_prime256.w[2] == 0)) && (((sig_n_prime256.w[1] < sig_x.w[1]) || (((sig_n_prime256.w[1] == sig_x.w[1]) && (sig_n_prime256.w[0] < sig_x.w[0]))))))) != (((x.w[1] & 0x8000000000000000) == 0x8000000000000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[diff as usize], sig_y);
    if (((sig_n_prime192.w[2] == 0) && (sig_n_prime192.w[1] == sig_x.w[1])) && (sig_n_prime192.w[0] == sig_x.w[0])) {
        res = 1;
        return (res, pfpsf);
    }
    res = 0;
    if ((((sig_n_prime192.w[2] == 0) && (((sig_n_prime192.w[1] < sig_x.w[1]) || (((sig_n_prime192.w[1] == sig_x.w[1]) && (sig_n_prime192.w[0] < sig_x.w[0]))))))) != (((y.w[1] & 0x8000000000000000) == 0x8000000000000000))) {
        res = 1;
    }
    return (res, pfpsf);
}

pub fn bid128_quiet_ordered(mut x: BID_UINT128, mut y: BID_UINT128) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut pfpsf: u32 = 0;
    if ((((x.w[1] & 0x7c00000000000000) == 0x7c00000000000000)) || (((y.w[1] & 0x7c00000000000000) == 0x7c00000000000000))) {
        if (((x.w[1] & 0x7e00000000000000) == 0x7e00000000000000) || ((y.w[1] & 0x7e00000000000000) == 0x7e00000000000000)) {
            pfpsf |= 1;
        }
        res = 0;
        return (res, pfpsf);
    }
    res = 1;
    return (res, pfpsf);
}

pub fn bid128_quiet_unordered(mut x: BID_UINT128, mut y: BID_UINT128) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut pfpsf: u32 = 0;
    if ((((x.w[1] & 0x7c00000000000000) == 0x7c00000000000000)) || (((y.w[1] & 0x7c00000000000000) == 0x7c00000000000000))) {
        if (((x.w[1] & 0x7e00000000000000) == 0x7e00000000000000) || ((y.w[1] & 0x7e00000000000000) == 0x7e00000000000000)) {
            pfpsf |= 1;
        }
        res = 1;
        return (res, pfpsf);
    }
    res = 0;
    return (res, pfpsf);
}

pub fn bid128_signaling_greater(mut x: BID_UINT128, mut y: BID_UINT128) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut diff: i64 = 0;
    let mut sig_x: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut sig_y: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut sig_n_prime192: BID_UINT192 = BID_UINT192 { w: [0, 0, 0] };
    let mut sig_n_prime256: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut pfpsf: u32 = 0;
    let mut x_is_zero: i64 = 0;
    let mut y_is_zero: i64 = 0;
    let mut non_canon_x: i64 = 0;
    let mut non_canon_y: i64 = 0;
    if ((((x.w[1] & 0x7c00000000000000) == 0x7c00000000000000)) || (((y.w[1] & 0x7c00000000000000) == 0x7c00000000000000))) {
        pfpsf |= 1;
        res = 0;
        return (res, pfpsf);
    }
    if ((x.w[0] == y.w[0]) && (x.w[1] == y.w[1])) {
        res = 0;
        return (res, pfpsf);
    }
    if ((x.w[1] & 0x7800000000000000) == 0x7800000000000000) {
        if ((x.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = 0;
            return (res, pfpsf);
        } else {
            res = 0;
            if ((((y.w[1] & 0x7800000000000000) != 0x7800000000000000)) || (((y.w[1] & 0x8000000000000000) == 0x8000000000000000))) {
                res = 1;
            }
            return (res, pfpsf);
        }
    } else if ((y.w[1] & 0x7800000000000000) == 0x7800000000000000) {
        res = 0;
        if ((y.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    sig_x.w[1] = (x.w[1] & 0x0001ffffffffffff);
    sig_x.w[0] = x.w[0];
    exp_x = ((((go_checked_shr_u64(x.w[1], go_shift_count_u64((49) as u64)))) & 0x000000000003fff) as i64);
    if (((sig_x.w[1] > 0x0001ed09bead87c0) || (((sig_x.w[1] == 0x0001ed09bead87c0) && (sig_x.w[0] > 0x378d8e63ffffffff)))) || (((x.w[1] & 0x6000000000000000) == 0x6000000000000000))) {
        non_canon_x = 1;
    } else {
        non_canon_x = 0;
    }
    exp_y = ((((go_checked_shr_u64(y.w[1], go_shift_count_u64((49) as u64)))) & 0x0000000000003fff) as i64);
    sig_y.w[1] = (y.w[1] & 0x0001ffffffffffff);
    sig_y.w[0] = y.w[0];
    if (((sig_y.w[1] > 0x0001ed09bead87c0) || (((sig_y.w[1] == 0x0001ed09bead87c0) && (sig_y.w[0] > 0x378d8e63ffffffff)))) || (((y.w[1] & 0x6000000000000000) == 0x6000000000000000))) {
        non_canon_y = 1;
    } else {
        non_canon_y = 0;
    }
    if ((non_canon_x != 0) || (((sig_x.w[1] == 0) && (sig_x.w[0] == 0)))) {
        x_is_zero = 1;
    }
    if ((non_canon_y != 0) || (((sig_y.w[1] == 0) && (sig_y.w[0] == 0)))) {
        y_is_zero = 1;
    }
    if ((x_is_zero != 0) && (y_is_zero != 0)) {
        res = 0;
        return (res, pfpsf);
    } else if (x_is_zero != 0) {
        res = 0;
        if ((y.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    } else if (y_is_zero != 0) {
        res = 0;
        if ((x.w[1] & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((((x.w[1] ^ y.w[1]) & 0x8000000000000000)) == 0x8000000000000000) {
        res = 0;
        if ((y.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (exp_y == exp_x) {
        res = 0;
        if ((((sig_x.w[1] > sig_y.w[1]) || (((sig_x.w[1] == sig_y.w[1]) && (sig_x.w[0] >= sig_y.w[0]))))) != (((x.w[1] & 0x8000000000000000) == 0x8000000000000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((((sig_x.w[1] > sig_y.w[1]) || (((sig_x.w[1] == sig_y.w[1]) && (sig_x.w[0] > sig_y.w[0]))))) && (exp_x >= exp_y)) {
        res = 0;
        if ((x.w[1] & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((((sig_x.w[1] < sig_y.w[1]) || (((sig_x.w[1] == sig_y.w[1]) && (sig_x.w[0] < sig_y.w[0]))))) && (exp_x <= exp_y)) {
        res = 0;
        if ((x.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    diff = (exp_x.wrapping_sub(exp_y));
    if (diff > 0) {
        if (diff > 33) {
            res = 0;
            if ((x.w[1] & 0x8000000000000000) != 0x8000000000000000) {
                res = 1;
            }
            return (res, pfpsf);
        }
        if (diff > 19) {
            sig_n_prime256 = __mul_128x128_to_256(sig_x, bid_ten2k128[(diff.wrapping_sub(20)) as usize]);
            if ((((sig_n_prime256.w[3] == 0) && (sig_n_prime256.w[2] == 0)) && (sig_n_prime256.w[1] == sig_y.w[1])) && (sig_n_prime256.w[0] == sig_y.w[0])) {
                res = 0;
                return (res, pfpsf);
            }
            res = 0;
            if (((((((sig_n_prime256.w[3] > 0) || (sig_n_prime256.w[2] > 0))) || (sig_n_prime256.w[1] > sig_y.w[1])) || (((sig_n_prime256.w[1] == sig_y.w[1]) && (sig_n_prime256.w[0] > sig_y.w[0]))))) != (((y.w[1] & 0x8000000000000000) == 0x8000000000000000))) {
                res = 1;
            }
            return (res, pfpsf);
        }
        sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[diff as usize], sig_x);
        if (((sig_n_prime192.w[2] == 0) && (sig_n_prime192.w[1] == sig_y.w[1])) && (sig_n_prime192.w[0] == sig_y.w[0])) {
            res = 0;
            return (res, pfpsf);
        }
        res = 0;
        if (((((sig_n_prime192.w[2] > 0) || (sig_n_prime192.w[1] > sig_y.w[1])) || (((sig_n_prime192.w[1] == sig_y.w[1]) && (sig_n_prime192.w[0] > sig_y.w[0]))))) != (((y.w[1] & 0x8000000000000000) == 0x8000000000000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    diff = (exp_y.wrapping_sub(exp_x));
    if (diff > 33) {
        res = 0;
        if ((x.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (diff > 19) {
        sig_n_prime256 = __mul_128x128_to_256(sig_y, bid_ten2k128[(diff.wrapping_sub(20)) as usize]);
        if ((((sig_n_prime256.w[3] == 0) && (sig_n_prime256.w[2] == 0)) && (sig_n_prime256.w[1] == sig_x.w[1])) && (sig_n_prime256.w[0] == sig_x.w[0])) {
            res = 0;
            return (res, pfpsf);
        }
        res = 0;
        if (((((sig_n_prime256.w[3] != 0) || (sig_n_prime256.w[2] != 0)) || (((sig_n_prime256.w[1] > sig_x.w[1]) || (((sig_n_prime256.w[1] == sig_x.w[1]) && (sig_n_prime256.w[0] > sig_x.w[0]))))))) != (((x.w[1] & 0x8000000000000000) != 0x8000000000000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[diff as usize], sig_y);
    if (((sig_n_prime192.w[2] == 0) && (sig_n_prime192.w[1] == sig_x.w[1])) && (sig_n_prime192.w[0] == sig_x.w[0])) {
        res = 0;
        return (res, pfpsf);
    }
    res = 0;
    if ((((sig_n_prime192.w[2] != 0) || (((sig_n_prime192.w[1] > sig_x.w[1]) || (((sig_n_prime192.w[1] == sig_x.w[1]) && (sig_n_prime192.w[0] > sig_x.w[0]))))))) != (((y.w[1] & 0x8000000000000000) != 0x8000000000000000))) {
        res = 1;
    }
    return (res, pfpsf);
}

pub fn bid128_signaling_greater_equal(mut x: BID_UINT128, mut y: BID_UINT128) -> (i64, u32) {
    let (mut res, mut pfpsf) = bid128_quiet_greater_equal(x, y);
    if ((((x.w[1] & 0x7c00000000000000) == 0x7c00000000000000)) || (((y.w[1] & 0x7c00000000000000) == 0x7c00000000000000))) {
        pfpsf |= 1;
        return (0, pfpsf);
    }
    return (res, pfpsf);
}

pub fn bid128_signaling_greater_unordered(mut x: BID_UINT128, mut y: BID_UINT128) -> (i64, u32) {
    let (mut res, mut pfpsf) = bid128_quiet_greater_unordered(x, y);
    if ((((x.w[1] & 0x7c00000000000000) == 0x7c00000000000000)) || (((y.w[1] & 0x7c00000000000000) == 0x7c00000000000000))) {
        pfpsf |= 1;
        return (1, pfpsf);
    }
    return (res, pfpsf);
}

pub fn bid128_signaling_less(mut x: BID_UINT128, mut y: BID_UINT128) -> (i64, u32) {
    let (mut res, mut pfpsf) = bid128_quiet_less(x, y);
    if ((((x.w[1] & 0x7c00000000000000) == 0x7c00000000000000)) || (((y.w[1] & 0x7c00000000000000) == 0x7c00000000000000))) {
        pfpsf |= 1;
        return (0, pfpsf);
    }
    return (res, pfpsf);
}

pub fn bid128_signaling_less_equal(mut x: BID_UINT128, mut y: BID_UINT128) -> (i64, u32) {
    let (mut res, mut pfpsf) = bid128_quiet_less_equal(x, y);
    if ((((x.w[1] & 0x7c00000000000000) == 0x7c00000000000000)) || (((y.w[1] & 0x7c00000000000000) == 0x7c00000000000000))) {
        pfpsf |= 1;
        return (0, pfpsf);
    }
    return (res, pfpsf);
}

pub fn bid128_signaling_less_unordered(mut x: BID_UINT128, mut y: BID_UINT128) -> (i64, u32) {
    let (mut res, mut pfpsf) = bid128_quiet_less_unordered(x, y);
    if ((((x.w[1] & 0x7c00000000000000) == 0x7c00000000000000)) || (((y.w[1] & 0x7c00000000000000) == 0x7c00000000000000))) {
        pfpsf |= 1;
        return (1, pfpsf);
    }
    return (res, pfpsf);
}

pub fn bid128_signaling_not_greater(mut x: BID_UINT128, mut y: BID_UINT128) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut diff: i64 = 0;
    let mut sig_x: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut sig_y: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut sig_n_prime192: BID_UINT192 = BID_UINT192 { w: [0, 0, 0] };
    let mut sig_n_prime256: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut pfpsf: u32 = 0;
    let mut x_is_zero: i64 = 0;
    let mut y_is_zero: i64 = 0;
    let mut non_canon_x: i64 = 0;
    let mut non_canon_y: i64 = 0;
    if ((((x.w[1] & 0x7c00000000000000) == 0x7c00000000000000)) || (((y.w[1] & 0x7c00000000000000) == 0x7c00000000000000))) {
        pfpsf |= 1;
        res = 1;
        return (res, pfpsf);
    }
    if ((x.w[0] == y.w[0]) && (x.w[1] == y.w[1])) {
        res = 1;
        return (res, pfpsf);
    }
    if ((x.w[1] & 0x7800000000000000) == 0x7800000000000000) {
        if ((x.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
            return (res, pfpsf);
        } else {
            res = 0;
            if ((((y.w[1] & 0x7800000000000000) == 0x7800000000000000)) && (((y.w[1] & 0x8000000000000000) != 0x8000000000000000))) {
                res = 1;
            }
            return (res, pfpsf);
        }
    } else if ((y.w[1] & 0x7800000000000000) == 0x7800000000000000) {
        res = 0;
        if ((y.w[1] & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    sig_x.w[1] = (x.w[1] & 0x0001ffffffffffff);
    sig_x.w[0] = x.w[0];
    exp_x = ((((go_checked_shr_u64(x.w[1], go_shift_count_u64((49) as u64)))) & 0x000000000003fff) as i64);
    if (((sig_x.w[1] > 0x0001ed09bead87c0) || (((sig_x.w[1] == 0x0001ed09bead87c0) && (sig_x.w[0] > 0x378d8e63ffffffff)))) || (((x.w[1] & 0x6000000000000000) == 0x6000000000000000))) {
        non_canon_x = 1;
    } else {
        non_canon_x = 0;
    }
    exp_y = ((((go_checked_shr_u64(y.w[1], go_shift_count_u64((49) as u64)))) & 0x0000000000003fff) as i64);
    sig_y.w[1] = (y.w[1] & 0x0001ffffffffffff);
    sig_y.w[0] = y.w[0];
    if (((sig_y.w[1] > 0x0001ed09bead87c0) || (((sig_y.w[1] == 0x0001ed09bead87c0) && (sig_y.w[0] > 0x378d8e63ffffffff)))) || (((y.w[1] & 0x6000000000000000) == 0x6000000000000000))) {
        non_canon_y = 1;
    } else {
        non_canon_y = 0;
    }
    if ((non_canon_x != 0) || (((sig_x.w[1] == 0) && (sig_x.w[0] == 0)))) {
        x_is_zero = 1;
    }
    if ((non_canon_y != 0) || (((sig_y.w[1] == 0) && (sig_y.w[0] == 0)))) {
        y_is_zero = 1;
    }
    if ((x_is_zero != 0) && (y_is_zero != 0)) {
        res = 1;
        return (res, pfpsf);
    } else if (x_is_zero != 0) {
        res = 0;
        if ((y.w[1] & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    } else if (y_is_zero != 0) {
        res = 0;
        if ((x.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((((x.w[1] ^ y.w[1]) & 0x8000000000000000)) == 0x8000000000000000) {
        res = 0;
        if ((y.w[1] & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (exp_y == exp_x) {
        res = 0;
        if ((((sig_x.w[1] > sig_y.w[1]) || (((sig_x.w[1] == sig_y.w[1]) && (sig_x.w[0] >= sig_y.w[0]))))) != (((x.w[1] & 0x8000000000000000) != 0x8000000000000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((((sig_x.w[1] > sig_y.w[1]) || (((sig_x.w[1] == sig_y.w[1]) && (sig_x.w[0] > sig_y.w[0]))))) && (exp_x >= exp_y)) {
        res = 0;
        if ((x.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((((sig_x.w[1] < sig_y.w[1]) || (((sig_x.w[1] == sig_y.w[1]) && (sig_x.w[0] < sig_y.w[0]))))) && (exp_x <= exp_y)) {
        res = 0;
        if ((x.w[1] & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    diff = (exp_x.wrapping_sub(exp_y));
    if (diff > 0) {
        if (diff > 33) {
            res = 0;
            if ((x.w[1] & 0x8000000000000000) == 0x8000000000000000) {
                res = 1;
            }
            return (res, pfpsf);
        }
        if (diff > 19) {
            sig_n_prime256 = __mul_128x128_to_256(sig_x, bid_ten2k128[(diff.wrapping_sub(20)) as usize]);
            if ((((sig_n_prime256.w[3] == 0) && (sig_n_prime256.w[2] == 0)) && (sig_n_prime256.w[1] == sig_y.w[1])) && (sig_n_prime256.w[0] == sig_y.w[0])) {
                res = 1;
                return (res, pfpsf);
            }
            res = 0;
            if (((((((sig_n_prime256.w[3] > 0) || (sig_n_prime256.w[2] > 0))) || (sig_n_prime256.w[1] > sig_y.w[1])) || (((sig_n_prime256.w[1] == sig_y.w[1]) && (sig_n_prime256.w[0] > sig_y.w[0]))))) != (((y.w[1] & 0x8000000000000000) != 0x8000000000000000))) {
                res = 1;
            }
            return (res, pfpsf);
        }
        sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[diff as usize], sig_x);
        if (((sig_n_prime192.w[2] == 0) && (sig_n_prime192.w[1] == sig_y.w[1])) && (sig_n_prime192.w[0] == sig_y.w[0])) {
            res = 1;
            return (res, pfpsf);
        }
        res = 0;
        if (((((sig_n_prime192.w[2] > 0) || (sig_n_prime192.w[1] > sig_y.w[1])) || (((sig_n_prime192.w[1] == sig_y.w[1]) && (sig_n_prime192.w[0] > sig_y.w[0]))))) != (((y.w[1] & 0x8000000000000000) != 0x8000000000000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    diff = (exp_y.wrapping_sub(exp_x));
    if (diff > 33) {
        res = 0;
        if ((x.w[1] & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (diff > 19) {
        sig_n_prime256 = __mul_128x128_to_256(sig_y, bid_ten2k128[(diff.wrapping_sub(20)) as usize]);
        if ((((sig_n_prime256.w[3] == 0) && (sig_n_prime256.w[2] == 0)) && (sig_n_prime256.w[1] == sig_x.w[1])) && (sig_n_prime256.w[0] == sig_x.w[0])) {
            res = 1;
            return (res, pfpsf);
        }
        res = 0;
        if (((((sig_n_prime256.w[3] != 0) || (sig_n_prime256.w[2] != 0)) || (((sig_n_prime256.w[1] > sig_x.w[1]) || (((sig_n_prime256.w[1] == sig_x.w[1]) && (sig_n_prime256.w[0] > sig_x.w[0]))))))) != (((x.w[1] & 0x8000000000000000) == 0x8000000000000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[diff as usize], sig_y);
    if (((sig_n_prime192.w[2] == 0) && (sig_n_prime192.w[1] == sig_x.w[1])) && (sig_n_prime192.w[0] == sig_x.w[0])) {
        res = 1;
        return (res, pfpsf);
    }
    res = 0;
    if ((((sig_n_prime192.w[2] != 0) || (((sig_n_prime192.w[1] > sig_x.w[1]) || (((sig_n_prime192.w[1] == sig_x.w[1]) && (sig_n_prime192.w[0] > sig_x.w[0]))))))) != (((y.w[1] & 0x8000000000000000) == 0x8000000000000000))) {
        res = 1;
    }
    return (res, pfpsf);
}

pub fn bid128_signaling_not_less(mut x: BID_UINT128, mut y: BID_UINT128) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut diff: i64 = 0;
    let mut sig_x: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut sig_y: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut sig_n_prime192: BID_UINT192 = BID_UINT192 { w: [0, 0, 0] };
    let mut sig_n_prime256: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut pfpsf: u32 = 0;
    let mut x_is_zero: i64 = 0;
    let mut y_is_zero: i64 = 0;
    let mut non_canon_x: i64 = 0;
    let mut non_canon_y: i64 = 0;
    if ((((x.w[1] & 0x7c00000000000000) == 0x7c00000000000000)) || (((y.w[1] & 0x7c00000000000000) == 0x7c00000000000000))) {
        pfpsf |= 1;
        res = 1;
        return (res, pfpsf);
    }
    if ((x.w[0] == y.w[0]) && (x.w[1] == y.w[1])) {
        res = 1;
        return (res, pfpsf);
    }
    if ((x.w[1] & 0x7800000000000000) == 0x7800000000000000) {
        if ((x.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = 0;
            if ((((y.w[1] & 0x7800000000000000) == 0x7800000000000000)) && ((y.w[1] & 0x8000000000000000) == 0x8000000000000000)) {
                res = 1;
            }
            return (res, pfpsf);
        } else {
            res = 1;
            return (res, pfpsf);
        }
    } else if ((y.w[1] & 0x7800000000000000) == 0x7800000000000000) {
        res = 0;
        if ((y.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    sig_x.w[1] = (x.w[1] & 0x0001ffffffffffff);
    sig_x.w[0] = x.w[0];
    exp_x = ((((go_checked_shr_u64(x.w[1], go_shift_count_u64((49) as u64)))) & 0x000000000003fff) as i64);
    if (((sig_x.w[1] > 0x0001ed09bead87c0) || (((sig_x.w[1] == 0x0001ed09bead87c0) && (sig_x.w[0] > 0x378d8e63ffffffff)))) || (((x.w[1] & 0x6000000000000000) == 0x6000000000000000))) {
        non_canon_x = 1;
    } else {
        non_canon_x = 0;
    }
    exp_y = ((((go_checked_shr_u64(y.w[1], go_shift_count_u64((49) as u64)))) & 0x0000000000003fff) as i64);
    sig_y.w[1] = (y.w[1] & 0x0001ffffffffffff);
    sig_y.w[0] = y.w[0];
    if (((sig_y.w[1] > 0x0001ed09bead87c0) || (((sig_y.w[1] == 0x0001ed09bead87c0) && (sig_y.w[0] > 0x378d8e63ffffffff)))) || (((y.w[1] & 0x6000000000000000) == 0x6000000000000000))) {
        non_canon_y = 1;
    } else {
        non_canon_y = 0;
    }
    if ((non_canon_x != 0) || (((sig_x.w[1] == 0) && (sig_x.w[0] == 0)))) {
        x_is_zero = 1;
    }
    if ((non_canon_y != 0) || (((sig_y.w[1] == 0) && (sig_y.w[0] == 0)))) {
        y_is_zero = 1;
    }
    if ((x_is_zero != 0) && (y_is_zero != 0)) {
        res = 1;
        return (res, pfpsf);
    } else if (x_is_zero != 0) {
        res = 0;
        if ((y.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    } else if (y_is_zero != 0) {
        res = 0;
        if ((x.w[1] & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((((x.w[1] ^ y.w[1]) & 0x8000000000000000)) == 0x8000000000000000) {
        res = 0;
        if ((y.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (exp_y == exp_x) {
        res = 0;
        if ((((sig_x.w[1] > sig_y.w[1]) || (((sig_x.w[1] == sig_y.w[1]) && (sig_x.w[0] >= sig_y.w[0]))))) != (((x.w[1] & 0x8000000000000000) == 0x8000000000000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (((sig_x.w[1] >= sig_y.w[1]) && (sig_x.w[0] >= sig_y.w[0])) && (exp_x > exp_y)) {
        res = 0;
        if ((x.w[1] & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (((sig_x.w[1] <= sig_y.w[1]) && (sig_x.w[0] <= sig_y.w[0])) && (exp_x < exp_y)) {
        res = 0;
        if ((x.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    diff = (exp_x.wrapping_sub(exp_y));
    if (diff > 0) {
        if (diff > 33) {
            res = 0;
            if ((x.w[1] & 0x8000000000000000) != 0x8000000000000000) {
                res = 1;
            }
            return (res, pfpsf);
        }
        if (diff > 19) {
            sig_n_prime256 = __mul_128x128_to_256(sig_x, bid_ten2k128[(diff.wrapping_sub(20)) as usize]);
            if ((((sig_n_prime256.w[3] == 0) && (sig_n_prime256.w[2] == 0)) && (sig_n_prime256.w[1] == sig_y.w[1])) && (sig_n_prime256.w[0] == sig_y.w[0])) {
                res = 1;
                return (res, pfpsf);
            }
            res = 0;
            if (((((((sig_n_prime256.w[3] > 0) || (sig_n_prime256.w[2] > 0))) || (sig_n_prime256.w[1] > sig_y.w[1])) || (((sig_n_prime256.w[1] == sig_y.w[1]) && (sig_n_prime256.w[0] > sig_y.w[0]))))) != (((y.w[1] & 0x8000000000000000) == 0x8000000000000000))) {
                res = 1;
            }
            return (res, pfpsf);
        }
        sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[diff as usize], sig_x);
        if (((sig_n_prime192.w[2] == 0) && (sig_n_prime192.w[1] == sig_y.w[1])) && (sig_n_prime192.w[0] == sig_y.w[0])) {
            res = 1;
            return (res, pfpsf);
        }
        res = 0;
        if (((((sig_n_prime192.w[2] > 0) || (sig_n_prime192.w[1] > sig_y.w[1])) || (((sig_n_prime192.w[1] == sig_y.w[1]) && (sig_n_prime192.w[0] > sig_y.w[0]))))) != (((y.w[1] & 0x8000000000000000) == 0x8000000000000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    diff = (exp_y.wrapping_sub(exp_x));
    if (diff > 33) {
        res = 0;
        if ((x.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (diff > 19) {
        sig_n_prime256 = __mul_128x128_to_256(sig_y, bid_ten2k128[(diff.wrapping_sub(20)) as usize]);
        if ((((sig_n_prime256.w[3] == 0) && (sig_n_prime256.w[2] == 0)) && (sig_n_prime256.w[1] == sig_x.w[1])) && (sig_n_prime256.w[0] == sig_x.w[0])) {
            res = 1;
            return (res, pfpsf);
        }
        res = 0;
        if (((((sig_n_prime256.w[3] == 0) && (sig_n_prime256.w[2] == 0)) && (((sig_n_prime256.w[1] < sig_x.w[1]) || (((sig_n_prime256.w[1] == sig_x.w[1]) && (sig_n_prime256.w[0] < sig_x.w[0]))))))) != (((x.w[1] & 0x8000000000000000) == 0x8000000000000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[diff as usize], sig_y);
    if (((sig_n_prime192.w[2] == 0) && (sig_n_prime192.w[1] == sig_x.w[1])) && (sig_n_prime192.w[0] == sig_x.w[0])) {
        res = 1;
        return (res, pfpsf);
    }
    res = 0;
    if ((((sig_n_prime192.w[2] == 0) && (((sig_n_prime192.w[1] < sig_x.w[1]) || (((sig_n_prime192.w[1] == sig_x.w[1]) && (sig_n_prime192.w[0] < sig_x.w[0]))))))) != (((y.w[1] & 0x8000000000000000) == 0x8000000000000000))) {
        res = 1;
    }
    return (res, pfpsf);
}

pub fn bid128_total_order(mut x: BID_UINT128, mut y: BID_UINT128) -> i64 {
    let mut res: i64 = 0;
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut sig_x: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut sig_y: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut pyld_y: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut pyld_x: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut sig_n_prime192: BID_UINT192 = BID_UINT192 { w: [0, 0, 0] };
    let mut sig_n_prime256: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut x_is_zero: i64 = 0;
    let mut y_is_zero: i64 = 0;
    if ((x.w[1] & 0x7c00000000000000) == 0x7c00000000000000) {
        if ((x.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            if (((y.w[1] & 0x7c00000000000000) != 0x7c00000000000000) || ((y.w[1] & 0x8000000000000000) != 0x8000000000000000)) {
                res = 1;
                return res;
            } else {
                pyld_x.w[1] = (x.w[1] & 0x00003fffffffffff);
                pyld_x.w[0] = x.w[0];
                pyld_y.w[1] = (y.w[1] & 0x00003fffffffffff);
                pyld_y.w[0] = y.w[0];
                if ((pyld_x.w[1] > 0x0000314dc6448d93) || (((pyld_x.w[1] == 0x0000314dc6448d93) && (pyld_x.w[0] > 0x38c15b09ffffffff)))) {
                    pyld_x.w[1] = 0;
                    pyld_x.w[0] = 0;
                }
                if ((pyld_y.w[1] > 0x0000314dc6448d93) || (((pyld_y.w[1] == 0x0000314dc6448d93) && (pyld_y.w[0] > 0x38c15b09ffffffff)))) {
                    pyld_y.w[1] = 0;
                    pyld_y.w[0] = 0;
                }
                if (!(((((y.w[1] & 0x7e00000000000000) == 0x7e00000000000000)) != (((x.w[1] & 0x7e00000000000000) == 0x7e00000000000000))))) {
                    if ((pyld_x.w[1] > pyld_y.w[1]) || (((pyld_x.w[1] == pyld_y.w[1]) && (pyld_x.w[0] >= pyld_y.w[0])))) {
                        res = 1;
                    } else {
                        res = 0;
                    }
                    return res;
                } else {
                    res = 0;
                    if ((y.w[1] & 0x7e00000000000000) == 0x7e00000000000000) {
                        res = 1;
                    }
                    return res;
                }
            }
        } else {
            if (((y.w[1] & 0x7c00000000000000) != 0x7c00000000000000) || ((y.w[1] & 0x8000000000000000) == 0x8000000000000000)) {
                res = 0;
                return res;
            } else {
                pyld_x.w[1] = (x.w[1] & 0x00003fffffffffff);
                pyld_x.w[0] = x.w[0];
                pyld_y.w[1] = (y.w[1] & 0x00003fffffffffff);
                pyld_y.w[0] = y.w[0];
                if ((pyld_x.w[1] > 0x0000314dc6448d93) || (((pyld_x.w[1] == 0x0000314dc6448d93) && (pyld_x.w[0] > 0x38c15b09ffffffff)))) {
                    pyld_x.w[1] = 0;
                    pyld_x.w[0] = 0;
                }
                if ((pyld_y.w[1] > 0x0000314dc6448d93) || (((pyld_y.w[1] == 0x0000314dc6448d93) && (pyld_y.w[0] > 0x38c15b09ffffffff)))) {
                    pyld_y.w[1] = 0;
                    pyld_y.w[0] = 0;
                }
                if (!(((((y.w[1] & 0x7e00000000000000) == 0x7e00000000000000)) != (((x.w[1] & 0x7e00000000000000) == 0x7e00000000000000))))) {
                    if ((pyld_x.w[1] < pyld_y.w[1]) || (((pyld_x.w[1] == pyld_y.w[1]) && (pyld_x.w[0] <= pyld_y.w[0])))) {
                        res = 1;
                    } else {
                        res = 0;
                    }
                    return res;
                } else {
                    res = 0;
                    if ((x.w[1] & 0x7e00000000000000) == 0x7e00000000000000) {
                        res = 1;
                    }
                    return res;
                }
            }
        }
    } else if ((y.w[1] & 0x7c00000000000000) == 0x7c00000000000000) {
        res = 0;
        if ((y.w[1] & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return res;
    }
    if ((x.w[1] == y.w[1]) && (x.w[0] == y.w[0])) {
        res = 1;
        return res;
    }
    if ((((x.w[1] & 0x8000000000000000) == 0x8000000000000000)) != (((y.w[1] & 0x8000000000000000) == 0x8000000000000000))) {
        res = 0;
        if ((x.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return res;
    }
    if ((x.w[1] & 0x7800000000000000) == 0x7800000000000000) {
        if ((x.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
            return res;
        } else {
            res = 0;
            if ((y.w[1] & 0x7800000000000000) == 0x7800000000000000) {
                res = 1;
            }
            return res;
        }
    } else if ((y.w[1] & 0x7800000000000000) == 0x7800000000000000) {
        res = 0;
        if ((y.w[1] & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return res;
    }
    sig_x.w[1] = (x.w[1] & 0x0001ffffffffffff);
    sig_x.w[0] = x.w[0];
    exp_x = ((((go_checked_shr_u64(x.w[1], go_shift_count_u64((49) as u64)))) & 0x000000000003fff) as i64);
    if (((((((sig_x.w[1] > 0x0001ed09bead87c0) || (((sig_x.w[1] == 0x0001ed09bead87c0) && (sig_x.w[0] > 0x378d8e63ffffffff))))) && (((x.w[1] & 0x6000000000000000) != 0x6000000000000000)))) || (((x.w[1] & 0x6000000000000000) == 0x6000000000000000))) || (((sig_x.w[1] == 0) && (sig_x.w[0] == 0)))) {
        x_is_zero = 1;
        if ((x.w[1] & 0x6000000000000000) == 0x6000000000000000) {
            exp_x = ((((go_checked_shr_u64(x.w[1], go_shift_count_u64((47) as u64)))) & 0x000000000003fff) as i64);
        }
    }
    exp_y = ((((go_checked_shr_u64(y.w[1], go_shift_count_u64((49) as u64)))) & 0x0000000000003fff) as i64);
    sig_y.w[1] = (y.w[1] & 0x0001ffffffffffff);
    sig_y.w[0] = y.w[0];
    if (((((((sig_y.w[1] > 0x0001ed09bead87c0) || (((sig_y.w[1] == 0x0001ed09bead87c0) && (sig_y.w[0] > 0x378d8e63ffffffff))))) && (((y.w[1] & 0x6000000000000000) != 0x6000000000000000)))) || (((y.w[1] & 0x6000000000000000) == 0x6000000000000000))) || (((sig_y.w[1] == 0) && (sig_y.w[0] == 0)))) {
        y_is_zero = 1;
        if ((y.w[1] & 0x6000000000000000) == 0x6000000000000000) {
            exp_y = ((((go_checked_shr_u64(y.w[1], go_shift_count_u64((47) as u64)))) & 0x000000000003fff) as i64);
        }
    }
    if ((x_is_zero != 0) && (y_is_zero != 0)) {
        if (exp_x == exp_y) {
            res = 1;
            return res;
        }
        res = 0;
        if ((exp_x <= exp_y) != (((x.w[1] & 0x8000000000000000) == 0x8000000000000000))) {
            res = 1;
        }
        return res;
    }
    if (x_is_zero != 0) {
        res = 0;
        if ((y.w[1] & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return res;
    }
    if (y_is_zero != 0) {
        res = 0;
        if ((x.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return res;
    }
    if ((((sig_x.w[1] > sig_y.w[1]) || (((sig_x.w[1] == sig_y.w[1]) && (sig_x.w[0] > sig_y.w[0]))))) && (exp_x >= exp_y)) {
        res = 0;
        if ((x.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return res;
    }
    if ((((sig_x.w[1] < sig_y.w[1]) || (((sig_x.w[1] == sig_y.w[1]) && (sig_x.w[0] < sig_y.w[0]))))) && (exp_x <= exp_y)) {
        res = 0;
        if ((x.w[1] & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return res;
    }
    if (exp_x > exp_y) {
        if ((exp_x.wrapping_sub(exp_y)) > 33) {
            res = 0;
            if ((x.w[1] & 0x8000000000000000) == 0x8000000000000000) {
                res = 1;
            }
            return res;
        }
        if ((exp_x.wrapping_sub(exp_y)) > 19) {
            sig_n_prime256 = __mul_128x128_to_256(sig_x, bid_ten2k128[((exp_x.wrapping_sub(exp_y)).wrapping_sub(20)) as usize]);
            if ((((sig_n_prime256.w[3] == 0) && (sig_n_prime256.w[2] == 0)) && (sig_n_prime256.w[1] == sig_y.w[1])) && (sig_n_prime256.w[0] == sig_y.w[0])) {
                res = 0;
                if ((exp_x <= exp_y) != (((x.w[1] & 0x8000000000000000) == 0x8000000000000000))) {
                    res = 1;
                }
                return res;
            }
            res = 0;
            if (((((sig_n_prime256.w[3] == 0) && (sig_n_prime256.w[2] == 0)) && (((sig_n_prime256.w[1] < sig_y.w[1]) || (((sig_n_prime256.w[1] == sig_y.w[1]) && (sig_n_prime256.w[0] < sig_y.w[0]))))))) != (((x.w[1] & 0x8000000000000000) == 0x8000000000000000))) {
                res = 1;
            }
            return res;
        }
        sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[(exp_x.wrapping_sub(exp_y)) as usize], sig_x);
        if (((sig_n_prime192.w[2] == 0) && (sig_n_prime192.w[1] == sig_y.w[1])) && (sig_n_prime192.w[0] == sig_y.w[0])) {
            res = 0;
            if ((exp_x <= exp_y) != (((x.w[1] & 0x8000000000000000) == 0x8000000000000000))) {
                res = 1;
            }
            return res;
        }
        res = 0;
        if ((((sig_n_prime192.w[2] == 0) && (((sig_n_prime192.w[1] < sig_y.w[1]) || (((sig_n_prime192.w[1] == sig_y.w[1]) && (sig_n_prime192.w[0] < sig_y.w[0]))))))) != (((x.w[1] & 0x8000000000000000) == 0x8000000000000000))) {
            res = 1;
        }
        return res;
    }
    if ((exp_y.wrapping_sub(exp_x)) > 33) {
        res = 0;
        if ((x.w[1] & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return res;
    }
    if ((exp_y.wrapping_sub(exp_x)) > 19) {
        sig_n_prime256 = __mul_128x128_to_256(sig_y, bid_ten2k128[((exp_y.wrapping_sub(exp_x)).wrapping_sub(20)) as usize]);
        if ((((sig_n_prime256.w[3] == 0) && (sig_n_prime256.w[2] == 0)) && (sig_n_prime256.w[1] == sig_x.w[1])) && (sig_n_prime256.w[0] == sig_x.w[0])) {
            res = 0;
            if ((exp_x <= exp_y) != (((x.w[1] & 0x8000000000000000) == 0x8000000000000000))) {
                res = 1;
            }
            return res;
        }
        res = 0;
        if ((((((sig_n_prime256.w[3] != 0) || (sig_n_prime256.w[2] != 0)) || (sig_n_prime256.w[1] > sig_x.w[1])) || (((sig_n_prime256.w[1] == sig_x.w[1]) && (sig_n_prime256.w[0] > sig_x.w[0]))))) != (((x.w[1] & 0x8000000000000000) == 0x8000000000000000))) {
            res = 1;
        }
        return res;
    }
    sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[(exp_y.wrapping_sub(exp_x)) as usize], sig_y);
    if (((sig_n_prime192.w[2] == 0) && (sig_n_prime192.w[1] == sig_x.w[1])) && (sig_n_prime192.w[0] == sig_x.w[0])) {
        res = 0;
        if ((exp_x <= exp_y) != (((x.w[1] & 0x8000000000000000) == 0x8000000000000000))) {
            res = 1;
        }
        return res;
    }
    res = 0;
    if (((((sig_n_prime192.w[2] != 0) || (sig_n_prime192.w[1] > sig_x.w[1])) || (((sig_n_prime192.w[1] == sig_x.w[1]) && (sig_n_prime192.w[0] > sig_x.w[0]))))) != (((x.w[1] & 0x8000000000000000) == 0x8000000000000000))) {
        res = 1;
    }
    return res;
}

pub fn bid128_total_order_mag(mut x: BID_UINT128, mut y: BID_UINT128) -> i64 {
    let mut res: i64 = 0;
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut sig_x: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut sig_y: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut pyld_y: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut pyld_x: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut sig_n_prime192: BID_UINT192 = BID_UINT192 { w: [0, 0, 0] };
    let mut sig_n_prime256: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut x_is_zero: i64 = 0;
    let mut y_is_zero: i64 = 0;
    x.w[1] = (x.w[1] & 0x7fffffffffffffff);
    y.w[1] = (y.w[1] & 0x7fffffffffffffff);
    if ((x.w[1] & 0x7c00000000000000) == 0x7c00000000000000) {
        if ((y.w[1] & 0x7c00000000000000) != 0x7c00000000000000) {
            res = 0;
            return res;
        } else {
            pyld_x.w[1] = (x.w[1] & 0x00003fffffffffff);
            pyld_x.w[0] = x.w[0];
            pyld_y.w[1] = (y.w[1] & 0x00003fffffffffff);
            pyld_y.w[0] = y.w[0];
            if ((pyld_x.w[1] > 0x0000314dc6448d93) || (((pyld_x.w[1] == 0x0000314dc6448d93) && (pyld_x.w[0] > 0x38c15b09ffffffff)))) {
                pyld_x.w[1] = 0;
                pyld_x.w[0] = 0;
            }
            if ((pyld_y.w[1] > 0x0000314dc6448d93) || (((pyld_y.w[1] == 0x0000314dc6448d93) && (pyld_y.w[0] > 0x38c15b09ffffffff)))) {
                pyld_y.w[1] = 0;
                pyld_y.w[0] = 0;
            }
            if (!(((((y.w[1] & 0x7e00000000000000) == 0x7e00000000000000)) != (((x.w[1] & 0x7e00000000000000) == 0x7e00000000000000))))) {
                if ((pyld_x.w[1] < pyld_y.w[1]) || (((pyld_x.w[1] == pyld_y.w[1]) && (pyld_x.w[0] <= pyld_y.w[0])))) {
                    res = 1;
                } else {
                    res = 0;
                }
                return res;
            } else {
                res = 0;
                if ((x.w[1] & 0x7e00000000000000) == 0x7e00000000000000) {
                    res = 1;
                }
                return res;
            }
        }
    } else if ((y.w[1] & 0x7c00000000000000) == 0x7c00000000000000) {
        res = 1;
        return res;
    }
    if ((x.w[1] == y.w[1]) && (x.w[0] == y.w[0])) {
        res = 1;
        return res;
    }
    if ((x.w[1] & 0x7800000000000000) == 0x7800000000000000) {
        res = 0;
        if ((y.w[1] & 0x7800000000000000) == 0x7800000000000000) {
            res = 1;
        }
        return res;
    } else if ((y.w[1] & 0x7800000000000000) == 0x7800000000000000) {
        res = 1;
        return res;
    }
    sig_x.w[1] = (x.w[1] & 0x0001ffffffffffff);
    sig_x.w[0] = x.w[0];
    exp_x = ((((go_checked_shr_u64(x.w[1], go_shift_count_u64((49) as u64)))) & 0x000000000003fff) as i64);
    if (((((((sig_x.w[1] > 0x0001ed09bead87c0) || (((sig_x.w[1] == 0x0001ed09bead87c0) && (sig_x.w[0] > 0x378d8e63ffffffff))))) && (((x.w[1] & 0x6000000000000000) != 0x6000000000000000)))) || (((x.w[1] & 0x6000000000000000) == 0x6000000000000000))) || (((sig_x.w[1] == 0) && (sig_x.w[0] == 0)))) {
        x_is_zero = 1;
        if ((x.w[1] & 0x6000000000000000) == 0x6000000000000000) {
            exp_x = ((((go_checked_shr_u64(x.w[1], go_shift_count_u64((47) as u64)))) & 0x000000000003fff) as i64);
        }
    }
    exp_y = ((((go_checked_shr_u64(y.w[1], go_shift_count_u64((49) as u64)))) & 0x0000000000003fff) as i64);
    sig_y.w[1] = (y.w[1] & 0x0001ffffffffffff);
    sig_y.w[0] = y.w[0];
    if (((((((sig_y.w[1] > 0x0001ed09bead87c0) || (((sig_y.w[1] == 0x0001ed09bead87c0) && (sig_y.w[0] > 0x378d8e63ffffffff))))) && (((y.w[1] & 0x6000000000000000) != 0x6000000000000000)))) || (((y.w[1] & 0x6000000000000000) == 0x6000000000000000))) || (((sig_y.w[1] == 0) && (sig_y.w[0] == 0)))) {
        y_is_zero = 1;
        if ((y.w[1] & 0x6000000000000000) == 0x6000000000000000) {
            exp_y = ((((go_checked_shr_u64(y.w[1], go_shift_count_u64((47) as u64)))) & 0x000000000003fff) as i64);
        }
    }
    if ((x_is_zero != 0) && (y_is_zero != 0)) {
        if (exp_x == exp_y) {
            res = 1;
            return res;
        }
        res = 0;
        if (exp_x <= exp_y) {
            res = 1;
        }
        return res;
    }
    if (x_is_zero != 0) {
        res = 1;
        return res;
    }
    if (y_is_zero != 0) {
        res = 0;
        return res;
    }
    if ((((sig_x.w[1] > sig_y.w[1]) || (((sig_x.w[1] == sig_y.w[1]) && (sig_x.w[0] > sig_y.w[0]))))) && (exp_x >= exp_y)) {
        res = 0;
        return res;
    }
    if ((((sig_x.w[1] < sig_y.w[1]) || (((sig_x.w[1] == sig_y.w[1]) && (sig_x.w[0] < sig_y.w[0]))))) && (exp_x <= exp_y)) {
        res = 1;
        return res;
    }
    if (exp_x > exp_y) {
        if ((exp_x.wrapping_sub(exp_y)) > 33) {
            res = 0;
            return res;
        }
        if ((exp_x.wrapping_sub(exp_y)) > 19) {
            sig_n_prime256 = __mul_128x128_to_256(sig_x, bid_ten2k128[((exp_x.wrapping_sub(exp_y)).wrapping_sub(20)) as usize]);
            if ((((sig_n_prime256.w[3] == 0) && (sig_n_prime256.w[2] == 0)) && (sig_n_prime256.w[1] == sig_y.w[1])) && (sig_n_prime256.w[0] == sig_y.w[0])) {
                res = 0;
                return res;
            }
            res = 0;
            if (((sig_n_prime256.w[3] == 0) && (sig_n_prime256.w[2] == 0)) && (((sig_n_prime256.w[1] < sig_y.w[1]) || (((sig_n_prime256.w[1] == sig_y.w[1]) && (sig_n_prime256.w[0] < sig_y.w[0])))))) {
                res = 1;
            }
            return res;
        }
        sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[(exp_x.wrapping_sub(exp_y)) as usize], sig_x);
        if (((sig_n_prime192.w[2] == 0) && (sig_n_prime192.w[1] == sig_y.w[1])) && (sig_n_prime192.w[0] == sig_y.w[0])) {
            res = 0;
            return res;
        }
        res = 0;
        if ((sig_n_prime192.w[2] == 0) && (((sig_n_prime192.w[1] < sig_y.w[1]) || (((sig_n_prime192.w[1] == sig_y.w[1]) && (sig_n_prime192.w[0] < sig_y.w[0])))))) {
            res = 1;
        }
        return res;
    }
    if ((exp_y.wrapping_sub(exp_x)) > 33) {
        res = 1;
        return res;
    }
    if ((exp_y.wrapping_sub(exp_x)) > 19) {
        sig_n_prime256 = __mul_128x128_to_256(sig_y, bid_ten2k128[((exp_y.wrapping_sub(exp_x)).wrapping_sub(20)) as usize]);
        if ((((sig_n_prime256.w[3] == 0) && (sig_n_prime256.w[2] == 0)) && (sig_n_prime256.w[1] == sig_x.w[1])) && (sig_n_prime256.w[0] == sig_x.w[0])) {
            res = 1;
            return res;
        }
        res = 0;
        if ((((sig_n_prime256.w[3] != 0) || (sig_n_prime256.w[2] != 0)) || (sig_n_prime256.w[1] > sig_x.w[1])) || (((sig_n_prime256.w[1] == sig_x.w[1]) && (sig_n_prime256.w[0] > sig_x.w[0])))) {
            res = 1;
        }
        return res;
    }
    sig_n_prime192 = __mul_64x128_to_192(bid_ten2k64[(exp_y.wrapping_sub(exp_x)) as usize], sig_y);
    if (((sig_n_prime192.w[2] == 0) && (sig_n_prime192.w[1] == sig_x.w[1])) && (sig_n_prime192.w[0] == sig_x.w[0])) {
        res = 1;
        return res;
    }
    res = 0;
    if (((sig_n_prime192.w[2] != 0) || (sig_n_prime192.w[1] > sig_x.w[1])) || (((sig_n_prime192.w[1] == sig_x.w[1]) && (sig_n_prime192.w[0] > sig_x.w[0])))) {
        res = 1;
    }
    return res;
}
