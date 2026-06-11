// Auto-generated from bid32_logb.go by go2rs. Do not edit.

use super::prelude::*;

pub fn bid32_i_logb(mut x: u32) -> (i64, u32) {
    let mut sign_x: u32 = 0;
    let mut coefficient_x: u32 = 0;
    let mut bin_expon_cx: i64 = 0;
    let mut digits: i64 = 0;
    let mut exponent_x: i64 = 0;
    let mut pfpsf: u32 = 0;
    let (mut sign_x, mut exponent_x, mut coefficient_x, mut valid) = unpack_bid32(x);
    _ = sign_x;
    if (!valid) {
        pfpsf |= 1;
        let mut res: i64 = 0;
        if ((x & 0x7c000000) == 0x78000000) {
            res = 0x7fffffff;
        } else {
            res = (-0x80000000);
        }
        return (res, pfpsf);
    }
    if (coefficient_x >= 1000000) {
        digits = 7;
    } else {
        let mut dx = ((coefficient_x as f32) as f32).to_bits();
        bin_expon_cx = (((go_checked_shr_u32(dx, go_shift_count_u64((23) as u64))) as i64).wrapping_sub(127));
        digits = (bid_estimate_decimal_digits[bin_expon_cx as usize] as i64);
        if ((coefficient_x as u64) >= bid_power10_table_128[digits as usize].w[0]) {
            digits = digits.wrapping_add(1);
        }
    }
    exponent_x = (((exponent_x.wrapping_sub(101)).wrapping_add(digits)).wrapping_sub(1));
    return (exponent_x, pfpsf);
}

pub fn bid32_logb(mut x: u32) -> (u32, u32) {
    let mut sign_x: u32 = 0;
    let mut coefficient_x: u32 = 0;
    let mut exponent_x: i64 = 0;
    let mut res: u32 = 0;
    let mut pfpsf: u32 = 0;
    let (mut sign_x, mut exponent_x, mut coefficient_x, mut valid) = unpack_bid32(x);
    _ = sign_x;
    _ = exponent_x;
    if (!valid) {
        if ((x & 0x78000000) == 0x78000000) {
            if ((x & 0x7e000000) == 0x7e000000) {
                pfpsf |= 1;
            }
            res = (coefficient_x & 0xfdffffff);
            if ((x & 0x7c000000) == 0x78000000) {
                res &= 0x7fffffff;
            }
            return (res, pfpsf);
        }
        pfpsf |= 4;
        res = 0xf8000000;
        return (res, pfpsf);
    }
    let (mut ires, mut iflags) = bid32_i_logb(x);
    pfpsf |= iflags;
    if (ires < 0) {
        res = (0xb2800000 | ((ires.wrapping_neg()) as u32));
    } else {
        res = (0x32800000 | (ires as u32));
    }
    return (res, pfpsf);
}
