package testgen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sky1core/bid754/devtools/internal/genmarker"
)

// The goport decTest outputs run the fixed-width Decimal32/64/128 oracle-dispatch
// operation set directly against the Go BID mechanical port (no cgo), so the
// portable CI legs cross-check the product port against the IBM decTest expected
// values as an independent second data source. Phase 2 asserts, per executed
// case, the exact value, the exact quantum (cohort member), and IEEE
// exception-flag parity on the BID five-flag surface, with the documented
// decNumber-vs-Intel flag-divergence classes recorded as flag exemptions
// (value and quantum still compared) in a closed per-suite accounting.
const (
	dectestGoportDispatchPath         = "../bid754-go/generated_dectest_goport_dispatch_test.go"
	dectestGoportCasesPath            = "../bid754-go/generated_dectest_goport_cases_test.go"
	dectestGoportDispatchTemplatePath = "internal/testgen/dectest_templates/dectest_goport_dispatch.go.tmpl"
)

type dectestGoportSuiteCoverage struct {
	Name        string
	Cases       int
	Executed    int
	SkipReasons map[string]int
	// FlagExempt counts, per documented decNumber-vs-Intel flag-divergence
	// class, the EXECUTED cases whose expected flags are not compared (value
	// and quantum still are). Exempt cases stay inside Executed, so
	// executed = cases - sum(SkipReasons) keeps holding.
	FlagExempt map[string]int
}

func WriteDectestGoportOutputs(repoRoot string, spec SharedSpec) error {
	files, err := GenerateDectestGoportOutputs(repoRoot, spec)
	if err != nil {
		return err
	}
	for path, data := range files {
		fullPath := filepath.Join(repoRoot, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			return fmt.Errorf("mkdir %q: %w", filepath.Dir(fullPath), err)
		}
		if err := os.WriteFile(fullPath, data, 0o644); err != nil {
			return fmt.Errorf("write generated dectest goport output %q: %w", fullPath, err)
		}
	}
	return nil
}

func GenerateDectestGoportOutputs(repoRoot string, spec SharedSpec) (map[string][]byte, error) {
	coverage, err := countDectestGoportSuiteCoverage(repoRoot, spec)
	if err != nil {
		return nil, err
	}
	dispatchData, err := os.ReadFile(filepath.Join(repoRoot, dectestGoportDispatchTemplatePath))
	if err != nil {
		return nil, fmt.Errorf("read generated dectest goport dispatch template %q: %w", dectestGoportDispatchTemplatePath, err)
	}
	files := map[string][]byte{
		dectestGoportDispatchPath: []byte(dectestGeneratedSourceFromTemplate(dispatchData)),
		dectestGoportCasesPath:    []byte(dectestGoportCasesSource(coverage)),
	}
	return formatGeneratedGoOutputs(files)
}

func countDectestGoportSuiteCoverage(repoRoot string, spec SharedSpec) ([]dectestGoportSuiteCoverage, error) {
	coverage := make([]dectestGoportSuiteCoverage, 0, len(spec.DectestSuites))
	for _, suite := range spec.DectestSuites {
		if !isGoportDectestSuite(suite) {
			continue
		}
		item := dectestGoportSuiteCoverage{
			Name:        suite.Name,
			SkipReasons: map[string]int{},
			FlagExempt:  map[string]int{},
		}
		for _, testFile := range suite.Files {
			cases, err := parseDecTestFile(filepath.Join(repoRoot, testFile))
			if err != nil {
				return nil, fmt.Errorf("count generated dectest goport suite %q file %q: %w", suite.Name, testFile, err)
			}
			item.Cases += len(cases)
			for _, tc := range cases {
				if reason, ok := generatedDectestGoportSkipReason(suite, tc); ok {
					item.SkipReasons[reason]++
					continue
				}
				if reason, ok := generatedDectestGoportFlagExemptReason(tc); ok {
					item.FlagExempt[reason]++
				}
			}
		}
		item.Executed = item.Cases
		for _, n := range item.SkipReasons {
			item.Executed -= n
		}
		coverage = append(coverage, item)
	}
	return coverage, nil
}

func dectestGoportSuiteCoverageLiteral(coverage []dectestGoportSuiteCoverage) string {
	var b strings.Builder
	for i, item := range coverage {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString("\t{\n")
		fmt.Fprintf(&b, "\t\tName:     %q,\n", item.Name)
		fmt.Fprintf(&b, "\t\tCases:    %d,\n", item.Cases)
		fmt.Fprintf(&b, "\t\tExecuted: %d,\n", item.Executed)
		if len(item.SkipReasons) == 0 {
			b.WriteString("\t\tSkipReasons: map[string]int{},\n")
		} else {
			b.WriteString("\t\tSkipReasons: map[string]int{\n")
			for _, line := range strings.Split(stringIntMapLiteral(item.SkipReasons), "\n") {
				b.WriteString("\t\t")
				b.WriteString(line)
				b.WriteByte('\n')
			}
			b.WriteString("\t\t},\n")
		}
		if len(item.FlagExempt) == 0 {
			b.WriteString("\t\tFlagExempt: map[string]int{},\n")
		} else {
			b.WriteString("\t\tFlagExempt: map[string]int{\n")
			for _, line := range strings.Split(stringIntMapLiteral(item.FlagExempt), "\n") {
				b.WriteString("\t\t")
				b.WriteString(line)
				b.WriteByte('\n')
			}
			b.WriteString("\t\t},\n")
		}
		b.WriteString("\t},")
	}
	return b.String()
}

func dectestGoportCasesSource(coverage []dectestGoportSuiteCoverage) string {
	return strings.NewReplacer(
		"@@GOPORT_DECTEST_SUITE_COVERAGE@@", dectestGoportSuiteCoverageLiteral(coverage),
	).Replace(dectestGoportCasesTemplate)
}

var dectestGoportCasesTemplate = genmarker.Line("testgen") + `

package bid754

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/sky1core/bid754/bid754-go/internal/testspec"
)

// goportDectestSuiteCoverage pins, per fixed-width suite, the raw decTest case total,
// the executed oracle-op count, the mechanical skip-reason buckets, and the
// flag-exemption buckets of the portable Go mechanical-port leg. Executed +
// sum(SkipReasons) == Cases (exempt cases stay executed: only their flag
// comparison is waived, value and quantum are still asserted), so a generator
// that silently drops executed cases or widens an exemption moves a count here
// (and the external anchor) instead of passing quietly.
type goportDectestSuiteCoverage struct {
	Name        string
	Cases       int
	Executed    int
	SkipReasons map[string]int
	FlagExempt  map[string]int
}

var expectedGoportDectestSuiteCoverage = []goportDectestSuiteCoverage{
@@GOPORT_DECTEST_SUITE_COVERAGE@@
}

// TestGeneratedDectestSuitesGoPort cross-checks the Go BID mechanical port against the
// IBM decTest expected values for the fixed-width Decimal32/64/128 oracle-dispatch
// operation set (add/subtract/multiply/divide/quantize/compare/comparesig/tosci/toeng/
// tointegral/tointegralx). It is portable (no cgo, no build tags), so it runs in every
// non-short "go test ./..." of bid754-go and under make test-portable-dectest. This is
// an independent second-source cross-validation of operations already anchored by the
// Intel readtest and C FFI bit-compare domains; it does not replace them. Phase 2
// asserts, per executed case: the exact value (shared decTest value comparator),
// the exact quantum (cohort member: sign, integer coefficient, exponent), and
// IEEE exception-flag parity on the BID five-flag surface, where the case flags
// accumulate the operand-parse flags and the operation flags. Cases in a
// documented decNumber-vs-Intel flag-divergence class keep their value and
// quantum assertions but waive the flag comparison, counted per class against
// the pinned FlagExempt buckets.
func TestGeneratedDectestSuitesGoPort(t *testing.T) {
	if testing.Short() {
		t.Skip("goport decTest cases run in non-short mode; use make test-portable-dectest")
	}
	spec := goportLoadGeneratedDectestSpec(t)
	if len(spec.DectestSuites) == 0 {
		t.Fatal("expected generated dectest suites")
	}
	assertGoportDectestRuntimeSkipInventory(t, spec.DectestGoportRuntimeSkipInventory)

	// Input-absence policy (docs/TEST_GENERATION_SPEC.md portable-input rule: when
	// generation inputs such as devtools/tests/*.decTest are absent, input-dependent
	// regeneration sync tests are explicitly skipped). The pinned decTest inputs are
	// gitignored downloads, absent from a checkout that has not run
	// 'make setup-generation-inputs'. Probe every referenced input up front; if any
	// is absent, SKIP the whole portable leg rather than failing, mirroring the Rust
	// sibling leg generated_dectest_suites_go_port (bid754-rs/tests/dectest_generated.rs).
	// Skip is keyed ONLY on absence (os.IsNotExist): any other stat fault fails here,
	// and a present-but-unreadable/malformed input still fails at parse time below --
	// absence is never absorbed into a skip once the inputs are present, and whenever
	// the inputs ARE present the leg runs its full corpus. The checked-in runtime skip
	// verification above is input-independent (it reads the checked-in spec), so it still runs.
	for _, suite := range spec.DectestSuites {
		if !isGoportDectestRunnerSuite(suite.TestType) {
			continue
		}
		for _, testFile := range suite.Files {
			path := filepath.Join("..", "devtools", testFile)
			if _, err := os.Stat(path); err != nil {
				if os.IsNotExist(err) {
					t.Skipf("skipping goport decTest leg: pinned decTest input %q absent -- run 'make setup-generation-inputs' first", path)
				}
				t.Fatalf("stat pinned decTest input %q: %v", path, err)
			}
		}
	}

	totalFailed := 0
	for _, suite := range spec.DectestSuites {
		if !isGoportDectestRunnerSuite(suite.TestType) {
			continue
		}
		suite := suite
		t.Run(suite.Name, func(t *testing.T) {
			totalFailed += runGoportDectestSuite(t, suite)
		})
	}
	if totalFailed != 0 {
		t.Fatalf("goport decTest value/quantum/flag cross-check: %d case(s) diverged from the IBM expected case", totalFailed)
	}
}

func runGoportDectestSuite(t *testing.T, suite testspec.GeneratedDectestSuite) int {
	t.Helper()
	expected := expectedGoportDectestCoverageForSuite(t, suite.Name)
	executed := 0
	failed := 0
	logged := 0
	skipReasons := map[string]int{}
	flagExempt := map[string]int{}
	logf := func(format string, args ...any) {
		if logged < 25 {
			logged++
			t.Errorf(format, args...)
		}
	}
	for _, testFile := range suite.Files {
		// TestGeneratedDectestSuitesGoPort probes every referenced input for
		// existence up front and SKIPS the whole leg when a pinned
		// devtools/tests/*.decTest input is absent, so this error path is reached
		// only for a present-but-unreadable/malformed file -- a genuine read or
		// parse fault, not an unset-up checkout. Absence is a skip (handled by the
		// caller); a present-but-unreadable file still fails here.
		cases, err := parseDecTestFile(filepath.Join("..", "devtools", testFile))
		if err != nil {
			t.Fatalf("parseDecTestFile(%q): %v", testFile, err)
		}
		for _, tc := range cases {
			if reason, ok := dectestGoportSkipReason(suite.IgnoredOperations, tc); ok {
				skipReasons[reason]++
				continue
			}
			executed++
			got, gotFlags, err := runDectestGoportCase(tc, suite.TestType)
			if err != nil {
				failed++
				logf("goport decTest case %s (%s) execution error: %v", tc.ID, suite.TestType, err)
				continue
			}
			if !compareDecimalResults(tc.Result, got) {
				failed++
				logf("goport decTest case %s (%s) value mismatch [rounding %s]: expected %q, port produced %q", tc.ID, suite.TestType, tc.RoundingMode, tc.Result, got)
				continue
			}
			if dectestGoportQuantumComparedOperation(normalizeDecTestOperation(tc.Operation)) && !dectestGoportQuantumEqual(tc.Result, got) {
				failed++
				logf("goport decTest case %s (%s) quantum mismatch [rounding %s]: expected %q, port produced %q (same value, different cohort member)", tc.ID, suite.TestType, tc.RoundingMode, tc.Result, got)
				continue
			}
			// Condition-token validation runs BEFORE the exemption classifier:
			// an exemption only waives the flag comparison, never the
			// strict check that every expected condition token is a
			// recognized one.
			expectedFlags, ok := dectestGoportExpectedFlags(tc)
			if !ok {
				failed++
				logf("goport decTest case %s (%s) carries an unrecognized condition token: %v", tc.ID, suite.TestType, tc.Flags)
				continue
			}
			if reason, ok := dectestGoportFlagExemptReason(tc); ok {
				flagExempt[reason]++
				continue
			}
			if gotFlags&dectestGoportBIDFlagMask != expectedFlags {
				failed++
				logf("goport decTest case %s (%s) flag mismatch [rounding %s]: conditions %v project to %s, port raised %s", tc.ID, suite.TestType, tc.RoundingMode, tc.Flags, expectedFlags.String(), (gotFlags & dectestGoportBIDFlagMask).String())
			}
		}
	}
	assertGoportDectestSkipReasons(t, suite.Name, skipReasons, expected.SkipReasons)
	assertGoportDectestSkipReasons(t, suite.Name+" flag-exempt", flagExempt, expected.FlagExempt)
	if executed != expected.Executed {
		t.Errorf("goport decTest suite %q executed = %d, want %d", suite.Name, executed, expected.Executed)
	}
	t.Logf("%s: executed=%d failed=%d skipped=%d flagExempt=%d", suite.Name, executed, failed, goportSumSkipReasons(skipReasons), goportSumSkipReasons(flagExempt))
	return failed
}

// assertGoportDectestRuntimeSkipInventory checks the checked-in generated inventory
// (dectest_goport_runtime_skip_inventory) against the in-runner pinned coverage, so a
// self-consistent generator regression that relaxes both the inventory and its own runner
// pins together still fails the external verification_anchors count check.
func assertGoportDectestRuntimeSkipInventory(t *testing.T, inventories []testspec.GeneratedDectestRuntimeSkipInventory) {
	t.Helper()
	if len(inventories) != len(expectedGoportDectestSuiteCoverage) {
		t.Fatalf("goport dectest runtime skip inventory count = %d, want %d", len(inventories), len(expectedGoportDectestSuiteCoverage))
	}
	for i, inventory := range inventories {
		expected := expectedGoportDectestSuiteCoverage[i]
		if inventory.Suite != expected.Name {
			t.Fatalf("goport dectest runtime skip inventory[%d] suite = %q, want %q", i, inventory.Suite, expected.Name)
		}
		if inventory.Cases != expected.Cases {
			t.Fatalf("goport dectest runtime skip inventory suite %q cases = %d, want %d", inventory.Suite, inventory.Cases, expected.Cases)
		}
		assertGoportDectestSkipReasons(t, inventory.Suite, inventory.SkipReasons, expected.SkipReasons)
		assertGoportDectestSkipReasons(t, inventory.Suite+" flag-exempt", inventory.FlagExemptions, expected.FlagExempt)
		if got := inventory.Cases - goportSumSkipReasons(inventory.SkipReasons); got != expected.Executed {
			t.Fatalf("goport dectest runtime skip inventory suite %q executed (cases-skips) = %d, want %d", inventory.Suite, got, expected.Executed)
		}
	}
}

func expectedGoportDectestCoverageForSuite(t *testing.T, name string) goportDectestSuiteCoverage {
	t.Helper()
	for _, expected := range expectedGoportDectestSuiteCoverage {
		if expected.Name == name {
			return expected
		}
	}
	t.Fatalf("goport dectest suite %q missing expected coverage", name)
	return goportDectestSuiteCoverage{}
}

func assertGoportDectestSkipReasons(t *testing.T, suite string, got, want map[string]int) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("goport dectest suite %q skip reason bucket count = %d, want %d (got %v, want %v)", suite, len(got), len(want), got, want)
	}
	for key, wantValue := range want {
		if got[key] != wantValue {
			t.Fatalf("goport dectest suite %q skip reason count[%q] = %d, want %d", suite, key, got[key], wantValue)
		}
	}
}

func goportSumSkipReasons(reasons map[string]int) int {
	total := 0
	for _, n := range reasons {
		total += n
	}
	return total
}

func goportLoadGeneratedDectestSpec(t *testing.T) testspec.SharedSpec {
	t.Helper()
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("resolve generated dectest goport file path")
	}
	spec, err := testspec.LoadGenerated(filepath.Join(filepath.Dir(currentFile), "..", "devtools", "generated", "testspec", "spec_index.json"))
	if err != nil {
		t.Fatalf("load shared spec: %v", err)
	}
	return spec
}
`
