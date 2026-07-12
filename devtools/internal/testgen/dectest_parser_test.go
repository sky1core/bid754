package testgen

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseDecTestFileKeepsCasesWithColonInInlineComment(t *testing.T) {
	path := filepath.Join(t.TempDir(), "inline-comment.decTest")
	content := "version: 2.62\n" +
		"extended: 1\n" +
		"dectest: add\n" +
		"precision: 5 -- exact numeric directive with a comment\n" +
		"maxexponent: 10 -- boundary\n" +
		"minexponent: -9\n" +
		"clamp: 1\n" +
		"rounding: half_up\n" +
		"comment001 add -0 0 -> 0 -- note: colon belongs to the comment; ignore -> here\n" +
		"comment002 add '--1' 0 -> NaN -- quoted -- is an operand\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write dectest file: %v", err)
	}

	cases, err := parseDecTestFile(path)
	if err != nil {
		t.Fatalf("parseDecTestFile returned error: %v", err)
	}
	if len(cases) != 2 {
		t.Fatalf("parseDecTestFile returned %d cases, want 2", len(cases))
	}
	got := cases[0]
	if got.ID != "comment001" || got.Operation != "add" || got.Result != "0" {
		t.Fatalf("parsed case = %+v, want comment001 add -> 0", got)
	}
	if got.Precision != 5 || got.MaxExponent != 10 || got.MinExponent != -9 || got.Clamp != 1 || got.RoundingMode != "half_up" {
		t.Fatalf("parsed context = precision %d emax %d emin %d clamp %d rounding %q", got.Precision, got.MaxExponent, got.MinExponent, got.Clamp, got.RoundingMode)
	}
	if operands := cases[1].Operands; len(operands) != 2 || operands[0] != "'--1'" {
		t.Fatalf("quoted comment marker parsed as operands %v", operands)
	}
}

func TestParseDecTestFileRejectsMalformedDirectivesAndCases(t *testing.T) {
	testCases := []struct {
		name    string
		content string
	}{
		{name: "invalid precision", content: "precision: nope\n"},
		{name: "trailing precision junk", content: "precision: 5junk\n"},
		{name: "unknown rounding", content: "rounding: nearestish\n"},
		{name: "invalid clamp", content: "clamp: 2\n"},
		{name: "invalid extended", content: "extended: 2\n"},
		{name: "directive with arrow", content: "precision: 16 -> 7\n"},
		{name: "unknown directive", content: "mystery: value\n"},
		{name: "unexpected content", content: "orphan text\n"},
		{name: "missing result", content: "bad001 add 1 2 ->\n"},
		{name: "multiple arrows", content: "bad002 add 1 2 -> 3 -> 4\n"},
		{name: "unterminated quote", content: "bad003 add '1 2 -> 3\n"},
		{name: "missing operands", content: "bad004 add -> 3\n"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "malformed.decTest")
			if err := os.WriteFile(path, []byte(tc.content), 0o644); err != nil {
				t.Fatalf("write dectest file: %v", err)
			}
			if _, err := parseDecTestFile(path); err == nil {
				t.Fatalf("parseDecTestFile accepted malformed input %q", tc.content)
			}
		})
	}
}
