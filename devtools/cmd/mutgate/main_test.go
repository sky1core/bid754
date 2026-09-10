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
	mustGit("add", "--", "bid754-go/internal/bidgo/dummy.go", "devtools/tests/.keep")
	mustGit("commit", "-qm", "fixture")
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

// buildOutcomeFixtureBinary compiles a package whose tests reproduce every
// distinct stage outcome mutgate must classify — an assertion failure, a
// panic, a self-timeout, and a nonzero exit with no failing-test evidence —
// so runStage is exercised against real `go test -c` binaries rather than a
// hand-faked production backend.
func buildOutcomeFixtureBinary(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module mutgateoutcome\n\ngo 1.21\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	src := `package fixture

import (
	"flag"
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	flag.Parse()
	if flag.Lookup("test.run").Value.String() == "^TestHangs$" {
		if err := flag.Set("test.timeout", "20ms"); err != nil { os.Exit(2) }
	}
	os.Exit(m.Run())
}

func TestPasses(t *testing.T) {}

func TestSkips(t *testing.T) { t.Skip("required input absent") }

func TestAssertMismatch(t *testing.T) {
	t.Errorf("value mismatch: got 0x01 want 0x02")
}

func TestPanics(t *testing.T) {
	panic("boom from mutant")
}

func TestHangs(t *testing.T) {
	time.Sleep(30 * time.Second)
}

func TestExitsWithoutDiagnostic(t *testing.T) {
	os.Exit(3)
}
`
	if err := os.WriteFile(filepath.Join(dir, "fixture_test.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	bin := filepath.Join(dir, "fixture.test")
	cmd := exec.Command("go", "test", "-c", "-o", bin, ".")
	cmd.Dir = dir
	cmd.Env = append(append([]string{}, os.Environ()...), "GOFLAGS=", "CGO_ENABLED=0")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build outcome fixture binary: %v\n%s", err, out)
	}
	return bin
}

// TestRunStageClassifiesRealTestOutcomes drives runStage against a real
// compiled test binary and asserts each outcome maps to the correct verdict:
// a pass, an assertion failure ("killed" — the only class that means a
// value/flag mismatch), a panic and a self-timeout (detections kept distinct
// from an assertion kill), and a bare nonzero exit ("inconclusive", never
// mistaken for a mismatch).
func TestRunStageClassifiesRealTestOutcomes(t *testing.T) {
	bin := buildOutcomeFixtureBinary(t)
	e := &engine{
		cfg:   config{stageTimeout: 30 * time.Second},
		goDir: filepath.Dir(bin),
	}
	cases := []struct {
		run  string
		want string
	}{
		{"^TestPasses$", "pass"},
		{"^TestSkips$", "inconclusive"},
		{"^TestAssertMismatch$", "killed"},
		{"^TestPanics$", "panic"},
		{"^TestHangs$", "timeout"},
		{"^TestExitsWithoutDiagnostic$", "inconclusive"},
	}
	for _, c := range cases {
		verdict, out, _ := e.runStage(stage{Name: "probe", Binary: "portable", RunExpr: c.run}, bin)
		if verdict != c.want {
			t.Errorf("runStage %s: verdict = %q, want %q\noutput:\n%s", c.run, verdict, c.want, out)
			continue
		}
		switch verdict {
		case "timeout":
			if !strings.Contains(out, "panic: test timed out after 20ms") {
				t.Errorf("runtime timeout diagnostic missing; an outer deadline is insufficient: %s", out)
			}
		case "killed":
			if lines := failLines(out); len(lines) == 0 {
				t.Errorf("runStage %s: killed verdict carried no failing-test diagnostic (fail lines empty)", c.run)
			}
		case "panic":
			// The panic banner must survive as evidence even though it precedes
			// a long goroutine dump that a tail would drop.
			found := false
			for _, l := range failLines(out) {
				if strings.HasPrefix(l, "panic:") {
					found = true
				}
			}
			if !found {
				t.Errorf("runStage %s: panic verdict did not preserve the panic banner in fail lines: %v", c.run, failLines(out))
			}
		}
	}
}

func TestRunStageRejectsInheritedShardSelection(t *testing.T) {
	e := &engine{cfg: config{stageTimeout: time.Second}}
	for _, key := range []string{"BID754_TIER1_ARITH_SHARD_COUNT", "BID754_TIER1_COMPARE_CONVERSION_SHARD_INDEX", "BID754_D32_EXHAUSTIVE_SHARD_COUNT"} {
		t.Run(key, func(t *testing.T) {
			t.Setenv(key, "18446744073709551615")
			verdict, diagnostic, _ := e.runStage(stage{Name: "tier1rand"}, "must-not-execute")
			if verdict != "inconclusive" || !strings.Contains(diagnostic, "unsharded execution") || !strings.Contains(diagnostic, key) {
				t.Fatalf("%s: %s", verdict, diagnostic)
			}
		})
	}
}

func TestClassifyStageFailure(t *testing.T) {
	cases := []struct {
		name string
		out  string
		want string
	}{
		{"assertion", "=== RUN   TestX\n    x_test.go:9: got 1 want 2\n--- FAIL: TestX (0.00s)\nFAIL\n", "killed"},
		{"panic", "panic: boom\n\ngoroutine 1 [running]:\nmain.main()\n", "panic"},
		{"runtime-fatal", "fatal error: concurrent map writes\n\ngoroutine 5 [running]:\n", "panic"},
		{"self-timeout", "panic: test timed out after 2s\n\ngoroutine 1 [running]:\n", "timeout"},
		{"bare-exit", "", "inconclusive"},
		{"loader-error", "dyld[123]: Library not loaded: libfoo.dylib\n", "inconclusive"},
		// A panic banner and a --- FAIL line together (as a crashing test emits)
		// still classifies as a crash, not an assertion mismatch.
		{"panic-with-fail-line", "--- FAIL: TestX (0.00s)\npanic: boom [recovered]\n\tpanic: boom\n", "panic"},
		// "panic:" appearing only inside an assertion message (not line-anchored)
		// stays an assertion kill.
		{"fail-message-mentions-panic", "    x_test.go:9: expected no panic: got one\n--- FAIL: TestX (0.00s)\n", "killed"},
	}
	for _, c := range cases {
		if got := classifyStageFailure(c.out); got != c.want {
			t.Errorf("classifyStageFailure(%s) = %q, want %q", c.name, got, c.want)
		}
	}
}

func TestFailLinesCapturesBanners(t *testing.T) {
	lines := failLines("--- FAIL: TestX (0.00s)\n    x_test.go:9: got 1 want 2\n")
	if len(lines) == 0 || !strings.HasPrefix(lines[0], "--- FAIL") {
		t.Fatalf("assertion fail lines = %v", lines)
	}
	lines = failLines("panic: boom\n\ngoroutine 1 [running]:\n")
	if len(lines) == 0 || !strings.HasPrefix(lines[0], "panic:") {
		t.Fatalf("panic fail lines = %v", lines)
	}
	lines = failLines("panic: test timed out after 2s\n\ngoroutine 1 [running]:\n")
	if len(lines) == 0 || !strings.HasPrefix(lines[0], "panic:") {
		t.Fatalf("timeout fail lines = %v", lines)
	}
}

func TestHasLinePrefix(t *testing.T) {
	if !hasLinePrefix("foo\npanic: boom\n", "panic:") {
		t.Error("hasLinePrefix missed a line-anchored panic banner")
	}
	if hasLinePrefix("expected no panic: got one\n", "panic:") {
		t.Error("hasLinePrefix matched a non-anchored occurrence")
	}
	if !hasLinePrefix("\t  panic: boom", "panic:") {
		t.Error("hasLinePrefix should ignore leading whitespace")
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

// TestEveryStageBinaryHasBuildSpec closes the stage-catalog/build-table world.
//
// The failure this blocks is silent, not loud: a stage naming a test binary
// with no build spec used to fall through to the untagged default build, and
// under the wrong tags a long gate compiles its stub, whose test *skips*. A
// skipping test exits 0, so mutgate would score every mutant "survived"
// against a gate that never executed a single case -- a fabricated survivor
// list with no visible error anywhere in the run.
func TestEveryStageBinaryHasBuildSpec(t *testing.T) {
	for name, st := range stageCatalog {
		if st.RunExpr == "" {
			t.Errorf("stage %q has an empty -test.run expression", name)
		}
		if _, ok := binaryBuilds[st.Binary]; !ok {
			t.Errorf("stage %q names test binary %q with no entry in binaryBuilds", name, st.Binary)
		}
	}
	for key, b := range binaryBuilds {
		if b.Pkg == "" {
			t.Errorf("binaryBuilds[%q] has no package", key)
		}
		used := false
		for _, st := range stageCatalog {
			if st.Binary == key {
				used = true
				break
			}
		}
		if !used {
			t.Errorf("binaryBuilds[%q] is unused; a build spec no stage names is dead weight "+
				"that hides which gates the tool can actually run", key)
		}
	}
}

// TestResolveStagesRejectsMissingBuildSpec exercises the guard through the
// flag-parsing path a run actually takes.
//
// It installs a temporary entry in the package-level stageCatalog, which is
// safe only because no test in this package calls t.Parallel; adding a
// parallel test here would race against this mutation.
func TestResolveStagesRejectsMissingBuildSpec(t *testing.T) {
	const probe = "probe_missing_build_spec"
	stageCatalog[probe] = stage{Name: probe, Binary: "no_such_binary", RunExpr: "^TestNothing$"}
	defer delete(stageCatalog, probe)

	if _, err := resolveStages(config{stages: probe}); err == nil {
		t.Fatal("resolveStages accepted a stage whose test binary has no build spec")
	}
	if _, err := resolveStages(config{stages: "readtest"}); err != nil {
		t.Fatalf("resolveStages rejected a valid stage: %v", err)
	}
}

// TestStubBackedBinariesCarryTheirGateTags pins the build tags of every test
// binary whose gate has a skipping stub compiled under the complementary
// constraint.
//
// Only the tags separate the real gate from its stub, and a stub is the one
// wrong-tag outcome mutgate cannot notice: it defines the same test name, so
// the -test.list preflight still matches, the stub skips, the binary exits 0,
// and every mutant is scored "survived" against a gate that ran nothing.
// tier1long is deliberately absent -- the Tier 1 long gate ships no stub, so
// wrong tags there leave the test undefined and the preflight reports
// "nomatch" instead of a silent pass.
func TestStubBackedBinariesCarryTheirGateTags(t *testing.T) {
	required := map[string][]string{
		"native":        {"bid754_native"},
		"decnumber":     {"bid754_native", "bid754_decnumber_diff"},
		"d32exhaustive": {"bid754_native", "bid754_d32_exhaustive"},
	}
	for binary, tags := range required {
		spec, ok := binaryBuilds[binary]
		if !ok {
			t.Errorf("stub-backed binary %q has no build spec", binary)
			continue
		}
		// Exact identifier match, not substring: Go resolves build tags as whole
		// identifiers, so "bid754_native_wrong" or "bid754_d32_exhaustiveX"
		// satisfies a Contains check while still compiling the stub -- the very
		// silent outcome this test exists to block.
		have := map[string]bool{}
		for _, tg := range strings.Split(spec.Tags, ",") {
			have[strings.TrimSpace(tg)] = true
		}
		for _, tag := range tags {
			if !have[tag] {
				t.Errorf("binary %q builds without the exact tag %q, which compiles the gate's "+
					"skipping stub; tags are %q", binary, tag, spec.Tags)
			}
		}
	}
	// The stage that motivated the table must still resolve to one of them.
	st, ok := stageCatalog["d32exh"]
	if !ok {
		t.Fatal("d32exh stage missing from the catalog")
	}
	if _, ok := required[st.Binary]; !ok {
		t.Fatalf("d32exh names binary %q, which is not covered by the stub-backed tag table", st.Binary)
	}
}

func TestSelfCheckRequiresItsArithmeticDiagnostic(t *testing.T) {
	spec := selfChecks[0]
	for _, result := range []mutantResult{
		{Status: "killed", FailLines: []string{"--- FAIL: TestGeneratedDecnumberDifferentialStructured"}},
		{Status: "killed", FailLines: []string{"DIVERGENCE d128 mul mode=0"}},
		{Status: "panic", FailLines: []string{"DIVERGENCE d128 fma mode=0"}},
	} {
		if selfCheckDiagnosticMatches(spec, result) {
			t.Fatalf("accepted unintended failure: %+v", result)
		}
	}
	result := mutantResult{Status: "killed", FailLines: []string{"generated.go:42: DIVERGENCE d128 fma mode=0"}}
	if !selfCheckDiagnosticMatches(spec, result) {
		t.Fatal("rejected intended arithmetic diagnostic")
	}
}
