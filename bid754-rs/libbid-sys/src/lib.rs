#![allow(non_camel_case_types, non_snake_case)]

use std::ffi::c_char;
use std::os::raw::c_int;

pub type BID_UINT32 = u32;
pub type BID_UINT64 = u64;
pub type BID_SINT64 = i64;
pub type _IDEC_round = u32;
pub type _IDEC_flags = u32;
pub type class_t = c_int;

pub const BID_ROUNDING_TO_NEAREST: _IDEC_round = 0;
pub const BID_ROUNDING_DOWN: _IDEC_round = 1;
pub const BID_ROUNDING_UP: _IDEC_round = 2;
pub const BID_ROUNDING_TO_ZERO: _IDEC_round = 3;
pub const BID_ROUNDING_TIES_AWAY: _IDEC_round = 4;

#[derive(Clone, Copy, Debug, Default, PartialEq, Eq)]
#[repr(C, align(16))]
pub struct BID_UINT128 {
    pub w: [u64; 2],
}

impl BID_UINT128 {
    pub const fn new(lo: u64, hi: u64) -> Self {
        Self { w: [lo, hi] }
    }
}

unsafe extern "C" {
    // Arithmetic: BID32
    #[link_name = "__bid32_add"]
    pub fn bid32_add(x: BID_UINT32, y: BID_UINT32, rnd: _IDEC_round, pfpsf: *mut _IDEC_flags) -> BID_UINT32;
    #[link_name = "__bid32_sub"]
    pub fn bid32_sub(x: BID_UINT32, y: BID_UINT32, rnd: _IDEC_round, pfpsf: *mut _IDEC_flags) -> BID_UINT32;
    #[link_name = "__bid32_mul"]
    pub fn bid32_mul(x: BID_UINT32, y: BID_UINT32, rnd: _IDEC_round, pfpsf: *mut _IDEC_flags) -> BID_UINT32;
    #[link_name = "__bid32_div"]
    pub fn bid32_div(x: BID_UINT32, y: BID_UINT32, rnd: _IDEC_round, pfpsf: *mut _IDEC_flags) -> BID_UINT32;
    #[link_name = "__bid32_fma"]
    pub fn bid32_fma(x: BID_UINT32, y: BID_UINT32, z: BID_UINT32, rnd: _IDEC_round, pfpsf: *mut _IDEC_flags) -> BID_UINT32;
    #[link_name = "__bid32_sqrt"]
    pub fn bid32_sqrt(x: BID_UINT32, rnd: _IDEC_round, pfpsf: *mut _IDEC_flags) -> BID_UINT32;

    // Arithmetic: BID64
    #[link_name = "__bid64_add"]
    pub fn bid64_add(x: BID_UINT64, y: BID_UINT64, rnd: _IDEC_round, pfpsf: *mut _IDEC_flags) -> BID_UINT64;
    #[link_name = "__bid64_sub"]
    pub fn bid64_sub(x: BID_UINT64, y: BID_UINT64, rnd: _IDEC_round, pfpsf: *mut _IDEC_flags) -> BID_UINT64;
    #[link_name = "__bid64_mul"]
    pub fn bid64_mul(x: BID_UINT64, y: BID_UINT64, rnd: _IDEC_round, pfpsf: *mut _IDEC_flags) -> BID_UINT64;
    #[link_name = "__bid64_div"]
    pub fn bid64_div(x: BID_UINT64, y: BID_UINT64, rnd: _IDEC_round, pfpsf: *mut _IDEC_flags) -> BID_UINT64;
    #[link_name = "__bid64_fma"]
    pub fn bid64_fma(
        x: BID_UINT64,
        y: BID_UINT64,
        z: BID_UINT64,
        rnd: _IDEC_round,
        pfpsf: *mut _IDEC_flags,
    ) -> BID_UINT64;
    #[link_name = "__bid64_sqrt"]
    pub fn bid64_sqrt(x: BID_UINT64, rnd: _IDEC_round, pfpsf: *mut _IDEC_flags) -> BID_UINT64;

    // Arithmetic: BID128
    #[link_name = "__bid128_add"]
    pub fn bid128_add(x: BID_UINT128, y: BID_UINT128, rnd: _IDEC_round, pfpsf: *mut _IDEC_flags) -> BID_UINT128;
    #[link_name = "__bid128_sub"]
    pub fn bid128_sub(x: BID_UINT128, y: BID_UINT128, rnd: _IDEC_round, pfpsf: *mut _IDEC_flags) -> BID_UINT128;
    #[link_name = "__bid128_mul"]
    pub fn bid128_mul(x: BID_UINT128, y: BID_UINT128, rnd: _IDEC_round, pfpsf: *mut _IDEC_flags) -> BID_UINT128;
    #[link_name = "__bid128_div"]
    pub fn bid128_div(x: BID_UINT128, y: BID_UINT128, rnd: _IDEC_round, pfpsf: *mut _IDEC_flags) -> BID_UINT128;
    #[link_name = "__bid128_fma"]
    pub fn bid128_fma(
        x: BID_UINT128,
        y: BID_UINT128,
        z: BID_UINT128,
        rnd: _IDEC_round,
        pfpsf: *mut _IDEC_flags,
    ) -> BID_UINT128;
    #[link_name = "__bid128_sqrt"]
    pub fn bid128_sqrt(x: BID_UINT128, rnd: _IDEC_round, pfpsf: *mut _IDEC_flags) -> BID_UINT128;

    // Comparisons: BID32
    #[link_name = "__bid32_quiet_equal"]
    pub fn bid32_quiet_equal(x: BID_UINT32, y: BID_UINT32, pfpsf: *mut _IDEC_flags) -> c_int;
    #[link_name = "__bid32_quiet_greater"]
    pub fn bid32_quiet_greater(x: BID_UINT32, y: BID_UINT32, pfpsf: *mut _IDEC_flags) -> c_int;
    #[link_name = "__bid32_quiet_greater_equal"]
    pub fn bid32_quiet_greater_equal(x: BID_UINT32, y: BID_UINT32, pfpsf: *mut _IDEC_flags) -> c_int;
    #[link_name = "__bid32_quiet_less"]
    pub fn bid32_quiet_less(x: BID_UINT32, y: BID_UINT32, pfpsf: *mut _IDEC_flags) -> c_int;
    #[link_name = "__bid32_quiet_less_equal"]
    pub fn bid32_quiet_less_equal(x: BID_UINT32, y: BID_UINT32, pfpsf: *mut _IDEC_flags) -> c_int;
    #[link_name = "__bid32_quiet_not_equal"]
    pub fn bid32_quiet_not_equal(x: BID_UINT32, y: BID_UINT32, pfpsf: *mut _IDEC_flags) -> c_int;

    // Comparisons: BID64
    #[link_name = "__bid64_quiet_equal"]
    pub fn bid64_quiet_equal(x: BID_UINT64, y: BID_UINT64, pfpsf: *mut _IDEC_flags) -> c_int;
    #[link_name = "__bid64_quiet_greater"]
    pub fn bid64_quiet_greater(x: BID_UINT64, y: BID_UINT64, pfpsf: *mut _IDEC_flags) -> c_int;
    #[link_name = "__bid64_quiet_greater_equal"]
    pub fn bid64_quiet_greater_equal(x: BID_UINT64, y: BID_UINT64, pfpsf: *mut _IDEC_flags) -> c_int;
    #[link_name = "__bid64_quiet_less"]
    pub fn bid64_quiet_less(x: BID_UINT64, y: BID_UINT64, pfpsf: *mut _IDEC_flags) -> c_int;
    #[link_name = "__bid64_quiet_less_equal"]
    pub fn bid64_quiet_less_equal(x: BID_UINT64, y: BID_UINT64, pfpsf: *mut _IDEC_flags) -> c_int;
    #[link_name = "__bid64_quiet_not_equal"]
    pub fn bid64_quiet_not_equal(x: BID_UINT64, y: BID_UINT64, pfpsf: *mut _IDEC_flags) -> c_int;

    // Comparisons: BID128
    #[link_name = "__bid128_quiet_equal"]
    pub fn bid128_quiet_equal(x: BID_UINT128, y: BID_UINT128, pfpsf: *mut _IDEC_flags) -> c_int;
    #[link_name = "__bid128_quiet_greater"]
    pub fn bid128_quiet_greater(x: BID_UINT128, y: BID_UINT128, pfpsf: *mut _IDEC_flags) -> c_int;
    #[link_name = "__bid128_quiet_greater_equal"]
    pub fn bid128_quiet_greater_equal(x: BID_UINT128, y: BID_UINT128, pfpsf: *mut _IDEC_flags) -> c_int;
    #[link_name = "__bid128_quiet_less"]
    pub fn bid128_quiet_less(x: BID_UINT128, y: BID_UINT128, pfpsf: *mut _IDEC_flags) -> c_int;
    #[link_name = "__bid128_quiet_less_equal"]
    pub fn bid128_quiet_less_equal(x: BID_UINT128, y: BID_UINT128, pfpsf: *mut _IDEC_flags) -> c_int;
    #[link_name = "__bid128_quiet_not_equal"]
    pub fn bid128_quiet_not_equal(x: BID_UINT128, y: BID_UINT128, pfpsf: *mut _IDEC_flags) -> c_int;

    // Min/max
    #[link_name = "__bid32_minnum"]
    pub fn bid32_minnum(x: BID_UINT32, y: BID_UINT32, pfpsf: *mut _IDEC_flags) -> BID_UINT32;
    #[link_name = "__bid32_minnum_mag"]
    pub fn bid32_minnum_mag(x: BID_UINT32, y: BID_UINT32, pfpsf: *mut _IDEC_flags) -> BID_UINT32;
    #[link_name = "__bid32_maxnum"]
    pub fn bid32_maxnum(x: BID_UINT32, y: BID_UINT32, pfpsf: *mut _IDEC_flags) -> BID_UINT32;
    #[link_name = "__bid32_maxnum_mag"]
    pub fn bid32_maxnum_mag(x: BID_UINT32, y: BID_UINT32, pfpsf: *mut _IDEC_flags) -> BID_UINT32;
    #[link_name = "__bid64_minnum"]
    pub fn bid64_minnum(x: BID_UINT64, y: BID_UINT64, pfpsf: *mut _IDEC_flags) -> BID_UINT64;
    #[link_name = "__bid64_minnum_mag"]
    pub fn bid64_minnum_mag(x: BID_UINT64, y: BID_UINT64, pfpsf: *mut _IDEC_flags) -> BID_UINT64;
    #[link_name = "__bid64_maxnum"]
    pub fn bid64_maxnum(x: BID_UINT64, y: BID_UINT64, pfpsf: *mut _IDEC_flags) -> BID_UINT64;
    #[link_name = "__bid64_maxnum_mag"]
    pub fn bid64_maxnum_mag(x: BID_UINT64, y: BID_UINT64, pfpsf: *mut _IDEC_flags) -> BID_UINT64;
    #[link_name = "__bid128_minnum"]
    pub fn bid128_minnum(x: BID_UINT128, y: BID_UINT128, pfpsf: *mut _IDEC_flags) -> BID_UINT128;
    #[link_name = "__bid128_minnum_mag"]
    pub fn bid128_minnum_mag(x: BID_UINT128, y: BID_UINT128, pfpsf: *mut _IDEC_flags) -> BID_UINT128;
    #[link_name = "__bid128_maxnum"]
    pub fn bid128_maxnum(x: BID_UINT128, y: BID_UINT128, pfpsf: *mut _IDEC_flags) -> BID_UINT128;
    #[link_name = "__bid128_maxnum_mag"]
    pub fn bid128_maxnum_mag(x: BID_UINT128, y: BID_UINT128, pfpsf: *mut _IDEC_flags) -> BID_UINT128;

    // Conversions from integers
    #[link_name = "__bid32_from_int32"]
    pub fn bid32_from_int32(x: c_int, rnd: _IDEC_round, pfpsf: *mut _IDEC_flags) -> BID_UINT32;
    #[link_name = "__bid32_from_uint32"]
    pub fn bid32_from_uint32(x: BID_UINT32, rnd: _IDEC_round, pfpsf: *mut _IDEC_flags) -> BID_UINT32;
    #[link_name = "__bid32_from_int64"]
    pub fn bid32_from_int64(x: BID_SINT64, rnd: _IDEC_round, pfpsf: *mut _IDEC_flags) -> BID_UINT32;
    #[link_name = "__bid32_from_uint64"]
    pub fn bid32_from_uint64(x: BID_UINT64, rnd: _IDEC_round, pfpsf: *mut _IDEC_flags) -> BID_UINT32;
    #[link_name = "__bid64_from_int32"]
    pub fn bid64_from_int32(x: c_int) -> BID_UINT64;
    #[link_name = "__bid64_from_uint32"]
    pub fn bid64_from_uint32(x: BID_UINT32) -> BID_UINT64;
    #[link_name = "__bid64_from_int64"]
    pub fn bid64_from_int64(x: BID_SINT64, rnd: _IDEC_round, pfpsf: *mut _IDEC_flags) -> BID_UINT64;
    #[link_name = "__bid64_from_uint64"]
    pub fn bid64_from_uint64(x: BID_UINT64, rnd: _IDEC_round, pfpsf: *mut _IDEC_flags) -> BID_UINT64;
    #[link_name = "__bid128_from_int32"]
    pub fn bid128_from_int32(x: c_int) -> BID_UINT128;
    #[link_name = "__bid128_from_uint32"]
    pub fn bid128_from_uint32(x: BID_UINT32) -> BID_UINT128;
    #[link_name = "__bid128_from_int64"]
    pub fn bid128_from_int64(x: BID_SINT64) -> BID_UINT128;
    #[link_name = "__bid128_from_uint64"]
    pub fn bid128_from_uint64(x: BID_UINT64) -> BID_UINT128;

    // Conversions between decimal widths
    #[link_name = "__bid32_to_bid64"]
    pub fn bid32_to_bid64(x: BID_UINT32, pfpsf: *mut _IDEC_flags) -> BID_UINT64;
    #[link_name = "__bid32_to_bid128"]
    pub fn bid32_to_bid128(x: BID_UINT32, pfpsf: *mut _IDEC_flags) -> BID_UINT128;
    #[link_name = "__bid64_to_bid32"]
    pub fn bid64_to_bid32(x: BID_UINT64, rnd: _IDEC_round, pfpsf: *mut _IDEC_flags) -> BID_UINT32;
    #[link_name = "__bid64_to_bid128"]
    pub fn bid64_to_bid128(x: BID_UINT64, pfpsf: *mut _IDEC_flags) -> BID_UINT128;
    #[link_name = "__bid128_to_bid32"]
    pub fn bid128_to_bid32(x: BID_UINT128, rnd: _IDEC_round, pfpsf: *mut _IDEC_flags) -> BID_UINT32;
    #[link_name = "__bid128_to_bid64"]
    pub fn bid128_to_bid64(x: BID_UINT128, rnd: _IDEC_round, pfpsf: *mut _IDEC_flags) -> BID_UINT64;

    // String conversions
    #[link_name = "__bid32_to_string"]
    pub fn bid32_to_string(ps: *mut c_char, x: BID_UINT32, pfpsf: *mut _IDEC_flags);
    #[link_name = "__bid32_from_string"]
    pub fn bid32_from_string(ps: *const c_char, rnd: _IDEC_round, pfpsf: *mut _IDEC_flags) -> BID_UINT32;
    #[link_name = "__bid64_to_string"]
    pub fn bid64_to_string(ps: *mut c_char, x: BID_UINT64, pfpsf: *mut _IDEC_flags);
    #[link_name = "__bid64_from_string"]
    pub fn bid64_from_string(ps: *const c_char, rnd: _IDEC_round, pfpsf: *mut _IDEC_flags) -> BID_UINT64;
    #[link_name = "__bid128_to_string"]
    pub fn bid128_to_string(ps: *mut c_char, x: BID_UINT128, pfpsf: *mut _IDEC_flags);
    #[link_name = "__bid128_from_string"]
    pub fn bid128_from_string(ps: *const c_char, rnd: _IDEC_round, pfpsf: *mut _IDEC_flags) -> BID_UINT128;

    // Quantize
    #[link_name = "__bid32_quantize"]
    pub fn bid32_quantize(x: BID_UINT32, y: BID_UINT32, rnd: _IDEC_round, pfpsf: *mut _IDEC_flags) -> BID_UINT32;
    #[link_name = "__bid64_quantize"]
    pub fn bid64_quantize(x: BID_UINT64, y: BID_UINT64, rnd: _IDEC_round, pfpsf: *mut _IDEC_flags) -> BID_UINT64;
    #[link_name = "__bid128_quantize"]
    pub fn bid128_quantize(x: BID_UINT128, y: BID_UINT128, rnd: _IDEC_round, pfpsf: *mut _IDEC_flags) -> BID_UINT128;

    // Classification: BID32
    #[link_name = "__bid32_isSigned"]
    pub fn bid32_isSigned(x: BID_UINT32) -> c_int;
    #[link_name = "__bid32_isNormal"]
    pub fn bid32_isNormal(x: BID_UINT32) -> c_int;
    #[link_name = "__bid32_isSubnormal"]
    pub fn bid32_isSubnormal(x: BID_UINT32) -> c_int;
    #[link_name = "__bid32_isFinite"]
    pub fn bid32_isFinite(x: BID_UINT32) -> c_int;
    #[link_name = "__bid32_isZero"]
    pub fn bid32_isZero(x: BID_UINT32) -> c_int;
    #[link_name = "__bid32_isInf"]
    pub fn bid32_isInf(x: BID_UINT32) -> c_int;
    #[link_name = "__bid32_isSignaling"]
    pub fn bid32_isSignaling(x: BID_UINT32) -> c_int;
    #[link_name = "__bid32_isCanonical"]
    pub fn bid32_isCanonical(x: BID_UINT32) -> c_int;
    #[link_name = "__bid32_isNaN"]
    pub fn bid32_isNaN(x: BID_UINT32) -> c_int;
    #[link_name = "__bid32_class"]
    pub fn bid32_class(x: BID_UINT32) -> class_t;

    // Classification: BID64
    #[link_name = "__bid64_isSigned"]
    pub fn bid64_isSigned(x: BID_UINT64) -> c_int;
    #[link_name = "__bid64_isNormal"]
    pub fn bid64_isNormal(x: BID_UINT64) -> c_int;
    #[link_name = "__bid64_isSubnormal"]
    pub fn bid64_isSubnormal(x: BID_UINT64) -> c_int;
    #[link_name = "__bid64_isFinite"]
    pub fn bid64_isFinite(x: BID_UINT64) -> c_int;
    #[link_name = "__bid64_isZero"]
    pub fn bid64_isZero(x: BID_UINT64) -> c_int;
    #[link_name = "__bid64_isInf"]
    pub fn bid64_isInf(x: BID_UINT64) -> c_int;
    #[link_name = "__bid64_isSignaling"]
    pub fn bid64_isSignaling(x: BID_UINT64) -> c_int;
    #[link_name = "__bid64_isCanonical"]
    pub fn bid64_isCanonical(x: BID_UINT64) -> c_int;
    #[link_name = "__bid64_isNaN"]
    pub fn bid64_isNaN(x: BID_UINT64) -> c_int;
    #[link_name = "__bid64_class"]
    pub fn bid64_class(x: BID_UINT64) -> class_t;

    // Classification: BID128
    #[link_name = "__bid128_isSigned"]
    pub fn bid128_isSigned(x: BID_UINT128) -> c_int;
    #[link_name = "__bid128_isNormal"]
    pub fn bid128_isNormal(x: BID_UINT128) -> c_int;
    #[link_name = "__bid128_isSubnormal"]
    pub fn bid128_isSubnormal(x: BID_UINT128) -> c_int;
    #[link_name = "__bid128_isFinite"]
    pub fn bid128_isFinite(x: BID_UINT128) -> c_int;
    #[link_name = "__bid128_isZero"]
    pub fn bid128_isZero(x: BID_UINT128) -> c_int;
    #[link_name = "__bid128_isInf"]
    pub fn bid128_isInf(x: BID_UINT128) -> c_int;
    #[link_name = "__bid128_isSignaling"]
    pub fn bid128_isSignaling(x: BID_UINT128) -> c_int;
    #[link_name = "__bid128_isCanonical"]
    pub fn bid128_isCanonical(x: BID_UINT128) -> c_int;
    #[link_name = "__bid128_isNaN"]
    pub fn bid128_isNaN(x: BID_UINT128) -> c_int;
    #[link_name = "__bid128_class"]
    pub fn bid128_class(x: BID_UINT128) -> class_t;
}
