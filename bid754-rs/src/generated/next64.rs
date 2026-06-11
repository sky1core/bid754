// Auto-generated from next64.go by go2rs. Do not edit.

use super::prelude::*;

pub fn bid64_next_up(mut x: u64) -> (u64, u32) {
    let mut res: u64 = 0;
    let mut x_sign: u64 = 0;
    let mut x_exp: u64 = 0;
    let mut q1: i64 = 0;
    let mut ind: i64 = 0;
    let mut C1: u64 = 0;
    let mut pfpsf: u32 = 0;
    let mut C: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    if ((x & 0x7c00000000000000) == 0x7c00000000000000) {
        if ((x & 0x0003ffffffffffff) > 999999999999999) {
            x = (x & 0xfe00000000000000);
        } else {
            x = (x & 0xfe03ffffffffffff);
        }
        if ((x & 0x7e00000000000000) == 0x7e00000000000000) {
            pfpsf |= 1;
            res = (x & 0xfdffffffffffffff);
        } else {
            res = x;
        }
        return (res, pfpsf);
    } else if ((x & 0x7800000000000000) == 0x7800000000000000) {
        if ((x & 0x8000000000000000) == 0) {
            res = 0x7800000000000000;
        } else {
            res = 0xf7fb86f26fc0ffff;
        }
        return (res, pfpsf);
    }
    x_sign = (x & 0x8000000000000000);
    if ((x & 0x6000000000000000) == 0x6000000000000000) {
        x_exp = (go_checked_shr_u64((x & 0x1ff8000000000000), go_shift_count_u64((51) as u64)));
        C1 = ((x & 0x7ffffffffffff) | 0x20000000000000);
        if (C1 > 9999999999999999) {
            x_exp = 0;
            C1 = 0;
        }
    } else {
        x_exp = (go_checked_shr_u64((x & 0x7fe0000000000000), go_shift_count_u64((53) as u64)));
        C1 = (x & 0x1fffffffffffff);
    }
    if (C1 == 0x0) {
        res = 0x0000000000000001;
    } else {
        if (x == 0x77fb86f26fc0ffff) {
            res = 0x7800000000000000;
        } else if (x == 0x8000000000000001) {
            res = 0x8000000000000000;
        } else {
            C.w[0] = C1;
            C.w[1] = 0;
            q1 = __get_dec_digits64(C);
            if (q1 < 16) {
                if (x_exp > (((16 as i64).wrapping_sub(q1)) as u64)) {
                    ind = ((16 as i64).wrapping_sub(q1));
                    C1 = (C1.wrapping_mul(bid_ten2k64[ind as usize]));
                    x_exp = (x_exp.wrapping_sub(ind as u64));
                } else {
                    ind = (x_exp as i64);
                    C1 = (C1.wrapping_mul(bid_ten2k64[ind as usize]));
                    x_exp = 0;
                }
            }
            if (x_sign == 0) {
                C1 = C1.wrapping_add(1);
                if (C1 == 0x002386f26fc10000) {
                    C1 = 0x00038d7ea4c68000;
                    x_exp = x_exp.wrapping_add(1);
                }
            } else {
                C1 = C1.wrapping_sub(1);
                if ((C1 == 0x00038d7ea4c67fff) && (x_exp != 0)) {
                    C1 = 0x002386f26fc0ffff;
                    x_exp = x_exp.wrapping_sub(1);
                }
            }
            if ((C1 & 0x20000000000000) != 0) {
                res = (((x_sign | ((go_checked_shl_u64(x_exp, go_shift_count_u64((51) as u64))))) | 0x6000000000000000) | (C1 & 0x7ffffffffffff));
            } else {
                res = ((x_sign | ((go_checked_shl_u64(x_exp, go_shift_count_u64((53) as u64))))) | C1);
            }
        }
    }
    return (res, pfpsf);
}

pub fn bid64_next_down(mut x: u64) -> (u64, u32) {
    let mut res: u64 = 0;
    let mut x_sign: u64 = 0;
    let mut x_exp: u64 = 0;
    let mut q1: i64 = 0;
    let mut ind: i64 = 0;
    let mut C1: u64 = 0;
    let mut pfpsf: u32 = 0;
    let mut C: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    if ((x & 0x7c00000000000000) == 0x7c00000000000000) {
        if ((x & 0x0003ffffffffffff) > 999999999999999) {
            x = (x & 0xfe00000000000000);
        } else {
            x = (x & 0xfe03ffffffffffff);
        }
        if ((x & 0x7e00000000000000) == 0x7e00000000000000) {
            pfpsf |= 1;
            res = (x & 0xfdffffffffffffff);
        } else {
            res = x;
        }
        return (res, pfpsf);
    } else if ((x & 0x7800000000000000) == 0x7800000000000000) {
        if ((x & 0x8000000000000000) == 0x8000000000000000) {
            res = 0xf800000000000000;
        } else {
            res = 0x77fb86f26fc0ffff;
        }
        return (res, pfpsf);
    }
    x_sign = (x & 0x8000000000000000);
    if ((x & 0x6000000000000000) == 0x6000000000000000) {
        x_exp = (go_checked_shr_u64((x & 0x1ff8000000000000), go_shift_count_u64((51) as u64)));
        C1 = ((x & 0x7ffffffffffff) | 0x20000000000000);
        if (C1 > 9999999999999999) {
            x_exp = 0;
            C1 = 0;
        }
    } else {
        x_exp = (go_checked_shr_u64((x & 0x7fe0000000000000), go_shift_count_u64((53) as u64)));
        C1 = (x & 0x1fffffffffffff);
    }
    if (C1 == 0x0) {
        res = 0x8000000000000001;
    } else {
        if (x == 0xf7fb86f26fc0ffff) {
            res = 0xf800000000000000;
        } else if (x == 0x0000000000000001) {
            res = 0x0000000000000000;
        } else {
            C.w[0] = C1;
            C.w[1] = 0;
            q1 = __get_dec_digits64(C);
            if (q1 < 16) {
                if (x_exp > (((16 as i64).wrapping_sub(q1)) as u64)) {
                    ind = ((16 as i64).wrapping_sub(q1));
                    C1 = (C1.wrapping_mul(bid_ten2k64[ind as usize]));
                    x_exp = (x_exp.wrapping_sub(ind as u64));
                } else {
                    ind = (x_exp as i64);
                    C1 = (C1.wrapping_mul(bid_ten2k64[ind as usize]));
                    x_exp = 0;
                }
            }
            if (x_sign != 0) {
                C1 = C1.wrapping_add(1);
                if (C1 == 0x002386f26fc10000) {
                    C1 = 0x00038d7ea4c68000;
                    x_exp = x_exp.wrapping_add(1);
                }
            } else {
                C1 = C1.wrapping_sub(1);
                if ((C1 == 0x00038d7ea4c67fff) && (x_exp != 0)) {
                    C1 = 0x002386f26fc0ffff;
                    x_exp = x_exp.wrapping_sub(1);
                }
            }
            if ((C1 & 0x20000000000000) != 0) {
                res = (((x_sign | ((go_checked_shl_u64(x_exp, go_shift_count_u64((51) as u64))))) | 0x6000000000000000) | (C1 & 0x7ffffffffffff));
            } else {
                res = ((x_sign | ((go_checked_shl_u64(x_exp, go_shift_count_u64((53) as u64))))) | C1);
            }
        }
    }
    return (res, pfpsf);
}

pub fn bid64_next_after(mut x: u64, mut y: u64) -> (u64, u32) {
    let mut res: u64 = 0;
    let mut tmp1: u64 = 0;
    let mut tmp2: u64 = 0;
    let mut tmp_fpsf: u32 = 0;
    let mut pfpsf: u32 = 0;
    let mut res1: i64 = 0;
    let mut res2: i64 = 0;
    if ((((x & 0x7800000000000000) == 0x7800000000000000)) || (((y & 0x7800000000000000) == 0x7800000000000000))) {
        if ((x & 0x7c00000000000000) == 0x7c00000000000000) {
            if ((x & 0x0003ffffffffffff) > 999999999999999) {
                x = (x & 0xfe00000000000000);
            } else {
                x = (x & 0xfe03ffffffffffff);
            }
            if ((x & 0x7e00000000000000) == 0x7e00000000000000) {
                pfpsf |= 1;
                res = (x & 0xfdffffffffffffff);
            } else {
                if ((y & 0x7e00000000000000) == 0x7e00000000000000) {
                    pfpsf |= 1;
                }
                res = x;
            }
            return (res, pfpsf);
        } else if ((y & 0x7c00000000000000) == 0x7c00000000000000) {
            if ((y & 0x0003ffffffffffff) > 999999999999999) {
                y = (y & 0xfe00000000000000);
            } else {
                y = (y & 0xfe03ffffffffffff);
            }
            if ((y & 0x7e00000000000000) == 0x7e00000000000000) {
                pfpsf |= 1;
                res = (y & 0xfdffffffffffffff);
            } else {
                res = y;
            }
            return (res, pfpsf);
        } else {
            if ((x & 0x7800000000000000) == 0x7800000000000000) {
                x = (x & (0x8000000000000000 | 0x7800000000000000));
            }
            if ((y & 0x7800000000000000) == 0x7800000000000000) {
                y = (y & (0x8000000000000000 | 0x7800000000000000));
            }
        }
    }
    if ((x & 0x7800000000000000) != 0x7800000000000000) {
        if ((x & 0x6000000000000000) == 0x6000000000000000) {
            if ((((x & 0x7ffffffffffff) | 0x20000000000000)) > 9999999999999999) {
                x = ((x & 0x8000000000000000) | ((go_checked_shl_u64((x & 0x1ff8000000000000), go_shift_count_u64((2) as u64)))));
            }
        } else {
        }
    }
    tmp_fpsf = pfpsf;
    (res1, _) = bid64_quiet_equal(x, y);
    (res2, _) = bid64_quiet_greater(x, y);
    pfpsf = tmp_fpsf;
    if (res1 != 0) {
        res = ((y & 0x8000000000000000) | (x & 0x7fffffffffffffff));
    } else if (res2 != 0) {
        (res, tmp_fpsf) = bid64_next_down(x);
        pfpsf |= tmp_fpsf;
    } else {
        (res, tmp_fpsf) = bid64_next_up(x);
        pfpsf |= tmp_fpsf;
    }
    if ((((x & 0x7800000000000000) != 0x7800000000000000)) && (((res & 0x7800000000000000) == 0x7800000000000000))) {
        pfpsf |= 32;
        pfpsf |= 8;
    }
    tmp1 = 0x00038d7ea4c68000;
    tmp2 = (res & 0x7fffffffffffffff);
    tmp_fpsf = pfpsf;
    (res1, _) = bid64_quiet_greater(tmp1, tmp2);
    (res2, _) = bid64_quiet_not_equal(x, res);
    pfpsf = tmp_fpsf;
    if ((res1 != 0) && (res2 != 0)) {
        pfpsf |= 32;
        pfpsf |= 16;
    }
    return (res, pfpsf);
}
