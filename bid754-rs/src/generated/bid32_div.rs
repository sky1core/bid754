// Auto-generated from bid32_div.go by go2rs. Do not edit.

use super::prelude::*;

pub(crate) fn bid32_div_pure(mut x: u32, mut y: u32, mut rndMode: i64) -> u32 {
    let mut CA: u64 = 0;
    let mut sign_x: u32 = 0;
    let mut sign_y: u32 = 0;
    let mut coefficient_x: u32 = 0;
    let mut coefficient_y: u32 = 0;
    let mut A: u32 = 0;
    let mut B: u32 = 0;
    let mut Q: u32 = 0;
    let mut Q2: u32 = 0;
    let mut B2: u32 = 0;
    let mut B4: u32 = 0;
    let mut B5: u32 = 0;
    let mut R: u32 = 0;
    let mut T: u32 = 0;
    let mut DU: u32 = 0;
    let mut res: u32 = 0;
    let mut valid_x: bool = false;
    let mut valid_y: bool = false;
    let mut D: u32 = 0;
    let mut exponent_x: i64 = 0;
    let mut exponent_y: i64 = 0;
    let mut bin_expon_cx: i64 = 0;
    let mut diff_expon: i64 = 0;
    let mut ed1: i64 = 0;
    let mut ed2: i64 = 0;
    let mut bin_index: i64 = 0;
    let mut rmode: i64 = 0;
    let mut amount: i64 = 0;
    let mut nzeros: i64 = 0;
    let mut i: i64 = 0;
    let mut j: i64 = 0;
    let mut d5: i64 = 0;
    let mut digit_h: u32 = 0;
    let mut digit_low: u32 = 0;
    (sign_x, exponent_x, coefficient_x, valid_x) = unpack_bid32_add(x);
    (sign_y, exponent_y, coefficient_y, valid_y) = unpack_bid32_add(y);
    if (!valid_x) {
        if ((x & 0x7c000000) == 0x7c000000) {
            return (coefficient_x & 0xfdffffff);
        }
        if ((x & 0x78000000) == 0x78000000) {
            if ((y & 0x78000000) == 0x78000000) {
                if ((y & 0x7c000000) == 0x78000000) {
                    return 0x7c000000;
                }
            } else {
                return ((((x ^ y) & 0x80000000)) | 0x78000000);
            }
        }
        if ((((y & 0x78000000) != 0x78000000)) && (coefficient_y == 0)) {
            return 0x7c000000;
        }
        if ((y & 0x78000000) != 0x78000000) {
            if ((y & 0x60000000) == 0x60000000) {
                exponent_y = ((((go_checked_shr_u32(y, go_shift_count_u64((21) as u64)))) & 0xff) as i64);
            } else {
                exponent_y = ((((go_checked_shr_u32(y, go_shift_count_u64((23) as u64)))) & 0xff) as i64);
            }
            sign_y = (y & 0x80000000);
            exponent_x = ((exponent_x.wrapping_sub(exponent_y)).wrapping_add(101));
            if (exponent_x > 191) {
                exponent_x = 191;
            } else if (exponent_x < 0) {
                exponent_x = 0;
            }
            return ((sign_x ^ sign_y) | ((go_checked_shl_u32((exponent_x as u32), go_shift_count_u64((23) as u64)))));
        }
    }
    if (!valid_y) {
        if ((y & 0x7c000000) == 0x7c000000) {
            return (coefficient_y & 0xfdffffff);
        }
        if ((y & 0x78000000) == 0x78000000) {
            return ((x ^ y) & 0x80000000);
        }
        return ((sign_x ^ sign_y) | 0x78000000);
    }
    diff_expon = ((exponent_x.wrapping_sub(exponent_y)).wrapping_add(101));
    if (coefficient_x < coefficient_y) {
        let mut tempx = (coefficient_x as f32);
        let mut tempy = (coefficient_y as f32);
        bin_index = ((go_checked_shr_u32((((tempy as f32).to_bits().wrapping_sub((tempx as f32).to_bits()))), go_shift_count_u64((23) as u64))) as i64);
        A = (coefficient_x.wrapping_mul(bid_power10_index_binexp[bin_index as usize] as u32));
        B = coefficient_y;
        DU = (go_checked_shr_u32(((A.wrapping_sub(B))), go_shift_count_u64((31) as u64)));
        ed1 = ((6 as i64).wrapping_add(DU as i64));
        ed2 = ((bid_estimate_decimal_digits[bin_index as usize] as i64).wrapping_add(ed1));
        T = (bid_power10_table_128[ed1 as usize].w[0] as u32);
        CA = ((A as u64).wrapping_mul(T as u64));
        Q = 0;
        diff_expon = (diff_expon.wrapping_sub(ed2));
    } else {
        Q = (coefficient_x / coefficient_y);
        R = (coefficient_x.wrapping_sub((coefficient_y.wrapping_mul(Q))));
        let mut tempq = (Q as f32);
        bin_expon_cx = ((((go_checked_shr_u32((tempq as f32).to_bits(), go_shift_count_u64((23) as u64)))) as i64).wrapping_sub(0x7f));
        if (R == 0) {
            res = get_bid32((sign_x ^ sign_y), diff_expon, (Q as u64), rndMode);
            return res;
        }
        DU = (((bid_power10_index_binexp[bin_expon_cx as usize] as u32).wrapping_sub(Q)).wrapping_sub(1));
        DU = go_checked_shr_u32(DU, go_shift_count_u64((31) as u64));
        ed2 = (((7 as i64).wrapping_sub(bid_estimate_decimal_digits[bin_expon_cx as usize] as i64)).wrapping_sub(DU as i64));
        T = (bid_power10_table_128[ed2 as usize].w[0] as u32);
        CA = ((R as u64).wrapping_mul(T as u64));
        B = coefficient_y;
        Q = Q.wrapping_mul(bid_power10_table_128[ed2 as usize].w[0] as u32);
        diff_expon = diff_expon.wrapping_sub(ed2);
    }
    Q2 = ((CA / (B as u64)) as u32);
    B2 = (B.wrapping_add(B));
    B4 = (B2.wrapping_add(B2));
    R = ((CA.wrapping_sub(((Q2 as u64).wrapping_mul(B as u64)))) as u32);
    Q = Q.wrapping_add(Q2);
    if (R == 0) {
        if ((coefficient_x <= 1024) && (coefficient_y <= 1024)) {
            i = ((coefficient_y.wrapping_sub(1)) as i64);
            j = ((coefficient_x.wrapping_sub(1)) as i64);
            nzeros = ((ed2.wrapping_sub(bid_factors32[i as usize][0] as i64)).wrapping_add(bid_factors32[j as usize][0] as i64));
            d5 = ((ed2.wrapping_sub(bid_factors32[i as usize][1] as i64)).wrapping_add(bid_factors32[j as usize][1] as i64));
            if (d5 < nzeros) {
                nzeros = d5;
            }
            if (nzeros > 0) {
                let mut CT = ((Q as u64).wrapping_mul(bid_bid_reciprocals10_32[nzeros as usize]));
                CT = go_checked_shr_u64(CT, go_shift_count_u64((32) as u64));
                amount = (bid_bid_bid_recip_scale32[nzeros as usize] as i64);
                Q = ((go_checked_shr_u64(CT, go_shift_count_u64((amount as u64) as u64))) as u32);
                diff_expon = diff_expon.wrapping_add(nzeros);
            }
        } else {
            nzeros = 0;
            let mut PD = ((Q as u64).wrapping_mul(0x068DB8BB));
            digit_h = ((go_checked_shr_u64(PD, go_shift_count_u64((40) as u64))) as u32);
            digit_low = (Q.wrapping_sub((digit_h.wrapping_mul(10000))));
            if (digit_low == 0) {
                nzeros = nzeros.wrapping_add(4);
            } else {
                digit_h = digit_low;
            }
            if ((digit_h & 1) == 0) {
                nzeros = nzeros.wrapping_add(((3 & ((go_checked_shr_u8(bid_packed_10000_zeros[(go_checked_shr_u32(digit_h, go_shift_count_u64((3) as u64))) as usize], go_shift_count_u64((digit_h & 7) as u64))))) as i64));
            }
            if (nzeros > 0) {
                let mut CT = ((Q as u64).wrapping_mul(bid_bid_reciprocals10_32[nzeros as usize]));
                CT = go_checked_shr_u64(CT, go_shift_count_u64((32) as u64));
                amount = (bid_bid_bid_recip_scale32[nzeros as usize] as i64);
                Q = ((go_checked_shr_u64(CT, go_shift_count_u64((amount as u64) as u64))) as u32);
            }
            diff_expon = diff_expon.wrapping_add(nzeros);
        }
        if (diff_expon >= 0) {
            res = get_bid32((sign_x ^ sign_y), diff_expon, (Q as u64), rndMode);
            return res;
        }
    }
    if (diff_expon >= 0) {
        rmode = rndMode;
        if (((sign_x ^ sign_y) != 0) && (((rmode.wrapping_sub(1)) as u64) < 2)) {
            rmode = ((3 as i64).wrapping_sub(rmode));
        }
        match rmode {
            0 | 4 => {
                R = R.wrapping_add(R);
                R = (((go_checked_shl_u32(R, go_shift_count_u64((2) as u64)))).wrapping_add(R));
                B5 = (B4.wrapping_add(B));
                R = (B5.wrapping_sub(R));
                R = R.wrapping_sub(((((Q | ((go_checked_shr_i64(rmode, go_shift_count_u64((2) as u64))) as u32))) & 1)));
                D = (go_checked_shr_u32(R, go_shift_count_u64((31) as u64)));
                Q = Q.wrapping_add(D);
            }
            1 | 3 => {
            }
            _ => {
                Q = Q.wrapping_add(1);
            }
        }
        res = get_bid32((sign_x ^ sign_y), diff_expon, (Q as u64), rndMode);
        return res;
    } else {
        rmode = rndMode;
        let mut uf_pfpsf: u32 = 0;
        res = get_bid32_uf((sign_x ^ sign_y), diff_expon, (Q as u64), R, rmode, (&mut uf_pfpsf));
        return res;
    }
}
