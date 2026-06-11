// Auto-generated from bid128_frexp.go by go2rs. Do not edit.

use super::prelude::*;

pub fn bid128_frexp(mut x: BID_UINT128) -> (BID_UINT128, i64) {
    let mut res: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut sig_x: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut exp_x: u32 = 0;
    let mut x_nr_bits: i64 = 0;
    let mut q: i64 = 0;
    if ((x.w[1] & 0x7800000000000000) >= 0x7800000000000000) {
        let mut exp: i64 = 0;
        res = x;
        if ((x.w[1] & 0x7e00000000000000) == 0x7e00000000000000) {
            res.w[1] = (x.w[1] & 0xfdffffffffffffff);
        }
        return (res, exp);
    } else {
        if ((x.w[1] & 0x6000000000000000) == 0x6000000000000000) {
            let mut exp: i64 = 0;
            exp_x = ((go_checked_shr_u64((x.w[1] & 0x1fff800000000000), go_shift_count_u64((47) as u64))) as u32);
            res.w[1] = ((x.w[1] & 0x8000000000000000) | ((go_checked_shl_u64((exp_x as u64), go_shift_count_u64((49) as u64)))));
            res.w[0] = 0x0000000000000000;
            return (res, exp);
        }
        exp_x = ((go_checked_shr_u64((x.w[1] & 0x7ffe000000000000), go_shift_count_u64((49) as u64))) as u32);
        sig_x.w[1] = (x.w[1] & 0x1ffffffffffff);
        sig_x.w[0] = x.w[0];
        if (((sig_x.w[1] > 0x0001ed09bead87c0) || (((sig_x.w[1] == 0x0001ed09bead87c0) && (sig_x.w[0] > 0x378d8e63ffffffff)))) || (((sig_x.w[1] == 0x0) && (sig_x.w[0] == 0x0)))) {
            let mut exp: i64 = 0;
            res.w[1] = ((x.w[1] & 0x8000000000000000) | ((go_checked_shl_u64((exp_x as u64), go_shift_count_u64((49) as u64)))));
            res.w[0] = 0x0000000000000000;
            return (res, exp);
        } else {
        }
        if (sig_x.w[1] == 0) {
            if (sig_x.w[0] >= 0x0020000000000000) {
                if (sig_x.w[0] >= 0x0000000100000000) {
                    let mut tmp_ui64 = (((go_checked_shr_u64(sig_x.w[0], go_shift_count_u64((32) as u64))) as f64)).to_bits();
                    x_nr_bits = ((32 as i64).wrapping_add((((((((go_checked_shr_u64(tmp_ui64, go_shift_count_u64((52) as u64))) as u32)) & 0x7ff) as i64).wrapping_sub(0x3ff)))));
                } else {
                    let mut tmp_ui64 = (sig_x.w[0] as f64).to_bits();
                    x_nr_bits = (((((((go_checked_shr_u64(tmp_ui64, go_shift_count_u64((52) as u64))) as u32)) & 0x7ff) as i64).wrapping_sub(0x3ff)));
                }
            } else {
                let mut tmp_ui64 = (sig_x.w[0] as f64).to_bits();
                x_nr_bits = (((((((go_checked_shr_u64(tmp_ui64, go_shift_count_u64((52) as u64))) as u32)) & 0x7ff) as i64).wrapping_sub(0x3ff)));
            }
        } else {
            let mut tmp_ui64 = (sig_x.w[1] as f64).to_bits();
            x_nr_bits = ((64 as i64).wrapping_add((((((((go_checked_shr_u64(tmp_ui64, go_shift_count_u64((52) as u64))) as u32)) & 0x7ff) as i64).wrapping_sub(0x3ff)))));
        }
        q = (bid_nr_digits[x_nr_bits as usize].digits as i64);
        if (q == 0) {
            q = (bid_nr_digits[x_nr_bits as usize].digits1 as i64);
            if ((sig_x.w[1] > bid_nr_digits[x_nr_bits as usize].threshold_hi) || (((sig_x.w[1] == bid_nr_digits[x_nr_bits as usize].threshold_hi) && (sig_x.w[0] >= bid_nr_digits[x_nr_bits as usize].threshold_lo)))) {
                q = q.wrapping_add(1);
            }
        }
        let mut exp = (((exp_x as i64).wrapping_sub(6176)).wrapping_add(q));
        res.w[1] = ((x.w[1] & 0x8001ffffffffffff) | ((go_checked_shl_u64((((q.wrapping_neg()).wrapping_add(6176)) as u64), go_shift_count_u64((49) as u64)))));
        res.w[0] = x.w[0];
        return (res, exp);
    }
}
