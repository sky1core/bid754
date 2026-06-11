// Auto-generated from logb64.go by go2rs. Do not edit.

use super::prelude::*;

pub fn bid64_i_logb(mut x: u64) -> (i64, u32) {
    let mut sign_x: u64 = 0;
    let mut coefficient_x: u64 = 0;
    let mut exponent_x: i64 = 0;
    let mut bin_expon_cx: i64 = 0;
    let mut digits: i64 = 0;
    let mut res: i64 = 0;
    let mut pfpsf: u32 = 0;
    let mut valid_x: bool = false;
    (sign_x, exponent_x, coefficient_x, valid_x) = unpack_bid64(x);
    _ = sign_x;
    if (!valid_x) {
        pfpsf |= 1;
        if ((x & 0x7c00000000000000) == 0x7800000000000000) {
            res = 0x7fffffff;
        } else {
            res = (-2147483648);
        }
        return (res, pfpsf);
    }
    if (coefficient_x >= 1000000000000000) {
        digits = 16;
    } else {
        let mut dx = (coefficient_x as f64).to_bits();
        bin_expon_cx = (((go_checked_shr_u64(dx, go_shift_count_u64((52) as u64))) as i64).wrapping_sub(1023));
        digits = (bid_estimate_decimal_digits[bin_expon_cx as usize] as i64);
        if (coefficient_x >= bid_power10_table_128[digits as usize].w[0]) {
            digits = digits.wrapping_add(1);
        }
    }
    exponent_x = (((exponent_x.wrapping_sub(0x18e)).wrapping_add(digits)).wrapping_sub(1));
    return (exponent_x, pfpsf);
}

pub fn bid64_logb(mut x: u64) -> (u64, u32) {
    let mut ires: i64 = 0;
    let mut exponent_x: i64 = 0;
    let mut sign_x: u64 = 0;
    let mut coefficient_x: u64 = 0;
    let mut valid_x: bool = false;
    let mut res: u64 = 0;
    let mut pfpsf: u32 = 0;
    (sign_x, exponent_x, coefficient_x, valid_x) = unpack_bid64(x);
    _ = sign_x;
    _ = exponent_x;
    if (!valid_x) {
        if ((x & 0x7800000000000000) == 0x7800000000000000) {
            if ((x & 0x7e00000000000000) == 0x7e00000000000000) {
                pfpsf |= 1;
            }
            res = (coefficient_x & 0xfdffffffffffffff);
            if ((x & 0x7c00000000000000) == 0x7800000000000000) {
                res &= 0x7fffffffffffffff;
            }
            return (res, pfpsf);
        }
        pfpsf |= 4;
        res = 0xf800000000000000;
        return (res, pfpsf);
    }
    (ires, _) = bid64_i_logb(x);
    if (ires < 0) {
        res = (0xb1c0000000000000 | ((ires.wrapping_neg()) as u64));
    } else {
        res = (0x31c0000000000000 | (ires as u64));
    }
    return (res, pfpsf);
}
