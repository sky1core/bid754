package bid754

// Explicit-rounding-mode arithmetic port wrappers.
//
// The public arithmetic WithMode value-type methods route through these
// thin wrappers, which accept a public RoundingMode, reject an out-of-range
// mode through the flag channel (strict, no panic/trap -- docs/SPEC.md:
// no public API path may panic/trap on unsupported input), and otherwise
// forward to the existing internal decimal<w>BID<Op>PortModeFlags port callers
// (types_bidgo_runtime.go), which already carry the bidgo-domain rounding
// integer into the Go mechanical port. No arithmetic is reproduced here; the
// underlying internal/bidgo port is unchanged.
//
// An invalid mode mirrors the same rejection value the flag-carrying conversion
// surfaces already use (types_bidgo_invalid_mode.go): the canonical quiet NaN of
// the operation's width plus FlagInvalidOperation, so no sentinel is guessed.
// This matches decimal<w>BIDToDecimal32Port/ToDecimal64Port, the pre-existing
// mode-accepting wrappers with no NaN input of their own result width to mirror.

func decimal32BIDAddModeFlags(d, other Decimal32BID, mode RoundingMode) (Decimal32BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN32BID(), FlagInvalidOperation
	}
	return decimal32BIDAddPortModeFlags(d, other, rnd)
}

func decimal32BIDSubModeFlags(d, other Decimal32BID, mode RoundingMode) (Decimal32BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN32BID(), FlagInvalidOperation
	}
	return decimal32BIDSubPortModeFlags(d, other, rnd)
}

func decimal32BIDMulModeFlags(d, other Decimal32BID, mode RoundingMode) (Decimal32BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN32BID(), FlagInvalidOperation
	}
	return decimal32BIDMulPortModeFlags(d, other, rnd)
}

func decimal32BIDDivModeFlags(d, other Decimal32BID, mode RoundingMode) (Decimal32BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN32BID(), FlagInvalidOperation
	}
	return decimal32BIDDivPortModeFlags(d, other, rnd)
}

func decimal32BIDQuantizeModeFlags(d, other Decimal32BID, mode RoundingMode) (Decimal32BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN32BID(), FlagInvalidOperation
	}
	return decimal32BIDQuantizePortModeFlags(d, other, rnd)
}

func decimal32BIDFMAModeFlags(d, mul, add Decimal32BID, mode RoundingMode) (Decimal32BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN32BID(), FlagInvalidOperation
	}
	return decimal32BIDFMAPortMode(d, mul, add, rnd)
}

func decimal32BIDSqrtModeFlags(d Decimal32BID, mode RoundingMode) (Decimal32BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN32BID(), FlagInvalidOperation
	}
	return decimal32BIDSqrtPortModeFlags(d, rnd)
}

func decimal32BIDRoundIntegralExactModeFlags(d Decimal32BID, mode RoundingMode) (Decimal32BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN32BID(), FlagInvalidOperation
	}
	return decimal32BIDRoundIntegralExactPortModeFlags(d, rnd)
}

func decimal32BIDScaleBModeFlags(d Decimal32BID, exponent int, mode RoundingMode) (Decimal32BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN32BID(), FlagInvalidOperation
	}
	return decimal32BIDScaleBPortModeFlags(d, exponent, rnd)
}

func decimal64BIDAddModeFlags(d, other Decimal64BID, mode RoundingMode) (Decimal64BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN64BID(), FlagInvalidOperation
	}
	return decimal64BIDAddPortModeFlags(d, other, rnd)
}

func decimal64BIDSubModeFlags(d, other Decimal64BID, mode RoundingMode) (Decimal64BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN64BID(), FlagInvalidOperation
	}
	return decimal64BIDSubPortModeFlags(d, other, rnd)
}

func decimal64BIDMulModeFlags(d, other Decimal64BID, mode RoundingMode) (Decimal64BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN64BID(), FlagInvalidOperation
	}
	return decimal64BIDMulPortModeFlags(d, other, rnd)
}

func decimal64BIDDivModeFlags(d, other Decimal64BID, mode RoundingMode) (Decimal64BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN64BID(), FlagInvalidOperation
	}
	return decimal64BIDDivPortModeFlags(d, other, rnd)
}

func decimal64BIDQuantizeModeFlags(d, other Decimal64BID, mode RoundingMode) (Decimal64BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN64BID(), FlagInvalidOperation
	}
	return decimal64BIDQuantizePortModeFlags(d, other, rnd)
}

func decimal64BIDFMAModeFlags(d, mul, add Decimal64BID, mode RoundingMode) (Decimal64BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN64BID(), FlagInvalidOperation
	}
	return decimal64BIDFMAPortMode(d, mul, add, rnd)
}

func decimal64BIDSqrtModeFlags(d Decimal64BID, mode RoundingMode) (Decimal64BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN64BID(), FlagInvalidOperation
	}
	return decimal64BIDSqrtPortModeFlags(d, rnd)
}

func decimal64BIDRoundIntegralExactModeFlags(d Decimal64BID, mode RoundingMode) (Decimal64BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN64BID(), FlagInvalidOperation
	}
	return decimal64BIDRoundIntegralExactPortModeFlags(d, rnd)
}

func decimal64BIDScaleBModeFlags(d Decimal64BID, exponent int, mode RoundingMode) (Decimal64BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN64BID(), FlagInvalidOperation
	}
	return decimal64BIDScaleBPortModeFlags(d, exponent, rnd)
}

func decimal128BIDAddModeFlags(d, other Decimal128BID, mode RoundingMode) (Decimal128BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN128BID(), FlagInvalidOperation
	}
	return decimal128BIDAddPortModeFlags(d, other, rnd)
}

func decimal128BIDSubModeFlags(d, other Decimal128BID, mode RoundingMode) (Decimal128BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN128BID(), FlagInvalidOperation
	}
	return decimal128BIDSubPortModeFlags(d, other, rnd)
}

func decimal128BIDMulModeFlags(d, other Decimal128BID, mode RoundingMode) (Decimal128BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN128BID(), FlagInvalidOperation
	}
	return decimal128BIDMulPortModeFlags(d, other, rnd)
}

func decimal128BIDDivModeFlags(d, other Decimal128BID, mode RoundingMode) (Decimal128BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN128BID(), FlagInvalidOperation
	}
	return decimal128BIDDivPortModeFlags(d, other, rnd)
}

func decimal128BIDQuantizeModeFlags(d, other Decimal128BID, mode RoundingMode) (Decimal128BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN128BID(), FlagInvalidOperation
	}
	return decimal128BIDQuantizePortModeFlags(d, other, rnd)
}

func decimal128BIDFMAModeFlags(d, mul, add Decimal128BID, mode RoundingMode) (Decimal128BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN128BID(), FlagInvalidOperation
	}
	return decimal128BIDFMAPortMode(d, mul, add, rnd)
}

func decimal128BIDSqrtModeFlags(d Decimal128BID, mode RoundingMode) (Decimal128BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN128BID(), FlagInvalidOperation
	}
	return decimal128BIDSqrtPortModeFlags(d, rnd)
}

func decimal128BIDRoundIntegralExactModeFlags(d Decimal128BID, mode RoundingMode) (Decimal128BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN128BID(), FlagInvalidOperation
	}
	return decimal128BIDRoundIntegralExactPortModeFlags(d, rnd)
}

func decimal128BIDScaleBModeFlags(d Decimal128BID, exponent int, mode RoundingMode) (Decimal128BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN128BID(), FlagInvalidOperation
	}
	return decimal128BIDScaleBPortModeFlags(d, exponent, rnd)
}
