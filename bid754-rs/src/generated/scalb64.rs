// Auto-generated from scalb64.go by go2rs. Do not edit.

use super::prelude::*;

pub fn bid64_scalbn(mut x: u64, mut n: i64, mut rndMode: i64) -> (u64, u32) {
    let mut sign_x: u64 = 0;
    let mut coefficient_x: u64 = 0;
    let mut res: u64 = 0;
    let mut exp64: i64 = 0;
    let mut exponent_x: i64 = 0;
    let mut pfpsf: u32 = 0;
    let (mut sign_x, mut exponent_x, mut coefficient_x, mut valid_x) = unpack_bid64(x);
    if (!valid_x) {
        if ((x & 0x7e00000000000000) == 0x7e00000000000000) {
            pfpsf |= 1;
        }
        if (coefficient_x != 0) {
            res = (coefficient_x & 0xfdffffffffffffff);
        } else {
            exp64 = ((exponent_x as i64).wrapping_add(n as i64));
            if (exp64 < 0) {
                exp64 = 0;
            }
            if (exp64 > 0x2ff) {
                exp64 = 0x2ff;
            }
            exponent_x = (exp64 as i64);
            res = very_fast_get_bid64(sign_x, exponent_x, coefficient_x);
        }
        return (res, pfpsf);
    }
    exp64 = ((exponent_x as i64).wrapping_add(n as i64));
    exponent_x = (exp64 as i64);
    if ((exponent_x as u32) <= 0x2ff) {
        res = very_fast_get_bid64(sign_x, exponent_x, coefficient_x);
        return (res, pfpsf);
    }
    if (exp64 > 0x2ff) {
        while ((coefficient_x < 1000000000000000) && (exp64 > 0x2ff)) {
            coefficient_x = (((go_checked_shl_u64(coefficient_x, go_shift_count_u64((1) as u64)))).wrapping_add(((go_checked_shl_u64(coefficient_x, go_shift_count_u64((3) as u64))))));
            exponent_x = exponent_x.wrapping_sub(1);
            exp64 = exp64.wrapping_sub(1);
        }
        if (exp64 <= 0x2ff) {
            res = very_fast_get_bid64(sign_x, exponent_x, coefficient_x);
            return (res, pfpsf);
        }
        exponent_x = 0x7fffffff;
    }
    (res, pfpsf) = get_bid64_flags(sign_x, exponent_x, coefficient_x, rndMode);
    return (res, pfpsf);
}

pub fn bid64_scalbln(mut x: u64, mut n: i64, mut rndMode: i64) -> (u64, u32) {
    let mut res: u64 = 0;
    let mut n1: i32 = 0;
    n1 = (n as i32);
    if ((n1 as i64) < n) {
        n1 = 0x7fffffff;
    } else if ((n1 as i64) > n) {
        n1 = ((-0x7fffffff) - 1);
    }
    let (mut res, mut pfpsf) = bid64_scalbn(x, (n1 as i64), rndMode);
    return (res, pfpsf);
}

pub fn bid64_ldexp(mut x: u64, mut n: i64, mut rndMode: i64) -> (u64, u32) {
    let mut sign_x: u64 = 0;
    let mut coefficient_x: u64 = 0;
    let mut res: u64 = 0;
    let mut exp64: i64 = 0;
    let mut exponent_x: i64 = 0;
    let mut rmode: i64 = 0;
    let mut pfpsf: u32 = 0;
    let (mut sign_x, mut exponent_x, mut coefficient_x, mut valid_x) = unpack_bid64(x);
    if (!valid_x) {
        if ((x & 0x7e00000000000000) == 0x7e00000000000000) {
            pfpsf |= 1;
        }
        if (coefficient_x != 0) {
            res = (coefficient_x & 0xfdffffffffffffff);
        } else {
            exp64 = ((exponent_x as i64).wrapping_add(n as i64));
            if (exp64 < 0) {
                exp64 = 0;
            }
            if (exp64 > 0x2ff) {
                exp64 = 0x2ff;
            }
            exponent_x = (exp64 as i64);
            res = very_fast_get_bid64(sign_x, exponent_x, coefficient_x);
        }
        return (res, pfpsf);
    }
    exp64 = ((exponent_x as i64).wrapping_add(n as i64));
    exponent_x = (exp64 as i64);
    if ((exponent_x as u32) <= 0x2ff) {
        res = very_fast_get_bid64(sign_x, exponent_x, coefficient_x);
        return (res, pfpsf);
    }
    if (exp64 > 0x2ff) {
        while ((coefficient_x < 1000000000000000) && (exp64 > 0x2ff)) {
            coefficient_x = (((go_checked_shl_u64(coefficient_x, go_shift_count_u64((1) as u64)))).wrapping_add(((go_checked_shl_u64(coefficient_x, go_shift_count_u64((3) as u64))))));
            exponent_x = exponent_x.wrapping_sub(1);
            exp64 = exp64.wrapping_sub(1);
        }
        if (exp64 <= 0x2ff) {
            res = very_fast_get_bid64(sign_x, exponent_x, coefficient_x);
            return (res, pfpsf);
        }
        exponent_x = 0x7fffffff;
    }
    rmode = rndMode;
    (res, pfpsf) = get_bid64_flags(sign_x, exponent_x, coefficient_x, rmode);
    return (res, pfpsf);
}
