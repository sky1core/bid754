package bid754

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/sky1core/bid754/bid754-go/internal/decimalprobe"
	"github.com/sky1core/bid754/bid754-go/internal/decimalref"
)

type finiteObservation struct {
	Path     string `json:"path"`
	Bits     string `json:"bits"`
	Flags    uint32 `json:"flags"`
	HasFlags bool   `json:"has_flags"`
}

type finitePathsConfig struct {
	Seed     uint64 `json:"seed"`
	Language string `json:"language,omitempty"`
	Witness  string `json:"witness,omitempty"`
	Samples  int    `json:"samples"`
	Uniform  int    `json:"uniform"`
	Replay   string `json:"replay"`
	Failures string `json:"failures"`
}

type finitePathFinding struct {
	Version      int                      `json:"version"`
	Config       finitePathsConfig        `json:"config"`
	Sample       decimalprobe.Sample      `json:"sample"`
	Reduced      decimalprobe.Sample      `json:"reduced"`
	Expected     string                   `json:"expected"`
	Flags        uint32                   `json:"flags"`
	Observations []finiteObservation      `json:"observations"`
	Reason       string                   `json:"reason"`
	Shrink       decimalprobe.ShrinkStats `json:"shrink"`
	Source       map[string]string        `json:"source"`
	BigDecimal   *bigDecimalResult        `json:"bigdecimal,omitempty"`
}

func finitePathFailureKey(err error) string {
	parts := strings.SplitN(err.Error(), ":", 3)
	return strings.Join(parts[:min(2, len(parts))], ":")
}

func finitePathsDecode(data []byte, dst any) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return err
	}
	if err := dec.Decode(new(any)); err != io.EOF {
		return fmt.Errorf("trailing JSON data: %v", err)
	}
	return nil
}

func finitePathNames(c decimalref.Case, language string) map[string]bool {
	paths := map[string]bool{language + "/public/mode": true, language + "/port": true}
	if c.Mode == "nearest_even" {
		paths[language+"/public/flags"] = true
		if c.Op != "fma" {
			paths[language+"/public/value"] = false
		}
	}
	return paths
}

func finiteCheckPaths(c decimalref.Case, want decimalref.Result, observations []finiteObservation, languages []string) error {
	paths := make(map[string]bool)
	for _, language := range languages {
		for path, flags := range finitePathNames(c, language) {
			paths[path] = flags
		}
	}
	seen := make(map[string]bool)
	var first *finiteObservation
	for _, actual := range observations {
		hasFlags, ok := paths[actual.Path]
		if !ok || seen[actual.Path] || hasFlags != actual.HasFlags {
			return fmt.Errorf("invalid, duplicate or misclassified path %+v", actual)
		}
		seen[actual.Path] = true
		flags := actual.Flags
		if !hasFlags {
			if flags != 0 {
				return fmt.Errorf("value-only path %s supplied flags", actual.Path)
			}
			flags = want.Flags
		}
		if err := decimalref.CompareQuantum(c.Width, want, actual.Bits, flags); err != nil {
			return fmt.Errorf("%s: %w", actual.Path, err)
		}
		if first != nil && first.Bits != actual.Bits {
			return fmt.Errorf("%s: cross-path bits differ: %s vs %s from %s", actual.Path, actual.Bits, first.Bits, first.Path)
		}
		if first == nil {
			copy := actual
			first = &copy
		}
	}
	if len(seen) != len(paths) {
		return fmt.Errorf("missing execution paths: got %v expected %v", seen, paths)
	}
	return nil
}

type finiteProcessClient struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	reader *bufio.Reader
	output *io.PipeReader
	stderr bytes.Buffer
	done   chan error
}

func finiteStartProcess(ctx context.Context, path string, args ...string) (*finiteProcessClient, error) {
	return finiteStartProcessWithInputCancel(ctx, false, path, args...)
}

func finiteStartProcessWithInputCancel(ctx context.Context, closeInput bool, path string, args ...string) (*finiteProcessClient, error) {
	client := &finiteProcessClient{cmd: exec.CommandContext(ctx, path, args...), done: make(chan error, 1)}
	client.cmd.WaitDelay = 250 * time.Millisecond
	client.cmd.Stderr = &client.stderr
	var err error
	client.stdin, err = client.cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	if closeInput {
		client.cmd.Cancel = client.stdin.Close
		client.cmd.WaitDelay = 2 * time.Second
	}
	reader, writer := io.Pipe()
	client.reader, client.output = bufio.NewReader(reader), reader
	client.cmd.Stdout = writer
	if err := client.cmd.Start(); err != nil {
		client.stdin.Close()
		reader.Close()
		writer.Close()
		return nil, err
	}
	finished := make(chan struct{})
	go func() {
		err := client.cmd.Wait()
		writer.CloseWithError(err)
		client.done <- err
		close(finished)
	}()
	go func() {
		select {
		case <-ctx.Done():
			reader.CloseWithError(ctx.Err())
		case <-finished:
		}
	}()
	return client, nil
}

func (client *finiteProcessClient) exchangeLine(data []byte) ([]byte, error) {
	if _, err := client.stdin.Write(append(data, '\n')); err != nil {
		return nil, fmt.Errorf("process input: %w", err)
	}
	line, err := client.readLine()
	if err != nil {
		return nil, fmt.Errorf("process output: %w", err)
	}
	return line, nil
}

func (client *finiteProcessClient) readLine() ([]byte, error) {
	line, err := client.reader.ReadSlice('\n')
	return bytes.Clone(line), err
}

func (client *finiteProcessClient) rustExchange(c decimalref.Case) ([]finiteObservation, error) {
	data, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	line, err := client.exchangeLine(data)
	if err != nil {
		return nil, fmt.Errorf("Rust exchange: %w", err)
	}
	var observations []finiteObservation
	if err := finitePathsDecode(line, &observations); err != nil {
		return nil, err
	}
	return observations, nil
}

func (client *finiteProcessClient) close() error {
	closeErr := client.stdin.Close()
	if errors.Is(closeErr, os.ErrClosed) {
		closeErr = nil
	}
	rest, readErr := io.ReadAll(client.reader)
	waitErr := <-client.done
	client.output.Close()
	if closeErr != nil || readErr != nil || waitErr != nil || len(rest) != 0 {
		return fmt.Errorf("process shutdown: input=%v wait=%v read=%v extra=%q stderr=%s", closeErr, waitErr, readErr, rest, client.stderr.String())
	}
	return nil
}

func finiteTestContext(t *testing.T) context.Context {
	t.Helper()
	deadline, ok := t.Deadline()
	if !ok {
		t.Fatal("finite subprocess checks require a Go test timeout")
	}
	ctx, cancel := context.WithDeadline(context.Background(), deadline.Add(-time.Second))
	t.Cleanup(cancel)
	return ctx
}

func finiteRustPaths(t *testing.T, path string) func(decimalref.Case) ([]finiteObservation, error) {
	t.Helper()
	ctx := finiteTestContext(t)
	client, err := finiteStartProcess(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		err := client.close()
		if ctx.Err() == context.DeadlineExceeded {
			t.Log(`FINITE-PATH-EXECUTION {"version":1,"language":"rust","status":"timeout"}`)
		}
		if err != nil {
			t.Error(err)
		}
	})
	return client.rustExchange
}

func TestFiniteRustDeadline(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	client, err := finiteStartProcess(ctx, "sh", "-c", "exec sleep 300")
	if err != nil {
		t.Fatal(err)
	}
	started := time.Now()
	_, exchangeErr := client.exchangeLine([]byte("stall"))
	closeErr := client.close()
	if exchangeErr == nil || closeErr == nil || ctx.Err() != context.DeadlineExceeded {
		t.Fatalf("stalled subprocess not rejected: exchange=%v close=%v context=%v", exchangeErr, closeErr, ctx.Err())
	}
	if time.Since(started) > 3*time.Second || client.cmd.ProcessState == nil {
		t.Fatalf("subprocess was not reaped within deadline: elapsed=%s state=%v", time.Since(started), client.cmd.ProcessState)
	}
}

func finitePathSamples(cfg finitePathsConfig, visit func(decimalprobe.Sample) error) error {
	if cfg.Witness != "" {
		if cfg.Replay != "" {
			return fmt.Errorf("witness and replay are mutually exclusive")
		}
		width := map[string]int{"zero-quantum32": 32, "zero-quantum64": 64, "zero-quantum128": 128}[cfg.Witness]
		if width != 0 {
			s, err := decimalprobe.Boundary("quantize", width, "nearest_even", 0)
			if err != nil {
				return err
			}
			return visit(s)
		}
		s, expected, flags, _, _, err := finiteCampaignWitness(cfg.Witness)
		if err != nil {
			return err
		}
		want, err := decimalprobe.Validate(s)
		if err != nil {
			return err
		}
		if err := decimalref.CompareQuantum(s.Case.Width, want, expected, flags); err != nil {
			return fmt.Errorf("witness expected result: %w", err)
		}
		return visit(s)
	}
	if cfg.Replay != "" {
		data, err := os.ReadFile(cfg.Replay)
		if err != nil {
			return err
		}
		var finding finitePathFinding
		if err := finitePathsDecode(data, &finding); err != nil {
			return err
		}
		if finding.BigDecimal != nil && os.Getenv("BID754_BIGDECIMAL_CLASSES") == "" {
			return fmt.Errorf("BigDecimal finding replay requires BID754_BIGDECIMAL_CLASSES and BID754_BIGDECIMAL_JAVA")
		}
		if finding.Version != 1 {
			return fmt.Errorf("unsupported finding version %d", finding.Version)
		}
		want, err := decimalprobe.Validate(finding.Sample)
		if err != nil {
			return err
		}
		if err := decimalref.CompareQuantum(finding.Sample.Case.Width, want, finding.Expected, finding.Flags); err != nil {
			return fmt.Errorf("recorded expectation differs from current model: %w", err)
		}
		originalErr := visit(finding.Sample)
		reducedErr := visit(finding.Reduced)
		return errors.Join(originalErr, reducedErr)
	}
	rng := rand.New(rand.NewPCG(cfg.Seed, 754))
	for _, width := range []int{32, 64, 128} {
		for _, family := range decimalprobe.Families() {
			for i := 0; i < cfg.Samples; i++ {
				s, err := decimalprobe.Generate(family, width, "nearest_even", rng.Uint64(), rng.Uint64(), int32(rng.Uint32()), i%2 != 0)
				if err != nil {
					return err
				}
				for _, mode := range finiteModes {
					s.Case.Mode = mode
					if err := visit(s); err != nil {
						return err
					}
				}
			}
		}
		for _, op := range []string{"add", "sub", "mul", "div", "fma", "quantize"} {
			for i := 0; i < 20; i++ {
				s, err := decimalprobe.Boundary(op, width, "nearest_even", i)
				if err != nil {
					return err
				}
				for _, mode := range finiteModes {
					s.Case.Mode = mode
					if err := visit(s); err != nil {
						return err
					}
				}
			}
			for i := 0; i < cfg.Uniform; i++ {
				s, err := decimalprobe.Uniform(op, width, "nearest_even", rng.Uint64(), rng.Uint64(), int32(rng.Uint32()), i%2 != 0)
				if err != nil {
					return err
				}
				for _, mode := range finiteModes {
					s.Case.Mode = mode
					if err := visit(s); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func finitePathSource(rust string) (map[string]string, error) {
	source := map[string]string{"platform": runtime.GOOS + "/" + runtime.GOARCH, "go": runtime.Version()}
	for _, args := range [][]string{{"rev-parse", "HEAD"}, {"diff", "HEAD", "--no-ext-diff"}} {
		b, err := exec.Command("git", args...).Output()
		if err != nil {
			return nil, fmt.Errorf("source identity git %v: %w", args, err)
		}
		if args[0] == "rev-parse" {
			source["commit"] = strings.TrimSpace(string(b))
		} else {
			source["tracked_diff_sha256"] = fmt.Sprintf("%x", sha256.Sum256(b))
		}
	}
	binary, err := os.Executable()
	if err != nil {
		return nil, err
	}
	for name, path := range map[string]string{"go_binary_sha256": binary, "rust_binary_sha256": rust} {
		if path != "" {
			b, err := os.ReadFile(path)
			if err != nil {
				return nil, fmt.Errorf("executable identity %s: %w", name, err)
			}
			source[name] = fmt.Sprintf("%x", sha256.Sum256(b))
		}
	}
	return source, nil
}

func TestFiniteArithmeticPaths(t *testing.T) {
	cfg := finitePathsConfig{Seed: 754, Samples: 2, Uniform: 2, Failures: os.Getenv("BID754_FINITE_FAILURES")}
	if raw := os.Getenv("BID754_FINITE_PATHS"); raw != "" {
		if err := finitePathsDecode([]byte(raw), &cfg); err != nil {
			t.Fatal(err)
		}
	}
	if cfg.Samples < 1 || cfg.Uniform < 1 {
		t.Fatal("samples and uniform must be positive")
	}
	languages := []string{"go"}
	var rust func(decimalref.Case) ([]finiteObservation, error)
	rustPath := os.Getenv("BID754_FINITE_RUST")
	if cfg.Language != "" && cfg.Language != "go" && cfg.Language != "rust" {
		t.Fatal("language must be go or rust")
	}
	if cfg.Language == "go" {
		rustPath = ""
	}
	if cfg.Language == "rust" && rustPath == "" {
		t.Fatal("Rust-only execution requires BID754_FINITE_RUST")
	}
	if rustPath != "" {
		rust = finiteRustPaths(t, rustPath)
		languages = append(languages, "rust")
	}
	if cfg.Language == "rust" {
		languages = []string{"rust"}
	}
	evaluate := func(s decimalprobe.Sample) ([]finiteObservation, error) {
		var observations []finiteObservation
		if cfg.Language != "rust" {
			var err error
			observations, err = finitePublicPaths(s.Case)
			if err != nil {
				return nil, err
			}
		}
		if rust != nil {
			other, err := rust(s.Case)
			if err != nil {
				return nil, err
			}
			observations = append(observations, other...)
		}
		return observations, nil
	}
	java := startBigDecimal(t)
	javaCoverage := make(map[string]int)
	javaExcluded := make(map[string]int)
	javaChecked := 0
	coverage := make(map[string]int)
	flagCoverage := make(map[string]uint32)
	cases, checked := 0, 0
	err := finitePathSamples(cfg, func(s decimalprobe.Sample) error {
		want, err := decimalprobe.Validate(s)
		if err != nil {
			return fmt.Errorf("invalid oracle sample %+v: %w", s, err)
		}
		cases++
		observations, err := evaluate(s)
		if err != nil {
			return fmt.Errorf("execution %+v: %w", s, err)
		}
		var javaResult *bigDecimalResult
		var mismatch error
		if java != nil {
			result, err := java.evaluate(s.Case)
			if err != nil {
				return fmt.Errorf("BigDecimal execution %+v: %w", s, err)
			}
			javaResult = &result
			if result.Status == "excluded" {
				javaExcluded[result.Reason]++
			} else {
				javaChecked++
				for _, observation := range observations {
					key := fmt.Sprintf("width=%d op=%s mode=%s path=%s", s.Case.Width, s.Case.Op, s.Case.Mode, observation.Path)
					javaCoverage[key]++
				}
				mismatch = compareBigDecimal(s.Case.Width, result, observations)
			}
		}
		if mismatch == nil {
			mismatch = finiteCheckPaths(s.Case, want, observations, languages)
		}
		if mismatch != nil {
			failureKey := finitePathFailureKey(mismatch)
			reduced, stats, err := decimalprobe.Shrink(s, 128, func(candidate decimalprobe.Sample) (bool, error) {
				want, err := decimalprobe.Validate(candidate)
				if err != nil {
					return false, err
				}
				got, err := evaluate(candidate)
				if err != nil {
					return false, err
				}
				var failure error
				if java != nil {
					result, err := java.evaluate(candidate.Case)
					if err != nil {
						return false, err
					}
					if result.Status == "ok" {
						failure = compareBigDecimal(candidate.Case.Width, result, got)
					}
				}
				if failure == nil {
					failure = finiteCheckPaths(candidate.Case, want, got, languages)
				}
				return failure != nil && finitePathFailureKey(failure) == failureKey, nil
			})
			if err != nil {
				return fmt.Errorf("%v; shrinking failed: %w", mismatch, err)
			}
			expected, err := decimalref.Encode(s.Case.Width, want.Value)
			if err != nil {
				return err
			}
			source, err := finitePathSource(rustPath)
			if err != nil {
				return fmt.Errorf("%v; source identity: %w", mismatch, err)
			}
			if java != nil {
				for k, v := range java.source {
					source[k] = v
				}
			}
			finding := finitePathFinding{1, cfg, s, reduced, expected, want.Flags, observations, mismatch.Error(), stats, source, javaResult}
			data, err := json.Marshal(finding)
			if err != nil {
				return err
			}
			if cfg.Failures != "" && cfg.Replay == "" {
				file, err := os.OpenFile(cfg.Failures, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
				if err != nil {
					return fmt.Errorf("%v; saving finding: %w", mismatch, err)
				}
				_, writeErr := file.Write(append(data, '\n'))
				closeErr := file.Close()
				if writeErr != nil || closeErr != nil {
					return fmt.Errorf("saving finding: write=%v close=%v", writeErr, closeErr)
				}
			}
			return fmt.Errorf("FINITE-PATH-FINDING %s", data)
		}
		for _, observation := range observations {
			key := fmt.Sprintf("width=%d op=%s mode=%s family=%s path=%s", s.Case.Width, s.Case.Op, s.Case.Mode, s.Family, observation.Path)
			coverage[key]++
			if observation.HasFlags {
				flagKey := fmt.Sprintf("width=%d mode=%s path=%s", s.Case.Width, s.Case.Mode, observation.Path)
				flagCoverage[flagKey] |= observation.Flags
			}
			checked++
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Replay == "" && cfg.Witness == "" {
		if java != nil {
			for _, width := range []int{32, 64, 128} {
				for _, op := range []string{"add", "sub", "mul", "div", "fma", "quantize"} {
					for _, mode := range finiteModes {
						for _, language := range languages {
							for path := range finitePathNames(decimalref.Case{Op: op, Mode: mode}, language) {
								key := fmt.Sprintf("width=%d op=%s mode=%s path=%s", width, op, mode, path)
								if javaCoverage[key] == 0 {
									t.Fatalf("missing BigDecimal numeric comparisons: %s", key)
								}
							}
						}
					}
				}
			}
		}
		expected := make(map[string]int)
		for _, width := range []int{32, 64, 128} {
			families := make(map[string]string)
			for _, family := range decimalprobe.Families() {
				op, err := decimalprobe.FamilyOperation(family)
				if err != nil {
					t.Fatal(err)
				}
				families[family] = op
			}
			for _, op := range []string{"add", "sub", "mul", "div", "fma", "quantize"} {
				families["uniform-finite/"+op] = op
				families["boundary-finite/"+op] = op
			}
			for family, op := range families {
				n := cfg.Samples
				if strings.HasPrefix(family, "uniform-finite/") {
					family, n = "uniform-finite", cfg.Uniform
				}
				if strings.HasPrefix(family, "boundary-finite/") {
					family, n = "boundary-finite", 20
				}
				for _, mode := range finiteModes {
					for _, language := range languages {
						for path := range finitePathNames(decimalref.Case{Op: op, Mode: mode}, language) {
							expected[fmt.Sprintf("width=%d op=%s mode=%s family=%s path=%s", width, op, mode, family, path)] = n
						}
					}
				}
			}
		}
		for _, width := range []int{32, 64, 128} {
			for _, mode := range finiteModes {
				for _, language := range languages {
					for path, hasFlags := range finitePathNames(decimalref.Case{Op: "add", Mode: mode}, language) {
						if !hasFlags {
							continue
						}
						key := fmt.Sprintf("width=%d mode=%s path=%s", width, mode, path)
						if flagCoverage[key] != 0x3d {
							t.Fatalf("flag coverage %s got=%#x expected=0x3d", key, flagCoverage[key])
						}
					}
				}
			}
		}
		if len(coverage) != len(expected) {
			t.Fatalf("coverage cells=%d expected=%d", len(coverage), len(expected))
		}
		for cell, n := range expected {
			if coverage[cell] != n {
				t.Fatalf("coverage %s got=%d expected=%d", cell, coverage[cell], n)
			}
		}
	}
	keys := make([]string, 0, len(coverage))
	for key := range coverage {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		t.Logf("FINITE-PATH-COVERAGE %s cases=%d", key, coverage[key])
	}
	flagKeys := make([]string, 0, len(flagCoverage))
	for key := range flagCoverage {
		flagKeys = append(flagKeys, key)
	}
	sort.Strings(flagKeys)
	for _, key := range flagKeys {
		t.Logf("FINITE-FLAG-COVERAGE %s flags=%#x", key, flagCoverage[key])
	}
	if java != nil {
		javaObs, excluded := 0, 0
		for _, n := range javaCoverage {
			javaObs += n
		}
		for _, reason := range []string{"division-by-zero", "exponent-range", "quantize-precision"} {
			excluded += javaExcluded[reason]
			t.Logf("BIGDECIMAL-EXCLUDED reason=%s cases=%d", reason, javaExcluded[reason])
		}
		if javaChecked+excluded != cases {
			t.Fatalf("BigDecimal adjudication counts checked=%d excluded=%d generated=%d", javaChecked, excluded, cases)
		}
		javaKeys := make([]string, 0, len(javaCoverage))
		for key := range javaCoverage {
			javaKeys = append(javaKeys, key)
		}
		sort.Strings(javaKeys)
		for _, key := range javaKeys {
			t.Logf("BIGDECIMAL-COVERAGE %s comparisons=%d", key, javaCoverage[key])
		}
		t.Logf("BIGDECIMAL languages=%s generated=%d checked=%d excluded=%d observations=%d cells=%d", strings.Join(languages, ","), cases, javaChecked, excluded, javaObs, len(javaCoverage))
	}
	t.Logf("FINITE-PATHS languages=%s seed=%d samples=%d uniform=%d cases=%d observations=%d cells=%d flag_cells=%d replay=%t witness=%s", strings.Join(languages, ","), cfg.Seed, cfg.Samples, cfg.Uniform, cases, checked, len(coverage), len(flagCoverage), cfg.Replay != "", cfg.Witness)
}

func TestFinitePathCoverageContract(t *testing.T) {
	s, err := decimalprobe.Generate("add_tie", 32, "nearest_even", 0, 754, 0, false)
	if err != nil {
		t.Fatal(err)
	}
	want, err := decimalprobe.Validate(s)
	if err != nil {
		t.Fatal(err)
	}
	got, err := finitePublicPaths(s.Case)
	if err != nil {
		t.Fatal(err)
	}
	if err := finiteCheckPaths(s.Case, want, got, []string{"go"}); err != nil {
		t.Fatal(err)
	}
	if finiteCheckPaths(s.Case, want, got[:len(got)-1], []string{"go"}) == nil {
		t.Fatal("missing executor accepted")
	}
	if finiteCheckPaths(s.Case, want, append(got, got[0]), []string{"go"}) == nil {
		t.Fatal("duplicate executor accepted")
	}
	if finiteCheckPaths(s.Case, want, got, []string{"go", "rust"}) == nil {
		t.Fatal("missing Rust observations accepted")
	}
	for i, original := range got {
		if !original.HasFlags {
			continue
		}
		got[i].Flags ^= 0x20
		if finiteCheckPaths(s.Case, want, got, []string{"go"}) == nil {
			t.Fatalf("lost inexact accepted on %s", original.Path)
		}
		got[i] = original
	}
	groups := make(map[string]map[string]bool)
	cfg := finitePathsConfig{Seed: 827, Samples: 2, Uniform: 2}
	if err := finitePathSamples(cfg, func(s decimalprobe.Sample) error {
		key := fmt.Sprintf("%d/%s/%s/%v", s.Case.Width, s.Case.Op, s.Family, s.Case.Operands)
		if groups[key] == nil {
			groups[key] = make(map[string]bool)
		}
		groups[key][s.Case.Mode] = true
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	for key, modes := range groups {
		if len(modes) != 5 {
			t.Fatalf("same operands not tested in all five modes: %s %v", key, modes)
		}
	}
	var cfg2 finitePathsConfig
	if finitePathsDecode([]byte(`{"unknown":1}`), &cfg2) == nil || finitePathsDecode([]byte(`{} {}`), &cfg2) == nil {
		t.Fatal("ambiguous campaign configuration accepted")
	}
}

func TestFinitePathInputContract(t *testing.T) {
	for raw := ExceptionFlags(0); raw < 32; raw++ {
		want := uint32(raw&4)<<1 | uint32(raw&1)<<5 | uint32(raw&2)<<3 | uint32(raw&8)>>1 | uint32(raw&16)>>4
		got, err := finiteMapPublicFlags(raw)
		if err != nil || got != want {
			t.Fatalf("flags %#x: got %#x, want %#x, error=%v", raw, got, want, err)
		}
	}
	for _, raw := range []ExceptionFlags{-1, 32, 1 << 32, 1 << 40} {
		if _, err := finiteMapPublicFlags(raw); err == nil {
			t.Fatalf("unrepresentable flag word %#x was truncated", raw)
		}
	}
	base := decimalref.Case{Width: 32, Op: "add", Mode: "nearest_even", Operands: []string{"32800001", "32800001"}}
	var inputs [][]byte
	for _, change := range []func(*decimalref.Case){
		func(c *decimalref.Case) { c.Width = 256 },
		func(c *decimalref.Case) { c.Op = "missing" },
		func(c *decimalref.Case) { c.Mode = "missing" },
		func(c *decimalref.Case) { c.Operands = []string{"32800001"} },
		func(c *decimalref.Case) { c.Operands = []string{"1", "32800001"} },
		func(c *decimalref.Case) { c.Operands = []string{"032800001", "32800001"} },
		func(c *decimalref.Case) { c.Operands = []string{"3280000A", "32800001"} },
		func(c *decimalref.Case) {
			c.Width = 128
			c.Operands = []string{"3040000000000000_0000000000000001", "3040000000000000:0000000000000001"}
		},
	} {
		c := base
		change(&c)
		if _, err := finitePublicPaths(c); err == nil {
			t.Fatalf("invalid executor input accepted: %+v", c)
		}
		data, err := json.Marshal(c)
		if err != nil {
			t.Fatal(err)
		}
		inputs = append(inputs, data)
	}
	if binary := os.Getenv("BID754_FINITE_RUST"); binary != "" {
		inputs = append(inputs, []byte(`{"width":32,"op":"add","mode":"nearest_even","operands":["32800001","32800001"],"unknown":1}`), nil)
		for _, data := range inputs {
			cmd := exec.CommandContext(finiteTestContext(t), binary)
			cmd.WaitDelay = 250 * time.Millisecond
			cmd.Stdin = bytes.NewReader(append(data, '\n'))
			var out, stderr bytes.Buffer
			cmd.Stdout, cmd.Stderr = &out, &stderr
			err := cmd.Run()
			var exit *exec.ExitError
			if !errors.As(err, &exit) || exit.ExitCode() != 1 || out.Len() != 0 || !strings.Contains(stderr.String(), "finite_probe:") {
				t.Fatalf("invalid Rust input not rejected through error result: %s exit=%v out=%s stderr=%s", data, err, out.String(), stderr.String())
			}
		}
	}
}
