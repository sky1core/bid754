package bid754

import bidgo "github.com/sky1core/bid754/bid-go"

func decimal32BIDConvertToInt8Port(d Decimal32BID, mode RoundingMode) (int8, ExceptionFlags) {
	var result int8
	var flags uint32
	switch mode {
	case RoundNearestEven:
		result, flags = bidgo.Bid32ToInt8Rnint(d.ToUint32())
	case RoundNearestAway:
		result, flags = bidgo.Bid32ToInt8Rninta(d.ToUint32())
	case RoundTowardZero:
		result, flags = bidgo.Bid32ToInt8Int(d.ToUint32())
	case RoundTowardPositive:
		result, flags = bidgo.Bid32ToInt8Ceil(d.ToUint32())
	case RoundTowardNegative:
		result, flags = bidgo.Bid32ToInt8Floor(d.ToUint32())
	default:
		result, flags = bidgo.Bid32ToInt8Rnint(d.ToUint32())
	}
	return result, bidgoExceptionFlags(flags)
}

func decimal32BIDConvertToInt8ExactPort(d Decimal32BID, mode RoundingMode) (int8, ExceptionFlags) {
	var result int8
	var flags uint32
	switch mode {
	case RoundNearestEven:
		result, flags = bidgo.Bid32ToInt8Xrnint(d.ToUint32())
	case RoundNearestAway:
		result, flags = bidgo.Bid32ToInt8Xrninta(d.ToUint32())
	case RoundTowardZero:
		result, flags = bidgo.Bid32ToInt8Xint(d.ToUint32())
	case RoundTowardPositive:
		result, flags = bidgo.Bid32ToInt8Xceil(d.ToUint32())
	case RoundTowardNegative:
		result, flags = bidgo.Bid32ToInt8Xfloor(d.ToUint32())
	default:
		result, flags = bidgo.Bid32ToInt8Xrnint(d.ToUint32())
	}
	return result, bidgoExceptionFlags(flags)
}

func decimal32BIDConvertToInt16Port(d Decimal32BID, mode RoundingMode) (int16, ExceptionFlags) {
	var result int16
	var flags uint32
	switch mode {
	case RoundNearestEven:
		result, flags = bidgo.Bid32ToInt16Rnint(d.ToUint32())
	case RoundNearestAway:
		result, flags = bidgo.Bid32ToInt16Rninta(d.ToUint32())
	case RoundTowardZero:
		result, flags = bidgo.Bid32ToInt16Int(d.ToUint32())
	case RoundTowardPositive:
		result, flags = bidgo.Bid32ToInt16Ceil(d.ToUint32())
	case RoundTowardNegative:
		result, flags = bidgo.Bid32ToInt16Floor(d.ToUint32())
	default:
		result, flags = bidgo.Bid32ToInt16Rnint(d.ToUint32())
	}
	return result, bidgoExceptionFlags(flags)
}

func decimal32BIDConvertToInt16ExactPort(d Decimal32BID, mode RoundingMode) (int16, ExceptionFlags) {
	var result int16
	var flags uint32
	switch mode {
	case RoundNearestEven:
		result, flags = bidgo.Bid32ToInt16Xrnint(d.ToUint32())
	case RoundNearestAway:
		result, flags = bidgo.Bid32ToInt16Xrninta(d.ToUint32())
	case RoundTowardZero:
		result, flags = bidgo.Bid32ToInt16Xint(d.ToUint32())
	case RoundTowardPositive:
		result, flags = bidgo.Bid32ToInt16Xceil(d.ToUint32())
	case RoundTowardNegative:
		result, flags = bidgo.Bid32ToInt16Xfloor(d.ToUint32())
	default:
		result, flags = bidgo.Bid32ToInt16Xrnint(d.ToUint32())
	}
	return result, bidgoExceptionFlags(flags)
}

func decimal32BIDConvertToInt32Port(d Decimal32BID, mode RoundingMode) (int32, ExceptionFlags) {
	var result int32
	var flags uint32
	switch mode {
	case RoundNearestEven:
		result, flags = bidgo.Bid32ToInt32Rnint(d.ToUint32())
	case RoundNearestAway:
		result, flags = bidgo.Bid32ToInt32Rninta(d.ToUint32())
	case RoundTowardZero:
		result, flags = bidgo.Bid32ToInt32Int(d.ToUint32())
	case RoundTowardPositive:
		result, flags = bidgo.Bid32ToInt32Ceil(d.ToUint32())
	case RoundTowardNegative:
		result, flags = bidgo.Bid32ToInt32Floor(d.ToUint32())
	default:
		result, flags = bidgo.Bid32ToInt32Rnint(d.ToUint32())
	}
	return result, bidgoExceptionFlags(flags)
}

func decimal32BIDConvertToInt32ExactPort(d Decimal32BID, mode RoundingMode) (int32, ExceptionFlags) {
	var result int32
	var flags uint32
	switch mode {
	case RoundNearestEven:
		result, flags = bidgo.Bid32ToInt32Xrnint(d.ToUint32())
	case RoundNearestAway:
		result, flags = bidgo.Bid32ToInt32Xrninta(d.ToUint32())
	case RoundTowardZero:
		result, flags = bidgo.Bid32ToInt32Xint(d.ToUint32())
	case RoundTowardPositive:
		result, flags = bidgo.Bid32ToInt32Xceil(d.ToUint32())
	case RoundTowardNegative:
		result, flags = bidgo.Bid32ToInt32Xfloor(d.ToUint32())
	default:
		result, flags = bidgo.Bid32ToInt32Xrnint(d.ToUint32())
	}
	return result, bidgoExceptionFlags(flags)
}

func decimal32BIDConvertToInt64Port(d Decimal32BID, mode RoundingMode) (int64, ExceptionFlags) {
	var result int64
	var flags uint32
	switch mode {
	case RoundNearestEven:
		result, flags = bidgo.Bid32ToInt64Rnint(d.ToUint32())
	case RoundNearestAway:
		result, flags = bidgo.Bid32ToInt64Rninta(d.ToUint32())
	case RoundTowardZero:
		result, flags = bidgo.Bid32ToInt64Int(d.ToUint32())
	case RoundTowardPositive:
		result, flags = bidgo.Bid32ToInt64Ceil(d.ToUint32())
	case RoundTowardNegative:
		result, flags = bidgo.Bid32ToInt64Floor(d.ToUint32())
	default:
		result, flags = bidgo.Bid32ToInt64Rnint(d.ToUint32())
	}
	return result, bidgoExceptionFlags(flags)
}

func decimal32BIDConvertToInt64ExactPort(d Decimal32BID, mode RoundingMode) (int64, ExceptionFlags) {
	var result int64
	var flags uint32
	switch mode {
	case RoundNearestEven:
		result, flags = bidgo.Bid32ToInt64Xrnint(d.ToUint32())
	case RoundNearestAway:
		result, flags = bidgo.Bid32ToInt64Xrninta(d.ToUint32())
	case RoundTowardZero:
		result, flags = bidgo.Bid32ToInt64Xint(d.ToUint32())
	case RoundTowardPositive:
		result, flags = bidgo.Bid32ToInt64Xceil(d.ToUint32())
	case RoundTowardNegative:
		result, flags = bidgo.Bid32ToInt64Xfloor(d.ToUint32())
	default:
		result, flags = bidgo.Bid32ToInt64Xrnint(d.ToUint32())
	}
	return result, bidgoExceptionFlags(flags)
}

func decimal32BIDConvertToUint8Port(d Decimal32BID, mode RoundingMode) (uint8, ExceptionFlags) {
	var result uint8
	var flags uint32
	switch mode {
	case RoundNearestEven:
		result, flags = bidgo.Bid32ToUint8Rnint(d.ToUint32())
	case RoundNearestAway:
		result, flags = bidgo.Bid32ToUint8Rninta(d.ToUint32())
	case RoundTowardZero:
		result, flags = bidgo.Bid32ToUint8Int(d.ToUint32())
	case RoundTowardPositive:
		result, flags = bidgo.Bid32ToUint8Ceil(d.ToUint32())
	case RoundTowardNegative:
		result, flags = bidgo.Bid32ToUint8Floor(d.ToUint32())
	default:
		result, flags = bidgo.Bid32ToUint8Rnint(d.ToUint32())
	}
	return result, bidgoExceptionFlags(flags)
}

func decimal32BIDConvertToUint8ExactPort(d Decimal32BID, mode RoundingMode) (uint8, ExceptionFlags) {
	var result uint8
	var flags uint32
	switch mode {
	case RoundNearestEven:
		result, flags = bidgo.Bid32ToUint8Xrnint(d.ToUint32())
	case RoundNearestAway:
		result, flags = bidgo.Bid32ToUint8Xrninta(d.ToUint32())
	case RoundTowardZero:
		result, flags = bidgo.Bid32ToUint8Xint(d.ToUint32())
	case RoundTowardPositive:
		result, flags = bidgo.Bid32ToUint8Xceil(d.ToUint32())
	case RoundTowardNegative:
		result, flags = bidgo.Bid32ToUint8Xfloor(d.ToUint32())
	default:
		result, flags = bidgo.Bid32ToUint8Xrnint(d.ToUint32())
	}
	return result, bidgoExceptionFlags(flags)
}

func decimal32BIDConvertToUint16Port(d Decimal32BID, mode RoundingMode) (uint16, ExceptionFlags) {
	var result uint16
	var flags uint32
	switch mode {
	case RoundNearestEven:
		result, flags = bidgo.Bid32ToUint16Rnint(d.ToUint32())
	case RoundNearestAway:
		result, flags = bidgo.Bid32ToUint16Rninta(d.ToUint32())
	case RoundTowardZero:
		result, flags = bidgo.Bid32ToUint16Int(d.ToUint32())
	case RoundTowardPositive:
		result, flags = bidgo.Bid32ToUint16Ceil(d.ToUint32())
	case RoundTowardNegative:
		result, flags = bidgo.Bid32ToUint16Floor(d.ToUint32())
	default:
		result, flags = bidgo.Bid32ToUint16Rnint(d.ToUint32())
	}
	return result, bidgoExceptionFlags(flags)
}

func decimal32BIDConvertToUint16ExactPort(d Decimal32BID, mode RoundingMode) (uint16, ExceptionFlags) {
	var result uint16
	var flags uint32
	switch mode {
	case RoundNearestEven:
		result, flags = bidgo.Bid32ToUint16Xrnint(d.ToUint32())
	case RoundNearestAway:
		result, flags = bidgo.Bid32ToUint16Xrninta(d.ToUint32())
	case RoundTowardZero:
		result, flags = bidgo.Bid32ToUint16Xint(d.ToUint32())
	case RoundTowardPositive:
		result, flags = bidgo.Bid32ToUint16Xceil(d.ToUint32())
	case RoundTowardNegative:
		result, flags = bidgo.Bid32ToUint16Xfloor(d.ToUint32())
	default:
		result, flags = bidgo.Bid32ToUint16Xrnint(d.ToUint32())
	}
	return result, bidgoExceptionFlags(flags)
}

func decimal32BIDConvertToUint32Port(d Decimal32BID, mode RoundingMode) (uint32, ExceptionFlags) {
	var result uint32
	var flags uint32
	switch mode {
	case RoundNearestEven:
		result, flags = bidgo.Bid32ToUint32Rnint(d.ToUint32())
	case RoundNearestAway:
		result, flags = bidgo.Bid32ToUint32Rninta(d.ToUint32())
	case RoundTowardZero:
		result, flags = bidgo.Bid32ToUint32Int(d.ToUint32())
	case RoundTowardPositive:
		result, flags = bidgo.Bid32ToUint32Ceil(d.ToUint32())
	case RoundTowardNegative:
		result, flags = bidgo.Bid32ToUint32Floor(d.ToUint32())
	default:
		result, flags = bidgo.Bid32ToUint32Rnint(d.ToUint32())
	}
	return result, bidgoExceptionFlags(flags)
}

func decimal32BIDConvertToUint32ExactPort(d Decimal32BID, mode RoundingMode) (uint32, ExceptionFlags) {
	var result uint32
	var flags uint32
	switch mode {
	case RoundNearestEven:
		result, flags = bidgo.Bid32ToUint32Xrnint(d.ToUint32())
	case RoundNearestAway:
		result, flags = bidgo.Bid32ToUint32Xrninta(d.ToUint32())
	case RoundTowardZero:
		result, flags = bidgo.Bid32ToUint32Xint(d.ToUint32())
	case RoundTowardPositive:
		result, flags = bidgo.Bid32ToUint32Xceil(d.ToUint32())
	case RoundTowardNegative:
		result, flags = bidgo.Bid32ToUint32Xfloor(d.ToUint32())
	default:
		result, flags = bidgo.Bid32ToUint32Xrnint(d.ToUint32())
	}
	return result, bidgoExceptionFlags(flags)
}

func decimal32BIDConvertToUint64Port(d Decimal32BID, mode RoundingMode) (uint64, ExceptionFlags) {
	var result uint64
	var flags uint32
	switch mode {
	case RoundNearestEven:
		result, flags = bidgo.Bid32ToUint64Rnint(d.ToUint32())
	case RoundNearestAway:
		result, flags = bidgo.Bid32ToUint64Rninta(d.ToUint32())
	case RoundTowardZero:
		result, flags = bidgo.Bid32ToUint64Int(d.ToUint32())
	case RoundTowardPositive:
		result, flags = bidgo.Bid32ToUint64Ceil(d.ToUint32())
	case RoundTowardNegative:
		result, flags = bidgo.Bid32ToUint64Floor(d.ToUint32())
	default:
		result, flags = bidgo.Bid32ToUint64Rnint(d.ToUint32())
	}
	return result, bidgoExceptionFlags(flags)
}

func decimal32BIDConvertToUint64ExactPort(d Decimal32BID, mode RoundingMode) (uint64, ExceptionFlags) {
	var result uint64
	var flags uint32
	switch mode {
	case RoundNearestEven:
		result, flags = bidgo.Bid32ToUint64Xrnint(d.ToUint32())
	case RoundNearestAway:
		result, flags = bidgo.Bid32ToUint64Xrninta(d.ToUint32())
	case RoundTowardZero:
		result, flags = bidgo.Bid32ToUint64Xint(d.ToUint32())
	case RoundTowardPositive:
		result, flags = bidgo.Bid32ToUint64Xceil(d.ToUint32())
	case RoundTowardNegative:
		result, flags = bidgo.Bid32ToUint64Xfloor(d.ToUint32())
	default:
		result, flags = bidgo.Bid32ToUint64Xrnint(d.ToUint32())
	}
	return result, bidgoExceptionFlags(flags)
}

func decimal64BIDConvertToInt8Port(d Decimal64BID, mode RoundingMode) (int8, ExceptionFlags) {
	var result int8
	var flags uint32
	switch mode {
	case RoundNearestEven:
		result, flags = bidgo.Bid64ToInt8Rnint(d.ToUint64())
	case RoundNearestAway:
		result, flags = bidgo.Bid64ToInt8Rninta(d.ToUint64())
	case RoundTowardZero:
		result, flags = bidgo.Bid64ToInt8Int(d.ToUint64())
	case RoundTowardPositive:
		result, flags = bidgo.Bid64ToInt8Ceil(d.ToUint64())
	case RoundTowardNegative:
		result, flags = bidgo.Bid64ToInt8Floor(d.ToUint64())
	default:
		result, flags = bidgo.Bid64ToInt8Rnint(d.ToUint64())
	}
	return result, bidgoExceptionFlags(flags)
}

func decimal64BIDConvertToInt8ExactPort(d Decimal64BID, mode RoundingMode) (int8, ExceptionFlags) {
	var result int8
	var flags uint32
	switch mode {
	case RoundNearestEven:
		result, flags = bidgo.Bid64ToInt8Xrnint(d.ToUint64())
	case RoundNearestAway:
		result, flags = bidgo.Bid64ToInt8Xrninta(d.ToUint64())
	case RoundTowardZero:
		result, flags = bidgo.Bid64ToInt8Xint(d.ToUint64())
	case RoundTowardPositive:
		result, flags = bidgo.Bid64ToInt8Xceil(d.ToUint64())
	case RoundTowardNegative:
		result, flags = bidgo.Bid64ToInt8Xfloor(d.ToUint64())
	default:
		result, flags = bidgo.Bid64ToInt8Xrnint(d.ToUint64())
	}
	return result, bidgoExceptionFlags(flags)
}

func decimal64BIDConvertToInt16Port(d Decimal64BID, mode RoundingMode) (int16, ExceptionFlags) {
	var result int16
	var flags uint32
	switch mode {
	case RoundNearestEven:
		result, flags = bidgo.Bid64ToInt16Rnint(d.ToUint64())
	case RoundNearestAway:
		result, flags = bidgo.Bid64ToInt16Rninta(d.ToUint64())
	case RoundTowardZero:
		result, flags = bidgo.Bid64ToInt16Int(d.ToUint64())
	case RoundTowardPositive:
		result, flags = bidgo.Bid64ToInt16Ceil(d.ToUint64())
	case RoundTowardNegative:
		result, flags = bidgo.Bid64ToInt16Floor(d.ToUint64())
	default:
		result, flags = bidgo.Bid64ToInt16Rnint(d.ToUint64())
	}
	return result, bidgoExceptionFlags(flags)
}

func decimal64BIDConvertToInt16ExactPort(d Decimal64BID, mode RoundingMode) (int16, ExceptionFlags) {
	var result int16
	var flags uint32
	switch mode {
	case RoundNearestEven:
		result, flags = bidgo.Bid64ToInt16Xrnint(d.ToUint64())
	case RoundNearestAway:
		result, flags = bidgo.Bid64ToInt16Xrninta(d.ToUint64())
	case RoundTowardZero:
		result, flags = bidgo.Bid64ToInt16Xint(d.ToUint64())
	case RoundTowardPositive:
		result, flags = bidgo.Bid64ToInt16Xceil(d.ToUint64())
	case RoundTowardNegative:
		result, flags = bidgo.Bid64ToInt16Xfloor(d.ToUint64())
	default:
		result, flags = bidgo.Bid64ToInt16Xrnint(d.ToUint64())
	}
	return result, bidgoExceptionFlags(flags)
}

func decimal64BIDConvertToInt32Port(d Decimal64BID, mode RoundingMode) (int32, ExceptionFlags) {
	var result int32
	var flags uint32
	switch mode {
	case RoundNearestEven:
		result, flags = bidgo.Bid64ToInt32Rnint(d.ToUint64())
	case RoundNearestAway:
		result, flags = bidgo.Bid64ToInt32Rninta(d.ToUint64())
	case RoundTowardZero:
		result, flags = bidgo.Bid64ToInt32Int(d.ToUint64())
	case RoundTowardPositive:
		result, flags = bidgo.Bid64ToInt32Ceil(d.ToUint64())
	case RoundTowardNegative:
		result, flags = bidgo.Bid64ToInt32Floor(d.ToUint64())
	default:
		result, flags = bidgo.Bid64ToInt32Rnint(d.ToUint64())
	}
	return result, bidgoExceptionFlags(flags)
}

func decimal64BIDConvertToInt32ExactPort(d Decimal64BID, mode RoundingMode) (int32, ExceptionFlags) {
	var result int32
	var flags uint32
	switch mode {
	case RoundNearestEven:
		result, flags = bidgo.Bid64ToInt32Xrnint(d.ToUint64())
	case RoundNearestAway:
		result, flags = bidgo.Bid64ToInt32Xrninta(d.ToUint64())
	case RoundTowardZero:
		result, flags = bidgo.Bid64ToInt32Xint(d.ToUint64())
	case RoundTowardPositive:
		result, flags = bidgo.Bid64ToInt32Xceil(d.ToUint64())
	case RoundTowardNegative:
		result, flags = bidgo.Bid64ToInt32Xfloor(d.ToUint64())
	default:
		result, flags = bidgo.Bid64ToInt32Xrnint(d.ToUint64())
	}
	return result, bidgoExceptionFlags(flags)
}

func decimal64BIDConvertToInt64Port(d Decimal64BID, mode RoundingMode) (int64, ExceptionFlags) {
	var result int64
	var flags uint32
	switch mode {
	case RoundNearestEven:
		result, flags = bidgo.Bid64ToInt64Rnint(d.ToUint64())
	case RoundNearestAway:
		result, flags = bidgo.Bid64ToInt64Rninta(d.ToUint64())
	case RoundTowardZero:
		result, flags = bidgo.Bid64ToInt64Int(d.ToUint64())
	case RoundTowardPositive:
		result, flags = bidgo.Bid64ToInt64Ceil(d.ToUint64())
	case RoundTowardNegative:
		result, flags = bidgo.Bid64ToInt64Floor(d.ToUint64())
	default:
		result, flags = bidgo.Bid64ToInt64Rnint(d.ToUint64())
	}
	return result, bidgoExceptionFlags(flags)
}

func decimal64BIDConvertToInt64ExactPort(d Decimal64BID, mode RoundingMode) (int64, ExceptionFlags) {
	var result int64
	var flags uint32
	switch mode {
	case RoundNearestEven:
		result, flags = bidgo.Bid64ToInt64Xrnint(d.ToUint64())
	case RoundNearestAway:
		result, flags = bidgo.Bid64ToInt64Xrninta(d.ToUint64())
	case RoundTowardZero:
		result, flags = bidgo.Bid64ToInt64Xint(d.ToUint64())
	case RoundTowardPositive:
		result, flags = bidgo.Bid64ToInt64Xceil(d.ToUint64())
	case RoundTowardNegative:
		result, flags = bidgo.Bid64ToInt64Xfloor(d.ToUint64())
	default:
		result, flags = bidgo.Bid64ToInt64Xrnint(d.ToUint64())
	}
	return result, bidgoExceptionFlags(flags)
}

func decimal64BIDConvertToUint8Port(d Decimal64BID, mode RoundingMode) (uint8, ExceptionFlags) {
	var result uint8
	var flags uint32
	switch mode {
	case RoundNearestEven:
		result, flags = bidgo.Bid64ToUint8Rnint(d.ToUint64())
	case RoundNearestAway:
		result, flags = bidgo.Bid64ToUint8Rninta(d.ToUint64())
	case RoundTowardZero:
		result, flags = bidgo.Bid64ToUint8Int(d.ToUint64())
	case RoundTowardPositive:
		result, flags = bidgo.Bid64ToUint8Ceil(d.ToUint64())
	case RoundTowardNegative:
		result, flags = bidgo.Bid64ToUint8Floor(d.ToUint64())
	default:
		result, flags = bidgo.Bid64ToUint8Rnint(d.ToUint64())
	}
	return result, bidgoExceptionFlags(flags)
}

func decimal64BIDConvertToUint8ExactPort(d Decimal64BID, mode RoundingMode) (uint8, ExceptionFlags) {
	var result uint8
	var flags uint32
	switch mode {
	case RoundNearestEven:
		result, flags = bidgo.Bid64ToUint8Xrnint(d.ToUint64())
	case RoundNearestAway:
		result, flags = bidgo.Bid64ToUint8Xrninta(d.ToUint64())
	case RoundTowardZero:
		result, flags = bidgo.Bid64ToUint8Xint(d.ToUint64())
	case RoundTowardPositive:
		result, flags = bidgo.Bid64ToUint8Xceil(d.ToUint64())
	case RoundTowardNegative:
		result, flags = bidgo.Bid64ToUint8Xfloor(d.ToUint64())
	default:
		result, flags = bidgo.Bid64ToUint8Xrnint(d.ToUint64())
	}
	return result, bidgoExceptionFlags(flags)
}

func decimal64BIDConvertToUint16Port(d Decimal64BID, mode RoundingMode) (uint16, ExceptionFlags) {
	var result uint16
	var flags uint32
	switch mode {
	case RoundNearestEven:
		result, flags = bidgo.Bid64ToUint16Rnint(d.ToUint64())
	case RoundNearestAway:
		result, flags = bidgo.Bid64ToUint16Rninta(d.ToUint64())
	case RoundTowardZero:
		result, flags = bidgo.Bid64ToUint16Int(d.ToUint64())
	case RoundTowardPositive:
		result, flags = bidgo.Bid64ToUint16Ceil(d.ToUint64())
	case RoundTowardNegative:
		result, flags = bidgo.Bid64ToUint16Floor(d.ToUint64())
	default:
		result, flags = bidgo.Bid64ToUint16Rnint(d.ToUint64())
	}
	return result, bidgoExceptionFlags(flags)
}

func decimal64BIDConvertToUint16ExactPort(d Decimal64BID, mode RoundingMode) (uint16, ExceptionFlags) {
	var result uint16
	var flags uint32
	switch mode {
	case RoundNearestEven:
		result, flags = bidgo.Bid64ToUint16Xrnint(d.ToUint64())
	case RoundNearestAway:
		result, flags = bidgo.Bid64ToUint16Xrninta(d.ToUint64())
	case RoundTowardZero:
		result, flags = bidgo.Bid64ToUint16Xint(d.ToUint64())
	case RoundTowardPositive:
		result, flags = bidgo.Bid64ToUint16Xceil(d.ToUint64())
	case RoundTowardNegative:
		result, flags = bidgo.Bid64ToUint16Xfloor(d.ToUint64())
	default:
		result, flags = bidgo.Bid64ToUint16Xrnint(d.ToUint64())
	}
	return result, bidgoExceptionFlags(flags)
}

func decimal64BIDConvertToUint32Port(d Decimal64BID, mode RoundingMode) (uint32, ExceptionFlags) {
	var result uint32
	var flags uint32
	switch mode {
	case RoundNearestEven:
		result, flags = bidgo.Bid64ToUint32Rnint(d.ToUint64())
	case RoundNearestAway:
		result, flags = bidgo.Bid64ToUint32Rninta(d.ToUint64())
	case RoundTowardZero:
		result, flags = bidgo.Bid64ToUint32Int(d.ToUint64())
	case RoundTowardPositive:
		result, flags = bidgo.Bid64ToUint32Ceil(d.ToUint64())
	case RoundTowardNegative:
		result, flags = bidgo.Bid64ToUint32Floor(d.ToUint64())
	default:
		result, flags = bidgo.Bid64ToUint32Rnint(d.ToUint64())
	}
	return result, bidgoExceptionFlags(flags)
}

func decimal64BIDConvertToUint32ExactPort(d Decimal64BID, mode RoundingMode) (uint32, ExceptionFlags) {
	var result uint32
	var flags uint32
	switch mode {
	case RoundNearestEven:
		result, flags = bidgo.Bid64ToUint32Xrnint(d.ToUint64())
	case RoundNearestAway:
		result, flags = bidgo.Bid64ToUint32Xrninta(d.ToUint64())
	case RoundTowardZero:
		result, flags = bidgo.Bid64ToUint32Xint(d.ToUint64())
	case RoundTowardPositive:
		result, flags = bidgo.Bid64ToUint32Xceil(d.ToUint64())
	case RoundTowardNegative:
		result, flags = bidgo.Bid64ToUint32Xfloor(d.ToUint64())
	default:
		result, flags = bidgo.Bid64ToUint32Xrnint(d.ToUint64())
	}
	return result, bidgoExceptionFlags(flags)
}

func decimal64BIDConvertToUint64Port(d Decimal64BID, mode RoundingMode) (uint64, ExceptionFlags) {
	var result uint64
	var flags uint32
	switch mode {
	case RoundNearestEven:
		result, flags = bidgo.Bid64ToUint64Rnint(d.ToUint64())
	case RoundNearestAway:
		result, flags = bidgo.Bid64ToUint64Rninta(d.ToUint64())
	case RoundTowardZero:
		result, flags = bidgo.Bid64ToUint64Int(d.ToUint64())
	case RoundTowardPositive:
		result, flags = bidgo.Bid64ToUint64Ceil(d.ToUint64())
	case RoundTowardNegative:
		result, flags = bidgo.Bid64ToUint64Floor(d.ToUint64())
	default:
		result, flags = bidgo.Bid64ToUint64Rnint(d.ToUint64())
	}
	return result, bidgoExceptionFlags(flags)
}

func decimal64BIDConvertToUint64ExactPort(d Decimal64BID, mode RoundingMode) (uint64, ExceptionFlags) {
	var result uint64
	var flags uint32
	switch mode {
	case RoundNearestEven:
		result, flags = bidgo.Bid64ToUint64Xrnint(d.ToUint64())
	case RoundNearestAway:
		result, flags = bidgo.Bid64ToUint64Xrninta(d.ToUint64())
	case RoundTowardZero:
		result, flags = bidgo.Bid64ToUint64Xint(d.ToUint64())
	case RoundTowardPositive:
		result, flags = bidgo.Bid64ToUint64Xceil(d.ToUint64())
	case RoundTowardNegative:
		result, flags = bidgo.Bid64ToUint64Xfloor(d.ToUint64())
	default:
		result, flags = bidgo.Bid64ToUint64Xrnint(d.ToUint64())
	}
	return result, bidgoExceptionFlags(flags)
}

func decimal128BIDConvertToInt8Port(d Decimal128BID, mode RoundingMode) (int8, ExceptionFlags) {
	var result int8
	var flags uint32
	switch mode {
	case RoundNearestEven:
		result, flags = bidgo.Bid128ToInt8Rnint(decimal128BIDAsBidgo(d))
	case RoundNearestAway:
		result, flags = bidgo.Bid128ToInt8Rninta(decimal128BIDAsBidgo(d))
	case RoundTowardZero:
		result, flags = bidgo.Bid128ToInt8Int(decimal128BIDAsBidgo(d))
	case RoundTowardPositive:
		result, flags = bidgo.Bid128ToInt8Ceil(decimal128BIDAsBidgo(d))
	case RoundTowardNegative:
		result, flags = bidgo.Bid128ToInt8Floor(decimal128BIDAsBidgo(d))
	default:
		result, flags = bidgo.Bid128ToInt8Rnint(decimal128BIDAsBidgo(d))
	}
	return result, bidgoExceptionFlags(flags)
}

func decimal128BIDConvertToInt8ExactPort(d Decimal128BID, mode RoundingMode) (int8, ExceptionFlags) {
	var result int8
	var flags uint32
	switch mode {
	case RoundNearestEven:
		result, flags = bidgo.Bid128ToInt8Xrnint(decimal128BIDAsBidgo(d))
	case RoundNearestAway:
		result, flags = bidgo.Bid128ToInt8Xrninta(decimal128BIDAsBidgo(d))
	case RoundTowardZero:
		result, flags = bidgo.Bid128ToInt8Xint(decimal128BIDAsBidgo(d))
	case RoundTowardPositive:
		result, flags = bidgo.Bid128ToInt8Xceil(decimal128BIDAsBidgo(d))
	case RoundTowardNegative:
		result, flags = bidgo.Bid128ToInt8Xfloor(decimal128BIDAsBidgo(d))
	default:
		result, flags = bidgo.Bid128ToInt8Xrnint(decimal128BIDAsBidgo(d))
	}
	return result, bidgoExceptionFlags(flags)
}

func decimal128BIDConvertToInt16Port(d Decimal128BID, mode RoundingMode) (int16, ExceptionFlags) {
	var result int16
	var flags uint32
	switch mode {
	case RoundNearestEven:
		result, flags = bidgo.Bid128ToInt16Rnint(decimal128BIDAsBidgo(d))
	case RoundNearestAway:
		result, flags = bidgo.Bid128ToInt16Rninta(decimal128BIDAsBidgo(d))
	case RoundTowardZero:
		result, flags = bidgo.Bid128ToInt16Int(decimal128BIDAsBidgo(d))
	case RoundTowardPositive:
		result, flags = bidgo.Bid128ToInt16Ceil(decimal128BIDAsBidgo(d))
	case RoundTowardNegative:
		result, flags = bidgo.Bid128ToInt16Floor(decimal128BIDAsBidgo(d))
	default:
		result, flags = bidgo.Bid128ToInt16Rnint(decimal128BIDAsBidgo(d))
	}
	return result, bidgoExceptionFlags(flags)
}

func decimal128BIDConvertToInt16ExactPort(d Decimal128BID, mode RoundingMode) (int16, ExceptionFlags) {
	var result int16
	var flags uint32
	switch mode {
	case RoundNearestEven:
		result, flags = bidgo.Bid128ToInt16Xrnint(decimal128BIDAsBidgo(d))
	case RoundNearestAway:
		result, flags = bidgo.Bid128ToInt16Xrninta(decimal128BIDAsBidgo(d))
	case RoundTowardZero:
		result, flags = bidgo.Bid128ToInt16Xint(decimal128BIDAsBidgo(d))
	case RoundTowardPositive:
		result, flags = bidgo.Bid128ToInt16Xceil(decimal128BIDAsBidgo(d))
	case RoundTowardNegative:
		result, flags = bidgo.Bid128ToInt16Xfloor(decimal128BIDAsBidgo(d))
	default:
		result, flags = bidgo.Bid128ToInt16Xrnint(decimal128BIDAsBidgo(d))
	}
	return result, bidgoExceptionFlags(flags)
}

func decimal128BIDConvertToInt32Port(d Decimal128BID, mode RoundingMode) (int32, ExceptionFlags) {
	var result int32
	var flags uint32
	switch mode {
	case RoundNearestEven:
		result, flags = bidgo.Bid128ToInt32Rnint(decimal128BIDAsBidgo(d))
	case RoundNearestAway:
		result, flags = bidgo.Bid128ToInt32Rninta(decimal128BIDAsBidgo(d))
	case RoundTowardZero:
		result, flags = bidgo.Bid128ToInt32Int(decimal128BIDAsBidgo(d))
	case RoundTowardPositive:
		result, flags = bidgo.Bid128ToInt32Ceil(decimal128BIDAsBidgo(d))
	case RoundTowardNegative:
		result, flags = bidgo.Bid128ToInt32Floor(decimal128BIDAsBidgo(d))
	default:
		result, flags = bidgo.Bid128ToInt32Rnint(decimal128BIDAsBidgo(d))
	}
	return result, bidgoExceptionFlags(flags)
}

func decimal128BIDConvertToInt32ExactPort(d Decimal128BID, mode RoundingMode) (int32, ExceptionFlags) {
	var result int32
	var flags uint32
	switch mode {
	case RoundNearestEven:
		result, flags = bidgo.Bid128ToInt32Xrnint(decimal128BIDAsBidgo(d))
	case RoundNearestAway:
		result, flags = bidgo.Bid128ToInt32Xrninta(decimal128BIDAsBidgo(d))
	case RoundTowardZero:
		result, flags = bidgo.Bid128ToInt32Xint(decimal128BIDAsBidgo(d))
	case RoundTowardPositive:
		result, flags = bidgo.Bid128ToInt32Xceil(decimal128BIDAsBidgo(d))
	case RoundTowardNegative:
		result, flags = bidgo.Bid128ToInt32Xfloor(decimal128BIDAsBidgo(d))
	default:
		result, flags = bidgo.Bid128ToInt32Xrnint(decimal128BIDAsBidgo(d))
	}
	return result, bidgoExceptionFlags(flags)
}

func decimal128BIDConvertToInt64Port(d Decimal128BID, mode RoundingMode) (int64, ExceptionFlags) {
	var result int64
	var flags uint32
	switch mode {
	case RoundNearestEven:
		result, flags = bidgo.Bid128ToInt64Rnint(decimal128BIDAsBidgo(d))
	case RoundNearestAway:
		result, flags = bidgo.Bid128ToInt64Rninta(decimal128BIDAsBidgo(d))
	case RoundTowardZero:
		result, flags = bidgo.Bid128ToInt64Int(decimal128BIDAsBidgo(d))
	case RoundTowardPositive:
		result, flags = bidgo.Bid128ToInt64Ceil(decimal128BIDAsBidgo(d))
	case RoundTowardNegative:
		result, flags = bidgo.Bid128ToInt64Floor(decimal128BIDAsBidgo(d))
	default:
		result, flags = bidgo.Bid128ToInt64Rnint(decimal128BIDAsBidgo(d))
	}
	return result, bidgoExceptionFlags(flags)
}

func decimal128BIDConvertToInt64ExactPort(d Decimal128BID, mode RoundingMode) (int64, ExceptionFlags) {
	var result int64
	var flags uint32
	switch mode {
	case RoundNearestEven:
		result, flags = bidgo.Bid128ToInt64Xrnint(decimal128BIDAsBidgo(d))
	case RoundNearestAway:
		result, flags = bidgo.Bid128ToInt64Xrninta(decimal128BIDAsBidgo(d))
	case RoundTowardZero:
		result, flags = bidgo.Bid128ToInt64Xint(decimal128BIDAsBidgo(d))
	case RoundTowardPositive:
		result, flags = bidgo.Bid128ToInt64Xceil(decimal128BIDAsBidgo(d))
	case RoundTowardNegative:
		result, flags = bidgo.Bid128ToInt64Xfloor(decimal128BIDAsBidgo(d))
	default:
		result, flags = bidgo.Bid128ToInt64Xrnint(decimal128BIDAsBidgo(d))
	}
	return result, bidgoExceptionFlags(flags)
}

func decimal128BIDConvertToUint8Port(d Decimal128BID, mode RoundingMode) (uint8, ExceptionFlags) {
	var result uint8
	var flags uint32
	switch mode {
	case RoundNearestEven:
		result, flags = bidgo.Bid128ToUint8Rnint(decimal128BIDAsBidgo(d))
	case RoundNearestAway:
		result, flags = bidgo.Bid128ToUint8Rninta(decimal128BIDAsBidgo(d))
	case RoundTowardZero:
		result, flags = bidgo.Bid128ToUint8Int(decimal128BIDAsBidgo(d))
	case RoundTowardPositive:
		result, flags = bidgo.Bid128ToUint8Ceil(decimal128BIDAsBidgo(d))
	case RoundTowardNegative:
		result, flags = bidgo.Bid128ToUint8Floor(decimal128BIDAsBidgo(d))
	default:
		result, flags = bidgo.Bid128ToUint8Rnint(decimal128BIDAsBidgo(d))
	}
	return result, bidgoExceptionFlags(flags)
}

func decimal128BIDConvertToUint8ExactPort(d Decimal128BID, mode RoundingMode) (uint8, ExceptionFlags) {
	var result uint8
	var flags uint32
	switch mode {
	case RoundNearestEven:
		result, flags = bidgo.Bid128ToUint8Xrnint(decimal128BIDAsBidgo(d))
	case RoundNearestAway:
		result, flags = bidgo.Bid128ToUint8Xrninta(decimal128BIDAsBidgo(d))
	case RoundTowardZero:
		result, flags = bidgo.Bid128ToUint8Xint(decimal128BIDAsBidgo(d))
	case RoundTowardPositive:
		result, flags = bidgo.Bid128ToUint8Xceil(decimal128BIDAsBidgo(d))
	case RoundTowardNegative:
		result, flags = bidgo.Bid128ToUint8Xfloor(decimal128BIDAsBidgo(d))
	default:
		result, flags = bidgo.Bid128ToUint8Xrnint(decimal128BIDAsBidgo(d))
	}
	return result, bidgoExceptionFlags(flags)
}

func decimal128BIDConvertToUint16Port(d Decimal128BID, mode RoundingMode) (uint16, ExceptionFlags) {
	var result uint16
	var flags uint32
	switch mode {
	case RoundNearestEven:
		result, flags = bidgo.Bid128ToUint16Rnint(decimal128BIDAsBidgo(d))
	case RoundNearestAway:
		result, flags = bidgo.Bid128ToUint16Rninta(decimal128BIDAsBidgo(d))
	case RoundTowardZero:
		result, flags = bidgo.Bid128ToUint16Int(decimal128BIDAsBidgo(d))
	case RoundTowardPositive:
		result, flags = bidgo.Bid128ToUint16Ceil(decimal128BIDAsBidgo(d))
	case RoundTowardNegative:
		result, flags = bidgo.Bid128ToUint16Floor(decimal128BIDAsBidgo(d))
	default:
		result, flags = bidgo.Bid128ToUint16Rnint(decimal128BIDAsBidgo(d))
	}
	return result, bidgoExceptionFlags(flags)
}

func decimal128BIDConvertToUint16ExactPort(d Decimal128BID, mode RoundingMode) (uint16, ExceptionFlags) {
	var result uint16
	var flags uint32
	switch mode {
	case RoundNearestEven:
		result, flags = bidgo.Bid128ToUint16Xrnint(decimal128BIDAsBidgo(d))
	case RoundNearestAway:
		result, flags = bidgo.Bid128ToUint16Xrninta(decimal128BIDAsBidgo(d))
	case RoundTowardZero:
		result, flags = bidgo.Bid128ToUint16Xint(decimal128BIDAsBidgo(d))
	case RoundTowardPositive:
		result, flags = bidgo.Bid128ToUint16Xceil(decimal128BIDAsBidgo(d))
	case RoundTowardNegative:
		result, flags = bidgo.Bid128ToUint16Xfloor(decimal128BIDAsBidgo(d))
	default:
		result, flags = bidgo.Bid128ToUint16Xrnint(decimal128BIDAsBidgo(d))
	}
	return result, bidgoExceptionFlags(flags)
}

func decimal128BIDConvertToUint32Port(d Decimal128BID, mode RoundingMode) (uint32, ExceptionFlags) {
	var result uint32
	var flags uint32
	switch mode {
	case RoundNearestEven:
		result, flags = bidgo.Bid128ToUint32Rnint(decimal128BIDAsBidgo(d))
	case RoundNearestAway:
		result, flags = bidgo.Bid128ToUint32Rninta(decimal128BIDAsBidgo(d))
	case RoundTowardZero:
		result, flags = bidgo.Bid128ToUint32Int(decimal128BIDAsBidgo(d))
	case RoundTowardPositive:
		result, flags = bidgo.Bid128ToUint32Ceil(decimal128BIDAsBidgo(d))
	case RoundTowardNegative:
		result, flags = bidgo.Bid128ToUint32Floor(decimal128BIDAsBidgo(d))
	default:
		result, flags = bidgo.Bid128ToUint32Rnint(decimal128BIDAsBidgo(d))
	}
	return result, bidgoExceptionFlags(flags)
}

func decimal128BIDConvertToUint32ExactPort(d Decimal128BID, mode RoundingMode) (uint32, ExceptionFlags) {
	var result uint32
	var flags uint32
	switch mode {
	case RoundNearestEven:
		result, flags = bidgo.Bid128ToUint32Xrnint(decimal128BIDAsBidgo(d))
	case RoundNearestAway:
		result, flags = bidgo.Bid128ToUint32Xrninta(decimal128BIDAsBidgo(d))
	case RoundTowardZero:
		result, flags = bidgo.Bid128ToUint32Xint(decimal128BIDAsBidgo(d))
	case RoundTowardPositive:
		result, flags = bidgo.Bid128ToUint32Xceil(decimal128BIDAsBidgo(d))
	case RoundTowardNegative:
		result, flags = bidgo.Bid128ToUint32Xfloor(decimal128BIDAsBidgo(d))
	default:
		result, flags = bidgo.Bid128ToUint32Xrnint(decimal128BIDAsBidgo(d))
	}
	return result, bidgoExceptionFlags(flags)
}

func decimal128BIDConvertToUint64Port(d Decimal128BID, mode RoundingMode) (uint64, ExceptionFlags) {
	var result uint64
	var flags uint32
	switch mode {
	case RoundNearestEven:
		result, flags = bidgo.Bid128ToUint64Rnint(decimal128BIDAsBidgo(d))
	case RoundNearestAway:
		result, flags = bidgo.Bid128ToUint64Rninta(decimal128BIDAsBidgo(d))
	case RoundTowardZero:
		result, flags = bidgo.Bid128ToUint64Int(decimal128BIDAsBidgo(d))
	case RoundTowardPositive:
		result, flags = bidgo.Bid128ToUint64Ceil(decimal128BIDAsBidgo(d))
	case RoundTowardNegative:
		result, flags = bidgo.Bid128ToUint64Floor(decimal128BIDAsBidgo(d))
	default:
		result, flags = bidgo.Bid128ToUint64Rnint(decimal128BIDAsBidgo(d))
	}
	return result, bidgoExceptionFlags(flags)
}

func decimal128BIDConvertToUint64ExactPort(d Decimal128BID, mode RoundingMode) (uint64, ExceptionFlags) {
	var result uint64
	var flags uint32
	switch mode {
	case RoundNearestEven:
		result, flags = bidgo.Bid128ToUint64Xrnint(decimal128BIDAsBidgo(d))
	case RoundNearestAway:
		result, flags = bidgo.Bid128ToUint64Xrninta(decimal128BIDAsBidgo(d))
	case RoundTowardZero:
		result, flags = bidgo.Bid128ToUint64Xint(decimal128BIDAsBidgo(d))
	case RoundTowardPositive:
		result, flags = bidgo.Bid128ToUint64Xceil(decimal128BIDAsBidgo(d))
	case RoundTowardNegative:
		result, flags = bidgo.Bid128ToUint64Xfloor(decimal128BIDAsBidgo(d))
	default:
		result, flags = bidgo.Bid128ToUint64Xrnint(decimal128BIDAsBidgo(d))
	}
	return result, bidgoExceptionFlags(flags)
}

func decimal32BIDFromInt32Port(x int32, mode RoundingMode) (Decimal32BID, ExceptionFlags) {
	result, flags := bidgo.Bid32FromInt32(x, bidgoRoundingMode(mode))
	return Decimal32BID(result), bidgoExceptionFlags(flags)
}

func decimal32BIDFromUint32Port(x uint32, mode RoundingMode) (Decimal32BID, ExceptionFlags) {
	result, flags := bidgo.Bid32FromUint32(x, bidgoRoundingMode(mode))
	return Decimal32BID(result), bidgoExceptionFlags(flags)
}

func decimal32BIDFromInt64Port(x int64, mode RoundingMode) (Decimal32BID, ExceptionFlags) {
	result, flags := bidgo.Bid32FromInt64(x, bidgoRoundingMode(mode))
	return Decimal32BID(result), bidgoExceptionFlags(flags)
}

func decimal32BIDFromUint64Port(x uint64, mode RoundingMode) (Decimal32BID, ExceptionFlags) {
	result, flags := bidgo.Bid32FromUint64(x, bidgoRoundingMode(mode))
	return Decimal32BID(result), bidgoExceptionFlags(flags)
}

func decimal64BIDFromInt32Port(x int32) Decimal64BID {
	return Decimal64BID(bidgo.Bid64FromInt32(x))
}

func decimal64BIDFromUint32Port(x uint32) Decimal64BID {
	return Decimal64BID(bidgo.Bid64FromUint32(x))
}

func decimal64BIDFromInt64Port(x int64, mode RoundingMode) (Decimal64BID, ExceptionFlags) {
	result, flags := bidgo.Bid64FromInt64(x, bidgoRoundingMode(mode))
	return Decimal64BID(result), bidgoExceptionFlags(flags)
}

func decimal64BIDFromUint64Port(x uint64, mode RoundingMode) (Decimal64BID, ExceptionFlags) {
	result, flags := bidgo.Bid64FromUint64(x, bidgoRoundingMode(mode))
	return Decimal64BID(result), bidgoExceptionFlags(flags)
}

func decimal128BIDFromInt32Port(x int32) Decimal128BID {
	return decimal128BIDFromBidgo(bidgo.Bid128FromInt32(x))
}

func decimal128BIDFromUint32Port(x uint32) Decimal128BID {
	return decimal128BIDFromBidgo(bidgo.Bid128FromUint32(x))
}

func decimal128BIDFromInt64Port(x int64) Decimal128BID {
	return decimal128BIDFromBidgo(bidgo.Bid128FromInt64(x))
}

func decimal128BIDFromUint64Port(x uint64) Decimal128BID {
	return decimal128BIDFromBidgo(bidgo.Bid128FromUint64(x))
}
