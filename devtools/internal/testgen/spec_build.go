package testgen

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
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
	dectestInventories, err := buildDectestFileInventories(repoRoot, manifest.DectestSuites, spec.DectestSuites)
	if err != nil {
		return SharedSpec{}, err
	}
	spec.DectestFileInventories = dectestInventories
	dectestRuntimeSkipInventory, err := buildDectestRuntimeSkipInventory(repoRoot, spec.DectestSuites)
	if err != nil {
		return SharedSpec{}, err
	}
	spec.DectestRuntimeSkipInventory = dectestRuntimeSkipInventory
	dectestGoportRuntimeSkipInventory, err := buildDectestGoportRuntimeSkipInventory(repoRoot, spec.DectestSuites)
	if err != nil {
		return SharedSpec{}, err
	}
	spec.DectestGoportRuntimeSkipInventory = dectestGoportRuntimeSkipInventory

	for _, profile := range manifest.ReadProfiles {
		inventory, err := buildReadtestProfileInventory(repoRoot, profile)
		if err != nil {
			return SharedSpec{}, err
		}

		reads, err := expandReadTestProfile(repoRoot, profile)
		if err != nil {
			return SharedSpec{}, err
		}
		rowSkips := map[string]int{}
		accepted := 0
		for _, read := range reads {
			n, skips, err := appendGeneratedReadCases(repoRoot, &spec, read)
			if err != nil {
				return SharedSpec{}, err
			}
			accepted += n
			for reason, count := range skips {
				rowSkips[reason] += count
			}
		}
		inventory.RowsAccepted = accepted
		if len(rowSkips) > 0 {
			inventory.RowSkipReasons = rowSkips
		}
		spec.ReadtestProfileInventory = append(spec.ReadtestProfileInventory, inventory)
	}

	for _, read := range manifest.ReadTests {
		if _, _, err := appendGeneratedReadCases(repoRoot, &spec, read); err != nil {
			return SharedSpec{}, err
		}
	}

	for _, group := range manifest.ReadTestGroups {
		for _, read := range expandReadTestGroup(group) {
			if _, _, err := appendGeneratedReadCases(repoRoot, &spec, read); err != nil {
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

func buildDectestFileInventories(repoRoot string, suiteSpecs []DectestSuiteSpec, generatedSuites []GeneratedDectestSuite) ([]GeneratedDectestFileInventory, error) {
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

	inventories := make([]GeneratedDectestFileInventory, 0, len(files))
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

		inventories = append(inventories, GeneratedDectestFileInventory{
			File:                              file,
			Operations:                        operations,
			SelectedSuites:                    selectedSuites,
			UnsupportedBySuite:                unsupportedBySuite,
			UnsupportedReasonsBySuite:         unsupportedReasonsBySuite,
			UnsupportedClassificationsBySuite: unsupportedClassificationsBySuite,
		})
	}
	return inventories, nil
}

func buildDectestRuntimeSkipInventory(repoRoot string, suites []GeneratedDectestSuite) ([]GeneratedDectestRuntimeSkipInventory, error) {
	inventories := make([]GeneratedDectestRuntimeSkipInventory, 0, len(suites))
	for _, suite := range suites {
		inventory := GeneratedDectestRuntimeSkipInventory{
			Suite:       suite.Name,
			SkipReasons: map[string]int{},
		}
		for _, testFile := range suite.Files {
			cases, err := parseDecTestFile(filepath.Join(repoRoot, testFile))
			if err != nil {
				return nil, fmt.Errorf("build generated dectest runtime skip inventory suite %q file %q: %w", suite.Name, testFile, err)
			}
			inventory.Cases += len(cases)
			for _, tc := range cases {
				if reason, ok := generatedDectestSkipReason(suite, tc); ok {
					inventory.SkipReasons[reason]++
				}
			}
		}
		if len(inventory.SkipReasons) == 0 {
			inventory.SkipReasons = nil
		}
		inventories = append(inventories, inventory)
	}
	return inventories, nil
}

// buildDectestGoportRuntimeSkipInventory mirrors buildDectestRuntimeSkipInventory for
// the portable Go mechanical-port cross-check leg. It covers only the
// fixed-width Decimal32/64/128 suites (General is excluded whole-suite because
// its arbitrary-precision context does not map onto the fixed-width BID port),
// and buckets every non-executed case with a mechanical reason so
// executed = cases - sum(skip_reasons) stays a closed accounting per suite.
func buildDectestGoportRuntimeSkipInventory(repoRoot string, suites []GeneratedDectestSuite) ([]GeneratedDectestRuntimeSkipInventory, error) {
	inventories := make([]GeneratedDectestRuntimeSkipInventory, 0, len(suites))
	for _, suite := range suites {
		if !isGoportDectestSuite(suite) {
			continue
		}
		inventory := GeneratedDectestRuntimeSkipInventory{
			Suite:          suite.Name,
			SkipReasons:    map[string]int{},
			FlagExemptions: map[string]int{},
		}
		for _, testFile := range suite.Files {
			cases, err := parseDecTestFile(filepath.Join(repoRoot, testFile))
			if err != nil {
				return nil, fmt.Errorf("build generated dectest goport runtime skip inventory suite %q file %q: %w", suite.Name, testFile, err)
			}
			inventory.Cases += len(cases)
			for _, tc := range cases {
				if reason, ok := generatedDectestGoportSkipReason(suite, tc); ok {
					inventory.SkipReasons[reason]++
					continue
				}
				if reason, ok := generatedDectestGoportFlagExemptReason(tc); ok {
					inventory.FlagExemptions[reason]++
				}
			}
		}
		if len(inventory.SkipReasons) == 0 {
			inventory.SkipReasons = nil
		}
		if len(inventory.FlagExemptions) == 0 {
			inventory.FlagExemptions = nil
		}
		inventories = append(inventories, inventory)
	}
	return inventories, nil
}

// isGoportDectestSuite reports whether a suite is a fixed-width BID suite that
// the portable Go mechanical-port decTest leg iterates.
func isGoportDectestSuite(suite GeneratedDectestSuite) bool {
	switch suite.TestType {
	case "decimal32", "decimal64", "decimal128":
		return true
	default:
		return false
	}
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

func scanDecTestLine(line string) (contentEnd int, arrow int, directiveColon int, err error) {
	contentEnd = len(line)
	arrow = -1
	directiveColon = -1
	var quote byte

	for index := 0; index < len(line); index++ {
		char := line[index]
		if quote != 0 {
			if char == quote {
				quote = 0
			}
			continue
		}
		if char == '\'' || char == '"' {
			quote = char
			continue
		}
		if char == ':' && directiveColon < 0 {
			directiveColon = index
			continue
		}
		if index+1 >= len(line) || char != '-' {
			continue
		}

		switch line[index+1] {
		case '-':
			return index, arrow, directiveColon, nil
		case '>':
			if arrow >= 0 {
				return 0, 0, 0, fmt.Errorf("multiple -> separators")
			}
			arrow = index
			index++
		}
	}

	if quote != 0 {
		return 0, 0, 0, fmt.Errorf("unterminated quoted field")
	}
	return contentEnd, arrow, directiveColon, nil
}

func parseDecTestDirectiveInt(directive, value string) (int, error) {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid %s value %q", directive, value)
	}
	return parsed, nil
}

func normalizeDecTestRounding(value string) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "05up", "ceiling", "down", "floor", "half_down", "half_even", "half_up", "up":
		return value, nil
	default:
		return "", fmt.Errorf("invalid rounding value %q", value)
	}
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
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "--") {
			continue
		}

		contentEnd, arrow, directiveColon, err := scanDecTestLine(line)
		if err != nil {
			return nil, fmt.Errorf("parse dectest %q line %d: %w", path, lineNumber, err)
		}
		if directiveColon >= 0 && arrow >= 0 && directiveColon < arrow {
			return nil, fmt.Errorf("parse dectest %q line %d: directive must not contain a -> separator", path, lineNumber)
		}
		if arrow >= 0 {
			left := splitFields(line[:arrow])
			right := splitFields(line[arrow+2 : contentEnd])
			if len(left) < 3 || len(right) < 1 {
				return nil, fmt.Errorf("parse dectest %q line %d: case requires an id, operation, at least one operand, and result", path, lineNumber)
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
			continue
		}

		content := strings.TrimSpace(line[:contentEnd])
		if content == "" {
			continue
		}
		if directiveColon < 0 {
			return nil, fmt.Errorf("parse dectest %q line %d: unexpected content %q", path, lineNumber, content)
		}

		directive := strings.ToLower(strings.TrimSpace(content[:directiveColon]))
		value := strings.TrimSpace(content[directiveColon+1:])
		switch directive {
		case "rounding":
			rounding, err = normalizeDecTestRounding(value)
		case "precision":
			precision, err = parseDecTestDirectiveInt(directive, value)
			if err == nil && precision <= 0 {
				err = fmt.Errorf("precision must be positive, got %d", precision)
			}
		case "maxexponent":
			maxExponent, err = parseDecTestDirectiveInt(directive, value)
		case "minexponent":
			minExponent, err = parseDecTestDirectiveInt(directive, value)
		case "clamp":
			clamp, err = parseDecTestDirectiveInt(directive, value)
			if err == nil && clamp != 0 && clamp != 1 {
				err = fmt.Errorf("clamp must be 0 or 1, got %d", clamp)
			}
		case "extended":
			var extended int
			extended, err = parseDecTestDirectiveInt(directive, value)
			if err == nil && extended != 0 && extended != 1 {
				err = fmt.Errorf("extended must be 0 or 1, got %d", extended)
			}
		case "version", "dectest":
			if value == "" {
				err = fmt.Errorf("%s value must not be empty", directive)
			}
		default:
			err = fmt.Errorf("unknown directive %q", directive)
		}
		if err != nil {
			return nil, fmt.Errorf("parse dectest %q line %d: %w", path, lineNumber, err)
		}
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
