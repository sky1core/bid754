use bid754::bid64_from_string_raw;
use bid754::generated::add64::bid64_add_with_flags;
use bid754::generated::bid128_add::bid128_add;
use bid754::generated::bid128_div::bid128_div;
use bid754::generated::bid128_fma::bid128_fma;
use bid754::generated::bid128_mul::bid128_mul;
use bid754::generated::bid128_noncomp::bid128_is_finite;
use bid754::generated::bid128_sqrt::bid128_sqrt;
use bid754::generated::bid128_string::{bid128_from_string, bid128_to_string};
use bid754::generated::bid32_fma::bid32_fma;
use bid754::generated::bid32_noncomp::bid32_is_finite;
use bid754::generated::bid32_sqrt::bid32_sqrt;
use bid754::generated::bid32_status::{
    bid32_add_with_flags, bid32_div_with_flags, bid32_mul_with_flags,
};
use bid754::generated::bid32_string::{bid32_from_string_raw, bid32_to_string_raw};
use bid754::generated::div64::bid64_div_with_flags;
use bid754::generated::fma64::bid64_fma;
use bid754::generated::mul64::bid64_mul_with_flags;
use bid754::generated::noncomp64::bid64_is_finite;
use bid754::generated::sqrt64::bid64_sqrt;
use bid754::generated::string64::bid64_to_string;
use criterion::{black_box, criterion_group, criterion_main, Criterion};
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
    let inputs: BenchmarkInputs = serde_json::from_str(include_str!(
        "../../bid754-go/testdata/benchmark_inputs.json"
    ))
    .expect("parse benchmark input contract");
    assert_eq!(inputs.format_version, 1, "benchmark input format_version");
    for (name, pair) in [
        ("decimal32", &inputs.decimal32),
        ("decimal64", &inputs.decimal64),
        ("decimal128", &inputs.decimal128),
    ] {
        assert!(
            !pair.x.is_empty() && !pair.y.is_empty() && !pair.z.is_empty(),
            "benchmark input {name} requires non-empty x, y, and z"
        );
    }
    inputs
}

fn exact_bid32(input: &str) -> u32 {
    let (value, flags) = bid32_from_string_raw(input, 0);
    assert_eq!(flags, 0, "Decimal32 benchmark input flags");
    assert_ne!(
        bid32_is_finite(value),
        0,
        "Decimal32 benchmark input must be finite"
    );
    value
}

fn exact_bid64(input: &str) -> u64 {
    let (value, flags) = bid64_from_string_raw(input, 0);
    assert_eq!(flags, 0, "Decimal64 benchmark input flags");
    assert_ne!(
        bid64_is_finite(value),
        0,
        "Decimal64 benchmark input must be finite"
    );
    value
}

fn exact_bid128(input: &str) -> bid754::gen_types::BID_UINT128 {
    let (value, flags) = bid128_from_string(input, 0);
    assert_eq!(flags, 0, "Decimal128 benchmark input flags");
    assert_ne!(
        bid128_is_finite(value),
        0,
        "Decimal128 benchmark input must be finite"
    );
    value
}

fn bench_bid32(c: &mut Criterion) {
    let inputs = benchmark_inputs();
    let x = exact_bid32(&inputs.decimal32.x);
    let y = exact_bid32(&inputs.decimal32.y);
    let z = exact_bid32(&inputs.decimal32.z);

    let mut group = c.benchmark_group("bid32");
    group.bench_function("add", |b| {
        b.iter(|| {
            let (got, flags) = bid32_add_with_flags(black_box(x), black_box(y), 0);
            black_box((got, flags))
        })
    });
    group.bench_function("mul", |b| {
        b.iter(|| {
            let (got, flags) = bid32_mul_with_flags(black_box(x), black_box(y), 0);
            black_box((got, flags))
        })
    });
    group.bench_function("div", |b| {
        b.iter(|| {
            let (got, flags) = bid32_div_with_flags(black_box(x), black_box(y), 0);
            black_box((got, flags))
        })
    });
    group.bench_function("fma", |b| {
        b.iter(|| {
            let (got, flags) = bid32_fma(black_box(x), black_box(y), black_box(z), 0);
            black_box((got, flags))
        })
    });
    group.bench_function("sqrt", |b| {
        b.iter(|| {
            let (got, flags) = bid32_sqrt(black_box(x), 0);
            black_box((got, flags))
        })
    });
    group.bench_function("parse", |b| {
        b.iter(|| black_box(bid32_from_string_raw(inputs.decimal32.x.as_str(), 0)))
    });
    group.bench_function("to_string", |b| {
        b.iter(|| black_box(bid32_to_string_raw(x)))
    });
    group.finish();
}

fn bench_bid64(c: &mut Criterion) {
    let inputs = benchmark_inputs();
    let x = exact_bid64(&inputs.decimal64.x);
    let y = exact_bid64(&inputs.decimal64.y);
    let z = exact_bid64(&inputs.decimal64.z);

    let mut group = c.benchmark_group("bid64");
    group.bench_function("add", |b| {
        b.iter(|| {
            let (got, flags) = bid64_add_with_flags(black_box(x), black_box(y), 0);
            black_box((got, flags))
        })
    });
    group.bench_function("mul", |b| {
        b.iter(|| {
            let (got, flags) = bid64_mul_with_flags(black_box(x), black_box(y), 0);
            black_box((got, flags))
        })
    });
    group.bench_function("div", |b| {
        b.iter(|| {
            let (got, flags) = bid64_div_with_flags(black_box(x), black_box(y), 0);
            black_box((got, flags))
        })
    });
    group.bench_function("fma", |b| {
        b.iter(|| {
            let (got, flags) = bid64_fma(black_box(x), black_box(y), black_box(z), 0);
            black_box((got, flags))
        })
    });
    group.bench_function("sqrt", |b| {
        b.iter(|| {
            let (got, flags) = bid64_sqrt(black_box(x), 0);
            black_box((got, flags))
        })
    });
    group.bench_function("parse", |b| {
        b.iter(|| black_box(bid64_from_string_raw(inputs.decimal64.x.as_str(), 0)))
    });
    group.bench_function("to_string", |b| b.iter(|| black_box(bid64_to_string(x))));
    group.finish();
}

fn bench_bid128(c: &mut Criterion) {
    let inputs = benchmark_inputs();
    let x = exact_bid128(&inputs.decimal128.x);
    let y = exact_bid128(&inputs.decimal128.y);
    let z = exact_bid128(&inputs.decimal128.z);

    let mut group = c.benchmark_group("bid128");
    group.bench_function("add", |b| {
        b.iter(|| {
            let mut flags = 0u32;
            let got = bid128_add(black_box(x), black_box(y), 0, &mut flags);
            black_box((got, flags))
        })
    });
    group.bench_function("mul", |b| {
        b.iter(|| {
            let (got, flags) = bid128_mul(black_box(x), black_box(y), 0);
            black_box((got, flags))
        })
    });
    group.bench_function("div", |b| {
        b.iter(|| {
            let (got, flags) = bid128_div(black_box(x), black_box(y), 0);
            black_box((got, flags))
        })
    });
    group.bench_function("fma", |b| {
        b.iter(|| {
            let (got, flags) = bid128_fma(black_box(x), black_box(y), black_box(z), 0);
            black_box((got, flags))
        })
    });
    group.bench_function("sqrt", |b| {
        b.iter(|| {
            let (got, flags) = bid128_sqrt(black_box(x), 0);
            black_box((got, flags))
        })
    });
    group.bench_function("parse", |b| {
        b.iter(|| black_box(bid128_from_string(inputs.decimal128.x.as_str(), 0)))
    });
    group.bench_function("to_string", |b| b.iter(|| black_box(bid128_to_string(x))));
    group.finish();
}

criterion_group!(benches, bench_bid32, bench_bid64, bench_bid128);
criterion_main!(benches);
