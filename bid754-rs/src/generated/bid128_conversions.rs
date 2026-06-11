// Auto-generated from bid128_conversions.go by go2rs. Do not edit.

use super::prelude::*;

pub fn bid128_to_bid64(mut x: BID_UINT128, mut rnd_mode: i64) -> (u64, u32) {
    let mut CX: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut T128: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut TP128: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut Qh: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut Ql: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut Qh1: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut Stemp: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut Tmp: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut Tmp1: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut CX1: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut sign_x: u64 = 0;
    let mut carry: u64 = 0;
    let mut cy: u64 = 0;
    let mut res: u64 = 0;
    let mut D: i64 = 0;
    let mut exponent_x: i64 = 0;
    let mut extra_digits: i64 = 0;
    let mut amount: i64 = 0;
    let mut bin_expon_cx: i64 = 0;
    let mut rmode: u32 = 0;
    let mut status: u32 = 0;
    let mut uf_check: u32 = 0;
    let mut pfpsf: u32 = 0;
    let (mut sign_x, mut exponent_x, mut CX, mut valid) = unpack_bid128_value(x);
    if (!valid) {
        if (((go_checked_shl_u64(x.w[1], go_shift_count_u64((1) as u64)))) >= 0xf000000000000000) {
            Tmp.w[1] = (CX.w[1] & 0x00003fffffffffff);
            Tmp.w[0] = CX.w[0];
            TP128 = bid_reciprocals10_128[18];
            (Qh, Ql) = __mul_128x128_full(Tmp, TP128);
            amount = (bid_recip_scale[18] as i64);
            Tmp = __shr_128(Qh, (amount as u64));
            res = ((CX.w[1] & 0xfc00000000000000) | Tmp.w[0]);
            if ((x.w[1] & 0x7e00000000000000) == 0x7e00000000000000) {
                pfpsf |= 1;
            }
            return (res, pfpsf);
        }
        exponent_x = ((exponent_x.wrapping_sub(0x1820)).wrapping_add(0x18e));
        if (exponent_x < 0) {
            res = sign_x;
            return (res, pfpsf);
        }
        if (exponent_x > 0x2ff) {
            exponent_x = 0x2ff;
        }
        res = (sign_x | ((go_checked_shl_u64((exponent_x as u64), go_shift_count_u64((53) as u64)))));
        return (res, pfpsf);
    }
    if ((CX.w[1] != 0) || (CX.w[0] >= 10000000000000000)) {
        let mut f64_i: u32 = (0x5f800000 as u32);
        let mut fx_d = no_fma_mul_add_f32((CX.w[1] as f32), f32::from_bits(f64_i), (CX.w[0] as f32));
        let mut fx_i = (fx_d as f32).to_bits();
        bin_expon_cx = (((((go_checked_shr_u32(fx_i, go_shift_count_u64((23) as u64)))) & 0xff) as i64).wrapping_sub(0x7f));
        extra_digits = ((bid_estimate_decimal_digits[bin_expon_cx as usize] as i64).wrapping_sub(16));
        D = ((CX.w[1] as i64).wrapping_sub(bid_power10_index_binexp_128[bin_expon_cx as usize].w[1] as i64));
        if ((D > 0) || (((D == 0) && (CX.w[0] >= bid_power10_index_binexp_128[bin_expon_cx as usize].w[0])))) {
            extra_digits = extra_digits.wrapping_add(1);
        }
        exponent_x = exponent_x.wrapping_add(extra_digits);
        rmode = (rnd_mode as u32);
        if ((sign_x != 0) && (((rmode.wrapping_sub(1)) as u32) < 2)) {
            rmode = ((3 as u32).wrapping_sub(rmode));
        }
        if (exponent_x < (0x1820 - 0x18e)) {
            uf_check = 1;
            if ((((((extra_digits.wrapping_neg()).wrapping_add(exponent_x)).wrapping_sub(0x1820)).wrapping_add(0x18e)).wrapping_add(35)) >= 0) {
                if (exponent_x == ((0x1820 - 0x18e) - 1)) {
                    T128 = bid_round_const_table_128[rmode as usize][extra_digits as usize];
                    (CX1.w[0], carry) = __add_carry_out(T128.w[0], CX.w[0]);
                    CX1.w[1] = ((CX.w[1].wrapping_add(T128.w[1])).wrapping_add(carry));
                }
                extra_digits = (((extra_digits.wrapping_add(0x1820)).wrapping_sub(0x18e)).wrapping_sub(exponent_x));
                exponent_x = (0x1820 - 0x18e);
            } else {
                rmode = 3;
            }
        }
        T128 = bid_round_const_table_128[rmode as usize][extra_digits as usize];
        (CX.w[0], carry) = __add_carry_out(T128.w[0], CX.w[0]);
        CX.w[1] = ((CX.w[1].wrapping_add(T128.w[1])).wrapping_add(carry));
        TP128 = bid_reciprocals10_128[extra_digits as usize];
        (Qh, Ql) = __mul_128x128_full(CX, TP128);
        amount = (bid_recip_scale[extra_digits as usize] as i64);
        if (amount >= 64) {
            CX.w[0] = (go_checked_shr_u64(Qh.w[1], go_shift_count_u64((((amount.wrapping_sub(64)) as u64)) as u64)));
            CX.w[1] = 0;
        } else {
            CX = __shr_128(Qh, (amount as u64));
        }
        if (rnd_mode == 0) {
            if ((CX.w[0] & 1) != 0) {
                Qh1 = __shl_128_long(Qh, (((128 as i64).wrapping_sub(amount)) as u64));
                if (((Qh1.w[1] == 0) && (Qh1.w[0] == 0)) && (((Ql.w[1] < bid_reciprocals10_128[extra_digits as usize].w[1]) || (((Ql.w[1] == bid_reciprocals10_128[extra_digits as usize].w[1]) && (Ql.w[0] < bid_reciprocals10_128[extra_digits as usize].w[0])))))) {
                    CX.w[0] = CX.w[0].wrapping_sub(1);
                }
            }
        }
        status = 32;
        Qh1 = __shl_128_long(Qh, (((128 as i64).wrapping_sub(amount)) as u64));
        match rmode {
            0 | 4 => {
                if (((Qh1.w[1] == 0x8000000000000000) && (Qh1.w[0] == 0)) && (((Ql.w[1] < bid_reciprocals10_128[extra_digits as usize].w[1]) || (((Ql.w[1] == bid_reciprocals10_128[extra_digits as usize].w[1]) && (Ql.w[0] < bid_reciprocals10_128[extra_digits as usize].w[0])))))) {
                    status = 0;
                }
            }
            1 | 3 => {
                if (((Qh1.w[1] == 0) && (Qh1.w[0] == 0)) && (((Ql.w[1] < bid_reciprocals10_128[extra_digits as usize].w[1]) || (((Ql.w[1] == bid_reciprocals10_128[extra_digits as usize].w[1]) && (Ql.w[0] < bid_reciprocals10_128[extra_digits as usize].w[0])))))) {
                    status = 0;
                }
            }
            _ => {
                (Stemp.w[0], cy) = __add_carry_out(Ql.w[0], bid_reciprocals10_128[extra_digits as usize].w[0]);
                (Stemp.w[1], carry) = __add_carry_in_out(Ql.w[1], bid_reciprocals10_128[extra_digits as usize].w[1], cy);
                Qh = __shr_128_long(Qh1, (((128 as i64).wrapping_sub(amount)) as u64));
                Tmp.w[0] = 1;
                Tmp.w[1] = 0;
                Tmp1 = __shl_128_long(Tmp, (amount as u64));
                Qh.w[0] = Qh.w[0].wrapping_add(carry);
                if (Qh.w[0] < carry) {
                    Qh.w[1] = Qh.w[1].wrapping_add(1);
                }
                if __unsigned_compare_ge_128(Qh, Tmp1) {
                    status = 0;
                }
            }
        }
        if (status != 0) {
            if (uf_check != 0) {
                status |= 16;
            }
            pfpsf |= status;
        }
        _ = CX1;
        _ = Stemp;
    }
    let (mut res, mut flags) = get_bid64_flags(sign_x, ((exponent_x.wrapping_sub(0x1820)).wrapping_add(0x18e)), CX.w[0], rnd_mode);
    pfpsf |= flags;
    return (res, pfpsf);
}

pub fn bid128_to_bid32(mut x: BID_UINT128, mut rnd_mode: i64) -> (u32, u32) {
    let mut CX: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut T128: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut TP128: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut Qh: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut Ql: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut Qh1: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut Stemp: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut Tmp: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut Tmp1: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut CX1: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut sign_x: u64 = 0;
    let mut carry: u64 = 0;
    let mut cy: u64 = 0;
    let mut D: i64 = 0;
    let mut res: u32 = 0;
    let mut exponent_x: i64 = 0;
    let mut extra_digits: i64 = 0;
    let mut amount: i64 = 0;
    let mut bin_expon_cx: i64 = 0;
    let mut uf_check: i64 = 0;
    let mut rmode: u32 = 0;
    let mut status: u32 = 0;
    let mut pfpsf: u32 = 0;
    let (mut sign_x, mut exponent_x, mut CX, mut valid) = unpack_bid128_value(x);
    if (!valid) {
        if ((x.w[1] & 0x7800000000000000) == 0x7800000000000000) {
            Tmp.w[1] = (CX.w[1] & 0x00003fffffffffff);
            Tmp.w[0] = CX.w[0];
            TP128 = bid_reciprocals10_128[27];
            (Qh, Ql) = __mul_128x128_full(Tmp, TP128);
            amount = ((bid_recip_scale[27] as i64).wrapping_sub(64));
            res = (((((go_checked_shr_u64(CX.w[1], go_shift_count_u64((32) as u64)))) & 0xfc000000) as u32) | ((go_checked_shr_u64(Qh.w[1], go_shift_count_u64((amount as u64) as u64))) as u32));
            if ((x.w[1] & 0x7e00000000000000) == 0x7e00000000000000) {
                pfpsf |= 1;
            }
            _ = Ql;
            return (res, pfpsf);
        }
        exponent_x = ((exponent_x.wrapping_sub(0x1820)).wrapping_add(101));
        if (exponent_x < 0) {
            exponent_x = 0;
        }
        if (exponent_x > 191) {
            exponent_x = 191;
        }
        res = (((go_checked_shr_u64(sign_x, go_shift_count_u64((32) as u64))) as u32) | ((go_checked_shl_i64(exponent_x, go_shift_count_u64((23) as u64))) as u32));
        return (res, pfpsf);
    }
    if ((CX.w[1] != 0) || (CX.w[0] >= 10000000)) {
        let mut f64_i: u32 = (0x5f800000 as u32);
        let mut fx_d = no_fma_mul_add_f32((CX.w[1] as f32), f32::from_bits(f64_i), (CX.w[0] as f32));
        let mut fx_i = (fx_d as f32).to_bits();
        bin_expon_cx = (((((go_checked_shr_u32(fx_i, go_shift_count_u64((23) as u64)))) & 0xff) as i64).wrapping_sub(0x7f));
        extra_digits = ((bid_estimate_decimal_digits[bin_expon_cx as usize] as i64).wrapping_sub(7));
        D = ((CX.w[1] as i64).wrapping_sub(bid_power10_index_binexp_128[bin_expon_cx as usize].w[1] as i64));
        if ((D > 0) || (((D == 0) && (CX.w[0] >= bid_power10_index_binexp_128[bin_expon_cx as usize].w[0])))) {
            extra_digits = extra_digits.wrapping_add(1);
        }
        exponent_x = exponent_x.wrapping_add(extra_digits);
        rmode = (rnd_mode as u32);
        if ((sign_x != 0) && (((rmode.wrapping_sub(1)) as u32) < 2)) {
            rmode = ((3 as u32).wrapping_sub(rmode));
        }
        if (exponent_x < (0x1820 - 101)) {
            uf_check = 1;
            if ((((((extra_digits.wrapping_neg()).wrapping_add(exponent_x)).wrapping_sub(0x1820)).wrapping_add(101)).wrapping_add(35)) >= 0) {
                if (exponent_x == ((0x1820 - 101) - 1)) {
                    T128 = bid_round_const_table_128[rmode as usize][extra_digits as usize];
                    (CX1.w[0], carry) = __add_carry_out(T128.w[0], CX.w[0]);
                    CX1.w[1] = ((CX.w[1].wrapping_add(T128.w[1])).wrapping_add(carry));
                }
                extra_digits = (((extra_digits.wrapping_add(0x1820)).wrapping_sub(101)).wrapping_sub(exponent_x));
                exponent_x = (0x1820 - 101);
            } else {
                rmode = 3;
            }
        }
        T128 = bid_round_const_table_128[rmode as usize][extra_digits as usize];
        (CX.w[0], carry) = __add_carry_out(T128.w[0], CX.w[0]);
        CX.w[1] = ((CX.w[1].wrapping_add(T128.w[1])).wrapping_add(carry));
        TP128 = bid_reciprocals10_128[extra_digits as usize];
        (Qh, Ql) = __mul_128x128_full(CX, TP128);
        amount = (bid_recip_scale[extra_digits as usize] as i64);
        if (amount >= 64) {
            CX.w[0] = (go_checked_shr_u64(Qh.w[1], go_shift_count_u64((((amount.wrapping_sub(64)) as u64)) as u64)));
            CX.w[1] = 0;
        } else {
            CX = __shr_128(Qh, (amount as u64));
        }
        if (rnd_mode == 0) {
            if ((CX.w[0] & 1) != 0) {
                Qh1 = __shl_128_long(Qh, (((128 as i64).wrapping_sub(amount)) as u64));
                if (((Qh1.w[1] == 0) && (Qh1.w[0] == 0)) && (((Ql.w[1] < bid_reciprocals10_128[extra_digits as usize].w[1]) || (((Ql.w[1] == bid_reciprocals10_128[extra_digits as usize].w[1]) && (Ql.w[0] < bid_reciprocals10_128[extra_digits as usize].w[0])))))) {
                    CX.w[0] = CX.w[0].wrapping_sub(1);
                }
            }
        }
        status = 32;
        Qh1 = __shl_128_long(Qh, (((128 as i64).wrapping_sub(amount)) as u64));
        match rmode {
            0 | 4 => {
                if (((Qh1.w[1] == 0x8000000000000000) && (Qh1.w[0] == 0)) && (((Ql.w[1] < bid_reciprocals10_128[extra_digits as usize].w[1]) || (((Ql.w[1] == bid_reciprocals10_128[extra_digits as usize].w[1]) && (Ql.w[0] < bid_reciprocals10_128[extra_digits as usize].w[0])))))) {
                    status = 0;
                }
            }
            1 | 3 => {
                if (((Qh1.w[1] == 0) && (Qh1.w[0] == 0)) && (((Ql.w[1] < bid_reciprocals10_128[extra_digits as usize].w[1]) || (((Ql.w[1] == bid_reciprocals10_128[extra_digits as usize].w[1]) && (Ql.w[0] < bid_reciprocals10_128[extra_digits as usize].w[0])))))) {
                    status = 0;
                }
            }
            _ => {
                (Stemp.w[0], cy) = __add_carry_out(Ql.w[0], bid_reciprocals10_128[extra_digits as usize].w[0]);
                (Stemp.w[1], carry) = __add_carry_in_out(Ql.w[1], bid_reciprocals10_128[extra_digits as usize].w[1], cy);
                Qh = __shr_128_long(Qh1, (((128 as i64).wrapping_sub(amount)) as u64));
                Tmp.w[0] = 1;
                Tmp.w[1] = 0;
                Tmp1 = __shl_128_long(Tmp, (amount as u64));
                Qh.w[0] = Qh.w[0].wrapping_add(carry);
                if (Qh.w[0] < carry) {
                    Qh.w[1] = Qh.w[1].wrapping_add(1);
                }
                if __unsigned_compare_ge_128(Qh, Tmp1) {
                    status = 0;
                }
            }
        }
        if (status != 0) {
            if (uf_check != 0) {
                status |= 16;
            }
            pfpsf |= status;
        }
        _ = CX1;
        _ = Stemp;
    }
    res = get_bid32_flags(((go_checked_shr_u64(sign_x, go_shift_count_u64((32) as u64))) as u32), ((exponent_x.wrapping_sub(0x1820)).wrapping_add(101)), CX.w[0], rnd_mode, (&mut pfpsf));
    return (res, pfpsf);
}
