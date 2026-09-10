package verification

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestParserResourceEvidenceRequiresEveryLanguageAndCount(t *testing.T) {
	root := t.TempDir()
	writeFixture(t, root, "devtools/verification_anchors.json", `{"codec_parser_resource_cases":{"go":3,"rust":4,"java":5,"javascript":6,"python":7,"swift":8}}`)
	lines := []string{
		"PARSER-RESOURCE language=go cases=3\n",
		"PARSER-RESOURCE language=rust cases=4\n",
		"PARSER-RESOURCE language=java cases=5\n",
		"PARSER-RESOURCE language=javascript cases=6\n",
		"PARSER-RESOURCE language=python cases=7\n",
		"PARSER-RESOURCE language=swift cases=8\n",
	}
	log := strings.Join(lines, "")
	check := func(text string) error {
		writeFixture(t, root, "evidence.log", text)
		return CheckEvidence(root, filepath.Join(root, "evidence.log"), Evidence{ParserResource: true})
	}
	if err := check(log); err != nil {
		t.Fatal(err)
	}
	for _, line := range lines {
		for _, tc := range []struct{ name, replacement, want string }{
			{"missing", "", "missing parser resource execution"},
			{"zero", line[:strings.Index(line, "cases=")] + "cases=0\n", "incorrect case count"},
			{"reduced", strings.Replace(line, "cases=", "cases=1", 1), "incorrect case count"},
			{"duplicate", line + line, "duplicate"},
			{"commented", "skipped " + line, "missing parser resource execution"},
		} {
			t.Run(strings.TrimSpace(line)+"/"+tc.name, func(t *testing.T) {
				if err := check(strings.Replace(log, line, tc.replacement, 1)); err == nil || !strings.Contains(err.Error(), tc.want) {
					t.Fatalf("wanted %q, got %v", tc.want, err)
				}
			})
		}
	}
	if err := check(log + "PARSER-RESOURCE language=other cases=1\n"); err == nil {
		t.Fatal("unknown language accepted")
	}
	if err := check(log + "PARSER-RESOURCE language=go cases=invalid\n"); err == nil {
		t.Fatal("malformed duplicate result accepted")
	}
	writeFixture(t, root, "devtools/verification_anchors.json", `{"codec_parser_resource_cases":{"go":3}}`)
	if err := check(log); err == nil {
		t.Fatal("incomplete anchors accepted")
	}
}

func TestParserResourceGateAppliesToRequiredProfiles(t *testing.T) {
	plan, err := Load("../../verification_plan.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, profile := range []string{"codec-parser", "ci-portable", "full", "full-portable"} {
		for _, platform := range plan.Profiles[profile].Platforms {
			gates, err := plan.Select(profile, platform)
			if err != nil {
				t.Fatal(err)
			}
			found := false
			for _, gate := range gates {
				if gate.ID == "codec-parser-resource" && gate.Evidence.ParserResource {
					found = true
				}
			}
			if !found {
				t.Fatalf("parser resource gate missing from %s/%s", profile, platform)
			}
		}
	}
}
