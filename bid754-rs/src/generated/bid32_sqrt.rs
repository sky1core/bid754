// Auto-generated from bid32_sqrt.go by go2rs. Do not edit.

use super::prelude::*;

pub fn bid32_sqrt(mut x: u32, mut rnd_mode: i64) -> (u32, u32) {
    let mut CA: u64 = 0;
    let mut CT: u64 = 0;
    let mut sign_x: u32 = 0;
    let mut coefficient_x: u32 = 0;
    let mut Q: u32 = 0;
    let mut A10: u32 = 0;
    let mut QE: u32 = 0;
    let mut res: u32 = 0;
    let mut dq: f64 = 0.0;
    let mut dqe: f64 = 0.0;
    let mut exponent_x: i64 = 0;
    let mut exponent_q: i64 = 0;
    let mut bin_expon_cx: i64 = 0;
    let mut digits_x: i64 = 0;
    let mut scale: i64 = 0;
    let mut pfpsf: u32 = 0;
    let (mut sign_x, mut exponent_x, mut coefficient_x, mut valid) = unpack_bid32(x);
    if (coefficient_x == 0) {
        valid = false;
    }
    if (!valid) {
        if ((x & 0x78000000) == 0x78000000) {
            res = coefficient_x;
            if ((coefficient_x & 0xfc000000) == 0xf8000000) {
                res = 0x7c000000;
                pfpsf |= 1;
            }
            if ((x & 0x7e000000) == 0x7e000000) {
                pfpsf |= 1;
            }
            return ((res & 0xfdffffff), pfpsf);
        }
        exponent_x = (go_checked_shr_i64(((exponent_x.wrapping_add(101))), go_shift_count_u64((1) as u64)));
        res = (sign_x | ((go_checked_shl_u32((exponent_x as u32), go_shift_count_u64((23) as u64)))));
        return (res, pfpsf);
    }
    if ((sign_x != 0) && (coefficient_x != 0)) {
        res = 0x7c000000;
        pfpsf |= 1;
        return (res, pfpsf);
    }
    let mut tempx = ((coefficient_x as f32) as f32).to_bits();
    bin_expon_cx = (((((go_checked_shr_u32(tempx, go_shift_count_u64((23) as u64)))) & 0xff) as i64).wrapping_sub(0x7f));
    digits_x = (bid_estimate_decimal_digits[bin_expon_cx as usize] as i64);
    if ((coefficient_x as u64) >= (bid_power10_index_binexp[bin_expon_cx as usize] as u64)) {
        digits_x = digits_x.wrapping_add(1);
    }
    A10 = coefficient_x;
    if ((exponent_x & 1) == 0) {
        A10 = (((go_checked_shl_u32(A10, go_shift_count_u64((2) as u64)))).wrapping_add(A10));
        A10 = A10.wrapping_add(A10);
    }
    dqe = (A10 as f64).sqrt();
    QE = (dqe as u32);
    if ((QE.wrapping_mul(QE)) == A10) {
        res = very_fast_get_bid32(0, (go_checked_shr_i64(((exponent_x.wrapping_add(101))), go_shift_count_u64((1) as u64))), QE);
        return (res, pfpsf);
    }
    scale = ((13 as i64).wrapping_sub(digits_x));
    exponent_q = ((exponent_x.wrapping_add(101)).wrapping_sub(scale));
    scale = scale.wrapping_add(exponent_q & 1);
    CT = bid_power10_table_128[scale as usize].w[0];
    CA = ((coefficient_x as u64).wrapping_mul(CT));
    dq = (CA as f64).sqrt();
    exponent_q = (go_checked_shr_i64((exponent_q), go_shift_count_u64((1) as u64)));
    pfpsf |= 32;
    if ((rnd_mode & 3) == 0) {
        Q = ((dq + 0.5) as u32);
    } else {
        Q = (dq as u32);
        if (rnd_mode == 2) {
            Q = Q.wrapping_add(1);
        }
    }
    res = fast_get_bid32(0, exponent_q, Q);
    return (res, pfpsf);
}
