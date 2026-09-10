package verification

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

func checkParserResourceEvidence(root, log string) error {
	var anchors struct {
		Cases map[string]int `json:"codec_parser_resource_cases"`
	}
	raw, err := os.ReadFile(filepath.Join(root, "devtools/verification_anchors.json"))
	if err != nil {
		return err
	}
	if err := json.Unmarshal(raw, &anchors); err != nil {
		return err
	}
	languages := []string{"go", "rust", "java", "javascript", "python", "swift"}
	if len(anchors.Cases) != len(languages) {
		return fmt.Errorf("parser resource anchors require all six languages")
	}
	for _, language := range languages {
		if anchors.Cases[language] <= 0 {
			return fmt.Errorf("missing parser resource anchor for %s", language)
		}
	}
	pattern := regexp.MustCompile(`(?m)^PARSER-RESOURCE language=(\S+) cases=([0-9]+)$`)
	matches := pattern.FindAllStringSubmatch(log, -1)
	if len(regexp.MustCompile(`(?m)^PARSER-RESOURCE[^\n]*$`).FindAllString(log, -1)) != len(matches) {
		return fmt.Errorf("malformed parser resource evidence")
	}
	seen := map[string]bool{}
	for _, match := range matches {
		language := match[1]
		n, err := strconv.Atoi(match[2])
		if err != nil || anchors.Cases[language] <= 0 || n != anchors.Cases[language] || seen[language] {
			return fmt.Errorf("parser resource %s: duplicate, unknown, or incorrect case count %s (expected %d)", language, match[2], anchors.Cases[language])
		}
		seen[language] = true
	}
	for _, language := range languages {
		if !seen[language] {
			return fmt.Errorf("missing parser resource execution for %s", language)
		}
	}
	return nil
}
