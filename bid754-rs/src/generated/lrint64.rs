// Auto-generated from lrint64.go by go2rs. Do not edit.

use super::prelude::*;

pub(crate) fn bid_size_long() -> i64 {
    if (std::env::consts::OS == "windows") {
        return 4;
    }
    return ((usize::BITS as i64) / 8);
}

pub fn bid64_llrint(mut x: u64, mut rndMode: i64) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut pfpsf: u32 = 0;
    if (rndMode == 0) {
        (res, pfpsf) = bid64_to_int64_xrnint(x);
    } else if (rndMode == 4) {
        (res, pfpsf) = bid64_to_int64_xrninta(x);
    } else if (rndMode == 1) {
        (res, pfpsf) = bid64_to_int64_xfloor(x);
    } else if (rndMode == 2) {
        (res, pfpsf) = bid64_to_int64_xceil(x);
    } else {
        (res, pfpsf) = bid64_to_int64_xint(x);
    }
    return (res, pfpsf);
}

pub fn bid64_lrint(mut x: u64, mut rndMode: i64) -> (i64, u32) {
    let mut res32: i32 = 0;
    let mut res64: i64 = 0;
    let mut pfpsf: u32 = 0;
    if (bid_size_long() == 4) {
        if (rndMode == 0) {
            (res32, pfpsf) = bid64_to_int32_xrnint(x);
        } else if (rndMode == 4) {
            (res32, pfpsf) = bid64_to_int32_xrninta(x);
        } else if (rndMode == 1) {
            (res32, pfpsf) = bid64_to_int32_xfloor(x);
        } else if (rndMode == 2) {
            (res32, pfpsf) = bid64_to_int32_xceil(x);
        } else {
            (res32, pfpsf) = bid64_to_int32_xint(x);
        }
        return ((res32 as i64), pfpsf);
    }
    if (rndMode == 0) {
        (res64, pfpsf) = bid64_to_int64_xrnint(x);
    } else if (rndMode == 4) {
        (res64, pfpsf) = bid64_to_int64_xrninta(x);
    } else if (rndMode == 1) {
        (res64, pfpsf) = bid64_to_int64_xfloor(x);
    } else if (rndMode == 2) {
        (res64, pfpsf) = bid64_to_int64_xceil(x);
    } else {
        (res64, pfpsf) = bid64_to_int64_xint(x);
    }
    return ((res64 as i64), pfpsf);
}

pub fn bid64_llround(mut x: u64) -> (i64, u32) {
    let mut res: i64 = 0;
    let mut pfpsf: u32 = 0;
    (res, pfpsf) = bid64_to_int64_rninta(x);
    return ((res as i64), pfpsf);
}

pub fn bid64_lround(mut x: u64) -> (i64, u32) {
    let mut res32: i32 = 0;
    let mut res64: i64 = 0;
    let mut pfpsf: u32 = 0;
    if (bid_size_long() == 4) {
        (res32, pfpsf) = bid64_to_int32_rninta(x);
        return ((res32 as i64), pfpsf);
    }
    (res64, pfpsf) = bid64_to_int64_rninta(x);
    return ((res64 as i64), pfpsf);
}
