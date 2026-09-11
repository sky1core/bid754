package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func (e *engine) regenerateFiniteRust() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), e.cfg.buildTimeout)
	defer cancel()
	cmd := finiteCommandContext(ctx, "go", "run", "./tools/go2rs")
	cmd.Dir = filepath.Join(e.worktree, "devtools")
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func finiteNumericFinding(out, language string) json.RawMessage {
	const marker = "FINITE-PATH-FINDING "
	for _, line := range strings.Split(out, "\n") {
		_, raw, found := strings.Cut(line, marker)
		if !found {
			continue
		}
		var finding struct {
			Version      int    `json:"version"`
			Reason       string `json:"reason"`
			Expected     string `json:"expected"`
			Flags        uint32 `json:"flags"`
			Observations []struct {
				Path     string `json:"path"`
				Bits     string `json:"bits"`
				Flags    uint32 `json:"flags"`
				HasFlags bool   `json:"has_flags"`
			} `json:"observations"`
		}
		if json.Unmarshal([]byte(raw), &finding) != nil || finding.Version != 1 || !strings.HasPrefix(finding.Reason, language+"/") {
			continue
		}
		discrepancy := false
		for _, observation := range finding.Observations {
			if strings.HasPrefix(observation.Path, language+"/") && finding.Expected != "" && observation.Bits != "" &&
				(observation.Bits != finding.Expected || observation.HasFlags && observation.Flags != finding.Flags) {
				discrepancy = true
			}
		}
		if !discrepancy {
			continue
		}
		for _, failure := range []string{": BigDecimal value differs:", ": cross-path bits differ:", ": finite value differs:", ": finite cohort differs:", ": flags: got", ": class: got", ": sign: got", ": noncanonical actual", ": actual result is signaling NaN"} {
			if strings.Contains(finding.Reason, failure) {
				return json.RawMessage(raw)
			}
		}
	}
	return nil
}

func finiteIntendedFinding(probe exactProbe, raw json.RawMessage, language string) bool {
	var finding struct {
		Version int `json:"version"`
		Config  struct {
			Witness string `json:"witness"`
		} `json:"config"`
		Sample struct {
			Case struct {
				Width int    `json:"width"`
				Op    string `json:"op"`
				Mode  string `json:"mode"`
			} `json:"case"`
		} `json:"sample"`
		Reason       string `json:"reason"`
		Expected     string `json:"expected"`
		Flags        uint32 `json:"flags"`
		Observations []struct {
			Path     string `json:"path"`
			Bits     string `json:"bits"`
			Flags    uint32 `json:"flags"`
			HasFlags bool   `json:"has_flags"`
		} `json:"observations"`
	}
	if json.Unmarshal(raw, &finding) != nil || finding.Version != 1 || finding.Config.Witness != probe.Name {
		return false
	}
	width, op, fault := 32, "", ""
	switch probe.Name {
	case "quantize-midpoint":
		op, fault = "quantize", "midpoint"
	case "mul-midpoint":
		op, fault = "mul", "midpoint"
	case "mul-inexact":
		op, fault = "mul", "inexact"
	case "add-inexact":
		op, fault = "add", "inexact"
	case "unfused32":
		op, fault = "fma", "unfused"
	case "unfused64":
		width, op, fault = 64, "fma", "unfused"
	case "unfused128":
		width, op, fault = 128, "fma", "unfused"
	case "quantize-midpoint64":
		width, op, fault = 64, "quantize", "midpoint"
	case "quantize-midpoint128":
		width, op, fault = 128, "quantize", "midpoint"
	case "quantize-inexact64":
		width, op, fault = 64, "quantize", "inexact"
	case "quantize-inexact128":
		width, op, fault = 128, "quantize", "inexact"
	case "zero-quantum32":
		op, fault = "quantize", "quantum"
	case "zero-quantum64":
		width, op, fault = 64, "quantize", "quantum"
	case "zero-quantum128":
		width, op, fault = 128, "quantize", "quantum"
	default:
		return false
	}
	if finding.Sample.Case.Width != width || finding.Sample.Case.Op != op || finding.Sample.Case.Mode != "nearest_even" {
		return false
	}
	for _, actual := range finding.Observations {
		if actual.Path != language+"/public/mode" || !actual.HasFlags || actual.Bits == "" || finding.Expected == "" {
			continue
		}
		switch fault {
		case "midpoint":
			return actual.Bits != finding.Expected && actual.Flags == finding.Flags && (strings.Contains(finding.Reason, ": finite value differs:") || strings.Contains(finding.Reason, ": BigDecimal value differs:"))
		case "inexact":
			return actual.Bits == finding.Expected && finding.Flags == 0x20 && actual.Flags == 0
		case "unfused":
			return actual.Bits != finding.Expected && finding.Flags == 0 && actual.Flags == 0x20
		case "quantum":
			return actual.Bits != finding.Expected && finding.Flags == 0 && actual.Flags == 0 && strings.Contains(finding.Reason, ": finite cohort differs:")
		}
	}
	return false
}

func finiteExecutionTimedOut(out, language string) bool {
	for _, line := range strings.Split(out, "\n") {
		_, raw, found := strings.Cut(line, "FINITE-PATH-EXECUTION ")
		if !found {
			continue
		}
		var execution struct {
			Version  int    `json:"version"`
			Language string `json:"language"`
			Status   string `json:"status"`
		}
		if json.Unmarshal([]byte(raw), &execution) == nil && execution.Version == 1 && execution.Language == language && execution.Status == "timeout" {
			return true
		}
	}
	return false
}

func (e *engine) runFinitePathStage(st stage, binary string) (string, string, time.Duration) {
	started := time.Now()
	e.pathFinding = nil
	language := "go"
	env := append(append([]string{}, os.Environ()...), "BID754_FINITE_FAILURES=", "BID754_FINITE_RUST=")
	if st.Name == "rustfinite" {
		language = "rust"
		if out, err := e.regenerateFiniteRust(); err != nil {
			return "inconclusive", "Rust regeneration: " + err.Error() + "\n" + tail(out, 2000), time.Since(started)
		}
		target := filepath.Join(e.worktree, "bid754-rs", "target")
		ctx, cancel := context.WithTimeout(context.Background(), e.cfg.buildTimeout)
		cmd := finiteCommandContext(ctx, "cargo", "build", "--locked", "--features", "verification", "--example", "finite_probe", "--target-dir", target)
		cmd.Dir = filepath.Join(e.worktree, "bid754-rs")
		out, err := cmd.CombinedOutput()
		cancel()
		if err != nil {
			return "inconclusive", "Rust build: " + err.Error() + "\n" + tail(string(out), 2000), time.Since(started)
		}
		env = append(env, "BID754_FINITE_RUST="+filepath.Join(target, "debug", "examples", "finite_probe"))
	}
	return e.runFinitePathBinary(binary, language, env, started)
}

func (e *engine) runFinitePathBinary(binary, language string, env []string, started time.Time) (string, string, time.Duration) {
	config, _ := json.Marshal(map[string]any{"seed": e.cfg.seed, "samples": 2, "uniform": 2, "language": language, "witness": e.pathWitness})
	env = append(env, "BID754_FINITE_PATHS="+string(config))
	ctx, cancel := context.WithTimeout(context.Background(), e.cfg.stageTimeout)
	defer cancel()
	cmd := finiteCommandContext(ctx, binary, "-test.run", "^TestFiniteArithmeticPaths$", "-test.count=1", "-test.v", "-test.timeout", e.cfg.stageTimeout.String())
	cmd.Dir, cmd.Env = e.goDir, env
	out, err := cmd.CombinedOutput()
	full := string(out)
	if ctx.Err() != nil {
		return "timeout", full, time.Since(started)
	}
	if err != nil {
		if finiteExecutionTimedOut(full, language) || finiteExecutionTimedOut(full, "java") || classifyStageFailure(full) == "timeout" {
			return "timeout", full, time.Since(started)
		}
		e.pathFinding = finiteNumericFinding(full, language)
		if len(e.pathFinding) != 0 && strings.Contains(full, "--- FAIL: TestFiniteArithmeticPaths") {
			return "killed", full, time.Since(started)
		}
		if classifyStageFailure(full) == "panic" || strings.Contains(full, "panicked at") {
			return "panic", full, time.Since(started)
		}
		return "inconclusive", full, time.Since(started)
	}
	if !strings.Contains(full, "FINITE-PATHS languages="+language+" ") || requireSelectedPasses("TestFiniteArithmeticPaths", full) != nil {
		return "inconclusive", full, time.Since(started)
	}
	return "pass", full, time.Since(started)
}

func finiteQuantumProbes() []exactProbe {
	return []exactProbe{
		{Name: "zero-quantum32", File: "bid32_quantize.go", Function: "Bid32Quantize", Original: "res = very_fast_get_BID32(sign_x, exponent_y, 0)", Replacement: "res = very_fast_get_BID32(sign_x, exponent_x, 0)"},
		{Name: "zero-quantum64", File: "quantize64.go", Function: "Bid64Quantize", Original: "res = very_fast_get_BID64_small_mantissa(sign_x, exponent_y, 0)", Replacement: "res = very_fast_get_BID64_small_mantissa(sign_x, exponent_x, 0)"},
		{Name: "zero-quantum128", File: "bid128_quantize.go", Function: "Bid128Quantize", Original: "res = very_fast_get_BID128(sign_x, exponent_y, CX)", Replacement: "res = very_fast_get_BID128(sign_x, exponent_x, CX)"},
	}
}

func runFinitePathCheck(cfg config) error {
	if cfg.stages != "finitepaths" && cfg.stages != "rustfinite" && cfg.stages != "finitepaths,rustfinite" {
		return fmt.Errorf("pathcheck requires -stages finitepaths, rustfinite, or finitepaths,rustfinite")
	}
	var probes []exactProbe
	var err error
	if cfg.exactProbeSet == "quantum" {
		probes = finiteQuantumProbes()
	} else {
		probes, err = exactProbes(cfg.exactProbeSet)
	}
	if err != nil {
		return err
	}
	stages, err := resolveStages(cfg)
	if err != nil {
		return err
	}
	e, err := newEngine(cfg)
	if err != nil {
		return err
	}
	defer e.close()
	head, err := gitText(e.worktree, "rev-parse", "HEAD")
	if err != nil {
		return err
	}
	expected, err := gitText(e.repo, "rev-parse", cfg.commit+"^{commit}")
	if err != nil {
		return err
	}
	if head != expected {
		return fmt.Errorf("pathcheck worktree HEAD %s differs from requested snapshot %s", head, expected)
	}
	if err := exec.Command("git", "-C", e.worktree, "symbolic-ref", "-q", "HEAD").Run(); err == nil {
		return fmt.Errorf("pathcheck requires a detached worktree")
	} else {
		var exit *exec.ExitError
		if !errors.As(err, &exit) || exit.ExitCode() != 1 {
			return err
		}
	}
	if err := e.baseline(stages); err != nil {
		return err
	}
	failed := 0
	for _, p := range probes {
		e.pathWitness = p.Name
		site, pristine, err := exactProbeSite(e.worktree, p)
		if err != nil {
			return err
		}
		for _, st := range stages {
			e.pathFinding = nil
			result := e.evaluateMutant(site, pristine, []stage{st})
			language := "go"
			if st.Name == "rustfinite" {
				language = "rust"
			}
			intended := finiteIntendedFinding(p, e.pathFinding, language)
			if err := json.NewEncoder(e.jsonl).Encode(map[string]any{"intended_finding": intended, "type": "finite_path_probe", "probe": p.Name, "stage": st.Name, "result": result, "finding": e.pathFinding}); err != nil {
				return err
			}
			fmt.Printf("pathcheck %s %s status=%s numeric_finding=%t intended_finding=%t\n", p.Name, st.Name, result.Status, len(e.pathFinding) != 0, intended)
			if result.Status != "killed" || !intended {
				failed++
			}
		}
	}
	dirty, err := gitStatusPorcelain(e.worktree)
	if err != nil {
		return err
	}
	if dirty != "" {
		return fmt.Errorf("pathcheck left modified sources: %s", dirty)
	}
	if failed != 0 {
		return fmt.Errorf("%d finite path probes lack their intended numeric detection", failed)
	}
	return nil
}
