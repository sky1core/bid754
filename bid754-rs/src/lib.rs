//! Public Rust implementation crate for bid754, in pre-release form.
//!
//! This crate is generated from the Go mechanical port of Intel BID C. It is
//! still `publish = false`, and it does not define a stable external Rust API
//! yet: the user-facing Rust API layer is a later phase. The `pub` modules
//! below are exposed so repository integration tests, benches,
//! and generated verification harnesses can call the Go-to-Rust generated
//! implementation directly. Treat `gen_types`, `gen_constants`, `tables`, and
//! `generated` as generated/non-stable compatibility surface, not as a semver
//! contract.
//!
//! Standalone BID component encode/decode helpers live in `bid754-codec-rs`;
//! that package has its own public package boundary and generated vector
//! tests.

#![allow(
    non_snake_case,
    non_camel_case_types,
    non_upper_case_globals,
    unused_mut,
    unused_parens,
    unused_variables,
    dead_code,
    unused_assignments,
    unused_imports,
    arithmetic_overflow,
    overflowing_literals
)]

// Generated from symbol registry (devtools/tools/registry/symbols.json)
pub mod gen_types;
pub mod gen_constants;

// Generated table compatibility layer.
pub mod tables;

// Generated BID functions (from go2rs converter)
pub mod generated;

// BID component encode/decode helpers for BID codec vectors
pub mod bid_codec;

pub fn bid64_from_string_raw(s: impl AsRef<str>, rnd_mode: i32) -> (u64, u32) {
    generated::bid64_from_string::bid64_from_string(s, i64::from(rnd_mode))
}
