import { readFileSync } from "node:fs";
import {
  Kind,
  decode32,
  decodeBytes32,
  encode32,
  encodeBytes32,
  decode64,
  decodeBytes64,
  encode64,
  encodeBytes64,
  decode128,
  decodeBytes128,
  encode128,
  encodeBytes128,
  fromString,
  toString,
} from "bid754-codec";

const vectorsPath = process.argv[2] ?? "../bid754-codec-vectors/vectors.json";
const vectorFile = JSON.parse(readFileSync(vectorsPath, "utf8"));
const expectedFormatVersion = {{BID_CODEC_VECTOR_FORMAT_VERSION}};
if (vectorFile.format_version !== expectedFormatVersion) {
  throw new Error(`unsupported BID codec vectors format_version ${vectorFile.format_version}, want ${expectedFormatVersion}`);
}
const vectors = vectorFile.vectors;
{{BID_CODEC_JS_ANCHORS}}
const expected = {
  total: {{BID_CODEC_VECTOR_TOTAL}},
  bid32: {{BID_CODEC_BID32_TOTAL}},
  bid64: {{BID_CODEC_BID64_TOTAL}},
  bid128: {{BID_CODEC_BID128_TOTAL}},
  bid32Canonical: {{BID_CODEC_BID32_CANONICAL}},
  bid64Canonical: {{BID_CODEC_BID64_CANONICAL}},
  bid128Canonical: {{BID_CODEC_BID128_CANONICAL}},
};

function assert(condition, message) {
  if (!condition) {
    throw new Error(message);
  }
}

function kindFromString(s) {
  switch (s) {
    case "zero":
      return Kind.Zero;
    case "normal":
      return Kind.Normal;
    case "inf":
      return Kind.Infinity;
    case "qnan":
      return Kind.QNaN;
    case "snan":
      return Kind.SNaN;
    default:
      throw new Error(`unknown kind: ${s}`);
  }
}

function leBytes32(hex) {
  const value = Number(BigInt("0x" + hex));
  return new Uint8Array([
    value & 0xff,
    (value >>> 8) & 0xff,
    (value >>> 16) & 0xff,
    (value >>> 24) & 0xff,
  ]);
}

function leBytes64(hex) {
  const value = BigInt("0x" + hex);
  const bytes = new Uint8Array(8);
  new DataView(bytes.buffer).setBigUint64(0, value, true);
  return bytes;
}

function leBytes128(loHex, hiHex) {
  const bytes = new Uint8Array(16);
  bytes.set(leBytes64(loHex), 0);
  bytes.set(leBytes64(hiHex), 8);
  return bytes;
}

function bytesEqual(a, b) {
  return a.length === b.length && a.every((v, i) => v === b[i]);
}

function componentsEqual(a, b) {
  return a.sign === b.sign
    && a.kind === b.kind
    && a.exponent === b.exponent
    && a.coefficient === b.coefficient
    && (a.payload ?? 0n) === (b.payload ?? 0n);
}

function verifyErrorSemantics() {
  for (const [label, fn] of [
    ["decodeBytes32 short", () => decodeBytes32(new Uint8Array(3))],
    ["decodeBytes32 long", () => decodeBytes32(new Uint8Array(5))],
    ["decodeBytes64 short", () => decodeBytes64(new Uint8Array(7))],
    ["decodeBytes64 long", () => decodeBytes64(new Uint8Array(9))],
    ["decodeBytes128 short", () => decodeBytes128(new Uint8Array(15))],
    ["decodeBytes128 long", () => decodeBytes128(new Uint8Array(17))],
  ]) {
    assertThrows(fn, label);
  }
  // Malformed from_string inputs and out-of-range encode Components are the
  // generated reject_vectors domain (verifyRejectVectors), not a hardcoded list.
}

const expectedRejectTotal = {{BID_CODEC_REJECT_TOTAL}};
const expectedRejectConsumed = {{BID_CODEC_JS_REJECT_CONSUMED}};
const expectedRejectSkipped = {{BID_CODEC_JS_REJECT_SKIPPED}};
const rejectCapabilities = new Set([{{BID_CODEC_JS_REJECT_CAPS}}]);
const rejectUnsupported = new Set([{{BID_CODEC_JS_REJECT_UNSUPPORTED}}]);

function rejectComponents(r) {
  return {
    sign: r.sign ?? false,
    kind: kindFromString(r.kind),
    coefficient: r.coefficient ? BigInt(r.coefficient) : 0n,
    exponent: r.exponent ?? 0,
    payload: r.payload ? BigInt(r.payload) : 0n,
  };
}

function verifyRejectVectors() {
  const rejects = vectorFile.reject_vectors;
  assert(rejects.length === expectedRejectTotal, `reject_vectors total changed: ${rejects.length}`);
  let consumed = 0;
  let skipped = 0;
  const skipReasons = {};
  for (const r of rejects) {
    if (r.requires && !rejectCapabilities.has(r.requires)) {
      if (!rejectUnsupported.has(r.requires)) {
        throw new Error(`reject record requires tag outside the declared capability universe: ${r.requires}`);
      }
      skipped += 1;
      skipReasons[r.requires] = (skipReasons[r.requires] ?? 0) + 1;
      continue;
    }
    consumed += 1;
    // Record-field access, kind parsing, and Components construction happen
    // outside assertThrows (their failures crash the runner as harness
    // failures); only the public API call sits inside the error-expectation
    // callback, so an altered record cannot pass as a rejection.
    if (r.channel === "from_string") {
      const input = r.input;
      assertThrows(() => fromString(input), `reject from_string ${JSON.stringify(input)} (${r.reason})`);
    } else if (r.channel === "encode") {
      const c = rejectComponents(r);
      const label = `reject encode ${r.type} ${r.kind} (${r.reason})`;
      switch (r.type) {
        case "bid32":
          assertThrows(() => encode32(c), label);
          break;
        case "bid64":
          assertThrows(() => encode64(c), label);
          break;
        case "bid128":
          assertThrows(() => encode128(c), label);
          break;
        default:
          throw new Error(`unknown reject encode type: ${r.type}`);
      }
    } else if (r.channel === "to_string") {
      const c = rejectComponents(r);
      assertThrows(() => toString(c), `reject to_string ${r.kind} (${r.reason})`);
    } else {
      throw new Error(`unknown reject channel: ${r.channel}`);
    }
  }
  assert(consumed === expectedRejectConsumed, `reject consumed count changed: ${consumed}`);
  assert(skipped === expectedRejectSkipped, `reject skipped count changed: ${skipped}`);
  assert(consumed + skipped === rejects.length, "reject consumption does not partition the reject set");
  console.log(`reject_vectors: consumed=${consumed} skipped=${skipped} skipReasons=${JSON.stringify(skipReasons)}`);
}

const expectedStringTotal = {{BID_CODEC_STRING_TOTAL}};

function verifyStringVectors() {
  // string_vectors: the generated SUCCESS channel for the string surface. Each
  // record's input must parse and re-render as the exact expected string,
  // pinning fromString→toString agreement across all language consumers in the
  // encoding-unreachable Components region (above all int32-extreme exponents
  // whose adjusted exponent exceeds int32) plus the successful grammar-edge
  // normalizations. The closure leg then re-parses the expected rendering
  // itself: fromString(expected) must succeed and toString must reproduce it
  // exactly (parse(render(x)) is total and expected is a rendering fixed
  // point), so a parser that rejects its own renderer's output fails here.
  // The channel is capability-ungated: every record is consumed.
  const stringVectors = vectorFile.string_vectors;
  assert(
    Array.isArray(stringVectors) && stringVectors.length === expectedStringTotal,
    `string_vectors total changed: ${stringVectors?.length}`,
  );
  let consumed = 0;
  for (const sv of stringVectors) {
    consumed += 1;
    let c;
    try {
      c = fromString(sv.input);
    } catch (err) {
      throw new Error(`string_vectors fromString ${JSON.stringify(sv.input)} threw, want success: ${err}`);
    }
    const got = toString(c);
    assert(got === sv.expected, `string_vectors ${JSON.stringify(sv.input)}: toString got ${got}, want ${sv.expected}`);
    let reparsed;
    try {
      reparsed = fromString(sv.expected);
    } catch (err) {
      throw new Error(`string_vectors closure fromString ${JSON.stringify(sv.expected)}: rendering not reparseable: ${err}`);
    }
    const again = toString(reparsed);
    assert(again === sv.expected, `string_vectors closure ${JSON.stringify(sv.expected)}: re-rendered as ${again}, not a fixed point`);
  }
  assert(consumed === expectedStringTotal, `string_vectors consumed count changed: ${consumed}`);
  console.log(`string_vectors: consumed=${consumed}`);
}

function verifyTypeDomainRejects() {
  // Reject values the shared JSON schema cannot express, constructible only in
  // dynamically-typed languages (a non-boolean sign, wrong numeric field type,
  // out-of-int32/non-finite exponent, or kind outside the defined set).
  const cases = [{{BID_CODEC_JS_REJECT_TYPE_DOMAIN}}];
  for (const [id, comp] of cases) {
    assertThrows(() => encode32(comp), `type-domain reject ${id}`);
    assertThrows(() => toString(comp), `to_string type-domain reject ${id}`);
  }
}

function verifyRawDecodeRejects() {
  const cases = [{{BID_CODEC_JS_RAW_DECODE_REJECTS}}];
  for (const [id, invoke] of cases) {
    assertThrows(invoke, `raw decode reject ${id}`);
  }
}

function assertThrows(fn, label) {
  try {
    fn();
  } catch {
    return;
  }
  throw new Error(`${label} succeeded, want throw`);
}

function verifyAnchorVectors() {
  assert(anchorVectors.length === {{BID_CODEC_VECTOR_ANCHOR_COUNT}}, `BID codec anchor count changed: ${anchorVectors.length}`);
  for (const v of anchorVectors) {
    let c;
    if (v.type === "bid32") {
      c = decode32(Number(BigInt("0x" + v.hex)));
      assert(encode32(c) === (Number(BigInt("0x" + v.encoded_hex)) >>> 0), `${v.hex} anchor encode32 mismatch`);
    } else if (v.type === "bid64") {
      c = decode64(BigInt("0x" + v.hex));
      assert(encode64(c) === BigInt("0x" + v.encoded_hex), `${v.hex} anchor encode64 mismatch`);
    } else if (v.type === "bid128") {
      c = decode128(BigInt("0x" + v.hex), BigInt("0x" + v.hex_hi));
      const [lo, hi] = encode128(c);
      assert(lo === BigInt("0x" + v.encoded_hex) && hi === BigInt("0x" + v.encoded_hi), `${v.hex_hi}_${v.hex} anchor encode128 mismatch`);
    } else {
      throw new Error(`unknown anchor vector type: ${v.type}`);
    }
    assert(v.canonical === true, `${v.type} ${v.hex} anchor canonical mismatch`);
    assert(c.sign === v.sign, `${v.type} ${v.hex} anchor sign mismatch`);
    assert(c.kind === kindFromString(v.kind), `${v.type} ${v.hex} anchor kind mismatch`);
    assert(c.exponent === v.exponent, `${v.type} ${v.hex} anchor exponent mismatch`);
    if (v.kind !== "qnan" && v.kind !== "snan") {
      const wantCoeff = v.coefficient === "" ? 0n : BigInt(v.coefficient);
      assert(c.coefficient === wantCoeff, `${v.type} ${v.hex} anchor coefficient mismatch`);
    }
    assert((c.payload ?? 0n) === (v.payload === undefined ? 0n : BigInt(v.payload)), `${v.type} ${v.hex} anchor payload mismatch`);
    assert(toString(c) === v.decimal_string, `${v.type} ${v.hex} anchor toString mismatch`);
  }
}

let bid32 = 0;
let bid64 = 0;
let bid128 = 0;
let bid32Canonical = 0;
let bid64Canonical = 0;
let bid128Canonical = 0;
let decode = 0;
let encode = 0;

verifyAnchorVectors();

for (const v of vectors) {
  const expectedKind = kindFromString(v.kind);
  let c;
  if (v.type === "bid32") {
    bid32 += 1;
    if (v.canonical) bid32Canonical += 1;
    const bits = Number(BigInt("0x" + v.hex));
    c = decode32(bits);
    assert(componentsEqual(decodeBytes32(leBytes32(v.hex)), c), `${v.hex} decodeBytes32 mismatch`);
  } else if (v.type === "bid64") {
    bid64 += 1;
    if (v.canonical) bid64Canonical += 1;
    c = decode64(BigInt("0x" + v.hex));
    assert(componentsEqual(decodeBytes64(leBytes64(v.hex)), c), `${v.hex} decodeBytes64 mismatch`);
  } else if (v.type === "bid128") {
    bid128 += 1;
    if (v.canonical) bid128Canonical += 1;
    c = decode128(BigInt("0x" + v.hex), BigInt("0x" + v.hex_hi));
    assert(componentsEqual(decodeBytes128(leBytes128(v.hex, v.hex_hi)), c), `${v.hex_hi}_${v.hex} decodeBytes128 mismatch`);
  } else {
    throw new Error(`unknown vector type: ${v.type}`);
  }

  assert(c.sign === v.sign, `${v.type} ${v.hex} sign mismatch`);
  assert(c.kind === expectedKind, `${v.type} ${v.hex} kind mismatch`);
  assert(c.exponent === v.exponent, `${v.type} ${v.hex} exponent mismatch`);
  const wantCoeff = v.coefficient === "" ? 0n : BigInt(v.coefficient);
  assert(c.coefficient === wantCoeff, `${v.type} ${v.hex} coefficient mismatch`);
  if (v.payload !== undefined) {
    assert(c.payload === BigInt(v.payload), `${v.type} ${v.hex} payload mismatch`);
  }
  assert(toString(c) === v.decimal_string, `${v.type} ${v.hex} toString mismatch`);

  const parsed = fromString(v.decimal_string);
  if (v.type === "bid32") {
    assert(encode32(parsed) === (Number(BigInt("0x" + v.encoded_hex)) >>> 0), `${v.hex} fromString encode32 mismatch`);
    if (v.canonical) {
      assert(encode32(c) === (Number(BigInt("0x" + v.encoded_hex)) >>> 0), `${v.hex} encode32 mismatch`);
      assert(bytesEqual(encodeBytes32(c), leBytes32(v.encoded_hex)), `${v.hex} encodeBytes32 mismatch`);
      encode += 1;
    }
  } else if (v.type === "bid64") {
    assert(encode64(parsed) === BigInt("0x" + v.encoded_hex), `${v.hex} fromString encode64 mismatch`);
    if (v.canonical) {
      assert(encode64(c) === BigInt("0x" + v.encoded_hex), `${v.hex} encode64 mismatch`);
      assert(bytesEqual(encodeBytes64(c), leBytes64(v.encoded_hex)), `${v.hex} encodeBytes64 mismatch`);
      encode += 1;
    }
  } else {
    const [parsedLo, parsedHi] = encode128(parsed);
    assert(parsedLo === BigInt("0x" + v.encoded_hex) && parsedHi === BigInt("0x" + v.encoded_hi), `${v.hex} fromString encode128 mismatch`);
    if (v.canonical) {
      const [lo, hi] = encode128(c);
      assert(lo === BigInt("0x" + v.encoded_hex) && hi === BigInt("0x" + v.encoded_hi), `${v.hex} encode128 mismatch`);
      assert(bytesEqual(encodeBytes128(c), leBytes128(v.encoded_hex, v.encoded_hi)), `${v.hex} encodeBytes128 mismatch`);
      encode += 1;
    }
  }
  decode += 1;
}

assert(vectors.length === expected.total, `total vectors changed: ${vectors.length}`);
assert(bid32 === expected.bid32 && bid64 === expected.bid64 && bid128 === expected.bid128, "per-format vector count changed");
assert(bid32Canonical === expected.bid32Canonical && bid64Canonical === expected.bid64Canonical && bid128Canonical === expected.bid128Canonical, "canonical vector count changed");
verifyErrorSemantics();
verifyRejectVectors();
verifyStringVectors();
verifyTypeDomainRejects();
verifyRawDecodeRejects();
console.log(`BID codec JS package vectors: decode=${decode} encode=${encode}`);
