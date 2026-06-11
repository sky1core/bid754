// Auto-generated from bid128_from_int.go by go2rs. Do not edit.

use super::prelude::*;

pub fn bid128_from_int32(mut x: i32) -> BID_UINT128 {
    let mut res: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    if ((((x as u32) & 0x80000000)) == 0x80000000) {
        res.w[1] = 0xb040000000000000;
        res.w[0] = (((!(x as u32)).wrapping_add(1)) as u64);
    } else {
        res.w[1] = 0x3040000000000000;
        res.w[0] = ((x as u32) as u64);
    }
    return res;
}

pub fn bid128_from_uint32(mut x: u32) -> BID_UINT128 {
    let mut res: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    res.w[1] = 0x3040000000000000;
    res.w[0] = (x as u64);
    return res;
}

pub fn bid128_from_int64(mut x: i64) -> BID_UINT128 {
    let mut res: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    if ((((x as u64) & 0x8000000000000000)) == 0x8000000000000000) {
        res.w[1] = 0xb040000000000000;
        res.w[0] = ((!(x as u64)).wrapping_add(1));
    } else {
        res.w[1] = 0x3040000000000000;
        res.w[0] = (x as u64);
    }
    return res;
}

pub fn bid128_from_uint64(mut x: u64) -> BID_UINT128 {
    let mut res: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    res.w[1] = 0x3040000000000000;
    res.w[0] = x;
    return res;
}
