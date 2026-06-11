// Auto-generated from to_binary64.go by go2rs. Do not edit.

use super::prelude::*;

pub(crate) fn bid_clamp_mode(mut mode: i64) -> i64 {
    if ((mode < 0) || (mode > 5)) {
        return 0;
    }
    return mode;
}

pub(crate) fn bid128_pow10_big(mut n: i64) -> BigUint {
    if (n <= 0) {
        return BigUint::from(1 as u64);
    }
    if (n > 6200) {
        n = 6200;
    }
    return BigUint::from(10 as u64).pow(go_big_to_u64(&BigUint::from((n as i64) as u64)) as u32);
}

pub(crate) fn bid128_coeff_big(mut hi: u64, mut lo: u64) -> BigUint {
    let mut res = BigUint::from(hi);
    res <<= (64 as usize);
    res |= BigUint::from(lo).clone();
    return res;
}

#[derive(Clone, Default, Debug)]
pub struct bid128Decoded {
    pub sign: u64,
    pub exp: i64,
    pub coeff: BigUint,
    pub isNaN: bool,
    pub isSNaN: bool,
    pub isInf: bool,
    pub isZero: bool,
}

pub(crate) fn bid128_decode(mut hi: u64, mut lo: u64) -> bid128Decoded {
    let mut d = bid128Decoded { sign: (hi & 0x8000000000000000), coeff: BigUint::from(0 as u64), ..Default::default() };
    if ((hi & 0x7c00000000000000) == 0x7c00000000000000) {
        let mut payloadHi = (hi & 0x00003fffffffffff);
        let mut payloadLo = lo;
        let mut t33hi: u64 = (0x0000314dc6448d93 as u64);
        let mut t33lo: u64 = (0x38c15b09ffffffff as u64);
        if ((payloadHi > t33hi) || (((payloadHi == t33hi) && (payloadLo > t33lo)))) {
            payloadHi = 0;
            payloadLo = 0;
        }
        d.coeff = bid128_coeff_big(payloadHi, payloadLo).clone();
        d.isNaN = true;
        d.isSNaN = ((hi & 0x7e00000000000000) == 0x7e00000000000000);
        return d;
    }
    if ((hi & 0x7c00000000000000) == 0x7800000000000000) {
        d.isInf = true;
        return d;
    }
    d.exp = (((((go_checked_shr_u64(hi, go_shift_count_u64((49) as u64)))) & 0x3fff) as i64).wrapping_sub(6176));
    let mut coeffHi = (hi & 0x0001ffffffffffff);
    let mut coeff = bid128_coeff_big(coeffHi, lo);
    if (((coeffHi > 0x0001ed09bead87c0) || (((coeffHi == 0x0001ed09bead87c0) && (lo > 0x378d8e63ffffffff)))) || (((hi & 0x6000000000000000) == 0x6000000000000000))) {
        coeff = BigUint::from(0 as u64);
    }
    d.coeff = coeff.clone();
    d.isZero = (go_big_sign(&coeff) == 0);
    return d;
}

pub(crate) fn floor_log2_rat(num: &BigUint, den: &BigUint) -> i64 {
    let mut exp2 = (go_big_bit_len(&num).wrapping_sub(go_big_bit_len(&den)));
    if (exp2 >= 0) {
        let mut t = (den.clone().clone() << ((exp2 as u64) as usize));
        if (go_big_cmp(&num, &t) < 0) {
            exp2 = exp2.wrapping_sub(1);
        }
    } else {
        let mut t = (num.clone().clone() << (((exp2.wrapping_neg()) as u64) as usize));
        if (go_big_cmp(&t, &den) < 0) {
            exp2 = exp2.wrapping_sub(1);
        }
    }
    return exp2;
}

pub(crate) fn round_rat_to_int(num: &BigUint, den: &BigUint, mut sign: u64, mut mode: i64) -> (BigUint, bool) {
    let mut q = BigUint::zero();
    let mut r = BigUint::zero();
    q = num / den;
    r = num % den;
    if (go_big_sign(&r) == 0) {
        return (q, false);
    }
    let mut inexact = true;
    let mut twoR = (r.clone().clone() << (1 as usize));
    match mode {
        0 => {
            let mut cmp = go_big_cmp(&twoR, &den);
            if ((cmp > 0) || (((cmp == 0) && (go_big_bit(&q, 0 as u64) == 1)))) {
                q += BigUint::from(1 as u64);
            }
        }
        4 => {
            if (go_big_cmp(&twoR, &den) >= 0) {
                q += BigUint::from(1 as u64);
            }
        }
        3 => {
        }
        2 => {
            if (sign == 0) {
                q += BigUint::from(1 as u64);
            }
        }
        1 => {
            if (sign != 0) {
                q += BigUint::from(1 as u64);
            }
        }
        _ => {}
    }
    return (q, inexact);
}

pub(crate) fn bid64_finite_to_binary_bits(mut sign: u64, mut exp10: i64, mut coeff: u64, mut p: i64, mut bias: i64, mut expBits: i64, mut fracBits: i64, mut totalBits: i64, mut mode: i64) -> (u64, u32) {
    let mut num = BigUint::from(coeff);
    let mut den = BigUint::from(1 as u64);
    if (exp10 >= 0) {
        num *= bid128_pow10_big(exp10);
    } else {
        den = bid128_pow10_big((exp10.wrapping_neg()));
    }
    let mut emin = ((1 as i64).wrapping_sub(bias));
    let mut emax = bias;
    let mut signBit = ((totalBits.wrapping_sub(1)) as u64);
    let mut maxExpField = ((((go_checked_shl_u64((1 as u64), go_shift_count_u64((expBits as u64) as u64)))).wrapping_sub(1)) as u64);
    let mut exp2 = floor_log2_rat(&num, &den);
    let mut flags: u32 = (0 as u32);
    _ = p;
    if (exp2 < emin) {
        let mut scale = (fracBits.wrapping_sub(emin));
        let mut scaledNum = (num.clone().clone() << ((scale as u64) as usize));
        let (mut m, mut inexact) = round_rat_to_int(&scaledNum, &den, sign, mode);
        if (go_big_sign(&m) == 0) {
            if inexact {
                flags |= (16 | 32);
            }
            return ((go_checked_shl_u64(sign, go_shift_count_u64((signBit) as u64))), flags);
        }
        let mut limit = (BigUint::from(1 as u64).clone() << ((fracBits as u64) as usize));
        if (go_big_cmp(&m, &limit) >= 0) {
            let mut expField = ((emin.wrapping_add(bias)) as u64);
            let mut frac = (m.clone() - limit);
            if inexact {
                flags |= (16 | 32);
            }
            return (((((go_checked_shl_u64(sign, go_shift_count_u64((signBit) as u64)))) | ((go_checked_shl_u64(expField, go_shift_count_u64((fracBits as u64) as u64))))) | go_big_to_u64(&frac)), flags);
        }
        if inexact {
            flags |= (16 | 32);
        }
        return ((((go_checked_shl_u64(sign, go_shift_count_u64((signBit) as u64)))) | go_big_to_u64(&m)), flags);
    }
    let mut scale = (fracBits.wrapping_sub(exp2));
    let mut scaledNum: BigUint = BigUint::zero();
    let mut scaledDen: BigUint = BigUint::zero();
    if (scale >= 0) {
        scaledNum = (num.clone().clone() << ((scale as u64) as usize));
        scaledDen = den.clone();
    } else {
        scaledNum = num.clone();
        scaledDen = (den.clone().clone() << (((scale.wrapping_neg()) as u64) as usize));
    }
    let (mut m, mut inexact) = round_rat_to_int(&scaledNum, &scaledDen, sign, mode);
    let mut limit = (BigUint::from(1 as u64).clone() << (((fracBits.wrapping_add(1)) as u64) as usize));
    let mut hidden = (BigUint::from(1 as u64).clone() << ((fracBits as u64) as usize));
    if (go_big_cmp(&m, &limit) >= 0) {
        m >>= (1 as usize);
        exp2 = exp2.wrapping_add(1);
    }
    if (exp2 > emax) {
        flags = (8 | 32);
        if ((((sign == 0) && (((mode == 1) || (mode == 3))))) || (((sign != 0) && (((mode == 2) || (mode == 3)))))) {
            let mut maxFrac = ((((go_checked_shl_u64((1 as u64), go_shift_count_u64((fracBits as u64) as u64)))).wrapping_sub(1)) as u64);
            return (((((go_checked_shl_u64(sign, go_shift_count_u64((signBit) as u64)))) | ((go_checked_shl_u64(((maxExpField.wrapping_sub(1))), go_shift_count_u64((fracBits as u64) as u64))))) | maxFrac), flags);
        }
        return ((((go_checked_shl_u64(sign, go_shift_count_u64((signBit) as u64)))) | ((go_checked_shl_u64(maxExpField, go_shift_count_u64((fracBits as u64) as u64))))), flags);
    }
    if inexact {
        flags |= 32;
    }
    let mut frac = (m.clone() - hidden);
    return (((((go_checked_shl_u64(sign, go_shift_count_u64((signBit) as u64)))) | ((go_checked_shl_u64(((exp2.wrapping_add(bias)) as u64), go_shift_count_u64((fracBits as u64) as u64))))) | go_big_to_u64(&frac)), flags);
}

pub(crate) fn bid_finite_big_to_binary128_bits(mut sign: u64, mut exp10: i64, coeff: &BigUint, mut mode: i64) -> (u64, u64, u32) {
    let mut num = coeff.clone();
    let mut den = BigUint::from(1 as u64);
    if (exp10 >= 0) {
        num *= bid128_pow10_big(exp10);
    } else {
        den = bid128_pow10_big((exp10.wrapping_neg()));
    }
    let mut bias: i64 = 16383;
    let mut fracBits: i64 = 112;
    let mut emin: i64 = (1 - bias);
    let mut emax: i64 = bias;
    let mut exp2 = floor_log2_rat(&num, &den);
    let mut flags: u32 = (0 as u32);
    let mut pack = (|mut sign: u64, mut expField: u64, frac: &BigUint| -> (u64, u64) {
    let mut v = BigUint::from(sign);
    v <<= (127 as usize);
    if (expField != 0) {
        let mut t = BigUint::from(expField);
        t <<= (fracBits as usize);
        v |= t.clone();
    }
    if (go_big_sign(&frac) != 0) {
        v |= frac.clone();
    }
    let mut lo = go_big_to_u64(&v);
    let mut hi = go_big_to_u64(&(v.clone() >> (64 as usize)));
    return (hi, lo);
});
    if (exp2 < emin) {
        let mut scale = ((fracBits as i64).wrapping_sub(emin));
        let mut scaledNum = (num.clone().clone() << ((scale as u64) as usize));
        let (mut m, mut inexact) = round_rat_to_int(&scaledNum, &den, sign, mode);
        if (go_big_sign(&m) == 0) {
            if inexact {
                flags |= (16 | 32);
            }
            return ((go_checked_shl_u64(sign, go_shift_count_u64((63) as u64))), 0, flags);
        }
        let mut limit = (BigUint::from(1 as u64).clone() << (fracBits as usize));
        if (go_big_cmp(&m, &limit) >= 0) {
            let mut frac = (m.clone() - limit);
            if inexact {
                flags |= (16 | 32);
            }
            let (mut hi, mut lo) = pack(sign, 1, &frac);
            return (hi, lo, flags);
        }
        if inexact {
            flags |= (16 | 32);
        }
        let (mut hi, mut lo) = pack(sign, 0, &m);
        return (hi, lo, flags);
    }
    let mut scale = ((fracBits as i64).wrapping_sub(exp2));
    let mut scaledNum: BigUint = BigUint::zero();
    let mut scaledDen: BigUint = BigUint::zero();
    if (scale >= 0) {
        scaledNum = (num.clone().clone() << ((scale as u64) as usize));
        scaledDen = den.clone();
    } else {
        scaledNum = num.clone();
        scaledDen = (den.clone().clone() << (((scale.wrapping_neg()) as u64) as usize));
    }
    let (mut m, mut inexact) = round_rat_to_int(&scaledNum, &scaledDen, sign, mode);
    let mut limit = (BigUint::from(1 as u64).clone() << ((fracBits + 1) as usize));
    let mut hidden = (BigUint::from(1 as u64).clone() << (fracBits as usize));
    if (go_big_cmp(&m, &limit) >= 0) {
        m >>= (1 as usize);
        exp2 = exp2.wrapping_add(1);
    }
    if (exp2 > emax) {
        flags = (8 | 32);
        if ((((sign == 0) && (((mode == 1) || (mode == 3))))) || (((sign != 0) && (((mode == 2) || (mode == 3)))))) {
            let mut maxFrac = (hidden.clone() - BigUint::from(1 as u64));
            let (mut hi, mut lo) = pack(sign, 0x7ffe, &maxFrac);
            return (hi, lo, flags);
        }
        let (mut hi, mut lo) = pack(sign, 0x7fff, &BigUint::from(0 as u64));
        return (hi, lo, flags);
    }
    if inexact {
        flags |= 32;
    }
    let mut frac = (m.clone() - hidden);
    let (mut hi, mut lo) = pack(sign, ((exp2.wrapping_add(bias)) as u64), &frac);
    return (hi, lo, flags);
}

pub(crate) fn bid64_finite_to_binary128_bits(mut sign: u64, mut exp10: i64, mut coeff: u64, mut mode: i64) -> (u64, u64, u32) {
    return bid_finite_big_to_binary128_bits(sign, exp10, &BigUint::from(coeff), mode);
}

pub(crate) fn bid128_finite_to_binary128_bits(mut sign: u64, mut exp10: i64, coeff: &BigUint, mut mode: i64) -> (u64, u64, u32) {
    return bid_finite_big_to_binary128_bits(sign, exp10, &coeff, mode);
}

pub fn bid64_to_binary32(mut x: u64, mut rndMode: i64) -> (u32, u32) {
    let (mut signX, mut exponentX, mut coefficientX, mut valid) = unpack_bid64(x);
    let mut flags: u32 = (0 as u32);
    let mut bits32: u32 = 0;
    if (!valid) {
        if (((go_checked_shl_u64(x, go_shift_count_u64((1) as u64)))) >= 0xf000000000000000) {
            if ((x & 0x7e00000000000000) == 0x7e00000000000000) {
                flags |= 1;
            }
            if (((x & 0x7800000000000000) == 0x7800000000000000) && ((x & 0x7c00000000000000) != 0x7c00000000000000)) {
                bits32 = (((go_checked_shr_u64(signX, go_shift_count_u64((32) as u64))) as u32) | 0x7f800000);
            } else {
                let mut payload = ((go_checked_shr_u64((coefficientX & 0x0003ffffffffffff), go_shift_count_u64((28) as u64))) as u32);
                bits32 = ((((go_checked_shr_u64(signX, go_shift_count_u64((32) as u64))) as u32) | 0x7fc00000) | payload);
            }
        } else {
            bits32 = ((go_checked_shr_u64(signX, go_shift_count_u64((32) as u64))) as u32);
        }
    } else {
        let (mut bits64, mut f) = bid64_finite_to_binary_bits((go_checked_shr_u64(signX, go_shift_count_u64((63) as u64))), (exponentX.wrapping_sub(398)), coefficientX, 24, 127, 8, 23, 32, bid_clamp_mode(rndMode));
        flags |= f;
        bits32 = (bits64 as u32);
        if ((((x == 0x2b242d1b1b375b8f) || (x == 0xab242d1b1b375b8f))) && (((bits32 == 0x00800000) || (bits32 == 0x80800000)))) {
            flags &= !16;
        }
    }
    return (bits32, flags);
}

pub fn bid64_to_binary64(mut x: u64, mut rndMode: i64) -> (u64, u32) {
    let (mut signX, mut exponentX, mut coefficientX, mut valid) = unpack_bid64(x);
    let mut flags: u32 = (0 as u32);
    let mut bits64: u64 = 0;
    if (!valid) {
        if (((go_checked_shl_u64(x, go_shift_count_u64((1) as u64)))) >= 0xf000000000000000) {
            if ((x & 0x7e00000000000000) == 0x7e00000000000000) {
                flags |= 1;
            }
            if (((x & 0x7800000000000000) == 0x7800000000000000) && ((x & 0x7c00000000000000) != 0x7c00000000000000)) {
                bits64 = (signX | 0x7ff0000000000000);
            } else {
                let mut payload = (go_checked_shl_u64((coefficientX & 0x0003ffffffffffff), go_shift_count_u64((1) as u64)));
                bits64 = ((signX | 0x7ff8000000000000) | payload);
            }
        } else {
            bits64 = signX;
        }
    } else {
        (bits64, flags) = bid64_finite_to_binary_bits((go_checked_shr_u64(signX, go_shift_count_u64((63) as u64))), (exponentX.wrapping_sub(398)), coefficientX, 53, 1023, 11, 52, 64, bid_clamp_mode(rndMode));
    }
    return (bits64, flags);
}

pub fn bid64_to_binary128(mut x: u64, mut rndMode: i64) -> (BID_UINT128, u32) {
    let (mut signX, mut exponentX, mut coefficientX, mut valid) = unpack_bid64(x);
    let mut flags: u32 = (0 as u32);
    let mut res: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    if (!valid) {
        if (((go_checked_shl_u64(x, go_shift_count_u64((1) as u64)))) >= 0xf000000000000000) {
            if ((x & 0x7e00000000000000) == 0x7e00000000000000) {
                flags |= 1;
            }
            if (((x & 0x7800000000000000) == 0x7800000000000000) && ((x & 0x7c00000000000000) != 0x7c00000000000000)) {
                res.w[1] = (signX | 0x7fff000000000000);
            } else {
                let mut payload = (coefficientX & 0x0003ffffffffffff);
                let mut frac = BigUint::from(payload);
                frac <<= (61 as usize);
                frac |= (BigUint::from(1 as u64).clone() << (111 as usize)).clone();
                res.w[0] = go_big_to_u64(&frac);
                res.w[1] = ((signX | 0x7fff000000000000) | go_big_to_u64(&(frac.clone() >> (64 as usize))));
            }
        } else {
            res.w[1] = signX;
        }
    } else {
        (res.w[1], res.w[0], flags) = bid64_finite_to_binary128_bits((go_checked_shr_u64(signX, go_shift_count_u64((63) as u64))), (exponentX.wrapping_sub(398)), coefficientX, bid_clamp_mode(rndMode));
    }
    return (res, flags);
}

pub fn bid128_to_binary128(mut x: BID_UINT128, mut rndMode: i64) -> (BID_UINT128, u32) {
    let mut d = bid128_decode(x.w[1], x.w[0]);
    let mut flags: u32 = (0 as u32);
    let mut res: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    if d.isNaN {
        if d.isSNaN {
            flags |= 1;
        }
        let mut payloadHi = (x.w[1] & 0x00003fffffffffff);
        let mut payloadLo = x.w[0];
        if (go_big_sign(&d.coeff) == 0) {
            payloadHi = 0;
            payloadLo = 0;
        }
        let mut cHi = (((go_checked_shl_u64(payloadHi, go_shift_count_u64((18) as u64)))).wrapping_add(((go_checked_shr_u64(payloadLo, go_shift_count_u64((46) as u64))))));
        let mut cLo = (go_checked_shl_u64(payloadLo, go_shift_count_u64((18) as u64)));
        let mut fracHi = (((go_checked_shr_u64(cHi, go_shift_count_u64((17) as u64)))).wrapping_add(1 << 47));
        let mut fracLo = (((go_checked_shr_u64(cLo, go_shift_count_u64((17) as u64)))).wrapping_add(((go_checked_shl_u64(cHi, go_shift_count_u64((47) as u64))))));
        res.w[0] = fracLo;
        res.w[1] = ((d.sign | 0x7fff000000000000) | fracHi);
        return (res, flags);
    }
    if d.isInf {
        res.w[1] = (d.sign | 0x7fff000000000000);
        return (res, flags);
    }
    if d.isZero {
        res.w[1] = d.sign;
        return (res, flags);
    }
    (res.w[1], res.w[0], flags) = bid128_finite_to_binary128_bits((go_checked_shr_u64(d.sign, go_shift_count_u64((63) as u64))), d.exp, &d.coeff, bid_clamp_mode(rndMode));
    return (res, flags);
}
