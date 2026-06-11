#![cfg(feature = "ffi-fuzz")]

use std::mem::{align_of, size_of};

use bid754::gen_types::BID_UINT128 as RustBid128;
use bid754::generated::prelude::*;
use libbid_sys::BID_UINT128 as CbBid128;
use rand::rngs::StdRng;
use rand::{RngCore, SeedableRng};

const CASES_PER_FUNCTION: usize = 10_000;
const ROUNDING_MODES: [u32; 5] = [
    libbid_sys::BID_ROUNDING_TO_NEAREST,
    libbid_sys::BID_ROUNDING_DOWN,
    libbid_sys::BID_ROUNDING_UP,
    libbid_sys::BID_ROUNDING_TO_ZERO,
    libbid_sys::BID_ROUNDING_TIES_AWAY,
];

fn next_rounding_mode(rng: &mut StdRng) -> i32 {
    ROUNDING_MODES[(rng.next_u32() as usize) % ROUNDING_MODES.len()] as i32
}

fn next_bid32(rng: &mut StdRng) -> u32 {
    rng.next_u32()
}

fn next_bid64(rng: &mut StdRng) -> u64 {
    rng.next_u64()
}

fn next_bid128(rng: &mut StdRng) -> RustBid128 {
    RustBid128 {
        w: [rng.next_u64(), rng.next_u64()],
    }
}

fn to_c_bid128(x: RustBid128) -> CbBid128 {
    CbBid128 { w: x.w }
}

fn to_rust_bid128(x: CbBid128) -> RustBid128 {
    RustBid128 { w: x.w }
}

fn fmt_rust_bid128(x: RustBid128) -> String {
    format!("[lo=0x{:016x}, hi=0x{:016x}]", x.w[0], x.w[1])
}

macro_rules! fuzz_binary_u32 {
    ($test_name:ident, $rust_fn:expr, $c_fn:ident, $seed:expr) => {
        #[test]
        fn $test_name() {
            let mut rng = StdRng::seed_from_u64($seed);
            for case_idx in 0..CASES_PER_FUNCTION {
                let x = next_bid32(&mut rng);
                let y = next_bid32(&mut rng);
                let rnd = next_rounding_mode(&mut rng);
                let mut c_flags = 0u32;
                let c_result = unsafe { libbid_sys::$c_fn(x, y, rnd as u32, &mut c_flags) };
                let (rust_result, rust_flags) = $rust_fn(x, y, i64::from(rnd));
                assert_eq!(
                    rust_result,
                    c_result,
                    "{} result mismatch at case {}: x=0x{:08x}, y=0x{:08x}, rnd={}, rust=0x{:08x}/{:02x}, c=0x{:08x}/{:02x}",
                    stringify!($test_name),
                    case_idx,
                    x,
                    y,
                    rnd,
                    rust_result,
                    rust_flags,
                    c_result,
                    c_flags
                );
                assert_eq!(
                    rust_flags,
                    c_flags,
                    "{} flags mismatch at case {}: x=0x{:08x}, y=0x{:08x}, rnd={}, rust=0x{:08x}/{:02x}, c=0x{:08x}/{:02x}",
                    stringify!($test_name),
                    case_idx,
                    x,
                    y,
                    rnd,
                    rust_result,
                    rust_flags,
                    c_result,
                    c_flags
                );
            }
        }
    };
}

macro_rules! fuzz_unary_u32 {
    ($test_name:ident, $rust_fn:expr, $c_fn:ident, $seed:expr) => {
        #[test]
        fn $test_name() {
            let mut rng = StdRng::seed_from_u64($seed);
            for case_idx in 0..CASES_PER_FUNCTION {
                let x = next_bid32(&mut rng);
                let rnd = next_rounding_mode(&mut rng);
                let mut c_flags = 0u32;
                let c_result = unsafe { libbid_sys::$c_fn(x, rnd as u32, &mut c_flags) };
                let (rust_result, rust_flags) = $rust_fn(x, i64::from(rnd));
                assert_eq!(
                    rust_result,
                    c_result,
                    "{} result mismatch at case {}: x=0x{:08x}, rnd={}, rust=0x{:08x}/{:02x}, c=0x{:08x}/{:02x}",
                    stringify!($test_name),
                    case_idx,
                    x,
                    rnd,
                    rust_result,
                    rust_flags,
                    c_result,
                    c_flags
                );
                assert_eq!(
                    rust_flags,
                    c_flags,
                    "{} flags mismatch at case {}: x=0x{:08x}, rnd={}, rust=0x{:08x}/{:02x}, c=0x{:08x}/{:02x}",
                    stringify!($test_name),
                    case_idx,
                    x,
                    rnd,
                    rust_result,
                    rust_flags,
                    c_result,
                    c_flags
                );
            }
        }
    };
}

macro_rules! fuzz_ternary_u32 {
    ($test_name:ident, $rust_fn:expr, $c_fn:ident, $seed:expr) => {
        #[test]
        fn $test_name() {
            let mut rng = StdRng::seed_from_u64($seed);
            for case_idx in 0..CASES_PER_FUNCTION {
                let x = next_bid32(&mut rng);
                let y = next_bid32(&mut rng);
                let z = next_bid32(&mut rng);
                let rnd = next_rounding_mode(&mut rng);
                let mut c_flags = 0u32;
                let c_result = unsafe { libbid_sys::$c_fn(x, y, z, rnd as u32, &mut c_flags) };
                let (rust_result, rust_flags) = $rust_fn(x, y, z, i64::from(rnd));
                assert_eq!(
                    rust_result,
                    c_result,
                    "{} result mismatch at case {}: x=0x{:08x}, y=0x{:08x}, z=0x{:08x}, rnd={}, rust=0x{:08x}/{:02x}, c=0x{:08x}/{:02x}",
                    stringify!($test_name),
                    case_idx,
                    x,
                    y,
                    z,
                    rnd,
                    rust_result,
                    rust_flags,
                    c_result,
                    c_flags
                );
                assert_eq!(
                    rust_flags,
                    c_flags,
                    "{} flags mismatch at case {}: x=0x{:08x}, y=0x{:08x}, z=0x{:08x}, rnd={}, rust=0x{:08x}/{:02x}, c=0x{:08x}/{:02x}",
                    stringify!($test_name),
                    case_idx,
                    x,
                    y,
                    z,
                    rnd,
                    rust_result,
                    rust_flags,
                    c_result,
                    c_flags
                );
            }
        }
    };
}

macro_rules! fuzz_binary_u64 {
    ($test_name:ident, $rust_fn:expr, $c_fn:ident, $seed:expr) => {
        #[test]
        fn $test_name() {
            let mut rng = StdRng::seed_from_u64($seed);
            for case_idx in 0..CASES_PER_FUNCTION {
                let x = next_bid64(&mut rng);
                let y = next_bid64(&mut rng);
                let rnd = next_rounding_mode(&mut rng);
                let mut c_flags = 0u32;
                let c_result = unsafe { libbid_sys::$c_fn(x, y, rnd as u32, &mut c_flags) };
                let (rust_result, rust_flags) = $rust_fn(x, y, i64::from(rnd));
                assert_eq!(
                    rust_result,
                    c_result,
                    "{} result mismatch at case {}: x=0x{:016x}, y=0x{:016x}, rnd={}, rust=0x{:016x}/{:02x}, c=0x{:016x}/{:02x}",
                    stringify!($test_name),
                    case_idx,
                    x,
                    y,
                    rnd,
                    rust_result,
                    rust_flags,
                    c_result,
                    c_flags
                );
                assert_eq!(
                    rust_flags,
                    c_flags,
                    "{} flags mismatch at case {}: x=0x{:016x}, y=0x{:016x}, rnd={}, rust=0x{:016x}/{:02x}, c=0x{:016x}/{:02x}",
                    stringify!($test_name),
                    case_idx,
                    x,
                    y,
                    rnd,
                    rust_result,
                    rust_flags,
                    c_result,
                    c_flags
                );
            }
        }
    };
}

macro_rules! fuzz_unary_u64 {
    ($test_name:ident, $rust_fn:expr, $c_fn:ident, $seed:expr) => {
        #[test]
        fn $test_name() {
            let mut rng = StdRng::seed_from_u64($seed);
            for case_idx in 0..CASES_PER_FUNCTION {
                let x = next_bid64(&mut rng);
                let rnd = next_rounding_mode(&mut rng);
                let mut c_flags = 0u32;
                let c_result = unsafe { libbid_sys::$c_fn(x, rnd as u32, &mut c_flags) };
                let (rust_result, rust_flags) = $rust_fn(x, i64::from(rnd));
                assert_eq!(
                    rust_result,
                    c_result,
                    "{} result mismatch at case {}: x=0x{:016x}, rnd={}, rust=0x{:016x}/{:02x}, c=0x{:016x}/{:02x}",
                    stringify!($test_name),
                    case_idx,
                    x,
                    rnd,
                    rust_result,
                    rust_flags,
                    c_result,
                    c_flags
                );
                assert_eq!(
                    rust_flags,
                    c_flags,
                    "{} flags mismatch at case {}: x=0x{:016x}, rnd={}, rust=0x{:016x}/{:02x}, c=0x{:016x}/{:02x}",
                    stringify!($test_name),
                    case_idx,
                    x,
                    rnd,
                    rust_result,
                    rust_flags,
                    c_result,
                    c_flags
                );
            }
        }
    };
}

macro_rules! fuzz_ternary_u64 {
    ($test_name:ident, $rust_fn:expr, $c_fn:ident, $seed:expr) => {
        #[test]
        fn $test_name() {
            let mut rng = StdRng::seed_from_u64($seed);
            for case_idx in 0..CASES_PER_FUNCTION {
                let x = next_bid64(&mut rng);
                let y = next_bid64(&mut rng);
                let z = next_bid64(&mut rng);
                let rnd = next_rounding_mode(&mut rng);
                let mut c_flags = 0u32;
                let c_result = unsafe { libbid_sys::$c_fn(x, y, z, rnd as u32, &mut c_flags) };
                let (rust_result, rust_flags) = $rust_fn(x, y, z, i64::from(rnd));
                assert_eq!(
                    rust_result,
                    c_result,
                    "{} result mismatch at case {}: x=0x{:016x}, y=0x{:016x}, z=0x{:016x}, rnd={}, rust=0x{:016x}/{:02x}, c=0x{:016x}/{:02x}",
                    stringify!($test_name),
                    case_idx,
                    x,
                    y,
                    z,
                    rnd,
                    rust_result,
                    rust_flags,
                    c_result,
                    c_flags
                );
                assert_eq!(
                    rust_flags,
                    c_flags,
                    "{} flags mismatch at case {}: x=0x{:016x}, y=0x{:016x}, z=0x{:016x}, rnd={}, rust=0x{:016x}/{:02x}, c=0x{:016x}/{:02x}",
                    stringify!($test_name),
                    case_idx,
                    x,
                    y,
                    z,
                    rnd,
                    rust_result,
                    rust_flags,
                    c_result,
                    c_flags
                );
            }
        }
    };
}

macro_rules! fuzz_binary_u128 {
    ($test_name:ident, $rust_fn:expr, $c_fn:ident, $seed:expr) => {
        #[test]
        fn $test_name() {
            let mut rng = StdRng::seed_from_u64($seed);
            for case_idx in 0..CASES_PER_FUNCTION {
                let x = next_bid128(&mut rng);
                let y = next_bid128(&mut rng);
                let rnd = next_rounding_mode(&mut rng);
                let mut c_flags = 0u32;
                let c_result = to_rust_bid128(unsafe { libbid_sys::$c_fn(to_c_bid128(x), to_c_bid128(y), rnd as u32, &mut c_flags) });
                let (rust_result, rust_flags) = $rust_fn(x, y, i64::from(rnd));
                assert_eq!(
                    rust_result,
                    c_result,
                    "{} result mismatch at case {}: x={}, y={}, rnd={}, rust={}/{:02x}, c={}/{:02x}",
                    stringify!($test_name),
                    case_idx,
                    fmt_rust_bid128(x),
                    fmt_rust_bid128(y),
                    rnd,
                    fmt_rust_bid128(rust_result),
                    rust_flags,
                    fmt_rust_bid128(c_result),
                    c_flags
                );
                assert_eq!(
                    rust_flags,
                    c_flags,
                    "{} flags mismatch at case {}: x={}, y={}, rnd={}, rust={}/{:02x}, c={}/{:02x}",
                    stringify!($test_name),
                    case_idx,
                    fmt_rust_bid128(x),
                    fmt_rust_bid128(y),
                    rnd,
                    fmt_rust_bid128(rust_result),
                    rust_flags,
                    fmt_rust_bid128(c_result),
                    c_flags
                );
            }
        }
    };
}

macro_rules! fuzz_unary_u128 {
    ($test_name:ident, $rust_fn:expr, $c_fn:ident, $seed:expr) => {
        #[test]
        fn $test_name() {
            let mut rng = StdRng::seed_from_u64($seed);
            for case_idx in 0..CASES_PER_FUNCTION {
                let x = next_bid128(&mut rng);
                let rnd = next_rounding_mode(&mut rng);
                let mut c_flags = 0u32;
                let c_result = to_rust_bid128(unsafe {
                    libbid_sys::$c_fn(to_c_bid128(x), rnd as u32, &mut c_flags)
                });
                let (rust_result, rust_flags) = $rust_fn(x, i64::from(rnd));
                assert_eq!(
                    rust_result,
                    c_result,
                    "{} result mismatch at case {}: x={}, rnd={}, rust={}/{:02x}, c={}/{:02x}",
                    stringify!($test_name),
                    case_idx,
                    fmt_rust_bid128(x),
                    rnd,
                    fmt_rust_bid128(rust_result),
                    rust_flags,
                    fmt_rust_bid128(c_result),
                    c_flags
                );
                assert_eq!(
                    rust_flags,
                    c_flags,
                    "{} flags mismatch at case {}: x={}, rnd={}, rust={}/{:02x}, c={}/{:02x}",
                    stringify!($test_name),
                    case_idx,
                    fmt_rust_bid128(x),
                    rnd,
                    fmt_rust_bid128(rust_result),
                    rust_flags,
                    fmt_rust_bid128(c_result),
                    c_flags
                );
            }
        }
    };
}

macro_rules! fuzz_ternary_u128 {
    ($test_name:ident, $rust_fn:expr, $c_fn:ident, $seed:expr) => {
        #[test]
        fn $test_name() {
            let mut rng = StdRng::seed_from_u64($seed);
            for case_idx in 0..CASES_PER_FUNCTION {
                let x = next_bid128(&mut rng);
                let y = next_bid128(&mut rng);
                let z = next_bid128(&mut rng);
                let rnd = next_rounding_mode(&mut rng);
                let mut c_flags = 0u32;
                let c_result = to_rust_bid128(unsafe {
                    libbid_sys::$c_fn(to_c_bid128(x), to_c_bid128(y), to_c_bid128(z), rnd as u32, &mut c_flags)
                });
                let (rust_result, rust_flags) = $rust_fn(x, y, z, i64::from(rnd));
                assert_eq!(
                    rust_result,
                    c_result,
                    "{} result mismatch at case {}: x={}, y={}, z={}, rnd={}, rust={}/{:02x}, c={}/{:02x}",
                    stringify!($test_name),
                    case_idx,
                    fmt_rust_bid128(x),
                    fmt_rust_bid128(y),
                    fmt_rust_bid128(z),
                    rnd,
                    fmt_rust_bid128(rust_result),
                    rust_flags,
                    fmt_rust_bid128(c_result),
                    c_flags
                );
                assert_eq!(
                    rust_flags,
                    c_flags,
                    "{} flags mismatch at case {}: x={}, y={}, z={}, rnd={}, rust={}/{:02x}, c={}/{:02x}",
                    stringify!($test_name),
                    case_idx,
                    fmt_rust_bid128(x),
                    fmt_rust_bid128(y),
                    fmt_rust_bid128(z),
                    rnd,
                    fmt_rust_bid128(rust_result),
                    rust_flags,
                    fmt_rust_bid128(c_result),
                    c_flags
                );
            }
        }
    };
}

#[test]
fn bid128_ffi_layout_matches_c_expectation() {
    assert_eq!(
        size_of::<CbBid128>(),
        16,
        "C BID_UINT128 size must be 16 bytes"
    );
    assert_eq!(
        align_of::<CbBid128>(),
        16,
        "C BID_UINT128 alignment must be 16 bytes"
    );
    assert_eq!(
        size_of::<RustBid128>(),
        16,
        "Rust BID_UINT128 size must remain 16 bytes"
    );
}

fuzz_binary_u64!(
    fuzz_bid64_add_vs_c,
    |x, y, rnd| bid64_add_with_flags(x, y, rnd),
    bid64_add,
    0xdec7_5400_0000_0001
);
fuzz_binary_u64!(
    fuzz_bid64_sub_vs_c,
    |x, y, rnd| bid64_sub_with_flags(x, y, rnd),
    bid64_sub,
    0xdec7_5400_0000_0002
);
fuzz_binary_u64!(
    fuzz_bid64_mul_vs_c,
    |x, y, rnd| bid64_mul_with_flags(x, y, rnd),
    bid64_mul,
    0xdec7_5400_0000_0003
);
fuzz_binary_u64!(
    fuzz_bid64_div_vs_c,
    |x, y, rnd| bid64_div_with_flags(x, y, rnd),
    bid64_div,
    0xdec7_5400_0000_0004
);
fuzz_unary_u64!(
    fuzz_bid64_sqrt_vs_c,
    |x, rnd| bid64_sqrt(x, rnd),
    bid64_sqrt,
    0xdec7_5400_0000_0005
);
fuzz_ternary_u64!(
    fuzz_bid64_fma_vs_c,
    |x, y, z, rnd| bid64_fma(x, y, z, rnd),
    bid64_fma,
    0xdec7_5400_0000_0006
);

fuzz_binary_u128!(
    fuzz_bid128_add_vs_c,
    |x, y, rnd| {
        let mut flags = 0u32;
        let result = bid128_add(x, y, rnd, &mut flags);
        (result, flags)
    },
    bid128_add,
    0xdec7_5412_8000_0001
);
fuzz_binary_u128!(
    fuzz_bid128_sub_vs_c,
    |x, y, rnd| {
        let mut flags = 0u32;
        let result = bid128_sub(x, y, rnd, &mut flags);
        (result, flags)
    },
    bid128_sub,
    0xdec7_5412_8000_0002
);
fuzz_binary_u128!(
    fuzz_bid128_mul_vs_c,
    |x, y, rnd| bid128_mul(x, y, rnd),
    bid128_mul,
    0xdec7_5412_8000_0003
);
fuzz_binary_u128!(
    fuzz_bid128_div_vs_c,
    |x, y, rnd| bid128_div(x, y, rnd),
    bid128_div,
    0xdec7_5412_8000_0004
);
fuzz_unary_u128!(
    fuzz_bid128_sqrt_vs_c,
    |x, rnd| bid128_sqrt(x, rnd),
    bid128_sqrt,
    0xdec7_5412_8000_0005
);
fuzz_ternary_u128!(
    fuzz_bid128_fma_vs_c,
    |x, y, z, rnd| bid128_fma(x, y, z, rnd),
    bid128_fma,
    0xdec7_5412_8000_0006
);

fuzz_binary_u32!(
    fuzz_bid32_add_vs_c,
    |x, y, rnd| bid32_add_with_flags(x, y, rnd),
    bid32_add,
    0xdec7_5432_0000_0001
);
fuzz_binary_u32!(
    fuzz_bid32_sub_vs_c,
    |x, y, rnd| bid32_sub_with_flags(x, y, rnd),
    bid32_sub,
    0xdec7_5432_0000_0002
);
fuzz_binary_u32!(
    fuzz_bid32_mul_vs_c,
    |x, y, rnd| bid32_mul_with_flags(x, y, rnd),
    bid32_mul,
    0xdec7_5432_0000_0003
);
fuzz_binary_u32!(
    fuzz_bid32_div_vs_c,
    |x, y, rnd| bid32_div_with_flags(x, y, rnd),
    bid32_div,
    0xdec7_5432_0000_0004
);
fuzz_unary_u32!(
    fuzz_bid32_sqrt_vs_c,
    |x, rnd| bid32_sqrt(x, rnd),
    bid32_sqrt,
    0xdec7_5432_0000_0005
);
fuzz_ternary_u32!(
    fuzz_bid32_fma_vs_c,
    |x, y, z, rnd| bid32_fma(x, y, z, rnd),
    bid32_fma,
    0xdec7_5432_0000_0006
);
