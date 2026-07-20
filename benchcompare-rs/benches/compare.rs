// Comparative benchmarks: bid754 (generated Rust) vs
// rust_decimal on parse, to-string, parts encode/decode, add/mul/div.
//
// Fairness notes: operands are exactly representable in Decimal64 (<= 16
// significant digits) and in rust_decimal (96-bit, scale 0..=28); div rows
// compare cost, not identical results (different precision/rounding rules).

use std::str::FromStr;
use std::time::Duration;

use bid754::bid_codec;
use bid754::Decimal64;
use benchcompare::{parts_components, PARSE_INPUTS, PARTS_ROWS};
use criterion::{black_box, criterion_group, criterion_main, Criterion};
use rust_decimal::Decimal;

fn bench_all(c: &mut Criterion) {
    let bid_vals: Vec<Decimal64> = PARSE_INPUTS
        .iter()
        .map(|s| Decimal64::parse_raw(s).0)
        .collect();
    let rd_vals: Vec<Decimal> = PARSE_INPUTS
        .iter()
        .map(|s| Decimal::from_str(s).unwrap())
        .collect();
    let n = bid_vals.len();
    let pairs: Vec<(usize, usize)> = (0..n).map(|i| (i, (i * 7 + 3) % n)).collect();
    let div_pairs: Vec<(usize, usize)> = (0..n)
        .map(|i| {
            let mut j = (i * 5 + 1) % n;
            if rd_vals[j].is_zero() {
                j = (j + 1) % n;
            }
            (i, j)
        })
        .collect();
    let bid_bits: Vec<u64> = bid_vals.iter().map(|d| d.to_bits()).collect();

    let mut g = c.benchmark_group("compare");
    g.warm_up_time(Duration::from_millis(300))
        .measurement_time(Duration::from_millis(900))
        .sample_size(60);

    let mut k = 0usize;
    g.bench_function("parse_bid754", |b| {
        b.iter(|| {
            k = (k + 1) % n;
            black_box(Decimal64::parse_raw(black_box(PARSE_INPUTS[k])))
        })
    });
    g.bench_function("parse_rust_decimal", |b| {
        b.iter(|| {
            k = (k + 1) % n;
            black_box(Decimal::from_str(black_box(PARSE_INPUTS[k])))
        })
    });

    g.bench_function("to_string_bid754", |b| {
        b.iter(|| {
            k = (k + 1) % n;
            black_box(black_box(bid_vals[k]).to_string())
        })
    });
    g.bench_function("to_string_rust_decimal", |b| {
        b.iter(|| {
            k = (k + 1) % n;
            black_box(black_box(rd_vals[k]).to_string())
        })
    });

    g.bench_function("from_parts_bid754codec", |b| {
        b.iter(|| {
            k = (k + 1) % PARTS_ROWS.len();
            let (m, e) = PARTS_ROWS[k];
            let comp = parts_components(m, e);
            black_box(bid_codec::encode64(black_box(&comp)))
        })
    });
    g.bench_function("from_parts_rust_decimal", |b| {
        b.iter(|| {
            k = (k + 1) % PARTS_ROWS.len();
            let (m, e) = PARTS_ROWS[k];
            black_box(Decimal::new(black_box(m), (-e) as u32))
        })
    });

    g.bench_function("to_parts_bid754codec", |b| {
        b.iter(|| {
            k = (k + 1) % n;
            black_box(bid_codec::decode64(black_box(bid_bits[k])))
        })
    });
    g.bench_function("to_parts_rust_decimal", |b| {
        b.iter(|| {
            k = (k + 1) % n;
            let v = black_box(rd_vals[k]);
            black_box((v.mantissa(), v.scale()))
        })
    });

    g.bench_function("add_bid754", |b| {
        b.iter(|| {
            k = (k + 1) % pairs.len();
            let (i, j) = pairs[k];
            black_box(black_box(bid_vals[i]).add(black_box(bid_vals[j])))
        })
    });
    g.bench_function("add_rust_decimal", |b| {
        b.iter(|| {
            k = (k + 1) % pairs.len();
            let (i, j) = pairs[k];
            black_box(black_box(rd_vals[i]) + black_box(rd_vals[j]))
        })
    });

    g.bench_function("mul_bid754", |b| {
        b.iter(|| {
            k = (k + 1) % pairs.len();
            let (i, j) = pairs[k];
            black_box(black_box(bid_vals[i]).mul(black_box(bid_vals[j])))
        })
    });
    g.bench_function("mul_rust_decimal", |b| {
        b.iter(|| {
            k = (k + 1) % pairs.len();
            let (i, j) = pairs[k];
            black_box(black_box(rd_vals[i]) * black_box(rd_vals[j]))
        })
    });

    g.bench_function("div_bid754", |b| {
        b.iter(|| {
            k = (k + 1) % div_pairs.len();
            let (i, j) = div_pairs[k];
            black_box(black_box(bid_vals[i]).div(black_box(bid_vals[j])))
        })
    });
    g.bench_function("div_rust_decimal", |b| {
        b.iter(|| {
            k = (k + 1) % div_pairs.len();
            let (i, j) = div_pairs[k];
            black_box(black_box(rd_vals[i]) / black_box(rd_vals[j]))
        })
    });

    g.finish();
}

criterion_group!(benches, bench_all);
criterion_main!(benches);
