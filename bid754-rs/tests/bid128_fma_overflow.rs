use bid754::generated::prelude::*;

// Locks the fix for the Case (1''B) overflow-path transcription defect of the
// bid128_ext_fma port (see bid754-go/internal/bidgo/bid128_fma_overflow_test.go
// for the full provenance; this generated Rust implementation is produced from
// that Go port and carried the identical defect). With x = 1E+34,
// y = +/-Nmax, z = -/+Nmax the fused result overflows (exact product
// (10^34-1)E+6145), and the pre-fix code returned non-canonical bits with the
// unclamped biased exponent 12321 instead of the Intel-C-pinned +/-Inf (RN and
// the away-side directed modes) or +/-Nmax (the truncating directed modes).
#[test]
fn bid128_fma_case1ppb_overflow_encoding() {
    let one_e34 = BID_UINT128 {
        lo: 0x0000000000000001,
        hi: 0x3084000000000000,
    };
    let pos_nmax = BID_UINT128 {
        lo: 0x378d8e63ffffffff,
        hi: 0x5fffed09bead87c0,
    };
    let neg_nmax = BID_UINT128 {
        lo: 0x378d8e63ffffffff,
        hi: 0xdfffed09bead87c0,
    };
    let pos_inf = BID_UINT128 {
        lo: 0,
        hi: 0x7800000000000000,
    };
    let neg_inf = BID_UINT128 {
        lo: 0,
        hi: 0xf800000000000000,
    };
    // BID_OVERFLOW_EXCEPTION | BID_INEXACT_EXCEPTION
    let want_flags: u32 = 0x08 | 0x20;

    // (y, z, rounding mode, expected result)
    let cases: [(BID_UINT128, BID_UINT128, i64, BID_UINT128); 10] = [
        (pos_nmax, neg_nmax, 0, pos_inf),  // nearest even
        (pos_nmax, neg_nmax, 1, pos_nmax), // toward negative
        (pos_nmax, neg_nmax, 2, pos_inf),  // toward positive
        (pos_nmax, neg_nmax, 3, pos_nmax), // toward zero
        (pos_nmax, neg_nmax, 4, pos_inf),  // nearest away
        (neg_nmax, pos_nmax, 0, neg_inf),
        (neg_nmax, pos_nmax, 1, neg_inf),
        (neg_nmax, pos_nmax, 2, neg_nmax),
        (neg_nmax, pos_nmax, 3, neg_nmax),
        (neg_nmax, pos_nmax, 4, neg_inf),
    ];
    for (y, z, rnd_mode, want) in cases {
        let (got, flags) = bid128_fma(one_e34, y, z, rnd_mode);
        assert_eq!(
            (got, flags),
            (want, want_flags),
            "fma(1E+34, {:016x}:{:016x}, {:016x}:{:016x}) mode {} = {:016x}:{:016x}/{:02x}, want {:016x}:{:016x}/{:02x}",
            y.hi, y.lo, z.hi, z.lo, rnd_mode,
            got.hi, got.lo, flags, want.hi, want.lo, want_flags
        );
    }
}
