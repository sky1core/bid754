// Auto-generated from bid128_sqrt.go by go2rs. Do not edit.

use super::prelude::*;

pub(crate) fn short_sqrt128(mut A10: BID_UINT128) -> u64 {
    let mut ARS: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut S: BID_UINT192 = BID_UINT192 { w: [0, 0, 0] };
    let mut AE0: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut AE: BID_UINT192 = BID_UINT192 { w: [0, 0, 0] };
    let mut MY: u64 = 0;
    let mut ES: u64 = 0;
    let mut CY: u64 = 0;
    let mut l64 = f64::from_bits(0x43f0000000000000);
    let mut lx = no_fma_mul_add_f64((A10.w[1] as f64), l64, (A10.w[0] as f64));
    let mut ly_d = (1.0 / (lx).sqrt());
    let mut ly_i = (ly_d).to_bits();
    MY = ((ly_i & 0x000fffffffffffff) | 0x0010000000000000);
    let mut ey = (((0x3ff as u64).wrapping_sub(((go_checked_shr_u64(ly_i, go_shift_count_u64((52) as u64)))))) as i64);
    let mut ARS0 = __mul_64x128_to_192(MY, A10);
    ARS = __mul_64x192_to_256(MY, ARS0);
    let mut k = ((((go_checked_shl_i64(ey, go_shift_count_u64((1) as u64)))).wrapping_add(104)).wrapping_sub(64));
    if (k >= 128) {
        if (k > 128) {
            ES = (((go_checked_shr_u64(ARS.w[2], go_shift_count_u64((((k.wrapping_sub(128)) as u64)) as u64)))) | ((go_checked_shl_u64(ARS.w[3], go_shift_count_u64(((((192 as i64).wrapping_sub(k)) as u64)) as u64)))));
        } else {
            ES = ARS.w[2];
        }
    } else {
        if (k >= 64) {
            ARS.w[0] = ARS.w[1];
            ARS.w[1] = ARS.w[2];
            k = k.wrapping_sub(64);
        }
        if (k != 0) {
            let mut ARS_128 = __shr_128(BID_UINT128 { w: [ARS.w[0], ARS.w[1]], ..Default::default() }, (k as u64));
            ARS.w[0] = ARS_128.w[0];
            ARS.w[1] = ARS_128.w[1];
        }
        ES = ARS.w[0];
    }
    ES = ((go_checked_shr_i64((ES as i64), go_shift_count_u64((1) as u64))) as u64);
    if ((ES as i64) < 0) {
        ES = (ES.wrapping_neg());
        AE0 = __mul_64x192_to_256(ES, ARS0);
        AE.w[0] = AE0.w[1];
        AE.w[1] = AE0.w[2];
        AE.w[2] = AE0.w[3];
        (S.w[0], CY) = __add_carry_out(ARS0.w[0], AE.w[0]);
        (S.w[1], CY) = __add_carry_in_out(ARS0.w[1], AE.w[1], CY);
        S.w[2] = ((ARS0.w[2].wrapping_add(AE.w[2])).wrapping_add(CY));
    } else {
        AE0 = __mul_64x192_to_256(ES, ARS0);
        AE.w[0] = AE0.w[1];
        AE.w[1] = AE0.w[2];
        AE.w[2] = AE0.w[3];
        (S.w[0], CY) = __sub_borrow_out(ARS0.w[0], AE.w[0]);
        (S.w[1], CY) = __sub_borrow_in_out(ARS0.w[1], AE.w[1], CY);
        S.w[2] = ((ARS0.w[2].wrapping_sub(AE.w[2])).wrapping_sub(CY));
    }
    k = (ey.wrapping_add(51));
    if (k >= 64) {
        if (k >= 128) {
            S.w[0] = S.w[2];
            S.w[1] = 0;
            k = k.wrapping_sub(128);
        } else {
            S.w[0] = S.w[1];
            S.w[1] = S.w[2];
        }
        k = k.wrapping_sub(64);
    }
    if (k != 0) {
        let mut S_128 = __shr_128(BID_UINT128 { w: [S.w[0], S.w[1]], ..Default::default() }, (k as u64));
        S.w[0] = S_128.w[0];
        S.w[1] = S_128.w[1];
    }
    return (go_checked_shr_u64(((S.w[0].wrapping_add(1))), go_shift_count_u64((1) as u64)));
}

pub(crate) fn bid_long_sqrt128(mut C256: BID_UINT256) -> BID_UINT128 {
    let mut S: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut ES: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut ARS1: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut ES2: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut ARS00: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut AE: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut AE2: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut CY: u64 = 0;
    let mut MY: u64 = 0;
    let mut ES32: u64 = 0;
    let mut l64 = f64::from_bits(0x43f0000000000000);
    let mut l128 = (l64 * l64);
    let mut lx = (((C256.w[3] as f64) * l64) * l128);
    let mut l2 = ((C256.w[2] as f64) * l128);
    lx = (lx + l2);
    let mut l1 = ((C256.w[1] as f64) * l64);
    lx = (lx + l1);
    let mut l0 = (C256.w[0] as f64);
    lx = (lx + l0);
    let mut ly_d = (1.0 / (lx).sqrt());
    let mut ly_i = (ly_d).to_bits();
    MY = ((ly_i & 0x000fffffffffffff) | 0x0010000000000000);
    let mut ey = (((0x3ff as u64).wrapping_sub(((go_checked_shr_u64(ly_i, go_shift_count_u64((52) as u64)))))) as i64);
    let mut ARS0 = __mul_64x256_to_320(MY, C256);
    let mut ARS = __mul_64x320_to_384(MY, ARS0);
    let mut k = (((((go_checked_shl_i64(ey, go_shift_count_u64((1) as u64)))).wrapping_add(104)).wrapping_sub(128)).wrapping_sub(192));
    let mut k2 = ((64 as i64).wrapping_sub(k));
    ES.w[0] = (((go_checked_shr_u64(ARS.w[3], go_shift_count_u64((((k.wrapping_add(1)) as u64)) as u64)))) | ((go_checked_shl_u64(ARS.w[4], go_shift_count_u64((((k2.wrapping_sub(1)) as u64)) as u64)))));
    ES.w[1] = (((go_checked_shr_u64(ARS.w[4], go_shift_count_u64((k as u64) as u64)))) | ((go_checked_shl_u64(ARS.w[5], go_shift_count_u64((k2 as u64) as u64)))));
    ES.w[1] = ((go_checked_shr_i64((ES.w[1] as i64), go_shift_count_u64((1) as u64))) as u64);
    ARS1.w[0] = ARS0.w[3];
    ARS1.w[1] = ARS0.w[4];
    ARS00.w[0] = ARS0.w[1];
    ARS00.w[1] = ARS0.w[2];
    ARS00.w[2] = ARS0.w[3];
    ARS00.w[3] = ARS0.w[4];
    if ((ES.w[1] as i64) < 0) {
        ES.w[0] = (ES.w[0].wrapping_neg());
        ES.w[1] = (ES.w[1].wrapping_neg());
        if (ES.w[0] != 0) {
            ES.w[1] = ES.w[1].wrapping_sub(1);
        }
        AE = __mul_128x128_to_256(ES, ARS1);
        (S.w[0], CY) = __add_carry_out(ARS00.w[0], AE.w[0]);
        (S.w[1], CY) = __add_carry_in_out(ARS00.w[1], AE.w[1], CY);
        (S.w[2], CY) = __add_carry_in_out(ARS00.w[2], AE.w[2], CY);
        S.w[3] = ((ARS00.w[3].wrapping_add(AE.w[3])).wrapping_add(CY));
    } else {
        AE = __mul_128x128_to_256(ES, ARS1);
        (S.w[0], CY) = __sub_borrow_out(ARS00.w[0], AE.w[0]);
        (S.w[1], CY) = __sub_borrow_in_out(ARS00.w[1], AE.w[1], CY);
        (S.w[2], CY) = __sub_borrow_in_out(ARS00.w[2], AE.w[2], CY);
        S.w[3] = ((ARS00.w[3].wrapping_sub(AE.w[3])).wrapping_sub(CY));
    }
    ES32 = (ES.w[1].wrapping_add(((go_checked_shr_u64(ES.w[1], go_shift_count_u64((1) as u64))))));
    ES2 = __mul_64x64_to_128(ES32, ES.w[1]);
    AE2 = __mul_128x128_to_256(ES2, ARS1);
    (S.w[0], CY) = __add_carry_out(S.w[0], AE2.w[0]);
    (S.w[1], CY) = __add_carry_in_out(S.w[1], AE2.w[1], CY);
    (S.w[2], CY) = __add_carry_in_out(S.w[2], AE2.w[2], CY);
    S.w[3] = ((S.w[3].wrapping_add(AE2.w[3])).wrapping_add(CY));
    k = ((ey.wrapping_add(51)).wrapping_sub(128));
    k2 = ((64 as i64).wrapping_sub(k));
    S.w[0] = (((go_checked_shr_u64(S.w[1], go_shift_count_u64((k as u64) as u64)))) | ((go_checked_shl_u64(S.w[2], go_shift_count_u64((k2 as u64) as u64)))));
    S.w[1] = (((go_checked_shr_u64(S.w[2], go_shift_count_u64((k as u64) as u64)))) | ((go_checked_shl_u64(S.w[3], go_shift_count_u64((k2 as u64) as u64)))));
    S.w[0] = S.w[0].wrapping_add(1);
    if (S.w[0] == 0) {
        S.w[1] = S.w[1].wrapping_add(1);
    }
    let mut CS: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    CS.w[0] = (((go_checked_shl_u64(S.w[1], go_shift_count_u64((63) as u64)))) | ((go_checked_shr_u64(S.w[0], go_shift_count_u64((1) as u64)))));
    CS.w[1] = (go_checked_shr_u64(S.w[1], go_shift_count_u64((1) as u64)));
    return CS;
}

pub fn bid128_sqrt(mut x: BID_UINT128, mut rnd_mode: i64) -> (BID_UINT128, u32) {
    let mut M256: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut C256: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut C4: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut C8: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut CX: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut CX1: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut CX2: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut A10: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut S2: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut T128: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut TP128: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut CS: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut CSM: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut res: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut sign_x: u64 = 0;
    let mut Carry: u64 = 0;
    let mut D: i64 = 0;
    let mut exponent_x: i64 = 0;
    let mut bin_expon_cx: i64 = 0;
    let mut digits: i64 = 0;
    let mut scale: i64 = 0;
    let mut exponent_q: i64 = 0;
    let mut pfpsf: u32 = 0;
    let (mut sign_x, mut exponent_x, mut CX, mut validBool) = unpack_bid128_value(x);
    if (!validBool) {
        res.w[1] = CX.w[1];
        res.w[0] = CX.w[0];
        if ((x.w[1] & 0x7c00000000000000) == 0x7c00000000000000) {
            if ((x.w[1] & 0x7e00000000000000) == 0x7e00000000000000) {
                pfpsf |= 1;
            }
            res.w[1] = (CX.w[1] & 0xfdffffffffffffff);
            return (res, pfpsf);
        }
        if ((x.w[1] & 0x7800000000000000) == 0x7800000000000000) {
            res.w[1] = CX.w[1];
            if (sign_x != 0) {
                res.w[1] = 0x7c00000000000000;
                pfpsf |= 1;
            }
            return (res, pfpsf);
        }
        res.w[1] = (sign_x | ((go_checked_shl_u64(((go_checked_shr_u64(((exponent_x.wrapping_add(0x1820)) as u64), go_shift_count_u64((1) as u64)))), go_shift_count_u64((49) as u64)))));
        res.w[0] = 0;
        return (res, pfpsf);
    }
    if (sign_x != 0) {
        res.w[1] = 0x7c00000000000000;
        res.w[0] = 0;
        pfpsf |= 1;
        return (res, pfpsf);
    }
    let mut f64_i: u32 = (0x5f800000 as u32);
    let mut f64_d = f32::from_bits(f64_i);
    let mut fx_d = no_fma_mul_add_f32((CX.w[1] as f32), f64_d, (CX.w[0] as f32));
    let mut fx_i = (fx_d as f32).to_bits();
    bin_expon_cx = (((((go_checked_shr_u32(fx_i, go_shift_count_u64((23) as u64)))) & 0xff) as i64).wrapping_sub(0x7f));
    digits = (bid_estimate_decimal_digits[bin_expon_cx as usize] as i64);
    A10 = CX;
    if ((exponent_x & 1) != 0) {
        A10.w[1] = (((go_checked_shl_u64(CX.w[1], go_shift_count_u64((3) as u64)))) | ((go_checked_shr_u64(CX.w[0], go_shift_count_u64((61) as u64)))));
        A10.w[0] = (go_checked_shl_u64(CX.w[0], go_shift_count_u64((3) as u64)));
        CX2.w[1] = (((go_checked_shl_u64(CX.w[1], go_shift_count_u64((1) as u64)))) | ((go_checked_shr_u64(CX.w[0], go_shift_count_u64((63) as u64)))));
        CX2.w[0] = (go_checked_shl_u64(CX.w[0], go_shift_count_u64((1) as u64)));
        A10 = __add_128_128(A10, CX2);
    }
    CS.w[0] = short_sqrt128(A10);
    CS.w[1] = 0;
    if ((CS.w[0].wrapping_mul(CS.w[0])) == A10.w[0]) {
        S2 = __mul_64x64_to_128_fast(CS.w[0], CS.w[0]);
        if (S2.w[1] == A10.w[1]) {
            res = very_fast_get_bid128(0, (go_checked_shr_i64(((exponent_x.wrapping_add(0x1820))), go_shift_count_u64((1) as u64))), CS);
            return (res, pfpsf);
        }
    }
    D = ((CX.w[1] as i64).wrapping_sub(bid_power10_index_binexp_128[bin_expon_cx as usize].w[1] as i64));
    if ((D > 0) || (((D == 0) && (CX.w[0] >= bid_power10_index_binexp_128[bin_expon_cx as usize].w[0])))) {
        digits = digits.wrapping_add(1);
    }
    scale = ((67 as i64).wrapping_sub(digits));
    exponent_q = (exponent_x.wrapping_sub(scale));
    scale = scale.wrapping_add(exponent_q & 1);
    if (scale > 38) {
        T128 = bid_power10_table_128[(scale.wrapping_sub(37)) as usize];
        CX1 = __mul_128x128_low(CX, T128);
        TP128 = bid_power10_table_128[37];
        C256 = __mul_128x128_to_256(CX1, TP128);
    } else {
        T128 = bid_power10_table_128[scale as usize];
        C256 = __mul_128x128_to_256(CX, T128);
    }
    C4.w[3] = (((go_checked_shl_u64(C256.w[3], go_shift_count_u64((2) as u64)))) | ((go_checked_shr_u64(C256.w[2], go_shift_count_u64((62) as u64)))));
    C4.w[2] = (((go_checked_shl_u64(C256.w[2], go_shift_count_u64((2) as u64)))) | ((go_checked_shr_u64(C256.w[1], go_shift_count_u64((62) as u64)))));
    C4.w[1] = (((go_checked_shl_u64(C256.w[1], go_shift_count_u64((2) as u64)))) | ((go_checked_shr_u64(C256.w[0], go_shift_count_u64((62) as u64)))));
    C4.w[0] = (go_checked_shl_u64(C256.w[0], go_shift_count_u64((2) as u64)));
    CS = bid_long_sqrt128(C256);
    if ((rnd_mode & 3) == 0) {
        CSM.w[1] = (((go_checked_shl_u64(CS.w[1], go_shift_count_u64((1) as u64)))) | ((go_checked_shr_u64(CS.w[0], go_shift_count_u64((63) as u64)))));
        CSM.w[0] = (((CS.w[0].wrapping_add(CS.w[0]))) | 1);
        M256 = __sqr128_to_256(CSM);
        if ((C4.w[3] > M256.w[3]) || (((C4.w[3] == M256.w[3]) && (((C4.w[2] > M256.w[2]) || (((C4.w[2] == M256.w[2]) && (((C4.w[1] > M256.w[1]) || (((C4.w[1] == M256.w[1]) && (C4.w[0] > M256.w[0])))))))))))) {
            CS.w[0] = CS.w[0].wrapping_add(1);
            if (CS.w[0] == 0) {
                CS.w[1] = CS.w[1].wrapping_add(1);
            }
        } else {
            C8.w[1] = (((go_checked_shl_u64(CS.w[1], go_shift_count_u64((3) as u64)))) | ((go_checked_shr_u64(CS.w[0], go_shift_count_u64((61) as u64)))));
            C8.w[0] = (go_checked_shl_u64(CS.w[0], go_shift_count_u64((3) as u64)));
            (M256.w[0], Carry) = __sub_borrow_out(M256.w[0], C8.w[0]);
            (M256.w[1], Carry) = __sub_borrow_in_out(M256.w[1], C8.w[1], Carry);
            (M256.w[2], Carry) = __sub_borrow_in_out(M256.w[2], 0, Carry);
            M256.w[3] = (M256.w[3].wrapping_sub(Carry));
            if ((M256.w[3] > C4.w[3]) || (((M256.w[3] == C4.w[3]) && (((M256.w[2] > C4.w[2]) || (((M256.w[2] == C4.w[2]) && (((M256.w[1] > C4.w[1]) || (((M256.w[1] == C4.w[1]) && (M256.w[0] > C4.w[0])))))))))))) {
                if (CS.w[0] == 0) {
                    CS.w[1] = CS.w[1].wrapping_sub(1);
                }
                CS.w[0] = CS.w[0].wrapping_sub(1);
            }
        }
    } else {
        M256 = __sqr128_to_256(CS);
        C8.w[1] = (((go_checked_shl_u64(CS.w[1], go_shift_count_u64((1) as u64)))) | ((go_checked_shr_u64(CS.w[0], go_shift_count_u64((63) as u64)))));
        C8.w[0] = (go_checked_shl_u64(CS.w[0], go_shift_count_u64((1) as u64)));
        if ((M256.w[3] > C256.w[3]) || (((M256.w[3] == C256.w[3]) && (((M256.w[2] > C256.w[2]) || (((M256.w[2] == C256.w[2]) && (((M256.w[1] > C256.w[1]) || (((M256.w[1] == C256.w[1]) && (M256.w[0] > C256.w[0])))))))))))) {
            (M256.w[0], Carry) = __sub_borrow_out(M256.w[0], C8.w[0]);
            (M256.w[1], Carry) = __sub_borrow_in_out(M256.w[1], C8.w[1], Carry);
            (M256.w[2], Carry) = __sub_borrow_in_out(M256.w[2], 0, Carry);
            M256.w[3] = (M256.w[3].wrapping_sub(Carry));
            M256.w[0] = M256.w[0].wrapping_add(1);
            if (M256.w[0] == 0) {
                M256.w[1] = M256.w[1].wrapping_add(1);
                if (M256.w[1] == 0) {
                    M256.w[2] = M256.w[2].wrapping_add(1);
                    if (M256.w[2] == 0) {
                        M256.w[3] = M256.w[3].wrapping_add(1);
                    }
                }
            }
            if (CS.w[0] == 0) {
                CS.w[1] = CS.w[1].wrapping_sub(1);
            }
            CS.w[0] = CS.w[0].wrapping_sub(1);
            if ((M256.w[3] > C256.w[3]) || (((M256.w[3] == C256.w[3]) && (((M256.w[2] > C256.w[2]) || (((M256.w[2] == C256.w[2]) && (((M256.w[1] > C256.w[1]) || (((M256.w[1] == C256.w[1]) && (M256.w[0] > C256.w[0])))))))))))) {
                if (CS.w[0] == 0) {
                    CS.w[1] = CS.w[1].wrapping_sub(1);
                }
                CS.w[0] = CS.w[0].wrapping_sub(1);
            }
        } else {
            (M256.w[0], Carry) = __add_carry_out(M256.w[0], C8.w[0]);
            (M256.w[1], Carry) = __add_carry_in_out(M256.w[1], C8.w[1], Carry);
            (M256.w[2], Carry) = __add_carry_in_out(M256.w[2], 0, Carry);
            M256.w[3] = (M256.w[3].wrapping_add(Carry));
            M256.w[0] = M256.w[0].wrapping_add(1);
            if (M256.w[0] == 0) {
                M256.w[1] = M256.w[1].wrapping_add(1);
                if (M256.w[1] == 0) {
                    M256.w[2] = M256.w[2].wrapping_add(1);
                    if (M256.w[2] == 0) {
                        M256.w[3] = M256.w[3].wrapping_add(1);
                    }
                }
            }
            if ((M256.w[3] < C256.w[3]) || (((M256.w[3] == C256.w[3]) && (((M256.w[2] < C256.w[2]) || (((M256.w[2] == C256.w[2]) && (((M256.w[1] < C256.w[1]) || (((M256.w[1] == C256.w[1]) && (M256.w[0] <= C256.w[0])))))))))))) {
                CS.w[0] = CS.w[0].wrapping_add(1);
                if (CS.w[0] == 0) {
                    CS.w[1] = CS.w[1].wrapping_add(1);
                }
            }
        }
        if (rnd_mode == 2) {
            CS.w[0] = CS.w[0].wrapping_add(1);
            if (CS.w[0] == 0) {
                CS.w[1] = CS.w[1].wrapping_add(1);
            }
        }
    }
    pfpsf |= 32;
    res = bid_get_bid128_fast(0, (go_checked_shr_i64(((exponent_q.wrapping_add(0x1820))), go_shift_count_u64((1) as u64))), CS);
    return (res, pfpsf);
}
