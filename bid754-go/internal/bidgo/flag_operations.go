// Ported from: IntelRDFPMathLib20U4/LIBRARY/src/bid_flag_operations.c
// Version: Intel(R) Decimal Floating-Point Math Library 2.0 Update 4
//
// Non-computational operations on flags (IEEE 754-2019 section 5.7.4). This
// file is a mechanical translation of the explicit-argument build configuration
// this repository pins in bid_conf.h — !DECIMAL_CALL_BY_REFERENCE,
// !DECIMAL_GLOBAL_EXCEPTION_FLAGS, !DECIMAL_GLOBAL_ROUNDING — so the status
// word arrives as the pfpsf pointer parameter and the rounding direction
// arrives as the rnd_mode value parameter instead of the library globals
// _IDEC_glbflags / _IDEC_glbround. All logic and magic numbers are preserved
// exactly; _IDEC_flags and _IDEC_round are `unsigned int` in bid_conf.h, so
// both map to uint32.

package bidgo

// BidSignalException is the mechanical port of bid_signalException.
// flagsmask is the logical OR of the flags to be set, e.g.
// flagsmask = BID_INVALID_EXCEPTION | BID_ZERO_DIVIDE_EXCEPTION | BID_OVERFLOW_EXCEPTION
// BID_UNDERFLOW_EXCEPTION | BID_INEXACT_EXCEPTION to set all five IEEE 754
// exception flags.
func BidSignalException(flagsmask uint32, pfpsf *uint32) {
	*pfpsf = *pfpsf | (flagsmask & BID_IEEE_FLAGS)
}

// BidLowerFlags is the mechanical port of bid_lowerFlags.
// flagsmask is the logical OR of the flags to be cleared, e.g.
// flagsmask = BID_INVALID_EXCEPTION | BID_ZERO_DIVIDE_EXCEPTION | BID_OVERFLOW_EXCEPTION
// BID_UNDERFLOW_EXCEPTION | BID_INEXACT_EXCEPTION to clear all five IEEE 754
// exception flags.
func BidLowerFlags(flagsmask uint32, pfpsf *uint32) {
	*pfpsf = *pfpsf & ^(flagsmask & BID_IEEE_FLAGS)
}

// BidTestFlags is the mechanical port of bid_testFlags.
// The return value raised is the logical OR of the flags selected by flagsmask
// that are set; e.g. if
// flagsmask = BID_INVALID_EXCEPTION | BID_UNDERFLOW_EXCEPTION | BID_INEXACT_EXCEPTION and
// only the invalid and inexact flags are raised (set) then the return value
// is raised = BID_INVALID_EXCEPTION | BID_INEXACT_EXCEPTION.
func BidTestFlags(flagsmask uint32, pfpsf *uint32) uint32 {
	var raised uint32
	raised = *pfpsf & (flagsmask & BID_IEEE_FLAGS)
	return raised
}

// BidTestSavedFlags is the mechanical port of bid_testSavedFlags.
// The return value raised is the logical OR of the flags selected by flagsmask
// that are set in savedflags; e.g. if
// flagsmask = BID_INVALID_EXCEPTION | BID_UNDERFLOW_EXCEPTION | BID_INEXACT_EXCEPTION and
// only the invalid and inexact flags are raised (set) in savedflags
// then the return value is raised = BID_INVALID_EXCEPTION | BID_INEXACT_EXCEPTION.
// Note that the flags could be saved in a global variable, but this function
// would still expect that value as an argument passed by value.
func BidTestSavedFlags(savedflags uint32, flagsmask uint32) uint32 {
	var raised uint32
	raised = savedflags & (flagsmask & BID_IEEE_FLAGS)
	return raised
}

// BidRestoreFlags is the mechanical port of bid_restoreFlags.
// It restores the status flags selected by flagsmask to the values specified
// (as a logical OR) in flagsvalues; e.g. if
// flagsmask = BID_INVALID_EXCEPTION | BID_UNDERFLOW_EXCEPTION | BID_INEXACT_EXCEPTION
// and only the invalid and inexact flags are raised (set) in flagsvalues
// then upon return the invalid status flag will be set, the underflow status
// flag will be clear, and the inexact status flag will be set.
func BidRestoreFlags(flagsvalues uint32, flagsmask uint32, pfpsf *uint32) {
	*pfpsf = *pfpsf & ^(flagsmask & BID_IEEE_FLAGS)
	// clear flags that have to be restored
	*pfpsf = *pfpsf | (flagsvalues & (flagsmask & BID_IEEE_FLAGS))
	// restore flags
}

// BidSaveFlags is the mechanical port of bid_saveFlags.
// It returns the status flags specified (as a logical OR) in flagsmask; e.g. if
// flagsmask = BID_INVALID_EXCEPTION | BID_UNDERFLOW_EXCEPTION | BID_INEXACT_EXCEPTION
// and only the invalid and inexact flags are raised (set) in the status word,
// then the return value will have the invalid status flag set, the
// underflow status flag clear, and the inexact status flag set.
func BidSaveFlags(flagsmask uint32, pfpsf *uint32) uint32 {
	var flagsvalues uint32
	flagsvalues = *pfpsf & (flagsmask & BID_IEEE_FLAGS)
	return flagsvalues
}

// BidGetDecimalRoundingDirection is the mechanical port of
// bid_getDecimalRoundingDirection: it returns the current rounding mode, which
// in the pinned !DECIMAL_GLOBAL_ROUNDING configuration is the rnd_mode value
// parameter itself.
func BidGetDecimalRoundingDirection(rnd_mode uint32) uint32 {
	return rnd_mode
}

// BidSetDecimalRoundingDirection is the mechanical port of
// bid_setDecimalRoundingDirection: it sets the current rounding mode to the
// value in rounding_mode; however, when arguments are passed by value and the
// rounding mode is a local variable, this is not of any use, so the pinned
// configuration returns the accepted mode and leaves rnd_mode unchanged when
// rounding_mode is not one of the five valid modes.
func BidSetDecimalRoundingDirection(rounding_mode uint32, rnd_mode uint32) uint32 {
	if rounding_mode == BID_ROUNDING_TO_NEAREST ||
		rounding_mode == BID_ROUNDING_DOWN ||
		rounding_mode == BID_ROUNDING_UP ||
		rounding_mode == BID_ROUNDING_TO_ZERO ||
		rounding_mode == BID_ROUNDING_TIES_AWAY {
		return rounding_mode
	}
	return rnd_mode
}
