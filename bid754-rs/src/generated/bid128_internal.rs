// Auto-generated from bid128_internal.go by go2rs. Do not edit.

use super::prelude::*;

pub(crate) fn unpack_bid128(mut x: BID_UINT128) -> (u64, i64, BID_UINT128, u64) {
    let mut sign: u64 = 0;
    let mut exponent: i64 = 0;
    let mut coefficient: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut valid: u64 = 0;
    sign = (x.w[1] & 0x8000000000000000);
    if ((x.w[1] & 0x7800000000000000) >= 0x6000000000000000) {
        if ((x.w[1] & 0x7800000000000000) < 0x7800000000000000) {
            coefficient.w[0] = 0;
            coefficient.w[1] = 0;
            let mut ex = (go_checked_shr_u64(x.w[1], go_shift_count_u64((47) as u64)));
            exponent = ((ex as i64) & 0x3fff);
            return (sign, exponent, coefficient, 0);
        }
        let mut T33 = bid_power10_table_128[33];
        let mut coeff = BID_UINT128 { w: [x.w[0], (x.w[1] & 0x1ffffffffffff)], ..Default::default() };
        coefficient.w[0] = x.w[0];
        coefficient.w[1] = x.w[1];
        if __unsigned_compare_ge_128(coeff, T33) {
            coefficient.w[1] &= (!(0x1ffffffffffff as u64));
            coefficient.w[0] = 0;
        }
        exponent = 0;
        return (sign, exponent, coefficient, 0);
    }
    let mut coeff = BID_UINT128 { w: [x.w[0], (x.w[1] & 0x1ffffffffffff)], ..Default::default() };
    let mut T34 = bid_power10_table_128[34];
    if __unsigned_compare_ge_128(coeff, T34) {
        coeff.w[0] = 0;
        coeff.w[1] = 0;
    }
    coefficient.w[0] = coeff.w[0];
    coefficient.w[1] = coeff.w[1];
    let mut ex = (go_checked_shr_u64(x.w[1], go_shift_count_u64((49) as u64)));
    exponent = ((ex as i64) & 0x3fff);
    return (sign, exponent, coefficient, (coeff.w[0] | coeff.w[1]));
}

pub(crate) fn very_fast_get_bid128(mut sgn: u64, mut expon: i64, mut coeff: BID_UINT128) -> BID_UINT128 {
    let mut res: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    res.w[0] = coeff.w[0];
    res.w[1] = ((sgn | ((go_checked_shl_u64((expon as u64), go_shift_count_u64((49) as u64))))) | coeff.w[1]);
    return res;
}

pub(crate) fn unpack_bid128_value(mut x: BID_UINT128) -> (u64, i64, BID_UINT128, bool) {
    let mut sign_x: u64 = 0;
    let mut exponent_x: i64 = 0;
    let mut coefficient_x: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut valid: bool = false;
    sign_x = (x.w[1] & 0x8000000000000000);
    if ((x.w[1] & 0x7800000000000000) >= 0x6000000000000000) {
        if ((x.w[1] & 0x7800000000000000) < 0x7800000000000000) {
            coefficient_x.w[0] = 0;
            coefficient_x.w[1] = 0;
            let mut ex = (go_checked_shr_u64(x.w[1], go_shift_count_u64((47) as u64)));
            exponent_x = ((ex as i64) & 0x3fff);
            return (sign_x, exponent_x, coefficient_x, false);
        }
        let mut T33 = bid_power10_table_128[33];
        coefficient_x.w[0] = x.w[0];
        coefficient_x.w[1] = (x.w[1] & 0x00003fffffffffff);
        if __unsigned_compare_ge_128(coefficient_x, T33) {
            coefficient_x.w[1] = (x.w[1] & 0xfe00000000000000);
            coefficient_x.w[0] = 0;
        } else {
            coefficient_x.w[1] = (x.w[1] & 0xfe003fffffffffff);
        }
        if ((x.w[1] & 0x7c00000000000000) == 0x7800000000000000) {
            coefficient_x.w[0] = 0;
            coefficient_x.w[1] = (x.w[1] & 0xf800000000000000);
        }
        exponent_x = 0;
        return (sign_x, exponent_x, coefficient_x, false);
    }
    let mut coeff = BID_UINT128 { w: [x.w[0], (x.w[1] & 0x1ffffffffffff)], ..Default::default() };
    let mut T34 = bid_power10_table_128[34];
    if __unsigned_compare_ge_128(coeff, T34) {
        coeff.w[0] = 0;
        coeff.w[1] = 0;
    }
    coefficient_x.w[0] = coeff.w[0];
    coefficient_x.w[1] = coeff.w[1];
    let mut ex = (go_checked_shr_u64(x.w[1], go_shift_count_u64((49) as u64)));
    exponent_x = ((ex as i64) & 0x3fff);
    return (sign_x, exponent_x, coefficient_x, ((coeff.w[0] | coeff.w[1]) != 0));
}

pub fn bid128_is_na_n(mut x: BID_UINT128) -> i64 {
    if ((x.w[1] & 0x7c00000000000000) == 0x7c00000000000000) {
        return 1;
    }
    return 0;
}

pub fn bid128_is_inf(mut x: BID_UINT128) -> i64 {
    if ((((x.w[1] & 0x7800000000000000) == 0x7800000000000000)) && (((x.w[1] & 0x7c00000000000000) != 0x7c00000000000000))) {
        return 1;
    }
    return 0;
}

pub fn bid128_is_zero(mut x: BID_UINT128) -> i64 {
    let (_, _, mut coeff, mut valid) = unpack_bid128(x);
    if (((valid == 0) && (bid128_is_na_n(x) == 0)) && (bid128_is_inf(x) == 0)) {
        return 1;
    }
    if ((coeff.w[0] == 0) && (coeff.w[1] == 0)) {
        return 1;
    }
    return 0;
}
