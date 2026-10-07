package testgen

import "testing"

func TestBidCodecExactExpectedBits(t *testing.T) {
	for _, tc := range []struct {
		expected string
		width    int
		lo, hi   uint64
	}{
		{"+NaN123", 32, 0x7c00007b, 0},
		{"+NaN123", 64, 0x7c0000000000007b, 0},
		{"+NaN123", 128, 123, 0x7c00000000000000},
		{"-0E-2", 32, 0xb1800000, 0},
		{"-0E-2", 64, 0xb180000000000000, 0},
		{"-0E-2", 128, 0, 0xb03c000000000000},
		{"+1.100E+0", 32, 0x3100044c, 0},
		{"+1.100E+0", 64, 0x316000000000044c, 0},
		{"+9.999999999999999999999999999999999E+33", 128, 0x378d8e63ffffffff, 0x3041ed09bead87c0},
	} {
		lo, hi, err := bidCodecExactExpectedBits(tc.expected, tc.width)
		if err != nil || lo != tc.lo || hi != tc.hi {
			t.Errorf("%q d%d = %x:%x err=%v, want %x:%x", tc.expected, tc.width, hi, lo, err, tc.hi, tc.lo)
		}
	}
	for _, tc := range []struct {
		expected string
		width    int
	}{
		{"+NaN000123", 32},
		{"+1.0E0", 32},
		{"+1E+91", 32},
		{"+9.9999999E+7", 32},
		{"+NaN1000000", 32},
		{"+NaN999999999999999999999999999999999", 64},
	} {
		if _, _, err := bidCodecExactExpectedBits(tc.expected, tc.width); err == nil {
			t.Errorf("%q d%d unexpectedly accepted", tc.expected, tc.width)
		}
	}
}

func TestBidCodecExactParseExpectationDomain(t *testing.T) {
	rows, err := bidCodecExactParseExpectations()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 112 {
		t.Fatalf("exact width cases = %d, want 112", len(rows))
	}
	counts := map[string]int{}
	for _, sv := range bidCodecGoFullStringVectorRecords() {
		for _, class := range bidCodecGoFullStringVectorClasses[sv.Input] {
			counts[class]++
		}
	}
	if counts["exact"] != 112 || counts["rounded"] != 51 || counts["rejected"] != 23 {
		t.Fatalf("string_vectors width partition = %v, want exact=112 rounded=51 rejected=23", counts)
	}
	t.Logf("string_vectors=%d exact=%d rounded=%d rejected=%d", len(bidCodecGoFullStringVectorRecords()), counts["exact"], counts["rounded"], counts["rejected"])
}
