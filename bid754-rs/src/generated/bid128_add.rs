// Auto-generated from bid128_add.go by go2rs. Do not edit.

use super::prelude::*;

pub fn bid128_add(mut x: BID_UINT128, mut y: BID_UINT128, mut rnd_mode: i64, pfpsf: &mut u32) -> BID_UINT128 {
    let mut res: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    res.w[0] = 0xbaddbaddbaddbadd;
    res.w[1] = 0xbaddbaddbaddbadd;
    let mut x_sign: u64 = 0;
    let mut y_sign: u64 = 0;
    let mut tmp_sign: u64 = 0;
    let mut x_exp: u64 = 0;
    let mut y_exp: u64 = 0;
    let mut tmp_exp: u64 = 0;
    let mut C1_hi: u64 = 0;
    let mut C2_hi: u64 = 0;
    let mut tmp_signif_hi: u64 = 0;
    let mut C1_lo: u64 = 0;
    let mut C2_lo: u64 = 0;
    let mut tmp_signif_lo: u64 = 0;
    let mut tmp64: u64 = 0;
    let mut tmp64A: u64 = 0;
    let mut tmp64B: u64 = 0;
    let mut x_nr_bits: i64 = 0;
    let mut y_nr_bits: i64 = 0;
    let mut q1: i64 = 0;
    let mut q2: i64 = 0;
    let mut delta: i64 = 0;
    let mut scale: i64 = 0;
    let mut x1: i64 = 0;
    let mut ind: i64 = 0;
    let mut shift: i64 = 0;
    let mut tmp_inexact: i64 = 0;
    let mut halfulp64: u64 = 0;
    let mut halfulp128: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut C1: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut C2: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut ten2m1: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut highf2star: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut P256: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut Q256: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut R256: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut is_inexact: i64 = 0;
    let mut is_midpoint_lt_even: i64 = 0;
    let mut is_midpoint_gt_even: i64 = 0;
    let mut is_inexact_lt_midpoint: i64 = 0;
    let mut is_inexact_gt_midpoint: i64 = 0;
    let mut second_pass: i64 = 0;
    x_sign = (x.w[1] & 0x8000000000000000);
    y_sign = (y.w[1] & 0x8000000000000000);
    if ((((x.w[1] & 0x7800000000000000) == 0x7800000000000000)) || (((y.w[1] & 0x7800000000000000) == 0x7800000000000000))) {
        if ((x.w[1] & 0x7c00000000000000) == 0x7c00000000000000) {
            if ((((x.w[1] & 0x00003fffffffffff) > 0x0000314dc6448d93)) || (((((x.w[1] & 0x00003fffffffffff) == 0x0000314dc6448d93)) && (x.w[0] > 0x38c15b09ffffffff)))) {
                x.w[1] = (x.w[1] & 0xffffc00000000000);
                x.w[0] = 0x0;
            }
            if ((x.w[1] & 0x7e00000000000000) == 0x7e00000000000000) {
                (*pfpsf) |= 1;
                res.w[1] = (x.w[1] & 0xfc003fffffffffff);
                res.w[0] = x.w[0];
            } else {
                res.w[1] = (x.w[1] & 0xfc003fffffffffff);
                res.w[0] = x.w[0];
                if ((y.w[1] & 0x7e00000000000000) == 0x7e00000000000000) {
                    (*pfpsf) |= 1;
                }
            }
            return res;
        } else if ((y.w[1] & 0x7c00000000000000) == 0x7c00000000000000) {
            if ((((y.w[1] & 0x00003fffffffffff) > 0x0000314dc6448d93)) || (((((y.w[1] & 0x00003fffffffffff) == 0x0000314dc6448d93)) && (y.w[0] > 0x38c15b09ffffffff)))) {
                y.w[1] = (y.w[1] & 0xffffc00000000000);
                y.w[0] = 0x0;
            }
            if ((y.w[1] & 0x7e00000000000000) == 0x7e00000000000000) {
                (*pfpsf) |= 1;
                res.w[1] = (y.w[1] & 0xfc003fffffffffff);
                res.w[0] = y.w[0];
            } else {
                res.w[1] = (y.w[1] & 0xfc003fffffffffff);
                res.w[0] = y.w[0];
            }
            return res;
        } else {
            if ((x.w[1] & 0x7800000000000000) == 0x7800000000000000) {
                if ((y.w[1] & 0x7800000000000000) == 0x7800000000000000) {
                    if ((x.w[1] & 0x8000000000000000) == (y.w[1] & 0x8000000000000000)) {
                        res.w[1] = (x_sign | 0x7800000000000000);
                        res.w[0] = 0x0;
                    } else {
                        (*pfpsf) |= 1;
                        res.w[1] = 0x7c00000000000000;
                        res.w[0] = 0x0000000000000000;
                    }
                } else {
                    res.w[1] = (x_sign | 0x7800000000000000);
                    res.w[0] = 0x0;
                }
            } else {
                res.w[1] = (y_sign | 0x7800000000000000);
                res.w[0] = 0x0;
            }
            return res;
        }
    }
    C1_hi = (x.w[1] & 0x1ffffffffffff);
    C1_lo = x.w[0];
    if ((x.w[1] & 0x6000000000000000) == 0x6000000000000000) {
        x_exp = (((go_checked_shl_u64(x.w[1], go_shift_count_u64((2) as u64)))) & 0x7ffe000000000000);
        C1_hi = 0;
        C1_lo = 0;
    } else {
        x_exp = (x.w[1] & 0x7ffe000000000000);
        if ((C1_hi > 0x0001ed09bead87c0) || (((C1_hi == 0x0001ed09bead87c0) && (C1_lo > 0x378d8e63ffffffff)))) {
            C1_hi = 0;
            C1_lo = 0;
        } else {
        }
    }
    C2_hi = (y.w[1] & 0x1ffffffffffff);
    C2_lo = y.w[0];
    if ((y.w[1] & 0x6000000000000000) == 0x6000000000000000) {
        y_exp = (((go_checked_shl_u64(y.w[1], go_shift_count_u64((2) as u64)))) & 0x7ffe000000000000);
        C2_hi = 0;
        C2_lo = 0;
    } else {
        y_exp = (y.w[1] & 0x7ffe000000000000);
        if ((C2_hi > 0x0001ed09bead87c0) || (((C2_hi == 0x0001ed09bead87c0) && (C2_lo > 0x378d8e63ffffffff)))) {
            C2_hi = 0;
            C2_lo = 0;
        } else {
        }
    }
    if ((C1_hi == 0x0) && (C1_lo == 0x0)) {
        if ((C2_hi == 0x0) && (C2_lo == 0x0)) {
            if (x_exp < y_exp) {
                res.w[1] = x_exp;
            } else {
                res.w[1] = y_exp;
            }
            if ((x_sign != 0) && (y_sign != 0)) {
                res.w[1] = (res.w[1] | x_sign);
            } else if ((rnd_mode == 1) && (x_sign != y_sign)) {
                res.w[1] = (res.w[1] | 0x8000000000000000);
            }
            res.w[0] = 0;
        } else {
            if (y_exp <= x_exp) {
                res.w[1] = y.w[1];
                res.w[0] = y.w[0];
            } else {
                if (C2_hi == 0) {
                    if (C2_lo >= 0x0020000000000000) {
                        let mut tmp2 = (((go_checked_shr_u64(C2_lo, go_shift_count_u64((32) as u64))) as f64)).to_bits();
                        y_nr_bits = ((33 as i64).wrapping_add((((((((go_checked_shr_u64(tmp2, go_shift_count_u64((52) as u64))) as u32)) & 0x7ff)).wrapping_sub(0x3ff)) as i64)));
                    } else {
                        let mut tmp2 = (C2_lo as f64).to_bits();
                        y_nr_bits = ((1 as i64).wrapping_add((((((((go_checked_shr_u64(tmp2, go_shift_count_u64((52) as u64))) as u32)) & 0x7ff)).wrapping_sub(0x3ff)) as i64)));
                    }
                } else {
                    let mut tmp2 = (C2_hi as f64).to_bits();
                    y_nr_bits = ((65 as i64).wrapping_add((((((((go_checked_shr_u64(tmp2, go_shift_count_u64((52) as u64))) as u32)) & 0x7ff)).wrapping_sub(0x3ff)) as i64)));
                }
                q2 = (bid_nr_digits[(y_nr_bits.wrapping_sub(1)) as usize].digits as i64);
                if (q2 == 0) {
                    q2 = (bid_nr_digits[(y_nr_bits.wrapping_sub(1)) as usize].digits1 as i64);
                    if ((C2_hi > bid_nr_digits[(y_nr_bits.wrapping_sub(1)) as usize].threshold_hi) || (((C2_hi == bid_nr_digits[(y_nr_bits.wrapping_sub(1)) as usize].threshold_hi) && (C2_lo >= bid_nr_digits[(y_nr_bits.wrapping_sub(1)) as usize].threshold_lo)))) {
                        q2 = q2.wrapping_add(1);
                    }
                }
                scale = ((34 as i64).wrapping_sub(q2));
                ind = ((go_checked_shr_u64(((y_exp.wrapping_sub(x_exp))), go_shift_count_u64((49) as u64))) as i64);
                if (ind < scale) {
                    scale = ind;
                }
                if (scale == 0) {
                    res.w[1] = y.w[1];
                    res.w[0] = y.w[0];
                } else if (q2 <= 19) {
                    if (scale <= 19) {
                        res = __mul_64x64_to_128(C2_lo, bid_ten2k64[scale as usize]);
                    } else {
                        res = __mul_128x64_to_128(C2_lo, bid_ten2k128[(scale.wrapping_sub(20)) as usize]);
                    }
                } else {
                    C2.w[1] = C2_hi;
                    C2.w[0] = C2_lo;
                    res = __mul_128x64_to_128(bid_ten2k64[scale as usize], C2);
                }
                y_exp = (y_exp.wrapping_sub(((go_checked_shl_u64((scale as u64), go_shift_count_u64((49) as u64))))));
                res.w[1] = ((res.w[1] | y_sign) | y_exp);
            }
        }
        return res;
    } else if ((C2_hi == 0x0) && (C2_lo == 0x0)) {
        if (x_exp <= y_exp) {
            res.w[1] = x.w[1];
            res.w[0] = x.w[0];
        } else {
            if (C1_hi == 0) {
                if (C1_lo >= 0x0020000000000000) {
                    let mut tmp1 = (((go_checked_shr_u64(C1_lo, go_shift_count_u64((32) as u64))) as f64)).to_bits();
                    x_nr_bits = ((33 as i64).wrapping_add((((((((go_checked_shr_u64(tmp1, go_shift_count_u64((52) as u64))) as u32)) & 0x7ff)).wrapping_sub(0x3ff)) as i64)));
                } else {
                    let mut tmp1 = (C1_lo as f64).to_bits();
                    x_nr_bits = ((1 as i64).wrapping_add((((((((go_checked_shr_u64(tmp1, go_shift_count_u64((52) as u64))) as u32)) & 0x7ff)).wrapping_sub(0x3ff)) as i64)));
                }
            } else {
                let mut tmp1 = (C1_hi as f64).to_bits();
                x_nr_bits = ((65 as i64).wrapping_add((((((((go_checked_shr_u64(tmp1, go_shift_count_u64((52) as u64))) as u32)) & 0x7ff)).wrapping_sub(0x3ff)) as i64)));
            }
            q1 = (bid_nr_digits[(x_nr_bits.wrapping_sub(1)) as usize].digits as i64);
            if (q1 == 0) {
                q1 = (bid_nr_digits[(x_nr_bits.wrapping_sub(1)) as usize].digits1 as i64);
                if ((C1_hi > bid_nr_digits[(x_nr_bits.wrapping_sub(1)) as usize].threshold_hi) || (((C1_hi == bid_nr_digits[(x_nr_bits.wrapping_sub(1)) as usize].threshold_hi) && (C1_lo >= bid_nr_digits[(x_nr_bits.wrapping_sub(1)) as usize].threshold_lo)))) {
                    q1 = q1.wrapping_add(1);
                }
            }
            scale = ((34 as i64).wrapping_sub(q1));
            ind = ((go_checked_shr_u64(((x_exp.wrapping_sub(y_exp))), go_shift_count_u64((49) as u64))) as i64);
            if (ind < scale) {
                scale = ind;
            }
            if (scale == 0) {
                res.w[1] = x.w[1];
                res.w[0] = x.w[0];
            } else if (q1 <= 19) {
                if (scale <= 19) {
                    res = __mul_64x64_to_128(C1_lo, bid_ten2k64[scale as usize]);
                } else {
                    res = __mul_128x64_to_128(C1_lo, bid_ten2k128[(scale.wrapping_sub(20)) as usize]);
                }
            } else {
                C1.w[1] = C1_hi;
                C1.w[0] = C1_lo;
                res = __mul_128x64_to_128(bid_ten2k64[scale as usize], C1);
            }
            x_exp = (x_exp.wrapping_sub(((go_checked_shl_u64((scale as u64), go_shift_count_u64((49) as u64))))));
            res.w[1] = ((res.w[1] | x_sign) | x_exp);
        }
        return res;
    } else {
        if (x_exp < y_exp) {
            tmp_sign = x_sign;
            tmp_exp = x_exp;
            tmp_signif_hi = C1_hi;
            tmp_signif_lo = C1_lo;
            x_sign = y_sign;
            x_exp = y_exp;
            C1_hi = C2_hi;
            C1_lo = C2_lo;
            y_sign = tmp_sign;
            y_exp = tmp_exp;
            C2_hi = tmp_signif_hi;
            C2_lo = tmp_signif_lo;
        }
        if (C1_hi == 0) {
            if (C1_lo >= 0x0020000000000000) {
                let mut tmp1 = (((go_checked_shr_u64(C1_lo, go_shift_count_u64((32) as u64))) as f64)).to_bits();
                x_nr_bits = ((33 as i64).wrapping_add((((((((go_checked_shr_u64(tmp1, go_shift_count_u64((52) as u64))) as u32)) & 0x7ff)).wrapping_sub(0x3ff)) as i64)));
            } else {
                let mut tmp1 = (C1_lo as f64).to_bits();
                x_nr_bits = ((1 as i64).wrapping_add((((((((go_checked_shr_u64(tmp1, go_shift_count_u64((52) as u64))) as u32)) & 0x7ff)).wrapping_sub(0x3ff)) as i64)));
            }
        } else {
            let mut tmp1 = (C1_hi as f64).to_bits();
            x_nr_bits = ((65 as i64).wrapping_add((((((((go_checked_shr_u64(tmp1, go_shift_count_u64((52) as u64))) as u32)) & 0x7ff)).wrapping_sub(0x3ff)) as i64)));
        }
        q1 = (bid_nr_digits[(x_nr_bits.wrapping_sub(1)) as usize].digits as i64);
        if (q1 == 0) {
            q1 = (bid_nr_digits[(x_nr_bits.wrapping_sub(1)) as usize].digits1 as i64);
            if ((C1_hi > bid_nr_digits[(x_nr_bits.wrapping_sub(1)) as usize].threshold_hi) || (((C1_hi == bid_nr_digits[(x_nr_bits.wrapping_sub(1)) as usize].threshold_hi) && (C1_lo >= bid_nr_digits[(x_nr_bits.wrapping_sub(1)) as usize].threshold_lo)))) {
                q1 = q1.wrapping_add(1);
            }
        }
        if (C2_hi == 0) {
            if (C2_lo >= 0x0020000000000000) {
                let mut tmp2 = (((go_checked_shr_u64(C2_lo, go_shift_count_u64((32) as u64))) as f64)).to_bits();
                y_nr_bits = ((33 as i64).wrapping_add((((((((go_checked_shr_u64(tmp2, go_shift_count_u64((52) as u64))) as u32)) & 0x7ff)).wrapping_sub(0x3ff)) as i64)));
            } else {
                let mut tmp2 = (C2_lo as f64).to_bits();
                y_nr_bits = ((1 as i64).wrapping_add((((((((go_checked_shr_u64(tmp2, go_shift_count_u64((52) as u64))) as u32)) & 0x7ff)).wrapping_sub(0x3ff)) as i64)));
            }
        } else {
            let mut tmp2 = (C2_hi as f64).to_bits();
            y_nr_bits = ((65 as i64).wrapping_add((((((((go_checked_shr_u64(tmp2, go_shift_count_u64((52) as u64))) as u32)) & 0x7ff)).wrapping_sub(0x3ff)) as i64)));
        }
        q2 = (bid_nr_digits[(y_nr_bits.wrapping_sub(1)) as usize].digits as i64);
        if (q2 == 0) {
            q2 = (bid_nr_digits[(y_nr_bits.wrapping_sub(1)) as usize].digits1 as i64);
            if ((C2_hi > bid_nr_digits[(y_nr_bits.wrapping_sub(1)) as usize].threshold_hi) || (((C2_hi == bid_nr_digits[(y_nr_bits.wrapping_sub(1)) as usize].threshold_hi) && (C2_lo >= bid_nr_digits[(y_nr_bits.wrapping_sub(1)) as usize].threshold_lo)))) {
                q2 = q2.wrapping_add(1);
            }
        }
        delta = (((q1.wrapping_add(((go_checked_shr_u64(x_exp, go_shift_count_u64((49) as u64))) as i64))).wrapping_sub(q2)).wrapping_sub(((go_checked_shr_u64(y_exp, go_shift_count_u64((49) as u64))) as i64)));
        if (delta >= 34) {
            if (delta >= (34 + 1)) {
                if (q1 < 34) {
                    scale = ((34 as i64).wrapping_sub(q1));
                    if (q1 <= 19) {
                        if (scale <= 19) {
                            C1 = __mul_64x64_to_128(bid_ten2k64[scale as usize], C1_lo);
                        } else {
                            C1_lo = (C1_lo.wrapping_mul(bid_ten2k64[(scale.wrapping_sub(19)) as usize]));
                            C1 = __mul_64x64_to_128(bid_ten2k64[19], C1_lo);
                        }
                    } else {
                        C1.w[1] = C1_hi;
                        C1.w[0] = C1_lo;
                        C1 = __mul_128x64_to_128(bid_ten2k64[((34 as i64).wrapping_sub(q1)) as usize], C1);
                    }
                    x_exp = (x_exp.wrapping_sub(((go_checked_shl_u64((scale as u64), go_shift_count_u64((49) as u64))))));
                    C1_hi = C1.w[1];
                    C1_lo = C1.w[0];
                }
                if ((((((((rnd_mode == 0) || (rnd_mode == 4))) && (delta == (34 + 1))) && (C1_hi == 0x0000314dc6448d93)) && (C1_lo == 0x38c15b0a00000000)) && (x_sign != y_sign)) && (((((q2 <= 19) && (C2_lo > bid_midpoint64[(q2.wrapping_sub(1)) as usize]))) || (((q2 >= 20) && (((C2_hi > bid_midpoint128[(q2.wrapping_sub(20)) as usize].w[1]) || (((C2_hi == bid_midpoint128[(q2.wrapping_sub(20)) as usize].w[1]) && (C2_lo > bid_midpoint128[(q2.wrapping_sub(20)) as usize].w[0])))))))))) {
                    C1_hi = 0x0001ed09bead87c0;
                    C1_lo = 0x378d8e63ffffffff;
                    x_exp = (x_exp.wrapping_sub(0x2000000000000));
                }
                if (rnd_mode != 0) {
                    if (((((rnd_mode == 1) && (x_sign != 0)) && (y_sign != 0))) || ((((rnd_mode == 2) && (x_sign == 0)) && (y_sign == 0)))) {
                        C1_lo = (C1_lo.wrapping_add(1));
                        if (C1_lo == 0) {
                            C1_hi = (C1_hi.wrapping_add(1));
                        }
                        if ((C1_hi == 0x0001ed09bead87c0) && (C1_lo == 0x378d8e6400000000)) {
                            C1_hi = 0x0000314dc6448d93;
                            C1_lo = 0x38c15b0a00000000;
                            x_exp = (x_exp.wrapping_add(0x2000000000000));
                            if (x_exp == 0x6000000000000000) {
                                C1_hi = 0x7800000000000000;
                                C1_lo = 0x0;
                                x_exp = 0;
                                (*pfpsf) |= 8;
                            }
                        }
                    } else if ((((((rnd_mode == 1) && (x_sign == 0)) && (y_sign != 0))) || ((((rnd_mode == 2) && (x_sign != 0)) && (y_sign == 0)))) || (((rnd_mode == 3) && (x_sign != y_sign)))) {
                        C1_lo = (C1_lo.wrapping_sub(1));
                        if (C1_lo == 0xffffffffffffffff) {
                            C1_hi = (C1_hi.wrapping_sub(1));
                        }
                        if ((C1_hi == 0x0000314dc6448d93) && (C1_lo == 0x38c15b09ffffffff)) {
                            C1_hi = 0x0001ed09bead87c0;
                            C1_lo = 0x378d8e63ffffffff;
                            x_exp = (x_exp.wrapping_sub(0x2000000000000));
                        }
                    } else {
                    }
                }
                (*pfpsf) |= 32;
                res.w[1] = ((x_sign | x_exp) | C1_hi);
                res.w[0] = C1_lo;
            } else {
                if (((x_sign == y_sign) || (((q1 <= 20) && (((C1_hi != 0) || (C1_lo != bid_ten2k64[(q1.wrapping_sub(1)) as usize])))))) || (((q1 >= 21) && (((C1_hi != bid_ten2k128[(q1.wrapping_sub(21)) as usize].w[1]) || (C1_lo != bid_ten2k128[(q1.wrapping_sub(21)) as usize].w[0])))))) {
                    if (q2 <= 19) {
                        halfulp64 = bid_midpoint64[(q2.wrapping_sub(1)) as usize];
                        if (C2_lo < halfulp64) {
                            if (q1 < 34) {
                                scale = ((34 as i64).wrapping_sub(q1));
                                if (q1 <= 19) {
                                    if (scale <= 19) {
                                        C1 = __mul_64x64_to_128(bid_ten2k64[scale as usize], C1_lo);
                                    } else {
                                        C1_lo = (C1_lo.wrapping_mul(bid_ten2k64[(scale.wrapping_sub(19)) as usize]));
                                        C1 = __mul_64x64_to_128(bid_ten2k64[19], C1_lo);
                                    }
                                } else {
                                    C1.w[1] = C1_hi;
                                    C1.w[0] = C1_lo;
                                    C1 = __mul_128x64_to_128(bid_ten2k64[((34 as i64).wrapping_sub(q1)) as usize], C1);
                                }
                                x_exp = (x_exp.wrapping_sub(((go_checked_shl_u64((scale as u64), go_shift_count_u64((49) as u64))))));
                                C1_hi = C1.w[1];
                                C1_lo = C1.w[0];
                            }
                            if (rnd_mode != 0) {
                                if (((((rnd_mode == 1) && (x_sign != 0)) && (y_sign != 0))) || ((((rnd_mode == 2) && (x_sign == 0)) && (y_sign == 0)))) {
                                    C1_lo = (C1_lo.wrapping_add(1));
                                    if (C1_lo == 0) {
                                        C1_hi = (C1_hi.wrapping_add(1));
                                    }
                                    if ((C1_hi == 0x0001ed09bead87c0) && (C1_lo == 0x378d8e6400000000)) {
                                        C1_hi = 0x0000314dc6448d93;
                                        C1_lo = 0x38c15b0a00000000;
                                        x_exp = (x_exp.wrapping_add(0x2000000000000));
                                        if (x_exp == 0x6000000000000000) {
                                            C1_hi = 0x7800000000000000;
                                            C1_lo = 0x0;
                                            x_exp = 0;
                                            (*pfpsf) |= 8;
                                        }
                                    }
                                } else if ((((((rnd_mode == 1) && (x_sign == 0)) && (y_sign != 0))) || ((((rnd_mode == 2) && (x_sign != 0)) && (y_sign == 0)))) || (((rnd_mode == 3) && (x_sign != y_sign)))) {
                                    C1_lo = (C1_lo.wrapping_sub(1));
                                    if (C1_lo == 0xffffffffffffffff) {
                                        C1_hi = (C1_hi.wrapping_sub(1));
                                    }
                                    if ((C1_hi == 0x0000314dc6448d93) && (C1_lo == 0x38c15b09ffffffff)) {
                                        C1_hi = 0x0001ed09bead87c0;
                                        C1_lo = 0x378d8e63ffffffff;
                                        x_exp = (x_exp.wrapping_sub(0x2000000000000));
                                    }
                                } else {
                                }
                            }
                            (*pfpsf) |= 32;
                            res.w[1] = ((x_sign | x_exp) | C1_hi);
                            res.w[0] = C1_lo;
                        } else if ((C2_lo == halfulp64) && (((q1 < 34) || (((C1_lo & 0x1) == 0))))) {
                            if (q1 < 34) {
                                scale = ((34 as i64).wrapping_sub(q1));
                                if (q1 <= 19) {
                                    if (scale <= 19) {
                                        C1 = __mul_64x64_to_128(bid_ten2k64[scale as usize], C1_lo);
                                    } else {
                                        C1_lo = (C1_lo.wrapping_mul(bid_ten2k64[(scale.wrapping_sub(19)) as usize]));
                                        C1 = __mul_64x64_to_128(bid_ten2k64[19], C1_lo);
                                    }
                                } else {
                                    C1.w[1] = C1_hi;
                                    C1.w[0] = C1_lo;
                                    C1 = __mul_128x64_to_128(bid_ten2k64[((34 as i64).wrapping_sub(q1)) as usize], C1);
                                }
                                x_exp = (x_exp.wrapping_sub(((go_checked_shl_u64((scale as u64), go_shift_count_u64((49) as u64))))));
                                C1_hi = C1.w[1];
                                C1_lo = C1.w[0];
                            }
                            if (((((((rnd_mode == 0) && (x_sign == y_sign)) && ((C1_lo & 0x01) != 0))) || (((rnd_mode == 4) && (x_sign == y_sign)))) || ((((rnd_mode == 2) && (x_sign == 0)) && (y_sign == 0)))) || ((((rnd_mode == 1) && (x_sign != 0)) && (y_sign != 0)))) {
                                C1_lo = (C1_lo.wrapping_add(1));
                                if (C1_lo == 0) {
                                    C1_hi = (C1_hi.wrapping_add(1));
                                }
                                if ((C1_hi == 0x0001ed09bead87c0) && (C1_lo == 0x378d8e6400000000)) {
                                    C1_hi = 0x0000314dc6448d93;
                                    C1_lo = 0x38c15b0a00000000;
                                    x_exp = (x_exp.wrapping_add(0x2000000000000));
                                    if (x_exp == 0x6000000000000000) {
                                        C1_hi = 0x7800000000000000;
                                        C1_lo = 0x0;
                                        x_exp = 0;
                                        (*pfpsf) |= 8;
                                    }
                                }
                            } else if (((((((rnd_mode == 0) && (x_sign != y_sign)) && ((C1_lo & 0x01) != 0))) || ((((rnd_mode == 1) && (x_sign == 0)) && (y_sign != 0)))) || ((((rnd_mode == 2) && (x_sign != 0)) && (y_sign == 0)))) || (((rnd_mode == 3) && (x_sign != y_sign)))) {
                                C1_lo = (C1_lo.wrapping_sub(1));
                                if (C1_lo == 0xffffffffffffffff) {
                                    C1_hi = (C1_hi.wrapping_sub(1));
                                }
                                if ((C1_hi == 0x0000314dc6448d93) && (C1_lo == 0x38c15b09ffffffff)) {
                                    C1_hi = 0x0001ed09bead87c0;
                                    C1_lo = 0x378d8e63ffffffff;
                                    x_exp = (x_exp.wrapping_sub(0x2000000000000));
                                }
                            } else {
                            }
                            (*pfpsf) |= 32;
                            res.w[1] = ((x_sign | x_exp) | C1_hi);
                            res.w[0] = C1_lo;
                        } else {
                            if (q1 < 34) {
                                scale = ((34 as i64).wrapping_sub(q1));
                                if (q1 <= 19) {
                                    if (scale <= 19) {
                                        C1 = __mul_64x64_to_128(bid_ten2k64[scale as usize], C1_lo);
                                    } else {
                                        C1_lo = (C1_lo.wrapping_mul(bid_ten2k64[(scale.wrapping_sub(19)) as usize]));
                                        C1 = __mul_64x64_to_128(bid_ten2k64[19], C1_lo);
                                    }
                                } else {
                                    C1.w[1] = C1_hi;
                                    C1.w[0] = C1_lo;
                                    C1 = __mul_128x64_to_128(bid_ten2k64[((34 as i64).wrapping_sub(q1)) as usize], C1);
                                }
                                x_exp = (x_exp.wrapping_sub(((go_checked_shl_u64((scale as u64), go_shift_count_u64((49) as u64))))));
                                C1_hi = C1.w[1];
                                C1_lo = C1.w[0];
                                if ((C1_hi == 0x0001ed09bead87c0) && (C1_lo == 0x378d8e6400000000)) {
                                    C1_hi = 0x0000314dc6448d93;
                                    C1_lo = 0x38c15b0a00000000;
                                    x_exp = (x_exp.wrapping_add(0x2000000000000));
                                }
                            }
                            if (((((((rnd_mode == 0) && (x_sign != y_sign))) || ((((rnd_mode == 4) && (x_sign != y_sign)) && (C2_lo != halfulp64)))) || ((((rnd_mode == 1) && (x_sign == 0)) && (y_sign != 0)))) || ((((rnd_mode == 2) && (x_sign != 0)) && (y_sign == 0)))) || (((rnd_mode == 3) && (x_sign != y_sign)))) {
                                C1_lo = (C1_lo.wrapping_sub(1));
                                if (C1_lo == 0xffffffffffffffff) {
                                    C1_hi = C1_hi.wrapping_sub(1);
                                }
                                if ((C1_hi == 0x0000314dc6448d93) && (C1_lo == 0x38c15b09ffffffff)) {
                                    C1_hi = 0x0001ed09bead87c0;
                                    C1_lo = 0x378d8e63ffffffff;
                                    x_exp = (x_exp.wrapping_sub(0x2000000000000));
                                }
                            } else if ((((((rnd_mode == 0) && (x_sign == y_sign))) || (((rnd_mode == 4) && (x_sign == y_sign)))) || ((((rnd_mode == 1) && (x_sign != 0)) && (y_sign != 0)))) || ((((rnd_mode == 2) && (x_sign == 0)) && (y_sign == 0)))) {
                                C1_lo = (C1_lo.wrapping_add(1));
                                if (C1_lo == 0) {
                                    C1_hi = (C1_hi.wrapping_add(1));
                                }
                                if ((C1_hi == 0x0001ed09bead87c0) && (C1_lo == 0x378d8e6400000000)) {
                                    C1_hi = 0x0000314dc6448d93;
                                    C1_lo = 0x38c15b0a00000000;
                                    x_exp = (x_exp.wrapping_add(0x2000000000000));
                                    if (x_exp == 0x6000000000000000) {
                                        C1_hi = 0x7800000000000000;
                                        C1_lo = 0x0;
                                        x_exp = 0;
                                        (*pfpsf) |= 8;
                                    }
                                }
                            } else {
                            }
                            (*pfpsf) |= 32;
                            res.w[1] = ((x_sign | x_exp) | C1_hi);
                            res.w[0] = C1_lo;
                        }
                    } else {
                        halfulp128 = bid_midpoint128[(q2.wrapping_sub(20)) as usize];
                        if ((C2_hi < halfulp128.w[1]) || (((C2_hi == halfulp128.w[1]) && (C2_lo < halfulp128.w[0])))) {
                            if (q1 < 34) {
                                scale = ((34 as i64).wrapping_sub(q1));
                                if (q1 <= 19) {
                                    if (scale <= 19) {
                                        C1 = __mul_64x64_to_128(bid_ten2k64[scale as usize], C1_lo);
                                    } else {
                                        C1_lo = (C1_lo.wrapping_mul(bid_ten2k64[(scale.wrapping_sub(19)) as usize]));
                                        C1 = __mul_64x64_to_128(bid_ten2k64[19], C1_lo);
                                    }
                                } else {
                                    C1.w[1] = C1_hi;
                                    C1.w[0] = C1_lo;
                                    C1 = __mul_128x64_to_128(bid_ten2k64[((34 as i64).wrapping_sub(q1)) as usize], C1);
                                }
                                C1_hi = C1.w[1];
                                C1_lo = C1.w[0];
                                x_exp = (x_exp.wrapping_sub(((go_checked_shl_u64((scale as u64), go_shift_count_u64((49) as u64))))));
                            }
                            if (rnd_mode != 0) {
                                if (((((rnd_mode == 1) && (x_sign != 0)) && (y_sign != 0))) || ((((rnd_mode == 2) && (x_sign == 0)) && (y_sign == 0)))) {
                                    C1_lo = (C1_lo.wrapping_add(1));
                                    if (C1_lo == 0) {
                                        C1_hi = (C1_hi.wrapping_add(1));
                                    }
                                    if ((C1_hi == 0x0001ed09bead87c0) && (C1_lo == 0x378d8e6400000000)) {
                                        C1_hi = 0x0000314dc6448d93;
                                        C1_lo = 0x38c15b0a00000000;
                                        x_exp = (x_exp.wrapping_add(0x2000000000000));
                                        if (x_exp == 0x6000000000000000) {
                                            C1_hi = 0x7800000000000000;
                                            C1_lo = 0x0;
                                            x_exp = 0;
                                            (*pfpsf) |= 8;
                                        }
                                    }
                                } else if ((((((rnd_mode == 1) && (x_sign == 0)) && (y_sign != 0))) || ((((rnd_mode == 2) && (x_sign != 0)) && (y_sign == 0)))) || (((rnd_mode == 3) && (x_sign != y_sign)))) {
                                    C1_lo = (C1_lo.wrapping_sub(1));
                                    if (C1_lo == 0xffffffffffffffff) {
                                        C1_hi = (C1_hi.wrapping_sub(1));
                                    }
                                    if ((C1_hi == 0x0000314dc6448d93) && (C1_lo == 0x38c15b09ffffffff)) {
                                        C1_hi = 0x0001ed09bead87c0;
                                        C1_lo = 0x378d8e63ffffffff;
                                        x_exp = (x_exp.wrapping_sub(0x2000000000000));
                                    }
                                } else {
                                }
                            }
                            (*pfpsf) |= 32;
                            res.w[1] = ((x_sign | x_exp) | C1_hi);
                            res.w[0] = C1_lo;
                        } else if ((((C2_hi == halfulp128.w[1]) && (C2_lo == halfulp128.w[0]))) && (((q1 < 34) || (((C1_lo & 0x1) == 0))))) {
                            if (q1 < 34) {
                                scale = ((34 as i64).wrapping_sub(q1));
                                if (q1 <= 19) {
                                    if (scale <= 19) {
                                        C1 = __mul_64x64_to_128(bid_ten2k64[scale as usize], C1_lo);
                                    } else {
                                        C1_lo = (C1_lo.wrapping_mul(bid_ten2k64[(scale.wrapping_sub(19)) as usize]));
                                        C1 = __mul_64x64_to_128(bid_ten2k64[19], C1_lo);
                                    }
                                } else {
                                    C1.w[1] = C1_hi;
                                    C1.w[0] = C1_lo;
                                    C1 = __mul_128x64_to_128(bid_ten2k64[((34 as i64).wrapping_sub(q1)) as usize], C1);
                                }
                                x_exp = (x_exp.wrapping_sub(((go_checked_shl_u64((scale as u64), go_shift_count_u64((49) as u64))))));
                                C1_hi = C1.w[1];
                                C1_lo = C1.w[0];
                            }
                            if (rnd_mode != 0) {
                                if (((((rnd_mode == 4) && (x_sign == y_sign))) || ((((rnd_mode == 2) && (x_sign == 0)) && (y_sign == 0)))) || ((((rnd_mode == 1) && (x_sign != 0)) && (y_sign != 0)))) {
                                    C1_lo = (C1_lo.wrapping_add(1));
                                    if (C1_lo == 0) {
                                        C1_hi = (C1_hi.wrapping_add(1));
                                    }
                                    if ((C1_hi == 0x0001ed09bead87c0) && (C1_lo == 0x378d8e6400000000)) {
                                        C1_hi = 0x0000314dc6448d93;
                                        C1_lo = 0x38c15b0a00000000;
                                        x_exp = (x_exp.wrapping_add(0x2000000000000));
                                        if (x_exp == 0x6000000000000000) {
                                            C1_hi = 0x7800000000000000;
                                            C1_lo = 0x0;
                                            x_exp = 0;
                                            (*pfpsf) |= 8;
                                        }
                                    }
                                } else if ((((((rnd_mode == 1) && (x_sign == 0)) && (y_sign != 0))) || ((((rnd_mode == 2) && (x_sign != 0)) && (y_sign == 0)))) || (((rnd_mode == 3) && (x_sign != y_sign)))) {
                                    C1_lo = (C1_lo.wrapping_sub(1));
                                    if (C1_lo == 0xffffffffffffffff) {
                                        C1_hi = (C1_hi.wrapping_sub(1));
                                    }
                                    if ((C1_hi == 0x0000314dc6448d93) && (C1_lo == 0x38c15b09ffffffff)) {
                                        C1_hi = 0x0001ed09bead87c0;
                                        C1_lo = 0x378d8e63ffffffff;
                                        x_exp = (x_exp.wrapping_sub(0x2000000000000));
                                    }
                                } else {
                                }
                            }
                            (*pfpsf) |= 32;
                            res.w[1] = ((x_sign | x_exp) | C1_hi);
                            res.w[0] = C1_lo;
                        } else {
                            if (q1 < 34) {
                                scale = ((34 as i64).wrapping_sub(q1));
                                if (q1 <= 19) {
                                    if (scale <= 19) {
                                        C1 = __mul_64x64_to_128(bid_ten2k64[scale as usize], C1_lo);
                                    } else {
                                        C1_lo = (C1_lo.wrapping_mul(bid_ten2k64[(scale.wrapping_sub(19)) as usize]));
                                        C1 = __mul_64x64_to_128(bid_ten2k64[19], C1_lo);
                                    }
                                } else {
                                    C1.w[1] = C1_hi;
                                    C1.w[0] = C1_lo;
                                    C1 = __mul_128x64_to_128(bid_ten2k64[((34 as i64).wrapping_sub(q1)) as usize], C1);
                                }
                                C1_hi = C1.w[1];
                                C1_lo = C1.w[0];
                                x_exp = (x_exp.wrapping_sub(((go_checked_shl_u64((scale as u64), go_shift_count_u64((49) as u64))))));
                            }
                            if (((((((rnd_mode == 0) && (x_sign != y_sign))) || ((((rnd_mode == 4) && (x_sign != y_sign)) && (((C2_hi != halfulp128.w[1]) || (C2_lo != halfulp128.w[0])))))) || ((((rnd_mode == 1) && (x_sign == 0)) && (y_sign != 0)))) || ((((rnd_mode == 2) && (x_sign != 0)) && (y_sign == 0)))) || (((rnd_mode == 3) && (x_sign != y_sign)))) {
                                C1_lo = (C1_lo.wrapping_sub(1));
                                if (C1_lo == 0xffffffffffffffff) {
                                    C1_hi = C1_hi.wrapping_sub(1);
                                }
                                if ((C1_hi == 0x0000314dc6448d93) && (C1_lo == 0x38c15b09ffffffff)) {
                                    C1_hi = 0x0001ed09bead87c0;
                                    C1_lo = 0x378d8e63ffffffff;
                                    x_exp = (x_exp.wrapping_sub(0x2000000000000));
                                }
                            } else if ((((((rnd_mode == 0) && (x_sign == y_sign))) || (((rnd_mode == 4) && (x_sign == y_sign)))) || ((((rnd_mode == 1) && (x_sign != 0)) && (y_sign != 0)))) || ((((rnd_mode == 2) && (x_sign == 0)) && (y_sign == 0)))) {
                                C1_lo = (C1_lo.wrapping_add(1));
                                if (C1_lo == 0) {
                                    C1_hi = (C1_hi.wrapping_add(1));
                                }
                                if ((C1_hi == 0x0001ed09bead87c0) && (C1_lo == 0x378d8e6400000000)) {
                                    C1_hi = 0x0000314dc6448d93;
                                    C1_lo = 0x38c15b0a00000000;
                                    x_exp = (x_exp.wrapping_add(0x2000000000000));
                                    if (x_exp == 0x6000000000000000) {
                                        C1_hi = 0x7800000000000000;
                                        C1_lo = 0x0;
                                        x_exp = 0;
                                        (*pfpsf) |= 8;
                                    }
                                }
                            } else {
                            }
                            (*pfpsf) |= 32;
                            res.w[1] = ((x_sign | x_exp) | C1_hi);
                            res.w[0] = C1_lo;
                        }
                    }
                } else {
                    x1 = (q2.wrapping_sub(1));
                    scale = (((34 as i64).wrapping_sub(q1)).wrapping_add(1));
                    if (scale >= 20) {
                        C1 = __mul_128x64_to_128(C1_lo, bid_ten2k128[(scale.wrapping_sub(20)) as usize]);
                    } else {
                        if (q1 <= 19) {
                            C1 = __mul_64x64_to_128(C1_lo, bid_ten2k64[scale as usize]);
                        } else {
                            C1.w[1] = C1_hi;
                            C1.w[0] = C1_lo;
                            C1 = __mul_128x64_to_128(bid_ten2k64[scale as usize], C1);
                        }
                    }
                    tmp64 = C1.w[0];
                    ind = (x1.wrapping_sub(1));
                    if (ind >= 0) {
                        C2.w[0] = C2_lo;
                        C2.w[1] = C2_hi;
                        if (ind <= 18) {
                            C2.w[0] = (C2.w[0].wrapping_add(bid_midpoint64[ind as usize]));
                            if (C2.w[0] < C2_lo) {
                                C2.w[1] = C2.w[1].wrapping_add(1);
                            }
                        } else {
                            C2.w[0] = (C2.w[0].wrapping_add(bid_midpoint128[(ind.wrapping_sub(19)) as usize].w[0]));
                            C2.w[1] = (C2.w[1].wrapping_add(bid_midpoint128[(ind.wrapping_sub(19)) as usize].w[1]));
                            if (C2.w[0] < C2_lo) {
                                C2.w[1] = C2.w[1].wrapping_add(1);
                            }
                        }
                        R256 = __mul_128x128_to_256(C2, bid_ten2mk128[ind as usize]);
                        if (ind <= 2) {
                            highf2star.w[1] = 0x0;
                            highf2star.w[0] = 0x0;
                        } else if (ind <= 21) {
                            highf2star.w[1] = 0x0;
                            highf2star.w[0] = (R256.w[2] & bid_maskhigh128[ind as usize]);
                        } else {
                            highf2star.w[1] = (R256.w[3] & bid_maskhigh128[ind as usize]);
                            highf2star.w[0] = R256.w[2];
                        }
                        if (ind >= 3) {
                            shift = (bid_shiftright128[ind as usize] as i64);
                            if (shift < 64) {
                                R256.w[2] = (((go_checked_shr_u64(R256.w[2], go_shift_count_u64((shift as u64) as u64)))) | ((go_checked_shl_u64(R256.w[3], go_shift_count_u64(((((64 as i64).wrapping_sub(shift)) as u64)) as u64)))));
                                R256.w[3] = ((go_checked_shr_u64(R256.w[3], go_shift_count_u64((shift as u64) as u64))));
                            } else {
                                R256.w[2] = ((go_checked_shr_u64(R256.w[3], go_shift_count_u64((((shift.wrapping_sub(64)) as u64)) as u64))));
                                R256.w[3] = 0x0;
                            }
                        }
                        is_inexact_lt_midpoint = 0;
                        is_inexact_gt_midpoint = 0;
                        is_midpoint_lt_even = 0;
                        is_midpoint_gt_even = 0;
                        if (ind <= 2) {
                            if ((R256.w[1] > 0x8000000000000000) || (((R256.w[1] == 0x8000000000000000) && (R256.w[0] > 0x0)))) {
                                tmp64A = (R256.w[1].wrapping_sub(0x8000000000000000));
                                if ((tmp64A > bid_ten2mk128trunc[ind as usize].w[1]) || (((tmp64A == bid_ten2mk128trunc[ind as usize].w[1]) && (R256.w[0] >= bid_ten2mk128trunc[ind as usize].w[0])))) {
                                    (*pfpsf) |= 32;
                                    is_inexact_gt_midpoint = 1;
                                }
                            } else {
                                (*pfpsf) |= 32;
                                is_inexact_lt_midpoint = 1;
                            }
                        } else if (ind <= 21) {
                            if (((highf2star.w[1] > 0x0) || (((highf2star.w[1] == 0x0) && (highf2star.w[0] > bid_onehalf128[ind as usize])))) || ((((highf2star.w[1] == 0x0) && (highf2star.w[0] == bid_onehalf128[ind as usize])) && (((R256.w[1] != 0) || (R256.w[0] != 0)))))) {
                                tmp64A = (highf2star.w[0].wrapping_sub(bid_onehalf128[ind as usize]));
                                tmp64B = highf2star.w[1];
                                if (tmp64A > highf2star.w[0]) {
                                    tmp64B = tmp64B.wrapping_sub(1);
                                }
                                if ((((tmp64B != 0) || (tmp64A != 0)) || (R256.w[1] > bid_ten2mk128trunc[ind as usize].w[1])) || (((R256.w[1] == bid_ten2mk128trunc[ind as usize].w[1]) && (R256.w[0] > bid_ten2mk128trunc[ind as usize].w[0])))) {
                                    (*pfpsf) |= 32;
                                    is_inexact_gt_midpoint = 1;
                                }
                            } else {
                                (*pfpsf) |= 32;
                                is_inexact_lt_midpoint = 1;
                            }
                        } else {
                            if ((highf2star.w[1] > bid_onehalf128[ind as usize]) || (((highf2star.w[1] == bid_onehalf128[ind as usize]) && ((((highf2star.w[0] != 0) || (R256.w[1] != 0)) || (R256.w[0] != 0)))))) {
                                tmp64B = (highf2star.w[1].wrapping_sub(bid_onehalf128[ind as usize]));
                                if ((((tmp64B != 0) || (highf2star.w[0] != 0)) || (R256.w[1] > bid_ten2mk128trunc[ind as usize].w[1])) || (((R256.w[1] == bid_ten2mk128trunc[ind as usize].w[1]) && (R256.w[0] > bid_ten2mk128trunc[ind as usize].w[0])))) {
                                    (*pfpsf) |= 32;
                                    is_inexact_gt_midpoint = 1;
                                }
                            } else {
                                (*pfpsf) |= 32;
                                is_inexact_lt_midpoint = 1;
                            }
                        }
                        if ((((((R256.w[1] != 0) || (R256.w[0] != 0))) && (highf2star.w[1] == 0)) && (highf2star.w[0] == 0)) && (((R256.w[1] < bid_ten2mk128trunc[ind as usize].w[1]) || (((R256.w[1] == bid_ten2mk128trunc[ind as usize].w[1]) && (R256.w[0] <= bid_ten2mk128trunc[ind as usize].w[0])))))) {
                            if ((((tmp64.wrapping_add(R256.w[2]))) & 0x01) != 0) {
                                R256.w[2] = R256.w[2].wrapping_sub(1);
                                if (R256.w[2] == 0xffffffffffffffff) {
                                    R256.w[3] = R256.w[3].wrapping_sub(1);
                                }
                                is_midpoint_lt_even = 1;
                                is_inexact_lt_midpoint = 0;
                                is_inexact_gt_midpoint = 0;
                            } else {
                                is_midpoint_gt_even = 1;
                                is_inexact_lt_midpoint = 0;
                                is_inexact_gt_midpoint = 0;
                            }
                        }
                    } else {
                        R256.w[2] = C2_lo;
                        R256.w[3] = C2_hi;
                        is_midpoint_lt_even = 0;
                        is_midpoint_gt_even = 0;
                        is_inexact_lt_midpoint = 0;
                        is_inexact_gt_midpoint = 0;
                    }
                    C1.w[0] = (C1.w[0].wrapping_sub(R256.w[2]));
                    C1.w[1] = (C1.w[1].wrapping_sub(R256.w[3]));
                    if (C1.w[0] > tmp64) {
                        C1.w[1] = C1.w[1].wrapping_sub(1);
                    }
                    if (C1.w[1] >= 0x8000000000000000) {
                        C1.w[0] = (!C1.w[0]);
                        C1.w[0] = C1.w[0].wrapping_add(1);
                        C1.w[1] = (!C1.w[1]);
                        if (C1.w[0] == 0x0) {
                            C1.w[1] = C1.w[1].wrapping_add(1);
                        }
                        tmp_sign = y_sign;
                    } else {
                        tmp_sign = x_sign;
                    }
                    x_sign = tmp_sign;
                    if (x1 >= 1) {
                        y_exp = (y_exp.wrapping_add(((go_checked_shl_u64((x1 as u64), go_shift_count_u64((49) as u64))))));
                    }
                    C1_hi = C1.w[1];
                    C1_lo = C1.w[0];
                    if (rnd_mode != 0) {
                        if ((((x_sign == 0) && (((((rnd_mode == 2) && (is_inexact_lt_midpoint != 0))) || (((((rnd_mode == 4) || (rnd_mode == 2))) && (is_midpoint_gt_even != 0))))))) || (((x_sign != 0) && (((((rnd_mode == 1) && (is_inexact_lt_midpoint != 0))) || (((((rnd_mode == 4) || (rnd_mode == 1))) && (is_midpoint_gt_even != 0)))))))) {
                            C1_lo = (C1_lo.wrapping_add(1));
                            if (C1_lo == 0) {
                                C1_hi = (C1_hi.wrapping_add(1));
                            }
                            if ((C1_hi == 0x0001ed09bead87c0) && (C1_lo == 0x378d8e6400000000)) {
                                C1_hi = 0x0000314dc6448d93;
                                C1_lo = 0x38c15b0a00000000;
                                y_exp = (y_exp.wrapping_add(0x2000000000000));
                            }
                        } else if ((((is_midpoint_lt_even != 0) || (is_inexact_gt_midpoint != 0))) && (((((x_sign != 0) && (((rnd_mode == 2) || (rnd_mode == 3))))) || (((x_sign == 0) && (((rnd_mode == 1) || (rnd_mode == 3)))))))) {
                            C1_lo = (C1_lo.wrapping_sub(1));
                            if (C1_lo == 0xffffffffffffffff) {
                                C1_hi = C1_hi.wrapping_sub(1);
                            }
                            if ((C1_hi == 0x0000314dc6448d93) && (C1_lo == 0x38c15b09ffffffff)) {
                                C1_hi = 0x0001ed09bead87c0;
                                C1_lo = 0x378d8e63ffffffff;
                                y_exp = (y_exp.wrapping_sub(0x2000000000000));
                            }
                        } else {
                        }
                    }
                    res.w[1] = ((x_sign | y_exp) | C1_hi);
                    res.w[0] = C1_lo;
                }
            }
        } else {
            if (delta >= 0) {
                if (delta <= (((34 - 1) as i64).wrapping_sub(q2))) {
                    scale = ((delta.wrapping_sub(q1)).wrapping_add(q2));
                    if (scale >= 20) {
                        C1 = __mul_128x64_to_128(C1_lo, bid_ten2k128[(scale.wrapping_sub(20)) as usize]);
                        C1_hi = C1.w[1];
                        C1_lo = C1.w[0];
                    } else if (scale >= 1) {
                        if (q1 <= 19) {
                            C1 = __mul_64x64_to_128(C1_lo, bid_ten2k64[scale as usize]);
                        } else {
                            C1.w[1] = C1_hi;
                            C1.w[0] = C1_lo;
                            C1 = __mul_128x64_to_128(bid_ten2k64[scale as usize], C1);
                        }
                        C1_hi = C1.w[1];
                        C1_lo = C1.w[0];
                    } else {
                        C1.w[0] = C1_lo;
                    }
                    if (x_sign == y_sign) {
                        C1_lo = (C1_lo.wrapping_add(C2_lo));
                        C1_hi = (C1_hi.wrapping_add(C2_hi));
                        if (C1_lo < C1.w[0]) {
                            C1_hi = C1_hi.wrapping_add(1);
                        }
                    } else {
                        C1_lo = (C1_lo.wrapping_sub(C2_lo));
                        C1_hi = (C1_hi.wrapping_sub(C2_hi));
                        if (C1_lo > C1.w[0]) {
                            C1_hi = C1_hi.wrapping_sub(1);
                        }
                        if ((C1_lo == 0) && (C1_hi == 0)) {
                            if (x_exp < y_exp) {
                                res.w[1] = x_exp;
                            } else {
                                res.w[1] = y_exp;
                            }
                            res.w[0] = 0;
                            if (rnd_mode == 1) {
                                res.w[1] |= 0x8000000000000000;
                            }
                            return res;
                        }
                        if (C1_hi >= 0x8000000000000000) {
                            C1_lo = (!C1_lo);
                            C1_lo = C1_lo.wrapping_add(1);
                            C1_hi = (!C1_hi);
                            if (C1_lo == 0x0) {
                                C1_hi = C1_hi.wrapping_add(1);
                            }
                            x_sign = y_sign;
                        }
                    }
                    res.w[1] = ((x_sign | y_exp) | C1_hi);
                    res.w[0] = C1_lo;
                } else if (delta == ((34 as i64).wrapping_sub(q2))) {
                    scale = ((delta.wrapping_sub(q1)).wrapping_add(q2));
                    if (scale >= 20) {
                        C1 = __mul_128x64_to_128(C1_lo, bid_ten2k128[(scale.wrapping_sub(20)) as usize]);
                    } else if (scale >= 1) {
                        if (q1 <= 19) {
                            C1 = __mul_64x64_to_128(C1_lo, bid_ten2k64[scale as usize]);
                        } else {
                            C1.w[1] = C1_hi;
                            C1.w[0] = C1_lo;
                            C1 = __mul_128x64_to_128(bid_ten2k64[scale as usize], C1);
                        }
                    } else {
                        C1.w[1] = C1_hi;
                        C1.w[0] = C1_lo;
                    }
                    C1_hi = C1.w[1];
                    C1_lo = C1.w[0];
                    if (x_sign == y_sign) {
                        C1_lo = (C1_lo.wrapping_add(C2_lo));
                        C1_hi = (C1_hi.wrapping_add(C2_hi));
                        if (C1_lo < C1.w[0]) {
                            C1_hi = C1_hi.wrapping_add(1);
                        }
                        if ((C1_hi > 0x0001ed09bead87c0) || (((C1_hi == 0x0001ed09bead87c0) && (C1_lo >= 0x378d8e6400000000)))) {
                            if (C1_lo >= 0xfffffffffffffffb) {
                                C1_lo = (C1_lo.wrapping_add(5));
                                C1_hi = (C1_hi.wrapping_add(1));
                            } else {
                                C1_lo = (C1_lo.wrapping_add(5));
                            }
                            C1.w[1] = C1_hi;
                            C1.w[0] = C1_lo;
                            ten2m1.w[1] = 0x1999999999999999;
                            ten2m1.w[0] = 0x9999999999999a00;
                            P256 = __mul_128x128_to_256(C1, ten2m1);
                            if ((((P256.w[1] != 0) || (P256.w[0] != 0))) && (((P256.w[1] < 0x1999999999999999) || (((P256.w[1] == 0x1999999999999999) && (P256.w[0] <= 0x9999999999999999)))))) {
                                if ((P256.w[2] & 0x01) != 0) {
                                    is_midpoint_gt_even = 1;
                                    P256.w[2] = P256.w[2].wrapping_sub(1);
                                    if (P256.w[2] == 0xffffffffffffffff) {
                                        P256.w[3] = P256.w[3].wrapping_sub(1);
                                    }
                                } else {
                                    is_midpoint_lt_even = 1;
                                }
                            }
                            y_exp = (y_exp.wrapping_add(0x2000000000000));
                            if ((y_exp == 0x6000000000000000) && (((rnd_mode == 0) || (rnd_mode == 4)))) {
                                res.w[1] = (x_sign | 0x7800000000000000);
                                res.w[0] = 0x0;
                                (*pfpsf) |= 32;
                                (*pfpsf) |= 8;
                                return res;
                            }
                            if ((P256.w[1] > 0x8000000000000000) || (((P256.w[1] == 0x8000000000000000) && (P256.w[0] > 0x0)))) {
                                tmp64 = (P256.w[1].wrapping_sub(0x8000000000000000));
                                if ((tmp64 > 0x1999999999999999) || (((tmp64 == 0x1999999999999999) && (P256.w[0] >= 0x9999999999999999)))) {
                                    (*pfpsf) |= 32;
                                    is_inexact = 1;
                                }
                            } else {
                                (*pfpsf) |= 32;
                                is_inexact = 1;
                            }
                            C1_hi = P256.w[3];
                            C1_lo = P256.w[2];
                            if ((is_midpoint_gt_even == 0) && (is_midpoint_lt_even == 0)) {
                                is_inexact_lt_midpoint = bool_to_int(((is_inexact != 0) && ((P256.w[1] & 0x8000000000000000) != 0)));
                                is_inexact_gt_midpoint = bool_to_int(((is_inexact != 0) && ((P256.w[1] & 0x8000000000000000) == 0)));
                            }
                            if (rnd_mode != 0) {
                                if ((((x_sign == 0) && (((((rnd_mode == 2) && (is_inexact_lt_midpoint != 0))) || (((((rnd_mode == 4) || (rnd_mode == 2))) && (is_midpoint_gt_even != 0))))))) || (((x_sign != 0) && (((((rnd_mode == 1) && (is_inexact_lt_midpoint != 0))) || (((((rnd_mode == 4) || (rnd_mode == 1))) && (is_midpoint_gt_even != 0)))))))) {
                                    C1_lo = (C1_lo.wrapping_add(1));
                                    if (C1_lo == 0) {
                                        C1_hi = (C1_hi.wrapping_add(1));
                                    }
                                    if ((C1_hi == 0x0001ed09bead87c0) && (C1_lo == 0x378d8e6400000000)) {
                                        C1_hi = 0x0000314dc6448d93;
                                        C1_lo = 0x38c15b0a00000000;
                                        y_exp = (y_exp.wrapping_add(0x2000000000000));
                                    }
                                } else if ((((is_midpoint_lt_even != 0) || (is_inexact_gt_midpoint != 0))) && (((((x_sign != 0) && (((rnd_mode == 2) || (rnd_mode == 3))))) || (((x_sign == 0) && (((rnd_mode == 1) || (rnd_mode == 3)))))))) {
                                    C1_lo = (C1_lo.wrapping_sub(1));
                                    if (C1_lo == 0xffffffffffffffff) {
                                        C1_hi = C1_hi.wrapping_sub(1);
                                    }
                                    if ((C1_hi == 0x0000314dc6448d93) && (C1_lo == 0x38c15b09ffffffff)) {
                                        C1_hi = 0x0001ed09bead87c0;
                                        C1_lo = 0x378d8e63ffffffff;
                                        y_exp = (y_exp.wrapping_sub(0x2000000000000));
                                    }
                                } else {
                                }
                                if (y_exp == 0x6000000000000000) {
                                    if ((((rnd_mode == 1) && (x_sign != 0))) || (((rnd_mode == 2) && (x_sign == 0)))) {
                                        C1_hi = 0x7800000000000000;
                                        C1_lo = 0x0;
                                    } else {
                                        C1_hi = 0x5fffed09bead87c0;
                                        C1_lo = 0x378d8e63ffffffff;
                                    }
                                    y_exp = 0;
                                    (*pfpsf) |= 32;
                                    (*pfpsf) |= 8;
                                }
                            }
                        }
                    } else {
                        C1_lo = (C1_lo.wrapping_sub(C2_lo));
                        C1_hi = (C1_hi.wrapping_sub(C2_hi));
                        if (C1_lo > C1.w[0]) {
                            C1_hi = C1_hi.wrapping_sub(1);
                        }
                        if ((C1_lo == 0) && (C1_hi == 0)) {
                            if (x_exp < y_exp) {
                                res.w[1] = x_exp;
                            } else {
                                res.w[1] = y_exp;
                            }
                            res.w[0] = 0;
                            if (rnd_mode == 1) {
                                res.w[1] |= 0x8000000000000000;
                            }
                            return res;
                        }
                        if (C1_hi >= 0x8000000000000000) {
                            C1_lo = (!C1_lo);
                            C1_lo = C1_lo.wrapping_add(1);
                            C1_hi = (!C1_hi);
                            if (C1_lo == 0x0) {
                                C1_hi = C1_hi.wrapping_add(1);
                            }
                            x_sign = y_sign;
                        }
                    }
                    res.w[1] = ((x_sign | y_exp) | C1_hi);
                    res.w[0] = C1_lo;
                } else {
                    x1 = ((delta.wrapping_add(q2)).wrapping_sub(34));
                    'roundC2: loop {
                    scale = (((delta.wrapping_sub(q1)).wrapping_add(q2)).wrapping_sub(x1));
                    if (scale >= 20) {
                        C1 = __mul_128x64_to_128(C1_lo, bid_ten2k128[(scale.wrapping_sub(20)) as usize]);
                    } else if (scale >= 1) {
                        if (q1 <= 19) {
                            C1 = __mul_64x64_to_128(C1_lo, bid_ten2k64[scale as usize]);
                        } else {
                            C1.w[1] = C1_hi;
                            C1.w[0] = C1_lo;
                            C1 = __mul_128x64_to_128(bid_ten2k64[scale as usize], C1);
                        }
                    } else {
                        C1.w[1] = C1_hi;
                        C1.w[0] = C1_lo;
                    }
                    tmp64 = C1.w[0];
                    ind = (x1.wrapping_sub(1));
                    if (ind >= 0) {
                        C2.w[0] = C2_lo;
                        C2.w[1] = C2_hi;
                        if (ind <= 18) {
                            C2.w[0] = (C2.w[0].wrapping_add(bid_midpoint64[ind as usize]));
                            if (C2.w[0] < C2_lo) {
                                C2.w[1] = C2.w[1].wrapping_add(1);
                            }
                        } else {
                            C2.w[0] = (C2.w[0].wrapping_add(bid_midpoint128[(ind.wrapping_sub(19)) as usize].w[0]));
                            C2.w[1] = (C2.w[1].wrapping_add(bid_midpoint128[(ind.wrapping_sub(19)) as usize].w[1]));
                            if (C2.w[0] < C2_lo) {
                                C2.w[1] = C2.w[1].wrapping_add(1);
                            }
                        }
                        R256 = __mul_128x128_to_256(C2, bid_ten2mk128[ind as usize]);
                        if (ind <= 2) {
                            highf2star.w[1] = 0x0;
                            highf2star.w[0] = 0x0;
                        } else if (ind <= 21) {
                            highf2star.w[1] = 0x0;
                            highf2star.w[0] = (R256.w[2] & bid_maskhigh128[ind as usize]);
                        } else {
                            highf2star.w[1] = (R256.w[3] & bid_maskhigh128[ind as usize]);
                            highf2star.w[0] = R256.w[2];
                        }
                        if (ind >= 3) {
                            shift = (bid_shiftright128[ind as usize] as i64);
                            if (shift < 64) {
                                R256.w[2] = (((go_checked_shr_u64(R256.w[2], go_shift_count_u64((shift as u64) as u64)))) | ((go_checked_shl_u64(R256.w[3], go_shift_count_u64(((((64 as i64).wrapping_sub(shift)) as u64)) as u64)))));
                                R256.w[3] = ((go_checked_shr_u64(R256.w[3], go_shift_count_u64((shift as u64) as u64))));
                            } else {
                                R256.w[2] = ((go_checked_shr_u64(R256.w[3], go_shift_count_u64((((shift.wrapping_sub(64)) as u64)) as u64))));
                                R256.w[3] = 0x0;
                            }
                        }
                        if (second_pass != 0) {
                            is_inexact_lt_midpoint = 0;
                            is_inexact_gt_midpoint = 0;
                            is_midpoint_lt_even = 0;
                            is_midpoint_gt_even = 0;
                        }
                        if (ind <= 2) {
                            if ((R256.w[1] > 0x8000000000000000) || (((R256.w[1] == 0x8000000000000000) && (R256.w[0] > 0x0)))) {
                                tmp64A = (R256.w[1].wrapping_sub(0x8000000000000000));
                                if ((tmp64A > bid_ten2mk128trunc[ind as usize].w[1]) || (((tmp64A == bid_ten2mk128trunc[ind as usize].w[1]) && (R256.w[0] >= bid_ten2mk128trunc[ind as usize].w[0])))) {
                                    tmp_inexact = 1;
                                    if (x_sign == y_sign) {
                                        is_inexact_lt_midpoint = 1;
                                    } else {
                                        is_inexact_gt_midpoint = 1;
                                    }
                                }
                            } else {
                                tmp_inexact = 1;
                                if (x_sign == y_sign) {
                                    is_inexact_gt_midpoint = 1;
                                } else {
                                    is_inexact_lt_midpoint = 1;
                                }
                            }
                        } else if (ind <= 21) {
                            if (((highf2star.w[1] > 0x0) || (((highf2star.w[1] == 0x0) && (highf2star.w[0] > bid_onehalf128[ind as usize])))) || ((((highf2star.w[1] == 0x0) && (highf2star.w[0] == bid_onehalf128[ind as usize])) && (((R256.w[1] != 0) || (R256.w[0] != 0)))))) {
                                tmp64A = (highf2star.w[0].wrapping_sub(bid_onehalf128[ind as usize]));
                                tmp64B = highf2star.w[1];
                                if (tmp64A > highf2star.w[0]) {
                                    tmp64B = tmp64B.wrapping_sub(1);
                                }
                                if ((((tmp64B != 0) || (tmp64A != 0)) || (R256.w[1] > bid_ten2mk128trunc[ind as usize].w[1])) || (((R256.w[1] == bid_ten2mk128trunc[ind as usize].w[1]) && (R256.w[0] > bid_ten2mk128trunc[ind as usize].w[0])))) {
                                    tmp_inexact = 1;
                                    if (x_sign == y_sign) {
                                        is_inexact_lt_midpoint = 1;
                                    } else {
                                        is_inexact_gt_midpoint = 1;
                                    }
                                }
                            } else {
                                tmp_inexact = 1;
                                if (x_sign == y_sign) {
                                    is_inexact_gt_midpoint = 1;
                                } else {
                                    is_inexact_lt_midpoint = 1;
                                }
                            }
                        } else {
                            if ((highf2star.w[1] > bid_onehalf128[ind as usize]) || (((highf2star.w[1] == bid_onehalf128[ind as usize]) && ((((highf2star.w[0] != 0) || (R256.w[1] != 0)) || (R256.w[0] != 0)))))) {
                                tmp64B = (highf2star.w[1].wrapping_sub(bid_onehalf128[ind as usize]));
                                if ((((tmp64B != 0) || (highf2star.w[0] != 0)) || (R256.w[1] > bid_ten2mk128trunc[ind as usize].w[1])) || (((R256.w[1] == bid_ten2mk128trunc[ind as usize].w[1]) && (R256.w[0] > bid_ten2mk128trunc[ind as usize].w[0])))) {
                                    tmp_inexact = 1;
                                    if (x_sign == y_sign) {
                                        is_inexact_lt_midpoint = 1;
                                    } else {
                                        is_inexact_gt_midpoint = 1;
                                    }
                                }
                            } else {
                                tmp_inexact = 1;
                                if (x_sign == y_sign) {
                                    is_inexact_gt_midpoint = 1;
                                } else {
                                    is_inexact_lt_midpoint = 1;
                                }
                            }
                        }
                        if ((((((R256.w[1] != 0) || (R256.w[0] != 0))) && (highf2star.w[1] == 0)) && (highf2star.w[0] == 0)) && (((R256.w[1] < bid_ten2mk128trunc[ind as usize].w[1]) || (((R256.w[1] == bid_ten2mk128trunc[ind as usize].w[1]) && (R256.w[0] <= bid_ten2mk128trunc[ind as usize].w[0])))))) {
                            if ((((tmp64.wrapping_add(R256.w[2]))) & 0x01) != 0) {
                                R256.w[2] = R256.w[2].wrapping_sub(1);
                                if (R256.w[2] == 0xffffffffffffffff) {
                                    R256.w[3] = R256.w[3].wrapping_sub(1);
                                }
                                if (x_sign == y_sign) {
                                    is_midpoint_gt_even = 1;
                                } else {
                                    is_midpoint_lt_even = 1;
                                }
                                is_inexact_lt_midpoint = 0;
                                is_inexact_gt_midpoint = 0;
                            } else {
                                if (x_sign == y_sign) {
                                    is_midpoint_lt_even = 1;
                                } else {
                                    is_midpoint_gt_even = 1;
                                }
                                is_inexact_lt_midpoint = 0;
                                is_inexact_gt_midpoint = 0;
                            }
                        }
                    } else {
                        R256.w[2] = C2_lo;
                        R256.w[3] = C2_hi;
                        tmp_inexact = 0;
                        if (second_pass != 0) {
                            is_midpoint_lt_even = 0;
                            is_midpoint_gt_even = 0;
                            is_inexact_lt_midpoint = 0;
                            is_inexact_gt_midpoint = 0;
                        }
                    }
                    if (x_sign == y_sign) {
                        C1.w[0] = (C1.w[0].wrapping_add(R256.w[2]));
                        C1.w[1] = (C1.w[1].wrapping_add(R256.w[3]));
                        if (C1.w[0] < tmp64) {
                            C1.w[1] = C1.w[1].wrapping_add(1);
                        }
                        if ((C1.w[1] > 0x0001ed09bead87c0) || (((C1.w[1] == 0x0001ed09bead87c0) && (C1.w[0] >= 0x378d8e6400000000)))) {
                            if (C1.w[0] >= 0xfffffffffffffffb) {
                                C1.w[0] = (C1.w[0].wrapping_add(5));
                                C1.w[1] = (C1.w[1].wrapping_add(1));
                            } else {
                                C1.w[0] = (C1.w[0].wrapping_add(5));
                            }
                            Q256 = __mul_128x128_to_256(C1, bid_ten2mk128[0]);
                            if ((((Q256.w[1] != 0) || (Q256.w[0] != 0))) && (((Q256.w[1] < bid_ten2mk128trunc[0].w[1]) || (((Q256.w[1] == bid_ten2mk128trunc[0].w[1]) && (Q256.w[0] <= bid_ten2mk128trunc[0].w[0])))))) {
                                if (is_inexact_lt_midpoint != 0) {
                                    is_inexact_gt_midpoint = 1;
                                    is_inexact_lt_midpoint = 0;
                                    is_midpoint_gt_even = 0;
                                    is_midpoint_lt_even = 0;
                                } else if (is_inexact_gt_midpoint != 0) {
                                    Q256.w[2] = Q256.w[2].wrapping_sub(1);
                                    if (Q256.w[2] == 0xffffffffffffffff) {
                                        Q256.w[3] = Q256.w[3].wrapping_sub(1);
                                    }
                                    is_inexact_gt_midpoint = 0;
                                    is_inexact_lt_midpoint = 1;
                                    is_midpoint_gt_even = 0;
                                    is_midpoint_lt_even = 0;
                                } else if (is_midpoint_gt_even != 0) {
                                    is_inexact_gt_midpoint = 0;
                                    is_inexact_lt_midpoint = 1;
                                    is_midpoint_gt_even = 0;
                                    is_midpoint_lt_even = 0;
                                } else {
                                    if ((Q256.w[2] & 0x01) != 0) {
                                        Q256.w[2] = Q256.w[2].wrapping_sub(1);
                                        if (Q256.w[2] == 0xffffffffffffffff) {
                                            Q256.w[3] = Q256.w[3].wrapping_sub(1);
                                        }
                                        is_inexact_gt_midpoint = 0;
                                        is_inexact_lt_midpoint = 0;
                                        is_midpoint_gt_even = 1;
                                        is_midpoint_lt_even = 0;
                                    } else {
                                        is_inexact_gt_midpoint = 0;
                                        is_inexact_lt_midpoint = 0;
                                        is_midpoint_gt_even = 0;
                                        is_midpoint_lt_even = 1;
                                    }
                                }
                                tmp_inexact = 1;
                            } else {
                                if ((Q256.w[1] > 0x8000000000000000) || (((Q256.w[1] == 0x8000000000000000) && (Q256.w[0] > 0x0)))) {
                                    Q256.w[1] = (Q256.w[1].wrapping_sub(0x8000000000000000));
                                    if ((Q256.w[1] > bid_ten2mk128trunc[0].w[1]) || (((Q256.w[1] == bid_ten2mk128trunc[0].w[1]) && (Q256.w[0] > bid_ten2mk128trunc[0].w[0])))) {
                                        is_inexact_gt_midpoint = 0;
                                        is_inexact_lt_midpoint = 1;
                                        is_midpoint_gt_even = 0;
                                        is_midpoint_lt_even = 0;
                                        tmp_inexact = 1;
                                    } else {
                                        if (tmp_inexact != 0) {
                                            if (is_midpoint_lt_even != 0) {
                                                is_inexact_gt_midpoint = 1;
                                                is_midpoint_lt_even = 0;
                                            } else if (is_midpoint_gt_even != 0) {
                                                is_inexact_lt_midpoint = 1;
                                                is_midpoint_gt_even = 0;
                                            } else {
                                            }
                                        }
                                    }
                                } else {
                                    is_inexact_gt_midpoint = 1;
                                    is_inexact_lt_midpoint = 0;
                                    is_midpoint_gt_even = 0;
                                    is_midpoint_lt_even = 0;
                                    tmp_inexact = 1;
                                }
                            }
                            C1.w[1] = Q256.w[3];
                            C1.w[0] = Q256.w[2];
                            y_exp = (y_exp.wrapping_add(((go_checked_shl_u64(((x1.wrapping_add(1)) as u64), go_shift_count_u64((49) as u64))))));
                        } else {
                            y_exp = (y_exp.wrapping_add(((go_checked_shl_u64((x1 as u64), go_shift_count_u64((49) as u64))))));
                        }
                        if ((y_exp == 0x6000000000000000) && (((rnd_mode == 0) || (rnd_mode == 4)))) {
                            res.w[1] = (0x7800000000000000 | x_sign);
                            res.w[0] = 0x0;
                            (*pfpsf) |= 32;
                            (*pfpsf) |= 8;
                            return res;
                        }
                    } else {
                        C1.w[0] = (C1.w[0].wrapping_sub(R256.w[2]));
                        C1.w[1] = (C1.w[1].wrapping_sub(R256.w[3]));
                        if (C1.w[0] > tmp64) {
                            C1.w[1] = C1.w[1].wrapping_sub(1);
                        }
                        if (C1.w[1] >= 0x8000000000000000) {
                            C1.w[0] = (!C1.w[0]);
                            C1.w[0] = C1.w[0].wrapping_add(1);
                            C1.w[1] = (!C1.w[1]);
                            if (C1.w[0] == 0x0) {
                                C1.w[1] = C1.w[1].wrapping_add(1);
                            }
                            tmp_sign = y_sign;
                        } else {
                            tmp_sign = x_sign;
                        }
                        if ((((C1.w[1] < 0x0000314dc6448d93) || (((C1.w[1] == 0x0000314dc6448d93) && (C1.w[0] < 0x38c15b0a00000000))))) || ((((C1.w[1] == 0x0000314dc6448d93) && (C1.w[0] == 0x38c15b0a00000000)) && (((is_inexact_gt_midpoint != 0) || (is_midpoint_lt_even != 0)))))) {
                            x1 = (x1.wrapping_sub(1));
                            if (x1 >= 0) {
                                is_midpoint_lt_even = 0;
                                is_midpoint_gt_even = 0;
                                is_inexact_lt_midpoint = 0;
                                is_inexact_gt_midpoint = 0;
                                tmp_inexact = 0;
                                second_pass = 1;
                                continue 'roundC2;
                            }
                        }
                        if ((C1.w[1] == 0x0001ed09bead87c0) && (C1.w[0] == 0x378d8e6400000000)) {
                            C1.w[1] = 0x0000314dc6448d93;
                            C1.w[0] = 0x38c15b0a00000000;
                            y_exp = (y_exp.wrapping_add((((1 as u64) << 49))));
                        }
                        x_sign = tmp_sign;
                        if (x1 >= 1) {
                            y_exp = (y_exp.wrapping_add(((go_checked_shl_u64((x1 as u64), go_shift_count_u64((49) as u64))))));
                        }
                    }
                    break;
                    }
                    C1_hi = C1.w[1];
                    C1_lo = C1.w[0];
                    if (rnd_mode != 0) {
                        if ((((x_sign == 0) && (((((rnd_mode == 2) && (is_inexact_lt_midpoint != 0))) || (((((rnd_mode == 4) || (rnd_mode == 2))) && (is_midpoint_gt_even != 0))))))) || (((x_sign != 0) && (((((rnd_mode == 1) && (is_inexact_lt_midpoint != 0))) || (((((rnd_mode == 4) || (rnd_mode == 1))) && (is_midpoint_gt_even != 0)))))))) {
                            C1_lo = (C1_lo.wrapping_add(1));
                            if (C1_lo == 0) {
                                C1_hi = (C1_hi.wrapping_add(1));
                            }
                            if ((C1_hi == 0x0001ed09bead87c0) && (C1_lo == 0x378d8e6400000000)) {
                                C1_hi = 0x0000314dc6448d93;
                                C1_lo = 0x38c15b0a00000000;
                                y_exp = (y_exp.wrapping_add(0x2000000000000));
                            }
                        } else if ((((is_midpoint_lt_even != 0) || (is_inexact_gt_midpoint != 0))) && (((((x_sign != 0) && (((rnd_mode == 2) || (rnd_mode == 3))))) || (((x_sign == 0) && (((rnd_mode == 1) || (rnd_mode == 3)))))))) {
                            C1_lo = (C1_lo.wrapping_sub(1));
                            if (C1_lo == 0xffffffffffffffff) {
                                C1_hi = C1_hi.wrapping_sub(1);
                            }
                            if ((C1_hi == 0x0000314dc6448d93) && (C1_lo == 0x38c15b09ffffffff)) {
                                C1_hi = 0x0001ed09bead87c0;
                                C1_lo = 0x378d8e63ffffffff;
                                y_exp = (y_exp.wrapping_sub(0x2000000000000));
                            }
                        } else {
                        }
                        if (y_exp == 0x6000000000000000) {
                            if ((((rnd_mode == 1) && (x_sign != 0))) || (((rnd_mode == 2) && (x_sign == 0)))) {
                                C1_hi = 0x7800000000000000;
                                C1_lo = 0x0;
                            } else {
                                C1_hi = 0x5fffed09bead87c0;
                                C1_lo = 0x378d8e63ffffffff;
                            }
                            y_exp = 0;
                            (*pfpsf) |= 32;
                            (*pfpsf) |= 8;
                        }
                    }
                    res.w[1] = ((x_sign | y_exp) | C1_hi);
                    res.w[0] = C1_lo;
                    if (tmp_inexact != 0) {
                        (*pfpsf) |= 32;
                    }
                }
            } else {
                scale = ((delta.wrapping_sub(q1)).wrapping_add(q2));
                if (scale >= 20) {
                    C1 = __mul_128x64_to_128(C1_lo, bid_ten2k128[(scale.wrapping_sub(20)) as usize]);
                } else if (scale >= 1) {
                    if (q1 <= 19) {
                        C1 = __mul_64x64_to_128(C1_lo, bid_ten2k64[scale as usize]);
                    } else {
                        C1.w[1] = C1_hi;
                        C1.w[0] = C1_lo;
                        C1 = __mul_128x64_to_128(bid_ten2k64[scale as usize], C1);
                    }
                } else {
                    C1.w[1] = C1_hi;
                    C1.w[0] = C1_lo;
                }
                C1_hi = C1.w[1];
                C1_lo = C1.w[0];
                if (x_sign == y_sign) {
                    C1_lo = (C1_lo.wrapping_add(C2_lo));
                    C1_hi = (C1_hi.wrapping_add(C2_hi));
                    if (C1_lo < C1.w[0]) {
                        C1_hi = C1_hi.wrapping_add(1);
                    }
                    if ((C1_hi > 0x0001ed09bead87c0) || (((C1_hi == 0x0001ed09bead87c0) && (C1_lo >= 0x378d8e6400000000)))) {
                        if (C1_lo >= 0xfffffffffffffffb) {
                            C1_lo = (C1_lo.wrapping_add(5));
                            C1_hi = (C1_hi.wrapping_add(1));
                        } else {
                            C1_lo = (C1_lo.wrapping_add(5));
                        }
                        C1.w[1] = C1_hi;
                        C1.w[0] = C1_lo;
                        ten2m1.w[1] = 0x1999999999999999;
                        ten2m1.w[0] = 0x9999999999999a00;
                        P256 = __mul_128x128_to_256(C1, ten2m1);
                        if ((((P256.w[1] != 0) || (P256.w[0] != 0))) && (((P256.w[1] < 0x1999999999999999) || (((P256.w[1] == 0x1999999999999999) && (P256.w[0] <= 0x9999999999999999)))))) {
                            if ((P256.w[2] & 0x01) != 0) {
                                is_midpoint_gt_even = 1;
                                P256.w[2] = P256.w[2].wrapping_sub(1);
                                if (P256.w[2] == 0xffffffffffffffff) {
                                    P256.w[3] = P256.w[3].wrapping_sub(1);
                                }
                            } else {
                                is_midpoint_lt_even = 1;
                            }
                        }
                        y_exp = (y_exp.wrapping_add(0x2000000000000));
                        if ((y_exp == 0x6000000000000000) && (((rnd_mode == 0) || (rnd_mode == 4)))) {
                            res.w[1] = (x_sign | 0x7800000000000000);
                            res.w[0] = 0x0;
                            (*pfpsf) |= 32;
                            (*pfpsf) |= 8;
                            return res;
                        }
                        if ((P256.w[1] > 0x8000000000000000) || (((P256.w[1] == 0x8000000000000000) && (P256.w[0] > 0x0)))) {
                            tmp64 = (P256.w[1].wrapping_sub(0x8000000000000000));
                            if ((tmp64 > 0x1999999999999999) || (((tmp64 == 0x1999999999999999) && (P256.w[0] >= 0x9999999999999999)))) {
                                (*pfpsf) |= 32;
                                is_inexact = 1;
                            }
                        } else {
                            (*pfpsf) |= 32;
                            is_inexact = 1;
                        }
                        C1_hi = P256.w[3];
                        C1_lo = P256.w[2];
                        if ((is_midpoint_gt_even == 0) && (is_midpoint_lt_even == 0)) {
                            is_inexact_lt_midpoint = bool_to_int(((is_inexact != 0) && ((P256.w[1] & 0x8000000000000000) != 0)));
                            is_inexact_gt_midpoint = bool_to_int(((is_inexact != 0) && ((P256.w[1] & 0x8000000000000000) == 0)));
                        }
                        if (rnd_mode != 0) {
                            if ((((x_sign == 0) && (((((rnd_mode == 2) && (is_inexact_lt_midpoint != 0))) || (((((rnd_mode == 4) || (rnd_mode == 2))) && (is_midpoint_gt_even != 0))))))) || (((x_sign != 0) && (((((rnd_mode == 1) && (is_inexact_lt_midpoint != 0))) || (((((rnd_mode == 4) || (rnd_mode == 1))) && (is_midpoint_gt_even != 0)))))))) {
                                C1_lo = (C1_lo.wrapping_add(1));
                                if (C1_lo == 0) {
                                    C1_hi = (C1_hi.wrapping_add(1));
                                }
                                if ((C1_hi == 0x0001ed09bead87c0) && (C1_lo == 0x378d8e6400000000)) {
                                    C1_hi = 0x0000314dc6448d93;
                                    C1_lo = 0x38c15b0a00000000;
                                    y_exp = (y_exp.wrapping_add(0x2000000000000));
                                }
                            } else if ((((is_midpoint_lt_even != 0) || (is_inexact_gt_midpoint != 0))) && (((((x_sign != 0) && (((rnd_mode == 2) || (rnd_mode == 3))))) || (((x_sign == 0) && (((rnd_mode == 1) || (rnd_mode == 3)))))))) {
                                C1_lo = (C1_lo.wrapping_sub(1));
                                if (C1_lo == 0xffffffffffffffff) {
                                    C1_hi = C1_hi.wrapping_sub(1);
                                }
                                if ((C1_hi == 0x0000314dc6448d93) && (C1_lo == 0x38c15b09ffffffff)) {
                                    C1_hi = 0x0001ed09bead87c0;
                                    C1_lo = 0x378d8e63ffffffff;
                                    y_exp = (y_exp.wrapping_sub(0x2000000000000));
                                }
                            } else {
                            }
                            if (y_exp == 0x6000000000000000) {
                                if ((((rnd_mode == 1) && (x_sign != 0))) || (((rnd_mode == 2) && (x_sign == 0)))) {
                                    C1_hi = 0x7800000000000000;
                                    C1_lo = 0x0;
                                } else {
                                    C1_hi = 0x5fffed09bead87c0;
                                    C1_lo = 0x378d8e63ffffffff;
                                }
                                y_exp = 0;
                                (*pfpsf) |= 32;
                                (*pfpsf) |= 8;
                            }
                        }
                    }
                    res.w[1] = ((x_sign | y_exp) | C1_hi);
                    res.w[0] = C1_lo;
                } else {
                    C1_lo = (C2_lo.wrapping_sub(C1_lo));
                    C1_hi = (C2_hi.wrapping_sub(C1_hi));
                    if (C1_lo > C2_lo) {
                        C1_hi = C1_hi.wrapping_sub(1);
                    }
                    if (C1_hi >= 0x8000000000000000) {
                        C1_lo = (!C1_lo);
                        C1_lo = C1_lo.wrapping_add(1);
                        C1_hi = (!C1_hi);
                        if (C1_lo == 0x0) {
                            C1_hi = C1_hi.wrapping_add(1);
                        }
                        x_sign = y_sign;
                    }
                    if ((C1_lo == 0) && (C1_hi == 0)) {
                        if (x_exp < y_exp) {
                            res.w[1] = x_exp;
                        } else {
                            res.w[1] = y_exp;
                        }
                        res.w[0] = 0;
                        if (rnd_mode == 1) {
                            res.w[1] |= 0x8000000000000000;
                        }
                        return res;
                    }
                    res.w[1] = ((y_sign | y_exp) | C1_hi);
                    res.w[0] = C1_lo;
                }
            }
        }
        return res;
    }
}

pub(crate) fn bool_to_int(mut b: bool) -> i64 {
    if b {
        return 1;
    }
    return 0;
}

pub fn bid128_sub(mut x: BID_UINT128, mut y: BID_UINT128, mut rnd_mode: i64, pfpsf: &mut u32) -> BID_UINT128 {
    let mut y_sign: u64 = 0;
    if ((y.w[1] & 0x7c00000000000000) != 0x7c00000000000000) {
        y_sign = (y.w[1] & 0x8000000000000000);
        if (y_sign != 0) {
            y.w[1] = (y.w[1] & 0x7fffffffffffffff);
        } else {
            y.w[1] = (y.w[1] | 0x8000000000000000);
        }
    }
    return bid128_add(x, y, rnd_mode, pfpsf);
}
