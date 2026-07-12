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
  it("encodes the payload field", () => {
    const nan = comp(Kind.QNaN, false, 0n, 0, 999n);
    const [lo, hi] = encode128(nan);
    expect(lo).toBe(999n);
    expect(hi).toBe(0x7c00000000000000n);
  });
});

// --- BID128 full 110-bit NaN payload ---

describe("BID128 full NaN payload (110-bit)", () => {
  const ten33 = 1000000000000000000000000000000000n; // 10^33
  const twoTo64 = 1n << 64n; // 2^64, first payload needing the high word
  const maxPayload = ten33 - 1n; // widest canonical BID128 NaN payload (33 nines)

  // NaN bit layout: payload = ((hi & 0x00003fffffffffff) << 64) | lo
  function nanBits(payload: bigint, snan: boolean, sign: boolean): [bigint, bigint] {
    const mask = snan ? 0x7e00000000000000n : 0x7c00000000000000n;
    const signBit = sign ? 0x8000000000000000n : 0n;
    const lo = payload & 0xffffffffffffffffn;
    const hi = signBit | mask | (payload >> 64n);
    return [lo, hi];
  }

  for (const payload of [twoTo64, maxPayload]) {
    for (const kind of [Kind.QNaN, Kind.SNaN]) {
      for (const sign of [false, true]) {
        it(`decode->encode bit roundtrip payload=${payload} kind=${kind} sign=${sign}`, () => {
          const [lo, hi] = nanBits(payload, kind === Kind.SNaN, sign);
          const decoded = decode128(lo, hi);
          expect(decoded.kind).toBe(kind);
          expect(decoded.sign).toBe(sign);
          expect(decoded.payload).toBe(payload);
          const [gotLo, gotHi] = encode128(decoded);
          expect(gotLo).toBe(lo);
          expect(gotHi).toBe(hi);
        });

        it(`toString->fromString roundtrip payload=${payload} kind=${kind} sign=${sign}`, () => {
          const original = comp(kind, sign, 0n, 0, payload);
          const parsed = fromString(toString(original));
          expect(parsed.kind).toBe(kind);
          expect(parsed.sign).toBe(sign);
          expect(parsed.payload).toBe(payload);
        });
      }
    }
  }

  it("renders and parses the 2^64 payload decimal string", () => {
    expect(toString(comp(Kind.QNaN, false, 0n, 0, twoTo64))).toBe("+NaN18446744073709551616");
    expect(fromString("NaN18446744073709551616").payload).toBe(twoTo64);
  });

  it("encode128 accepts 10^33-1 and rejects 10^33 (encode boundary)", () => {
    expect(encode128(comp(Kind.QNaN, false, 0n, 0, maxPayload))).toBeInstanceOf(Array);
    expect(() => encode128(comp(Kind.QNaN, false, 0n, 0, ten33))).toThrow(/exceeds max/);
  });

  it("fromString accepts 10^33-1 and rejects 10^33 (parse boundary)", () => {
    expect(fromString("NaN" + "9".repeat(33)).payload).toBe(maxPayload);
    // "NaN1000000000000000000000000000000000" is exactly 10^33 (1 + 33 zeros).
    expect(() => fromString("NaN1" + "0".repeat(33))).toThrow(/exceeds schema max/);
    expect(() => fromString("SNaN1" + "0".repeat(33))).toThrow(/exceeds schema max/);
  });

  it("decode128 normalizes a payload at or above 10^33 to 0 (non-canonical)", () => {
    // payHi = 0x00003fffffffffff, lo = all ones -> ~2^110-1, far above 10^33.
    const decoded = decode128(0xffffffffffffffffn, 0x7c00000000000000n | 0x00003fffffffffffn);
    expect(decoded.kind).toBe(Kind.QNaN);
    expect(decoded.payload).toBe(0n);
    // The exact 10^33 bit pattern also normalizes to 0 (boundary is inclusive).
    const [lo33, hi33] = nanBits(ten33, false, false);
    expect(decode128(lo33, hi33).payload).toBe(0n);
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
    for (const input of ["NaNabc", "SNaN-1", "1.2.3", "1E", "1Eabc", "1E2147483648"]) {
      expect(() => fromString(input)).toThrow();
    }
  });
});

// --- fromString strict ASCII grammar (reject contract) ---

describe("fromString strict ASCII grammar", () => {
  // Cross-language divergence regression cases: every one of these was accepted
  // (or silently altered) by at least one language's standard-library parser.
  const rejected: [string, string][] = [
    ["NaN+5", "sign inside NaN payload"],
    ["1E１", "fullwidth digit in exponent"],
    ["1E1_0", "underscore digit group in exponent"],
    ["１２３", "fullwidth mantissa digits"],
    ["NaN١٢", "arabic-indic digits in payload"],
    ["1E 5", "embedded whitespace in exponent"],
    [" 1", "leading U+00A0 (Unicode whitespace, not ASCII)"],
    ["½", "vulgar fraction one half"],
  ];

  for (const [input, why] of rejected) {
    it(`rejects ${JSON.stringify(input)} (${why})`, () => {
      expect(() => fromString(input)).toThrow();
    });
  }

  it("still accepts the valid ASCII grammar", () => {
    expect(fromString("1.5").coefficient).toBe(15n);
    expect(fromString("1.5").exponent).toBe(-1);
    expect(fromString("+1.23E+5").coefficient).toBe(123n);
    expect(fromString("+1.23E+5").exponent).toBe(3);
    expect(fromString("-inf").kind).toBe(Kind.Infinity);
    expect(fromString("-inf").sign).toBe(true);
    expect(fromString("NaN123").payload).toBe(123n);
    expect(fromString("1.").coefficient).toBe(1n);
    expect(fromString(".5").coefficient).toBe(5n);
    expect(fromString(".5").exponent).toBe(-1);
    expect(fromString("007").coefficient).toBe(7n);
  });

  it("trims ASCII whitespace but rejects Unicode whitespace", () => {
    expect(fromString(" \t1.5\n ").coefficient).toBe(15n); // ASCII ws stripped
    expect(() => fromString(" 1.5")).toThrow(); // NBSP leading -> reject
    expect(() => fromString("1.5 ")).toThrow(); // em-space trailing -> reject
  });
});

describe("fromString schema-wide coefficient cap (10^34-1)", () => {
  const max34 = 9999999999999999999999999999999999n;

  it("accepts 34 nines (exactly 10^34-1)", () => {
    const c = fromString("9".repeat(34));
    expect(c.kind).toBe(Kind.Normal);
    expect(c.coefficient).toBe(max34);
  });

  it("rejects 35 nines (above the schema cap)", () => {
    expect(() => fromString("9".repeat(35))).toThrow(/exceeds schema max/);
  });

  it("rejects exactly 10^34 (the smallest out-of-cap value)", () => {
    expect(() => fromString("1" + "0".repeat(34))).toThrow(/exceeds schema max/);
  });

  it("accepts 40 leading zeros + 1 (value-based, not digit-count-based)", () => {
    const c = fromString("0".repeat(40) + "1");
    expect(c.kind).toBe(Kind.Normal);
    expect(c.coefficient).toBe(1n);
  });

  it("rejects a 40-digit value (cross-language fixed-width wrap regression)", () => {
    expect(() => fromString("1" + "0".repeat(39))).toThrow(/exceeds schema max/);
  });
});

describe("fromString exponent: shared 2^53 literal bound, int32 final exponent", () => {
  // The exponent literal must be below the shared exact-integer bound 2^53 in
  // magnitude (the fromString grammar constant every language consumer
  // enforces; Number.isSafeInteger pins it here — see parseExponentLiteral);
  // only the fraction-adjusted FINAL exponent must fit int32, so every
  // toString rendering (adjusted exponent up to int32 max + 33) reparses
  // (round-trip closure).
  it("accepts a literal past int32 when the adjusted final exponent fits", () => {
    // literal 2147483649 > int32 max; adjusted 2147483649-3 = 2147483646 fits.
    expect(fromString("0.001E2147483649").exponent).toBe(2147483646);
    // the exact shape toString renders at the top edge.
    expect(fromString("1.0E2147483648").exponent).toBe(2147483647);
  });

  it("accepts int32-max literal when the adjusted value fits", () => {
    const c = fromString("1.5E2147483647");
    expect(c.coefficient).toBe(15n);
    expect(c.exponent).toBe(2147483646);
  });

  it("rejects an adjusted final exponent outside int32 (one past either edge)", () => {
    // literal -2147483648 parses; adjusted -2147483648-1 = -2147483649 does not fit.
    expect(() => fromString("1.0E-2147483648")).toThrow();
    // literal 2147483649 parses; adjusted 2147483649-1 = 2147483648 does not fit.
    expect(() => fromString("1.0E+2147483649")).toThrow();
  });

  it("rejects a literal at/beyond the shared 2^53 bound through the same error channel", () => {
    expect(() => fromString("1E9007199254740992")).toThrow(); // +2^53, at the shared bound
    expect(() => fromString("1E-9007199254740992")).toThrow(); // -2^53, at the shared bound
    expect(() => fromString("1E" + "9".repeat(25))).toThrow(); // far past the bound
  });

  it("closes the round-trip at the rendered edge (parse(render(x)) fixed point)", () => {
    const rendered = toString(fromString("10E2147483647"));
    expect(rendered).toBe("+1.0E+2147483648");
    expect(toString(fromString(rendered))).toBe(rendered);
  });
});

// --- encode reject contract ---

// Build a Components with arbitrary (possibly ill-typed) fields for reject tests.
function raw(fields: Partial<Record<keyof Components, unknown>>): Components {
  return {
    sign: false,
    coefficient: 0n,
    exponent: 0,
    kind: Kind.Normal,
    payload: 0n,
    ...fields,
  } as unknown as Components;
}

describe("encode32 reject contract", () => {
  it("rejects coefficient above 10^7-1 but accepts the max", () => {
    expect(encode32(comp(Kind.Normal, false, 9999999n, 0))).toBeTypeOf("number");
    expect(() => encode32(comp(Kind.Normal, false, 10000000n, 0))).toThrow(
      /coefficient 10000000 exceeds max 9999999/,
    );
  });

  it("rejects exponent outside [-101,+90] but accepts the bounds", () => {
    expect(encode32(comp(Kind.Normal, false, 1n, 90))).toBeTypeOf("number");
    expect(encode32(comp(Kind.Normal, false, 1n, -101))).toBeTypeOf("number");
    expect(() => encode32(comp(Kind.Normal, false, 1n, 91))).toThrow(/exponent 91 outside range/);
    expect(() => encode32(comp(Kind.Normal, false, 1n, -102))).toThrow(
      /exponent -102 outside range/,
    );
  });

  it("rejects NaN payload at or above 10^6 but accepts the max", () => {
    expect(encode32(comp(Kind.QNaN, false, 0n, 0, 999999n))).toBeTypeOf("number");
    expect(() => encode32(comp(Kind.QNaN, false, 0n, 0, 1000000n))).toThrow(
      /payload 1000000 exceeds max 999999/,
    );
    expect(() => encode32(comp(Kind.SNaN, false, 0n, 0, 1000000n))).toThrow();
  });

  it("rejects negative, non-bigint, non-integer, and unknown-kind fields", () => {
    expect(() => encode32(raw({ kind: Kind.Normal, coefficient: -1n }))).toThrow(/is negative/);
    expect(() => encode32(raw({ kind: Kind.QNaN, payload: -1n }))).toThrow(/is negative/);
    expect(() => encode32(raw({ kind: Kind.Normal, coefficient: 5 }))).toThrow(
      /coefficient must be a bigint/,
    );
    expect(() => encode32(raw({ kind: Kind.QNaN, payload: 5 }))).toThrow(/payload must be a bigint/);
    expect(() => encode32(raw({ kind: Kind.Normal, coefficient: 1n, exponent: 1.5 }))).toThrow(
      /not an integer/,
    );
    expect(() => encode32(raw({ kind: Kind.Normal, coefficient: 1n, exponent: NaN }))).toThrow(
      /not an integer/,
    );
    expect(() => encode32(raw({ kind: Kind.Normal, coefficient: 1n, exponent: Infinity }))).toThrow(
      /not an integer/,
    );
    expect(() => encode32(raw({ kind: 99 }))).toThrow(/unrecognized kind 99/);
  });

  it("rejects a BID64 sNaN payload fed into encode32 (no crash, no truncation)", () => {
    // Reachable normal-usage path: decode64 a canonical sNaN whose payload exceeds
    // the BID32 payload domain, then encode32 the resulting Components.
    const c = decode64(0x7e00000000000000n | 5000000000n);
    expect(c.kind).toBe(Kind.SNaN);
    expect(c.payload).toBe(5000000000n);
    expect(() => encode32(c)).toThrow(/payload 5000000000 exceeds max 999999/);
  });
});

describe("encode64 reject contract", () => {
  it("rejects coefficient above 10^16-1 but accepts the max", () => {
    expect(encode64(comp(Kind.Normal, false, 9999999999999999n, 0))).toBeTypeOf("bigint");
    expect(() => encode64(comp(Kind.Normal, false, 10000000000000000n, 0))).toThrow(
      /exceeds max 9999999999999999/,
    );
  });

  it("rejects exponent outside [-398,+369] but accepts the bounds", () => {
    expect(encode64(comp(Kind.Normal, false, 1n, 369))).toBeTypeOf("bigint");
    expect(encode64(comp(Kind.Normal, false, 1n, -398))).toBeTypeOf("bigint");
    expect(() => encode64(comp(Kind.Normal, false, 1n, 370))).toThrow(/outside range/);
    expect(() => encode64(comp(Kind.Normal, false, 1n, -399))).toThrow(/outside range/);
  });

  it("rejects NaN payload at or above 10^15 but accepts the max", () => {
    expect(encode64(comp(Kind.QNaN, false, 0n, 0, 999999999999999n))).toBeTypeOf("bigint");
    expect(() => encode64(comp(Kind.QNaN, false, 0n, 0, 1000000000000000n))).toThrow(
      /exceeds max 999999999999999/,
    );
  });
});

describe("encode128 reject contract", () => {
  const max128 = 9999999999999999999999999999999999n; // 10^34 - 1

  it("rejects coefficient above 10^34-1 but accepts the max", () => {
    expect(encode128(comp(Kind.Normal, false, max128, 0))).toBeInstanceOf(Array);
    expect(() => encode128(comp(Kind.Normal, false, max128 + 1n, 0))).toThrow(
      /exceeds max 9999999999999999999999999999999999/,
    );
    // Well above 2^128, previously silently truncated by the high-bit mask.
    expect(() => encode128(comp(Kind.Normal, false, 1n << 200n, 0))).toThrow();
  });

  it("rejects exponent outside [-6176,+6111] but accepts the bounds", () => {
    expect(encode128(comp(Kind.Normal, false, 1n, 6111))).toBeInstanceOf(Array);
    expect(encode128(comp(Kind.Normal, false, 1n, -6176))).toBeInstanceOf(Array);
    expect(() => encode128(comp(Kind.Normal, false, 1n, 6112))).toThrow(/outside range/);
    expect(() => encode128(comp(Kind.Normal, false, 1n, -6177))).toThrow(/outside range/);
  });

  it("accepts a high NaN payload below 10^33 but rejects one at or above it", () => {
    const ten33 = 1000000000000000000000000000000000n; // 10^33
    // Payloads above 2^64 are now in range (full 110-bit payload), not truncated.
    expect(encode128(comp(Kind.QNaN, false, 0n, 0, 0xffffffffffffffffn))).toBeInstanceOf(Array);
    expect(encode128(comp(Kind.QNaN, false, 0n, 0, 1n << 64n))).toBeInstanceOf(Array);
    expect(encode128(comp(Kind.QNaN, false, 0n, 0, ten33 - 1n))).toBeInstanceOf(Array);
    expect(encode128(comp(Kind.SNaN, false, 0n, 0, ten33 - 1n))).toBeInstanceOf(Array);
    // At or above 10^33 is non-canonical and rejected (no silent truncation).
    expect(() => encode128(comp(Kind.QNaN, false, 0n, 0, ten33))).toThrow(/exceeds max/);
    expect(() => encode128(comp(Kind.SNaN, false, 0n, 0, ten33))).toThrow(/exceeds max/);
    // Negative payload rejected (unsigned field).
    expect(() => encode128(comp(Kind.QNaN, false, 0n, 0, -1n))).toThrow(/is negative/);
  });
});

describe("encodeBytes reject contract", () => {
  it("propagates encode rejection through the byte helpers", () => {
    expect(() => encodeBytes32(comp(Kind.Normal, false, 10000000n, 0))).toThrow();
    expect(() => encodeBytes64(comp(Kind.Normal, false, 10000000000000000n, 0))).toThrow();
    expect(() =>
      encodeBytes128(comp(Kind.Normal, false, 10000000000000000000000000000000000n, 0)),
    ).toThrow();
  });
});

describe("encode in-range bits are unchanged", () => {
  it("keeps canonical encodings identical", () => {
    expect(encode32(comp(Kind.Normal, false, 1n, 0))).toBe(0x32800001);
    expect(encode32(comp(Kind.Zero, true, 0n, 0))).toBe(0xb2800000 >>> 0);
    expect(encode32(comp(Kind.QNaN, false, 0n, 0, 1n))).toBe(0x7c000001);
    expect(encode64(comp(Kind.Normal, false, 1n, 0))).toBe(0x31c0000000000001n);
    expect(encode64(comp(Kind.Zero, true, 0n, 0))).toBe(0xb1c0000000000000n);
    const [lo, hi] = encode128(comp(Kind.Normal, false, 1n, 0));
    expect(lo).toBe(1n);
    expect(hi).toBe(0x3040000000000000n);
    const [zlo, zhi] = encode128(comp(Kind.Zero, true, 0n, -6176));
    expect(zlo).toBe(0n);
    expect(zhi).toBe(0x8000000000000000n);
  });
});
