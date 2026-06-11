// context_v2.go - arithmetic context carrying rounding mode and flags.
package bid754

import "sync/atomic"

// ArithmeticContext carries the rounding mode and accumulated exception
// flags for context-based operations. Format and precision are implied by
// the BID value types themselves.
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
func (ctx *ArithmeticContext) WithRounding(mode RoundingMode) *ArithmeticContext {
	newCtx := ctx.Clone()
	newCtx.RoundingMode = mode
	return newCtx
}

// === Global default context ===

var defaultArithmeticRoundingMode atomic.Int32

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
func SetDefaultRounding(mode RoundingMode) {
	defaultArithmeticRoundingMode.Store(int32(mode))
}

// === Context-based operations (optional) ===

// Most operations are value-type methods; these helpers exist for callers
// that need an explicit rounding mode with flag accumulation.

// Add32BIDWithContext returns a + b rounded with the context mode and
// accumulates the raised flags into ctx. A nil ctx uses the global default
// rounding mode.
func Add32BIDWithContext(a, b Decimal32BID, ctx *ArithmeticContext) Decimal32BID {
	result, flags := decimal32BIDAddPortModeFlags(a, b, contextBIDRoundingMode(ctx))
	accumulateContextFlags(ctx, flags)
	return result
}

// Add64BIDWithContext returns a + b rounded with the context mode and
// accumulates the raised flags into ctx. A nil ctx uses the global default
// rounding mode.
func Add64BIDWithContext(a, b Decimal64BID, ctx *ArithmeticContext) Decimal64BID {
	result, flags := decimal64BIDAddPortModeFlags(a, b, contextBIDRoundingMode(ctx))
	accumulateContextFlags(ctx, flags)
	return result
}

// Add128BIDWithContext returns a + b rounded with the context mode and
// accumulates the raised flags into ctx. A nil ctx uses the global default
// rounding mode.
func Add128BIDWithContext(a, b Decimal128BID, ctx *ArithmeticContext) Decimal128BID {
	result, flags := decimal128BIDAddPortModeFlags(a, b, contextBIDRoundingMode(ctx))
	accumulateContextFlags(ctx, flags)
	return result
}

func contextBIDRoundingMode(ctx *ArithmeticContext) int {
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
