package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type exactConfig struct {
	Campaign string `json:"campaign"`
	Seed     int64  `json:"seed"`
	Cases    int    `json:"cases"`
	CPUNS    int64  `json:"cpu_ns"`
	Witness  string `json:"witness,omitempty"`
}

type exactFinding struct {
	Index  int `json:"index"`
	Sample struct {
		Version int    `json:"version"`
		Family  string `json:"family"`
		Case    struct {
			Width    int      `json:"width"`
			Op       string   `json:"op"`
			Mode     string   `json:"mode"`
			Operands []string `json:"operands"`
		} `json:"case"`
	} `json:"sample"`
	ExpectedRaw     string `json:"expected_raw"`
	ExpectedFlags   uint32 `json:"expected_flags"`
	ActualRaw       string `json:"actual_raw"`
	ActualFlags     uint32 `json:"actual_flags"`
	ValueMismatch   bool   `json:"value_mismatch"`
	FlagsMismatch   bool   `json:"flags_mismatch"`
	Rounding        string `json:"rounding"`
	Diagnostic      string `json:"diagnostic"`
	IntendedWitness bool   `json:"intended_witness"`
	ForbiddenRaw    string `json:"forbidden_raw,omitempty"`
	ForbiddenFlags  uint32 `json:"forbidden_flags,omitempty"`
}

type exactReport struct {
	Version          int            `json:"version"`
	Config           exactConfig    `json:"config"`
	GeneratorVersion int            `json:"generator_version"`
	ModelVersion     int            `json:"model_version"`
	Executed         int            `json:"executed"`
	Lanes            map[string]int `json:"lanes"`
	Mismatches       int            `json:"mismatches"`
	Findings         []exactFinding `json:"findings"`
	Digest           string         `json:"input_sha256"`
	CPUNS            int64          `json:"cpu_ns"`
	WallNS           int64          `json:"wall_ns"`
	Stop             string         `json:"stop"`
}

type exactRun struct {
	Config       exactConfig  `json:"config"`
	Status       string       `json:"status"`
	Report       *exactReport `json:"report,omitempty"`
	ProcessCPUNS int64        `json:"process_cpu_ns"`
	WallNS       int64        `json:"wall_ns"`
	Diagnostic   string       `json:"diagnostic,omitempty"`
}

type exactEvidence struct {
	SourceCommit string        `json:"source_commit"`
	SourceTree   string        `json:"source_tree"`
	GeneratorSHA string        `json:"generator_sha256"`
	ModelSHA     string        `json:"model_sha256"`
	HarnessSHA   string        `json:"harness_sha256"`
	ToolSHA      string        `json:"mutgate_sha256"`
	GoVersion    string        `json:"go_version"`
	Platform     string        `json:"platform"`
	MutationID   string        `json:"mutation_id"`
	Selection    string        `json:"selection"`
	BudgetScope  string        `json:"budget_scope"`
	Comparison   string        `json:"comparison"`
	ProbeSet     string        `json:"probe_set,omitempty"`
	TuningSeeds  []int64       `json:"declared_tuning_seeds,omitempty"`
	Configs      []exactConfig `json:"campaigns"`
	Baseline     []exactRun    `json:"baseline"`
	Runs         []exactRun    `json:"runs,omitempty"`
}

func registerExactFlags(cfg *config) {
	flag.StringVar(&cfg.exactCampaigns, "exact-campaigns", "relations,uniform-finite", "exactprobe campaigns, each executed independently for every seed")
	flag.StringVar(&cfg.exactSeeds, "exact-seeds", "", "exactprobe explicit comma-separated distinct seeds (required)")
	flag.StringVar(&cfg.exactTuningSeeds, "exact-tuning-seeds", "0,754", "exactcheck declared generator-tuning seeds; heldout seeds must be disjoint; extend to include all seeds used for tuning")
	flag.IntVar(&cfg.exactCases, "exact-cases", 300, "exactprobe cases per campaign/seed; set 0 for CPU budget")
	flag.DurationVar(&cfg.exactCPU, "exact-cpu-budget", 0, "opt-in process user+system CPU per campaign/seed, excluding compilation; requires -exact-cases=0; wall timeout is only a safety limit")
	flag.StringVar(&cfg.exactProbeSet, "exact-probes", "calibration", "exactcheck probe set: calibration|heldout (distinct mechanical mutations and fixed witnesses)")
	flag.StringVar(&cfg.snapshotFiles, "snapshot-files", "", "snapshot mode: explicit CSV of additional untracked source paths to capture with tracked working-tree edits; prints dangling commit, changes no branch/index")
}

func hasExactStage(stages string) bool {
	for _, name := range strings.Split(stages, ",") {
		if strings.TrimSpace(name) == "exactprobe" {
			return true
		}
	}
	return false
}

func exactConfigs(cfg config) ([]exactConfig, error) {
	if cfg.exactCases < 0 || cfg.exactCPU < 0 || (cfg.exactCases == 0) == (cfg.exactCPU == 0) {
		return nil, errors.New("exactprobe requires positive -exact-cases OR -exact-cpu-budget (with -exact-cases=0)")
	}
	if cfg.exactSeeds == "" {
		return nil, errors.New("exactprobe requires explicit -exact-seeds")
	}
	seeds, err := exactSeedList(cfg.exactSeeds)
	if err != nil {
		return nil, err
	}
	var out []exactConfig
	campaigns := map[string]bool{}
	for _, name := range strings.Split(cfg.exactCampaigns, ",") {
		if (name != "relations" && name != "uniform-finite") || campaigns[name] {
			return nil, fmt.Errorf("invalid or duplicate exact campaign %q", name)
		}
		campaigns[name] = true
		for _, seed := range seeds {
			out = append(out, exactConfig{Campaign: name, Seed: seed, Cases: cfg.exactCases, CPUNS: int64(cfg.exactCPU)})
		}
	}
	return out, nil
}

func gitText(repo string, args ...string) (string, error) {
	out, err := exec.Command("git", append([]string{"-C", repo}, args...)...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %v: %w: %s", args, err, out)
	}
	return strings.TrimSpace(string(out)), nil
}

func hashSources(root string, paths ...string) (string, error) {
	h := sha256.New()
	for _, path := range paths {
		err := filepath.WalkDir(filepath.Join(root, path), func(p string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if !strings.HasSuffix(p, ".go") {
				return nil
			}
			rel, err := filepath.Rel(root, p)
			if err != nil {
				return err
			}
			b, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			fmt.Fprintf(h, "%s\x00%d\x00", rel, len(b))
			h.Write(b)
			return nil
		})
		if err != nil {
			return "", err
		}
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func exactSource(cfg config) (*exactEvidence, error) {
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		return nil, errors.New("exactprobe CPU accounting requires Linux or macOS")
	}
	head, err := gitText(cfg.worktree, "rev-parse", "HEAD")
	if err != nil {
		return nil, err
	}
	expected, err := gitText(cfg.repo, "rev-parse", cfg.commit+"^{commit}")
	if err != nil {
		return nil, err
	}
	if head != expected {
		return nil, fmt.Errorf("exactprobe worktree HEAD %s differs from requested snapshot %s", head, expected)
	}
	cmd := exec.Command("git", "-C", cfg.worktree, "symbolic-ref", "-q", "HEAD")
	if err := cmd.Run(); err == nil {
		return nil, errors.New("exactprobe requires a detached worktree")
	} else {
		var exit *exec.ExitError
		if !errors.As(err, &exit) || exit.ExitCode() != 1 {
			return nil, err
		}
	}
	tree, err := gitText(cfg.worktree, "rev-parse", "HEAD^{tree}")
	if err != nil {
		return nil, err
	}
	configs, err := exactConfigs(cfg)
	if err != nil {
		return nil, err
	}
	e := &exactEvidence{SourceCommit: head, SourceTree: tree, MutationID: "baseline", Comparison: "independent decimalref numeric value/class/sign/canonicality and raw IEEE flags; not cohort equality", GoVersion: runtime.Version(), Platform: runtime.GOOS + "/" + runtime.GOARCH, Configs: configs, Selection: "relations: cycle families x widths x modes; uniform-finite: cycle operations x widths x modes; ChaCha8 keyed by full little-endian uint64 seed in a 32-byte key; entropy per case", BudgetScope: "RUSAGE_SELF user+system delta, generation+model+public kernel+comparison; includes GC; excludes build/startup; checked between cases; overshoot measured; no exact equal-CPU claim"}
	e.GeneratorSHA, err = hashSources(cfg.worktree, "bid754-go/internal/decimalprobe")
	if err != nil {
		return nil, err
	}
	e.ModelSHA, err = hashSources(cfg.worktree, "bid754-go/internal/decimalref")
	if err != nil {
		return nil, err
	}
	e.HarnessSHA, err = hashSources(cfg.worktree, "bid754-go/finite_arithmetic_campaign_test.go", "bid754-go/finite_arithmetic_fuzz_test.go")
	if err != nil {
		return nil, err
	}
	executable, err := os.Executable()
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(executable)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(b)
	e.ToolSHA = hex.EncodeToString(sum[:])
	return e, nil
}

func parseExactReport(out string, cfg exactConfig) (*exactReport, error) {
	var report *exactReport
	for _, line := range strings.Split(out, "\n") {
		if !strings.HasPrefix(line, "EXACTPROBE_REPORT ") {
			continue
		}
		if report != nil {
			return nil, errors.New("duplicate exact report")
		}
		report = new(exactReport)
		dec := json.NewDecoder(strings.NewReader(strings.TrimPrefix(line, "EXACTPROBE_REPORT ")))
		dec.DisallowUnknownFields()
		if err := dec.Decode(report); err != nil {
			return nil, err
		}
		if err := dec.Decode(new(any)); err != io.EOF {
			return nil, errors.New("trailing report data")
		}
	}
	if report == nil {
		return nil, errors.New("missing exact campaign completion evidence")
	}
	r := report
	if r.Version != 1 || r.GeneratorVersion != 1 || r.ModelVersion != 1 || r.Config != cfg || r.Executed <= 0 || r.Mismatches < 0 || r.Mismatches > r.Executed || r.CPUNS < 0 || r.WallNS <= 0 {
		return nil, errors.New("invalid exact report accounting/configuration")
	}
	if b, err := hex.DecodeString(r.Digest); err != nil || len(b) != 32 {
		return nil, errors.New("invalid input digest")
	}
	if cfg.Cases > 0 && (r.Executed != cfg.Cases || r.Stop != "cases") {
		return nil, errors.New("incomplete declared case campaign")
	}
	if cfg.CPUNS > 0 && (r.CPUNS < cfg.CPUNS || r.Stop != "process_cpu") {
		return nil, errors.New("incomplete process CPU campaign")
	}
	if len(r.Findings) != min(r.Mismatches, 8) {
		return nil, errors.New("missing mismatch evidence")
	}
	laneTotal := 0
	for _, count := range r.Lanes {
		if count <= 0 {
			return nil, errors.New("invalid lane count")
		}
		laneTotal += count
	}
	if laneTotal != r.Executed {
		return nil, errors.New("lane counts do not match executed cases")
	}
	last := -1
	for _, f := range r.Findings {
		c := f.Sample.Case
		if f.Index <= last || f.Index >= r.Executed || f.Sample.Version != 1 || (c.Width != 32 && c.Width != 64 && c.Width != 128) || len(c.Operands) < 2 || f.ExpectedRaw == "" || f.ActualRaw == "" || f.Diagnostic == "" || (!f.ValueMismatch && !f.FlagsMismatch) || f.FlagsMismatch != (f.ExpectedFlags != f.ActualFlags) {
			return nil, errors.New("invalid arithmetic finding evidence")
		}
		if f.ValueMismatch && f.ExpectedRaw == f.ActualRaw {
			return nil, errors.New("identical raw values cannot prove value mismatch")
		}
		last = f.Index
	}
	return report, nil
}

func exactVerdict(out string, runErr error, deadline bool, cfg exactConfig) (string, *exactReport, string) {
	if deadline {
		return "timeout", nil, "wall safety deadline"
	}
	failure := classifyStageFailure(out)
	if runErr != nil && (failure == "panic" || failure == "timeout") {
		return failure, nil, tail(out, 1500)
	}
	report, err := parseExactReport(out, cfg)
	if err != nil {
		return "inconclusive", nil, err.Error() + "\n" + tail(out, 1500)
	}
	if runErr == nil && report.Mismatches == 0 {
		return "pass", report, ""
	}
	if runErr != nil && report.Mismatches > 0 && strings.Contains(out, "--- FAIL: TestFiniteArithmeticCampaign ") && strings.Contains(out, fmt.Sprintf("EXACTPROBE_ARITHMETIC_MISMATCH count=%d", report.Mismatches)) {
		return "killed", report, "EXACTPROBE_ARITHMETIC_MISMATCH"
	}
	return "inconclusive", report, "exit status and precise arithmetic evidence disagree"
}

func (e *engine) runExactStage(bin string) (string, string, time.Duration) {
	start := time.Now()
	e.exactRuns = nil
	if e.exact == nil {
		return "inconclusive", "missing exact source metadata", time.Since(start)
	}
	overall := "pass"
	var diagnostics []string
	for _, cfg := range e.exact.Configs {
		data, _ := json.Marshal(cfg)
		ctx, cancel := context.WithTimeout(context.Background(), e.cfg.stageTimeout)
		cmd := exec.CommandContext(ctx, bin, "-test.run", "^TestFiniteArithmeticCampaign$", "-test.count=1", "-test.timeout", e.cfg.stageTimeout.String())
		cmd.Dir = e.goDir
		for _, v := range os.Environ() {
			if !strings.HasPrefix(v, "BID754_EXACTPROBE=") && !strings.HasPrefix(v, "GOFLAGS=") && !strings.HasPrefix(v, "GOMAXPROCS=") {
				cmd.Env = append(cmd.Env, v)
			}
		}
		cmd.Env = append(cmd.Env, "GOFLAGS=", "GOMAXPROCS=1", "BID754_EXACTPROBE="+string(data))
		begin := time.Now()
		out, err := cmd.CombinedOutput()
		verdict, report, diagnostic := exactVerdict(string(out), err, ctx.Err() == context.DeadlineExceeded, cfg)
		if verdict == "killed" && len(e.exact.Baseline) > 0 && !exactBaselineCoversFinding(cfg, report, e.exact.Baseline) {
			verdict, diagnostic = "inconclusive", "arithmetic discrepancy lacks a passing pristine input prefix; CPU budgets can execute different case counts"
		}
		cancel()
		run := exactRun{Config: cfg, Status: verdict, Report: report, WallNS: time.Since(begin).Nanoseconds(), Diagnostic: diagnostic}
		if cmd.ProcessState != nil {
			run.ProcessCPUNS = (cmd.ProcessState.UserTime() + cmd.ProcessState.SystemTime()).Nanoseconds()
		}
		e.exactRuns = append(e.exactRuns, run)
		if verdict != "pass" {
			diagnostics = append(diagnostics, string(out))
			overall = aggregateExactVerdict(overall, verdict)
		}
	}
	return overall, strings.Join(diagnostics, "\n"), time.Since(start)
}

func aggregateExactVerdict(a, b string) string {
	rank := map[string]int{"pass": 0, "inconclusive": 1, "timeout": 2, "panic": 3, "killed": 4}
	if _, ok := rank[a]; !ok {
		a = "inconclusive"
	}
	if _, ok := rank[b]; !ok {
		b = "inconclusive"
	}
	if rank[b] > rank[a] {
		return b
	}
	return a
}

func snapshotSource(cfg config) error {
	parent, err := gitText(cfg.repo, "rev-parse", cfg.commit+"^{commit}")
	if err != nil {
		return err
	}
	sourceIndex, err := exec.Command("git", "-C", cfg.repo, "ls-files", "--stage", "-z").Output()
	if err != nil {
		return fmt.Errorf("snapshot source index: %w", err)
	}
	dir, err := os.MkdirTemp("", "mutgate-index-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	index := filepath.Join(dir, "index")
	git := func(input string, args ...string) (string, error) {
		cmd := exec.Command("git", append([]string{"-C", cfg.repo}, args...)...)
		for _, v := range os.Environ() {
			if !strings.HasPrefix(v, "GIT_INDEX_FILE=") {
				cmd.Env = append(cmd.Env, v)
			}
		}
		cmd.Env = append(cmd.Env, "GIT_INDEX_FILE="+index)
		cmd.Stdin = strings.NewReader(input)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("snapshot git %v: %w: %s", args, err, out)
		}
		return strings.TrimSpace(string(out)), nil
	}
	if _, err = git("", "read-tree", "--empty"); err != nil {
		return err
	}
	if _, err = git(string(sourceIndex), "update-index", "-z", "--index-info"); err != nil {
		return err
	}
	if _, err = git("", "write-tree"); err != nil {
		return err
	}
	if _, err = git("", "add", "-u", "--", "."); err != nil {
		return err
	}
	if cfg.snapshotFiles != "" {
		for _, p := range strings.Split(cfg.snapshotFiles, ",") {
			clean := filepath.Clean(p)
			if filepath.IsAbs(p) || clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
				return fmt.Errorf("snapshot path must be repo-relative source: %q", p)
			}
			if _, err = git("", "add", "--", clean); err != nil {
				return err
			}
		}
	}
	tree, err := git("", "write-tree")
	if err != nil {
		return err
	}
	commit, err := git("Temporary exactprobe source snapshot\n", "commit-tree", tree, "-p", parent)
	if err != nil {
		return err
	}
	fmt.Println(commit)
	return nil
}

func exactBuildFault(out string, runErr, contextErr error) string {
	if contextErr != nil {
		return "build_timeout"
	}
	var exit *exec.ExitError
	if !errors.As(runErr, &exit) {
		return "build_exec_error"
	}
	for _, line := range strings.Split(out, "\n") {
		parts := strings.SplitN(line, ":", 4)
		if len(parts) != 4 || !strings.HasSuffix(parts[0], ".go") {
			continue
		}
		_, lineErr := strconv.Atoi(parts[1])
		_, colErr := strconv.Atoi(parts[2])
		if lineErr == nil && colErr == nil {
			return "compile_error"
		}
	}
	return "build_infrastructure_error"
}

func exactSeedList(text string) ([]int64, error) {
	var seeds []int64
	seen := map[int64]bool{}
	for _, v := range strings.Split(text, ",") {
		seed, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
		if err != nil || seen[seed] {
			return nil, fmt.Errorf("invalid or duplicate exact seed %q", v)
		}
		seen[seed] = true
		seeds = append(seeds, seed)
	}
	return seeds, nil
}

func exactBaselineCoversFinding(cfg exactConfig, report *exactReport, baseline []exactRun) bool {
	if report == nil {
		return false
	}
	for _, run := range baseline {
		if run.Config != cfg || run.Status != "pass" || run.Report == nil {
			continue
		}
		if cfg.Cases > 0 && run.Report.Digest != report.Digest {
			return false
		}
		for _, finding := range report.Findings {
			if finding.Index < run.Report.Executed {
				return true
			}
		}
	}
	return false
}
