package bid754

// Concurrency safety harness for the public BID surface. These tests only earn
// their value under `go test -race` (see the `make test-race` target and the
// CI native job step): without the race detector they merely exercise the code
// concurrently and would miss the data races they exist to catch. They assert
// no expected values; the oracle is the race detector plus the absence of a
// panic under contention.

import (
	"sync"
	"testing"
)

// concurrentValueSpecs are parsed once and then shared (read-only) across
// goroutines. The mix spans finite, zero, signed, wide-exponent, and special
// (NaN/Inf) encodings so the value-type methods run their special-case paths
// under contention too.
var concurrentValueSpecs = []string{
	"0", "1", "-1", "2", "0.5", "-0.25",
	"3.141592653589793", "-2.718281828459045",
	"1E10", "-1E-10", "9999999", "123456.789",
	"inf", "-inf", "nan", "snan",
}

func makeConcurrentDecimal32(t *testing.T) []Decimal32BID {
	t.Helper()
	out := make([]Decimal32BID, 0, len(concurrentValueSpecs))
	for _, s := range concurrentValueSpecs {
		if d, err := NewDecimal32BIDDirect(s); err == nil {
			out = append(out, d)
		}
	}
	if len(out) < 4 {
		t.Fatalf("decimal32 seed set too small: %d", len(out))
	}
	return out
}

func makeConcurrentDecimal64(t *testing.T) []Decimal64BID {
	t.Helper()
	out := make([]Decimal64BID, 0, len(concurrentValueSpecs))
	for _, s := range concurrentValueSpecs {
		if d, err := NewDecimal64BIDDirect(s); err == nil {
			out = append(out, d)
		}
	}
	if len(out) < 4 {
		t.Fatalf("decimal64 seed set too small: %d", len(out))
	}
	return out
}

func makeConcurrentDecimal128(t *testing.T) []Decimal128BID {
	t.Helper()
	out := make([]Decimal128BID, 0, len(concurrentValueSpecs))
	for _, s := range concurrentValueSpecs {
		if d, err := NewDecimal128BIDDirect(s); err == nil {
			out = append(out, d)
		}
	}
	if len(out) < 4 {
		t.Fatalf("decimal128 seed set too small: %d", len(out))
	}
	return out
}

// TestConcurrentValueTypeOperationsRaceFree drives the public value-type
// methods (Add/Sub/Mul/Div/Quantize/String/ConvertToInt64) plus the raw parse
// entrypoints from many goroutines over a shared, read-only value set. What it
// catches: if any of these methods gained hidden shared mutable state (a
// package-level scratch buffer, memoization cache, or in-place mutation of the
// value type instead of copy semantics), the race detector flags it here. A
// green run under -race is the evidence that the BID value types are safe to
// share by copy across goroutines.
func TestConcurrentValueTypeOperationsRaceFree(t *testing.T) {
	d32 := makeConcurrentDecimal32(t)
	d64 := makeConcurrentDecimal64(t)
	d128 := makeConcurrentDecimal128(t)

	const workers = 8
	const iterations = 2000

	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(base int) {
			defer wg.Done()
			for n := 0; n < iterations; n++ {
				i := (n + base) % len(d64)
				j := (n*3 + base) % len(d64)

				a64, b64 := d64[i], d64[j]
				_ = a64.Add(b64).String()
				_ = a64.Sub(b64).String()
				_ = a64.Mul(b64).String()
				_ = a64.Div(b64).String()
				_ = a64.Quantize(b64).String()
				_, _ = a64.ConvertToInt64(RoundTowardZero)

				k := (n + base) % len(d32)
				l := (n*3 + base) % len(d32)
				a32, b32 := d32[k], d32[l]
				_ = a32.Add(b32).String()
				_ = a32.Mul(b32).String()
				_ = a32.Div(b32).String()
				_ = a32.Quantize(b32).String()

				p := (n + base) % len(d128)
				q := (n*3 + base) % len(d128)
				a128, b128 := d128[p], d128[q]
				_ = a128.Add(b128).String()
				_ = a128.Mul(b128).String()
				_ = a128.Div(b128).String()

				// Raw parse channel: independent parses must not share state.
				spec := concurrentValueSpecs[(n+base)%len(concurrentValueSpecs)]
				pd64, _ := ParseDecimal64BIDRaw(spec)
				_ = pd64.String()
				pd32, _ := ParseDecimal32BIDRaw(spec)
				_ = pd32.String()
			}
		}(w)
	}
	wg.Wait()
}

// TestConcurrentPerGoroutineContextRaceFree runs the documented carrier
// pattern concurrently: each goroutine owns its own *ArithmeticContext,
// feeds its rounding mode into *WithMode operations, and accumulates the
// returned flags via SetFlag. Per-goroutine contexts must not touch shared
// state, so flag accumulation into a private context stays local and the
// race detector must stay silent.
func TestConcurrentPerGoroutineContextRaceFree(t *testing.T) {
	modes := []RoundingMode{
		RoundNearestEven,
		RoundNearestAway,
		RoundTowardZero,
		RoundTowardPositive,
		RoundTowardNegative,
	}

	a32 := mustDecimal32BID(t, "9.999999")
	b32 := mustDecimal32BID(t, "9.999999")
	a64 := mustDecimal64BID(t, "9.999999999999999")
	b64 := mustDecimal64BID(t, "9.999999999999999")
	a128 := mustDecimal128BID(t, "9.999999999999999999999999999999999")
	b128 := mustDecimal128BID(t, "9.999999999999999999999999999999999")

	const workers = 6
	const iterations = 2000

	var wg sync.WaitGroup

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(base int) {
			defer wg.Done()
			ctx := NewArithmeticContext().WithRounding(modes[base%len(modes)])
			for n := 0; n < iterations; n++ {
				v32, f32 := a32.AddWithMode(b32, ctx.RoundingMode)
				_ = v32
				ctx.SetFlag(f32)
				v64, f64 := a64.AddWithMode(b64, ctx.RoundingMode)
				_ = v64
				ctx.SetFlag(f64)
				v128, f128 := a128.AddWithMode(b128, ctx.RoundingMode)
				_ = v128
				ctx.SetFlag(f128)
				_ = ctx.SaveAllFlags()
			}
		}(w)
	}

	wg.Wait()
}
