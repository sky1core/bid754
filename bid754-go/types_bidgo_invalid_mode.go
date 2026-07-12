package bid754

import (
	"unsafe"

	bidgo "github.com/sky1core/bid754/bid754-go/internal/bidgo"
)

// Invalid-RoundingMode rejection for flag-carrying public surfaces.
//
// A public operation that takes a RoundingMode returns a value alongside its
// ExceptionFlags rather than an error, so its idiomatic failure channel is the
// flag set, not a Go error. When such a surface is handed a RoundingMode
// outside the five defined constants it cannot round, so it rejects the call
// through that flag channel instead of panicking (docs/SPEC.md: no public API
// path may panic/trap on unsupported input): it yields the value the same
// operation produces for a NaN input of the source width and raises
// FlagInvalidOperation. Measured against the Intel BID port, a NaN-to-integer
// conversion returns a fixed per-target-type sentinel and a NaN-to-float or
// NaN-to-decimal conversion returns a NaN; the rejection mirrors exactly that
// value so no sentinel is guessed. A decimal-producing surface that has no NaN
// input to mirror (integer-to-decimal constructors, context arithmetic)
// returns the canonical quiet NaN of the target width instead.
//
// This is a routing/plumbing behavior; the underlying Go mechanical port in
// internal/bidgo is unchanged.

// Canonical quiet-NaN combination-field bit patterns (Intel BID: 0x7c…, zero
// payload) for the 32- and 64-bit widths.
const (
	bidNaNBits32 uint32 = 0x7c000000
	bidNaNBits64 uint64 = 0x7c00000000000000
)

// canonicalQNaN32BID returns the canonical Decimal32 quiet NaN used as the
// rejection result of a Decimal32-producing surface handed an invalid mode.
func canonicalQNaN32BID() Decimal32BID { return Decimal32BID(bidNaNBits32) }

// canonicalQNaN64BID returns the canonical Decimal64 quiet NaN.
func canonicalQNaN64BID() Decimal64BID { return Decimal64BID(bidNaNBits64) }

// canonicalQNaN128BID returns the canonical Decimal128 quiet NaN. The high word
// carries the 0x7c… combination field; the payload words are zero.
func canonicalQNaN128BID() Decimal128BID {
	bits := bidUint128Words{w: [2]uint64{0, 0x7c00000000000000}}
	return *(*Decimal128BID)(unsafe.Pointer(&bits))
}

// bidNaN128Bidgo returns the canonical Decimal128 quiet NaN in the bidgo
// 128-bit representation, for mirroring NaN-input results on Decimal128 sources.
func bidNaN128Bidgo() bidgo.BID_UINT128 {
	return decimal128BIDAsBidgo(canonicalQNaN128BID())
}
