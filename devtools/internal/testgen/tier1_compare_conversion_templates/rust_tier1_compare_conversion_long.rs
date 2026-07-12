#![cfg(feature = "tier1-long")]

use bid754::generated::bid32_exports::{bid32_max_num, bid32_max_num_mag, bid32_min_num, bid32_min_num_mag};
use bid754::{Decimal128, Decimal32, Decimal64, ExceptionFlags, RoundingMode};
use libbid_sys::BID_UINT128 as C128;

const BOUNDARY32_COUNT: u64 = @@TIER1_BOUNDARY32_COUNT@@;
const BOUNDARY64_COUNT: u64 = @@TIER1_BOUNDARY64_COUNT@@;
const BOUNDARY128_COUNT: u64 = @@TIER1_BOUNDARY128_COUNT@@;
const SEMANTIC32_COUNT: u64 = @@TIER1_CONVERSION_SEMANTIC32_COUNT@@;
const SEMANTIC64_COUNT: u64 = @@TIER1_CONVERSION_SEMANTIC64_COUNT@@;
const SEMANTIC128_COUNT: u64 = @@TIER1_CONVERSION_SEMANTIC128_COUNT@@;
const CONSTRUCTOR_I32_COUNT: u64 = @@TIER1_CONSTRUCTOR_INT32_COUNT@@;
const CONSTRUCTOR_U32_COUNT: u64 = @@TIER1_CONSTRUCTOR_UINT32_COUNT@@;
const CONSTRUCTOR_I64_COUNT: u64 = @@TIER1_CONSTRUCTOR_INT64_COUNT@@;
const CONSTRUCTOR_U64_COUNT: u64 = @@TIER1_CONSTRUCTOR_UINT64_COUNT@@;

const COMPARE_STRUCTURED32: u64 = @@TIER1_COMPARE_STRUCTURED32_COUNT@@;
const COMPARE_STRUCTURED64: u64 = @@TIER1_COMPARE_STRUCTURED64_COUNT@@;
const COMPARE_STRUCTURED128: u64 = @@TIER1_COMPARE_STRUCTURED128_COUNT@@;
const COMPARE_RANDOM32: u64 = @@TIER1_COMPARE_RANDOM32_COUNT@@;
const COMPARE_RANDOM64: u64 = @@TIER1_COMPARE_RANDOM64_COUNT@@;
const COMPARE_RANDOM128: u64 = @@TIER1_COMPARE_RANDOM128_COUNT@@;
const COMPARE_TOTAL32: u64 = @@TIER1_COMPARE_TOTAL32_COUNT@@;
const COMPARE_TOTAL64: u64 = @@TIER1_COMPARE_TOTAL64_COUNT@@;
const COMPARE_TOTAL128: u64 = @@TIER1_COMPARE_TOTAL128_COUNT@@;

const TO_INT_STRUCTURED32: u64 = @@TIER1_TO_INT_STRUCTURED32_COUNT@@;
const TO_INT_STRUCTURED64: u64 = @@TIER1_TO_INT_STRUCTURED64_COUNT@@;
const TO_INT_STRUCTURED128: u64 = @@TIER1_TO_INT_STRUCTURED128_COUNT@@;
const TO_INT_TOTAL32: u64 = @@TIER1_TO_INT_TOTAL32_COUNT@@;
const TO_INT_TOTAL64: u64 = @@TIER1_TO_INT_TOTAL64_COUNT@@;
const TO_INT_TOTAL128: u64 = @@TIER1_TO_INT_TOTAL128_COUNT@@;
const WIDTH_STRUCTURED32: u64 = @@TIER1_WIDTH_STRUCTURED32_COUNT@@;
const WIDTH_STRUCTURED64: u64 = @@TIER1_WIDTH_STRUCTURED64_COUNT@@;
const WIDTH_STRUCTURED128: u64 = @@TIER1_WIDTH_STRUCTURED128_COUNT@@;
const WIDTH_TOTAL32: u64 = @@TIER1_WIDTH_TOTAL32_COUNT@@;
const WIDTH_TOTAL64: u64 = @@TIER1_WIDTH_TOTAL64_COUNT@@;
const WIDTH_TOTAL128: u64 = @@TIER1_WIDTH_TOTAL128_COUNT@@;
const CONSTRUCTOR_STRUCTURED: u64 = @@TIER1_CONSTRUCTOR_STRUCTURED_COUNT@@;
const CONSTRUCTOR_TOTAL: u64 = @@TIER1_CONSTRUCTOR_TOTAL_COUNT@@;
const CONSTRUCTOR_CONVENIENCE: u64 = @@TIER1_CONSTRUCTOR_CONVENIENCE_COUNT@@;
const CONVERSION_STRUCTURED: u64 = @@TIER1_CONVERSION_STRUCTURED_COUNT@@;
const CONVERSION_RANDOM: u64 = @@TIER1_CONVERSION_RANDOM_COUNT@@;
const CONVERSION_TOTAL: u64 = @@TIER1_CONVERSION_TOTAL_COUNT@@;
const QUIET_PREDICATE_COUNT: u64 = 6;
const MINMAX_OPERATION_COUNT: u64 = 4;
const TO_INT_OPERATION_COUNT: u64 = 80;
const WIDTH_OPERATION_COUNT32: u64 = 2;
const WIDTH_OPERATION_COUNT64: u64 = 6;
const WIDTH_OPERATION_COUNT128: u64 = 10;
const CONSTRUCTOR_OPERATION_COUNT: u64 = 36;

const COMPARE_RANDOM_PAIRS32: u64 = 1 << 20;
const COMPARE_RANDOM_PAIRS64: u64 = 1 << 20;
const COMPARE_RANDOM_PAIRS128: u64 = 1 << 19;
const TO_INT_RANDOM32: u64 = 1 << 18;
const TO_INT_RANDOM64: u64 = 1 << 18;
const TO_INT_RANDOM128: u64 = 1 << 17;
const WIDTH_RANDOM32: u64 = 1 << 18;
const WIDTH_RANDOM64: u64 = 1 << 18;
const WIDTH_RANDOM128: u64 = 1 << 17;
const CONSTRUCTOR_RANDOM: u64 = 1 << 20;

#[derive(Clone, Copy, Debug)]
struct Mode { public: RoundingMode, native: u32 }

const MODES: [Mode; 5] = [
    Mode { public: RoundingMode::NearestEven, native: 0 },
    Mode { public: RoundingMode::NearestAway, native: 4 },
    Mode { public: RoundingMode::TowardZero, native: 3 },
    Mode { public: RoundingMode::TowardPositive, native: 2 },
    Mode { public: RoundingMode::TowardNegative, native: 1 },
];

#[derive(Clone, Copy, Debug, PartialEq, Eq)]
struct Words { lo: u64, hi: u64 }

#[derive(Clone, Copy, Debug)]
enum RawDecimal { D32(u32), D64(u64), D128(Words) }

#[derive(Clone, Copy, Debug)]
enum QuietOp { Equal, NotEqual, Less, LessEqual, Greater, GreaterEqual }
const QUIET_OPS: [QuietOp; 6] = [
    QuietOp::Equal, QuietOp::NotEqual, QuietOp::Less, QuietOp::LessEqual,
    QuietOp::Greater, QuietOp::GreaterEqual,
];

#[derive(Clone, Copy, Debug)]
enum MinMaxOp { MinNum, MaxNum, MinNumMag, MaxNumMag }
const MINMAX_OPS: [MinMaxOp; 4] = [MinMaxOp::MinNum, MinMaxOp::MaxNum, MinMaxOp::MinNumMag, MinMaxOp::MaxNumMag];

#[derive(Clone, Copy, Debug)]
enum IntKind { I8, I16, I32, I64, U8, U16, U32, U64 }
const INT_KINDS: [IntKind; 8] = [
    IntKind::I8, IntKind::I16, IntKind::I32, IntKind::I64,
    IntKind::U8, IntKind::U16, IntKind::U32, IntKind::U64,
];

#[derive(Clone, Copy, Debug)]
enum IntSuffix { Rnint, Rninta, Int, Ceil, Floor, Xrnint, Xrninta, Xint, Xceil, Xfloor }

#[derive(Clone, Copy, Debug)]
struct ToIntOp { kind: IntKind, exact: bool, mode: Mode, suffix: IntSuffix }

#[derive(Clone, Copy, Debug)]
struct WidthOp { source: u8, dest: u8, mode: Option<Mode> }

#[derive(Clone, Copy, Debug)]
enum ConstructorKind { I32, U32, I64, U64 }

#[derive(Clone, Copy, Debug)]
struct ConstructorOp { dest: u8, kind: ConstructorKind, mode: Option<Mode> }

#[derive(Clone, Copy, Debug, PartialEq, Eq)]
enum Bits { D32(u32), D64(u64), D128(Words) }

#[derive(Clone, Copy)]
struct Shard { count: u64, index: u64 }

impl Shard {
    fn load() -> Self {
        let count = std::env::var("BID754_TIER1_COMPARE_CONVERSION_SHARD_COUNT").ok();
        let index = std::env::var("BID754_TIER1_COMPARE_CONVERSION_SHARD_INDEX").ok();
        match (count, index) {
            (None, None) => Shard { count: 1, index: 0 },
            (Some(count), Some(index)) => {
                let count = count.parse::<u64>().expect("invalid Tier 1 compare/conversion shard count");
                let index = index.parse::<u64>().expect("invalid Tier 1 compare/conversion shard index");
                assert!(count > 0 && index < count, "invalid Tier 1 compare/conversion shard");
                Shard { count, index }
            }
            _ => panic!("BID754_TIER1_COMPARE_CONVERSION_SHARD_COUNT and BID754_TIER1_COMPARE_CONVERSION_SHARD_INDEX must be set together"),
        }
    }
    fn owns(self, case_index: u64) -> bool { case_index % self.count == self.index }
    fn owned_count(self, total: u64) -> u64 {
        if total <= self.index { 0 } else { 1 + (total - 1 - self.index) / self.count }
    }
}

unsafe extern "C" {
@@TIER1_TO_INT_EXTERN_DECLS@@
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

fn splitmix64(mut value: u64) -> u64 {
    value = value.wrapping_add(0x9e3779b97f4a7c15);
    value = (value ^ (value >> 30)).wrapping_mul(0xbf58476d1ce4e5b9);
    value = (value ^ (value >> 27)).wrapping_mul(0x94d049bb133111eb);
    value ^ (value >> 31)
}

fn random_word(seed: u64, case_index: u64, lane: u64) -> u64 {
    splitmix64(seed ^ case_index.wrapping_mul(0xd1342543de82ef95) ^ lane.wrapping_mul(0x9e3779b97f4a7c15))
}

fn check_quiet32(op: QuietOp, x: u32, y: u32) {
    let mut native_flags = 0u32;
    let native = unsafe { match op {
        QuietOp::Equal => libbid_sys::bid32_quiet_equal(x, y, &mut native_flags),
        QuietOp::NotEqual => libbid_sys::bid32_quiet_not_equal(x, y, &mut native_flags),
        QuietOp::Less => libbid_sys::bid32_quiet_less(x, y, &mut native_flags),
        QuietOp::LessEqual => libbid_sys::bid32_quiet_less_equal(x, y, &mut native_flags),
        QuietOp::Greater => libbid_sys::bid32_quiet_greater(x, y, &mut native_flags),
        QuietOp::GreaterEqual => libbid_sys::bid32_quiet_greater_equal(x, y, &mut native_flags),
    }} != 0;
    let left = Decimal32::from_bits(x); let right = Decimal32::from_bits(y);
    let (public, flags) = match op {
        QuietOp::Equal => left.quiet_eq(right), QuietOp::NotEqual => left.quiet_ne(right),
        QuietOp::Less => left.quiet_lt(right), QuietOp::LessEqual => left.quiet_le(right),
        QuietOp::Greater => left.quiet_gt(right), QuietOp::GreaterEqual => left.quiet_ge(right),
    };
    assert_eq!((public, public_raw_flags(flags)), (native, native_flags), "Decimal32 {op:?} x={x:08x} y={y:08x}");
}

fn check_quiet64(op: QuietOp, x: u64, y: u64) {
    let mut native_flags = 0u32;
    let native = unsafe { match op {
        QuietOp::Equal => libbid_sys::bid64_quiet_equal(x, y, &mut native_flags),
        QuietOp::NotEqual => libbid_sys::bid64_quiet_not_equal(x, y, &mut native_flags),
        QuietOp::Less => libbid_sys::bid64_quiet_less(x, y, &mut native_flags),
        QuietOp::LessEqual => libbid_sys::bid64_quiet_less_equal(x, y, &mut native_flags),
        QuietOp::Greater => libbid_sys::bid64_quiet_greater(x, y, &mut native_flags),
        QuietOp::GreaterEqual => libbid_sys::bid64_quiet_greater_equal(x, y, &mut native_flags),
    }} != 0;
    let left = Decimal64::from_bits(x); let right = Decimal64::from_bits(y);
    let (public, flags) = match op {
        QuietOp::Equal => left.quiet_eq(right), QuietOp::NotEqual => left.quiet_ne(right),
        QuietOp::Less => left.quiet_lt(right), QuietOp::LessEqual => left.quiet_le(right),
        QuietOp::Greater => left.quiet_gt(right), QuietOp::GreaterEqual => left.quiet_ge(right),
    };
    assert_eq!((public, public_raw_flags(flags)), (native, native_flags), "Decimal64 {op:?} x={x:016x} y={y:016x}");
}

fn check_quiet128(op: QuietOp, x: Words, y: Words) {
    let mut native_flags = 0u32;
    let native = unsafe { match op {
        QuietOp::Equal => libbid_sys::bid128_quiet_equal(c128(x), c128(y), &mut native_flags),
        QuietOp::NotEqual => libbid_sys::bid128_quiet_not_equal(c128(x), c128(y), &mut native_flags),
        QuietOp::Less => libbid_sys::bid128_quiet_less(c128(x), c128(y), &mut native_flags),
        QuietOp::LessEqual => libbid_sys::bid128_quiet_less_equal(c128(x), c128(y), &mut native_flags),
        QuietOp::Greater => libbid_sys::bid128_quiet_greater(c128(x), c128(y), &mut native_flags),
        QuietOp::GreaterEqual => libbid_sys::bid128_quiet_greater_equal(c128(x), c128(y), &mut native_flags),
    }} != 0;
    let left = decimal128(x); let right = decimal128(y);
    let (public, flags) = match op {
        QuietOp::Equal => left.quiet_eq(right), QuietOp::NotEqual => left.quiet_ne(right),
        QuietOp::Less => left.quiet_lt(right), QuietOp::LessEqual => left.quiet_le(right),
        QuietOp::Greater => left.quiet_gt(right), QuietOp::GreaterEqual => left.quiet_ge(right),
    };
    assert_eq!((public, public_raw_flags(flags)), (native, native_flags), "Decimal128 {op:?} x={x:?} y={y:?}");
}

fn check_minmax32(op: MinMaxOp, x: u32, y: u32) {
    let mut native_flags = 0u32;
    let native = unsafe { match op {
        MinMaxOp::MinNum => libbid_sys::bid32_minnum(x, y, &mut native_flags),
        MinMaxOp::MaxNum => libbid_sys::bid32_maxnum(x, y, &mut native_flags),
        MinMaxOp::MinNumMag => libbid_sys::bid32_minnum_mag(x, y, &mut native_flags),
        MinMaxOp::MaxNumMag => libbid_sys::bid32_maxnum_mag(x, y, &mut native_flags),
    }};
    let left = Decimal32::from_bits(x); let right = Decimal32::from_bits(y);
    let (public, flags) = match op {
        MinMaxOp::MinNum => left.min_num(right), MinMaxOp::MaxNum => left.max_num(right),
        MinMaxOp::MinNumMag => left.min_num_mag(right), MinMaxOp::MaxNumMag => left.max_num_mag(right),
    };
    assert_eq!((public.to_bits(), public_raw_flags(flags)), (native, native_flags), "Decimal32 {op:?} x={x:08x} y={y:08x}");
    // The value-only generated bodies (bid32_minnum_pure / bid32_maxnum_pure /
    // bid32_minnum_mag_pure / bid32_maxnum_mag_pure) are separate
    // implementations from the status-aware bodies, so they are compared
    // against the native value bits directly, mirroring the Go runner.
    let pure = match op {
        MinMaxOp::MinNum => bid32_min_num(x, y), MinMaxOp::MaxNum => bid32_max_num(x, y),
        MinMaxOp::MinNumMag => bid32_min_num_mag(x, y), MinMaxOp::MaxNumMag => bid32_max_num_mag(x, y),
    };
    assert_eq!(pure, native, "Decimal32 pure {op:?} x={x:08x} y={y:08x}");
}

fn check_minmax64(op: MinMaxOp, x: u64, y: u64) {
    let mut native_flags = 0u32;
    let native = unsafe { match op {
        MinMaxOp::MinNum => libbid_sys::bid64_minnum(x, y, &mut native_flags),
        MinMaxOp::MaxNum => libbid_sys::bid64_maxnum(x, y, &mut native_flags),
        MinMaxOp::MinNumMag => libbid_sys::bid64_minnum_mag(x, y, &mut native_flags),
        MinMaxOp::MaxNumMag => libbid_sys::bid64_maxnum_mag(x, y, &mut native_flags),
    }};
    let left = Decimal64::from_bits(x); let right = Decimal64::from_bits(y);
    let (public, flags) = match op {
        MinMaxOp::MinNum => left.min_num(right), MinMaxOp::MaxNum => left.max_num(right),
        MinMaxOp::MinNumMag => left.min_num_mag(right), MinMaxOp::MaxNumMag => left.max_num_mag(right),
    };
    assert_eq!((public.to_bits(), public_raw_flags(flags)), (native, native_flags), "Decimal64 {op:?} x={x:016x} y={y:016x}");
}

fn check_minmax128(op: MinMaxOp, x: Words, y: Words) {
    let mut native_flags = 0u32;
    let native = unsafe { match op {
        MinMaxOp::MinNum => libbid_sys::bid128_minnum(c128(x), c128(y), &mut native_flags),
        MinMaxOp::MaxNum => libbid_sys::bid128_maxnum(c128(x), c128(y), &mut native_flags),
        MinMaxOp::MinNumMag => libbid_sys::bid128_minnum_mag(c128(x), c128(y), &mut native_flags),
        MinMaxOp::MaxNumMag => libbid_sys::bid128_maxnum_mag(c128(x), c128(y), &mut native_flags),
    }};
    let left = decimal128(x); let right = decimal128(y);
    let (public, flags) = match op {
        MinMaxOp::MinNum => left.min_num(right), MinMaxOp::MaxNum => left.max_num(right),
        MinMaxOp::MinNumMag => left.min_num_mag(right), MinMaxOp::MaxNumMag => left.max_num_mag(right),
    };
    assert_eq!((decimal128_words(public), public_raw_flags(flags)), (c128_words(native), native_flags), "Decimal128 {op:?} x={x:?} y={y:?}");
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

fn expected_quiet(op: QuietOp, relation: i8) -> bool {
    match op {
        QuietOp::Equal => relation == 0,
        QuietOp::NotEqual => relation != 0,
        QuietOp::Less => relation == -1,
        QuietOp::LessEqual => relation == -1 || relation == 0,
        QuietOp::Greater => relation == 1,
        QuietOp::GreaterEqual => relation == 1 || relation == 0,
    }
}

fn check_semantic32(x: u32, y: u32, relation: i8, expected_flags: ExceptionFlags) {
    let left = Decimal32::from_bits(x); let right = Decimal32::from_bits(y);
    for &op in &QUIET_OPS {
        let (got, flags) = match op {
            QuietOp::Equal => left.quiet_eq(right), QuietOp::NotEqual => left.quiet_ne(right),
            QuietOp::Less => left.quiet_lt(right), QuietOp::LessEqual => left.quiet_le(right),
            QuietOp::Greater => left.quiet_gt(right), QuietOp::GreaterEqual => left.quiet_ge(right),
        };
        assert_eq!((got, flags), (expected_quiet(op, relation), expected_flags), "Decimal32 semantic {op:?}");
    }
}
fn check_semantic64(x: u64, y: u64, relation: i8, expected_flags: ExceptionFlags) {
    let left = Decimal64::from_bits(x); let right = Decimal64::from_bits(y);
    for &op in &QUIET_OPS {
        let (got, flags) = match op {
            QuietOp::Equal => left.quiet_eq(right), QuietOp::NotEqual => left.quiet_ne(right),
            QuietOp::Less => left.quiet_lt(right), QuietOp::LessEqual => left.quiet_le(right),
            QuietOp::Greater => left.quiet_gt(right), QuietOp::GreaterEqual => left.quiet_ge(right),
        };
        assert_eq!((got, flags), (expected_quiet(op, relation), expected_flags), "Decimal64 semantic {op:?}");
    }
}
fn check_semantic128(x: Words, y: Words, relation: i8, expected_flags: ExceptionFlags) {
    let left = decimal128(x); let right = decimal128(y);
    for &op in &QUIET_OPS {
        let (got, flags) = match op {
            QuietOp::Equal => left.quiet_eq(right), QuietOp::NotEqual => left.quiet_ne(right),
            QuietOp::Less => left.quiet_lt(right), QuietOp::LessEqual => left.quiet_le(right),
            QuietOp::Greater => left.quiet_gt(right), QuietOp::GreaterEqual => left.quiet_ge(right),
        };
        assert_eq!((got, flags), (expected_quiet(op, relation), expected_flags), "Decimal128 semantic {op:?}");
    }
}

fn to_int_ops() -> Vec<ToIntOp> {
    let mut result = Vec::with_capacity(80);
    for &kind in &INT_KINDS {
        for exact in [false, true] {
            for &mode in &MODES {
                let suffix = match (exact, mode.public) {
                    (false, RoundingMode::NearestEven) => IntSuffix::Rnint,
                    (false, RoundingMode::NearestAway) => IntSuffix::Rninta,
                    (false, RoundingMode::TowardZero) => IntSuffix::Int,
                    (false, RoundingMode::TowardPositive) => IntSuffix::Ceil,
                    (false, RoundingMode::TowardNegative) => IntSuffix::Floor,
                    (true, RoundingMode::NearestEven) => IntSuffix::Xrnint,
                    (true, RoundingMode::NearestAway) => IntSuffix::Xrninta,
                    (true, RoundingMode::TowardZero) => IntSuffix::Xint,
                    (true, RoundingMode::TowardPositive) => IntSuffix::Xceil,
                    (true, RoundingMode::TowardNegative) => IntSuffix::Xfloor,
                };
                result.push(ToIntOp { kind, exact, mode, suffix });
            }
        }
    }
    result
}

fn native_to_int(value: RawDecimal, op: ToIntOp) -> (u64, u32) {
    let mut flags = 0u32;
    let result = unsafe { match (value, op.kind, op.suffix) {
@@TIER1_TO_INT_NATIVE_DISPATCH@@
    }};
    (result, flags)
}

macro_rules! public_to_int_value {
    ($value:expr, $op:expr) => {{
        let (result, flags) = match ($op.kind, $op.exact) {
            (IntKind::I8, false) => { let (v, f) = $value.to_i8($op.mode.public); (v as i64 as u64, f) },
            (IntKind::I8, true) => { let (v, f) = $value.to_i8_exact($op.mode.public); (v as i64 as u64, f) },
            (IntKind::I16, false) => { let (v, f) = $value.to_i16($op.mode.public); (v as i64 as u64, f) },
            (IntKind::I16, true) => { let (v, f) = $value.to_i16_exact($op.mode.public); (v as i64 as u64, f) },
            (IntKind::I32, false) => { let (v, f) = $value.to_i32($op.mode.public); (v as i64 as u64, f) },
            (IntKind::I32, true) => { let (v, f) = $value.to_i32_exact($op.mode.public); (v as i64 as u64, f) },
            (IntKind::I64, false) => { let (v, f) = $value.to_i64($op.mode.public); (v as u64, f) },
            (IntKind::I64, true) => { let (v, f) = $value.to_i64_exact($op.mode.public); (v as u64, f) },
            (IntKind::U8, false) => { let (v, f) = $value.to_u8($op.mode.public); (v as u64, f) },
            (IntKind::U8, true) => { let (v, f) = $value.to_u8_exact($op.mode.public); (v as u64, f) },
            (IntKind::U16, false) => { let (v, f) = $value.to_u16($op.mode.public); (v as u64, f) },
            (IntKind::U16, true) => { let (v, f) = $value.to_u16_exact($op.mode.public); (v as u64, f) },
            (IntKind::U32, false) => { let (v, f) = $value.to_u32($op.mode.public); (v as u64, f) },
            (IntKind::U32, true) => { let (v, f) = $value.to_u32_exact($op.mode.public); (v as u64, f) },
            (IntKind::U64, false) => { let (v, f) = $value.to_u64($op.mode.public); (v, f) },
            (IntKind::U64, true) => { let (v, f) = $value.to_u64_exact($op.mode.public); (v, f) },
        };
        (result, public_raw_flags(flags))
    }};
}

fn public_to_int(value: RawDecimal, op: ToIntOp) -> (u64, u32) {
    match value {
        RawDecimal::D32(x) => public_to_int_value!(Decimal32::from_bits(x), op),
        RawDecimal::D64(x) => public_to_int_value!(Decimal64::from_bits(x), op),
        RawDecimal::D128(x) => public_to_int_value!(decimal128(x), op),
    }
}

fn check_to_int(value: RawDecimal, op: ToIntOp) {
    assert_eq!(public_to_int(value, op), native_to_int(value, op), "BID-to-integer mismatch value={value:?} op={op:?}");
}

fn width_ops(source: u8) -> Vec<WidthOp> {
    match source {
        32 => vec![WidthOp { source: 32, dest: 64, mode: None }, WidthOp { source: 32, dest: 128, mode: None }],
        64 => {
            let mut v: Vec<_> = MODES.iter().copied().map(|mode| WidthOp { source: 64, dest: 32, mode: Some(mode) }).collect();
            v.push(WidthOp { source: 64, dest: 128, mode: None }); v
        }
        128 => {
            let mut v: Vec<_> = MODES.iter().copied().map(|mode| WidthOp { source: 128, dest: 32, mode: Some(mode) }).collect();
            v.extend(MODES.iter().copied().map(|mode| WidthOp { source: 128, dest: 64, mode: Some(mode) })); v
        }
        _ => panic!("unsupported width source"),
    }
}

fn native_width(value: RawDecimal, op: WidthOp) -> (Bits, u32) {
    let mut flags = 0u32;
    let bits = unsafe { match (value, op.source, op.dest) {
        (RawDecimal::D32(x), 32, 64) => Bits::D64(libbid_sys::bid32_to_bid64(x, &mut flags)),
        (RawDecimal::D32(x), 32, 128) => Bits::D128(c128_words(libbid_sys::bid32_to_bid128(x, &mut flags))),
        (RawDecimal::D64(x), 64, 32) => Bits::D32(libbid_sys::bid64_to_bid32(x, op.mode.unwrap().native, &mut flags)),
        (RawDecimal::D64(x), 64, 128) => Bits::D128(c128_words(libbid_sys::bid64_to_bid128(x, &mut flags))),
        (RawDecimal::D128(x), 128, 32) => Bits::D32(libbid_sys::bid128_to_bid32(c128(x), op.mode.unwrap().native, &mut flags)),
        (RawDecimal::D128(x), 128, 64) => Bits::D64(libbid_sys::bid128_to_bid64(c128(x), op.mode.unwrap().native, &mut flags)),
        _ => panic!("invalid width operation {op:?} for {value:?}"),
    }};
    (bits, flags)
}

fn public_width(value: RawDecimal, op: WidthOp) -> (Bits, u32) {
    let (bits, flags) = match (value, op.source, op.dest) {
        (RawDecimal::D32(x), 32, 64) => { let (v, f) = Decimal32::from_bits(x).to_decimal64(); (Bits::D64(v.to_bits()), f) },
        (RawDecimal::D32(x), 32, 128) => { let (v, f) = Decimal32::from_bits(x).to_decimal128(); (Bits::D128(decimal128_words(v)), f) },
        (RawDecimal::D64(x), 64, 32) => { let (v, f) = Decimal64::from_bits(x).to_decimal32(op.mode.unwrap().public); (Bits::D32(v.to_bits()), f) },
        (RawDecimal::D64(x), 64, 128) => { let (v, f) = Decimal64::from_bits(x).to_decimal128(); (Bits::D128(decimal128_words(v)), f) },
        (RawDecimal::D128(x), 128, 32) => { let (v, f) = decimal128(x).to_decimal32(op.mode.unwrap().public); (Bits::D32(v.to_bits()), f) },
        (RawDecimal::D128(x), 128, 64) => { let (v, f) = decimal128(x).to_decimal64(op.mode.unwrap().public); (Bits::D64(v.to_bits()), f) },
        _ => panic!("invalid public width operation {op:?} for {value:?}"),
    };
    (bits, public_raw_flags(flags))
}

fn check_width(value: RawDecimal, op: WidthOp) {
    assert_eq!(public_width(value, op), native_width(value, op), "width-conversion mismatch value={value:?} op={op:?}");
}

fn constructor_ops() -> Vec<ConstructorOp> {
    let mut result = Vec::with_capacity(36);
    for kind in [ConstructorKind::I32, ConstructorKind::U32, ConstructorKind::I64, ConstructorKind::U64] {
        result.extend(MODES.iter().copied().map(|mode| ConstructorOp { dest: 32, kind, mode: Some(mode) }));
    }
    result.push(ConstructorOp { dest: 64, kind: ConstructorKind::I32, mode: None });
    result.push(ConstructorOp { dest: 64, kind: ConstructorKind::U32, mode: None });
    for kind in [ConstructorKind::I64, ConstructorKind::U64] {
        result.extend(MODES.iter().copied().map(|mode| ConstructorOp { dest: 64, kind, mode: Some(mode) }));
    }
    for kind in [ConstructorKind::I32, ConstructorKind::U32, ConstructorKind::I64, ConstructorKind::U64] {
        result.push(ConstructorOp { dest: 128, kind, mode: None });
    }
    result
}

fn native_constructor(raw: u64, op: ConstructorOp) -> (Bits, u32) {
    let mut flags = 0u32;
    let bits = unsafe { match (op.dest, op.kind) {
        (32, ConstructorKind::I32) => Bits::D32(libbid_sys::bid32_from_int32(raw as u32 as i32, op.mode.unwrap().native, &mut flags)),
        (32, ConstructorKind::U32) => Bits::D32(libbid_sys::bid32_from_uint32(raw as u32, op.mode.unwrap().native, &mut flags)),
        (32, ConstructorKind::I64) => Bits::D32(libbid_sys::bid32_from_int64(raw as i64, op.mode.unwrap().native, &mut flags)),
        (32, ConstructorKind::U64) => Bits::D32(libbid_sys::bid32_from_uint64(raw, op.mode.unwrap().native, &mut flags)),
        (64, ConstructorKind::I32) => Bits::D64(libbid_sys::bid64_from_int32(raw as u32 as i32)),
        (64, ConstructorKind::U32) => Bits::D64(libbid_sys::bid64_from_uint32(raw as u32)),
        (64, ConstructorKind::I64) => Bits::D64(libbid_sys::bid64_from_int64(raw as i64, op.mode.unwrap().native, &mut flags)),
        (64, ConstructorKind::U64) => Bits::D64(libbid_sys::bid64_from_uint64(raw, op.mode.unwrap().native, &mut flags)),
        (128, ConstructorKind::I32) => Bits::D128(c128_words(libbid_sys::bid128_from_int32(raw as u32 as i32))),
        (128, ConstructorKind::U32) => Bits::D128(c128_words(libbid_sys::bid128_from_uint32(raw as u32))),
        (128, ConstructorKind::I64) => Bits::D128(c128_words(libbid_sys::bid128_from_int64(raw as i64))),
        (128, ConstructorKind::U64) => Bits::D128(c128_words(libbid_sys::bid128_from_uint64(raw))),
        _ => panic!("invalid constructor {op:?}"),
    }};
    (bits, flags)
}

fn public_constructor(raw: u64, op: ConstructorOp) -> (Bits, u32) {
    let (bits, flags) = match (op.dest, op.kind) {
        (32, ConstructorKind::I32) => { let (v, f) = Decimal32::from_i32(raw as u32 as i32, op.mode.unwrap().public); (Bits::D32(v.to_bits()), f) },
        (32, ConstructorKind::U32) => { let (v, f) = Decimal32::from_u32(raw as u32, op.mode.unwrap().public); (Bits::D32(v.to_bits()), f) },
        (32, ConstructorKind::I64) => { let (v, f) = Decimal32::from_i64(raw as i64, op.mode.unwrap().public); (Bits::D32(v.to_bits()), f) },
        (32, ConstructorKind::U64) => { let (v, f) = Decimal32::from_u64(raw, op.mode.unwrap().public); (Bits::D32(v.to_bits()), f) },
        (64, ConstructorKind::I32) => (Bits::D64(Decimal64::from(raw as u32 as i32).to_bits()), ExceptionFlags::empty()),
        (64, ConstructorKind::U32) => (Bits::D64(Decimal64::from(raw as u32).to_bits()), ExceptionFlags::empty()),
        (64, ConstructorKind::I64) => { let (v, f) = Decimal64::from_i64(raw as i64, op.mode.unwrap().public); (Bits::D64(v.to_bits()), f) },
        (64, ConstructorKind::U64) => { let (v, f) = Decimal64::from_u64(raw, op.mode.unwrap().public); (Bits::D64(v.to_bits()), f) },
        (128, ConstructorKind::I32) => (Bits::D128(decimal128_words(Decimal128::from(raw as u32 as i32))), ExceptionFlags::empty()),
        (128, ConstructorKind::U32) => (Bits::D128(decimal128_words(Decimal128::from(raw as u32))), ExceptionFlags::empty()),
        (128, ConstructorKind::I64) => (Bits::D128(decimal128_words(Decimal128::from(raw as i64))), ExceptionFlags::empty()),
        (128, ConstructorKind::U64) => (Bits::D128(decimal128_words(Decimal128::from(raw))), ExceptionFlags::empty()),
        _ => panic!("invalid public constructor {op:?}"),
    };
    (bits, public_raw_flags(flags))
}

fn check_constructor(raw: u64, op: ConstructorOp) {
    assert_eq!(public_constructor(raw, op), native_constructor(raw, op), "integer constructor mismatch raw={raw:016x} op={op:?}");
}

#[test]
fn tier1_compare_conversion_corpus_contract() {
    assert_eq!(std::mem::size_of::<C128>(), 16);
    assert_eq!(BOUNDARY32.len(), BOUNDARY32_COUNT as usize);
    assert_eq!(BOUNDARY64.len(), BOUNDARY64_COUNT as usize);
    assert_eq!(BOUNDARY128.len(), BOUNDARY128_COUNT as usize);
    assert_eq!(COMPARE_TOTAL32, COMPARE_STRUCTURED32 + COMPARE_RANDOM32);
    assert_eq!(COMPARE_TOTAL64, COMPARE_STRUCTURED64 + COMPARE_RANDOM64);
    assert_eq!(COMPARE_TOTAL128, COMPARE_STRUCTURED128 + COMPARE_RANDOM128);
    assert_eq!(CONVERSION_TOTAL, CONVERSION_STRUCTURED + CONVERSION_RANDOM);
    assert_eq!(QUIET_OPS.len(), QUIET_PREDICATE_COUNT as usize);
    assert_eq!(MINMAX_OPS.len(), MINMAX_OPERATION_COUNT as usize);
    assert_eq!(to_int_ops().len(), TO_INT_OPERATION_COUNT as usize);
    assert_eq!(width_ops(32).len(), WIDTH_OPERATION_COUNT32 as usize);
    assert_eq!(width_ops(64).len(), WIDTH_OPERATION_COUNT64 as usize);
    assert_eq!(width_ops(128).len(), WIDTH_OPERATION_COUNT128 as usize);
    assert_eq!(constructor_ops().len(), CONSTRUCTOR_OPERATION_COUNT as usize);
    assert_eq!(random_word(0xdec75432c04d5001, 0, 0), @@TIER1_RANDOM_SAMPLE0@@);
    assert_eq!(random_word(0xdec75464c0a70001, (1 << 18) - 1, 0), @@TIER1_RANDOM_SAMPLE1@@);
    assert_eq!(random_word(0xdec754c0c0de0001, (1 << 17) - 1, 1), @@TIER1_RANDOM_SAMPLE2@@);
}

#[test]
fn tier1_quiet_comparison_semantic_matrix() {
    let invalid = ExceptionFlags::INVALID_OPERATION;
    for (x, y, relation, flags) in [
        (0x32800001, 0x32800001, 0, ExceptionFlags::empty()), (0xb2800001, 0x32800001, -1, ExceptionFlags::empty()),
        (0x32800001, 0xb2800001, 1, ExceptionFlags::empty()), (0x00000000, 0x80000000, 0, ExceptionFlags::empty()),
        (0x78000000, 0x32800001, 1, ExceptionFlags::empty()), (0x32800001, 0x78000000, -1, ExceptionFlags::empty()),
        (0x7c000001, 0x32800001, 2, ExceptionFlags::empty()), (0x32800001, 0x7c000001, 2, ExceptionFlags::empty()),
        (0x7e000001, 0x32800001, 2, invalid), (0x32800001, 0x7e000001, 2, invalid),
        (0xf8000000, 0x32800001, -1, ExceptionFlags::empty()), (0x32800001, 0xf8000000, 1, ExceptionFlags::empty()),
        (0xf8000000, 0x78000000, -1, ExceptionFlags::empty()), (0x78000000, 0xf8000000, 1, ExceptionFlags::empty()),
        (0x78000000, 0x78000000, 0, ExceptionFlags::empty()), (0xf8000000, 0xf8000000, 0, ExceptionFlags::empty()),
    ] { check_semantic32(x, y, relation, flags); }
    for (x, y, relation, flags) in [
        (0x31c0000000000001, 0x31c0000000000001, 0, ExceptionFlags::empty()), (0xb1c0000000000001, 0x31c0000000000001, -1, ExceptionFlags::empty()),
        (0x31c0000000000001, 0xb1c0000000000001, 1, ExceptionFlags::empty()), (0, 0x8000000000000000, 0, ExceptionFlags::empty()),
        (0x7800000000000000, 0x31c0000000000001, 1, ExceptionFlags::empty()), (0x31c0000000000001, 0x7800000000000000, -1, ExceptionFlags::empty()),
        (0x7c00000000000001, 0x31c0000000000001, 2, ExceptionFlags::empty()), (0x31c0000000000001, 0x7c00000000000001, 2, ExceptionFlags::empty()),
        (0x7e00000000000001, 0x31c0000000000001, 2, invalid), (0x31c0000000000001, 0x7e00000000000001, 2, invalid),
        (0xf800000000000000, 0x31c0000000000001, -1, ExceptionFlags::empty()), (0x31c0000000000001, 0xf800000000000000, 1, ExceptionFlags::empty()),
        (0xf800000000000000, 0x7800000000000000, -1, ExceptionFlags::empty()), (0x7800000000000000, 0xf800000000000000, 1, ExceptionFlags::empty()),
        (0x7800000000000000, 0x7800000000000000, 0, ExceptionFlags::empty()), (0xf800000000000000, 0xf800000000000000, 0, ExceptionFlags::empty()),
    ] { check_semantic64(x, y, relation, flags); }
    let one = Words { lo: 1, hi: 0x3040000000000000 };
    let neg_one = Words { lo: 1, hi: 0xb040000000000000 };
    for (x, y, relation, flags) in [
        (one, one, 0, ExceptionFlags::empty()), (neg_one, one, -1, ExceptionFlags::empty()), (one, neg_one, 1, ExceptionFlags::empty()),
        (Words { lo: 0, hi: 0 }, Words { lo: 0, hi: 0x8000000000000000 }, 0, ExceptionFlags::empty()),
        (Words { lo: 0, hi: 0x7800000000000000 }, one, 1, ExceptionFlags::empty()), (one, Words { lo: 0, hi: 0x7800000000000000 }, -1, ExceptionFlags::empty()),
        (Words { lo: 1, hi: 0x7c00000000000000 }, one, 2, ExceptionFlags::empty()), (one, Words { lo: 1, hi: 0x7c00000000000000 }, 2, ExceptionFlags::empty()),
        (Words { lo: 1, hi: 0x7e00000000000000 }, one, 2, invalid), (one, Words { lo: 1, hi: 0x7e00000000000000 }, 2, invalid),
        (Words { lo: 0, hi: 0xf800000000000000 }, one, -1, ExceptionFlags::empty()), (one, Words { lo: 0, hi: 0xf800000000000000 }, 1, ExceptionFlags::empty()),
        (Words { lo: 0, hi: 0xf800000000000000 }, Words { lo: 0, hi: 0x7800000000000000 }, -1, ExceptionFlags::empty()),
        (Words { lo: 0, hi: 0x7800000000000000 }, Words { lo: 0, hi: 0xf800000000000000 }, 1, ExceptionFlags::empty()),
        (Words { lo: 0, hi: 0x7800000000000000 }, Words { lo: 0, hi: 0x7800000000000000 }, 0, ExceptionFlags::empty()),
        (Words { lo: 0, hi: 0xf800000000000000 }, Words { lo: 0, hi: 0xf800000000000000 }, 0, ExceptionFlags::empty()),
    ] { check_semantic128(x, y, relation, flags); }
}

#[test]
fn tier1_comparison_minmax_structured_native_differential() {
    let shard = Shard::load();
    let mut count = 0u64;
    visit_pairs32(|x, y| { for &op in &QUIET_OPS { if shard.owns(count) { check_quiet32(op, x, y); } count += 1; } for &op in &MINMAX_OPS { if shard.owns(count) { check_minmax32(op, x, y); } count += 1; } });
    assert_eq!(count, COMPARE_STRUCTURED32); eprintln!("Rust Decimal32 structured compare/minmax: {}/{}", shard.owned_count(count), count);
    count = 0;
    visit_pairs64(|x, y| { for &op in &QUIET_OPS { if shard.owns(count) { check_quiet64(op, x, y); } count += 1; } for &op in &MINMAX_OPS { if shard.owns(count) { check_minmax64(op, x, y); } count += 1; } });
    assert_eq!(count, COMPARE_STRUCTURED64); eprintln!("Rust Decimal64 structured compare/minmax: {}/{}", shard.owned_count(count), count);
    count = 0;
    visit_pairs128(|x, y| { for &op in &QUIET_OPS { if shard.owns(count) { check_quiet128(op, x, y); } count += 1; } for &op in &MINMAX_OPS { if shard.owns(count) { check_minmax128(op, x, y); } count += 1; } });
    assert_eq!(count, COMPARE_STRUCTURED128); eprintln!("Rust Decimal128 structured compare/minmax: {}/{}", shard.owned_count(count), count);
}

#[test]
fn tier1_comparison_minmax_deterministic_random_native_differential() {
    let shard = Shard::load();
    let mut count = 0u64;
    for i in 0..COMPARE_RANDOM_PAIRS32 { let x = random_word(0xdec75432c04d5001, i, 0) as u32; let y = random_word(0xdec75432c04d5001, i, 1) as u32; for &op in &QUIET_OPS { if shard.owns(count) { check_quiet32(op, x, y); } count += 1; } for &op in &MINMAX_OPS { if shard.owns(count) { check_minmax32(op, x, y); } count += 1; } }
    assert_eq!(count, COMPARE_RANDOM32); eprintln!("Rust Decimal32 random compare/minmax: {}/{}", shard.owned_count(count), count);
    count = 0;
    for i in 0..COMPARE_RANDOM_PAIRS64 { let x = random_word(0xdec75464c04d5001, i, 0); let y = random_word(0xdec75464c04d5001, i, 1); for &op in &QUIET_OPS { if shard.owns(count) { check_quiet64(op, x, y); } count += 1; } for &op in &MINMAX_OPS { if shard.owns(count) { check_minmax64(op, x, y); } count += 1; } }
    assert_eq!(count, COMPARE_RANDOM64); eprintln!("Rust Decimal64 random compare/minmax: {}/{}", shard.owned_count(count), count);
    count = 0;
    for i in 0..COMPARE_RANDOM_PAIRS128 { let x = Words { lo: random_word(0xdec754c0c04d5001, i, 0), hi: random_word(0xdec754c0c04d5001, i, 1) }; let y = Words { lo: random_word(0xdec754c0c04d5001, i, 2), hi: random_word(0xdec754c0c04d5001, i, 3) }; for &op in &QUIET_OPS { if shard.owns(count) { check_quiet128(op, x, y); } count += 1; } for &op in &MINMAX_OPS { if shard.owns(count) { check_minmax128(op, x, y); } count += 1; } }
    assert_eq!(count, COMPARE_RANDOM128); eprintln!("Rust Decimal128 random compare/minmax: {}/{}", shard.owned_count(count), count);
}

fn check_convenience_contracts() {
    for &value in CONSTRUCTOR_I32 {
        let (rounded, flags) = Decimal32::from_i32(value, RoundingMode::NearestEven);
        match Decimal32::from_int(value) {
            Err(_) => assert!(flags.contains(ExceptionFlags::INEXACT), "Decimal32::from_int unexpectedly rejected exact {value}"),
            Ok(got) => { assert!(!flags.contains(ExceptionFlags::INEXACT)); assert_eq!(got.to_bits(), rounded.to_bits()); }
        }
    }
    for &value in CONSTRUCTOR_I64 {
        let (rounded, flags) = Decimal64::from_i64(value, RoundingMode::NearestEven);
        match Decimal64::from_int(value) {
            Err(_) => assert!(flags.contains(ExceptionFlags::INEXACT), "Decimal64::from_int unexpectedly rejected exact {value}"),
            Ok(got) => { assert!(!flags.contains(ExceptionFlags::INEXACT)); assert_eq!(got.to_bits(), rounded.to_bits()); }
        }
        assert_eq!(Decimal128::from_int(value).unwrap().to_le_bytes(), Decimal128::from(value).to_le_bytes());
    }
}

#[test]
fn tier1_conversion_structured_native_differential() {
    assert_eq!(BOUNDARY32.len(), BOUNDARY32_COUNT as usize); assert_eq!(BOUNDARY64.len(), BOUNDARY64_COUNT as usize); assert_eq!(BOUNDARY128.len(), BOUNDARY128_COUNT as usize);
    assert_eq!(SEMANTIC32.len(), SEMANTIC32_COUNT as usize); assert_eq!(SEMANTIC64.len(), SEMANTIC64_COUNT as usize); assert_eq!(SEMANTIC128.len(), SEMANTIC128_COUNT as usize);
    assert_eq!(CONSTRUCTOR_I32.len(), CONSTRUCTOR_I32_COUNT as usize); assert_eq!(CONSTRUCTOR_U32.len(), CONSTRUCTOR_U32_COUNT as usize); assert_eq!(CONSTRUCTOR_I64.len(), CONSTRUCTOR_I64_COUNT as usize); assert_eq!(CONSTRUCTOR_U64.len(), CONSTRUCTOR_U64_COUNT as usize);
    let shard = Shard::load(); let int_ops = to_int_ops(); let ctor_ops = constructor_ops(); assert_eq!(int_ops.len(), 80); assert_eq!(ctor_ops.len(), 36);
    let mut total = 0u64; let mut executed = 0u64;
    let mut count = 0u64;
    for &x in BOUNDARY32.iter().chain(SEMANTIC32) { for &op in &int_ops { if shard.owns(count) { check_to_int(RawDecimal::D32(x), op); } count += 1; } }
    assert_eq!(count, TO_INT_STRUCTURED32); total += count; executed += shard.owned_count(count);
    count = 0; for &x in BOUNDARY64.iter().chain(SEMANTIC64) { for &op in &int_ops { if shard.owns(count) { check_to_int(RawDecimal::D64(x), op); } count += 1; } }
    assert_eq!(count, TO_INT_STRUCTURED64); total += count; executed += shard.owned_count(count);
    count = 0; for &x in BOUNDARY128.iter().chain(SEMANTIC128) { for &op in &int_ops { if shard.owns(count) { check_to_int(RawDecimal::D128(x), op); } count += 1; } }
    assert_eq!(count, TO_INT_STRUCTURED128); total += count; executed += shard.owned_count(count);
    count = 0; for &x in BOUNDARY32.iter().chain(SEMANTIC32) { for op in width_ops(32) { if shard.owns(count) { check_width(RawDecimal::D32(x), op); } count += 1; } }
    assert_eq!(count, WIDTH_STRUCTURED32); total += count; executed += shard.owned_count(count);
    count = 0; for &x in BOUNDARY64.iter().chain(SEMANTIC64) { for op in width_ops(64) { if shard.owns(count) { check_width(RawDecimal::D64(x), op); } count += 1; } }
    assert_eq!(count, WIDTH_STRUCTURED64); total += count; executed += shard.owned_count(count);
    count = 0; for &x in BOUNDARY128.iter().chain(SEMANTIC128) { for op in width_ops(128) { if shard.owns(count) { check_width(RawDecimal::D128(x), op); } count += 1; } }
    assert_eq!(count, WIDTH_STRUCTURED128); total += count; executed += shard.owned_count(count);
    count = 0;
    for &op in &ctor_ops {
        match op.kind {
            ConstructorKind::I32 => for &x in CONSTRUCTOR_I32 { if shard.owns(count) { check_constructor(x as u32 as u64, op); } count += 1; },
            ConstructorKind::U32 => for &x in CONSTRUCTOR_U32 { if shard.owns(count) { check_constructor(x as u64, op); } count += 1; },
            ConstructorKind::I64 => for &x in CONSTRUCTOR_I64 { if shard.owns(count) { check_constructor(x as u64, op); } count += 1; },
            ConstructorKind::U64 => for &x in CONSTRUCTOR_U64 { if shard.owns(count) { check_constructor(x, op); } count += 1; },
        }
    }
    assert_eq!(count, CONSTRUCTOR_STRUCTURED); total += count; executed += shard.owned_count(count);
    check_convenience_contracts(); assert_eq!(CONSTRUCTOR_CONVENIENCE, CONSTRUCTOR_I32_COUNT + 2 * CONSTRUCTOR_I64_COUNT);
    assert_eq!(total, CONVERSION_STRUCTURED);
    eprintln!("Rust structured Tier 1 conversion exact comparisons: {executed}/{total}; convenience={CONSTRUCTOR_CONVENIENCE}");
}

#[test]
fn tier1_conversion_deterministic_random_native_differential() {
    let shard = Shard::load(); let int_ops = to_int_ops(); let ctor_ops = constructor_ops(); let mut total = 0u64; let mut executed = 0u64; let mut count = 0u64;
    for i in 0..TO_INT_RANDOM32 { if shard.owns(count) { check_to_int(RawDecimal::D32(random_word(0xdec75432c0a70001, i, 0) as u32), int_ops[i as usize % int_ops.len()]); } count += 1; }
    assert_eq!(count + TO_INT_STRUCTURED32, TO_INT_TOTAL32); total += count; executed += shard.owned_count(count);
    count = 0; for i in 0..TO_INT_RANDOM64 { if shard.owns(count) { check_to_int(RawDecimal::D64(random_word(0xdec75464c0a70001, i, 0)), int_ops[i as usize % int_ops.len()]); } count += 1; }
    assert_eq!(count + TO_INT_STRUCTURED64, TO_INT_TOTAL64); total += count; executed += shard.owned_count(count);
    count = 0; for i in 0..TO_INT_RANDOM128 { if shard.owns(count) { check_to_int(RawDecimal::D128(Words { lo: random_word(0xdec754c0c0a70001, i, 0), hi: random_word(0xdec754c0c0a70001, i, 1) }), int_ops[i as usize % int_ops.len()]); } count += 1; }
    assert_eq!(count + TO_INT_STRUCTURED128, TO_INT_TOTAL128); total += count; executed += shard.owned_count(count);
    let width32 = width_ops(32); count = 0; for i in 0..WIDTH_RANDOM32 { if shard.owns(count) { check_width(RawDecimal::D32(random_word(0xdec75432c0de0001, i, 0) as u32), width32[i as usize % width32.len()]); } count += 1; }
    assert_eq!(count + WIDTH_STRUCTURED32, WIDTH_TOTAL32); total += count; executed += shard.owned_count(count);
    let width64 = width_ops(64); count = 0; for i in 0..WIDTH_RANDOM64 { if shard.owns(count) { check_width(RawDecimal::D64(random_word(0xdec75464c0de0001, i, 0)), width64[i as usize % width64.len()]); } count += 1; }
    assert_eq!(count + WIDTH_STRUCTURED64, WIDTH_TOTAL64); total += count; executed += shard.owned_count(count);
    let width128 = width_ops(128); count = 0; for i in 0..WIDTH_RANDOM128 { if shard.owns(count) { check_width(RawDecimal::D128(Words { lo: random_word(0xdec754c0c0de0001, i, 0), hi: random_word(0xdec754c0c0de0001, i, 1) }), width128[i as usize % width128.len()]); } count += 1; }
    assert_eq!(count + WIDTH_STRUCTURED128, WIDTH_TOTAL128); total += count; executed += shard.owned_count(count);
    count = 0; for i in 0..CONSTRUCTOR_RANDOM { if shard.owns(count) { check_constructor(random_word(0xdec754c0c0570001, i, 0), ctor_ops[i as usize % ctor_ops.len()]); } count += 1; }
    assert_eq!(count + CONSTRUCTOR_STRUCTURED, CONSTRUCTOR_TOTAL); total += count; executed += shard.owned_count(count);
    assert_eq!(total, CONVERSION_RANDOM); assert_eq!(CONVERSION_STRUCTURED + total, CONVERSION_TOTAL);
    eprintln!("Rust deterministic random Tier 1 conversion exact comparisons: {executed}/{total}");
}

const PROBES32: [u32; 12] = [0x00000000, 0x80000000, 0x32800001, 0xb2800001, 0x00000001, 0x77f8967f, 0xf7f8967f, 0x78000000, 0x7c000001, 0x7e000001, 0x60000000, 0x5f800000];
const PROBES64: [u64; 12] = [0x0000000000000000, 0x8000000000000000, 0x31c0000000000001, 0xb1c0000000000001, 0x0000000000000001, 0x77fb86f26fc0ffff, 0xf7fb86f26fc0ffff, 0x7800000000000000, 0x7c00000000000001, 0x7e00000000000001, 0x6000000000000000, 0x5fe0000000000000];
const PROBES128: [Words; 12] = [
    Words { lo: 0, hi: 0 }, Words { lo: 0, hi: 0x8000000000000000 }, Words { lo: 1, hi: 0x3040000000000000 }, Words { lo: 1, hi: 0xb040000000000000 },
    Words { lo: 1, hi: 0 }, Words { lo: 0x378d8e63ffffffff, hi: 0x5fffed09bead87c0 }, Words { lo: 0x378d8e63ffffffff, hi: 0xdfffed09bead87c0 },
    Words { lo: 0, hi: 0x7800000000000000 }, Words { lo: 1, hi: 0x7c00000000000000 }, Words { lo: 1, hi: 0x7e00000000000000 },
    Words { lo: 0, hi: 0x6000000000000000 }, Words { lo: 0, hi: 0x5ffe000000000000 },
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
const SEMANTIC32: &[u32] = &[
@@TIER1_CONVERSION_SEMANTIC32_VALUES@@
];
const SEMANTIC64: &[u64] = &[
@@TIER1_CONVERSION_SEMANTIC64_VALUES@@
];
const SEMANTIC128: &[Words] = &[
@@TIER1_CONVERSION_SEMANTIC128_VALUES@@
];
const CONSTRUCTOR_I32: &[i32] = &[
@@TIER1_CONSTRUCTOR_INT32_VALUES@@
];
const CONSTRUCTOR_U32: &[u32] = &[
@@TIER1_CONSTRUCTOR_UINT32_VALUES@@
];
const CONSTRUCTOR_I64: &[i64] = &[
@@TIER1_CONSTRUCTOR_INT64_VALUES@@
];
const CONSTRUCTOR_U64: &[u64] = &[
@@TIER1_CONSTRUCTOR_UINT64_VALUES@@
];
