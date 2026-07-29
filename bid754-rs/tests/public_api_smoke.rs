//! Representative public-API smoke tests.
//!
//! Systematic public-API parity coverage lives in the generated parity test.
//! These focused cases check representative crate-root entrypoints and compare
//! each wrapper with its direct generated-port call.

use bid754::{
    Binary128, Context, Decimal128, Decimal32, Decimal64, DecimalClass, ExceptionFlags,
    RoundingMode,
};
use std::convert::TryFrom;
use std::str::FromStr;

#[test]
fn parse_routes_through_port() {
    for s in ["1.5", "0", "-3.25", "123456.789", "1E10", "9999999999999999"] {
        let d = Decimal64::parse(s).expect("valid decimal parses");
        let (raw_bits, _) = bid754::bid64_from_string_raw(s, 0);
        assert_eq!(d.to_bits(), raw_bits, "parse({s}) must route through the port");
    }
}

#[test]
fn parse_rejects_malformed() {
    assert!(Decimal64::parse("").is_err());
    assert!(Decimal64::parse("   ").is_err());
    assert!(Decimal64::parse("garbage").is_err());
    assert!(Decimal64::parse("1.2.3").is_err());
    assert!(Decimal64::parse("nan(1)").is_err());
}

#[test]
fn parse_accepts_nan_literals() {
    // The mechanical port maps malformed text to a NaN too, so the parse Ok/Err
    // decision cannot be "is the result NaN"; it uses the NaN-literal grammar.
    for s in ["nan", "NaN", "-nan", "snan", "SNaN123", "+qnan", "nan42"] {
        assert!(
            Decimal64::parse(s).is_ok(),
            "parse({s}) is a valid NaN literal and must be Ok"
        );
    }
}

#[test]
fn from_str_matches_parse() {
    let a = Decimal64::from_str("42.5").expect("valid");
    let b = Decimal64::parse("42.5").expect("valid");
    assert_eq!(a, b);
    assert!("garbage".parse::<Decimal64>().is_err());
}

#[test]
fn parse_raw_rejects_malformed_through_flags() {
    // The raw public surface declares IEEE exception flags as its failure
    // channel. Intel from_string itself maps malformed text to qNaN or zero
    // with no status, so the public wrapper must supply the missing invalid
    // signal and canonical invalid result.
    for s in ["garbage", "", ".", "1e2junk", "1 ", "--1"] {
        let (value, flags) = Decimal64::parse_raw(s);
        assert_eq!(
            value.to_bits(),
            0x7c00_0000_0000_0000,
            "parse_raw({s}) bits"
        );
        assert_eq!(
            flags,
            ExceptionFlags::INVALID_OPERATION,
            "parse_raw({s}) flags"
        );
    }

    // Valid finite input still exposes the mechanical port result and flags.
    let (value, flags) = Decimal64::parse_raw("12.75");
    let (raw_bits, raw_flags) = bid754::bid64_from_string_raw("12.75", 0);
    assert_eq!(value.to_bits(), raw_bits, "valid parse_raw bits");
    assert_eq!(
        flags.bits(),
        map_bidgo_flags(raw_flags),
        "valid parse_raw flags"
    );

    // An out-of-range exponent raises overflow+inexact (this is where the port
    // does raise flags), proving flags are actually plumbed through.
    let (_, big_flags) = Decimal64::parse_raw("1e999");
    assert!(big_flags.contains(ExceptionFlags::OVERFLOW));
    assert!(big_flags.contains(ExceptionFlags::INEXACT));
}

/// Maps the raw bidgo status word into the public ExceptionFlags bit layout,
/// independently of the wrapper's own converter (so a shared-converter bug is
/// visible), mirroring the Go parity runner's independent-literal approach.
fn map_bidgo_flags(raw: u32) -> u32 {
    let mut out = 0u32;
    if raw & 32 != 0 {
        out |= ExceptionFlags::INEXACT.bits();
    }
    if raw & 16 != 0 {
        out |= ExceptionFlags::UNDERFLOW.bits();
    }
    if raw & 8 != 0 {
        out |= ExceptionFlags::OVERFLOW.bits();
    }
    if raw & 4 != 0 {
        out |= ExceptionFlags::DIVISION_BY_ZERO.bits();
    }
    if raw & 1 != 0 {
        out |= ExceptionFlags::INVALID_OPERATION.bits();
    }
    out
}

#[test]
fn add_routes_through_port() {
    let a = Decimal64::parse("2.5").expect("valid");
    let b = Decimal64::parse("1.25").expect("valid");
    let want = bid754::generated::add64::bid64_add(a.to_bits(), b.to_bits(), 0);
    assert_eq!(a.add(b).to_bits(), want);
    // Addition is commutative bit-for-bit.
    assert_eq!(a.add(b), b.add(a));
}

#[test]
fn add_with_flags_routes_through_port() {
    let a = Decimal64::parse("1").expect("valid");
    let b = Decimal64::parse("2").expect("valid");
    let (sum, flags) = a.add_with_flags(b);
    let (want_bits, _) = bid754::generated::add64::bid64_add_with_flags(a.to_bits(), b.to_bits(), 0);
    assert_eq!(sum.to_bits(), want_bits);
    // 1 + 2 = 3 is exact, so no flags are raised.
    assert!(flags.is_empty());
}

#[test]
fn display_routes_through_port() {
    let d = Decimal64::parse("2.5").expect("valid");
    assert_eq!(
        d.to_string(),
        bid754::generated::string64::bid64_to_string(d.to_bits())
    );
}

#[test]
fn rounding_mode_is_a_closed_ieee_set() {
    assert_eq!(RoundingMode::try_from(0), Ok(RoundingMode::NearestEven));
    assert_eq!(RoundingMode::try_from(1), Ok(RoundingMode::NearestAway));
    assert_eq!(RoundingMode::try_from(2), Ok(RoundingMode::TowardZero));
    assert_eq!(RoundingMode::try_from(3), Ok(RoundingMode::TowardPositive));
    assert_eq!(RoundingMode::try_from(4), Ok(RoundingMode::TowardNegative));
    // The non-IEEE decTest-compat NearestDown(5) is not representable publicly.
    assert!(RoundingMode::try_from(5).is_err());
    assert!(RoundingMode::try_from(99).is_err());
}

#[test]
fn exception_flags_vocabulary_matches_go() {
    assert_eq!(ExceptionFlags::empty().to_string(), "None");
    let f = ExceptionFlags::INEXACT | ExceptionFlags::OVERFLOW;
    assert_eq!(f.to_string(), "Inexact|Overflow");
    assert!(f.contains(ExceptionFlags::INEXACT));
    assert!(!f.contains(ExceptionFlags::UNDERFLOW));
    assert!(!f.is_empty());
    assert!(ExceptionFlags::empty().is_empty());
    // All eight flags render.
    let all = ExceptionFlags::INEXACT
        | ExceptionFlags::UNDERFLOW
        | ExceptionFlags::OVERFLOW
        | ExceptionFlags::DIVISION_BY_ZERO
        | ExceptionFlags::INVALID_OPERATION
        | ExceptionFlags::SUBNORMAL
        | ExceptionFlags::ROUNDED
        | ExceptionFlags::CLAMPED;
    assert_eq!(
        all.to_string(),
        "Inexact|Underflow|Overflow|DivisionByZero|InvalidOperation|Subnormal|Rounded|Clamped"
    );
}

#[test]
fn decimal_class_uses_gda_spellings() {
    assert_eq!(DecimalClass::SignalingNaN.to_string(), "sNaN");
    assert_eq!(DecimalClass::QuietNaN.to_string(), "NaN");
    assert_eq!(DecimalClass::PositiveNormal.to_string(), "+Normal");
    assert_eq!(DecimalClass::NegativeInfinity.to_string(), "-Infinity");
    assert_eq!(DecimalClass::PositiveZero.to_string(), "+Zero");
}

#[test]
fn value_types_are_fixed_width_with_bit_roundtrip() {
    assert_eq!(std::mem::size_of::<Decimal32>(), 4);
    assert_eq!(std::mem::size_of::<Decimal64>(), 8);
    assert_eq!(std::mem::size_of::<Decimal128>(), 16);
    assert_eq!(std::mem::size_of::<Binary128>(), 16);

    for bit in 0..32 {
        let bits = 1u32 << bit;
        assert_eq!(Decimal32::from_bits(bits).to_bits(), bits, "bit {bit}");
    }
    for bit in 0..64 {
        let bits = 1u64 << bit;
        assert_eq!(Decimal64::from_bits(bits).to_bits(), bits, "bit {bit}");
    }
    for bit in 0..128 {
        let mut bytes = [0u8; 16];
        bytes[bit / 8] = 1u8 << (bit % 8);
        assert_eq!(
            Decimal128::from_le_bytes(bytes).to_le_bytes(),
            bytes,
            "decimal128 bit {bit}"
        );
        assert_eq!(
            Binary128::from_le_bytes(bytes).to_le_bytes(),
            bytes,
            "binary128 bit {bit}"
        );
    }
}

#[test]
fn context_accumulates_flags() {
    let mut ctx = Context::new();
    assert_eq!(ctx.rounding, RoundingMode::NearestEven);
    assert!(ctx.flags.is_empty());

    ctx.set_flag(ExceptionFlags::INEXACT);
    assert!(ctx.has_flag(ExceptionFlags::INEXACT));
    let saved = ctx.save_all_flags();

    ctx.clear_flag(ExceptionFlags::INEXACT);
    assert!(!ctx.has_flag(ExceptionFlags::INEXACT));

    // restore_flags is a masked write (5.7.4 restoreFlags): the mask selects
    // which bits are taken from `saved`; every bit outside the mask keeps its
    // current value. OVERFLOW is raised here and is not in the mask, so it must
    // survive a restore whose `saved` snapshot does not contain it. The full
    // masked-write contract is pinned in tests/context_save_restore_flags.rs.
    ctx.set_flag(ExceptionFlags::OVERFLOW);
    ctx.restore_flags(saved, ExceptionFlags::INEXACT);
    assert!(ctx.has_flag(ExceptionFlags::INEXACT));
    assert!(
        ctx.has_flag(ExceptionFlags::OVERFLOW),
        "restore_flags must preserve bits outside the mask, got {}",
        ctx.flags
    );

    ctx.clear_all_flags();
    assert!(ctx.flags.is_empty());

    let c2 = Context::with_rounding(RoundingMode::TowardZero);
    assert_eq!(c2.rounding, RoundingMode::TowardZero);
}

#[test]
fn parse_error_carries_input() {
    let err = Decimal64::parse("garbage").unwrap_err();
    assert_eq!(err.input(), "garbage");
    assert!(err.to_string().contains("garbage"));
}

// Representative Decimal64 smoke checks complement the systematic generated
// public-API parity runner. They verify crate-root exposure and representative
// wrapper behavior, including exact NaN payload round trips.

#[test]
fn nan_payload_round_trips_through_parse_and_display() {
    for s in ["NaN", "-NaN", "NaN123", "+SNaN42", "snan7"] {
        let d = Decimal64::parse(s).expect("NaN literal parses");
        let rendered = d.to_string();
        let reparsed = Decimal64::parse(&rendered).expect("rendered NaN literal reparses");
        // Compare bits, not `==`: Decimal64's `==` is IEEE quiet equality, so a
        // NaN never equals itself. The round-trip we care about here is
        // bit-identity of the payload.
        assert_eq!(
            d.to_bits(),
            reparsed.to_bits(),
            "parse({s}) -> Display -> parse must round-trip bit-for-bit, rendered as {rendered}"
        );
    }
    // A zero payload is not canonical (mirrors Go payloadString) and is
    // suppressed on display, exactly like a payload-less literal.
    let zero_payload = Decimal64::parse("NaN0").expect("valid");
    assert_eq!(zero_payload.to_string(), "+NaN");
}

#[test]
fn decimal64_arithmetic_predicate_and_compare_family_routes_through_port() {
    let a = Decimal64::parse("2.5").expect("valid");
    let b = Decimal64::parse("1.25").expect("valid");

    let want_sub = bid754::generated::add64::bid64_sub(a.to_bits(), b.to_bits(), 0);
    assert_eq!(a.sub(b).to_bits(), want_sub);

    assert!(!a.is_zero());
    assert!(Decimal64::parse("0").expect("valid").is_zero());
    assert_eq!(a.sign(), 1);
    assert_eq!(a.class(), DecimalClass::PositiveNormal);
    assert_eq!(Decimal64::RADIX, 10);

    let (eq, flags) = a.quiet_eq(a);
    assert!(eq, "a must quiet-equal itself");
    assert!(flags.is_empty());
    assert_eq!(a.total_cmp(b), core::cmp::Ordering::Greater);
}

#[test]
fn decimal64_conversion_family_routes_through_port() {
    let d = Decimal64::parse("42").expect("valid");

    let (i, flags) = d.to_i32(RoundingMode::NearestEven);
    let (want_i, want_raw) = bid754::generated::to_int32::bid64_to_int32_rnint(d.to_bits());
    assert_eq!(i, want_i);
    assert_eq!(flags.bits(), map_bidgo_flags(want_raw));

    let (widened, flags) = d.to_decimal128();
    let (want_bits, want_raw) = bid754::generated::to_bid12864::bid64_to_bid128(d.to_bits());
    let mut want_le = [0u8; 16];
    want_le[0..8].copy_from_slice(&want_bits.lo.to_le_bytes());
    want_le[8..16].copy_from_slice(&want_bits.hi.to_le_bytes());
    assert_eq!(widened.to_le_bytes(), want_le);
    assert_eq!(flags.bits(), map_bidgo_flags(want_raw));
}

#[test]
fn eq_and_partial_ord_use_ieee_quiet_semantics() {
    // Decimal64's `==` is IEEE quiet equality (like f64), NOT bit-equality.
    let nan = Decimal64::parse("NaN").expect("valid");
    assert!(nan != nan, "quiet == makes a NaN unequal to itself");
    assert_eq!(nan.partial_cmp(&nan), None, "a NaN is unordered");

    let pos_zero = Decimal64::parse("0").expect("valid");
    let neg_zero = Decimal64::parse("-0").expect("valid");
    // -0 and +0 have distinct bit patterns but are numerically equal.
    assert_ne!(pos_zero.to_bits(), neg_zero.to_bits(), "distinct bit patterns");
    assert_eq!(pos_zero, neg_zero, "quiet == makes -0 equal +0");

    let one = Decimal64::parse("1").expect("valid");
    let two = Decimal64::parse("2").expect("valid");
    assert!(one < two);
    assert!(two > one);
    assert_eq!(one.partial_cmp(&two), Some(core::cmp::Ordering::Less));
}

#[test]
fn context_carries_mode_into_with_mode_operations() {
    let mut ctx = Context::with_rounding(RoundingMode::TowardZero);
    // 16-digit operands whose 17-digit exact sum must round: the result
    // differs between TowardZero and NearestEven (so a mode miswire changes
    // the bits) and the operation raises INEXACT (so the flag-accumulation
    // assertion below cannot pass vacuously on empty flags).
    let a = Decimal64::parse("9.999999999999999").expect("valid");
    let b = Decimal64::parse("9.999999999999999").expect("valid");
    let (sum, flags) = a.add_with_mode(b, ctx.rounding);
    ctx.set_flag(flags);
    // bidgo-domain TowardZero = 3 (independent literal, not the wrapper's own
    // to_bidgo_rounding converter, so a shared-converter bug stays visible).
    let (want_bits, want_raw) = bid754::generated::add64::bid64_add_with_flags(a.to_bits(), b.to_bits(), 3);
    let (nearest_bits, _) = bid754::generated::add64::bid64_add_with_flags(a.to_bits(), b.to_bits(), 0);
    assert_ne!(want_bits, nearest_bits, "operands must discriminate the rounding mode");
    assert_eq!(sum.to_bits(), want_bits);
    assert!(!ctx.flags.is_empty(), "operands must raise flags so accumulation is observable");
    assert_eq!(ctx.flags.bits(), map_bidgo_flags(want_raw));
}

// Representative ZERO/ONE/PI/E smoke checks complement the systematic
// const == parse(literal) parity gate (generated as
// generated_public_api_const_parity_{32,64,128} in
// tests/public_parity_generated.rs, one case per constant). This is one
// representative constant per width proving the constants are
// exposed at the crate root, and agree with a runtime parse of their own
// documented literal.

#[test]
fn constants_agree_with_parsing_their_own_literal() {
    assert_eq!(Decimal32::PI.to_bits(), Decimal32::parse("3.141593").expect("valid").to_bits());
    assert_eq!(Decimal64::ONE.to_bits(), Decimal64::parse("1").expect("valid").to_bits());
    assert_eq!(
        Decimal128::E.to_le_bytes(),
        Decimal128::parse("2.718281828459045235360287471352662")
            .expect("valid")
            .to_le_bytes()
    );
}
