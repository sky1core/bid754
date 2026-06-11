// Auto-generated from reduce64.go by go2rs. Do not edit.

use super::prelude::*;

pub fn bid64_reduce(mut x: u64) -> (u64, u32) {
    let mut xSign: u64 = 0;
    let mut xExp: u64 = 0;
    let mut c1: u64 = 0;
    if ((x & 0x7c00000000000000) == 0x7c00000000000000) {
        if ((x & 0x0003ffffffffffff) > 999999999999999) {
            x = (x & 0xfe00000000000000);
        } else {
            x = (x & 0xfe03ffffffffffff);
        }
        if ((x & 0x7e00000000000000) == 0x7e00000000000000) {
            return ((x & 0xfdffffffffffffff), 1);
        }
        return (x, 0);
    }
    if ((x & 0x7800000000000000) == 0x7800000000000000) {
        return (((x & 0x8000000000000000) | 0x7800000000000000), 0);
    }
    xSign = (x & 0x8000000000000000);
    if ((x & 0x6000000000000000) == 0x6000000000000000) {
        c1 = ((x & 0x7ffffffffffff) | 0x20000000000000);
        xExp = (go_checked_shr_u64((x & 0x1ff8000000000000), go_shift_count_u64((51) as u64)));
        if ((c1 > 9999999999999999) || (c1 == 0)) {
            return ((xSign | (((0x18e as u64) << 53))), 0);
        }
    } else {
        c1 = (x & 0x1fffffffffffff);
        xExp = (go_checked_shr_u64((x & 0x7fe0000000000000), go_shift_count_u64((53) as u64)));
        if (c1 == 0) {
            return ((xSign | (((0x18e as u64) << 53))), 0);
        }
    }
    while (((c1 != 0) && ((c1 % 10) == 0)) && (xExp < 0x2ff)) {
        c1 = (c1 / 10);
        xExp = (xExp.wrapping_add(1));
    }
    if (c1 < 0x20000000000000) {
        return (((xSign | ((go_checked_shl_u64(xExp, go_shift_count_u64((53) as u64))))) | c1), 0);
    }
    return ((((xSign | 0x6000000000000000) | ((go_checked_shl_u64(xExp, go_shift_count_u64((51) as u64))))) | (c1 & 0x7ffffffffffff)), 0);
}
