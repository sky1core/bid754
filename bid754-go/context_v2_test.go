package bid754

import (
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
				_ = contextBIDRoundingMode(ctx)
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

	if got := nearest.String(); got != "+2.000000e1" {
		t.Fatalf("nearest-even result = %q, want %q", got, "+2.000000e1")
	}
	if got := toZero.String(); got != "+1.999999e1" {
		t.Fatalf("toward-zero result = %q, want %q", got, "+1.999999e1")
	}
}

func TestAddBIDWithContextAccumulatesFlags(t *testing.T) {
	t.Run("decimal32", func(t *testing.T) {
		a := mustDecimal32BID(t, "9.999999")
		b := mustDecimal32BID(t, "9.999999")
		ctx := &ArithmeticContext{RoundingMode: RoundTowardZero, Flags: FlagDivisionByZero}

		_, wantFlags := decimal32BIDAddPortModeFlags(a, b, bidgoRoundingMode(ctx.RoundingMode))
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

		_, wantFlags := decimal64BIDAddPortModeFlags(a, b, bidgoRoundingMode(ctx.RoundingMode))
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

		_, wantFlags := decimal128BIDAddPortModeFlags(a, b, bidgoRoundingMode(ctx.RoundingMode))
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
