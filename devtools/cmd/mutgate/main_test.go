package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestIntLitDelta(t *testing.T) {
	cases := []struct {
		lit   string
		delta int64
		want  string
		ok    bool
	}{
		{"5", 1, "6", true},
		{"5", -1, "4", true},
		{"0", -1, "(-1)", true},
		{"0", 1, "1", true},
		{"0x1f", 1, "0x20", true},
		{"0x0", -1, "(-1)", true},
		{"0xffffffffffffffff", 1, "", false},
		{"0Xff", -1, "0Xfe", true},
		{"0b101", 1, "0b110", true},
		{"0o17", 1, "0o20", true},
		{"017", 1, "020", true},
		{"1_000", 1, "1001", true},
		{"9999999", 1, "10000000", true},
	}
	for _, c := range cases {
		got, ok := intLitDelta(c.lit, c.delta)
		if ok != c.ok || got != c.want {
			t.Errorf("intLitDelta(%q,%d) = %q,%v want %q,%v", c.lit, c.delta, got, ok, c.want, c.ok)
		}
	}
}

// gitFixtureRepo builds a minimal committed repository shaped like the paths
// mutgate touches: a bid754-go tree (mutation surface) and a devtools/tests
// dir (decTest copy source read by setupWorktree).
func gitFixtureRepo(t *testing.T) string {
	t.Helper()
	repo := t.TempDir()
	mustGit := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", repo}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	mustGit("init", "-q")
	mustGit("config", "user.email", "mutgate-test@example.invalid")
	mustGit("config", "user.name", "mutgate test")
	if err := os.MkdirAll(filepath.Join(repo, "bid754-go", "internal", "bidgo"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(repo, "devtools", "tests"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, "bid754-go", "internal", "bidgo", "dummy.go"), []byte("package bidgo\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, "devtools", "tests", ".keep"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	mustGit("add", "-A")
	mustGit("commit", "-q", "-m", "fixture")
	return repo
}

// TestSetupWorktreeRejectsDirtyReuse reproduces the failure mode where a
// leftover mutation from an interrupted run silently poisoned every later
// verdict: reusing a dirty worktree must now be a hard setup error.
func TestSetupWorktreeRejectsDirtyReuse(t *testing.T) {
	repo := gitFixtureRepo(t)
	worktree := filepath.Join(t.TempDir(), "mutwork")
	cfg := config{repo: repo, worktree: worktree, commit: "HEAD"}

	if err := setupWorktree(cfg); err != nil {
		t.Fatalf("initial setup: %v", err)
	}
	// Clean reuse stays allowed.
	if err := setupWorktree(cfg); err != nil {
		t.Fatalf("clean reuse should succeed: %v", err)
	}
	// Simulate an interrupted run: a mutation is still applied in the reused
	// worktree.
	mutated := filepath.Join(worktree, "bid754-go", "internal", "bidgo", "dummy.go")
	if err := os.WriteFile(mutated, []byte("package bidgo // mutated leftover\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := setupWorktree(cfg)
	if err == nil {
		t.Fatal("setupWorktree reused a dirty worktree without error; leftover mutations would poison every mutant verdict")
	}
	if !strings.Contains(err.Error(), "dirty worktree") {
		t.Fatalf("setupWorktree dirty-reuse error should name the dirty worktree, got: %v", err)
	}
}

// TestCheckWorktreeCleanAfterRunFailsOnLeftoverMutation reproduces the
// failure mode where a dirty exit only printed a WARNING and returned exit
// code 0: the post-run clean check must now return an error that propagates
// to a non-zero process exit.
func TestCheckWorktreeCleanAfterRunFailsOnLeftoverMutation(t *testing.T) {
	repo := gitFixtureRepo(t)
	if err := checkWorktreeCleanAfterRun(repo); err != nil {
		t.Fatalf("clean tree should pass: %v", err)
	}
	mutated := filepath.Join(repo, "bid754-go", "internal", "bidgo", "dummy.go")
	if err := os.WriteFile(mutated, []byte("package bidgo // leftover\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := checkWorktreeCleanAfterRun(repo)
	if err == nil {
		t.Fatal("checkWorktreeCleanAfterRun accepted a tree with a leftover mutation")
	}
	if !strings.Contains(err.Error(), "not clean after run") {
		t.Fatalf("unexpected error text: %v", err)
	}
}

// buildFixtureTestBinary compiles a one-test package so runStage exercises a
// real `go test -c` binary, exactly like production stage runs.
func buildFixtureTestBinary(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module mutgatefixture\n\ngo 1.21\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	src := "package fixture\n\nimport \"testing\"\n\nfunc TestAlpha(t *testing.T) {}\n"
	if err := os.WriteFile(filepath.Join(dir, "fixture_test.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	bin := filepath.Join(dir, "fixture.test")
	cmd := exec.Command("go", "test", "-c", "-o", bin, ".")
	cmd.Dir = dir
	cmd.Env = append(append([]string{}, os.Environ()...), "GOFLAGS=", "CGO_ENABLED=0")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build fixture test binary: %v\n%s", err, out)
	}
	return bin
}

// TestRunStageFailsWhenRunExprMatchesNoTests reproduces the failure mode
// where a -test.run expression selecting zero tests exited 0 and was counted
// as a stage pass (miscounting the mutant as survived): zero selected tests
// must now yield the "nomatch" verdict, and a matching expression must still
// pass.
func TestRunStageFailsWhenRunExprMatchesNoTests(t *testing.T) {
	bin := buildFixtureTestBinary(t)
	e := &engine{
		cfg:   config{stageTimeout: 60 * time.Second},
		goDir: filepath.Dir(bin),
	}

	verdict, note, _ := e.runStage(stage{Name: "probe", Binary: "portable", RunExpr: "^TestDoesNotExist$"}, bin)
	if verdict != "nomatch" {
		t.Fatalf("zero-match run expression: verdict = %q (note %q), want \"nomatch\"", verdict, note)
	}

	verdict, note, _ = e.runStage(stage{Name: "probe", Binary: "portable", RunExpr: "^TestAlpha$"}, bin)
	if verdict != "pass" {
		t.Fatalf("matching run expression: verdict = %q (note %q), want \"pass\"", verdict, note)
	}
}

func TestCountListedTests(t *testing.T) {
	if got := countListedTests(""); got != 0 {
		t.Fatalf("empty list output counted %d tests", got)
	}
	if got := countListedTests("TestAlpha\nTestBeta\n"); got != 2 {
		t.Fatalf("two-test list output counted %d tests", got)
	}
}

func TestParseStrata(t *testing.T) {
	q, order, err := parseStrata("aor=5,cmp=4")
	if err != nil || q["aor"] != 5 || q["cmp"] != 4 || len(order) != 2 || order[0] != "aor" {
		t.Fatalf("parseStrata: q=%v order=%v err=%v", q, order, err)
	}
	if _, _, err := parseStrata("bogus"); err == nil {
		t.Fatal("parseStrata should reject entries without '='")
	}
	q, order, err = parseStrata("")
	if q != nil || order != nil || err != nil {
		t.Fatalf("empty strata should be nil,nil,nil; got %v %v %v", q, order, err)
	}
}
