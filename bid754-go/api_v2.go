// api_v2.go - high-level convenience API over the BID value types.
package bid754

import (
	"fmt"
	"strings"
)

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

// NewDecimal32 parses a decimal string literal exactly into a Decimal32BID.
// It returns an error instead of silently rounding or changing range; use
// NewDecimal32WithFlags when a rounded finite result is desired.
func NewDecimal32(s string) (Decimal32BID, error) {
	return NewDecimal32BIDDirect(s)
}

// NewDecimal32WithFlags parses a decimal string literal into a Decimal32BID
// and also returns the exception flags raised while parsing, e.g. FlagInexact
// when the literal is not exactly representable in 7 digits, or FlagOverflow
// for an out-of-range exponent. Malformed input, a NaN payload that cannot be
// represented in Decimal32, and a written cohort whose quantum or coefficient
// the width would otherwise coerce silently return an error.
func NewDecimal32WithFlags(s string) (Decimal32BID, ExceptionFlags, error) {
	return newDecimal32BIDWithFlagsPort(s)
}

// NewDecimal32WithMode parses s as a Decimal32BID value rounding excess
// precision with mode, and returns the exception flags raised while parsing
// alongside the NewDecimal32WithFlags error contract (a non-nil error if and
// only if the input string is rejected). A RoundingMode outside the defined
// constants raises FlagInvalidOperation and returns a canonical quiet NaN
// with a nil error, instead of panicking: the flag channel is the repo-wide
// invalid-mode failure channel, keeping it structurally distinguishable from
// a string rejection (zero flags, non-nil error). String rejection takes
// precedence when both the string and mode are invalid, so a bad mode cannot
// mask a rejected input string.
func NewDecimal32WithMode(s string, mode RoundingMode) (Decimal32BID, ExceptionFlags, error) {
	return newDecimal32BIDWithModePort(s, mode)
}

// NewDecimal32FromInt converts an int32 exactly into a Decimal32BID. It returns
// a quiet NaN and an error when the integer is not exactly representable; use
// NewDecimal32FromInt32 for an explicitly rounded conversion with flags.
func NewDecimal32FromInt(i int32) (Decimal32BID, error) {
	result, flags := decimal32BIDFromInt32Port(i, RoundNearestEven)
	if flags&FlagInexact != 0 {
		return canonicalQNaN32BID(), fmt.Errorf("integer %d is not exactly representable as Decimal32BID", i)
	}
	return result, nil
}

// NewDecimal32FromInt32 converts int32 to Decimal32BID with the requested rounding mode and returned flags.
// Passing a RoundingMode outside the defined constants raises
// FlagInvalidOperation and returns a canonical quiet NaN, instead of panicking.
func NewDecimal32FromInt32(x int32, mode RoundingMode) (Decimal32BID, ExceptionFlags) {
	return decimal32BIDFromInt32Port(x, mode)
}

// NewDecimal32FromUint32 converts uint32 to Decimal32BID with the requested rounding mode and returned flags.
// Passing a RoundingMode outside the defined constants raises
// FlagInvalidOperation and returns a canonical quiet NaN, instead of panicking.
func NewDecimal32FromUint32(x uint32, mode RoundingMode) (Decimal32BID, ExceptionFlags) {
	return decimal32BIDFromUint32Port(x, mode)
}

// NewDecimal32FromInt64 converts int64 to Decimal32BID with the requested rounding mode and returned flags.
// Passing a RoundingMode outside the defined constants raises
// FlagInvalidOperation and returns a canonical quiet NaN, instead of panicking.
func NewDecimal32FromInt64(x int64, mode RoundingMode) (Decimal32BID, ExceptionFlags) {
	return decimal32BIDFromInt64Port(x, mode)
}

// NewDecimal32FromUint64 converts uint64 to Decimal32BID with the requested rounding mode and returned flags.
// Passing a RoundingMode outside the defined constants raises
// FlagInvalidOperation and returns a canonical quiet NaN, instead of panicking.
func NewDecimal32FromUint64(x uint64, mode RoundingMode) (Decimal32BID, ExceptionFlags) {
	return decimal32BIDFromUint64Port(x, mode)
}

// NewDecimal64 parses a decimal string literal exactly into a Decimal64BID.
// It returns an error instead of silently rounding or changing range; use
// NewDecimal64WithFlags when a rounded finite result is desired.
func NewDecimal64(s string) (Decimal64BID, error) {
	return NewDecimal64BIDDirect(s)
}

// NewDecimal64WithFlags parses a decimal string literal into a Decimal64BID
// and also returns the exception flags raised while parsing, e.g. FlagInexact
// when the literal is not exactly representable in 16 digits, or FlagOverflow
// for an out-of-range exponent. Malformed input, a NaN payload that cannot be
// represented in Decimal64, and a written cohort whose quantum or coefficient
// the width would otherwise coerce silently return an error.
func NewDecimal64WithFlags(s string) (Decimal64BID, ExceptionFlags, error) {
	return newDecimal64BIDWithFlagsPort(s)
}

// NewDecimal64WithMode parses s as a Decimal64BID value rounding excess
// precision with mode, and returns the exception flags raised while parsing
// alongside the NewDecimal64WithFlags error contract (a non-nil error if and
// only if the input string is rejected). A RoundingMode outside the defined
// constants raises FlagInvalidOperation and returns a canonical quiet NaN
// with a nil error, instead of panicking: the flag channel is the repo-wide
// invalid-mode failure channel, keeping it structurally distinguishable from
// a string rejection (zero flags, non-nil error). String rejection takes
// precedence when both inputs are invalid.
func NewDecimal64WithMode(s string, mode RoundingMode) (Decimal64BID, ExceptionFlags, error) {
	return newDecimal64BIDWithModePort(s, mode)
}

// NewDecimal64FromInt converts an int64 exactly into a Decimal64BID. It returns
// a quiet NaN and an error when the integer is not exactly representable; use
// NewDecimal64FromInt64 for an explicitly rounded conversion with flags.
func NewDecimal64FromInt(i int64) (Decimal64BID, error) {
	result, flags := decimal64BIDFromInt64Port(i, RoundNearestEven)
	if flags&FlagInexact != 0 {
		return canonicalQNaN64BID(), fmt.Errorf("integer %d is not exactly representable as Decimal64BID", i)
	}
	return result, nil
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
// Passing a RoundingMode outside the defined constants raises
// FlagInvalidOperation and returns a canonical quiet NaN, instead of panicking.
func NewDecimal64FromInt64(x int64, mode RoundingMode) (Decimal64BID, ExceptionFlags) {
	return decimal64BIDFromInt64Port(x, mode)
}

// NewDecimal64FromUint64 converts uint64 to Decimal64BID with the requested rounding mode and returned flags.
// Passing a RoundingMode outside the defined constants raises
// FlagInvalidOperation and returns a canonical quiet NaN, instead of panicking.
func NewDecimal64FromUint64(x uint64, mode RoundingMode) (Decimal64BID, ExceptionFlags) {
	return decimal64BIDFromUint64Port(x, mode)
}

// NewDecimal128 parses a decimal string literal exactly into a Decimal128BID.
// It returns an error instead of silently rounding or changing range; use
// NewDecimal128WithFlags when a rounded finite result is desired.
func NewDecimal128(s string) (Decimal128BID, error) {
	return NewDecimal128BIDDirect(s)
}

// NewDecimal128WithFlags parses a decimal string literal into a Decimal128BID
// and also returns the exception flags raised while parsing, e.g. FlagInexact
// when the literal is not exactly representable in 34 digits, or FlagOverflow
// for an out-of-range exponent. Malformed input, a NaN payload that cannot be
// represented in Decimal128, and a written cohort whose quantum or coefficient
// the width would otherwise coerce silently return an error.
func NewDecimal128WithFlags(s string) (Decimal128BID, ExceptionFlags, error) {
	return newDecimal128BIDWithFlagsPort(s)
}

// NewDecimal128WithMode parses s as a Decimal128BID value rounding excess
// precision with mode, and returns the exception flags raised while parsing
// alongside the NewDecimal128WithFlags error contract (a non-nil error if and
// only if the input string is rejected). A RoundingMode outside the defined
// constants raises FlagInvalidOperation and returns a canonical quiet NaN
// with a nil error, instead of panicking: the flag channel is the repo-wide
// invalid-mode failure channel, keeping it structurally distinguishable from
// a string rejection (zero flags, non-nil error). String rejection takes
// precedence when both inputs are invalid.
func NewDecimal128WithMode(s string, mode RoundingMode) (Decimal128BID, ExceptionFlags, error) {
	return newDecimal128BIDWithModePort(s, mode)
}

// NewDecimal128FromInt converts an int64 exactly into a Decimal128BID.
func NewDecimal128FromInt(i int64) (Decimal128BID, error) {
	return decimal128BIDFromInt64Port(i), nil
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

// IsValidDecimalString reports whether s is a complete decimal literal that
// is exactly representable in at least one of the three BID widths.
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
// decimal literal s requires. Trailing zeros do not raise the requirement:
// "1.2300000" needs 3 significant digits. This is the canonical minimal
// precision of the value, not the literal's written digit count. This is a
// syntax/precision query, so a finite literal may exceed every supported
// width's exponent range; malformed input returns an error.
func GetRequiredPrecision(s string) (int, error) {
	_, nanLiteral := parseBIDNaNLiteral(s)
	if !nanLiteral && !validBIDFiniteLiteral(s) {
		return 0, fmt.Errorf("invalid decimal string: %q", s)
	}
	return determinePrecisionFromString(s), nil
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

// Common decimal values in each BID width. Pi and E carry the maximum number
// of significant digits each width can represent. The backing values are not
// exported: accessors return copies so callers cannot replace the process-wide
// value or introduce a data race by reassigning it.
var (
	zero32BID, _  = NewDecimal32BIDDirect("0")
	zero64BID, _  = NewDecimal64BIDDirect("0")
	zero128BID, _ = NewDecimal128BIDDirect("0")

	one32BID, _  = NewDecimal32BIDDirect("1")
	one64BID, _  = NewDecimal64BIDDirect("1")
	one128BID, _ = NewDecimal128BIDDirect("1")

	pi32BID, _  = NewDecimal32BIDDirect("3.141593")
	pi64BID, _  = NewDecimal64BIDDirect("3.141592653589793")
	pi128BID, _ = NewDecimal128BIDDirect("3.141592653589793238462643383279503")

	e32BID, _  = NewDecimal32BIDDirect("2.718282")
	e64BID, _  = NewDecimal64BIDDirect("2.718281828459045")
	e128BID, _ = NewDecimal128BIDDirect("2.718281828459045235360287471352662")
)

// Zero32BID returns the Decimal32BID representation of zero.
func Zero32BID() Decimal32BID { return zero32BID }

// Zero64BID returns the Decimal64BID representation of zero.
func Zero64BID() Decimal64BID { return zero64BID }

// Zero128BID returns the Decimal128BID representation of zero.
func Zero128BID() Decimal128BID { return zero128BID }

// One32BID returns the Decimal32BID representation of one.
func One32BID() Decimal32BID { return one32BID }

// One64BID returns the Decimal64BID representation of one.
func One64BID() Decimal64BID { return one64BID }

// One128BID returns the Decimal128BID representation of one.
func One128BID() Decimal128BID { return one128BID }

// Pi32BID returns pi rounded to Decimal32BID precision.
func Pi32BID() Decimal32BID { return pi32BID }

// Pi64BID returns pi rounded to Decimal64BID precision.
func Pi64BID() Decimal64BID { return pi64BID }

// Pi128BID returns pi rounded to Decimal128BID precision.
func Pi128BID() Decimal128BID { return pi128BID }

// E32BID returns Euler's number rounded to Decimal32BID precision.
func E32BID() Decimal32BID { return e32BID }

// E64BID returns Euler's number rounded to Decimal64BID precision.
func E64BID() Decimal64BID { return e64BID }

// E128BID returns Euler's number rounded to Decimal128BID precision.
func E128BID() Decimal128BID { return e128BID }
