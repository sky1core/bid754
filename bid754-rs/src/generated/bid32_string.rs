// Auto-generated from bid32_string.go by go2rs. Do not edit.

use super::prelude::*;
use core::fmt::Write as _;

pub fn bid32_to_string_raw(mut x: u32) -> String {
    let mut CT: u64 = 0;
    let mut d: i64 = 0;
    let mut istart: i64 = 0;
    let mut istart0: i64 = 0;
    let mut sign_x: u32 = 0;
    let mut coefficient_x: u32 = 0;
    let mut exponent_x: i64 = 0;
    let mut ps: [u8; 64] = [0; 64];
    let (mut sign_x, mut exponent_x, mut coefficient_x, mut valid) = unpack_bid32(x);
    if (!valid) {
        if (sign_x != 0) {
            ps[0] = b'-';
        } else {
            ps[0] = b'+';
        }
        if ((x & 0x7c000000) == 0x7c000000) {
            if ((x & 0x7e000000) == 0x7e000000) {
                return (go_string_from_bytes(&mut ps[..1 as usize]) + "SNaN");
            }
            return (go_string_from_bytes(&mut ps[..1 as usize]) + "NaN");
        }
        if ((x & 0x78000000) == 0x78000000) {
            return (go_string_from_bytes(&mut ps[..1 as usize]) + "Inf");
        }
        istart = 1;
        ps[istart as usize] = b'0';
        istart = istart.wrapping_add(1);
    } else {
        if (sign_x != 0) {
            ps[0] = b'-';
        } else {
            ps[0] = b'+';
        }
        istart = 1;
        if (coefficient_x >= 1000000) {
            CT = ((coefficient_x as u64).wrapping_mul(0x431BDE83));
            CT = go_checked_shr_u64(CT, go_shift_count_u64((32) as u64));
            d = ((go_checked_shr_u64(CT, go_shift_count_u64((50 - 32) as u64))) as i64);
            ps[istart as usize] = ((d as u8).wrapping_add(b'0'));
            istart = istart.wrapping_add(1);
            coefficient_x = coefficient_x.wrapping_sub(((d as u32).wrapping_mul(1000000)));
            CT = ((coefficient_x as u64).wrapping_mul(0x20C49BA6));
            CT = go_checked_shr_u64(CT, go_shift_count_u64((32) as u64));
            d = ((go_checked_shr_u64(CT, go_shift_count_u64((39 - 32) as u64))) as i64);
            ps[istart as usize] = bid_midi_tbl[d as usize][0];
            istart = istart.wrapping_add(1);
            ps[istart as usize] = bid_midi_tbl[d as usize][1];
            istart = istart.wrapping_add(1);
            ps[istart as usize] = bid_midi_tbl[d as usize][2];
            istart = istart.wrapping_add(1);
            d = ((coefficient_x as i64).wrapping_sub((d.wrapping_mul(1000))));
            ps[istart as usize] = bid_midi_tbl[d as usize][0];
            istart = istart.wrapping_add(1);
            ps[istart as usize] = bid_midi_tbl[d as usize][1];
            istart = istart.wrapping_add(1);
            ps[istart as usize] = bid_midi_tbl[d as usize][2];
            istart = istart.wrapping_add(1);
        } else if (coefficient_x >= 1000) {
            CT = ((coefficient_x as u64).wrapping_mul(0x20C49BA6));
            CT = go_checked_shr_u64(CT, go_shift_count_u64((32) as u64));
            d = ((go_checked_shr_u64(CT, go_shift_count_u64((39 - 32) as u64))) as i64);
            istart0 = istart;
            ps[istart as usize] = bid_midi_tbl[d as usize][0];
            if (ps[istart as usize] != b'0') {
                istart = istart.wrapping_add(1);
            }
            ps[istart as usize] = bid_midi_tbl[d as usize][1];
            if ((ps[istart as usize] != b'0') || (istart != istart0)) {
                istart = istart.wrapping_add(1);
            }
            ps[istart as usize] = bid_midi_tbl[d as usize][2];
            istart = istart.wrapping_add(1);
            d = ((coefficient_x as i64).wrapping_sub((d.wrapping_mul(1000))));
            ps[istart as usize] = bid_midi_tbl[d as usize][0];
            istart = istart.wrapping_add(1);
            ps[istart as usize] = bid_midi_tbl[d as usize][1];
            istart = istart.wrapping_add(1);
            ps[istart as usize] = bid_midi_tbl[d as usize][2];
            istart = istart.wrapping_add(1);
        } else {
            d = (coefficient_x as i64);
            istart0 = istart;
            ps[istart as usize] = bid_midi_tbl[d as usize][0];
            if (ps[istart as usize] != b'0') {
                istart = istart.wrapping_add(1);
            }
            ps[istart as usize] = bid_midi_tbl[d as usize][1];
            if ((ps[istart as usize] != b'0') || (istart != istart0)) {
                istart = istart.wrapping_add(1);
            }
            ps[istart as usize] = bid_midi_tbl[d as usize][2];
            istart = istart.wrapping_add(1);
        }
    }
    if (!valid) {
        ps[istart as usize] = b'E';
        istart = istart.wrapping_add(1);
        exponent_x = exponent_x.wrapping_sub(101);
        if (exponent_x < 0) {
            ps[istart as usize] = b'-';
            istart = istart.wrapping_add(1);
            exponent_x = (exponent_x.wrapping_neg());
        } else {
            ps[istart as usize] = b'+';
            istart = istart.wrapping_add(1);
        }
        istart0 = istart;
        ps[istart as usize] = bid_midi_tbl[exponent_x as usize][0];
        if (ps[istart as usize] != b'0') {
            istart = istart.wrapping_add(1);
        }
        ps[istart as usize] = bid_midi_tbl[exponent_x as usize][1];
        if ((ps[istart as usize] != b'0') || (istart != istart0)) {
            istart = istart.wrapping_add(1);
        }
        ps[istart as usize] = bid_midi_tbl[exponent_x as usize][2];
        istart = istart.wrapping_add(1);
        return go_string_from_bytes(&mut ps[..istart as usize]);
    }
    let mut digits = go_string_from_bytes(&mut ps[1 as usize..istart as usize]);
    let mut adjustedExp = (((exponent_x - 101) + (digits.len() as i64)) - 1);
    let mut out = String::with_capacity(digits.len() + 12);
    out.push(ps[0] as char);
    out.push_str(&digits[..1 as usize]);
    if ((digits.len() as i64) > 1) {
        out.push('.');
        out.push_str(&digits[1 as usize..]);
    }
    if (adjustedExp != 0) {
        out.push('e');
        let _ = write!(&mut out, "{}", adjustedExp);
    }
    return out;
}

pub fn bid32_from_string_raw(ps: impl AsRef<str>, mut rnd_mode: i64) -> (u32, u32) {
    let mut ps = ps.as_ref().to_string();
    let mut sign_x: u64 = 0;
    let mut coefficient_x: u64 = 0;
    let mut rounded: u64 = 0;
    let mut expon_x: i64 = 0;
    let mut sgn_expon: i64 = 0;
    let mut ndigits: i64 = 0;
    let mut add_expon: i64 = 0;
    let mut midpoint: i64 = 0;
    let mut rounded_up: i64 = 0;
    let mut dround: i64 = 0;
    let mut dec_expon_scale: i64 = 0;
    let mut right_radix_leading_zeros: i64 = 0;
    let mut rdx_pt_enc: i64 = 0;
    let mut pfpsf: u32 = 0;
    let mut res: u64 = 0;
    let s = (ps).trim_start_matches(|c| " \t".contains(c)).as_bytes();
    if ((s.len() as i64) == 0) {
        return (0x7c000000, 0);
    }
    let mut c = s[0];
    let mut idx: i64 = 0;
    if ((((c != b'.') && (c != b'-')) && (c != b'+')) && (((c < b'0') || (c > b'9')))) {
        if (s.eq_ignore_ascii_case(b"inf") || s.eq_ignore_ascii_case(b"infinity")) {
            return (0x78000000, 0);
        }
        if ((s.len() >= 4) && s[..4].eq_ignore_ascii_case(b"snan")) {
            return (0x7e000000, 0);
        }
        return (0x7c000000, 0);
    }
    if ((s.len() as i64) > 1) {
        let sl1 = &s[1 as usize..];
        if (sl1.eq_ignore_ascii_case(b"inf") || sl1.eq_ignore_ascii_case(b"infinity")) {
            if (c == b'+') {
                return (0x78000000, 0);
            } else if (c == b'-') {
                return (0xf8000000, 0);
            }
            return (0x7c000000, 0);
        }
        if ((sl1.len() >= 4) && sl1[..4].eq_ignore_ascii_case(b"snan")) {
            if (c == b'-') {
                return (0xfe000000, 0);
            }
            return (0x7e000000, 0);
        }
        if (sl1.eq_ignore_ascii_case(b"nan")) {
            if (c == b'-') {
                return (0xfc000000, 0);
            }
            return (0x7c000000, 0);
        }
    }
    if (c == b'-') {
        sign_x = 0x80000000;
    } else {
        sign_x = 0;
    }
    if ((c == b'-') || (c == b'+')) {
        idx = idx.wrapping_add(1);
        if (idx >= (s.len() as i64)) {
            return (((0x7c000000 | sign_x) as u32), 0);
        }
        c = s[idx as usize];
    }
    if ((c != b'.') && (((c < b'0') || (c > b'9')))) {
        return (((0x7c000000 | sign_x) as u32), 0);
    }
    rdx_pt_enc = 0;
    if ((idx < (s.len() as i64)) && (((s[idx as usize] == b'0') || (s[idx as usize] == b'.')))) {
        if (s[idx as usize] == b'.') {
            rdx_pt_enc = 1;
            idx = idx.wrapping_add(1);
        }
        while ((idx < (s.len() as i64)) && (s[idx as usize] == b'0')) {
            idx = idx.wrapping_add(1);
            if (rdx_pt_enc != 0) {
                right_radix_leading_zeros = right_radix_leading_zeros.wrapping_add(1);
            }
            if ((idx < (s.len() as i64)) && (s[idx as usize] == b'.')) {
                if (rdx_pt_enc == 0) {
                    rdx_pt_enc = 1;
                    if ((idx.wrapping_add(1)) >= (s.len() as i64)) {
                        right_radix_leading_zeros = ((101 as i64).wrapping_sub(right_radix_leading_zeros));
                        if (right_radix_leading_zeros < 0) {
                            right_radix_leading_zeros = 0;
                        }
                        res = (((go_checked_shl_u64((right_radix_leading_zeros as u64), go_shift_count_u64((23) as u64)))) | sign_x);
                        return ((res as u32), 0);
                    }
                    idx = idx.wrapping_add(1);
                } else {
                    return (((0x7c000000 | sign_x) as u32), 0);
                }
            } else if (idx >= (s.len() as i64)) {
                right_radix_leading_zeros = ((101 as i64).wrapping_sub(right_radix_leading_zeros));
                if (right_radix_leading_zeros < 0) {
                    right_radix_leading_zeros = 0;
                }
                res = (((go_checked_shl_u64((right_radix_leading_zeros as u64), go_shift_count_u64((23) as u64)))) | sign_x);
                return ((res as u32), 0);
            }
        }
    }
    if (idx >= (s.len() as i64)) {
        right_radix_leading_zeros = ((101 as i64).wrapping_sub(right_radix_leading_zeros));
        if (right_radix_leading_zeros < 0) {
            right_radix_leading_zeros = 0;
        }
        res = (((go_checked_shl_u64((right_radix_leading_zeros as u64), go_shift_count_u64((23) as u64)))) | sign_x);
        return ((res as u32), 0);
    }
    c = s[idx as usize];
    ndigits = 0;
    while ((idx < (s.len() as i64)) && (((((c >= b'0') && (c <= b'9'))) || (c == b'.')))) {
        if (c == b'.') {
            if (rdx_pt_enc != 0) {
                return (((0x7c000000 | sign_x) as u32), 0);
            }
            rdx_pt_enc = 1;
            idx = idx.wrapping_add(1);
            if (idx < (s.len() as i64)) {
                c = s[idx as usize];
            }
            continue;
        }
        dec_expon_scale = dec_expon_scale.wrapping_add(rdx_pt_enc);
        ndigits = ndigits.wrapping_add(1);
        if (ndigits <= 7) {
            coefficient_x = (((go_checked_shl_u64(coefficient_x, go_shift_count_u64((1) as u64)))).wrapping_add(((go_checked_shl_u64(coefficient_x, go_shift_count_u64((3) as u64))))));
            coefficient_x = coefficient_x.wrapping_add(((c.wrapping_sub(b'0')) as u64));
        } else if (ndigits == 8) {
            match rnd_mode {
                0 => {
                    if ((c == b'5') && ((coefficient_x & 1) == 0)) {
                        midpoint = 1;
                    }
                    if ((c > b'5') || (((c == b'5') && ((coefficient_x & 1) != 0)))) {
                        coefficient_x = coefficient_x.wrapping_add(1);
                        rounded_up = 1;
                    }
                }
                1 => {
                    if (sign_x != 0) {
                        if (c > b'0') {
                            coefficient_x = coefficient_x.wrapping_add(1);
                            rounded_up = 1;
                        } else {
                            dround = 1;
                        }
                    }
                }
                2 => {
                    if (sign_x == 0) {
                        if (c > b'0') {
                            coefficient_x = coefficient_x.wrapping_add(1);
                            rounded_up = 1;
                        } else {
                            dround = 1;
                        }
                    }
                }
                4 => {
                    if (c >= b'5') {
                        coefficient_x = coefficient_x.wrapping_add(1);
                        rounded_up = 1;
                    }
                }
                _ => {}
            }
            if (coefficient_x == 10000000) {
                coefficient_x = 1000000;
                add_expon = 1;
            }
            if (c > b'0') {
                rounded = 1;
            }
            add_expon = add_expon.wrapping_add(1);
        } else {
            add_expon = add_expon.wrapping_add(1);
            if ((midpoint != 0) && (c > b'0')) {
                coefficient_x = coefficient_x.wrapping_add(1);
                midpoint = 0;
                rounded_up = 1;
            }
            if (c > b'0') {
                rounded = 1;
                if (dround != 0) {
                    dround = 0;
                    coefficient_x = coefficient_x.wrapping_add(1);
                    rounded_up = 1;
                    if (coefficient_x == 10000000) {
                        coefficient_x = 1000000;
                        add_expon = add_expon.wrapping_add(1);
                    }
                }
            }
        }
        idx = idx.wrapping_add(1);
        if (idx < (s.len() as i64)) {
            c = s[idx as usize];
        } else {
            c = 0;
        }
    }
    add_expon = add_expon.wrapping_sub(((dec_expon_scale.wrapping_add(right_radix_leading_zeros))));
    if (idx >= (s.len() as i64)) {
        if (rounded != 0) {
            pfpsf |= 32;
        }
        res = (get_bid32_flags((sign_x as u32), (add_expon.wrapping_add(101)), coefficient_x, rnd_mode, (&mut pfpsf)) as u64);
        return ((res as u32), pfpsf);
    }
    c = s[idx as usize];
    if ((c != b'E') && (c != b'e')) {
        return (((0x7c000000 | sign_x) as u32), 0);
    }
    idx = idx.wrapping_add(1);
    if (idx >= (s.len() as i64)) {
        return (((0x7c000000 | sign_x) as u32), 0);
    }
    c = s[idx as usize];
    if (c == b'-') {
        sgn_expon = 1;
    }
    if ((c == b'-') || (c == b'+')) {
        idx = idx.wrapping_add(1);
        if (idx >= (s.len() as i64)) {
            return (((0x7c000000 | sign_x) as u32), 0);
        }
        c = s[idx as usize];
    }
    if ((c < b'0') || (c > b'9')) {
        return (((0x7c000000 | sign_x) as u32), 0);
    }
    while (((idx < (s.len() as i64)) && (s[idx as usize] >= b'0')) && (s[idx as usize] <= b'9')) {
        if (expon_x < (1 << 20)) {
            expon_x = (((go_checked_shl_i64(expon_x, go_shift_count_u64((1) as u64)))).wrapping_add(((go_checked_shl_i64(expon_x, go_shift_count_u64((3) as u64))))));
            expon_x = expon_x.wrapping_add(((s[idx as usize].wrapping_sub(b'0')) as i64));
        }
        idx = idx.wrapping_add(1);
    }
    if (idx < (s.len() as i64)) {
        return (((0x7c000000 | sign_x) as u32), 0);
    }
    if (rounded != 0) {
        pfpsf |= 32;
    }
    if (sgn_expon != 0) {
        expon_x = (expon_x.wrapping_neg());
    }
    expon_x = expon_x.wrapping_add((add_expon.wrapping_add(101)));
    if (expon_x < 0) {
        if (rounded_up != 0) {
            coefficient_x = coefficient_x.wrapping_sub(1);
        }
        rnd_mode = 0;
        res = (get_bid32_uf((sign_x as u32), expon_x, coefficient_x, (rounded as u32), rnd_mode, (&mut pfpsf)) as u64);
        return ((res as u32), pfpsf);
    }
    res = (get_bid32_flags((sign_x as u32), expon_x, coefficient_x, rnd_mode, (&mut pfpsf)) as u64);
    return ((res as u32), pfpsf);
}
