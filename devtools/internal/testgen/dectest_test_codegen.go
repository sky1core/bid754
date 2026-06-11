package testgen

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	dectestGeneratedNativeTestPath = "../bid754-go/generated_dectest_cases_native_test.go"
	dectestGeneratedStubTestPath   = "../bid754-go/generated_dectest_cases_stub_test.go"
	dectestGeneratedDispatchPath   = "../bid754-go/generated_dectest_dispatch.go"
)

type dectestExecutorTemplate struct {
	OutputPath   string
	TemplatePath string
}

var dectestExecutorTemplates = []dectestExecutorTemplate{
	{OutputPath: "../bid754-go/dectest_class.go", TemplatePath: "internal/testgen/dectest_templates/dectest_class.go.tmpl"},
	{OutputPath: "../bid754-go/dectest_compare.go", TemplatePath: "internal/testgen/dectest_templates/dectest_compare.go.tmpl"},
	{OutputPath: "../bid754-go/dectest_comparetotal.go", TemplatePath: "internal/testgen/dectest_templates/dectest_comparetotal.go.tmpl"},
	{OutputPath: "../bid754-go/dectest_copy.go", TemplatePath: "internal/testgen/dectest_templates/dectest_copy.go.tmpl"},
	{OutputPath: "../bid754-go/dectest_driver.go", TemplatePath: "internal/testgen/dectest_templates/dectest_driver.go.tmpl"},
	{OutputPath: "../bid754-go/dectest_spec_test.go", TemplatePath: "internal/testgen/dectest_templates/dectest_spec_test.go.tmpl"},
	{OutputPath: "../bid754-go/dectest_fma.go", TemplatePath: "internal/testgen/dectest_templates/dectest_fma.go.tmpl"},
	{OutputPath: "../bid754-go/dectest_helpers.go", TemplatePath: "internal/testgen/dectest_templates/dectest_helpers.go.tmpl"},
	{OutputPath: "../bid754-go/dectest_logb.go", TemplatePath: "internal/testgen/dectest_templates/dectest_logb.go.tmpl"},
	{OutputPath: "../bid754-go/dectest_minmax.go", TemplatePath: "internal/testgen/dectest_templates/dectest_minmax.go.tmpl"},
	{OutputPath: "../bid754-go/dectest_native.go", TemplatePath: "internal/testgen/dectest_templates/dectest_native.go.tmpl"},
	{OutputPath: "../bid754-go/dectest_native_stub.go", TemplatePath: "internal/testgen/dectest_templates/dectest_native_stub.go.tmpl"},
	{OutputPath: "../bid754-go/dectest_next.go", TemplatePath: "internal/testgen/dectest_templates/dectest_next.go.tmpl"},
	{OutputPath: "../bid754-go/dectest_nexttoward.go", TemplatePath: "internal/testgen/dectest_templates/dectest_nexttoward.go.tmpl"},
	{OutputPath: "../bid754-go/dectest_reduce.go", TemplatePath: "internal/testgen/dectest_templates/dectest_reduce.go.tmpl"},
	{OutputPath: "../bid754-go/dectest_remainder.go", TemplatePath: "internal/testgen/dectest_templates/dectest_remainder.go.tmpl"},
	{OutputPath: "../bid754-go/dectest_remaindernear.go", TemplatePath: "internal/testgen/dectest_templates/dectest_remaindernear.go.tmpl"},
	{OutputPath: "../bid754-go/dectest_samequantum.go", TemplatePath: "internal/testgen/dectest_templates/dectest_samequantum.go.tmpl"},
	{OutputPath: "../bid754-go/dectest_scaleb.go", TemplatePath: "internal/testgen/dectest_templates/dectest_scaleb.go.tmpl"},
	{OutputPath: "../bid754-go/dectest_tointegral.go", TemplatePath: "internal/testgen/dectest_templates/dectest_tointegral.go.tmpl"},
	{OutputPath: "../bid754-go/dectest_unary.go", TemplatePath: "internal/testgen/dectest_templates/dectest_unary.go.tmpl"},
}

type dectestGeneratedSuiteCoverage struct {
	Name        string
	Files       int
	Cases       int
	SkipReasons map[string]int
}

type dectestDispatchSpec struct {
	Operations    []string
	Arity         int
	Executor      string
	ResultCompare string
	FlagCheck     string
}

var dectestDispatchSpecs = []dectestDispatchSpec{
	{
		Operations:    []string{"class"},
		Arity:         -1,
		Executor:      "executeDecTestClassOperation",
		ResultCompare: "generatedDectestCompareTokenResult",
		FlagCheck:     "generatedDectestFlagCheckNone",
	},
	{
		Operations:    []string{"samequantum"},
		Arity:         -1,
		Executor:      "executeDecTestSameQuantumOperation",
		ResultCompare: "generatedDectestCompareTokenResult",
		FlagCheck:     "generatedDectestFlagCheckNone",
	},
	{
		Operations:    []string{"nexttoward"},
		Arity:         -1,
		Executor:      "executeDecTestNextTowardOperation",
		ResultCompare: "generatedDectestCompareDecimalResult",
		FlagCheck:     "generatedDectestFlagCheckAlways",
	},
	{
		Operations:    []string{"nextplus", "nextminus"},
		Arity:         -1,
		Executor:      "executeDecTestNextOperation",
		ResultCompare: "generatedDectestCompareDecimalResult",
		FlagCheck:     "generatedDectestFlagCheckAlways",
	},
	{
		Operations:    []string{"min", "max", "minmag", "maxmag"},
		Arity:         -1,
		Executor:      "executeDecTestMinMaxOperation",
		ResultCompare: "generatedDectestCompareDecimalResult",
		FlagCheck:     "generatedDectestFlagCheckAlways",
	},
	{
		Operations:    []string{"comparetotal", "comparetotmag"},
		Arity:         -1,
		Executor:      "executeDecTestCompareTotalOperation",
		ResultCompare: "generatedDectestCompareDecimalResult",
		FlagCheck:     "generatedDectestFlagCheckAlways",
	},
	{
		Operations:    []string{"abs", "plus", "minus"},
		Arity:         -1,
		Executor:      "executeDecTestUnaryOperation",
		ResultCompare: "generatedDectestCompareDecimalResult",
		FlagCheck:     "generatedDectestFlagCheckAlways",
	},
	{
		Operations:    []string{"fma"},
		Arity:         -1,
		Executor:      "executeDecTestFMAOperation",
		ResultCompare: "generatedDectestCompareDecimalResult",
		FlagCheck:     "generatedDectestFlagCheckAlways",
	},
	{
		Operations:    []string{"logb"},
		Arity:         -1,
		Executor:      "executeDecTestLogBOperation",
		ResultCompare: "generatedDectestCompareDecimalResult",
		FlagCheck:     "generatedDectestFlagCheckAlways",
	},
	{
		Operations:    []string{"scaleb"},
		Arity:         -1,
		Executor:      "executeDecTestScaleBOperation",
		ResultCompare: "generatedDectestCompareDecimalResult",
		FlagCheck:     "generatedDectestFlagCheckAlways",
	},
	{
		Operations:    []string{"remaindernear"},
		Arity:         -1,
		Executor:      "executeDecTestRemainderNearOperation",
		ResultCompare: "generatedDectestCompareDecimalResult",
		FlagCheck:     "generatedDectestFlagCheckAlways",
	},
	{
		Operations:    []string{"remainder"},
		Arity:         -1,
		Executor:      "executeDecTestRemainderOperation",
		ResultCompare: "generatedDectestCompareDecimalResult",
		FlagCheck:     "generatedDectestFlagCheckAlways",
	},
	{
		Operations:    []string{"reduce"},
		Arity:         -1,
		Executor:      "executeDecTestReduceOperation",
		ResultCompare: "generatedDectestCompareDecimalResult",
		FlagCheck:     "generatedDectestFlagCheckAlways",
	},
	{
		Operations:    []string{"copy", "copyabs", "copynegate", "copysign"},
		Arity:         -1,
		Executor:      "executeDecTestCopyOperation",
		ResultCompare: "generatedDectestCompareDecimalResult",
		FlagCheck:     "generatedDectestFlagCheckNone",
	},
	{
		Operations:    []string{"add", "subtract", "multiply", "divide", "quantize"},
		Arity:         2,
		Executor:      "executeDecTestOperation",
		ResultCompare: "generatedDectestCompareDecimalResult",
		FlagCheck:     "generatedDectestFlagCheckNative",
	},
	{
		Operations:    []string{"compare", "comparesig"},
		Arity:         2,
		Executor:      "executeDecTestOperation",
		ResultCompare: "generatedDectestCompareDecimalResult",
		FlagCheck:     "generatedDectestFlagCheckNative",
	},
	{
		Operations:    []string{"tosci", "toeng", "tointegral", "tointegralx"},
		Arity:         1,
		Executor:      "executeDecTestReadOperation",
		ResultCompare: "generatedDectestCompareDecimalResult",
		FlagCheck:     "generatedDectestFlagCheckNative",
	},
}

func WriteDectestTestOutputs(repoRoot string, spec SharedSpec) error {
	files, err := GenerateDectestTestOutputs(repoRoot, spec)
	if err != nil {
		return err
	}
	for path, data := range files {
		fullPath := filepath.Join(repoRoot, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			return fmt.Errorf("mkdir %q: %w", filepath.Dir(fullPath), err)
		}
		if err := os.WriteFile(fullPath, data, 0o644); err != nil {
			return fmt.Errorf("write generated dectest test %q: %w", fullPath, err)
		}
	}
	return nil
}

func GenerateDectestTestOutputs(repoRoot string, spec SharedSpec) (map[string][]byte, error) {
	coverage, err := countDectestGeneratedSuiteCoverage(repoRoot, spec)
	if err != nil {
		return nil, err
	}
	dispatchSource, err := dectestDispatchSource(spec)
	if err != nil {
		return nil, err
	}
	files := map[string][]byte{
		dectestGeneratedNativeTestPath: []byte(dectestNativeTestSource(coverage)),
		dectestGeneratedDispatchPath:   []byte(dispatchSource),
		dectestGeneratedStubTestPath: []byte(`// Code generated by testgen; DO NOT EDIT.
//go:build !cgo || !bid754_native

package bid754

import "testing"

func TestGeneratedDectestSuites(t *testing.T) {
	t.Skip("generated dectest suites require cgo and bid754_native")
}
`),
	}
	executorOutputs, err := dectestExecutorTemplateOutputs(repoRoot)
	if err != nil {
		return nil, err
	}
	for path, data := range executorOutputs {
		files[path] = data
	}
	return formatGeneratedGoOutputs(files)
}

func dectestExecutorTemplateOutputs(repoRoot string) (map[string][]byte, error) {
	outputs := make(map[string][]byte, len(dectestExecutorTemplates))
	for _, item := range dectestExecutorTemplates {
		data, err := os.ReadFile(filepath.Join(repoRoot, item.TemplatePath))
		if err != nil {
			return nil, fmt.Errorf("read generated dectest executor template %q: %w", item.TemplatePath, err)
		}
		outputs[item.OutputPath] = []byte(dectestGeneratedSourceFromTemplate(data))
	}
	return outputs, nil
}

func dectestGeneratedSourceFromTemplate(data []byte) string {
	body := strings.TrimLeft(string(data), "\n")
	body = strings.TrimPrefix(body, "// Code generated by testgen; DO NOT EDIT.\n\n")
	return "// Code generated by testgen; DO NOT EDIT.\n" + body
}

func countDectestGeneratedSuiteCoverage(repoRoot string, spec SharedSpec) ([]dectestGeneratedSuiteCoverage, error) {
	coverage := make([]dectestGeneratedSuiteCoverage, 0, len(spec.DectestSuites))
	for _, suite := range spec.DectestSuites {
		item := dectestGeneratedSuiteCoverage{
			Name:        suite.Name,
			Files:       len(suite.Files),
			SkipReasons: map[string]int{},
		}
		for _, testFile := range suite.Files {
			cases, err := parseDecTestFile(filepath.Join(repoRoot, testFile))
			if err != nil {
				return nil, fmt.Errorf("count generated dectest suite %q file %q: %w", suite.Name, testFile, err)
			}
			item.Cases += len(cases)
			for _, tc := range cases {
				if reason, ok := generatedDectestSkipReason(suite, tc); ok {
					item.SkipReasons[reason]++
				}
			}
		}
		coverage = append(coverage, item)
	}
	return coverage, nil
}

func dectestSuiteCoverageLiteral(coverage []dectestGeneratedSuiteCoverage) string {
	var b strings.Builder
	for i, item := range coverage {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString("\t{\n")
		fmt.Fprintf(&b, "\t\tName:  %q,\n", item.Name)
		fmt.Fprintf(&b, "\t\tFiles: %d,\n", item.Files)
		fmt.Fprintf(&b, "\t\tCases: %d,\n", item.Cases)
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
		b.WriteString("\t},")
	}
	return b.String()
}

func dectestDispatchSource(spec SharedSpec) (string, error) {
	supportedOps := generatedDectestSupportedOperations(spec)
	dispatchSpecs, err := selectedDectestDispatchSpecs(supportedOps)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	b.WriteString(`// Code generated by testgen; DO NOT EDIT.

package bid754

import (
	"fmt"
	"path/filepath"
)

type generatedDectestRunSuite struct {
	Name              string
	Pattern           string
	TestType          string
	Files             []string
	IgnoredOperations []string
}

type generatedDectestFlagCheckMode int

const (
	generatedDectestFlagCheckNone generatedDectestFlagCheckMode = iota
	generatedDectestFlagCheckAlways
	generatedDectestFlagCheckNative
)

type generatedDectestExecFunc func(decTestCase, string) (decTestExecResult, error)
type generatedDectestResultCompareFunc func(expected, actual string) bool

`)
	b.WriteString(dectestRunSuitesSource(spec.DectestSuites))
	b.WriteByte('\n')
	b.WriteString(dectestGetTestFilesSource(spec.DectestSuites))
	b.WriteByte('\n')
	b.WriteString(dectestRunCaseSource(dispatchSpecs))
	return b.String(), nil
}

func generatedDectestSupportedOperations(spec SharedSpec) map[string]struct{} {
	ops := map[string]struct{}{}
	for _, suite := range spec.DectestSuites {
		for _, op := range suite.SupportedOperations {
			normalized := normalizeDecTestOperation(op)
			if normalized == "" {
				continue
			}
			ops[normalized] = struct{}{}
		}
	}
	return ops
}

func selectedDectestDispatchSpecs(supportedOps map[string]struct{}) ([]dectestDispatchSpec, error) {
	byOp := map[string]dectestDispatchSpec{}
	for _, spec := range dectestDispatchSpecs {
		for _, op := range spec.Operations {
			normalized := normalizeDecTestOperation(op)
			if _, exists := byOp[normalized]; exists {
				return nil, fmt.Errorf("duplicate generated dectest dispatch operation %q", normalized)
			}
			byOp[normalized] = spec
		}
	}

	missing := make([]string, 0)
	for op := range supportedOps {
		if _, ok := byOp[op]; !ok {
			missing = append(missing, op)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return nil, fmt.Errorf("generated dectest dispatch missing supported operations: %s", strings.Join(missing, ", "))
	}

	selected := make([]dectestDispatchSpec, 0, len(dectestDispatchSpecs))
	for _, spec := range dectestDispatchSpecs {
		ops := make([]string, 0, len(spec.Operations))
		for _, op := range spec.Operations {
			normalized := normalizeDecTestOperation(op)
			if _, ok := supportedOps[normalized]; ok {
				ops = append(ops, normalized)
			}
		}
		if len(ops) == 0 {
			continue
		}
		selected = append(selected, dectestDispatchSpec{
			Operations:    ops,
			Arity:         spec.Arity,
			Executor:      spec.Executor,
			ResultCompare: spec.ResultCompare,
			FlagCheck:     spec.FlagCheck,
		})
	}
	return selected, nil
}

func dectestRunSuitesSource(suites []GeneratedDectestSuite) string {
	var b strings.Builder
	b.WriteString("func loadGeneratedDectestRunSuites() ([]generatedDectestRunSuite, error) {\n")
	b.WriteString("\treturn []generatedDectestRunSuite{\n")
	for _, suite := range suites {
		b.WriteString("\t\t{\n")
		fmt.Fprintf(&b, "\t\t\tName:              %q,\n", suite.Name)
		fmt.Fprintf(&b, "\t\t\tPattern:           %q,\n", suite.Pattern)
		fmt.Fprintf(&b, "\t\t\tTestType:          %q,\n", suite.TestType)
		fmt.Fprintf(&b, "\t\t\tFiles:             []string{%s},\n", quotedStringList(suite.Files))
		fmt.Fprintf(&b, "\t\t\tIgnoredOperations: []string{%s},\n", quotedStringList(suite.IgnoredOperations))
		b.WriteString("\t\t},\n")
	}
	b.WriteString("\t}, nil\n")
	b.WriteString("}\n")
	return b.String()
}

func dectestGetTestFilesSource(suites []GeneratedDectestSuite) string {
	var b strings.Builder
	b.WriteString("func getTestFiles(dir, pattern string) ([]string, error) {\n")
	b.WriteString("\tswitch pattern {\n")
	for _, suite := range suites {
		bases := make([]string, 0, len(suite.Files))
		for _, file := range suite.Files {
			bases = append(bases, filepath.Base(file))
		}
		fmt.Fprintf(&b, "\tcase %q:\n", suite.Pattern)
		b.WriteString("\t\treturn []string{")
		for i, base := range bases {
			if i > 0 {
				b.WriteString(", ")
			}
			fmt.Fprintf(&b, "filepath.Join(dir, %q)", base)
		}
		b.WriteString("}, nil\n")
	}
	b.WriteString("\tdefault:\n")
	b.WriteString("\t\treturn nil, fmt.Errorf(\"no generated dectest suite for pattern %q\", pattern)\n")
	b.WriteString("\t}\n")
	b.WriteString("}\n")
	return b.String()
}

func dectestRunCaseSource(specs []dectestDispatchSpec) string {
	var b strings.Builder
	b.WriteString(`func runDecTestCaseV2(tc decTestCase, testType string) error {
	switch normalizeDecTestOperation(tc.Operation) {
`)
	for _, spec := range specs {
		fmt.Fprintf(&b, "\tcase %s:\n", quotedCaseList(spec.Operations))
		fmt.Fprintf(
			&b,
			"\t\treturn runGeneratedDectestCase(tc, testType, %d, %s, %s, %s)\n",
			spec.Arity,
			spec.Executor,
			spec.ResultCompare,
			spec.FlagCheck,
		)
	}
	b.WriteString(`	default:
		return fmt.Errorf("%w: %s", errUnsupportedDecTestOperation, tc.Operation)
	}
}

func runGeneratedDectestCase(
	tc decTestCase,
	testType string,
	arity int,
	exec generatedDectestExecFunc,
	resultMatches generatedDectestResultCompareFunc,
	flagCheck generatedDectestFlagCheckMode,
) error {
	if arity >= 0 && len(tc.Operands) != arity {
		return fmt.Errorf("%s requires %d operands, got %d", tc.Operation, arity, len(tc.Operands))
	}

	execResult, err := exec(tc, testType)
	if err != nil {
		return err
	}

	if !resultMatches(tc.Result, execResult.Result) {
		return fmt.Errorf("result mismatch: expected %s, got %s", tc.Result, execResult.Result)
	}

	switch flagCheck {
	case generatedDectestFlagCheckNone:
		return nil
	case generatedDectestFlagCheckAlways:
		if !compareDecTestFlags(tc.Flags, execResult.Flags) {
			return fmt.Errorf("flag mismatch: expected %v, got %s", tc.Flags, execResult.Flags.String())
		}
	case generatedDectestFlagCheckNative:
		if supportsDecTestFlagVerification() && !compareDecTestFlags(tc.Flags, execResult.Flags) {
			return fmt.Errorf("flag mismatch: expected %v, got %s", tc.Flags, execResult.Flags.String())
		}
	default:
		return fmt.Errorf("unsupported generated decTest flag check mode: %d", flagCheck)
	}

	return nil
}

func generatedDectestCompareDecimalResult(expected, actual string) bool {
	return compareDecimalResults(expected, actual)
}

func generatedDectestCompareTokenResult(expected, actual string) bool {
	return compareDecTestTokenResult(expected, actual)
}
`)
	return b.String()
}

func quotedStringList(values []string) string {
	var b strings.Builder
	for i, value := range values {
		if i > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "%q", value)
	}
	return b.String()
}

func quotedCaseList(values []string) string {
	return quotedStringList(values)
}

func dectestNativeTestSource(coverage []dectestGeneratedSuiteCoverage) string {
	return strings.NewReplacer(
		"@@DECTEST_SUITE_COVERAGE@@", dectestSuiteCoverageLiteral(coverage),
	).Replace(`// Code generated by testgen; DO NOT EDIT.
//go:build cgo && bid754_native

package bid754

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/sky1core/bid754/bid754-go/internal/testspec"
)

type generatedDectestSuiteCoverage struct {
	Name        string
	Files       int
	Cases       int
	SkipReasons map[string]int
}

var expectedGeneratedDectestSuiteCoverage = []generatedDectestSuiteCoverage{
@@DECTEST_SUITE_COVERAGE@@
}

func TestGeneratedDectestSuites(t *testing.T) {
	requireNative(t)
	if testing.Short() {
		t.Skip("IBM decTest suite is long-running and is skipped under -short.")
	}

	spec := loadGeneratedDectestSpecForTest(t)
	if len(spec.DectestSuites) == 0 {
		t.Fatal("expected generated dectest suites")
	}
	assertGeneratedDectestSuiteCoverage(t, spec.DectestSuites)
	assertGeneratedDectestRuntimeSkipAudit(t, spec.DectestRuntimeSkipAudit)

	var totals []decTestSuiteTotals
	for _, suite := range spec.DectestSuites {
		suite := suite
		t.Run(suite.Name, func(t *testing.T) {
			totals = append(totals, runGeneratedDectestSuite(t, suite))
		})
	}

	if err := decTestFailureError(totals); err != nil {
		t.Fatal(err)
	}
}

func assertGeneratedDectestSuiteCoverage(t *testing.T, suites []testspec.GeneratedDectestSuite) {
	t.Helper()
	if len(suites) != len(expectedGeneratedDectestSuiteCoverage) {
		t.Fatalf("generated dectest suite count = %d, want %d", len(suites), len(expectedGeneratedDectestSuiteCoverage))
	}
	for i, suite := range suites {
		expected := expectedGeneratedDectestSuiteCoverage[i]
		if suite.Name != expected.Name {
			t.Fatalf("generated dectest suite[%d] = %q, want %q", i, suite.Name, expected.Name)
		}
		if len(suite.Files) != expected.Files {
			t.Fatalf("generated dectest suite %q file count = %d, want %d", suite.Name, len(suite.Files), expected.Files)
		}
		gotCases := 0
		for _, testFile := range suite.Files {
			cases, err := parseDecTestFile(filepath.Join("..", "devtools", testFile))
			if err != nil {
				t.Fatalf("parseDecTestFile(%q): %v", testFile, err)
			}
			gotCases += len(cases)
		}
		if gotCases != expected.Cases {
			t.Fatalf("generated dectest suite %q raw case count = %d, want %d", suite.Name, gotCases, expected.Cases)
		}
	}
}

func assertGeneratedDectestRuntimeSkipAudit(t *testing.T, audits []testspec.GeneratedDectestRuntimeSkipAudit) {
	t.Helper()
	if len(audits) != len(expectedGeneratedDectestSuiteCoverage) {
		t.Fatalf("generated dectest runtime skip audit count = %d, want %d", len(audits), len(expectedGeneratedDectestSuiteCoverage))
	}
	for i, audit := range audits {
		expected := expectedGeneratedDectestSuiteCoverage[i]
		if audit.Suite != expected.Name {
			t.Fatalf("generated dectest runtime skip audit[%d] suite = %q, want %q", i, audit.Suite, expected.Name)
		}
		if audit.Cases != expected.Cases {
			t.Fatalf("generated dectest runtime skip audit suite %q cases = %d, want %d", audit.Suite, audit.Cases, expected.Cases)
		}
		assertGeneratedDectestSkipReasons(t, audit.Suite, audit.SkipReasons, expected.SkipReasons)
	}
}

func expectedGeneratedDectestCoverageForSuite(t *testing.T, name string) generatedDectestSuiteCoverage {
	t.Helper()
	for _, expected := range expectedGeneratedDectestSuiteCoverage {
		if expected.Name == name {
			return expected
		}
	}
	t.Fatalf("generated dectest suite %q missing expected coverage", name)
	return generatedDectestSuiteCoverage{}
}

func assertGeneratedDectestSkipReasons(t *testing.T, suite string, got, want map[string]int) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("generated dectest suite %q skip reason bucket count = %d, want %d", suite, len(got), len(want))
	}
	for key, wantValue := range want {
		if got[key] != wantValue {
			t.Fatalf("generated dectest suite %q skip reason count[%q] = %d, want %d", suite, key, got[key], wantValue)
		}
	}
}

func loadGeneratedDectestSpecForTest(t *testing.T) testspec.SharedSpec {
	t.Helper()
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("resolve generated dectest file path")
	}
	spec, err := testspec.LoadGenerated(filepath.Join(filepath.Dir(currentFile), "..", "devtools", "generated", "testspec", "spec_index.json"))
	if err != nil {
		t.Fatalf("load shared spec: %v", err)
	}
	return spec
}

func runGeneratedDectestSuite(t *testing.T, suite testspec.GeneratedDectestSuite) decTestSuiteTotals {
	t.Helper()

	result := decTestSuiteTotals{Name: suite.Name}
	skipReasons := map[string]int{}
	for _, testFile := range suite.Files {
		cases, err := parseDecTestFile(filepath.Join("..", "devtools", testFile))
		if err != nil {
			t.Fatalf("parseDecTestFile(%q): %v", testFile, err)
		}
		for _, tc := range cases {
			if reason, ok := decTestSkipReason(suite.IgnoredOperations, tc, suite.TestType); ok {
				result.Skipped++
				skipReasons[reason]++
				continue
			}
			if err := runDecTestCaseV2(tc, suite.TestType); err != nil {
				result.Failed++
				t.Logf("generated dectest case %s (%s) failed: %v", tc.ID, suite.TestType, err)
				continue
			}
			result.Passed++
		}
	}
	expected := expectedGeneratedDectestCoverageForSuite(t, suite.Name)
	assertGeneratedDectestSkipReasons(t, suite.Name, skipReasons, expected.SkipReasons)
	t.Logf("%s: passed=%d failed=%d skipped=%d", suite.Name, result.Passed, result.Failed, result.Skipped)
	return result
}
`)
}
