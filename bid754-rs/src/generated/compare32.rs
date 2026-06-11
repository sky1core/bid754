// Auto-generated from compare32.go by go2rs. Do not edit.

use super::prelude::*;

pub fn bid32_quiet_equal(mut x: u32, mut y: u32) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut exp_t: i64 = 0;
    let mut sig_x: u32 = 0;
    let mut sig_y: u32 = 0;
    let mut sig_t: u32 = 0;
    let mut pfpsf: u32 = 0;
    let mut x_is_zero: i64 = 0;
    let mut y_is_zero: i64 = 0;
    let mut non_canon_x: i64 = 0;
    let mut non_canon_y: i64 = 0;
    let mut lcv: i64 = 0;
    if ((((x & 0x7c000000) == 0x7c000000)) || (((y & 0x7c000000) == 0x7c000000))) {
        if (((x & 0x7e000000) == 0x7e000000) || ((y & 0x7e000000) == 0x7e000000)) {
            pfpsf |= 1;
        }
        res = 0;
        return (res, pfpsf);
    }
    if (x == y) {
        res = 1;
        return (res, pfpsf);
    }
    if ((((x & 0x78000000) == 0x78000000)) && (((y & 0x78000000) == 0x78000000))) {
        res = 0;
        if ((((x ^ y) & 0x80000000)) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((((x & 0x78000000) == 0x78000000)) || (((y & 0x78000000) == 0x78000000))) {
        res = 0;
        return (res, pfpsf);
    }
    if ((x & 0x60000000) == 0x60000000) {
        exp_x = ((go_checked_shr_u32((x & 0x1fe00000), go_shift_count_u64((21) as u64))) as i64);
        sig_x = ((x & 0x1fffff) | 0x800000);
        if (sig_x > 9999999) {
            non_canon_x = 1;
        } else {
            non_canon_x = 0;
        }
    } else {
        exp_x = ((go_checked_shr_u32((x & 0x7f800000), go_shift_count_u64((23) as u64))) as i64);
        sig_x = (x & 0x7fffff);
        non_canon_x = 0;
    }
    if ((y & 0x60000000) == 0x60000000) {
        exp_y = ((go_checked_shr_u32((y & 0x1fe00000), go_shift_count_u64((21) as u64))) as i64);
        sig_y = ((y & 0x1fffff) | 0x800000);
        if (sig_y > 9999999) {
            non_canon_y = 1;
        } else {
            non_canon_y = 0;
        }
    } else {
        exp_y = ((go_checked_shr_u32((y & 0x7f800000), go_shift_count_u64((23) as u64))) as i64);
        sig_y = (y & 0x7fffff);
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
    if ((((x ^ y) & 0x80000000)) != 0) {
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
    if ((exp_y.wrapping_sub(exp_x)) > 6) {
        res = 0;
        return (res, pfpsf);
    }
    lcv = 0;
    while (lcv < ((exp_y.wrapping_sub(exp_x)))) {
        sig_y = (sig_y.wrapping_mul(10));
        if (sig_y > 9999999) {
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

pub fn bid32_quiet_greater(mut x: u32, mut y: u32) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut sig_x: u32 = 0;
    let mut sig_y: u32 = 0;
    let mut sig_n_prime: u64 = 0;
    let mut pfpsf: u32 = 0;
    let mut x_is_zero: i64 = 0;
    let mut y_is_zero: i64 = 0;
    let mut non_canon_x: i64 = 0;
    let mut non_canon_y: i64 = 0;
    if ((((x & 0x7c000000) == 0x7c000000)) || (((y & 0x7c000000) == 0x7c000000))) {
        if (((x & 0x7e000000) == 0x7e000000) || ((y & 0x7e000000) == 0x7e000000)) {
            pfpsf |= 1;
        }
        res = 0;
        return (res, pfpsf);
    }
    if (x == y) {
        res = 0;
        return (res, pfpsf);
    }
    if ((x & 0x78000000) == 0x78000000) {
        if ((x & 0x80000000) == 0x80000000) {
            res = 0;
            return (res, pfpsf);
        } else {
            res = 0;
            if ((((y & 0x78000000) != 0x78000000)) || (((y & 0x80000000) == 0x80000000))) {
                res = 1;
            }
            return (res, pfpsf);
        }
    } else if ((y & 0x78000000) == 0x78000000) {
        res = 0;
        if ((y & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((x & 0x60000000) == 0x60000000) {
        exp_x = ((go_checked_shr_u32((x & 0x1fe00000), go_shift_count_u64((21) as u64))) as i64);
        sig_x = ((x & 0x1fffff) | 0x800000);
        if (sig_x > 9999999) {
            non_canon_x = 1;
        } else {
            non_canon_x = 0;
        }
    } else {
        exp_x = ((go_checked_shr_u32((x & 0x7f800000), go_shift_count_u64((23) as u64))) as i64);
        sig_x = (x & 0x7fffff);
        non_canon_x = 0;
    }
    if ((y & 0x60000000) == 0x60000000) {
        exp_y = ((go_checked_shr_u32((y & 0x1fe00000), go_shift_count_u64((21) as u64))) as i64);
        sig_y = ((y & 0x1fffff) | 0x800000);
        if (sig_y > 9999999) {
            non_canon_y = 1;
        } else {
            non_canon_y = 0;
        }
    } else {
        exp_y = ((go_checked_shr_u32((y & 0x7f800000), go_shift_count_u64((23) as u64))) as i64);
        sig_y = (y & 0x7fffff);
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
        if ((y & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    } else if (y_is_zero != 0) {
        res = 0;
        if ((x & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((((x ^ y) & 0x80000000)) == 0x80000000) {
        res = 0;
        if ((y & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((sig_x > sig_y) && (exp_x > exp_y)) {
        res = 0;
        if ((x & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((sig_x < sig_y) && (exp_x < exp_y)) {
        res = 0;
        if ((x & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((exp_x.wrapping_sub(exp_y)) > 6) {
        if ((x & 0x80000000) != 0) {
            res = 0;
        } else {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((exp_y.wrapping_sub(exp_x)) > 6) {
        if ((x & 0x80000000) != 0) {
            res = 1;
        } else {
            res = 0;
        }
        return (res, pfpsf);
    }
    if (exp_x > exp_y) {
        sig_n_prime = ((sig_x as u64).wrapping_mul(bid32_mult_factor[(exp_x.wrapping_sub(exp_y)) as usize]));
        if ((0 == 0) && ((sig_n_prime == (sig_y as u64)))) {
            res = 0;
            return (res, pfpsf);
        }
        res = 0;
        if ((((0 > 0) || (sig_n_prime > (sig_y as u64)))) != (((x & 0x80000000) == 0x80000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    sig_n_prime = ((sig_y as u64).wrapping_mul(bid32_mult_factor[(exp_y.wrapping_sub(exp_x)) as usize]));
    if ((0 == 0) && ((sig_n_prime == (sig_x as u64)))) {
        res = 0;
        return (res, pfpsf);
    }
    res = 0;
    if ((((sig_x as u64) > sig_n_prime)) != (((x & 0x80000000) == 0x80000000))) {
        res = 1;
    }
    return (res, pfpsf);
}

pub fn bid32_quiet_greater_equal(mut x: u32, mut y: u32) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut sig_x: u32 = 0;
    let mut sig_y: u32 = 0;
    let mut sig_n_prime: u64 = 0;
    let mut pfpsf: u32 = 0;
    let mut x_is_zero: i64 = 0;
    let mut y_is_zero: i64 = 0;
    let mut non_canon_x: i64 = 0;
    let mut non_canon_y: i64 = 0;
    if ((((x & 0x7c000000) == 0x7c000000)) || (((y & 0x7c000000) == 0x7c000000))) {
        if (((x & 0x7e000000) == 0x7e000000) || ((y & 0x7e000000) == 0x7e000000)) {
            pfpsf |= 1;
        }
        res = 0;
        return (res, pfpsf);
    }
    if (x == y) {
        res = 1;
        return (res, pfpsf);
    }
    if ((x & 0x78000000) == 0x78000000) {
        if ((x & 0x80000000) == 0x80000000) {
            res = 0;
            if ((((y & 0x78000000) == 0x78000000)) && ((y & 0x80000000) == 0x80000000)) {
                res = 1;
            }
            return (res, pfpsf);
        } else {
            res = 1;
            return (res, pfpsf);
        }
    } else if ((y & 0x78000000) == 0x78000000) {
        res = 0;
        if ((y & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((x & 0x60000000) == 0x60000000) {
        exp_x = ((go_checked_shr_u32((x & 0x1fe00000), go_shift_count_u64((21) as u64))) as i64);
        sig_x = ((x & 0x1fffff) | 0x800000);
        if (sig_x > 9999999) {
            non_canon_x = 1;
        } else {
            non_canon_x = 0;
        }
    } else {
        exp_x = ((go_checked_shr_u32((x & 0x7f800000), go_shift_count_u64((23) as u64))) as i64);
        sig_x = (x & 0x7fffff);
        non_canon_x = 0;
    }
    if ((y & 0x60000000) == 0x60000000) {
        exp_y = ((go_checked_shr_u32((y & 0x1fe00000), go_shift_count_u64((21) as u64))) as i64);
        sig_y = ((y & 0x1fffff) | 0x800000);
        if (sig_y > 9999999) {
            non_canon_y = 1;
        } else {
            non_canon_y = 0;
        }
    } else {
        exp_y = ((go_checked_shr_u32((y & 0x7f800000), go_shift_count_u64((23) as u64))) as i64);
        sig_y = (y & 0x7fffff);
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
        if ((y & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    } else if (y_is_zero != 0) {
        res = 0;
        if ((x & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((((x ^ y) & 0x80000000)) == 0x80000000) {
        res = 0;
        if ((y & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((sig_x > sig_y) && (exp_x >= exp_y)) {
        res = 0;
        if ((x & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((sig_x < sig_y) && (exp_x <= exp_y)) {
        res = 0;
        if ((x & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((exp_x.wrapping_sub(exp_y)) > 6) {
        res = 0;
        if ((x & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((exp_y.wrapping_sub(exp_x)) > 6) {
        res = 0;
        if ((x & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (exp_x > exp_y) {
        sig_n_prime = ((sig_x as u64).wrapping_mul(bid32_mult_factor[(exp_x.wrapping_sub(exp_y)) as usize]));
        if ((0 == 0) && ((sig_n_prime == (sig_y as u64)))) {
            res = 1;
            return (res, pfpsf);
        }
        res = 0;
        if (((sig_n_prime < (sig_y as u64))) != (((x & 0x80000000) != 0x80000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    sig_n_prime = ((sig_y as u64).wrapping_mul(bid32_mult_factor[(exp_y.wrapping_sub(exp_x)) as usize]));
    if ((0 == 0) && ((sig_n_prime == (sig_x as u64)))) {
        res = 1;
        return (res, pfpsf);
    }
    res = 0;
    if ((((sig_x as u64) < sig_n_prime)) != (((x & 0x80000000) != 0x80000000))) {
        res = 1;
    }
    return (res, pfpsf);
}

pub fn bid32_quiet_greater_unordered(mut x: u32, mut y: u32) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut sig_x: u32 = 0;
    let mut sig_y: u32 = 0;
    let mut sig_n_prime: u64 = 0;
    let mut pfpsf: u32 = 0;
    let mut x_is_zero: i64 = 0;
    let mut y_is_zero: i64 = 0;
    let mut non_canon_x: i64 = 0;
    let mut non_canon_y: i64 = 0;
    if ((((x & 0x7c000000) == 0x7c000000)) || (((y & 0x7c000000) == 0x7c000000))) {
        if (((x & 0x7e000000) == 0x7e000000) || ((y & 0x7e000000) == 0x7e000000)) {
            pfpsf |= 1;
        }
        res = 1;
        return (res, pfpsf);
    }
    if (x == y) {
        res = 0;
        return (res, pfpsf);
    }
    if ((x & 0x78000000) == 0x78000000) {
        if ((x & 0x80000000) == 0x80000000) {
            res = 0;
            return (res, pfpsf);
        } else {
            res = 0;
            if ((((y & 0x78000000) != 0x78000000)) || (((y & 0x80000000) == 0x80000000))) {
                res = 1;
            }
            return (res, pfpsf);
        }
    } else if ((y & 0x78000000) == 0x78000000) {
        res = 0;
        if ((y & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((x & 0x60000000) == 0x60000000) {
        exp_x = ((go_checked_shr_u32((x & 0x1fe00000), go_shift_count_u64((21) as u64))) as i64);
        sig_x = ((x & 0x1fffff) | 0x800000);
        if (sig_x > 9999999) {
            non_canon_x = 1;
        } else {
            non_canon_x = 0;
        }
    } else {
        exp_x = ((go_checked_shr_u32((x & 0x7f800000), go_shift_count_u64((23) as u64))) as i64);
        sig_x = (x & 0x7fffff);
        non_canon_x = 0;
    }
    if ((y & 0x60000000) == 0x60000000) {
        exp_y = ((go_checked_shr_u32((y & 0x1fe00000), go_shift_count_u64((21) as u64))) as i64);
        sig_y = ((y & 0x1fffff) | 0x800000);
        if (sig_y > 9999999) {
            non_canon_y = 1;
        } else {
            non_canon_y = 0;
        }
    } else {
        exp_y = ((go_checked_shr_u32((y & 0x7f800000), go_shift_count_u64((23) as u64))) as i64);
        sig_y = (y & 0x7fffff);
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
        if ((y & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    } else if (y_is_zero != 0) {
        res = 0;
        if ((x & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((((x ^ y) & 0x80000000)) == 0x80000000) {
        res = 0;
        if ((y & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((sig_x > sig_y) && (exp_x >= exp_y)) {
        res = 0;
        if ((x & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((sig_x < sig_y) && (exp_x <= exp_y)) {
        res = 0;
        if ((x & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((exp_x.wrapping_sub(exp_y)) > 6) {
        res = 0;
        if ((x & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((exp_y.wrapping_sub(exp_x)) > 6) {
        res = 0;
        if ((x & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (exp_x > exp_y) {
        sig_n_prime = ((sig_x as u64).wrapping_mul(bid32_mult_factor[(exp_x.wrapping_sub(exp_y)) as usize]));
        if ((0 == 0) && ((sig_n_prime == (sig_y as u64)))) {
            res = 0;
            return (res, pfpsf);
        }
        res = 0;
        if ((((0 > 0) || (sig_n_prime > (sig_y as u64)))) != (((x & 0x80000000) == 0x80000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    sig_n_prime = ((sig_y as u64).wrapping_mul(bid32_mult_factor[(exp_y.wrapping_sub(exp_x)) as usize]));
    if ((0 == 0) && ((sig_n_prime == (sig_x as u64)))) {
        res = 0;
        return (res, pfpsf);
    }
    res = 0;
    if ((((sig_x as u64) > sig_n_prime)) != (((x & 0x80000000) == 0x80000000))) {
        res = 1;
    }
    return (res, pfpsf);
}

pub fn bid32_quiet_less(mut x: u32, mut y: u32) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut sig_x: u32 = 0;
    let mut sig_y: u32 = 0;
    let mut sig_n_prime: u64 = 0;
    let mut pfpsf: u32 = 0;
    let mut x_is_zero: i64 = 0;
    let mut y_is_zero: i64 = 0;
    let mut non_canon_x: i64 = 0;
    let mut non_canon_y: i64 = 0;
    if ((((x & 0x7c000000) == 0x7c000000)) || (((y & 0x7c000000) == 0x7c000000))) {
        if (((x & 0x7e000000) == 0x7e000000) || ((y & 0x7e000000) == 0x7e000000)) {
            pfpsf |= 1;
        }
        res = 0;
        return (res, pfpsf);
    }
    if (x == y) {
        res = 0;
        return (res, pfpsf);
    }
    if ((x & 0x78000000) == 0x78000000) {
        if ((x & 0x80000000) == 0x80000000) {
            res = 0;
            if ((((y & 0x78000000) != 0x78000000)) || ((y & 0x80000000) != 0x80000000)) {
                res = 1;
            }
            return (res, pfpsf);
        } else {
            res = 0;
            return (res, pfpsf);
        }
    } else if ((y & 0x78000000) == 0x78000000) {
        res = 0;
        if ((y & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((x & 0x60000000) == 0x60000000) {
        exp_x = ((go_checked_shr_u32((x & 0x1fe00000), go_shift_count_u64((21) as u64))) as i64);
        sig_x = ((x & 0x1fffff) | 0x800000);
        if (sig_x > 9999999) {
            non_canon_x = 1;
        } else {
            non_canon_x = 0;
        }
    } else {
        exp_x = ((go_checked_shr_u32((x & 0x7f800000), go_shift_count_u64((23) as u64))) as i64);
        sig_x = (x & 0x7fffff);
        non_canon_x = 0;
    }
    if ((y & 0x60000000) == 0x60000000) {
        exp_y = ((go_checked_shr_u32((y & 0x1fe00000), go_shift_count_u64((21) as u64))) as i64);
        sig_y = ((y & 0x1fffff) | 0x800000);
        if (sig_y > 9999999) {
            non_canon_y = 1;
        } else {
            non_canon_y = 0;
        }
    } else {
        exp_y = ((go_checked_shr_u32((y & 0x7f800000), go_shift_count_u64((23) as u64))) as i64);
        sig_y = (y & 0x7fffff);
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
        if ((y & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    } else if (y_is_zero != 0) {
        res = 0;
        if ((x & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((((x ^ y) & 0x80000000)) == 0x80000000) {
        res = 0;
        if ((y & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((sig_x > sig_y) && (exp_x >= exp_y)) {
        res = 0;
        if ((x & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((sig_x < sig_y) && (exp_x <= exp_y)) {
        res = 0;
        if ((x & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((exp_x.wrapping_sub(exp_y)) > 6) {
        res = 0;
        if ((x & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((exp_y.wrapping_sub(exp_x)) > 6) {
        res = 0;
        if ((x & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (exp_x > exp_y) {
        sig_n_prime = ((sig_x as u64).wrapping_mul(bid32_mult_factor[(exp_x.wrapping_sub(exp_y)) as usize]));
        if ((0 == 0) && ((sig_n_prime == (sig_y as u64)))) {
            res = 0;
            return (res, pfpsf);
        }
        res = 0;
        if (((sig_n_prime < (sig_y as u64))) != (((x & 0x80000000) == 0x80000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    sig_n_prime = ((sig_y as u64).wrapping_mul(bid32_mult_factor[(exp_y.wrapping_sub(exp_x)) as usize]));
    if ((0 == 0) && ((sig_n_prime == (sig_x as u64)))) {
        res = 0;
        return (res, pfpsf);
    }
    res = 0;
    if ((((sig_x as u64) < sig_n_prime)) != (((x & 0x80000000) == 0x80000000))) {
        res = 1;
    }
    return (res, pfpsf);
}

pub fn bid32_quiet_less_equal(mut x: u32, mut y: u32) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut sig_x: u32 = 0;
    let mut sig_y: u32 = 0;
    let mut sig_n_prime: u64 = 0;
    let mut pfpsf: u32 = 0;
    let mut x_is_zero: i64 = 0;
    let mut y_is_zero: i64 = 0;
    let mut non_canon_x: i64 = 0;
    let mut non_canon_y: i64 = 0;
    if ((((x & 0x7c000000) == 0x7c000000)) || (((y & 0x7c000000) == 0x7c000000))) {
        if (((x & 0x7e000000) == 0x7e000000) || ((y & 0x7e000000) == 0x7e000000)) {
            pfpsf |= 1;
        }
        res = 0;
        return (res, pfpsf);
    }
    if (x == y) {
        res = 1;
        return (res, pfpsf);
    }
    if ((x & 0x78000000) == 0x78000000) {
        if ((x & 0x80000000) == 0x80000000) {
            res = 1;
            return (res, pfpsf);
        } else {
            res = 1;
            if ((((y & 0x78000000) != 0x78000000)) || (((y & 0x80000000) == 0x80000000))) {
                res = 0;
            }
            return (res, pfpsf);
        }
    } else if ((y & 0x78000000) == 0x78000000) {
        res = 0;
        if ((y & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((x & 0x60000000) == 0x60000000) {
        exp_x = ((go_checked_shr_u32((x & 0x1fe00000), go_shift_count_u64((21) as u64))) as i64);
        sig_x = ((x & 0x1fffff) | 0x800000);
        if (sig_x > 9999999) {
            non_canon_x = 1;
        } else {
            non_canon_x = 0;
        }
    } else {
        exp_x = ((go_checked_shr_u32((x & 0x7f800000), go_shift_count_u64((23) as u64))) as i64);
        sig_x = (x & 0x7fffff);
        non_canon_x = 0;
    }
    if ((y & 0x60000000) == 0x60000000) {
        exp_y = ((go_checked_shr_u32((y & 0x1fe00000), go_shift_count_u64((21) as u64))) as i64);
        sig_y = ((y & 0x1fffff) | 0x800000);
        if (sig_y > 9999999) {
            non_canon_y = 1;
        } else {
            non_canon_y = 0;
        }
    } else {
        exp_y = ((go_checked_shr_u32((y & 0x7f800000), go_shift_count_u64((23) as u64))) as i64);
        sig_y = (y & 0x7fffff);
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
        if ((y & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    } else if (y_is_zero != 0) {
        res = 0;
        if ((x & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((((x ^ y) & 0x80000000)) == 0x80000000) {
        res = 0;
        if ((y & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((sig_x > sig_y) && (exp_x >= exp_y)) {
        res = 0;
        if ((x & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((sig_x < sig_y) && (exp_x <= exp_y)) {
        res = 0;
        if ((x & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((exp_x.wrapping_sub(exp_y)) > 6) {
        res = 0;
        if ((x & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((exp_y.wrapping_sub(exp_x)) > 6) {
        res = 0;
        if ((x & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (exp_x > exp_y) {
        sig_n_prime = ((sig_x as u64).wrapping_mul(bid32_mult_factor[(exp_x.wrapping_sub(exp_y)) as usize]));
        if ((0 == 0) && ((sig_n_prime == (sig_y as u64)))) {
            res = 1;
            return (res, pfpsf);
        }
        res = 0;
        if (((sig_n_prime < (sig_y as u64))) != (((x & 0x80000000) == 0x80000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    sig_n_prime = ((sig_y as u64).wrapping_mul(bid32_mult_factor[(exp_y.wrapping_sub(exp_x)) as usize]));
    if ((0 == 0) && ((sig_n_prime == (sig_x as u64)))) {
        res = 1;
        return (res, pfpsf);
    }
    res = 0;
    if ((((sig_x as u64) < sig_n_prime)) != (((x & 0x80000000) == 0x80000000))) {
        res = 1;
    }
    return (res, pfpsf);
}

pub fn bid32_quiet_less_unordered(mut x: u32, mut y: u32) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut sig_x: u32 = 0;
    let mut sig_y: u32 = 0;
    let mut sig_n_prime: u64 = 0;
    let mut pfpsf: u32 = 0;
    let mut x_is_zero: i64 = 0;
    let mut y_is_zero: i64 = 0;
    let mut non_canon_x: i64 = 0;
    let mut non_canon_y: i64 = 0;
    if ((((x & 0x7c000000) == 0x7c000000)) || (((y & 0x7c000000) == 0x7c000000))) {
        if (((x & 0x7e000000) == 0x7e000000) || ((y & 0x7e000000) == 0x7e000000)) {
            pfpsf |= 1;
        }
        res = 1;
        return (res, pfpsf);
    }
    if (x == y) {
        res = 0;
        return (res, pfpsf);
    }
    if ((x & 0x78000000) == 0x78000000) {
        if ((x & 0x80000000) == 0x80000000) {
            res = 0;
            if ((((y & 0x78000000) != 0x78000000)) || ((y & 0x80000000) != 0x80000000)) {
                res = 1;
            }
            return (res, pfpsf);
        } else {
            res = 0;
            return (res, pfpsf);
        }
    } else if ((y & 0x78000000) == 0x78000000) {
        res = 0;
        if ((y & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((x & 0x60000000) == 0x60000000) {
        exp_x = ((go_checked_shr_u32((x & 0x1fe00000), go_shift_count_u64((21) as u64))) as i64);
        sig_x = ((x & 0x1fffff) | 0x800000);
        if (sig_x > 9999999) {
            non_canon_x = 1;
        } else {
            non_canon_x = 0;
        }
    } else {
        exp_x = ((go_checked_shr_u32((x & 0x7f800000), go_shift_count_u64((23) as u64))) as i64);
        sig_x = (x & 0x7fffff);
        non_canon_x = 0;
    }
    if ((y & 0x60000000) == 0x60000000) {
        exp_y = ((go_checked_shr_u32((y & 0x1fe00000), go_shift_count_u64((21) as u64))) as i64);
        sig_y = ((y & 0x1fffff) | 0x800000);
        if (sig_y > 9999999) {
            non_canon_y = 1;
        } else {
            non_canon_y = 0;
        }
    } else {
        exp_y = ((go_checked_shr_u32((y & 0x7f800000), go_shift_count_u64((23) as u64))) as i64);
        sig_y = (y & 0x7fffff);
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
        if ((y & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    } else if (y_is_zero != 0) {
        res = 0;
        if ((x & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((((x ^ y) & 0x80000000)) == 0x80000000) {
        res = 0;
        if ((y & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((sig_x > sig_y) && (exp_x >= exp_y)) {
        res = 0;
        if ((x & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((sig_x < sig_y) && (exp_x <= exp_y)) {
        res = 0;
        if ((x & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((exp_x.wrapping_sub(exp_y)) > 6) {
        res = 0;
        if ((x & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((exp_y.wrapping_sub(exp_x)) > 6) {
        res = 0;
        if ((x & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (exp_x > exp_y) {
        sig_n_prime = ((sig_x as u64).wrapping_mul(bid32_mult_factor[(exp_x.wrapping_sub(exp_y)) as usize]));
        if ((0 == 0) && ((sig_n_prime == (sig_y as u64)))) {
            res = 0;
            return (res, pfpsf);
        }
        res = 0;
        if (((sig_n_prime < (sig_y as u64))) != (((x & 0x80000000) == 0x80000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    sig_n_prime = ((sig_y as u64).wrapping_mul(bid32_mult_factor[(exp_y.wrapping_sub(exp_x)) as usize]));
    if ((0 == 0) && ((sig_n_prime == (sig_x as u64)))) {
        res = 0;
        return (res, pfpsf);
    }
    res = 0;
    if ((((sig_x as u64) < sig_n_prime)) != (((x & 0x80000000) == 0x80000000))) {
        res = 1;
    }
    return (res, pfpsf);
}

pub fn bid32_quiet_not_equal(mut x: u32, mut y: u32) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut exp_t: i64 = 0;
    let mut sig_x: u32 = 0;
    let mut sig_y: u32 = 0;
    let mut sig_t: u32 = 0;
    let mut pfpsf: u32 = 0;
    let mut x_is_zero: i64 = 0;
    let mut y_is_zero: i64 = 0;
    let mut non_canon_x: i64 = 0;
    let mut non_canon_y: i64 = 0;
    let mut lcv: i64 = 0;
    if ((((x & 0x7c000000) == 0x7c000000)) || (((y & 0x7c000000) == 0x7c000000))) {
        if (((x & 0x7e000000) == 0x7e000000) || ((y & 0x7e000000) == 0x7e000000)) {
            pfpsf |= 1;
        }
        res = 1;
        return (res, pfpsf);
    }
    if (x == y) {
        res = 0;
        return (res, pfpsf);
    }
    if ((((x & 0x78000000) == 0x78000000)) && (((y & 0x78000000) == 0x78000000))) {
        res = 0;
        if ((((x ^ y) & 0x80000000)) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((((x & 0x78000000) == 0x78000000)) || (((y & 0x78000000) == 0x78000000))) {
        res = 1;
        return (res, pfpsf);
    }
    if ((x & 0x60000000) == 0x60000000) {
        exp_x = ((go_checked_shr_u32((x & 0x1fe00000), go_shift_count_u64((21) as u64))) as i64);
        sig_x = ((x & 0x1fffff) | 0x800000);
        if (sig_x > 9999999) {
            non_canon_x = 1;
        } else {
            non_canon_x = 0;
        }
    } else {
        exp_x = ((go_checked_shr_u32((x & 0x7f800000), go_shift_count_u64((23) as u64))) as i64);
        sig_x = (x & 0x7fffff);
        non_canon_x = 0;
    }
    if ((y & 0x60000000) == 0x60000000) {
        exp_y = ((go_checked_shr_u32((y & 0x1fe00000), go_shift_count_u64((21) as u64))) as i64);
        sig_y = ((y & 0x1fffff) | 0x800000);
        if (sig_y > 9999999) {
            non_canon_y = 1;
        } else {
            non_canon_y = 0;
        }
    } else {
        exp_y = ((go_checked_shr_u32((y & 0x7f800000), go_shift_count_u64((23) as u64))) as i64);
        sig_y = (y & 0x7fffff);
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
    if ((((x ^ y) & 0x80000000)) != 0) {
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
    if ((exp_y.wrapping_sub(exp_x)) > 6) {
        res = 1;
        return (res, pfpsf);
    }
    lcv = 0;
    while (lcv < ((exp_y.wrapping_sub(exp_x)))) {
        sig_y = (sig_y.wrapping_mul(10));
        if (sig_y > 9999999) {
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

pub fn bid32_quiet_not_greater(mut x: u32, mut y: u32) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut sig_x: u32 = 0;
    let mut sig_y: u32 = 0;
    let mut sig_n_prime: u64 = 0;
    let mut pfpsf: u32 = 0;
    let mut x_is_zero: i64 = 0;
    let mut y_is_zero: i64 = 0;
    let mut non_canon_x: i64 = 0;
    let mut non_canon_y: i64 = 0;
    if ((((x & 0x7c000000) == 0x7c000000)) || (((y & 0x7c000000) == 0x7c000000))) {
        if (((x & 0x7e000000) == 0x7e000000) || ((y & 0x7e000000) == 0x7e000000)) {
            pfpsf |= 1;
        }
        res = 1;
        return (res, pfpsf);
    }
    if (x == y) {
        res = 1;
        return (res, pfpsf);
    }
    if ((x & 0x78000000) == 0x78000000) {
        if ((x & 0x80000000) == 0x80000000) {
            res = 1;
            return (res, pfpsf);
        }
        res = 1;
        if ((((y & 0x78000000) != 0x78000000)) || (((y & 0x80000000) == 0x80000000))) {
            res = 0;
        }
        return (res, pfpsf);
    } else if ((y & 0x78000000) == 0x78000000) {
        res = 0;
        if ((y & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((x & 0x60000000) == 0x60000000) {
        exp_x = ((go_checked_shr_u32((x & 0x1fe00000), go_shift_count_u64((21) as u64))) as i64);
        sig_x = ((x & 0x1fffff) | 0x800000);
        if (sig_x > 9999999) {
            non_canon_x = 1;
        } else {
            non_canon_x = 0;
        }
    } else {
        exp_x = ((go_checked_shr_u32((x & 0x7f800000), go_shift_count_u64((23) as u64))) as i64);
        sig_x = (x & 0x7fffff);
        non_canon_x = 0;
    }
    if ((y & 0x60000000) == 0x60000000) {
        exp_y = ((go_checked_shr_u32((y & 0x1fe00000), go_shift_count_u64((21) as u64))) as i64);
        sig_y = ((y & 0x1fffff) | 0x800000);
        if (sig_y > 9999999) {
            non_canon_y = 1;
        } else {
            non_canon_y = 0;
        }
    } else {
        exp_y = ((go_checked_shr_u32((y & 0x7f800000), go_shift_count_u64((23) as u64))) as i64);
        sig_y = (y & 0x7fffff);
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
        if ((y & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (y_is_zero != 0) {
        res = 0;
        if ((x & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((((x ^ y) & 0x80000000)) == 0x80000000) {
        res = 0;
        if ((y & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((sig_x > sig_y) && (exp_x >= exp_y)) {
        res = 0;
        if ((x & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((sig_x < sig_y) && (exp_x <= exp_y)) {
        res = 0;
        if ((x & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((exp_x.wrapping_sub(exp_y)) > 6) {
        res = 0;
        if ((x & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((exp_y.wrapping_sub(exp_x)) > 6) {
        res = 0;
        if ((x & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (exp_x > exp_y) {
        sig_n_prime = ((sig_x as u64).wrapping_mul(bid32_mult_factor[(exp_x.wrapping_sub(exp_y)) as usize]));
        if ((0 == 0) && ((sig_n_prime == (sig_y as u64)))) {
            res = 1;
            return (res, pfpsf);
        }
        res = 0;
        if (((sig_n_prime < (sig_y as u64))) != (((x & 0x80000000) == 0x80000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    sig_n_prime = ((sig_y as u64).wrapping_mul(bid32_mult_factor[(exp_y.wrapping_sub(exp_x)) as usize]));
    if ((0 == 0) && ((sig_n_prime == (sig_x as u64)))) {
        res = 1;
        return (res, pfpsf);
    }
    res = 0;
    if ((((sig_x as u64) < sig_n_prime)) != (((x & 0x80000000) == 0x80000000))) {
        res = 1;
    }
    return (res, pfpsf);
}

pub fn bid32_quiet_not_less(mut x: u32, mut y: u32) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut sig_x: u32 = 0;
    let mut sig_y: u32 = 0;
    let mut sig_n_prime: u64 = 0;
    let mut pfpsf: u32 = 0;
    let mut x_is_zero: i64 = 0;
    let mut y_is_zero: i64 = 0;
    let mut non_canon_x: i64 = 0;
    let mut non_canon_y: i64 = 0;
    if ((((x & 0x7c000000) == 0x7c000000)) || (((y & 0x7c000000) == 0x7c000000))) {
        if (((x & 0x7e000000) == 0x7e000000) || ((y & 0x7e000000) == 0x7e000000)) {
            pfpsf |= 1;
        }
        res = 1;
        return (res, pfpsf);
    }
    if (x == y) {
        res = 1;
        return (res, pfpsf);
    }
    if ((x & 0x78000000) == 0x78000000) {
        if ((x & 0x80000000) == 0x80000000) {
            res = 0;
            if ((((y & 0x78000000) == 0x78000000)) && ((y & 0x80000000) == 0x80000000)) {
                res = 1;
            }
            return (res, pfpsf);
        }
        res = 1;
        return (res, pfpsf);
    } else if ((y & 0x78000000) == 0x78000000) {
        res = 0;
        if ((y & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((x & 0x60000000) == 0x60000000) {
        exp_x = ((go_checked_shr_u32((x & 0x1fe00000), go_shift_count_u64((21) as u64))) as i64);
        sig_x = ((x & 0x1fffff) | 0x800000);
        if (sig_x > 9999999) {
            non_canon_x = 1;
        } else {
            non_canon_x = 0;
        }
    } else {
        exp_x = ((go_checked_shr_u32((x & 0x7f800000), go_shift_count_u64((23) as u64))) as i64);
        sig_x = (x & 0x7fffff);
        non_canon_x = 0;
    }
    if ((y & 0x60000000) == 0x60000000) {
        exp_y = ((go_checked_shr_u32((y & 0x1fe00000), go_shift_count_u64((21) as u64))) as i64);
        sig_y = ((y & 0x1fffff) | 0x800000);
        if (sig_y > 9999999) {
            non_canon_y = 1;
        } else {
            non_canon_y = 0;
        }
    } else {
        exp_y = ((go_checked_shr_u32((y & 0x7f800000), go_shift_count_u64((23) as u64))) as i64);
        sig_y = (y & 0x7fffff);
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
        if ((y & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (y_is_zero != 0) {
        res = 0;
        if ((x & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((((x ^ y) & 0x80000000)) == 0x80000000) {
        res = 0;
        if ((y & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((sig_x > sig_y) && (exp_x >= exp_y)) {
        res = 0;
        if ((x & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((sig_x < sig_y) && (exp_x <= exp_y)) {
        res = 0;
        if ((x & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((exp_x.wrapping_sub(exp_y)) > 6) {
        res = 0;
        if ((x & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((exp_y.wrapping_sub(exp_x)) > 6) {
        res = 0;
        if ((x & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (exp_x > exp_y) {
        sig_n_prime = ((sig_x as u64).wrapping_mul(bid32_mult_factor[(exp_x.wrapping_sub(exp_y)) as usize]));
        if ((0 == 0) && ((sig_n_prime == (sig_y as u64)))) {
            res = 1;
            return (res, pfpsf);
        }
        res = 0;
        if (((sig_n_prime < (sig_y as u64))) != (((x & 0x80000000) != 0x80000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    sig_n_prime = ((sig_y as u64).wrapping_mul(bid32_mult_factor[(exp_y.wrapping_sub(exp_x)) as usize]));
    if ((0 == 0) && ((sig_n_prime == (sig_x as u64)))) {
        res = 1;
        return (res, pfpsf);
    }
    res = 0;
    if ((((sig_x as u64) < sig_n_prime)) != (((x & 0x80000000) != 0x80000000))) {
        res = 1;
    }
    return (res, pfpsf);
}

pub fn bid32_quiet_ordered(mut x: u32, mut y: u32) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut pfpsf: u32 = 0;
    if ((((x & 0x7c000000) == 0x7c000000)) || (((y & 0x7c000000) == 0x7c000000))) {
        if (((x & 0x7e000000) == 0x7e000000) || ((y & 0x7e000000) == 0x7e000000)) {
            pfpsf |= 1;
        }
        res = 0;
        return (res, pfpsf);
    }
    res = 1;
    return (res, pfpsf);
}

pub fn bid32_quiet_unordered(mut x: u32, mut y: u32) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut pfpsf: u32 = 0;
    if ((((x & 0x7c000000) == 0x7c000000)) || (((y & 0x7c000000) == 0x7c000000))) {
        if (((x & 0x7e000000) == 0x7e000000) || ((y & 0x7e000000) == 0x7e000000)) {
            pfpsf |= 1;
        }
        res = 1;
        return (res, pfpsf);
    }
    res = 0;
    return (res, pfpsf);
}

pub fn bid32_signaling_greater(mut x: u32, mut y: u32) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut sig_x: u32 = 0;
    let mut sig_y: u32 = 0;
    let mut sig_n_prime: u64 = 0;
    let mut pfpsf: u32 = 0;
    let mut x_is_zero: i64 = 0;
    let mut y_is_zero: i64 = 0;
    let mut non_canon_x: i64 = 0;
    let mut non_canon_y: i64 = 0;
    if ((((x & 0x7c000000) == 0x7c000000)) || (((y & 0x7c000000) == 0x7c000000))) {
        pfpsf |= 1;
        res = 0;
        return (res, pfpsf);
    }
    if (x == y) {
        res = 0;
        return (res, pfpsf);
    }
    if ((x & 0x78000000) == 0x78000000) {
        if ((x & 0x80000000) == 0x80000000) {
            res = 0;
            return (res, pfpsf);
        }
        res = 0;
        if ((((y & 0x78000000) != 0x78000000)) || (((y & 0x80000000) == 0x80000000))) {
            res = 1;
        }
        return (res, pfpsf);
    } else if ((y & 0x78000000) == 0x78000000) {
        res = 0;
        if ((y & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((x & 0x60000000) == 0x60000000) {
        exp_x = ((go_checked_shr_u32((x & 0x1fe00000), go_shift_count_u64((21) as u64))) as i64);
        sig_x = ((x & 0x1fffff) | 0x800000);
        if (sig_x > 9999999) {
            non_canon_x = 1;
        } else {
            non_canon_x = 0;
        }
    } else {
        exp_x = ((go_checked_shr_u32((x & 0x7f800000), go_shift_count_u64((23) as u64))) as i64);
        sig_x = (x & 0x7fffff);
        non_canon_x = 0;
    }
    if ((y & 0x60000000) == 0x60000000) {
        exp_y = ((go_checked_shr_u32((y & 0x1fe00000), go_shift_count_u64((21) as u64))) as i64);
        sig_y = ((y & 0x1fffff) | 0x800000);
        if (sig_y > 9999999) {
            non_canon_y = 1;
        } else {
            non_canon_y = 0;
        }
    } else {
        exp_y = ((go_checked_shr_u32((y & 0x7f800000), go_shift_count_u64((23) as u64))) as i64);
        sig_y = (y & 0x7fffff);
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
        if ((y & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (y_is_zero != 0) {
        res = 0;
        if ((x & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((((x ^ y) & 0x80000000)) == 0x80000000) {
        res = 0;
        if ((y & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((sig_x > sig_y) && (exp_x >= exp_y)) {
        res = 0;
        if ((x & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((sig_x < sig_y) && (exp_x <= exp_y)) {
        res = 0;
        if ((x & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((exp_x.wrapping_sub(exp_y)) > 6) {
        res = 0;
        if ((x & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((exp_y.wrapping_sub(exp_x)) > 6) {
        res = 0;
        if ((x & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (exp_x > exp_y) {
        sig_n_prime = ((sig_x as u64).wrapping_mul(bid32_mult_factor[(exp_x.wrapping_sub(exp_y)) as usize]));
        if ((0 == 0) && ((sig_n_prime == (sig_y as u64)))) {
            res = 0;
            return (res, pfpsf);
        }
        res = 0;
        if ((((0 > 0) || (sig_n_prime > (sig_y as u64)))) != (((x & 0x80000000) == 0x80000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    sig_n_prime = ((sig_y as u64).wrapping_mul(bid32_mult_factor[(exp_y.wrapping_sub(exp_x)) as usize]));
    if ((0 == 0) && ((sig_n_prime == (sig_x as u64)))) {
        res = 0;
        return (res, pfpsf);
    }
    res = 0;
    if ((((sig_x as u64) > sig_n_prime)) != (((x & 0x80000000) == 0x80000000))) {
        res = 1;
    }
    return (res, pfpsf);
}

pub fn bid32_signaling_greater_equal(mut x: u32, mut y: u32) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut sig_x: u32 = 0;
    let mut sig_y: u32 = 0;
    let mut sig_n_prime: u64 = 0;
    let mut pfpsf: u32 = 0;
    let mut x_is_zero: i64 = 0;
    let mut y_is_zero: i64 = 0;
    let mut non_canon_x: i64 = 0;
    let mut non_canon_y: i64 = 0;
    if ((((x & 0x7c000000) == 0x7c000000)) || (((y & 0x7c000000) == 0x7c000000))) {
        pfpsf |= 1;
        res = 0;
        return (res, pfpsf);
    }
    if (x == y) {
        res = 1;
        return (res, pfpsf);
    }
    if ((x & 0x78000000) == 0x78000000) {
        if ((x & 0x80000000) == 0x80000000) {
            res = 0;
            if ((((y & 0x78000000) == 0x78000000)) && ((y & 0x80000000) == 0x80000000)) {
                res = 1;
            }
            return (res, pfpsf);
        }
        res = 1;
        return (res, pfpsf);
    } else if ((y & 0x78000000) == 0x78000000) {
        res = 0;
        if ((y & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((x & 0x60000000) == 0x60000000) {
        exp_x = ((go_checked_shr_u32((x & 0x1fe00000), go_shift_count_u64((21) as u64))) as i64);
        sig_x = ((x & 0x1fffff) | 0x800000);
        if (sig_x > 9999999) {
            non_canon_x = 1;
        } else {
            non_canon_x = 0;
        }
    } else {
        exp_x = ((go_checked_shr_u32((x & 0x7f800000), go_shift_count_u64((23) as u64))) as i64);
        sig_x = (x & 0x7fffff);
        non_canon_x = 0;
    }
    if ((y & 0x60000000) == 0x60000000) {
        exp_y = ((go_checked_shr_u32((y & 0x1fe00000), go_shift_count_u64((21) as u64))) as i64);
        sig_y = ((y & 0x1fffff) | 0x800000);
        if (sig_y > 9999999) {
            non_canon_y = 1;
        } else {
            non_canon_y = 0;
        }
    } else {
        exp_y = ((go_checked_shr_u32((y & 0x7f800000), go_shift_count_u64((23) as u64))) as i64);
        sig_y = (y & 0x7fffff);
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
        if ((y & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (y_is_zero != 0) {
        res = 0;
        if ((x & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((((x ^ y) & 0x80000000)) == 0x80000000) {
        res = 0;
        if ((y & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((sig_x > sig_y) && (exp_x >= exp_y)) {
        res = 0;
        if ((x & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((sig_x < sig_y) && (exp_x <= exp_y)) {
        res = 0;
        if ((x & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((exp_x.wrapping_sub(exp_y)) > 6) {
        res = 0;
        if ((x & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((exp_y.wrapping_sub(exp_x)) > 6) {
        res = 0;
        if ((x & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (exp_x > exp_y) {
        sig_n_prime = ((sig_x as u64).wrapping_mul(bid32_mult_factor[(exp_x.wrapping_sub(exp_y)) as usize]));
        if ((0 == 0) && ((sig_n_prime == (sig_y as u64)))) {
            res = 1;
            return (res, pfpsf);
        }
        res = 0;
        if (((sig_n_prime < (sig_y as u64))) != (((x & 0x80000000) != 0x80000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    sig_n_prime = ((sig_y as u64).wrapping_mul(bid32_mult_factor[(exp_y.wrapping_sub(exp_x)) as usize]));
    if ((0 == 0) && ((sig_n_prime == (sig_x as u64)))) {
        res = 1;
        return (res, pfpsf);
    }
    res = 0;
    if ((((sig_x as u64) < sig_n_prime)) != (((x & 0x80000000) != 0x80000000))) {
        res = 1;
    }
    return (res, pfpsf);
}

pub fn bid32_signaling_greater_unordered(mut x: u32, mut y: u32) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut sig_x: u32 = 0;
    let mut sig_y: u32 = 0;
    let mut sig_n_prime: u64 = 0;
    let mut pfpsf: u32 = 0;
    let mut x_is_zero: i64 = 0;
    let mut y_is_zero: i64 = 0;
    let mut non_canon_x: i64 = 0;
    let mut non_canon_y: i64 = 0;
    if ((((x & 0x7c000000) == 0x7c000000)) || (((y & 0x7c000000) == 0x7c000000))) {
        pfpsf |= 1;
        res = 1;
        return (res, pfpsf);
    }
    if (x == y) {
        res = 0;
        return (res, pfpsf);
    }
    if ((x & 0x78000000) == 0x78000000) {
        if ((x & 0x80000000) == 0x80000000) {
            res = 0;
            return (res, pfpsf);
        }
        res = 0;
        if ((((y & 0x78000000) != 0x78000000)) || (((y & 0x80000000) == 0x80000000))) {
            res = 1;
        }
        return (res, pfpsf);
    } else if ((y & 0x78000000) == 0x78000000) {
        res = 0;
        if ((y & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((x & 0x60000000) == 0x60000000) {
        exp_x = ((go_checked_shr_u32((x & 0x1fe00000), go_shift_count_u64((21) as u64))) as i64);
        sig_x = ((x & 0x1fffff) | 0x800000);
        if (sig_x > 9999999) {
            non_canon_x = 1;
        } else {
            non_canon_x = 0;
        }
    } else {
        exp_x = ((go_checked_shr_u32((x & 0x7f800000), go_shift_count_u64((23) as u64))) as i64);
        sig_x = (x & 0x7fffff);
        non_canon_x = 0;
    }
    if ((y & 0x60000000) == 0x60000000) {
        exp_y = ((go_checked_shr_u32((y & 0x1fe00000), go_shift_count_u64((21) as u64))) as i64);
        sig_y = ((y & 0x1fffff) | 0x800000);
        if (sig_y > 9999999) {
            non_canon_y = 1;
        } else {
            non_canon_y = 0;
        }
    } else {
        exp_y = ((go_checked_shr_u32((y & 0x7f800000), go_shift_count_u64((23) as u64))) as i64);
        sig_y = (y & 0x7fffff);
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
        if ((y & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (y_is_zero != 0) {
        res = 0;
        if ((x & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((((x ^ y) & 0x80000000)) == 0x80000000) {
        res = 0;
        if ((y & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((sig_x > sig_y) && (exp_x >= exp_y)) {
        res = 0;
        if ((x & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((sig_x < sig_y) && (exp_x <= exp_y)) {
        res = 0;
        if ((x & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((exp_x.wrapping_sub(exp_y)) > 6) {
        res = 0;
        if ((x & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((exp_y.wrapping_sub(exp_x)) > 6) {
        res = 0;
        if ((x & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (exp_x > exp_y) {
        sig_n_prime = ((sig_x as u64).wrapping_mul(bid32_mult_factor[(exp_x.wrapping_sub(exp_y)) as usize]));
        if ((0 == 0) && ((sig_n_prime == (sig_y as u64)))) {
            res = 0;
            return (res, pfpsf);
        }
        res = 0;
        if ((((0 > 0) || (sig_n_prime > (sig_y as u64)))) != (((x & 0x80000000) == 0x80000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    sig_n_prime = ((sig_y as u64).wrapping_mul(bid32_mult_factor[(exp_y.wrapping_sub(exp_x)) as usize]));
    if ((0 == 0) && ((sig_n_prime == (sig_x as u64)))) {
        res = 0;
        return (res, pfpsf);
    }
    res = 0;
    if ((((sig_x as u64) > sig_n_prime)) != (((x & 0x80000000) == 0x80000000))) {
        res = 1;
    }
    return (res, pfpsf);
}

pub fn bid32_signaling_less_equal(mut x: u32, mut y: u32) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut sig_x: u32 = 0;
    let mut sig_y: u32 = 0;
    let mut sig_n_prime: u64 = 0;
    let mut pfpsf: u32 = 0;
    let mut x_is_zero: i64 = 0;
    let mut y_is_zero: i64 = 0;
    let mut non_canon_x: i64 = 0;
    let mut non_canon_y: i64 = 0;
    if ((((x & 0x7c000000) == 0x7c000000)) || (((y & 0x7c000000) == 0x7c000000))) {
        pfpsf |= 1;
        res = 0;
        return (res, pfpsf);
    }
    if (x == y) {
        res = 1;
        return (res, pfpsf);
    }
    if ((x & 0x78000000) == 0x78000000) {
        if ((x & 0x80000000) == 0x80000000) {
            res = 1;
            return (res, pfpsf);
        }
        res = 1;
        if ((((y & 0x78000000) != 0x78000000)) || (((y & 0x80000000) == 0x80000000))) {
            res = 0;
        }
        return (res, pfpsf);
    } else if ((y & 0x78000000) == 0x78000000) {
        res = 0;
        if ((y & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((x & 0x60000000) == 0x60000000) {
        exp_x = ((go_checked_shr_u32((x & 0x1fe00000), go_shift_count_u64((21) as u64))) as i64);
        sig_x = ((x & 0x1fffff) | 0x800000);
        if (sig_x > 9999999) {
            non_canon_x = 1;
        } else {
            non_canon_x = 0;
        }
    } else {
        exp_x = ((go_checked_shr_u32((x & 0x7f800000), go_shift_count_u64((23) as u64))) as i64);
        sig_x = (x & 0x7fffff);
        non_canon_x = 0;
    }
    if ((y & 0x60000000) == 0x60000000) {
        exp_y = ((go_checked_shr_u32((y & 0x1fe00000), go_shift_count_u64((21) as u64))) as i64);
        sig_y = ((y & 0x1fffff) | 0x800000);
        if (sig_y > 9999999) {
            non_canon_y = 1;
        } else {
            non_canon_y = 0;
        }
    } else {
        exp_y = ((go_checked_shr_u32((y & 0x7f800000), go_shift_count_u64((23) as u64))) as i64);
        sig_y = (y & 0x7fffff);
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
        if ((y & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (y_is_zero != 0) {
        res = 0;
        if ((x & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((((x ^ y) & 0x80000000)) == 0x80000000) {
        res = 0;
        if ((y & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((sig_x > sig_y) && (exp_x >= exp_y)) {
        res = 0;
        if ((x & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((sig_x < sig_y) && (exp_x <= exp_y)) {
        res = 0;
        if ((x & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((exp_x.wrapping_sub(exp_y)) > 6) {
        res = 0;
        if ((x & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((exp_y.wrapping_sub(exp_x)) > 6) {
        res = 0;
        if ((x & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (exp_x > exp_y) {
        sig_n_prime = ((sig_x as u64).wrapping_mul(bid32_mult_factor[(exp_x.wrapping_sub(exp_y)) as usize]));
        if ((0 == 0) && ((sig_n_prime == (sig_y as u64)))) {
            res = 1;
            return (res, pfpsf);
        }
        res = 0;
        if (((sig_n_prime < (sig_y as u64))) != (((x & 0x80000000) == 0x80000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    sig_n_prime = ((sig_y as u64).wrapping_mul(bid32_mult_factor[(exp_y.wrapping_sub(exp_x)) as usize]));
    if ((0 == 0) && ((sig_n_prime == (sig_x as u64)))) {
        res = 1;
        return (res, pfpsf);
    }
    res = 0;
    if ((((sig_x as u64) < sig_n_prime)) != (((x & 0x80000000) == 0x80000000))) {
        res = 1;
    }
    return (res, pfpsf);
}

pub fn bid32_signaling_less_unordered(mut x: u32, mut y: u32) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut sig_x: u32 = 0;
    let mut sig_y: u32 = 0;
    let mut sig_n_prime: u64 = 0;
    let mut pfpsf: u32 = 0;
    let mut x_is_zero: i64 = 0;
    let mut y_is_zero: i64 = 0;
    let mut non_canon_x: i64 = 0;
    let mut non_canon_y: i64 = 0;
    if ((((x & 0x7c000000) == 0x7c000000)) || (((y & 0x7c000000) == 0x7c000000))) {
        pfpsf |= 1;
        res = 1;
        return (res, pfpsf);
    }
    if (x == y) {
        res = 0;
        return (res, pfpsf);
    }
    if ((x & 0x78000000) == 0x78000000) {
        if ((x & 0x80000000) == 0x80000000) {
            res = 0;
            if ((((y & 0x78000000) != 0x78000000)) || ((y & 0x80000000) != 0x80000000)) {
                res = 1;
            }
            return (res, pfpsf);
        }
        res = 0;
        return (res, pfpsf);
    } else if ((y & 0x78000000) == 0x78000000) {
        res = 0;
        if ((y & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((x & 0x60000000) == 0x60000000) {
        exp_x = ((go_checked_shr_u32((x & 0x1fe00000), go_shift_count_u64((21) as u64))) as i64);
        sig_x = ((x & 0x1fffff) | 0x800000);
        if (sig_x > 9999999) {
            non_canon_x = 1;
        } else {
            non_canon_x = 0;
        }
    } else {
        exp_x = ((go_checked_shr_u32((x & 0x7f800000), go_shift_count_u64((23) as u64))) as i64);
        sig_x = (x & 0x7fffff);
        non_canon_x = 0;
    }
    if ((y & 0x60000000) == 0x60000000) {
        exp_y = ((go_checked_shr_u32((y & 0x1fe00000), go_shift_count_u64((21) as u64))) as i64);
        sig_y = ((y & 0x1fffff) | 0x800000);
        if (sig_y > 9999999) {
            non_canon_y = 1;
        } else {
            non_canon_y = 0;
        }
    } else {
        exp_y = ((go_checked_shr_u32((y & 0x7f800000), go_shift_count_u64((23) as u64))) as i64);
        sig_y = (y & 0x7fffff);
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
        if ((y & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (y_is_zero != 0) {
        res = 0;
        if ((x & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((((x ^ y) & 0x80000000)) == 0x80000000) {
        res = 0;
        if ((y & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((sig_x > sig_y) && (exp_x >= exp_y)) {
        res = 0;
        if ((x & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((sig_x < sig_y) && (exp_x <= exp_y)) {
        res = 0;
        if ((x & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((exp_x.wrapping_sub(exp_y)) > 6) {
        res = 0;
        if ((x & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((exp_y.wrapping_sub(exp_x)) > 6) {
        res = 0;
        if ((x & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (exp_x > exp_y) {
        sig_n_prime = ((sig_x as u64).wrapping_mul(bid32_mult_factor[(exp_x.wrapping_sub(exp_y)) as usize]));
        if ((0 == 0) && ((sig_n_prime == (sig_y as u64)))) {
            res = 0;
            return (res, pfpsf);
        }
        res = 0;
        if (((sig_n_prime < (sig_y as u64))) != (((x & 0x80000000) == 0x80000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    sig_n_prime = ((sig_y as u64).wrapping_mul(bid32_mult_factor[(exp_y.wrapping_sub(exp_x)) as usize]));
    if ((0 == 0) && ((sig_n_prime == (sig_x as u64)))) {
        res = 0;
        return (res, pfpsf);
    }
    res = 0;
    if ((((sig_x as u64) < sig_n_prime)) != (((x & 0x80000000) == 0x80000000))) {
        res = 1;
    }
    return (res, pfpsf);
}

pub fn bid32_signaling_not_greater(mut x: u32, mut y: u32) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut sig_x: u32 = 0;
    let mut sig_y: u32 = 0;
    let mut sig_n_prime: u64 = 0;
    let mut pfpsf: u32 = 0;
    let mut x_is_zero: i64 = 0;
    let mut y_is_zero: i64 = 0;
    let mut non_canon_x: i64 = 0;
    let mut non_canon_y: i64 = 0;
    if ((((x & 0x7c000000) == 0x7c000000)) || (((y & 0x7c000000) == 0x7c000000))) {
        pfpsf |= 1;
        res = 1;
        return (res, pfpsf);
    }
    if (x == y) {
        res = 1;
        return (res, pfpsf);
    }
    if ((x & 0x78000000) == 0x78000000) {
        if ((x & 0x80000000) == 0x80000000) {
            res = 1;
            return (res, pfpsf);
        }
        res = 1;
        if ((((y & 0x78000000) != 0x78000000)) || (((y & 0x80000000) == 0x80000000))) {
            res = 0;
        }
        return (res, pfpsf);
    } else if ((y & 0x78000000) == 0x78000000) {
        res = 0;
        if ((y & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((x & 0x60000000) == 0x60000000) {
        exp_x = ((go_checked_shr_u32((x & 0x1fe00000), go_shift_count_u64((21) as u64))) as i64);
        sig_x = ((x & 0x1fffff) | 0x800000);
        if (sig_x > 9999999) {
            non_canon_x = 1;
        } else {
            non_canon_x = 0;
        }
    } else {
        exp_x = ((go_checked_shr_u32((x & 0x7f800000), go_shift_count_u64((23) as u64))) as i64);
        sig_x = (x & 0x7fffff);
        non_canon_x = 0;
    }
    if ((y & 0x60000000) == 0x60000000) {
        exp_y = ((go_checked_shr_u32((y & 0x1fe00000), go_shift_count_u64((21) as u64))) as i64);
        sig_y = ((y & 0x1fffff) | 0x800000);
        if (sig_y > 9999999) {
            non_canon_y = 1;
        } else {
            non_canon_y = 0;
        }
    } else {
        exp_y = ((go_checked_shr_u32((y & 0x7f800000), go_shift_count_u64((23) as u64))) as i64);
        sig_y = (y & 0x7fffff);
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
        if ((y & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (y_is_zero != 0) {
        res = 0;
        if ((x & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((((x ^ y) & 0x80000000)) == 0x80000000) {
        res = 0;
        if ((y & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((sig_x > sig_y) && (exp_x >= exp_y)) {
        res = 0;
        if ((x & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((sig_x < sig_y) && (exp_x <= exp_y)) {
        res = 0;
        if ((x & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((exp_x.wrapping_sub(exp_y)) > 6) {
        res = 0;
        if ((x & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((exp_y.wrapping_sub(exp_x)) > 6) {
        res = 0;
        if ((x & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (exp_x > exp_y) {
        sig_n_prime = ((sig_x as u64).wrapping_mul(bid32_mult_factor[(exp_x.wrapping_sub(exp_y)) as usize]));
        if ((0 == 0) && ((sig_n_prime == (sig_y as u64)))) {
            res = 1;
            return (res, pfpsf);
        }
        res = 0;
        if (((sig_n_prime < (sig_y as u64))) != (((x & 0x80000000) == 0x80000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    sig_n_prime = ((sig_y as u64).wrapping_mul(bid32_mult_factor[(exp_y.wrapping_sub(exp_x)) as usize]));
    if ((0 == 0) && ((sig_n_prime == (sig_x as u64)))) {
        res = 1;
        return (res, pfpsf);
    }
    res = 0;
    if ((((sig_x as u64) < sig_n_prime)) != (((x & 0x80000000) == 0x80000000))) {
        res = 1;
    }
    return (res, pfpsf);
}

pub fn bid32_signaling_not_less(mut x: u32, mut y: u32) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut sig_x: u32 = 0;
    let mut sig_y: u32 = 0;
    let mut sig_n_prime: u64 = 0;
    let mut pfpsf: u32 = 0;
    let mut x_is_zero: i64 = 0;
    let mut y_is_zero: i64 = 0;
    let mut non_canon_x: i64 = 0;
    let mut non_canon_y: i64 = 0;
    if ((((x & 0x7c000000) == 0x7c000000)) || (((y & 0x7c000000) == 0x7c000000))) {
        pfpsf |= 1;
        res = 1;
        return (res, pfpsf);
    }
    if (x == y) {
        res = 1;
        return (res, pfpsf);
    }
    if ((x & 0x78000000) == 0x78000000) {
        if ((x & 0x80000000) == 0x80000000) {
            res = 0;
            if ((((y & 0x78000000) == 0x78000000)) && ((y & 0x80000000) == 0x80000000)) {
                res = 1;
            }
            return (res, pfpsf);
        }
        res = 1;
        return (res, pfpsf);
    } else if ((y & 0x78000000) == 0x78000000) {
        res = 0;
        if ((y & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((x & 0x60000000) == 0x60000000) {
        exp_x = ((go_checked_shr_u32((x & 0x1fe00000), go_shift_count_u64((21) as u64))) as i64);
        sig_x = ((x & 0x1fffff) | 0x800000);
        if (sig_x > 9999999) {
            non_canon_x = 1;
        } else {
            non_canon_x = 0;
        }
    } else {
        exp_x = ((go_checked_shr_u32((x & 0x7f800000), go_shift_count_u64((23) as u64))) as i64);
        sig_x = (x & 0x7fffff);
        non_canon_x = 0;
    }
    if ((y & 0x60000000) == 0x60000000) {
        exp_y = ((go_checked_shr_u32((y & 0x1fe00000), go_shift_count_u64((21) as u64))) as i64);
        sig_y = ((y & 0x1fffff) | 0x800000);
        if (sig_y > 9999999) {
            non_canon_y = 1;
        } else {
            non_canon_y = 0;
        }
    } else {
        exp_y = ((go_checked_shr_u32((y & 0x7f800000), go_shift_count_u64((23) as u64))) as i64);
        sig_y = (y & 0x7fffff);
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
        if ((y & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (y_is_zero != 0) {
        res = 0;
        if ((x & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((((x ^ y) & 0x80000000)) == 0x80000000) {
        res = 0;
        if ((y & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((sig_x > sig_y) && (exp_x >= exp_y)) {
        res = 0;
        if ((x & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((sig_x < sig_y) && (exp_x <= exp_y)) {
        res = 0;
        if ((x & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((exp_x.wrapping_sub(exp_y)) > 6) {
        res = 0;
        if ((x & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((exp_y.wrapping_sub(exp_x)) > 6) {
        res = 0;
        if ((x & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (exp_x > exp_y) {
        sig_n_prime = ((sig_x as u64).wrapping_mul(bid32_mult_factor[(exp_x.wrapping_sub(exp_y)) as usize]));
        if ((0 == 0) && ((sig_n_prime == (sig_y as u64)))) {
            res = 1;
            return (res, pfpsf);
        }
        res = 0;
        if (((sig_n_prime < (sig_y as u64))) != (((x & 0x80000000) != 0x80000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    sig_n_prime = ((sig_y as u64).wrapping_mul(bid32_mult_factor[(exp_y.wrapping_sub(exp_x)) as usize]));
    if ((0 == 0) && ((sig_n_prime == (sig_x as u64)))) {
        res = 1;
        return (res, pfpsf);
    }
    res = 0;
    if ((((sig_x as u64) < sig_n_prime)) != (((x & 0x80000000) != 0x80000000))) {
        res = 1;
    }
    return (res, pfpsf);
}

pub fn bid32_signaling_less(mut x: u32, mut y: u32) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut sig_x: u32 = 0;
    let mut sig_y: u32 = 0;
    let mut sig_n_prime: u64 = 0;
    let mut x_is_zero: i64 = 0;
    let mut y_is_zero: i64 = 0;
    let mut non_canon_x: i64 = 0;
    let mut non_canon_y: i64 = 0;
    let mut pfpsf: u32 = 0;
    if ((((x & 0x7c000000) == 0x7c000000)) || (((y & 0x7c000000) == 0x7c000000))) {
        pfpsf |= 1;
        res = 0;
        return (res, pfpsf);
    }
    if (x == y) {
        res = 0;
        return (res, pfpsf);
    }
    if ((x & 0x78000000) == 0x78000000) {
        if ((x & 0x80000000) == 0x80000000) {
            res = 0;
            if ((((y & 0x78000000) != 0x78000000)) || ((y & 0x80000000) != 0x80000000)) {
                res = 1;
            }
            return (res, pfpsf);
        }
        res = 0;
        return (res, pfpsf);
    } else if ((y & 0x78000000) == 0x78000000) {
        res = 0;
        if ((y & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((x & 0x60000000) == 0x60000000) {
        exp_x = ((go_checked_shr_u32((x & 0x1fe00000), go_shift_count_u64((21) as u64))) as i64);
        sig_x = ((x & 0x1fffff) | 0x800000);
        if (sig_x > 9999999) {
            non_canon_x = 1;
        }
    } else {
        exp_x = ((go_checked_shr_u32((x & 0x7f800000), go_shift_count_u64((23) as u64))) as i64);
        sig_x = (x & 0x7fffff);
    }
    if ((y & 0x60000000) == 0x60000000) {
        exp_y = ((go_checked_shr_u32((y & 0x1fe00000), go_shift_count_u64((21) as u64))) as i64);
        sig_y = ((y & 0x1fffff) | 0x800000);
        if (sig_y > 9999999) {
            non_canon_y = 1;
        }
    } else {
        exp_y = ((go_checked_shr_u32((y & 0x7f800000), go_shift_count_u64((23) as u64))) as i64);
        sig_y = (y & 0x7fffff);
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
        if ((y & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (y_is_zero != 0) {
        res = 0;
        if ((x & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((((x ^ y) & 0x80000000)) == 0x80000000) {
        res = 0;
        if ((y & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((sig_x > sig_y) && (exp_x >= exp_y)) {
        res = 0;
        if ((x & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((sig_x < sig_y) && (exp_x <= exp_y)) {
        res = 0;
        if ((x & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((exp_x.wrapping_sub(exp_y)) > 6) {
        res = 0;
        if ((x & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if ((exp_y.wrapping_sub(exp_x)) > 6) {
        res = 0;
        if ((x & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return (res, pfpsf);
    }
    if (exp_x > exp_y) {
        sig_n_prime = ((sig_x as u64).wrapping_mul(bid32_mult_factor[(exp_x.wrapping_sub(exp_y)) as usize]));
        if (sig_n_prime == (sig_y as u64)) {
            res = 0;
            return (res, pfpsf);
        }
        res = 0;
        if (((sig_n_prime < (sig_y as u64))) != (((x & 0x80000000) == 0x80000000))) {
            res = 1;
        }
        return (res, pfpsf);
    }
    sig_n_prime = ((sig_y as u64).wrapping_mul(bid32_mult_factor[(exp_y.wrapping_sub(exp_x)) as usize]));
    if (sig_n_prime == (sig_x as u64)) {
        res = 0;
        return (res, pfpsf);
    }
    res = 0;
    if ((((sig_x as u64) < sig_n_prime)) != (((x & 0x80000000) == 0x80000000))) {
        res = 1;
    }
    return (res, pfpsf);
}
