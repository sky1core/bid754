// Auto-generated from bid128_rem.go by go2rs. Do not edit.

use super::prelude::*;

pub fn bid128_rem(mut x: BID_UINT128, mut y: BID_UINT128) -> (BID_UINT128, u32) {
    let mut P256: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut CX: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut CY: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut CX2: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut CQ: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut CR: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut T: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut CXS: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut P128: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut res: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut sign_x: u64 = 0;
    let mut sign_y: u64 = 0;
    let mut valid_y: bool = false;
    let mut D: i64 = 0;
    let mut exponent_x: i64 = 0;
    let mut exponent_y: i64 = 0;
    let mut diff_expon: i64 = 0;
    let mut bin_expon_cx: i64 = 0;
    let mut scale: i64 = 0;
    let mut scale0: i64 = 0;
    let mut pfpsf: u32 = 0;
    (sign_y, exponent_y, CY, valid_y) = unpack_bid128_value(y);
    _ = sign_y;
    let mut valid_x: bool = false;
    (sign_x, exponent_x, CX, valid_x) = unpack_bid128_value(x);
    if (!valid_x) {
        if ((y.w[1] & 0x7e00000000000000) == 0x7e00000000000000) {
            pfpsf |= 1;
        }
        if ((x.w[1] & 0x7c00000000000000) == 0x7c00000000000000) {
            if ((x.w[1] & 0x7e00000000000000) == 0x7e00000000000000) {
                pfpsf |= 1;
            }
            res.w[1] = (CX.w[1] & 0xfdffffffffffffff);
            res.w[0] = CX.w[0];
            return (res, pfpsf);
        }
        if ((x.w[1] & 0x7800000000000000) == 0x7800000000000000) {
            if ((y.w[1] & 0x7c00000000000000) != 0x7c00000000000000) {
                pfpsf |= 1;
                res.w[1] = 0x7c00000000000000;
                res.w[0] = 0;
                return (res, pfpsf);
            }
        }
        if ((CY.w[1] == 0) && (CY.w[0] == 0)) {
            pfpsf |= 1;
            res.w[1] = 0x7c00000000000000;
            res.w[0] = 0;
            return (res, pfpsf);
        }
        if (valid_y || (((y.w[1] & 0x7c00000000000000) == 0x7800000000000000))) {
            if ((exponent_x > exponent_y) && (((y.w[1] & 0x7c00000000000000) != 0x7800000000000000))) {
                exponent_x = exponent_y;
            }
            res.w[1] = (sign_x | ((go_checked_shl_u64((exponent_x as u64), go_shift_count_u64((49) as u64)))));
            res.w[0] = 0;
            return (res, pfpsf);
        }
    }
    if (!valid_y) {
        if ((y.w[1] & 0x7c00000000000000) == 0x7c00000000000000) {
            if ((y.w[1] & 0x7e00000000000000) == 0x7e00000000000000) {
                pfpsf |= 1;
            }
            res.w[1] = (CY.w[1] & 0xfdffffffffffffff);
            res.w[0] = CY.w[0];
            return (res, pfpsf);
        }
        if ((y.w[1] & 0x7800000000000000) == 0x7800000000000000) {
            res.w[1] = x.w[1];
            res.w[0] = x.w[0];
            return (res, pfpsf);
        }
        pfpsf |= 1;
        res.w[1] = 0x7c00000000000000;
        res.w[0] = 0;
        return (res, pfpsf);
    }
    diff_expon = (exponent_x.wrapping_sub(exponent_y));
    if (diff_expon <= 0) {
        diff_expon = (diff_expon.wrapping_neg());
        if (diff_expon > 34) {
            res = x;
            return (res, pfpsf);
        }
        T = bid_power10_table_128[diff_expon as usize];
        P256 = __mul_128x128_to_256(CY, T);
        if ((P256.w[2] != 0) || (P256.w[3] != 0)) {
            res = x;
            return (res, pfpsf);
        }
        CX2.w[1] = (((go_checked_shl_u64(CX.w[1], go_shift_count_u64((1) as u64)))) | ((go_checked_shr_u64(CX.w[0], go_shift_count_u64((63) as u64)))));
        CX2.w[0] = (go_checked_shl_u64(CX.w[0], go_shift_count_u64((1) as u64)));
        let mut P256_128 = BID_UINT128 { w: [P256.w[0], P256.w[1]], ..Default::default() };
        if __unsigned_compare_ge_128(P256_128, CX2) {
            res = x;
            return (res, pfpsf);
        }
        P128.w[0] = P256.w[0];
        P128.w[1] = P256.w[1];
        (CQ, CR) = bid___div_128_by_128(CX, P128);
        CX2.w[1] = (((go_checked_shl_u64(CR.w[1], go_shift_count_u64((1) as u64)))) | ((go_checked_shr_u64(CR.w[0], go_shift_count_u64((63) as u64)))));
        CX2.w[0] = (go_checked_shl_u64(CR.w[0], go_shift_count_u64((1) as u64)));
        if (__unsigned_compare_gt_128(CX2, P256_128) || ((((CX2.w[1] == P256.w[1]) && (CX2.w[0] == P256.w[0])) && ((CQ.w[0] & 1) != 0)))) {
            CR = __sub_128_128(P256_128, CR);
            sign_x ^= 0x8000000000000000;
        }
        res = very_fast_get_bid128(sign_x, exponent_x, CR);
        return (res, pfpsf);
    }
    let mut f64_d = f32::from_bits(0x5f800000);
    scale0 = 38;
    if (CY.w[1] == 0) {
        scale0 = 34;
    }
    while (diff_expon > 0) {
        let mut fx_d = no_fma_mul_add_f32((CX.w[1] as f32), f64_d, (CX.w[0] as f32));
        let mut fx_i = (fx_d as f32).to_bits();
        bin_expon_cx = (((((go_checked_shr_u32(fx_i, go_shift_count_u64((23) as u64)))) & 0xff) as i64).wrapping_sub(0x7f));
        scale = (scale0.wrapping_sub(bid_estimate_decimal_digits[bin_expon_cx as usize] as i64));
        D = ((CX.w[1] as i64).wrapping_sub(bid_power10_index_binexp_128[bin_expon_cx as usize].w[1] as i64));
        if ((D > 0) || (((D == 0) && (CX.w[0] >= bid_power10_index_binexp_128[bin_expon_cx as usize].w[0])))) {
            scale = scale.wrapping_sub(1);
        }
        if (diff_expon >= scale) {
            diff_expon = diff_expon.wrapping_sub(scale);
        } else {
            scale = diff_expon;
            diff_expon = 0;
        }
        T = bid_power10_table_128[scale as usize];
        CXS = __mul_128x128_low(CX, T);
        (CQ, CX) = bid___div_128_by_128(CXS, CY);
        if ((CX.w[1] == 0) && (CX.w[0] == 0)) {
            res = very_fast_get_bid128(sign_x, exponent_y, CX);
            return (res, pfpsf);
        }
    }
    CX2.w[1] = (((go_checked_shl_u64(CX.w[1], go_shift_count_u64((1) as u64)))) | ((go_checked_shr_u64(CX.w[0], go_shift_count_u64((63) as u64)))));
    CX2.w[0] = (go_checked_shl_u64(CX.w[0], go_shift_count_u64((1) as u64)));
    if (__unsigned_compare_gt_128(CX2, CY) || ((((CX2.w[1] == CY.w[1]) && (CX2.w[0] == CY.w[0])) && ((CQ.w[0] & 1) != 0)))) {
        CX = __sub_128_128(CY, CX);
        sign_x ^= 0x8000000000000000;
    }
    res = very_fast_get_bid128(sign_x, exponent_y, CX);
    return (res, pfpsf);
}

pub fn bid128_fmod(mut x: BID_UINT128, mut y: BID_UINT128) -> (BID_UINT128, u32) {
    let mut P256: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut CX: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut CY: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut CQ: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut CR: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut T: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut CXS: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut P128: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut res: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut sign_x: u64 = 0;
    let mut sign_y: u64 = 0;
    let mut valid_y: bool = false;
    let mut D: i64 = 0;
    let mut exponent_x: i64 = 0;
    let mut exponent_y: i64 = 0;
    let mut diff_expon: i64 = 0;
    let mut bin_expon_cx: i64 = 0;
    let mut scale: i64 = 0;
    let mut scale0: i64 = 0;
    let mut pfpsf: u32 = 0;
    (sign_y, exponent_y, CY, valid_y) = unpack_bid128_value(y);
    _ = sign_y;
    let mut valid_x: bool = false;
    (sign_x, exponent_x, CX, valid_x) = unpack_bid128_value(x);
    if (!valid_x) {
        if ((y.w[1] & 0x7e00000000000000) == 0x7e00000000000000) {
            pfpsf |= 1;
        }
        if ((x.w[1] & 0x7c00000000000000) == 0x7c00000000000000) {
            if ((x.w[1] & 0x7e00000000000000) == 0x7e00000000000000) {
                pfpsf |= 1;
            }
            res.w[1] = (CX.w[1] & 0xfdffffffffffffff);
            res.w[0] = CX.w[0];
            return (res, pfpsf);
        }
        if ((x.w[1] & 0x7800000000000000) == 0x7800000000000000) {
            if ((y.w[1] & 0x7c00000000000000) != 0x7c00000000000000) {
                pfpsf |= 1;
                res.w[1] = 0x7c00000000000000;
                res.w[0] = 0;
                return (res, pfpsf);
            }
        }
        if ((CY.w[1] == 0) && (CY.w[0] == 0)) {
            pfpsf |= 1;
            res.w[1] = 0x7c00000000000000;
            res.w[0] = 0;
            return (res, pfpsf);
        }
        if (valid_y || (((y.w[1] & 0x7c00000000000000) == 0x7800000000000000))) {
            if ((exponent_x > exponent_y) && (((y.w[1] & 0x7c00000000000000) != 0x7800000000000000))) {
                exponent_x = exponent_y;
            }
            res.w[1] = (sign_x | ((go_checked_shl_u64((exponent_x as u64), go_shift_count_u64((49) as u64)))));
            res.w[0] = 0;
            return (res, pfpsf);
        }
    }
    if (!valid_y) {
        if ((y.w[1] & 0x7c00000000000000) == 0x7c00000000000000) {
            if ((y.w[1] & 0x7e00000000000000) == 0x7e00000000000000) {
                pfpsf |= 1;
            }
            res.w[1] = (CY.w[1] & 0xfdffffffffffffff);
            res.w[0] = CY.w[0];
            return (res, pfpsf);
        }
        if ((y.w[1] & 0x7800000000000000) == 0x7800000000000000) {
            res.w[1] = x.w[1];
            res.w[0] = x.w[0];
            return (res, pfpsf);
        }
        pfpsf |= 1;
        res.w[1] = 0x7c00000000000000;
        res.w[0] = 0;
        return (res, pfpsf);
    }
    diff_expon = (exponent_x.wrapping_sub(exponent_y));
    if (diff_expon <= 0) {
        diff_expon = (diff_expon.wrapping_neg());
        if (diff_expon > 34) {
            res = x;
            return (res, pfpsf);
        }
        T = bid_power10_table_128[diff_expon as usize];
        P256 = __mul_128x128_to_256(CY, T);
        if ((P256.w[2] != 0) || (P256.w[3] != 0)) {
            res = x;
            return (res, pfpsf);
        }
        if __unsigned_compare_gt_128(BID_UINT128 { w: [P256.w[0], P256.w[1]], ..Default::default() }, CX) {
            res = x;
            return (res, pfpsf);
        }
        P128.w[0] = P256.w[0];
        P128.w[1] = P256.w[1];
        (_, CR) = bid___div_128_by_128(CX, P128);
        res = very_fast_get_bid128(sign_x, exponent_x, CR);
        return (res, pfpsf);
    }
    let mut f64_d = f32::from_bits(0x5f800000);
    scale0 = 38;
    if (CY.w[1] == 0) {
        scale0 = 34;
    }
    while (diff_expon > 0) {
        let mut fx_d = no_fma_mul_add_f32((CX.w[1] as f32), f64_d, (CX.w[0] as f32));
        let mut fx_i = (fx_d as f32).to_bits();
        bin_expon_cx = (((((go_checked_shr_u32(fx_i, go_shift_count_u64((23) as u64)))) & 0xff) as i64).wrapping_sub(0x7f));
        scale = (scale0.wrapping_sub(bid_estimate_decimal_digits[bin_expon_cx as usize] as i64));
        D = ((CX.w[1] as i64).wrapping_sub(bid_power10_index_binexp_128[bin_expon_cx as usize].w[1] as i64));
        if ((D > 0) || (((D == 0) && (CX.w[0] >= bid_power10_index_binexp_128[bin_expon_cx as usize].w[0])))) {
            scale = scale.wrapping_sub(1);
        }
        if (diff_expon >= scale) {
            diff_expon = diff_expon.wrapping_sub(scale);
        } else {
            scale = diff_expon;
            diff_expon = 0;
        }
        T = bid_power10_table_128[scale as usize];
        CXS = __mul_128x128_low(CX, T);
        (CQ, CX) = bid___div_128_by_128(CXS, CY);
        _ = CQ;
        if ((CX.w[1] == 0) && (CX.w[0] == 0)) {
            res = very_fast_get_bid128(sign_x, exponent_y, CX);
            return (res, pfpsf);
        }
    }
    res = very_fast_get_bid128(sign_x, exponent_y, CX);
    return (res, pfpsf);
}
