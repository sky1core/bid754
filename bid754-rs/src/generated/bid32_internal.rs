// Auto-generated from bid32_internal.go by go2rs. Do not edit.

use super::prelude::*;

pub(crate) fn very_fast_get_bid32(mut sgn: u32, mut expon: i64, mut coeff: u32) -> u32 {
    let mut r: u32 = 0;
    let mut mask: u32 = 0;
    mask = (1 << 23);
    if (coeff < mask) {
        r = (expon as u32);
        r = go_checked_shl_u32(r, go_shift_count_u64((23) as u64));
        r |= (coeff | sgn);
        return r;
    }
    r = (expon as u32);
    r = go_checked_shl_u32(r, go_shift_count_u64((21) as u64));
    r |= (sgn | 0x60000000);
    mask = ((1 << 21) - 1);
    coeff &= mask;
    r |= coeff;
    return r;
}

pub(crate) fn fast_get_bid32(mut sgn: u32, mut expon: i64, mut coeff: u32) -> u32 {
    let mut r: u32 = 0;
    let mut mask: u32 = 0;
    mask = (1 << 23);
    if (coeff > 9999999) {
        expon = expon.wrapping_add(1);
        coeff = 1000000;
    }
    if (coeff < mask) {
        r = (expon as u32);
        r = go_checked_shl_u32(r, go_shift_count_u64((23) as u64));
        r |= (coeff | sgn);
        return r;
    }
    r = (expon as u32);
    r = go_checked_shl_u32(r, go_shift_count_u64((21) as u64));
    r |= (sgn | 0x60000000);
    mask = ((1 << 21) - 1);
    coeff &= mask;
    r |= coeff;
    return r;
}

pub(crate) fn get_bid32(mut sgn: u32, mut expon: i64, mut coeff: u64, mut rmode: i64) -> u32 {
    let mut Q: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut _C64: u64 = 0;
    let mut remainder_h: u64 = 0;
    let mut r: u32 = 0;
    let mut mask: u32 = 0;
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
                if ((rmode == 1) && (sgn != 0)) {
                    return 0x80000001;
                }
                if ((rmode == 2) && (sgn == 0)) {
                    return 1;
                }
                return sgn;
            }
            if ((sgn != 0) && (((rmode.wrapping_sub(1)) as u64) < 2)) {
                rmode = ((3 as i64).wrapping_sub(rmode));
            }
            extra_digits = (expon.wrapping_neg());
            coeff = coeff.wrapping_add(bid_round_const_table[rmode as usize][extra_digits as usize]);
            Q = __mul_64x64_to_128(coeff, bid_reciprocals10_64[extra_digits as usize]);
            amount = (bid_short_recip_scale[extra_digits as usize] as i64);
            _C64 = (go_checked_shr_u64(Q.w[1], go_shift_count_u64((amount as u64) as u64)));
            if (rmode == 0) {
                if ((_C64 & 1) != 0) {
                    amount2 = ((64 as i64).wrapping_sub(amount));
                    remainder_h = (go_checked_shr_u64(((!(0 as u64))), go_shift_count_u64((amount2 as u64) as u64)));
                    remainder_h = (remainder_h & Q.w[1]);
                    if ((remainder_h == 0) && (Q.w[0] < bid_reciprocals10_64[extra_digits as usize])) {
                        _C64 = _C64.wrapping_sub(1);
                    }
                }
            }
            return (sgn | (_C64 as u32));
        }
        if (coeff == 0) {
            if (expon > 191) {
                expon = 191;
            }
        }
        while ((coeff < 1000000) && (expon >= (3 * 64))) {
            expon = expon.wrapping_sub(1);
            coeff = (((go_checked_shl_u64(coeff, go_shift_count_u64((3) as u64)))).wrapping_add(((go_checked_shl_u64(coeff, go_shift_count_u64((1) as u64))))));
        }
        if (expon > 191) {
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
                        r = (0x80000000 | 0x77f8967f);
                    }
                }
                _ => {}
            }
            return r;
        }
    }
    mask = (1 << 23);
    if ((coeff as u32) < mask) {
        r = (expon as u32);
        r = go_checked_shl_u32(r, go_shift_count_u64((23) as u64));
        r |= (((coeff as u32) | sgn));
        return r;
    }
    r = (expon as u32);
    r = go_checked_shl_u32(r, go_shift_count_u64((21) as u64));
    r |= (sgn | 0x60000000);
    mask = ((1 << 21) - 1);
    let mut coeff64 = (coeff as u32);
    coeff64 &= mask;
    r |= coeff64;
    return r;
}

pub(crate) fn get_bid32_flags(mut sgn: u32, mut expon: i64, mut coeff: u64, mut rmode: i64, pfpsf: &mut u32) -> u32 {
    let mut Q: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut _C64: u64 = 0;
    let mut remainder_h: u64 = 0;
    let mut carry: u64 = 0;
    let mut Stemp: u64 = 0;
    let mut r: u32 = 0;
    let mut mask: u32 = 0;
    let mut extra_digits: i64 = 0;
    let mut amount: i64 = 0;
    let mut amount2: i64 = 0;
    let mut status: u32 = 0;
    if (coeff > 9999999) {
        expon = expon.wrapping_add(1);
        coeff = 1000000;
    }
    if ((expon as u64) > 191) {
        if (expon < 0) {
            if ((expon.wrapping_add(7)) < 0) {
                (*pfpsf) |= (16 | 32);
                if ((rmode == 1) && (sgn != 0)) {
                    return 0x80000001;
                }
                if ((rmode == 2) && (sgn == 0)) {
                    return 1;
                }
                return sgn;
            }
            if ((sgn != 0) && (((rmode.wrapping_sub(1)) as u64) < 2)) {
                rmode = ((3 as i64).wrapping_sub(rmode));
            }
            extra_digits = (expon.wrapping_neg());
            coeff = coeff.wrapping_add(bid_round_const_table[rmode as usize][extra_digits as usize]);
            Q = __mul_64x64_to_128(coeff, bid_reciprocals10_64[extra_digits as usize]);
            amount = (bid_short_recip_scale[extra_digits as usize] as i64);
            _C64 = (go_checked_shr_u64(Q.w[1], go_shift_count_u64((amount as u64) as u64)));
            if ((rmode == 0) && ((_C64 & 1) != 0)) {
                amount2 = ((64 as i64).wrapping_sub(amount));
                remainder_h = (!(0 as u64));
                remainder_h = go_checked_shr_u64(remainder_h, go_shift_count_u64((amount2 as u64) as u64));
                remainder_h &= Q.w[1];
                if ((remainder_h == 0) && (Q.w[0] < bid_reciprocals10_64[extra_digits as usize])) {
                    _C64 = _C64.wrapping_sub(1);
                }
            }
            if ((((*pfpsf) & 32)) != 0) {
                (*pfpsf) |= 16;
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
                        (Stemp, carry) = go_add64(Q.w[0], bid_reciprocals10_64[extra_digits as usize], 0);
                        _ = Stemp;
                        if ((((go_checked_shr_u64(remainder_h, go_shift_count_u64(((((64 as i64).wrapping_sub(amount)) as u64)) as u64)))).wrapping_add(carry)) >= ((go_checked_shl_u64((1 as u64), go_shift_count_u64((amount as u64) as u64))))) {
                            status = 0;
                        }
                    }
                }
                if (status != 0) {
                    (*pfpsf) |= (16 | status);
                }
            }
            return (sgn | (_C64 as u32));
        }
        if (coeff == 0) {
            if (expon > 191) {
                expon = 191;
            }
        }
        while ((coeff < 1000000) && (expon > 191)) {
            coeff = (((go_checked_shl_u64(coeff, go_shift_count_u64((3) as u64)))).wrapping_add(((go_checked_shl_u64(coeff, go_shift_count_u64((1) as u64))))));
            expon = expon.wrapping_sub(1);
        }
        if ((expon as u64) > 191) {
            (*pfpsf) |= (8 | 32);
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
            return r;
        }
    }
    mask = (1 << 23);
    if ((coeff as u32) < mask) {
        r = (expon as u32);
        r = go_checked_shl_u32(r, go_shift_count_u64((23) as u64));
        r |= (((coeff as u32) | sgn));
        return r;
    }
    r = (expon as u32);
    r = go_checked_shl_u32(r, go_shift_count_u64((21) as u64));
    r |= (sgn | 0x60000000);
    mask = ((1 << 21) - 1);
    r |= (((coeff as u32) & mask));
    return r;
}

pub(crate) fn get_bid32_uf(mut sgn: u32, mut expon: i64, mut coeff: u64, mut R: u32, mut rmode: i64, pfpsf: &mut u32) -> u32 {
    let mut Q: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut _C64: u64 = 0;
    let mut remainder_h: u64 = 0;
    let mut r: u32 = 0;
    let mut mask: u32 = 0;
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
                (*pfpsf) |= (16 | 32);
                if ((rmode == 1) && (sgn != 0)) {
                    return 0x80000001;
                }
                if ((rmode == 2) && (sgn == 0)) {
                    return 1;
                }
                return sgn;
            }
            if ((sgn != 0) && (((rmode.wrapping_sub(1)) as u64) < 2)) {
                rmode = ((3 as i64).wrapping_sub(rmode));
            }
            coeff = (((go_checked_shl_u64(coeff, go_shift_count_u64((3) as u64)))).wrapping_add(((go_checked_shl_u64(coeff, go_shift_count_u64((1) as u64))))));
            if (R != 0) {
                coeff |= 1;
            }
            extra_digits = ((1 as i64).wrapping_sub(expon));
            coeff = coeff.wrapping_add(bid_round_const_table[rmode as usize][extra_digits as usize]);
            Q = __mul_64x64_to_128(coeff, bid_reciprocals10_64[extra_digits as usize]);
            amount = (bid_short_recip_scale[extra_digits as usize] as i64);
            _C64 = (go_checked_shr_u64(Q.w[1], go_shift_count_u64((amount as u64) as u64)));
            if (rmode == 0) {
                if ((_C64 & 1) != 0) {
                    amount2 = ((64 as i64).wrapping_sub(amount));
                    remainder_h = (go_checked_shr_u64(((!(0 as u64))), go_shift_count_u64((amount2 as u64) as u64)));
                    remainder_h = (remainder_h & Q.w[1]);
                    if ((remainder_h == 0) && (Q.w[0] < bid_reciprocals10_64[extra_digits as usize])) {
                        _C64 = _C64.wrapping_sub(1);
                    }
                }
            }
            if ((((*pfpsf) & 32)) != 0) {
                (*pfpsf) |= 16;
            } else {
                let mut status: u32 = (32 as u32);
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
                        let mut carry: u64 = 0;
                        (_, carry) = __add_carry_out(Q.w[0], bid_reciprocals10_64[extra_digits as usize]);
                        if ((((go_checked_shr_u64(remainder_h, go_shift_count_u64(((((64 as i64).wrapping_sub(amount)) as u64)) as u64)))).wrapping_add(carry)) >= ((go_checked_shl_u64((1 as u64), go_shift_count_u64((amount as u64) as u64))))) {
                            status = 0;
                        }
                    }
                }
                if (status != 0) {
                    (*pfpsf) |= (16 | status);
                }
            }
            return (sgn | (_C64 as u32));
        }
        if (coeff == 0) {
            if (expon > 191) {
                expon = 191;
            }
        }
        while ((coeff < 1000000) && (expon >= (3 * 64))) {
            expon = expon.wrapping_sub(1);
            coeff = (((go_checked_shl_u64(coeff, go_shift_count_u64((3) as u64)))).wrapping_add(((go_checked_shl_u64(coeff, go_shift_count_u64((1) as u64))))));
        }
        if (expon > 191) {
            (*pfpsf) |= (8 | 32);
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
                        r = (0x80000000 | 0x77f8967f);
                    }
                }
                _ => {}
            }
            return r;
        }
    }
    mask = (1 << 23);
    if ((coeff as u32) < mask) {
        r = (expon as u32);
        r = go_checked_shl_u32(r, go_shift_count_u64((23) as u64));
        r |= (((coeff as u32) | sgn));
        return r;
    }
    r = (expon as u32);
    r = go_checked_shl_u32(r, go_shift_count_u64((21) as u64));
    r |= (sgn | 0x60000000);
    mask = ((1 << 21) - 1);
    let mut coeff64 = (coeff as u32);
    coeff64 &= mask;
    r |= coeff64;
    return r;
}

pub(crate) fn unpack_bid32_intel(mut x: u32) -> (u32, i64, u32, bool) {
    let mut sign: u32 = 0;
    let mut exponent: i64 = 0;
    let mut coefficient: u32 = 0;
    let mut valid: bool = false;
    sign = (x & 0x80000000);
    if ((x & 0x60000000) == 0x60000000) {
        coefficient = ((x & 0x1fffff) | 0x800000);
        if ((x & 0x78000000) == 0x78000000) {
            exponent = 0;
            coefficient = (x & 0xfe0fffff);
            if ((x & 0x000fffff) >= 1000000) {
                coefficient = (x & 0xfe000000);
            }
            if ((x & 0x7c000000) == 0x78000000) {
                coefficient = (x & 0xf8000000);
            }
            return (sign, exponent, coefficient, false);
        }
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
