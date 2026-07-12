package genmarker

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestMarkerLinesMatchPattern proves every marker line this package can emit
// is discoverable by the coverage check's regex. A marker variant that stops
// matching Pattern would make its generator's outputs invisible to
// check_generated_marker_coverage.sh.
func TestMarkerLinesMatchPattern(t *testing.T) {
	re := regexp.MustCompile(Pattern)
	lines := map[string]string{
		"Line":               Line("testgen"),
		"Line with source":   Line("go2rs from bid128_add.go"),
		"HashLine":           HashLine("testgen"),
		"DotLine":            DotLine("tools/codegen"),
		"DotLine with flags": DotLine("tools/codegen --target=readtest-rust"),
	}
	for name, line := range lines {
		if strings.Contains(line, "\n") {
			t.Errorf("%s output %q spans multiple lines; markers must be a single line", name, line)
			continue
		}
		if !re.MatchString(line) {
			t.Errorf("%s output %q does not match coverage Pattern %q", name, line, Pattern)
		}
	}
}

// TestPatternMatchesCoverageScript pins Pattern to the marker_regex used by
// devtools/scripts/check_generated_marker_coverage.sh. If either side changes
// without the other, generators and the coverage check drift apart and newly
// generated artifacts can escape verification.
func TestPatternMatchesCoverageScript(t *testing.T) {
	scriptPath := filepath.Join("..", "..", "scripts", "check_generated_marker_coverage.sh")
	data, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("ReadFile(%q): %v", scriptPath, err)
	}
	const prefix = "marker_regex='"
	var scriptPattern string
	var found bool
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, prefix) {
			continue
		}
		if found {
			t.Fatalf("%s defines marker_regex more than once", scriptPath)
		}
		rest := strings.TrimPrefix(trimmed, prefix)
		if !strings.HasSuffix(rest, "'") {
			t.Fatalf("%s marker_regex line %q is not a simple single-quoted assignment", scriptPath, trimmed)
		}
		scriptPattern = strings.TrimSuffix(rest, "'")
		found = true
	}
	if !found {
		t.Fatalf("%s does not define marker_regex; the coverage check contract moved", scriptPath)
	}
	if scriptPattern != Pattern {
		t.Fatalf("marker regex drift:\n  script  (%s): %q\n  genmarker.Pattern: %q\nchange both sides together", scriptPath, scriptPattern, Pattern)
	}
}

// TestNoHardcodedMarkerLiteralsInEmitters walks the devtools module and fails
// if any non-test Go source outside this package hardcodes the "DO NOT EDIT"
// marker literal. Emitters must build marker lines via genmarker so a marker
// text change cannot fork per generator (the go2rs "Auto-generated" drift
// previously hid 104 generated Rust files from the coverage check).
func TestNoHardcodedMarkerLiteralsInEmitters(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve devtools root: %v", err)
	}
	selfDir, err := filepath.Abs(".")
	if err != nil {
		t.Fatalf("resolve genmarker dir: %v", err)
	}
	// Directories that hold generated outputs or third-party payloads rather
	// than generator sources.
	skipDirs := map[string]bool{
		"generated":   true,
		"third_party": true,
		"testdata":    true,
		"docker":      true,
	}
	marker := []byte("DO NOT EDIT")
	walkErr := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if skipDirs[d.Name()] || path == selfDir {
				return fs.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".go") || strings.HasSuffix(d.Name(), "_test.go") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if bytes.Contains(data, marker) {
			t.Errorf("%s hardcodes the generated-code marker literal; emit it via devtools/internal/genmarker instead", path)
		}
		return nil
	})
	if walkErr != nil {
		t.Fatalf("walk %s: %v", root, walkErr)
	}
}
