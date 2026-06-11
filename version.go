package bid754

// Is754Version1985 reports whether this library implements IEEE 754-1985.
//
// IEEE 754-1985 did not include decimal floating-point formats, and the
// pinned Intel BID upstream reports the same (bid_is754() returns 0 in
// third_party/intel_dfp/LIBRARY/src/bid_flag_operations.c).
func Is754Version1985() bool {
	return false
}

// Is754Version2008 reports whether this library implements IEEE 754-2008.
//
// The supported BID decimal surface implements the IEEE 754-2008 decimal
// requirements; the pinned Intel BID upstream reports the same (bid_is754R()
// returns 1 in third_party/intel_dfp/LIBRARY/src/bid_flag_operations.c).
func Is754Version2008() bool {
	return true
}

// Is754Version2019 reports whether this library implements IEEE 754-2019.
//
// The supported BID decimal surface implements the IEEE 754-2019 mandatory
// operations; intentional IEEE-conformance deviations from the pinned Intel
// BID C upstream are documented in IEEE754_SPEC.md.
func Is754Version2019() bool {
	return true
}
