package verification

import (
	"strings"
	"testing"
)

func TestCodecAdjudicationRequiresEachLanguageAndItsAnchoredCounts(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "devtools/verification_anchors.json", `{
"bid_codec_parse_oracle_tuples":15,
"bid_codec_vectors_total":3,
"bid_codec_vectors_by_width":{"bid32":1,"bid64":1,"bid128":1},
"bid_codec_canonical_vectors_by_width":{"bid32":1,"bid64":1,"bid128":1},
"bid_codec_reject_vectors_total":1,
"bid_codec_reject_vectors_consumed_by_language":{"go":1,"rust":1,"rust_full":1,"java":1,"python":1,"js":1,"swift":1},
"bid_codec_string_vectors_consumed_by_language":{"go":1,"rust":1,"rust_full":1,"java":1,"python":1,"js":1,"swift":1},
"bid_codec_reject_vectors_go_full_consumed":1,"bid_codec_reject_vectors_go_full_channel_skipped":0,
"bid_codec_string_vectors_go_full_consumed":1,
"bid_codec_reject_vectors_rust_full_parse_consumed":1,"bid_codec_reject_vectors_rust_full_parse_channel_skipped":0
}`)
	common := "reject_vectors: consumed=1 skipped=0\nstring_vectors: consumed=1\n"
	log := "==> Go BID codec vector tests: bid754-codec-go\n" + common + "decode: 3 pass, 0 fail\nroundtrip: 3 pass, 0 fail\n" +
		"==> Rust BID codec vector tests: bid754-codec-rs\n" + common + "bid32 decode: 1 vectors passed\nbid64 decode: 1 vectors passed\nbid128 decode: 1 vectors passed\nbid32 roundtrip: 1 canonical vectors passed\nbid64 roundtrip: 1 canonical vectors passed\nbid128 roundtrip: 1 canonical vectors passed\n" +
		"==> Rust bid754 BID codec vector tests: bid754-rs\n" + common + "BID codec vectors: decode_passed=3 encode_passed=3 skipped=0\n" +
		"==> Go bid754 public parse BID codec vector tests: bid754-go\ngo_full reject_vectors: consumed=1 channel_skipped=0\ngo_full string_vectors: consumed=1\nparse oracle tuples=15 comparator baseline/bits/flags and all-mode-pair checks executed\n--- PASS: TestGoFullBidCodecParseComparatorStrength (0.01s)\nd32 mode_pairs=10/10 directed_substitutions=20/20\nd64 mode_pairs=10/10 directed_substitutions=20/20\nd128 mode_pairs=10/10 directed_substitutions=20/20\n" +
		"==> Rust bid754 public parse BID codec vector tests: bid754-rs\nrust_full_parse reject_vectors: consumed=1 channel_skipped=0\nrust_full_parse rounded oracle tuples=15\ntest test_rust_full_parse_comparator_strength ... ok\nd32 mode_pairs=10/10 directed_substitutions=20/20\nd64 mode_pairs=10/10 directed_substitutions=20/20\nd128 mode_pairs=10/10 directed_substitutions=20/20\ntest result: ok. 3 passed; 0 failed; 0 ignored; 0 measured; 0 filtered out;\n" +
		"==> Java BID codec vector tests: bid754-codec-java\n" + common + "BID codec Java vectors: decode=3 encode=3\n" +
		"==> Python BID codec vector tests: bid754-codec-py\n" + common +
		"tests/test_vectors.py::test_decode32[vec0] PASSED\ntests/test_vectors.py::test_decode64[vec0] PASSED\ntests/test_vectors.py::test_decode128[vec0] PASSED\n" +
		"tests/test_vectors.py::test_roundtrip32[vec0] PASSED\ntests/test_vectors.py::test_roundtrip64[vec0] PASSED\ntests/test_vectors.py::test_roundtrip128[vec0] PASSED\n" +
		"==> JavaScript/TypeScript BID codec vector tests: bid754-codec-js\n" + common + "BID codec JS package vectors: decode=3 encode=3\n" +
		"==> Swift BID codec vector tests: bid754-codec-swift\n" + common + "BID codec Swift vectors: decode=3 encode=3\n"
	if err := checkCodecEvidence(root, log); err != nil {
		t.Fatalf("valid evidence rejected: %v", err)
	}
	for _, tc := range []struct{ name, old, new, want string }{
		{"missing language", "==> Swift BID codec vector tests: bid754-codec-swift", "skipped swift", "missing subsequent consumer"},
		{"short corpus", "decode: 3 pass, 0 fail", "decode: 2 pass, 0 fail", "codec go missing anchored evidence"},
		{"missing mode coverage", "d64 mode_pairs=10/10 directed_substitutions=20/20", "d64 mode_pairs=9/10 directed_substitutions=18/20", "codec go_parse missing anchored evidence"},
		{"reduced oracle", "parse oracle tuples=15 comparator", "parse oracle tuples=10 comparator", "codec go_parse missing anchored evidence"},
		{"reported case failure", "decode: 3 pass, 0 fail", "decode: 3 pass, 1 fail", "codec go missing anchored evidence"},
		{"missing decode execution", "tests/test_vectors.py::test_decode64[vec0] PASSED", "", "codec python decode64 executed=0, anchored=1"},
		{"language evidence substituted", "BID codec Swift vectors: decode=3 encode=3", "BID codec Java vectors: decode=3 encode=3", "codec swift missing anchored evidence"},
		{"wrong reject count", "reject_vectors: consumed=1 skipped=0", "reject_vectors: consumed=0 skipped=1", "codec go missing anchored evidence"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mutant := strings.Replace(log, tc.old, tc.new, 1)
			if err := checkCodecEvidence(root, mutant); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("want %q, got %v", tc.want, err)
			}
		})
	}
}
