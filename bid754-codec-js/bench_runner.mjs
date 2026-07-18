// Standalone BID codec benchmark leg (JavaScript, dependency-free).
//
// Operands come from the hand-pinned shared contract
// ../bid754-codec-go/testdata/codec_benchmark_operands.json, consumed by all
// four codec benchmark legs (Go testing.B, Rust criterion, this script, and
// bid754-codec-py/benchmarks/bench_runner.py). Setup re-checks the contract's
// exactness round trips (encode(decode(bits)) == bits, toString(decode(bits))
// == decimal_string, encode(fromString(string)) == bits) and throws on any
// violation, so the run rejects invalid or inexact operands instead of timing
// them. Every row rotates operands with an i % n counter and feeds results
// into an observable sink printed at the end, matching the other legs.
// Interpreted-leg caveat: each timed iteration pays a fixed harness dispatch
// cost (one closure call plus the sink update), uniform across rows, so
// within-leg comparisons are sound but absolute ns/op includes that constant;
// prefer the compiled legs for cross-language absolute comparisons.
// Requires the built package (npm run build); imports resolve through the
// package self-reference. Benchmark infrastructure, not a regular
// verification domain.
import { readFileSync } from "node:fs";
import {
  Kind,
  decode32,
  decode64,
  decode128,
  encode32,
  encode64,
  encode128,
  fromString,
  toString,
} from "bid754-codec";

const operandPath = new URL(
  "../bid754-codec-go/testdata/codec_benchmark_operands.json",
  import.meta.url,
);
const file = JSON.parse(readFileSync(operandPath, "utf8"));
if (file.format_version !== 1) {
  throw new Error(`benchmark operand format_version ${file.format_version}, want 1`);
}
for (const width of ["decimal32", "decimal64", "decimal128"]) {
  if (!Array.isArray(file[width]) || file[width].length === 0) {
    throw new Error(`benchmark operand contract requires non-empty ${width} entries`);
  }
}

function parseHexWord(width, index, field, hex, bits) {
  const digits = bits / 4;
  if (typeof hex !== "string" || hex.length !== digits || !/^[0-9a-f]+$/.test(hex)) {
    throw new Error(`${width}[${index}].${field}: hex ${JSON.stringify(hex)} must be exactly ${digits} lowercase hex digits`);
  }
  return BigInt(`0x${hex}`);
}

const operands32 = file.decimal32.map((entry, i) => {
  if (entry.hex_hi !== undefined && entry.hex_hi !== null) {
    throw new Error(`decimal32[${i}]: hex_hi is only valid for decimal128 entries`);
  }
  const bits = Number(parseHexWord("decimal32", i, "hex", entry.hex, 32));
  const comp = decode32(bits);
  if (comp.kind !== Kind.Normal) {
    throw new Error(`decimal32[${i}] must decode Normal`);
  }
  if (encode32(comp) !== bits) {
    throw new Error(`decimal32[${i}] operand is not canonical`);
  }
  if (toString(comp) !== entry.decimal_string) {
    throw new Error(`decimal32[${i}] canonical string mismatch`);
  }
  if (encode32(fromString(entry.decimal_string)) !== bits) {
    throw new Error(`decimal32[${i}] string operand is not exactly representable`);
  }
  return { bits, comp, string: entry.decimal_string };
});

const operands64 = file.decimal64.map((entry, i) => {
  if (entry.hex_hi !== undefined && entry.hex_hi !== null) {
    throw new Error(`decimal64[${i}]: hex_hi is only valid for decimal128 entries`);
  }
  const bits = parseHexWord("decimal64", i, "hex", entry.hex, 64);
  const comp = decode64(bits);
  if (comp.kind !== Kind.Normal) {
    throw new Error(`decimal64[${i}] must decode Normal`);
  }
  if (encode64(comp) !== bits) {
    throw new Error(`decimal64[${i}] operand is not canonical`);
  }
  if (toString(comp) !== entry.decimal_string) {
    throw new Error(`decimal64[${i}] canonical string mismatch`);
  }
  if (encode64(fromString(entry.decimal_string)) !== bits) {
    throw new Error(`decimal64[${i}] string operand is not exactly representable`);
  }
  return { bits, comp, string: entry.decimal_string };
});

const operands128 = file.decimal128.map((entry, i) => {
  const lo = parseHexWord("decimal128", i, "hex", entry.hex, 64);
  const hi = parseHexWord("decimal128", i, "hex_hi", entry.hex_hi, 64);
  const comp = decode128(lo, hi);
  if (comp.kind !== Kind.Normal) {
    throw new Error(`decimal128[${i}] must decode Normal`);
  }
  const [encodedLo, encodedHi] = encode128(comp);
  if (encodedLo !== lo || encodedHi !== hi) {
    throw new Error(`decimal128[${i}] operand is not canonical`);
  }
  if (toString(comp) !== entry.decimal_string) {
    throw new Error(`decimal128[${i}] canonical string mismatch`);
  }
  const [reLo, reHi] = encode128(fromString(entry.decimal_string));
  if (reLo !== lo || reHi !== hi) {
    throw new Error(`decimal128[${i}] string operand is not exactly representable`);
  }
  return { lo, hi, comp, string: entry.decimal_string };
});

// Two sinks so the timed loops never pay Number<->BigInt conversion cost:
// bigint results XOR into sinkBig, small numeric observations XOR into
// sinkNum. Both are printed at the end so every result stays observable.
let sinkBig = 0n;
let sinkNum = 0;

function sinkComponents(comp) {
  sinkBig ^= comp.coefficient ^ comp.payload;
  sinkNum ^= comp.exponent ^ comp.kind ^ (comp.sign ? 1 : 0);
}

function sinkString(s) {
  sinkNum ^= s.length;
}

const SAMPLES = Number.parseInt(process.argv[2] ?? "5", 10);
const TARGET_SAMPLE_NS = 100_000_000n; // ~100ms per timed sample
const CALIBRATION_ITERS = 2_000;

function runBatch(fn, iters) {
  const start = process.hrtime.bigint();
  for (let i = 0; i < iters; i++) {
    fn(i);
  }
  return process.hrtime.bigint() - start;
}

function benchRow(name, fn) {
  // Warmup plus calibration: size the timed batch to ~TARGET_SAMPLE_NS.
  let calibrated = runBatch(fn, CALIBRATION_ITERS);
  if (calibrated <= 0n) {
    calibrated = 1n;
  }
  let iters = Number((TARGET_SAMPLE_NS * BigInt(CALIBRATION_ITERS)) / calibrated);
  iters = Math.max(1_000, Math.min(20_000_000, iters));
  const samples = [];
  for (let s = 0; s < SAMPLES; s++) {
    const elapsed = runBatch(fn, iters);
    samples.push(Number(elapsed) / iters);
  }
  const sorted = [...samples].sort((a, b) => a - b);
  const median = sorted[Math.floor(sorted.length / 2)];
  const rendered = samples.map((v) => v.toFixed(1)).join(",");
  console.log(`BENCH ${name} ns_op_median=${median.toFixed(1)} iters=${iters} samples_ns_op=[${rendered}]`);
}

const n32 = operands32.length;
const n64 = operands64.length;
const n128 = operands128.length;

console.log(`BENCH-LEG codec-js node=${process.version} samples=${SAMPLES} operands=${n32}/${n64}/${n128}`);

benchRow("codec_bid32/decode", (i) => {
  sinkComponents(decode32(operands32[i % n32].bits));
});
benchRow("codec_bid32/encode", (i) => {
  sinkNum ^= encode32(operands32[i % n32].comp);
});
benchRow("codec_bid32/to_string", (i) => {
  sinkString(toString(operands32[i % n32].comp));
});
benchRow("codec_bid32/from_string", (i) => {
  sinkComponents(fromString(operands32[i % n32].string));
});

benchRow("codec_bid64/decode", (i) => {
  sinkComponents(decode64(operands64[i % n64].bits));
});
benchRow("codec_bid64/encode", (i) => {
  sinkBig ^= encode64(operands64[i % n64].comp);
});
benchRow("codec_bid64/to_string", (i) => {
  sinkString(toString(operands64[i % n64].comp));
});
benchRow("codec_bid64/from_string", (i) => {
  sinkComponents(fromString(operands64[i % n64].string));
});

benchRow("codec_bid128/decode", (i) => {
  sinkComponents(decode128(operands128[i % n128].lo, operands128[i % n128].hi));
});
benchRow("codec_bid128/encode", (i) => {
  const [lo, hi] = encode128(operands128[i % n128].comp);
  sinkBig ^= lo ^ hi;
});
benchRow("codec_bid128/to_string", (i) => {
  sinkString(toString(operands128[i % n128].comp));
});
benchRow("codec_bid128/from_string", (i) => {
  sinkComponents(fromString(operands128[i % n128].string));
});

console.log(`BENCH-SINK ${sinkBig}/${sinkNum}`);
