// BID (Binary Integer Decimal) codec for IEEE 754 decimal floating-point.
// Mechanical translation of Go implementation: bidcodec/decimal.go

export enum Kind {
  Normal = 0,
  Zero = 1,
  Infinity = 2,
  QNaN = 3,
  SNaN = 4,
}

export interface Components {
  sign: boolean;
  coefficient: bigint;
  exponent: number;
  kind: Kind;
  payload: bigint;
}

function c(
  sign: boolean,
  kind: Kind,
  coefficient: bigint = 0n,
  exponent: number = 0,
  payload: bigint = 0n,
): Components {
  return { sign, coefficient, exponent, kind, payload };
}

// --- BID32 ---

const bid32NaNMask = 0x7c000000;
const bid32SNaNMask = 0x7e000000;
const bid32InfMask = 0x78000000;
const bid32SignMask = 0x80000000;
const bid32SteerMask = 0x60000000;
const bid32ExpMask32 = 0xff;
const bid32Bias = 101;

export function decode32(v: number): Components {
  // Ensure unsigned 32-bit
  v = v >>> 0;

  const sign = (v & bid32SignMask) !== 0;

  // NaN
  if ((v & bid32NaNMask) === bid32NaNMask) {
    const kind = (v & bid32SNaNMask) === bid32SNaNMask ? Kind.SNaN : Kind.QNaN;
    let payload = BigInt(v & 0x000fffff);
    if (payload > 999999n) {
      payload = 0n; // non-canonical
    }
    return c(sign, kind, 0n, 0, payload);
  }

  // Infinity
  if ((v & bid32InfMask) === bid32InfMask) {
    return c(sign, Kind.Infinity);
  }

  let exp: number;
  let coeff: number;
  if ((v & bid32SteerMask) === bid32SteerMask) {
    // special encoding (implicit high bit)
    exp = (v >>> 21) & bid32ExpMask32;
    coeff = (v & 0x001fffff) | 0x00800000;
    if (coeff >= 10000000) {
      coeff = 0; // non-canonical
    }
  } else {
    exp = (v >>> 23) & bid32ExpMask32;
    coeff = v & 0x007fffff;
  }

  if (coeff === 0) {
    return c(sign, Kind.Zero, 0n, exp - bid32Bias);
  }
  return c(sign, Kind.Normal, BigInt(coeff), exp - bid32Bias);
}

export function encode32(comp: Components): number {
  let sgn = 0;
  if (comp.sign) {
    sgn = bid32SignMask;
  }

  switch (comp.kind) {
    case Kind.Infinity:
      return (sgn | 0x78000000) >>> 0;
    case Kind.QNaN:
      return (sgn | 0x7c000000 | (Number(comp.payload) & 0x000fffff)) >>> 0;
    case Kind.SNaN:
      return (sgn | 0x7e000000 | (Number(comp.payload) & 0x000fffff)) >>> 0;
    case Kind.Zero: {
      let exp = comp.exponent + bid32Bias;
      if (exp < 0) exp = 0;
      else if (exp > 191) exp = 191;
      return (sgn | (exp << 23)) >>> 0;
    }
  }

  // Normal
  const coeff = Number(comp.coefficient);
  let exp = comp.exponent + bid32Bias;
  if (exp < 0) exp = 0;
  else if (exp > 191) exp = 191;

  if (coeff < 0x800000) {
    return (sgn | (exp << 23) | coeff) >>> 0;
  }
  return (sgn | 0x60000000 | (exp << 21) | (coeff & 0x001fffff)) >>> 0;
}

// --- BID64 ---

const bid64NaNMask = 0x7c00000000000000n;
const bid64SNaNMask = 0x7e00000000000000n;
const bid64InfMask = 0x7800000000000000n;
const bid64SignMask = 0x8000000000000000n;
const bid64SteerMask = 0x6000000000000000n;
const bid64ExpMask = 0x3ffn;
const bid64MaxCoeff = 9999999999999999n;
const bid64Bias = 398;

export function decode64(v: bigint): Components {
  v = BigInt.asUintN(64, v);

  const sign = (v & bid64SignMask) !== 0n;

  // NaN
  if ((v & bid64NaNMask) === bid64NaNMask) {
    const kind = (v & bid64SNaNMask) === bid64SNaNMask ? Kind.SNaN : Kind.QNaN;
    let payload = v & 0x0003ffffffffffffn;
    if (payload > 999999999999999n) {
      payload = 0n; // non-canonical
    }
    return c(sign, kind, 0n, 0, payload);
  }

  // Infinity
  if ((v & bid64InfMask) === bid64InfMask) {
    return c(sign, Kind.Infinity);
  }

  let exp: number;
  let coeff: bigint;
  if ((v & bid64SteerMask) === bid64SteerMask) {
    exp = Number((v >> 51n) & bid64ExpMask);
    coeff = (v & 0x0007ffffffffffffn) | 0x0020000000000000n;
    if (coeff > bid64MaxCoeff) {
      coeff = 0n; // non-canonical
    }
  } else {
    exp = Number((v >> 53n) & bid64ExpMask);
    coeff = v & 0x001fffffffffffffn;
  }

  if (coeff === 0n) {
    return c(sign, Kind.Zero, 0n, exp - bid64Bias);
  }
  return c(sign, Kind.Normal, coeff, exp - bid64Bias);
}

export function encode64(comp: Components): bigint {
  let sgn = 0n;
  if (comp.sign) {
    sgn = bid64SignMask;
  }

  switch (comp.kind) {
    case Kind.Infinity:
      return BigInt.asUintN(64, sgn | 0x7800000000000000n);
    case Kind.QNaN:
      return BigInt.asUintN(
        64,
        sgn | 0x7c00000000000000n | (comp.payload & 0x0003ffffffffffffn),
      );
    case Kind.SNaN:
      return BigInt.asUintN(
        64,
        sgn | 0x7e00000000000000n | (comp.payload & 0x0003ffffffffffffn),
      );
    case Kind.Zero: {
      let exp = comp.exponent + bid64Bias;
      if (exp < 0) exp = 0;
      else if (exp > 767) exp = 767;
      return BigInt.asUintN(64, sgn | (BigInt(exp) << 53n));
    }
  }

  // Normal
  const coeff = comp.coefficient;
  let exp = comp.exponent + bid64Bias;
  if (exp < 0) exp = 0;
  else if (exp > 767) exp = 767;

  if (coeff < 0x20000000000000n) {
    return BigInt.asUintN(64, sgn | (BigInt(exp) << 53n) | coeff);
  }
  return BigInt.asUintN(
    64,
    sgn | bid64SteerMask | (BigInt(exp) << 51n) | (coeff & 0x0007ffffffffffffn),
  );
}

// --- BID128 ---

const bid128NaNMask = 0x7c00000000000000n;
const bid128SNaNMask = 0x7e00000000000000n;
const bid128InfMask = 0x7800000000000000n;
const bid128SignMask = 0x8000000000000000n;
const bid128SteerMask = 0x6000000000000000n;
const bid128ExpMask = 0x3fffn;
const bid128Bias = 6176;

const ten34 = 10000000000000000000000000000000000n; // 10^34
const ten33 = 1000000000000000000000000000000000n; // 10^33

export function decode128(lo: bigint, hi: bigint): Components {
  lo = BigInt.asUintN(64, lo);
  hi = BigInt.asUintN(64, hi);

  const sign = (hi & bid128SignMask) !== 0n;

  // NaN
  if ((hi & bid128NaNMask) === bid128NaNMask) {
    const kind = (hi & bid128SNaNMask) === bid128SNaNMask ? Kind.SNaN : Kind.QNaN;
    // payload: hi[45:0] and lo[63:0] = 110 bits
    const payHi = hi & 0x00003fffffffffffn;
    const coeff = (payHi << 64n) | lo;
    if (coeff >= ten33) {
      return c(sign, kind);
    }
    return c(sign, kind, 0n, 0, lo);
  }

  // Infinity
  if ((hi & bid128InfMask) === bid128InfMask) {
    return c(sign, Kind.Infinity);
  }

  let exp: number;
  let coeffHi: bigint;
  if ((hi & bid128SteerMask) === bid128SteerMask) {
    exp = Number((hi >> 47n) & bid128ExpMask);
    coeffHi = (hi & 0x00007fffffffffffn) | 0x0020000000000000n;
  } else {
    exp = Number((hi >> 49n) & bid128ExpMask);
    coeffHi = hi & 0x0001ffffffffffffn;
  }

  let coeff = (coeffHi << 64n) | lo;

  if (coeff >= ten34) {
    coeff = 0n;
  }

  if (coeff === 0n) {
    return c(sign, Kind.Zero, 0n, exp - bid128Bias);
  }
  return c(sign, Kind.Normal, coeff, exp - bid128Bias);
}

export function encode128(comp: Components): [bigint, bigint] {
  let sgn = 0n;
  if (comp.sign) {
    sgn = bid128SignMask;
  }

  switch (comp.kind) {
    case Kind.Infinity:
      return [0n, BigInt.asUintN(64, sgn | 0x7800000000000000n)];
    case Kind.QNaN: {
      return [
        BigInt.asUintN(64, comp.payload),
        BigInt.asUintN(64, sgn | 0x7c00000000000000n),
      ];
    }
    case Kind.SNaN: {
      return [
        BigInt.asUintN(64, comp.payload),
        BigInt.asUintN(64, sgn | 0x7e00000000000000n),
      ];
    }
    case Kind.Zero: {
      let exp = comp.exponent + bid128Bias;
      if (exp < 0) exp = 0;
      else if (exp > 12287) exp = 12287;
      return [0n, BigInt.asUintN(64, sgn | (BigInt(exp) << 49n))];
    }
  }

  // Normal: coefficient as 128 bits
  const coeffHi = comp.coefficient >> 64n;
  const coeffLo = comp.coefficient & 0xffffffffffffffffn;

  let exp = comp.exponent + bid128Bias;
  if (exp < 0) exp = 0;
  else if (exp > 12287) exp = 12287;

  const lo = BigInt.asUintN(64, coeffLo);
  const hi = BigInt.asUintN(
    64,
    sgn | (BigInt(exp) << 49n) | (coeffHi & 0x0001ffffffffffffn),
  );
  return [lo, hi];
}

// --- Byte encoding/decoding (little-endian Uint8Array) ---

export function decodeBytes32(buf: Uint8Array): Components {
  if (buf.length !== 4) throw new Error(`decodeBytes32: expected 4 bytes, got ${buf.length}`);
  const v = buf[0] | (buf[1] << 8) | (buf[2] << 16) | (buf[3] << 24);
  return decode32(v >>> 0);
}

export function encodeBytes32(comp: Components): Uint8Array {
  const v = encode32(comp);
  const buf = new Uint8Array(4);
  buf[0] = v & 0xff;
  buf[1] = (v >>> 8) & 0xff;
  buf[2] = (v >>> 16) & 0xff;
  buf[3] = (v >>> 24) & 0xff;
  return buf;
}

export function decodeBytes64(buf: Uint8Array): Components {
  if (buf.length !== 8) throw new Error(`decodeBytes64: expected 8 bytes, got ${buf.length}`);
  const dv = new DataView(buf.buffer, buf.byteOffset, buf.byteLength);
  const v = dv.getBigUint64(0, true); // little-endian
  return decode64(v);
}

export function encodeBytes64(comp: Components): Uint8Array {
  const v = encode64(comp);
  const buf = new Uint8Array(8);
  const dv = new DataView(buf.buffer, buf.byteOffset, buf.byteLength);
  dv.setBigUint64(0, v, true); // little-endian
  return buf;
}

export function decodeBytes128(buf: Uint8Array): Components {
  if (buf.length !== 16) throw new Error(`decodeBytes128: expected 16 bytes, got ${buf.length}`);
  const dv = new DataView(buf.buffer, buf.byteOffset, buf.byteLength);
  const lo = dv.getBigUint64(0, true); // little-endian
  const hi = dv.getBigUint64(8, true);
  return decode128(lo, hi);
}

export function encodeBytes128(comp: Components): Uint8Array {
  const [lo, hi] = encode128(comp);
  const buf = new Uint8Array(16);
  const dv = new DataView(buf.buffer, buf.byteOffset, buf.byteLength);
  dv.setBigUint64(0, lo, true); // little-endian
  dv.setBigUint64(8, hi, true);
  return buf;
}

// --- IEEE 754 string conversion ---

export function toString(comp: Components): string {
  const prefix = comp.sign ? "-" : "+";

  switch (comp.kind) {
    case Kind.Infinity:
      return prefix + "Inf";
    case Kind.QNaN:
      if (comp.payload !== 0n) {
        return `${prefix}NaN${comp.payload}`;
      }
      return prefix + "NaN";
    case Kind.SNaN:
      if (comp.payload !== 0n) {
        return `${prefix}SNaN${comp.payload}`;
      }
      return prefix + "SNaN";
    case Kind.Zero:
      if (comp.exponent === 0) {
        return prefix + "0";
      }
      return `${prefix}0E${comp.exponent >= 0 ? "+" : ""}${comp.exponent}`;
  }

  // Normal
  const digits = comp.coefficient.toString();
  const exp = comp.exponent + digits.length - 1;
  const expStr = `${exp >= 0 ? "+" : ""}${exp}`;
  if (digits.length === 1) {
    return `${prefix}${digits}E${expStr}`;
  }
  return `${prefix}${digits[0]}.${digits.slice(1)}E${expStr}`;
}

export function fromString(s: string): Components {
  s = s.trim();
  if (s.length === 0) throw new Error("empty string");

  let sign = false;
  if (s[0] === "+") {
    s = s.slice(1);
  } else if (s[0] === "-") {
    sign = true;
    s = s.slice(1);
  }

  const upper = s.toUpperCase();
  if (upper === "INF" || upper === "INFINITY") {
    return c(sign, Kind.Infinity);
  }
  if (upper.startsWith("SNAN")) {
    const payload = s.length > 4 ? parseUint64Payload(s.slice(4)) : 0n;
    return c(sign, Kind.SNaN, 0n, 0, payload);
  }
  if (upper.startsWith("NAN")) {
    const payload = s.length > 3 ? parseUint64Payload(s.slice(3)) : 0n;
    return c(sign, Kind.QNaN, 0n, 0, payload);
  }

  // Parse number: digits, decimal point, exponent
  let digits = "";
  let expAdjust = 0;
  let foundDot = false;
  let i = 0;
  while (i < s.length && s[i] !== "E" && s[i] !== "e") {
    if (s[i] === ".") {
      if (foundDot) throw new Error("multiple decimal points");
      foundDot = true;
    } else if (s[i] >= "0" && s[i] <= "9") {
      digits += s[i];
      if (foundDot) {
        expAdjust--;
      }
    } else {
      throw new Error(`unexpected character: ${s[i]}`);
    }
    i++;
  }

  let expPart = 0;
  if (i < s.length && (s[i] === "E" || s[i] === "e")) {
    i++;
    expPart = parseInt32Exponent(s.slice(i));
  }

  if (digits.length === 0) throw new Error("no digits");

  // Remove leading zeros
  let start = 0;
  while (start < digits.length - 1 && digits[start] === "0") {
    start++;
  }
  digits = digits.slice(start);

  const coeff = BigInt(digits);
  const exponent = checkedInt32(expPart + expAdjust, "exponent");

  if (coeff === 0n) {
    return c(sign, Kind.Zero, 0n, exponent);
  }
  return c(sign, Kind.Normal, coeff, exponent);
}

function parseUint64Payload(s: string): bigint {
  if (!/^[0-9]+$/.test(s)) throw new Error(`invalid NaN payload: ${s}`);
  const payload = BigInt(s);
  if (payload > 0xffffffffffffffffn) throw new Error(`invalid NaN payload: ${s}`);
  return payload;
}

function parseInt32Exponent(s: string): number {
  if (!/^[+-]?[0-9]+$/.test(s)) throw new Error(`invalid exponent: ${s}`);
  return checkedInt32(Number(s), "exponent");
}

function checkedInt32(n: number, label: string): number {
  if (!Number.isSafeInteger(n) || n < -2147483648 || n > 2147483647) {
    throw new Error(`${label} out of int32 range: ${n}`);
  }
  return n;
}
