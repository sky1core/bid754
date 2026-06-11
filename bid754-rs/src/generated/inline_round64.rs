// Auto-generated from inline_round64.go by go2rs. Do not edit.

use super::prelude::*;

pub(crate) fn __low_64(mut q: BID_UINT128) -> u64 {
    return q.w[0];
}

pub(crate) fn __mul_64x64_to_64(mut cx: u64, mut cy: u64) -> u64 {
    return (cx.wrapping_mul(cy));
}

pub(crate) fn __mul_64x128_short(mut a: u64, mut b: BID_UINT128) -> BID_UINT128 {
    let mut ql: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut ALBH_L: u64 = 0;
    ALBH_L = __mul_64x64_to_64(a, b.w[1]);
    ql = __mul_64x64_to_128(a, b.w[0]);
    ql.w[1] = ql.w[1].wrapping_add(ALBH_L);
    return ql;
}

pub(crate) fn __scale128_10(mut tmp: BID_UINT128) -> BID_UINT128 {
    let mut tmp2: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut tmp8: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    tmp2.w[1] = (((go_checked_shl_u64(tmp.w[1], go_shift_count_u64((1) as u64)))) | ((go_checked_shr_u64(tmp.w[0], go_shift_count_u64((63) as u64)))));
    tmp2.w[0] = (go_checked_shl_u64(tmp.w[0], go_shift_count_u64((1) as u64)));
    tmp8.w[1] = (((go_checked_shl_u64(tmp.w[1], go_shift_count_u64((3) as u64)))) | ((go_checked_shr_u64(tmp.w[0], go_shift_count_u64((61) as u64)))));
    tmp8.w[0] = (go_checked_shl_u64(tmp.w[0], go_shift_count_u64((3) as u64)));
    return __add_128_128(tmp2, tmp8);
}

pub(crate) fn __bid_simple_round64_sticky(mut sign: u64, mut exponent: i64, mut P: BID_UINT128, mut extra_digits: i64, mut rounding_mode: i64, fpsc: &mut u32) -> u64 {
    let mut Q_high: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut Q_low: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut C128: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut C64: u64 = 0;
    let mut amount: i64 = 0;
    let mut rmode: i64 = 0;
    rmode = rounding_mode;
    if ((sign != 0) && (((rmode.wrapping_sub(1)) as u64) < 2)) {
        rmode = ((3 as i64).wrapping_sub(rmode));
    }
    P = __add_128_64(P, bid_round_const_table[rmode as usize][extra_digits as usize]);
    (Q_high, Q_low) = __mul_128x128_full(P, bid_reciprocals10_128[extra_digits as usize]);
    _ = Q_low;
    amount = (bid_recip_scale[extra_digits as usize] as i64);
    C128 = __shr_128(Q_high, (amount as u64));
    C64 = __low_64(C128);
    (*fpsc) |= 32;
    return get_bid64_with_flags(sign, exponent, C64, rounding_mode, fpsc);
}

pub(crate) fn __bid_full_round64(mut sign: u64, mut exponent: i64, mut P: BID_UINT128, mut extra_digits: i64, mut rounding_mode: i64, fpsc: &mut u32) -> u64 {
    let mut Q_high: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut Q_low: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut C128: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut Stemp: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut remainder_h: u64 = 0;
    let mut C64: u64 = 0;
    let mut carry: u64 = 0;
    let mut CY: u64 = 0;
    let mut amount: i64 = 0;
    let mut amount2: i64 = 0;
    let mut rmode: i64 = 0;
    let mut status: u32 = 0;
    if (exponent < 0) {
        if ((exponent >= (-16)) && (((extra_digits.wrapping_add(exponent)) < 0))) {
            extra_digits = (exponent.wrapping_neg());
            status = 16;
        }
    }
    if (extra_digits > 0) {
        exponent = exponent.wrapping_add(extra_digits);
        rmode = rounding_mode;
        if ((sign != 0) && (((rmode.wrapping_sub(1)) as u64) < 2)) {
            rmode = ((3 as i64).wrapping_sub(rmode));
        }
        P = __add_128_128(P, bid_round_const_table_128[rmode as usize][extra_digits as usize]);
        (Q_high, Q_low) = __mul_128x128_full(P, bid_reciprocals10_128[extra_digits as usize]);
        amount = (bid_recip_scale[extra_digits as usize] as i64);
        C128 = __shr_128_long(Q_high, (amount as u64));
        C64 = __low_64(C128);
        if (rmode == 0) {
            if ((C64 & 1) != 0) {
                amount2 = ((64 as i64).wrapping_sub(amount));
                remainder_h = 0;
                remainder_h = remainder_h.wrapping_sub(1);
                remainder_h = go_checked_shr_u64(remainder_h, go_shift_count_u64((amount2 as u64) as u64));
                remainder_h = (remainder_h & Q_high.w[0]);
                if ((remainder_h == 0) && (((Q_low.w[1] < bid_reciprocals10_128[extra_digits as usize].w[1]) || (((Q_low.w[1] == bid_reciprocals10_128[extra_digits as usize].w[1]) && (Q_low.w[0] < bid_reciprocals10_128[extra_digits as usize].w[0])))))) {
                    C64 = C64.wrapping_sub(1);
                }
            }
        }
        status |= 32;
        remainder_h = (go_checked_shl_u64(Q_high.w[0], go_shift_count_u64(((((64 as i64).wrapping_sub(amount)) as u64)) as u64)));
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
                (Stemp.w[0], CY) = __add_carry_out(Q_low.w[0], bid_reciprocals10_128[extra_digits as usize].w[0]);
                (Stemp.w[1], carry) = __add_carry_in_out(Q_low.w[1], bid_reciprocals10_128[extra_digits as usize].w[1], CY);
                _ = Stemp;
                if ((((go_checked_shr_u64(remainder_h, go_shift_count_u64(((((64 as i64).wrapping_sub(amount)) as u64)) as u64)))).wrapping_add(carry)) >= ((go_checked_shl_u64((1 as u64), go_shift_count_u64((amount as u64) as u64))))) {
                    status = 0;
                }
            }
        }
        (*fpsc) |= status;
    } else {
        C64 = P.w[0];
        if (C64 == 0) {
            sign = 0;
            if (rounding_mode == 1) {
                sign = 0x8000000000000000;
            }
        }
    }
    return get_bid64_with_flags(sign, exponent, C64, rounding_mode, fpsc);
}

pub(crate) fn __bid_full_round64_remainder(mut sign: u64, mut exponent: i64, mut P: BID_UINT128, mut extra_digits: i64, mut remainder_P: u64, mut rounding_mode: i64, fpsc: &mut u32, mut uf_status: u32) -> u64 {
    let mut Q_high: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut Q_low: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut C128: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut Stemp: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut remainder_h: u64 = 0;
    let mut C64: u64 = 0;
    let mut carry: u64 = 0;
    let mut CY: u64 = 0;
    let mut amount: i64 = 0;
    let mut amount2: i64 = 0;
    let mut rmode: i64 = 0;
    let mut status = uf_status;
    rmode = rounding_mode;
    if ((sign != 0) && (((rmode.wrapping_sub(1)) as u64) < 2)) {
        rmode = ((3 as i64).wrapping_sub(rmode));
    }
    if ((rmode == 2) && (remainder_P != 0)) {
        P.w[0] = P.w[0].wrapping_add(1);
        if (P.w[0] == 0) {
            P.w[1] = P.w[1].wrapping_add(1);
        }
    }
    if (extra_digits != 0) {
        P = __add_128_64(P, bid_round_const_table[rmode as usize][extra_digits as usize]);
        (Q_high, Q_low) = __mul_128x128_full(P, bid_reciprocals10_128[extra_digits as usize]);
        amount = (bid_recip_scale[extra_digits as usize] as i64);
        C128 = __shr_128(Q_high, (amount as u64));
        C64 = __low_64(C128);
        if (rmode == 0) {
            if ((remainder_P == 0) && ((C64 & 1) != 0)) {
                amount2 = ((64 as i64).wrapping_sub(amount));
                remainder_h = 0;
                remainder_h = remainder_h.wrapping_sub(1);
                remainder_h = go_checked_shr_u64(remainder_h, go_shift_count_u64((amount2 as u64) as u64));
                remainder_h = (remainder_h & Q_high.w[0]);
                if ((remainder_h == 0) && (((Q_low.w[1] < bid_reciprocals10_128[extra_digits as usize].w[1]) || (((Q_low.w[1] == bid_reciprocals10_128[extra_digits as usize].w[1]) && (Q_low.w[0] < bid_reciprocals10_128[extra_digits as usize].w[0])))))) {
                    C64 = C64.wrapping_sub(1);
                }
            }
        }
        status |= 32;
        if (remainder_P == 0) {
            remainder_h = (go_checked_shl_u64(Q_high.w[0], go_shift_count_u64(((((64 as i64).wrapping_sub(amount)) as u64)) as u64)));
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
                    (Stemp.w[0], CY) = __add_carry_out(Q_low.w[0], bid_reciprocals10_128[extra_digits as usize].w[0]);
                    (Stemp.w[1], carry) = __add_carry_in_out(Q_low.w[1], bid_reciprocals10_128[extra_digits as usize].w[1], CY);
                    _ = Stemp;
                    if ((((go_checked_shr_u64(remainder_h, go_shift_count_u64(((((64 as i64).wrapping_sub(amount)) as u64)) as u64)))).wrapping_add(carry)) >= ((go_checked_shl_u64((1 as u64), go_shift_count_u64((amount as u64) as u64))))) {
                        status = 0;
                    }
                }
            }
        }
        (*fpsc) |= status;
    } else {
        C64 = P.w[0];
        if (remainder_P != 0) {
            (*fpsc) |= (uf_status | 32);
        }
    }
    return get_bid64_with_flags(sign, (exponent.wrapping_add(extra_digits)), C64, rounding_mode, fpsc);
}

pub(crate) fn __truncate(mut P: BID_UINT128, mut extra_digits: i64) -> u64 {
    let mut Q_high: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut Q_low: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut C128: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut C64: u64 = 0;
    let mut amount: i64 = 0;
    (Q_high, Q_low) = __mul_128x128_full(P, bid_reciprocals10_128[extra_digits as usize]);
    _ = Q_low;
    amount = (bid_recip_scale[extra_digits as usize] as i64);
    C128 = __shr_128(Q_high, (amount as u64));
    C64 = __low_64(C128);
    return C64;
}

pub(crate) fn __get_dec_digits64(mut X: BID_UINT128) -> i64 {
    let mut digits_x: i64 = 0;
    let mut bin_expon_cx: i64 = 0;
    if (X.w[1] == 0) {
        if (X.w[0] == 0) {
            return 0;
        }
        let mut tempx = (X.w[0] as f64).to_bits();
        bin_expon_cx = (((go_checked_shr_u64((tempx & 0x7ff0000000000000), go_shift_count_u64((52) as u64))) as i64).wrapping_sub(0x3ff));
        digits_x = (bid_estimate_decimal_digits[bin_expon_cx as usize] as i64);
        if (X.w[0] >= bid_power10_table_128[digits_x as usize].w[0]) {
            digits_x = digits_x.wrapping_add(1);
        }
        return digits_x;
    }
    let mut tempx = (X.w[1] as f64).to_bits();
    bin_expon_cx = (((go_checked_shr_u64((tempx & 0x7ff0000000000000), go_shift_count_u64((52) as u64))) as i64).wrapping_sub(0x3ff));
    digits_x = (bid_estimate_decimal_digits[(bin_expon_cx.wrapping_add(64)) as usize] as i64);
    if __unsigned_compare_ge_128(X, bid_power10_table_128[digits_x as usize]) {
        digits_x = digits_x.wrapping_add(1);
    }
    return digits_x;
}

pub fn bid_normalize(mut sign_z: u64, mut exponent_z: i64, mut coefficient_z: u64, mut round_dir: u64, mut round_flag: i64, mut rounding_mode: i64, fpsc: &mut u32) -> u64 {
    let mut D: i64 = 0;
    let mut digits_z: i64 = 0;
    let mut bin_expon: i64 = 0;
    let mut scale: i64 = 0;
    let mut rmode: i64 = 0;
    rmode = rounding_mode;
    if ((sign_z != 0) && (((rmode.wrapping_sub(1)) as u64) < 2)) {
        rmode = ((3 as i64).wrapping_sub(rmode));
    }
    let mut tempx = (coefficient_z as f64).to_bits();
    bin_expon = (((go_checked_shr_u64((tempx & 0x7ff0000000000000), go_shift_count_u64((52) as u64))) as i64).wrapping_sub(0x3ff));
    digits_z = (bid_estimate_decimal_digits[bin_expon as usize] as i64);
    if (coefficient_z >= bid_power10_table_128[digits_z as usize].w[0]) {
        digits_z = digits_z.wrapping_add(1);
    }
    scale = ((16 as i64).wrapping_sub(digits_z));
    exponent_z = exponent_z.wrapping_sub(scale);
    if (exponent_z < 0) {
        scale = scale.wrapping_add(exponent_z);
        exponent_z = 0;
    }
    coefficient_z = coefficient_z.wrapping_mul(bid_power10_table_128[scale as usize].w[0]);
    if (round_flag != 0) {
        (*fpsc) |= 32;
        if (coefficient_z < 1000000000000000) {
            (*fpsc) |= 16;
        } else if ((((coefficient_z == 1000000000000000) && (exponent_z == 0)) && ((((round_dir ^ sign_z) as i64) < 0))) && (round_flag != 0)) {
            (*fpsc) |= 16;
        }
    }
    if ((round_flag != 0) && ((rmode & 3) != 0)) {
        D = ((round_dir ^ sign_z) as i64);
        if (rmode == 2) {
            if (D >= 0) {
                coefficient_z = coefficient_z.wrapping_add(1);
            }
        } else {
            if (D < 0) {
                coefficient_z = coefficient_z.wrapping_sub(1);
            }
            if ((coefficient_z < 1000000000000000) && (exponent_z != 0)) {
                coefficient_z = 9999999999999999;
                exponent_z = exponent_z.wrapping_sub(1);
            }
        }
    }
    return get_bid64_with_flags(sign_z, exponent_z, coefficient_z, rounding_mode, fpsc);
}

pub(crate) fn add_zero64(mut exponent_y: i64, mut sign_z: u64, mut exponent_z: i64, mut coefficient_z: u64, prounding_mode: &mut i64, fpsc: &mut u32) -> u64 {
    let mut bin_expon: i64 = 0;
    let mut scale_k: i64 = 0;
    let mut scale_cz: i64 = 0;
    let mut diff_expon: i64 = 0;
    diff_expon = (exponent_z.wrapping_sub(exponent_y));
    let mut tempx = (coefficient_z as f64).to_bits();
    bin_expon = (((go_checked_shr_u64((tempx & 0x7ff0000000000000), go_shift_count_u64((52) as u64))) as i64).wrapping_sub(0x3ff));
    scale_cz = (bid_estimate_decimal_digits[bin_expon as usize] as i64);
    if (coefficient_z >= bid_power10_table_128[scale_cz as usize].w[0]) {
        scale_cz = scale_cz.wrapping_add(1);
    }
    scale_k = ((16 as i64).wrapping_sub(scale_cz));
    if (diff_expon < scale_k) {
        scale_k = diff_expon;
    }
    coefficient_z = coefficient_z.wrapping_mul(bid_power10_table_128[scale_k as usize].w[0]);
    return get_bid64_with_flags(sign_z, (exponent_z.wrapping_sub(scale_k)), coefficient_z, (*prounding_mode), fpsc);
}
