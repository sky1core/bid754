package genmarker

import (
	"bytes"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestEmittedMarkersAreDiscoveredByCoverageChecker(t *testing.T) {
	root := t.TempDir()
	scripts := filepath.Join(root, "devtools", "scripts")
	if err := os.MkdirAll(scripts, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"generated_artifacts.py", "check_generated_marker_coverage.sh"} {
		body, err := os.ReadFile(filepath.Join("..", "..", "scripts", name))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(scripts, name), body, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(scripts, "generated_marker_exceptions.txt"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	markers := map[string]string{"line.go": Line("testgen"), "source.go": Line("go2rs from bid128_add.go"), "hash.py": HashLine("testgen"), "dot.go": DotLine("tools/codegen"), "flags.go": DotLine("tools/codegen --target=readtest-rust")}
	for name, marker := range markers {
		if err := os.WriteFile(filepath.Join(root, name), []byte(marker+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	manifest := filepath.Join(root, "devtools", "generated_artifacts.json")
	if err := os.WriteFile(manifest, []byte(`{"files":["line.go"],"directories":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	git := exec.Command("git", "init", "-q", root)
	if out, err := git.CombinedOutput(); err != nil {
		t.Fatalf("git init: %v %s", err, out)
	}
	check := func() ([]byte, error) {
		cmd := exec.Command("bash", filepath.Join(scripts, "check_generated_marker_coverage.sh"))
		cmd.Dir = root
		return cmd.CombinedOutput()
	}
	out, err := check()
	if err == nil || !bytes.Contains(out, []byte("dot.go")) || !bytes.Contains(out, []byte("hash.py")) || !bytes.Contains(out, []byte("source.go")) || !bytes.Contains(out, []byte("flags.go")) {
		t.Fatalf("unregistered emitted variants not discovered: %v\n%s", err, out)
	}
	if err := os.WriteFile(manifest, []byte(`{"files":["line.go","source.go","hash.py","dot.go","flags.go"],"directories":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if out, err := check(); err != nil {
		t.Fatalf("registered emitted variants rejected: %v\n%s", err, out)
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
