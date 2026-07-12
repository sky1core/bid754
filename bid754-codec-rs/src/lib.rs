//! BID (Binary Integer Decimal) encoding/decoding for IEEE 754 decimal floating-point.
//!
//! Extracts {sign, coefficient, exponent} components from BID32/64/128 encoded values,
//! enabling conversion to any language's native decimal library.

#![cfg_attr(not(feature = "std"), no_std)]

extern crate alloc;

use alloc::string::String;

/// Classifies a decimal value.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Kind {
    Normal,
    Zero,
    Infinity,
    QNaN,
    SNaN,
}

/// Decomposed parts of a BID-encoded decimal.
///
/// ```text
/// value = (-1)^sign * coefficient * 10^exponent
/// ```
///
/// For special values (Infinity, NaN), coefficient is zero and the NaN payload
/// is stored in `payload`. `payload` is the full BID128 110-bit NaN payload as
/// an unsigned integer, subject to the schema-wide value limit of `10^33` (the
/// widest canonical BID128 NaN payload, mirroring the `10^34-1` coefficient
/// cap); BID32/BID64 payloads are unaffected subsets. It is a `u128` (not a
/// `u64`) so that payloads above `2^64` are represented rather than truncated;
/// the unsigned type also makes a negative payload unconstructible, so no
/// negative-payload check is needed.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Components {
    pub sign: bool,
    pub coefficient: u128,
    pub exponent: i32,
    pub kind: Kind,
    pub payload: u128,
}

impl Components {
    fn new_normal(sign: bool, coefficient: u128, exponent: i32) -> Self {
        Self {
            sign,
            coefficient,
            exponent,
            kind: Kind::Normal,
            payload: 0,
        }
    }
    fn new_zero(sign: bool, exponent: i32) -> Self {
        Self {
            sign,
            coefficient: 0,
            exponent,
            kind: Kind::Zero,
            payload: 0,
        }
    }
    fn new_inf(sign: bool) -> Self {
        Self {
            sign,
            coefficient: 0,
            exponent: 0,
            kind: Kind::Infinity,
            payload: 0,
        }
    }
    fn new_nan(sign: bool, kind: Kind, payload: u128) -> Self {
        Self {
            sign,
            coefficient: 0,
            exponent: 0,
            kind,
            payload,
        }
    }
}

fn validate_kind_fields(operation: &str, c: &Components) -> Result<(), String> {
    match c.kind {
        Kind::Normal if c.coefficient == 0 => Err(alloc::format!(
            "{operation}: normal value cannot carry a zero coefficient"
        )),
        Kind::Normal if c.payload != 0 => Err(alloc::format!(
            "{operation}: normal value cannot carry NaN payload {}",
            c.payload
        )),
        Kind::Zero if c.coefficient != 0 => Err(alloc::format!(
            "{operation}: zero value cannot carry coefficient {}",
            c.coefficient
        )),
        Kind::Zero if c.payload != 0 => Err(alloc::format!(
            "{operation}: zero value cannot carry NaN payload {}",
            c.payload
        )),
        Kind::Infinity if c.coefficient != 0 => Err(alloc::format!(
            "{operation}: infinity cannot carry coefficient {}",
            c.coefficient
        )),
        Kind::Infinity if c.exponent != 0 => Err(alloc::format!(
            "{operation}: infinity cannot carry exponent {}",
            c.exponent
        )),
        Kind::Infinity if c.payload != 0 => Err(alloc::format!(
            "{operation}: infinity cannot carry NaN payload {}",
            c.payload
        )),
        Kind::QNaN | Kind::SNaN if c.coefficient != 0 => Err(alloc::format!(
            "{operation}: NaN cannot carry coefficient {}",
            c.coefficient
        )),
        Kind::QNaN | Kind::SNaN if c.exponent != 0 => Err(alloc::format!(
            "{operation}: NaN cannot carry exponent {}",
            c.exponent
        )),
        _ => Ok(()),
    }
}

// --- BID32 constants ---

const BID32_NAN_MASK: u32 = 0x7c000000;
const BID32_SNAN_MASK: u32 = 0x7e000000;
const BID32_INF_MASK: u32 = 0x78000000;
const BID32_SIGN_MASK: u32 = 0x80000000;
const BID32_STEER_MASK: u32 = 0x60000000;
const BID32_EXP_MASK: u32 = 0xff;
const BID32_BIAS: i32 = 101;
const BID32_MAX_COEFF: u128 = 9_999_999; // 10^7-1
const BID32_MIN_EXP: i32 = -101; // unbiased exponent lower bound (biased 0)
const BID32_MAX_EXP: i32 = 90; // unbiased exponent upper bound (biased 191)
const BID32_PAYLOAD_LIMIT: u64 = 1_000_000; // 10^6; canonical NaN payload must be below this

/// Decode a BID32-encoded `u32` into components.
pub fn decode32(v: u32) -> Components {
    let sign = v & BID32_SIGN_MASK != 0;

    // NaN
    if v & BID32_NAN_MASK == BID32_NAN_MASK {
        let kind = if v & BID32_SNAN_MASK == BID32_SNAN_MASK {
            Kind::SNaN
        } else {
            Kind::QNaN
        };
        let mut payload = (v & 0x000fffff) as u128;
        if payload > 999999 {
            payload = 0; // non-canonical
        }
        return Components::new_nan(sign, kind, payload);
    }
    // Infinity
    if v & BID32_INF_MASK == BID32_INF_MASK {
        return Components::new_inf(sign);
    }

    let (exp, coeff);
    if v & BID32_STEER_MASK == BID32_STEER_MASK {
        // special encoding (implicit high bit)
        exp = ((v >> 21) & BID32_EXP_MASK) as i32;
        let c = (v & 0x001fffff) | 0x00800000;
        coeff = if c >= 10000000 { 0 } else { c };
    } else {
        exp = ((v >> 23) & BID32_EXP_MASK) as i32;
        coeff = v & 0x007fffff;
    }

    if coeff == 0 {
        return Components::new_zero(sign, exp - BID32_BIAS);
    }
    Components::new_normal(sign, coeff as u128, exp - BID32_BIAS)
}

/// Validate an unbiased BID32 exponent and return its biased form. Rejects
/// out-of-range exponents instead of clamping them.
fn bid32_biased_exp(exp: i32) -> Result<u32, String> {
    if !(BID32_MIN_EXP..=BID32_MAX_EXP).contains(&exp) {
        return Err(alloc::format!(
            "bid32 encode: exponent {} out of range [{}, {}]",
            exp,
            BID32_MIN_EXP,
            BID32_MAX_EXP
        ));
    }
    Ok((exp + BID32_BIAS) as u32)
}

fn check_bid32_payload(p: u128) -> Result<(), String> {
    if p >= BID32_PAYLOAD_LIMIT as u128 {
        return Err(alloc::format!(
            "bid32 encode: NaN payload {} exceeds max {}",
            p,
            BID32_PAYLOAD_LIMIT - 1
        ));
    }
    Ok(())
}

/// Encode components into a BID32 `u32`.
///
/// This is a validating packing API: it rejects any `Components` field that is
/// not representable in BID32 exactly as supplied, returning `Err` rather than
/// silently truncating, masking, or clamping. In-range values encode unchanged.
///   - coefficient (Normal): 1 ..= 9_999_999
///   - exponent (Zero/Normal): unbiased -101 ..= 90
///   - payload (QNaN/SNaN): 0 ..= 999_999
///
/// The `Components` field types make several invalid inputs unconstructible, so
/// the type system represents the constraint there: `coefficient` is unsigned `u128`
/// (no negatives, nothing above `2^128`), `payload` is unsigned `u128` (no
/// negatives), and `kind` is a closed enum (no unrecognized variant).
pub fn encode32(c: &Components) -> Result<u32, String> {
    validate_kind_fields("bid32 encode", c)?;
    let sgn: u32 = if c.sign { BID32_SIGN_MASK } else { 0 };
    match c.kind {
        Kind::Infinity => Ok(sgn | 0x78000000),
        Kind::QNaN => {
            check_bid32_payload(c.payload)?;
            Ok(sgn | 0x7c000000 | ((c.payload as u32) & 0x000fffff))
        }
        Kind::SNaN => {
            check_bid32_payload(c.payload)?;
            Ok(sgn | 0x7e000000 | ((c.payload as u32) & 0x000fffff))
        }
        Kind::Zero => {
            let exp = bid32_biased_exp(c.exponent)?;
            Ok(sgn | (exp << 23))
        }
        Kind::Normal => {
            if c.coefficient > BID32_MAX_COEFF {
                return Err(alloc::format!(
                    "bid32 encode: coefficient {} exceeds max {}",
                    c.coefficient,
                    BID32_MAX_COEFF
                ));
            }
            let exp = bid32_biased_exp(c.exponent)?;
            let coeff = c.coefficient as u32; // safe: coefficient <= 9_999_999
            if coeff < 0x800000 {
                Ok(sgn | (exp << 23) | coeff)
            } else {
                Ok(sgn | 0x60000000 | (exp << 21) | (coeff & 0x001fffff))
            }
        }
    }
}

// --- BID64 constants ---

const BID64_NAN_MASK: u64 = 0x7c00000000000000;
const BID64_SNAN_MASK: u64 = 0x7e00000000000000;
const BID64_INF_MASK: u64 = 0x7800000000000000;
const BID64_SIGN_MASK: u64 = 0x8000000000000000;
const BID64_STEER_MASK: u64 = 0x6000000000000000;
const BID64_EXP_MASK: u64 = 0x3ff;
const BID64_MAX_COEFF: u64 = 9999999999999999; // 10^16-1
const BID64_BIAS: i32 = 398;
const BID64_MIN_EXP: i32 = -398; // unbiased exponent lower bound (biased 0)
const BID64_MAX_EXP: i32 = 369; // unbiased exponent upper bound (biased 767)
const BID64_PAYLOAD_LIMIT: u64 = 1_000_000_000_000_000; // 10^15; canonical NaN payload must be below this

/// Decode a BID64-encoded `u64` into components.
pub fn decode64(v: u64) -> Components {
    let sign = v & BID64_SIGN_MASK != 0;

    if v & BID64_NAN_MASK == BID64_NAN_MASK {
        let kind = if v & BID64_SNAN_MASK == BID64_SNAN_MASK {
            Kind::SNaN
        } else {
            Kind::QNaN
        };
        let mut payload = (v & 0x0003ffffffffffff) as u128;
        if payload > 999999999999999 {
            payload = 0;
        }
        return Components::new_nan(sign, kind, payload);
    }
    if v & BID64_INF_MASK == BID64_INF_MASK {
        return Components::new_inf(sign);
    }

    let (exp, coeff);
    if v & BID64_STEER_MASK == BID64_STEER_MASK {
        exp = ((v >> 51) & BID64_EXP_MASK) as i32;
        let c = (v & 0x0007ffffffffffff) | 0x0020000000000000;
        coeff = if c > BID64_MAX_COEFF { 0 } else { c };
    } else {
        exp = ((v >> 53) & BID64_EXP_MASK) as i32;
        coeff = v & 0x001fffffffffffff;
    }

    if coeff == 0 {
        return Components::new_zero(sign, exp - BID64_BIAS);
    }
    Components::new_normal(sign, coeff as u128, exp - BID64_BIAS)
}

/// Validate an unbiased BID64 exponent and return its biased form. Rejects
/// out-of-range exponents instead of clamping them.
fn bid64_biased_exp(exp: i32) -> Result<u64, String> {
    if !(BID64_MIN_EXP..=BID64_MAX_EXP).contains(&exp) {
        return Err(alloc::format!(
            "bid64 encode: exponent {} out of range [{}, {}]",
            exp,
            BID64_MIN_EXP,
            BID64_MAX_EXP
        ));
    }
    Ok((exp + BID64_BIAS) as u64)
}

fn check_bid64_payload(p: u128) -> Result<(), String> {
    if p >= BID64_PAYLOAD_LIMIT as u128 {
        return Err(alloc::format!(
            "bid64 encode: NaN payload {} exceeds max {}",
            p,
            BID64_PAYLOAD_LIMIT - 1
        ));
    }
    Ok(())
}

/// Encode components into a BID64 `u64`.
///
/// Validating packing API (see [`encode32`]); rejects fields not representable
/// in BID64 exactly as supplied.
///   - coefficient (Normal): 1 ..= 9_999_999_999_999_999
///   - exponent (Zero/Normal): unbiased -398 ..= 369
///   - payload (QNaN/SNaN): 0 ..= 999_999_999_999_999
pub fn encode64(c: &Components) -> Result<u64, String> {
    validate_kind_fields("bid64 encode", c)?;
    let sgn: u64 = if c.sign { BID64_SIGN_MASK } else { 0 };
    match c.kind {
        Kind::Infinity => Ok(sgn | 0x7800000000000000),
        Kind::QNaN => {
            check_bid64_payload(c.payload)?;
            // check passed: payload < 10^15 < 2^50, so `as u64` keeps every bit.
            Ok(sgn | 0x7c00000000000000 | ((c.payload as u64) & 0x0003ffffffffffff))
        }
        Kind::SNaN => {
            check_bid64_payload(c.payload)?;
            Ok(sgn | 0x7e00000000000000 | ((c.payload as u64) & 0x0003ffffffffffff))
        }
        Kind::Zero => {
            let exp = bid64_biased_exp(c.exponent)?;
            Ok(sgn | (exp << 53))
        }
        Kind::Normal => {
            if c.coefficient > BID64_MAX_COEFF as u128 {
                return Err(alloc::format!(
                    "bid64 encode: coefficient {} exceeds max {}",
                    c.coefficient,
                    BID64_MAX_COEFF
                ));
            }
            let exp = bid64_biased_exp(c.exponent)?;
            let coeff = c.coefficient as u64; // safe: coefficient <= 10^16-1
            if coeff < 0x20000000000000 {
                Ok(sgn | (exp << 53) | coeff)
            } else {
                Ok(sgn | BID64_STEER_MASK | (exp << 51) | (coeff & 0x0007ffffffffffff))
            }
        }
    }
}

// --- BID128 constants ---

const BID128_NAN_MASK: u64 = 0x7c00000000000000;
const BID128_SNAN_MASK: u64 = 0x7e00000000000000;
const BID128_INF_MASK: u64 = 0x7800000000000000;
const BID128_SIGN_MASK: u64 = 0x8000000000000000;
const BID128_STEER_MASK: u64 = 0x6000000000000000;
const BID128_EXP_MASK: u64 = 0x3fff;
const BID128_BIAS: i32 = 6176;
const BID128_MIN_EXP: i32 = -6176; // unbiased exponent lower bound (biased 0)
const BID128_MAX_EXP: i32 = 6111; // unbiased exponent upper bound (biased 12287)

/// 10^34 - max coefficient + 1 for BID128
const TEN34: u128 = 10_000_000_000_000_000_000_000_000_000_000_000;
/// 10^33 - max NaN payload + 1
const TEN33: u128 = 1_000_000_000_000_000_000_000_000_000_000_000;

/// Decode BID128 from (lo, hi) pair into components.
pub fn decode128(lo: u64, hi: u64) -> Components {
    let sign = hi & BID128_SIGN_MASK != 0;

    if hi & BID128_NAN_MASK == BID128_NAN_MASK {
        let kind = if hi & BID128_SNAN_MASK == BID128_SNAN_MASK {
            Kind::SNaN
        } else {
            Kind::QNaN
        };
        // payload: hi[45:0] and lo[63:0] = full 110 bits
        let pay_hi = hi & 0x00003fffffffffff;
        let payload = ((pay_hi as u128) << 64) | (lo as u128);
        if payload >= TEN33 {
            // Non-canonical payload (>= 10^33) normalizes to 0, the same
            // boundary the encode128 reject contract enforces.
            return Components::new_nan(sign, kind, 0);
        }
        return Components::new_nan(sign, kind, payload);
    }
    if hi & BID128_INF_MASK == BID128_INF_MASK {
        return Components::new_inf(sign);
    }

    let (exp, coeff_hi);
    if hi & BID128_STEER_MASK == BID128_STEER_MASK {
        exp = ((hi >> 47) & BID128_EXP_MASK) as i32;
        coeff_hi = (hi & 0x00007fffffffffff) | 0x0020000000000000;
    } else {
        exp = ((hi >> 49) & BID128_EXP_MASK) as i32;
        coeff_hi = hi & 0x0001ffffffffffff;
    }

    let coeff = ((coeff_hi as u128) << 64) | (lo as u128);
    let coeff = if coeff >= TEN34 { 0 } else { coeff };

    if coeff == 0 {
        return Components::new_zero(sign, exp - BID128_BIAS);
    }
    Components::new_normal(sign, coeff, exp - BID128_BIAS)
}

/// Validate an unbiased BID128 exponent and return its biased form. Rejects
/// out-of-range exponents instead of clamping them.
fn bid128_biased_exp(exp: i32) -> Result<u64, String> {
    if !(BID128_MIN_EXP..=BID128_MAX_EXP).contains(&exp) {
        return Err(alloc::format!(
            "bid128 encode: exponent {} out of range [{}, {}]",
            exp,
            BID128_MIN_EXP,
            BID128_MAX_EXP
        ));
    }
    Ok((exp + BID128_BIAS) as u64)
}

/// Reject a NaN payload at or above the canonical BID128 limit (`10^33`, the
/// no-silent-failure boundary). `payload` is `u128`, so a negative value is
/// unconstructible and needs no check.
fn check_bid128_payload(p: u128) -> Result<(), String> {
    if p >= TEN33 {
        return Err(alloc::format!(
            "bid128 encode: NaN payload {} exceeds max {}",
            p,
            TEN33 - 1
        ));
    }
    Ok(())
}

/// Encode components into BID128 as `(lo, hi)`.
///
/// Validating packing API (see [`encode32`]); rejects fields not representable
/// in BID128 exactly as supplied.
///   - coefficient (Normal): 1 ..= 10^34-1
///   - exponent (Zero/Normal): unbiased -6176 ..= 6111
///   - payload (QNaN/SNaN): 0 ..= 10^33-1, split across the (lo, hi) payload
///     words; a payload at or above 10^33 is rejected (the no-silent-failure
///     boundary).
pub fn encode128(c: &Components) -> Result<(u64, u64), String> {
    validate_kind_fields("bid128 encode", c)?;
    let sgn: u64 = if c.sign { BID128_SIGN_MASK } else { 0 };
    match c.kind {
        Kind::Infinity => Ok((0, sgn | 0x7800000000000000)),
        Kind::QNaN => {
            check_bid128_payload(c.payload)?;
            // payload < 10^33 < 2^110, so its high word occupies only the 46-bit
            // payload field (bits 64..109), never the NaN/sign bits.
            Ok((
                c.payload as u64,
                sgn | 0x7c00000000000000 | ((c.payload >> 64) as u64),
            ))
        }
        Kind::SNaN => {
            check_bid128_payload(c.payload)?;
            Ok((
                c.payload as u64,
                sgn | 0x7e00000000000000 | ((c.payload >> 64) as u64),
            ))
        }
        Kind::Zero => {
            let exp = bid128_biased_exp(c.exponent)?;
            Ok((0, sgn | (exp << 49)))
        }
        Kind::Normal => {
            if c.coefficient >= TEN34 {
                return Err(alloc::format!(
                    "bid128 encode: coefficient {} exceeds max {}",
                    c.coefficient,
                    TEN34 - 1
                ));
            }
            let exp = bid128_biased_exp(c.exponent)?;
            let coeff_lo = c.coefficient as u64;
            // coefficient < 10^34 guarantees coeff_hi < 2^49, so it fits the
            // 49-bit coefficient field without colliding with the exponent; no
            // masking (which would silently drop bits) is needed.
            let coeff_hi = (c.coefficient >> 64) as u64;
            let lo = coeff_lo;
            let hi = sgn | (exp << 49) | coeff_hi;
            Ok((lo, hi))
        }
    }
}

// --- Byte-level convenience (little-endian) ---

/// Decode 4 bytes (little-endian) as BID32.
pub fn decode32_bytes(b: &[u8; 4]) -> Components {
    decode32(u32::from_le_bytes(*b))
}

/// Try to decode 4 bytes (little-endian) as BID32.
pub fn try_decode32_bytes(b: &[u8]) -> Result<Components, String> {
    let raw: [u8; 4] = b
        .try_into()
        .map_err(|_| alloc::format!("decode32_bytes: expected 4 bytes, got {}", b.len()))?;
    Ok(decode32_bytes(&raw))
}

/// Decode 8 bytes (little-endian) as BID64.
pub fn decode64_bytes(b: &[u8; 8]) -> Components {
    decode64(u64::from_le_bytes(*b))
}

/// Try to decode 8 bytes (little-endian) as BID64.
pub fn try_decode64_bytes(b: &[u8]) -> Result<Components, String> {
    let raw: [u8; 8] = b
        .try_into()
        .map_err(|_| alloc::format!("decode64_bytes: expected 8 bytes, got {}", b.len()))?;
    Ok(decode64_bytes(&raw))
}

/// Decode 16 bytes (little-endian) as BID128.
pub fn decode128_bytes(b: &[u8; 16]) -> Components {
    let lo = u64::from_le_bytes([b[0], b[1], b[2], b[3], b[4], b[5], b[6], b[7]]);
    let hi = u64::from_le_bytes([b[8], b[9], b[10], b[11], b[12], b[13], b[14], b[15]]);
    decode128(lo, hi)
}

/// Try to decode 16 bytes (little-endian) as BID128.
pub fn try_decode128_bytes(b: &[u8]) -> Result<Components, String> {
    let raw: [u8; 16] = b
        .try_into()
        .map_err(|_| alloc::format!("decode128_bytes: expected 16 bytes, got {}", b.len()))?;
    Ok(decode128_bytes(&raw))
}

/// Encode components as 4 bytes (little-endian) BID32. Rejects invalid
/// `Components` with the same contract as [`encode32`].
pub fn encode32_bytes(c: &Components) -> Result<[u8; 4], String> {
    Ok(encode32(c)?.to_le_bytes())
}

/// Encode components as 8 bytes (little-endian) BID64. Rejects invalid
/// `Components` with the same contract as [`encode64`].
pub fn encode64_bytes(c: &Components) -> Result<[u8; 8], String> {
    Ok(encode64(c)?.to_le_bytes())
}

/// Encode components as 16 bytes (little-endian) BID128. Rejects invalid
/// `Components` with the same contract as [`encode128`].
pub fn encode128_bytes(c: &Components) -> Result<[u8; 16], String> {
    let (lo, hi) = encode128(c)?;
    let mut buf = [0u8; 16];
    buf[..8].copy_from_slice(&lo.to_le_bytes());
    buf[8..].copy_from_slice(&hi.to_le_bytes());
    Ok(buf)
}

// --- String conversion ---

/// Convert valid components to IEEE 754 string representation.
///
/// Examples: `"+1.2345E+2"`, `"-Inf"`, `"+NaN"`
pub fn to_string(c: &Components) -> Result<String, String> {
    validate_string_components(c)?;
    let prefix = if c.sign { "-" } else { "+" };
    Ok(match c.kind {
        Kind::Infinity => {
            let mut s = String::from(prefix);
            s.push_str("Inf");
            s
        }
        Kind::QNaN => {
            let mut s = String::from(prefix);
            s.push_str("NaN");
            if c.payload != 0 {
                use alloc::format;
                s.push_str(&format!("{}", c.payload));
            }
            s
        }
        Kind::SNaN => {
            let mut s = String::from(prefix);
            s.push_str("SNaN");
            if c.payload != 0 {
                use alloc::format;
                s.push_str(&format!("{}", c.payload));
            }
            s
        }
        Kind::Zero => {
            if c.exponent == 0 {
                let mut s = String::from(prefix);
                s.push('0');
                s
            } else {
                use alloc::format;
                format!("{}0E{:+}", prefix, c.exponent)
            }
        }
        Kind::Normal => {
            use alloc::format;
            let digits = format!("{}", c.coefficient);
            let exp = c.exponent as i64 + digits.len() as i64 - 1;
            if digits.len() == 1 {
                format!("{}{}E{:+}", prefix, digits, exp)
            } else {
                format!("{}{}.{}E{:+}", prefix, &digits[..1], &digits[1..], exp)
            }
        }
    })
}

fn validate_string_components(c: &Components) -> Result<(), String> {
    validate_kind_fields("BID codec string", c)?;
    match c.kind {
        Kind::Normal if c.coefficient >= TEN34 => Err(alloc::format!(
            "BID codec string: coefficient {} exceeds schema max {}",
            c.coefficient,
            TEN34 - 1
        )),
        Kind::QNaN | Kind::SNaN if c.payload >= TEN33 => Err(alloc::format!(
            "BID codec string: NaN payload {} exceeds schema max {}",
            c.payload,
            TEN33 - 1
        )),
        _ => Ok(()),
    }
}

/// Parse an IEEE 754 string into components using one strict ASCII grammar
/// (identical across the cross-language codec packages).
///
/// The whole input must be ASCII: any non-ASCII byte anywhere is malformed,
/// including Unicode digit variants and Unicode whitespace. Only ASCII
/// whitespace may surround the token (removed by one surrounding trim); there is
/// no whitespace inside the token. After the optional single leading `+`/`-`,
/// the input is either a special token (`Inf`/`Infinity`, or `NaN`/`SNaN`
/// followed by an optional unsigned ASCII-digit payload whose value must be
/// below the schema-wide NaN payload limit `10^33`, matched case-insensitively)
/// or a number: ASCII digits with at
/// most one `.` and at least one digit, optionally followed by `E`/`e`, one
/// optional sign, and ASCII exponent digits where the exponent literal must be
/// below the shared exact-integer literal bound 2^53 in magnitude (the widest
/// bound every consumer's number type can check exactly; a literal at or
/// beyond it is rejected through the same error channel) and the
/// fraction-adjusted FINAL exponent must fit a signed 32-bit integer — every
/// `to_string` rendering (adjusted exponent up to `i32::MAX + 33`, far below
/// 2^53) reparses through `from_string` (round-trip closure). Underscores and
/// payload-internal signs are malformed everywhere.
///
/// `from_string` validates grammar and schema limits only, identical in all six
/// language packages: the parsed coefficient value must not exceed the
/// schema-wide maximum coefficient `10^34-1` (the largest value any supported
/// BID width can hold — a schema constant, not per-width validation), and the
/// parsed payload value must be below the schema-wide NaN payload limit `10^33`
/// (the widest canonical BID128 NaN payload, the same kind of schema constant).
/// BID width-range validation is the `encode*` contract.
pub fn from_string(s: &str) -> Result<Components, String> {
    // (1) Whole-input ASCII scan, before any trimming, so non-ASCII whitespace
    // and Unicode digit variants are rejected rather than trimmed or parsed.
    if let Some(pos) = s.bytes().position(|b| b > 0x7f) {
        return Err(alloc::format!("non-ASCII byte at position {}", pos));
    }

    // (2) Trim only ASCII whitespace. `str::trim` is avoided: it strips Unicode
    // whitespace, which would let the grammar diverge across languages.
    let s = trim_ascii_space(s);
    if s.is_empty() {
        return Err("empty string".into());
    }

    // (3) Parse. Optional single leading sign.
    let (sign, s) = if let Some(rest) = s.strip_prefix('-') {
        (true, rest)
    } else if let Some(rest) = s.strip_prefix('+') {
        (false, rest)
    } else {
        (false, s)
    };
    if s.is_empty() {
        return Err("no digits".into());
    }

    let upper = s.to_ascii_uppercase();
    if upper == "INF" || upper == "INFINITY" {
        return Ok(Components::new_inf(sign));
    }
    if upper.starts_with("SNAN") {
        let payload = parse_nan_payload(&s[4..])?;
        return Ok(Components::new_nan(sign, Kind::SNaN, payload));
    }
    if upper.starts_with("NAN") {
        let payload = parse_nan_payload(&s[3..])?;
        return Ok(Components::new_nan(sign, Kind::QNaN, payload));
    }

    // Number: ASCII digits, at most one '.', at least one digit.
    let bytes = s.as_bytes();
    let mut coeff: u128 = 0;
    let mut have_digit = false;
    let mut exp_adjust: i64 = 0;
    let mut found_dot = false;
    let mut i = 0;

    while i < bytes.len() && bytes[i] != b'E' && bytes[i] != b'e' {
        let b = bytes[i];
        if b == b'.' {
            if found_dot {
                return Err("multiple decimal points".into());
            }
            found_dot = true;
        } else if b.is_ascii_digit() {
            // Checked accumulation: a value that overflows u128 is necessarily
            // above the 10^34-1 schema cap enforced below, so it fails with the
            // same schema error instead of wrapping.
            coeff = coeff
                .checked_mul(10)
                .and_then(|v| v.checked_add((b - b'0') as u128))
                .ok_or_else(|| alloc::format!("coefficient exceeds schema max {}", TEN34 - 1))?;
            have_digit = true;
            if found_dot {
                // i64 fraction counting, matching the Go consumer's width: the
                // checked_sub can only fail after ~2^63 fraction digits, a
                // region where the true final exponent is far below i32::MIN,
                // so its rejection is the mathematically correct verdict (an
                // i32 counter here would reject constructible ~2^31-digit
                // inputs that Go/Python/Swift accept).
                exp_adjust = exp_adjust
                    .checked_sub(1)
                    .ok_or_else(|| String::from("exponent out of i32 range"))?;
            }
        } else {
            return Err(alloc::format!(
                "unexpected character {:?} in coefficient",
                b as char
            ));
        }
        i += 1;
    }

    if !have_digit {
        return Err("no digits".into());
    }

    // Optional exponent: E/e, one optional sign, at least one ASCII digit.
    let exp_part: i64 = if i < bytes.len() {
        // The loop only stops early on 'E'/'e', so bytes[i] is the marker.
        parse_exponent_literal(&s[i + 1..])?
    } else {
        0
    };

    // Schema-wide coefficient cap: the parsed value (not the digit count) must
    // not exceed 10^34-1, the largest coefficient any supported BID width can
    // hold. This is a shared schema constant, identical in all six language
    // packages, so fixed-width-integer and big-integer languages fail the same
    // inputs the same way. Per-width range validation stays in encode*.
    if coeff >= TEN34 {
        return Err(alloc::format!(
            "coefficient {} exceeds schema max {}",
            coeff,
            TEN34 - 1
        ));
    }

    // Only the fraction-adjusted FINAL exponent must fit i32; the literal was
    // allowed past int32 (below the shared 2^53 bound) so every to_string
    // rendering (adjusted-exponent literal at most i32::MAX + 33) reparses
    // successfully. Exactness of this fold: |exp_part| < 2^53 and exp_adjust
    // is in (i64::MIN, 0], so checked_add can fail only when exp_adjust <
    // i64::MIN + 2^53 (about 2^63 fraction digits), a region where the true
    // final exponent is far below i32::MIN — the checked rejection is the
    // mathematically correct verdict there, so the accepted-input set equals
    // the mathematical rule on every representable input.
    let exponent64 = exp_part
        .checked_add(exp_adjust)
        .ok_or_else(|| String::from("exponent out of i32 range"))?;
    let exponent = i32::try_from(exponent64)
        .map_err(|_| alloc::format!("exponent {} out of i32 range", exponent64))?;

    if coeff == 0 {
        return Ok(Components::new_zero(sign, exponent));
    }
    Ok(Components::new_normal(sign, coeff, exponent))
}

/// Trim only ASCII whitespace (TAB, LF, VT, FF, CR, SPACE) from both ends. This
/// intentionally does not use `str::trim`/`str::trim_ascii`, whose whitespace
/// sets differ (Unicode set, and `is_ascii_whitespace` excludes VT), which would
/// let the grammar diverge across languages.
fn trim_ascii_space(s: &str) -> &str {
    let is_space = |b: u8| matches!(b, 0x09 | 0x0a | 0x0b | 0x0c | 0x0d | 0x20);
    let bytes = s.as_bytes();
    let mut start = 0;
    while start < bytes.len() && is_space(bytes[start]) {
        start += 1;
    }
    let mut end = bytes.len();
    while end > start && is_space(bytes[end - 1]) {
        end -= 1;
    }
    &s[start..end]
}

/// Report whether `s` is non-empty and made solely of ASCII digits 0-9 (no
/// sign, underscore, whitespace, or Unicode digit variants).
fn is_ascii_digits(s: &str) -> bool {
    !s.is_empty() && s.bytes().all(|b| b.is_ascii_digit())
}

/// Parse an optional NaN payload. Empty means a bare NaN/SNaN (payload 0);
/// otherwise it must be unsigned ASCII digits whose value is below the
/// schema-wide NaN payload limit `10^33` (the widest canonical BID128 NaN
/// payload, replacing the former 64-bit fit rule so fixed-width and big-integer
/// languages fail the same inputs the same way). The explicit ASCII pre-check
/// prevents delegating a signed substring to `u128::from_str`, which would
/// otherwise accept a leading `+`.
fn parse_nan_payload(s: &str) -> Result<u128, String> {
    if s.is_empty() {
        return Ok(0);
    }
    if !is_ascii_digits(s) {
        return Err(alloc::format!(
            "invalid NaN payload {:?}: must be unsigned ASCII digits",
            s
        ));
    }
    let v = s
        .parse::<u128>()
        .map_err(|_| alloc::format!("NaN payload {:?} out of u128 range", s))?;
    if v >= TEN33 {
        return Err(alloc::format!(
            "NaN payload {} exceeds schema max {}",
            v,
            TEN33 - 1
        ));
    }
    Ok(v)
}

/// The shared exact-integer exponent-literal bound 2^53: the widest bound
/// every language consumer's number type can check exactly (JavaScript's
/// safe-integer range pins it). A literal at or beyond this magnitude is
/// rejected in every consumer through the same error channel, so every
/// consumer decides each input its runtime can represent by the same
/// mathematical rule (literal below 2^53, fraction-adjusted final exponent in
/// i32) — a fixed-width fraction counter can force a rejection only in
/// regions (over ~2^63 fraction digits) where that rule itself rejects.
const SHARED_EXPONENT_LITERAL_BOUND: i64 = 1 << 53;

/// Parse an exponent literal: an optional single leading sign followed by at
/// least one ASCII digit. Underscores, embedded whitespace, and Unicode digits
/// are malformed, and the literal's magnitude must be below the shared
/// exact-integer bound 2^53 (a literal at or beyond it — including anything
/// past i64 — is rejected through the same error channel). The caller checks
/// the fraction-adjusted FINAL exponent against the signed 32-bit range — the
/// literal itself is allowed past i32 so every to_string rendering
/// (adjusted-exponent literal at most i32::MAX + 33, far below 2^53) reparses
/// successfully. The explicit ASCII pre-check keeps the standard parser from
/// widening the grammar.
fn parse_exponent_literal(s: &str) -> Result<i64, String> {
    let body = s.strip_prefix(['+', '-']).unwrap_or(s);
    if !is_ascii_digits(body) {
        return Err(alloc::format!("invalid exponent {:?}", s));
    }
    match s.parse::<i64>() {
        Ok(v) if v > -SHARED_EXPONENT_LITERAL_BOUND && v < SHARED_EXPONENT_LITERAL_BOUND => Ok(v),
        _ => Err(alloc::format!(
            "exponent literal {:?} at or above the shared exact-integer bound 2^53",
            s
        )),
    }
}

// --- Tests ---

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_decode32_basic() {
        let cases: &[(u32, Components)] = &[
            (0x32800000, Components::new_zero(false, 0)),
            (0xb2800000, Components::new_zero(true, 0)),
            (0x32800001, Components::new_normal(false, 1, 0)),
            (0xb2800001, Components::new_normal(true, 1, 0)),
            (0x78000000, Components::new_inf(false)),
            (0xf8000000, Components::new_inf(true)),
            (0x7c000000, Components::new_nan(false, Kind::QNaN, 0)),
            (0x7e000000, Components::new_nan(false, Kind::SNaN, 0)),
            (0x77f8967f, Components::new_normal(false, 9999999, 90)),
        ];
        for (v, want) in cases {
            let got = decode32(*v);
            assert_eq!(got, *want, "decode32(0x{:08x})", v);
        }
    }

    #[test]
    fn test_roundtrip32() {
        let values: &[u32] = &[
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
        for &v in values {
            let c = decode32(v);
            let got = encode32(&c).unwrap();
            assert_eq!(got, v, "roundtrip32 0x{:08x}: got 0x{:08x}", v, got);
        }
    }

    #[test]
    fn test_decode64_basic() {
        let cases: &[(u64, Kind, i32)] = &[
            (0x31c0000000000000, Kind::Zero, 0),
            (0x31c0000000000001, Kind::Normal, 0),
            (0x7800000000000000, Kind::Infinity, 0),
            (0x7c00000000000000, Kind::QNaN, 0),
        ];
        for (v, kind, exp) in cases {
            let got = decode64(*v);
            assert_eq!(got.kind, *kind, "decode64(0x{:016x}) kind", v);
            assert_eq!(got.exponent, *exp, "decode64(0x{:016x}) exp", v);
        }
    }

    #[test]
    fn test_roundtrip64() {
        let values: &[u64] = &[
            0x31c0000000000000, // +0
            0xb1c0000000000000, // -0
            0x31c0000000000001, // +1
            0x7800000000000000, // +inf
            0x7c00000000000000, // NaN
            0x7e00000000000000, // sNaN
        ];
        for &v in values {
            let c = decode64(v);
            let got = encode64(&c).unwrap();
            assert_eq!(got, v, "roundtrip64 0x{:016x}: got 0x{:016x}", v, got);
        }
    }

    #[test]
    fn test_decode128_basic() {
        let lo: u64 = 0x0000000000000001;
        let hi: u64 = (6176u64) << 49;
        let c = decode128(lo, hi);
        assert_eq!(c.kind, Kind::Normal);
        assert_eq!(c.exponent, 0);
        assert_eq!(c.coefficient, 1);
        assert!(!c.sign);
    }

    #[test]
    fn test_roundtrip128() {
        let cases: &[(u64, u64)] = &[
            (0, (6176u64) << 49),                    // +0
            (0, BID128_SIGN_MASK | (6176u64) << 49), // -0
            (1, (6176u64) << 49),                    // +1
            (0, 0x7800000000000000),                 // +inf
            (0, 0x7c00000000000000),                 // NaN
        ];
        for &(lo, hi) in cases {
            let c = decode128(lo, hi);
            let (got_lo, got_hi) = encode128(&c).unwrap();
            assert_eq!(
                (got_lo, got_hi),
                (lo, hi),
                "roundtrip128 {:016x}_{:016x}: got {:016x}_{:016x}",
                hi,
                lo,
                got_hi,
                got_lo
            );
        }
    }

    #[test]
    fn test_to_string() {
        assert_eq!(to_string(&Components::new_inf(false)).unwrap(), "+Inf");
        assert_eq!(to_string(&Components::new_inf(true)).unwrap(), "-Inf");
        assert_eq!(
            to_string(&Components::new_nan(false, Kind::QNaN, 0)).unwrap(),
            "+NaN"
        );
        assert_eq!(
            to_string(&Components::new_nan(false, Kind::SNaN, 123)).unwrap(),
            "+SNaN123"
        );
        assert_eq!(to_string(&Components::new_zero(false, 0)).unwrap(), "+0");
        assert_eq!(to_string(&Components::new_zero(true, -5)).unwrap(), "-0E-5");
        assert_eq!(
            to_string(&Components::new_normal(false, 12345, -2)).unwrap(),
            "+1.2345E+2"
        );
        assert_eq!(
            to_string(&Components::new_normal(true, 5, 3)).unwrap(),
            "-5E+3"
        );
    }

    #[test]
    fn test_from_string() {
        let cases: &[(&str, Components)] = &[
            ("+Inf", Components::new_inf(false)),
            ("-Inf", Components::new_inf(true)),
            ("Infinity", Components::new_inf(false)),
            ("NaN", Components::new_nan(false, Kind::QNaN, 0)),
            ("SNaN123", Components::new_nan(false, Kind::SNaN, 123)),
            ("-NaN", Components::new_nan(true, Kind::QNaN, 0)),
            ("0", Components::new_zero(false, 0)),
            ("123.45", Components::new_normal(false, 12345, -2)),
            ("+1.23E+5", Components::new_normal(false, 123, 3)),
            ("-100", Components::new_normal(true, 100, 0)),
            ("1E-10", Components::new_normal(false, 1, -10)),
        ];
        for (s, want) in cases {
            let got =
                from_string(s).unwrap_or_else(|e| panic!("from_string({:?}) failed: {}", s, e));
            assert_eq!(got, *want, "from_string({:?})", s);
        }
    }

    #[test]
    fn test_from_string_errors() {
        for input in [
            "",
            "abc",
            "NaNabc",
            "SNaN-1",
            "1.2.3",
            "1E",
            "1Eabc",
            "1E2147483648",
            // Strict-ASCII grammar divergence cases (previously accepted through
            // a lenient stdlib integer parser or Unicode-aware trim).
            "NaN+5",                    // payload leading sign
            "1E\u{FF11}",               // fullwidth Unicode digit in exponent
            "1E1_0",                    // underscore digit group in exponent
            "\u{FF11}\u{FF12}\u{FF13}", // fullwidth Unicode digits in coefficient
            "1E 5",                     // embedded ASCII whitespace in exponent
            "\u{00A0}1",                // leading NBSP (Unicode whitespace) before token
        ] {
            assert!(
                from_string(input).is_err(),
                "from_string({input:?}) succeeded"
            );
        }
    }

    #[test]
    fn test_from_string_valid() {
        for input in ["1.5", "+1.23E+5", "-inf", "NaN123", "1.", ".5", "007"] {
            assert!(from_string(input).is_ok(), "from_string({input:?}) failed");
        }
    }

    /// Pins the schema-wide coefficient cap: the parsed value (leading zeros
    /// removed) must not exceed 10^34-1, the largest coefficient any supported
    /// BID width can hold. Value-based, not digit-count-based, identical in all
    /// six language packages.
    #[test]
    fn test_from_string_schema_coefficient_cap() {
        // 34 nines = 10^34-1, the schema max: accepted with the exact value.
        let c = from_string(&"9".repeat(34)).unwrap();
        assert_eq!(c.kind, Kind::Normal);
        assert_eq!(c.coefficient, TEN34 - 1);
        // 35 nines: value above the schema cap, rejected.
        assert!(from_string(&"9".repeat(35)).is_err());
        // 41 digits but value 1: accepted, because the cap applies to the
        // parsed value, not the digit count.
        let mut s = "0".repeat(40);
        s.push('1');
        let c = from_string(&s).unwrap();
        assert_eq!(c.kind, Kind::Normal);
        assert_eq!(c.coefficient, 1);
    }

    /// Pins the exponent rule: the exponent literal must be below the shared
    /// exact-integer bound 2^53 in magnitude (identical in every language
    /// consumer), and only the fraction-adjusted FINAL exponent must fit a
    /// signed 32-bit integer — so every to_string rendering (adjusted
    /// exponent up to i32::MAX + 33) reparses (round-trip closure).
    #[test]
    fn test_from_string_exponent_i32_bounds() {
        // Literal exceeds i32 but the fraction-adjusted final exponent fits:
        // accepted (the shape to_string renders for near-max exponents).
        let c = from_string("0.001E2147483649").unwrap();
        assert_eq!(c.exponent, 2147483646);
        let c = from_string("1.0E2147483648").unwrap();
        assert_eq!(c.exponent, 2147483647);
        // Literal and fraction-adjusted value both fit: accepted.
        let c = from_string("1.5E2147483647").unwrap();
        assert_eq!(c.exponent, 2147483646);
        // Fraction-adjusted final exponent leaves i32 (one past either edge):
        // rejected.
        assert!(from_string("1.0E-2147483648").is_err());
        assert!(from_string("1.0E+2147483649").is_err());
        // Literal at/beyond the shared 2^53 exact-integer bound: rejected at
        // the literal step (same error channel), at the exact edge in both
        // signs and far past i64.
        assert!(from_string("1E9007199254740992").is_err());
        assert!(from_string("1E-9007199254740992").is_err());
        assert!(from_string(&alloc::format!("1E{}", "9".repeat(25))).is_err());
        // The i64::MIN literal now also fails the 2^53 literal bound; the fold
        // behind it stays checked, so nothing wraps or traps on this shape.
        assert!(from_string("1.0E-9223372036854775808").is_err());
        // Round-trip closure at the rendered edge: parse(render(x)) succeeds
        // and is a fixed point.
        let first = from_string("10E2147483647").unwrap();
        let rendered = to_string(&first).unwrap();
        assert_eq!(rendered, "+1.0E+2147483648");
        let again = from_string(&rendered).unwrap();
        assert_eq!(to_string(&again).unwrap(), rendered);
    }

    #[test]
    fn test_encode_boundaries() {
        // Coefficient upper bound (Normal): exact max ok, +1 rejected.
        assert!(encode32(&Components::new_normal(false, 9_999_999, 0)).is_ok());
        assert!(encode32(&Components::new_normal(false, 10_000_000, 0)).is_err());
        assert!(encode64(&Components::new_normal(false, 9_999_999_999_999_999, 0)).is_ok());
        assert!(encode64(&Components::new_normal(false, 10_000_000_000_000_000, 0)).is_err());
        assert!(encode128(&Components::new_normal(false, TEN34 - 1, 0)).is_ok());
        assert!(encode128(&Components::new_normal(false, TEN34, 0)).is_err());

        // Exponent range (Normal): exact bounds ok, one past rejected.
        assert!(encode32(&Components::new_normal(false, 1, 90)).is_ok());
        assert!(encode32(&Components::new_normal(false, 1, -101)).is_ok());
        assert!(encode32(&Components::new_normal(false, 1, 91)).is_err());
        assert!(encode32(&Components::new_normal(false, 1, -102)).is_err());
        assert!(encode64(&Components::new_normal(false, 1, 369)).is_ok());
        assert!(encode64(&Components::new_normal(false, 1, -398)).is_ok());
        assert!(encode64(&Components::new_normal(false, 1, 370)).is_err());
        assert!(encode64(&Components::new_normal(false, 1, -399)).is_err());
        assert!(encode128(&Components::new_normal(false, 1, 6111)).is_ok());
        assert!(encode128(&Components::new_normal(false, 1, -6176)).is_ok());
        assert!(encode128(&Components::new_normal(false, 1, 6112)).is_err());
        assert!(encode128(&Components::new_normal(false, 1, -6177)).is_err());

        // Zero exponent range mirrors Normal.
        assert!(encode32(&Components::new_zero(false, 90)).is_ok());
        assert!(encode32(&Components::new_zero(false, 91)).is_err());

        // NaN payload limit, per width (10^6 / 10^15 / 10^33). The bid128
        // payload now uses the full u128, so its 10^33 boundary is exercised.
        assert!(encode32(&Components::new_nan(false, Kind::QNaN, 999_999)).is_ok());
        assert!(encode32(&Components::new_nan(false, Kind::SNaN, 1_000_000)).is_err());
        assert!(encode64(&Components::new_nan(false, Kind::QNaN, 999_999_999_999_999)).is_ok());
        assert!(encode64(&Components::new_nan(
            false,
            Kind::SNaN,
            1_000_000_000_000_000
        ))
        .is_err());
        assert!(encode128(&Components::new_nan(false, Kind::QNaN, TEN33 - 1)).is_ok());
        assert!(encode128(&Components::new_nan(false, Kind::SNaN, TEN33)).is_err());

        // Byte encoders share the reject contract.
        assert!(encode32_bytes(&Components::new_normal(false, 10_000_000, 0)).is_err());
        assert!(encode64_bytes(&Components::new_normal(false, 10_000_000_000_000_000, 0)).is_err());
        assert!(encode128_bytes(&Components::new_normal(false, TEN34, 0)).is_err());
        assert!(encode128_bytes(&Components::new_nan(false, Kind::QNaN, TEN33)).is_err());

        // Field-domain violations that the Go package constructs via big.Int/Kind
        // are unconstructible here; the type system represents that constraint instead:
        //   - a negative coefficient or payload: both are unsigned (`u128`);
        //   - a nil coefficient: `u128` always holds a value (0 for non-Normal);
        //   - an unrecognized `kind`: `Kind` is a closed enum;
        //   - a coefficient at or above 2^128: not representable in `u128`.
        // Those reject records are skipped in Rust with the type system reported
        // as the constraint, matching the cross-language schema note.
    }

    #[test]
    fn test_try_decode_bytes_errors() {
        assert!(try_decode32_bytes(&[0; 3]).is_err());
        assert!(try_decode32_bytes(&[0; 5]).is_err());
        assert!(try_decode64_bytes(&[0; 7]).is_err());
        assert!(try_decode64_bytes(&[0; 9]).is_err());
        assert!(try_decode128_bytes(&[0; 15]).is_err());
        assert!(try_decode128_bytes(&[0; 17]).is_err());
    }

    #[test]
    fn test_string_roundtrip() {
        // Decode BID64 -> to_string -> from_string -> encode BID64
        let values: &[u64] = &[
            0x31c0000000000001, // +1
            0x31c0000000000064, // +100
        ];
        for &v in values {
            let c = decode64(v);
            let s = to_string(&c).unwrap();
            let c2 = from_string(&s).unwrap();
            let v2 = encode64(&c2).unwrap();
            assert_eq!(
                v, v2,
                "string roundtrip 0x{:016x} -> {:?} -> 0x{:016x}",
                v, s, v2
            );
        }
    }

    #[test]
    fn test_bytes_roundtrip32() {
        let v: u32 = 0x32800001; // +1
        let bytes = v.to_le_bytes();
        let c = decode32_bytes(&bytes);
        assert_eq!(c.kind, Kind::Normal);
        assert_eq!(c.coefficient, 1);
        let enc = encode32_bytes(&c).unwrap();
        assert_eq!(enc, bytes);
    }

    #[test]
    fn test_bytes_roundtrip64() {
        let v: u64 = 0x31c0000000000001; // +1
        let bytes = v.to_le_bytes();
        let c = decode64_bytes(&bytes);
        assert_eq!(c.kind, Kind::Normal);
        assert_eq!(c.coefficient, 1);
        let enc = encode64_bytes(&c).unwrap();
        assert_eq!(enc, bytes);
    }

    #[test]
    fn test_bytes_roundtrip128() {
        let lo: u64 = 1;
        let hi: u64 = (6176u64) << 49;
        let mut bytes = [0u8; 16];
        bytes[..8].copy_from_slice(&lo.to_le_bytes());
        bytes[8..].copy_from_slice(&hi.to_le_bytes());
        let c = decode128_bytes(&bytes);
        assert_eq!(c.kind, Kind::Normal);
        assert_eq!(c.coefficient, 1);
        let enc = encode128_bytes(&c).unwrap();
        assert_eq!(enc, bytes);
    }

    /// Build the (lo, hi) BID128 words for a NaN carrying the given payload (up
    /// to 110 bits), signaling flag, and sign. Packs the payload the same way
    /// `encode128` does but without the canonical-limit check, so tests can
    /// construct non-canonical (>= 10^33) inputs directly.
    fn nan_bits128(payload: u128, signaling: bool, sign: bool) -> (u64, u64) {
        let lo = payload as u64;
        let mut hi = ((payload >> 64) as u64) & 0x00003fffffffffff;
        hi |= if signaling {
            0x7e00000000000000
        } else {
            0x7c00000000000000
        };
        if sign {
            hi |= BID128_SIGN_MASK;
        }
        (lo, hi)
    }

    /// Exercises the full 110-bit BID128 NaN payload (values above 2^64 that
    /// populate the high payload word) through the bit round-trip
    /// (Components -> bits -> Components -> bits) and the string round-trip
    /// (Components -> to_string -> from_string -> bits).
    #[test]
    fn test_bid128_full_payload_roundtrip() {
        let payloads: [u128; 2] = [
            18446744073709551616, // 2^64: first value needing the high word
            TEN33 - 1,            // 10^33-1: widest canonical payload
        ];
        for &p in &payloads {
            for kind in [Kind::QNaN, Kind::SNaN] {
                for sign in [false, true] {
                    let inc = Components::new_nan(sign, kind, p);

                    let (lo, hi) = encode128(&inc).expect("encode128");
                    // The high payload word must actually carry bits, proving we
                    // are not silently dropping the payload above 2^64.
                    assert_ne!(
                        hi & 0x00003fffffffffff,
                        0,
                        "payload {} did not populate the high payload word",
                        p
                    );

                    // bits -> Components -> bits
                    let got = decode128(lo, hi);
                    assert_eq!(got.kind, kind);
                    assert_eq!(got.payload, p, "decode128 payload");
                    let (lo2, hi2) = encode128(&got).expect("re-encode128");
                    assert_eq!((lo2, hi2), (lo, hi), "bit round-trip");

                    // Components -> string -> Components -> bits
                    let s = to_string(&inc).unwrap();
                    let parsed =
                        from_string(&s).unwrap_or_else(|e| panic!("from_string({:?}): {}", s, e));
                    let (lo3, hi3) = encode128(&parsed).expect("encode128(from_string)");
                    assert_eq!((lo3, hi3), (lo, hi), "string round-trip via {:?}", s);
                }
            }
        }
    }

    /// Pins the from_string NaN payload boundary at the schema-wide 10^33 limit:
    /// 10^33-1 parses, 10^33 is rejected (for both NaN and SNaN).
    #[test]
    fn test_bid128_payload_boundary_from_string() {
        let c = from_string("NaN999999999999999999999999999999999").expect("10^33-1 payload");
        assert_eq!(c.kind, Kind::QNaN);
        assert_eq!(c.payload, TEN33 - 1);
        assert!(from_string("NaN1000000000000000000000000000000000").is_err());
        assert!(from_string("SNaN1000000000000000000000000000000000").is_err());
    }

    /// Verifies that a BID128 NaN whose encoded payload bits are at or above
    /// 10^33 (non-canonical) decodes to payload 0, the same normalization
    /// boundary the encode reject contract enforces.
    #[test]
    fn test_bid128_noncanonical_payload_decodes_to_zero() {
        let (lo, hi) = nan_bits128(TEN33, false, false);
        let got = decode128(lo, hi);
        assert_eq!(got.kind, Kind::QNaN);
        assert_eq!(got.payload, 0, "payload=10^33 should normalize to 0");

        // All 110 payload bits set (2^110-1), far above 10^33.
        let got = decode128(0xffffffffffffffff, 0x7c003fffffffffff);
        assert_eq!(got.kind, Kind::QNaN);
        assert_eq!(got.payload, 0, "all-payload-bits-set should normalize to 0");
    }
}
