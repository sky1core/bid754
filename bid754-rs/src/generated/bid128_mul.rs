// Auto-generated from bid128_mul.go by go2rs. Do not edit.

use super::prelude::*;

pub fn bid128_mul(mut x: BID_UINT128, mut y: BID_UINT128, mut rnd_mode: i64) -> (BID_UINT128, u32) {
    let mut res: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut x_sign: u64 = 0;
    let mut y_sign: u64 = 0;
    let mut p_sign: u64 = 0;
    let mut x_exp: u64 = 0;
    let mut y_exp: u64 = 0;
    let mut p_exp: u64 = 0;
    let mut true_p_exp: i64 = 0;
    let mut C1: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut C2: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut pfpsf: u32 = 0;
    if (!(((((((x.w[1] & 0x7c00000000000000) == 0x7c00000000000000)) || (((y.w[1] & 0x7c00000000000000) == 0x7c00000000000000))) || (((x.w[1] & 0x7c00000000000000) == 0x7800000000000000))) || (((y.w[1] & 0x7c00000000000000) == 0x7800000000000000))))) {
        x_sign = (x.w[1] & 0x8000000000000000);
        C1.w[1] = (x.w[1] & 0x1ffffffffffff);
        C1.w[0] = x.w[0];
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
        y_sign = (y.w[1] & 0x8000000000000000);
        C2.w[1] = (y.w[1] & 0x1ffffffffffff);
        C2.w[0] = y.w[0];
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
        p_sign = (x_sign ^ y_sign);
        true_p_exp = (((((go_checked_shr_u64(x_exp, go_shift_count_u64((49) as u64))) as i64).wrapping_sub(6176)).wrapping_add(((go_checked_shr_u64(y_exp, go_shift_count_u64((49) as u64))) as i64))).wrapping_sub(6176));
        if (true_p_exp < (-6176)) {
            p_exp = 0;
        } else if (true_p_exp > 6111) {
            p_exp = (((6111 + 6176) as u64) << 49);
        } else {
            p_exp = (go_checked_shl_u64(((true_p_exp.wrapping_add(6176)) as u64), go_shift_count_u64((49) as u64)));
        }
        if ((((C1.w[1] == 0) && (C1.w[0] == 0))) || (((C2.w[1] == 0) && (C2.w[0] == 0)))) {
            res.w[1] = (p_sign | p_exp);
            res.w[0] = 0;
            return (res, pfpsf);
        }
    }
    let mut z = BID_UINT128 { w: [0x0000000000000000, 0x5ffe000000000000], ..Default::default() };
    return bid128_fma(y, x, z, rnd_mode);
}
