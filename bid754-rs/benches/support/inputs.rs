// Shared benchmark input deserialization, validation, and preparation.
//
// Both the Criterion target and the native wiring verifier include this file,
// so neither can silently use a different fixture or preparation path.

#[derive(serde::Deserialize)]
struct BenchmarkInputPair {
    x: String,
    y: String,
    z: String,
}

#[derive(serde::Deserialize)]
struct BenchmarkInputContract {
    format_version: u32,
    integer_operand: i64,
    scale_exponent: i64,
    decimal32: BenchmarkInputPair,
    decimal64: BenchmarkInputPair,
    decimal128: BenchmarkInputPair,
}

struct PreparedBenchmarkInputs {
    integer_operand: i64,
    scale_exponent: i64,
    decimal32_x_text: String,
    decimal64_x_text: String,
    decimal128_x_text: String,
    x32: u32,
    y32: u32,
    z32: u32,
    x64: u64,
    y64: u64,
    z64: u64,
    x128: bid754::gen_types::BID_UINT128,
    y128: bid754::gen_types::BID_UINT128,
    z128: bid754::gen_types::BID_UINT128,
    integer32: u32,
    integer64: u64,
    integer128: bid754::gen_types::BID_UINT128,
}

fn exact_benchmark_bid32(input: &str) -> u32 {
    let (value, flags) = bid754::generated::bid32_string::bid32_from_string_raw(input, 0);
    assert_eq!(flags, 0, "Decimal32 benchmark input flags");
    assert_ne!(
        bid754::generated::bid32_noncomp::bid32_is_finite(value),
        0,
        "Decimal32 benchmark input must be finite"
    );
    let public = bid754::Decimal32::parse(input)
        .unwrap_or_else(|err| panic!("Decimal32 benchmark input {input:?} is not exact: {err}"));
    assert_eq!(
        value,
        public.to_bits(),
        "Decimal32 benchmark raw/public parse bits"
    );
    value
}

fn exact_benchmark_bid64(input: &str) -> u64 {
    let (value, flags) = bid754::bid64_from_string_raw(input, 0);
    assert_eq!(flags, 0, "Decimal64 benchmark input flags");
    assert_ne!(
        bid754::generated::noncomp64::bid64_is_finite(value),
        0,
        "Decimal64 benchmark input must be finite"
    );
    let public = bid754::Decimal64::parse(input)
        .unwrap_or_else(|err| panic!("Decimal64 benchmark input {input:?} is not exact: {err}"));
    assert_eq!(
        value,
        public.to_bits(),
        "Decimal64 benchmark raw/public parse bits"
    );
    value
}

fn exact_benchmark_bid128(input: &str) -> bid754::gen_types::BID_UINT128 {
    let (value, flags) = bid754::generated::bid128_string::bid128_from_string(input, 0);
    assert_eq!(flags, 0, "Decimal128 benchmark input flags");
    assert_ne!(
        bid754::generated::bid128_noncomp::bid128_is_finite(value),
        0,
        "Decimal128 benchmark input must be finite"
    );
    let public = bid754::Decimal128::parse(input)
        .unwrap_or_else(|err| panic!("Decimal128 benchmark input {input:?} is not exact: {err}"));
    let public_bits = public.to_le_bytes();
    let public_lo = u64::from_le_bytes(
        public_bits[..8]
            .try_into()
            .expect("Decimal128 low BID bytes"),
    );
    let public_hi = u64::from_le_bytes(
        public_bits[8..]
            .try_into()
            .expect("Decimal128 high BID bytes"),
    );
    assert_eq!(
        (value.lo, value.hi),
        (public_lo, public_hi),
        "Decimal128 benchmark raw/public parse bits"
    );
    value
}

fn load_benchmark_input_contract() -> BenchmarkInputContract {
    serde_json::from_str(include_str!(
        "../../../bid754-go/testdata/benchmark_inputs.json"
    ))
    .expect("parse benchmark input contract")
}

fn prepare_benchmark_inputs_from_contract(
    contract: BenchmarkInputContract,
) -> PreparedBenchmarkInputs {
    assert_eq!(contract.format_version, 2, "benchmark input format_version");
    assert_ne!(contract.integer_operand, 0, "benchmark integer_operand");
    assert_ne!(contract.scale_exponent, 0, "benchmark scale_exponent");
    assert!(
        i32::try_from(contract.scale_exponent).is_ok(),
        "benchmark scale_exponent must fit the Intel C int32 contract"
    );
    for (name, pair) in [
        ("decimal32", &contract.decimal32),
        ("decimal64", &contract.decimal64),
        ("decimal128", &contract.decimal128),
    ] {
        assert!(
            !pair.x.is_empty() && !pair.y.is_empty() && !pair.z.is_empty(),
            "benchmark input {name} requires non-empty x, y, and z"
        );
    }

    let x32 = exact_benchmark_bid32(&contract.decimal32.x);
    let y32 = exact_benchmark_bid32(&contract.decimal32.y);
    let z32 = exact_benchmark_bid32(&contract.decimal32.z);
    let x64 = exact_benchmark_bid64(&contract.decimal64.x);
    let y64 = exact_benchmark_bid64(&contract.decimal64.y);
    let z64 = exact_benchmark_bid64(&contract.decimal64.z);
    let x128 = exact_benchmark_bid128(&contract.decimal128.x);
    let y128 = exact_benchmark_bid128(&contract.decimal128.y);
    let z128 = exact_benchmark_bid128(&contract.decimal128.z);

    assert_eq!(
        bid754::generated::bid32_noncomp::bid32_is_signed(x32),
        0,
        "Decimal32 sqrt operand sign"
    );
    assert_eq!(
        bid754::generated::noncomp64::bid64_is_signed(x64),
        0,
        "Decimal64 sqrt operand sign"
    );
    assert_eq!(
        bid754::generated::bid128_noncomp::bid128_is_signed(x128),
        0,
        "Decimal128 sqrt operand sign"
    );
    assert_eq!(
        bid754::generated::bid32_noncomp::bid32_is_zero32(x32),
        0,
        "Decimal32 remainder divisor"
    );
    assert_eq!(
        bid754::generated::bid32_noncomp::bid32_is_zero32(y32),
        0,
        "Decimal32 division divisor"
    );
    assert_eq!(
        bid754::generated::noncomp64::bid64_is_zero(x64),
        0,
        "Decimal64 remainder/mixed operand"
    );
    assert_eq!(
        bid754::generated::noncomp64::bid64_is_zero(y64),
        0,
        "Decimal64 division/mixed divisor"
    );
    assert_eq!(
        bid754::generated::bid128_internal::bid128_is_zero(x128),
        0,
        "Decimal128 remainder/mixed operand"
    );
    assert_eq!(
        bid754::generated::bid128_internal::bid128_is_zero(y128),
        0,
        "Decimal128 division/mixed divisor"
    );

    let (integer32, flags32) =
        bid754::generated::bid32_to_int::bid32_from_int64(contract.integer_operand, 0);
    assert_eq!(flags32, 0, "Decimal32 integer benchmark operand flags");
    assert_eq!(
        bid754::generated::bid32_to_int::bid32_to_int64_rnint(integer32),
        (contract.integer_operand, 0),
        "Decimal32 integer benchmark operand round trip"
    );
    let (integer64, flags64) =
        bid754::generated::convert64::bid64_from_int64(contract.integer_operand, 0);
    assert_eq!(flags64, 0, "Decimal64 integer benchmark operand flags");
    assert_eq!(
        bid754::generated::to_int64_signed::bid64_to_int64_rnint(integer64),
        (contract.integer_operand, 0),
        "Decimal64 integer benchmark operand round trip"
    );
    let integer128 =
        bid754::generated::bid128_from_int::bid128_from_int64(contract.integer_operand);
    assert_eq!(
        bid754::generated::bid128_to_int::bid128_to_int64_rnint(integer128),
        (contract.integer_operand, 0),
        "Decimal128 integer benchmark operand round trip"
    );

    let (scaled32, scale_flags32) =
        bid754::generated::bid32_status::bid32_scalbn_with_flags(x32, contract.scale_exponent, 0);
    assert_eq!(scale_flags32, 0, "Decimal32 scaleB benchmark flags");
    assert_ne!(
        bid754::generated::bid32_noncomp::bid32_is_finite(scaled32),
        0,
        "Decimal32 scaleB result"
    );
    let (scaled64, scale_flags64) =
        bid754::generated::scalb64::bid64_scalbn(x64, contract.scale_exponent, 0);
    assert_eq!(scale_flags64, 0, "Decimal64 scaleB benchmark flags");
    assert_ne!(
        bid754::generated::noncomp64::bid64_is_finite(scaled64),
        0,
        "Decimal64 scaleB result"
    );
    let mut scale_flags128 = 0;
    let scaled128 = bid754::generated::bid128_misc::bid128_scalbn(
        x128,
        contract.scale_exponent,
        0,
        &mut scale_flags128,
    );
    assert_eq!(scale_flags128, 0, "Decimal128 scaleB benchmark flags");
    assert_ne!(
        bid754::generated::bid128_noncomp::bid128_is_finite(scaled128),
        0,
        "Decimal128 scaleB result"
    );

    PreparedBenchmarkInputs {
        integer_operand: contract.integer_operand,
        scale_exponent: contract.scale_exponent,
        decimal32_x_text: contract.decimal32.x,
        decimal64_x_text: contract.decimal64.x,
        decimal128_x_text: contract.decimal128.x,
        x32,
        y32,
        z32,
        x64,
        y64,
        z64,
        x128,
        y128,
        z128,
        integer32,
        integer64,
        integer128,
    }
}

fn prepare_benchmark_inputs() -> PreparedBenchmarkInputs {
    prepare_benchmark_inputs_from_contract(load_benchmark_input_contract())
}
