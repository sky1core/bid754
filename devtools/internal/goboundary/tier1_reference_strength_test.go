package goboundary

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestTier1ReferenceFaultInjection(t *testing.T) {
	source, err := filepath.Abs("../../../bid754-go")
	if err != nil {
		t.Fatal(err)
	}
	module := t.TempDir()
	copyGoModuleSources(t, source, module)
	evalPath := filepath.Join(module, "internal/tier1ref/evaluate.go")
	baseline, err := os.ReadFile(evalPath)
	if err != nil {
		t.Fatal(err)
	}

	if out, err := runGo(module, "0", "test", "-count=1", "-timeout=120s", "./internal/tier1ref/"); err != nil {
		t.Fatalf("baseline model tests failed: %v\n%s", err, out)
	}
	t.Log("TIER1-REFERENCE-STRENGTH baseline: internal/tier1ref model tests pass")

	faults := []struct {
		name      string
		witness   string
		old       string
		new       string
		assertion bool
	}{
		{
			"quiet_less_predicate", "TestQuietPredicatesTruthTable",
			"return !unordered && order < 0", "return !unordered && order <= 0", true,
		},
		{
			"remainder_zero_sign", "TestRemainderWitnesses",
			"negative := rem.Sign() < 0 || rem.Sign() == 0 && a.Negative", "negative := rem.Sign() < 0", true,
		},
		{
			"to_integer_upper_bound", "TestIntegerConversionWitnesses",
			"if n.Cmp(lo) < 0 || n.Cmp(hi) > 0 {", "if n.Cmp(lo) < 0 || n.Cmp(hi) >= 0 {", true,
		},
		{
			"scaleb_shift_direction", "TestScaleWitnesses",
			"exponent.Add(exponent, shift)", "exponent.Sub(exponent, shift)", true,
		},
		{
			"underflow_flag", "TestScaleWitnesses",
			"flags |= Underflow", "flags |= 0", true,
		},
		{"crash_control", "TestScaleWitnesses", "exponent.Add(exponent, shift)", "panic(shift.String())", false},
	}

	for _, f := range faults {
		t.Run(f.name, func(t *testing.T) {
			if c := strings.Count(string(baseline), f.old); c != 1 {
				t.Fatalf("fault anchor %q occurs %d times, expected exactly 1", f.old, c)
			}
			mutant := strings.Replace(string(baseline), f.old, f.new, 1)
			if err := os.WriteFile(evalPath, []byte(mutant), 0600); err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := os.WriteFile(evalPath, baseline, 0600); err != nil {
					t.Fatal(err)
				}
			}()

			out, err := runGo(module, "0", "test", "-run", "^"+f.witness+"$", "-count=1", "-timeout=120s", "./internal/tier1ref/")
			if got := tier1WitnessAssertion(out, err, f.witness); got != f.assertion {
				t.Fatalf("mutant %s assertion=%v want %v: %v\n%s", f.name, got, f.assertion, err, out)
			}
			if !f.assertion {
				if !strings.Contains(string(out), "panic:") {
					t.Fatalf("crash control did not panic: %v\n%s", err, out)
				}
				t.Log("TIER1-REFERENCE-STRENGTH crash control rejected")
				return
			}
			t.Logf("TIER1-REFERENCE-STRENGTH mutant %s rejected by %s assertion", f.name, f.witness)
		})
	}
}

func tier1WitnessAssertion(out []byte, err error, witness string) bool {
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 1 {
		return false
	}
	s := string(out)
	for _, bad := range []string{"[build failed]", "test timed out", "Test killed"} {
		if strings.Contains(s, bad) {
			return false
		}
	}
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		for _, prefix := range []string{"panic:", "fatal error:", "signal:", "exit status "} {
			if strings.HasPrefix(line, prefix) {
				return false
			}
		}
	}
	return strings.Contains(s, "--- FAIL: "+witness+" (") &&
		regexp.MustCompile(`(?m)^\s+\S+_test\.go:[0-9]+: .* want`).MatchString(s) &&
		regexp.MustCompile(`(?m)^FAIL\s+github\.com/sky1core/bid754/bid754-go/internal/tier1ref\s`).MatchString(s)
}
