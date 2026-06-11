import { describe, expect, it } from "vitest";
import {
  Kind,
  Components,
  decode32,
  encode32,
  decode64,
  encode64,
  decode128,
  encode128,
  decodeBytes32,
  encodeBytes32,
  decodeBytes64,
  encodeBytes64,
  decodeBytes128,
  encodeBytes128,
  toString,
  fromString,
} from "./index.js";

function comp(
  kind: Kind,
  sign = false,
  coefficient = 0n,
  exponent = 0,
  payload = 0n,
): Components {
  return { sign, coefficient, exponent, kind, payload };
}

// --- BID32 ---

describe("decode32", () => {
  it("decodes zero", () => {
    const c = decode32(0x32800000);
    expect(c.kind).toBe(Kind.Zero);
    expect(c.sign).toBe(false);
    expect(c.exponent).toBe(0);
  });

  it("decodes negative zero", () => {
    const c = decode32(0xb2800000);
    expect(c.kind).toBe(Kind.Zero);
    expect(c.sign).toBe(true);
    expect(c.exponent).toBe(0);
  });

  it("decodes one", () => {
    const c = decode32(0x32800001);
    expect(c.kind).toBe(Kind.Normal);
    expect(c.coefficient).toBe(1n);
    expect(c.exponent).toBe(0);
  });

  it("decodes negative one", () => {
    const c = decode32(0xb2800001);
    expect(c.kind).toBe(Kind.Normal);
    expect(c.sign).toBe(true);
    expect(c.coefficient).toBe(1n);
    expect(c.exponent).toBe(0);
  });

  it("decodes +inf", () => {
    const c = decode32(0x78000000);
    expect(c.kind).toBe(Kind.Infinity);
    expect(c.sign).toBe(false);
  });

  it("decodes -inf", () => {
    const c = decode32(0xf8000000);
    expect(c.kind).toBe(Kind.Infinity);
    expect(c.sign).toBe(true);
  });

  it("decodes QNaN", () => {
    const c = decode32(0x7c000000);
    expect(c.kind).toBe(Kind.QNaN);
  });

  it("decodes SNaN", () => {
    const c = decode32(0x7e000000);
    expect(c.kind).toBe(Kind.SNaN);
  });

  it("decodes max value (special encoding)", () => {
    const c = decode32(0x77f8967f);
    expect(c.kind).toBe(Kind.Normal);
    expect(c.coefficient).toBe(9999999n);
    expect(c.exponent).toBe(90);
  });
});

describe("encode32 roundtrip", () => {
  const values = [
    0x32800000, // +0
    0xb2800000, // -0
    0x32800001, // +1
    0x32800064, // +100
    0x77f8967f, // 9999999 * 10^90 (special encoding)
    0x78000000, // +inf
    0xf8000000, // -inf
    0x7c000000, // NaN
    0x7e000000, // sNaN
  ];

  for (const v of values) {
    it(`roundtrip 0x${v.toString(16).padStart(8, "0")}`, () => {
      const c = decode32(v);
      const got = encode32(c);
      expect(got).toBe(v >>> 0);
    });
  }
});

// --- BID64 ---

describe("decode64", () => {
  it("decodes zero", () => {
    const c = decode64(0x31c0000000000000n);
    expect(c.kind).toBe(Kind.Zero);
    expect(c.exponent).toBe(0);
  });

  it("decodes one", () => {
    const c = decode64(0x31c0000000000001n);
    expect(c.kind).toBe(Kind.Normal);
    expect(c.coefficient).toBe(1n);
    expect(c.exponent).toBe(0);
  });

  it("decodes negative zero", () => {
    const c = decode64(0xb1c0000000000000n);
    expect(c.kind).toBe(Kind.Zero);
    expect(c.sign).toBe(true);
  });

  it("decodes +inf", () => {
    const c = decode64(0x7800000000000000n);
    expect(c.kind).toBe(Kind.Infinity);
  });

  it("decodes QNaN", () => {
    const c = decode64(0x7c00000000000000n);
    expect(c.kind).toBe(Kind.QNaN);
  });

  it("decodes SNaN", () => {
    const c = decode64(0x7e00000000000000n);
    expect(c.kind).toBe(Kind.SNaN);
  });
});

describe("encode64 roundtrip", () => {
  const values = [
    0x31c0000000000000n, // +0
    0xb1c0000000000000n, // -0
    0x31c0000000000001n, // +1
    0x7800000000000000n, // +inf
    0x7c00000000000000n, // NaN
    0x7e00000000000000n, // sNaN
  ];

  for (const v of values) {
    it(`roundtrip 0x${v.toString(16).padStart(16, "0")}`, () => {
      const c = decode64(v);
      const got = encode64(c);
      expect(got).toBe(v);
    });
  }
});

// --- BID128 ---

describe("decode128", () => {
  it("decodes +1", () => {
    const lo = 0x0000000000000001n;
    const hi = BigInt(6176) << 49n;
    const c = decode128(lo, hi);
    expect(c.kind).toBe(Kind.Normal);
    expect(c.exponent).toBe(0);
    expect(c.coefficient).toBe(1n);
    expect(c.sign).toBe(false);
  });

  it("decodes +inf", () => {
    const c = decode128(0n, 0x7800000000000000n);
    expect(c.kind).toBe(Kind.Infinity);
    expect(c.sign).toBe(false);
  });

  it("decodes QNaN", () => {
    const c = decode128(0n, 0x7c00000000000000n);
    expect(c.kind).toBe(Kind.QNaN);
  });

  it("decodes zero", () => {
    const hi = BigInt(6176) << 49n;
    const c = decode128(0n, hi);
    expect(c.kind).toBe(Kind.Zero);
    expect(c.exponent).toBe(0);
  });
});

describe("encode128 roundtrip", () => {
  const signMask = 0x8000000000000000n;
  const cases: [bigint, bigint][] = [
    [0n, BigInt(6176) << 49n], // +0
    [0n, signMask | (BigInt(6176) << 49n)], // -0
    [1n, BigInt(6176) << 49n], // +1
    [0n, 0x7800000000000000n], // +inf
    [0n, 0x7c00000000000000n], // NaN
  ];

  for (const [lo, hi] of cases) {
    it(`roundtrip hi=0x${hi.toString(16)} lo=0x${lo.toString(16)}`, () => {
      const c = decode128(lo, hi);
      const [gotLo, gotHi] = encode128(c);
      expect(gotLo).toBe(lo);
      expect(gotHi).toBe(hi);
    });
  }
});

describe("encode128 NaN payload", () => {
  it("uses the payload field instead of coefficient bits", () => {
    const nan = comp(Kind.QNaN, false, (1n << 80n) | 12345n, 0, 999n);
    const [lo, hi] = encode128(nan);
    expect(lo).toBe(999n);
    expect(hi).toBe(0x7c00000000000000n);
  });
});

// --- Cross-format consistency ---

describe("cross-format", () => {
  it("+1 encodes consistently across formats", () => {
    const one = comp(Kind.Normal, false, 1n, 0);

    const v32 = encode32(one);
    const c32 = decode32(v32);
    expect(c32.coefficient).toBe(1n);
    expect(c32.exponent).toBe(0);

    const v64 = encode64(one);
    const c64 = decode64(v64);
    expect(c64.coefficient).toBe(1n);
    expect(c64.exponent).toBe(0);

    const [lo128, hi128] = encode128(one);
    const c128 = decode128(lo128, hi128);
    expect(c128.coefficient).toBe(1n);
    expect(c128.exponent).toBe(0);
  });

  it("NaN payload roundtrips for BID32", () => {
    const nan = comp(Kind.QNaN, false, 0n, 0, 12345n);
    const v = encode32(nan);
    const c = decode32(v);
    expect(c.kind).toBe(Kind.QNaN);
    expect(c.payload).toBe(12345n);
  });

  it("NaN payload roundtrips for BID64", () => {
    const nan = comp(Kind.QNaN, false, 0n, 0, 12345n);
    const v = encode64(nan);
    const c = decode64(v);
    expect(c.kind).toBe(Kind.QNaN);
    expect(c.payload).toBe(12345n);
  });
});

// --- Known BID encodings (from Go test) ---

describe("known encodings", () => {
  it("BID32 +123.45 = 12345 * 10^-2", () => {
    const c = comp(Kind.Normal, false, 12345n, -2);
    const v = encode32(c);
    const d = decode32(v);
    expect(d.kind).toBe(Kind.Normal);
    expect(d.coefficient).toBe(12345n);
    expect(d.exponent).toBe(-2);
  });

  it("BID64 +123.45 = 12345 * 10^-2", () => {
    const c = comp(Kind.Normal, false, 12345n, -2);
    const v = encode64(c);
    const d = decode64(v);
    expect(d.kind).toBe(Kind.Normal);
    expect(d.coefficient).toBe(12345n);
    expect(d.exponent).toBe(-2);
  });

  it("BID128 +123.45 = 12345 * 10^-2", () => {
    const c = comp(Kind.Normal, false, 12345n, -2);
    const [lo, hi] = encode128(c);
    const d = decode128(lo, hi);
    expect(d.kind).toBe(Kind.Normal);
    expect(d.coefficient).toBe(12345n);
    expect(d.exponent).toBe(-2);
  });
});

// --- Byte encoding/decoding ---

describe("decodeBytes32 / encodeBytes32", () => {
  it("roundtrips +1", () => {
    const original = comp(Kind.Normal, false, 1n, 0);
    const bytes = encodeBytes32(original);
    expect(bytes.length).toBe(4);
    const decoded = decodeBytes32(bytes);
    expect(decoded.kind).toBe(Kind.Normal);
    expect(decoded.coefficient).toBe(1n);
    expect(decoded.exponent).toBe(0);
  });

  it("matches encode32 little-endian layout", () => {
    const original = comp(Kind.Normal, false, 12345n, -2);
    const v = encode32(original);
    const bytes = encodeBytes32(original);
    expect(bytes[0]).toBe(v & 0xff);
    expect(bytes[1]).toBe((v >>> 8) & 0xff);
    expect(bytes[2]).toBe((v >>> 16) & 0xff);
    expect(bytes[3]).toBe((v >>> 24) & 0xff);
  });

  it("throws on short buffer", () => {
    expect(() => decodeBytes32(new Uint8Array(3))).toThrow();
  });

  it("throws on long buffer", () => {
    expect(() => decodeBytes32(new Uint8Array(5))).toThrow();
  });
});

describe("decodeBytes64 / encodeBytes64", () => {
  it("roundtrips +1", () => {
    const original = comp(Kind.Normal, false, 1n, 0);
    const bytes = encodeBytes64(original);
    expect(bytes.length).toBe(8);
    const decoded = decodeBytes64(bytes);
    expect(decoded.kind).toBe(Kind.Normal);
    expect(decoded.coefficient).toBe(1n);
    expect(decoded.exponent).toBe(0);
  });

  it("roundtrips +inf", () => {
    const original = comp(Kind.Infinity, false);
    const bytes = encodeBytes64(original);
    const decoded = decodeBytes64(bytes);
    expect(decoded.kind).toBe(Kind.Infinity);
    expect(decoded.sign).toBe(false);
  });

  it("throws on short buffer", () => {
    expect(() => decodeBytes64(new Uint8Array(7))).toThrow();
  });

  it("throws on long buffer", () => {
    expect(() => decodeBytes64(new Uint8Array(9))).toThrow();
  });
});

describe("decodeBytes128 / encodeBytes128", () => {
  it("roundtrips +1", () => {
    const original = comp(Kind.Normal, false, 1n, 0);
    const bytes = encodeBytes128(original);
    expect(bytes.length).toBe(16);
    const decoded = decodeBytes128(bytes);
    expect(decoded.kind).toBe(Kind.Normal);
    expect(decoded.coefficient).toBe(1n);
    expect(decoded.exponent).toBe(0);
  });

  it("throws on short buffer", () => {
    expect(() => decodeBytes128(new Uint8Array(15))).toThrow();
  });

  it("throws on long buffer", () => {
    expect(() => decodeBytes128(new Uint8Array(17))).toThrow();
  });
});

// --- toString / fromString ---

describe("toString", () => {
  it("formats +1", () => {
    expect(toString(comp(Kind.Normal, false, 1n, 0))).toBe("+1E+0");
  });

  it("formats -123.45 (12345 * 10^-2)", () => {
    expect(toString(comp(Kind.Normal, true, 12345n, -2))).toBe("-1.2345E+2");
  });

  it("formats +0", () => {
    expect(toString(comp(Kind.Zero, false, 0n, 0))).toBe("+0");
  });

  it("formats +0E-5", () => {
    expect(toString(comp(Kind.Zero, false, 0n, -5))).toBe("+0E-5");
  });

  it("formats +Inf", () => {
    expect(toString(comp(Kind.Infinity, false))).toBe("+Inf");
  });

  it("formats -Inf", () => {
    expect(toString(comp(Kind.Infinity, true))).toBe("-Inf");
  });

  it("formats +NaN", () => {
    expect(toString(comp(Kind.QNaN, false))).toBe("+NaN");
  });

  it("formats +NaN123", () => {
    expect(toString(comp(Kind.QNaN, false, 0n, 0, 123n))).toBe("+NaN123");
  });

  it("formats +SNaN", () => {
    expect(toString(comp(Kind.SNaN, false))).toBe("+SNaN");
  });

  it("formats -SNaN456", () => {
    expect(toString(comp(Kind.SNaN, true, 0n, 0, 456n))).toBe("-SNaN456");
  });
});

describe("fromString", () => {
  it("parses +1", () => {
    const c = fromString("+1");
    expect(c.kind).toBe(Kind.Normal);
    expect(c.coefficient).toBe(1n);
    expect(c.exponent).toBe(0);
    expect(c.sign).toBe(false);
  });

  it("parses -123.45", () => {
    const c = fromString("-123.45");
    expect(c.kind).toBe(Kind.Normal);
    expect(c.sign).toBe(true);
    expect(c.coefficient).toBe(12345n);
    expect(c.exponent).toBe(-2);
  });

  it("parses 1.23E+5", () => {
    const c = fromString("1.23E+5");
    expect(c.kind).toBe(Kind.Normal);
    expect(c.coefficient).toBe(123n);
    expect(c.exponent).toBe(3);
  });

  it("parses INF", () => {
    expect(fromString("INF").kind).toBe(Kind.Infinity);
    expect(fromString("-Infinity").kind).toBe(Kind.Infinity);
    expect(fromString("-Infinity").sign).toBe(true);
  });

  it("parses NaN", () => {
    expect(fromString("NaN").kind).toBe(Kind.QNaN);
    expect(fromString("NaN123").payload).toBe(123n);
  });

  it("parses SNaN", () => {
    expect(fromString("SNaN").kind).toBe(Kind.SNaN);
    expect(fromString("SNaN456").payload).toBe(456n);
  });

  it("parses 0", () => {
    const c = fromString("0");
    expect(c.kind).toBe(Kind.Zero);
    expect(c.exponent).toBe(0);
  });

  it("parses 0.00", () => {
    const c = fromString("0.00");
    expect(c.kind).toBe(Kind.Zero);
    expect(c.exponent).toBe(-2);
  });

  it("throws on empty string", () => {
    expect(() => fromString("")).toThrow();
  });

  it("throws on invalid char", () => {
    expect(() => fromString("12x3")).toThrow();
  });

  it("throws on malformed payload, exponent, and decimal point", () => {
    for (const input of ["NaNabc", "SNaN-1", "1.2.3", "1E", "1Eabc", "1E2147483648", "1.0E2147483648"]) {
      expect(() => fromString(input)).toThrow();
    }
  });
});
