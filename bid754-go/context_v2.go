// context_v2.go - arithmetic context carrying rounding mode and flags.
package bid754

// ArithmeticContext carries a rounding mode and accumulated exception flags
// for callers that thread an explicit rounding policy through the *WithMode
// operations and accumulate the returned flags via SetFlag. Format and
// precision are implied by the BID value types themselves. An
// ArithmeticContext is not goroutine-safe; use one context per goroutine or
// add external synchronization. Flags are sticky across operations: between
// logically independent computations, clear them (ClearAllFlags, or
// SaveAllFlags/RestoreFlags) or start a fresh context, or an earlier
// computation's flags read as the later one's.
type ArithmeticContext struct {
	RoundingMode RoundingMode
	Flags        ExceptionFlags
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
// The mode is not validated here: a mode outside the defined constants is
// rejected by the *WithMode operation it is eventually passed to, through
// that operation's flag channel (a NaN-class result plus
// FlagInvalidOperation), not by panicking.
func (ctx *ArithmeticContext) WithRounding(mode RoundingMode) *ArithmeticContext {
	newCtx := ctx.Clone()
	newCtx.RoundingMode = mode
	return newCtx
}
