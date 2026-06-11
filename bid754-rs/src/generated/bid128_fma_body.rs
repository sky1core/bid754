// Auto-generated from bid128_fma_body.go by go2rs. Do not edit.

use super::prelude::*;

pub(crate) fn bid_fma_main_body(mut p34: i64, res: &mut BID_UINT128, ptr_is_midpoint_lt_even: &mut i64, ptr_is_midpoint_gt_even: &mut i64, ptr_is_inexact_lt_midpoint: &mut i64, ptr_is_inexact_gt_midpoint: &mut i64, mut p_sign: u64, mut z_sign: u64, z_exp_ptr: &mut u64, p_exp_ptr: &mut u64, mut q3: i64, mut q4: i64, e3_ptr: &mut i64, e4_ptr: &mut i64, mut delta: i64, C3: &mut BID_UINT128, mut C4: BID_UINT256, mut rnd_mode: i64, pfpsf: &mut u32) {
    let mut z_exp = (*z_exp_ptr);
    let mut p_exp = (*p_exp_ptr);
    let mut e3 = (*e3_ptr);
    let mut e4 = (*e4_ptr);
    let mut is_midpoint_lt_even: i64 = 0;
    let mut is_midpoint_gt_even: i64 = 0;
    let mut is_inexact_lt_midpoint: i64 = 0;
    let mut is_inexact_gt_midpoint: i64 = 0;
    if (delta >= 0) {
        bid_fma_delta_ge_zero(p34, res, (&mut is_midpoint_lt_even), (&mut is_midpoint_gt_even), (&mut is_inexact_lt_midpoint), (&mut is_inexact_gt_midpoint), p_sign, z_sign, (&mut z_exp), (&mut p_exp), q3, q4, (&mut e3), (&mut e4), delta, C3, C4, rnd_mode, pfpsf);
    } else {
        bid_fma_delta_lt_zero(p34, res, (&mut is_midpoint_lt_even), (&mut is_midpoint_gt_even), (&mut is_inexact_lt_midpoint), (&mut is_inexact_gt_midpoint), p_sign, z_sign, (&mut z_exp), (&mut p_exp), q3, q4, (&mut e3), (&mut e4), delta, C3, C4, rnd_mode, pfpsf);
    }
    (*ptr_is_midpoint_lt_even) = is_midpoint_lt_even;
    (*ptr_is_midpoint_gt_even) = is_midpoint_gt_even;
    (*ptr_is_inexact_lt_midpoint) = is_inexact_lt_midpoint;
    (*ptr_is_inexact_gt_midpoint) = is_inexact_gt_midpoint;
    (*z_exp_ptr) = z_exp;
    (*p_exp_ptr) = p_exp;
    (*e3_ptr) = e3;
    (*e4_ptr) = e4;
}

pub(crate) fn bid_fma_delta_ge_zero(mut p34: i64, res: &mut BID_UINT128, ptr_is_midpoint_lt_even: &mut i64, ptr_is_midpoint_gt_even: &mut i64, ptr_is_inexact_lt_midpoint: &mut i64, ptr_is_inexact_gt_midpoint: &mut i64, mut p_sign: u64, mut z_sign: u64, z_exp_ptr: &mut u64, p_exp_ptr: &mut u64, mut q3: i64, mut q4: i64, e3_ptr: &mut i64, e4_ptr: &mut i64, mut delta: i64, C3: &mut BID_UINT128, mut C4: BID_UINT256, mut rnd_mode: i64, pfpsf: &mut u32) {
    let mut z_exp = (*z_exp_ptr);
    let mut p_exp = (*p_exp_ptr);
    let mut e3 = (*e3_ptr);
    let mut e4 = (*e4_ptr);
    let mut is_midpoint_lt_even: i64 = 0;
    let mut is_midpoint_gt_even: i64 = 0;
    let mut is_inexact_lt_midpoint: i64 = 0;
    let mut is_inexact_gt_midpoint: i64 = 0;
    let mut is_midpoint_lt_even0: i64 = 0;
    let mut is_midpoint_gt_even0: i64 = 0;
    let mut is_inexact_lt_midpoint0: i64 = 0;
    let mut is_inexact_gt_midpoint0: i64 = 0;
    let mut incr_exp: i64 = 0;
    let mut lt_half_ulp: i64 = 0;
    let mut eq_half_ulp: i64 = 0;
    let mut gt_half_ulp: i64 = 0;
    let mut is_tiny: i64 = 0;
    let mut R64: u64 = 0;
    let mut P128: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut R128: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut P192: BID_UINT192 = BID_UINT192 { w: [0, 0, 0] };
    let mut R192: BID_UINT192 = BID_UINT192 { w: [0, 0, 0] };
    let mut R256: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut scale: i64 = 0;
    let mut ind: i64 = 0;
    let mut x0: i64 = 0;
    let mut tmp_sign: u64 = 0;
    _ = is_midpoint_lt_even0;
    _ = is_midpoint_gt_even0;
    _ = is_inexact_lt_midpoint0;
    _ = is_inexact_gt_midpoint0;
    _ = is_tiny;
    _ = x0;
    _ = R128;
    _ = R192;
    _ = R256;
    _ = R64;
    _ = P128;
    _ = P192;
    _ = ind;
    _ = tmp_sign;
    'done: {
    if ((p34 <= (delta.wrapping_sub(1))) || (((p34 == delta) && ((e3.wrapping_add(6176)) < (p34.wrapping_sub(q3)))))) {
        if ((((q3.wrapping_add(e3))) > ((p34.wrapping_add(0x17df)))) && (p34 <= (delta.wrapping_sub(1)))) {
            if (rnd_mode == 0) {
                res.w[1] = (z_sign | 0x7800000000000000);
                res.w[0] = 0x0000000000000000;
                (*pfpsf) |= (32 | 8);
            } else {
                if (p_sign == z_sign) {
                    is_inexact_lt_midpoint = 1;
                } else {
                    is_inexact_gt_midpoint = 1;
                }
                scale = (p34.wrapping_sub(q3));
                if (scale == 0) {
                    res.w[1] = (z_sign | C3.w[1]);
                    res.w[0] = C3.w[0];
                } else {
                    if (q3 <= 19) {
                        if (scale <= 19) {
                            (*res) = __mul_64x64_to_128(C3.w[0], bid_ten2k64[scale as usize]);
                        } else {
                            (*res) = __mul_64x128_to_128(C3.w[0], bid_ten2k128[(scale.wrapping_sub(20)) as usize]);
                        }
                    } else {
                        (*res) = __mul_64x128_to_128(bid_ten2k64[scale as usize], (*C3));
                    }
                }
                e3 = (e3.wrapping_sub(scale));
                res.w[1] = (z_sign | res.w[1]);
                bid_rounding_correction(rnd_mode, is_inexact_lt_midpoint, is_inexact_gt_midpoint, is_midpoint_lt_even, is_midpoint_gt_even, e3, res, pfpsf);
            }
            (*ptr_is_midpoint_lt_even) = is_midpoint_lt_even;
            (*ptr_is_midpoint_gt_even) = is_midpoint_gt_even;
            (*ptr_is_inexact_lt_midpoint) = is_inexact_lt_midpoint;
            (*ptr_is_inexact_gt_midpoint) = is_inexact_gt_midpoint;
            break 'done;
        }
        if (q3 < p34) {
            scale = (p34.wrapping_sub(q3));
            ind = (e3.wrapping_add(6176));
            if (ind < scale) {
                scale = ind;
            }
            if (scale == 0) {
                res.w[1] = C3.w[1];
                res.w[0] = C3.w[0];
            } else if (q3 <= 19) {
                if (scale <= 19) {
                    (*res) = __mul_64x64_to_128(C3.w[0], bid_ten2k64[scale as usize]);
                } else {
                    (*res) = __mul_64x128_to_128(C3.w[0], bid_ten2k128[(scale.wrapping_sub(20)) as usize]);
                }
            } else {
                (*res) = __mul_64x128_to_128(bid_ten2k64[scale as usize], (*C3));
            }
            z_exp = (z_exp.wrapping_sub(((go_checked_shl_u64((scale as u64), go_shift_count_u64((49) as u64))))));
            e3 = (e3.wrapping_sub(scale));
            res.w[1] = ((z_sign | (z_exp & 0x7ffe000000000000)) | res.w[1]);
            if ((scale.wrapping_add(q3)) < p34) {
                (*pfpsf) |= 16;
            }
        } else {
            scale = 0;
            res.w[1] = ((z_sign | ((go_checked_shl_u64(((e3.wrapping_add(6176)) as u64), go_shift_count_u64((49) as u64))))) | C3.w[1]);
            res.w[0] = C3.w[0];
        }
        if ((p_sign != z_sign) && ((delta == (((q3.wrapping_add(scale)).wrapping_add(1)))))) {
            if (((((q3 <= 19) && (C3.w[0] != bid_ten2k64[(q3.wrapping_sub(1)) as usize]))) || (((q3 == 20) && (((C3.w[1] != 0) || (C3.w[0] != bid_ten2k64[19])))))) || (((q3 >= 21) && (((C3.w[1] != bid_ten2k128[(q3.wrapping_sub(21)) as usize].w[1]) || (C3.w[0] != bid_ten2k128[(q3.wrapping_sub(21)) as usize].w[0])))))) {
                is_inexact_gt_midpoint = 1;
            } else {
                if (q4 == 1) {
                    R64 = C4.w[0];
                } else {
                    if (q4 <= 18) {
                        R64 = bid_round64_2_18(q4, (q4.wrapping_sub(1)), C4.w[0], (&mut incr_exp), (&mut is_midpoint_lt_even), (&mut is_midpoint_gt_even), (&mut is_inexact_lt_midpoint), (&mut is_inexact_gt_midpoint));
                    } else if (q4 <= 38) {
                        P128.w[1] = C4.w[1];
                        P128.w[0] = C4.w[0];
                        (R128, incr_exp, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint) = bid_round128_19_38(q4, (q4.wrapping_sub(1)), P128);
                        R64 = R128.w[0];
                    } else if (q4 <= 57) {
                        P192.w[2] = C4.w[2];
                        P192.w[1] = C4.w[1];
                        P192.w[0] = C4.w[0];
                        (R192, incr_exp, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint) = bid_round192_39_57(q4, (q4.wrapping_sub(1)), P192);
                        R64 = R192.w[0];
                    } else {
                        (R256, incr_exp, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint) = bid_round256_58_76(q4, (q4.wrapping_sub(1)), C4);
                        R64 = R256.w[0];
                    }
                    if (incr_exp != 0) {
                        R64 = 10;
                    }
                }
                if (((((R64 == 5) && (is_inexact_lt_midpoint == 0)) && (is_inexact_gt_midpoint == 0)) && (is_midpoint_lt_even == 0)) && (is_midpoint_gt_even == 0)) {
                    is_inexact_lt_midpoint = 0;
                    is_inexact_gt_midpoint = 0;
                    is_midpoint_lt_even = 1;
                    is_midpoint_gt_even = 0;
                } else if (((e3 == -6176) || (R64 < 5)) || (((R64 == 5) && (is_inexact_gt_midpoint != 0)))) {
                    is_inexact_lt_midpoint = 0;
                    is_inexact_gt_midpoint = 1;
                    is_midpoint_lt_even = 0;
                    is_midpoint_gt_even = 0;
                } else {
                    is_inexact_lt_midpoint = 1;
                    is_inexact_gt_midpoint = 0;
                    is_midpoint_lt_even = 0;
                    is_midpoint_gt_even = 0;
                    if (((q3.wrapping_add(scale))) <= 19) {
                        res.w[1] = 0;
                        res.w[0] = bid_ten2k64[(q3.wrapping_add(scale)) as usize];
                    } else {
                        res.w[1] = bid_ten2k128[((q3.wrapping_add(scale)).wrapping_sub(20)) as usize].w[1];
                        res.w[0] = bid_ten2k128[((q3.wrapping_add(scale)).wrapping_sub(20)) as usize].w[0];
                    }
                    res.w[0] = (res.w[0].wrapping_sub(1));
                    z_exp = (z_exp.wrapping_sub(0x2000000000000));
                    e3 = (e3.wrapping_sub(1));
                    res.w[1] = ((z_sign | ((go_checked_shl_u64(((e3.wrapping_add(6176)) as u64), go_shift_count_u64((49) as u64))))) | res.w[1]);
                }
                if (e3 == -6176) {
                    (*pfpsf) |= 16;
                }
            }
            (*pfpsf) |= 32;
        } else {
            if (p_sign == z_sign) {
                is_inexact_lt_midpoint = 1;
            } else {
                is_inexact_gt_midpoint = 1;
            }
            (*pfpsf) |= 32;
        }
        if ((((e3 == -6176) && (((q3.wrapping_add(scale))) < p34))) || ((((((e3 == -6176) && (((q3.wrapping_add(scale))) == p34)) && ((res.w[1] & 0x1ffffffffffff) == 0x0000314dc6448d93)) && (res.w[0] == 0x38c15b0a00000000)) && (z_sign != p_sign)))) {
            (*pfpsf) |= 16;
        }
        if (rnd_mode != 0) {
            bid_rounding_correction(rnd_mode, is_inexact_lt_midpoint, is_inexact_gt_midpoint, is_midpoint_lt_even, is_midpoint_gt_even, e3, res, pfpsf);
        }
        (*ptr_is_midpoint_lt_even) = is_midpoint_lt_even;
        (*ptr_is_midpoint_gt_even) = is_midpoint_gt_even;
        (*ptr_is_inexact_lt_midpoint) = is_inexact_lt_midpoint;
        (*ptr_is_inexact_gt_midpoint) = is_inexact_gt_midpoint;
        break 'done;
    } else if (p34 == delta) {
        scale = (p34.wrapping_sub(q3));
        if (scale == 0) {
            res.w[1] = C3.w[1];
            res.w[0] = C3.w[0];
        } else if (q3 <= 19) {
            if (scale <= 19) {
                (*res) = __mul_64x64_to_128(C3.w[0], bid_ten2k64[scale as usize]);
            } else {
                (*res) = __mul_64x128_to_128(C3.w[0], bid_ten2k128[(scale.wrapping_sub(20)) as usize]);
            }
        } else {
            (*res) = __mul_64x128_to_128(bid_ten2k64[scale as usize], (*C3));
        }
        z_exp = (z_exp.wrapping_sub(((go_checked_shl_u64((scale as u64), go_shift_count_u64((49) as u64))))));
        e3 = (e3.wrapping_sub(scale));
        lt_half_ulp = 0;
        eq_half_ulp = 0;
        gt_half_ulp = 0;
        if (q4 <= 19) {
            if (C4.w[0] < bid_midpoint64[(q4.wrapping_sub(1)) as usize]) {
                lt_half_ulp = 1;
            } else if (C4.w[0] == bid_midpoint64[(q4.wrapping_sub(1)) as usize]) {
                eq_half_ulp = 1;
            } else {
                gt_half_ulp = 1;
            }
        } else if (q4 <= 38) {
            if ((C4.w[2] == 0) && (((C4.w[1] < bid_midpoint128[(q4.wrapping_sub(20)) as usize].w[1]) || (((C4.w[1] == bid_midpoint128[(q4.wrapping_sub(20)) as usize].w[1]) && (C4.w[0] < bid_midpoint128[(q4.wrapping_sub(20)) as usize].w[0])))))) {
                lt_half_ulp = 1;
            } else if (((C4.w[2] == 0) && (C4.w[1] == bid_midpoint128[(q4.wrapping_sub(20)) as usize].w[1])) && (C4.w[0] == bid_midpoint128[(q4.wrapping_sub(20)) as usize].w[0])) {
                eq_half_ulp = 1;
            } else {
                gt_half_ulp = 1;
            }
        } else if (q4 <= 58) {
            if ((C4.w[3] == 0) && ((((C4.w[2] < bid_midpoint192[(q4.wrapping_sub(39)) as usize].w[2]) || (((C4.w[2] == bid_midpoint192[(q4.wrapping_sub(39)) as usize].w[2]) && (C4.w[1] < bid_midpoint192[(q4.wrapping_sub(39)) as usize].w[1])))) || ((((C4.w[2] == bid_midpoint192[(q4.wrapping_sub(39)) as usize].w[2]) && (C4.w[1] == bid_midpoint192[(q4.wrapping_sub(39)) as usize].w[1])) && (C4.w[0] < bid_midpoint192[(q4.wrapping_sub(39)) as usize].w[0])))))) {
                lt_half_ulp = 1;
            } else if ((((C4.w[3] == 0) && (C4.w[2] == bid_midpoint192[(q4.wrapping_sub(39)) as usize].w[2])) && (C4.w[1] == bid_midpoint192[(q4.wrapping_sub(39)) as usize].w[1])) && (C4.w[0] == bid_midpoint192[(q4.wrapping_sub(39)) as usize].w[0])) {
                eq_half_ulp = 1;
            } else {
                gt_half_ulp = 1;
            }
        } else {
            if ((((C4.w[3] < bid_midpoint256[(q4.wrapping_sub(59)) as usize].w[3]) || (((C4.w[3] == bid_midpoint256[(q4.wrapping_sub(59)) as usize].w[3]) && (C4.w[2] < bid_midpoint256[(q4.wrapping_sub(59)) as usize].w[2])))) || ((((C4.w[3] == bid_midpoint256[(q4.wrapping_sub(59)) as usize].w[3]) && (C4.w[2] == bid_midpoint256[(q4.wrapping_sub(59)) as usize].w[2])) && (C4.w[1] < bid_midpoint256[(q4.wrapping_sub(59)) as usize].w[1])))) || (((((C4.w[3] == bid_midpoint256[(q4.wrapping_sub(59)) as usize].w[3]) && (C4.w[2] == bid_midpoint256[(q4.wrapping_sub(59)) as usize].w[2])) && (C4.w[1] == bid_midpoint256[(q4.wrapping_sub(59)) as usize].w[1])) && (C4.w[0] < bid_midpoint256[(q4.wrapping_sub(59)) as usize].w[0])))) {
                lt_half_ulp = 1;
            } else if ((((C4.w[3] == bid_midpoint256[(q4.wrapping_sub(59)) as usize].w[3]) && (C4.w[2] == bid_midpoint256[(q4.wrapping_sub(59)) as usize].w[2])) && (C4.w[1] == bid_midpoint256[(q4.wrapping_sub(59)) as usize].w[1])) && (C4.w[0] == bid_midpoint256[(q4.wrapping_sub(59)) as usize].w[0])) {
                eq_half_ulp = 1;
            } else {
                gt_half_ulp = 1;
            }
        }
        if (p_sign == z_sign) {
            if (lt_half_ulp != 0) {
                res.w[1] = ((z_sign | (z_exp & 0x7ffe000000000000)) | res.w[1]);
                is_inexact_lt_midpoint = 1;
            } else if ((((eq_half_ulp != 0) && ((res.w[0] & 0x01) != 0))) || (gt_half_ulp != 0)) {
                res.w[0] = res.w[0].wrapping_add(1);
                if (res.w[0] == 0x0) {
                    res.w[1] = res.w[1].wrapping_add(1);
                }
                if (((res.w[1] & 0x1ffffffffffff) == 0x0001ed09bead87c0) && (res.w[0] == 0x378d8e6400000000)) {
                    e3 = (e3.wrapping_add(1));
                    z_exp = (((go_checked_shl_u64(((e3.wrapping_add(6176)) as u64), go_shift_count_u64((49) as u64)))) & 0x7ffe000000000000);
                    res.w[1] = 0x0000314dc6448d93;
                    res.w[0] = 0x38c15b0a00000000;
                }
                res.w[1] = ((z_sign | (z_exp & 0x7ffe000000000000)) | res.w[1]);
                if (eq_half_ulp != 0) {
                    is_midpoint_lt_even = 1;
                } else {
                    is_inexact_gt_midpoint = 1;
                }
            } else {
                res.w[1] = ((z_sign | (z_exp & 0x7ffe000000000000)) | res.w[1]);
                is_midpoint_gt_even = 1;
            }
            (*pfpsf) |= 32;
            if ((e3 > 0x17df) && (rnd_mode == 0)) {
                res.w[1] = (z_sign | 0x7800000000000000);
                res.w[0] = 0x0000000000000000;
                (*pfpsf) |= (32 | 8);
                (*ptr_is_midpoint_lt_even) = is_midpoint_lt_even;
                (*ptr_is_midpoint_gt_even) = is_midpoint_gt_even;
                (*ptr_is_inexact_lt_midpoint) = is_inexact_lt_midpoint;
                (*ptr_is_inexact_gt_midpoint) = is_inexact_gt_midpoint;
                break 'done;
            }
            if (rnd_mode != 0) {
                bid_rounding_correction(rnd_mode, is_inexact_lt_midpoint, is_inexact_gt_midpoint, is_midpoint_lt_even, is_midpoint_gt_even, e3, res, pfpsf);
                z_exp = (res.w[1] & 0x7ffe000000000000);
            }
        } else {
            bid_fma_case1pp_b_psign_ne_zsign(p34, res, (&mut is_midpoint_lt_even), (&mut is_midpoint_gt_even), (&mut is_inexact_lt_midpoint), (&mut is_inexact_gt_midpoint), p_sign, z_sign, (&mut z_exp), q3, q4, (&mut e3), scale, C3, C4, lt_half_ulp, eq_half_ulp, gt_half_ulp, rnd_mode, pfpsf);
        }
        res.w[1] = ((z_sign | (z_exp & 0x7ffe000000000000)) | (res.w[1] & 0x1ffffffffffff));
        (*ptr_is_midpoint_lt_even) = is_midpoint_lt_even;
        (*ptr_is_midpoint_gt_even) = is_midpoint_gt_even;
        (*ptr_is_inexact_lt_midpoint) = is_inexact_lt_midpoint;
        (*ptr_is_inexact_gt_midpoint) = is_inexact_gt_midpoint;
        break 'done;
    } else if ((((((((((q3 <= delta) && (delta < p34)) && (p34 < (delta.wrapping_add(q4))))) || (((q3 <= delta) && ((delta.wrapping_add(q4)) <= p34)))) || (((delta < q3) && (p34 < (delta.wrapping_add(q4)))))) || ((((delta < q3) && (q3 <= (delta.wrapping_add(q4)))) && ((delta.wrapping_add(q4)) <= p34)))) || (((delta.wrapping_add(q4)) < q3)))) && (!(((delta <= 1) && (p_sign != z_sign))))) {
        bid_fma_cases_2_to_6(p34, res, (&mut is_midpoint_lt_even), (&mut is_midpoint_gt_even), (&mut is_inexact_lt_midpoint), (&mut is_inexact_gt_midpoint), p_sign, z_sign, (&mut z_exp), q3, q4, (&mut e3), (&mut e4), delta, C3, C4, rnd_mode, pfpsf);
        (*ptr_is_midpoint_lt_even) = is_midpoint_lt_even;
        (*ptr_is_midpoint_gt_even) = is_midpoint_gt_even;
        (*ptr_is_inexact_lt_midpoint) = is_inexact_lt_midpoint;
        (*ptr_is_inexact_gt_midpoint) = is_inexact_gt_midpoint;
        break 'done;
    } else {
        if ((delta.wrapping_add(q4)) < q3) {
            P128.w[1] = C3.w[1];
            P128.w[0] = C3.w[0];
            C3.w[1] = C4.w[1];
            C3.w[0] = C4.w[0];
            C4.w[1] = P128.w[1];
            C4.w[0] = P128.w[0];
            ind = q3;
            q3 = q4;
            q4 = ind;
            ind = e3;
            e3 = e4;
            e4 = ind;
            tmp_sign = z_sign;
            z_sign = p_sign;
            p_sign = tmp_sign;
        } else {
            delta = (delta.wrapping_neg());
        }
        bid_add_and_round(q3, q4, e4, delta, p34, z_sign, p_sign, (*C3), C4, rnd_mode, (&mut is_midpoint_lt_even), (&mut is_midpoint_gt_even), (&mut is_inexact_lt_midpoint), (&mut is_inexact_gt_midpoint), pfpsf, res);
        (*ptr_is_midpoint_lt_even) = is_midpoint_lt_even;
        (*ptr_is_midpoint_gt_even) = is_midpoint_gt_even;
        (*ptr_is_inexact_lt_midpoint) = is_inexact_lt_midpoint;
        (*ptr_is_inexact_gt_midpoint) = is_inexact_gt_midpoint;
        break 'done;
    }
    }
    (*z_exp_ptr) = z_exp;
    (*p_exp_ptr) = p_exp;
    (*e3_ptr) = e3;
    (*e4_ptr) = e4;
}

pub(crate) fn bid_fma_delta_lt_zero(mut p34: i64, res: &mut BID_UINT128, ptr_is_midpoint_lt_even: &mut i64, ptr_is_midpoint_gt_even: &mut i64, ptr_is_inexact_lt_midpoint: &mut i64, ptr_is_inexact_gt_midpoint: &mut i64, mut p_sign: u64, mut z_sign: u64, z_exp_ptr: &mut u64, p_exp_ptr: &mut u64, mut q3: i64, mut q4: i64, e3_ptr: &mut i64, e4_ptr: &mut i64, mut delta: i64, C3: &mut BID_UINT128, mut C4: BID_UINT256, mut rnd_mode: i64, pfpsf: &mut u32) {
    let mut z_exp = (*z_exp_ptr);
    let mut p_exp = (*p_exp_ptr);
    let mut e3 = (*e3_ptr);
    let mut e4 = (*e4_ptr);
    let mut is_midpoint_lt_even: i64 = 0;
    let mut is_midpoint_gt_even: i64 = 0;
    let mut is_inexact_lt_midpoint: i64 = 0;
    let mut is_inexact_gt_midpoint: i64 = 0;
    let mut is_midpoint_lt_even0: i64 = 0;
    let mut is_midpoint_gt_even0: i64 = 0;
    let mut is_inexact_lt_midpoint0: i64 = 0;
    let mut is_inexact_gt_midpoint0: i64 = 0;
    let mut incr_exp: i64 = 0;
    let mut lsb: i64 = 0;
    let mut lt_half_ulp: i64 = 0;
    let mut eq_half_ulp: i64 = 0;
    let mut is_tiny: i64 = 0;
    let mut R64: u64 = 0;
    let mut P128: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut R128: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut P192: BID_UINT192 = BID_UINT192 { w: [0, 0, 0] };
    let mut R192: BID_UINT192 = BID_UINT192 { w: [0, 0, 0] };
    let mut R256: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut scale: i64 = 0;
    let mut ind: i64 = 0;
    let mut x0: i64 = 0;
    let mut tmp_sign: u64 = 0;
    delta = (delta.wrapping_neg());
    if ((p34 < q4) && (q4 <= delta)) {
        bid_fma_case7(p34, res, (&mut is_midpoint_lt_even), (&mut is_midpoint_gt_even), (&mut is_inexact_lt_midpoint), (&mut is_inexact_gt_midpoint), p_sign, z_sign, q3, q4, (&mut e4), delta, C3, C4, rnd_mode, pfpsf);
    } else if ((((((((q4 <= p34) && (p34 <= delta))) || ((((q4 <= delta) && (delta < p34)) && (p34 < (delta.wrapping_add(q3)))))) || (((q4 <= delta) && ((delta.wrapping_add(q3)) <= p34)))) || ((((delta < q4) && (q4 <= p34)) && (p34 < (delta.wrapping_add(q3)))))) || ((((delta < q4) && (q4 <= (delta.wrapping_add(q3)))) && ((delta.wrapping_add(q3)) <= p34)))) || ((((delta.wrapping_add(q3)) < q4) && (q4 <= p34)))) {
        P128.w[1] = C3.w[1];
        P128.w[0] = C3.w[0];
        C3.w[1] = C4.w[1];
        C3.w[0] = C4.w[0];
        C4.w[1] = P128.w[1];
        C4.w[0] = P128.w[0];
        ind = q3;
        q3 = q4;
        q4 = ind;
        ind = e3;
        e3 = e4;
        e4 = ind;
        tmp_sign = z_sign;
        z_sign = p_sign;
        p_sign = tmp_sign;
        let mut tmp64 = z_exp;
        z_exp = p_exp;
        p_exp = tmp64;
        delta = (((q3.wrapping_add(e3)).wrapping_sub(q4)).wrapping_sub(e4));
        bid_fma_delta_ge_zero(p34, res, (&mut is_midpoint_lt_even), (&mut is_midpoint_gt_even), (&mut is_inexact_lt_midpoint), (&mut is_inexact_gt_midpoint), p_sign, z_sign, (&mut z_exp), (&mut p_exp), q3, q4, (&mut e3), (&mut e4), delta, C3, C4, rnd_mode, pfpsf);
    } else if (((((p34 <= delta) && (delta < q4)) && (q4 < (delta.wrapping_add(q3))))) || ((((delta < p34) && (p34 < q4)) && (q4 < (delta.wrapping_add(q3)))))) {
        bid_fma_cases_11_12(p34, res, (&mut is_midpoint_lt_even), (&mut is_midpoint_gt_even), (&mut is_inexact_lt_midpoint), (&mut is_inexact_gt_midpoint), p_sign, z_sign, q3, q4, (&mut e3), (&mut e4), delta, C3, C4, rnd_mode, pfpsf);
    } else if (((((p34 <= delta) && ((delta.wrapping_add(q3)) <= q4))) || ((((delta < p34) && (p34 < (delta.wrapping_add(q3)))) && ((delta.wrapping_add(q3)) <= q4)))) || ((((delta.wrapping_add(q3)) <= p34) && (p34 < q4)))) {
        bid_add_and_round(q3, q4, e4, delta, p34, z_sign, p_sign, (*C3), C4, rnd_mode, (&mut is_midpoint_lt_even), (&mut is_midpoint_gt_even), (&mut is_inexact_lt_midpoint), (&mut is_inexact_gt_midpoint), pfpsf, res);
    } else {
    }
    _ = R64;
    _ = P128;
    _ = R128;
    _ = P192;
    _ = R192;
    _ = R256;
    _ = scale;
    _ = ind;
    _ = x0;
    _ = incr_exp;
    _ = lsb;
    _ = lt_half_ulp;
    _ = eq_half_ulp;
    _ = is_tiny;
    _ = is_midpoint_lt_even0;
    _ = is_midpoint_gt_even0;
    _ = is_inexact_lt_midpoint0;
    _ = is_inexact_gt_midpoint0;
    _ = tmp_sign;
    (*ptr_is_midpoint_lt_even) = is_midpoint_lt_even;
    (*ptr_is_midpoint_gt_even) = is_midpoint_gt_even;
    (*ptr_is_inexact_lt_midpoint) = is_inexact_lt_midpoint;
    (*ptr_is_inexact_gt_midpoint) = is_inexact_gt_midpoint;
    (*z_exp_ptr) = z_exp;
    (*p_exp_ptr) = p_exp;
    (*e3_ptr) = e3;
    (*e4_ptr) = e4;
}

pub(crate) fn bid_fma_case1pp_b_psign_ne_zsign(mut p34: i64, res: &mut BID_UINT128, ptr_is_midpoint_lt_even: &mut i64, ptr_is_midpoint_gt_even: &mut i64, ptr_is_inexact_lt_midpoint: &mut i64, ptr_is_inexact_gt_midpoint: &mut i64, mut p_sign: u64, mut z_sign: u64, z_exp_ptr: &mut u64, mut q3: i64, mut q4: i64, e3_ptr: &mut i64, mut scale: i64, C3: &mut BID_UINT128, mut C4: BID_UINT256, mut lt_half_ulp: i64, mut eq_half_ulp: i64, mut gt_half_ulp: i64, mut rnd_mode: i64, pfpsf: &mut u32) {
    let mut z_exp = (*z_exp_ptr);
    let mut e3 = (*e3_ptr);
    let mut is_midpoint_lt_even: i64 = 0;
    let mut is_midpoint_gt_even: i64 = 0;
    let mut is_inexact_lt_midpoint: i64 = 0;
    let mut is_inexact_gt_midpoint: i64 = 0;
    let mut incr_exp: i64 = 0;
    let mut R64: u64 = 0;
    let mut R128: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut P128: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut P192: BID_UINT192 = BID_UINT192 { w: [0, 0, 0] };
    let mut R192: BID_UINT192 = BID_UINT192 { w: [0, 0, 0] };
    let mut R256: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    if ((res.w[1] != 0x0000314dc6448d93) || (res.w[0] != 0x38c15b0a00000000)) {
        if (lt_half_ulp != 0) {
            res.w[1] = ((z_sign | (z_exp & 0x7ffe000000000000)) | res.w[1]);
            is_inexact_gt_midpoint = 1;
        } else if ((((eq_half_ulp != 0) && ((res.w[0] & 0x01) != 0))) || (gt_half_ulp != 0)) {
            res.w[0] = res.w[0].wrapping_sub(1);
            if (res.w[0] == 0xffffffffffffffff) {
                res.w[1] = res.w[1].wrapping_sub(1);
            }
            res.w[1] = ((z_sign | (z_exp & 0x7ffe000000000000)) | res.w[1]);
            if (eq_half_ulp != 0) {
                is_midpoint_gt_even = 1;
            } else {
                is_inexact_lt_midpoint = 1;
            }
        } else {
            res.w[1] = ((z_sign | (z_exp & 0x7ffe000000000000)) | res.w[1]);
            is_midpoint_lt_even = 1;
        }
        if (e3 > 0x17df) {
            if (rnd_mode == 0) {
                res.w[1] = (z_sign | 0x7800000000000000);
                res.w[0] = 0x0000000000000000;
                (*pfpsf) |= (32 | 8);
            } else {
                bid_rounding_correction(rnd_mode, is_inexact_lt_midpoint, is_inexact_gt_midpoint, is_midpoint_lt_even, is_midpoint_gt_even, e3, res, pfpsf);
            }
            (*ptr_is_midpoint_lt_even) = is_midpoint_lt_even;
            (*ptr_is_midpoint_gt_even) = is_midpoint_gt_even;
            (*ptr_is_inexact_lt_midpoint) = is_inexact_lt_midpoint;
            (*ptr_is_inexact_gt_midpoint) = is_inexact_gt_midpoint;
            (*z_exp_ptr) = z_exp;
            (*e3_ptr) = e3;
            return;
        }
        (*pfpsf) |= 32;
        if (rnd_mode != 0) {
            bid_rounding_correction(rnd_mode, is_inexact_lt_midpoint, is_inexact_gt_midpoint, is_midpoint_lt_even, is_midpoint_gt_even, e3, res, pfpsf);
        }
        z_exp = (res.w[1] & 0x7ffe000000000000);
    } else {
        e3 = (((go_checked_shr_u64(z_exp, go_shift_count_u64((49) as u64))) as i64).wrapping_sub(6176));
        if (e3 > -6176) {
            if (q4 == 1) {
                res.w[1] = 0x0001ed09bead87c0;
                res.w[0] = ((0x378d8e6400000000 as u64).wrapping_sub(C4.w[0]));
                z_exp = (z_exp.wrapping_sub(0x2000000000000));
                e3 = (e3.wrapping_sub(1));
                res.w[1] = ((z_sign | (z_exp & 0x7ffe000000000000)) | res.w[1]);
            } else {
                if (q4 <= 18) {
                    R64 = bid_round64_2_18(q4, (q4.wrapping_sub(1)), C4.w[0], (&mut incr_exp), (&mut is_midpoint_lt_even), (&mut is_midpoint_gt_even), (&mut is_inexact_lt_midpoint), (&mut is_inexact_gt_midpoint));
                } else if (q4 <= 38) {
                    P128.w[1] = C4.w[1];
                    P128.w[0] = C4.w[0];
                    (R128, incr_exp, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint) = bid_round128_19_38(q4, (q4.wrapping_sub(1)), P128);
                    R64 = R128.w[0];
                } else if (q4 <= 57) {
                    P192.w[2] = C4.w[2];
                    P192.w[1] = C4.w[1];
                    P192.w[0] = C4.w[0];
                    (R192, incr_exp, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint) = bid_round192_39_57(q4, (q4.wrapping_sub(1)), P192);
                    R64 = R192.w[0];
                } else {
                    (R256, incr_exp, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint) = bid_round256_58_76(q4, (q4.wrapping_sub(1)), C4);
                    R64 = R256.w[0];
                }
                if ((((is_midpoint_lt_even == 0) && (is_midpoint_gt_even == 0)) && (is_inexact_lt_midpoint == 0)) && (is_inexact_gt_midpoint == 0)) {
                    z_exp = (z_exp.wrapping_sub(0x2000000000000));
                    e3 = (e3.wrapping_sub(1));
                    res.w[1] = ((z_sign | (z_exp & 0x7ffe000000000000)) | 0x0001ed09bead87c0);
                    res.w[0] = ((0x378d8e6400000000 as u64).wrapping_sub(R64));
                } else {
                    if (incr_exp != 0) {
                        R64 = 10;
                    }
                    res.w[1] = 0x0001ed09bead87c0;
                    res.w[0] = ((0x378d8e6400000000 as u64).wrapping_sub(R64));
                    z_exp = (z_exp.wrapping_sub(0x2000000000000));
                    e3 = (e3.wrapping_sub(1));
                    if (is_inexact_lt_midpoint != 0) {
                        is_inexact_lt_midpoint = 0;
                        is_inexact_gt_midpoint = 1;
                    } else if (is_inexact_gt_midpoint != 0) {
                        is_inexact_gt_midpoint = 0;
                        is_inexact_lt_midpoint = 1;
                    } else if (is_midpoint_lt_even != 0) {
                        is_midpoint_lt_even = 0;
                        is_midpoint_gt_even = 1;
                    } else if (is_midpoint_gt_even != 0) {
                        is_midpoint_gt_even = 0;
                        is_midpoint_lt_even = 1;
                    }
                    if (e3 > 0x17df) {
                        if (rnd_mode == 0) {
                            res.w[1] = (z_sign | 0x7800000000000000);
                            res.w[0] = 0x0000000000000000;
                            (*pfpsf) |= (32 | 8);
                        } else {
                            bid_rounding_correction(rnd_mode, is_inexact_lt_midpoint, is_inexact_gt_midpoint, is_midpoint_lt_even, is_midpoint_gt_even, e3, res, pfpsf);
                        }
                        (*ptr_is_midpoint_lt_even) = is_midpoint_lt_even;
                        (*ptr_is_midpoint_gt_even) = is_midpoint_gt_even;
                        (*ptr_is_inexact_lt_midpoint) = is_inexact_lt_midpoint;
                        (*ptr_is_inexact_gt_midpoint) = is_inexact_gt_midpoint;
                        (*z_exp_ptr) = z_exp;
                        (*e3_ptr) = e3;
                        return;
                    }
                    (*pfpsf) |= 32;
                    res.w[1] = ((z_sign | ((go_checked_shl_u64(((e3.wrapping_add(6176)) as u64), go_shift_count_u64((49) as u64))))) | res.w[1]);
                    if (rnd_mode != 0) {
                        bid_rounding_correction(rnd_mode, is_inexact_lt_midpoint, is_inexact_gt_midpoint, is_midpoint_lt_even, is_midpoint_gt_even, e3, res, pfpsf);
                    }
                    z_exp = (res.w[1] & 0x7ffe000000000000);
                }
            }
        } else {
            if (gt_half_ulp != 0) {
                res.w[1] = 0x0000314dc6448d93;
                res.w[0] = 0x38c15b09ffffffff;
            } else {
                res.w[1] = 0x0000314dc6448d93;
                res.w[0] = 0x38c15b0a00000000;
            }
            res.w[1] = ((z_sign | (z_exp & 0x7ffe000000000000)) | res.w[1]);
            (*pfpsf) |= 16;
            if (eq_half_ulp != 0) {
                is_midpoint_lt_even = 1;
            } else if (lt_half_ulp != 0) {
                is_inexact_gt_midpoint = 1;
            } else {
                is_inexact_lt_midpoint = 1;
            }
            if (rnd_mode != 0) {
                bid_rounding_correction(rnd_mode, is_inexact_lt_midpoint, is_inexact_gt_midpoint, is_midpoint_lt_even, is_midpoint_gt_even, e3, res, pfpsf);
                z_exp = (res.w[1] & 0x7ffe000000000000);
            }
        }
        if ((((is_inexact_lt_midpoint != 0) || (is_inexact_gt_midpoint != 0)) || (is_midpoint_lt_even != 0)) || (is_midpoint_gt_even != 0)) {
            (*pfpsf) |= 32;
        }
    }
    (*ptr_is_midpoint_lt_even) = is_midpoint_lt_even;
    (*ptr_is_midpoint_gt_even) = is_midpoint_gt_even;
    (*ptr_is_inexact_lt_midpoint) = is_inexact_lt_midpoint;
    (*ptr_is_inexact_gt_midpoint) = is_inexact_gt_midpoint;
    (*z_exp_ptr) = z_exp;
    (*e3_ptr) = e3;
}

pub(crate) fn bid_fma_cases_2_to_6(mut p34: i64, res: &mut BID_UINT128, ptr_is_midpoint_lt_even: &mut i64, ptr_is_midpoint_gt_even: &mut i64, ptr_is_inexact_lt_midpoint: &mut i64, ptr_is_inexact_gt_midpoint: &mut i64, mut p_sign: u64, mut z_sign: u64, z_exp_ptr: &mut u64, mut q3: i64, mut q4: i64, e3_ptr: &mut i64, e4_ptr: &mut i64, mut delta: i64, C3: &mut BID_UINT128, mut C4: BID_UINT256, mut rnd_mode: i64, pfpsf: &mut u32) {
    let mut e3 = (*e3_ptr);
    let mut scale: i64 = 0;
    let mut x0: i64 = 0;
    let mut is_midpoint_lt_even: i64 = 0;
    let mut is_midpoint_gt_even: i64 = 0;
    let mut is_inexact_lt_midpoint: i64 = 0;
    let mut is_inexact_gt_midpoint: i64 = 0;
    let mut is_midpoint_lt_even0: i64 = 0;
    let mut is_midpoint_gt_even0: i64 = 0;
    let mut is_inexact_lt_midpoint0: i64 = 0;
    let mut is_inexact_gt_midpoint0: i64 = 0;
    let mut incr_exp: i64 = 0;
    let mut is_tiny: i64 = 0;
    let mut R128: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut P128: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut R64: u64 = 0;
    let mut P192: BID_UINT192 = BID_UINT192 { w: [0, 0, 0] };
    let mut R192: BID_UINT192 = BID_UINT192 { w: [0, 0, 0] };
    let mut R256: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut lsb: u64 = 0;
    let mut tmp64: u64 = 0;
    let mut ind: i64 = 0;
    if (((((q3 <= delta) && (delta < p34)) && (p34 < (delta.wrapping_add(q4))))) || (((delta < q3) && (p34 < (delta.wrapping_add(q4)))))) {
        scale = (p34.wrapping_sub(q3));
        x0 = ((delta.wrapping_add(q4)).wrapping_sub(p34));
    } else if ((delta.wrapping_add(q4)) < q3) {
        scale = ((q3.wrapping_sub(delta)).wrapping_sub(q4));
        if (q4 <= 19) {
            if (scale <= 19) {
                P128 = __mul_64x64_to_128(C4.w[0], bid_ten2k64[scale as usize]);
            } else {
                P128 = __mul_128x64_to_128(C4.w[0], bid_ten2k128[(scale.wrapping_sub(20)) as usize]);
            }
        } else {
            let mut C4_128: BID_UINT128 = BID_UINT128 { w: [0, 0] };
            C4_128.w[0] = C4.w[0];
            C4_128.w[1] = C4.w[1];
            P128 = __mul_128x64_to_128(bid_ten2k64[scale as usize], C4_128);
        }
        C4.w[0] = P128.w[0];
        C4.w[1] = P128.w[1];
        scale = 0;
        x0 = 0;
    } else {
        scale = ((delta.wrapping_add(q4)).wrapping_sub(q3));
        x0 = 0;
    }
    loop {
        if (scale == 0) {
            res.w[1] = C3.w[1];
            res.w[0] = C3.w[0];
        } else if (q3 <= 19) {
            if (scale <= 19) {
                (*res) = __mul_64x64_to_128(C3.w[0], bid_ten2k64[scale as usize]);
            } else {
                (*res) = __mul_128x64_to_128(C3.w[0], bid_ten2k128[(scale.wrapping_sub(20)) as usize]);
            }
        } else {
            (*res) = __mul_128x64_to_128(bid_ten2k64[scale as usize], (*C3));
        }
        e3 = (e3.wrapping_sub(scale));
        if (x0 == 0) {
            R128.w[1] = C4.w[1];
            R128.w[0] = C4.w[0];
        } else if (q4 <= 18) {
            R64 = bid_round64_2_18(q4, x0, C4.w[0], (&mut incr_exp), (&mut is_midpoint_lt_even), (&mut is_midpoint_gt_even), (&mut is_inexact_lt_midpoint), (&mut is_inexact_gt_midpoint));
            if (incr_exp != 0) {
                R64 = bid_ten2k64[(q4.wrapping_sub(x0)) as usize];
            }
            R128.w[1] = 0;
            R128.w[0] = R64;
        } else if (q4 <= 38) {
            P128.w[1] = C4.w[1];
            P128.w[0] = C4.w[0];
            (R128, incr_exp, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint) = bid_round128_19_38(q4, x0, P128);
            if (incr_exp != 0) {
                if ((q4.wrapping_sub(x0)) <= 19) {
                    R128.w[0] = bid_ten2k64[(q4.wrapping_sub(x0)) as usize];
                } else {
                    R128.w[0] = bid_ten2k128[((q4.wrapping_sub(x0)).wrapping_sub(20)) as usize].w[0];
                    R128.w[1] = bid_ten2k128[((q4.wrapping_sub(x0)).wrapping_sub(20)) as usize].w[1];
                }
            }
        } else if (q4 <= 57) {
            P192.w[2] = C4.w[2];
            P192.w[1] = C4.w[1];
            P192.w[0] = C4.w[0];
            (R192, incr_exp, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint) = bid_round192_39_57(q4, x0, P192);
            if (incr_exp != 0) {
                if ((q4.wrapping_sub(x0)) <= 19) {
                    R192.w[0] = bid_ten2k64[(q4.wrapping_sub(x0)) as usize];
                } else {
                    R192.w[0] = bid_ten2k128[((q4.wrapping_sub(x0)).wrapping_sub(20)) as usize].w[0];
                    R192.w[1] = bid_ten2k128[((q4.wrapping_sub(x0)).wrapping_sub(20)) as usize].w[1];
                }
            }
            R128.w[1] = R192.w[1];
            R128.w[0] = R192.w[0];
        } else {
            (R256, incr_exp, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint) = bid_round256_58_76(q4, x0, C4);
            if (incr_exp != 0) {
                if ((q4.wrapping_sub(x0)) <= 19) {
                    R256.w[0] = bid_ten2k64[(q4.wrapping_sub(x0)) as usize];
                } else {
                    R256.w[0] = bid_ten2k128[((q4.wrapping_sub(x0)).wrapping_sub(20)) as usize].w[0];
                    R256.w[1] = bid_ten2k128[((q4.wrapping_sub(x0)).wrapping_sub(20)) as usize].w[1];
                }
            }
            R128.w[1] = R256.w[1];
            R128.w[0] = R256.w[0];
        }
        if (z_sign == p_sign) {
            lsb = (res.w[0] & 0x01);
            res.w[0] = (res.w[0].wrapping_add(R128.w[0]));
            res.w[1] = (res.w[1].wrapping_add(R128.w[1]));
            if (res.w[0] < R128.w[0]) {
                res.w[1] = res.w[1].wrapping_add(1);
            }
            if ((res.w[1] > 0x0001ed09bead87c0) || (((res.w[1] == 0x0001ed09bead87c0) && (res.w[0] > 0x378d8e63ffffffff)))) {
                is_inexact_lt_midpoint0 = is_inexact_lt_midpoint;
                is_inexact_gt_midpoint0 = is_inexact_gt_midpoint;
                is_midpoint_lt_even0 = is_midpoint_lt_even;
                is_midpoint_gt_even0 = is_midpoint_gt_even;
                is_inexact_lt_midpoint = 0;
                is_inexact_gt_midpoint = 0;
                is_midpoint_lt_even = 0;
                is_midpoint_gt_even = 0;
                P128.w[1] = res.w[1];
                P128.w[0] = res.w[0];
                ((*res), incr_exp, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint) = bid_round128_19_38(35, 1, P128);
                _ = incr_exp;
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
                e3 = (e3.wrapping_add(1));
                if ((((is_midpoint_lt_even == 0) && (is_midpoint_gt_even == 0)) && (is_inexact_lt_midpoint == 0)) && (is_inexact_gt_midpoint == 0)) {
                    if ((((is_midpoint_lt_even0 != 0) || (is_midpoint_gt_even0 != 0)) || (is_inexact_lt_midpoint0 != 0)) || (is_inexact_gt_midpoint0 != 0)) {
                        is_inexact_lt_midpoint = 1;
                    }
                }
            } else {
                res.w[1] = (res.w[1] & 0x1ffffffffffff);
                if (lsb == 1) {
                    if (is_midpoint_gt_even != 0) {
                        is_midpoint_gt_even = 0;
                        is_midpoint_lt_even = 1;
                        res.w[0] = res.w[0].wrapping_add(1);
                        if (res.w[0] == 0x0) {
                            res.w[1] = res.w[1].wrapping_add(1);
                        }
                        if ((res.w[1] == 0x0001ed09bead87c0) && (res.w[0] == 0x378d8e6400000000)) {
                            res.w[1] = 0x0000314dc6448d93;
                            res.w[0] = 0x38c15b0a00000000;
                            e3 = e3.wrapping_add(1);
                        }
                    } else if (is_midpoint_lt_even != 0) {
                        is_midpoint_lt_even = 0;
                        is_midpoint_gt_even = 1;
                        res.w[0] = res.w[0].wrapping_sub(1);
                        if (res.w[0] == 0xffffffffffffffff) {
                            res.w[1] = res.w[1].wrapping_sub(1);
                        }
                        if ((res.w[1] == 0x0) && (res.w[0] == 0x0)) {
                            if (rnd_mode != 1) {
                                z_sign = 0x0000000000000000;
                            } else {
                                z_sign = 0x8000000000000000;
                            }
                            res.w[1] = 0x0;
                            res.w[0] = 0x0;
                            (*ptr_is_midpoint_lt_even) = is_midpoint_lt_even;
                            (*ptr_is_midpoint_gt_even) = is_midpoint_gt_even;
                            (*ptr_is_inexact_lt_midpoint) = is_inexact_lt_midpoint;
                            (*ptr_is_inexact_gt_midpoint) = is_inexact_gt_midpoint;
                            (*e3_ptr) = e3;
                            return;
                        }
                    } else {
                    }
                }
            }
        } else {
            lsb = (res.w[0] & 0x01);
            tmp64 = res.w[0];
            res.w[0] = (res.w[0].wrapping_sub(R128.w[0]));
            res.w[1] = (res.w[1].wrapping_sub(R128.w[1]));
            if (res.w[0] > tmp64) {
                res.w[1] = res.w[1].wrapping_sub(1);
            }
            if (((e3 > -6176) && (((((res.w[1] < 0x0000314dc6448d93) || (((res.w[1] == 0x0000314dc6448d93) && (res.w[0] < 0x38c15b0a00000000))))) || (((((is_inexact_lt_midpoint | is_midpoint_gt_even) != 0) && (res.w[1] == 0x0000314dc6448d93)) && (res.w[0] == 0x38c15b0a00000000)))))) && (x0 >= 1)) {
                x0 = (x0.wrapping_sub(1));
                e3 = (e3.wrapping_add(scale));
                scale = (scale.wrapping_add(1));
                is_inexact_lt_midpoint = 0;
                is_inexact_gt_midpoint = 0;
                is_midpoint_lt_even = 0;
                is_midpoint_gt_even = 0;
                incr_exp = 0;
                continue;
            }
            if (is_inexact_lt_midpoint != 0) {
                is_inexact_lt_midpoint = 0;
                is_inexact_gt_midpoint = 1;
            } else if (is_inexact_gt_midpoint != 0) {
                is_inexact_gt_midpoint = 0;
                is_inexact_lt_midpoint = 1;
            } else if (lsb == 0) {
                if (is_midpoint_lt_even != 0) {
                    is_midpoint_lt_even = 0;
                    is_midpoint_gt_even = 1;
                } else if (is_midpoint_gt_even != 0) {
                    is_midpoint_gt_even = 0;
                    is_midpoint_lt_even = 1;
                } else {
                }
            } else if (lsb == 1) {
                if (is_midpoint_lt_even != 0) {
                    res.w[0] = res.w[0].wrapping_add(1);
                    if (res.w[0] == 0x0) {
                        res.w[1] = res.w[1].wrapping_add(1);
                    }
                    if ((res.w[1] == 0x0001ed09bead87c0) && (res.w[0] == 0x378d8e6400000000)) {
                        res.w[1] = 0x0000314dc6448d93;
                        res.w[0] = 0x38c15b0a00000000;
                        e3 = e3.wrapping_add(1);
                    }
                } else if (is_midpoint_gt_even != 0) {
                    res.w[0] = res.w[0].wrapping_sub(1);
                    if (res.w[0] == 0xffffffffffffffff) {
                        res.w[1] = res.w[1].wrapping_sub(1);
                    }
                    if ((res.w[1] == 0x0) && (res.w[0] == 0x0)) {
                        if (rnd_mode != 1) {
                            z_sign = 0x0000000000000000;
                        } else {
                            z_sign = 0x8000000000000000;
                        }
                        res.w[1] = 0x0;
                        res.w[0] = 0x0;
                        (*ptr_is_midpoint_lt_even) = is_midpoint_lt_even;
                        (*ptr_is_midpoint_gt_even) = is_midpoint_gt_even;
                        (*ptr_is_inexact_lt_midpoint) = is_inexact_lt_midpoint;
                        (*ptr_is_inexact_gt_midpoint) = is_inexact_gt_midpoint;
                        (*e3_ptr) = e3;
                        return;
                    }
                } else {
                }
            } else {
            }
        }
        if (e3 == -6176) {
            if (((res.w[1] & 0x1ffffffffffff) < 0x0000314dc6448d93) || ((((res.w[1] & 0x1ffffffffffff) == 0x0000314dc6448d93) && (res.w[0] < 0x38c15b0a00000000)))) {
                is_tiny = 1;
            }
            if (((((res.w[1] & 0x7fffffffffffffff) == 0x0000314dc6448d93)) && (res.w[0] == 0x38c15b0a00000000)) && (z_sign != p_sign)) {
                is_tiny = 1;
            }
        } else if (e3 < -6176) {
            is_tiny = 1;
            x0 = ((-6176 as i64).wrapping_sub(e3));
            is_inexact_lt_midpoint0 = is_inexact_lt_midpoint;
            is_inexact_gt_midpoint0 = is_inexact_gt_midpoint;
            is_midpoint_lt_even0 = is_midpoint_lt_even;
            is_midpoint_gt_even0 = is_midpoint_gt_even;
            is_inexact_lt_midpoint = 0;
            is_inexact_gt_midpoint = 0;
            is_midpoint_lt_even = 0;
            is_midpoint_gt_even = 0;
            if (res.w[1] == 0x0) {
                ind = 1;
                while (ind <= 19) {
                    if (res.w[0] < bid_ten2k64[ind as usize]) {
                        break;
                    }
                    ind = ind.wrapping_add(1);
                }
            } else if ((res.w[1] < bid_ten2k128[0].w[1]) || (((res.w[1] == bid_ten2k128[0].w[1]) && (res.w[0] < bid_ten2k128[0].w[0])))) {
                ind = 20;
            } else {
                ind = 1;
                while (ind <= 18) {
                    if ((res.w[1] < bid_ten2k128[ind as usize].w[1]) || (((res.w[1] == bid_ten2k128[ind as usize].w[1]) && (res.w[0] < bid_ten2k128[ind as usize].w[0])))) {
                        break;
                    }
                    ind = ind.wrapping_add(1);
                }
                ind = (ind.wrapping_add(20));
            }
            if (x0 == ind) {
                res.w[1] = 0x0;
                res.w[0] = 0x1;
                is_inexact_gt_midpoint = 1;
            } else if (ind <= 18) {
                R64 = bid_round64_2_18(ind, x0, res.w[0], (&mut incr_exp), (&mut is_midpoint_lt_even), (&mut is_midpoint_gt_even), (&mut is_inexact_lt_midpoint), (&mut is_inexact_gt_midpoint));
                if (incr_exp != 0) {
                    R64 = bid_ten2k64[(ind.wrapping_sub(x0)) as usize];
                }
                res.w[1] = 0;
                res.w[0] = R64;
            } else if (ind <= 38) {
                P128.w[1] = res.w[1];
                P128.w[0] = res.w[0];
                ((*res), incr_exp, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint) = bid_round128_19_38(ind, x0, P128);
                if (incr_exp != 0) {
                    if ((ind.wrapping_sub(x0)) <= 19) {
                        res.w[0] = bid_ten2k64[(ind.wrapping_sub(x0)) as usize];
                    } else {
                        res.w[0] = bid_ten2k128[((ind.wrapping_sub(x0)).wrapping_sub(20)) as usize].w[0];
                        res.w[1] = bid_ten2k128[((ind.wrapping_sub(x0)).wrapping_sub(20)) as usize].w[1];
                    }
                }
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
            e3 = (e3.wrapping_add(x0));
            if ((((is_midpoint_lt_even == 0) && (is_midpoint_gt_even == 0)) && (is_inexact_lt_midpoint == 0)) && (is_inexact_gt_midpoint == 0)) {
                if ((((is_midpoint_lt_even0 != 0) || (is_midpoint_gt_even0 != 0)) || (is_inexact_lt_midpoint0 != 0)) || (is_inexact_gt_midpoint0 != 0)) {
                    is_inexact_lt_midpoint = 1;
                }
            }
        } else {
        }
        if ((((is_inexact_lt_midpoint != 0) || (is_inexact_gt_midpoint != 0)) || (is_midpoint_lt_even != 0)) || (is_midpoint_gt_even != 0)) {
            (*pfpsf) |= 32;
            if (is_tiny != 0) {
                (*pfpsf) |= 16;
            }
        }
        if ((res.w[1] == 0x0001ed09bead87c0) && (res.w[0] == 0x378d8e6400000000)) {
            res.w[1] = 0x0000314dc6448d93;
            res.w[0] = 0x38c15b0a00000000;
            e3 = (e3.wrapping_add(1));
        }
        res.w[1] = ((z_sign | ((go_checked_shl_u64(((e3.wrapping_add(6176)) as u64), go_shift_count_u64((49) as u64))))) | res.w[1]);
        if ((rnd_mode == 0) && (e3 > 0x17df)) {
            res.w[1] = (z_sign | 0x7800000000000000);
            res.w[0] = 0x0000000000000000;
            (*pfpsf) |= (32 | 8);
        }
        if (rnd_mode != 0) {
            bid_rounding_correction(rnd_mode, is_inexact_lt_midpoint, is_inexact_gt_midpoint, is_midpoint_lt_even, is_midpoint_gt_even, e3, res, pfpsf);
        }
        (*ptr_is_midpoint_lt_even) = is_midpoint_lt_even;
        (*ptr_is_midpoint_gt_even) = is_midpoint_gt_even;
        (*ptr_is_inexact_lt_midpoint) = is_inexact_lt_midpoint;
        (*ptr_is_inexact_gt_midpoint) = is_inexact_gt_midpoint;
        (*e3_ptr) = e3;
        break;
    }
}

pub(crate) fn bid_fma_case7(mut p34: i64, res: &mut BID_UINT128, ptr_is_midpoint_lt_even: &mut i64, ptr_is_midpoint_gt_even: &mut i64, ptr_is_inexact_lt_midpoint: &mut i64, ptr_is_inexact_gt_midpoint: &mut i64, mut p_sign: u64, mut z_sign: u64, mut q3: i64, mut q4: i64, e4_ptr: &mut i64, mut delta: i64, C3: &mut BID_UINT128, mut C4: BID_UINT256, mut rnd_mode: i64, pfpsf: &mut u32) {
    let mut e4 = (*e4_ptr);
    let mut is_midpoint_lt_even: i64 = 0;
    let mut is_midpoint_gt_even: i64 = 0;
    let mut is_inexact_lt_midpoint: i64 = 0;
    let mut is_inexact_gt_midpoint: i64 = 0;
    let mut incr_exp: i64 = 0;
    let mut P128: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut P192: BID_UINT192 = BID_UINT192 { w: [0, 0, 0] };
    let mut R192: BID_UINT192 = BID_UINT192 { w: [0, 0, 0] };
    let mut R256: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut x0: i64 = 0;
    x0 = (q4.wrapping_sub(p34));
    if (q4 <= 38) {
        P128.w[1] = C4.w[1];
        P128.w[0] = C4.w[0];
        ((*res), incr_exp, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint) = bid_round128_19_38(q4, x0, P128);
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
    if (incr_exp != 0) {
        e4 = (e4.wrapping_add(1));
    }
    if ((((is_midpoint_lt_even == 0) && (is_midpoint_gt_even == 0)) && (is_inexact_lt_midpoint == 0)) && (is_inexact_gt_midpoint == 0)) {
        if (p_sign == z_sign) {
            is_inexact_lt_midpoint = 1;
        } else {
            if ((res.w[1] != 0x0000314dc6448d93) || (res.w[0] != 0x38c15b0a00000000)) {
                is_inexact_gt_midpoint = 1;
            } else {
                if (delta > (p34.wrapping_add(1))) {
                    is_inexact_gt_midpoint = 1;
                } else {
                    if (q3 <= 19) {
                        if (C3.w[0] < bid_midpoint64[(q3.wrapping_sub(1)) as usize]) {
                            is_inexact_gt_midpoint = 1;
                        } else if (C3.w[0] == bid_midpoint64[(q3.wrapping_sub(1)) as usize]) {
                            is_midpoint_lt_even = 1;
                        } else {
                            res.w[1] = 0x0001ed09bead87c0;
                            res.w[0] = 0x378d8e63ffffffff;
                            e4 = (e4.wrapping_sub(1));
                            is_inexact_lt_midpoint = 1;
                        }
                    } else {
                        if ((C3.w[1] < bid_midpoint128[(q3.wrapping_sub(20)) as usize].w[1]) || (((C3.w[1] == bid_midpoint128[(q3.wrapping_sub(20)) as usize].w[1]) && (C3.w[0] < bid_midpoint128[(q3.wrapping_sub(20)) as usize].w[0])))) {
                            is_inexact_gt_midpoint = 1;
                        } else if ((C3.w[1] == bid_midpoint128[(q3.wrapping_sub(20)) as usize].w[1]) && (C3.w[0] == bid_midpoint128[(q3.wrapping_sub(20)) as usize].w[0])) {
                            is_midpoint_lt_even = 1;
                        } else {
                            res.w[1] = 0x0001ed09bead87c0;
                            res.w[0] = 0x378d8e63ffffffff;
                            e4 = (e4.wrapping_sub(1));
                            is_inexact_lt_midpoint = 1;
                        }
                    }
                }
            }
        }
    } else if (is_midpoint_lt_even != 0) {
        if (z_sign != p_sign) {
            res.w[0] = (res.w[0].wrapping_sub(1));
            if (res.w[0] == 0xffffffffffffffff) {
                res.w[1] = res.w[1].wrapping_sub(1);
            }
            if ((res.w[1] == 0x0000314dc6448d93) && (res.w[0] == 0x38c15b09ffffffff)) {
                res.w[1] = 0x0001ed09bead87c0;
                res.w[0] = 0x378d8e63ffffffff;
                e4 = (e4.wrapping_sub(1));
            }
            is_midpoint_lt_even = 0;
            is_inexact_lt_midpoint = 1;
        } else {
            is_midpoint_lt_even = 0;
            is_inexact_gt_midpoint = 1;
        }
    } else if (is_midpoint_gt_even != 0) {
        if (z_sign == p_sign) {
            res.w[0] = (res.w[0].wrapping_add(1));
            if (res.w[0] == 0x0000000000000000) {
                res.w[1] = res.w[1].wrapping_add(1);
            }
            is_midpoint_gt_even = 0;
            is_inexact_gt_midpoint = 1;
        } else {
            is_midpoint_gt_even = 0;
            is_inexact_lt_midpoint = 1;
        }
    }
    if ((rnd_mode == 0) && (e4 > 0x17df)) {
        res.w[1] = (p_sign | 0x7800000000000000);
        res.w[0] = 0x0000000000000000;
        (*pfpsf) |= (8 | 32);
    } else {
        let mut p_exp = (go_checked_shl_u64(((e4.wrapping_add(6176)) as u64), go_shift_count_u64((49) as u64)));
        res.w[1] = ((p_sign | (p_exp & 0x7ffe000000000000)) | res.w[1]);
    }
    if (rnd_mode != 0) {
        bid_rounding_correction(rnd_mode, is_inexact_lt_midpoint, is_inexact_gt_midpoint, is_midpoint_lt_even, is_midpoint_gt_even, e4, res, pfpsf);
    }
    if ((((is_inexact_lt_midpoint != 0) || (is_inexact_gt_midpoint != 0)) || (is_midpoint_lt_even != 0)) || (is_midpoint_gt_even != 0)) {
        (*pfpsf) |= 32;
    }
    (*ptr_is_midpoint_lt_even) = is_midpoint_lt_even;
    (*ptr_is_midpoint_gt_even) = is_midpoint_gt_even;
    (*ptr_is_inexact_lt_midpoint) = is_inexact_lt_midpoint;
    (*ptr_is_inexact_gt_midpoint) = is_inexact_gt_midpoint;
    (*e4_ptr) = e4;
}

pub(crate) fn bid_fma_cases_11_12(mut p34: i64, res: &mut BID_UINT128, ptr_is_midpoint_lt_even: &mut i64, ptr_is_midpoint_gt_even: &mut i64, ptr_is_inexact_lt_midpoint: &mut i64, ptr_is_inexact_gt_midpoint: &mut i64, mut p_sign: u64, mut z_sign: u64, mut q3: i64, mut q4: i64, e3_ptr: &mut i64, e4_ptr: &mut i64, mut delta: i64, C3: &mut BID_UINT128, mut C4: BID_UINT256, mut rnd_mode: i64, pfpsf: &mut u32) {
    let mut e3 = (*e3_ptr);
    let mut e4 = (*e4_ptr);
    let mut is_midpoint_lt_even: i64 = 0;
    let mut is_midpoint_gt_even: i64 = 0;
    let mut is_inexact_lt_midpoint: i64 = 0;
    let mut is_inexact_gt_midpoint: i64 = 0;
    let mut is_inexact_lt_midpoint0: i64 = 0;
    let mut is_inexact_gt_midpoint0: i64 = 0;
    let mut is_midpoint_lt_even0: i64 = 0;
    let mut is_midpoint_gt_even0: i64 = 0;
    let mut R64: u64 = 0;
    let mut R128: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut P128: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut P192_128: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut R192: BID_UINT192 = BID_UINT192 { w: [0, 0, 0] };
    let mut R256: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut incr_exp: i64 = 0;
    let mut ind: i64 = 0;
    let mut x0: i64 = 0;
    let mut lsb: u64 = 0;
    let mut is_tiny: i64 = 0;
    let mut lt_half_ulp: i64 = 0;
    let mut eq_half_ulp: i64 = 0;
    let mut gt_half_ulp: i64 = 0;
    x0 = (e4.wrapping_sub(e3));
    if (q3 <= 18) {
        R64 = bid_round64_2_18(q3, x0, C3.w[0], (&mut incr_exp), (&mut is_midpoint_lt_even), (&mut is_midpoint_gt_even), (&mut is_inexact_lt_midpoint), (&mut is_inexact_gt_midpoint));
        C3.w[0] = R64;
    } else if (q3 <= 38) {
        (R128, incr_exp, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint) = bid_round128_19_38(q3, x0, (*C3));
        C3.w[1] = R128.w[1];
        C3.w[0] = R128.w[0];
    }
    if (incr_exp != 0) {
        P128.w[1] = C3.w[1];
        P128.w[0] = C3.w[0];
        (*C3) = __mul_64x128_to_128(bid_ten2k64[1], P128);
    }
    e3 = (e3.wrapping_add(x0));
    R256.w[3] = 0;
    R256.w[2] = 0;
    R256.w[1] = C3.w[1];
    R256.w[0] = C3.w[0];
    if (p_sign == z_sign) {
        R256 = bid_add256(C4, R256);
    } else {
        R256 = bid_sub256(C4, R256);
        lsb = (C4.w[0] & 0x01);
        if (is_inexact_lt_midpoint != 0) {
            is_inexact_lt_midpoint = 0;
            is_inexact_gt_midpoint = 1;
        } else if (is_inexact_gt_midpoint != 0) {
            is_inexact_gt_midpoint = 0;
            is_inexact_lt_midpoint = 1;
        } else if (lsb == 0) {
            if (is_midpoint_lt_even != 0) {
                is_midpoint_lt_even = 0;
                is_midpoint_gt_even = 1;
            } else if (is_midpoint_gt_even != 0) {
                is_midpoint_gt_even = 0;
                is_midpoint_lt_even = 1;
            }
        } else if (lsb == 1) {
            if (is_midpoint_lt_even != 0) {
                R256.w[0] = R256.w[0].wrapping_add(1);
                if (R256.w[0] == 0) {
                    R256.w[1] = R256.w[1].wrapping_add(1);
                    if (R256.w[1] == 0) {
                        R256.w[2] = R256.w[2].wrapping_add(1);
                        if (R256.w[2] == 0) {
                            R256.w[3] = R256.w[3].wrapping_add(1);
                        }
                    }
                }
            } else if (is_midpoint_gt_even != 0) {
                R256.w[0] = R256.w[0].wrapping_sub(1);
                if (R256.w[0] == 0xffffffffffffffff) {
                    R256.w[1] = R256.w[1].wrapping_sub(1);
                    if (R256.w[1] == 0xffffffffffffffff) {
                        R256.w[2] = R256.w[2].wrapping_sub(1);
                        if (R256.w[2] == 0xffffffffffffffff) {
                            R256.w[3] = R256.w[3].wrapping_sub(1);
                        }
                    }
                }
            }
        }
    }
    ind = bid_bid_nr_digits256(R256);
    if (ind < p34) {
    } else if (ind == p34) {
        res.w[1] = R256.w[1];
        res.w[0] = R256.w[0];
    } else {
        x0 = (ind.wrapping_sub(p34));
        is_inexact_lt_midpoint0 = is_inexact_lt_midpoint;
        is_inexact_gt_midpoint0 = is_inexact_gt_midpoint;
        is_midpoint_lt_even0 = is_midpoint_lt_even;
        is_midpoint_gt_even0 = is_midpoint_gt_even;
        is_inexact_lt_midpoint = 0;
        is_inexact_gt_midpoint = 0;
        is_midpoint_lt_even = 0;
        is_midpoint_gt_even = 0;
        if (ind <= 38) {
            P128.w[1] = R256.w[1];
            P128.w[0] = R256.w[0];
            (R128, incr_exp, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint) = bid_round128_19_38(ind, x0, P128);
        } else if (ind <= 57) {
            let mut P192: BID_UINT192 = BID_UINT192 { w: [0, 0, 0] };
            P192.w[2] = R256.w[2];
            P192.w[1] = R256.w[1];
            P192.w[0] = R256.w[0];
            (R192, incr_exp, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint) = bid_round192_39_57(ind, x0, P192);
            R128.w[1] = R192.w[1];
            R128.w[0] = R192.w[0];
        } else {
            (R256, incr_exp, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint) = bid_round256_58_76(ind, x0, R256);
            R128.w[1] = R256.w[1];
            R128.w[0] = R256.w[0];
        }
        e4 = ((e4.wrapping_add(x0)).wrapping_add(incr_exp));
        res.w[1] = R128.w[1];
        res.w[0] = R128.w[0];
        if ((((is_inexact_gt_midpoint0 != 0) || (is_midpoint_lt_even0 != 0))) && (is_midpoint_lt_even != 0)) {
            res.w[0] = res.w[0].wrapping_sub(1);
            if (res.w[0] == 0xffffffffffffffff) {
                res.w[1] = res.w[1].wrapping_sub(1);
            }
            is_midpoint_lt_even = 0;
            is_inexact_lt_midpoint = 1;
            if ((res.w[1] == 0x0000314dc6448d93) && (res.w[0] == 0x38c15b09ffffffff)) {
                res.w[1] = 0x0001ed09bead87c0;
                res.w[0] = 0x378d8e63ffffffff;
                e4 = e4.wrapping_sub(1);
            }
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
        }
    }
    if (rnd_mode == 0) {
        if (e4 < -6176) {
            is_tiny = 1;
        }
    } else {
        P128.w[1] = ((p_sign | 0x3040000000000000) | res.w[1]);
        P128.w[0] = res.w[0];
        bid_rounding_correction(rnd_mode, is_inexact_lt_midpoint, is_inexact_gt_midpoint, is_midpoint_lt_even, is_midpoint_gt_even, 0, (&mut P128), pfpsf);
        let mut scale = (((go_checked_shr_u64((P128.w[1] & 0x7ffe000000000000), go_shift_count_u64((49) as u64))) as i64).wrapping_sub(6176));
        if ((e4.wrapping_add(scale)) < -6176) {
            is_tiny = 1;
        }
    }
    res.w[1] = ((p_sign | ((go_checked_shl_u64(((e4.wrapping_add(6176)) as u64), go_shift_count_u64((49) as u64))))) | res.w[1]);
    ind = p34;
    if ((rnd_mode == 0) && (((ind.wrapping_add(e4))) > ((p34.wrapping_add(0x17df))))) {
        res.w[1] = (p_sign | 0x7800000000000000);
        res.w[0] = 0x0000000000000000;
        (*pfpsf) |= (32 | 8);
        (*ptr_is_midpoint_lt_even) = is_midpoint_lt_even;
        (*ptr_is_midpoint_gt_even) = is_midpoint_gt_even;
        (*ptr_is_inexact_lt_midpoint) = is_inexact_lt_midpoint;
        (*ptr_is_inexact_gt_midpoint) = is_inexact_gt_midpoint;
        return;
    }
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
        if (x0 > ind) {
            is_inexact_lt_midpoint = 1;
            res.w[1] = (p_sign | 0x0000000000000000);
            res.w[0] = 0x0000000000000000;
            e4 = -6176;
        } else if (x0 == ind) {
            R128.w[1] = (res.w[1] & 0x1ffffffffffff);
            R128.w[0] = res.w[0];
            if (ind <= 19) {
                if (R128.w[0] < bid_midpoint64[(ind.wrapping_sub(1)) as usize]) {
                    lt_half_ulp = 1;
                    is_inexact_lt_midpoint = 1;
                } else if (R128.w[0] == bid_midpoint64[(ind.wrapping_sub(1)) as usize]) {
                    eq_half_ulp = 1;
                    is_midpoint_gt_even = 1;
                } else {
                    gt_half_ulp = 1;
                    is_inexact_gt_midpoint = 1;
                }
            } else {
                if ((R128.w[1] < bid_midpoint128[(ind.wrapping_sub(20)) as usize].w[1]) || (((R128.w[1] == bid_midpoint128[(ind.wrapping_sub(20)) as usize].w[1]) && (R128.w[0] < bid_midpoint128[(ind.wrapping_sub(20)) as usize].w[0])))) {
                    lt_half_ulp = 1;
                    is_inexact_lt_midpoint = 1;
                } else if ((R128.w[1] == bid_midpoint128[(ind.wrapping_sub(20)) as usize].w[1]) && (R128.w[0] == bid_midpoint128[(ind.wrapping_sub(20)) as usize].w[0])) {
                    eq_half_ulp = 1;
                    is_midpoint_gt_even = 1;
                } else {
                    gt_half_ulp = 1;
                    is_inexact_gt_midpoint = 1;
                }
            }
            _ = lt_half_ulp;
            _ = eq_half_ulp;
            if ((lt_half_ulp != 0) || (eq_half_ulp != 0)) {
                res.w[1] = 0x0000000000000000;
                res.w[0] = 0x0000000000000000;
            } else {
                res.w[1] = 0x0000000000000000;
                res.w[0] = 0x0000000000000001;
            }
            _ = gt_half_ulp;
            res.w[1] = (p_sign | res.w[1]);
            e4 = -6176;
        } else {
            if (ind <= 18) {
                R64 = bid_round64_2_18(ind, x0, res.w[0], (&mut incr_exp), (&mut is_midpoint_lt_even), (&mut is_midpoint_gt_even), (&mut is_inexact_lt_midpoint), (&mut is_inexact_gt_midpoint));
                res.w[1] = 0x0;
                res.w[0] = R64;
            } else if (ind <= 38) {
                P128.w[1] = (res.w[1] & 0x1ffffffffffff);
                P128.w[0] = res.w[0];
                (P192_128, incr_exp, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint) = bid_round128_19_38(ind, x0, P128);
                res.w[1] = P192_128.w[1];
                res.w[0] = P192_128.w[0];
            }
            e4 = (e4.wrapping_add(x0));
            if (incr_exp != 0) {
                P128.w[1] = (res.w[1] & 0x1ffffffffffff);
                P128.w[0] = res.w[0];
                (*res) = __mul_64x128_to_128(bid_ten2k64[1], P128);
            }
            res.w[1] = ((p_sign | ((go_checked_shl_u64(((e4.wrapping_add(6176)) as u64), go_shift_count_u64((49) as u64))))) | (res.w[1] & 0x1ffffffffffff));
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
            }
        }
    }
    if (rnd_mode != 0) {
        bid_rounding_correction(rnd_mode, is_inexact_lt_midpoint, is_inexact_gt_midpoint, is_midpoint_lt_even, is_midpoint_gt_even, e4, res, pfpsf);
    }
    if (((((res.w[1] & 0x7fffffffffffffff) == 0x0000314dc6448d93) && (res.w[0] == 0x38c15b0a00000000))) && (((((((rnd_mode == 0) || (rnd_mode == 4))) && (((is_midpoint_lt_even != 0) || (is_inexact_gt_midpoint != 0))))) || (((((((rnd_mode == 2) && (((res.w[1] & 0x8000000000000000) == 0)))) || (((rnd_mode == 1) && (((res.w[1] & 0x8000000000000000) != 0)))))) && (((((is_midpoint_lt_even != 0) || (is_midpoint_gt_even != 0)) || (is_inexact_lt_midpoint != 0)) || (is_inexact_gt_midpoint != 0)))))))) {
        is_tiny = 1;
    }
    if ((((is_midpoint_lt_even != 0) || (is_midpoint_gt_even != 0)) || (is_inexact_lt_midpoint != 0)) || (is_inexact_gt_midpoint != 0)) {
        (*pfpsf) |= 32;
        if (is_tiny != 0) {
            (*pfpsf) |= 16;
        }
    }
    (*ptr_is_midpoint_lt_even) = is_midpoint_lt_even;
    (*ptr_is_midpoint_gt_even) = is_midpoint_gt_even;
    (*ptr_is_inexact_lt_midpoint) = is_inexact_lt_midpoint;
    (*ptr_is_inexact_gt_midpoint) = is_inexact_gt_midpoint;
    (*e4_ptr) = e4;
}

pub(crate) fn bid_add_and_round(mut q3: i64, mut q4: i64, mut e4: i64, mut delta: i64, mut p34: i64, mut z_sign: u64, mut p_sign: u64, mut C3: BID_UINT128, mut C4: BID_UINT256, mut rnd_mode: i64, ptr_is_midpoint_lt_even: &mut i64, ptr_is_midpoint_gt_even: &mut i64, ptr_is_inexact_lt_midpoint: &mut i64, ptr_is_inexact_gt_midpoint: &mut i64, ptrfpsf: &mut u32, ptrres: &mut BID_UINT128) {
    let mut scale: i64 = 0;
    let mut x0: i64 = 0;
    let mut ind: i64 = 0;
    let mut R64: u64 = 0;
    let mut P128: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut R128: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut P192: BID_UINT192 = BID_UINT192 { w: [0, 0, 0] };
    let mut R192: BID_UINT192 = BID_UINT192 { w: [0, 0, 0] };
    let mut R256: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut is_midpoint_lt_even: i64 = 0;
    let mut is_midpoint_gt_even: i64 = 0;
    let mut is_inexact_lt_midpoint: i64 = 0;
    let mut is_inexact_gt_midpoint: i64 = 0;
    let mut is_midpoint_lt_even0: i64 = 0;
    let mut is_midpoint_gt_even0: i64 = 0;
    let mut is_inexact_lt_midpoint0: i64 = 0;
    let mut is_inexact_gt_midpoint0: i64 = 0;
    let mut incr_exp: i64 = 0;
    let mut is_tiny: i64 = 0;
    let mut lt_half_ulp: i64 = 0;
    let mut eq_half_ulp: i64 = 0;
    let mut res = (*ptrres);
    scale = ((q4.wrapping_sub(delta)).wrapping_sub(q3));
    if (scale == 0) {
        R256.w[3] = 0x0;
        R256.w[2] = 0x0;
        R256.w[1] = C3.w[1];
        R256.w[0] = C3.w[0];
    } else if (scale <= 19) {
        P128.w[1] = 0;
        P128.w[0] = bid_ten2k64[scale as usize];
        R256 = __mul_128x128_to_256(P128, C3);
    } else if (scale <= 38) {
        R256 = __mul_128x128_to_256(bid_ten2k128[(scale.wrapping_sub(20)) as usize], C3);
    } else if (scale <= 57) {
        R128 = __mul_64x128_to_128(bid_ten2k64[(scale.wrapping_sub(38)) as usize], C3);
        R256 = __mul_128x128_to_256(R128, bid_ten2k128[18]);
    } else {
        R128 = __mul_64x128_to_128(C3.w[0], bid_ten2k128[(scale.wrapping_sub(58)) as usize]);
        R256 = __mul_128x128_to_256(R128, bid_ten2k128[18]);
    }
    if (p_sign == z_sign) {
        R256 = bid_add256(C4, R256);
    } else {
        if ((((R256.w[3] > C4.w[3]) || (((R256.w[3] == C4.w[3]) && (R256.w[2] > C4.w[2])))) || ((((R256.w[3] == C4.w[3]) && (R256.w[2] == C4.w[2])) && (R256.w[1] > C4.w[1])))) || (((((R256.w[3] == C4.w[3]) && (R256.w[2] == C4.w[2])) && (R256.w[1] == C4.w[1])) && (R256.w[0] >= C4.w[0])))) {
            R256 = bid_sub256(R256, C4);
            p_sign = z_sign;
        } else {
            R256 = bid_sub256(C4, R256);
        }
        if ((((R256.w[3] == 0x0) && (R256.w[2] == 0x0)) && (R256.w[1] == 0x0)) && (R256.w[0] == 0x0)) {
            if (rnd_mode != 1) {
                p_sign = 0x0000000000000000;
            } else {
                p_sign = 0x8000000000000000;
            }
            if (e4 < (-6176)) {
                e4 = -6176;
            }
            res.w[1] = (p_sign | ((go_checked_shl_u64(((e4.wrapping_add(6176)) as u64), go_shift_count_u64((49) as u64)))));
            res.w[0] = 0x0;
            (*ptrres) = res;
            return;
        }
    }
    ind = bid_bid_nr_digits256(R256);
    if (ind <= p34) {
        if ((ind.wrapping_add(e4)) < (p34.wrapping_add(-6176))) {
            is_tiny = 1;
        }
        res.w[1] = ((p_sign | ((go_checked_shl_u64(((e4.wrapping_add(6176)) as u64), go_shift_count_u64((49) as u64))))) | R256.w[1]);
        res.w[0] = R256.w[0];
    } else {
        x0 = (ind.wrapping_sub(p34));
        if (ind <= 38) {
            P128.w[1] = R256.w[1];
            P128.w[0] = R256.w[0];
            (R128, incr_exp, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint) = bid_round128_19_38(ind, x0, P128);
        } else if (ind <= 57) {
            P192.w[2] = R256.w[2];
            P192.w[1] = R256.w[1];
            P192.w[0] = R256.w[0];
            (R192, incr_exp, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint) = bid_round192_39_57(ind, x0, P192);
            R128.w[1] = R192.w[1];
            R128.w[0] = R192.w[0];
        } else {
            (R256, incr_exp, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint) = bid_round256_58_76(ind, x0, R256);
            R128.w[1] = R256.w[1];
            R128.w[0] = R256.w[0];
        }
        if ((e4.wrapping_add(x0)) < -6176) {
            is_tiny = 1;
        }
        e4 = ((e4.wrapping_add(x0)).wrapping_add(incr_exp));
        if (rnd_mode == 0) {
        } else {
            P128.w[1] = ((p_sign | 0x3040000000000000) | R128.w[1]);
            P128.w[0] = R128.w[0];
            bid_rounding_correction(rnd_mode, is_inexact_lt_midpoint, is_inexact_gt_midpoint, is_midpoint_lt_even, is_midpoint_gt_even, 0, (&mut P128), ptrfpsf);
            scale = (((go_checked_shr_u64((P128.w[1] & 0x7ffe000000000000), go_shift_count_u64((49) as u64))) as i64).wrapping_sub(6176));
        }
        ind = p34;
        res.w[1] = ((p_sign | ((go_checked_shl_u64(((e4.wrapping_add(6176)) as u64), go_shift_count_u64((49) as u64))))) | R128.w[1]);
        res.w[0] = R128.w[0];
    }
    if ((rnd_mode == 0) && (((ind.wrapping_add(e4))) > ((p34.wrapping_add(0x17df))))) {
        res.w[1] = (p_sign | 0x7800000000000000);
        res.w[0] = 0x0000000000000000;
        (*ptrres) = res;
        (*ptrfpsf) |= (32 | 8);
        return;
    }
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
        if (x0 > ind) {
            is_inexact_lt_midpoint = 1;
            res.w[1] = (p_sign | 0x0000000000000000);
            res.w[0] = 0x0000000000000000;
            e4 = -6176;
        } else if (x0 == ind) {
            R128.w[1] = (res.w[1] & 0x1ffffffffffff);
            R128.w[0] = res.w[0];
            if (ind <= 19) {
                if (R128.w[0] < bid_midpoint64[(ind.wrapping_sub(1)) as usize]) {
                    lt_half_ulp = 1;
                    is_inexact_lt_midpoint = 1;
                } else if (R128.w[0] == bid_midpoint64[(ind.wrapping_sub(1)) as usize]) {
                    eq_half_ulp = 1;
                    is_midpoint_gt_even = 1;
                } else {
                    is_inexact_gt_midpoint = 1;
                }
            } else {
                if ((R128.w[1] < bid_midpoint128[(ind.wrapping_sub(20)) as usize].w[1]) || (((R128.w[1] == bid_midpoint128[(ind.wrapping_sub(20)) as usize].w[1]) && (R128.w[0] < bid_midpoint128[(ind.wrapping_sub(20)) as usize].w[0])))) {
                    lt_half_ulp = 1;
                    is_inexact_lt_midpoint = 1;
                } else if ((R128.w[1] == bid_midpoint128[(ind.wrapping_sub(20)) as usize].w[1]) && (R128.w[0] == bid_midpoint128[(ind.wrapping_sub(20)) as usize].w[0])) {
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
            res.w[1] = (p_sign | res.w[1]);
            e4 = -6176;
        } else {
            if (ind <= 18) {
                R64 = bid_round64_2_18(ind, x0, res.w[0], (&mut incr_exp), (&mut is_midpoint_lt_even), (&mut is_midpoint_gt_even), (&mut is_inexact_lt_midpoint), (&mut is_inexact_gt_midpoint));
                res.w[1] = 0x0;
                res.w[0] = R64;
            } else if (ind <= 38) {
                P128.w[1] = (res.w[1] & 0x1ffffffffffff);
                P128.w[0] = res.w[0];
                (res, incr_exp, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint) = bid_round128_19_38(ind, x0, P128);
            }
            e4 = (e4.wrapping_add(x0));
            if (incr_exp != 0) {
                P128.w[1] = (res.w[1] & 0x1ffffffffffff);
                P128.w[0] = res.w[0];
                res = __mul_64x128_to_128(bid_ten2k64[1], P128);
            }
            res.w[1] = ((p_sign | ((go_checked_shl_u64(((e4.wrapping_add(6176)) as u64), go_shift_count_u64((49) as u64))))) | (res.w[1] & 0x1ffffffffffff));
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
            }
        }
    }
    if (rnd_mode != 0) {
        bid_rounding_correction(rnd_mode, is_inexact_lt_midpoint, is_inexact_gt_midpoint, is_midpoint_lt_even, is_midpoint_gt_even, e4, (&mut res), ptrfpsf);
    }
    if ((((is_midpoint_lt_even != 0) || (is_midpoint_gt_even != 0)) || (is_inexact_lt_midpoint != 0)) || (is_inexact_gt_midpoint != 0)) {
        (*ptrfpsf) |= 32;
        if (is_tiny != 0) {
            (*ptrfpsf) |= 16;
        }
    }
    (*ptr_is_midpoint_lt_even) = is_midpoint_lt_even;
    (*ptr_is_midpoint_gt_even) = is_midpoint_gt_even;
    (*ptr_is_inexact_lt_midpoint) = is_inexact_lt_midpoint;
    (*ptr_is_inexact_gt_midpoint) = is_inexact_gt_midpoint;
    (*ptrres) = res;
}

pub(crate) fn bid_bid_nr_digits256(mut R256: BID_UINT256) -> i64 {
    let mut ind: i64 = 0;
    if (((R256.w[3] == 0x0) && (R256.w[2] == 0x0)) && (R256.w[1] == 0x0)) {
        ind = 1;
        while (ind <= 19) {
            if (R256.w[0] < bid_ten2k64[ind as usize]) {
                break;
            }
            ind = ind.wrapping_add(1);
        }
    } else if (((R256.w[3] == 0x0) && (R256.w[2] == 0x0)) && (((R256.w[1] < bid_ten2k128[0].w[1]) || (((R256.w[1] == bid_ten2k128[0].w[1]) && (R256.w[0] < bid_ten2k128[0].w[0])))))) {
        ind = 20;
    } else if ((R256.w[3] == 0x0) && (R256.w[2] == 0x0)) {
        ind = 1;
        while (ind <= 18) {
            if ((R256.w[1] < bid_ten2k128[ind as usize].w[1]) || (((R256.w[1] == bid_ten2k128[ind as usize].w[1]) && (R256.w[0] < bid_ten2k128[ind as usize].w[0])))) {
                break;
            }
            ind = ind.wrapping_add(1);
        }
        ind = (ind.wrapping_add(20));
    } else if ((R256.w[3] == 0x0) && ((((R256.w[2] < bid_ten2k256[0].w[2]) || (((R256.w[2] == bid_ten2k256[0].w[2]) && (R256.w[1] < bid_ten2k256[0].w[1])))) || ((((R256.w[2] == bid_ten2k256[0].w[2]) && (R256.w[1] == bid_ten2k256[0].w[1])) && (R256.w[0] < bid_ten2k256[0].w[0])))))) {
        ind = 39;
    } else {
        ind = 1;
        while (ind <= 29) {
            if ((((R256.w[3] < bid_ten2k256[ind as usize].w[3]) || (((R256.w[3] == bid_ten2k256[ind as usize].w[3]) && (R256.w[2] < bid_ten2k256[ind as usize].w[2])))) || ((((R256.w[3] == bid_ten2k256[ind as usize].w[3]) && (R256.w[2] == bid_ten2k256[ind as usize].w[2])) && (R256.w[1] < bid_ten2k256[ind as usize].w[1])))) || (((((R256.w[3] == bid_ten2k256[ind as usize].w[3]) && (R256.w[2] == bid_ten2k256[ind as usize].w[2])) && (R256.w[1] == bid_ten2k256[ind as usize].w[1])) && (R256.w[0] < bid_ten2k256[ind as usize].w[0])))) {
                break;
            }
            ind = ind.wrapping_add(1);
        }
        ind = (ind.wrapping_add(39));
    }
    return ind;
}
