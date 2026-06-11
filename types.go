package bid754

import (
	"strings"
)

// String returns the encoding format name.
func (f EncodingFormat) String() string {
	switch f {
	case EncodingBID:
		return "BID"
	default:
		return "Unknown"
	}
}

// String returns the rounding mode name.
func (r RoundingMode) String() string {
	switch r {
	case RoundNearestEven:
		return "RoundNearestEven"
	case RoundNearestAway:
		return "RoundNearestAway"
	case RoundTowardZero:
		return "RoundTowardZero"
	case RoundTowardPositive:
		return "RoundTowardPositive"
	case RoundTowardNegative:
		return "RoundTowardNegative"
	default:
		return "Unknown"
	}
}

// String returns the raised flag names joined by "|", or "None".
func (f ExceptionFlags) String() string {
	if f == 0 {
		return "None"
	}

	var flags []string
	if f&FlagInexact != 0 {
		flags = append(flags, "Inexact")
	}
	if f&FlagUnderflow != 0 {
		flags = append(flags, "Underflow")
	}
	if f&FlagOverflow != 0 {
		flags = append(flags, "Overflow")
	}
	if f&FlagDivisionByZero != 0 {
		flags = append(flags, "DivisionByZero")
	}
	if f&FlagInvalidOperation != 0 {
		flags = append(flags, "InvalidOperation")
	}
	if f&FlagSubnormal != 0 {
		flags = append(flags, "Subnormal")
	}
	if f&FlagRounded != 0 {
		flags = append(flags, "Rounded")
	}
	if f&FlagClamped != 0 {
		flags = append(flags, "Clamped")
	}

	return strings.Join(flags, "|")
}

// HasFlag reports whether any of the given flags is set in this saved flag
// value (IEEE 754-2019 5.7.4 testSavedFlags).
func (f ExceptionFlags) HasFlag(flag ExceptionFlags) bool {
	return f&flag != 0
}

// Common interfaces.

// Decimal is the base interface implemented by every decimal value type.
type Decimal interface {
	String() string
	IsZero() bool
	IsNaN() bool
	IsInf() bool
	IsSignMinus() bool
	Sign() int
}

// Decimal32Interface is implemented by 32-bit decimal value types.
type Decimal32Interface interface {
	Decimal
	ToUint32() uint32
}

// Decimal64Interface is implemented by 64-bit decimal value types.
type Decimal64Interface interface {
	Decimal
	ToUint64() uint64
}

// Decimal128Interface is implemented by 128-bit decimal value types.
type Decimal128Interface interface {
	Decimal
	ToBytes() [16]byte
}

// Binary128 is a fixed-width IEEE 754 binary128 bit pattern.
type Binary128 [16]byte

// ToBytes returns the binary128 bit pattern as 16 bytes.
func (b Binary128) ToBytes() [16]byte {
	return [16]byte(b)
}

// Note: the BID public value-type surface routes through the Go mechanical
// port via the types_bidgo_runtime.go helpers.
