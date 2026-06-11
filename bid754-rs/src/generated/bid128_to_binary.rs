// Auto-generated from bid128_to_binary.go by go2rs. Do not edit.

use super::prelude::*;

pub(crate) fn clz64_nz(mut n: u64) -> i64 {
    return (n).leading_zeros() as i64;
}

pub(crate) fn clz64(mut n: u64) -> i64 {
    if (n == 0) {
        return 64;
    }
    return clz64_nz(n);
}

pub(crate) fn clz128_nz(mut n_hi: u64, mut n_lo: u64) -> i64 {
    if (n_hi == 0) {
        return ((64 as i64).wrapping_add(clz64_nz(n_lo)));
    }
    return clz64_nz(n_hi);
}

pub(crate) fn sll128_short(mut hi: u64, mut lo: u64, mut c: u64) -> (u64, u64) {
    hi = (((go_checked_shl_u64(hi, go_shift_count_u64((c) as u64)))).wrapping_add(((go_checked_shr_u64(lo, go_shift_count_u64(((((64 as u64).wrapping_sub(c)))) as u64))))));
    lo = (go_checked_shl_u64(lo, go_shift_count_u64((c) as u64)));
    return (hi, lo);
}

pub(crate) fn sll128(mut hi: u64, mut lo: u64, mut c: u64) -> (u64, u64) {
    if (c == 0) {
        return (hi, lo);
    }
    if (c >= 64) {
        return ((go_checked_shl_u64(lo, go_shift_count_u64((((c.wrapping_sub(64)))) as u64))), 0);
    }
    return sll128_short(hi, lo, c);
}

pub(crate) fn srl256_short(mut x3: u64, mut x2: u64, mut x1: u64, mut x0: u64, mut c: u64) -> (u64, u64, u64, u64) {
    x0 = (((go_checked_shl_u64(x1, go_shift_count_u64(((((64 as u64).wrapping_sub(c)))) as u64)))).wrapping_add(((go_checked_shr_u64(x0, go_shift_count_u64((c) as u64))))));
    x1 = (((go_checked_shl_u64(x2, go_shift_count_u64(((((64 as u64).wrapping_sub(c)))) as u64)))).wrapping_add(((go_checked_shr_u64(x1, go_shift_count_u64((c) as u64))))));
    x2 = (((go_checked_shl_u64(x3, go_shift_count_u64(((((64 as u64).wrapping_sub(c)))) as u64)))).wrapping_add(((go_checked_shr_u64(x2, go_shift_count_u64((c) as u64))))));
    x3 = (go_checked_shr_u64(x3, go_shift_count_u64((c) as u64)));
    return (x3, x2, x1, x0);
}

pub(crate) fn lt128(mut x_hi: u64, mut x_lo: u64, mut y_hi: u64, mut y_lo: u64) -> bool {
    return ((x_hi < y_hi) || (((x_hi == y_hi) && (x_lo < y_lo))));
}

pub(crate) fn le128(mut x_hi: u64, mut x_lo: u64, mut y_hi: u64, mut y_lo: u64) -> bool {
    return ((x_hi < y_hi) || (((x_hi == y_hi) && (x_lo <= y_lo))));
}

pub(crate) fn __mul_128x256_to_384(mut A: BID_UINT128, mut B: BID_UINT256) -> BID_UINT384 {
    let mut P: BID_UINT384 = BID_UINT384 { w: [0, 0, 0, 0, 0, 0] };
    let mut CY: u64 = 0;
    let mut P0 = __mul_64x256_to_320(A.w[0], B);
    let mut P1 = __mul_64x256_to_320(A.w[1], B);
    P.w[0] = P0.w[0];
    (P.w[1], CY) = __add_carry_out(P1.w[0], P0.w[1]);
    (P.w[2], CY) = __add_carry_in_out(P1.w[1], P0.w[2], CY);
    (P.w[3], CY) = __add_carry_in_out(P1.w[2], P0.w[3], CY);
    (P.w[4], CY) = __add_carry_in_out(P1.w[3], P0.w[4], CY);
    P.w[5] = (P1.w[4].wrapping_add(CY));
    return P;
}

pub(crate) fn unpack_bid128_binarydecimal(mut x: BID_UINT128) -> (i64, i64, i64, BID_UINT128, bool, bool, bool, u64, u64, bool) {
    let mut s: i64 = 0;
    let mut e: i64 = 0;
    let mut k: i64 = 0;
    let mut c: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut isZero: bool = false;
    let mut isInf: bool = false;
    let mut isNaN: bool = false;
    let mut nanPayloadHi: u64 = 0;
    let mut nanPayloadLo: u64 = 0;
    let mut isSNaN: bool = false;
    s = ((go_checked_shr_u64(x.w[1], go_shift_count_u64((63) as u64))) as i64);
    if (((x.w[1] & (3 << 61))) == (3 << 61)) {
        if (((x.w[1] & (0xF << 59))) == (0xF << 59)) {
            if (((x.w[1] & (0x1F << 58))) != (0x1F << 58)) {
                isInf = true;
                return (s, e, k, c, isZero, isInf, isNaN, nanPayloadHi, nanPayloadLo, isSNaN);
            }
            if (((x.w[1] & (1 << 57))) != 0) {
                isSNaN = true;
            }
            isNaN = true;
            if lt128(54210108624275, 4089650035136921599, (x.w[1] & 0x3FFFFFFFFFFF), x.w[0]) {
                nanPayloadHi = 0;
                nanPayloadLo = 0;
            } else {
                nanPayloadHi = (((go_checked_shl_u64(x.w[1], go_shift_count_u64((18) as u64)))).wrapping_add(((go_checked_shr_u64(x.w[0], go_shift_count_u64((46) as u64))))));
                nanPayloadLo = (go_checked_shl_u64(x.w[0], go_shift_count_u64((18) as u64)));
            }
            return (s, e, k, c, isZero, isInf, isNaN, nanPayloadHi, nanPayloadLo, isSNaN);
        }
        isZero = true;
        return (s, e, k, c, isZero, isInf, isNaN, nanPayloadHi, nanPayloadLo, isSNaN);
    }
    e = (((((go_checked_shr_u64(x.w[1], go_shift_count_u64((49) as u64)))) & (((1 << 14) - 1))) as i64).wrapping_sub(6176));
    c.w[1] = (x.w[1] & (((1 << 49) - 1)));
    c.w[0] = x.w[0];
    if lt128(542101086242752, 4003012203950112767, c.w[1], c.w[0]) {
        c.w[1] = 0;
        c.w[0] = 0;
    }
    if ((c.w[1] == 0) && (c.w[0] == 0)) {
        isZero = true;
        return (s, e, k, c, isZero, isInf, isNaN, nanPayloadHi, nanPayloadLo, isSNaN);
    }
    k = (clz128_nz(c.w[1], c.w[0]).wrapping_sub(15));
    (c.w[1], c.w[0]) = sll128(c.w[1], c.w[0], (k as u64));
    return (s, e, k, c, isZero, isInf, isNaN, nanPayloadHi, nanPayloadLo, isSNaN);
}

pub(crate) fn return_binary32_pack(mut s: i64, mut e: i64, mut c: u64) -> f32 {
    let mut bits = ((((go_checked_shl_u32((s as u32), go_shift_count_u64((31) as u64)))).wrapping_add(((go_checked_shl_u32((e as u32), go_shift_count_u64((23) as u64)))))).wrapping_add(c as u32));
    return f32::from_bits(bits);
}

pub(crate) fn return_binary64_pack(mut s: i64, mut e: i64, mut c: u64) -> f64 {
    let mut bits = ((((go_checked_shl_u64((s as u64), go_shift_count_u64((63) as u64)))).wrapping_add(((go_checked_shl_u64((e as u64), go_shift_count_u64((52) as u64)))))).wrapping_add(c));
    return f64::from_bits(bits);
}

pub fn bid128_to_binary32(mut x: BID_UINT128, mut rnd_mode: i64, pfpsf: &mut u32) -> f32 {
    let mut c_prov: u64 = 0;
    let mut c: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut m_min: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut e_out: i64 = 0;
    let mut r: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut z: BID_UINT384 = BID_UINT384 { w: [0, 0, 0, 0, 0, 0] };
    let (mut s, mut e, mut k, mut c, mut isZero, mut isInf, mut isNaN, mut nanPayloadHi, mut nanPayloadLo, mut isSNaN) = unpack_bid128_binarydecimal(x);
    if isZero {
        return return_binary32_pack(s, 0, 0);
    }
    if isInf {
        return return_binary32_pack(s, 255, 0);
    }
    if isNaN {
        if isSNaN {
            (*pfpsf) |= 1;
        }
        _ = nanPayloadLo;
        return return_binary32_pack(s, 255, (((go_checked_shr_u64(nanPayloadHi, go_shift_count_u64((42) as u64)))).wrapping_add(1 << 22)));
    }
    if (e >= 39) {
        (*pfpsf) |= (8 | 32);
        if ((rnd_mode == 3) || ((rnd_mode == bool_to_rnd_mode(s != 0)))) {
            return return_binary32_pack(s, 254, ((1 << 23) - 1));
        }
        return return_binary32_pack(s, 255, 0);
    }
    if (e <= (-80)) {
        e = (-80);
    }
    m_min = bid_breakpoints_binary32[(e.wrapping_add(80)) as usize];
    e_out = ((bid_exponents_binary32[(e.wrapping_add(80)) as usize] as i64).wrapping_sub(k));
    if le128(c.w[1], c.w[0], m_min.w[1], m_min.w[0]) {
        r = bid_multipliers1_binary32[(e.wrapping_add(80)) as usize];
    } else {
        r = bid_multipliers2_binary32[(e.wrapping_add(80)) as usize];
        e_out = (e_out.wrapping_add(1));
    }
    z = __mul_128x256_to_384(c, r);
    if (e_out < 1) {
        let mut d = ((1 as i64).wrapping_sub(e_out));
        if (d > 26) {
            d = 26;
        }
        e_out = 1;
        (z.w[5], z.w[4], z.w[3], z.w[2]) = srl256_short(z.w[5], z.w[4], z.w[3], z.w[2], (d as u64));
    }
    c_prov = z.w[5];
    let mut rbIdx = ((((go_checked_shl_i64(rnd_mode, go_shift_count_u64((2) as u64)))).wrapping_add(((go_checked_shl_i64((s & 1), go_shift_count_u64((1) as u64)))))).wrapping_add(((c_prov & 1) as i64)));
    if lt128(bid_roundbound_128[rbIdx as usize].w[1], bid_roundbound_128[rbIdx as usize].w[0], z.w[4], z.w[3]) {
        c_prov = (c_prov.wrapping_add(1));
        if (c_prov == (1 << 24)) {
            c_prov = (1 << 23);
            e_out = (e_out.wrapping_add(1));
        } else if (((c_prov == (1 << 23))) && (e_out == 1)) {
            if ((((((rnd_mode & 3) == 0)) && ((z.w[4] < (3 << 62))))) || (((((rnd_mode.wrapping_add(((s & 1) as i64))) == 2)) && ((z.w[4] < (1 << 63)))))) {
                (*pfpsf) |= 16;
            }
        }
    }
    if (e_out >= 255) {
        (*pfpsf) |= (8 | 32);
        if ((rnd_mode == 3) || ((rnd_mode == bool_to_rnd_mode(s != 0)))) {
            return return_binary32_pack(s, 254, ((1 << 23) - 1));
        }
        return return_binary32_pack(s, 255, 0);
    }
    if (c_prov < (1 << 23)) {
        e_out = 0;
    } else {
        c_prov = (c_prov & (((1 << 23) - 1)));
    }
    if ((z.w[4] != 0) || (z.w[3] != 0)) {
        (*pfpsf) |= 32;
        if (e_out == 0) {
            (*pfpsf) |= 16;
        }
    }
    return return_binary32_pack(s, e_out, c_prov);
}

pub fn bid128_to_binary64(mut x: BID_UINT128, mut rnd_mode: i64, pfpsf: &mut u32) -> f64 {
    let mut c_prov: u64 = 0;
    let mut c: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut m_min: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut e_out: i64 = 0;
    let mut r: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut z: BID_UINT384 = BID_UINT384 { w: [0, 0, 0, 0, 0, 0] };
    let (mut s, mut e, mut k, mut c, mut isZero, mut isInf, mut isNaN, mut nanPayloadHi, mut nanPayloadLo, mut isSNaN) = unpack_bid128_binarydecimal(x);
    if isZero {
        return return_binary64_pack(s, 0, 0);
    }
    if isInf {
        return return_binary64_pack(s, 2047, 0);
    }
    if isNaN {
        if isSNaN {
            (*pfpsf) |= 1;
        }
        _ = nanPayloadLo;
        return return_binary64_pack(s, 2047, (((go_checked_shr_u64(nanPayloadHi, go_shift_count_u64((13) as u64)))).wrapping_add(1 << 51)));
    }
    (c.w[1], c.w[0]) = sll128_short(c.w[1], c.w[0], 6);
    if (e >= 309) {
        (*pfpsf) |= (8 | 32);
        if ((rnd_mode == 3) || ((rnd_mode == bool_to_rnd_mode(s != 0)))) {
            return return_binary64_pack(s, 2046, ((1 << 52) - 1));
        }
        return return_binary64_pack(s, 2047, 0);
    }
    if (e <= (-358)) {
        e = (-358);
    }
    m_min = bid_breakpoints_binary64[(e.wrapping_add(358)) as usize];
    e_out = ((bid_exponents_binary64[(e.wrapping_add(358)) as usize] as i64).wrapping_sub(k));
    if le128(c.w[1], c.w[0], m_min.w[1], m_min.w[0]) {
        r = bid_multipliers1_binary64[(e.wrapping_add(358)) as usize];
    } else {
        r = bid_multipliers2_binary64[(e.wrapping_add(358)) as usize];
        e_out = (e_out.wrapping_add(1));
    }
    z = __mul_128x256_to_384(c, r);
    if (e_out < 1) {
        let mut d = ((1 as i64).wrapping_sub(e_out));
        if (d > 55) {
            d = 55;
        }
        e_out = 1;
        (z.w[5], z.w[4], z.w[3], z.w[2]) = srl256_short(z.w[5], z.w[4], z.w[3], z.w[2], (d as u64));
    }
    c_prov = z.w[5];
    let mut rbIdx = ((((go_checked_shl_i64(rnd_mode, go_shift_count_u64((2) as u64)))).wrapping_add(((go_checked_shl_i64((s & 1), go_shift_count_u64((1) as u64)))))).wrapping_add(((c_prov & 1) as i64)));
    if lt128(bid_roundbound_128[rbIdx as usize].w[1], bid_roundbound_128[rbIdx as usize].w[0], z.w[4], z.w[3]) {
        c_prov = (c_prov.wrapping_add(1));
        if (c_prov == (1 << 53)) {
            c_prov = (1 << 52);
            e_out = (e_out.wrapping_add(1));
        } else if (((c_prov == (1 << 52))) && (e_out == 1)) {
            if ((((((rnd_mode & 3) == 0)) && ((z.w[4] < (3 << 62))))) || (((((rnd_mode.wrapping_add(((s & 1) as i64))) == 2)) && ((z.w[4] < (1 << 63)))))) {
                (*pfpsf) |= 16;
            }
        }
    }
    if (e_out >= 2047) {
        (*pfpsf) |= (8 | 32);
        if ((rnd_mode == 3) || ((rnd_mode == bool_to_rnd_mode(s != 0)))) {
            return return_binary64_pack(s, 2046, ((1 << 52) - 1));
        }
        return return_binary64_pack(s, 2047, 0);
    }
    if (c_prov < (1 << 52)) {
        e_out = 0;
    } else {
        c_prov = (c_prov & (((1 << 52) - 1)));
    }
    if ((z.w[4] != 0) || (z.w[3] != 0)) {
        (*pfpsf) |= 32;
        if (e_out == 0) {
            (*pfpsf) |= 16;
        }
    }
    return return_binary64_pack(s, e_out, c_prov);
}

pub(crate) fn bool_to_rnd_mode(mut neg: bool) -> i64 {
    if neg {
        return 2;
    }
    return 1;
}
