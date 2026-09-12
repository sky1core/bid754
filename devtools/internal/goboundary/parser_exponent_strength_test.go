package goboundary

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParserExponentFaultInjection(t *testing.T) {
	source, err := filepath.Abs("../../../bid754-go")
	if err != nil {
		t.Fatal(err)
	}
	module := t.TempDir()
	copyGoModuleSources(t, source, module)
	const witness = "TestParserExponentCancellation"
	for _, fault := range []struct {
		width, input, file, limit, cap string
	}{
		{"d32", "big_n10485760", "bid32_string.go", "len(s) + 2*DECIMAL_EXPONENT_BIAS_32", "1 << 20"},
		{"d64", "big_n10485760", "bid64_from_string.go", "len(str) + 2*DECIMAL_EXPONENT_BIAS", "1 << 20"},
		{"d128", "big_n10000000", "bid128_string.go", "len(str) + 2*EXPONENT_BIAS128", "9999999"},
	} {
		t.Run(fault.width, func(t *testing.T) {
			run := func() ([]byte, error) {
				selection := "^" + witness + "$/^" + fault.width + "$/^" + fault.input + "$"
				return runGo(module, "0", "test", "-count=1", "-v", "-timeout=120s", "-run="+selection, ".")
			}
			if out, err := run(); err != nil {
				t.Fatalf("baseline parser witness failed: %v\n%s", err, out)
			}
			path := filepath.Join(module, "internal/bidgo", fault.file)
			baseline, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Count(string(baseline), fault.limit) != 1 {
				t.Fatal("parser exponent limit mutation site changed")
			}
			mutant := strings.Replace(string(baseline), fault.limit, fault.cap, 1)
			if err := os.WriteFile(path, []byte(mutant), 0600); err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := os.WriteFile(path, baseline, 0600); err != nil {
					t.Fatal(err)
				}
			}()
			out, err := run()
			if !productionWitnessAssertion(out, err, witness, "want exact cohort") {
				t.Fatalf("parser exponent cap mutant was not rejected by the numeric assertion: %v\n%s", err, out)
			}
			t.Logf("PARSER-EXPONENT-STRENGTH width=%s cap=%s assertion=true", fault.width, fault.cap)
		})
	}
}
