// Auto-generated from bid32_to_bid64.go by go2rs. Do not edit.

use super::prelude::*;

pub fn bid32_to_bid64(mut x: u32) -> (u64, u32) {
    let mut res: u64 = 0;
    let mut sign_x: u32 = 0;
    let mut coefficient_x: u32 = 0;
    let mut exponent_x: i64 = 0;
    let mut pfpsf: u32 = 0;
    let (mut sign_x, mut exponent_x, mut coefficient_x, mut valid) = unpack_bid32(x);
    if (!valid) {
        if ((x & 0x78000000) == 0x78000000) {
            if ((x & 0x7e000000) == 0x7e000000) {
                pfpsf |= 1;
            }
            res = ((coefficient_x & 0x000fffff) as u64);
            res = res.wrapping_mul(1000000000);
            res |= (((go_checked_shl_u64((coefficient_x as u64), go_shift_count_u64((32) as u64)))) & 0xfc00000000000000);
            return (res, pfpsf);
        }
    }
    res = very_fast_get_bid64_small_mantissa((go_checked_shl_u64((sign_x as u64), go_shift_count_u64((32) as u64))), ((exponent_x.wrapping_add(0x18e)).wrapping_sub(101)), (coefficient_x as u64));
    return (res, pfpsf);
}
