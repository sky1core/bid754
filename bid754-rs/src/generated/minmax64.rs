// Auto-generated from minmax64.go by go2rs. Do not edit.

use super::prelude::*;

pub fn bid64_min_num(mut x: u64, mut y: u64) -> (u64, u32) {
    let mut res: u64 = 0;
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut sig_x: u64 = 0;
    let mut sig_y: u64 = 0;
    let mut sig_n_prime: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut x_is_zero: i64 = 0;
    let mut y_is_zero: i64 = 0;
    let mut flags: u32 = 0;
    if ((x & 0x7c00000000000000) == 0x7c00000000000000) {
        x = (x & 0xfe03ffffffffffff);
        if ((x & 0x0003ffffffffffff) > 999999999999999) {
            x = (x & 0xfe00000000000000);
        }
    } else if ((x & 0x7800000000000000) == 0x7800000000000000) {
        x = (x & (0x8000000000000000 | 0x7800000000000000));
    } else {
        if ((x & 0x6000000000000000) == 0x6000000000000000) {
            if ((((x & 0x7ffffffffffff) | 0x20000000000000)) > 9999999999999999) {
                x = ((x & 0x8000000000000000) | ((go_checked_shl_u64((x & 0x1ff8000000000000), go_shift_count_u64((2) as u64)))));
            }
        }
    }
    if ((y & 0x7c00000000000000) == 0x7c00000000000000) {
        y = (y & 0xfe03ffffffffffff);
        if ((y & 0x0003ffffffffffff) > 999999999999999) {
            y = (y & 0xfe00000000000000);
        }
    } else if ((y & 0x7800000000000000) == 0x7800000000000000) {
        y = (y & (0x8000000000000000 | 0x7800000000000000));
    } else {
        if ((y & 0x6000000000000000) == 0x6000000000000000) {
            if ((((y & 0x7ffffffffffff) | 0x20000000000000)) > 9999999999999999) {
                y = ((y & 0x8000000000000000) | ((go_checked_shl_u64((y & 0x1ff8000000000000), go_shift_count_u64((2) as u64)))));
            }
        }
    }
    if ((x & 0x7c00000000000000) == 0x7c00000000000000) {
        if ((x & 0x7e00000000000000) == 0x7e00000000000000) {
            flags |= 1;
            x = (x & 0xfdffffffffffffff);
            res = x;
        } else {
            if ((y & 0x7c00000000000000) == 0x7c00000000000000) {
                if ((y & 0x7e00000000000000) == 0x7e00000000000000) {
                    flags |= 1;
                }
                res = x;
            } else {
                res = y;
            }
        }
        return (res, flags);
    } else if ((y & 0x7c00000000000000) == 0x7c00000000000000) {
        if ((y & 0x7e00000000000000) == 0x7e00000000000000) {
            flags |= 1;
            y = (y & 0xfdffffffffffffff);
            res = y;
        } else {
            res = x;
        }
        return (res, flags);
    }
    if (x == y) {
        res = x;
        return (res, flags);
    }
    if ((x & 0x7800000000000000) == 0x7800000000000000) {
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = x;
            return (res, flags);
        }
        res = y;
        return (res, flags);
    } else if ((y & 0x7800000000000000) == 0x7800000000000000) {
        if ((y & 0x8000000000000000) == 0x8000000000000000) {
            res = y;
        } else {
            res = x;
        }
        return (res, flags);
    }
    if ((x & 0x6000000000000000) == 0x6000000000000000) {
        exp_x = ((go_checked_shr_u64((x & 0x1ff8000000000000), go_shift_count_u64((51) as u64))) as i64);
        sig_x = ((x & 0x7ffffffffffff) | 0x20000000000000);
    } else {
        exp_x = ((go_checked_shr_u64((x & 0x7fe0000000000000), go_shift_count_u64((53) as u64))) as i64);
        sig_x = (x & 0x1fffffffffffff);
    }
    if ((y & 0x6000000000000000) == 0x6000000000000000) {
        exp_y = ((go_checked_shr_u64((y & 0x1ff8000000000000), go_shift_count_u64((51) as u64))) as i64);
        sig_y = ((y & 0x7ffffffffffff) | 0x20000000000000);
    } else {
        exp_y = ((go_checked_shr_u64((y & 0x7fe0000000000000), go_shift_count_u64((53) as u64))) as i64);
        sig_y = (y & 0x1fffffffffffff);
    }
    if (sig_x == 0) {
        x_is_zero = 1;
    }
    if (sig_y == 0) {
        y_is_zero = 1;
    }
    if ((x_is_zero != 0) && (y_is_zero != 0)) {
        res = y;
        return (res, flags);
    } else if (x_is_zero != 0) {
        if ((y & 0x8000000000000000) == 0x8000000000000000) {
            res = y;
        } else {
            res = x;
        }
        return (res, flags);
    } else if (y_is_zero != 0) {
        if ((x & 0x8000000000000000) != 0x8000000000000000) {
            res = y;
        } else {
            res = x;
        }
        return (res, flags);
    }
    if ((((x ^ y) & 0x8000000000000000)) == 0x8000000000000000) {
        if ((y & 0x8000000000000000) == 0x8000000000000000) {
            res = y;
        } else {
            res = x;
        }
        return (res, flags);
    }
    if ((sig_x > sig_y) && (exp_x >= exp_y)) {
        if ((x & 0x8000000000000000) != 0x8000000000000000) {
            res = y;
        } else {
            res = x;
        }
        return (res, flags);
    }
    if ((sig_x < sig_y) && (exp_x <= exp_y)) {
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = y;
        } else {
            res = x;
        }
        return (res, flags);
    }
    if ((exp_x.wrapping_sub(exp_y)) > 15) {
        if ((x & 0x8000000000000000) != 0x8000000000000000) {
            res = y;
        } else {
            res = x;
        }
        return (res, flags);
    }
    if ((exp_y.wrapping_sub(exp_x)) > 15) {
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = y;
        } else {
            res = x;
        }
        return (res, flags);
    }
    if (exp_x > exp_y) {
        sig_n_prime = __mul_64x64_to_128(sig_x, bid_mult_factor_minmax[(exp_x.wrapping_sub(exp_y)) as usize]);
        if ((sig_n_prime.w[1] == 0) && (sig_n_prime.w[0] == sig_y)) {
            res = y;
            return (res, flags);
        }
        let mut cond = ((sig_n_prime.w[1] > 0) || (sig_n_prime.w[0] > sig_y));
        let mut sign = ((x & 0x8000000000000000) == 0x8000000000000000);
        if (cond != sign) {
            res = y;
        } else {
            res = x;
        }
        return (res, flags);
    }
    sig_n_prime = __mul_64x64_to_128(sig_y, bid_mult_factor_minmax[(exp_y.wrapping_sub(exp_x)) as usize]);
    if ((sig_n_prime.w[1] == 0) && (sig_n_prime.w[0] == sig_x)) {
        res = y;
        return (res, flags);
    }
    let mut cond = ((sig_n_prime.w[1] == 0) && (sig_x > sig_n_prime.w[0]));
    let mut sign = ((x & 0x8000000000000000) == 0x8000000000000000);
    if (cond != sign) {
        res = y;
    } else {
        res = x;
    }
    return (res, flags);
}

pub fn bid64_min_num_mag(mut x: u64, mut y: u64) -> (u64, u32) {
    let mut res: u64 = 0;
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut sig_x: u64 = 0;
    let mut sig_y: u64 = 0;
    let mut sig_n_prime: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut flags: u32 = 0;
    if ((x & 0x7c00000000000000) == 0x7c00000000000000) {
        x = (x & 0xfe03ffffffffffff);
        if ((x & 0x0003ffffffffffff) > 999999999999999) {
            x = (x & 0xfe00000000000000);
        }
    } else if ((x & 0x7800000000000000) == 0x7800000000000000) {
        x = (x & (0x8000000000000000 | 0x7800000000000000));
    } else {
        if ((x & 0x6000000000000000) == 0x6000000000000000) {
            if ((((x & 0x7ffffffffffff) | 0x20000000000000)) > 9999999999999999) {
                x = ((x & 0x8000000000000000) | ((go_checked_shl_u64((x & 0x1ff8000000000000), go_shift_count_u64((2) as u64)))));
            }
        }
    }
    if ((y & 0x7c00000000000000) == 0x7c00000000000000) {
        y = (y & 0xfe03ffffffffffff);
        if ((y & 0x0003ffffffffffff) > 999999999999999) {
            y = (y & 0xfe00000000000000);
        }
    } else if ((y & 0x7800000000000000) == 0x7800000000000000) {
        y = (y & (0x8000000000000000 | 0x7800000000000000));
    } else {
        if ((y & 0x6000000000000000) == 0x6000000000000000) {
            if ((((y & 0x7ffffffffffff) | 0x20000000000000)) > 9999999999999999) {
                y = ((y & 0x8000000000000000) | ((go_checked_shl_u64((y & 0x1ff8000000000000), go_shift_count_u64((2) as u64)))));
            }
        }
    }
    if ((x & 0x7c00000000000000) == 0x7c00000000000000) {
        if ((x & 0x7e00000000000000) == 0x7e00000000000000) {
            flags |= 1;
            x = (x & 0xfdffffffffffffff);
            res = x;
        } else {
            if ((y & 0x7c00000000000000) == 0x7c00000000000000) {
                if ((y & 0x7e00000000000000) == 0x7e00000000000000) {
                    flags |= 1;
                }
                res = x;
            } else {
                res = y;
            }
        }
        return (res, flags);
    } else if ((y & 0x7c00000000000000) == 0x7c00000000000000) {
        if ((y & 0x7e00000000000000) == 0x7e00000000000000) {
            flags |= 1;
            y = (y & 0xfdffffffffffffff);
            res = y;
        } else {
            res = x;
        }
        return (res, flags);
    }
    if (x == y) {
        res = x;
        return (res, flags);
    }
    if ((x & 0x7800000000000000) == 0x7800000000000000) {
        if (((x & 0x8000000000000000) == 0x8000000000000000) && ((y & 0x7800000000000000) == 0x7800000000000000)) {
            res = x;
        } else {
            res = y;
        }
        return (res, flags);
    } else if ((y & 0x7800000000000000) == 0x7800000000000000) {
        res = x;
        return (res, flags);
    }
    if ((x & 0x6000000000000000) == 0x6000000000000000) {
        exp_x = ((go_checked_shr_u64((x & 0x1ff8000000000000), go_shift_count_u64((51) as u64))) as i64);
        sig_x = ((x & 0x7ffffffffffff) | 0x20000000000000);
    } else {
        exp_x = ((go_checked_shr_u64((x & 0x7fe0000000000000), go_shift_count_u64((53) as u64))) as i64);
        sig_x = (x & 0x1fffffffffffff);
    }
    if ((y & 0x6000000000000000) == 0x6000000000000000) {
        exp_y = ((go_checked_shr_u64((y & 0x1ff8000000000000), go_shift_count_u64((51) as u64))) as i64);
        sig_y = ((y & 0x7ffffffffffff) | 0x20000000000000);
    } else {
        exp_y = ((go_checked_shr_u64((y & 0x7fe0000000000000), go_shift_count_u64((53) as u64))) as i64);
        sig_y = (y & 0x1fffffffffffff);
    }
    if (sig_x == 0) {
        res = x;
        return (res, flags);
    }
    if (sig_y == 0) {
        res = y;
        return (res, flags);
    }
    if ((sig_x > sig_y) && (exp_x >= exp_y)) {
        res = y;
        return (res, flags);
    }
    if ((sig_x < sig_y) && (exp_x <= exp_y)) {
        res = x;
        return (res, flags);
    }
    if ((exp_x.wrapping_sub(exp_y)) > 15) {
        res = y;
        return (res, flags);
    }
    if ((exp_y.wrapping_sub(exp_x)) > 15) {
        res = x;
        return (res, flags);
    }
    if (exp_x > exp_y) {
        sig_n_prime = __mul_64x64_to_128(sig_x, bid_mult_factor_minmax[(exp_x.wrapping_sub(exp_y)) as usize]);
        if ((sig_n_prime.w[1] == 0) && (sig_n_prime.w[0] == sig_y)) {
            if ((y & 0x8000000000000000) == 0x8000000000000000) {
                res = y;
            } else {
                res = x;
            }
            return (res, flags);
        }
        if ((sig_n_prime.w[1] != 0) || (sig_n_prime.w[0] > sig_y)) {
            res = y;
        } else {
            res = x;
        }
        return (res, flags);
    }
    sig_n_prime = __mul_64x64_to_128(sig_y, bid_mult_factor_minmax[(exp_y.wrapping_sub(exp_x)) as usize]);
    if ((sig_n_prime.w[1] == 0) && (sig_n_prime.w[0] == sig_x)) {
        if ((y & 0x8000000000000000) == 0x8000000000000000) {
            res = y;
        } else {
            res = x;
        }
        return (res, flags);
    }
    if ((sig_n_prime.w[1] == 0) && (sig_x > sig_n_prime.w[0])) {
        res = y;
    } else {
        res = x;
    }
    return (res, flags);
}

pub fn bid64_max_num(mut x: u64, mut y: u64) -> (u64, u32) {
    let mut res: u64 = 0;
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut sig_x: u64 = 0;
    let mut sig_y: u64 = 0;
    let mut sig_n_prime: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut x_is_zero: i64 = 0;
    let mut y_is_zero: i64 = 0;
    let mut flags: u32 = 0;
    if ((x & 0x7c00000000000000) == 0x7c00000000000000) {
        x = (x & 0xfe03ffffffffffff);
        if ((x & 0x0003ffffffffffff) > 999999999999999) {
            x = (x & 0xfe00000000000000);
        }
    } else if ((x & 0x7800000000000000) == 0x7800000000000000) {
        x = (x & (0x8000000000000000 | 0x7800000000000000));
    } else {
        if ((x & 0x6000000000000000) == 0x6000000000000000) {
            if ((((x & 0x7ffffffffffff) | 0x20000000000000)) > 9999999999999999) {
                x = ((x & 0x8000000000000000) | ((go_checked_shl_u64((x & 0x1ff8000000000000), go_shift_count_u64((2) as u64)))));
            }
        }
    }
    if ((y & 0x7c00000000000000) == 0x7c00000000000000) {
        y = (y & 0xfe03ffffffffffff);
        if ((y & 0x0003ffffffffffff) > 999999999999999) {
            y = (y & 0xfe00000000000000);
        }
    } else if ((y & 0x7800000000000000) == 0x7800000000000000) {
        y = (y & (0x8000000000000000 | 0x7800000000000000));
    } else {
        if ((y & 0x6000000000000000) == 0x6000000000000000) {
            if ((((y & 0x7ffffffffffff) | 0x20000000000000)) > 9999999999999999) {
                y = ((y & 0x8000000000000000) | ((go_checked_shl_u64((y & 0x1ff8000000000000), go_shift_count_u64((2) as u64)))));
            }
        }
    }
    if ((x & 0x7c00000000000000) == 0x7c00000000000000) {
        if ((x & 0x7e00000000000000) == 0x7e00000000000000) {
            flags |= 1;
            x = (x & 0xfdffffffffffffff);
            res = x;
        } else {
            if ((y & 0x7c00000000000000) == 0x7c00000000000000) {
                if ((y & 0x7e00000000000000) == 0x7e00000000000000) {
                    flags |= 1;
                }
                res = x;
            } else {
                res = y;
            }
        }
        return (res, flags);
    } else if ((y & 0x7c00000000000000) == 0x7c00000000000000) {
        if ((y & 0x7e00000000000000) == 0x7e00000000000000) {
            flags |= 1;
            y = (y & 0xfdffffffffffffff);
            res = y;
        } else {
            res = x;
        }
        return (res, flags);
    }
    if (x == y) {
        res = x;
        return (res, flags);
    }
    if ((x & 0x7800000000000000) == 0x7800000000000000) {
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = y;
        } else {
            res = x;
        }
        return (res, flags);
    } else if ((y & 0x7800000000000000) == 0x7800000000000000) {
        if ((y & 0x8000000000000000) == 0x8000000000000000) {
            res = x;
        } else {
            res = y;
        }
        return (res, flags);
    }
    if ((x & 0x6000000000000000) == 0x6000000000000000) {
        exp_x = ((go_checked_shr_u64((x & 0x1ff8000000000000), go_shift_count_u64((51) as u64))) as i64);
        sig_x = ((x & 0x7ffffffffffff) | 0x20000000000000);
    } else {
        exp_x = ((go_checked_shr_u64((x & 0x7fe0000000000000), go_shift_count_u64((53) as u64))) as i64);
        sig_x = (x & 0x1fffffffffffff);
    }
    if ((y & 0x6000000000000000) == 0x6000000000000000) {
        exp_y = ((go_checked_shr_u64((y & 0x1ff8000000000000), go_shift_count_u64((51) as u64))) as i64);
        sig_y = ((y & 0x7ffffffffffff) | 0x20000000000000);
    } else {
        exp_y = ((go_checked_shr_u64((y & 0x7fe0000000000000), go_shift_count_u64((53) as u64))) as i64);
        sig_y = (y & 0x1fffffffffffff);
    }
    if (sig_x == 0) {
        x_is_zero = 1;
    }
    if (sig_y == 0) {
        y_is_zero = 1;
    }
    if ((x_is_zero != 0) && (y_is_zero != 0)) {
        res = y;
        return (res, flags);
    } else if (x_is_zero != 0) {
        if ((y & 0x8000000000000000) == 0x8000000000000000) {
            res = x;
        } else {
            res = y;
        }
        return (res, flags);
    } else if (y_is_zero != 0) {
        if ((x & 0x8000000000000000) != 0x8000000000000000) {
            res = x;
        } else {
            res = y;
        }
        return (res, flags);
    }
    if ((((x ^ y) & 0x8000000000000000)) == 0x8000000000000000) {
        if ((y & 0x8000000000000000) == 0x8000000000000000) {
            res = x;
        } else {
            res = y;
        }
        return (res, flags);
    }
    if ((sig_x > sig_y) && (exp_x >= exp_y)) {
        if ((x & 0x8000000000000000) != 0x8000000000000000) {
            res = x;
        } else {
            res = y;
        }
        return (res, flags);
    }
    if ((sig_x < sig_y) && (exp_x <= exp_y)) {
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = x;
        } else {
            res = y;
        }
        return (res, flags);
    }
    if ((exp_x.wrapping_sub(exp_y)) > 15) {
        if ((x & 0x8000000000000000) != 0x8000000000000000) {
            res = x;
        } else {
            res = y;
        }
        return (res, flags);
    }
    if ((exp_y.wrapping_sub(exp_x)) > 15) {
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = x;
        } else {
            res = y;
        }
        return (res, flags);
    }
    if (exp_x > exp_y) {
        sig_n_prime = __mul_64x64_to_128(sig_x, bid_mult_factor_minmax[(exp_x.wrapping_sub(exp_y)) as usize]);
        if ((sig_n_prime.w[1] == 0) && (sig_n_prime.w[0] == sig_y)) {
            res = y;
            return (res, flags);
        }
        let mut cond = ((sig_n_prime.w[1] > 0) || (sig_n_prime.w[0] > sig_y));
        let mut sign = ((x & 0x8000000000000000) == 0x8000000000000000);
        if (cond != sign) {
            res = x;
        } else {
            res = y;
        }
        return (res, flags);
    }
    sig_n_prime = __mul_64x64_to_128(sig_y, bid_mult_factor_minmax[(exp_y.wrapping_sub(exp_x)) as usize]);
    if ((sig_n_prime.w[1] == 0) && (sig_n_prime.w[0] == sig_x)) {
        res = y;
        return (res, flags);
    }
    let mut cond = ((sig_n_prime.w[1] == 0) && (sig_x > sig_n_prime.w[0]));
    let mut sign = ((x & 0x8000000000000000) == 0x8000000000000000);
    if (cond != sign) {
        res = x;
    } else {
        res = y;
    }
    return (res, flags);
}

pub fn bid64_max_num_mag(mut x: u64, mut y: u64) -> (u64, u32) {
    let mut res: u64 = 0;
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut sig_x: u64 = 0;
    let mut sig_y: u64 = 0;
    let mut sig_n_prime: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut flags: u32 = 0;
    if ((x & 0x7c00000000000000) == 0x7c00000000000000) {
        x = (x & 0xfe03ffffffffffff);
        if ((x & 0x0003ffffffffffff) > 999999999999999) {
            x = (x & 0xfe00000000000000);
        }
    } else if ((x & 0x7800000000000000) == 0x7800000000000000) {
        x = (x & (0x8000000000000000 | 0x7800000000000000));
    } else {
        if ((x & 0x6000000000000000) == 0x6000000000000000) {
            if ((((x & 0x7ffffffffffff) | 0x20000000000000)) > 9999999999999999) {
                x = ((x & 0x8000000000000000) | ((go_checked_shl_u64((x & 0x1ff8000000000000), go_shift_count_u64((2) as u64)))));
            }
        }
    }
    if ((y & 0x7c00000000000000) == 0x7c00000000000000) {
        y = (y & 0xfe03ffffffffffff);
        if ((y & 0x0003ffffffffffff) > 999999999999999) {
            y = (y & 0xfe00000000000000);
        }
    } else if ((y & 0x7800000000000000) == 0x7800000000000000) {
        y = (y & (0x8000000000000000 | 0x7800000000000000));
    } else {
        if ((y & 0x6000000000000000) == 0x6000000000000000) {
            if ((((y & 0x7ffffffffffff) | 0x20000000000000)) > 9999999999999999) {
                y = ((y & 0x8000000000000000) | ((go_checked_shl_u64((y & 0x1ff8000000000000), go_shift_count_u64((2) as u64)))));
            }
        }
    }
    if ((x & 0x7c00000000000000) == 0x7c00000000000000) {
        if ((x & 0x7e00000000000000) == 0x7e00000000000000) {
            flags |= 1;
            x = (x & 0xfdffffffffffffff);
            res = x;
        } else {
            if ((y & 0x7c00000000000000) == 0x7c00000000000000) {
                if ((y & 0x7e00000000000000) == 0x7e00000000000000) {
                    flags |= 1;
                }
                res = x;
            } else {
                res = y;
            }
        }
        return (res, flags);
    } else if ((y & 0x7c00000000000000) == 0x7c00000000000000) {
        if ((y & 0x7e00000000000000) == 0x7e00000000000000) {
            flags |= 1;
            y = (y & 0xfdffffffffffffff);
            res = y;
        } else {
            res = x;
        }
        return (res, flags);
    }
    if (x == y) {
        res = x;
        return (res, flags);
    }
    if ((x & 0x7800000000000000) == 0x7800000000000000) {
        if (((x & 0x8000000000000000) == 0x8000000000000000) && ((y & 0x7800000000000000) == 0x7800000000000000)) {
            res = y;
        } else {
            res = x;
        }
        return (res, flags);
    } else if ((y & 0x7800000000000000) == 0x7800000000000000) {
        res = y;
        return (res, flags);
    }
    if ((x & 0x6000000000000000) == 0x6000000000000000) {
        exp_x = ((go_checked_shr_u64((x & 0x1ff8000000000000), go_shift_count_u64((51) as u64))) as i64);
        sig_x = ((x & 0x7ffffffffffff) | 0x20000000000000);
    } else {
        exp_x = ((go_checked_shr_u64((x & 0x7fe0000000000000), go_shift_count_u64((53) as u64))) as i64);
        sig_x = (x & 0x1fffffffffffff);
    }
    if ((y & 0x6000000000000000) == 0x6000000000000000) {
        exp_y = ((go_checked_shr_u64((y & 0x1ff8000000000000), go_shift_count_u64((51) as u64))) as i64);
        sig_y = ((y & 0x7ffffffffffff) | 0x20000000000000);
    } else {
        exp_y = ((go_checked_shr_u64((y & 0x7fe0000000000000), go_shift_count_u64((53) as u64))) as i64);
        sig_y = (y & 0x1fffffffffffff);
    }
    if (sig_x == 0) {
        res = y;
        return (res, flags);
    }
    if (sig_y == 0) {
        res = x;
        return (res, flags);
    }
    if ((sig_x > sig_y) && (exp_x >= exp_y)) {
        res = x;
        return (res, flags);
    }
    if ((sig_x < sig_y) && (exp_x <= exp_y)) {
        res = y;
        return (res, flags);
    }
    if ((exp_x.wrapping_sub(exp_y)) > 15) {
        res = x;
        return (res, flags);
    }
    if ((exp_y.wrapping_sub(exp_x)) > 15) {
        res = y;
        return (res, flags);
    }
    if (exp_x > exp_y) {
        sig_n_prime = __mul_64x64_to_128(sig_x, bid_mult_factor_minmax[(exp_x.wrapping_sub(exp_y)) as usize]);
        if ((sig_n_prime.w[1] == 0) && (sig_n_prime.w[0] == sig_y)) {
            if ((y & 0x8000000000000000) == 0x8000000000000000) {
                res = x;
            } else {
                res = y;
            }
            return (res, flags);
        }
        if ((sig_n_prime.w[1] != 0) || (sig_n_prime.w[0] > sig_y)) {
            res = x;
        } else {
            res = y;
        }
        return (res, flags);
    }
    sig_n_prime = __mul_64x64_to_128(sig_y, bid_mult_factor_minmax[(exp_y.wrapping_sub(exp_x)) as usize]);
    if ((sig_n_prime.w[1] == 0) && (sig_n_prime.w[0] == sig_x)) {
        if ((y & 0x8000000000000000) == 0x8000000000000000) {
            res = x;
        } else {
            res = y;
        }
        return (res, flags);
    }
    if ((sig_n_prime.w[1] == 0) && (sig_x > sig_n_prime.w[0])) {
        res = x;
    } else {
        res = y;
    }
    return (res, flags);
}
