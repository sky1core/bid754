//! Shared operand set for the comparative benchmarks (benches/compare.rs)
//! plus the operand-contract test that keeps it honest: every parse input
//! must be exact and flag-clean on the BID side and parseable by
//! rust_decimal, and every parts row must be encodable on both sides. If an
//! operand edit breaks that shared contract, the test fails before any
//! benchmark number is produced.

pub const PARSE_INPUTS: &[&str] = &[
    "0", "1", "42", "1.5", "-2.25", "12345.67", "0.001", "-999999.99",
    "123456789.123456", "9999999999999.99", "3.14159265358979", "-0.00000001",
];

pub const PARTS_ROWS: &[(i64, i32)] = &[
    (0, 0), (1, 0), (42, 0), (15, -1), (-225, -2), (1234567, -2),
    (1, -3), (-99999999, -2), (123456789123456, -6), (999999999999999, -2),
    (314159265358979, -14), (-1, -8),
];

/// Builds the codec components for one parts row.
pub fn parts_components(m: i64, e: i32) -> bid754::bid_codec::Components {
    let mag = m.unsigned_abs() as u128;
    bid754::bid_codec::Components {
        sign: m < 0,
        coefficient: mag,
        exponent: e,
        kind: if mag == 0 {
            bid754::bid_codec::Kind::Zero
        } else {
            bid754::bid_codec::Kind::Normal
        },
        payload: 0,
    }
}

#[cfg(test)]
mod tests {
    use std::str::FromStr;

    use bid754::{bid_codec, Decimal64};
    use rust_decimal::Decimal;

    use super::*;

    #[test]
    fn operand_contract() {
        for s in PARSE_INPUTS {
            let (_, flags) = Decimal64::parse_raw(s);
            assert!(
                flags.is_empty(),
                "parse input {s:?} is not exact/clean for Decimal64 (flags {flags:?})"
            );
            Decimal::from_str(s).unwrap_or_else(|e| panic!("rust_decimal rejects {s:?}: {e}"));
        }
        for &(m, e) in PARTS_ROWS {
            bid_codec::encode64(&parts_components(m, e))
                .unwrap_or_else(|err| panic!("encode64 rejects ({m}, {e}): {err}"));
            let scale = -e;
            assert!(
                (0..=28).contains(&scale),
                "parts row ({m}, {e}) is outside rust_decimal's scale range"
            );
        }
    }
}
