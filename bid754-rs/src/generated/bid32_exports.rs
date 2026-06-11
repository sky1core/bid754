// Auto-generated from bid32_exports.go by go2rs. Do not edit.

use super::prelude::*;

pub fn bid32_add(mut x: u32, mut y: u32, mut rndMode: i64) -> u32 {
    return bid32_add_pure(x, y, rndMode);
}

pub fn bid32_sub(mut x: u32, mut y: u32, mut rndMode: i64) -> u32 {
    return bid32_sub_pure(x, y, rndMode);
}

pub fn bid32_mul(mut x: u32, mut y: u32, mut rndMode: i64) -> u32 {
    return bid32_mul_pure(x, y, rndMode);
}

pub fn bid32_div(mut x: u32, mut y: u32, mut rndMode: i64) -> u32 {
    return bid32_div_pure(x, y, rndMode);
}

pub fn bid32_min_num(mut x: u32, mut y: u32) -> u32 {
    return bid32_minnum_pure(x, y);
}

pub fn bid32_max_num(mut x: u32, mut y: u32) -> u32 {
    return bid32_maxnum_pure(x, y);
}

pub fn bid32_min_num_mag(mut x: u32, mut y: u32) -> u32 {
    return bid32_minnum_mag_pure(x, y);
}

pub fn bid32_max_num_mag(mut x: u32, mut y: u32) -> u32 {
    return bid32_maxnum_mag_pure(x, y);
}

pub fn bid32_same_quantum(mut x: u32, mut y: u32) -> bool {
    return bid32_same_quantum_pure(x, y);
}

pub fn bid32_quantum(mut x: u32) -> u32 {
    return bid32_quantum_pure(x);
}

pub fn bid32_is_na_n(mut x: u32) -> bool {
    return (bid32_is_na_n32(x) != 0);
}

pub fn bid32_is_inf(mut x: u32) -> bool {
    return (bid32_is_inf32(x) != 0);
}

pub fn bid32_is_zero(mut x: u32) -> bool {
    return (bid32_is_zero32(x) != 0);
}

pub fn bid32_abs(mut x: u32) -> u32 {
    return (x & 0x7fffffff);
}

pub fn bid32_negate(mut x: u32) -> u32 {
    return (x ^ 0x80000000);
}

pub fn bid32_to_string(mut x: u32) -> String {
    return bid32_to_string_raw(x);
}

pub fn bid32_from_string(s: impl AsRef<str>) -> u32 {
    let mut s = s.as_ref().to_string();
    let (mut r, _) = bid32_from_string_raw(s, 0);
    return r;
}
