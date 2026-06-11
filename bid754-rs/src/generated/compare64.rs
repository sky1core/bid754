// Auto-generated from compare64.go by go2rs. Do not edit.

use super::prelude::*;

pub fn bid64_quiet_equal(mut x: u64, mut y: u64) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut exp_t: i64 = 0;
    let mut sig_x: u64 = 0;
    let mut sig_y: u64 = 0;
    let mut sig_t: u64 = 0;
    let mut pfpsf: u32 = 0;
    let mut x_is_zero: i64 = 0;
    let mut y_is_zero: i64 = 0;
    let mut non_canon_x: i64 = 0;
    let mut non_canon_y: i64 = 0;
    let mut lcv: i64 = 0;
    if ((((x & 0x7c00000000000000) == 0x7c00000000000000)) || (((y & 0x7c00000000000000) == 0x7c00000000000000))) {
        if (((x & 0x7e00000000000000) == 0x7e00000000000000) || ((y & 0x7e00000000000000) == 0x7e00000000000000)) {
            pfpsf |= 1;
        }
        res = 0;
        return (res, pfpsf);
    }
    if (x == y) {
        res = 1;
        return (res, pfpsf);
    }
    if ((((x & 0x7800000000000000) == 0x7800000000000000)) && (((y & 0x7800000000000000) == 0x7800000000000000))) {
        res = 0;
        if ((((x ^ y) & 0x8000000000000000)) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((((x & 0x7800000000000000) == 0x7800000000000000)) || (((y & 0x7800000000000000) == 0x7800000000000000))) {
        res = 0;
        return (res, pfpsf);
    }
    if ((x & 0x6000000000000000) == 0x6000000000000000) {
        exp_x = ((go_checked_shr_u64((x & 0x1ff8000000000000), go_shift_count_u64((51) as u64))) as i64);
        sig_x = ((x & 0x7ffffffffffff) | 0x20000000000000);
        if (sig_x > 9999999999999999) {
            non_canon_x = 1;
        } else {
            non_canon_x = 0;
        }
    } else {
        exp_x = ((go_checked_shr_u64((x & 0x7fe0000000000000), go_shift_count_u64((53) as u64))) as i64);
        sig_x = (x & 0x1fffffffffffff);
        non_canon_x = 0;
    }
    if ((y & 0x6000000000000000) == 0x6000000000000000) {
        exp_y = ((go_checked_shr_u64((y & 0x1ff8000000000000), go_shift_count_u64((51) as u64))) as i64);
        sig_y = ((y & 0x7ffffffffffff) | 0x20000000000000);
        if (sig_y > 9999999999999999) {
            non_canon_y = 1;
        } else {
            non_canon_y = 0;
        }
    } else {
        exp_y = ((go_checked_shr_u64((y & 0x7fe0000000000000), go_shift_count_u64((53) as u64))) as i64);
        sig_y = (y & 0x1fffffffffffff);
        non_canon_y = 0;
    }
    if ((non_canon_x != 0) || (sig_x == 0)) {
        x_is_zero = 1;
    }
    if ((non_canon_y != 0) || (sig_y == 0)) {
        y_is_zero = 1;
    }
    if ((x_is_zero != 0) && (y_is_zero != 0)) {
        res = 1;
        return (res, pfpsf);
    } else if ((((x_is_zero != 0) && (y_is_zero == 0))) || (((x_is_zero == 0) && (y_is_zero != 0)))) {
        res = 0;
        return (res, pfpsf);
    }
    if ((((x ^ y) & 0x8000000000000000)) != 0) {
        res = 0;
        return (res, pfpsf);
    }
    if (exp_x > exp_y) {
        exp_t = exp_x;
        exp_x = exp_y;
        exp_y = exp_t;
        sig_t = sig_x;
        sig_x = sig_y;
        sig_y = sig_t;
    }
    if ((exp_y.wrapping_sub(exp_x)) > 15) {
        res = 0;
        return (res, pfpsf);
    }
    lcv = 0;
    while (lcv < ((exp_y.wrapping_sub(exp_x)))) {
        sig_y = (sig_y.wrapping_mul(10));
        if (sig_y > 9999999999999999) {
            res = 0;
            return (res, pfpsf);
        }
        lcv = lcv.wrapping_add(1);
    }
    res = 0;
    if (sig_y == sig_x) {
        res = 1;
    }
    return (res, pfpsf);
}

pub fn bid64_quiet_greater(mut x: u64, mut y: u64) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut sig_x: u64 = 0;
    let mut sig_y: u64 = 0;
    let mut sig_n_prime: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut pfpsf: u32 = 0;
    let mut x_is_zero: i64 = 0;
    let mut y_is_zero: i64 = 0;
    let mut non_canon_x: i64 = 0;
    let mut non_canon_y: i64 = 0;
    if ((((x & 0x7c00000000000000) == 0x7c00000000000000)) || (((y & 0x7c00000000000000) == 0x7c00000000000000))) {
        if (((x & 0x7e00000000000000) == 0x7e00000000000000) || ((y & 0x7e00000000000000) == 0x7e00000000000000)) {
            pfpsf |= 1;
        }
        res = 0;
        return (res, pfpsf);
    }
    if (x == y) {
        res = 0;
        return (res, pfpsf);
    }
    if ((x & 0x7800000000000000) == 0x7800000000000000) {
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = 0;
            return (res, pfpsf);
        } else {
            res = 0;
            if ((((y & 0x7800000000000000) != 0x7800000000000000)) || (((y & 0x8000000000000000) == 0x8000000000000000))) {
                res = 1;
            }
            return (res, pfpsf);
        }
    } else if ((y & 0x7800000000000000) == 0x7800000000000000) {
        res = 0;
        if ((y & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((x & 0x6000000000000000) == 0x6000000000000000) {
        exp_x = ((go_checked_shr_u64((x & 0x1ff8000000000000), go_shift_count_u64((51) as u64))) as i64);
        sig_x = ((x & 0x7ffffffffffff) | 0x20000000000000);
        if (sig_x > 9999999999999999) {
            non_canon_x = 1;
        } else {
            non_canon_x = 0;
        }
    } else {
        exp_x = ((go_checked_shr_u64((x & 0x7fe0000000000000), go_shift_count_u64((53) as u64))) as i64);
        sig_x = (x & 0x1fffffffffffff);
        non_canon_x = 0;
    }
    if ((y & 0x6000000000000000) == 0x6000000000000000) {
        exp_y = ((go_checked_shr_u64((y & 0x1ff8000000000000), go_shift_count_u64((51) as u64))) as i64);
        sig_y = ((y & 0x7ffffffffffff) | 0x20000000000000);
        if (sig_y > 9999999999999999) {
            non_canon_y = 1;
        } else {
            non_canon_y = 0;
        }
    } else {
        exp_y = ((go_checked_shr_u64((y & 0x7fe0000000000000), go_shift_count_u64((53) as u64))) as i64);
        sig_y = (y & 0x1fffffffffffff);
        non_canon_y = 0;
    }
    if ((non_canon_x != 0) || (sig_x == 0)) {
        x_is_zero = 1;
    }
    if ((non_canon_y != 0) || (sig_y == 0)) {
        y_is_zero = 1;
    }
    if ((x_is_zero != 0) && (y_is_zero != 0)) {
        res = 0;
        return (res, pfpsf);
    } else if (x_is_zero != 0) {
        res = 0;
        if ((y & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    } else if (y_is_zero != 0) {
        res = 0;
        if ((x & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((((x ^ y) & 0x8000000000000000)) == 0x8000000000000000) {
        res = 0;
        if ((y & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((sig_x > sig_y) && (exp_x > exp_y)) {
        res = 0;
        if ((x & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((sig_x < sig_y) && (exp_x < exp_y)) {
        res = 0;
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((exp_x.wrapping_sub(exp_y)) > 15) {
        if ((x & 0x8000000000000000) != 0) {
            res = 0;
        } else {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((exp_y.wrapping_sub(exp_x)) > 15) {
        if ((x & 0x8000000000000000) != 0) {
            res = 1;
        } else {
            res = 0;
        }
        return (res, pfpsf);
    }
    if (exp_x > exp_y) {
        sig_n_prime = __mul_64x64_to_128(sig_x, bid_mult_factor[(exp_x.wrapping_sub(exp_y)) as usize]);
        if ((sig_n_prime.w[1] == 0) && (sig_n_prime.w[0] == sig_y)) {
            res = 0;
            return (res, pfpsf);
        }
        res = 0;
        if ((((sig_n_prime.w[1] > 0) || (sig_n_prime.w[0] > sig_y))) != (((x & 0x8000000000000000) == 0x8000000000000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    sig_n_prime = __mul_64x64_to_128(sig_y, bid_mult_factor[(exp_y.wrapping_sub(exp_x)) as usize]);
    if ((sig_n_prime.w[1] == 0) && (sig_n_prime.w[0] == sig_x)) {
        res = 0;
        return (res, pfpsf);
    }
    res = 0;
    if ((((sig_n_prime.w[1] == 0) && (sig_x > sig_n_prime.w[0]))) != (((x & 0x8000000000000000) == 0x8000000000000000))) {
        res = 1;
    }
    return (res, pfpsf);
}

pub fn bid64_quiet_greater_equal(mut x: u64, mut y: u64) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut sig_x: u64 = 0;
    let mut sig_y: u64 = 0;
    let mut sig_n_prime: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut pfpsf: u32 = 0;
    let mut x_is_zero: i64 = 0;
    let mut y_is_zero: i64 = 0;
    let mut non_canon_x: i64 = 0;
    let mut non_canon_y: i64 = 0;
    if ((((x & 0x7c00000000000000) == 0x7c00000000000000)) || (((y & 0x7c00000000000000) == 0x7c00000000000000))) {
        if (((x & 0x7e00000000000000) == 0x7e00000000000000) || ((y & 0x7e00000000000000) == 0x7e00000000000000)) {
            pfpsf |= 1;
        }
        res = 0;
        return (res, pfpsf);
    }
    if (x == y) {
        res = 1;
        return (res, pfpsf);
    }
    if ((x & 0x7800000000000000) == 0x7800000000000000) {
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = 0;
            if ((((y & 0x7800000000000000) == 0x7800000000000000)) && ((y & 0x8000000000000000) == 0x8000000000000000)) {
                res = 1;
            }
            return (res, pfpsf);
        } else {
            res = 1;
            return (res, pfpsf);
        }
    } else if ((y & 0x7800000000000000) == 0x7800000000000000) {
        res = 0;
        if ((y & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((x & 0x6000000000000000) == 0x6000000000000000) {
        exp_x = ((go_checked_shr_u64((x & 0x1ff8000000000000), go_shift_count_u64((51) as u64))) as i64);
        sig_x = ((x & 0x7ffffffffffff) | 0x20000000000000);
        if (sig_x > 9999999999999999) {
            non_canon_x = 1;
        } else {
            non_canon_x = 0;
        }
    } else {
        exp_x = ((go_checked_shr_u64((x & 0x7fe0000000000000), go_shift_count_u64((53) as u64))) as i64);
        sig_x = (x & 0x1fffffffffffff);
        non_canon_x = 0;
    }
    if ((y & 0x6000000000000000) == 0x6000000000000000) {
        exp_y = ((go_checked_shr_u64((y & 0x1ff8000000000000), go_shift_count_u64((51) as u64))) as i64);
        sig_y = ((y & 0x7ffffffffffff) | 0x20000000000000);
        if (sig_y > 9999999999999999) {
            non_canon_y = 1;
        } else {
            non_canon_y = 0;
        }
    } else {
        exp_y = ((go_checked_shr_u64((y & 0x7fe0000000000000), go_shift_count_u64((53) as u64))) as i64);
        sig_y = (y & 0x1fffffffffffff);
        non_canon_y = 0;
    }
    if ((non_canon_x != 0) || (sig_x == 0)) {
        x_is_zero = 1;
    }
    if ((non_canon_y != 0) || (sig_y == 0)) {
        y_is_zero = 1;
    }
    if ((x_is_zero != 0) && (y_is_zero != 0)) {
        res = 1;
        return (res, pfpsf);
    } else if (x_is_zero != 0) {
        res = 0;
        if ((y & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    } else if (y_is_zero != 0) {
        res = 0;
        if ((x & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((((x ^ y) & 0x8000000000000000)) == 0x8000000000000000) {
        res = 0;
        if ((y & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((sig_x > sig_y) && (exp_x >= exp_y)) {
        res = 0;
        if ((x & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((sig_x < sig_y) && (exp_x <= exp_y)) {
        res = 0;
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((exp_x.wrapping_sub(exp_y)) > 15) {
        res = 0;
        if ((x & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((exp_y.wrapping_sub(exp_x)) > 15) {
        res = 0;
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (exp_x > exp_y) {
        sig_n_prime = __mul_64x64_to_128(sig_x, bid_mult_factor[(exp_x.wrapping_sub(exp_y)) as usize]);
        if ((sig_n_prime.w[1] == 0) && (sig_n_prime.w[0] == sig_y)) {
            res = 1;
            return (res, pfpsf);
        }
        res = 0;
        if ((((sig_n_prime.w[1] == 0) && (sig_n_prime.w[0] < sig_y))) != (((x & 0x8000000000000000) != 0x8000000000000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    sig_n_prime = __mul_64x64_to_128(sig_y, bid_mult_factor[(exp_y.wrapping_sub(exp_x)) as usize]);
    if ((sig_n_prime.w[1] == 0) && (sig_n_prime.w[0] == sig_x)) {
        res = 1;
        return (res, pfpsf);
    }
    res = 0;
    if ((((sig_n_prime.w[1] > 0) || (sig_x < sig_n_prime.w[0]))) != (((x & 0x8000000000000000) != 0x8000000000000000))) {
        res = 1;
    }
    return (res, pfpsf);
}

pub fn bid64_quiet_greater_unordered(mut x: u64, mut y: u64) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut sig_x: u64 = 0;
    let mut sig_y: u64 = 0;
    let mut sig_n_prime: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut pfpsf: u32 = 0;
    let mut x_is_zero: i64 = 0;
    let mut y_is_zero: i64 = 0;
    let mut non_canon_x: i64 = 0;
    let mut non_canon_y: i64 = 0;
    if ((((x & 0x7c00000000000000) == 0x7c00000000000000)) || (((y & 0x7c00000000000000) == 0x7c00000000000000))) {
        if (((x & 0x7e00000000000000) == 0x7e00000000000000) || ((y & 0x7e00000000000000) == 0x7e00000000000000)) {
            pfpsf |= 1;
        }
        res = 1;
        return (res, pfpsf);
    }
    if (x == y) {
        res = 0;
        return (res, pfpsf);
    }
    if ((x & 0x7800000000000000) == 0x7800000000000000) {
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = 0;
            return (res, pfpsf);
        } else {
            res = 0;
            if ((((y & 0x7800000000000000) != 0x7800000000000000)) || (((y & 0x8000000000000000) == 0x8000000000000000))) {
                res = 1;
            }
            return (res, pfpsf);
        }
    } else if ((y & 0x7800000000000000) == 0x7800000000000000) {
        res = 0;
        if ((y & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((x & 0x6000000000000000) == 0x6000000000000000) {
        exp_x = ((go_checked_shr_u64((x & 0x1ff8000000000000), go_shift_count_u64((51) as u64))) as i64);
        sig_x = ((x & 0x7ffffffffffff) | 0x20000000000000);
        if (sig_x > 9999999999999999) {
            non_canon_x = 1;
        } else {
            non_canon_x = 0;
        }
    } else {
        exp_x = ((go_checked_shr_u64((x & 0x7fe0000000000000), go_shift_count_u64((53) as u64))) as i64);
        sig_x = (x & 0x1fffffffffffff);
        non_canon_x = 0;
    }
    if ((y & 0x6000000000000000) == 0x6000000000000000) {
        exp_y = ((go_checked_shr_u64((y & 0x1ff8000000000000), go_shift_count_u64((51) as u64))) as i64);
        sig_y = ((y & 0x7ffffffffffff) | 0x20000000000000);
        if (sig_y > 9999999999999999) {
            non_canon_y = 1;
        } else {
            non_canon_y = 0;
        }
    } else {
        exp_y = ((go_checked_shr_u64((y & 0x7fe0000000000000), go_shift_count_u64((53) as u64))) as i64);
        sig_y = (y & 0x1fffffffffffff);
        non_canon_y = 0;
    }
    if ((non_canon_x != 0) || (sig_x == 0)) {
        x_is_zero = 1;
    }
    if ((non_canon_y != 0) || (sig_y == 0)) {
        y_is_zero = 1;
    }
    if ((x_is_zero != 0) && (y_is_zero != 0)) {
        res = 0;
        return (res, pfpsf);
    } else if (x_is_zero != 0) {
        res = 0;
        if ((y & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    } else if (y_is_zero != 0) {
        res = 0;
        if ((x & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((((x ^ y) & 0x8000000000000000)) == 0x8000000000000000) {
        res = 0;
        if ((y & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((sig_x > sig_y) && (exp_x >= exp_y)) {
        res = 0;
        if ((x & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((sig_x < sig_y) && (exp_x <= exp_y)) {
        res = 0;
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((exp_x.wrapping_sub(exp_y)) > 15) {
        res = 0;
        if ((x & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((exp_y.wrapping_sub(exp_x)) > 15) {
        res = 0;
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (exp_x > exp_y) {
        sig_n_prime = __mul_64x64_to_128(sig_x, bid_mult_factor[(exp_x.wrapping_sub(exp_y)) as usize]);
        if ((sig_n_prime.w[1] == 0) && (sig_n_prime.w[0] == sig_y)) {
            res = 0;
            return (res, pfpsf);
        }
        res = 0;
        if ((((sig_n_prime.w[1] > 0) || (sig_n_prime.w[0] > sig_y))) != (((x & 0x8000000000000000) == 0x8000000000000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    sig_n_prime = __mul_64x64_to_128(sig_y, bid_mult_factor[(exp_y.wrapping_sub(exp_x)) as usize]);
    if ((sig_n_prime.w[1] == 0) && (sig_n_prime.w[0] == sig_x)) {
        res = 0;
        return (res, pfpsf);
    }
    res = 0;
    if ((((sig_n_prime.w[1] == 0) && (sig_x > sig_n_prime.w[0]))) != (((x & 0x8000000000000000) == 0x8000000000000000))) {
        res = 1;
    }
    return (res, pfpsf);
}

pub fn bid64_quiet_less(mut x: u64, mut y: u64) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut sig_x: u64 = 0;
    let mut sig_y: u64 = 0;
    let mut sig_n_prime: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut pfpsf: u32 = 0;
    let mut x_is_zero: i64 = 0;
    let mut y_is_zero: i64 = 0;
    let mut non_canon_x: i64 = 0;
    let mut non_canon_y: i64 = 0;
    if ((((x & 0x7c00000000000000) == 0x7c00000000000000)) || (((y & 0x7c00000000000000) == 0x7c00000000000000))) {
        if (((x & 0x7e00000000000000) == 0x7e00000000000000) || ((y & 0x7e00000000000000) == 0x7e00000000000000)) {
            pfpsf |= 1;
        }
        res = 0;
        return (res, pfpsf);
    }
    if (x == y) {
        res = 0;
        return (res, pfpsf);
    }
    if ((x & 0x7800000000000000) == 0x7800000000000000) {
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = 0;
            if ((((y & 0x7800000000000000) != 0x7800000000000000)) || ((y & 0x8000000000000000) != 0x8000000000000000)) {
                res = 1;
            }
            return (res, pfpsf);
        } else {
            res = 0;
            return (res, pfpsf);
        }
    } else if ((y & 0x7800000000000000) == 0x7800000000000000) {
        res = 0;
        if ((y & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((x & 0x6000000000000000) == 0x6000000000000000) {
        exp_x = ((go_checked_shr_u64((x & 0x1ff8000000000000), go_shift_count_u64((51) as u64))) as i64);
        sig_x = ((x & 0x7ffffffffffff) | 0x20000000000000);
        if (sig_x > 9999999999999999) {
            non_canon_x = 1;
        } else {
            non_canon_x = 0;
        }
    } else {
        exp_x = ((go_checked_shr_u64((x & 0x7fe0000000000000), go_shift_count_u64((53) as u64))) as i64);
        sig_x = (x & 0x1fffffffffffff);
        non_canon_x = 0;
    }
    if ((y & 0x6000000000000000) == 0x6000000000000000) {
        exp_y = ((go_checked_shr_u64((y & 0x1ff8000000000000), go_shift_count_u64((51) as u64))) as i64);
        sig_y = ((y & 0x7ffffffffffff) | 0x20000000000000);
        if (sig_y > 9999999999999999) {
            non_canon_y = 1;
        } else {
            non_canon_y = 0;
        }
    } else {
        exp_y = ((go_checked_shr_u64((y & 0x7fe0000000000000), go_shift_count_u64((53) as u64))) as i64);
        sig_y = (y & 0x1fffffffffffff);
        non_canon_y = 0;
    }
    if ((non_canon_x != 0) || (sig_x == 0)) {
        x_is_zero = 1;
    }
    if ((non_canon_y != 0) || (sig_y == 0)) {
        y_is_zero = 1;
    }
    if ((x_is_zero != 0) && (y_is_zero != 0)) {
        res = 0;
        return (res, pfpsf);
    } else if (x_is_zero != 0) {
        res = 0;
        if ((y & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    } else if (y_is_zero != 0) {
        res = 0;
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((((x ^ y) & 0x8000000000000000)) == 0x8000000000000000) {
        res = 0;
        if ((y & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((sig_x > sig_y) && (exp_x >= exp_y)) {
        res = 0;
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((sig_x < sig_y) && (exp_x <= exp_y)) {
        res = 0;
        if ((x & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((exp_x.wrapping_sub(exp_y)) > 15) {
        res = 0;
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((exp_y.wrapping_sub(exp_x)) > 15) {
        res = 0;
        if ((x & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (exp_x > exp_y) {
        sig_n_prime = __mul_64x64_to_128(sig_x, bid_mult_factor[(exp_x.wrapping_sub(exp_y)) as usize]);
        if ((sig_n_prime.w[1] == 0) && (sig_n_prime.w[0] == sig_y)) {
            res = 0;
            return (res, pfpsf);
        }
        res = 0;
        if ((((sig_n_prime.w[1] == 0) && (sig_n_prime.w[0] < sig_y))) != (((x & 0x8000000000000000) == 0x8000000000000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    sig_n_prime = __mul_64x64_to_128(sig_y, bid_mult_factor[(exp_y.wrapping_sub(exp_x)) as usize]);
    if ((sig_n_prime.w[1] == 0) && (sig_n_prime.w[0] == sig_x)) {
        res = 0;
        return (res, pfpsf);
    }
    res = 0;
    if ((((sig_n_prime.w[1] > 0) || (sig_x < sig_n_prime.w[0]))) != (((x & 0x8000000000000000) == 0x8000000000000000))) {
        res = 1;
    }
    return (res, pfpsf);
}

pub fn bid64_quiet_less_equal(mut x: u64, mut y: u64) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut sig_x: u64 = 0;
    let mut sig_y: u64 = 0;
    let mut sig_n_prime: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut pfpsf: u32 = 0;
    let mut x_is_zero: i64 = 0;
    let mut y_is_zero: i64 = 0;
    let mut non_canon_x: i64 = 0;
    let mut non_canon_y: i64 = 0;
    if ((((x & 0x7c00000000000000) == 0x7c00000000000000)) || (((y & 0x7c00000000000000) == 0x7c00000000000000))) {
        if (((x & 0x7e00000000000000) == 0x7e00000000000000) || ((y & 0x7e00000000000000) == 0x7e00000000000000)) {
            pfpsf |= 1;
        }
        res = 0;
        return (res, pfpsf);
    }
    if (x == y) {
        res = 1;
        return (res, pfpsf);
    }
    if ((x & 0x7800000000000000) == 0x7800000000000000) {
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
            return (res, pfpsf);
        } else {
            res = 1;
            if ((((y & 0x7800000000000000) != 0x7800000000000000)) || (((y & 0x8000000000000000) == 0x8000000000000000))) {
                res = 0;
            }
            return (res, pfpsf);
        }
    } else if ((y & 0x7800000000000000) == 0x7800000000000000) {
        res = 0;
        if ((y & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((x & 0x6000000000000000) == 0x6000000000000000) {
        exp_x = ((go_checked_shr_u64((x & 0x1ff8000000000000), go_shift_count_u64((51) as u64))) as i64);
        sig_x = ((x & 0x7ffffffffffff) | 0x20000000000000);
        if (sig_x > 9999999999999999) {
            non_canon_x = 1;
        } else {
            non_canon_x = 0;
        }
    } else {
        exp_x = ((go_checked_shr_u64((x & 0x7fe0000000000000), go_shift_count_u64((53) as u64))) as i64);
        sig_x = (x & 0x1fffffffffffff);
        non_canon_x = 0;
    }
    if ((y & 0x6000000000000000) == 0x6000000000000000) {
        exp_y = ((go_checked_shr_u64((y & 0x1ff8000000000000), go_shift_count_u64((51) as u64))) as i64);
        sig_y = ((y & 0x7ffffffffffff) | 0x20000000000000);
        if (sig_y > 9999999999999999) {
            non_canon_y = 1;
        } else {
            non_canon_y = 0;
        }
    } else {
        exp_y = ((go_checked_shr_u64((y & 0x7fe0000000000000), go_shift_count_u64((53) as u64))) as i64);
        sig_y = (y & 0x1fffffffffffff);
        non_canon_y = 0;
    }
    if ((non_canon_x != 0) || (sig_x == 0)) {
        x_is_zero = 1;
    }
    if ((non_canon_y != 0) || (sig_y == 0)) {
        y_is_zero = 1;
    }
    if ((x_is_zero != 0) && (y_is_zero != 0)) {
        res = 1;
        return (res, pfpsf);
    } else if (x_is_zero != 0) {
        res = 0;
        if ((y & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    } else if (y_is_zero != 0) {
        res = 0;
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((((x ^ y) & 0x8000000000000000)) == 0x8000000000000000) {
        res = 0;
        if ((y & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((sig_x > sig_y) && (exp_x >= exp_y)) {
        res = 0;
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((sig_x < sig_y) && (exp_x <= exp_y)) {
        res = 0;
        if ((x & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((exp_x.wrapping_sub(exp_y)) > 15) {
        res = 0;
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((exp_y.wrapping_sub(exp_x)) > 15) {
        res = 0;
        if ((x & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (exp_x > exp_y) {
        sig_n_prime = __mul_64x64_to_128(sig_x, bid_mult_factor[(exp_x.wrapping_sub(exp_y)) as usize]);
        if ((sig_n_prime.w[1] == 0) && (sig_n_prime.w[0] == sig_y)) {
            res = 1;
            return (res, pfpsf);
        }
        res = 0;
        if ((((sig_n_prime.w[1] == 0) && (sig_n_prime.w[0] < sig_y))) != (((x & 0x8000000000000000) == 0x8000000000000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    sig_n_prime = __mul_64x64_to_128(sig_y, bid_mult_factor[(exp_y.wrapping_sub(exp_x)) as usize]);
    if ((sig_n_prime.w[1] == 0) && (sig_n_prime.w[0] == sig_x)) {
        res = 1;
        return (res, pfpsf);
    }
    res = 0;
    if ((((sig_n_prime.w[1] > 0) || (sig_x < sig_n_prime.w[0]))) != (((x & 0x8000000000000000) == 0x8000000000000000))) {
        res = 1;
    }
    return (res, pfpsf);
}

pub fn bid64_quiet_less_unordered(mut x: u64, mut y: u64) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut sig_x: u64 = 0;
    let mut sig_y: u64 = 0;
    let mut sig_n_prime: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut pfpsf: u32 = 0;
    let mut x_is_zero: i64 = 0;
    let mut y_is_zero: i64 = 0;
    let mut non_canon_x: i64 = 0;
    let mut non_canon_y: i64 = 0;
    if ((((x & 0x7c00000000000000) == 0x7c00000000000000)) || (((y & 0x7c00000000000000) == 0x7c00000000000000))) {
        if (((x & 0x7e00000000000000) == 0x7e00000000000000) || ((y & 0x7e00000000000000) == 0x7e00000000000000)) {
            pfpsf |= 1;
        }
        res = 1;
        return (res, pfpsf);
    }
    if (x == y) {
        res = 0;
        return (res, pfpsf);
    }
    if ((x & 0x7800000000000000) == 0x7800000000000000) {
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = 0;
            if ((((y & 0x7800000000000000) != 0x7800000000000000)) || ((y & 0x8000000000000000) != 0x8000000000000000)) {
                res = 1;
            }
            return (res, pfpsf);
        } else {
            res = 0;
            return (res, pfpsf);
        }
    } else if ((y & 0x7800000000000000) == 0x7800000000000000) {
        res = 0;
        if ((y & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((x & 0x6000000000000000) == 0x6000000000000000) {
        exp_x = ((go_checked_shr_u64((x & 0x1ff8000000000000), go_shift_count_u64((51) as u64))) as i64);
        sig_x = ((x & 0x7ffffffffffff) | 0x20000000000000);
        if (sig_x > 9999999999999999) {
            non_canon_x = 1;
        } else {
            non_canon_x = 0;
        }
    } else {
        exp_x = ((go_checked_shr_u64((x & 0x7fe0000000000000), go_shift_count_u64((53) as u64))) as i64);
        sig_x = (x & 0x1fffffffffffff);
        non_canon_x = 0;
    }
    if ((y & 0x6000000000000000) == 0x6000000000000000) {
        exp_y = ((go_checked_shr_u64((y & 0x1ff8000000000000), go_shift_count_u64((51) as u64))) as i64);
        sig_y = ((y & 0x7ffffffffffff) | 0x20000000000000);
        if (sig_y > 9999999999999999) {
            non_canon_y = 1;
        } else {
            non_canon_y = 0;
        }
    } else {
        exp_y = ((go_checked_shr_u64((y & 0x7fe0000000000000), go_shift_count_u64((53) as u64))) as i64);
        sig_y = (y & 0x1fffffffffffff);
        non_canon_y = 0;
    }
    if ((non_canon_x != 0) || (sig_x == 0)) {
        x_is_zero = 1;
    }
    if ((non_canon_y != 0) || (sig_y == 0)) {
        y_is_zero = 1;
    }
    if ((x_is_zero != 0) && (y_is_zero != 0)) {
        res = 0;
        return (res, pfpsf);
    } else if (x_is_zero != 0) {
        res = 0;
        if ((y & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    } else if (y_is_zero != 0) {
        res = 0;
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((((x ^ y) & 0x8000000000000000)) == 0x8000000000000000) {
        res = 0;
        if ((y & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((sig_x > sig_y) && (exp_x >= exp_y)) {
        res = 0;
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((sig_x < sig_y) && (exp_x <= exp_y)) {
        res = 0;
        if ((x & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((exp_x.wrapping_sub(exp_y)) > 15) {
        res = 0;
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((exp_y.wrapping_sub(exp_x)) > 15) {
        res = 0;
        if ((x & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (exp_x > exp_y) {
        sig_n_prime = __mul_64x64_to_128(sig_x, bid_mult_factor[(exp_x.wrapping_sub(exp_y)) as usize]);
        if ((sig_n_prime.w[1] == 0) && (sig_n_prime.w[0] == sig_y)) {
            res = 0;
            return (res, pfpsf);
        }
        res = 0;
        if ((((sig_n_prime.w[1] == 0) && (sig_n_prime.w[0] < sig_y))) != (((x & 0x8000000000000000) == 0x8000000000000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    sig_n_prime = __mul_64x64_to_128(sig_y, bid_mult_factor[(exp_y.wrapping_sub(exp_x)) as usize]);
    if ((sig_n_prime.w[1] == 0) && (sig_n_prime.w[0] == sig_x)) {
        res = 0;
        return (res, pfpsf);
    }
    res = 0;
    if ((((sig_n_prime.w[1] > 0) || (sig_x < sig_n_prime.w[0]))) != (((x & 0x8000000000000000) == 0x8000000000000000))) {
        res = 1;
    }
    return (res, pfpsf);
}

pub fn bid64_quiet_not_equal(mut x: u64, mut y: u64) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut exp_t: i64 = 0;
    let mut sig_x: u64 = 0;
    let mut sig_y: u64 = 0;
    let mut sig_t: u64 = 0;
    let mut pfpsf: u32 = 0;
    let mut x_is_zero: i64 = 0;
    let mut y_is_zero: i64 = 0;
    let mut non_canon_x: i64 = 0;
    let mut non_canon_y: i64 = 0;
    let mut lcv: i64 = 0;
    if ((((x & 0x7c00000000000000) == 0x7c00000000000000)) || (((y & 0x7c00000000000000) == 0x7c00000000000000))) {
        if (((x & 0x7e00000000000000) == 0x7e00000000000000) || ((y & 0x7e00000000000000) == 0x7e00000000000000)) {
            pfpsf |= 1;
        }
        res = 1;
        return (res, pfpsf);
    }
    if (x == y) {
        res = 0;
        return (res, pfpsf);
    }
    if ((((x & 0x7800000000000000) == 0x7800000000000000)) && (((y & 0x7800000000000000) == 0x7800000000000000))) {
        res = 0;
        if ((((x ^ y) & 0x8000000000000000)) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((((x & 0x7800000000000000) == 0x7800000000000000)) || (((y & 0x7800000000000000) == 0x7800000000000000))) {
        res = 1;
        return (res, pfpsf);
    }
    if ((x & 0x6000000000000000) == 0x6000000000000000) {
        exp_x = ((go_checked_shr_u64((x & 0x1ff8000000000000), go_shift_count_u64((51) as u64))) as i64);
        sig_x = ((x & 0x7ffffffffffff) | 0x20000000000000);
        if (sig_x > 9999999999999999) {
            non_canon_x = 1;
        } else {
            non_canon_x = 0;
        }
    } else {
        exp_x = ((go_checked_shr_u64((x & 0x7fe0000000000000), go_shift_count_u64((53) as u64))) as i64);
        sig_x = (x & 0x1fffffffffffff);
        non_canon_x = 0;
    }
    if ((y & 0x6000000000000000) == 0x6000000000000000) {
        exp_y = ((go_checked_shr_u64((y & 0x1ff8000000000000), go_shift_count_u64((51) as u64))) as i64);
        sig_y = ((y & 0x7ffffffffffff) | 0x20000000000000);
        if (sig_y > 9999999999999999) {
            non_canon_y = 1;
        } else {
            non_canon_y = 0;
        }
    } else {
        exp_y = ((go_checked_shr_u64((y & 0x7fe0000000000000), go_shift_count_u64((53) as u64))) as i64);
        sig_y = (y & 0x1fffffffffffff);
        non_canon_y = 0;
    }
    if ((non_canon_x != 0) || (sig_x == 0)) {
        x_is_zero = 1;
    }
    if ((non_canon_y != 0) || (sig_y == 0)) {
        y_is_zero = 1;
    }
    if ((x_is_zero != 0) && (y_is_zero != 0)) {
        res = 0;
        return (res, pfpsf);
    } else if ((((x_is_zero != 0) && (y_is_zero == 0))) || (((x_is_zero == 0) && (y_is_zero != 0)))) {
        res = 1;
        return (res, pfpsf);
    }
    if ((((x ^ y) & 0x8000000000000000)) != 0) {
        res = 1;
        return (res, pfpsf);
    }
    if (exp_x > exp_y) {
        exp_t = exp_x;
        exp_x = exp_y;
        exp_y = exp_t;
        sig_t = sig_x;
        sig_x = sig_y;
        sig_y = sig_t;
    }
    if ((exp_y.wrapping_sub(exp_x)) > 15) {
        res = 1;
        return (res, pfpsf);
    }
    lcv = 0;
    while (lcv < ((exp_y.wrapping_sub(exp_x)))) {
        sig_y = (sig_y.wrapping_mul(10));
        if (sig_y > 9999999999999999) {
            res = 1;
            return (res, pfpsf);
        }
        lcv = lcv.wrapping_add(1);
    }
    res = 0;
    if (sig_y != sig_x) {
        res = 1;
    }
    return (res, pfpsf);
}

pub fn bid64_quiet_not_greater(mut x: u64, mut y: u64) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut sig_x: u64 = 0;
    let mut sig_y: u64 = 0;
    let mut sig_n_prime: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut pfpsf: u32 = 0;
    let mut x_is_zero: i64 = 0;
    let mut y_is_zero: i64 = 0;
    let mut non_canon_x: i64 = 0;
    let mut non_canon_y: i64 = 0;
    if ((((x & 0x7c00000000000000) == 0x7c00000000000000)) || (((y & 0x7c00000000000000) == 0x7c00000000000000))) {
        if (((x & 0x7e00000000000000) == 0x7e00000000000000) || ((y & 0x7e00000000000000) == 0x7e00000000000000)) {
            pfpsf |= 1;
        }
        res = 1;
        return (res, pfpsf);
    }
    if (x == y) {
        res = 1;
        return (res, pfpsf);
    }
    if ((x & 0x7800000000000000) == 0x7800000000000000) {
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
            return (res, pfpsf);
        }
        res = 1;
        if ((((y & 0x7800000000000000) != 0x7800000000000000)) || (((y & 0x8000000000000000) == 0x8000000000000000))) {
            res = 0;
        }
        return (res, pfpsf);
    } else if ((y & 0x7800000000000000) == 0x7800000000000000) {
        res = 0;
        if ((y & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((x & 0x6000000000000000) == 0x6000000000000000) {
        exp_x = ((go_checked_shr_u64((x & 0x1ff8000000000000), go_shift_count_u64((51) as u64))) as i64);
        sig_x = ((x & 0x7ffffffffffff) | 0x20000000000000);
        if (sig_x > 9999999999999999) {
            non_canon_x = 1;
        } else {
            non_canon_x = 0;
        }
    } else {
        exp_x = ((go_checked_shr_u64((x & 0x7fe0000000000000), go_shift_count_u64((53) as u64))) as i64);
        sig_x = (x & 0x1fffffffffffff);
        non_canon_x = 0;
    }
    if ((y & 0x6000000000000000) == 0x6000000000000000) {
        exp_y = ((go_checked_shr_u64((y & 0x1ff8000000000000), go_shift_count_u64((51) as u64))) as i64);
        sig_y = ((y & 0x7ffffffffffff) | 0x20000000000000);
        if (sig_y > 9999999999999999) {
            non_canon_y = 1;
        } else {
            non_canon_y = 0;
        }
    } else {
        exp_y = ((go_checked_shr_u64((y & 0x7fe0000000000000), go_shift_count_u64((53) as u64))) as i64);
        sig_y = (y & 0x1fffffffffffff);
        non_canon_y = 0;
    }
    if ((non_canon_x != 0) || (sig_x == 0)) {
        x_is_zero = 1;
    }
    if ((non_canon_y != 0) || (sig_y == 0)) {
        y_is_zero = 1;
    }
    if ((x_is_zero != 0) && (y_is_zero != 0)) {
        res = 1;
        return (res, pfpsf);
    }
    if (x_is_zero != 0) {
        res = 0;
        if ((y & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (y_is_zero != 0) {
        res = 0;
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((((x ^ y) & 0x8000000000000000)) == 0x8000000000000000) {
        res = 0;
        if ((y & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((sig_x > sig_y) && (exp_x >= exp_y)) {
        res = 0;
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((sig_x < sig_y) && (exp_x <= exp_y)) {
        res = 0;
        if ((x & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((exp_x.wrapping_sub(exp_y)) > 15) {
        res = 0;
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((exp_y.wrapping_sub(exp_x)) > 15) {
        res = 0;
        if ((x & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (exp_x > exp_y) {
        sig_n_prime = __mul_64x64_to_128(sig_x, bid_mult_factor[(exp_x.wrapping_sub(exp_y)) as usize]);
        if ((sig_n_prime.w[1] == 0) && (sig_n_prime.w[0] == sig_y)) {
            res = 1;
            return (res, pfpsf);
        }
        res = 0;
        if ((((sig_n_prime.w[1] == 0) && (sig_n_prime.w[0] < sig_y))) != (((x & 0x8000000000000000) == 0x8000000000000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    sig_n_prime = __mul_64x64_to_128(sig_y, bid_mult_factor[(exp_y.wrapping_sub(exp_x)) as usize]);
    if ((sig_n_prime.w[1] == 0) && (sig_n_prime.w[0] == sig_x)) {
        res = 1;
        return (res, pfpsf);
    }
    res = 0;
    if ((((sig_n_prime.w[1] > 0) || (sig_x < sig_n_prime.w[0]))) != (((x & 0x8000000000000000) == 0x8000000000000000))) {
        res = 1;
    }
    return (res, pfpsf);
}

pub fn bid64_quiet_not_less(mut x: u64, mut y: u64) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut sig_x: u64 = 0;
    let mut sig_y: u64 = 0;
    let mut sig_n_prime: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut pfpsf: u32 = 0;
    let mut x_is_zero: i64 = 0;
    let mut y_is_zero: i64 = 0;
    let mut non_canon_x: i64 = 0;
    let mut non_canon_y: i64 = 0;
    if ((((x & 0x7c00000000000000) == 0x7c00000000000000)) || (((y & 0x7c00000000000000) == 0x7c00000000000000))) {
        if (((x & 0x7e00000000000000) == 0x7e00000000000000) || ((y & 0x7e00000000000000) == 0x7e00000000000000)) {
            pfpsf |= 1;
        }
        res = 1;
        return (res, pfpsf);
    }
    if (x == y) {
        res = 1;
        return (res, pfpsf);
    }
    if ((x & 0x7800000000000000) == 0x7800000000000000) {
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = 0;
            if ((((y & 0x7800000000000000) == 0x7800000000000000)) && ((y & 0x8000000000000000) == 0x8000000000000000)) {
                res = 1;
            }
            return (res, pfpsf);
        }
        res = 1;
        return (res, pfpsf);
    } else if ((y & 0x7800000000000000) == 0x7800000000000000) {
        res = 0;
        if ((y & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((x & 0x6000000000000000) == 0x6000000000000000) {
        exp_x = ((go_checked_shr_u64((x & 0x1ff8000000000000), go_shift_count_u64((51) as u64))) as i64);
        sig_x = ((x & 0x7ffffffffffff) | 0x20000000000000);
        if (sig_x > 9999999999999999) {
            non_canon_x = 1;
        } else {
            non_canon_x = 0;
        }
    } else {
        exp_x = ((go_checked_shr_u64((x & 0x7fe0000000000000), go_shift_count_u64((53) as u64))) as i64);
        sig_x = (x & 0x1fffffffffffff);
        non_canon_x = 0;
    }
    if ((y & 0x6000000000000000) == 0x6000000000000000) {
        exp_y = ((go_checked_shr_u64((y & 0x1ff8000000000000), go_shift_count_u64((51) as u64))) as i64);
        sig_y = ((y & 0x7ffffffffffff) | 0x20000000000000);
        if (sig_y > 9999999999999999) {
            non_canon_y = 1;
        } else {
            non_canon_y = 0;
        }
    } else {
        exp_y = ((go_checked_shr_u64((y & 0x7fe0000000000000), go_shift_count_u64((53) as u64))) as i64);
        sig_y = (y & 0x1fffffffffffff);
        non_canon_y = 0;
    }
    if ((non_canon_x != 0) || (sig_x == 0)) {
        x_is_zero = 1;
    }
    if ((non_canon_y != 0) || (sig_y == 0)) {
        y_is_zero = 1;
    }
    if ((x_is_zero != 0) && (y_is_zero != 0)) {
        res = 1;
        return (res, pfpsf);
    }
    if (x_is_zero != 0) {
        res = 0;
        if ((y & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (y_is_zero != 0) {
        res = 0;
        if ((x & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((((x ^ y) & 0x8000000000000000)) == 0x8000000000000000) {
        res = 0;
        if ((y & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((sig_x > sig_y) && (exp_x >= exp_y)) {
        res = 0;
        if ((x & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((sig_x < sig_y) && (exp_x <= exp_y)) {
        res = 0;
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((exp_x.wrapping_sub(exp_y)) > 15) {
        res = 0;
        if ((x & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((exp_y.wrapping_sub(exp_x)) > 15) {
        res = 0;
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (exp_x > exp_y) {
        sig_n_prime = __mul_64x64_to_128(sig_x, bid_mult_factor[(exp_x.wrapping_sub(exp_y)) as usize]);
        if ((sig_n_prime.w[1] == 0) && (sig_n_prime.w[0] == sig_y)) {
            res = 1;
            return (res, pfpsf);
        }
        res = 0;
        if ((((sig_n_prime.w[1] == 0) && (sig_n_prime.w[0] < sig_y))) != (((x & 0x8000000000000000) != 0x8000000000000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    sig_n_prime = __mul_64x64_to_128(sig_y, bid_mult_factor[(exp_y.wrapping_sub(exp_x)) as usize]);
    if ((sig_n_prime.w[1] == 0) && (sig_n_prime.w[0] == sig_x)) {
        res = 1;
        return (res, pfpsf);
    }
    res = 0;
    if ((((sig_n_prime.w[1] > 0) || (sig_x < sig_n_prime.w[0]))) != (((x & 0x8000000000000000) != 0x8000000000000000))) {
        res = 1;
    }
    return (res, pfpsf);
}

pub fn bid64_quiet_ordered(mut x: u64, mut y: u64) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut pfpsf: u32 = 0;
    if ((((x & 0x7c00000000000000) == 0x7c00000000000000)) || (((y & 0x7c00000000000000) == 0x7c00000000000000))) {
        if (((x & 0x7e00000000000000) == 0x7e00000000000000) || ((y & 0x7e00000000000000) == 0x7e00000000000000)) {
            pfpsf |= 1;
        }
        res = 0;
        return (res, pfpsf);
    }
    res = 1;
    return (res, pfpsf);
}

pub fn bid64_quiet_unordered(mut x: u64, mut y: u64) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut pfpsf: u32 = 0;
    if ((((x & 0x7c00000000000000) == 0x7c00000000000000)) || (((y & 0x7c00000000000000) == 0x7c00000000000000))) {
        if (((x & 0x7e00000000000000) == 0x7e00000000000000) || ((y & 0x7e00000000000000) == 0x7e00000000000000)) {
            pfpsf |= 1;
        }
        res = 1;
        return (res, pfpsf);
    }
    res = 0;
    return (res, pfpsf);
}

pub fn bid64_signaling_greater(mut x: u64, mut y: u64) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut sig_x: u64 = 0;
    let mut sig_y: u64 = 0;
    let mut sig_n_prime: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut pfpsf: u32 = 0;
    let mut x_is_zero: i64 = 0;
    let mut y_is_zero: i64 = 0;
    let mut non_canon_x: i64 = 0;
    let mut non_canon_y: i64 = 0;
    if ((((x & 0x7c00000000000000) == 0x7c00000000000000)) || (((y & 0x7c00000000000000) == 0x7c00000000000000))) {
        pfpsf |= 1;
        res = 0;
        return (res, pfpsf);
    }
    if (x == y) {
        res = 0;
        return (res, pfpsf);
    }
    if ((x & 0x7800000000000000) == 0x7800000000000000) {
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = 0;
            return (res, pfpsf);
        }
        res = 0;
        if ((((y & 0x7800000000000000) != 0x7800000000000000)) || (((y & 0x8000000000000000) == 0x8000000000000000))) {
            res = 1;
        }
        return (res, pfpsf);
    } else if ((y & 0x7800000000000000) == 0x7800000000000000) {
        res = 0;
        if ((y & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((x & 0x6000000000000000) == 0x6000000000000000) {
        exp_x = ((go_checked_shr_u64((x & 0x1ff8000000000000), go_shift_count_u64((51) as u64))) as i64);
        sig_x = ((x & 0x7ffffffffffff) | 0x20000000000000);
        if (sig_x > 9999999999999999) {
            non_canon_x = 1;
        } else {
            non_canon_x = 0;
        }
    } else {
        exp_x = ((go_checked_shr_u64((x & 0x7fe0000000000000), go_shift_count_u64((53) as u64))) as i64);
        sig_x = (x & 0x1fffffffffffff);
        non_canon_x = 0;
    }
    if ((y & 0x6000000000000000) == 0x6000000000000000) {
        exp_y = ((go_checked_shr_u64((y & 0x1ff8000000000000), go_shift_count_u64((51) as u64))) as i64);
        sig_y = ((y & 0x7ffffffffffff) | 0x20000000000000);
        if (sig_y > 9999999999999999) {
            non_canon_y = 1;
        } else {
            non_canon_y = 0;
        }
    } else {
        exp_y = ((go_checked_shr_u64((y & 0x7fe0000000000000), go_shift_count_u64((53) as u64))) as i64);
        sig_y = (y & 0x1fffffffffffff);
        non_canon_y = 0;
    }
    if ((non_canon_x != 0) || (sig_x == 0)) {
        x_is_zero = 1;
    }
    if ((non_canon_y != 0) || (sig_y == 0)) {
        y_is_zero = 1;
    }
    if ((x_is_zero != 0) && (y_is_zero != 0)) {
        res = 0;
        return (res, pfpsf);
    }
    if (x_is_zero != 0) {
        res = 0;
        if ((y & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (y_is_zero != 0) {
        res = 0;
        if ((x & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((((x ^ y) & 0x8000000000000000)) == 0x8000000000000000) {
        res = 0;
        if ((y & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((sig_x > sig_y) && (exp_x >= exp_y)) {
        res = 0;
        if ((x & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((sig_x < sig_y) && (exp_x <= exp_y)) {
        res = 0;
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((exp_x.wrapping_sub(exp_y)) > 15) {
        res = 0;
        if ((x & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((exp_y.wrapping_sub(exp_x)) > 15) {
        res = 0;
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (exp_x > exp_y) {
        sig_n_prime = __mul_64x64_to_128(sig_x, bid_mult_factor[(exp_x.wrapping_sub(exp_y)) as usize]);
        if ((sig_n_prime.w[1] == 0) && (sig_n_prime.w[0] == sig_y)) {
            res = 0;
            return (res, pfpsf);
        }
        res = 0;
        if ((((sig_n_prime.w[1] > 0) || (sig_n_prime.w[0] > sig_y))) != (((x & 0x8000000000000000) == 0x8000000000000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    sig_n_prime = __mul_64x64_to_128(sig_y, bid_mult_factor[(exp_y.wrapping_sub(exp_x)) as usize]);
    if ((sig_n_prime.w[1] == 0) && (sig_n_prime.w[0] == sig_x)) {
        res = 0;
        return (res, pfpsf);
    }
    res = 0;
    if ((((sig_n_prime.w[1] == 0) && (sig_x > sig_n_prime.w[0]))) != (((x & 0x8000000000000000) == 0x8000000000000000))) {
        res = 1;
    }
    return (res, pfpsf);
}

pub fn bid64_signaling_greater_equal(mut x: u64, mut y: u64) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut sig_x: u64 = 0;
    let mut sig_y: u64 = 0;
    let mut sig_n_prime: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut pfpsf: u32 = 0;
    let mut x_is_zero: i64 = 0;
    let mut y_is_zero: i64 = 0;
    let mut non_canon_x: i64 = 0;
    let mut non_canon_y: i64 = 0;
    if ((((x & 0x7c00000000000000) == 0x7c00000000000000)) || (((y & 0x7c00000000000000) == 0x7c00000000000000))) {
        pfpsf |= 1;
        res = 0;
        return (res, pfpsf);
    }
    if (x == y) {
        res = 1;
        return (res, pfpsf);
    }
    if ((x & 0x7800000000000000) == 0x7800000000000000) {
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = 0;
            if ((((y & 0x7800000000000000) == 0x7800000000000000)) && ((y & 0x8000000000000000) == 0x8000000000000000)) {
                res = 1;
            }
            return (res, pfpsf);
        }
        res = 1;
        return (res, pfpsf);
    } else if ((y & 0x7800000000000000) == 0x7800000000000000) {
        res = 0;
        if ((y & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((x & 0x6000000000000000) == 0x6000000000000000) {
        exp_x = ((go_checked_shr_u64((x & 0x1ff8000000000000), go_shift_count_u64((51) as u64))) as i64);
        sig_x = ((x & 0x7ffffffffffff) | 0x20000000000000);
        if (sig_x > 9999999999999999) {
            non_canon_x = 1;
        } else {
            non_canon_x = 0;
        }
    } else {
        exp_x = ((go_checked_shr_u64((x & 0x7fe0000000000000), go_shift_count_u64((53) as u64))) as i64);
        sig_x = (x & 0x1fffffffffffff);
        non_canon_x = 0;
    }
    if ((y & 0x6000000000000000) == 0x6000000000000000) {
        exp_y = ((go_checked_shr_u64((y & 0x1ff8000000000000), go_shift_count_u64((51) as u64))) as i64);
        sig_y = ((y & 0x7ffffffffffff) | 0x20000000000000);
        if (sig_y > 9999999999999999) {
            non_canon_y = 1;
        } else {
            non_canon_y = 0;
        }
    } else {
        exp_y = ((go_checked_shr_u64((y & 0x7fe0000000000000), go_shift_count_u64((53) as u64))) as i64);
        sig_y = (y & 0x1fffffffffffff);
        non_canon_y = 0;
    }
    if ((non_canon_x != 0) || (sig_x == 0)) {
        x_is_zero = 1;
    }
    if ((non_canon_y != 0) || (sig_y == 0)) {
        y_is_zero = 1;
    }
    if ((x_is_zero != 0) && (y_is_zero != 0)) {
        res = 1;
        return (res, pfpsf);
    }
    if (x_is_zero != 0) {
        res = 0;
        if ((y & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (y_is_zero != 0) {
        res = 0;
        if ((x & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((((x ^ y) & 0x8000000000000000)) == 0x8000000000000000) {
        res = 0;
        if ((y & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((sig_x > sig_y) && (exp_x >= exp_y)) {
        res = 0;
        if ((x & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((sig_x < sig_y) && (exp_x <= exp_y)) {
        res = 0;
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((exp_x.wrapping_sub(exp_y)) > 15) {
        res = 0;
        if ((x & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((exp_y.wrapping_sub(exp_x)) > 15) {
        res = 0;
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (exp_x > exp_y) {
        sig_n_prime = __mul_64x64_to_128(sig_x, bid_mult_factor[(exp_x.wrapping_sub(exp_y)) as usize]);
        if ((sig_n_prime.w[1] == 0) && (sig_n_prime.w[0] == sig_y)) {
            res = 1;
            return (res, pfpsf);
        }
        res = 0;
        if ((((sig_n_prime.w[1] == 0) && (sig_n_prime.w[0] < sig_y))) != (((x & 0x8000000000000000) != 0x8000000000000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    sig_n_prime = __mul_64x64_to_128(sig_y, bid_mult_factor[(exp_y.wrapping_sub(exp_x)) as usize]);
    if ((sig_n_prime.w[1] == 0) && (sig_n_prime.w[0] == sig_x)) {
        res = 1;
        return (res, pfpsf);
    }
    res = 0;
    if ((((sig_n_prime.w[1] > 0) || (sig_x < sig_n_prime.w[0]))) != (((x & 0x8000000000000000) != 0x8000000000000000))) {
        res = 1;
    }
    return (res, pfpsf);
}

pub fn bid64_signaling_greater_unordered(mut x: u64, mut y: u64) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut sig_x: u64 = 0;
    let mut sig_y: u64 = 0;
    let mut sig_n_prime: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut pfpsf: u32 = 0;
    let mut x_is_zero: i64 = 0;
    let mut y_is_zero: i64 = 0;
    let mut non_canon_x: i64 = 0;
    let mut non_canon_y: i64 = 0;
    if ((((x & 0x7c00000000000000) == 0x7c00000000000000)) || (((y & 0x7c00000000000000) == 0x7c00000000000000))) {
        pfpsf |= 1;
        res = 1;
        return (res, pfpsf);
    }
    if (x == y) {
        res = 0;
        return (res, pfpsf);
    }
    if ((x & 0x7800000000000000) == 0x7800000000000000) {
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = 0;
            return (res, pfpsf);
        }
        res = 0;
        if ((((y & 0x7800000000000000) != 0x7800000000000000)) || (((y & 0x8000000000000000) == 0x8000000000000000))) {
            res = 1;
        }
        return (res, pfpsf);
    } else if ((y & 0x7800000000000000) == 0x7800000000000000) {
        res = 0;
        if ((y & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((x & 0x6000000000000000) == 0x6000000000000000) {
        exp_x = ((go_checked_shr_u64((x & 0x1ff8000000000000), go_shift_count_u64((51) as u64))) as i64);
        sig_x = ((x & 0x7ffffffffffff) | 0x20000000000000);
        if (sig_x > 9999999999999999) {
            non_canon_x = 1;
        } else {
            non_canon_x = 0;
        }
    } else {
        exp_x = ((go_checked_shr_u64((x & 0x7fe0000000000000), go_shift_count_u64((53) as u64))) as i64);
        sig_x = (x & 0x1fffffffffffff);
        non_canon_x = 0;
    }
    if ((y & 0x6000000000000000) == 0x6000000000000000) {
        exp_y = ((go_checked_shr_u64((y & 0x1ff8000000000000), go_shift_count_u64((51) as u64))) as i64);
        sig_y = ((y & 0x7ffffffffffff) | 0x20000000000000);
        if (sig_y > 9999999999999999) {
            non_canon_y = 1;
        } else {
            non_canon_y = 0;
        }
    } else {
        exp_y = ((go_checked_shr_u64((y & 0x7fe0000000000000), go_shift_count_u64((53) as u64))) as i64);
        sig_y = (y & 0x1fffffffffffff);
        non_canon_y = 0;
    }
    if ((non_canon_x != 0) || (sig_x == 0)) {
        x_is_zero = 1;
    }
    if ((non_canon_y != 0) || (sig_y == 0)) {
        y_is_zero = 1;
    }
    if ((x_is_zero != 0) && (y_is_zero != 0)) {
        res = 0;
        return (res, pfpsf);
    }
    if (x_is_zero != 0) {
        res = 0;
        if ((y & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (y_is_zero != 0) {
        res = 0;
        if ((x & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((((x ^ y) & 0x8000000000000000)) == 0x8000000000000000) {
        res = 0;
        if ((y & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((sig_x > sig_y) && (exp_x >= exp_y)) {
        res = 0;
        if ((x & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((sig_x < sig_y) && (exp_x <= exp_y)) {
        res = 0;
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((exp_x.wrapping_sub(exp_y)) > 15) {
        res = 0;
        if ((x & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((exp_y.wrapping_sub(exp_x)) > 15) {
        res = 0;
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (exp_x > exp_y) {
        sig_n_prime = __mul_64x64_to_128(sig_x, bid_mult_factor[(exp_x.wrapping_sub(exp_y)) as usize]);
        if ((sig_n_prime.w[1] == 0) && (sig_n_prime.w[0] == sig_y)) {
            res = 0;
            return (res, pfpsf);
        }
        res = 0;
        if ((((sig_n_prime.w[1] > 0) || (sig_n_prime.w[0] > sig_y))) != (((x & 0x8000000000000000) == 0x8000000000000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    sig_n_prime = __mul_64x64_to_128(sig_y, bid_mult_factor[(exp_y.wrapping_sub(exp_x)) as usize]);
    if ((sig_n_prime.w[1] == 0) && (sig_n_prime.w[0] == sig_x)) {
        res = 0;
        return (res, pfpsf);
    }
    res = 0;
    if ((((sig_n_prime.w[1] == 0) && (sig_x > sig_n_prime.w[0]))) != (((x & 0x8000000000000000) == 0x8000000000000000))) {
        res = 1;
    }
    return (res, pfpsf);
}

pub fn bid64_signaling_less_equal(mut x: u64, mut y: u64) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut sig_x: u64 = 0;
    let mut sig_y: u64 = 0;
    let mut sig_n_prime: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut pfpsf: u32 = 0;
    let mut x_is_zero: i64 = 0;
    let mut y_is_zero: i64 = 0;
    let mut non_canon_x: i64 = 0;
    let mut non_canon_y: i64 = 0;
    if ((((x & 0x7c00000000000000) == 0x7c00000000000000)) || (((y & 0x7c00000000000000) == 0x7c00000000000000))) {
        pfpsf |= 1;
        res = 0;
        return (res, pfpsf);
    }
    if (x == y) {
        res = 1;
        return (res, pfpsf);
    }
    if ((x & 0x7800000000000000) == 0x7800000000000000) {
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
            return (res, pfpsf);
        }
        res = 1;
        if ((((y & 0x7800000000000000) != 0x7800000000000000)) || (((y & 0x8000000000000000) == 0x8000000000000000))) {
            res = 0;
        }
        return (res, pfpsf);
    } else if ((y & 0x7800000000000000) == 0x7800000000000000) {
        res = 0;
        if ((y & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((x & 0x6000000000000000) == 0x6000000000000000) {
        exp_x = ((go_checked_shr_u64((x & 0x1ff8000000000000), go_shift_count_u64((51) as u64))) as i64);
        sig_x = ((x & 0x7ffffffffffff) | 0x20000000000000);
        if (sig_x > 9999999999999999) {
            non_canon_x = 1;
        } else {
            non_canon_x = 0;
        }
    } else {
        exp_x = ((go_checked_shr_u64((x & 0x7fe0000000000000), go_shift_count_u64((53) as u64))) as i64);
        sig_x = (x & 0x1fffffffffffff);
        non_canon_x = 0;
    }
    if ((y & 0x6000000000000000) == 0x6000000000000000) {
        exp_y = ((go_checked_shr_u64((y & 0x1ff8000000000000), go_shift_count_u64((51) as u64))) as i64);
        sig_y = ((y & 0x7ffffffffffff) | 0x20000000000000);
        if (sig_y > 9999999999999999) {
            non_canon_y = 1;
        } else {
            non_canon_y = 0;
        }
    } else {
        exp_y = ((go_checked_shr_u64((y & 0x7fe0000000000000), go_shift_count_u64((53) as u64))) as i64);
        sig_y = (y & 0x1fffffffffffff);
        non_canon_y = 0;
    }
    if ((non_canon_x != 0) || (sig_x == 0)) {
        x_is_zero = 1;
    }
    if ((non_canon_y != 0) || (sig_y == 0)) {
        y_is_zero = 1;
    }
    if ((x_is_zero != 0) && (y_is_zero != 0)) {
        res = 1;
        return (res, pfpsf);
    }
    if (x_is_zero != 0) {
        res = 0;
        if ((y & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (y_is_zero != 0) {
        res = 0;
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((((x ^ y) & 0x8000000000000000)) == 0x8000000000000000) {
        res = 0;
        if ((y & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((sig_x > sig_y) && (exp_x >= exp_y)) {
        res = 0;
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((sig_x < sig_y) && (exp_x <= exp_y)) {
        res = 0;
        if ((x & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((exp_x.wrapping_sub(exp_y)) > 15) {
        res = 0;
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((exp_y.wrapping_sub(exp_x)) > 15) {
        res = 0;
        if ((x & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (exp_x > exp_y) {
        sig_n_prime = __mul_64x64_to_128(sig_x, bid_mult_factor[(exp_x.wrapping_sub(exp_y)) as usize]);
        if ((sig_n_prime.w[1] == 0) && (sig_n_prime.w[0] == sig_y)) {
            res = 1;
            return (res, pfpsf);
        }
        res = 0;
        if ((((sig_n_prime.w[1] == 0) && (sig_n_prime.w[0] < sig_y))) != (((x & 0x8000000000000000) == 0x8000000000000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    sig_n_prime = __mul_64x64_to_128(sig_y, bid_mult_factor[(exp_y.wrapping_sub(exp_x)) as usize]);
    if ((sig_n_prime.w[1] == 0) && (sig_n_prime.w[0] == sig_x)) {
        res = 1;
        return (res, pfpsf);
    }
    res = 0;
    if ((((sig_n_prime.w[1] > 0) || (sig_x < sig_n_prime.w[0]))) != (((x & 0x8000000000000000) == 0x8000000000000000))) {
        res = 1;
    }
    return (res, pfpsf);
}

pub fn bid64_signaling_less_unordered(mut x: u64, mut y: u64) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut sig_x: u64 = 0;
    let mut sig_y: u64 = 0;
    let mut sig_n_prime: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut pfpsf: u32 = 0;
    let mut x_is_zero: i64 = 0;
    let mut y_is_zero: i64 = 0;
    let mut non_canon_x: i64 = 0;
    let mut non_canon_y: i64 = 0;
    if ((((x & 0x7c00000000000000) == 0x7c00000000000000)) || (((y & 0x7c00000000000000) == 0x7c00000000000000))) {
        pfpsf |= 1;
        res = 1;
        return (res, pfpsf);
    }
    if (x == y) {
        res = 0;
        return (res, pfpsf);
    }
    if ((x & 0x7800000000000000) == 0x7800000000000000) {
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = 0;
            if ((((y & 0x7800000000000000) != 0x7800000000000000)) || ((y & 0x8000000000000000) != 0x8000000000000000)) {
                res = 1;
            }
            return (res, pfpsf);
        }
        res = 0;
        return (res, pfpsf);
    } else if ((y & 0x7800000000000000) == 0x7800000000000000) {
        res = 0;
        if ((y & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((x & 0x6000000000000000) == 0x6000000000000000) {
        exp_x = ((go_checked_shr_u64((x & 0x1ff8000000000000), go_shift_count_u64((51) as u64))) as i64);
        sig_x = ((x & 0x7ffffffffffff) | 0x20000000000000);
        if (sig_x > 9999999999999999) {
            non_canon_x = 1;
        } else {
            non_canon_x = 0;
        }
    } else {
        exp_x = ((go_checked_shr_u64((x & 0x7fe0000000000000), go_shift_count_u64((53) as u64))) as i64);
        sig_x = (x & 0x1fffffffffffff);
        non_canon_x = 0;
    }
    if ((y & 0x6000000000000000) == 0x6000000000000000) {
        exp_y = ((go_checked_shr_u64((y & 0x1ff8000000000000), go_shift_count_u64((51) as u64))) as i64);
        sig_y = ((y & 0x7ffffffffffff) | 0x20000000000000);
        if (sig_y > 9999999999999999) {
            non_canon_y = 1;
        } else {
            non_canon_y = 0;
        }
    } else {
        exp_y = ((go_checked_shr_u64((y & 0x7fe0000000000000), go_shift_count_u64((53) as u64))) as i64);
        sig_y = (y & 0x1fffffffffffff);
        non_canon_y = 0;
    }
    if ((non_canon_x != 0) || (sig_x == 0)) {
        x_is_zero = 1;
    }
    if ((non_canon_y != 0) || (sig_y == 0)) {
        y_is_zero = 1;
    }
    if ((x_is_zero != 0) && (y_is_zero != 0)) {
        res = 0;
        return (res, pfpsf);
    }
    if (x_is_zero != 0) {
        res = 0;
        if ((y & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (y_is_zero != 0) {
        res = 0;
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((((x ^ y) & 0x8000000000000000)) == 0x8000000000000000) {
        res = 0;
        if ((y & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((sig_x > sig_y) && (exp_x >= exp_y)) {
        res = 0;
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((sig_x < sig_y) && (exp_x <= exp_y)) {
        res = 0;
        if ((x & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((exp_x.wrapping_sub(exp_y)) > 15) {
        res = 0;
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((exp_y.wrapping_sub(exp_x)) > 15) {
        res = 0;
        if ((x & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (exp_x > exp_y) {
        sig_n_prime = __mul_64x64_to_128(sig_x, bid_mult_factor[(exp_x.wrapping_sub(exp_y)) as usize]);
        if ((sig_n_prime.w[1] == 0) && (sig_n_prime.w[0] == sig_y)) {
            res = 0;
            return (res, pfpsf);
        }
        res = 0;
        if ((((sig_n_prime.w[1] == 0) && (sig_n_prime.w[0] < sig_y))) != (((x & 0x8000000000000000) == 0x8000000000000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    sig_n_prime = __mul_64x64_to_128(sig_y, bid_mult_factor[(exp_y.wrapping_sub(exp_x)) as usize]);
    if ((sig_n_prime.w[1] == 0) && (sig_n_prime.w[0] == sig_x)) {
        res = 0;
        return (res, pfpsf);
    }
    res = 0;
    if ((((sig_n_prime.w[1] > 0) || (sig_x < sig_n_prime.w[0]))) != (((x & 0x8000000000000000) == 0x8000000000000000))) {
        res = 1;
    }
    return (res, pfpsf);
}

pub fn bid64_signaling_not_greater(mut x: u64, mut y: u64) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut sig_x: u64 = 0;
    let mut sig_y: u64 = 0;
    let mut sig_n_prime: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut pfpsf: u32 = 0;
    let mut x_is_zero: i64 = 0;
    let mut y_is_zero: i64 = 0;
    let mut non_canon_x: i64 = 0;
    let mut non_canon_y: i64 = 0;
    if ((((x & 0x7c00000000000000) == 0x7c00000000000000)) || (((y & 0x7c00000000000000) == 0x7c00000000000000))) {
        pfpsf |= 1;
        res = 1;
        return (res, pfpsf);
    }
    if (x == y) {
        res = 1;
        return (res, pfpsf);
    }
    if ((x & 0x7800000000000000) == 0x7800000000000000) {
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
            return (res, pfpsf);
        }
        res = 1;
        if ((((y & 0x7800000000000000) != 0x7800000000000000)) || (((y & 0x8000000000000000) == 0x8000000000000000))) {
            res = 0;
        }
        return (res, pfpsf);
    } else if ((y & 0x7800000000000000) == 0x7800000000000000) {
        res = 0;
        if ((y & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((x & 0x6000000000000000) == 0x6000000000000000) {
        exp_x = ((go_checked_shr_u64((x & 0x1ff8000000000000), go_shift_count_u64((51) as u64))) as i64);
        sig_x = ((x & 0x7ffffffffffff) | 0x20000000000000);
        if (sig_x > 9999999999999999) {
            non_canon_x = 1;
        } else {
            non_canon_x = 0;
        }
    } else {
        exp_x = ((go_checked_shr_u64((x & 0x7fe0000000000000), go_shift_count_u64((53) as u64))) as i64);
        sig_x = (x & 0x1fffffffffffff);
        non_canon_x = 0;
    }
    if ((y & 0x6000000000000000) == 0x6000000000000000) {
        exp_y = ((go_checked_shr_u64((y & 0x1ff8000000000000), go_shift_count_u64((51) as u64))) as i64);
        sig_y = ((y & 0x7ffffffffffff) | 0x20000000000000);
        if (sig_y > 9999999999999999) {
            non_canon_y = 1;
        } else {
            non_canon_y = 0;
        }
    } else {
        exp_y = ((go_checked_shr_u64((y & 0x7fe0000000000000), go_shift_count_u64((53) as u64))) as i64);
        sig_y = (y & 0x1fffffffffffff);
        non_canon_y = 0;
    }
    if ((non_canon_x != 0) || (sig_x == 0)) {
        x_is_zero = 1;
    }
    if ((non_canon_y != 0) || (sig_y == 0)) {
        y_is_zero = 1;
    }
    if ((x_is_zero != 0) && (y_is_zero != 0)) {
        res = 1;
        return (res, pfpsf);
    }
    if (x_is_zero != 0) {
        res = 0;
        if ((y & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (y_is_zero != 0) {
        res = 0;
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((((x ^ y) & 0x8000000000000000)) == 0x8000000000000000) {
        res = 0;
        if ((y & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((sig_x > sig_y) && (exp_x >= exp_y)) {
        res = 0;
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((sig_x < sig_y) && (exp_x <= exp_y)) {
        res = 0;
        if ((x & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((exp_x.wrapping_sub(exp_y)) > 15) {
        res = 0;
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((exp_y.wrapping_sub(exp_x)) > 15) {
        res = 0;
        if ((x & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (exp_x > exp_y) {
        sig_n_prime = __mul_64x64_to_128(sig_x, bid_mult_factor[(exp_x.wrapping_sub(exp_y)) as usize]);
        if ((sig_n_prime.w[1] == 0) && (sig_n_prime.w[0] == sig_y)) {
            res = 1;
            return (res, pfpsf);
        }
        res = 0;
        if ((((sig_n_prime.w[1] == 0) && (sig_n_prime.w[0] < sig_y))) != (((x & 0x8000000000000000) == 0x8000000000000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    sig_n_prime = __mul_64x64_to_128(sig_y, bid_mult_factor[(exp_y.wrapping_sub(exp_x)) as usize]);
    if ((sig_n_prime.w[1] == 0) && (sig_n_prime.w[0] == sig_x)) {
        res = 1;
        return (res, pfpsf);
    }
    res = 0;
    if ((((sig_n_prime.w[1] > 0) || (sig_x < sig_n_prime.w[0]))) != (((x & 0x8000000000000000) == 0x8000000000000000))) {
        res = 1;
    }
    return (res, pfpsf);
}

pub fn bid64_signaling_not_less(mut x: u64, mut y: u64) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut sig_x: u64 = 0;
    let mut sig_y: u64 = 0;
    let mut sig_n_prime: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut pfpsf: u32 = 0;
    let mut x_is_zero: i64 = 0;
    let mut y_is_zero: i64 = 0;
    let mut non_canon_x: i64 = 0;
    let mut non_canon_y: i64 = 0;
    if ((((x & 0x7c00000000000000) == 0x7c00000000000000)) || (((y & 0x7c00000000000000) == 0x7c00000000000000))) {
        pfpsf |= 1;
        res = 1;
        return (res, pfpsf);
    }
    if (x == y) {
        res = 1;
        return (res, pfpsf);
    }
    if ((x & 0x7800000000000000) == 0x7800000000000000) {
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = 0;
            if ((((y & 0x7800000000000000) == 0x7800000000000000)) && ((y & 0x8000000000000000) == 0x8000000000000000)) {
                res = 1;
            }
            return (res, pfpsf);
        }
        res = 1;
        return (res, pfpsf);
    } else if ((y & 0x7800000000000000) == 0x7800000000000000) {
        res = 0;
        if ((y & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((x & 0x6000000000000000) == 0x6000000000000000) {
        exp_x = ((go_checked_shr_u64((x & 0x1ff8000000000000), go_shift_count_u64((51) as u64))) as i64);
        sig_x = ((x & 0x7ffffffffffff) | 0x20000000000000);
        if (sig_x > 9999999999999999) {
            non_canon_x = 1;
        } else {
            non_canon_x = 0;
        }
    } else {
        exp_x = ((go_checked_shr_u64((x & 0x7fe0000000000000), go_shift_count_u64((53) as u64))) as i64);
        sig_x = (x & 0x1fffffffffffff);
        non_canon_x = 0;
    }
    if ((y & 0x6000000000000000) == 0x6000000000000000) {
        exp_y = ((go_checked_shr_u64((y & 0x1ff8000000000000), go_shift_count_u64((51) as u64))) as i64);
        sig_y = ((y & 0x7ffffffffffff) | 0x20000000000000);
        if (sig_y > 9999999999999999) {
            non_canon_y = 1;
        } else {
            non_canon_y = 0;
        }
    } else {
        exp_y = ((go_checked_shr_u64((y & 0x7fe0000000000000), go_shift_count_u64((53) as u64))) as i64);
        sig_y = (y & 0x1fffffffffffff);
        non_canon_y = 0;
    }
    if ((non_canon_x != 0) || (sig_x == 0)) {
        x_is_zero = 1;
    }
    if ((non_canon_y != 0) || (sig_y == 0)) {
        y_is_zero = 1;
    }
    if ((x_is_zero != 0) && (y_is_zero != 0)) {
        res = 1;
        return (res, pfpsf);
    }
    if (x_is_zero != 0) {
        res = 0;
        if ((y & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (y_is_zero != 0) {
        res = 0;
        if ((x & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((((x ^ y) & 0x8000000000000000)) == 0x8000000000000000) {
        res = 0;
        if ((y & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((sig_x > sig_y) && (exp_x >= exp_y)) {
        res = 0;
        if ((x & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((sig_x < sig_y) && (exp_x <= exp_y)) {
        res = 0;
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((exp_x.wrapping_sub(exp_y)) > 15) {
        res = 0;
        if ((x & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((exp_y.wrapping_sub(exp_x)) > 15) {
        res = 0;
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (exp_x > exp_y) {
        sig_n_prime = __mul_64x64_to_128(sig_x, bid_mult_factor[(exp_x.wrapping_sub(exp_y)) as usize]);
        if ((sig_n_prime.w[1] == 0) && (sig_n_prime.w[0] == sig_y)) {
            res = 1;
            return (res, pfpsf);
        }
        res = 0;
        if ((((sig_n_prime.w[1] == 0) && (sig_n_prime.w[0] < sig_y))) != (((x & 0x8000000000000000) != 0x8000000000000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    sig_n_prime = __mul_64x64_to_128(sig_y, bid_mult_factor[(exp_y.wrapping_sub(exp_x)) as usize]);
    if ((sig_n_prime.w[1] == 0) && (sig_n_prime.w[0] == sig_x)) {
        res = 1;
        return (res, pfpsf);
    }
    res = 0;
    if ((((sig_n_prime.w[1] > 0) || (sig_x < sig_n_prime.w[0]))) != (((x & 0x8000000000000000) != 0x8000000000000000))) {
        res = 1;
    }
    return (res, pfpsf);
}
