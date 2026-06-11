// Auto-generated from bid32_mul.go by go2rs. Do not edit.

use super::prelude::*;

pub(crate) fn bid32_mul_pure(mut x: u32, mut y: u32, mut rndMode: i64) -> u32 {
    let mut Tmp: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut P: u64 = 0;
    let mut Q: u64 = 0;
    let mut R: u64 = 0;
    let mut sign_x: u32 = 0;
    let mut sign_y: u32 = 0;
    let mut coefficient_x: u32 = 0;
    let mut coefficient_y: u32 = 0;
    let mut res: u32 = 0;
    let mut valid_x: bool = false;
    let mut valid_y: bool = false;
    let mut exponent_x: i64 = 0;
    let mut exponent_y: i64 = 0;
    let mut bin_expon_p: i64 = 0;
    let mut amount: i64 = 0;
    let mut n_digits: i64 = 0;
    let mut extra_digits: i64 = 0;
    let mut rmode: i64 = 0;
    (sign_x, exponent_x, coefficient_x, valid_x) = unpack_bid32_add(x);
    (sign_y, exponent_y, coefficient_y, valid_y) = unpack_bid32_add(y);
    if (!valid_x) {
        if ((x & 0x7c000000) == 0x7c000000) {
            return (coefficient_x & 0xfdffffff);
        }
        if ((x & 0x78000000) == 0x78000000) {
            if ((((y & 0x78000000) != 0x78000000)) && (coefficient_y == 0)) {
                return 0x7c000000;
            }
            if ((y & 0x7c000000) == 0x7c000000) {
                return (coefficient_y & 0xfdffffff);
            }
            return ((((x ^ y) & 0x80000000)) | 0x78000000);
        }
        if ((y & 0x78000000) != 0x78000000) {
            if ((y & 0x60000000) == 0x60000000) {
                exponent_y = ((((go_checked_shr_u32(y, go_shift_count_u64((21) as u64)))) & 0xff) as i64);
            } else {
                exponent_y = ((((go_checked_shr_u32(y, go_shift_count_u64((23) as u64)))) & 0xff) as i64);
            }
            sign_y = (y & 0x80000000);
            exponent_x = exponent_x.wrapping_add((exponent_y.wrapping_sub(101)));
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
            if (coefficient_x == 0) {
                return 0x7c000000;
            }
            return ((((x ^ y) & 0x80000000)) | 0x78000000);
        }
        exponent_x = exponent_x.wrapping_add((exponent_y.wrapping_sub(101)));
        if (exponent_x > 191) {
            exponent_x = 191;
        } else if (exponent_x < 0) {
            exponent_x = 0;
        }
        return ((sign_x ^ sign_y) | ((go_checked_shl_u32((exponent_x as u32), go_shift_count_u64((23) as u64)))));
    }
    P = ((coefficient_x as u64).wrapping_mul(coefficient_y as u64));
    let mut tempx = (P as f64);
    bin_expon_p = (((go_checked_shr_u64((((tempx).to_bits() & 0x7ff0000000000000)), go_shift_count_u64((52) as u64))) as i64).wrapping_sub(0x3ff));
    n_digits = (bid_estimate_decimal_digits[bin_expon_p as usize] as i64);
    if (P >= bid_power10_table_128[n_digits as usize].w[0]) {
        n_digits = n_digits.wrapping_add(1);
    }
    exponent_x = exponent_x.wrapping_add((exponent_y.wrapping_sub(101)));
    if (n_digits <= 7) {
        extra_digits = 0;
    } else {
        extra_digits = (n_digits.wrapping_sub(7));
    }
    exponent_x = exponent_x.wrapping_add(extra_digits);
    if (extra_digits == 0) {
        res = get_bid32((sign_x ^ sign_y), exponent_x, P, rndMode);
        return res;
    }
    rmode = rndMode;
    if (((sign_x ^ sign_y) != 0) && (((rmode.wrapping_sub(1)) as u64) < 2)) {
        rmode = ((3 as i64).wrapping_sub(rmode));
    }
    if (exponent_x < 0) {
        rmode = 3;
    }
    P = P.wrapping_add(bid_round_const_table[rmode as usize][extra_digits as usize]);
    Tmp = __mul_64x64_to_128(P, bid_reciprocals10_64[extra_digits as usize]);
    amount = (bid_short_recip_scale[extra_digits as usize] as i64);
    Q = (go_checked_shr_u64(Tmp.w[1], go_shift_count_u64((amount as u64) as u64)));
    R = (P.wrapping_sub((Q.wrapping_mul(bid_power10_table_128[extra_digits as usize].w[0]))));
    if (rmode == 0) {
        if (R == 0) {
            Q &= 0xfffffffe;
        }
    }
    if (((exponent_x == (-1)) && (Q == 9999999)) && (rndMode != 3)) {
        rmode = rndMode;
        if (((sign_x ^ sign_y) != 0) && (((rmode.wrapping_sub(1)) as u64) < 2)) {
            rmode = ((3 as i64).wrapping_sub(rmode));
        }
        if ((((R != 0) && (rmode == 2))) || ((((rmode & 3) == 0) && ((R.wrapping_add(R)) >= bid_power10_table_128[extra_digits as usize].w[0])))) {
            res = very_fast_get_bid32((sign_x ^ sign_y), 0, 1000000);
            return res;
        }
    }
    let mut uf_pfpsf: u32 = 0;
    res = get_bid32_uf((sign_x ^ sign_y), exponent_x, Q, (R as u32), rndMode, (&mut uf_pfpsf));
    return res;
}
