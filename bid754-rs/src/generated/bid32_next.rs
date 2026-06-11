// Auto-generated from bid32_next.go by go2rs. Do not edit.

use super::prelude::*;

pub fn bid32_next_up(mut x: u32) -> (u32, u32) {
    let mut res: u32 = 0;
    let mut x_sign: u32 = 0;
    let mut x_exp: u32 = 0;
    let mut x_nr_bits: i64 = 0;
    let mut q1: i64 = 0;
    let mut ind: i64 = 0;
    let mut C1: u32 = 0;
    let mut pfpsf: u32 = 0;
    if ((x & 0x7c000000) == 0x7c000000) {
        if ((x & 0x000fffff) > 999999) {
            x = (x & 0xfe000000);
        } else {
            x = (x & 0xfe0fffff);
        }
        if ((x & 0x7e000000) == 0x7e000000) {
            pfpsf |= 1;
            res = (x & 0xfdffffff);
        } else {
            res = x;
        }
        return (res, pfpsf);
    } else if ((x & 0x78000000) == 0x78000000) {
        if ((x & 0x80000000) == 0) {
            res = 0x78000000;
        } else {
            res = 0xf7f8967f;
        }
        return (res, pfpsf);
    }
    x_sign = (x & 0x80000000);
    if ((x & 0x60000000) == 0x60000000) {
        x_exp = (go_checked_shr_u32((x & 0x1fe00000), go_shift_count_u64((21) as u64)));
        C1 = ((x & 0x1fffff) | 0x800000);
        if (C1 > 9999999) {
            x_exp = 0;
            C1 = 0;
        }
    } else {
        x_exp = (go_checked_shr_u32((x & 0x7f800000), go_shift_count_u64((23) as u64)));
        C1 = (x & 0x7fffff);
    }
    if (C1 == 0) {
        res = 0x00000001;
    } else {
        if (x == 0x77f8967f) {
            res = 0x78000000;
        } else if (x == 0x80000001) {
            res = 0x80000000;
        } else {
            let mut tmp1 = ((C1 as f32) as f32).to_bits();
            x_nr_bits = (((1 as i64).wrapping_add(((((go_checked_shr_u32(tmp1, go_shift_count_u64((23) as u64)))) & 0xff) as i64))).wrapping_sub(0x7f));
            q1 = (bid_nr_digits[(x_nr_bits.wrapping_sub(1)) as usize].digits as i64);
            if (q1 == 0) {
                q1 = (bid_nr_digits[(x_nr_bits.wrapping_sub(1)) as usize].digits1 as i64);
                if ((C1 as u64) >= bid_nr_digits[(x_nr_bits.wrapping_sub(1)) as usize].threshold_lo) {
                    q1 = q1.wrapping_add(1);
                }
            }
            if (q1 < 7) {
                if (x_exp > (((7 as i64).wrapping_sub(q1)) as u32)) {
                    ind = ((7 as i64).wrapping_sub(q1));
                    C1 = (C1.wrapping_mul(bid_ten2k64[ind as usize] as u32));
                    x_exp = (x_exp.wrapping_sub(ind as u32));
                } else {
                    ind = (x_exp as i64);
                    C1 = (C1.wrapping_mul(bid_ten2k64[ind as usize] as u32));
                    x_exp = 0;
                }
            }
            if (x_sign == 0) {
                C1 = C1.wrapping_add(1);
                if (C1 == 0x989680) {
                    C1 = 0x0f4240;
                    x_exp = x_exp.wrapping_add(1);
                }
            } else {
                C1 = C1.wrapping_sub(1);
                if ((C1 == 0x0f423f) && (x_exp != 0)) {
                    C1 = 0x98967f;
                    x_exp = x_exp.wrapping_sub(1);
                }
            }
            if ((C1 & 0x800000) != 0) {
                res = (((x_sign | ((go_checked_shl_u32(x_exp, go_shift_count_u64((21) as u64))))) | 0x60000000) | (C1 & 0x1fffff));
            } else {
                res = ((x_sign | ((go_checked_shl_u32(x_exp, go_shift_count_u64((23) as u64))))) | C1);
            }
        }
    }
    return (res, pfpsf);
}

pub fn bid32_next_down(mut x: u32) -> (u32, u32) {
    let (mut res, mut flags) = bid32_next_up(x ^ 0x80000000);
    return ((res ^ 0x80000000), flags);
}

pub fn bid32_next_after(mut x: u32, mut y: u32) -> (u32, u32) {
    let mut res: u32 = 0;
    let mut pfpsf: u32 = 0;
    if ((x & 0x7c000000) == 0x7c000000) {
        if ((x & 0x000fffff) > 999999) {
            x = (x & 0xfe000000);
        } else {
            x = (x & 0xfe0fffff);
        }
        if ((x & 0x7e000000) == 0x7e000000) {
            pfpsf |= 1;
            res = (x & 0xfdffffff);
        } else {
            res = x;
        }
        if ((y & 0x7e000000) == 0x7e000000) {
            pfpsf |= 1;
        }
        return (res, pfpsf);
    }
    if ((y & 0x7c000000) == 0x7c000000) {
        if ((y & 0x000fffff) > 999999) {
            y = (y & 0xfe000000);
        } else {
            y = (y & 0xfe0fffff);
        }
        if ((y & 0x7e000000) == 0x7e000000) {
            pfpsf |= 1;
            res = (y & 0xfdffffff);
        } else {
            res = y;
        }
        return (res, pfpsf);
    }
    let (mut eqRes, _) = bid32_quiet_equal(x, y);
    if (eqRes != 0) {
        res = y;
        return (res, pfpsf);
    }
    let (mut lessRes, _) = bid32_quiet_greater(x, y);
    if (lessRes != 0) {
        (res, pfpsf) = bid32_next_down(x);
    } else {
        (res, pfpsf) = bid32_next_up(x);
    }
    if ((((x & 0x78000000) != 0x78000000)) && (((res & 0x78000000) == 0x78000000))) {
        pfpsf |= (32 | 8);
    }
    let mut tmp1: u32 = (0x00784000 as u32);
    let mut tmp2 = (res & 0x7fffffff);
    let (mut gtRes, _) = bid32_quiet_greater(tmp1, tmp2);
    let (mut neRes, _) = bid32_quiet_not_equal(x, res);
    if ((gtRes != 0) && (neRes != 0)) {
        pfpsf |= (32 | 16);
    }
    return (res, pfpsf);
}
