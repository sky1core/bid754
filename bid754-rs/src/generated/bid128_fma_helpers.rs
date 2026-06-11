// Auto-generated from bid128_fma_helpers.go by go2rs. Do not edit.

use super::prelude::*;

pub(crate) fn bid_add256(mut x: BID_UINT256, mut y: BID_UINT256) -> BID_UINT256 {
    let mut z: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    z.w[0] = (x.w[0].wrapping_add(y.w[0]));
    if (z.w[0] < x.w[0]) {
        x.w[1] = x.w[1].wrapping_add(1);
        if (x.w[1] == 0) {
            x.w[2] = x.w[2].wrapping_add(1);
            if (x.w[2] == 0) {
                x.w[3] = x.w[3].wrapping_add(1);
            }
        }
    }
    z.w[1] = (x.w[1].wrapping_add(y.w[1]));
    if (z.w[1] < x.w[1]) {
        x.w[2] = x.w[2].wrapping_add(1);
        if (x.w[2] == 0) {
            x.w[3] = x.w[3].wrapping_add(1);
        }
    }
    z.w[2] = (x.w[2].wrapping_add(y.w[2]));
    if (z.w[2] < x.w[2]) {
        x.w[3] = x.w[3].wrapping_add(1);
    }
    z.w[3] = (x.w[3].wrapping_add(y.w[3]));
    return z;
}

pub(crate) fn bid_sub256(mut x: BID_UINT256, mut y: BID_UINT256) -> BID_UINT256 {
    let mut z: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    z.w[0] = (x.w[0].wrapping_sub(y.w[0]));
    if (z.w[0] > x.w[0]) {
        x.w[1] = x.w[1].wrapping_sub(1);
        if (x.w[1] == 0xffffffffffffffff) {
            x.w[2] = x.w[2].wrapping_sub(1);
            if (x.w[2] == 0xffffffffffffffff) {
                x.w[3] = x.w[3].wrapping_sub(1);
            }
        }
    }
    z.w[1] = (x.w[1].wrapping_sub(y.w[1]));
    if (z.w[1] > x.w[1]) {
        x.w[2] = x.w[2].wrapping_sub(1);
        if (x.w[2] == 0xffffffffffffffff) {
            x.w[3] = x.w[3].wrapping_sub(1);
        }
    }
    z.w[2] = (x.w[2].wrapping_sub(y.w[2]));
    if (z.w[2] > x.w[2]) {
        x.w[3] = x.w[3].wrapping_sub(1);
    }
    z.w[3] = (x.w[3].wrapping_sub(y.w[3]));
    return z;
}

pub(crate) fn bid_rounding_correction(mut rnd_mode: i64, mut is_inexact_lt_midpoint: i64, mut is_inexact_gt_midpoint: i64, mut is_midpoint_lt_even: i64, mut is_midpoint_gt_even: i64, mut unbexp: i64, ptrres: &mut BID_UINT128, ptrfpsf: &mut u32) {
    let mut res = (*ptrres);
    let mut sign: u64 = 0;
    let mut exp: u64 = 0;
    let mut C_hi: u64 = 0;
    let mut C_lo: u64 = 0;
    if ((((is_inexact_lt_midpoint != 0) || (is_inexact_gt_midpoint != 0)) || (is_midpoint_lt_even != 0)) || (is_midpoint_gt_even != 0)) {
        (*ptrfpsf) |= 32;
    }
    sign = (res.w[1] & 0x8000000000000000);
    exp = (go_checked_shl_u64(((unbexp.wrapping_add(6176)) as u64), go_shift_count_u64((49) as u64)));
    C_hi = (res.w[1] & 0x1ffffffffffff);
    C_lo = res.w[0];
    if ((((sign == 0) && (((((rnd_mode == 2) && (is_inexact_lt_midpoint != 0))) || (((((rnd_mode == 4) || (rnd_mode == 2))) && (is_midpoint_gt_even != 0))))))) || (((sign != 0) && (((((rnd_mode == 1) && (is_inexact_lt_midpoint != 0))) || (((((rnd_mode == 4) || (rnd_mode == 1))) && (is_midpoint_gt_even != 0)))))))) {
        C_lo = (C_lo.wrapping_add(1));
        if (C_lo == 0) {
            C_hi = (C_hi.wrapping_add(1));
        }
        if ((C_hi == 0x0001ed09bead87c0) && (C_lo == 0x378d8e6400000000)) {
            C_hi = 0x0000314dc6448d93;
            C_lo = 0x38c15b0a00000000;
            unbexp = (unbexp.wrapping_add(1));
            exp = (go_checked_shl_u64(((unbexp.wrapping_add(6176)) as u64), go_shift_count_u64((49) as u64)));
        }
    } else if ((((is_midpoint_lt_even != 0) || (is_inexact_gt_midpoint != 0))) && (((((sign != 0) && (((rnd_mode == 2) || (rnd_mode == 3))))) || (((sign == 0) && (((rnd_mode == 1) || (rnd_mode == 3)))))))) {
        C_lo = (C_lo.wrapping_sub(1));
        if (C_lo == 0xffffffffffffffff) {
            C_hi = C_hi.wrapping_sub(1);
        }
        if ((C_hi == 0x0000314dc6448d93) && (C_lo == 0x38c15b09ffffffff)) {
            if (exp > 0) {
                C_hi = 0x0001ed09bead87c0;
                C_lo = 0x378d8e63ffffffff;
                unbexp = (unbexp.wrapping_sub(1));
                exp = (go_checked_shl_u64(((unbexp.wrapping_add(6176)) as u64), go_shift_count_u64((49) as u64)));
            } else {
                (*ptrfpsf) |= 16;
            }
        }
    }
    if (unbexp > 0x17df) {
        (*ptrfpsf) |= (32 | 8);
        exp = 0;
        if (sign == 0) {
            if ((rnd_mode == 2) || (rnd_mode == 4)) {
                C_hi = 0x7800000000000000;
                C_lo = 0x0000000000000000;
            } else {
                C_hi = 0x5fffed09bead87c0;
                C_lo = 0x378d8e63ffffffff;
            }
        } else {
            if ((rnd_mode == 1) || (rnd_mode == 4)) {
                C_hi = 0xf800000000000000;
                C_lo = 0x0000000000000000;
            } else {
                C_hi = 0xdfffed09bead87c0;
                C_lo = 0x378d8e63ffffffff;
            }
        }
    }
    res.w[1] = ((sign | exp) | C_hi);
    res.w[0] = C_lo;
    (*ptrres) = res;
}

pub(crate) fn bid128_count_digits(mut C: BID_UINT128) -> i64 {
    let mut x_nr_bits: i64 = 0;
    if (C.w[1] == 0) {
        if (C.w[0] == 0) {
            return 0;
        }
        if (C.w[0] >= 0x0020000000000000) {
            let mut tmp = (((go_checked_shr_u64(C.w[0], go_shift_count_u64((32) as u64))) as f64)).to_bits();
            x_nr_bits = ((33 as i64).wrapping_add(((((((go_checked_shr_u64(tmp, go_shift_count_u64((52) as u64)))) & 0x7ff)).wrapping_sub(0x3ff)) as i64)));
        } else {
            let mut tmp = (C.w[0] as f64).to_bits();
            x_nr_bits = ((1 as i64).wrapping_add(((((((go_checked_shr_u64(tmp, go_shift_count_u64((52) as u64)))) & 0x7ff)).wrapping_sub(0x3ff)) as i64)));
        }
    } else {
        let mut tmp = (C.w[1] as f64).to_bits();
        x_nr_bits = ((65 as i64).wrapping_add(((((((go_checked_shr_u64(tmp, go_shift_count_u64((52) as u64)))) & 0x7ff)).wrapping_sub(0x3ff)) as i64)));
    }
    let mut q = (bid_nr_digits[(x_nr_bits.wrapping_sub(1)) as usize].digits as i64);
    if (q == 0) {
        q = (bid_nr_digits[(x_nr_bits.wrapping_sub(1)) as usize].digits1 as i64);
        if ((C.w[1] > bid_nr_digits[(x_nr_bits.wrapping_sub(1)) as usize].threshold_hi) || (((C.w[1] == bid_nr_digits[(x_nr_bits.wrapping_sub(1)) as usize].threshold_hi) && (C.w[0] >= bid_nr_digits[(x_nr_bits.wrapping_sub(1)) as usize].threshold_lo)))) {
            q = q.wrapping_add(1);
        }
    }
    return q;
}
