// Auto-generated from frexp64.go by go2rs. Do not edit.

use super::prelude::*;

pub fn bid64_frexp(mut x: u64) -> (u64, i64) {
    let mut res: u64 = 0;
    let mut sig_x: u64 = 0;
    let mut exp_x: u64 = 0;
    let mut C: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut q: i64 = 0;
    if (((x & 0x7c00000000000000) == 0x7c00000000000000) || ((x & 0x7800000000000000) == 0x7800000000000000)) {
        res = x;
        if ((x & 0x7e00000000000000) == 0x7e00000000000000) {
            res = (x & 0xfdffffffffffffff);
        }
        return (res, 0);
    } else {
        if ((x & 0x6000000000000000) == 0x6000000000000000) {
            sig_x = ((x & 0x7ffffffffffff) | 0x20000000000000);
            exp_x = (go_checked_shr_u64((x & 0x1ff8000000000000), go_shift_count_u64((51) as u64)));
            if ((sig_x > 9999999999999999) || (sig_x == 0)) {
                res = ((x & 0x8000000000000000) | ((go_checked_shl_u64(exp_x, go_shift_count_u64((53) as u64)))));
                return (res, 0);
            }
        } else {
            sig_x = (x & 0x1fffffffffffff);
            exp_x = (go_checked_shr_u64((x & 0x7fe0000000000000), go_shift_count_u64((53) as u64)));
            if (sig_x == 0x0) {
                res = x;
                return (res, 0);
            }
        }
        if (sig_x >= 0x0020000000000000) {
            q = 16;
        } else {
            C.w[0] = sig_x;
            C.w[1] = 0;
            q = __get_dec_digits64(C);
        }
        if (sig_x < 0x0020000000000000) {
            res = ((x & 0x801fffffffffffff) | ((go_checked_shl_u64((((q.wrapping_neg()).wrapping_add(398)) as u64), go_shift_count_u64((53) as u64)))));
        } else {
            res = ((x & 0xe007ffffffffffff) | ((go_checked_shl_u64((((q.wrapping_neg()).wrapping_add(398)) as u64), go_shift_count_u64((51) as u64)))));
        }
        return (res, (((exp_x as i64).wrapping_sub(398)).wrapping_add(q)));
    }
}
