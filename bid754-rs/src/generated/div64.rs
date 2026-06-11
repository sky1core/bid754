// Auto-generated from div64.go by go2rs. Do not edit.

use super::prelude::*;

pub fn bid64_div(mut x: u64, mut y: u64, mut rndMode: i64) -> u64 {
    let mut CA: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut sign_x: u64 = 0;
    let mut sign_y: u64 = 0;
    let mut coefficient_x: u64 = 0;
    let mut coefficient_y: u64 = 0;
    let mut A: u64 = 0;
    let mut B: u64 = 0;
    let mut Q: u64 = 0;
    let mut Q2: u64 = 0;
    let mut R: u64 = 0;
    let mut T: u64 = 0;
    let mut DU: u64 = 0;
    let mut res: u64 = 0;
    let mut B2: u64 = 0;
    let mut B4: u64 = 0;
    let mut B5: u64 = 0;
    let mut valid_x: bool = false;
    let mut valid_y: bool = false;
    let mut exponent_x: i64 = 0;
    let mut exponent_y: i64 = 0;
    let mut bin_expon_cx: i64 = 0;
    let mut diff_expon: i64 = 0;
    let mut ed1: i64 = 0;
    let mut ed2: i64 = 0;
    let mut bin_index: i64 = 0;
    let mut rmode: i64 = 0;
    let mut D: i64 = 0;
    let mut db: f64 = 0.0;
    (sign_x, exponent_x, coefficient_x, valid_x) = unpack_bid64(x);
    (sign_y, exponent_y, coefficient_y, valid_y) = unpack_bid64(y);
    if (!valid_x) {
        if ((x & 0x7c00000000000000) == 0x7c00000000000000) {
            return (coefficient_x & 0xfdffffffffffffff);
        }
        if ((x & 0x7800000000000000) == 0x7800000000000000) {
            if ((y & 0x7800000000000000) == 0x7800000000000000) {
                if ((y & 0x7c00000000000000) == 0x7800000000000000) {
                    return 0x7c00000000000000;
                }
            } else {
                return ((((x ^ y) & 0x8000000000000000)) | 0x7800000000000000);
            }
        }
        if ((((y & 0x7800000000000000) != 0x7800000000000000)) && (coefficient_y == 0)) {
            return 0x7c00000000000000;
        }
        if ((y & 0x7800000000000000) != 0x7800000000000000) {
            if ((y & 0x6000000000000000) == 0x6000000000000000) {
                exponent_y = (((((go_checked_shr_u64(y, go_shift_count_u64((51) as u64))) as u32) & 0x3ff)) as i64);
            } else {
                exponent_y = (((((go_checked_shr_u64(y, go_shift_count_u64((53) as u64))) as u32) & 0x3ff)) as i64);
            }
            sign_y = (y & 0x8000000000000000);
            exponent_x = ((exponent_x.wrapping_sub(exponent_y)).wrapping_add(0x18e));
            if (exponent_x > 0x2ff) {
                exponent_x = 0x2ff;
            } else if (exponent_x < 0) {
                exponent_x = 0;
            }
            return ((sign_x ^ sign_y) | ((go_checked_shl_u64((exponent_x as u64), go_shift_count_u64((53) as u64)))));
        }
    }
    if (!valid_y) {
        if ((y & 0x7c00000000000000) == 0x7c00000000000000) {
            return (coefficient_y & 0xfdffffffffffffff);
        }
        if ((y & 0x7800000000000000) == 0x7800000000000000) {
            return ((x ^ y) & 0x8000000000000000);
        }
        return ((sign_x ^ sign_y) | 0x7800000000000000);
    }
    diff_expon = ((exponent_x.wrapping_sub(exponent_y)).wrapping_add(0x18e));
    if (coefficient_x < coefficient_y) {
        let mut tempx = ((coefficient_x as f32) as f32).to_bits();
        let mut tempy = ((coefficient_y as f32) as f32).to_bits();
        bin_index = ((go_checked_shr_u32(((tempy.wrapping_sub(tempx))), go_shift_count_u64((23) as u64))) as i64);
        A = (coefficient_x.wrapping_mul(bid_power10_index_binexp[bin_index as usize]));
        B = coefficient_y;
        let mut temp_b = (B as f64);
        DU = (go_checked_shr_u64(((A.wrapping_sub(B))), go_shift_count_u64((63) as u64)));
        ed1 = ((15 as i64).wrapping_add(DU as i64));
        ed2 = ((bid_estimate_decimal_digits[bin_index as usize] as i64).wrapping_add(ed1));
        T = bid_power10_table_128[ed1 as usize].w[0];
        CA = __mul_64x64_to_128(A, T);
        Q = 0;
        diff_expon = (diff_expon.wrapping_sub(ed2));
        if (coefficient_y < 0x0020000000000000) {
            let mut temp_b_bits = (temp_b).to_bits();
            temp_b_bits = temp_b_bits.wrapping_add(1);
            db = f64::from_bits(temp_b_bits);
        } else {
            db = (((B.wrapping_add(2)).wrapping_add(B & 1)) as f64);
        }
    } else {
        let mut A2 = (coefficient_x | 1);
        let mut da = (A2 as f64);
        db = (coefficient_y as f64);
        let mut dq = (da / db);
        Q = (dq as u64);
        R = (coefficient_x.wrapping_sub((coefficient_y.wrapping_mul(Q))));
        let mut tempq = (dq).to_bits();
        bin_expon_cx = ((((go_checked_shr_u64(tempq, go_shift_count_u64((52) as u64)))) as i64).wrapping_sub(0x3ff));
        D = (go_checked_shr_i64((R as i64), go_shift_count_u64((63) as u64)));
        Q = Q.wrapping_add(D as u64);
        R = R.wrapping_add((coefficient_y & (D as u64)));
        if ((R as i64) <= 0) {
            res = get_bid64((sign_x ^ sign_y), diff_expon, (Q.wrapping_add(R)), rndMode);
            return res;
        }
        DU = ((bid_power10_index_binexp[bin_expon_cx as usize].wrapping_sub(Q)).wrapping_sub(1));
        DU = go_checked_shr_u64(DU, go_shift_count_u64((63) as u64));
        ed2 = (((16 as i64).wrapping_sub(bid_estimate_decimal_digits[bin_expon_cx as usize] as i64)).wrapping_sub(DU as i64));
        T = bid_power10_table_128[ed2 as usize].w[0];
        CA = __mul_64x64_to_128(R, T);
        B = coefficient_y;
        Q = Q.wrapping_mul(bid_power10_table_128[ed2 as usize].w[0]);
        diff_expon = diff_expon.wrapping_sub(ed2);
    }
    if (CA.w[1] == 0) {
        Q2 = (CA.w[0] / B);
        B2 = (B.wrapping_add(B));
        B4 = (B2.wrapping_add(B2));
        R = (CA.w[0].wrapping_sub((Q2.wrapping_mul(B))));
        Q = Q.wrapping_add(Q2);
    } else {
        let mut t_scale = f64::from_bits(0x43f0000000000000);
        let mut da_h = (CA.w[1] as f64);
        let mut da_l = (CA.w[0] as f64);
        let mut da = no_fma_mul_add_f64(da_h, t_scale, da_l);
        let mut dq = (da / db);
        Q2 = (dq as u64);
        R = (CA.w[0].wrapping_sub((Q2.wrapping_mul(B))));
        D = (go_checked_shr_i64((R as i64), go_shift_count_u64((63) as u64)));
        Q2 = Q2.wrapping_add(D as u64);
        R = R.wrapping_add((B & (D as u64)));
        B2 = (B.wrapping_add(B));
        B4 = (B2.wrapping_add(B2));
        R = (R.wrapping_sub(B4));
        D = (go_checked_shr_i64((R as i64), go_shift_count_u64((63) as u64)));
        R = R.wrapping_add((B4 & (D as u64)));
        Q2 = Q2.wrapping_add((((!(D as u64))) & 4));
        R = (R.wrapping_sub(B2));
        D = (go_checked_shr_i64((R as i64), go_shift_count_u64((63) as u64)));
        R = R.wrapping_add((B2 & (D as u64)));
        Q2 = Q2.wrapping_add((((!(D as u64))) & 2));
        R = (R.wrapping_sub(B));
        D = (go_checked_shr_i64((R as i64), go_shift_count_u64((63) as u64)));
        R = R.wrapping_add((B & (D as u64)));
        Q2 = Q2.wrapping_add((((!(D as u64))) & 1));
        Q = Q.wrapping_add(Q2);
    }
    if (R == 0) {
        let mut nzeros: i64 = 0;
        if ((coefficient_x <= 1024) && (coefficient_y <= 1024)) {
            let mut i = ((coefficient_y as i64).wrapping_sub(1));
            let mut j = ((coefficient_x as i64).wrapping_sub(1));
            nzeros = ((ed2.wrapping_sub(bid_factors[i as usize][0] as i64)).wrapping_add(bid_factors[j as usize][0] as i64));
            let mut d5 = ((ed2.wrapping_sub(bid_factors[i as usize][1] as i64)).wrapping_add(bid_factors[j as usize][1] as i64));
            if (d5 < nzeros) {
                nzeros = d5;
            }
            if (nzeros > 0) {
                let mut CT = __mul_64x64_to_128(Q, bid_reciprocals10_64[nzeros as usize]);
                let mut amount = (bid_short_recip_scale[nzeros as usize] as i64);
                Q = (go_checked_shr_u64(CT.w[1], go_shift_count_u64((amount as u64) as u64)));
            }
            diff_expon = diff_expon.wrapping_add(nzeros);
        } else {
            let mut tdigit: [u32; 3] = [0; 3];
            tdigit[0] = ((Q & 0x3ffffff) as u32);
            tdigit[1] = 0;
            let mut QX = (go_checked_shr_u64(Q, go_shift_count_u64((26) as u64)));
            let mut QX32 = (QX as u32);
            nzeros = 0;
            let mut j: i64 = 0;
            while (QX32 != 0) {
                let mut k = ((QX32 & 127) as i64);
                tdigit[0] = tdigit[0].wrapping_add(bid_convert_table[j as usize][k as usize][0]);
                tdigit[1] = tdigit[1].wrapping_add(bid_convert_table[j as usize][k as usize][1]);
                if (tdigit[0] >= 100000000) {
                    tdigit[0] = tdigit[0].wrapping_sub(100000000);
                    tdigit[1] = tdigit[1].wrapping_add(1);
                }
                j = (j.wrapping_add(1));
                QX32 = (go_checked_shr_u32(QX32, go_shift_count_u64((7) as u64)));
            }
            let mut digit = tdigit[0];
            if ((digit == 0) && (tdigit[1] == 0)) {
                nzeros = nzeros.wrapping_add(16);
            } else {
                if (digit == 0) {
                    nzeros = nzeros.wrapping_add(8);
                    digit = tdigit[1];
                }
                let mut PD = ((digit as u64).wrapping_mul(0x068DB8BB));
                let mut digit_h = ((go_checked_shr_u64(PD, go_shift_count_u64((40) as u64))) as u32);
                let mut digit_low = (digit.wrapping_sub((digit_h.wrapping_mul(10000))));
                if (digit_low == 0) {
                    nzeros = nzeros.wrapping_add(4);
                } else {
                    digit_h = digit_low;
                }
                if ((digit_h & 1) == 0) {
                    nzeros = nzeros.wrapping_add(((3 & ((go_checked_shr_u32((bid_packed_10000_zeros[(go_checked_shr_u32(digit_h, go_shift_count_u64((3) as u64))) as usize] as u32), go_shift_count_u64((digit_h & 7) as u64))))) as i64));
                }
            }
            if (nzeros > 0) {
                let mut CT = __mul_64x64_to_128(Q, bid_reciprocals10_64[nzeros as usize]);
                let mut amount = (bid_short_recip_scale[nzeros as usize] as i64);
                Q = (go_checked_shr_u64(CT.w[1], go_shift_count_u64((amount as u64) as u64)));
            }
            diff_expon = diff_expon.wrapping_add(nzeros);
        }
        if (diff_expon >= 0) {
            res = fast_get_bid64_check_of((sign_x ^ sign_y), diff_expon, Q, rndMode);
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
                R = (((go_checked_shl_u64(R, go_shift_count_u64((2) as u64)))).wrapping_add(R));
                B5 = (B4.wrapping_add(B));
                R = (B5.wrapping_sub(R));
                R = R.wrapping_sub((((Q | ((go_checked_shr_u64((rmode as u64), go_shift_count_u64((2) as u64)))))) & 1));
                Q = Q.wrapping_add((go_checked_shr_u64(R, go_shift_count_u64((63) as u64))));
            }
            1 | 3 => {
            }
            _ => {
                Q = Q.wrapping_add(1);
            }
        }
        res = fast_get_bid64_check_of((sign_x ^ sign_y), diff_expon, Q, rndMode);
        return res;
    } else {
        rmode = rndMode;
        res = get_bid64_uf((sign_x ^ sign_y), diff_expon, Q, R, rmode);
        return res;
    }
}

pub fn bid64_div_with_flags(mut x: u64, mut y: u64, mut rndMode: i64) -> (u64, u32) {
    let mut CA: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut sign_x: u64 = 0;
    let mut sign_y: u64 = 0;
    let mut coefficient_x: u64 = 0;
    let mut coefficient_y: u64 = 0;
    let mut A: u64 = 0;
    let mut B: u64 = 0;
    let mut Q: u64 = 0;
    let mut Q2: u64 = 0;
    let mut R: u64 = 0;
    let mut T: u64 = 0;
    let mut DU: u64 = 0;
    let mut res: u64 = 0;
    let mut B2: u64 = 0;
    let mut B4: u64 = 0;
    let mut B5: u64 = 0;
    let mut valid_x: bool = false;
    let mut valid_y: bool = false;
    let mut exponent_x: i64 = 0;
    let mut exponent_y: i64 = 0;
    let mut bin_expon_cx: i64 = 0;
    let mut diff_expon: i64 = 0;
    let mut ed1: i64 = 0;
    let mut ed2: i64 = 0;
    let mut bin_index: i64 = 0;
    let mut rmode: i64 = 0;
    let mut D: i64 = 0;
    let mut db: f64 = 0.0;
    let mut pfpsf: u32 = 0;
    (sign_x, exponent_x, coefficient_x, valid_x) = unpack_bid64(x);
    (sign_y, exponent_y, coefficient_y, valid_y) = unpack_bid64(y);
    if (!valid_x) {
        if ((y & 0x7e00000000000000) == 0x7e00000000000000) {
            pfpsf |= 1;
        }
        if ((x & 0x7c00000000000000) == 0x7c00000000000000) {
            if ((x & 0x7e00000000000000) == 0x7e00000000000000) {
                pfpsf |= 1;
            }
            return ((coefficient_x & 0xfdffffffffffffff), pfpsf);
        }
        if ((x & 0x7800000000000000) == 0x7800000000000000) {
            if ((y & 0x7800000000000000) == 0x7800000000000000) {
                if ((y & 0x7c00000000000000) == 0x7800000000000000) {
                    pfpsf |= 1;
                    return (0x7c00000000000000, pfpsf);
                }
            } else {
                return (((((x ^ y) & 0x8000000000000000)) | 0x7800000000000000), pfpsf);
            }
        }
        if ((((y & 0x7800000000000000) != 0x7800000000000000)) && (coefficient_y == 0)) {
            pfpsf |= 1;
            return (0x7c00000000000000, pfpsf);
        }
        if ((y & 0x7800000000000000) != 0x7800000000000000) {
            if ((y & 0x6000000000000000) == 0x6000000000000000) {
                exponent_y = (((((go_checked_shr_u64(y, go_shift_count_u64((51) as u64))) as u32) & 0x3ff)) as i64);
            } else {
                exponent_y = (((((go_checked_shr_u64(y, go_shift_count_u64((53) as u64))) as u32) & 0x3ff)) as i64);
            }
            sign_y = (y & 0x8000000000000000);
            exponent_x = ((exponent_x.wrapping_sub(exponent_y)).wrapping_add(0x18e));
            if (exponent_x > 0x2ff) {
                exponent_x = 0x2ff;
            } else if (exponent_x < 0) {
                exponent_x = 0;
            }
            return (((sign_x ^ sign_y) | ((go_checked_shl_u64((exponent_x as u64), go_shift_count_u64((53) as u64))))), pfpsf);
        }
    }
    if (!valid_y) {
        if ((y & 0x7c00000000000000) == 0x7c00000000000000) {
            if ((y & 0x7e00000000000000) == 0x7e00000000000000) {
                pfpsf |= 1;
            }
            return ((coefficient_y & 0xfdffffffffffffff), pfpsf);
        }
        if ((y & 0x7800000000000000) == 0x7800000000000000) {
            return (((x ^ y) & 0x8000000000000000), pfpsf);
        }
        pfpsf |= 4;
        return (((sign_x ^ sign_y) | 0x7800000000000000), pfpsf);
    }
    diff_expon = ((exponent_x.wrapping_sub(exponent_y)).wrapping_add(0x18e));
    if (coefficient_x < coefficient_y) {
        let mut tempx = ((coefficient_x as f32) as f32).to_bits();
        let mut tempy = ((coefficient_y as f32) as f32).to_bits();
        bin_index = ((go_checked_shr_u32(((tempy.wrapping_sub(tempx))), go_shift_count_u64((23) as u64))) as i64);
        A = (coefficient_x.wrapping_mul(bid_power10_index_binexp[bin_index as usize]));
        B = coefficient_y;
        let mut temp_b = (B as f64);
        DU = (go_checked_shr_u64(((A.wrapping_sub(B))), go_shift_count_u64((63) as u64)));
        ed1 = ((15 as i64).wrapping_add(DU as i64));
        ed2 = ((bid_estimate_decimal_digits[bin_index as usize] as i64).wrapping_add(ed1));
        T = bid_power10_table_128[ed1 as usize].w[0];
        CA = __mul_64x64_to_128(A, T);
        Q = 0;
        diff_expon = (diff_expon.wrapping_sub(ed2));
        if (coefficient_y < 0x0020000000000000) {
            let mut temp_b_bits = (temp_b).to_bits();
            temp_b_bits = temp_b_bits.wrapping_add(1);
            db = f64::from_bits(temp_b_bits);
        } else {
            db = (((B.wrapping_add(2)).wrapping_add(B & 1)) as f64);
        }
    } else {
        let mut A2 = (coefficient_x | 1);
        let mut da = (A2 as f64);
        db = (coefficient_y as f64);
        let mut dq = (da / db);
        Q = (dq as u64);
        R = (coefficient_x.wrapping_sub((coefficient_y.wrapping_mul(Q))));
        let mut tempq = (dq).to_bits();
        bin_expon_cx = ((((go_checked_shr_u64(tempq, go_shift_count_u64((52) as u64)))) as i64).wrapping_sub(0x3ff));
        D = (go_checked_shr_i64((R as i64), go_shift_count_u64((63) as u64)));
        Q = Q.wrapping_add(D as u64);
        R = R.wrapping_add((coefficient_y & (D as u64)));
        if ((R as i64) <= 0) {
            let (mut res, mut flags) = get_bid64_flags((sign_x ^ sign_y), diff_expon, (Q.wrapping_add(R)), rndMode);
            pfpsf |= flags;
            return (res, pfpsf);
        }
        DU = ((bid_power10_index_binexp[bin_expon_cx as usize].wrapping_sub(Q)).wrapping_sub(1));
        DU = go_checked_shr_u64(DU, go_shift_count_u64((63) as u64));
        ed2 = (((16 as i64).wrapping_sub(bid_estimate_decimal_digits[bin_expon_cx as usize] as i64)).wrapping_sub(DU as i64));
        T = bid_power10_table_128[ed2 as usize].w[0];
        CA = __mul_64x64_to_128(R, T);
        B = coefficient_y;
        Q = Q.wrapping_mul(bid_power10_table_128[ed2 as usize].w[0]);
        diff_expon = diff_expon.wrapping_sub(ed2);
    }
    if (CA.w[1] == 0) {
        Q2 = (CA.w[0] / B);
        B2 = (B.wrapping_add(B));
        B4 = (B2.wrapping_add(B2));
        R = (CA.w[0].wrapping_sub((Q2.wrapping_mul(B))));
        Q = Q.wrapping_add(Q2);
    } else {
        let mut t_scale = f64::from_bits(0x43f0000000000000);
        let mut da_h = (CA.w[1] as f64);
        let mut da_l = (CA.w[0] as f64);
        let mut da = no_fma_mul_add_f64(da_h, t_scale, da_l);
        let mut dq = (da / db);
        Q2 = (dq as u64);
        R = (CA.w[0].wrapping_sub((Q2.wrapping_mul(B))));
        D = (go_checked_shr_i64((R as i64), go_shift_count_u64((63) as u64)));
        Q2 = Q2.wrapping_add(D as u64);
        R = R.wrapping_add((B & (D as u64)));
        B2 = (B.wrapping_add(B));
        B4 = (B2.wrapping_add(B2));
        R = (R.wrapping_sub(B4));
        D = (go_checked_shr_i64((R as i64), go_shift_count_u64((63) as u64)));
        R = R.wrapping_add((B4 & (D as u64)));
        Q2 = Q2.wrapping_add((((!(D as u64))) & 4));
        R = (R.wrapping_sub(B2));
        D = (go_checked_shr_i64((R as i64), go_shift_count_u64((63) as u64)));
        R = R.wrapping_add((B2 & (D as u64)));
        Q2 = Q2.wrapping_add((((!(D as u64))) & 2));
        R = (R.wrapping_sub(B));
        D = (go_checked_shr_i64((R as i64), go_shift_count_u64((63) as u64)));
        R = R.wrapping_add((B & (D as u64)));
        Q2 = Q2.wrapping_add((((!(D as u64))) & 1));
        Q = Q.wrapping_add(Q2);
    }
    if (R != 0) {
        pfpsf |= 32;
    }
    if (R == 0) {
        let mut nzeros: i64 = 0;
        if ((coefficient_x <= 1024) && (coefficient_y <= 1024)) {
            let mut i = ((coefficient_y as i64).wrapping_sub(1));
            let mut j = ((coefficient_x as i64).wrapping_sub(1));
            nzeros = ((ed2.wrapping_sub(bid_factors[i as usize][0] as i64)).wrapping_add(bid_factors[j as usize][0] as i64));
            let mut d5 = ((ed2.wrapping_sub(bid_factors[i as usize][1] as i64)).wrapping_add(bid_factors[j as usize][1] as i64));
            if (d5 < nzeros) {
                nzeros = d5;
            }
            if (nzeros > 0) {
                let mut CT = __mul_64x64_to_128(Q, bid_reciprocals10_64[nzeros as usize]);
                let mut amount = (bid_short_recip_scale[nzeros as usize] as i64);
                Q = (go_checked_shr_u64(CT.w[1], go_shift_count_u64((amount as u64) as u64)));
            }
            diff_expon = diff_expon.wrapping_add(nzeros);
        } else {
            let mut tdigit: [u32; 3] = [0; 3];
            tdigit[0] = ((Q & 0x3ffffff) as u32);
            tdigit[1] = 0;
            let mut QX = (go_checked_shr_u64(Q, go_shift_count_u64((26) as u64)));
            let mut QX32 = (QX as u32);
            nzeros = 0;
            let mut j: i64 = 0;
            while (QX32 != 0) {
                let mut k = ((QX32 & 127) as i64);
                tdigit[0] = tdigit[0].wrapping_add(bid_convert_table[j as usize][k as usize][0]);
                tdigit[1] = tdigit[1].wrapping_add(bid_convert_table[j as usize][k as usize][1]);
                if (tdigit[0] >= 100000000) {
                    tdigit[0] = tdigit[0].wrapping_sub(100000000);
                    tdigit[1] = tdigit[1].wrapping_add(1);
                }
                j = (j.wrapping_add(1));
                QX32 = (go_checked_shr_u32(QX32, go_shift_count_u64((7) as u64)));
            }
            let mut digit = tdigit[0];
            if ((digit == 0) && (tdigit[1] == 0)) {
                nzeros = nzeros.wrapping_add(16);
            } else {
                if (digit == 0) {
                    nzeros = nzeros.wrapping_add(8);
                    digit = tdigit[1];
                }
                let mut PD = ((digit as u64).wrapping_mul(0x068DB8BB));
                let mut digit_h = ((go_checked_shr_u64(PD, go_shift_count_u64((40) as u64))) as u32);
                let mut digit_low = (digit.wrapping_sub((digit_h.wrapping_mul(10000))));
                if (digit_low == 0) {
                    nzeros = nzeros.wrapping_add(4);
                } else {
                    digit_h = digit_low;
                }
                if ((digit_h & 1) == 0) {
                    nzeros = nzeros.wrapping_add(((3 & ((go_checked_shr_u32((bid_packed_10000_zeros[(go_checked_shr_u32(digit_h, go_shift_count_u64((3) as u64))) as usize] as u32), go_shift_count_u64((digit_h & 7) as u64))))) as i64));
                }
            }
            if (nzeros > 0) {
                let mut CT = __mul_64x64_to_128(Q, bid_reciprocals10_64[nzeros as usize]);
                let mut amount = (bid_short_recip_scale[nzeros as usize] as i64);
                Q = (go_checked_shr_u64(CT.w[1], go_shift_count_u64((amount as u64) as u64)));
            }
            diff_expon = diff_expon.wrapping_add(nzeros);
        }
        if (diff_expon >= 0) {
            let (mut res, mut flags) = fast_get_bid64_check_of_flags((sign_x ^ sign_y), diff_expon, Q, rndMode);
            pfpsf |= flags;
            return (res, pfpsf);
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
                R = (((go_checked_shl_u64(R, go_shift_count_u64((2) as u64)))).wrapping_add(R));
                B5 = (B4.wrapping_add(B));
                R = (B5.wrapping_sub(R));
                R = R.wrapping_sub((((Q | ((go_checked_shr_u64((rmode as u64), go_shift_count_u64((2) as u64)))))) & 1));
                Q = Q.wrapping_add((go_checked_shr_u64(R, go_shift_count_u64((63) as u64))));
            }
            1 | 3 => {
            }
            _ => {
                Q = Q.wrapping_add(1);
            }
        }
        let (mut res, mut flags) = fast_get_bid64_check_of_flags((sign_x ^ sign_y), diff_expon, Q, rndMode);
        pfpsf |= flags;
        return (res, pfpsf);
    } else {
        if ((diff_expon.wrapping_add(16)) < 0) {
            pfpsf |= 32;
        }
        rmode = rndMode;
        res = get_bid64_uf_with_flags((sign_x ^ sign_y), diff_expon, Q, R, rmode, (&mut pfpsf));
        return (res, pfpsf);
    }
}
