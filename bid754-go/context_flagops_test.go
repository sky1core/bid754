package bid754

import "testing"

// ArithmeticContext.ClearFlag, Clone, and WithRounding are IEEE 754-2019 5.7.4
// context plumbing implemented directly in Go (no Intel port counterpart), so
// the public-API parity gate excludes them and no generated domain exercises
// them. These hand-written tests pin their semantics: masked flag lowering that
// preserves the rest, independent copies, and mode replacement that leaves the
// source untouched.

func TestArithmeticContextClearFlagLowersOnlySelected(t *testing.T) {
	ctx := NewArithmeticContext()
	ctx.SetFlag(FlagInvalidOperation | FlagOverflow | FlagInexact)

	ctx.ClearFlag(FlagOverflow)
	if ctx.HasFlag(FlagOverflow) {
		t.Fatalf("ClearFlag(FlagOverflow) left FlagOverflow raised: %#x", ctx.Flags)
	}
	if !ctx.HasFlag(FlagInvalidOperation) || !ctx.HasFlag(FlagInexact) {
		t.Fatalf("ClearFlag(FlagOverflow) lowered unrelated flags: %#x", ctx.Flags)
	}
	if want := FlagInvalidOperation | FlagInexact; ctx.Flags != want {
		t.Fatalf("Flags after single clear = %#x, want %#x", ctx.Flags, want)
	}

	// Clearing multiple flags at once, and clearing an unset flag is a no-op.
	ctx.ClearFlag(FlagInvalidOperation | FlagInexact)
	if ctx.Flags != 0 {
		t.Fatalf("Flags after clearing all raised = %#x, want 0", ctx.Flags)
	}
	ctx.ClearFlag(FlagUnderflow)
	if ctx.Flags != 0 {
		t.Fatalf("ClearFlag on an unset flag changed Flags to %#x, want 0", ctx.Flags)
	}
}

func TestArithmeticContextCloneIsIndependent(t *testing.T) {
	orig := NewArithmeticContext()
	orig.RoundingMode = RoundTowardPositive
	orig.SetFlag(FlagInexact)

	clone := orig.Clone()
	if clone == orig {
		t.Fatal("Clone returned the same pointer, not a copy")
	}
	if clone.RoundingMode != RoundTowardPositive || clone.Flags != FlagInexact {
		t.Fatalf("Clone = {%v, %#x}, want {RoundTowardPositive, %#x}", clone.RoundingMode, clone.Flags, FlagInexact)
	}

	// Mutating the clone must not touch the original.
	clone.SetFlag(FlagOverflow)
	clone.RoundingMode = RoundTowardZero
	if orig.HasFlag(FlagOverflow) {
		t.Fatalf("mutating clone flags affected original: %#x", orig.Flags)
	}
	if orig.RoundingMode != RoundTowardPositive {
		t.Fatalf("mutating clone rounding mode affected original: %v", orig.RoundingMode)
	}

	// And mutating the original must not touch the clone.
	orig.SetFlag(FlagInvalidOperation)
	if clone.HasFlag(FlagInvalidOperation) {
		t.Fatalf("mutating original flags affected clone: %#x", clone.Flags)
	}
}

func TestArithmeticContextWithRoundingCopiesAndPreserves(t *testing.T) {
	orig := NewArithmeticContext()
	orig.SetFlag(FlagUnderflow | FlagInexact)

	derived := orig.WithRounding(RoundTowardNegative)
	if derived == orig {
		t.Fatal("WithRounding returned the same pointer, not a copy")
	}
	if derived.RoundingMode != RoundTowardNegative {
		t.Fatalf("derived rounding mode = %v, want RoundTowardNegative", derived.RoundingMode)
	}
	// Flags are carried over unchanged.
	if want := FlagUnderflow | FlagInexact; derived.Flags != want {
		t.Fatalf("derived flags = %#x, want %#x (preserved)", derived.Flags, want)
	}
	// The source keeps its own mode and flags.
	if orig.RoundingMode != RoundNearestEven {
		t.Fatalf("WithRounding changed the source mode to %v, want RoundNearestEven", orig.RoundingMode)
	}
	if want := FlagUnderflow | FlagInexact; orig.Flags != want {
		t.Fatalf("WithRounding changed the source flags to %#x, want %#x", orig.Flags, want)
	}

	// The derived context is independent of the source.
	derived.SetFlag(FlagOverflow)
	if orig.HasFlag(FlagOverflow) {
		t.Fatalf("mutating WithRounding result affected source: %#x", orig.Flags)
	}
}
