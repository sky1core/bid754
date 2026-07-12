#![cfg(feature = "tier1-long")]
#![allow(clippy::too_many_arguments)]

use bid754::generated::bid32_exports::{bid32_add, bid32_div, bid32_mul, bid32_sub};
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
const SCALE_TUPLE_HASH_OFFSET: u64 = 14695981039346656037;
const SCALE_TUPLE_HASH_PRIME: u64 = 1099511628211;

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

fn check_rounded32(op: RoundedOp, x: u32, y: u32, mode: Mode) {
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
    assert_eq!((public.to_bits(), public_raw_flags(flags)), (native, native_flags),
        "Decimal32 {op:?} mismatch x={x:08x} y={y:08x} mode={}", mode.name);

    let flagless = match op {
        RoundedOp::Add => Some(bid32_add(x, y, i64::from(mode.native))),
        RoundedOp::Sub => Some(bid32_sub(x, y, i64::from(mode.native))),
        RoundedOp::Mul => Some(bid32_mul(x, y, i64::from(mode.native))),
        RoundedOp::Div => Some(bid32_div(x, y, i64::from(mode.native))),
        RoundedOp::Quantize => None,
    };
    if let Some(flagless) = flagless {
        assert_eq!(flagless, native,
            "Decimal32 flagless {op:?} mismatch x={x:08x} y={y:08x} mode={}", mode.name);
    }
}

fn check_rounded64(op: RoundedOp, x: u64, y: u64, mode: Mode) {
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
    assert_eq!((public.to_bits(), public_raw_flags(flags)), (native, native_flags),
        "Decimal64 {op:?} mismatch x={x:016x} y={y:016x} mode={}", mode.name);
}

fn check_rounded128(op: RoundedOp, x: Words, y: Words, mode: Mode) {
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
    assert_eq!((decimal128_words(public), public_raw_flags(flags)), (c128_words(native), native_flags),
        "Decimal128 {op:?} mismatch x={x:?} y={y:?} mode={}", mode.name);
}

fn check_fma32(x: u32, y: u32, z: u32, mode: Mode) {
    let mut native_flags = 0u32;
    let native = unsafe { libbid_sys::bid32_fma(x, y, z, mode.native, &mut native_flags) };
    let (public, flags) = Decimal32::from_bits(x)
        .fma_with_mode(Decimal32::from_bits(y), Decimal32::from_bits(z), mode.public);
    assert_eq!((public.to_bits(), public_raw_flags(flags)), (native, native_flags),
        "Decimal32 fma mismatch x={x:08x} y={y:08x} z={z:08x} mode={}", mode.name);
}

fn check_fma64(x: u64, y: u64, z: u64, mode: Mode) {
    let mut native_flags = 0u32;
    let native = unsafe { libbid_sys::bid64_fma(x, y, z, mode.native, &mut native_flags) };
    let (public, flags) = Decimal64::from_bits(x)
        .fma_with_mode(Decimal64::from_bits(y), Decimal64::from_bits(z), mode.public);
    assert_eq!((public.to_bits(), public_raw_flags(flags)), (native, native_flags),
        "Decimal64 fma mismatch x={x:016x} y={y:016x} z={z:016x} mode={}", mode.name);
}

fn check_fma128(x: Words, y: Words, z: Words, mode: Mode) {
    let mut native_flags = 0u32;
    let native = unsafe { libbid_sys::bid128_fma(c128(x), c128(y), c128(z), mode.native, &mut native_flags) };
    let (public, flags) = decimal128(x).fma_with_mode(decimal128(y), decimal128(z), mode.public);
    assert_eq!((decimal128_words(public), public_raw_flags(flags)), (c128_words(native), native_flags),
        "Decimal128 fma mismatch x={x:?} y={y:?} z={z:?} mode={}", mode.name);
}

fn check_sqrt32(x: u32, mode: Mode) {
    let mut native_flags = 0u32;
    let native = unsafe { libbid_sys::bid32_sqrt(x, mode.native, &mut native_flags) };
    let (public, flags) = Decimal32::from_bits(x).sqrt_with_mode(mode.public);
    assert_eq!((public.to_bits(), public_raw_flags(flags)), (native, native_flags),
        "Decimal32 sqrt mismatch x={x:08x} mode={}", mode.name);
}

fn check_sqrt64(x: u64, mode: Mode) {
    let mut native_flags = 0u32;
    let native = unsafe { libbid_sys::bid64_sqrt(x, mode.native, &mut native_flags) };
    let (public, flags) = Decimal64::from_bits(x).sqrt_with_mode(mode.public);
    assert_eq!((public.to_bits(), public_raw_flags(flags)), (native, native_flags),
        "Decimal64 sqrt mismatch x={x:016x} mode={}", mode.name);
}

fn check_sqrt128(x: Words, mode: Mode) {
    let mut native_flags = 0u32;
    let native = unsafe { libbid_sys::bid128_sqrt(c128(x), mode.native, &mut native_flags) };
    let (public, flags) = decimal128(x).sqrt_with_mode(mode.public);
    assert_eq!((decimal128_words(public), public_raw_flags(flags)), (c128_words(native), native_flags),
        "Decimal128 sqrt mismatch x={x:?} mode={}", mode.name);
}

fn check_unrounded32(op: UnroundedOp, x: u32, y: u32) {
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
    assert_eq!((public.to_bits(), public_raw_flags(flags)), (native, native_flags),
        "Decimal32 {op:?} mismatch x={x:08x} y={y:08x}");
}

fn check_unrounded64(op: UnroundedOp, x: u64, y: u64) {
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
    assert_eq!((public.to_bits(), public_raw_flags(flags)), (native, native_flags),
        "Decimal64 {op:?} mismatch x={x:016x} y={y:016x}");
}

fn check_unrounded128(op: UnroundedOp, x: Words, y: Words) {
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
    assert_eq!((decimal128_words(public), public_raw_flags(flags)), (c128_words(native), native_flags),
        "Decimal128 {op:?} mismatch x={x:?} y={y:?}");
}

fn check_scale32(x: u32, exponent: i64, mode: Mode) {
    let mut native_flags = 0u32;
    let native = unsafe { native_bid32_scalbln(x, exponent as c_long, mode.native, &mut native_flags) };
    let (public, flags) = Decimal32::from_bits(x).scaleb_with_mode(exponent, mode.public);
    assert_eq!((public.to_bits(), public_raw_flags(flags)), (native, native_flags),
        "Decimal32 scaleB mismatch x={x:08x} exponent={exponent} mode={}", mode.name);
}

fn check_scale64(x: u64, exponent: i64, mode: Mode) {
    let mut native_flags = 0u32;
    let native = unsafe { native_bid64_scalbln(x, exponent as c_long, mode.native, &mut native_flags) };
    let (public, flags) = Decimal64::from_bits(x).scaleb_with_mode(exponent, mode.public);
    assert_eq!((public.to_bits(), public_raw_flags(flags)), (native, native_flags),
        "Decimal64 scaleB mismatch x={x:016x} exponent={exponent} mode={}", mode.name);
}

fn check_scale128(x: Words, exponent: i64, mode: Mode) {
    let mut native_flags = 0u32;
    let native = unsafe { native_bid128_scalbln(c128(x), exponent as c_long, mode.native, &mut native_flags) };
    let (public, flags) = decimal128(x).scaleb_with_mode(exponent, mode.public);
    assert_eq!((decimal128_words(public), public_raw_flags(flags)), (c128_words(native), native_flags),
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
    assert_scale_corpus_contract(32, 0xdec7543253414c45, 1, SCALE_FINITE_TRANSITION_LIMIT32 as i64, RANDOM_CASES32, SCALE_MODE_CROSS_GROUPS32, SCALE_TUPLE_HASH32);
    assert_scale_corpus_contract(64, 0xdec7546453414c45, 1, SCALE_FINITE_TRANSITION_LIMIT64 as i64, RANDOM_CASES64, SCALE_MODE_CROSS_GROUPS64, SCALE_TUPLE_HASH64);
    assert_scale_corpus_contract(128, 0xdec7541253414c45, 2, SCALE_FINITE_TRANSITION_LIMIT128 as i64, RANDOM_CASES128, SCALE_MODE_CROSS_GROUPS128, SCALE_TUPLE_HASH128);
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
    for (op_index, &op) in ROUNDED_OPS.iter().enumerate() {
        let seed = 0xdec7543200000000 ^ op_index as u64;
        for i in 0..RANDOM_CASES32 { if shard.owns(count) { check_rounded32(op, random_word(seed, i, 0) as u32, random_word(seed, i, 1) as u32, MODES[i as usize % MODES.len()]); } count += 1; }
    }
    for (op_index, &op) in UNROUNDED_OPS.iter().enumerate() {
        let seed = 0xdec7543210000000 ^ op_index as u64;
        for i in 0..RANDOM_CASES32 { if shard.owns(count) { check_unrounded32(op, random_word(seed, i, 0) as u32, random_word(seed, i, 1) as u32); } count += 1; }
    }
    let seed = 0xdec7543253414c45;
    for i in 0..RANDOM_CASES32 { if shard.owns(count) { check_scale32(random_scale_operand32(seed, i, SCALE_FINITE_TRANSITION_LIMIT32 as i64), random_scale_exponent(seed, i, 1, SCALE_FINITE_TRANSITION_LIMIT32 as i64), MODES[i as usize % MODES.len()]); } count += 1; }
    let seed = 0xdec7543220000000u64;
    for i in 0..RANDOM_CASES32 { if shard.owns(count) { check_fma32(random_word(seed, i, 0) as u32, random_word(seed, i, 1) as u32, random_word(seed, i, 2) as u32, MODES[i as usize % MODES.len()]); } count += 1; }
    let seed = 0xdec7543230000000u64;
    for i in 0..RANDOM_CASES32 { if shard.owns(count) { check_sqrt32(random_word(seed, i, 0) as u32, MODES[i as usize % MODES.len()]); } count += 1; }
    assert_eq!(count, RANDOM32_COUNT);
    eprintln!("Rust Decimal32 random Tier 1 exact comparisons: {}/{}", shard.owned_count(count), count);

    count = 0;
    for (op_index, &op) in ROUNDED_OPS.iter().enumerate() {
        let seed = 0xdec7546400000000 ^ op_index as u64;
        for i in 0..RANDOM_CASES64 { if shard.owns(count) { check_rounded64(op, random_word(seed, i, 0), random_word(seed, i, 1), MODES[i as usize % MODES.len()]); } count += 1; }
    }
    for (op_index, &op) in UNROUNDED_OPS.iter().enumerate() {
        let seed = 0xdec7546410000000 ^ op_index as u64;
        for i in 0..RANDOM_CASES64 { if shard.owns(count) { check_unrounded64(op, random_word(seed, i, 0), random_word(seed, i, 1)); } count += 1; }
    }
    let seed = 0xdec7546453414c45;
    for i in 0..RANDOM_CASES64 { if shard.owns(count) { check_scale64(random_scale_operand64(seed, i, SCALE_FINITE_TRANSITION_LIMIT64 as i64), random_scale_exponent(seed, i, 1, SCALE_FINITE_TRANSITION_LIMIT64 as i64), MODES[i as usize % MODES.len()]); } count += 1; }
    let seed = 0xdec7546420000000u64;
    for i in 0..RANDOM_CASES64 { if shard.owns(count) { check_fma64(random_word(seed, i, 0), random_word(seed, i, 1), random_word(seed, i, 2), MODES[i as usize % MODES.len()]); } count += 1; }
    let seed = 0xdec7546430000000u64;
    for i in 0..RANDOM_CASES64 { if shard.owns(count) { check_sqrt64(random_word(seed, i, 0), MODES[i as usize % MODES.len()]); } count += 1; }
    assert_eq!(count, RANDOM64_COUNT);
    eprintln!("Rust Decimal64 random Tier 1 exact comparisons: {}/{}", shard.owned_count(count), count);

    count = 0;
    for (op_index, &op) in ROUNDED_OPS.iter().enumerate() {
        let seed = 0xdec7541200000000 ^ op_index as u64;
        for i in 0..RANDOM_CASES128 {
            if shard.owns(count) { check_rounded128(op, Words { lo: random_word(seed, i, 0), hi: random_word(seed, i, 1) }, Words { lo: random_word(seed, i, 2), hi: random_word(seed, i, 3) }, MODES[i as usize % MODES.len()]); }
            count += 1;
        }
    }
    for (op_index, &op) in UNROUNDED_OPS.iter().enumerate() {
        let seed = 0xdec7541210000000 ^ op_index as u64;
        for i in 0..RANDOM_CASES128 {
            if shard.owns(count) { check_unrounded128(op, Words { lo: random_word(seed, i, 0), hi: random_word(seed, i, 1) }, Words { lo: random_word(seed, i, 2), hi: random_word(seed, i, 3) }); }
            count += 1;
        }
    }
    let seed = 0xdec7541253414c45;
    for i in 0..RANDOM_CASES128 {
        if shard.owns(count) { check_scale128(random_scale_operand128(seed, i, SCALE_FINITE_TRANSITION_LIMIT128 as i64), random_scale_exponent(seed, i, 2, SCALE_FINITE_TRANSITION_LIMIT128 as i64), MODES[i as usize % MODES.len()]); }
        count += 1;
    }
    let seed = 0xdec7541220000000u64;
    for i in 0..RANDOM_CASES128 {
        if shard.owns(count) { check_fma128(Words { lo: random_word(seed, i, 0), hi: random_word(seed, i, 1) }, Words { lo: random_word(seed, i, 2), hi: random_word(seed, i, 3) }, Words { lo: random_word(seed, i, 4), hi: random_word(seed, i, 5) }, MODES[i as usize % MODES.len()]); }
        count += 1;
    }
    let seed = 0xdec7541230000000u64;
    for i in 0..RANDOM_CASES128 {
        if shard.owns(count) { check_sqrt128(Words { lo: random_word(seed, i, 0), hi: random_word(seed, i, 1) }, MODES[i as usize % MODES.len()]); }
        count += 1;
    }
    assert_eq!(count, RANDOM128_COUNT);
    eprintln!("Rust Decimal128 random Tier 1 exact comparisons: {}/{}", shard.owned_count(count), count);
}

const PROBES32: [u32; 12] = [
    0x00000000, 0x80000000, 0x32800001, 0xb2800001, 0x00000001, 0x77f8967f,
    0xf7f8967f, 0x78000000, 0x7c000001, 0x7e000001, 0x60000000, 0x5f800000,
];

const PROBES64: [u64; 12] = [
    0x0000000000000000, 0x8000000000000000, 0x31c0000000000001, 0xb1c0000000000001,
    0x0000000000000001, 0x77fb86f26fc0ffff, 0xf7fb86f26fc0ffff, 0x7800000000000000,
    0x7c00000000000001, 0x7e00000000000001, 0x6000000000000000, 0x5fe0000000000000,
];

const PROBES128: [Words; 12] = [
    Words { lo: 0x0000000000000000, hi: 0x0000000000000000 },
    Words { lo: 0x0000000000000000, hi: 0x8000000000000000 },
    Words { lo: 0x0000000000000001, hi: 0x3040000000000000 },
    Words { lo: 0x0000000000000001, hi: 0xb040000000000000 },
    Words { lo: 0x0000000000000001, hi: 0x0000000000000000 },
    Words { lo: 0x378d8e63ffffffff, hi: 0x5fffed09bead87c0 },
    Words { lo: 0x378d8e63ffffffff, hi: 0xdfffed09bead87c0 },
    Words { lo: 0x0000000000000000, hi: 0x7800000000000000 },
    Words { lo: 0x0000000000000001, hi: 0x7c00000000000000 },
    Words { lo: 0x0000000000000001, hi: 0x7e00000000000000 },
    Words { lo: 0x0000000000000000, hi: 0x6000000000000000 },
    Words { lo: 0x0000000000000000, hi: 0x5ffe000000000000 },
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
