// Auto-generated from bid128_ldexp.go by go2rs. Do not edit.

use super::prelude::*;

pub fn bid128_ldexp(mut x: BID_UINT128, mut n: i64, mut rnd_mode: i64) -> (BID_UINT128, u32) {
    let mut CX: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut CX2: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut CBID_X8: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut res: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut exp64: i64 = 0;
    let mut sign_x: u64 = 0;
    let mut exponent_x: i64 = 0;
    let mut pfpsf: u32 = 0;
    let (mut sign_x, mut exponent_x, mut CX, mut valid) = unpack_bid128_value(x);
    if (!valid) {
        if ((x.w[1] & 0x7e00000000000000) == 0x7e00000000000000) {
            pfpsf |= 1;
        }
        res.w[1] = (CX.w[1] & 0xfdffffffffffffff);
        res.w[0] = CX.w[0];
        if (CX.w[1] == 0) {
            exp64 = ((exponent_x as i64).wrapping_add(n as i64));
            if (exp64 < 0) {
                exp64 = 0;
            }
            if (exp64 > 0x2fff) {
                exp64 = 0x2fff;
            }
            exponent_x = (exp64 as i64);
            res = very_fast_get_bid128(sign_x, exponent_x, CX);
        }
        return (res, pfpsf);
    }
    exp64 = ((exponent_x as i64).wrapping_add(n as i64));
    exponent_x = (exp64 as i64);
    if ((exponent_x as u32) <= 0x2fff) {
        res = very_fast_get_bid128(sign_x, exponent_x, CX);
        return (res, pfpsf);
    }
    if (exp64 > 0x2fff) {
        if (CX.w[1] < 0x314dc6448d93) {
            loop {
                CBID_X8.w[1] = (((go_checked_shl_u64(CX.w[1], go_shift_count_u64((3) as u64)))) | ((go_checked_shr_u64(CX.w[0], go_shift_count_u64((61) as u64)))));
                CBID_X8.w[0] = (go_checked_shl_u64(CX.w[0], go_shift_count_u64((3) as u64)));
                CX2.w[1] = (((go_checked_shl_u64(CX.w[1], go_shift_count_u64((1) as u64)))) | ((go_checked_shr_u64(CX.w[0], go_shift_count_u64((63) as u64)))));
                CX2.w[0] = (go_checked_shl_u64(CX.w[0], go_shift_count_u64((1) as u64)));
                CX = __add_128_128(CX2, CBID_X8);
                exponent_x = exponent_x.wrapping_sub(1);
                exp64 = exp64.wrapping_sub(1);
                if (!(((CX.w[1] < 0x314dc6448d93) && (exp64 > 0x2fff)))) {
                    break;
                }
            }
        }
        if (exp64 <= 0x2fff) {
            res = very_fast_get_bid128(sign_x, exponent_x, CX);
            return (res, pfpsf);
        } else {
            exponent_x = 0x7fffffff;
        }
    }
    let mut rmode = rnd_mode;
    res = bid_get_bid128(sign_x, exponent_x, CX, rmode, (&mut pfpsf));
    return (res, pfpsf);
}
