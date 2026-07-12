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

// --- Encode validation (reject contract) ---
//
// Encode* are validating packing APIs: a Components value whose fields are not
// representable in the target BID width is rejected (thrown), never silently
// truncated, masked, clamped, or coerced, and never crashes the process.
// Validation runs before any narrowing numeric conversion, so in-range
// Components encode to exactly the same bits as before.

const bid32EncMaxCoeff = 9999999n; // 10^7 - 1
const bid32EncMaxPayload = 999999n; // 10^6 - 1 (payload at/above 10^6 is non-canonical)
const bid32EncExpMin = -101;
const bid32EncExpMax = 90;

const bid64EncMaxCoeff = 9999999999999999n; // 10^16 - 1
const bid64EncMaxPayload = 999999999999999n; // 10^15 - 1
const bid64EncExpMin = -398;
const bid64EncExpMax = 369;

const bid128EncMaxCoeff = 9999999999999999999999999999999999n; // 10^34 - 1
// The schema-wide maximum coefficient equals the coefficient maximum of the
// widest supported BID width (BID128). fromString enforces it as a schema
// limit so all six languages parse the identical value space; encode* enforce
// the per-width ranges.
const schemaMaxCoefficient = bid128EncMaxCoeff;
// The BID128 Components.payload holds the full 110-bit NaN payload. The BID128
// canonical NaN payload limit is 10^33 (the widest canonical BID128 NaN
// payload); a payload at or above it is non-canonical and rejected, mirroring
// the BID32/BID64 payload caps above. encode128 splits an in-range payload into
// its low and high 64-bit words. This is the schema-wide NaN payload limit and
// equals fromString's payload cap.
const bid128EncMaxPayload = 999999999999999999999999999999999n; // 10^33 - 1
const bid128EncExpMin = -6176;
const bid128EncExpMax = 6111;

const validKinds: ReadonlySet<Kind> = new Set([
  Kind.Normal,
  Kind.Zero,
  Kind.Infinity,
  Kind.QNaN,
  Kind.SNaN,
]);

function checkKind(kind: Kind, operation: string): void {
  if (!validKinds.has(kind)) {
    throw new Error(`${operation}: unrecognized kind ${String(kind)}`);
  }
}

function checkSign(sign: unknown, operation: string): void {
  if (typeof sign !== "boolean") {
    throw new Error(`${operation}: sign must be a boolean, got ${typeof sign}`);
  }
}

function checkZeroBigIntField(
  value: unknown,
  field: string,
  kind: Kind,
  operation: string,
): void {
  if (typeof value !== "bigint") {
    throw new Error(`${operation}: ${field} must be a bigint, got ${typeof value}`);
  }
  if (value !== 0n) {
    throw new Error(`${operation}: ${Kind[kind]} cannot carry ${field} ${value}`);
  }
}

function checkZeroExponent(exp: number, kind: Kind, operation: string): void {
  if (!Number.isInteger(exp)) {
    throw new Error(`${operation}: exponent ${exp} is not an integer`);
  }
  if (exp !== 0) {
    throw new Error(`${operation}: ${Kind[kind]} cannot carry exponent ${exp}`);
  }
}

function checkKindFields(comp: Components, operation: string): void {
  switch (comp.kind) {
    case Kind.Normal:
      if (comp.coefficient === 0n) {
        throw new Error(`${operation}: Normal cannot carry a zero coefficient`);
      }
      checkZeroBigIntField(comp.payload, "NaN payload", comp.kind, operation);
      break;
    case Kind.Zero:
      checkZeroBigIntField(comp.coefficient, "coefficient", comp.kind, operation);
      checkZeroBigIntField(comp.payload, "NaN payload", comp.kind, operation);
      break;
    case Kind.Infinity:
      checkZeroBigIntField(comp.coefficient, "coefficient", comp.kind, operation);
      checkZeroExponent(comp.exponent, comp.kind, operation);
      checkZeroBigIntField(comp.payload, "NaN payload", comp.kind, operation);
      break;
    case Kind.QNaN:
    case Kind.SNaN:
      checkZeroBigIntField(comp.coefficient, "coefficient", comp.kind, operation);
      checkZeroExponent(comp.exponent, comp.kind, operation);
      break;
  }
}

function checkEncExponent(exp: number, min: number, max: number, width: string): number {
  // Rejects NaN, Infinity, and non-integers (e.g. 1.5) as well as out-of-range.
  if (!Number.isInteger(exp)) {
    throw new Error(`${width} encode: exponent ${exp} is not an integer`);
  }
  if (exp < min || exp > max) {
    throw new Error(`${width} encode: exponent ${exp} outside range [${min},${max}]`);
  }
  return exp;
}

function checkEncCoefficient(coeff: unknown, max: bigint, width: string): bigint {
  if (typeof coeff !== "bigint") {
    throw new Error(`${width} encode: coefficient must be a bigint, got ${typeof coeff}`);
  }
  if (coeff < 0n) {
    throw new Error(`${width} encode: coefficient ${coeff} is negative`);
  }
  if (coeff > max) {
    throw new Error(`${width} encode: coefficient ${coeff} exceeds max ${max}`);
  }
  return coeff;
}

function checkEncPayload(payload: unknown, max: bigint, width: string): bigint {
  if (typeof payload !== "bigint") {
    throw new Error(`${width} encode: payload must be a bigint, got ${typeof payload}`);
  }
  if (payload < 0n) {
    throw new Error(`${width} encode: payload ${payload} is negative`);
  }
  if (payload > max) {
    throw new Error(`${width} encode: payload ${payload} exceeds max ${max}`);
  }
  return payload;
}

// --- BID32 ---

const bid32NaNMask = 0x7c000000;
const bid32SNaNMask = 0x7e000000;
const bid32InfMask = 0x78000000;
const bid32SignMask = 0x80000000;
const bid32SteerMask = 0x60000000;
const bid32ExpMask32 = 0xff;
const bid32Bias = 101;

function checkUint32Word(value: unknown, operation: string): number {
  if (typeof value !== "number" || !Number.isInteger(value)) {
    throw new Error(`${operation}: word must be an integer number`);
  }
  if (value < 0 || value > 0xffffffff) {
    throw new Error(`${operation}: word ${value} outside unsigned 32-bit range [0,4294967295]`);
  }
  return value;
}

function checkUint64Word(value: unknown, operation: string, field: string): bigint {
  if (typeof value !== "bigint") {
    throw new Error(`${operation}: ${field} must be a bigint, got ${typeof value}`);
  }
  if (value < 0n || value > 0xffffffffffffffffn) {
    throw new Error(
      `${operation}: ${field} ${value} outside unsigned 64-bit range [0,18446744073709551615]`,
    );
  }
  return value;
}

export function decode32(v: number): Components {
  v = checkUint32Word(v, "bid32 decode");

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
  checkSign(comp.sign, "bid32 encode");
  checkKind(comp.kind, "bid32 encode");
  checkKindFields(comp, "bid32 encode");

  let sgn = 0;
  if (comp.sign) {
    sgn = bid32SignMask;
  }

  switch (comp.kind) {
    case Kind.Infinity:
      return (sgn | 0x78000000) >>> 0;
    case Kind.QNaN: {
      const payload = checkEncPayload(comp.payload, bid32EncMaxPayload, "bid32");
      return (sgn | 0x7c000000 | Number(payload)) >>> 0;
    }
    case Kind.SNaN: {
      const payload = checkEncPayload(comp.payload, bid32EncMaxPayload, "bid32");
      return (sgn | 0x7e000000 | Number(payload)) >>> 0;
    }
    case Kind.Zero: {
      const exp =
        checkEncExponent(comp.exponent, bid32EncExpMin, bid32EncExpMax, "bid32") + bid32Bias;
      return (sgn | (exp << 23)) >>> 0;
    }
  }

  // Normal: validate before the narrowing Number() conversion so an out-of-range
  // coefficient can never lose precision or be masked into a wrong value.
  const coeff = Number(checkEncCoefficient(comp.coefficient, bid32EncMaxCoeff, "bid32"));
  const exp =
    checkEncExponent(comp.exponent, bid32EncExpMin, bid32EncExpMax, "bid32") + bid32Bias;

  if (coeff < 0x800000) {
    return (sgn | (exp << 23) | coeff) >>> 0;
  }
  // coeff is in [0x800000, 9999999]; the 0x800000 bit is implicit in the special
  // encoding, so `& 0x001fffff` strips only that known implicit bit, not data.
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
  v = checkUint64Word(v, "bid64 decode", "word");

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
  checkSign(comp.sign, "bid64 encode");
  checkKind(comp.kind, "bid64 encode");
  checkKindFields(comp, "bid64 encode");

  let sgn = 0n;
  if (comp.sign) {
    sgn = bid64SignMask;
  }

  switch (comp.kind) {
    case Kind.Infinity:
      return BigInt.asUintN(64, sgn | 0x7800000000000000n);
    case Kind.QNaN: {
      const payload = checkEncPayload(comp.payload, bid64EncMaxPayload, "bid64");
      return BigInt.asUintN(64, sgn | 0x7c00000000000000n | payload);
    }
    case Kind.SNaN: {
      const payload = checkEncPayload(comp.payload, bid64EncMaxPayload, "bid64");
      return BigInt.asUintN(64, sgn | 0x7e00000000000000n | payload);
    }
    case Kind.Zero: {
      const exp =
        checkEncExponent(comp.exponent, bid64EncExpMin, bid64EncExpMax, "bid64") + bid64Bias;
      return BigInt.asUintN(64, sgn | (BigInt(exp) << 53n));
    }
  }

  // Normal
  const coeff = checkEncCoefficient(comp.coefficient, bid64EncMaxCoeff, "bid64");
  const exp =
    checkEncExponent(comp.exponent, bid64EncExpMin, bid64EncExpMax, "bid64") + bid64Bias;

  if (coeff < 0x20000000000000n) {
    return BigInt.asUintN(64, sgn | (BigInt(exp) << 53n) | coeff);
  }
  // coeff is in [2^53, 10^16-1]; the 2^53 bit is implicit in the steer encoding,
  // so `& 0x0007ffffffffffff` strips only that known implicit bit, not data.
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
  lo = checkUint64Word(lo, "bid128 decode", "lo word");
  hi = checkUint64Word(hi, "bid128 decode", "hi word");

  const sign = (hi & bid128SignMask) !== 0n;

  // NaN
  if ((hi & bid128NaNMask) === bid128NaNMask) {
    const kind = (hi & bid128SNaNMask) === bid128SNaNMask ? Kind.SNaN : Kind.QNaN;
    // payload: hi[45:0] and lo[63:0] = 110 bits, preserved in full.
    const payHi = hi & 0x00003fffffffffffn;
    const payload = (payHi << 64n) | lo;
    if (payload >= ten33) {
      // non-canonical payload (>= 10^33) normalizes to 0
      return c(sign, kind);
    }
    return c(sign, kind, 0n, 0, payload);
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
  checkSign(comp.sign, "bid128 encode");
  checkKind(comp.kind, "bid128 encode");
  checkKindFields(comp, "bid128 encode");

  let sgn = 0n;
  if (comp.sign) {
    sgn = bid128SignMask;
  }

  switch (comp.kind) {
    case Kind.Infinity:
      return [0n, BigInt.asUintN(64, sgn | 0x7800000000000000n)];
    case Kind.QNaN: {
      const payload = checkEncPayload(comp.payload, bid128EncMaxPayload, "bid128");
      // Validated payload < 10^33, so payHi occupies only the 46-bit payload
      // field (bits 64..109) and never collides with the NaN/sign bits.
      const payHi = payload >> 64n;
      const payLo = payload & 0xffffffffffffffffn;
      return [BigInt.asUintN(64, payLo), BigInt.asUintN(64, sgn | 0x7c00000000000000n | payHi)];
    }
    case Kind.SNaN: {
      const payload = checkEncPayload(comp.payload, bid128EncMaxPayload, "bid128");
      const payHi = payload >> 64n;
      const payLo = payload & 0xffffffffffffffffn;
      return [BigInt.asUintN(64, payLo), BigInt.asUintN(64, sgn | 0x7e00000000000000n | payHi)];
    }
    case Kind.Zero: {
      const exp =
        checkEncExponent(comp.exponent, bid128EncExpMin, bid128EncExpMax, "bid128") + bid128Bias;
      return [0n, BigInt.asUintN(64, sgn | (BigInt(exp) << 49n))];
    }
  }

  // Normal: coefficient as 128 bits. Validated coeff < 10^34, so coeffHi < 2^49
  // and needs no masking to fit under the exponent field.
  const coeff = checkEncCoefficient(comp.coefficient, bid128EncMaxCoeff, "bid128");
  const exp =
    checkEncExponent(comp.exponent, bid128EncExpMin, bid128EncExpMax, "bid128") + bid128Bias;

  const coeffHi = coeff >> 64n;
  const coeffLo = coeff & 0xffffffffffffffffn;
  const lo = BigInt.asUintN(64, coeffLo);
  const hi = BigInt.asUintN(64, sgn | (BigInt(exp) << 49n) | coeffHi);
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
  validateStringComponents(comp);
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

function validateStringComponents(comp: Components): void {
  checkSign(comp.sign, "BID codec string");
  checkKind(comp.kind, "BID codec string");
  if (!Number.isInteger(comp.exponent) || comp.exponent < -2147483648 || comp.exponent > 2147483647) {
    throw new Error(`BID codec string: exponent ${comp.exponent} is not a signed 32-bit integer`);
  }
  if (typeof comp.coefficient !== "bigint") {
    throw new Error(`BID codec string: coefficient must be a bigint, got ${typeof comp.coefficient}`);
  }
  if (typeof comp.payload !== "bigint") {
    throw new Error(`BID codec string: payload must be a bigint, got ${typeof comp.payload}`);
  }
  checkKindFields(comp, "BID codec string");
  if (comp.kind === Kind.Normal) {
    if (comp.coefficient < 0n) {
      throw new Error(`BID codec string: coefficient ${comp.coefficient} is negative`);
    }
    if (comp.coefficient > schemaMaxCoefficient) {
      throw new Error(`BID codec string: coefficient ${comp.coefficient} exceeds schema max ${schemaMaxCoefficient}`);
    }
  }
  if (comp.kind === Kind.QNaN || comp.kind === Kind.SNaN) {
    if (comp.payload < 0n) {
      throw new Error(`BID codec string: NaN payload ${comp.payload} is negative`);
    }
    if (comp.payload > bid128EncMaxPayload) {
      throw new Error(`BID codec string: NaN payload ${comp.payload} exceeds schema max ${bid128EncMaxPayload}`);
    }
  }
}

export function fromString(s: string): Components {
  // The whole input must be ASCII, checked before trimming: any code unit above
  // 0x7F anywhere is malformed (Unicode digit variants such as U+FF11/U+0661,
  // Unicode whitespace such as U+00A0, fractions such as U+00BD, etc.).
  // String.prototype.trim() strips Unicode whitespace, which would let a leading
  // U+00A0 pass silently, so it is not used.
  for (let k = 0; k < s.length; k++) {
    if (s.charCodeAt(k) > 0x7f) {
      throw new Error(`non-ASCII character at index ${k}`);
    }
  }
  // Trim only ASCII whitespace {TAB, LF, VT, FF, CR, SPACE}.
  s = trimAsciiWhitespace(s);
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
    const payload = s.length > 4 ? parsePayload(s.slice(4)) : 0n;
    return c(sign, Kind.SNaN, 0n, 0, payload);
  }
  if (upper.startsWith("NAN")) {
    const payload = s.length > 3 ? parsePayload(s.slice(3)) : 0n;
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
    expPart = parseExponentLiteral(s.slice(i));
  }

  if (digits.length === 0) throw new Error("no digits");

  // Remove leading zeros
  let start = 0;
  while (start < digits.length - 1 && digits[start] === "0") {
    start++;
  }
  digits = digits.slice(start);

  const coeff = BigInt(digits);
  // Schema-wide maximum coefficient: the parsed value (after leading-zero
  // removal) must not exceed 10^34-1, the largest coefficient any supported BID
  // width can hold. This is a schema constant, not per-width validation (which
  // stays in encode*): it makes big-integer and fixed-width-integer languages
  // fail the same inputs the same way instead of wrapping or diverging.
  if (coeff > schemaMaxCoefficient) {
    throw new Error(`coefficient ${coeff} exceeds schema max ${schemaMaxCoefficient}`);
  }
  // Only the fraction-adjusted FINAL exponent must fit int32; the literal was
  // allowed past int32 (below the shared 2^53 exact-integer bound) so every
  // toString rendering (adjusted-exponent literal at most int32 max + 33)
  // reparses successfully. Exactness of this fold: |expPart| < 2^53 and
  // |expAdjust| is bounded by the input length (far below 2^30, the engine
  // string-length ceiling), so either the sum stays within the exact double
  // range and is computed exactly, or its true magnitude exceeds 2^53 — a
  // region entirely outside int32, where any rounding (error at most a few
  // ULPs, each >= 2) cannot move the value into the int32 window, so
  // checkedInt32's verdict equals the mathematical one either way.
  const exponent = checkedInt32(expPart + expAdjust, "exponent");

  if (coeff === 0n) {
    return c(sign, Kind.Zero, 0n, exponent);
  }
  return c(sign, Kind.Normal, coeff, exponent);
}

const asciiWhitespace: ReadonlySet<number> = new Set([0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x20]);

function trimAsciiWhitespace(s: string): string {
  let start = 0;
  let end = s.length;
  while (start < end && asciiWhitespace.has(s.charCodeAt(start))) start++;
  while (end > start && asciiWhitespace.has(s.charCodeAt(end - 1))) end--;
  return s.slice(start, end);
}

function parsePayload(s: string): bigint {
  if (!/^[0-9]+$/.test(s)) throw new Error(`invalid NaN payload: ${s}`);
  const payload = BigInt(s);
  // Schema-wide NaN payload limit: reject at or above 10^33 (the widest
  // canonical BID128 NaN payload), the same value encode128 rejects.
  if (payload >= ten33) {
    throw new Error(`NaN payload ${payload} exceeds schema max ${ten33 - 1n}`);
  }
  return payload;
}

// parseExponentLiteral parses an exponent literal: one optional sign then
// ASCII digits, with NO int32 range check — the caller checks the
// fraction-adjusted FINAL exponent against int32, so every toString rendering
// (adjusted-exponent literal at most int32 max + 33) reparses successfully
// (round-trip closure). The literal must be a safe integer (magnitude below
// 2^53): that IS the shared exact-integer literal bound of the fromString
// grammar itself — every language consumer rejects a literal at or beyond
// 2^53 through the same error channel (Number.isSafeInteger pins the bound
// here; the fixed-width and big-integer consumers enforce the same constant
// explicitly), so the seven accepted-input sets are mathematically
// identical, not merely observationally so.
function parseExponentLiteral(s: string): number {
  if (!/^[+-]?[0-9]+$/.test(s)) throw new Error(`invalid exponent: ${s}`);
  const n = Number(s);
  if (!Number.isSafeInteger(n)) {
    throw new Error(`exponent literal at or above the shared exact-integer bound 2^53: ${s}`);
  }
  return n;
}

function checkedInt32(n: number, label: string): number {
  if (!Number.isSafeInteger(n) || n < -2147483648 || n > 2147483647) {
    throw new Error(`${label} out of int32 range: ${n}`);
  }
  return n;
}
