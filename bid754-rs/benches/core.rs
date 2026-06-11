use criterion::{black_box, criterion_group, criterion_main, Criterion};
use bid754::generated::add64::bid64_add_with_flags;
use bid754::generated::bid128_add::bid128_add;
use bid754::generated::bid128_div::bid128_div;
use bid754::generated::bid128_mul::bid128_mul;
use bid754::generated::bid128_string::{bid128_from_string, bid128_to_string};
use bid754::generated::bid32_status::{
    bid32_add_with_flags, bid32_div_with_flags, bid32_mul_with_flags,
};
use bid754::generated::bid32_string::{bid32_from_string_raw, bid32_to_string_raw};
use bid754::generated::div64::bid64_div_with_flags;
use bid754::generated::mul64::bid64_mul_with_flags;
use bid754::generated::string64::bid64_to_string;
use bid754::bid64_from_string_raw;

fn bench_bid32(c: &mut Criterion) {
    let (x, _) = bid32_from_string_raw("123.456", 0);
    let (y, _) = bid32_from_string_raw("789.012", 0);

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
    group.bench_function("parse", |b| {
        b.iter(|| black_box(bid32_from_string_raw("123.456", 0)))
    });
    group.bench_function("to_string", |b| {
        b.iter(|| black_box(bid32_to_string_raw(x)))
    });
    group.finish();
}

fn bench_bid64(c: &mut Criterion) {
    let (x, _) = bid64_from_string_raw("123456789.123456789", 0);
    let (y, _) = bid64_from_string_raw("987654321.987654321", 0);

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
    group.bench_function("parse", |b| {
        b.iter(|| black_box(bid64_from_string_raw("123456789.123456789", 0)))
    });
    group.bench_function("to_string", |b| {
        b.iter(|| black_box(bid64_to_string(x)))
    });
    group.finish();
}

fn bench_bid128(c: &mut Criterion) {
    let (x, _) = bid128_from_string("12345678901234567890.12345678901234", 0);
    let (y, _) = bid128_from_string("98765432109876543210.98765432109876", 0);

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
    group.bench_function("parse", |b| {
        b.iter(|| {
            black_box(bid128_from_string(
                "12345678901234567890.12345678901234",
                0,
            ))
        })
    });
    group.bench_function("to_string", |b| {
        b.iter(|| black_box(bid128_to_string(x)))
    });
    group.finish();
}

criterion_group!(benches, bench_bid32, bench_bid64, bench_bid128);
criterion_main!(benches);
