// Auto-generated from bid128_noncomp.go by go2rs. Do not edit.

use super::prelude::*;

pub fn bid128_is_signed(mut x: BID_UINT128) -> i64 {
    if ((x.w[1] & 0x8000000000000000) == 0x8000000000000000) {
        return 1;
    }
    return 0;
}

pub fn bid128_is_normal(mut x: BID_UINT128) -> i64 {
    if ((x.w[1] & 0x7800000000000000) == 0x7800000000000000) {
        return 0;
    }
    let mut C1_hi = (x.w[1] & 0x1ffffffffffff);
    let mut C1_lo = x.w[0];
    if ((C1_hi == 0) && (C1_lo == 0)) {
        return 0;
    }
    if ((((((C1_hi > 0x0001ed09bead87c0) || (((C1_hi == 0x0001ed09bead87c0) && (C1_lo > 0x378d8e63ffffffff))))) && (((x.w[1] & 0x6000000000000000) != 0x6000000000000000)))) || (((x.w[1] & 0x6000000000000000) == 0x6000000000000000))) {
        return 0;
    }
    let mut x_nr_bits: i64 = 0;
    if (C1_hi == 0) {
        if (C1_lo >= 0x0020000000000000) {
            let mut tmp1 = (((go_checked_shr_u64(C1_lo, go_shift_count_u64((32) as u64))) as f64)).to_bits();
            x_nr_bits = ((33 as i64).wrapping_add(((((((go_checked_shr_u64(tmp1, go_shift_count_u64((52) as u64)))) & 0x7ff)).wrapping_sub(0x3ff)) as i64)));
        } else {
            let mut tmp1 = (C1_lo as f64).to_bits();
            x_nr_bits = ((1 as i64).wrapping_add(((((((go_checked_shr_u64(tmp1, go_shift_count_u64((52) as u64)))) & 0x7ff)).wrapping_sub(0x3ff)) as i64)));
        }
    } else {
        let mut tmp1 = (C1_hi as f64).to_bits();
        x_nr_bits = ((65 as i64).wrapping_add(((((((go_checked_shr_u64(tmp1, go_shift_count_u64((52) as u64)))) & 0x7ff)).wrapping_sub(0x3ff)) as i64)));
    }
    let mut q = (bid_nr_digits[(x_nr_bits.wrapping_sub(1)) as usize].digits as i64);
    if (q == 0) {
        q = (bid_nr_digits[(x_nr_bits.wrapping_sub(1)) as usize].digits1 as i64);
        if ((C1_hi > bid_nr_digits[(x_nr_bits.wrapping_sub(1)) as usize].threshold_hi) || (((C1_hi == bid_nr_digits[(x_nr_bits.wrapping_sub(1)) as usize].threshold_hi) && (C1_lo >= bid_nr_digits[(x_nr_bits.wrapping_sub(1)) as usize].threshold_lo)))) {
            q = q.wrapping_add(1);
        }
    }
    let mut x_exp = (x.w[1] & 0x7ffe000000000000);
    let mut exp = (((go_checked_shr_u64(x_exp, go_shift_count_u64((49) as u64))) as i64).wrapping_sub(6176));
    if ((exp.wrapping_add(q)) <= (-6143)) {
        return 0;
    }
    return 1;
}

pub fn bid128_is_subnormal(mut x: BID_UINT128) -> i64 {
    if ((x.w[1] & 0x7800000000000000) == 0x7800000000000000) {
        return 0;
    }
    let mut C1_hi = (x.w[1] & 0x1ffffffffffff);
    let mut C1_lo = x.w[0];
    if ((C1_hi == 0) && (C1_lo == 0)) {
        return 0;
    }
    if ((((((C1_hi > 0x0001ed09bead87c0) || (((C1_hi == 0x0001ed09bead87c0) && (C1_lo > 0x378d8e63ffffffff))))) && (((x.w[1] & 0x6000000000000000) != 0x6000000000000000)))) || (((x.w[1] & 0x6000000000000000) == 0x6000000000000000))) {
        return 0;
    }
    let mut x_nr_bits: i64 = 0;
    if (C1_hi == 0) {
        if (C1_lo >= 0x0020000000000000) {
            let mut tmp1 = (((go_checked_shr_u64(C1_lo, go_shift_count_u64((32) as u64))) as f64)).to_bits();
            x_nr_bits = ((33 as i64).wrapping_add(((((((go_checked_shr_u64(tmp1, go_shift_count_u64((52) as u64)))) & 0x7ff)).wrapping_sub(0x3ff)) as i64)));
        } else {
            let mut tmp1 = (C1_lo as f64).to_bits();
            x_nr_bits = ((1 as i64).wrapping_add(((((((go_checked_shr_u64(tmp1, go_shift_count_u64((52) as u64)))) & 0x7ff)).wrapping_sub(0x3ff)) as i64)));
        }
    } else {
        let mut tmp1 = (C1_hi as f64).to_bits();
        x_nr_bits = ((65 as i64).wrapping_add(((((((go_checked_shr_u64(tmp1, go_shift_count_u64((52) as u64)))) & 0x7ff)).wrapping_sub(0x3ff)) as i64)));
    }
    let mut q = (bid_nr_digits[(x_nr_bits.wrapping_sub(1)) as usize].digits as i64);
    if (q == 0) {
        q = (bid_nr_digits[(x_nr_bits.wrapping_sub(1)) as usize].digits1 as i64);
        if ((C1_hi > bid_nr_digits[(x_nr_bits.wrapping_sub(1)) as usize].threshold_hi) || (((C1_hi == bid_nr_digits[(x_nr_bits.wrapping_sub(1)) as usize].threshold_hi) && (C1_lo >= bid_nr_digits[(x_nr_bits.wrapping_sub(1)) as usize].threshold_lo)))) {
            q = q.wrapping_add(1);
        }
    }
    let mut x_exp = (x.w[1] & 0x7ffe000000000000);
    let mut exp = (((go_checked_shr_u64(x_exp, go_shift_count_u64((49) as u64))) as i64).wrapping_sub(6176));
    if ((exp.wrapping_add(q)) <= (-6143)) {
        return 1;
    }
    return 0;
}

pub fn bid128_is_finite(mut x: BID_UINT128) -> i64 {
    if ((x.w[1] & 0x7800000000000000) != 0x7800000000000000) {
        return 1;
    }
    return 0;
}

pub fn bid128_is_signaling(mut x: BID_UINT128) -> i64 {
    if ((x.w[1] & 0x7e00000000000000) == 0x7e00000000000000) {
        return 1;
    }
    return 0;
}

pub fn bid128_is_canonical(mut x: BID_UINT128) -> i64 {
    if ((x.w[1] & 0x7c00000000000000) == 0x7c00000000000000) {
        if ((x.w[1] & 0x01ffc00000000000) != 0) {
            return 0;
        }
        let mut sig_x_hi = (x.w[1] & 0x00003fffffffffff);
        let mut sig_x_lo = x.w[0];
        if ((sig_x_hi < 0x0000314dc6448d93) || (((sig_x_hi == 0x0000314dc6448d93) && (sig_x_lo < 0x38c15b0a00000000)))) {
            return 1;
        }
        return 0;
    }
    if ((x.w[1] & 0x7800000000000000) == 0x7800000000000000) {
        if (((x.w[1] & 0x03ffffffffffffff) != 0) || (x.w[0] != 0)) {
            return 0;
        }
        return 1;
    }
    let mut sig_x_hi = (x.w[1] & 0x0001ffffffffffff);
    let mut sig_x_lo = x.w[0];
    if (((sig_x_hi > 0x0001ed09bead87c0) || (((sig_x_hi == 0x0001ed09bead87c0) && (sig_x_lo > 0x378d8e63ffffffff)))) || (((x.w[1] & 0x6000000000000000) == 0x6000000000000000))) {
        return 0;
    }
    return 1;
}

pub fn bid128_copy(mut x: BID_UINT128) -> BID_UINT128 {
    return x;
}

pub fn bid128_negate(mut x: BID_UINT128) -> BID_UINT128 {
    return BID_UINT128 { w: [x.w[0], (x.w[1] ^ 0x8000000000000000)], ..Default::default() };
}

pub fn bid128_abs(mut x: BID_UINT128) -> BID_UINT128 {
    return BID_UINT128 { w: [x.w[0], (x.w[1] & 0x7fffffffffffffff)], ..Default::default() };
}

pub fn bid128_copy_sign(mut x: BID_UINT128, mut y: BID_UINT128) -> BID_UINT128 {
    return BID_UINT128 { w: [x.w[0], ((x.w[1] & 0x7fffffffffffffff) | (y.w[1] & 0x8000000000000000))], ..Default::default() };
}

pub fn bid128_radix() -> i64 {
    return 10;
}

pub fn bid128_inf() -> BID_UINT128 {
    return BID_UINT128 { w: [0x0000000000000000, 0x7800000000000000], ..Default::default() };
}

pub fn bid128_na_n(tagp: impl AsRef<str>) -> BID_UINT128 {
    let mut tagp = tagp.as_ref().to_string();
    let mut res = BID_UINT128 { w: [0x0000000000000000, 0x7c00000000000000], ..Default::default() };
    if (tagp == "") {
        return res;
    }
    let (mut x, _) = bid128_from_string(tagp, 0);
    x.w[1] = (x.w[1] & 0x00003fffffffffff);
    res.w[1] = (res.w[1] | x.w[1]);
    res.w[0] = x.w[0];
    return res;
}

pub fn bid128_same_quantum(mut x: BID_UINT128, mut y: BID_UINT128) -> i64 {
    if (((x.w[1] & 0x7800000000000000) == 0x7800000000000000) || ((y.w[1] & 0x7800000000000000) == 0x7800000000000000)) {
        if (((x.w[1] & 0x7800000000000000) == 0x7800000000000000) && ((y.w[1] & 0x7800000000000000) == 0x7800000000000000)) {
            if (((x.w[1] & 0x7c00000000000000) == 0x7c00000000000000) || ((y.w[1] & 0x7c00000000000000) == 0x7c00000000000000)) {
                if (((x.w[1] & 0x7c00000000000000) == 0x7c00000000000000) && ((y.w[1] & 0x7c00000000000000) == 0x7c00000000000000)) {
                    return 1;
                }
                return 0;
            }
            return 1;
        }
        return 0;
    }
    let mut exp_x: u64 = 0;
    let mut exp_y: u64 = 0;
    if ((x.w[1] & 0x6000000000000000) == 0x6000000000000000) {
        exp_x = (((go_checked_shl_u64(x.w[1], go_shift_count_u64((2) as u64)))) & 0x7ffe000000000000);
    } else {
        exp_x = (x.w[1] & 0x7ffe000000000000);
    }
    if ((y.w[1] & 0x6000000000000000) == 0x6000000000000000) {
        exp_y = (((go_checked_shl_u64(y.w[1], go_shift_count_u64((2) as u64)))) & 0x7ffe000000000000);
    } else {
        exp_y = (y.w[1] & 0x7ffe000000000000);
    }
    if (exp_x == exp_y) {
        return 1;
    }
    return 0;
}
