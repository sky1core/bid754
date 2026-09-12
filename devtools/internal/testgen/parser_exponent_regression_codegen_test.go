package testgen

import (
	"testing"
)

func TestParserExpRefEncode(t *testing.T) {
	for _, tc := range []struct {
		width  int
		neg    bool
		q      int
		wantHi uint64
		wantLo uint64
	}{
		{32, false, 0, 0, 0x32800001},
		{64, false, 0, 0, 0x31c0000000000001},
		{128, false, 0, 0x3040000000000000, 0x1},
		{128, true, 0, 0xb040000000000000, 0x1},
		{64, false, 369, 0, 0x5fe0000000000001},
		{64, false, -398, 0, 0x1},
	} {
		hi, lo := parserExpRefEncode(tc.width, tc.neg, tc.q)
		if hi != tc.wantHi || lo != tc.wantLo {
			t.Errorf("parserExpRefEncode(%d,%t,%d) = %#x,%#x; want %#x,%#x", tc.width, tc.neg, tc.q, hi, lo, tc.wantHi, tc.wantLo)
		}
	}
}

func TestParserExponentRegressionCoverage(t *testing.T) {
	cases := parserExponentRegressionCases()
	if len(cases) != 49 {
		t.Fatalf("parser cases=%d want49", len(cases))
	}
	seen := map[string]bool{}
	counts := map[int]int{}
	negativeExponents := map[int]int{}
	for _, c := range cases {
		if seen[c.ID] {
			t.Errorf("duplicate case %s", c.ID)
		}
		seen[c.ID] = true
		counts[c.Width]++
		if c.IntegerZeros > 0 {
			if c.ExpMag != -c.IntegerZeros {
				t.Errorf("%s does not cancel to 1", c.ID)
			}
			negativeExponents[c.Width]++
		} else if c.Quantum != c.ExpMag-c.FracZeros-1 {
			t.Errorf("%s has an inconsistent exact quantum", c.ID)
		}
	}
	for _, width := range []int{32, 64, 128} {
		if counts[width] < 15 || negativeExponents[width] != 4 {
			t.Errorf("width=%d cases=%d negative_exponent_cases=%d", width, counts[width], negativeExponents[width])
		}
	}
}
