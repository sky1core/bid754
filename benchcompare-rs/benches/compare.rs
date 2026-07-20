// Comparative benchmarks: bid754 (generated Rust) vs rust_decimal on parse,
// to-string, parts encode/decode, and add/mul/div, at every product width
// (Decimal32/64/128).
//
// Fairness notes: each width's operands are exactly representable at that
// BID width and in rust_decimal (96-bit mantissa, scale 0..=28; the d128
// list is capped at 28 significant digits for that reason); div rows
// compare cost, not identical results (different precision/rounding rules);
// rust_decimal stores (mantissa, scale) natively, so its parts accessors
// are near-free by construction while the BID side decodes the interchange
// encoding — the parts rows expose that asymmetry rather than hiding it.

use std::str::FromStr;
use std::time::Duration;

use benchcompare::{
    div_pair_indices, pair_indices, parts_components, PARSE_INPUTS_D128, PARSE_INPUTS_D32,
    PARSE_INPUTS_D64, PARTS_ROWS_D128, PARTS_ROWS_D32, PARTS_ROWS_D64,
};
use bid754::bid_codec;
use bid754::{Decimal128, Decimal32, Decimal64};
use criterion::{black_box, criterion_group, criterion_main, Criterion};
use rust_decimal::Decimal;

macro_rules! bench_width {
    ($g:expr, $w:literal, $bid:ident, $inputs:expr, $parts:expr,
     $encode:expr, $decode_bits:expr) => {{
        let bid_vals: Vec<$bid> = $inputs.iter().map(|s| $bid::parse_raw(s).0).collect();
        let rd_vals: Vec<Decimal> = $inputs.iter().map(|s| Decimal::from_str(s).unwrap()).collect();
        let n = bid_vals.len();
        let pairs = pair_indices(n);
        let div_pairs = div_pair_indices(n);
        let mut k = 0usize;

        $g.bench_function(concat!($w, "_parse_bid754"), |b| {
            b.iter(|| {
                k = (k + 1) % n;
                black_box($bid::parse_raw(black_box($inputs[k])))
            })
        });
        $g.bench_function(concat!($w, "_parse_rust_decimal"), |b| {
            b.iter(|| {
                k = (k + 1) % n;
                black_box(Decimal::from_str(black_box($inputs[k])))
            })
        });

        $g.bench_function(concat!($w, "_to_string_bid754"), |b| {
            b.iter(|| {
                k = (k + 1) % n;
                black_box(black_box(bid_vals[k]).to_string())
            })
        });
        $g.bench_function(concat!($w, "_to_string_rust_decimal"), |b| {
            b.iter(|| {
                k = (k + 1) % n;
                black_box(black_box(rd_vals[k]).to_string())
            })
        });

        $g.bench_function(concat!($w, "_from_parts_bid754codec"), |b| {
            b.iter(|| {
                k = (k + 1) % $parts.len();
                let (m, e) = $parts[k];
                let comp = parts_components(m, e);
                black_box($encode(black_box(&comp)))
            })
        });
        $g.bench_function(concat!($w, "_from_parts_rust_decimal"), |b| {
            b.iter(|| {
                k = (k + 1) % $parts.len();
                let (m, e) = $parts[k];
                black_box(Decimal::new(black_box(m), (-e) as u32))
            })
        });

        $g.bench_function(concat!($w, "_to_parts_bid754codec"), |b| {
            b.iter(|| {
                k = (k + 1) % n;
                black_box($decode_bits(black_box(bid_vals[k])))
            })
        });
        $g.bench_function(concat!($w, "_to_parts_rust_decimal"), |b| {
            b.iter(|| {
                k = (k + 1) % n;
                let v = black_box(rd_vals[k]);
                black_box((v.mantissa(), v.scale()))
            })
        });

        $g.bench_function(concat!($w, "_add_bid754"), |b| {
            b.iter(|| {
                k = (k + 1) % pairs.len();
                let (i, j) = pairs[k];
                black_box(black_box(bid_vals[i]).add(black_box(bid_vals[j])))
            })
        });
        $g.bench_function(concat!($w, "_add_rust_decimal"), |b| {
            b.iter(|| {
                k = (k + 1) % pairs.len();
                let (i, j) = pairs[k];
                black_box(black_box(rd_vals[i]) + black_box(rd_vals[j]))
            })
        });

        $g.bench_function(concat!($w, "_mul_bid754"), |b| {
            b.iter(|| {
                k = (k + 1) % pairs.len();
                let (i, j) = pairs[k];
                black_box(black_box(bid_vals[i]).mul(black_box(bid_vals[j])))
            })
        });
        $g.bench_function(concat!($w, "_mul_rust_decimal"), |b| {
            b.iter(|| {
                k = (k + 1) % pairs.len();
                let (i, j) = pairs[k];
                black_box(black_box(rd_vals[i]) * black_box(rd_vals[j]))
            })
        });

        $g.bench_function(concat!($w, "_div_bid754"), |b| {
            b.iter(|| {
                k = (k + 1) % div_pairs.len();
                let (i, j) = div_pairs[k];
                black_box(black_box(bid_vals[i]).div(black_box(bid_vals[j])))
            })
        });
        $g.bench_function(concat!($w, "_div_rust_decimal"), |b| {
            b.iter(|| {
                k = (k + 1) % div_pairs.len();
                let (i, j) = div_pairs[k];
                black_box(black_box(rd_vals[i]) / black_box(rd_vals[j]))
            })
        });
    }};
}

fn bench_all(c: &mut Criterion) {
    let mut g = c.benchmark_group("compare");
    g.warm_up_time(Duration::from_millis(300))
        .measurement_time(Duration::from_millis(900))
        .sample_size(60);

    bench_width!(
        g, "d32", Decimal32, PARSE_INPUTS_D32, PARTS_ROWS_D32,
        |c: &bid_codec::Components| bid_codec::encode32(c),
        |v: Decimal32| bid_codec::decode32(v.to_bits())
    );
    bench_width!(
        g, "d64", Decimal64, PARSE_INPUTS_D64, PARTS_ROWS_D64,
        |c: &bid_codec::Components| bid_codec::encode64(c),
        |v: Decimal64| bid_codec::decode64(v.to_bits())
    );
    bench_width!(
        g, "d128", Decimal128, PARSE_INPUTS_D128, PARTS_ROWS_D128,
        |c: &bid_codec::Components| bid_codec::encode128(c),
        |v: Decimal128| {
            let raw = v.to_le_bytes();
            bid_codec::decode128(
                u64::from_le_bytes(raw[0..8].try_into().unwrap()),
                u64::from_le_bytes(raw[8..16].try_into().unwrap()),
            )
        }
    );

    g.finish();
}

criterion_group!(benches, bench_all);
criterion_main!(benches);
