// Auto-generated from to_uint64_support.go by go2rs. Do not edit.

use super::prelude::*;

pub(crate) fn __mul_128x64_to_128(mut a: u64, mut b: BID_UINT128) -> BID_UINT128 {
    return __mul_64x128_short(a, b);
}
