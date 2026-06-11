package bid754

import bidgo "github.com/sky1core/bid754/bid754-go/internal/bidgo"

func decimal32BIDQuietEqualPort(d, other Decimal32BID) (bool, ExceptionFlags) {
	truth, flags := bidgo.Bid32QuietEqual(d.ToUint32(), other.ToUint32())
	return truth != 0, bidgoExceptionFlags(flags)
}

func decimal32BIDQuietNotEqualPort(d, other Decimal32BID) (bool, ExceptionFlags) {
	truth, flags := bidgo.Bid32QuietNotEqual(d.ToUint32(), other.ToUint32())
	return truth != 0, bidgoExceptionFlags(flags)
}

func decimal32BIDQuietGreaterPort(d, other Decimal32BID) (bool, ExceptionFlags) {
	truth, flags := bidgo.Bid32QuietGreater(d.ToUint32(), other.ToUint32())
	return truth != 0, bidgoExceptionFlags(flags)
}

func decimal32BIDQuietGreaterEqualPort(d, other Decimal32BID) (bool, ExceptionFlags) {
	truth, flags := bidgo.Bid32QuietGreaterEqual(d.ToUint32(), other.ToUint32())
	return truth != 0, bidgoExceptionFlags(flags)
}

func decimal32BIDQuietGreaterUnorderedPort(d, other Decimal32BID) (bool, ExceptionFlags) {
	truth, flags := bidgo.Bid32QuietGreaterUnordered(d.ToUint32(), other.ToUint32())
	return truth != 0, bidgoExceptionFlags(flags)
}

func decimal32BIDQuietLessPort(d, other Decimal32BID) (bool, ExceptionFlags) {
	truth, flags := bidgo.Bid32QuietLess(d.ToUint32(), other.ToUint32())
	return truth != 0, bidgoExceptionFlags(flags)
}

func decimal32BIDQuietLessEqualPort(d, other Decimal32BID) (bool, ExceptionFlags) {
	truth, flags := bidgo.Bid32QuietLessEqual(d.ToUint32(), other.ToUint32())
	return truth != 0, bidgoExceptionFlags(flags)
}

func decimal32BIDQuietLessUnorderedPort(d, other Decimal32BID) (bool, ExceptionFlags) {
	truth, flags := bidgo.Bid32QuietLessUnordered(d.ToUint32(), other.ToUint32())
	return truth != 0, bidgoExceptionFlags(flags)
}

func decimal32BIDQuietNotGreaterPort(d, other Decimal32BID) (bool, ExceptionFlags) {
	truth, flags := bidgo.Bid32QuietNotGreater(d.ToUint32(), other.ToUint32())
	return truth != 0, bidgoExceptionFlags(flags)
}

func decimal32BIDQuietNotLessPort(d, other Decimal32BID) (bool, ExceptionFlags) {
	truth, flags := bidgo.Bid32QuietNotLess(d.ToUint32(), other.ToUint32())
	return truth != 0, bidgoExceptionFlags(flags)
}

func decimal32BIDQuietOrderedPort(d, other Decimal32BID) (bool, ExceptionFlags) {
	truth, flags := bidgo.Bid32QuietOrdered(d.ToUint32(), other.ToUint32())
	return truth != 0, bidgoExceptionFlags(flags)
}

func decimal32BIDQuietUnorderedPort(d, other Decimal32BID) (bool, ExceptionFlags) {
	truth, flags := bidgo.Bid32QuietUnordered(d.ToUint32(), other.ToUint32())
	return truth != 0, bidgoExceptionFlags(flags)
}

func decimal32BIDSignalingGreaterPort(d, other Decimal32BID) (bool, ExceptionFlags) {
	truth, flags := bidgo.Bid32SignalingGreater(d.ToUint32(), other.ToUint32())
	return truth != 0, bidgoExceptionFlags(flags)
}

func decimal32BIDSignalingGreaterEqualPort(d, other Decimal32BID) (bool, ExceptionFlags) {
	truth, flags := bidgo.Bid32SignalingGreaterEqual(d.ToUint32(), other.ToUint32())
	return truth != 0, bidgoExceptionFlags(flags)
}

func decimal32BIDSignalingGreaterUnorderedPort(d, other Decimal32BID) (bool, ExceptionFlags) {
	truth, flags := bidgo.Bid32SignalingGreaterUnordered(d.ToUint32(), other.ToUint32())
	return truth != 0, bidgoExceptionFlags(flags)
}

func decimal32BIDSignalingLessPort(d, other Decimal32BID) (bool, ExceptionFlags) {
	truth, flags := bidgo.Bid32SignalingLess(d.ToUint32(), other.ToUint32())
	return truth != 0, bidgoExceptionFlags(flags)
}

func decimal32BIDSignalingLessEqualPort(d, other Decimal32BID) (bool, ExceptionFlags) {
	truth, flags := bidgo.Bid32SignalingLessEqual(d.ToUint32(), other.ToUint32())
	return truth != 0, bidgoExceptionFlags(flags)
}

func decimal32BIDSignalingLessUnorderedPort(d, other Decimal32BID) (bool, ExceptionFlags) {
	truth, flags := bidgo.Bid32SignalingLessUnordered(d.ToUint32(), other.ToUint32())
	return truth != 0, bidgoExceptionFlags(flags)
}

func decimal32BIDSignalingNotGreaterPort(d, other Decimal32BID) (bool, ExceptionFlags) {
	truth, flags := bidgo.Bid32SignalingNotGreater(d.ToUint32(), other.ToUint32())
	return truth != 0, bidgoExceptionFlags(flags)
}

func decimal32BIDSignalingNotLessPort(d, other Decimal32BID) (bool, ExceptionFlags) {
	truth, flags := bidgo.Bid32SignalingNotLess(d.ToUint32(), other.ToUint32())
	return truth != 0, bidgoExceptionFlags(flags)
}

func decimal32BIDSignalingEqualPort(d, other Decimal32BID) (bool, ExceptionFlags) {
	ge, geFlags := bidgo.Bid32SignalingGreaterEqual(d.ToUint32(), other.ToUint32())
	le, leFlags := bidgo.Bid32SignalingLessEqual(d.ToUint32(), other.ToUint32())
	// Upstream has no signaling equal entrypoint; GE && LE preserves IEEE equality truth and OR preserves the returned invalid flag bit.
	return ge != 0 && le != 0, bidgoExceptionFlags(geFlags | leFlags)
}

func decimal32BIDSignalingNotEqualPort(d, other Decimal32BID) (bool, ExceptionFlags) {
	eq, flags := decimal32BIDSignalingEqualPort(d, other)
	return !eq, flags
}

func decimal64BIDQuietEqualPort(d, other Decimal64BID) (bool, ExceptionFlags) {
	truth, flags := bidgo.Bid64QuietEqual(d.ToUint64(), other.ToUint64())
	return truth != 0, bidgoExceptionFlags(flags)
}

func decimal64BIDQuietNotEqualPort(d, other Decimal64BID) (bool, ExceptionFlags) {
	truth, flags := bidgo.Bid64QuietNotEqual(d.ToUint64(), other.ToUint64())
	return truth != 0, bidgoExceptionFlags(flags)
}

func decimal64BIDQuietGreaterPort(d, other Decimal64BID) (bool, ExceptionFlags) {
	truth, flags := bidgo.Bid64QuietGreater(d.ToUint64(), other.ToUint64())
	return truth != 0, bidgoExceptionFlags(flags)
}

func decimal64BIDQuietGreaterEqualPort(d, other Decimal64BID) (bool, ExceptionFlags) {
	truth, flags := bidgo.Bid64QuietGreaterEqual(d.ToUint64(), other.ToUint64())
	return truth != 0, bidgoExceptionFlags(flags)
}

func decimal64BIDQuietGreaterUnorderedPort(d, other Decimal64BID) (bool, ExceptionFlags) {
	truth, flags := bidgo.Bid64QuietGreaterUnordered(d.ToUint64(), other.ToUint64())
	return truth != 0, bidgoExceptionFlags(flags)
}

func decimal64BIDQuietLessPort(d, other Decimal64BID) (bool, ExceptionFlags) {
	truth, flags := bidgo.Bid64QuietLess(d.ToUint64(), other.ToUint64())
	return truth != 0, bidgoExceptionFlags(flags)
}

func decimal64BIDQuietLessEqualPort(d, other Decimal64BID) (bool, ExceptionFlags) {
	truth, flags := bidgo.Bid64QuietLessEqual(d.ToUint64(), other.ToUint64())
	return truth != 0, bidgoExceptionFlags(flags)
}

func decimal64BIDQuietLessUnorderedPort(d, other Decimal64BID) (bool, ExceptionFlags) {
	truth, flags := bidgo.Bid64QuietLessUnordered(d.ToUint64(), other.ToUint64())
	return truth != 0, bidgoExceptionFlags(flags)
}

func decimal64BIDQuietNotGreaterPort(d, other Decimal64BID) (bool, ExceptionFlags) {
	truth, flags := bidgo.Bid64QuietNotGreater(d.ToUint64(), other.ToUint64())
	return truth != 0, bidgoExceptionFlags(flags)
}

func decimal64BIDQuietNotLessPort(d, other Decimal64BID) (bool, ExceptionFlags) {
	truth, flags := bidgo.Bid64QuietNotLess(d.ToUint64(), other.ToUint64())
	return truth != 0, bidgoExceptionFlags(flags)
}

func decimal64BIDQuietOrderedPort(d, other Decimal64BID) (bool, ExceptionFlags) {
	truth, flags := bidgo.Bid64QuietOrdered(d.ToUint64(), other.ToUint64())
	return truth != 0, bidgoExceptionFlags(flags)
}

func decimal64BIDQuietUnorderedPort(d, other Decimal64BID) (bool, ExceptionFlags) {
	truth, flags := bidgo.Bid64QuietUnordered(d.ToUint64(), other.ToUint64())
	return truth != 0, bidgoExceptionFlags(flags)
}

func decimal64BIDSignalingGreaterPort(d, other Decimal64BID) (bool, ExceptionFlags) {
	truth, flags := bidgo.Bid64SignalingGreater(d.ToUint64(), other.ToUint64())
	return truth != 0, bidgoExceptionFlags(flags)
}

func decimal64BIDSignalingGreaterEqualPort(d, other Decimal64BID) (bool, ExceptionFlags) {
	truth, flags := bidgo.Bid64SignalingGreaterEqual(d.ToUint64(), other.ToUint64())
	return truth != 0, bidgoExceptionFlags(flags)
}

func decimal64BIDSignalingGreaterUnorderedPort(d, other Decimal64BID) (bool, ExceptionFlags) {
	truth, flags := bidgo.Bid64SignalingGreaterUnordered(d.ToUint64(), other.ToUint64())
	return truth != 0, bidgoExceptionFlags(flags)
}

func decimal64BIDSignalingLessPort(d, other Decimal64BID) (bool, ExceptionFlags) {
	truth, flags := bidgo.Bid64SignalingLess(d.ToUint64(), other.ToUint64())
	return truth != 0, bidgoExceptionFlags(flags)
}

func decimal64BIDSignalingLessEqualPort(d, other Decimal64BID) (bool, ExceptionFlags) {
	truth, flags := bidgo.Bid64SignalingLessEqual(d.ToUint64(), other.ToUint64())
	return truth != 0, bidgoExceptionFlags(flags)
}

func decimal64BIDSignalingLessUnorderedPort(d, other Decimal64BID) (bool, ExceptionFlags) {
	truth, flags := bidgo.Bid64SignalingLessUnordered(d.ToUint64(), other.ToUint64())
	return truth != 0, bidgoExceptionFlags(flags)
}

func decimal64BIDSignalingNotGreaterPort(d, other Decimal64BID) (bool, ExceptionFlags) {
	truth, flags := bidgo.Bid64SignalingNotGreater(d.ToUint64(), other.ToUint64())
	return truth != 0, bidgoExceptionFlags(flags)
}

func decimal64BIDSignalingNotLessPort(d, other Decimal64BID) (bool, ExceptionFlags) {
	truth, flags := bidgo.Bid64SignalingNotLess(d.ToUint64(), other.ToUint64())
	return truth != 0, bidgoExceptionFlags(flags)
}

func decimal64BIDSignalingEqualPort(d, other Decimal64BID) (bool, ExceptionFlags) {
	ge, geFlags := bidgo.Bid64SignalingGreaterEqual(d.ToUint64(), other.ToUint64())
	le, leFlags := bidgo.Bid64SignalingLessEqual(d.ToUint64(), other.ToUint64())
	// Upstream has no signaling equal entrypoint; GE && LE preserves IEEE equality truth and OR preserves the returned invalid flag bit.
	return ge != 0 && le != 0, bidgoExceptionFlags(geFlags | leFlags)
}

func decimal64BIDSignalingNotEqualPort(d, other Decimal64BID) (bool, ExceptionFlags) {
	eq, flags := decimal64BIDSignalingEqualPort(d, other)
	return !eq, flags
}

func decimal128BIDQuietEqualPort(d, other Decimal128BID) (bool, ExceptionFlags) {
	truth, flags := bidgo.Bid128QuietEqual(decimal128BIDAsBidgo(d), decimal128BIDAsBidgo(other))
	return truth != 0, bidgoExceptionFlags(flags)
}

func decimal128BIDQuietNotEqualPort(d, other Decimal128BID) (bool, ExceptionFlags) {
	truth, flags := bidgo.Bid128QuietNotEqual(decimal128BIDAsBidgo(d), decimal128BIDAsBidgo(other))
	return truth != 0, bidgoExceptionFlags(flags)
}

func decimal128BIDQuietGreaterPort(d, other Decimal128BID) (bool, ExceptionFlags) {
	truth, flags := bidgo.Bid128QuietGreater(decimal128BIDAsBidgo(d), decimal128BIDAsBidgo(other))
	return truth != 0, bidgoExceptionFlags(flags)
}

func decimal128BIDQuietGreaterEqualPort(d, other Decimal128BID) (bool, ExceptionFlags) {
	truth, flags := bidgo.Bid128QuietGreaterEqual(decimal128BIDAsBidgo(d), decimal128BIDAsBidgo(other))
	return truth != 0, bidgoExceptionFlags(flags)
}

func decimal128BIDQuietGreaterUnorderedPort(d, other Decimal128BID) (bool, ExceptionFlags) {
	truth, flags := bidgo.Bid128QuietGreaterUnordered(decimal128BIDAsBidgo(d), decimal128BIDAsBidgo(other))
	return truth != 0, bidgoExceptionFlags(flags)
}

func decimal128BIDQuietLessPort(d, other Decimal128BID) (bool, ExceptionFlags) {
	truth, flags := bidgo.Bid128QuietLess(decimal128BIDAsBidgo(d), decimal128BIDAsBidgo(other))
	return truth != 0, bidgoExceptionFlags(flags)
}

func decimal128BIDQuietLessEqualPort(d, other Decimal128BID) (bool, ExceptionFlags) {
	truth, flags := bidgo.Bid128QuietLessEqual(decimal128BIDAsBidgo(d), decimal128BIDAsBidgo(other))
	return truth != 0, bidgoExceptionFlags(flags)
}

func decimal128BIDQuietLessUnorderedPort(d, other Decimal128BID) (bool, ExceptionFlags) {
	truth, flags := bidgo.Bid128QuietLessUnordered(decimal128BIDAsBidgo(d), decimal128BIDAsBidgo(other))
	return truth != 0, bidgoExceptionFlags(flags)
}

func decimal128BIDQuietNotGreaterPort(d, other Decimal128BID) (bool, ExceptionFlags) {
	truth, flags := bidgo.Bid128QuietNotGreater(decimal128BIDAsBidgo(d), decimal128BIDAsBidgo(other))
	return truth != 0, bidgoExceptionFlags(flags)
}

func decimal128BIDQuietNotLessPort(d, other Decimal128BID) (bool, ExceptionFlags) {
	truth, flags := bidgo.Bid128QuietNotLess(decimal128BIDAsBidgo(d), decimal128BIDAsBidgo(other))
	return truth != 0, bidgoExceptionFlags(flags)
}

func decimal128BIDQuietOrderedPort(d, other Decimal128BID) (bool, ExceptionFlags) {
	truth, flags := bidgo.Bid128QuietOrdered(decimal128BIDAsBidgo(d), decimal128BIDAsBidgo(other))
	return truth != 0, bidgoExceptionFlags(flags)
}

func decimal128BIDQuietUnorderedPort(d, other Decimal128BID) (bool, ExceptionFlags) {
	truth, flags := bidgo.Bid128QuietUnordered(decimal128BIDAsBidgo(d), decimal128BIDAsBidgo(other))
	return truth != 0, bidgoExceptionFlags(flags)
}

func decimal128BIDSignalingGreaterPort(d, other Decimal128BID) (bool, ExceptionFlags) {
	truth, flags := bidgo.Bid128SignalingGreater(decimal128BIDAsBidgo(d), decimal128BIDAsBidgo(other))
	return truth != 0, bidgoExceptionFlags(flags)
}

func decimal128BIDSignalingGreaterEqualPort(d, other Decimal128BID) (bool, ExceptionFlags) {
	truth, flags := bidgo.Bid128SignalingGreaterEqual(decimal128BIDAsBidgo(d), decimal128BIDAsBidgo(other))
	return truth != 0, bidgoExceptionFlags(flags)
}

func decimal128BIDSignalingGreaterUnorderedPort(d, other Decimal128BID) (bool, ExceptionFlags) {
	truth, flags := bidgo.Bid128SignalingGreaterUnordered(decimal128BIDAsBidgo(d), decimal128BIDAsBidgo(other))
	return truth != 0, bidgoExceptionFlags(flags)
}

func decimal128BIDSignalingLessPort(d, other Decimal128BID) (bool, ExceptionFlags) {
	truth, flags := bidgo.Bid128SignalingLess(decimal128BIDAsBidgo(d), decimal128BIDAsBidgo(other))
	return truth != 0, bidgoExceptionFlags(flags)
}

func decimal128BIDSignalingLessEqualPort(d, other Decimal128BID) (bool, ExceptionFlags) {
	truth, flags := bidgo.Bid128SignalingLessEqual(decimal128BIDAsBidgo(d), decimal128BIDAsBidgo(other))
	return truth != 0, bidgoExceptionFlags(flags)
}

func decimal128BIDSignalingLessUnorderedPort(d, other Decimal128BID) (bool, ExceptionFlags) {
	truth, flags := bidgo.Bid128SignalingLessUnordered(decimal128BIDAsBidgo(d), decimal128BIDAsBidgo(other))
	return truth != 0, bidgoExceptionFlags(flags)
}

func decimal128BIDSignalingNotGreaterPort(d, other Decimal128BID) (bool, ExceptionFlags) {
	truth, flags := bidgo.Bid128SignalingNotGreater(decimal128BIDAsBidgo(d), decimal128BIDAsBidgo(other))
	return truth != 0, bidgoExceptionFlags(flags)
}

func decimal128BIDSignalingNotLessPort(d, other Decimal128BID) (bool, ExceptionFlags) {
	truth, flags := bidgo.Bid128SignalingNotLess(decimal128BIDAsBidgo(d), decimal128BIDAsBidgo(other))
	return truth != 0, bidgoExceptionFlags(flags)
}

func decimal128BIDSignalingEqualPort(d, other Decimal128BID) (bool, ExceptionFlags) {
	ge, geFlags := bidgo.Bid128SignalingGreaterEqual(decimal128BIDAsBidgo(d), decimal128BIDAsBidgo(other))
	le, leFlags := bidgo.Bid128SignalingLessEqual(decimal128BIDAsBidgo(d), decimal128BIDAsBidgo(other))
	// Upstream has no signaling equal entrypoint; GE && LE preserves IEEE equality truth and OR preserves the returned invalid flag bit.
	return ge != 0 && le != 0, bidgoExceptionFlags(geFlags | leFlags)
}

func decimal128BIDSignalingNotEqualPort(d, other Decimal128BID) (bool, ExceptionFlags) {
	eq, flags := decimal128BIDSignalingEqualPort(d, other)
	return !eq, flags
}
