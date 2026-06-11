package testgen

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type parsedReadtestCase struct {
	Function string
	Rounding int
	Operands []string
	Result   string
	Status   string
	Line     int
}

type readtestFunctionSpec struct {
	Name    string
	Inputs  []string
	Output  string
	Compare string
}

func expandReadTestGroup(group ReadTestGroupSpec) []ReadTestSpec {
	reads := make([]ReadTestSpec, 0, len(group.Cases))
	for _, tc := range group.Cases {
		reads = append(reads, ReadTestSpec{
			Name:          tc.Name,
			Group:         group.Name,
			Format:        group.Format,
			Header:        group.Header,
			Source:        group.Source,
			Function:      tc.Function,
			Kind:          tc.Kind,
			Statuses:      append([]string(nil), group.Statuses...),
			RoundingModes: append([]int(nil), group.RoundingModes...),
			Limit:         tc.Limit,
		})
	}
	return reads
}

func expandReadTestProfile(repoRoot string, profile ReadTestProfileSpec) ([]ReadTestSpec, error) {
	specs, err := parseReadtestFunctionSpecs(filepath.Join(repoRoot, profile.Header))
	if err != nil {
		return nil, err
	}

	allowedFormats := make(map[string]struct{}, len(profile.Formats))
	for _, format := range profile.Formats {
		allowedFormats[strings.ToLower(strings.TrimSpace(format))] = struct{}{}
	}

	reads := make([]ReadTestSpec, 0, len(specs))
	for _, fn := range specs {
		read, ok := buildProfileReadTest(profile, fn, allowedFormats)
		if !ok {
			continue
		}
		reads = append(reads, read)
	}
	if len(reads) == 0 {
		return nil, fmt.Errorf("readtest profile %q matched no functions", profile.Name)
	}
	return reads, nil
}

func buildReadtestProfileAudit(repoRoot string, profile ReadTestProfileSpec) (GeneratedReadtestProfileAudit, error) {
	specs, err := parseReadtestFunctionSpecs(filepath.Join(repoRoot, profile.Header))
	if err != nil {
		return GeneratedReadtestProfileAudit{}, err
	}

	allowedFormats := make(map[string]struct{}, len(profile.Formats))
	for _, format := range profile.Formats {
		allowedFormats[strings.ToLower(strings.TrimSpace(format))] = struct{}{}
	}

	audit := GeneratedReadtestProfileAudit{
		Profile:        profile.Name,
		Header:         profile.Header,
		Source:         profile.Source,
		Selection:      profile.Selection,
		TotalFunctions: len(specs),
		Functions:      make([]GeneratedReadtestFunctionAudit, 0, len(specs)),
	}
	for _, fn := range specs {
		functionAudit := buildReadtestFunctionAudit(profile, fn, allowedFormats)
		if functionAudit.Selected {
			audit.SelectedFunctions++
		} else {
			audit.ExcludedFunctions++
		}
		audit.Functions = append(audit.Functions, functionAudit)
	}
	return audit, nil
}

func buildReadtestFunctionAudit(profile ReadTestProfileSpec, fn readtestFunctionSpec, allowedFormats map[string]struct{}) GeneratedReadtestFunctionAudit {
	audit := GeneratedReadtestFunctionAudit{
		Function:     fn.Name,
		OutputType:   fn.Output,
		InputTypes:   append([]string(nil), fn.Inputs...),
		CompareGroup: fn.Compare,
	}

	if read, ok := buildProfileReadTest(profile, fn, allowedFormats); ok {
		audit.Selected = true
		audit.Format = read.Format
		audit.Kind = read.Kind
		audit.Group = read.Group
		audit.Reason = "selected by readtest profile"
		audit.Classification = "selected"
		return audit
	}

	audit.Reason, audit.Classification = readtestProfileExclusion(profile, fn, allowedFormats)
	return audit
}

func readtestProfileExclusion(profile ReadTestProfileSpec, fn readtestFunctionSpec, allowedFormats map[string]struct{}) (string, string) {
	if profile.Selection != "repo_supported_surface" {
		return "readtest profile selection is not supported by this generator", "generator_profile_unsupported"
	}
	if fn.Compare == "CMP_RELATIVEERR" {
		return "CMP_RELATIVEERR is excluded from the current readtest regular verification profile", "optional_not_required"
	}
	if isHistoricalReadtestSkipFunction(fn.Name) {
		return historicalReadtestSkipReason(fn.Name)
	}
	if isCurrentSpecReadtestExcludedFunction(fn.Name) {
		return "function is excluded by the current spec-phase readtest exclusion list", "current_spec_excluded"
	}
	format, _, ok := classifySupportedReadtestSurface(fn)
	if !ok {
		return unsupportedReadtestSurfaceReason(fn)
	}
	if _, ok := allowedFormats[format]; !ok && format != "status" {
		return "function format is outside the readtest profile's configured formats", "out_of_scope_not_required"
	}
	return "function was not selected by the readtest profile for an unclassified reason", "blocked_required_review"
}

func historicalReadtestSkipReason(name string) (string, string) {
	switch {
	case strings.Contains(name, "binary80"):
		return "binary80 interchange is outside the current supported BID decimal scope", "out_of_scope_not_required"
	case strings.HasPrefix(name, "binary32_to_") || strings.HasPrefix(name, "binary64_to_") || strings.HasPrefix(name, "binary128_to_"):
		return "reverse binary-to-BID conversion is outside the current supported surface", "out_of_scope_not_required"
	case strings.Contains(name, "dpd"):
		return "DPD interchange is outside the BID-only project scope", "out_of_scope_not_required"
	case strings.HasPrefix(name, "bid_fe"):
		return "floating-point environment helper API is outside the public Go mechanical-port verification surface", "out_of_scope_not_required"
	case name == "bid_is754" || name == "bid_is754R":
		return "non-IEEE marker helper is outside the supported BID decimal operation surface", "out_of_scope_not_required"
	case isMixedWidthIntelReadtestExtension(name):
		return "mixed-width Intel extension is not part of the current mandatory BID fixed-width surface", "optional_scope_gap"
	default:
		return "historical explicit readtest skip", "blocked_required_review"
	}
}

func unsupportedReadtestSurfaceReason(fn readtestFunctionSpec) (string, string) {
	switch {
	case fn.Name == "str64":
		return "Intel readtest identity helper is not a public BID operation surface", "out_of_scope_not_required"
	case strings.HasPrefix(fn.Name, "bid_strtod") || strings.HasPrefix(fn.Name, "bid_wcstod"):
		return "C strtod/wcstod compatibility helpers are outside the public Go mechanical-port string API; bid*_from_string is the generated string conversion path", "out_of_scope_not_required"
	case isMixedWidthIntelReadtestExtension(fn.Name):
		return "mixed-width Intel extension is not part of the current mandatory BID fixed-width surface", "optional_scope_gap"
	case strings.Contains(fn.Name, "binary80"):
		return "binary80 interchange is outside the current supported BID decimal scope", "out_of_scope_not_required"
	case strings.Contains(fn.Name, "dpd"):
		return "DPD interchange is outside the BID-only project scope", "out_of_scope_not_required"
	case strings.HasPrefix(fn.Name, "binary"):
		return "binary-to-BID conversion is outside the current supported surface", "out_of_scope_not_required"
	case strings.HasPrefix(fn.Name, "bid_fe"):
		return "floating-point environment helper API is outside the public Go mechanical-port verification surface", "out_of_scope_not_required"
	default:
		return "function signature is outside the current generated readtest adapter surface", "blocked_required_review"
	}
}

func isMixedWidthIntelReadtestExtension(name string) bool {
	return strings.HasPrefix(name, "bid64dq") ||
		strings.HasPrefix(name, "bid64qd") ||
		strings.HasPrefix(name, "bid64qq") ||
		strings.HasPrefix(name, "bid64ddq") ||
		strings.HasPrefix(name, "bid64dqd") ||
		strings.HasPrefix(name, "bid64dqq") ||
		strings.HasPrefix(name, "bid64qdd") ||
		strings.HasPrefix(name, "bid64qdq") ||
		strings.HasPrefix(name, "bid64q_") ||
		strings.HasPrefix(name, "bid128dq") ||
		strings.HasPrefix(name, "bid128qd") ||
		strings.HasPrefix(name, "bid128dd") ||
		strings.HasPrefix(name, "bid128dqd") ||
		strings.HasPrefix(name, "bid128dqq") ||
		strings.HasPrefix(name, "bid128qdd") ||
		strings.HasPrefix(name, "bid128qdq") ||
		strings.HasPrefix(name, "bid128qqd") ||
		strings.HasPrefix(name, "bid128d_")
}

func buildProfileReadTest(profile ReadTestProfileSpec, fn readtestFunctionSpec, allowedFormats map[string]struct{}) (ReadTestSpec, bool) {
	if profile.Selection != "repo_supported_surface" {
		return ReadTestSpec{}, false
	}
	if fn.Compare == "CMP_RELATIVEERR" {
		return ReadTestSpec{}, false
	}
	if isHistoricalReadtestSkipFunction(fn.Name) {
		return ReadTestSpec{}, false
	}
	if isCurrentSpecReadtestExcludedFunction(fn.Name) {
		return ReadTestSpec{}, false
	}
	format, kind, ok := classifySupportedReadtestSurface(fn)
	if !ok {
		return ReadTestSpec{}, false
	}
	if _, ok := allowedFormats[format]; !ok && format != "status" {
		return ReadTestSpec{}, false
	}
	caseFormat := format
	if outputFormat, ok := readtestFormatFromToken(fn.Output); ok {
		caseFormat = outputFormat
	}
	group := caseFormat + "_" + readtestGroupSuffix(kind)
	if kind == "status_control" {
		group = "status_control_operations"
	}
	return ReadTestSpec{
		Name:          fn.Name,
		Group:         group,
		Format:        caseFormat,
		Header:        profile.Header,
		Source:        profile.Source,
		Function:      fn.Name,
		Kind:          kind,
		OutputType:    fn.Output,
		InputTypes:    append([]string(nil), fn.Inputs...),
		CompareGroup:  fn.Compare,
		Statuses:      append([]string(nil), profile.Statuses...),
		RoundingModes: append([]int(nil), profile.RoundingModes...),
	}, true
}

func readtestGroupSuffix(kind string) string {
	switch kind {
	case "from_string", "to_string":
		return "strings"
	case "status_control":
		return "status_control"
	default:
		return "operations"
	}
}

func classifySupportedReadtestSurface(fn readtestFunctionSpec) (format string, kind string, ok bool) {
	if isFlagSubsetReadtestFunction(fn.Name) || isDecimalRoundingDirectionReadtestFunction(fn.Name) {
		if fn.Output != "OP_BID_UINT32" || fn.Compare != "CMP_FUZZYSTATUS" {
			return "", "", false
		}
		for _, input := range fn.Inputs {
			if input != "OP_BID_UINT32" {
				return "", "", false
			}
		}
		switch fn.Name {
		case "bid_getDecimalRoundingDirection", "bid_setDecimalRoundingDirection":
			if len(fn.Inputs) != 1 {
				return "", "", false
			}
		case "bid_restoreFlags":
			if len(fn.Inputs) != 3 {
				return "", "", false
			}
		default:
			if len(fn.Inputs) != 2 {
				return "", "", false
			}
		}
		return "status", "status_control", true
	}

	format, opToken, ok := readtestFormatFromFunctionName(fn.Name)
	if !ok {
		return "", "", false
	}
	if fn.Output != opToken && !isSupportedReadtestScalarOutput(fn.Output) && !isSupportedReadtestDecimalOutput(fn.Output) {
		switch fn.Name {
		case "bid32_from_string", "bid64_from_string", "bid128_from_string", "bid32_to_string", "bid64_to_string", "bid128_to_string":
		default:
			return "", "", false
		}
	}

	switch fn.Name {
	case "bid32_from_string", "bid64_from_string", "bid128_from_string":
		return format, "from_string", true
	case "bid32_to_string", "bid64_to_string", "bid128_to_string":
		return format, "to_string", true
	}

	switch {
	case len(fn.Inputs) == 1 && fn.Output == opToken && fn.Inputs[0] == opToken:
		return format, "unary_op", true
	case len(fn.Inputs) == 1 && fn.Inputs[0] == opToken && isSupportedReadtestScalarOutput(fn.Output):
		return format, "unary_op", true
	case len(fn.Inputs) == 1 && isSupportedReadtestDecimalOutput(fn.Output) && isSupportedReadtestInput(fn.Inputs[0]):
		return format, "unary_op", true
	case len(fn.Inputs) == 2 && fn.Output == opToken && fn.Inputs[0] == opToken && fn.Inputs[1] == opToken:
		return format, "binary_op", true
	case len(fn.Inputs) == 2 && fn.Inputs[0] == opToken && fn.Inputs[1] == opToken && isSupportedReadtestScalarOutput(fn.Output):
		return format, "binary_op", true
	case len(fn.Inputs) == 2 && isSupportedReadtestDecimalOutput(fn.Output) && isSupportedReadtestInput(fn.Inputs[0]) && isSupportedReadtestInput(fn.Inputs[1]):
		return format, "binary_op", true
	case len(fn.Inputs) == 3 && fn.Inputs[0] == opToken && fn.Inputs[1] == opToken && fn.Inputs[2] == opToken && fn.Output == opToken:
		return format, "ternary_op", true
	default:
		return "", "", false
	}
}

func isFlagSubsetReadtestFunction(name string) bool {
	switch name {
	case "bid_testFlags",
		"bid_lowerFlags",
		"bid_signalException",
		"bid_saveFlags",
		"bid_restoreFlags",
		"bid_testSavedFlags":
		return true
	default:
		return false
	}
}

func isDecimalRoundingDirectionReadtestFunction(name string) bool {
	switch name {
	case "bid_getDecimalRoundingDirection", "bid_setDecimalRoundingDirection":
		return true
	default:
		return false
	}
}

func isSupportedReadtestScalarOutput(output string) bool {
	switch output {
	case "OP_BIN32",
		"OP_BIN64",
		"OP_BIN128",
		"OP_INT8",
		"OP_INT16",
		"OP_INT32",
		"OP_INT64",
		"OP_LINT",
		"OP_BID_UINT8",
		"OP_BID_UINT16",
		"OP_BID_UINT32",
		"OP_BID_UINT64":
		return true
	default:
		return false
	}
}

func isSupportedReadtestDecimalOutput(output string) bool {
	return output == "OP_DEC32" || output == "OP_DEC64" || output == "OP_DEC128"
}

func isSupportedReadtestInput(input string) bool {
	return isSupportedReadtestDecimalOutput(input) || isSupportedReadtestScalarOutput(input)
}

func readtestFormatFromFunctionName(name string) (format string, opToken string, ok bool) {
	switch {
	case strings.HasPrefix(name, "bid32_"):
		return "decimal32", "OP_DEC32", true
	case strings.HasPrefix(name, "bid64_"):
		return "decimal64", "OP_DEC64", true
	case strings.HasPrefix(name, "bid128_"):
		return "decimal128", "OP_DEC128", true
	default:
		return "", "", false
	}
}

func readtestFormatFromToken(token string) (format string, ok bool) {
	switch token {
	case "OP_DEC32":
		return "decimal32", true
	case "OP_DEC64":
		return "decimal64", true
	case "OP_DEC128":
		return "decimal128", true
	default:
		return "", false
	}
}

func isHistoricalReadtestSkipFunction(name string) bool {
	switch name {
	case "bid32_to_binary80",
		"bid64_to_binary80",
		"bid128_to_binary80",
		"binary32_to_bid32", "binary32_to_bid64", "binary32_to_bid128",
		"binary64_to_bid32", "binary64_to_bid64", "binary64_to_bid128",
		"binary80_to_bid32", "binary80_to_bid64", "binary80_to_bid128",
		"binary128_to_bid32", "binary128_to_bid64", "binary128_to_bid128",
		"bid_to_dpd32", "bid_to_dpd64", "bid_to_dpd128",
		"bid_dpd_to_bid32", "bid_dpd_to_bid64", "bid_dpd_to_bid128",
		"bid_feclearexcept", "bid_fegetexceptflag", "bid_feraiseexcept", "bid_fesetexceptflag", "bid_fetestexcept",
		"bid_is754", "bid_is754R",
		"bid64ddq_fma", "bid64dqd_fma", "bid64dq_add", "bid64dq_sub", "bid64qd_add", "bid64qd_sub",
		"bid64qq_add", "bid64qq_sub", "bid64qq_mul", "bid64qq_div", "bid64qq_fma", "bid64qqq_fma":
		return true
	default:
		return false
	}
}

func isCurrentSpecReadtestExcludedFunction(name string) bool {
	switch name {
	default:
		return false
	}
}

func parseReadtestFunctionSpecs(path string) ([]readtestFunctionSpec, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open readtest header %q: %w", path, err)
	}
	defer file.Close()

	var specs []readtestFunctionSpec
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)

	reFuncName := regexp.MustCompile(`strcmp\s*\(\s*func\s*,\s*"([^"]+)"\s*\)`)
	reGetTest := regexp.MustCompile(`GETTEST(\d*)\s*\(([^)]+)\)`)
	reCheck := regexp.MustCompile(`check_results\s*\(\s*(\w+)\s*\)`)

	var current readtestFunctionSpec
	flush := func() {
		if current.Name != "" && current.Output != "" {
			specs = append(specs, current)
		}
		current = readtestFunctionSpec{}
	}

	for scanner.Scan() {
		line := scanner.Text()
		if m := reFuncName.FindStringSubmatch(line); m != nil {
			flush()
			current.Name = m[1]
		}
		if m := reGetTest.FindStringSubmatch(line); m != nil {
			parts := strings.Split(m[2], ",")
			for i := range parts {
				parts[i] = strings.TrimSpace(parts[i])
			}
			if len(parts) > 0 {
				current.Output = parts[0]
			}
			if len(parts) > 1 {
				current.Inputs = append([]string(nil), parts[1:]...)
			}
		}
		if m := reCheck.FindStringSubmatch(line); m != nil {
			current.Compare = m[1]
		}
	}
	flush()
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan readtest header %q: %w", path, err)
	}
	return specs, nil
}

func appendGeneratedReadCases(repoRoot string, spec *SharedSpec, read ReadTestSpec) error {
	headerPath := filepath.Join(repoRoot, read.Header)
	if err := verifyReadtestFunction(headerPath, read.Function); err != nil {
		return err
	}

	cases, err := parseReadtestSubset(filepath.Join(repoRoot, read.Source), read)
	if err != nil {
		return err
	}
	for _, tc := range cases {
		spec.ReadCases = append(spec.ReadCases, GeneratedReadCase{
			Suite:                   read.Name,
			Group:                   read.Group,
			Format:                  read.Format,
			Header:                  filepath.ToSlash(read.Header),
			Source:                  filepath.ToSlash(read.Source),
			ID:                      fmt.Sprintf("%s_%03d", read.Name, len(spec.ReadCases)+1),
			Line:                    tc.Line,
			Function:                tc.Function,
			Kind:                    read.Kind,
			OutputType:              read.OutputType,
			InputTypes:              append([]string(nil), read.InputTypes...),
			CompareGroup:            read.CompareGroup,
			NativeCompareSkipReason: read.NativeCompareSkipReason,
			Operands:                append([]string(nil), tc.Operands...),
			Expected:                tc.Result,
			Status:                  tc.Status,
			Rounding:                tc.Rounding,
		})
	}
	if len(cases) == 0 {
		return fmt.Errorf("readtest %q matched no cases", read.Name)
	}
	return nil
}

var (
	readtestStatusPattern    = regexp.MustCompile(`(?i)^(?:0x)?[0-9a-f]+$`)
	readtestBitsPattern      = regexp.MustCompile(`(?i)^\[[0-9a-fA-F,]+\]$`)
	supportedReadtestLiteral = regexp.MustCompile(`(?i)^[+-]?(?:(?:\d+(?:\.\d*)?|\.\d+)(?:[eE][+-]?\d+)?|inf(?:inity)?|nan|snan)$`)
)

func verifyReadtestFunction(path string, function string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read readtest header %q: %w", path, err)
	}

	needle := regexp.MustCompile(fmt.Sprintf(`strcmp\s*\(\s*func\s*,\s*%q\s*\)`, function))
	if !needle.Match(data) {
		return fmt.Errorf("readtest function %q not found in %q", function, path)
	}
	return nil
}

func parseReadtestSubset(path string, spec ReadTestSpec) ([]parsedReadtestCase, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open readtest %q: %w", path, err)
	}
	defer file.Close()

	statuses := make(map[string]struct{}, len(spec.Statuses))
	for _, status := range spec.Statuses {
		statuses[strings.ToUpper(status)] = struct{}{}
	}

	roundings := make(map[int]struct{}, len(spec.RoundingModes))
	for _, rounding := range spec.RoundingModes {
		roundings[rounding] = struct{}{}
	}

	var cases []parsedReadtestCase
	scanner := bufio.NewScanner(file)
	for lineNo := 1; scanner.Scan(); lineNo++ {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "--") {
			continue
		}
		if strings.Contains(strings.ToLower(line), "longintsize=32") {
			continue
		}

		fields := splitFields(line)
		if len(fields) < 5 || fields[0] != spec.Function {
			continue
		}

		rounding, err := strconv.Atoi(fields[1])
		if err != nil {
			return nil, fmt.Errorf("parse readtest rounding %q at %s:%d: %w", fields[1], path, lineNo, err)
		}
		if len(roundings) > 0 {
			if _, ok := roundings[rounding]; !ok {
				continue
			}
		}

		if len(fields) < 4 {
			continue
		}
		statusIndex := -1
		for i := len(fields) - 1; i >= 3; i-- {
			if readtestStatusPattern.MatchString(strings.ToUpper(fields[i])) {
				statusIndex = i
				break
			}
		}
		if statusIndex < 3 {
			return nil, fmt.Errorf("parse readtest status at %s:%d", path, lineNo)
		}
		operands := append([]string(nil), fields[2:statusIndex-1]...)
		result := fields[statusIndex-1]
		status := strings.ToUpper(fields[statusIndex])
		if len(statuses) > 0 {
			if _, ok := statuses[status]; !ok {
				continue
			}
		}
		result = repairKnownReadtestResult(spec.Function, lineNo, operands, result)
		if !supportsReadtestCase(spec, operands, result) {
			continue
		}

		cases = append(cases, parsedReadtestCase{
			Function: fields[0],
			Rounding: rounding,
			Operands: append([]string(nil), operands...),
			Result:   result,
			Status:   status,
			Line:     lineNo,
		})
		if spec.Limit > 0 && len(cases) == spec.Limit {
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan readtest %q: %w", path, err)
	}

	return cases, nil
}

func repairKnownReadtestResult(function string, line int, operands []string, result string) string {
	// Intel's published readtest.in (including the official netlib tarball) has
	// one truncated decimal128 result literal at bid128_fdim line 93914.
	// The surrounding rounding variants and the native implementation agree the
	// intended exact bits are 3040...0000, so normalize that one upstream defect
	// before the generated harness reasons about the case.
	if function == "bid128_fdim" &&
		line == 93914 &&
		len(operands) == 2 &&
		strings.EqualFold(strings.TrimSpace(operands[0]), "[DFFFED09BEAD87C0378D8E63FFFFFFFF]") &&
		strings.EqualFold(strings.TrimSpace(operands[1]), "[5FFFED09BEAD87C0378D8E63FFFFFFFF]") &&
		strings.EqualFold(strings.TrimSpace(result), "[304000000000000000000000000000]") {
		return "[30400000000000000000000000000000]"
	}
	return result
}

func supportsReadtestCase(spec ReadTestSpec, operands []string, expected string) bool {
	if len(spec.InputTypes) > 0 {
		if len(operands) != len(spec.InputTypes) {
			return false
		}
		for i, inputType := range spec.InputTypes {
			if !supportsReadtestValue(inputType, operands[i], false) {
				return false
			}
		}
		return supportsReadtestValue(spec.OutputType, expected, true)
	}

	switch spec.Kind {
	case "from_string":
		return len(operands) == 1 && supportedReadtestLiteral.MatchString(operands[0]) && readtestBitsPattern.MatchString(expected)
	case "to_string":
		if len(operands) != 1 || !readtestBitsPattern.MatchString(operands[0]) {
			return false
		}
		return !strings.Contains(strings.ToLower(expected), "snan")
	case "binary_op", "unary_op":
		if !readtestBitsPattern.MatchString(expected) {
			return false
		}
		for _, operand := range operands {
			if readtestBitsPattern.MatchString(operand) {
				continue
			}
			if !supportedReadtestLiteral.MatchString(operand) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func supportsReadtestValue(kind, value string, isResult bool) bool {
	value = strings.TrimSpace(value)
	switch kind {
	case "OP_DEC32", "OP_DEC64":
		if strings.Contains(value, ",") {
			return false
		}
		if readtestBitsPattern.MatchString(value) {
			return true
		}
		return supportedReadtestLiteral.MatchString(value)
	case "OP_DEC128":
		if readtestBitsPattern.MatchString(value) {
			return true
		}
		return supportedReadtestLiteral.MatchString(value)
	case "OP_STRING":
		return value != "" && (!isResult || !strings.Contains(strings.ToLower(value), "snan"))
	case "OP_BIN32", "OP_BIN64", "OP_BIN128":
		return readtestBitsPattern.MatchString(value)
	case "OP_INT8", "OP_INT16", "OP_INT32", "OP_INT64", "OP_LINT",
		"OP_BID_UINT8", "OP_BID_UINT16", "OP_BID_UINT32", "OP_BID_UINT64":
		if _, err := strconv.ParseInt(value, 10, 64); err == nil {
			return true
		}
		if _, err := strconv.ParseUint(value, 10, 64); err == nil {
			return true
		}
		return false
	default:
		return false
	}
}
