package bid754

import (
	"encoding/binary"
	"fmt"
	"math"
	"strings"

	bidgo "github.com/sky1core/bid754/bid754-go/internal/bidgo"
)

const (
	bidgoRoundingNearestEven    = 0
	bidgoRoundingTowardNegative = 1
	bidgoRoundingTowardPositive = 2
	bidgoRoundingTowardZero     = 3
	bidgoRoundingNearestAway    = 4
)

// defaultBIDRoundingMode is the bidgo-domain rounding mode used when an
// operation does not take an explicit RoundingMode. It is defined in the
// bidgo constant domain so it stays correct even if the public RoundingMode
// enum order ever changes.
const defaultBIDRoundingMode = bidgoRoundingNearestEven

// bidgoRoundingMode maps a public RoundingMode to its bidgo-domain value. The
// bool reports whether mode is one of the five defined constants; it is false
// for any other value. Callers reject an invalid mode through their own
// declared failure channel instead of panicking (docs/SPEC.md: no public API
// path may panic/trap on unsupported input).
func bidgoRoundingMode(mode RoundingMode) (int, bool) {
	switch mode {
	case RoundNearestEven:
		return bidgoRoundingNearestEven, true
	case RoundNearestAway:
		return bidgoRoundingNearestAway, true
	case RoundTowardZero:
		return bidgoRoundingTowardZero, true
	case RoundTowardPositive:
		return bidgoRoundingTowardPositive, true
	case RoundTowardNegative:
		return bidgoRoundingTowardNegative, true
	default:
		return 0, false
	}
}

func bidgoExceptionFlags(flags uint32) ExceptionFlags {
	var result ExceptionFlags
	if flags&bidgo.BID_INEXACT_EXCEPTION != 0 {
		result |= FlagInexact
	}
	if flags&bidgo.BID_UNDERFLOW_EXCEPTION != 0 {
		result |= FlagUnderflow
	}
	if flags&bidgo.BID_OVERFLOW_EXCEPTION != 0 {
		result |= FlagOverflow
	}
	if flags&bidgo.BID_ZERO_DIVIDE_EXCEPTION != 0 {
		result |= FlagDivisionByZero
	}
	if flags&bidgo.BID_INVALID_EXCEPTION != 0 {
		result |= FlagInvalidOperation
	}
	return result
}

func decimal32BIDAddPort(d, other Decimal32BID) Decimal32BID {
	return decimal32BIDAddPortMode(d, other, defaultBIDRoundingMode)
}

func decimal32BIDAddPortFlags(d, other Decimal32BID) (Decimal32BID, ExceptionFlags) {
	return decimal32BIDAddPortModeFlags(d, other, defaultBIDRoundingMode)
}

func decimal32BIDAddPortMode(d, other Decimal32BID, rndMode int) Decimal32BID {
	return Decimal32BID(bidgo.Bid32Add(d.ToUint32(), other.ToUint32(), rndMode))
}

func decimal32BIDAddPortModeFlags(d, other Decimal32BID, rndMode int) (Decimal32BID, ExceptionFlags) {
	result, flags := bidgo.Bid32AddWithFlags(d.ToUint32(), other.ToUint32(), rndMode)
	return Decimal32BID(result), bidgoExceptionFlags(flags)
}

func decimal32BIDSubPort(d, other Decimal32BID) Decimal32BID {
	return Decimal32BID(bidgo.Bid32Sub(d.ToUint32(), other.ToUint32(), defaultBIDRoundingMode))
}

func decimal32BIDSubPortFlags(d, other Decimal32BID) (Decimal32BID, ExceptionFlags) {
	return decimal32BIDSubPortModeFlags(d, other, defaultBIDRoundingMode)
}

func decimal32BIDSubPortModeFlags(d, other Decimal32BID, rndMode int) (Decimal32BID, ExceptionFlags) {
	result, flags := bidgo.Bid32SubWithFlags(d.ToUint32(), other.ToUint32(), rndMode)
	return Decimal32BID(result), bidgoExceptionFlags(flags)
}

func decimal32BIDMulPort(d, other Decimal32BID) Decimal32BID {
	return Decimal32BID(bidgo.Bid32Mul(d.ToUint32(), other.ToUint32(), defaultBIDRoundingMode))
}

func decimal32BIDMulPortFlags(d, other Decimal32BID) (Decimal32BID, ExceptionFlags) {
	result, flags := bidgo.Bid32MulWithFlags(d.ToUint32(), other.ToUint32(), defaultBIDRoundingMode)
	return Decimal32BID(result), bidgoExceptionFlags(flags)
}

func decimal32BIDMulPortModeFlags(d, other Decimal32BID, rndMode int) (Decimal32BID, ExceptionFlags) {
	result, flags := bidgo.Bid32MulWithFlags(d.ToUint32(), other.ToUint32(), rndMode)
	return Decimal32BID(result), bidgoExceptionFlags(flags)
}

func decimal32BIDFMAPort(d, mul, add Decimal32BID) (Decimal32BID, ExceptionFlags) {
	return decimal32BIDFMAPortMode(d, mul, add, defaultBIDRoundingMode)
}

func decimal32BIDFMAPortMode(d, mul, add Decimal32BID, rndMode int) (Decimal32BID, ExceptionFlags) {
	result, flags := bidgo.Bid32Fma(d.ToUint32(), mul.ToUint32(), add.ToUint32(), rndMode)
	return Decimal32BID(result), bidgoExceptionFlags(flags)
}

func decimal32BIDSqrtPort(d Decimal32BID) (Decimal32BID, ExceptionFlags) {
	result, flags := bidgo.Bid32Sqrt(d.ToUint32(), defaultBIDRoundingMode)
	return Decimal32BID(result), bidgoExceptionFlags(flags)
}

func decimal32BIDSqrtPortModeFlags(d Decimal32BID, rndMode int) (Decimal32BID, ExceptionFlags) {
	result, flags := bidgo.Bid32Sqrt(d.ToUint32(), rndMode)
	return Decimal32BID(result), bidgoExceptionFlags(flags)
}

func decimal32BIDDivPort(d, other Decimal32BID) Decimal32BID {
	return Decimal32BID(bidgo.Bid32Div(d.ToUint32(), other.ToUint32(), defaultBIDRoundingMode))
}

func decimal32BIDDivPortFlags(d, other Decimal32BID) (Decimal32BID, ExceptionFlags) {
	result, flags := bidgo.Bid32DivWithFlags(d.ToUint32(), other.ToUint32(), defaultBIDRoundingMode)
	return Decimal32BID(result), bidgoExceptionFlags(flags)
}

func decimal32BIDDivPortModeFlags(d, other Decimal32BID, rndMode int) (Decimal32BID, ExceptionFlags) {
	result, flags := bidgo.Bid32DivWithFlags(d.ToUint32(), other.ToUint32(), rndMode)
	return Decimal32BID(result), bidgoExceptionFlags(flags)
}

func decimal32BIDRemainderPort(d, other Decimal32BID) (Decimal32BID, ExceptionFlags) {
	result, flags := bidgo.Bid32Rem(d.ToUint32(), other.ToUint32())
	return Decimal32BID(result), bidgoExceptionFlags(flags)
}

func decimal32BIDFmodPort(d, other Decimal32BID) (Decimal32BID, ExceptionFlags) {
	result, flags := bidgo.Bid32Fmod(d.ToUint32(), other.ToUint32())
	return Decimal32BID(result), bidgoExceptionFlags(flags)
}

func decimal32BIDQuantizePort(d, other Decimal32BID) Decimal32BID {
	result, _ := bidgo.Bid32Quantize(d.ToUint32(), other.ToUint32(), defaultBIDRoundingMode)
	return Decimal32BID(result)
}

func decimal32BIDQuantizePortFlags(d, other Decimal32BID) (Decimal32BID, ExceptionFlags) {
	result, flags := bidgo.Bid32Quantize(d.ToUint32(), other.ToUint32(), defaultBIDRoundingMode)
	return Decimal32BID(result), bidgoExceptionFlags(flags)
}

func decimal32BIDQuantizePortModeFlags(d, other Decimal32BID, rndMode int) (Decimal32BID, ExceptionFlags) {
	result, flags := bidgo.Bid32Quantize(d.ToUint32(), other.ToUint32(), rndMode)
	return Decimal32BID(result), bidgoExceptionFlags(flags)
}

func decimal32BIDSameQuantumPort(d, other Decimal32BID) bool {
	return bidgo.Bid32SameQuantum(d.ToUint32(), other.ToUint32())
}

func decimal32BIDMinNumPort(d, other Decimal32BID) (Decimal32BID, ExceptionFlags) {
	result, flags := bidgo.Bid32MinNumWithFlags(d.ToUint32(), other.ToUint32())
	return Decimal32BID(result), bidgoExceptionFlags(flags)
}

func decimal32BIDMaxNumPort(d, other Decimal32BID) (Decimal32BID, ExceptionFlags) {
	result, flags := bidgo.Bid32MaxNumWithFlags(d.ToUint32(), other.ToUint32())
	return Decimal32BID(result), bidgoExceptionFlags(flags)
}

func decimal32BIDMinNumMagPort(d, other Decimal32BID) (Decimal32BID, ExceptionFlags) {
	result, flags := bidgo.Bid32MinNumMagWithFlags(d.ToUint32(), other.ToUint32())
	return Decimal32BID(result), bidgoExceptionFlags(flags)
}

func decimal32BIDMaxNumMagPort(d, other Decimal32BID) (Decimal32BID, ExceptionFlags) {
	result, flags := bidgo.Bid32MaxNumMagWithFlags(d.ToUint32(), other.ToUint32())
	return Decimal32BID(result), bidgoExceptionFlags(flags)
}

func decimal32BIDCompareTotalPort(d, other Decimal32BID) int {
	return totalOrderComparison(bidgo.Bid32TotalOrder(d.ToUint32(), other.ToUint32()), bidgo.Bid32TotalOrder(other.ToUint32(), d.ToUint32()))
}

func decimal32BIDCompareTotalMagPort(d, other Decimal32BID) int {
	return totalOrderComparison(bidgo.Bid32TotalOrderMag(d.ToUint32(), other.ToUint32()), bidgo.Bid32TotalOrderMag(other.ToUint32(), d.ToUint32()))
}

func decimal32BIDRoundIntegralExactPort(d Decimal32BID) Decimal32BID {
	result, _ := bidgo.Bid32RoundIntegralExact(d.ToUint32(), defaultBIDRoundingMode)
	return Decimal32BID(result)
}

func decimal32BIDRoundIntegralExactPortFlags(d Decimal32BID) (Decimal32BID, ExceptionFlags) {
	result, flags := bidgo.Bid32RoundIntegralExact(d.ToUint32(), defaultBIDRoundingMode)
	return Decimal32BID(result), bidgoExceptionFlags(flags)
}

func decimal32BIDRoundIntegralExactPortModeFlags(d Decimal32BID, rndMode int) (Decimal32BID, ExceptionFlags) {
	result, flags := bidgo.Bid32RoundIntegralExact(d.ToUint32(), rndMode)
	return Decimal32BID(result), bidgoExceptionFlags(flags)
}

func decimal32BIDRoundIntegralNearestEvenPort(d Decimal32BID) (Decimal32BID, ExceptionFlags) {
	result, flags := bidgo.Bid32RoundIntegralNearestEven(d.ToUint32())
	return Decimal32BID(result), bidgoExceptionFlags(flags)
}

func decimal32BIDRoundIntegralNearestAwayPort(d Decimal32BID) (Decimal32BID, ExceptionFlags) {
	result, flags := bidgo.Bid32RoundIntegralNearestAway(d.ToUint32())
	return Decimal32BID(result), bidgoExceptionFlags(flags)
}

func decimal32BIDRoundIntegralZeroPort(d Decimal32BID) (Decimal32BID, ExceptionFlags) {
	result, flags := bidgo.Bid32RoundIntegralZero(d.ToUint32())
	return Decimal32BID(result), bidgoExceptionFlags(flags)
}

func decimal32BIDRoundIntegralPositivePort(d Decimal32BID) (Decimal32BID, ExceptionFlags) {
	result, flags := bidgo.Bid32RoundIntegralPositive(d.ToUint32())
	return Decimal32BID(result), bidgoExceptionFlags(flags)
}

func decimal32BIDRoundIntegralNegativePort(d Decimal32BID) (Decimal32BID, ExceptionFlags) {
	result, flags := bidgo.Bid32RoundIntegralNegative(d.ToUint32())
	return Decimal32BID(result), bidgoExceptionFlags(flags)
}

func decimal32BIDLogBPort(d Decimal32BID) (Decimal32BID, ExceptionFlags) {
	result, flags := bidgo.Bid32Logb(d.ToUint32())
	return Decimal32BID(result), bidgoExceptionFlags(flags)
}

func decimal32BIDScaleBPort(d Decimal32BID, exponent int) (Decimal32BID, ExceptionFlags) {
	result, flags := bidgo.Bid32ScalblnWithFlags(d.ToUint32(), int64(exponent), defaultBIDRoundingMode)
	return Decimal32BID(result), bidgoExceptionFlags(flags)
}

func decimal32BIDScaleBPortModeFlags(d Decimal32BID, exponent int, rndMode int) (Decimal32BID, ExceptionFlags) {
	result, flags := bidgo.Bid32ScalblnWithFlags(d.ToUint32(), int64(exponent), rndMode)
	return Decimal32BID(result), bidgoExceptionFlags(flags)
}

func decimal32BIDCopyPort(d Decimal32BID) Decimal32BID {
	return Decimal32BID(bidgo.Bid32Copy(d.ToUint32()))
}

func decimal32BIDAbsPort(d Decimal32BID) Decimal32BID {
	return Decimal32BID(bidgo.Bid32Abs(d.ToUint32()))
}

func decimal32BIDNegatePort(d Decimal32BID) Decimal32BID {
	return Decimal32BID(bidgo.Bid32Negate(d.ToUint32()))
}

func decimal32BIDCopySignPort(d, signSource Decimal32BID) Decimal32BID {
	return Decimal32BID(bidgo.Bid32CopySign(d.ToUint32(), signSource.ToUint32()))
}

func decimal32BIDStringPort(d Decimal32BID) string {
	if s, ok := formatDecimal32BIDNaN(d.ToUint32()); ok {
		return s
	}
	return bidgo.Bid32ToString(d.ToUint32())
}

func decimal32BIDToBinary32Port(d Decimal32BID, mode RoundingMode) (float32, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		bits, _ := bidgo.Bid32ToBinary32(bidNaNBits32, bidgoRoundingNearestEven)
		return math.Float32frombits(bits), FlagInvalidOperation
	}
	result, flags := bidgo.Bid32ToBinary32(d.ToUint32(), rnd)
	return math.Float32frombits(result), bidgoExceptionFlags(flags)
}

func decimal32BIDToBinary64Port(d Decimal32BID, mode RoundingMode) (float64, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		bits, _ := bidgo.Bid32ToBinary64(bidNaNBits32, bidgoRoundingNearestEven)
		return math.Float64frombits(bits), FlagInvalidOperation
	}
	result, flags := bidgo.Bid32ToBinary64(d.ToUint32(), rnd)
	return math.Float64frombits(result), bidgoExceptionFlags(flags)
}

func decimal32BIDToBinary128Port(d Decimal32BID, mode RoundingMode) (Binary128, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		bits, _ := bidgo.Bid32ToBinary128(bidNaNBits32, bidgoRoundingNearestEven)
		return binary128FromBidgo(bits), FlagInvalidOperation
	}
	result, flags := bidgo.Bid32ToBinary128(d.ToUint32(), rnd)
	return binary128FromBidgo(result), bidgoExceptionFlags(flags)
}

// decimal32BIDToDecimal128Port and decimal64BIDToDecimal128Port are the two
// format-widening conversions to Decimal128. Both write their result through
// decimal128BIDSetBidgo rather than assigning the value form, which drops the
// encode's store/reload round trip. That is worth measuring only where the
// encode is a large enough share of the operation: it moves the bid64 row and
// is flat on the bid32 row, which does more work per call. Both use the
// pointer form so the pair stays uniform. Every other path that uses the
// value-form encode still pays that round trip, which inlining does not
// remove: it is negligible against the arithmetic operations, and
// taking it off the remaining cheap ones is unmeasured work, not a decision
// taken here.
func decimal32BIDToDecimal128Port(d Decimal32BID) (out Decimal128BID, flags ExceptionFlags) {
	result, rawFlags := bidgo.Bid32ToBid128(d.ToUint32())
	decimal128BIDSetBidgo(&out, result)
	return out, bidgoExceptionFlags(rawFlags)
}

func decimal32BIDToDecimal64Port(d Decimal32BID) (Decimal64BID, ExceptionFlags) {
	result, flags := bidgo.Bid32ToBid64(d.ToUint32())
	return Decimal64BID(result), bidgoExceptionFlags(flags)
}

func decimal32BIDNextTowardPort(d Decimal32BID, target Decimal128BID) (Decimal32BID, ExceptionFlags) {
	result, flags := bidgo.Bid32NextToward(d.ToUint32(), decimal128BIDAsBidgo(target))
	return Decimal32BID(result), bidgoExceptionFlags(flags)
}

func decimal32BIDNextPlusPort(d Decimal32BID) (Decimal32BID, ExceptionFlags) {
	result, flags := bidgo.Bid32NextUp(d.ToUint32())
	return Decimal32BID(result), bidgoExceptionFlags(flags)
}

func decimal32BIDNextMinusPort(d Decimal32BID) (Decimal32BID, ExceptionFlags) {
	result, flags := bidgo.Bid32NextDown(d.ToUint32())
	return Decimal32BID(result), bidgoExceptionFlags(flags)
}

func decimal32BIDPrettyStringPort(d Decimal32BID) string {
	return formatPrettyBIDString(decimal32BIDStringPort(d), decimal32Precision)
}

func decimal32BIDIsZeroPort(d Decimal32BID) bool {
	return bidgo.Bid32IsZero(d.ToUint32())
}

func decimal32BIDIsNaNPort(d Decimal32BID) bool {
	return bidgo.Bid32IsNaN(d.ToUint32())
}

func decimal32BIDIsInfPort(d Decimal32BID) bool {
	return bidgo.Bid32IsInf(d.ToUint32())
}

func decimal32BIDIsNormalPort(d Decimal32BID) bool {
	return bidgo.Bid32IsNormal(d.ToUint32()) != 0
}

func decimal32BIDIsFinitePort(d Decimal32BID) bool {
	return bidgo.Bid32IsFinite(d.ToUint32()) != 0
}

func decimal32BIDIsSubnormalPort(d Decimal32BID) bool {
	return bidgo.Bid32IsSubnormal(d.ToUint32()) != 0
}

func decimal32BIDIsSignalingPort(d Decimal32BID) bool {
	return bidgo.Bid32IsSignaling(d.ToUint32()) != 0
}

func decimal32BIDIsCanonicalPort(d Decimal32BID) bool {
	return bidgo.Bid32IsCanonical(d.ToUint32()) != 0
}

func decimal32BIDRadixPort() int {
	return bidgo.Bid32Radix()
}

func decimal32BIDIsSignMinusPort(d Decimal32BID) bool {
	return bidgo.Bid32IsSigned(d.ToUint32()) != 0
}

func decimal32BIDClassPort(d Decimal32BID) DecimalClass {
	return decimalClassFromBIDClass(bidgo.Bid32Class(d.ToUint32()))
}

func decimal32BIDSignPort(d Decimal32BID) int {
	if decimal32BIDIsZeroPort(d) {
		return 0
	}
	if bidgo.Bid32IsSigned(d.ToUint32()) != 0 {
		return -1
	}
	return 1
}

func newDecimal32BIDDirectPort(s string) (Decimal32BID, error) {
	result, flags := parseDecimal32BIDPort(s)
	if rejectedBIDStringInput(flags) || unrepresentableBIDStringFlags(flags) {
		return 0, fmt.Errorf("invalid decimal string: %s", s)
	}
	return result, nil
}

func newDecimal32BIDWithFlagsPort(s string) (Decimal32BID, ExceptionFlags, error) {
	result, flags := parseDecimal32BIDPort(s)
	if rejectedBIDStringInput(flags) {
		return 0, 0, fmt.Errorf("invalid decimal string: %s", s)
	}
	return result, flags, nil
}

// newDecimal32BIDWithModePort mirrors newDecimal32BIDWithFlagsPort with the
// caller's rounding mode carried into the parse. When the mode is invalid, the
// WithFlags input-rejection contract is checked first so an invalid mode cannot
// mask a malformed or unrepresentable string. An accepted string plus an
// invalid mode is rejected through the repo-wide flag channel (canonical quiet
// NaN + FlagInvalidOperation, nil error); a rejected string keeps the ordinary
// zero value, zero flags, non-nil error result.
func newDecimal32BIDWithModePort(s string, mode RoundingMode) (Decimal32BID, ExceptionFlags, error) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		if _, _, err := newDecimal32BIDWithFlagsPort(s); err != nil {
			return 0, 0, err
		}
		return canonicalQNaN32BID(), FlagInvalidOperation, nil
	}
	result, flags := parseDecimal32BIDPublicMode(s, rnd)
	if rejectedBIDStringInput(flags) {
		return 0, 0, fmt.Errorf("invalid decimal string: %s", s)
	}
	return result, flags, nil
}

func parseDecimal32BIDPort(s string) (Decimal32BID, ExceptionFlags) {
	return parseDecimal32BIDPublicMode(s, defaultBIDRoundingMode)
}

// parseDecimal32BIDPublicMode adds the public complete-syntax and
// no-silent-coercion contracts to the raw mechanical-port conversion. The
// generated decTest loader calls parseDecimal32BIDPortMode directly because
// IBM operands are converted under their declared context and accumulate
// conversion status as part of the case.
//
// The finite arm scans the input once: for a non-NaN result,
// invalidBIDStringInput reduces to !validBIDFiniteLiteral because a valid
// finite literal always contains a digit or an infinity spelling, so the
// whitespace-only arm can never accept on its own; the same parsed literal
// then feeds the silent-cohort check that previously re-parsed the input.
func parseDecimal32BIDPublicMode(s string, rndMode int) (Decimal32BID, ExceptionFlags) {
	result, flags, rawStatus := parseDecimal32BIDPortModeWithRawStatus(s, rndMode)
	if bidgo.Bid32IsNaN(result.ToUint32()) {
		if invalidBIDStringInput(s, true) {
			return canonicalQNaN32BID(), FlagInvalidOperation
		}
		return result, flags
	}
	literal, ok := parseBIDFiniteLiteral(s)
	if !ok {
		return canonicalQNaN32BID(), FlagInvalidOperation
	}
	if rawStatus == bidgo.BID_EXACT_STATUS && literal.cohortUnrepresentable(decimal32MinQuantum, decimal32MaxQuantum, decimal32Precision) {
		return canonicalQNaN32BID(), FlagInvalidOperation
	}
	return result, flags
}

func parseDecimal32BIDPortMode(s string, rndMode int) (Decimal32BID, ExceptionFlags) {
	result, flags, _ := parseDecimal32BIDPortModeWithRawStatus(s, rndMode)
	return result, flags
}

func parseDecimal32BIDPortModeWithRawStatus(s string, rndMode int) (Decimal32BID, ExceptionFlags, uint32) {
	if result, ok := parseDecimal32BIDNaN(s); ok {
		return result, 0, bidgo.BID_EXACT_STATUS
	}
	if _, ok := parseBIDNaNLiteral(s); ok {
		return canonicalQNaN32BID(), FlagInvalidOperation, bidgo.BID_INVALID_EXCEPTION
	}
	result, rawFlags := bidgo.Bid32FromStringRaw(s, rndMode)
	return Decimal32BID(result), bidgoExceptionFlags(rawFlags), rawFlags
}

func decimal64BIDAddPort(d, other Decimal64BID) Decimal64BID {
	return decimal64BIDAddPortMode(d, other, defaultBIDRoundingMode)
}

func decimal64BIDAddPortFlags(d, other Decimal64BID) (Decimal64BID, ExceptionFlags) {
	return decimal64BIDAddPortModeFlags(d, other, defaultBIDRoundingMode)
}

func decimal64BIDAddPortMode(d, other Decimal64BID, rndMode int) Decimal64BID {
	return Decimal64BID(bidgo.Bid64Add(d.ToUint64(), other.ToUint64(), rndMode))
}

func decimal64BIDAddPortModeFlags(d, other Decimal64BID, rndMode int) (Decimal64BID, ExceptionFlags) {
	result, flags := bidgo.Bid64AddWithFlags(d.ToUint64(), other.ToUint64(), rndMode)
	return Decimal64BID(result), bidgoExceptionFlags(flags)
}

func decimal64BIDSubPort(d, other Decimal64BID) Decimal64BID {
	return Decimal64BID(bidgo.Bid64Sub(d.ToUint64(), other.ToUint64(), defaultBIDRoundingMode))
}

func decimal64BIDSubPortFlags(d, other Decimal64BID) (Decimal64BID, ExceptionFlags) {
	return decimal64BIDSubPortModeFlags(d, other, defaultBIDRoundingMode)
}

func decimal64BIDSubPortModeFlags(d, other Decimal64BID, rndMode int) (Decimal64BID, ExceptionFlags) {
	result, flags := bidgo.Bid64SubWithFlags(d.ToUint64(), other.ToUint64(), rndMode)
	return Decimal64BID(result), bidgoExceptionFlags(flags)
}

func decimal64BIDMulPort(d, other Decimal64BID) Decimal64BID {
	return Decimal64BID(bidgo.Bid64Mul(d.ToUint64(), other.ToUint64(), defaultBIDRoundingMode))
}

func decimal64BIDMulPortFlags(d, other Decimal64BID) (Decimal64BID, ExceptionFlags) {
	result, flags := bidgo.Bid64MulWithFlags(d.ToUint64(), other.ToUint64(), defaultBIDRoundingMode)
	return Decimal64BID(result), bidgoExceptionFlags(flags)
}

func decimal64BIDMulPortModeFlags(d, other Decimal64BID, rndMode int) (Decimal64BID, ExceptionFlags) {
	result, flags := bidgo.Bid64MulWithFlags(d.ToUint64(), other.ToUint64(), rndMode)
	return Decimal64BID(result), bidgoExceptionFlags(flags)
}

func decimal64BIDFMAPort(d, mul, add Decimal64BID) (Decimal64BID, ExceptionFlags) {
	return decimal64BIDFMAPortMode(d, mul, add, defaultBIDRoundingMode)
}

func decimal64BIDFMAPortMode(d, mul, add Decimal64BID, rndMode int) (Decimal64BID, ExceptionFlags) {
	result, flags := bidgo.Bid64Fma(d.ToUint64(), mul.ToUint64(), add.ToUint64(), rndMode)
	return Decimal64BID(result), bidgoExceptionFlags(flags)
}

func decimal64BIDSqrtPort(d Decimal64BID) (Decimal64BID, ExceptionFlags) {
	result, flags := bidgo.Bid64Sqrt(d.ToUint64(), defaultBIDRoundingMode)
	return Decimal64BID(result), bidgoExceptionFlags(flags)
}

func decimal64BIDSqrtPortModeFlags(d Decimal64BID, rndMode int) (Decimal64BID, ExceptionFlags) {
	result, flags := bidgo.Bid64Sqrt(d.ToUint64(), rndMode)
	return Decimal64BID(result), bidgoExceptionFlags(flags)
}

func decimal64BIDDivPort(d, other Decimal64BID) Decimal64BID {
	return Decimal64BID(bidgo.Bid64Div(d.ToUint64(), other.ToUint64(), defaultBIDRoundingMode))
}

func decimal64BIDDivPortFlags(d, other Decimal64BID) (Decimal64BID, ExceptionFlags) {
	result, flags := bidgo.Bid64DivWithFlags(d.ToUint64(), other.ToUint64(), defaultBIDRoundingMode)
	return Decimal64BID(result), bidgoExceptionFlags(flags)
}

func decimal64BIDDivPortModeFlags(d, other Decimal64BID, rndMode int) (Decimal64BID, ExceptionFlags) {
	result, flags := bidgo.Bid64DivWithFlags(d.ToUint64(), other.ToUint64(), rndMode)
	return Decimal64BID(result), bidgoExceptionFlags(flags)
}

func decimal64BIDRemainderPort(d, other Decimal64BID) (Decimal64BID, ExceptionFlags) {
	result, flags := bidgo.Bid64Rem(d.ToUint64(), other.ToUint64())
	return Decimal64BID(result), bidgoExceptionFlags(flags)
}

func decimal64BIDFmodPort(d, other Decimal64BID) (Decimal64BID, ExceptionFlags) {
	result, flags := bidgo.Bid64Fmod(d.ToUint64(), other.ToUint64())
	return Decimal64BID(result), bidgoExceptionFlags(flags)
}

func decimal64BIDQuantizePort(d, other Decimal64BID) Decimal64BID {
	result, _ := bidgo.Bid64Quantize(d.ToUint64(), other.ToUint64(), defaultBIDRoundingMode)
	return Decimal64BID(result)
}

func decimal64BIDQuantizePortFlags(d, other Decimal64BID) (Decimal64BID, ExceptionFlags) {
	result, flags := bidgo.Bid64Quantize(d.ToUint64(), other.ToUint64(), defaultBIDRoundingMode)
	return Decimal64BID(result), bidgoExceptionFlags(flags)
}

func decimal64BIDQuantizePortModeFlags(d, other Decimal64BID, rndMode int) (Decimal64BID, ExceptionFlags) {
	result, flags := bidgo.Bid64Quantize(d.ToUint64(), other.ToUint64(), rndMode)
	return Decimal64BID(result), bidgoExceptionFlags(flags)
}

func decimal64BIDSameQuantumPort(d, other Decimal64BID) bool {
	return bidgo.Bid64SameQuantum(d.ToUint64(), other.ToUint64()) != 0
}

func decimal64BIDMinNumPort(d, other Decimal64BID) (Decimal64BID, ExceptionFlags) {
	result, flags := bidgo.Bid64MinNum(d.ToUint64(), other.ToUint64())
	return Decimal64BID(result), bidgoExceptionFlags(flags)
}

func decimal64BIDMaxNumPort(d, other Decimal64BID) (Decimal64BID, ExceptionFlags) {
	result, flags := bidgo.Bid64MaxNum(d.ToUint64(), other.ToUint64())
	return Decimal64BID(result), bidgoExceptionFlags(flags)
}

func decimal64BIDMinNumMagPort(d, other Decimal64BID) (Decimal64BID, ExceptionFlags) {
	result, flags := bidgo.Bid64MinNumMag(d.ToUint64(), other.ToUint64())
	return Decimal64BID(result), bidgoExceptionFlags(flags)
}

func decimal64BIDMaxNumMagPort(d, other Decimal64BID) (Decimal64BID, ExceptionFlags) {
	result, flags := bidgo.Bid64MaxNumMag(d.ToUint64(), other.ToUint64())
	return Decimal64BID(result), bidgoExceptionFlags(flags)
}

func decimal64BIDCompareTotalPort(d, other Decimal64BID) int {
	return totalOrderComparison(bidgo.Bid64TotalOrder(d.ToUint64(), other.ToUint64()), bidgo.Bid64TotalOrder(other.ToUint64(), d.ToUint64()))
}

func decimal64BIDCompareTotalMagPort(d, other Decimal64BID) int {
	return totalOrderComparison(bidgo.Bid64TotalOrderMag(d.ToUint64(), other.ToUint64()), bidgo.Bid64TotalOrderMag(other.ToUint64(), d.ToUint64()))
}

func decimal64BIDRoundIntegralExactPort(d Decimal64BID) Decimal64BID {
	result, _ := bidgo.Bid64RoundIntegralExact(d.ToUint64(), defaultBIDRoundingMode)
	return Decimal64BID(result)
}

func decimal64BIDRoundIntegralExactPortFlags(d Decimal64BID) (Decimal64BID, ExceptionFlags) {
	result, flags := bidgo.Bid64RoundIntegralExact(d.ToUint64(), defaultBIDRoundingMode)
	return Decimal64BID(result), bidgoExceptionFlags(flags)
}

func decimal64BIDRoundIntegralExactPortModeFlags(d Decimal64BID, rndMode int) (Decimal64BID, ExceptionFlags) {
	result, flags := bidgo.Bid64RoundIntegralExact(d.ToUint64(), rndMode)
	return Decimal64BID(result), bidgoExceptionFlags(flags)
}

func decimal64BIDRoundIntegralNearestEvenPort(d Decimal64BID) (Decimal64BID, ExceptionFlags) {
	result, flags := bidgo.Bid64RoundIntegralNearestEven(d.ToUint64())
	return Decimal64BID(result), bidgoExceptionFlags(flags)
}

func decimal64BIDRoundIntegralNearestAwayPort(d Decimal64BID) (Decimal64BID, ExceptionFlags) {
	result, flags := bidgo.Bid64RoundIntegralNearestAway(d.ToUint64())
	return Decimal64BID(result), bidgoExceptionFlags(flags)
}

func decimal64BIDRoundIntegralZeroPort(d Decimal64BID) (Decimal64BID, ExceptionFlags) {
	result, flags := bidgo.Bid64RoundIntegralZero(d.ToUint64())
	return Decimal64BID(result), bidgoExceptionFlags(flags)
}

func decimal64BIDRoundIntegralPositivePort(d Decimal64BID) (Decimal64BID, ExceptionFlags) {
	result, flags := bidgo.Bid64RoundIntegralPositive(d.ToUint64())
	return Decimal64BID(result), bidgoExceptionFlags(flags)
}

func decimal64BIDRoundIntegralNegativePort(d Decimal64BID) (Decimal64BID, ExceptionFlags) {
	result, flags := bidgo.Bid64RoundIntegralNegative(d.ToUint64())
	return Decimal64BID(result), bidgoExceptionFlags(flags)
}

func decimal64BIDLogBPort(d Decimal64BID) (Decimal64BID, ExceptionFlags) {
	result, flags := bidgo.Bid64Logb(d.ToUint64())
	return Decimal64BID(result), bidgoExceptionFlags(flags)
}

func decimal64BIDScaleBPort(d Decimal64BID, exponent int) (Decimal64BID, ExceptionFlags) {
	result, flags := bidgo.Bid64Scalbln(d.ToUint64(), int64(exponent), defaultBIDRoundingMode)
	return Decimal64BID(result), bidgoExceptionFlags(flags)
}

func decimal64BIDScaleBPortModeFlags(d Decimal64BID, exponent int, rndMode int) (Decimal64BID, ExceptionFlags) {
	result, flags := bidgo.Bid64Scalbln(d.ToUint64(), int64(exponent), rndMode)
	return Decimal64BID(result), bidgoExceptionFlags(flags)
}

func decimal64BIDCopyPort(d Decimal64BID) Decimal64BID {
	return Decimal64BID(bidgo.Bid64Copy(d.ToUint64()))
}

func decimal64BIDAbsPort(d Decimal64BID) Decimal64BID {
	return Decimal64BID(bidgo.Bid64Abs(d.ToUint64()))
}

func decimal64BIDNegatePort(d Decimal64BID) Decimal64BID {
	return Decimal64BID(bidgo.Bid64Negate(d.ToUint64()))
}

func decimal64BIDCopySignPort(d, signSource Decimal64BID) Decimal64BID {
	return Decimal64BID(bidgo.Bid64CopySign(d.ToUint64(), signSource.ToUint64()))
}

func decimal64BIDStringPort(d Decimal64BID) string {
	if s, ok := formatDecimal64BIDNaN(d.ToUint64()); ok {
		return s
	}
	return bidgo.Bid64ToString(d.ToUint64())
}

func decimal64BIDToBinary32Port(d Decimal64BID, mode RoundingMode) (float32, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		bits, _ := bidgo.Bid64ToBinary32(bidNaNBits64, bidgoRoundingNearestEven)
		return math.Float32frombits(bits), FlagInvalidOperation
	}
	result, flags := bidgo.Bid64ToBinary32(d.ToUint64(), rnd)
	return math.Float32frombits(result), bidgoExceptionFlags(flags)
}

func decimal64BIDToBinary64Port(d Decimal64BID, mode RoundingMode) (float64, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		bits, _ := bidgo.Bid64ToBinary64(bidNaNBits64, bidgoRoundingNearestEven)
		return math.Float64frombits(bits), FlagInvalidOperation
	}
	result, flags := bidgo.Bid64ToBinary64(d.ToUint64(), rnd)
	return math.Float64frombits(result), bidgoExceptionFlags(flags)
}

func decimal64BIDToBinary128Port(d Decimal64BID, mode RoundingMode) (Binary128, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		bits, _ := bidgo.Bid64ToBinary128(bidNaNBits64, bidgoRoundingNearestEven)
		return binary128FromBidgo(bits), FlagInvalidOperation
	}
	result, flags := bidgo.Bid64ToBinary128(d.ToUint64(), rnd)
	return binary128FromBidgo(result), bidgoExceptionFlags(flags)
}

// See decimal32BIDToDecimal128Port for why this conversion writes its result
// through decimal128BIDSetBidgo.
func decimal64BIDToDecimal128Port(d Decimal64BID) (out Decimal128BID, flags ExceptionFlags) {
	result, rawFlags := bidgo.Bid64ToBid128(d.ToUint64())
	decimal128BIDSetBidgo(&out, result)
	return out, bidgoExceptionFlags(rawFlags)
}

func decimal64BIDToDecimal32Port(d Decimal64BID, mode RoundingMode) (Decimal32BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN32BID(), FlagInvalidOperation
	}
	result, flags := bidgo.Bid64ToBid32(d.ToUint64(), rnd)
	return Decimal32BID(result), bidgoExceptionFlags(flags)
}

func decimal64BIDNextTowardPort(d Decimal64BID, target Decimal128BID) (Decimal64BID, ExceptionFlags) {
	result, flags := bidgo.Bid64NextToward(d.ToUint64(), decimal128BIDAsBidgo(target))
	return Decimal64BID(result), bidgoExceptionFlags(flags)
}

func decimal64BIDNextPlusPort(d Decimal64BID) (Decimal64BID, ExceptionFlags) {
	result, flags := bidgo.Bid64NextUp(d.ToUint64())
	return Decimal64BID(result), bidgoExceptionFlags(flags)
}

func decimal64BIDNextMinusPort(d Decimal64BID) (Decimal64BID, ExceptionFlags) {
	result, flags := bidgo.Bid64NextDown(d.ToUint64())
	return Decimal64BID(result), bidgoExceptionFlags(flags)
}

func decimal64BIDPrettyStringPort(d Decimal64BID) string {
	return formatPrettyBIDString(decimal64BIDStringPort(d), decimal64Precision)
}

func decimal64BIDIsZeroPort(d Decimal64BID) bool {
	return bidgo.Bid64IsZero(d.ToUint64()) != 0
}

func decimal64BIDIsNaNPort(d Decimal64BID) bool {
	return bidgo.Bid64IsNaN(d.ToUint64()) != 0
}

func decimal64BIDIsInfPort(d Decimal64BID) bool {
	return bidgo.Bid64IsInf(d.ToUint64()) != 0
}

func decimal64BIDIsNormalPort(d Decimal64BID) bool {
	return bidgo.Bid64IsNormal(d.ToUint64()) != 0
}

func decimal64BIDIsFinitePort(d Decimal64BID) bool {
	return bidgo.Bid64IsFinite(d.ToUint64()) != 0
}

func decimal64BIDIsSubnormalPort(d Decimal64BID) bool {
	return bidgo.Bid64IsSubnormal(d.ToUint64()) != 0
}

func decimal64BIDIsSignalingPort(d Decimal64BID) bool {
	return bidgo.Bid64IsSignaling(d.ToUint64()) != 0
}

func decimal64BIDIsCanonicalPort(d Decimal64BID) bool {
	return bidgo.Bid64IsCanonical(d.ToUint64()) != 0
}

func decimal64BIDRadixPort() int {
	return bidgo.Bid64Radix()
}

func decimal64BIDIsSignMinusPort(d Decimal64BID) bool {
	return bidgo.Bid64IsSigned(d.ToUint64()) != 0
}

func decimal64BIDClassPort(d Decimal64BID) DecimalClass {
	return decimalClassFromBIDClass(bidgo.Bid64Class(d.ToUint64()))
}

func decimal64BIDSignPort(d Decimal64BID) int {
	if decimal64BIDIsZeroPort(d) {
		return 0
	}
	if bidgo.Bid64IsSigned(d.ToUint64()) != 0 {
		return -1
	}
	return 1
}

func newDecimal64BIDDirectPort(s string) (Decimal64BID, error) {
	result, flags := parseDecimal64BIDPort(s)
	if rejectedBIDStringInput(flags) || unrepresentableBIDStringFlags(flags) {
		return 0, fmt.Errorf("invalid decimal string: %s", s)
	}
	return result, nil
}

func newDecimal64BIDWithFlagsPort(s string) (Decimal64BID, ExceptionFlags, error) {
	result, flags := parseDecimal64BIDPort(s)
	if rejectedBIDStringInput(flags) {
		return 0, 0, fmt.Errorf("invalid decimal string: %s", s)
	}
	return result, flags, nil
}

// newDecimal64BIDWithModePort mirrors newDecimal64BIDWithFlagsPort with the
// caller's rounding mode carried into the parse. WithFlags input rejection is
// checked before an invalid-mode result so a bad mode cannot mask a rejected
// string.
func newDecimal64BIDWithModePort(s string, mode RoundingMode) (Decimal64BID, ExceptionFlags, error) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		if _, _, err := newDecimal64BIDWithFlagsPort(s); err != nil {
			return 0, 0, err
		}
		return canonicalQNaN64BID(), FlagInvalidOperation, nil
	}
	result, flags := parseDecimal64BIDPublicMode(s, rnd)
	if rejectedBIDStringInput(flags) {
		return 0, 0, fmt.Errorf("invalid decimal string: %s", s)
	}
	return result, flags, nil
}

func parseDecimal64BIDPort(s string) (Decimal64BID, ExceptionFlags) {
	return parseDecimal64BIDPublicMode(s, defaultBIDRoundingMode)
}

// parseDecimal64BIDPublicMode is the Decimal64 counterpart of
// parseDecimal32BIDPublicMode, sharing its single-scan structure and
// equivalence argument.
func parseDecimal64BIDPublicMode(s string, rndMode int) (Decimal64BID, ExceptionFlags) {
	result, flags, rawStatus := parseDecimal64BIDPortModeWithRawStatus(s, rndMode)
	if bidgo.Bid64IsNaN(result.ToUint64()) != 0 {
		if invalidBIDStringInput(s, true) {
			return canonicalQNaN64BID(), FlagInvalidOperation
		}
		return result, flags
	}
	literal, ok := parseBIDFiniteLiteral(s)
	if !ok {
		return canonicalQNaN64BID(), FlagInvalidOperation
	}
	if rawStatus == bidgo.BID_EXACT_STATUS && literal.cohortUnrepresentable(decimal64MinQuantum, decimal64MaxQuantum, decimal64Precision) {
		return canonicalQNaN64BID(), FlagInvalidOperation
	}
	return result, flags
}

func parseDecimal64BIDPortMode(s string, rndMode int) (Decimal64BID, ExceptionFlags) {
	result, flags, _ := parseDecimal64BIDPortModeWithRawStatus(s, rndMode)
	return result, flags
}

func parseDecimal64BIDPortModeWithRawStatus(s string, rndMode int) (Decimal64BID, ExceptionFlags, uint32) {
	if result, ok := parseDecimal64BIDNaN(s); ok {
		return result, 0, bidgo.BID_EXACT_STATUS
	}
	if _, ok := parseBIDNaNLiteral(s); ok {
		return canonicalQNaN64BID(), FlagInvalidOperation, bidgo.BID_INVALID_EXCEPTION
	}
	result, rawFlags := bidgo.Bid64FromString(s, rndMode)
	return Decimal64BID(result), bidgoExceptionFlags(rawFlags), rawFlags
}

const (
	decimal32MinQuantum  = -101
	decimal32MaxQuantum  = 90
	decimal32Precision   = 7
	decimal64MinQuantum  = -398
	decimal64MaxQuantum  = 369
	decimal64Precision   = 16
	decimal128MinQuantum = -6176
	decimal128MaxQuantum = 6111
	decimal128Precision  = 34
)

// rejectedBIDStringInput reports whether a public error-returning string
// parser must reject a parseDecimal*BIDPublicMode result. That parse already
// reports every public-contract rejection through FlagInvalidOperation:
// malformed syntax and a NaN result whose input does not spell a NaN literal
// become canonical quiet NaN + FlagInvalidOperation, a silently coerced
// cohort does the same, and an oversized NaN payload carries the raw parser's
// FlagInvalidOperation. Conversely an accepted public-mode result never
// carries FlagInvalidOperation, so the error-channel decision needs only the
// returned flags and no second scan of the input.
func rejectedBIDStringInput(flags ExceptionFlags) bool {
	return flags&FlagInvalidOperation != 0
}

// unrepresentableBIDStringFlags reports whether converting a syntactically
// valid finite literal changed its numeric value or range. Error-only string
// constructors cannot expose these flags, so they must fail instead of
// silently returning the rounded, overflowed, or underflowed result.
func unrepresentableBIDStringFlags(flags ExceptionFlags) bool {
	const mask = FlagInexact | FlagUnderflow | FlagOverflow
	return flags&mask != 0
}

// invalidBIDStringInput reports whether a public string-parse input violates
// the complete literal grammar. The mechanical-port parser never fails: it
// usually maps malformed input to a NaN result, maps a radix point with no
// significand digits (for example "." or ".e1") to a finite signed zero, and
// Decimal128 accepts a valid exponent prefix followed by junk. The public
// literal contract rejects all of those forms. A NaN result is only
// legitimate when the input itself spells a NaN literal:
//
//	[+|-] (nan | qnan | snan) [decimal digits]   (case-insensitive)
//
// which is exactly the grammar parseBIDNaNLiteral recognizes. Per-width NaN
// payload range is enforced by rejectedBIDStringInput through the raw
// parser's FlagInvalidOperation result. The Intel port has no parenthesized
// payload form, so inputs like "nan(123)" are rejected.
//
// The public-mode parsers now call this only with resultIsNaN=true: their
// finite arm folds the equivalent syntax check into the single
// parseBIDFiniteLiteral scan that also feeds the silent-cohort decision. The
// finite arm here still states the full contract for that reduction and for
// validBIDFiniteLiteral's remaining direct callers.
func invalidBIDStringInput(input string, resultIsNaN bool) bool {
	if strings.TrimSpace(input) == "" {
		return true
	}
	if !resultIsNaN {
		return !validBIDFiniteLiteral(input)
	}
	_, ok := parseBIDNaNLiteral(input)
	return !ok
}

// validBIDFiniteLiteral checks the complete finite-result portion of the
// public literal contract. parseBIDFiniteLiteral performs the actual parse so
// syntax validation and silent-cohort detection cannot drift apart.
func validBIDFiniteLiteral(input string) bool {
	_, ok := parseBIDFiniteLiteral(input)
	return ok
}

type bidFiniteLiteral struct {
	// infinite marks an exact infinity spelling, which has no written cohort.
	infinite bool
	// quantum is the written cohort quantum (explicit exponent minus the
	// fractional digit count) when quantumOutsideInt64 is false.
	quantum int64
	// quantumOutsideInt64 reports that the exact quantum does not fit int64.
	// Every BID width's quantum range is a tiny sub-range of int64, so the
	// cohort decision needs only this fact, never the wide value itself, and
	// the overflow direction is irrelevant: either sign is unrepresentable.
	quantumOutsideInt64 bool
	coefficientDigits   int
}

// parseBIDFiniteLiteral parses the complete finite literal grammar and returns
// its written cohort: quantum (explicit exponent minus fractional digit count)
// and coefficient digit count after leading zeros. An exact infinity spelling
// is valid and reports infinite instead of a quantum. Only leading ASCII
// space/tab is accepted, matching the Intel port and public NaN grammar;
// trailing or internal whitespace is rejected.
//
// The quantum is computed in exact uint64/int64 arithmetic with no big-integer
// allocation, and the result is decision-equivalent to arbitrary-precision
// arithmetic for every input, including a huge exponent canceling a huge
// fractional digit count:
//
//   - fractionalDigits F counts input bytes, so 0 <= F < 2^63.
//   - If the exponent digit string overflows uint64 (true magnitude >= 2^64),
//     the exact quantum is -(M+F) <= -2^64 for a negative exponent and
//     M-F > 2^64-2^63 = 2^63 for a positive one; both lie outside int64, which
//     is exactly what arbitrary precision would report.
//   - Otherwise the magnitude M fits uint64 and bidLiteralQuantum computes
//     +-M-F exactly, again reporting outside-int64 only when the exact value
//     is outside int64.
func parseBIDFiniteLiteral(input string) (bidFiniteLiteral, bool) {
	rest := strings.TrimLeft(input, " \t")
	if rest == "" {
		return bidFiniteLiteral{}, false
	}
	if rest[0] == '+' || rest[0] == '-' {
		rest = rest[1:]
	}
	if strings.EqualFold(rest, "inf") || strings.EqualFold(rest, "infinity") {
		return bidFiniteLiteral{infinite: true}, true
	}

	seenDigit := false
	seenPoint := false
	fractionalDigits := 0
	coefficientDigits := 0
	i := 0
	for ; i < len(rest); i++ {
		switch c := rest[i]; {
		case c >= '0' && c <= '9':
			seenDigit = true
			if coefficientDigits > 0 || c != '0' {
				coefficientDigits++
			}
			if seenPoint {
				fractionalDigits++
			}
		case c == '.' && !seenPoint:
			seenPoint = true
		default:
			goto exponent
		}
	}
	if !seenDigit {
		return bidFiniteLiteral{}, false
	}
	return bidFiniteLiteral{
		quantum:           -int64(fractionalDigits),
		coefficientDigits: coefficientDigits,
	}, true

exponent:
	if !seenDigit || (rest[i] != 'e' && rest[i] != 'E') {
		return bidFiniteLiteral{}, false
	}
	i++
	exponentNegative := false
	if i < len(rest) && (rest[i] == '+' || rest[i] == '-') {
		exponentNegative = rest[i] == '-'
		i++
	}
	exponentStart := i
	exponentMagnitude := uint64(0)
	magnitudeOverflows := false
	for ; i < len(rest) && rest[i] >= '0' && rest[i] <= '9'; i++ {
		d := uint64(rest[i] - '0')
		if exponentMagnitude > (math.MaxUint64-d)/10 {
			// The true magnitude is already >= 2^64 here, and appending
			// further digits only grows it; keep scanning for syntax only.
			magnitudeOverflows = true
			continue
		}
		exponentMagnitude = exponentMagnitude*10 + d
	}
	if i != len(rest) || i == exponentStart {
		return bidFiniteLiteral{}, false
	}
	literal := bidFiniteLiteral{coefficientDigits: coefficientDigits}
	literal.quantum, literal.quantumOutsideInt64 = bidLiteralQuantum(exponentMagnitude, magnitudeOverflows, exponentNegative, uint64(fractionalDigits))
	return literal, true
}

// bidLiteralQuantum computes quantum = +-magnitude - fractionalDigits exactly,
// reporting outside-int64 instead of a value when the exact result does not
// fit. magnitudeOverflows means the true magnitude is >= 2^64 while
// fractionalDigits < 2^63 (it counts input bytes), so the exact quantum is
// then outside int64 for either exponent sign.
func bidLiteralQuantum(magnitude uint64, magnitudeOverflows, negative bool, fractionalDigits uint64) (int64, bool) {
	if magnitudeOverflows {
		return 0, true
	}
	if negative {
		// quantum = -(magnitude + fractionalDigits); representable iff the
		// exact sum is at most 2^63 (int64's most negative value is -2^63).
		sum := magnitude + fractionalDigits
		if sum < magnitude || sum > 1<<63 {
			return 0, true
		}
		if sum == 1<<63 {
			return math.MinInt64, false
		}
		return -int64(sum), false
	}
	if magnitude >= fractionalDigits {
		diff := magnitude - fractionalDigits
		if diff > math.MaxInt64 {
			return 0, true
		}
		return int64(diff), false
	}
	// fractionalDigits < 2^63, so the negative difference always fits.
	return -int64(fractionalDigits - magnitude), false
}

// cohortUnrepresentable reports whether a complete numeric finite literal
// asks for a quantum or coefficient the target BID width cannot encode. The
// Intel from-string port can silently rescale a trailing-zero coefficient or
// clamp an out-of-range zero while keeping the numeric value exact and
// therefore raising no status. Public APIs must detect that otherwise-silent
// cohort coercion explicitly.
func (l bidFiniteLiteral) cohortUnrepresentable(minQuantum, maxQuantum int64, precision int) bool {
	if l.infinite {
		return false
	}
	if l.coefficientDigits > precision || l.quantumOutsideInt64 {
		return true
	}
	return l.quantum < minQuantum || l.quantum > maxQuantum
}

func trimTrailingIntegerSuffix(s string) string {
	if strings.HasSuffix(s, ".0") {
		return s[:len(s)-2]
	}
	return s
}

// formatPrettyBIDString keeps the display-oriented plain form only when that
// spelling remains a representable cohort for the source width. Expanding a
// positive exponent can append more coefficient digits than the width holds;
// for example Decimal32 1960256E+14 used to become the 21-digit plain literal
// 196025600000000000000, which the exact public parser must reject. Selecting
// the port's exponent form in that case keeps the same value and makes every
// finite PrettyString result consumable by the matching public constructor.
func formatPrettyBIDString(raw string, precision int) string {
	formatted := trimTrailingIntegerSuffix(optimizedFormatDecimalString(raw))
	// A spelling no longer than the precision cannot contain too many
	// coefficient digits, so the common short-value path needs no second scan.
	if len(formatted) <= precision {
		return formatted
	}
	if coefficientDigits, plainFinite := plainBIDCoefficientDigits(formatted); plainFinite && coefficientDigits > precision {
		return strings.TrimPrefix(raw, "+")
	}
	return formatted
}

// plainBIDCoefficientDigits returns the significant written coefficient digit
// count for a plain finite literal. Leading zeros do not consume precision;
// once the first non-zero digit appears, trailing zeros do. Scientific and
// special spellings return false because formatPrettyBIDString only needs to
// police the new plain cohort created by exponent expansion. Moving a decimal
// point to plain notation preserves the source quantum or raises it by
// trimming fractional zeros, while an integer expansion has quantum zero, so
// a valid BID source remains inside every width's quantum range. The only new
// unrepresentable field can therefore be the written coefficient precision.
func plainBIDCoefficientDigits(s string) (int, bool) {
	if s == "" {
		return 0, false
	}
	i := 0
	if s[0] == '+' || s[0] == '-' {
		i++
	}
	seenDigit := false
	seenPoint := false
	digits := 0
	for ; i < len(s); i++ {
		switch c := s[i]; {
		case c >= '0' && c <= '9':
			seenDigit = true
			if digits > 0 || c != '0' {
				digits++
			}
		case c == '.' && !seenPoint:
			seenPoint = true
		default:
			return 0, false
		}
	}
	return digits, seenDigit
}

func decimal128BIDStringPort(d Decimal128BID) string {
	if s, ok := formatDecimal128BIDNaN(d); ok {
		return s
	}
	return bidgo.Bid128ToString(decimal128BIDAsBidgo(d))
}

func decimal128BIDToBinary32Port(d Decimal128BID, mode RoundingMode) (float32, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		var discard uint32
		nan := bidgo.Bid128ToBinary32(bidNaN128Bidgo(), bidgoRoundingNearestEven, &discard)
		return nan, FlagInvalidOperation
	}
	var flags uint32
	result := bidgo.Bid128ToBinary32(decimal128BIDAsBidgo(d), rnd, &flags)
	return result, bidgoExceptionFlags(flags)
}

func decimal128BIDToBinary64Port(d Decimal128BID, mode RoundingMode) (float64, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		var discard uint32
		nan := bidgo.Bid128ToBinary64(bidNaN128Bidgo(), bidgoRoundingNearestEven, &discard)
		return nan, FlagInvalidOperation
	}
	var flags uint32
	result := bidgo.Bid128ToBinary64(decimal128BIDAsBidgo(d), rnd, &flags)
	return result, bidgoExceptionFlags(flags)
}

func decimal128BIDToBinary128Port(d Decimal128BID, mode RoundingMode) (Binary128, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		bits, _ := bidgo.Bid128ToBinary128(bidNaN128Bidgo(), bidgoRoundingNearestEven)
		return binary128FromBidgo(bits), FlagInvalidOperation
	}
	result, flags := bidgo.Bid128ToBinary128(decimal128BIDAsBidgo(d), rnd)
	return binary128FromBidgo(result), bidgoExceptionFlags(flags)
}

func decimal128BIDNextTowardPort(d Decimal128BID, target Decimal128BID) (Decimal128BID, ExceptionFlags) {
	result, flags := bidgo.Bid128NextToward(decimal128BIDAsBidgo(d), decimal128BIDAsBidgo(target))
	return decimal128BIDFromBidgo(result), bidgoExceptionFlags(flags)
}

func decimal128BIDNextPlusPort(d Decimal128BID) (Decimal128BID, ExceptionFlags) {
	result, flags := bidgo.Bid128NextUp(decimal128BIDAsBidgo(d))
	return decimal128BIDFromBidgo(result), bidgoExceptionFlags(flags)
}

func decimal128BIDNextMinusPort(d Decimal128BID) (Decimal128BID, ExceptionFlags) {
	result, flags := bidgo.Bid128NextDown(decimal128BIDAsBidgo(d))
	return decimal128BIDFromBidgo(result), bidgoExceptionFlags(flags)
}

func decimal128BIDAddPort(d, other Decimal128BID) Decimal128BID {
	return decimal128BIDAddPortMode(d, other, defaultBIDRoundingMode)
}

func decimal128BIDAddPortFlags(d, other Decimal128BID) (Decimal128BID, ExceptionFlags) {
	return decimal128BIDAddPortModeFlags(d, other, defaultBIDRoundingMode)
}

func decimal128BIDPrettyStringPort(d Decimal128BID) string {
	return formatPrettyBIDString(decimal128BIDStringPort(d), decimal128Precision)
}

func decimal128BIDAddPortMode(d, other Decimal128BID, rndMode int) Decimal128BID {
	var flags uint32
	return decimal128BIDFromBidgo(bidgo.Bid128Add(decimal128BIDAsBidgo(d), decimal128BIDAsBidgo(other), rndMode, &flags))
}

func decimal128BIDAddPortModeFlags(d, other Decimal128BID, rndMode int) (Decimal128BID, ExceptionFlags) {
	var flags uint32
	result := bidgo.Bid128Add(decimal128BIDAsBidgo(d), decimal128BIDAsBidgo(other), rndMode, &flags)
	return decimal128BIDFromBidgo(result), bidgoExceptionFlags(flags)
}

func decimal128BIDSubPort(d, other Decimal128BID) Decimal128BID {
	var flags uint32
	return decimal128BIDFromBidgo(bidgo.Bid128Sub(decimal128BIDAsBidgo(d), decimal128BIDAsBidgo(other), defaultBIDRoundingMode, &flags))
}

func decimal128BIDSubPortFlags(d, other Decimal128BID) (Decimal128BID, ExceptionFlags) {
	return decimal128BIDSubPortModeFlags(d, other, defaultBIDRoundingMode)
}

func decimal128BIDSubPortModeFlags(d, other Decimal128BID, rndMode int) (Decimal128BID, ExceptionFlags) {
	var flags uint32
	result := bidgo.Bid128Sub(decimal128BIDAsBidgo(d), decimal128BIDAsBidgo(other), rndMode, &flags)
	return decimal128BIDFromBidgo(result), bidgoExceptionFlags(flags)
}

func decimal128BIDMulPort(d, other Decimal128BID) Decimal128BID {
	result, _ := bidgo.Bid128Mul(decimal128BIDAsBidgo(d), decimal128BIDAsBidgo(other), defaultBIDRoundingMode)
	return decimal128BIDFromBidgo(result)
}

func decimal128BIDMulPortFlags(d, other Decimal128BID) (Decimal128BID, ExceptionFlags) {
	result, flags := bidgo.Bid128Mul(decimal128BIDAsBidgo(d), decimal128BIDAsBidgo(other), defaultBIDRoundingMode)
	return decimal128BIDFromBidgo(result), bidgoExceptionFlags(flags)
}

func decimal128BIDMulPortModeFlags(d, other Decimal128BID, rndMode int) (Decimal128BID, ExceptionFlags) {
	result, flags := bidgo.Bid128Mul(decimal128BIDAsBidgo(d), decimal128BIDAsBidgo(other), rndMode)
	return decimal128BIDFromBidgo(result), bidgoExceptionFlags(flags)
}

func decimal128BIDFMAPort(d, mul, add Decimal128BID) (Decimal128BID, ExceptionFlags) {
	return decimal128BIDFMAPortMode(d, mul, add, defaultBIDRoundingMode)
}

func decimal128BIDFMAPortMode(d, mul, add Decimal128BID, rndMode int) (Decimal128BID, ExceptionFlags) {
	result, flags := bidgo.Bid128Fma(decimal128BIDAsBidgo(d), decimal128BIDAsBidgo(mul), decimal128BIDAsBidgo(add), rndMode)
	return decimal128BIDFromBidgo(result), bidgoExceptionFlags(flags)
}

func decimal128BIDSqrtPort(d Decimal128BID) (Decimal128BID, ExceptionFlags) {
	result, flags := bidgo.Bid128Sqrt(decimal128BIDAsBidgo(d), defaultBIDRoundingMode)
	return decimal128BIDFromBidgo(result), bidgoExceptionFlags(flags)
}

func decimal128BIDSqrtPortModeFlags(d Decimal128BID, rndMode int) (Decimal128BID, ExceptionFlags) {
	result, flags := bidgo.Bid128Sqrt(decimal128BIDAsBidgo(d), rndMode)
	return decimal128BIDFromBidgo(result), bidgoExceptionFlags(flags)
}

func decimal128BIDDivPort(d, other Decimal128BID) Decimal128BID {
	result, _ := bidgo.Bid128Div(decimal128BIDAsBidgo(d), decimal128BIDAsBidgo(other), defaultBIDRoundingMode)
	return decimal128BIDFromBidgo(result)
}

func decimal128BIDDivPortFlags(d, other Decimal128BID) (Decimal128BID, ExceptionFlags) {
	result, flags := bidgo.Bid128Div(decimal128BIDAsBidgo(d), decimal128BIDAsBidgo(other), defaultBIDRoundingMode)
	return decimal128BIDFromBidgo(result), bidgoExceptionFlags(flags)
}

func decimal128BIDDivPortModeFlags(d, other Decimal128BID, rndMode int) (Decimal128BID, ExceptionFlags) {
	result, flags := bidgo.Bid128Div(decimal128BIDAsBidgo(d), decimal128BIDAsBidgo(other), rndMode)
	return decimal128BIDFromBidgo(result), bidgoExceptionFlags(flags)
}

func decimal128BIDRemainderPort(d, other Decimal128BID) (Decimal128BID, ExceptionFlags) {
	result, flags := bidgo.Bid128Rem(decimal128BIDAsBidgo(d), decimal128BIDAsBidgo(other))
	return decimal128BIDFromBidgo(result), bidgoExceptionFlags(flags)
}

func decimal128BIDFmodPort(d, other Decimal128BID) (Decimal128BID, ExceptionFlags) {
	result, flags := bidgo.Bid128Fmod(decimal128BIDAsBidgo(d), decimal128BIDAsBidgo(other))
	return decimal128BIDFromBidgo(result), bidgoExceptionFlags(flags)
}

func decimal128BIDQuantizePort(d, other Decimal128BID) Decimal128BID {
	result, _ := bidgo.Bid128Quantize(decimal128BIDAsBidgo(d), decimal128BIDAsBidgo(other), defaultBIDRoundingMode)
	return decimal128BIDFromBidgo(result)
}

func decimal128BIDQuantizePortFlags(d, other Decimal128BID) (Decimal128BID, ExceptionFlags) {
	result, flags := bidgo.Bid128Quantize(decimal128BIDAsBidgo(d), decimal128BIDAsBidgo(other), defaultBIDRoundingMode)
	return decimal128BIDFromBidgo(result), bidgoExceptionFlags(flags)
}

func decimal128BIDQuantizePortModeFlags(d, other Decimal128BID, rndMode int) (Decimal128BID, ExceptionFlags) {
	result, flags := bidgo.Bid128Quantize(decimal128BIDAsBidgo(d), decimal128BIDAsBidgo(other), rndMode)
	return decimal128BIDFromBidgo(result), bidgoExceptionFlags(flags)
}

func decimal128BIDSameQuantumPort(d, other Decimal128BID) bool {
	return bidgo.Bid128SameQuantum(decimal128BIDAsBidgo(d), decimal128BIDAsBidgo(other)) != 0
}

func decimal128BIDMinNumPort(d, other Decimal128BID) (Decimal128BID, ExceptionFlags) {
	var flags uint32
	result := bidgo.Bid128Minnum(decimal128BIDAsBidgo(d), decimal128BIDAsBidgo(other), &flags)
	return decimal128BIDFromBidgo(result), bidgoExceptionFlags(flags)
}

func decimal128BIDMaxNumPort(d, other Decimal128BID) (Decimal128BID, ExceptionFlags) {
	var flags uint32
	result := bidgo.Bid128Maxnum(decimal128BIDAsBidgo(d), decimal128BIDAsBidgo(other), &flags)
	return decimal128BIDFromBidgo(result), bidgoExceptionFlags(flags)
}

func decimal128BIDMinNumMagPort(d, other Decimal128BID) (Decimal128BID, ExceptionFlags) {
	var flags uint32
	result := bidgo.Bid128MinnumMag(decimal128BIDAsBidgo(d), decimal128BIDAsBidgo(other), &flags)
	return decimal128BIDFromBidgo(result), bidgoExceptionFlags(flags)
}

func decimal128BIDMaxNumMagPort(d, other Decimal128BID) (Decimal128BID, ExceptionFlags) {
	var flags uint32
	result := bidgo.Bid128MaxnumMag(decimal128BIDAsBidgo(d), decimal128BIDAsBidgo(other), &flags)
	return decimal128BIDFromBidgo(result), bidgoExceptionFlags(flags)
}

func decimal128BIDCompareTotalPort(d, other Decimal128BID) int {
	return totalOrderComparison(bidgo.Bid128TotalOrder(decimal128BIDAsBidgo(d), decimal128BIDAsBidgo(other)), bidgo.Bid128TotalOrder(decimal128BIDAsBidgo(other), decimal128BIDAsBidgo(d)))
}

func decimal128BIDCompareTotalMagPort(d, other Decimal128BID) int {
	return totalOrderComparison(bidgo.Bid128TotalOrderMag(decimal128BIDAsBidgo(d), decimal128BIDAsBidgo(other)), bidgo.Bid128TotalOrderMag(decimal128BIDAsBidgo(other), decimal128BIDAsBidgo(d)))
}

func totalOrderComparison(leftLE, rightLE int) int {
	switch {
	case leftLE != 0 && rightLE != 0:
		return 0
	case leftLE != 0:
		return -1
	default:
		return 1
	}
}

func decimal128BIDRoundIntegralExactPort(d Decimal128BID) Decimal128BID {
	var flags uint32
	return decimal128BIDFromBidgo(bidgo.Bid128RoundIntegralExact(decimal128BIDAsBidgo(d), defaultBIDRoundingMode, &flags))
}

func decimal128BIDRoundIntegralExactPortFlags(d Decimal128BID) (Decimal128BID, ExceptionFlags) {
	var flags uint32
	result := bidgo.Bid128RoundIntegralExact(decimal128BIDAsBidgo(d), defaultBIDRoundingMode, &flags)
	return decimal128BIDFromBidgo(result), bidgoExceptionFlags(flags)
}

func decimal128BIDRoundIntegralExactPortModeFlags(d Decimal128BID, rndMode int) (Decimal128BID, ExceptionFlags) {
	var flags uint32
	result := bidgo.Bid128RoundIntegralExact(decimal128BIDAsBidgo(d), rndMode, &flags)
	return decimal128BIDFromBidgo(result), bidgoExceptionFlags(flags)
}

func decimal128BIDRoundIntegralNearestEvenPort(d Decimal128BID) (Decimal128BID, ExceptionFlags) {
	var flags uint32
	result := bidgo.Bid128RoundIntegralNearestEven(decimal128BIDAsBidgo(d), &flags)
	return decimal128BIDFromBidgo(result), bidgoExceptionFlags(flags)
}

func decimal128BIDRoundIntegralNearestAwayPort(d Decimal128BID) (Decimal128BID, ExceptionFlags) {
	var flags uint32
	result := bidgo.Bid128RoundIntegralNearestAway(decimal128BIDAsBidgo(d), &flags)
	return decimal128BIDFromBidgo(result), bidgoExceptionFlags(flags)
}

func decimal128BIDRoundIntegralZeroPort(d Decimal128BID) (Decimal128BID, ExceptionFlags) {
	var flags uint32
	result := bidgo.Bid128RoundIntegralZero(decimal128BIDAsBidgo(d), &flags)
	return decimal128BIDFromBidgo(result), bidgoExceptionFlags(flags)
}

func decimal128BIDRoundIntegralPositivePort(d Decimal128BID) (Decimal128BID, ExceptionFlags) {
	var flags uint32
	result := bidgo.Bid128RoundIntegralPositive(decimal128BIDAsBidgo(d), &flags)
	return decimal128BIDFromBidgo(result), bidgoExceptionFlags(flags)
}

func decimal128BIDRoundIntegralNegativePort(d Decimal128BID) (Decimal128BID, ExceptionFlags) {
	var flags uint32
	result := bidgo.Bid128RoundIntegralNegative(decimal128BIDAsBidgo(d), &flags)
	return decimal128BIDFromBidgo(result), bidgoExceptionFlags(flags)
}

func decimal128BIDLogBPort(d Decimal128BID) (Decimal128BID, ExceptionFlags) {
	var flags uint32
	result := bidgo.Bid128Logb(decimal128BIDAsBidgo(d), &flags)
	return decimal128BIDFromBidgo(result), bidgoExceptionFlags(flags)
}

func decimal128BIDScaleBPort(d Decimal128BID, exponent int) (Decimal128BID, ExceptionFlags) {
	var flags uint32
	result := bidgo.Bid128Scalbln(decimal128BIDAsBidgo(d), int64(exponent), defaultBIDRoundingMode, &flags)
	return decimal128BIDFromBidgo(result), bidgoExceptionFlags(flags)
}

func decimal128BIDScaleBPortModeFlags(d Decimal128BID, exponent int, rndMode int) (Decimal128BID, ExceptionFlags) {
	var flags uint32
	result := bidgo.Bid128Scalbln(decimal128BIDAsBidgo(d), int64(exponent), rndMode, &flags)
	return decimal128BIDFromBidgo(result), bidgoExceptionFlags(flags)
}

func decimal128BIDCopyPort(d Decimal128BID) Decimal128BID {
	return decimal128BIDFromBidgo(bidgo.Bid128Copy(decimal128BIDAsBidgo(d)))
}

func decimal128BIDAbsPort(d Decimal128BID) Decimal128BID {
	return decimal128BIDFromBidgo(bidgo.Bid128Abs(decimal128BIDAsBidgo(d)))
}

func decimal128BIDNegatePort(d Decimal128BID) Decimal128BID {
	return decimal128BIDFromBidgo(bidgo.Bid128Negate(decimal128BIDAsBidgo(d)))
}

func decimal128BIDCopySignPort(d, signSource Decimal128BID) Decimal128BID {
	return decimal128BIDFromBidgo(bidgo.Bid128CopySign(decimal128BIDAsBidgo(d), decimal128BIDAsBidgo(signSource)))
}

func decimal128BIDIsZeroPort(d Decimal128BID) bool {
	return bidgo.Bid128IsZero(decimal128BIDAsBidgo(d)) != 0
}

func decimal128BIDIsNaNPort(d Decimal128BID) bool {
	return bidgo.Bid128IsNaN(decimal128BIDAsBidgo(d)) != 0
}

func decimal128BIDIsInfPort(d Decimal128BID) bool {
	return bidgo.Bid128IsInf(decimal128BIDAsBidgo(d)) != 0
}

func decimal128BIDIsNormalPort(d Decimal128BID) bool {
	return bidgo.Bid128IsNormal(decimal128BIDAsBidgo(d)) != 0
}

func decimal128BIDIsFinitePort(d Decimal128BID) bool {
	return bidgo.Bid128IsFinite(decimal128BIDAsBidgo(d)) != 0
}

func decimal128BIDIsSubnormalPort(d Decimal128BID) bool {
	return bidgo.Bid128IsSubnormal(decimal128BIDAsBidgo(d)) != 0
}

func decimal128BIDIsSignalingPort(d Decimal128BID) bool {
	return bidgo.Bid128IsSignaling(decimal128BIDAsBidgo(d)) != 0
}

func decimal128BIDIsCanonicalPort(d Decimal128BID) bool {
	return bidgo.Bid128IsCanonical(decimal128BIDAsBidgo(d)) != 0
}

func decimal128BIDRadixPort() int {
	return bidgo.Bid128Radix()
}

func decimal128BIDToDecimal64Port(d Decimal128BID, mode RoundingMode) (Decimal64BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN64BID(), FlagInvalidOperation
	}
	result, flags := bidgo.Bid128ToBid64(decimal128BIDAsBidgo(d), rnd)
	return Decimal64BID(result), bidgoExceptionFlags(flags)
}

func decimal128BIDToDecimal32Port(d Decimal128BID, mode RoundingMode) (Decimal32BID, ExceptionFlags) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		return canonicalQNaN32BID(), FlagInvalidOperation
	}
	result, flags := bidgo.Bid128ToBid32(decimal128BIDAsBidgo(d), rnd)
	return Decimal32BID(result), bidgoExceptionFlags(flags)
}

func decimal128BIDIsSignMinusPort(d Decimal128BID) bool {
	return bidgo.Bid128IsSigned(decimal128BIDAsBidgo(d)) != 0
}

func decimal128BIDClassPort(d Decimal128BID) DecimalClass {
	return decimalClassFromBIDClass(bidgo.Bid128Class(decimal128BIDAsBidgo(d)))
}

func decimal128BIDSignPort(d Decimal128BID) int {
	if decimal128BIDIsZeroPort(d) {
		return 0
	}
	if bidgo.Bid128IsSigned(decimal128BIDAsBidgo(d)) != 0 {
		return -1
	}
	return 1
}

func newDecimal128BIDDirectPort(s string) (Decimal128BID, error) {
	result, flags := parseDecimal128BIDPort(s)
	if rejectedBIDStringInput(flags) || unrepresentableBIDStringFlags(flags) {
		return Decimal128BID{}, fmt.Errorf("invalid decimal string: %s", s)
	}
	return result, nil
}

func newDecimal128BIDWithFlagsPort(s string) (Decimal128BID, ExceptionFlags, error) {
	result, flags := parseDecimal128BIDPort(s)
	if rejectedBIDStringInput(flags) {
		return Decimal128BID{}, 0, fmt.Errorf("invalid decimal string: %s", s)
	}
	return result, flags, nil
}

// newDecimal128BIDWithModePort mirrors newDecimal128BIDWithFlagsPort with the
// caller's rounding mode carried into the parse. WithFlags input rejection is
// checked before an invalid-mode result so a bad mode cannot mask a rejected
// string.
func newDecimal128BIDWithModePort(s string, mode RoundingMode) (Decimal128BID, ExceptionFlags, error) {
	rnd, ok := bidgoRoundingMode(mode)
	if !ok {
		if _, _, err := newDecimal128BIDWithFlagsPort(s); err != nil {
			return Decimal128BID{}, 0, err
		}
		return canonicalQNaN128BID(), FlagInvalidOperation, nil
	}
	result, flags := parseDecimal128BIDPublicMode(s, rnd)
	if rejectedBIDStringInput(flags) {
		return Decimal128BID{}, 0, fmt.Errorf("invalid decimal string: %s", s)
	}
	return result, flags, nil
}

func parseDecimal128BIDPort(s string) (Decimal128BID, ExceptionFlags) {
	return parseDecimal128BIDPublicMode(s, defaultBIDRoundingMode)
}

// parseDecimal128BIDPublicMode is the Decimal128 counterpart of
// parseDecimal32BIDPublicMode, sharing its single-scan structure and
// equivalence argument.
func parseDecimal128BIDPublicMode(s string, rndMode int) (Decimal128BID, ExceptionFlags) {
	result, flags, rawStatus := parseDecimal128BIDPortModeWithRawStatus(s, rndMode)
	if bidgo.Bid128IsNaN(decimal128BIDAsBidgo(result)) != 0 {
		if invalidBIDStringInput(s, true) {
			return canonicalQNaN128BID(), FlagInvalidOperation
		}
		return result, flags
	}
	literal, ok := parseBIDFiniteLiteral(s)
	if !ok {
		return canonicalQNaN128BID(), FlagInvalidOperation
	}
	if rawStatus == bidgo.BID_EXACT_STATUS && literal.cohortUnrepresentable(decimal128MinQuantum, decimal128MaxQuantum, decimal128Precision) {
		return canonicalQNaN128BID(), FlagInvalidOperation
	}
	return result, flags
}

func parseDecimal128BIDPortMode(s string, rndMode int) (Decimal128BID, ExceptionFlags) {
	result, flags, _ := parseDecimal128BIDPortModeWithRawStatus(s, rndMode)
	return result, flags
}

func parseDecimal128BIDPortModeWithRawStatus(s string, rndMode int) (Decimal128BID, ExceptionFlags, uint32) {
	if result, ok := parseDecimal128BIDNaN(s); ok {
		return result, 0, bidgo.BID_EXACT_STATUS
	}
	if _, ok := parseBIDNaNLiteral(s); ok {
		return canonicalQNaN128BID(), FlagInvalidOperation, bidgo.BID_INVALID_EXCEPTION
	}
	result, rawFlags := bidgo.Bid128FromString(s, rndMode)
	return decimal128BIDFromBidgo(result), bidgoExceptionFlags(rawFlags), rawFlags
}

// decimal128BIDAsBidgo decodes d's little-endian [16]byte image into the port
// (hi, lo) word representation. The Decimal128BID byte contract is
// little-endian on every platform, so the decode is explicit rather than a
// native-endian pointer reinterpretation (which byte-swapped the words on
// big-endian platforms). Both the hand-written routing layer and the generated
// FFI bit-compare runner call it, so a byte-order change here moves results in
// both.
func decimal128BIDAsBidgo(d Decimal128BID) bidgo.BID_UINT128 {
	return bidgo.Bid128FromWords(binary.LittleEndian.Uint64(d[8:16]), binary.LittleEndian.Uint64(d[0:8]))
}

// decimal128BIDFromBidgo encodes x's (hi, lo) words as the little-endian
// [16]byte image. The result is a named return so the two explicit stores write
// the result slot directly instead of filling a separate local that is then
// copied out.
//
// This is the value-form encode. decimal128BIDSetBidgo is the pointer form and
// binary128FromBidgo is the binary128-typed one; all three write the same two
// words in the same order, so a byte-order change to one must be made to all
// three. The generated public-API parity runner checks the encoded image
// against its own independent little-endian oracle, and the linux/s390x leg
// runs that runner, so a divergence between the three fails there.
func decimal128BIDFromBidgo(x bidgo.BID_UINT128) (d Decimal128BID) {
	hi, lo := bidgo.Bid128Words(x)
	binary.LittleEndian.PutUint64(d[0:8], lo)
	binary.LittleEndian.PutUint64(d[8:16], hi)
	return
}

// decimal128BIDSetBidgo writes x's (hi, lo) words as the little-endian
// [16]byte image of *d. It is the pointer form of decimal128BIDFromBidgo, for
// callers whose destination is already an addressable [16]byte they own, such
// as a named result. Assigning the value form to such a destination costs an
// extra store/reload round trip: the explicit byte-order stores make the
// array address-taken, so SSA cannot forward the intermediate away.
func decimal128BIDSetBidgo(d *Decimal128BID, x bidgo.BID_UINT128) {
	hi, lo := bidgo.Bid128Words(x)
	binary.LittleEndian.PutUint64(d[0:8], lo)
	binary.LittleEndian.PutUint64(d[8:16], hi)
}

// decimal128BIDWords returns the (hi, lo) 64-bit words stored in d's
// little-endian [16]byte image.
func decimal128BIDWords(d Decimal128BID) (hi, lo uint64) {
	return bidgo.Bid128Words(decimal128BIDAsBidgo(d))
}

// decimal128BIDFromWords builds the little-endian [16]byte image of the
// (hi, lo) 64-bit words.
func decimal128BIDFromWords(hi, lo uint64) Decimal128BID {
	return decimal128BIDFromBidgo(bidgo.Bid128FromWords(hi, lo))
}

// binary128FromBidgo encodes x's (hi, lo) words as the little-endian [16]byte
// binary128 image, with the same named-result reasoning as
// decimal128BIDFromBidgo.
func binary128FromBidgo(x bidgo.BID_UINT128) (b Binary128) {
	hi, lo := bidgo.Bid128Words(x)
	binary.LittleEndian.PutUint64(b[0:8], lo)
	binary.LittleEndian.PutUint64(b[8:16], hi)
	return
}
