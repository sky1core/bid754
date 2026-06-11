// Auto-generated from bid128_round.go by go2rs. Do not edit.

use super::prelude::*;

pub(crate) fn bid_round128_19_38(mut q: i64, mut x: i64, mut C: BID_UINT128) -> (BID_UINT128, i64, i64, i64, i64, i64) {
    let mut Cstar: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut incr_exp: i64 = 0;
    let mut is_midpoint_lt_even: i64 = 0;
    let mut is_midpoint_gt_even: i64 = 0;
    let mut is_inexact_lt_midpoint: i64 = 0;
    let mut is_inexact_gt_midpoint: i64 = 0;
    let mut P256: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut fstar: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut tmp64: u64 = 0;
    let mut shift: i64 = 0;
    let mut ind: i64 = 0;
    ind = (x.wrapping_sub(1));
    if (ind <= 18) {
        tmp64 = C.w[0];
        C.w[0] = (C.w[0].wrapping_add(bid_midpoint64[ind as usize]));
        if (C.w[0] < tmp64) {
            C.w[1] = C.w[1].wrapping_add(1);
        }
    } else {
        tmp64 = C.w[0];
        C.w[0] = (C.w[0].wrapping_add(bid_midpoint128[(ind.wrapping_sub(19)) as usize].w[0]));
        if (C.w[0] < tmp64) {
            C.w[1] = C.w[1].wrapping_add(1);
        }
        C.w[1] = (C.w[1].wrapping_add(bid_midpoint128[(ind.wrapping_sub(19)) as usize].w[1]));
    }
    P256 = __mul_128x128_to_256(C, bid_Kx128[ind as usize]);
    shift = (bid_Ex128m128[ind as usize] as i64);
    if (ind <= 18) {
        Cstar.w[0] = (((go_checked_shr_u64(P256.w[2], go_shift_count_u64((shift as u64) as u64)))) | ((go_checked_shl_u64(P256.w[3], go_shift_count_u64(((((64 as i64).wrapping_sub(shift)) as u64)) as u64)))));
        Cstar.w[1] = ((go_checked_shr_u64(P256.w[3], go_shift_count_u64((shift as u64) as u64))));
        fstar.w[0] = P256.w[0];
        fstar.w[1] = P256.w[1];
        fstar.w[2] = (P256.w[2] & bid_mask128[ind as usize]);
        fstar.w[3] = 0x0;
    } else {
        Cstar.w[0] = (go_checked_shr_u64(P256.w[3], go_shift_count_u64((shift as u64) as u64)));
        Cstar.w[1] = 0x0;
        fstar.w[0] = P256.w[0];
        fstar.w[1] = P256.w[1];
        fstar.w[2] = P256.w[2];
        fstar.w[3] = (P256.w[3] & bid_mask128[ind as usize]);
    }
    if (ind <= 18) {
        if ((fstar.w[2] > bid_half128[ind as usize]) || (((fstar.w[2] == bid_half128[ind as usize]) && (((fstar.w[1] != 0) || (fstar.w[0] != 0)))))) {
            tmp64 = (fstar.w[2].wrapping_sub(bid_half128[ind as usize]));
            if (((tmp64 != 0) || (fstar.w[1] > bid_ten2mxtrunc128[ind as usize].w[1])) || (((fstar.w[1] == bid_ten2mxtrunc128[ind as usize].w[1]) && (fstar.w[0] > bid_ten2mxtrunc128[ind as usize].w[0])))) {
                is_inexact_lt_midpoint = 1;
            }
        } else {
            is_inexact_gt_midpoint = 1;
        }
    } else {
        if ((fstar.w[3] > bid_half128[ind as usize]) || (((fstar.w[3] == bid_half128[ind as usize]) && ((((fstar.w[2] != 0) || (fstar.w[1] != 0)) || (fstar.w[0] != 0)))))) {
            tmp64 = (fstar.w[3].wrapping_sub(bid_half128[ind as usize]));
            if ((((tmp64 != 0) || (fstar.w[2] != 0)) || (fstar.w[1] > bid_ten2mxtrunc128[ind as usize].w[1])) || (((fstar.w[1] == bid_ten2mxtrunc128[ind as usize].w[1]) && (fstar.w[0] > bid_ten2mxtrunc128[ind as usize].w[0])))) {
                is_inexact_lt_midpoint = 1;
            }
        } else {
            is_inexact_gt_midpoint = 1;
        }
    }
    if (((fstar.w[3] == 0) && (fstar.w[2] == 0)) && (((fstar.w[1] < bid_ten2mxtrunc128[ind as usize].w[1]) || (((fstar.w[1] == bid_ten2mxtrunc128[ind as usize].w[1]) && (fstar.w[0] <= bid_ten2mxtrunc128[ind as usize].w[0])))))) {
        if ((Cstar.w[0] & 0x01) != 0) {
            Cstar.w[0] = Cstar.w[0].wrapping_sub(1);
            if (Cstar.w[0] == 0xffffffffffffffff) {
                Cstar.w[1] = Cstar.w[1].wrapping_sub(1);
            }
            is_midpoint_gt_even = 1;
            is_inexact_lt_midpoint = 0;
            is_inexact_gt_midpoint = 0;
        } else {
            is_midpoint_lt_even = 1;
            is_inexact_lt_midpoint = 0;
            is_inexact_gt_midpoint = 0;
        }
    }
    ind = (q.wrapping_sub(x));
    if (ind <= 19) {
        if ((Cstar.w[1] == 0x0) && (Cstar.w[0] == bid_ten2k64[ind as usize])) {
            Cstar.w[0] = bid_ten2k64[(ind.wrapping_sub(1)) as usize];
            incr_exp = 1;
        } else {
            incr_exp = 0;
        }
    } else if (ind == 20) {
        if ((Cstar.w[1] == bid_ten2k128[0].w[1]) && (Cstar.w[0] == bid_ten2k128[0].w[0])) {
            Cstar.w[0] = bid_ten2k64[19];
            Cstar.w[1] = 0x0;
            incr_exp = 1;
        } else {
            incr_exp = 0;
        }
    } else {
        if ((Cstar.w[1] == bid_ten2k128[(ind.wrapping_sub(20)) as usize].w[1]) && (Cstar.w[0] == bid_ten2k128[(ind.wrapping_sub(20)) as usize].w[0])) {
            Cstar.w[0] = bid_ten2k128[(ind.wrapping_sub(21)) as usize].w[0];
            Cstar.w[1] = bid_ten2k128[(ind.wrapping_sub(21)) as usize].w[1];
            incr_exp = 1;
        } else {
            incr_exp = 0;
        }
    }
    return (Cstar, incr_exp, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint);
}

pub(crate) fn bid_round192_39_57(mut q: i64, mut x: i64, mut C: BID_UINT192) -> (BID_UINT192, i64, i64, i64, i64, i64) {
    let mut Cstar: BID_UINT192 = BID_UINT192 { w: [0, 0, 0] };
    let mut incr_exp: i64 = 0;
    let mut is_midpoint_lt_even: i64 = 0;
    let mut is_midpoint_gt_even: i64 = 0;
    let mut is_inexact_lt_midpoint: i64 = 0;
    let mut is_inexact_gt_midpoint: i64 = 0;
    let mut P384: BID_UINT384 = BID_UINT384 { w: [0, 0, 0, 0, 0, 0] };
    let mut fstar: BID_UINT384 = BID_UINT384 { w: [0, 0, 0, 0, 0, 0] };
    let mut tmp64: u64 = 0;
    let mut shift: i64 = 0;
    let mut ind: i64 = 0;
    ind = (x.wrapping_sub(1));
    if (ind <= 18) {
        tmp64 = C.w[0];
        C.w[0] = (C.w[0].wrapping_add(bid_midpoint64[ind as usize]));
        if (C.w[0] < tmp64) {
            C.w[1] = C.w[1].wrapping_add(1);
            if (C.w[1] == 0x0) {
                C.w[2] = C.w[2].wrapping_add(1);
            }
        }
    } else if (ind <= 37) {
        tmp64 = C.w[0];
        C.w[0] = (C.w[0].wrapping_add(bid_midpoint128[(ind.wrapping_sub(19)) as usize].w[0]));
        if (C.w[0] < tmp64) {
            C.w[1] = C.w[1].wrapping_add(1);
            if (C.w[1] == 0x0) {
                C.w[2] = C.w[2].wrapping_add(1);
            }
        }
        tmp64 = C.w[1];
        C.w[1] = (C.w[1].wrapping_add(bid_midpoint128[(ind.wrapping_sub(19)) as usize].w[1]));
        if (C.w[1] < tmp64) {
            C.w[2] = C.w[2].wrapping_add(1);
        }
    } else {
        tmp64 = C.w[0];
        C.w[0] = (C.w[0].wrapping_add(bid_midpoint192[(ind.wrapping_sub(38)) as usize].w[0]));
        if (C.w[0] < tmp64) {
            C.w[1] = C.w[1].wrapping_add(1);
            if (C.w[1] == 0x0) {
                C.w[2] = C.w[2].wrapping_add(1);
            }
        }
        tmp64 = C.w[1];
        C.w[1] = (C.w[1].wrapping_add(bid_midpoint192[(ind.wrapping_sub(38)) as usize].w[1]));
        if (C.w[1] < tmp64) {
            C.w[2] = C.w[2].wrapping_add(1);
        }
        C.w[2] = (C.w[2].wrapping_add(bid_midpoint192[(ind.wrapping_sub(38)) as usize].w[2]));
    }
    P384 = __mul_192x192_to_384(C, bid_Kx192[ind as usize]);
    shift = (bid_Ex192m192[ind as usize] as i64);
    if (ind <= 18) {
        Cstar.w[2] = ((go_checked_shr_u64(P384.w[5], go_shift_count_u64((shift as u64) as u64))));
        Cstar.w[1] = (((go_checked_shl_u64(P384.w[5], go_shift_count_u64(((((64 as i64).wrapping_sub(shift)) as u64)) as u64)))) | ((go_checked_shr_u64(P384.w[4], go_shift_count_u64((shift as u64) as u64)))));
        Cstar.w[0] = (((go_checked_shl_u64(P384.w[4], go_shift_count_u64(((((64 as i64).wrapping_sub(shift)) as u64)) as u64)))) | ((go_checked_shr_u64(P384.w[3], go_shift_count_u64((shift as u64) as u64)))));
        fstar.w[5] = 0x0;
        fstar.w[4] = 0x0;
        fstar.w[3] = (P384.w[3] & bid_mask192[ind as usize]);
        fstar.w[2] = P384.w[2];
        fstar.w[1] = P384.w[1];
        fstar.w[0] = P384.w[0];
    } else if (ind <= 37) {
        Cstar.w[2] = 0x0;
        Cstar.w[1] = (go_checked_shr_u64(P384.w[5], go_shift_count_u64((shift as u64) as u64)));
        Cstar.w[0] = (((go_checked_shl_u64(P384.w[5], go_shift_count_u64(((((64 as i64).wrapping_sub(shift)) as u64)) as u64)))) | ((go_checked_shr_u64(P384.w[4], go_shift_count_u64((shift as u64) as u64)))));
        fstar.w[5] = 0x0;
        fstar.w[4] = (P384.w[4] & bid_mask192[ind as usize]);
        fstar.w[3] = P384.w[3];
        fstar.w[2] = P384.w[2];
        fstar.w[1] = P384.w[1];
        fstar.w[0] = P384.w[0];
    } else {
        Cstar.w[2] = 0x0;
        Cstar.w[1] = 0x0;
        Cstar.w[0] = (go_checked_shr_u64(P384.w[5], go_shift_count_u64((shift as u64) as u64)));
        fstar.w[5] = (P384.w[5] & bid_mask192[ind as usize]);
        fstar.w[4] = P384.w[4];
        fstar.w[3] = P384.w[3];
        fstar.w[2] = P384.w[2];
        fstar.w[1] = P384.w[1];
        fstar.w[0] = P384.w[0];
    }
    if (ind <= 18) {
        if ((fstar.w[3] > bid_half192[ind as usize]) || (((fstar.w[3] == bid_half192[ind as usize]) && ((((fstar.w[2] != 0) || (fstar.w[1] != 0)) || (fstar.w[0] != 0)))))) {
            tmp64 = (fstar.w[3].wrapping_sub(bid_half192[ind as usize]));
            if ((((tmp64 != 0) || (fstar.w[2] > bid_ten2mxtrunc192[ind as usize].w[2])) || (((fstar.w[2] == bid_ten2mxtrunc192[ind as usize].w[2]) && (fstar.w[1] > bid_ten2mxtrunc192[ind as usize].w[1])))) || ((((fstar.w[2] == bid_ten2mxtrunc192[ind as usize].w[2]) && (fstar.w[1] == bid_ten2mxtrunc192[ind as usize].w[1])) && (fstar.w[0] > bid_ten2mxtrunc192[ind as usize].w[0])))) {
                is_inexact_lt_midpoint = 1;
            }
        } else {
            is_inexact_gt_midpoint = 1;
        }
    } else if (ind <= 37) {
        if ((fstar.w[4] > bid_half192[ind as usize]) || (((fstar.w[4] == bid_half192[ind as usize]) && (((((fstar.w[3] != 0) || (fstar.w[2] != 0)) || (fstar.w[1] != 0)) || (fstar.w[0] != 0)))))) {
            tmp64 = (fstar.w[4].wrapping_sub(bid_half192[ind as usize]));
            if (((((tmp64 != 0) || (fstar.w[3] != 0)) || (fstar.w[2] > bid_ten2mxtrunc192[ind as usize].w[2])) || (((fstar.w[2] == bid_ten2mxtrunc192[ind as usize].w[2]) && (fstar.w[1] > bid_ten2mxtrunc192[ind as usize].w[1])))) || ((((fstar.w[2] == bid_ten2mxtrunc192[ind as usize].w[2]) && (fstar.w[1] == bid_ten2mxtrunc192[ind as usize].w[1])) && (fstar.w[0] > bid_ten2mxtrunc192[ind as usize].w[0])))) {
                is_inexact_lt_midpoint = 1;
            }
        } else {
            is_inexact_gt_midpoint = 1;
        }
    } else {
        if ((fstar.w[5] > bid_half192[ind as usize]) || (((fstar.w[5] == bid_half192[ind as usize]) && ((((((fstar.w[4] != 0) || (fstar.w[3] != 0)) || (fstar.w[2] != 0)) || (fstar.w[1] != 0)) || (fstar.w[0] != 0)))))) {
            tmp64 = (fstar.w[5].wrapping_sub(bid_half192[ind as usize]));
            if ((((((tmp64 != 0) || (fstar.w[4] != 0)) || (fstar.w[3] != 0)) || (fstar.w[2] > bid_ten2mxtrunc192[ind as usize].w[2])) || (((fstar.w[2] == bid_ten2mxtrunc192[ind as usize].w[2]) && (fstar.w[1] > bid_ten2mxtrunc192[ind as usize].w[1])))) || ((((fstar.w[2] == bid_ten2mxtrunc192[ind as usize].w[2]) && (fstar.w[1] == bid_ten2mxtrunc192[ind as usize].w[1])) && (fstar.w[0] > bid_ten2mxtrunc192[ind as usize].w[0])))) {
                is_inexact_lt_midpoint = 1;
            }
        } else {
            is_inexact_gt_midpoint = 1;
        }
    }
    if ((((fstar.w[5] == 0) && (fstar.w[4] == 0)) && (fstar.w[3] == 0)) && ((((fstar.w[2] < bid_ten2mxtrunc192[ind as usize].w[2]) || (((fstar.w[2] == bid_ten2mxtrunc192[ind as usize].w[2]) && (fstar.w[1] < bid_ten2mxtrunc192[ind as usize].w[1])))) || ((((fstar.w[2] == bid_ten2mxtrunc192[ind as usize].w[2]) && (fstar.w[1] == bid_ten2mxtrunc192[ind as usize].w[1])) && (fstar.w[0] <= bid_ten2mxtrunc192[ind as usize].w[0])))))) {
        if ((Cstar.w[0] & 0x01) != 0) {
            Cstar.w[0] = Cstar.w[0].wrapping_sub(1);
            if (Cstar.w[0] == 0xffffffffffffffff) {
                Cstar.w[1] = Cstar.w[1].wrapping_sub(1);
                if (Cstar.w[1] == 0xffffffffffffffff) {
                    Cstar.w[2] = Cstar.w[2].wrapping_sub(1);
                }
            }
            is_midpoint_gt_even = 1;
            is_inexact_lt_midpoint = 0;
            is_inexact_gt_midpoint = 0;
        } else {
            is_midpoint_lt_even = 1;
            is_inexact_lt_midpoint = 0;
            is_inexact_gt_midpoint = 0;
        }
    }
    ind = (q.wrapping_sub(x));
    if (ind <= 19) {
        if (((Cstar.w[2] == 0x0) && (Cstar.w[1] == 0x0)) && (Cstar.w[0] == bid_ten2k64[ind as usize])) {
            Cstar.w[0] = bid_ten2k64[(ind.wrapping_sub(1)) as usize];
            incr_exp = 1;
        } else {
            incr_exp = 0;
        }
    } else if (ind == 20) {
        if (((Cstar.w[2] == 0x0) && (Cstar.w[1] == bid_ten2k128[0].w[1])) && (Cstar.w[0] == bid_ten2k128[0].w[0])) {
            Cstar.w[0] = bid_ten2k64[19];
            Cstar.w[1] = 0x0;
            incr_exp = 1;
        } else {
            incr_exp = 0;
        }
    } else if (ind <= 38) {
        if (((Cstar.w[2] == 0x0) && (Cstar.w[1] == bid_ten2k128[(ind.wrapping_sub(20)) as usize].w[1])) && (Cstar.w[0] == bid_ten2k128[(ind.wrapping_sub(20)) as usize].w[0])) {
            Cstar.w[0] = bid_ten2k128[(ind.wrapping_sub(21)) as usize].w[0];
            Cstar.w[1] = bid_ten2k128[(ind.wrapping_sub(21)) as usize].w[1];
            incr_exp = 1;
        } else {
            incr_exp = 0;
        }
    } else if (ind == 39) {
        if (((Cstar.w[2] == bid_ten2k256[0].w[2]) && (Cstar.w[1] == bid_ten2k256[0].w[1])) && (Cstar.w[0] == bid_ten2k256[0].w[0])) {
            Cstar.w[0] = bid_ten2k128[18].w[0];
            Cstar.w[1] = bid_ten2k128[18].w[1];
            Cstar.w[2] = 0x0;
            incr_exp = 1;
        } else {
            incr_exp = 0;
        }
    } else {
        if (((Cstar.w[2] == bid_ten2k256[(ind.wrapping_sub(39)) as usize].w[2]) && (Cstar.w[1] == bid_ten2k256[(ind.wrapping_sub(39)) as usize].w[1])) && (Cstar.w[0] == bid_ten2k256[(ind.wrapping_sub(39)) as usize].w[0])) {
            Cstar.w[0] = bid_ten2k256[(ind.wrapping_sub(40)) as usize].w[0];
            Cstar.w[1] = bid_ten2k256[(ind.wrapping_sub(40)) as usize].w[1];
            Cstar.w[2] = bid_ten2k256[(ind.wrapping_sub(40)) as usize].w[2];
            incr_exp = 1;
        } else {
            incr_exp = 0;
        }
    }
    return (Cstar, incr_exp, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint);
}

pub(crate) fn bid_round256_58_76(mut q: i64, mut x: i64, mut C: BID_UINT256) -> (BID_UINT256, i64, i64, i64, i64, i64) {
    let mut Cstar: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut incr_exp: i64 = 0;
    let mut is_midpoint_lt_even: i64 = 0;
    let mut is_midpoint_gt_even: i64 = 0;
    let mut is_inexact_lt_midpoint: i64 = 0;
    let mut is_inexact_gt_midpoint: i64 = 0;
    let mut P512: BID_UINT512 = BID_UINT512 { w: [0, 0, 0, 0, 0, 0, 0, 0] };
    let mut fstar: BID_UINT512 = BID_UINT512 { w: [0, 0, 0, 0, 0, 0, 0, 0] };
    let mut tmp64: u64 = 0;
    let mut shift: i64 = 0;
    let mut ind: i64 = 0;
    ind = (x.wrapping_sub(1));
    if (ind <= 18) {
        tmp64 = C.w[0];
        C.w[0] = (C.w[0].wrapping_add(bid_midpoint64[ind as usize]));
        if (C.w[0] < tmp64) {
            C.w[1] = C.w[1].wrapping_add(1);
            if (C.w[1] == 0x0) {
                C.w[2] = C.w[2].wrapping_add(1);
                if (C.w[2] == 0x0) {
                    C.w[3] = C.w[3].wrapping_add(1);
                }
            }
        }
    } else if (ind <= 37) {
        tmp64 = C.w[0];
        C.w[0] = (C.w[0].wrapping_add(bid_midpoint128[(ind.wrapping_sub(19)) as usize].w[0]));
        if (C.w[0] < tmp64) {
            C.w[1] = C.w[1].wrapping_add(1);
            if (C.w[1] == 0x0) {
                C.w[2] = C.w[2].wrapping_add(1);
                if (C.w[2] == 0x0) {
                    C.w[3] = C.w[3].wrapping_add(1);
                }
            }
        }
        tmp64 = C.w[1];
        C.w[1] = (C.w[1].wrapping_add(bid_midpoint128[(ind.wrapping_sub(19)) as usize].w[1]));
        if (C.w[1] < tmp64) {
            C.w[2] = C.w[2].wrapping_add(1);
            if (C.w[2] == 0x0) {
                C.w[3] = C.w[3].wrapping_add(1);
            }
        }
    } else if (ind <= 57) {
        tmp64 = C.w[0];
        C.w[0] = (C.w[0].wrapping_add(bid_midpoint192[(ind.wrapping_sub(38)) as usize].w[0]));
        if (C.w[0] < tmp64) {
            C.w[1] = C.w[1].wrapping_add(1);
            if (C.w[1] == 0x0) {
                C.w[2] = C.w[2].wrapping_add(1);
                if (C.w[2] == 0x0) {
                    C.w[3] = C.w[3].wrapping_add(1);
                }
            }
        }
        tmp64 = C.w[1];
        C.w[1] = (C.w[1].wrapping_add(bid_midpoint192[(ind.wrapping_sub(38)) as usize].w[1]));
        if (C.w[1] < tmp64) {
            C.w[2] = C.w[2].wrapping_add(1);
            if (C.w[2] == 0x0) {
                C.w[3] = C.w[3].wrapping_add(1);
            }
        }
        tmp64 = C.w[2];
        C.w[2] = (C.w[2].wrapping_add(bid_midpoint192[(ind.wrapping_sub(38)) as usize].w[2]));
        if (C.w[2] < tmp64) {
            C.w[3] = C.w[3].wrapping_add(1);
        }
    } else {
        tmp64 = C.w[0];
        C.w[0] = (C.w[0].wrapping_add(bid_midpoint256[(ind.wrapping_sub(58)) as usize].w[0]));
        if (C.w[0] < tmp64) {
            C.w[1] = C.w[1].wrapping_add(1);
            if (C.w[1] == 0x0) {
                C.w[2] = C.w[2].wrapping_add(1);
                if (C.w[2] == 0x0) {
                    C.w[3] = C.w[3].wrapping_add(1);
                }
            }
        }
        tmp64 = C.w[1];
        C.w[1] = (C.w[1].wrapping_add(bid_midpoint256[(ind.wrapping_sub(58)) as usize].w[1]));
        if (C.w[1] < tmp64) {
            C.w[2] = C.w[2].wrapping_add(1);
            if (C.w[2] == 0x0) {
                C.w[3] = C.w[3].wrapping_add(1);
            }
        }
        tmp64 = C.w[2];
        C.w[2] = (C.w[2].wrapping_add(bid_midpoint256[(ind.wrapping_sub(58)) as usize].w[2]));
        if (C.w[2] < tmp64) {
            C.w[3] = C.w[3].wrapping_add(1);
        }
        C.w[3] = (C.w[3].wrapping_add(bid_midpoint256[(ind.wrapping_sub(58)) as usize].w[3]));
    }
    P512 = __mul_256x256_to_512(C, bid_Kx256[ind as usize]);
    shift = (bid_Ex256m256[ind as usize] as i64);
    if (ind <= 18) {
        Cstar.w[3] = ((go_checked_shr_u64(P512.w[7], go_shift_count_u64((shift as u64) as u64))));
        Cstar.w[2] = (((go_checked_shl_u64(P512.w[7], go_shift_count_u64(((((64 as i64).wrapping_sub(shift)) as u64)) as u64)))) | ((go_checked_shr_u64(P512.w[6], go_shift_count_u64((shift as u64) as u64)))));
        Cstar.w[1] = (((go_checked_shl_u64(P512.w[6], go_shift_count_u64(((((64 as i64).wrapping_sub(shift)) as u64)) as u64)))) | ((go_checked_shr_u64(P512.w[5], go_shift_count_u64((shift as u64) as u64)))));
        Cstar.w[0] = (((go_checked_shl_u64(P512.w[5], go_shift_count_u64(((((64 as i64).wrapping_sub(shift)) as u64)) as u64)))) | ((go_checked_shr_u64(P512.w[4], go_shift_count_u64((shift as u64) as u64)))));
        fstar.w[7] = 0x0;
        fstar.w[6] = 0x0;
        fstar.w[5] = 0x0;
        fstar.w[4] = (P512.w[4] & bid_mask256[ind as usize]);
        fstar.w[3] = P512.w[3];
        fstar.w[2] = P512.w[2];
        fstar.w[1] = P512.w[1];
        fstar.w[0] = P512.w[0];
    } else if (ind <= 37) {
        Cstar.w[3] = 0x0;
        Cstar.w[2] = (go_checked_shr_u64(P512.w[7], go_shift_count_u64((shift as u64) as u64)));
        Cstar.w[1] = (((go_checked_shl_u64(P512.w[7], go_shift_count_u64(((((64 as i64).wrapping_sub(shift)) as u64)) as u64)))) | ((go_checked_shr_u64(P512.w[6], go_shift_count_u64((shift as u64) as u64)))));
        Cstar.w[0] = (((go_checked_shl_u64(P512.w[6], go_shift_count_u64(((((64 as i64).wrapping_sub(shift)) as u64)) as u64)))) | ((go_checked_shr_u64(P512.w[5], go_shift_count_u64((shift as u64) as u64)))));
        fstar.w[7] = 0x0;
        fstar.w[6] = 0x0;
        fstar.w[5] = (P512.w[5] & bid_mask256[ind as usize]);
        fstar.w[4] = P512.w[4];
        fstar.w[3] = P512.w[3];
        fstar.w[2] = P512.w[2];
        fstar.w[1] = P512.w[1];
        fstar.w[0] = P512.w[0];
    } else if (ind <= 56) {
        Cstar.w[3] = 0x0;
        Cstar.w[2] = 0x0;
        Cstar.w[1] = (go_checked_shr_u64(P512.w[7], go_shift_count_u64((shift as u64) as u64)));
        Cstar.w[0] = (((go_checked_shl_u64(P512.w[7], go_shift_count_u64(((((64 as i64).wrapping_sub(shift)) as u64)) as u64)))) | ((go_checked_shr_u64(P512.w[6], go_shift_count_u64((shift as u64) as u64)))));
        fstar.w[7] = 0x0;
        fstar.w[6] = (P512.w[6] & bid_mask256[ind as usize]);
        fstar.w[5] = P512.w[5];
        fstar.w[4] = P512.w[4];
        fstar.w[3] = P512.w[3];
        fstar.w[2] = P512.w[2];
        fstar.w[1] = P512.w[1];
        fstar.w[0] = P512.w[0];
    } else if (ind == 57) {
        Cstar.w[3] = 0x0;
        Cstar.w[2] = 0x0;
        Cstar.w[1] = 0x0;
        Cstar.w[0] = P512.w[7];
        fstar.w[7] = 0x0;
        fstar.w[6] = P512.w[6];
        fstar.w[5] = P512.w[5];
        fstar.w[4] = P512.w[4];
        fstar.w[3] = P512.w[3];
        fstar.w[2] = P512.w[2];
        fstar.w[1] = P512.w[1];
        fstar.w[0] = P512.w[0];
    } else {
        Cstar.w[3] = 0x0;
        Cstar.w[2] = 0x0;
        Cstar.w[1] = 0x0;
        Cstar.w[0] = (go_checked_shr_u64(P512.w[7], go_shift_count_u64((shift as u64) as u64)));
        fstar.w[7] = (P512.w[7] & bid_mask256[ind as usize]);
        fstar.w[6] = P512.w[6];
        fstar.w[5] = P512.w[5];
        fstar.w[4] = P512.w[4];
        fstar.w[3] = P512.w[3];
        fstar.w[2] = P512.w[2];
        fstar.w[1] = P512.w[1];
        fstar.w[0] = P512.w[0];
    }
    if (ind <= 18) {
        if ((fstar.w[4] > bid_half256[ind as usize]) || (((fstar.w[4] == bid_half256[ind as usize]) && (((((fstar.w[3] != 0) || (fstar.w[2] != 0)) || (fstar.w[1] != 0)) || (fstar.w[0] != 0)))))) {
            tmp64 = (fstar.w[4].wrapping_sub(bid_half256[ind as usize]));
            if (((((tmp64 != 0) || (fstar.w[3] > bid_ten2mxtrunc256[ind as usize].w[2])) || (((fstar.w[3] == bid_ten2mxtrunc256[ind as usize].w[3]) && (fstar.w[2] > bid_ten2mxtrunc256[ind as usize].w[2])))) || ((((fstar.w[3] == bid_ten2mxtrunc256[ind as usize].w[3]) && (fstar.w[2] == bid_ten2mxtrunc256[ind as usize].w[2])) && (fstar.w[1] > bid_ten2mxtrunc256[ind as usize].w[1])))) || (((((fstar.w[3] == bid_ten2mxtrunc256[ind as usize].w[3]) && (fstar.w[2] == bid_ten2mxtrunc256[ind as usize].w[2])) && (fstar.w[1] == bid_ten2mxtrunc256[ind as usize].w[1])) && (fstar.w[0] > bid_ten2mxtrunc256[ind as usize].w[0])))) {
                is_inexact_lt_midpoint = 1;
            }
        } else {
            is_inexact_gt_midpoint = 1;
        }
    } else if (ind <= 37) {
        if ((fstar.w[5] > bid_half256[ind as usize]) || (((fstar.w[5] == bid_half256[ind as usize]) && ((((((fstar.w[4] != 0) || (fstar.w[3] != 0)) || (fstar.w[2] != 0)) || (fstar.w[1] != 0)) || (fstar.w[0] != 0)))))) {
            tmp64 = (fstar.w[5].wrapping_sub(bid_half256[ind as usize]));
            if ((((((tmp64 != 0) || (fstar.w[4] != 0)) || (fstar.w[3] > bid_ten2mxtrunc256[ind as usize].w[3])) || (((fstar.w[3] == bid_ten2mxtrunc256[ind as usize].w[3]) && (fstar.w[2] > bid_ten2mxtrunc256[ind as usize].w[2])))) || ((((fstar.w[3] == bid_ten2mxtrunc256[ind as usize].w[3]) && (fstar.w[2] == bid_ten2mxtrunc256[ind as usize].w[2])) && (fstar.w[1] > bid_ten2mxtrunc256[ind as usize].w[1])))) || (((((fstar.w[3] == bid_ten2mxtrunc256[ind as usize].w[3]) && (fstar.w[2] == bid_ten2mxtrunc256[ind as usize].w[2])) && (fstar.w[1] == bid_ten2mxtrunc256[ind as usize].w[1])) && (fstar.w[0] > bid_ten2mxtrunc256[ind as usize].w[0])))) {
                is_inexact_lt_midpoint = 1;
            }
        } else {
            is_inexact_gt_midpoint = 1;
        }
    } else if (ind <= 57) {
        if ((fstar.w[6] > bid_half256[ind as usize]) || (((fstar.w[6] == bid_half256[ind as usize]) && (((((((fstar.w[5] != 0) || (fstar.w[4] != 0)) || (fstar.w[3] != 0)) || (fstar.w[2] != 0)) || (fstar.w[1] != 0)) || (fstar.w[0] != 0)))))) {
            tmp64 = (fstar.w[6].wrapping_sub(bid_half256[ind as usize]));
            if (((((((tmp64 != 0) || (fstar.w[5] != 0)) || (fstar.w[4] != 0)) || (fstar.w[3] > bid_ten2mxtrunc256[ind as usize].w[3])) || (((fstar.w[3] == bid_ten2mxtrunc256[ind as usize].w[3]) && (fstar.w[2] > bid_ten2mxtrunc256[ind as usize].w[2])))) || ((((fstar.w[3] == bid_ten2mxtrunc256[ind as usize].w[3]) && (fstar.w[2] == bid_ten2mxtrunc256[ind as usize].w[2])) && (fstar.w[1] > bid_ten2mxtrunc256[ind as usize].w[1])))) || (((((fstar.w[3] == bid_ten2mxtrunc256[ind as usize].w[3]) && (fstar.w[2] == bid_ten2mxtrunc256[ind as usize].w[2])) && (fstar.w[1] == bid_ten2mxtrunc256[ind as usize].w[1])) && (fstar.w[0] > bid_ten2mxtrunc256[ind as usize].w[0])))) {
                is_inexact_lt_midpoint = 1;
            }
        } else {
            is_inexact_gt_midpoint = 1;
        }
    } else {
        if ((fstar.w[7] > bid_half256[ind as usize]) || (((fstar.w[7] == bid_half256[ind as usize]) && ((((((((fstar.w[6] != 0) || (fstar.w[5] != 0)) || (fstar.w[4] != 0)) || (fstar.w[3] != 0)) || (fstar.w[2] != 0)) || (fstar.w[1] != 0)) || (fstar.w[0] != 0)))))) {
            tmp64 = (fstar.w[7].wrapping_sub(bid_half256[ind as usize]));
            if ((((((((tmp64 != 0) || (fstar.w[6] != 0)) || (fstar.w[5] != 0)) || (fstar.w[4] != 0)) || (fstar.w[3] > bid_ten2mxtrunc256[ind as usize].w[3])) || (((fstar.w[3] == bid_ten2mxtrunc256[ind as usize].w[3]) && (fstar.w[2] > bid_ten2mxtrunc256[ind as usize].w[2])))) || ((((fstar.w[3] == bid_ten2mxtrunc256[ind as usize].w[3]) && (fstar.w[2] == bid_ten2mxtrunc256[ind as usize].w[2])) && (fstar.w[1] > bid_ten2mxtrunc256[ind as usize].w[1])))) || (((((fstar.w[3] == bid_ten2mxtrunc256[ind as usize].w[3]) && (fstar.w[2] == bid_ten2mxtrunc256[ind as usize].w[2])) && (fstar.w[1] == bid_ten2mxtrunc256[ind as usize].w[1])) && (fstar.w[0] > bid_ten2mxtrunc256[ind as usize].w[0])))) {
                is_inexact_lt_midpoint = 1;
            }
        } else {
            is_inexact_gt_midpoint = 1;
        }
    }
    if (((((fstar.w[7] == 0) && (fstar.w[6] == 0)) && (fstar.w[5] == 0)) && (fstar.w[4] == 0)) && (((((fstar.w[3] < bid_ten2mxtrunc256[ind as usize].w[3]) || (((fstar.w[3] == bid_ten2mxtrunc256[ind as usize].w[3]) && (fstar.w[2] < bid_ten2mxtrunc256[ind as usize].w[2])))) || ((((fstar.w[3] == bid_ten2mxtrunc256[ind as usize].w[3]) && (fstar.w[2] == bid_ten2mxtrunc256[ind as usize].w[2])) && (fstar.w[1] < bid_ten2mxtrunc256[ind as usize].w[1])))) || (((((fstar.w[3] == bid_ten2mxtrunc256[ind as usize].w[3]) && (fstar.w[2] == bid_ten2mxtrunc256[ind as usize].w[2])) && (fstar.w[1] == bid_ten2mxtrunc256[ind as usize].w[1])) && (fstar.w[0] <= bid_ten2mxtrunc256[ind as usize].w[0])))))) {
        if ((Cstar.w[0] & 0x01) != 0) {
            Cstar.w[0] = Cstar.w[0].wrapping_sub(1);
            if (Cstar.w[0] == 0xffffffffffffffff) {
                Cstar.w[1] = Cstar.w[1].wrapping_sub(1);
                if (Cstar.w[1] == 0xffffffffffffffff) {
                    Cstar.w[2] = Cstar.w[2].wrapping_sub(1);
                    if (Cstar.w[2] == 0xffffffffffffffff) {
                        Cstar.w[3] = Cstar.w[3].wrapping_sub(1);
                    }
                }
            }
            is_midpoint_gt_even = 1;
            is_inexact_lt_midpoint = 0;
            is_inexact_gt_midpoint = 0;
        } else {
            is_midpoint_lt_even = 1;
            is_inexact_lt_midpoint = 0;
            is_inexact_gt_midpoint = 0;
        }
    }
    ind = (q.wrapping_sub(x));
    if (ind <= 19) {
        if ((((Cstar.w[3] == 0x0) && (Cstar.w[2] == 0x0)) && (Cstar.w[1] == 0x0)) && (Cstar.w[0] == bid_ten2k64[ind as usize])) {
            Cstar.w[0] = bid_ten2k64[(ind.wrapping_sub(1)) as usize];
            incr_exp = 1;
        } else {
            incr_exp = 0;
        }
    } else if (ind == 20) {
        if ((((Cstar.w[3] == 0x0) && (Cstar.w[2] == 0x0)) && (Cstar.w[1] == bid_ten2k128[0].w[1])) && (Cstar.w[0] == bid_ten2k128[0].w[0])) {
            Cstar.w[0] = bid_ten2k64[19];
            Cstar.w[1] = 0x0;
            incr_exp = 1;
        } else {
            incr_exp = 0;
        }
    } else if (ind <= 38) {
        if ((((Cstar.w[3] == 0x0) && (Cstar.w[2] == 0x0)) && (Cstar.w[1] == bid_ten2k128[(ind.wrapping_sub(20)) as usize].w[1])) && (Cstar.w[0] == bid_ten2k128[(ind.wrapping_sub(20)) as usize].w[0])) {
            Cstar.w[0] = bid_ten2k128[(ind.wrapping_sub(21)) as usize].w[0];
            Cstar.w[1] = bid_ten2k128[(ind.wrapping_sub(21)) as usize].w[1];
            incr_exp = 1;
        } else {
            incr_exp = 0;
        }
    } else if (ind == 39) {
        if ((((Cstar.w[3] == 0x0) && (Cstar.w[2] == bid_ten2k256[0].w[2])) && (Cstar.w[1] == bid_ten2k256[0].w[1])) && (Cstar.w[0] == bid_ten2k256[0].w[0])) {
            Cstar.w[0] = bid_ten2k128[18].w[0];
            Cstar.w[1] = bid_ten2k128[18].w[1];
            Cstar.w[2] = 0x0;
            incr_exp = 1;
        } else {
            incr_exp = 0;
        }
    } else if (ind <= 57) {
        if ((((Cstar.w[3] == 0x0) && (Cstar.w[2] == bid_ten2k256[(ind.wrapping_sub(39)) as usize].w[2])) && (Cstar.w[1] == bid_ten2k256[(ind.wrapping_sub(39)) as usize].w[1])) && (Cstar.w[0] == bid_ten2k256[(ind.wrapping_sub(39)) as usize].w[0])) {
            Cstar.w[0] = bid_ten2k256[(ind.wrapping_sub(40)) as usize].w[0];
            Cstar.w[1] = bid_ten2k256[(ind.wrapping_sub(40)) as usize].w[1];
            Cstar.w[2] = bid_ten2k256[(ind.wrapping_sub(40)) as usize].w[2];
            incr_exp = 1;
        } else {
            incr_exp = 0;
        }
    } else {
        if ((((Cstar.w[3] == bid_ten2k256[(ind.wrapping_sub(39)) as usize].w[3]) && (Cstar.w[2] == bid_ten2k256[(ind.wrapping_sub(39)) as usize].w[2])) && (Cstar.w[1] == bid_ten2k256[(ind.wrapping_sub(39)) as usize].w[1])) && (Cstar.w[0] == bid_ten2k256[(ind.wrapping_sub(39)) as usize].w[0])) {
            Cstar.w[0] = bid_ten2k256[(ind.wrapping_sub(40)) as usize].w[0];
            Cstar.w[1] = bid_ten2k256[(ind.wrapping_sub(40)) as usize].w[1];
            Cstar.w[2] = bid_ten2k256[(ind.wrapping_sub(40)) as usize].w[2];
            Cstar.w[3] = bid_ten2k256[(ind.wrapping_sub(40)) as usize].w[3];
            incr_exp = 1;
        } else {
            incr_exp = 0;
        }
    }
    return (Cstar, incr_exp, is_midpoint_lt_even, is_midpoint_gt_even, is_inexact_lt_midpoint, is_inexact_gt_midpoint);
}
