import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";
import {
  Kind,
  type Components,
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
} from "./index.js";

interface Vector {
  type: "bid32" | "bid64" | "bid128";
  hex: string;
  hex_hi?: string;
  sign: boolean;
  coefficient: string; // "" for zero/inf/nan-without-payload
  exponent: number;
  kind: string; // "zero" | "normal" | "inf" | "qnan" | "snan"
  payload?: string;
  decimal_string: string;
  canonical: boolean;
  encoded_hex: string;
  encoded_hi?: string;
}

interface RejectVector {
  channel: "from_string" | "encode";
  type?: "bid32" | "bid64" | "bid128";
  input?: string;
  sign?: boolean;
  kind?: string;
  coefficient?: string;
  exponent?: number;
  payload?: string;
  reason: string;
  requires?: string;
}

interface StringVector {
  input: string;
  expected: string;
}

interface VectorFile {
  format_version: number;
  vectors: Vector[];
  reject_vectors: RejectVector[];
  string_vectors: StringVector[];
}

const vectorsPath = resolve(__dirname, "../../bid754-codec-vectors/vectors.json");
const vectorFile: VectorFile = JSON.parse(readFileSync(vectorsPath, "utf-8"));
const expectedFormatVersion = {{BID_CODEC_VECTOR_FORMAT_VERSION}};
if (vectorFile.format_version !== expectedFormatVersion) {
  throw new Error(`unsupported BID codec vectors format_version ${vectorFile.format_version}, want ${expectedFormatVersion}`);
}
const vectors: Vector[] = vectorFile.vectors;
const anchorVectors: Vector[] = {{BID_CODEC_JS_ANCHOR_ARRAY}};

function kindFromString(s: string): Kind {
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

const bid32Vectors = vectors.filter((v) => v.type === "bid32");
const bid64Vectors = vectors.filter((v) => v.type === "bid64");
const bid128Vectors = vectors.filter((v) => v.type === "bid128");
const bid32Canonical = bid32Vectors.filter((v) => v.canonical);
const bid64Canonical = bid64Vectors.filter((v) => v.canonical);
const bid128Canonical = bid128Vectors.filter((v) => v.canonical);

const expectedTotal = {{BID_CODEC_VECTOR_TOTAL}};
const expectedBid32 = {{BID_CODEC_BID32_TOTAL}};
const expectedBid64 = {{BID_CODEC_BID64_TOTAL}};
const expectedBid128 = {{BID_CODEC_BID128_TOTAL}};
const expectedBid32Canonical = {{BID_CODEC_BID32_CANONICAL}};
const expectedBid64Canonical = {{BID_CODEC_BID64_CANONICAL}};
const expectedBid128Canonical = {{BID_CODEC_BID128_CANONICAL}};

describe("vectors: coverage profile", () => {
  it("matches generated vector counts", () => {
    expect(vectors).toHaveLength(expectedTotal);
    expect(bid32Vectors).toHaveLength(expectedBid32);
    expect(bid64Vectors).toHaveLength(expectedBid64);
    expect(bid128Vectors).toHaveLength(expectedBid128);
    expect(bid32Canonical).toHaveLength(expectedBid32Canonical);
    expect(bid64Canonical).toHaveLength(expectedBid64Canonical);
    expect(bid128Canonical).toHaveLength(expectedBid128Canonical);
  });
});

describe("vectors: anchor contract", () => {
  it.each(anchorVectors)("matches hardcoded anchor $type $hex", (vec) => {
    let c;
    if (vec.type === "bid32") {
      c = decode32(Number(BigInt("0x" + vec.hex)));
      expect(encode32(c)).toBe(Number(BigInt("0x" + vec.encoded_hex)) >>> 0);
    } else if (vec.type === "bid64") {
      c = decode64(BigInt("0x" + vec.hex));
      expect(encode64(c)).toBe(BigInt("0x" + vec.encoded_hex));
    } else {
      c = decode128(BigInt("0x" + vec.hex), BigInt("0x" + vec.hex_hi!));
      const [lo, hi] = encode128(c);
      expect(lo).toBe(BigInt("0x" + vec.encoded_hex));
      expect(hi).toBe(BigInt("0x" + vec.encoded_hi!));
    }
    expect(vec.canonical).toBe(true);
    expect(c.sign).toBe(vec.sign);
    expect(c.kind).toBe(kindFromString(vec.kind));
    expect(c.exponent).toBe(vec.exponent);
    if (vec.kind !== "qnan" && vec.kind !== "snan") {
      expect(c.coefficient).toBe(vec.coefficient === "" ? 0n : BigInt(vec.coefficient));
    }
    expect(c.payload ?? 0n).toBe(vec.payload === undefined ? 0n : BigInt(vec.payload));
    expect(toString(c)).toBe(vec.decimal_string);
  });
});

describe("vectors: error semantics", () => {
  it("rejects invalid byte lengths", () => {
    expect(() => decodeBytes32(new Uint8Array(3))).toThrow();
    expect(() => decodeBytes32(new Uint8Array(5))).toThrow();
    expect(() => decodeBytes64(new Uint8Array(7))).toThrow();
    expect(() => decodeBytes64(new Uint8Array(9))).toThrow();
    expect(() => decodeBytes128(new Uint8Array(15))).toThrow();
    expect(() => decodeBytes128(new Uint8Array(17))).toThrow();
  });

  // Malformed from_string inputs and out-of-range encode Components are the
  // generated reject_vectors domain (below), not a hardcoded list.
});

const expectedRejectTotal = {{BID_CODEC_REJECT_TOTAL}};
const expectedRejectConsumed = {{BID_CODEC_JS_REJECT_CONSUMED}};
const expectedRejectSkipped = {{BID_CODEC_JS_REJECT_SKIPPED}};
const rejectCapabilities = new Set<string>([{{BID_CODEC_JS_REJECT_CAPS}}]);
const rejectUnsupported = new Set<string>([{{BID_CODEC_JS_REJECT_UNSUPPORTED}}]);
const rejectVectors: RejectVector[] = vectorFile.reject_vectors;
const consumedRejects = rejectVectors.filter((r) => !r.requires || rejectCapabilities.has(r.requires));
const skippedRejects = rejectVectors.filter((r) => r.requires && !rejectCapabilities.has(r.requires));

function rejectComponents(r: RejectVector): Components {
  return {
    sign: r.sign ?? false,
    kind: kindFromString(r.kind ?? ""),
    coefficient: r.coefficient ? BigInt(r.coefficient) : 0n,
    exponent: r.exponent ?? 0,
    payload: r.payload ? BigInt(r.payload) : 0n,
  };
}

describe("vectors: reject domain", () => {
  it("has the pinned reject count and consumption split", () => {
    expect(rejectVectors).toHaveLength(expectedRejectTotal);
    expect(consumedRejects).toHaveLength(expectedRejectConsumed);
    expect(skippedRejects).toHaveLength(expectedRejectSkipped);
    expect(consumedRejects.length + skippedRejects.length).toBe(rejectVectors.length);
  });

  it("skips only records whose requires tag is a declared-unsupported capability", () => {
    for (const r of skippedRejects) {
      expect(r.requires !== undefined && rejectUnsupported.has(r.requires)).toBe(true);
    }
  });

  it.each(consumedRejects)(
    "rejects $channel $reason",
    (r: RejectVector) => {
      // Record-field access, kind parsing, and Components construction happen
      // outside expect().toThrow() (their failures fail the test as harness
      // failures); only the public API call sits inside the error-expectation
      // callback, so an altered record cannot pass as a rejection.
      if (r.channel === "from_string") {
        const input = r.input ?? "";
        expect(() => fromString(input)).toThrow();
      } else if (r.channel === "encode") {
        const c = rejectComponents(r);
        switch (r.type) {
          case "bid32":
            expect(() => encode32(c)).toThrow();
            break;
          case "bid64":
            expect(() => encode64(c)).toThrow();
            break;
          case "bid128":
            expect(() => encode128(c)).toThrow();
            break;
          default:
            throw new Error(`unknown reject encode type: ${String(r.type)}`);
        }
      } else if (r.channel === "to_string") {
        const c = rejectComponents(r);
        expect(() => toString(c)).toThrow();
      } else {
        throw new Error(`unknown reject channel: ${String(r.channel)}`);
      }
    },
  );
});

const expectedStringTotal = {{BID_CODEC_STRING_TOTAL}};
const stringVectors: StringVector[] = vectorFile.string_vectors;

describe("vectors: string success channel", () => {
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
  it("has the pinned string vector count", () => {
    expect(stringVectors).toHaveLength(expectedStringTotal);
  });

  it.each(stringVectors)("round-trips $input", (sv: StringVector) => {
    const c = fromString(sv.input); // must succeed; a throw fails the harness
    expect(toString(c)).toBe(sv.expected);
    const reparsed = fromString(sv.expected); // closure: rendering must reparse
    expect(toString(reparsed)).toBe(sv.expected); // fixed point
  });
});

const typeDomainRejects: [string, unknown][] = [{{BID_CODEC_JS_REJECT_TYPE_DOMAIN}}];

describe("vectors: reject type domain", () => {
  // Reject values the shared JSON schema cannot express, constructible only in
  // dynamically-typed languages (a non-boolean sign, wrong numeric field type,
  // out-of-int32/non-finite exponent, or kind outside the defined set).
  it.each(typeDomainRejects)("rejects %s", (_id, comp) => {
    expect(() => encode32(comp as Components)).toThrow();
    expect(() => toString(comp as Components)).toThrow();
  });
});

const rawDecodeRejects: [string, () => unknown][] = [{{BID_CODEC_JS_RAW_DECODE_REJECTS}}];

describe("vectors: reject raw decode domain", () => {
  // TypeScript declarations do not constrain runtime JavaScript callers. Require an
  // exact uint32 number or uint64 bigint before any narrowing/coercing bit op.
  it.each(rawDecodeRejects)("rejects %s", (_id, invoke) => {
    expect(invoke).toThrow();
  });
});

function leBytes32(hex: string): Uint8Array {
  const value = Number(BigInt("0x" + hex));
  return new Uint8Array([
    value & 0xff,
    (value >>> 8) & 0xff,
    (value >>> 16) & 0xff,
    (value >>> 24) & 0xff,
  ]);
}

function leBytes64(hex: string): Uint8Array {
  const value = BigInt("0x" + hex);
  const bytes = new Uint8Array(8);
  const dv = new DataView(bytes.buffer);
  dv.setBigUint64(0, value, true);
  return bytes;
}

function leBytes128(loHex: string, hiHex: string): Uint8Array {
  const bytes = new Uint8Array(16);
  bytes.set(leBytes64(loHex), 0);
  bytes.set(leBytes64(hiHex), 8);
  return bytes;
}

// --- BID32 ---

describe("vectors: bid32 decode", () => {
  it.each(bid32Vectors)("decode $hex", (vec) => {
    const bits = Number(BigInt("0x" + vec.hex));
    const c = decode32(bits);

    expect(c.sign).toBe(vec.sign);
    expect(c.kind).toBe(kindFromString(vec.kind));
    expect(c.exponent).toBe(vec.exponent);

    const expectedCoeff = vec.coefficient === "" ? 0n : BigInt(vec.coefficient);
    expect(c.coefficient).toBe(expectedCoeff);

    if (vec.payload !== undefined) {
      expect(c.payload).toBe(BigInt(vec.payload));
    }
    expect(decodeBytes32(leBytes32(vec.hex))).toEqual(c);
    expect(toString(c)).toBe(vec.decimal_string);
    expect(encode32(fromString(vec.decimal_string))).toBe(Number(BigInt("0x" + vec.encoded_hex)) >>> 0);
  });
});

describe("vectors: bid32 roundtrip", () => {
  it.each(bid32Canonical)("roundtrip $hex", (vec) => {
    const bits = Number(BigInt("0x" + vec.hex));
    const c = decode32(bits);
    const encoded = encode32(c);
    const expectedBits = Number(BigInt("0x" + vec.encoded_hex));
    expect(encoded).toBe(expectedBits >>> 0);
    expect(encodeBytes32(c)).toEqual(leBytes32(vec.encoded_hex));
  });
});

// --- BID64 ---

describe("vectors: bid64 decode", () => {
  it.each(bid64Vectors)("decode $hex", (vec) => {
    const bits = BigInt("0x" + vec.hex);
    const c = decode64(bits);

    expect(c.sign).toBe(vec.sign);
    expect(c.kind).toBe(kindFromString(vec.kind));
    expect(c.exponent).toBe(vec.exponent);

    const expectedCoeff = vec.coefficient === "" ? 0n : BigInt(vec.coefficient);
    expect(c.coefficient).toBe(expectedCoeff);

    if (vec.payload !== undefined) {
      expect(c.payload).toBe(BigInt(vec.payload));
    }
    expect(decodeBytes64(leBytes64(vec.hex))).toEqual(c);
    expect(toString(c)).toBe(vec.decimal_string);
    expect(encode64(fromString(vec.decimal_string))).toBe(BigInt("0x" + vec.encoded_hex));
  });
});

describe("vectors: bid64 roundtrip", () => {
  it.each(bid64Canonical)("roundtrip $hex", (vec) => {
    const bits = BigInt("0x" + vec.hex);
    const c = decode64(bits);
    const encoded = encode64(c);
    const expectedBits = BigInt("0x" + vec.encoded_hex);
    expect(encoded).toBe(expectedBits);
    expect(encodeBytes64(c)).toEqual(leBytes64(vec.encoded_hex));
  });
});

// --- BID128 ---

describe("vectors: bid128 decode", () => {
  it.each(bid128Vectors)("decode $hex/$hex_hi", (vec) => {
    const lo = BigInt("0x" + vec.hex);
    const hi = BigInt("0x" + vec.hex_hi!);
    const c = decode128(lo, hi);

    expect(c.sign).toBe(vec.sign);
    expect(c.kind).toBe(kindFromString(vec.kind));
    expect(c.exponent).toBe(vec.exponent);

    const expectedCoeff = vec.coefficient === "" ? 0n : BigInt(vec.coefficient);
    expect(c.coefficient).toBe(expectedCoeff);

    if (vec.payload !== undefined) {
      expect(c.payload).toBe(BigInt(vec.payload));
    }
    expect(decodeBytes128(leBytes128(vec.hex, vec.hex_hi!))).toEqual(c);
    expect(toString(c)).toBe(vec.decimal_string);
    const [parsedLo, parsedHi] = encode128(fromString(vec.decimal_string));
    expect(parsedLo).toBe(BigInt("0x" + vec.encoded_hex));
    expect(parsedHi).toBe(BigInt("0x" + vec.encoded_hi!));
  });
});

describe("vectors: bid128 roundtrip", () => {
  it.each(bid128Canonical)("roundtrip $hex/$hex_hi", (vec) => {
    const lo = BigInt("0x" + vec.hex);
    const hi = BigInt("0x" + vec.hex_hi!);
    const c = decode128(lo, hi);
    const [encodedLo, encodedHi] = encode128(c);
    const expectedLo = BigInt("0x" + vec.encoded_hex);
    const expectedHi = BigInt("0x" + vec.encoded_hi!);
    expect(encodedLo).toBe(expectedLo);
    expect(encodedHi).toBe(expectedHi);
    expect(encodeBytes128(c)).toEqual(leBytes128(vec.encoded_hex, vec.encoded_hi!));
  });
});
