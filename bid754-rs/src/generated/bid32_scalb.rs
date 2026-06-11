// Auto-generated from bid32_scalb.go by go2rs. Do not edit.

use super::prelude::*;

pub fn bid32_scalbn(mut x: u32, mut n: i64, mut rnd_mode: i64) -> (u32, u32) {
    let mut sign_x: u32 = 0;
    let mut coefficient_x: u32 = 0;
    let mut res: u32 = 0;
    let mut exp64: i64 = 0;
    let mut exponent_x: i64 = 0;
    let mut rmode: i64 = 0;
    let mut pfpsf: u32 = 0;
    let (mut sign_x, mut exponent_x, mut coefficient_x, mut valid) = unpack_bid32(x);
    if (!valid) {
        if ((x & 0x7e000000) == 0x7e000000) {
            pfpsf |= 1;
        }
        if (coefficient_x != 0) {
            res = (coefficient_x & 0xfdffffff);
        } else {
            exp64 = ((exponent_x as i64).wrapping_add(n as i64));
            if (exp64 < 0) {
                exp64 = 0;
            }
            if (exp64 > (191 as i64)) {
                exp64 = (191 as i64);
            }
            exponent_x = (exp64 as i64);
            res = very_fast_get_bid32(sign_x, exponent_x, coefficient_x);
        }
        return (res, pfpsf);
    }
    exp64 = ((exponent_x as i64).wrapping_add(n as i64));
    exponent_x = (exp64 as i64);
    if ((exponent_x as u32) <= 191) {
        res = very_fast_get_bid32(sign_x, exponent_x, coefficient_x);
        return (res, pfpsf);
    }
    if (exp64 > (191 as i64)) {
        while ((coefficient_x < 1000000) && (exp64 > (191 as i64))) {
            coefficient_x = (((go_checked_shl_u32(coefficient_x, go_shift_count_u64((1) as u64)))).wrapping_add(((go_checked_shl_u32(coefficient_x, go_shift_count_u64((3) as u64))))));
            exponent_x = exponent_x.wrapping_sub(1);
            exp64 = exp64.wrapping_sub(1);
        }
        if (exp64 <= (191 as i64)) {
            res = very_fast_get_bid32(sign_x, exponent_x, coefficient_x);
            return (res, pfpsf);
        } else {
            exponent_x = 0x7fffffff;
        }
    }
    rmode = rnd_mode;
    res = get_bid32(sign_x, exponent_x, (coefficient_x as u64), rmode);
    return (res, pfpsf);
}
