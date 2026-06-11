// Auto-generated from fma64.go by go2rs. Do not edit.

use super::prelude::*;

pub fn bid64_fma(mut x: u64, mut y: u64, mut z: u64, mut rndMode: i64) -> (u64, u32) {
    let mut P: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut CT: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut CZ: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut sign_x: u64 = 0;
    let mut sign_y: u64 = 0;
    let mut coefficient_x: u64 = 0;
    let mut coefficient_y: u64 = 0;
    let mut sign_z: u64 = 0;
    let mut coefficient_z: u64 = 0;
    let mut C64: u64 = 0;
    let mut remainder_y: u64 = 0;
    let mut res: u64 = 0;
    let mut CYh: u64 = 0;
    let mut CY0L: u64 = 0;
    let mut T: u64 = 0;
    let mut valid_x: bool = false;
    let mut valid_y: bool = false;
    let mut valid_z: bool = false;
    let mut extra_digits: i64 = 0;
    let mut exponent_x: i64 = 0;
    let mut exponent_y: i64 = 0;
    let mut bin_expon_cx: i64 = 0;
    let mut bin_expon_cy: i64 = 0;
    let mut bin_expon_product: i64 = 0;
    let mut digits_p: i64 = 0;
    let mut bp: i64 = 0;
    let mut final_exponent: i64 = 0;
    let mut exponent_z: i64 = 0;
    let mut digits_z: i64 = 0;
    let mut ez: i64 = 0;
    let mut ey: i64 = 0;
    let mut scale_z: i64 = 0;
    let mut uf_status: u32 = 0;
    let mut pfpsf: u32 = 0;
    (sign_x, exponent_x, coefficient_x, valid_x) = unpack_bid64(x);
    (sign_y, exponent_y, coefficient_y, valid_y) = unpack_bid64(y);
    (sign_z, exponent_z, coefficient_z, valid_z) = unpack_bid64(z);
    if (((!valid_x) || (!valid_y)) || (!valid_z)) {
        if ((y & 0x7c00000000000000) == 0x7c00000000000000) {
            y = (y & 0xfe03ffffffffffff);
            if ((y & 0x0003ffffffffffff) > 999999999999999) {
                y = (y & 0xfe00000000000000);
            }
            if ((y & 0x7e00000000000000) == 0x7e00000000000000) {
                pfpsf |= 1;
                res = (y & 0xfdffffffffffffff);
            } else {
                res = y;
                if (((z & 0x7e00000000000000) == 0x7e00000000000000) || ((x & 0x7e00000000000000) == 0x7e00000000000000)) {
                    pfpsf |= 1;
                }
            }
            return (res, pfpsf);
        } else if ((z & 0x7c00000000000000) == 0x7c00000000000000) {
            z = (z & 0xfe03ffffffffffff);
            if ((z & 0x0003ffffffffffff) > 999999999999999) {
                z = (z & 0xfe00000000000000);
            }
            if ((z & 0x7e00000000000000) == 0x7e00000000000000) {
                pfpsf |= 1;
                res = (z & 0xfdffffffffffffff);
            } else {
                res = z;
                if ((x & 0x7e00000000000000) == 0x7e00000000000000) {
                    pfpsf |= 1;
                }
            }
            return (res, pfpsf);
        } else if ((x & 0x7c00000000000000) == 0x7c00000000000000) {
            x = (x & 0xfe03ffffffffffff);
            if ((x & 0x0003ffffffffffff) > 999999999999999) {
                x = (x & 0xfe00000000000000);
            }
            if ((x & 0x7e00000000000000) == 0x7e00000000000000) {
                pfpsf |= 1;
                res = (x & 0xfdffffffffffffff);
            } else {
                res = x;
            }
            return (res, pfpsf);
        }
        if (!valid_x) {
            if ((x & 0x7800000000000000) == 0x7800000000000000) {
                if (coefficient_y == 0) {
                    if ((z & 0x7e00000000000000) != 0x7c00000000000000) {
                        pfpsf |= 1;
                    }
                    return (0x7c00000000000000, pfpsf);
                }
                if ((((z & 0x7c00000000000000) == 0x7800000000000000)) && ((((((x ^ y) ^ z)) & 0x8000000000000000)) != 0)) {
                    pfpsf |= 1;
                    return (0x7c00000000000000, pfpsf);
                }
                return (((((x ^ y) & 0x8000000000000000)) | 0x7800000000000000), pfpsf);
            }
            if ((((y & 0x7800000000000000) != 0x7800000000000000)) && (((z & 0x7800000000000000) != 0x7800000000000000))) {
                if (coefficient_z != 0) {
                    exponent_y = ((exponent_x.wrapping_sub(0x18e)).wrapping_add(exponent_y));
                    sign_z = (z & 0x8000000000000000);
                    if (exponent_y >= exponent_z) {
                        return (z, pfpsf);
                    }
                    res = add_zero64(exponent_y, sign_z, exponent_z, coefficient_z, (&mut rndMode), (&mut pfpsf));
                    return (res, pfpsf);
                }
            }
        }
        if (!valid_y) {
            if ((y & 0x7800000000000000) == 0x7800000000000000) {
                if (coefficient_x == 0) {
                    pfpsf |= 1;
                    return (0x7c00000000000000, pfpsf);
                }
                if ((((z & 0x7c00000000000000) == 0x7800000000000000)) && ((((((x ^ y) ^ z)) & 0x8000000000000000)) != 0)) {
                    pfpsf |= 1;
                    return (0x7c00000000000000, pfpsf);
                }
                return (((((x ^ y) & 0x8000000000000000)) | 0x7800000000000000), pfpsf);
            }
            if ((z & 0x7800000000000000) != 0x7800000000000000) {
                if (coefficient_z != 0) {
                    exponent_y = exponent_y.wrapping_add((exponent_x.wrapping_sub(0x18e)));
                    sign_z = (z & 0x8000000000000000);
                    if (exponent_y >= exponent_z) {
                        return (z, pfpsf);
                    }
                    res = add_zero64(exponent_y, sign_z, exponent_z, coefficient_z, (&mut rndMode), (&mut pfpsf));
                    return (res, pfpsf);
                }
            }
        }
        if (!valid_z) {
            if ((z & 0x7800000000000000) == 0x7800000000000000) {
                return ((coefficient_z & 0xfdffffffffffffff), pfpsf);
            }
            if ((coefficient_x == 0) || (coefficient_y == 0)) {
                exponent_x = exponent_x.wrapping_add((exponent_y.wrapping_sub(0x18e)));
                if (exponent_x > 0x2ff) {
                    exponent_x = 0x2ff;
                } else if (exponent_x < 0) {
                    exponent_x = 0;
                }
                if (exponent_x <= exponent_z) {
                    res = (go_checked_shl_u64((exponent_x as u64), go_shift_count_u64((53) as u64)));
                } else {
                    res = (go_checked_shl_u64((exponent_z as u64), go_shift_count_u64((53) as u64)));
                }
                if ((sign_x ^ sign_y) == sign_z) {
                    res |= sign_z;
                } else if (rndMode == 1) {
                    res |= 0x8000000000000000;
                }
                return (res, pfpsf);
            }
        }
    }
    let mut tempx = (coefficient_x as f64).to_bits();
    bin_expon_cx = ((go_checked_shr_u64((tempx & 0x7ff0000000000000), go_shift_count_u64((52) as u64))) as i64);
    let mut tempy = (coefficient_y as f64).to_bits();
    bin_expon_cy = ((go_checked_shr_u64((tempy & 0x7ff0000000000000), go_shift_count_u64((52) as u64))) as i64);
    bin_expon_product = (bin_expon_cx.wrapping_add(bin_expon_cy));
    if (bin_expon_product < (51 + (2 * 0x3ff))) {
        C64 = (coefficient_x.wrapping_mul(coefficient_y));
        final_exponent = ((exponent_x.wrapping_add(exponent_y)).wrapping_sub(0x18e));
        if ((final_exponent > 0) || (coefficient_z == 0)) {
            res = bid_get_add64((sign_x ^ sign_y), final_exponent, C64, sign_z, exponent_z, coefficient_z, rndMode, (&mut pfpsf));
            return (res, pfpsf);
        }
        P.w[0] = C64;
        P.w[1] = 0;
        extra_digits = 0;
    } else {
        if (coefficient_z == 0) {
            let (mut res, mut flags) = bid64_mul_with_flags(x, y, rndMode);
            pfpsf |= flags;
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
    }
    if ((final_exponent as u64) >= (3 * 256)) {
        if (final_exponent < 0) {
            tempx = (coefficient_z as f64).to_bits();
            bin_expon_cx = (((go_checked_shr_u64((tempx & 0x7ff0000000000000), go_shift_count_u64((52) as u64))) as i64).wrapping_sub(0x3ff));
            digits_z = (bid_estimate_decimal_digits[bin_expon_cx as usize] as i64);
            if (coefficient_z >= bid_power10_table_128[digits_z as usize].w[0]) {
                digits_z = digits_z.wrapping_add(1);
            }
            if ((((final_exponent.wrapping_add(16)) < 0)) || (((exponent_z.wrapping_add(digits_z)) > ((33 as i64).wrapping_add(final_exponent))))) {
                res = bid_normalize(sign_z, exponent_z, coefficient_z, (sign_x ^ sign_y), 1, rndMode, (&mut pfpsf));
                return (res, pfpsf);
            }
            ez = ((exponent_z.wrapping_add(digits_z)).wrapping_sub(16));
            if (ez < 0) {
                ez = 0;
            }
            scale_z = (exponent_z.wrapping_sub(ez));
            coefficient_z = coefficient_z.wrapping_mul(bid_power10_table_128[scale_z as usize].w[0]);
            ey = (final_exponent.wrapping_sub(extra_digits));
            extra_digits = (ez.wrapping_sub(ey));
            if (extra_digits > 17) {
                CYh = __truncate(P, 16);
                T = bid_power10_table_128[16].w[0];
                CY0L = __mul_64x64_to_64(CYh, T);
                remainder_y = (P.w[0].wrapping_sub(CY0L));
                extra_digits = extra_digits.wrapping_sub(16);
                P.w[0] = CYh;
                P.w[1] = 0;
            } else {
                remainder_y = 0;
            }
            CZ = __mul_64x64_to_128(coefficient_z, bid_power10_table_128[extra_digits as usize].w[0]);
            if (sign_z == (sign_y ^ sign_x)) {
                CT = __add_128_128(CZ, P);
                if __unsigned_compare_ge_128(CT, bid_power10_table_128[((16 as i64).wrapping_add(extra_digits)) as usize]) {
                    extra_digits = extra_digits.wrapping_add(1);
                    ez = ez.wrapping_add(1);
                }
            } else {
                if ((remainder_y != 0) && (__unsigned_compare_ge_128(CZ, P))) {
                    P.w[0] = P.w[0].wrapping_add(1);
                    if (P.w[0] == 0) {
                        P.w[1] = P.w[1].wrapping_add(1);
                    }
                }
                CT = __sub_128_128(CZ, P);
                if ((CT.w[1] as i64) < 0) {
                    sign_z = (sign_y ^ sign_x);
                    CT.w[0] = ((0 as u64).wrapping_sub(CT.w[0]));
                    CT.w[1] = ((0 as u64).wrapping_sub(CT.w[1]));
                    if (CT.w[0] != 0) {
                        CT.w[1] = CT.w[1].wrapping_sub(1);
                    }
                } else if ((CT.w[1] | CT.w[0]) == 0) {
                    if (rndMode != 1) {
                        sign_z = 0;
                    } else {
                        sign_z = 0x8000000000000000;
                    }
                }
                if ((ez != 0) && __unsigned_compare_gt_128(bid_power10_table_128[((15 as i64).wrapping_add(extra_digits)) as usize], CT)) {
                    extra_digits = extra_digits.wrapping_sub(1);
                    ez = ez.wrapping_sub(1);
                }
            }
            uf_status = 0;
            if ((ez == 0) && __unsigned_compare_gt_128(bid_power10_table_128[(extra_digits.wrapping_add(15)) as usize], CT)) {
                uf_status = 16;
            }
            res = __bid_full_round64_remainder(sign_z, (ez.wrapping_sub(extra_digits)), CT, extra_digits, remainder_y, rndMode, (&mut pfpsf), uf_status);
            return (res, pfpsf);
        } else {
            if (((sign_z == (sign_x ^ sign_y))) || ((final_exponent > ((3 * 256) + 15)))) {
                let (mut res, mut flags) = fast_get_bid64_check_of_flags((sign_x ^ sign_y), final_exponent, 1000000000000000, rndMode);
                pfpsf |= flags;
                return (res, pfpsf);
            }
        }
    }
    if (extra_digits > 0) {
        res = bid_get_add128(sign_z, exponent_z, coefficient_z, (sign_x ^ sign_y), final_exponent, P, extra_digits, rndMode, (&mut pfpsf));
        return (res, pfpsf);
    }
    C64 = __low_64(P);
    res = bid_get_add64((sign_x ^ sign_y), ((exponent_x.wrapping_add(exponent_y)).wrapping_sub(0x18e)), C64, sign_z, exponent_z, coefficient_z, rndMode, (&mut pfpsf));
    return (res, pfpsf);
}
