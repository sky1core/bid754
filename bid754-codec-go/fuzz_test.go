package bidcodec

// Auxiliary fuzz targets for the standalone BID codec. These are exploration
// tools, not a regular generated verification domain: the generated channels
// (vectors / reject_vectors / string_vectors under make test-bidcodec) pin the
// cross-language contract; these targets hunt for inputs those channels do
// not carry yet. Their seed corpus replays as ordinary test cases under a
// plain `go test ./...`, so no extra wiring is needed and the module stays
// stdlib-only.
//
// The properties fuzzed here are the library's own closure contract:
// FromString never panics, and every successful parse (and every decode of
// arbitrary bits) renders through ToString to a string that FromString
// reparses to the same rendering (a ToString fixed point). The committed
// corpus entry 10E2147483647 under testdata/fuzz/FuzzFromStringRoundTrip is
// the fuzz-found input that exposed the pre-closure exponent contract (its
// rendering "+1.0E+2147483648" was rejected by the parser's own old
// literal-must-fit-int32 rule); it now replays as a passing regression seed.

import "testing"

// FuzzFromStringRoundTrip: FromString must never panic; on success, the
// rendering must reparse and be render-stable.
func FuzzFromStringRoundTrip(f *testing.F) {
	for _, s := range []string{
		"", "1", "-1.5E+3", "NaN123", "SNaN", "Inf", ".5", "5.",
		"1E+2147483647", "00.001E-99", "\t1\r",
		"9999999999999999999999999999999999",
		"10E2147483647", "1E9007199254740992",
	} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		c, err := FromString(s)
		if err != nil {
			return
		}
		rendered, err := ToString(c)
		if err != nil {
			t.Fatalf("parsed Components failed ToString: %v", err)
		}
		reparsed, err := FromString(rendered)
		if err != nil {
			t.Fatalf("rendering not reparseable: %q -> %q: %v", s, rendered, err)
		}
		if again, err := ToString(reparsed); err != nil || again != rendered {
			t.Fatalf("rendering not a fixed point: %q -> %q -> %q", s, rendered, again)
		}
	})
}

// FuzzDecodeToStringReparse: Decode32/64/128 are total over arbitrary bits;
// the rendering of any decode result must reparse and be render-stable.
func FuzzDecodeToStringReparse(f *testing.F) {
	f.Add(uint32(0), uint64(0), uint64(0), uint64(0))
	f.Add(uint32(0x7c000001), uint64(0x7c00000000000001), uint64(0), uint64(0x7c00000000000001))
	f.Add(uint32(0x77f8967f), uint64(0x77fb86f26fc0ffff), uint64(0x378d8e63ffffffff), uint64(0x5fffed09bead87c0))
	f.Fuzz(func(t *testing.T, v32 uint32, v64, lo, hi uint64) {
		for _, c := range []Components{
			Decode32(v32),
			Decode64(v64),
			Decode128(lo, hi),
		} {
			rendered, err := ToString(c)
			if err != nil {
				t.Fatalf("decoded Components failed ToString: %v", err)
			}
			reparsed, err := FromString(rendered)
			if err != nil {
				t.Fatalf("decode rendering not reparseable: %q: %v", rendered, err)
			}
			if again, err := ToString(reparsed); err != nil || again != rendered {
				t.Fatalf("decode rendering not a fixed point: %q -> %q", rendered, again)
			}
		}
	})
}
