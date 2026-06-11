// Auto-generated from bid32_rem.go by go2rs. Do not edit.

use super::prelude::*;

pub fn bid32_rem(mut x: u32, mut y: u32) -> (u32, u32) {
    let mut CX: u64 = 0;
    let mut Q64: u64 = 0;
    let mut CYL: u64 = 0;
    let mut CY: u32 = 0;
    let mut sign_x: u32 = 0;
    let mut sign_y: u32 = 0;
    let mut coefficient_x: u32 = 0;
    let mut coefficient_y: u32 = 0;
    let mut res: u32 = 0;
    let mut Q: u32 = 0;
    let mut R: u32 = 0;
    let mut R2: u32 = 0;
    let mut T: u32 = 0;
    let mut exponent_x: i64 = 0;
    let mut exponent_y: i64 = 0;
    let mut bin_expon: i64 = 0;
    let mut e_scale: i64 = 0;
    let mut digits_x: i64 = 0;
    let mut diff_expon: i64 = 0;
    let mut pfpsf: u32 = 0;
    let (mut sign_y, mut exponent_y, mut coefficient_y, mut valid_y) = unpack_bid32(y);
    let (mut sign_x, mut exponent_x, mut coefficient_x, mut valid_x) = unpack_bid32(x);
    _ = sign_y;
    if (coefficient_x == 0) {
        valid_x = false;
    }
    if (coefficient_y == 0) {
        valid_y = false;
    }
    if (!valid_x) {
        if ((y & 0x7e000000) == 0x7e000000) {
            pfpsf |= 1;
        }
        if ((x & 0x7c000000) == 0x7c000000) {
            if ((x & 0x7e000000) == 0x7e000000) {
                pfpsf |= 1;
            }
            res = (coefficient_x & 0xfdffffff);
            return (res, pfpsf);
        }
        if ((x & 0x78000000) == 0x78000000) {
            if ((y & 0x7c000000) != 0x7c000000) {
                pfpsf |= 1;
                res = 0x7c000000;
                return (res, pfpsf);
            }
        }
        if ((((y & 0x78000000) < 0x78000000)) && (coefficient_y != 0)) {
            if ((y & 0x60000000) == 0x60000000) {
                exponent_y = ((((go_checked_shr_u32(y, go_shift_count_u64((21) as u64)))) & 0xff) as i64);
            } else {
                exponent_y = ((((go_checked_shr_u32(y, go_shift_count_u64((23) as u64)))) & 0xff) as i64);
            }
            if (exponent_y < exponent_x) {
                exponent_x = exponent_y;
            }
            res = (((go_checked_shl_u32((exponent_x as u32), go_shift_count_u64((23) as u64)))) | sign_x);
            return (res, pfpsf);
        }
    }
    if (!valid_y) {
        if ((y & 0x7c000000) == 0x7c000000) {
            if ((y & 0x7e000000) == 0x7e000000) {
                pfpsf |= 1;
            }
            res = (coefficient_y & 0xfdffffff);
            return (res, pfpsf);
        }
        if ((y & 0x78000000) == 0x78000000) {
            res = very_fast_get_bid32(sign_x, exponent_x, coefficient_x);
            return (res, pfpsf);
        }
        pfpsf |= 1;
        res = 0x7c000000;
        return (res, pfpsf);
    }
    diff_expon = (exponent_x.wrapping_sub(exponent_y));
    if (diff_expon <= 0) {
        diff_expon = (diff_expon.wrapping_neg());
        if (diff_expon > 7) {
            res = x;
            return (res, pfpsf);
        }
        T = (bid_power10_table_128[diff_expon as usize].w[0] as u32);
        CYL = ((coefficient_y as u64).wrapping_mul(T as u64));
        if (CYL > ((go_checked_shl_u32(coefficient_x, go_shift_count_u64((1) as u64))) as u64)) {
            res = x;
            return (res, pfpsf);
        }
        CY = (CYL as u32);
        Q = (coefficient_x / CY);
        R = (coefficient_x.wrapping_sub((Q.wrapping_mul(CY))));
        R2 = (R.wrapping_add(R));
        if ((R2 > CY) || (((R2 == CY) && ((Q & 1) != 0)))) {
            R = (CY.wrapping_sub(R));
            sign_x ^= 0x80000000;
        }
        res = very_fast_get_bid32(sign_x, exponent_x, R);
        return (res, pfpsf);
    }
    CX = (coefficient_x as u64);
    while (diff_expon > 0) {
        let mut tempx = ((CX as f32) as f32).to_bits();
        bin_expon = (((((go_checked_shr_u32(tempx, go_shift_count_u64((23) as u64)))) & 0xff) as i64).wrapping_sub(0x7f));
        digits_x = (bid_estimate_decimal_digits[bin_expon as usize] as i64);
        e_scale = ((18 as i64).wrapping_sub(digits_x));
        if (diff_expon >= e_scale) {
            diff_expon = diff_expon.wrapping_sub(e_scale);
        } else {
            e_scale = diff_expon;
            diff_expon = 0;
        }
        CX = CX.wrapping_mul(bid_power10_table_128[e_scale as usize].w[0]);
        Q64 = (CX / (coefficient_y as u64));
        CX = CX.wrapping_sub((Q64.wrapping_mul(coefficient_y as u64)));
        if (CX == 0) {
            res = very_fast_get_bid32(sign_x, exponent_y, 0);
            return (res, pfpsf);
        }
    }
    coefficient_x = (CX as u32);
    R2 = (coefficient_x.wrapping_add(coefficient_x));
    if ((R2 > coefficient_y) || (((R2 == coefficient_y) && ((Q64 & 1) != 0)))) {
        coefficient_x = (coefficient_y.wrapping_sub(coefficient_x));
        sign_x ^= 0x80000000;
    }
    res = very_fast_get_bid32(sign_x, exponent_y, coefficient_x);
    return (res, pfpsf);
}
