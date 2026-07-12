package bid754

import (
	"math"
	"testing"
)

// TestScaleBLargeExponentDomain locks the fix for the int64->C-int ScaleB
// domain-aliasing defect. The public ScaleB/ScaleBWithMode take a Go int
// exponent (64-bit), which used to be handed straight to the int-domain
// bid<w>_scalbn port; inside that port a |n| >= 2^32 aliased through a
// uint32 truncation back into the valid exponent window and silently
// returned the operand unchanged with no status flags (e.g. 1.ScaleB(1<<32)
// == +1E+0, no flags).
//
// The public path now routes through the Intel bid<w>_scalbln long-int
// entrypoints, which clamp n into [INT32_MIN, INT32_MAX] inside the port
// before scaling. So every |n| here (all far outside the representable
// exponent range for these widths) resolves to a genuine overflow or
// underflow: a large positive n -> signed infinity with Overflow|Inexact,
// a large negative n -> signed zero with Underflow|Inexact, and the operand
// sign is preserved through both.
//
// Note: this test asserts 64-bit int semantics (the defect is specific to a
// Go int wider than the C int contract); the shifts are built at runtime to
// avoid a constant-overflow on a hypothetical 32-bit int build.
func TestScaleBLargeExponentDomain(t *testing.T) {
	two31 := int(1) << 31
	two32 := int(1) << 32
	// Boundary and beyond-boundary exponents on both sides of the int32 range.
	hugePos := []int{two31 - 1, two31, two31 + 1, two32, two32 + 5, math.MaxInt64}
	hugeNeg := []int{-(two31 - 1), -two31, -(two31 + 1), -two32 + 5, math.MinInt64}

	type probe struct {
		name       string
		neg        bool // operand is -1 (true) or +1 (false)
		scaleB     func(n int) (inf, zero, signMinus bool, flags ExceptionFlags)
		scaleBMode func(n int, mode RoundingMode) (inf, zero, signMinus bool, flags ExceptionFlags)
	}

	mk32 := func(lit string, neg bool) probe {
		v := mustDecimal32BID(t, lit)
		return probe{name: "Decimal32BID(" + lit + ")", neg: neg,
			scaleB: func(n int) (bool, bool, bool, ExceptionFlags) {
				r, f := v.ScaleB(n)
				return r.IsInf(), r.IsZero(), r.IsSignMinus(), f
			},
			scaleBMode: func(n int, m RoundingMode) (bool, bool, bool, ExceptionFlags) {
				r, f := v.ScaleBWithMode(n, m)
				return r.IsInf(), r.IsZero(), r.IsSignMinus(), f
			}}
	}
	mk64 := func(lit string, neg bool) probe {
		v := mustDecimal64BID(t, lit)
		return probe{name: "Decimal64BID(" + lit + ")", neg: neg,
			scaleB: func(n int) (bool, bool, bool, ExceptionFlags) {
				r, f := v.ScaleB(n)
				return r.IsInf(), r.IsZero(), r.IsSignMinus(), f
			},
			scaleBMode: func(n int, m RoundingMode) (bool, bool, bool, ExceptionFlags) {
				r, f := v.ScaleBWithMode(n, m)
				return r.IsInf(), r.IsZero(), r.IsSignMinus(), f
			}}
	}
	mk128 := func(lit string, neg bool) probe {
		v := mustDecimal128BID(t, lit)
		return probe{name: "Decimal128BID(" + lit + ")", neg: neg,
			scaleB: func(n int) (bool, bool, bool, ExceptionFlags) {
				r, f := v.ScaleB(n)
				return r.IsInf(), r.IsZero(), r.IsSignMinus(), f
			},
			scaleBMode: func(n int, m RoundingMode) (bool, bool, bool, ExceptionFlags) {
				r, f := v.ScaleBWithMode(n, m)
				return r.IsInf(), r.IsZero(), r.IsSignMinus(), f
			}}
	}

	probes := []probe{
		mk32("1", false), mk32("-1", true),
		mk64("1", false), mk64("-1", true),
		mk128("1", false), mk128("-1", true),
	}

	// The mode entrypoint is exercised at nearest-even so its expected result
	// matches the default ScaleB (both saturate overflow to infinity); a
	// separate check below confirms a directed mode reaches the same clamped
	// scaler.
	check := func(t *testing.T, name string, inf, zero, signMinus bool, flags ExceptionFlags, wantInf bool, wantSign bool) {
		t.Helper()
		if wantInf {
			if !inf || zero {
				t.Errorf("%s: want signed infinity, got inf=%v zero=%v", name, inf, zero)
			}
			if flags != FlagOverflow|FlagInexact {
				t.Errorf("%s: flags = %v, want Overflow|Inexact", name, flags)
			}
		} else {
			if inf || !zero {
				t.Errorf("%s: want signed zero, got inf=%v zero=%v", name, inf, zero)
			}
			if flags != FlagUnderflow|FlagInexact {
				t.Errorf("%s: flags = %v, want Underflow|Inexact", name, flags)
			}
		}
		if signMinus != wantSign {
			t.Errorf("%s: IsSignMinus = %v, want %v (operand sign not preserved)", name, signMinus, wantSign)
		}
	}

	for _, p := range probes {
		for _, n := range hugePos {
			inf, zero, sign, flags := p.scaleB(n)
			check(t, p.name+".ScaleB(+huge)", inf, zero, sign, flags, true, p.neg)
			infM, zeroM, signM, flagsM := p.scaleBMode(n, RoundNearestEven)
			check(t, p.name+".ScaleBWithMode(+huge,nearest)", infM, zeroM, signM, flagsM, true, p.neg)
		}
		for _, n := range hugeNeg {
			inf, zero, sign, flags := p.scaleB(n)
			check(t, p.name+".ScaleB(-huge)", inf, zero, sign, flags, false, p.neg)
			infM, zeroM, signM, flagsM := p.scaleBMode(n, RoundNearestEven)
			check(t, p.name+".ScaleBWithMode(-huge,nearest)", infM, zeroM, signM, flagsM, false, p.neg)
		}
	}
}

// TestScaleBLargeExponentBugFingerprints pins the two concrete cases from the
// original defect report against their corrected results, so a regression to
// the aliasing behavior (operand returned unchanged, no flags) fails here.
func TestScaleBLargeExponentBugFingerprints(t *testing.T) {
	one := mustDecimal64BID(t, "1")

	// Was: +1E+0 with no flags (n mod 2^32 aliased to 0).
	if r, f := one.ScaleB(int(1) << 32); !r.IsInf() || r.IsSignMinus() || f != FlagOverflow|FlagInexact {
		t.Errorf("1.ScaleB(1<<32) = (%s, %v), want +Inf with Overflow|Inexact", r.String(), f)
	}
	// Was: +1E+0 with no flags (math.MinInt64 truncated to 0).
	if r, f := one.ScaleB(math.MinInt64); !r.IsZero() || r.IsSignMinus() || f != FlagUnderflow|FlagInexact {
		t.Errorf("1.ScaleB(math.MinInt64) = (%s, %v), want +0 with Underflow|Inexact", r.String(), f)
	}
}

// TestScaleBWithModeDirectedOverflowReachesClampedScaler confirms the explicit
// rounding mode is carried through the clamped long-int scaler: round-toward-
// zero turns a saturating overflow into the largest finite magnitude instead
// of infinity, still flagging Overflow|Inexact. This proves ScaleBWithMode's
// mode argument reaches the port after the domain clamp, not just that the
// value channel changed.
func TestScaleBWithModeDirectedOverflowReachesClampedScaler(t *testing.T) {
	one := mustDecimal64BID(t, "1")
	r, f := one.ScaleBWithMode(int(1)<<40, RoundTowardZero)
	if r.IsInf() || !r.IsFinite() {
		t.Errorf("1.ScaleBWithMode(1<<40, towardZero) = %s, want largest finite (not infinity)", r.String())
	}
	if f != FlagOverflow|FlagInexact {
		t.Errorf("1.ScaleBWithMode(1<<40, towardZero) flags = %v, want Overflow|Inexact", f)
	}
}
