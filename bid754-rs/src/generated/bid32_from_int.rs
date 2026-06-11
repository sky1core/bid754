// Auto-generated from bid32_from_int.go by go2rs. Do not edit.

use super::prelude::*;

pub fn bid32_from_int32(mut x: i32, mut rnd_mode: i64) -> (u32, u32) {
    let mut res: u32 = 0;
    let mut res64: u64 = 0;
    let mut x_sign: u32 = 0;
    let mut C: u32 = 0;
    let mut q: u32 = 0;
    let mut ind: u32 = 0;
    let mut incr_exp: i64 = 0;
    let mut is_midpoint_lt_even: i64 = 0;
    let mut is_midpoint_gt_even: i64 = 0;
    let mut is_inexact_lt_midpoint: i64 = 0;
    let mut is_inexact_gt_midpoint: i64 = 0;
    let mut pfpsf: u32 = 0;
    x_sign = ((x as u32) & 0x80000000);
    if (x_sign != 0) {
        C = ((!(x as u32)).wrapping_add(1));
    } else {
        C = (x as u32);
    }
    if (C <= 0x98967f) {
        if (C < 0x00800000) {
            res = ((x_sign | 0x32800000) | C);
        } else {
            res = ((x_sign | 0x6ca00000) | (C & 0x001fffff));
        }
    } else {
        if (C < 0x05f5e100) {
            q = 8;
            ind = 1;
        } else if (C < 0x3b9aca00) {
            q = 9;
            ind = 2;
        } else {
            q = 10;
            ind = 3;
        }
        res64 = bid_round64_2_18((q as i64), (ind as i64), (C as u64), (&mut incr_exp), (&mut is_midpoint_lt_even), (&mut is_midpoint_gt_even), (&mut is_inexact_lt_midpoint), (&mut is_inexact_gt_midpoint));
        res = (res64 as u32);
        if (incr_exp != 0) {
            ind = ind.wrapping_add(1);
        }
        if ((((is_inexact_lt_midpoint != 0) || (is_inexact_gt_midpoint != 0)) || (is_midpoint_lt_even != 0)) || (is_midpoint_gt_even != 0)) {
            pfpsf |= 32;
        }
        if (rnd_mode != 0) {
            if ((((x_sign == 0) && (((((rnd_mode == 2) && (is_inexact_lt_midpoint != 0))) || (((((rnd_mode == 4) || (rnd_mode == 2))) && (is_midpoint_gt_even != 0))))))) || (((x_sign != 0) && (((((rnd_mode == 1) && (is_inexact_lt_midpoint != 0))) || (((((rnd_mode == 4) || (rnd_mode == 1))) && (is_midpoint_gt_even != 0)))))))) {
                res = (res.wrapping_add(1));
                if (res == 10000000) {
                    res = 1000000;
                    ind = (ind.wrapping_add(1));
                }
            } else if ((((is_midpoint_lt_even != 0) || (is_inexact_gt_midpoint != 0))) && (((((x_sign != 0) && (((rnd_mode == 2) || (rnd_mode == 3))))) || (((x_sign == 0) && (((rnd_mode == 1) || (rnd_mode == 3)))))))) {
                res = (res.wrapping_sub(1));
                if (res == 999999) {
                    res = 9999999;
                    ind = (ind.wrapping_sub(1));
                }
            }
        }
        if (res < 0x00800000) {
            res = ((x_sign | ((go_checked_shl_u32(((ind.wrapping_add(101))), go_shift_count_u64((23) as u64))))) | res);
        } else {
            res = (((x_sign | 0x60000000) | ((go_checked_shl_u32(((ind.wrapping_add(101))), go_shift_count_u64((21) as u64))))) | (res & 0x001fffff));
        }
    }
    return (res, pfpsf);
}

pub fn bid32_from_uint32(mut x: u32, mut rnd_mode: i64) -> (u32, u32) {
    let mut res: u32 = 0;
    let mut res64: u64 = 0;
    let mut C: u32 = 0;
    let mut q: u32 = 0;
    let mut ind: u32 = 0;
    let mut incr_exp: i64 = 0;
    let mut is_midpoint_lt_even: i64 = 0;
    let mut is_midpoint_gt_even: i64 = 0;
    let mut is_inexact_lt_midpoint: i64 = 0;
    let mut is_inexact_gt_midpoint: i64 = 0;
    let mut pfpsf: u32 = 0;
    C = x;
    if (C <= 0x98967f) {
        if (C < 0x00800000) {
            res = (0x32800000 | C);
        } else {
            res = (0x6ca00000 | (C & 0x001fffff));
        }
    } else {
        if (C < 0x05f5e100) {
            q = 8;
            ind = 1;
        } else if (C < 0x3b9aca00) {
            q = 9;
            ind = 2;
        } else {
            q = 10;
            ind = 3;
        }
        res64 = bid_round64_2_18((q as i64), (ind as i64), (C as u64), (&mut incr_exp), (&mut is_midpoint_lt_even), (&mut is_midpoint_gt_even), (&mut is_inexact_lt_midpoint), (&mut is_inexact_gt_midpoint));
        res = (res64 as u32);
        if (incr_exp != 0) {
            ind = ind.wrapping_add(1);
        }
        if ((((is_inexact_lt_midpoint != 0) || (is_inexact_gt_midpoint != 0)) || (is_midpoint_lt_even != 0)) || (is_midpoint_gt_even != 0)) {
            pfpsf |= 32;
        }
        if (rnd_mode != 0) {
            if ((((rnd_mode == 2) && (is_inexact_lt_midpoint != 0))) || (((((rnd_mode == 4) || (rnd_mode == 2))) && (is_midpoint_gt_even != 0)))) {
                res = (res.wrapping_add(1));
                if (res == 10000000) {
                    res = 1000000;
                    ind = (ind.wrapping_add(1));
                }
            } else if ((((is_midpoint_lt_even != 0) || (is_inexact_gt_midpoint != 0))) && (((rnd_mode == 1) || (rnd_mode == 3)))) {
                res = (res.wrapping_sub(1));
                if (res == 999999) {
                    res = 9999999;
                    ind = (ind.wrapping_sub(1));
                }
            }
        }
        if (res < 0x00800000) {
            res = (((go_checked_shl_u32(((ind.wrapping_add(101))), go_shift_count_u64((23) as u64)))) | res);
        } else {
            res = ((0x60000000 | ((go_checked_shl_u32(((ind.wrapping_add(101))), go_shift_count_u64((21) as u64))))) | (res & 0x001fffff));
        }
    }
    return (res, pfpsf);
}
