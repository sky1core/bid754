// Auto-generated from bid128_modf.go by go2rs. Do not edit.

use super::prelude::*;

pub fn bid128_modf(mut x: BID_UINT128) -> (BID_UINT128, BID_UINT128, u32) {
    let mut res: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut xi: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut pfpsf: u32 = 0;
    xi = bid128_round_integral_zero(x, (&mut pfpsf));
    if ((x.w[1] & 0x7c00000000000000) == 0x7800000000000000) {
        res.w[1] = ((x.w[1] & 0x8000000000000000) | 0x5ffe000000000000);
        res.w[0] = 0;
    } else {
        res = bid128_sub(x, xi, 0, (&mut pfpsf));
    }
    xi.w[1] |= (x.w[1] & 0x8000000000000000);
    res.w[1] |= (x.w[1] & 0x8000000000000000);
    return (res, xi, pfpsf);
}
