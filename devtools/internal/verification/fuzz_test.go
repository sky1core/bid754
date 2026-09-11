package verification

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestGoFuzzEvidenceRequiresExploration(t *testing.T) {
	const target = "FuzzFiniteArithmeticBigDecimal"
	const baseline = "fuzz: elapsed: 1s, gathering baseline coverage: 12/12 completed, now fuzzing with 2 workers\n"
	const progress = "fuzz: elapsed: 3s, execs: 100 (33/sec), new interesting: 1 (total: 13)\n"
	const pass = "--- PASS: " + target + " (3.00s)\n"
	const log = "=== RUN   " + target + "\n" + baseline + progress + pass + "PASS\n"
	root := t.TempDir()
	check := func(text string) error {
		writeFixture(t, root, "fuzz.log", text)
		return CheckEvidence(root, filepath.Join(root, "fuzz.log"), Evidence{GoFuzz: target})
	}
	if err := check(log); err != nil {
		t.Fatal(err)
	}
	for name, candidate := range map[string]string{
		"seed replay only":            "=== RUN   " + target + "\n" + pass,
		"no mutation inputs":          strings.Replace(log, "execs: 100", "execs: 12", 1),
		"baseline incomplete":         strings.Replace(log, "12/12", "11/12", 1),
		"no coverage instrumentation": strings.ReplaceAll(log, "gathering baseline coverage", "testing seed corpus"),
		"missing progress":            strings.Replace(log, progress, "", 1),
		"foreign target":              strings.ReplaceAll(log, target, "FuzzOther"),
		"skipped target":              strings.Replace(log, pass, "--- SKIP: "+target+" (3.00s)\n", 1),
		"failed execution":            log + "FAIL\n",
		"duplicate evidence":          log + log,
		"counter overflow":            strings.Replace(log, "execs: 100", "execs: 18446744073709551616", 1),
		"decreasing counter":          strings.Replace(log, pass, strings.Replace(progress, "execs: 100", "execs: 99", 1)+pass, 1),
		"progress after pass":         log + progress,
	} {
		t.Run(name, func(t *testing.T) {
			if err := check(candidate); err == nil {
				t.Fatal("incomplete or invalid exploration accepted")
			}
		})
	}
}

func TestNumericProfileIncludesRegressionAndDiscovery(t *testing.T) {
	plan, err := Load("../../verification_plan.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, platform := range []string{"darwin/arm64", "linux/amd64", "linux/arm64"} {
		gates, err := plan.Select("numeric", platform)
		if err != nil {
			t.Fatal(err)
		}
		seen := map[string]Gate{}
		for _, gate := range gates {
			seen[gate.ID] = gate
		}
		for _, id := range []string{"go-tests", "go-public-routing", "go-readtest", "go-dectest", "rust-tests", "finite-reference", "finite-paths", "bigdecimal", "bigdecimal-fuzz"} {
			if _, ok := seen[id]; !ok {
				t.Fatalf("numeric profile missing %s on %s", id, platform)
			}
		}
		if seen["bigdecimal-fuzz"].Evidence.GoFuzz != "FuzzFiniteArithmeticBigDecimal" || plan.Profiles["numeric"].Scope == "full" {
			t.Fatal("numeric exploration misclassified as full verification or missing fuzz evidence")
		}
	}
}
