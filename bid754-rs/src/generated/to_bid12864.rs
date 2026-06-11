// Auto-generated from to_bid12864.go by go2rs. Do not edit.

use super::prelude::*;

pub fn bid64_to_bid128(mut x: u64) -> (BID_UINT128, u32) {
    let mut new_coeff: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut res: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut sign_x: u64 = 0;
    let mut exponent_x: i64 = 0;
    let mut coefficient_x: u64 = 0;
    let mut pfpsf: u32 = 0;
    let (mut sign_x, mut exponent_x, mut coefficient_x, mut valid) = unpack_bid64(x);
    if (!valid) {
        if (((go_checked_shl_u64(x, go_shift_count_u64((1) as u64)))) >= 0xf000000000000000) {
            if ((x & 0x7e00000000000000) == 0x7e00000000000000) {
                pfpsf |= 1;
            }
            res.w[0] = (coefficient_x & 0x0003ffffffffffff);
            (res.w[1], res.w[0]) = go_mul64(res.w[0], bid_power10_table_128[18].w[0]);
            res.w[1] |= (coefficient_x & 0xfc00000000000000);
            return (res, pfpsf);
        }
    }
    new_coeff.w[0] = coefficient_x;
    new_coeff.w[1] = 0;
    res.w[0] = new_coeff.w[0];
    res.w[1] = (sign_x | ((go_checked_shl_u64((((exponent_x.wrapping_add(6176)).wrapping_sub(398)) as u64), go_shift_count_u64((49) as u64)))));
    return (res, pfpsf);
}
