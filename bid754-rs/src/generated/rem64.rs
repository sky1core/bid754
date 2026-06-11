// Auto-generated from rem64.go by go2rs. Do not edit.

use super::prelude::*;

pub fn bid64_rem(mut x: u64, mut y: u64) -> (u64, u32) {
    let mut CY: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut sign_x: u64 = 0;
    let mut sign_y: u64 = 0;
    let mut coefficient_x: u64 = 0;
    let mut coefficient_y: u64 = 0;
    let mut res: u64 = 0;
    let mut Q: u64 = 0;
    let mut R: u64 = 0;
    let mut R2: u64 = 0;
    let mut T: u64 = 0;
    let mut valid_y: bool = false;
    let mut valid_x: bool = false;
    let mut exponent_x: i64 = 0;
    let mut exponent_y: i64 = 0;
    let mut bin_expon: i64 = 0;
    let mut e_scale: i64 = 0;
    let mut digits_x: i64 = 0;
    let mut diff_expon: i64 = 0;
    let mut pfpsf: u32 = 0;
    (sign_y, exponent_y, coefficient_y, valid_y) = unpack_bid64(y);
    (sign_x, exponent_x, coefficient_x, valid_x) = unpack_bid64(x);
    _ = sign_y;
    if (!valid_x) {
        if ((y & 0x7e00000000000000) == 0x7e00000000000000) {
            pfpsf |= 1;
        }
        if ((x & 0x7c00000000000000) == 0x7c00000000000000) {
            if ((x & 0x7e00000000000000) == 0x7e00000000000000) {
                pfpsf |= 1;
            }
            res = (coefficient_x & 0xfdffffffffffffff);
            return (res, pfpsf);
        }
        if ((x & 0x7800000000000000) == 0x7800000000000000) {
            if ((y & 0x7c00000000000000) != 0x7c00000000000000) {
                pfpsf |= 1;
                res = 0x7c00000000000000;
                return (res, pfpsf);
            }
        }
        if ((((y & 0x7800000000000000) < 0x7800000000000000)) && (coefficient_y != 0)) {
            if ((y & 0x6000000000000000) == 0x6000000000000000) {
                exponent_y = ((((go_checked_shr_u64(y, go_shift_count_u64((51) as u64)))) & 0x3ff) as i64);
            } else {
                exponent_y = ((((go_checked_shr_u64(y, go_shift_count_u64((53) as u64)))) & 0x3ff) as i64);
            }
            if (exponent_y < exponent_x) {
                exponent_x = exponent_y;
            }
            x = (exponent_x as u64);
            x = go_checked_shl_u64(x, go_shift_count_u64((53) as u64));
            res = (x | sign_x);
            return (res, pfpsf);
        }
    }
    if (!valid_y) {
        if ((y & 0x7c00000000000000) == 0x7c00000000000000) {
            if ((y & 0x7e00000000000000) == 0x7e00000000000000) {
                pfpsf |= 1;
            }
            res = (coefficient_y & 0xfdffffffffffffff);
            return (res, pfpsf);
        }
        if ((y & 0x7800000000000000) == 0x7800000000000000) {
            res = very_fast_get_bid64(sign_x, exponent_x, coefficient_x);
            return (res, pfpsf);
        }
        pfpsf |= 1;
        res = 0x7c00000000000000;
        return (res, pfpsf);
    }
    diff_expon = (exponent_x.wrapping_sub(exponent_y));
    if (diff_expon <= 0) {
        diff_expon = (diff_expon.wrapping_neg());
        if (diff_expon > 16) {
            res = x;
            return (res, pfpsf);
        }
        T = bid_power10_table_128[diff_expon as usize].w[0];
        CY = __mul_64x64_to_128(coefficient_y, T);
        if ((CY.w[1] != 0) || (CY.w[0] > ((go_checked_shl_u64(coefficient_x, go_shift_count_u64((1) as u64)))))) {
            res = x;
            return (res, pfpsf);
        }
        Q = (coefficient_x / CY.w[0]);
        R = (coefficient_x.wrapping_sub((Q.wrapping_mul(CY.w[0]))));
        R2 = (R.wrapping_add(R));
        if ((R2 > CY.w[0]) || (((R2 == CY.w[0]) && ((Q & 1) != 0)))) {
            R = (CY.w[0].wrapping_sub(R));
            sign_x ^= 0x8000000000000000;
        }
        res = very_fast_get_bid64(sign_x, exponent_x, R);
        return (res, pfpsf);
    }
    while (diff_expon > 0) {
        let mut tempx = ((coefficient_x as f32) as f32).to_bits();
        bin_expon = (((((go_checked_shr_u32(tempx, go_shift_count_u64((23) as u64)))) & 0xff) as i64).wrapping_sub(0x7f));
        digits_x = (bid_estimate_decimal_digits[bin_expon as usize] as i64);
        e_scale = ((18 as i64).wrapping_sub(digits_x));
        if (diff_expon >= e_scale) {
            diff_expon = diff_expon.wrapping_sub(e_scale);
        } else {
            e_scale = diff_expon;
            diff_expon = 0;
        }
        coefficient_x = coefficient_x.wrapping_mul(bid_power10_table_128[e_scale as usize].w[0]);
        Q = (coefficient_x / coefficient_y);
        coefficient_x = coefficient_x.wrapping_sub((Q.wrapping_mul(coefficient_y)));
        if (coefficient_x == 0) {
            res = very_fast_get_bid64_small_mantissa(sign_x, exponent_y, 0);
            return (res, pfpsf);
        }
    }
    R2 = (coefficient_x.wrapping_add(coefficient_x));
    if ((R2 > coefficient_y) || (((R2 == coefficient_y) && ((Q & 1) != 0)))) {
        coefficient_x = (coefficient_y.wrapping_sub(coefficient_x));
        sign_x ^= 0x8000000000000000;
    }
    res = very_fast_get_bid64(sign_x, exponent_y, coefficient_x);
    return (res, pfpsf);
}
