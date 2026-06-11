// Auto-generated from bid128_string.go by go2rs. Do not edit.

use super::prelude::*;

pub(crate) fn l1_split_mi_di_6(mut X: u64, mut MiDi: &mut [u32], ptr: &mut i64) {
    let mut L1_Xhi_64 = (go_checked_shr_u64(((((go_checked_shr_u64(X, go_shift_count_u64((28) as u64)))).wrapping_mul(0x89705f41))), go_shift_count_u64((33) as u64)));
    let mut L1_Xlo_64 = (X.wrapping_sub((L1_Xhi_64.wrapping_mul(0x3b9aca00 as u64))));
    if (L1_Xlo_64 >= (0x3b9aca00 as u64)) {
        L1_Xlo_64 = L1_Xlo_64.wrapping_sub(0x3b9aca00 as u64);
        L1_Xhi_64 = L1_Xhi_64.wrapping_add(1);
    }
    let mut L1_X_hi = (L1_Xhi_64 as u32);
    let mut L1_X_lo = (L1_Xlo_64 as u32);
    l0_split_mi_di_3((L1_X_hi as u64), MiDi, ptr);
    l0_split_mi_di_3((L1_X_lo as u64), MiDi, ptr);
}

pub fn bid128_to_string(mut x: BID_UINT128) -> String {
    let mut str: [u8; 128] = [0; 128];
    let mut k: u64 = 0;
    let mut d0: u64 = 0;
    let mut d123: u64 = 0;
    let mut zero_digit: u64 = (b'0' as u64);
    let mut HI_18Dig: u64 = 0;
    let mut LO_18Dig: u64 = 0;
    let mut Tmp: u64 = 0;
    let mut MiDi: [u32; 12] = [0; 12];
    let mut midi_ind: i64 = 0;
    let mut k_lcv: i64 = 0;
    let mut length: i64 = 0;
    if ((x.w[1] & 0x7800000000000000) == 0x7800000000000000) {
        if ((x.w[1] & 0x7c00000000000000) == 0x7c00000000000000) {
            if ((x.w[1] & 0x7e00000000000000) == 0x7e00000000000000) {
                if ((x.w[1] as i64) < 0) {
                    str[0] = b'-';
                } else {
                    str[0] = b'+';
                }
                str[1] = b'S';
                str[2] = b'N';
                str[3] = b'a';
                str[4] = b'N';
                return go_string_from_bytes(&mut str[..5 as usize]);
            } else {
                if ((x.w[1] as i64) < 0) {
                    str[0] = b'-';
                } else {
                    str[0] = b'+';
                }
                str[1] = b'N';
                str[2] = b'a';
                str[3] = b'N';
                return go_string_from_bytes(&mut str[..4 as usize]);
            }
        } else {
            if ((x.w[1] & 0x8000000000000000) == 0) {
                str[0] = b'+';
                str[1] = b'I';
                str[2] = b'n';
                str[3] = b'f';
                return go_string_from_bytes(&mut str[..4 as usize]);
            } else {
                str[0] = b'-';
                str[1] = b'I';
                str[2] = b'n';
                str[3] = b'f';
                return go_string_from_bytes(&mut str[..4 as usize]);
            }
        }
    } else if ((((x.w[1] & 0x1ffffffffffff) == 0)) && (x.w[0] == 0)) {
        length = 0;
        if ((x.w[1] & 0x8000000000000000) != 0) {
            str[length as usize] = b'-';
        } else {
            str[length as usize] = b'+';
        }
        length = length.wrapping_add(1);
        str[length as usize] = b'0';
        length = length.wrapping_add(1);
        str[length as usize] = b'E';
        length = length.wrapping_add(1);
        let mut exp = (((go_checked_shr_u64((x.w[1] & 0x7ffe000000000000), go_shift_count_u64((49) as u64))) as i64).wrapping_sub(6176));
        if (exp > (((0x5ffe >> 1) - (6176)))) {
            exp = (((go_checked_shr_u64(((((go_checked_shl_u64(x.w[1], go_shift_count_u64((2) as u64)))) & 0x7ffe000000000000)), go_shift_count_u64((49) as u64))) as i64).wrapping_sub(6176));
        }
        if (exp >= 0) {
            str[length as usize] = b'+';
            length = length.wrapping_add(1);
            let mut s = format!("{}", exp);
            go_copy_str(&mut str[length as usize..], &s);
            length = length.wrapping_add((s.len() as i64));
        } else {
            let mut s = format!("{}", exp);
            go_copy_str(&mut str[length as usize..], &s);
            length = length.wrapping_add((s.len() as i64));
        }
        return go_string_from_bytes(&mut str[..length as usize]);
    } else {
        let mut x_sign = (x.w[1] & 0x8000000000000000);
        let mut x_exp = (x.w[1] & 0x7ffe000000000000);
        if ((x.w[1] & 0x6000000000000000) == 0x6000000000000000) {
            x_exp = (((go_checked_shl_u64(x.w[1], go_shift_count_u64((2) as u64)))) & 0x7ffe000000000000);
        }
        let mut C1: BID_UINT128 = BID_UINT128 { w: [0, 0] };
        C1.w[1] = (x.w[1] & 0x1ffffffffffff);
        C1.w[0] = x.w[0];
        let mut exp = (((go_checked_shr_u64(x_exp, go_shift_count_u64((49) as u64))) as i64).wrapping_sub(6176));
        _ = x_sign;
        if (x_sign != 0) {
            str[k as usize] = b'-';
        } else {
            str[k as usize] = b'+';
        }
        k = k.wrapping_add(1);
        if ((((C1.w[1] > 0x0001ed09bead87c0) || (((C1.w[1] == 0x0001ed09bead87c0) && (C1.w[0] > 0x378d8e63ffffffff)))) || (((x.w[1] & 0x6000000000000000) == 0x6000000000000000))) || (((C1.w[1] == 0) && (C1.w[0] == 0)))) {
            str[k as usize] = b'0';
            k = k.wrapping_add(1);
        } else {
            Tmp = (go_checked_shr_u64(C1.w[0], go_shift_count_u64((59) as u64)));
            LO_18Dig = (go_checked_shr_u64(((go_checked_shl_u64(C1.w[0], go_shift_count_u64((5) as u64)))), go_shift_count_u64((5) as u64)));
            Tmp = Tmp.wrapping_add(((go_checked_shl_u64(C1.w[1], go_shift_count_u64((5) as u64)))));
            HI_18Dig = 0;
            k_lcv = 0;
            while (Tmp != 0) {
                midi_ind = ((Tmp & 0x000000000000003F) as i64);
                midi_ind = go_checked_shl_i64(midi_ind, go_shift_count_u64((1) as u64));
                Tmp = go_checked_shr_u64(Tmp, go_shift_count_u64((6) as u64));
                HI_18Dig = HI_18Dig.wrapping_add(mod10_18_tbl[k_lcv as usize][midi_ind as usize]);
                midi_ind = midi_ind.wrapping_add(1);
                LO_18Dig = LO_18Dig.wrapping_add(mod10_18_tbl[k_lcv as usize][midi_ind as usize]);
                k_lcv = k_lcv.wrapping_add(1);
                l0_normalize_10to18((&mut HI_18Dig), (&mut LO_18Dig));
            }
            let mut ptr: i64 = 0;
            if (HI_18Dig == 0) {
                l1_split_mi_di_6_lead(LO_18Dig, &mut MiDi[..], (&mut ptr));
            } else {
                l1_split_mi_di_6_lead(HI_18Dig, &mut MiDi[..], (&mut ptr));
                l1_split_mi_di_6(LO_18Dig, &mut MiDi[..], (&mut ptr));
            }
            length = ptr;
            let mut c_ptr_start = (k as i64);
            let mut c_ptr = c_ptr_start;
            l0_mi_di2_str_lead(MiDi[0], &mut str[..], (&mut c_ptr));
            k_lcv = 1;
            while (k_lcv < length) {
                l0_mi_di2_str(MiDi[k_lcv as usize], &mut str[..], (&mut c_ptr));
                k_lcv = k_lcv.wrapping_add(1);
            }
            k = (k.wrapping_add(((c_ptr.wrapping_sub(c_ptr_start)) as u64)));
        }
        str[k as usize] = b'E';
        k = k.wrapping_add(1);
        if (exp < 0) {
            exp = (exp.wrapping_neg());
            str[k as usize] = b'-';
        } else {
            str[k as usize] = b'+';
        }
        k = k.wrapping_add(1);
        d0 = ((go_checked_shr_u64((((exp as u64).wrapping_mul(0x418a))), go_shift_count_u64((24) as u64))) as u64);
        d123 = ((exp as u64).wrapping_sub(((1000 as u64).wrapping_mul(d0))));
        if (d0 != 0) {
            str[k as usize] = ((d0.wrapping_add(zero_digit)) as u8);
            k = k.wrapping_add(1);
            str[k as usize] = bid_midi_tbl[d123 as usize][0];
            k = k.wrapping_add(1);
            str[k as usize] = bid_midi_tbl[d123 as usize][1];
            k = k.wrapping_add(1);
            str[k as usize] = bid_midi_tbl[d123 as usize][2];
            k = k.wrapping_add(1);
        } else {
            if (d123 < 10) {
                str[k as usize] = ((d123.wrapping_add(zero_digit)) as u8);
                k = k.wrapping_add(1);
            } else if (d123 < 100) {
                str[k as usize] = bid_midi_tbl[d123 as usize][1];
                k = k.wrapping_add(1);
                str[k as usize] = bid_midi_tbl[d123 as usize][2];
                k = k.wrapping_add(1);
            } else {
                str[k as usize] = bid_midi_tbl[d123 as usize][0];
                k = k.wrapping_add(1);
                str[k as usize] = bid_midi_tbl[d123 as usize][1];
                k = k.wrapping_add(1);
                str[k as usize] = bid_midi_tbl[d123 as usize][2];
                k = k.wrapping_add(1);
            }
        }
    }
    return go_string_from_bytes(&mut str[..k as usize]);
}

pub fn bid128_from_string(str: impl AsRef<str>, mut rnd_mode: i64) -> (BID_UINT128, u32) {
    let mut str = str.as_ref().to_string();
    let mut res: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut pfpsf: u32 = 0;
    let mut CX: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut sign_x: u64 = 0;
    let mut coeff_high: u64 = 0;
    let mut coeff_low: u64 = 0;
    let mut coeff2: u64 = 0;
    let mut coeff_l2: u64 = 0;
    let mut carry: u64 = 0;
    let mut scale_high: u64 = 0;
    let mut right_radix_leading_zeros: u64 = 0;
    let mut ndigits_before: i64 = 0;
    let mut ndigits_after: i64 = 0;
    let mut ndigits_total: i64 = 0;
    let mut dec_expon: i64 = 0;
    let mut sgn_exp: i64 = 0;
    let mut i: i64 = 0;
    let mut d2: i64 = 0;
    let mut rdx_pt_enc: i64 = 0;
    let mut set_inexact: i64 = 0;
    let mut min_digits: i64 = 0;
    let mut sticky_bit: i64 = 0;
    let mut buffer: [u8; 100] = [0; 100];
    let mut c: u8 = 0;
    let ps = str.as_bytes();
    let mut ps_idx: usize = 0;
    macro_rules! ps_at {
        ($offset:expr) => {
            *ps.get(ps_idx + ($offset as usize)).unwrap_or(&0)
        };
    }
    right_radix_leading_zeros = 0;
    rdx_pt_enc = 0;
    while ((((ps_at!(0) == b' ') || (ps_at!(0) == b'\t'))) && (ps_at!(0) != 0)) {
        ps_idx += 1;
    }
    c = ps_at!(0);
    if ((c == 0) || (((((c != b'.') && (c != b'-')) && (c != b'+')) && ((((c.wrapping_sub(b'0')) as u64) > 9))))) {
        res.w[0] = 0;
        if (((((tolower_macro(ps_at!(0)) == b'i') && (tolower_macro(ps_at!(1)) == b'n')) && (tolower_macro(ps_at!(2)) == b'f'))) && (((ps_at!(3) == 0) || (((((((tolower_macro(ps_at!(3)) == b'i') && (tolower_macro(ps_at!(4)) == b'n')) && (tolower_macro(ps_at!(5)) == b'i')) && (tolower_macro(ps_at!(6)) == b't')) && (tolower_macro(ps_at!(7)) == b'y')) && (ps_at!(8) == 0)))))) {
            res.w[1] = 0x7800000000000000;
            return (res, pfpsf);
        }
        if ((((tolower_macro(ps_at!(0)) == b's') && (tolower_macro(ps_at!(1)) == b'n')) && (tolower_macro(ps_at!(2)) == b'a')) && (tolower_macro(ps_at!(3)) == b'n')) {
            res.w[1] = 0x7e00000000000000;
            return (res, pfpsf);
        }
        res.w[1] = 0x7c00000000000000;
        return (res, pfpsf);
    }
    if (((((tolower_macro(ps_at!(1)) == b'i') && (tolower_macro(ps_at!(2)) == b'n')) && (tolower_macro(ps_at!(3)) == b'f'))) && (((ps_at!(4) == 0) || (((((((tolower_macro(ps_at!(4)) == b'i') && (tolower_macro(ps_at!(5)) == b'n')) && (tolower_macro(ps_at!(6)) == b'i')) && (tolower_macro(ps_at!(7)) == b't')) && (tolower_macro(ps_at!(8)) == b'y')) && (ps_at!(9) == 0)))))) {
        res.w[0] = 0;
        if (c == b'+') {
            res.w[1] = 0x7800000000000000;
        } else if (c == b'-') {
            res.w[1] = 0xf800000000000000;
        } else {
            res.w[1] = 0x7c00000000000000;
        }
        return (res, pfpsf);
    }
    if ((((tolower_macro(ps_at!(1)) == b's') && (tolower_macro(ps_at!(2)) == b'n')) && (tolower_macro(ps_at!(3)) == b'a')) && (tolower_macro(ps_at!(4)) == b'n')) {
        res.w[0] = 0;
        if (c == b'-') {
            res.w[1] = 0xfe00000000000000;
        } else {
            res.w[1] = 0x7e00000000000000;
        }
        return (res, pfpsf);
    }
    if (c == b'-') {
        sign_x = 0x8000000000000000;
    } else {
        sign_x = 0;
    }
    if ((c == b'-') || (c == b'+')) {
        ps_idx += 1;
    }
    c = ps_at!(0);
    if ((c != b'.') && ((((c.wrapping_sub(b'0')) as u64) > 9))) {
        res.w[1] = (0x7c00000000000000 | sign_x);
        res.w[0] = 0;
        return (res, pfpsf);
    }
    if (c == b'.') {
        rdx_pt_enc = 1;
        ps_idx += 1;
    }
    if (ps_at!(0) == b'0') {
        while (ps_at!(0) == b'0') {
            ps_idx += 1;
            if (rdx_pt_enc != 0) {
                right_radix_leading_zeros = right_radix_leading_zeros.wrapping_add(1);
            }
            if (ps_at!(0) == b'.') {
                if (rdx_pt_enc == 0) {
                    rdx_pt_enc = 1;
                    if (ps_at!(1) == 0) {
                        res.w[1] = ((((0x3040000000000000 as u64).wrapping_sub(((go_checked_shl_u64(right_radix_leading_zeros, go_shift_count_u64((49) as u64))))))) | sign_x);
                        res.w[0] = 0;
                        return (res, pfpsf);
                    }
                    ps_idx += 1;
                } else {
                    res.w[1] = (0x7c00000000000000 | sign_x);
                    res.w[0] = 0;
                    return (res, pfpsf);
                }
            } else if (ps_at!(0) == 0) {
                if (right_radix_leading_zeros > 6176) {
                    right_radix_leading_zeros = 6176;
                }
                res.w[1] = ((((0x3040000000000000 as u64).wrapping_sub(((go_checked_shl_u64(right_radix_leading_zeros, go_shift_count_u64((49) as u64))))))) | sign_x);
                res.w[0] = 0;
                return (res, pfpsf);
            }
        }
    }
    c = ps_at!(0);
    ndigits_before = 0;
    ndigits_after = 0;
    ndigits_total = 0;
    sgn_exp = 0;
    if (rdx_pt_enc == 0) {
        while (((c.wrapping_sub(b'0')) as u64) <= 9) {
            if (ndigits_before < 34) {
                buffer[ndigits_before as usize] = c;
            } else if (ndigits_before < 100) {
                buffer[ndigits_before as usize] = c;
                if (c > b'0') {
                    set_inexact = 1;
                }
            } else if (c > b'0') {
                set_inexact = 1;
                sticky_bit = 1;
            }
            ps_idx += 1;
            c = ps_at!(0);
            ndigits_before = ndigits_before.wrapping_add(1);
        }
        ndigits_total = ndigits_before;
        if (c == b'.') {
            ps_idx += 1;
            c = ps_at!(0);
            if (c != 0) {
                while (((c.wrapping_sub(b'0')) as u64) <= 9) {
                    if (ndigits_total < 34) {
                        buffer[ndigits_total as usize] = c;
                    } else if (ndigits_total < 100) {
                        buffer[ndigits_total as usize] = c;
                        if (c > b'0') {
                            set_inexact = 1;
                        }
                    } else if (c > b'0') {
                        set_inexact = 1;
                        sticky_bit = 1;
                    }
                    ps_idx += 1;
                    c = ps_at!(0);
                    ndigits_total = ndigits_total.wrapping_add(1);
                }
                ndigits_after = (ndigits_total.wrapping_sub(ndigits_before));
            }
        }
    } else {
        c = ps_at!(0);
        ndigits_total = 0;
        while (((c.wrapping_sub(b'0')) as u64) <= 9) {
            if (ndigits_total < 34) {
                buffer[ndigits_total as usize] = c;
            } else if (ndigits_total < 100) {
                buffer[ndigits_total as usize] = c;
                if (c > b'0') {
                    set_inexact = 1;
                }
            } else if (c > b'0') {
                set_inexact = 1;
                sticky_bit = 1;
            }
            ps_idx += 1;
            c = ps_at!(0);
            ndigits_total = ndigits_total.wrapping_add(1);
        }
        ndigits_after = (ndigits_total.wrapping_sub(ndigits_before));
    }
    dec_expon = 0;
    if (c != 0) {
        if ((c != b'e') && (c != b'E')) {
            res.w[1] = 0x7c00000000000000;
            res.w[0] = 0;
            return (res, pfpsf);
        }
        ps_idx += 1;
        c = ps_at!(0);
        if (((((c.wrapping_sub(b'0')) as u64) > 9)) && (((((c != b'+') && (c != b'-'))) || ((((ps_at!(1).wrapping_sub(b'0')) as u64) > 9))))) {
            res.w[1] = 0x7c00000000000000;
            res.w[0] = 0;
            return (res, pfpsf);
        }
        if (c == b'-') {
            sgn_exp = (-1);
            ps_idx += 1;
            c = ps_at!(0);
        } else if (c == b'+') {
            ps_idx += 1;
            c = ps_at!(0);
        }
        dec_expon = ((c.wrapping_sub(b'0')) as i64);
        i = 1;
        ps_idx += 1;
        if (dec_expon == 0) {
            while (ps_at!(0) == b'0') {
                ps_idx += 1;
            }
        }
        c = (ps_at!(0).wrapping_sub(b'0'));
        while (((c as u64) <= 9) && (i < 7)) {
            d2 = (dec_expon.wrapping_add(dec_expon));
            dec_expon = ((((go_checked_shl_i64(d2, go_shift_count_u64((2) as u64)))).wrapping_add(d2)).wrapping_add(c as i64));
            ps_idx += 1;
            c = (ps_at!(0).wrapping_sub(b'0'));
            i = i.wrapping_add(1);
        }
    }
    dec_expon = (((dec_expon.wrapping_add(sgn_exp))) ^ sgn_exp);
    if (ndigits_total <= 34) {
        dec_expon = dec_expon.wrapping_add((((0x1820 as i64).wrapping_sub(ndigits_after)).wrapping_sub(right_radix_leading_zeros as i64)));
        if (dec_expon < 0) {
            res.w[1] = (0 | sign_x);
            res.w[0] = 0;
        }
        if (ndigits_total == 0) {
            CX.w[0] = 0;
            CX.w[1] = 0;
        } else if (ndigits_total <= 19) {
            coeff_high = ((buffer[0].wrapping_sub(b'0')) as u64);
            i = 1;
            while (i < ndigits_total) {
                coeff2 = (coeff_high.wrapping_add(coeff_high));
                coeff_high = ((((go_checked_shl_u64(coeff2, go_shift_count_u64((2) as u64)))).wrapping_add(coeff2)).wrapping_add(((buffer[i as usize].wrapping_sub(b'0')) as u64)));
                i = i.wrapping_add(1);
            }
            CX.w[0] = coeff_high;
            CX.w[1] = 0;
        } else {
            coeff_high = ((buffer[0].wrapping_sub(b'0')) as u64);
            i = 1;
            while (i < (ndigits_total.wrapping_sub(17))) {
                coeff2 = (coeff_high.wrapping_add(coeff_high));
                coeff_high = ((((go_checked_shl_u64(coeff2, go_shift_count_u64((2) as u64)))).wrapping_add(coeff2)).wrapping_add(((buffer[i as usize].wrapping_sub(b'0')) as u64)));
                i = i.wrapping_add(1);
            }
            coeff_low = ((buffer[i as usize].wrapping_sub(b'0')) as u64);
            i = i.wrapping_add(1);
            while (i < ndigits_total) {
                coeff_l2 = (coeff_low.wrapping_add(coeff_low));
                coeff_low = ((((go_checked_shl_u64(coeff_l2, go_shift_count_u64((2) as u64)))).wrapping_add(coeff_l2)).wrapping_add(((buffer[i as usize].wrapping_sub(b'0')) as u64)));
                i = i.wrapping_add(1);
            }
            scale_high = 100000000000000000;
            CX = __mul_64x64_to_128_fast(coeff_high, scale_high);
            CX.w[0] = CX.w[0].wrapping_add(coeff_low);
            if (CX.w[0] < coeff_low) {
                CX.w[1] = CX.w[1].wrapping_add(1);
            }
        }
        res = bid_get_bid128(sign_x, dec_expon, CX, rnd_mode, (&mut pfpsf));
        return (res, pfpsf);
    } else {
        dec_expon = dec_expon.wrapping_add((((ndigits_before.wrapping_add(0x1820)).wrapping_sub(34)).wrapping_sub(right_radix_leading_zeros as i64)));
        if (dec_expon < 0) {
            res.w[1] = (0 | sign_x);
            res.w[0] = 0;
        }
        coeff_high = ((buffer[0].wrapping_sub(b'0')) as u64);
        i = 1;
        while (i < (34 - 17)) {
            coeff2 = (coeff_high.wrapping_add(coeff_high));
            coeff_high = ((((go_checked_shl_u64(coeff2, go_shift_count_u64((2) as u64)))).wrapping_add(coeff2)).wrapping_add(((buffer[i as usize].wrapping_sub(b'0')) as u64)));
            i = i.wrapping_add(1);
        }
        coeff_low = ((buffer[i as usize].wrapping_sub(b'0')) as u64);
        i = i.wrapping_add(1);
        while (i < 34) {
            coeff_l2 = (coeff_low.wrapping_add(coeff_low));
            coeff_low = ((((go_checked_shl_u64(coeff_l2, go_shift_count_u64((2) as u64)))).wrapping_add(coeff_l2)).wrapping_add(((buffer[i as usize].wrapping_sub(b'0')) as u64)));
            i = i.wrapping_add(1);
        }
        match rnd_mode {
            0 => {
                carry = (go_checked_shr_u64((((((b'4' as i64) as i64).wrapping_sub(buffer[i as usize] as i64)) as u32) as u64), go_shift_count_u64((31) as u64)));
                if (((((buffer[i as usize] == b'5') && ((coeff_low & 1) == 0)) && (sticky_bit == 0))) || (dec_expon < 0)) {
                    if (dec_expon >= 0) {
                        carry = 0;
                        i = i.wrapping_add(1);
                    }
                    min_digits = ndigits_total;
                    if (min_digits > 100) {
                        min_digits = 100;
                    }
                    carry = (sticky_bit as u64);
                    while ((carry == 0) && (i < min_digits)) {
                        if (buffer[i as usize] > b'0') {
                            carry = 1;
                            break;
                        }
                        i = i.wrapping_add(1);
                    }
                }
            }
            1 => {
                carry = 0;
                if (sign_x != 0) {
                    min_digits = ndigits_total;
                    if (min_digits > 100) {
                        min_digits = 100;
                    }
                    carry = (sticky_bit as u64);
                    while ((carry == 0) && (i < min_digits)) {
                        if (buffer[i as usize] > b'0') {
                            carry = 1;
                            break;
                        }
                        i = i.wrapping_add(1);
                    }
                }
            }
            2 => {
                carry = 0;
                if (sign_x == 0) {
                    min_digits = ndigits_total;
                    if (min_digits > 100) {
                        min_digits = 100;
                    }
                    carry = (sticky_bit as u64);
                    while ((carry == 0) && (i < min_digits)) {
                        if (buffer[i as usize] > b'0') {
                            carry = 1;
                            break;
                        }
                        i = i.wrapping_add(1);
                    }
                }
            }
            3 => {
                carry = 0;
            }
            4 => {
                carry = (go_checked_shr_u64((((((b'4' as i64) as i64).wrapping_sub(buffer[i as usize] as i64)) as u32) as u64), go_shift_count_u64((31) as u64)));
                if (dec_expon < 0) {
                    min_digits = ndigits_total;
                    if (min_digits > 100) {
                        min_digits = 100;
                    }
                    carry = (sticky_bit as u64);
                    while ((carry == 0) && (i < min_digits)) {
                        if (buffer[i as usize] > b'0') {
                            carry = 1;
                            break;
                        }
                        i = i.wrapping_add(1);
                    }
                }
            }
            _ => {
                carry = 0;
            }
        }
        scale_high = 100000000000000000;
        if (dec_expon < 0) {
            if (dec_expon > (-34)) {
                scale_high = 1000000000000000000;
                coeff_low = (((go_checked_shl_u64(coeff_low, go_shift_count_u64((3) as u64)))).wrapping_add(((go_checked_shl_u64(coeff_low, go_shift_count_u64((1) as u64))))));
                dec_expon = dec_expon.wrapping_sub(1);
            }
            if ((dec_expon == (-34)) && (coeff_high > 50000000000000000)) {
                carry = 0;
            }
        }
        CX = __mul_64x64_to_128_fast(coeff_high, scale_high);
        coeff_low = coeff_low.wrapping_add(carry);
        CX.w[0] = CX.w[0].wrapping_add(coeff_low);
        if (CX.w[0] < coeff_low) {
            CX.w[1] = CX.w[1].wrapping_add(1);
        }
        if (set_inexact != 0) {
            pfpsf |= 32;
        }
        res = bid_get_bid128(sign_x, dec_expon, CX, rnd_mode, (&mut pfpsf));
        return (res, pfpsf);
    }
}
