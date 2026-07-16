package bid754

import bidgo "github.com/sky1core/bid754/bid754-go/internal/bidgo"

// FMA64DDQBIDWithMode returns x*y+z fused once to Decimal64, with D/D/Q
// operand widths and the requested rounding mode.
func FMA64DDQBIDWithMode(x, y Decimal64BID, z Decimal128BID, mode RoundingMode) (Decimal64BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN64BID(), FlagInvalidOperation
	}
	result, flags := bidgo.Bid64ddqFma(x.ToUint64(), y.ToUint64(), decimal128BIDAsBidgo(z), rnd)
	return Decimal64BID(result), bidgoExceptionFlags(flags)
}

// FMA64DQDBIDWithMode returns x*y+z fused once to Decimal64, with D/Q/D
// operand widths and the requested rounding mode.
func FMA64DQDBIDWithMode(x Decimal64BID, y Decimal128BID, z Decimal64BID, mode RoundingMode) (Decimal64BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN64BID(), FlagInvalidOperation
	}
	result, flags := bidgo.Bid64dqdFma(x.ToUint64(), decimal128BIDAsBidgo(y), z.ToUint64(), rnd)
	return Decimal64BID(result), bidgoExceptionFlags(flags)
}

// FMA64DQQBIDWithMode returns x*y+z fused once to Decimal64, with D/Q/Q
// operand widths and the requested rounding mode.
func FMA64DQQBIDWithMode(x Decimal64BID, y, z Decimal128BID, mode RoundingMode) (Decimal64BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN64BID(), FlagInvalidOperation
	}
	result, flags := bidgo.Bid64dqqFma(x.ToUint64(), decimal128BIDAsBidgo(y), decimal128BIDAsBidgo(z), rnd)
	return Decimal64BID(result), bidgoExceptionFlags(flags)
}

// FMA64QDDBIDWithMode returns x*y+z fused once to Decimal64, with Q/D/D
// operand widths and the requested rounding mode.
func FMA64QDDBIDWithMode(x Decimal128BID, y, z Decimal64BID, mode RoundingMode) (Decimal64BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN64BID(), FlagInvalidOperation
	}
	result, flags := bidgo.Bid64qddFma(decimal128BIDAsBidgo(x), y.ToUint64(), z.ToUint64(), rnd)
	return Decimal64BID(result), bidgoExceptionFlags(flags)
}

// FMA64QDQBIDWithMode returns x*y+z fused once to Decimal64, with Q/D/Q
// operand widths and the requested rounding mode.
func FMA64QDQBIDWithMode(x Decimal128BID, y Decimal64BID, z Decimal128BID, mode RoundingMode) (Decimal64BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN64BID(), FlagInvalidOperation
	}
	result, flags := bidgo.Bid64qdqFma(decimal128BIDAsBidgo(x), y.ToUint64(), decimal128BIDAsBidgo(z), rnd)
	return Decimal64BID(result), bidgoExceptionFlags(flags)
}

// FMA64QQDBIDWithMode returns x*y+z fused once to Decimal64, with Q/Q/D
// operand widths and the requested rounding mode.
func FMA64QQDBIDWithMode(x, y Decimal128BID, z Decimal64BID, mode RoundingMode) (Decimal64BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN64BID(), FlagInvalidOperation
	}
	result, flags := bidgo.Bid64qqdFma(decimal128BIDAsBidgo(x), decimal128BIDAsBidgo(y), z.ToUint64(), rnd)
	return Decimal64BID(result), bidgoExceptionFlags(flags)
}

// FMA64QQQBIDWithMode returns x*y+z fused once to Decimal64 from three
// Decimal128 operands with the requested rounding mode.
func FMA64QQQBIDWithMode(x, y, z Decimal128BID, mode RoundingMode) (Decimal64BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN64BID(), FlagInvalidOperation
	}
	result, flags := bidgo.Bid64qqqFma(decimal128BIDAsBidgo(x), decimal128BIDAsBidgo(y), decimal128BIDAsBidgo(z), rnd)
	return Decimal64BID(result), bidgoExceptionFlags(flags)
}

// FMA128DDDBIDWithMode returns x*y+z fused to Decimal128 from three
// Decimal64 operands with the requested rounding mode.
func FMA128DDDBIDWithMode(x, y, z Decimal64BID, mode RoundingMode) (Decimal128BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN128BID(), FlagInvalidOperation
	}
	result, flags := bidgo.Bid128dddFma(x.ToUint64(), y.ToUint64(), z.ToUint64(), rnd)
	return decimal128BIDFromBidgo(result), bidgoExceptionFlags(flags)
}

// FMA128DDQBIDWithMode returns x*y+z fused to Decimal128, with D/D/Q operand
// widths and the requested rounding mode.
func FMA128DDQBIDWithMode(x, y Decimal64BID, z Decimal128BID, mode RoundingMode) (Decimal128BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN128BID(), FlagInvalidOperation
	}
	result, flags := bidgo.Bid128ddqFma(x.ToUint64(), y.ToUint64(), decimal128BIDAsBidgo(z), rnd)
	return decimal128BIDFromBidgo(result), bidgoExceptionFlags(flags)
}

// FMA128DQDBIDWithMode returns x*y+z fused to Decimal128, with D/Q/D operand
// widths and the requested rounding mode.
func FMA128DQDBIDWithMode(x Decimal64BID, y Decimal128BID, z Decimal64BID, mode RoundingMode) (Decimal128BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN128BID(), FlagInvalidOperation
	}
	result, flags := bidgo.Bid128dqdFma(x.ToUint64(), decimal128BIDAsBidgo(y), z.ToUint64(), rnd)
	return decimal128BIDFromBidgo(result), bidgoExceptionFlags(flags)
}

// FMA128DQQBIDWithMode returns x*y+z fused to Decimal128, with D/Q/Q operand
// widths and the requested rounding mode.
func FMA128DQQBIDWithMode(x Decimal64BID, y, z Decimal128BID, mode RoundingMode) (Decimal128BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN128BID(), FlagInvalidOperation
	}
	result, flags := bidgo.Bid128dqqFma(x.ToUint64(), decimal128BIDAsBidgo(y), decimal128BIDAsBidgo(z), rnd)
	return decimal128BIDFromBidgo(result), bidgoExceptionFlags(flags)
}

// FMA128QDDBIDWithMode returns x*y+z fused to Decimal128, with Q/D/D operand
// widths and the requested rounding mode.
func FMA128QDDBIDWithMode(x Decimal128BID, y, z Decimal64BID, mode RoundingMode) (Decimal128BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN128BID(), FlagInvalidOperation
	}
	result, flags := bidgo.Bid128qddFma(decimal128BIDAsBidgo(x), y.ToUint64(), z.ToUint64(), rnd)
	return decimal128BIDFromBidgo(result), bidgoExceptionFlags(flags)
}

// FMA128QDQBIDWithMode returns x*y+z fused to Decimal128, with Q/D/Q operand
// widths and the requested rounding mode.
func FMA128QDQBIDWithMode(x Decimal128BID, y Decimal64BID, z Decimal128BID, mode RoundingMode) (Decimal128BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN128BID(), FlagInvalidOperation
	}
	result, flags := bidgo.Bid128qdqFma(decimal128BIDAsBidgo(x), y.ToUint64(), decimal128BIDAsBidgo(z), rnd)
	return decimal128BIDFromBidgo(result), bidgoExceptionFlags(flags)
}

// FMA128QQDBIDWithMode returns x*y+z fused to Decimal128, with Q/Q/D operand
// widths and the requested rounding mode.
func FMA128QQDBIDWithMode(x, y Decimal128BID, z Decimal64BID, mode RoundingMode) (Decimal128BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN128BID(), FlagInvalidOperation
	}
	result, flags := bidgo.Bid128qqdFma(decimal128BIDAsBidgo(x), decimal128BIDAsBidgo(y), z.ToUint64(), rnd)
	return decimal128BIDFromBidgo(result), bidgoExceptionFlags(flags)
}

// Sqrt64QBIDWithMode returns the Decimal64 square root of a Decimal128
// operand with the requested rounding mode.
func Sqrt64QBIDWithMode(x Decimal128BID, mode RoundingMode) (Decimal64BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN64BID(), FlagInvalidOperation
	}
	result, flags := bidgo.Bid64qSqrt(decimal128BIDAsBidgo(x), rnd)
	return Decimal64BID(result), bidgoExceptionFlags(flags)
}

// Sqrt128DBIDWithMode returns the Decimal128 square root of a Decimal64
// operand with the requested rounding mode.
func Sqrt128DBIDWithMode(x Decimal64BID, mode RoundingMode) (Decimal128BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN128BID(), FlagInvalidOperation
	}
	result, flags := bidgo.Bid128dSqrt(x.ToUint64(), rnd)
	return decimal128BIDFromBidgo(result), bidgoExceptionFlags(flags)
}
