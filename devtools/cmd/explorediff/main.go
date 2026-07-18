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
	"encoding/json"
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
	repo     string
	seedText string
	cases    int
	bias     float64
	ops      string
	widths   string
	modes    string
	out      string
}

func main() {
	var cfg config
	flag.StringVar(&cfg.repo, "repo", "", "repository root holding bid754-go and the pinned Intel build (default: git toplevel of cwd)")
	flag.StringVar(&cfg.seedText, "seed", "", "case-stream seed, decimal or 0x hex uint64 (default: fresh from the current time)")
	flag.IntVar(&cfg.cases, "cases", 20000, "cases per (width, op) target; each case runs under every selected mode")
	flag.Float64Var(&cfg.bias, "bias", 0.25, "probability in [0,1] that a case is boundary-biased (pool draw + exponent correlation)")
	flag.StringVar(&cfg.ops, "ops", "add,sub,mul,div,fma,sqrt,quantize", "CSV of Tier 1 arithmetic ops")
	flag.StringVar(&cfg.widths, "widths", "32,64,128", "CSV of decimal widths")
	flag.StringVar(&cfg.modes, "modes", "nearest_even,toward_negative,toward_positive,toward_zero,nearest_away", "CSV of rounding-mode names")
	flag.StringVar(&cfg.out, "out", "", "JSONL findings/summary path (default: <repo>/test_results/explore_fresh_seed_<utc>_seed<seed>.jsonl)")
	flag.Parse()

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
	return fmt.Sprintf("(cd %s/devtools && bin=$(mktemp -d)/explorediff && go build -o \"$bin\" ./cmd/explorediff && \"$bin\" -repo %s -seed %d -cases %d -bias %g -ops %s -widths %s -modes %s)",
		repo, repo, seed, cfg.cases, cfg.bias, cfg.ops, cfg.widths, cfg.modes)
}

func gitTopLevel() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", fmt.Errorf("resolve repo root (pass -repo or run inside the repository): %v", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// describeCommit reports the repo state recorded into the config record.
func describeCommit(repo string) string {
	head, err := exec.Command("git", "-C", repo, "rev-parse", "--short", "HEAD").Output()
	if err != nil {
		return "unknown"
	}
	desc := strings.TrimSpace(string(head))
	status, err := exec.Command("git", "-C", repo, "status", "--porcelain", "--untracked-files=no").Output()
	if err == nil && strings.TrimSpace(string(status)) != "" {
		desc += "-dirty"
	}
	return desc
}

func run(cfg config) (int, error) {
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
	outPath := cfg.out
	if outPath == "" {
		outPath = filepath.Join(repo, "test_results",
			fmt.Sprintf("explore_fresh_seed_%s_seed%d.jsonl", time.Now().UTC().Format("20060102T150405Z"), seed))
	}
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return 1, err
	}

	binDir, err := os.MkdirTemp("", "explorediff-bin-")
	if err != nil {
		return 1, err
	}
	defer os.RemoveAll(binDir)
	binPath := filepath.Join(binDir, "explorenative")
	build := exec.Command("go", "build", "-o", binPath, "-tags", "bid754_native", "./internal/cmd/explorenative")
	build.Dir = goDir
	build.Env = append(append([]string{}, os.Environ()...), "GOFLAGS=", "CGO_ENABLED=1")
	if out, err := build.CombinedOutput(); err != nil {
		return 1, fmt.Errorf("build explorenative: %v\n%s", err, out)
	}

	fmt.Printf("explorediff: seed=%d (%s) cases=%d bias=%g ops=%s widths=%s modes=%s\n",
		seed, seedSource, cfg.cases, cfg.bias, cfg.ops, cfg.widths, cfg.modes)
	fmt.Printf("explorediff: repo=%s commit=%s\n", repo, describeCommit(repo))
	fmt.Printf("explorediff: findings JSONL -> %s\n", outPath)

	outFile, err := os.Create(outPath)
	if err != nil {
		return 1, err
	}
	defer outFile.Close()
	fileBuf := bufio.NewWriterSize(outFile, 1<<16)

	runner := exec.Command(binPath,
		"-seed", strconv.FormatUint(seed, 10),
		"-cases", strconv.Itoa(cfg.cases),
		"-bias", strconv.FormatFloat(cfg.bias, 'g', -1, 64),
		"-ops", cfg.ops, "-widths", cfg.widths, "-modes", cfg.modes,
		"-commit", describeCommit(repo))
	runner.Dir = goDir
	runner.Stderr = os.Stderr
	stdout, err := runner.StdoutPipe()
	if err != nil {
		return 1, err
	}
	if err := runner.Start(); err != nil {
		return 1, err
	}

	mismatches, sawSummary, err := relayRecords(stdout, fileBuf, os.Stdout)
	if err != nil {
		// The relay stopped consuming stdout; kill the producer so Wait
		// cannot block on a full pipe.
		_ = runner.Process.Kill()
	}
	waitErr := runner.Wait()
	flushErr := fileBuf.Flush()
	if err != nil {
		return 1, err
	}
	if waitErr != nil {
		return 1, fmt.Errorf("explorenative run failed: %v", waitErr)
	}
	if flushErr != nil {
		return 1, flushErr
	}
	if !sawSummary {
		return 1, fmt.Errorf("explorenative exited without a summary record; %s is truncated", outPath)
	}

	fmt.Printf("explorediff: reproduce with:\n  %s\n", reproCommand(cfg, repo, seed))
	if mismatches > 0 {
		fmt.Printf("explorediff: RESULT counterexamples found: %d mismatch record(s) in %s\n", mismatches, outPath)
		fmt.Println("explorediff: triage them through the existing manual procedures (regression vectors / sentinels / corpus promotion)")
		return 3, nil
	}
	fmt.Println("explorediff: RESULT no mismatch in this stream")
	return 0, nil
}

// relayRecords streams every subprocess JSONL line into the findings file
// while rendering console progress: per-target summaries, the first
// mismatches (all of them are always in the file), and the run total.
func relayRecords(in io.Reader, file io.Writer, console io.Writer) (int, bool, error) {
	const consoleMismatchLimit = 20
	type record struct {
		Type        string `json:"type"`
		Target      string `json:"target"`
		Mode        string `json:"mode"`
		CaseIndex   int    `json:"case_index"`
		X           string `json:"x"`
		Y           string `json:"y"`
		Z           string `json:"z"`
		CBits       string `json:"c_bits"`
		CFlags      string `json:"c_flags"`
		GoBits      string `json:"go_bits"`
		GoFlags     string `json:"go_flags"`
		Targets     int    `json:"targets"`
		Cases       int    `json:"cases"`
		Comparisons int    `json:"comparisons"`
		Mismatches  int    `json:"mismatches"`
		ElapsedMS   int64  `json:"elapsed_ms"`
	}
	scanner := bufio.NewScanner(in)
	scanner.Buffer(make([]byte, 0, 1<<20), 1<<20)
	mismatches := 0
	sawSummary := false
	for scanner.Scan() {
		line := scanner.Bytes()
		if _, err := file.Write(line); err != nil {
			return mismatches, sawSummary, err
		}
		if _, err := file.Write([]byte{'\n'}); err != nil {
			return mismatches, sawSummary, err
		}
		var rec record
		if err := json.Unmarshal(line, &rec); err != nil {
			return mismatches, sawSummary, fmt.Errorf("bad subprocess record %q: %v", line, err)
		}
		switch rec.Type {
		case "mismatch":
			mismatches++
			if mismatches <= consoleMismatchLimit {
				operands := rec.X
				if rec.Y != "" {
					operands += " y=" + rec.Y
				}
				if rec.Z != "" {
					operands += " z=" + rec.Z
				}
				fmt.Fprintf(console, "MISMATCH %s %s case=%d x=%s C=%s/%s go=%s/%s\n",
					rec.Target, rec.Mode, rec.CaseIndex, operands,
					rec.CBits, rec.CFlags, rec.GoBits, rec.GoFlags)
			} else if mismatches == consoleMismatchLimit+1 {
				fmt.Fprintln(console, "... further mismatches recorded in the JSONL only")
			}
		case "target_summary":
			fmt.Fprintf(console, "target %-14s cases=%d comparisons=%d mismatches=%d elapsed=%dms\n",
				rec.Target, rec.Cases, rec.Comparisons, rec.Mismatches, rec.ElapsedMS)
		case "summary":
			sawSummary = true
			fmt.Fprintf(console, "TOTAL targets=%d cases=%d comparisons=%d mismatches=%d elapsed=%dms\n",
				rec.Targets, rec.Cases, rec.Comparisons, rec.Mismatches, rec.ElapsedMS)
		}
	}
	return mismatches, sawSummary, scanner.Err()
}
