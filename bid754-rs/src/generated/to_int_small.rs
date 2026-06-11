// Auto-generated from to_int_small.go by go2rs. Do not edit.

use super::prelude::*;

pub fn bid64_to_int8_rnint(mut x: u64) -> (i8, u32) {
    let mut res: i32 = 0;
    let mut sgn_mask: i32 = 0;
    let mut pfpsf: u32 = 0;
    let mut saved_fpsc: u32 = 0;
    saved_fpsc = pfpsf;
    (res, pfpsf) = bid64_to_int32_rnint(x);
    sgn_mask = (res & -128);
    if ((sgn_mask != 0) && (sgn_mask != -128)) {
        pfpsf = (saved_fpsc | 1);
        res = 128;
    }
    return ((res as i8), pfpsf);
}

pub fn bid64_to_int8_xrnint(mut x: u64) -> (i8, u32) {
    let mut res: i32 = 0;
    let mut sgn_mask: i32 = 0;
    let mut pfpsf: u32 = 0;
    let mut saved_fpsc: u32 = 0;
    saved_fpsc = pfpsf;
    (res, pfpsf) = bid64_to_int32_xrnint(x);
    sgn_mask = (res & -128);
    if ((sgn_mask != 0) && (sgn_mask != -128)) {
        pfpsf = (saved_fpsc | 1);
        res = 128;
    }
    return ((res as i8), pfpsf);
}

pub fn bid64_to_int16_rnint(mut x: u64) -> (i16, u32) {
    let mut res: i32 = 0;
    let mut sgn_mask: i32 = 0;
    let mut pfpsf: u32 = 0;
    let mut saved_fpsc: u32 = 0;
    saved_fpsc = pfpsf;
    (res, pfpsf) = bid64_to_int32_rnint(x);
    sgn_mask = (res & -32768);
    if ((sgn_mask != 0) && (sgn_mask != -32768)) {
        pfpsf = (saved_fpsc | 1);
        res = 0x8000;
    }
    return ((res as i16), pfpsf);
}

pub fn bid64_to_int16_xrnint(mut x: u64) -> (i16, u32) {
    let mut res: i32 = 0;
    let mut sgn_mask: i32 = 0;
    let mut pfpsf: u32 = 0;
    let mut saved_fpsc: u32 = 0;
    saved_fpsc = pfpsf;
    (res, pfpsf) = bid64_to_int32_xrnint(x);
    sgn_mask = (res & -32768);
    if ((sgn_mask != 0) && (sgn_mask != -32768)) {
        pfpsf = (saved_fpsc | 1);
        res = 0x8000;
    }
    return ((res as i16), pfpsf);
}

pub fn bid64_to_int8_rninta(mut x: u64) -> (i8, u32) {
    let mut res: i32 = 0;
    let mut sgn_mask: i32 = 0;
    let mut pfpsf: u32 = 0;
    let mut saved_fpsc: u32 = 0;
    saved_fpsc = pfpsf;
    (res, pfpsf) = bid64_to_int32_rninta(x);
    sgn_mask = (res & -128);
    if ((sgn_mask != 0) && (sgn_mask != -128)) {
        pfpsf = (saved_fpsc | 1);
        res = 128;
    }
    return ((res as i8), pfpsf);
}

pub fn bid64_to_int8_xrninta(mut x: u64) -> (i8, u32) {
    let mut res: i32 = 0;
    let mut sgn_mask: i32 = 0;
    let mut pfpsf: u32 = 0;
    let mut saved_fpsc: u32 = 0;
    saved_fpsc = pfpsf;
    (res, pfpsf) = bid64_to_int32_xrninta(x);
    sgn_mask = (res & -128);
    if ((sgn_mask != 0) && (sgn_mask != -128)) {
        pfpsf = (saved_fpsc | 1);
        res = 128;
    }
    return ((res as i8), pfpsf);
}

pub fn bid64_to_int16_rninta(mut x: u64) -> (i16, u32) {
    let mut res: i32 = 0;
    let mut sgn_mask: i32 = 0;
    let mut pfpsf: u32 = 0;
    let mut saved_fpsc: u32 = 0;
    saved_fpsc = pfpsf;
    (res, pfpsf) = bid64_to_int32_rninta(x);
    sgn_mask = (res & -32768);
    if ((sgn_mask != 0) && (sgn_mask != -32768)) {
        pfpsf = (saved_fpsc | 1);
        res = 0x8000;
    }
    return ((res as i16), pfpsf);
}

pub fn bid64_to_int16_xrninta(mut x: u64) -> (i16, u32) {
    let mut res: i32 = 0;
    let mut sgn_mask: i32 = 0;
    let mut pfpsf: u32 = 0;
    let mut saved_fpsc: u32 = 0;
    saved_fpsc = pfpsf;
    (res, pfpsf) = bid64_to_int32_xrninta(x);
    sgn_mask = (res & -32768);
    if ((sgn_mask != 0) && (sgn_mask != -32768)) {
        pfpsf = (saved_fpsc | 1);
        res = 0x8000;
    }
    return ((res as i16), pfpsf);
}

pub fn bid64_to_int8_int(mut x: u64) -> (i8, u32) {
    let mut res: i32 = 0;
    let mut sgn_mask: i32 = 0;
    let mut pfpsf: u32 = 0;
    let mut saved_fpsc: u32 = 0;
    saved_fpsc = pfpsf;
    (res, pfpsf) = bid64_to_int32_int(x);
    sgn_mask = (res & -128);
    if ((sgn_mask != 0) && (sgn_mask != -128)) {
        pfpsf = (saved_fpsc | 1);
        res = 128;
    }
    return ((res as i8), pfpsf);
}

pub fn bid64_to_int8_xint(mut x: u64) -> (i8, u32) {
    let mut res: i32 = 0;
    let mut sgn_mask: i32 = 0;
    let mut pfpsf: u32 = 0;
    let mut saved_fpsc: u32 = 0;
    saved_fpsc = pfpsf;
    (res, pfpsf) = bid64_to_int32_xint(x);
    sgn_mask = (res & -128);
    if ((sgn_mask != 0) && (sgn_mask != -128)) {
        pfpsf = (saved_fpsc | 1);
        res = 128;
    }
    return ((res as i8), pfpsf);
}

pub fn bid64_to_int16_int(mut x: u64) -> (i16, u32) {
    let mut res: i32 = 0;
    let mut sgn_mask: i32 = 0;
    let mut pfpsf: u32 = 0;
    let mut saved_fpsc: u32 = 0;
    saved_fpsc = pfpsf;
    (res, pfpsf) = bid64_to_int32_int(x);
    sgn_mask = (res & -32768);
    if ((sgn_mask != 0) && (sgn_mask != -32768)) {
        pfpsf = (saved_fpsc | 1);
        res = 0x8000;
    }
    return ((res as i16), pfpsf);
}

pub fn bid64_to_int16_xint(mut x: u64) -> (i16, u32) {
    let mut res: i32 = 0;
    let mut sgn_mask: i32 = 0;
    let mut pfpsf: u32 = 0;
    let mut saved_fpsc: u32 = 0;
    saved_fpsc = pfpsf;
    (res, pfpsf) = bid64_to_int32_xint(x);
    sgn_mask = (res & -32768);
    if ((sgn_mask != 0) && (sgn_mask != -32768)) {
        pfpsf = (saved_fpsc | 1);
        res = 0x8000;
    }
    return ((res as i16), pfpsf);
}

pub fn bid64_to_int8_floor(mut x: u64) -> (i8, u32) {
    let mut res: i32 = 0;
    let mut sgn_mask: i32 = 0;
    let mut pfpsf: u32 = 0;
    let mut saved_fpsc: u32 = 0;
    saved_fpsc = pfpsf;
    (res, pfpsf) = bid64_to_int32_floor(x);
    sgn_mask = (res & -128);
    if ((sgn_mask != 0) && (sgn_mask != -128)) {
        pfpsf = (saved_fpsc | 1);
        res = 128;
    }
    return ((res as i8), pfpsf);
}

pub fn bid64_to_int8_xfloor(mut x: u64) -> (i8, u32) {
    let mut res: i32 = 0;
    let mut sgn_mask: i32 = 0;
    let mut pfpsf: u32 = 0;
    let mut saved_fpsc: u32 = 0;
    saved_fpsc = pfpsf;
    (res, pfpsf) = bid64_to_int32_xfloor(x);
    sgn_mask = (res & -128);
    if ((sgn_mask != 0) && (sgn_mask != -128)) {
        pfpsf = (saved_fpsc | 1);
        res = 128;
    }
    return ((res as i8), pfpsf);
}

pub fn bid64_to_int16_floor(mut x: u64) -> (i16, u32) {
    let mut res: i32 = 0;
    let mut sgn_mask: i32 = 0;
    let mut pfpsf: u32 = 0;
    let mut saved_fpsc: u32 = 0;
    saved_fpsc = pfpsf;
    (res, pfpsf) = bid64_to_int32_floor(x);
    sgn_mask = (res & -32768);
    if ((sgn_mask != 0) && (sgn_mask != -32768)) {
        pfpsf = (saved_fpsc | 1);
        res = 0x8000;
    }
    return ((res as i16), pfpsf);
}

pub fn bid64_to_int16_xfloor(mut x: u64) -> (i16, u32) {
    let mut res: i32 = 0;
    let mut sgn_mask: i32 = 0;
    let mut pfpsf: u32 = 0;
    let mut saved_fpsc: u32 = 0;
    saved_fpsc = pfpsf;
    (res, pfpsf) = bid64_to_int32_xfloor(x);
    sgn_mask = (res & -32768);
    if ((sgn_mask != 0) && (sgn_mask != -32768)) {
        pfpsf = (saved_fpsc | 1);
        res = 0x8000;
    }
    return ((res as i16), pfpsf);
}

pub fn bid64_to_int8_ceil(mut x: u64) -> (i8, u32) {
    let mut res: i32 = 0;
    let mut sgn_mask: i32 = 0;
    let mut pfpsf: u32 = 0;
    let mut saved_fpsc: u32 = 0;
    saved_fpsc = pfpsf;
    (res, pfpsf) = bid64_to_int32_ceil(x);
    sgn_mask = (res & -128);
    if ((sgn_mask != 0) && (sgn_mask != -128)) {
        pfpsf = (saved_fpsc | 1);
        res = 128;
    }
    return ((res as i8), pfpsf);
}

pub fn bid64_to_int8_xceil(mut x: u64) -> (i8, u32) {
    let mut res: i32 = 0;
    let mut sgn_mask: i32 = 0;
    let mut pfpsf: u32 = 0;
    let mut saved_fpsc: u32 = 0;
    saved_fpsc = pfpsf;
    (res, pfpsf) = bid64_to_int32_xceil(x);
    sgn_mask = (res & -128);
    if ((sgn_mask != 0) && (sgn_mask != -128)) {
        pfpsf = (saved_fpsc | 1);
        res = 128;
    }
    return ((res as i8), pfpsf);
}

pub fn bid64_to_int16_ceil(mut x: u64) -> (i16, u32) {
    let mut res: i32 = 0;
    let mut sgn_mask: i32 = 0;
    let mut pfpsf: u32 = 0;
    let mut saved_fpsc: u32 = 0;
    saved_fpsc = pfpsf;
    (res, pfpsf) = bid64_to_int32_ceil(x);
    sgn_mask = (res & -32768);
    if ((sgn_mask != 0) && (sgn_mask != -32768)) {
        pfpsf = (saved_fpsc | 1);
        res = 0x8000;
    }
    return ((res as i16), pfpsf);
}

pub fn bid64_to_int16_xceil(mut x: u64) -> (i16, u32) {
    let mut res: i32 = 0;
    let mut sgn_mask: i32 = 0;
    let mut pfpsf: u32 = 0;
    let mut saved_fpsc: u32 = 0;
    saved_fpsc = pfpsf;
    (res, pfpsf) = bid64_to_int32_xceil(x);
    sgn_mask = (res & -32768);
    if ((sgn_mask != 0) && (sgn_mask != -32768)) {
        pfpsf = (saved_fpsc | 1);
        res = 0x8000;
    }
    return ((res as i16), pfpsf);
}
