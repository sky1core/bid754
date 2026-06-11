// Auto-generated from internal.go by go2rs. Do not edit.

use super::prelude::*;

pub(crate) fn __shr_128(mut a: BID_UINT128, mut k: u64) -> BID_UINT128 {
    let mut q: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    q.w[0] = (go_checked_shr_u64(a.w[0], go_shift_count_u64((k) as u64)));
    q.w[0] |= (go_checked_shl_u64(a.w[1], go_shift_count_u64(((((64 as u64).wrapping_sub(k)))) as u64)));
    q.w[1] = (go_checked_shr_u64(a.w[1], go_shift_count_u64((k) as u64)));
    return q;
}

pub(crate) fn __shr_128_long(mut a: BID_UINT128, mut k: u64) -> BID_UINT128 {
    let mut q: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    if (k < 64) {
        q.w[0] = (go_checked_shr_u64(a.w[0], go_shift_count_u64((k) as u64)));
        q.w[0] |= (go_checked_shl_u64(a.w[1], go_shift_count_u64(((((64 as u64).wrapping_sub(k)))) as u64)));
        q.w[1] = (go_checked_shr_u64(a.w[1], go_shift_count_u64((k) as u64)));
    } else {
        q.w[0] = (go_checked_shr_u64(a.w[1], go_shift_count_u64((((k.wrapping_sub(64)))) as u64)));
        q.w[1] = 0;
    }
    return q;
}

pub(crate) fn __shl_128_long(mut a: BID_UINT128, mut k: u64) -> BID_UINT128 {
    let mut q: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    if (k < 64) {
        q.w[1] = (go_checked_shl_u64(a.w[1], go_shift_count_u64((k) as u64)));
        q.w[1] |= (go_checked_shr_u64(a.w[0], go_shift_count_u64(((((64 as u64).wrapping_sub(k)))) as u64)));
        q.w[0] = (go_checked_shl_u64(a.w[0], go_shift_count_u64((k) as u64)));
    } else {
        q.w[1] = (go_checked_shl_u64(a.w[0], go_shift_count_u64((((k.wrapping_sub(64)))) as u64)));
        q.w[0] = 0;
    }
    return q;
}

pub(crate) fn __unsigned_compare_gt_128(mut a: BID_UINT128, mut b: BID_UINT128) -> bool {
    return ((a.w[1] > b.w[1]) || (((a.w[1] == b.w[1]) && (a.w[0] > b.w[0]))));
}

pub(crate) fn __unsigned_compare_ge_128(mut a: BID_UINT128, mut b: BID_UINT128) -> bool {
    return ((a.w[1] > b.w[1]) || (((a.w[1] == b.w[1]) && (a.w[0] >= b.w[0]))));
}

pub(crate) fn __test_equal_128(mut a: BID_UINT128, mut b: BID_UINT128) -> bool {
    return ((a.w[1] == b.w[1]) && (a.w[0] == b.w[0]));
}

pub(crate) fn __add_128_64(mut a: BID_UINT128, mut b: u64) -> BID_UINT128 {
    let mut r: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut r64h = a.w[1];
    r.w[0] = (b.wrapping_add(a.w[0]));
    if (r.w[0] < b) {
        r64h = r64h.wrapping_add(1);
    }
    r.w[1] = r64h;
    return r;
}

pub(crate) fn __sub_128_64(mut a: BID_UINT128, mut b: u64) -> BID_UINT128 {
    let mut r: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut r64h = a.w[1];
    if (a.w[0] < b) {
        r64h = r64h.wrapping_sub(1);
    }
    r.w[1] = r64h;
    r.w[0] = (a.w[0].wrapping_sub(b));
    return r;
}

pub(crate) fn __add_128_128(mut a: BID_UINT128, mut b: BID_UINT128) -> BID_UINT128 {
    let mut q: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    q.w[1] = (a.w[1].wrapping_add(b.w[1]));
    q.w[0] = (b.w[0].wrapping_add(a.w[0]));
    if (q.w[0] < b.w[0]) {
        q.w[1] = q.w[1].wrapping_add(1);
    }
    return q;
}

pub(crate) fn __sub_128_128(mut a: BID_UINT128, mut b: BID_UINT128) -> BID_UINT128 {
    let mut q: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    q.w[1] = (a.w[1].wrapping_sub(b.w[1]));
    q.w[0] = (a.w[0].wrapping_sub(b.w[0]));
    if (a.w[0] < b.w[0]) {
        q.w[1] = q.w[1].wrapping_sub(1);
    }
    return q;
}

pub(crate) fn __add_carry_out(mut x: u64, mut y: u64) -> (u64, u64) {
    let mut s: u64 = 0;
    let mut cy: u64 = 0;
    s = (x.wrapping_add(y));
    if (s < x) {
        cy = 1;
    }
    return (s, cy);
}

pub(crate) fn __add_carry_in_out(mut x: u64, mut y: u64, mut ci: u64) -> (u64, u64) {
    let mut s: u64 = 0;
    let mut cy: u64 = 0;
    let mut x1 = (x.wrapping_add(ci));
    s = (x1.wrapping_add(y));
    if ((s < x1) || (x1 < ci)) {
        cy = 1;
    }
    return (s, cy);
}

pub(crate) fn __sub_borrow_out(mut x: u64, mut y: u64) -> (u64, u64) {
    let mut s: u64 = 0;
    let mut cy: u64 = 0;
    s = (x.wrapping_sub(y));
    if (s > x) {
        cy = 1;
    }
    return (s, cy);
}

pub(crate) fn __mul_64x64_to_128(mut cx: u64, mut cy: u64) -> BID_UINT128 {
    let (mut hi, mut lo) = go_mul64(cx, cy);
    return BID_UINT128 { w: [lo, hi], ..Default::default() };
}

pub(crate) fn __mul_64x64_to_128_fast(mut cx: u64, mut cy: u64) -> BID_UINT128 {
    return __mul_64x64_to_128(cx, cy);
}

pub(crate) fn __mul_64x128_full(mut a: u64, mut b: BID_UINT128) -> (u64, BID_UINT128) {
    let mut ph: u64 = 0;
    let mut ql: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut albh = __mul_64x64_to_128(a, b.w[1]);
    let mut albl = __mul_64x64_to_128(a, b.w[0]);
    ql.w[0] = albl.w[0];
    let mut qm2 = __add_128_64(albh, albl.w[1]);
    ql.w[1] = qm2.w[0];
    ph = qm2.w[1];
    return (ph, ql);
}

pub(crate) fn __mul_64x128_to_192(mut a: u64, mut b: BID_UINT128) -> BID_UINT192 {
    let mut q: BID_UINT192 = BID_UINT192 { w: [0, 0, 0] };
    let mut albh = __mul_64x64_to_128(a, b.w[1]);
    let mut albl = __mul_64x64_to_128(a, b.w[0]);
    q.w[0] = albl.w[0];
    let mut qm2 = __add_128_64(albh, albl.w[1]);
    q.w[1] = qm2.w[0];
    q.w[2] = qm2.w[1];
    return q;
}

pub(crate) fn __mul_128x128_to_256(mut a: BID_UINT128, mut b: BID_UINT128) -> BID_UINT256 {
    let mut p256: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut cy1: u64 = 0;
    let mut cy2: u64 = 0;
    let (mut phl, mut qll) = __mul_64x128_full(a.w[0], b);
    let (mut phh, mut qlh) = __mul_64x128_full(a.w[1], b);
    p256.w[0] = qll.w[0];
    (p256.w[1], cy1) = __add_carry_out(qlh.w[0], qll.w[1]);
    (p256.w[2], cy2) = __add_carry_in_out(qlh.w[1], phl, cy1);
    p256.w[3] = (phh.wrapping_add(cy2));
    return p256;
}

pub(crate) fn unpack_bid64(mut x: u64) -> (u64, i64, u64, bool) {
    let mut sign: u64 = 0;
    let mut exponent: i64 = 0;
    let mut coefficient: u64 = 0;
    let mut valid: bool = false;
    sign = (x & 0x8000000000000000);
    if ((x & 0x6000000000000000) == 0x6000000000000000) {
        coefficient = ((x & 0x7ffffffffffff) | 0x20000000000000);
        if ((x & 0x7800000000000000) == 0x7800000000000000) {
            exponent = 0;
            coefficient = (x & 0xfe03ffffffffffff);
            if ((x & 0x0003ffffffffffff) >= 1000000000000000) {
                coefficient = (x & 0xfe00000000000000);
            }
            if ((x & 0x7c00000000000000) == 0x7800000000000000) {
                coefficient = (x & 0xf800000000000000);
            }
            return (sign, exponent, coefficient, false);
        }
        if (coefficient >= 10000000000000000) {
            coefficient = 0;
        }
        let mut tmp = (go_checked_shr_u64(x, go_shift_count_u64((51) as u64)));
        exponent = ((tmp & 0x3ff) as i64);
        return (sign, exponent, coefficient, (coefficient != 0));
    }
    let mut tmp = (go_checked_shr_u64(x, go_shift_count_u64((53) as u64)));
    exponent = ((tmp & 0x3ff) as i64);
    coefficient = (x & 0x1fffffffffffff);
    return (sign, exponent, coefficient, (coefficient != 0));
}

pub(crate) fn very_fast_get_bid64(mut sgn: u64, mut expon: i64, mut coeff: u64) -> u64 {
    let mut r: u64 = 0;
    let mut mask: u64 = ((1 as u64) << 53);
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

pub(crate) fn fast_get_bid64(mut sgn: u64, mut expon: i64, mut coeff: u64) -> u64 {
    return very_fast_get_bid64(sgn, expon, coeff);
}

pub(crate) fn fast_get_bid64_check_of(mut sgn: u64, mut expon: i64, mut coeff: u64, mut rmode: i64) -> u64 {
    let mut r: u64 = 0;
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
    let mut mask: u64 = ((1 as u64) << 53);
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

pub(crate) fn get_bid64(mut sgn: u64, mut expon: i64, mut coeff: u64, mut rmode: i64) -> u64 {
    let mut r: u64 = 0;
    if (coeff > 9999999999999999) {
        expon = expon.wrapping_add(1);
        coeff = 1000000000000000;
    }
    if ((expon as u64) >= (3 * 256)) {
        if (expon < 0) {
            if ((expon.wrapping_add(16)) < 0) {
                if ((rmode == 1) && (sgn != 0)) {
                    return 0x8000000000000001;
                }
                if ((rmode == 2) && (sgn == 0)) {
                    return 1;
                }
                return sgn;
            }
            if ((sgn != 0) && (((rmode == 1) || (rmode == 2)))) {
                if (rmode == 1) {
                    rmode = 2;
                } else {
                    rmode = 1;
                }
            }
            let mut extraDigits = (expon.wrapping_neg());
            coeff = coeff.wrapping_add(bid_round_const_table[rmode as usize][extraDigits as usize]);
            let (mut qh, mut qLow) = __mul_64x128_full(coeff, bid_reciprocals10_128[extraDigits as usize]);
            let mut amount = (bid_recip_scale[extraDigits as usize] as i64);
            let mut c64 = (go_checked_shr_u64(qh, go_shift_count_u64((amount as u64) as u64)));
            if (rmode == 0) {
                if ((c64 & 1) != 0) {
                    let mut amount2 = ((64 as i64).wrapping_sub(amount));
                    let mut remainderH = (go_checked_shr_u64(((!(0 as u64))), go_shift_count_u64((amount2 as u64) as u64)));
                    remainderH = (remainderH & qh);
                    if ((remainderH == 0) && (((qLow.w[1] < bid_reciprocals10_128[extraDigits as usize].w[1]) || (((qLow.w[1] == bid_reciprocals10_128[extraDigits as usize].w[1]) && (qLow.w[0] < bid_reciprocals10_128[extraDigits as usize].w[0])))))) {
                        c64 = c64.wrapping_sub(1);
                    }
                }
            } else if (rmode == 5) {
                let mut amount2 = ((64 as i64).wrapping_sub(amount));
                let mut remainderH = (go_checked_shr_u64(((!(0 as u64))), go_shift_count_u64((amount2 as u64) as u64)));
                remainderH = (remainderH & qh);
                if ((remainderH == 0) && (((qLow.w[1] < bid_reciprocals10_128[extraDigits as usize].w[1]) || (((qLow.w[1] == bid_reciprocals10_128[extraDigits as usize].w[1]) && (qLow.w[0] < bid_reciprocals10_128[extraDigits as usize].w[0])))))) {
                    c64 = c64.wrapping_sub(1);
                }
            }
            return (sgn | c64);
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
    return very_fast_get_bid64(sgn, expon, coeff);
}

pub(crate) fn get_binary_exponent(mut coefficient: u64) -> i64 {
    let mut f = (coefficient as f64);
    let mut bits = (f).to_bits();
    return (((go_checked_shr_u64((bits & 0x7ff0000000000000), go_shift_count_u64((52) as u64))) as i64).wrapping_sub(0x3ff));
}

pub(crate) fn __tight_bin_range_128(mut P: BID_UINT128, mut binExpon: i64) -> i64 {
    let mut M: u64 = 1;
    let mut bp = binExpon;
    if (bp < 63) {
        M = go_checked_shl_u64(M, go_shift_count_u64((((bp.wrapping_add(1)) as u64)) as u64));
        if (P.w[0] >= M) {
            bp = bp.wrapping_add(1);
        }
    } else if (bp > 64) {
        M = go_checked_shl_u64(M, go_shift_count_u64(((((bp.wrapping_add(1)).wrapping_sub(64)) as u64)) as u64));
        if ((P.w[1] > M) || (((P.w[1] == M) && (P.w[0] != 0)))) {
            bp = bp.wrapping_add(1);
        }
    } else if (P.w[1] != 0) {
        bp = bp.wrapping_add(1);
    }
    return bp;
}

pub(crate) fn __mul_128x128_full(mut A: BID_UINT128, mut B: BID_UINT128) -> (BID_UINT128, BID_UINT128) {
    let mut Qh: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut Ql: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut ALBH = __mul_64x64_to_128(A.w[0], B.w[1]);
    let mut AHBL = __mul_64x64_to_128(B.w[0], A.w[1]);
    let mut ALBL = __mul_64x64_to_128(A.w[0], B.w[0]);
    let mut AHBH = __mul_64x64_to_128(A.w[1], B.w[1]);
    let mut QM = __add_128_128(ALBH, AHBL);
    Ql.w[0] = ALBL.w[0];
    let mut QM2 = __add_128_64(QM, ALBL.w[1]);
    Qh = __add_128_64(AHBH, QM2.w[1]);
    Ql.w[1] = QM2.w[0];
    return (Qh, Ql);
}

pub(crate) fn very_fast_get_bid64_small_mantissa(mut sgn: u64, mut expon: i64, mut coeff: u64) -> u64 {
    let mut r = (expon as u64);
    r = go_checked_shl_u64(r, go_shift_count_u64((53) as u64));
    r |= (coeff | sgn);
    return r;
}

pub(crate) fn get_bid64_small_mantissa(mut sgn: u64, mut expon: i64, mut coeff: u64, mut rmode: i64) -> u64 {
    let mut Q_low: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut _C64: u64 = 0;
    let mut remainder_h: u64 = 0;
    let mut QH: u64 = 0;
    let mut extra_digits: i64 = 0;
    let mut amount: i64 = 0;
    let mut amount2: i64 = 0;
    if ((expon as u64) >= (3 * 256)) {
        if (expon < 0) {
            if ((expon.wrapping_add(16)) < 0) {
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
            } else if (rmode == 5) {
                amount2 = ((64 as i64).wrapping_sub(amount));
                remainder_h = (go_checked_shr_u64(((!(0 as u64))), go_shift_count_u64((amount2 as u64) as u64)));
                remainder_h = (remainder_h & QH);
                if ((remainder_h == 0) && (((Q_low.w[1] < bid_reciprocals10_128[extra_digits as usize].w[1]) || (((Q_low.w[1] == bid_reciprocals10_128[extra_digits as usize].w[1]) && (Q_low.w[0] < bid_reciprocals10_128[extra_digits as usize].w[0])))))) {
                    _C64 = _C64.wrapping_sub(1);
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
            let mut r = (sgn | 0x7800000000000000);
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
    return very_fast_get_bid64_small_mantissa(sgn, expon, coeff);
}

pub(crate) fn get_bid64_small_mantissa_flags(mut sgn: u64, mut expon: i64, mut coeff: u64, mut rmode: i64, fpsc: &mut u32) -> u64 {
    let mut C128: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut Q_low: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut Stemp: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut r: u64 = 0;
    let mut mask: u64 = 0;
    let mut _C64: u64 = 0;
    let mut remainder_h: u64 = 0;
    let mut QH: u64 = 0;
    let mut carry: u64 = 0;
    let mut CY: u64 = 0;
    let mut extra_digits: i64 = 0;
    let mut amount: i64 = 0;
    let mut amount2: i64 = 0;
    let mut status: u32 = 0;
    if ((expon as u64) >= (3 * 256)) {
        if (expon < 0) {
            if ((expon.wrapping_add(16)) < 0) {
                (*fpsc) |= (16 | 32);
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
            C128.w[0] = (coeff.wrapping_add(bid_round_const_table[rmode as usize][extra_digits as usize]));
            (QH, Q_low) = __mul_64x128_full(C128.w[0], bid_reciprocals10_128[extra_digits as usize]);
            amount = (bid_recip_scale[extra_digits as usize] as i64);
            _C64 = (go_checked_shr_u64(QH, go_shift_count_u64((amount as u64) as u64)));
            if (rmode == 0) {
                if ((_C64 & 1) != 0) {
                    amount2 = ((64 as i64).wrapping_sub(amount));
                    remainder_h = 0;
                    remainder_h = remainder_h.wrapping_sub(1);
                    remainder_h = go_checked_shr_u64(remainder_h, go_shift_count_u64((amount2 as u64) as u64));
                    remainder_h = (remainder_h & QH);
                    if ((remainder_h == 0) && (((Q_low.w[1] < bid_reciprocals10_128[extra_digits as usize].w[1]) || (((Q_low.w[1] == bid_reciprocals10_128[extra_digits as usize].w[1]) && (Q_low.w[0] < bid_reciprocals10_128[extra_digits as usize].w[0])))))) {
                        _C64 = _C64.wrapping_sub(1);
                    }
                }
            }
            if ((((*fpsc) & 32)) != 0) {
                (*fpsc) |= 16;
            } else {
                status = 32;
                remainder_h = (go_checked_shl_u64(QH, go_shift_count_u64(((((64 as u64).wrapping_sub(amount as u64)))) as u64)));
                match rmode {
                    0 => {
                    }
                    4 => {
                        if ((remainder_h == 0x8000000000000000) && (((Q_low.w[1] < bid_reciprocals10_128[extra_digits as usize].w[1]) || (((Q_low.w[1] == bid_reciprocals10_128[extra_digits as usize].w[1]) && (Q_low.w[0] < bid_reciprocals10_128[extra_digits as usize].w[0])))))) {
                            status = 0;
                        }
                    }
                    1 => {
                    }
                    3 => {
                        if ((remainder_h == 0) && (((Q_low.w[1] < bid_reciprocals10_128[extra_digits as usize].w[1]) || (((Q_low.w[1] == bid_reciprocals10_128[extra_digits as usize].w[1]) && (Q_low.w[0] < bid_reciprocals10_128[extra_digits as usize].w[0])))))) {
                            status = 0;
                        }
                    }
                    _ => {
                        (Stemp.w[0], CY) = __add_carry_out(Q_low.w[0], bid_reciprocals10_128[extra_digits as usize].w[0]);
                        (Stemp.w[1], carry) = __add_carry_in_out(Q_low.w[1], bid_reciprocals10_128[extra_digits as usize].w[1], CY);
                        if ((((go_checked_shr_u64(remainder_h, go_shift_count_u64(((((64 as u64).wrapping_sub(amount as u64)))) as u64)))).wrapping_add(carry)) >= ((go_checked_shl_u64((1 as u64), go_shift_count_u64((amount as u64) as u64))))) {
                            status = 0;
                        }
                    }
                }
                if (status != 0) {
                    (*fpsc) |= (16 | status);
                }
            }
            return (sgn | _C64);
        }
        while ((coeff < 1000000000000000) && (expon >= (3 * 256))) {
            expon = expon.wrapping_sub(1);
            coeff = (((go_checked_shl_u64(coeff, go_shift_count_u64((3) as u64)))).wrapping_add(((go_checked_shl_u64(coeff, go_shift_count_u64((1) as u64))))));
        }
        if (expon > 0x2ff) {
            (*fpsc) |= (8 | 32);
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
        } else {
            mask = 1;
            mask = go_checked_shl_u64(mask, go_shift_count_u64((53) as u64));
            if (coeff >= mask) {
                r = (expon as u64);
                r = go_checked_shl_u64(r, go_shift_count_u64((51) as u64));
                r |= (sgn | 0x6000000000000000);
                mask = (((go_checked_shr_u64(mask, go_shift_count_u64((2) as u64)))).wrapping_sub(1));
                coeff &= mask;
                r |= coeff;
                return r;
            }
        }
    }
    r = (expon as u64);
    r = go_checked_shl_u64(r, go_shift_count_u64((53) as u64));
    r |= (coeff | sgn);
    return r;
}

pub(crate) fn rounding_mode_to_bid(mut mode: i32) -> i64 {
    match mode {
        0 => {
            return 0;
        }
        4 => {
            return 4;
        }
        3 => {
            return 3;
        }
        2 => {
            return 2;
        }
        1 => {
            return 1;
        }
        5 => {
            return 5;
        }
        _ => {
            return 0;
        }
    }
}

pub(crate) fn get_bid64_uf(mut sgn: u64, mut expon: i64, mut coeff: u64, mut R: u64, mut rmode: i64) -> u64 {
    let mut Q_low: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut _C64: u64 = 0;
    let mut remainder_h: u64 = 0;
    let mut QH: u64 = 0;
    let mut extra_digits: i64 = 0;
    let mut amount: i64 = 0;
    let mut amount2: i64 = 0;
    if ((expon.wrapping_add(16)) < 0) {
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
    } else if (rmode == 5) {
        amount2 = ((64 as i64).wrapping_sub(amount));
        remainder_h = (go_checked_shr_u64(((!(0 as u64))), go_shift_count_u64((amount2 as u64) as u64)));
        remainder_h = (remainder_h & QH);
        if ((remainder_h == 0) && (((Q_low.w[1] < bid_reciprocals10_128[extra_digits as usize].w[1]) || (((Q_low.w[1] == bid_reciprocals10_128[extra_digits as usize].w[1]) && (Q_low.w[0] < bid_reciprocals10_128[extra_digits as usize].w[0])))))) {
            _C64 = _C64.wrapping_sub(1);
        }
    }
    return (sgn | _C64);
}

pub(crate) fn get_bid64_flags(mut sgn: u64, mut expon: i64, mut coeff: u64, mut rmode: i64) -> (u64, u32) {
    let mut Q_low: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut QH: u64 = 0;
    let mut r: u64 = 0;
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
                status = (16 | 32);
                if ((rmode == 1) && (sgn != 0)) {
                    return (0x8000000000000001, status);
                }
                if ((rmode == 2) && (sgn == 0)) {
                    return (1, status);
                }
                return (sgn, status);
            }
            if ((sgn != 0) && (((rmode.wrapping_sub(1)) as u64) < 2)) {
                rmode = ((3 as i64).wrapping_sub(rmode));
            }
            extra_digits = (expon.wrapping_neg());
            coeff = coeff.wrapping_add(bid_round_const_table[rmode as usize][extra_digits as usize]);
            (QH, Q_low) = __mul_64x128_full(coeff, bid_reciprocals10_128[extra_digits as usize]);
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
            status = 32;
            remainder_h = (go_checked_shl_u64(QH, go_shift_count_u64(((((64 as u64).wrapping_sub(amount as u64)))) as u64)));
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
                    let mut Stemp_w0: u64 = 0;
                    (Stemp_w0, CY) = __add_carry_out(Q_low.w[0], bid_reciprocals10_128[extra_digits as usize].w[0]);
                    (_, carry) = __add_carry_in_out(Q_low.w[1], bid_reciprocals10_128[extra_digits as usize].w[1], CY);
                    _ = Stemp_w0;
                    if ((((go_checked_shr_u64(remainder_h, go_shift_count_u64(((((64 as u64).wrapping_sub(amount as u64)))) as u64)))).wrapping_add(carry)) >= ((go_checked_shl_u64((1 as u64), go_shift_count_u64((amount as u64) as u64))))) {
                        status = 0;
                    }
                }
            }
            if (status != 0) {
                status = (16 | status);
            }
            return ((sgn | _C64), status);
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
            status = (8 | 32);
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
            return (r, status);
        }
    }
    return (very_fast_get_bid64(sgn, expon, coeff), 0);
}

pub(crate) fn __mul_192x192_to_384(mut a: BID_UINT192, mut b: BID_UINT192) -> BID_UINT384 {
    let mut p: BID_UINT384 = BID_UINT384 { w: [0, 0, 0, 0, 0, 0] };
    let mut cy: u64 = 0;
    let mut p00 = __mul_64x64_to_128(a.w[0], b.w[0]);
    let mut p01 = __mul_64x64_to_128(a.w[0], b.w[1]);
    let mut p02 = __mul_64x64_to_128(a.w[0], b.w[2]);
    let mut p10 = __mul_64x64_to_128(a.w[1], b.w[0]);
    let mut p11 = __mul_64x64_to_128(a.w[1], b.w[1]);
    let mut p12 = __mul_64x64_to_128(a.w[1], b.w[2]);
    let mut p20 = __mul_64x64_to_128(a.w[2], b.w[0]);
    let mut p21 = __mul_64x64_to_128(a.w[2], b.w[1]);
    let mut p22 = __mul_64x64_to_128(a.w[2], b.w[2]);
    p.w[0] = p00.w[0];
    p.w[1] = (p00.w[1].wrapping_add(p01.w[0]));
    cy = 0;
    if (p.w[1] < p00.w[1]) {
        cy = cy.wrapping_add(1);
    }
    let mut tmp = p.w[1];
    p.w[1] = p.w[1].wrapping_add(p10.w[0]);
    if (p.w[1] < tmp) {
        cy = cy.wrapping_add(1);
    }
    p.w[2] = (cy.wrapping_add(p01.w[1]));
    cy = 0;
    if (p.w[2] < p01.w[1]) {
        cy = cy.wrapping_add(1);
    }
    tmp = p.w[2];
    p.w[2] = p.w[2].wrapping_add(p02.w[0]);
    if (p.w[2] < tmp) {
        cy = cy.wrapping_add(1);
    }
    tmp = p.w[2];
    p.w[2] = p.w[2].wrapping_add(p10.w[1]);
    if (p.w[2] < tmp) {
        cy = cy.wrapping_add(1);
    }
    tmp = p.w[2];
    p.w[2] = p.w[2].wrapping_add(p11.w[0]);
    if (p.w[2] < tmp) {
        cy = cy.wrapping_add(1);
    }
    tmp = p.w[2];
    p.w[2] = p.w[2].wrapping_add(p20.w[0]);
    if (p.w[2] < tmp) {
        cy = cy.wrapping_add(1);
    }
    p.w[3] = (cy.wrapping_add(p02.w[1]));
    cy = 0;
    if (p.w[3] < p02.w[1]) {
        cy = cy.wrapping_add(1);
    }
    tmp = p.w[3];
    p.w[3] = p.w[3].wrapping_add(p11.w[1]);
    if (p.w[3] < tmp) {
        cy = cy.wrapping_add(1);
    }
    tmp = p.w[3];
    p.w[3] = p.w[3].wrapping_add(p12.w[0]);
    if (p.w[3] < tmp) {
        cy = cy.wrapping_add(1);
    }
    tmp = p.w[3];
    p.w[3] = p.w[3].wrapping_add(p20.w[1]);
    if (p.w[3] < tmp) {
        cy = cy.wrapping_add(1);
    }
    tmp = p.w[3];
    p.w[3] = p.w[3].wrapping_add(p21.w[0]);
    if (p.w[3] < tmp) {
        cy = cy.wrapping_add(1);
    }
    p.w[4] = (cy.wrapping_add(p12.w[1]));
    cy = 0;
    if (p.w[4] < p12.w[1]) {
        cy = cy.wrapping_add(1);
    }
    tmp = p.w[4];
    p.w[4] = p.w[4].wrapping_add(p21.w[1]);
    if (p.w[4] < tmp) {
        cy = cy.wrapping_add(1);
    }
    tmp = p.w[4];
    p.w[4] = p.w[4].wrapping_add(p22.w[0]);
    if (p.w[4] < tmp) {
        cy = cy.wrapping_add(1);
    }
    p.w[5] = (cy.wrapping_add(p22.w[1]));
    return p;
}

pub(crate) fn __mul_256x256_to_512(mut a: BID_UINT256, mut b: BID_UINT256) -> BID_UINT512 {
    let mut p: BID_UINT512 = BID_UINT512 { w: [0, 0, 0, 0, 0, 0, 0, 0] };
    let mut aL = BID_UINT128 { w: [a.w[0], a.w[1]], ..Default::default() };
    let mut aH = BID_UINT128 { w: [a.w[2], a.w[3]], ..Default::default() };
    let mut bL = BID_UINT128 { w: [b.w[0], b.w[1]], ..Default::default() };
    let mut bH = BID_UINT128 { w: [b.w[2], b.w[3]], ..Default::default() };
    let mut p0 = __mul_128x128_to_256(aL, bL);
    let mut p1 = __mul_128x128_to_256(aH, bL);
    let mut p2 = __mul_128x128_to_256(aL, bH);
    let mut p3 = __mul_128x128_to_256(aH, bH);
    p.w[0] = p0.w[0];
    p.w[1] = p0.w[1];
    let mut cy: u64 = 0;
    p.w[2] = (p0.w[2].wrapping_add(p1.w[0]));
    cy = 0;
    if (p.w[2] < p0.w[2]) {
        cy = 1;
    }
    p.w[3] = ((p0.w[3].wrapping_add(p1.w[1])).wrapping_add(cy));
    cy = 0;
    if ((p.w[3] < p1.w[1]) || (((cy == 0) && (p.w[3] < p0.w[3])))) {
        cy = 1;
    }
    let mut c4 = (p1.w[2].wrapping_add(cy));
    cy = 0;
    if (c4 < p1.w[2]) {
        cy = 1;
    }
    let mut c5 = (p1.w[3].wrapping_add(cy));
    let mut tmp = p.w[2];
    p.w[2] = p.w[2].wrapping_add(p2.w[0]);
    cy = 0;
    if (p.w[2] < tmp) {
        cy = 1;
    }
    tmp = p.w[3];
    p.w[3] = p.w[3].wrapping_add((p2.w[1].wrapping_add(cy)));
    cy = 0;
    if ((p.w[3] < tmp) || (((p.w[3] == tmp) && (cy > 0)))) {
        cy = 1;
    }
    tmp = c4;
    c4 = c4.wrapping_add((p2.w[2].wrapping_add(cy)));
    cy = 0;
    if ((c4 < tmp) || (((c4 == tmp) && (cy > 0)))) {
        cy = 1;
    }
    c5 = c5.wrapping_add((p2.w[3].wrapping_add(cy)));
    p.w[4] = (c4.wrapping_add(p3.w[0]));
    cy = 0;
    if (p.w[4] < c4) {
        cy = 1;
    }
    p.w[5] = ((c5.wrapping_add(p3.w[1])).wrapping_add(cy));
    cy = 0;
    if ((p.w[5] < c5) || (((p.w[5] == c5) && (cy > 0)))) {
        cy = 1;
    }
    p.w[6] = (p3.w[2].wrapping_add(cy));
    cy = 0;
    if (p.w[6] < p3.w[2]) {
        cy = 1;
    }
    p.w[7] = (p3.w[3].wrapping_add(cy));
    return p;
}

pub(crate) fn __mul_64x128_to_128(mut a: u64, mut b: BID_UINT128) -> BID_UINT128 {
    let (_, mut ql) = __mul_64x128_full(a, b);
    return ql;
}

pub(crate) fn __mul_64x192_to_256(mut A: u64, mut B: BID_UINT192) -> BID_UINT256 {
    let mut P: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut c: u64 = 0;
    let mut lP0 = __mul_64x64_to_128(A, B.w[0]);
    let mut lP1 = __mul_64x64_to_128(A, B.w[1]);
    let mut lP2 = __mul_64x64_to_128(A, B.w[2]);
    P.w[0] = lP0.w[0];
    (P.w[1], c) = __add_carry_out(lP1.w[0], lP0.w[1]);
    (P.w[2], c) = __add_carry_in_out(lP2.w[0], lP1.w[1], c);
    P.w[3] = (lP2.w[1].wrapping_add(c));
    return P;
}

pub(crate) fn __mul_64x256_to_320(mut A: u64, mut B: BID_UINT256) -> BID_UINT320 {
    let mut P: BID_UINT320 = BID_UINT320 { w: [0, 0, 0, 0, 0] };
    let mut c: u64 = 0;
    let mut lP0 = __mul_64x64_to_128(A, B.w[0]);
    let mut lP1 = __mul_64x64_to_128(A, B.w[1]);
    let mut lP2 = __mul_64x64_to_128(A, B.w[2]);
    let mut lP3 = __mul_64x64_to_128(A, B.w[3]);
    P.w[0] = lP0.w[0];
    (P.w[1], c) = __add_carry_out(lP1.w[0], lP0.w[1]);
    (P.w[2], c) = __add_carry_in_out(lP2.w[0], lP1.w[1], c);
    (P.w[3], c) = __add_carry_in_out(lP3.w[0], lP2.w[1], c);
    P.w[4] = (lP3.w[1].wrapping_add(c));
    return P;
}

pub(crate) fn __mul_64x320_to_384(mut A: u64, mut B: BID_UINT320) -> BID_UINT384 {
    let mut P: BID_UINT384 = BID_UINT384 { w: [0, 0, 0, 0, 0, 0] };
    let mut c: u64 = 0;
    let mut lP0 = __mul_64x64_to_128(A, B.w[0]);
    let mut lP1 = __mul_64x64_to_128(A, B.w[1]);
    let mut lP2 = __mul_64x64_to_128(A, B.w[2]);
    let mut lP3 = __mul_64x64_to_128(A, B.w[3]);
    let mut lP4 = __mul_64x64_to_128(A, B.w[4]);
    P.w[0] = lP0.w[0];
    (P.w[1], c) = __add_carry_out(lP1.w[0], lP0.w[1]);
    (P.w[2], c) = __add_carry_in_out(lP2.w[0], lP1.w[1], c);
    (P.w[3], c) = __add_carry_in_out(lP3.w[0], lP2.w[1], c);
    (P.w[4], c) = __add_carry_in_out(lP4.w[0], lP3.w[1], c);
    P.w[5] = (lP4.w[1].wrapping_add(c));
    return P;
}

pub(crate) fn __sqr128_to_256(mut A: BID_UINT128) -> BID_UINT256 {
    let mut P256: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut c1: u64 = 0;
    let mut c2: u64 = 0;
    let mut Qhh = __mul_64x64_to_128(A.w[1], A.w[1]);
    let mut Qlh = __mul_64x64_to_128(A.w[0], A.w[1]);
    Qhh.w[1] = Qhh.w[1].wrapping_add(((go_checked_shr_u64(Qlh.w[1], go_shift_count_u64((63) as u64)))));
    Qlh.w[1] = (((Qlh.w[1].wrapping_add(Qlh.w[1]))) | ((go_checked_shr_u64(Qlh.w[0], go_shift_count_u64((63) as u64)))));
    Qlh.w[0] = Qlh.w[0].wrapping_add(Qlh.w[0]);
    let mut Qll = __mul_64x64_to_128(A.w[0], A.w[0]);
    (P256.w[1], c1) = __add_carry_out(Qlh.w[0], Qll.w[1]);
    P256.w[0] = Qll.w[0];
    (P256.w[2], c2) = __add_carry_in_out(Qlh.w[1], Qhh.w[0], c1);
    P256.w[3] = (Qhh.w[1].wrapping_add(c2));
    return P256;
}

pub(crate) fn bid_get_bid128_fast(mut sgn: u64, mut expon: i64, mut coeff: BID_UINT128) -> BID_UINT128 {
    let mut res: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    if ((coeff.w[1] == 0x0001ed09bead87c0) && (coeff.w[0] == 0x378d8e6400000000)) {
        expon = expon.wrapping_add(1);
        coeff.w[1] = 0x0000314dc6448d93;
        coeff.w[0] = 0x38c15b0a00000000;
    }
    res.w[0] = coeff.w[0];
    let mut tmp = (expon as u64);
    tmp = go_checked_shl_u64(tmp, go_shift_count_u64((49) as u64));
    res.w[1] = ((sgn | tmp) | coeff.w[1]);
    return res;
}

pub(crate) fn no_fma_mul_add_f64(mut a: f64, mut b: f64, mut c: f64) -> f64 {
    return (f64::from_bits((a * b).to_bits()) + c);
}

pub(crate) fn no_fma_mul_add_f32(mut a: f32, mut b: f32, mut c: f32) -> f32 {
    return (f32::from_bits(((a * b) as f32).to_bits()) + c);
}
