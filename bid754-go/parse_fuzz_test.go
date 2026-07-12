package bid754

// Auxiliary portable fuzz target for the public raw parse surface. This is an
// exploration tool, not a regular generated verification domain (the readtest
// string shards pin from_string/to_string values; this target only hunts for
// panics). It needs no native prerequisite and no build tag, and its seed
// corpus replays as an ordinary test case under a plain `go test ./...`, so
// no extra wiring is needed and the module stays stdlib-only.

import "testing"

// FuzzParseNoPanic: the public raw parse entrypoints and the String rendering
// of their results must never panic, whatever the input. Raw parses report
// problems through ExceptionFlags rather than errors, so the only failure
// mode this target asserts is a panic/crash.
func FuzzParseNoPanic(f *testing.F) {
	for _, s := range []string{
		"", "1", "+1E+0", "-inf", "nan", "snan123", "1e999999999",
		"0.00000000000000000000000001", "1..2", "١٢٣", "1E", "-",
		"9999999999999999999999999999999999E+6111",
	} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		d32, _ := ParseDecimal32BIDRaw(s)
		_ = d32.String()
		d64, _ := ParseDecimal64BIDRaw(s)
		_ = d64.String()
		d128, _ := ParseDecimal128BIDRaw(s)
		_ = d128.String()
	})
}
