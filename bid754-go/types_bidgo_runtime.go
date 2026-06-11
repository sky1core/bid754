package bid754

import (
	"fmt"
	"math"
	"strings"
	"unsafe"

	bidgo "github.com/sky1core/bid754/bid-go"
)

const defaultBIDRoundingMode = int(RoundNearestEven)

const (
	bidgoRoundingNearestEven    = 0
	bidgoRoundingTowardNegative = 1
	bidgoRoundingTowardPositive = 2
	bidgoRoundingTowardZero     = 3
	bidgoRoundingNearestAway    = 4
)

func bidgoRoundingMode(mode RoundingMode) int {
	switch mode {
	case RoundNearestEven:
		return bidgoRoundingNearestEven
	case RoundNearestAway:
		return bidgoRoundingNearestAway
	case RoundTowardZero:
		return bidgoRoundingTowardZero
	case RoundTowardPositive:
		return bidgoRoundingTowardPositive
	case RoundTowardNegative:
		return bidgoRoundingTowardNegative
	default:
		return defaultBIDRoundingMode
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

func decimal32BIDDivPort(d, other Decimal32BID) Decimal32BID {
	return Decimal32BID(bidgo.Bid32Div(d.ToUint32(), other.ToUint32(), defaultBIDRoundingMode))
}

func decimal32BIDDivPortFlags(d, other Decimal32BID) (Decimal32BID, ExceptionFlags) {
	result, flags := bidgo.Bid32DivWithFlags(d.ToUint32(), other.ToUint32(), defaultBIDRoundingMode)
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
	result, flags := bidgo.Bid32Scalbn(d.ToUint32(), exponent, defaultBIDRoundingMode)
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
	result, flags := bidgo.Bid32ToBinary32(d.ToUint32(), bidgoRoundingMode(mode))
	return math.Float32frombits(result), bidgoExceptionFlags(flags)
}

func decimal32BIDToBinary64Port(d Decimal32BID, mode RoundingMode) (float64, ExceptionFlags) {
	result, flags := bidgo.Bid32ToBinary64(d.ToUint32(), bidgoRoundingMode(mode))
	return math.Float64frombits(result), bidgoExceptionFlags(flags)
}

func decimal32BIDToBinary128Port(d Decimal32BID, mode RoundingMode) (Binary128, ExceptionFlags) {
	result, flags := bidgo.Bid32ToBinary128(d.ToUint32(), bidgoRoundingMode(mode))
	return binary128FromBidgo(result), bidgoExceptionFlags(flags)
}

func decimal32BIDToDecimal128Port(d Decimal32BID) (Decimal128BID, ExceptionFlags) {
	result, flags := bidgo.Bid32ToBid128(d.ToUint32())
	return decimal128BIDFromBidgo(result), bidgoExceptionFlags(flags)
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
	return trimTrailingIntegerSuffix(optimizedFormatDecimalString(decimal32BIDStringPort(d)))
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
	result, _ := parseDecimal32BIDPort(s)
	if invalidBIDStringInput(s, bidgo.Bid32IsNaN(result.ToUint32())) {
		return 0, fmt.Errorf("invalid decimal string: %s", s)
	}
	return result, nil
}

func parseDecimal32BIDPort(s string) (Decimal32BID, ExceptionFlags) {
	return parseDecimal32BIDPortMode(s, defaultBIDRoundingMode)
}

func parseDecimal32BIDPortMode(s string, rndMode int) (Decimal32BID, ExceptionFlags) {
	if result, ok := parseDecimal32BIDNaN(s); ok {
		return result, 0
	}
	result, flags := bidgo.Bid32FromStringRaw(s, rndMode)
	return Decimal32BID(result), bidgoExceptionFlags(flags)
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

func decimal64BIDDivPort(d, other Decimal64BID) Decimal64BID {
	return Decimal64BID(bidgo.Bid64Div(d.ToUint64(), other.ToUint64(), defaultBIDRoundingMode))
}

func decimal64BIDDivPortFlags(d, other Decimal64BID) (Decimal64BID, ExceptionFlags) {
	result, flags := bidgo.Bid64DivWithFlags(d.ToUint64(), other.ToUint64(), defaultBIDRoundingMode)
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
	result, flags := bidgo.Bid64Scalbn(d.ToUint64(), exponent, defaultBIDRoundingMode)
	return Decimal64BID(result), bidgoExceptionFlags(flags)
}

func decimal64BIDReducePort(d Decimal64BID) (Decimal64BID, ExceptionFlags) {
	result, flags := bidgo.Bid64Reduce(d.ToUint64())
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
	result, flags := bidgo.Bid64ToBinary32(d.ToUint64(), bidgoRoundingMode(mode))
	return math.Float32frombits(result), bidgoExceptionFlags(flags)
}

func decimal64BIDToBinary64Port(d Decimal64BID, mode RoundingMode) (float64, ExceptionFlags) {
	result, flags := bidgo.Bid64ToBinary64(d.ToUint64(), bidgoRoundingMode(mode))
	return math.Float64frombits(result), bidgoExceptionFlags(flags)
}

func decimal64BIDToBinary128Port(d Decimal64BID, mode RoundingMode) (Binary128, ExceptionFlags) {
	result, flags := bidgo.Bid64ToBinary128(d.ToUint64(), bidgoRoundingMode(mode))
	return binary128FromBidgo(result), bidgoExceptionFlags(flags)
}

func decimal64BIDToDecimal128Port(d Decimal64BID) (Decimal128BID, ExceptionFlags) {
	result, flags := bidgo.Bid64ToBid128(d.ToUint64())
	return decimal128BIDFromBidgo(result), bidgoExceptionFlags(flags)
}

func decimal64BIDToDecimal32Port(d Decimal64BID, mode RoundingMode) (Decimal32BID, ExceptionFlags) {
	result, flags := bidgo.Bid64ToBid32(d.ToUint64(), bidgoRoundingMode(mode))
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
	return trimTrailingIntegerSuffix(optimizedFormatDecimalString(decimal64BIDStringPort(d)))
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
	result, _ := parseDecimal64BIDPort(s)
	if invalidBIDStringInput(s, bidgo.Bid64IsNaN(result.ToUint64()) != 0) {
		return 0, fmt.Errorf("invalid decimal string: %s", s)
	}
	return result, nil
}

func parseDecimal64BIDPort(s string) (Decimal64BID, ExceptionFlags) {
	return parseDecimal64BIDPortMode(s, defaultBIDRoundingMode)
}

func parseDecimal64BIDPortMode(s string, rndMode int) (Decimal64BID, ExceptionFlags) {
	if result, ok := parseDecimal64BIDNaN(s); ok {
		return result, 0
	}
	result, flags := bidgo.Bid64FromString(s, rndMode)
	return Decimal64BID(result), bidgoExceptionFlags(flags)
}

func invalidBIDStringInput(input string, resultIsNaN bool) bool {
	return strings.TrimSpace(input) == "" || (resultIsNaN && !strings.Contains(strings.ToLower(input), "nan"))
}

func trimTrailingIntegerSuffix(s string) string {
	if strings.HasSuffix(s, ".0") {
		return s[:len(s)-2]
	}
	return s
}

func decimal128BIDStringPort(d Decimal128BID) string {
	if s, ok := formatDecimal128BIDNaN(d); ok {
		return s
	}
	return bidgo.Bid128ToString(decimal128BIDAsBidgo(d))
}

func decimal128BIDToBinary32Port(d Decimal128BID, mode RoundingMode) (float32, ExceptionFlags) {
	var flags uint32
	result := bidgo.Bid128ToBinary32(decimal128BIDAsBidgo(d), bidgoRoundingMode(mode), &flags)
	return result, bidgoExceptionFlags(flags)
}

func decimal128BIDToBinary64Port(d Decimal128BID, mode RoundingMode) (float64, ExceptionFlags) {
	var flags uint32
	result := bidgo.Bid128ToBinary64(decimal128BIDAsBidgo(d), bidgoRoundingMode(mode), &flags)
	return result, bidgoExceptionFlags(flags)
}

func decimal128BIDToBinary128Port(d Decimal128BID, mode RoundingMode) (Binary128, ExceptionFlags) {
	result, flags := bidgo.Bid128ToBinary128(decimal128BIDAsBidgo(d), bidgoRoundingMode(mode))
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
	return trimTrailingIntegerSuffix(optimizedFormatDecimalString(decimal128BIDStringPort(d)))
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

func decimal128BIDDivPort(d, other Decimal128BID) Decimal128BID {
	result, _ := bidgo.Bid128Div(decimal128BIDAsBidgo(d), decimal128BIDAsBidgo(other), defaultBIDRoundingMode)
	return decimal128BIDFromBidgo(result)
}

func decimal128BIDDivPortFlags(d, other Decimal128BID) (Decimal128BID, ExceptionFlags) {
	result, flags := bidgo.Bid128Div(decimal128BIDAsBidgo(d), decimal128BIDAsBidgo(other), defaultBIDRoundingMode)
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
	result := bidgo.Bid128Scalbn(decimal128BIDAsBidgo(d), exponent, defaultBIDRoundingMode, &flags)
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
	result, flags := bidgo.Bid128ToBid64(decimal128BIDAsBidgo(d), bidgoRoundingMode(mode))
	return Decimal64BID(result), bidgoExceptionFlags(flags)
}

func decimal128BIDToDecimal32Port(d Decimal128BID, mode RoundingMode) (Decimal32BID, ExceptionFlags) {
	result, flags := bidgo.Bid128ToBid32(decimal128BIDAsBidgo(d), bidgoRoundingMode(mode))
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
	result, _ := parseDecimal128BIDPort(s)
	if invalidBIDStringInput(s, bidgo.Bid128IsNaN(decimal128BIDAsBidgo(result)) != 0) {
		return Decimal128BID{}, fmt.Errorf("invalid decimal string: %s", s)
	}
	return result, nil
}

func parseDecimal128BIDPort(s string) (Decimal128BID, ExceptionFlags) {
	return parseDecimal128BIDPortMode(s, defaultBIDRoundingMode)
}

func parseDecimal128BIDPortMode(s string, rndMode int) (Decimal128BID, ExceptionFlags) {
	if result, ok := parseDecimal128BIDNaN(s); ok {
		return result, 0
	}
	result, flags := bidgo.Bid128FromString(s, rndMode)
	return decimal128BIDFromBidgo(result), bidgoExceptionFlags(flags)
}

func decimal128BIDAsBidgo(d Decimal128BID) bidgo.BID_UINT128 {
	raw := d.ToBytes()
	return *(*bidgo.BID_UINT128)(unsafe.Pointer(&raw))
}

func decimal128BIDFromBidgo(x bidgo.BID_UINT128) Decimal128BID {
	return *(*Decimal128BID)(unsafe.Pointer(&x))
}

func binary128FromBidgo(x bidgo.BID_UINT128) Binary128 {
	return *(*Binary128)(unsafe.Pointer(&x))
}
