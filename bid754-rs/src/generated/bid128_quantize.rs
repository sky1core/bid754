// Auto-generated from bid128_quantize.go by go2rs. Do not edit.

use super::prelude::*;

pub fn bid128_quantize(mut x: BID_UINT128, mut y: BID_UINT128, mut rnd_mode: i64) -> (BID_UINT128, u32) {
    let mut CT: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut CX: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut CY: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut T: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut CX2: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut CR: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut Stemp: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut res: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut REM_H: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut C2N: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut sign_x: u64 = 0;
    let mut remainder_h: u64 = 0;
    let mut carry: u64 = 0;
    let mut CY64: u64 = 0;
    let mut valid_x: bool = false;
    let mut exponent_x: i64 = 0;
    let mut exponent_y: i64 = 0;
    let mut digits_x: i64 = 0;
    let mut extra_digits: i64 = 0;
    let mut amount: i64 = 0;
    let mut expon_diff: i64 = 0;
    let mut total_digits: i64 = 0;
    let mut bin_expon_cx: i64 = 0;
    let mut rmode: i64 = 0;
    let mut status: i64 = 0;
    let mut pfpsf: u32 = 0;
    (sign_x, exponent_x, CX, valid_x) = unpack_bid128_value(x);
    let (_, mut exponent_y, mut CY, mut valid_y) = unpack_bid128_value(y);
    if (!valid_y) {
        if ((x.w[1] & 0x7e00000000000000) == 0x7e00000000000000) {
            pfpsf |= 1;
        }
        if ((y.w[1] & 0x7c00000000000000) == 0x7c00000000000000) {
            if ((y.w[1] & 0x7e00000000000000) == 0x7e00000000000000) {
                pfpsf |= 1;
            }
            if ((x.w[1] & 0x7c00000000000000) != 0x7c00000000000000) {
                res.w[1] = (CY.w[1] & 0xfdffffffffffffff);
                res.w[0] = CY.w[0];
            } else {
                res.w[1] = (CX.w[1] & 0xfdffffffffffffff);
                res.w[0] = CX.w[0];
            }
            return (res, pfpsf);
        }
        if ((y.w[1] & 0x7800000000000000) == 0x7800000000000000) {
            if ((x.w[1] & 0x7c00000000000000) < 0x7800000000000000) {
                pfpsf |= 1;
                res.w[1] = 0x7c00000000000000;
                res.w[0] = 0;
                return (res, pfpsf);
            } else if ((x.w[1] & 0x7c00000000000000) <= 0x7800000000000000) {
                res.w[1] = (CX.w[1] & 0xfdffffffffffffff);
                res.w[0] = CX.w[0];
                return (res, pfpsf);
            }
        }
    }
    if (!valid_x) {
        if ((x.w[1] & 0x7c00000000000000) == 0x7800000000000000) {
            pfpsf |= 1;
            res.w[1] = 0x7c00000000000000;
            res.w[0] = 0;
            return (res, pfpsf);
        } else if ((x.w[1] & 0x7c00000000000000) == 0x7c00000000000000) {
            if ((x.w[1] & 0x7e00000000000000) == 0x7e00000000000000) {
                pfpsf |= 1;
            }
            res.w[1] = (CX.w[1] & 0xfdffffffffffffff);
            res.w[0] = CX.w[0];
            return (res, pfpsf);
        }
        if ((CX.w[1] == 0) && (CX.w[0] == 0)) {
            res = very_fast_get_bid128(sign_x, exponent_y, CX);
            return (res, pfpsf);
        }
    }
    if (CX.w[1] != 0) {
        let mut tempx = ((CX.w[1] as f32) as f32).to_bits();
        bin_expon_cx = ((((((go_checked_shr_u32(tempx, go_shift_count_u64((23) as u64)))) & 0xff) as i64).wrapping_sub(0x7f)).wrapping_add(64));
    } else {
        let mut tempx = ((CX.w[0] as f32) as f32).to_bits();
        bin_expon_cx = (((((go_checked_shr_u32(tempx, go_shift_count_u64((23) as u64)))) & 0xff) as i64).wrapping_sub(0x7f));
    }
    digits_x = (bid_estimate_decimal_digits[bin_expon_cx as usize] as i64);
    if ((CX.w[1] > bid_power10_table_128[digits_x as usize].w[1]) || (((CX.w[1] == bid_power10_table_128[digits_x as usize].w[1]) && (CX.w[0] >= bid_power10_table_128[digits_x as usize].w[0])))) {
        digits_x = digits_x.wrapping_add(1);
    }
    expon_diff = (exponent_x.wrapping_sub(exponent_y));
    total_digits = (digits_x.wrapping_add(expon_diff));
    if ((total_digits as u32) <= 34) {
        if (expon_diff >= 0) {
            T = bid_power10_table_128[expon_diff as usize];
            CX2 = __mul_128x128_low(T, CX);
            res = very_fast_get_bid128(sign_x, exponent_y, CX2);
            return (res, pfpsf);
        }
        rmode = rnd_mode;
        if ((sign_x != 0) && (((rmode.wrapping_sub(1)) as u64) < 2)) {
            rmode = ((3 as i64).wrapping_sub(rmode));
        }
        extra_digits = (expon_diff.wrapping_neg());
        CX = __add_128_128(CX, bid_round_const_table_128[rmode as usize][extra_digits as usize]);
        CT = __mul_128x128_to_256(CX, bid_reciprocals10_128[extra_digits as usize]);
        amount = (bid_recip_scale[extra_digits as usize] as i64);
        CX2.w[0] = CT.w[2];
        CX2.w[1] = CT.w[3];
        if (amount >= 64) {
            CR.w[1] = 0;
            CR.w[0] = (go_checked_shr_u64(CX2.w[1], go_shift_count_u64((((amount.wrapping_sub(64)) as u64)) as u64)));
        } else {
            CR = __shr_128(CX2, (amount as u64));
        }
        if (rnd_mode == 0) {
            if ((CR.w[0] & 1) != 0) {
                if (amount >= 64) {
                    remainder_h = (CX2.w[0] | ((go_checked_shl_u64(CX2.w[1], go_shift_count_u64(((((128 as i64).wrapping_sub(amount)) as u64)) as u64)))));
                } else {
                    remainder_h = (go_checked_shl_u64(CX2.w[0], go_shift_count_u64(((((64 as i64).wrapping_sub(amount)) as u64)) as u64)));
                }
                if ((remainder_h == 0) && (((CT.w[1] < bid_reciprocals10_128[extra_digits as usize].w[1]) || (((CT.w[1] == bid_reciprocals10_128[extra_digits as usize].w[1]) && (CT.w[0] < bid_reciprocals10_128[extra_digits as usize].w[0])))))) {
                    CR.w[0] = CR.w[0].wrapping_sub(1);
                }
            }
        }
        status = (32 as i64);
        if (amount >= 64) {
            REM_H.w[1] = (go_checked_shl_u64(CX2.w[1], go_shift_count_u64(((((128 as i64).wrapping_sub(amount)) as u64)) as u64)));
            REM_H.w[0] = CX2.w[0];
        } else {
            REM_H.w[1] = (go_checked_shl_u64(CX2.w[0], go_shift_count_u64(((((64 as i64).wrapping_sub(amount)) as u64)) as u64)));
            REM_H.w[0] = 0;
        }
        match rmode {
            0 | 4 => {
                if (((REM_H.w[1] == 0x8000000000000000) && (REM_H.w[0] == 0)) && (((CT.w[1] < bid_reciprocals10_128[extra_digits as usize].w[1]) || (((CT.w[1] == bid_reciprocals10_128[extra_digits as usize].w[1]) && (CT.w[0] < bid_reciprocals10_128[extra_digits as usize].w[0])))))) {
                    status = (0 as i64);
                }
            }
            1 | 3 => {
                if (((REM_H.w[1] | REM_H.w[0]) == 0) && (((CT.w[1] < bid_reciprocals10_128[extra_digits as usize].w[1]) || (((CT.w[1] == bid_reciprocals10_128[extra_digits as usize].w[1]) && (CT.w[0] < bid_reciprocals10_128[extra_digits as usize].w[0])))))) {
                    status = (0 as i64);
                }
            }
            _ => {
                (Stemp.w[0], CY64) = __add_carry_out(CT.w[0], bid_reciprocals10_128[extra_digits as usize].w[0]);
                (Stemp.w[1], carry) = __add_carry_in_out(CT.w[1], bid_reciprocals10_128[extra_digits as usize].w[1], CY64);
                if (amount < 64) {
                    C2N.w[1] = 0;
                    C2N.w[0] = (go_checked_shl_u64((1 as u64), go_shift_count_u64((amount as u64) as u64)));
                    REM_H.w[0] = (go_checked_shr_u64(REM_H.w[1], go_shift_count_u64(((((64 as i64).wrapping_sub(amount)) as u64)) as u64)));
                    REM_H.w[1] = 0;
                } else {
                    C2N.w[1] = (go_checked_shl_u64((1 as u64), go_shift_count_u64((((amount.wrapping_sub(64)) as u64)) as u64)));
                    C2N.w[0] = 0;
                    REM_H.w[1] = go_checked_shr_u64(REM_H.w[1], go_shift_count_u64(((((128 as i64).wrapping_sub(amount)) as u64)) as u64));
                }
                REM_H.w[0] = REM_H.w[0].wrapping_add(carry);
                if (REM_H.w[0] < carry) {
                    REM_H.w[1] = REM_H.w[1].wrapping_add(1);
                }
                if __unsigned_compare_ge_128(REM_H, C2N) {
                    status = (0 as i64);
                }
            }
        }
        pfpsf |= (status as u32);
        res = very_fast_get_bid128(sign_x, exponent_y, CR);
        return (res, pfpsf);
    }
    if (total_digits < 0) {
        CR.w[1] = 0;
        CR.w[0] = 0;
        rmode = rnd_mode;
        if ((sign_x != 0) && (((rmode.wrapping_sub(1)) as u64) < 2)) {
            rmode = ((3 as i64).wrapping_sub(rmode));
        }
        if (rmode == 2) {
            CR.w[0] = 1;
        }
        pfpsf |= 32;
        res = very_fast_get_bid128(sign_x, exponent_y, CR);
        return (res, pfpsf);
    }
    pfpsf |= 1;
    res.w[1] = 0x7c00000000000000;
    res.w[0] = 0;
    return (res, pfpsf);
}
