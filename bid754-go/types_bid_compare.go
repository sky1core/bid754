package bid754

// QuietEqual reports whether d == other for Decimal32BID under the IEEE 754 quiet comparison predicate, without invalid for quiet NaNs, and returns the exception flags raised by the operation.
func (d Decimal32BID) QuietEqual(other Decimal32BID) (bool, ExceptionFlags) {
	return decimal32BIDQuietEqualPort(d, other)
}

// QuietNotEqual reports whether d != other for Decimal32BID under the IEEE 754 quiet comparison predicate, without invalid for quiet NaNs, and returns the exception flags raised by the operation.
func (d Decimal32BID) QuietNotEqual(other Decimal32BID) (bool, ExceptionFlags) {
	return decimal32BIDQuietNotEqualPort(d, other)
}

// QuietGreater reports whether d > other for Decimal32BID under the IEEE 754 quiet comparison predicate, without invalid for quiet NaNs, and returns the exception flags raised by the operation.
func (d Decimal32BID) QuietGreater(other Decimal32BID) (bool, ExceptionFlags) {
	return decimal32BIDQuietGreaterPort(d, other)
}

// QuietGreaterEqual reports whether d >= other for Decimal32BID under the IEEE 754 quiet comparison predicate, without invalid for quiet NaNs, and returns the exception flags raised by the operation.
func (d Decimal32BID) QuietGreaterEqual(other Decimal32BID) (bool, ExceptionFlags) {
	return decimal32BIDQuietGreaterEqualPort(d, other)
}

// QuietGreaterUnordered reports whether d > other or the operands are unordered for Decimal32BID under the IEEE 754 quiet comparison predicate, without invalid for quiet NaNs, and returns the exception flags raised by the operation.
func (d Decimal32BID) QuietGreaterUnordered(other Decimal32BID) (bool, ExceptionFlags) {
	return decimal32BIDQuietGreaterUnorderedPort(d, other)
}

// QuietLess reports whether d < other for Decimal32BID under the IEEE 754 quiet comparison predicate, without invalid for quiet NaNs, and returns the exception flags raised by the operation.
func (d Decimal32BID) QuietLess(other Decimal32BID) (bool, ExceptionFlags) {
	return decimal32BIDQuietLessPort(d, other)
}

// QuietLessEqual reports whether d <= other for Decimal32BID under the IEEE 754 quiet comparison predicate, without invalid for quiet NaNs, and returns the exception flags raised by the operation.
func (d Decimal32BID) QuietLessEqual(other Decimal32BID) (bool, ExceptionFlags) {
	return decimal32BIDQuietLessEqualPort(d, other)
}

// QuietLessUnordered reports whether d < other or the operands are unordered for Decimal32BID under the IEEE 754 quiet comparison predicate, without invalid for quiet NaNs, and returns the exception flags raised by the operation.
func (d Decimal32BID) QuietLessUnordered(other Decimal32BID) (bool, ExceptionFlags) {
	return decimal32BIDQuietLessUnorderedPort(d, other)
}

// QuietNotGreater reports whether d is not greater than other for Decimal32BID under the IEEE 754 quiet comparison predicate, without invalid for quiet NaNs, and returns the exception flags raised by the operation.
func (d Decimal32BID) QuietNotGreater(other Decimal32BID) (bool, ExceptionFlags) {
	return decimal32BIDQuietNotGreaterPort(d, other)
}

// QuietNotLess reports whether d is not less than other for Decimal32BID under the IEEE 754 quiet comparison predicate, without invalid for quiet NaNs, and returns the exception flags raised by the operation.
func (d Decimal32BID) QuietNotLess(other Decimal32BID) (bool, ExceptionFlags) {
	return decimal32BIDQuietNotLessPort(d, other)
}

// QuietOrdered reports whether d and other are ordered Decimal32BID operands under the IEEE 754 quiet comparison predicate, without invalid for quiet NaNs, and returns the exception flags raised by the operation.
func (d Decimal32BID) QuietOrdered(other Decimal32BID) (bool, ExceptionFlags) {
	return decimal32BIDQuietOrderedPort(d, other)
}

// QuietUnordered reports whether d and other are unordered Decimal32BID operands under the IEEE 754 quiet comparison predicate, without invalid for quiet NaNs, and returns the exception flags raised by the operation.
func (d Decimal32BID) QuietUnordered(other Decimal32BID) (bool, ExceptionFlags) {
	return decimal32BIDQuietUnorderedPort(d, other)
}

// SignalingEqual derives equality from signaling GreaterEqual and LessEqual port calls.
// Returned flags are the OR of those port-returned flags.
func (d Decimal32BID) SignalingEqual(other Decimal32BID) (bool, ExceptionFlags) {
	return decimal32BIDSignalingEqualPort(d, other)
}

// SignalingNotEqual is the negation of SignalingEqual with the same port-returned flags.
func (d Decimal32BID) SignalingNotEqual(other Decimal32BID) (bool, ExceptionFlags) {
	return decimal32BIDSignalingNotEqualPort(d, other)
}

// SignalingGreater reports whether d > other for Decimal32BID under the IEEE 754 signaling comparison predicate, raising invalid for NaN comparisons, and returns the exception flags raised by the operation.
func (d Decimal32BID) SignalingGreater(other Decimal32BID) (bool, ExceptionFlags) {
	return decimal32BIDSignalingGreaterPort(d, other)
}

// SignalingGreaterEqual reports whether d >= other for Decimal32BID under the IEEE 754 signaling comparison predicate, raising invalid for NaN comparisons, and returns the exception flags raised by the operation.
func (d Decimal32BID) SignalingGreaterEqual(other Decimal32BID) (bool, ExceptionFlags) {
	return decimal32BIDSignalingGreaterEqualPort(d, other)
}

// SignalingGreaterUnordered reports whether d > other or the operands are unordered for Decimal32BID under the IEEE 754 signaling comparison predicate, raising invalid for NaN comparisons, and returns the exception flags raised by the operation.
func (d Decimal32BID) SignalingGreaterUnordered(other Decimal32BID) (bool, ExceptionFlags) {
	return decimal32BIDSignalingGreaterUnorderedPort(d, other)
}

// SignalingLess reports whether d < other for Decimal32BID under the IEEE 754 signaling comparison predicate, raising invalid for NaN comparisons, and returns the exception flags raised by the operation.
func (d Decimal32BID) SignalingLess(other Decimal32BID) (bool, ExceptionFlags) {
	return decimal32BIDSignalingLessPort(d, other)
}

// SignalingLessEqual reports whether d <= other for Decimal32BID under the IEEE 754 signaling comparison predicate, raising invalid for NaN comparisons, and returns the exception flags raised by the operation.
func (d Decimal32BID) SignalingLessEqual(other Decimal32BID) (bool, ExceptionFlags) {
	return decimal32BIDSignalingLessEqualPort(d, other)
}

// SignalingLessUnordered reports whether d < other or the operands are unordered for Decimal32BID under the IEEE 754 signaling comparison predicate, raising invalid for NaN comparisons, and returns the exception flags raised by the operation.
func (d Decimal32BID) SignalingLessUnordered(other Decimal32BID) (bool, ExceptionFlags) {
	return decimal32BIDSignalingLessUnorderedPort(d, other)
}

// SignalingNotGreater reports whether d is not greater than other for Decimal32BID under the IEEE 754 signaling comparison predicate, raising invalid for NaN comparisons, and returns the exception flags raised by the operation.
func (d Decimal32BID) SignalingNotGreater(other Decimal32BID) (bool, ExceptionFlags) {
	return decimal32BIDSignalingNotGreaterPort(d, other)
}

// SignalingNotLess reports whether d is not less than other for Decimal32BID under the IEEE 754 signaling comparison predicate, raising invalid for NaN comparisons, and returns the exception flags raised by the operation.
func (d Decimal32BID) SignalingNotLess(other Decimal32BID) (bool, ExceptionFlags) {
	return decimal32BIDSignalingNotLessPort(d, other)
}

// QuietEqual reports whether d == other for Decimal64BID under the IEEE 754 quiet comparison predicate, without invalid for quiet NaNs, and returns the exception flags raised by the operation.
func (d Decimal64BID) QuietEqual(other Decimal64BID) (bool, ExceptionFlags) {
	return decimal64BIDQuietEqualPort(d, other)
}

// QuietNotEqual reports whether d != other for Decimal64BID under the IEEE 754 quiet comparison predicate, without invalid for quiet NaNs, and returns the exception flags raised by the operation.
func (d Decimal64BID) QuietNotEqual(other Decimal64BID) (bool, ExceptionFlags) {
	return decimal64BIDQuietNotEqualPort(d, other)
}

// QuietGreater reports whether d > other for Decimal64BID under the IEEE 754 quiet comparison predicate, without invalid for quiet NaNs, and returns the exception flags raised by the operation.
func (d Decimal64BID) QuietGreater(other Decimal64BID) (bool, ExceptionFlags) {
	return decimal64BIDQuietGreaterPort(d, other)
}

// QuietGreaterEqual reports whether d >= other for Decimal64BID under the IEEE 754 quiet comparison predicate, without invalid for quiet NaNs, and returns the exception flags raised by the operation.
func (d Decimal64BID) QuietGreaterEqual(other Decimal64BID) (bool, ExceptionFlags) {
	return decimal64BIDQuietGreaterEqualPort(d, other)
}

// QuietGreaterUnordered reports whether d > other or the operands are unordered for Decimal64BID under the IEEE 754 quiet comparison predicate, without invalid for quiet NaNs, and returns the exception flags raised by the operation.
func (d Decimal64BID) QuietGreaterUnordered(other Decimal64BID) (bool, ExceptionFlags) {
	return decimal64BIDQuietGreaterUnorderedPort(d, other)
}

// QuietLess reports whether d < other for Decimal64BID under the IEEE 754 quiet comparison predicate, without invalid for quiet NaNs, and returns the exception flags raised by the operation.
func (d Decimal64BID) QuietLess(other Decimal64BID) (bool, ExceptionFlags) {
	return decimal64BIDQuietLessPort(d, other)
}

// QuietLessEqual reports whether d <= other for Decimal64BID under the IEEE 754 quiet comparison predicate, without invalid for quiet NaNs, and returns the exception flags raised by the operation.
func (d Decimal64BID) QuietLessEqual(other Decimal64BID) (bool, ExceptionFlags) {
	return decimal64BIDQuietLessEqualPort(d, other)
}

// QuietLessUnordered reports whether d < other or the operands are unordered for Decimal64BID under the IEEE 754 quiet comparison predicate, without invalid for quiet NaNs, and returns the exception flags raised by the operation.
func (d Decimal64BID) QuietLessUnordered(other Decimal64BID) (bool, ExceptionFlags) {
	return decimal64BIDQuietLessUnorderedPort(d, other)
}

// QuietNotGreater reports whether d is not greater than other for Decimal64BID under the IEEE 754 quiet comparison predicate, without invalid for quiet NaNs, and returns the exception flags raised by the operation.
func (d Decimal64BID) QuietNotGreater(other Decimal64BID) (bool, ExceptionFlags) {
	return decimal64BIDQuietNotGreaterPort(d, other)
}

// QuietNotLess reports whether d is not less than other for Decimal64BID under the IEEE 754 quiet comparison predicate, without invalid for quiet NaNs, and returns the exception flags raised by the operation.
func (d Decimal64BID) QuietNotLess(other Decimal64BID) (bool, ExceptionFlags) {
	return decimal64BIDQuietNotLessPort(d, other)
}

// QuietOrdered reports whether d and other are ordered Decimal64BID operands under the IEEE 754 quiet comparison predicate, without invalid for quiet NaNs, and returns the exception flags raised by the operation.
func (d Decimal64BID) QuietOrdered(other Decimal64BID) (bool, ExceptionFlags) {
	return decimal64BIDQuietOrderedPort(d, other)
}

// QuietUnordered reports whether d and other are unordered Decimal64BID operands under the IEEE 754 quiet comparison predicate, without invalid for quiet NaNs, and returns the exception flags raised by the operation.
func (d Decimal64BID) QuietUnordered(other Decimal64BID) (bool, ExceptionFlags) {
	return decimal64BIDQuietUnorderedPort(d, other)
}

// SignalingEqual derives equality from signaling GreaterEqual and LessEqual port calls.
// Returned flags are the OR of those port-returned flags.
func (d Decimal64BID) SignalingEqual(other Decimal64BID) (bool, ExceptionFlags) {
	return decimal64BIDSignalingEqualPort(d, other)
}

// SignalingNotEqual is the negation of SignalingEqual with the same port-returned flags.
func (d Decimal64BID) SignalingNotEqual(other Decimal64BID) (bool, ExceptionFlags) {
	return decimal64BIDSignalingNotEqualPort(d, other)
}

// SignalingGreater reports whether d > other for Decimal64BID under the IEEE 754 signaling comparison predicate, raising invalid for NaN comparisons, and returns the exception flags raised by the operation.
func (d Decimal64BID) SignalingGreater(other Decimal64BID) (bool, ExceptionFlags) {
	return decimal64BIDSignalingGreaterPort(d, other)
}

// SignalingGreaterEqual reports whether d >= other for Decimal64BID under the IEEE 754 signaling comparison predicate, raising invalid for NaN comparisons, and returns the exception flags raised by the operation.
func (d Decimal64BID) SignalingGreaterEqual(other Decimal64BID) (bool, ExceptionFlags) {
	return decimal64BIDSignalingGreaterEqualPort(d, other)
}

// SignalingGreaterUnordered reports whether d > other or the operands are unordered for Decimal64BID under the IEEE 754 signaling comparison predicate, raising invalid for NaN comparisons, and returns the exception flags raised by the operation.
func (d Decimal64BID) SignalingGreaterUnordered(other Decimal64BID) (bool, ExceptionFlags) {
	return decimal64BIDSignalingGreaterUnorderedPort(d, other)
}

// SignalingLess reports whether d < other for Decimal64BID under the IEEE 754 signaling comparison predicate, raising invalid for NaN comparisons, and returns the exception flags raised by the operation.
func (d Decimal64BID) SignalingLess(other Decimal64BID) (bool, ExceptionFlags) {
	return decimal64BIDSignalingLessPort(d, other)
}

// SignalingLessEqual reports whether d <= other for Decimal64BID under the IEEE 754 signaling comparison predicate, raising invalid for NaN comparisons, and returns the exception flags raised by the operation.
func (d Decimal64BID) SignalingLessEqual(other Decimal64BID) (bool, ExceptionFlags) {
	return decimal64BIDSignalingLessEqualPort(d, other)
}

// SignalingLessUnordered reports whether d < other or the operands are unordered for Decimal64BID under the IEEE 754 signaling comparison predicate, raising invalid for NaN comparisons, and returns the exception flags raised by the operation.
func (d Decimal64BID) SignalingLessUnordered(other Decimal64BID) (bool, ExceptionFlags) {
	return decimal64BIDSignalingLessUnorderedPort(d, other)
}

// SignalingNotGreater reports whether d is not greater than other for Decimal64BID under the IEEE 754 signaling comparison predicate, raising invalid for NaN comparisons, and returns the exception flags raised by the operation.
func (d Decimal64BID) SignalingNotGreater(other Decimal64BID) (bool, ExceptionFlags) {
	return decimal64BIDSignalingNotGreaterPort(d, other)
}

// SignalingNotLess reports whether d is not less than other for Decimal64BID under the IEEE 754 signaling comparison predicate, raising invalid for NaN comparisons, and returns the exception flags raised by the operation.
func (d Decimal64BID) SignalingNotLess(other Decimal64BID) (bool, ExceptionFlags) {
	return decimal64BIDSignalingNotLessPort(d, other)
}

// QuietEqual reports whether d == other for Decimal128BID under the IEEE 754 quiet comparison predicate, without invalid for quiet NaNs, and returns the exception flags raised by the operation.
func (d Decimal128BID) QuietEqual(other Decimal128BID) (bool, ExceptionFlags) {
	return decimal128BIDQuietEqualPort(d, other)
}

// QuietNotEqual reports whether d != other for Decimal128BID under the IEEE 754 quiet comparison predicate, without invalid for quiet NaNs, and returns the exception flags raised by the operation.
func (d Decimal128BID) QuietNotEqual(other Decimal128BID) (bool, ExceptionFlags) {
	return decimal128BIDQuietNotEqualPort(d, other)
}

// QuietGreater reports whether d > other for Decimal128BID under the IEEE 754 quiet comparison predicate, without invalid for quiet NaNs, and returns the exception flags raised by the operation.
func (d Decimal128BID) QuietGreater(other Decimal128BID) (bool, ExceptionFlags) {
	return decimal128BIDQuietGreaterPort(d, other)
}

// QuietGreaterEqual reports whether d >= other for Decimal128BID under the IEEE 754 quiet comparison predicate, without invalid for quiet NaNs, and returns the exception flags raised by the operation.
func (d Decimal128BID) QuietGreaterEqual(other Decimal128BID) (bool, ExceptionFlags) {
	return decimal128BIDQuietGreaterEqualPort(d, other)
}

// QuietGreaterUnordered reports whether d > other or the operands are unordered for Decimal128BID under the IEEE 754 quiet comparison predicate, without invalid for quiet NaNs, and returns the exception flags raised by the operation.
func (d Decimal128BID) QuietGreaterUnordered(other Decimal128BID) (bool, ExceptionFlags) {
	return decimal128BIDQuietGreaterUnorderedPort(d, other)
}

// QuietLess reports whether d < other for Decimal128BID under the IEEE 754 quiet comparison predicate, without invalid for quiet NaNs, and returns the exception flags raised by the operation.
func (d Decimal128BID) QuietLess(other Decimal128BID) (bool, ExceptionFlags) {
	return decimal128BIDQuietLessPort(d, other)
}

// QuietLessEqual reports whether d <= other for Decimal128BID under the IEEE 754 quiet comparison predicate, without invalid for quiet NaNs, and returns the exception flags raised by the operation.
func (d Decimal128BID) QuietLessEqual(other Decimal128BID) (bool, ExceptionFlags) {
	return decimal128BIDQuietLessEqualPort(d, other)
}

// QuietLessUnordered reports whether d < other or the operands are unordered for Decimal128BID under the IEEE 754 quiet comparison predicate, without invalid for quiet NaNs, and returns the exception flags raised by the operation.
func (d Decimal128BID) QuietLessUnordered(other Decimal128BID) (bool, ExceptionFlags) {
	return decimal128BIDQuietLessUnorderedPort(d, other)
}

// QuietNotGreater reports whether d is not greater than other for Decimal128BID under the IEEE 754 quiet comparison predicate, without invalid for quiet NaNs, and returns the exception flags raised by the operation.
func (d Decimal128BID) QuietNotGreater(other Decimal128BID) (bool, ExceptionFlags) {
	return decimal128BIDQuietNotGreaterPort(d, other)
}

// QuietNotLess reports whether d is not less than other for Decimal128BID under the IEEE 754 quiet comparison predicate, without invalid for quiet NaNs, and returns the exception flags raised by the operation.
func (d Decimal128BID) QuietNotLess(other Decimal128BID) (bool, ExceptionFlags) {
	return decimal128BIDQuietNotLessPort(d, other)
}

// QuietOrdered reports whether d and other are ordered Decimal128BID operands under the IEEE 754 quiet comparison predicate, without invalid for quiet NaNs, and returns the exception flags raised by the operation.
func (d Decimal128BID) QuietOrdered(other Decimal128BID) (bool, ExceptionFlags) {
	return decimal128BIDQuietOrderedPort(d, other)
}

// QuietUnordered reports whether d and other are unordered Decimal128BID operands under the IEEE 754 quiet comparison predicate, without invalid for quiet NaNs, and returns the exception flags raised by the operation.
func (d Decimal128BID) QuietUnordered(other Decimal128BID) (bool, ExceptionFlags) {
	return decimal128BIDQuietUnorderedPort(d, other)
}

// SignalingEqual derives equality from signaling GreaterEqual and LessEqual port calls.
// Returned flags are the OR of those port-returned flags.
func (d Decimal128BID) SignalingEqual(other Decimal128BID) (bool, ExceptionFlags) {
	return decimal128BIDSignalingEqualPort(d, other)
}

// SignalingNotEqual is the negation of SignalingEqual with the same port-returned flags.
func (d Decimal128BID) SignalingNotEqual(other Decimal128BID) (bool, ExceptionFlags) {
	return decimal128BIDSignalingNotEqualPort(d, other)
}

// SignalingGreater reports whether d > other for Decimal128BID under the IEEE 754 signaling comparison predicate, raising invalid for NaN comparisons, and returns the exception flags raised by the operation.
func (d Decimal128BID) SignalingGreater(other Decimal128BID) (bool, ExceptionFlags) {
	return decimal128BIDSignalingGreaterPort(d, other)
}

// SignalingGreaterEqual reports whether d >= other for Decimal128BID under the IEEE 754 signaling comparison predicate, raising invalid for NaN comparisons, and returns the exception flags raised by the operation.
func (d Decimal128BID) SignalingGreaterEqual(other Decimal128BID) (bool, ExceptionFlags) {
	return decimal128BIDSignalingGreaterEqualPort(d, other)
}

// SignalingGreaterUnordered reports whether d > other or the operands are unordered for Decimal128BID under the IEEE 754 signaling comparison predicate, raising invalid for NaN comparisons, and returns the exception flags raised by the operation.
func (d Decimal128BID) SignalingGreaterUnordered(other Decimal128BID) (bool, ExceptionFlags) {
	return decimal128BIDSignalingGreaterUnorderedPort(d, other)
}

// SignalingLess reports whether d < other for Decimal128BID under the IEEE 754 signaling comparison predicate, raising invalid for NaN comparisons, and returns the exception flags raised by the operation.
func (d Decimal128BID) SignalingLess(other Decimal128BID) (bool, ExceptionFlags) {
	return decimal128BIDSignalingLessPort(d, other)
}

// SignalingLessEqual reports whether d <= other for Decimal128BID under the IEEE 754 signaling comparison predicate, raising invalid for NaN comparisons, and returns the exception flags raised by the operation.
func (d Decimal128BID) SignalingLessEqual(other Decimal128BID) (bool, ExceptionFlags) {
	return decimal128BIDSignalingLessEqualPort(d, other)
}

// SignalingLessUnordered reports whether d < other or the operands are unordered for Decimal128BID under the IEEE 754 signaling comparison predicate, raising invalid for NaN comparisons, and returns the exception flags raised by the operation.
func (d Decimal128BID) SignalingLessUnordered(other Decimal128BID) (bool, ExceptionFlags) {
	return decimal128BIDSignalingLessUnorderedPort(d, other)
}

// SignalingNotGreater reports whether d is not greater than other for Decimal128BID under the IEEE 754 signaling comparison predicate, raising invalid for NaN comparisons, and returns the exception flags raised by the operation.
func (d Decimal128BID) SignalingNotGreater(other Decimal128BID) (bool, ExceptionFlags) {
	return decimal128BIDSignalingNotGreaterPort(d, other)
}

// SignalingNotLess reports whether d is not less than other for Decimal128BID under the IEEE 754 signaling comparison predicate, raising invalid for NaN comparisons, and returns the exception flags raised by the operation.
func (d Decimal128BID) SignalingNotLess(other Decimal128BID) (bool, ExceptionFlags) {
	return decimal128BIDSignalingNotLessPort(d, other)
}
