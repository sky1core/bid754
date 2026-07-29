package bid754

import (
	"testing"
)

func TestArithmeticContextSaveRestoreFlags(t *testing.T) {
	ctx := &ArithmeticContext{Flags: FlagInexact | FlagOverflow | FlagClamped}
	saved := ctx.SaveAllFlags()
	if saved != FlagInexact|FlagOverflow|FlagClamped {
		t.Fatalf("SaveAllFlags() = %s, want %s", saved, FlagInexact|FlagOverflow|FlagClamped)
	}
	if !saved.HasFlag(FlagOverflow | FlagInvalidOperation) {
		t.Fatalf("saved HasFlag mask = false, want true for any saved bit")
	}
	if saved.HasFlag(FlagDivisionByZero | FlagInvalidOperation) {
		t.Fatalf("saved HasFlag absent mask = true, want false")
	}

	ctx.ClearAllFlags()
	ctx.SetFlag(FlagDivisionByZero | FlagInvalidOperation)
	ctx.RestoreFlags(saved, FlagInexact|FlagOverflow|FlagInvalidOperation)

	want := FlagDivisionByZero | FlagInexact | FlagOverflow
	if ctx.Flags != want {
		t.Fatalf("RestoreFlags partial mask = %s, want %s", ctx.Flags, want)
	}
}

// TestArithmeticContextCarriesModeIntoWithModeOperations pins the carrier
// pattern that replaced the removed context-arithmetic helpers: the caller
// reads the context's rounding mode into a *WithMode operation and
// accumulates the returned flags into the context via SetFlag.
func TestArithmeticContextCarriesModeIntoWithModeOperations(t *testing.T) {
	a := mustDecimal32BID(t, "9.999999")
	b := mustDecimal32BID(t, "9.999999")

	nearestCtx := NewArithmeticContext()
	nearest, nearestFlags := a.AddWithMode(b, nearestCtx.RoundingMode)
	nearestCtx.SetFlag(nearestFlags)
	if got := nearest.String(); got != "+2000000E-5" {
		t.Fatalf("nearest-even result = %q, want %q", got, "+2000000E-5")
	}

	toZeroCtx := NewArithmeticContext().WithRounding(RoundTowardZero)
	toZero, toZeroFlags := a.AddWithMode(b, toZeroCtx.RoundingMode)
	toZeroCtx.SetFlag(toZeroFlags)
	if got := toZero.String(); got != "+1999999E-5" {
		t.Fatalf("toward-zero result = %q, want %q", got, "+1999999E-5")
	}

	if toZeroFlags == 0 {
		t.Fatal("test case did not produce flags")
	}
	if toZeroCtx.Flags != toZeroFlags {
		t.Fatalf("context flags = %s, want %s", toZeroCtx.Flags, toZeroFlags)
	}
}

func mustDecimal32BID(t *testing.T, s string) Decimal32BID {
	t.Helper()
	d, err := NewDecimal32BIDDirect(s)
	if err != nil {
		t.Fatalf("NewDecimal32BIDDirect(%q): %v", s, err)
	}
	return d
}

func mustDecimal64BID(t *testing.T, s string) Decimal64BID {
	t.Helper()
	d, err := NewDecimal64BIDDirect(s)
	if err != nil {
		t.Fatalf("NewDecimal64BIDDirect(%q): %v", s, err)
	}
	return d
}

func mustDecimal128BID(t *testing.T, s string) Decimal128BID {
	t.Helper()
	d, err := NewDecimal128BIDDirect(s)
	if err != nil {
		t.Fatalf("NewDecimal128BIDDirect(%q): %v", s, err)
	}
	return d
}
