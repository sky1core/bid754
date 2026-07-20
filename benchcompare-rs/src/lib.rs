//! Shared per-width operand sets for the comparative benchmarks
//! (benches/compare.rs) plus the operand-contract test that keeps them
//! honest: every parse input must be exact and flag-clean at its BID width
//! and parseable by rust_decimal, every parts row must be encodable on both
//! sides, and every benchmark operand pair must stay inside rust_decimal's
//! checked add/mul/div range (its operators panic on overflow). If an
//! operand edit breaks that shared contract, the test fails before any
//! benchmark number is produced.

pub const PARSE_INPUTS_D32: &[&str] = &[
    "0", "1", "42", "1.5", "-2.25", "1234.56", "0.001", "-9999.99",
    "1234567", "-999999.9", "3.14159", "-0.0001",
];

pub const PARSE_INPUTS_D64: &[&str] = &[
    "0", "1", "42", "1.5", "-2.25", "12345.67", "0.001", "-999999.99",
    "123456789.123456", "9999999999999.99", "3.14159265358979", "-0.00000001",
];

/// d128 list: <= 28 significant digits and value magnitudes within
/// [1e-8, 1e12], so the list stays exact in rust_decimal (96-bit mantissa,
/// scale <= 28) and its pairwise products/quotients stay in range.
pub const PARSE_INPUTS_D128: &[&str] = &[
    "0", "1", "42", "1.5", "-2.25", "123456789012.3456789012345678",
    "0.001", "-999999999999.9999999999999999", "1234567890.123456789012345678",
    "3.141592653589793238462643383", "0.9876543210987654321098765432", "-0.00000001",
];

pub const PARTS_ROWS_D32: &[(i64, i32)] = &[
    (0, 0), (1, 0), (42, 0), (15, -1), (-225, -2), (123456, -2),
    (1, -3), (-9999999, -2), (1234567, -4), (9999999, -1),
    (314159, -5), (-1, -7),
];

pub const PARTS_ROWS_D64: &[(i64, i32)] = &[
    (0, 0), (1, 0), (42, 0), (15, -1), (-225, -2), (1234567, -2),
    (1, -3), (-99999999, -2), (123456789123456, -6), (999999999999999, -2),
    (314159265358979, -14), (-1, -8),
];

pub const PARTS_ROWS_D128: &[(i64, i32)] = &[
    (0, 0), (1, 0), (42, 0), (15, -1), (-225, -2), (1234567, -2),
    (1, -3), (-99999999, -2), (123456789012345678, -9),
    (999999999999999999, -12), (314159265358979323, -17), (-1, -8),
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

/// Operand index pairs shared by every width's add/mul rows.
pub fn pair_indices(n: usize) -> Vec<(usize, usize)> {
    (0..n).map(|i| (i, (i * 7 + 3) % n)).collect()
}

/// Divisor index pairs: index 0 is the literal "0" in every width list and
/// is never used as a divisor.
pub fn div_pair_indices(n: usize) -> Vec<(usize, usize)> {
    (0..n)
        .map(|i| {
            let j = (i * 5 + 1) % n;
            (i, if j == 0 { 1 } else { j })
        })
        .collect()
}

#[cfg(test)]
mod tests {
    use std::str::FromStr;

    use bid754::{bid_codec, Decimal128, Decimal32, Decimal64};
    use rust_decimal::Decimal;

    use super::*;

    fn check_parse_side(width: &str, inputs: &[&str], flags_clean: impl Fn(&str) -> bool) -> Vec<Decimal> {
        inputs
            .iter()
            .map(|s| {
                assert!(flags_clean(s), "{width} parse input {s:?} is not exact/clean for BID");
                Decimal::from_str(s)
                    .unwrap_or_else(|e| panic!("rust_decimal rejects {width} input {s:?}: {e}"))
            })
            .collect()
    }

    fn check_pair_ranges(width: &str, vals: &[Decimal]) {
        for &(i, j) in &pair_indices(vals.len()) {
            assert!(
                vals[i].checked_add(vals[j]).is_some(),
                "{width} add pair ({i},{j}) overflows rust_decimal"
            );
            assert!(
                vals[i].checked_mul(vals[j]).is_some(),
                "{width} mul pair ({i},{j}) overflows rust_decimal"
            );
        }
        for &(i, j) in &div_pair_indices(vals.len()) {
            assert!(!vals[j].is_zero(), "{width} div pair ({i},{j}) has a zero divisor");
            assert!(
                vals[i].checked_div(vals[j]).is_some(),
                "{width} div pair ({i},{j}) overflows rust_decimal"
            );
        }
    }

    #[test]
    fn operand_contract() {
        let d32 = check_parse_side("d32", PARSE_INPUTS_D32, |s| Decimal32::parse_raw(s).1.is_empty());
        let d64 = check_parse_side("d64", PARSE_INPUTS_D64, |s| Decimal64::parse_raw(s).1.is_empty());
        let d128 = check_parse_side("d128", PARSE_INPUTS_D128, |s| {
            Decimal128::parse_raw(s).1.is_empty()
        });
        assert_eq!(d32.len(), d64.len());
        assert_eq!(d64.len(), d128.len());
        check_pair_ranges("d32", &d32);
        check_pair_ranges("d64", &d64);
        check_pair_ranges("d128", &d128);

        for &(m, e) in PARTS_ROWS_D32 {
            bid_codec::encode32(&parts_components(m, e))
                .unwrap_or_else(|err| panic!("encode32 rejects ({m}, {e}): {err}"));
            assert!((0..=28).contains(&-e), "d32 parts row ({m}, {e}) outside rust_decimal scale");
        }
        for &(m, e) in PARTS_ROWS_D64 {
            bid_codec::encode64(&parts_components(m, e))
                .unwrap_or_else(|err| panic!("encode64 rejects ({m}, {e}): {err}"));
            assert!((0..=28).contains(&-e), "d64 parts row ({m}, {e}) outside rust_decimal scale");
        }
        for &(m, e) in PARTS_ROWS_D128 {
            bid_codec::encode128(&parts_components(m, e))
                .unwrap_or_else(|err| panic!("encode128 rejects ({m}, {e}): {err}"));
            assert!((0..=28).contains(&-e), "d128 parts row ({m}, {e}) outside rust_decimal scale");
        }
    }
}
