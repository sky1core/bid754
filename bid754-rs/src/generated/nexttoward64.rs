// Auto-generated from nexttoward64.go by go2rs. Do not edit.

use super::prelude::*;

pub(crate) fn bid128_na_n_to_bid64(mut hi: u64, mut lo: u64) -> (u64, u32) {
    let mut payloadHi = (hi & 0x00003fffffffffff);
    let mut payloadLo = lo;
    let mut t33hi: u64 = (0x0000314dc6448d93 as u64);
    let mut t33lo: u64 = (0x38c15b09ffffffff as u64);
    if ((payloadHi > t33hi) || (((payloadHi == t33hi) && (payloadLo > t33lo)))) {
        payloadHi = 0;
        payloadLo = 0;
    }
    let mut payload = bid128_coeff_big(payloadHi, payloadLo);
    payload /= BigUint::from(1000000000000000000 as u64);
    return (((hi & 0xfc00000000000000) | go_big_to_u64(&payload)), (|| -> u32 {
    if ((hi & 0x7e00000000000000) == 0x7e00000000000000) {
        return 1;
    }
    return 0;
})());
}

pub(crate) fn bid64_canonicalize_non_canonical_finite(mut x: u64) -> u64 {
    if ((x & 0x7800000000000000) == 0x7800000000000000) {
        return x;
    }
    if ((x & 0x6000000000000000) == 0x6000000000000000) {
        if ((((x & 0x7ffffffffffff) | 0x20000000000000)) > 9999999999999999) {
            return ((x & 0x8000000000000000) | ((go_checked_shl_u64((x & 0x1ff8000000000000), go_shift_count_u64((2) as u64)))));
        }
    }
    return x;
}

pub(crate) fn bid64_decode_for_compare(mut x: u64) -> (u64, i64, BigUint, bool) {
    let mut sign: u64 = 0;
    let mut exp: i64 = 0;
    let mut coeff: BigUint = BigUint::zero();
    let mut isZero: bool = false;
    let (mut sign, mut exp, mut c) = bid64_unpack_finite_for_round_local(x);
    coeff = BigUint::from(c);
    return (sign, exp, coeff, (c == 0));
}

pub(crate) fn bid64_compare_to_bid128(mut x: u64, mut y: BID_UINT128) -> i64 {
    x = bid64_canonicalize_non_canonical_finite(x);
    let (mut xSign, mut xExp, mut xCoeff, mut xZero) = bid64_decode_for_compare(x);
    let mut yd = bid128_decode(y.w[1], y.w[0]);
    if (bid64_is_inf(x) != 0) {
        if yd.isInf {
            if (xSign == yd.sign) {
                return 0;
            }
            if (xSign != 0) {
                return (-1);
            }
            return 1;
        }
        if (xSign != 0) {
            return (-1);
        }
        return 1;
    }
    if yd.isInf {
        if (yd.sign != 0) {
            return 1;
        }
        return (-1);
    }
    if (xZero && yd.isZero) {
        return 0;
    }
    if (xSign != yd.sign) {
        if (xZero && yd.isZero) {
            return 0;
        }
        if (xSign != 0) {
            return (-1);
        }
        return 1;
    }
    let mut xc = xCoeff.clone();
    let mut yc = yd.coeff.clone();
    if (xExp > yd.exp) {
        xc *= bid128_pow10_big((xExp.wrapping_sub(yd.exp)));
    } else if (yd.exp > xExp) {
        yc *= bid128_pow10_big((yd.exp.wrapping_sub(xExp)));
    }
    let mut cmp = go_big_cmp(&xc, &yc);
    if (xSign != 0) {
        cmp = (cmp.wrapping_neg());
    }
    return cmp;
}

pub fn bid64_next_toward(mut x: u64, mut y: BID_UINT128) -> (u64, u32) {
    let mut res: u64 = 0;
    let mut tmp1: u64 = 0;
    let mut tmp2: u64 = 0;
    let mut pfpsf: u32 = 0;
    let mut res1: i64 = 0;
    let mut res2: i64 = 0;
    let mut yd = bid128_decode(y.w[1], y.w[0]);
    if ((x & 0x7c00000000000000) == 0x7c00000000000000) {
        if ((x & 0x0003ffffffffffff) > 999999999999999) {
            x = (x & 0xfe00000000000000);
        } else {
            x = (x & 0xfe03ffffffffffff);
        }
        if ((x & 0x7e00000000000000) == 0x7e00000000000000) {
            pfpsf |= 1;
            res = (x & 0xfdffffffffffffff);
        } else {
            if yd.isSNaN {
                pfpsf |= 1;
            }
            res = x;
        }
        return (res, pfpsf);
    } else if yd.isNaN {
        (res, pfpsf) = bid128_na_n_to_bid64(y.w[1], y.w[0]);
        return (res, pfpsf);
    } else {
        if ((x & 0x7800000000000000) == 0x7800000000000000) {
            x = (x & (0x8000000000000000 | 0x7800000000000000));
        }
    }
    if ((x & 0x7800000000000000) != 0x7800000000000000) {
        x = bid64_canonicalize_non_canonical_finite(x);
    }
    res2 = bid64_compare_to_bid128(x, y);
    if (res2 == 0) {
        res = ((y.w[1] & 0x8000000000000000) | (x & 0x7fffffffffffffff));
    } else if (res2 > 0) {
        (res, _) = bid64_next_down(x);
    } else {
        (res, _) = bid64_next_up(x);
    }
    if ((((x & 0x7800000000000000) != 0x7800000000000000)) && (((res & 0x7800000000000000) == 0x7800000000000000))) {
        pfpsf |= 32;
        pfpsf |= 8;
    }
    tmp1 = 0x00038d7ea4c68000;
    tmp2 = (res & 0x7fffffffffffffff);
    (res1, _) = bid64_quiet_greater(tmp1, tmp2);
    (res2, _) = bid64_quiet_not_equal(x, res);
    if ((res1 != 0) && (res2 != 0)) {
        pfpsf |= 32;
        pfpsf |= 16;
    }
    return (res, pfpsf);
}

pub fn bid64_nexttoward(mut x: u64, mut y: BID_UINT128) -> (u64, u32) {
    bid64_next_toward(x, y)
}
