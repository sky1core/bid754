// Auto-generated from noncomp64.go by go2rs. Do not edit.

use super::prelude::*;

pub fn bid64_is_signed(mut x: u64) -> i64 {
    if ((x & 0x8000000000000000) == 0x8000000000000000) {
        return 1;
    }
    return 0;
}

pub fn bid64_is_na_n(mut x: u64) -> i64 {
    if ((x & 0x7c00000000000000) == 0x7c00000000000000) {
        return 1;
    }
    return 0;
}

pub fn bid64_is_finite(mut x: u64) -> i64 {
    if ((x & 0x7800000000000000) != 0x7800000000000000) {
        return 1;
    }
    return 0;
}

pub fn bid64_is_inf(mut x: u64) -> i64 {
    if (((x & 0x7800000000000000) == 0x7800000000000000) && ((x & 0x7c00000000000000) != 0x7c00000000000000)) {
        return 1;
    }
    return 0;
}

pub fn bid64_is_signaling(mut x: u64) -> i64 {
    if ((x & 0x7e00000000000000) == 0x7e00000000000000) {
        return 1;
    }
    return 0;
}

pub fn bid64_is_canonical(mut x: u64) -> i64 {
    if ((x & 0x7c00000000000000) == 0x7c00000000000000) {
        if ((x & 0x01fc000000000000) != 0) {
            return 0;
        } else if ((x & 0x0003ffffffffffff) > 999999999999999) {
            return 0;
        } else {
            return 1;
        }
    } else if ((x & 0x7800000000000000) == 0x7800000000000000) {
        if ((x & 0x03ffffffffffffff) != 0) {
            return 0;
        } else {
            return 1;
        }
    } else if ((x & 0x6000000000000000) == 0x6000000000000000) {
        if ((((x & 0x7ffffffffffff) | 0x20000000000000)) <= 9999999999999999) {
            return 1;
        }
        return 0;
    } else {
        return 1;
    }
}

pub fn bid64_is_zero(mut x: u64) -> i64 {
    if ((x & 0x7800000000000000) == 0x7800000000000000) {
        return 0;
    } else if ((x & 0x6000000000000000) == 0x6000000000000000) {
        if ((((x & 0x7ffffffffffff) | 0x20000000000000)) > 9999999999999999) {
            return 1;
        }
        return 0;
    } else {
        if ((x & 0x1fffffffffffff) == 0) {
            return 1;
        }
        return 0;
    }
}

pub fn bid64_is_normal(mut x: u64) -> i64 {
    let mut sig_x_prime: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut sig_x: u64 = 0;
    let mut exp_x: u32 = 0;
    if ((x & 0x7800000000000000) == 0x7800000000000000) {
        return 0;
    }
    if ((x & 0x6000000000000000) == 0x6000000000000000) {
        sig_x = ((x & 0x7ffffffffffff) | 0x20000000000000);
        if ((sig_x > 9999999999999999) || (sig_x == 0)) {
            return 0;
        }
        exp_x = ((go_checked_shr_u64((x & 0x1ff8000000000000), go_shift_count_u64((51) as u64))) as u32);
    } else {
        sig_x = (x & 0x1fffffffffffff);
        if (sig_x == 0) {
            return 0;
        }
        exp_x = ((go_checked_shr_u64((x & 0x7fe0000000000000), go_shift_count_u64((53) as u64))) as u32);
    }
    if (exp_x < 15) {
        sig_x_prime = __mul_64x64_to_128(sig_x, bid_mult_factor[exp_x as usize]);
        if ((sig_x_prime.w[1] == 0) && (sig_x_prime.w[0] < 1000000000000000)) {
            return 0;
        }
        return 1;
    }
    return 1;
}

pub fn bid64_is_subnormal(mut x: u64) -> i64 {
    let mut sig_x_prime: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut sig_x: u64 = 0;
    let mut exp_x: u32 = 0;
    if ((x & 0x7800000000000000) == 0x7800000000000000) {
        return 0;
    }
    if ((x & 0x6000000000000000) == 0x6000000000000000) {
        sig_x = ((x & 0x7ffffffffffff) | 0x20000000000000);
        if ((sig_x > 9999999999999999) || (sig_x == 0)) {
            return 0;
        }
        exp_x = ((go_checked_shr_u64((x & 0x1ff8000000000000), go_shift_count_u64((51) as u64))) as u32);
    } else {
        sig_x = (x & 0x1fffffffffffff);
        if (sig_x == 0) {
            return 0;
        }
        exp_x = ((go_checked_shr_u64((x & 0x7fe0000000000000), go_shift_count_u64((53) as u64))) as u32);
    }
    if (exp_x < 15) {
        sig_x_prime = __mul_64x64_to_128(sig_x, bid_mult_factor[exp_x as usize]);
        if ((sig_x_prime.w[1] == 0) && (sig_x_prime.w[0] < 1000000000000000)) {
            return 1;
        }
        return 0;
    }
    return 0;
}

pub fn bid64_copy(mut x: u64) -> u64 {
    return x;
}

pub fn bid64_negate(mut x: u64) -> u64 {
    return (x ^ 0x8000000000000000);
}

pub fn bid64_abs(mut x: u64) -> u64 {
    return (x & (!0x8000000000000000));
}

pub fn bid64_copy_sign(mut x: u64, mut y: u64) -> u64 {
    return (((x & (!0x8000000000000000))) | (y & 0x8000000000000000));
}

pub fn bid64_same_quantum(mut x: u64, mut y: u64) -> i64 {
    let mut exp_x: u32 = 0;
    let mut exp_y: u32 = 0;
    if (((x & 0x7c00000000000000) == 0x7c00000000000000) || ((y & 0x7c00000000000000) == 0x7c00000000000000)) {
        if (((x & 0x7c00000000000000) == 0x7c00000000000000) && ((y & 0x7c00000000000000) == 0x7c00000000000000)) {
            return 1;
        }
        return 0;
    }
    if (((x & 0x7800000000000000) == 0x7800000000000000) || ((y & 0x7800000000000000) == 0x7800000000000000)) {
        if (((x & 0x7800000000000000) == 0x7800000000000000) && ((y & 0x7800000000000000) == 0x7800000000000000)) {
            return 1;
        }
        return 0;
    }
    if ((x & 0x6000000000000000) == 0x6000000000000000) {
        exp_x = ((go_checked_shr_u64((x & 0x1ff8000000000000), go_shift_count_u64((51) as u64))) as u32);
    } else {
        exp_x = ((go_checked_shr_u64((x & 0x7fe0000000000000), go_shift_count_u64((53) as u64))) as u32);
    }
    if ((y & 0x6000000000000000) == 0x6000000000000000) {
        exp_y = ((go_checked_shr_u64((y & 0x1ff8000000000000), go_shift_count_u64((51) as u64))) as u32);
    } else {
        exp_y = ((go_checked_shr_u64((y & 0x7fe0000000000000), go_shift_count_u64((53) as u64))) as u32);
    }
    if (exp_x == exp_y) {
        return 1;
    }
    return 0;
}

pub fn bid64_radix() -> i64 {
    return 10;
}

pub fn bid64_class(mut x: u64) -> i64 {
    let mut sig_x_prime: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut sig_x: u64 = 0;
    let mut exp_x: i64 = 0;
    if ((x & 0x7c00000000000000) == 0x7c00000000000000) {
        if ((x & 0x7e00000000000000) == 0x7e00000000000000) {
            return 0;
        }
        return 1;
    } else if ((x & 0x7800000000000000) == 0x7800000000000000) {
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            return 2;
        }
        return 9;
    } else if ((x & 0x6000000000000000) == 0x6000000000000000) {
        sig_x = ((x & 0x7ffffffffffff) | 0x20000000000000);
        if ((sig_x > 9999999999999999) || (sig_x == 0)) {
            if ((x & 0x8000000000000000) == 0x8000000000000000) {
                return 5;
            }
            return 6;
        }
        exp_x = ((go_checked_shr_u64((x & 0x1ff8000000000000), go_shift_count_u64((51) as u64))) as i64);
    } else {
        sig_x = (x & 0x1fffffffffffff);
        if (sig_x == 0) {
            if ((x & 0x8000000000000000) == 0x8000000000000000) {
                return 5;
            }
            return 6;
        }
        exp_x = ((go_checked_shr_u64((x & 0x7fe0000000000000), go_shift_count_u64((53) as u64))) as i64);
    }
    if (exp_x < 15) {
        sig_x_prime = __mul_64x64_to_128(sig_x, bid_mult_factor[exp_x as usize]);
        if ((sig_x_prime.w[1] == 0) && (sig_x_prime.w[0] < 1000000000000000)) {
            if ((x & 0x8000000000000000) == 0x8000000000000000) {
                return 4;
            }
            return 7;
        }
    }
    if ((x & 0x8000000000000000) == 0x8000000000000000) {
        return 3;
    }
    return 8;
}

pub fn bid64_inf() -> u64 {
    return 0x7800000000000000;
}

pub fn bid64_na_n() -> u64 {
    return 0x7c00000000000000;
}

pub fn bid64_total_order(mut x: u64, mut y: u64) -> i64 {
    let mut res: i64 = 0;
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut sig_x: u64 = 0;
    let mut sig_y: u64 = 0;
    let mut pyld_y: u64 = 0;
    let mut pyld_x: u64 = 0;
    let mut sig_n_prime: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut x_is_zero: i64 = 0;
    let mut y_is_zero: i64 = 0;
    if ((x & 0x7c00000000000000) == 0x7c00000000000000) {
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            if (((y & 0x7c00000000000000) != 0x7c00000000000000) || ((y & 0x8000000000000000) != 0x8000000000000000)) {
                res = 1;
                return res;
            }
            let mut xIsSNaN = ((x & 0x7e00000000000000) == 0x7e00000000000000);
            let mut yIsSNaN = ((y & 0x7e00000000000000) == 0x7e00000000000000);
            if (xIsSNaN == yIsSNaN) {
                pyld_y = (y & 0x0003ffffffffffff);
                pyld_x = (x & 0x0003ffffffffffff);
                if ((pyld_y > 999999999999999) || (pyld_y == 0)) {
                    res = 1;
                    return res;
                }
                if ((pyld_x > 999999999999999) || (pyld_x == 0)) {
                    res = 0;
                    return res;
                }
                if (pyld_x >= pyld_y) {
                    res = 1;
                } else {
                    res = 0;
                }
                return res;
            }
            if ((y & 0x7e00000000000000) == 0x7e00000000000000) {
                res = 1;
            } else {
                res = 0;
            }
            return res;
        }
        if (((y & 0x7c00000000000000) != 0x7c00000000000000) || ((y & 0x8000000000000000) == 0x8000000000000000)) {
            res = 0;
            return res;
        }
        let mut xIsSNaN = ((x & 0x7e00000000000000) == 0x7e00000000000000);
        let mut yIsSNaN = ((y & 0x7e00000000000000) == 0x7e00000000000000);
        if (xIsSNaN == yIsSNaN) {
            pyld_y = (y & 0x0003ffffffffffff);
            pyld_x = (x & 0x0003ffffffffffff);
            if ((pyld_x > 999999999999999) || (pyld_x == 0)) {
                res = 1;
                return res;
            }
            if ((pyld_y > 999999999999999) || (pyld_y == 0)) {
                res = 0;
                return res;
            }
            if (pyld_x <= pyld_y) {
                res = 1;
            } else {
                res = 0;
            }
            return res;
        }
        if ((x & 0x7e00000000000000) == 0x7e00000000000000) {
            res = 1;
        } else {
            res = 0;
        }
        return res;
    } else if ((y & 0x7c00000000000000) == 0x7c00000000000000) {
        if ((y & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        } else {
            res = 0;
        }
        return res;
    }
    if (x == y) {
        res = 1;
        return res;
    }
    if ((((x & 0x8000000000000000) == 0x8000000000000000)) != (((y & 0x8000000000000000) == 0x8000000000000000))) {
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        } else {
            res = 0;
        }
        return res;
    }
    if ((x & 0x7800000000000000) == 0x7800000000000000) {
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
            return res;
        }
        if ((y & 0x7800000000000000) == 0x7800000000000000) {
            res = 1;
        } else {
            res = 0;
        }
        return res;
    } else if ((y & 0x7800000000000000) == 0x7800000000000000) {
        if ((y & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        } else {
            res = 0;
        }
        return res;
    }
    if ((x & 0x6000000000000000) == 0x6000000000000000) {
        exp_x = ((go_checked_shr_u64((x & 0x1ff8000000000000), go_shift_count_u64((51) as u64))) as i64);
        sig_x = ((x & 0x7ffffffffffff) | 0x20000000000000);
        if ((sig_x > 9999999999999999) || (sig_x == 0)) {
            x_is_zero = 1;
        }
    } else {
        exp_x = ((go_checked_shr_u64((x & 0x7fe0000000000000), go_shift_count_u64((53) as u64))) as i64);
        sig_x = (x & 0x1fffffffffffff);
        if (sig_x == 0) {
            x_is_zero = 1;
        }
    }
    if ((y & 0x6000000000000000) == 0x6000000000000000) {
        exp_y = ((go_checked_shr_u64((y & 0x1ff8000000000000), go_shift_count_u64((51) as u64))) as i64);
        sig_y = ((y & 0x7ffffffffffff) | 0x20000000000000);
        if ((sig_y > 9999999999999999) || (sig_y == 0)) {
            y_is_zero = 1;
        }
    } else {
        exp_y = ((go_checked_shr_u64((y & 0x7fe0000000000000), go_shift_count_u64((53) as u64))) as i64);
        sig_y = (y & 0x1fffffffffffff);
        if (sig_y == 0) {
            y_is_zero = 1;
        }
    }
    if ((x_is_zero != 0) && (y_is_zero != 0)) {
        if ((((x & 0x8000000000000000) == 0x8000000000000000)) == (((y & 0x8000000000000000) == 0x8000000000000000))) {
            if (exp_x == exp_y) {
                res = 1;
                return res;
            }
            if ((exp_x <= exp_y) != (((x & 0x8000000000000000) == 0x8000000000000000))) {
                res = 1;
            } else {
                res = 0;
            }
            return res;
        }
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        } else {
            res = 0;
        }
        return res;
    }
    if (x_is_zero != 0) {
        if ((y & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        } else {
            res = 0;
        }
        return res;
    }
    if (y_is_zero != 0) {
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        } else {
            res = 0;
        }
        return res;
    }
    if ((sig_x > sig_y) && (exp_x >= exp_y)) {
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        } else {
            res = 0;
        }
        return res;
    }
    if ((sig_x < sig_y) && (exp_x <= exp_y)) {
        if ((x & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        } else {
            res = 0;
        }
        return res;
    }
    if ((exp_x.wrapping_sub(exp_y)) > 15) {
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = 1;
        } else {
            res = 0;
        }
        return res;
    }
    if ((exp_y.wrapping_sub(exp_x)) > 15) {
        if ((x & 0x8000000000000000) != 0x8000000000000000) {
            res = 1;
        } else {
            res = 0;
        }
        return res;
    }
    if (exp_x > exp_y) {
        sig_n_prime = __mul_64x64_to_128(sig_x, bid_mult_factor[(exp_x.wrapping_sub(exp_y)) as usize]);
        if ((sig_n_prime.w[1] == 0) && (sig_n_prime.w[0] == sig_y)) {
            if ((exp_x <= exp_y) != (((x & 0x8000000000000000) == 0x8000000000000000))) {
                res = 1;
            } else {
                res = 0;
            }
            return res;
        }
        let mut cond = ((sig_n_prime.w[1] == 0) && (sig_n_prime.w[0] < sig_y));
        let mut sign = ((x & 0x8000000000000000) == 0x8000000000000000);
        if (cond != sign) {
            res = 1;
        } else {
            res = 0;
        }
        return res;
    }
    sig_n_prime = __mul_64x64_to_128(sig_y, bid_mult_factor[(exp_y.wrapping_sub(exp_x)) as usize]);
    if ((sig_n_prime.w[1] == 0) && (sig_n_prime.w[0] == sig_x)) {
        if ((exp_x <= exp_y) != (((x & 0x8000000000000000) == 0x8000000000000000))) {
            res = 1;
        } else {
            res = 0;
        }
        return res;
    }
    let mut cond = ((sig_n_prime.w[1] > 0) || (sig_x < sig_n_prime.w[0]));
    let mut sign = ((x & 0x8000000000000000) == 0x8000000000000000);
    if (cond != sign) {
        res = 1;
    } else {
        res = 0;
    }
    return res;
}

pub fn bid64_total_order_mag(mut x: u64, mut y: u64) -> i64 {
    let mut res: i64 = 0;
    let mut exp_x: i64 = 0;
    let mut exp_y: i64 = 0;
    let mut sig_x: u64 = 0;
    let mut sig_y: u64 = 0;
    let mut pyld_y: u64 = 0;
    let mut pyld_x: u64 = 0;
    let mut sig_n_prime: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut x_is_zero: i64 = 0;
    let mut y_is_zero: i64 = 0;
    if ((x & 0x7c00000000000000) == 0x7c00000000000000) {
        if ((y & 0x7c00000000000000) != 0x7c00000000000000) {
            res = 0;
            return res;
        }
        let mut xIsSNaN = ((x & 0x7e00000000000000) == 0x7e00000000000000);
        let mut yIsSNaN = ((y & 0x7e00000000000000) == 0x7e00000000000000);
        if (xIsSNaN == yIsSNaN) {
            pyld_y = (y & 0x0003ffffffffffff);
            pyld_x = (x & 0x0003ffffffffffff);
            if ((pyld_x > 999999999999999) || (pyld_x == 0)) {
                res = 1;
                return res;
            }
            if ((pyld_y > 999999999999999) || (pyld_y == 0)) {
                res = 0;
                return res;
            }
            if (pyld_x <= pyld_y) {
                res = 1;
            } else {
                res = 0;
            }
            return res;
        }
        if ((x & 0x7e00000000000000) == 0x7e00000000000000) {
            res = 1;
        } else {
            res = 0;
        }
        return res;
    } else if ((y & 0x7c00000000000000) == 0x7c00000000000000) {
        res = 1;
        return res;
    }
    if (((x & (!0x8000000000000000))) == ((y & (!0x8000000000000000)))) {
        res = 1;
        return res;
    }
    if ((x & 0x7800000000000000) == 0x7800000000000000) {
        if ((y & 0x7800000000000000) == 0x7800000000000000) {
            res = 1;
        } else {
            res = 0;
        }
        return res;
    } else if ((y & 0x7800000000000000) == 0x7800000000000000) {
        res = 1;
        return res;
    }
    if ((x & 0x6000000000000000) == 0x6000000000000000) {
        exp_x = ((go_checked_shr_u64((x & 0x1ff8000000000000), go_shift_count_u64((51) as u64))) as i64);
        sig_x = ((x & 0x7ffffffffffff) | 0x20000000000000);
        if ((sig_x > 9999999999999999) || (sig_x == 0)) {
            x_is_zero = 1;
        }
    } else {
        exp_x = ((go_checked_shr_u64((x & 0x7fe0000000000000), go_shift_count_u64((53) as u64))) as i64);
        sig_x = (x & 0x1fffffffffffff);
        if (sig_x == 0) {
            x_is_zero = 1;
        }
    }
    if ((y & 0x6000000000000000) == 0x6000000000000000) {
        exp_y = ((go_checked_shr_u64((y & 0x1ff8000000000000), go_shift_count_u64((51) as u64))) as i64);
        sig_y = ((y & 0x7ffffffffffff) | 0x20000000000000);
        if ((sig_y > 9999999999999999) || (sig_y == 0)) {
            y_is_zero = 1;
        }
    } else {
        exp_y = ((go_checked_shr_u64((y & 0x7fe0000000000000), go_shift_count_u64((53) as u64))) as i64);
        sig_y = (y & 0x1fffffffffffff);
        if (sig_y == 0) {
            y_is_zero = 1;
        }
    }
    if ((x_is_zero != 0) && (y_is_zero != 0)) {
        if (exp_x <= exp_y) {
            res = 1;
        } else {
            res = 0;
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
    if ((sig_x > sig_y) && (exp_x >= exp_y)) {
        res = 0;
        return res;
    }
    if ((sig_x < sig_y) && (exp_x <= exp_y)) {
        res = 1;
        return res;
    }
    if ((exp_x.wrapping_sub(exp_y)) > 15) {
        res = 0;
        return res;
    }
    if ((exp_y.wrapping_sub(exp_x)) > 15) {
        res = 1;
        return res;
    }
    if (exp_x > exp_y) {
        sig_n_prime = __mul_64x64_to_128(sig_x, bid_mult_factor[(exp_x.wrapping_sub(exp_y)) as usize]);
        if ((sig_n_prime.w[1] == 0) && (sig_n_prime.w[0] == sig_y)) {
            res = 0;
            return res;
        }
        if ((sig_n_prime.w[1] == 0) && (sig_n_prime.w[0] < sig_y)) {
            res = 1;
        } else {
            res = 0;
        }
        return res;
    }
    sig_n_prime = __mul_64x64_to_128(sig_y, bid_mult_factor[(exp_y.wrapping_sub(exp_x)) as usize]);
    if ((sig_n_prime.w[1] == 0) && (sig_n_prime.w[0] == sig_x)) {
        res = 1;
        return res;
    }
    if ((sig_n_prime.w[1] > 0) || (sig_x < sig_n_prime.w[0])) {
        res = 1;
    } else {
        res = 0;
    }
    return res;
}

pub fn bid64_quantum(mut x: u64) -> u64 {
    if ((x & 0x7800000000000000) == 0x7800000000000000) {
        return (x & (!0x8000000000000000));
    }
    if ((x & 0x7c00000000000000) == 0x7c00000000000000) {
        return (x & 0x7dffffffffffffff);
    }
    let mut intExp: i64 = 0;
    if ((x & 0x6000000000000000) == 0x6000000000000000) {
        intExp = (((((go_checked_shr_u64(x, go_shift_count_u64((51) as u64)))) & 0x3ff) as i64).wrapping_sub(398));
    } else {
        intExp = (((((go_checked_shr_u64(x, go_shift_count_u64((53) as u64)))) & 0x3ff) as i64).wrapping_sub(398));
    }
    return (((go_checked_shl_u64((intExp as u64), go_shift_count_u64((53) as u64)))).wrapping_add(0x31c0000000000001));
}

pub fn bid64_quantexp(mut x: u64) -> (i32, u32) {
    if ((((x & 0x7800000000000000) == 0x7800000000000000)) || (((x & 0x7c00000000000000) == 0x7c00000000000000))) {
        return ((-2147483648), 1);
    }
    if ((x & 0x6000000000000000) == 0x6000000000000000) {
        return ((((((go_checked_shr_u64(x, go_shift_count_u64((51) as u64)))) & 0x3ff) as i32).wrapping_sub(398)), 0);
    }
    return ((((((go_checked_shr_u64(x, go_shift_count_u64((53) as u64)))) & 0x3ff) as i32).wrapping_sub(398)), 0);
}

pub fn bid64_ll_quantexp(mut x: u64) -> (i64, u32) {
    if ((((x & 0x7800000000000000) == 0x7800000000000000)) || (((x & 0x7c00000000000000) == 0x7c00000000000000))) {
        return ((-9223372036854775808), 1);
    }
    if ((x & 0x6000000000000000) == 0x6000000000000000) {
        return ((((((go_checked_shr_u64(x, go_shift_count_u64((51) as u64)))) & 0x3ff) as i64).wrapping_sub(398)), 0);
    }
    return ((((((go_checked_shr_u64(x, go_shift_count_u64((53) as u64)))) & 0x3ff) as i64).wrapping_sub(398)), 0);
}

pub fn bid64_signaling_less(mut x: u64, mut y: u64) -> (i64, u32) {
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
            if ((((y & 0x7800000000000000) != 0x7800000000000000)) || (((y & 0x8000000000000000) != 0x8000000000000000))) {
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
