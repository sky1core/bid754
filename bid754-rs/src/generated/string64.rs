// Auto-generated from string64.go by go2rs. Do not edit.

use super::prelude::*;

pub(crate) fn l0_normalize_10to18(X_hi: &mut u64, X_lo: &mut u64) {
    let mut L0_tmp = ((*X_lo).wrapping_add(0x21f494c589c0000));
    if ((L0_tmp & 0x1000000000000000) != 0) {
        (*X_hi) = ((*X_hi).wrapping_add(1));
        (*X_lo) = (go_checked_shr_u64(((go_checked_shl_u64(L0_tmp, go_shift_count_u64((4) as u64)))), go_shift_count_u64((4) as u64)));
    }
}

pub(crate) fn l0_split_mi_di_2(mut X: u32, mut MiDi: &mut [u32], ptr: &mut i64) {
    let mut L0_head = (go_checked_shr_u32(X, go_shift_count_u64((10) as u64)));
    let mut L0_tail = (((X & 0x03FF).wrapping_add(((go_checked_shl_u32(L0_head, go_shift_count_u64((5) as u64)))))).wrapping_sub(((go_checked_shl_u32(L0_head, go_shift_count_u64((3) as u64))))));
    let mut L0_tmp = (go_checked_shr_u32(L0_tail, go_shift_count_u64((10) as u64)));
    L0_head = L0_head.wrapping_add(L0_tmp);
    L0_tail = (((L0_tail & 0x03FF).wrapping_add(((go_checked_shl_u32(L0_tmp, go_shift_count_u64((5) as u64)))))).wrapping_sub(((go_checked_shl_u32(L0_tmp, go_shift_count_u64((3) as u64))))));
    if (L0_tail > 999) {
        L0_tail = L0_tail.wrapping_sub(1000);
        L0_head = L0_head.wrapping_add(1);
    }
    MiDi[(*ptr) as usize] = L0_head;
    (*ptr) = (*ptr).wrapping_add(1);
    MiDi[(*ptr) as usize] = L0_tail;
    (*ptr) = (*ptr).wrapping_add(1);
}

pub(crate) fn l0_split_mi_di_3(mut X: u64, mut MiDi: &mut [u32], ptr: &mut i64) {
    let mut L0_X = (X as u32);
    let mut L0_head = (go_checked_shr_u32(((((go_checked_shr_u32(L0_X, go_shift_count_u64((17) as u64)))).wrapping_mul(34359))), go_shift_count_u64((18) as u64)));
    L0_X = L0_X.wrapping_sub((L0_head.wrapping_mul(1000000)));
    if (L0_X >= 1000000) {
        L0_X = L0_X.wrapping_sub(1000000);
        L0_head = L0_head.wrapping_add(1);
    }
    let mut L0_mid = (go_checked_shr_u32(L0_X, go_shift_count_u64((10) as u64)));
    let mut L0_tail = (((L0_X & 0x03FF).wrapping_add(((go_checked_shl_u32(L0_mid, go_shift_count_u64((5) as u64)))))).wrapping_sub(((go_checked_shl_u32(L0_mid, go_shift_count_u64((3) as u64))))));
    let mut L0_tmp = (go_checked_shr_u32(L0_tail, go_shift_count_u64((10) as u64)));
    L0_mid = L0_mid.wrapping_add(L0_tmp);
    L0_tail = (((L0_tail & 0x3FF).wrapping_add(((go_checked_shl_u32(L0_tmp, go_shift_count_u64((5) as u64)))))).wrapping_sub(((go_checked_shl_u32(L0_tmp, go_shift_count_u64((3) as u64))))));
    if (L0_tail > 999) {
        L0_tail = L0_tail.wrapping_sub(1000);
        L0_mid = L0_mid.wrapping_add(1);
    }
    MiDi[(*ptr) as usize] = L0_head;
    (*ptr) = (*ptr).wrapping_add(1);
    MiDi[(*ptr) as usize] = L0_mid;
    (*ptr) = (*ptr).wrapping_add(1);
    MiDi[(*ptr) as usize] = L0_tail;
    (*ptr) = (*ptr).wrapping_add(1);
}

pub(crate) fn l1_split_mi_di_6_lead(mut X: u64, mut MiDi: &mut [u32], ptr: &mut i64) {
    if (X >= (0x3b9aca00 as u64)) {
        let mut L1_Xhi_64 = (go_checked_shr_u64(((((go_checked_shr_u64(X, go_shift_count_u64((28) as u64)))).wrapping_mul(0x89705f41))), go_shift_count_u64((33) as u64)));
        let mut L1_Xlo_64 = (X.wrapping_sub((L1_Xhi_64.wrapping_mul(0x3b9aca00 as u64))));
        if (L1_Xlo_64 >= (0x3b9aca00 as u64)) {
            L1_Xlo_64 = L1_Xlo_64.wrapping_sub(0x3b9aca00 as u64);
            L1_Xhi_64 = L1_Xhi_64.wrapping_add(1);
        }
        let mut L1_X_hi = (L1_Xhi_64 as u32);
        let mut L1_X_lo = (L1_Xlo_64 as u32);
        if (L1_X_hi >= 0xf4240) {
            l0_split_mi_di_3((L1_X_hi as u64), MiDi, ptr);
            l0_split_mi_di_3((L1_X_lo as u64), MiDi, ptr);
        } else if (L1_X_hi >= 0x3e8) {
            l0_split_mi_di_2(L1_X_hi, MiDi, ptr);
            l0_split_mi_di_3((L1_X_lo as u64), MiDi, ptr);
        } else {
            MiDi[(*ptr) as usize] = L1_X_hi;
            (*ptr) = (*ptr).wrapping_add(1);
            l0_split_mi_di_3((L1_X_lo as u64), MiDi, ptr);
        }
    } else {
        let mut L1_X_lo = (X as u32);
        if (L1_X_lo >= 0xf4240) {
            l0_split_mi_di_3((L1_X_lo as u64), MiDi, ptr);
        } else if (L1_X_lo >= 0x3e8) {
            l0_split_mi_di_2(L1_X_lo, MiDi, ptr);
        } else {
            MiDi[(*ptr) as usize] = L1_X_lo;
            (*ptr) = (*ptr).wrapping_add(1);
        }
    }
}

pub(crate) fn l0_mi_di2_str(mut X: u32, mut ps: &mut [u8], c_ptr: &mut i64) {
    let mut src = bid_midi_tbl[X as usize];
    ps[(*c_ptr) as usize] = src[0];
    (*c_ptr) = (*c_ptr).wrapping_add(1);
    ps[(*c_ptr) as usize] = src[1];
    (*c_ptr) = (*c_ptr).wrapping_add(1);
    ps[(*c_ptr) as usize] = src[2];
    (*c_ptr) = (*c_ptr).wrapping_add(1);
}

pub(crate) fn l0_mi_di2_str_lead(mut X: u32, mut ps: &mut [u8], c_ptr: &mut i64) {
    let mut src = bid_midi_tbl[X as usize];
    if (X >= 100) {
        ps[(*c_ptr) as usize] = src[0];
        (*c_ptr) = (*c_ptr).wrapping_add(1);
        ps[(*c_ptr) as usize] = src[1];
        (*c_ptr) = (*c_ptr).wrapping_add(1);
        ps[(*c_ptr) as usize] = src[2];
        (*c_ptr) = (*c_ptr).wrapping_add(1);
    } else if (X >= 10) {
        ps[(*c_ptr) as usize] = src[1];
        (*c_ptr) = (*c_ptr).wrapping_add(1);
        ps[(*c_ptr) as usize] = src[2];
        (*c_ptr) = (*c_ptr).wrapping_add(1);
    } else {
        ps[(*c_ptr) as usize] = src[2];
        (*c_ptr) = (*c_ptr).wrapping_add(1);
    }
}

pub fn bid64_to_string(mut x: u64) -> String {
    let mut ps: [u8; 64] = [0; 64];
    let mut istart: i64 = 0;
    let (mut sign_x, mut exponent_x, mut coefficient_x, mut valid) = unpack_bid64(x);
    if (!valid) {
        if ((x & 0x7800000000000000) == 0x7800000000000000) {
            if ((x & 0x7c00000000000000) == 0x7c00000000000000) {
                if (sign_x != 0) {
                    ps[0] = b'-';
                } else {
                    ps[0] = b'+';
                }
                ps[1] = b'S';
                let mut j: i64 = 2;
                if ((x & 0x7e00000000000000) != 0x7e00000000000000) {
                    j = 1;
                }
                ps[j as usize] = b'N';
                j = j.wrapping_add(1);
                ps[j as usize] = b'a';
                j = j.wrapping_add(1);
                ps[j as usize] = b'N';
                j = j.wrapping_add(1);
                return go_string_from_bytes(&mut ps[..j as usize]);
            }
            if (sign_x != 0) {
                ps[0] = b'-';
            } else {
                ps[0] = b'+';
            }
            ps[1] = b'I';
            ps[2] = b'n';
            ps[3] = b'f';
            return go_string_from_bytes(&mut ps[..4 as usize]);
        }
        istart = 1;
        if (sign_x != 0) {
            ps[0] = b'-';
        } else {
            ps[0] = b'+';
        }
        ps[istart as usize] = b'0';
        istart = istart.wrapping_add(1);
        ps[istart as usize] = b'E';
        istart = istart.wrapping_add(1);
        exponent_x = exponent_x.wrapping_sub(398);
        if (exponent_x < 0) {
            ps[istart as usize] = b'-';
            istart = istart.wrapping_add(1);
            exponent_x = (exponent_x.wrapping_neg());
        } else {
            ps[istart as usize] = b'+';
            istart = istart.wrapping_add(1);
        }
        if (exponent_x != 0) {
            let mut tempx = (exponent_x as f32);
            let mut bin_expon_cx = (((((go_checked_shr_u32((tempx as f32).to_bits(), go_shift_count_u64((23) as u64)))) & 0xff) as i64).wrapping_sub(0x7f));
            let mut digits_x = (bid_estimate_decimal_digits[bin_expon_cx as usize] as i64);
            if ((exponent_x as u64) >= bid_power10_table_128[digits_x as usize].w[0]) {
                digits_x = digits_x.wrapping_add(1);
            }
            let mut j = ((istart.wrapping_add(digits_x)).wrapping_sub(1));
            istart = (j.wrapping_add(1));
            let mut ER10: u64 = (0x1999999a as u64);
            let mut exp = exponent_x;
            while (exp > 9) {
                let mut D = ((exp as u64).wrapping_mul(ER10));
                D = go_checked_shr_u64(D, go_shift_count_u64((32) as u64));
                exp = ((exp.wrapping_sub(((go_checked_shl_u64(D, go_shift_count_u64((1) as u64))) as i64))).wrapping_sub(((go_checked_shl_u64(D, go_shift_count_u64((3) as u64))) as i64)));
                ps[j as usize] = ((b'0' as u8).wrapping_add(exp as u8));
                j = j.wrapping_sub(1);
                exp = (D as i64);
            }
            ps[j as usize] = ((b'0' as u8).wrapping_add(exp as u8));
        } else {
            ps[istart as usize] = b'0';
            istart = istart.wrapping_add(1);
        }
        return go_string_from_bytes(&mut ps[..istart as usize]);
    }
    exponent_x = exponent_x.wrapping_sub(0x18e);
    istart = 1;
    if (sign_x != 0) {
        ps[0] = b'-';
    } else {
        ps[0] = b'+';
    }
    if ((coefficient_x > 9999999999999999) || (coefficient_x == 0)) {
        ps[istart as usize] = b'0';
        istart = istart.wrapping_add(1);
    } else {
        let mut MiDi: [u32; 12] = [0; 12];
        let mut ptr: i64 = 0;
        let mut Tmp = (go_checked_shr_u64(coefficient_x, go_shift_count_u64((59) as u64)));
        let mut LO_18Dig = (go_checked_shr_u64(((go_checked_shl_u64(coefficient_x, go_shift_count_u64((5) as u64)))), go_shift_count_u64((5) as u64)));
        let mut HI_18Dig: u64 = (0 as u64);
        let mut k_lcv: i64 = 0;
        while (Tmp != 0) {
            let mut midi_ind = ((Tmp & 0x000000000000003F) as i64);
            midi_ind = go_checked_shl_i64(midi_ind, go_shift_count_u64((1) as u64));
            Tmp = go_checked_shr_u64(Tmp, go_shift_count_u64((6) as u64));
            HI_18Dig = HI_18Dig.wrapping_add(mod10_18_tbl[k_lcv as usize][midi_ind as usize]);
            midi_ind = midi_ind.wrapping_add(1);
            LO_18Dig = LO_18Dig.wrapping_add(mod10_18_tbl[k_lcv as usize][midi_ind as usize]);
            k_lcv = k_lcv.wrapping_add(1);
            l0_normalize_10to18((&mut HI_18Dig), (&mut LO_18Dig));
        }
        l1_split_mi_di_6_lead(LO_18Dig, &mut MiDi[..], (&mut ptr));
        let mut length = ptr;
        let mut c_ptr = istart;
        l0_mi_di2_str_lead(MiDi[0], &mut ps[..], (&mut c_ptr));
        let mut k: i64 = 1;
        while (k < length) {
            l0_mi_di2_str(MiDi[k as usize], &mut ps[..], (&mut c_ptr));
            k = k.wrapping_add(1);
        }
        istart = c_ptr;
    }
    ps[istart as usize] = b'E';
    istart = istart.wrapping_add(1);
    if (exponent_x < 0) {
        ps[istart as usize] = b'-';
        istart = istart.wrapping_add(1);
        exponent_x = (exponent_x.wrapping_neg());
    } else {
        ps[istart as usize] = b'+';
        istart = istart.wrapping_add(1);
    }
    if (exponent_x != 0) {
        let mut tempx = (exponent_x as f32);
        let mut bin_expon_cx = (((((go_checked_shr_u32((tempx as f32).to_bits(), go_shift_count_u64((23) as u64)))) & 0xff) as i64).wrapping_sub(0x7f));
        let mut digits_x = (bid_estimate_decimal_digits[bin_expon_cx as usize] as i64);
        if ((exponent_x as u64) >= bid_power10_table_128[digits_x as usize].w[0]) {
            digits_x = digits_x.wrapping_add(1);
        }
        let mut j = ((istart.wrapping_add(digits_x)).wrapping_sub(1));
        istart = (j.wrapping_add(1));
        let mut ER10: u64 = (0x1999999a as u64);
        let mut exp = exponent_x;
        while (exp > 9) {
            let mut D = ((exp as u64).wrapping_mul(ER10));
            D = go_checked_shr_u64(D, go_shift_count_u64((32) as u64));
            exp = ((exp.wrapping_sub(((go_checked_shl_u64(D, go_shift_count_u64((1) as u64))) as i64))).wrapping_sub(((go_checked_shl_u64(D, go_shift_count_u64((3) as u64))) as i64)));
            ps[j as usize] = ((b'0' as u8).wrapping_add(exp as u8));
            j = j.wrapping_sub(1);
            exp = (D as i64);
        }
        ps[j as usize] = ((b'0' as u8).wrapping_add(exp as u8));
    } else {
        ps[istart as usize] = b'0';
        istart = istart.wrapping_add(1);
    }
    return go_string_from_bytes(&mut ps[..istart as usize]);
}
