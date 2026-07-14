package bid754

import bidgo "github.com/sky1core/bid754/bid754-go/internal/bidgo"

// Add64DQBIDWithMode returns the Decimal64 result of left + right, where left
// is Decimal64 and right is Decimal128, rounded with mode.
func Add64DQBIDWithMode(left Decimal64BID, right Decimal128BID, mode RoundingMode) (Decimal64BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN64BID(), FlagInvalidOperation
	}
	result, flags := bidgo.Bid64dqAdd(left.ToUint64(), decimal128BIDAsBidgo(right), rnd)
	return Decimal64BID(result), bidgoExceptionFlags(flags)
}

// Add64QDBIDWithMode returns the Decimal64 result of left + right, where left
// is Decimal128 and right is Decimal64, rounded with mode.
func Add64QDBIDWithMode(left Decimal128BID, right Decimal64BID, mode RoundingMode) (Decimal64BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN64BID(), FlagInvalidOperation
	}
	result, flags := bidgo.Bid64qdAdd(decimal128BIDAsBidgo(left), right.ToUint64(), rnd)
	return Decimal64BID(result), bidgoExceptionFlags(flags)
}

// Add64QQBIDWithMode returns the Decimal64 result of two Decimal128 operands,
// rounded once to Decimal64 with mode.
func Add64QQBIDWithMode(left, right Decimal128BID, mode RoundingMode) (Decimal64BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN64BID(), FlagInvalidOperation
	}
	result, flags := bidgo.Bid64qqAdd(decimal128BIDAsBidgo(left), decimal128BIDAsBidgo(right), rnd)
	return Decimal64BID(result), bidgoExceptionFlags(flags)
}

// Sub64DQBIDWithMode returns the Decimal64 result of left - right, where left
// is Decimal64 and right is Decimal128, rounded with mode.
func Sub64DQBIDWithMode(left Decimal64BID, right Decimal128BID, mode RoundingMode) (Decimal64BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN64BID(), FlagInvalidOperation
	}
	result, flags := bidgo.Bid64dqSub(left.ToUint64(), decimal128BIDAsBidgo(right), rnd)
	return Decimal64BID(result), bidgoExceptionFlags(flags)
}

// Sub64QDBIDWithMode returns the Decimal64 result of left - right, where left
// is Decimal128 and right is Decimal64, rounded with mode.
func Sub64QDBIDWithMode(left Decimal128BID, right Decimal64BID, mode RoundingMode) (Decimal64BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN64BID(), FlagInvalidOperation
	}
	result, flags := bidgo.Bid64qdSub(decimal128BIDAsBidgo(left), right.ToUint64(), rnd)
	return Decimal64BID(result), bidgoExceptionFlags(flags)
}

// Sub64QQBIDWithMode returns the Decimal64 result of two Decimal128 operands,
// rounded once to Decimal64 with mode.
func Sub64QQBIDWithMode(left, right Decimal128BID, mode RoundingMode) (Decimal64BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN64BID(), FlagInvalidOperation
	}
	result, flags := bidgo.Bid64qqSub(decimal128BIDAsBidgo(left), decimal128BIDAsBidgo(right), rnd)
	return Decimal64BID(result), bidgoExceptionFlags(flags)
}

// Mul64DQBIDWithMode returns the Decimal64 result of left * right, where left
// is Decimal64 and right is Decimal128, rounded with mode.
func Mul64DQBIDWithMode(left Decimal64BID, right Decimal128BID, mode RoundingMode) (Decimal64BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN64BID(), FlagInvalidOperation
	}
	result, flags := bidgo.Bid64dqMul(left.ToUint64(), decimal128BIDAsBidgo(right), rnd)
	return Decimal64BID(result), bidgoExceptionFlags(flags)
}

// Mul64QDBIDWithMode returns the Decimal64 result of left * right, where left
// is Decimal128 and right is Decimal64, rounded with mode.
func Mul64QDBIDWithMode(left Decimal128BID, right Decimal64BID, mode RoundingMode) (Decimal64BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN64BID(), FlagInvalidOperation
	}
	result, flags := bidgo.Bid64qdMul(decimal128BIDAsBidgo(left), right.ToUint64(), rnd)
	return Decimal64BID(result), bidgoExceptionFlags(flags)
}

// Mul64QQBIDWithMode returns the Decimal64 result of two Decimal128 operands,
// rounded once to Decimal64 with mode.
func Mul64QQBIDWithMode(left, right Decimal128BID, mode RoundingMode) (Decimal64BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN64BID(), FlagInvalidOperation
	}
	result, flags := bidgo.Bid64qqMul(decimal128BIDAsBidgo(left), decimal128BIDAsBidgo(right), rnd)
	return Decimal64BID(result), bidgoExceptionFlags(flags)
}

// Div64DQBIDWithMode returns the Decimal64 result of left / right, where left
// is Decimal64 and right is Decimal128, rounded with mode.
func Div64DQBIDWithMode(left Decimal64BID, right Decimal128BID, mode RoundingMode) (Decimal64BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN64BID(), FlagInvalidOperation
	}
	result, flags := bidgo.Bid64dqDiv(left.ToUint64(), decimal128BIDAsBidgo(right), rnd)
	return Decimal64BID(result), bidgoExceptionFlags(flags)
}

// Div64QDBIDWithMode returns the Decimal64 result of left / right, where left
// is Decimal128 and right is Decimal64, rounded with mode.
func Div64QDBIDWithMode(left Decimal128BID, right Decimal64BID, mode RoundingMode) (Decimal64BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN64BID(), FlagInvalidOperation
	}
	result, flags := bidgo.Bid64qdDiv(decimal128BIDAsBidgo(left), right.ToUint64(), rnd)
	return Decimal64BID(result), bidgoExceptionFlags(flags)
}

// Div64QQBIDWithMode returns the Decimal64 result of two Decimal128 operands,
// rounded once to Decimal64 with mode.
func Div64QQBIDWithMode(left, right Decimal128BID, mode RoundingMode) (Decimal64BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN64BID(), FlagInvalidOperation
	}
	result, flags := bidgo.Bid64qqDiv(decimal128BIDAsBidgo(left), decimal128BIDAsBidgo(right), rnd)
	return Decimal64BID(result), bidgoExceptionFlags(flags)
}

// Add128DDBIDWithMode returns the Decimal128 result of two Decimal64 operands,
// rounded with mode.
func Add128DDBIDWithMode(left, right Decimal64BID, mode RoundingMode) (Decimal128BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN128BID(), FlagInvalidOperation
	}
	result, flags := bidgo.Bid128ddAdd(left.ToUint64(), right.ToUint64(), rnd)
	return decimal128BIDFromBidgo(result), bidgoExceptionFlags(flags)
}

// Add128DQBIDWithMode returns the Decimal128 result of left + right, where
// left is Decimal64 and right is Decimal128, rounded with mode.
func Add128DQBIDWithMode(left Decimal64BID, right Decimal128BID, mode RoundingMode) (Decimal128BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN128BID(), FlagInvalidOperation
	}
	result, flags := bidgo.Bid128dqAdd(left.ToUint64(), decimal128BIDAsBidgo(right), rnd)
	return decimal128BIDFromBidgo(result), bidgoExceptionFlags(flags)
}

// Add128QDBIDWithMode returns the Decimal128 result of left + right, where
// left is Decimal128 and right is Decimal64, rounded with mode.
func Add128QDBIDWithMode(left Decimal128BID, right Decimal64BID, mode RoundingMode) (Decimal128BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN128BID(), FlagInvalidOperation
	}
	result, flags := bidgo.Bid128qdAdd(decimal128BIDAsBidgo(left), right.ToUint64(), rnd)
	return decimal128BIDFromBidgo(result), bidgoExceptionFlags(flags)
}

// Sub128DDBIDWithMode returns the Decimal128 result of left - right, where both
// operands are Decimal64, rounded with mode.
func Sub128DDBIDWithMode(left, right Decimal64BID, mode RoundingMode) (Decimal128BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN128BID(), FlagInvalidOperation
	}
	result, flags := bidgo.Bid128ddSub(left.ToUint64(), right.ToUint64(), rnd)
	return decimal128BIDFromBidgo(result), bidgoExceptionFlags(flags)
}

// Sub128DQBIDWithMode returns the Decimal128 result of left - right, where
// left is Decimal64 and right is Decimal128, rounded with mode.
func Sub128DQBIDWithMode(left Decimal64BID, right Decimal128BID, mode RoundingMode) (Decimal128BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN128BID(), FlagInvalidOperation
	}
	result, flags := bidgo.Bid128dqSub(left.ToUint64(), decimal128BIDAsBidgo(right), rnd)
	return decimal128BIDFromBidgo(result), bidgoExceptionFlags(flags)
}

// Sub128QDBIDWithMode returns the Decimal128 result of left - right, where
// left is Decimal128 and right is Decimal64, rounded with mode.
func Sub128QDBIDWithMode(left Decimal128BID, right Decimal64BID, mode RoundingMode) (Decimal128BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN128BID(), FlagInvalidOperation
	}
	result, flags := bidgo.Bid128qdSub(decimal128BIDAsBidgo(left), right.ToUint64(), rnd)
	return decimal128BIDFromBidgo(result), bidgoExceptionFlags(flags)
}

// Mul128DDBIDWithMode returns the Decimal128 result of two Decimal64 operands,
// rounded with mode.
func Mul128DDBIDWithMode(left, right Decimal64BID, mode RoundingMode) (Decimal128BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN128BID(), FlagInvalidOperation
	}
	result, flags := bidgo.Bid128ddMul(left.ToUint64(), right.ToUint64(), rnd)
	return decimal128BIDFromBidgo(result), bidgoExceptionFlags(flags)
}

// Mul128DQBIDWithMode returns the Decimal128 result of left * right, where
// left is Decimal64 and right is Decimal128, rounded with mode.
func Mul128DQBIDWithMode(left Decimal64BID, right Decimal128BID, mode RoundingMode) (Decimal128BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN128BID(), FlagInvalidOperation
	}
	result, flags := bidgo.Bid128dqMul(left.ToUint64(), decimal128BIDAsBidgo(right), rnd)
	return decimal128BIDFromBidgo(result), bidgoExceptionFlags(flags)
}

// Mul128QDBIDWithMode returns the Decimal128 result of left * right, where
// left is Decimal128 and right is Decimal64, rounded with mode.
func Mul128QDBIDWithMode(left Decimal128BID, right Decimal64BID, mode RoundingMode) (Decimal128BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN128BID(), FlagInvalidOperation
	}
	result, flags := bidgo.Bid128qdMul(decimal128BIDAsBidgo(left), right.ToUint64(), rnd)
	return decimal128BIDFromBidgo(result), bidgoExceptionFlags(flags)
}

// Div128DDBIDWithMode returns the Decimal128 result of two Decimal64 operands,
// rounded with mode.
func Div128DDBIDWithMode(left, right Decimal64BID, mode RoundingMode) (Decimal128BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN128BID(), FlagInvalidOperation
	}
	result, flags := bidgo.Bid128ddDiv(left.ToUint64(), right.ToUint64(), rnd)
	return decimal128BIDFromBidgo(result), bidgoExceptionFlags(flags)
}

// Div128DQBIDWithMode returns the Decimal128 result of left / right, where
// left is Decimal64 and right is Decimal128, rounded with mode.
func Div128DQBIDWithMode(left Decimal64BID, right Decimal128BID, mode RoundingMode) (Decimal128BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN128BID(), FlagInvalidOperation
	}
	result, flags := bidgo.Bid128dqDiv(left.ToUint64(), decimal128BIDAsBidgo(right), rnd)
	return decimal128BIDFromBidgo(result), bidgoExceptionFlags(flags)
}

// Div128QDBIDWithMode returns the Decimal128 result of left / right, where
// left is Decimal128 and right is Decimal64, rounded with mode.
func Div128QDBIDWithMode(left Decimal128BID, right Decimal64BID, mode RoundingMode) (Decimal128BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN128BID(), FlagInvalidOperation
	}
	result, flags := bidgo.Bid128qdDiv(decimal128BIDAsBidgo(left), right.ToUint64(), rnd)
	return decimal128BIDFromBidgo(result), bidgoExceptionFlags(flags)
}
