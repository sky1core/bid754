// Auto-generated from bid32_quantize.go by go2rs. Do not edit.

use super::prelude::*;

pub fn bid32_quantize(mut x: u32, mut y: u32, mut rnd_mode: i64) -> (u32, u32) {
    let mut CT: u64 = 0;
    let mut sign_x: u32 = 0;
    let mut coefficient_x: u32 = 0;
    let mut coefficient_y: u32 = 0;
    let mut remainder_h: u32 = 0;
    let mut C64: u32 = 0;
    let mut CT0: u32 = 0;
    let mut carry: u32 = 0;
    let mut res: u32 = 0;
    let mut exponent_x: i64 = 0;
    let mut exponent_y: i64 = 0;
    let mut digits_x: i64 = 0;
    let mut extra_digits: i64 = 0;
    let mut amount: i64 = 0;
    let mut amount2: i64 = 0;
    let mut expon_diff: i64 = 0;
    let mut total_digits: i64 = 0;
    let mut bin_expon_cx: i64 = 0;
    let mut rmode: i64 = 0;
    let mut status: u32 = 0;
    let mut pfpsf: u32 = 0;
    let mut valid_x: bool = false;
    (sign_x, exponent_x, coefficient_x, valid_x) = unpack_bid32(x);
    let (_, mut exponent_y, mut coefficient_y, mut valid_y) = unpack_bid32(y);
    if (coefficient_x == 0) {
        valid_x = false;
    }
    if (coefficient_y == 0) {
        valid_y = false;
    }
    if (!valid_y) {
        if ((x & 0x7e000000) == 0x7e000000) {
            pfpsf |= 1;
        }
        if (((((go_checked_shl_u32(coefficient_x, go_shift_count_u64((1) as u64)))) == 0xf0000000)) && ((((go_checked_shl_u32(coefficient_y, go_shift_count_u64((1) as u64)))) == 0xf0000000))) {
            res = coefficient_x;
            return (res, pfpsf);
        }
        if ((y & 0x78000000) == 0x78000000) {
            if ((((y & 0x7e000000) == 0x7e000000)) || (((((y & 0x7c000000) == 0x78000000)) && (((x & 0x7c000000) < 0x78000000))))) {
                pfpsf |= 1;
            }
            if ((y & 0x7c000000) != 0x7c000000) {
                coefficient_y = 0;
            }
            if ((x & 0x7c000000) != 0x7c000000) {
                res = (0x7c000000 | (coefficient_y & 0xfdffffff));
                if ((((y & 0x7c000000) != 0x7c000000)) && (((x & 0x7c000000) == 0x78000000))) {
                    res = x;
                }
                return (res, pfpsf);
            }
        }
    }
    if (!valid_x) {
        if ((x & 0x78000000) == 0x78000000) {
            if ((((x & 0x7e000000) == 0x7e000000)) || (((x & 0x7c000000) == 0x78000000))) {
                pfpsf |= 1;
            }
            if ((x & 0x7c000000) != 0x7c000000) {
                coefficient_x = 0;
            }
            res = (0x7c000000 | (coefficient_x & 0xfdffffff));
            return (res, pfpsf);
        }
        res = very_fast_get_bid32(sign_x, exponent_y, 0);
        return (res, pfpsf);
    }
    let mut tempx = ((coefficient_x as f32) as f32).to_bits();
    bin_expon_cx = (((((go_checked_shr_u32(tempx, go_shift_count_u64((23) as u64)))) & 0xff) as i64).wrapping_sub(0x7f));
    digits_x = (bid_estimate_decimal_digits[bin_expon_cx as usize] as i64);
    if ((coefficient_x as u64) >= bid_power10_table_128[digits_x as usize].w[0]) {
        digits_x = digits_x.wrapping_add(1);
    }
    expon_diff = (exponent_x.wrapping_sub(exponent_y));
    total_digits = (digits_x.wrapping_add(expon_diff));
    if (((total_digits.wrapping_add(1)) as u32) <= 8) {
        if (expon_diff >= 0) {
            coefficient_x = coefficient_x.wrapping_mul(bid_power10_table_128[expon_diff as usize].w[0] as u32);
            res = very_fast_get_bid32(sign_x, exponent_y, coefficient_x);
            return (res, pfpsf);
        }
        extra_digits = (expon_diff.wrapping_neg());
        rmode = rnd_mode;
        if ((sign_x != 0) && (((rmode.wrapping_sub(1)) as u64) < 2)) {
            rmode = ((3 as i64).wrapping_sub(rmode));
        }
        coefficient_x = coefficient_x.wrapping_add(bid_round_const_table[rmode as usize][extra_digits as usize] as u32);
        CT = ((coefficient_x as u64).wrapping_mul(bid_bid_reciprocals10_32[extra_digits as usize]));
        amount = (bid_bid_bid_recip_scale32[extra_digits as usize] as i64);
        CT0 = ((go_checked_shr_u64(CT, go_shift_count_u64((32) as u64))) as u32);
        C64 = (go_checked_shr_u32(CT0, go_shift_count_u64((amount as u64) as u64)));
        if (rnd_mode == 0) {
            if ((C64 & 1) != 0) {
                amount2 = ((32 as i64).wrapping_sub(amount));
                remainder_h = 0;
                remainder_h = remainder_h.wrapping_sub(1);
                remainder_h = go_checked_shr_u32(remainder_h, go_shift_count_u64((amount2 as u64) as u64));
                remainder_h = (remainder_h & CT0);
                if ((remainder_h == 0) && (((CT as u32) < (bid_bid_reciprocals10_32[extra_digits as usize] as u32)))) {
                    C64 = C64.wrapping_sub(1);
                }
            }
        }
        status = 32;
        remainder_h = (go_checked_shl_u32(CT0, go_shift_count_u64(((((32 as i64).wrapping_sub(amount)) as u64)) as u64)));
        match rmode {
            0 | 4 => {
                if ((remainder_h == 0x80000000) && (((CT as u32) < (bid_bid_reciprocals10_32[extra_digits as usize] as u32)))) {
                    status = 0;
                }
            }
            1 | 3 => {
                if ((remainder_h == 0) && (((CT as u32) < (bid_bid_reciprocals10_32[extra_digits as usize] as u32)))) {
                    status = 0;
                }
            }
            _ => {
                if (((CT as u32).wrapping_add(bid_bid_reciprocals10_32[extra_digits as usize] as u32)) < (CT as u32)) {
                    carry = 1;
                } else {
                    carry = 0;
                }
                if ((((go_checked_shr_u32(remainder_h, go_shift_count_u64(((((32 as i64).wrapping_sub(amount)) as u64)) as u64)))).wrapping_add(carry)) >= ((go_checked_shl_u32((1 as u32), go_shift_count_u64((amount as u64) as u64))))) {
                    status = 0;
                }
            }
        }
        pfpsf |= status;
        res = very_fast_get_bid32(sign_x, exponent_y, C64);
        return (res, pfpsf);
    }
    if (total_digits < 0) {
        pfpsf |= 32;
        C64 = 0;
        rmode = rnd_mode;
        if ((sign_x != 0) && (((rmode.wrapping_sub(1)) as u64) < 2)) {
            rmode = ((3 as i64).wrapping_sub(rmode));
        }
        if (rmode == 2) {
            C64 = 1;
        }
        res = very_fast_get_bid32(sign_x, exponent_y, C64);
        return (res, pfpsf);
    }
    pfpsf |= 1;
    res = 0x7c000000;
    return (res, pfpsf);
}
