// Auto-generated from mul64.go by go2rs. Do not edit.

use super::prelude::*;

pub fn bid64_mul(mut x: u64, mut y: u64, mut rndMode: i64) -> u64 {
    let mut P: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut C128: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut Q_high: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut Q_low: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut sign_x: u64 = 0;
    let mut sign_y: u64 = 0;
    let mut coefficient_x: u64 = 0;
    let mut coefficient_y: u64 = 0;
    let mut C64: u64 = 0;
    let mut remainder_h: u64 = 0;
    let mut res: u64 = 0;
    let mut valid_x: bool = false;
    let mut valid_y: bool = false;
    let mut extra_digits: i64 = 0;
    let mut exponent_x: i64 = 0;
    let mut exponent_y: i64 = 0;
    let mut bin_expon_cx: i64 = 0;
    let mut bin_expon_cy: i64 = 0;
    let mut bin_expon_product: i64 = 0;
    let mut rmode: i64 = 0;
    let mut digits_p: i64 = 0;
    let mut bp: i64 = 0;
    let mut amount: i64 = 0;
    let mut final_exponent: i64 = 0;
    let mut round_up: i64 = 0;
    (sign_x, exponent_x, coefficient_x, valid_x) = unpack_bid64(x);
    (sign_y, exponent_y, coefficient_y, valid_y) = unpack_bid64(y);
    if (!valid_x) {
        if ((y & 0x7e00000000000000) == 0x7e00000000000000) {
        }
        if ((x & 0x7c00000000000000) == 0x7c00000000000000) {
            if ((x & 0x7e00000000000000) == 0x7e00000000000000) {
            }
            return (coefficient_x & 0xfdffffffffffffff);
        }
        if ((x & 0x7800000000000000) == 0x7800000000000000) {
            if ((((y & 0x7800000000000000) != 0x7800000000000000)) && (coefficient_y == 0)) {
                return 0x7c00000000000000;
            }
            if ((y & 0x7c00000000000000) == 0x7c00000000000000) {
                return (coefficient_y & 0xfdffffffffffffff);
            }
            return ((((x ^ y) & 0x8000000000000000)) | 0x7800000000000000);
        }
        if ((y & 0x7800000000000000) != 0x7800000000000000) {
            if ((y & 0x6000000000000000) == 0x6000000000000000) {
                exponent_y = (((((go_checked_shr_u64(y, go_shift_count_u64((51) as u64))) as u32) & 0x3ff)) as i64);
            } else {
                exponent_y = (((((go_checked_shr_u64(y, go_shift_count_u64((53) as u64))) as u32) & 0x3ff)) as i64);
            }
            sign_y = (y & 0x8000000000000000);
            exponent_x = exponent_x.wrapping_add((exponent_y.wrapping_sub(0x18e)));
            if (exponent_x > 0x2ff) {
                exponent_x = 0x2ff;
            } else if (exponent_x < 0) {
                exponent_x = 0;
            }
            return ((sign_x ^ sign_y) | ((go_checked_shl_u64((exponent_x as u64), go_shift_count_u64((53) as u64)))));
        }
    }
    if (!valid_y) {
        if ((y & 0x7c00000000000000) == 0x7c00000000000000) {
            if ((y & 0x7e00000000000000) == 0x7e00000000000000) {
            }
            return (coefficient_y & 0xfdffffffffffffff);
        }
        if ((y & 0x7800000000000000) == 0x7800000000000000) {
            if (coefficient_x == 0) {
                return 0x7c00000000000000;
            }
            return ((((x ^ y) & 0x8000000000000000)) | 0x7800000000000000);
        }
        exponent_x = exponent_x.wrapping_add((exponent_y.wrapping_sub(0x18e)));
        if (exponent_x > 0x2ff) {
            exponent_x = 0x2ff;
        } else if (exponent_x < 0) {
            exponent_x = 0;
        }
        return ((sign_x ^ sign_y) | ((go_checked_shl_u64((exponent_x as u64), go_shift_count_u64((53) as u64)))));
    }
    let mut tempx = (coefficient_x as f64).to_bits();
    bin_expon_cx = ((go_checked_shr_u64((tempx & 0x7ff0000000000000), go_shift_count_u64((52) as u64))) as i64);
    let mut tempy = (coefficient_y as f64).to_bits();
    bin_expon_cy = ((go_checked_shr_u64((tempy & 0x7ff0000000000000), go_shift_count_u64((52) as u64))) as i64);
    bin_expon_product = (bin_expon_cx.wrapping_add(bin_expon_cy));
    if (bin_expon_product < (51 + (2 * 0x3ff))) {
        C64 = (coefficient_x.wrapping_mul(coefficient_y));
        res = get_bid64_small_mantissa((sign_x ^ sign_y), ((exponent_x.wrapping_add(exponent_y)).wrapping_sub(0x18e)), C64, rndMode);
        return res;
    }
    P = __mul_64x64_to_128(coefficient_x, coefficient_y);
    bin_expon_product = bin_expon_product.wrapping_sub(2 * 0x3ff);
    bp = __tight_bin_range_128(P, bin_expon_product);
    digits_p = (bid_estimate_decimal_digits[bp as usize] as i64);
    if (!__unsigned_compare_gt_128(bid_power10_table_128[digits_p as usize], P)) {
        digits_p = digits_p.wrapping_add(1);
    }
    extra_digits = (digits_p.wrapping_sub(16));
    final_exponent = (((exponent_x.wrapping_add(exponent_y)).wrapping_add(extra_digits)).wrapping_sub(0x18e));
    rmode = rndMode;
    if (((sign_x ^ sign_y) != 0) && (((rmode.wrapping_sub(1)) as u64) < 2)) {
        rmode = ((3 as i64).wrapping_sub(rmode));
    }
    round_up = 0;
    if ((final_exponent as u64) >= (3 * 256)) {
        if (final_exponent < 0) {
            if ((final_exponent.wrapping_add(16)) < 0) {
                res = (sign_x ^ sign_y);
                if (rmode == 2) {
                    res |= 1;
                }
                return res;
            }
            extra_digits = extra_digits.wrapping_sub(final_exponent);
            final_exponent = 0;
            if (extra_digits > 17) {
                (Q_high, Q_low) = __mul_128x128_full(P, bid_reciprocals10_128[16]);
                amount = (bid_recip_scale[16] as i64);
                P = __shr_128(Q_high, (amount as u64));
                let mut amount2 = ((64 as i64).wrapping_sub(amount));
                remainder_h = (go_checked_shr_u64(((!(0 as u64))), go_shift_count_u64((amount2 as u64) as u64)));
                remainder_h = (remainder_h & Q_high.w[0]);
                extra_digits = extra_digits.wrapping_sub(16);
                if ((remainder_h != 0) || (((Q_low.w[1] > bid_reciprocals10_128[16].w[1]) || (((Q_low.w[1] == bid_reciprocals10_128[16].w[1]) && (Q_low.w[0] >= bid_reciprocals10_128[16].w[0])))))) {
                    round_up = 1;
                    P.w[0] = (((go_checked_shl_u64(P.w[0], go_shift_count_u64((3) as u64)))).wrapping_add(((go_checked_shl_u64(P.w[0], go_shift_count_u64((1) as u64))))));
                    P.w[0] |= 1;
                    extra_digits = extra_digits.wrapping_add(1);
                }
            }
        } else {
            res = fast_get_bid64_check_of((sign_x ^ sign_y), final_exponent, 1000000000000000, rndMode);
            return res;
        }
    }
    if (extra_digits > 0) {
        P = __add_128_64(P, bid_round_const_table[rmode as usize][extra_digits as usize]);
        (Q_high, Q_low) = __mul_128x128_full(P, bid_reciprocals10_128[extra_digits as usize]);
        amount = (bid_recip_scale[extra_digits as usize] as i64);
        C128 = __shr_128(Q_high, (amount as u64));
        C64 = C128.w[0];
        if (rmode == 0) {
            if (((C64 & 1) != 0) && (round_up == 0)) {
                remainder_h = (go_checked_shl_u64(Q_high.w[0], go_shift_count_u64(((((64 as u64).wrapping_sub(amount as u64)))) as u64)));
                if ((remainder_h == 0) && (((Q_low.w[1] < bid_reciprocals10_128[extra_digits as usize].w[1]) || (((Q_low.w[1] == bid_reciprocals10_128[extra_digits as usize].w[1]) && (Q_low.w[0] < bid_reciprocals10_128[extra_digits as usize].w[0])))))) {
                    C64 = C64.wrapping_sub(1);
                }
            }
        }
        res = fast_get_bid64_check_of((sign_x ^ sign_y), final_exponent, C64, rndMode);
        return res;
    }
    C64 = P.w[0];
    res = get_bid64((sign_x ^ sign_y), ((exponent_x.wrapping_add(exponent_y)).wrapping_sub(0x18e)), C64, rndMode);
    return res;
}

pub fn bid64_mul_with_flags(mut x: u64, mut y: u64, mut rndMode: i64) -> (u64, u32) {
    let mut P: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut C128: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut Q_high: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut Q_low: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut sign_x: u64 = 0;
    let mut sign_y: u64 = 0;
    let mut coefficient_x: u64 = 0;
    let mut coefficient_y: u64 = 0;
    let mut C64: u64 = 0;
    let mut remainder_h: u64 = 0;
    let mut res: u64 = 0;
    let mut valid_x: bool = false;
    let mut valid_y: bool = false;
    let mut extra_digits: i64 = 0;
    let mut exponent_x: i64 = 0;
    let mut exponent_y: i64 = 0;
    let mut bin_expon_cx: i64 = 0;
    let mut bin_expon_cy: i64 = 0;
    let mut bin_expon_product: i64 = 0;
    let mut rmode: i64 = 0;
    let mut digits_p: i64 = 0;
    let mut bp: i64 = 0;
    let mut amount: i64 = 0;
    let mut final_exponent: i64 = 0;
    let mut round_up: i64 = 0;
    let mut pfpsf: u32 = 0;
    let mut uf_status: u32 = 0;
    (sign_x, exponent_x, coefficient_x, valid_x) = unpack_bid64(x);
    (sign_y, exponent_y, coefficient_y, valid_y) = unpack_bid64(y);
    if (!valid_x) {
        if ((y & 0x7e00000000000000) == 0x7e00000000000000) {
            pfpsf |= 1;
        }
        if ((x & 0x7c00000000000000) == 0x7c00000000000000) {
            if ((x & 0x7e00000000000000) == 0x7e00000000000000) {
                pfpsf |= 1;
            }
            return ((coefficient_x & 0xfdffffffffffffff), pfpsf);
        }
        if ((x & 0x7800000000000000) == 0x7800000000000000) {
            if ((((y & 0x7800000000000000) != 0x7800000000000000)) && (coefficient_y == 0)) {
                pfpsf |= 1;
                return (0x7c00000000000000, pfpsf);
            }
            if ((y & 0x7c00000000000000) == 0x7c00000000000000) {
                return ((coefficient_y & 0xfdffffffffffffff), pfpsf);
            }
            return (((((x ^ y) & 0x8000000000000000)) | 0x7800000000000000), pfpsf);
        }
        if ((y & 0x7800000000000000) != 0x7800000000000000) {
            if ((y & 0x6000000000000000) == 0x6000000000000000) {
                exponent_y = (((((go_checked_shr_u64(y, go_shift_count_u64((51) as u64))) as u32) & 0x3ff)) as i64);
            } else {
                exponent_y = (((((go_checked_shr_u64(y, go_shift_count_u64((53) as u64))) as u32) & 0x3ff)) as i64);
            }
            sign_y = (y & 0x8000000000000000);
            exponent_x = exponent_x.wrapping_add((exponent_y.wrapping_sub(0x18e)));
            if (exponent_x > 0x2ff) {
                exponent_x = 0x2ff;
            } else if (exponent_x < 0) {
                exponent_x = 0;
            }
            return (((sign_x ^ sign_y) | ((go_checked_shl_u64((exponent_x as u64), go_shift_count_u64((53) as u64))))), pfpsf);
        }
    }
    if (!valid_y) {
        if ((y & 0x7c00000000000000) == 0x7c00000000000000) {
            if ((y & 0x7e00000000000000) == 0x7e00000000000000) {
                pfpsf |= 1;
            }
            return ((coefficient_y & 0xfdffffffffffffff), pfpsf);
        }
        if ((y & 0x7800000000000000) == 0x7800000000000000) {
            if (coefficient_x == 0) {
                pfpsf |= 1;
                return (0x7c00000000000000, pfpsf);
            }
            return (((((x ^ y) & 0x8000000000000000)) | 0x7800000000000000), pfpsf);
        }
        exponent_x = exponent_x.wrapping_add((exponent_y.wrapping_sub(0x18e)));
        if (exponent_x > 0x2ff) {
            exponent_x = 0x2ff;
        } else if (exponent_x < 0) {
            exponent_x = 0;
        }
        return (((sign_x ^ sign_y) | ((go_checked_shl_u64((exponent_x as u64), go_shift_count_u64((53) as u64))))), pfpsf);
    }
    let mut tempx = (coefficient_x as f64).to_bits();
    bin_expon_cx = ((go_checked_shr_u64((tempx & 0x7ff0000000000000), go_shift_count_u64((52) as u64))) as i64);
    let mut tempy = (coefficient_y as f64).to_bits();
    bin_expon_cy = ((go_checked_shr_u64((tempy & 0x7ff0000000000000), go_shift_count_u64((52) as u64))) as i64);
    bin_expon_product = (bin_expon_cx.wrapping_add(bin_expon_cy));
    if (bin_expon_product < (51 + (2 * 0x3ff))) {
        C64 = (coefficient_x.wrapping_mul(coefficient_y));
        res = get_bid64_small_mantissa_flags((sign_x ^ sign_y), ((exponent_x.wrapping_add(exponent_y)).wrapping_sub(0x18e)), C64, rndMode, (&mut pfpsf));
        return (res, pfpsf);
    }
    P = __mul_64x64_to_128(coefficient_x, coefficient_y);
    bin_expon_product = bin_expon_product.wrapping_sub(2 * 0x3ff);
    bp = __tight_bin_range_128(P, bin_expon_product);
    digits_p = (bid_estimate_decimal_digits[bp as usize] as i64);
    if (!__unsigned_compare_gt_128(bid_power10_table_128[digits_p as usize], P)) {
        digits_p = digits_p.wrapping_add(1);
    }
    extra_digits = (digits_p.wrapping_sub(16));
    final_exponent = (((exponent_x.wrapping_add(exponent_y)).wrapping_add(extra_digits)).wrapping_sub(0x18e));
    rmode = rndMode;
    if (((sign_x ^ sign_y) != 0) && (((rmode.wrapping_sub(1)) as u64) < 2)) {
        rmode = ((3 as i64).wrapping_sub(rmode));
    }
    round_up = 0;
    if ((final_exponent as u64) >= (3 * 256)) {
        if (final_exponent < 0) {
            if ((final_exponent.wrapping_add(16)) < 0) {
                res = (sign_x ^ sign_y);
                pfpsf |= (16 | 32);
                if (rmode == 2) {
                    res |= 1;
                }
                return (res, pfpsf);
            }
            uf_status = 16;
            extra_digits = extra_digits.wrapping_sub(final_exponent);
            final_exponent = 0;
            if (extra_digits > 17) {
                (Q_high, Q_low) = __mul_128x128_full(P, bid_reciprocals10_128[16]);
                amount = (bid_recip_scale[16] as i64);
                P = __shr_128(Q_high, (amount as u64));
                let mut amount2 = ((64 as i64).wrapping_sub(amount));
                remainder_h = (go_checked_shr_u64(((!(0 as u64))), go_shift_count_u64((amount2 as u64) as u64)));
                remainder_h = (remainder_h & Q_high.w[0]);
                extra_digits = extra_digits.wrapping_sub(16);
                if ((remainder_h != 0) || (((Q_low.w[1] > bid_reciprocals10_128[16].w[1]) || (((Q_low.w[1] == bid_reciprocals10_128[16].w[1]) && (Q_low.w[0] >= bid_reciprocals10_128[16].w[0])))))) {
                    round_up = 1;
                    pfpsf |= (16 | 32);
                    P.w[0] = (((go_checked_shl_u64(P.w[0], go_shift_count_u64((3) as u64)))).wrapping_add(((go_checked_shl_u64(P.w[0], go_shift_count_u64((1) as u64))))));
                    P.w[0] |= 1;
                    extra_digits = extra_digits.wrapping_add(1);
                }
            }
        } else {
            let (mut res, mut flags) = fast_get_bid64_check_of_flags((sign_x ^ sign_y), final_exponent, 1000000000000000, rndMode);
            pfpsf |= flags;
            return (res, pfpsf);
        }
    }
    if (extra_digits > 0) {
        P = __add_128_64(P, bid_round_const_table[rmode as usize][extra_digits as usize]);
        (Q_high, Q_low) = __mul_128x128_full(P, bid_reciprocals10_128[extra_digits as usize]);
        amount = (bid_recip_scale[extra_digits as usize] as i64);
        C128 = __shr_128(Q_high, (amount as u64));
        C64 = C128.w[0];
        if (rmode == 0) {
            if (((C64 & 1) != 0) && (round_up == 0)) {
                remainder_h = (go_checked_shl_u64(Q_high.w[0], go_shift_count_u64(((((64 as u64).wrapping_sub(amount as u64)))) as u64)));
                if ((remainder_h == 0) && (((Q_low.w[1] < bid_reciprocals10_128[extra_digits as usize].w[1]) || (((Q_low.w[1] == bid_reciprocals10_128[extra_digits as usize].w[1]) && (Q_low.w[0] < bid_reciprocals10_128[extra_digits as usize].w[0])))))) {
                    C64 = C64.wrapping_sub(1);
                }
            }
        }
        let mut status = ((32 as u32) | uf_status);
        remainder_h = (go_checked_shl_u64(Q_high.w[0], go_shift_count_u64(((((64 as u64).wrapping_sub(amount as u64)))) as u64)));
        match rmode {
            0 | 4 => {
                if ((remainder_h == 0x8000000000000000) && (((Q_low.w[1] < bid_reciprocals10_128[extra_digits as usize].w[1]) || (((Q_low.w[1] == bid_reciprocals10_128[extra_digits as usize].w[1]) && (Q_low.w[0] < bid_reciprocals10_128[extra_digits as usize].w[0])))))) {
                    status = 0;
                }
            }
            1 | 3 => {
                if ((remainder_h == 0) && (((Q_low.w[1] < bid_reciprocals10_128[extra_digits as usize].w[1]) || (((Q_low.w[1] == bid_reciprocals10_128[extra_digits as usize].w[1]) && (Q_low.w[0] < bid_reciprocals10_128[extra_digits as usize].w[0])))))) {
                    status = 0;
                }
            }
            _ => {
                let mut CY: u64 = 0;
                let mut carry: u64 = 0;
                let (mut Stemp_w0, mut CY) = go_add64(Q_low.w[0], bid_reciprocals10_128[extra_digits as usize].w[0], 0);
                (_, carry) = go_add64(Q_low.w[1], bid_reciprocals10_128[extra_digits as usize].w[1], CY);
                _ = Stemp_w0;
                if ((((go_checked_shr_u64(remainder_h, go_shift_count_u64(((((64 as u64).wrapping_sub(amount as u64)))) as u64)))).wrapping_add(carry)) >= ((go_checked_shl_u64((1 as u64), go_shift_count_u64((amount as u64) as u64))))) {
                    status = 0;
                }
            }
        }
        pfpsf |= status;
        let (mut res, mut flags) = fast_get_bid64_check_of_flags((sign_x ^ sign_y), final_exponent, C64, rndMode);
        pfpsf |= flags;
        return (res, pfpsf);
    }
    C64 = P.w[0];
    res = get_bid64((sign_x ^ sign_y), ((exponent_x.wrapping_add(exponent_y)).wrapping_sub(0x18e)), C64, rndMode);
    return (res, pfpsf);
}
