package bid754

// ConvertToInt8 converts d to int8 with the requested IEEE rounding mode.
// Passing a RoundingMode outside the defined constants panics.
func (d Decimal32BID) ConvertToInt8(mode RoundingMode) (int8, ExceptionFlags) {
	return decimal32BIDConvertToInt8Port(d, mode)
}

// ConvertToInt16 converts d to int16 with the requested IEEE rounding mode.
// Passing a RoundingMode outside the defined constants panics.
func (d Decimal32BID) ConvertToInt16(mode RoundingMode) (int16, ExceptionFlags) {
	return decimal32BIDConvertToInt16Port(d, mode)
}

// ConvertToInt32 converts d to int32 with the requested IEEE rounding mode.
// Passing a RoundingMode outside the defined constants panics.
func (d Decimal32BID) ConvertToInt32(mode RoundingMode) (int32, ExceptionFlags) {
	return decimal32BIDConvertToInt32Port(d, mode)
}

// ConvertToInt64 converts d to int64 with the requested IEEE rounding mode.
// Passing a RoundingMode outside the defined constants panics.
func (d Decimal32BID) ConvertToInt64(mode RoundingMode) (int64, ExceptionFlags) {
	return decimal32BIDConvertToInt64Port(d, mode)
}

// ConvertToUint8 converts d to uint8 with the requested IEEE rounding mode.
// Passing a RoundingMode outside the defined constants panics.
func (d Decimal32BID) ConvertToUint8(mode RoundingMode) (uint8, ExceptionFlags) {
	return decimal32BIDConvertToUint8Port(d, mode)
}

// ConvertToUint16 converts d to uint16 with the requested IEEE rounding mode.
// Passing a RoundingMode outside the defined constants panics.
func (d Decimal32BID) ConvertToUint16(mode RoundingMode) (uint16, ExceptionFlags) {
	return decimal32BIDConvertToUint16Port(d, mode)
}

// ConvertToUint32 converts d to uint32 with the requested IEEE rounding mode.
// Passing a RoundingMode outside the defined constants panics.
func (d Decimal32BID) ConvertToUint32(mode RoundingMode) (uint32, ExceptionFlags) {
	return decimal32BIDConvertToUint32Port(d, mode)
}

// ConvertToUint64 converts d to uint64 with the requested IEEE rounding mode.
// Passing a RoundingMode outside the defined constants panics.
func (d Decimal32BID) ConvertToUint64(mode RoundingMode) (uint64, ExceptionFlags) {
	return decimal32BIDConvertToUint64Port(d, mode)
}

// ConvertToInt8Exact converts d to int8 and signals FlagInexact when the integer result is not exact.
// Passing a RoundingMode outside the defined constants panics.
func (d Decimal32BID) ConvertToInt8Exact(mode RoundingMode) (int8, ExceptionFlags) {
	return decimal32BIDConvertToInt8ExactPort(d, mode)
}

// ConvertToInt16Exact converts d to int16 and signals FlagInexact when the integer result is not exact.
// Passing a RoundingMode outside the defined constants panics.
func (d Decimal32BID) ConvertToInt16Exact(mode RoundingMode) (int16, ExceptionFlags) {
	return decimal32BIDConvertToInt16ExactPort(d, mode)
}

// ConvertToInt32Exact converts d to int32 and signals FlagInexact when the integer result is not exact.
// Passing a RoundingMode outside the defined constants panics.
func (d Decimal32BID) ConvertToInt32Exact(mode RoundingMode) (int32, ExceptionFlags) {
	return decimal32BIDConvertToInt32ExactPort(d, mode)
}

// ConvertToInt64Exact converts d to int64 and signals FlagInexact when the integer result is not exact.
// Passing a RoundingMode outside the defined constants panics.
func (d Decimal32BID) ConvertToInt64Exact(mode RoundingMode) (int64, ExceptionFlags) {
	return decimal32BIDConvertToInt64ExactPort(d, mode)
}

// ConvertToUint8Exact converts d to uint8 and signals FlagInexact when the integer result is not exact.
// Passing a RoundingMode outside the defined constants panics.
func (d Decimal32BID) ConvertToUint8Exact(mode RoundingMode) (uint8, ExceptionFlags) {
	return decimal32BIDConvertToUint8ExactPort(d, mode)
}

// ConvertToUint16Exact converts d to uint16 and signals FlagInexact when the integer result is not exact.
// Passing a RoundingMode outside the defined constants panics.
func (d Decimal32BID) ConvertToUint16Exact(mode RoundingMode) (uint16, ExceptionFlags) {
	return decimal32BIDConvertToUint16ExactPort(d, mode)
}

// ConvertToUint32Exact converts d to uint32 and signals FlagInexact when the integer result is not exact.
// Passing a RoundingMode outside the defined constants panics.
func (d Decimal32BID) ConvertToUint32Exact(mode RoundingMode) (uint32, ExceptionFlags) {
	return decimal32BIDConvertToUint32ExactPort(d, mode)
}

// ConvertToUint64Exact converts d to uint64 and signals FlagInexact when the integer result is not exact.
// Passing a RoundingMode outside the defined constants panics.
func (d Decimal32BID) ConvertToUint64Exact(mode RoundingMode) (uint64, ExceptionFlags) {
	return decimal32BIDConvertToUint64ExactPort(d, mode)
}

// ConvertToInt8 converts d to int8 with the requested IEEE rounding mode.
// Passing a RoundingMode outside the defined constants panics.
func (d Decimal64BID) ConvertToInt8(mode RoundingMode) (int8, ExceptionFlags) {
	return decimal64BIDConvertToInt8Port(d, mode)
}

// ConvertToInt16 converts d to int16 with the requested IEEE rounding mode.
// Passing a RoundingMode outside the defined constants panics.
func (d Decimal64BID) ConvertToInt16(mode RoundingMode) (int16, ExceptionFlags) {
	return decimal64BIDConvertToInt16Port(d, mode)
}

// ConvertToInt32 converts d to int32 with the requested IEEE rounding mode.
// Passing a RoundingMode outside the defined constants panics.
func (d Decimal64BID) ConvertToInt32(mode RoundingMode) (int32, ExceptionFlags) {
	return decimal64BIDConvertToInt32Port(d, mode)
}

// ConvertToInt64 converts d to int64 with the requested IEEE rounding mode.
// Passing a RoundingMode outside the defined constants panics.
func (d Decimal64BID) ConvertToInt64(mode RoundingMode) (int64, ExceptionFlags) {
	return decimal64BIDConvertToInt64Port(d, mode)
}

// ConvertToUint8 converts d to uint8 with the requested IEEE rounding mode.
// Passing a RoundingMode outside the defined constants panics.
func (d Decimal64BID) ConvertToUint8(mode RoundingMode) (uint8, ExceptionFlags) {
	return decimal64BIDConvertToUint8Port(d, mode)
}

// ConvertToUint16 converts d to uint16 with the requested IEEE rounding mode.
// Passing a RoundingMode outside the defined constants panics.
func (d Decimal64BID) ConvertToUint16(mode RoundingMode) (uint16, ExceptionFlags) {
	return decimal64BIDConvertToUint16Port(d, mode)
}

// ConvertToUint32 converts d to uint32 with the requested IEEE rounding mode.
// Passing a RoundingMode outside the defined constants panics.
func (d Decimal64BID) ConvertToUint32(mode RoundingMode) (uint32, ExceptionFlags) {
	return decimal64BIDConvertToUint32Port(d, mode)
}

// ConvertToUint64 converts d to uint64 with the requested IEEE rounding mode.
// Passing a RoundingMode outside the defined constants panics.
func (d Decimal64BID) ConvertToUint64(mode RoundingMode) (uint64, ExceptionFlags) {
	return decimal64BIDConvertToUint64Port(d, mode)
}

// ConvertToInt8Exact converts d to int8 and signals FlagInexact when the integer result is not exact.
// Passing a RoundingMode outside the defined constants panics.
func (d Decimal64BID) ConvertToInt8Exact(mode RoundingMode) (int8, ExceptionFlags) {
	return decimal64BIDConvertToInt8ExactPort(d, mode)
}

// ConvertToInt16Exact converts d to int16 and signals FlagInexact when the integer result is not exact.
// Passing a RoundingMode outside the defined constants panics.
func (d Decimal64BID) ConvertToInt16Exact(mode RoundingMode) (int16, ExceptionFlags) {
	return decimal64BIDConvertToInt16ExactPort(d, mode)
}

// ConvertToInt32Exact converts d to int32 and signals FlagInexact when the integer result is not exact.
// Passing a RoundingMode outside the defined constants panics.
func (d Decimal64BID) ConvertToInt32Exact(mode RoundingMode) (int32, ExceptionFlags) {
	return decimal64BIDConvertToInt32ExactPort(d, mode)
}

// ConvertToInt64Exact converts d to int64 and signals FlagInexact when the integer result is not exact.
// Passing a RoundingMode outside the defined constants panics.
func (d Decimal64BID) ConvertToInt64Exact(mode RoundingMode) (int64, ExceptionFlags) {
	return decimal64BIDConvertToInt64ExactPort(d, mode)
}

// ConvertToUint8Exact converts d to uint8 and signals FlagInexact when the integer result is not exact.
// Passing a RoundingMode outside the defined constants panics.
func (d Decimal64BID) ConvertToUint8Exact(mode RoundingMode) (uint8, ExceptionFlags) {
	return decimal64BIDConvertToUint8ExactPort(d, mode)
}

// ConvertToUint16Exact converts d to uint16 and signals FlagInexact when the integer result is not exact.
// Passing a RoundingMode outside the defined constants panics.
func (d Decimal64BID) ConvertToUint16Exact(mode RoundingMode) (uint16, ExceptionFlags) {
	return decimal64BIDConvertToUint16ExactPort(d, mode)
}

// ConvertToUint32Exact converts d to uint32 and signals FlagInexact when the integer result is not exact.
// Passing a RoundingMode outside the defined constants panics.
func (d Decimal64BID) ConvertToUint32Exact(mode RoundingMode) (uint32, ExceptionFlags) {
	return decimal64BIDConvertToUint32ExactPort(d, mode)
}

// ConvertToUint64Exact converts d to uint64 and signals FlagInexact when the integer result is not exact.
// Passing a RoundingMode outside the defined constants panics.
func (d Decimal64BID) ConvertToUint64Exact(mode RoundingMode) (uint64, ExceptionFlags) {
	return decimal64BIDConvertToUint64ExactPort(d, mode)
}

// ConvertToInt8 converts d to int8 with the requested IEEE rounding mode.
// Passing a RoundingMode outside the defined constants panics.
func (d Decimal128BID) ConvertToInt8(mode RoundingMode) (int8, ExceptionFlags) {
	return decimal128BIDConvertToInt8Port(d, mode)
}

// ConvertToInt16 converts d to int16 with the requested IEEE rounding mode.
// Passing a RoundingMode outside the defined constants panics.
func (d Decimal128BID) ConvertToInt16(mode RoundingMode) (int16, ExceptionFlags) {
	return decimal128BIDConvertToInt16Port(d, mode)
}

// ConvertToInt32 converts d to int32 with the requested IEEE rounding mode.
// Passing a RoundingMode outside the defined constants panics.
func (d Decimal128BID) ConvertToInt32(mode RoundingMode) (int32, ExceptionFlags) {
	return decimal128BIDConvertToInt32Port(d, mode)
}

// ConvertToInt64 converts d to int64 with the requested IEEE rounding mode.
// Passing a RoundingMode outside the defined constants panics.
func (d Decimal128BID) ConvertToInt64(mode RoundingMode) (int64, ExceptionFlags) {
	return decimal128BIDConvertToInt64Port(d, mode)
}

// ConvertToUint8 converts d to uint8 with the requested IEEE rounding mode.
// Passing a RoundingMode outside the defined constants panics.
func (d Decimal128BID) ConvertToUint8(mode RoundingMode) (uint8, ExceptionFlags) {
	return decimal128BIDConvertToUint8Port(d, mode)
}

// ConvertToUint16 converts d to uint16 with the requested IEEE rounding mode.
// Passing a RoundingMode outside the defined constants panics.
func (d Decimal128BID) ConvertToUint16(mode RoundingMode) (uint16, ExceptionFlags) {
	return decimal128BIDConvertToUint16Port(d, mode)
}

// ConvertToUint32 converts d to uint32 with the requested IEEE rounding mode.
// Passing a RoundingMode outside the defined constants panics.
func (d Decimal128BID) ConvertToUint32(mode RoundingMode) (uint32, ExceptionFlags) {
	return decimal128BIDConvertToUint32Port(d, mode)
}

// ConvertToUint64 converts d to uint64 with the requested IEEE rounding mode.
// Passing a RoundingMode outside the defined constants panics.
func (d Decimal128BID) ConvertToUint64(mode RoundingMode) (uint64, ExceptionFlags) {
	return decimal128BIDConvertToUint64Port(d, mode)
}

// ConvertToInt8Exact converts d to int8 and signals FlagInexact when the integer result is not exact.
// Passing a RoundingMode outside the defined constants panics.
func (d Decimal128BID) ConvertToInt8Exact(mode RoundingMode) (int8, ExceptionFlags) {
	return decimal128BIDConvertToInt8ExactPort(d, mode)
}

// ConvertToInt16Exact converts d to int16 and signals FlagInexact when the integer result is not exact.
// Passing a RoundingMode outside the defined constants panics.
func (d Decimal128BID) ConvertToInt16Exact(mode RoundingMode) (int16, ExceptionFlags) {
	return decimal128BIDConvertToInt16ExactPort(d, mode)
}

// ConvertToInt32Exact converts d to int32 and signals FlagInexact when the integer result is not exact.
// Passing a RoundingMode outside the defined constants panics.
func (d Decimal128BID) ConvertToInt32Exact(mode RoundingMode) (int32, ExceptionFlags) {
	return decimal128BIDConvertToInt32ExactPort(d, mode)
}

// ConvertToInt64Exact converts d to int64 and signals FlagInexact when the integer result is not exact.
// Passing a RoundingMode outside the defined constants panics.
func (d Decimal128BID) ConvertToInt64Exact(mode RoundingMode) (int64, ExceptionFlags) {
	return decimal128BIDConvertToInt64ExactPort(d, mode)
}

// ConvertToUint8Exact converts d to uint8 and signals FlagInexact when the integer result is not exact.
// Passing a RoundingMode outside the defined constants panics.
func (d Decimal128BID) ConvertToUint8Exact(mode RoundingMode) (uint8, ExceptionFlags) {
	return decimal128BIDConvertToUint8ExactPort(d, mode)
}

// ConvertToUint16Exact converts d to uint16 and signals FlagInexact when the integer result is not exact.
// Passing a RoundingMode outside the defined constants panics.
func (d Decimal128BID) ConvertToUint16Exact(mode RoundingMode) (uint16, ExceptionFlags) {
	return decimal128BIDConvertToUint16ExactPort(d, mode)
}

// ConvertToUint32Exact converts d to uint32 and signals FlagInexact when the integer result is not exact.
// Passing a RoundingMode outside the defined constants panics.
func (d Decimal128BID) ConvertToUint32Exact(mode RoundingMode) (uint32, ExceptionFlags) {
	return decimal128BIDConvertToUint32ExactPort(d, mode)
}

// ConvertToUint64Exact converts d to uint64 and signals FlagInexact when the integer result is not exact.
// Passing a RoundingMode outside the defined constants panics.
func (d Decimal128BID) ConvertToUint64Exact(mode RoundingMode) (uint64, ExceptionFlags) {
	return decimal128BIDConvertToUint64ExactPort(d, mode)
}
