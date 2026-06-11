// Auto-generated from add64.go by go2rs. Do not edit.

use super::prelude::*;

pub fn bid64_sub(mut x: u64, mut y: u64, mut rndMode: i64) -> u64 {
    let (mut result, _) = bid64_sub_with_flags(x, y, rndMode);
    return result;
}

pub fn bid64_sub_with_flags(mut x: u64, mut y: u64, mut rndMode: i64) -> (u64, u32) {
    if ((y & 0x7c00000000000000) != 0x7c00000000000000) {
        y ^= 0x8000000000000000;
    }
    return bid64_add_with_flags(x, y, rndMode);
}

pub fn bid64_add(mut x: u64, mut y: u64, mut rndMode: i64) -> u64 {
    let (mut result, _) = bid64_add_with_flags(x, y, rndMode);
    return result;
}

pub fn bid64_add_with_flags(mut x: u64, mut y: u64, mut rndMode: i64) -> (u64, u32) {
    let mut CA: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut CT: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut CT_new: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut sign_x: u64 = 0;
    let mut sign_y: u64 = 0;
    let mut coefficient_x: u64 = 0;
    let mut coefficient_y: u64 = 0;
    let mut C64_new: u64 = 0;
    let mut valid_x: bool = false;
    let mut valid_y: bool = false;
    let mut res: u64 = 0;
    let mut pfpsf: u32 = 0;
    let mut sign_a: u64 = 0;
    let mut sign_b: u64 = 0;
    let mut coefficient_a: u64 = 0;
    let mut coefficient_b: u64 = 0;
    let mut sign_s: u64 = 0;
    let mut sign_ab: u64 = 0;
    let mut rem_a: u64 = 0;
    let mut saved_ca: u64 = 0;
    let mut saved_cb: u64 = 0;
    let mut C0_64: u64 = 0;
    let mut C64: u64 = 0;
    let mut remainder_h: u64 = 0;
    let mut T1: u64 = 0;
    let mut carry: u64 = 0;
    let mut tmp: u64 = 0;
    let mut exponent_x: i64 = 0;
    let mut exponent_y: i64 = 0;
    let mut exponent_a: i64 = 0;
    let mut exponent_b: i64 = 0;
    let mut diff_dec_expon: i64 = 0;
    let mut bin_expon_ca: i64 = 0;
    let mut extra_digits: i64 = 0;
    let mut amount: i64 = 0;
    let mut scale_k: i64 = 0;
    let mut scale_ca: i64 = 0;
    let mut rmode: i64 = 0;
    let mut status: u32 = 0;
    (sign_x, exponent_x, coefficient_x, valid_x) = unpack_bid64(x);
    (sign_y, exponent_y, coefficient_y, valid_y) = unpack_bid64(y);
    if (!valid_x) {
        if ((x & 0x7c00000000000000) == 0x7c00000000000000) {
            if ((((x & 0x7e00000000000000) == 0x7e00000000000000)) || (((y & 0x7e00000000000000) == 0x7e00000000000000))) {
                pfpsf |= 1;
            }
            res = (coefficient_x & 0xfdffffffffffffff);
            return (res, pfpsf);
        }
        if ((x & 0x7800000000000000) == 0x7800000000000000) {
            if ((y & 0x7c00000000000000) == 0x7800000000000000) {
                if (sign_x == (y & 0x8000000000000000)) {
                    res = coefficient_x;
                    return (res, pfpsf);
                }
                pfpsf |= 1;
                res = 0x7c00000000000000;
                return (res, pfpsf);
            }
            if ((y & 0x7c00000000000000) == 0x7c00000000000000) {
                res = (coefficient_y & 0xfdffffffffffffff);
                if ((y & 0x7e00000000000000) == 0x7e00000000000000) {
                    pfpsf |= 1;
                }
                return (res, pfpsf);
            }
            res = coefficient_x;
            return (res, pfpsf);
        }
        if ((((y & 0x7800000000000000) != 0x7800000000000000)) && (coefficient_y != 0)) {
            if (exponent_y <= exponent_x) {
                res = y;
                return (res, pfpsf);
            }
        }
    }
    if (!valid_y) {
        if ((y & 0x7800000000000000) == 0x7800000000000000) {
            if ((y & 0x7e00000000000000) == 0x7e00000000000000) {
                pfpsf |= 1;
            }
            res = (coefficient_y & 0xfdffffffffffffff);
            return (res, pfpsf);
        }
        if (coefficient_x == 0) {
            if (exponent_x <= exponent_y) {
                res = (go_checked_shl_u64((exponent_x as u64), go_shift_count_u64((53) as u64)));
            } else {
                res = (go_checked_shl_u64((exponent_y as u64), go_shift_count_u64((53) as u64)));
            }
            if (sign_x == sign_y) {
                res |= sign_x;
            }
            if ((rndMode == 1) && (sign_x != sign_y)) {
                res |= 0x8000000000000000;
            }
            return (res, pfpsf);
        } else if (exponent_y >= exponent_x) {
            res = x;
            return (res, pfpsf);
        }
    }
    if (exponent_x < exponent_y) {
        sign_a = sign_y;
        exponent_a = exponent_y;
        coefficient_a = coefficient_y;
        sign_b = sign_x;
        exponent_b = exponent_x;
        coefficient_b = coefficient_x;
    } else {
        sign_a = sign_x;
        exponent_a = exponent_x;
        coefficient_a = coefficient_x;
        sign_b = sign_y;
        exponent_b = exponent_y;
        coefficient_b = coefficient_y;
    }
    diff_dec_expon = (exponent_a.wrapping_sub(exponent_b));
    let mut tempx = f64::from_bits((coefficient_a as f64).to_bits());
    bin_expon_ca = (((go_checked_shr_u64((((tempx).to_bits() & 0x7ff0000000000000)), go_shift_count_u64((52) as u64))) as i64).wrapping_sub(0x3ff));
    if (diff_dec_expon > 16) {
        scale_ca = (bid_estimate_decimal_digits[bin_expon_ca as usize] as i64);
        if (coefficient_a >= bid_power10_table_128[scale_ca as usize].w[0]) {
            scale_ca = scale_ca.wrapping_add(1);
        }
        scale_k = ((16 as i64).wrapping_sub(scale_ca));
        coefficient_a = coefficient_a.wrapping_mul(bid_power10_table_128[scale_k as usize].w[0]);
        diff_dec_expon = diff_dec_expon.wrapping_sub(scale_k);
        exponent_a = exponent_a.wrapping_sub(scale_k);
        tempx = (coefficient_a as f64);
        bin_expon_ca = (((go_checked_shr_u64((((tempx).to_bits() & 0x7ff0000000000000)), go_shift_count_u64((52) as u64))) as i64).wrapping_sub(0x3ff));
        if (diff_dec_expon > 16) {
            if (coefficient_b != 0) {
                pfpsf |= 32;
            }
            if (((rndMode & 3) != 0) && (coefficient_b != 0)) {
                match rndMode {
                    1 => {
                        if (sign_b != 0) {
                            coefficient_a = coefficient_a.wrapping_sub(((((go_checked_shr_i64((sign_a as i64), go_shift_count_u64((63) as u64)))) | 1) as u64));
                            if (coefficient_a < 1000000000000000) {
                                exponent_a = exponent_a.wrapping_sub(1);
                                coefficient_a = 9999999999999999;
                            } else if (coefficient_a >= 10000000000000000) {
                                exponent_a = exponent_a.wrapping_add(1);
                                coefficient_a = 1000000000000000;
                            }
                        }
                    }
                    2 => {
                        if (sign_b == 0) {
                            coefficient_a = coefficient_a.wrapping_add(((((go_checked_shr_i64((sign_a as i64), go_shift_count_u64((63) as u64)))) | 1) as u64));
                            if (coefficient_a < 1000000000000000) {
                                exponent_a = exponent_a.wrapping_sub(1);
                                coefficient_a = 9999999999999999;
                            } else if (coefficient_a >= 10000000000000000) {
                                exponent_a = exponent_a.wrapping_add(1);
                                coefficient_a = 1000000000000000;
                            }
                        }
                    }
                    _ => {
                        if (sign_a != sign_b) {
                            coefficient_a = coefficient_a.wrapping_sub(1);
                            if (coefficient_a < 1000000000000000) {
                                exponent_a = exponent_a.wrapping_sub(1);
                                coefficient_a = 9999999999999999;
                            }
                        }
                    }
                }
            } else {
                if ((((coefficient_a == 1000000000000000) && (diff_dec_expon == (16 + 1))) && ((sign_a ^ sign_b) != 0)) && (coefficient_b > 5000000000000000)) {
                    coefficient_a = 9999999999999999;
                    exponent_a = exponent_a.wrapping_sub(1);
                }
            }
            let (mut res, mut flags) = fast_get_bid64_check_of_flags(sign_a, exponent_a, coefficient_a, rndMode);
            pfpsf |= flags;
            return (res, pfpsf);
        }
    }
    if ((bin_expon_ca.wrapping_add(bid_estimate_bin_expon[diff_dec_expon as usize] as i64)) < 60) {
        coefficient_a = coefficient_a.wrapping_mul(bid_power10_table_128[diff_dec_expon as usize].w[0]);
        sign_b = ((go_checked_shr_i64((sign_b as i64), go_shift_count_u64((63) as u64))) as u64);
        coefficient_b = (((coefficient_b.wrapping_add(sign_b))) ^ sign_b);
        sign_a = ((go_checked_shr_i64((sign_a as i64), go_shift_count_u64((63) as u64))) as u64);
        coefficient_a = (((coefficient_a.wrapping_add(sign_a))) ^ sign_a);
        coefficient_a = coefficient_a.wrapping_add(coefficient_b);
        sign_s = ((go_checked_shr_i64((coefficient_a as i64), go_shift_count_u64((63) as u64))) as u64);
        coefficient_a = (((coefficient_a.wrapping_add(sign_s))) ^ sign_s);
        sign_s &= 0x8000000000000000;
        if (coefficient_a < bid_power10_table_128[16].w[0]) {
            if (((rndMode == 1) && (coefficient_a == 0)) && (sign_a != sign_b)) {
                sign_s = 0x8000000000000000;
            }
            res = very_fast_get_bid64(sign_s, exponent_b, coefficient_a);
            return (res, pfpsf);
        }
        if (coefficient_a < bid_power10_table_128[17].w[0]) {
            extra_digits = 1;
        } else if (coefficient_a < bid_power10_table_128[18].w[0]) {
            extra_digits = 2;
        } else {
            extra_digits = 3;
        }
        rmode = rndMode;
        if ((sign_s != 0) && (((rmode.wrapping_sub(1)) as u64) < 2)) {
            rmode = ((3 as i64).wrapping_sub(rmode));
        }
        coefficient_a = coefficient_a.wrapping_add(bid_round_const_table[rmode as usize][extra_digits as usize]);
        CT = __mul_64x64_to_128(coefficient_a, bid_reciprocals10_64[extra_digits as usize]);
        amount = (bid_short_recip_scale[extra_digits as usize] as i64);
        C64 = (go_checked_shr_u64(CT.w[1], go_shift_count_u64((amount as u64) as u64)));
    } else {
        sign_s = sign_a;
        rmode = rndMode;
        if ((sign_s != 0) && (((rmode.wrapping_sub(1)) as u64) < 2)) {
            rmode = ((3 as i64).wrapping_sub(rmode));
        }
        scale_ca = (bid_estimate_decimal_digits[bin_expon_ca as usize] as i64);
        sign_ab = (sign_a ^ sign_b);
        sign_ab = ((go_checked_shr_i64((sign_ab as i64), go_shift_count_u64((63) as u64))) as u64);
        T1 = bid_power10_table_128[((16 as i64).wrapping_sub(diff_dec_expon)) as usize].w[0];
        if (coefficient_a >= bid_power10_table_128[scale_ca as usize].w[0]) {
            scale_ca = scale_ca.wrapping_add(1);
        }
        scale_k = ((16 as i64).wrapping_sub(scale_ca));
        saved_ca = (coefficient_a.wrapping_sub(T1));
        coefficient_a = (((saved_ca as i64).wrapping_mul(bid_power10_table_128[scale_k as usize].w[0] as i64)) as u64);
        extra_digits = (diff_dec_expon.wrapping_sub(scale_k));
        saved_cb = (((coefficient_b.wrapping_add(sign_ab))) ^ sign_ab);
        coefficient_b = ((saved_cb.wrapping_add(10000000000000000)).wrapping_add(bid_round_const_table[rmode as usize][extra_digits as usize]));
        CT = __mul_64x64_to_128(coefficient_b, bid_reciprocals10_64[extra_digits as usize]);
        amount = (bid_short_recip_scale[extra_digits as usize] as i64);
        C0_64 = (go_checked_shr_u64(CT.w[1], go_shift_count_u64((amount as u64) as u64)));
        C64 = (C0_64.wrapping_add(((coefficient_a as i64) as u64)));
        if ((((C64.wrapping_sub(1000000000000000)).wrapping_sub(1)) as u64) > (9000000000000000 - 2)) {
            if (C64 >= 10000000000000000) {
                if (scale_k == 0) {
                    saved_ca = (saved_ca.wrapping_add(T1));
                    CA = __mul_64x64_to_128(saved_ca, 0x3333333333333334);
                    coefficient_a = (go_checked_shr_u64(CA.w[1], go_shift_count_u64((1) as u64)));
                    rem_a = ((saved_ca.wrapping_sub(((go_checked_shl_u64(coefficient_a, go_shift_count_u64((3) as u64)))))).wrapping_sub(((go_checked_shl_u64(coefficient_a, go_shift_count_u64((1) as u64))))));
                    coefficient_a = (coefficient_a.wrapping_sub(T1));
                    saved_cb = saved_cb.wrapping_add((rem_a.wrapping_mul(bid_power10_table_128[diff_dec_expon as usize].w[0])));
                } else {
                    coefficient_a = (((((saved_ca.wrapping_sub(T1)).wrapping_sub(((go_checked_shl_u64(T1, go_shift_count_u64((3) as u64)))))) as i64).wrapping_mul((bid_power10_table_128[(scale_k.wrapping_sub(1)) as usize].w[0] as i64))) as u64);
                }
                extra_digits = extra_digits.wrapping_add(1);
                coefficient_b = ((saved_cb.wrapping_add(100000000000000000)).wrapping_add(bid_round_const_table[rmode as usize][extra_digits as usize]));
                CT = __mul_64x64_to_128(coefficient_b, bid_reciprocals10_64[extra_digits as usize]);
                amount = (bid_short_recip_scale[extra_digits as usize] as i64);
                C0_64 = (go_checked_shr_u64(CT.w[1], go_shift_count_u64((amount as u64) as u64)));
                C64 = (C0_64.wrapping_add(((coefficient_a as i64) as u64)));
            } else if (C64 <= 1000000000000000) {
                coefficient_a = (((saved_ca as i64).wrapping_mul((bid_power10_table_128[(scale_k.wrapping_add(1)) as usize].w[0] as i64))) as u64);
                exponent_b = exponent_b.wrapping_sub(1);
                coefficient_b = (((((go_checked_shl_u64(saved_cb, go_shift_count_u64((3) as u64)))).wrapping_add(((go_checked_shl_u64(saved_cb, go_shift_count_u64((1) as u64)))))).wrapping_add(100000000000000000)).wrapping_add(bid_round_const_table[rmode as usize][extra_digits as usize]));
                CT_new = __mul_64x64_to_128(coefficient_b, bid_reciprocals10_64[extra_digits as usize]);
                amount = (bid_short_recip_scale[extra_digits as usize] as i64);
                C0_64 = (go_checked_shr_u64(CT_new.w[1], go_shift_count_u64((amount as u64) as u64)));
                C64_new = (C0_64.wrapping_add(((coefficient_a as i64) as u64)));
                if (C64_new < 10000000000000000) {
                    C64 = C64_new;
                    CT = CT_new;
                } else {
                    exponent_b = exponent_b.wrapping_add(1);
                }
            }
        }
    }
    if (rmode == 0) {
        if ((C64 & 1) != 0) {
            remainder_h = (go_checked_shl_u64(CT.w[1], go_shift_count_u64(((((64 as u64).wrapping_sub(amount as u64)))) as u64)));
            if ((remainder_h == 0) && (CT.w[0] < bid_reciprocals10_64[extra_digits as usize])) {
                C64 = C64.wrapping_sub(1);
            }
        }
    }
    status = 32;
    remainder_h = (go_checked_shl_u64(CT.w[1], go_shift_count_u64(((((64 as u64).wrapping_sub(amount as u64)))) as u64)));
    match rmode {
        0 | 4 => {
            if ((remainder_h == 0x8000000000000000) && (CT.w[0] < bid_reciprocals10_64[extra_digits as usize])) {
                status = 0;
            }
        }
        1 | 3 => {
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
    let (mut res, mut flags) = fast_get_bid64_check_of_flags(sign_s, (exponent_b.wrapping_add(extra_digits)), C64, rndMode);
    pfpsf |= flags;
    return (res, pfpsf);
}

pub(crate) fn fast_get_bid64_check_of_flags(mut sgn: u64, mut expon: i64, mut coeff: u64, mut rmode: i64) -> (u64, u32) {
    let mut r: u64 = 0;
    let mut flags: u32 = 0;
    if ((expon as u64) >= ((3 * 256) - 1)) {
        if ((expon == ((3 * 256) - 1)) && (coeff == 10000000000000000)) {
            expon = (3 * 256);
            coeff = 1000000000000000;
        }
        if ((expon as u64) >= (3 * 256)) {
            while ((coeff < 1000000000000000) && (expon >= (3 * 256))) {
                expon = expon.wrapping_sub(1);
                coeff = (((go_checked_shl_u64(coeff, go_shift_count_u64((3) as u64)))).wrapping_add(((go_checked_shl_u64(coeff, go_shift_count_u64((1) as u64))))));
            }
            if (expon > 0x2ff) {
                flags |= (8 | 32);
                r = (sgn | 0x7800000000000000);
                match rmode {
                    1 => {
                        if (sgn == 0) {
                            r = 0x77fb86f26fc0ffff;
                        }
                    }
                    3 => {
                        r = (sgn | 0x77fb86f26fc0ffff);
                    }
                    2 => {
                        if (sgn != 0) {
                            r = 0xf7fb86f26fc0ffff;
                        }
                    }
                    _ => {}
                }
                return (r, flags);
            }
        }
    }
    let mut mask: u64 = ((1 as u64) << 53);
    if (coeff < mask) {
        r = (expon as u64);
        r = go_checked_shl_u64(r, go_shift_count_u64((53) as u64));
        r |= (coeff | sgn);
        return (r, flags);
    }
    if (coeff == 10000000000000000) {
        r = ((expon.wrapping_add(1)) as u64);
        r = go_checked_shl_u64(r, go_shift_count_u64((53) as u64));
        r |= (1000000000000000 | sgn);
        return (r, flags);
    }
    r = (expon as u64);
    r = go_checked_shl_u64(r, go_shift_count_u64((51) as u64));
    r |= (sgn | 0x6000000000000000);
    mask = (((go_checked_shr_u64(mask, go_shift_count_u64((2) as u64)))).wrapping_sub(1));
    coeff &= mask;
    r |= coeff;
    return (r, flags);
}
