// context_v2.go - arithmetic context carrying rounding mode and flags.
package bid754

import "sync/atomic"

// ArithmeticContext carries the rounding mode and accumulated exception
// flags for context-based operations. Format and precision are implied by
// the BID value types themselves. An ArithmeticContext is not goroutine-safe;
// use one context per goroutine or add external synchronization.
type ArithmeticContext struct {
	RoundingMode RoundingMode
	Flags        ExceptionFlags
	// Precision is implied by the value type. The BID public helpers route
	// through the Go mechanical port for all three widths.
}

// NewArithmeticContext returns a context with round-to-nearest-even and no
// raised flags.
func NewArithmeticContext() *ArithmeticContext {
	return &ArithmeticContext{
		RoundingMode: RoundNearestEven,
		Flags:        0,
	}
}

// SetFlag raises the given exception flags (IEEE 754-2019 5.7.4 raiseFlags).
func (ctx *ArithmeticContext) SetFlag(flag ExceptionFlags) {
	ctx.Flags |= flag
}

// ClearFlag lowers the given exception flags (IEEE 754-2019 5.7.4 lowerFlags).
func (ctx *ArithmeticContext) ClearFlag(flag ExceptionFlags) {
	ctx.Flags &^= flag
}

// HasFlag reports whether any of the given flags is raised (IEEE 754-2019
// 5.7.4 testFlags).
func (ctx *ArithmeticContext) HasFlag(flag ExceptionFlags) bool {
	return ctx.Flags&flag != 0
}

// ClearAllFlags lowers every exception flag (IEEE 754-2019 5.7.4 lowerFlags
// applied to the full flag group).
func (ctx *ArithmeticContext) ClearAllFlags() {
	ctx.Flags = 0
}

// SaveAllFlags returns a snapshot of the accumulated exception flags
// (IEEE 754-2019 5.7.4 saveAllFlags).
func (ctx *ArithmeticContext) SaveAllFlags() ExceptionFlags {
	return ctx.Flags
}

// RestoreFlags restores the flags selected by mask to their values in saved
// and preserves the rest (IEEE 754-2019 5.7.4 restoreFlags). The whole
// ExceptionFlags domain is public; no implicit masking is applied.
func (ctx *ArithmeticContext) RestoreFlags(saved ExceptionFlags, mask ExceptionFlags) {
	ctx.Flags = (ctx.Flags &^ mask) | (saved & mask)
}

// Clone returns a copy of the context.
func (ctx *ArithmeticContext) Clone() *ArithmeticContext {
	return &ArithmeticContext{
		RoundingMode: ctx.RoundingMode,
		Flags:        ctx.Flags,
	}
}

// WithRounding returns a copy of the context with the given rounding mode.
// The mode is validated when the context is used by an operation: a mode
// outside the defined constants makes that operation reject the call through
// its flag channel (canonical quiet NaN result plus FlagInvalidOperation
// accumulated into the context), not panic.
func (ctx *ArithmeticContext) WithRounding(mode RoundingMode) *ArithmeticContext {
	newCtx := ctx.Clone()
	newCtx.RoundingMode = mode
	return newCtx
}

// === Global default context ===

// RoundingMode has int width. Int64 preserves every value on Go targets where
// int is 32 or 64 bits, so an unsupported wide mode reaches validation intact;
// narrowing to Int32 would silently alias values such as 1<<32 to a valid mode.
var defaultArithmeticRoundingMode atomic.Int64

// DefaultArithmeticContext returns a context snapshotting the global default
// rounding mode, with no raised flags.
func DefaultArithmeticContext() *ArithmeticContext {
	return &ArithmeticContext{
		RoundingMode: defaultRoundingMode(),
		Flags:        0,
	}
}

// SetDefaultRounding atomically sets the global default rounding mode used
// by DefaultArithmeticContext and by context operations given a nil context.
// A mode outside the defined constants is stored as-is; a later operation that
// resolves it returns a canonical quiet NaN result rather than panicking. When
// that operation has a non-nil context, FlagInvalidOperation is accumulated
// into it; when the context is nil there is no flag field to accumulate into,
// so the flag is dropped and the canonical quiet NaN result is the only
// observable signal of the rejection.
func SetDefaultRounding(mode RoundingMode) {
	defaultArithmeticRoundingMode.Store(int64(mode))
}

// === Context-based operations (optional) ===

// Most operations are value-type methods; these helpers exist for callers
// that need an explicit rounding mode with flag accumulation. A resolved
// rounding mode outside the defined constants is rejected with a canonical
// quiet NaN result, not by panicking, and FlagInvalidOperation is accumulated
// into ctx. A nil ctx has no flag field: the raised flags (including this
// invalid-mode flag) are dropped, and the canonical quiet NaN result is the
// only observable signal of the rejection.

// Add32BIDWithContext returns a + b rounded with the context mode and
// accumulates the raised flags into ctx. A nil ctx uses the global default
// rounding mode and, having no flag field, drops every raised flag. A
// resolved rounding mode outside the defined constants is rejected with a
// canonical quiet NaN result (no panic); FlagInvalidOperation is accumulated
// into ctx when it is non-nil, and with a nil ctx the canonical quiet NaN is
// the only observable signal.
func Add32BIDWithContext(a, b Decimal32BID, ctx *ArithmeticContext) Decimal32BID {
	rnd, ok := contextBIDRoundingMode(ctx)
	if !ok {
		accumulateContextFlags(ctx, FlagInvalidOperation)
		return canonicalQNaN32BID()
	}
	result, flags := decimal32BIDAddPortModeFlags(a, b, rnd)
	accumulateContextFlags(ctx, flags)
	return result
}

// Add64BIDWithContext returns a + b rounded with the context mode and
// accumulates the raised flags into ctx. A nil ctx uses the global default
// rounding mode and, having no flag field, drops every raised flag. A
// resolved rounding mode outside the defined constants is rejected with a
// canonical quiet NaN result (no panic); FlagInvalidOperation is accumulated
// into ctx when it is non-nil, and with a nil ctx the canonical quiet NaN is
// the only observable signal.
func Add64BIDWithContext(a, b Decimal64BID, ctx *ArithmeticContext) Decimal64BID {
	rnd, ok := contextBIDRoundingMode(ctx)
	if !ok {
		accumulateContextFlags(ctx, FlagInvalidOperation)
		return canonicalQNaN64BID()
	}
	result, flags := decimal64BIDAddPortModeFlags(a, b, rnd)
	accumulateContextFlags(ctx, flags)
	return result
}

// Add128BIDWithContext returns a + b rounded with the context mode and
// accumulates the raised flags into ctx. A nil ctx uses the global default
// rounding mode and, having no flag field, drops every raised flag. A
// resolved rounding mode outside the defined constants is rejected with a
// canonical quiet NaN result (no panic); FlagInvalidOperation is accumulated
// into ctx when it is non-nil, and with a nil ctx the canonical quiet NaN is
// the only observable signal.
func Add128BIDWithContext(a, b Decimal128BID, ctx *ArithmeticContext) Decimal128BID {
	rnd, ok := contextBIDRoundingMode(ctx)
	if !ok {
		accumulateContextFlags(ctx, FlagInvalidOperation)
		return canonicalQNaN128BID()
	}
	result, flags := decimal128BIDAddPortModeFlags(a, b, rnd)
	accumulateContextFlags(ctx, flags)
	return result
}

// contextBIDRoundingMode resolves the bidgo-domain rounding mode for ctx (or
// the global default when ctx is nil). The bool is false when the resolved
// RoundingMode is outside the defined constants, letting the caller reject the
// operation through its flag channel instead of panicking.
func contextBIDRoundingMode(ctx *ArithmeticContext) (int, bool) {
	if ctx == nil {
		return bidgoRoundingMode(defaultRoundingMode())
	}
	return bidgoRoundingMode(ctx.RoundingMode)
}

func defaultRoundingMode() RoundingMode {
	return RoundingMode(defaultArithmeticRoundingMode.Load())
}

func accumulateContextFlags(ctx *ArithmeticContext, flags ExceptionFlags) {
	if ctx != nil {
		ctx.Flags |= flags
	}
}
