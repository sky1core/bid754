package bid754

// Add returns the IEEE 754 Decimal32BID sum d + other.
func (d Decimal32BID) Add(other Decimal32BID) Decimal32BID { return decimal32BIDAddPort(d, other) }

// AddWithFlags returns the IEEE 754 Decimal32BID sum d + other and the exception flags raised by the operation.
func (d Decimal32BID) AddWithFlags(other Decimal32BID) (Decimal32BID, ExceptionFlags) {
	return decimal32BIDAddPortFlags(d, other)
}

// Sub returns the IEEE 754 Decimal32BID difference d - other.
func (d Decimal32BID) Sub(other Decimal32BID) Decimal32BID { return decimal32BIDSubPort(d, other) }

// SubWithFlags returns the IEEE 754 Decimal32BID difference d - other and the exception flags raised by the operation.
func (d Decimal32BID) SubWithFlags(other Decimal32BID) (Decimal32BID, ExceptionFlags) {
	return decimal32BIDSubPortFlags(d, other)
}

// Mul returns the IEEE 754 Decimal32BID product d * other.
func (d Decimal32BID) Mul(other Decimal32BID) Decimal32BID { return decimal32BIDMulPort(d, other) }

// MulWithFlags returns the IEEE 754 Decimal32BID product d * other and the exception flags raised by the operation.
func (d Decimal32BID) MulWithFlags(other Decimal32BID) (Decimal32BID, ExceptionFlags) {
	return decimal32BIDMulPortFlags(d, other)
}

// FMA returns the IEEE 754 Decimal32BID fused multiply-add d*mul + add and the exception flags raised by the operation.
func (d Decimal32BID) FMA(mul, add Decimal32BID) (Decimal32BID, ExceptionFlags) {
	return decimal32BIDFMAPort(d, mul, add)
}

// Sqrt returns the IEEE 754 Decimal32BID square root of d and the exception flags raised by the operation.
func (d Decimal32BID) Sqrt() (Decimal32BID, ExceptionFlags) { return decimal32BIDSqrtPort(d) }

// Div returns the IEEE 754 Decimal32BID quotient d / other.
func (d Decimal32BID) Div(other Decimal32BID) Decimal32BID { return decimal32BIDDivPort(d, other) }

// DivWithFlags returns the IEEE 754 Decimal32BID quotient d / other and the exception flags raised by the operation.
func (d Decimal32BID) DivWithFlags(other Decimal32BID) (Decimal32BID, ExceptionFlags) {
	return decimal32BIDDivPortFlags(d, other)
}

// Remainder returns the IEEE 754-2019 clause 5.3.1 remainder (remainder-near),
// routed through the Go mechanical port of Intel bid32_rem.
func (d Decimal32BID) Remainder(other Decimal32BID) (Decimal32BID, ExceptionFlags) {
	return decimal32BIDRemainderPort(d, other)
}

// Fmod returns the truncated-division remainder (Intel bid32_fmod; the GDA
// decTest "remainder" operation), which is not the IEEE 754 remainder.
func (d Decimal32BID) Fmod(other Decimal32BID) (Decimal32BID, ExceptionFlags) {
	return decimal32BIDFmodPort(d, other)
}

// Quantize returns d quantized to the exponent of other as a Decimal32BID value.
func (d Decimal32BID) Quantize(other Decimal32BID) Decimal32BID {
	return decimal32BIDQuantizePort(d, other)
}

// QuantizeWithFlags returns d quantized to the exponent of other as a Decimal32BID value and the exception flags raised by the operation.
func (d Decimal32BID) QuantizeWithFlags(other Decimal32BID) (Decimal32BID, ExceptionFlags) {
	return decimal32BIDQuantizePortFlags(d, other)
}

// SameQuantum reports whether d and other have the same quantum.
func (d Decimal32BID) SameQuantum(other Decimal32BID) bool {
	return decimal32BIDSameQuantumPort(d, other)
}

// MinNum returns the IEEE 754 Decimal32BID minimum numeric value of d and other and the exception flags raised by the operation.
func (d Decimal32BID) MinNum(other Decimal32BID) (Decimal32BID, ExceptionFlags) {
	return decimal32BIDMinNumPort(d, other)
}

// MaxNum returns the IEEE 754 Decimal32BID maximum numeric value of d and other and the exception flags raised by the operation.
func (d Decimal32BID) MaxNum(other Decimal32BID) (Decimal32BID, ExceptionFlags) {
	return decimal32BIDMaxNumPort(d, other)
}

// MinNumMag returns the IEEE 754 Decimal32BID operand with minimum numeric magnitude and the exception flags raised by the operation.
func (d Decimal32BID) MinNumMag(other Decimal32BID) (Decimal32BID, ExceptionFlags) {
	return decimal32BIDMinNumMagPort(d, other)
}

// MaxNumMag returns the IEEE 754 Decimal32BID operand with maximum numeric magnitude and the exception flags raised by the operation.
func (d Decimal32BID) MaxNumMag(other Decimal32BID) (Decimal32BID, ExceptionFlags) {
	return decimal32BIDMaxNumMagPort(d, other)
}

// CompareTotal returns the IEEE 754 totalOrder comparison of d and other: -1 if d orders before other, 0 if equal, +1 if after.
func (d Decimal32BID) CompareTotal(other Decimal32BID) int {
	return decimal32BIDCompareTotalPort(d, other)
}

// CompareTotalMag returns the IEEE 754 totalOrderMag comparison of |d| and |other|: -1, 0, or +1.
func (d Decimal32BID) CompareTotalMag(other Decimal32BID) int {
	return decimal32BIDCompareTotalMagPort(d, other)
}

// RoundIntegralExact returns d rounded to an integral Decimal32BID value using round-to-nearest-even (it does not consult SetDefaultRounding).
func (d Decimal32BID) RoundIntegralExact() Decimal32BID { return decimal32BIDRoundIntegralExactPort(d) }

// RoundIntegralExactWithFlags returns d rounded to an integral Decimal32BID value using round-to-nearest-even (it does not consult SetDefaultRounding) and the exception flags raised by the operation.
func (d Decimal32BID) RoundIntegralExactWithFlags() (Decimal32BID, ExceptionFlags) {
	return decimal32BIDRoundIntegralExactPortFlags(d)
}

// RoundIntegralNearestEven implements IEEE roundToIntegralTiesToEven.
func (d Decimal32BID) RoundIntegralNearestEven() (Decimal32BID, ExceptionFlags) {
	return decimal32BIDRoundIntegralNearestEvenPort(d)
}

// RoundIntegralNearestAway implements IEEE roundToIntegralTiesToAway.
func (d Decimal32BID) RoundIntegralNearestAway() (Decimal32BID, ExceptionFlags) {
	return decimal32BIDRoundIntegralNearestAwayPort(d)
}

// RoundIntegralZero implements IEEE roundToIntegralTowardZero.
func (d Decimal32BID) RoundIntegralZero() (Decimal32BID, ExceptionFlags) {
	return decimal32BIDRoundIntegralZeroPort(d)
}

// RoundIntegralPositive implements IEEE roundToIntegralTowardPositive.
func (d Decimal32BID) RoundIntegralPositive() (Decimal32BID, ExceptionFlags) {
	return decimal32BIDRoundIntegralPositivePort(d)
}

// RoundIntegralNegative implements IEEE roundToIntegralTowardNegative.
func (d Decimal32BID) RoundIntegralNegative() (Decimal32BID, ExceptionFlags) {
	return decimal32BIDRoundIntegralNegativePort(d)
}

// LogB returns the IEEE 754 logB result for d and the exception flags raised by the operation.
func (d Decimal32BID) LogB() (Decimal32BID, ExceptionFlags) {
	return decimal32BIDLogBPort(d)
}

// ScaleB returns the IEEE 754 scaleB result for d scaled by exponent and the exception flags raised by the operation.
func (d Decimal32BID) ScaleB(exponent int) (Decimal32BID, ExceptionFlags) {
	return decimal32BIDScaleBPort(d, exponent)
}

// Copy returns a Decimal32BID copy of d.
func (d Decimal32BID) Copy() Decimal32BID { return decimal32BIDCopyPort(d) }

// Abs returns the Decimal32BID value with the magnitude of d and a positive sign.
func (d Decimal32BID) Abs() Decimal32BID { return decimal32BIDAbsPort(d) }

// Negate returns the Decimal32BID value with the magnitude of d and the opposite sign.
func (d Decimal32BID) Negate() Decimal32BID { return decimal32BIDNegatePort(d) }

// CopySign returns the Decimal32BID value with the magnitude of d and the sign of signSource.
func (d Decimal32BID) CopySign(signSource Decimal32BID) Decimal32BID {
	return decimal32BIDCopySignPort(d, signSource)
}

// String returns the decimal string form of d produced by the Go mechanical port.
func (d Decimal32BID) String() string { return decimal32BIDStringPort(d) }

// PrettyString returns the display-oriented string form of d.
func (d Decimal32BID) PrettyString() string { return decimal32BIDPrettyStringPort(d) }

// ToBinary32 converts d to binary32 using mode and returns the exception flags raised by the operation.
// A RoundingMode outside the defined constants panics.
func (d Decimal32BID) ToBinary32(mode RoundingMode) (float32, ExceptionFlags) {
	return decimal32BIDToBinary32Port(d, mode)
}

// ToBinary64 converts d to binary64 using mode and returns the exception flags raised by the operation.
// A RoundingMode outside the defined constants panics.
func (d Decimal32BID) ToBinary64(mode RoundingMode) (float64, ExceptionFlags) {
	return decimal32BIDToBinary64Port(d, mode)
}

// ToBinary128 converts d to binary128 using mode and returns the exception flags raised by the operation.
// A RoundingMode outside the defined constants panics.
func (d Decimal32BID) ToBinary128(mode RoundingMode) (Binary128, ExceptionFlags) {
	return decimal32BIDToBinary128Port(d, mode)
}

// ToDecimal128 converts d to the wider Decimal128BID format and returns the exception flags raised by the operation.
func (d Decimal32BID) ToDecimal128() (Decimal128BID, ExceptionFlags) {
	return decimal32BIDToDecimal128Port(d)
}

// ToDecimal64 converts to the wider Decimal64 format (IEEE 754 convertFormat,
// clause 5.4.2); widening is exact.
func (d Decimal32BID) ToDecimal64() (Decimal64BID, ExceptionFlags) {
	return decimal32BIDToDecimal64Port(d)
}

// NextToward returns the next representable Decimal32BID value from d toward target and the exception flags raised by the operation.
func (d Decimal32BID) NextToward(target Decimal128BID) (Decimal32BID, ExceptionFlags) {
	return decimal32BIDNextTowardPort(d, target)
}

// NextPlus returns the next representable Decimal32BID value greater than d and the exception flags raised by the operation.
func (d Decimal32BID) NextPlus() (Decimal32BID, ExceptionFlags) {
	return decimal32BIDNextPlusPort(d)
}

// NextMinus returns the next representable Decimal32BID value less than d and the exception flags raised by the operation.
func (d Decimal32BID) NextMinus() (Decimal32BID, ExceptionFlags) {
	return decimal32BIDNextMinusPort(d)
}

// IsZero reports whether d is zero.
func (d Decimal32BID) IsZero() bool { return decimal32BIDIsZeroPort(d) }

// IsNaN reports whether d is a NaN.
func (d Decimal32BID) IsNaN() bool { return decimal32BIDIsNaNPort(d) }

// IsInf reports whether d is an infinity.
func (d Decimal32BID) IsInf() bool { return decimal32BIDIsInfPort(d) }

// IsNormal reports whether d is normal.
func (d Decimal32BID) IsNormal() bool { return decimal32BIDIsNormalPort(d) }

// IsFinite reports whether d is finite.
func (d Decimal32BID) IsFinite() bool { return decimal32BIDIsFinitePort(d) }

// IsSubnormal reports whether d is subnormal.
func (d Decimal32BID) IsSubnormal() bool { return decimal32BIDIsSubnormalPort(d) }

// IsSignaling reports whether d is a signaling NaN.
func (d Decimal32BID) IsSignaling() bool { return decimal32BIDIsSignalingPort(d) }

// IsCanonical reports whether d is canonical.
func (d Decimal32BID) IsCanonical() bool { return decimal32BIDIsCanonicalPort(d) }

// Radix returns the radix of the Decimal32BID format.
func (d Decimal32BID) Radix() int { return decimal32BIDRadixPort() }

// IsSignMinus reports whether d has a negative sign.
func (d Decimal32BID) IsSignMinus() bool {
	return decimal32BIDIsSignMinusPort(d)
}

// Class returns the IEEE 754 class of d.
func (d Decimal32BID) Class() DecimalClass {
	return decimal32BIDClassPort(d)
}

// Sign returns 0 if d is zero (either sign), and otherwise -1 or +1 from the sign bit of d (including for NaNs).
func (d Decimal32BID) Sign() int { return decimal32BIDSignPort(d) }

// Add returns the IEEE 754 Decimal64BID sum d + other.
func (d Decimal64BID) Add(other Decimal64BID) Decimal64BID { return decimal64BIDAddPort(d, other) }

// AddWithFlags returns the IEEE 754 Decimal64BID sum d + other and the exception flags raised by the operation.
func (d Decimal64BID) AddWithFlags(other Decimal64BID) (Decimal64BID, ExceptionFlags) {
	return decimal64BIDAddPortFlags(d, other)
}

// Sub returns the IEEE 754 Decimal64BID difference d - other.
func (d Decimal64BID) Sub(other Decimal64BID) Decimal64BID { return decimal64BIDSubPort(d, other) }

// SubWithFlags returns the IEEE 754 Decimal64BID difference d - other and the exception flags raised by the operation.
func (d Decimal64BID) SubWithFlags(other Decimal64BID) (Decimal64BID, ExceptionFlags) {
	return decimal64BIDSubPortFlags(d, other)
}

// Mul returns the IEEE 754 Decimal64BID product d * other.
func (d Decimal64BID) Mul(other Decimal64BID) Decimal64BID { return decimal64BIDMulPort(d, other) }

// MulWithFlags returns the IEEE 754 Decimal64BID product d * other and the exception flags raised by the operation.
func (d Decimal64BID) MulWithFlags(other Decimal64BID) (Decimal64BID, ExceptionFlags) {
	return decimal64BIDMulPortFlags(d, other)
}

// FMA returns the IEEE 754 Decimal64BID fused multiply-add d*mul + add and the exception flags raised by the operation.
func (d Decimal64BID) FMA(mul, add Decimal64BID) (Decimal64BID, ExceptionFlags) {
	return decimal64BIDFMAPort(d, mul, add)
}

// Sqrt returns the IEEE 754 Decimal64BID square root of d and the exception flags raised by the operation.
func (d Decimal64BID) Sqrt() (Decimal64BID, ExceptionFlags) { return decimal64BIDSqrtPort(d) }

// Div returns the IEEE 754 Decimal64BID quotient d / other.
func (d Decimal64BID) Div(other Decimal64BID) Decimal64BID { return decimal64BIDDivPort(d, other) }

// DivWithFlags returns the IEEE 754 Decimal64BID quotient d / other and the exception flags raised by the operation.
func (d Decimal64BID) DivWithFlags(other Decimal64BID) (Decimal64BID, ExceptionFlags) {
	return decimal64BIDDivPortFlags(d, other)
}

// Remainder returns the IEEE 754-2019 clause 5.3.1 remainder (remainder-near),
// routed through the Go mechanical port of Intel bid64_rem.
func (d Decimal64BID) Remainder(other Decimal64BID) (Decimal64BID, ExceptionFlags) {
	return decimal64BIDRemainderPort(d, other)
}

// Fmod returns the truncated-division remainder (Intel bid64_fmod; the GDA
// decTest "remainder" operation), which is not the IEEE 754 remainder.
func (d Decimal64BID) Fmod(other Decimal64BID) (Decimal64BID, ExceptionFlags) {
	return decimal64BIDFmodPort(d, other)
}

// Quantize returns d quantized to the exponent of other as a Decimal64BID value.
func (d Decimal64BID) Quantize(other Decimal64BID) Decimal64BID {
	return decimal64BIDQuantizePort(d, other)
}

// QuantizeWithFlags returns d quantized to the exponent of other as a Decimal64BID value and the exception flags raised by the operation.
func (d Decimal64BID) QuantizeWithFlags(other Decimal64BID) (Decimal64BID, ExceptionFlags) {
	return decimal64BIDQuantizePortFlags(d, other)
}

// SameQuantum reports whether d and other have the same quantum.
func (d Decimal64BID) SameQuantum(other Decimal64BID) bool {
	return decimal64BIDSameQuantumPort(d, other)
}

// MinNum returns the IEEE 754 Decimal64BID minimum numeric value of d and other and the exception flags raised by the operation.
func (d Decimal64BID) MinNum(other Decimal64BID) (Decimal64BID, ExceptionFlags) {
	return decimal64BIDMinNumPort(d, other)
}

// MaxNum returns the IEEE 754 Decimal64BID maximum numeric value of d and other and the exception flags raised by the operation.
func (d Decimal64BID) MaxNum(other Decimal64BID) (Decimal64BID, ExceptionFlags) {
	return decimal64BIDMaxNumPort(d, other)
}

// MinNumMag returns the IEEE 754 Decimal64BID operand with minimum numeric magnitude and the exception flags raised by the operation.
func (d Decimal64BID) MinNumMag(other Decimal64BID) (Decimal64BID, ExceptionFlags) {
	return decimal64BIDMinNumMagPort(d, other)
}

// MaxNumMag returns the IEEE 754 Decimal64BID operand with maximum numeric magnitude and the exception flags raised by the operation.
func (d Decimal64BID) MaxNumMag(other Decimal64BID) (Decimal64BID, ExceptionFlags) {
	return decimal64BIDMaxNumMagPort(d, other)
}

// CompareTotal returns the IEEE 754 totalOrder comparison of d and other: -1 if d orders before other, 0 if equal, +1 if after.
func (d Decimal64BID) CompareTotal(other Decimal64BID) int {
	return decimal64BIDCompareTotalPort(d, other)
}

// CompareTotalMag returns the IEEE 754 totalOrderMag comparison of |d| and |other|: -1, 0, or +1.
func (d Decimal64BID) CompareTotalMag(other Decimal64BID) int {
	return decimal64BIDCompareTotalMagPort(d, other)
}

// RoundIntegralExact returns d rounded to an integral Decimal64BID value using round-to-nearest-even (it does not consult SetDefaultRounding).
func (d Decimal64BID) RoundIntegralExact() Decimal64BID { return decimal64BIDRoundIntegralExactPort(d) }

// RoundIntegralExactWithFlags returns d rounded to an integral Decimal64BID value using round-to-nearest-even (it does not consult SetDefaultRounding) and the exception flags raised by the operation.
func (d Decimal64BID) RoundIntegralExactWithFlags() (Decimal64BID, ExceptionFlags) {
	return decimal64BIDRoundIntegralExactPortFlags(d)
}

// RoundIntegralNearestEven implements IEEE roundToIntegralTiesToEven.
func (d Decimal64BID) RoundIntegralNearestEven() (Decimal64BID, ExceptionFlags) {
	return decimal64BIDRoundIntegralNearestEvenPort(d)
}

// RoundIntegralNearestAway implements IEEE roundToIntegralTiesToAway.
func (d Decimal64BID) RoundIntegralNearestAway() (Decimal64BID, ExceptionFlags) {
	return decimal64BIDRoundIntegralNearestAwayPort(d)
}

// RoundIntegralZero implements IEEE roundToIntegralTowardZero.
func (d Decimal64BID) RoundIntegralZero() (Decimal64BID, ExceptionFlags) {
	return decimal64BIDRoundIntegralZeroPort(d)
}

// RoundIntegralPositive implements IEEE roundToIntegralTowardPositive.
func (d Decimal64BID) RoundIntegralPositive() (Decimal64BID, ExceptionFlags) {
	return decimal64BIDRoundIntegralPositivePort(d)
}

// RoundIntegralNegative implements IEEE roundToIntegralTowardNegative.
func (d Decimal64BID) RoundIntegralNegative() (Decimal64BID, ExceptionFlags) {
	return decimal64BIDRoundIntegralNegativePort(d)
}

// LogB returns the IEEE 754 logB result for d and the exception flags raised by the operation.
func (d Decimal64BID) LogB() (Decimal64BID, ExceptionFlags) {
	return decimal64BIDLogBPort(d)
}

// ScaleB returns the IEEE 754 scaleB result for d scaled by exponent and the exception flags raised by the operation.
func (d Decimal64BID) ScaleB(exponent int) (Decimal64BID, ExceptionFlags) {
	return decimal64BIDScaleBPort(d, exponent)
}

// Reduce returns d in reduced Decimal64BID form and the exception flags raised by the operation.
func (d Decimal64BID) Reduce() (Decimal64BID, ExceptionFlags) {
	return decimal64BIDReducePort(d)
}

// Copy returns a Decimal64BID copy of d.
func (d Decimal64BID) Copy() Decimal64BID { return decimal64BIDCopyPort(d) }

// Abs returns the Decimal64BID value with the magnitude of d and a positive sign.
func (d Decimal64BID) Abs() Decimal64BID { return decimal64BIDAbsPort(d) }

// Negate returns the Decimal64BID value with the magnitude of d and the opposite sign.
func (d Decimal64BID) Negate() Decimal64BID { return decimal64BIDNegatePort(d) }

// CopySign returns the Decimal64BID value with the magnitude of d and the sign of signSource.
func (d Decimal64BID) CopySign(signSource Decimal64BID) Decimal64BID {
	return decimal64BIDCopySignPort(d, signSource)
}

// String returns the decimal string form of d produced by the Go mechanical port.
func (d Decimal64BID) String() string { return decimal64BIDStringPort(d) }

// PrettyString returns the display-oriented string form of d.
func (d Decimal64BID) PrettyString() string { return decimal64BIDPrettyStringPort(d) }

// ToBinary32 converts d to binary32 using mode and returns the exception flags raised by the operation.
// A RoundingMode outside the defined constants panics.
func (d Decimal64BID) ToBinary32(mode RoundingMode) (float32, ExceptionFlags) {
	return decimal64BIDToBinary32Port(d, mode)
}

// ToBinary64 converts d to binary64 using mode and returns the exception flags raised by the operation.
// A RoundingMode outside the defined constants panics.
func (d Decimal64BID) ToBinary64(mode RoundingMode) (float64, ExceptionFlags) {
	return decimal64BIDToBinary64Port(d, mode)
}

// ToBinary128 converts d to binary128 using mode and returns the exception flags raised by the operation.
// A RoundingMode outside the defined constants panics.
func (d Decimal64BID) ToBinary128(mode RoundingMode) (Binary128, ExceptionFlags) {
	return decimal64BIDToBinary128Port(d, mode)
}

// ToDecimal128 converts d to the wider Decimal128BID format and returns the exception flags raised by the operation.
func (d Decimal64BID) ToDecimal128() (Decimal128BID, ExceptionFlags) {
	return decimal64BIDToDecimal128Port(d)
}

// ToDecimal32 converts to the narrower Decimal32 format (IEEE 754
// convertFormat, clause 5.4.2), rounding with the given mode. A RoundingMode
// outside the defined constants panics.
func (d Decimal64BID) ToDecimal32(mode RoundingMode) (Decimal32BID, ExceptionFlags) {
	return decimal64BIDToDecimal32Port(d, mode)
}

// NextToward returns the next representable Decimal64BID value from d toward target and the exception flags raised by the operation.
func (d Decimal64BID) NextToward(target Decimal128BID) (Decimal64BID, ExceptionFlags) {
	return decimal64BIDNextTowardPort(d, target)
}

// NextPlus returns the next representable Decimal64BID value greater than d and the exception flags raised by the operation.
func (d Decimal64BID) NextPlus() (Decimal64BID, ExceptionFlags) {
	return decimal64BIDNextPlusPort(d)
}

// NextMinus returns the next representable Decimal64BID value less than d and the exception flags raised by the operation.
func (d Decimal64BID) NextMinus() (Decimal64BID, ExceptionFlags) {
	return decimal64BIDNextMinusPort(d)
}

// IsZero reports whether d is zero.
func (d Decimal64BID) IsZero() bool { return decimal64BIDIsZeroPort(d) }

// IsNaN reports whether d is a NaN.
func (d Decimal64BID) IsNaN() bool { return decimal64BIDIsNaNPort(d) }

// IsInf reports whether d is an infinity.
func (d Decimal64BID) IsInf() bool { return decimal64BIDIsInfPort(d) }

// IsNormal reports whether d is normal.
func (d Decimal64BID) IsNormal() bool { return decimal64BIDIsNormalPort(d) }

// IsFinite reports whether d is finite.
func (d Decimal64BID) IsFinite() bool { return decimal64BIDIsFinitePort(d) }

// IsSubnormal reports whether d is subnormal.
func (d Decimal64BID) IsSubnormal() bool { return decimal64BIDIsSubnormalPort(d) }

// IsSignaling reports whether d is a signaling NaN.
func (d Decimal64BID) IsSignaling() bool { return decimal64BIDIsSignalingPort(d) }

// IsCanonical reports whether d is canonical.
func (d Decimal64BID) IsCanonical() bool { return decimal64BIDIsCanonicalPort(d) }

// Radix returns the radix of the Decimal64BID format.
func (d Decimal64BID) Radix() int { return decimal64BIDRadixPort() }

// IsSignMinus reports whether d has a negative sign.
func (d Decimal64BID) IsSignMinus() bool {
	return decimal64BIDIsSignMinusPort(d)
}

// Class returns the IEEE 754 class of d.
func (d Decimal64BID) Class() DecimalClass {
	return decimal64BIDClassPort(d)
}

// Sign returns 0 if d is zero (either sign), and otherwise -1 or +1 from the sign bit of d (including for NaNs).
func (d Decimal64BID) Sign() int { return decimal64BIDSignPort(d) }

// Add returns the IEEE 754 Decimal128BID sum d + other.
func (d Decimal128BID) Add(other Decimal128BID) Decimal128BID { return decimal128BIDAddPort(d, other) }

// AddWithFlags returns the IEEE 754 Decimal128BID sum d + other and the exception flags raised by the operation.
func (d Decimal128BID) AddWithFlags(other Decimal128BID) (Decimal128BID, ExceptionFlags) {
	return decimal128BIDAddPortFlags(d, other)
}

// Sub returns the IEEE 754 Decimal128BID difference d - other.
func (d Decimal128BID) Sub(other Decimal128BID) Decimal128BID { return decimal128BIDSubPort(d, other) }

// SubWithFlags returns the IEEE 754 Decimal128BID difference d - other and the exception flags raised by the operation.
func (d Decimal128BID) SubWithFlags(other Decimal128BID) (Decimal128BID, ExceptionFlags) {
	return decimal128BIDSubPortFlags(d, other)
}

// Mul returns the IEEE 754 Decimal128BID product d * other.
func (d Decimal128BID) Mul(other Decimal128BID) Decimal128BID { return decimal128BIDMulPort(d, other) }

// MulWithFlags returns the IEEE 754 Decimal128BID product d * other and the exception flags raised by the operation.
func (d Decimal128BID) MulWithFlags(other Decimal128BID) (Decimal128BID, ExceptionFlags) {
	return decimal128BIDMulPortFlags(d, other)
}

// FMA returns the IEEE 754 Decimal128BID fused multiply-add d*mul + add and the exception flags raised by the operation.
func (d Decimal128BID) FMA(mul, add Decimal128BID) (Decimal128BID, ExceptionFlags) {
	return decimal128BIDFMAPort(d, mul, add)
}

// Sqrt returns the IEEE 754 Decimal128BID square root of d and the exception flags raised by the operation.
func (d Decimal128BID) Sqrt() (Decimal128BID, ExceptionFlags) { return decimal128BIDSqrtPort(d) }

// Div returns the IEEE 754 Decimal128BID quotient d / other.
func (d Decimal128BID) Div(other Decimal128BID) Decimal128BID { return decimal128BIDDivPort(d, other) }

// DivWithFlags returns the IEEE 754 Decimal128BID quotient d / other and the exception flags raised by the operation.
func (d Decimal128BID) DivWithFlags(other Decimal128BID) (Decimal128BID, ExceptionFlags) {
	return decimal128BIDDivPortFlags(d, other)
}

// Remainder returns the IEEE 754-2019 clause 5.3.1 remainder (remainder-near),
// routed through the Go mechanical port of Intel bid128_rem.
func (d Decimal128BID) Remainder(other Decimal128BID) (Decimal128BID, ExceptionFlags) {
	return decimal128BIDRemainderPort(d, other)
}

// Fmod returns the truncated-division remainder (Intel bid128_fmod; the GDA
// decTest "remainder" operation), which is not the IEEE 754 remainder.
func (d Decimal128BID) Fmod(other Decimal128BID) (Decimal128BID, ExceptionFlags) {
	return decimal128BIDFmodPort(d, other)
}

// ToDecimal64 converts to the narrower Decimal64 format (IEEE 754
// convertFormat, clause 5.4.2), rounding with the given mode. A RoundingMode
// outside the defined constants panics.
func (d Decimal128BID) ToDecimal64(mode RoundingMode) (Decimal64BID, ExceptionFlags) {
	return decimal128BIDToDecimal64Port(d, mode)
}

// ToDecimal32 converts to the narrower Decimal32 format (IEEE 754
// convertFormat, clause 5.4.2), rounding with the given mode. A RoundingMode
// outside the defined constants panics.
func (d Decimal128BID) ToDecimal32(mode RoundingMode) (Decimal32BID, ExceptionFlags) {
	return decimal128BIDToDecimal32Port(d, mode)
}

// Quantize returns d quantized to the exponent of other as a Decimal128BID value.
func (d Decimal128BID) Quantize(other Decimal128BID) Decimal128BID {
	return decimal128BIDQuantizePort(d, other)
}

// QuantizeWithFlags returns d quantized to the exponent of other as a Decimal128BID value and the exception flags raised by the operation.
func (d Decimal128BID) QuantizeWithFlags(other Decimal128BID) (Decimal128BID, ExceptionFlags) {
	return decimal128BIDQuantizePortFlags(d, other)
}

// SameQuantum reports whether d and other have the same quantum.
func (d Decimal128BID) SameQuantum(other Decimal128BID) bool {
	return decimal128BIDSameQuantumPort(d, other)
}

// MinNum returns the IEEE 754 Decimal128BID minimum numeric value of d and other and the exception flags raised by the operation.
func (d Decimal128BID) MinNum(other Decimal128BID) (Decimal128BID, ExceptionFlags) {
	return decimal128BIDMinNumPort(d, other)
}

// MaxNum returns the IEEE 754 Decimal128BID maximum numeric value of d and other and the exception flags raised by the operation.
func (d Decimal128BID) MaxNum(other Decimal128BID) (Decimal128BID, ExceptionFlags) {
	return decimal128BIDMaxNumPort(d, other)
}

// MinNumMag returns the IEEE 754 Decimal128BID operand with minimum numeric magnitude and the exception flags raised by the operation.
func (d Decimal128BID) MinNumMag(other Decimal128BID) (Decimal128BID, ExceptionFlags) {
	return decimal128BIDMinNumMagPort(d, other)
}

// MaxNumMag returns the IEEE 754 Decimal128BID operand with maximum numeric magnitude and the exception flags raised by the operation.
func (d Decimal128BID) MaxNumMag(other Decimal128BID) (Decimal128BID, ExceptionFlags) {
	return decimal128BIDMaxNumMagPort(d, other)
}

// CompareTotal returns the IEEE 754 totalOrder comparison of d and other: -1 if d orders before other, 0 if equal, +1 if after.
func (d Decimal128BID) CompareTotal(other Decimal128BID) int {
	return decimal128BIDCompareTotalPort(d, other)
}

// CompareTotalMag returns the IEEE 754 totalOrderMag comparison of |d| and |other|: -1, 0, or +1.
func (d Decimal128BID) CompareTotalMag(other Decimal128BID) int {
	return decimal128BIDCompareTotalMagPort(d, other)
}

// RoundIntegralExact returns d rounded to an integral Decimal128BID value using round-to-nearest-even (it does not consult SetDefaultRounding).
func (d Decimal128BID) RoundIntegralExact() Decimal128BID {
	return decimal128BIDRoundIntegralExactPort(d)
}

// RoundIntegralExactWithFlags returns d rounded to an integral Decimal128BID value using round-to-nearest-even (it does not consult SetDefaultRounding) and the exception flags raised by the operation.
func (d Decimal128BID) RoundIntegralExactWithFlags() (Decimal128BID, ExceptionFlags) {
	return decimal128BIDRoundIntegralExactPortFlags(d)
}

// RoundIntegralNearestEven implements IEEE roundToIntegralTiesToEven.
func (d Decimal128BID) RoundIntegralNearestEven() (Decimal128BID, ExceptionFlags) {
	return decimal128BIDRoundIntegralNearestEvenPort(d)
}

// RoundIntegralNearestAway implements IEEE roundToIntegralTiesToAway.
func (d Decimal128BID) RoundIntegralNearestAway() (Decimal128BID, ExceptionFlags) {
	return decimal128BIDRoundIntegralNearestAwayPort(d)
}

// RoundIntegralZero implements IEEE roundToIntegralTowardZero.
func (d Decimal128BID) RoundIntegralZero() (Decimal128BID, ExceptionFlags) {
	return decimal128BIDRoundIntegralZeroPort(d)
}

// RoundIntegralPositive implements IEEE roundToIntegralTowardPositive.
func (d Decimal128BID) RoundIntegralPositive() (Decimal128BID, ExceptionFlags) {
	return decimal128BIDRoundIntegralPositivePort(d)
}

// RoundIntegralNegative implements IEEE roundToIntegralTowardNegative.
func (d Decimal128BID) RoundIntegralNegative() (Decimal128BID, ExceptionFlags) {
	return decimal128BIDRoundIntegralNegativePort(d)
}

// LogB returns the IEEE 754 logB result for d and the exception flags raised by the operation.
func (d Decimal128BID) LogB() (Decimal128BID, ExceptionFlags) {
	return decimal128BIDLogBPort(d)
}

// ScaleB returns the IEEE 754 scaleB result for d scaled by exponent and the exception flags raised by the operation.
func (d Decimal128BID) ScaleB(exponent int) (Decimal128BID, ExceptionFlags) {
	return decimal128BIDScaleBPort(d, exponent)
}

// Copy returns a Decimal128BID copy of d.
func (d Decimal128BID) Copy() Decimal128BID { return decimal128BIDCopyPort(d) }

// Abs returns the Decimal128BID value with the magnitude of d and a positive sign.
func (d Decimal128BID) Abs() Decimal128BID { return decimal128BIDAbsPort(d) }

// Negate returns the Decimal128BID value with the magnitude of d and the opposite sign.
func (d Decimal128BID) Negate() Decimal128BID { return decimal128BIDNegatePort(d) }

// CopySign returns the Decimal128BID value with the magnitude of d and the sign of signSource.
func (d Decimal128BID) CopySign(signSource Decimal128BID) Decimal128BID {
	return decimal128BIDCopySignPort(d, signSource)
}

// String returns the decimal string form of d produced by the Go mechanical port.
func (d Decimal128BID) String() string { return decimal128BIDStringPort(d) }

// PrettyString returns the display-oriented string form of d.
func (d Decimal128BID) PrettyString() string {
	return decimal128BIDPrettyStringPort(d)
}

// ToBinary32 converts d to binary32 using mode and returns the exception flags raised by the operation.
// A RoundingMode outside the defined constants panics.
func (d Decimal128BID) ToBinary32(mode RoundingMode) (float32, ExceptionFlags) {
	return decimal128BIDToBinary32Port(d, mode)
}

// ToBinary64 converts d to binary64 using mode and returns the exception flags raised by the operation.
// A RoundingMode outside the defined constants panics.
func (d Decimal128BID) ToBinary64(mode RoundingMode) (float64, ExceptionFlags) {
	return decimal128BIDToBinary64Port(d, mode)
}

// ToBinary128 converts d to binary128 using mode and returns the exception flags raised by the operation.
// A RoundingMode outside the defined constants panics.
func (d Decimal128BID) ToBinary128(mode RoundingMode) (Binary128, ExceptionFlags) {
	return decimal128BIDToBinary128Port(d, mode)
}

// NextToward returns the next representable Decimal128BID value from d toward target and the exception flags raised by the operation.
func (d Decimal128BID) NextToward(target Decimal128BID) (Decimal128BID, ExceptionFlags) {
	return decimal128BIDNextTowardPort(d, target)
}

// NextPlus returns the next representable Decimal128BID value greater than d and the exception flags raised by the operation.
func (d Decimal128BID) NextPlus() (Decimal128BID, ExceptionFlags) {
	return decimal128BIDNextPlusPort(d)
}

// NextMinus returns the next representable Decimal128BID value less than d and the exception flags raised by the operation.
func (d Decimal128BID) NextMinus() (Decimal128BID, ExceptionFlags) {
	return decimal128BIDNextMinusPort(d)
}

// IsZero reports whether d is zero.
func (d Decimal128BID) IsZero() bool { return decimal128BIDIsZeroPort(d) }

// IsNaN reports whether d is a NaN.
func (d Decimal128BID) IsNaN() bool { return decimal128BIDIsNaNPort(d) }

// IsInf reports whether d is an infinity.
func (d Decimal128BID) IsInf() bool { return decimal128BIDIsInfPort(d) }

// IsNormal reports whether d is normal.
func (d Decimal128BID) IsNormal() bool { return decimal128BIDIsNormalPort(d) }

// IsFinite reports whether d is finite.
func (d Decimal128BID) IsFinite() bool { return decimal128BIDIsFinitePort(d) }

// IsSubnormal reports whether d is subnormal.
func (d Decimal128BID) IsSubnormal() bool { return decimal128BIDIsSubnormalPort(d) }

// IsSignaling reports whether d is a signaling NaN.
func (d Decimal128BID) IsSignaling() bool { return decimal128BIDIsSignalingPort(d) }

// IsCanonical reports whether d is canonical.
func (d Decimal128BID) IsCanonical() bool { return decimal128BIDIsCanonicalPort(d) }

// Radix returns the radix of the Decimal128BID format.
func (d Decimal128BID) Radix() int { return decimal128BIDRadixPort() }

// IsSignMinus reports whether d has a negative sign.
func (d Decimal128BID) IsSignMinus() bool {
	return decimal128BIDIsSignMinusPort(d)
}

// Class returns the IEEE 754 class of d.
func (d Decimal128BID) Class() DecimalClass {
	return decimal128BIDClassPort(d)
}

// Sign returns 0 if d is zero (either sign), and otherwise -1 or +1 from the sign bit of d (including for NaNs).
func (d Decimal128BID) Sign() int { return decimal128BIDSignPort(d) }

// NewDecimal32BIDDirect parses s as a Decimal32BID value, reporting parse problems as an error (ParseDecimal32BIDRaw returns the exception flags instead).
func NewDecimal32BIDDirect(s string) (Decimal32BID, error) { return newDecimal32BIDDirectPort(s) }

// NewDecimal64BIDDirect parses s as a Decimal64BID value, reporting parse problems as an error (ParseDecimal64BIDRaw returns the exception flags instead).
func NewDecimal64BIDDirect(s string) (Decimal64BID, error) { return newDecimal64BIDDirectPort(s) }

// NewDecimal128BIDDirect parses s as a Decimal128BID value, reporting parse problems as an error (ParseDecimal128BIDRaw returns the exception flags instead).
func NewDecimal128BIDDirect(s string) (Decimal128BID, error) {
	return newDecimal128BIDDirectPort(s)
}

// ParseDecimal32BIDRaw parses s as a Decimal32BID value and returns the exception flags raised by the operation.
func ParseDecimal32BIDRaw(s string) (Decimal32BID, ExceptionFlags) {
	return parseDecimal32BIDPort(s)
}

// ParseDecimal64BIDRaw parses s as a Decimal64BID value and returns the exception flags raised by the operation.
func ParseDecimal64BIDRaw(s string) (Decimal64BID, ExceptionFlags) {
	return parseDecimal64BIDPort(s)
}

// ParseDecimal128BIDRaw parses s as a Decimal128BID value and returns the exception flags raised by the operation.
func ParseDecimal128BIDRaw(s string) (Decimal128BID, ExceptionFlags) {
	return parseDecimal128BIDPort(s)
}
