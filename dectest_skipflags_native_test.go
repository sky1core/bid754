//go:build cgo && bid754_native

package bid754

import "testing"

func TestShouldSkipDecTestFlagsAllowsNativeArithmeticFlagEdges(t *testing.T) {
	testCases := []decTestCase{
		{Operation: "divide", Flags: []string{"Division_undefined"}},
		{Operation: "add", Flags: []string{"Clamped"}},
	}

	for _, tc := range testCases {
		if shouldSkipDecTestFlags(tc) {
			t.Fatalf("expected native flag verification to keep %v runnable", tc.Flags)
		}
	}
}
