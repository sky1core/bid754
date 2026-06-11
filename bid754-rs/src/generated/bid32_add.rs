// Auto-generated from bid32_add.go by go2rs. Do not edit.

use super::prelude::*;

pub(crate) fn bid32_add_pure(mut x: u32, mut y: u32, mut rndMode: i64) -> u32 {
    let mut Tmp: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut S: i64 = 0;
    let mut sign_ab: i64 = 0;
    let mut SU: u64 = 0;
    let mut CB: u64 = 0;
    let mut P: u64 = 0;
    let mut Q: u64 = 0;
    let mut R: u64 = 0;
    let mut sign_x: u32 = 0;
    let mut sign_y: u32 = 0;
    let mut coefficient_x: u32 = 0;
    let mut coefficient_y: u32 = 0;
    let mut res: u32 = 0;
    let mut sign_a: u32 = 0;
    let mut sign_b: u32 = 0;
    let mut coefficient_a: u32 = 0;
    let mut coefficient_b: u32 = 0;
    let mut valid_x: bool = false;
    let mut valid_y: bool = false;
    let mut exponent_x: i64 = 0;
    let mut exponent_y: i64 = 0;
    let mut bin_expon: i64 = 0;
    let mut amount: i64 = 0;
    let mut n_digits: i64 = 0;
    let mut extra_digits: i64 = 0;
    let mut rmode: i64 = 0;
    let mut exponent_a: i64 = 0;
    let mut exponent_b: i64 = 0;
    let mut scale_ca: i64 = 0;
    let mut diff_dec_expon: i64 = 0;
    let mut d2: i64 = 0;
    (sign_x, exponent_x, coefficient_x, valid_x) = unpack_bid32_add(x);
    (sign_y, exponent_y, coefficient_y, valid_y) = unpack_bid32_add(y);
    if (!valid_x) {
        if ((x & 0x7c000000) == 0x7c000000) {
            res = (coefficient_x & 0xfdffffff);
            return res;
        }
        if ((x & 0x78000000) == 0x78000000) {
            if ((y & 0x7c000000) == 0x78000000) {
                if (sign_x == (y & 0x80000000)) {
                    res = coefficient_x;
                    return res;
                }
                res = 0x7c000000;
                return res;
            }
            if ((y & 0x7c000000) == 0x7c000000) {
                res = (coefficient_y & 0xfdffffff);
                return res;
            }
            res = coefficient_x;
            return res;
        }
        if ((((y & 0x78000000) != 0x78000000)) && (coefficient_y != 0)) {
            if (exponent_y <= exponent_x) {
                res = y;
                return res;
            }
        }
    }
    if (!valid_y) {
        if ((y & 0x78000000) == 0x78000000) {
            res = (coefficient_y & 0xfdffffff);
            return res;
        }
        if (coefficient_x == 0) {
            if (exponent_x <= exponent_y) {
                res = (go_checked_shl_u32((exponent_x as u32), go_shift_count_u64((23) as u64)));
            } else {
                res = (go_checked_shl_u32((exponent_y as u32), go_shift_count_u64((23) as u64)));
            }
            if (sign_x == sign_y) {
                res |= sign_x;
            }
            if ((rndMode == 1) && (sign_x != sign_y)) {
                res |= 0x80000000;
            }
            return res;
        } else if (exponent_y >= exponent_x) {
            res = x;
            return res;
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
    if (diff_dec_expon > 7) {
        let mut tempx = (coefficient_a as f64);
        bin_expon = (((go_checked_shr_u64((((tempx).to_bits() & 0x7ff0000000000000)), go_shift_count_u64((52) as u64))) as i64).wrapping_sub(0x3ff));
        scale_ca = (bid_estimate_decimal_digits[bin_expon as usize] as i64);
        d2 = ((16 as i64).wrapping_sub(scale_ca));
        if (diff_dec_expon > d2) {
            diff_dec_expon = d2;
            exponent_b = (exponent_a.wrapping_sub(diff_dec_expon));
        }
    }
    sign_ab = (go_checked_shl_i64(((sign_a ^ sign_b) as i64), go_shift_count_u64((32) as u64)));
    sign_ab = (go_checked_shr_i64(sign_ab, go_shift_count_u64((63) as u64)));
    CB = ((((coefficient_b as i64).wrapping_add(sign_ab)) as u64) ^ (sign_ab as u64));
    SU = ((coefficient_a as u64).wrapping_mul(bid_power10_table_128[diff_dec_expon as usize].w[0]));
    S = ((SU as i64).wrapping_add(CB as i64));
    if (S < 0) {
        sign_a ^= 0x80000000;
        S = (S.wrapping_neg());
    }
    P = (S as u64);
    if (P == 0) {
        sign_a = 0;
        if (rndMode == 1) {
            sign_a = 0x80000000;
        }
        if (coefficient_a == 0) {
            sign_a = sign_x;
        }
        n_digits = 0;
    } else {
        let mut tempx = (P as f64);
        bin_expon = (((go_checked_shr_u64((((tempx).to_bits() & 0x7ff0000000000000)), go_shift_count_u64((52) as u64))) as i64).wrapping_sub(0x3ff));
        n_digits = (bid_estimate_decimal_digits[bin_expon as usize] as i64);
        if (P >= bid_power10_table_128[n_digits as usize].w[0]) {
            n_digits = n_digits.wrapping_add(1);
        }
    }
    if (n_digits <= 7) {
        res = get_bid32(sign_a, exponent_b, P, rndMode);
        return res;
    }
    extra_digits = (n_digits.wrapping_sub(7));
    rmode = rndMode;
    if ((sign_a != 0) && (((rmode.wrapping_sub(1)) as u64) < 2)) {
        rmode = ((3 as i64).wrapping_sub(rmode));
    }
    P = P.wrapping_add(bid_round_const_table[rmode as usize][extra_digits as usize]);
    Tmp = __mul_64x64_to_128(P, bid_reciprocals10_64[extra_digits as usize]);
    amount = (bid_short_recip_scale[extra_digits as usize] as i64);
    Q = (go_checked_shr_u64(Tmp.w[1], go_shift_count_u64((amount as u64) as u64)));
    R = (P.wrapping_sub((Q.wrapping_mul(bid_power10_table_128[extra_digits as usize].w[0]))));
    if (rmode == 0) {
        if (R == 0) {
            Q &= 0xfffffffe;
        }
    }
    res = get_bid32(sign_a, (exponent_b.wrapping_add(extra_digits)), Q, rndMode);
    return res;
}

pub(crate) fn unpack_bid32_add(mut x: u32) -> (u32, i64, u32, bool) {
    let mut sign: u32 = 0;
    let mut exponent: i64 = 0;
    let mut coefficient: u32 = 0;
    let mut valid: bool = false;
    sign = (x & 0x80000000);
    if ((x & 0x60000000) == 0x60000000) {
        if ((x & 0x78000000) == 0x78000000) {
            coefficient = (x & 0xfe0fffff);
            if ((x & 0x000fffff) >= 1000000) {
                coefficient = (x & 0xfe000000);
            }
            if ((x & 0x7c000000) == 0x78000000) {
                coefficient = (x & 0xf8000000);
            }
            return (sign, 0, coefficient, false);
        }
        coefficient = ((x & 0x1fffff) | 0x800000);
        if (coefficient >= 10000000) {
            coefficient = 0;
        }
        exponent = ((((go_checked_shr_u32(x, go_shift_count_u64((21) as u64)))) & 255) as i64);
        return (sign, exponent, coefficient, (coefficient != 0));
    }
    exponent = ((((go_checked_shr_u32(x, go_shift_count_u64((23) as u64)))) & 255) as i64);
    coefficient = (x & 0x7fffff);
    return (sign, exponent, coefficient, (coefficient != 0));
}

pub(crate) fn bid32_sub_pure(mut x: u32, mut y: u32, mut rndMode: i64) -> u32 {
    if ((y & 0x7c000000) != 0x7c000000) {
        y ^= 0x80000000;
    }
    return bid32_add_pure(x, y, rndMode);
}
