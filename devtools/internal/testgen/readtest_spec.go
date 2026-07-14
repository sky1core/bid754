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

func buildReadtestProfileInventory(repoRoot string, profile ReadTestProfileSpec) (GeneratedReadtestProfileInventory, error) {
	specs, err := parseReadtestFunctionSpecs(filepath.Join(repoRoot, profile.Header))
	if err != nil {
		return GeneratedReadtestProfileInventory{}, err
	}

	allowedFormats := make(map[string]struct{}, len(profile.Formats))
	for _, format := range profile.Formats {
		allowedFormats[strings.ToLower(strings.TrimSpace(format))] = struct{}{}
	}

	inventory := GeneratedReadtestProfileInventory{
		Profile:        profile.Name,
		Header:         profile.Header,
		Source:         profile.Source,
		Selection:      profile.Selection,
		TotalFunctions: len(specs),
		Functions:      make([]GeneratedReadtestFunctionInventory, 0, len(specs)),
	}
	for _, fn := range specs {
		functionInventory := buildReadtestFunctionInventory(profile, fn, allowedFormats)
		if functionInventory.Selected {
			inventory.SelectedFunctions++
		} else {
			inventory.ExcludedFunctions++
		}
		inventory.Functions = append(inventory.Functions, functionInventory)
	}
	return inventory, nil
}

func buildReadtestFunctionInventory(profile ReadTestProfileSpec, fn readtestFunctionSpec, allowedFormats map[string]struct{}) GeneratedReadtestFunctionInventory {
	inventory := GeneratedReadtestFunctionInventory{
		Function:     fn.Name,
		OutputType:   fn.Output,
		InputTypes:   append([]string(nil), fn.Inputs...),
		CompareGroup: fn.Compare,
	}

	if read, ok := buildProfileReadTest(profile, fn, allowedFormats); ok {
		inventory.Selected = true
		inventory.Format = read.Format
		inventory.Kind = read.Kind
		inventory.Group = read.Group
		inventory.Reason = "selected by readtest profile"
		inventory.Classification = "selected"
		return inventory
	}

	inventory.Reason, inventory.Classification = readtestProfileExclusion(profile, fn, allowedFormats)
	return inventory
}

func readtestProfileExclusion(profile ReadTestProfileSpec, fn readtestFunctionSpec, allowedFormats map[string]struct{}) (string, string) {
	if profile.Selection != "repo_supported_surface" {
		return "readtest profile selection is not supported by this generator", "generator_profile_unsupported"
	}
	if fn.Compare == "CMP_RELATIVEERR" {
		return "CMP_RELATIVEERR is excluded from the current regular readtest verification profile", "optional_not_required"
	}
	if isHistoricalReadtestSkipFunction(fn.Name) {
		return historicalReadtestSkipReason(fn.Name)
	}
	if isCurrentSpecReadtestExcludedFunction(fn.Name) {
		return "function is excluded by the current declared readtest exclusion list", "current_spec_excluded"
	}
	format, _, ok := classifySupportedReadtestSurface(fn)
	if !ok {
		return unsupportedReadtestSurfaceReason(fn)
	}
	if _, ok := allowedFormats[format]; !ok && format != "status" {
		return "function format is outside the readtest profile's configured formats", "out_of_scope_not_required"
	}
	return "function was not selected by the readtest profile for an unclassified reason", "unresolved_required_review"
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
		return "IEEE 754-2019 5.7.1 version-conformance predicate outside the supported BID decimal operation surface", "out_of_scope_not_required"
	case isMixedWidthIntelReadtestExtension(name):
		return "mixed-width Intel extension is not part of the current mandatory BID fixed-width surface", "optional_scope_gap"
	default:
		return "historical explicit readtest skip", "unresolved_required_review"
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
		return "function signature is outside the current generated readtest adapter surface", "unresolved_required_review"
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

// isTier1MixedWidthIntelReadtestFunction is the closed Intel BID C surface
// selected by the project's Tier 1 arithmetic contract.  Keep this list
// explicit: Decimal32-destination combinations do not exist upstream, while
// the remaining mixed FMA/sqrt functions are Tier 2 and stay excluded.
func isTier1MixedWidthIntelReadtestFunction(name string) bool {
	switch name {
	case "bid64dq_add", "bid64qd_add", "bid64qq_add",
		"bid64dq_sub", "bid64qd_sub", "bid64qq_sub",
		"bid64dq_mul", "bid64qd_mul", "bid64qq_mul",
		"bid64dq_div", "bid64qd_div", "bid64qq_div",
		"bid128dd_add", "bid128dq_add", "bid128qd_add",
		"bid128dd_sub", "bid128dq_sub", "bid128qd_sub",
		"bid128dd_mul", "bid128dq_mul", "bid128qd_mul",
		"bid128dd_div", "bid128dq_div", "bid128qd_div":
		return true
	default:
		return false
	}
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
	case isTier1MixedWidthIntelReadtestFunction(name) && strings.HasPrefix(name, "bid64"):
		return "decimal64", "OP_DEC64", true
	case isTier1MixedWidthIntelReadtestFunction(name) && strings.HasPrefix(name, "bid128"):
		return "decimal128", "OP_DEC128", true
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
		"bid64ddq_fma", "bid64dqd_fma", "bid64qq_fma", "bid64qqq_fma":
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

// appendGeneratedReadCases appends the generated cases for one readtest and
// returns the accepted case count and the per-reason row-skip counts for its
// function (see parseReadtestSubset).
func appendGeneratedReadCases(repoRoot string, spec *SharedSpec, read ReadTestSpec) (int, map[string]int, error) {
	headerPath := filepath.Join(repoRoot, read.Header)
	if err := verifyReadtestFunction(headerPath, read.Function); err != nil {
		return 0, nil, err
	}

	cases, rowSkips, err := parseReadtestSubset(filepath.Join(repoRoot, read.Source), read)
	if err != nil {
		return 0, nil, err
	}
	for _, tc := range cases {
		spec.ReadCases = append(spec.ReadCases, GeneratedReadCase{
			Suite:  read.Name,
			Group:  read.Group,
			Format: read.Format,
			Header: filepath.ToSlash(read.Header),
			Source: filepath.ToSlash(read.Source),
			// Keep IDs tied to the pinned source row, not global append order,
			// so adding an unrelated suite does not renumber existing cases.
			ID:                      fmt.Sprintf("%s_line_%d", read.Name, tc.Line),
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
		return 0, nil, fmt.Errorf("readtest %q matched no cases", read.Name)
	}
	return len(cases), rowSkips, nil
}

var (
	readtestStatusPattern = regexp.MustCompile(`(?i)^(?:0x)?[0-9a-f]+$`)
	readtestBitsPattern   = regexp.MustCompile(`(?i)^\[[0-9a-fA-F,]+\]$`)
	// NaN spellings mirror the tokens Intel readtest.c getop32/64/128 hand to
	// bid*_from_string: plain "NaN", "sNaN", and "qNaN" (case-insensitive). The
	// "q" prefix was previously rejected, dropping every qNaN-operand row.
	supportedReadtestLiteral = regexp.MustCompile(`(?i)^[+-]?(?:(?:\d+(?:\.\d*)?|\.\d+)(?:[eE][+-]?\d+)?|inf(?:inity)?|[qs]?nan)$`)
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

// readtest row-skip reason keys recorded for rows that belong to the target
// function but are not turned into generated cases. They are aggregated per
// profile in the generated inventory (and anchored) so a parser change that
// silently drops rows moves a named reason count, not only the total.
const (
	readtestRowSkipLongIntSize32     = "longintsize32"
	readtestRowSkipShortRow          = "short_row"
	readtestRowSkipRoundingFiltered  = "rounding_filtered"
	readtestRowSkipStatusFiltered    = "status_filtered"
	readtestRowSkipUnsupportedResult = "unsupported_literal"
)

// parseReadtestSubset returns the generated cases for one readtest function and,
// as its second result, per-reason counts of rows that belong to that function
// but were dropped. Rows for other functions are the scan mechanism and are not
// counted; the function-name check runs first precisely so every subsequent
// drop is attributable to spec.Function.
func parseReadtestSubset(path string, spec ReadTestSpec) ([]parsedReadtestCase, map[string]int, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, fmt.Errorf("open readtest %q: %w", path, err)
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

	rowSkips := map[string]int{}
	var cases []parsedReadtestCase
	scanner := bufio.NewScanner(file)
	for lineNo := 1; scanner.Scan(); lineNo++ {
		// Intel readtest.c removes the comment before tokenizing a row:
		//   p = strstr(line, "--"); if (p) *p = 0;   (readtest.c ~1834)
		// so the first "--" substring and everything after it is dropped, then
		// trailing spaces are trimmed. Replicate that exactly. Without it a
		// trailing hex-shaped comment token (e.g.
		//   bid128_llround 0 [..] 1 00 -- 1
		// where the comment "1" matches the status pattern) is picked up by the
		// back-to-front status scan below, which shifts the operand/result/status
		// split and silently drops the row as unsupported. A comment-only line
		// (including a "--"-disabled data row) collapses to empty and is skipped
		// like before.
		line := scanner.Text()
		if idx := strings.Index(line, "--"); idx >= 0 {
			line = line[:idx]
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := splitFields(line)
		if len(fields) == 0 || fields[0] != spec.Function {
			// Row for another function: this is how a per-function scan walks
			// the shared readtest.in, not a drop of this function's data.
			continue
		}

		// From here the row belongs to spec.Function, so every skip below is
		// counted under a named reason.
		if strings.Contains(strings.ToLower(line), "longintsize=32") {
			rowSkips[readtestRowSkipLongIntSize32]++
			continue
		}
		if len(fields) < 5 {
			rowSkips[readtestRowSkipShortRow]++
			continue
		}

		rounding, err := strconv.Atoi(fields[1])
		if err != nil {
			return nil, nil, fmt.Errorf("parse readtest rounding %q at %s:%d: %w", fields[1], path, lineNo, err)
		}
		if len(roundings) > 0 {
			if _, ok := roundings[rounding]; !ok {
				rowSkips[readtestRowSkipRoundingFiltered]++
				continue
			}
		}

		statusIndex := -1
		for i := len(fields) - 1; i >= 3; i-- {
			if readtestStatusPattern.MatchString(strings.ToUpper(fields[i])) {
				statusIndex = i
				break
			}
		}
		if statusIndex < 3 {
			return nil, nil, fmt.Errorf("parse readtest status at %s:%d", path, lineNo)
		}
		operands := append([]string(nil), fields[2:statusIndex-1]...)
		result := fields[statusIndex-1]
		status := strings.ToUpper(fields[statusIndex])
		if len(statuses) > 0 {
			if _, ok := statuses[status]; !ok {
				rowSkips[readtestRowSkipStatusFiltered]++
				continue
			}
		}
		// readtest.c first tokenizes a row by the number of fields and only
		// afterwards lets the selected GETTEST1/2/3 macro consume the operands
		// declared for that function. A handful of upstream unary rows carry a
		// second, deliberately ignored operand; keeping it here would reject a
		// row that the canonical harness executes. Drop only operands beyond the
		// header-derived input arity, exactly as GETTEST does. A short row still
		// fails supportsReadtestCase below.
		if len(spec.InputTypes) > 0 && len(operands) > len(spec.InputTypes) {
			operands = operands[:len(spec.InputTypes)]
		}
		operands = repairKnownReadtestOperands(spec.Function, lineNo, operands)
		result = repairKnownReadtestResult(spec.Function, lineNo, operands, result)
		if !supportsReadtestCase(spec, operands, result) {
			rowSkips[readtestRowSkipUnsupportedResult]++
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
		return nil, nil, fmt.Errorf("scan readtest %q: %w", path, err)
	}

	return cases, rowSkips, nil
}

func repairKnownReadtestOperands(function string, line int, operands []string) []string {
	// Intel's published readtest.in has one bid128qd_div row whose Decimal64
	// operand is missing the closing bracket. The canonical getop64 macro still
	// executes it because sscanf consumes the 16 hex digits after '[' without
	// requiring ']'. Normalize only that pinned upstream defect to the same bits
	// so generated runners exercise the row instead of silently dropping it.
	if function == "bid128qd_div" &&
		line == 11608 &&
		len(operands) == 2 &&
		strings.EqualFold(strings.TrimSpace(operands[0]), "[80000000000000000000000000000001]") &&
		strings.EqualFold(strings.TrimSpace(operands[1]), "[2fe0000000000005") {
		repaired := append([]string(nil), operands...)
		repaired[1] = "[2fe0000000000005]"
		return repaired
	}
	return operands
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
	// The operand of bid*_from_string is the string under test, not a decimal
	// literal that the harness parses before the call. Invalid spellings are
	// authoritative negative parser cases and must reach the implementation.
	if spec.Kind == "from_string" {
		return len(operands) == 1 && operands[0] != "" && supportsReadtestValue(spec.OutputType, expected, true)
	}
	// Intel declares bid*_nan's tag argument as OP_DEC* in readtest.h even
	// though the implementation receives the original C string (including the
	// special NULL token). Keep that verification-only source grammar intact.
	if strings.HasSuffix(spec.Function, "_nan") {
		return len(operands) == 1 && operands[0] != "" && supportsReadtestValue(spec.OutputType, expected, true)
	}
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
	case "OP_DEC32":
		if strings.Contains(value, ",") {
			return false
		}
		if readtestBitsPattern.MatchString(value) {
			return true
		}
		return supportedReadtestLiteral.MatchString(value)
	case "OP_DEC64":
		// Intel getop64 uses sscanf("%016llx"), so a bracketed comma form
		// consumes the first word and ignores the suffix. Pinned nextafter /
		// nexttoward rows exercise this exact upstream behavior.
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
	case "OP_INT8", "OP_INT16", "OP_INT32", "OP_INT64", "OP_LINT":
		return supportsReadtestSignedIntegerToken(value)
	case "OP_BID_UINT8", "OP_BID_UINT16", "OP_BID_UINT32", "OP_BID_UINT64":
		return supportsReadtestUnsignedIntegerToken(value)
	default:
		return false
	}
}

// readtestIntegerTokenPrefix mirrors the lexical prefix accepted by
// readtest.c's getop*i/getop*u scanf conversions. Bracketed values are hex bit
// patterns; decimal conversions accept a leading sign and stop at the first
// non-digit (for example "1.0" is the integer 1). Width conversion is
// performed later by the generated dispatcher, matching the C assignment.
func readtestIntegerTokenPrefix(value string) (string, int, bool) {
	value = strings.TrimSpace(value)
	if strings.HasPrefix(value, "[") {
		end := 1
		for end < len(value) {
			c := value[end]
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
				break
			}
			end++
		}
		if end == 1 {
			return "", 0, false
		}
		return value[1:end], 16, true
	}

	end := 0
	if end < len(value) && (value[end] == '+' || value[end] == '-') {
		end++
	}
	digits := end
	for end < len(value) && value[end] >= '0' && value[end] <= '9' {
		end++
	}
	if end == digits {
		return "", 0, false
	}
	return value[:end], 10, true
}

func supportsReadtestSignedIntegerToken(value string) bool {
	prefix, base, ok := readtestIntegerTokenPrefix(value)
	if !ok {
		return false
	}
	if base == 16 {
		_, err := strconv.ParseUint(prefix, 16, 64)
		return err == nil
	}
	_, err := strconv.ParseInt(prefix, 10, 64)
	return err == nil
}

func supportsReadtestUnsignedIntegerToken(value string) bool {
	prefix, base, ok := readtestIntegerTokenPrefix(value)
	if !ok {
		return false
	}
	if base == 16 {
		_, err := strconv.ParseUint(prefix, 16, 64)
		return err == nil
	}
	prefix = strings.TrimPrefix(prefix, "+")
	prefix = strings.TrimPrefix(prefix, "-")
	_, err := strconv.ParseUint(prefix, 10, 64)
	return err == nil
}
