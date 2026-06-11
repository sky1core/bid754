package testgen

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type symbolFile struct {
	Symbols []symbolSpec `json:"symbols"`
}

type symbolSpec struct {
	Name        string   `json:"name"`
	LinkName    string   `json:"link_name"`
	ReturnType  string   `json:"return_type"`
	Parameters  []string `json:"parameters"`
	Declaration string   `json:"declaration"`
}

type parsedCase struct {
	ID           string
	Operation    string
	Operands     []string
	Result       string
	Flags        []string
	Precision    int
	MaxExponent  int
	MinExponent  int
	Clamp        int
	RoundingMode string
}

func buildSpec(repoRoot string, manifest Manifest) (SharedSpec, error) {
	spec := SharedSpec{
		DectestSuites: make([]GeneratedDectestSuite, 0, len(manifest.DectestSuites)),
	}

	for _, suite := range manifest.DectestSuites {
		files, err := selectDectestSuiteFiles(repoRoot, suite)
		if err != nil {
			return SharedSpec{}, err
		}
		spec.DectestSuites = append(spec.DectestSuites, GeneratedDectestSuite{
			Name:                suite.Name,
			Pattern:             suite.Pattern,
			TestType:            suite.TestType,
			Files:               files,
			SupportedOperations: append([]string(nil), suite.SupportedOperations...),
			IgnoredOperations:   append([]string(nil), suite.IgnoredOperations...),
		})
	}
	dectestAudits, err := buildDectestFileAudits(repoRoot, manifest.DectestSuites, spec.DectestSuites)
	if err != nil {
		return SharedSpec{}, err
	}
	spec.DectestFileAudits = dectestAudits
	dectestRuntimeSkipAudit, err := buildDectestRuntimeSkipAudit(repoRoot, spec.DectestSuites)
	if err != nil {
		return SharedSpec{}, err
	}
	spec.DectestRuntimeSkipAudit = dectestRuntimeSkipAudit

	for _, profile := range manifest.ReadProfiles {
		audit, err := buildReadtestProfileAudit(repoRoot, profile)
		if err != nil {
			return SharedSpec{}, err
		}
		spec.ReadtestProfileAudit = append(spec.ReadtestProfileAudit, audit)

		reads, err := expandReadTestProfile(repoRoot, profile)
		if err != nil {
			return SharedSpec{}, err
		}
		for _, read := range reads {
			if err := appendGeneratedReadCases(repoRoot, &spec, read); err != nil {
				return SharedSpec{}, err
			}
		}
	}

	for _, read := range manifest.ReadTests {
		if err := appendGeneratedReadCases(repoRoot, &spec, read); err != nil {
			return SharedSpec{}, err
		}
	}

	for _, group := range manifest.ReadTestGroups {
		for _, read := range expandReadTestGroup(group) {
			if err := appendGeneratedReadCases(repoRoot, &spec, read); err != nil {
				return SharedSpec{}, err
			}
		}
	}

	for _, fuzz := range manifest.FuzzTests {
		ops := make(map[string]struct{}, len(fuzz.Operations))
		for _, op := range fuzz.Operations {
			ops[op] = struct{}{}
		}
		selected := 0
		seen := map[string]struct{}{}
		for _, source := range fuzz.Sources {
			cases, err := parseDecTestFile(filepath.Join(repoRoot, source))
			if err != nil {
				return SharedSpec{}, err
			}
			for _, tc := range cases {
				if _, ok := ops[tc.Operation]; !ok || len(tc.Operands) != 2 {
					continue
				}
				key := strings.Join([]string{tc.Operation, tc.Operands[0], tc.Operands[1], tc.Result, fmt.Sprint(tc.Precision)}, "\x00")
				if _, ok := seen[key]; ok {
					continue
				}
				seen[key] = struct{}{}
				spec.FuzzCases = append(spec.FuzzCases, GeneratedFuzzCase{
					Suite:        fuzz.Name,
					TestType:     fuzz.TestType,
					Source:       filepath.ToSlash(source),
					ID:           tc.ID,
					Operation:    tc.Operation,
					Operands:     append([]string(nil), tc.Operands...),
					Expected:     tc.Result,
					Precision:    tc.Precision,
					RoundingMode: tc.RoundingMode,
					MaxExponent:  tc.MaxExponent,
					MinExponent:  tc.MinExponent,
					Clamp:        tc.Clamp,
				})
				selected++
				if selected == fuzz.Limit {
					break
				}
			}
			if selected == fuzz.Limit {
				break
			}
		}
		if selected == 0 {
			return SharedSpec{}, fmt.Errorf("fuzz suite %q matched no cases", fuzz.Name)
		}
	}

	for _, ffi := range manifest.FFITests {
		cases, err := buildFFICases(repoRoot, ffi)
		if err != nil {
			return SharedSpec{}, err
		}
		spec.FFICases = append(spec.FFICases, cases...)
	}

	return spec, nil
}

func buildDectestFileAudits(repoRoot string, suiteSpecs []DectestSuiteSpec, generatedSuites []GeneratedDectestSuite) ([]GeneratedDectestFileAudit, error) {
	selectedByFile := map[string][]string{}
	for _, suite := range generatedSuites {
		for _, file := range suite.Files {
			selectedByFile[file] = append(selectedByFile[file], suite.Name)
		}
	}

	dirs := map[string]struct{}{}
	for _, suite := range suiteSpecs {
		dirs[suite.Directory] = struct{}{}
	}
	var files []string
	for dir := range dirs {
		entries, err := os.ReadDir(filepath.Join(repoRoot, dir))
		if err != nil {
			return nil, fmt.Errorf("read dectest directory %q: %w", dir, err)
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".decTest") {
				continue
			}
			files = append(files, filepath.ToSlash(filepath.Join(dir, entry.Name())))
		}
	}
	sort.Strings(files)

	audits := make([]GeneratedDectestFileAudit, 0, len(files))
	for _, file := range files {
		ops, err := scanDecTestOperations(filepath.Join(repoRoot, file))
		if err != nil {
			return nil, err
		}
		operations := sortedOperationKeys(ops)
		selectedSuites := append([]string(nil), selectedByFile[file]...)
		sort.Strings(selectedSuites)

		unsupportedBySuite := map[string][]string{}
		unsupportedReasonsBySuite := map[string]map[string]string{}
		unsupportedClassificationsBySuite := map[string]map[string]string{}
		name := filepath.Base(file)
		for _, suite := range suiteSpecs {
			if !matchesDectestSuitePattern(name, suite) {
				continue
			}
			unsupported := unsupportedDectestOperations(ops, suite)
			if len(unsupported) > 0 {
				unsupportedBySuite[suite.Name] = unsupported
				unsupportedReasonsBySuite[suite.Name] = unsupportedDectestReasons(unsupported, suite)
				unsupportedClassificationsBySuite[suite.Name] = unsupportedDectestClassifications(unsupported, suite)
			}
		}
		if len(unsupportedBySuite) == 0 {
			unsupportedBySuite = nil
			unsupportedReasonsBySuite = nil
			unsupportedClassificationsBySuite = nil
		}

		audits = append(audits, GeneratedDectestFileAudit{
			File:                              file,
			Operations:                        operations,
			SelectedSuites:                    selectedSuites,
			UnsupportedBySuite:                unsupportedBySuite,
			UnsupportedReasonsBySuite:         unsupportedReasonsBySuite,
			UnsupportedClassificationsBySuite: unsupportedClassificationsBySuite,
		})
	}
	return audits, nil
}

func buildDectestRuntimeSkipAudit(repoRoot string, suites []GeneratedDectestSuite) ([]GeneratedDectestRuntimeSkipAudit, error) {
	audits := make([]GeneratedDectestRuntimeSkipAudit, 0, len(suites))
	for _, suite := range suites {
		audit := GeneratedDectestRuntimeSkipAudit{
			Suite:       suite.Name,
			SkipReasons: map[string]int{},
		}
		for _, testFile := range suite.Files {
			cases, err := parseDecTestFile(filepath.Join(repoRoot, testFile))
			if err != nil {
				return nil, fmt.Errorf("build generated dectest runtime skip audit suite %q file %q: %w", suite.Name, testFile, err)
			}
			audit.Cases += len(cases)
			for _, tc := range cases {
				if reason, ok := generatedDectestSkipReason(suite, tc); ok {
					audit.SkipReasons[reason]++
				}
			}
		}
		if len(audit.SkipReasons) == 0 {
			audit.SkipReasons = nil
		}
		audits = append(audits, audit)
	}
	return audits, nil
}

func sortedOperationKeys(ops map[string]struct{}) []string {
	keys := make([]string, 0, len(ops))
	for op := range ops {
		if op == "" {
			continue
		}
		keys = append(keys, op)
	}
	sort.Strings(keys)
	return keys
}

func unsupportedDectestOperations(ops map[string]struct{}, suite DectestSuiteSpec) []string {
	supported := normalizeOperationSet(suite.SupportedOperations)
	ignored := normalizeOperationSet(suite.IgnoredOperations)
	var unsupported []string
	for op := range ops {
		if _, ok := ignored[op]; ok {
			continue
		}
		if _, ok := supported[op]; !ok {
			unsupported = append(unsupported, op)
		}
	}
	sort.Strings(unsupported)
	return unsupported
}

func unsupportedDectestReasons(unsupported []string, suite DectestSuiteSpec) map[string]string {
	if len(unsupported) == 0 {
		return nil
	}
	reasons := make(map[string]string, len(unsupported))
	for _, op := range unsupported {
		reasons[op] = unsupportedDectestReason(op, suite)
	}
	return reasons
}

func unsupportedDectestClassifications(unsupported []string, suite DectestSuiteSpec) map[string]string {
	if len(unsupported) == 0 {
		return nil
	}
	classifications := make(map[string]string, len(unsupported))
	for _, op := range unsupported {
		classifications[op] = unsupportedDectestClassification(op, suite)
	}
	return classifications
}

func unsupportedDectestReason(op string, suite DectestSuiteSpec) string {
	if suite.TestType == "general" {
		switch op {
		case "exp", "ln", "log10", "power":
			return "general recommended math operation is outside the current mandatory BID fixed-width surface"
		case "and", "or", "xor", "invert", "rotate", "shift":
			return "general decimal logical/digit operation has no current Go BID mechanical-port public path"
		case "divideint":
			return "general integer-quotient divide operation has no current Go BID mechanical-port adapter"
		case "rescale", "trim":
			return "general GDA operation has no current BID fixed-width public surface"
		case "squareroot":
			return "general arbitrary-precision square root is not the fixed-width BID sqrt verification path"
		default:
			return "general arbitrary-precision decTest operation is not selected for the current BID fixed-width surface"
		}
	}

	switch op {
	case "canonical":
		return "tagged literal DPD/encoding canonicalization is outside the current BID-only surface"
	case "and", "or", "xor", "invert", "rotate", "shift":
		return "decimal logical/digit operation has no current Go BID mechanical-port public path"
	case "divideint":
		return "integer-quotient divide operation has no current Go BID mechanical-port adapter"
	case "reduce":
		return "Decimal128 reduce has no current Go BID mechanical-port public path"
	default:
		return "operation is not in the current generated decTest supported surface"
	}
}

func unsupportedDectestClassification(op string, suite DectestSuiteSpec) string {
	if suite.TestType == "general" {
		switch op {
		case "exp", "ln", "log10", "power":
			return "optional_not_required"
		case "and", "or", "xor", "invert", "rotate", "shift", "divideint", "rescale", "trim":
			return "out_of_scope_not_required"
		default:
			return "out_of_scope_not_required"
		}
	}

	switch op {
	case "canonical":
		return "out_of_scope_not_required"
	case "and", "or", "xor", "invert", "rotate", "shift", "divideint":
		return "out_of_scope_not_required"
	case "reduce":
		return "optional_scope_gap"
	default:
		return "unsupported_unclassified"
	}
}

func selectDectestSuiteFiles(repoRoot string, suite DectestSuiteSpec) ([]string, error) {
	dir := filepath.Join(repoRoot, suite.Directory)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read dectest directory %q: %w", suite.Directory, err)
	}

	supported := normalizeOperationSet(suite.SupportedOperations)
	ignored := normalizeOperationSet(suite.IgnoredOperations)
	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !matchesDectestSuitePattern(name, suite) {
			continue
		}

		relPath := filepath.Join(suite.Directory, name)
		ops, err := scanDecTestOperations(filepath.Join(repoRoot, relPath))
		if err != nil {
			return nil, err
		}
		if shouldSelectDecTestFile(ops, supported, ignored) {
			files = append(files, filepath.ToSlash(relPath))
		}
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("dectest suite %q matched no files", suite.Name)
	}
	return files, nil
}

func matchesDectestSuitePattern(name string, suite DectestSuiteSpec) bool {
	matched, err := filepath.Match(suite.Pattern, name)
	if err != nil || !matched {
		return false
	}
	for _, prefix := range suite.ExcludePrefixes {
		if strings.HasPrefix(name, prefix) {
			return false
		}
	}
	return true
}

func normalizeOperationSet(ops []string) map[string]struct{} {
	set := make(map[string]struct{}, len(ops))
	for _, op := range ops {
		normalized := normalizeDecTestOperation(op)
		if normalized == "" {
			continue
		}
		set[normalized] = struct{}{}
	}
	return set
}

func scanDecTestOperations(path string) (map[string]struct{}, error) {
	cases, err := parseDecTestFile(path)
	if err != nil {
		return nil, err
	}
	ops := make(map[string]struct{}, len(cases))
	for _, tc := range cases {
		ops[normalizeDecTestOperation(tc.Operation)] = struct{}{}
	}
	return ops, nil
}

func shouldSelectDecTestFile(ops, supported, ignored map[string]struct{}) bool {
	hasSupported := false
	for op := range ops {
		if _, ok := ignored[op]; ok {
			continue
		}
		if _, ok := supported[op]; !ok {
			return false
		}
		hasSupported = true
	}
	return hasSupported
}

func loadSymbolFile(path string) (symbolFile, error) {
	var symbols symbolFile

	data, err := os.ReadFile(path)
	if err != nil {
		return symbols, fmt.Errorf("read symbol file %q: %w", path, err)
	}
	if err := json.Unmarshal(data, &symbols); err != nil {
		return symbols, fmt.Errorf("parse symbol file %q: %w", path, err)
	}
	return symbols, nil
}

func parseDecTestFile(path string) ([]parsedCase, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open dectest %q: %w", path, err)
	}
	defer file.Close()

	rounding := "half_even"
	precision := 9
	maxExponent := 384
	minExponent := -383
	clamp := 0

	var cases []parsedCase
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "--") {
			continue
		}
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) != 2 {
				continue
			}
			key := strings.ToLower(strings.TrimSpace(parts[0]))
			value := strings.TrimSpace(parts[1])
			switch key {
			case "rounding":
				rounding = value
			case "precision":
				fmt.Sscanf(value, "%d", &precision)
			case "maxexponent":
				fmt.Sscanf(value, "%d", &maxExponent)
			case "minexponent":
				fmt.Sscanf(value, "%d", &minExponent)
			case "clamp":
				fmt.Sscanf(value, "%d", &clamp)
			}
			continue
		}
		if !strings.Contains(line, "->") {
			continue
		}
		parts := strings.SplitN(line, "->", 2)
		left := splitFields(parts[0])
		rightText := strings.TrimSpace(parts[1])
		if commentIdx := strings.Index(rightText, "--"); commentIdx >= 0 {
			rightText = strings.TrimSpace(rightText[:commentIdx])
		}
		right := splitFields(rightText)
		if len(left) < 2 || len(right) < 1 {
			continue
		}
		cases = append(cases, parsedCase{
			ID:           left[0],
			Operation:    left[1],
			Operands:     append([]string(nil), left[2:]...),
			Result:       right[0],
			Flags:        append([]string(nil), right[1:]...),
			Precision:    precision,
			MaxExponent:  maxExponent,
			MinExponent:  minExponent,
			Clamp:        clamp,
			RoundingMode: rounding,
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan dectest %q: %w", path, err)
	}
	return cases, nil
}

func normalizeDecTestOperation(op string) string {
	op = strings.TrimSpace(strings.ToLower(op))
	return strings.ReplaceAll(op, "_", "")
}

func splitFields(input string) []string {
	var fields []string
	var current strings.Builder
	var quote rune

	flush := func() {
		if current.Len() == 0 {
			return
		}
		fields = append(fields, current.String())
		current.Reset()
	}

	for _, r := range input {
		switch {
		case quote != 0:
			current.WriteRune(r)
			if r == quote {
				quote = 0
			}
		case r == '\'' || r == '"':
			current.WriteRune(r)
			quote = r
		case r == ' ' || r == '\t':
			flush()
		default:
			current.WriteRune(r)
		}
	}

	flush()
	return fields
}
