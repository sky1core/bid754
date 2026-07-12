package bid754

import "testing"

// Hand-written flag/rounding mapping anchor for the public-API parity gate.
//
// The generated parity runner (generated_public_parity_{dispatch,cases}_test.go)
// compares each public wrapper against an independent port invocation, mapping
// the port's raw status word and rounding mode with its OWN numeric literals
// (mapPortFlagsForParity, publicParityModes) rather than the wrappers'
// bidgoExceptionFlags / bidgoRoundingMode converters. That independence is only
// meaningful if the wrapper converters themselves are pinned somewhere the
// generator cannot touch: a converter bug (e.g. mapping the port invalid bit to
// the wrong public flag, or RoundTowardPositive to the wrong port integer) does
// not change any case count, so no counting gate and no reproducibility gate
// can see it. Only a hand-written semantic anchor that calls the converters
// with hardcoded numeric expectations can.
//
// This file is deliberately kept OUTSIDE the generation path, exactly like
// devtools/verification_anchors.json and
// readtest_comparator_strength_test.go: if it were generated, the same change
// that drifts a converter could regenerate the anchor to match. Do not move
// these assertions into a generated file, and do not relax a numeric
// expectation to make a drifted converter pass — fix the converter.
//
// The pinned bit values are the Intel BID_*_EXCEPTION masks from
// devtools/third_party/intel_dfp/LIBRARY/src/bid_functions.h
// (BID_INVALID=0x01, BID_ZERO_DIVIDE=0x04, BID_OVERFLOW=0x08,
// BID_UNDERFLOW=0x10, BID_INEXACT=0x20); the pinned rounding integers are the
// port-domain rounding selectors (nearest-even 0, toward-negative 1,
// toward-positive 2, toward-zero 3, nearest-away 4).

func TestPublicFlagMappingAnchor(t *testing.T) {
	cases := []struct {
		name string
		bit  uint32
		want ExceptionFlags
	}{
		{"invalid", 0x01, FlagInvalidOperation},
		{"divByZero", 0x04, FlagDivisionByZero},
		{"overflow", 0x08, FlagOverflow},
		{"underflow", 0x10, FlagUnderflow},
		{"inexact", 0x20, FlagInexact},
	}
	for _, c := range cases {
		if got := bidgoExceptionFlags(c.bit); got != c.want {
			t.Errorf("bidgoExceptionFlags(%#x) = %v, want %v", c.bit, got, c.want)
		}
	}

	// Intel BID_DENORMAL_EXCEPTION (0x02, DEC_FE_UNNORMAL in the pinned
	// bid_functions.h) is not one of the five IEEE exception flags the public
	// surface exposes: it must map to nothing — in particular it must not
	// propagate into FlagSubnormal or any other public flag.
	if got := bidgoExceptionFlags(0x02); got != 0 {
		t.Errorf("bidgoExceptionFlags(0x02 BID_DENORMAL_EXCEPTION) = %v, want 0", got)
	}

	// The five bits together map to exactly the union of the five public flags
	// and nothing else (no Subnormal/Rounded/Clamped propagation), and the ignored
	// denormal bit does not change that union.
	all := FlagInvalidOperation | FlagDivisionByZero | FlagOverflow | FlagUnderflow | FlagInexact
	if got := bidgoExceptionFlags(0x01 | 0x04 | 0x08 | 0x10 | 0x20); got != all {
		t.Errorf("bidgoExceptionFlags(all five bits) = %v, want %v", got, all)
	}
	if got := bidgoExceptionFlags(0x01 | 0x02 | 0x04 | 0x08 | 0x10 | 0x20); got != all {
		t.Errorf("bidgoExceptionFlags(all five bits | 0x02) = %v, want %v", got, all)
	}

	// No status bit — the ignored denormal bit included — produces any of the
	// display-only public flags.
	for bit := uint32(0x01); bit <= 0x20; bit <<= 1 {
		if bidgoExceptionFlags(bit)&(FlagSubnormal|FlagRounded|FlagClamped) != 0 {
			t.Errorf("bidgoExceptionFlags(%#x) mapped a display-only flag", bit)
		}
	}
}

func TestPublicRoundingMappingAnchor(t *testing.T) {
	cases := []struct {
		mode RoundingMode
		want int
	}{
		{RoundNearestEven, 0},
		{RoundTowardNegative, 1},
		{RoundTowardPositive, 2},
		{RoundTowardZero, 3},
		{RoundNearestAway, 4},
	}
	for _, c := range cases {
		got, ok := bidgoRoundingMode(c.mode)
		if !ok || got != c.want {
			t.Errorf("bidgoRoundingMode(%v) = %d, %v, want %d, true", c.mode, got, ok, c.want)
		}
	}

	// A RoundingMode outside the five defined constants is rejected strict
	// (ok == false), never silently coerced to a valid selector. This is the
	// pinned contract that lets every public wrapper reject an undefined mode
	// through its own failure channel instead of panicking; do not relax it to
	// return a valid selector for an undefined mode.
	for _, bad := range []RoundingMode{RoundingMode(-1), RoundingMode(5), RoundingMode(99)} {
		if got, ok := bidgoRoundingMode(bad); ok {
			t.Errorf("bidgoRoundingMode(%d) = %d, true, want _, false", int(bad), got)
		}
	}

	if defaultBIDRoundingMode != 0 {
		t.Errorf("defaultBIDRoundingMode = %d, want 0 (round-nearest-even)", defaultBIDRoundingMode)
	}
}
