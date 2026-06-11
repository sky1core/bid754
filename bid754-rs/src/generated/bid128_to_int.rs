// Auto-generated from bid128_to_int.go by go2rs. Do not edit.

use super::prelude::*;

pub(crate) fn bid128_unpack_for_int(mut x: BID_UINT128) -> (u64, u64, BID_UINT128, bool) {
    let mut x_sign: u64 = 0;
    let mut x_exp: u64 = 0;
    let mut C1: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut is_special: bool = false;
    x_sign = (x.w[1] & 0x8000000000000000);
    x_exp = (x.w[1] & 0x7ffe000000000000);
    C1.w[1] = (x.w[1] & 0x1ffffffffffff);
    C1.w[0] = x.w[0];
    is_special = ((x.w[1] & 0x7800000000000000) == 0x7800000000000000);
    return (x_sign, x_exp, C1, is_special);
}

pub(crate) fn bid128_is_nan_for_int(mut x: BID_UINT128) -> bool {
    return ((x.w[1] & 0x7c00000000000000) == 0x7c00000000000000);
}

pub(crate) fn bid128_is_snan_for_int(mut x: BID_UINT128) -> bool {
    return ((x.w[1] & 0x7e00000000000000) == 0x7e00000000000000);
}

pub(crate) fn bid128_is_noncanonical(mut C1: BID_UINT128, mut x: BID_UINT128) -> bool {
    return (((C1.w[1] > 0x0001ed09bead87c0) || (((C1.w[1] == 0x0001ed09bead87c0) && (C1.w[0] > 0x378d8e63ffffffff)))) || (((x.w[1] & 0x6000000000000000) == 0x6000000000000000)));
}

pub(crate) fn bid128_nr_digits(mut C1: BID_UINT128) -> (i64, u64) {
    let mut q: i64 = 0;
    let mut x_nr_bits: u64 = 0;
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
    return (q, x_nr_bits);
}

pub(crate) fn bid128_check_overflow_10(mut C1: BID_UINT128, mut x_sign: u64, mut q: i64, mut neg_limit: u64, mut neg_cmp_ge: bool, mut pos_limit: u64, mut pos_cmp_ge: bool) -> (bool, u32) {
    let mut C: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut tmp64: u64 = 0;
    let mut pfpsf: u32 = 0;
    if (x_sign != 0) {
        if (q <= 11) {
            tmp64 = (C1.w[0].wrapping_mul(bid_ten2k64[((11 as i64).wrapping_sub(q)) as usize]));
            if neg_cmp_ge {
                if (tmp64 >= neg_limit) {
                    pfpsf |= 1;
                    return (true, pfpsf);
                }
            } else {
                if (tmp64 > neg_limit) {
                    pfpsf |= 1;
                    return (true, pfpsf);
                }
            }
        } else {
            tmp64 = neg_limit;
            if ((q.wrapping_sub(11)) <= 19) {
                C = __mul_64x64_to_128(tmp64, bid_ten2k64[(q.wrapping_sub(11)) as usize]);
            } else {
                C = __mul_128x64_to_128(tmp64, bid_ten2k128[(q.wrapping_sub(31)) as usize]);
            }
            if neg_cmp_ge {
                if ((C1.w[1] > C.w[1]) || (((C1.w[1] == C.w[1]) && (C1.w[0] >= C.w[0])))) {
                    pfpsf |= 1;
                    return (true, pfpsf);
                }
            } else {
                if ((C1.w[1] > C.w[1]) || (((C1.w[1] == C.w[1]) && (C1.w[0] > C.w[0])))) {
                    pfpsf |= 1;
                    return (true, pfpsf);
                }
            }
        }
    } else {
        if (q <= 11) {
            tmp64 = (C1.w[0].wrapping_mul(bid_ten2k64[((11 as i64).wrapping_sub(q)) as usize]));
            if pos_cmp_ge {
                if (tmp64 >= pos_limit) {
                    pfpsf |= 1;
                    return (true, pfpsf);
                }
            } else {
                if (tmp64 > pos_limit) {
                    pfpsf |= 1;
                    return (true, pfpsf);
                }
            }
        } else {
            tmp64 = pos_limit;
            if ((q.wrapping_sub(11)) <= 19) {
                C = __mul_64x64_to_128(tmp64, bid_ten2k64[(q.wrapping_sub(11)) as usize]);
            } else {
                C = __mul_128x64_to_128(tmp64, bid_ten2k128[(q.wrapping_sub(31)) as usize]);
            }
            if pos_cmp_ge {
                if ((C1.w[1] > C.w[1]) || (((C1.w[1] == C.w[1]) && (C1.w[0] >= C.w[0])))) {
                    pfpsf |= 1;
                    return (true, pfpsf);
                }
            } else {
                if ((C1.w[1] > C.w[1]) || (((C1.w[1] == C.w[1]) && (C1.w[0] > C.w[0])))) {
                    pfpsf |= 1;
                    return (true, pfpsf);
                }
            }
        }
    }
    return (false, 0);
}

pub(crate) fn bid128_round_rnint_common(mut C1: BID_UINT128, mut ind: i64) -> u64 {
    let mut Cstar_w0: u64 = 0;
    let mut Cstar: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut fstar: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut P256: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut tmp64 = C1.w[0];
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
    if ((ind.wrapping_sub(1)) <= 21) {
        Cstar.w[1] = P256.w[3];
        Cstar.w[0] = P256.w[2];
        fstar.w[3] = 0;
        fstar.w[2] = (P256.w[2] & bid_maskhigh128[(ind.wrapping_sub(1)) as usize]);
        fstar.w[1] = P256.w[1];
        fstar.w[0] = P256.w[0];
    } else {
        Cstar.w[1] = 0;
        Cstar.w[0] = P256.w[3];
        fstar.w[3] = (P256.w[3] & bid_maskhigh128[(ind.wrapping_sub(1)) as usize]);
        fstar.w[2] = P256.w[2];
        fstar.w[1] = P256.w[1];
        fstar.w[0] = P256.w[0];
    }
    let mut shift = (bid_shiftright128[(ind.wrapping_sub(1)) as usize] as i64);
    if ((ind.wrapping_sub(1)) <= 21) {
        Cstar.w[0] = (((go_checked_shr_u64(Cstar.w[0], go_shift_count_u64((shift as u64) as u64)))) | ((go_checked_shl_u64(Cstar.w[1], go_shift_count_u64(((((64 as i64).wrapping_sub(shift)) as u64)) as u64)))));
    } else {
        Cstar.w[0] = (go_checked_shr_u64(Cstar.w[0], go_shift_count_u64((((shift.wrapping_sub(64)) as u64)) as u64)));
    }
    if ((((fstar.w[3] == 0) && (fstar.w[2] == 0)) && (((fstar.w[1] != 0) || (fstar.w[0] != 0)))) && (((fstar.w[1] < bid_ten2mk128trunc[(ind.wrapping_sub(1)) as usize].w[1]) || (((fstar.w[1] == bid_ten2mk128trunc[(ind.wrapping_sub(1)) as usize].w[1]) && (fstar.w[0] <= bid_ten2mk128trunc[(ind.wrapping_sub(1)) as usize].w[0])))))) {
        if ((Cstar.w[0] & 0x01) != 0) {
            Cstar.w[0] = Cstar.w[0].wrapping_sub(1);
        }
    }
    return Cstar.w[0];
}

pub(crate) fn bid128_round_floor_ceil_int_common(mut C1: BID_UINT128, mut ind: i64, mut x_sign: u64, mut mode: i64) -> u64 {
    let mut Cstar_w0: u64 = 0;
    let mut Cstar: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut fstar: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut P256: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut is_inexact_lt_midpoint: i64 = 0;
    let mut is_inexact_gt_midpoint: i64 = 0;
    let mut is_midpoint_lt_even: i64 = 0;
    let mut is_midpoint_gt_even: i64 = 0;
    _ = is_midpoint_gt_even;
    let mut tmp64 = C1.w[0];
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
    if ((ind.wrapping_sub(1)) <= 21) {
        Cstar.w[1] = P256.w[3];
        Cstar.w[0] = P256.w[2];
        fstar.w[3] = 0;
        fstar.w[2] = (P256.w[2] & bid_maskhigh128[(ind.wrapping_sub(1)) as usize]);
        fstar.w[1] = P256.w[1];
        fstar.w[0] = P256.w[0];
    } else {
        Cstar.w[1] = 0;
        Cstar.w[0] = P256.w[3];
        fstar.w[3] = (P256.w[3] & bid_maskhigh128[(ind.wrapping_sub(1)) as usize]);
        fstar.w[2] = P256.w[2];
        fstar.w[1] = P256.w[1];
        fstar.w[0] = P256.w[0];
    }
    let mut shift = (bid_shiftright128[(ind.wrapping_sub(1)) as usize] as i64);
    if ((ind.wrapping_sub(1)) <= 21) {
        Cstar.w[0] = (((go_checked_shr_u64(Cstar.w[0], go_shift_count_u64((shift as u64) as u64)))) | ((go_checked_shl_u64(Cstar.w[1], go_shift_count_u64(((((64 as i64).wrapping_sub(shift)) as u64)) as u64)))));
    } else {
        Cstar.w[0] = (go_checked_shr_u64(Cstar.w[0], go_shift_count_u64((((shift.wrapping_sub(64)) as u64)) as u64)));
    }
    let mut tmp64A: u64 = 0;
    if ((ind.wrapping_sub(1)) <= 2) {
        if ((fstar.w[1] > 0x8000000000000000) || (((fstar.w[1] == 0x8000000000000000) && (fstar.w[0] > 0x0)))) {
            tmp64 = (fstar.w[1].wrapping_sub(0x8000000000000000));
            if ((tmp64 > bid_ten2mk128trunc[(ind.wrapping_sub(1)) as usize].w[1]) || (((tmp64 == bid_ten2mk128trunc[(ind.wrapping_sub(1)) as usize].w[1]) && (fstar.w[0] >= bid_ten2mk128trunc[(ind.wrapping_sub(1)) as usize].w[0])))) {
                is_inexact_lt_midpoint = 1;
            }
        } else {
            is_inexact_gt_midpoint = 1;
        }
    } else if ((ind.wrapping_sub(1)) <= 21) {
        if (((fstar.w[3] > 0x0) || (((fstar.w[3] == 0x0) && (fstar.w[2] > bid_onehalf128[(ind.wrapping_sub(1)) as usize])))) || ((((fstar.w[3] == 0x0) && (fstar.w[2] == bid_onehalf128[(ind.wrapping_sub(1)) as usize])) && (((fstar.w[1] != 0) || (fstar.w[0] != 0)))))) {
            tmp64 = (fstar.w[2].wrapping_sub(bid_onehalf128[(ind.wrapping_sub(1)) as usize]));
            tmp64A = fstar.w[3];
            if (tmp64 > fstar.w[2]) {
                tmp64A = tmp64A.wrapping_sub(1);
            }
            if ((((tmp64A != 0) || (tmp64 != 0)) || (fstar.w[1] > bid_ten2mk128trunc[(ind.wrapping_sub(1)) as usize].w[1])) || (((fstar.w[1] == bid_ten2mk128trunc[(ind.wrapping_sub(1)) as usize].w[1]) && (fstar.w[0] > bid_ten2mk128trunc[(ind.wrapping_sub(1)) as usize].w[0])))) {
                is_inexact_lt_midpoint = 1;
            }
        } else {
            is_inexact_gt_midpoint = 1;
        }
    } else {
        if ((fstar.w[3] > bid_onehalf128[(ind.wrapping_sub(1)) as usize]) || (((fstar.w[3] == bid_onehalf128[(ind.wrapping_sub(1)) as usize]) && ((((fstar.w[2] != 0) || (fstar.w[1] != 0)) || (fstar.w[0] != 0)))))) {
            tmp64 = (fstar.w[3].wrapping_sub(bid_onehalf128[(ind.wrapping_sub(1)) as usize]));
            if ((((tmp64 != 0) || (fstar.w[2] != 0)) || (fstar.w[1] > bid_ten2mk128trunc[(ind.wrapping_sub(1)) as usize].w[1])) || (((fstar.w[1] == bid_ten2mk128trunc[(ind.wrapping_sub(1)) as usize].w[1]) && (fstar.w[0] > bid_ten2mk128trunc[(ind.wrapping_sub(1)) as usize].w[0])))) {
                is_inexact_lt_midpoint = 1;
            }
        } else {
            is_inexact_gt_midpoint = 1;
        }
    }
    if ((((fstar.w[3] == 0) && (fstar.w[2] == 0)) && (((fstar.w[1] != 0) || (fstar.w[0] != 0)))) && (((fstar.w[1] < bid_ten2mk128trunc[(ind.wrapping_sub(1)) as usize].w[1]) || (((fstar.w[1] == bid_ten2mk128trunc[(ind.wrapping_sub(1)) as usize].w[1]) && (fstar.w[0] <= bid_ten2mk128trunc[(ind.wrapping_sub(1)) as usize].w[0])))))) {
        if ((Cstar.w[0] & 0x01) != 0) {
            Cstar.w[0] = Cstar.w[0].wrapping_sub(1);
            is_midpoint_gt_even = 1;
            is_inexact_lt_midpoint = 0;
            is_inexact_gt_midpoint = 0;
        } else {
            is_midpoint_lt_even = 1;
            is_inexact_lt_midpoint = 0;
            is_inexact_gt_midpoint = 0;
        }
    }
    match mode {
        0 => {
            if ((x_sign != 0) && (((is_midpoint_gt_even != 0) || (is_inexact_lt_midpoint != 0)))) {
                Cstar.w[0] = (Cstar.w[0].wrapping_add(1));
            } else if ((x_sign == 0) && (((is_midpoint_lt_even != 0) || (is_inexact_gt_midpoint != 0)))) {
                Cstar.w[0] = (Cstar.w[0].wrapping_sub(1));
            }
        }
        1 => {
            if ((x_sign != 0) && (((is_midpoint_lt_even != 0) || (is_inexact_gt_midpoint != 0)))) {
                Cstar.w[0] = (Cstar.w[0].wrapping_sub(1));
            } else if ((x_sign == 0) && (((is_midpoint_gt_even != 0) || (is_inexact_lt_midpoint != 0)))) {
                Cstar.w[0] = (Cstar.w[0].wrapping_add(1));
            }
        }
        2 => {
            if ((is_midpoint_lt_even != 0) || (is_inexact_gt_midpoint != 0)) {
                Cstar.w[0] = (Cstar.w[0].wrapping_sub(1));
            }
        }
        _ => {}
    }
    return Cstar.w[0];
}

pub(crate) fn bid128_trunc_inexact_common(mut C1: BID_UINT128, mut ind: i64) -> (u64, bool) {
    let mut Cstar_w0: u64 = 0;
    let mut inexact: bool = false;
    let mut Cstar: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut fstar: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut P256: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    P256 = __mul_128x128_to_256(C1, bid_ten2mk128[(ind.wrapping_sub(1)) as usize]);
    if ((ind.wrapping_sub(1)) <= 21) {
        Cstar.w[1] = P256.w[3];
        Cstar.w[0] = P256.w[2];
        fstar.w[3] = 0;
        fstar.w[2] = (P256.w[2] & bid_maskhigh128[(ind.wrapping_sub(1)) as usize]);
        fstar.w[1] = P256.w[1];
        fstar.w[0] = P256.w[0];
    } else {
        Cstar.w[1] = 0;
        Cstar.w[0] = P256.w[3];
        fstar.w[3] = (P256.w[3] & bid_maskhigh128[(ind.wrapping_sub(1)) as usize]);
        fstar.w[2] = P256.w[2];
        fstar.w[1] = P256.w[1];
        fstar.w[0] = P256.w[0];
    }
    let mut shift = (bid_shiftright128[(ind.wrapping_sub(1)) as usize] as i64);
    if ((ind.wrapping_sub(1)) <= 21) {
        Cstar.w[0] = (((go_checked_shr_u64(Cstar.w[0], go_shift_count_u64((shift as u64) as u64)))) | ((go_checked_shl_u64(Cstar.w[1], go_shift_count_u64(((((64 as i64).wrapping_sub(shift)) as u64)) as u64)))));
    } else {
        Cstar.w[0] = (go_checked_shr_u64(Cstar.w[0], go_shift_count_u64((((shift.wrapping_sub(64)) as u64)) as u64)));
    }
    if ((ind.wrapping_sub(1)) <= 2) {
        inexact = ((fstar.w[1] > bid_ten2mk128trunc[(ind.wrapping_sub(1)) as usize].w[1]) || (((fstar.w[1] == bid_ten2mk128trunc[(ind.wrapping_sub(1)) as usize].w[1]) && (fstar.w[0] > bid_ten2mk128trunc[(ind.wrapping_sub(1)) as usize].w[0]))));
    } else if ((ind.wrapping_sub(1)) <= 21) {
        inexact = (((fstar.w[2] != 0) || (fstar.w[1] > bid_ten2mk128trunc[(ind.wrapping_sub(1)) as usize].w[1])) || (((fstar.w[1] == bid_ten2mk128trunc[(ind.wrapping_sub(1)) as usize].w[1]) && (fstar.w[0] > bid_ten2mk128trunc[(ind.wrapping_sub(1)) as usize].w[0]))));
    } else {
        inexact = ((((fstar.w[3] != 0) || (fstar.w[2] != 0)) || (fstar.w[1] > bid_ten2mk128trunc[(ind.wrapping_sub(1)) as usize].w[1])) || (((fstar.w[1] == bid_ten2mk128trunc[(ind.wrapping_sub(1)) as usize].w[1]) && (fstar.w[0] > bid_ten2mk128trunc[(ind.wrapping_sub(1)) as usize].w[0]))));
    }
    return (Cstar.w[0], inexact);
}

pub(crate) fn bid128_round_trunc_mode_common(mut C1: BID_UINT128, mut ind: i64, mut x_sign: u64, mut mode: i64, mut setInexact: bool) -> (u64, u32) {
    let mut Cstar_w0: u64 = 0;
    let mut pfpsf: u32 = 0;
    let (mut Cstar_w0, mut inexact) = bid128_trunc_inexact_common(C1, ind);
    match mode {
        0 => {
            if ((x_sign != 0) && inexact) {
                Cstar_w0 = Cstar_w0.wrapping_add(1);
            }
        }
        1 => {
            if ((x_sign == 0) && inexact) {
                Cstar_w0 = Cstar_w0.wrapping_add(1);
            }
        }
        2 => {
        }
        _ => {}
    }
    if (setInexact && inexact) {
        pfpsf |= 32;
    }
    return (Cstar_w0, pfpsf);
}

pub(crate) fn bid128_round_xrnint_common(mut C1: BID_UINT128, mut ind: i64) -> (u64, u32) {
    let mut Cstar_w0: u64 = 0;
    let mut pfpsf: u32 = 0;
    let mut Cstar: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut fstar: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut P256: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut tmp64 = C1.w[0];
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
    if ((ind.wrapping_sub(1)) <= 21) {
        Cstar.w[1] = P256.w[3];
        Cstar.w[0] = P256.w[2];
        fstar.w[3] = 0;
        fstar.w[2] = (P256.w[2] & bid_maskhigh128[(ind.wrapping_sub(1)) as usize]);
        fstar.w[1] = P256.w[1];
        fstar.w[0] = P256.w[0];
    } else {
        Cstar.w[1] = 0;
        Cstar.w[0] = P256.w[3];
        fstar.w[3] = (P256.w[3] & bid_maskhigh128[(ind.wrapping_sub(1)) as usize]);
        fstar.w[2] = P256.w[2];
        fstar.w[1] = P256.w[1];
        fstar.w[0] = P256.w[0];
    }
    let mut shift = (bid_shiftright128[(ind.wrapping_sub(1)) as usize] as i64);
    if ((ind.wrapping_sub(1)) <= 21) {
        Cstar.w[0] = (((go_checked_shr_u64(Cstar.w[0], go_shift_count_u64((shift as u64) as u64)))) | ((go_checked_shl_u64(Cstar.w[1], go_shift_count_u64(((((64 as i64).wrapping_sub(shift)) as u64)) as u64)))));
    } else {
        Cstar.w[0] = (go_checked_shr_u64(Cstar.w[0], go_shift_count_u64((((shift.wrapping_sub(64)) as u64)) as u64)));
    }
    let mut tmp64A: u64 = 0;
    if ((ind.wrapping_sub(1)) <= 2) {
        if ((fstar.w[1] > 0x8000000000000000) || (((fstar.w[1] == 0x8000000000000000) && (fstar.w[0] > 0x0)))) {
            tmp64 = (fstar.w[1].wrapping_sub(0x8000000000000000));
            if ((tmp64 > bid_ten2mk128trunc[(ind.wrapping_sub(1)) as usize].w[1]) || (((tmp64 == bid_ten2mk128trunc[(ind.wrapping_sub(1)) as usize].w[1]) && (fstar.w[0] >= bid_ten2mk128trunc[(ind.wrapping_sub(1)) as usize].w[0])))) {
                pfpsf |= 32;
            }
        } else {
            pfpsf |= 32;
        }
    } else if ((ind.wrapping_sub(1)) <= 21) {
        if (((fstar.w[3] > 0x0) || (((fstar.w[3] == 0x0) && (fstar.w[2] > bid_onehalf128[(ind.wrapping_sub(1)) as usize])))) || ((((fstar.w[3] == 0x0) && (fstar.w[2] == bid_onehalf128[(ind.wrapping_sub(1)) as usize])) && (((fstar.w[1] != 0) || (fstar.w[0] != 0)))))) {
            tmp64 = (fstar.w[2].wrapping_sub(bid_onehalf128[(ind.wrapping_sub(1)) as usize]));
            tmp64A = fstar.w[3];
            if (tmp64 > fstar.w[2]) {
                tmp64A = tmp64A.wrapping_sub(1);
            }
            if ((((tmp64A != 0) || (tmp64 != 0)) || (fstar.w[1] > bid_ten2mk128trunc[(ind.wrapping_sub(1)) as usize].w[1])) || (((fstar.w[1] == bid_ten2mk128trunc[(ind.wrapping_sub(1)) as usize].w[1]) && (fstar.w[0] > bid_ten2mk128trunc[(ind.wrapping_sub(1)) as usize].w[0])))) {
                pfpsf |= 32;
            }
        } else {
            pfpsf |= 32;
        }
    } else {
        if ((fstar.w[3] > bid_onehalf128[(ind.wrapping_sub(1)) as usize]) || (((fstar.w[3] == bid_onehalf128[(ind.wrapping_sub(1)) as usize]) && ((((fstar.w[2] != 0) || (fstar.w[1] != 0)) || (fstar.w[0] != 0)))))) {
            tmp64 = (fstar.w[3].wrapping_sub(bid_onehalf128[(ind.wrapping_sub(1)) as usize]));
            if ((((tmp64 != 0) || (fstar.w[2] != 0)) || (fstar.w[1] > bid_ten2mk128trunc[(ind.wrapping_sub(1)) as usize].w[1])) || (((fstar.w[1] == bid_ten2mk128trunc[(ind.wrapping_sub(1)) as usize].w[1]) && (fstar.w[0] > bid_ten2mk128trunc[(ind.wrapping_sub(1)) as usize].w[0])))) {
                pfpsf |= 32;
            }
        } else {
            pfpsf |= 32;
        }
    }
    if ((((fstar.w[3] == 0) && (fstar.w[2] == 0)) && (((fstar.w[1] != 0) || (fstar.w[0] != 0)))) && (((fstar.w[1] < bid_ten2mk128trunc[(ind.wrapping_sub(1)) as usize].w[1]) || (((fstar.w[1] == bid_ten2mk128trunc[(ind.wrapping_sub(1)) as usize].w[1]) && (fstar.w[0] <= bid_ten2mk128trunc[(ind.wrapping_sub(1)) as usize].w[0])))))) {
        if ((Cstar.w[0] & 0x01) != 0) {
            Cstar.w[0] = Cstar.w[0].wrapping_sub(1);
        }
    }
    return (Cstar.w[0], pfpsf);
}

pub(crate) fn bid128_round_xfloor_xceil_xint_common(mut C1: BID_UINT128, mut ind: i64, mut x_sign: u64, mut mode: i64) -> (u64, u32) {
    let mut Cstar_w0: u64 = 0;
    let mut pfpsf: u32 = 0;
    let mut Cstar: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut fstar: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut P256: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut is_inexact_lt_midpoint: i64 = 0;
    let mut is_inexact_gt_midpoint: i64 = 0;
    let mut is_midpoint_lt_even: i64 = 0;
    let mut is_midpoint_gt_even: i64 = 0;
    _ = is_midpoint_gt_even;
    let mut tmp64 = C1.w[0];
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
    if ((ind.wrapping_sub(1)) <= 21) {
        Cstar.w[1] = P256.w[3];
        Cstar.w[0] = P256.w[2];
        fstar.w[3] = 0;
        fstar.w[2] = (P256.w[2] & bid_maskhigh128[(ind.wrapping_sub(1)) as usize]);
        fstar.w[1] = P256.w[1];
        fstar.w[0] = P256.w[0];
    } else {
        Cstar.w[1] = 0;
        Cstar.w[0] = P256.w[3];
        fstar.w[3] = (P256.w[3] & bid_maskhigh128[(ind.wrapping_sub(1)) as usize]);
        fstar.w[2] = P256.w[2];
        fstar.w[1] = P256.w[1];
        fstar.w[0] = P256.w[0];
    }
    let mut shift = (bid_shiftright128[(ind.wrapping_sub(1)) as usize] as i64);
    if ((ind.wrapping_sub(1)) <= 21) {
        Cstar.w[0] = (((go_checked_shr_u64(Cstar.w[0], go_shift_count_u64((shift as u64) as u64)))) | ((go_checked_shl_u64(Cstar.w[1], go_shift_count_u64(((((64 as i64).wrapping_sub(shift)) as u64)) as u64)))));
    } else {
        Cstar.w[0] = (go_checked_shr_u64(Cstar.w[0], go_shift_count_u64((((shift.wrapping_sub(64)) as u64)) as u64)));
    }
    let mut tmp64A: u64 = 0;
    if ((ind.wrapping_sub(1)) <= 2) {
        if ((fstar.w[1] > 0x8000000000000000) || (((fstar.w[1] == 0x8000000000000000) && (fstar.w[0] > 0x0)))) {
            tmp64 = (fstar.w[1].wrapping_sub(0x8000000000000000));
            if ((tmp64 > bid_ten2mk128trunc[(ind.wrapping_sub(1)) as usize].w[1]) || (((tmp64 == bid_ten2mk128trunc[(ind.wrapping_sub(1)) as usize].w[1]) && (fstar.w[0] >= bid_ten2mk128trunc[(ind.wrapping_sub(1)) as usize].w[0])))) {
                is_inexact_lt_midpoint = 1;
            }
        } else {
            is_inexact_gt_midpoint = 1;
        }
    } else if ((ind.wrapping_sub(1)) <= 21) {
        if (((fstar.w[3] > 0x0) || (((fstar.w[3] == 0x0) && (fstar.w[2] > bid_onehalf128[(ind.wrapping_sub(1)) as usize])))) || ((((fstar.w[3] == 0x0) && (fstar.w[2] == bid_onehalf128[(ind.wrapping_sub(1)) as usize])) && (((fstar.w[1] != 0) || (fstar.w[0] != 0)))))) {
            tmp64 = (fstar.w[2].wrapping_sub(bid_onehalf128[(ind.wrapping_sub(1)) as usize]));
            tmp64A = fstar.w[3];
            if (tmp64 > fstar.w[2]) {
                tmp64A = tmp64A.wrapping_sub(1);
            }
            if ((((tmp64A != 0) || (tmp64 != 0)) || (fstar.w[1] > bid_ten2mk128trunc[(ind.wrapping_sub(1)) as usize].w[1])) || (((fstar.w[1] == bid_ten2mk128trunc[(ind.wrapping_sub(1)) as usize].w[1]) && (fstar.w[0] > bid_ten2mk128trunc[(ind.wrapping_sub(1)) as usize].w[0])))) {
                is_inexact_lt_midpoint = 1;
            }
        } else {
            is_inexact_gt_midpoint = 1;
        }
    } else {
        if ((fstar.w[3] > bid_onehalf128[(ind.wrapping_sub(1)) as usize]) || (((fstar.w[3] == bid_onehalf128[(ind.wrapping_sub(1)) as usize]) && ((((fstar.w[2] != 0) || (fstar.w[1] != 0)) || (fstar.w[0] != 0)))))) {
            tmp64 = (fstar.w[3].wrapping_sub(bid_onehalf128[(ind.wrapping_sub(1)) as usize]));
            if ((((tmp64 != 0) || (fstar.w[2] != 0)) || (fstar.w[1] > bid_ten2mk128trunc[(ind.wrapping_sub(1)) as usize].w[1])) || (((fstar.w[1] == bid_ten2mk128trunc[(ind.wrapping_sub(1)) as usize].w[1]) && (fstar.w[0] > bid_ten2mk128trunc[(ind.wrapping_sub(1)) as usize].w[0])))) {
                is_inexact_lt_midpoint = 1;
            }
        } else {
            is_inexact_gt_midpoint = 1;
        }
    }
    if ((((fstar.w[3] == 0) && (fstar.w[2] == 0)) && (((fstar.w[1] != 0) || (fstar.w[0] != 0)))) && (((fstar.w[1] < bid_ten2mk128trunc[(ind.wrapping_sub(1)) as usize].w[1]) || (((fstar.w[1] == bid_ten2mk128trunc[(ind.wrapping_sub(1)) as usize].w[1]) && (fstar.w[0] <= bid_ten2mk128trunc[(ind.wrapping_sub(1)) as usize].w[0])))))) {
        if ((Cstar.w[0] & 0x01) != 0) {
            Cstar.w[0] = Cstar.w[0].wrapping_sub(1);
            is_midpoint_gt_even = 1;
            is_inexact_lt_midpoint = 0;
            is_inexact_gt_midpoint = 0;
        } else {
            is_midpoint_lt_even = 1;
            is_inexact_lt_midpoint = 0;
            is_inexact_gt_midpoint = 0;
        }
    }
    if ((((is_inexact_lt_midpoint != 0) || (is_inexact_gt_midpoint != 0)) || (is_midpoint_lt_even != 0)) || (is_midpoint_gt_even != 0)) {
        pfpsf |= 32;
    }
    match mode {
        0 => {
            if ((x_sign != 0) && (((is_midpoint_gt_even != 0) || (is_inexact_lt_midpoint != 0)))) {
                Cstar.w[0] = (Cstar.w[0].wrapping_add(1));
            } else if ((x_sign == 0) && (((is_midpoint_lt_even != 0) || (is_inexact_gt_midpoint != 0)))) {
                Cstar.w[0] = (Cstar.w[0].wrapping_sub(1));
            }
        }
        1 => {
            if ((x_sign != 0) && (((is_midpoint_lt_even != 0) || (is_inexact_gt_midpoint != 0)))) {
                Cstar.w[0] = (Cstar.w[0].wrapping_sub(1));
            } else if ((x_sign == 0) && (((is_midpoint_gt_even != 0) || (is_inexact_lt_midpoint != 0)))) {
                Cstar.w[0] = (Cstar.w[0].wrapping_add(1));
            }
        }
        2 => {
            if ((is_midpoint_lt_even != 0) || (is_inexact_gt_midpoint != 0)) {
                Cstar.w[0] = (Cstar.w[0].wrapping_sub(1));
            }
        }
        _ => {}
    }
    return (Cstar.w[0], pfpsf);
}

pub(crate) fn bid128_round_rninta_common(mut C1: BID_UINT128, mut ind: i64) -> u64 {
    let mut Cstar_w0: u64 = 0;
    let mut Cstar: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut P256: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut tmp64 = C1.w[0];
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
    if ((ind.wrapping_sub(1)) <= 21) {
        Cstar.w[1] = P256.w[3];
        Cstar.w[0] = P256.w[2];
    } else {
        Cstar.w[1] = 0;
        Cstar.w[0] = P256.w[3];
    }
    let mut shift = (bid_shiftright128[(ind.wrapping_sub(1)) as usize] as i64);
    if ((ind.wrapping_sub(1)) <= 21) {
        Cstar.w[0] = (((go_checked_shr_u64(Cstar.w[0], go_shift_count_u64((shift as u64) as u64)))) | ((go_checked_shl_u64(Cstar.w[1], go_shift_count_u64(((((64 as i64).wrapping_sub(shift)) as u64)) as u64)))));
    } else {
        Cstar.w[0] = (go_checked_shr_u64(Cstar.w[0], go_shift_count_u64((((shift.wrapping_sub(64)) as u64)) as u64)));
    }
    return Cstar.w[0];
}

pub(crate) fn bid128_round_xrninta_common(mut C1: BID_UINT128, mut ind: i64) -> (u64, u32) {
    let mut Cstar_w0: u64 = 0;
    let mut pfpsf: u32 = 0;
    let mut Cstar: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut fstar: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut P256: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut tmp64 = C1.w[0];
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
    if ((ind.wrapping_sub(1)) <= 21) {
        Cstar.w[1] = P256.w[3];
        Cstar.w[0] = P256.w[2];
        fstar.w[3] = 0;
        fstar.w[2] = (P256.w[2] & bid_maskhigh128[(ind.wrapping_sub(1)) as usize]);
        fstar.w[1] = P256.w[1];
        fstar.w[0] = P256.w[0];
    } else {
        Cstar.w[1] = 0;
        Cstar.w[0] = P256.w[3];
        fstar.w[3] = (P256.w[3] & bid_maskhigh128[(ind.wrapping_sub(1)) as usize]);
        fstar.w[2] = P256.w[2];
        fstar.w[1] = P256.w[1];
        fstar.w[0] = P256.w[0];
    }
    let mut shift = (bid_shiftright128[(ind.wrapping_sub(1)) as usize] as i64);
    if ((ind.wrapping_sub(1)) <= 21) {
        Cstar.w[0] = (((go_checked_shr_u64(Cstar.w[0], go_shift_count_u64((shift as u64) as u64)))) | ((go_checked_shl_u64(Cstar.w[1], go_shift_count_u64(((((64 as i64).wrapping_sub(shift)) as u64)) as u64)))));
    } else {
        Cstar.w[0] = (go_checked_shr_u64(Cstar.w[0], go_shift_count_u64((((shift.wrapping_sub(64)) as u64)) as u64)));
    }
    let mut tmp64A: u64 = 0;
    if ((ind.wrapping_sub(1)) <= 2) {
        if ((fstar.w[1] > 0x8000000000000000) || (((fstar.w[1] == 0x8000000000000000) && (fstar.w[0] > 0x0)))) {
            tmp64 = (fstar.w[1].wrapping_sub(0x8000000000000000));
            if ((tmp64 > bid_ten2mk128trunc[(ind.wrapping_sub(1)) as usize].w[1]) || (((tmp64 == bid_ten2mk128trunc[(ind.wrapping_sub(1)) as usize].w[1]) && (fstar.w[0] >= bid_ten2mk128trunc[(ind.wrapping_sub(1)) as usize].w[0])))) {
                pfpsf |= 32;
            }
        } else {
            pfpsf |= 32;
        }
    } else if ((ind.wrapping_sub(1)) <= 21) {
        if (((fstar.w[3] > 0x0) || (((fstar.w[3] == 0x0) && (fstar.w[2] > bid_onehalf128[(ind.wrapping_sub(1)) as usize])))) || ((((fstar.w[3] == 0x0) && (fstar.w[2] == bid_onehalf128[(ind.wrapping_sub(1)) as usize])) && (((fstar.w[1] != 0) || (fstar.w[0] != 0)))))) {
            tmp64 = (fstar.w[2].wrapping_sub(bid_onehalf128[(ind.wrapping_sub(1)) as usize]));
            tmp64A = fstar.w[3];
            if (tmp64 > fstar.w[2]) {
                tmp64A = tmp64A.wrapping_sub(1);
            }
            if ((((tmp64A != 0) || (tmp64 != 0)) || (fstar.w[1] > bid_ten2mk128trunc[(ind.wrapping_sub(1)) as usize].w[1])) || (((fstar.w[1] == bid_ten2mk128trunc[(ind.wrapping_sub(1)) as usize].w[1]) && (fstar.w[0] > bid_ten2mk128trunc[(ind.wrapping_sub(1)) as usize].w[0])))) {
                pfpsf |= 32;
            }
        } else {
            pfpsf |= 32;
        }
    } else {
        if ((fstar.w[3] > bid_onehalf128[(ind.wrapping_sub(1)) as usize]) || (((fstar.w[3] == bid_onehalf128[(ind.wrapping_sub(1)) as usize]) && ((((fstar.w[2] != 0) || (fstar.w[1] != 0)) || (fstar.w[0] != 0)))))) {
            tmp64 = (fstar.w[3].wrapping_sub(bid_onehalf128[(ind.wrapping_sub(1)) as usize]));
            if ((((tmp64 != 0) || (fstar.w[2] != 0)) || (fstar.w[1] > bid_ten2mk128trunc[(ind.wrapping_sub(1)) as usize].w[1])) || (((fstar.w[1] == bid_ten2mk128trunc[(ind.wrapping_sub(1)) as usize].w[1]) && (fstar.w[0] > bid_ten2mk128trunc[(ind.wrapping_sub(1)) as usize].w[0])))) {
                pfpsf |= 32;
            }
        } else {
            pfpsf |= 32;
        }
    }
    return (Cstar.w[0], pfpsf);
}

pub fn bid128_to_int32_rnint(mut x: BID_UINT128) -> (i32, u32) {
    let mut res: i32 = 0;
    let mut pfpsf: u32 = 0;
    let (mut x_sign, mut x_exp, mut C1, mut is_special) = bid128_unpack_for_int(x);
    if is_special {
        pfpsf |= 1;
        return (((-0x80000000) as i32), pfpsf);
    }
    if bid128_is_noncanonical(C1, x) {
        return (0, 0);
    }
    if ((C1.w[1] == 0) && (C1.w[0] == 0)) {
        return (0, 0);
    }
    let (mut q, _) = bid128_nr_digits(C1);
    let mut exp = (((go_checked_shr_u64(x_exp, go_shift_count_u64((49) as u64))) as i64).wrapping_sub(6176));
    if ((q.wrapping_add(exp)) > 10) {
        pfpsf |= 1;
        return (((-0x80000000) as i32), pfpsf);
    } else if ((q.wrapping_add(exp)) == 10) {
        let (mut invalid, mut f) = bid128_check_overflow_10(C1, x_sign, q, 0x500000005, false, 0x4fffffffb, true);
        if invalid {
            return (((-0x80000000) as i32), f);
        }
    }
    if ((q.wrapping_add(exp)) < 0) {
        return (0, 0);
    } else if ((q.wrapping_add(exp)) == 0) {
        let mut ind = (q.wrapping_sub(1));
        if (ind <= 18) {
            if ((C1.w[1] == 0) && (C1.w[0] <= bid_midpoint64[ind as usize])) {
                res = 0;
            } else if (x_sign != 0) {
                res = (-1);
            } else {
                res = 1;
            }
        } else {
            if ((C1.w[1] < bid_midpoint128[(ind.wrapping_sub(19)) as usize].w[1]) || (((C1.w[1] == bid_midpoint128[(ind.wrapping_sub(19)) as usize].w[1]) && (C1.w[0] <= bid_midpoint128[(ind.wrapping_sub(19)) as usize].w[0])))) {
                res = 0;
            } else if (x_sign != 0) {
                res = (-1);
            } else {
                res = 1;
            }
        }
    } else {
        if (exp < 0) {
            let mut ind = (exp.wrapping_neg());
            let mut Cstar_w0 = bid128_round_rnint_common(C1, ind);
            if (x_sign != 0) {
                res = ((Cstar_w0 as i32).wrapping_neg());
            } else {
                res = (Cstar_w0 as i32);
            }
        } else if (exp == 0) {
            if (x_sign != 0) {
                res = ((C1.w[0] as i32).wrapping_neg());
            } else {
                res = (C1.w[0] as i32);
            }
        } else {
            if (x_sign != 0) {
                res = (((C1.w[0].wrapping_mul(bid_ten2k64[exp as usize])) as i32).wrapping_neg());
            } else {
                res = ((C1.w[0].wrapping_mul(bid_ten2k64[exp as usize])) as i32);
            }
        }
    }
    return (res, pfpsf);
}

pub fn bid128_to_int32_xrnint(mut x: BID_UINT128) -> (i32, u32) {
    let mut res: i32 = 0;
    let mut pfpsf: u32 = 0;
    let (mut x_sign, mut x_exp, mut C1, mut is_special) = bid128_unpack_for_int(x);
    if is_special {
        pfpsf |= 1;
        return (((-0x80000000) as i32), pfpsf);
    }
    if bid128_is_noncanonical(C1, x) {
        return (0, 0);
    }
    if ((C1.w[1] == 0) && (C1.w[0] == 0)) {
        return (0, 0);
    }
    let (mut q, _) = bid128_nr_digits(C1);
    let mut exp = (((go_checked_shr_u64(x_exp, go_shift_count_u64((49) as u64))) as i64).wrapping_sub(6176));
    if ((q.wrapping_add(exp)) > 10) {
        pfpsf |= 1;
        return (((-0x80000000) as i32), pfpsf);
    } else if ((q.wrapping_add(exp)) == 10) {
        let (mut invalid, mut f) = bid128_check_overflow_10(C1, x_sign, q, 0x500000005, false, 0x4fffffffb, true);
        if invalid {
            return (((-0x80000000) as i32), f);
        }
    }
    if ((q.wrapping_add(exp)) < 0) {
        pfpsf |= 32;
        return (0, pfpsf);
    } else if ((q.wrapping_add(exp)) == 0) {
        let mut ind = (q.wrapping_sub(1));
        if (ind <= 18) {
            if ((C1.w[1] == 0) && (C1.w[0] <= bid_midpoint64[ind as usize])) {
                res = 0;
            } else if (x_sign != 0) {
                res = (-1);
            } else {
                res = 1;
            }
        } else {
            if ((C1.w[1] < bid_midpoint128[(ind.wrapping_sub(19)) as usize].w[1]) || (((C1.w[1] == bid_midpoint128[(ind.wrapping_sub(19)) as usize].w[1]) && (C1.w[0] <= bid_midpoint128[(ind.wrapping_sub(19)) as usize].w[0])))) {
                res = 0;
            } else if (x_sign != 0) {
                res = (-1);
            } else {
                res = 1;
            }
        }
        pfpsf |= 32;
    } else {
        if (exp < 0) {
            let mut ind = (exp.wrapping_neg());
            let (mut Cstar_w0, mut f) = bid128_round_xrnint_common(C1, ind);
            pfpsf |= f;
            if (x_sign != 0) {
                res = ((Cstar_w0 as i32).wrapping_neg());
            } else {
                res = (Cstar_w0 as i32);
            }
        } else if (exp == 0) {
            if (x_sign != 0) {
                res = ((C1.w[0] as i32).wrapping_neg());
            } else {
                res = (C1.w[0] as i32);
            }
        } else {
            if (x_sign != 0) {
                res = (((C1.w[0].wrapping_mul(bid_ten2k64[exp as usize])) as i32).wrapping_neg());
            } else {
                res = ((C1.w[0].wrapping_mul(bid_ten2k64[exp as usize])) as i32);
            }
        }
    }
    return (res, pfpsf);
}

pub fn bid128_to_int32_floor(mut x: BID_UINT128) -> (i32, u32) {
    let mut res: i32 = 0;
    let mut pfpsf: u32 = 0;
    let (mut x_sign, mut x_exp, mut C1, mut is_special) = bid128_unpack_for_int(x);
    if is_special {
        pfpsf |= 1;
        return (((-0x80000000) as i32), pfpsf);
    }
    if bid128_is_noncanonical(C1, x) {
        return (0, 0);
    }
    if ((C1.w[1] == 0) && (C1.w[0] == 0)) {
        return (0, 0);
    }
    let (mut q, _) = bid128_nr_digits(C1);
    let mut exp = (((go_checked_shr_u64(x_exp, go_shift_count_u64((49) as u64))) as i64).wrapping_sub(6176));
    if ((q.wrapping_add(exp)) > 10) {
        pfpsf |= 1;
        return (((-0x80000000) as i32), pfpsf);
    } else if ((q.wrapping_add(exp)) == 10) {
        let (mut invalid, mut f) = bid128_check_overflow_10(C1, x_sign, q, 0x500000000, false, 0x500000000, true);
        if invalid {
            return (((-0x80000000) as i32), f);
        }
    }
    if ((q.wrapping_add(exp)) <= 0) {
        if (x_sign != 0) {
            return ((-1), 0);
        }
        return (0, 0);
    } else {
        if (exp < 0) {
            let mut ind = (exp.wrapping_neg());
            let mut Cstar_w0 = bid128_round_floor_ceil_int_common(C1, ind, x_sign, 0);
            if (x_sign != 0) {
                res = ((Cstar_w0 as i32).wrapping_neg());
            } else {
                res = (Cstar_w0 as i32);
            }
        } else if (exp == 0) {
            if (x_sign != 0) {
                res = ((C1.w[0] as i32).wrapping_neg());
            } else {
                res = (C1.w[0] as i32);
            }
        } else {
            if (x_sign != 0) {
                res = (((C1.w[0].wrapping_mul(bid_ten2k64[exp as usize])) as i32).wrapping_neg());
            } else {
                res = ((C1.w[0].wrapping_mul(bid_ten2k64[exp as usize])) as i32);
            }
        }
    }
    return (res, pfpsf);
}

pub fn bid128_to_int32_xfloor(mut x: BID_UINT128) -> (i32, u32) {
    let mut res: i32 = 0;
    let mut pfpsf: u32 = 0;
    let (mut x_sign, mut x_exp, mut C1, mut is_special) = bid128_unpack_for_int(x);
    if is_special {
        pfpsf |= 1;
        return (((-0x80000000) as i32), pfpsf);
    }
    if bid128_is_noncanonical(C1, x) {
        return (0, 0);
    }
    if ((C1.w[1] == 0) && (C1.w[0] == 0)) {
        return (0, 0);
    }
    let (mut q, _) = bid128_nr_digits(C1);
    let mut exp = (((go_checked_shr_u64(x_exp, go_shift_count_u64((49) as u64))) as i64).wrapping_sub(6176));
    if ((q.wrapping_add(exp)) > 10) {
        pfpsf |= 1;
        return (((-0x80000000) as i32), pfpsf);
    } else if ((q.wrapping_add(exp)) == 10) {
        let (mut invalid, mut f) = bid128_check_overflow_10(C1, x_sign, q, 0x500000000, false, 0x500000000, true);
        if invalid {
            return (((-0x80000000) as i32), f);
        }
    }
    if ((q.wrapping_add(exp)) <= 0) {
        if (x_sign != 0) {
            pfpsf |= 32;
            return ((-1), pfpsf);
        }
        pfpsf |= 32;
        return (0, pfpsf);
    } else {
        if (exp < 0) {
            let mut ind = (exp.wrapping_neg());
            let (mut Cstar_w0, mut f) = bid128_round_xfloor_xceil_xint_common(C1, ind, x_sign, 0);
            pfpsf |= f;
            if (x_sign != 0) {
                res = ((Cstar_w0 as i32).wrapping_neg());
            } else {
                res = (Cstar_w0 as i32);
            }
        } else if (exp == 0) {
            if (x_sign != 0) {
                res = ((C1.w[0] as i32).wrapping_neg());
            } else {
                res = (C1.w[0] as i32);
            }
        } else {
            if (x_sign != 0) {
                res = (((C1.w[0].wrapping_mul(bid_ten2k64[exp as usize])) as i32).wrapping_neg());
            } else {
                res = ((C1.w[0].wrapping_mul(bid_ten2k64[exp as usize])) as i32);
            }
        }
    }
    return (res, pfpsf);
}

pub fn bid128_to_int32_ceil(mut x: BID_UINT128) -> (i32, u32) {
    let mut res: i32 = 0;
    let mut pfpsf: u32 = 0;
    let (mut x_sign, mut x_exp, mut C1, mut is_special) = bid128_unpack_for_int(x);
    if is_special {
        pfpsf |= 1;
        return (((-0x80000000) as i32), pfpsf);
    }
    if bid128_is_noncanonical(C1, x) {
        return (0, 0);
    }
    if ((C1.w[1] == 0) && (C1.w[0] == 0)) {
        return (0, 0);
    }
    let (mut q, _) = bid128_nr_digits(C1);
    let mut exp = (((go_checked_shr_u64(x_exp, go_shift_count_u64((49) as u64))) as i64).wrapping_sub(6176));
    if ((q.wrapping_add(exp)) > 10) {
        pfpsf |= 1;
        return (((-0x80000000) as i32), pfpsf);
    } else if ((q.wrapping_add(exp)) == 10) {
        let (mut invalid, mut f) = bid128_check_overflow_10(C1, x_sign, q, 0x50000000a, true, 0x4fffffff6, false);
        if invalid {
            return (((-0x80000000) as i32), f);
        }
    }
    if ((q.wrapping_add(exp)) <= 0) {
        if (x_sign != 0) {
            return (0, 0);
        }
        return (1, 0);
    } else {
        if (exp < 0) {
            let mut ind = (exp.wrapping_neg());
            let mut Cstar_w0 = bid128_round_floor_ceil_int_common(C1, ind, x_sign, 1);
            if (x_sign != 0) {
                res = ((Cstar_w0 as i32).wrapping_neg());
            } else {
                res = (Cstar_w0 as i32);
            }
        } else if (exp == 0) {
            if (x_sign != 0) {
                res = ((C1.w[0] as i32).wrapping_neg());
            } else {
                res = (C1.w[0] as i32);
            }
        } else {
            if (x_sign != 0) {
                res = (((C1.w[0].wrapping_mul(bid_ten2k64[exp as usize])) as i32).wrapping_neg());
            } else {
                res = ((C1.w[0].wrapping_mul(bid_ten2k64[exp as usize])) as i32);
            }
        }
    }
    return (res, pfpsf);
}

pub fn bid128_to_int32_xceil(mut x: BID_UINT128) -> (i32, u32) {
    let mut res: i32 = 0;
    let mut pfpsf: u32 = 0;
    let (mut x_sign, mut x_exp, mut C1, mut is_special) = bid128_unpack_for_int(x);
    if is_special {
        pfpsf |= 1;
        return (((-0x80000000) as i32), pfpsf);
    }
    if bid128_is_noncanonical(C1, x) {
        return (0, 0);
    }
    if ((C1.w[1] == 0) && (C1.w[0] == 0)) {
        return (0, 0);
    }
    let (mut q, _) = bid128_nr_digits(C1);
    let mut exp = (((go_checked_shr_u64(x_exp, go_shift_count_u64((49) as u64))) as i64).wrapping_sub(6176));
    if ((q.wrapping_add(exp)) > 10) {
        pfpsf |= 1;
        return (((-0x80000000) as i32), pfpsf);
    } else if ((q.wrapping_add(exp)) == 10) {
        let (mut invalid, mut f) = bid128_check_overflow_10(C1, x_sign, q, 0x50000000a, true, 0x4fffffff6, false);
        if invalid {
            return (((-0x80000000) as i32), f);
        }
    }
    if ((q.wrapping_add(exp)) <= 0) {
        if (x_sign != 0) {
            pfpsf |= 32;
            return (0, pfpsf);
        }
        pfpsf |= 32;
        return (1, pfpsf);
    } else {
        if (exp < 0) {
            let mut ind = (exp.wrapping_neg());
            let (mut Cstar_w0, mut f) = bid128_round_xfloor_xceil_xint_common(C1, ind, x_sign, 1);
            pfpsf |= f;
            if (x_sign != 0) {
                res = ((Cstar_w0 as i32).wrapping_neg());
            } else {
                res = (Cstar_w0 as i32);
            }
        } else if (exp == 0) {
            if (x_sign != 0) {
                res = ((C1.w[0] as i32).wrapping_neg());
            } else {
                res = (C1.w[0] as i32);
            }
        } else {
            if (x_sign != 0) {
                res = (((C1.w[0].wrapping_mul(bid_ten2k64[exp as usize])) as i32).wrapping_neg());
            } else {
                res = ((C1.w[0].wrapping_mul(bid_ten2k64[exp as usize])) as i32);
            }
        }
    }
    return (res, pfpsf);
}

pub fn bid128_to_int32_int(mut x: BID_UINT128) -> (i32, u32) {
    let mut res: i32 = 0;
    let mut pfpsf: u32 = 0;
    let (mut x_sign, mut x_exp, mut C1, mut is_special) = bid128_unpack_for_int(x);
    if is_special {
        pfpsf |= 1;
        return (((-0x80000000) as i32), pfpsf);
    }
    if bid128_is_noncanonical(C1, x) {
        return (0, 0);
    }
    if ((C1.w[1] == 0) && (C1.w[0] == 0)) {
        return (0, 0);
    }
    let (mut q, _) = bid128_nr_digits(C1);
    let mut exp = (((go_checked_shr_u64(x_exp, go_shift_count_u64((49) as u64))) as i64).wrapping_sub(6176));
    if ((q.wrapping_add(exp)) > 10) {
        pfpsf |= 1;
        return (((-0x80000000) as i32), pfpsf);
    } else if ((q.wrapping_add(exp)) == 10) {
        let (mut invalid, mut f) = bid128_check_overflow_10(C1, x_sign, q, 0x50000000a, true, 0x500000000, true);
        if invalid {
            return (((-0x80000000) as i32), f);
        }
    }
    if ((q.wrapping_add(exp)) <= 0) {
        return (0, 0);
    } else {
        if (exp < 0) {
            let mut ind = (exp.wrapping_neg());
            let mut Cstar_w0 = bid128_round_floor_ceil_int_common(C1, ind, x_sign, 2);
            if (x_sign != 0) {
                res = ((Cstar_w0 as i32).wrapping_neg());
            } else {
                res = (Cstar_w0 as i32);
            }
        } else if (exp == 0) {
            if (x_sign != 0) {
                res = ((C1.w[0] as i32).wrapping_neg());
            } else {
                res = (C1.w[0] as i32);
            }
        } else {
            if (x_sign != 0) {
                res = (((C1.w[0].wrapping_mul(bid_ten2k64[exp as usize])) as i32).wrapping_neg());
            } else {
                res = ((C1.w[0].wrapping_mul(bid_ten2k64[exp as usize])) as i32);
            }
        }
    }
    return (res, pfpsf);
}

pub fn bid128_to_int32_xint(mut x: BID_UINT128) -> (i32, u32) {
    let mut res: i32 = 0;
    let mut pfpsf: u32 = 0;
    let (mut x_sign, mut x_exp, mut C1, mut is_special) = bid128_unpack_for_int(x);
    if is_special {
        pfpsf |= 1;
        return (((-0x80000000) as i32), pfpsf);
    }
    if bid128_is_noncanonical(C1, x) {
        return (0, 0);
    }
    if ((C1.w[1] == 0) && (C1.w[0] == 0)) {
        return (0, 0);
    }
    let (mut q, _) = bid128_nr_digits(C1);
    let mut exp = (((go_checked_shr_u64(x_exp, go_shift_count_u64((49) as u64))) as i64).wrapping_sub(6176));
    if ((q.wrapping_add(exp)) > 10) {
        pfpsf |= 1;
        return (((-0x80000000) as i32), pfpsf);
    } else if ((q.wrapping_add(exp)) == 10) {
        let (mut invalid, mut f) = bid128_check_overflow_10(C1, x_sign, q, 0x50000000a, true, 0x500000000, true);
        if invalid {
            return (((-0x80000000) as i32), f);
        }
    }
    if ((q.wrapping_add(exp)) <= 0) {
        pfpsf |= 32;
        return (0, pfpsf);
    } else {
        if (exp < 0) {
            let mut ind = (exp.wrapping_neg());
            let (mut Cstar_w0, mut f) = bid128_round_xfloor_xceil_xint_common(C1, ind, x_sign, 2);
            pfpsf |= f;
            if (x_sign != 0) {
                res = ((Cstar_w0 as i32).wrapping_neg());
            } else {
                res = (Cstar_w0 as i32);
            }
        } else if (exp == 0) {
            if (x_sign != 0) {
                res = ((C1.w[0] as i32).wrapping_neg());
            } else {
                res = (C1.w[0] as i32);
            }
        } else {
            if (x_sign != 0) {
                res = (((C1.w[0].wrapping_mul(bid_ten2k64[exp as usize])) as i32).wrapping_neg());
            } else {
                res = ((C1.w[0].wrapping_mul(bid_ten2k64[exp as usize])) as i32);
            }
        }
    }
    return (res, pfpsf);
}

pub fn bid128_to_int32_rninta(mut x: BID_UINT128) -> (i32, u32) {
    let mut res: i32 = 0;
    let mut pfpsf: u32 = 0;
    let (mut x_sign, mut x_exp, mut C1, mut is_special) = bid128_unpack_for_int(x);
    if is_special {
        pfpsf |= 1;
        return (((-0x80000000) as i32), pfpsf);
    }
    if bid128_is_noncanonical(C1, x) {
        return (0, 0);
    }
    if ((C1.w[1] == 0) && (C1.w[0] == 0)) {
        return (0, 0);
    }
    let (mut q, _) = bid128_nr_digits(C1);
    let mut exp = (((go_checked_shr_u64(x_exp, go_shift_count_u64((49) as u64))) as i64).wrapping_sub(6176));
    if ((q.wrapping_add(exp)) > 10) {
        pfpsf |= 1;
        return (((-0x80000000) as i32), pfpsf);
    } else if ((q.wrapping_add(exp)) == 10) {
        let (mut invalid, mut f) = bid128_check_overflow_10(C1, x_sign, q, 0x500000005, true, 0x4fffffffb, true);
        if invalid {
            return (((-0x80000000) as i32), f);
        }
    }
    if ((q.wrapping_add(exp)) < 0) {
        return (0, 0);
    } else if ((q.wrapping_add(exp)) == 0) {
        let mut ind = (q.wrapping_sub(1));
        if (ind <= 18) {
            if ((C1.w[1] == 0) && (C1.w[0] < bid_midpoint64[ind as usize])) {
                res = 0;
            } else if (x_sign != 0) {
                res = (-1);
            } else {
                res = 1;
            }
        } else {
            if ((C1.w[1] < bid_midpoint128[(ind.wrapping_sub(19)) as usize].w[1]) || (((C1.w[1] == bid_midpoint128[(ind.wrapping_sub(19)) as usize].w[1]) && (C1.w[0] < bid_midpoint128[(ind.wrapping_sub(19)) as usize].w[0])))) {
                res = 0;
            } else if (x_sign != 0) {
                res = (-1);
            } else {
                res = 1;
            }
        }
    } else {
        if (exp < 0) {
            let mut ind = (exp.wrapping_neg());
            let mut Cstar_w0 = bid128_round_rninta_common(C1, ind);
            if (x_sign != 0) {
                res = ((Cstar_w0 as i32).wrapping_neg());
            } else {
                res = (Cstar_w0 as i32);
            }
        } else if (exp == 0) {
            if (x_sign != 0) {
                res = ((C1.w[0] as i32).wrapping_neg());
            } else {
                res = (C1.w[0] as i32);
            }
        } else {
            if (x_sign != 0) {
                res = (((C1.w[0].wrapping_mul(bid_ten2k64[exp as usize])) as i32).wrapping_neg());
            } else {
                res = ((C1.w[0].wrapping_mul(bid_ten2k64[exp as usize])) as i32);
            }
        }
    }
    return (res, pfpsf);
}

pub fn bid128_to_int32_xrninta(mut x: BID_UINT128) -> (i32, u32) {
    let mut res: i32 = 0;
    let mut pfpsf: u32 = 0;
    let (mut x_sign, mut x_exp, mut C1, mut is_special) = bid128_unpack_for_int(x);
    if is_special {
        pfpsf |= 1;
        return (((-0x80000000) as i32), pfpsf);
    }
    if bid128_is_noncanonical(C1, x) {
        return (0, 0);
    }
    if ((C1.w[1] == 0) && (C1.w[0] == 0)) {
        return (0, 0);
    }
    let (mut q, _) = bid128_nr_digits(C1);
    let mut exp = (((go_checked_shr_u64(x_exp, go_shift_count_u64((49) as u64))) as i64).wrapping_sub(6176));
    if ((q.wrapping_add(exp)) > 10) {
        pfpsf |= 1;
        return (((-0x80000000) as i32), pfpsf);
    } else if ((q.wrapping_add(exp)) == 10) {
        let (mut invalid, mut f) = bid128_check_overflow_10(C1, x_sign, q, 0x500000005, true, 0x4fffffffb, true);
        if invalid {
            return (((-0x80000000) as i32), f);
        }
    }
    if ((q.wrapping_add(exp)) < 0) {
        pfpsf |= 32;
        return (0, pfpsf);
    } else if ((q.wrapping_add(exp)) == 0) {
        let mut ind = (q.wrapping_sub(1));
        if (ind <= 18) {
            if ((C1.w[1] == 0) && (C1.w[0] < bid_midpoint64[ind as usize])) {
                res = 0;
            } else if (x_sign != 0) {
                res = (-1);
            } else {
                res = 1;
            }
        } else {
            if ((C1.w[1] < bid_midpoint128[(ind.wrapping_sub(19)) as usize].w[1]) || (((C1.w[1] == bid_midpoint128[(ind.wrapping_sub(19)) as usize].w[1]) && (C1.w[0] < bid_midpoint128[(ind.wrapping_sub(19)) as usize].w[0])))) {
                res = 0;
            } else if (x_sign != 0) {
                res = (-1);
            } else {
                res = 1;
            }
        }
        pfpsf |= 32;
    } else {
        if (exp < 0) {
            let mut ind = (exp.wrapping_neg());
            let (mut Cstar_w0, mut f) = bid128_round_xrninta_common(C1, ind);
            pfpsf |= f;
            if (x_sign != 0) {
                res = ((Cstar_w0 as i32).wrapping_neg());
            } else {
                res = (Cstar_w0 as i32);
            }
        } else if (exp == 0) {
            if (x_sign != 0) {
                res = ((C1.w[0] as i32).wrapping_neg());
            } else {
                res = (C1.w[0] as i32);
            }
        } else {
            if (x_sign != 0) {
                res = (((C1.w[0].wrapping_mul(bid_ten2k64[exp as usize])) as i32).wrapping_neg());
            } else {
                res = ((C1.w[0].wrapping_mul(bid_ten2k64[exp as usize])) as i32);
            }
        }
    }
    return (res, pfpsf);
}

pub(crate) fn bid128_check_overflow_19(mut x: BID_UINT128, mut x_sign: u64, mut C1: BID_UINT128, mut q: i64, mut neg_hi: u64, mut neg_lo: u64, mut neg_cmp_ge: bool, mut pos_hi: u64, mut pos_lo: u64, mut pos_cmp_ge: bool) -> (bool, u32, BID_UINT128) {
    let mut C: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut pfpsf: u32 = 0;
    if (x_sign != 0) {
        C.w[1] = neg_hi;
        C.w[0] = neg_lo;
        if (q <= 19) {
            C1 = __mul_64x64_to_128(C1.w[0], bid_ten2k64[((20 as i64).wrapping_sub(q)) as usize]);
        } else if (q == 20) {
        } else {
            C = __mul_128x64_to_128(bid_ten2k64[(q.wrapping_sub(20)) as usize], C);
        }
        if neg_cmp_ge {
            if ((C1.w[1] > C.w[1]) || (((C1.w[1] == C.w[1]) && (C1.w[0] >= C.w[0])))) {
                pfpsf |= 1;
                return (true, pfpsf, C1);
            }
        } else {
            if ((C1.w[1] > C.w[1]) || (((C1.w[1] == C.w[1]) && (C1.w[0] > C.w[0])))) {
                pfpsf |= 1;
                return (true, pfpsf, C1);
            }
        }
    } else {
        C.w[1] = pos_hi;
        C.w[0] = pos_lo;
        if (q <= 19) {
            C1 = __mul_64x64_to_128(C1.w[0], bid_ten2k64[((20 as i64).wrapping_sub(q)) as usize]);
        } else if (q == 20) {
        } else {
            C = __mul_128x64_to_128(bid_ten2k64[(q.wrapping_sub(20)) as usize], C);
        }
        if pos_cmp_ge {
            if ((C1.w[1] > C.w[1]) || (((C1.w[1] == C.w[1]) && (C1.w[0] >= C.w[0])))) {
                pfpsf |= 1;
                return (true, pfpsf, C1);
            }
        } else {
            if ((C1.w[1] > C.w[1]) || (((C1.w[1] == C.w[1]) && (C1.w[0] > C.w[0])))) {
                pfpsf |= 1;
                return (true, pfpsf, C1);
            }
        }
    }
    return (false, 0, C1);
}

pub(crate) fn bid128_cmp_128(mut a: BID_UINT128, mut b: BID_UINT128) -> i64 {
    if (a.w[1] < b.w[1]) {
        return (-1);
    }
    if (a.w[1] > b.w[1]) {
        return 1;
    }
    if (a.w[0] < b.w[0]) {
        return (-1);
    }
    if (a.w[0] > b.w[0]) {
        return 1;
    }
    return 0;
}

pub(crate) fn bid128_check_overflow_20(mut C1: BID_UINT128, mut x_sign: u64, mut q: i64, mut neg_hi: u64, mut neg_lo: u64, mut neg_cmp_ge: bool, mut pos_hi: u64, mut pos_lo: u64, mut pos_cmp_ge: bool) -> (bool, u32) {
    let mut scaled: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut limit: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut cmp: i64 = 0;
    let mut pfpsf: u32 = 0;
    limit.w[1] = pos_hi;
    limit.w[0] = pos_lo;
    if (x_sign != 0) {
        limit.w[1] = neg_hi;
        limit.w[0] = neg_lo;
    }
    if (q < 21) {
        if (q == 1) {
            scaled = __mul_128x64_to_128(C1.w[0], bid_ten2k128[0]);
        } else if (q <= 19) {
            scaled = __mul_64x64_to_128(C1.w[0], bid_ten2k64[((21 as i64).wrapping_sub(q)) as usize]);
        } else {
            scaled = __mul_128x64_to_128(bid_ten2k64[1], C1);
        }
        cmp = bid128_cmp_128(scaled, limit);
    } else if (q == 21) {
        cmp = bid128_cmp_128(C1, limit);
    } else {
        limit = __mul_128x64_to_128(bid_ten2k64[(q.wrapping_sub(21)) as usize], limit);
        cmp = bid128_cmp_128(C1, limit);
    }
    if (x_sign != 0) {
        if ((cmp > 0) || ((neg_cmp_ge && (cmp == 0)))) {
            pfpsf |= 1;
            return (true, pfpsf);
        }
    } else {
        if ((cmp > 0) || ((pos_cmp_ge && (cmp == 0)))) {
            pfpsf |= 1;
            return (true, pfpsf);
        }
    }
    return (false, 0);
}

pub(crate) fn bid128_int64_core(mut x: BID_UINT128, mut roundMode: i64, mut setInexact: bool, mut maxQE: i64, mut invalidRes: i64, mut negHi: u64, mut negLo: u64, mut negCmpGe: bool, mut posHi: u64, mut posLo: u64, mut posCmpGe: bool, mut qpExpEqHandler: fn(u64, BID_UINT128, i64, i64) -> (i64, u32), mut qpExpLeqHandler: fn(u64, BID_UINT128, i64, i64) -> (i64, u32)) -> (i64, u32) {
    let mut pfpsf: u32 = 0;
    let (mut x_sign, mut x_exp, mut C1, mut is_special) = bid128_unpack_for_int(x);
    if is_special {
        pfpsf |= 1;
        return (invalidRes, pfpsf);
    }
    if bid128_is_noncanonical(C1, x) {
        return (0, 0);
    }
    if ((C1.w[1] == 0) && (C1.w[0] == 0)) {
        return (0, 0);
    }
    let (mut q, _) = bid128_nr_digits(C1);
    let mut exp = (((go_checked_shr_u64(x_exp, go_shift_count_u64((49) as u64))) as i64).wrapping_sub(6176));
    if ((q.wrapping_add(exp)) > maxQE) {
        pfpsf |= 1;
        return (invalidRes, pfpsf);
    } else if ((q.wrapping_add(exp)) == maxQE) {
        let (mut invalid, mut f, mut newC1) = bid128_check_overflow_19(x, x_sign, C1, q, negHi, negLo, negCmpGe, posHi, posLo, posCmpGe);
        if invalid {
            return (invalidRes, f);
        }
        C1 = newC1;
        C1.w[1] = (x.w[1] & 0x1ffffffffffff);
        C1.w[0] = x.w[0];
    }
    return qpExpLeqHandler(x_sign, C1, q, exp);
}

pub fn bid128_to_int64_rnint(mut x: BID_UINT128) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut pfpsf: u32 = 0;
    let (mut x_sign, mut x_exp, mut C1, mut is_special) = bid128_unpack_for_int(x);
    if is_special {
        pfpsf |= 1;
        return ((-0x8000000000000000), pfpsf);
    }
    if bid128_is_noncanonical(C1, x) {
        return (0, 0);
    }
    if ((C1.w[1] == 0) && (C1.w[0] == 0)) {
        return (0, 0);
    }
    let (mut q, _) = bid128_nr_digits(C1);
    let mut exp = (((go_checked_shr_u64(x_exp, go_shift_count_u64((49) as u64))) as i64).wrapping_sub(6176));
    if ((q.wrapping_add(exp)) > 19) {
        pfpsf |= 1;
        return ((-0x8000000000000000), pfpsf);
    } else if ((q.wrapping_add(exp)) == 19) {
        let (mut invalid, mut f, _) = bid128_check_overflow_19(x, x_sign, C1, q, 0x0000000000000005, 0x0000000000000005, false, 0x0000000000000004, 0xfffffffffffffffb, true);
        if invalid {
            return ((-0x8000000000000000), f);
        }
        C1.w[1] = (x.w[1] & 0x1ffffffffffff);
        C1.w[0] = x.w[0];
    }
    if ((q.wrapping_add(exp)) < 0) {
        return (0, 0);
    } else if ((q.wrapping_add(exp)) == 0) {
        let mut ind = (q.wrapping_sub(1));
        if (ind <= 18) {
            if ((C1.w[1] == 0) && (C1.w[0] <= bid_midpoint64[ind as usize])) {
                res = 0;
            } else if (x_sign != 0) {
                res = (-1);
            } else {
                res = 1;
            }
        } else {
            if ((C1.w[1] < bid_midpoint128[(ind.wrapping_sub(19)) as usize].w[1]) || (((C1.w[1] == bid_midpoint128[(ind.wrapping_sub(19)) as usize].w[1]) && (C1.w[0] <= bid_midpoint128[(ind.wrapping_sub(19)) as usize].w[0])))) {
                res = 0;
            } else if (x_sign != 0) {
                res = (-1);
            } else {
                res = 1;
            }
        }
    } else {
        if (exp < 0) {
            let mut ind = (exp.wrapping_neg());
            let mut Cstar_w0 = bid128_round_rnint_common(C1, ind);
            if (x_sign != 0) {
                res = ((Cstar_w0 as i64).wrapping_neg());
            } else {
                res = (Cstar_w0 as i64);
            }
        } else if (exp == 0) {
            if (x_sign != 0) {
                res = ((C1.w[0] as i64).wrapping_neg());
            } else {
                res = (C1.w[0] as i64);
            }
        } else {
            if (x_sign != 0) {
                res = (((C1.w[0].wrapping_mul(bid_ten2k64[exp as usize])) as i64).wrapping_neg());
            } else {
                res = ((C1.w[0].wrapping_mul(bid_ten2k64[exp as usize])) as i64);
            }
        }
    }
    return (res, pfpsf);
}

pub fn bid128_to_int64_xrnint(mut x: BID_UINT128) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut pfpsf: u32 = 0;
    let (mut x_sign, mut x_exp, mut C1, mut is_special) = bid128_unpack_for_int(x);
    if is_special {
        pfpsf |= 1;
        return ((-0x8000000000000000), pfpsf);
    }
    if bid128_is_noncanonical(C1, x) {
        return (0, 0);
    }
    if ((C1.w[1] == 0) && (C1.w[0] == 0)) {
        return (0, 0);
    }
    let (mut q, _) = bid128_nr_digits(C1);
    let mut exp = (((go_checked_shr_u64(x_exp, go_shift_count_u64((49) as u64))) as i64).wrapping_sub(6176));
    if ((q.wrapping_add(exp)) > 19) {
        pfpsf |= 1;
        return ((-0x8000000000000000), pfpsf);
    } else if ((q.wrapping_add(exp)) == 19) {
        let (mut invalid, mut f, _) = bid128_check_overflow_19(x, x_sign, C1, q, 0x0000000000000005, 0x0000000000000005, false, 0x0000000000000004, 0xfffffffffffffffb, true);
        if invalid {
            return ((-0x8000000000000000), f);
        }
        C1.w[1] = (x.w[1] & 0x1ffffffffffff);
        C1.w[0] = x.w[0];
    }
    if ((q.wrapping_add(exp)) < 0) {
        pfpsf |= 32;
        return (0, pfpsf);
    } else if ((q.wrapping_add(exp)) == 0) {
        let mut ind = (q.wrapping_sub(1));
        if (ind <= 18) {
            if ((C1.w[1] == 0) && (C1.w[0] <= bid_midpoint64[ind as usize])) {
                res = 0;
            } else if (x_sign != 0) {
                res = (-1);
            } else {
                res = 1;
            }
        } else {
            if ((C1.w[1] < bid_midpoint128[(ind.wrapping_sub(19)) as usize].w[1]) || (((C1.w[1] == bid_midpoint128[(ind.wrapping_sub(19)) as usize].w[1]) && (C1.w[0] <= bid_midpoint128[(ind.wrapping_sub(19)) as usize].w[0])))) {
                res = 0;
            } else if (x_sign != 0) {
                res = (-1);
            } else {
                res = 1;
            }
        }
        pfpsf |= 32;
    } else {
        if (exp < 0) {
            let mut ind = (exp.wrapping_neg());
            let (mut Cstar_w0, mut f) = bid128_round_xrnint_common(C1, ind);
            pfpsf |= f;
            if (x_sign != 0) {
                res = ((Cstar_w0 as i64).wrapping_neg());
            } else {
                res = (Cstar_w0 as i64);
            }
        } else if (exp == 0) {
            if (x_sign != 0) {
                res = ((C1.w[0] as i64).wrapping_neg());
            } else {
                res = (C1.w[0] as i64);
            }
        } else {
            if (x_sign != 0) {
                res = (((C1.w[0].wrapping_mul(bid_ten2k64[exp as usize])) as i64).wrapping_neg());
            } else {
                res = ((C1.w[0].wrapping_mul(bid_ten2k64[exp as usize])) as i64);
            }
        }
    }
    return (res, pfpsf);
}

pub fn bid128_to_int64_floor(mut x: BID_UINT128) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut pfpsf: u32 = 0;
    let (mut x_sign, mut x_exp, mut C1, mut is_special) = bid128_unpack_for_int(x);
    if is_special {
        pfpsf |= 1;
        return ((-0x8000000000000000), pfpsf);
    }
    if bid128_is_noncanonical(C1, x) {
        return (0, 0);
    }
    if ((C1.w[1] == 0) && (C1.w[0] == 0)) {
        return (0, 0);
    }
    let (mut q, _) = bid128_nr_digits(C1);
    let mut exp = (((go_checked_shr_u64(x_exp, go_shift_count_u64((49) as u64))) as i64).wrapping_sub(6176));
    if ((q.wrapping_add(exp)) > 19) {
        pfpsf |= 1;
        return ((-0x8000000000000000), pfpsf);
    } else if ((q.wrapping_add(exp)) == 19) {
        let (mut invalid, mut f, _) = bid128_check_overflow_19(x, x_sign, C1, q, 0x0000000000000005, 0x0000000000000000, false, 0x0000000000000005, 0x0000000000000000, true);
        if invalid {
            return ((-0x8000000000000000), f);
        }
        C1.w[1] = (x.w[1] & 0x1ffffffffffff);
        C1.w[0] = x.w[0];
    }
    if ((q.wrapping_add(exp)) <= 0) {
        if (x_sign != 0) {
            return ((-1), 0);
        }
        return (0, 0);
    } else {
        if (exp < 0) {
            let mut ind = (exp.wrapping_neg());
            let (mut Cstar_w0, _) = bid128_round_trunc_mode_common(C1, ind, x_sign, 0, false);
            if (x_sign != 0) {
                res = ((Cstar_w0 as i64).wrapping_neg());
            } else {
                res = (Cstar_w0 as i64);
            }
        } else if (exp == 0) {
            if (x_sign != 0) {
                res = ((C1.w[0] as i64).wrapping_neg());
            } else {
                res = (C1.w[0] as i64);
            }
        } else {
            if (x_sign != 0) {
                res = (((C1.w[0].wrapping_mul(bid_ten2k64[exp as usize])) as i64).wrapping_neg());
            } else {
                res = ((C1.w[0].wrapping_mul(bid_ten2k64[exp as usize])) as i64);
            }
        }
    }
    return (res, pfpsf);
}

pub fn bid128_to_int64_xfloor(mut x: BID_UINT128) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut pfpsf: u32 = 0;
    let (mut x_sign, mut x_exp, mut C1, mut is_special) = bid128_unpack_for_int(x);
    if is_special {
        pfpsf |= 1;
        return ((-0x8000000000000000), pfpsf);
    }
    if bid128_is_noncanonical(C1, x) {
        return (0, 0);
    }
    if ((C1.w[1] == 0) && (C1.w[0] == 0)) {
        return (0, 0);
    }
    let (mut q, _) = bid128_nr_digits(C1);
    let mut exp = (((go_checked_shr_u64(x_exp, go_shift_count_u64((49) as u64))) as i64).wrapping_sub(6176));
    if ((q.wrapping_add(exp)) > 19) {
        pfpsf |= 1;
        return ((-0x8000000000000000), pfpsf);
    } else if ((q.wrapping_add(exp)) == 19) {
        let (mut invalid, mut f, _) = bid128_check_overflow_19(x, x_sign, C1, q, 0x0000000000000005, 0x0000000000000000, false, 0x0000000000000005, 0x0000000000000000, true);
        if invalid {
            return ((-0x8000000000000000), f);
        }
        C1.w[1] = (x.w[1] & 0x1ffffffffffff);
        C1.w[0] = x.w[0];
    }
    if ((q.wrapping_add(exp)) <= 0) {
        pfpsf |= 32;
        if (x_sign != 0) {
            return ((-1), pfpsf);
        }
        return (0, pfpsf);
    } else {
        if (exp < 0) {
            let mut ind = (exp.wrapping_neg());
            let (mut Cstar_w0, mut f) = bid128_round_trunc_mode_common(C1, ind, x_sign, 0, true);
            pfpsf |= f;
            if (x_sign != 0) {
                res = ((Cstar_w0 as i64).wrapping_neg());
            } else {
                res = (Cstar_w0 as i64);
            }
        } else if (exp == 0) {
            if (x_sign != 0) {
                res = ((C1.w[0] as i64).wrapping_neg());
            } else {
                res = (C1.w[0] as i64);
            }
        } else {
            if (x_sign != 0) {
                res = (((C1.w[0].wrapping_mul(bid_ten2k64[exp as usize])) as i64).wrapping_neg());
            } else {
                res = ((C1.w[0].wrapping_mul(bid_ten2k64[exp as usize])) as i64);
            }
        }
    }
    return (res, pfpsf);
}

pub fn bid128_to_int64_ceil(mut x: BID_UINT128) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut pfpsf: u32 = 0;
    let (mut x_sign, mut x_exp, mut C1, mut is_special) = bid128_unpack_for_int(x);
    if is_special {
        pfpsf |= 1;
        return ((-0x8000000000000000), pfpsf);
    }
    if bid128_is_noncanonical(C1, x) {
        return (0, 0);
    }
    if ((C1.w[1] == 0) && (C1.w[0] == 0)) {
        return (0, 0);
    }
    let (mut q, _) = bid128_nr_digits(C1);
    let mut exp = (((go_checked_shr_u64(x_exp, go_shift_count_u64((49) as u64))) as i64).wrapping_sub(6176));
    if ((q.wrapping_add(exp)) > 19) {
        pfpsf |= 1;
        return ((-0x8000000000000000), pfpsf);
    } else if ((q.wrapping_add(exp)) == 19) {
        let (mut invalid, mut f, _) = bid128_check_overflow_19(x, x_sign, C1, q, 0x0000000000000005, 0x000000000000000a, true, 0x0000000000000004, 0xfffffffffffffff6, false);
        if invalid {
            return ((-0x8000000000000000), f);
        }
        C1.w[1] = (x.w[1] & 0x1ffffffffffff);
        C1.w[0] = x.w[0];
    }
    if ((q.wrapping_add(exp)) <= 0) {
        if (x_sign != 0) {
            return (0, 0);
        }
        return (1, 0);
    } else {
        if (exp < 0) {
            let mut ind = (exp.wrapping_neg());
            let (mut Cstar_w0, _) = bid128_round_trunc_mode_common(C1, ind, x_sign, 1, false);
            if (x_sign != 0) {
                res = ((Cstar_w0 as i64).wrapping_neg());
            } else {
                res = (Cstar_w0 as i64);
            }
        } else if (exp == 0) {
            if (x_sign != 0) {
                res = ((C1.w[0] as i64).wrapping_neg());
            } else {
                res = (C1.w[0] as i64);
            }
        } else {
            if (x_sign != 0) {
                res = (((C1.w[0].wrapping_mul(bid_ten2k64[exp as usize])) as i64).wrapping_neg());
            } else {
                res = ((C1.w[0].wrapping_mul(bid_ten2k64[exp as usize])) as i64);
            }
        }
    }
    return (res, pfpsf);
}

pub fn bid128_to_int64_xceil(mut x: BID_UINT128) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut pfpsf: u32 = 0;
    let (mut x_sign, mut x_exp, mut C1, mut is_special) = bid128_unpack_for_int(x);
    if is_special {
        pfpsf |= 1;
        return ((-0x8000000000000000), pfpsf);
    }
    if bid128_is_noncanonical(C1, x) {
        return (0, 0);
    }
    if ((C1.w[1] == 0) && (C1.w[0] == 0)) {
        return (0, 0);
    }
    let (mut q, _) = bid128_nr_digits(C1);
    let mut exp = (((go_checked_shr_u64(x_exp, go_shift_count_u64((49) as u64))) as i64).wrapping_sub(6176));
    if ((q.wrapping_add(exp)) > 19) {
        pfpsf |= 1;
        return ((-0x8000000000000000), pfpsf);
    } else if ((q.wrapping_add(exp)) == 19) {
        let (mut invalid, mut f, _) = bid128_check_overflow_19(x, x_sign, C1, q, 0x0000000000000005, 0x000000000000000a, true, 0x0000000000000004, 0xfffffffffffffff6, false);
        if invalid {
            return ((-0x8000000000000000), f);
        }
        C1.w[1] = (x.w[1] & 0x1ffffffffffff);
        C1.w[0] = x.w[0];
    }
    if ((q.wrapping_add(exp)) <= 0) {
        pfpsf |= 32;
        if (x_sign != 0) {
            return (0, pfpsf);
        }
        return (1, pfpsf);
    } else {
        if (exp < 0) {
            let mut ind = (exp.wrapping_neg());
            let (mut Cstar_w0, mut f) = bid128_round_trunc_mode_common(C1, ind, x_sign, 1, true);
            pfpsf |= f;
            if (x_sign != 0) {
                res = ((Cstar_w0 as i64).wrapping_neg());
            } else {
                res = (Cstar_w0 as i64);
            }
        } else if (exp == 0) {
            if (x_sign != 0) {
                res = ((C1.w[0] as i64).wrapping_neg());
            } else {
                res = (C1.w[0] as i64);
            }
        } else {
            if (x_sign != 0) {
                res = (((C1.w[0].wrapping_mul(bid_ten2k64[exp as usize])) as i64).wrapping_neg());
            } else {
                res = ((C1.w[0].wrapping_mul(bid_ten2k64[exp as usize])) as i64);
            }
        }
    }
    return (res, pfpsf);
}

pub fn bid128_to_int64_int(mut x: BID_UINT128) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut pfpsf: u32 = 0;
    let (mut x_sign, mut x_exp, mut C1, mut is_special) = bid128_unpack_for_int(x);
    if is_special {
        pfpsf |= 1;
        return ((-0x8000000000000000), pfpsf);
    }
    if bid128_is_noncanonical(C1, x) {
        return (0, 0);
    }
    if ((C1.w[1] == 0) && (C1.w[0] == 0)) {
        return (0, 0);
    }
    let (mut q, _) = bid128_nr_digits(C1);
    let mut exp = (((go_checked_shr_u64(x_exp, go_shift_count_u64((49) as u64))) as i64).wrapping_sub(6176));
    if ((q.wrapping_add(exp)) > 19) {
        pfpsf |= 1;
        return ((-0x8000000000000000), pfpsf);
    } else if ((q.wrapping_add(exp)) == 19) {
        let (mut invalid, mut f, _) = bid128_check_overflow_19(x, x_sign, C1, q, 0x0000000000000005, 0x000000000000000a, true, 0x0000000000000005, 0x0000000000000000, true);
        if invalid {
            return ((-0x8000000000000000), f);
        }
        C1.w[1] = (x.w[1] & 0x1ffffffffffff);
        C1.w[0] = x.w[0];
    }
    if ((q.wrapping_add(exp)) <= 0) {
        return (0, 0);
    } else {
        if (exp < 0) {
            let mut ind = (exp.wrapping_neg());
            let (mut Cstar_w0, _) = bid128_round_trunc_mode_common(C1, ind, x_sign, 2, false);
            if (x_sign != 0) {
                res = ((Cstar_w0 as i64).wrapping_neg());
            } else {
                res = (Cstar_w0 as i64);
            }
        } else if (exp == 0) {
            if (x_sign != 0) {
                res = ((C1.w[0] as i64).wrapping_neg());
            } else {
                res = (C1.w[0] as i64);
            }
        } else {
            if (x_sign != 0) {
                res = (((C1.w[0].wrapping_mul(bid_ten2k64[exp as usize])) as i64).wrapping_neg());
            } else {
                res = ((C1.w[0].wrapping_mul(bid_ten2k64[exp as usize])) as i64);
            }
        }
    }
    return (res, pfpsf);
}

pub fn bid128_to_int64_xint(mut x: BID_UINT128) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut pfpsf: u32 = 0;
    let (mut x_sign, mut x_exp, mut C1, mut is_special) = bid128_unpack_for_int(x);
    if is_special {
        pfpsf |= 1;
        return ((-0x8000000000000000), pfpsf);
    }
    if bid128_is_noncanonical(C1, x) {
        return (0, 0);
    }
    if ((C1.w[1] == 0) && (C1.w[0] == 0)) {
        return (0, 0);
    }
    let (mut q, _) = bid128_nr_digits(C1);
    let mut exp = (((go_checked_shr_u64(x_exp, go_shift_count_u64((49) as u64))) as i64).wrapping_sub(6176));
    if ((q.wrapping_add(exp)) > 19) {
        pfpsf |= 1;
        return ((-0x8000000000000000), pfpsf);
    } else if ((q.wrapping_add(exp)) == 19) {
        let (mut invalid, mut f, _) = bid128_check_overflow_19(x, x_sign, C1, q, 0x0000000000000005, 0x000000000000000a, true, 0x0000000000000005, 0x0000000000000000, true);
        if invalid {
            return ((-0x8000000000000000), f);
        }
        C1.w[1] = (x.w[1] & 0x1ffffffffffff);
        C1.w[0] = x.w[0];
    }
    if ((q.wrapping_add(exp)) <= 0) {
        pfpsf |= 32;
        return (0, pfpsf);
    } else {
        if (exp < 0) {
            let mut ind = (exp.wrapping_neg());
            let (mut Cstar_w0, mut f) = bid128_round_trunc_mode_common(C1, ind, x_sign, 2, true);
            pfpsf |= f;
            if (x_sign != 0) {
                res = ((Cstar_w0 as i64).wrapping_neg());
            } else {
                res = (Cstar_w0 as i64);
            }
        } else if (exp == 0) {
            if (x_sign != 0) {
                res = ((C1.w[0] as i64).wrapping_neg());
            } else {
                res = (C1.w[0] as i64);
            }
        } else {
            if (x_sign != 0) {
                res = (((C1.w[0].wrapping_mul(bid_ten2k64[exp as usize])) as i64).wrapping_neg());
            } else {
                res = ((C1.w[0].wrapping_mul(bid_ten2k64[exp as usize])) as i64);
            }
        }
    }
    return (res, pfpsf);
}

pub fn bid128_to_int64_rninta(mut x: BID_UINT128) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut pfpsf: u32 = 0;
    let (mut x_sign, mut x_exp, mut C1, mut is_special) = bid128_unpack_for_int(x);
    if is_special {
        pfpsf |= 1;
        return ((-0x8000000000000000), pfpsf);
    }
    if bid128_is_noncanonical(C1, x) {
        return (0, 0);
    }
    if ((C1.w[1] == 0) && (C1.w[0] == 0)) {
        return (0, 0);
    }
    let (mut q, _) = bid128_nr_digits(C1);
    let mut exp = (((go_checked_shr_u64(x_exp, go_shift_count_u64((49) as u64))) as i64).wrapping_sub(6176));
    if ((q.wrapping_add(exp)) > 19) {
        pfpsf |= 1;
        return ((-0x8000000000000000), pfpsf);
    } else if ((q.wrapping_add(exp)) == 19) {
        let (mut invalid, mut f, _) = bid128_check_overflow_19(x, x_sign, C1, q, 0x0000000000000005, 0x0000000000000005, true, 0x0000000000000004, 0xfffffffffffffffb, true);
        if invalid {
            return ((-0x8000000000000000), f);
        }
        C1.w[1] = (x.w[1] & 0x1ffffffffffff);
        C1.w[0] = x.w[0];
    }
    if ((q.wrapping_add(exp)) < 0) {
        return (0, 0);
    } else if ((q.wrapping_add(exp)) == 0) {
        let mut ind = (q.wrapping_sub(1));
        if (ind <= 18) {
            if ((C1.w[1] == 0) && (C1.w[0] < bid_midpoint64[ind as usize])) {
                res = 0;
            } else if (x_sign != 0) {
                res = (-1);
            } else {
                res = 1;
            }
        } else {
            if ((C1.w[1] < bid_midpoint128[(ind.wrapping_sub(19)) as usize].w[1]) || (((C1.w[1] == bid_midpoint128[(ind.wrapping_sub(19)) as usize].w[1]) && (C1.w[0] < bid_midpoint128[(ind.wrapping_sub(19)) as usize].w[0])))) {
                res = 0;
            } else if (x_sign != 0) {
                res = (-1);
            } else {
                res = 1;
            }
        }
    } else {
        if (exp < 0) {
            let mut ind = (exp.wrapping_neg());
            let mut Cstar_w0 = bid128_round_rninta_common(C1, ind);
            if (x_sign != 0) {
                res = ((Cstar_w0 as i64).wrapping_neg());
            } else {
                res = (Cstar_w0 as i64);
            }
        } else if (exp == 0) {
            if (x_sign != 0) {
                res = ((C1.w[0] as i64).wrapping_neg());
            } else {
                res = (C1.w[0] as i64);
            }
        } else {
            if (x_sign != 0) {
                res = (((C1.w[0].wrapping_mul(bid_ten2k64[exp as usize])) as i64).wrapping_neg());
            } else {
                res = ((C1.w[0].wrapping_mul(bid_ten2k64[exp as usize])) as i64);
            }
        }
    }
    return (res, pfpsf);
}

pub fn bid128_to_int64_xrninta(mut x: BID_UINT128) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut pfpsf: u32 = 0;
    let (mut x_sign, mut x_exp, mut C1, mut is_special) = bid128_unpack_for_int(x);
    if is_special {
        pfpsf |= 1;
        return ((-0x8000000000000000), pfpsf);
    }
    if bid128_is_noncanonical(C1, x) {
        return (0, 0);
    }
    if ((C1.w[1] == 0) && (C1.w[0] == 0)) {
        return (0, 0);
    }
    let (mut q, _) = bid128_nr_digits(C1);
    let mut exp = (((go_checked_shr_u64(x_exp, go_shift_count_u64((49) as u64))) as i64).wrapping_sub(6176));
    if ((q.wrapping_add(exp)) > 19) {
        pfpsf |= 1;
        return ((-0x8000000000000000), pfpsf);
    } else if ((q.wrapping_add(exp)) == 19) {
        let (mut invalid, mut f, _) = bid128_check_overflow_19(x, x_sign, C1, q, 0x0000000000000005, 0x0000000000000005, true, 0x0000000000000004, 0xfffffffffffffffb, true);
        if invalid {
            return ((-0x8000000000000000), f);
        }
        C1.w[1] = (x.w[1] & 0x1ffffffffffff);
        C1.w[0] = x.w[0];
    }
    if ((q.wrapping_add(exp)) < 0) {
        pfpsf |= 32;
        return (0, pfpsf);
    } else if ((q.wrapping_add(exp)) == 0) {
        let mut ind = (q.wrapping_sub(1));
        if (ind <= 18) {
            if ((C1.w[1] == 0) && (C1.w[0] < bid_midpoint64[ind as usize])) {
                res = 0;
            } else if (x_sign != 0) {
                res = (-1);
            } else {
                res = 1;
            }
        } else {
            if ((C1.w[1] < bid_midpoint128[(ind.wrapping_sub(19)) as usize].w[1]) || (((C1.w[1] == bid_midpoint128[(ind.wrapping_sub(19)) as usize].w[1]) && (C1.w[0] < bid_midpoint128[(ind.wrapping_sub(19)) as usize].w[0])))) {
                res = 0;
            } else if (x_sign != 0) {
                res = (-1);
            } else {
                res = 1;
            }
        }
        pfpsf |= 32;
    } else {
        if (exp < 0) {
            let mut ind = (exp.wrapping_neg());
            let (mut Cstar_w0, mut f) = bid128_round_xrninta_common(C1, ind);
            pfpsf |= f;
            if (x_sign != 0) {
                res = ((Cstar_w0 as i64).wrapping_neg());
            } else {
                res = (Cstar_w0 as i64);
            }
        } else if (exp == 0) {
            if (x_sign != 0) {
                res = ((C1.w[0] as i64).wrapping_neg());
            } else {
                res = (C1.w[0] as i64);
            }
        } else {
            if (x_sign != 0) {
                res = (((C1.w[0].wrapping_mul(bid_ten2k64[exp as usize])) as i64).wrapping_neg());
            } else {
                res = ((C1.w[0].wrapping_mul(bid_ten2k64[exp as usize])) as i64);
            }
        }
    }
    return (res, pfpsf);
}

pub(crate) fn bid128_uint_rnint_qpexp0(mut C1: BID_UINT128, mut x_sign: u64, mut q: i64) -> (u32, u32) {
    let mut ind = (q.wrapping_sub(1));
    if (ind <= 18) {
        if ((C1.w[1] == 0) && (C1.w[0] <= bid_midpoint64[ind as usize])) {
            return (0, 0);
        } else if (x_sign == 0) {
            return (1, 0);
        } else {
            return (0x80000000, 1);
        }
    } else {
        if ((C1.w[1] < bid_midpoint128[(ind.wrapping_sub(19)) as usize].w[1]) || (((C1.w[1] == bid_midpoint128[(ind.wrapping_sub(19)) as usize].w[1]) && (C1.w[0] <= bid_midpoint128[(ind.wrapping_sub(19)) as usize].w[0])))) {
            return (0, 0);
        } else if (x_sign == 0) {
            return (1, 0);
        } else {
            return (0x80000000, 1);
        }
    }
}

pub fn bid128_to_uint32_rnint(mut x: BID_UINT128) -> (u32, u32) {
    let mut pfpsf: u32 = 0;
    let (mut x_sign, mut x_exp, mut C1, mut is_special) = bid128_unpack_for_int(x);
    if is_special {
        pfpsf |= 1;
        return (0x80000000, pfpsf);
    }
    if bid128_is_noncanonical(C1, x) {
        return (0, 0);
    }
    if ((C1.w[1] == 0) && (C1.w[0] == 0)) {
        return (0, 0);
    }
    let (mut q, _) = bid128_nr_digits(C1);
    let mut exp = (((go_checked_shr_u64(x_exp, go_shift_count_u64((49) as u64))) as i64).wrapping_sub(6176));
    if ((q.wrapping_add(exp)) > 10) {
        pfpsf |= 1;
        return (0x80000000, pfpsf);
    } else if ((q.wrapping_add(exp)) == 10) {
        let (mut invalid, mut f) = bid128_check_overflow_10(C1, x_sign, q, 0x05, false, 0x9fffffffb, true);
        if invalid {
            return (0x80000000, f);
        }
    }
    if ((q.wrapping_add(exp)) < 0) {
        return (0, 0);
    } else if ((q.wrapping_add(exp)) == 0) {
        return bid128_uint_rnint_qpexp0(C1, x_sign, q);
    } else {
        if (x_sign != 0) {
            pfpsf |= 1;
            return (0x80000000, pfpsf);
        }
        if (exp < 0) {
            let mut ind = (exp.wrapping_neg());
            let mut Cstar_w0 = bid128_round_rnint_common(C1, ind);
            return ((Cstar_w0 as u32), pfpsf);
        } else if (exp == 0) {
            return ((C1.w[0] as u32), 0);
        } else {
            return (((C1.w[0].wrapping_mul(bid_ten2k64[exp as usize])) as u32), 0);
        }
    }
}

pub fn bid128_to_uint32_xrnint(mut x: BID_UINT128) -> (u32, u32) {
    let mut pfpsf: u32 = 0;
    let (mut x_sign, mut x_exp, mut C1, mut is_special) = bid128_unpack_for_int(x);
    if is_special {
        pfpsf |= 1;
        return (0x80000000, pfpsf);
    }
    if bid128_is_noncanonical(C1, x) {
        return (0, 0);
    }
    if ((C1.w[1] == 0) && (C1.w[0] == 0)) {
        return (0, 0);
    }
    let (mut q, _) = bid128_nr_digits(C1);
    let mut exp = (((go_checked_shr_u64(x_exp, go_shift_count_u64((49) as u64))) as i64).wrapping_sub(6176));
    if ((q.wrapping_add(exp)) > 10) {
        pfpsf |= 1;
        return (0x80000000, pfpsf);
    } else if ((q.wrapping_add(exp)) == 10) {
        let (mut invalid, mut f) = bid128_check_overflow_10(C1, x_sign, q, 0x05, false, 0x9fffffffb, true);
        if invalid {
            return (0x80000000, f);
        }
    }
    if ((q.wrapping_add(exp)) < 0) {
        pfpsf |= 32;
        return (0, pfpsf);
    } else if ((q.wrapping_add(exp)) == 0) {
        let (mut r, mut f) = bid128_uint_rnint_qpexp0(C1, x_sign, q);
        if (f != 0) {
            return (r, f);
        }
        return (r, 32);
    } else {
        if (x_sign != 0) {
            pfpsf |= 1;
            return (0x80000000, pfpsf);
        }
        if (exp < 0) {
            let mut ind = (exp.wrapping_neg());
            let (mut Cstar_w0, mut f) = bid128_round_xrnint_common(C1, ind);
            pfpsf |= f;
            return ((Cstar_w0 as u32), pfpsf);
        } else if (exp == 0) {
            return ((C1.w[0] as u32), 0);
        } else {
            return (((C1.w[0].wrapping_mul(bid_ten2k64[exp as usize])) as u32), 0);
        }
    }
}

pub fn bid128_to_uint32_floor(mut x: BID_UINT128) -> (u32, u32) {
    let mut pfpsf: u32 = 0;
    let (mut x_sign, mut x_exp, mut C1, mut is_special) = bid128_unpack_for_int(x);
    if is_special {
        pfpsf |= 1;
        return (0x80000000, pfpsf);
    }
    if bid128_is_noncanonical(C1, x) {
        return (0, 0);
    }
    if ((C1.w[1] == 0) && (C1.w[0] == 0)) {
        return (0, 0);
    }
    let (mut q, _) = bid128_nr_digits(C1);
    let mut exp = (((go_checked_shr_u64(x_exp, go_shift_count_u64((49) as u64))) as i64).wrapping_sub(6176));
    if ((q.wrapping_add(exp)) > 10) {
        pfpsf |= 1;
        return (0x80000000, pfpsf);
    } else if ((q.wrapping_add(exp)) == 10) {
        let (mut invalid, mut f) = bid128_check_overflow_10(C1, x_sign, q, 0x0a, true, 0xa00000000, true);
        if invalid {
            return (0x80000000, f);
        }
    }
    if ((q.wrapping_add(exp)) <= 0) {
        if (x_sign != 0) {
            pfpsf |= 1;
            return (0x80000000, pfpsf);
        }
        return (0, 0);
    } else {
        if (x_sign != 0) {
            pfpsf |= 1;
            return (0x80000000, pfpsf);
        }
        if (exp < 0) {
            let mut ind = (exp.wrapping_neg());
            let mut Cstar_w0 = bid128_round_floor_ceil_int_common(C1, ind, x_sign, 0);
            return ((Cstar_w0 as u32), pfpsf);
        } else if (exp == 0) {
            return ((C1.w[0] as u32), 0);
        } else {
            return (((C1.w[0].wrapping_mul(bid_ten2k64[exp as usize])) as u32), 0);
        }
    }
}

pub fn bid128_to_uint32_xfloor(mut x: BID_UINT128) -> (u32, u32) {
    let mut pfpsf: u32 = 0;
    let (mut x_sign, mut x_exp, mut C1, mut is_special) = bid128_unpack_for_int(x);
    if is_special {
        pfpsf |= 1;
        return (0x80000000, pfpsf);
    }
    if bid128_is_noncanonical(C1, x) {
        return (0, 0);
    }
    if ((C1.w[1] == 0) && (C1.w[0] == 0)) {
        return (0, 0);
    }
    let (mut q, _) = bid128_nr_digits(C1);
    let mut exp = (((go_checked_shr_u64(x_exp, go_shift_count_u64((49) as u64))) as i64).wrapping_sub(6176));
    if ((q.wrapping_add(exp)) > 10) {
        pfpsf |= 1;
        return (0x80000000, pfpsf);
    } else if ((q.wrapping_add(exp)) == 10) {
        let (mut invalid, mut f) = bid128_check_overflow_10(C1, x_sign, q, 0x0a, true, 0xa00000000, true);
        if invalid {
            return (0x80000000, f);
        }
    }
    if ((q.wrapping_add(exp)) <= 0) {
        if (x_sign != 0) {
            pfpsf |= 1;
            return (0x80000000, pfpsf);
        }
        pfpsf |= 32;
        return (0, pfpsf);
    } else {
        if (x_sign != 0) {
            pfpsf |= 1;
            return (0x80000000, pfpsf);
        }
        if (exp < 0) {
            let mut ind = (exp.wrapping_neg());
            let (mut Cstar_w0, mut f) = bid128_round_xfloor_xceil_xint_common(C1, ind, x_sign, 0);
            pfpsf |= f;
            return ((Cstar_w0 as u32), pfpsf);
        } else if (exp == 0) {
            return ((C1.w[0] as u32), 0);
        } else {
            return (((C1.w[0].wrapping_mul(bid_ten2k64[exp as usize])) as u32), 0);
        }
    }
}

pub fn bid128_to_uint32_ceil(mut x: BID_UINT128) -> (u32, u32) {
    let mut pfpsf: u32 = 0;
    let (mut x_sign, mut x_exp, mut C1, mut is_special) = bid128_unpack_for_int(x);
    if is_special {
        pfpsf |= 1;
        return (0x80000000, pfpsf);
    }
    if bid128_is_noncanonical(C1, x) {
        return (0, 0);
    }
    if ((C1.w[1] == 0) && (C1.w[0] == 0)) {
        return (0, 0);
    }
    let (mut q, _) = bid128_nr_digits(C1);
    let mut exp = (((go_checked_shr_u64(x_exp, go_shift_count_u64((49) as u64))) as i64).wrapping_sub(6176));
    if ((q.wrapping_add(exp)) > 10) {
        pfpsf |= 1;
        return (0x80000000, pfpsf);
    } else if ((q.wrapping_add(exp)) == 10) {
        let (mut invalid, mut f) = bid128_check_overflow_10(C1, x_sign, q, 0x0a, true, 0x9fffffff6, false);
        if invalid {
            return (0x80000000, f);
        }
    }
    if ((q.wrapping_add(exp)) <= 0) {
        if (x_sign != 0) {
            return (0, 0);
        }
        return (1, 0);
    } else {
        if (x_sign != 0) {
            pfpsf |= 1;
            return (0x80000000, pfpsf);
        }
        if (exp < 0) {
            let mut ind = (exp.wrapping_neg());
            let mut Cstar_w0 = bid128_round_floor_ceil_int_common(C1, ind, x_sign, 1);
            return ((Cstar_w0 as u32), pfpsf);
        } else if (exp == 0) {
            return ((C1.w[0] as u32), 0);
        } else {
            return (((C1.w[0].wrapping_mul(bid_ten2k64[exp as usize])) as u32), 0);
        }
    }
}

pub fn bid128_to_uint32_xceil(mut x: BID_UINT128) -> (u32, u32) {
    let mut pfpsf: u32 = 0;
    let (mut x_sign, mut x_exp, mut C1, mut is_special) = bid128_unpack_for_int(x);
    if is_special {
        pfpsf |= 1;
        return (0x80000000, pfpsf);
    }
    if bid128_is_noncanonical(C1, x) {
        return (0, 0);
    }
    if ((C1.w[1] == 0) && (C1.w[0] == 0)) {
        return (0, 0);
    }
    let (mut q, _) = bid128_nr_digits(C1);
    let mut exp = (((go_checked_shr_u64(x_exp, go_shift_count_u64((49) as u64))) as i64).wrapping_sub(6176));
    if ((q.wrapping_add(exp)) > 10) {
        pfpsf |= 1;
        return (0x80000000, pfpsf);
    } else if ((q.wrapping_add(exp)) == 10) {
        let (mut invalid, mut f) = bid128_check_overflow_10(C1, x_sign, q, 0x0a, true, 0x9fffffff6, false);
        if invalid {
            return (0x80000000, f);
        }
    }
    if ((q.wrapping_add(exp)) <= 0) {
        if (x_sign != 0) {
            pfpsf |= 32;
            return (0, pfpsf);
        }
        pfpsf |= 32;
        return (1, pfpsf);
    } else {
        if (x_sign != 0) {
            pfpsf |= 1;
            return (0x80000000, pfpsf);
        }
        if (exp < 0) {
            let mut ind = (exp.wrapping_neg());
            let (mut Cstar_w0, mut f) = bid128_round_xfloor_xceil_xint_common(C1, ind, x_sign, 1);
            pfpsf |= f;
            return ((Cstar_w0 as u32), pfpsf);
        } else if (exp == 0) {
            return ((C1.w[0] as u32), 0);
        } else {
            return (((C1.w[0].wrapping_mul(bid_ten2k64[exp as usize])) as u32), 0);
        }
    }
}

pub fn bid128_to_uint32_int(mut x: BID_UINT128) -> (u32, u32) {
    let mut pfpsf: u32 = 0;
    let (mut x_sign, mut x_exp, mut C1, mut is_special) = bid128_unpack_for_int(x);
    if is_special {
        pfpsf |= 1;
        return (0x80000000, pfpsf);
    }
    if bid128_is_noncanonical(C1, x) {
        return (0, 0);
    }
    if ((C1.w[1] == 0) && (C1.w[0] == 0)) {
        return (0, 0);
    }
    let (mut q, _) = bid128_nr_digits(C1);
    let mut exp = (((go_checked_shr_u64(x_exp, go_shift_count_u64((49) as u64))) as i64).wrapping_sub(6176));
    if ((q.wrapping_add(exp)) > 10) {
        pfpsf |= 1;
        return (0x80000000, pfpsf);
    } else if ((q.wrapping_add(exp)) == 10) {
        let (mut invalid, mut f) = bid128_check_overflow_10(C1, x_sign, q, 0x0a, true, 0xa00000000, true);
        if invalid {
            return (0x80000000, f);
        }
    }
    if ((q.wrapping_add(exp)) <= 0) {
        if (x_sign != 0) {
            return (0, 0);
        }
        return (0, 0);
    } else {
        if (x_sign != 0) {
            pfpsf |= 1;
            return (0x80000000, pfpsf);
        }
        if (exp < 0) {
            let mut ind = (exp.wrapping_neg());
            let mut Cstar_w0 = bid128_round_floor_ceil_int_common(C1, ind, x_sign, 2);
            return ((Cstar_w0 as u32), pfpsf);
        } else if (exp == 0) {
            return ((C1.w[0] as u32), 0);
        } else {
            return (((C1.w[0].wrapping_mul(bid_ten2k64[exp as usize])) as u32), 0);
        }
    }
}

pub fn bid128_to_uint32_xint(mut x: BID_UINT128) -> (u32, u32) {
    let mut pfpsf: u32 = 0;
    let (mut x_sign, mut x_exp, mut C1, mut is_special) = bid128_unpack_for_int(x);
    if is_special {
        pfpsf |= 1;
        return (0x80000000, pfpsf);
    }
    if bid128_is_noncanonical(C1, x) {
        return (0, 0);
    }
    if ((C1.w[1] == 0) && (C1.w[0] == 0)) {
        return (0, 0);
    }
    let (mut q, _) = bid128_nr_digits(C1);
    let mut exp = (((go_checked_shr_u64(x_exp, go_shift_count_u64((49) as u64))) as i64).wrapping_sub(6176));
    if ((q.wrapping_add(exp)) > 10) {
        pfpsf |= 1;
        return (0x80000000, pfpsf);
    } else if ((q.wrapping_add(exp)) == 10) {
        let (mut invalid, mut f) = bid128_check_overflow_10(C1, x_sign, q, 0x0a, true, 0xa00000000, true);
        if invalid {
            return (0x80000000, f);
        }
    }
    if ((q.wrapping_add(exp)) <= 0) {
        if (x_sign != 0) {
            pfpsf |= 32;
            return (0, pfpsf);
        }
        pfpsf |= 32;
        return (0, pfpsf);
    } else {
        if (x_sign != 0) {
            pfpsf |= 1;
            return (0x80000000, pfpsf);
        }
        if (exp < 0) {
            let mut ind = (exp.wrapping_neg());
            let (mut Cstar_w0, mut f) = bid128_round_xfloor_xceil_xint_common(C1, ind, x_sign, 2);
            pfpsf |= f;
            return ((Cstar_w0 as u32), pfpsf);
        } else if (exp == 0) {
            return ((C1.w[0] as u32), 0);
        } else {
            return (((C1.w[0].wrapping_mul(bid_ten2k64[exp as usize])) as u32), 0);
        }
    }
}

pub fn bid128_to_uint32_rninta(mut x: BID_UINT128) -> (u32, u32) {
    let mut pfpsf: u32 = 0;
    let (mut x_sign, mut x_exp, mut C1, mut is_special) = bid128_unpack_for_int(x);
    if is_special {
        pfpsf |= 1;
        return (0x80000000, pfpsf);
    }
    if bid128_is_noncanonical(C1, x) {
        return (0, 0);
    }
    if ((C1.w[1] == 0) && (C1.w[0] == 0)) {
        return (0, 0);
    }
    let (mut q, _) = bid128_nr_digits(C1);
    let mut exp = (((go_checked_shr_u64(x_exp, go_shift_count_u64((49) as u64))) as i64).wrapping_sub(6176));
    if ((q.wrapping_add(exp)) > 10) {
        pfpsf |= 1;
        return (0x80000000, pfpsf);
    } else if ((q.wrapping_add(exp)) == 10) {
        let (mut invalid, mut f) = bid128_check_overflow_10(C1, x_sign, q, 0x05, true, 0x9fffffffb, true);
        if invalid {
            return (0x80000000, f);
        }
    }
    if ((q.wrapping_add(exp)) < 0) {
        return (0, 0);
    } else if ((q.wrapping_add(exp)) == 0) {
        let mut ind = (q.wrapping_sub(1));
        if (ind <= 18) {
            if ((C1.w[1] == 0) && (C1.w[0] < bid_midpoint64[ind as usize])) {
                return (0, 0);
            } else if (x_sign == 0) {
                return (1, 0);
            } else {
                return (0x80000000, 1);
            }
        } else {
            if ((C1.w[1] < bid_midpoint128[(ind.wrapping_sub(19)) as usize].w[1]) || (((C1.w[1] == bid_midpoint128[(ind.wrapping_sub(19)) as usize].w[1]) && (C1.w[0] < bid_midpoint128[(ind.wrapping_sub(19)) as usize].w[0])))) {
                return (0, 0);
            } else if (x_sign == 0) {
                return (1, 0);
            } else {
                return (0x80000000, 1);
            }
        }
    } else {
        if (x_sign != 0) {
            pfpsf |= 1;
            return (0x80000000, pfpsf);
        }
        if (exp < 0) {
            let mut ind = (exp.wrapping_neg());
            let mut Cstar_w0 = bid128_round_rninta_common(C1, ind);
            return ((Cstar_w0 as u32), pfpsf);
        } else if (exp == 0) {
            return ((C1.w[0] as u32), 0);
        } else {
            return (((C1.w[0].wrapping_mul(bid_ten2k64[exp as usize])) as u32), 0);
        }
    }
}

pub fn bid128_to_uint32_xrninta(mut x: BID_UINT128) -> (u32, u32) {
    let mut pfpsf: u32 = 0;
    let (mut x_sign, mut x_exp, mut C1, mut is_special) = bid128_unpack_for_int(x);
    if is_special {
        pfpsf |= 1;
        return (0x80000000, pfpsf);
    }
    if bid128_is_noncanonical(C1, x) {
        return (0, 0);
    }
    if ((C1.w[1] == 0) && (C1.w[0] == 0)) {
        return (0, 0);
    }
    let (mut q, _) = bid128_nr_digits(C1);
    let mut exp = (((go_checked_shr_u64(x_exp, go_shift_count_u64((49) as u64))) as i64).wrapping_sub(6176));
    if ((q.wrapping_add(exp)) > 10) {
        pfpsf |= 1;
        return (0x80000000, pfpsf);
    } else if ((q.wrapping_add(exp)) == 10) {
        let (mut invalid, mut f) = bid128_check_overflow_10(C1, x_sign, q, 0x05, true, 0x9fffffffb, true);
        if invalid {
            return (0x80000000, f);
        }
    }
    if ((q.wrapping_add(exp)) < 0) {
        pfpsf |= 32;
        return (0, pfpsf);
    } else if ((q.wrapping_add(exp)) == 0) {
        let mut ind = (q.wrapping_sub(1));
        if (ind <= 18) {
            if ((C1.w[1] == 0) && (C1.w[0] < bid_midpoint64[ind as usize])) {
                return (0, 32);
            } else if (x_sign == 0) {
                return (1, 32);
            } else {
                return (0x80000000, 1);
            }
        } else {
            if ((C1.w[1] < bid_midpoint128[(ind.wrapping_sub(19)) as usize].w[1]) || (((C1.w[1] == bid_midpoint128[(ind.wrapping_sub(19)) as usize].w[1]) && (C1.w[0] < bid_midpoint128[(ind.wrapping_sub(19)) as usize].w[0])))) {
                return (0, 32);
            } else if (x_sign == 0) {
                return (1, 32);
            } else {
                return (0x80000000, 1);
            }
        }
    } else {
        if (x_sign != 0) {
            pfpsf |= 1;
            return (0x80000000, pfpsf);
        }
        if (exp < 0) {
            let mut ind = (exp.wrapping_neg());
            let (mut Cstar_w0, mut f) = bid128_round_xrninta_common(C1, ind);
            pfpsf |= f;
            return ((Cstar_w0 as u32), pfpsf);
        } else if (exp == 0) {
            return ((C1.w[0] as u32), 0);
        } else {
            return (((C1.w[0].wrapping_mul(bid_ten2k64[exp as usize])) as u32), 0);
        }
    }
}

pub fn bid128_to_uint64_rnint(mut x: BID_UINT128) -> (u64, u32) {
    return bid128_to_uint64_core(x, 0, false);
}

pub fn bid128_to_uint64_xrnint(mut x: BID_UINT128) -> (u64, u32) {
    return bid128_to_uint64_core(x, 0, true);
}

pub fn bid128_to_uint64_floor(mut x: BID_UINT128) -> (u64, u32) {
    return bid128_to_uint64_core(x, 1, false);
}

pub fn bid128_to_uint64_xfloor(mut x: BID_UINT128) -> (u64, u32) {
    return bid128_to_uint64_core(x, 1, true);
}

pub fn bid128_to_uint64_ceil(mut x: BID_UINT128) -> (u64, u32) {
    return bid128_to_uint64_core(x, 2, false);
}

pub fn bid128_to_uint64_xceil(mut x: BID_UINT128) -> (u64, u32) {
    return bid128_to_uint64_core(x, 2, true);
}

pub fn bid128_to_uint64_int(mut x: BID_UINT128) -> (u64, u32) {
    return bid128_to_uint64_core(x, 3, false);
}

pub fn bid128_to_uint64_xint(mut x: BID_UINT128) -> (u64, u32) {
    return bid128_to_uint64_core(x, 3, true);
}

pub fn bid128_to_uint64_rninta(mut x: BID_UINT128) -> (u64, u32) {
    return bid128_to_uint64_core(x, 4, false);
}

pub fn bid128_to_uint64_xrninta(mut x: BID_UINT128) -> (u64, u32) {
    return bid128_to_uint64_core(x, 4, true);
}

pub(crate) fn bid128_uint_midpoint_cmp(mut C1: BID_UINT128, mut q: i64) -> i64 {
    let mut ind = (q.wrapping_sub(1));
    if (ind <= 18) {
        if (C1.w[1] == 0) {
            if (C1.w[0] < bid_midpoint64[ind as usize]) {
                return (-1);
            }
            if (C1.w[0] > bid_midpoint64[ind as usize]) {
                return 1;
            }
            return 0;
        }
        return 1;
    }
    if (C1.w[1] < bid_midpoint128[(ind.wrapping_sub(19)) as usize].w[1]) {
        return (-1);
    }
    if (C1.w[1] > bid_midpoint128[(ind.wrapping_sub(19)) as usize].w[1]) {
        return 1;
    }
    if (C1.w[0] < bid_midpoint128[(ind.wrapping_sub(19)) as usize].w[0]) {
        return (-1);
    }
    if (C1.w[0] > bid_midpoint128[(ind.wrapping_sub(19)) as usize].w[0]) {
        return 1;
    }
    return 0;
}

pub(crate) fn bid128_to_uint64_core(mut x: BID_UINT128, mut mode: i64, mut setInexact: bool) -> (u64, u32) {
    let mut pfpsf: u32 = 0;
    let (mut x_sign, mut x_exp, mut C1, mut is_special) = bid128_unpack_for_int(x);
    if is_special {
        pfpsf |= 1;
        return (0x8000000000000000, pfpsf);
    }
    if bid128_is_noncanonical(C1, x) {
        return (0, 0);
    }
    if ((C1.w[1] == 0) && (C1.w[0] == 0)) {
        return (0, 0);
    }
    let (mut q, _) = bid128_nr_digits(C1);
    let mut exp = (((go_checked_shr_u64(x_exp, go_shift_count_u64((49) as u64))) as i64).wrapping_sub(6176));
    if ((q.wrapping_add(exp)) > 20) {
        pfpsf |= 1;
        return (0x8000000000000000, pfpsf);
    } else if ((q.wrapping_add(exp)) == 20) {
        let mut invalid: bool = false;
        match mode {
            0 => {
                (invalid, pfpsf) = bid128_check_overflow_20(C1, x_sign, q, 0x0000000000000000, 0x0000000000000005, false, 0x0000000000000009, 0xfffffffffffffffb, true);
            }
            1 => {
                (invalid, pfpsf) = bid128_check_overflow_20(C1, x_sign, q, 0x0000000000000000, 0x000000000000000a, true, 0x000000000000000a, 0x0000000000000000, true);
            }
            2 => {
                (invalid, pfpsf) = bid128_check_overflow_20(C1, x_sign, q, 0x0000000000000000, 0x000000000000000a, true, 0x0000000000000009, 0xfffffffffffffff6, false);
            }
            3 => {
                (invalid, pfpsf) = bid128_check_overflow_20(C1, x_sign, q, 0x0000000000000000, 0x000000000000000a, true, 0x000000000000000a, 0x0000000000000000, true);
            }
            _ => {
                (invalid, pfpsf) = bid128_check_overflow_20(C1, x_sign, q, 0x0000000000000000, 0x0000000000000005, true, 0x0000000000000009, 0xfffffffffffffffb, true);
            }
        }
        if invalid {
            return (0x8000000000000000, pfpsf);
        }
    }
    if ((q.wrapping_add(exp)) < 0) {
        match mode {
            0 | 4 => {
                if setInexact {
                    pfpsf |= 32;
                }
                return (0, pfpsf);
            }
            1 => {
                if (x_sign != 0) {
                    pfpsf |= 1;
                    return (0x8000000000000000, pfpsf);
                }
                if setInexact {
                    pfpsf |= 32;
                }
                return (0, pfpsf);
            }
            2 => {
                if (x_sign != 0) {
                    if setInexact {
                        pfpsf |= 32;
                    }
                    return (0, pfpsf);
                }
                if setInexact {
                    pfpsf |= 32;
                }
                return (1, pfpsf);
            }
            3 => {
                if setInexact {
                    pfpsf |= 32;
                }
                return (0, pfpsf);
            }
            _ => {}
        }
    }
    if ((q.wrapping_add(exp)) == 0) {
        let mut cmp = bid128_uint_midpoint_cmp(C1, q);
        match mode {
            0 => {
                if (cmp <= 0) {
                    if setInexact {
                        pfpsf |= 32;
                    }
                    return (0, pfpsf);
                }
                if (x_sign == 0) {
                    if setInexact {
                        pfpsf |= 32;
                    }
                    return (1, pfpsf);
                }
                pfpsf |= 1;
                return (0x8000000000000000, pfpsf);
            }
            1 => {
                if (x_sign != 0) {
                    pfpsf |= 1;
                    return (0x8000000000000000, pfpsf);
                }
                if setInexact {
                    pfpsf |= 32;
                }
                return (0, pfpsf);
            }
            2 => {
                if (x_sign != 0) {
                    if setInexact {
                        pfpsf |= 32;
                    }
                    return (0, pfpsf);
                }
                if setInexact {
                    pfpsf |= 32;
                }
                return (1, pfpsf);
            }
            3 => {
                if setInexact {
                    pfpsf |= 32;
                }
                return (0, pfpsf);
            }
            _ => {
                if (cmp < 0) {
                    if setInexact {
                        pfpsf |= 32;
                    }
                    return (0, pfpsf);
                }
                if (x_sign == 0) {
                    if setInexact {
                        pfpsf |= 32;
                    }
                    return (1, pfpsf);
                }
                pfpsf |= 1;
                return (0x8000000000000000, pfpsf);
            }
        }
    }
    if (x_sign != 0) {
        pfpsf |= 1;
        return (0x8000000000000000, pfpsf);
    }
    if (exp < 0) {
        let mut ind = (exp.wrapping_neg());
        match mode {
            0 => {
                if setInexact {
                    let (mut res, mut f) = bid128_round_xrnint_common(C1, ind);
                    return (res, f);
                }
                return (bid128_round_rnint_common(C1, ind), 0);
            }
            1 => {
                let (mut res, mut f) = bid128_round_trunc_mode_common(C1, ind, 0, 0, setInexact);
                return (res, f);
            }
            2 => {
                let (mut res, mut f) = bid128_round_trunc_mode_common(C1, ind, 0, 1, setInexact);
                return (res, f);
            }
            3 => {
                let (mut res, mut f) = bid128_round_trunc_mode_common(C1, ind, 0, 2, setInexact);
                return (res, f);
            }
            _ => {
                if setInexact {
                    let (mut res, mut f) = bid128_round_xrninta_common(C1, ind);
                    return (res, f);
                }
                return (bid128_round_rninta_common(C1, ind), 0);
            }
        }
    } else if (exp == 0) {
        return (C1.w[0], 0);
    }
    return ((C1.w[0].wrapping_mul(bid_ten2k64[exp as usize])), 0);
}

pub(crate) fn bid128_to_small_int(mut r#fn: fn(BID_UINT128) -> (i32, u32), mut x: BID_UINT128, mut sizeMask: i32, mut invalidResult: i8) -> (i8, u32) {
    let (mut v, mut f) = r#fn(x);
    if ((f & 1) != 0) {
        return (invalidResult, f);
    }
    if (((v & sizeMask) != 0) && ((v & sizeMask) != sizeMask)) {
        return (invalidResult, 1);
    }
    return ((v as i8), f);
}

pub(crate) fn bid128_to_small_uint(mut r#fn: fn(BID_UINT128) -> (u32, u32), mut x: BID_UINT128, mut sizeMask: u32, mut invalidResult: u8) -> (u8, u32) {
    let (mut v, mut f) = r#fn(x);
    if ((f & 1) != 0) {
        return (invalidResult, f);
    }
    if ((v & sizeMask) != 0) {
        return (invalidResult, 1);
    }
    return ((v as u8), f);
}

pub(crate) fn bid128_to_small_int16(mut r#fn: fn(BID_UINT128) -> (i32, u32), mut x: BID_UINT128, mut sizeMask: i32, mut invalidResult: i16) -> (i16, u32) {
    let (mut v, mut f) = r#fn(x);
    if ((f & 1) != 0) {
        return (invalidResult, f);
    }
    if (((v & sizeMask) != 0) && ((v & sizeMask) != sizeMask)) {
        return (invalidResult, 1);
    }
    return ((v as i16), f);
}

pub(crate) fn bid128_to_small_uint16(mut r#fn: fn(BID_UINT128) -> (u32, u32), mut x: BID_UINT128, mut sizeMask: u32, mut invalidResult: u16) -> (u16, u32) {
    let (mut v, mut f) = r#fn(x);
    if ((f & 1) != 0) {
        return (invalidResult, f);
    }
    if ((v & sizeMask) != 0) {
        return (invalidResult, 1);
    }
    return ((v as u16), f);
}

pub fn bid128_to_int8_rnint(mut x: BID_UINT128) -> (i8, u32) {
    return bid128_to_small_int(bid128_to_int32_rnint, x, (-128), (-128));
}

pub fn bid128_to_int8_xrnint(mut x: BID_UINT128) -> (i8, u32) {
    return bid128_to_small_int(bid128_to_int32_xrnint, x, (-128), (-128));
}

pub fn bid128_to_int8_rninta(mut x: BID_UINT128) -> (i8, u32) {
    return bid128_to_small_int(bid128_to_int32_rninta, x, (-128), (-128));
}

pub fn bid128_to_int8_xrninta(mut x: BID_UINT128) -> (i8, u32) {
    return bid128_to_small_int(bid128_to_int32_xrninta, x, (-128), (-128));
}

pub fn bid128_to_int8_int(mut x: BID_UINT128) -> (i8, u32) {
    return bid128_to_small_int(bid128_to_int32_int, x, (-128), (-128));
}

pub fn bid128_to_int8_xint(mut x: BID_UINT128) -> (i8, u32) {
    return bid128_to_small_int(bid128_to_int32_xint, x, (-128), (-128));
}

pub fn bid128_to_int8_floor(mut x: BID_UINT128) -> (i8, u32) {
    return bid128_to_small_int(bid128_to_int32_floor, x, (-128), (-128));
}

pub fn bid128_to_int8_xfloor(mut x: BID_UINT128) -> (i8, u32) {
    return bid128_to_small_int(bid128_to_int32_xfloor, x, (-128), (-128));
}

pub fn bid128_to_int8_ceil(mut x: BID_UINT128) -> (i8, u32) {
    return bid128_to_small_int(bid128_to_int32_ceil, x, (-128), (-128));
}

pub fn bid128_to_int8_xceil(mut x: BID_UINT128) -> (i8, u32) {
    return bid128_to_small_int(bid128_to_int32_xceil, x, (-128), (-128));
}

pub fn bid128_to_int16_rnint(mut x: BID_UINT128) -> (i16, u32) {
    return bid128_to_small_int16(bid128_to_int32_rnint, x, (-32768), (-32768));
}

pub fn bid128_to_int16_xrnint(mut x: BID_UINT128) -> (i16, u32) {
    return bid128_to_small_int16(bid128_to_int32_xrnint, x, (-32768), (-32768));
}

pub fn bid128_to_int16_rninta(mut x: BID_UINT128) -> (i16, u32) {
    return bid128_to_small_int16(bid128_to_int32_rninta, x, (-32768), (-32768));
}

pub fn bid128_to_int16_xrninta(mut x: BID_UINT128) -> (i16, u32) {
    return bid128_to_small_int16(bid128_to_int32_xrninta, x, (-32768), (-32768));
}

pub fn bid128_to_int16_int(mut x: BID_UINT128) -> (i16, u32) {
    return bid128_to_small_int16(bid128_to_int32_int, x, (-32768), (-32768));
}

pub fn bid128_to_int16_xint(mut x: BID_UINT128) -> (i16, u32) {
    return bid128_to_small_int16(bid128_to_int32_xint, x, (-32768), (-32768));
}

pub fn bid128_to_int16_floor(mut x: BID_UINT128) -> (i16, u32) {
    return bid128_to_small_int16(bid128_to_int32_floor, x, (-32768), (-32768));
}

pub fn bid128_to_int16_xfloor(mut x: BID_UINT128) -> (i16, u32) {
    return bid128_to_small_int16(bid128_to_int32_xfloor, x, (-32768), (-32768));
}

pub fn bid128_to_int16_ceil(mut x: BID_UINT128) -> (i16, u32) {
    return bid128_to_small_int16(bid128_to_int32_ceil, x, (-32768), (-32768));
}

pub fn bid128_to_int16_xceil(mut x: BID_UINT128) -> (i16, u32) {
    return bid128_to_small_int16(bid128_to_int32_xceil, x, (-32768), (-32768));
}

pub fn bid128_to_uint8_rnint(mut x: BID_UINT128) -> (u8, u32) {
    return bid128_to_small_uint(bid128_to_uint32_rnint, x, 0xffffff00, 0x80);
}

pub fn bid128_to_uint8_xrnint(mut x: BID_UINT128) -> (u8, u32) {
    return bid128_to_small_uint(bid128_to_uint32_xrnint, x, 0xffffff00, 0x80);
}

pub fn bid128_to_uint8_rninta(mut x: BID_UINT128) -> (u8, u32) {
    return bid128_to_small_uint(bid128_to_uint32_rninta, x, 0xffffff00, 0x80);
}

pub fn bid128_to_uint8_xrninta(mut x: BID_UINT128) -> (u8, u32) {
    return bid128_to_small_uint(bid128_to_uint32_xrninta, x, 0xffffff00, 0x80);
}

pub fn bid128_to_uint8_int(mut x: BID_UINT128) -> (u8, u32) {
    return bid128_to_small_uint(bid128_to_uint32_int, x, 0xffffff00, 0x80);
}

pub fn bid128_to_uint8_xint(mut x: BID_UINT128) -> (u8, u32) {
    return bid128_to_small_uint(bid128_to_uint32_xint, x, 0xffffff00, 0x80);
}

pub fn bid128_to_uint8_floor(mut x: BID_UINT128) -> (u8, u32) {
    return bid128_to_small_uint(bid128_to_uint32_floor, x, 0xffffff00, 0x80);
}

pub fn bid128_to_uint8_xfloor(mut x: BID_UINT128) -> (u8, u32) {
    return bid128_to_small_uint(bid128_to_uint32_xfloor, x, 0xffffff00, 0x80);
}

pub fn bid128_to_uint8_ceil(mut x: BID_UINT128) -> (u8, u32) {
    return bid128_to_small_uint(bid128_to_uint32_ceil, x, 0xffffff00, 0x80);
}

pub fn bid128_to_uint8_xceil(mut x: BID_UINT128) -> (u8, u32) {
    return bid128_to_small_uint(bid128_to_uint32_xceil, x, 0xffffff00, 0x80);
}

pub fn bid128_to_uint16_rnint(mut x: BID_UINT128) -> (u16, u32) {
    return bid128_to_small_uint16(bid128_to_uint32_rnint, x, 0xffff0000, 0x8000);
}

pub fn bid128_to_uint16_xrnint(mut x: BID_UINT128) -> (u16, u32) {
    return bid128_to_small_uint16(bid128_to_uint32_xrnint, x, 0xffff0000, 0x8000);
}

pub fn bid128_to_uint16_rninta(mut x: BID_UINT128) -> (u16, u32) {
    return bid128_to_small_uint16(bid128_to_uint32_rninta, x, 0xffff0000, 0x8000);
}

pub fn bid128_to_uint16_xrninta(mut x: BID_UINT128) -> (u16, u32) {
    return bid128_to_small_uint16(bid128_to_uint32_xrninta, x, 0xffff0000, 0x8000);
}

pub fn bid128_to_uint16_int(mut x: BID_UINT128) -> (u16, u32) {
    return bid128_to_small_uint16(bid128_to_uint32_int, x, 0xffff0000, 0x8000);
}

pub fn bid128_to_uint16_xint(mut x: BID_UINT128) -> (u16, u32) {
    return bid128_to_small_uint16(bid128_to_uint32_xint, x, 0xffff0000, 0x8000);
}

pub fn bid128_to_uint16_floor(mut x: BID_UINT128) -> (u16, u32) {
    return bid128_to_small_uint16(bid128_to_uint32_floor, x, 0xffff0000, 0x8000);
}

pub fn bid128_to_uint16_xfloor(mut x: BID_UINT128) -> (u16, u32) {
    return bid128_to_small_uint16(bid128_to_uint32_xfloor, x, 0xffff0000, 0x8000);
}

pub fn bid128_to_uint16_ceil(mut x: BID_UINT128) -> (u16, u32) {
    return bid128_to_small_uint16(bid128_to_uint32_ceil, x, 0xffff0000, 0x8000);
}

pub fn bid128_to_uint16_xceil(mut x: BID_UINT128) -> (u16, u32) {
    return bid128_to_small_uint16(bid128_to_uint32_xceil, x, 0xffff0000, 0x8000);
}

pub fn bid128_llrint(mut x: BID_UINT128, mut rnd_mode: i64) -> (i64, u32) {
    match rnd_mode {
        0 => {
            return bid128_to_int64_xrnint(x);
        }
        4 => {
            return bid128_to_int64_xrninta(x);
        }
        1 => {
            return bid128_to_int64_xfloor(x);
        }
        2 => {
            return bid128_to_int64_xceil(x);
        }
        _ => {
            return bid128_to_int64_xint(x);
        }
    }
}

pub fn bid128_lrint(mut x: BID_UINT128, mut rnd_mode: i64) -> (i64, u32) {
    return bid128_llrint(x, rnd_mode);
}

pub fn bid128_llround(mut x: BID_UINT128) -> (i64, u32) {
    return bid128_to_int64_rninta(x);
}

pub fn bid128_lround(mut x: BID_UINT128) -> (i64, u32) {
    return bid128_llround(x);
}
