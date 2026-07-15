package bid754

import "testing"

type roundIntegralExactModeCase struct {
	input string
	mode  RoundingMode
	want  string
}

var roundIntegralExactModeCases = []roundIntegralExactModeCase{
	{input: "2.5", mode: RoundNearestEven, want: "2"},
	{input: "2.5", mode: RoundNearestAway, want: "3"},
	{input: "2.5", mode: RoundTowardZero, want: "2"},
	{input: "2.5", mode: RoundTowardPositive, want: "3"},
	{input: "2.5", mode: RoundTowardNegative, want: "2"},
	{input: "3.5", mode: RoundNearestEven, want: "4"},
	{input: "3.5", mode: RoundNearestAway, want: "4"},
	{input: "3.5", mode: RoundTowardZero, want: "3"},
	{input: "3.5", mode: RoundTowardPositive, want: "4"},
	{input: "3.5", mode: RoundTowardNegative, want: "3"},
	{input: "-2.5", mode: RoundNearestEven, want: "-2"},
	{input: "-2.5", mode: RoundNearestAway, want: "-3"},
	{input: "-2.5", mode: RoundTowardZero, want: "-2"},
	{input: "-2.5", mode: RoundTowardPositive, want: "-2"},
	{input: "-2.5", mode: RoundTowardNegative, want: "-3"},
}

func TestDecimal32RoundIntegralExactWithMode(t *testing.T) {
	for _, tc := range roundIntegralExactModeCases {
		input := mustDecimal32BID(t, tc.input)
		want := mustDecimal32BID(t, tc.want)
		got, flags := input.RoundIntegralExactWithMode(tc.mode)
		if got != want || flags != FlagInexact {
			t.Errorf("RoundIntegralExactWithMode(%s, %s) = %s/%s, want %s/%s", tc.input, tc.mode, got, flags, want, FlagInexact)
		}
	}
	integer := mustDecimal32BID(t, "2")
	for _, mode := range []RoundingMode{RoundNearestEven, RoundNearestAway, RoundTowardZero, RoundTowardPositive, RoundTowardNegative} {
		if got, flags := integer.RoundIntegralExactWithMode(mode); got != integer || flags != 0 {
			t.Errorf("RoundIntegralExactWithMode(2, %s) = %s/%s, want %s/None", mode, got, flags, integer)
		}
		if got, flags := mustDecimal32BID(t, "2.0").RoundIntegralExactWithMode(mode); got != integer || flags != 0 {
			t.Errorf("RoundIntegralExactWithMode(2.0, %s) = %s/%s, want %s/None", mode, got, flags, integer)
		}
	}
	if got, flags := mustDecimal32BID(t, "sNaN").RoundIntegralExactWithMode(RoundNearestEven); !got.IsNaN() || flags != FlagInvalidOperation {
		t.Errorf("RoundIntegralExactWithMode(sNaN) = %s/%s, want NaN/%s", got, flags, FlagInvalidOperation)
	}
	if got, flags := integer.RoundIntegralExactWithMode(RoundingMode(99)); got != canonicalQNaN32BID() || flags != FlagInvalidOperation {
		t.Errorf("RoundIntegralExactWithMode(2, invalid) = %s/%s, want canonical qNaN/%s", got, flags, FlagInvalidOperation)
	}
}

func TestDecimal64RoundIntegralExactWithMode(t *testing.T) {
	for _, tc := range roundIntegralExactModeCases {
		input := mustDecimal64BID(t, tc.input)
		want := mustDecimal64BID(t, tc.want)
		got, flags := input.RoundIntegralExactWithMode(tc.mode)
		if got != want || flags != FlagInexact {
			t.Errorf("RoundIntegralExactWithMode(%s, %s) = %s/%s, want %s/%s", tc.input, tc.mode, got, flags, want, FlagInexact)
		}
	}
	integer := mustDecimal64BID(t, "2")
	for _, mode := range []RoundingMode{RoundNearestEven, RoundNearestAway, RoundTowardZero, RoundTowardPositive, RoundTowardNegative} {
		if got, flags := integer.RoundIntegralExactWithMode(mode); got != integer || flags != 0 {
			t.Errorf("RoundIntegralExactWithMode(2, %s) = %s/%s, want %s/None", mode, got, flags, integer)
		}
		if got, flags := mustDecimal64BID(t, "2.0").RoundIntegralExactWithMode(mode); got != integer || flags != 0 {
			t.Errorf("RoundIntegralExactWithMode(2.0, %s) = %s/%s, want %s/None", mode, got, flags, integer)
		}
	}
	if got, flags := mustDecimal64BID(t, "sNaN").RoundIntegralExactWithMode(RoundNearestEven); !got.IsNaN() || flags != FlagInvalidOperation {
		t.Errorf("RoundIntegralExactWithMode(sNaN) = %s/%s, want NaN/%s", got, flags, FlagInvalidOperation)
	}
	if got, flags := integer.RoundIntegralExactWithMode(RoundingMode(99)); got != canonicalQNaN64BID() || flags != FlagInvalidOperation {
		t.Errorf("RoundIntegralExactWithMode(2, invalid) = %s/%s, want canonical qNaN/%s", got, flags, FlagInvalidOperation)
	}
}

func TestDecimal128RoundIntegralExactWithMode(t *testing.T) {
	for _, tc := range roundIntegralExactModeCases {
		input := mustDecimal128BID(t, tc.input)
		want := mustDecimal128BID(t, tc.want)
		got, flags := input.RoundIntegralExactWithMode(tc.mode)
		if got != want || flags != FlagInexact {
			t.Errorf("RoundIntegralExactWithMode(%s, %s) = %s/%s, want %s/%s", tc.input, tc.mode, got, flags, want, FlagInexact)
		}
	}
	integer := mustDecimal128BID(t, "2")
	for _, mode := range []RoundingMode{RoundNearestEven, RoundNearestAway, RoundTowardZero, RoundTowardPositive, RoundTowardNegative} {
		if got, flags := integer.RoundIntegralExactWithMode(mode); got != integer || flags != 0 {
			t.Errorf("RoundIntegralExactWithMode(2, %s) = %s/%s, want %s/None", mode, got, flags, integer)
		}
		if got, flags := mustDecimal128BID(t, "2.0").RoundIntegralExactWithMode(mode); got != integer || flags != 0 {
			t.Errorf("RoundIntegralExactWithMode(2.0, %s) = %s/%s, want %s/None", mode, got, flags, integer)
		}
	}
	if got, flags := mustDecimal128BID(t, "sNaN").RoundIntegralExactWithMode(RoundNearestEven); !got.IsNaN() || flags != FlagInvalidOperation {
		t.Errorf("RoundIntegralExactWithMode(sNaN) = %s/%s, want NaN/%s", got, flags, FlagInvalidOperation)
	}
	if got, flags := integer.RoundIntegralExactWithMode(RoundingMode(99)); got != canonicalQNaN128BID() || flags != FlagInvalidOperation {
		t.Errorf("RoundIntegralExactWithMode(2, invalid) = %s/%s, want canonical qNaN/%s", got, flags, FlagInvalidOperation)
	}
}
