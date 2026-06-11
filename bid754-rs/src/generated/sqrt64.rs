// Auto-generated from sqrt64.go by go2rs. Do not edit.

use super::prelude::*;

pub fn bid64_sqrt(mut x: u64, mut rndMode: i64) -> (u64, u32) {
    let mut CA: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut sign_x: u64 = 0;
    let mut coefficient_x: u64 = 0;
    let mut Q: u64 = 0;
    let mut Q2: u64 = 0;
    let mut A10: u64 = 0;
    let mut C4: u64 = 0;
    let mut R: u64 = 0;
    let mut R2: u64 = 0;
    let mut QE: u64 = 0;
    let mut res: u64 = 0;
    let mut D: i64 = 0;
    let mut t_scale: u64 = 0;
    let mut da: f64 = 0.0;
    let mut dq: f64 = 0.0;
    let mut da_h: f64 = 0.0;
    let mut da_l: f64 = 0.0;
    let mut dqe: f64 = 0.0;
    let mut exponent_x: i64 = 0;
    let mut exponent_q: i64 = 0;
    let mut bin_expon_cx: i64 = 0;
    let mut digits_x: i64 = 0;
    let mut scale: i64 = 0;
    let mut pfpsf: u32 = 0;
    let mut valid: bool = false;
    (sign_x, exponent_x, coefficient_x, valid) = unpack_bid64(x);
    if (!valid) {
        if ((x & 0x7800000000000000) == 0x7800000000000000) {
            res = coefficient_x;
            if ((coefficient_x & 0xfc00000000000000) == 0xf800000000000000) {
                res = 0x7c00000000000000;
                pfpsf |= 1;
            }
            if ((x & 0x7e00000000000000) == 0x7e00000000000000) {
                pfpsf |= 1;
            }
            return ((res & 0xfdffffffffffffff), pfpsf);
        }
        exponent_x = (go_checked_shr_i64(((exponent_x.wrapping_add(0x18e))), go_shift_count_u64((1) as u64)));
        res = (sign_x | ((go_checked_shl_u64((exponent_x as u64), go_shift_count_u64((53) as u64)))));
        return (res, pfpsf);
    }
    if ((sign_x != 0) && (coefficient_x != 0)) {
        res = 0x7c00000000000000;
        pfpsf |= 1;
        return (res, pfpsf);
    }
    t_scale = 0x43f0000000000000;
    bin_expon_cx = (((((go_checked_shr_u32(((coefficient_x as f32) as f32).to_bits(), go_shift_count_u64((23) as u64)))) & 0xff) as i64).wrapping_sub(0x7f));
    digits_x = (bid_estimate_decimal_digits[bin_expon_cx as usize] as i64);
    if (coefficient_x >= bid_power10_index_binexp[bin_expon_cx as usize]) {
        digits_x = digits_x.wrapping_add(1);
    }
    A10 = coefficient_x;
    if ((exponent_x & 1) != 0) {
        A10 = (((go_checked_shl_u64(A10, go_shift_count_u64((2) as u64)))).wrapping_add(A10));
        A10 = A10.wrapping_add(A10);
    }
    dqe = (A10 as f64).sqrt();
    QE = (dqe as u64);
    if ((QE.wrapping_mul(QE)) == A10) {
        res = very_fast_get_bid64(0, (go_checked_shr_i64(((exponent_x.wrapping_add(0x18e))), go_shift_count_u64((1) as u64))), QE);
        return (res, pfpsf);
    }
    scale = ((31 as i64).wrapping_sub(digits_x));
    exponent_q = (exponent_x.wrapping_sub(scale));
    scale = scale.wrapping_add(exponent_q & 1);
    let mut CT = bid_power10_table_128[scale as usize];
    CA = __mul_64x128_short(coefficient_x, CT);
    da_h = (CA.w[1] as f64);
    da_l = (CA.w[0] as f64);
    da = no_fma_mul_add_f64(da_h, f64::from_bits(t_scale), da_l);
    dq = (da).sqrt();
    Q = (dq as u64);
    R = ((go_checked_shr_i64(((CA.w[0].wrapping_sub((Q.wrapping_mul(Q)))) as i64), go_shift_count_u64((63) as u64))) as u64);
    D = (((R.wrapping_add(R)).wrapping_add(1)) as i64);
    exponent_q = (go_checked_shr_i64(((exponent_q.wrapping_add(0x18e))), go_shift_count_u64((1) as u64)));
    pfpsf |= 32;
    if ((rndMode & 3) == 0) {
        Q2 = ((Q.wrapping_add(Q)).wrapping_add(D as u64));
        C4 = (go_checked_shl_u64(CA.w[0], go_shift_count_u64((2) as u64)));
        R2 = ((go_checked_shr_i64((((Q2.wrapping_mul(Q2)).wrapping_sub(C4)) as i64), go_shift_count_u64((63) as u64))) as u64);
        Q = Q.wrapping_add(((D as u64) & (R ^ R2)));
    } else {
        C4 = CA.w[0];
        Q = Q.wrapping_add(D as u64);
        if ((((Q.wrapping_mul(Q)).wrapping_sub(C4)) as i64) > 0) {
            Q = Q.wrapping_sub(1);
        }
        if (rndMode == 2) {
            Q = Q.wrapping_add(1);
        }
    }
    res = fast_get_bid64(0, exponent_q, Q);
    return (res, pfpsf);
}
