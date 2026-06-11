// Auto-generated from bid128_nearbyint.go by go2rs. Do not edit.

use super::prelude::*;

pub fn bid128_nearbyint(mut x: BID_UINT128, mut rnd_mode: i64) -> (BID_UINT128, u32) {
    let mut res: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut x_sign: u64 = 0;
    let mut x_exp: u64 = 0;
    let mut exp: i64 = 0;
    let mut tmp64: u64 = 0;
    let mut x_nr_bits: u64 = 0;
    let mut q: i64 = 0;
    let mut ind: i64 = 0;
    let mut shift: i64 = 0;
    let mut C1: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut fstar: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut P256: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut pfpsf: u32 = 0;
    if ((x.w[1] & 0x7800000000000000) == 0x7800000000000000) {
        if ((x.w[1] & 0x7c00000000000000) == 0x7c00000000000000) {
            if ((((x.w[1] & 0x00003fffffffffff) > 0x0000314dc6448d93)) || ((((x.w[1] & 0x00003fffffffffff) == 0x0000314dc6448d93) && (x.w[0] > 0x38c15b09ffffffff)))) {
                x.w[1] = (x.w[1] & 0xffffc00000000000);
                x.w[0] = 0x0;
            }
            if ((x.w[1] & 0x7e00000000000000) == 0x7e00000000000000) {
                pfpsf |= 1;
                res.w[1] = (x.w[1] & 0xfc003fffffffffff);
                res.w[0] = x.w[0];
            } else {
                res.w[1] = (x.w[1] & 0xfc003fffffffffff);
                res.w[0] = x.w[0];
            }
            return (res, pfpsf);
        } else {
            if ((x.w[1] & 0x8000000000000000) == 0x0) {
                res.w[1] = 0x7800000000000000;
                res.w[0] = 0x0000000000000000;
            } else {
                res.w[1] = 0xf800000000000000;
                res.w[0] = 0x0000000000000000;
            }
            return (res, pfpsf);
        }
    }
    x_sign = (x.w[1] & 0x8000000000000000);
    C1.w[1] = (x.w[1] & 0x1ffffffffffff);
    C1.w[0] = x.w[0];
    if ((x.w[1] & 0x6000000000000000) == 0x6000000000000000) {
        x_exp = (((go_checked_shl_u64(x.w[1], go_shift_count_u64((2) as u64)))) & 0x7ffe000000000000);
        C1.w[1] = 0;
        C1.w[0] = 0;
    } else {
        x_exp = (x.w[1] & 0x7ffe000000000000);
        if ((C1.w[1] > 0x0001ed09bead87c0) || (((C1.w[1] == 0x0001ed09bead87c0) && (C1.w[0] > 0x378d8e63ffffffff)))) {
            C1.w[1] = 0;
            C1.w[0] = 0;
        }
    }
    if ((C1.w[1] == 0x0) && (C1.w[0] == 0x0)) {
        if (x_exp <= (0x1820 << 49)) {
            res.w[1] = ((x.w[1] & 0x8000000000000000) | 0x3040000000000000);
        } else {
            res.w[1] = (x_sign | x_exp);
        }
        res.w[0] = 0x0000000000000000;
        return (res, pfpsf);
    }
    match rnd_mode {
        0 | 4 => {
            if (x_exp <= 0x2ffa000000000000) {
                res.w[1] = (x_sign | 0x3040000000000000);
                res.w[0] = 0x0000000000000000;
                return (res, pfpsf);
            }
        }
        1 => {
            if (x_exp <= 0x2ffc000000000000) {
                if (x_sign != 0) {
                    res.w[1] = 0xb040000000000000;
                    res.w[0] = 0x0000000000000001;
                } else {
                    res.w[1] = 0x3040000000000000;
                    res.w[0] = 0x0000000000000000;
                }
                return (res, pfpsf);
            }
        }
        2 => {
            if (x_exp <= 0x2ffc000000000000) {
                if (x_sign != 0) {
                    res.w[1] = 0xb040000000000000;
                    res.w[0] = 0x0000000000000000;
                } else {
                    res.w[1] = 0x3040000000000000;
                    res.w[0] = 0x0000000000000001;
                }
                return (res, pfpsf);
            }
        }
        3 => {
            if (x_exp <= 0x2ffc000000000000) {
                res.w[1] = (x_sign | 0x3040000000000000);
                res.w[0] = 0x0000000000000000;
                return (res, pfpsf);
            }
        }
        _ => {}
    }
    let mut tmp1: u64 = 0;
    if (C1.w[1] == 0) {
        if (C1.w[0] >= 0x0020000000000000) {
            tmp1 = (((go_checked_shr_u64(C1.w[0], go_shift_count_u64((32) as u64))) as f64)).to_bits();
            x_nr_bits = ((33 as u64).wrapping_add(((((((go_checked_shr_u64(tmp1, go_shift_count_u64((52) as u64))) as u32) & 0x7ff)).wrapping_sub(0x3ff)) as u64)));
        } else {
            tmp1 = (C1.w[0] as f64).to_bits();
            x_nr_bits = ((1 as u64).wrapping_add(((((((go_checked_shr_u64(tmp1, go_shift_count_u64((52) as u64))) as u32) & 0x7ff)).wrapping_sub(0x3ff)) as u64)));
        }
    } else {
        tmp1 = (C1.w[1] as f64).to_bits();
        x_nr_bits = ((65 as u64).wrapping_add(((((((go_checked_shr_u64(tmp1, go_shift_count_u64((52) as u64))) as u32) & 0x7ff)).wrapping_sub(0x3ff)) as u64)));
    }
    q = (bid_nr_digits[(x_nr_bits.wrapping_sub(1)) as usize].digits as i64);
    if (q == 0) {
        q = (bid_nr_digits[(x_nr_bits.wrapping_sub(1)) as usize].digits1 as i64);
        if ((C1.w[1] > bid_nr_digits[(x_nr_bits.wrapping_sub(1)) as usize].threshold_hi) || (((C1.w[1] == bid_nr_digits[(x_nr_bits.wrapping_sub(1)) as usize].threshold_hi) && (C1.w[0] >= bid_nr_digits[(x_nr_bits.wrapping_sub(1)) as usize].threshold_lo)))) {
            q = q.wrapping_add(1);
        }
    }
    exp = (((go_checked_shr_u64(x_exp, go_shift_count_u64((49) as u64))) as i64).wrapping_sub(6176));
    if (exp >= 0) {
        res.w[1] = x.w[1];
        res.w[0] = x.w[0];
        return (res, pfpsf);
    }
    match rnd_mode {
        0 => {
            if ((q.wrapping_add(exp)) >= 0) {
                ind = (exp.wrapping_neg());
                tmp64 = C1.w[0];
                if (ind <= 19) {
                    C1.w[0] = (C1.w[0].wrapping_add(bid_midpoint64[(ind.wrapping_sub(1)) as usize]));
                } else {
                    C1.w[0] = (C1.w[0].wrapping_add(bid_midpoint128[(ind.wrapping_sub(20)) as usize].w[0]));
                    C1.w[1] = (C1.w[1].wrapping_add(bid_midpoint128[(ind.wrapping_sub(20)) as usize].w[1]));
                }
                if (C1.w[0] < tmp64) {
                    C1.w[1] = C1.w[1].wrapping_add(1);
                }
                P256 = __mul_128x128_to_256(C1, bid_ten2mk128[(ind.wrapping_sub(1)) as usize]);
                if ((ind.wrapping_sub(1)) <= 2) {
                    res.w[1] = P256.w[3];
                    res.w[0] = P256.w[2];
                    fstar.w[1] = P256.w[1];
                    fstar.w[0] = P256.w[0];
                    if ((((res.w[0] & 0x0000000000000001) != 0)) && ((((fstar.w[1] < bid_ten2mk128[(ind.wrapping_sub(1)) as usize].w[1])) || (((fstar.w[1] == bid_ten2mk128[(ind.wrapping_sub(1)) as usize].w[1]) && (fstar.w[0] < bid_ten2mk128[(ind.wrapping_sub(1)) as usize].w[0])))))) {
                        res.w[0] = res.w[0].wrapping_sub(1);
                    }
                } else if ((ind.wrapping_sub(1)) <= 21) {
                    shift = (bid_shiftright128[(ind.wrapping_sub(1)) as usize] as i64);
                    res.w[1] = (go_checked_shr_u64(P256.w[3], go_shift_count_u64((shift as u64) as u64)));
                    res.w[0] = (((go_checked_shl_u64(P256.w[3], go_shift_count_u64(((((64 as i64).wrapping_sub(shift)) as u64)) as u64)))) | ((go_checked_shr_u64(P256.w[2], go_shift_count_u64((shift as u64) as u64)))));
                    fstar.w[2] = (P256.w[2] & bid_maskhigh128[(ind.wrapping_sub(1)) as usize]);
                    fstar.w[1] = P256.w[1];
                    fstar.w[0] = P256.w[0];
                    if (((((res.w[0] & 0x0000000000000001) != 0)) && (fstar.w[2] == 0)) && (((fstar.w[1] < bid_ten2mk128[(ind.wrapping_sub(1)) as usize].w[1]) || (((fstar.w[1] == bid_ten2mk128[(ind.wrapping_sub(1)) as usize].w[1]) && (fstar.w[0] < bid_ten2mk128[(ind.wrapping_sub(1)) as usize].w[0])))))) {
                        res.w[0] = res.w[0].wrapping_sub(1);
                    }
                } else {
                    shift = ((bid_shiftright128[(ind.wrapping_sub(1)) as usize] as i64).wrapping_sub(64));
                    res.w[1] = 0;
                    res.w[0] = (go_checked_shr_u64(P256.w[3], go_shift_count_u64((shift as u64) as u64)));
                    fstar.w[3] = (P256.w[3] & bid_maskhigh128[(ind.wrapping_sub(1)) as usize]);
                    fstar.w[2] = P256.w[2];
                    fstar.w[1] = P256.w[1];
                    fstar.w[0] = P256.w[0];
                    if ((((((res.w[0] & 0x0000000000000001) != 0)) && (fstar.w[3] == 0)) && (fstar.w[2] == 0)) && (((fstar.w[1] < bid_ten2mk128[(ind.wrapping_sub(1)) as usize].w[1]) || (((fstar.w[1] == bid_ten2mk128[(ind.wrapping_sub(1)) as usize].w[1]) && (fstar.w[0] < bid_ten2mk128[(ind.wrapping_sub(1)) as usize].w[0])))))) {
                        res.w[0] = res.w[0].wrapping_sub(1);
                    }
                }
                res.w[1] = ((x_sign | 0x3040000000000000) | res.w[1]);
                return (res, pfpsf);
            } else {
                res.w[1] = (x_sign | 0x3040000000000000);
                res.w[0] = 0x0000000000000000;
                return (res, pfpsf);
            }
        }
        4 => {
            if ((q.wrapping_add(exp)) >= 0) {
                ind = (exp.wrapping_neg());
                tmp64 = C1.w[0];
                if (ind <= 19) {
                    C1.w[0] = (C1.w[0].wrapping_add(bid_midpoint64[(ind.wrapping_sub(1)) as usize]));
                } else {
                    C1.w[0] = (C1.w[0].wrapping_add(bid_midpoint128[(ind.wrapping_sub(20)) as usize].w[0]));
                    C1.w[1] = (C1.w[1].wrapping_add(bid_midpoint128[(ind.wrapping_sub(20)) as usize].w[1]));
                }
                if (C1.w[0] < tmp64) {
                    C1.w[1] = C1.w[1].wrapping_add(1);
                }
                P256 = __mul_128x128_to_256(C1, bid_ten2mk128[(ind.wrapping_sub(1)) as usize]);
                if ((ind.wrapping_sub(1)) <= 2) {
                    res.w[1] = P256.w[3];
                    res.w[0] = P256.w[2];
                } else if ((ind.wrapping_sub(1)) <= 21) {
                    shift = (bid_shiftright128[(ind.wrapping_sub(1)) as usize] as i64);
                    res.w[1] = (go_checked_shr_u64(P256.w[3], go_shift_count_u64((shift as u64) as u64)));
                    res.w[0] = (((go_checked_shl_u64(P256.w[3], go_shift_count_u64(((((64 as i64).wrapping_sub(shift)) as u64)) as u64)))) | ((go_checked_shr_u64(P256.w[2], go_shift_count_u64((shift as u64) as u64)))));
                } else {
                    shift = ((bid_shiftright128[(ind.wrapping_sub(1)) as usize] as i64).wrapping_sub(64));
                    res.w[1] = 0;
                    res.w[0] = (go_checked_shr_u64(P256.w[3], go_shift_count_u64((shift as u64) as u64)));
                }
                res.w[1] |= (x_sign | 0x3040000000000000);
                return (res, pfpsf);
            } else {
                res.w[1] = (x_sign | 0x3040000000000000);
                res.w[0] = 0x0000000000000000;
                return (res, pfpsf);
            }
        }
        1 => {
            if ((q.wrapping_add(exp)) > 0) {
                ind = (exp.wrapping_neg());
                P256 = __mul_128x128_to_256(C1, bid_ten2mk128[(ind.wrapping_sub(1)) as usize]);
                if ((ind.wrapping_sub(1)) <= 2) {
                    res.w[1] = P256.w[3];
                    res.w[0] = P256.w[2];
                    if (((P256.w[1] > bid_ten2mk128[(ind.wrapping_sub(1)) as usize].w[1])) || (((P256.w[1] == bid_ten2mk128[(ind.wrapping_sub(1)) as usize].w[1]) && (P256.w[0] >= bid_ten2mk128[(ind.wrapping_sub(1)) as usize].w[0])))) {
                        if (x_sign != 0) {
                            res.w[0] = res.w[0].wrapping_add(1);
                            if (res.w[0] == 0) {
                                res.w[1] = res.w[1].wrapping_add(1);
                            }
                        }
                    }
                } else if ((ind.wrapping_sub(1)) <= 21) {
                    shift = (bid_shiftright128[(ind.wrapping_sub(1)) as usize] as i64);
                    res.w[1] = (go_checked_shr_u64(P256.w[3], go_shift_count_u64((shift as u64) as u64)));
                    res.w[0] = (((go_checked_shl_u64(P256.w[3], go_shift_count_u64(((((64 as i64).wrapping_sub(shift)) as u64)) as u64)))) | ((go_checked_shr_u64(P256.w[2], go_shift_count_u64((shift as u64) as u64)))));
                    fstar.w[2] = (P256.w[2] & bid_maskhigh128[(ind.wrapping_sub(1)) as usize]);
                    fstar.w[1] = P256.w[1];
                    fstar.w[0] = P256.w[0];
                    if (((fstar.w[2] != 0) || (fstar.w[1] > bid_ten2mk128[(ind.wrapping_sub(1)) as usize].w[1])) || (((fstar.w[1] == bid_ten2mk128[(ind.wrapping_sub(1)) as usize].w[1]) && (fstar.w[0] >= bid_ten2mk128[(ind.wrapping_sub(1)) as usize].w[0])))) {
                        if (x_sign != 0) {
                            res.w[0] = res.w[0].wrapping_add(1);
                            if (res.w[0] == 0) {
                                res.w[1] = res.w[1].wrapping_add(1);
                            }
                        }
                    }
                } else {
                    shift = ((bid_shiftright128[(ind.wrapping_sub(1)) as usize] as i64).wrapping_sub(64));
                    res.w[1] = 0;
                    res.w[0] = (go_checked_shr_u64(P256.w[3], go_shift_count_u64((shift as u64) as u64)));
                    fstar.w[3] = (P256.w[3] & bid_maskhigh128[(ind.wrapping_sub(1)) as usize]);
                    fstar.w[2] = P256.w[2];
                    fstar.w[1] = P256.w[1];
                    fstar.w[0] = P256.w[0];
                    if ((((fstar.w[3] != 0) || (fstar.w[2] != 0)) || (fstar.w[1] > bid_ten2mk128[(ind.wrapping_sub(1)) as usize].w[1])) || (((fstar.w[1] == bid_ten2mk128[(ind.wrapping_sub(1)) as usize].w[1]) && (fstar.w[0] >= bid_ten2mk128[(ind.wrapping_sub(1)) as usize].w[0])))) {
                        if (x_sign != 0) {
                            res.w[0] = res.w[0].wrapping_add(1);
                            if (res.w[0] == 0) {
                                res.w[1] = res.w[1].wrapping_add(1);
                            }
                        }
                    }
                }
                res.w[1] = ((x_sign | 0x3040000000000000) | res.w[1]);
                return (res, pfpsf);
            } else {
                if (x_sign != 0) {
                    res.w[1] = 0xb040000000000000;
                    res.w[0] = 0x0000000000000001;
                } else {
                    res.w[1] = 0x3040000000000000;
                    res.w[0] = 0x0000000000000000;
                }
                return (res, pfpsf);
            }
        }
        2 => {
            if ((q.wrapping_add(exp)) > 0) {
                ind = (exp.wrapping_neg());
                P256 = __mul_128x128_to_256(C1, bid_ten2mk128[(ind.wrapping_sub(1)) as usize]);
                if ((ind.wrapping_sub(1)) <= 2) {
                    res.w[1] = P256.w[3];
                    res.w[0] = P256.w[2];
                    if (((P256.w[1] > bid_ten2mk128[(ind.wrapping_sub(1)) as usize].w[1])) || (((P256.w[1] == bid_ten2mk128[(ind.wrapping_sub(1)) as usize].w[1]) && (P256.w[0] >= bid_ten2mk128[(ind.wrapping_sub(1)) as usize].w[0])))) {
                        if (x_sign == 0) {
                            res.w[0] = res.w[0].wrapping_add(1);
                            if (res.w[0] == 0) {
                                res.w[1] = res.w[1].wrapping_add(1);
                            }
                        }
                    }
                } else if ((ind.wrapping_sub(1)) <= 21) {
                    shift = (bid_shiftright128[(ind.wrapping_sub(1)) as usize] as i64);
                    res.w[1] = (go_checked_shr_u64(P256.w[3], go_shift_count_u64((shift as u64) as u64)));
                    res.w[0] = (((go_checked_shl_u64(P256.w[3], go_shift_count_u64(((((64 as i64).wrapping_sub(shift)) as u64)) as u64)))) | ((go_checked_shr_u64(P256.w[2], go_shift_count_u64((shift as u64) as u64)))));
                    fstar.w[2] = (P256.w[2] & bid_maskhigh128[(ind.wrapping_sub(1)) as usize]);
                    fstar.w[1] = P256.w[1];
                    fstar.w[0] = P256.w[0];
                    if (((fstar.w[2] != 0) || (fstar.w[1] > bid_ten2mk128[(ind.wrapping_sub(1)) as usize].w[1])) || (((fstar.w[1] == bid_ten2mk128[(ind.wrapping_sub(1)) as usize].w[1]) && (fstar.w[0] >= bid_ten2mk128[(ind.wrapping_sub(1)) as usize].w[0])))) {
                        if (x_sign == 0) {
                            res.w[0] = res.w[0].wrapping_add(1);
                            if (res.w[0] == 0) {
                                res.w[1] = res.w[1].wrapping_add(1);
                            }
                        }
                    }
                } else {
                    shift = ((bid_shiftright128[(ind.wrapping_sub(1)) as usize] as i64).wrapping_sub(64));
                    res.w[1] = 0;
                    res.w[0] = (go_checked_shr_u64(P256.w[3], go_shift_count_u64((shift as u64) as u64)));
                    fstar.w[3] = (P256.w[3] & bid_maskhigh128[(ind.wrapping_sub(1)) as usize]);
                    fstar.w[2] = P256.w[2];
                    fstar.w[1] = P256.w[1];
                    fstar.w[0] = P256.w[0];
                    if ((((fstar.w[3] != 0) || (fstar.w[2] != 0)) || (fstar.w[1] > bid_ten2mk128[(ind.wrapping_sub(1)) as usize].w[1])) || (((fstar.w[1] == bid_ten2mk128[(ind.wrapping_sub(1)) as usize].w[1]) && (fstar.w[0] >= bid_ten2mk128[(ind.wrapping_sub(1)) as usize].w[0])))) {
                        if (x_sign == 0) {
                            res.w[0] = res.w[0].wrapping_add(1);
                            if (res.w[0] == 0) {
                                res.w[1] = res.w[1].wrapping_add(1);
                            }
                        }
                    }
                }
                res.w[1] = ((x_sign | 0x3040000000000000) | res.w[1]);
                return (res, pfpsf);
            } else {
                if (x_sign != 0) {
                    res.w[1] = 0xb040000000000000;
                    res.w[0] = 0x0000000000000000;
                } else {
                    res.w[1] = 0x3040000000000000;
                    res.w[0] = 0x0000000000000001;
                }
                return (res, pfpsf);
            }
        }
        3 => {
            if ((q.wrapping_add(exp)) > 0) {
                ind = (exp.wrapping_neg());
                P256 = __mul_128x128_to_256(C1, bid_ten2mk128[(ind.wrapping_sub(1)) as usize]);
                if ((ind.wrapping_sub(1)) <= 2) {
                    res.w[1] = P256.w[3];
                    res.w[0] = P256.w[2];
                } else if ((ind.wrapping_sub(1)) <= 21) {
                    shift = (bid_shiftright128[(ind.wrapping_sub(1)) as usize] as i64);
                    res.w[1] = (go_checked_shr_u64(P256.w[3], go_shift_count_u64((shift as u64) as u64)));
                    res.w[0] = (((go_checked_shl_u64(P256.w[3], go_shift_count_u64(((((64 as i64).wrapping_sub(shift)) as u64)) as u64)))) | ((go_checked_shr_u64(P256.w[2], go_shift_count_u64((shift as u64) as u64)))));
                } else {
                    shift = ((bid_shiftright128[(ind.wrapping_sub(1)) as usize] as i64).wrapping_sub(64));
                    res.w[1] = 0;
                    res.w[0] = (go_checked_shr_u64(P256.w[3], go_shift_count_u64((shift as u64) as u64)));
                }
                res.w[1] = ((x_sign | 0x3040000000000000) | res.w[1]);
                return (res, pfpsf);
            } else {
                res.w[1] = (x_sign | 0x3040000000000000);
                res.w[0] = 0x0000000000000000;
                return (res, pfpsf);
            }
        }
        _ => {}
    }
    return (res, pfpsf);
}
