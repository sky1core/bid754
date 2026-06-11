// Auto-generated from bid32_minmax.go by go2rs. Do not edit.

use super::prelude::*;

pub(crate) fn bid32_minnum_pure(mut x: u32, mut y: u32) -> u32 {
    let mut res: u32 = 0;
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut sig_x: u32 = 0;
    let mut sig_y: u32 = 0;
    let mut sig_n_prime: u64 = 0;
    let mut x_is_zero: bool = false;
    let mut y_is_zero: bool = false;
    if ((x & 0x7c000000) == 0x7c000000) {
        x = (x & 0xfe0fffff);
        if ((x & 0x000fffff) > 999999) {
            x = (x & 0xfe000000);
        }
    } else if ((x & 0x78000000) == 0x78000000) {
        x = (x & (0x80000000 | 0x78000000));
    } else {
        if ((x & 0x60000000) == 0x60000000) {
            if ((((x & 0x1fffff) | 0x800000)) > 9999999) {
                x = ((x & 0x80000000) | ((go_checked_shl_u32((x & 0x1fe00000), go_shift_count_u64((2) as u64)))));
            }
        }
    }
    if ((y & 0x7c000000) == 0x7c000000) {
        y = (y & 0xfe0fffff);
        if ((y & 0x000fffff) > 999999) {
            y = (y & 0xfe000000);
        }
    } else if ((y & 0x78000000) == 0x78000000) {
        y = (y & (0x80000000 | 0x78000000));
    } else {
        if ((y & 0x60000000) == 0x60000000) {
            if ((((y & 0x1fffff) | 0x800000)) > 9999999) {
                y = ((y & 0x80000000) | ((go_checked_shl_u32((y & 0x1fe00000), go_shift_count_u64((2) as u64)))));
            }
        }
    }
    if ((x & 0x7c000000) == 0x7c000000) {
        if ((x & 0x7e000000) == 0x7e000000) {
            x = (x & 0xfdffffff);
            res = x;
        } else {
            if ((y & 0x7c000000) == 0x7c000000) {
                res = x;
            } else {
                res = y;
            }
        }
        return res;
    } else if ((y & 0x7c000000) == 0x7c000000) {
        if ((y & 0x7e000000) == 0x7e000000) {
            y = (y & 0xfdffffff);
            res = y;
        } else {
            res = x;
        }
        return res;
    }
    if (x == y) {
        return x;
    }
    if ((x & 0x78000000) == 0x78000000) {
        if ((x & 0x80000000) == 0x80000000) {
            return x;
        }
        return y;
    } else if ((y & 0x78000000) == 0x78000000) {
        if ((y & 0x80000000) == 0x80000000) {
            return y;
        }
        return x;
    }
    if ((x & 0x60000000) == 0x60000000) {
        exp_x = ((go_checked_shr_u32((x & 0x1fe00000), go_shift_count_u64((21) as u64))) as i64);
        sig_x = ((x & 0x1fffff) | 0x800000);
    } else {
        exp_x = ((go_checked_shr_u32((x & 0x7f800000), go_shift_count_u64((23) as u64))) as i64);
        sig_x = (x & 0x7fffff);
    }
    if ((y & 0x60000000) == 0x60000000) {
        exp_y = ((go_checked_shr_u32((y & 0x1fe00000), go_shift_count_u64((21) as u64))) as i64);
        sig_y = ((y & 0x1fffff) | 0x800000);
    } else {
        exp_y = ((go_checked_shr_u32((y & 0x7f800000), go_shift_count_u64((23) as u64))) as i64);
        sig_y = (y & 0x7fffff);
    }
    if (sig_x == 0) {
        x_is_zero = true;
    }
    if (sig_y == 0) {
        y_is_zero = true;
    }
    if (x_is_zero && y_is_zero) {
        return y;
    } else if x_is_zero {
        if ((y & 0x80000000) == 0x80000000) {
            return y;
        }
        return x;
    } else if y_is_zero {
        if ((x & 0x80000000) != 0x80000000) {
            return y;
        }
        return x;
    }
    if ((((x ^ y) & 0x80000000)) == 0x80000000) {
        if ((y & 0x80000000) == 0x80000000) {
            return y;
        }
        return x;
    }
    if ((sig_x > sig_y) && (exp_x >= exp_y)) {
        if ((x & 0x80000000) != 0x80000000) {
            return y;
        }
        return x;
    }
    if ((sig_x < sig_y) && (exp_x <= exp_y)) {
        if ((x & 0x80000000) == 0x80000000) {
            return y;
        }
        return x;
    }
    if ((exp_x.wrapping_sub(exp_y)) > 6) {
        if ((x & 0x80000000) != 0x80000000) {
            return y;
        }
        return x;
    }
    if ((exp_y.wrapping_sub(exp_x)) > 6) {
        if ((x & 0x80000000) == 0x80000000) {
            return y;
        }
        return x;
    }
    if (exp_x > exp_y) {
        sig_n_prime = ((sig_x as u64).wrapping_mul((bid_mult_factor32[(exp_x.wrapping_sub(exp_y)) as usize] as u64)));
        if (sig_n_prime == (sig_y as u64)) {
            return y;
        }
        if (((sig_n_prime > (sig_y as u64))) != (((x & 0x80000000) == 0x80000000))) {
            return y;
        }
        return x;
    }
    sig_n_prime = ((sig_y as u64).wrapping_mul((bid_mult_factor32[(exp_y.wrapping_sub(exp_x)) as usize] as u64)));
    if (sig_n_prime == (sig_x as u64)) {
        return y;
    }
    if ((((sig_x as u64) > sig_n_prime)) != (((x & 0x80000000) == 0x80000000))) {
        return y;
    }
    return x;
}

pub(crate) fn bid32_maxnum_pure(mut x: u32, mut y: u32) -> u32 {
    let mut res: u32 = 0;
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut sig_x: u32 = 0;
    let mut sig_y: u32 = 0;
    let mut sig_n_prime: u64 = 0;
    let mut x_is_zero: bool = false;
    let mut y_is_zero: bool = false;
    if ((x & 0x7c000000) == 0x7c000000) {
        x = (x & 0xfe0fffff);
        if ((x & 0x000fffff) > 999999) {
            x = (x & 0xfe000000);
        }
    } else if ((x & 0x78000000) == 0x78000000) {
        x = (x & (0x80000000 | 0x78000000));
    } else {
        if ((x & 0x60000000) == 0x60000000) {
            if ((((x & 0x1fffff) | 0x800000)) > 9999999) {
                x = ((x & 0x80000000) | ((go_checked_shl_u32((x & 0x1fe00000), go_shift_count_u64((2) as u64)))));
            }
        }
    }
    if ((y & 0x7c000000) == 0x7c000000) {
        y = (y & 0xfe0fffff);
        if ((y & 0x000fffff) > 999999) {
            y = (y & 0xfe000000);
        }
    } else if ((y & 0x78000000) == 0x78000000) {
        y = (y & (0x80000000 | 0x78000000));
    } else {
        if ((y & 0x60000000) == 0x60000000) {
            if ((((y & 0x1fffff) | 0x800000)) > 9999999) {
                y = ((y & 0x80000000) | ((go_checked_shl_u32((y & 0x1fe00000), go_shift_count_u64((2) as u64)))));
            }
        }
    }
    if ((x & 0x7c000000) == 0x7c000000) {
        if ((x & 0x7e000000) == 0x7e000000) {
            x = (x & 0xfdffffff);
            res = x;
        } else {
            if ((y & 0x7c000000) == 0x7c000000) {
                res = x;
            } else {
                res = y;
            }
        }
        return res;
    } else if ((y & 0x7c000000) == 0x7c000000) {
        if ((y & 0x7e000000) == 0x7e000000) {
            y = (y & 0xfdffffff);
            res = y;
        } else {
            res = x;
        }
        return res;
    }
    if (x == y) {
        return x;
    }
    if ((x & 0x78000000) == 0x78000000) {
        if ((x & 0x80000000) == 0x80000000) {
            return y;
        }
        return x;
    } else if ((y & 0x78000000) == 0x78000000) {
        if ((y & 0x80000000) == 0x80000000) {
            return x;
        }
        return y;
    }
    if ((x & 0x60000000) == 0x60000000) {
        exp_x = ((go_checked_shr_u32((x & 0x1fe00000), go_shift_count_u64((21) as u64))) as i64);
        sig_x = ((x & 0x1fffff) | 0x800000);
    } else {
        exp_x = ((go_checked_shr_u32((x & 0x7f800000), go_shift_count_u64((23) as u64))) as i64);
        sig_x = (x & 0x7fffff);
    }
    if ((y & 0x60000000) == 0x60000000) {
        exp_y = ((go_checked_shr_u32((y & 0x1fe00000), go_shift_count_u64((21) as u64))) as i64);
        sig_y = ((y & 0x1fffff) | 0x800000);
    } else {
        exp_y = ((go_checked_shr_u32((y & 0x7f800000), go_shift_count_u64((23) as u64))) as i64);
        sig_y = (y & 0x7fffff);
    }
    if (sig_x == 0) {
        x_is_zero = true;
    }
    if (sig_y == 0) {
        y_is_zero = true;
    }
    if (x_is_zero && y_is_zero) {
        return y;
    } else if x_is_zero {
        if ((y & 0x80000000) == 0x80000000) {
            return x;
        }
        return y;
    } else if y_is_zero {
        if ((x & 0x80000000) != 0x80000000) {
            return x;
        }
        return y;
    }
    if ((((x ^ y) & 0x80000000)) == 0x80000000) {
        if ((y & 0x80000000) == 0x80000000) {
            return x;
        }
        return y;
    }
    if ((sig_x > sig_y) && (exp_x >= exp_y)) {
        if ((x & 0x80000000) != 0x80000000) {
            return x;
        }
        return y;
    }
    if ((sig_x < sig_y) && (exp_x <= exp_y)) {
        if ((x & 0x80000000) == 0x80000000) {
            return x;
        }
        return y;
    }
    if ((exp_x.wrapping_sub(exp_y)) > 6) {
        if ((x & 0x80000000) != 0x80000000) {
            return x;
        }
        return y;
    }
    if ((exp_y.wrapping_sub(exp_x)) > 6) {
        if ((x & 0x80000000) == 0x80000000) {
            return x;
        }
        return y;
    }
    if (exp_x > exp_y) {
        sig_n_prime = ((sig_x as u64).wrapping_mul((bid_mult_factor32[(exp_x.wrapping_sub(exp_y)) as usize] as u64)));
        if (sig_n_prime == (sig_y as u64)) {
            return y;
        }
        if (((sig_n_prime > (sig_y as u64))) != (((x & 0x80000000) == 0x80000000))) {
            return x;
        }
        return y;
    }
    sig_n_prime = ((sig_y as u64).wrapping_mul((bid_mult_factor32[(exp_y.wrapping_sub(exp_x)) as usize] as u64)));
    if (sig_n_prime == (sig_x as u64)) {
        return y;
    }
    if ((((sig_x as u64) > sig_n_prime)) != (((x & 0x80000000) == 0x80000000))) {
        return x;
    }
    return y;
}

pub(crate) fn bid32_minnum_mag_pure(mut x: u32, mut y: u32) -> u32 {
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut sig_x: u32 = 0;
    let mut sig_y: u32 = 0;
    let mut sig_n_prime: u64 = 0;
    if ((x & 0x7c000000) == 0x7c000000) {
        x = (x & 0xfe0fffff);
        if ((x & 0x000fffff) > 999999) {
            x = (x & 0xfe000000);
        }
    } else if ((x & 0x78000000) == 0x78000000) {
        x = (x & (0x80000000 | 0x78000000));
    } else {
        if ((x & 0x60000000) == 0x60000000) {
            if ((((x & 0x1fffff) | 0x800000)) > 9999999) {
                x = ((x & 0x80000000) | ((go_checked_shl_u32((x & 0x1fe00000), go_shift_count_u64((2) as u64)))));
            }
        }
    }
    if ((y & 0x7c000000) == 0x7c000000) {
        y = (y & 0xfe0fffff);
        if ((y & 0x000fffff) > 999999) {
            y = (y & 0xfe000000);
        }
    } else if ((y & 0x78000000) == 0x78000000) {
        y = (y & (0x80000000 | 0x78000000));
    } else {
        if ((y & 0x60000000) == 0x60000000) {
            if ((((y & 0x1fffff) | 0x800000)) > 9999999) {
                y = ((y & 0x80000000) | ((go_checked_shl_u32((y & 0x1fe00000), go_shift_count_u64((2) as u64)))));
            }
        }
    }
    if ((x & 0x7c000000) == 0x7c000000) {
        if ((x & 0x7e000000) == 0x7e000000) {
            return (x & 0xfdffffff);
        }
        if ((y & 0x7c000000) == 0x7c000000) {
            return x;
        }
        return y;
    } else if ((y & 0x7c000000) == 0x7c000000) {
        if ((y & 0x7e000000) == 0x7e000000) {
            return (y & 0xfdffffff);
        }
        return x;
    }
    if (x == y) {
        return x;
    }
    if ((x & 0x78000000) == 0x78000000) {
        if (((x & 0x80000000) == 0x80000000) && ((y & 0x78000000) == 0x78000000)) {
            return x;
        }
        return y;
    } else if ((y & 0x78000000) == 0x78000000) {
        return x;
    }
    if ((x & 0x60000000) == 0x60000000) {
        exp_x = ((go_checked_shr_u32((x & 0x1fe00000), go_shift_count_u64((21) as u64))) as i64);
        sig_x = ((x & 0x1fffff) | 0x800000);
    } else {
        exp_x = ((go_checked_shr_u32((x & 0x7f800000), go_shift_count_u64((23) as u64))) as i64);
        sig_x = (x & 0x7fffff);
    }
    if ((y & 0x60000000) == 0x60000000) {
        exp_y = ((go_checked_shr_u32((y & 0x1fe00000), go_shift_count_u64((21) as u64))) as i64);
        sig_y = ((y & 0x1fffff) | 0x800000);
    } else {
        exp_y = ((go_checked_shr_u32((y & 0x7f800000), go_shift_count_u64((23) as u64))) as i64);
        sig_y = (y & 0x7fffff);
    }
    if (sig_x == 0) {
        return x;
    }
    if (sig_y == 0) {
        return y;
    }
    if ((sig_x > sig_y) && (exp_x >= exp_y)) {
        return y;
    }
    if ((sig_x < sig_y) && (exp_x <= exp_y)) {
        return x;
    }
    if ((exp_x.wrapping_sub(exp_y)) > 6) {
        return y;
    }
    if ((exp_y.wrapping_sub(exp_x)) > 6) {
        return x;
    }
    if (exp_x > exp_y) {
        sig_n_prime = ((sig_x as u64).wrapping_mul((bid_mult_factor32[(exp_x.wrapping_sub(exp_y)) as usize] as u64)));
        if (sig_n_prime == (sig_y as u64)) {
            if ((y & 0x80000000) == 0x80000000) {
                return y;
            }
            return x;
        }
        if (sig_n_prime > (sig_y as u64)) {
            return y;
        }
        return x;
    }
    sig_n_prime = ((sig_y as u64).wrapping_mul((bid_mult_factor32[(exp_y.wrapping_sub(exp_x)) as usize] as u64)));
    if (sig_n_prime == (sig_x as u64)) {
        if ((y & 0x80000000) == 0x80000000) {
            return y;
        }
        return x;
    }
    if ((sig_x as u64) > sig_n_prime) {
        return y;
    }
    return x;
}

pub(crate) fn bid32_maxnum_mag_pure(mut x: u32, mut y: u32) -> u32 {
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut sig_x: u32 = 0;
    let mut sig_y: u32 = 0;
    let mut sig_n_prime: u64 = 0;
    if ((x & 0x7c000000) == 0x7c000000) {
        x = (x & 0xfe0fffff);
        if ((x & 0x000fffff) > 999999) {
            x = (x & 0xfe000000);
        }
    } else if ((x & 0x78000000) == 0x78000000) {
        x = (x & (0x80000000 | 0x78000000));
    } else {
        if ((x & 0x60000000) == 0x60000000) {
            if ((((x & 0x1fffff) | 0x800000)) > 9999999) {
                x = ((x & 0x80000000) | ((go_checked_shl_u32((x & 0x1fe00000), go_shift_count_u64((2) as u64)))));
            }
        }
    }
    if ((y & 0x7c000000) == 0x7c000000) {
        y = (y & 0xfe0fffff);
        if ((y & 0x000fffff) > 999999) {
            y = (y & 0xfe000000);
        }
    } else if ((y & 0x78000000) == 0x78000000) {
        y = (y & (0x80000000 | 0x78000000));
    } else {
        if ((y & 0x60000000) == 0x60000000) {
            if ((((y & 0x1fffff) | 0x800000)) > 9999999) {
                y = ((y & 0x80000000) | ((go_checked_shl_u32((y & 0x1fe00000), go_shift_count_u64((2) as u64)))));
            }
        }
    }
    if ((x & 0x7c000000) == 0x7c000000) {
        if ((x & 0x7e000000) == 0x7e000000) {
            return (x & 0xfdffffff);
        }
        if ((y & 0x7c000000) == 0x7c000000) {
            return x;
        }
        return y;
    } else if ((y & 0x7c000000) == 0x7c000000) {
        if ((y & 0x7e000000) == 0x7e000000) {
            return (y & 0xfdffffff);
        }
        return x;
    }
    if (x == y) {
        return x;
    }
    if ((x & 0x78000000) == 0x78000000) {
        if (((x & 0x80000000) == 0x80000000) && ((y & 0x78000000) == 0x78000000)) {
            return y;
        }
        return x;
    } else if ((y & 0x78000000) == 0x78000000) {
        return y;
    }
    if ((x & 0x60000000) == 0x60000000) {
        exp_x = ((go_checked_shr_u32((x & 0x1fe00000), go_shift_count_u64((21) as u64))) as i64);
        sig_x = ((x & 0x1fffff) | 0x800000);
    } else {
        exp_x = ((go_checked_shr_u32((x & 0x7f800000), go_shift_count_u64((23) as u64))) as i64);
        sig_x = (x & 0x7fffff);
    }
    if ((y & 0x60000000) == 0x60000000) {
        exp_y = ((go_checked_shr_u32((y & 0x1fe00000), go_shift_count_u64((21) as u64))) as i64);
        sig_y = ((y & 0x1fffff) | 0x800000);
    } else {
        exp_y = ((go_checked_shr_u32((y & 0x7f800000), go_shift_count_u64((23) as u64))) as i64);
        sig_y = (y & 0x7fffff);
    }
    if (sig_x == 0) {
        return y;
    }
    if (sig_y == 0) {
        return x;
    }
    if ((sig_x > sig_y) && (exp_x >= exp_y)) {
        return x;
    }
    if ((sig_x < sig_y) && (exp_x <= exp_y)) {
        return y;
    }
    if ((exp_x.wrapping_sub(exp_y)) > 6) {
        return x;
    }
    if ((exp_y.wrapping_sub(exp_x)) > 6) {
        return y;
    }
    if (exp_x > exp_y) {
        sig_n_prime = ((sig_x as u64).wrapping_mul((bid_mult_factor32[(exp_x.wrapping_sub(exp_y)) as usize] as u64)));
        if (sig_n_prime == (sig_y as u64)) {
            if ((y & 0x80000000) == 0x80000000) {
                return x;
            }
            return y;
        }
        if (sig_n_prime > (sig_y as u64)) {
            return x;
        }
        return y;
    }
    sig_n_prime = ((sig_y as u64).wrapping_mul((bid_mult_factor32[(exp_y.wrapping_sub(exp_x)) as usize] as u64)));
    if (sig_n_prime == (sig_x as u64)) {
        if ((y & 0x80000000) == 0x80000000) {
            return x;
        }
        return y;
    }
    if ((sig_x as u64) > sig_n_prime) {
        return x;
    }
    return y;
}

pub(crate) fn bid32_same_quantum_pure(mut x: u32, mut y: u32) -> bool {
    let mut exp_x: u32 = 0;
    let mut exp_y: u32 = 0;
    if (((x & 0x7c000000) == 0x7c000000) || ((y & 0x7c000000) == 0x7c000000)) {
        return (((x & 0x7c000000) == 0x7c000000) && ((y & 0x7c000000) == 0x7c000000));
    }
    if (((x & 0x78000000) == 0x78000000) || ((y & 0x78000000) == 0x78000000)) {
        return (((x & 0x78000000) == 0x78000000) && ((y & 0x78000000) == 0x78000000));
    }
    if ((x & 0x60000000) == 0x60000000) {
        exp_x = (go_checked_shr_u32((x & 0x1fe00000), go_shift_count_u64((21) as u64)));
    } else {
        exp_x = (go_checked_shr_u32((x & 0x7f800000), go_shift_count_u64((23) as u64)));
    }
    if ((y & 0x60000000) == 0x60000000) {
        exp_y = (go_checked_shr_u32((y & 0x1fe00000), go_shift_count_u64((21) as u64)));
    } else {
        exp_y = (go_checked_shr_u32((y & 0x7f800000), go_shift_count_u64((23) as u64)));
    }
    return (exp_x == exp_y);
}

pub(crate) fn bid32_quantum_pure(mut x: u32) -> u32 {
    let mut int_exp: i64 = 0;
    if ((x & 0x78000000) == 0x78000000) {
        return (x & (!(0x80000000 as u32)));
    }
    if ((x & 0x7c000000) == 0x7c000000) {
        return (x & 0x7fffffff);
    }
    int_exp = (((((go_checked_shr_u32(x, go_shift_count_u64((23) as u64)))) & 0xff) as i64).wrapping_sub(101));
    return (((go_checked_shl_i64(((int_exp.wrapping_add(101))), go_shift_count_u64((23) as u64))) as u32).wrapping_add(1));
}
