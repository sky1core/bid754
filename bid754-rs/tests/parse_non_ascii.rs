//! Regression coverage for non-ASCII input on the public parse surface.
//!
//! Go slices strings by byte and never faults, but the generated Rust operates
//! on &str, whose slicing panics when the cut lands inside a multi-byte
//! character. A prefix probe written as a slice therefore passes every Go-side
//! gate and still panics here on ordinary rejected input such as "1234é".
//!
//! The shared verification corpora are pure ASCII, so no generated domain
//! exercises this class. These rows cover it directly: the public parse API
//! must reject the input through its error type, never by unwinding.

use bid754::{Decimal128, Decimal32, Decimal64};

/// Inputs whose byte offsets 1..n land inside a multi-byte character, plus the
/// dotted capital I that Unicode case mapping folds to ASCII 'i'.
const NON_ASCII_REJECTS: &[&str] = &[
    "1234é",
    "1.23é",
    "-123é",
    "+123é",
    "1234中",
    "aaaé",
    "snané",
    "snaé",
    "é",
    "éé",
    "ééé",
    "1é",
    "12é",
    "123é",
    "İnf",
    "İNFINITY",
    "infé",
    "nané",
    "énan",
    "12345678é",
    "1.2345678é",
    "\u{1F600}",
    "1\u{1F600}",
    "12\u{1F600}",
    "123\u{1F600}",
];

#[test]
fn decimal32_rejects_non_ascii_without_panic() {
    for s in NON_ASCII_REJECTS {
        assert!(
            Decimal32::parse_with_flags(s).is_err(),
            "Decimal32 accepted non-ASCII input {s:?}"
        );
    }
}

#[test]
fn decimal64_rejects_non_ascii_without_panic() {
    for s in NON_ASCII_REJECTS {
        assert!(
            Decimal64::parse_with_flags(s).is_err(),
            "Decimal64 accepted non-ASCII input {s:?}"
        );
    }
}

#[test]
fn decimal128_rejects_non_ascii_without_panic() {
    for s in NON_ASCII_REJECTS {
        assert!(
            Decimal128::parse_with_flags(s).is_err(),
            "Decimal128 accepted non-ASCII input {s:?}"
        );
    }
}

/// Sweeps every byte offset of a multi-byte string through the parser, which is
/// the shape that surfaces a &str slice panic regardless of which probe holds
/// the slice.
#[test]
fn every_byte_offset_of_multibyte_input_is_safe() {
    for base in ["é", "aé", "aaé", "aaaé", "中文字", "İnf", "snané", "1234é"] {
        for end in 1..=base.len() {
            let Some(prefix) = base.get(..end) else {
                continue;
            };
            let _ = Decimal32::parse_with_flags(prefix);
            let _ = Decimal64::parse_with_flags(prefix);
            let _ = Decimal128::parse_with_flags(prefix);
        }
        let _ = Decimal32::parse_with_flags(base);
    }
}
