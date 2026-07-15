package bid754

import (
	"strconv"
	"sync"
	"testing"
)

func TestDefaultArithmeticContextReturnsSnapshot(t *testing.T) {
	previous := DefaultArithmeticContext().RoundingMode
	t.Cleanup(func() {
		SetDefaultRounding(previous)
	})

	SetDefaultRounding(RoundNearestEven)
	ctx := DefaultArithmeticContext()
	if ctx.RoundingMode != RoundNearestEven {
		t.Fatalf("default rounding = %s, want %s", ctx.RoundingMode, RoundNearestEven)
	}

	ctx.RoundingMode = RoundTowardZero
	if got := DefaultArithmeticContext().RoundingMode; got != RoundNearestEven {
		t.Fatalf("mutating returned default context changed global rounding to %s", got)
	}

	SetDefaultRounding(RoundTowardPositive)
	if got := DefaultArithmeticContext().RoundingMode; got != RoundTowardPositive {
		t.Fatalf("default rounding after SetDefaultRounding = %s, want %s", got, RoundTowardPositive)
	}
}

func TestDefaultArithmeticContextPreservesWideInvalidRoundingMode(t *testing.T) {
	if strconv.IntSize < 64 {
		t.Skip("wide RoundingMode truncation requires a 64-bit int")
	}

	previous := DefaultArithmeticContext().RoundingMode
	t.Cleanup(func() { SetDefaultRounding(previous) })

	one := mustDecimal64BID(t, "1")
	tests := []struct {
		name string
		raw  int64
	}{
		{name: "positive wraps to nearest even", raw: int64(1) << 32},
		{name: "positive wraps to toward negative", raw: (int64(1) << 32) + int64(RoundTowardNegative)},
		{name: "negative wraps to nearest even", raw: -(int64(1) << 32)},
		{name: "negative wraps to nearest away", raw: -(int64(1) << 32) + int64(RoundNearestAway)},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			bad := RoundingMode(tc.raw)
			SetDefaultRounding(bad)
			if got := DefaultArithmeticContext().RoundingMode; got != bad {
				t.Errorf("DefaultArithmeticContext().RoundingMode = %d after SetDefaultRounding(%d), want the unsupported mode preserved for rejection", got, bad)
			}
			if got := Add64BIDWithContext(one, one, nil); got != canonicalQNaN64BID() {
				t.Errorf("Add64BIDWithContext with wide invalid default %d = %s, want canonical qNaN", bad, got.String())
			}
		})
	}
}

func TestDefaultArithmeticContextConcurrentAccess(t *testing.T) {
	previous := DefaultArithmeticContext().RoundingMode
	t.Cleanup(func() {
		SetDefaultRounding(previous)
	})

	modes := []RoundingMode{
		RoundNearestEven,
		RoundNearestAway,
		RoundTowardZero,
		RoundTowardPositive,
		RoundTowardNegative,
	}

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(offset int) {
			defer wg.Done()
			for n := 0; n < 1000; n++ {
				SetDefaultRounding(modes[(n+offset)%len(modes)])
			}
		}(i)
	}
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for n := 0; n < 1000; n++ {
				ctx := DefaultArithmeticContext()
				_, _ = contextBIDRoundingMode(ctx)
				ctx.RoundingMode = RoundTowardZero
			}
		}()
	}
	wg.Wait()
}

func TestAdd32BIDWithContextUsesRoundingMode(t *testing.T) {
	a, err := NewDecimal32BIDDirect("9.999999")
	if err != nil {
		t.Fatalf("NewDecimal32BIDDirect(a): %v", err)
	}
	b, err := NewDecimal32BIDDirect("9.999999")
	if err != nil {
		t.Fatalf("NewDecimal32BIDDirect(b): %v", err)
	}

	nearest := Add32BIDWithContext(a, b, &ArithmeticContext{RoundingMode: RoundNearestEven})
	toZero := Add32BIDWithContext(a, b, &ArithmeticContext{RoundingMode: RoundTowardZero})

	if got := nearest.String(); got != "+2000000E-5" {
		t.Fatalf("nearest-even result = %q, want %q", got, "+2000000E-5")
	}
	if got := toZero.String(); got != "+1999999E-5" {
		t.Fatalf("toward-zero result = %q, want %q", got, "+1999999E-5")
	}
}

func TestAddBIDWithContextAccumulatesFlags(t *testing.T) {
	t.Run("decimal32", func(t *testing.T) {
		a := mustDecimal32BID(t, "9.999999")
		b := mustDecimal32BID(t, "9.999999")
		ctx := &ArithmeticContext{RoundingMode: RoundTowardZero, Flags: FlagDivisionByZero}

		_, wantFlags := decimal32BIDAddPortModeFlags(a, b, bidgoRoundingModeForTest(t, ctx.RoundingMode))
		if wantFlags == 0 {
			t.Fatal("test case did not produce flags")
		}

		Add32BIDWithContext(a, b, ctx)
		assertContextFlags(t, ctx, FlagDivisionByZero|wantFlags)
	})

	t.Run("decimal64", func(t *testing.T) {
		a := mustDecimal64BID(t, "9.999999999999999")
		b := mustDecimal64BID(t, "9.999999999999999")
		ctx := &ArithmeticContext{RoundingMode: RoundTowardZero, Flags: FlagDivisionByZero}

		_, wantFlags := decimal64BIDAddPortModeFlags(a, b, bidgoRoundingModeForTest(t, ctx.RoundingMode))
		if wantFlags == 0 {
			t.Fatal("test case did not produce flags")
		}

		Add64BIDWithContext(a, b, ctx)
		assertContextFlags(t, ctx, FlagDivisionByZero|wantFlags)
	})

	t.Run("decimal128", func(t *testing.T) {
		a := mustDecimal128BID(t, "9.999999999999999999999999999999999")
		b := mustDecimal128BID(t, "9.999999999999999999999999999999999")
		ctx := &ArithmeticContext{RoundingMode: RoundTowardZero, Flags: FlagDivisionByZero}

		_, wantFlags := decimal128BIDAddPortModeFlags(a, b, bidgoRoundingModeForTest(t, ctx.RoundingMode))
		if wantFlags == 0 {
			t.Fatal("test case did not produce flags")
		}

		Add128BIDWithContext(a, b, ctx)
		assertContextFlags(t, ctx, FlagDivisionByZero|wantFlags)
	})
}

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

// TestWithContextNilContextBadGlobalModeReturnsCanonicalQNaN pins the nil-ctx
// behavior when the global default rounding mode is outside the defined
// constants: the operation cannot resolve a rounding mode, so it rejects the
// call with a canonical quiet NaN result and must not panic. A nil context has
// no Flags field, so FlagInvalidOperation has nowhere to accumulate and is
// dropped -- the canonical quiet NaN is the only observable signal (this is the
// documented contract on SetDefaultRounding and the Add<w>BIDWithContext
// helpers). The contrast case confirms the flag is genuinely raised and only
// the nil-context case drops it.
func TestWithContextNilContextBadGlobalModeReturnsCanonicalQNaN(t *testing.T) {
	previous := DefaultArithmeticContext().RoundingMode
	t.Cleanup(func() { SetDefaultRounding(previous) })
	SetDefaultRounding(RoundingMode(99)) // outside the defined constants

	a32, b32 := mustDecimal32BID(t, "1"), mustDecimal32BID(t, "2")
	if got := Add32BIDWithContext(a32, b32, nil); got != canonicalQNaN32BID() {
		t.Errorf("Add32BIDWithContext(1,2,nil) with invalid global mode = %s, want canonical qNaN", got.String())
	}
	a64, b64 := mustDecimal64BID(t, "1"), mustDecimal64BID(t, "2")
	if got := Add64BIDWithContext(a64, b64, nil); got != canonicalQNaN64BID() {
		t.Errorf("Add64BIDWithContext(1,2,nil) with invalid global mode = %s, want canonical qNaN", got.String())
	}
	a128, b128 := mustDecimal128BID(t, "1"), mustDecimal128BID(t, "2")
	if got := Add128BIDWithContext(a128, b128, nil); got != canonicalQNaN128BID() {
		t.Errorf("Add128BIDWithContext(1,2,nil) with invalid global mode = %s, want canonical qNaN", got.String())
	}

	// A non-nil context DOES observe the invalid-mode rejection through its
	// flag channel, so the flag is genuinely raised; only the nil-context path
	// has nowhere to record it.
	ctx := &ArithmeticContext{RoundingMode: RoundingMode(99)}
	if got := Add64BIDWithContext(a64, b64, ctx); got != canonicalQNaN64BID() {
		t.Errorf("Add64BIDWithContext(1,2,badCtx) = %s, want canonical qNaN", got.String())
	}
	if ctx.Flags&FlagInvalidOperation == 0 {
		t.Errorf("Add64BIDWithContext(badCtx) flags = %v, want FlagInvalidOperation accumulated", ctx.Flags)
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

func assertContextFlags(t *testing.T, ctx *ArithmeticContext, want ExceptionFlags) {
	t.Helper()
	if ctx.Flags != want {
		t.Fatalf("context flags = %s, want %s", ctx.Flags, want)
	}
}
