// Auto-generated from bid32_noncomp.go by go2rs. Do not edit.

use super::prelude::*;

pub fn bid32_is_signed(mut x: u32) -> i64 {
    if ((x & 0x80000000) == 0x80000000) {
        return 1;
    }
    return 0;
}

pub fn bid32_is_normal(mut x: u32) -> i64 {
    let mut res: i64 = 0;
    let mut sig_x_prime: u64 = 0;
    let mut sig_x: u32 = 0;
    let mut exp_x: u32 = 0;
    if ((x & 0x78000000) == 0x78000000) {
        res = 0;
    } else {
        if ((x & 0x60000000) == 0x60000000) {
            sig_x = ((x & 0x1fffff) | 0x800000);
            if ((sig_x > 9999999) || (sig_x == 0)) {
                return 0;
            }
            exp_x = (go_checked_shr_u32((x & 0x1fe00000), go_shift_count_u64((21) as u64)));
        } else {
            sig_x = (x & 0x7fffff);
            if (sig_x == 0) {
                return 0;
            }
            exp_x = (go_checked_shr_u32((x & 0x7f800000), go_shift_count_u64((23) as u64)));
        }
        if (exp_x < 6) {
            sig_x_prime = ((sig_x as u64).wrapping_mul(bid32_mult_factor[exp_x as usize]));
            if (sig_x_prime < 1000000) {
                res = 0;
            } else {
                res = 1;
            }
        } else {
            res = 1;
        }
    }
    return res;
}

pub fn bid32_is_subnormal(mut x: u32) -> i64 {
    let mut res: i64 = 0;
    let mut sig_x_prime: u64 = 0;
    let mut sig_x: u32 = 0;
    let mut exp_x: u32 = 0;
    if ((x & 0x78000000) == 0x78000000) {
        res = 0;
    } else {
        if ((x & 0x60000000) == 0x60000000) {
            sig_x = ((x & 0x1fffff) | 0x800000);
            if ((sig_x > 9999999) || (sig_x == 0)) {
                return 0;
            }
            exp_x = (go_checked_shr_u32((x & 0x1fe00000), go_shift_count_u64((21) as u64)));
        } else {
            sig_x = (x & 0x7fffff);
            if (sig_x == 0) {
                return 0;
            }
            exp_x = (go_checked_shr_u32((x & 0x7f800000), go_shift_count_u64((23) as u64)));
        }
        if (exp_x < 6) {
            sig_x_prime = ((sig_x as u64).wrapping_mul(bid32_mult_factor[exp_x as usize]));
            if (sig_x_prime < 1000000) {
                res = 1;
            } else {
                res = 0;
            }
        } else {
            res = 0;
        }
    }
    return res;
}

pub fn bid32_is_finite(mut x: u32) -> i64 {
    if ((x & 0x78000000) != 0x78000000) {
        return 1;
    }
    return 0;
}

pub fn bid32_is_zero32(mut x: u32) -> i64 {
    let mut res: i64 = 0;
    if ((x & 0x78000000) == 0x78000000) {
        res = 0;
    } else if ((x & 0x60000000) == 0x60000000) {
        if ((((x & 0x1fffff) | 0x800000)) > 9999999) {
            res = 1;
        } else {
            res = 0;
        }
    } else {
        if ((x & 0x7fffff) == 0) {
            res = 1;
        } else {
            res = 0;
        }
    }
    return res;
}

pub fn bid32_is_inf32(mut x: u32) -> i64 {
    if ((((x & 0x78000000) == 0x78000000)) && (((x & 0x7c000000) != 0x7c000000))) {
        return 1;
    }
    return 0;
}

pub fn bid32_is_signaling(mut x: u32) -> i64 {
    if ((x & 0x7e000000) == 0x7e000000) {
        return 1;
    }
    return 0;
}

pub fn bid32_is_na_n32(mut x: u32) -> i64 {
    if ((x & 0x7c000000) == 0x7c000000) {
        return 1;
    }
    return 0;
}

pub fn bid32_is_canonical(mut x: u32) -> i64 {
    if ((x & 0x7c000000) == 0x7c000000) {
        if ((x & 0x01f00000) != 0) {
            return 0;
        }
        if ((x & 0x000fffff) > 999999) {
            return 0;
        }
        return 1;
    }
    if ((x & 0x78000000) == 0x78000000) {
        if ((x & 0x03ffffff) != 0) {
            return 0;
        }
        return 1;
    }
    if ((x & 0x60000000) == 0x60000000) {
        if ((((x & 0x1fffff) | 0x800000)) > 9999999) {
            return 0;
        }
    }
    return 1;
}

pub fn bid32_copy(mut x: u32) -> u32 {
    return x;
}

pub fn bid32_copy_sign(mut x: u32, mut y: u32) -> u32 {
    return ((x & 0x7fffffff) | (y & 0x80000000));
}

pub fn bid32_radix() -> i64 {
    return 10;
}

pub fn bid32_total_order(mut x: u32, mut y: u32) -> i64 {
    let mut res: i64 = 0;
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut sig_x: u32 = 0;
    let mut sig_y: u32 = 0;
    let mut pyld_y: u32 = 0;
    let mut pyld_x: u32 = 0;
    let mut sig_n_prime: u64 = 0;
    let mut x_is_zero: i64 = 0;
    let mut y_is_zero: i64 = 0;
    if ((x & 0x7c000000) == 0x7c000000) {
        if ((x & 0x80000000) == 0x80000000) {
            if (((y & 0x7c000000) != 0x7c000000) || ((y & 0x80000000) != 0x80000000)) {
                res = 1;
                return res;
            } else {
                if (!(((((y & 0x7e000000) == 0x7e000000)) != (((x & 0x7e000000) == 0x7e000000))))) {
                    pyld_y = (y & 0x000fffff);
                    pyld_x = (x & 0x000fffff);
                    if ((pyld_y > 999999) || (pyld_y == 0)) {
                        res = 1;
                        return res;
                    }
                    if ((pyld_x > 999999) || (pyld_x == 0)) {
                        res = 0;
                        return res;
                    }
                    res = 0;
                    if (pyld_x >= pyld_y) {
                        res = 1;
                    }
                    return res;
                } else {
                    res = 0;
                    if ((y & 0x7e000000) == 0x7e000000) {
                        res = 1;
                    }
                    return res;
                }
            }
        } else {
            if (((y & 0x7c000000) != 0x7c000000) || ((y & 0x80000000) == 0x80000000)) {
                res = 0;
                return res;
            } else {
                if (!(((((y & 0x7e000000) == 0x7e000000)) != (((x & 0x7e000000) == 0x7e000000))))) {
                    pyld_y = (y & 0x000fffff);
                    pyld_x = (x & 0x000fffff);
                    if ((pyld_x > 999999) || (pyld_x == 0)) {
                        res = 1;
                        return res;
                    }
                    if ((pyld_y > 999999) || (pyld_y == 0)) {
                        res = 0;
                        return res;
                    }
                    res = 0;
                    if (pyld_x <= pyld_y) {
                        res = 1;
                    }
                    return res;
                } else {
                    res = 0;
                    if ((x & 0x7e000000) == 0x7e000000) {
                        res = 1;
                    }
                    return res;
                }
            }
        }
    } else if ((y & 0x7c000000) == 0x7c000000) {
        res = 0;
        if ((y & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return res;
    }
    if (x == y) {
        res = 1;
        return res;
    }
    if ((((x & 0x80000000) == 0x80000000)) != (((y & 0x80000000) == 0x80000000))) {
        res = 0;
        if ((x & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return res;
    }
    if ((x & 0x78000000) == 0x78000000) {
        if ((x & 0x80000000) == 0x80000000) {
            res = 1;
            return res;
        } else {
            res = 0;
            if ((y & 0x78000000) == 0x78000000) {
                res = 1;
            }
            return res;
        }
    } else if ((y & 0x78000000) == 0x78000000) {
        res = 0;
        if ((y & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return res;
    }
    if ((x & 0x60000000) == 0x60000000) {
        exp_x = ((go_checked_shr_u32((x & 0x1fe00000), go_shift_count_u64((21) as u64))) as i64);
        sig_x = ((x & 0x1fffff) | 0x800000);
        if ((sig_x > 9999999) || (sig_x == 0)) {
            x_is_zero = 1;
        }
    } else {
        exp_x = ((go_checked_shr_u32((x & 0x7f800000), go_shift_count_u64((23) as u64))) as i64);
        sig_x = (x & 0x7fffff);
        if (sig_x == 0) {
            x_is_zero = 1;
        }
    }
    if ((y & 0x60000000) == 0x60000000) {
        exp_y = ((go_checked_shr_u32((y & 0x1fe00000), go_shift_count_u64((21) as u64))) as i64);
        sig_y = ((y & 0x1fffff) | 0x800000);
        if ((sig_y > 9999999) || (sig_y == 0)) {
            y_is_zero = 1;
        }
    } else {
        exp_y = ((go_checked_shr_u32((y & 0x7f800000), go_shift_count_u64((23) as u64))) as i64);
        sig_y = (y & 0x7fffff);
        if (sig_y == 0) {
            y_is_zero = 1;
        }
    }
    if ((x_is_zero != 0) && (y_is_zero != 0)) {
        if (exp_x == exp_y) {
            res = 1;
            return res;
        }
        res = 0;
        if ((exp_x <= exp_y) != (((x & 0x80000000) == 0x80000000))) {
            res = 1;
        }
        return res;
    }
    if (x_is_zero != 0) {
        res = 0;
        if ((y & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return res;
    }
    if (y_is_zero != 0) {
        res = 0;
        if ((x & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return res;
    }
    if ((sig_x > sig_y) && (exp_x >= exp_y)) {
        res = 0;
        if ((x & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return res;
    }
    if ((sig_x < sig_y) && (exp_x <= exp_y)) {
        res = 0;
        if ((x & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return res;
    }
    if ((exp_x.wrapping_sub(exp_y)) > 6) {
        res = 0;
        if ((x & 0x80000000) == 0x80000000) {
            res = 1;
        }
        return res;
    }
    if ((exp_y.wrapping_sub(exp_x)) > 6) {
        res = 0;
        if ((x & 0x80000000) != 0x80000000) {
            res = 1;
        }
        return res;
    }
    if (exp_x > exp_y) {
        sig_n_prime = ((sig_x as u64).wrapping_mul(bid32_mult_factor[(exp_x.wrapping_sub(exp_y)) as usize]));
        if (sig_n_prime == (sig_y as u64)) {
            res = 0;
            if ((exp_x <= exp_y) != (((x & 0x80000000) == 0x80000000))) {
                res = 1;
            }
            return res;
        }
        res = 0;
        if (((sig_n_prime < (sig_y as u64))) != (((x & 0x80000000) == 0x80000000))) {
            res = 1;
        }
        return res;
    }
    sig_n_prime = ((sig_y as u64).wrapping_mul(bid32_mult_factor[(exp_y.wrapping_sub(exp_x)) as usize]));
    if (sig_n_prime == (sig_x as u64)) {
        res = 0;
        if ((exp_x <= exp_y) != (((x & 0x80000000) == 0x80000000))) {
            res = 1;
        }
        return res;
    }
    res = 0;
    if ((((sig_x as u64) < sig_n_prime)) != (((x & 0x80000000) == 0x80000000))) {
        res = 1;
    }
    return res;
}

pub fn bid32_total_order_mag(mut x: u32, mut y: u32) -> i64 {
    return bid32_total_order((x & 0x7fffffff), (y & 0x7fffffff));
}
