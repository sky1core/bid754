package bidgo

import (
	"strings"
	"testing"
)

// cTolowerMacro is a literal transcription of the pinned Intel predecessor
// macro (devtools/third_party/intel_dfp/LIBRARY/src/bid_internal.h:94):
//
//	#define tolower_macro(x) (((unsigned char)((x)-'A')<=('Z'-'A'))?((x)-'A'+'a'):(x))
//
// The from_string special-case probes are the only fold site, so pinning the
// port helper against this transcription closes the fold semantics directly
// rather than by spot-checking a few literals.
func cTolowerMacro(x byte) byte {
	if byte(x-'A') <= byte('Z'-'A') {
		return x - 'A' + 'a'
	}
	return x
}

// TestTolowerMacroMatchesIntelDefinition closes the port helper over the whole
// byte domain, so a future edit that widens or narrows the folded range fails
// here instead of silently changing which literals from_string accepts. All
// three widths route their special-case probes through this one helper.
func TestTolowerMacroMatchesIntelDefinition(t *testing.T) {
	for i := 0; i < 256; i++ {
		b := byte(i)
		if got, want := tolower_macro(b), cTolowerMacro(b); got != want {
			t.Fatalf("tolower_macro(0x%02x) = 0x%02x, Intel macro = 0x%02x", b, got, want)
		}
	}
}

// TestEqualFoldASCIIMatchesPerByteFold checks the string helpers against a
// per-byte application of the pinned macro, which is exactly how the C probes
// are written (tolower_macro(ps[0]) == 'i' && ...).
func TestEqualFoldASCIIMatchesPerByteFold(t *testing.T) {
	lits := []string{"inf", "infinity", "nan", "snan"}
	inputs := []string{
		"", "i", "in", "inf", "INF", "Inf", "iNF", "infi", "infinity", "INFINITY",
		"InFiNiTy", "nan", "NAN", "nAn", "snan", "SNAN", "sNaN", "snan123",
		"snana", "snn", "qnan", "inf ", " inf", "in\x00", "\x00inf", "innf",
	}
	for _, lit := range lits {
		for _, in := range inputs {
			perByte := len(in) == len(lit)
			if perByte {
				for i := 0; i < len(in); i++ {
					if cTolowerMacro(in[i]) != lit[i] {
						perByte = false
						break
					}
				}
			}
			if got := equalFoldASCII(in, lit); got != perByte {
				t.Fatalf("equalFoldASCII(%q, %q) = %v, per-byte fold = %v", in, lit, got, perByte)
			}

			prefix := len(in) >= len(lit)
			if prefix {
				for i := 0; i < len(lit); i++ {
					if cTolowerMacro(in[i]) != lit[i] {
						prefix = false
						break
					}
				}
			}
			if got := hasPrefixFoldASCII(in, lit); got != prefix {
				t.Fatalf("hasPrefixFoldASCII(%q, %q) = %v, per-byte fold = %v", in, lit, got, prefix)
			}
		}
	}
}

// TestBid32FromStringSpecialCaseAcceptance pins the ASCII special-case surface
// the fold serves. Every row here is unchanged by the fold rewrite; the rows
// exist so a regression in the probe restructuring shows up as a decoded value
// change rather than only as a benchmark difference.
func TestBid32FromStringSpecialCaseAcceptance(t *testing.T) {
	cases := []struct {
		in   string
		want uint32
	}{
		{"inf", 0x78000000},
		{"INF", 0x78000000},
		{"Inf", 0x78000000},
		{"infinity", 0x78000000},
		{"INFINITY", 0x78000000},
		{"InFiNiTy", 0x78000000},
		{"+inf", 0x78000000},
		{"+INFINITY", 0x78000000},
		{"-inf", 0xf8000000},
		{"-Infinity", 0xf8000000},
		{"snan", 0x7e000000},
		{"SNaN", 0x7e000000},
		{"-snan", 0xfe000000},
		{"+snan", 0x7e000000},
		{"nan", 0x7c000000},
		{"-nan", 0xfc000000},
		{"+NaN", 0x7c000000},
		{"infi", 0x7c000000},
		{"infinit", 0x7c000000},
		{"infinityy", 0x7c000000},
		{"qnan", 0x7c000000},
		{"", 0x7c000000},
		{"   inf", 0x78000000},
		{"\tinf", 0x78000000},
	}
	for _, tc := range cases {
		got, flags := Bid32FromStringRaw(tc.in, 0)
		if got != tc.want {
			t.Errorf("Bid32FromStringRaw(%q) = 0x%08x, want 0x%08x", tc.in, got, tc.want)
		}
		if flags != 0 {
			t.Errorf("Bid32FromStringRaw(%q) raised flags 0x%x, want 0", tc.in, flags)
		}
	}
}

// TestBid32FromStringFoldIsASCIIOnly pins the one input class whose decoding
// changed when the probes moved from strings.ToLower to the pinned macro.
//
// Go's strings.ToLower applies Unicode simple case mapping, under which
// U+0130 (LATIN CAPITAL LETTER I WITH DOT ABOVE) lowers to ASCII 'i'. An
// exhaustive rune sweep shows U+0130 is the only non-ASCII rune that lowers
// into the {i,n,f,s,a,t,y} letter set these literals use, so it was the only
// input class where the previous whole-string ToLower comparison accepted a
// spelling the pinned C rejects: tolower_macro folds bytes, never runes, so
// the C sees the UTF-8 lead byte 0xC4 and falls through to qNaN. The generated
// Rust probes (eq_ignore_ascii_case) already matched the C here, so the port
// now agrees with both instead of standing alone.
func TestBid32FromStringFoldIsASCIIOnly(t *testing.T) {
	const dottedCapitalI = "İ"

	for _, in := range []string{
		dottedCapitalI + "nf",
		dottedCapitalI + "NF",
		dottedCapitalI + "nfinity",
		"+" + dottedCapitalI + "nf",
		"-" + dottedCapitalI + "nf",
	} {
		if strings.ToLower(in) != strings.ToLower(strings.ToValidUTF8(in, "")) {
			t.Fatalf("test input %q is not valid UTF-8", in)
		}
		got, flags := Bid32FromStringRaw(in, 0)
		if got != 0x7c000000 && got != 0xfc000000 {
			t.Errorf("Bid32FromStringRaw(%q) = 0x%08x, want a qNaN encoding "+
				"(the pinned C folds bytes, not runes)", in, got)
		}
		if flags != 0 {
			t.Errorf("Bid32FromStringRaw(%q) raised flags 0x%x, want 0", in, flags)
		}
	}
}

// TestBid32FromStringRejectsMultiByteInputWithoutPanic covers the input class
// that separates Go string slicing from its generated Rust counterpart.
//
// Go slices strings by byte and never faults; Rust &str slicing panics when the
// cut lands inside a multi-byte character. A prefix probe written as a slice
// therefore parses fine in Go and panics in the generated Rust on ordinary
// rejected input such as "1234é", which the public API contract forbids
// ("never panic/trap").
//
// These rows pin the Go side of that pair: the decoded value for non-ASCII
// input. The Rust side, where the panic actually occurs, is pinned separately
// by bid754-rs/tests/parse_non_ascii.rs.
func TestBid32FromStringRejectsMultiByteInputWithoutPanic(t *testing.T) {
	for _, in := range []string{
		"1234é", "1.23é", "-123é", "+123é", "1234中", "aaaé", "snané", "snaé",
		"é", "éé", "ééé", "1é", "12é", "123é", "İnf", "İNFINITY",
		"infé", "nané", "énan", "12345678é", "1.2345678é",
	} {
		got, flags := Bid32FromStringRaw(in, 0)
		// NaN-class, quiet or signaling: "snané" legitimately hits the sNaN
		// probe, which matches only the leading four bytes and ignores the
		// trailing payload exactly as the pinned C does.
		if got&NAN_MASK32 != NAN_MASK32 {
			t.Errorf("Bid32FromStringRaw(%q) = 0x%08x, want a NaN encoding", in, got)
		}
		if flags != 0 {
			t.Errorf("Bid32FromStringRaw(%q) raised flags 0x%x, want 0", in, flags)
		}
	}
}

// TestFoldHelpersStayInBoundsOnMultiByteInput exercises the probes across every
// offset of a multi-byte string and only checks that they stay in bounds.
//
// It deliberately does NOT pin the character-boundary property. Go slices
// strings by byte and never faults, so no Go test can observe the &str slicing
// rule that made a sliced prefix probe panic in the generated Rust — this test
// passes against both the indexed and the sliced form. The actual guard for
// that failure mode is bid754-rs/tests/parse_non_ascii.rs, which fails with the
// boundary panic when the sliced form is restored. What this test does catch is
// a probe that drops its length guard and reads past the end.
func TestFoldHelpersStayInBoundsOnMultiByteInput(t *testing.T) {
	for _, s := range []string{"é", "aé", "aaé", "aaaé", "中文字", "İnf", "snané"} {
		for _, lit := range []string{"inf", "infinity", "nan", "snan"} {
			_ = equalFoldASCII(s, lit)
			_ = hasPrefixFoldASCII(s, lit)
			for i := 1; i < len(s); i++ {
				_ = hasPrefixFoldASCII(s[:i], lit)
			}
		}
	}
}

// TestBid32FromStringParseDoesNotAllocate closes the regression the fold
// rewrite removes: the previous probes called strings.ToLower on every parse,
// which allocated for any input containing an upper-case byte (including the
// exponent 'E' of an ordinary numeric literal).
func TestBid32FromStringParseDoesNotAllocate(t *testing.T) {
	for _, in := range []string{"1.234E+5", "-9.999999E-95", "0", "inf", "-sNaN"} {
		allocs := testing.AllocsPerRun(100, func() {
			Bid32FromStringRaw(in, 0)
		})
		if allocs != 0 {
			t.Errorf("Bid32FromStringRaw(%q) allocated %.0f times per run, want 0", in, allocs)
		}
	}
}
