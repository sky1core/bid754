// Hand-written coverage gate for BidSetDecimalRoundingDirection (the mechanical
// port of bid_setDecimalRoundingDirection, flag_operations.go). It lives
// OUTSIDE every generation path and must remain hand-written and unchanged by
// any emitter.
//
// Why this gate exists: the generated readtest rows for this function
// (devtools/generated/testspec/readtest/bid_setDecimalRoundingDirection.json,
// 26 rows) only ever pass requested ∈ {-1, 0, 1, 5, 6, 8}, so of the five
// accept arms in the ported disjunction only BID_ROUNDING_TO_NEAREST (0) and
// BID_ROUNDING_DOWN (1) are ever taken. The UP (2), TO_ZERO (3) and TIES_AWAY
// (4) arms could be deleted or miswired and every gate would still pass,
// including the native C differential leg, which replays the same corpus and so
// has the same blind spot. Source-to-source review is otherwise the only
// defense; this test is the mechanical one. The reject arm and the non-standard
// bidgo mode 5 are also pinned here so the whole contract is asserted in one
// place, but arms 2/3/4 are the coverage that exists nowhere else.
//
// Expected values are derived from the pinned Intel C, not captured from the
// port:
//   - devtools/third_party/intel_dfp/LIBRARY/src/bid_flag_operations.c:299-311
//     is the !DECIMAL_CALL_BY_REFERENCE && !DECIMAL_GLOBAL_ROUNDING build this
//     repository pins: on a valid rounding_mode it returns rounding_mode,
//     otherwise it returns the current rnd_mode unchanged. Valid means exactly
//     one of BID_ROUNDING_TO_NEAREST, _DOWN, _UP, _TO_ZERO, _TIES_AWAY.
//   - devtools/third_party/intel_dfp/LIBRARY/src/bid_functions.h:140-144 pins
//     those five macros to 0, 1, 2, 3, 4.
//
// The requested modes below are therefore written as bare numeric literals
// rather than through the bidgo constants: the numbers are the contract, so a
// renumbered package constant must fail this test instead of moving with it.
package bidgo

import "testing"

// setDecimalRoundingDirectionValidModes are the five values pinned C accepts,
// as the raw macro numbers from bid_functions.h:140-144.
var setDecimalRoundingDirectionValidModes = []struct {
	name      string
	requested uint32
}{
	{name: "BID_ROUNDING_TO_NEAREST", requested: 0},
	{name: "BID_ROUNDING_DOWN", requested: 1},
	{name: "BID_ROUNDING_UP", requested: 2},
	{name: "BID_ROUNDING_TO_ZERO", requested: 3},
	{name: "BID_ROUNDING_TIES_AWAY", requested: 4},
}

// setDecimalRoundingDirectionInvalidModes are values pinned C rejects, i.e.
// everything outside 0..4. Mode 5 is the non-standard bidgo
// BID_ROUNDING_NEAREST_DOWN kept for decTest compatibility (internal.go:105);
// pinned C has no such mode, so this entrypoint must reject it like any other
// out-of-range value.
var setDecimalRoundingDirectionInvalidModes = []struct {
	name      string
	requested uint32
}{
	{name: "bidgo_nonstandard_nearest_down_5", requested: 5},
	{name: "6", requested: 6},
	{name: "7", requested: 7},
	{name: "8", requested: 8},
	{name: "0x0000ffff", requested: 0x0000ffff},
	{name: "0x80000000", requested: 0x80000000},
	{name: "0xffffffff_negative_one", requested: 0xffffffff},
}

// setDecimalRoundingDirectionCurrentModes are the current-mode (rnd_mode)
// values every case is run against: the five valid modes plus a value that is
// not a rounding mode at all. The last one matters for the accept cases, where
// returning the current mode instead of the requested mode must be visible; the
// entrypoint returns rnd_mode verbatim on rejection, so it is not required to
// be a valid mode.
var setDecimalRoundingDirectionCurrentModes = []uint32{0, 1, 2, 3, 4, 0xa5a5a5a5}

// TestBidSetDecimalRoundingDirectionAcceptsEveryValidMode pins the accept arm:
// for each of the five valid modes the requested mode comes back, whatever the
// current mode is (bid_flag_operations.c:303-309).
func TestBidSetDecimalRoundingDirectionAcceptsEveryValidMode(t *testing.T) {
	for _, mode := range setDecimalRoundingDirectionValidModes {
		discriminating := 0
		for _, current := range setDecimalRoundingDirectionCurrentModes {
			if current == mode.requested {
				// A current mode equal to the requested mode cannot tell the
				// accept arm from the reject arm; skip it as evidence.
				continue
			}
			discriminating++
			got := BidSetDecimalRoundingDirection(mode.requested, current)
			if got != mode.requested {
				t.Errorf("BidSetDecimalRoundingDirection(%d /* %s */, %#x) = %#x, want %d (pinned C returns the requested mode)",
					mode.requested, mode.name, current, got, mode.requested)
			}
		}
		if discriminating == 0 {
			t.Errorf("%s (%d): no current mode differs from the requested mode, so the accept arm is untested",
				mode.name, mode.requested)
		}
	}
}

// TestBidSetDecimalRoundingDirectionRejectsInvalidModeAndKeepsCurrent pins the
// reject arm: an out-of-range requested mode leaves the current mode in place
// (bid_flag_operations.c:310).
func TestBidSetDecimalRoundingDirectionRejectsInvalidModeAndKeepsCurrent(t *testing.T) {
	for _, mode := range setDecimalRoundingDirectionInvalidModes {
		for _, current := range setDecimalRoundingDirectionCurrentModes {
			got := BidSetDecimalRoundingDirection(mode.requested, current)
			if got != current {
				t.Errorf("BidSetDecimalRoundingDirection(%#x /* %s */, %#x) = %#x, want %#x (pinned C returns the unchanged current mode)",
					mode.requested, mode.name, current, got, current)
			}
		}
	}
}

// TestBidSetDecimalRoundingDirectionModeTablesPinIntelMacroValues keeps the two
// tables above honest. Without it, deleting an accept arm from the port
// together with its row here would still be green; with it, the valid-mode set
// is pinned to the exact five numbers of bid_functions.h:140-144 and the
// invalid list cannot silently absorb one of them.
func TestBidSetDecimalRoundingDirectionModeTablesPinIntelMacroValues(t *testing.T) {
	wantValid := []uint32{0, 1, 2, 3, 4}
	if len(setDecimalRoundingDirectionValidModes) != len(wantValid) {
		t.Fatalf("valid-mode table has %d entries, want %d (bid_functions.h:140-144)",
			len(setDecimalRoundingDirectionValidModes), len(wantValid))
	}
	for i, want := range wantValid {
		if got := setDecimalRoundingDirectionValidModes[i].requested; got != want {
			t.Errorf("valid-mode table entry %d (%s) = %d, want %d (bid_functions.h:140-144)",
				i, setDecimalRoundingDirectionValidModes[i].name, got, want)
		}
	}

	valid := make(map[uint32]bool, len(wantValid))
	for _, want := range wantValid {
		valid[want] = true
	}
	if len(setDecimalRoundingDirectionInvalidModes) == 0 {
		t.Fatal("invalid-mode table is empty, so the reject arm is untested")
	}
	for _, mode := range setDecimalRoundingDirectionInvalidModes {
		if valid[mode.requested] {
			t.Errorf("invalid-mode table entry %s (%d) is a mode pinned C accepts", mode.name, mode.requested)
		}
	}
}
