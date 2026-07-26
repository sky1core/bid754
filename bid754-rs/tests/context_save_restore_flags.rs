//! IEEE 754-2019 5.7.4 `saveAllFlags`/`restoreFlags` contract pins for the
//! public Rust `Context`.
//!
//! These are context flag plumbing with no Intel port counterpart, so the
//! public-API parity gate excludes them (`excluded_context_flag_plumbing` in
//! `devtools/generated/testspec/public_api_routing_inventory.json`) and no
//! generated verification domain exercises them. That makes a hand-written pin
//! the only thing standing between the two languages' 5.7.4 public contract and
//! silent divergence: `restore_flags` is a **masked write**, not a full
//! overwrite of the flag word.
//!
//! The first test mirrors the Go `TestArithmeticContextSaveRestoreFlags`
//! (`bid754-go/context_v2_test.go`) case-for-case and value-for-value, so a
//! reviewer can put the two side by side.
//!
//! Vocabulary note: Go's value-level `ExceptionFlags.HasFlag` is an *any-bit*
//! test, while the Rust value-level `ExceptionFlags::contains` is an *all-bits*
//! test. The any-bit form used by `Context::has_flag` is written out here as
//! `!(flags & mask).is_empty()` so the Go assertions translate exactly.

use bid754::{Context, ExceptionFlags};

/// Mirror of Go `TestArithmeticContextSaveRestoreFlags`. The saved snapshot, the
/// flags raised before the restore, and the mask are the same values as the Go
/// test, chosen so that every masked-write role occurs in one call:
///
/// | flag              | in `saved` | in `mask` | raised before | expected after                        |
/// |-------------------|------------|-----------|---------------|---------------------------------------|
/// | INEXACT           | yes        | yes       | no            | raised — taken from `saved`           |
/// | OVERFLOW          | yes        | yes       | no            | raised — taken from `saved`           |
/// | INVALID_OPERATION | no         | yes       | yes           | lowered — in mask, absent from `saved`|
/// | DIVISION_BY_ZERO  | no         | no        | yes           | raised — outside mask, preserved      |
/// | CLAMPED           | yes        | no        | no            | clear — outside mask, not restored    |
#[test]
fn save_all_flags_then_restore_flags_is_a_masked_write() {
    let mut ctx = Context::new();
    ctx.set_flag(ExceptionFlags::INEXACT | ExceptionFlags::OVERFLOW | ExceptionFlags::CLAMPED);

    let saved = ctx.save_all_flags();
    assert_eq!(
        saved,
        ExceptionFlags::INEXACT | ExceptionFlags::OVERFLOW | ExceptionFlags::CLAMPED,
        "save_all_flags must snapshot exactly the raised flags"
    );

    // 5.7.4 testSavedFlags on the snapshot: any-bit, so a mask that overlaps
    // the snapshot in a single bit still reports true, and a mask disjoint from
    // it reports false.
    assert!(
        !(saved & (ExceptionFlags::OVERFLOW | ExceptionFlags::INVALID_OPERATION)).is_empty(),
        "a mask overlapping the snapshot in one bit must test true, saved = {saved}"
    );
    assert!(
        (saved & (ExceptionFlags::DIVISION_BY_ZERO | ExceptionFlags::INVALID_OPERATION)).is_empty(),
        "a mask disjoint from the snapshot must test false, saved = {saved}"
    );

    ctx.clear_all_flags();
    ctx.set_flag(ExceptionFlags::DIVISION_BY_ZERO | ExceptionFlags::INVALID_OPERATION);
    ctx.restore_flags(
        saved,
        ExceptionFlags::INEXACT | ExceptionFlags::OVERFLOW | ExceptionFlags::INVALID_OPERATION,
    );

    let want =
        ExceptionFlags::DIVISION_BY_ZERO | ExceptionFlags::INEXACT | ExceptionFlags::OVERFLOW;
    assert_eq!(
        ctx.flags, want,
        "restore_flags partial mask = {}, want {want}",
        ctx.flags
    );

    // The same expectation stated per role, so a failure names which part of the
    // masked write broke rather than only showing two flag words.
    assert!(
        ctx.has_flag(ExceptionFlags::INEXACT) && ctx.has_flag(ExceptionFlags::OVERFLOW),
        "masked bits present in `saved` must be raised, got {}",
        ctx.flags
    );
    assert!(
        !ctx.has_flag(ExceptionFlags::INVALID_OPERATION),
        "a masked bit absent from `saved` must be lowered, got {}",
        ctx.flags
    );
    assert!(
        ctx.has_flag(ExceptionFlags::DIVISION_BY_ZERO),
        "a raised bit outside the mask must be preserved, got {}",
        ctx.flags
    );
    assert!(
        !ctx.has_flag(ExceptionFlags::CLAMPED),
        "a bit present in `saved` but outside the mask must not be restored, got {}",
        ctx.flags
    );

    // Direct pin against the pre-fix Rust behavior: a full overwrite
    // (`self.flags = saved`) would leave the flag word exactly equal to `saved`
    // for these operands. The masked write must not.
    assert_ne!(
        ctx.flags, saved,
        "restore_flags must not overwrite the whole flag word with `saved`"
    );
}

/// The mask boundaries of the same masked write: an empty mask changes nothing,
/// and a mask covering the whole public flag domain reproduces `saved` exactly.
/// No implicit masking is applied to either end (the Go `RestoreFlags` doc
/// contract: the whole `ExceptionFlags` domain is public).
#[test]
fn restore_flags_mask_boundaries() {
    // The whole public flag domain, spelled out from the eight public constants
    // (`ExceptionFlags` exposes no `all()` constructor).
    let all = ExceptionFlags::INEXACT
        | ExceptionFlags::UNDERFLOW
        | ExceptionFlags::OVERFLOW
        | ExceptionFlags::DIVISION_BY_ZERO
        | ExceptionFlags::INVALID_OPERATION
        | ExceptionFlags::SUBNORMAL
        | ExceptionFlags::ROUNDED
        | ExceptionFlags::CLAMPED;

    let raised = ExceptionFlags::DIVISION_BY_ZERO | ExceptionFlags::ROUNDED;
    let saved = ExceptionFlags::INEXACT | ExceptionFlags::SUBNORMAL;

    let mut ctx = Context::new();
    ctx.set_flag(raised);
    ctx.restore_flags(saved, ExceptionFlags::empty());
    assert_eq!(
        ctx.flags, raised,
        "an empty mask selects no bit from `saved`, so the flag word is unchanged"
    );

    let mut ctx = Context::new();
    ctx.set_flag(raised);
    ctx.restore_flags(saved, all);
    assert_eq!(
        ctx.flags, saved,
        "a mask covering the whole public flag domain restores `saved` exactly"
    );
}
