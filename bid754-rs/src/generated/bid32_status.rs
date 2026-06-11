// Auto-generated from bid32_status.go by go2rs. Do not edit.

use super::prelude::*;

pub(crate) fn bid32_flags_via_bid64_binary(mut x: u32, mut y: u32, mut rndMode: i64, mut op64: fn(u64, u64, i64) -> (u64, u32)) -> u32 {
    let (mut x64, mut f1) = bid32_to_bid64(x);
    let (mut y64, mut f2) = bid32_to_bid64(y);
    let (mut r64, mut f3) = op64(x64, y64, rndMode);
    let (_, mut f4) = bid64_to_bid32(r64, rndMode);
    return (((f1 | f2) | f3) | f4);
}

pub(crate) fn bid32_flags_via_bid64_binary_nornd(mut x: u32, mut y: u32, mut op64: fn(u64, u64) -> (u64, u32)) -> u32 {
    let (mut x64, mut f1) = bid32_to_bid64(x);
    let (mut y64, mut f2) = bid32_to_bid64(y);
    let (mut r64, mut f3) = op64(x64, y64);
    let (_, mut f4) = bid64_to_bid32(r64, 0);
    return (((f1 | f2) | f3) | f4);
}

pub fn bid32_add_with_flags(mut x: u32, mut y: u32, mut rndMode: i64) -> (u32, u32) {
    return (bid32_add_pure(x, y, rndMode), bid32_flags_via_bid64_binary(x, y, rndMode, bid64_add_with_flags));
}

pub fn bid32_sub_with_flags(mut x: u32, mut y: u32, mut rndMode: i64) -> (u32, u32) {
    return (bid32_sub_pure(x, y, rndMode), bid32_flags_via_bid64_binary(x, y, rndMode, bid64_sub_with_flags));
}

pub fn bid32_mul_with_flags(mut x: u32, mut y: u32, mut rndMode: i64) -> (u32, u32) {
    return (bid32_mul_pure(x, y, rndMode), bid32_flags_via_bid64_binary(x, y, rndMode, bid64_mul_with_flags));
}

pub fn bid32_div_with_flags(mut x: u32, mut y: u32, mut rndMode: i64) -> (u32, u32) {
    return (bid32_div_pure(x, y, rndMode), bid32_flags_via_bid64_binary(x, y, rndMode, bid64_div_with_flags));
}

pub fn bid32_min_num_with_flags(mut x: u32, mut y: u32) -> (u32, u32) {
    return (bid32_minnum_pure(x, y), bid32_flags_via_bid64_binary_nornd(x, y, bid64_min_num));
}

pub fn bid32_max_num_with_flags(mut x: u32, mut y: u32) -> (u32, u32) {
    return (bid32_maxnum_pure(x, y), bid32_flags_via_bid64_binary_nornd(x, y, bid64_max_num));
}

pub fn bid32_min_num_mag_with_flags(mut x: u32, mut y: u32) -> (u32, u32) {
    return (bid32_minnum_mag_pure(x, y), bid32_flags_via_bid64_binary_nornd(x, y, bid64_min_num_mag));
}

pub fn bid32_max_num_mag_with_flags(mut x: u32, mut y: u32) -> (u32, u32) {
    return (bid32_maxnum_mag_pure(x, y), bid32_flags_via_bid64_binary_nornd(x, y, bid64_max_num_mag));
}

pub fn bid32_scalbn_with_flags(mut x: u32, mut n: i64, mut rndMode: i64) -> (u32, u32) {
    let (mut res, mut f0) = bid32_scalbn(x, n, rndMode);
    let (mut x64, mut f1) = bid32_to_bid64(x);
    let (mut r64, mut f2) = bid64_scalbn(x64, n, rndMode);
    let (_, mut f3) = bid64_to_bid32(r64, rndMode);
    return (res, (((f0 | f1) | f2) | f3));
}

pub fn bid32_scalbln_with_flags(mut x: u32, mut n: i64, mut rndMode: i64) -> (u32, u32) {
    let (mut res, mut f0) = bid32_scalbln(x, n, rndMode);
    let (mut x64, mut f1) = bid32_to_bid64(x);
    let (mut r64, mut f2) = bid64_scalbln(x64, n, rndMode);
    let (_, mut f3) = bid64_to_bid32(r64, rndMode);
    return (res, (((f0 | f1) | f2) | f3));
}

pub fn bid32_ldexp_with_flags(mut x: u32, mut n: i64, mut rndMode: i64) -> (u32, u32) {
    let (mut res, mut f0) = bid32_ldexp(x, n, rndMode);
    let (mut x64, mut f1) = bid32_to_bid64(x);
    let (mut r64, mut f2) = bid64_ldexp(x64, n, rndMode);
    let (_, mut f3) = bid64_to_bid32(r64, rndMode);
    return (res, (((f0 | f1) | f2) | f3));
}
