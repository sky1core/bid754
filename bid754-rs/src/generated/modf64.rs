// Auto-generated from modf64.go by go2rs. Do not edit.

use super::prelude::*;

pub(crate) fn bid64_unpack_finite_for_round_local(mut x: u64) -> (u64, i64, u64) {
    let mut xSign = (x & 0x8000000000000000);
    if ((x & 0x6000000000000000) == 0x6000000000000000) {
        let mut exp = (((go_checked_shr_u64((x & 0x1ff8000000000000), go_shift_count_u64((51) as u64))) as i64).wrapping_sub(398));
        let mut coeff = ((x & 0x7ffffffffffff) | 0x20000000000000);
        if (coeff > 9999999999999999) {
            coeff = 0;
        }
        return (xSign, exp, coeff);
    }
    let mut exp = (((go_checked_shr_u64((x & 0x7fe0000000000000), go_shift_count_u64((53) as u64))) as i64).wrapping_sub(398));
    let mut coeff = (x & 0x1fffffffffffff);
    return (xSign, exp, coeff);
}

pub fn bid64_modf(mut x: u64) -> (u64, u64, u32) {
    let (mut xi, mut flags) = bid64_round_integral_zero(x);
    let mut res: u64 = 0;
    if ((x & 0x7c00000000000000) == 0x7800000000000000) {
        res = ((x & 0x8000000000000000) | 0x5fe0000000000000);
    } else {
        let (mut r, mut subFlags) = bid64_sub_with_flags(x, xi, 0);
        res = r;
        flags |= subFlags;
    }
    let mut iptr = (xi | (x & 0x8000000000000000));
    res |= (x & 0x8000000000000000);
    return (res, iptr, flags);
}
