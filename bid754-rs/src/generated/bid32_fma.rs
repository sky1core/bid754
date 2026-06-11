// Auto-generated from bid32_fma.go by go2rs. Do not edit.

use super::prelude::*;

pub(crate) fn add_zero32(mut exponent_y: i64, mut sign_z: u32, mut exponent_z: i64, mut coefficient_z: u32, prounding_mode: &mut i64, fpsc: &mut u32) -> u32 {
    let mut bin_expon: i64 = 0;
    let mut scale_k: i64 = 0;
    let mut scale_cz: i64 = 0;
    let mut diff_expon: i64 = 0;
    diff_expon = (exponent_z.wrapping_sub(exponent_y));
    let mut tempx = (coefficient_z as f64).to_bits();
    bin_expon = ((((go_checked_shr_u64((tempx & 0x7ff0000000000000), go_shift_count_u64((52) as u64)))) as i64).wrapping_sub(0x3ff));
    scale_cz = (bid_estimate_decimal_digits[bin_expon as usize] as i64);
    if ((coefficient_z as u64) >= bid_power10_table_128[scale_cz as usize].w[0]) {
        scale_cz = scale_cz.wrapping_add(1);
    }
    scale_k = ((7 as i64).wrapping_sub(scale_cz));
    if (diff_expon < scale_k) {
        scale_k = diff_expon;
    }
    coefficient_z = coefficient_z.wrapping_mul(bid_power10_table_128[scale_k as usize].w[0] as u32);
    return get_bid32(sign_z, (exponent_z.wrapping_sub(scale_k)), (coefficient_z as u64), (*prounding_mode));
}

pub fn bid32_fma(mut x: u32, mut y: u32, mut z: u32, mut rnd_mode: i64) -> (u32, u32) {
    let mut P: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut Tmp: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut CB: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut Q_high: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut Q_low: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut Stemp: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut C128: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut P0: u64 = 0;
    let mut C64: u64 = 0;
    let mut remainder_h: u64 = 0;
    let mut rem_l: u64 = 0;
    let mut carry: u64 = 0;
    let mut CY: u64 = 0;
    let mut coefficient_a: u64 = 0;
    let mut coefficient_b: u64 = 0;
    let mut sign_ab: u64 = 0;
    let mut sign_x: u32 = 0;
    let mut sign_y: u32 = 0;
    let mut coefficient_x: u32 = 0;
    let mut coefficient_y: u32 = 0;
    let mut sign_z: u32 = 0;
    let mut coefficient_z: u32 = 0;
    let mut R: u32 = 0;
    let mut sign_a: u32 = 0;
    let mut sign_b: u32 = 0;
    let mut res: u32 = 0;
    let mut extra_digits: i64 = 0;
    let mut exponent_x: i64 = 0;
    let mut exponent_y: i64 = 0;
    let mut exponent_z: i64 = 0;
    let mut bin_expon: i64 = 0;
    let mut rmode: i64 = 0;
    let mut inexact: i64 = 0;
    let mut n_digits: i64 = 0;
    let mut amount: i64 = 0;
    let mut exponent_a: i64 = 0;
    let mut exponent_b: i64 = 0;
    let mut diff_dec_expon: i64 = 0;
    let mut d2: i64 = 0;
    let mut scale_ca: i64 = 0;
    let mut status: u32 = 0;
    let mut pfpsf: u32 = 0;
    let (mut sign_x, mut exponent_x, mut coefficient_x, mut valid_x) = unpack_bid32(x);
    let (mut sign_y, mut exponent_y, mut coefficient_y, mut valid_y) = unpack_bid32(y);
    let (mut sign_z, mut exponent_z, mut coefficient_z, mut valid_z) = unpack_bid32(z);
    if (coefficient_x == 0) {
        valid_x = false;
    }
    if (coefficient_y == 0) {
        valid_y = false;
    }
    if (coefficient_z == 0) {
        valid_z = false;
    }
    if (((!valid_x) || (!valid_y)) || (!valid_z)) {
        if ((y & 0x7c000000) == 0x7c000000) {
            if (((((x & 0x7e000000) == 0x7e000000)) || (((y & 0x7e000000) == 0x7e000000))) || (((z & 0x7e000000) == 0x7e000000))) {
                pfpsf |= 1;
            }
            res = (coefficient_y & 0xfdffffff);
            return (res, pfpsf);
        }
        if ((z & 0x7c000000) == 0x7c000000) {
            if ((((x & 0x7e000000) == 0x7e000000)) || (((z & 0x7e000000) == 0x7e000000))) {
                pfpsf |= 1;
            }
            res = (coefficient_z & 0xfdffffff);
            return (res, pfpsf);
        }
        if ((x & 0x7c000000) == 0x7c000000) {
            if ((x & 0x7e000000) == 0x7e000000) {
                pfpsf |= 1;
            }
            res = (coefficient_x & 0xfdffffff);
            return (res, pfpsf);
        }
        if (!valid_x) {
            if ((x & 0x78000000) == 0x78000000) {
                if (coefficient_y == 0) {
                    if ((z & 0x7e000000) != 0x7c000000) {
                        pfpsf |= 1;
                    }
                    return (0x7c000000, pfpsf);
                }
                if ((((z & 0x7c000000) == 0x78000000)) && ((((((x ^ y) ^ z)) & 0x80000000)) != 0)) {
                    pfpsf |= 1;
                    return (0x7c000000, pfpsf);
                }
                return (((((x ^ y) & 0x80000000)) | 0x78000000), pfpsf);
            }
            if ((((y & 0x78000000) != 0x78000000)) && (((z & 0x78000000) != 0x78000000))) {
                if (coefficient_z != 0) {
                    exponent_y = ((exponent_x.wrapping_sub(101)).wrapping_add(exponent_y));
                    sign_z = (z & 0x80000000);
                    if (exponent_y >= exponent_z) {
                        return (z, pfpsf);
                    }
                    res = add_zero32(exponent_y, sign_z, exponent_z, coefficient_z, (&mut rnd_mode), (&mut pfpsf));
                    return (res, pfpsf);
                }
            }
        }
        if (!valid_y) {
            if ((y & 0x78000000) == 0x78000000) {
                if (coefficient_x == 0) {
                    pfpsf |= 1;
                    return (0x7c000000, pfpsf);
                }
                if ((((z & 0x7c000000) == 0x78000000)) && ((((((x ^ y) ^ z)) & 0x80000000)) != 0)) {
                    pfpsf |= 1;
                    return (0x7c000000, pfpsf);
                }
                return (((((x ^ y) & 0x80000000)) | 0x78000000), pfpsf);
            }
            if ((z & 0x78000000) != 0x78000000) {
                if (coefficient_z != 0) {
                    exponent_y = exponent_y.wrapping_add((exponent_x.wrapping_sub(101)));
                    sign_z = (z & 0x80000000);
                    if (exponent_y >= exponent_z) {
                        return (z, pfpsf);
                    }
                    res = add_zero32(exponent_y, sign_z, exponent_z, coefficient_z, (&mut rnd_mode), (&mut pfpsf));
                    return (res, pfpsf);
                }
            }
        }
        if (!valid_z) {
            if ((z & 0x78000000) == 0x78000000) {
                res = (coefficient_z & 0xfdffffff);
                return (res, pfpsf);
            }
            if ((coefficient_x == 0) || (coefficient_y == 0)) {
                exponent_x = exponent_x.wrapping_add((exponent_y.wrapping_sub(101)));
                if (exponent_x > 191) {
                    exponent_x = 191;
                } else if (exponent_x < 0) {
                    exponent_x = 0;
                }
                if (exponent_x <= exponent_z) {
                    res = (go_checked_shl_u32((exponent_x as u32), go_shift_count_u64((23) as u64)));
                } else {
                    res = (go_checked_shl_u32((exponent_z as u32), go_shift_count_u64((23) as u64)));
                }
                if ((sign_x ^ sign_y) == sign_z) {
                    res |= sign_z;
                } else if (rnd_mode == 1) {
                    res |= 0x80000000;
                }
                return (res, pfpsf);
            }
            d2 = ((exponent_x.wrapping_add(exponent_y)).wrapping_sub(101));
            if (exponent_z > d2) {
                exponent_z = d2;
            }
        }
    }
    P0 = ((coefficient_x as u64).wrapping_mul(coefficient_y as u64));
    exponent_x = exponent_x.wrapping_add((exponent_y.wrapping_sub(101)));
    if (exponent_x < exponent_z) {
        sign_a = sign_z;
        exponent_a = exponent_z;
        coefficient_a = (coefficient_z as u64);
        sign_b = (sign_x ^ sign_y);
        exponent_b = exponent_x;
        coefficient_b = P0;
    } else {
        sign_a = (sign_x ^ sign_y);
        exponent_a = exponent_x;
        coefficient_a = P0;
        sign_b = sign_z;
        exponent_b = exponent_z;
        coefficient_b = (coefficient_z as u64);
    }
    diff_dec_expon = (exponent_a.wrapping_sub(exponent_b));
    if (diff_dec_expon > 17) {
        let mut tempx = (coefficient_a as f64).to_bits();
        bin_expon = ((((go_checked_shr_u64((tempx & 0x7ff0000000000000), go_shift_count_u64((52) as u64)))) as i64).wrapping_sub(0x3ff));
        scale_ca = (bid_estimate_decimal_digits[bin_expon as usize] as i64);
        d2 = ((31 as i64).wrapping_sub(scale_ca));
        if (diff_dec_expon > d2) {
            diff_dec_expon = d2;
            exponent_b = (exponent_a.wrapping_sub(diff_dec_expon));
        }
        if (coefficient_b != 0) {
            inexact = 1;
        }
    }
    sign_ab = ((go_checked_shl_i64((((sign_a ^ sign_b) as i32) as i64), go_shift_count_u64((32) as u64))) as u64);
    sign_ab = ((go_checked_shr_i64((sign_ab as i64), go_shift_count_u64((63) as u64))) as u64);
    CB.w[0] = (((coefficient_b.wrapping_add(sign_ab))) ^ sign_ab);
    CB.w[1] = ((go_checked_shr_i64((CB.w[0] as i64), go_shift_count_u64((63) as u64))) as u64);
    (_, Tmp) = __mul_64x128_full(coefficient_a, bid_power10_table_128[diff_dec_expon as usize]);
    P = __add_128_128(Tmp, CB);
    if ((P.w[1] as i64) < 0) {
        sign_a ^= 0x80000000;
        P.w[1] = ((0 as u64).wrapping_sub(P.w[1]));
        if (P.w[0] != 0) {
            P.w[1] = P.w[1].wrapping_sub(1);
        }
        P.w[0] = ((0 as u64).wrapping_sub(P.w[0]));
    }
    if (P.w[1] != 0) {
        let mut tempx = (P.w[1] as f64).to_bits();
        bin_expon = (((((go_checked_shr_u64((tempx & 0x7ff0000000000000), go_shift_count_u64((52) as u64)))) as i64).wrapping_sub(0x3ff)).wrapping_add(64));
        n_digits = (bid_estimate_decimal_digits[bin_expon as usize] as i64);
        if __unsigned_compare_ge_128(P, bid_power10_table_128[n_digits as usize]) {
            n_digits = n_digits.wrapping_add(1);
        }
    } else {
        if (P.w[0] != 0) {
            let mut tempx = (P.w[0] as f64).to_bits();
            bin_expon = ((((go_checked_shr_u64((tempx & 0x7ff0000000000000), go_shift_count_u64((52) as u64)))) as i64).wrapping_sub(0x3ff));
            n_digits = (bid_estimate_decimal_digits[bin_expon as usize] as i64);
            if (P.w[0] >= bid_power10_table_128[n_digits as usize].w[0]) {
                n_digits = n_digits.wrapping_add(1);
            }
        } else {
            sign_a = 0;
            if (rnd_mode == 1) {
                sign_a = 0x80000000;
            }
            if (coefficient_a == 0) {
                sign_a = sign_x;
            }
            n_digits = 0;
        }
    }
    if (n_digits <= 7) {
        res = get_bid32_uf(sign_a, exponent_b, ((P.w[0] as u32) as u64), 0, rnd_mode, (&mut pfpsf));
        return (res, pfpsf);
    }
    extra_digits = (n_digits.wrapping_sub(7));
    rmode = rnd_mode;
    if ((sign_a != 0) && (((rmode.wrapping_sub(1)) as u64) < 2)) {
        rmode = ((3 as i64).wrapping_sub(rmode));
    }
    if ((exponent_b.wrapping_add(extra_digits)) < 0) {
        rmode = 3;
    }
    if (extra_digits <= 18) {
        P = __add_128_64(P, bid_round_const_table[rmode as usize][extra_digits as usize]);
    } else {
        Stemp = __mul_64x64_to_128(bid_round_const_table[rmode as usize][18], bid_power10_table_128[(extra_digits.wrapping_sub(18)) as usize].w[0]);
        P = __add_128_128(P, Stemp);
        if (rmode == 2) {
            P = __add_128_64(P, bid_round_const_table[rmode as usize][(extra_digits.wrapping_sub(18)) as usize]);
        }
    }
    (Q_high, Q_low) = __mul_128x128_full(P, bid_reciprocals10_128[extra_digits as usize]);
    amount = (bid_recip_scale[extra_digits as usize] as i64);
    C128 = __shr_128_long(Q_high, (amount as u64));
    C64 = __low_64(C128);
    if (rmode == 0) {
        if ((C64 & 1) != 0) {
            rem_l = Q_high.w[0];
            if (amount < 64) {
                remainder_h = (go_checked_shl_u64(Q_high.w[0], go_shift_count_u64(((((64 as i64).wrapping_sub(amount)) as u64)) as u64)));
                rem_l = 0;
            } else {
                remainder_h = (go_checked_shl_u64(Q_high.w[1], go_shift_count_u64(((((128 as i64).wrapping_sub(amount)) as u64)) as u64)));
            }
            if (((remainder_h | rem_l) == 0) && (((Q_low.w[1] < bid_reciprocals10_128[extra_digits as usize].w[1]) || (((Q_low.w[1] == bid_reciprocals10_128[extra_digits as usize].w[1]) && (Q_low.w[0] < bid_reciprocals10_128[extra_digits as usize].w[0])))))) {
                C64 = C64.wrapping_sub(1);
            }
        }
    }
    status = 32;
    rem_l = Q_high.w[0];
    if (amount < 64) {
        remainder_h = (go_checked_shl_u64(Q_high.w[0], go_shift_count_u64(((((64 as i64).wrapping_sub(amount)) as u64)) as u64)));
        rem_l = 0;
    } else {
        remainder_h = (go_checked_shl_u64(Q_high.w[1], go_shift_count_u64(((((128 as i64).wrapping_sub(amount)) as u64)) as u64)));
    }
    match rmode {
        0 | 4 => {
            if ((((remainder_h == 0x8000000000000000) && (rem_l == 0))) && (((Q_low.w[1] < bid_reciprocals10_128[extra_digits as usize].w[1]) || (((Q_low.w[1] == bid_reciprocals10_128[extra_digits as usize].w[1]) && (Q_low.w[0] < bid_reciprocals10_128[extra_digits as usize].w[0])))))) {
                status = 0;
            }
        }
        1 | 3 => {
            if (((remainder_h | rem_l) == 0) && (((Q_low.w[1] < bid_reciprocals10_128[extra_digits as usize].w[1]) || (((Q_low.w[1] == bid_reciprocals10_128[extra_digits as usize].w[1]) && (Q_low.w[0] < bid_reciprocals10_128[extra_digits as usize].w[0])))))) {
                status = 0;
            }
        }
        _ => {
            (Stemp.w[0], CY) = __add_carry_out(Q_low.w[0], bid_reciprocals10_128[extra_digits as usize].w[0]);
            (Stemp.w[1], carry) = __add_carry_in_out(Q_low.w[1], bid_reciprocals10_128[extra_digits as usize].w[1], CY);
            _ = Stemp;
            if (amount < 64) {
                if ((((go_checked_shr_u64(remainder_h, go_shift_count_u64(((((64 as i64).wrapping_sub(amount)) as u64)) as u64)))).wrapping_add(carry)) >= ((go_checked_shl_u64((1 as u64), go_shift_count_u64((amount as u64) as u64))))) {
                    if (inexact == 0) {
                        status = 0;
                    }
                }
            } else {
                rem_l = rem_l.wrapping_add(carry);
                remainder_h = go_checked_shr_u64(remainder_h, go_shift_count_u64(((((128 as i64).wrapping_sub(amount)) as u64)) as u64));
                if ((carry != 0) && (rem_l == 0)) {
                    remainder_h = remainder_h.wrapping_add(1);
                }
                if ((remainder_h >= ((go_checked_shl_u64((1 as u64), go_shift_count_u64((((amount.wrapping_sub(64)) as u64)) as u64))))) && (inexact == 0)) {
                    status = 0;
                }
            }
        }
    }
    pfpsf |= status;
    R = 0;
    if (status != 0) {
        R = 1;
    }
    res = get_bid32_uf(sign_a, (exponent_b.wrapping_add(extra_digits)), ((C64 as u32) as u64), (R as u32), rnd_mode, (&mut pfpsf));
    return (res, pfpsf);
}
