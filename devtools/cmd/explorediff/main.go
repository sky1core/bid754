// Command explorediff is a fresh-seed exploration fuzzer for the Tier 1
// arithmetic surface: it draws a new random case stream on every run (or a
// pinned one with -seed) and differentially compares the Go mechanical port
// against pinned Intel BID C with exact (result bits, raw flags) equality
// over add/sub/mul/div/fma/sqrt/quantize x Decimal32/64/128 x the five
// rounding modes.
//
// Status: discovery/audit tool, the same standing as devtools/cmd/mutgate —
// NOT a verification domain and never part of make verify-all. The pinned
// generated gates prove the pinned corpora; this tool exists to look where
// they have not looked yet. Every mismatch is recorded as one JSONL line
// (operation, width, mode, operand bits, both results and flag words) plus
// config/summary records, and the exact reproduction command is printed.
// Findings enter the tree only through the existing manual procedures
// (regression vectors, routing sentinels, corpus promotion).
//
// devtools is a stdlib-only module and must not require bid754-go
// (docs/SPEC.md inter-component dependency rules), so the case generation
// and both differential legs live in the bid754-go module as
// bid754-go/internal/cmd/explorenative; this driver builds that command with
// the native tags and runs it as a subprocess — a filesystem relationship,
// not a module dependency, exactly like the sentinel codegen's
// sentineloracle. The Intel BID build under devtools/third_party/intel_dfp
// must exist (make setup-native); in a worktree that directory already
// exists (it carries tracked files), so symlink its lib, src, include, and
// LIBRARY subdirectories from the primary checkout instead.
//
// Exit codes: 0 = run completed with no mismatch, 3 = run completed and
// recorded mismatches, 1 = the run itself failed. `go run` folds every
// nonzero child exit into 1, so invoke a built binary directly (as the
// printed reproduction command and make explore-fresh-seed do) whenever the
// 0/3/1 distinction matters.
package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type config struct {
	repo           string
	campaign       string
	replay         string
	shrinkAttempts int
	seedText       string
	cases          int
	bias           float64
	ops            string
	opsSet         bool
	widths         string
	modes          string
	out            string
}

func main() {
	var cfg config
	flag.StringVar(&cfg.repo, "repo", "", "repository root holding bid754-go and the pinned Intel build (default: git toplevel of cwd)")
	flag.StringVar(&cfg.campaign, "campaign", "legacy", "campaign: legacy (default, unchanged C-vs-Go fresh-seed fuzz), relations (decimalprobe relational generators + decimalref model oracle), or uniform-finite (uniform finite operands + model oracle)")
	flag.StringVar(&cfg.replay, "replay", "", "exact fixed-input JSONL replay: re-run recorded sample/finding cases through C, the Go port, and the model (no seed regeneration)")
	flag.IntVar(&cfg.shrinkAttempts, "shrink-attempts", 128, "max reference-campaign shrink attempts per finding; 0 disables shrinking (recorded disabled)")
	flag.StringVar(&cfg.seedText, "seed", "", "case-stream seed, decimal or 0x hex uint64 (default: fresh from the current time)")
	flag.IntVar(&cfg.cases, "cases", 20000, "explicit budget: legacy=cases per (width, op) target x every mode; relations=samples per (family, width, mode); uniform-finite=samples per (op, width, mode)")
	flag.Float64Var(&cfg.bias, "bias", 0.25, "legacy only: probability in [0,1] that a case is boundary-biased (pool draw + exponent correlation)")
	flag.StringVar(&cfg.ops, "ops", "", "legacy/uniform-finite CSV of ops; reference campaigns support add/sub/mul/div/fma/quantize and reject sqrt")
	flag.StringVar(&cfg.widths, "widths", "32,64,128", "CSV of decimal widths")
	flag.StringVar(&cfg.modes, "modes", "nearest_even,toward_negative,toward_positive,toward_zero,nearest_away", "CSV of rounding-mode names")
	flag.StringVar(&cfg.out, "out", "", "JSONL findings/summary path (default: <repo>/test_results/explore_<campaign>_<utc>_seed<seed>.jsonl)")
	if err := validateFlagArgs(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	flag.Parse()
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "ops" {
			cfg.opsSet = true
		}
	})

	code, err := run(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "explorediff: %v\n", err)
		os.Exit(1)
	}
	os.Exit(code)
}

// resolveSeed turns the -seed flag into the effective run seed. An empty
// flag draws a fresh seed from the wall clock (scrambled so consecutive
// runs differ in every bit, not just the low ones).
func resolveSeed(text string, now int64) (uint64, string, error) {
	if text == "" {
		return scramble(uint64(now)), "time", nil
	}
	seed, err := strconv.ParseUint(text, 0, 64)
	if err != nil {
		return 0, "", fmt.Errorf("invalid -seed %q: %v", text, err)
	}
	return seed, "flag", nil
}

// scramble is one SplitMix64 output step (same finalizer the case streams
// use), applied to the raw clock value.
func scramble(x uint64) uint64 {
	x += 0x9e3779b97f4a7c15
	x = (x ^ (x >> 30)) * 0xbf58476d1ce4e5b9
	x = (x ^ (x >> 27)) * 0x94d049bb133111eb
	return x ^ (x >> 31)
}

// reproCommand renders the exact command that replays this run. It builds
// and runs the binary directly instead of `go run`, which would fold the
// exit-3 counterexample signal into exit 1.
func reproCommand(cfg config, repo string, seed uint64) string {
	base := fmt.Sprintf("(cd %s && bin=$(mktemp -d)/explorediff && go build -o \"$bin\" ./cmd/explorediff && \"$bin\" -repo %s",
		shellQuote(filepath.Join(repo, "devtools")), shellQuote(repo))
	if cfg.replay != "" {
		return base + fmt.Sprintf(" -replay %s)", shellQuote(cfg.replay))
	}
	return base + fmt.Sprintf(" -campaign %s -seed %d -cases %d -bias %g -ops %s -widths %s -modes %s -shrink-attempts %d)",
		shellQuote(cfg.campaign), seed, cfg.cases, cfg.bias, shellQuote(cfg.ops), shellQuote(cfg.widths), shellQuote(cfg.modes), cfg.shrinkAttempts)
}

func gitTopLevel() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", fmt.Errorf("resolve repo root (pass -repo or run inside the repository): %v", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func run(cfg config) (code int, runErr error) {
	console := consoleWriter(os.Stdout)
	defer func() {
		runErr = errors.Join(runErr, console.err)
		if runErr != nil {
			code = 1
		}
	}()
	if err := resolveConfig(&cfg); err != nil {
		return 1, err
	}
	repo := cfg.repo
	var err error
	if repo == "" {
		if repo, err = gitTopLevel(); err != nil {
			return 1, err
		}
	}
	if repo, err = filepath.Abs(repo); err != nil {
		return 1, err
	}
	goDir := filepath.Join(repo, "bid754-go")
	if _, err := os.Stat(goDir); err != nil {
		return 1, fmt.Errorf("%s does not look like a bid754 checkout: %v", repo, err)
	}
	intelLib := filepath.Join(repo, "devtools", "third_party", "intel_dfp", "lib", "libbid.a")
	if _, err := os.Stat(intelLib); err != nil {
		return 1, fmt.Errorf("pinned Intel build missing (%s): run make setup-native; in a worktree, symlink the lib, src, include, and LIBRARY subdirectories of devtools/third_party/intel_dfp from the primary checkout (the directory itself already exists there)", intelLib)
	}

	seed, seedSource, err := resolveSeed(cfg.seedText, time.Now().UnixNano())
	if err != nil {
		return 1, err
	}
	label := cfg.campaign
	if cfg.replay != "" {
		label = "replay"
	}
	outPath := cfg.out
	if outPath == "" {
		outPath = filepath.Join(repo, "test_results",
			fmt.Sprintf("explore_%s_%s_seed%d.jsonl", label, time.Now().UTC().Format("20060102T150405Z"), seed))
	}
	if _, err := os.Lstat(outPath); err == nil {
		return 1, fmt.Errorf("output already exists: %s", outPath)
	} else if !os.IsNotExist(err) {
		return 1, err
	}

	binDir, err := os.MkdirTemp("", "explorediff-bin-")
	if err != nil {
		return 1, err
	}
	defer os.RemoveAll(binDir)
	binPath := filepath.Join(binDir, "explorenative")
	frozenRepo, sourceID, snapshot, err := freezeSource(repo, binDir)
	if err != nil {
		return 1, err
	}
	goDir = filepath.Join(frozenRepo, "bid754-go")
	intelLib = filepath.Join(frozenRepo, "devtools/third_party/intel_dfp/lib/libbid.a")
	librarySHA, err := hashFile(intelLib)
	if err != nil {
		return 1, err
	}
	archivePath := filepath.Join(frozenRepo, "devtools/third_party/intel_dfp/IntelRDFPMathLib20U4.tar.gz")
	archiveSHA, err := hashFile(archivePath)
	if err != nil {
		return 1, err
	}
	checkSource := func() error {
		if err := verifyFrozenSource(frozenRepo, snapshot, sourceID); err != nil {
			return err
		}
		for path, want := range map[string]string{intelLib: librarySHA, archivePath: archiveSHA} {
			got, err := hashFile(path)
			if err != nil {
				return err
			}
			if got != want {
				return fmt.Errorf("C provenance changed: %s", path)
			}
		}
		return nil
	}
	build := exec.Command("go", "build", "-a", "-o", binPath, "-tags", "bid754_native", "./internal/cmd/explorenative")
	build.Dir = goDir
	build.Env = append(buildEnvironment(), "GOCACHE="+filepath.Join(binDir, "go-cache"))
	if out, err := build.CombinedOutput(); err != nil {
		return 1, fmt.Errorf("build explorenative: %v\n%s", err, out)
	}

	if err := checkSource(); err != nil {
		return 1, err
	}
	binarySHA, err := hashFile(binPath)
	if err != nil {
		return 1, err
	}
	if cfg.replay != "" {
		fmt.Fprintf(console, "explorediff: campaign=replay replay=%s\n", cfg.replay)
	} else {
		fmt.Fprintf(console, "explorediff: campaign=%s seed=%d (%s) cases=%d bias=%g ops=%s widths=%s modes=%s shrink-attempts=%d\n",
			cfg.campaign, seed, seedSource, cfg.cases, cfg.bias, cfg.ops, cfg.widths, cfg.modes, cfg.shrinkAttempts)
	}
	fmt.Fprintf(console, "explorediff: repo=%s commit=%s\n", repo, sourceID)
	fmt.Fprintf(console, "explorediff: findings JSONL -> %s\n", outPath)

	args := []string{
		"-campaign", cfg.campaign, "-shrink-attempts", strconv.Itoa(cfg.shrinkAttempts),
		"-seed", strconv.FormatUint(seed, 10),
		"-cases", strconv.Itoa(cfg.cases),
		"-bias", strconv.FormatFloat(cfg.bias, 'g', -1, 64),
		"-ops", cfg.ops, "-widths", cfg.widths, "-modes", cfg.modes,
		"-commit", sourceID,
	}
	if cfg.replay != "" {
		replayAbs, err := filepath.Abs(cfg.replay)
		if err != nil {
			return 1, err
		}
		if _, err := os.Stat(replayAbs); err != nil {
			return 1, fmt.Errorf("replay file: %v", err)
		}
		args = append(args, "-replay", replayAbs)
	}
	validate := exec.Command(binPath, append(append([]string{}, args...), "-validate-only")...)
	validate.Dir = goDir
	if out, err := validate.CombinedOutput(); err != nil {
		return 1, fmt.Errorf("input preflight: %w: %s", err, out)
	}
	spool, err := os.CreateTemp(binDir, "records-*.jsonl")
	if err != nil {
		return 1, err
	}
	fileBuf := bufio.NewWriterSize(spool, 1<<16)
	published := false
	publish := func() error {
		if published {
			return nil
		}
		published = true
		return publishRecords(fileBuf, spool, outPath)
	}
	defer func() {
		runErr = errors.Join(runErr, publish())
		if runErr != nil {
			code = 1
		}
	}()
	args = append(args, "-source-snapshot", snapshot, "-source-id", sourceID, "-c-library-sha256", librarySHA, "-binary-sha256", binarySHA, "-c-archive-sha256", archiveSHA)
	runner := exec.Command(binPath, args...)
	runner.Dir = goDir
	runner.Stderr = os.Stderr
	stdout, err := runner.StdoutPipe()
	if err != nil {
		return 1, err
	}
	if err := runner.Start(); err != nil {
		return 1, err
	}

	counterexamples, sawSummary, err := relayRecords(stdout, fileBuf, console)
	if err != nil {
		// The relay stopped consuming stdout; kill the producer so Wait
		// cannot block on a full pipe.
		_ = runner.Process.Kill()
	}
	waitErr := runner.Wait()
	identityErr := checkSource()
	gotBinary, hashErr := hashFile(binPath)
	if hashErr == nil && gotBinary != binarySHA {
		hashErr = fmt.Errorf("built binary changed during run")
	}
	if waitErr != nil {
		waitErr = fmt.Errorf("explorenative run failed: %w", waitErr)
	}
	err = errors.Join(err, waitErr, identityErr, hashErr, publish())
	if err != nil {
		return 1, err
	}
	if !sawSummary {
		return 1, fmt.Errorf("explorenative exited without a summary record; %s is truncated", outPath)
	}

	fmt.Fprintf(console, "explorediff: reproduce with:\n  %s\n", reproCommand(cfg, repo, seed))
	if counterexamples > 0 {
		fmt.Fprintf(console, "explorediff: RESULT counterexamples found: %d record(s) in %s\n", counterexamples, outPath)
		fmt.Fprintln(console, "explorediff: auxiliary exploration only; a reference/candidate disagreement is a finding to triage against the spec, not an automatic product bug")
		return 3, nil
	}
	fmt.Fprintln(console, "explorediff: RESULT no counterexample in this stream")
	return 0, nil
}

func publishRecords(fileBuf *bufio.Writer, spool *os.File, outPath string) error {
	flushErr := fileBuf.Flush()
	if _, err := spool.Seek(0, 0); err != nil {
		return errors.Join(flushErr, err, spool.Close())
	}
	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		return errors.Join(flushErr, err, spool.Close())
	}
	outFile, err := os.OpenFile(outPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return errors.Join(flushErr, err, spool.Close())
	}
	_, copyErr := io.Copy(outFile, spool)
	return errors.Join(flushErr, copyErr, outFile.Close(), spool.Close())
}
