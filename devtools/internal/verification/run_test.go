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
	"time"
)

var nativeTagSettings = []struct{ name, value string }{
	{"NATIVE_TAGS", "-tags bid754_native"},
	{"TIER1_LONG_NATIVE_TAGS", "-tags bid754_native,bid754_tier1_long"},
	{"DECNUMBER_DIFF_NATIVE_TAGS", "-tags bid754_native,bid754_decnumber_diff"},
	{"D32_EXHAUSTIVE_NATIVE_TAGS", "-tags bid754_native,bid754_d32_exhaustive"},
}

func verificationFixture(t *testing.T) (string, Plan) {
	t.Helper()
	t.Setenv("BID754_SNAPSHOT_ARCHIVE", "")
	t.Setenv("BID754_SNAPSHOT_ID", "")
	t.Setenv("GOFLAGS", "")
	t.Setenv("MAKEFILES", "")
	t.Setenv("GNUMAKEFLAGS", "")
	t.Setenv("MAKE", "")
	if err := os.Unsetenv("MAKE"); err != nil {
		t.Fatal(err)
	}
	for _, setting := range nativeTagSettings {
		t.Setenv(setting.name, "")
		if err := os.Unsetenv(setting.name); err != nil {
			t.Fatal(err)
		}
	}
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

func TestEvidenceAcceptsRelativeLogPath(t *testing.T) {
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	writeFixture(t, dir, "evidence.log", "=== RUN   TestWitness\n--- PASS: TestWitness (0.00s)\nPASS\nok\texample.com/witness\t0.01s\n")
	abs := filepath.Join(dir, "evidence.log")
	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(previousDir); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	rel, err := filepath.Rel(cwd, abs)
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{abs, rel} {
		if err := CheckEvidence(root, path, Evidence{Passes: []string{"TestWitness"}}); err != nil {
			t.Errorf("valid evidence at %q: %v", path, err)
		}
		if err := CheckEvidence(root, path, Evidence{Passes: []string{"TestMissing"}}); err == nil || !strings.Contains(err.Error(), `missing: "--- PASS: TestMissing"`) {
			t.Errorf("incomplete evidence at %q: %v", path, err)
		}
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

func TestRunnerRejectsExternalMakefileTestSelection(t *testing.T) {
	root, p := verificationFixture(t)
	writeFixture(t, root, "go.mod", "module example.com/verification-witness\n\ngo 1.24.0\n")
	writeFixture(t, root, "witness_test.go", "package witness\nimport \"testing\"\nfunc TestOrdinary(t *testing.T) {}\nfunc TestRequiredBehavior(t *testing.T) {}\n")
	writeFixture(t, root, "Makefile", "check-scripts:\n\t@go test -count=1 -v .\n")
	p.Gates[0].Comparison = "package-tests"
	p.Gates[0].Evidence.Patterns = []string{`(?m)^ok\s+example.com/verification-witness\s`}
	baseline, err := Run(root, p, "scripts", filepath.Join(root, "test_results/baseline"), "makefiles", io.Discard)
	if err != nil || baseline.Status != "passed" {
		t.Fatalf("baseline: %v %+v", err, baseline)
	}
	writeFixture(t, root, "witness_test.go", "package witness\nimport \"testing\"\nfunc TestOrdinary(t *testing.T) {}\nfunc TestRequiredBehavior(t *testing.T) { t.Fatal(\"required failure witness\") }\n")
	failed, err := Run(root, p, "scripts", filepath.Join(root, "test_results/failing"), "makefiles", io.Discard)
	if err == nil || failed.Status != "failed" || len(failed.Gates) != 1 {
		t.Fatalf("required Go test did not fail: %v %+v", err, failed)
	}
	log, err := os.ReadFile(filepath.Join(root, "test_results/failing/shell.log"))
	if err != nil || !strings.Contains(string(log), "required failure witness") {
		t.Fatalf("unrelated failure: %v %s", err, log)
	}
	foreign := t.TempDir()
	writeFixture(t, foreign, "selection.mk", "export GOFLAGS := -skip=TestRequiredBehavior\n")
	t.Setenv("MAKEFILES", filepath.Join(foreign, "selection.mk"))
	dir := filepath.Join(root, "test_results/foreign")
	rejected, err := Run(root, p, "scripts", dir, "makefiles", io.Discard)
	if err == nil || !strings.Contains(err.Error(), "MAKEFILES") || rejected.Status != "failed" || len(rejected.Gates) != 0 {
		t.Fatalf("external Makefile changed required test selection: %v %+v", err, rejected)
	}
	var saved Result
	if err := ReadJSON(filepath.Join(dir, "result.json"), &saved); err != nil {
		t.Fatal(err)
	}
	if saved.Error != rejected.Error || saved.Status != "failed" || saved.Configuration["MAKEFILES"] != os.Getenv("MAKEFILES") {
		t.Fatalf("external Makefile rejection not recorded: %+v", saved)
	}
}

func TestRunnerRejectsInheritedMakeOverrides(t *testing.T) {
	root, p := verificationFixture(t)
	for _, name := range []string{"MAKE", "MAKEFILES", "GNUMAKEFLAGS"} {
		t.Run(name, func(t *testing.T) {
			values := []string{"untracked.mk", "-n", "-i", "GOFLAGS=-skip=TestRequiredBehavior"}
			if name == "MAKE" {
				values = append(values, "", "true")
			}
			for i, value := range values {
				t.Setenv(name, value)
				dir := filepath.Join(root, "test_results/"+name+"-"+strconv.Itoa(i))
				rejected, err := Run(root, p, "scripts", dir, "make-overrides", io.Discard)
				if err == nil || !strings.Contains(err.Error(), name) || rejected.Status != "failed" || len(rejected.Gates) != 0 {
					t.Fatalf("inherited %s=%q changed the plan: %v %+v", name, value, err, rejected)
				}
				var saved Result
				if err := ReadJSON(filepath.Join(dir, "result.json"), &saved); err != nil {
					t.Fatal(err)
				}
				if saved.Error != rejected.Error || saved.Status != "failed" || saved.Configuration[name] != value {
					t.Fatalf("Make override rejection not recorded: %+v", saved)
				}
			}
		})
	}
}

func TestRunnerRejectsRecursiveMakeReplacement(t *testing.T) {
	root, p := verificationFixture(t)
	writeFixture(t, root, "go.mod", "module example.com/verification-witness\n\ngo 1.24.0\n")
	writeFixture(t, root, "witness_test.go", "package witness\nimport \"testing\"\nfunc TestOrdinary(t *testing.T) {}\n")
	makefile := "check-scripts:\n\t@go test -count=1 .\n\t@$(MAKE) --no-print-directory required-child\n\nrequired-child:\n\t@echo recursive child witness\n\t@true\n"
	writeFixture(t, root, "Makefile", makefile)
	p.Gates[0].Comparison = "package-tests"
	p.Gates[0].Evidence.Patterns = []string{`(?m)^ok\s+example.com/verification-witness\s`}
	for _, mode := range []string{"unset", "canonical"} {
		if mode == "canonical" {
			t.Setenv("MAKE", "make")
		}
		dir := filepath.Join(root, "test_results/"+mode)
		baseline, err := Run(root, p, "scripts", dir, "recursive-make", io.Discard)
		if err != nil || baseline.Status != "passed" || baseline.Configuration["MAKE"] != "make" {
			t.Fatalf("%s baseline: %v %+v", mode, err, baseline)
		}
		log, err := os.ReadFile(filepath.Join(dir, "shell.log"))
		if err != nil || !strings.Contains(string(log), "recursive child witness") {
			t.Fatalf("%s required child did not run: %v %s", mode, err, log)
		}
	}
	writeFixture(t, root, "Makefile", strings.ReplaceAll(makefile, "@true", "@false"))
	failed, err := Run(root, p, "scripts", filepath.Join(root, "test_results/failing"), "recursive-make", io.Discard)
	if err == nil || failed.Status != "failed" || len(failed.Gates) != 1 {
		t.Fatalf("required child did not fail: %v %+v", err, failed)
	}
	log, err := os.ReadFile(filepath.Join(root, "test_results/failing/shell.log"))
	if err != nil || !strings.Contains(string(log), "recursive child witness") {
		t.Fatalf("unrelated failure: %v %s", err, log)
	}
	t.Setenv("MAKE", "true")
	rejected, err := Run(root, p, "scripts", filepath.Join(root, "test_results/replaced"), "recursive-make", io.Discard)
	if err == nil || !strings.Contains(err.Error(), "MAKE") || rejected.Status != "failed" || len(rejected.Gates) != 0 {
		t.Fatalf("recursive Make was replaced: %v %+v", err, rejected)
	}
}

func TestAdjudicatorRejectsAlteredMakeConfiguration(t *testing.T) {
	root, p := verificationFixture(t)
	dir := filepath.Join(root, "test_results/baseline")
	baseline, err := Run(root, p, "scripts", dir, "make-overrides", io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(baseline)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"MAKE", "MAKEFILES", "GNUMAKEFLAGS"} {
		values := []string{"missing", "untracked.mk", "-n", "-i", "GOFLAGS=-skip=TestRequiredBehavior"}
		if name == "MAKE" {
			values = append(values, "", "true")
		}
		for _, value := range values {
			var candidate Result
			if err := json.Unmarshal(encoded, &candidate); err != nil {
				t.Fatal(err)
			}
			if value == "missing" {
				delete(candidate.Configuration, name)
			} else {
				candidate.Configuration[name] = value
			}
			if err := Validate(root, p, dir, candidate, baseline.SnapshotID, "make-overrides"); err == nil || !strings.Contains(err.Error(), name) {
				t.Errorf("altered %s=%q accepted: %v", name, value, err)
			}
		}
	}
}

func TestRunnerRejectsInheritedNativeTagArguments(t *testing.T) {
	root, p := verificationFixture(t)
	writeFixture(t, root, "Makefile", "NATIVE_TAGS ?= -tags bid754_native\nTIER1_LONG_NATIVE_TAGS ?= -tags bid754_native,bid754_tier1_long\nDECNUMBER_DIFF_NATIVE_TAGS ?= -tags bid754_native,bid754_decnumber_diff\nD32_EXHAUSTIVE_NATIVE_TAGS ?= -tags bid754_native,bid754_d32_exhaustive\ncheck-scripts:\n\t@printf '%s\\n' 'NATIVE_TAGS=$(NATIVE_TAGS)' 'TIER1_LONG_NATIVE_TAGS=$(TIER1_LONG_NATIVE_TAGS)' 'DECNUMBER_DIFF_NATIVE_TAGS=$(DECNUMBER_DIFF_NATIVE_TAGS)' 'D32_EXHAUSTIVE_NATIVE_TAGS=$(D32_EXHAUSTIVE_NATIVE_TAGS)'\n\t@python3 -B devtools/scripts/check_scripts.py\n")
	for _, mode := range []string{"unset", "canonical"} {
		if mode == "canonical" {
			for _, setting := range nativeTagSettings {
				t.Setenv(setting.name, setting.value)
			}
		}
		dir := filepath.Join(root, "test_results/"+mode)
		baseline, err := Run(root, p, "scripts", dir, "native-tags", io.Discard)
		if err != nil || baseline.Status != "passed" {
			t.Fatalf("%s baseline: %v", mode, err)
		}
		log, err := os.ReadFile(filepath.Join(dir, "shell.log"))
		if err != nil {
			t.Fatal(err)
		}
		for _, setting := range nativeTagSettings {
			if baseline.Configuration[setting.name] != setting.value || !strings.Contains(string(log), setting.name+"="+setting.value+"\n") {
				t.Errorf("%s %s configuration/make arguments: %+v %s", mode, setting.name, baseline.Configuration, log)
			}
		}
	}
	for _, setting := range nativeTagSettings {
		t.Run(setting.name, func(t *testing.T) {
			for i, extra := range []string{" -skip=^TestConcurrent", " -run=^$", " -list=.", " -short", "", " -tags=other"} {
				value := setting.value + extra
				if extra == "" {
					value = ""
				}
				t.Setenv(setting.name, value)
				dir := filepath.Join(root, "test_results/"+setting.name+"-"+strconv.Itoa(i))
				failed, err := Run(root, p, "scripts", dir, "native-tags", io.Discard)
				if err == nil || !strings.Contains(err.Error(), setting.name) || len(failed.Gates) != 0 || failed.Status != "failed" || failed.Configuration[setting.name] != value {
					t.Errorf("noncanonical %s=%q accepted: %v %+v", setting.name, value, err, failed)
				}
				var saved Result
				if err := ReadJSON(filepath.Join(dir, "result.json"), &saved); err != nil {
					t.Fatal(err)
				}
				if saved.Status != "failed" || saved.Configuration[setting.name] != value || saved.Error != failed.Error {
					t.Errorf("rejected configuration not recorded: %+v", saved)
				}
			}
		})
	}
}

func TestAdjudicatorRejectsAlteredNativeTagConfiguration(t *testing.T) {
	root, p := verificationFixture(t)
	dir := filepath.Join(root, "test_results/baseline")
	baseline, err := Run(root, p, "scripts", dir, "native-tags", io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(baseline)
	if err != nil {
		t.Fatal(err)
	}
	for _, setting := range nativeTagSettings {
		t.Run(setting.name, func(t *testing.T) {
			for _, value := range []string{"missing", "", setting.value + " -skip=^TestConcurrent", setting.value + " -run=^$", setting.value + " -list=.", setting.value + " -short"} {
				var candidate Result
				if err := json.Unmarshal(encoded, &candidate); err != nil {
					t.Fatal(err)
				}
				if value == "missing" {
					delete(candidate.Configuration, setting.name)
				} else {
					candidate.Configuration[setting.name] = value
				}
				if err := Validate(root, p, dir, candidate, baseline.SnapshotID, "native-tags"); err == nil || !strings.Contains(err.Error(), setting.name) {
					t.Errorf("altered %s=%q accepted: %v", setting.name, value, err)
				}
			}
		})
	}
}

func nativeVerificationFixture(t *testing.T) (string, Plan) {
	t.Helper()
	archive := "devtools/third_party/intel_dfp/IntelRDFPMathLib20U4.tar.gz"
	for _, rel := range []string{archive, "devtools/third_party/intel_dfp/lib/libbid.a"} {
		if _, err := os.Stat(filepath.Join("../../..", rel)); os.IsNotExist(err) {
			t.Skipf("native prerequisite integration dependency absent: %s", rel)
		} else if err != nil {
			t.Fatal(err)
		}
	}
	root, p := verificationFixture(t)
	t.Setenv("INTEL_DFP_OPT_CFLAGS", "-O3 -ffp-contract=off")
	writeFixture(t, root, ".gitignore", "test_results/\ndevtools/third_party/\n.env.sh\n")
	for _, rel := range []string{"devtools/scripts/setup_generation_inputs.sh", archive, "devtools/third_party/intel_dfp/lib/libbid.a"} {
		raw, err := os.ReadFile(filepath.Join("../../..", rel))
		if err != nil {
			t.Fatal(err)
		}
		writeFixture(t, root, rel, string(raw))
	}
	intel := filepath.Join(root, "devtools/third_party/intel_dfp")
	cmd := exec.Command("tar", "-xzf", filepath.Join(root, archive), "-C", intel)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("extract pinned archive: %v %s", err, out)
	}
	if err := os.Symlink("LIBRARY/src", filepath.Join(intel, "src")); err != nil {
		t.Fatal(err)
	}
	machine, err := exec.Command("uname", "-m").Output()
	if err != nil {
		t.Fatal(err)
	}
	aux := ""
	if arch := strings.TrimSpace(string(machine)); arch == "arm64" || arch == "aarch64" {
		aux = "-DBID_SIZE_LONG=8"
	}
	writeFixture(t, root, "devtools/third_party/intel_dfp/lib/.libbid.build-flags", "CALL_BY_REF=0\nGLOBAL_RND=0\nGLOBAL_FLAGS=0\nUNCHANGED_BINARY_FLAGS=0\nCFLAGS_AUX="+aux+"\nCFLAGS_OPT=-O3 -ffp-contract=off\n")
	writeFixture(t, root, ".env.sh", "#!/bin/bash\nexport CGO_ENABLED=1\n")
	p.Gates[0].Prerequisites = []string{"native"}
	p.Gates[0].Evidence.Patterns = []string{`(?m)^SCRIPT-SYNTAX-CHECK checked=3$`}
	return root, p
}

func TestRunnerChecksNativePrerequisiteWithPinnedInputs(t *testing.T) {
	root, p := nativeVerificationFixture(t)
	if err := os.Rename(filepath.Join(root, ".env.sh"), filepath.Join(root, "test_results/native-env.sh")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("test_results/native-env.sh", filepath.Join(root, ".env.sh")); err != nil {
		t.Fatal(err)
	}
	realBash, err := exec.LookPath("bash")
	if err != nil {
		t.Fatal(err)
	}
	bin := filepath.Join(root, "test_results/bin")
	writeFixture(t, bin, "bash", "#!/bin/sh\nif [ \"$1\" = devtools/scripts/setup_generation_inputs.sh ] && [ \"$2\" = verify-intel ]; then\n  printf '%s\\n' verify-intel >> \"$BID754_TEST_CHECKER_CALLS\"\nfi\nexec \"$BID754_TEST_REAL_BASH\" \"$@\"\n")
	if err := os.Chmod(filepath.Join(bin, "bash"), 0700); err != nil {
		t.Fatal(err)
	}
	calls := filepath.Join(root, "test_results/checker-calls.txt")
	t.Setenv("BID754_TEST_REAL_BASH", realBash)
	t.Setenv("BID754_TEST_CHECKER_CALLS", calls)
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	second := p.Gates[0]
	second.ID = "shell-second"
	p.Gates = append(p.Gates, second)
	started := time.Now()
	baseline, err := Run(root, p, "scripts", filepath.Join(root, "test_results/baseline"), "native-inputs", io.Discard)
	t.Logf("native prerequisite baseline Run elapsed=%s", time.Since(started))
	if err != nil || baseline.Status != "passed" || len(baseline.Gates) != 2 {
		t.Fatalf("pinned-input baseline: %v %+v", err, baseline)
	}
	raw, err := os.ReadFile(calls)
	if err != nil || string(raw) != "verify-intel\n" {
		t.Fatalf("repeated native prerequisite checker calls: %q %v", raw, err)
	}
	for _, tc := range []struct {
		name, rel, want string
		mutate          func(*testing.T, string)
	}{
		{"missing-stamp", "devtools/third_party/intel_dfp/lib/.libbid.build-flags", "native library or build stamp missing", nil},
		{"mismatched-stamp", "devtools/third_party/intel_dfp/lib/.libbid.build-flags", "native build stamp does not match pinned flags", func(t *testing.T, path string) { writeFixture(t, root, path, "CALL_BY_REF=1\n") }},
		{"external-source", "devtools/third_party/intel_dfp/src", "src alias does not resolve to pinned LIBRARY/src", func(t *testing.T, path string) {
			foreign := t.TempDir()
			writeFixture(t, foreign, "bid_conf.h", "foreign source\n")
			if err := os.Symlink(foreign, filepath.Join(root, path)); err != nil {
				t.Fatal(err)
			}
		}},
		{"changed-source", "devtools/third_party/intel_dfp/LIBRARY/src/bid_conf.h", "input differs from pinned archive: LIBRARY/src/bid_conf.h", func(t *testing.T, path string) { writeFixture(t, root, path, "foreign source\n") }},
		{"changed-archive", "devtools/third_party/intel_dfp/IntelRDFPMathLib20U4.tar.gz", "checksum mismatch", func(t *testing.T, path string) { writeFixture(t, root, path, "foreign archive\n") }},
		{"empty-library", "devtools/third_party/intel_dfp/lib/libbid.a", "native library or build stamp missing", func(t *testing.T, path string) { writeFixture(t, root, path, "") }},
		{"missing-env", ".env.sh", ".env.sh", nil},
		{"empty-env", ".env.sh", ".env.sh must be a nonempty regular file", func(t *testing.T, path string) { writeFixture(t, root, path, "") }},
		{"directory-env", ".env.sh", ".env.sh must be a nonempty regular file", func(t *testing.T, path string) {
			if err := os.Mkdir(filepath.Join(root, path), 0700); err != nil {
				t.Fatal(err)
			}
		}},
		{"dangling-env", ".env.sh", ".env.sh", func(t *testing.T, path string) {
			if err := os.Symlink("missing-env.sh", filepath.Join(root, path)); err != nil {
				t.Fatal(err)
			}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(root, tc.rel)
			backup := filepath.Join(t.TempDir(), "original")
			if err := os.Rename(path, backup); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if _, err := os.Lstat(path); err == nil {
					if err := os.Rename(path, filepath.Join(t.TempDir(), "mutant")); err != nil {
						t.Fatal(err)
					}
				}
				if err := os.Rename(backup, path); err != nil {
					t.Fatal(err)
				}
			})
			if tc.mutate != nil {
				tc.mutate(t, tc.rel)
			}
			dir := filepath.Join(root, "test_results/"+tc.name)
			failed, err := Run(root, p, "scripts", dir, "native-inputs", io.Discard)
			if err == nil || !strings.Contains(err.Error(), tc.want) || failed.Status != "failed" || len(failed.Gates) != 0 {
				t.Errorf("native defect not rejected for %q: %v %+v", tc.want, err, failed)
			}
			t.Logf("rejected %s before gate execution: %s", tc.name, failed.Error)
			var saved Result
			if err := ReadJSON(filepath.Join(dir, "result.json"), &saved); err != nil {
				t.Fatal(err)
			}
			if saved.Error != failed.Error || saved.Status != "failed" || len(saved.Gates) != 0 {
				t.Errorf("native prerequisite failure not recorded: %+v", saved)
			}
		})
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
