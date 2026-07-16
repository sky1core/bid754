//! Regression coverage for the exact generated-Rust calls timed by the
//! Criterion benchmark target.

use std::collections::{BTreeMap, BTreeSet};
use std::ffi::{c_char, c_int, CStr, CString};
use std::hint::black_box;

use bid754::bid64_from_string_raw;
use bid754::gen_types::BID_UINT128 as Rust128;
use bid754::generated::prelude::*;
use libbid_sys::BID_UINT128 as C128;

include!("../../benches/support/inputs.rs");
include!("../../benches/support/rows.rs");

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
    #[link_name = "__bid32_scalbn"]
    fn native_bid32_scalbn(x: u32, exponent: c_int, rounding: u32, flags: *mut u32) -> u32;
    #[link_name = "__bid64_scalbn"]
    fn native_bid64_scalbn(x: u64, exponent: c_int, rounding: u32, flags: *mut u32) -> u64;
    #[link_name = "__bid128_scalbn"]
    fn native_bid128_scalbn(x: C128, exponent: c_int, rounding: u32, flags: *mut u32) -> C128;
    #[link_name = "__bid32_to_int64_rnint"]
    fn native_bid32_to_int64_rnint(x: u32, flags: *mut u32) -> i64;
    #[link_name = "__bid64_to_int64_rnint"]
    fn native_bid64_to_int64_rnint(x: u64, flags: *mut u32) -> i64;
    #[link_name = "__bid128_to_int64_rnint"]
    fn native_bid128_to_int64_rnint(x: C128, flags: *mut u32) -> i64;

    #[link_name = "__bid64dq_add"]
    fn native_bid64dq_add(x: u64, y: C128, rounding: u32, flags: *mut u32) -> u64;
    #[link_name = "__bid64qd_add"]
    fn native_bid64qd_add(x: C128, y: u64, rounding: u32, flags: *mut u32) -> u64;
    #[link_name = "__bid64qq_add"]
    fn native_bid64qq_add(x: C128, y: C128, rounding: u32, flags: *mut u32) -> u64;
    #[link_name = "__bid64dq_sub"]
    fn native_bid64dq_sub(x: u64, y: C128, rounding: u32, flags: *mut u32) -> u64;
    #[link_name = "__bid64qd_sub"]
    fn native_bid64qd_sub(x: C128, y: u64, rounding: u32, flags: *mut u32) -> u64;
    #[link_name = "__bid64qq_sub"]
    fn native_bid64qq_sub(x: C128, y: C128, rounding: u32, flags: *mut u32) -> u64;
    #[link_name = "__bid64dq_mul"]
    fn native_bid64dq_mul(x: u64, y: C128, rounding: u32, flags: *mut u32) -> u64;
    #[link_name = "__bid64qd_mul"]
    fn native_bid64qd_mul(x: C128, y: u64, rounding: u32, flags: *mut u32) -> u64;
    #[link_name = "__bid64qq_mul"]
    fn native_bid64qq_mul(x: C128, y: C128, rounding: u32, flags: *mut u32) -> u64;
    #[link_name = "__bid64dq_div"]
    fn native_bid64dq_div(x: u64, y: C128, rounding: u32, flags: *mut u32) -> u64;
    #[link_name = "__bid64qd_div"]
    fn native_bid64qd_div(x: C128, y: u64, rounding: u32, flags: *mut u32) -> u64;
    #[link_name = "__bid64qq_div"]
    fn native_bid64qq_div(x: C128, y: C128, rounding: u32, flags: *mut u32) -> u64;

    #[link_name = "__bid128dd_add"]
    fn native_bid128dd_add(x: u64, y: u64, rounding: u32, flags: *mut u32) -> C128;
    #[link_name = "__bid128dq_add"]
    fn native_bid128dq_add(x: u64, y: C128, rounding: u32, flags: *mut u32) -> C128;
    #[link_name = "__bid128qd_add"]
    fn native_bid128qd_add(x: C128, y: u64, rounding: u32, flags: *mut u32) -> C128;
    #[link_name = "__bid128dd_sub"]
    fn native_bid128dd_sub(x: u64, y: u64, rounding: u32, flags: *mut u32) -> C128;
    #[link_name = "__bid128dq_sub"]
    fn native_bid128dq_sub(x: u64, y: C128, rounding: u32, flags: *mut u32) -> C128;
    #[link_name = "__bid128qd_sub"]
    fn native_bid128qd_sub(x: C128, y: u64, rounding: u32, flags: *mut u32) -> C128;
    #[link_name = "__bid128dd_mul"]
    fn native_bid128dd_mul(x: u64, y: u64, rounding: u32, flags: *mut u32) -> C128;
    #[link_name = "__bid128dq_mul"]
    fn native_bid128dq_mul(x: u64, y: C128, rounding: u32, flags: *mut u32) -> C128;
    #[link_name = "__bid128qd_mul"]
    fn native_bid128qd_mul(x: C128, y: u64, rounding: u32, flags: *mut u32) -> C128;
    #[link_name = "__bid128dd_div"]
    fn native_bid128dd_div(x: u64, y: u64, rounding: u32, flags: *mut u32) -> C128;
    #[link_name = "__bid128dq_div"]
    fn native_bid128dq_div(x: u64, y: C128, rounding: u32, flags: *mut u32) -> C128;
    #[link_name = "__bid128qd_div"]
    fn native_bid128qd_div(x: C128, y: u64, rounding: u32, flags: *mut u32) -> C128;
}

fn c128(value: Rust128) -> C128 {
    C128::new(value.lo, value.hi)
}

fn c_scale_exponent(value: i64) -> c_int {
    c_int::try_from(value).expect("benchmark scale exponent must fit the Intel C int32 contract")
}

#[derive(Clone, Copy, Debug, Eq, PartialEq)]
enum ValueKind {
    D32,
    D64,
    D128,
    I64,
    Predicate,
    Text,
}

#[derive(Clone, Copy, Debug, Eq, PartialEq)]
enum OperandShape {
    One(ValueKind),
    Two(ValueKind, ValueKind),
    Three(ValueKind, ValueKind, ValueKind),
}

#[derive(Clone, Copy, Debug, Eq, PartialEq)]
enum RoundingContract {
    ExplicitNearestEven,
    FixedNearestEven,
    NotApplicable,
}

#[derive(Clone, Copy, Debug, Eq, PartialEq)]
enum StatusContract {
    TupleFlags,
    OutFlags,
    NoFlags,
}

#[derive(Clone, Copy, Debug, Eq, PartialEq)]
struct Contract {
    group: &'static str,
    name: &'static str,
    operands: OperandShape,
    result: ValueKind,
    rounding: RoundingContract,
    status: StatusContract,
}

macro_rules! value_kind {
    (D32) => {
        ValueKind::D32
    };
    (D64) => {
        ValueKind::D64
    };
    (D128) => {
        ValueKind::D128
    };
    (I64) => {
        ValueKind::I64
    };
    (Predicate) => {
        ValueKind::Predicate
    };
    (Text) => {
        ValueKind::Text
    };
}

macro_rules! operand_shape {
    ([$a:ident]) => {
        OperandShape::One(value_kind!($a))
    };
    ([$a:ident, $b:ident]) => {
        OperandShape::Two(value_kind!($a), value_kind!($b))
    };
    ([$a:ident, $b:ident, $c:ident]) => {
        OperandShape::Three(value_kind!($a), value_kind!($b), value_kind!($c))
    };
}

macro_rules! rounding_contract {
    (ExplicitNearestEven) => {
        RoundingContract::ExplicitNearestEven
    };
    (FixedNearestEven) => {
        RoundingContract::FixedNearestEven
    };
    (NotApplicable) => {
        RoundingContract::NotApplicable
    };
}

macro_rules! status_contract {
    (TupleFlags) => {
        StatusContract::TupleFlags
    };
    (OutFlags) => {
        StatusContract::OutFlags
    };
    (NoFlags) => {
        StatusContract::NoFlags
    };
}

#[derive(Clone, Debug, Eq, PartialEq)]
enum Observation {
    D32(u32, Option<u32>),
    D64(u64, Option<u32>),
    D128(u64, u64, Option<u32>),
    I64(i64, Option<u32>),
    Predicate(i64, Option<u32>),
    Text(String),
}

macro_rules! observe {
    (D32, $status:ident, $call:expr) => {{
        let (bits, flags) = $call;
        Observation::D32(bits, Some(flags))
    }};
    (D64, $status:ident, $call:expr) => {{
        let (bits, flags) = $call;
        Observation::D64(bits, Some(flags))
    }};
    (D128, NoFlags, $call:expr) => {{
        let bits = $call;
        Observation::D128(bits.lo, bits.hi, None)
    }};
    (D128, $status:ident, $call:expr) => {{
        let (bits, flags) = $call;
        Observation::D128(bits.lo, bits.hi, Some(flags))
    }};
    (I64, $status:ident, $call:expr) => {{
        let (value, flags) = $call;
        Observation::I64(value, Some(flags))
    }};
    (Predicate, $status:ident, $call:expr) => {{
        let (value, flags) = $call;
        Observation::Predicate(value, Some(flags))
    }};
    (Text, NoFlags, $call:expr) => {
        Observation::Text($call)
    };
}

#[derive(Clone, Copy, Debug, Eq, Ord, PartialEq, PartialOrd)]
enum RowId {
    Bid32Add,
    Bid32Mul,
    Bid32Sub,
    Bid32Div,
    Bid32Fma,
    Bid32Sqrt,
    Bid32Remainder,
    Bid32Fmod,
    Bid32Quantize,
    Bid32Scaleb,
    Bid32QuietEqual,
    Bid32Minnum,
    Bid32Maxnum,
    Bid32FromInt64,
    Bid32ToInt64,
    Bid32ToDecimal64,
    Bid32ToDecimal128,
    Bid32Parse,
    Bid32ToString,
    Bid64Add,
    Bid64Mul,
    Bid64Sub,
    Bid64Div,
    Bid64Fma,
    Bid64Sqrt,
    Bid64Remainder,
    Bid64Fmod,
    Bid64Quantize,
    Bid64Scaleb,
    Bid64QuietEqual,
    Bid64Minnum,
    Bid64Maxnum,
    Bid64FromInt64,
    Bid64ToInt64,
    Bid64ToDecimal32,
    Bid64ToDecimal128,
    Bid64Parse,
    Bid64ToString,
    Bid128Add,
    Bid128Mul,
    Bid128Sub,
    Bid128Div,
    Bid128Fma,
    Bid128Sqrt,
    Bid128Remainder,
    Bid128Fmod,
    Bid128Quantize,
    Bid128Scaleb,
    Bid128QuietEqual,
    Bid128Minnum,
    Bid128Maxnum,
    Bid128FromInt64,
    Bid128ToInt64,
    Bid128ToDecimal32,
    Bid128ToDecimal64,
    Bid128Parse,
    Bid128ToString,
    Bid64MixedDqAdd,
    Bid64MixedQdAdd,
    Bid64MixedQqAdd,
    Bid64MixedDqSub,
    Bid64MixedQdSub,
    Bid64MixedQqSub,
    Bid64MixedDqMul,
    Bid64MixedQdMul,
    Bid64MixedQqMul,
    Bid64MixedDqDiv,
    Bid64MixedQdDiv,
    Bid64MixedQqDiv,
    Bid128MixedDdAdd,
    Bid128MixedDqAdd,
    Bid128MixedQdAdd,
    Bid128MixedDdSub,
    Bid128MixedDqSub,
    Bid128MixedQdSub,
    Bid128MixedDdMul,
    Bid128MixedDqMul,
    Bid128MixedQdMul,
    Bid128MixedDdDiv,
    Bid128MixedDqDiv,
    Bid128MixedQdDiv,
}

fn contract(
    group: &'static str,
    name: &'static str,
    operands: OperandShape,
    result: ValueKind,
    rounding: RoundingContract,
    status: StatusContract,
) -> Contract {
    Contract {
        group,
        name,
        operands,
        result,
        rounding,
        status,
    }
}

macro_rules! c {
    ($group:literal, $name:literal, [$($operand:ident),+], $result:ident, $rounding:ident, $status:ident) => {
        contract(
            $group,
            $name,
            operand_shape!([$($operand),+]),
            value_kind!($result),
            rounding_contract!($rounding),
            status_contract!($status),
        )
    };
}

fn expected_contract(id: RowId) -> Contract {
    use RowId::*;
    match id {
        Bid32Add => c!(
            "bid32",
            "add",
            [D32, D32],
            D32,
            ExplicitNearestEven,
            TupleFlags
        ),
        Bid32Mul => c!(
            "bid32",
            "mul",
            [D32, D32],
            D32,
            ExplicitNearestEven,
            TupleFlags
        ),
        Bid32Sub => c!(
            "bid32",
            "sub",
            [D32, D32],
            D32,
            ExplicitNearestEven,
            TupleFlags
        ),
        Bid32Div => c!(
            "bid32",
            "div",
            [D32, D32],
            D32,
            ExplicitNearestEven,
            TupleFlags
        ),
        Bid32Fma => c!(
            "bid32",
            "fma",
            [D32, D32, D32],
            D32,
            ExplicitNearestEven,
            TupleFlags
        ),
        Bid32Sqrt => c!("bid32", "sqrt", [D32], D32, ExplicitNearestEven, TupleFlags),
        Bid32Remainder => c!(
            "bid32",
            "remainder",
            [D32, D32],
            D32,
            NotApplicable,
            TupleFlags
        ),
        Bid32Fmod => c!("bid32", "fmod", [D32, D32], D32, NotApplicable, TupleFlags),
        Bid32Quantize => c!(
            "bid32",
            "quantize",
            [D32, D32],
            D32,
            ExplicitNearestEven,
            TupleFlags
        ),
        Bid32Scaleb => c!(
            "bid32",
            "scaleb",
            [D32, I64],
            D32,
            ExplicitNearestEven,
            TupleFlags
        ),
        Bid32QuietEqual => c!(
            "bid32",
            "quiet_equal",
            [D32, D32],
            Predicate,
            NotApplicable,
            TupleFlags
        ),
        Bid32Minnum => c!(
            "bid32",
            "minnum",
            [D32, D32],
            D32,
            NotApplicable,
            TupleFlags
        ),
        Bid32Maxnum => c!(
            "bid32",
            "maxnum",
            [D32, D32],
            D32,
            NotApplicable,
            TupleFlags
        ),
        Bid32FromInt64 => c!(
            "bid32",
            "from_int64",
            [I64],
            D32,
            ExplicitNearestEven,
            TupleFlags
        ),
        Bid32ToInt64 => c!(
            "bid32",
            "to_int64",
            [D32],
            I64,
            FixedNearestEven,
            TupleFlags
        ),
        Bid32ToDecimal64 => c!(
            "bid32",
            "to_decimal64",
            [D32],
            D64,
            NotApplicable,
            TupleFlags
        ),
        Bid32ToDecimal128 => c!(
            "bid32",
            "to_decimal128",
            [D32],
            D128,
            NotApplicable,
            TupleFlags
        ),
        Bid32Parse => c!(
            "bid32",
            "parse",
            [Text],
            D32,
            ExplicitNearestEven,
            TupleFlags
        ),
        Bid32ToString => c!("bid32", "to_string", [D32], Text, NotApplicable, NoFlags),
        Bid64Add => c!(
            "bid64",
            "add",
            [D64, D64],
            D64,
            ExplicitNearestEven,
            TupleFlags
        ),
        Bid64Mul => c!(
            "bid64",
            "mul",
            [D64, D64],
            D64,
            ExplicitNearestEven,
            TupleFlags
        ),
        Bid64Sub => c!(
            "bid64",
            "sub",
            [D64, D64],
            D64,
            ExplicitNearestEven,
            TupleFlags
        ),
        Bid64Div => c!(
            "bid64",
            "div",
            [D64, D64],
            D64,
            ExplicitNearestEven,
            TupleFlags
        ),
        Bid64Fma => c!(
            "bid64",
            "fma",
            [D64, D64, D64],
            D64,
            ExplicitNearestEven,
            TupleFlags
        ),
        Bid64Sqrt => c!("bid64", "sqrt", [D64], D64, ExplicitNearestEven, TupleFlags),
        Bid64Remainder => c!(
            "bid64",
            "remainder",
            [D64, D64],
            D64,
            NotApplicable,
            TupleFlags
        ),
        Bid64Fmod => c!("bid64", "fmod", [D64, D64], D64, NotApplicable, TupleFlags),
        Bid64Quantize => c!(
            "bid64",
            "quantize",
            [D64, D64],
            D64,
            ExplicitNearestEven,
            TupleFlags
        ),
        Bid64Scaleb => c!(
            "bid64",
            "scaleb",
            [D64, I64],
            D64,
            ExplicitNearestEven,
            TupleFlags
        ),
        Bid64QuietEqual => c!(
            "bid64",
            "quiet_equal",
            [D64, D64],
            Predicate,
            NotApplicable,
            TupleFlags
        ),
        Bid64Minnum => c!(
            "bid64",
            "minnum",
            [D64, D64],
            D64,
            NotApplicable,
            TupleFlags
        ),
        Bid64Maxnum => c!(
            "bid64",
            "maxnum",
            [D64, D64],
            D64,
            NotApplicable,
            TupleFlags
        ),
        Bid64FromInt64 => c!(
            "bid64",
            "from_int64",
            [I64],
            D64,
            ExplicitNearestEven,
            TupleFlags
        ),
        Bid64ToInt64 => c!(
            "bid64",
            "to_int64",
            [D64],
            I64,
            FixedNearestEven,
            TupleFlags
        ),
        Bid64ToDecimal32 => c!(
            "bid64",
            "to_decimal32",
            [D64],
            D32,
            ExplicitNearestEven,
            TupleFlags
        ),
        Bid64ToDecimal128 => c!(
            "bid64",
            "to_decimal128",
            [D64],
            D128,
            NotApplicable,
            TupleFlags
        ),
        Bid64Parse => c!(
            "bid64",
            "parse",
            [Text],
            D64,
            ExplicitNearestEven,
            TupleFlags
        ),
        Bid64ToString => c!("bid64", "to_string", [D64], Text, NotApplicable, NoFlags),
        Bid128Add => c!(
            "bid128",
            "add",
            [D128, D128],
            D128,
            ExplicitNearestEven,
            OutFlags
        ),
        Bid128Mul => c!(
            "bid128",
            "mul",
            [D128, D128],
            D128,
            ExplicitNearestEven,
            TupleFlags
        ),
        Bid128Sub => c!(
            "bid128",
            "sub",
            [D128, D128],
            D128,
            ExplicitNearestEven,
            OutFlags
        ),
        Bid128Div => c!(
            "bid128",
            "div",
            [D128, D128],
            D128,
            ExplicitNearestEven,
            TupleFlags
        ),
        Bid128Fma => c!(
            "bid128",
            "fma",
            [D128, D128, D128],
            D128,
            ExplicitNearestEven,
            TupleFlags
        ),
        Bid128Sqrt => c!(
            "bid128",
            "sqrt",
            [D128],
            D128,
            ExplicitNearestEven,
            TupleFlags
        ),
        Bid128Remainder => c!(
            "bid128",
            "remainder",
            [D128, D128],
            D128,
            NotApplicable,
            TupleFlags
        ),
        Bid128Fmod => c!(
            "bid128",
            "fmod",
            [D128, D128],
            D128,
            NotApplicable,
            TupleFlags
        ),
        Bid128Quantize => c!(
            "bid128",
            "quantize",
            [D128, D128],
            D128,
            ExplicitNearestEven,
            TupleFlags
        ),
        Bid128Scaleb => c!(
            "bid128",
            "scaleb",
            [D128, I64],
            D128,
            ExplicitNearestEven,
            OutFlags
        ),
        Bid128QuietEqual => c!(
            "bid128",
            "quiet_equal",
            [D128, D128],
            Predicate,
            NotApplicable,
            TupleFlags
        ),
        Bid128Minnum => c!(
            "bid128",
            "minnum",
            [D128, D128],
            D128,
            NotApplicable,
            OutFlags
        ),
        Bid128Maxnum => c!(
            "bid128",
            "maxnum",
            [D128, D128],
            D128,
            NotApplicable,
            OutFlags
        ),
        Bid128FromInt64 => c!("bid128", "from_int64", [I64], D128, NotApplicable, NoFlags),
        Bid128ToInt64 => c!(
            "bid128",
            "to_int64",
            [D128],
            I64,
            FixedNearestEven,
            TupleFlags
        ),
        Bid128ToDecimal32 => c!(
            "bid128",
            "to_decimal32",
            [D128],
            D32,
            ExplicitNearestEven,
            TupleFlags
        ),
        Bid128ToDecimal64 => c!(
            "bid128",
            "to_decimal64",
            [D128],
            D64,
            ExplicitNearestEven,
            TupleFlags
        ),
        Bid128Parse => c!(
            "bid128",
            "parse",
            [Text],
            D128,
            ExplicitNearestEven,
            TupleFlags
        ),
        Bid128ToString => c!("bid128", "to_string", [D128], Text, NotApplicable, NoFlags),
        Bid64MixedDqAdd => c!(
            "bid64_mixed",
            "dq_add",
            [D64, D128],
            D64,
            ExplicitNearestEven,
            TupleFlags
        ),
        Bid64MixedQdAdd => c!(
            "bid64_mixed",
            "qd_add",
            [D128, D64],
            D64,
            ExplicitNearestEven,
            TupleFlags
        ),
        Bid64MixedQqAdd => c!(
            "bid64_mixed",
            "qq_add",
            [D128, D128],
            D64,
            ExplicitNearestEven,
            TupleFlags
        ),
        Bid64MixedDqSub => c!(
            "bid64_mixed",
            "dq_sub",
            [D64, D128],
            D64,
            ExplicitNearestEven,
            TupleFlags
        ),
        Bid64MixedQdSub => c!(
            "bid64_mixed",
            "qd_sub",
            [D128, D64],
            D64,
            ExplicitNearestEven,
            TupleFlags
        ),
        Bid64MixedQqSub => c!(
            "bid64_mixed",
            "qq_sub",
            [D128, D128],
            D64,
            ExplicitNearestEven,
            TupleFlags
        ),
        Bid64MixedDqMul => c!(
            "bid64_mixed",
            "dq_mul",
            [D64, D128],
            D64,
            ExplicitNearestEven,
            TupleFlags
        ),
        Bid64MixedQdMul => c!(
            "bid64_mixed",
            "qd_mul",
            [D128, D64],
            D64,
            ExplicitNearestEven,
            TupleFlags
        ),
        Bid64MixedQqMul => c!(
            "bid64_mixed",
            "qq_mul",
            [D128, D128],
            D64,
            ExplicitNearestEven,
            TupleFlags
        ),
        Bid64MixedDqDiv => c!(
            "bid64_mixed",
            "dq_div",
            [D64, D128],
            D64,
            ExplicitNearestEven,
            TupleFlags
        ),
        Bid64MixedQdDiv => c!(
            "bid64_mixed",
            "qd_div",
            [D128, D64],
            D64,
            ExplicitNearestEven,
            TupleFlags
        ),
        Bid64MixedQqDiv => c!(
            "bid64_mixed",
            "qq_div",
            [D128, D128],
            D64,
            ExplicitNearestEven,
            TupleFlags
        ),
        Bid128MixedDdAdd => c!(
            "bid128_mixed",
            "dd_add",
            [D64, D64],
            D128,
            ExplicitNearestEven,
            TupleFlags
        ),
        Bid128MixedDqAdd => c!(
            "bid128_mixed",
            "dq_add",
            [D64, D128],
            D128,
            ExplicitNearestEven,
            TupleFlags
        ),
        Bid128MixedQdAdd => c!(
            "bid128_mixed",
            "qd_add",
            [D128, D64],
            D128,
            ExplicitNearestEven,
            TupleFlags
        ),
        Bid128MixedDdSub => c!(
            "bid128_mixed",
            "dd_sub",
            [D64, D64],
            D128,
            ExplicitNearestEven,
            TupleFlags
        ),
        Bid128MixedDqSub => c!(
            "bid128_mixed",
            "dq_sub",
            [D64, D128],
            D128,
            ExplicitNearestEven,
            TupleFlags
        ),
        Bid128MixedQdSub => c!(
            "bid128_mixed",
            "qd_sub",
            [D128, D64],
            D128,
            ExplicitNearestEven,
            TupleFlags
        ),
        Bid128MixedDdMul => c!(
            "bid128_mixed",
            "dd_mul",
            [D64, D64],
            D128,
            ExplicitNearestEven,
            TupleFlags
        ),
        Bid128MixedDqMul => c!(
            "bid128_mixed",
            "dq_mul",
            [D64, D128],
            D128,
            ExplicitNearestEven,
            TupleFlags
        ),
        Bid128MixedQdMul => c!(
            "bid128_mixed",
            "qd_mul",
            [D128, D64],
            D128,
            ExplicitNearestEven,
            TupleFlags
        ),
        Bid128MixedDdDiv => c!(
            "bid128_mixed",
            "dd_div",
            [D64, D64],
            D128,
            ExplicitNearestEven,
            TupleFlags
        ),
        Bid128MixedDqDiv => c!(
            "bid128_mixed",
            "dq_div",
            [D64, D128],
            D128,
            ExplicitNearestEven,
            TupleFlags
        ),
        Bid128MixedQdDiv => c!(
            "bid128_mixed",
            "qd_div",
            [D128, D64],
            D128,
            ExplicitNearestEven,
            TupleFlags
        ),
    }
}

macro_rules! native_d32 {
    ($function:path; $($argument:expr),*) => {{
        let mut flags = 0;
        let bits = unsafe { $function($($argument,)* &mut flags) };
        Observation::D32(bits, Some(flags))
    }};
}

macro_rules! native_d64 {
    ($function:path; $($argument:expr),*) => {{
        let mut flags = 0;
        let bits = unsafe { $function($($argument,)* &mut flags) };
        Observation::D64(bits, Some(flags))
    }};
}

macro_rules! native_d128 {
    ($function:path; $($argument:expr),*) => {{
        let mut flags = 0;
        let bits = unsafe { $function($($argument,)* &mut flags) };
        Observation::D128(bits.w[0], bits.w[1], Some(flags))
    }};
}

macro_rules! native_i64 {
    ($function:path; $($argument:expr),*) => {{
        let mut flags = 0;
        let value = unsafe { $function($($argument,)* &mut flags) };
        Observation::I64(value, Some(flags))
    }};
}

macro_rules! native_predicate {
    ($function:path; $($argument:expr),*) => {{
        let mut flags = 0;
        let value = unsafe { $function($($argument,)* &mut flags) };
        Observation::Predicate(i64::from(value), Some(flags))
    }};
}

fn native_parse32(text: &str) -> Observation {
    let text = CString::new(text).unwrap();
    native_d32!(libbid_sys::bid32_from_string; text.as_ptr(), 0)
}

fn native_parse64(text: &str) -> Observation {
    let text = CString::new(text).unwrap();
    native_d64!(libbid_sys::bid64_from_string; text.as_ptr(), 0)
}

fn native_parse128(text: &str) -> Observation {
    let text = CString::new(text).unwrap();
    native_d128!(libbid_sys::bid128_from_string; text.as_ptr(), 0)
}

fn native_string(call: impl FnOnce(*mut c_char, *mut u32)) -> Observation {
    let mut buffer = [0 as c_char; 128];
    let mut flags = 0;
    call(buffer.as_mut_ptr(), &mut flags);
    assert_eq!(flags, 0, "Intel BID to-string status flags");
    let text = unsafe { CStr::from_ptr(buffer.as_ptr()) }
        .to_str()
        .expect("Intel BID to-string output must be UTF-8")
        .to_owned();
    Observation::Text(text)
}

fn native_string32(value: u32) -> Observation {
    native_string(|buffer, flags| unsafe {
        libbid_sys::bid32_to_string(buffer, value, flags);
    })
}

fn native_string64(value: u64) -> Observation {
    native_string(|buffer, flags| unsafe {
        libbid_sys::bid64_to_string(buffer, value, flags);
    })
}

fn native_string128(value: Rust128) -> Observation {
    native_string(|buffer, flags| unsafe {
        libbid_sys::bid128_to_string(buffer, c128(value), flags);
    })
}

fn native_expected(id: RowId, i: &PreparedBenchmarkInputs) -> Observation {
    use RowId::*;
    match id {
        Bid32Add => native_d32!(libbid_sys::bid32_add; i.x32, i.y32, 0),
        Bid32Mul => native_d32!(libbid_sys::bid32_mul; i.x32, i.y32, 0),
        Bid32Sub => native_d32!(libbid_sys::bid32_sub; i.x32, i.y32, 0),
        Bid32Div => native_d32!(libbid_sys::bid32_div; i.x32, i.y32, 0),
        Bid32Fma => native_d32!(libbid_sys::bid32_fma; i.x32, i.y32, i.z32, 0),
        Bid32Sqrt => native_d32!(libbid_sys::bid32_sqrt; i.x32, 0),
        Bid32Remainder => native_d32!(native_bid32_rem; i.y32, i.x32),
        Bid32Fmod => native_d32!(native_bid32_fmod; i.y32, i.x32),
        Bid32Quantize => native_d32!(libbid_sys::bid32_quantize; i.x32, i.integer32, 0),
        Bid32Scaleb => {
            native_d32!(native_bid32_scalbn; i.x32, c_scale_exponent(i.scale_exponent), 0)
        }
        Bid32QuietEqual => native_predicate!(libbid_sys::bid32_quiet_equal; i.x32, i.y32),
        Bid32Minnum => native_d32!(libbid_sys::bid32_minnum; i.x32, i.y32),
        Bid32Maxnum => native_d32!(libbid_sys::bid32_maxnum; i.x32, i.y32),
        Bid32FromInt64 => native_d32!(libbid_sys::bid32_from_int64; i.integer_operand, 0),
        Bid32ToInt64 => native_i64!(native_bid32_to_int64_rnint; i.integer32),
        Bid32ToDecimal64 => native_d64!(libbid_sys::bid32_to_bid64; i.x32),
        Bid32ToDecimal128 => native_d128!(libbid_sys::bid32_to_bid128; i.x32),
        Bid32Parse => native_parse32(&i.decimal32_x_text),
        Bid32ToString => native_string32(i.x32),
        Bid64Add => native_d64!(libbid_sys::bid64_add; i.x64, i.y64, 0),
        Bid64Mul => native_d64!(libbid_sys::bid64_mul; i.x64, i.y64, 0),
        Bid64Sub => native_d64!(libbid_sys::bid64_sub; i.x64, i.y64, 0),
        Bid64Div => native_d64!(libbid_sys::bid64_div; i.x64, i.y64, 0),
        Bid64Fma => native_d64!(libbid_sys::bid64_fma; i.x64, i.y64, i.z64, 0),
        Bid64Sqrt => native_d64!(libbid_sys::bid64_sqrt; i.x64, 0),
        Bid64Remainder => native_d64!(native_bid64_rem; i.y64, i.x64),
        Bid64Fmod => native_d64!(native_bid64_fmod; i.y64, i.x64),
        Bid64Quantize => native_d64!(libbid_sys::bid64_quantize; i.x64, i.integer64, 0),
        Bid64Scaleb => {
            native_d64!(native_bid64_scalbn; i.x64, c_scale_exponent(i.scale_exponent), 0)
        }
        Bid64QuietEqual => native_predicate!(libbid_sys::bid64_quiet_equal; i.x64, i.y64),
        Bid64Minnum => native_d64!(libbid_sys::bid64_minnum; i.x64, i.y64),
        Bid64Maxnum => native_d64!(libbid_sys::bid64_maxnum; i.x64, i.y64),
        Bid64FromInt64 => native_d64!(libbid_sys::bid64_from_int64; i.integer_operand, 0),
        Bid64ToInt64 => native_i64!(native_bid64_to_int64_rnint; i.integer64),
        Bid64ToDecimal32 => native_d32!(libbid_sys::bid64_to_bid32; i.x64, 0),
        Bid64ToDecimal128 => native_d128!(libbid_sys::bid64_to_bid128; i.x64),
        Bid64Parse => native_parse64(&i.decimal64_x_text),
        Bid64ToString => native_string64(i.x64),
        Bid128Add => native_d128!(libbid_sys::bid128_add; c128(i.x128), c128(i.y128), 0),
        Bid128Mul => native_d128!(libbid_sys::bid128_mul; c128(i.x128), c128(i.y128), 0),
        Bid128Sub => native_d128!(libbid_sys::bid128_sub; c128(i.x128), c128(i.y128), 0),
        Bid128Div => native_d128!(libbid_sys::bid128_div; c128(i.x128), c128(i.y128), 0),
        Bid128Fma => native_d128!(
            libbid_sys::bid128_fma;
            c128(i.x128),
            c128(i.y128),
            c128(i.z128),
            0
        ),
        Bid128Sqrt => native_d128!(libbid_sys::bid128_sqrt; c128(i.x128), 0),
        Bid128Remainder => native_d128!(native_bid128_rem; c128(i.y128), c128(i.x128)),
        Bid128Fmod => native_d128!(native_bid128_fmod; c128(i.y128), c128(i.x128)),
        Bid128Quantize => native_d128!(
            libbid_sys::bid128_quantize;
            c128(i.x128),
            c128(i.integer128),
            0
        ),
        Bid128Scaleb => native_d128!(
            native_bid128_scalbn;
            c128(i.x128),
            c_scale_exponent(i.scale_exponent),
            0
        ),
        Bid128QuietEqual => native_predicate!(
            libbid_sys::bid128_quiet_equal;
            c128(i.x128),
            c128(i.y128)
        ),
        Bid128Minnum => native_d128!(libbid_sys::bid128_minnum; c128(i.x128), c128(i.y128)),
        Bid128Maxnum => native_d128!(libbid_sys::bid128_maxnum; c128(i.x128), c128(i.y128)),
        Bid128FromInt64 => {
            let bits = unsafe { libbid_sys::bid128_from_int64(i.integer_operand) };
            Observation::D128(bits.w[0], bits.w[1], None)
        }
        Bid128ToInt64 => native_i64!(native_bid128_to_int64_rnint; c128(i.integer128)),
        Bid128ToDecimal32 => native_d32!(libbid_sys::bid128_to_bid32; c128(i.x128), 0),
        Bid128ToDecimal64 => native_d64!(libbid_sys::bid128_to_bid64; c128(i.x128), 0),
        Bid128Parse => native_parse128(&i.decimal128_x_text),
        Bid128ToString => native_string128(i.x128),
        Bid64MixedDqAdd => native_d64!(native_bid64dq_add; i.x64, c128(i.y128), 0),
        Bid64MixedQdAdd => native_d64!(native_bid64qd_add; c128(i.x128), i.y64, 0),
        Bid64MixedQqAdd => native_d64!(native_bid64qq_add; c128(i.x128), c128(i.y128), 0),
        Bid64MixedDqSub => native_d64!(native_bid64dq_sub; i.x64, c128(i.y128), 0),
        Bid64MixedQdSub => native_d64!(native_bid64qd_sub; c128(i.x128), i.y64, 0),
        Bid64MixedQqSub => native_d64!(native_bid64qq_sub; c128(i.x128), c128(i.y128), 0),
        Bid64MixedDqMul => native_d64!(native_bid64dq_mul; i.x64, c128(i.y128), 0),
        Bid64MixedQdMul => native_d64!(native_bid64qd_mul; c128(i.x128), i.y64, 0),
        Bid64MixedQqMul => native_d64!(native_bid64qq_mul; c128(i.x128), c128(i.y128), 0),
        Bid64MixedDqDiv => native_d64!(native_bid64dq_div; i.x64, c128(i.y128), 0),
        Bid64MixedQdDiv => native_d64!(native_bid64qd_div; c128(i.x128), i.y64, 0),
        Bid64MixedQqDiv => native_d64!(native_bid64qq_div; c128(i.x128), c128(i.y128), 0),
        Bid128MixedDdAdd => native_d128!(native_bid128dd_add; i.x64, i.y64, 0),
        Bid128MixedDqAdd => native_d128!(native_bid128dq_add; i.x64, c128(i.y128), 0),
        Bid128MixedQdAdd => native_d128!(native_bid128qd_add; c128(i.x128), i.y64, 0),
        Bid128MixedDdSub => native_d128!(native_bid128dd_sub; i.x64, i.y64, 0),
        Bid128MixedDqSub => native_d128!(native_bid128dq_sub; i.x64, c128(i.y128), 0),
        Bid128MixedQdSub => native_d128!(native_bid128qd_sub; c128(i.x128), i.y64, 0),
        Bid128MixedDdMul => native_d128!(native_bid128dd_mul; i.x64, i.y64, 0),
        Bid128MixedDqMul => native_d128!(native_bid128dq_mul; i.x64, c128(i.y128), 0),
        Bid128MixedQdMul => native_d128!(native_bid128qd_mul; c128(i.x128), i.y64, 0),
        Bid128MixedDdDiv => native_d128!(native_bid128dd_div; i.x64, i.y64, 0),
        Bid128MixedDqDiv => native_d128!(native_bid128dq_div; i.x64, c128(i.y128), 0),
        Bid128MixedQdDiv => native_d128!(native_bid128qd_div; c128(i.x128), i.y64, 0),
    }
}

fn assert_text_round_trip(id: RowId, actual: &Observation, i: &PreparedBenchmarkInputs) {
    match (id, actual) {
        (RowId::Bid32ToString, Observation::Text(text)) => assert_eq!(
            native_parse32(text),
            Observation::D32(i.x32, Some(0)),
            "generated Rust bid32_to_string must round-trip through Intel BID C"
        ),
        (RowId::Bid64ToString, Observation::Text(text)) => assert_eq!(
            native_parse64(text),
            Observation::D64(i.x64, Some(0)),
            "generated Rust bid64_to_string must round-trip through Intel BID C"
        ),
        (RowId::Bid128ToString, Observation::Text(text)) => assert_eq!(
            native_parse128(text),
            Observation::D128(i.x128.lo, i.x128.hi, Some(0)),
            "generated Rust bid128_to_string must round-trip through Intel BID C"
        ),
        _ => {}
    }
}

struct Harness<'a> {
    inputs: &'a PreparedBenchmarkInputs,
    seen: BTreeSet<RowId>,
    observations: BTreeMap<RowId, Observation>,
    groups: BTreeSet<&'static str>,
    active_group: Option<&'static str>,
    same_width: usize,
    mixed: usize,
}

impl Harness<'_> {
    fn verify(&mut self, id: RowId, declared: Contract, actual: Observation) {
        assert!(self.seen.insert(id), "duplicate benchmark row {id:?}");
        assert_eq!(
            self.active_group,
            Some(declared.group),
            "row {id:?} is emitted under the wrong benchmark group"
        );
        assert_eq!(
            declared,
            expected_contract(id),
            "contract mismatch for {id:?}"
        );
        assert_text_round_trip(id, &actual, self.inputs);
        let expected = native_expected(id, self.inputs);
        assert_eq!(actual, expected, "wiring mismatch for {id:?}");
        assert!(
            self.observations.insert(id, expected).is_none(),
            "duplicate benchmark observation {id:?}"
        );
        if declared.group.ends_with("_mixed") {
            self.mixed += 1;
        } else {
            self.same_width += 1;
        }
    }
}

macro_rules! verify_wiring_row {
    (
        $harness:ident,
        $id:ident,
        $group:literal,
        $name:literal,
        $operands:tt,
        $result:ident,
        $rounding:ident,
        $status:ident,
        $call:expr
    ) => {
        $harness.verify(
            RowId::$id,
            Contract {
                group: $group,
                name: $name,
                operands: operand_shape!($operands),
                result: value_kind!($result),
                rounding: rounding_contract!($rounding),
                status: status_contract!($status),
            },
            observe!($result, $status, $call),
        );
    };
}

macro_rules! verify_wiring_group {
    ($harness:ident, $prepared:ident, $name:literal, $count:literal, $rows:ident) => {{
        assert!(
            $harness.groups.insert($name),
            "duplicate benchmark group {}",
            $name
        );
        assert!(
            $harness.active_group.replace($name).is_none(),
            "nested benchmark group {}",
            $name
        );
        let before = $harness.seen.len();
        $rows!(verify_wiring_row, $harness, $prepared);
        assert_eq!(
            $harness.seen.len() - before,
            $count,
            "benchmark row count for group {}",
            $name
        );
        $harness.active_group = None;
    }};
}

struct BenchmarkWiringFixture {
    name: &'static str,
    contract: BenchmarkInputContract,
}

fn benchmark_input_pair(x: &str, y: &str, z: &str) -> BenchmarkInputPair {
    BenchmarkInputPair {
        x: x.to_owned(),
        y: y.to_owned(),
        z: z.to_owned(),
    }
}

fn benchmark_wiring_fixtures() -> Vec<BenchmarkWiringFixture> {
    let production = load_benchmark_input_contract();
    let integer_operand = production.integer_operand;
    let scale_exponent = production.scale_exponent;
    vec![
        BenchmarkWiringFixture {
            name: "production",
            contract: production,
        },
        BenchmarkWiringFixture {
            name: "x_less_than_y",
            contract: BenchmarkInputContract {
                format_version: 2,
                integer_operand,
                scale_exponent,
                decimal32: benchmark_input_pair("4.25", "33", "2.5"),
                decimal64: benchmark_input_pair("7.125", "55", "3.25"),
                decimal128: benchmark_input_pair("11.0625", "96", "5.5"),
            },
        },
        BenchmarkWiringFixture {
            name: "x_greater_than_y",
            contract: BenchmarkInputContract {
                format_version: 2,
                integer_operand,
                scale_exponent,
                decimal32: benchmark_input_pair("33", "4.25", "2.5"),
                decimal64: benchmark_input_pair("55", "7.125", "3.25"),
                decimal128: benchmark_input_pair("96", "11.0625", "5.5"),
            },
        },
    ]
}

fn observation_fingerprint(observation: &Observation) -> String {
    format!("{observation:?}")
}

#[test]
fn benchmark_rows_match_the_intel_bid_oracle() {
    let mut fixture_observations = Vec::new();

    for fixture in benchmark_wiring_fixtures() {
        let prepared = if fixture.name == "production" {
            // Exercise the exact loader used by the Criterion target.
            prepare_benchmark_inputs()
        } else {
            prepare_benchmark_inputs_from_contract(fixture.contract)
        };
        let mut harness = Harness {
            inputs: &prepared,
            seen: BTreeSet::new(),
            observations: BTreeMap::new(),
            groups: BTreeSet::new(),
            active_group: None,
            same_width: 0,
            mixed: 0,
        };

        benchmark_groups!(verify_wiring_group, harness, prepared);

        assert_eq!(
            harness.same_width, 57,
            "{} same-width benchmark row count",
            fixture.name
        );
        assert_eq!(
            harness.mixed, 24,
            "{} mixed-width benchmark row count",
            fixture.name
        );
        assert_eq!(
            harness.seen.len(),
            81,
            "{} unique benchmark row count",
            fixture.name
        );
        assert_eq!(
            harness.observations.len(),
            81,
            "{} unique benchmark observation count",
            fixture.name
        );
        assert_eq!(
            harness.groups.len(),
            5,
            "{} benchmark group count",
            fixture.name
        );
        assert_eq!(
            harness.active_group, None,
            "{} finished benchmark group",
            fixture.name
        );
        fixture_observations.push((fixture.name, harness.observations));
    }

    for group in ["bid32", "bid64", "bid128", "bid64_mixed", "bid128_mixed"] {
        let mut fingerprints = BTreeMap::new();
        for id in fixture_observations[0]
            .1
            .keys()
            .copied()
            .filter(|id| expected_contract(*id).group == group)
        {
            let fingerprint = fixture_observations
                .iter()
                .map(|(fixture, observations)| {
                    let observation = observations
                        .get(&id)
                        .unwrap_or_else(|| panic!("{fixture} missing oracle observation {id:?}"));
                    format!("{fixture}={}", observation_fingerprint(observation))
                })
                .collect::<Vec<_>>()
                .join("|");
            if let Some(previous) = fingerprints.insert(fingerprint, id) {
                panic!(
                    "{group} rows {previous:?} and {id:?} have indistinguishable Intel BID oracle observations across all fixtures"
                );
            }
        }
    }
}

// ---------------------------------------------------------------------------
// Shared hand-pinned benchmark row descriptor conformance.
//
// bid754-go/testdata/benchmark_rows.json pins every benchmark row of all four
// measured layers. The Go native preflight closed-world-compares the three
// Go-side layers against it; this leg exact-matches the descriptor's rust
// layer against the Contract metadata declared row-by-row in
// benches/support/rows.rs (the same declarations the Criterion emitter and
// the Intel BID oracle test above consume).
// ---------------------------------------------------------------------------

#[derive(Debug, serde::Deserialize)]
#[serde(deny_unknown_fields)]
struct SharedDescriptor {
    comment: Vec<String>,
    format_version: u32,
    layer_counts: BTreeMap<String, usize>,
    rows: Vec<SharedDescriptorRow>,
}

#[derive(Debug, serde::Deserialize)]
#[serde(deny_unknown_fields)]
struct SharedDescriptorRow {
    layer: String,
    group: String,
    name: String,
    op: String,
    operands: Vec<String>,
    result: String,
    rounding: String,
    status: String,
}

fn operand_kind_for_token(token: &str) -> ValueKind {
    match token {
        "x32" | "y32" | "z32" | "integer32" => ValueKind::D32,
        "x64" | "y64" | "z64" | "integer64" => ValueKind::D64,
        "x128" | "y128" | "z128" | "integer128" => ValueKind::D128,
        "integer_operand" | "scale_exponent" => ValueKind::I64,
        "decimal32_x_text" | "decimal64_x_text" | "decimal128_x_text" => ValueKind::Text,
        other => panic!("unknown shared descriptor operand token {other:?}"),
    }
}

fn shape_kinds(shape: OperandShape) -> Vec<ValueKind> {
    match shape {
        OperandShape::One(a) => vec![a],
        OperandShape::Two(a, b) => vec![a, b],
        OperandShape::Three(a, b, c) => vec![a, b, c],
    }
}

fn value_kind_token(kind: ValueKind) -> &'static str {
    match kind {
        ValueKind::D32 => "d32",
        ValueKind::D64 => "d64",
        ValueKind::D128 => "d128",
        ValueKind::I64 => "i64",
        ValueKind::Predicate => "predicate",
        ValueKind::Text => "text",
    }
}

fn rounding_token(rounding: RoundingContract) -> &'static str {
    match rounding {
        RoundingContract::ExplicitNearestEven => "explicit_nearest_even",
        RoundingContract::FixedNearestEven => "fixed_nearest_even",
        RoundingContract::NotApplicable => "not_applicable",
    }
}

fn status_matches_descriptor(status: StatusContract, token: &str) -> bool {
    match status {
        StatusContract::TupleFlags | StatusContract::OutFlags => token == "flags_observed",
        StatusContract::NoFlags => token == "value_only",
    }
}

macro_rules! collect_contract_row {
    (
        $rows_vec:ident,
        $id:ident,
        $group:literal,
        $name:literal,
        $operands:tt,
        $result:ident,
        $rounding:ident,
        $status:ident,
        $call:expr
    ) => {
        $rows_vec.push((
            RowId::$id,
            Contract {
                group: $group,
                name: $name,
                operands: operand_shape!($operands),
                result: value_kind!($result),
                rounding: rounding_contract!($rounding),
                status: status_contract!($status),
            },
        ));
    };
}

macro_rules! collect_contract_group {
    ($rows_vec:ident, $prepared:ident, $name:literal, $count:literal, $rows:ident) => {{
        let before = $rows_vec.len();
        $rows!(collect_contract_row, $rows_vec, $prepared);
        assert_eq!(
            $rows_vec.len() - before,
            $count,
            "declared benchmark row count for group {}",
            $name
        );
    }};
}

#[test]
fn benchmark_contracts_match_shared_descriptor() {
    let descriptor: SharedDescriptor = serde_json::from_str(include_str!(
        "../../../bid754-go/testdata/benchmark_rows.json"
    ))
    .expect("parse shared benchmark row descriptor");
    assert_eq!(
        descriptor.format_version, 1,
        "shared descriptor format_version"
    );
    assert!(
        !descriptor.comment.is_empty(),
        "shared descriptor must carry its hand-maintenance comment"
    );
    assert_eq!(
        descriptor.rows.len(),
        descriptor.layer_counts.values().sum::<usize>(),
        "shared descriptor row total vs layer_counts"
    );

    // The collector captures each row's declared contract tokens; the call
    // expressions are bound but never evaluated, so no prepared inputs exist.
    let mut declared: Vec<(RowId, Contract)> = Vec::new();
    benchmark_groups!(collect_contract_group, declared, unused_prepared_inputs);

    let mut by_key: BTreeMap<(String, String), (RowId, Contract)> = BTreeMap::new();
    for (id, contract) in declared {
        assert!(
            by_key
                .insert(
                    (contract.group.to_owned(), contract.name.to_owned()),
                    (id, contract)
                )
                .is_none(),
            "duplicate declared benchmark row {id:?}"
        );
    }

    let rust_rows: Vec<&SharedDescriptorRow> = descriptor
        .rows
        .iter()
        .filter(|row| row.layer == "rust")
        .collect();
    assert_eq!(
        Some(&rust_rows.len()),
        descriptor.layer_counts.get("rust"),
        "shared descriptor rust layer count"
    );
    assert_eq!(
        rust_rows.len(),
        by_key.len(),
        "declared Criterion rows vs shared descriptor rust rows"
    );

    for row in rust_rows {
        let (id, contract) = by_key
            .remove(&(row.group.clone(), row.name.clone()))
            .unwrap_or_else(|| {
                panic!(
                    "shared descriptor rust row {}/{} has no declared Criterion row",
                    row.group, row.name
                )
            });
        assert!(
            !row.op.is_empty(),
            "shared descriptor rust row {}/{} is missing its Intel C op identity",
            row.group,
            row.name
        );
        let declared_operands = shape_kinds(contract.operands);
        let descriptor_operands: Vec<ValueKind> = row
            .operands
            .iter()
            .map(|token| operand_kind_for_token(token))
            .collect();
        assert_eq!(
            declared_operands, descriptor_operands,
            "operand shape for {id:?}"
        );
        assert_eq!(
            value_kind_token(contract.result),
            row.result,
            "result kind for {id:?}"
        );
        assert_eq!(
            rounding_token(contract.rounding),
            row.rounding,
            "rounding contract for {id:?}"
        );
        assert!(
            status_matches_descriptor(contract.status, &row.status),
            "status contract for {id:?}: declared {:?}, descriptor {:?}",
            contract.status,
            row.status
        );
    }
    assert!(
        by_key.is_empty(),
        "declared Criterion rows missing from the shared descriptor: {:?}",
        by_key.keys().collect::<Vec<_>>()
    );
}

#[test]
fn benchmark_setup_rejects_publicly_unrepresentable_cohorts() {
    for (name, mutate) in [
        (
            "decimal32",
            (|contract: &mut BenchmarkInputContract| {
                contract.decimal32.x = "1000000.0".to_owned();
            }) as fn(&mut BenchmarkInputContract),
        ),
        ("decimal64", |contract: &mut BenchmarkInputContract| {
            contract.decimal64.x = "1000000000000000.0".to_owned();
        }),
        ("decimal128", |contract: &mut BenchmarkInputContract| {
            contract.decimal128.x = "1000000000000000000000000000000000.0".to_owned();
        }),
    ] {
        let mut contract = load_benchmark_input_contract();
        mutate(&mut contract);
        let result = std::panic::catch_unwind(|| prepare_benchmark_inputs_from_contract(contract));
        assert!(
            result.is_err(),
            "{name} benchmark setup accepted an unrepresentable cohort"
        );
    }
}

#[test]
fn benchmark_exact_parsers_accept_cohort_boundaries() {
    for input in ["1e-101", "9999999e90"] {
        exact_benchmark_bid32(input);
    }
    for input in ["1e-398", "9999999999999999e369"] {
        exact_benchmark_bid64(input);
    }
    for input in ["1e-6176", "9999999999999999999999999999999999e6111"] {
        exact_benchmark_bid128(input);
    }
}
