use bid754::{Decimal128, Decimal32, Decimal64, RoundingMode};
use serde::Deserialize;

#[derive(Deserialize)]
struct BenchmarkInputPair {
    x: String,
    y: String,
    /// Third fma operand (result = x*y + z); sqrt reuses the non-negative x
    /// operand and needs no dedicated input.
    z: String,
}

#[derive(Deserialize)]
struct BenchmarkInputs {
    format_version: u32,
    integer_operand: i64,
    scale_exponent: i64,
    decimal32: BenchmarkInputPair,
    decimal64: BenchmarkInputPair,
    decimal128: BenchmarkInputPair,
}

fn benchmark_inputs() -> BenchmarkInputs {
    serde_json::from_str(include_str!(
        "../../bid754-go/testdata/benchmark_inputs.json"
    ))
    .expect("parse benchmark input contract")
}

#[test]
fn benchmark_inputs_are_shared_finite_and_exact() {
    let inputs = benchmark_inputs();
    assert_eq!(inputs.format_version, 2);
    assert_ne!(
        inputs.integer_operand, 0,
        "integer_operand must be non-zero"
    );
    assert_ne!(inputs.scale_exponent, 0, "scale_exponent must be non-zero");
    assert!(
        i32::try_from(inputs.scale_exponent).is_ok(),
        "scale_exponent must fit the Intel C int32 contract"
    );

    for input in [
        &inputs.decimal32.x,
        &inputs.decimal32.y,
        &inputs.decimal32.z,
    ] {
        let value =
            Decimal32::parse(input).unwrap_or_else(|err| panic!("Decimal32 {input:?}: {err}"));
        assert!(value.is_finite(), "Decimal32 {input:?} must be finite");
    }
    for input in [
        &inputs.decimal64.x,
        &inputs.decimal64.y,
        &inputs.decimal64.z,
    ] {
        let value =
            Decimal64::parse(input).unwrap_or_else(|err| panic!("Decimal64 {input:?}: {err}"));
        assert!(value.is_finite(), "Decimal64 {input:?} must be finite");
    }
    for input in [
        &inputs.decimal128.x,
        &inputs.decimal128.y,
        &inputs.decimal128.z,
    ] {
        let value =
            Decimal128::parse(input).unwrap_or_else(|err| panic!("Decimal128 {input:?}: {err}"));
        assert!(value.is_finite(), "Decimal128 {input:?} must be finite");
    }

    // The sqrt benchmark rows reuse the x operands; a negative x would
    // silently turn them into NaN/invalid paths instead of square roots.
    // Check the parsed value's sign and NaN class, not the raw text, so a
    // disguised spelling (e.g. " -1") cannot slip past the guard.
    let x32 = Decimal32::parse(&inputs.decimal32.x).expect("parse decimal32 x");
    let y32 = Decimal32::parse(&inputs.decimal32.y).expect("parse decimal32 y");
    assert!(
        !x32.is_nan() && !x32.is_sign_negative(),
        "benchmark input decimal32 x = {:?} parses negative or NaN; sqrt benchmarks reuse x",
        inputs.decimal32.x
    );
    let x64 = Decimal64::parse(&inputs.decimal64.x).expect("parse decimal64 x");
    let y64 = Decimal64::parse(&inputs.decimal64.y).expect("parse decimal64 y");
    assert!(
        !x64.is_nan() && !x64.is_sign_negative(),
        "benchmark input decimal64 x = {:?} parses negative or NaN; sqrt benchmarks reuse x",
        inputs.decimal64.x
    );
    let x128 = Decimal128::parse(&inputs.decimal128.x).expect("parse decimal128 x");
    let y128 = Decimal128::parse(&inputs.decimal128.y).expect("parse decimal128 y");
    assert!(
        !x128.is_nan() && !x128.is_sign_negative(),
        "benchmark input decimal128 x = {:?} parses negative or NaN; sqrt benchmarks reuse x",
        inputs.decimal128.x
    );
    assert!(
        !x32.is_zero() && !y32.is_zero(),
        "Decimal32 division/remainder operands must be non-zero"
    );
    assert!(
        !x64.is_zero() && !y64.is_zero(),
        "Decimal64 division/remainder and mixed operands must be non-zero"
    );
    assert!(
        !x128.is_zero() && !y128.is_zero(),
        "Decimal128 division/remainder and mixed operands must be non-zero"
    );

    let (int32, flags32) = Decimal32::from_i64(inputs.integer_operand, RoundingMode::NearestEven);
    assert!(
        flags32.is_empty(),
        "Decimal32 integer_operand must be exact"
    );
    assert_eq!(
        int32.to_i64(RoundingMode::NearestEven),
        (inputs.integer_operand, flags32),
        "Decimal32 integer_operand must round-trip exactly"
    );
    let (int64, flags64) = Decimal64::from_i64(inputs.integer_operand, RoundingMode::NearestEven);
    assert!(
        flags64.is_empty(),
        "Decimal64 integer_operand must be exact"
    );
    assert_eq!(
        int64.to_i64(RoundingMode::NearestEven),
        (inputs.integer_operand, flags64),
        "Decimal64 integer_operand must round-trip exactly"
    );
    let int128 = Decimal128::from(inputs.integer_operand);
    let (round_trip128, flags128) = int128.to_i64(RoundingMode::NearestEven);
    assert!(
        flags128.is_empty(),
        "Decimal128 integer_operand must be exact"
    );
    assert_eq!(round_trip128, inputs.integer_operand);

    let (scaled32, flags32) = x32.scaleb(inputs.scale_exponent);
    assert!(flags32.is_empty() && scaled32.is_finite());
    let (scaled64, flags64) = x64.scaleb(inputs.scale_exponent);
    assert!(flags64.is_empty() && scaled64.is_finite());
    let (scaled128, flags128) = x128.scaleb(inputs.scale_exponent);
    assert!(flags128.is_empty() && scaled128.is_finite());
}
