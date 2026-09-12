package goboundary

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestTier1ConversionFaultInjection(t *testing.T) {
	source, err := filepath.Abs("../../../bid754-go")
	if err != nil {
		t.Fatal(err)
	}
	module := t.TempDir()
	copyGoModuleSources(t, source, module)
	const witness = "TestTier1RoundingBoundaryGo"
	run := func() ([]byte, error) {
		return runGo(module, "0", "test", "-count=1", "-v", "-timeout=120s", "-run=^"+witness+"$", ".")
	}
	if out, err := run(); err != nil {
		t.Fatalf("baseline conversion tests failed: %v\n%s", err, out)
	}
	for _, fault := range []struct {
		name, file, function, body string
		assertion                  bool
	}{
		{"signed_double_rounding", "bid32_to_int.go", "Bid32FromInt64", "r, f := Bid64FromInt64(x, rnd_mode); v, g := Bid64ToBid32(r, rnd_mode); return v, f | g", true},
		{"unsigned_double_rounding", "bid32_to_int.go", "Bid32FromUint64", "r, f := Bid64FromUint64(x, rnd_mode); v, g := Bid64ToBid32(r, rnd_mode); return v, f | g", true},
		{"narrowing_double_rounding", "bid128_conversions.go", "Bid128ToBid32", "r, f := Bid128ToBid64(x, rnd_mode); v, g := Bid64ToBid32(r, rnd_mode); return v, f | g", true},
		{"crash_control", "bid32_to_int.go", "Bid32FromInt64", "panic(x)", false},
	} {
		t.Run(fault.name, func(t *testing.T) {
			path := filepath.Join(module, "internal/bidgo", fault.file)
			baseline, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			mutant := replaceProductionFunctionBody(t, baseline, fault.function, fault.body)
			if err := os.WriteFile(path, mutant, 0600); err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := os.WriteFile(path, baseline, 0600); err != nil {
					t.Fatal(err)
				}
			}()
			out, err := run()
			if got := productionWitnessAssertion(out, err, witness, "decimal numeric value differs"); got != fault.assertion {
				t.Fatalf("mutant %s assertion=%v want %v: %v\n%s", fault.name, got, fault.assertion, err, out)
			}
			if !fault.assertion && !strings.Contains(string(out), "panic:") {
				t.Fatalf("crash control did not panic: %v\n%s", err, out)
			}
			t.Logf("TIER1-CONVERSION-STRENGTH mutant=%s assertion=%v", fault.name, fault.assertion)
		})
	}
}

func replaceProductionFunctionBody(t *testing.T, source []byte, name, body string) []byte {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "production.go", source, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != name || fn.Body == nil {
			continue
		}
		start, end := fset.Position(fn.Body.Pos()).Offset, fset.Position(fn.Body.End()).Offset
		return []byte(string(source[:start]) + "{\n" + body + "\n}" + string(source[end:]))
	}
	t.Fatalf("production function %s not found", name)
	return nil
}

func productionWitnessAssertion(out []byte, err error, witness, diagnostic string) bool {
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 1 {
		return false
	}
	s := string(out)
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		for _, bad := range []string{"panic:", "fatal error:", "signal:", "exit status "} {
			if strings.HasPrefix(line, bad) {
				return false
			}
		}
	}
	return !strings.Contains(s, "[build failed]") && !strings.Contains(s, "test timed out") &&
		strings.Contains(s, "--- FAIL: "+witness+" (") &&
		strings.Contains(s, diagnostic) &&
		strings.Contains(s, "FAIL\tgithub.com/sky1core/bid754/bid754-go\t")
}
