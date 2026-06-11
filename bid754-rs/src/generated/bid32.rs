// Auto-generated from bid32.go by go2rs. Do not edit.

use super::prelude::*;

pub static errInvalidFormat32: &'static str = "invalid decimal32 format";

pub(crate) fn unpack_bid32(mut x: u32) -> (u32, i64, u32, bool) {
    let mut sign: u32 = 0;
    let mut exponent: i64 = 0;
    let mut coefficient: u32 = 0;
    let mut valid: bool = false;
    sign = (x & 0x80000000);
    if ((x & 0x60000000) == 0x60000000) {
        if ((x & 0x78000000) == 0x78000000) {
            coefficient = (x & 0xfe0fffff);
            if ((x & 0x000fffff) >= 1000000) {
                coefficient = (x & 0xfe000000);
            }
            if ((x & 0x7c000000) == 0x78000000) {
                coefficient = (x & 0xf8000000);
            }
            exponent = 0;
            return (sign, exponent, coefficient, false);
        }
        coefficient = ((x & 0x1fffff) | 0x800000);
        if (coefficient >= 10000000) {
            coefficient = 0;
        }
        exponent = ((((go_checked_shr_u32(x, go_shift_count_u64((21) as u64)))) & 255) as i64);
    } else {
        exponent = ((((go_checked_shr_u32(x, go_shift_count_u64((23) as u64)))) & 255) as i64);
        coefficient = (x & 0x7fffff);
    }
    return (sign, exponent, coefficient, (coefficient != 0));
}

pub(crate) fn encode_bid32(mut sign: i64, mut exponent: i64, mut coefficient: u32) -> u32 {
    let mut sgn: u32 = 0;
    if (sign != 0) {
        sgn = 0x80000000;
    }
    if (exponent < 0) {
        exponent = 0;
    } else if (exponent > 191) {
        exponent = 191;
    }
    if (coefficient >= 0x800000) {
        return (((sgn | 0x60000000) | ((go_checked_shl_u32((exponent as u32), go_shift_count_u64((21) as u64))))) | (coefficient & 0x1fffff));
    }
    return ((sgn | ((go_checked_shl_u32((exponent as u32), go_shift_count_u64((23) as u64))))) | coefficient);
}

pub(crate) fn uitoa32(mut n: u32) -> String {
    if (n == 0) {
        return "0".to_string();
    }
    let mut buf: [u8; 10] = [0; 10];
    let mut i: i64 = 10;
    while (n > 0) {
        i = i.wrapping_sub(1);
        buf[i as usize] = ((b'0' + ((n % 10) as u8)) as u8);
        n /= 10;
    }
    return go_string_from_bytes(&mut buf[i as usize..]);
}

pub fn parse_decimal32_pure(s: impl AsRef<str>) -> (Decimal32Pure, &'static str) {
    let mut s = s.as_ref().to_string();
    return parse_decimal32_pure_with_mode(s, (0 as i32));
}

pub fn parse_decimal32_pure_with_mode(s: impl AsRef<str>, mut mode: i32) -> (Decimal32Pure, &'static str) {
    let mut s = s.as_ref().to_string();
    s = (s).trim().to_string();
    if ((s.len() as i64) == 0) {
        return (0, errInvalidFormat32);
    }
    let mut sign: i64 = 0;
    if (s.as_bytes()[0] == b'-') {
        sign = 1;
        s = (&s[1 as usize..]).to_string();
    } else if (s.as_bytes()[0] == b'+') {
        s = (&s[1 as usize..]).to_string();
    }
    let mut upper = (s).to_ascii_uppercase();
    if ((upper == "INF") || (upper == "INFINITY")) {
        let mut r: u32 = 0x78000000;
        if (sign != 0) {
            r |= 0x80000000;
        }
        return ((r as Decimal32Pure), "");
    }
    if ((upper == "NAN") || (upper == "QNAN")) {
        let mut r: u32 = 0x7c000000;
        if (sign != 0) {
            r |= 0x80000000;
        }
        return ((r as Decimal32Pure), "");
    }
    if (upper == "SNAN") {
        let mut r: u32 = 0x7e000000;
        if (sign != 0) {
            r |= 0x80000000;
        }
        return ((r as Decimal32Pure), "");
    }
    let mut coefficient: u64 = 0;
    let mut exponent: i64 = 0;
    let mut hasDecimal: bool = false;
    let mut decimalPos: i64 = 0;
    let mut digits: i64 = 0;
    let mut i: i64 = 0;
    while (i < (s.len() as i64)) {
        let mut c = s.as_bytes()[i as usize];
        if ((c >= b'0') && (c <= b'9')) {
            if (digits < 16) {
                coefficient = ((coefficient.wrapping_mul(10)).wrapping_add(((c.wrapping_sub(b'0')) as u64)));
                digits = digits.wrapping_add(1);
            } else {
                exponent = exponent.wrapping_add(1);
            }
            if hasDecimal {
                decimalPos = decimalPos.wrapping_add(1);
            }
        } else if (c == b'.') {
            if hasDecimal {
                return (0, errInvalidFormat32);
            }
            hasDecimal = true;
        } else if ((c == b'e') || (c == b'E')) {
            let mut expStr = &s[(i.wrapping_add(1)) as usize..];
            let (mut exp, mut err) = go_atoi(expStr);
            if err.is_some() {
                return (0, errInvalidFormat32);
            }
            exponent = exponent.wrapping_add(exp);
            break;
        } else {
            return (0, errInvalidFormat32);
        }
        i = i.wrapping_add(1);
    }
    if hasDecimal {
        exponent = exponent.wrapping_sub(decimalPos);
    }
    if (coefficient == 0) {
        return ((encode_bid32(sign, 101, 0) as Decimal32Pure), "");
    }
    while (coefficient > (0x98967f as u64)) {
        let mut lastDigit = (coefficient % 10);
        coefficient /= 10;
        exponent = exponent.wrapping_add(1);
        if ((mode == 0) || (mode == 4)) {
            if ((lastDigit > 5) || (((lastDigit == 5) && (mode == 4)))) {
                coefficient = coefficient.wrapping_add(1);
            } else if (((lastDigit == 5) && (mode == 0)) && ((coefficient % 2) == 1)) {
                coefficient = coefficient.wrapping_add(1);
            }
        }
    }
    let mut biasedExp = (exponent.wrapping_add(101));
    if (biasedExp > 191) {
        let mut r: u32 = 0x78000000;
        if (sign != 0) {
            r |= 0x80000000;
        }
        return ((r as Decimal32Pure), "");
    }
    if (biasedExp < 0) {
        while ((biasedExp < 0) && (coefficient > 0)) {
            coefficient /= 10;
            biasedExp = biasedExp.wrapping_add(1);
        }
        if (coefficient == 0) {
            return ((encode_bid32(sign, 0, 0) as Decimal32Pure), "");
        }
    }
    return ((encode_bid32(sign, biasedExp, (coefficient as u32)) as Decimal32Pure), "");
}

pub(crate) fn from_decimal64(mut d: Decimal64Pure, mut mode: i32) -> Decimal32Pure {
    let (mut res, _) = bid64_to_bid32((d as u64), rounding_mode_to_bid(mode as i32));
    return (res as Decimal32Pure);
}
