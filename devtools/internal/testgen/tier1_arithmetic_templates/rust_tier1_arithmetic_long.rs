#![cfg(feature = "tier1-long")]
#![allow(clippy::too_many_arguments)]

use bid754::generated::add64::{bid64_add, bid64_sub};
use bid754::generated::bid32_exports::{bid32_add, bid32_div, bid32_mul, bid32_sub};
use bid754::generated::div64::bid64_div;
use bid754::generated::mul64::bid64_mul;
use bid754::{Decimal128, Decimal32, Decimal64, ExceptionFlags, RoundingMode};
use libbid_sys::BID_UINT128 as C128;
use std::ffi::c_long;

const BOUNDARY32_COUNT: u64 = @@TIER1_BOUNDARY32_COUNT@@;
const BOUNDARY64_COUNT: u64 = @@TIER1_BOUNDARY64_COUNT@@;
const BOUNDARY128_COUNT: u64 = @@TIER1_BOUNDARY128_COUNT@@;
const STRUCTURED32_COUNT: u64 = @@TIER1_STRUCTURED32_COUNT@@;
const STRUCTURED64_COUNT: u64 = @@TIER1_STRUCTURED64_COUNT@@;
const STRUCTURED128_COUNT: u64 = @@TIER1_STRUCTURED128_COUNT@@;
const RANDOM32_COUNT: u64 = @@TIER1_RANDOM32_COUNT@@;
const RANDOM64_COUNT: u64 = @@TIER1_RANDOM64_COUNT@@;
const RANDOM128_COUNT: u64 = @@TIER1_RANDOM128_COUNT@@;
const TOTAL32_COUNT: u64 = @@TIER1_TOTAL32_COUNT@@;
const TOTAL64_COUNT: u64 = @@TIER1_TOTAL64_COUNT@@;
const TOTAL128_COUNT: u64 = @@TIER1_TOTAL128_COUNT@@;
const SEMANTIC_ROUNDED32_COUNT: usize = @@TIER1_SEMANTIC_ROUNDED32_COUNT@@;
const SEMANTIC_ROUNDED64_COUNT: usize = @@TIER1_SEMANTIC_ROUNDED64_COUNT@@;
const SEMANTIC_ROUNDED128_COUNT: usize = @@TIER1_SEMANTIC_ROUNDED128_COUNT@@;
const SEMANTIC_SCALE32_COUNT: usize = @@TIER1_SEMANTIC_SCALE32_COUNT@@;
const SEMANTIC_SCALE64_COUNT: usize = @@TIER1_SEMANTIC_SCALE64_COUNT@@;
const SEMANTIC_SCALE128_COUNT: usize = @@TIER1_SEMANTIC_SCALE128_COUNT@@;
const SEMANTIC_REMAINDER32_COUNT: usize = @@TIER1_SEMANTIC_REMAINDER32_COUNT@@;
const SEMANTIC_REMAINDER64_COUNT: usize = @@TIER1_SEMANTIC_REMAINDER64_COUNT@@;
const SEMANTIC_REMAINDER128_COUNT: usize = @@TIER1_SEMANTIC_REMAINDER128_COUNT@@;
const SEMANTIC_FMA32_COUNT: usize = @@TIER1_SEMANTIC_FMA32_COUNT@@;
const SEMANTIC_FMA64_COUNT: usize = @@TIER1_SEMANTIC_FMA64_COUNT@@;
const SEMANTIC_FMA128_COUNT: usize = @@TIER1_SEMANTIC_FMA128_COUNT@@;
const SEMANTIC_SQRT32_COUNT: usize = @@TIER1_SEMANTIC_SQRT32_COUNT@@;
const SEMANTIC_SQRT64_COUNT: usize = @@TIER1_SEMANTIC_SQRT64_COUNT@@;
const SEMANTIC_SQRT128_COUNT: usize = @@TIER1_SEMANTIC_SQRT128_COUNT@@;
const RANDOM_OPS: u64 = 10;
const RANDOM_CASES32: u64 = @@TIER1_RANDOM_CASES_PER_OP32@@;
const RANDOM_CASES64: u64 = @@TIER1_RANDOM_CASES_PER_OP64@@;
const RANDOM_CASES128: u64 = @@TIER1_RANDOM_CASES_PER_OP128@@;
const SCALE_FINITE_TRANSITION_LIMIT32: u64 = @@TIER1_SCALE_FINITE_TRANSITION_LIMIT32@@;
const SCALE_FINITE_TRANSITION_LIMIT64: u64 = @@TIER1_SCALE_FINITE_TRANSITION_LIMIT64@@;
const SCALE_FINITE_TRANSITION_LIMIT128: u64 = @@TIER1_SCALE_FINITE_TRANSITION_LIMIT128@@;
const SCALE_RANDOM_STRATA: u64 = @@TIER1_SCALE_RANDOM_STRATA@@;
const SCALE_MODE_CROSS: u64 = @@TIER1_SCALE_MODE_CROSS@@;
const SCALE_MODE_CROSS_GROUPS32: u64 = @@TIER1_SCALE_MODE_CROSS_GROUPS32@@;
const SCALE_MODE_CROSS_GROUPS64: u64 = @@TIER1_SCALE_MODE_CROSS_GROUPS64@@;
const SCALE_MODE_CROSS_GROUPS128: u64 = @@TIER1_SCALE_MODE_CROSS_GROUPS128@@;
const SCALE_TUPLE_HASH32: u64 = @@TIER1_SCALE_TUPLE_HASH32@@;
const SCALE_TUPLE_HASH64: u64 = @@TIER1_SCALE_TUPLE_HASH64@@;
const SCALE_TUPLE_HASH128: u64 = @@TIER1_SCALE_TUPLE_HASH128@@;
const PAIR_STREAM_HASH32: u64 = @@TIER1_PAIR_STREAM_HASH32@@;
const PAIR_STREAM_HASH64: u64 = @@TIER1_PAIR_STREAM_HASH64@@;
const PAIR_STREAM_HASH128: u64 = @@TIER1_PAIR_STREAM_HASH128@@;
const FMA_TRIPLE_STREAM_HASH32: u64 = @@TIER1_FMA_TRIPLE_STREAM_HASH32@@;
const FMA_TRIPLE_STREAM_HASH64: u64 = @@TIER1_FMA_TRIPLE_STREAM_HASH64@@;
const FMA_TRIPLE_STREAM_HASH128: u64 = @@TIER1_FMA_TRIPLE_STREAM_HASH128@@;
const RANDOM_STREAM_HASH32: u64 = @@TIER1_RANDOM_STREAM_HASH32@@;
const RANDOM_STREAM_HASH64: u64 = @@TIER1_RANDOM_STREAM_HASH64@@;
const RANDOM_STREAM_HASH128: u64 = @@TIER1_RANDOM_STREAM_HASH128@@;
const SCALE_TUPLE_HASH_OFFSET: u64 = 14695981039346656037;
const SCALE_TUPLE_HASH_PRIME: u64 = 1099511628211;

// The deterministic-random seeds are shared by the differential blocks and
// the corpus stream contract, so a seed edit that reaches only one of them is
// impossible by construction.
const RANDOM_ROUNDED_SEED_BASE32: u64 = 0xdec7543200000000;
const RANDOM_UNROUNDED_SEED_BASE32: u64 = 0xdec7543210000000;
const RANDOM_FMA_SEED32: u64 = 0xdec7543220000000;
const RANDOM_SQRT_SEED32: u64 = 0xdec7543230000000;
const RANDOM_ROUNDED_SEED_BASE64: u64 = 0xdec7546400000000;
const RANDOM_UNROUNDED_SEED_BASE64: u64 = 0xdec7546410000000;
const RANDOM_FMA_SEED64: u64 = 0xdec7546420000000;
const RANDOM_SQRT_SEED64: u64 = 0xdec7546430000000;
const RANDOM_ROUNDED_SEED_BASE128: u64 = 0xdec7541200000000;
const RANDOM_UNROUNDED_SEED_BASE128: u64 = 0xdec7541210000000;
const RANDOM_FMA_SEED128: u64 = 0xdec7541220000000;
const RANDOM_SQRT_SEED128: u64 = 0xdec7541230000000;
const SCALE_SEED32: u64 = 0xdec7543253414c45;
const SCALE_SEED64: u64 = 0xdec7546453414c45;
const SCALE_SEED128: u64 = 0xdec7541253414c45;

#[derive(Clone, Copy, Debug)]
struct Mode {
    name: &'static str,
    public: RoundingMode,
    native: u32,
}

const MODES: [Mode; 5] = [
    Mode { name: "nearest_even", public: RoundingMode::NearestEven, native: 0 },
    Mode { name: "nearest_away", public: RoundingMode::NearestAway, native: 4 },
    Mode { name: "toward_zero", public: RoundingMode::TowardZero, native: 3 },
    Mode { name: "toward_positive", public: RoundingMode::TowardPositive, native: 2 },
    Mode { name: "toward_negative", public: RoundingMode::TowardNegative, native: 1 },
];

#[derive(Clone, Copy, Debug)]
enum RoundedOp { Add, Sub, Mul, Div, Quantize }

#[derive(Clone, Copy, Debug)]
enum UnroundedOp { Remainder, Fmod }

const ROUNDED_OPS: [RoundedOp; 5] = [
    RoundedOp::Add, RoundedOp::Sub, RoundedOp::Mul, RoundedOp::Div, RoundedOp::Quantize,
];
const UNROUNDED_OPS: [UnroundedOp; 2] = [UnroundedOp::Remainder, UnroundedOp::Fmod];

const SCALE_EXPONENTS: [i64; 25] = [
    i64::MIN, -2147483649, -2147483648, -6177, -6176, -1000, -399, -398,
    -102, -101, -1, 0, 1, 90, 91, 369, 370, 398, 399, 1000, 6176, 6177,
    2147483647, 2147483648, i64::MAX,
];

#[derive(Clone, Copy, Debug, PartialEq, Eq)]
struct Words { lo: u64, hi: u64 }

#[derive(Clone, Copy)]
struct Rounded32 { op: RoundedOp, x: u32, y: u32 }
#[derive(Clone, Copy)]
struct Rounded64 { op: RoundedOp, x: u64, y: u64 }
#[derive(Clone, Copy)]
struct Rounded128 { op: RoundedOp, x: Words, y: Words }
#[derive(Clone, Copy)]
struct Scale32 { x: u32, exponent: i64 }
#[derive(Clone, Copy)]
struct Scale64 { x: u64, exponent: i64 }
#[derive(Clone, Copy)]
struct Scale128 { x: Words, exponent: i64 }
#[derive(Clone, Copy)]
struct Pair32 { x: u32, y: u32 }
#[derive(Clone, Copy)]
struct Pair64 { x: u64, y: u64 }
#[derive(Clone, Copy)]
struct Pair128 { x: Words, y: Words }
#[derive(Clone, Copy)]
struct Triple32 { x: u32, y: u32, z: u32 }
#[derive(Clone, Copy)]
struct Triple64 { x: u64, y: u64, z: u64 }
#[derive(Clone, Copy)]
struct Triple128 { x: Words, y: Words, z: Words }

#[derive(Clone, Copy)]
struct Shard { count: u64, index: u64 }

impl Shard {
    fn load() -> Self {
        let count = std::env::var("BID754_TIER1_ARITH_SHARD_COUNT").ok();
        let index = std::env::var("BID754_TIER1_ARITH_SHARD_INDEX").ok();
        match (count, index) {
            (None, None) => Shard { count: 1, index: 0 },
            (Some(count), Some(index)) => {
                let count = count.parse::<u64>().expect("invalid BID754_TIER1_ARITH_SHARD_COUNT");
                let index = index.parse::<u64>().expect("invalid BID754_TIER1_ARITH_SHARD_INDEX");
                assert!(count > 0, "BID754_TIER1_ARITH_SHARD_COUNT must be positive");
                assert!(index < count, "Tier 1 arithmetic shard index must be below shard count");
                Shard { count, index }
            }
            _ => panic!("BID754_TIER1_ARITH_SHARD_COUNT and BID754_TIER1_ARITH_SHARD_INDEX must be set together"),
        }
    }

    fn owns(self, case_index: u64) -> bool { case_index % self.count == self.index }

    fn owned_count(self, total: u64) -> u64 {
        if total <= self.index { 0 } else { 1 + (total - 1 - self.index) / self.count }
    }
}

unsafe extern "C" {
    #[link_name = "__bid32_rem"]
    fn native_bid32_rem(x: u32, y: u32, flags: *mut u32) -> u32;
    #[link_name = "__bid32_fmod"]
    fn native_bid32_fmod(x: u32, y: u32, flags: *mut u32) -> u32;
    #[link_name = "__bid64_rem"]
    fn native_bid64_rem(x: u64, y: u64, flags: *mut u32) -> u64;
    #[link_name = "__bid64_fmod"]
    fn native_bid64_fmod(x: u64, y: u64, flags: *mut u32) -> u64;
    #[link_name = "__bid128_rem"]
    fn native_bid128_rem(x: C128, y: C128, flags: *mut u32) -> C128;
    #[link_name = "__bid128_fmod"]
    fn native_bid128_fmod(x: C128, y: C128, flags: *mut u32) -> C128;
    #[link_name = "__bid32_scalbln"]
    fn native_bid32_scalbln(x: u32, exponent: c_long, rounding: u32, flags: *mut u32) -> u32;
    #[link_name = "__bid64_scalbln"]
    fn native_bid64_scalbln(x: u64, exponent: c_long, rounding: u32, flags: *mut u32) -> u64;
    #[link_name = "__bid128_scalbln"]
    fn native_bid128_scalbln(x: C128, exponent: c_long, rounding: u32, flags: *mut u32) -> C128;
}

fn c128(words: Words) -> C128 { C128::new(words.lo, words.hi) }

fn decimal128(words: Words) -> Decimal128 {
    let mut raw = [0u8; 16];
    raw[0..8].copy_from_slice(&words.lo.to_le_bytes());
    raw[8..16].copy_from_slice(&words.hi.to_le_bytes());
    Decimal128::from_le_bytes(raw)
}

fn decimal128_words(value: Decimal128) -> Words {
    let raw = value.to_le_bytes();
    Words {
        lo: u64::from_le_bytes(raw[0..8].try_into().unwrap()),
        hi: u64::from_le_bytes(raw[8..16].try_into().unwrap()),
    }
}

fn c128_words(value: C128) -> Words { Words { lo: value.w[0], hi: value.w[1] } }

fn public_raw_flags(flags: ExceptionFlags) -> u32 {
    let known = ExceptionFlags::INEXACT | ExceptionFlags::UNDERFLOW | ExceptionFlags::OVERFLOW |
        ExceptionFlags::DIVISION_BY_ZERO | ExceptionFlags::INVALID_OPERATION;
    assert!(flags.difference(known).is_empty(), "unknown public flags: {flags}");
    let mut raw = 0u32;
    if flags.contains(ExceptionFlags::INVALID_OPERATION) { raw |= 0x01; }
    if flags.contains(ExceptionFlags::DIVISION_BY_ZERO) { raw |= 0x04; }
    if flags.contains(ExceptionFlags::OVERFLOW) { raw |= 0x08; }
    if flags.contains(ExceptionFlags::UNDERFLOW) { raw |= 0x10; }
    if flags.contains(ExceptionFlags::INEXACT) { raw |= 0x20; }
    raw
}

// legs_* compute every leg of one comparison from one decoded case. They are
// the single operand/mode/dispatch fan-out shared by the differential checks
// and the routing sentinels, so a slot swap, a mode miswire, or a
// dispatch-row mislabel in this glue skews the differential's legs
// identically (an agreed-upon wrong answer the differential cannot see)
// while the pinned sentinel rows diverge and fail.

fn legs_rounded32(op: RoundedOp, x: u32, y: u32, mode: Mode) -> (u32, u32, u32, u32, Option<u32>) {
    let mut native_flags = 0u32;
    let native = unsafe {
        match op {
            RoundedOp::Add => libbid_sys::bid32_add(x, y, mode.native, &mut native_flags),
            RoundedOp::Sub => libbid_sys::bid32_sub(x, y, mode.native, &mut native_flags),
            RoundedOp::Mul => libbid_sys::bid32_mul(x, y, mode.native, &mut native_flags),
            RoundedOp::Div => libbid_sys::bid32_div(x, y, mode.native, &mut native_flags),
            RoundedOp::Quantize => libbid_sys::bid32_quantize(x, y, mode.native, &mut native_flags),
        }
    };
    let left = Decimal32::from_bits(x);
    let right = Decimal32::from_bits(y);
    let (public, flags) = match op {
        RoundedOp::Add => left.add_with_mode(right, mode.public),
        RoundedOp::Sub => left.sub_with_mode(right, mode.public),
        RoundedOp::Mul => left.mul_with_mode(right, mode.public),
        RoundedOp::Div => left.div_with_mode(right, mode.public),
        RoundedOp::Quantize => left.quantize_with_mode(right, mode.public),
    };
    let flagless = match op {
        RoundedOp::Add => Some(bid32_add(x, y, i64::from(mode.native))),
        RoundedOp::Sub => Some(bid32_sub(x, y, i64::from(mode.native))),
        RoundedOp::Mul => Some(bid32_mul(x, y, i64::from(mode.native))),
        RoundedOp::Div => Some(bid32_div(x, y, i64::from(mode.native))),
        RoundedOp::Quantize => None,
    };
    (native, native_flags, public.to_bits(), public_raw_flags(flags), flagless)
}

fn check_rounded32(op: RoundedOp, x: u32, y: u32, mode: Mode) {
    let (native, native_flags, public, public_flags, flagless) = legs_rounded32(op, x, y, mode);
    assert_eq!((public, public_flags), (native, native_flags),
        "Decimal32 {op:?} mismatch x={x:08x} y={y:08x} mode={}", mode.name);
    if let Some(flagless) = flagless {
        assert_eq!(flagless, native,
            "Decimal32 flagless {op:?} mismatch x={x:08x} y={y:08x} mode={}", mode.name);
    }
}

fn legs_rounded64(op: RoundedOp, x: u64, y: u64, mode: Mode) -> (u64, u32, u64, u32, Option<u64>) {
    let mut native_flags = 0u32;
    let native = unsafe {
        match op {
            RoundedOp::Add => libbid_sys::bid64_add(x, y, mode.native, &mut native_flags),
            RoundedOp::Sub => libbid_sys::bid64_sub(x, y, mode.native, &mut native_flags),
            RoundedOp::Mul => libbid_sys::bid64_mul(x, y, mode.native, &mut native_flags),
            RoundedOp::Div => libbid_sys::bid64_div(x, y, mode.native, &mut native_flags),
            RoundedOp::Quantize => libbid_sys::bid64_quantize(x, y, mode.native, &mut native_flags),
        }
    };
    let left = Decimal64::from_bits(x);
    let right = Decimal64::from_bits(y);
    let (public, flags) = match op {
        RoundedOp::Add => left.add_with_mode(right, mode.public),
        RoundedOp::Sub => left.sub_with_mode(right, mode.public),
        RoundedOp::Mul => left.mul_with_mode(right, mode.public),
        RoundedOp::Div => left.div_with_mode(right, mode.public),
        RoundedOp::Quantize => left.quantize_with_mode(right, mode.public),
    };
    let flagless = match op {
        RoundedOp::Add => Some(bid64_add(x, y, i64::from(mode.native))),
        RoundedOp::Sub => Some(bid64_sub(x, y, i64::from(mode.native))),
        RoundedOp::Mul => Some(bid64_mul(x, y, i64::from(mode.native))),
        RoundedOp::Div => Some(bid64_div(x, y, i64::from(mode.native))),
        RoundedOp::Quantize => None,
    };
    (native, native_flags, public.to_bits(), public_raw_flags(flags), flagless)
}

fn check_rounded64(op: RoundedOp, x: u64, y: u64, mode: Mode) {
    let (native, native_flags, public, public_flags, flagless) = legs_rounded64(op, x, y, mode);
    assert_eq!((public, public_flags), (native, native_flags),
        "Decimal64 {op:?} mismatch x={x:016x} y={y:016x} mode={}", mode.name);
    if let Some(flagless) = flagless {
        assert_eq!(flagless, native,
            "Decimal64 flagless {op:?} mismatch x={x:016x} y={y:016x} mode={}", mode.name);
    }
}

fn legs_rounded128(op: RoundedOp, x: Words, y: Words, mode: Mode) -> (Words, u32, Words, u32) {
    let mut native_flags = 0u32;
    let native = unsafe {
        match op {
            RoundedOp::Add => libbid_sys::bid128_add(c128(x), c128(y), mode.native, &mut native_flags),
            RoundedOp::Sub => libbid_sys::bid128_sub(c128(x), c128(y), mode.native, &mut native_flags),
            RoundedOp::Mul => libbid_sys::bid128_mul(c128(x), c128(y), mode.native, &mut native_flags),
            RoundedOp::Div => libbid_sys::bid128_div(c128(x), c128(y), mode.native, &mut native_flags),
            RoundedOp::Quantize => libbid_sys::bid128_quantize(c128(x), c128(y), mode.native, &mut native_flags),
        }
    };
    let left = decimal128(x);
    let right = decimal128(y);
    let (public, flags) = match op {
        RoundedOp::Add => left.add_with_mode(right, mode.public),
        RoundedOp::Sub => left.sub_with_mode(right, mode.public),
        RoundedOp::Mul => left.mul_with_mode(right, mode.public),
        RoundedOp::Div => left.div_with_mode(right, mode.public),
        RoundedOp::Quantize => left.quantize_with_mode(right, mode.public),
    };
    (c128_words(native), native_flags, decimal128_words(public), public_raw_flags(flags))
}

fn check_rounded128(op: RoundedOp, x: Words, y: Words, mode: Mode) {
    let (native, native_flags, public, public_flags) = legs_rounded128(op, x, y, mode);
    assert_eq!((public, public_flags), (native, native_flags),
        "Decimal128 {op:?} mismatch x={x:?} y={y:?} mode={}", mode.name);
}

fn legs_fma32(x: u32, y: u32, z: u32, mode: Mode) -> (u32, u32, u32, u32) {
    let mut native_flags = 0u32;
    let native = unsafe { libbid_sys::bid32_fma(x, y, z, mode.native, &mut native_flags) };
    let (public, flags) = Decimal32::from_bits(x)
        .fma_with_mode(Decimal32::from_bits(y), Decimal32::from_bits(z), mode.public);
    (native, native_flags, public.to_bits(), public_raw_flags(flags))
}

fn check_fma32(x: u32, y: u32, z: u32, mode: Mode) {
    let (native, native_flags, public, public_flags) = legs_fma32(x, y, z, mode);
    assert_eq!((public, public_flags), (native, native_flags),
        "Decimal32 fma mismatch x={x:08x} y={y:08x} z={z:08x} mode={}", mode.name);
}

fn legs_fma64(x: u64, y: u64, z: u64, mode: Mode) -> (u64, u32, u64, u32) {
    let mut native_flags = 0u32;
    let native = unsafe { libbid_sys::bid64_fma(x, y, z, mode.native, &mut native_flags) };
    let (public, flags) = Decimal64::from_bits(x)
        .fma_with_mode(Decimal64::from_bits(y), Decimal64::from_bits(z), mode.public);
    (native, native_flags, public.to_bits(), public_raw_flags(flags))
}

fn check_fma64(x: u64, y: u64, z: u64, mode: Mode) {
    let (native, native_flags, public, public_flags) = legs_fma64(x, y, z, mode);
    assert_eq!((public, public_flags), (native, native_flags),
        "Decimal64 fma mismatch x={x:016x} y={y:016x} z={z:016x} mode={}", mode.name);
}

fn legs_fma128(x: Words, y: Words, z: Words, mode: Mode) -> (Words, u32, Words, u32) {
    let mut native_flags = 0u32;
    let native = unsafe { libbid_sys::bid128_fma(c128(x), c128(y), c128(z), mode.native, &mut native_flags) };
    let (public, flags) = decimal128(x).fma_with_mode(decimal128(y), decimal128(z), mode.public);
    (c128_words(native), native_flags, decimal128_words(public), public_raw_flags(flags))
}

fn check_fma128(x: Words, y: Words, z: Words, mode: Mode) {
    let (native, native_flags, public, public_flags) = legs_fma128(x, y, z, mode);
    assert_eq!((public, public_flags), (native, native_flags),
        "Decimal128 fma mismatch x={x:?} y={y:?} z={z:?} mode={}", mode.name);
}

fn legs_sqrt32(x: u32, mode: Mode) -> (u32, u32, u32, u32) {
    let mut native_flags = 0u32;
    let native = unsafe { libbid_sys::bid32_sqrt(x, mode.native, &mut native_flags) };
    let (public, flags) = Decimal32::from_bits(x).sqrt_with_mode(mode.public);
    (native, native_flags, public.to_bits(), public_raw_flags(flags))
}

fn check_sqrt32(x: u32, mode: Mode) {
    let (native, native_flags, public, public_flags) = legs_sqrt32(x, mode);
    assert_eq!((public, public_flags), (native, native_flags),
        "Decimal32 sqrt mismatch x={x:08x} mode={}", mode.name);
}

fn legs_sqrt64(x: u64, mode: Mode) -> (u64, u32, u64, u32) {
    let mut native_flags = 0u32;
    let native = unsafe { libbid_sys::bid64_sqrt(x, mode.native, &mut native_flags) };
    let (public, flags) = Decimal64::from_bits(x).sqrt_with_mode(mode.public);
    (native, native_flags, public.to_bits(), public_raw_flags(flags))
}

fn check_sqrt64(x: u64, mode: Mode) {
    let (native, native_flags, public, public_flags) = legs_sqrt64(x, mode);
    assert_eq!((public, public_flags), (native, native_flags),
        "Decimal64 sqrt mismatch x={x:016x} mode={}", mode.name);
}

fn legs_sqrt128(x: Words, mode: Mode) -> (Words, u32, Words, u32) {
    let mut native_flags = 0u32;
    let native = unsafe { libbid_sys::bid128_sqrt(c128(x), mode.native, &mut native_flags) };
    let (public, flags) = decimal128(x).sqrt_with_mode(mode.public);
    (c128_words(native), native_flags, decimal128_words(public), public_raw_flags(flags))
}

fn check_sqrt128(x: Words, mode: Mode) {
    let (native, native_flags, public, public_flags) = legs_sqrt128(x, mode);
    assert_eq!((public, public_flags), (native, native_flags),
        "Decimal128 sqrt mismatch x={x:?} mode={}", mode.name);
}

fn legs_unrounded32(op: UnroundedOp, x: u32, y: u32) -> (u32, u32, u32, u32) {
    let mut native_flags = 0u32;
    let native = unsafe { match op {
        UnroundedOp::Remainder => native_bid32_rem(x, y, &mut native_flags),
        UnroundedOp::Fmod => native_bid32_fmod(x, y, &mut native_flags),
    }};
    let left = Decimal32::from_bits(x);
    let right = Decimal32::from_bits(y);
    let (public, flags) = match op {
        UnroundedOp::Remainder => left.remainder(right),
        UnroundedOp::Fmod => left.fmod(right),
    };
    (native, native_flags, public.to_bits(), public_raw_flags(flags))
}

fn check_unrounded32(op: UnroundedOp, x: u32, y: u32) {
    let (native, native_flags, public, public_flags) = legs_unrounded32(op, x, y);
    assert_eq!((public, public_flags), (native, native_flags),
        "Decimal32 {op:?} mismatch x={x:08x} y={y:08x}");
}

fn legs_unrounded64(op: UnroundedOp, x: u64, y: u64) -> (u64, u32, u64, u32) {
    let mut native_flags = 0u32;
    let native = unsafe { match op {
        UnroundedOp::Remainder => native_bid64_rem(x, y, &mut native_flags),
        UnroundedOp::Fmod => native_bid64_fmod(x, y, &mut native_flags),
    }};
    let left = Decimal64::from_bits(x);
    let right = Decimal64::from_bits(y);
    let (public, flags) = match op {
        UnroundedOp::Remainder => left.remainder(right),
        UnroundedOp::Fmod => left.fmod(right),
    };
    (native, native_flags, public.to_bits(), public_raw_flags(flags))
}

fn check_unrounded64(op: UnroundedOp, x: u64, y: u64) {
    let (native, native_flags, public, public_flags) = legs_unrounded64(op, x, y);
    assert_eq!((public, public_flags), (native, native_flags),
        "Decimal64 {op:?} mismatch x={x:016x} y={y:016x}");
}

fn legs_unrounded128(op: UnroundedOp, x: Words, y: Words) -> (Words, u32, Words, u32) {
    let mut native_flags = 0u32;
    let native = unsafe { match op {
        UnroundedOp::Remainder => native_bid128_rem(c128(x), c128(y), &mut native_flags),
        UnroundedOp::Fmod => native_bid128_fmod(c128(x), c128(y), &mut native_flags),
    }};
    let left = decimal128(x);
    let right = decimal128(y);
    let (public, flags) = match op {
        UnroundedOp::Remainder => left.remainder(right),
        UnroundedOp::Fmod => left.fmod(right),
    };
    (c128_words(native), native_flags, decimal128_words(public), public_raw_flags(flags))
}

fn check_unrounded128(op: UnroundedOp, x: Words, y: Words) {
    let (native, native_flags, public, public_flags) = legs_unrounded128(op, x, y);
    assert_eq!((public, public_flags), (native, native_flags),
        "Decimal128 {op:?} mismatch x={x:?} y={y:?}");
}

fn legs_scale32(x: u32, exponent: i64, mode: Mode) -> (u32, u32, u32, u32) {
    let mut native_flags = 0u32;
    let native = unsafe { native_bid32_scalbln(x, exponent as c_long, mode.native, &mut native_flags) };
    let (public, flags) = Decimal32::from_bits(x).scaleb_with_mode(exponent, mode.public);
    (native, native_flags, public.to_bits(), public_raw_flags(flags))
}

fn check_scale32(x: u32, exponent: i64, mode: Mode) {
    let (native, native_flags, public, public_flags) = legs_scale32(x, exponent, mode);
    assert_eq!((public, public_flags), (native, native_flags),
        "Decimal32 scaleB mismatch x={x:08x} exponent={exponent} mode={}", mode.name);
}

fn legs_scale64(x: u64, exponent: i64, mode: Mode) -> (u64, u32, u64, u32) {
    let mut native_flags = 0u32;
    let native = unsafe { native_bid64_scalbln(x, exponent as c_long, mode.native, &mut native_flags) };
    let (public, flags) = Decimal64::from_bits(x).scaleb_with_mode(exponent, mode.public);
    (native, native_flags, public.to_bits(), public_raw_flags(flags))
}

fn check_scale64(x: u64, exponent: i64, mode: Mode) {
    let (native, native_flags, public, public_flags) = legs_scale64(x, exponent, mode);
    assert_eq!((public, public_flags), (native, native_flags),
        "Decimal64 scaleB mismatch x={x:016x} exponent={exponent} mode={}", mode.name);
}

fn legs_scale128(x: Words, exponent: i64, mode: Mode) -> (Words, u32, Words, u32) {
    let mut native_flags = 0u32;
    let native = unsafe { native_bid128_scalbln(c128(x), exponent as c_long, mode.native, &mut native_flags) };
    let (public, flags) = decimal128(x).scaleb_with_mode(exponent, mode.public);
    (c128_words(native), native_flags, decimal128_words(public), public_raw_flags(flags))
}

fn check_scale128(x: Words, exponent: i64, mode: Mode) {
    let (native, native_flags, public, public_flags) = legs_scale128(x, exponent, mode);
    assert_eq!((public, public_flags), (native, native_flags),
        "Decimal128 scaleB mismatch x={x:?} exponent={exponent} mode={}", mode.name);
}

fn visit_pairs32(mut visit: impl FnMut(u32, u32)) {
    for &x in BOUNDARY32 { for &y in &PROBES32 { visit(x, y); visit(y, x); } }
    for &x in &PROBES32 { for &y in &PROBES32 { visit(x, y); } }
}

fn visit_pairs64(mut visit: impl FnMut(u64, u64)) {
    for &x in BOUNDARY64 { for &y in &PROBES64 { visit(x, y); visit(y, x); } }
    for &x in &PROBES64 { for &y in &PROBES64 { visit(x, y); } }
}

fn visit_pairs128(mut visit: impl FnMut(Words, Words)) {
    for &x in BOUNDARY128 { for &y in &PROBES128 { visit(x, y); visit(y, x); } }
    for &x in &PROBES128 { for &y in &PROBES128 { visit(x, y); } }
}

fn visit_triples32(mut visit: impl FnMut(u32, u32, u32)) {
    for (i, &x) in BOUNDARY32.iter().enumerate() {
        for (j, &y) in PROBES32.iter().enumerate() {
            let z = PROBES32[(i + j) % PROBES32.len()];
            visit(x, y, z);
            visit(y, z, x);
            visit(z, x, y);
        }
    }
    for &x in &PROBES32 { for &y in &PROBES32 { for &z in &PROBES32 { visit(x, y, z); } } }
}

fn visit_triples64(mut visit: impl FnMut(u64, u64, u64)) {
    for (i, &x) in BOUNDARY64.iter().enumerate() {
        for (j, &y) in PROBES64.iter().enumerate() {
            let z = PROBES64[(i + j) % PROBES64.len()];
            visit(x, y, z);
            visit(y, z, x);
            visit(z, x, y);
        }
    }
    for &x in &PROBES64 { for &y in &PROBES64 { for &z in &PROBES64 { visit(x, y, z); } } }
}

fn visit_triples128(mut visit: impl FnMut(Words, Words, Words)) {
    for (i, &x) in BOUNDARY128.iter().enumerate() {
        for (j, &y) in PROBES128.iter().enumerate() {
            let z = PROBES128[(i + j) % PROBES128.len()];
            visit(x, y, z);
            visit(y, z, x);
            visit(z, x, y);
        }
    }
    for &x in &PROBES128 { for &y in &PROBES128 { for &z in &PROBES128 { visit(x, y, z); } } }
}

// Recomputes the pair, fma-triple, and non-ScaleB random operand streams
// through the exact visit functions and seed/lane/mode derivations the
// differentials consume, comparing the digests against the generator-anchored
// constants. The Go runner carries the same contract, so a corpus drift in
// either template fails here instead of silently splitting the languages.
fn assert_corpus_stream_contracts() {
    let mut digest = scale_tuple_hash_mix(SCALE_TUPLE_HASH_OFFSET, 32);
    let mut visits = 0u64;
    visit_pairs32(|x, y| {
        digest = scale_tuple_hash_mix(digest, x as u64);
        digest = scale_tuple_hash_mix(digest, 0);
        digest = scale_tuple_hash_mix(digest, y as u64);
        digest = scale_tuple_hash_mix(digest, 0);
        visits += 1;
    });
    assert_eq!(scale_tuple_hash_mix(digest, visits), PAIR_STREAM_HASH32, "Decimal32 pair stream hash drift");
    let mut digest = scale_tuple_hash_mix(SCALE_TUPLE_HASH_OFFSET, 64);
    let mut visits = 0u64;
    visit_pairs64(|x, y| {
        digest = scale_tuple_hash_mix(digest, x);
        digest = scale_tuple_hash_mix(digest, 0);
        digest = scale_tuple_hash_mix(digest, y);
        digest = scale_tuple_hash_mix(digest, 0);
        visits += 1;
    });
    assert_eq!(scale_tuple_hash_mix(digest, visits), PAIR_STREAM_HASH64, "Decimal64 pair stream hash drift");
    let mut digest = scale_tuple_hash_mix(SCALE_TUPLE_HASH_OFFSET, 128);
    let mut visits = 0u64;
    visit_pairs128(|x, y| {
        digest = scale_tuple_hash_mix(digest, x.lo);
        digest = scale_tuple_hash_mix(digest, x.hi);
        digest = scale_tuple_hash_mix(digest, y.lo);
        digest = scale_tuple_hash_mix(digest, y.hi);
        visits += 1;
    });
    assert_eq!(scale_tuple_hash_mix(digest, visits), PAIR_STREAM_HASH128, "Decimal128 pair stream hash drift");

    let mut digest = scale_tuple_hash_mix(SCALE_TUPLE_HASH_OFFSET, 32);
    let mut visits = 0u64;
    visit_triples32(|x, y, z| {
        digest = scale_tuple_hash_mix(digest, x as u64);
        digest = scale_tuple_hash_mix(digest, 0);
        digest = scale_tuple_hash_mix(digest, y as u64);
        digest = scale_tuple_hash_mix(digest, 0);
        digest = scale_tuple_hash_mix(digest, z as u64);
        digest = scale_tuple_hash_mix(digest, 0);
        visits += 1;
    });
    assert_eq!(scale_tuple_hash_mix(digest, visits), FMA_TRIPLE_STREAM_HASH32, "Decimal32 fma triple stream hash drift");
    let mut digest = scale_tuple_hash_mix(SCALE_TUPLE_HASH_OFFSET, 64);
    let mut visits = 0u64;
    visit_triples64(|x, y, z| {
        digest = scale_tuple_hash_mix(digest, x);
        digest = scale_tuple_hash_mix(digest, 0);
        digest = scale_tuple_hash_mix(digest, y);
        digest = scale_tuple_hash_mix(digest, 0);
        digest = scale_tuple_hash_mix(digest, z);
        digest = scale_tuple_hash_mix(digest, 0);
        visits += 1;
    });
    assert_eq!(scale_tuple_hash_mix(digest, visits), FMA_TRIPLE_STREAM_HASH64, "Decimal64 fma triple stream hash drift");
    let mut digest = scale_tuple_hash_mix(SCALE_TUPLE_HASH_OFFSET, 128);
    let mut visits = 0u64;
    visit_triples128(|x, y, z| {
        digest = scale_tuple_hash_mix(digest, x.lo);
        digest = scale_tuple_hash_mix(digest, x.hi);
        digest = scale_tuple_hash_mix(digest, y.lo);
        digest = scale_tuple_hash_mix(digest, y.hi);
        digest = scale_tuple_hash_mix(digest, z.lo);
        digest = scale_tuple_hash_mix(digest, z.hi);
        visits += 1;
    });
    assert_eq!(scale_tuple_hash_mix(digest, visits), FMA_TRIPLE_STREAM_HASH128, "Decimal128 fma triple stream hash drift");

    let random_stream = |bits: u32, rounded_base: u64, unrounded_base: u64, fma_seed: u64, sqrt_seed: u64, cases: u64| -> u64 {
        let mut digest = scale_tuple_hash_mix(SCALE_TUPLE_HASH_OFFSET, bits as u64);
        let mix_operand = |digest: &mut u64, seed: u64, case_index: u64, lane: u64| {
            match bits {
                32 => *digest = mix_operand32(*digest, random_operand32(seed, case_index, lane)),
                64 => *digest = mix_operand64(*digest, random_operand64(seed, case_index, lane)),
                _ => *digest = mix_operand128(*digest, random_operand128(seed, case_index, lane)),
            }
        };
        let mut total = 0u64;
        for op_index in 0..ROUNDED_OPS.len() as u64 {
            let seed = rounded_base ^ op_index;
            for i in 0..cases {
                mix_operand(&mut digest, seed, i, 0);
                mix_operand(&mut digest, seed, i, 1);
                digest = scale_tuple_hash_mix(digest, MODES[i as usize % MODES.len()].native as u64);
                total += 1;
            }
        }
        for op_index in 0..UNROUNDED_OPS.len() as u64 {
            let seed = unrounded_base ^ op_index;
            for i in 0..cases {
                mix_operand(&mut digest, seed, i, 0);
                mix_operand(&mut digest, seed, i, 1);
                total += 1;
            }
        }
        for i in 0..cases {
            mix_operand(&mut digest, fma_seed, i, 0);
            mix_operand(&mut digest, fma_seed, i, 1);
            mix_operand(&mut digest, fma_seed, i, 2);
            digest = scale_tuple_hash_mix(digest, MODES[i as usize % MODES.len()].native as u64);
            total += 1;
        }
        for i in 0..cases {
            mix_operand(&mut digest, sqrt_seed, i, 0);
            digest = scale_tuple_hash_mix(digest, MODES[i as usize % MODES.len()].native as u64);
            total += 1;
        }
        scale_tuple_hash_mix(digest, total)
    };
    assert_eq!(random_stream(32, RANDOM_ROUNDED_SEED_BASE32, RANDOM_UNROUNDED_SEED_BASE32, RANDOM_FMA_SEED32, RANDOM_SQRT_SEED32, RANDOM_CASES32), RANDOM_STREAM_HASH32, "Decimal32 random stream hash drift");
    assert_eq!(random_stream(64, RANDOM_ROUNDED_SEED_BASE64, RANDOM_UNROUNDED_SEED_BASE64, RANDOM_FMA_SEED64, RANDOM_SQRT_SEED64, RANDOM_CASES64), RANDOM_STREAM_HASH64, "Decimal64 random stream hash drift");
    assert_eq!(random_stream(128, RANDOM_ROUNDED_SEED_BASE128, RANDOM_UNROUNDED_SEED_BASE128, RANDOM_FMA_SEED128, RANDOM_SQRT_SEED128, RANDOM_CASES128), RANDOM_STREAM_HASH128, "Decimal128 random stream hash drift");
}

#[test]
fn tier1_arithmetic_corpus_contract() {
    assert_eq!(std::mem::size_of::<c_long>(), 8, "Tier 1 native oracle requires LP64");
    assert_eq!(SCALE_MODE_CROSS as usize, MODES.len(), "ScaleB mode cross must equal the rounding-mode census");
    assert_eq!(BOUNDARY32.len(), BOUNDARY32_COUNT as usize);
    assert_eq!(BOUNDARY64.len(), BOUNDARY64_COUNT as usize);
    assert_eq!(BOUNDARY128.len(), BOUNDARY128_COUNT as usize);
    assert_eq!(SEMANTIC_ROUNDED32.len(), SEMANTIC_ROUNDED32_COUNT);
    assert_eq!(SEMANTIC_ROUNDED64.len(), SEMANTIC_ROUNDED64_COUNT);
    assert_eq!(SEMANTIC_ROUNDED128.len(), SEMANTIC_ROUNDED128_COUNT);
    assert_eq!(SEMANTIC_SCALE32.len(), SEMANTIC_SCALE32_COUNT);
    assert_eq!(SEMANTIC_SCALE64.len(), SEMANTIC_SCALE64_COUNT);
    assert_eq!(SEMANTIC_SCALE128.len(), SEMANTIC_SCALE128_COUNT);
    assert_eq!(SEMANTIC_REMAINDER32.len(), SEMANTIC_REMAINDER32_COUNT);
    assert_eq!(SEMANTIC_REMAINDER64.len(), SEMANTIC_REMAINDER64_COUNT);
    assert_eq!(SEMANTIC_REMAINDER128.len(), SEMANTIC_REMAINDER128_COUNT);
    assert_eq!(SEMANTIC_FMA32.len(), SEMANTIC_FMA32_COUNT);
    assert_eq!(SEMANTIC_FMA64.len(), SEMANTIC_FMA64_COUNT);
    assert_eq!(SEMANTIC_FMA128.len(), SEMANTIC_FMA128_COUNT);
    assert_eq!(SEMANTIC_SQRT32.len(), SEMANTIC_SQRT32_COUNT);
    assert_eq!(SEMANTIC_SQRT64.len(), SEMANTIC_SQRT64_COUNT);
    assert_eq!(SEMANTIC_SQRT128.len(), SEMANTIC_SQRT128_COUNT);
    assert_eq!(RANDOM32_COUNT, RANDOM_CASES32 * RANDOM_OPS);
    assert_eq!(RANDOM64_COUNT, RANDOM_CASES64 * RANDOM_OPS);
    assert_eq!(RANDOM128_COUNT, RANDOM_CASES128 * RANDOM_OPS);
    assert_eq!(random_word(0xdec7543200000000, 0, 0), @@TIER1_RANDOM_SAMPLE0@@);
    assert_eq!(random_word(0xdec7546400000004, (1 << 20) - 1, 1), @@TIER1_RANDOM_SAMPLE1@@);
    assert_eq!(random_word(0xdec7541253414c45, (1 << 19) - 1, 2), @@TIER1_RANDOM_SAMPLE2@@);
    assert_scale_corpus_contract(32, SCALE_SEED32, 1, SCALE_FINITE_TRANSITION_LIMIT32 as i64, RANDOM_CASES32, SCALE_MODE_CROSS_GROUPS32, SCALE_TUPLE_HASH32);
    assert_scale_corpus_contract(64, SCALE_SEED64, 1, SCALE_FINITE_TRANSITION_LIMIT64 as i64, RANDOM_CASES64, SCALE_MODE_CROSS_GROUPS64, SCALE_TUPLE_HASH64);
    assert_scale_corpus_contract(128, SCALE_SEED128, 2, SCALE_FINITE_TRANSITION_LIMIT128 as i64, RANDOM_CASES128, SCALE_MODE_CROSS_GROUPS128, SCALE_TUPLE_HASH128);
    assert_corpus_stream_contracts();
    assert_eq!(TOTAL32_COUNT, STRUCTURED32_COUNT + RANDOM32_COUNT);
    assert_eq!(TOTAL64_COUNT, STRUCTURED64_COUNT + RANDOM64_COUNT);
    assert_eq!(TOTAL128_COUNT, STRUCTURED128_COUNT + RANDOM128_COUNT);
}

#[test]
fn tier1_arithmetic_structured_native_differential() {
    let shard = Shard::load();
    let mut count = 0u64;
    for &op in &ROUNDED_OPS {
        visit_pairs32(|x, y| { for &mode in &MODES { if shard.owns(count) { check_rounded32(op, x, y, mode); } count += 1; } });
    }
    for &tc in SEMANTIC_ROUNDED32 { for &mode in &MODES { if shard.owns(count) { check_rounded32(tc.op, tc.x, tc.y, mode); } count += 1; } }
    visit_triples32(|x, y, z| { for &mode in &MODES { if shard.owns(count) { check_fma32(x, y, z, mode); } count += 1; } });
    for &tc in SEMANTIC_FMA32 { for &mode in &MODES { if shard.owns(count) { check_fma32(tc.x, tc.y, tc.z, mode); } count += 1; } }
    for &op in &UNROUNDED_OPS { visit_pairs32(|x, y| { if shard.owns(count) { check_unrounded32(op, x, y); } count += 1; }); }
    for &tc in SEMANTIC_REMAINDER32 {
        for &op in &UNROUNDED_OPS { if shard.owns(count) { check_unrounded32(op, tc.x, tc.y); } count += 1; }
        if shard.count == 1 {
            assert_ne!(Decimal32::from_bits(tc.x).remainder(Decimal32::from_bits(tc.y)).0.to_bits(),
                Decimal32::from_bits(tc.x).fmod(Decimal32::from_bits(tc.y)).0.to_bits(), "Decimal32 remainder/fmod discriminator collapsed");
        }
    }
    for &x in BOUNDARY32 { for &mode in &MODES { if shard.owns(count) { check_sqrt32(x, mode); } count += 1; } }
    for &x in SEMANTIC_SQRT32 { for &mode in &MODES { if shard.owns(count) { check_sqrt32(x, mode); } count += 1; } }
    for &x in BOUNDARY32 { for &exponent in &SCALE_EXPONENTS { for &mode in &MODES { if shard.owns(count) { check_scale32(x, exponent, mode); } count += 1; } } }
    for &tc in SEMANTIC_SCALE32 { for &mode in &MODES { if shard.owns(count) { check_scale32(tc.x, tc.exponent, mode); } count += 1; } }
    assert_eq!(count, STRUCTURED32_COUNT);
    eprintln!("Rust Decimal32 structured Tier 1 exact comparisons: {}/{}", shard.owned_count(count), count);

    count = 0;
    for &op in &ROUNDED_OPS {
        visit_pairs64(|x, y| { for &mode in &MODES { if shard.owns(count) { check_rounded64(op, x, y, mode); } count += 1; } });
    }
    for &tc in SEMANTIC_ROUNDED64 { for &mode in &MODES { if shard.owns(count) { check_rounded64(tc.op, tc.x, tc.y, mode); } count += 1; } }
    visit_triples64(|x, y, z| { for &mode in &MODES { if shard.owns(count) { check_fma64(x, y, z, mode); } count += 1; } });
    for &tc in SEMANTIC_FMA64 { for &mode in &MODES { if shard.owns(count) { check_fma64(tc.x, tc.y, tc.z, mode); } count += 1; } }
    for &op in &UNROUNDED_OPS { visit_pairs64(|x, y| { if shard.owns(count) { check_unrounded64(op, x, y); } count += 1; }); }
    for &tc in SEMANTIC_REMAINDER64 {
        for &op in &UNROUNDED_OPS { if shard.owns(count) { check_unrounded64(op, tc.x, tc.y); } count += 1; }
        if shard.count == 1 {
            assert_ne!(Decimal64::from_bits(tc.x).remainder(Decimal64::from_bits(tc.y)).0.to_bits(),
                Decimal64::from_bits(tc.x).fmod(Decimal64::from_bits(tc.y)).0.to_bits(), "Decimal64 remainder/fmod discriminator collapsed");
        }
    }
    for &x in BOUNDARY64 { for &mode in &MODES { if shard.owns(count) { check_sqrt64(x, mode); } count += 1; } }
    for &x in SEMANTIC_SQRT64 { for &mode in &MODES { if shard.owns(count) { check_sqrt64(x, mode); } count += 1; } }
    for &x in BOUNDARY64 { for &exponent in &SCALE_EXPONENTS { for &mode in &MODES { if shard.owns(count) { check_scale64(x, exponent, mode); } count += 1; } } }
    for &tc in SEMANTIC_SCALE64 { for &mode in &MODES { if shard.owns(count) { check_scale64(tc.x, tc.exponent, mode); } count += 1; } }
    assert_eq!(count, STRUCTURED64_COUNT);
    eprintln!("Rust Decimal64 structured Tier 1 exact comparisons: {}/{}", shard.owned_count(count), count);

    count = 0;
    for &op in &ROUNDED_OPS {
        visit_pairs128(|x, y| { for &mode in &MODES { if shard.owns(count) { check_rounded128(op, x, y, mode); } count += 1; } });
    }
    for &tc in SEMANTIC_ROUNDED128 { for &mode in &MODES { if shard.owns(count) { check_rounded128(tc.op, tc.x, tc.y, mode); } count += 1; } }
    visit_triples128(|x, y, z| { for &mode in &MODES { if shard.owns(count) { check_fma128(x, y, z, mode); } count += 1; } });
    for &tc in SEMANTIC_FMA128 { for &mode in &MODES { if shard.owns(count) { check_fma128(tc.x, tc.y, tc.z, mode); } count += 1; } }
    for &op in &UNROUNDED_OPS { visit_pairs128(|x, y| { if shard.owns(count) { check_unrounded128(op, x, y); } count += 1; }); }
    for &tc in SEMANTIC_REMAINDER128 {
        for &op in &UNROUNDED_OPS { if shard.owns(count) { check_unrounded128(op, tc.x, tc.y); } count += 1; }
        if shard.count == 1 {
            assert_ne!(decimal128_words(decimal128(tc.x).remainder(decimal128(tc.y)).0),
                decimal128_words(decimal128(tc.x).fmod(decimal128(tc.y)).0), "Decimal128 remainder/fmod discriminator collapsed");
        }
    }
    for &x in BOUNDARY128 { for &mode in &MODES { if shard.owns(count) { check_sqrt128(x, mode); } count += 1; } }
    for &x in SEMANTIC_SQRT128 { for &mode in &MODES { if shard.owns(count) { check_sqrt128(x, mode); } count += 1; } }
    for &x in BOUNDARY128 { for &exponent in &SCALE_EXPONENTS { for &mode in &MODES { if shard.owns(count) { check_scale128(x, exponent, mode); } count += 1; } } }
    for &tc in SEMANTIC_SCALE128 { for &mode in &MODES { if shard.owns(count) { check_scale128(tc.x, tc.exponent, mode); } count += 1; } }
    assert_eq!(count, STRUCTURED128_COUNT);
    eprintln!("Rust Decimal128 structured Tier 1 exact comparisons: {}/{}", shard.owned_count(count), count);
}

fn splitmix64(mut value: u64) -> u64 {
    value = value.wrapping_add(0x9e3779b97f4a7c15);
    value = (value ^ (value >> 30)).wrapping_mul(0xbf58476d1ce4e5b9);
    value = (value ^ (value >> 27)).wrapping_mul(0x94d049bb133111eb);
    value ^ (value >> 31)
}

fn random_word(seed: u64, case_index: u64, lane: u64) -> u64 {
    splitmix64(seed ^ case_index.wrapping_mul(0xd1342543de82ef95) ^ lane.wrapping_mul(0x9e3779b97f4a7c15))
}

// mix_operand* fold one drawn operand into a stream digest. The corpus
// stream contract and the deterministic-random differentials share these
// mixers, so the digest always reflects the operands as consumed.
fn mix_operand32(digest: u64, value: u32) -> u64 {
    scale_tuple_hash_mix(scale_tuple_hash_mix(digest, value as u64), 0)
}

fn mix_operand64(digest: u64, value: u64) -> u64 {
    scale_tuple_hash_mix(scale_tuple_hash_mix(digest, value), 0)
}

fn mix_operand128(digest: u64, words: Words) -> u64 {
    scale_tuple_hash_mix(scale_tuple_hash_mix(digest, words.lo), words.hi)
}

// random_operand* derive the deterministic-random operands for one logical
// operand slot. The differential blocks and the corpus stream contract both
// consume these helpers, so a lane or truncation drift cannot reach only one
// of them.
fn random_operand32(seed: u64, case_index: u64, lane: u64) -> u32 {
    random_word(seed, case_index, lane) as u32
}

fn random_operand64(seed: u64, case_index: u64, lane: u64) -> u64 {
    random_word(seed, case_index, lane)
}

fn random_operand128(seed: u64, case_index: u64, lane: u64) -> Words {
    Words {
        lo: random_word(seed, case_index, lane * 2),
        hi: random_word(seed, case_index, lane * 2 + 1),
    }
}

fn scale_mode_cross_groups(transition_limit: i64) -> u64 {
    (transition_limit * 2 + 1) as u64
}

fn scale_mode_cross_case(case_index: u64, transition_limit: i64) -> bool {
    if case_index % SCALE_RANDOM_STRATA != 1 {
        return false;
    }
    let slot = case_index / SCALE_RANDOM_STRATA;
    slot < scale_mode_cross_groups(transition_limit) * SCALE_MODE_CROSS
}

fn scale_mode_cross_group(case_index: u64) -> u64 {
    (case_index / SCALE_RANDOM_STRATA) / SCALE_MODE_CROSS
}

fn random_scale_exponent(seed: u64, case_index: u64, lane: u64, transition_limit: i64) -> i64 {
    match case_index % SCALE_RANDOM_STRATA {
        0 => random_word(seed, case_index, lane) as i64,
        1 => {
            if scale_mode_cross_case(case_index, transition_limit) {
                scale_mode_cross_group(case_index) as i64 - transition_limit
            } else {
                (random_word(seed, case_index, lane) % (transition_limit * 2 + 1) as u64) as i64 - transition_limit
            }
        }
        2 => (random_word(seed, case_index, lane) % (transition_limit * 4 + 1) as u64) as i64 - transition_limit * 2,
        _ => (random_word(seed, case_index, lane) % (transition_limit * 2 + 1) as u64) as i64 - transition_limit,
    }
}

fn random_scale_operand_index(case_index: u64, transition_limit: i64) -> u64 {
    if scale_mode_cross_case(case_index, transition_limit) {
        scale_mode_cross_group(case_index)
    } else {
        case_index
    }
}

fn scale_mode_cross_sign(group: u64) -> u64 {
    group & 1
}

fn random_scale_operand32(seed: u64, case_index: u64, transition_limit: i64) -> u32 {
    if scale_mode_cross_case(case_index, transition_limit) {
        let group = scale_mode_cross_group(case_index);
        let exponent = group as i64 - transition_limit;
        let sign = (scale_mode_cross_sign(group) << 31) as u32;
        return if exponent > 0 {
            sign | 0x00000001
        } else if exponent < 0 {
            sign | 0x77f8967f
        } else {
            sign | 0x32800001
        };
    }
    let mut raw = random_word(seed, random_scale_operand_index(case_index, transition_limit), 0) as u32;
    if case_index % SCALE_RANDOM_STRATA != 0 {
        raw = raw & !0x20000000 | 1;
    }
    raw
}

fn random_scale_operand64(seed: u64, case_index: u64, transition_limit: i64) -> u64 {
    if scale_mode_cross_case(case_index, transition_limit) {
        let group = scale_mode_cross_group(case_index);
        let exponent = group as i64 - transition_limit;
        let sign = scale_mode_cross_sign(group) << 63;
        return if exponent > 0 {
            sign | 0x0000000000000001
        } else if exponent < 0 {
            sign | 0x77fb86f26fc0ffff
        } else {
            sign | 0x31c0000000000001
        };
    }
    let mut raw = random_word(seed, random_scale_operand_index(case_index, transition_limit), 0);
    if case_index % SCALE_RANDOM_STRATA != 0 {
        raw = raw & !0x2000000000000000 | 1;
    }
    raw
}

fn random_scale_operand128(seed: u64, case_index: u64, transition_limit: i64) -> Words {
    if scale_mode_cross_case(case_index, transition_limit) {
        let group = scale_mode_cross_group(case_index);
        let exponent = group as i64 - transition_limit;
        let sign = scale_mode_cross_sign(group) << 63;
        return if exponent > 0 {
            Words { lo: 0x0000000000000001, hi: sign }
        } else if exponent < 0 {
            Words { lo: 0x378d8e63ffffffff, hi: sign | 0x5fffed09bead87c0 }
        } else {
            Words { lo: 0x0000000000000001, hi: sign | 0x3040000000000000 }
        };
    }
    let operand_index = random_scale_operand_index(case_index, transition_limit);
    let mut raw = Words {
        lo: random_word(seed, operand_index, 0),
        hi: random_word(seed, operand_index, 1),
    };
    if case_index % SCALE_RANDOM_STRATA != 0 {
        raw.lo |= 1;
        raw.hi &= !0x2001000000000000;
    }
    raw
}

fn scale_operand32_is_canonical_nonzero_finite(raw: u32) -> bool {
    if raw & 0x60000000 == 0x60000000 {
        if raw & 0x78000000 == 0x78000000 {
            return false;
        }
        let coefficient = raw & 0x001fffff | 0x00800000;
        return coefficient < 10000000;
    }
    raw & 0x007fffff != 0
}

fn scale_operand64_is_canonical_nonzero_finite(raw: u64) -> bool {
    if raw & 0x6000000000000000 == 0x6000000000000000 {
        if raw & 0x7800000000000000 == 0x7800000000000000 {
            return false;
        }
        let coefficient = raw & 0x0007ffffffffffff | 0x0020000000000000;
        return coefficient < 10000000000000000;
    }
    raw & 0x001fffffffffffff != 0
}

fn scale_operand128_is_canonical_nonzero_finite(raw: Words) -> bool {
    if raw.hi & 0x7800000000000000 >= 0x6000000000000000 {
        return false;
    }
    let coefficient_hi = raw.hi & 0x0001ffffffffffff;
    if coefficient_hi == 0 && raw.lo == 0 {
        return false;
    }
    coefficient_hi < 0x0001ed09bead87c0
        || (coefficient_hi == 0x0001ed09bead87c0 && raw.lo <= 0x378d8e63ffffffff)
}

fn scale_operand(bits: u32, seed: u64, case_index: u64, transition_limit: i64) -> (u64, u64, bool) {
    match bits {
        32 => {
            let raw = random_scale_operand32(seed, case_index, transition_limit);
            (raw as u64, 0, scale_operand32_is_canonical_nonzero_finite(raw))
        }
        64 => {
            let raw = random_scale_operand64(seed, case_index, transition_limit);
            (raw, 0, scale_operand64_is_canonical_nonzero_finite(raw))
        }
        128 => {
            let raw = random_scale_operand128(seed, case_index, transition_limit);
            (raw.lo, raw.hi, scale_operand128_is_canonical_nonzero_finite(raw))
        }
        _ => panic!("unsupported Tier 1 ScaleB corpus width {bits}"),
    }
}

fn scale_tuple_hash_mix(digest: u64, word: u64) -> u64 {
    (digest ^ word).wrapping_mul(SCALE_TUPLE_HASH_PRIME)
}

fn assert_scale_corpus_contract(bits: u32, seed: u64, lane: u64, transition_limit: i64, cases: u64, want_groups: u64, want_hash: u64) {
    assert_eq!(scale_mode_cross_groups(transition_limit), want_groups, "Decimal{bits} ScaleB mode-cross group count drift");
    let mut seen = vec![false; (want_groups * SCALE_MODE_CROSS) as usize];
    let mut mode_sets = vec![0u8; want_groups as usize];
    let mut operand_lo = vec![0u64; want_groups as usize];
    let mut operand_hi = vec![0u64; want_groups as usize];
    let mut operand_seen = vec![false; want_groups as usize];
    let mut digest = scale_tuple_hash_mix(SCALE_TUPLE_HASH_OFFSET, bits as u64);
    let mut mode_cross_cases = 0u64;
    for i in 0..cases {
        let (lo, hi, finite) = scale_operand(bits, seed, i, transition_limit);
        let exponent = random_scale_exponent(seed, i, lane, transition_limit);
        let mode = MODES[i as usize % MODES.len()];
        digest = scale_tuple_hash_mix(digest, i);
        digest = scale_tuple_hash_mix(digest, lo);
        digest = scale_tuple_hash_mix(digest, hi);
        digest = scale_tuple_hash_mix(digest, exponent as u64);
        digest = scale_tuple_hash_mix(digest, mode.native as u64);

        match i % SCALE_RANDOM_STRATA {
            0 => assert_eq!(exponent, random_word(seed, i, lane) as i64, "Decimal{bits} ScaleB full-domain exponent drift at case {i}"),
            1 => {
                if !scale_mode_cross_case(i, transition_limit) {
                    assert!((-transition_limit..=transition_limit).contains(&exponent) && finite,
                        "Decimal{bits} ScaleB surplus in-range case {i} exponent={exponent} finite={finite}");
                    continue;
                }
                mode_cross_cases += 1;
                let group = scale_mode_cross_group(i);
                let group_index = group as usize;
                assert_eq!(exponent, group as i64 - transition_limit, "Decimal{bits} ScaleB mode-cross exponent drift at case {i}");
                assert_eq!(random_scale_operand_index(i, transition_limit), group, "Decimal{bits} ScaleB mode-cross operand index drift at case {i}");
                assert!(finite, "Decimal{bits} ScaleB mode-cross operand is not canonical nonzero finite at case {i}: hi={hi:016x} lo={lo:016x}");
                if operand_seen[group_index] {
                    assert_eq!((lo, hi), (operand_lo[group_index], operand_hi[group_index]), "Decimal{bits} ScaleB mode-cross operand drift in group {group}");
                } else {
                    operand_lo[group_index] = lo;
                    operand_hi[group_index] = hi;
                    operand_seen[group_index] = true;
                }
                let mode_index = (i % SCALE_MODE_CROSS) as usize;
                mode_sets[group_index] |= 1u8 << mode_index;
                seen[group_index * SCALE_MODE_CROSS as usize + mode_index] = true;
            }
            2 => assert!((-transition_limit * 2..=transition_limit * 2).contains(&exponent) && finite,
                "Decimal{bits} ScaleB transition-window case {i} exponent={exponent} finite={finite}"),
            3 => assert!((-transition_limit..=transition_limit).contains(&exponent) && finite,
                "Decimal{bits} ScaleB in-range case {i} exponent={exponent} finite={finite}"),
            stratum => panic!("unexpected ScaleB stratum {stratum}"),
        }
    }
    digest = scale_tuple_hash_mix(digest, cases);
    assert_eq!(digest, want_hash, "Decimal{bits} ScaleB tuple stream hash drift");
    assert_eq!(mode_cross_cases, want_groups * SCALE_MODE_CROSS, "Decimal{bits} ScaleB complete mode-cross case count drift");
    let missing = seen.iter().filter(|&&covered| !covered).count();
    assert_eq!(missing, 0, "Decimal{bits} ScaleB random corpus misses {missing}/{} finite-transition exponent×rounding-mode cells", seen.len());
    for (group, &modes) in mode_sets.iter().enumerate() {
        assert_eq!(modes, 0x1f, "Decimal{bits} ScaleB group {group} native-mode index set drift");
    }
}

#[test]
fn tier1_arithmetic_deterministic_random_native_differential() {
    let shard = Shard::load();
    let mut count = 0u64;
    let mut consumed = 0u64;
    let mut digest = scale_tuple_hash_mix(SCALE_TUPLE_HASH_OFFSET, 32);
    for (op_index, &op) in ROUNDED_OPS.iter().enumerate() {
        let seed = RANDOM_ROUNDED_SEED_BASE32 ^ op_index as u64;
        for i in 0..RANDOM_CASES32 {
            let x = random_operand32(seed, i, 0);
            let y = random_operand32(seed, i, 1);
            let mode = MODES[i as usize % MODES.len()];
            digest = mix_operand32(digest, x);
            digest = mix_operand32(digest, y);
            digest = scale_tuple_hash_mix(digest, mode.native as u64);
            consumed += 1;
            if shard.owns(count) { check_rounded32(op, x, y, mode); }
            count += 1;
        }
    }
    for (op_index, &op) in UNROUNDED_OPS.iter().enumerate() {
        let seed = RANDOM_UNROUNDED_SEED_BASE32 ^ op_index as u64;
        for i in 0..RANDOM_CASES32 {
            let x = random_operand32(seed, i, 0);
            let y = random_operand32(seed, i, 1);
            digest = mix_operand32(digest, x);
            digest = mix_operand32(digest, y);
            consumed += 1;
            if shard.owns(count) { check_unrounded32(op, x, y); }
            count += 1;
        }
    }
    let seed = SCALE_SEED32;
    for i in 0..RANDOM_CASES32 { if shard.owns(count) { check_scale32(random_scale_operand32(seed, i, SCALE_FINITE_TRANSITION_LIMIT32 as i64), random_scale_exponent(seed, i, 1, SCALE_FINITE_TRANSITION_LIMIT32 as i64), MODES[i as usize % MODES.len()]); } count += 1; }
    for i in 0..RANDOM_CASES32 {
        let x = random_operand32(RANDOM_FMA_SEED32, i, 0);
        let y = random_operand32(RANDOM_FMA_SEED32, i, 1);
        let z = random_operand32(RANDOM_FMA_SEED32, i, 2);
        let mode = MODES[i as usize % MODES.len()];
        digest = mix_operand32(digest, x);
        digest = mix_operand32(digest, y);
        digest = mix_operand32(digest, z);
        digest = scale_tuple_hash_mix(digest, mode.native as u64);
        consumed += 1;
        if shard.owns(count) { check_fma32(x, y, z, mode); }
        count += 1;
    }
    for i in 0..RANDOM_CASES32 {
        let x = random_operand32(RANDOM_SQRT_SEED32, i, 0);
        let mode = MODES[i as usize % MODES.len()];
        digest = mix_operand32(digest, x);
        digest = scale_tuple_hash_mix(digest, mode.native as u64);
        consumed += 1;
        if shard.owns(count) { check_sqrt32(x, mode); }
        count += 1;
    }
    assert_eq!(scale_tuple_hash_mix(digest, consumed), RANDOM_STREAM_HASH32, "Decimal32 random differential consumed-stream hash drift");
    assert_eq!(count, RANDOM32_COUNT);
    eprintln!("Rust Decimal32 random Tier 1 exact comparisons: {}/{}", shard.owned_count(count), count);

    count = 0;
    let mut consumed = 0u64;
    let mut digest = scale_tuple_hash_mix(SCALE_TUPLE_HASH_OFFSET, 64);
    for (op_index, &op) in ROUNDED_OPS.iter().enumerate() {
        let seed = RANDOM_ROUNDED_SEED_BASE64 ^ op_index as u64;
        for i in 0..RANDOM_CASES64 {
            let x = random_operand64(seed, i, 0);
            let y = random_operand64(seed, i, 1);
            let mode = MODES[i as usize % MODES.len()];
            digest = mix_operand64(digest, x);
            digest = mix_operand64(digest, y);
            digest = scale_tuple_hash_mix(digest, mode.native as u64);
            consumed += 1;
            if shard.owns(count) { check_rounded64(op, x, y, mode); }
            count += 1;
        }
    }
    for (op_index, &op) in UNROUNDED_OPS.iter().enumerate() {
        let seed = RANDOM_UNROUNDED_SEED_BASE64 ^ op_index as u64;
        for i in 0..RANDOM_CASES64 {
            let x = random_operand64(seed, i, 0);
            let y = random_operand64(seed, i, 1);
            digest = mix_operand64(digest, x);
            digest = mix_operand64(digest, y);
            consumed += 1;
            if shard.owns(count) { check_unrounded64(op, x, y); }
            count += 1;
        }
    }
    let seed = SCALE_SEED64;
    for i in 0..RANDOM_CASES64 { if shard.owns(count) { check_scale64(random_scale_operand64(seed, i, SCALE_FINITE_TRANSITION_LIMIT64 as i64), random_scale_exponent(seed, i, 1, SCALE_FINITE_TRANSITION_LIMIT64 as i64), MODES[i as usize % MODES.len()]); } count += 1; }
    for i in 0..RANDOM_CASES64 {
        let x = random_operand64(RANDOM_FMA_SEED64, i, 0);
        let y = random_operand64(RANDOM_FMA_SEED64, i, 1);
        let z = random_operand64(RANDOM_FMA_SEED64, i, 2);
        let mode = MODES[i as usize % MODES.len()];
        digest = mix_operand64(digest, x);
        digest = mix_operand64(digest, y);
        digest = mix_operand64(digest, z);
        digest = scale_tuple_hash_mix(digest, mode.native as u64);
        consumed += 1;
        if shard.owns(count) { check_fma64(x, y, z, mode); }
        count += 1;
    }
    for i in 0..RANDOM_CASES64 {
        let x = random_operand64(RANDOM_SQRT_SEED64, i, 0);
        let mode = MODES[i as usize % MODES.len()];
        digest = mix_operand64(digest, x);
        digest = scale_tuple_hash_mix(digest, mode.native as u64);
        consumed += 1;
        if shard.owns(count) { check_sqrt64(x, mode); }
        count += 1;
    }
    assert_eq!(scale_tuple_hash_mix(digest, consumed), RANDOM_STREAM_HASH64, "Decimal64 random differential consumed-stream hash drift");
    assert_eq!(count, RANDOM64_COUNT);
    eprintln!("Rust Decimal64 random Tier 1 exact comparisons: {}/{}", shard.owned_count(count), count);

    count = 0;
    let mut consumed = 0u64;
    let mut digest = scale_tuple_hash_mix(SCALE_TUPLE_HASH_OFFSET, 128);
    for (op_index, &op) in ROUNDED_OPS.iter().enumerate() {
        let seed = RANDOM_ROUNDED_SEED_BASE128 ^ op_index as u64;
        for i in 0..RANDOM_CASES128 {
            let x = random_operand128(seed, i, 0);
            let y = random_operand128(seed, i, 1);
            let mode = MODES[i as usize % MODES.len()];
            digest = mix_operand128(digest, x);
            digest = mix_operand128(digest, y);
            digest = scale_tuple_hash_mix(digest, mode.native as u64);
            consumed += 1;
            if shard.owns(count) { check_rounded128(op, x, y, mode); }
            count += 1;
        }
    }
    for (op_index, &op) in UNROUNDED_OPS.iter().enumerate() {
        let seed = RANDOM_UNROUNDED_SEED_BASE128 ^ op_index as u64;
        for i in 0..RANDOM_CASES128 {
            let x = random_operand128(seed, i, 0);
            let y = random_operand128(seed, i, 1);
            digest = mix_operand128(digest, x);
            digest = mix_operand128(digest, y);
            consumed += 1;
            if shard.owns(count) { check_unrounded128(op, x, y); }
            count += 1;
        }
    }
    let seed = SCALE_SEED128;
    for i in 0..RANDOM_CASES128 {
        if shard.owns(count) { check_scale128(random_scale_operand128(seed, i, SCALE_FINITE_TRANSITION_LIMIT128 as i64), random_scale_exponent(seed, i, 2, SCALE_FINITE_TRANSITION_LIMIT128 as i64), MODES[i as usize % MODES.len()]); }
        count += 1;
    }
    for i in 0..RANDOM_CASES128 {
        let x = random_operand128(RANDOM_FMA_SEED128, i, 0);
        let y = random_operand128(RANDOM_FMA_SEED128, i, 1);
        let z = random_operand128(RANDOM_FMA_SEED128, i, 2);
        let mode = MODES[i as usize % MODES.len()];
        digest = mix_operand128(digest, x);
        digest = mix_operand128(digest, y);
        digest = mix_operand128(digest, z);
        digest = scale_tuple_hash_mix(digest, mode.native as u64);
        consumed += 1;
        if shard.owns(count) { check_fma128(x, y, z, mode); }
        count += 1;
    }
    for i in 0..RANDOM_CASES128 {
        let x = random_operand128(RANDOM_SQRT_SEED128, i, 0);
        let mode = MODES[i as usize % MODES.len()];
        digest = mix_operand128(digest, x);
        digest = scale_tuple_hash_mix(digest, mode.native as u64);
        consumed += 1;
        if shard.owns(count) { check_sqrt128(x, mode); }
        count += 1;
    }
    assert_eq!(scale_tuple_hash_mix(digest, consumed), RANDOM_STREAM_HASH128, "Decimal128 random differential consumed-stream hash drift");
    assert_eq!(count, RANDOM128_COUNT);
    eprintln!("Rust Decimal128 random Tier 1 exact comparisons: {}/{}", shard.owned_count(count), count);
}

// Routing sentinels: generator-selected known-answer rows that bind the
// runner glue (operand slots, rounding-mode wiring, dispatch-row labels) to
// values pinned outside the runtime. Each row's expected (bits, flags) was
// computed at generation time through the public bid754-go API and is
// byte-equal pinned in devtools/verification_sentinels.json; at runtime the
// Intel C leg, the generated Rust public leg, and (where exported) the
// flagless leg must reproduce the pinned answer exactly. A glue bug that
// skews every leg the same way — invisible to the differential — diverges
// from the pin here. The diverged set names the broken leg axis:
// {C,public} means shared runner glue (slot/mode/dispatch), {public} a
// generated-Rust regression, {C} an FFI/link regression.

fn sentinel_mode(row: &str, native: u32) -> Mode {
    for mode in MODES {
        if mode.native == native {
            return mode;
        }
    }
    panic!("routing sentinel row [{row}]: native mode {native} is not in the runner mode table");
}

fn sentinel_mode_label(mode: Mode) -> String {
    format!("{}(native {})", mode.name, mode.native)
}

fn sentinel_hex32(row: &str, text: &str) -> u32 {
    assert!(text.len() == 8, "routing sentinel row [{row}]: expected 8 hex digits, got {text:?}");
    u32::from_str_radix(text, 16)
        .unwrap_or_else(|err| panic!("routing sentinel row [{row}]: bad 32-bit hex {text:?}: {err}"))
}

fn sentinel_hex64(row: &str, text: &str) -> u64 {
    assert!(text.len() == 16, "routing sentinel row [{row}]: expected 16 hex digits, got {text:?}");
    u64::from_str_radix(text, 16)
        .unwrap_or_else(|err| panic!("routing sentinel row [{row}]: bad 64-bit hex {text:?}: {err}"))
}

fn sentinel_words128(row: &str, text: &str) -> Words {
    assert!(
        text.len() == 33 && text.as_bytes()[16] == b':',
        "routing sentinel row [{row}]: expected <hi16>:<lo16> hex words, got {text:?}"
    );
    Words { hi: sentinel_hex64(row, &text[..16]), lo: sentinel_hex64(row, &text[17..]) }
}

fn sentinel_result32(bits: u32, flags: u32) -> String {
    format!("{bits:08x}/{flags:08x}")
}

fn sentinel_result64(bits: u64, flags: u32) -> String {
    format!("{bits:016x}/{flags:08x}")
}

fn sentinel_result128(words: Words, flags: u32) -> String {
    format!("{:016x}:{:016x}/{flags:08x}", words.hi, words.lo)
}

fn sentinel_assert(row: &str, mode_label: &str, pinned: &str, native: &str, public: &str, flagless: Option<&str>) {
    let pinned_bits = pinned.split('/').next().unwrap_or("");
    let flagless_matches = flagless.map_or(true, |bits| bits == pinned_bits);
    if native == pinned && public == pinned && flagless_matches {
        return;
    }
    let mut diverged = Vec::new();
    if native != pinned {
        diverged.push("C");
    }
    if public != pinned {
        diverged.push("public");
    }
    if !flagless_matches {
        diverged.push("flagless");
    }
    panic!(
        "routing sentinel mismatch [{row}]:\n  pinned={pinned} C={native} public={public} flagless={}\n  mode={mode_label} diverged={{{}}}",
        flagless.unwrap_or("-"),
        diverged.join(",")
    );
}

fn sentinel_rounded32(row: &str, op: RoundedOp, x: u32, y: u32, mode: Mode, pinned: &str) {
    let (native, native_flags, public, public_flags, flagless) = legs_rounded32(op, x, y, mode);
    let flagless_text = flagless.map(|bits| format!("{bits:08x}"));
    sentinel_assert(
        row,
        &sentinel_mode_label(mode),
        pinned,
        &sentinel_result32(native, native_flags),
        &sentinel_result32(public, public_flags),
        flagless_text.as_deref(),
    );
}

fn sentinel_rounded64(row: &str, op: RoundedOp, x: u64, y: u64, mode: Mode, pinned: &str) {
    let (native, native_flags, public, public_flags, flagless) = legs_rounded64(op, x, y, mode);
    let flagless_text = flagless.map(|bits| format!("{bits:016x}"));
    sentinel_assert(
        row,
        &sentinel_mode_label(mode),
        pinned,
        &sentinel_result64(native, native_flags),
        &sentinel_result64(public, public_flags),
        flagless_text.as_deref(),
    );
}

fn sentinel_rounded128(row: &str, op: RoundedOp, x: Words, y: Words, mode: Mode, pinned: &str) {
    let (native, native_flags, public, public_flags) = legs_rounded128(op, x, y, mode);
    sentinel_assert(
        row,
        &sentinel_mode_label(mode),
        pinned,
        &sentinel_result128(native, native_flags),
        &sentinel_result128(public, public_flags),
        None,
    );
}

fn sentinel_unrounded32(row: &str, op: UnroundedOp, x: u32, y: u32, pinned: &str) {
    let (native, native_flags, public, public_flags) = legs_unrounded32(op, x, y);
    sentinel_assert(
        row,
        "(none)",
        pinned,
        &sentinel_result32(native, native_flags),
        &sentinel_result32(public, public_flags),
        None,
    );
}

fn sentinel_unrounded64(row: &str, op: UnroundedOp, x: u64, y: u64, pinned: &str) {
    let (native, native_flags, public, public_flags) = legs_unrounded64(op, x, y);
    sentinel_assert(
        row,
        "(none)",
        pinned,
        &sentinel_result64(native, native_flags),
        &sentinel_result64(public, public_flags),
        None,
    );
}

fn sentinel_unrounded128(row: &str, op: UnroundedOp, x: Words, y: Words, pinned: &str) {
    let (native, native_flags, public, public_flags) = legs_unrounded128(op, x, y);
    sentinel_assert(
        row,
        "(none)",
        pinned,
        &sentinel_result128(native, native_flags),
        &sentinel_result128(public, public_flags),
        None,
    );
}

fn sentinel_fma32(row: &str, x: u32, y: u32, z: u32, mode: Mode, pinned: &str) {
    let (native, native_flags, public, public_flags) = legs_fma32(x, y, z, mode);
    sentinel_assert(
        row,
        &sentinel_mode_label(mode),
        pinned,
        &sentinel_result32(native, native_flags),
        &sentinel_result32(public, public_flags),
        None,
    );
}

fn sentinel_fma64(row: &str, x: u64, y: u64, z: u64, mode: Mode, pinned: &str) {
    let (native, native_flags, public, public_flags) = legs_fma64(x, y, z, mode);
    sentinel_assert(
        row,
        &sentinel_mode_label(mode),
        pinned,
        &sentinel_result64(native, native_flags),
        &sentinel_result64(public, public_flags),
        None,
    );
}

fn sentinel_fma128(row: &str, x: Words, y: Words, z: Words, mode: Mode, pinned: &str) {
    let (native, native_flags, public, public_flags) = legs_fma128(x, y, z, mode);
    sentinel_assert(
        row,
        &sentinel_mode_label(mode),
        pinned,
        &sentinel_result128(native, native_flags),
        &sentinel_result128(public, public_flags),
        None,
    );
}

fn sentinel_sqrt32(row: &str, x: u32, mode: Mode, pinned: &str) {
    let (native, native_flags, public, public_flags) = legs_sqrt32(x, mode);
    sentinel_assert(
        row,
        &sentinel_mode_label(mode),
        pinned,
        &sentinel_result32(native, native_flags),
        &sentinel_result32(public, public_flags),
        None,
    );
}

fn sentinel_sqrt64(row: &str, x: u64, mode: Mode, pinned: &str) {
    let (native, native_flags, public, public_flags) = legs_sqrt64(x, mode);
    sentinel_assert(
        row,
        &sentinel_mode_label(mode),
        pinned,
        &sentinel_result64(native, native_flags),
        &sentinel_result64(public, public_flags),
        None,
    );
}

fn sentinel_sqrt128(row: &str, x: Words, mode: Mode, pinned: &str) {
    let (native, native_flags, public, public_flags) = legs_sqrt128(x, mode);
    sentinel_assert(
        row,
        &sentinel_mode_label(mode),
        pinned,
        &sentinel_result128(native, native_flags),
        &sentinel_result128(public, public_flags),
        None,
    );
}

fn sentinel_scale32(row: &str, x: u32, exponent: i64, mode: Mode, pinned: &str) {
    let (native, native_flags, public, public_flags) = legs_scale32(x, exponent, mode);
    sentinel_assert(
        row,
        &sentinel_mode_label(mode),
        pinned,
        &sentinel_result32(native, native_flags),
        &sentinel_result32(public, public_flags),
        None,
    );
}

fn sentinel_scale64(row: &str, x: u64, exponent: i64, mode: Mode, pinned: &str) {
    let (native, native_flags, public, public_flags) = legs_scale64(x, exponent, mode);
    sentinel_assert(
        row,
        &sentinel_mode_label(mode),
        pinned,
        &sentinel_result64(native, native_flags),
        &sentinel_result64(public, public_flags),
        None,
    );
}

fn sentinel_scale128(row: &str, x: Words, exponent: i64, mode: Mode, pinned: &str) {
    let (native, native_flags, public, public_flags) = legs_scale128(x, exponent, mode);
    sentinel_assert(
        row,
        &sentinel_mode_label(mode),
        pinned,
        &sentinel_result128(native, native_flags),
        &sentinel_result128(public, public_flags),
        None,
    );
}

fn sentinel_rounded_op(row: &str, name: &str) -> RoundedOp {
    match name {
        "add" => RoundedOp::Add,
        "sub" => RoundedOp::Sub,
        "mul" => RoundedOp::Mul,
        "div" => RoundedOp::Div,
        "quantize" => RoundedOp::Quantize,
        _ => panic!("routing sentinel row [{row}]: unknown rounded operation {name:?}"),
    }
}

fn check_routing_sentinel_row(row: &str) {
    let fields: Vec<&str> = row.split(' ').collect();
    assert!(
        fields.len() >= 5 && fields[fields.len() - 2] == "->",
        "routing sentinel row [{row}]: malformed layout"
    );
    let width = fields[0];
    let operation = fields[1];
    let pinned = fields[fields.len() - 1];
    let mut x_text: Option<&str> = None;
    let mut y_text: Option<&str> = None;
    let mut z_text: Option<&str> = None;
    let mut exponent: Option<i64> = None;
    let mut mode: Option<Mode> = None;
    for field in &fields[2..fields.len() - 2] {
        if let Some(text) = field.strip_prefix("x=") {
            x_text = Some(text);
        } else if let Some(text) = field.strip_prefix("y=") {
            y_text = Some(text);
        } else if let Some(text) = field.strip_prefix("z=") {
            z_text = Some(text);
        } else if let Some(text) = field.strip_prefix("n=") {
            exponent = Some(text.parse::<i64>().unwrap_or_else(|err| {
                panic!("routing sentinel row [{row}]: bad scaleb exponent {text:?}: {err}")
            }));
        } else if let Some(text) = field.strip_prefix("m=") {
            let native = text.parse::<u32>().unwrap_or_else(|err| {
                panic!("routing sentinel row [{row}]: bad native mode {text:?}: {err}")
            });
            mode = Some(sentinel_mode(row, native));
        } else {
            panic!("routing sentinel row [{row}]: unknown field {field:?}");
        }
    }
    let require_shape = |need_x: bool, need_y: bool, need_z: bool, need_n: bool, need_mode: bool| {
        assert!(
            x_text.is_some() == need_x
                && y_text.is_some() == need_y
                && z_text.is_some() == need_z
                && exponent.is_some() == need_n
                && mode.is_some() == need_mode,
            "routing sentinel row [{row}]: field shape does not match operation {operation:?}"
        );
    };
    match operation {
        "add" | "sub" | "mul" | "div" | "quantize" => {
            require_shape(true, true, false, false, true);
            let op = sentinel_rounded_op(row, operation);
            let (x, y, mode) = (x_text.unwrap(), y_text.unwrap(), mode.unwrap());
            match width {
                "d32" => sentinel_rounded32(row, op, sentinel_hex32(row, x), sentinel_hex32(row, y), mode, pinned),
                "d64" => sentinel_rounded64(row, op, sentinel_hex64(row, x), sentinel_hex64(row, y), mode, pinned),
                "d128" => sentinel_rounded128(row, op, sentinel_words128(row, x), sentinel_words128(row, y), mode, pinned),
                _ => panic!("routing sentinel row [{row}]: unknown width {width:?}"),
            }
        }
        "remainder" | "fmod" => {
            require_shape(true, true, false, false, false);
            let op = if operation == "remainder" { UnroundedOp::Remainder } else { UnroundedOp::Fmod };
            let (x, y) = (x_text.unwrap(), y_text.unwrap());
            match width {
                "d32" => sentinel_unrounded32(row, op, sentinel_hex32(row, x), sentinel_hex32(row, y), pinned),
                "d64" => sentinel_unrounded64(row, op, sentinel_hex64(row, x), sentinel_hex64(row, y), pinned),
                "d128" => sentinel_unrounded128(row, op, sentinel_words128(row, x), sentinel_words128(row, y), pinned),
                _ => panic!("routing sentinel row [{row}]: unknown width {width:?}"),
            }
        }
        "fma" => {
            require_shape(true, true, true, false, true);
            let (x, y, z, mode) = (x_text.unwrap(), y_text.unwrap(), z_text.unwrap(), mode.unwrap());
            match width {
                "d32" => sentinel_fma32(row, sentinel_hex32(row, x), sentinel_hex32(row, y), sentinel_hex32(row, z), mode, pinned),
                "d64" => sentinel_fma64(row, sentinel_hex64(row, x), sentinel_hex64(row, y), sentinel_hex64(row, z), mode, pinned),
                "d128" => sentinel_fma128(row, sentinel_words128(row, x), sentinel_words128(row, y), sentinel_words128(row, z), mode, pinned),
                _ => panic!("routing sentinel row [{row}]: unknown width {width:?}"),
            }
        }
        "sqrt" => {
            require_shape(true, false, false, false, true);
            let (x, mode) = (x_text.unwrap(), mode.unwrap());
            match width {
                "d32" => sentinel_sqrt32(row, sentinel_hex32(row, x), mode, pinned),
                "d64" => sentinel_sqrt64(row, sentinel_hex64(row, x), mode, pinned),
                "d128" => sentinel_sqrt128(row, sentinel_words128(row, x), mode, pinned),
                _ => panic!("routing sentinel row [{row}]: unknown width {width:?}"),
            }
        }
        "scaleb" => {
            require_shape(true, false, false, true, true);
            let (x, n, mode) = (x_text.unwrap(), exponent.unwrap(), mode.unwrap());
            match width {
                "d32" => sentinel_scale32(row, sentinel_hex32(row, x), n, mode, pinned),
                "d64" => sentinel_scale64(row, sentinel_hex64(row, x), n, mode, pinned),
                "d128" => sentinel_scale128(row, sentinel_words128(row, x), n, mode, pinned),
                _ => panic!("routing sentinel row [{row}]: unknown width {width:?}"),
            }
        }
        _ => panic!("routing sentinel row [{row}]: unknown operation {operation:?}"),
    }
}

// Runs every pinned sentinel row on every leg. Deliberately ignores the
// shard environment: the rows are few and every shard configuration (and the
// -full gate) must execute all of them.
#[test]
fn tier1_arithmetic_routing_sentinels() {
    assert!(
        !ROUTING_SENTINEL_ROWS.is_empty(),
        "generated Tier 1 arithmetic routing sentinel row set is empty"
    );
    for row in ROUTING_SENTINEL_ROWS {
        check_routing_sentinel_row(row);
    }
    let n = ROUTING_SENTINEL_ROWS.len();
    println!("Rust Tier 1 arithmetic routing sentinels: {n}/{n}");
}

// The canonical sentinel row set. The identical byte sequence is pinned by
// hand in devtools/verification_sentinels.json and emitted into the generated
// Go runner; TestVerificationAnchorsMatchGeneratedArtifacts requires the
// three copies to match exactly.
const ROUTING_SENTINEL_ROWS: [&str; @@TIER1_ARITH_SENTINEL_COUNT@@] = [
@@TIER1_ARITH_SENTINEL_ROWS@@
];

const PROBES32: [u32; 12] = [
@@TIER1_PROBES32_VALUES@@
];

const PROBES64: [u64; 12] = [
@@TIER1_PROBES64_VALUES@@
];

const PROBES128: [Words; 12] = [
@@TIER1_PROBES128_VALUES@@
];

const SEMANTIC_ROUNDED32: &[Rounded32] = &[
@@TIER1_SEMANTIC_ROUNDED32_VALUES@@
];
const SEMANTIC_ROUNDED64: &[Rounded64] = &[
@@TIER1_SEMANTIC_ROUNDED64_VALUES@@
];
const SEMANTIC_ROUNDED128: &[Rounded128] = &[
@@TIER1_SEMANTIC_ROUNDED128_VALUES@@
];
const SEMANTIC_SCALE32: &[Scale32] = &[
@@TIER1_SEMANTIC_SCALE32_VALUES@@
];
const SEMANTIC_SCALE64: &[Scale64] = &[
@@TIER1_SEMANTIC_SCALE64_VALUES@@
];
const SEMANTIC_SCALE128: &[Scale128] = &[
@@TIER1_SEMANTIC_SCALE128_VALUES@@
];
const SEMANTIC_FMA32: &[Triple32] = &[
@@TIER1_SEMANTIC_FMA32_VALUES@@
];
const SEMANTIC_FMA64: &[Triple64] = &[
@@TIER1_SEMANTIC_FMA64_VALUES@@
];
const SEMANTIC_FMA128: &[Triple128] = &[
@@TIER1_SEMANTIC_FMA128_VALUES@@
];
const SEMANTIC_SQRT32: &[u32] = &[
@@TIER1_SEMANTIC_SQRT32_VALUES@@
];
const SEMANTIC_SQRT64: &[u64] = &[
@@TIER1_SEMANTIC_SQRT64_VALUES@@
];
const SEMANTIC_SQRT128: &[Words] = &[
@@TIER1_SEMANTIC_SQRT128_VALUES@@
];
const SEMANTIC_REMAINDER32: &[Pair32] = &[
@@TIER1_SEMANTIC_REMAINDER32_VALUES@@
];
const SEMANTIC_REMAINDER64: &[Pair64] = &[
@@TIER1_SEMANTIC_REMAINDER64_VALUES@@
];
const SEMANTIC_REMAINDER128: &[Pair128] = &[
@@TIER1_SEMANTIC_REMAINDER128_VALUES@@
];
const BOUNDARY32: &[u32] = &[
@@TIER1_BOUNDARY32_VALUES@@
];
const BOUNDARY64: &[u64] = &[
@@TIER1_BOUNDARY64_VALUES@@
];
const BOUNDARY128: &[Words] = &[
@@TIER1_BOUNDARY128_VALUES@@
];
