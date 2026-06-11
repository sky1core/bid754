// Auto-generated from quantize64.go by go2rs. Do not edit.

use super::prelude::*;

pub fn bid64_quantize(mut x: u64, mut y: u64, mut rndMode: i64) -> (u64, u32) {
    let mut CT: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut sign_x: u64 = 0;
    let mut sign_y: u64 = 0;
    let mut coefficient_x: u64 = 0;
    let mut coefficient_y: u64 = 0;
    let mut remainder_h: u64 = 0;
    let mut C64: u64 = 0;
    let mut valid_x: bool = false;
    let mut valid_y: bool = false;
    let mut tmp: u64 = 0;
    let mut carry: u64 = 0;
    let mut res: u64 = 0;
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
    (sign_x, exponent_x, coefficient_x, valid_x) = unpack_bid64(x);
    (sign_y, exponent_y, coefficient_y, valid_y) = unpack_bid64(y);
    if (!valid_y) {
        if ((x & 0x7e00000000000000) == 0x7e00000000000000) {
            pfpsf |= 1;
        }
        if ((((go_checked_shl_u64(coefficient_x, go_shift_count_u64((1) as u64)))) == 0xf000000000000000) && (((go_checked_shl_u64(coefficient_y, go_shift_count_u64((1) as u64)))) == 0xf000000000000000)) {
            res = coefficient_x;
            return (res, pfpsf);
        }
        if ((y & 0x7800000000000000) == 0x7800000000000000) {
            if ((((y & 0x7e00000000000000) == 0x7e00000000000000)) || (((((y & 0x7c00000000000000) == 0x7800000000000000)) && (((x & 0x7c00000000000000) < 0x7800000000000000))))) {
                pfpsf |= 1;
            }
            if ((y & 0x7c00000000000000) != 0x7c00000000000000) {
                coefficient_y = 0;
            }
            if ((x & 0x7c00000000000000) != 0x7c00000000000000) {
                res = (0x7c00000000000000 | (coefficient_y & 0xfdffffffffffffff));
                if ((((y & 0x7c00000000000000) != 0x7c00000000000000)) && (((x & 0x7c00000000000000) == 0x7800000000000000))) {
                    res = x;
                }
                return (res, pfpsf);
            }
        }
    }
    _ = sign_y;
    if (!valid_x) {
        if ((x & 0x7800000000000000) == 0x7800000000000000) {
            if ((((x & 0x7e00000000000000) == 0x7e00000000000000)) || (((x & 0x7c00000000000000) == 0x7800000000000000))) {
                pfpsf |= 1;
            }
            if ((x & 0x7c00000000000000) != 0x7c00000000000000) {
                coefficient_x = 0;
            }
            res = (0x7c00000000000000 | (coefficient_x & 0xfdffffffffffffff));
            return (res, pfpsf);
        }
        res = very_fast_get_bid64_small_mantissa(sign_x, exponent_y, 0);
        return (res, pfpsf);
    }
    let mut tempx = ((coefficient_x as f32) as f32).to_bits();
    bin_expon_cx = (((((go_checked_shr_u32(tempx, go_shift_count_u64((23) as u64)))) & 0xff) as i64).wrapping_sub(0x7f));
    digits_x = (bid_estimate_decimal_digits[bin_expon_cx as usize] as i64);
    if (coefficient_x >= bid_power10_table_128[digits_x as usize].w[0]) {
        digits_x = digits_x.wrapping_add(1);
    }
    expon_diff = (exponent_x.wrapping_sub(exponent_y));
    total_digits = (digits_x.wrapping_add(expon_diff));
    if (((total_digits.wrapping_add(1)) as u32) <= 17) {
        if (expon_diff >= 0) {
            coefficient_x = coefficient_x.wrapping_mul(bid_power10_table_128[expon_diff as usize].w[0]);
            res = very_fast_get_bid64(sign_x, exponent_y, coefficient_x);
            return (res, pfpsf);
        }
        extra_digits = (expon_diff.wrapping_neg());
        rmode = rndMode;
        if ((sign_x != 0) && (((rmode.wrapping_sub(1)) as u64) < 2)) {
            rmode = ((3 as i64).wrapping_sub(rmode));
        }
        coefficient_x = coefficient_x.wrapping_add(bid_round_const_table[rmode as usize][extra_digits as usize]);
        CT = __mul_64x64_to_128(coefficient_x, bid_reciprocals10_64[extra_digits as usize]);
        amount = (bid_short_recip_scale[extra_digits as usize] as i64);
        C64 = (go_checked_shr_u64(CT.w[1], go_shift_count_u64((amount as u64) as u64)));
        if (rndMode == 0) {
            if ((C64 & 1) != 0) {
                amount2 = ((64 as i64).wrapping_sub(amount));
                remainder_h = 0;
                remainder_h = remainder_h.wrapping_sub(1);
                remainder_h = go_checked_shr_u64(remainder_h, go_shift_count_u64((amount2 as u64) as u64));
                remainder_h = (remainder_h & CT.w[1]);
                if ((remainder_h == 0) && (CT.w[0] < bid_reciprocals10_64[extra_digits as usize])) {
                    C64 = C64.wrapping_sub(1);
                }
            }
        }
        status = 32;
        remainder_h = (go_checked_shl_u64(CT.w[1], go_shift_count_u64(((((64 as u64).wrapping_sub(amount as u64)))) as u64)));
        match rmode {
            0 => {
            }
            4 => {
                if ((remainder_h == 0x8000000000000000) && (CT.w[0] < bid_reciprocals10_64[extra_digits as usize])) {
                    status = 0;
                }
            }
            1 => {
            }
            3 => {
                if ((remainder_h == 0) && (CT.w[0] < bid_reciprocals10_64[extra_digits as usize])) {
                    status = 0;
                }
            }
            _ => {
                (tmp, carry) = __add_carry_out(CT.w[0], bid_reciprocals10_64[extra_digits as usize]);
                _ = tmp;
                if ((((go_checked_shr_u64(remainder_h, go_shift_count_u64(((((64 as i64).wrapping_sub(amount)) as u64)) as u64)))).wrapping_add(carry)) >= ((go_checked_shl_u64((1 as u64), go_shift_count_u64((amount as u64) as u64))))) {
                    status = 0;
                }
            }
        }
        pfpsf |= status;
        res = very_fast_get_bid64_small_mantissa(sign_x, exponent_y, C64);
        return (res, pfpsf);
    }
    if (total_digits < 0) {
        pfpsf |= 32;
        C64 = 0;
        rmode = rndMode;
        if ((sign_x != 0) && (((rmode.wrapping_sub(1)) as u64) < 2)) {
            rmode = ((3 as i64).wrapping_sub(rmode));
        }
        if (rmode == 2) {
            C64 = 1;
        }
        res = very_fast_get_bid64_small_mantissa(sign_x, exponent_y, C64);
        return (res, pfpsf);
    }
    pfpsf |= 1;
    res = 0x7c00000000000000;
    return (res, pfpsf);
}
