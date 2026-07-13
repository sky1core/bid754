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
    for (name, x) in [
        ("decimal32", &inputs.decimal32.x),
        ("decimal64", &inputs.decimal64.x),
        ("decimal128", &inputs.decimal128.x),
    ] {
        assert!(
            !x.starts_with('-'),
            "benchmark input {name} x = {x:?} is negative; sqrt benchmarks reuse x"
        );
    }
}
