package verification

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func checkCodecEvidence(root, log string) error {
	var anchors struct {
		Total             int            `json:"bid_codec_vectors_total"`
		ByWidth           map[string]int `json:"bid_codec_vectors_by_width"`
		Canonical         map[string]int `json:"bid_codec_canonical_vectors_by_width"`
		RejectTotal       int            `json:"bid_codec_reject_vectors_total"`
		RejectConsumed    map[string]int `json:"bid_codec_reject_vectors_consumed_by_language"`
		StringConsumed    map[string]int `json:"bid_codec_string_vectors_consumed_by_language"`
		GoParseReject     int            `json:"bid_codec_reject_vectors_go_full_consumed"`
		GoParseSkipped    int            `json:"bid_codec_reject_vectors_go_full_channel_skipped"`
		GoParseString     int            `json:"bid_codec_string_vectors_go_full_consumed"`
		RustParseReject   int            `json:"bid_codec_reject_vectors_rust_full_parse_consumed"`
		RustParseSkipped  int            `json:"bid_codec_reject_vectors_rust_full_parse_channel_skipped"`
		ParseOracleTuples int            `json:"bid_codec_parse_oracle_tuples"`
	}
	raw, err := os.ReadFile(filepath.Join(root, "devtools/verification_anchors.json"))
	if err != nil {
		return err
	}
	if err := json.Unmarshal(raw, &anchors); err != nil {
		return err
	}
	canonical := 0
	for _, width := range []string{"bid32", "bid64", "bid128"} {
		if anchors.ByWidth[width] <= 0 || anchors.Canonical[width] <= 0 {
			return fmt.Errorf("missing codec width anchor %s", width)
		}
		canonical += anchors.Canonical[width]
	}
	if anchors.Total <= 0 || anchors.RejectTotal <= 0 || anchors.ParseOracleTuples <= 0 {
		return fmt.Errorf("missing codec total anchors")
	}
	sections := []struct{ language, heading string }{
		{"go", "==> Go BID codec vector tests: bid754-codec-go"},
		{"rust", "==> Rust BID codec vector tests: bid754-codec-rs"},
		{"rust_full", "==> Rust bid754 BID codec vector tests: bid754-rs"},
		{"go_parse", "==> Go bid754 public parse BID codec vector tests: bid754-go"},
		{"rust_parse", "==> Rust bid754 public parse BID codec vector tests: bid754-rs"},
		{"java", "==> Java BID codec vector tests: bid754-codec-java"},
		{"python", "==> Python BID codec vector tests: bid754-codec-py"},
		{"js", "==> JavaScript/TypeScript BID codec vector tests: bid754-codec-js"},
		{"swift", "==> Swift BID codec vector tests: bid754-codec-swift"},
	}
	last := -1
	for i, section := range sections {
		marker := section.heading + "\n"
		if strings.Count(log, marker) != 1 {
			return fmt.Errorf("codec %s missing or duplicate consumer execution", section.language)
		}
		start := strings.Index(log, marker)
		if start <= last {
			return fmt.Errorf("codec consumer execution order mismatch")
		}
		last = start
		end := len(log)
		if i+1 < len(sections) {
			end = strings.Index(log, sections[i+1].heading+"\n")
			if end <= start {
				return fmt.Errorf("codec %s missing subsequent consumer", section.language)
			}
		}
		part := log[start+len(marker) : end]
		patterns := []string{}
		if section.language == "go_parse" {
			patterns = append(patterns, fmt.Sprintf(`go_full reject_vectors: consumed=%d channel_skipped=%d\b`, anchors.GoParseReject, anchors.GoParseSkipped), fmt.Sprintf(`go_full string_vectors: consumed=%d\b`, anchors.GoParseString))
			patterns = append(patterns, fmt.Sprintf(`parse oracle tuples=%d comparator baseline/bits/flags and all-mode-pair checks executed\b`, anchors.ParseOracleTuples), `(?m)^--- PASS: TestGoFullBidCodecParseComparatorStrength \(`)
		} else if section.language == "rust_parse" {
			patterns = append(patterns, fmt.Sprintf(`rust_full_parse reject_vectors: consumed=%d channel_skipped=%d\b`, anchors.RustParseReject, anchors.RustParseSkipped))
			patterns = append(patterns, fmt.Sprintf(`rust_full_parse rounded oracle tuples=%d\b`, anchors.ParseOracleTuples), `(?m)^test test_rust_full_parse_comparator_strength \.\.\. ok$`, `test result: ok\. 3 passed; 0 failed; 0 ignored; 0 measured; 0 filtered out;`)
		} else {
			consumed, stringsCount := anchors.RejectConsumed[section.language], anchors.StringConsumed[section.language]
			if consumed <= 0 || stringsCount <= 0 {
				return fmt.Errorf("codec %s missing consumer anchors", section.language)
			}
			patterns = append(patterns, fmt.Sprintf(`reject_vectors: consumed=%d skipped=%d\b`, consumed, anchors.RejectTotal-consumed), fmt.Sprintf(`string_vectors: consumed=%d\b`, stringsCount))
		}
		switch section.language {
		case "go":
			patterns = append(patterns, fmt.Sprintf(`decode: %d pass, 0 fail\b`, anchors.Total), fmt.Sprintf(`roundtrip: %d pass, 0 fail\b`, canonical))
		case "rust":
			for _, width := range []string{"bid32", "bid64", "bid128"} {
				patterns = append(patterns, fmt.Sprintf(`%s decode: %d vectors passed\b`, width, anchors.ByWidth[width]), fmt.Sprintf(`%s roundtrip: %d canonical vectors passed\b`, width, anchors.Canonical[width]))
			}
		case "rust_full":
			patterns = append(patterns, fmt.Sprintf(`BID codec vectors: decode_passed=%d encode_passed=%d skipped=%d\b`, anchors.Total, canonical, anchors.Total-canonical))
		case "java":
			patterns = append(patterns, fmt.Sprintf(`BID codec Java vectors: decode=%d encode=%d\b`, anchors.Total, canonical))
		case "js":
			patterns = append(patterns, fmt.Sprintf(`BID codec JS package vectors: decode=%d encode=%d\b`, anchors.Total, canonical))
		case "swift":
			patterns = append(patterns, fmt.Sprintf(`BID codec Swift vectors: decode=%d encode=%d\b`, anchors.Total, canonical))
		case "python":
			for _, width := range []string{"32", "64", "128"} {
				for operation, want := range map[string]int{"decode": anchors.ByWidth["bid"+width], "roundtrip": anchors.Canonical["bid"+width]} {
					pattern := `(?m)^tests/test_vectors\.py::test_` + operation + width + `\[[^\n]+\] PASSED(?:\s|$)`
					if got := len(regexp.MustCompile(pattern).FindAllStringIndex(part, -1)); got != want {
						return fmt.Errorf("codec python %s%s executed=%d, anchored=%d", operation, width, got, want)
					}
				}
			}
		}
		if section.language == "go_parse" || section.language == "rust_parse" {
			for _, width := range []int{32, 64, 128} {
				patterns = append(patterns, fmt.Sprintf(`d%d mode_pairs=10/10 directed_substitutions=20/20\b`, width))
			}
		}
		for _, pattern := range patterns {
			if !regexp.MustCompile(pattern).MatchString(part) {
				return fmt.Errorf("codec %s missing anchored evidence %s", section.language, pattern)
			}
		}
	}
	return nil
}
