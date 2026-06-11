// Auto-generated from bid128_misc.go by go2rs. Do not edit.

use super::prelude::*;

pub fn bid128_class(mut x: BID_UINT128) -> i64 {
    let mut res: i64 = 0;
    let mut sig_x_prime256: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut sig_x_prime192: BID_UINT192 = BID_UINT192 { w: [0, 0, 0] };
    let mut sig_x: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut exp_x: i64 = 0;
    if ((x.w[1] & 0x7c00000000000000) == 0x7c00000000000000) {
        if ((x.w[1] & 0x7e00000000000000) == 0x7e00000000000000) {
            res = 0;
        } else {
            res = 1;
        }
        return res;
    }
    if ((x.w[1] & 0x7800000000000000) == 0x7800000000000000) {
        if ((x.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = 2;
        } else {
            res = 9;
        }
        return res;
    }
    sig_x.w[1] = (x.w[1] & 0x0001ffffffffffff);
    sig_x.w[0] = x.w[0];
    if ((((sig_x.w[1] > 0x0001ed09bead87c0) || (((sig_x.w[1] == 0x0001ed09bead87c0) && (sig_x.w[0] > 0x378d8e63ffffffff)))) || (((x.w[1] & 0x6000000000000000) == 0x6000000000000000))) || (((sig_x.w[1] == 0) && (sig_x.w[0] == 0)))) {
        if ((x.w[1] & 0x8000000000000000) == 0x8000000000000000) {
            res = 5;
        } else {
            res = 6;
        }
        return res;
    }
    exp_x = ((((go_checked_shr_u64(x.w[1], go_shift_count_u64((49) as u64)))) & 0x000000000003fff) as i64);
    if (exp_x < 33) {
        if (exp_x > 19) {
            sig_x_prime256 = __mul_128x128_to_256(sig_x, bid_ten2k128[(exp_x.wrapping_sub(20)) as usize]);
            if (((sig_x_prime256.w[3] == 0) && (sig_x_prime256.w[2] == 0)) && (((sig_x_prime256.w[1] < 0x0000314dc6448d93) || (((sig_x_prime256.w[1] == 0x0000314dc6448d93) && (sig_x_prime256.w[0] < 0x38c15b0a00000000)))))) {
                if ((x.w[1] & 0x8000000000000000) == 0x8000000000000000) {
                    res = 4;
                } else {
                    res = 7;
                }
                return res;
            }
        } else {
            sig_x_prime192 = __mul_64x128_to_192(bid_ten2k64[exp_x as usize], sig_x);
            if ((sig_x_prime192.w[2] == 0) && (((sig_x_prime192.w[1] < 0x0000314dc6448d93) || (((sig_x_prime192.w[1] == 0x0000314dc6448d93) && (sig_x_prime192.w[0] < 0x38c15b0a00000000)))))) {
                if ((x.w[1] & 0x8000000000000000) == 0x8000000000000000) {
                    res = 4;
                } else {
                    res = 7;
                }
                return res;
            }
        }
    }
    if ((x.w[1] & 0x8000000000000000) == 0x8000000000000000) {
        res = 3;
    } else {
        res = 8;
    }
    return res;
}

pub fn bid128_llquantexp(mut x: BID_UINT128, pfpsf: &mut u32) -> i64 {
    let mut res: i64 = 0;
    if ((x.w[1] & 0x7800000000000000) == 0x7800000000000000) {
        (*pfpsf) |= 1;
        res = ((-1) << 63);
        return res;
    }
    if ((x.w[1] & 0x6000000000000000) == 0x6000000000000000) {
        res = (((((go_checked_shr_u64(x.w[1], go_shift_count_u64((47) as u64)))) & 0x3fff) as i64).wrapping_sub(6176));
    } else {
        res = (((((go_checked_shr_u64(x.w[1], go_shift_count_u64((49) as u64)))) & 0x3fff) as i64).wrapping_sub(6176));
    }
    return res;
}

pub fn bid128_quantexp(mut x: BID_UINT128, pfpsf: &mut u32) -> i32 {
    if ((x.w[1] & 0x7800000000000000) == 0x7800000000000000) {
        (*pfpsf) |= 1;
        return (-0x80000000);
    }
    if ((x.w[1] & 0x6000000000000000) == 0x6000000000000000) {
        return (((((go_checked_shr_u64(x.w[1], go_shift_count_u64((47) as u64)))) & 0x3fff) as i32).wrapping_sub(6176));
    }
    return (((((go_checked_shr_u64(x.w[1], go_shift_count_u64((49) as u64)))) & 0x3fff) as i32).wrapping_sub(6176));
}

pub fn bid128_quantum(mut x: BID_UINT128) -> BID_UINT128 {
    let mut res: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut int_exp: i64 = 0;
    if ((x.w[1] & 0x7c00000000000000) == 0x7800000000000000) {
        res.w[1] = 0x7800000000000000;
        res.w[0] = 0x0000000000000000;
        return res;
    } else if ((x.w[1] & 0x7c00000000000000) == 0x7c00000000000000) {
        res.w[1] = (x.w[1] & 0xfdffffffffffffff);
        res.w[0] = x.w[0];
        return res;
    }
    if ((x.w[1] & 0x6000000000000000) == 0x6000000000000000) {
        int_exp = (((((go_checked_shr_u64(x.w[1], go_shift_count_u64((47) as u64)))) & 0x3fff) as i64).wrapping_sub(6176));
    } else {
        int_exp = (((((go_checked_shr_u64(x.w[1], go_shift_count_u64((49) as u64)))) & 0x3fff) as i64).wrapping_sub(6176));
    }
    res.w[1] = (((go_checked_shl_u64(((int_exp as i64) as u64), go_shift_count_u64((49) as u64)))).wrapping_add(0x3040000000000000));
    res.w[0] = 0x0000000000000001;
    return res;
}

pub fn bid128_scalbn(mut x: BID_UINT128, mut n: i64, mut rnd_mode: i64, pfpsf: &mut u32) -> BID_UINT128 {
    let mut CX: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut CX2: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut CBID_X8: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut res: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut exp64: i64 = 0;
    let mut sign_x: u64 = 0;
    let mut exponent_x: i64 = 0;
    let (mut sign_x, mut exponent_x, mut CX, mut valid) = unpack_bid128_value(x);
    if (!valid) {
        if ((x.w[1] & 0x7e00000000000000) == 0x7e00000000000000) {
            (*pfpsf) |= 1;
        }
        res.w[1] = (CX.w[1] & 0xfdffffffffffffff);
        res.w[0] = CX.w[0];
        if ((CX.w[1] == 0) && (CX.w[0] == 0)) {
            exp64 = ((exponent_x as i64).wrapping_add(n as i64));
            if (exp64 < 0) {
                exp64 = 0;
            }
            if (exp64 > 0x2fff) {
                exp64 = 0x2fff;
            }
            exponent_x = (exp64 as i64);
            res = very_fast_get_bid128(sign_x, exponent_x, CX);
        }
        return res;
    }
    exp64 = ((exponent_x as i64).wrapping_add(n as i64));
    exponent_x = (exp64 as i64);
    if ((exponent_x as u32) <= (0x2fff as u32)) {
        res = very_fast_get_bid128(sign_x, exponent_x, CX);
        return res;
    }
    if (exp64 > (0x2fff as i64)) {
        if (CX.w[1] < 0x314dc6448d93) {
            loop {
                CBID_X8.w[1] = (((go_checked_shl_u64(CX.w[1], go_shift_count_u64((3) as u64)))) | ((go_checked_shr_u64(CX.w[0], go_shift_count_u64((61) as u64)))));
                CBID_X8.w[0] = (go_checked_shl_u64(CX.w[0], go_shift_count_u64((3) as u64)));
                CX2.w[1] = (((go_checked_shl_u64(CX.w[1], go_shift_count_u64((1) as u64)))) | ((go_checked_shr_u64(CX.w[0], go_shift_count_u64((63) as u64)))));
                CX2.w[0] = (go_checked_shl_u64(CX.w[0], go_shift_count_u64((1) as u64)));
                CX = __add_128_128(CX2, CBID_X8);
                exponent_x = exponent_x.wrapping_sub(1);
                exp64 = exp64.wrapping_sub(1);
                if (!(((CX.w[1] < 0x314dc6448d93) && (exp64 > (0x2fff as i64))))) {
                    break;
                }
            }
        }
        if (exp64 <= (0x2fff as i64)) {
            res = very_fast_get_bid128(sign_x, exponent_x, CX);
            return res;
        } else {
            exponent_x = 0x7fffffff;
        }
    }
    res = bid_get_bid128(sign_x, exponent_x, CX, rnd_mode, pfpsf);
    return res;
}

pub fn bid128_scalbln(mut x: BID_UINT128, mut n: i64, mut rnd_mode: i64, pfpsf: &mut u32) -> BID_UINT128 {
    let mut n1 = (n as i32);
    n1 = if ((i64::from(n1)) < n) {
        i32::MAX
    } else if ((i64::from(n1)) > n) {
        i32::MIN
    } else {
        n1
    };
    return bid128_scalbn(x, (n1 as i64), rnd_mode, pfpsf);
}

pub fn bid128_ilogb(mut x: BID_UINT128, pfpsf: &mut u32) -> i64 {
    let mut CX: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut D: i64 = 0;
    let mut exponent_x: i64 = 0;
    let mut bin_expon_cx: i64 = 0;
    let mut digits: i64 = 0;
    let mut res: i64 = 0;
    let (_, mut exponent_x, mut CX, mut valid) = unpack_bid128_value(x);
    if (!valid) {
        (*pfpsf) |= 1;
        if ((x.w[1] & 0x7c00000000000000) == 0x7800000000000000) {
            res = 0x7fffffff;
        } else {
            res = ((((-1) << 31) as i32) as i64);
        }
        return res;
    }
    let mut f64_i: u32 = (0x5f800000 as u32);
    let mut fx_d = no_fma_mul_add_f32((CX.w[1] as f32), f32::from_bits(f64_i), (CX.w[0] as f32));
    let mut fx_i = (fx_d as f32).to_bits();
    bin_expon_cx = (((((go_checked_shr_u32(fx_i, go_shift_count_u64((23) as u64)))) & 0xff) as i64).wrapping_sub(0x7f));
    digits = (bid_estimate_decimal_digits[bin_expon_cx as usize] as i64);
    D = ((CX.w[1] as i64).wrapping_sub(bid_power10_index_binexp_128[bin_expon_cx as usize].w[1] as i64));
    if ((D > 0) || (((D == 0) && (CX.w[0] >= bid_power10_index_binexp_128[bin_expon_cx as usize].w[0])))) {
        digits = digits.wrapping_add(1);
    }
    exponent_x = (((exponent_x.wrapping_sub(0x1820)).wrapping_sub(1)).wrapping_add(digits));
    return exponent_x;
}

pub fn bid128_logb(mut x: BID_UINT128, pfpsf: &mut u32) -> BID_UINT128 {
    let mut ires: i64 = 0;
    let mut res: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut CX: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let (_, _, mut CX, mut valid) = unpack_bid128_value(x);
    if (!valid) {
        if ((x.w[1] & 0x7800000000000000) == 0x7800000000000000) {
            if ((x.w[1] & 0x7e00000000000000) == 0x7e00000000000000) {
                (*pfpsf) |= 1;
            }
            res.w[1] = (CX.w[1] & 0xfdffffffffffffff);
            res.w[0] = CX.w[0];
            if ((x.w[1] & 0x7c00000000000000) == 0x7800000000000000) {
                res.w[1] = (res.w[1] & 0x7fffffffffffffff);
            }
            return res;
        }
        (*pfpsf) |= 4;
        res.w[1] = 0xf800000000000000;
        res.w[0] = 0;
        return res;
    }
    ires = bid128_ilogb(x, pfpsf);
    if ((ires & 0x80000000) != 0) {
        res.w[1] = 0xb040000000000000;
        res.w[0] = ((((ires as i32) as i64).wrapping_neg()) as u64);
    } else {
        res.w[1] = 0x3040000000000000;
        res.w[0] = (ires as u64);
    }
    return res;
}

pub fn bid128_fdim(mut x: BID_UINT128, mut y: BID_UINT128, mut rnd_mode: i64, pfpsf: &mut u32) -> BID_UINT128 {
    let mut res: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut tmp_fpsf: u32 = 0;
    tmp_fpsf = (*pfpsf);
    let (mut cmpres, _) = bid128_quiet_greater(x, y);
    (*pfpsf) = tmp_fpsf;
    if (((((x.w[1] & 0x7c00000000000000) != 0x7c00000000000000)) && (((y.w[1] & 0x7c00000000000000) != 0x7c00000000000000))) && (cmpres == 0)) {
        res.w[1] = 0x3040000000000000;
        res.w[0] = 0x0000000000000000;
        return res;
    }
    res = bid128_sub(x, y, rnd_mode, pfpsf);
    return res;
}
