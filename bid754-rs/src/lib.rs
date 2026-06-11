//! Repository-internal Rust implementation crate for bid754.
//!
//! This crate is not a stable public Rust API and is not published. The
//! `pub` modules below are exposed so repository integration tests, benches,
//! and generated verification harnesses can call the Go-to-Rust generated
//! implementation directly. Treat `gen_types`, `gen_constants`, `tables`, and
//! `generated` as generated/internal compatibility surface, not as a semver
//! contract.
//!
//! Standalone BID component encode/decode helpers live in `bid-codec-rs`; that
//! package has its own public package boundary and generated vector tests.

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

// Generated from symbol registry (tools/registry/symbols.json)
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
