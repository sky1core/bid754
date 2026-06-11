// Auto-generated from to_uint32_int.go by go2rs. Do not edit.

use super::prelude::*;

pub fn bid64_to_uint32_int(mut x: u64) -> (u32, u32) {
    let mut res: u32 = 0;
    let mut x_sign: u64 = 0;
    let mut x_exp: u64 = 0;
    let mut exp: i64 = 0;
    let mut tmp64: u64 = 0;
    let mut tmp1: u64 = 0;
    let mut x_nr_bits: i64 = 0;
    let mut q: i64 = 0;
    let mut ind: i64 = 0;
    let mut shift: i64 = 0;
    let mut C1: u64 = 0;
    let mut Cstar: u64 = 0;
    let mut P128: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut pfpsf: u32 = 0;
    if (((x & 0x7c00000000000000) == 0x7c00000000000000) || ((x & 0x7800000000000000) == 0x7800000000000000)) {
        pfpsf |= 1;
        res = 0x80000000;
        return (res, pfpsf);
    }
    x_sign = (x & 0x8000000000000000);
    if ((x & 0x6000000000000000) == 0x6000000000000000) {
        x_exp = (go_checked_shr_u64((x & 0x1ff8000000000000), go_shift_count_u64((51) as u64)));
        C1 = ((x & 0x7ffffffffffff) | 0x20000000000000);
        if (C1 > 9999999999999999) {
            x_exp = 0;
            C1 = 0;
        }
    } else {
        x_exp = (go_checked_shr_u64((x & 0x7fe0000000000000), go_shift_count_u64((53) as u64)));
        C1 = (x & 0x1fffffffffffff);
    }
    if (C1 == 0x0) {
        res = 0x00000000;
        return (res, pfpsf);
    }
    if (C1 >= 0x0020000000000000) {
        tmp1 = (((go_checked_shr_u64(C1, go_shift_count_u64((32) as u64))) as f64)).to_bits();
        x_nr_bits = ((33 as i64).wrapping_add(((((((go_checked_shr_u64(tmp1, go_shift_count_u64((52) as u64)))) & 0x7ff)).wrapping_sub(0x3ff)) as i64)));
    } else {
        tmp1 = (C1 as f64).to_bits();
        x_nr_bits = ((1 as i64).wrapping_add(((((((go_checked_shr_u64(tmp1, go_shift_count_u64((52) as u64)))) & 0x7ff)).wrapping_sub(0x3ff)) as i64)));
    }
    q = (bid_nr_digits[(x_nr_bits.wrapping_sub(1)) as usize].digits as i64);
    if (q == 0) {
        q = (bid_nr_digits[(x_nr_bits.wrapping_sub(1)) as usize].digits1 as i64);
        if (C1 >= bid_nr_digits[(x_nr_bits.wrapping_sub(1)) as usize].threshold_lo) {
            q = q.wrapping_add(1);
        }
    }
    exp = ((x_exp as i64).wrapping_sub(398));
    if (((q.wrapping_add(exp))) > 10) {
        pfpsf |= 1;
        res = 0x80000000;
        return (res, pfpsf);
    } else if (((q.wrapping_add(exp))) == 10) {
        if (x_sign != 0) {
            pfpsf |= 1;
            res = 0x80000000;
            return (res, pfpsf);
        } else {
            if (q <= 11) {
                tmp64 = (C1.wrapping_mul(bid_ten2k64[((11 as i64).wrapping_sub(q)) as usize]));
                if (tmp64 >= 0xa00000000) {
                    pfpsf |= 1;
                    res = 0x80000000;
                    return (res, pfpsf);
                }
            } else {
                tmp64 = ((0xa00000000 as u64).wrapping_mul(bid_ten2k64[(q.wrapping_sub(11)) as usize]));
                if (C1 >= tmp64) {
                    pfpsf |= 1;
                    res = 0x80000000;
                    return (res, pfpsf);
                }
            }
        }
    }
    if (((q.wrapping_add(exp))) <= 0) {
        res = 0x00000000;
        return (res, pfpsf);
    } else {
        if (x_sign != 0) {
            pfpsf |= 1;
            res = 0x80000000;
            return (res, pfpsf);
        }
        if (exp < 0) {
            ind = (exp.wrapping_neg());
            P128 = __mul_64x64_to_128(C1, bid_ten2mk64[(ind.wrapping_sub(1)) as usize]);
            Cstar = P128.w[1];
            shift = (bid_shiftright128[(ind.wrapping_sub(1)) as usize] as i64);
            Cstar = (go_checked_shr_u64(Cstar, go_shift_count_i64((shift) as i64)));
            res = (Cstar as u32);
        } else if (exp == 0) {
            res = (C1 as u32);
        } else {
            res = ((C1.wrapping_mul(bid_ten2k64[exp as usize])) as u32);
        }
    }
    return (res, pfpsf);
}

pub fn bid64_to_uint32_xint(mut x: u64) -> (u32, u32) {
    let mut res: u32 = 0;
    let mut x_sign: u64 = 0;
    let mut x_exp: u64 = 0;
    let mut exp: i64 = 0;
    let mut tmp64: u64 = 0;
    let mut tmp1: u64 = 0;
    let mut x_nr_bits: i64 = 0;
    let mut q: i64 = 0;
    let mut ind: i64 = 0;
    let mut shift: i64 = 0;
    let mut C1: u64 = 0;
    let mut Cstar: u64 = 0;
    let mut fstar: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut P128: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut pfpsf: u32 = 0;
    if (((x & 0x7c00000000000000) == 0x7c00000000000000) || ((x & 0x7800000000000000) == 0x7800000000000000)) {
        pfpsf |= 1;
        res = 0x80000000;
        return (res, pfpsf);
    }
    x_sign = (x & 0x8000000000000000);
    if ((x & 0x6000000000000000) == 0x6000000000000000) {
        x_exp = (go_checked_shr_u64((x & 0x1ff8000000000000), go_shift_count_u64((51) as u64)));
        C1 = ((x & 0x7ffffffffffff) | 0x20000000000000);
        if (C1 > 9999999999999999) {
            x_exp = 0;
            C1 = 0;
        }
    } else {
        x_exp = (go_checked_shr_u64((x & 0x7fe0000000000000), go_shift_count_u64((53) as u64)));
        C1 = (x & 0x1fffffffffffff);
    }
    if (C1 == 0x0) {
        res = 0x00000000;
        return (res, pfpsf);
    }
    if (C1 >= 0x0020000000000000) {
        tmp1 = (((go_checked_shr_u64(C1, go_shift_count_u64((32) as u64))) as f64)).to_bits();
        x_nr_bits = ((33 as i64).wrapping_add(((((((go_checked_shr_u64(tmp1, go_shift_count_u64((52) as u64)))) & 0x7ff)).wrapping_sub(0x3ff)) as i64)));
    } else {
        tmp1 = (C1 as f64).to_bits();
        x_nr_bits = ((1 as i64).wrapping_add(((((((go_checked_shr_u64(tmp1, go_shift_count_u64((52) as u64)))) & 0x7ff)).wrapping_sub(0x3ff)) as i64)));
    }
    q = (bid_nr_digits[(x_nr_bits.wrapping_sub(1)) as usize].digits as i64);
    if (q == 0) {
        q = (bid_nr_digits[(x_nr_bits.wrapping_sub(1)) as usize].digits1 as i64);
        if (C1 >= bid_nr_digits[(x_nr_bits.wrapping_sub(1)) as usize].threshold_lo) {
            q = q.wrapping_add(1);
        }
    }
    exp = ((x_exp as i64).wrapping_sub(398));
    if (((q.wrapping_add(exp))) > 10) {
        pfpsf |= 1;
        res = 0x80000000;
        return (res, pfpsf);
    } else if (((q.wrapping_add(exp))) == 10) {
        if (x_sign != 0) {
            pfpsf |= 1;
            res = 0x80000000;
            return (res, pfpsf);
        } else {
            if (q <= 11) {
                tmp64 = (C1.wrapping_mul(bid_ten2k64[((11 as i64).wrapping_sub(q)) as usize]));
                if (tmp64 >= 0xa00000000) {
                    pfpsf |= 1;
                    res = 0x80000000;
                    return (res, pfpsf);
                }
            } else {
                tmp64 = ((0xa00000000 as u64).wrapping_mul(bid_ten2k64[(q.wrapping_sub(11)) as usize]));
                if (C1 >= tmp64) {
                    pfpsf |= 1;
                    res = 0x80000000;
                    return (res, pfpsf);
                }
            }
        }
    }
    if (((q.wrapping_add(exp))) <= 0) {
        pfpsf |= 32;
        res = 0x00000000;
        return (res, pfpsf);
    } else {
        if (x_sign != 0) {
            pfpsf |= 1;
            res = 0x80000000;
            return (res, pfpsf);
        }
        if (exp < 0) {
            ind = (exp.wrapping_neg());
            P128 = __mul_64x64_to_128(C1, bid_ten2mk64[(ind.wrapping_sub(1)) as usize]);
            Cstar = P128.w[1];
            fstar.w[1] = (P128.w[1] & bid_maskhigh128[(ind.wrapping_sub(1)) as usize]);
            fstar.w[0] = P128.w[0];
            shift = (bid_shiftright128[(ind.wrapping_sub(1)) as usize] as i64);
            Cstar = (go_checked_shr_u64(Cstar, go_shift_count_i64((shift) as i64)));
            if ((ind.wrapping_sub(1)) <= 2) {
                if (fstar.w[0] > bid_ten2mk128trunc[(ind.wrapping_sub(1)) as usize].w[1]) {
                    pfpsf |= 32;
                }
            } else {
                if ((fstar.w[1] != 0) || (fstar.w[0] > bid_ten2mk128trunc[(ind.wrapping_sub(1)) as usize].w[1])) {
                    pfpsf |= 32;
                }
            }
            res = (Cstar as u32);
        } else if (exp == 0) {
            res = (C1 as u32);
        } else {
            res = ((C1.wrapping_mul(bid_ten2k64[exp as usize])) as u32);
        }
    }
    return (res, pfpsf);
}
