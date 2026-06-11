//! BID (Binary Integer Decimal) encoding/decoding for IEEE 754 decimal floating-point.
//!
//! Extracts {sign, coefficient, exponent} components from BID32/64/128 encoded values,
//! enabling conversion to any language's native decimal library.

#![cfg_attr(not(feature = "std"), no_std)]

extern crate alloc;

use alloc::string::String;
use alloc::vec::Vec;
use core::fmt;

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
/// For special values (Infinity, NaN), coefficient is zero and NaN payload
/// is stored in `payload`.
#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Components {
    pub sign: bool,
    pub coefficient: u128,
    pub exponent: i32,
    pub kind: Kind,
    pub payload: u64,
}

impl Components {
    fn new_normal(sign: bool, coefficient: u128, exponent: i32) -> Self {
        Self { sign, coefficient, exponent, kind: Kind::Normal, payload: 0 }
    }
    fn new_zero(sign: bool, exponent: i32) -> Self {
        Self { sign, coefficient: 0, exponent, kind: Kind::Zero, payload: 0 }
    }
    fn new_inf(sign: bool) -> Self {
        Self { sign, coefficient: 0, exponent: 0, kind: Kind::Infinity, payload: 0 }
    }
    fn new_nan(sign: bool, kind: Kind, payload: u64) -> Self {
        Self { sign, coefficient: 0, exponent: 0, kind, payload }
    }
}

impl fmt::Display for Components {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{}", to_string(self))
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

/// Decode a BID32-encoded `u32` into components.
pub fn decode32(v: u32) -> Components {
    let sign = v & BID32_SIGN_MASK != 0;

    // NaN
    if v & BID32_NAN_MASK == BID32_NAN_MASK {
        let kind = if v & BID32_SNAN_MASK == BID32_SNAN_MASK { Kind::SNaN } else { Kind::QNaN };
        let mut payload = (v & 0x000fffff) as u64;
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

/// Encode components into a BID32 `u32`.
///
/// Coefficient must be <= 9_999_999. Exponent range: -101..=90.
pub fn encode32(c: &Components) -> u32 {
    let sgn: u32 = if c.sign { BID32_SIGN_MASK } else { 0 };
    match c.kind {
        Kind::Infinity => sgn | 0x78000000,
        Kind::QNaN => sgn | 0x7c000000 | ((c.payload as u32) & 0x000fffff),
        Kind::SNaN => sgn | 0x7e000000 | ((c.payload as u32) & 0x000fffff),
        Kind::Zero => {
            let mut exp = c.exponent + BID32_BIAS;
            exp = exp.clamp(0, 191);
            sgn | ((exp as u32) << 23)
        }
        Kind::Normal => {
            let coeff = c.coefficient as u32;
            let mut exp = c.exponent + BID32_BIAS;
            exp = exp.clamp(0, 191);
            if coeff < 0x800000 {
                sgn | ((exp as u32) << 23) | coeff
            } else {
                sgn | 0x60000000 | ((exp as u32) << 21) | (coeff & 0x001fffff)
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
const BID64_MAX_COEFF: u64 = 9999999999999999;
const BID64_BIAS: i32 = 398;

/// Decode a BID64-encoded `u64` into components.
pub fn decode64(v: u64) -> Components {
    let sign = v & BID64_SIGN_MASK != 0;

    if v & BID64_NAN_MASK == BID64_NAN_MASK {
        let kind = if v & BID64_SNAN_MASK == BID64_SNAN_MASK { Kind::SNaN } else { Kind::QNaN };
        let mut payload = v & 0x0003ffffffffffff;
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

/// Encode components into a BID64 `u64`.
pub fn encode64(c: &Components) -> u64 {
    let sgn: u64 = if c.sign { BID64_SIGN_MASK } else { 0 };
    match c.kind {
        Kind::Infinity => sgn | 0x7800000000000000,
        Kind::QNaN => sgn | 0x7c00000000000000 | (c.payload & 0x0003ffffffffffff),
        Kind::SNaN => sgn | 0x7e00000000000000 | (c.payload & 0x0003ffffffffffff),
        Kind::Zero => {
            let mut exp = c.exponent + BID64_BIAS;
            exp = exp.clamp(0, 767);
            sgn | ((exp as u64) << 53)
        }
        Kind::Normal => {
            let coeff = c.coefficient as u64;
            let mut exp = c.exponent + BID64_BIAS;
            exp = exp.clamp(0, 767);
            if coeff < 0x20000000000000 {
                sgn | ((exp as u64) << 53) | coeff
            } else {
                sgn | BID64_STEER_MASK | ((exp as u64) << 51) | (coeff & 0x0007ffffffffffff)
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

/// 10^34 - max coefficient + 1 for BID128
const TEN34: u128 = 10_000_000_000_000_000_000_000_000_000_000_000;
/// 10^33 - max NaN payload + 1
const TEN33: u128 = 1_000_000_000_000_000_000_000_000_000_000_000;

/// Decode BID128 from (lo, hi) pair into components.
pub fn decode128(lo: u64, hi: u64) -> Components {
    let sign = hi & BID128_SIGN_MASK != 0;

    if hi & BID128_NAN_MASK == BID128_NAN_MASK {
        let kind = if hi & BID128_SNAN_MASK == BID128_SNAN_MASK { Kind::SNaN } else { Kind::QNaN };
        let pay_hi = hi & 0x00003fffffffffff;
        let coeff = ((pay_hi as u128) << 64) | (lo as u128);
        if coeff >= TEN33 {
            return Components::new_nan(sign, kind, 0);
        }
        // simplified: lo only for payload (matches Go)
        return Components::new_nan(sign, kind, lo);
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

/// Encode components into BID128 as `(lo, hi)`.
pub fn encode128(c: &Components) -> (u64, u64) {
    let sgn: u64 = if c.sign { BID128_SIGN_MASK } else { 0 };
    match c.kind {
        Kind::Infinity => (0, sgn | 0x7800000000000000),
        Kind::QNaN => (c.payload, sgn | 0x7c00000000000000),
        Kind::SNaN => (c.payload, sgn | 0x7e00000000000000),
        Kind::Zero => {
            let mut exp = c.exponent + BID128_BIAS;
            exp = exp.clamp(0, 12287);
            (0, sgn | ((exp as u64) << 49))
        }
        Kind::Normal => {
            let coeff_lo = c.coefficient as u64;
            let coeff_hi = (c.coefficient >> 64) as u64;
            let mut exp = c.exponent + BID128_BIAS;
            exp = exp.clamp(0, 12287);
            let lo = coeff_lo;
            let hi = sgn | ((exp as u64) << 49) | (coeff_hi & 0x0001ffffffffffff);
            (lo, hi)
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
    let raw: [u8; 4] = b.try_into().map_err(|_| alloc::format!("decode32_bytes: expected 4 bytes, got {}", b.len()))?;
    Ok(decode32_bytes(&raw))
}

/// Decode 8 bytes (little-endian) as BID64.
pub fn decode64_bytes(b: &[u8; 8]) -> Components {
    decode64(u64::from_le_bytes(*b))
}

/// Try to decode 8 bytes (little-endian) as BID64.
pub fn try_decode64_bytes(b: &[u8]) -> Result<Components, String> {
    let raw: [u8; 8] = b.try_into().map_err(|_| alloc::format!("decode64_bytes: expected 8 bytes, got {}", b.len()))?;
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
    let raw: [u8; 16] = b.try_into().map_err(|_| alloc::format!("decode128_bytes: expected 16 bytes, got {}", b.len()))?;
    Ok(decode128_bytes(&raw))
}

/// Encode components as 4 bytes (little-endian) BID32.
pub fn encode32_bytes(c: &Components) -> [u8; 4] {
    encode32(c).to_le_bytes()
}

/// Encode components as 8 bytes (little-endian) BID64.
pub fn encode64_bytes(c: &Components) -> [u8; 8] {
    encode64(c).to_le_bytes()
}

/// Encode components as 16 bytes (little-endian) BID128.
pub fn encode128_bytes(c: &Components) -> [u8; 16] {
    let (lo, hi) = encode128(c);
    let mut buf = [0u8; 16];
    buf[..8].copy_from_slice(&lo.to_le_bytes());
    buf[8..].copy_from_slice(&hi.to_le_bytes());
    buf
}

// --- String conversion ---

/// Convert components to IEEE 754 string representation.
///
/// Examples: `"+1.2345E+2"`, `"-Inf"`, `"+NaN"`
pub fn to_string(c: &Components) -> String {
    let prefix = if c.sign { "-" } else { "+" };
    match c.kind {
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
    }
}

/// Parse an IEEE 754 string into components.
///
/// Supports: `"123.45"`, `"+1.23E+5"`, `"-INF"`, `"NaN"`, `"SNaN123"`
pub fn from_string(s: &str) -> Result<Components, String> {
    let s = s.trim();
    if s.is_empty() {
        return Err("empty string".into());
    }

    let (sign, s) = if let Some(rest) = s.strip_prefix('-') {
        (true, rest)
    } else if let Some(rest) = s.strip_prefix('+') {
        (false, rest)
    } else {
        (false, s)
    };

    let upper = s.to_ascii_uppercase();
    if upper == "INF" || upper == "INFINITY" {
        return Ok(Components::new_inf(sign));
    }
    if let Some(rest) = upper.strip_prefix("SNAN") {
        let payload = if rest.is_empty() {
            0
        } else {
            rest.parse::<u64>().map_err(|e| alloc::format!("invalid payload: {}", e))?
        };
        return Ok(Components::new_nan(sign, Kind::SNaN, payload));
    }
    if let Some(rest) = upper.strip_prefix("NAN") {
        let payload = if rest.is_empty() {
            0
        } else {
            rest.parse::<u64>().map_err(|e| alloc::format!("invalid payload: {}", e))?
        };
        return Ok(Components::new_nan(sign, Kind::QNaN, payload));
    }

    // Parse number: digits, decimal point, exponent
    let bytes = s.as_bytes();
    let mut digits: Vec<u8> = Vec::new();
    let mut exp_adjust: i32 = 0;
    let mut found_dot = false;
    let mut i = 0;

    while i < bytes.len() && bytes[i] != b'E' && bytes[i] != b'e' {
        if bytes[i] == b'.' {
            if found_dot {
                return Err("multiple decimal points".into());
            }
            found_dot = true;
        } else if bytes[i].is_ascii_digit() {
            digits.push(bytes[i]);
            if found_dot {
                exp_adjust -= 1;
            }
        } else {
            return Err(alloc::format!("unexpected character: {}", bytes[i] as char));
        }
        i += 1;
    }

    let exp_part: i32 = if i < bytes.len() && (bytes[i] == b'E' || bytes[i] == b'e') {
        i += 1;
        let exp_str = &s[i..];
        exp_str.parse::<i32>().map_err(|e| alloc::format!("invalid exponent: {}", e))?
    } else {
        0
    };

    if digits.is_empty() {
        return Err("no digits".into());
    }

    // Remove leading zeros (keep at least one digit)
    let mut start = 0;
    while start < digits.len() - 1 && digits[start] == b'0' {
        start += 1;
    }
    let digits = &digits[start..];

    // Parse coefficient from digit bytes
    let mut coeff: u128 = 0;
    for &d in digits {
        coeff = coeff * 10 + (d - b'0') as u128;
    }

    let exponent = exp_part
        .checked_add(exp_adjust)
        .ok_or_else(|| String::from("exponent out of int32 range"))?;

    if coeff == 0 {
        return Ok(Components::new_zero(sign, exponent));
    }
    Ok(Components::new_normal(sign, coeff, exponent))
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
            let got = encode32(&c);
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
            let got = encode64(&c);
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
            (0, (6176u64) << 49),                           // +0
            (0, BID128_SIGN_MASK | (6176u64) << 49),        // -0
            (1, (6176u64) << 49),                           // +1
            (0, 0x7800000000000000),                        // +inf
            (0, 0x7c00000000000000),                        // NaN
        ];
        for &(lo, hi) in cases {
            let c = decode128(lo, hi);
            let (got_lo, got_hi) = encode128(&c);
            assert_eq!(
                (got_lo, got_hi), (lo, hi),
                "roundtrip128 {:016x}_{:016x}: got {:016x}_{:016x}",
                hi, lo, got_hi, got_lo
            );
        }
    }

    #[test]
    fn test_to_string() {
        assert_eq!(to_string(&Components::new_inf(false)), "+Inf");
        assert_eq!(to_string(&Components::new_inf(true)), "-Inf");
        assert_eq!(to_string(&Components::new_nan(false, Kind::QNaN, 0)), "+NaN");
        assert_eq!(to_string(&Components::new_nan(false, Kind::SNaN, 123)), "+SNaN123");
        assert_eq!(to_string(&Components::new_zero(false, 0)), "+0");
        assert_eq!(to_string(&Components::new_zero(true, -5)), "-0E-5");
        assert_eq!(to_string(&Components::new_normal(false, 12345, -2)), "+1.2345E+2");
        assert_eq!(to_string(&Components::new_normal(true, 5, 3)), "-5E+3");
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
            let got = from_string(s).unwrap_or_else(|e| panic!("from_string({:?}) failed: {}", s, e));
            assert_eq!(got, *want, "from_string({:?})", s);
        }
    }

    #[test]
    fn test_from_string_errors() {
        for input in ["", "abc", "NaNabc", "SNaN-1", "1.2.3", "1E", "1Eabc", "1E2147483648", "1.0E2147483648"] {
            assert!(from_string(input).is_err(), "from_string({input:?}) succeeded");
        }
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
            let s = to_string(&c);
            let c2 = from_string(&s).unwrap();
            let v2 = encode64(&c2);
            assert_eq!(v, v2, "string roundtrip 0x{:016x} -> {:?} -> 0x{:016x}", v, s, v2);
        }
    }

    #[test]
    fn test_display_trait() {
        let c = Components::new_normal(false, 42, 0);
        let s = alloc::format!("{}", c);
        assert_eq!(s, "+4.2E+1");
    }

    #[test]
    fn test_bytes_roundtrip32() {
        let v: u32 = 0x32800001; // +1
        let bytes = v.to_le_bytes();
        let c = decode32_bytes(&bytes);
        assert_eq!(c.kind, Kind::Normal);
        assert_eq!(c.coefficient, 1);
        let enc = encode32_bytes(&c);
        assert_eq!(enc, bytes);
    }

    #[test]
    fn test_bytes_roundtrip64() {
        let v: u64 = 0x31c0000000000001; // +1
        let bytes = v.to_le_bytes();
        let c = decode64_bytes(&bytes);
        assert_eq!(c.kind, Kind::Normal);
        assert_eq!(c.coefficient, 1);
        let enc = encode64_bytes(&c);
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
        let enc = encode128_bytes(&c);
        assert_eq!(enc, bytes);
    }
}
