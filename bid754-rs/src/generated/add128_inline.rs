// Auto-generated from add128_inline.go by go2rs. Do not edit.

use super::prelude::*;

pub(crate) fn bid_get_add128(mut sign_x: u64, mut exponent_x: i64, mut coefficient_x: u64, mut sign_y: u64, mut final_exponent_y: i64, mut CY: BID_UINT128, mut extra_digits: i64, mut rounding_mode: i64, fpsc: &mut u32) -> u64 {
    let mut CY_L: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut CX: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut FS: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut F: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut CT: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut ST: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut T2: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut CYh: u64 = 0;
    let mut CY0L: u64 = 0;
    let mut T: u64 = 0;
    let mut S: u64 = 0;
    let mut coefficient_y: u64 = 0;
    let mut remainder_y: u64 = 0;
    let mut D: i64 = 0;
    let mut diff_dec_expon: i64 = 0;
    let mut extra_digits2: i64 = 0;
    let mut exponent_y: i64 = 0;
    let mut extra_dx: i64 = 0;
    let mut diff_dec2: i64 = 0;
    let mut bin_expon_cx: i64 = 0;
    let mut digits_x: i64 = 0;
    let mut rmode: i64 = 0;
    let mut status: u32 = 0;
    _ = status;
    exponent_y = (final_exponent_y.wrapping_sub(extra_digits));
    if (exponent_x > exponent_y) {
        let mut tempx = (coefficient_x as f64).to_bits();
        bin_expon_cx = (((go_checked_shr_u64((tempx & 0x7ff0000000000000), go_shift_count_u64((52) as u64))) as i64).wrapping_sub(0x3ff));
        digits_x = (bid_estimate_decimal_digits[bin_expon_cx as usize] as i64);
        if (coefficient_x >= bid_power10_table_128[digits_x as usize].w[0]) {
            digits_x = digits_x.wrapping_add(1);
        }
        extra_dx = ((16 as i64).wrapping_sub(digits_x));
        coefficient_x = coefficient_x.wrapping_mul(bid_power10_table_128[extra_dx as usize].w[0]);
        if (((sign_x ^ sign_y) != 0) && (coefficient_x == 1000000000000000)) {
            extra_dx = extra_dx.wrapping_add(1);
            coefficient_x = 10000000000000000;
        }
        exponent_x = exponent_x.wrapping_sub(extra_dx);
        if (exponent_x > exponent_y) {
            diff_dec_expon = (exponent_x.wrapping_sub(exponent_y));
            if (exponent_x <= (final_exponent_y.wrapping_add(1))) {
                CX = __mul_64x64_to_128(coefficient_x, bid_power10_table_128[diff_dec_expon as usize].w[0]);
                if (sign_x == sign_y) {
                    CT = __add_128_128(CY, CX);
                    if (exponent_x > final_exponent_y) {
                        extra_digits = extra_digits.wrapping_add(1);
                    }
                    if __unsigned_compare_ge_128(CT, bid_power10_table_128[((16 as i64).wrapping_add(extra_digits)) as usize]) {
                        extra_digits = extra_digits.wrapping_add(1);
                    }
                } else {
                    CT = __sub_128_128(CY, CX);
                    if ((CT.w[1] as i64) < 0) {
                        CT.w[0] = ((0 as u64).wrapping_sub(CT.w[0]));
                        CT.w[1] = ((0 as u64).wrapping_sub(CT.w[1]));
                        if (CT.w[0] != 0) {
                            CT.w[1] = CT.w[1].wrapping_sub(1);
                        }
                        sign_y = sign_x;
                    } else if ((CT.w[1] | CT.w[0]) == 0) {
                        if (rounding_mode != 1) {
                            sign_y = 0;
                        } else {
                            sign_y = 0x8000000000000000;
                        }
                    }
                    if ((exponent_x.wrapping_add(1)) >= final_exponent_y) {
                        extra_digits = (__get_dec_digits64(CT).wrapping_sub(16));
                        if (extra_digits <= 0) {
                            if ((CT.w[0] == 0) && (rounding_mode == 1)) {
                                sign_y = 0x8000000000000000;
                            }
                            return get_bid64_with_flags(sign_y, exponent_y, CT.w[0], rounding_mode, fpsc);
                        }
                    } else if __unsigned_compare_gt_128(bid_power10_table_128[((15 as i64).wrapping_add(extra_digits)) as usize], CT) {
                        extra_digits = extra_digits.wrapping_sub(1);
                    }
                }
                return __bid_full_round64(sign_y, exponent_y, CT, extra_digits, rounding_mode, fpsc);
            }
            diff_dec2 = (exponent_x.wrapping_sub(final_exponent_y));
            if (diff_dec2 >= 17) {
                if ((rounding_mode & 3) != 0) {
                    match rounding_mode {
                        2 => {
                            if (sign_y == 0) {
                                D = ((sign_x ^ sign_y) as i64);
                                D = go_checked_shr_i64(D, go_shift_count_u64((63) as u64));
                                D = ((D.wrapping_add(D)).wrapping_add(1));
                                coefficient_x = coefficient_x.wrapping_add(D as u64);
                            }
                        }
                        1 => {
                            if (sign_y != 0) {
                                D = ((sign_x ^ sign_y) as i64);
                                D = go_checked_shr_i64(D, go_shift_count_u64((63) as u64));
                                D = ((D.wrapping_add(D)).wrapping_add(1));
                                coefficient_x = coefficient_x.wrapping_add(D as u64);
                            }
                        }
                        3 => {
                            if (sign_y != sign_x) {
                                D = (0 - 1);
                                coefficient_x = coefficient_x.wrapping_add(D as u64);
                            }
                        }
                        _ => {
                        }
                    }
                    if (coefficient_x < 1000000000000000) {
                        coefficient_x = coefficient_x.wrapping_sub(D as u64);
                        coefficient_x = (((D as u64).wrapping_add(((go_checked_shl_u64(coefficient_x, go_shift_count_u64((1) as u64)))))).wrapping_add(((go_checked_shl_u64(coefficient_x, go_shift_count_u64((3) as u64))))));
                        exponent_x = exponent_x.wrapping_sub(1);
                    }
                }
                if ((CY.w[1] | CY.w[0]) != 0) {
                    (*fpsc) |= 32;
                }
                return get_bid64_with_flags(sign_x, exponent_x, coefficient_x, rounding_mode, fpsc);
            }
            CYh = __truncate(CY, extra_digits);
            T = bid_power10_table_128[extra_digits as usize].w[0];
            CY0L = __mul_64x64_to_64(CYh, T);
            remainder_y = (CY.w[0].wrapping_sub(CY0L));
            CX = __mul_64x64_to_128(coefficient_x, bid_power10_table_128[diff_dec2 as usize].w[0]);
            if (sign_x == sign_y) {
                CT = __add_128_64(CX, CYh);
                if __unsigned_compare_ge_128(CT, bid_power10_table_128[((16 as i64).wrapping_add(diff_dec2)) as usize]) {
                    diff_dec2 = diff_dec2.wrapping_add(1);
                }
            } else {
                if (remainder_y != 0) {
                    CYh = CYh.wrapping_add(1);
                }
                CT = __sub_128_64(CX, CYh);
                if __unsigned_compare_gt_128(bid_power10_table_128[((15 as i64).wrapping_add(diff_dec2)) as usize], CT) {
                    diff_dec2 = diff_dec2.wrapping_sub(1);
                }
            }
            return __bid_full_round64_remainder(sign_x, final_exponent_y, CT, diff_dec2, remainder_y, rounding_mode, fpsc, 0);
        }
    }
    diff_dec_expon = (exponent_y.wrapping_sub(exponent_x));
    if (diff_dec_expon > 16) {
        rmode = rounding_mode;
        if ((sign_x ^ sign_y) != 0) {
            if (CY.w[0] == 0) {
                CY.w[1] = CY.w[1].wrapping_sub(1);
            }
            CY.w[0] = CY.w[0].wrapping_sub(1);
            if __unsigned_compare_gt_128(bid_power10_table_128[((15 as i64).wrapping_add(extra_digits)) as usize], CY) {
                if ((rmode & 3) != 0) {
                    extra_digits = extra_digits.wrapping_sub(1);
                    final_exponent_y = final_exponent_y.wrapping_sub(1);
                } else {
                    CY.w[0] = 1000000000000000;
                    CY.w[1] = 0;
                    extra_digits = 0;
                }
            }
        }
        CY = __scale128_10(CY);
        extra_digits = extra_digits.wrapping_add(1);
        CY.w[0] |= 1;
        return __bid_simple_round64_sticky(sign_y, final_exponent_y, CY, extra_digits, rmode, fpsc);
    }
    sign_x ^= sign_y;
    sign_x = ((go_checked_shr_i64((sign_x as i64), go_shift_count_u64((63) as u64))) as u64);
    CX.w[0] = (((coefficient_x.wrapping_add(sign_x))) ^ sign_x);
    CX.w[1] = sign_x;
    diff_dec2 = (final_exponent_y.wrapping_sub(exponent_x));
    if (diff_dec2 <= 17) {
        S = bid_power10_table_128[diff_dec_expon as usize].w[0];
        CY_L = __mul_64x128_short(S, CY);
        ST = __add_128_128(CY_L, CX);
        extra_digits2 = (__get_dec_digits64(ST).wrapping_sub(16));
        return __bid_full_round64(sign_y, exponent_x, ST, extra_digits2, rounding_mode, fpsc);
    }
    CYh = __truncate(CY, extra_digits);
    T = bid_power10_table_128[extra_digits as usize].w[0];
    CY0L = __mul_64x64_to_64(CYh, T);
    coefficient_y = (CY.w[0].wrapping_sub(CY0L));
    rmode = rounding_mode;
    if ((sign_y != 0) && (((rmode.wrapping_sub(1)) as u64) < 2)) {
        rmode = ((3 as i64).wrapping_sub(rmode));
    }
    if ((rmode & 3) == 0) {
        coefficient_y = coefficient_y.wrapping_add(bid_round_const_table[rmode as usize][extra_digits as usize]);
    }
    S = bid_power10_table_128[diff_dec_expon as usize].w[0];
    F = __mul_64x64_to_128(coefficient_y, S);
    FS = __add_128_128(F, CX);
    if (rmode == 0) {
        T2 = bid_power10_table_128[(diff_dec_expon.wrapping_add(extra_digits)) as usize];
        if (__unsigned_compare_gt_128(FS, T2) || ((((CYh & 1) != 0) && __test_equal_128(FS, T2)))) {
            CYh = CYh.wrapping_add(1);
            FS = __sub_128_128(FS, T2);
        }
    }
    if (rmode == 4) {
        T2 = bid_power10_table_128[(diff_dec_expon.wrapping_add(extra_digits)) as usize];
        if __unsigned_compare_ge_128(FS, T2) {
            CYh = CYh.wrapping_add(1);
            FS = __sub_128_128(FS, T2);
        }
    }
    match rmode {
        1 | 3 => {
            if ((FS.w[1] as i64) < 0) {
                CYh = CYh.wrapping_sub(1);
                if (CYh < 1000000000000000) {
                    CYh = 9999999999999999;
                    final_exponent_y = final_exponent_y.wrapping_sub(1);
                }
            } else {
                T2 = bid_power10_table_128[(diff_dec_expon.wrapping_add(extra_digits)) as usize];
                if __unsigned_compare_ge_128(FS, T2) {
                    CYh = CYh.wrapping_add(1);
                    FS = __sub_128_128(FS, T2);
                }
            }
        }
        2 => {
            if !((FS.w[1] as i64) < 0) {
                T2 = bid_power10_table_128[(diff_dec_expon.wrapping_add(extra_digits)) as usize];
                if __unsigned_compare_gt_128(FS, T2) {
                    CYh = CYh.wrapping_add(2);
                    FS = __sub_128_128(FS, T2);
                } else if ((FS.w[1] == T2.w[1]) && (FS.w[0] == T2.w[0])) {
                    CYh = CYh.wrapping_add(1);
                    FS.w[1] = 0;
                    FS.w[0] = 0;
                } else if ((FS.w[1] | FS.w[0]) != 0) {
                    CYh = CYh.wrapping_add(1);
                }
            }
        }
        _ => {
        }
    }
    status = 32;
    if ((rmode & 3) == 0) {
        if (((FS.w[1] == bid_round_const_table_128[0][(diff_dec_expon.wrapping_add(extra_digits)) as usize].w[1])) && ((FS.w[0] == bid_round_const_table_128[0][(diff_dec_expon.wrapping_add(extra_digits)) as usize].w[0]))) {
            status = 0;
        }
    } else if ((FS.w[1] == 0) && (FS.w[0] == 0)) {
        status = 0;
    }
    (*fpsc) |= status;
    return get_bid64_with_flags(sign_y, final_exponent_y, CYh, rounding_mode, fpsc);
}
