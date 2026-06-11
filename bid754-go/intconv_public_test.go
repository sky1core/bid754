package bid754

import (
	"testing"

	bidgo "github.com/sky1core/bid754/bid754-go/internal/bidgo"
)

type intConvInput32 struct {
	name  string
	value Decimal32BID
}

type intConvInput64 struct {
	name  string
	value Decimal64BID
}

type intConvInput128 struct {
	name  string
	value Decimal128BID
}

type intConvPort32 struct {
	name   string
	mode   RoundingMode
	direct func(uint32) (any, uint32)
}

type intConvPort64 struct {
	name   string
	mode   RoundingMode
	direct func(uint64) (any, uint32)
}

type intConvPort128 struct {
	name   string
	mode   RoundingMode
	direct func(bidgo.BID_UINT128) (any, uint32)
}

func intConvInputs32(t *testing.T) []intConvInput32 {
	t.Helper()
	return []intConvInput32{
		{name: "positive half", value: mustDecimal32BID(t, "1.5")},
		{name: "negative half", value: mustDecimal32BID(t, "-1.5")},
		{name: "even half", value: mustDecimal32BID(t, "2.5")},
		{name: "qnan", value: mustDecimal32BID(t, "NaN")},
		{name: "inf", value: mustDecimal32BID(t, "Infinity")},
		{name: "large positive", value: mustDecimal32BID(t, "1E20")},
		{name: "large negative", value: mustDecimal32BID(t, "-1E20")},
		{name: "int8 boundary", value: mustDecimal32BID(t, "127.5")},
		{name: "int8 negative boundary", value: mustDecimal32BID(t, "-128.5")},
		{name: "int32 overflow", value: mustDecimal32BID(t, "2147484000")},
		{name: "snan", value: mustDecimal32BID(t, "sNaN")},
	}
}

func intConvInputs64(t *testing.T) []intConvInput64 {
	t.Helper()
	return []intConvInput64{
		{name: "positive half", value: mustDecimal64BID(t, "1.5")},
		{name: "negative half", value: mustDecimal64BID(t, "-1.5")},
		{name: "even half", value: mustDecimal64BID(t, "2.5")},
		{name: "qnan", value: mustDecimal64BID(t, "NaN")},
		{name: "inf", value: mustDecimal64BID(t, "Infinity")},
		{name: "large positive", value: mustDecimal64BID(t, "1E20")},
		{name: "large negative", value: mustDecimal64BID(t, "-1E20")},
		{name: "int8 boundary", value: mustDecimal64BID(t, "127.5")},
		{name: "int8 negative boundary", value: mustDecimal64BID(t, "-128.5")},
		{name: "int32 overflow", value: mustDecimal64BID(t, "2147483648")},
		{name: "int64 overflow", value: mustDecimal64BID(t, "9223372036854775808")},
		{name: "snan", value: mustDecimal64BID(t, "sNaN")},
	}
}

func intConvInputs128(t *testing.T) []intConvInput128 {
	t.Helper()
	return []intConvInput128{
		{name: "positive half", value: mustDecimal128BID(t, "1.5")},
		{name: "negative half", value: mustDecimal128BID(t, "-1.5")},
		{name: "even half", value: mustDecimal128BID(t, "2.5")},
		{name: "qnan", value: mustDecimal128BID(t, "NaN")},
		{name: "inf", value: mustDecimal128BID(t, "Infinity")},
		{name: "large positive", value: mustDecimal128BID(t, "1E20")},
		{name: "large negative", value: mustDecimal128BID(t, "-1E20")},
		{name: "int8 boundary", value: mustDecimal128BID(t, "127.5")},
		{name: "int8 negative boundary", value: mustDecimal128BID(t, "-128.5")},
		{name: "int32 overflow", value: mustDecimal128BID(t, "2147483648")},
		{name: "int64 overflow", value: mustDecimal128BID(t, "9223372036854775808")},
		{name: "snan", value: mustDecimal128BID(t, "sNaN")},
	}
}
func TestConvertToIntegerPublicRoutingDecimal32(t *testing.T) {
	tests := []struct {
		name   string
		public func(Decimal32BID, RoundingMode) (any, ExceptionFlags)
		ports  []intConvPort32
	}{
		{name: "ConvertToInt8", public: func(d Decimal32BID, mode RoundingMode) (any, ExceptionFlags) {
			result, flags := d.ConvertToInt8(mode)
			return result, flags
		}, ports: []intConvPort32{
			{name: "nearestEven", mode: RoundNearestEven, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToInt8Rnint(x) }},
			{name: "nearestAway", mode: RoundNearestAway, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToInt8Rninta(x) }},
			{name: "towardZero", mode: RoundTowardZero, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToInt8Int(x) }},
			{name: "towardPositive", mode: RoundTowardPositive, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToInt8Ceil(x) }},
			{name: "towardNegative", mode: RoundTowardNegative, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToInt8Floor(x) }},
			{name: "unknownDefault", mode: RoundingMode(99), direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToInt8Rnint(x) }},
		}},
		{name: "ConvertToInt8Exact", public: func(d Decimal32BID, mode RoundingMode) (any, ExceptionFlags) {
			result, flags := d.ConvertToInt8Exact(mode)
			return result, flags
		}, ports: []intConvPort32{
			{name: "nearestEven", mode: RoundNearestEven, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToInt8Xrnint(x) }},
			{name: "nearestAway", mode: RoundNearestAway, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToInt8Xrninta(x) }},
			{name: "towardZero", mode: RoundTowardZero, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToInt8Xint(x) }},
			{name: "towardPositive", mode: RoundTowardPositive, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToInt8Xceil(x) }},
			{name: "towardNegative", mode: RoundTowardNegative, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToInt8Xfloor(x) }},
			{name: "unknownDefault", mode: RoundingMode(99), direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToInt8Xrnint(x) }},
		}},
		{name: "ConvertToInt16", public: func(d Decimal32BID, mode RoundingMode) (any, ExceptionFlags) {
			result, flags := d.ConvertToInt16(mode)
			return result, flags
		}, ports: []intConvPort32{
			{name: "nearestEven", mode: RoundNearestEven, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToInt16Rnint(x) }},
			{name: "nearestAway", mode: RoundNearestAway, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToInt16Rninta(x) }},
			{name: "towardZero", mode: RoundTowardZero, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToInt16Int(x) }},
			{name: "towardPositive", mode: RoundTowardPositive, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToInt16Ceil(x) }},
			{name: "towardNegative", mode: RoundTowardNegative, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToInt16Floor(x) }},
			{name: "unknownDefault", mode: RoundingMode(99), direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToInt16Rnint(x) }},
		}},
		{name: "ConvertToInt16Exact", public: func(d Decimal32BID, mode RoundingMode) (any, ExceptionFlags) {
			result, flags := d.ConvertToInt16Exact(mode)
			return result, flags
		}, ports: []intConvPort32{
			{name: "nearestEven", mode: RoundNearestEven, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToInt16Xrnint(x) }},
			{name: "nearestAway", mode: RoundNearestAway, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToInt16Xrninta(x) }},
			{name: "towardZero", mode: RoundTowardZero, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToInt16Xint(x) }},
			{name: "towardPositive", mode: RoundTowardPositive, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToInt16Xceil(x) }},
			{name: "towardNegative", mode: RoundTowardNegative, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToInt16Xfloor(x) }},
			{name: "unknownDefault", mode: RoundingMode(99), direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToInt16Xrnint(x) }},
		}},
		{name: "ConvertToInt32", public: func(d Decimal32BID, mode RoundingMode) (any, ExceptionFlags) {
			result, flags := d.ConvertToInt32(mode)
			return result, flags
		}, ports: []intConvPort32{
			{name: "nearestEven", mode: RoundNearestEven, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToInt32Rnint(x) }},
			{name: "nearestAway", mode: RoundNearestAway, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToInt32Rninta(x) }},
			{name: "towardZero", mode: RoundTowardZero, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToInt32Int(x) }},
			{name: "towardPositive", mode: RoundTowardPositive, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToInt32Ceil(x) }},
			{name: "towardNegative", mode: RoundTowardNegative, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToInt32Floor(x) }},
			{name: "unknownDefault", mode: RoundingMode(99), direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToInt32Rnint(x) }},
		}},
		{name: "ConvertToInt32Exact", public: func(d Decimal32BID, mode RoundingMode) (any, ExceptionFlags) {
			result, flags := d.ConvertToInt32Exact(mode)
			return result, flags
		}, ports: []intConvPort32{
			{name: "nearestEven", mode: RoundNearestEven, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToInt32Xrnint(x) }},
			{name: "nearestAway", mode: RoundNearestAway, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToInt32Xrninta(x) }},
			{name: "towardZero", mode: RoundTowardZero, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToInt32Xint(x) }},
			{name: "towardPositive", mode: RoundTowardPositive, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToInt32Xceil(x) }},
			{name: "towardNegative", mode: RoundTowardNegative, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToInt32Xfloor(x) }},
			{name: "unknownDefault", mode: RoundingMode(99), direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToInt32Xrnint(x) }},
		}},
		{name: "ConvertToInt64", public: func(d Decimal32BID, mode RoundingMode) (any, ExceptionFlags) {
			result, flags := d.ConvertToInt64(mode)
			return result, flags
		}, ports: []intConvPort32{
			{name: "nearestEven", mode: RoundNearestEven, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToInt64Rnint(x) }},
			{name: "nearestAway", mode: RoundNearestAway, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToInt64Rninta(x) }},
			{name: "towardZero", mode: RoundTowardZero, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToInt64Int(x) }},
			{name: "towardPositive", mode: RoundTowardPositive, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToInt64Ceil(x) }},
			{name: "towardNegative", mode: RoundTowardNegative, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToInt64Floor(x) }},
			{name: "unknownDefault", mode: RoundingMode(99), direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToInt64Rnint(x) }},
		}},
		{name: "ConvertToInt64Exact", public: func(d Decimal32BID, mode RoundingMode) (any, ExceptionFlags) {
			result, flags := d.ConvertToInt64Exact(mode)
			return result, flags
		}, ports: []intConvPort32{
			{name: "nearestEven", mode: RoundNearestEven, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToInt64Xrnint(x) }},
			{name: "nearestAway", mode: RoundNearestAway, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToInt64Xrninta(x) }},
			{name: "towardZero", mode: RoundTowardZero, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToInt64Xint(x) }},
			{name: "towardPositive", mode: RoundTowardPositive, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToInt64Xceil(x) }},
			{name: "towardNegative", mode: RoundTowardNegative, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToInt64Xfloor(x) }},
			{name: "unknownDefault", mode: RoundingMode(99), direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToInt64Xrnint(x) }},
		}},
		{name: "ConvertToUint8", public: func(d Decimal32BID, mode RoundingMode) (any, ExceptionFlags) {
			result, flags := d.ConvertToUint8(mode)
			return result, flags
		}, ports: []intConvPort32{
			{name: "nearestEven", mode: RoundNearestEven, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToUint8Rnint(x) }},
			{name: "nearestAway", mode: RoundNearestAway, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToUint8Rninta(x) }},
			{name: "towardZero", mode: RoundTowardZero, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToUint8Int(x) }},
			{name: "towardPositive", mode: RoundTowardPositive, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToUint8Ceil(x) }},
			{name: "towardNegative", mode: RoundTowardNegative, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToUint8Floor(x) }},
			{name: "unknownDefault", mode: RoundingMode(99), direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToUint8Rnint(x) }},
		}},
		{name: "ConvertToUint8Exact", public: func(d Decimal32BID, mode RoundingMode) (any, ExceptionFlags) {
			result, flags := d.ConvertToUint8Exact(mode)
			return result, flags
		}, ports: []intConvPort32{
			{name: "nearestEven", mode: RoundNearestEven, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToUint8Xrnint(x) }},
			{name: "nearestAway", mode: RoundNearestAway, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToUint8Xrninta(x) }},
			{name: "towardZero", mode: RoundTowardZero, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToUint8Xint(x) }},
			{name: "towardPositive", mode: RoundTowardPositive, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToUint8Xceil(x) }},
			{name: "towardNegative", mode: RoundTowardNegative, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToUint8Xfloor(x) }},
			{name: "unknownDefault", mode: RoundingMode(99), direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToUint8Xrnint(x) }},
		}},
		{name: "ConvertToUint16", public: func(d Decimal32BID, mode RoundingMode) (any, ExceptionFlags) {
			result, flags := d.ConvertToUint16(mode)
			return result, flags
		}, ports: []intConvPort32{
			{name: "nearestEven", mode: RoundNearestEven, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToUint16Rnint(x) }},
			{name: "nearestAway", mode: RoundNearestAway, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToUint16Rninta(x) }},
			{name: "towardZero", mode: RoundTowardZero, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToUint16Int(x) }},
			{name: "towardPositive", mode: RoundTowardPositive, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToUint16Ceil(x) }},
			{name: "towardNegative", mode: RoundTowardNegative, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToUint16Floor(x) }},
			{name: "unknownDefault", mode: RoundingMode(99), direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToUint16Rnint(x) }},
		}},
		{name: "ConvertToUint16Exact", public: func(d Decimal32BID, mode RoundingMode) (any, ExceptionFlags) {
			result, flags := d.ConvertToUint16Exact(mode)
			return result, flags
		}, ports: []intConvPort32{
			{name: "nearestEven", mode: RoundNearestEven, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToUint16Xrnint(x) }},
			{name: "nearestAway", mode: RoundNearestAway, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToUint16Xrninta(x) }},
			{name: "towardZero", mode: RoundTowardZero, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToUint16Xint(x) }},
			{name: "towardPositive", mode: RoundTowardPositive, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToUint16Xceil(x) }},
			{name: "towardNegative", mode: RoundTowardNegative, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToUint16Xfloor(x) }},
			{name: "unknownDefault", mode: RoundingMode(99), direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToUint16Xrnint(x) }},
		}},
		{name: "ConvertToUint32", public: func(d Decimal32BID, mode RoundingMode) (any, ExceptionFlags) {
			result, flags := d.ConvertToUint32(mode)
			return result, flags
		}, ports: []intConvPort32{
			{name: "nearestEven", mode: RoundNearestEven, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToUint32Rnint(x) }},
			{name: "nearestAway", mode: RoundNearestAway, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToUint32Rninta(x) }},
			{name: "towardZero", mode: RoundTowardZero, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToUint32Int(x) }},
			{name: "towardPositive", mode: RoundTowardPositive, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToUint32Ceil(x) }},
			{name: "towardNegative", mode: RoundTowardNegative, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToUint32Floor(x) }},
			{name: "unknownDefault", mode: RoundingMode(99), direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToUint32Rnint(x) }},
		}},
		{name: "ConvertToUint32Exact", public: func(d Decimal32BID, mode RoundingMode) (any, ExceptionFlags) {
			result, flags := d.ConvertToUint32Exact(mode)
			return result, flags
		}, ports: []intConvPort32{
			{name: "nearestEven", mode: RoundNearestEven, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToUint32Xrnint(x) }},
			{name: "nearestAway", mode: RoundNearestAway, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToUint32Xrninta(x) }},
			{name: "towardZero", mode: RoundTowardZero, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToUint32Xint(x) }},
			{name: "towardPositive", mode: RoundTowardPositive, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToUint32Xceil(x) }},
			{name: "towardNegative", mode: RoundTowardNegative, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToUint32Xfloor(x) }},
			{name: "unknownDefault", mode: RoundingMode(99), direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToUint32Xrnint(x) }},
		}},
		{name: "ConvertToUint64", public: func(d Decimal32BID, mode RoundingMode) (any, ExceptionFlags) {
			result, flags := d.ConvertToUint64(mode)
			return result, flags
		}, ports: []intConvPort32{
			{name: "nearestEven", mode: RoundNearestEven, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToUint64Rnint(x) }},
			{name: "nearestAway", mode: RoundNearestAway, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToUint64Rninta(x) }},
			{name: "towardZero", mode: RoundTowardZero, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToUint64Int(x) }},
			{name: "towardPositive", mode: RoundTowardPositive, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToUint64Ceil(x) }},
			{name: "towardNegative", mode: RoundTowardNegative, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToUint64Floor(x) }},
			{name: "unknownDefault", mode: RoundingMode(99), direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToUint64Rnint(x) }},
		}},
		{name: "ConvertToUint64Exact", public: func(d Decimal32BID, mode RoundingMode) (any, ExceptionFlags) {
			result, flags := d.ConvertToUint64Exact(mode)
			return result, flags
		}, ports: []intConvPort32{
			{name: "nearestEven", mode: RoundNearestEven, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToUint64Xrnint(x) }},
			{name: "nearestAway", mode: RoundNearestAway, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToUint64Xrninta(x) }},
			{name: "towardZero", mode: RoundTowardZero, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToUint64Xint(x) }},
			{name: "towardPositive", mode: RoundTowardPositive, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToUint64Xceil(x) }},
			{name: "towardNegative", mode: RoundTowardNegative, direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToUint64Xfloor(x) }},
			{name: "unknownDefault", mode: RoundingMode(99), direct: func(x uint32) (any, uint32) { return bidgo.Bid32ToUint64Xrnint(x) }},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, port := range tt.ports {
				for _, input := range intConvInputs32(t) {
					got, gotFlags := tt.public(input.value, port.mode)
					want, wantFlags := port.direct(input.value.ToUint32())
					if got != want || gotFlags != bidgoExceptionFlags(wantFlags) {
						t.Fatalf("%s/%s(%s) = %v/%s, want %v/%s", tt.name, port.name, input.name, got, gotFlags, want, bidgoExceptionFlags(wantFlags))
					}
				}
			}
		})
	}
}

func TestConvertToIntegerPublicRoutingDecimal64(t *testing.T) {
	tests := []struct {
		name   string
		public func(Decimal64BID, RoundingMode) (any, ExceptionFlags)
		ports  []intConvPort64
	}{
		{name: "ConvertToInt8", public: func(d Decimal64BID, mode RoundingMode) (any, ExceptionFlags) {
			result, flags := d.ConvertToInt8(mode)
			return result, flags
		}, ports: []intConvPort64{
			{name: "nearestEven", mode: RoundNearestEven, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToInt8Rnint(x) }},
			{name: "nearestAway", mode: RoundNearestAway, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToInt8Rninta(x) }},
			{name: "towardZero", mode: RoundTowardZero, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToInt8Int(x) }},
			{name: "towardPositive", mode: RoundTowardPositive, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToInt8Ceil(x) }},
			{name: "towardNegative", mode: RoundTowardNegative, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToInt8Floor(x) }},
			{name: "unknownDefault", mode: RoundingMode(99), direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToInt8Rnint(x) }},
		}},
		{name: "ConvertToInt8Exact", public: func(d Decimal64BID, mode RoundingMode) (any, ExceptionFlags) {
			result, flags := d.ConvertToInt8Exact(mode)
			return result, flags
		}, ports: []intConvPort64{
			{name: "nearestEven", mode: RoundNearestEven, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToInt8Xrnint(x) }},
			{name: "nearestAway", mode: RoundNearestAway, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToInt8Xrninta(x) }},
			{name: "towardZero", mode: RoundTowardZero, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToInt8Xint(x) }},
			{name: "towardPositive", mode: RoundTowardPositive, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToInt8Xceil(x) }},
			{name: "towardNegative", mode: RoundTowardNegative, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToInt8Xfloor(x) }},
			{name: "unknownDefault", mode: RoundingMode(99), direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToInt8Xrnint(x) }},
		}},
		{name: "ConvertToInt16", public: func(d Decimal64BID, mode RoundingMode) (any, ExceptionFlags) {
			result, flags := d.ConvertToInt16(mode)
			return result, flags
		}, ports: []intConvPort64{
			{name: "nearestEven", mode: RoundNearestEven, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToInt16Rnint(x) }},
			{name: "nearestAway", mode: RoundNearestAway, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToInt16Rninta(x) }},
			{name: "towardZero", mode: RoundTowardZero, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToInt16Int(x) }},
			{name: "towardPositive", mode: RoundTowardPositive, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToInt16Ceil(x) }},
			{name: "towardNegative", mode: RoundTowardNegative, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToInt16Floor(x) }},
			{name: "unknownDefault", mode: RoundingMode(99), direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToInt16Rnint(x) }},
		}},
		{name: "ConvertToInt16Exact", public: func(d Decimal64BID, mode RoundingMode) (any, ExceptionFlags) {
			result, flags := d.ConvertToInt16Exact(mode)
			return result, flags
		}, ports: []intConvPort64{
			{name: "nearestEven", mode: RoundNearestEven, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToInt16Xrnint(x) }},
			{name: "nearestAway", mode: RoundNearestAway, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToInt16Xrninta(x) }},
			{name: "towardZero", mode: RoundTowardZero, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToInt16Xint(x) }},
			{name: "towardPositive", mode: RoundTowardPositive, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToInt16Xceil(x) }},
			{name: "towardNegative", mode: RoundTowardNegative, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToInt16Xfloor(x) }},
			{name: "unknownDefault", mode: RoundingMode(99), direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToInt16Xrnint(x) }},
		}},
		{name: "ConvertToInt32", public: func(d Decimal64BID, mode RoundingMode) (any, ExceptionFlags) {
			result, flags := d.ConvertToInt32(mode)
			return result, flags
		}, ports: []intConvPort64{
			{name: "nearestEven", mode: RoundNearestEven, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToInt32Rnint(x) }},
			{name: "nearestAway", mode: RoundNearestAway, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToInt32Rninta(x) }},
			{name: "towardZero", mode: RoundTowardZero, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToInt32Int(x) }},
			{name: "towardPositive", mode: RoundTowardPositive, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToInt32Ceil(x) }},
			{name: "towardNegative", mode: RoundTowardNegative, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToInt32Floor(x) }},
			{name: "unknownDefault", mode: RoundingMode(99), direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToInt32Rnint(x) }},
		}},
		{name: "ConvertToInt32Exact", public: func(d Decimal64BID, mode RoundingMode) (any, ExceptionFlags) {
			result, flags := d.ConvertToInt32Exact(mode)
			return result, flags
		}, ports: []intConvPort64{
			{name: "nearestEven", mode: RoundNearestEven, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToInt32Xrnint(x) }},
			{name: "nearestAway", mode: RoundNearestAway, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToInt32Xrninta(x) }},
			{name: "towardZero", mode: RoundTowardZero, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToInt32Xint(x) }},
			{name: "towardPositive", mode: RoundTowardPositive, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToInt32Xceil(x) }},
			{name: "towardNegative", mode: RoundTowardNegative, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToInt32Xfloor(x) }},
			{name: "unknownDefault", mode: RoundingMode(99), direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToInt32Xrnint(x) }},
		}},
		{name: "ConvertToInt64", public: func(d Decimal64BID, mode RoundingMode) (any, ExceptionFlags) {
			result, flags := d.ConvertToInt64(mode)
			return result, flags
		}, ports: []intConvPort64{
			{name: "nearestEven", mode: RoundNearestEven, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToInt64Rnint(x) }},
			{name: "nearestAway", mode: RoundNearestAway, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToInt64Rninta(x) }},
			{name: "towardZero", mode: RoundTowardZero, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToInt64Int(x) }},
			{name: "towardPositive", mode: RoundTowardPositive, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToInt64Ceil(x) }},
			{name: "towardNegative", mode: RoundTowardNegative, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToInt64Floor(x) }},
			{name: "unknownDefault", mode: RoundingMode(99), direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToInt64Rnint(x) }},
		}},
		{name: "ConvertToInt64Exact", public: func(d Decimal64BID, mode RoundingMode) (any, ExceptionFlags) {
			result, flags := d.ConvertToInt64Exact(mode)
			return result, flags
		}, ports: []intConvPort64{
			{name: "nearestEven", mode: RoundNearestEven, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToInt64Xrnint(x) }},
			{name: "nearestAway", mode: RoundNearestAway, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToInt64Xrninta(x) }},
			{name: "towardZero", mode: RoundTowardZero, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToInt64Xint(x) }},
			{name: "towardPositive", mode: RoundTowardPositive, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToInt64Xceil(x) }},
			{name: "towardNegative", mode: RoundTowardNegative, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToInt64Xfloor(x) }},
			{name: "unknownDefault", mode: RoundingMode(99), direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToInt64Xrnint(x) }},
		}},
		{name: "ConvertToUint8", public: func(d Decimal64BID, mode RoundingMode) (any, ExceptionFlags) {
			result, flags := d.ConvertToUint8(mode)
			return result, flags
		}, ports: []intConvPort64{
			{name: "nearestEven", mode: RoundNearestEven, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToUint8Rnint(x) }},
			{name: "nearestAway", mode: RoundNearestAway, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToUint8Rninta(x) }},
			{name: "towardZero", mode: RoundTowardZero, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToUint8Int(x) }},
			{name: "towardPositive", mode: RoundTowardPositive, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToUint8Ceil(x) }},
			{name: "towardNegative", mode: RoundTowardNegative, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToUint8Floor(x) }},
			{name: "unknownDefault", mode: RoundingMode(99), direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToUint8Rnint(x) }},
		}},
		{name: "ConvertToUint8Exact", public: func(d Decimal64BID, mode RoundingMode) (any, ExceptionFlags) {
			result, flags := d.ConvertToUint8Exact(mode)
			return result, flags
		}, ports: []intConvPort64{
			{name: "nearestEven", mode: RoundNearestEven, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToUint8Xrnint(x) }},
			{name: "nearestAway", mode: RoundNearestAway, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToUint8Xrninta(x) }},
			{name: "towardZero", mode: RoundTowardZero, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToUint8Xint(x) }},
			{name: "towardPositive", mode: RoundTowardPositive, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToUint8Xceil(x) }},
			{name: "towardNegative", mode: RoundTowardNegative, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToUint8Xfloor(x) }},
			{name: "unknownDefault", mode: RoundingMode(99), direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToUint8Xrnint(x) }},
		}},
		{name: "ConvertToUint16", public: func(d Decimal64BID, mode RoundingMode) (any, ExceptionFlags) {
			result, flags := d.ConvertToUint16(mode)
			return result, flags
		}, ports: []intConvPort64{
			{name: "nearestEven", mode: RoundNearestEven, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToUint16Rnint(x) }},
			{name: "nearestAway", mode: RoundNearestAway, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToUint16Rninta(x) }},
			{name: "towardZero", mode: RoundTowardZero, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToUint16Int(x) }},
			{name: "towardPositive", mode: RoundTowardPositive, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToUint16Ceil(x) }},
			{name: "towardNegative", mode: RoundTowardNegative, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToUint16Floor(x) }},
			{name: "unknownDefault", mode: RoundingMode(99), direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToUint16Rnint(x) }},
		}},
		{name: "ConvertToUint16Exact", public: func(d Decimal64BID, mode RoundingMode) (any, ExceptionFlags) {
			result, flags := d.ConvertToUint16Exact(mode)
			return result, flags
		}, ports: []intConvPort64{
			{name: "nearestEven", mode: RoundNearestEven, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToUint16Xrnint(x) }},
			{name: "nearestAway", mode: RoundNearestAway, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToUint16Xrninta(x) }},
			{name: "towardZero", mode: RoundTowardZero, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToUint16Xint(x) }},
			{name: "towardPositive", mode: RoundTowardPositive, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToUint16Xceil(x) }},
			{name: "towardNegative", mode: RoundTowardNegative, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToUint16Xfloor(x) }},
			{name: "unknownDefault", mode: RoundingMode(99), direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToUint16Xrnint(x) }},
		}},
		{name: "ConvertToUint32", public: func(d Decimal64BID, mode RoundingMode) (any, ExceptionFlags) {
			result, flags := d.ConvertToUint32(mode)
			return result, flags
		}, ports: []intConvPort64{
			{name: "nearestEven", mode: RoundNearestEven, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToUint32Rnint(x) }},
			{name: "nearestAway", mode: RoundNearestAway, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToUint32Rninta(x) }},
			{name: "towardZero", mode: RoundTowardZero, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToUint32Int(x) }},
			{name: "towardPositive", mode: RoundTowardPositive, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToUint32Ceil(x) }},
			{name: "towardNegative", mode: RoundTowardNegative, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToUint32Floor(x) }},
			{name: "unknownDefault", mode: RoundingMode(99), direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToUint32Rnint(x) }},
		}},
		{name: "ConvertToUint32Exact", public: func(d Decimal64BID, mode RoundingMode) (any, ExceptionFlags) {
			result, flags := d.ConvertToUint32Exact(mode)
			return result, flags
		}, ports: []intConvPort64{
			{name: "nearestEven", mode: RoundNearestEven, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToUint32Xrnint(x) }},
			{name: "nearestAway", mode: RoundNearestAway, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToUint32Xrninta(x) }},
			{name: "towardZero", mode: RoundTowardZero, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToUint32Xint(x) }},
			{name: "towardPositive", mode: RoundTowardPositive, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToUint32Xceil(x) }},
			{name: "towardNegative", mode: RoundTowardNegative, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToUint32Xfloor(x) }},
			{name: "unknownDefault", mode: RoundingMode(99), direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToUint32Xrnint(x) }},
		}},
		{name: "ConvertToUint64", public: func(d Decimal64BID, mode RoundingMode) (any, ExceptionFlags) {
			result, flags := d.ConvertToUint64(mode)
			return result, flags
		}, ports: []intConvPort64{
			{name: "nearestEven", mode: RoundNearestEven, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToUint64Rnint(x) }},
			{name: "nearestAway", mode: RoundNearestAway, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToUint64Rninta(x) }},
			{name: "towardZero", mode: RoundTowardZero, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToUint64Int(x) }},
			{name: "towardPositive", mode: RoundTowardPositive, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToUint64Ceil(x) }},
			{name: "towardNegative", mode: RoundTowardNegative, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToUint64Floor(x) }},
			{name: "unknownDefault", mode: RoundingMode(99), direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToUint64Rnint(x) }},
		}},
		{name: "ConvertToUint64Exact", public: func(d Decimal64BID, mode RoundingMode) (any, ExceptionFlags) {
			result, flags := d.ConvertToUint64Exact(mode)
			return result, flags
		}, ports: []intConvPort64{
			{name: "nearestEven", mode: RoundNearestEven, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToUint64Xrnint(x) }},
			{name: "nearestAway", mode: RoundNearestAway, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToUint64Xrninta(x) }},
			{name: "towardZero", mode: RoundTowardZero, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToUint64Xint(x) }},
			{name: "towardPositive", mode: RoundTowardPositive, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToUint64Xceil(x) }},
			{name: "towardNegative", mode: RoundTowardNegative, direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToUint64Xfloor(x) }},
			{name: "unknownDefault", mode: RoundingMode(99), direct: func(x uint64) (any, uint32) { return bidgo.Bid64ToUint64Xrnint(x) }},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, port := range tt.ports {
				for _, input := range intConvInputs64(t) {
					got, gotFlags := tt.public(input.value, port.mode)
					want, wantFlags := port.direct(input.value.ToUint64())
					if got != want || gotFlags != bidgoExceptionFlags(wantFlags) {
						t.Fatalf("%s/%s(%s) = %v/%s, want %v/%s", tt.name, port.name, input.name, got, gotFlags, want, bidgoExceptionFlags(wantFlags))
					}
				}
			}
		})
	}
}

func TestConvertToIntegerPublicRoutingDecimal128(t *testing.T) {
	tests := []struct {
		name   string
		public func(Decimal128BID, RoundingMode) (any, ExceptionFlags)
		ports  []intConvPort128
	}{
		{name: "ConvertToInt8", public: func(d Decimal128BID, mode RoundingMode) (any, ExceptionFlags) {
			result, flags := d.ConvertToInt8(mode)
			return result, flags
		}, ports: []intConvPort128{
			{name: "nearestEven", mode: RoundNearestEven, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToInt8Rnint(x) }},
			{name: "nearestAway", mode: RoundNearestAway, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToInt8Rninta(x) }},
			{name: "towardZero", mode: RoundTowardZero, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToInt8Int(x) }},
			{name: "towardPositive", mode: RoundTowardPositive, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToInt8Ceil(x) }},
			{name: "towardNegative", mode: RoundTowardNegative, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToInt8Floor(x) }},
			{name: "unknownDefault", mode: RoundingMode(99), direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToInt8Rnint(x) }},
		}},
		{name: "ConvertToInt8Exact", public: func(d Decimal128BID, mode RoundingMode) (any, ExceptionFlags) {
			result, flags := d.ConvertToInt8Exact(mode)
			return result, flags
		}, ports: []intConvPort128{
			{name: "nearestEven", mode: RoundNearestEven, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToInt8Xrnint(x) }},
			{name: "nearestAway", mode: RoundNearestAway, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToInt8Xrninta(x) }},
			{name: "towardZero", mode: RoundTowardZero, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToInt8Xint(x) }},
			{name: "towardPositive", mode: RoundTowardPositive, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToInt8Xceil(x) }},
			{name: "towardNegative", mode: RoundTowardNegative, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToInt8Xfloor(x) }},
			{name: "unknownDefault", mode: RoundingMode(99), direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToInt8Xrnint(x) }},
		}},
		{name: "ConvertToInt16", public: func(d Decimal128BID, mode RoundingMode) (any, ExceptionFlags) {
			result, flags := d.ConvertToInt16(mode)
			return result, flags
		}, ports: []intConvPort128{
			{name: "nearestEven", mode: RoundNearestEven, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToInt16Rnint(x) }},
			{name: "nearestAway", mode: RoundNearestAway, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToInt16Rninta(x) }},
			{name: "towardZero", mode: RoundTowardZero, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToInt16Int(x) }},
			{name: "towardPositive", mode: RoundTowardPositive, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToInt16Ceil(x) }},
			{name: "towardNegative", mode: RoundTowardNegative, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToInt16Floor(x) }},
			{name: "unknownDefault", mode: RoundingMode(99), direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToInt16Rnint(x) }},
		}},
		{name: "ConvertToInt16Exact", public: func(d Decimal128BID, mode RoundingMode) (any, ExceptionFlags) {
			result, flags := d.ConvertToInt16Exact(mode)
			return result, flags
		}, ports: []intConvPort128{
			{name: "nearestEven", mode: RoundNearestEven, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToInt16Xrnint(x) }},
			{name: "nearestAway", mode: RoundNearestAway, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToInt16Xrninta(x) }},
			{name: "towardZero", mode: RoundTowardZero, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToInt16Xint(x) }},
			{name: "towardPositive", mode: RoundTowardPositive, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToInt16Xceil(x) }},
			{name: "towardNegative", mode: RoundTowardNegative, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToInt16Xfloor(x) }},
			{name: "unknownDefault", mode: RoundingMode(99), direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToInt16Xrnint(x) }},
		}},
		{name: "ConvertToInt32", public: func(d Decimal128BID, mode RoundingMode) (any, ExceptionFlags) {
			result, flags := d.ConvertToInt32(mode)
			return result, flags
		}, ports: []intConvPort128{
			{name: "nearestEven", mode: RoundNearestEven, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToInt32Rnint(x) }},
			{name: "nearestAway", mode: RoundNearestAway, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToInt32Rninta(x) }},
			{name: "towardZero", mode: RoundTowardZero, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToInt32Int(x) }},
			{name: "towardPositive", mode: RoundTowardPositive, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToInt32Ceil(x) }},
			{name: "towardNegative", mode: RoundTowardNegative, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToInt32Floor(x) }},
			{name: "unknownDefault", mode: RoundingMode(99), direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToInt32Rnint(x) }},
		}},
		{name: "ConvertToInt32Exact", public: func(d Decimal128BID, mode RoundingMode) (any, ExceptionFlags) {
			result, flags := d.ConvertToInt32Exact(mode)
			return result, flags
		}, ports: []intConvPort128{
			{name: "nearestEven", mode: RoundNearestEven, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToInt32Xrnint(x) }},
			{name: "nearestAway", mode: RoundNearestAway, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToInt32Xrninta(x) }},
			{name: "towardZero", mode: RoundTowardZero, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToInt32Xint(x) }},
			{name: "towardPositive", mode: RoundTowardPositive, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToInt32Xceil(x) }},
			{name: "towardNegative", mode: RoundTowardNegative, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToInt32Xfloor(x) }},
			{name: "unknownDefault", mode: RoundingMode(99), direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToInt32Xrnint(x) }},
		}},
		{name: "ConvertToInt64", public: func(d Decimal128BID, mode RoundingMode) (any, ExceptionFlags) {
			result, flags := d.ConvertToInt64(mode)
			return result, flags
		}, ports: []intConvPort128{
			{name: "nearestEven", mode: RoundNearestEven, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToInt64Rnint(x) }},
			{name: "nearestAway", mode: RoundNearestAway, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToInt64Rninta(x) }},
			{name: "towardZero", mode: RoundTowardZero, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToInt64Int(x) }},
			{name: "towardPositive", mode: RoundTowardPositive, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToInt64Ceil(x) }},
			{name: "towardNegative", mode: RoundTowardNegative, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToInt64Floor(x) }},
			{name: "unknownDefault", mode: RoundingMode(99), direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToInt64Rnint(x) }},
		}},
		{name: "ConvertToInt64Exact", public: func(d Decimal128BID, mode RoundingMode) (any, ExceptionFlags) {
			result, flags := d.ConvertToInt64Exact(mode)
			return result, flags
		}, ports: []intConvPort128{
			{name: "nearestEven", mode: RoundNearestEven, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToInt64Xrnint(x) }},
			{name: "nearestAway", mode: RoundNearestAway, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToInt64Xrninta(x) }},
			{name: "towardZero", mode: RoundTowardZero, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToInt64Xint(x) }},
			{name: "towardPositive", mode: RoundTowardPositive, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToInt64Xceil(x) }},
			{name: "towardNegative", mode: RoundTowardNegative, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToInt64Xfloor(x) }},
			{name: "unknownDefault", mode: RoundingMode(99), direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToInt64Xrnint(x) }},
		}},
		{name: "ConvertToUint8", public: func(d Decimal128BID, mode RoundingMode) (any, ExceptionFlags) {
			result, flags := d.ConvertToUint8(mode)
			return result, flags
		}, ports: []intConvPort128{
			{name: "nearestEven", mode: RoundNearestEven, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToUint8Rnint(x) }},
			{name: "nearestAway", mode: RoundNearestAway, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToUint8Rninta(x) }},
			{name: "towardZero", mode: RoundTowardZero, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToUint8Int(x) }},
			{name: "towardPositive", mode: RoundTowardPositive, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToUint8Ceil(x) }},
			{name: "towardNegative", mode: RoundTowardNegative, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToUint8Floor(x) }},
			{name: "unknownDefault", mode: RoundingMode(99), direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToUint8Rnint(x) }},
		}},
		{name: "ConvertToUint8Exact", public: func(d Decimal128BID, mode RoundingMode) (any, ExceptionFlags) {
			result, flags := d.ConvertToUint8Exact(mode)
			return result, flags
		}, ports: []intConvPort128{
			{name: "nearestEven", mode: RoundNearestEven, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToUint8Xrnint(x) }},
			{name: "nearestAway", mode: RoundNearestAway, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToUint8Xrninta(x) }},
			{name: "towardZero", mode: RoundTowardZero, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToUint8Xint(x) }},
			{name: "towardPositive", mode: RoundTowardPositive, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToUint8Xceil(x) }},
			{name: "towardNegative", mode: RoundTowardNegative, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToUint8Xfloor(x) }},
			{name: "unknownDefault", mode: RoundingMode(99), direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToUint8Xrnint(x) }},
		}},
		{name: "ConvertToUint16", public: func(d Decimal128BID, mode RoundingMode) (any, ExceptionFlags) {
			result, flags := d.ConvertToUint16(mode)
			return result, flags
		}, ports: []intConvPort128{
			{name: "nearestEven", mode: RoundNearestEven, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToUint16Rnint(x) }},
			{name: "nearestAway", mode: RoundNearestAway, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToUint16Rninta(x) }},
			{name: "towardZero", mode: RoundTowardZero, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToUint16Int(x) }},
			{name: "towardPositive", mode: RoundTowardPositive, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToUint16Ceil(x) }},
			{name: "towardNegative", mode: RoundTowardNegative, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToUint16Floor(x) }},
			{name: "unknownDefault", mode: RoundingMode(99), direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToUint16Rnint(x) }},
		}},
		{name: "ConvertToUint16Exact", public: func(d Decimal128BID, mode RoundingMode) (any, ExceptionFlags) {
			result, flags := d.ConvertToUint16Exact(mode)
			return result, flags
		}, ports: []intConvPort128{
			{name: "nearestEven", mode: RoundNearestEven, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToUint16Xrnint(x) }},
			{name: "nearestAway", mode: RoundNearestAway, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToUint16Xrninta(x) }},
			{name: "towardZero", mode: RoundTowardZero, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToUint16Xint(x) }},
			{name: "towardPositive", mode: RoundTowardPositive, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToUint16Xceil(x) }},
			{name: "towardNegative", mode: RoundTowardNegative, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToUint16Xfloor(x) }},
			{name: "unknownDefault", mode: RoundingMode(99), direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToUint16Xrnint(x) }},
		}},
		{name: "ConvertToUint32", public: func(d Decimal128BID, mode RoundingMode) (any, ExceptionFlags) {
			result, flags := d.ConvertToUint32(mode)
			return result, flags
		}, ports: []intConvPort128{
			{name: "nearestEven", mode: RoundNearestEven, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToUint32Rnint(x) }},
			{name: "nearestAway", mode: RoundNearestAway, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToUint32Rninta(x) }},
			{name: "towardZero", mode: RoundTowardZero, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToUint32Int(x) }},
			{name: "towardPositive", mode: RoundTowardPositive, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToUint32Ceil(x) }},
			{name: "towardNegative", mode: RoundTowardNegative, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToUint32Floor(x) }},
			{name: "unknownDefault", mode: RoundingMode(99), direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToUint32Rnint(x) }},
		}},
		{name: "ConvertToUint32Exact", public: func(d Decimal128BID, mode RoundingMode) (any, ExceptionFlags) {
			result, flags := d.ConvertToUint32Exact(mode)
			return result, flags
		}, ports: []intConvPort128{
			{name: "nearestEven", mode: RoundNearestEven, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToUint32Xrnint(x) }},
			{name: "nearestAway", mode: RoundNearestAway, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToUint32Xrninta(x) }},
			{name: "towardZero", mode: RoundTowardZero, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToUint32Xint(x) }},
			{name: "towardPositive", mode: RoundTowardPositive, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToUint32Xceil(x) }},
			{name: "towardNegative", mode: RoundTowardNegative, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToUint32Xfloor(x) }},
			{name: "unknownDefault", mode: RoundingMode(99), direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToUint32Xrnint(x) }},
		}},
		{name: "ConvertToUint64", public: func(d Decimal128BID, mode RoundingMode) (any, ExceptionFlags) {
			result, flags := d.ConvertToUint64(mode)
			return result, flags
		}, ports: []intConvPort128{
			{name: "nearestEven", mode: RoundNearestEven, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToUint64Rnint(x) }},
			{name: "nearestAway", mode: RoundNearestAway, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToUint64Rninta(x) }},
			{name: "towardZero", mode: RoundTowardZero, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToUint64Int(x) }},
			{name: "towardPositive", mode: RoundTowardPositive, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToUint64Ceil(x) }},
			{name: "towardNegative", mode: RoundTowardNegative, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToUint64Floor(x) }},
			{name: "unknownDefault", mode: RoundingMode(99), direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToUint64Rnint(x) }},
		}},
		{name: "ConvertToUint64Exact", public: func(d Decimal128BID, mode RoundingMode) (any, ExceptionFlags) {
			result, flags := d.ConvertToUint64Exact(mode)
			return result, flags
		}, ports: []intConvPort128{
			{name: "nearestEven", mode: RoundNearestEven, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToUint64Xrnint(x) }},
			{name: "nearestAway", mode: RoundNearestAway, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToUint64Xrninta(x) }},
			{name: "towardZero", mode: RoundTowardZero, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToUint64Xint(x) }},
			{name: "towardPositive", mode: RoundTowardPositive, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToUint64Xceil(x) }},
			{name: "towardNegative", mode: RoundTowardNegative, direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToUint64Xfloor(x) }},
			{name: "unknownDefault", mode: RoundingMode(99), direct: func(x bidgo.BID_UINT128) (any, uint32) { return bidgo.Bid128ToUint64Xrnint(x) }},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, port := range tt.ports {
				for _, input := range intConvInputs128(t) {
					got, gotFlags := tt.public(input.value, port.mode)
					want, wantFlags := port.direct(decimal128BIDAsBidgo(input.value))
					if got != want || gotFlags != bidgoExceptionFlags(wantFlags) {
						t.Fatalf("%s/%s(%s) = %v/%s, want %v/%s", tt.name, port.name, input.name, got, gotFlags, want, bidgoExceptionFlags(wantFlags))
					}
				}
			}
		})
	}
}

func TestConvertToIntegerModeAndFlagMatrix(t *testing.T) {
	t.Run("decimal32", func(t *testing.T) {
		frac := mustDecimal32BID(t, "1.5")
		assertIntConvModeDifference(t, []any{
			mustConvertToAny(frac.ConvertToInt32(RoundNearestEven)),
			mustConvertToAny(frac.ConvertToInt32(RoundTowardZero)),
			mustConvertToAny(frac.ConvertToInt32(RoundTowardNegative)),
		})
		nearest, nearestFlags := frac.ConvertToInt32(RoundNearestEven)
		unknown, unknownFlags := frac.ConvertToInt32(RoundingMode(99))
		if nearest != unknown || nearestFlags != unknownFlags {
			t.Fatalf("Decimal32 ConvertToInt32 unknown mode = %d/%s, want nearest %d/%s", unknown, unknownFlags, nearest, nearestFlags)
		}
		_, nonExactFlags := frac.ConvertToInt32(RoundTowardZero)
		_, exactFlags := frac.ConvertToInt32Exact(RoundTowardZero)
		if nonExactFlags.HasFlag(FlagInexact) || !exactFlags.HasFlag(FlagInexact) {
			t.Fatalf("Decimal32 inexact flags non-exact=%s exact=%s", nonExactFlags, exactFlags)
		}
		_, invalidFlags := mustDecimal32BID(t, "NaN").ConvertToInt32(RoundNearestEven)
		if !invalidFlags.HasFlag(FlagInvalidOperation) {
			t.Fatalf("Decimal32 NaN ConvertToInt32 flags = %s, want Invalid", invalidFlags)
		}
	})
	t.Run("decimal64", func(t *testing.T) {
		frac := mustDecimal64BID(t, "1.5")
		assertIntConvModeDifference(t, []any{
			mustConvertToAny(frac.ConvertToInt32(RoundNearestEven)),
			mustConvertToAny(frac.ConvertToInt32(RoundTowardZero)),
			mustConvertToAny(frac.ConvertToInt32(RoundTowardNegative)),
		})
		nearest, nearestFlags := frac.ConvertToInt32(RoundNearestEven)
		unknown, unknownFlags := frac.ConvertToInt32(RoundingMode(99))
		if nearest != unknown || nearestFlags != unknownFlags {
			t.Fatalf("Decimal64 ConvertToInt32 unknown mode = %d/%s, want nearest %d/%s", unknown, unknownFlags, nearest, nearestFlags)
		}
		_, nonExactFlags := frac.ConvertToInt32(RoundTowardZero)
		_, exactFlags := frac.ConvertToInt32Exact(RoundTowardZero)
		if nonExactFlags.HasFlag(FlagInexact) || !exactFlags.HasFlag(FlagInexact) {
			t.Fatalf("Decimal64 inexact flags non-exact=%s exact=%s", nonExactFlags, exactFlags)
		}
		_, invalidFlags := mustDecimal64BID(t, "Infinity").ConvertToInt32(RoundNearestEven)
		if !invalidFlags.HasFlag(FlagInvalidOperation) {
			t.Fatalf("Decimal64 Infinity ConvertToInt32 flags = %s, want Invalid", invalidFlags)
		}
	})
	t.Run("decimal128", func(t *testing.T) {
		frac := mustDecimal128BID(t, "1.5")
		assertIntConvModeDifference(t, []any{
			mustConvertToAny(frac.ConvertToInt32(RoundNearestEven)),
			mustConvertToAny(frac.ConvertToInt32(RoundTowardZero)),
			mustConvertToAny(frac.ConvertToInt32(RoundTowardNegative)),
		})
		nearest, nearestFlags := frac.ConvertToInt32(RoundNearestEven)
		unknown, unknownFlags := frac.ConvertToInt32(RoundingMode(99))
		if nearest != unknown || nearestFlags != unknownFlags {
			t.Fatalf("Decimal128 ConvertToInt32 unknown mode = %d/%s, want nearest %d/%s", unknown, unknownFlags, nearest, nearestFlags)
		}
		_, nonExactFlags := frac.ConvertToInt32(RoundTowardZero)
		_, exactFlags := frac.ConvertToInt32Exact(RoundTowardZero)
		if nonExactFlags.HasFlag(FlagInexact) || !exactFlags.HasFlag(FlagInexact) {
			t.Fatalf("Decimal128 inexact flags non-exact=%s exact=%s", nonExactFlags, exactFlags)
		}
		_, invalidFlags := mustDecimal128BID(t, "NaN").ConvertToInt32(RoundNearestEven)
		if !invalidFlags.HasFlag(FlagInvalidOperation) {
			t.Fatalf("Decimal128 NaN ConvertToInt32 flags = %s, want Invalid", invalidFlags)
		}
	})
}

func mustConvertToAny[T comparable](value T, _ ExceptionFlags) any {
	return value
}

func assertIntConvModeDifference(t *testing.T, values []any) {
	t.Helper()
	seen := map[any]bool{}
	for _, value := range values {
		seen[value] = true
	}
	if len(seen) < 2 {
		t.Fatalf("mode switch test did not produce distinct values: %v", values)
	}
}

func testBidgoRoundingMode(mode RoundingMode) int {
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
		return bidgoRoundingNearestEven
	}
}

func TestFromIntPublicRouting(t *testing.T) {
	modes := []RoundingMode{RoundNearestEven, RoundNearestAway, RoundTowardZero, RoundTowardPositive, RoundTowardNegative, RoundingMode(99)}
	for _, mode := range modes {
		t.Run(mode.String(), func(t *testing.T) {
			bidMode := testBidgoRoundingMode(mode)

			got32i32, got32i32Flags := NewDecimal32FromInt32(123456789, mode)
			want32i32, want32i32Flags := bidgo.Bid32FromInt32(123456789, bidMode)
			assertDecimal32Bits(t, "NewDecimal32FromInt32", got32i32, got32i32Flags, Decimal32BID(want32i32), bidgoExceptionFlags(want32i32Flags))

			got32u32, got32u32Flags := NewDecimal32FromUint32(4000000000, mode)
			want32u32, want32u32Flags := bidgo.Bid32FromUint32(4000000000, bidMode)
			assertDecimal32Bits(t, "NewDecimal32FromUint32", got32u32, got32u32Flags, Decimal32BID(want32u32), bidgoExceptionFlags(want32u32Flags))

			got32i64, got32i64Flags := NewDecimal32FromInt64(9223372036854775807, mode)
			want32i64, want32i64Flags := bidgo.Bid32FromInt64(9223372036854775807, bidMode)
			assertDecimal32Bits(t, "NewDecimal32FromInt64", got32i64, got32i64Flags, Decimal32BID(want32i64), bidgoExceptionFlags(want32i64Flags))

			got32u64, got32u64Flags := NewDecimal32FromUint64(^uint64(0), mode)
			want32u64, want32u64Flags := bidgo.Bid32FromUint64(^uint64(0), bidMode)
			assertDecimal32Bits(t, "NewDecimal32FromUint64", got32u64, got32u64Flags, Decimal32BID(want32u64), bidgoExceptionFlags(want32u64Flags))

			got64i64, got64i64Flags := NewDecimal64FromInt64(9223372036854775807, mode)
			want64i64, want64i64Flags := bidgo.Bid64FromInt64(9223372036854775807, bidMode)
			assertDecimal64Bits(t, "NewDecimal64FromInt64", got64i64, got64i64Flags, Decimal64BID(want64i64), bidgoExceptionFlags(want64i64Flags))

			got64u64, got64u64Flags := NewDecimal64FromUint64(^uint64(0), mode)
			want64u64, want64u64Flags := bidgo.Bid64FromUint64(^uint64(0), bidMode)
			assertDecimal64Bits(t, "NewDecimal64FromUint64", got64u64, got64u64Flags, Decimal64BID(want64u64), bidgoExceptionFlags(want64u64Flags))
		})
	}

	if got, want := NewDecimal64FromInt32(-123456789), Decimal64BID(bidgo.Bid64FromInt32(-123456789)); got != want {
		t.Fatalf("NewDecimal64FromInt32 bits = %016x, want %016x", got.ToUint64(), want.ToUint64())
	}
	if got, want := NewDecimal64FromUint32(4000000000), Decimal64BID(bidgo.Bid64FromUint32(4000000000)); got != want {
		t.Fatalf("NewDecimal64FromUint32 bits = %016x, want %016x", got.ToUint64(), want.ToUint64())
	}
	if got, want := NewDecimal128FromInt32(-123456789), decimal128BIDFromBidgo(bidgo.Bid128FromInt32(-123456789)); got != want {
		t.Fatalf("NewDecimal128FromInt32 = %s, want %s", got.String(), want.String())
	}
	if got, want := NewDecimal128FromUint32(4000000000), decimal128BIDFromBidgo(bidgo.Bid128FromUint32(4000000000)); got != want {
		t.Fatalf("NewDecimal128FromUint32 = %s, want %s", got.String(), want.String())
	}
	if got, want := NewDecimal128FromInt64(9223372036854775807), decimal128BIDFromBidgo(bidgo.Bid128FromInt64(9223372036854775807)); got != want {
		t.Fatalf("NewDecimal128FromInt64 = %s, want %s", got.String(), want.String())
	}
	if got, want := NewDecimal128FromUint64(^uint64(0)), decimal128BIDFromBidgo(bidgo.Bid128FromUint64(^uint64(0))); got != want {
		t.Fatalf("NewDecimal128FromUint64 = %s, want %s", got.String(), want.String())
	}

	_, d32Flags := NewDecimal32FromInt32(123456789, RoundNearestEven)
	if !d32Flags.HasFlag(FlagInexact) {
		t.Fatalf("NewDecimal32FromInt32(123456789) flags = %s, want Inexact", d32Flags)
	}
	_, d64Flags := NewDecimal64FromInt64(9223372036854775807, RoundNearestEven)
	if !d64Flags.HasFlag(FlagInexact) {
		t.Fatalf("NewDecimal64FromInt64(max) flags = %s, want Inexact", d64Flags)
	}

	d32Nearest, _ := NewDecimal32FromInt32(123456789, RoundNearestEven)
	d32Zero, _ := NewDecimal32FromInt32(123456789, RoundTowardZero)
	d32Positive, _ := NewDecimal32FromInt32(123456789, RoundTowardPositive)
	d32Unknown, _ := NewDecimal32FromInt32(123456789, RoundingMode(99))
	if d32Unknown != d32Nearest {
		t.Fatalf("NewDecimal32FromInt32 unknown mode = %08x, want nearest-even %08x", d32Unknown.ToUint32(), d32Nearest.ToUint32())
	}
	assertIntConvModeDifference(t, []any{d32Nearest, d32Zero, d32Positive})

	d64Nearest, _ := NewDecimal64FromUint64(^uint64(0), RoundNearestEven)
	d64Zero, _ := NewDecimal64FromUint64(^uint64(0), RoundTowardZero)
	d64Positive, _ := NewDecimal64FromUint64(^uint64(0), RoundTowardPositive)
	d64Unknown, _ := NewDecimal64FromUint64(^uint64(0), RoundingMode(99))
	if d64Unknown != d64Nearest {
		t.Fatalf("NewDecimal64FromUint64 unknown mode = %016x, want nearest-even %016x", d64Unknown.ToUint64(), d64Nearest.ToUint64())
	}
	assertIntConvModeDifference(t, []any{d64Nearest, d64Zero, d64Positive})
}

func assertDecimal32Bits(t *testing.T, name string, got Decimal32BID, gotFlags ExceptionFlags, want Decimal32BID, wantFlags ExceptionFlags) {
	t.Helper()
	if got != want || gotFlags != wantFlags {
		t.Fatalf("%s = %08x/%s, want %08x/%s", name, got.ToUint32(), gotFlags, want.ToUint32(), wantFlags)
	}
}

func assertDecimal64Bits(t *testing.T, name string, got Decimal64BID, gotFlags ExceptionFlags, want Decimal64BID, wantFlags ExceptionFlags) {
	t.Helper()
	if got != want || gotFlags != wantFlags {
		t.Fatalf("%s = %016x/%s, want %016x/%s", name, got.ToUint64(), gotFlags, want.ToUint64(), wantFlags)
	}
}

func TestConvertToIntegerModesActuallyDiffer(t *testing.T) {
	half := mustDecimal64BID(t, "1.5")
	up, _ := half.ConvertToInt32(RoundTowardPositive)
	down, _ := half.ConvertToInt32(RoundTowardNegative)
	if up == down {
		t.Fatalf("RoundTowardPositive and RoundTowardNegative must differ on 1.5: both %d", up)
	}
	evenHalf := mustDecimal64BID(t, "2.5")
	even, _ := evenHalf.ConvertToInt32(RoundNearestEven)
	away, _ := evenHalf.ConvertToInt32(RoundNearestAway)
	if even == away {
		t.Fatalf("RoundNearestEven and RoundNearestAway must differ on 2.5: both %d", even)
	}
}
