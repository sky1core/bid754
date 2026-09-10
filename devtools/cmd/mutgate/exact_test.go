package main

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestExactConfigRejectsAmbiguousBudgetsAndSeeds(t *testing.T) {
	valid := config{exactCampaigns: "relations,uniform-finite", exactSeeds: "1701,1702", exactCases: 300}
	configs, err := exactConfigs(valid)
	if err != nil || len(configs) != 4 {
		t.Fatalf("configs=%v err=%v", configs, err)
	}
	for _, change := range []func(*config){
		func(c *config) { c.exactSeeds = "" }, func(c *config) { c.exactSeeds = "1,1" }, func(c *config) { c.exactCPU = time.Second }, func(c *config) { c.exactCases = 0 }, func(c *config) { c.exactCampaigns = "relations,relations" }, func(c *config) { c.exactCPU = -1 },
	} {
		cfg := valid
		change(&cfg)
		if _, err := exactConfigs(cfg); err == nil {
			t.Errorf("accepted ambiguous config: %+v", cfg)
		}
	}
	valid.exactCases = 0
	valid.exactCPU = time.Millisecond
	if _, err := exactConfigs(valid); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(stageOrderDefault, []string{"readtest", "dectest", "parity"}) {
		t.Fatal("default stage chain changed")
	}
}

func TestExactProbeSitesAreRealAndDisjoint(t *testing.T) {
	root, err := gitTopLevel(".")
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, set := range []string{"calibration", "heldout"} {
		probes, err := exactProbes(set)
		if err != nil {
			t.Fatal(err)
		}
		for _, probe := range probes {
			site, src, err := exactProbeSite(root, probe)
			if err != nil {
				t.Fatal(err)
			}
			key := site.File + ":" + site.Func + ":" + site.OrigFull
			if seen[key] {
				t.Fatalf("heldout mutation overlaps calibration: %s", site.ID())
			}
			seen[key] = true
			if site.Offset < 0 || site.End > len(src) || site.OrigFull != string(src[site.Offset:site.End]) {
				t.Fatalf("invalid source span: %+v", site)
			}
		}
	}
	widthsSeen := map[string]bool{}
	widthsProbes, err := exactProbes("widths")
	if err != nil {
		t.Fatal(err)
	}
	for _, probe := range widthsProbes {
		site, src, err := exactProbeSite(root, probe)
		if err != nil {
			t.Fatal(err)
		}
		key := site.File + ":" + site.Func + ":" + site.OrigFull
		if widthsSeen[key] {
			t.Fatalf("widths mutation is not single-site: %s", site.ID())
		}
		widthsSeen[key] = true
		if site.Offset < 0 || site.End > len(src) || site.OrigFull != string(src[site.Offset:site.End]) {
			t.Fatalf("invalid source span: %+v", site)
		}
	}
}

func TestExactCampaignRealBinary(t *testing.T) {
	root, err := gitTopLevel(".")
	if err != nil {
		t.Fatal(err)
	}
	bin := filepath.Join(t.TempDir(), "finite.test")
	e := &engine{repo: root, goDir: filepath.Join(root, "bid754-go"), cfg: config{stageTimeout: 10 * time.Second}}
	build := exec.Command("go", e.buildArgs("portable", bin)...)
	build.Dir = e.goDir
	build.Env = e.buildEnv("portable")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build actual public test binary: %v\n%s", err, out)
	}
	for _, campaign := range []string{"relations", "uniform-finite"} {
		cfg := exactConfig{Campaign: campaign, Seed: 1701, Cases: 300}
		e.exact = &exactEvidence{Configs: []exactConfig{cfg}}
		verdict, out, _ := e.runExactStage(bin)
		if verdict != "pass" {
			t.Fatalf("%s baseline: %s\n%s", campaign, verdict, out)
		}
		first := e.exactRuns[0].Report
		if first.Executed != 300 || first.Mismatches != 0 {
			t.Fatalf("incomplete baseline: %+v", first)
		}
		verdict, out, _ = e.runExactStage(bin)
		if verdict != "pass" || e.exactRuns[0].Report.Digest != first.Digest {
			t.Fatalf("seed replay failed: %s %s", verdict, out)
		}
		alias := cfg
		alias.Seed += 2147483647
		e.exact.Configs = []exactConfig{alias}
		verdict, out, _ = e.runExactStage(bin)
		if verdict != "pass" || e.exactRuns[0].Report.Digest == first.Digest {
			t.Fatalf("distinct full-width seeds alias: %s %s", verdict, out)
		}
		cfg.Cases = 0
		cfg.CPUNS = int64(5 * time.Millisecond)
		e.exact.Configs = []exactConfig{cfg}
		verdict, out, _ = e.runExactStage(bin)
		if verdict != "pass" {
			t.Fatalf("CPU campaign: %s %s", verdict, out)
		}
		measured := e.exactRuns[0].Report
		t.Logf("%s requested_cpu_ns=%d measured_cpu_ns=%d cases=%d process_cpu_ns=%d", campaign, cfg.CPUNS, measured.CPUNS, measured.Executed, e.exactRuns[0].ProcessCPUNS)
		if measured.CPUNS < cfg.CPUNS || measured.Stop != "process_cpu" {
			t.Fatalf("CPU budget not measured: %+v", measured)
		}
	}
	for _, set := range []string{"calibration", "heldout", "widths"} {
		probes, _ := exactProbes(set)
		for _, p := range probes {
			e.exact.Configs = []exactConfig{{Campaign: "witness", Cases: 1, Witness: p.Name}}
			verdict, out, _ := e.runExactStage(bin)
			if verdict != "pass" {
				t.Fatalf("independent pinned witness %s: %s\n%s", p.Name, verdict, out)
			}
		}
	}
	e.exact.Configs = []exactConfig{{Campaign: "invalid", Cases: 1}}
	verdict, _, _ := e.runExactStage(bin)
	if verdict != "inconclusive" {
		t.Fatalf("configuration failure counted as %s", verdict)
	}
	fixture := buildOutcomeFixtureBinary(t)
	e.exact.Configs = []exactConfig{{Campaign: "relations", Seed: 1701, Cases: 1}}
	verdict, _, _ = e.runExactStage(fixture)
	if verdict != "inconclusive" {
		t.Fatalf("missing real campaign counted as %s", verdict)
	}
	e.exact = nil
	e.goDir = t.TempDir()
	verdict, out, _ := e.runStage(stageCatalog["dectest"], bin)
	if verdict != "inconclusive" || !strings.Contains(out, "pinned decTest input") {
		t.Fatalf("real decTest without inputs falsely passed: %s %s", verdict, out)
	}
}

func TestExactVerdictAggregationIsOrderIndependent(t *testing.T) {
	for _, a := range []string{"pass", "inconclusive", "timeout", "panic", "killed"} {
		for _, b := range []string{"pass", "inconclusive", "timeout", "panic", "killed"} {
			if aggregateExactVerdict(a, b) != aggregateExactVerdict(b, a) {
				t.Fatalf("order-dependent verdicts %s %s", a, b)
			}
		}
	}
	if aggregateExactVerdict("inconclusive", "killed") != "killed" {
		t.Fatal("confirmed arithmetic detection was hidden")
	}
}

func TestExactVerdictRequiresIntendedEvidence(t *testing.T) {
	cfg := exactConfig{Campaign: "witness", Cases: 1, Witness: "mul-inexact"}
	report := exactReport{Version: 1, Config: cfg, GeneratorVersion: 1, ModelVersion: 1, Executed: 1, Mismatches: 1, Digest: strings.Repeat("a", 64), CPUNS: 1, WallNS: 1, Stop: "cases", Lanes: map[string]int{"mul/d32/nearest_even": 1}}
	f := exactFinding{ExpectedRaw: "32800001", ActualRaw: "32800001", ExpectedFlags: 0x20, FlagsMismatch: true, Diagnostic: "flags differ"}
	f.Sample.Version = 1
	f.Sample.Family = "uniform-finite"
	f.Sample.Case.Width = 32
	f.Sample.Case.Op = "mul"
	f.Sample.Case.Mode = "nearest_even"
	f.Sample.Case.Operands = []string{"32800001", "32800001"}
	report.Findings = []exactFinding{f}
	data, _ := json.Marshal(report)
	evidence := "EXACTPROBE_REPORT " + string(data) + "\n    campaign_test.go:1: EXACTPROBE_ARITHMETIC_MISMATCH count=1\n--- FAIL: TestFiniteArithmeticCampaign (0.00s)\n"
	verdict, _, _ := exactVerdict(evidence, errors.New("exit 1"), false, cfg)
	if verdict != "killed" {
		t.Fatalf("precise arithmetic finding got %s", verdict)
	}
	for _, c := range []struct {
		out      string
		err      error
		deadline bool
		want     string
	}{
		{"--- FAIL: TestFiniteArithmeticCampaign (0.00s)\n", errors.New("exit 1"), false, "inconclusive"},
		{strings.ReplaceAll(evidence, "EXACTPROBE_ARITHMETIC_MISMATCH", "unrelated assertion"), errors.New("exit 1"), false, "inconclusive"},
		{evidence, nil, false, "inconclusive"},
		{evidence + "panic: broken kernel\n", errors.New("exit 2"), false, "panic"},
		{evidence, errors.New("deadline"), true, "timeout"},
	} {
		got, _, _ := exactVerdict(c.out, c.err, c.deadline, cfg)
		if got != c.want {
			t.Errorf("got %s want %s", got, c.want)
		}
	}
	report.Findings[0].FlagsMismatch = false
	data, _ = json.Marshal(report)
	if _, err := parseExactReport("EXACTPROBE_REPORT "+string(data), cfg); err == nil {
		t.Fatal("accepted non-arithmetic finding")
	}
}

func TestExactCLIAndSnapshotKeepPrimaryIndex(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "mutgate")
	cmd := exec.Command("go", "build", "-o", bin, ".")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build CLI: %v\n%s", err, out)
	}
	repo := gitFixtureRepo(t)
	mustGit := func(args ...string) string {
		t.Helper()
		s, err := gitText(repo, args...)
		if err != nil {
			t.Fatal(err)
		}
		return s
	}
	tracked := filepath.Join(repo, "bid754-go", "internal", "bidgo", "dummy.go")
	if err := os.WriteFile(tracked, []byte("package bidgo\nvar staged = 1\n"), 0644); err != nil {
		t.Fatal(err)
	}
	mustGit("add", "--", "bid754-go/internal/bidgo/dummy.go")
	if err := os.WriteFile(tracked, []byte("package bidgo\nvar working = 2\n"), 0644); err != nil {
		t.Fatal(err)
	}
	extra := "bid754-go/new_test.go"
	if err := os.WriteFile(filepath.Join(repo, extra), []byte("package bidgo\n"), 0644); err != nil {
		t.Fatal(err)
	}
	beforeIndex, err := os.ReadFile(filepath.Join(repo, ".git", "index"))
	if err != nil {
		t.Fatal(err)
	}
	beforeHead := mustGit("rev-parse", "HEAD")
	cmd = exec.Command(bin, "-mode", "snapshot", "-repo", repo, "-snapshot-files", extra)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("snapshot CLI: %v\n%s", err, out)
	}
	snapshot := strings.TrimSpace(string(out))
	afterIndex, err := os.ReadFile(filepath.Join(repo, ".git", "index"))
	if err != nil {
		t.Fatal(err)
	}
	if string(beforeIndex) != string(afterIndex) || mustGit("rev-parse", "HEAD") != beforeHead {
		t.Fatal("snapshot changed primary index or branch")
	}
	if mustGit("show", snapshot+":bid754-go/internal/bidgo/dummy.go") != "package bidgo\nvar working = 2" {
		t.Fatal("snapshot missed working-tree bytes")
	}
	if mustGit("show", snapshot+":"+extra) != "package bidgo" {
		t.Fatal("snapshot missed new file")
	}
	if mustGit("rev-parse", snapshot+"^") != beforeHead {
		t.Fatal("snapshot parent mismatch")
	}
	wt := filepath.Join(t.TempDir(), "not-created")
	cmd = exec.Command(bin, "-mode", "run", "-repo", repo, "-worktree", wt, "-stages", "exactprobe", "-jsonl", filepath.Join(t.TempDir(), "out.jsonl"))
	out, err = cmd.CombinedOutput()
	if err == nil || !strings.Contains(string(out), "explicit -exact-seeds") {
		t.Fatalf("CLI did not reject missing campaign seed: %v %s", err, out)
	}
	if _, err := os.Stat(wt); !os.IsNotExist(err) {
		t.Fatal("invalid exact configuration created a worktree")
	}
}

func TestExactCheckRejectsTuningSeedOverlapBeforeSetup(t *testing.T) {
	cfg := config{stages: "exactprobe", exactCampaigns: "relations", exactCases: 1, exactSeeds: "1701,1702", exactTuningSeeds: "754,1702", exactProbeSet: "heldout"}
	if err := runExactCheck(cfg); err == nil || !strings.Contains(err.Error(), "overlaps declared tuning seeds") {
		t.Fatalf("heldout overlap was not rejected before workspace setup: %v", err)
	}
}

func TestSnapshotStagedPaths(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "mutgate")
	if out, err := exec.Command("go", "build", "-o", bin, ".").CombinedOutput(); err != nil {
		t.Fatalf("build CLI: %v\n%s", err, out)
	}
	for _, kind := range []string{"regular", "split", "linked", "unmerged"} {
		t.Run(kind, func(t *testing.T) {
			repo := gitFixtureRepo(t)
			git := func(args ...string) string {
				t.Helper()
				out, err := gitText(repo, args...)
				if err != nil {
					t.Fatal(err)
				}
				return out
			}
			if kind == "linked" {
				linked := filepath.Join(t.TempDir(), "linked")
				git("worktree", "add", "--detach", linked, "HEAD")
				repo = linked
			}
			original := "bid754-go/internal/bidgo/dummy.go"
			added := "bid754-go/added \t\n_test.go"
			renamed := "bid754-go/internal/bidgo/renamed.go"
			untracked := "bid754-go/untracked_test.go"
			write := func(path, content string) {
				t.Helper()
				if err := os.WriteFile(filepath.Join(repo, path), []byte(content), 0644); err != nil {
					t.Fatal(err)
				}
			}
			write(added, "package bid754\nvar staged = 1\n")
			git("add", "--", added)
			write(added, "package bid754\nvar working = 2\n")
			git("mv", "--", original, renamed)
			write(renamed, "package bidgo\nvar renamedWorking = 3\n")
			git("update-index", "--force-remove", "--", "devtools/tests/.keep")
			write(untracked, "package bid754\n")
			if kind == "split" {
				git("update-index", "--split-index")
			}
			if kind == "unmerged" {
				blob := git("rev-parse", "HEAD:"+original)
				cmd := exec.Command("git", "-C", repo, "update-index", "--index-info")
				cmd.Stdin = strings.NewReader("0 " + strings.Repeat("0", len(blob)) + "\t" + renamed + "\n" +
					"100644 " + blob + " 1\t" + renamed + "\n" +
					"100644 " + blob + " 2\t" + renamed + "\n" +
					"100644 " + blob + " 3\t" + renamed + "\n")
				if out, err := cmd.CombinedOutput(); err != nil {
					t.Fatalf("prepare unmerged index: %v\n%s", err, out)
				}
			}
			index := git("rev-parse", "--git-path", "index")
			if !filepath.IsAbs(index) {
				index = filepath.Join(repo, index)
			}
			beforeIndex, err := os.ReadFile(index)
			if err != nil {
				t.Fatal(err)
			}
			beforeHead := git("rev-parse", "HEAD")
			out, snapshotErr := exec.Command(bin, "-mode", "snapshot", "-repo", repo).CombinedOutput()
			afterIndex, err := os.ReadFile(index)
			if err != nil {
				t.Fatal(err)
			}
			if string(beforeIndex) != string(afterIndex) || git("rev-parse", "HEAD") != beforeHead {
				t.Fatal("snapshot changed source index or HEAD")
			}
			if kind == "unmerged" {
				if snapshotErr == nil || !strings.Contains(string(out), "unmerged") {
					t.Fatalf("unmerged index was not rejected: %v\n%s", snapshotErr, out)
				}
				return
			}
			if snapshotErr != nil {
				t.Fatalf("snapshot CLI: %v\n%s", snapshotErr, out)
			}
			snapshot := strings.TrimSpace(string(out))
			for path, want := range map[string]string{
				added:   "package bid754\nvar working = 2",
				renamed: "package bidgo\nvar renamedWorking = 3",
			} {
				if got := git("show", snapshot+":"+path); got != want {
					t.Errorf("snapshot %q: got %q, want %q", path, got, want)
				}
			}
			for _, path := range []string{original, "devtools/tests/.keep", untracked} {
				if err := exec.Command("git", "-C", repo, "cat-file", "-e", snapshot+":"+path).Run(); err == nil {
					t.Errorf("snapshot included deleted or untracked path %q", path)
				}
			}
			if got := git("rev-parse", snapshot+"^"); got != beforeHead {
				t.Errorf("snapshot parent %s, want %s", got, beforeHead)
			}
		})
	}
}

func TestIntendedExactProbeRejectsOtherArithmeticFailures(t *testing.T) {
	probe := exactProbe{Name: "mul-midpoint"}
	finding := exactFinding{ValueMismatch: true, Rounding: "tie", ForbiddenRaw: "32800001"}
	finding.Sample.Case.Mode = "nearest_even"
	finding.Sample.Case.Width = 32
	run := exactRun{Config: exactConfig{Campaign: "witness", Witness: probe.Name}, Status: "killed", Report: &exactReport{Findings: []exactFinding{finding}}}
	if intendedExactFinding(probe, run) {
		t.Fatal("arbitrary arithmetic mismatch satisfied the intended witness")
	}
	run.Report.Findings[0].IntendedWitness = true
	if !intendedExactFinding(probe, run) {
		t.Fatal("precise midpoint witness was rejected")
	}
	run.Report.Findings[0].Sample.Case.Width = 128
	if intendedExactFinding(probe, run) {
		t.Fatal("another width credited to the witness")
	}
	run.Report.Findings[0].Sample.Case.Width = 32
	run.Status = "panic"
	if intendedExactFinding(probe, run) {
		t.Fatal("panic satisfied the intended witness")
	}
}

func TestExactBuildFaultRequiresCompilerEvidence(t *testing.T) {
	runErr := exec.Command("go", "tool", "compile", filepath.Join(t.TempDir(), "absent.go")).Run()
	if runErr == nil {
		t.Fatal("expected real compiler failure")
	}
	for _, c := range []struct {
		out        string
		contextErr error
		want       string
	}{
		{"internal/bidgo/bid32_mul.go:10:2: undefined: missing", nil, "compile_error"},
		{"go: cache permission denied", nil, "build_infrastructure_error"},
		{"internal/bidgo/bid32_mul.go:10:2: undefined: missing", errors.New("deadline"), "build_timeout"},
	} {
		if got := exactBuildFault(c.out, runErr, c.contextErr); got != c.want {
			t.Errorf("got %s want %s", got, c.want)
		}
	}
}

func TestExactCPUFindingsRequirePristinePrefix(t *testing.T) {
	cfg := exactConfig{Campaign: "relations", Seed: 1701, CPUNS: 1000}
	baseline := []exactRun{{Config: cfg, Status: "pass", Report: &exactReport{Executed: 10}}}
	report := &exactReport{Executed: 20, Findings: []exactFinding{{Index: 15}}}
	if exactBaselineCoversFinding(cfg, report, baseline) {
		t.Fatal("CPU budget extended past the pristine prefix but was counted as a mutation kill")
	}
	report.Findings[0].Index = 9
	if !exactBaselineCoversFinding(cfg, report, baseline) {
		t.Fatal("covered deterministic prefix finding was rejected")
	}
	cfg.CPUNS = 0
	cfg.Cases = 10
	baseline[0].Config = cfg
	baseline[0].Report.Digest = "pristine-inputs"
	report.Digest = "different-inputs"
	if exactBaselineCoversFinding(cfg, report, baseline) {
		t.Fatal("fixed cases with different inputs were counted as a mutation kill")
	}
}
