// api_v2.go - high-level convenience API over the BID value types.
package bid754

import (
	"fmt"
	"strings"
)

// ParseDecimal parses s and returns the narrowest BID type whose precision
// holds the literal: Decimal32BID, Decimal64BID, or Decimal128BID.
func ParseDecimal(s string) (interface{}, error) {
	if payloadPrecision, ok := determineNaNPayloadPrecision(s); ok {
		switch {
		case payloadPrecision <= 6:
			return NewDecimal32(s)
		case payloadPrecision <= 15:
			return NewDecimal64(s)
		default:
			return NewDecimal128(s)
		}
	}

	precision := determinePrecisionFromString(s)

	switch {
	case precision <= 7:
		return NewDecimal32(s)
	case precision <= 16:
		return NewDecimal64(s)
	default:
		return NewDecimal128(s)
	}
}

// determinePrecisionFromString reports the significant-digit count s needs.
func determinePrecisionFromString(s string) int {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return 1
	}

	lower := strings.ToLower(trimmed)
	if lower == "inf" || lower == "+inf" || lower == "-inf" ||
		lower == "infinity" || lower == "+infinity" || lower == "-infinity" {
		return 1
	}
	if _, ok := parseBIDNaNLiteral(trimmed); ok {
		return 1
	}

	if idx := strings.IndexAny(trimmed, "eE"); idx >= 0 {
		trimmed = trimmed[:idx]
	}

	var digits strings.Builder
	for _, ch := range trimmed {
		if ch >= '0' && ch <= '9' {
			digits.WriteRune(ch)
		}
	}

	significant := strings.TrimLeft(digits.String(), "0")
	significant = strings.TrimRight(significant, "0")
	if significant == "" {
		return 1
	}
	return len(significant)
}

func determineNaNPayloadPrecision(s string) (int, bool) {
	lit, ok := parseBIDNaNLiteral(s)
	if !ok {
		return 0, false
	}
	if lit.payload == "" {
		return 1, true
	}
	return len(lit.payload), true
}

// NewDecimal32 parses a decimal string literal into a Decimal32BID.
func NewDecimal32(s string) (Decimal32BID, error) {
	return NewDecimal32BIDDirect(s)
}

// NewDecimal32FromInt converts an int32 into a Decimal32BID via its decimal
// string form.
func NewDecimal32FromInt(i int32) (Decimal32BID, error) {
	s := fmt.Sprintf("%d", i)
	return NewDecimal32(s)
}

// NewDecimal32FromInt32 converts int32 to Decimal32BID with the requested rounding mode and returned flags.
// Unknown RoundingMode values are routed as RoundNearestEven.
func NewDecimal32FromInt32(x int32, mode RoundingMode) (Decimal32BID, ExceptionFlags) {
	return decimal32BIDFromInt32Port(x, mode)
}

// NewDecimal32FromUint32 converts uint32 to Decimal32BID with the requested rounding mode and returned flags.
// Unknown RoundingMode values are routed as RoundNearestEven.
func NewDecimal32FromUint32(x uint32, mode RoundingMode) (Decimal32BID, ExceptionFlags) {
	return decimal32BIDFromUint32Port(x, mode)
}

// NewDecimal32FromInt64 converts int64 to Decimal32BID with the requested rounding mode and returned flags.
// Unknown RoundingMode values are routed as RoundNearestEven.
func NewDecimal32FromInt64(x int64, mode RoundingMode) (Decimal32BID, ExceptionFlags) {
	return decimal32BIDFromInt64Port(x, mode)
}

// NewDecimal32FromUint64 converts uint64 to Decimal32BID with the requested rounding mode and returned flags.
// Unknown RoundingMode values are routed as RoundNearestEven.
func NewDecimal32FromUint64(x uint64, mode RoundingMode) (Decimal32BID, ExceptionFlags) {
	return decimal32BIDFromUint64Port(x, mode)
}

// NewDecimal64 parses a decimal string literal into a Decimal64BID.
func NewDecimal64(s string) (Decimal64BID, error) {
	return NewDecimal64BIDDirect(s)
}

// NewDecimal64FromInt converts an int64 into a Decimal64BID via its decimal
// string form.
func NewDecimal64FromInt(i int64) (Decimal64BID, error) {
	s := fmt.Sprintf("%d", i)
	return NewDecimal64(s)
}

// NewDecimal64FromInt32 converts int32 to Decimal64BID exactly.
func NewDecimal64FromInt32(x int32) Decimal64BID {
	return decimal64BIDFromInt32Port(x)
}

// NewDecimal64FromUint32 converts uint32 to Decimal64BID exactly.
func NewDecimal64FromUint32(x uint32) Decimal64BID {
	return decimal64BIDFromUint32Port(x)
}

// NewDecimal64FromInt64 converts int64 to Decimal64BID with the requested rounding mode and returned flags.
// Unknown RoundingMode values are routed as RoundNearestEven.
func NewDecimal64FromInt64(x int64, mode RoundingMode) (Decimal64BID, ExceptionFlags) {
	return decimal64BIDFromInt64Port(x, mode)
}

// NewDecimal64FromUint64 converts uint64 to Decimal64BID with the requested rounding mode and returned flags.
// Unknown RoundingMode values are routed as RoundNearestEven.
func NewDecimal64FromUint64(x uint64, mode RoundingMode) (Decimal64BID, ExceptionFlags) {
	return decimal64BIDFromUint64Port(x, mode)
}

// NewDecimal128 parses a decimal string literal into a Decimal128BID.
func NewDecimal128(s string) (Decimal128BID, error) {
	return NewDecimal128BIDDirect(s)
}

// NewDecimal128FromInt converts an int64 into a Decimal128BID via its decimal
// string form.
func NewDecimal128FromInt(i int64) (Decimal128BID, error) {
	s := fmt.Sprintf("%d", i)
	return NewDecimal128(s)
}

// NewDecimal128FromInt32 converts int32 to Decimal128BID exactly.
func NewDecimal128FromInt32(x int32) Decimal128BID {
	return decimal128BIDFromInt32Port(x)
}

// NewDecimal128FromUint32 converts uint32 to Decimal128BID exactly.
func NewDecimal128FromUint32(x uint32) Decimal128BID {
	return decimal128BIDFromUint32Port(x)
}

// NewDecimal128FromInt64 converts int64 to Decimal128BID exactly.
func NewDecimal128FromInt64(x int64) Decimal128BID {
	return decimal128BIDFromInt64Port(x)
}

// NewDecimal128FromUint64 converts uint64 to Decimal128BID exactly.
func NewDecimal128FromUint64(x uint64) Decimal128BID {
	return decimal128BIDFromUint64Port(x)
}

// IsValidDecimalString reports whether s parses as a decimal literal in any
// of the three BID widths.
func IsValidDecimalString(s string) bool {
	_, err := NewDecimal32BIDDirect(s)
	if err == nil {
		return true
	}

	_, err = NewDecimal64BIDDirect(s)
	if err == nil {
		return true
	}

	_, err = NewDecimal128BIDDirect(s)
	return err == nil
}

// GetRequiredPrecision returns the minimum significant-digit precision the
// decimal literal s requires.
func GetRequiredPrecision(s string) int {
	return determinePrecisionFromString(s)
}

// AddSlice32BID returns the left-to-right sum of values, or zero for an
// empty slice.
func AddSlice32BID(values []Decimal32BID) Decimal32BID {
	if len(values) == 0 {
		zero, _ := NewDecimal32BIDDirect("0")
		return zero
	}

	result := values[0]
	for i := 1; i < len(values); i++ {
		result = result.Add(values[i])
	}
	return result
}

// AddSlice32BIDWithFlags returns the left-to-right sum of values together
// with the union of the exception flags raised by each step.
func AddSlice32BIDWithFlags(values []Decimal32BID) (Decimal32BID, ExceptionFlags) {
	if len(values) == 0 {
		zero, _ := NewDecimal32BIDDirect("0")
		return zero, 0
	}

	result := values[0]
	var flags ExceptionFlags
	for i := 1; i < len(values); i++ {
		var stepFlags ExceptionFlags
		result, stepFlags = result.AddWithFlags(values[i])
		flags |= stepFlags
	}
	return result, flags
}

// AddSlice64BID returns the left-to-right sum of values, or zero for an
// empty slice.
func AddSlice64BID(values []Decimal64BID) Decimal64BID {
	if len(values) == 0 {
		zero, _ := NewDecimal64BIDDirect("0")
		return zero
	}

	result := values[0]
	for i := 1; i < len(values); i++ {
		result = result.Add(values[i])
	}
	return result
}

// AddSlice64BIDWithFlags returns the left-to-right sum of values together
// with the union of the exception flags raised by each step.
func AddSlice64BIDWithFlags(values []Decimal64BID) (Decimal64BID, ExceptionFlags) {
	if len(values) == 0 {
		zero, _ := NewDecimal64BIDDirect("0")
		return zero, 0
	}

	result := values[0]
	var flags ExceptionFlags
	for i := 1; i < len(values); i++ {
		var stepFlags ExceptionFlags
		result, stepFlags = result.AddWithFlags(values[i])
		flags |= stepFlags
	}
	return result, flags
}

// AddSlice128BID returns the left-to-right sum of values, or zero for an
// empty slice.
func AddSlice128BID(values []Decimal128BID) Decimal128BID {
	if len(values) == 0 {
		zero, _ := NewDecimal128BIDDirect("0")
		return zero
	}

	result := values[0]
	for i := 1; i < len(values); i++ {
		result = result.Add(values[i])
	}
	return result
}

// AddSlice128BIDWithFlags returns the left-to-right sum of values together
// with the union of the exception flags raised by each step.
func AddSlice128BIDWithFlags(values []Decimal128BID) (Decimal128BID, ExceptionFlags) {
	if len(values) == 0 {
		zero, _ := NewDecimal128BIDDirect("0")
		return zero, 0
	}

	result := values[0]
	var flags ExceptionFlags
	for i := 1; i < len(values); i++ {
		var stepFlags ExceptionFlags
		result, stepFlags = result.AddWithFlags(values[i])
		flags |= stepFlags
	}
	return result, flags
}

// Common decimal constants in each BID width. Pi and E carry the maximum
// number of significant digits each width can represent.
var (
	Zero32BID, _  = NewDecimal32BIDDirect("0")
	Zero64BID, _  = NewDecimal64BIDDirect("0")
	Zero128BID, _ = NewDecimal128BIDDirect("0")

	One32BID, _  = NewDecimal32BIDDirect("1")
	One64BID, _  = NewDecimal64BIDDirect("1")
	One128BID, _ = NewDecimal128BIDDirect("1")

	Pi32BID, _  = NewDecimal32BIDDirect("3.141593")
	Pi64BID, _  = NewDecimal64BIDDirect("3.141592653589793")
	Pi128BID, _ = NewDecimal128BIDDirect("3.141592653589793238462643383279503")

	E32BID, _  = NewDecimal32BIDDirect("2.718282")
	E64BID, _  = NewDecimal64BIDDirect("2.718281828459045")
	E128BID, _ = NewDecimal128BIDDirect("2.718281828459045235360287471352662")
)
