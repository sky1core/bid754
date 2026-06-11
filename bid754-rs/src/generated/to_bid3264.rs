// Auto-generated from to_bid3264.go by go2rs. Do not edit.

use super::prelude::*;

pub(crate) fn bid32_get_no_flags(mut sgn: u32, mut expon: i64, mut coeff: u64, mut rmode: i64) -> u32 {
    if (coeff > 9999999) {
        expon = expon.wrapping_add(1);
        coeff = 1000000;
    }
    if ((expon as u64) > 191) {
        if (expon < 0) {
            if ((expon.wrapping_add(7)) < 0) {
                if ((rmode == 1) && (sgn != 0)) {
                    return 0x80000001;
                }
                if ((rmode == 2) && (sgn == 0)) {
                    return 0x00000001;
                }
                return sgn;
            }
            if ((sgn != 0) && (((rmode.wrapping_sub(1)) as u64) < 2)) {
                rmode = ((3 as i64).wrapping_sub(rmode));
            }
            let mut extraDigits = (expon.wrapping_neg());
            coeff = coeff.wrapping_add(bid_round_const_table[rmode as usize][extraDigits as usize]);
            let mut Q = __mul_64x64_to_128(coeff, bid_reciprocals10_64[extraDigits as usize]);
            let mut amount = (bid_short_recip_scale[extraDigits as usize] as i64);
            coeff = (go_checked_shr_u64(Q.w[1], go_shift_count_i64((amount) as i64)));
            if ((rmode == 0) && ((coeff & 1) != 0)) {
                let mut remainder_h = (Q.w[1] & ((((go_checked_shl_u64((1 as u64), go_shift_count_i64((amount) as i64)))).wrapping_sub(1))));
                if ((remainder_h == 0) && (Q.w[0] < bid_reciprocals10_64[extraDigits as usize])) {
                    coeff = coeff.wrapping_sub(1);
                }
            }
            return (sgn | (coeff as u32));
        }
        if ((coeff == 0) && (expon > 191)) {
            expon = 191;
        }
        while ((coeff < 1000000) && (expon >= 192)) {
            expon = expon.wrapping_sub(1);
            coeff = (((go_checked_shl_u64(coeff, go_shift_count_u64((3) as u64)))).wrapping_add(((go_checked_shl_u64(coeff, go_shift_count_u64((1) as u64))))));
        }
        if (expon > 191) {
            let mut r = (sgn | 0x78000000);
            match rmode {
                1 => {
                    if (sgn == 0) {
                        r = 0x77f8967f;
                    }
                }
                3 => {
                    r = (sgn | 0x77f8967f);
                }
                2 => {
                    if (sgn != 0) {
                        r = 0xf7f8967f;
                    }
                }
                _ => {}
            }
            return r;
        }
    }
    return very_fast_get_bid32(sgn, expon, (coeff as u32));
}

pub(crate) fn bid32_get_with_flags(mut sgn: u32, mut expon: i64, mut coeff: u64, mut rmode: i64, mut flags: u32) -> (u32, u32) {
    let mut Q: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut C64: u64 = 0;
    let mut remainder_h: u64 = 0;
    let mut carry: u64 = 0;
    let mut Stemp: u64 = 0;
    let mut r: u32 = 0;
    let mut mask: u32 = 0;
    let mut status: u32 = 0;
    let mut extra_digits: i64 = 0;
    let mut amount: i64 = 0;
    let mut amount2: i64 = 0;
    if (coeff > 9999999) {
        expon = expon.wrapping_add(1);
        coeff = 1000000;
    }
    if ((expon as u64) > 191) {
        if (expon < 0) {
            if ((expon.wrapping_add(7)) < 0) {
                flags |= (16 | 32);
                if ((rmode == 1) && (sgn != 0)) {
                    return (0x80000001, flags);
                }
                if ((rmode == 2) && (sgn == 0)) {
                    return (1, flags);
                }
                return (sgn, flags);
            }
            if ((sgn != 0) && (((rmode.wrapping_sub(1)) as u64) < 2)) {
                rmode = ((3 as i64).wrapping_sub(rmode));
            }
            extra_digits = (expon.wrapping_neg());
            coeff = coeff.wrapping_add(bid_round_const_table[rmode as usize][extra_digits as usize]);
            Q = __mul_64x64_to_128(coeff, bid_reciprocals10_64[extra_digits as usize]);
            amount = (bid_short_recip_scale[extra_digits as usize] as i64);
            C64 = (go_checked_shr_u64(Q.w[1], go_shift_count_u64((amount as u64) as u64)));
            if ((rmode == 0) && ((C64 & 1) != 0)) {
                amount2 = ((64 as i64).wrapping_sub(amount));
                remainder_h = (!(0 as u64));
                remainder_h = go_checked_shr_u64(remainder_h, go_shift_count_u64((amount2 as u64) as u64));
                remainder_h &= Q.w[1];
                if ((remainder_h == 0) && (Q.w[0] < bid_reciprocals10_64[extra_digits as usize])) {
                    C64 = C64.wrapping_sub(1);
                }
            }
            if ((flags & 32) != 0) {
                flags |= 16;
            } else {
                status = 32;
                remainder_h = (go_checked_shl_u64(Q.w[1], go_shift_count_u64(((((64 as i64).wrapping_sub(amount)) as u64)) as u64)));
                match rmode {
                    0 | 4 => {
                        if ((remainder_h == 0x8000000000000000) && (Q.w[0] < bid_reciprocals10_64[extra_digits as usize])) {
                            status = 0;
                        }
                    }
                    1 | 3 => {
                        if ((remainder_h == 0) && (Q.w[0] < bid_reciprocals10_64[extra_digits as usize])) {
                            status = 0;
                        }
                    }
                    _ => {
                        (Stemp, carry) = __add_carry_out(Q.w[0], bid_reciprocals10_64[extra_digits as usize]);
                        _ = Stemp;
                        if ((((go_checked_shr_u64(remainder_h, go_shift_count_u64(((((64 as i64).wrapping_sub(amount)) as u64)) as u64)))).wrapping_add(carry)) >= ((go_checked_shl_u64((1 as u64), go_shift_count_u64((amount as u64) as u64))))) {
                            status = 0;
                        }
                    }
                }
                if (status != 0) {
                    flags |= (16 | status);
                }
            }
            return ((sgn | (C64 as u32)), flags);
        }
        if ((coeff == 0) && (expon > 191)) {
            expon = 191;
        }
        while ((coeff < 1000000) && (expon > 191)) {
            coeff = (((go_checked_shl_u64(coeff, go_shift_count_u64((3) as u64)))).wrapping_add(((go_checked_shl_u64(coeff, go_shift_count_u64((1) as u64))))));
            expon = expon.wrapping_sub(1);
        }
        if ((expon as u64) > 191) {
            flags |= (8 | 32);
            r = (sgn | 0x78000000);
            match rmode {
                1 => {
                    if (sgn == 0) {
                        r = 0x77f8967f;
                    }
                }
                3 => {
                    r = (sgn | 0x77f8967f);
                }
                2 => {
                    if (sgn != 0) {
                        r = (sgn | 0x77f8967f);
                    }
                }
                _ => {}
            }
            return (r, flags);
        }
    }
    mask = (1 << 23);
    if (coeff < (mask as u64)) {
        r = (expon as u32);
        r = go_checked_shl_u32(r, go_shift_count_u64((23) as u64));
        r |= (((coeff as u32) | sgn));
        return (r, flags);
    }
    r = (expon as u32);
    r = go_checked_shl_u32(r, go_shift_count_u64((21) as u64));
    r |= (sgn | 0x60000000);
    mask = ((1 << 21) - 1);
    r |= (((coeff as u32) & mask));
    return (r, flags);
}

pub fn bid64_to_bid32(mut x: u64, mut rndMode: i64) -> (u32, u32) {
    let mut Q: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut sign_x: u64 = 0;
    let mut coefficient_x: u64 = 0;
    let mut remainder_h: u64 = 0;
    let mut carry: u64 = 0;
    let mut Stemp: u64 = 0;
    let mut res: u32 = 0;
    let mut t64: u64 = 0;
    let mut exponent_x: i64 = 0;
    let mut bin_expon_cx: i64 = 0;
    let mut extra_digits: i64 = 0;
    let mut rmode: i64 = 0;
    let mut amount: i64 = 0;
    let mut status: u32 = 0;
    let (mut sign_x, mut exponent_x, mut coefficient_x, mut valid) = unpack_bid64(x);
    if (!valid) {
        if ((x & 0x7800000000000000) == 0x7800000000000000) {
            t64 = (coefficient_x & 0x0003ffffffffffff);
            res = ((t64 / 1000000000) as u32);
            res |= ((((go_checked_shr_u64(coefficient_x, go_shift_count_u64((32) as u64)))) & 0xfc000000) as u32);
            if ((x & 0x7e00000000000000) == 0x7e00000000000000) {
                status |= 1;
            }
            return (res, status);
        }
        exponent_x = ((exponent_x.wrapping_sub(0x18e)).wrapping_add(101));
        if (exponent_x < 0) {
            exponent_x = 0;
        }
        if (exponent_x > 191) {
            exponent_x = 191;
        }
        res = (((go_checked_shr_u64(sign_x, go_shift_count_u64((32) as u64))) as u32) | ((go_checked_shl_i64(exponent_x, go_shift_count_u64((23) as u64))) as u32));
        return (res, status);
    }
    exponent_x = ((exponent_x.wrapping_sub(0x18e)).wrapping_add(101));
    if (coefficient_x >= 10000000) {
        let mut tempx = ((coefficient_x as f32) as f32).to_bits();
        bin_expon_cx = (((((go_checked_shr_u32(tempx, go_shift_count_u64((23) as u64)))) & 0xff) as i64).wrapping_sub(0x7f));
        extra_digits = ((bid_estimate_decimal_digits[bin_expon_cx as usize] as i64).wrapping_sub(7));
        if (coefficient_x >= bid_power10_index_binexp[bin_expon_cx as usize]) {
            extra_digits = extra_digits.wrapping_add(1);
        }
        rmode = rndMode;
        if ((sign_x != 0) && (((rmode.wrapping_sub(1)) as u64) < 2)) {
            rmode = ((3 as i64).wrapping_sub(rmode));
        }
        exponent_x = exponent_x.wrapping_add(extra_digits);
        if ((exponent_x < 0) && (((exponent_x.wrapping_add(7)) >= 0))) {
            status = 16;
            extra_digits = extra_digits.wrapping_sub(exponent_x);
            exponent_x = 0;
        }
        coefficient_x = coefficient_x.wrapping_add(bid_round_const_table[rmode as usize][extra_digits as usize]);
        Q = __mul_64x64_to_128(coefficient_x, bid_reciprocals10_64[extra_digits as usize]);
        amount = (bid_short_recip_scale[extra_digits as usize] as i64);
        coefficient_x = (go_checked_shr_u64(Q.w[1], go_shift_count_i64((amount) as i64)));
        if (rmode == 0) {
            if ((coefficient_x & 1) != 0) {
                remainder_h = (go_checked_shl_u64(Q.w[1], go_shift_count_i64(((((64 as i64).wrapping_sub(amount)))) as i64)));
                if ((remainder_h == 0) && (Q.w[0] < bid_reciprocals10_64[extra_digits as usize])) {
                    coefficient_x = coefficient_x.wrapping_sub(1);
                }
            }
        }
        status |= 32;
        remainder_h = (go_checked_shl_u64(Q.w[1], go_shift_count_i64(((((64 as i64).wrapping_sub(amount)))) as i64)));
        match rmode {
            0 | 4 => {
                if ((remainder_h == 0x8000000000000000) && (Q.w[0] < bid_reciprocals10_64[extra_digits as usize])) {
                    status = 0;
                }
            }
            1 | 3 => {
                if ((remainder_h == 0) && (Q.w[0] < bid_reciprocals10_64[extra_digits as usize])) {
                    status = 0;
                }
            }
            _ => {
                (Stemp, carry) = __add_carry_out(Q.w[0], bid_reciprocals10_64[extra_digits as usize]);
                _ = Stemp;
                if ((((go_checked_shr_u64(remainder_h, go_shift_count_u64(((((64 as i64).wrapping_sub(amount)) as u64)) as u64)))).wrapping_add(carry)) >= ((go_checked_shl_u64((1 as u64), go_shift_count_u64((amount as u64) as u64))))) {
                    status = 0;
                }
            }
        }
    }
    (res, status) = bid32_get_with_flags(((go_checked_shr_u64(sign_x, go_shift_count_u64((32) as u64))) as u32), exponent_x, coefficient_x, rndMode, status);
    return (res, status);
}
