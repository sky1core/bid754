// Auto-generated from bid64_from_string.go by go2rs. Do not edit.

use super::prelude::*;

pub(crate) fn tolower_macro(mut x: u8) -> u8 {
    if ((x >= b'A') && (x <= b'Z')) {
        return (x.wrapping_add(b'a' - b'A'));
    }
    return x;
}

pub(crate) fn bid64_from_string(str: impl AsRef<str>, mut rnd_mode: i64) -> (u64, u32) {
    let mut str = str.as_ref().to_string();
    let mut res: u64 = 0;
    let mut pfpsf: u32 = 0;
    let ps = str.as_bytes();
    let mut ps_idx: usize = 0;
    macro_rules! ps_at {
        ($offset:expr) => {
            *ps.get(ps_idx + ($offset as usize)).unwrap_or(&0)
        };
    }
    let mut sign_x: u64 = 0;
    let mut coefficient_x: u64 = 0;
    let mut rounded: u64 = 0;
    let mut expon_x: i64 = 0;
    let mut sgn_expon: i64 = 0;
    let mut ndigits: i64 = 0;
    let mut add_expon: i64 = 0;
    let mut midpoint: i64 = 0;
    let mut rounded_up: i64 = 0;
    let mut dround: i64 = 0;
    let mut dec_expon_scale: i64 = 0;
    let mut right_radix_leading_zeros: i64 = 0;
    let mut rdx_pt_enc: i64 = 0;
    let mut c: u8 = 0;
    while ((((ps_at!(0) == b' ') || (ps_at!(0) == b'\t'))) && (ps_at!(0) != 0)) {
        ps_idx += 1;
    }
    c = ps_at!(0);
    if ((c == 0) || (((((c != b'.') && (c != b'-')) && (c != b'+')) && (((c < b'0') || (c > b'9')))))) {
        if ((((tolower_macro(ps_at!(0)) == b'i') && (tolower_macro(ps_at!(1)) == b'n')) && (tolower_macro(ps_at!(2)) == b'f')) && (((ps_at!(3) == 0) || (((((((tolower_macro(ps_at!(3)) == b'i') && (tolower_macro(ps_at!(4)) == b'n')) && (tolower_macro(ps_at!(5)) == b'i')) && (tolower_macro(ps_at!(6)) == b't')) && (tolower_macro(ps_at!(7)) == b'y')) && (ps_at!(8) == 0)))))) {
            res = 0x7800000000000000;
            return (res, pfpsf);
        }
        if ((((tolower_macro(ps_at!(0)) == b's') && (tolower_macro(ps_at!(1)) == b'n')) && (tolower_macro(ps_at!(2)) == b'a')) && (tolower_macro(ps_at!(3)) == b'n')) {
            res = 0x7e00000000000000;
            return (res, pfpsf);
        }
        res = 0x7c00000000000000;
        return (res, pfpsf);
    }
    if (((((tolower_macro(ps_at!(1)) == b'i') && (tolower_macro(ps_at!(2)) == b'n')) && (tolower_macro(ps_at!(3)) == b'f'))) && (((ps_at!(4) == 0) || (((((((tolower_macro(ps_at!(4)) == b'i') && (tolower_macro(ps_at!(5)) == b'n')) && (tolower_macro(ps_at!(6)) == b'i')) && (tolower_macro(ps_at!(7)) == b't')) && (tolower_macro(ps_at!(8)) == b'y')) && (ps_at!(9) == 0)))))) {
        if (c == b'+') {
            res = 0x7800000000000000;
        } else if (c == b'-') {
            res = 0xf800000000000000;
        } else {
            res = 0x7c00000000000000;
        }
        return (res, pfpsf);
    }
    if ((((tolower_macro(ps_at!(1)) == b's') && (tolower_macro(ps_at!(2)) == b'n')) && (tolower_macro(ps_at!(3)) == b'a')) && (tolower_macro(ps_at!(4)) == b'n')) {
        if (c == b'-') {
            res = 0xfe00000000000000;
        } else {
            res = 0x7e00000000000000;
        }
        return (res, pfpsf);
    }
    if (c == b'-') {
        sign_x = 0x8000000000000000;
    } else {
        sign_x = 0;
    }
    if ((c == b'-') || (c == b'+')) {
        ps_idx += 1;
        c = ps_at!(0);
    }
    if ((c != b'.') && (((c < b'0') || (c > b'9')))) {
        res = (0x7c00000000000000 | sign_x);
        return (res, pfpsf);
    }
    rdx_pt_enc = 0;
    if ((ps_at!(0) == b'0') || (ps_at!(0) == b'.')) {
        if (ps_at!(0) == b'.') {
            rdx_pt_enc = 1;
            ps_idx += 1;
        }
        while (ps_at!(0) == b'0') {
            ps_idx += 1;
            if (rdx_pt_enc != 0) {
                right_radix_leading_zeros = right_radix_leading_zeros.wrapping_add(1);
            }
            if (ps_at!(0) == b'.') {
                if (rdx_pt_enc == 0) {
                    rdx_pt_enc = 1;
                    if (ps_at!(1) == 0) {
                        res = (((go_checked_shl_u64((((398 as i64).wrapping_sub(right_radix_leading_zeros)) as u64), go_shift_count_u64((53) as u64)))) | sign_x);
                        return (res, pfpsf);
                    }
                    ps_idx += 1;
                } else {
                    res = (0x7c00000000000000 | sign_x);
                    return (res, pfpsf);
                }
            } else if (ps_at!(0) == 0) {
                res = (((go_checked_shl_u64((((398 as i64).wrapping_sub(right_radix_leading_zeros)) as u64), go_shift_count_u64((53) as u64)))) | sign_x);
                return (res, pfpsf);
            }
        }
    }
    c = ps_at!(0);
    ndigits = 0;
    while ((((c >= b'0') && (c <= b'9'))) || (c == b'.')) {
        if (c == b'.') {
            if (rdx_pt_enc != 0) {
                res = (0x7c00000000000000 | sign_x);
                return (res, pfpsf);
            }
            rdx_pt_enc = 1;
            ps_idx += 1;
            c = ps_at!(0);
            continue;
        }
        dec_expon_scale = dec_expon_scale.wrapping_add(rdx_pt_enc);
        ndigits = ndigits.wrapping_add(1);
        if (ndigits <= 16) {
            coefficient_x = (((go_checked_shl_u64(coefficient_x, go_shift_count_u64((1) as u64)))).wrapping_add(((go_checked_shl_u64(coefficient_x, go_shift_count_u64((3) as u64))))));
            coefficient_x = coefficient_x.wrapping_add(((c.wrapping_sub(b'0')) as u64));
        } else if (ndigits == 17) {
            let mut doOverflowCheck = false;
            match rnd_mode {
                0 => {
                    if ((c == b'5') && ((coefficient_x & 1) == 0)) {
                        midpoint = 1;
                    } else {
                        midpoint = 0;
                    }
                    if ((c > b'5') || (((c == b'5') && ((coefficient_x & 1) != 0)))) {
                        coefficient_x = coefficient_x.wrapping_add(1);
                        rounded_up = 1;
                    } else {
                        doOverflowCheck = true;
                    }
                }
                1 => {
                    if (sign_x != 0) {
                        if (c > b'0') {
                            coefficient_x = coefficient_x.wrapping_add(1);
                            rounded_up = 1;
                        } else {
                            dround = 1;
                        }
                    }
                }
                2 => {
                    if (sign_x == 0) {
                        if (c > b'0') {
                            coefficient_x = coefficient_x.wrapping_add(1);
                            rounded_up = 1;
                        } else {
                            dround = 1;
                        }
                    }
                }
                4 => {
                    if (c >= b'5') {
                        coefficient_x = coefficient_x.wrapping_add(1);
                        rounded_up = 1;
                    }
                }
                _ => {
                }
            }
            if doOverflowCheck {
                if (coefficient_x == 10000000000000000) {
                    coefficient_x = 1000000000000000;
                    add_expon = 1;
                }
            }
            if (c > b'0') {
                rounded = 1;
            }
            add_expon = add_expon.wrapping_add(1);
        } else {
            add_expon = add_expon.wrapping_add(1);
            if ((midpoint != 0) && (c > b'0')) {
                coefficient_x = coefficient_x.wrapping_add(1);
                midpoint = 0;
                rounded_up = 1;
            }
            if (c > b'0') {
                rounded = 1;
                if (dround != 0) {
                    dround = 0;
                    coefficient_x = coefficient_x.wrapping_add(1);
                    rounded_up = 1;
                    if (coefficient_x == 10000000000000000) {
                        coefficient_x = 1000000000000000;
                        add_expon = add_expon.wrapping_add(1);
                    }
                }
            }
        }
        ps_idx += 1;
        c = ps_at!(0);
    }
    add_expon = add_expon.wrapping_sub(((dec_expon_scale.wrapping_add(right_radix_leading_zeros))));
    if (c == 0) {
        if (rounded != 0) {
            pfpsf |= 32;
        }
        res = fast_get_bid64_check_of_with_flags(sign_x, (add_expon.wrapping_add(0x18e)), coefficient_x, rnd_mode, (&mut pfpsf));
        return (res, pfpsf);
    }
    if ((c != b'E') && (c != b'e')) {
        res = (0x7c00000000000000 | sign_x);
        return (res, pfpsf);
    }
    ps_idx += 1;
    c = ps_at!(0);
    if (c == b'-') {
        sgn_expon = 1;
    } else {
        sgn_expon = 0;
    }
    if ((c == b'-') || (c == b'+')) {
        ps_idx += 1;
        c = ps_at!(0);
    }
    if (((c == 0) || (c < b'0')) || (c > b'9')) {
        res = (0x7c00000000000000 | sign_x);
        return (res, pfpsf);
    }
    while ((c >= b'0') && (c <= b'9')) {
        if (expon_x < (1 << 20)) {
            expon_x = (((go_checked_shl_i64(expon_x, go_shift_count_u64((1) as u64)))).wrapping_add(((go_checked_shl_i64(expon_x, go_shift_count_u64((3) as u64))))));
            expon_x = expon_x.wrapping_add(((c.wrapping_sub(b'0')) as i64));
        }
        ps_idx += 1;
        c = ps_at!(0);
    }
    if (c != 0) {
        res = (0x7c00000000000000 | sign_x);
        return (res, pfpsf);
    }
    if (rounded != 0) {
        pfpsf |= 32;
    }
    if (sgn_expon != 0) {
        expon_x = (expon_x.wrapping_neg());
    }
    expon_x = expon_x.wrapping_add((add_expon.wrapping_add(0x18e)));
    if (expon_x < 0) {
        if (rounded_up != 0) {
            coefficient_x = coefficient_x.wrapping_sub(1);
        }
        rnd_mode = 0;
        res = get_bid64_uf_with_flags(sign_x, expon_x, coefficient_x, rounded, rnd_mode, (&mut pfpsf));
        return (res, pfpsf);
    }
    res = get_bid64_with_flags(sign_x, expon_x, coefficient_x, rnd_mode, (&mut pfpsf));
    return (res, pfpsf);
}

pub(crate) fn fast_get_bid64_check_of_with_flags(mut sgn: u64, mut expon: i64, mut coeff: u64, mut rmode: i64, pfpsf: &mut u32) -> u64 {
    let mut r: u64 = 0;
    let mut mask: u64 = 0;
    if ((expon as u64) >= ((3 * 256) - 1)) {
        if ((expon == ((3 * 256) - 1)) && (coeff == 10000000000000000)) {
            expon = (3 * 256);
            coeff = 1000000000000000;
        }
        if ((expon as u64) >= (3 * 256)) {
            while ((coeff < 1000000000000000) && (expon >= (3 * 256))) {
                expon = expon.wrapping_sub(1);
                coeff = (((go_checked_shl_u64(coeff, go_shift_count_u64((3) as u64)))).wrapping_add(((go_checked_shl_u64(coeff, go_shift_count_u64((1) as u64))))));
            }
            if (expon > 0x2ff) {
                (*pfpsf) |= (8 | 32);
                r = (sgn | 0x7800000000000000);
                match rmode {
                    1 => {
                        if (sgn == 0) {
                            r = 0x77fb86f26fc0ffff;
                        }
                    }
                    3 => {
                        r = (sgn | 0x77fb86f26fc0ffff);
                    }
                    2 => {
                        if (sgn != 0) {
                            r = 0xf7fb86f26fc0ffff;
                        }
                    }
                    _ => {}
                }
                return r;
            }
        }
    }
    mask = 1;
    mask = go_checked_shl_u64(mask, go_shift_count_u64((53) as u64));
    if (coeff < mask) {
        r = (expon as u64);
        r = go_checked_shl_u64(r, go_shift_count_u64((53) as u64));
        r |= (coeff | sgn);
        return r;
    }
    if (coeff == 10000000000000000) {
        r = ((expon.wrapping_add(1)) as u64);
        r = go_checked_shl_u64(r, go_shift_count_u64((53) as u64));
        r |= (1000000000000000 | sgn);
        return r;
    }
    r = (expon as u64);
    r = go_checked_shl_u64(r, go_shift_count_u64((51) as u64));
    r |= (sgn | 0x6000000000000000);
    mask = (((go_checked_shr_u64(mask, go_shift_count_u64((2) as u64)))).wrapping_sub(1));
    coeff &= mask;
    r |= coeff;
    return r;
}

pub(crate) fn get_bid64_uf_with_flags(mut sgn: u64, mut expon: i64, mut coeff: u64, mut R: u64, mut rmode: i64, pfpsf: &mut u32) -> u64 {
    let mut Q_low: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut Stemp: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut _C64: u64 = 0;
    let mut remainder_h: u64 = 0;
    let mut QH: u64 = 0;
    let mut carry: u64 = 0;
    let mut CY: u64 = 0;
    let mut extra_digits: i64 = 0;
    let mut amount: i64 = 0;
    let mut amount2: i64 = 0;
    let mut status: u32 = 0;
    if ((expon.wrapping_add(16)) < 0) {
        (*pfpsf) |= (16 | 32);
        if ((rmode == 1) && (sgn != 0)) {
            return 0x8000000000000001;
        }
        if ((rmode == 2) && (sgn == 0)) {
            return 1;
        }
        return sgn;
    }
    coeff = (((go_checked_shl_u64(coeff, go_shift_count_u64((3) as u64)))).wrapping_add(((go_checked_shl_u64(coeff, go_shift_count_u64((1) as u64))))));
    if ((sgn != 0) && (((rmode.wrapping_sub(1)) as u64) < 2)) {
        rmode = ((3 as i64).wrapping_sub(rmode));
    }
    if (R != 0) {
        coeff |= 1;
    }
    extra_digits = ((1 as i64).wrapping_sub(expon));
    let mut C128w0 = (coeff.wrapping_add(bid_round_const_table[rmode as usize][extra_digits as usize]));
    (QH, Q_low) = __mul_64x128_full(C128w0, bid_reciprocals10_128[extra_digits as usize]);
    amount = (bid_recip_scale[extra_digits as usize] as i64);
    _C64 = (go_checked_shr_u64(QH, go_shift_count_u64((amount as u64) as u64)));
    if (rmode == 0) {
        if ((_C64 & 1) != 0) {
            amount2 = ((64 as i64).wrapping_sub(amount));
            remainder_h = (go_checked_shr_u64(((!(0 as u64))), go_shift_count_u64((amount2 as u64) as u64)));
            remainder_h = (remainder_h & QH);
            if ((remainder_h == 0) && (((Q_low.w[1] < bid_reciprocals10_128[extra_digits as usize].w[1]) || (((Q_low.w[1] == bid_reciprocals10_128[extra_digits as usize].w[1]) && (Q_low.w[0] < bid_reciprocals10_128[extra_digits as usize].w[0])))))) {
                _C64 = _C64.wrapping_sub(1);
            }
        }
    }
    if ((((*pfpsf) & 32)) != 0) {
        (*pfpsf) |= 16;
    } else {
        status = 32;
        remainder_h = (go_checked_shl_u64(QH, go_shift_count_u64(((((64 as i64).wrapping_sub(amount)) as u64)) as u64)));
        match rmode {
            0 | 4 => {
                if ((remainder_h == 0x8000000000000000) && (((Q_low.w[1] < bid_reciprocals10_128[extra_digits as usize].w[1]) || (((Q_low.w[1] == bid_reciprocals10_128[extra_digits as usize].w[1]) && (Q_low.w[0] < bid_reciprocals10_128[extra_digits as usize].w[0])))))) {
                    status = 0;
                }
            }
            1 | 3 => {
                if ((remainder_h == 0) && (((Q_low.w[1] < bid_reciprocals10_128[extra_digits as usize].w[1]) || (((Q_low.w[1] == bid_reciprocals10_128[extra_digits as usize].w[1]) && (Q_low.w[0] < bid_reciprocals10_128[extra_digits as usize].w[0])))))) {
                    status = 0;
                }
            }
            _ => {
                (Stemp.w[0], CY) = __add_carry_out(Q_low.w[0], bid_reciprocals10_128[extra_digits as usize].w[0]);
                (Stemp.w[1], carry) = __add_carry_in_out(Q_low.w[1], bid_reciprocals10_128[extra_digits as usize].w[1], CY);
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
    return (sgn | _C64);
}

pub(crate) fn get_bid64_with_flags(mut sgn: u64, mut expon: i64, mut coeff: u64, mut rmode: i64, pfpsf: &mut u32) -> u64 {
    let mut Q_low: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut Stemp: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut QH: u64 = 0;
    let mut r: u64 = 0;
    let mut mask: u64 = 0;
    let mut _C64: u64 = 0;
    let mut remainder_h: u64 = 0;
    let mut CY: u64 = 0;
    let mut carry: u64 = 0;
    let mut extra_digits: i64 = 0;
    let mut amount: i64 = 0;
    let mut amount2: i64 = 0;
    let mut status: u32 = 0;
    if (coeff > 9999999999999999) {
        expon = expon.wrapping_add(1);
        coeff = 1000000000000000;
    }
    if ((expon as u64) >= (3 * 256)) {
        if (expon < 0) {
            if ((expon.wrapping_add(16)) < 0) {
                (*pfpsf) |= (16 | 32);
                if ((rmode == 1) && (sgn != 0)) {
                    return 0x8000000000000001;
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
            let mut C128w0 = (coeff.wrapping_add(bid_round_const_table[rmode as usize][extra_digits as usize]));
            (QH, Q_low) = __mul_64x128_full(C128w0, bid_reciprocals10_128[extra_digits as usize]);
            amount = (bid_recip_scale[extra_digits as usize] as i64);
            _C64 = (go_checked_shr_u64(QH, go_shift_count_u64((amount as u64) as u64)));
            if (rmode == 0) {
                if ((_C64 & 1) != 0) {
                    amount2 = ((64 as i64).wrapping_sub(amount));
                    remainder_h = (go_checked_shr_u64(((!(0 as u64))), go_shift_count_u64((amount2 as u64) as u64)));
                    remainder_h = (remainder_h & QH);
                    if ((remainder_h == 0) && (((Q_low.w[1] < bid_reciprocals10_128[extra_digits as usize].w[1]) || (((Q_low.w[1] == bid_reciprocals10_128[extra_digits as usize].w[1]) && (Q_low.w[0] < bid_reciprocals10_128[extra_digits as usize].w[0])))))) {
                        _C64 = _C64.wrapping_sub(1);
                    }
                }
            }
            if ((((*pfpsf) & 32)) != 0) {
                (*pfpsf) |= 16;
            } else {
                status = 32;
                remainder_h = (go_checked_shl_u64(QH, go_shift_count_u64(((((64 as i64).wrapping_sub(amount)) as u64)) as u64)));
                match rmode {
                    0 | 4 => {
                        if ((remainder_h == 0x8000000000000000) && (((Q_low.w[1] < bid_reciprocals10_128[extra_digits as usize].w[1]) || (((Q_low.w[1] == bid_reciprocals10_128[extra_digits as usize].w[1]) && (Q_low.w[0] < bid_reciprocals10_128[extra_digits as usize].w[0])))))) {
                            status = 0;
                        }
                    }
                    1 | 3 => {
                        if ((remainder_h == 0) && (((Q_low.w[1] < bid_reciprocals10_128[extra_digits as usize].w[1]) || (((Q_low.w[1] == bid_reciprocals10_128[extra_digits as usize].w[1]) && (Q_low.w[0] < bid_reciprocals10_128[extra_digits as usize].w[0])))))) {
                            status = 0;
                        }
                    }
                    _ => {
                        (Stemp.w[0], CY) = __add_carry_out(Q_low.w[0], bid_reciprocals10_128[extra_digits as usize].w[0]);
                        (Stemp.w[1], carry) = __add_carry_in_out(Q_low.w[1], bid_reciprocals10_128[extra_digits as usize].w[1], CY);
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
            return (sgn | _C64);
        }
        if (coeff == 0) {
            if (expon > 0x2ff) {
                expon = 0x2ff;
            }
        }
        while ((coeff < 1000000000000000) && (expon >= (3 * 256))) {
            expon = expon.wrapping_sub(1);
            coeff = (((go_checked_shl_u64(coeff, go_shift_count_u64((3) as u64)))).wrapping_add(((go_checked_shl_u64(coeff, go_shift_count_u64((1) as u64))))));
        }
        if (expon > 0x2ff) {
            (*pfpsf) |= (8 | 32);
            r = (sgn | 0x7800000000000000);
            match rmode {
                1 => {
                    if (sgn == 0) {
                        r = 0x77fb86f26fc0ffff;
                    }
                }
                3 => {
                    r = (sgn | 0x77fb86f26fc0ffff);
                }
                2 => {
                    if (sgn != 0) {
                        r = 0xf7fb86f26fc0ffff;
                    }
                }
                _ => {}
            }
            return r;
        }
    }
    mask = 1;
    mask = go_checked_shl_u64(mask, go_shift_count_u64((53) as u64));
    if (coeff < mask) {
        r = (expon as u64);
        r = go_checked_shl_u64(r, go_shift_count_u64((53) as u64));
        r |= (coeff | sgn);
        return r;
    }
    if (coeff == 10000000000000000) {
        r = ((expon.wrapping_add(1)) as u64);
        r = go_checked_shl_u64(r, go_shift_count_u64((53) as u64));
        r |= (1000000000000000 | sgn);
        return r;
    }
    r = (expon as u64);
    r = go_checked_shl_u64(r, go_shift_count_u64((51) as u64));
    r |= (sgn | 0x6000000000000000);
    mask = (((go_checked_shr_u64(mask, go_shift_count_u64((2) as u64)))).wrapping_sub(1));
    coeff &= mask;
    r |= coeff;
    return r;
}
