package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestFinitePathMutationRequiresNumericFinding(t *testing.T) {
	for _, reason := range []string{"go/public/mode: flags: got 0x0, want 0x20", "go/port: finite cohort differs: got 10E0, want 1E1"} {
		out := `finite_arithmetic_paths_test.go:1: FINITE-PATH-FINDING {"version":1,"reason":"` + reason + `","expected":"32800001","flags":32,"observations":[{"path":"go/public/mode","bits":"32800001","flags":0,"has_flags":true}]}`
		if len(finiteNumericFinding(out, "go")) == 0 {
			t.Fatalf("numeric failure not recognized: %s", out)
		}
		if len(finiteNumericFinding(out, "rust")) != 0 {
			t.Fatal("Go-only failure credited to Rust")
		}
	}
	for _, out := range []string{
		"--- FAIL: TestFiniteArithmeticPaths (0.00s)\nexit status 1",
		`FINITE-PATH-FINDING {"version":1,"reason":"missing execution paths"}`,
		`FINITE-PATH-FINDING {"version":1,"reason":"rust/port: flags: got 0, want 32","expected":"32800001","flags":32,"observations":[{"path":"rust/port","bits":"32800001","flags":32,"has_flags":true}]}`,
		`FINITE-PATH-FINDING {"version":1,"reason":"rust/port: execution failed"}`,
		`FINITE-PATH-FINDING {"version":2,"reason":"rust/port: flags: got 0, want 1"}`,
		`FINITE-PATH-FINDING {"version":1,"reason":"rust/port: flags: got 0, want 1"} {}`,
	} {
		if len(finiteNumericFinding(out, "rust")) != 0 {
			t.Fatalf("infrastructure or invalid output counted as numeric detection: %s", out)
		}
	}
}

func TestFiniteQuantumProbeSites(t *testing.T) {
	root, err := gitTopLevel(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, probe := range finiteQuantumProbes() {
		if _, _, err := exactProbeSite(root, probe); err != nil {
			t.Fatal(err)
		}
	}
}

func TestFinitePathMutationRequiresIntendedOperationAndFault(t *testing.T) {
	base := `{"version":1,"config":{"witness":"unfused128"},"sample":{"case":{"width":128,"op":"fma","mode":"nearest_even"}},"reason":"rust/public/mode: flags: got 0x20, want 0","expected":"3040000000000000:0000000000000001","flags":0,"observations":[{"path":"rust/public/mode","bits":"3082000000000000:0000000000000000","flags":32,"has_flags":true}]}`
	probe := exactProbe{Name: "unfused128"}
	if !finiteIntendedFinding(probe, json.RawMessage(base), "rust") {
		t.Fatal("intended FMA fault rejected")
	}
	for _, change := range []func(map[string]any){
		func(f map[string]any) { f["sample"].(map[string]any)["case"].(map[string]any)["op"] = "mul" },
		func(f map[string]any) { f["sample"].(map[string]any)["case"].(map[string]any)["width"] = 64 },
		func(f map[string]any) { f["sample"].(map[string]any)["case"].(map[string]any)["mode"] = "down" },
		func(f map[string]any) { f["config"].(map[string]any)["witness"] = "unfused64" },
		func(f map[string]any) { f["observations"].([]any)[0].(map[string]any)["flags"] = 0 },
		func(f map[string]any) { f["observations"].([]any)[0].(map[string]any)["bits"] = f["expected"] },
	} {
		var f map[string]any
		if err := json.Unmarshal([]byte(base), &f); err != nil {
			t.Fatal(err)
		}
		change(f)
		raw, err := json.Marshal(f)
		if err != nil {
			t.Fatal(err)
		}
		if finiteIntendedFinding(probe, raw, "rust") {
			t.Fatalf("unrelated fault credited: %s", raw)
		}
	}
	if finiteIntendedFinding(probe, json.RawMessage(base), "go") {
		t.Fatal("Rust finding credited to Go")
	}
}

func TestFinitePathRustTimeoutClassification(t *testing.T) {
	root, err := gitTopLevel(".")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	binary := filepath.Join(dir, "finite.test")
	e := &engine{repo: root, goDir: filepath.Join(root, "bid754-go"), cfg: config{stageTimeout: 10 * time.Second}}
	build := exec.Command("go", e.buildArgs("portable", binary)...)
	build.Dir, build.Env = e.goDir, e.buildEnv("portable")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build public test binary: %v\n%s", err, out)
	}
	innerTimeout := filepath.Join(dir, "inner-timeout")
	if err := os.WriteFile(innerTimeout, []byte("#!/bin/sh\nsleep 2\nexec '"+strings.ReplaceAll(binary, "'", "'\"'\"'")+"' \"$@\" -test.timeout=2s\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	stall := filepath.Join(dir, "stall")
	if err := os.WriteFile(stall, []byte("#!/bin/sh\nexec sleep 300\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	env := append(os.Environ(), "BID754_FINITE_RUST="+stall, "BID754_FINITE_FAILURES=")
	status, out, elapsed := e.runFinitePathBinary(innerTimeout, "rust", env, time.Now())
	if status != "timeout" || !finiteExecutionTimedOut(out, "rust") || !strings.Contains(out, "--- FAIL: TestFiniteArithmeticPaths") || len(e.pathFinding) != 0 {
		t.Fatalf("inner Rust timeout classified as %s: %s", status, out)
	}
	if elapsed > e.cfg.stageTimeout+finiteWaitDelay {
		t.Fatalf("timeout exceeded bound: %s", elapsed)
	}
	if finiteExecutionTimedOut(out, "go") {
		t.Fatal("Rust deadline attributed to Go")
	}
	for _, invalid := range []string{
		`FINITE-PATH-EXECUTION {"version":2,"language":"rust","status":"timeout"}`,
		`FINITE-PATH-EXECUTION {"version":1,"language":"rust","status":"pass"}`,
		`FINITE-PATH-EXECUTION {"version":1,"language":"rust","status":"timeout"} {}`,
	} {
		if finiteExecutionTimedOut(invalid, "rust") {
			t.Fatalf("invalid timeout evidence accepted: %s", invalid)
		}
	}
}
