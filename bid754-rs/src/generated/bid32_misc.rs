// Auto-generated from bid32_misc.go by go2rs. Do not edit.

use super::prelude::*;

pub fn bid32_nearby_int(mut x: u32, mut rnd_mode: i64) -> (u32, u32) {
    let (mut x64, mut f1) = bid32_to_bid64(x);
    let (mut res64, mut f2) = bid64_nearby_int(x64, rnd_mode);
    let (mut res, mut f3) = bid64_to_bid32(res64, 0);
    return (res, ((f1 | f2) | f3));
}

pub fn bid32_fdim(mut x: u32, mut y: u32, mut rnd_mode: i64) -> (u32, u32) {
    let mut res: u32 = 0;
    let mut pfpsf: u32 = 0;
    let (mut cmpres, _) = bid32_quiet_greater(x, y);
    if (((((x & 0x7c000000) != 0x7c000000)) && (((y & 0x7c000000) != 0x7c000000))) && (cmpres == 0)) {
        res = 0x32800000;
        return (res, pfpsf);
    }
    return bid32_sub_with_flags(x, y, rnd_mode);
}

pub fn bid32_scalbln(mut x: u32, mut n: i64, mut rnd_mode: i64) -> (u32, u32) {
    let mut n1 = (n as i32);
    if ((n1 as i64) < n) {
        n1 = 0x7fffffff;
    } else if ((n1 as i64) > n) {
        n1 = (-0x80000000);
    }
    return bid32_scalbn(x, (n1 as i64), rnd_mode);
}

pub fn bid32_modf(mut x: u32) -> (u32, u32, u32) {
    let (mut x64, mut f0) = bid32_to_bid64(x);
    let (mut frac64, mut iptr64, mut flags) = bid64_modf(x64);
    let (mut frac, mut f1) = bid64_to_bid32(frac64, 0);
    let (mut iptr, mut f2) = bid64_to_bid32(iptr64, 0);
    return (frac, iptr, (((f0 | flags) | f1) | f2));
}

pub fn bid32_ldexp(mut x: u32, mut n: i64, mut rnd_mode: i64) -> (u32, u32) {
    return bid32_scalbn(x, n, rnd_mode);
}

pub fn bid32_frexp(mut x: u32) -> (u32, i64, u32) {
    let mut sig_x: u32 = 0;
    let mut res: u32 = 0;
    let mut exp_x: u32 = 0;
    let mut pfpsf: u32 = 0;
    if ((x & 0x78000000) == 0x78000000) {
        res = x;
        if ((x & 0x7e000000) == 0x7e000000) {
            res = (x & 0xfdffffff);
        }
        return (res, 0, pfpsf);
    }
    if ((x & 0x60000000) == 0x60000000) {
        exp_x = (go_checked_shr_u32((x & 0x1fe00000), go_shift_count_u64((21) as u64)));
        sig_x = ((x & 0x1fffff) | 0x800000);
        if ((sig_x > 9999999) || (sig_x == 0)) {
            res = ((x & 0x80000000) | ((go_checked_shl_u32(exp_x, go_shift_count_u64((23) as u64)))));
            return (res, 0, pfpsf);
        }
    } else {
        exp_x = (go_checked_shr_u32((x & 0x7f800000), go_shift_count_u64((23) as u64)));
        sig_x = (x & 0x7fffff);
        if (sig_x == 0) {
            res = ((x & 0x80000000) | ((go_checked_shl_u32(exp_x, go_shift_count_u64((23) as u64)))));
            return (res, 0, pfpsf);
        }
    }
    let mut tmp = ((sig_x as f32) as f32).to_bits();
    let mut x_nr_bits = (((1 as i64).wrapping_add(((((go_checked_shr_u32(tmp, go_shift_count_u64((23) as u64)))) & 0xff) as i64))).wrapping_sub(0x7f));
    let mut q = (bid_nr_digits[(x_nr_bits.wrapping_sub(1)) as usize].digits as i64);
    if (q == 0) {
        q = (bid_nr_digits[(x_nr_bits.wrapping_sub(1)) as usize].digits1 as i64);
        if ((sig_x as u64) >= bid_nr_digits[(x_nr_bits.wrapping_sub(1)) as usize].threshold_lo) {
            q = q.wrapping_add(1);
        }
    }
    let mut exp = (((exp_x as i64).wrapping_sub(101)).wrapping_add(q));
    if (sig_x < 0x00800000) {
        res = ((x & 0x807fffff) | ((go_checked_shl_u32((((q.wrapping_neg()).wrapping_add(101)) as u32), go_shift_count_u64((23) as u64)))));
    } else {
        res = ((x & 0xe01fffff) | ((go_checked_shl_u32((((q.wrapping_neg()).wrapping_add(101)) as u32), go_shift_count_u64((21) as u64)))));
    }
    return (res, exp, pfpsf);
}

pub fn bid32_fmod(mut x: u32, mut y: u32) -> (u32, u32) {
    let mut CX: u64 = 0;
    let mut Q64: u64 = 0;
    let mut CYL: u64 = 0;
    let mut CY: u32 = 0;
    let mut sign_x: u32 = 0;
    let mut sign_y: u32 = 0;
    let mut coefficient_x: u32 = 0;
    let mut coefficient_y: u32 = 0;
    let mut res: u32 = 0;
    let mut Q: u32 = 0;
    let mut R: u32 = 0;
    let mut T: u32 = 0;
    let mut exponent_x: i64 = 0;
    let mut exponent_y: i64 = 0;
    let mut bin_expon: i64 = 0;
    let mut e_scale: i64 = 0;
    let mut digits_x: i64 = 0;
    let mut diff_expon: i64 = 0;
    let mut pfpsf: u32 = 0;
    let (mut sign_y, mut exponent_y, mut coefficient_y, mut valid_y) = unpack_bid32(y);
    let (mut sign_x, mut exponent_x, mut coefficient_x, mut valid_x) = unpack_bid32(x);
    _ = sign_y;
    if (!valid_x) {
        if ((y & 0x7e000000) == 0x7e000000) {
            pfpsf |= 1;
        }
        if ((x & 0x7c000000) == 0x7c000000) {
            if ((x & 0x7e000000) == 0x7e000000) {
                pfpsf |= 1;
            }
            res = (coefficient_x & 0xfdffffff);
            return (res, pfpsf);
        }
        if ((x & 0x78000000) == 0x78000000) {
            if ((y & 0x7c000000) != 0x7c000000) {
                pfpsf |= 1;
                return (0x7c000000, pfpsf);
            }
        }
        if ((((y & 0x78000000) < 0x78000000)) && (coefficient_y != 0)) {
            if ((y & 0x60000000) == 0x60000000) {
                exponent_y = ((((go_checked_shr_u32(y, go_shift_count_u64((21) as u64)))) & 0xff) as i64);
            } else {
                exponent_y = ((((go_checked_shr_u32(y, go_shift_count_u64((23) as u64)))) & 0xff) as i64);
            }
            if (exponent_y < exponent_x) {
                exponent_x = exponent_y;
            }
            res = (((go_checked_shl_u32((exponent_x as u32), go_shift_count_u64((23) as u64)))) | sign_x);
            return (res, pfpsf);
        }
    }
    if (!valid_y) {
        if ((y & 0x7c000000) == 0x7c000000) {
            if ((y & 0x7e000000) == 0x7e000000) {
                pfpsf |= 1;
            }
            res = (coefficient_y & 0xfdffffff);
            return (res, pfpsf);
        }
        if ((y & 0x78000000) == 0x78000000) {
            res = very_fast_get_bid32(sign_x, exponent_x, coefficient_x);
            return (res, pfpsf);
        }
        pfpsf |= 1;
        return (0x7c000000, pfpsf);
    }
    diff_expon = (exponent_x.wrapping_sub(exponent_y));
    if (diff_expon <= 0) {
        diff_expon = (diff_expon.wrapping_neg());
        if (diff_expon > 7) {
            res = x;
            return (res, pfpsf);
        }
        T = (bid_power10_table_128[diff_expon as usize].w[0] as u32);
        CYL = ((coefficient_y as u64).wrapping_mul(T as u64));
        if (CYL > (coefficient_x as u64)) {
            res = x;
            return (res, pfpsf);
        }
        CY = (CYL as u32);
        Q = (coefficient_x / CY);
        R = (coefficient_x.wrapping_sub((Q.wrapping_mul(CY))));
        res = very_fast_get_bid32(sign_x, exponent_x, R);
        return (res, pfpsf);
    }
    CX = (coefficient_x as u64);
    while (diff_expon > 0) {
        let mut tempx = ((CX as f32) as f32).to_bits();
        bin_expon = (((((go_checked_shr_u32(tempx, go_shift_count_u64((23) as u64)))) & 0xff) as i64).wrapping_sub(0x7f));
        digits_x = (bid_estimate_decimal_digits[bin_expon as usize] as i64);
        e_scale = ((18 as i64).wrapping_sub(digits_x));
        if (diff_expon >= e_scale) {
            diff_expon = diff_expon.wrapping_sub(e_scale);
        } else {
            e_scale = diff_expon;
            diff_expon = 0;
        }
        CX = CX.wrapping_mul(bid_power10_table_128[e_scale as usize].w[0]);
        Q64 = (CX / (coefficient_y as u64));
        CX = CX.wrapping_sub((Q64.wrapping_mul(coefficient_y as u64)));
        if (CX == 0) {
            res = very_fast_get_bid32(sign_x, exponent_y, 0);
            return (res, pfpsf);
        }
    }
    res = very_fast_get_bid32(sign_x, exponent_y, (CX as u32));
    return (res, pfpsf);
}

pub fn bid32_class(mut x: u32) -> i64 {
    let mut sig_x: u32 = 0;
    let mut exp_x: i64 = 0;
    if ((x & 0x7c000000) == 0x7c000000) {
        if ((x & 0x7e000000) == 0x7e000000) {
            return 0;
        }
        return 1;
    }
    if ((x & 0x78000000) == 0x78000000) {
        if ((x & 0x80000000) == 0x80000000) {
            return 2;
        }
        return 9;
    }
    if ((x & 0x60000000) == 0x60000000) {
        exp_x = ((go_checked_shr_u32((x & 0x1fe00000), go_shift_count_u64((21) as u64))) as i64);
        sig_x = ((x & 0x1fffff) | 0x800000);
        if (sig_x > 9999999) {
            sig_x = 0;
        }
    } else {
        exp_x = ((go_checked_shr_u32((x & 0x7f800000), go_shift_count_u64((23) as u64))) as i64);
        sig_x = (x & 0x7fffff);
    }
    if (sig_x == 0) {
        if ((x & 0x80000000) == 0x80000000) {
            return 5;
        }
        return 6;
    }
    if (exp_x < 6) {
        let mut sig_x_prime = ((sig_x as u64).wrapping_mul(bid32_mult_factor[exp_x as usize]));
        if (sig_x_prime < 1000000) {
            if ((x & 0x80000000) == 0x80000000) {
                return 4;
            }
            return 7;
        }
    }
    if ((x & 0x80000000) == 0x80000000) {
        return 3;
    }
    return 8;
}

pub fn bid32_quantexp(mut x: u32) -> (i32, u32) {
    let mut pfpsf: u32 = 0;
    if ((x & 0x78000000) == 0x78000000) {
        pfpsf |= 1;
        return ((-0x80000000), pfpsf);
    }
    let mut exp: i64 = 0;
    if ((x & 0x60000000) == 0x60000000) {
        exp = ((go_checked_shr_u32((x & 0x1fe00000), go_shift_count_u64((21) as u64))) as i64);
    } else {
        exp = ((go_checked_shr_u32((x & 0x7f800000), go_shift_count_u64((23) as u64))) as i64);
    }
    return (((exp.wrapping_sub(101)) as i32), pfpsf);
}

pub fn bid32_ll_quantexp(mut x: u32) -> (i64, u32) {
    let mut pfpsf: u32 = 0;
    if ((x & 0x78000000) == 0x78000000) {
        pfpsf |= 1;
        return ((-0x8000000000000000), pfpsf);
    }
    let mut exp: i64 = 0;
    if ((x & 0x60000000) == 0x60000000) {
        exp = ((go_checked_shr_u32((x & 0x1fe00000), go_shift_count_u64((21) as u64))) as i64);
    } else {
        exp = ((go_checked_shr_u32((x & 0x7f800000), go_shift_count_u64((23) as u64))) as i64);
    }
    return (((exp.wrapping_sub(101)) as i64), pfpsf);
}

pub fn bid32_inf() -> u32 {
    return 0x78000000;
}

pub fn bid32_na_n() -> u32 {
    return 0x7c000000;
}

pub fn bid32_next_toward(mut x: u32, mut y: BID_UINT128) -> (u32, u32) {
    let mut res: u32 = 0;
    let mut x128: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut tmp128: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut tmp1: u32 = 0;
    let mut tmp2: u32 = 0;
    let mut tmp_fpsf: u32 = 0;
    let mut res1: i64 = 0;
    let mut res2: i64 = 0;
    let mut flags: u32 = 0;
    if (((((x & 0x78000000) == 0x78000000)) || (((y.w[1] & 0x7c00000000000000) == 0x7c00000000000000))) || (((y.w[1] & 0x7c00000000000000) == 0x7800000000000000))) {
        if ((x & 0x7c000000) == 0x7c000000) {
            if ((x & 0x000fffff) > 999999) {
                x = (x & 0xfe000000);
            } else {
                x = (x & 0xfe0fffff);
            }
            if ((x & 0x7e000000) == 0x7e000000) {
                flags |= 1;
                res = (x & 0xfdffffff);
            } else {
                if ((y.w[1] & 0x7e00000000000000) == 0x7e00000000000000) {
                    flags |= 1;
                }
                res = x;
            }
            return (res, flags);
        } else if ((y.w[1] & 0x7c00000000000000) == 0x7c00000000000000) {
            if ((((y.w[1] & 0x00003fffffffffff) > 0x0000314dc6448d93)) || ((((y.w[1] & 0x00003fffffffffff) == 0x0000314dc6448d93) && (y.w[0] > 0x38c15b09ffffffff)))) {
                y.w[1] = (y.w[1] & 0xffffc00000000000);
                y.w[0] = 0;
            }
            if ((y.w[1] & 0x7e00000000000000) == 0x7e00000000000000) {
                flags |= 1;
                tmp128.w[1] = (y.w[1] & 0xfc003fffffffffff);
                tmp128.w[0] = y.w[0];
            } else {
                tmp128.w[1] = (y.w[1] & 0xfc003fffffffffff);
                tmp128.w[0] = y.w[0];
            }
            (res, _) = bid128_to_bid32(tmp128, 0);
            return (res, flags);
        } else {
            if ((x & 0x78000000) == 0x78000000) {
                x = (x & (0x80000000 | 0x78000000));
            }
            if ((y.w[1] & 0x7c00000000000000) == 0x7800000000000000) {
                y.w[1] = (y.w[1] & (0x8000000000000000 | 0x7800000000000000));
                y.w[0] = 0;
            }
        }
    }
    if ((x & 0x78000000) != 0x78000000) {
        if ((x & 0x60000000) == 0x60000000) {
            if ((((x & 0x1fffff) | 0x800000)) > 9999999) {
                x = ((x & 0x80000000) | ((go_checked_shl_u32((x & 0x1fe00000), go_shift_count_u64((2) as u64)))));
            }
        }
    }
    tmp_fpsf = flags;
    (x128, _) = bid32_to_bid128(x);
    (res1, _) = bid128_quiet_equal(x128, y);
    (res2, _) = bid128_quiet_greater(x128, y);
    flags = tmp_fpsf;
    if (res1 != 0) {
        res = (((go_checked_shr_u64((y.w[1] & 0x8000000000000000), go_shift_count_u64((32) as u64))) as u32) | (x & 0x7fffffff));
    } else if (res2 != 0) {
        (res, _) = bid32_next_down(x);
    } else {
        (res, _) = bid32_next_up(x);
    }
    if ((((x & 0x78000000) != 0x78000000)) && (((res & 0x78000000) == 0x78000000))) {
        flags |= 32;
        flags |= 8;
    }
    tmp1 = 0x0f4240;
    tmp2 = (res & 0x7fffffff);
    tmp_fpsf = flags;
    (res1, _) = bid32_quiet_greater(tmp1, tmp2);
    (res2, _) = bid32_quiet_not_equal(x, res);
    flags = tmp_fpsf;
    if ((res1 != 0) && (res2 != 0)) {
        flags |= 32;
        flags |= 16;
    }
    return (res, flags);
}

pub fn bid32_to_binary32(mut x: u32, mut rndMode: i64) -> (u32, u32) {
    if ((x & 0x7c000000) == 0x7c000000) {
        let mut flags: u32 = 0;
        let mut payload = (x & 0x000fffff);
        if (payload > 999999) {
            payload = 0;
        }
        if ((x & 0x7e000000) == 0x7e000000) {
            flags |= 1;
        }
        return (((((x & 0x80000000) | 0x7f800000) | 0x00400000) | ((((go_checked_shl_u32(payload, go_shift_count_u64((2) as u64)))) & 0x003fffff))), flags);
    }
    let (mut x64, mut f0) = bid32_to_bid64(x);
    let (mut res, mut f1) = bid64_to_binary32(x64, rndMode);
    return (res, (f0 | f1));
}

pub fn bid32_to_binary64(mut x: u32, mut rndMode: i64) -> (u64, u32) {
    if ((x & 0x7c000000) == 0x7c000000) {
        let mut flags: u32 = 0;
        let mut payload = ((x & 0x000fffff) as u64);
        if (payload > 999999) {
            payload = 0;
        }
        if ((x & 0x7e000000) == 0x7e000000) {
            flags |= 1;
        }
        return ((((((go_checked_shl_u64(((x & 0x80000000) as u64), go_shift_count_u64((32) as u64)))) | 0x7ff0000000000000) | 0x0008000000000000) | ((((go_checked_shl_u64(payload, go_shift_count_u64((31) as u64)))) & 0x0007ffffffffffff))), flags);
    }
    let (mut x64, mut f0) = bid32_to_bid64(x);
    let (mut res, mut f1) = bid64_to_binary64(x64, rndMode);
    return (res, (f0 | f1));
}

pub fn bid32_to_binary128(mut x: u32, mut rndMode: i64) -> (BID_UINT128, u32) {
    if ((x & 0x7c000000) == 0x7c000000) {
        let mut res: BID_UINT128 = BID_UINT128 { w: [0, 0] };
        let mut flags: u32 = 0;
        let mut payload = ((x & 0x000fffff) as u64);
        if (payload > 999999) {
            payload = 0;
        }
        if ((x & 0x7e000000) == 0x7e000000) {
            flags |= 1;
        }
        res.w[1] = (((((go_checked_shl_u64(((x & 0x80000000) as u64), go_shift_count_u64((32) as u64)))) | 0x7fff000000000000) | 0x0000800000000000) | ((((go_checked_shl_u64(payload, go_shift_count_u64((27) as u64)))) & 0x00007fffffffffff)));
        res.w[0] = 0;
        return (res, flags);
    }
    let (mut x64, mut f0) = bid32_to_bid64(x);
    let (mut res, mut f1) = bid64_to_binary128(x64, rndMode);
    return (res, (f0 | f1));
}

pub fn bid32_to_bid128(mut x: u32) -> (BID_UINT128, u32) {
    let (mut x64, mut f1) = bid32_to_bid64(x);
    let (mut res, mut f2) = bid64_to_bid128(x64);
    return (res, (f1 | f2));
}

pub fn bid32_nexttoward(mut x: u32, mut y: BID_UINT128) -> (u32, u32) {
    bid32_next_toward(x, y)
}
