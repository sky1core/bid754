import assert from 'node:assert/strict';
import { spawnSync } from 'node:child_process';
import { fileURLToPath } from 'node:url';
import * as codec from './dist/esm/index.js';

const widths = [
  { bits: 32, p: 7, qmin: -101, qmax: 90 },
  { bits: 64, p: 16, qmin: -398, qmax: 369 },
  { bits: 128, p: 34, qmin: -6176, qmax: 6111 },
];
const lengths = [34, 35, 6176, 6177, 61760];
const nativeBigInt = globalThis.BigInt;
const nativeNumber = globalThis.Number;
let cases = 0;
let active = null;
const stats = { coefficient: 0, payload: 0, exponent: 0, bigintCalls: 0, numberCalls: 0 };

function installConversionAccounting() {
  globalThis.BigInt = new Proxy(nativeBigInt, {
    apply(target, receiver, args) {
      if (active !== null) {
        active.bigintCalls++;
        stats.bigintCalls++;
        const value = args[0];
        if (typeof value !== 'string' || !/^[0-9]+$/.test(value) || value.length > active.limit) {
          active.violation = 'BigInt received a non-decimal or oversized argument';
        } else {
          stats[active.field] = Math.max(stats[active.field], value.length);
        }
      }
      return Reflect.apply(target, receiver, args);
    },
  });
  globalThis.Number = new Proxy(nativeNumber, {
    apply(target, receiver, args) {
      if (active !== null && typeof args[0] === 'string') {
        active.numberCalls++;
        stats.numberCalls++;
        const value = args[0];
        const digits = value[0] === '+' || value[0] === '-' ? value.length - 1 : value.length;
        if (!/^[+-]?[0-9]+$/.test(value) || digits > 16) active.violation = 'Number received more than 16 exponent digits';
        stats.exponent = Math.max(stats.exponent, digits);
      }
      return Reflect.apply(target, receiver, args);
    },
  });
}
function components(coefficient = 0n, exponent = 0, sign = false, kind = coefficient === 0n ? codec.Kind.Zero : codec.Kind.Normal, payload = 0n) {
  return { sign, kind, coefficient, exponent, payload };
}
function nan(payload = 0n, sign = false, kind = codec.Kind.QNaN) {
  return components(0n, 0, sign, kind, payload);
}
function boundedError(error, name) {
  assert.ok(error instanceof Error, `${name}: exception type`);
  assert.ok(Buffer.byteLength(error.message, 'utf8') <= 256, `${name}: error exceeds 256 bytes`);
}
function parseCase(test) {
  const { name, input, expected, errorPart = '', field = 'coefficient' } = test;
  active = { field, limit: field === 'payload' ? 33 : 34, bigintCalls: 0, numberCalls: 0, violation: null };
  let value;
  let failure;
  const audit = active;
  try { value = codec.fromString(input); } catch (error) { failure = error; } finally { active = null; }
  assert.equal(audit.violation, null, `${name}: ${audit.violation}`);
  assert.ok(audit.bigintCalls <= 1, `${name}: more than one BigInt conversion`);
  assert.ok(audit.numberCalls <= 1, `${name}: repeated exponent conversions`);
  if (expected === null) {
    assert.notEqual(failure, undefined, `${name}: accepted malformed input`);
    boundedError(failure, name);
    assert.ok(failure.message.includes(errorPart), `${name}: wrong rejection boundary`);
  } else {
    assert.equal(failure, undefined, `${name}: unexpected error ${failure?.message}`);
    assert.deepEqual(value, expected, `${name}: parsed fields`);
    assert.deepEqual(codec.fromString(codec.toString(value)), expected, `${name}: cohort closure`);
  }
  cases++;
}
function widthCase(width, value, valid) {
  const name = `bid${width.bits} width boundary`;
  let decoded;
  let failure;
  try {
    const words = codec[`encode${width.bits}`](value);
    decoded = width.bits === 128 ? codec.decode128(...words) : codec[`decode${width.bits}`](words);
  } catch (error) { failure = error; }
  if (valid) {
    assert.equal(failure, undefined, `${name}: unexpected rejection`);
    assert.deepEqual(decoded, value, `${name}: decoded fields`);
  } else {
    assert.notEqual(failure, undefined, `${name}: encoded unrepresentable input`);
    boundedError(failure, name);
  }
  cases++;
}
function widthBoundaries() {
  for (const width of widths) {
    for (const digits of [width.p - 1, width.p, width.p + 1]) {
      for (const q of [width.qmin - 1, width.qmin, width.qmax, width.qmax + 1]) {
        const text = '9'.repeat(digits);
        const value = components(nativeBigInt(text), q);
        parseCase({ name: `width digits=${digits} q=${q}`, input: `${text}e${q}`, expected: digits <= 34 ? value : null });
        widthCase(width, value, digits <= width.p && q >= width.qmin && q <= width.qmax);
      }
    }
    for (const q of [width.qmin - 1, width.qmin, width.qmax, width.qmax + 1]) {
      const value = components(0n, q, true);
      parseCase({ name: 'signed zero width', input: `-0e${q}`, expected: value });
      widthCase(width, value, q >= width.qmin && q <= width.qmax);
    }
    for (const digits of [width.p - 2, width.p - 1, width.p]) {
      const text = '9'.repeat(digits);
      const value = nan(nativeBigInt(text), true, codec.Kind.SNaN);
      parseCase({ name: 'payload width', input: `-sNaN${text}`, expected: digits <= 33 ? value : null, field: 'payload' });
      widthCase(width, value, digits < width.p);
    }
  }
}
function parserCases() {
  const tests = [];
  const add = (name, input, expected, field = 'coefficient', errorPart = '') => tests.push({ name, input, expected, field, errorPart });
  add('schema34', '9'.repeat(34), components(nativeBigInt('9'.repeat(34))));
  add('schema35', '9'.repeat(35), null);
  add('payload33', 'NaN' + '9'.repeat(33), nan(nativeBigInt('9'.repeat(33))), 'payload');
  add('payload34', 'NaN' + '9'.repeat(34), null, 'payload');
  for (const q of [-2147483649, -2147483648, -2147483647, 2147483646, 2147483647, 2147483648]) {
    add(`int32 ${q}`, `1e${q}`, q >= -2147483648 && q <= 2147483647 ? components(1n, q) : null);
  }
  for (const sign of ['', '-']) {
    add('literal inside', `1e${sign}9007199254740991`, null, 'coefficient', 'int32');
    add('literal at bound', `1e${sign}9007199254740992`, null, 'coefficient', '2^53');
    add('literal outside', `1e${sign}9007199254740993`, null, 'coefficient', '2^53');
  }
  add('fraction upper inside', '1.0e2147483648', components(10n, 2147483647));
  add('fraction upper outside', '1.0e2147483649', null);
  add('fraction lower inside', '1.0e-2147483647', components(10n, -2147483648));
  add('fraction lower outside', '1.0e-2147483648', null);
  add('cohort', '-0001.100', components(1100n, -3, true));
  add('trim6', '\t\n\v\f\r -iNfInItY \t\n\v\f\r', components(0n, 0, true, codec.Kind.Infinity));
  add('negative exponent zero', '-0e-0', components(0n, 0, true));
  for (const input of ['', '+', '-', '.', 'e1', '1e', '1e+', '1e-', '1e+-1', '1.2.3', '1e1.0',
    '1 0', 'NaN+1', 'SNaN-1', 'Infinityx', '\x001', '1\x1f', '1_0', '0x10', '\u00a01', '１', '1\x7f', '1\n0']) {
    add('grammar', input, null, /nan/i.test(input) ? 'payload' : 'coefficient');
  }
  for (const n of lengths) {
    const zeros = '0'.repeat(n);
    add(`leading coefficient ${n}`, ` -${zeros}1200 `, components(1200n, 0, true));
    add(`leading payload ${n}`, ` -sNaN${zeros}12 `, nan(12n, true, codec.Kind.SNaN), 'payload');
    add(`leading exponent ${n}`, `1e-${zeros}12`, components(1n, -12));
    add(`leading fraction ${n}`, `0.${zeros}1200e${n + 4}`, components(1200n));
    add(`allzero ${n}`, `-${zeros}`, components(0n, 0, true));
    add(`fractionzero ${n}`, `-0.${zeros}`, components(0n, -n, true));
    add(`payloadzero ${n}`, `NaN${zeros}`, nan(), 'payload');
    add(`exponentzero ${n}`, `-0e-${zeros}`, components(0n, 0, true));
    add(`oversize coefficient ${n}`, '9'.repeat(n), n <= 34 ? components(nativeBigInt('9'.repeat(n))) : null);
    add(`oversize payload ${n}`, 'NaN' + '9'.repeat(n), null, 'payload');
    add(`oversize exponent ${n}`, '1e' + '9'.repeat(n), null);
    add(`long trim ${n}`, ' '.repeat(n) + '1' + '\t'.repeat(n), components(1n));
    for (const tail of ['x', 'é', '\n0']) {
      for (const prefix of ['NaN', '1e', '1.', '']) add(`malformed tail ${n}`, prefix + zeros + tail, null, prefix === 'NaN' ? 'payload' : 'coefficient');
    }
  }
  return tests;
}
function boundedHeapChild() {
  const stressDigits = 4 * 1024 * 1024;
  const shapes = ['coefficient', 'payload', 'exponent', 'fraction', 'zero', 'malformed'];
  for (let i = 0; i < 4096; i++) codec.fromString('001.20e-2');
  for (const shape of shapes) {
    let input;
    let expected;
    const zeros = '0'.repeat(stressDigits);
    if (shape === 'coefficient') { input = ` -${zeros}1200 `; expected = components(1200n, 0, true); }
    if (shape === 'payload') { input = ` -sNaN${zeros}12 `; expected = nan(12n, true, codec.Kind.SNaN); }
    if (shape === 'exponent') { input = `1e-${zeros}12`; expected = components(1n, -12); }
    if (shape === 'fraction') { input = `0.${zeros}1200e${stressDigits + 4}`; expected = components(1200n); }
    if (shape === 'zero') { input = `-0.${zeros}`; expected = components(0n, -stressDigits, true); }
    if (shape === 'malformed') { input = `NaN${zeros}x`; expected = null; }
    input = Buffer.from(input, 'ascii').toString('ascii');
    const test = { name: `boundedheap ${shape}`, input, expected, field: shape === 'payload' || shape === 'malformed' ? 'payload' : 'coefficient' };
    for (let i = 0; i < 2; i++) {
      try { codec.fromString(input); } catch (error) { assert.equal(expected, null); boundedError(error, shape); }
    }
    for (let i = 0; i < 4; i++) parseCase(test);
  }
  console.log(`MEMORY javascript method=boundedheap_child old_space_MiB=32 semi_space_MiB=1 stress_digits=${stressDigits} shapes=${shapes.length} repetitions=4 warmup=4096_short_plus_2_per_shape input_generation=before_asserted_calls input_flattening=before_warmup budget_basis=runtime_plus_4MiB_input_headroom_excludes_input_sized_digit_ropes limitation=live_heap_budget_not_total_allocation_bound`);
  console.log(`HEAP-CASES ${cases}`);
}

installConversionAccounting();
try {
  if (process.argv.includes('--heap-child')) {
    boundedHeapChild();
  } else {
    widthBoundaries();
    const tests = parserCases();
    for (let i = 0; i < 4096; i++) codec.fromString('001.20e-2');
    for (const test of tests) parseCase(test);
    const child = spawnSync(process.execPath, ['--max-old-space-size=32', '--max-semi-space-size=1', fileURLToPath(import.meta.url), '--heap-child'], { encoding: 'utf8', maxBuffer: 1024 * 1024 });
    assert.equal(child.error, undefined, `memory child launch failed: ${child.error}`);
    assert.equal(child.status, 0, `memory child failed signal=${child.signal}\n${child.stderr}\n${child.stdout}`);
    const count = /^HEAP-CASES (\d+)$/m.exec(child.stdout);
    assert.ok(count, 'memory child missing asserted case count');
    assert.equal(nativeNumber(count[1]), 24, 'memory child shape/repetition count');
    cases += nativeNumber(count[1]);
    console.log(child.stdout.trim());
    console.log(`CONVERSION javascript method=transparent_native_BigInt_Number_Proxy BigInt_calls=${stats.bigintCalls} Number_string_calls=${stats.numberCalls} max_coefficient_digits=${stats.coefficient} max_payload_digits=${stats.payload} max_exponent_conversion_digits=${stats.exponent} budgets=34/33/16 max_BigInt_calls_per_parse=1 exponent=checked_number_accumulation`);
    console.log(`PARSER-RESOURCE language=javascript cases=${cases}`);
  }
} finally {
  globalThis.BigInt = nativeBigInt;
  globalThis.Number = nativeNumber;
}
