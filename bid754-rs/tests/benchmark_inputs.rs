use bid754::{Decimal128, Decimal32, Decimal64};
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
    assert_eq!(inputs.format_version, 1);

    for input in [&inputs.decimal32.x, &inputs.decimal32.y, &inputs.decimal32.z] {
        let value =
            Decimal32::parse(input).unwrap_or_else(|err| panic!("Decimal32 {input:?}: {err}"));
        assert!(value.is_finite(), "Decimal32 {input:?} must be finite");
    }
    for input in [&inputs.decimal64.x, &inputs.decimal64.y, &inputs.decimal64.z] {
        let value =
            Decimal64::parse(input).unwrap_or_else(|err| panic!("Decimal64 {input:?}: {err}"));
        assert!(value.is_finite(), "Decimal64 {input:?} must be finite");
    }
    for input in [&inputs.decimal128.x, &inputs.decimal128.y, &inputs.decimal128.z] {
        let value =
            Decimal128::parse(input).unwrap_or_else(|err| panic!("Decimal128 {input:?}: {err}"));
        assert!(value.is_finite(), "Decimal128 {input:?} must be finite");
    }

    // The sqrt benchmark rows reuse the x operands; a negative x would
    // silently turn them into NaN/invalid paths instead of square roots.
    // Check the parsed value's sign and NaN class, not the raw text, so a
    // disguised spelling (e.g. " -1") cannot slip past the guard.
    let x32 = Decimal32::parse(&inputs.decimal32.x).expect("parse decimal32 x");
    assert!(
        !x32.is_nan() && !x32.is_sign_negative(),
        "benchmark input decimal32 x = {:?} parses negative or NaN; sqrt benchmarks reuse x",
        inputs.decimal32.x
    );
    let x64 = Decimal64::parse(&inputs.decimal64.x).expect("parse decimal64 x");
    assert!(
        !x64.is_nan() && !x64.is_sign_negative(),
        "benchmark input decimal64 x = {:?} parses negative or NaN; sqrt benchmarks reuse x",
        inputs.decimal64.x
    );
    let x128 = Decimal128::parse(&inputs.decimal128.x).expect("parse decimal128 x");
    assert!(
        !x128.is_nan() && !x128.is_sign_negative(),
        "benchmark input decimal128 x = {:?} parses negative or NaN; sqrt benchmarks reuse x",
        inputs.decimal128.x
    );
}
