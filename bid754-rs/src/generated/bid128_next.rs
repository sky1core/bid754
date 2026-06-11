// Auto-generated from bid128_next.go by go2rs. Do not edit.

use super::prelude::*;

pub fn bid128_next_up(mut x: BID_UINT128) -> (BID_UINT128, u32) {
    let mut res: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut x_sign: u64 = 0;
    let mut x_exp: u64 = 0;
    let mut exp: i64 = 0;
    let mut x_nr_bits: i64 = 0;
    let mut q1: i64 = 0;
    let mut ind: i64 = 0;
    let mut C1: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut pfpsf: u32 = 0;
    x_sign = (x.w[1] & 0x8000000000000000);
    C1.w[1] = (x.w[1] & 0x1ffffffffffff);
    C1.w[0] = x.w[0];
    if ((x.w[1] & 0x7800000000000000) == 0x7800000000000000) {
        if ((x.w[1] & 0x7c00000000000000) == 0x7c00000000000000) {
            if ((((x.w[1] & 0x00003fffffffffff) > 0x0000314dc6448d93)) || (((((x.w[1] & 0x00003fffffffffff) == 0x0000314dc6448d93)) && (x.w[0] > 0x38c15b09ffffffff)))) {
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
        } else {
            if (x_sign == 0) {
                res.w[1] = 0x7800000000000000;
                res.w[0] = 0x0000000000000000;
            } else {
                res.w[1] = 0xdfffed09bead87c0;
                res.w[0] = 0x378d8e63ffffffff;
            }
        }
        return (res, pfpsf);
    }
    if ((x.w[1] & 0x6000000000000000) == 0x6000000000000000) {
        x_exp = (((go_checked_shl_u64(x.w[1], go_shift_count_u64((2) as u64)))) & 0x7ffe000000000000);
        C1.w[1] = 0;
        C1.w[0] = 0;
    } else {
        x_exp = (x.w[1] & 0x7ffe000000000000);
        if ((C1.w[1] > 0x0001ed09bead87c0) || (((C1.w[1] == 0x0001ed09bead87c0) && (C1.w[0] > 0x378d8e63ffffffff)))) {
            C1.w[1] = 0;
            C1.w[0] = 0;
        } else {
        }
    }
    if ((C1.w[1] == 0x0) && (C1.w[0] == 0x0)) {
        res.w[1] = 0x0000000000000000;
        res.w[0] = 0x0000000000000001;
    } else {
        if ((x.w[1] == 0x5fffed09bead87c0) && (x.w[0] == 0x378d8e63ffffffff)) {
            res.w[1] = 0x7800000000000000;
            res.w[0] = 0x0000000000000000;
        } else if ((x.w[1] == 0x8000000000000000) && (x.w[0] == 0x0000000000000001)) {
            res.w[1] = 0x8000000000000000;
            res.w[0] = 0x0000000000000000;
        } else {
            if (C1.w[1] == 0) {
                if (C1.w[0] >= 0x0020000000000000) {
                    if (C1.w[0] >= 0x0000000100000000) {
                        let mut tmp1 = (((go_checked_shr_u64(C1.w[0], go_shift_count_u64((32) as u64))) as f64)).to_bits();
                        x_nr_bits = ((33 as i64).wrapping_add((((((((go_checked_shr_u64(tmp1, go_shift_count_u64((52) as u64))) as u32)) & 0x7ff)).wrapping_sub(0x3ff)) as i64)));
                    } else {
                        let mut tmp1 = (C1.w[0] as f64).to_bits();
                        x_nr_bits = ((1 as i64).wrapping_add((((((((go_checked_shr_u64(tmp1, go_shift_count_u64((52) as u64))) as u32)) & 0x7ff)).wrapping_sub(0x3ff)) as i64)));
                    }
                } else {
                    let mut tmp1 = (C1.w[0] as f64).to_bits();
                    x_nr_bits = ((1 as i64).wrapping_add((((((((go_checked_shr_u64(tmp1, go_shift_count_u64((52) as u64))) as u32)) & 0x7ff)).wrapping_sub(0x3ff)) as i64)));
                }
            } else {
                let mut tmp1 = (C1.w[1] as f64).to_bits();
                x_nr_bits = ((65 as i64).wrapping_add((((((((go_checked_shr_u64(tmp1, go_shift_count_u64((52) as u64))) as u32)) & 0x7ff)).wrapping_sub(0x3ff)) as i64)));
            }
            q1 = (bid_nr_digits[(x_nr_bits.wrapping_sub(1)) as usize].digits as i64);
            if (q1 == 0) {
                q1 = (bid_nr_digits[(x_nr_bits.wrapping_sub(1)) as usize].digits1 as i64);
                if ((C1.w[1] > bid_nr_digits[(x_nr_bits.wrapping_sub(1)) as usize].threshold_hi) || (((C1.w[1] == bid_nr_digits[(x_nr_bits.wrapping_sub(1)) as usize].threshold_hi) && (C1.w[0] >= bid_nr_digits[(x_nr_bits.wrapping_sub(1)) as usize].threshold_lo)))) {
                    q1 = q1.wrapping_add(1);
                }
            }
            if (q1 < 34) {
                exp = (((go_checked_shr_u64(x_exp, go_shift_count_u64((49) as u64))) as i64).wrapping_sub(6176));
                if ((exp.wrapping_add(6176)) > ((34 as i64).wrapping_sub(q1))) {
                    ind = ((34 as i64).wrapping_sub(q1));
                    if (q1 <= 19) {
                        if (ind <= 19) {
                            C1 = __mul_64x64_to_128(C1.w[0], bid_ten2k64[ind as usize]);
                        } else {
                            C1 = __mul_128x64_to_128(C1.w[0], bid_ten2k128[(ind.wrapping_sub(20)) as usize]);
                        }
                    } else {
                        if (ind <= 14) {
                            C1 = __mul_128x64_to_128(bid_ten2k64[ind as usize], C1);
                        } else if (ind <= 19) {
                            C1 = __mul_64x64_to_128(C1.w[0], bid_ten2k64[ind as usize]);
                        } else {
                            C1 = __mul_128x64_to_128(C1.w[0], bid_ten2k128[(ind.wrapping_sub(20)) as usize]);
                        }
                    }
                    x_exp = (x_exp.wrapping_sub(((go_checked_shl_u64((ind as u64), go_shift_count_u64((49) as u64))))));
                } else {
                    ind = (exp.wrapping_add(6176));
                    if (ind <= 19) {
                        if (q1 <= 19) {
                            C1 = __mul_64x64_to_128(C1.w[0], bid_ten2k64[ind as usize]);
                        } else {
                            C1 = __mul_128x64_to_128(bid_ten2k64[ind as usize], C1);
                        }
                    } else {
                        C1 = __mul_128x64_to_128(C1.w[0], bid_ten2k128[(ind.wrapping_sub(20)) as usize]);
                    }
                    x_exp = 0;
                }
            }
            if (x_sign == 0) {
                C1.w[0] = C1.w[0].wrapping_add(1);
                if (C1.w[0] == 0) {
                    C1.w[1] = C1.w[1].wrapping_add(1);
                }
                if ((C1.w[1] == 0x0001ed09bead87c0) && (C1.w[0] == 0x378d8e6400000000)) {
                    C1.w[1] = 0x0000314dc6448d93;
                    C1.w[0] = 0x38c15b0a00000000;
                    x_exp = (x_exp.wrapping_add(0x2000000000000));
                }
            } else {
                C1.w[0] = C1.w[0].wrapping_sub(1);
                if (C1.w[0] == 0xffffffffffffffff) {
                    C1.w[1] = C1.w[1].wrapping_sub(1);
                }
                if (((x_exp != 0) && (C1.w[1] == 0x0000314dc6448d93)) && (C1.w[0] == 0x38c15b09ffffffff)) {
                    C1.w[1] = 0x0001ed09bead87c0;
                    C1.w[0] = 0x378d8e63ffffffff;
                    x_exp = (x_exp.wrapping_sub(0x2000000000000));
                }
            }
            res.w[1] = ((x_sign | x_exp) | C1.w[1]);
            res.w[0] = C1.w[0];
        }
    }
    return (res, pfpsf);
}

pub fn bid128_next_down(mut x: BID_UINT128) -> (BID_UINT128, u32) {
    let mut res: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut x_sign: u64 = 0;
    let mut x_exp: u64 = 0;
    let mut exp: i64 = 0;
    let mut x_nr_bits: i64 = 0;
    let mut q1: i64 = 0;
    let mut ind: i64 = 0;
    let mut C1: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut pfpsf: u32 = 0;
    x_sign = (x.w[1] & 0x8000000000000000);
    C1.w[1] = (x.w[1] & 0x1ffffffffffff);
    C1.w[0] = x.w[0];
    if ((x.w[1] & 0x7800000000000000) == 0x7800000000000000) {
        if ((x.w[1] & 0x7c00000000000000) == 0x7c00000000000000) {
            if ((((x.w[1] & 0x00003fffffffffff) > 0x0000314dc6448d93)) || (((((x.w[1] & 0x00003fffffffffff) == 0x0000314dc6448d93)) && (x.w[0] > 0x38c15b09ffffffff)))) {
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
        } else {
            if (x_sign == 0) {
                res.w[1] = 0x5fffed09bead87c0;
                res.w[0] = 0x378d8e63ffffffff;
            } else {
                res.w[1] = 0xf800000000000000;
                res.w[0] = 0x0000000000000000;
            }
        }
        return (res, pfpsf);
    }
    if ((x.w[1] & 0x6000000000000000) == 0x6000000000000000) {
        x_exp = (((go_checked_shl_u64(x.w[1], go_shift_count_u64((2) as u64)))) & 0x7ffe000000000000);
        C1.w[1] = 0;
        C1.w[0] = 0;
    } else {
        x_exp = (x.w[1] & 0x7ffe000000000000);
        if ((C1.w[1] > 0x0001ed09bead87c0) || (((C1.w[1] == 0x0001ed09bead87c0) && (C1.w[0] > 0x378d8e63ffffffff)))) {
            C1.w[1] = 0;
            C1.w[0] = 0;
        } else {
        }
    }
    if ((C1.w[1] == 0x0) && (C1.w[0] == 0x0)) {
        res.w[1] = 0x8000000000000000;
        res.w[0] = 0x0000000000000001;
    } else {
        if ((x.w[1] == 0xdfffed09bead87c0) && (x.w[0] == 0x378d8e63ffffffff)) {
            res.w[1] = 0xf800000000000000;
            res.w[0] = 0x0000000000000000;
        } else if ((x.w[1] == 0x0) && (x.w[0] == 0x0000000000000001)) {
            res.w[1] = 0x0000000000000000;
            res.w[0] = 0x0000000000000000;
        } else {
            if (C1.w[1] == 0) {
                if (C1.w[0] >= 0x0020000000000000) {
                    if (C1.w[0] >= 0x0000000100000000) {
                        let mut tmp1 = (((go_checked_shr_u64(C1.w[0], go_shift_count_u64((32) as u64))) as f64)).to_bits();
                        x_nr_bits = ((33 as i64).wrapping_add((((((((go_checked_shr_u64(tmp1, go_shift_count_u64((52) as u64))) as u32)) & 0x7ff)).wrapping_sub(0x3ff)) as i64)));
                    } else {
                        let mut tmp1 = (C1.w[0] as f64).to_bits();
                        x_nr_bits = ((1 as i64).wrapping_add((((((((go_checked_shr_u64(tmp1, go_shift_count_u64((52) as u64))) as u32)) & 0x7ff)).wrapping_sub(0x3ff)) as i64)));
                    }
                } else {
                    let mut tmp1 = (C1.w[0] as f64).to_bits();
                    x_nr_bits = ((1 as i64).wrapping_add((((((((go_checked_shr_u64(tmp1, go_shift_count_u64((52) as u64))) as u32)) & 0x7ff)).wrapping_sub(0x3ff)) as i64)));
                }
            } else {
                let mut tmp1 = (C1.w[1] as f64).to_bits();
                x_nr_bits = ((65 as i64).wrapping_add((((((((go_checked_shr_u64(tmp1, go_shift_count_u64((52) as u64))) as u32)) & 0x7ff)).wrapping_sub(0x3ff)) as i64)));
            }
            q1 = (bid_nr_digits[(x_nr_bits.wrapping_sub(1)) as usize].digits as i64);
            if (q1 == 0) {
                q1 = (bid_nr_digits[(x_nr_bits.wrapping_sub(1)) as usize].digits1 as i64);
                if ((C1.w[1] > bid_nr_digits[(x_nr_bits.wrapping_sub(1)) as usize].threshold_hi) || (((C1.w[1] == bid_nr_digits[(x_nr_bits.wrapping_sub(1)) as usize].threshold_hi) && (C1.w[0] >= bid_nr_digits[(x_nr_bits.wrapping_sub(1)) as usize].threshold_lo)))) {
                    q1 = q1.wrapping_add(1);
                }
            }
            if (q1 < 34) {
                exp = (((go_checked_shr_u64(x_exp, go_shift_count_u64((49) as u64))) as i64).wrapping_sub(6176));
                if ((exp.wrapping_add(6176)) > ((34 as i64).wrapping_sub(q1))) {
                    ind = ((34 as i64).wrapping_sub(q1));
                    if (q1 <= 19) {
                        if (ind <= 19) {
                            C1 = __mul_64x64_to_128(C1.w[0], bid_ten2k64[ind as usize]);
                        } else {
                            C1 = __mul_128x64_to_128(C1.w[0], bid_ten2k128[(ind.wrapping_sub(20)) as usize]);
                        }
                    } else {
                        if (ind <= 14) {
                            C1 = __mul_128x64_to_128(bid_ten2k64[ind as usize], C1);
                        } else if (ind <= 19) {
                            C1 = __mul_64x64_to_128(C1.w[0], bid_ten2k64[ind as usize]);
                        } else {
                            C1 = __mul_128x64_to_128(C1.w[0], bid_ten2k128[(ind.wrapping_sub(20)) as usize]);
                        }
                    }
                    x_exp = (x_exp.wrapping_sub(((go_checked_shl_u64((ind as u64), go_shift_count_u64((49) as u64))))));
                } else {
                    ind = (exp.wrapping_add(6176));
                    if (ind <= 19) {
                        if (q1 <= 19) {
                            C1 = __mul_64x64_to_128(C1.w[0], bid_ten2k64[ind as usize]);
                        } else {
                            C1 = __mul_128x64_to_128(bid_ten2k64[ind as usize], C1);
                        }
                    } else {
                        C1 = __mul_128x64_to_128(C1.w[0], bid_ten2k128[(ind.wrapping_sub(20)) as usize]);
                    }
                    x_exp = 0;
                }
            }
            if (x_sign != 0) {
                C1.w[0] = C1.w[0].wrapping_add(1);
                if (C1.w[0] == 0) {
                    C1.w[1] = C1.w[1].wrapping_add(1);
                }
                if ((C1.w[1] == 0x0001ed09bead87c0) && (C1.w[0] == 0x378d8e6400000000)) {
                    C1.w[1] = 0x0000314dc6448d93;
                    C1.w[0] = 0x38c15b0a00000000;
                    x_exp = (x_exp.wrapping_add(0x2000000000000));
                }
            } else {
                C1.w[0] = C1.w[0].wrapping_sub(1);
                if (C1.w[0] == 0xffffffffffffffff) {
                    C1.w[1] = C1.w[1].wrapping_sub(1);
                }
                if (((x_exp != 0) && (C1.w[1] == 0x0000314dc6448d93)) && (C1.w[0] == 0x38c15b09ffffffff)) {
                    C1.w[1] = 0x0001ed09bead87c0;
                    C1.w[0] = 0x378d8e63ffffffff;
                    x_exp = (x_exp.wrapping_sub(0x2000000000000));
                }
            }
            res.w[1] = ((x_sign | x_exp) | C1.w[1]);
            res.w[0] = C1.w[0];
        }
    }
    return (res, pfpsf);
}

pub fn bid128_next_after(mut x: BID_UINT128, mut y: BID_UINT128) -> (BID_UINT128, u32) {
    let mut xnswp: BID_UINT128 = x;
    let mut ynswp: BID_UINT128 = y;
    let mut res: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut tmp1: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut tmp2: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut tmp3: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut tmp_fpsf: u32 = 0;
    let mut res1: i64 = 0;
    let mut res2: i64 = 0;
    let mut x_exp: u64 = 0;
    let mut pfpsf: u32 = 0;
    _ = xnswp;
    _ = ynswp;
    if ((((x.w[1] & 0x7800000000000000) == 0x7800000000000000)) || (((y.w[1] & 0x7800000000000000) == 0x7800000000000000))) {
        if ((x.w[1] & 0x7c00000000000000) == 0x7c00000000000000) {
            if ((((x.w[1] & 0x00003fffffffffff) > 0x0000314dc6448d93)) || (((((x.w[1] & 0x00003fffffffffff) == 0x0000314dc6448d93)) && (x.w[0] > 0x38c15b09ffffffff)))) {
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
                if ((y.w[1] & 0x7e00000000000000) == 0x7e00000000000000) {
                    pfpsf |= 1;
                }
            }
            return (res, pfpsf);
        } else if ((y.w[1] & 0x7c00000000000000) == 0x7c00000000000000) {
            if ((((y.w[1] & 0x00003fffffffffff) > 0x0000314dc6448d93)) || (((((y.w[1] & 0x00003fffffffffff) == 0x0000314dc6448d93)) && (y.w[0] > 0x38c15b09ffffffff)))) {
                y.w[1] = (y.w[1] & 0xffffc00000000000);
                y.w[0] = 0x0;
            }
            if ((y.w[1] & 0x7e00000000000000) == 0x7e00000000000000) {
                pfpsf |= 1;
                res.w[1] = (y.w[1] & 0xfc003fffffffffff);
                res.w[0] = y.w[0];
            } else {
                res.w[1] = (y.w[1] & 0xfc003fffffffffff);
                res.w[0] = y.w[0];
            }
            return (res, pfpsf);
        } else {
            if ((x.w[1] & 0x7c00000000000000) == 0x7800000000000000) {
                x.w[1] = (x.w[1] & (0x8000000000000000 | 0x7800000000000000));
                x.w[0] = 0x0;
            }
            if ((y.w[1] & 0x7c00000000000000) == 0x7800000000000000) {
                y.w[1] = (y.w[1] & (0x8000000000000000 | 0x7800000000000000));
                y.w[0] = 0x0;
            }
        }
    }
    if ((x.w[1] & 0x7c00000000000000) != 0x7800000000000000) {
        if ((x.w[1] & 0x6000000000000000) == 0x6000000000000000) {
            x_exp = (((go_checked_shl_u64(x.w[1], go_shift_count_u64((2) as u64)))) & 0x7ffe000000000000);
            x.w[1] = ((x.w[1] & 0x8000000000000000) | x_exp);
            x.w[0] = 0x0;
        } else {
            x_exp = (x.w[1] & 0x7ffe000000000000);
            if (((x.w[1] & 0x1ffffffffffff) > 0x0001ed09bead87c0) || ((((x.w[1] & 0x1ffffffffffff) == 0x0001ed09bead87c0) && (x.w[0] > 0x378d8e63ffffffff)))) {
                x.w[1] = ((x.w[1] & 0x8000000000000000) | x_exp);
                x.w[0] = 0x0;
            } else {
            }
        }
    }
    tmp_fpsf = pfpsf;
    xnswp = x;
    ynswp = y;
    (res1, _) = bid128_quiet_equal(xnswp, ynswp);
    (res2, _) = bid128_quiet_greater(xnswp, ynswp);
    pfpsf = tmp_fpsf;
    if (res1 != 0) {
        res.w[1] = ((x.w[1] & 0x7fffffffffffffff) | (y.w[1] & 0x8000000000000000));
        res.w[0] = x.w[0];
    } else if (res2 != 0) {
        (res, tmp_fpsf) = bid128_next_down(xnswp);
        pfpsf |= tmp_fpsf;
    } else {
        (res, tmp_fpsf) = bid128_next_up(xnswp);
        pfpsf |= tmp_fpsf;
    }
    if ((((x.w[1] & 0x7800000000000000) != 0x7800000000000000)) && (((res.w[1] & 0x7800000000000000) == 0x7800000000000000))) {
        pfpsf |= 32;
        pfpsf |= 8;
    }
    tmp1.w[1] = 0x0000314dc6448d93;
    tmp1.w[0] = 0x38c15b0a00000000;
    tmp2.w[1] = (res.w[1] & 0x7fffffffffffffff);
    tmp2.w[0] = res.w[0];
    tmp3.w[1] = res.w[1];
    tmp3.w[0] = res.w[0];
    tmp_fpsf = pfpsf;
    (res1, _) = bid128_quiet_greater(tmp1, tmp2);
    (res2, _) = bid128_quiet_not_equal(xnswp, tmp3);
    pfpsf = tmp_fpsf;
    if ((res1 != 0) && (res2 != 0)) {
        pfpsf |= 32;
        pfpsf |= 16;
    }
    return (res, pfpsf);
}

pub fn bid128_next_toward(mut x: BID_UINT128, mut y: BID_UINT128) -> (BID_UINT128, u32) {
    let (mut res, mut pfpsf) = bid128_next_after(x, y);
    return (res, pfpsf);
}
