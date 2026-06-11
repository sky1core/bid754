// Auto-generated from bid128_div.go by go2rs. Do not edit.

use super::prelude::*;

pub(crate) fn __mul_128x128_low(mut A: BID_UINT128, mut B: BID_UINT128) -> BID_UINT128 {
    let mut Ql: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut ALBL = __mul_64x64_to_128(A.w[0], B.w[0]);
    let mut QM64 = ((B.w[0].wrapping_mul(A.w[1])).wrapping_add((A.w[0].wrapping_mul(B.w[1]))));
    Ql.w[0] = ALBL.w[0];
    Ql.w[1] = (QM64.wrapping_add(ALBL.w[1]));
    return Ql;
}

pub(crate) fn __sub_borrow_in_out(mut x: u64, mut y: u64, mut ci: u64) -> (u64, u64) {
    let mut s: u64 = 0;
    let mut co: u64 = 0;
    let mut x1 = (x.wrapping_sub(ci));
    if (x1 > x) {
        co = 1;
    }
    s = (x1.wrapping_sub(y));
    if (s > x1) {
        co = 1;
    }
    return (s, co);
}

pub(crate) fn bid___div_128_by_128(mut CX0: BID_UINT128, mut CY: BID_UINT128) -> (BID_UINT128, BID_UINT128) {
    let mut CQ: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut CR: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut CY36: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut CY51: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut A2: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut CQT: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut Q: u64 = 0;
    if ((CX0.w[1] == 0) && (CY.w[1] == 0)) {
        CQ.w[0] = (CX0.w[0] / CY.w[0]);
        CQ.w[1] = 0;
        CR.w[1] = 0;
        CR.w[0] = (CX0.w[0].wrapping_sub((CQ.w[0].wrapping_mul(CY.w[0]))));
        return (CQ, CR);
    }
    let mut CX = CX0;
    let mut t64 = f64::from_bits(0x43f0000000000000);
    let mut lx = no_fma_mul_add_f64((CX.w[1] as f64), t64, (CX.w[0] as f64));
    let mut ly = no_fma_mul_add_f64((CY.w[1] as f64), t64, (CY.w[0] as f64));
    let mut lq = (lx / ly);
    CY36.w[1] = (go_checked_shr_u64(CY.w[0], go_shift_count_u64((64 - 36) as u64)));
    CY36.w[0] = (go_checked_shl_u64(CY.w[0], go_shift_count_u64((36) as u64)));
    CQ.w[1] = 0;
    CQ.w[0] = 0;
    if (((CY.w[1] == 0) && (CY36.w[1] == 0)) && (CX.w[1] >= CY36.w[0])) {
        let mut d60 = f64::from_bits(0x3c30000000000000);
        lq *= d60;
        Q = ((lq as u64).wrapping_sub(4));
        A2 = __mul_64x64_to_128(Q, CY.w[0]);
        A2.w[1] = (((go_checked_shl_u64(A2.w[1], go_shift_count_u64((60) as u64)))) | ((go_checked_shr_u64(A2.w[0], go_shift_count_u64((64 - 60) as u64)))));
        A2.w[0] = go_checked_shl_u64(A2.w[0], go_shift_count_u64((60) as u64));
        CX = __sub_128_128(CX, A2);
        lx = no_fma_mul_add_f64((CX.w[1] as f64), t64, (CX.w[0] as f64));
        lq = (lx / ly);
        CQ.w[1] = (go_checked_shr_u64(Q, go_shift_count_u64((64 - 60) as u64)));
        CQ.w[0] = (go_checked_shl_u64(Q, go_shift_count_u64((60) as u64)));
    }
    CY51.w[1] = (((go_checked_shl_u64(CY.w[1], go_shift_count_u64((51) as u64)))) | ((go_checked_shr_u64(CY.w[0], go_shift_count_u64((64 - 51) as u64)))));
    CY51.w[0] = (go_checked_shl_u64(CY.w[0], go_shift_count_u64((51) as u64)));
    if ((CY.w[1] < ((1 << (64 - 51)))) && __unsigned_compare_gt_128(CX, CY51)) {
        let mut d49 = f64::from_bits(0x3ce0000000000000);
        lq *= d49;
        Q = ((lq as u64).wrapping_sub(1));
        A2 = __mul_64x64_to_128(Q, CY.w[0]);
        A2.w[1] = A2.w[1].wrapping_add((Q.wrapping_mul(CY.w[1])));
        A2.w[1] = (((go_checked_shl_u64(A2.w[1], go_shift_count_u64((49) as u64)))) | ((go_checked_shr_u64(A2.w[0], go_shift_count_u64((64 - 49) as u64)))));
        A2.w[0] = go_checked_shl_u64(A2.w[0], go_shift_count_u64((49) as u64));
        CX = __sub_128_128(CX, A2);
        CQT.w[1] = (go_checked_shr_u64(Q, go_shift_count_u64((64 - 49) as u64)));
        CQT.w[0] = (go_checked_shl_u64(Q, go_shift_count_u64((49) as u64)));
        CQ = __add_128_128(CQ, CQT);
        lx = no_fma_mul_add_f64((CX.w[1] as f64), t64, (CX.w[0] as f64));
        lq = (lx / ly);
    }
    Q = (lq as u64);
    A2 = __mul_64x64_to_128(Q, CY.w[0]);
    A2.w[1] = A2.w[1].wrapping_add((Q.wrapping_mul(CY.w[1])));
    CX = __sub_128_128(CX, A2);
    if ((CX.w[1] as i64) < 0) {
        Q = Q.wrapping_sub(1);
        CX.w[0] = CX.w[0].wrapping_add(CY.w[0]);
        if (CX.w[0] < CY.w[0]) {
            CX.w[1] = CX.w[1].wrapping_add(1);
        }
        CX.w[1] = CX.w[1].wrapping_add(CY.w[1]);
        if ((CX.w[1] as i64) < 0) {
            Q = Q.wrapping_sub(1);
            CX.w[0] = CX.w[0].wrapping_add(CY.w[0]);
            if (CX.w[0] < CY.w[0]) {
                CX.w[1] = CX.w[1].wrapping_add(1);
            }
            CX.w[1] = CX.w[1].wrapping_add(CY.w[1]);
        }
    } else if __unsigned_compare_ge_128(CX, CY) {
        Q = Q.wrapping_add(1);
        CX = __sub_128_128(CX, CY);
    }
    CQ = __add_128_64(CQ, Q);
    CR.w[1] = CX.w[1];
    CR.w[0] = CX.w[0];
    return (CQ, CR);
}

pub(crate) fn bid___div_256_by_128(pCQ: &mut BID_UINT128, pCA4: &mut BID_UINT256, mut CY: BID_UINT128) {
    let mut CA4: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut CA2: [u64; 3] = [0; 3];
    let mut CQ: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut A2: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut A2h: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut CQT: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut Q: u64 = 0;
    let mut carry64: u64 = 0;
    CA4.w[3] = pCA4.w[3];
    CA4.w[2] = pCA4.w[2];
    CA4.w[1] = pCA4.w[1];
    CA4.w[0] = pCA4.w[0];
    CQ.w[1] = pCQ.w[1];
    CQ.w[0] = pCQ.w[0];
    let mut t64 = f64::from_bits(0x43f0000000000000);
    let mut d128 = (t64 * t64);
    let mut d192 = (d128 * t64);
    let mut lx = no_fma_mul_add_f64((CA4.w[3] as f64), d192, no_fma_mul_add_f64((CA4.w[2] as f64), d128, no_fma_mul_add_f64((CA4.w[1] as f64), t64, (CA4.w[0] as f64))));
    let mut ly = no_fma_mul_add_f64((CY.w[1] as f64), t64, (CY.w[0] as f64));
    let mut lq = (lx / ly);
    let mut CY36_2: u64 = 0;
    let mut CY36_1: u64 = 0;
    let mut CY36_0: u64 = 0;
    CY36_2 = (go_checked_shr_u64(CY.w[1], go_shift_count_u64((64 - 36) as u64)));
    CY36_1 = (((go_checked_shl_u64(CY.w[1], go_shift_count_u64((36) as u64)))) | ((go_checked_shr_u64(CY.w[0], go_shift_count_u64((64 - 36) as u64)))));
    CY36_0 = (go_checked_shl_u64(CY.w[0], go_shift_count_u64((36) as u64)));
    if ((CA4.w[3] > CY36_2) || (((CA4.w[3] == CY36_2) && (((CA4.w[2] > CY36_1) || (((CA4.w[2] == CY36_1) && (CA4.w[1] >= CY36_0)))))))) {
        let mut d60 = f64::from_bits(0x3c30000000000000);
        lq *= d60;
        Q = ((lq as u64).wrapping_sub(4));
        let mut tmp192 = __mul_64x128_to_192(Q, CY);
        CA2[2] = (((go_checked_shl_u64(tmp192.w[2], go_shift_count_u64((60) as u64)))) | ((go_checked_shr_u64(tmp192.w[1], go_shift_count_u64((64 - 60) as u64)))));
        CA2[1] = (((go_checked_shl_u64(tmp192.w[1], go_shift_count_u64((60) as u64)))) | ((go_checked_shr_u64(tmp192.w[0], go_shift_count_u64((64 - 60) as u64)))));
        CA2[0] = (go_checked_shl_u64(tmp192.w[0], go_shift_count_u64((60) as u64)));
        (CA4.w[0], carry64) = __sub_borrow_out(CA4.w[0], CA2[0]);
        (CA4.w[1], carry64) = __sub_borrow_in_out(CA4.w[1], CA2[1], carry64);
        CA4.w[2] = ((CA4.w[2].wrapping_sub(CA2[2])).wrapping_sub(carry64));
        lx = no_fma_mul_add_f64((CA4.w[2] as f64), d128, no_fma_mul_add_f64((CA4.w[1] as f64), t64, (CA4.w[0] as f64)));
        lq = (lx / ly);
        CQT.w[1] = (go_checked_shr_u64(Q, go_shift_count_u64((64 - 60) as u64)));
        CQT.w[0] = (go_checked_shl_u64(Q, go_shift_count_u64((60) as u64)));
        CQ = __add_128_128(CQ, CQT);
    }
    let mut CY51_2: u64 = 0;
    let mut CY51_1: u64 = 0;
    let mut CY51_0: u64 = 0;
    CY51_2 = (go_checked_shr_u64(CY.w[1], go_shift_count_u64((64 - 51) as u64)));
    CY51_1 = (((go_checked_shl_u64(CY.w[1], go_shift_count_u64((51) as u64)))) | ((go_checked_shr_u64(CY.w[0], go_shift_count_u64((64 - 51) as u64)))));
    CY51_0 = (go_checked_shl_u64(CY.w[0], go_shift_count_u64((51) as u64)));
    let mut ca4_128 = BID_UINT128 { w: [CA4.w[0], CA4.w[1]], ..Default::default() };
    let mut cy51_128 = BID_UINT128 { w: [CY51_0, CY51_1], ..Default::default() };
    if ((CA4.w[2] > CY51_2) || (((CA4.w[2] == CY51_2) && __unsigned_compare_gt_128(ca4_128, cy51_128)))) {
        let mut d49 = f64::from_bits(0x3ce0000000000000);
        lq *= d49;
        Q = ((lq as u64).wrapping_sub(1));
        A2 = __mul_64x64_to_128(Q, CY.w[0]);
        A2h = __mul_64x64_to_128(Q, CY.w[1]);
        A2.w[1] = A2.w[1].wrapping_add(A2h.w[0]);
        if (A2.w[1] < A2h.w[0]) {
            A2h.w[1] = A2h.w[1].wrapping_add(1);
        }
        CA2[2] = (((go_checked_shl_u64(A2h.w[1], go_shift_count_u64((49) as u64)))) | ((go_checked_shr_u64(A2.w[1], go_shift_count_u64((64 - 49) as u64)))));
        CA2[1] = (((go_checked_shl_u64(A2.w[1], go_shift_count_u64((49) as u64)))) | ((go_checked_shr_u64(A2.w[0], go_shift_count_u64((64 - 49) as u64)))));
        CA2[0] = (go_checked_shl_u64(A2.w[0], go_shift_count_u64((49) as u64)));
        (CA4.w[0], carry64) = __sub_borrow_out(CA4.w[0], CA2[0]);
        (CA4.w[1], carry64) = __sub_borrow_in_out(CA4.w[1], CA2[1], carry64);
        CA4.w[2] = ((CA4.w[2].wrapping_sub(CA2[2])).wrapping_sub(carry64));
        CQT.w[1] = (go_checked_shr_u64(Q, go_shift_count_u64((64 - 49) as u64)));
        CQT.w[0] = (go_checked_shl_u64(Q, go_shift_count_u64((49) as u64)));
        CQ = __add_128_128(CQ, CQT);
        lx = no_fma_mul_add_f64((CA4.w[2] as f64), d128, no_fma_mul_add_f64((CA4.w[1] as f64), t64, (CA4.w[0] as f64)));
        lq = (lx / ly);
    }
    Q = (lq as u64);
    A2 = __mul_64x64_to_128(Q, CY.w[0]);
    A2.w[1] = A2.w[1].wrapping_add((Q.wrapping_mul(CY.w[1])));
    let mut tmpCA = BID_UINT128 { w: [CA4.w[0], CA4.w[1]], ..Default::default() };
    tmpCA = __sub_128_128(tmpCA, A2);
    CA4.w[0] = tmpCA.w[0];
    CA4.w[1] = tmpCA.w[1];
    if ((CA4.w[1] as i64) < 0) {
        Q = Q.wrapping_sub(1);
        CA4.w[0] = CA4.w[0].wrapping_add(CY.w[0]);
        if (CA4.w[0] < CY.w[0]) {
            CA4.w[1] = CA4.w[1].wrapping_add(1);
        }
        CA4.w[1] = CA4.w[1].wrapping_add(CY.w[1]);
        if ((CA4.w[1] as i64) < 0) {
            Q = Q.wrapping_sub(1);
            CA4.w[0] = CA4.w[0].wrapping_add(CY.w[0]);
            if (CA4.w[0] < CY.w[0]) {
                CA4.w[1] = CA4.w[1].wrapping_add(1);
            }
            CA4.w[1] = CA4.w[1].wrapping_add(CY.w[1]);
        }
    } else if ((CA4.w[1] > CY.w[1]) || (((CA4.w[1] == CY.w[1]) && (CA4.w[0] >= CY.w[0])))) {
        Q = Q.wrapping_add(1);
        let mut tmpCA2 = BID_UINT128 { w: [CA4.w[0], CA4.w[1]], ..Default::default() };
        tmpCA2 = __sub_128_128(tmpCA2, CY);
        CA4.w[0] = tmpCA2.w[0];
        CA4.w[1] = tmpCA2.w[1];
    }
    CQ = __add_128_64(CQ, Q);
    pCQ.w[1] = CQ.w[1];
    pCQ.w[0] = CQ.w[0];
    pCA4.w[1] = CA4.w[1];
    pCA4.w[0] = CA4.w[0];
}

pub(crate) fn handle_uf_128(mut sgn: u64, mut expon: i64, mut CQ: BID_UINT128, mut prounding_mode: i64, fpsc: &mut u32) -> BID_UINT128 {
    let mut res: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut T128: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut TP128: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut Qh: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut Ql: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut Qh1: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut Stemp: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut Tmp: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut Tmp1: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut carry: u64 = 0;
    let mut CY: u64 = 0;
    let mut ed2: i64 = 0;
    let mut amount: i64 = 0;
    let mut rmode: u64 = 0;
    let mut status: u32 = 0;
    if ((expon.wrapping_add(34)) < 0) {
        (*fpsc) |= (16 | 32);
        res.w[1] = sgn;
        res.w[0] = 0;
        if ((((sgn != 0) && (prounding_mode == 1))) || (((sgn == 0) && (prounding_mode == 2)))) {
            res.w[0] = 1;
        }
        return res;
    }
    ed2 = ((0 as i64).wrapping_sub(expon));
    rmode = (prounding_mode as u64);
    if ((sgn != 0) && (((rmode.wrapping_sub(1)) as u64) < 2)) {
        rmode = ((3 as u64).wrapping_sub(rmode));
    }
    T128 = bid_round_const_table_128[rmode as usize][ed2 as usize];
    (CQ.w[0], carry) = __add_carry_out(T128.w[0], CQ.w[0]);
    CQ.w[1] = ((CQ.w[1].wrapping_add(T128.w[1])).wrapping_add(carry));
    TP128 = bid_reciprocals10_128[ed2 as usize];
    (Qh, Ql) = __mul_128x128_full(CQ, TP128);
    amount = (bid_recip_scale[ed2 as usize] as i64);
    if (amount >= 64) {
        CQ.w[0] = (go_checked_shr_u64(Qh.w[1], go_shift_count_u64((((amount.wrapping_sub(64)) as u64)) as u64)));
        CQ.w[1] = 0;
    } else {
        CQ = __shr_128(Qh, (amount as u64));
    }
    expon = 0;
    if (prounding_mode == 0) {
        if ((CQ.w[0] & 1) != 0) {
            Qh1 = __shl_128_long(Qh, (((128 as i64).wrapping_sub(amount)) as u64));
            if (((Qh1.w[1] == 0) && (Qh1.w[0] == 0)) && (((Ql.w[1] < bid_reciprocals10_128[ed2 as usize].w[1]) || (((Ql.w[1] == bid_reciprocals10_128[ed2 as usize].w[1]) && (Ql.w[0] < bid_reciprocals10_128[ed2 as usize].w[0])))))) {
                CQ.w[0] = CQ.w[0].wrapping_sub(1);
            }
        }
    }
    if ((((*fpsc) & 32)) != 0) {
        (*fpsc) |= 16;
    } else {
        status = 32;
        Qh1 = __shl_128_long(Qh, (((128 as i64).wrapping_sub(amount)) as u64));
        match rmode {
            0 | 4 => {
                if (((Qh1.w[1] == 0x8000000000000000) && (Qh1.w[0] == 0)) && (((Ql.w[1] < bid_reciprocals10_128[ed2 as usize].w[1]) || (((Ql.w[1] == bid_reciprocals10_128[ed2 as usize].w[1]) && (Ql.w[0] < bid_reciprocals10_128[ed2 as usize].w[0])))))) {
                    status = 0;
                }
            }
            1 | 3 => {
                if (((Qh1.w[1] == 0) && (Qh1.w[0] == 0)) && (((Ql.w[1] < bid_reciprocals10_128[ed2 as usize].w[1]) || (((Ql.w[1] == bid_reciprocals10_128[ed2 as usize].w[1]) && (Ql.w[0] < bid_reciprocals10_128[ed2 as usize].w[0])))))) {
                    status = 0;
                }
            }
            _ => {
                (Stemp.w[0], CY) = __add_carry_out(Ql.w[0], bid_reciprocals10_128[ed2 as usize].w[0]);
                (Stemp.w[1], carry) = __add_carry_in_out(Ql.w[1], bid_reciprocals10_128[ed2 as usize].w[1], CY);
                _ = Stemp;
                Qh = __shr_128_long(Qh1, (((128 as i64).wrapping_sub(amount)) as u64));
                Tmp.w[0] = 1;
                Tmp.w[1] = 0;
                Tmp1 = __shl_128_long(Tmp, (amount as u64));
                Qh.w[0] = Qh.w[0].wrapping_add(carry);
                if (Qh.w[0] < carry) {
                    Qh.w[1] = Qh.w[1].wrapping_add(1);
                }
                if __unsigned_compare_ge_128(Qh, Tmp1) {
                    status = 0;
                }
            }
        }
        if (status != 0) {
            (*fpsc) |= (16 | status);
        }
    }
    res.w[1] = (sgn | CQ.w[1]);
    res.w[0] = CQ.w[0];
    return res;
}

pub(crate) fn bid_handle_uf_128_rem(mut sgn: u64, mut expon: i64, mut CQ: BID_UINT128, mut R: u64, mut prounding_mode: i64, fpsc: &mut u32) -> BID_UINT128 {
    let mut res: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut T128: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut TP128: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut Qh: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut Ql: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut Qh1: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut Stemp: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut Tmp: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut Tmp1: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut CQ2: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut CQ8: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut carry: u64 = 0;
    let mut CY: u64 = 0;
    let mut ed2: i64 = 0;
    let mut amount: i64 = 0;
    let mut rmode: u64 = 0;
    let mut status: u32 = 0;
    if ((expon.wrapping_add(34)) < 0) {
        (*fpsc) |= (16 | 32);
        res.w[1] = sgn;
        res.w[0] = 0;
        if ((((sgn != 0) && (prounding_mode == 1))) || (((sgn == 0) && (prounding_mode == 2)))) {
            res.w[0] = 1;
        }
        return res;
    }
    CQ2.w[1] = (((go_checked_shl_u64(CQ.w[1], go_shift_count_u64((1) as u64)))) | ((go_checked_shr_u64(CQ.w[0], go_shift_count_u64((63) as u64)))));
    CQ2.w[0] = (go_checked_shl_u64(CQ.w[0], go_shift_count_u64((1) as u64)));
    CQ8.w[1] = (((go_checked_shl_u64(CQ.w[1], go_shift_count_u64((3) as u64)))) | ((go_checked_shr_u64(CQ.w[0], go_shift_count_u64((61) as u64)))));
    CQ8.w[0] = (go_checked_shl_u64(CQ.w[0], go_shift_count_u64((3) as u64)));
    CQ = __add_128_128(CQ2, CQ8);
    if (R != 0) {
        CQ.w[0] |= 1;
    }
    ed2 = ((1 as i64).wrapping_sub(expon));
    rmode = (prounding_mode as u64);
    if ((sgn != 0) && (((rmode.wrapping_sub(1)) as u64) < 2)) {
        rmode = ((3 as u64).wrapping_sub(rmode));
    }
    T128 = bid_round_const_table_128[rmode as usize][ed2 as usize];
    (CQ.w[0], carry) = __add_carry_out(T128.w[0], CQ.w[0]);
    CQ.w[1] = ((CQ.w[1].wrapping_add(T128.w[1])).wrapping_add(carry));
    TP128 = bid_reciprocals10_128[ed2 as usize];
    (Qh, Ql) = __mul_128x128_full(CQ, TP128);
    amount = (bid_recip_scale[ed2 as usize] as i64);
    if (amount >= 64) {
        CQ.w[0] = (go_checked_shr_u64(Qh.w[1], go_shift_count_u64((((amount.wrapping_sub(64)) as u64)) as u64)));
        CQ.w[1] = 0;
    } else {
        CQ = __shr_128(Qh, (amount as u64));
    }
    expon = 0;
    if (prounding_mode == 0) {
        if ((CQ.w[0] & 1) != 0) {
            Qh1 = __shl_128_long(Qh, (((128 as i64).wrapping_sub(amount)) as u64));
            if (((Qh1.w[1] == 0) && (Qh1.w[0] == 0)) && (((Ql.w[1] < bid_reciprocals10_128[ed2 as usize].w[1]) || (((Ql.w[1] == bid_reciprocals10_128[ed2 as usize].w[1]) && (Ql.w[0] < bid_reciprocals10_128[ed2 as usize].w[0])))))) {
                CQ.w[0] = CQ.w[0].wrapping_sub(1);
            }
        }
    }
    if ((((*fpsc) & 32)) != 0) {
        (*fpsc) |= 16;
    } else {
        status = 32;
        Qh1 = __shl_128_long(Qh, (((128 as i64).wrapping_sub(amount)) as u64));
        match rmode {
            0 | 4 => {
                if (((Qh1.w[1] == 0x8000000000000000) && (Qh1.w[0] == 0)) && (((Ql.w[1] < bid_reciprocals10_128[ed2 as usize].w[1]) || (((Ql.w[1] == bid_reciprocals10_128[ed2 as usize].w[1]) && (Ql.w[0] < bid_reciprocals10_128[ed2 as usize].w[0])))))) {
                    status = 0;
                }
            }
            1 | 3 => {
                if (((Qh1.w[1] == 0) && (Qh1.w[0] == 0)) && (((Ql.w[1] < bid_reciprocals10_128[ed2 as usize].w[1]) || (((Ql.w[1] == bid_reciprocals10_128[ed2 as usize].w[1]) && (Ql.w[0] < bid_reciprocals10_128[ed2 as usize].w[0])))))) {
                    status = 0;
                }
            }
            _ => {
                (Stemp.w[0], CY) = __add_carry_out(Ql.w[0], bid_reciprocals10_128[ed2 as usize].w[0]);
                (Stemp.w[1], carry) = __add_carry_in_out(Ql.w[1], bid_reciprocals10_128[ed2 as usize].w[1], CY);
                _ = Stemp;
                Qh = __shr_128_long(Qh1, (((128 as i64).wrapping_sub(amount)) as u64));
                Tmp.w[0] = 1;
                Tmp.w[1] = 0;
                Tmp1 = __shl_128_long(Tmp, (amount as u64));
                Qh.w[0] = Qh.w[0].wrapping_add(carry);
                if (Qh.w[0] < carry) {
                    Qh.w[1] = Qh.w[1].wrapping_add(1);
                }
                if __unsigned_compare_ge_128(Qh, Tmp1) {
                    status = 0;
                }
            }
        }
        if (status != 0) {
            (*fpsc) |= (16 | status);
        }
    }
    res.w[1] = (sgn | CQ.w[1]);
    res.w[0] = CQ.w[0];
    return res;
}

pub(crate) fn bid_get_bid128(mut sgn: u64, mut expon: i64, mut coeff: BID_UINT128, mut prounding_mode: i64, fpsc: &mut u32) -> BID_UINT128 {
    let mut res: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut T: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut tmp: u64 = 0;
    let mut tmp2: u64 = 0;
    if ((coeff.w[1] == 0x0001ed09bead87c0) && (coeff.w[0] == 0x378d8e6400000000)) {
        expon = expon.wrapping_add(1);
        coeff.w[1] = 0x0000314dc6448d93;
        coeff.w[0] = 0x38c15b0a00000000;
    }
    if ((expon < 0) || (expon > 0x2fff)) {
        if (expon < 0) {
            return handle_uf_128(sgn, expon, coeff, prounding_mode, fpsc);
        }
        if ((expon.wrapping_sub(34)) <= 0x2fff) {
            T = bid_power10_table_128[(34 - 1) as usize];
            while (__unsigned_compare_gt_128(T, coeff) && (expon > 0x2fff)) {
                coeff.w[1] = (((((go_checked_shl_u64(coeff.w[1], go_shift_count_u64((3) as u64)))).wrapping_add(((go_checked_shl_u64(coeff.w[1], go_shift_count_u64((1) as u64)))))).wrapping_add(((go_checked_shr_u64(coeff.w[0], go_shift_count_u64((61) as u64)))))).wrapping_add(((go_checked_shr_u64(coeff.w[0], go_shift_count_u64((63) as u64))))));
                tmp2 = (go_checked_shl_u64(coeff.w[0], go_shift_count_u64((3) as u64)));
                coeff.w[0] = (((go_checked_shl_u64(coeff.w[0], go_shift_count_u64((1) as u64)))).wrapping_add(tmp2));
                if (coeff.w[0] < tmp2) {
                    coeff.w[1] = coeff.w[1].wrapping_add(1);
                }
                expon = expon.wrapping_sub(1);
            }
        }
        if (expon > 0x2fff) {
            if ((coeff.w[1] == 0) && (coeff.w[0] == 0)) {
                res.w[1] = (sgn | (((0x2fff as u64) << 49)));
                res.w[0] = 0;
                return res;
            }
            (*fpsc) |= (8 | 32);
            if (((prounding_mode == 3) || (((sgn != 0) && (prounding_mode == 2)))) || (((sgn == 0) && (prounding_mode == 1)))) {
                res.w[1] = (sgn | 0x5fffed09bead87c0);
                res.w[0] = 0x378d8e63ffffffff;
            } else {
                res.w[1] = (sgn | 0x7800000000000000);
                res.w[0] = 0;
            }
            return res;
        }
    }
    res.w[0] = coeff.w[0];
    tmp = (expon as u64);
    tmp = go_checked_shl_u64(tmp, go_shift_count_u64((49) as u64));
    res.w[1] = ((sgn | tmp) | coeff.w[1]);
    return res;
}

pub fn bid128_div(mut x: BID_UINT128, mut y: BID_UINT128, mut rnd_mode: i64) -> (BID_UINT128, u32) {
    let mut CA4: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut CA4r: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut P256: BID_UINT256 = BID_UINT256 { w: [0, 0, 0, 0] };
    let mut CX: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut CY: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut T128: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut CQ: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut CR: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut CA: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut TP128: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut Qh: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut res: BID_UINT128 = BID_UINT128 { w: [0, 0] };
    let mut sign_x: u64 = 0;
    let mut sign_y: u64 = 0;
    let mut T: u64 = 0;
    let mut carry64: u64 = 0;
    let mut D: u64 = 0;
    let mut Q_high: u64 = 0;
    let mut Q_low: u64 = 0;
    let mut QX: u64 = 0;
    let mut PD: u64 = 0;
    let mut valid_y: bool = false;
    let mut QX32: u32 = 0;
    let mut digit: u32 = 0;
    let mut digit_h: u32 = 0;
    let mut digit_low: u32 = 0;
    let mut tdigit: [u32; 3] = [0; 3];
    let mut exponent_x: i64 = 0;
    let mut exponent_y: i64 = 0;
    let mut bin_index: i64 = 0;
    let mut bin_expon: i64 = 0;
    let mut diff_expon: i64 = 0;
    let mut ed2: i64 = 0;
    let mut digits_q: i64 = 0;
    let mut amount: i64 = 0;
    let mut nzeros: i64 = 0;
    let mut i: i64 = 0;
    let mut j: i64 = 0;
    let mut k: i64 = 0;
    let mut d5: i64 = 0;
    let mut rmode: u64 = 0;
    let mut pfpsf: u32 = 0;
    (sign_y, exponent_y, CY, valid_y) = unpack_bid128_value(y);
    _ = valid_y;
    let (mut sign_x_raw, mut exponent_x_raw, mut CX_raw, mut valid_x) = unpack_bid128_value(x);
    sign_x = sign_x_raw;
    exponent_x = exponent_x_raw;
    CX = CX_raw;
    if (!valid_x) {
        if ((x.w[1] & 0x7c00000000000000) == 0x7c00000000000000) {
            if (((x.w[1] & 0x7e00000000000000) == 0x7e00000000000000) || ((y.w[1] & 0x7e00000000000000) == 0x7e00000000000000)) {
                pfpsf |= 1;
            }
            res.w[1] = ((CX.w[1]) & 0xfdffffffffffffff);
            res.w[0] = CX.w[0];
            return (res, pfpsf);
        }
        if ((x.w[1] & 0x7800000000000000) == 0x7800000000000000) {
            if ((y.w[1] & 0x7c00000000000000) == 0x7800000000000000) {
                pfpsf |= 1;
                res.w[1] = 0x7c00000000000000;
                res.w[0] = 0;
                return (res, pfpsf);
            }
            if ((y.w[1] & 0x7c00000000000000) != 0x7c00000000000000) {
                res.w[1] = ((((x.w[1] ^ y.w[1]) & 0x8000000000000000)) | 0x7800000000000000);
                res.w[0] = 0;
                return (res, pfpsf);
            }
        }
        if ((y.w[1] & 0x7800000000000000) < 0x7800000000000000) {
            if ((CY.w[0] == 0) && ((CY.w[1] & 0x0001ffffffffffff) == 0)) {
                pfpsf |= 1;
                res.w[1] = 0x7c00000000000000;
                res.w[0] = 0;
                return (res, pfpsf);
            }
            res.w[1] = ((x.w[1] ^ y.w[1]) & 0x8000000000000000);
            exponent_x = ((exponent_x.wrapping_sub(exponent_y)).wrapping_add(0x1820));
            if (exponent_x > 0x2fff) {
                exponent_x = 0x2fff;
            } else if (exponent_x < 0) {
                exponent_x = 0;
            }
            res.w[1] |= (go_checked_shl_u64((exponent_x as u64), go_shift_count_u64((49) as u64)));
            res.w[0] = 0;
            return (res, pfpsf);
        }
    }
    if (!valid_y) {
        if ((y.w[1] & 0x7c00000000000000) == 0x7c00000000000000) {
            if ((y.w[1] & 0x7e00000000000000) == 0x7e00000000000000) {
                pfpsf |= 1;
            }
            res.w[1] = (CY.w[1] & 0xfdffffffffffffff);
            res.w[0] = CY.w[0];
            return (res, pfpsf);
        }
        if ((y.w[1] & 0x7800000000000000) == 0x7800000000000000) {
            res.w[1] = (sign_x ^ sign_y);
            res.w[0] = 0;
            return (res, pfpsf);
        }
        pfpsf |= 4;
        res.w[1] = ((((x.w[1] ^ y.w[1]) & 0x8000000000000000)) | 0x7800000000000000);
        res.w[0] = 0;
        return (res, pfpsf);
    }
    diff_expon = ((exponent_x.wrapping_sub(exponent_y)).wrapping_add(0x1820));
    if __unsigned_compare_gt_128(CY, CX) {
        let mut f64_i: u32 = (0x5f800000 as u32);
        let mut f64_d = f32::from_bits(f64_i);
        let mut fx_d = no_fma_mul_add_f32((CX.w[1] as f32), f64_d, (CX.w[0] as f32));
        let mut fy_d = no_fma_mul_add_f32((CY.w[1] as f32), f64_d, (CY.w[0] as f32));
        let mut fx_i = (fx_d as f32).to_bits();
        let mut fy_i = (fy_d as f32).to_bits();
        bin_index = ((go_checked_shr_u32(((fy_i.wrapping_sub(fx_i))), go_shift_count_u64((23) as u64))) as i64);
        if (CX.w[1] != 0) {
            T = bid_power10_index_binexp_128[bin_index as usize].w[0];
            CA = __mul_64x128_short(T, CX);
        } else {
            T128 = bid_power10_index_binexp_128[bin_index as usize];
            CA = __mul_64x128_short(CX.w[0], T128);
        }
        ed2 = 33;
        if __unsigned_compare_gt_128(CY, CA) {
            ed2 = ed2.wrapping_add(1);
        }
        T128 = bid_power10_table_128[ed2 as usize];
        CA4 = __mul_128x128_to_256(CA, T128);
        ed2 = ed2.wrapping_add(bid_estimate_decimal_digits[bin_index as usize] as i64);
        CQ.w[0] = 0;
        CQ.w[1] = 0;
        diff_expon = (diff_expon.wrapping_sub(ed2));
    } else {
        (CQ, CR) = bid___div_128_by_128(CX, CY);
        if ((CR.w[1] == 0) && (CR.w[0] == 0)) {
            res = bid_get_bid128((sign_x ^ sign_y), diff_expon, CQ, rnd_mode, (&mut pfpsf));
            return (res, pfpsf);
        }
        let mut f64_i: u32 = (0x5f800000 as u32);
        let mut f64_d = f32::from_bits(f64_i);
        let mut fx_d = no_fma_mul_add_f32((CQ.w[1] as f32), f64_d, (CQ.w[0] as f32));
        let mut fx_i = (fx_d as f32).to_bits();
        bin_expon = ((go_checked_shr_u32(((fx_i.wrapping_sub(0x3f800000))), go_shift_count_u64((23) as u64))) as i64);
        digits_q = (bid_estimate_decimal_digits[bin_expon as usize] as i64);
        TP128.w[0] = bid_power10_index_binexp_128[bin_expon as usize].w[0];
        TP128.w[1] = bid_power10_index_binexp_128[bin_expon as usize].w[1];
        if __unsigned_compare_ge_128(CQ, TP128) {
            digits_q = digits_q.wrapping_add(1);
        }
        ed2 = ((34 as i64).wrapping_sub(digits_q));
        T128.w[0] = bid_power10_table_128[ed2 as usize].w[0];
        T128.w[1] = bid_power10_table_128[ed2 as usize].w[1];
        CA4 = __mul_128x128_to_256(CR, T128);
        diff_expon = (diff_expon.wrapping_sub(ed2));
        CQ = __mul_128x128_low(CQ, T128);
    }
    bid___div_256_by_128((&mut CQ), (&mut CA4), CY);
    if ((CA4.w[0] != 0) || (CA4.w[1] != 0)) {
        pfpsf |= 32;
    } else {
        if ((((CX.w[1] == 0) && (CY.w[1] == 0)) && (CX.w[0] <= 1024)) && (CY.w[0] <= 1024)) {
            i = ((CY.w[0] as i64).wrapping_sub(1));
            j = ((CX.w[0] as i64).wrapping_sub(1));
            nzeros = ((ed2.wrapping_sub(bid_factors[i as usize][0] as i64)).wrapping_add(bid_factors[j as usize][0] as i64));
            d5 = ((ed2.wrapping_sub(bid_factors[i as usize][1] as i64)).wrapping_add(bid_factors[j as usize][1] as i64));
            if (d5 < nzeros) {
                nzeros = d5;
            }
            (Qh, _) = __mul_128x128_full(CQ, bid_reciprocals10_128[nzeros as usize]);
            amount = (bid_recip_scale[nzeros as usize] as i64);
            CQ = __shr_128_long(Qh, (amount as u64));
            diff_expon = diff_expon.wrapping_add(nzeros);
        } else {
            T128.w[0] = 0x44909befeb9fad49;
            T128.w[1] = 0x000b877aa3236a4b;
            P256 = __mul_128x128_to_256(CQ, T128);
            Q_high = (((go_checked_shr_u64(P256.w[2], go_shift_count_u64((44) as u64)))) | ((go_checked_shl_u64(P256.w[3], go_shift_count_u64((64 - 44) as u64)))));
            Q_low = (CQ.w[0].wrapping_sub((Q_high.wrapping_mul(100000000000000000))));
            if (Q_low == 0) {
                diff_expon = diff_expon.wrapping_add(17);
                tdigit[0] = ((Q_high as u32) & 0x3ffffff);
                tdigit[1] = 0;
                QX = (go_checked_shr_u64(Q_high, go_shift_count_u64((26) as u64)));
                QX32 = (QX as u32);
                nzeros = 0;
                j = 0;
                while (QX32 != 0) {
                    k = ((QX32 & 127) as i64);
                    tdigit[0] = tdigit[0].wrapping_add(bid_convert_table[j as usize][k as usize][0]);
                    tdigit[1] = tdigit[1].wrapping_add(bid_convert_table[j as usize][k as usize][1]);
                    if (tdigit[0] >= 100000000) {
                        tdigit[0] = tdigit[0].wrapping_sub(100000000);
                        tdigit[1] = tdigit[1].wrapping_add(1);
                    }
                    j = (j.wrapping_add(1));
                    QX32 = (go_checked_shr_u32(QX32, go_shift_count_u64((7) as u64)));
                }
                if (tdigit[1] >= 100000000) {
                    tdigit[1] = tdigit[1].wrapping_sub(100000000);
                    if (tdigit[1] >= 100000000) {
                        tdigit[1] = tdigit[1].wrapping_sub(100000000);
                    }
                }
                digit = tdigit[0];
                if ((digit == 0) && (tdigit[1] == 0)) {
                    nzeros = nzeros.wrapping_add(16);
                } else {
                    if (digit == 0) {
                        nzeros = nzeros.wrapping_add(8);
                        digit = tdigit[1];
                    }
                    PD = ((digit as u64).wrapping_mul(0x068DB8BB));
                    digit_h = ((go_checked_shr_u64(PD, go_shift_count_u64((40) as u64))) as u32);
                    digit_low = (digit.wrapping_sub((digit_h.wrapping_mul(10000))));
                    if (digit_low == 0) {
                        nzeros = nzeros.wrapping_add(4);
                    } else {
                        digit_h = digit_low;
                    }
                    if ((digit_h & 1) == 0) {
                        nzeros = nzeros.wrapping_add(((3 & ((go_checked_shr_u8(bid_packed_10000_zeros[(go_checked_shr_u32(digit_h, go_shift_count_u64((3) as u64))) as usize], go_shift_count_u64((digit_h & 7) as u64))) as u32)) as i64));
                    }
                }
                if (nzeros != 0) {
                    CQ = __mul_64x64_to_128(Q_high, bid_reciprocals10_64[nzeros as usize]);
                    amount = (bid_short_recip_scale[nzeros as usize] as i64);
                    CQ.w[0] = (go_checked_shr_u64(CQ.w[1], go_shift_count_u64((amount as u64) as u64)));
                } else {
                    CQ.w[0] = Q_high;
                }
                CQ.w[1] = 0;
                diff_expon = diff_expon.wrapping_add(nzeros);
            } else {
                tdigit[0] = ((Q_low as u32) & 0x3ffffff);
                tdigit[1] = 0;
                QX = (go_checked_shr_u64(Q_low, go_shift_count_u64((26) as u64)));
                QX32 = (QX as u32);
                nzeros = 0;
                j = 0;
                while (QX32 != 0) {
                    k = ((QX32 & 127) as i64);
                    tdigit[0] = tdigit[0].wrapping_add(bid_convert_table[j as usize][k as usize][0]);
                    tdigit[1] = tdigit[1].wrapping_add(bid_convert_table[j as usize][k as usize][1]);
                    if (tdigit[0] >= 100000000) {
                        tdigit[0] = tdigit[0].wrapping_sub(100000000);
                        tdigit[1] = tdigit[1].wrapping_add(1);
                    }
                    j = (j.wrapping_add(1));
                    QX32 = (go_checked_shr_u32(QX32, go_shift_count_u64((7) as u64)));
                }
                if (tdigit[1] >= 100000000) {
                    tdigit[1] = tdigit[1].wrapping_sub(100000000);
                    if (tdigit[1] >= 100000000) {
                        tdigit[1] = tdigit[1].wrapping_sub(100000000);
                    }
                }
                digit = tdigit[0];
                if ((digit == 0) && (tdigit[1] == 0)) {
                    nzeros = nzeros.wrapping_add(16);
                } else {
                    if (digit == 0) {
                        nzeros = nzeros.wrapping_add(8);
                        digit = tdigit[1];
                    }
                    PD = ((digit as u64).wrapping_mul(0x068DB8BB));
                    digit_h = ((go_checked_shr_u64(PD, go_shift_count_u64((40) as u64))) as u32);
                    digit_low = (digit.wrapping_sub((digit_h.wrapping_mul(10000))));
                    if (digit_low == 0) {
                        nzeros = nzeros.wrapping_add(4);
                    } else {
                        digit_h = digit_low;
                    }
                    if ((digit_h & 1) == 0) {
                        nzeros = nzeros.wrapping_add(((3 & ((go_checked_shr_u8(bid_packed_10000_zeros[(go_checked_shr_u32(digit_h, go_shift_count_u64((3) as u64))) as usize], go_shift_count_u64((digit_h & 7) as u64))) as u32)) as i64));
                    }
                }
                if (nzeros != 0) {
                    (Qh, _) = __mul_128x128_full(CQ, bid_reciprocals10_128[nzeros as usize]);
                    amount = (bid_recip_scale[nzeros as usize] as i64);
                    CQ = __shr_128(Qh, (amount as u64));
                }
                diff_expon = diff_expon.wrapping_add(nzeros);
            }
        }
        res = bid_get_bid128((sign_x ^ sign_y), diff_expon, CQ, rnd_mode, (&mut pfpsf));
        return (res, pfpsf);
    }
    if (diff_expon >= 0) {
        rmode = (rnd_mode as u64);
        if (((sign_x ^ sign_y) != 0) && (((rmode.wrapping_sub(1)) as u64) < 2)) {
            rmode = ((3 as u64).wrapping_sub(rmode));
        }
        match rmode {
            0 => {
                CA4r.w[1] = (((CA4.w[1].wrapping_add(CA4.w[1]))) | ((go_checked_shr_u64(CA4.w[0], go_shift_count_u64((63) as u64)))));
                CA4r.w[0] = (CA4.w[0].wrapping_add(CA4.w[0]));
                (CA4r.w[0], carry64) = __sub_borrow_out(CA4r.w[0], CY.w[0]);
                CA4r.w[1] = ((CA4r.w[1].wrapping_sub(CY.w[1])).wrapping_sub(carry64));
                if ((CA4r.w[1] | CA4r.w[0]) != 0) {
                    D = 1;
                } else {
                    D = 0;
                }
                carry64 = ((((1 as i64).wrapping_add((go_checked_shr_i64((CA4r.w[1] as i64), go_shift_count_u64((63) as u64))))) as u64) & (((CQ.w[0]) | D)));
                CQ.w[0] = CQ.w[0].wrapping_add(carry64);
                if (CQ.w[0] < carry64) {
                    CQ.w[1] = CQ.w[1].wrapping_add(1);
                }
            }
            4 => {
                CA4r.w[1] = (((CA4.w[1].wrapping_add(CA4.w[1]))) | ((go_checked_shr_u64(CA4.w[0], go_shift_count_u64((63) as u64)))));
                CA4r.w[0] = (CA4.w[0].wrapping_add(CA4.w[0]));
                (CA4r.w[0], carry64) = __sub_borrow_out(CA4r.w[0], CY.w[0]);
                CA4r.w[1] = ((CA4r.w[1].wrapping_sub(CY.w[1])).wrapping_sub(carry64));
                if ((CA4r.w[1] | CA4r.w[0]) != 0) {
                    D = 0;
                } else {
                    D = 1;
                }
                carry64 = ((((1 as i64).wrapping_add((go_checked_shr_i64((CA4r.w[1] as i64), go_shift_count_u64((63) as u64))))) as u64) | D);
                CQ.w[0] = CQ.w[0].wrapping_add(carry64);
                if (CQ.w[0] < carry64) {
                    CQ.w[1] = CQ.w[1].wrapping_add(1);
                }
            }
            1 | 3 => {
            }
            _ => {
                CQ.w[0] = CQ.w[0].wrapping_add(1);
                if (CQ.w[0] == 0) {
                    CQ.w[1] = CQ.w[1].wrapping_add(1);
                }
            }
        }
    } else {
        if ((CA4.w[0] != 0) || (CA4.w[1] != 0)) {
            pfpsf |= 32;
        }
        res = bid_handle_uf_128_rem((sign_x ^ sign_y), diff_expon, CQ, (CA4.w[1] | CA4.w[0]), rnd_mode, (&mut pfpsf));
        return (res, pfpsf);
    }
    res = bid_get_bid128((sign_x ^ sign_y), diff_expon, CQ, rnd_mode, (&mut pfpsf));
    return (res, pfpsf);
}
