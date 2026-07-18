//! Standalone BID codec benchmark leg (Rust, criterion).
//!
//! Operands come from the hand-pinned shared contract
//! `bid754-codec-go/testdata/codec_benchmark_operands.json`, consumed by all
//! four codec benchmark legs (Go `testing.B`, this criterion leg, and the
//! dependency-free JS/Python scripts). `benchmark_operands` re-checks the
//! contract's exactness round trips (encode(decode(bits)) == bits,
//! to_string(decode(bits)) == decimal_string, encode(from_string(string)) ==
//! bits) and panics on any violation, so benchmark setup rejects invalid or
//! inexact operands instead of timing them. Benchmark infrastructure only;
//! not a regular verification domain.

use bid754_codec::{
    decode128, decode32, decode64, encode128, encode32, encode64, from_string, to_string,
    Components, Kind,
};
use criterion::{black_box, criterion_group, criterion_main, Criterion};
use serde::Deserialize;

#[derive(Deserialize)]
struct OperandEntry {
    hex: String,
    #[serde(default)]
    hex_hi: Option<String>,
    decimal_string: String,
}

#[derive(Deserialize)]
struct OperandFile {
    format_version: u32,
    decimal32: Vec<OperandEntry>,
    decimal64: Vec<OperandEntry>,
    decimal128: Vec<OperandEntry>,
}

struct Operand32 {
    bits: u32,
    comp: Components,
    string: String,
}

struct Operand64 {
    bits: u64,
    comp: Components,
    string: String,
}

struct Operand128 {
    lo: u64,
    hi: u64,
    comp: Components,
    string: String,
}

struct Operands {
    decimal32: Vec<Operand32>,
    decimal64: Vec<Operand64>,
    decimal128: Vec<Operand128>,
}

fn parse_hex_word(width: &str, index: usize, field: &str, hex: &str, bits: u32) -> u64 {
    let digits = (bits / 4) as usize;
    // Lowercase-only, matching the Go/JS/Python legs exactly, so all four
    // legs accept and reject the same contract file (from_str_radix alone
    // would also take uppercase digits and a leading sign).
    assert!(
        hex.len() == digits && hex.bytes().all(|b| matches!(b, b'0'..=b'9' | b'a'..=b'f')),
        "{width}[{index}].{field}: hex {hex:?} must be exactly {digits} lowercase hex digits"
    );
    u64::from_str_radix(hex, 16)
        .unwrap_or_else(|err| panic!("{width}[{index}].{field}: invalid hex {hex:?}: {err}"))
}

fn benchmark_operands() -> Operands {
    let file: OperandFile = serde_json::from_str(include_str!(
        "../../bid754-codec-go/testdata/codec_benchmark_operands.json"
    ))
    .expect("parse benchmark operand contract");
    assert_eq!(file.format_version, 1, "benchmark operand format_version");
    assert!(
        !file.decimal32.is_empty() && !file.decimal64.is_empty() && !file.decimal128.is_empty(),
        "benchmark operand contract requires non-empty decimal32/decimal64/decimal128 entry lists"
    );

    let decimal32 = file
        .decimal32
        .iter()
        .enumerate()
        .map(|(i, entry)| {
            assert!(
                entry.hex_hi.is_none(),
                "decimal32[{i}]: hex_hi is only valid for decimal128 entries"
            );
            let bits = parse_hex_word("decimal32", i, "hex", &entry.hex, 32) as u32;
            let comp = decode32(bits);
            assert_eq!(comp.kind, Kind::Normal, "decimal32[{i}] must decode Normal");
            let encoded = encode32(&comp).expect("re-encode decoded decimal32 operand");
            assert_eq!(encoded, bits, "decimal32[{i}] operand is not canonical");
            let rendered = to_string(&comp).expect("render decimal32 operand");
            assert_eq!(
                rendered, entry.decimal_string,
                "decimal32[{i}] canonical string mismatch"
            );
            let parsed = from_string(&entry.decimal_string).expect("parse decimal32 operand string");
            let reencoded =
                encode32(&parsed).expect("decimal32 string operand must be exactly representable");
            assert_eq!(reencoded, bits, "decimal32[{i}] string round trip drifts");
            Operand32 {
                bits,
                comp,
                string: entry.decimal_string.clone(),
            }
        })
        .collect();

    let decimal64 = file
        .decimal64
        .iter()
        .enumerate()
        .map(|(i, entry)| {
            assert!(
                entry.hex_hi.is_none(),
                "decimal64[{i}]: hex_hi is only valid for decimal128 entries"
            );
            let bits = parse_hex_word("decimal64", i, "hex", &entry.hex, 64);
            let comp = decode64(bits);
            assert_eq!(comp.kind, Kind::Normal, "decimal64[{i}] must decode Normal");
            let encoded = encode64(&comp).expect("re-encode decoded decimal64 operand");
            assert_eq!(encoded, bits, "decimal64[{i}] operand is not canonical");
            let rendered = to_string(&comp).expect("render decimal64 operand");
            assert_eq!(
                rendered, entry.decimal_string,
                "decimal64[{i}] canonical string mismatch"
            );
            let parsed = from_string(&entry.decimal_string).expect("parse decimal64 operand string");
            let reencoded =
                encode64(&parsed).expect("decimal64 string operand must be exactly representable");
            assert_eq!(reencoded, bits, "decimal64[{i}] string round trip drifts");
            Operand64 {
                bits,
                comp,
                string: entry.decimal_string.clone(),
            }
        })
        .collect();

    let decimal128 = file
        .decimal128
        .iter()
        .enumerate()
        .map(|(i, entry)| {
            let lo = parse_hex_word("decimal128", i, "hex", &entry.hex, 64);
            let hi_hex = entry
                .hex_hi
                .as_deref()
                .unwrap_or_else(|| panic!("decimal128[{i}]: missing hex_hi"));
            let hi = parse_hex_word("decimal128", i, "hex_hi", hi_hex, 64);
            let comp = decode128(lo, hi);
            assert_eq!(comp.kind, Kind::Normal, "decimal128[{i}] must decode Normal");
            let (encoded_lo, encoded_hi) =
                encode128(&comp).expect("re-encode decoded decimal128 operand");
            assert_eq!(
                (encoded_lo, encoded_hi),
                (lo, hi),
                "decimal128[{i}] operand is not canonical"
            );
            let rendered = to_string(&comp).expect("render decimal128 operand");
            assert_eq!(
                rendered, entry.decimal_string,
                "decimal128[{i}] canonical string mismatch"
            );
            let parsed =
                from_string(&entry.decimal_string).expect("parse decimal128 operand string");
            let (reencoded_lo, reencoded_hi) =
                encode128(&parsed).expect("decimal128 string operand must be exactly representable");
            assert_eq!(
                (reencoded_lo, reencoded_hi),
                (lo, hi),
                "decimal128[{i}] string round trip drifts"
            );
            Operand128 {
                lo,
                hi,
                comp,
                string: entry.decimal_string.clone(),
            }
        })
        .collect();

    Operands {
        decimal32,
        decimal64,
        decimal128,
    }
}

/// Registers one benchmark row that rotates through the shared operand
/// entries with an index counter (`i % n`), matching the Go `testing.B` leg's
/// per-iteration operand selection, and consumes the result through
/// `black_box` (operand inputs are `black_box`ed inside each row's closure).
fn bench_rotating<T, R>(c: &mut Criterion, name: &str, entries: &[T], mut op: impl FnMut(&T) -> R) {
    let n = entries.len();
    let mut i = 0usize;
    c.bench_function(name, |b| {
        b.iter(|| {
            let entry = &entries[i % n];
            i += 1;
            black_box(op(entry))
        })
    });
}

fn codec_benchmarks(c: &mut Criterion) {
    let ops = benchmark_operands();

    let entries = &ops.decimal32;
    bench_rotating(c, "codec_bid32/decode", entries, |e| {
        decode32(black_box(e.bits))
    });
    bench_rotating(c, "codec_bid32/encode", entries, |e| {
        encode32(black_box(&e.comp))
    });
    bench_rotating(c, "codec_bid32/to_string", entries, |e| {
        to_string(black_box(&e.comp))
    });
    bench_rotating(c, "codec_bid32/from_string", entries, |e| {
        from_string(black_box(e.string.as_str()))
    });

    let entries = &ops.decimal64;
    bench_rotating(c, "codec_bid64/decode", entries, |e| {
        decode64(black_box(e.bits))
    });
    bench_rotating(c, "codec_bid64/encode", entries, |e| {
        encode64(black_box(&e.comp))
    });
    bench_rotating(c, "codec_bid64/to_string", entries, |e| {
        to_string(black_box(&e.comp))
    });
    bench_rotating(c, "codec_bid64/from_string", entries, |e| {
        from_string(black_box(e.string.as_str()))
    });

    let entries = &ops.decimal128;
    bench_rotating(c, "codec_bid128/decode", entries, |e| {
        decode128(black_box(e.lo), black_box(e.hi))
    });
    bench_rotating(c, "codec_bid128/encode", entries, |e| {
        encode128(black_box(&e.comp))
    });
    bench_rotating(c, "codec_bid128/to_string", entries, |e| {
        to_string(black_box(&e.comp))
    });
    bench_rotating(c, "codec_bid128/from_string", entries, |e| {
        from_string(black_box(e.string.as_str()))
    });
}

criterion_group!(benches, codec_benchmarks);
criterion_main!(benches);
