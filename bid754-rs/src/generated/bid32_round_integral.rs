// Auto-generated from bid32_round_integral.go by go2rs. Do not edit.

use super::prelude::*;

pub(crate) fn bid32_to_bid64_local(mut x: u32) -> u64 {
    let (mut sign, mut exp, mut coeff, mut valid) = unpack_bid32(x);
    if (!valid) {
        if ((x & 0x7c000000) == 0x7c000000) {
            let mut payload = ((x & 0x000fffff) as u64);
            if ((x & 0x7e000000) == 0x7e000000) {
                return (((go_checked_shl_u64((sign as u64), go_shift_count_u64((32) as u64))) | 0x7e00000000000000) | payload);
            }
            return (((go_checked_shl_u64((sign as u64), go_shift_count_u64((32) as u64))) | 0x7c00000000000000) | payload);
        }
        if ((x & 0x78000000) == 0x78000000) {
            return ((go_checked_shl_u64((sign as u64), go_shift_count_u64((32) as u64))) | 0x7800000000000000);
        }
        let mut exp64 = (exp.wrapping_add(0x18e - 101));
        return ((go_checked_shl_u64((sign as u64), go_shift_count_u64((32) as u64))) | ((go_checked_shl_u64((exp64 as u64), go_shift_count_u64((53) as u64)))));
    }
    let mut exp64 = (exp.wrapping_add(0x18e - 101));
    return (((go_checked_shl_u64((sign as u64), go_shift_count_u64((32) as u64))) | ((go_checked_shl_u64((exp64 as u64), go_shift_count_u64((53) as u64))))) | (coeff as u64));
}

pub(crate) fn bid64_to_bid32_local(mut x: u64) -> u32 {
    let (mut r, _) = bid64_to_bid32(x, 0);
    return r;
}

pub fn bid32_round_integral_exact(mut x: u32, mut rndMode: i64) -> (u32, u32) {
    let (mut x64, mut flags0) = bid32_to_bid64(x);
    let (mut res64, mut flags1) = bid64_round_integral_exact(x64, rndMode);
    let (mut res, mut flags2) = bid64_to_bid32(res64, 0);
    return (res, ((flags0 | flags1) | flags2));
}

pub fn bid32_round_integral_nearest_even(mut x: u32) -> (u32, u32) {
    let (mut x64, mut flags0) = bid32_to_bid64(x);
    let (mut res64, mut flags1) = bid64_round_integral_nearest_even(x64);
    let (mut res, mut flags2) = bid64_to_bid32(res64, 0);
    return (res, ((flags0 | flags1) | flags2));
}

pub fn bid32_round_integral_negative(mut x: u32) -> (u32, u32) {
    let (mut x64, mut flags0) = bid32_to_bid64(x);
    let (mut res64, mut flags1) = bid64_round_integral_negative(x64);
    let (mut res, mut flags2) = bid64_to_bid32(res64, 0);
    return (res, ((flags0 | flags1) | flags2));
}

pub fn bid32_round_integral_positive(mut x: u32) -> (u32, u32) {
    let (mut x64, mut flags0) = bid32_to_bid64(x);
    let (mut res64, mut flags1) = bid64_round_integral_positive(x64);
    let (mut res, mut flags2) = bid64_to_bid32(res64, 0);
    return (res, ((flags0 | flags1) | flags2));
}

pub fn bid32_round_integral_zero(mut x: u32) -> (u32, u32) {
    let (mut x64, mut flags0) = bid32_to_bid64(x);
    let (mut res64, mut flags1) = bid64_round_integral_zero(x64);
    let (mut res, mut flags2) = bid64_to_bid32(res64, 0);
    return (res, ((flags0 | flags1) | flags2));
}

pub fn bid32_round_integral_nearest_away(mut x: u32) -> (u32, u32) {
    let (mut x64, mut flags0) = bid32_to_bid64(x);
    let (mut res64, mut flags1) = bid64_round_integral_nearest_away(x64);
    let (mut res, mut flags2) = bid64_to_bid32(res64, 0);
    return (res, ((flags0 | flags1) | flags2));
}
