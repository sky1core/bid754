// Auto-generated from to_uint_small.go by go2rs. Do not edit.

use super::prelude::*;

pub fn bid64_to_uint8_rnint(mut x: u64) -> (u8, u32) {
    let mut res: u32 = 0;
    let mut pfpsf: u32 = 0;
    let mut saved_fpsc: u32 = 0;
    saved_fpsc = pfpsf;
    (res, pfpsf) = bid64_to_uint32_rnint(x);
    if ((res & 0xffffff00) != 0) {
        pfpsf = (saved_fpsc | 1);
        res = 128;
    }
    return ((res as u8), pfpsf);
}

pub fn bid64_to_uint8_xrnint(mut x: u64) -> (u8, u32) {
    let mut res: u32 = 0;
    let mut pfpsf: u32 = 0;
    let mut saved_fpsc: u32 = 0;
    saved_fpsc = pfpsf;
    (res, pfpsf) = bid64_to_uint32_xrnint(x);
    if ((res & 0xffffff00) != 0) {
        pfpsf = (saved_fpsc | 1);
        res = 128;
    }
    return ((res as u8), pfpsf);
}

pub fn bid64_to_uint8_rninta(mut x: u64) -> (u8, u32) {
    let mut res: u32 = 0;
    let mut pfpsf: u32 = 0;
    let mut saved_fpsc: u32 = 0;
    saved_fpsc = pfpsf;
    (res, pfpsf) = bid64_to_uint32_rninta(x);
    if ((res & 0xffffff00) != 0) {
        pfpsf = (saved_fpsc | 1);
        res = 128;
    }
    return ((res as u8), pfpsf);
}

pub fn bid64_to_uint8_xrninta(mut x: u64) -> (u8, u32) {
    let mut res: u32 = 0;
    let mut pfpsf: u32 = 0;
    let mut saved_fpsc: u32 = 0;
    saved_fpsc = pfpsf;
    (res, pfpsf) = bid64_to_uint32_xrninta(x);
    if ((res & 0xffffff00) != 0) {
        pfpsf = (saved_fpsc | 1);
        res = 128;
    }
    return ((res as u8), pfpsf);
}

pub fn bid64_to_uint8_int(mut x: u64) -> (u8, u32) {
    let mut res: u32 = 0;
    let mut pfpsf: u32 = 0;
    let mut saved_fpsc: u32 = 0;
    saved_fpsc = pfpsf;
    (res, pfpsf) = bid64_to_uint32_int(x);
    if ((res & 0xffffff00) != 0) {
        pfpsf = (saved_fpsc | 1);
        res = 128;
    }
    return ((res as u8), pfpsf);
}

pub fn bid64_to_uint8_xint(mut x: u64) -> (u8, u32) {
    let mut res: u32 = 0;
    let mut pfpsf: u32 = 0;
    let mut saved_fpsc: u32 = 0;
    saved_fpsc = pfpsf;
    (res, pfpsf) = bid64_to_uint32_xint(x);
    if ((res & 0xffffff00) != 0) {
        pfpsf = (saved_fpsc | 1);
        res = 128;
    }
    return ((res as u8), pfpsf);
}

pub fn bid64_to_uint8_floor(mut x: u64) -> (u8, u32) {
    let mut res: u32 = 0;
    let mut pfpsf: u32 = 0;
    let mut saved_fpsc: u32 = 0;
    saved_fpsc = pfpsf;
    (res, pfpsf) = bid64_to_uint32_floor(x);
    if ((res & 0xffffff00) != 0) {
        pfpsf = (saved_fpsc | 1);
        res = 128;
    }
    return ((res as u8), pfpsf);
}

pub fn bid64_to_uint8_ceil(mut x: u64) -> (u8, u32) {
    let mut res: u32 = 0;
    let mut pfpsf: u32 = 0;
    let mut saved_fpsc: u32 = 0;
    saved_fpsc = pfpsf;
    (res, pfpsf) = bid64_to_uint32_ceil(x);
    if ((res & 0xffffff00) != 0) {
        pfpsf = (saved_fpsc | 1);
        res = 128;
    }
    return ((res as u8), pfpsf);
}

pub fn bid64_to_uint8_xfloor(mut x: u64) -> (u8, u32) {
    let mut res: u32 = 0;
    let mut pfpsf: u32 = 0;
    let mut saved_fpsc: u32 = 0;
    saved_fpsc = pfpsf;
    (res, pfpsf) = bid64_to_uint32_xfloor(x);
    if ((res & 0xffffff00) != 0) {
        pfpsf = (saved_fpsc | 1);
        res = 128;
    }
    return ((res as u8), pfpsf);
}

pub fn bid64_to_uint8_xceil(mut x: u64) -> (u8, u32) {
    let mut res: u32 = 0;
    let mut pfpsf: u32 = 0;
    let mut saved_fpsc: u32 = 0;
    saved_fpsc = pfpsf;
    (res, pfpsf) = bid64_to_uint32_xceil(x);
    if ((res & 0xffffff00) != 0) {
        pfpsf = (saved_fpsc | 1);
        res = 128;
    }
    return ((res as u8), pfpsf);
}

pub fn bid64_to_uint16_rnint(mut x: u64) -> (u16, u32) {
    let mut res: u32 = 0;
    let mut pfpsf: u32 = 0;
    let mut saved_fpsc: u32 = 0;
    saved_fpsc = pfpsf;
    (res, pfpsf) = bid64_to_uint32_rnint(x);
    if ((res & 0xffff0000) != 0) {
        pfpsf = (saved_fpsc | 1);
        res = 0x8000;
    }
    return ((res as u16), pfpsf);
}

pub fn bid64_to_uint16_xrnint(mut x: u64) -> (u16, u32) {
    let mut res: u32 = 0;
    let mut pfpsf: u32 = 0;
    let mut saved_fpsc: u32 = 0;
    saved_fpsc = pfpsf;
    (res, pfpsf) = bid64_to_uint32_xrnint(x);
    if ((res & 0xffff0000) != 0) {
        pfpsf = (saved_fpsc | 1);
        res = 0x8000;
    }
    return ((res as u16), pfpsf);
}

pub fn bid64_to_uint16_rninta(mut x: u64) -> (u16, u32) {
    let mut res: u32 = 0;
    let mut pfpsf: u32 = 0;
    let mut saved_fpsc: u32 = 0;
    saved_fpsc = pfpsf;
    (res, pfpsf) = bid64_to_uint32_rninta(x);
    if ((res & 0xffff0000) != 0) {
        pfpsf = (saved_fpsc | 1);
        res = 0x8000;
    }
    return ((res as u16), pfpsf);
}

pub fn bid64_to_uint16_xrninta(mut x: u64) -> (u16, u32) {
    let mut res: u32 = 0;
    let mut pfpsf: u32 = 0;
    let mut saved_fpsc: u32 = 0;
    saved_fpsc = pfpsf;
    (res, pfpsf) = bid64_to_uint32_xrninta(x);
    if ((res & 0xffff0000) != 0) {
        pfpsf = (saved_fpsc | 1);
        res = 0x8000;
    }
    return ((res as u16), pfpsf);
}

pub fn bid64_to_uint16_int(mut x: u64) -> (u16, u32) {
    let mut res: u32 = 0;
    let mut pfpsf: u32 = 0;
    let mut saved_fpsc: u32 = 0;
    saved_fpsc = pfpsf;
    (res, pfpsf) = bid64_to_uint32_int(x);
    if ((res & 0xffff0000) != 0) {
        pfpsf = (saved_fpsc | 1);
        res = 0x8000;
    }
    return ((res as u16), pfpsf);
}

pub fn bid64_to_uint16_xint(mut x: u64) -> (u16, u32) {
    let mut res: u32 = 0;
    let mut pfpsf: u32 = 0;
    let mut saved_fpsc: u32 = 0;
    saved_fpsc = pfpsf;
    (res, pfpsf) = bid64_to_uint32_xint(x);
    if ((res & 0xffff0000) != 0) {
        pfpsf = (saved_fpsc | 1);
        res = 0x8000;
    }
    return ((res as u16), pfpsf);
}

pub fn bid64_to_uint16_floor(mut x: u64) -> (u16, u32) {
    let mut res: u32 = 0;
    let mut pfpsf: u32 = 0;
    let mut saved_fpsc: u32 = 0;
    saved_fpsc = pfpsf;
    (res, pfpsf) = bid64_to_uint32_floor(x);
    if ((res & 0xffff0000) != 0) {
        pfpsf = (saved_fpsc | 1);
        res = 0x8000;
    }
    return ((res as u16), pfpsf);
}

pub fn bid64_to_uint16_ceil(mut x: u64) -> (u16, u32) {
    let mut res: u32 = 0;
    let mut pfpsf: u32 = 0;
    let mut saved_fpsc: u32 = 0;
    saved_fpsc = pfpsf;
    (res, pfpsf) = bid64_to_uint32_ceil(x);
    if ((res & 0xffff0000) != 0) {
        pfpsf = (saved_fpsc | 1);
        res = 0x8000;
    }
    return ((res as u16), pfpsf);
}

pub fn bid64_to_uint16_xfloor(mut x: u64) -> (u16, u32) {
    let mut res: u32 = 0;
    let mut pfpsf: u32 = 0;
    let mut saved_fpsc: u32 = 0;
    saved_fpsc = pfpsf;
    (res, pfpsf) = bid64_to_uint32_xfloor(x);
    if ((res & 0xffff0000) != 0) {
        pfpsf = (saved_fpsc | 1);
        res = 0x8000;
    }
    return ((res as u16), pfpsf);
}

pub fn bid64_to_uint16_xceil(mut x: u64) -> (u16, u32) {
    let mut res: u32 = 0;
    let mut pfpsf: u32 = 0;
    let mut saved_fpsc: u32 = 0;
    saved_fpsc = pfpsf;
    (res, pfpsf) = bid64_to_uint32_xceil(x);
    if ((res & 0xffff0000) != 0) {
        pfpsf = (saved_fpsc | 1);
        res = 0x8000;
    }
    return ((res as u16), pfpsf);
}
