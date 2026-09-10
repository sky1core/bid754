package verification

import (
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

func verificationFixture(t *testing.T) (string, Plan) {
	t.Helper()
	t.Setenv("BID754_SNAPSHOT_ARCHIVE", "")
	t.Setenv("BID754_SNAPSHOT_ID", "")
	t.Setenv("GOFLAGS", "")
	root := t.TempDir()
	repo, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("git", "init", "--quiet", root)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init: %v %s", err, out)
	}
	for _, rel := range []string{"devtools/scripts/lib/worktree_files.py", "devtools/scripts/lib/source_snapshot.py", "devtools/scripts/check_scripts.py"} {
		raw, err := os.ReadFile(filepath.Join(repo, rel))
		if err != nil {
			t.Fatal(err)
		}
		writeFixture(t, root, rel, string(raw))
	}
	writeFixture(t, root, ".gitignore", "test_results/\n")
	writeFixture(t, root, "devtools/verification_anchors.json", "{}\n")
	writeFixture(t, root, "first.sh", "#!/usr/bin/env bash\ntrue\n")
	writeFixture(t, root, "second.sh", "#!/usr/bin/env bash\ntrue\n")
	writeFixture(t, root, "Makefile", "check-scripts:\n\t@python3 -B devtools/scripts/check_scripts.py\n")
	if err := os.MkdirAll(filepath.Join(root, "test_results"), 0700); err != nil {
		t.Fatal(err)
	}
	p := Plan{Version: 1, Profiles: map[string]Profile{"scripts": {Groups: []string{"scripts"}, Scope: "profile", Platforms: []string{runtime.GOOS + "/" + runtime.GOARCH}}}, Gates: []Gate{{ID: "shell", Target: "check-scripts", Group: "scripts", Comparison: "syntax", Evidence: Evidence{Patterns: []string{`(?m)^SCRIPT-SYNTAX-CHECK checked=2$`}}}}}
	return root, p
}

func writeFixture(t *testing.T, root, rel, text string) {
	t.Helper()
	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(text), 0600); err != nil {
		t.Fatal(err)
	}
}

func TestRunnerRejectsNoOpAndDetectsTheIntendedSyntaxDefect(t *testing.T) {
	root, p := verificationFixture(t)
	baseline, err := Run(root, p, "scripts", filepath.Join(root, "test_results/baseline"), "audit", io.Discard)
	if err != nil || baseline.Status != "passed" {
		t.Fatalf("baseline: %v %+v", err, baseline)
	}
	writeFixture(t, root, "second.sh", "#!/usr/bin/env bash\nif\n")
	failed, err := Run(root, p, "scripts", filepath.Join(root, "test_results/mutant"), "audit", io.Discard)
	if err == nil || failed.Status != "failed" || len(failed.Gates) != 1 {
		t.Fatalf("mutant was not rejected: %v %+v", err, failed)
	}
	log, readErr := os.ReadFile(filepath.Join(root, "test_results/mutant/shell.log"))
	if readErr != nil {
		t.Fatal(readErr)
	}
	if !strings.Contains(string(log), "shell syntax check failed: second.sh") || !strings.Contains(string(log), "syntax error") {
		t.Fatalf("mutant failed for unrelated reason: %s", log)
	}
	writeFixture(t, root, "Makefile", "check-scripts:\n\t@true\n")
	noop, err := Run(root, p, "scripts", filepath.Join(root, "test_results/noop"), "audit", io.Discard)
	if err == nil || !strings.Contains(err.Error(), "missing execution evidence") || noop.Gates[0].ExitCode != 0 {
		t.Fatalf("zero exit without execution evidence was not rejected: %v %+v", err, noop)
	}
}

func TestRunnerRejectsInheritedTestSelection(t *testing.T) {
	root, p := verificationFixture(t)
	baseline, err := Run(root, p, "scripts", filepath.Join(root, "test_results/baseline"), "configuration", io.Discard)
	if err != nil || baseline.Status != "passed" {
		t.Fatalf("baseline: %v", err)
	}
	for i, flags := range []string{"-skip=^TestConcurrent", "-run=^$", "-short", "-tags=bid754_native"} {
		t.Setenv("GOFLAGS", flags)
		failed, err := Run(root, p, "scripts", filepath.Join(root, "test_results/filtered-"+strconv.Itoa(i)), "configuration", io.Discard)
		if err == nil || !strings.Contains(err.Error(), "require empty GOFLAGS") || len(failed.Gates) != 0 || failed.Configuration["GOFLAGS"] != flags {
			t.Fatalf("noncanonical scope %q accepted: %v %+v", flags, err, failed)
		}
	}
}

func TestAdjudicatorRejectsMissingMixedStaleAndWeakenedEvidence(t *testing.T) {
	root, p := verificationFixture(t)
	dir := filepath.Join(root, "test_results/baseline")
	baseline, err := Run(root, p, "scripts", dir, "invocation-1", io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(baseline)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, want string
		mutate     func(*Result)
	}{
		{"missing gate", "missing required gates", func(r *Result) { r.Gates = nil }},
		{"mixed invocation", "mismatched verification invocation", func(r *Result) { r.Invocation = "invocation-0" }},
		{"mixed gate run", "identity or comparison mismatch", func(r *Result) { r.Gates[0].RunID = "other" }},
		{"wrong snapshot", "snapshot mismatch", func(r *Result) { r.SnapshotID = strings.Repeat("a", 64) }},
		{"stale plan", "execution plan mismatch", func(r *Result) { r.PlanHash = strings.Repeat("a", 64) }},
		{"stale corpus", "anchors mismatch", func(r *Result) { r.AnchorsHash = strings.Repeat("a", 64) }},
		{"weakened comparator", "identity or comparison mismatch", func(r *Result) { r.Gates[0].Comparison = "exit-only" }},
		{"partial promoted to full", "scope mismatch", func(r *Result) { r.Scope = "full" }},
		{"missing configuration", "Go verification configuration", func(r *Result) { r.Configuration = nil }},
		{"filtered configuration", "Go verification configuration", func(r *Result) { r.Configuration["GOFLAGS"] = "-skip=^TestConcurrent" }},
		{"missing tool", "missing tool version", func(r *Result) { delete(r.Tools, "go") }},
		{"log traversal", "invalid log path", func(r *Result) { r.Gates[0].Log = "../shell.log" }},
		{"log altered", "log hash mismatch", func(r *Result) { r.Gates[0].LogHash = strings.Repeat("a", 64) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var candidate Result
			if err := json.Unmarshal(encoded, &candidate); err != nil {
				t.Fatal(err)
			}
			tc.mutate(&candidate)
			if err := Validate(root, p, dir, candidate, baseline.SnapshotID, "invocation-1"); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("want %q, got %v", tc.want, err)
			}
		})
	}
	logPath := filepath.Join(dir, "shell.log")
	raw, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	raw = []byte(strings.ReplaceAll(string(raw), "run="+baseline.RunID, "run=old-run"))
	if err := os.WriteFile(logPath, raw, 0600); err != nil {
		t.Fatal(err)
	}
	baseline.Gates[0].LogHash = Hash(raw)
	if err := Validate(root, p, dir, baseline, baseline.SnapshotID, "invocation-1"); err == nil || !strings.Contains(err.Error(), "stale log identity") {
		t.Fatalf("stale log accepted with updated checksum: %v", err)
	}
}

func TestCanonicalPlanDoesNotPromotePartialOrRequireExhaustiveByDefault(t *testing.T) {
	p, err := Load("../../verification_plan.json")
	if err != nil {
		t.Fatal(err)
	}
	if p.Profiles["full-portable"].Scope != "partial" {
		t.Fatal("native omission is not classified partial")
	}
	full, err := p.Select("full", "linux/amd64")
	if err != nil {
		t.Fatal(err)
	}
	ids := map[string]bool{}
	for _, g := range full {
		if strings.Contains(g.Target, "exhaustive") {
			t.Fatalf("long exhaustive gate inserted into routine closure: %s", g.ID)
		}
		ids[g.ID] = true
	}
	for _, id := range []string{"native-decnumber", "rust-native-fuzz", "codec-vectors", "codec-parser-resource", "generated-artifacts"} {
		if !ids[id] {
			t.Fatalf("missing required gate %s", id)
		}
	}
	if _, err := p.Select("native", "linux/arm64"); err == nil {
		t.Fatal("unsupported native profile accepted")
	}
}

func TestMatrixRejectsMissingRequiredPlatform(t *testing.T) {
	root, p := verificationFixture(t)
	p.Matrix = []RequiredRun{{Profile: "scripts", Platform: runtime.GOOS + "/" + runtime.GOARCH}}
	dir := filepath.Join(root, "test_results/baseline")
	result, err := Run(root, p, "scripts", dir, "matrix", io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateMatrix(root, p, filepath.Join(root, "test_results"), result.SnapshotID, "matrix"); err != nil {
		t.Fatal(err)
	}
	empty := t.TempDir()
	if err := ValidateMatrix(root, p, empty, result.SnapshotID, "matrix"); err == nil || !strings.Contains(err.Error(), "missing required matrix result") {
		t.Fatalf("missing platform accepted: %v", err)
	}
}
