// Auto-generated from round_integral64.go by go2rs. Do not edit.

use super::prelude::*;

pub fn bid64_round_integral_exact(mut x: u64, mut rndMode: i64) -> (u64, u32) {
    let mut res: u64 = 0xbaddbaddbaddbadd;
    let mut x_sign: u64 = 0;
    let mut x_nr_bits: i64 = 0;
    let mut q: i64 = 0;
    let mut ind: i64 = 0;
    let mut shift: i64 = 0;
    let mut C1: u64 = 0;
    let mut fstar: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut P128: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut pfpsf: u32 = 0;
    let mut exp: i64 = 0;
    let mut tmp1: u64 = 0;
    x_sign = (x & 0x8000000000000000);
    if ((x & 0x7c00000000000000) == 0x7c00000000000000) {
        if ((x & 0x0003ffffffffffff) > 999999999999999) {
            x = (x & 0xfe00000000000000);
        } else {
            x = (x & 0xfe03ffffffffffff);
        }
        if ((x & 0x7e00000000000000) == 0x7e00000000000000) {
            pfpsf |= 1;
            res = (x & 0xfdffffffffffffff);
        } else {
            res = x;
        }
        return (res, pfpsf);
    } else if ((x & 0x7800000000000000) == 0x7800000000000000) {
        res = (x_sign | 0x7800000000000000);
        return (res, pfpsf);
    }
    if ((x & 0x6000000000000000) == 0x6000000000000000) {
        exp = (((go_checked_shr_u64((x & 0x1ff8000000000000), go_shift_count_u64((51) as u64))) as i64).wrapping_sub(398));
        C1 = ((x & 0x7ffffffffffff) | 0x20000000000000);
        if (C1 > 9999999999999999) {
            C1 = 0;
        }
    } else {
        exp = (((go_checked_shr_u64((x & 0x7fe0000000000000), go_shift_count_u64((53) as u64))) as i64).wrapping_sub(398));
        C1 = (x & 0x1fffffffffffff);
    }
    if (C1 == 0) {
        if (exp < 0) {
            exp = 0;
        }
        res = (x_sign | ((go_checked_shl_u64((((exp as u64).wrapping_add(398))), go_shift_count_u64((53) as u64)))));
        return (res, pfpsf);
    }
    match rndMode {
        0 | 4 => {
            if (exp <= (-17)) {
                res = (x_sign | 0x31c0000000000000);
                pfpsf |= 32;
                return (res, pfpsf);
            }
        }
        1 => {
            if (exp <= (-16)) {
                if (x_sign != 0) {
                    res = 0xb1c0000000000001;
                } else {
                    res = 0x31c0000000000000;
                }
                pfpsf |= 32;
                return (res, pfpsf);
            }
        }
        2 => {
            if (exp <= (-16)) {
                if (x_sign != 0) {
                    res = 0xb1c0000000000000;
                } else {
                    res = 0x31c0000000000001;
                }
                pfpsf |= 32;
                return (res, pfpsf);
            }
        }
        3 => {
            if (exp <= (-16)) {
                res = (x_sign | 0x31c0000000000000);
                pfpsf |= 32;
                return (res, pfpsf);
            }
        }
        _ => {}
    }
    if (C1 >= 0x0020000000000000) {
        q = 16;
    } else {
        tmp1 = (C1 as f64).to_bits();
        x_nr_bits = ((1 as i64).wrapping_add(((((((go_checked_shr_u64(tmp1, go_shift_count_u64((52) as u64)))) & 0x7ff)).wrapping_sub(0x3ff)) as i64)));
        q = (bid_estimate_decimal_digits[(x_nr_bits.wrapping_sub(1)) as usize] as i64);
        if (C1 >= bid_power10_table_128[q as usize].w[0]) {
            q = q.wrapping_add(1);
        }
    }
    if (exp >= 0) {
        res = x;
        return (res, pfpsf);
    }
    match rndMode {
        0 => {
            if (((q.wrapping_add(exp))) >= 0) {
                ind = (exp.wrapping_neg());
                C1 = (C1.wrapping_add(bid_midpoint64[(ind.wrapping_sub(1)) as usize]));
                P128 = __mul_64x64_to_128(C1, bid_ten2mk64_round64[(ind.wrapping_sub(1)) as usize]);
                if (((ind.wrapping_sub(1))) <= 2) {
                    res = P128.w[1];
                    fstar.w[1] = 0;
                    fstar.w[0] = P128.w[0];
                } else if (((ind.wrapping_sub(1))) <= 21) {
                    shift = (bid_shiftright128_round64[(ind.wrapping_sub(1)) as usize] as i64);
                    res = ((go_checked_shr_u64(P128.w[1], go_shift_count_i64((shift) as i64))));
                    fstar.w[1] = (P128.w[1] & bid_maskhigh128_round64[(ind.wrapping_sub(1)) as usize]);
                    fstar.w[0] = P128.w[0];
                }
                if ((((res & 0x0000000000000001) != 0) && (fstar.w[1] == 0)) && ((fstar.w[0] < bid_ten2mk64_round64[(ind.wrapping_sub(1)) as usize]))) {
                    res = res.wrapping_sub(1);
                }
                if (((ind.wrapping_sub(1))) <= 2) {
                    if (fstar.w[0] > 0x8000000000000000) {
                        if (((fstar.w[0].wrapping_sub(0x8000000000000000))) > bid_ten2mk64_round64[(ind.wrapping_sub(1)) as usize]) {
                            pfpsf |= 32;
                        }
                    } else {
                        pfpsf |= 32;
                    }
                } else {
                    if ((fstar.w[1] > bid_onehalf128_round64[(ind.wrapping_sub(1)) as usize]) || (((fstar.w[1] == bid_onehalf128_round64[(ind.wrapping_sub(1)) as usize]) && (fstar.w[0] != 0)))) {
                        if ((fstar.w[1] > bid_onehalf128_round64[(ind.wrapping_sub(1)) as usize]) || (fstar.w[0] > bid_ten2mk64_round64[(ind.wrapping_sub(1)) as usize])) {
                            pfpsf |= 32;
                        }
                    } else {
                        pfpsf |= 32;
                    }
                }
                res = ((x_sign | 0x31c0000000000000) | res);
                return (res, pfpsf);
            } else {
                res = (x_sign | 0x31c0000000000000);
                pfpsf |= 32;
                return (res, pfpsf);
            }
        }
        4 => {
            if (((q.wrapping_add(exp))) >= 0) {
                ind = (exp.wrapping_neg());
                C1 = (C1.wrapping_add(bid_midpoint64[(ind.wrapping_sub(1)) as usize]));
                P128 = __mul_64x64_to_128(C1, bid_ten2mk64_round64[(ind.wrapping_sub(1)) as usize]);
                if (((ind.wrapping_sub(1))) <= 2) {
                    res = P128.w[1];
                    fstar.w[1] = 0;
                    fstar.w[0] = P128.w[0];
                } else if (((ind.wrapping_sub(1))) <= 21) {
                    shift = (bid_shiftright128_round64[(ind.wrapping_sub(1)) as usize] as i64);
                    res = ((go_checked_shr_u64(P128.w[1], go_shift_count_i64((shift) as i64))));
                    fstar.w[1] = (P128.w[1] & bid_maskhigh128_round64[(ind.wrapping_sub(1)) as usize]);
                    fstar.w[0] = P128.w[0];
                }
                if (((ind.wrapping_sub(1))) <= 2) {
                    if (fstar.w[0] > 0x8000000000000000) {
                        if (((fstar.w[0].wrapping_sub(0x8000000000000000))) > bid_ten2mk64_round64[(ind.wrapping_sub(1)) as usize]) {
                            pfpsf |= 32;
                        }
                    } else {
                        pfpsf |= 32;
                    }
                } else {
                    if ((fstar.w[1] > bid_onehalf128_round64[(ind.wrapping_sub(1)) as usize]) || (((fstar.w[1] == bid_onehalf128_round64[(ind.wrapping_sub(1)) as usize]) && (fstar.w[0] != 0)))) {
                        if ((fstar.w[1] > bid_onehalf128_round64[(ind.wrapping_sub(1)) as usize]) || (fstar.w[0] > bid_ten2mk64_round64[(ind.wrapping_sub(1)) as usize])) {
                            pfpsf |= 32;
                        }
                    } else {
                        pfpsf |= 32;
                    }
                }
                res = ((x_sign | 0x31c0000000000000) | res);
                return (res, pfpsf);
            } else {
                res = (x_sign | 0x31c0000000000000);
                pfpsf |= 32;
                return (res, pfpsf);
            }
        }
        1 => {
            if (((q.wrapping_add(exp))) > 0) {
                ind = (exp.wrapping_neg());
                P128 = __mul_64x64_to_128(C1, bid_ten2mk64_round64[(ind.wrapping_sub(1)) as usize]);
                if (((ind.wrapping_sub(1))) <= 2) {
                    res = P128.w[1];
                    fstar.w[1] = 0;
                    fstar.w[0] = P128.w[0];
                } else if (((ind.wrapping_sub(1))) <= 21) {
                    shift = (bid_shiftright128_round64[(ind.wrapping_sub(1)) as usize] as i64);
                    res = ((go_checked_shr_u64(P128.w[1], go_shift_count_i64((shift) as i64))));
                    fstar.w[1] = (P128.w[1] & bid_maskhigh128_round64[(ind.wrapping_sub(1)) as usize]);
                    fstar.w[0] = P128.w[0];
                }
                if ((fstar.w[1] != 0) || ((fstar.w[0] >= bid_ten2mk64_round64[(ind.wrapping_sub(1)) as usize]))) {
                    if (x_sign != 0) {
                        res = res.wrapping_add(1);
                    }
                    pfpsf |= 32;
                }
                res = ((x_sign | 0x31c0000000000000) | res);
                return (res, pfpsf);
            } else {
                if (x_sign != 0) {
                    res = 0xb1c0000000000001;
                } else {
                    res = 0x31c0000000000000;
                }
                pfpsf |= 32;
                return (res, pfpsf);
            }
        }
        2 => {
            if (((q.wrapping_add(exp))) > 0) {
                ind = (exp.wrapping_neg());
                P128 = __mul_64x64_to_128(C1, bid_ten2mk64_round64[(ind.wrapping_sub(1)) as usize]);
                if (((ind.wrapping_sub(1))) <= 2) {
                    res = P128.w[1];
                    fstar.w[1] = 0;
                    fstar.w[0] = P128.w[0];
                } else if (((ind.wrapping_sub(1))) <= 21) {
                    shift = (bid_shiftright128_round64[(ind.wrapping_sub(1)) as usize] as i64);
                    res = ((go_checked_shr_u64(P128.w[1], go_shift_count_i64((shift) as i64))));
                    fstar.w[1] = (P128.w[1] & bid_maskhigh128_round64[(ind.wrapping_sub(1)) as usize]);
                    fstar.w[0] = P128.w[0];
                }
                if ((fstar.w[1] != 0) || ((fstar.w[0] >= bid_ten2mk64_round64[(ind.wrapping_sub(1)) as usize]))) {
                    if (x_sign == 0) {
                        res = res.wrapping_add(1);
                    }
                    pfpsf |= 32;
                }
                res = ((x_sign | 0x31c0000000000000) | res);
                return (res, pfpsf);
            } else {
                if (x_sign != 0) {
                    res = 0xb1c0000000000000;
                } else {
                    res = 0x31c0000000000001;
                }
                pfpsf |= 32;
                return (res, pfpsf);
            }
        }
        3 => {
            if (((q.wrapping_add(exp))) >= 0) {
                ind = (exp.wrapping_neg());
                P128 = __mul_64x64_to_128(C1, bid_ten2mk64_round64[(ind.wrapping_sub(1)) as usize]);
                if (((ind.wrapping_sub(1))) <= 2) {
                    res = P128.w[1];
                    fstar.w[1] = 0;
                    fstar.w[0] = P128.w[0];
                } else if (((ind.wrapping_sub(1))) <= 21) {
                    shift = (bid_shiftright128_round64[(ind.wrapping_sub(1)) as usize] as i64);
                    res = ((go_checked_shr_u64(P128.w[1], go_shift_count_i64((shift) as i64))));
                    fstar.w[1] = (P128.w[1] & bid_maskhigh128_round64[(ind.wrapping_sub(1)) as usize]);
                    fstar.w[0] = P128.w[0];
                }
                if ((fstar.w[1] != 0) || ((fstar.w[0] >= bid_ten2mk64_round64[(ind.wrapping_sub(1)) as usize]))) {
                    pfpsf |= 32;
                }
                res = ((x_sign | 0x31c0000000000000) | res);
                return (res, pfpsf);
            } else {
                res = (x_sign | 0x31c0000000000000);
                pfpsf |= 32;
                return (res, pfpsf);
            }
        }
        _ => {}
    }
    return (res, pfpsf);
}

pub fn bid64_nearby_int(mut x: u64, mut rndMode: i64) -> (u64, u32) {
    let mut res: u64 = 0xbaddbaddbaddbadd;
    let mut x_sign: u64 = 0;
    let mut x_nr_bits: i64 = 0;
    let mut q: i64 = 0;
    let mut ind: i64 = 0;
    let mut shift: i64 = 0;
    let mut C1: u64 = 0;
    let mut fstar: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut P128: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut pfpsf: u32 = 0;
    let mut exp: i64 = 0;
    let mut tmp1: u64 = 0;
    x_sign = (x & 0x8000000000000000);
    if ((x & 0x7c00000000000000) == 0x7c00000000000000) {
        if ((x & 0x0003ffffffffffff) > 999999999999999) {
            x = (x & 0xfe00000000000000);
        } else {
            x = (x & 0xfe03ffffffffffff);
        }
        if ((x & 0x7e00000000000000) == 0x7e00000000000000) {
            pfpsf |= 1;
            res = (x & 0xfdffffffffffffff);
        } else {
            res = x;
        }
        return (res, pfpsf);
    } else if ((x & 0x7800000000000000) == 0x7800000000000000) {
        res = (x_sign | 0x7800000000000000);
        return (res, pfpsf);
    }
    if ((x & 0x6000000000000000) == 0x6000000000000000) {
        exp = (((go_checked_shr_u64((x & 0x1ff8000000000000), go_shift_count_u64((51) as u64))) as i64).wrapping_sub(398));
        C1 = ((x & 0x7ffffffffffff) | 0x20000000000000);
        if (C1 > 9999999999999999) {
            C1 = 0;
        }
    } else {
        exp = (((go_checked_shr_u64((x & 0x7fe0000000000000), go_shift_count_u64((53) as u64))) as i64).wrapping_sub(398));
        C1 = (x & 0x1fffffffffffff);
    }
    if (C1 == 0) {
        if (exp < 0) {
            exp = 0;
        }
        res = (x_sign | ((go_checked_shl_u64((((exp as u64).wrapping_add(398))), go_shift_count_u64((53) as u64)))));
        return (res, pfpsf);
    }
    match rndMode {
        0 | 4 => {
            if (exp <= (-17)) {
                res = (x_sign | 0x31c0000000000000);
                return (res, pfpsf);
            }
        }
        1 => {
            if (exp <= (-16)) {
                if (x_sign != 0) {
                    res = 0xb1c0000000000001;
                } else {
                    res = 0x31c0000000000000;
                }
                return (res, pfpsf);
            }
        }
        2 => {
            if (exp <= (-16)) {
                if (x_sign != 0) {
                    res = 0xb1c0000000000000;
                } else {
                    res = 0x31c0000000000001;
                }
                return (res, pfpsf);
            }
        }
        3 => {
            if (exp <= (-16)) {
                res = (x_sign | 0x31c0000000000000);
                return (res, pfpsf);
            }
        }
        _ => {}
    }
    if (C1 >= 0x0020000000000000) {
        q = 16;
    } else {
        tmp1 = (C1 as f64).to_bits();
        x_nr_bits = ((1 as i64).wrapping_add(((((((go_checked_shr_u64(tmp1, go_shift_count_u64((52) as u64)))) & 0x7ff)).wrapping_sub(0x3ff)) as i64)));
        q = (bid_estimate_decimal_digits[(x_nr_bits.wrapping_sub(1)) as usize] as i64);
        if (C1 >= bid_power10_table_128[q as usize].w[0]) {
            q = q.wrapping_add(1);
        }
    }
    if (exp >= 0) {
        res = x;
        return (res, pfpsf);
    }
    match rndMode {
        0 => {
            if (((q.wrapping_add(exp))) >= 0) {
                ind = (exp.wrapping_neg());
                C1 = (C1.wrapping_add(bid_midpoint64[(ind.wrapping_sub(1)) as usize]));
                P128 = __mul_64x64_to_128(C1, bid_ten2mk64_round64[(ind.wrapping_sub(1)) as usize]);
                if (((ind.wrapping_sub(1))) <= 2) {
                    res = P128.w[1];
                    fstar.w[1] = 0;
                    fstar.w[0] = P128.w[0];
                } else if (((ind.wrapping_sub(1))) <= 21) {
                    shift = (bid_shiftright128_round64[(ind.wrapping_sub(1)) as usize] as i64);
                    res = ((go_checked_shr_u64(P128.w[1], go_shift_count_i64((shift) as i64))));
                    fstar.w[1] = (P128.w[1] & bid_maskhigh128_round64[(ind.wrapping_sub(1)) as usize]);
                    fstar.w[0] = P128.w[0];
                }
                if ((((res & 0x0000000000000001) != 0) && (fstar.w[1] == 0)) && ((fstar.w[0] < bid_ten2mk64_round64[(ind.wrapping_sub(1)) as usize]))) {
                    res = res.wrapping_sub(1);
                }
                res = ((x_sign | 0x31c0000000000000) | res);
                return (res, pfpsf);
            } else {
                res = (x_sign | 0x31c0000000000000);
                return (res, pfpsf);
            }
        }
        4 => {
            if (((q.wrapping_add(exp))) >= 0) {
                ind = (exp.wrapping_neg());
                C1 = (C1.wrapping_add(bid_midpoint64[(ind.wrapping_sub(1)) as usize]));
                P128 = __mul_64x64_to_128(C1, bid_ten2mk64_round64[(ind.wrapping_sub(1)) as usize]);
                if (((ind.wrapping_sub(1))) <= 2) {
                    res = P128.w[1];
                } else if (((ind.wrapping_sub(1))) <= 21) {
                    shift = (bid_shiftright128_round64[(ind.wrapping_sub(1)) as usize] as i64);
                    res = ((go_checked_shr_u64(P128.w[1], go_shift_count_i64((shift) as i64))));
                }
                res = ((x_sign | 0x31c0000000000000) | res);
                return (res, pfpsf);
            } else {
                res = (x_sign | 0x31c0000000000000);
                return (res, pfpsf);
            }
        }
        1 => {
            if (((q.wrapping_add(exp))) > 0) {
                ind = (exp.wrapping_neg());
                P128 = __mul_64x64_to_128(C1, bid_ten2mk64_round64[(ind.wrapping_sub(1)) as usize]);
                if (((ind.wrapping_sub(1))) <= 2) {
                    res = P128.w[1];
                    fstar.w[1] = 0;
                    fstar.w[0] = P128.w[0];
                } else if (((ind.wrapping_sub(1))) <= 21) {
                    shift = (bid_shiftright128_round64[(ind.wrapping_sub(1)) as usize] as i64);
                    res = ((go_checked_shr_u64(P128.w[1], go_shift_count_i64((shift) as i64))));
                    fstar.w[1] = (P128.w[1] & bid_maskhigh128_round64[(ind.wrapping_sub(1)) as usize]);
                    fstar.w[0] = P128.w[0];
                }
                if ((fstar.w[1] != 0) || ((fstar.w[0] >= bid_ten2mk64_round64[(ind.wrapping_sub(1)) as usize]))) {
                    if (x_sign != 0) {
                        res = res.wrapping_add(1);
                    }
                }
                res = ((x_sign | 0x31c0000000000000) | res);
                return (res, pfpsf);
            } else {
                if (x_sign != 0) {
                    res = 0xb1c0000000000001;
                } else {
                    res = 0x31c0000000000000;
                }
                return (res, pfpsf);
            }
        }
        2 => {
            if (((q.wrapping_add(exp))) > 0) {
                ind = (exp.wrapping_neg());
                P128 = __mul_64x64_to_128(C1, bid_ten2mk64_round64[(ind.wrapping_sub(1)) as usize]);
                if (((ind.wrapping_sub(1))) <= 2) {
                    res = P128.w[1];
                    fstar.w[1] = 0;
                    fstar.w[0] = P128.w[0];
                } else if (((ind.wrapping_sub(1))) <= 21) {
                    shift = (bid_shiftright128_round64[(ind.wrapping_sub(1)) as usize] as i64);
                    res = ((go_checked_shr_u64(P128.w[1], go_shift_count_i64((shift) as i64))));
                    fstar.w[1] = (P128.w[1] & bid_maskhigh128_round64[(ind.wrapping_sub(1)) as usize]);
                    fstar.w[0] = P128.w[0];
                }
                if ((fstar.w[1] != 0) || ((fstar.w[0] >= bid_ten2mk64_round64[(ind.wrapping_sub(1)) as usize]))) {
                    if (x_sign == 0) {
                        res = res.wrapping_add(1);
                    }
                }
                res = ((x_sign | 0x31c0000000000000) | res);
                return (res, pfpsf);
            } else {
                if (x_sign != 0) {
                    res = 0xb1c0000000000000;
                } else {
                    res = 0x31c0000000000001;
                }
                return (res, pfpsf);
            }
        }
        3 => {
            if (((q.wrapping_add(exp))) >= 0) {
                ind = (exp.wrapping_neg());
                P128 = __mul_64x64_to_128(C1, bid_ten2mk64_round64[(ind.wrapping_sub(1)) as usize]);
                if (((ind.wrapping_sub(1))) <= 2) {
                    res = P128.w[1];
                } else if (((ind.wrapping_sub(1))) <= 21) {
                    shift = (bid_shiftright128_round64[(ind.wrapping_sub(1)) as usize] as i64);
                    res = ((go_checked_shr_u64(P128.w[1], go_shift_count_i64((shift) as i64))));
                }
                res = ((x_sign | 0x31c0000000000000) | res);
                return (res, pfpsf);
            } else {
                res = (x_sign | 0x31c0000000000000);
                return (res, pfpsf);
            }
        }
        _ => {}
    }
    return (res, pfpsf);
}

pub fn bid64_round_integral_nearest_even(mut x: u64) -> (u64, u32) {
    let mut res: u64 = 0xbaddbaddbaddbadd;
    let mut x_sign: u64 = 0;
    let mut x_nr_bits: i64 = 0;
    let mut q: i64 = 0;
    let mut ind: i64 = 0;
    let mut shift: i64 = 0;
    let mut C1: u64 = 0;
    let mut fstar: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut P128: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut pfpsf: u32 = 0;
    let mut exp: i64 = 0;
    let mut tmp1: u64 = 0;
    x_sign = (x & 0x8000000000000000);
    if ((x & 0x7c00000000000000) == 0x7c00000000000000) {
        if ((x & 0x0003ffffffffffff) > 999999999999999) {
            x = (x & 0xfe00000000000000);
        } else {
            x = (x & 0xfe03ffffffffffff);
        }
        if ((x & 0x7e00000000000000) == 0x7e00000000000000) {
            pfpsf |= 1;
            res = (x & 0xfdffffffffffffff);
        } else {
            res = x;
        }
        return (res, pfpsf);
    } else if ((x & 0x7800000000000000) == 0x7800000000000000) {
        res = (x_sign | 0x7800000000000000);
        return (res, pfpsf);
    }
    if ((x & 0x6000000000000000) == 0x6000000000000000) {
        exp = (((go_checked_shr_u64((x & 0x1ff8000000000000), go_shift_count_u64((51) as u64))) as i64).wrapping_sub(398));
        C1 = ((x & 0x7ffffffffffff) | 0x20000000000000);
        if (C1 > 9999999999999999) {
            C1 = 0;
        }
    } else {
        exp = (((go_checked_shr_u64((x & 0x7fe0000000000000), go_shift_count_u64((53) as u64))) as i64).wrapping_sub(398));
        C1 = (x & 0x1fffffffffffff);
    }
    if (C1 == 0) {
        if (exp < 0) {
            exp = 0;
        }
        res = (x_sign | ((go_checked_shl_u64((((exp as u64).wrapping_add(398))), go_shift_count_u64((53) as u64)))));
        return (res, pfpsf);
    }
    if (exp <= (-17)) {
        res = (x_sign | 0x31c0000000000000);
        return (res, pfpsf);
    }
    if (C1 >= 0x0020000000000000) {
        q = 16;
    } else {
        tmp1 = (C1 as f64).to_bits();
        x_nr_bits = ((1 as i64).wrapping_add(((((((go_checked_shr_u64(tmp1, go_shift_count_u64((52) as u64)))) & 0x7ff)).wrapping_sub(0x3ff)) as i64)));
        q = (bid_estimate_decimal_digits[(x_nr_bits.wrapping_sub(1)) as usize] as i64);
        if (C1 >= bid_power10_table_128[q as usize].w[0]) {
            q = q.wrapping_add(1);
        }
    }
    if (exp >= 0) {
        res = x;
        return (res, pfpsf);
    } else if (((q.wrapping_add(exp))) >= 0) {
        ind = (exp.wrapping_neg());
        C1 = (C1.wrapping_add(bid_midpoint64[(ind.wrapping_sub(1)) as usize]));
        P128 = __mul_64x64_to_128(C1, bid_ten2mk64_round64[(ind.wrapping_sub(1)) as usize]);
        if (((ind.wrapping_sub(1))) <= 2) {
            res = P128.w[1];
            fstar.w[1] = 0;
            fstar.w[0] = P128.w[0];
        } else if (((ind.wrapping_sub(1))) <= 21) {
            shift = (bid_shiftright128_round64[(ind.wrapping_sub(1)) as usize] as i64);
            res = ((go_checked_shr_u64(P128.w[1], go_shift_count_i64((shift) as i64))));
            fstar.w[1] = (P128.w[1] & bid_maskhigh128_round64[(ind.wrapping_sub(1)) as usize]);
            fstar.w[0] = P128.w[0];
        }
        if ((((res & 0x0000000000000001) != 0) && (fstar.w[1] == 0)) && ((fstar.w[0] < bid_ten2mk64_round64[(ind.wrapping_sub(1)) as usize]))) {
            res = res.wrapping_sub(1);
        }
        res = ((x_sign | 0x31c0000000000000) | res);
        return (res, pfpsf);
    } else {
        res = (x_sign | 0x31c0000000000000);
        return (res, pfpsf);
    }
}

pub fn bid64_round_integral_negative(mut x: u64) -> (u64, u32) {
    let mut res: u64 = 0xbaddbaddbaddbadd;
    let mut x_sign: u64 = 0;
    let mut x_nr_bits: i64 = 0;
    let mut q: i64 = 0;
    let mut ind: i64 = 0;
    let mut shift: i64 = 0;
    let mut C1: u64 = 0;
    let mut fstar: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut P128: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut pfpsf: u32 = 0;
    let mut exp: i64 = 0;
    let mut tmp1: u64 = 0;
    x_sign = (x & 0x8000000000000000);
    if ((x & 0x7c00000000000000) == 0x7c00000000000000) {
        if ((x & 0x0003ffffffffffff) > 999999999999999) {
            x = (x & 0xfe00000000000000);
        } else {
            x = (x & 0xfe03ffffffffffff);
        }
        if ((x & 0x7e00000000000000) == 0x7e00000000000000) {
            pfpsf |= 1;
            res = (x & 0xfdffffffffffffff);
        } else {
            res = x;
        }
        return (res, pfpsf);
    } else if ((x & 0x7800000000000000) == 0x7800000000000000) {
        res = (x_sign | 0x7800000000000000);
        return (res, pfpsf);
    }
    if ((x & 0x6000000000000000) == 0x6000000000000000) {
        exp = (((go_checked_shr_u64((x & 0x1ff8000000000000), go_shift_count_u64((51) as u64))) as i64).wrapping_sub(398));
        C1 = ((x & 0x7ffffffffffff) | 0x20000000000000);
        if (C1 > 9999999999999999) {
            C1 = 0;
        }
    } else {
        exp = (((go_checked_shr_u64((x & 0x7fe0000000000000), go_shift_count_u64((53) as u64))) as i64).wrapping_sub(398));
        C1 = (x & 0x1fffffffffffff);
    }
    if (C1 == 0) {
        if (exp < 0) {
            exp = 0;
        }
        res = (x_sign | ((go_checked_shl_u64((((exp as u64).wrapping_add(398))), go_shift_count_u64((53) as u64)))));
        return (res, pfpsf);
    }
    if (exp <= (-16)) {
        if (x_sign != 0) {
            res = 0xb1c0000000000001;
        } else {
            res = 0x31c0000000000000;
        }
        return (res, pfpsf);
    }
    if (C1 >= 0x0020000000000000) {
        q = 16;
    } else {
        tmp1 = (C1 as f64).to_bits();
        x_nr_bits = ((1 as i64).wrapping_add(((((((go_checked_shr_u64(tmp1, go_shift_count_u64((52) as u64)))) & 0x7ff)).wrapping_sub(0x3ff)) as i64)));
        q = (bid_estimate_decimal_digits[(x_nr_bits.wrapping_sub(1)) as usize] as i64);
        if (C1 >= bid_power10_table_128[q as usize].w[0]) {
            q = q.wrapping_add(1);
        }
    }
    if (exp >= 0) {
        res = x;
        return (res, pfpsf);
    } else if (((q.wrapping_add(exp))) > 0) {
        ind = (exp.wrapping_neg());
        P128 = __mul_64x64_to_128(C1, bid_ten2mk64_round64[(ind.wrapping_sub(1)) as usize]);
        if (((ind.wrapping_sub(1))) <= 2) {
            res = P128.w[1];
            fstar.w[1] = 0;
            fstar.w[0] = P128.w[0];
        } else if (((ind.wrapping_sub(1))) <= 21) {
            shift = (bid_shiftright128_round64[(ind.wrapping_sub(1)) as usize] as i64);
            res = ((go_checked_shr_u64(P128.w[1], go_shift_count_i64((shift) as i64))));
            fstar.w[1] = (P128.w[1] & bid_maskhigh128_round64[(ind.wrapping_sub(1)) as usize]);
            fstar.w[0] = P128.w[0];
        }
        if ((x_sign != 0) && (((fstar.w[1] != 0) || ((fstar.w[0] >= bid_ten2mk64_round64[(ind.wrapping_sub(1)) as usize]))))) {
            res = res.wrapping_add(1);
        }
        res = ((x_sign | 0x31c0000000000000) | res);
        return (res, pfpsf);
    } else {
        if (x_sign != 0) {
            res = 0xb1c0000000000001;
        } else {
            res = 0x31c0000000000000;
        }
        return (res, pfpsf);
    }
}

pub fn bid64_round_integral_positive(mut x: u64) -> (u64, u32) {
    let mut res: u64 = 0xbaddbaddbaddbadd;
    let mut x_sign: u64 = 0;
    let mut x_nr_bits: i64 = 0;
    let mut q: i64 = 0;
    let mut ind: i64 = 0;
    let mut shift: i64 = 0;
    let mut C1: u64 = 0;
    let mut fstar: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut P128: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut pfpsf: u32 = 0;
    let mut exp: i64 = 0;
    let mut tmp1: u64 = 0;
    x_sign = (x & 0x8000000000000000);
    if ((x & 0x7c00000000000000) == 0x7c00000000000000) {
        if ((x & 0x0003ffffffffffff) > 999999999999999) {
            x = (x & 0xfe00000000000000);
        } else {
            x = (x & 0xfe03ffffffffffff);
        }
        if ((x & 0x7e00000000000000) == 0x7e00000000000000) {
            pfpsf |= 1;
            res = (x & 0xfdffffffffffffff);
        } else {
            res = x;
        }
        return (res, pfpsf);
    } else if ((x & 0x7800000000000000) == 0x7800000000000000) {
        res = (x_sign | 0x7800000000000000);
        return (res, pfpsf);
    }
    if ((x & 0x6000000000000000) == 0x6000000000000000) {
        exp = (((go_checked_shr_u64((x & 0x1ff8000000000000), go_shift_count_u64((51) as u64))) as i64).wrapping_sub(398));
        C1 = ((x & 0x7ffffffffffff) | 0x20000000000000);
        if (C1 > 9999999999999999) {
            C1 = 0;
        }
    } else {
        exp = (((go_checked_shr_u64((x & 0x7fe0000000000000), go_shift_count_u64((53) as u64))) as i64).wrapping_sub(398));
        C1 = (x & 0x1fffffffffffff);
    }
    if (C1 == 0) {
        if (exp < 0) {
            exp = 0;
        }
        res = (x_sign | ((go_checked_shl_u64((((exp as u64).wrapping_add(398))), go_shift_count_u64((53) as u64)))));
        return (res, pfpsf);
    }
    if (exp <= (-16)) {
        if (x_sign != 0) {
            res = 0xb1c0000000000000;
        } else {
            res = 0x31c0000000000001;
        }
        return (res, pfpsf);
    }
    if (C1 >= 0x0020000000000000) {
        q = 16;
    } else {
        tmp1 = (C1 as f64).to_bits();
        x_nr_bits = ((1 as i64).wrapping_add(((((((go_checked_shr_u64(tmp1, go_shift_count_u64((52) as u64)))) & 0x7ff)).wrapping_sub(0x3ff)) as i64)));
        q = (bid_estimate_decimal_digits[(x_nr_bits.wrapping_sub(1)) as usize] as i64);
        if (C1 >= bid_power10_table_128[q as usize].w[0]) {
            q = q.wrapping_add(1);
        }
    }
    if (exp >= 0) {
        res = x;
        return (res, pfpsf);
    } else if (((q.wrapping_add(exp))) > 0) {
        ind = (exp.wrapping_neg());
        P128 = __mul_64x64_to_128(C1, bid_ten2mk64_round64[(ind.wrapping_sub(1)) as usize]);
        if (((ind.wrapping_sub(1))) <= 2) {
            res = P128.w[1];
            fstar.w[1] = 0;
            fstar.w[0] = P128.w[0];
        } else if (((ind.wrapping_sub(1))) <= 21) {
            shift = (bid_shiftright128_round64[(ind.wrapping_sub(1)) as usize] as i64);
            res = ((go_checked_shr_u64(P128.w[1], go_shift_count_i64((shift) as i64))));
            fstar.w[1] = (P128.w[1] & bid_maskhigh128_round64[(ind.wrapping_sub(1)) as usize]);
            fstar.w[0] = P128.w[0];
        }
        if ((x_sign == 0) && (((fstar.w[1] != 0) || ((fstar.w[0] >= bid_ten2mk64_round64[(ind.wrapping_sub(1)) as usize]))))) {
            res = res.wrapping_add(1);
        }
        res = ((x_sign | 0x31c0000000000000) | res);
        return (res, pfpsf);
    } else {
        if (x_sign != 0) {
            res = 0xb1c0000000000000;
        } else {
            res = 0x31c0000000000001;
        }
        return (res, pfpsf);
    }
}

pub fn bid64_round_integral_zero(mut x: u64) -> (u64, u32) {
    let mut res: u64 = 0xbaddbaddbaddbadd;
    let mut x_sign: u64 = 0;
    let mut x_nr_bits: i64 = 0;
    let mut q: i64 = 0;
    let mut ind: i64 = 0;
    let mut shift: i64 = 0;
    let mut C1: u64 = 0;
    let mut P128: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut pfpsf: u32 = 0;
    let mut exp: i64 = 0;
    let mut tmp1: u64 = 0;
    x_sign = (x & 0x8000000000000000);
    if ((x & 0x7c00000000000000) == 0x7c00000000000000) {
        if ((x & 0x0003ffffffffffff) > 999999999999999) {
            x = (x & 0xfe00000000000000);
        } else {
            x = (x & 0xfe03ffffffffffff);
        }
        if ((x & 0x7e00000000000000) == 0x7e00000000000000) {
            pfpsf |= 1;
            res = (x & 0xfdffffffffffffff);
        } else {
            res = x;
        }
        return (res, pfpsf);
    } else if ((x & 0x7800000000000000) == 0x7800000000000000) {
        res = (x_sign | 0x7800000000000000);
        return (res, pfpsf);
    }
    if ((x & 0x6000000000000000) == 0x6000000000000000) {
        exp = (((go_checked_shr_u64((x & 0x1ff8000000000000), go_shift_count_u64((51) as u64))) as i64).wrapping_sub(398));
        C1 = ((x & 0x7ffffffffffff) | 0x20000000000000);
        if (C1 > 9999999999999999) {
            C1 = 0;
        }
    } else {
        exp = (((go_checked_shr_u64((x & 0x7fe0000000000000), go_shift_count_u64((53) as u64))) as i64).wrapping_sub(398));
        C1 = (x & 0x1fffffffffffff);
    }
    if (C1 == 0) {
        if (exp < 0) {
            exp = 0;
        }
        res = (x_sign | ((go_checked_shl_u64((((exp as u64).wrapping_add(398))), go_shift_count_u64((53) as u64)))));
        return (res, pfpsf);
    }
    if (exp <= (-16)) {
        res = (x_sign | 0x31c0000000000000);
        return (res, pfpsf);
    }
    if (C1 >= 0x0020000000000000) {
        q = 16;
    } else {
        tmp1 = (C1 as f64).to_bits();
        x_nr_bits = ((1 as i64).wrapping_add(((((((go_checked_shr_u64(tmp1, go_shift_count_u64((52) as u64)))) & 0x7ff)).wrapping_sub(0x3ff)) as i64)));
        q = (bid_estimate_decimal_digits[(x_nr_bits.wrapping_sub(1)) as usize] as i64);
        if (C1 >= bid_power10_table_128[q as usize].w[0]) {
            q = q.wrapping_add(1);
        }
    }
    if (exp >= 0) {
        res = x;
        return (res, pfpsf);
    } else if (((q.wrapping_add(exp))) >= 0) {
        ind = (exp.wrapping_neg());
        P128 = __mul_64x64_to_128(C1, bid_ten2mk64_round64[(ind.wrapping_sub(1)) as usize]);
        if (((ind.wrapping_sub(1))) <= 2) {
            res = P128.w[1];
        } else if (((ind.wrapping_sub(1))) <= 21) {
            shift = (bid_shiftright128_round64[(ind.wrapping_sub(1)) as usize] as i64);
            res = ((go_checked_shr_u64(P128.w[1], go_shift_count_i64((shift) as i64))));
        }
        res = ((x_sign | 0x31c0000000000000) | res);
        return (res, pfpsf);
    } else {
        res = (x_sign | 0x31c0000000000000);
        return (res, pfpsf);
    }
}

pub fn bid64_round_integral_nearest_away(mut x: u64) -> (u64, u32) {
    let mut res: u64 = 0xbaddbaddbaddbadd;
    let mut x_sign: u64 = 0;
    let mut x_nr_bits: i64 = 0;
    let mut q: i64 = 0;
    let mut ind: i64 = 0;
    let mut shift: i64 = 0;
    let mut C1: u64 = 0;
    let mut P128: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut pfpsf: u32 = 0;
    let mut exp: i64 = 0;
    let mut tmp1: u64 = 0;
    x_sign = (x & 0x8000000000000000);
    if ((x & 0x7c00000000000000) == 0x7c00000000000000) {
        if ((x & 0x0003ffffffffffff) > 999999999999999) {
            x = (x & 0xfe00000000000000);
        } else {
            x = (x & 0xfe03ffffffffffff);
        }
        if ((x & 0x7e00000000000000) == 0x7e00000000000000) {
            pfpsf |= 1;
            res = (x & 0xfdffffffffffffff);
        } else {
            res = x;
        }
        return (res, pfpsf);
    } else if ((x & 0x7800000000000000) == 0x7800000000000000) {
        res = (x_sign | 0x7800000000000000);
        return (res, pfpsf);
    }
    if ((x & 0x6000000000000000) == 0x6000000000000000) {
        exp = (((go_checked_shr_u64((x & 0x1ff8000000000000), go_shift_count_u64((51) as u64))) as i64).wrapping_sub(398));
        C1 = ((x & 0x7ffffffffffff) | 0x20000000000000);
        if (C1 > 9999999999999999) {
            C1 = 0;
        }
    } else {
        exp = (((go_checked_shr_u64((x & 0x7fe0000000000000), go_shift_count_u64((53) as u64))) as i64).wrapping_sub(398));
        C1 = (x & 0x1fffffffffffff);
    }
    if (C1 == 0) {
        if (exp < 0) {
            exp = 0;
        }
        res = (x_sign | ((go_checked_shl_u64((((exp as u64).wrapping_add(398))), go_shift_count_u64((53) as u64)))));
        return (res, pfpsf);
    }
    if (exp <= (-17)) {
        res = (x_sign | 0x31c0000000000000);
        return (res, pfpsf);
    }
    if (C1 >= 0x0020000000000000) {
        q = 16;
    } else {
        tmp1 = (C1 as f64).to_bits();
        x_nr_bits = ((1 as i64).wrapping_add(((((((go_checked_shr_u64(tmp1, go_shift_count_u64((52) as u64)))) & 0x7ff)).wrapping_sub(0x3ff)) as i64)));
        q = (bid_estimate_decimal_digits[(x_nr_bits.wrapping_sub(1)) as usize] as i64);
        if (C1 >= bid_power10_table_128[q as usize].w[0]) {
            q = q.wrapping_add(1);
        }
    }
    if (exp >= 0) {
        res = x;
        return (res, pfpsf);
    } else if (((q.wrapping_add(exp))) >= 0) {
        ind = (exp.wrapping_neg());
        C1 = (C1.wrapping_add(bid_midpoint64[(ind.wrapping_sub(1)) as usize]));
        P128 = __mul_64x64_to_128(C1, bid_ten2mk64_round64[(ind.wrapping_sub(1)) as usize]);
        if (((ind.wrapping_sub(1))) <= 2) {
            res = P128.w[1];
        } else if (((ind.wrapping_sub(1))) <= 21) {
            shift = (bid_shiftright128_round64[(ind.wrapping_sub(1)) as usize] as i64);
            res = ((go_checked_shr_u64(P128.w[1], go_shift_count_i64((shift) as i64))));
        }
        res = ((x_sign | 0x31c0000000000000) | res);
        return (res, pfpsf);
    } else {
        res = (x_sign | 0x31c0000000000000);
        return (res, pfpsf);
    }
}
