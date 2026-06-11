// Auto-generated from fdim64.go by go2rs. Do not edit.

use super::prelude::*;

pub fn bid64_fdim(mut x: u64, mut y: u64, mut rndMode: i64) -> (u64, u32) {
    let mut res: u64 = 0;
    let mut cmpres: i64 = 0;
    let mut tmp_pfpsf: u32 = 0;
    tmp_pfpsf = 0;
    _ = tmp_pfpsf;
    (cmpres, _) = bid64_quiet_greater(x, y);
    if (((((x & 0x7c00000000000000) != 0x7c00000000000000)) && (((y & 0x7c00000000000000) != 0x7c00000000000000))) && (cmpres == 0)) {
        res = 0x31c0000000000000;
        return (res, 0);
    }
    let (mut res, mut pfpsf) = bid64_sub_with_flags(x, y, rndMode);
    return (res, pfpsf);
}
