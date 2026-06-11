// Auto-generated from bid128_fma.go by go2rs. Do not edit.

use super::prelude::*;

pub(crate) fn bid128_ext_fma(mut x: BID_UINT128, mut y: BID_UINT128, mut z: BID_UINT128, mut rnd_mode: i64, pfpsf: &mut u32) -> (BID_UINT128, i64, i64, i64, i64) {
    let mut res: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut is_midpoint_lt_even: i64 = 0;
    let mut is_midpoint_gt_even: i64 = 0;
    let mut is_inexact_lt_midpoint: i64 = 0;
    let mut is_inexact_gt_midpoint: i64 = 0;
    res = BID_UINT128 { w: [0xbaddbaddbaddbadd, 0xbaddbaddbaddbadd], ..Default::default() };
    let mut x_sign: u64 = 0;
    let mut y_sign: u64 = 0;
    let mut z_sign: u64 = 0;
    let mut p_sign: u64 = 0;
    let mut x_exp: u64 = 0;
    let mut y_exp: u64 = 0;
    let mut z_exp: u64 = 0;
    let mut p_exp: u64 = 0;
    let mut true_p_exp: i64 = 0;
    let mut C1: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut C2: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut C3: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    if ((y.w[1] & 0x7c00000000000000) == 0x7c00000000000000) {
        if ((((y.w[1] & 0x00003fffffffffff) > 0x0000314dc6448d93)) || ((((y.w[1] & 0x00003fffffffffff) == 0x0000314dc6448d93) && (y.w[0] > 0x38c15b09ffffffff)))) {
            y.w[1] = (y.w[1] & 0xffffc00000000000);
            y.w[0] = 0;
        }
        if ((y.w[1] & 0x7e00000000000000) == 0x7e00000000000000) {
            (*pfpsf) |= 1;
            res.w[1] = (y.w[1] & 0xfc003fffffffffff);
            res.w[0] = y.w[0];
        } else {
            res.w[1] = (y.w[1] & 0xfc003fffffffffff);
            res.w[0] = y.w[0];
            if (((z.w[1] & 0x7e00000000000000) == 0x7e00000000000000) || ((x.w[1] & 0x7e00000000000000) == 0x7e00000000000000)) {
                (*pfpsf) |= 1;
            }
        }
        return (res, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint);
    } else if ((z.w[1] & 0x7c00000000000000) == 0x7c00000000000000) {
        if ((((z.w[1] & 0x00003fffffffffff) > 0x0000314dc6448d93)) || ((((z.w[1] & 0x00003fffffffffff) == 0x0000314dc6448d93) && (z.w[0] > 0x38c15b09ffffffff)))) {
            z.w[1] = (z.w[1] & 0xffffc00000000000);
            z.w[0] = 0;
        }
        if ((z.w[1] & 0x7e00000000000000) == 0x7e00000000000000) {
            (*pfpsf) |= 1;
            res.w[1] = (z.w[1] & 0xfc003fffffffffff);
            res.w[0] = z.w[0];
        } else {
            res.w[1] = (z.w[1] & 0xfc003fffffffffff);
            res.w[0] = z.w[0];
            if ((x.w[1] & 0x7e00000000000000) == 0x7e00000000000000) {
                (*pfpsf) |= 1;
            }
        }
        return (res, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint);
    } else if ((x.w[1] & 0x7c00000000000000) == 0x7c00000000000000) {
        if ((((x.w[1] & 0x00003fffffffffff) > 0x0000314dc6448d93)) || ((((x.w[1] & 0x00003fffffffffff) == 0x0000314dc6448d93) && (x.w[0] > 0x38c15b09ffffffff)))) {
            x.w[1] = (x.w[1] & 0xffffc00000000000);
            x.w[0] = 0;
        }
        if ((x.w[1] & 0x7e00000000000000) == 0x7e00000000000000) {
            (*pfpsf) |= 1;
            res.w[1] = (x.w[1] & 0xfc003fffffffffff);
            res.w[0] = x.w[0];
        } else {
            res.w[1] = (x.w[1] & 0xfc003fffffffffff);
            res.w[0] = x.w[0];
        }
        return (res, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint);
    }
    x_sign = (x.w[1] & 0x8000000000000000);
    C1.w[1] = (x.w[1] & 0x1ffffffffffff);
    C1.w[0] = x.w[0];
    if ((x.w[1] & 0x7c00000000000000) != 0x7800000000000000) {
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
    }
    y_sign = (y.w[1] & 0x8000000000000000);
    C2.w[1] = (y.w[1] & 0x1ffffffffffff);
    C2.w[0] = y.w[0];
    if ((y.w[1] & 0x7c00000000000000) != 0x7800000000000000) {
        if ((y.w[1] & 0x6000000000000000) == 0x6000000000000000) {
            y_exp = (((go_checked_shl_u64(y.w[1], go_shift_count_u64((2) as u64)))) & 0x7ffe000000000000);
            C2.w[1] = 0;
            C2.w[0] = 0;
        } else {
            y_exp = (y.w[1] & 0x7ffe000000000000);
            if ((C2.w[1] > 0x0001ed09bead87c0) || (((C2.w[1] == 0x0001ed09bead87c0) && (C2.w[0] > 0x378d8e63ffffffff)))) {
                C2.w[1] = 0;
                C2.w[0] = 0;
            }
        }
    }
    z_sign = (z.w[1] & 0x8000000000000000);
    C3.w[1] = (z.w[1] & 0x1ffffffffffff);
    C3.w[0] = z.w[0];
    if ((z.w[1] & 0x7c00000000000000) != 0x7800000000000000) {
        if ((z.w[1] & 0x6000000000000000) == 0x6000000000000000) {
            z_exp = (((go_checked_shl_u64(z.w[1], go_shift_count_u64((2) as u64)))) & 0x7ffe000000000000);
            C3.w[1] = 0;
            C3.w[0] = 0;
        } else {
            z_exp = (z.w[1] & 0x7ffe000000000000);
            if ((C3.w[1] > 0x0001ed09bead87c0) || (((C3.w[1] == 0x0001ed09bead87c0) && (C3.w[0] > 0x378d8e63ffffffff)))) {
                C3.w[1] = 0;
                C3.w[0] = 0;
            }
        }
    }
    p_sign = (x_sign ^ y_sign);
    if ((x.w[1] & 0x7c00000000000000) == 0x7800000000000000) {
        if ((y.w[1] & 0x7c00000000000000) == 0x7800000000000000) {
            if ((z.w[1] & 0x7c00000000000000) == 0x7800000000000000) {
                if (p_sign == z_sign) {
                    res.w[1] = (z_sign | 0x7800000000000000);
                    res.w[0] = 0;
                } else {
                    res.w[1] = 0x7c00000000000000;
                    res.w[0] = 0;
                    (*pfpsf) |= 1;
                }
            } else {
                res.w[1] = (p_sign | 0x7800000000000000);
                res.w[0] = 0;
            }
        } else if ((C2.w[1] != 0) || (C2.w[0] != 0)) {
            if ((z.w[1] & 0x7c00000000000000) == 0x7800000000000000) {
                if (p_sign == z_sign) {
                    res.w[1] = (z_sign | 0x7800000000000000);
                    res.w[0] = 0;
                } else {
                    res.w[1] = 0x7c00000000000000;
                    res.w[0] = 0;
                    (*pfpsf) |= 1;
                }
            } else {
                res.w[1] = (p_sign | 0x7800000000000000);
                res.w[0] = 0;
            }
        } else {
            res.w[1] = 0x7c00000000000000;
            res.w[0] = 0;
            (*pfpsf) |= 1;
        }
        return (res, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint);
    } else if ((y.w[1] & 0x7c00000000000000) == 0x7800000000000000) {
        if ((z.w[1] & 0x7c00000000000000) == 0x7800000000000000) {
            if ((p_sign != z_sign) || (((C1.w[1] == 0) && (C1.w[0] == 0)))) {
                res.w[1] = 0x7c00000000000000;
                res.w[0] = 0;
                (*pfpsf) |= 1;
            } else {
                res.w[1] = (z_sign | 0x7800000000000000);
                res.w[0] = 0;
            }
        } else if ((C1.w[1] == 0) && (C1.w[0] == 0)) {
            res.w[1] = 0x7c00000000000000;
            res.w[0] = 0;
            (*pfpsf) |= 1;
        } else {
            res.w[1] = (p_sign | 0x7800000000000000);
            res.w[0] = 0;
        }
        return (res, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint);
    } else if ((z.w[1] & 0x7c00000000000000) == 0x7800000000000000) {
        res.w[1] = (z_sign | 0x7800000000000000);
        res.w[0] = 0;
        return (res, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint);
    }
    true_p_exp = (((((go_checked_shr_u64(x_exp, go_shift_count_u64((49) as u64))) as i64).wrapping_sub(6176)).wrapping_add(((go_checked_shr_u64(y_exp, go_shift_count_u64((49) as u64))) as i64))).wrapping_sub(6176));
    if (true_p_exp < (-6176)) {
        p_exp = 0;
    } else {
        p_exp = (go_checked_shl_u64(((true_p_exp.wrapping_add(6176)) as u64), go_shift_count_u64((49) as u64)));
    }
    if (((((((C1.w[1] == 0) && (C1.w[0] == 0))) || (((C2.w[1] == 0) && (C2.w[0] == 0))))) && (C3.w[1] == 0)) && (C3.w[0] == 0)) {
        if (p_exp < z_exp) {
            res.w[1] = p_exp;
        } else {
            res.w[1] = z_exp;
        }
        if (p_sign == z_sign) {
            res.w[1] |= z_sign;
            res.w[0] = 0;
        } else {
            if (rnd_mode == 1) {
                res.w[1] |= 0x8000000000000000;
                res.w[0] = 0;
            } else {
                res.w[0] = 0;
            }
        }
        return (res, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint);
    }
    let mut q1: i64 = 0;
    let mut q2: i64 = 0;
    let mut q3: i64 = 0;
    let mut q4: i64 = 0;
    let mut e1: i64 = 0;
    let mut e2: i64 = 0;
    let mut e3: i64 = 0;
    let mut e4: i64 = 0;
    let mut C4: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut scale: i64 = 0;
    let mut ind: i64 = 0;
    if ((C1.w[1] != 0) || (C1.w[0] != 0)) {
        q1 = bid128_count_digits(C1);
    }
    if ((C2.w[1] != 0) || (C2.w[0] != 0)) {
        q2 = bid128_count_digits(C2);
    }
    if ((C3.w[1] != 0) || (C3.w[0] != 0)) {
        q3 = bid128_count_digits(C3);
    }
    if ((((C1.w[1] == 0) && (C1.w[0] == 0))) || (((C2.w[1] == 0) && (C2.w[0] == 0)))) {
        let mut p34: i64 = 34;
        if (z_exp <= p_exp) {
            res.w[1] = ((z_sign | (z_exp & 0x7ffe000000000000)) | C3.w[1]);
            res.w[0] = C3.w[0];
        } else {
            scale = (p34.wrapping_sub(q3));
            ind = ((go_checked_shr_u64(((z_exp.wrapping_sub(p_exp))), go_shift_count_u64((49) as u64))) as i64);
            if (ind < scale) {
                scale = ind;
            }
            if (scale == 0) {
                res.w[1] = z.w[1];
                res.w[0] = z.w[0];
            } else if (q3 <= 19) {
                if (scale <= 19) {
                    res = __mul_64x64_to_128(C3.w[0], bid_ten2k64[scale as usize]);
                } else {
                    (_, res) = __mul_64x128_full(C3.w[0], bid_ten2k128[(scale.wrapping_sub(20)) as usize]);
                }
            } else {
                (_, res) = __mul_64x128_full(bid_ten2k64[scale as usize], C3);
            }
            z_exp = (z_exp.wrapping_sub(((go_checked_shl_u64((scale as u64), go_shift_count_u64((49) as u64))))));
            res.w[1] = ((z_sign | (z_exp & 0x7ffe000000000000)) | res.w[1]);
        }
        return (res, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint);
    }
    e1 = (((go_checked_shr_u64(x_exp, go_shift_count_u64((49) as u64))) as i64).wrapping_sub(6176));
    e2 = (((go_checked_shr_u64(y_exp, go_shift_count_u64((49) as u64))) as i64).wrapping_sub(6176));
    e3 = (((go_checked_shr_u64(z_exp, go_shift_count_u64((49) as u64))) as i64).wrapping_sub(6176));
    e4 = (e1.wrapping_add(e2));
    C4.w[3] = 0;
    C4.w[2] = 0;
    C4.w[1] = 0;
    C4.w[0] = 0;
    if ((q1.wrapping_add(q2)) <= 19) {
        C4.w[0] = (C1.w[0].wrapping_mul(C2.w[0]));
        if (C4.w[0] < bid_ten2k64[((q1.wrapping_add(q2)).wrapping_sub(1)) as usize]) {
            q4 = ((q1.wrapping_add(q2)).wrapping_sub(1));
        } else {
            q4 = (q1.wrapping_add(q2));
        }
    } else if ((q1.wrapping_add(q2)) == 20) {
        let mut tmp128 = __mul_64x64_to_128(C1.w[0], C2.w[0]);
        C4.w[0] = tmp128.w[0];
        C4.w[1] = tmp128.w[1];
        if ((C4.w[1] == 0) && (C4.w[0] < bid_ten2k64[19])) {
            q4 = 19;
        } else {
            q4 = 20;
        }
    } else if ((q1.wrapping_add(q2)) <= 38) {
        let mut tmp128: BID_UINT128 = BID_UINT128 { w: [0, 0] };
        if (q1 <= 19) {
            (_, tmp128) = __mul_64x128_full(C1.w[0], C2);
        } else {
            (_, tmp128) = __mul_64x128_full(C2.w[0], C1);
        }
        C4.w[0] = tmp128.w[0];
        C4.w[1] = tmp128.w[1];
        if ((C4.w[1] < bid_ten2k128[((q1.wrapping_add(q2)).wrapping_sub(21)) as usize].w[1]) || (((C4.w[1] == bid_ten2k128[((q1.wrapping_add(q2)).wrapping_sub(21)) as usize].w[1]) && (C4.w[0] < bid_ten2k128[((q1.wrapping_add(q2)).wrapping_sub(21)) as usize].w[0])))) {
            q4 = ((q1.wrapping_add(q2)).wrapping_sub(1));
        } else {
            q4 = (q1.wrapping_add(q2));
        }
    } else if ((q1.wrapping_add(q2)) == 39) {
        C4 = __mul_128x128_to_256(C1, C2);
        if ((C4.w[2] == 0) && (((C4.w[1] < bid_ten2k128[18].w[1]) || (((C4.w[1] == bid_ten2k128[18].w[1]) && (C4.w[0] < bid_ten2k128[18].w[0])))))) {
            q4 = 38;
        } else {
            q4 = 39;
        }
    } else if ((q1.wrapping_add(q2)) <= 57) {
        C4 = __mul_128x128_to_256(C1, C2);
        if ((C4.w[2] < bid_ten2k256[((q1.wrapping_add(q2)).wrapping_sub(40)) as usize].w[2]) || (((C4.w[2] == bid_ten2k256[((q1.wrapping_add(q2)).wrapping_sub(40)) as usize].w[2]) && (((C4.w[1] < bid_ten2k256[((q1.wrapping_add(q2)).wrapping_sub(40)) as usize].w[1]) || (((C4.w[1] == bid_ten2k256[((q1.wrapping_add(q2)).wrapping_sub(40)) as usize].w[1]) && (C4.w[0] < bid_ten2k256[((q1.wrapping_add(q2)).wrapping_sub(40)) as usize].w[0])))))))) {
            q4 = ((q1.wrapping_add(q2)).wrapping_sub(1));
        } else {
            q4 = (q1.wrapping_add(q2));
        }
    } else if ((q1.wrapping_add(q2)) == 58) {
        C4 = __mul_128x128_to_256(C1, C2);
        if ((C4.w[3] == 0) && (((C4.w[2] < bid_ten2k256[18].w[2]) || (((C4.w[2] == bid_ten2k256[18].w[2]) && (((C4.w[1] < bid_ten2k256[18].w[1]) || (((C4.w[1] == bid_ten2k256[18].w[1]) && (C4.w[0] < bid_ten2k256[18].w[0])))))))))) {
            q4 = 57;
        } else {
            q4 = 58;
        }
    } else {
        C4 = __mul_128x128_to_256(C1, C2);
        if ((C4.w[3] < bid_ten2k256[((q1.wrapping_add(q2)).wrapping_sub(40)) as usize].w[3]) || (((C4.w[3] == bid_ten2k256[((q1.wrapping_add(q2)).wrapping_sub(40)) as usize].w[3]) && (((C4.w[2] < bid_ten2k256[((q1.wrapping_add(q2)).wrapping_sub(40)) as usize].w[2]) || (((C4.w[2] == bid_ten2k256[((q1.wrapping_add(q2)).wrapping_sub(40)) as usize].w[2]) && (((C4.w[1] < bid_ten2k256[((q1.wrapping_add(q2)).wrapping_sub(40)) as usize].w[1]) || (((C4.w[1] == bid_ten2k256[((q1.wrapping_add(q2)).wrapping_sub(40)) as usize].w[1]) && (C4.w[0] < bid_ten2k256[((q1.wrapping_add(q2)).wrapping_sub(40)) as usize].w[0])))))))))))) {
            q4 = ((q1.wrapping_add(q2)).wrapping_sub(1));
        } else {
            q4 = (q1.wrapping_add(q2));
        }
    }
    let mut save_fpsf: u32 = 0;
    let mut is_midpoint_lt_even0: i64 = 0;
    let mut is_midpoint_gt_even0: i64 = 0;
    let mut is_inexact_lt_midpoint0: i64 = 0;
    let mut is_inexact_gt_midpoint0: i64 = 0;
    let mut incr_exp: i64 = 0;
    let mut lt_half_ulp: i64 = 0;
    let mut eq_half_ulp: i64 = 0;
    let mut is_tiny: i64 = 0;
    let mut R64: u64 = 0;
    let mut P128: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut P192: BID_UINT192 = BID_UINT192 { w: [0, 0, 0] };
    let mut R192: BID_UINT192 = BID_UINT192 { w: [0, 0, 0] };
    let mut R256: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut x0: i64 = 0;
    let mut p34: i64 = 34;
    if ((C3.w[1] == 0x0) && (C3.w[0] == 0x0)) {
        save_fpsf = (*pfpsf);
        (*pfpsf) = 0;
        if (q4 > p34) {
            x0 = (q4.wrapping_sub(p34));
            if (q4 <= 38) {
                P128.w[1] = C4.w[1];
                P128.w[0] = C4.w[0];
                (res, incr_exp, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint) = bid_round128_19_38(q4, x0, P128);
            } else if (q4 <= 57) {
                P192.w[2] = C4.w[2];
                P192.w[1] = C4.w[1];
                P192.w[0] = C4.w[0];
                (R192, incr_exp, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint) = bid_round192_39_57(q4, x0, P192);
                res.w[0] = R192.w[0];
                res.w[1] = R192.w[1];
            } else {
                (R256, incr_exp, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint) = bid_round256_58_76(q4, x0, C4);
                res.w[0] = R256.w[0];
                res.w[1] = R256.w[1];
            }
            e4 = (e4.wrapping_add(x0));
            q4 = p34;
            if (incr_exp != 0) {
                e4 = (e4.wrapping_add(1));
                if ((q4.wrapping_add(e4)) == ((-6176 as i64).wrapping_add(p34))) {
                    (*pfpsf) |= (32 | 16);
                }
            }
        } else {
            if ((((q4.wrapping_add(e4)) <= (p34.wrapping_add(0x17df)))) && (e4 > 0x17df)) {
                scale = (e4.wrapping_sub(0x17df));
                if (q4 <= 19) {
                    if (scale <= 19) {
                        res = __mul_64x64_to_128(C4.w[0], bid_ten2k64[scale as usize]);
                    } else {
                        res = __mul_64x128_to_128(C4.w[0], bid_ten2k128[(scale.wrapping_sub(20)) as usize]);
                    }
                } else {
                    res = __mul_64x128_to_128(bid_ten2k64[scale as usize], BID_UINT128 { w: [C4.w[0], C4.w[1]], ..Default::default() });
                }
                e4 = (e4.wrapping_sub(scale));
                q4 = (q4.wrapping_add(scale));
            } else {
                res.w[1] = C4.w[1];
                res.w[0] = C4.w[0];
            }
        }
        if ((q4.wrapping_add(e4)) > (p34.wrapping_add(0x17df))) {
            if (rnd_mode == 0) {
                res.w[1] = (p_sign | 0x7800000000000000);
                res.w[0] = 0x0000000000000000;
                (*pfpsf) |= (32 | 8);
            } else {
                res.w[1] = (p_sign | res.w[1]);
                bid_rounding_correction(rnd_mode, is_inexact_lt_midpoint, is_inexact_gt_midpoint, is_midpoint_lt_even, is_midpoint_gt_even, e4, (&mut res), pfpsf);
            }
            (*pfpsf) |= save_fpsf;
            return (res, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint);
        }
        if ((q4.wrapping_add(e4)) < ((-6176 as i64).wrapping_add(p34))) {
            is_tiny = 1;
            if (e4 < -6176) {
                x0 = ((-6176 as i64).wrapping_sub(e4));
                is_inexact_lt_midpoint0 = is_inexact_lt_midpoint;
                is_inexact_gt_midpoint0 = is_inexact_gt_midpoint;
                is_midpoint_lt_even0 = is_midpoint_lt_even;
                is_midpoint_gt_even0 = is_midpoint_gt_even;
                is_inexact_lt_midpoint = 0;
                is_inexact_gt_midpoint = 0;
                is_midpoint_lt_even = 0;
                is_midpoint_gt_even = 0;
                if (x0 < q4) {
                    if (q4 <= 18) {
                        R64 = bid_round64_2_18(q4, x0, res.w[0], (&mut incr_exp), (&mut is_midpoint_lt_even), (&mut is_midpoint_gt_even), (&mut is_inexact_lt_midpoint), (&mut is_inexact_gt_midpoint));
                        if (incr_exp != 0) {
                            R64 = bid_ten2k64[(q4.wrapping_sub(x0)) as usize];
                        }
                        res.w[0] = R64;
                    } else {
                        P128.w[1] = res.w[1];
                        P128.w[0] = res.w[0];
                        (res, incr_exp, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint) = bid_round128_19_38(q4, x0, P128);
                        if (incr_exp != 0) {
                            if ((q4.wrapping_sub(x0)) <= 19) {
                                res.w[0] = bid_ten2k64[(q4.wrapping_sub(x0)) as usize];
                            } else {
                                res.w[0] = bid_ten2k128[((q4.wrapping_sub(x0)).wrapping_sub(20)) as usize].w[0];
                                res.w[1] = bid_ten2k128[((q4.wrapping_sub(x0)).wrapping_sub(20)) as usize].w[1];
                            }
                        }
                    }
                    e4 = (e4.wrapping_add(x0));
                } else if (x0 == q4) {
                    if (q4 <= 19) {
                        if (res.w[0] < bid_midpoint64[(q4.wrapping_sub(1)) as usize]) {
                            lt_half_ulp = 1;
                            is_inexact_lt_midpoint = 1;
                        } else if (res.w[0] == bid_midpoint64[(q4.wrapping_sub(1)) as usize]) {
                            eq_half_ulp = 1;
                            is_midpoint_gt_even = 1;
                        } else {
                            is_inexact_gt_midpoint = 1;
                        }
                    } else {
                        if ((res.w[1] < bid_midpoint128[(q4.wrapping_sub(20)) as usize].w[1]) || (((res.w[1] == bid_midpoint128[(q4.wrapping_sub(20)) as usize].w[1]) && (res.w[0] < bid_midpoint128[(q4.wrapping_sub(20)) as usize].w[0])))) {
                            lt_half_ulp = 1;
                            is_inexact_lt_midpoint = 1;
                        } else if ((res.w[1] == bid_midpoint128[(q4.wrapping_sub(20)) as usize].w[1]) && (res.w[0] == bid_midpoint128[(q4.wrapping_sub(20)) as usize].w[0])) {
                            eq_half_ulp = 1;
                            is_midpoint_gt_even = 1;
                        } else {
                            is_inexact_gt_midpoint = 1;
                        }
                    }
                    if ((lt_half_ulp != 0) || (eq_half_ulp != 0)) {
                        res.w[1] = 0x0000000000000000;
                        res.w[0] = 0x0000000000000000;
                    } else {
                        res.w[1] = 0x0000000000000000;
                        res.w[0] = 0x0000000000000001;
                    }
                    e4 = -6176;
                } else {
                    res.w[1] = 0;
                    res.w[0] = 0;
                    e4 = -6176;
                    is_inexact_lt_midpoint = 1;
                }
                if ((((is_inexact_gt_midpoint0 != 0) || (is_midpoint_lt_even0 != 0))) && (is_midpoint_lt_even != 0)) {
                    res.w[0] = res.w[0].wrapping_sub(1);
                    if (res.w[0] == 0xffffffffffffffff) {
                        res.w[1] = res.w[1].wrapping_sub(1);
                    }
                    is_midpoint_lt_even = 0;
                    is_inexact_lt_midpoint = 1;
                } else if ((((is_inexact_lt_midpoint0 != 0) || (is_midpoint_gt_even0 != 0))) && (is_midpoint_gt_even != 0)) {
                    res.w[0] = res.w[0].wrapping_add(1);
                    if (res.w[0] == 0) {
                        res.w[1] = res.w[1].wrapping_add(1);
                    }
                    is_midpoint_gt_even = 0;
                    is_inexact_gt_midpoint = 1;
                } else if ((((is_midpoint_lt_even == 0) && (is_midpoint_gt_even == 0)) && (is_inexact_lt_midpoint == 0)) && (is_inexact_gt_midpoint == 0)) {
                    if ((is_inexact_gt_midpoint0 != 0) || (is_midpoint_lt_even0 != 0)) {
                        is_inexact_gt_midpoint = 1;
                    }
                    if ((is_inexact_lt_midpoint0 != 0) || (is_midpoint_gt_even0 != 0)) {
                        is_inexact_lt_midpoint = 1;
                    }
                } else if ((is_midpoint_gt_even != 0) && (((is_inexact_gt_midpoint0 != 0) || (is_midpoint_lt_even0 != 0)))) {
                    is_inexact_lt_midpoint = 1;
                    is_inexact_gt_midpoint = 0;
                    is_midpoint_lt_even = 0;
                    is_midpoint_gt_even = 0;
                } else if ((is_midpoint_lt_even != 0) && (((is_inexact_lt_midpoint0 != 0) || (is_midpoint_gt_even0 != 0)))) {
                    is_inexact_lt_midpoint = 0;
                    is_inexact_gt_midpoint = 1;
                    is_midpoint_lt_even = 0;
                    is_midpoint_gt_even = 0;
                } else {
                }
            } else {
                if (e3 < e4) {
                    scale = (p34.wrapping_sub(q4));
                    ind = (e4.wrapping_sub(e3));
                    if (ind < scale) {
                        scale = ind;
                    }
                    if (scale == 0) {
                    } else if (q4 <= 19) {
                        if (scale <= 19) {
                            res = __mul_64x64_to_128(res.w[0], bid_ten2k64[scale as usize]);
                        } else {
                            res = __mul_64x128_to_128(res.w[0], bid_ten2k128[(scale.wrapping_sub(20)) as usize]);
                        }
                    } else {
                        res = __mul_64x128_to_128(bid_ten2k64[scale as usize], res);
                    }
                    e4 = (e4.wrapping_sub(scale));
                }
            }
            if ((((is_inexact_lt_midpoint != 0) || (is_inexact_gt_midpoint != 0)) || (is_midpoint_lt_even != 0)) || (is_midpoint_gt_even != 0)) {
                (*pfpsf) |= 32;
                (*pfpsf) |= 16;
            }
            res.w[1] = ((p_sign | ((go_checked_shl_u64(((e4.wrapping_add(6176)) as u64), go_shift_count_u64((49) as u64))))) | res.w[1]);
            if (rnd_mode != 0) {
                bid_rounding_correction(rnd_mode, is_inexact_lt_midpoint, is_inexact_gt_midpoint, is_midpoint_lt_even, is_midpoint_gt_even, e4, (&mut res), pfpsf);
            }
            (*pfpsf) |= save_fpsf;
            return (res, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint);
        }
        res.w[1] = ((p_sign | ((go_checked_shl_u64(((e4.wrapping_add(6176)) as u64), go_shift_count_u64((49) as u64))))) | res.w[1]);
        if (rnd_mode != 0) {
            bid_rounding_correction(rnd_mode, is_inexact_lt_midpoint, is_inexact_gt_midpoint, is_midpoint_lt_even, is_midpoint_gt_even, e4, (&mut res), pfpsf);
            if (e4 == -6176) {
                if (((res.w[1] & 0x1ffffffffffff) < 0x0000314dc6448d93) || ((((res.w[1] & 0x1ffffffffffff) == 0x0000314dc6448d93) && (res.w[0] < 0x38c15b0a00000000)))) {
                    is_tiny = 1;
                }
            }
        }
        if ((((is_inexact_lt_midpoint != 0) || (is_inexact_gt_midpoint != 0)) || (is_midpoint_lt_even != 0)) || (is_midpoint_gt_even != 0)) {
            (*pfpsf) |= 32;
            if (is_tiny != 0) {
                (*pfpsf) |= 16;
            }
        }
        if ((((*pfpsf) & 32)) == 0) {
            p_exp = (res.w[1] & 0x7ffe000000000000);
            if (z_exp < p_exp) {
                C3.w[1] = (res.w[1] & 0x1ffffffffffff);
                C3.w[0] = res.w[0];
                scale = (p34.wrapping_sub(q4));
                ind = ((go_checked_shr_u64(((p_exp.wrapping_sub(z_exp))), go_shift_count_u64((49) as u64))) as i64);
                if (ind < scale) {
                    scale = ind;
                }
                p_exp = (p_exp.wrapping_sub(((go_checked_shl_u64((scale as u64), go_shift_count_u64((49) as u64))))));
                if (scale == 0) {
                } else if (q4 <= 19) {
                    if (scale <= 19) {
                        res = __mul_64x64_to_128(C3.w[0], bid_ten2k64[scale as usize]);
                    } else {
                        res = __mul_64x128_to_128(C3.w[0], bid_ten2k128[(scale.wrapping_sub(20)) as usize]);
                    }
                    res.w[1] = ((p_sign | (p_exp & 0x7ffe000000000000)) | res.w[1]);
                } else {
                    res = __mul_64x128_to_128(bid_ten2k64[scale as usize], C3);
                    res.w[1] = ((p_sign | (p_exp & 0x7ffe000000000000)) | res.w[1]);
                }
            }
        }
        (*pfpsf) |= save_fpsf;
        return (res, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint);
    }
    let mut delta = (((q3.wrapping_add(e3)).wrapping_sub(q4)).wrapping_sub(e4));
    bid_fma_main_body(p34, (&mut res), (&mut is_midpoint_lt_even), (&mut is_midpoint_gt_even), (&mut is_inexact_lt_midpoint), (&mut is_inexact_gt_midpoint), p_sign, z_sign, (&mut z_exp), (&mut p_exp), q3, q4, (&mut e3), (&mut e4), delta, (&mut C3), C4, rnd_mode, pfpsf);
    return (res, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint);
}

pub fn bid128_fma(mut x: BID_UINT128, mut y: BID_UINT128, mut z: BID_UINT128, mut rnd_mode: i64) -> (BID_UINT128, u32) {
    let mut pfpsf: u32 = 0;
    let (mut res, _, _, _, _) = bid128_ext_fma(x, y, z, rnd_mode, (&mut pfpsf));
    return (res, pfpsf);
}
