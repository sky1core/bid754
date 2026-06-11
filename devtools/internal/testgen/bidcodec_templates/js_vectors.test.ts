import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";
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

interface VectorFile {
  format_version: number;
  vectors: Vector[];
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

const expectedTotal = 15046;
const expectedBid32 = 5019;
const expectedBid64 = 5017;
const expectedBid128 = 5010;
const expectedBid32Canonical = 4464;
const expectedBid64Canonical = 4210;
const expectedBid128Canonical = 3641;

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

  it("rejects malformed strings", () => {
    for (const input of ["", "NaNabc", "SNaN-1", "1.2.3", "1E", "1Eabc", "1E2147483648", "1.0E2147483648"]) {
      expect(() => fromString(input)).toThrow();
    }
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
