// Auto-generated from convert64.go by go2rs. Do not edit.

use super::prelude::*;

pub fn bid64_from_int32(mut x: i32) -> u64 {
    let mut res: u64 = 0;
    if ((((x as u32) & 0x80000000)) == 0x80000000) {
        x = ((!x).wrapping_add(1));
        res = (((x as u32) as u64) | 0xb1c0000000000000);
    } else {
        res = ((x as u64) | 0x31c0000000000000);
    }
    return res;
}

pub fn bid64_from_uint32(mut x: u32) -> u64 {
    let mut res = ((x as u64) | 0x31c0000000000000);
    return res;
}

pub fn bid64_from_int64(mut x: i64, mut rndMode: i64) -> (u64, u32) {
    let mut res: u64 = 0;
    let mut pfpsf: u32 = 0;
    let mut C: u64 = 0;
    let mut incr_exp: i64 = 0;
    let mut is_midpoint_lt_even: i64 = 0;
    let mut is_midpoint_gt_even: i64 = 0;
    let mut is_inexact_lt_midpoint: i64 = 0;
    let mut is_inexact_gt_midpoint: i64 = 0;
    let mut x_sign = ((x as u64) & 0x8000000000000000);
    if (x_sign != 0) {
        C = ((!(x as u64)).wrapping_add(1));
    } else {
        C = (x as u64);
    }
    if (C <= 0x2386f26fc0ffff) {
        if (C < 0x0020000000000000) {
            res = ((x_sign | 0x31c0000000000000) | C);
        } else {
            res = ((x_sign | 0x6c70000000000000) | (C & 0x0007ffffffffffff));
        }
    } else {
        let mut q: u32 = 0;
        let mut ind: u32 = 0;
        if (C < 0x16345785d8a0000) {
            q = 17;
            ind = 1;
        } else if (C < 0xde0b6b3a7640000) {
            q = 18;
            ind = 2;
        } else {
            q = 19;
            ind = 3;
        }
        res = bid_round64_2_18((q as i64), (ind as i64), C, (&mut incr_exp), (&mut is_midpoint_lt_even), (&mut is_midpoint_gt_even), (&mut is_inexact_lt_midpoint), (&mut is_inexact_gt_midpoint));
        if (incr_exp != 0) {
            ind = ind.wrapping_add(1);
        }
        if ((((is_inexact_lt_midpoint != 0) || (is_inexact_gt_midpoint != 0)) || (is_midpoint_lt_even != 0)) || (is_midpoint_gt_even != 0)) {
            pfpsf |= 32;
        }
        if (rndMode != 0) {
            if ((((x_sign == 0) && (((((rndMode == 2) && (is_inexact_lt_midpoint != 0))) || (((((rndMode == 4) || (rndMode == 2))) && (is_midpoint_gt_even != 0))))))) || (((x_sign != 0) && (((((rndMode == 1) && (is_inexact_lt_midpoint != 0))) || (((((rndMode == 4) || (rndMode == 1))) && (is_midpoint_gt_even != 0)))))))) {
                res = (res.wrapping_add(1));
                if (res == 0x002386f26fc10000) {
                    res = 0x00038d7ea4c68000;
                    ind = (ind.wrapping_add(1));
                }
            } else if ((((is_midpoint_lt_even != 0) || (is_inexact_gt_midpoint != 0))) && (((((x_sign != 0) && (((rndMode == 2) || (rndMode == 3))))) || (((x_sign == 0) && (((rndMode == 1) || (rndMode == 3)))))))) {
                res = (res.wrapping_sub(1));
                if (res == 0x00038d7ea4c67fff) {
                    res = 0x002386f26fc0ffff;
                    ind = (ind.wrapping_sub(1));
                }
            }
        }
        if (res < 0x0020000000000000) {
            res = ((x_sign | ((go_checked_shl_u64(((ind.wrapping_add(398)) as u64), go_shift_count_u64((53) as u64))))) | res);
        } else {
            res = (((x_sign | 0x6000000000000000) | ((go_checked_shl_u64(((ind.wrapping_add(398)) as u64), go_shift_count_u64((51) as u64))))) | (res & 0x0007ffffffffffff));
        }
    }
    return (res, pfpsf);
}

pub fn bid64_from_uint64(mut x: u64, mut rndMode: i64) -> (u64, u32) {
    let mut res: u64 = 0;
    let mut pfpsf: u32 = 0;
    let mut incr_exp: i64 = 0;
    let mut is_midpoint_lt_even: i64 = 0;
    let mut is_midpoint_gt_even: i64 = 0;
    let mut is_inexact_lt_midpoint: i64 = 0;
    let mut is_inexact_gt_midpoint: i64 = 0;
    if (x <= 0x2386f26fc0ffff) {
        if (x < 0x0020000000000000) {
            res = (0x31c0000000000000 | x);
        } else {
            res = (0x6c70000000000000 | (x & 0x0007ffffffffffff));
        }
    } else {
        let mut q: u32 = 0;
        let mut ind: u32 = 0;
        if (x < 0x16345785d8a0000) {
            q = 17;
            ind = 1;
        } else if (x < 0xde0b6b3a7640000) {
            q = 18;
            ind = 2;
        } else if (x < 0x8ac7230489e80000) {
            q = 19;
            ind = 3;
        } else {
            q = 20;
            ind = 4;
        }
        if (q <= 19) {
            res = bid_round64_2_18((q as i64), (ind as i64), x, (&mut incr_exp), (&mut is_midpoint_lt_even), (&mut is_midpoint_gt_even), (&mut is_inexact_lt_midpoint), (&mut is_inexact_gt_midpoint));
        } else {
            res = bid_round128_19_38_for64((q as i64), (ind as i64), x, (&mut incr_exp), (&mut is_midpoint_lt_even), (&mut is_midpoint_gt_even), (&mut is_inexact_lt_midpoint), (&mut is_inexact_gt_midpoint));
        }
        if (incr_exp != 0) {
            ind = ind.wrapping_add(1);
        }
        if ((((is_inexact_lt_midpoint != 0) || (is_inexact_gt_midpoint != 0)) || (is_midpoint_lt_even != 0)) || (is_midpoint_gt_even != 0)) {
            pfpsf |= 32;
        }
        if (rndMode != 0) {
            if ((((rndMode == 2) && (is_inexact_lt_midpoint != 0))) || (((((rndMode == 4) || (rndMode == 2))) && (is_midpoint_gt_even != 0)))) {
                res = (res.wrapping_add(1));
                if (res == 0x002386f26fc10000) {
                    res = 0x00038d7ea4c68000;
                    ind = (ind.wrapping_add(1));
                }
            } else if ((((is_midpoint_lt_even != 0) || (is_inexact_gt_midpoint != 0))) && (((rndMode == 1) || (rndMode == 3)))) {
                res = (res.wrapping_sub(1));
                if (res == 0x00038d7ea4c67fff) {
                    res = 0x002386f26fc0ffff;
                    ind = (ind.wrapping_sub(1));
                }
            }
        }
        if (res < 0x0020000000000000) {
            res = (((go_checked_shl_u64(((ind.wrapping_add(398)) as u64), go_shift_count_u64((53) as u64)))) | res);
        } else {
            res = ((0x6000000000000000 | ((go_checked_shl_u64(((ind.wrapping_add(398)) as u64), go_shift_count_u64((51) as u64))))) | (res & 0x0007ffffffffffff));
        }
    }
    return (res, pfpsf);
}

pub(crate) fn bid_round64_2_18(mut q: i64, mut x: i64, mut C: u64, incr_exp: &mut i64, is_midpoint_lt_even: &mut i64, is_midpoint_gt_even: &mut i64, is_inexact_lt_midpoint: &mut i64, is_inexact_gt_midpoint: &mut i64) -> u64 {
    let mut P128: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut fstar: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut Cstar: u64 = 0;
    let mut tmp64: u64 = 0;
    let mut shift: i64 = 0;
    let mut ind: i64 = 0;
    ind = (x.wrapping_sub(1));
    C = (C.wrapping_add(bid_midpoint64[ind as usize]));
    P128 = __mul_64x64_to_128(C, bid_Kx64[ind as usize]);
    shift = (bid_Ex64m64[ind as usize] as i64);
    Cstar = (go_checked_shr_u64(P128.w[1], go_shift_count_i64((shift) as i64)));
    fstar.w[1] = (P128.w[1] & bid_mask64[ind as usize]);
    fstar.w[0] = P128.w[0];
    if ((fstar.w[1] > bid_half64[ind as usize]) || (((fstar.w[1] == bid_half64[ind as usize]) && (fstar.w[0] != 0)))) {
        tmp64 = (fstar.w[1].wrapping_sub(bid_half64[ind as usize]));
        if ((tmp64 != 0) || (fstar.w[0] > bid_ten2mxtrunc64[ind as usize])) {
            (*is_inexact_lt_midpoint) = 1;
        }
    } else {
        (*is_inexact_gt_midpoint) = 1;
    }
    if ((fstar.w[1] == 0) && (fstar.w[0] <= bid_ten2mxtrunc64[ind as usize])) {
        if ((Cstar & 0x01) != 0) {
            Cstar = Cstar.wrapping_sub(1);
            (*is_midpoint_gt_even) = 1;
            (*is_inexact_lt_midpoint) = 0;
            (*is_inexact_gt_midpoint) = 0;
        } else {
            (*is_midpoint_lt_even) = 1;
            (*is_inexact_lt_midpoint) = 0;
            (*is_inexact_gt_midpoint) = 0;
        }
    }
    ind = (q.wrapping_sub(x));
    if (Cstar == bid_ten2k64[ind as usize]) {
        Cstar = bid_ten2k64[(ind.wrapping_sub(1)) as usize];
        (*incr_exp) = 1;
    } else {
        (*incr_exp) = 0;
    }
    return Cstar;
}

pub(crate) fn bid_round128_19_38_for64(mut q: i64, mut x: i64, mut C: u64, incr_exp: &mut i64, is_midpoint_lt_even: &mut i64, is_midpoint_gt_even: &mut i64, is_inexact_lt_midpoint: &mut i64, is_inexact_gt_midpoint: &mut i64) -> u64 {
    let mut P256: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut fstar: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut Cstar: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut C128: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut tmp64: u64 = 0;
    let mut shift: i64 = 0;
    let mut ind: i64 = 0;
    (*incr_exp) = 0;
    (*is_midpoint_lt_even) = 0;
    (*is_midpoint_gt_even) = 0;
    (*is_inexact_lt_midpoint) = 0;
    (*is_inexact_gt_midpoint) = 0;
    ind = (x.wrapping_sub(1));
    if ((ind < 0) || (ind > 18)) {
        return 0;
    }
    C128.w[0] = C;
    C128.w[1] = 0;
    tmp64 = C128.w[0];
    C128.w[0] = (C128.w[0].wrapping_add(bid_midpoint64[ind as usize]));
    if (C128.w[0] < tmp64) {
        C128.w[1] = C128.w[1].wrapping_add(1);
    }
    P256 = __mul_128x128_to_256(C128, bid_Kx128_for64[ind as usize]);
    shift = (bid_Ex128m128_for64[ind as usize] as i64);
    Cstar.w[0] = (((go_checked_shr_u64(P256.w[2], go_shift_count_i64((shift) as i64)))) | ((go_checked_shl_u64(P256.w[3], go_shift_count_i64(((((64 as i64).wrapping_sub(shift)))) as i64)))));
    Cstar.w[1] = (go_checked_shr_u64(P256.w[3], go_shift_count_i64((shift) as i64)));
    fstar.w[0] = P256.w[0];
    fstar.w[1] = P256.w[1];
    fstar.w[2] = (P256.w[2] & bid_mask128_for64[ind as usize]);
    fstar.w[3] = 0;
    if ((fstar.w[2] > bid_half128_for64[ind as usize]) || (((fstar.w[2] == bid_half128_for64[ind as usize]) && (((fstar.w[1] != 0) || (fstar.w[0] != 0)))))) {
        tmp64 = (fstar.w[2].wrapping_sub(bid_half128_for64[ind as usize]));
        if (((tmp64 != 0) || (fstar.w[1] > bid_ten2mxtrunc128_for64[ind as usize].w[1])) || (((fstar.w[1] == bid_ten2mxtrunc128_for64[ind as usize].w[1]) && (fstar.w[0] > bid_ten2mxtrunc128_for64[ind as usize].w[0])))) {
            (*is_inexact_lt_midpoint) = 1;
        }
    } else {
        (*is_inexact_gt_midpoint) = 1;
    }
    if (((fstar.w[3] == 0) && (fstar.w[2] == 0)) && (((fstar.w[1] < bid_ten2mxtrunc128_for64[ind as usize].w[1]) || (((fstar.w[1] == bid_ten2mxtrunc128_for64[ind as usize].w[1]) && (fstar.w[0] <= bid_ten2mxtrunc128_for64[ind as usize].w[0])))))) {
        if ((Cstar.w[0] & 0x01) != 0) {
            Cstar.w[0] = Cstar.w[0].wrapping_sub(1);
            if (Cstar.w[0] == 0xffffffffffffffff) {
                Cstar.w[1] = Cstar.w[1].wrapping_sub(1);
            }
            (*is_midpoint_gt_even) = 1;
            (*is_inexact_lt_midpoint) = 0;
            (*is_inexact_gt_midpoint) = 0;
        } else {
            (*is_midpoint_lt_even) = 1;
            (*is_inexact_lt_midpoint) = 0;
            (*is_inexact_gt_midpoint) = 0;
        }
    }
    ind = (q.wrapping_sub(x));
    if (ind <= 19) {
        if ((Cstar.w[1] == 0x0) && (Cstar.w[0] == bid_ten2k64[ind as usize])) {
            Cstar.w[0] = bid_ten2k64[(ind.wrapping_sub(1)) as usize];
            (*incr_exp) = 1;
        } else {
            (*incr_exp) = 0;
        }
    } else if (ind == 20) {
        if ((Cstar.w[1] == 0x0000000000000005) && (Cstar.w[0] == 0x6bc75e2d63100000)) {
            Cstar.w[0] = bid_ten2k64[19];
            Cstar.w[1] = 0x0;
            (*incr_exp) = 1;
        } else {
            (*incr_exp) = 0;
        }
    } else {
        (*incr_exp) = 0;
    }
    return Cstar.w[0];
}
