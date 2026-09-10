package bidgo

import (
	"fmt"
	"strings"
	"testing"
)

var parserResourceBits BID_UINT128
var parserResourceFlags uint32

var resourceParsers = []struct {
	name  string
	parse func(string, int) (BID_UINT128, uint32)
}{
	{"bid64", func(s string, mode int) (BID_UINT128, uint32) {
		bits, flags := bid64_from_string(s, mode)
		return BID_UINT128{lo: bits}, flags
	}},
	{"bid128", Bid128FromString},
}

func TestRawStringParserResources(t *testing.T) {
	for _, size := range []int{64, 1 << 20} {
		zeros := strings.Repeat("0", size)
		for _, tc := range []struct{ name, input, equivalent string }{
			{"leading_zeros", zeros + "123.50", "123.50"},
			{"zero", zeros, "0"},
			{"whitespace", strings.Repeat(" \t", size/2) + "-123.50", "-123.50"},
			{"exponent_zeros", "1e+" + zeros + "2", "1e+2"},
			{"invalid_tail", zeros + "1.2.3", "NaN"},
			{"invalid_prefix", "x" + zeros, "NaN"},
			{"nul_tail", "123.50\x00" + zeros + "x", "123.50"},
			{"snan_payload", "-sNaN" + zeros, "-sNaN"},
		} {
			for _, parser := range resourceParsers {
				t.Run(fmt.Sprintf("%s/%s/%d", parser.name, tc.name, size), func(t *testing.T) {
					for mode := 0; mode <= 4; mode++ {
						want, wantFlags := parser.parse(tc.equivalent, mode)
						got, flags := parser.parse(tc.input, mode)
						if got != want || flags != wantFlags {
							t.Fatalf("mode %d: got %x/%x, want %x/%x", mode, got, flags, want, wantFlags)
						}
					}
					allocs := testing.AllocsPerRun(10, func() {
						parserResourceBits, parserResourceFlags = parser.parse(tc.input, BID_ROUNDING_TO_NEAREST)
					})
					if allocs != 0 {
						t.Fatalf("input bytes %d: got %.0f allocations per parse, want 0", len(tc.input), allocs)
					}
				})
			}
		}
	}
}

func TestRawStringParserLookahead(t *testing.T) {
	for _, parser := range resourceParsers {
		for _, token := range []string{"Infinity", "sNaN", "-Infinity", "+sNaN", "-1.25e-20", "+", "-", ".", "1e+", "é", "\xff"} {
			for end := 0; end <= len(token); end++ {
				input := token[:end]
				for mode := 0; mode <= 4; mode++ {
					got, flags := parser.parse(input, mode)
					want, wantFlags := parser.parse(input+"\x00ignored", mode)
					if got != want || flags != wantFlags {
						t.Fatalf("%s %q mode %d: got %x/%x, NUL terminated %x/%x", parser.name, input, mode, got, flags, want, wantFlags)
					}
				}
			}
		}
	}
}

func TestRawStringParserPrefixAcceptance(t *testing.T) {
	for mode := 0; mode <= 4; mode++ {
		for _, input := range []string{"1e2tail", "1e2é", "1e00000002tail"} {
			got64, flags64 := bid64_from_string(input, mode)
			if got64 != 0x7c00000000000000 || flags64 != 0 {
				t.Fatalf("bid64 %q mode %d: got %x/%x", input, mode, got64, flags64)
			}
			got128, flags128 := Bid128FromString(input, mode)
			want128, wantFlags128 := Bid128FromString("1e2", mode)
			if got128 != want128 || flags128 != wantFlags128 {
				t.Fatalf("bid128 %q mode %d: got %x/%x, want %x/%x", input, mode, got128, flags128, want128, wantFlags128)
			}
		}
	}
}

func BenchmarkRawStringParser(b *testing.B) {
	for _, size := range []int{16, 1024, 1 << 20} {
		for _, tc := range []struct{ name, input string }{
			{"leading_zeros", strings.Repeat("0", size) + "123.50"},
			{"invalid_prefix", "x" + strings.Repeat("0", size)},
		} {
			for _, parser := range resourceParsers {
				b.Run(fmt.Sprintf("%s/%s/%d", parser.name, tc.name, size), func(b *testing.B) {
					b.ReportAllocs()
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						parserResourceBits, parserResourceFlags = parser.parse(tc.input, BID_ROUNDING_TO_NEAREST)
					}
				})
			}
		}
	}
}
