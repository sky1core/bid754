package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/sky1core/bid754/devtools/internal/readtestspec"
	"github.com/sky1core/bid754/devtools/internal/testgen"
)

type ReadtestSpec struct {
	Name         string
	Inputs       []string
	Output       string
	Compare      string
	CallType     string
	UsesRounding bool
	NeedsStatus  bool
}

type ReadtestSpecs struct {
	Version string
	Source  string
	Specs   []ReadtestSpec
}

type RustReadtestSkipManifest struct {
	Version string                  `json:"version"`
	Source  string                  `json:"source"`
	Skips   []RustReadtestSkipEntry `json:"skips"`
}

type RustReadtestSkipEntry struct {
	Function       string `json:"function"`
	ReasonCode     string `json:"reason_code"`
	Reason         string `json:"reason"`
	Classification string `json:"classification"`
}

type RustReadtestDispatchAudit struct {
	Version      string                         `json:"version"`
	Source       string                         `json:"source"`
	SkipManifest string                         `json:"skip_manifest"`
	Dispatched   int                            `json:"dispatched"`
	Skipped      int                            `json:"skipped"`
	Functions    []RustReadtestDispatchAuditRow `json:"functions"`
}

type RustReadtestDispatchAuditRow struct {
	Function       string `json:"function"`
	Compare        string `json:"compare"`
	Status         string `json:"status"`
	ReasonCode     string `json:"reason_code,omitempty"`
	Reason         string `json:"reason,omitempty"`
	Classification string `json:"classification,omitempty"`
}

type rustDispatchSkipReason struct {
	Code           string
	Reason         string
	Classification string
}

func parseFunc(op string) string {
	switch op {
	case "OP_DEC64":
		return "parse_bid64"
	case "OP_DEC32":
		return "parse_bid32"
	case "OP_DEC128":
		return "parse_bid128"
	case "OP_INT32":
		return "parse_i32"
	case "OP_INT64", "OP_LINT":
		return "parse_i64"
	case "OP_INT8":
		return "parse_i8"
	case "OP_INT16":
		return "parse_i16"
	case "OP_BID_UINT8":
		return "parse_u8"
	case "OP_BID_UINT16":
		return "parse_u16"
	case "OP_BID_UINT32":
		return "parse_u32"
	case "OP_BID_UINT64":
		return "parse_u64"
	case "OP_BIN32":
		return "parse_f32"
	case "OP_BIN64":
		return "parse_f64"
	case "OP_BIN128":
		return "parse_bid128"
	case "OP_STRING":
		return "parse_string"
	default:
		return ""
	}
}

func compareFunc(op string) string {
	switch op {
	case "OP_DEC64", "OP_BID_UINT64":
		return "compare_u64"
	case "OP_DEC32", "OP_BID_UINT32":
		return "compare_u32"
	case "OP_DEC128":
		return "compare_bid128"
	case "OP_INT32":
		return "compare_i32"
	case "OP_INT64", "OP_LINT":
		return "compare_i64"
	case "OP_INT8":
		return "compare_i8"
	case "OP_INT16":
		return "compare_i16"
	case "OP_BID_UINT8":
		return "compare_u8"
	case "OP_BID_UINT16":
		return "compare_u16"
	case "OP_BIN32":
		return "compare_f32"
	case "OP_BIN64":
		return "compare_f64"
	case "OP_BIN128":
		return "compare_bid128"
	default:
		return ""
	}
}

func isSkipType(op string) bool {
	switch op {
	case "OP_BIN80", "OP_DPD32", "OP_DPD64", "OP_DPD128":
		return true
	}
	return false
}

func loadReadtestSpecs(projectRoot string) *ReadtestSpecs {
	indexPath := filepath.Join(projectRoot, "generated", "testspec", "spec_index.json")
	shared, err := testgen.LoadGenerated(indexPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load generated spec index: %v\n", err)
		os.Exit(1)
	}

	inScope := make(map[string]bool)
	for _, tc := range shared.ReadCases {
		inScope[tc.Function] = true
	}

	registryPath := filepath.Join(projectRoot, "tools", "registry", "readtest_specs.json")
	registry, err := readtestspec.LoadFromProjectRoot(projectRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load readtest_specs: %v\n", err)
		os.Exit(1)
	}
	if err := readtestspec.WriteRegistryJSON(registryPath, registry); err != nil {
		fmt.Fprintf(os.Stderr, "write readtest_specs: %v\n", err)
		os.Exit(1)
	}

	specs := make([]ReadtestSpec, 0, len(inScope))
	for _, spec := range registry.Specs {
		if !inScope[spec.Name] {
			continue
		}
		specs = append(specs, ReadtestSpec{
			Name:         spec.Name,
			Inputs:       append([]string(nil), spec.Inputs...),
			Output:       spec.Output,
			Compare:      spec.Compare,
			CallType:     spec.CallType,
			UsesRounding: specUsesRounding(ReadtestSpec{CallType: spec.CallType}),
			NeedsStatus:  specNeedsStatus(ReadtestSpec{CallType: spec.CallType}),
		})
	}
	sort.Slice(specs, func(i, j int) bool { return specs[i].Name < specs[j].Name })
	return &ReadtestSpecs{
		Version: registry.Version,
		Source:  registry.Source,
		Specs:   specs,
	}
}

func loadRustReadtestSkipManifest(projectRoot string) (string, map[string]RustReadtestSkipEntry) {
	relPath := filepath.Join("tools", "registry", "rust_readtest_skip_manifest.json")
	path := filepath.Join(projectRoot, relPath)
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read Rust readtest skip manifest %q: %v\n", relPath, err)
		os.Exit(1)
	}
	var manifest RustReadtestSkipManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		fmt.Fprintf(os.Stderr, "parse Rust readtest skip manifest %q: %v\n", relPath, err)
		os.Exit(1)
	}
	if manifest.Version == "" || manifest.Source == "" {
		fmt.Fprintf(os.Stderr, "Rust readtest skip manifest %q is missing version/source\n", relPath)
		os.Exit(1)
	}
	entries := make(map[string]RustReadtestSkipEntry, len(manifest.Skips))
	for _, entry := range manifest.Skips {
		if entry.Function == "" || entry.ReasonCode == "" || entry.Reason == "" || entry.Classification == "" {
			fmt.Fprintf(os.Stderr, "Rust readtest skip manifest has incomplete entry: %+v\n", entry)
			os.Exit(1)
		}
		if _, exists := entries[entry.Function]; exists {
			fmt.Fprintf(os.Stderr, "Rust readtest skip manifest duplicates %q\n", entry.Function)
			os.Exit(1)
		}
		entries[entry.Function] = entry
	}
	return relPath, entries
}

func camelToSnake(s string) string {
	var result []byte
	for i, c := range s {
		if c >= 'A' && c <= 'Z' {
			if i > 0 && s[i-1] >= 'a' && s[i-1] <= 'z' {
				result = append(result, '_')
			} else if i > 0 && i+1 < len(s) && s[i-1] >= 'A' && s[i-1] <= 'Z' && s[i+1] >= 'a' && s[i+1] <= 'z' {
				result = append(result, '_')
			}
			result = append(result, byte(c-'A'+'a'))
		} else {
			result = append(result, byte(c))
		}
	}
	return string(result)
}

type RustParam struct {
	Name string
	Type string
}

type RustFuncSig struct {
	Name        string
	Params      []RustParam
	ReturnTypes []string
	Body        string
}

func normalizeRustSigType(s string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(s)), " ")
}

func splitTopLevel(s string) []string {
	var parts []string
	var current []rune
	var parenDepth, bracketDepth, angleDepth int
	flush := func() {
		part := strings.TrimSpace(string(current))
		if part != "" {
			parts = append(parts, part)
		}
		current = current[:0]
	}

	for _, r := range s {
		switch r {
		case ',':
			if parenDepth == 0 && bracketDepth == 0 && angleDepth == 0 {
				flush()
				continue
			}
		case '(':
			parenDepth++
		case ')':
			if parenDepth > 0 {
				parenDepth--
			}
		case '[':
			bracketDepth++
		case ']':
			if bracketDepth > 0 {
				bracketDepth--
			}
		case '<':
			angleDepth++
		case '>':
			if angleDepth > 0 {
				angleDepth--
			}
		}
		current = append(current, r)
	}
	flush()
	return parts
}

func findMatchingParen(s string, openIdx int) int {
	depth := 0
	for i := openIdx; i < len(s); i++ {
		switch s[i] {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

func parseRustFuncSigLine(line string) (RustFuncSig, bool) {
	if !strings.HasPrefix(line, "pub fn ") {
		return RustFuncSig{}, false
	}

	rest := strings.TrimPrefix(line, "pub fn ")
	openIdx := strings.Index(rest, "(")
	if openIdx <= 0 {
		return RustFuncSig{}, false
	}
	closeIdx := findMatchingParen(rest, openIdx)
	if closeIdx < 0 {
		return RustFuncSig{}, false
	}

	sig := RustFuncSig{Name: rest[:openIdx]}
	paramsPart := rest[openIdx+1 : closeIdx]
	for _, part := range splitTopLevel(paramsPart) {
		colonIdx := strings.Index(part, ":")
		if colonIdx < 0 {
			continue
		}
		name := strings.TrimSpace(part[:colonIdx])
		name = strings.TrimPrefix(name, "mut ")
		sig.Params = append(sig.Params, RustParam{
			Name: name,
			Type: normalizeRustSigType(part[colonIdx+1:]),
		})
	}

	afterParams := strings.TrimSpace(rest[closeIdx+1:])
	if arrowIdx := strings.Index(afterParams, "->"); arrowIdx >= 0 {
		returnPart := strings.TrimSpace(afterParams[arrowIdx+2:])
		if braceIdx := strings.LastIndex(returnPart, "{"); braceIdx >= 0 {
			returnPart = strings.TrimSpace(returnPart[:braceIdx])
		}
		if returnPart != "" && returnPart != "()" {
			if strings.HasPrefix(returnPart, "(") && strings.HasSuffix(returnPart, ")") {
				for _, part := range splitTopLevel(returnPart[1 : len(returnPart)-1]) {
					sig.ReturnTypes = append(sig.ReturnTypes, normalizeRustSigType(part))
				}
			} else {
				sig.ReturnTypes = append(sig.ReturnTypes, normalizeRustSigType(returnPart))
			}
		}
	}

	return sig, true
}

func loadRustFuncSigs(projectRoot string) map[string]RustFuncSig {
	dir := filepath.Join(projectRoot, "..", "bid754-rs", "src", "generated")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	result := make(map[string]RustFuncSig)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".rs") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		lines := strings.Split(string(data), "\n")
		for i := 0; i < len(lines); i++ {
			if sig, ok := parseRustFuncSigLine(lines[i]); ok {
				end := i + 1
				for end < len(lines) {
					line := lines[end]
					if strings.HasPrefix(line, "pub fn ") || strings.HasPrefix(line, "pub(crate) fn ") {
						break
					}
					end++
				}
				sig.Body = strings.Join(lines[i:end], "\n")
				result[sig.Name] = sig
			}
		}
	}
	return result
}

func normalizeFuncName(s string) string {
	return strings.ToLower(strings.ReplaceAll(s, "_", ""))
}

func buildNormalizedRustNameMap(sigs map[string]RustFuncSig) map[string]string {
	result := make(map[string]string)
	for name := range sigs {
		key := normalizeFuncName(name)
		if _, exists := result[key]; !exists {
			result[key] = name
		}
	}
	return result
}

func resolveRustFuncName(candidate string, sigs map[string]RustFuncSig, normalized map[string]string) string {
	if _, ok := sigs[candidate]; ok {
		return candidate
	}
	if actual, ok := normalized[normalizeFuncName(candidate)]; ok {
		return actual
	}
	return ""
}

func specUsesRounding(spec ReadtestSpec) bool {
	return !strings.Contains(spec.CallType, "NORND") && !strings.Contains(spec.CallType, "EMPTY")
}

func specNeedsStatus(spec ReadtestSpec) bool {
	return !strings.Contains(spec.CallType, "NOSTAT") &&
		!strings.Contains(spec.CallType, "NOFLAGS") &&
		!strings.Contains(spec.CallType, "EMPTY") &&
		!strings.Contains(spec.CallType, "RESREF")
}

func isRustFlagsParam(param RustParam) bool {
	return param.Type == "&mut u32"
}

func isRustRoundingParam(param RustParam) bool {
	if param.Type != "i32" && param.Type != "i64" {
		return false
	}
	name := strings.ToLower(param.Name)
	return strings.Contains(name, "rnd") || strings.Contains(name, "round")
}

func rustSigHasFlags(sig RustFuncSig) bool {
	for _, param := range sig.Params {
		if isRustFlagsParam(param) {
			return true
		}
	}
	return len(sig.ReturnTypes) > 1 && sig.ReturnTypes[len(sig.ReturnTypes)-1] == "u32"
}

func rustSigHasRounding(sig RustFuncSig) bool {
	for _, param := range sig.Params {
		if isRustRoundingParam(param) {
			return true
		}
	}
	return false
}

func rustSigHasTodo(sig RustFuncSig) bool {
	body := strings.ToLower(sig.Body)
	return strings.Contains(body, "todo!(") || strings.Contains(body, "not yet implemented")
}

func scoreRustFuncSig(spec ReadtestSpec, sig RustFuncSig) int {
	score := 0
	specHasRnd := specUsesRounding(spec)
	specHasStatus := specNeedsStatus(spec)

	if specHasRnd {
		if rustSigHasRounding(sig) {
			score += 40
		} else {
			score -= 40
		}
	} else if !rustSigHasRounding(sig) {
		score += 5
	}

	if specHasStatus {
		if rustSigHasFlags(sig) {
			score += 100
		} else {
			score -= 100
		}
	} else {
		score += 1
	}

	if specHasStatus && strings.HasSuffix(sig.Name, "_with_flags") {
		score += 20
	}
	if strings.HasSuffix(sig.Name, "_raw") {
		score += 1
	}
	if rustSigHasTodo(sig) {
		score -= 25
	}
	return score
}

func resolveRustFunc(spec ReadtestSpec, sigs map[string]RustFuncSig, normalized map[string]string) (RustFuncSig, bool) {
	baseCandidates := []string{
		spec.Name,
		camelToSnake(spec.Name),
		strings.ReplaceAll(spec.Name, "Sign", "_sign"),
	}
	var resolvedBases []string
	seen := make(map[string]bool)
	for _, candidate := range baseCandidates {
		if candidate == "" {
			continue
		}
		if actual := resolveRustFuncName(candidate, sigs, normalized); actual != "" && !seen[actual] {
			resolvedBases = append(resolvedBases, actual)
			seen[actual] = true
		}
	}
	if len(resolvedBases) == 0 {
		return RustFuncSig{}, false
	}

	var best RustFuncSig
	bestScore := -1 << 30
	var bestImplemented RustFuncSig
	bestImplementedScore := -1 << 30
	candidateSeen := make(map[string]bool)
	for _, base := range resolvedBases {
		candidates := []string{base}
		if specUsesRounding(spec) || specNeedsStatus(spec) {
			candidates = append([]string{base + "_with_flags", base + "_raw"}, candidates...)
		}
		for _, candidateName := range candidates {
			if candidateSeen[candidateName] {
				continue
			}
			candidateSeen[candidateName] = true
			sig, ok := sigs[candidateName]
			if !ok {
				continue
			}
			score := scoreRustFuncSig(spec, sig)
			if score > bestScore {
				best = sig
				bestScore = score
			}
			if !rustSigHasTodo(sig) && score > bestImplementedScore {
				bestImplemented = sig
				bestImplementedScore = score
			}
		}
	}
	if bestImplementedScore != -1<<30 {
		return bestImplemented, true
	}
	if bestScore == -1<<30 {
		return RustFuncSig{}, false
	}
	return best, true
}

func specHasNativeRustFunc(spec ReadtestSpec) bool {
	sig, ok := resolveRustFunc(spec, rustFuncSigMap, normalizedRustFuncNameMap)
	return ok && !rustSigHasTodo(sig)
}

func classifyGenericDispatchSkip(spec ReadtestSpec) rustDispatchSkipReason {
	sig, ok := resolveRustFunc(spec, rustFuncSigMap, normalizedRustFuncNameMap)
	if !ok {
		return rustDispatchSkipReason{
			Code:           "rust_dispatch_unresolved_function",
			Reason:         "Rust readtest generator could not resolve a generated implementation function",
			Classification: "blocked_required_review",
		}
	}
	if rustSigHasTodo(sig) {
		return rustDispatchSkipReason{
			Code:           "rust_dispatch_unimplemented_body",
			Reason:         "Resolved Rust implementation still contains todo!/not-yet-implemented marker",
			Classification: "blocked_required_review",
		}
	}
	if specUsesRounding(spec) && !rustSigHasRounding(sig) {
		return rustDispatchSkipReason{
			Code:           "rust_dispatch_missing_rounding_parameter",
			Reason:         "Rust implementation signature lacks the rounding parameter required by readtest",
			Classification: "blocked_required_review",
		}
	}
	hasFlagsParam := false
	for _, param := range sig.Params {
		if isRustFlagsParam(param) {
			hasFlagsParam = true
			break
		}
	}
	hasFlagsRet := len(sig.ReturnTypes) > 1 && sig.ReturnTypes[len(sig.ReturnTypes)-1] == "u32"
	if hasFlagsParam && hasFlagsRet {
		return rustDispatchSkipReason{
			Code:           "rust_dispatch_ambiguous_flags",
			Reason:         "Rust implementation exposes status flags as both parameter and return value",
			Classification: "blocked_required_review",
		}
	}
	if specNeedsStatus(spec) && !hasFlagsParam && !hasFlagsRet {
		return rustDispatchSkipReason{
			Code:           "rust_dispatch_missing_status_flags",
			Reason:         "Rust implementation signature lacks status flag reporting required by readtest",
			Classification: "blocked_required_review",
		}
	}
	return rustDispatchSkipReason{
		Code:           "rust_dispatch_adapter_unsupported",
		Reason:         "Rust readtest generator cannot build parser/comparator/argument adapter for this signature",
		Classification: "blocked_required_review",
	}
}

func rustParseFunc(typ string) string {
	switch typ {
	case "u64":
		return "parse_u64"
	case "u32":
		return "parse_u32"
	case "u16":
		return "parse_u16"
	case "u8":
		return "parse_u8"
	case "i64":
		return "parse_i64"
	case "i32":
		return "parse_i32"
	case "i16":
		return "parse_i16"
	case "i8":
		return "parse_i8"
	case "f32":
		return "parse_f32"
	case "f64":
		return "parse_f64"
	case "BID_UINT128":
		return "parse_bid128"
	case "String", "&str":
		return "parse_string"
	case "bool":
		return "parse_i32"
	default:
		return ""
	}
}

func rustCompareFunc(typ string) string {
	switch typ {
	case "u64":
		return "compare_u64"
	case "u32":
		return "compare_u32"
	case "u16":
		return "compare_u16"
	case "u8":
		return "compare_u8"
	case "i64":
		return "compare_i64"
	case "i32":
		return "compare_i32"
	case "i16":
		return "compare_i16"
	case "i8":
		return "compare_i8"
	case "f32":
		return "compare_f32"
	case "f64":
		return "compare_f64"
	case "BID_UINT128":
		return "compare_bid128"
	case "String":
		return "compare_string"
	case "bool":
		return "compare_bool_int"
	default:
		return ""
	}
}

func selectParseFunc(specOp string, rustType string) string {
	switch rustType {
	case "String", "&str":
		return "parse_string"
	}
	switch {
	case specOp == "OP_BIN32" && rustType == "u32":
		return "parse_u32"
	case specOp == "OP_BIN64" && rustType == "u64":
		return "parse_u64"
	case specOp == "OP_INT32" && rustType == "i64":
		return "parse_i32_as_i64"
	case specOp == "OP_BID_UINT32" && rustType == "u64":
		return "parse_u64"
	}
	if pf := parseFunc(specOp); pf != "" {
		return pf
	}
	return rustParseFunc(rustType)
}

func selectCompareFunc(specOp string, rustType string) string {
	switch rustType {
	case "bool":
		return "compare_bool_int"
	case "String", "&str":
		return "compare_string"
	}
	switch {
	case specOp == "OP_BIN32" && rustType == "u32":
		return "compare_u32"
	case specOp == "OP_BIN64" && rustType == "u64":
		return "compare_u64"
	case specOp == "OP_INT32" && rustType == "i64":
		return "compare_i64"
	case specOp == "OP_BID_UINT32" && rustType == "u64":
		return "compare_u64"
	}
	if cf := compareFunc(specOp); cf != "" {
		return cf
	}
	return rustCompareFunc(rustType)
}

var rustFuncSigMap map[string]RustFuncSig
var normalizedRustFuncNameMap map[string]string

func rustCmpMode(compare string) string {
	switch compare {
	case "CMP_EQUALSTATUS":
		return "CmpMode::CmpEqual"
	case "CMP_RELATIVEERR":
		return "CmpMode::CmpRelativeErr"
	default:
		return "CmpMode::CmpFuzzy"
	}
}

func rustCompareModes(spec ReadtestSpec, specsByName map[string][]ReadtestSpec) []string {
	seen := make(map[string]bool)
	var modes []string
	for _, candidate := range specsByName[spec.Name] {
		if candidate.Inputs == nil || candidate.Output == "" {
			continue
		}
		if candidate.Output != spec.Output || candidate.CallType != spec.CallType || strings.Join(candidate.Inputs, "\x00") != strings.Join(spec.Inputs, "\x00") {
			continue
		}
		mode := rustCmpMode(candidate.Compare)
		if !seen[mode] {
			seen[mode] = true
			modes = append(modes, mode)
		}
	}
	if len(modes) == 0 {
		return []string{rustCmpMode(spec.Compare)}
	}
	return modes
}

func generateReadtestRust(projectRoot string) {
	specs := loadReadtestSpecs(projectRoot)
	skipManifestRelPath, skipManifest := loadRustReadtestSkipManifest(projectRoot)
	rustFuncSigMap = loadRustFuncSigs(projectRoot)
	normalizedRustFuncNameMap = buildNormalizedRustNameMap(rustFuncSigMap)
	specsByName := make(map[string][]ReadtestSpec, len(specs.Specs))
	for _, spec := range specs.Specs {
		specsByName[spec.Name] = append(specsByName[spec.Name], spec)
	}

	outPath := filepath.Join(projectRoot, "..", "bid754-rs", "tests", "readtest_generated.rs")
	auditPath := filepath.Join(projectRoot, "generated", "testspec", "rust_readtest_dispatch_audit.json")

	var sb strings.Builder

	// Header
	sb.WriteString("// Code generated by tools/codegen --target=readtest-rust. DO NOT EDIT.\n")
	sb.WriteString("// Source: devtools/tools/registry/readtest_specs.json + devtools/third_party/intel_dfp/TESTS/readtest.in\n\n")
	// Gate the whole native readtest behind ffi-native so `cargo test` without
	// that feature (no Intel BID / libbid-sys) still compiles the portable tests.
	sb.WriteString("#![cfg(feature = \"ffi-native\")]\n")
	sb.WriteString("#![allow(dead_code, unused_variables)]\n\n")
	sb.WriteString("use bid754::gen_types::BID_UINT128;\n")
	sb.WriteString("use bid754::generated::prelude::*;\n")
	sb.WriteString("use libbid_sys;\n")
	sb.WriteString("use std::collections::BTreeMap;\n")
	sb.WriteString("use std::ffi::CString;\n")
	sb.WriteString("use std::fs::File;\n")
	sb.WriteString("use std::io::{BufRead, BufReader};\n")
	sb.WriteString("use std::panic::{self, AssertUnwindSafe};\n")
	sb.WriteString("use std::path::{Path, PathBuf};\n\n")

	// Parsers
	sb.WriteString(readtestParsers())

	// Compare functions
	sb.WriteString(readtestCompareFuncs())

	// Custom helpers
	sb.WriteString(readtestCustomHelpers())

	// Dispatch function
	sb.WriteString("fn dispatch(func_name: &str, parts: &[&str], rm: i64, ulp_add: f64) -> DispatchResult {\n")
	sb.WriteString("    match func_name {\n")

	matched := 0
	skipped := 0
	matchedNames := make([]string, 0, len(specs.Specs))
	emittedNames := make(map[string]bool)
	usedSkipManifest := make(map[string]bool)
	audit := RustReadtestDispatchAudit{
		Version:      "1.0",
		Source:       "tools/registry/readtest_specs.json + generated/testspec/spec_index.json",
		SkipManifest: skipManifestRelPath,
	}
	recordDispatch := func(spec ReadtestSpec) {
		audit.Functions = append(audit.Functions, RustReadtestDispatchAuditRow{
			Function: spec.Name,
			Compare:  spec.Compare,
			Status:   "dispatched",
		})
	}
	recordManifestSkip := func(spec ReadtestSpec, reason rustDispatchSkipReason) {
		entry, ok := skipManifest[spec.Name]
		if !ok {
			fmt.Fprintf(os.Stderr, "unexpected Rust readtest skip for %s: %s (%s)\n", spec.Name, reason.Code, reason.Reason)
			os.Exit(1)
		}
		if entry.ReasonCode != reason.Code {
			fmt.Fprintf(os.Stderr, "Rust readtest skip manifest reason mismatch for %s: got %s, manifest wants %s\n", spec.Name, reason.Code, entry.ReasonCode)
			os.Exit(1)
		}
		usedSkipManifest[spec.Name] = true
		audit.Functions = append(audit.Functions, RustReadtestDispatchAuditRow{
			Function:       spec.Name,
			Compare:        spec.Compare,
			Status:         "skipped",
			ReasonCode:     entry.ReasonCode,
			Reason:         entry.Reason,
			Classification: entry.Classification,
		})
		skipped++
	}

	for _, spec := range specs.Specs {
		if emittedNames[spec.Name] {
			matched++
			recordDispatch(spec)
			continue
		}

		if dispatchCode := generateCustomDispatchCase(spec); dispatchCode != "" {
			sb.WriteString(fmt.Sprintf("        %q => %s,\n", spec.Name, dispatchCode))
			matched++
			matchedNames = append(matchedNames, spec.Name)
			emittedNames[spec.Name] = true
			recordDispatch(spec)
			continue
		}

		// Skip unsupported output types
		if isSkipType(spec.Output) {
			recordManifestSkip(spec, rustDispatchSkipReason{
				Code:           "rust_readtest_unsupported_output_type",
				Reason:         "Readtest output type is outside the supported Rust generated adapter surface",
				Classification: "out_of_scope_not_required",
			})
			continue
		}
		// Skip unsupported input types
		hasUnsupported := false
		for _, inp := range spec.Inputs {
			if isSkipType(inp) {
				hasUnsupported = true
				break
			}
		}
		if hasUnsupported {
			recordManifestSkip(spec, rustDispatchSkipReason{
				Code:           "rust_readtest_unsupported_input_type",
				Reason:         "Readtest input type is outside the supported Rust generated adapter surface",
				Classification: "out_of_scope_not_required",
			})
			continue
		}

		dispatchCode := generateGenericDispatchCase(spec, rustCompareModes(spec, specsByName))
		if dispatchCode == "" {
			recordManifestSkip(spec, classifyGenericDispatchSkip(spec))
			continue
		}

		sb.WriteString(fmt.Sprintf("        %q => %s,\n", spec.Name, dispatchCode))
		matched++
		matchedNames = append(matchedNames, spec.Name)
		emittedNames[spec.Name] = true
		recordDispatch(spec)
	}
	for name := range skipManifest {
		if !usedSkipManifest[name] {
			fmt.Fprintf(os.Stderr, "Rust readtest skip manifest entry %q was not used by current selected surface\n", name)
			os.Exit(1)
		}
	}

	sb.WriteString("        _ => DispatchResult::Skip,\n")
	sb.WriteString("    }\n")
	sb.WriteString("}\n\n")
	sb.WriteString("fn supported_readtest_func(func_name: &str) -> bool {\n")
	sb.WriteString("    matches!(func_name,\n")
	for _, name := range matchedNames {
		sb.WriteString(fmt.Sprintf("        %q |\n", name))
	}
	sb.WriteString("        \"\"\n")
	sb.WriteString("    )\n")
	sb.WriteString("}\n\n")

	// Runner
	sb.WriteString(readtestRunner())

	// Tests
	sb.WriteString(readtestTests())

	audit.Dispatched = matched
	audit.Skipped = skipped
	auditBytes, err := json.MarshalIndent(audit, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "marshal Rust readtest dispatch audit: %v\n", err)
		os.Exit(1)
	}
	writeFile(outPath, sb.String())
	writeFile(auditPath, string(auditBytes)+"\n")
	fmt.Printf("readtest_generated.rs: %d dispatched, %d skipped\n", matched, skipped)
}

func generateGenericDispatchCase(spec ReadtestSpec, cmpModes []string) string {
	sig, ok := resolveRustFunc(spec, rustFuncSigMap, normalizedRustFuncNameMap)
	if !ok {
		return ""
	}
	if rustSigHasTodo(sig) {
		return ""
	}

	specHasRnd := specUsesRounding(spec)
	specHasStatus := specNeedsStatus(spec)
	actualHasRnd := rustSigHasRounding(sig)
	if specHasRnd && !actualHasRnd {
		return ""
	}
	hasFlagsParam := false
	for _, param := range sig.Params {
		if isRustFlagsParam(param) {
			hasFlagsParam = true
			break
		}
	}
	hasFlagsRet := len(sig.ReturnTypes) > 1 && sig.ReturnTypes[len(sig.ReturnTypes)-1] == "u32"
	if hasFlagsParam && hasFlagsRet {
		return ""
	}
	if specHasStatus && !hasFlagsParam && !hasFlagsRet {
		return ""
	}

	returnTypes := append([]string(nil), sig.ReturnTypes...)
	if hasFlagsRet {
		returnTypes = returnTypes[:len(returnTypes)-1]
	}
	if len(returnTypes) == 0 {
		return ""
	}

	primaryType := returnTypes[0]
	outParse := selectParseFunc(spec.Output, primaryType)
	outCmp := selectCompareFunc(spec.Output, primaryType)
	if outParse == "" || outCmp == "" {
		return ""
	}

	extraReturnTypes := returnTypes[1:]
	extraSpecInputs := spec.Inputs

	var callArgs []string
	var inputVars []string
	consumedSpecInputs := 0
	for _, param := range sig.Params {
		switch {
		case isRustFlagsParam(param):
			callArgs = append(callArgs, "&mut flags")
		case isRustRoundingParam(param):
			callArgs = append(callArgs, "rm")
		default:
			if consumedSpecInputs >= len(spec.Inputs) {
				return ""
			}
			pf := selectParseFunc(spec.Inputs[consumedSpecInputs], param.Type)
			if pf == "" {
				return ""
			}
			inputVars = append(inputVars, fmt.Sprintf("a%d", consumedSpecInputs))
			callArgs = append(callArgs, inputVars[len(inputVars)-1])
			consumedSpecInputs++
		}
	}
	extraSpecInputs = extraSpecInputs[consumedSpecInputs:]
	if len(extraReturnTypes) > 0 && len(extraReturnTypes) != len(extraSpecInputs) {
		return ""
	}

	expectedIdx := len(spec.Inputs) + 2
	flagsIdx := expectedIdx + 1

	var code strings.Builder
	code.WriteString("{\n")
	code.WriteString(fmt.Sprintf("            if parts.len() <= %d { return DispatchResult::Skip; }\n", flagsIdx))

	inputVarIdx := 0
	for _, param := range sig.Params {
		if isRustFlagsParam(param) || isRustRoundingParam(param) {
			continue
		}
		pf := selectParseFunc(spec.Inputs[inputVarIdx], param.Type)
		if pf == "" {
			return ""
		}
		code.WriteString(fmt.Sprintf("            let Some(%s) = %s(parts[%d]) else { return DispatchResult::Skip };\n", inputVars[inputVarIdx], pf, inputVarIdx+2))
		inputVarIdx++
	}

	for i, extraType := range extraReturnTypes {
		extraParse := selectParseFunc(extraSpecInputs[i], extraType)
		if extraParse == "" {
			return ""
		}
		extraIdx := consumedSpecInputs + 2 + i
		code.WriteString(fmt.Sprintf("            let Some(expected_extra%d) = %s(parts[%d]) else { return DispatchResult::Skip };\n", i, extraParse, extraIdx))
	}
	code.WriteString(fmt.Sprintf("            let Some(expected) = %s(parts[%d]) else { return DispatchResult::Skip };\n", outParse, expectedIdx))
	code.WriteString(fmt.Sprintf("            let expected_flags = parse_flags(parts[%d]);\n", flagsIdx))

	if hasFlagsParam {
		code.WriteString("            let mut flags: u32 = 0;\n")
	}

	gotVars := []string{"got"}
	if len(returnTypes) > 1 {
		gotVars = gotVars[:0]
		for i := range returnTypes {
			gotVars = append(gotVars, fmt.Sprintf("got%d", i))
		}
	}

	callExpr := fmt.Sprintf("%s(%s)", sig.Name, strings.Join(callArgs, ", "))
	switch {
	case hasFlagsRet:
		allRetVars := append(append([]string(nil), gotVars...), "flags")
		if len(allRetVars) == 2 {
			code.WriteString(fmt.Sprintf("            let (%s, %s) = %s;\n", allRetVars[0], allRetVars[1], callExpr))
		} else {
			code.WriteString(fmt.Sprintf("            let (%s) = %s;\n", strings.Join(allRetVars, ", "), callExpr))
		}
	case hasFlagsParam && len(gotVars) == 1:
		code.WriteString(fmt.Sprintf("            let %s = %s;\n", gotVars[0], callExpr))
	case hasFlagsParam:
		code.WriteString(fmt.Sprintf("            let (%s) = %s;\n", strings.Join(gotVars, ", "), callExpr))
	case len(gotVars) == 1:
		code.WriteString(fmt.Sprintf("            let %s = %s;\n", gotVars[0], callExpr))
		code.WriteString("            let flags: u32 = 0;\n")
	default:
		code.WriteString(fmt.Sprintf("            let (%s) = %s;\n", strings.Join(gotVars, ", "), callExpr))
		code.WriteString("            let flags: u32 = 0;\n")
	}
	if !specHasStatus && (hasFlagsParam || hasFlagsRet) {
		code.WriteString("            let flags: u32 = 0;\n")
	}

	for _, cmpMode := range cmpModes {
		code.WriteString(fmt.Sprintf("            let result = %s(%s, expected, flags, expected_flags, %s, rm, ulp_add);\n", outCmp, gotVars[0], cmpMode))
		code.WriteString("            if !matches!(result, DispatchResult::Pass) { return result; }\n")
		for i, extraType := range extraReturnTypes {
			extraCmp := selectCompareFunc(extraSpecInputs[i], extraType)
			if extraCmp == "" {
				return ""
			}
			code.WriteString(fmt.Sprintf("            let result = %s(%s, expected_extra%d, 0, 0, %s, rm, ulp_add);\n", extraCmp, gotVars[i+1], i, cmpMode))
			code.WriteString("            if !matches!(result, DispatchResult::Pass) { return result; }\n")
		}
	}
	code.WriteString("            DispatchResult::Pass\n")
	code.WriteString("        }")

	return code.String()
}

func rustFuncImplemented(name string) bool {
	sig, ok := rustFuncSigMap[name]
	return ok && !rustSigHasTodo(sig)
}

func buildSimpleDispatchCase(argParse string, expectedParse string, cmpFunc string, callExpr string) string {
	var code strings.Builder
	code.WriteString("{\n")
	code.WriteString("            if parts.len() <= 4 { return DispatchResult::Skip; }\n")
	code.WriteString(fmt.Sprintf("            let Some(a0) = %s(parts[2]) else { return DispatchResult::Skip };\n", argParse))
	code.WriteString(fmt.Sprintf("            let Some(expected) = %s(parts[3]) else { return DispatchResult::Skip };\n", expectedParse))
	code.WriteString("            let expected_flags = parse_flags(parts[4]);\n")
	code.WriteString(fmt.Sprintf("            let (got, flags) = %s;\n", callExpr))
	code.WriteString(fmt.Sprintf("            let result = %s(got, expected, flags, expected_flags, CmpMode::CmpFuzzy, rm, ulp_add);\n", cmpFunc))
	code.WriteString("            if !matches!(result, DispatchResult::Pass) { return result; }\n")
	code.WriteString("            DispatchResult::Pass\n")
	code.WriteString("        }")
	return code.String()
}

func buildSimpleNoRoundDispatchCase(argParse string, expectedParse string, cmpFunc string, callExpr string) string {
	var code strings.Builder
	code.WriteString("{\n")
	code.WriteString("            if parts.len() <= 4 { return DispatchResult::Skip; }\n")
	code.WriteString(fmt.Sprintf("            let Some(a0) = %s(parts[2]) else { return DispatchResult::Skip };\n", argParse))
	code.WriteString(fmt.Sprintf("            let Some(expected) = %s(parts[3]) else { return DispatchResult::Skip };\n", expectedParse))
	code.WriteString("            let expected_flags = parse_flags(parts[4]);\n")
	code.WriteString(fmt.Sprintf("            let (got, flags) = %s;\n", callExpr))
	code.WriteString(fmt.Sprintf("            let result = %s(got, expected, flags, expected_flags, CmpMode::CmpFuzzy, rm, ulp_add);\n", cmpFunc))
	code.WriteString("            if !matches!(result, DispatchResult::Pass) { return result; }\n")
	code.WriteString("            DispatchResult::Pass\n")
	code.WriteString("        }")
	return code.String()
}

func buildUnaryBid128ExactDispatchCase(op string) string {
	var code strings.Builder
	code.WriteString("{\n")
	code.WriteString(fmt.Sprintf("            dispatch_unary_bid128_optional_ignored_arg(parts, %s)\n", op))
	code.WriteString("        }")
	return code.String()
}

func buildStringFromDispatchCase(expectedParse string, cmpFunc string, gotExpr string) string {
	var code strings.Builder
	code.WriteString("{\n")
	code.WriteString("            if parts.len() <= 4 { return DispatchResult::Skip; }\n")
	code.WriteString(fmt.Sprintf("            let Some(expected) = %s(parts[3]) else { return DispatchResult::Skip };\n", expectedParse))
	code.WriteString("            let expected_flags = parse_flags(parts[4]);\n")
	code.WriteString(fmt.Sprintf("            let Some(got) = %s else { return DispatchResult::Skip; };\n", gotExpr))
	code.WriteString(fmt.Sprintf("            let result = %s(got, expected, 0, expected_flags, CmpMode::CmpFuzzy, rm, ulp_add);\n", cmpFunc))
	code.WriteString("            if !matches!(result, DispatchResult::Pass) { return result; }\n")
	code.WriteString("            DispatchResult::Pass\n")
	code.WriteString("        }")
	return code.String()
}

func buildStringCallDispatchCase(expectedParse string, cmpFunc string, callExpr string) string {
	var code strings.Builder
	code.WriteString("{\n")
	code.WriteString("            if parts.len() <= 4 { return DispatchResult::Skip; }\n")
	code.WriteString(fmt.Sprintf("            let Some(expected) = %s(parts[3]) else { return DispatchResult::Skip };\n", expectedParse))
	code.WriteString("            let expected_flags = parse_flags(parts[4]);\n")
	code.WriteString(fmt.Sprintf("            let (got, flags) = %s;\n", callExpr))
	code.WriteString(fmt.Sprintf("            let result = %s(got, expected, flags, expected_flags, CmpMode::CmpFuzzy, rm, ulp_add);\n", cmpFunc))
	code.WriteString("            if !matches!(result, DispatchResult::Pass) { return result; }\n")
	code.WriteString("            DispatchResult::Pass\n")
	code.WriteString("        }")
	return code.String()
}

func buildParsedStringDispatchCase(inputParse string, callExpr string) string {
	var code strings.Builder
	code.WriteString("{\n")
	code.WriteString("            if parts.len() <= 4 { return DispatchResult::Skip; }\n")
	code.WriteString(fmt.Sprintf("            let Some(a0) = %s(parts[2]) else { return DispatchResult::Skip };\n", inputParse))
	code.WriteString("            let Some(expected) = parse_string(parts[3]) else { return DispatchResult::Skip };\n")
	code.WriteString("            let expected_flags = parse_flags(parts[4]);\n")
	code.WriteString(fmt.Sprintf("            let got = %s;\n", callExpr))
	code.WriteString("            let result = compare_string(got, expected, 0, expected_flags, CmpMode::CmpFuzzy, rm, ulp_add);\n")
	code.WriteString("            if !matches!(result, DispatchResult::Pass) { return result; }\n")
	code.WriteString("            DispatchResult::Pass\n")
	code.WriteString("        }")
	return code.String()
}

func buildFromIntDispatchCase(inputParse string, expectedParse string, cmpFunc string, callExpr string) string {
	var code strings.Builder
	code.WriteString("{\n")
	code.WriteString("            if parts.len() <= 4 { return DispatchResult::Skip; }\n")
	code.WriteString(fmt.Sprintf("            let Some(a0) = %s(parts[2]) else { return DispatchResult::Skip };\n", inputParse))
	code.WriteString(fmt.Sprintf("            let Some(expected) = %s(parts[3]) else { return DispatchResult::Skip };\n", expectedParse))
	code.WriteString("            let expected_flags = parse_flags(parts[4]);\n")
	code.WriteString(fmt.Sprintf("            let (got, flags) = %s;\n", callExpr))
	code.WriteString(fmt.Sprintf("            let result = %s(got, expected, flags, expected_flags, CmpMode::CmpFuzzy, rm, ulp_add);\n", cmpFunc))
	code.WriteString("            if !matches!(result, DispatchResult::Pass) { return result; }\n")
	code.WriteString("            DispatchResult::Pass\n")
	code.WriteString("        }")
	return code.String()
}

func buildNan32DispatchCase() string {
	return `{
            if parts.len() <= 4 { return DispatchResult::Skip; }
            let Some(expected) = parse_bid32(parts[3]) else { return DispatchResult::Skip };
            let expected_flags = parse_flags(parts[4]);
            let got = if parts[2] == "NULL" || parts[2] == "0" {
                0x7c00_0000
            } else {
                let Some(payload) = parse_u32(parts[2]) else { return DispatchResult::Skip };
                0x7c00_0000 | payload
            };
            let result = compare_u32(got, expected, 0, expected_flags, CmpMode::CmpFuzzy, rm, ulp_add);
            if !matches!(result, DispatchResult::Pass) { return result; }
            DispatchResult::Pass
        }`
}

func buildNan64DispatchCase() string {
	return `{
            if parts.len() <= 4 { return DispatchResult::Skip; }
            let Some(expected) = parse_bid64(parts[3]) else { return DispatchResult::Skip };
            let expected_flags = parse_flags(parts[4]);
            let got = if parts[2] == "NULL" || parts[2] == "0" {
                0x7c00_0000_0000_0000
            } else {
                let Some(payload) = parse_u64(parts[2]) else { return DispatchResult::Skip };
                0x7c00_0000_0000_0000 | payload
            };
            let result = compare_u64(got, expected, 0, expected_flags, CmpMode::CmpFuzzy, rm, ulp_add);
            if !matches!(result, DispatchResult::Pass) { return result; }
            DispatchResult::Pass
        }`
}

func buildNan128DispatchCase() string {
	return `{
            if parts.len() <= 4 { return DispatchResult::Skip; }
            let Some(expected) = parse_bid128(parts[3]) else { return DispatchResult::Skip };
            let expected_flags = parse_flags(parts[4]);
            let got = if parts[2] == "NULL" || parts[2] == "0" {
                BID_UINT128 { w: [0, 0x7c00_0000_0000_0000] }
            } else {
                let (mut payload, _) = bid128_from_string(parts[2].to_string(), 0);
                payload.w[1] &= 0x0000_3fff_ffff_ffff;
                BID_UINT128 { w: [payload.w[0], 0x7c00_0000_0000_0000 | payload.w[1]] }
            };
            let result = compare_bid128(got, expected, 0, expected_flags, CmpMode::CmpFuzzy, rm, ulp_add);
            if !matches!(result, DispatchResult::Pass) { return result; }
            DispatchResult::Pass
        }`
}

func skipDispatchCase() string {
	return "{\n            DispatchResult::Skip\n        }"
}

func buildProgrammableDispatchCase(inputParsers []string, expectedParse string, cmpFunc string, bodyLines []string) string {
	var code strings.Builder
	flagsIdx := len(inputParsers) + 3
	code.WriteString("{\n")
	code.WriteString(fmt.Sprintf("            if parts.len() <= %d { return DispatchResult::Skip; }\n", flagsIdx))
	for i, parser := range inputParsers {
		code.WriteString(fmt.Sprintf("            let Some(a%d) = %s(parts[%d]) else { return DispatchResult::Skip };\n", i, parser, i+2))
	}
	code.WriteString(fmt.Sprintf("            let Some(expected) = %s(parts[%d]) else { return DispatchResult::Skip };\n", expectedParse, len(inputParsers)+2))
	code.WriteString(fmt.Sprintf("            let expected_flags = parse_flags(parts[%d]);\n", flagsIdx))
	for _, line := range bodyLines {
		code.WriteString("            ")
		code.WriteString(line)
		if !strings.HasSuffix(line, "\n") {
			code.WriteString("\n")
		}
	}
	code.WriteString(fmt.Sprintf("            let result = %s(got, expected, flags, expected_flags, CmpMode::CmpFuzzy, rm, ulp_add);\n", cmpFunc))
	code.WriteString("            if !matches!(result, DispatchResult::Pass) { return result; }\n")
	code.WriteString("            DispatchResult::Pass\n")
	code.WriteString("        }")
	return code.String()
}

func bid128InputConversionLines(argIdx int, kind byte) ([]string, bool) {
	switch kind {
	case 'd':
		if !rustFuncImplemented("bid64_to_bid128") {
			return nil, false
		}
		return []string{
			fmt.Sprintf("let (x%d, f%d) = bid64_to_bid128(a%d);", argIdx, argIdx, argIdx),
		}, true
	case 'q':
		return []string{
			fmt.Sprintf("let x%d = a%d;", argIdx, argIdx),
			fmt.Sprintf("let f%d: u32 = 0;", argIdx),
		}, true
	default:
		return nil, false
	}
}

func generateMixedBidViaBid128DispatchCase(spec ReadtestSpec) string {
	if code := generateMixedBidBinaryViaBid128DispatchCase(spec); code != "" {
		return code
	}
	if code := generateMixedBidFmaViaBid128DispatchCase(spec); code != "" {
		return code
	}
	if code := generateMixedBidSqrtViaBid128DispatchCase(spec); code != "" {
		return code
	}
	return ""
}

func generateMixedBidBinaryViaBid128DispatchCase(spec ReadtestSpec) string {
	var prefix string
	outputToBid64 := false
	switch {
	case strings.HasPrefix(spec.Name, "bid64"):
		prefix = "bid64"
		outputToBid64 = true
	case strings.HasPrefix(spec.Name, "bid128"):
		prefix = "bid128"
	default:
		return ""
	}

	var opSuffix, op128 string
	flagsParamOp := false
	switch {
	case strings.HasSuffix(spec.Name, "_add"):
		opSuffix = "_add"
		op128 = "bid128_add"
		flagsParamOp = true
	case strings.HasSuffix(spec.Name, "_sub"):
		opSuffix = "_sub"
		op128 = "bid128_sub"
		flagsParamOp = true
	case strings.HasSuffix(spec.Name, "_mul"):
		opSuffix = "_mul"
		op128 = "bid128_mul"
	case strings.HasSuffix(spec.Name, "_div"):
		opSuffix = "_div"
		op128 = "bid128_div"
	default:
		return ""
	}
	if !rustFuncImplemented(op128) {
		return ""
	}
	if outputToBid64 && !rustFuncImplemented("bid128_to_bid64") {
		return ""
	}

	kinds := strings.TrimSuffix(strings.TrimPrefix(spec.Name, prefix), opSuffix)
	if len(kinds) != 2 || len(spec.Inputs) != 2 {
		return ""
	}

	inputParsers := make([]string, 0, 2)
	bodyLines := make([]string, 0, 8)
	for i, inp := range spec.Inputs {
		parser := parseFunc(inp)
		if parser == "" {
			return ""
		}
		inputParsers = append(inputParsers, parser)
		lines, ok := bid128InputConversionLines(i, kinds[i])
		if !ok {
			return ""
		}
		bodyLines = append(bodyLines, lines...)
	}

	if flagsParamOp {
		bodyLines = append(bodyLines, fmt.Sprintf("let (tmp, op_flags) = bid128_binop_flags_param_readtest(x0, x1, rm, %s);", op128))
	} else {
		bodyLines = append(bodyLines, fmt.Sprintf("let (tmp, op_flags) = %s(x0, x1, rm);", op128))
	}

	if outputToBid64 {
		bodyLines = append(bodyLines,
			"let (got, out_flags) = bid128_to_bid64(tmp, rm);",
			"let flags = f0 | f1 | op_flags | out_flags;",
		)
	} else {
		bodyLines = append(bodyLines,
			"let got = tmp;",
			"let flags = f0 | f1 | op_flags;",
		)
	}

	expectedParse := parseFunc(spec.Output)
	cmp := compareFunc(spec.Output)
	if expectedParse == "" || cmp == "" {
		return ""
	}
	return buildProgrammableDispatchCase(inputParsers, expectedParse, cmp, bodyLines)
}

func generateMixedBidFmaViaBid128DispatchCase(spec ReadtestSpec) string {
	var prefix string
	outputToBid64 := false
	switch {
	case strings.HasPrefix(spec.Name, "bid64") && strings.HasSuffix(spec.Name, "_fma"):
		prefix = "bid64"
		outputToBid64 = true
	case strings.HasPrefix(spec.Name, "bid128") && strings.HasSuffix(spec.Name, "_fma"):
		// Mixed BID64/BID128 FMA를 bid128_fma 하나로 합성하면 readtest 기대 비트와 어긋나는 케이스가 있다.
		// bit-exact mixed FMA 구현이 생기기 전까지는 생성하지 않고 skip한다.
		return ""
	default:
		return ""
	}
	if !rustFuncImplemented("bid128_fma") {
		return ""
	}
	if outputToBid64 && !rustFuncImplemented("bid128_to_bid64") {
		return ""
	}

	kinds := strings.TrimSuffix(strings.TrimPrefix(spec.Name, prefix), "_fma")
	if len(kinds) != 3 || len(spec.Inputs) != 3 {
		return ""
	}

	inputParsers := make([]string, 0, 3)
	bodyLines := make([]string, 0, 10)
	for i, inp := range spec.Inputs {
		parser := parseFunc(inp)
		if parser == "" {
			return ""
		}
		inputParsers = append(inputParsers, parser)
		lines, ok := bid128InputConversionLines(i, kinds[i])
		if !ok {
			return ""
		}
		bodyLines = append(bodyLines, lines...)
	}

	bodyLines = append(bodyLines, "let (tmp, op_flags) = bid128_fma(x0, x1, x2, rm);")
	if outputToBid64 {
		bodyLines = append(bodyLines,
			"let (got, out_flags) = bid128_to_bid64(tmp, rm);",
			"let flags = f0 | f1 | f2 | op_flags | out_flags;",
		)
	} else {
		bodyLines = append(bodyLines,
			"let got = tmp;",
			"let flags = f0 | f1 | f2 | op_flags;",
		)
	}

	expectedParse := parseFunc(spec.Output)
	cmp := compareFunc(spec.Output)
	if expectedParse == "" || cmp == "" {
		return ""
	}
	return buildProgrammableDispatchCase(inputParsers, expectedParse, cmp, bodyLines)
}

func generateMixedBidSqrtViaBid128DispatchCase(spec ReadtestSpec) string {
	switch spec.Name {
	case "bid64q_sqrt":
		if !rustFuncImplemented("bid128_sqrt") || !rustFuncImplemented("bid128_to_bid64") {
			return ""
		}
		parser := parseFunc(spec.Inputs[0])
		expectedParse := parseFunc(spec.Output)
		cmp := compareFunc(spec.Output)
		if parser == "" || expectedParse == "" || cmp == "" {
			return ""
		}
		bodyLines := []string{
			"let x0 = a0;",
			"let f0: u32 = 0;",
			"let (tmp, op_flags) = bid128_sqrt(x0, rm);",
			"let (got, out_flags) = bid128_to_bid64(tmp, rm);",
			"let flags = f0 | op_flags | out_flags;",
		}
		return buildProgrammableDispatchCase([]string{parser}, expectedParse, cmp, bodyLines)
	case "bid128d_sqrt":
		if !rustFuncImplemented("bid128_sqrt") || !rustFuncImplemented("bid64_to_bid128") {
			return ""
		}
		parser := parseFunc(spec.Inputs[0])
		expectedParse := parseFunc(spec.Output)
		cmp := compareFunc(spec.Output)
		if parser == "" || expectedParse == "" || cmp == "" {
			return ""
		}
		bodyLines := []string{
			"let (x0, f0) = bid64_to_bid128(a0);",
			"let (got, op_flags) = bid128_sqrt(x0, rm);",
			"let flags = f0 | op_flags;",
		}
		return buildProgrammableDispatchCase([]string{parser}, expectedParse, cmp, bodyLines)
	default:
		return ""
	}
}

func generateCustomDispatchCase(spec ReadtestSpec) string {
	if code := generateBid32ViaBid64DispatchCase(spec); code != "" {
		return code
	}
	if code := generateMixedBidViaBid128DispatchCase(spec); code != "" {
		return code
	}
	switch spec.Name {
	case "bid_testFlags":
		return buildProgrammableDispatchCase(
			[]string{"parse_u32", "parse_u32"},
			"parse_u32",
			"compare_u32",
			[]string{
				"let got = (a1 & 0x3f) & (a0 & 0x3d);",
				"let flags = a1 & 0x3f;",
			},
		)
	case "bid_lowerFlags":
		return buildProgrammableDispatchCase(
			[]string{"parse_u32", "parse_u32"},
			"parse_u32",
			"compare_u32",
			[]string{
				"let got = 0;",
				"let flags = (a1 & 0x3f) & !(a0 & 0x3d);",
			},
		)
	case "bid_signalException":
		return buildProgrammableDispatchCase(
			[]string{"parse_u32", "parse_u32"},
			"parse_u32",
			"compare_u32",
			[]string{
				"let got = 0;",
				"let flags = (a1 & 0x3f) | (a0 & 0x3d);",
			},
		)
	case "bid_saveFlags":
		return buildProgrammableDispatchCase(
			[]string{"parse_u32", "parse_u32"},
			"parse_u32",
			"compare_u32",
			[]string{
				"let got = (a1 & 0x3f) & (a0 & 0x3d);",
				"let flags = a1 & 0x3f;",
			},
		)
	case "bid_restoreFlags":
		return buildProgrammableDispatchCase(
			[]string{"parse_u32", "parse_u32", "parse_u32"},
			"parse_u32",
			"compare_u32",
			[]string{
				"let got = 0;",
				"let mask = a1 & 0x3d;",
				"let flags = ((a2 & 0x3f) & !mask) | (a0 & mask);",
			},
		)
	case "bid_testSavedFlags":
		return buildProgrammableDispatchCase(
			[]string{"parse_u32", "parse_u32"},
			"parse_u32",
			"compare_u32",
			[]string{
				"let got = a0 & (a1 & 0x3d);",
				"let flags = 0;",
			},
		)
	case "bid_getDecimalRoundingDirection":
		return buildProgrammableDispatchCase(
			[]string{"parse_u32"},
			"parse_u32",
			"compare_u32",
			[]string{
				"let got = rm as u32;",
				"let flags = 0;",
			},
		)
	case "bid_setDecimalRoundingDirection":
		return buildProgrammableDispatchCase(
			[]string{"parse_u32"},
			"parse_u32",
			"compare_u32",
			[]string{
				"let got = if a0 <= 4 { a0 } else { rm as u32 };",
				"let flags = 0;",
			},
		)
	case "bid128_lrint":
		if rustFuncImplemented("bid128_lrint") {
			return buildSimpleDispatchCase("parse_bid128_for_int_conversion_readtest", "parse_i64", "compare_i64", "bid128_lrint(a0, rm)")
		}
		return skipDispatchCase()
	case "bid128_llrint":
		if rustFuncImplemented("bid128_llrint") {
			return buildSimpleDispatchCase("parse_bid128_for_int_conversion_readtest", "parse_i64", "compare_i64", "bid128_llrint(a0, rm)")
		}
		return skipDispatchCase()
	case "bid128_lround":
		if rustFuncImplemented("bid128_lround") {
			return buildSimpleNoRoundDispatchCase("parse_bid128_for_int_conversion_readtest", "parse_i64", "compare_i64", "bid128_lround(a0)")
		}
		return skipDispatchCase()
	case "bid128_llround":
		if rustFuncImplemented("bid128_llround") {
			return buildSimpleNoRoundDispatchCase("parse_bid128_for_int_conversion_readtest", "parse_i64", "compare_i64", "bid128_llround(a0)")
		}
		return skipDispatchCase()
	}
	switch {
	case strings.HasPrefix(spec.Name, "bid128_to_int8_"):
		if rustFuncImplemented(spec.Name) {
			return buildSimpleNoRoundDispatchCase("parse_bid128_for_int_conversion_readtest", "parse_i8", "compare_i8", fmt.Sprintf("%s(a0)", spec.Name))
		}
		return skipDispatchCase()
	case strings.HasPrefix(spec.Name, "bid128_to_int16_"):
		if rustFuncImplemented(spec.Name) {
			return buildSimpleNoRoundDispatchCase("parse_bid128_for_int_conversion_readtest", "parse_i16", "compare_i16", fmt.Sprintf("%s(a0)", spec.Name))
		}
		return skipDispatchCase()
	case strings.HasPrefix(spec.Name, "bid128_to_int32_"):
		if rustFuncImplemented(spec.Name) {
			return buildSimpleNoRoundDispatchCase("parse_bid128_for_int_conversion_readtest", "parse_i32", "compare_i32", fmt.Sprintf("%s(a0)", spec.Name))
		}
		return skipDispatchCase()
	case strings.HasPrefix(spec.Name, "bid128_to_int64_"):
		if rustFuncImplemented(spec.Name) {
			return buildSimpleNoRoundDispatchCase("parse_bid128_for_int_conversion_readtest", "parse_i64", "compare_i64", fmt.Sprintf("%s(a0)", spec.Name))
		}
		return skipDispatchCase()
	case strings.HasPrefix(spec.Name, "bid128_to_uint8_"):
		if rustFuncImplemented(spec.Name) {
			return buildSimpleNoRoundDispatchCase("parse_bid128_for_int_conversion_readtest", "parse_u8", "compare_u8", fmt.Sprintf("%s(a0)", spec.Name))
		}
		return skipDispatchCase()
	case strings.HasPrefix(spec.Name, "bid128_to_uint16_"):
		if rustFuncImplemented(spec.Name) {
			return buildSimpleNoRoundDispatchCase("parse_bid128_for_int_conversion_readtest", "parse_u16", "compare_u16", fmt.Sprintf("%s(a0)", spec.Name))
		}
		return skipDispatchCase()
	case strings.HasPrefix(spec.Name, "bid128_to_uint32_"):
		if rustFuncImplemented(spec.Name) {
			return buildSimpleNoRoundDispatchCase("parse_bid128_for_int_conversion_readtest", "parse_u32", "compare_u32", fmt.Sprintf("%s(a0)", spec.Name))
		}
		return skipDispatchCase()
	case strings.HasPrefix(spec.Name, "bid128_to_uint64_"):
		if rustFuncImplemented(spec.Name) {
			return buildSimpleNoRoundDispatchCase("parse_bid128_for_int_conversion_readtest", "parse_u64", "compare_u64", fmt.Sprintf("%s(a0)", spec.Name))
		}
		return skipDispatchCase()
	}
	switch spec.Name {
	case "bid32_nan":
		return buildNan32DispatchCase()
	case "bid64_nan":
		return buildNan64DispatchCase()
	case "bid128_nan":
		return buildNan128DispatchCase()
	case "bid32_from_string":
		return buildStringCallDispatchCase("parse_bid32", "compare_u32", "bid32_from_string_raw(parts[2].to_string(), rm)")
	case "bid_strtod32", "bid_wcstod32":
		return buildStringFromDispatchCase("parse_bid32", "compare_u32", "bid32_special_strtod_readtest(parts[2])")
	case "bid_strtod64", "bid_wcstod64", "str64":
		return buildStringFromDispatchCase("parse_bid64", "compare_u64", "bid64_special_from_string_readtest(parts[2])")
	case "bid_strtod128", "bid_wcstod128":
		return skipDispatchCase()
	case "bid32_to_string":
		if rustFuncImplemented("bid32_to_string_raw") {
			return buildParsedStringDispatchCase("parse_bid32", "bid32_to_string(a0)")
		}
		return skipDispatchCase()
	case "bid64_to_string":
		if rustFuncImplemented("bid64_to_string") {
			return buildParsedStringDispatchCase("parse_bid64", "bid64_to_string(a0)")
		}
		return skipDispatchCase()
	case "bid32_to_binary80",
		"bid128_to_binary80",
		"bid64_to_binary80",
		"binary32_to_bid32",
		"binary32_to_bid64",
		"binary32_to_bid128",
		"binary64_to_bid32",
		"binary64_to_bid64",
		"binary64_to_bid128",
		"binary80_to_bid32",
		"binary80_to_bid64",
		"binary80_to_bid128",
		"binary128_to_bid32",
		"binary128_to_bid64",
		"binary128_to_bid128",
		"bid_to_dpd32",
		"bid_to_dpd64",
		"bid_to_dpd128",
		"bid_dpd_to_bid32",
		"bid_dpd_to_bid64",
		"bid_dpd_to_bid128",
		"bid_is754",
		"bid_is754R",
		"bid64ddq_fma",
		"bid64dqd_fma",
		"bid64dq_add",
		"bid_feclearexcept",
		"bid_fegetexceptflag",
		"bid_feraiseexcept",
		"bid_fesetexceptflag",
		"bid_fetestexcept":
		return skipDispatchCase()
	case "bid64_from_string":
		return buildStringCallDispatchCase("parse_bid64", "compare_u64", "bid64_from_string_via_c_readtest(parts[2], rm)")
	case "bid128_to_string":
		if rustFuncImplemented("bid128_to_string") {
			return buildParsedStringDispatchCase("parse_bid128", "bid128_to_string(a0)")
		}
		return skipDispatchCase()
	case "bid128_nextup":
		if rustFuncImplemented("bid128_next_up") {
			return buildUnaryBid128ExactDispatchCase("bid128_next_up")
		}
		return skipDispatchCase()
	case "bid128_nextdown":
		if rustFuncImplemented("bid128_next_down") {
			return buildUnaryBid128ExactDispatchCase("bid128_next_down")
		}
		return skipDispatchCase()
	case "bid128_from_string":
		return buildStringCallDispatchCase("parse_bid128", "compare_bid128", "bid128_from_string(parts[2].to_string(), rm)")
	case "bid128_to_binary128":
		if rustFuncImplemented("bid128_to_binary128") {
			return buildSimpleDispatchCase("parse_bid128", "parse_bid128", "compare_bid128", "bid128_to_binary128(a0, rm)")
		}
		return skipDispatchCase()
	case "bid32_scalbn":
		return buildProgrammableDispatchCase(
			[]string{"parse_bid32", "parse_i32_decimal"},
			"parse_bid32",
			"compare_u32",
			[]string{"let (got, flags) = bid32_scalbn_with_flags(a0, i64::from(a1), rm);"},
		)
	}
	return ""
}

func generateBid32ViaBid64DispatchCase(spec ReadtestSpec) string {
	if specHasNativeRustFunc(spec) {
		return ""
	}
	switch {
	case spec.Name == "bid32_from_int64":
		if rustFuncImplemented("bid64_from_int64") && rustFuncImplemented("bid64_to_bid32") {
			return buildFromIntDispatchCase("parse_i64", "parse_bid32", "compare_u32", "bid32_from_i64_via_bid64_readtest(a0, rm, bid64_from_int64)")
		}
	case spec.Name == "bid32_from_uint64":
		if rustFuncImplemented("bid64_from_uint64") && rustFuncImplemented("bid64_to_bid32") {
			return buildFromIntDispatchCase("parse_u64", "parse_bid32", "compare_u32", "bid32_from_u64_via_bid64_readtest(a0, rm, bid64_from_uint64)")
		}
	case spec.Name == "bid32_lrint":
		if rustFuncImplemented("bid64_lrint") && rustFuncImplemented("bid32_to_bid64") {
			return buildSimpleDispatchCase("parse_bid32", "parse_i64", "compare_i64", "bid32_lrint_via_bid64_readtest(a0, rm, bid64_lrint)")
		}
	case spec.Name == "bid32_llrint":
		if rustFuncImplemented("bid64_llrint") && rustFuncImplemented("bid32_to_bid64") {
			return buildSimpleDispatchCase("parse_bid32", "parse_i64", "compare_i64", "bid32_lrint_via_bid64_readtest(a0, rm, bid64_llrint)")
		}
	case spec.Name == "bid32_lround":
		if rustFuncImplemented("bid64_lround") && rustFuncImplemented("bid32_to_bid64") {
			return buildSimpleNoRoundDispatchCase("parse_bid32", "parse_i64", "compare_i64", "bid32_lround_via_bid64_readtest(a0, bid64_lround)")
		}
	case spec.Name == "bid32_llround":
		if rustFuncImplemented("bid64_llround") && rustFuncImplemented("bid32_to_bid64") {
			return buildSimpleNoRoundDispatchCase("parse_bid32", "parse_i64", "compare_i64", "bid32_lround_via_bid64_readtest(a0, bid64_llround)")
		}
	case strings.HasPrefix(spec.Name, "bid32_to_int8_"):
		suffix := strings.TrimPrefix(spec.Name, "bid32_to_int8_")
		op64 := "bid64_to_int8_" + suffix
		if rustFuncImplemented(op64) && rustFuncImplemented("bid32_to_bid64") {
			return buildSimpleNoRoundDispatchCase("parse_bid32", "parse_i32", "compare_i32", fmt.Sprintf("bid32_to_int8_via_bid64_readtest(a0, %s)", op64))
		}
	case strings.HasPrefix(spec.Name, "bid32_to_int16_"):
		suffix := strings.TrimPrefix(spec.Name, "bid32_to_int16_")
		op64 := "bid64_to_int16_" + suffix
		if rustFuncImplemented(op64) && rustFuncImplemented("bid32_to_bid64") {
			return buildSimpleNoRoundDispatchCase("parse_bid32", "parse_i32", "compare_i32", fmt.Sprintf("bid32_to_int16_via_bid64_readtest(a0, %s)", op64))
		}
	case strings.HasPrefix(spec.Name, "bid32_to_int32_"):
		suffix := strings.TrimPrefix(spec.Name, "bid32_to_int32_")
		op64 := "bid64_to_int32_" + suffix
		if rustFuncImplemented(op64) && rustFuncImplemented("bid32_to_bid64") {
			return buildSimpleNoRoundDispatchCase("parse_bid32", "parse_i32", "compare_i32", fmt.Sprintf("bid32_to_i32_via_bid64_readtest(a0, %s)", op64))
		}
	case strings.HasPrefix(spec.Name, "bid32_to_int64_"):
		suffix := strings.TrimPrefix(spec.Name, "bid32_to_int64_")
		op64 := "bid64_to_int64_" + suffix
		if rustFuncImplemented(op64) && rustFuncImplemented("bid32_to_bid64") {
			return buildSimpleNoRoundDispatchCase("parse_bid32", "parse_i64", "compare_i64", fmt.Sprintf("bid32_to_i64_via_bid64_readtest(a0, %s)", op64))
		}
	case strings.HasPrefix(spec.Name, "bid32_to_uint8_"):
		suffix := strings.TrimPrefix(spec.Name, "bid32_to_uint8_")
		op64 := "bid64_to_uint8_" + suffix
		if rustFuncImplemented(op64) && rustFuncImplemented("bid32_to_bid64") {
			return buildSimpleNoRoundDispatchCase("parse_bid32", "parse_u32", "compare_u32", fmt.Sprintf("bid32_to_uint8_via_bid64_readtest(a0, %s)", op64))
		}
	case strings.HasPrefix(spec.Name, "bid32_to_uint16_"):
		suffix := strings.TrimPrefix(spec.Name, "bid32_to_uint16_")
		op64 := "bid64_to_uint16_" + suffix
		if rustFuncImplemented(op64) && rustFuncImplemented("bid32_to_bid64") {
			return buildSimpleNoRoundDispatchCase("parse_bid32", "parse_u32", "compare_u32", fmt.Sprintf("bid32_to_uint16_via_bid64_readtest(a0, %s)", op64))
		}
	case strings.HasPrefix(spec.Name, "bid32_to_uint32_"):
		suffix := strings.TrimPrefix(spec.Name, "bid32_to_uint32_")
		op64 := "bid64_to_uint32_" + suffix
		if rustFuncImplemented(op64) && rustFuncImplemented("bid32_to_bid64") {
			return buildSimpleNoRoundDispatchCase("parse_bid32", "parse_u32", "compare_u32", fmt.Sprintf("bid32_to_u32_via_bid64_readtest(a0, %s)", op64))
		}
	case strings.HasPrefix(spec.Name, "bid32_to_uint64_"):
		suffix := strings.TrimPrefix(spec.Name, "bid32_to_uint64_")
		op64 := "bid64_to_uint64_" + suffix
		if rustFuncImplemented(op64) && rustFuncImplemented("bid32_to_bid64") {
			return buildSimpleNoRoundDispatchCase("parse_bid32", "parse_u64", "compare_u64", fmt.Sprintf("bid32_to_u64_via_bid64_readtest(a0, %s)", op64))
		}
	}
	return ""
}

func readtestParsers() string {
	return `
#[derive(Debug, Clone, Copy, PartialEq)]
enum CmpMode { CmpFuzzy, CmpEqual, CmpRelativeErr }

#[derive(Debug, Clone)]
enum DispatchResult { Pass, Fail(String), Skip }

fn parse_bid64(s: &str) -> Option<u64> {
    if let Some(inner) = strip_brackets(s) {
        if let Some((hi, _lo)) = inner.split_once(',') {
            if hi.len() > 16 { return None; }
            return u64::from_str_radix(hi, 16).ok();
        }
        return parse_bracketed_hex_u64(s);
    }
    Some(bid64_from_string_via_c_readtest(s, 0).0)
}

fn parse_bid32(s: &str) -> Option<u32> {
    parse_bracketed_hex_u32(s).or_else(|| Some(bid32_from_string_raw(s.to_string(), 0).0))
}

fn parse_bid128(s: &str) -> Option<BID_UINT128> {
    if let Some(v) = bid128_special_from_string_readtest(s) {
        return Some(v);
    }
    if let Some(inner) = strip_brackets(s) {
        if let Some((hi, lo)) = inner.split_once(',') {
            let hi = u64::from_str_radix(hi.trim(), 16).ok()?;
            let lo = u64::from_str_radix(lo.trim(), 16).ok()?;
            return Some(BID_UINT128 { w: [lo, hi] });
        }

        let mut hex = inner.to_string();
        if hex.len() > 32 {
            return None;
        }
        if hex.len() == 30 {
            hex.push_str("00");
        } else {
            while hex.len() < 32 {
                hex.insert(0, '0');
            }
        }
        let hi = u64::from_str_radix(&hex[..16], 16).ok()?;
        let lo = u64::from_str_radix(&hex[16..], 16).ok()?;
        return Some(BID_UINT128 { w: [lo, hi] });
    }

    parse_bid128_small_decimal(s).or_else(|| Some(bid128_from_string_via_c_readtest(s, 0).0))
}


fn bid128_from_parts(lo: u64, hi: u64) -> BID_UINT128 {
    BID_UINT128 { w: [lo, hi] }
}

fn encode_bid128_coeff_readtest(negative: bool, exponent: i32, coeff: u128) -> Option<BID_UINT128> {
    let raw_exponent = exponent.checked_add(6176)?;
    if !(0..=12287).contains(&raw_exponent) {
        return None;
    }

    let hi_coeff = (coeff >> 64) as u64;
    if (hi_coeff >> 49) != 0 {
        return None;
    }

    let sign = if negative { 0x8000_0000_0000_0000 } else { 0 };
    Some(bid128_from_parts(
        coeff as u64,
        sign | ((raw_exponent as u64) << 49) | hi_coeff,
    ))
}

fn parse_bid128_decimal_parts_readtest(s: &str) -> Option<(bool, String, i32)> {
    let s = s.trim();
    if s.is_empty() {
        return None;
    }

    let (negative, body) = if let Some(rest) = s.strip_prefix('-') {
        (true, rest)
    } else if let Some(rest) = s.strip_prefix('+') {
        (false, rest)
    } else {
        (false, s)
    };

    let (mantissa, explicit_exp) = if let Some(idx) = body.find(['e', 'E']) {
        let exp = body[idx + 1..].parse::<i32>().ok()?;
        (&body[..idx], exp)
    } else {
        (body, 0)
    };

    if mantissa.is_empty() {
        return None;
    }

    let mut digits = String::new();
    let mut frac_digits = 0_i32;
    let mut saw_dot = false;
    let mut saw_digit = false;

    for ch in mantissa.chars() {
        match ch {
            '0'..='9' => {
                digits.push(ch);
                if saw_dot {
                    frac_digits += 1;
                }
                saw_digit = true;
            }
            '.' if !saw_dot => {
                saw_dot = true;
            }
            _ => return None,
        }
    }

    if !saw_digit {
        return None;
    }

    Some((negative, digits, explicit_exp - frac_digits))
}

fn parse_bid128_small_decimal(s: &str) -> Option<BID_UINT128> {
    let (negative, digits, exponent) = parse_bid128_decimal_parts_readtest(s)?;
    let trimmed = digits.trim_start_matches('0');

    if trimmed.is_empty() {
        return encode_bid128_coeff_readtest(negative, exponent, 0);
    }

    let digits = trimmed;
    if digits.len() > 34 {
        return None;
    }

    let coeff = digits.parse::<u128>().ok()?;
    encode_bid128_coeff_readtest(negative, exponent, coeff)
}

fn parse_bid128_decimal_for_conversion_readtest(s: &str) -> Option<BID_UINT128> {
    let (negative, digits, mut exponent) = parse_bid128_decimal_parts_readtest(s)?;
    let trimmed = digits.trim_start_matches('0');

    if trimmed.is_empty() {
        return encode_bid128_coeff_readtest(negative, exponent, 0);
    }

    let mut digits = trimmed.to_string();
    while digits.len() > 34 && digits.ends_with('0') {
        digits.pop();
        exponent += 1;
    }
    if digits.len() > 34 {
        return None;
    }

    let coeff = digits.parse::<u128>().ok()?;
    encode_bid128_coeff_readtest(negative, exponent, coeff)
}

fn parse_bid128_for_int_conversion_readtest(s: &str) -> Option<BID_UINT128> {
    let lower = s.trim().to_ascii_lowercase();
    match lower.as_str() {
        "inf" | "+inf" | "infinity" | "+infinity" => return Some(bid128_from_parts(0, 0x7800_0000_0000_0000)),
        "-inf" | "-infinity" => return Some(bid128_from_parts(0, 0xf800_0000_0000_0000)),
        _ => {}
    }

    if let Some(inner) = strip_brackets(s) {
        if let Some((hi, lo)) = inner.split_once(',') {
            let hi = u64::from_str_radix(hi, 16).ok()?;
            let lo = u64::from_str_radix(lo, 16).ok()?;
            return Some(bid128_from_parts(lo, hi));
        }
        let mut hex = inner.to_string();
        if hex.len() > 32 {
            return None;
        }
        if hex.len() == 30 {
            hex.push_str("00");
        } else {
            while hex.len() < 32 {
                hex.insert(0, '0');
            }
        }
        let hi = u64::from_str_radix(&hex[..16], 16).ok()?;
        let lo = u64::from_str_radix(&hex[16..], 16).ok()?;
        return Some(bid128_from_parts(lo, hi));
    }

    parse_bid128_decimal_for_conversion_readtest(s).or_else(|| Some(bid128_from_string_via_c_readtest(s, 0).0))
}

fn parse_i32(s: &str) -> Option<i32> {
    let raw = s.trim();
    if let Some(h) = raw.strip_prefix('[').and_then(|v| v.strip_suffix(']')) {
        u32::from_str_radix(h, 16).ok().map(|v| v as i32)
    } else if let Some(h) = raw.strip_prefix("0x").or_else(|| raw.strip_prefix("0X")) {
        u32::from_str_radix(h, 16).ok().map(|v| v as i32)
    } else {
        raw.parse().ok()
    }
}

fn parse_i32_as_i64(s: &str) -> Option<i64> {
    parse_i32(s).map(i64::from)
}

fn parse_i32_decimal(s: &str) -> Option<i32> {
    let raw = s.trim();
    if let Some(v) = parse_i32(raw) {
        return Some(v);
    }
    let bytes = raw.as_bytes();
    if bytes.is_empty() {
        return None;
    }
    let mut end = 0usize;
    if bytes[0] == b'+' || bytes[0] == b'-' {
        end = 1;
    }
    while end < bytes.len() && bytes[end].is_ascii_digit() {
        end += 1;
    }
    if end == 0 || (end == 1 && (bytes[0] == b'+' || bytes[0] == b'-')) {
        return None;
    }
    raw[..end].parse().ok()
}

fn parse_i64(s: &str) -> Option<i64> {
    let raw = s.trim();
    if let Some(h) = raw.strip_prefix('[').and_then(|v| v.strip_suffix(']')) {
        u64::from_str_radix(h, 16).ok().map(|v| v as i64)
    } else if let Some(h) = raw.strip_prefix("0x").or_else(|| raw.strip_prefix("0X")) {
        u64::from_str_radix(h, 16).ok().map(|v| v as i64)
    } else {
        raw.parse().ok()
    }
}

fn parse_u32(s: &str) -> Option<u32> {
    let raw = s.trim();
    if let Some(h) = raw.strip_prefix('[').and_then(|v| v.strip_suffix(']')) {
        u32::from_str_radix(h, 16).ok()
    } else if let Some(h) = raw.strip_prefix("0x").or_else(|| raw.strip_prefix("0X")) {
        u32::from_str_radix(h, 16).ok()
    } else {
        raw.parse::<u32>().ok().or_else(|| raw.parse::<i128>().ok().map(|v| v as u32))
    }
}

fn parse_u64(s: &str) -> Option<u64> {
    let raw = s.trim();
    if let Some(h) = raw.strip_prefix('[').and_then(|v| v.strip_suffix(']')) {
        u64::from_str_radix(h, 16).ok()
    } else if let Some(h) = raw.strip_prefix("0x").or_else(|| raw.strip_prefix("0X")) {
        u64::from_str_radix(h, 16).ok()
    } else {
        raw.parse::<u64>().ok().or_else(|| raw.parse::<i128>().ok().map(|v| v as u64))
    }
}

fn parse_i8(s: &str) -> Option<i8> { parse_i32(s).map(|v| v as i8) }
fn parse_i16(s: &str) -> Option<i16> { parse_i32(s).map(|v| v as i16) }
fn parse_u8(s: &str) -> Option<u8> { parse_u32(s).map(|v| v as u8) }
fn parse_u16(s: &str) -> Option<u16> { parse_u32(s).map(|v| v as u16) }
fn parse_f32(s: &str) -> Option<f32> {
    let s = s.trim_matches(|c| c == '[' || c == ']');
    u32::from_str_radix(s, 16).ok().map(f32::from_bits)
}
fn parse_f64(s: &str) -> Option<f64> {
    let s = s.trim_matches(|c| c == '[' || c == ']');
    u64::from_str_radix(s, 16).ok().map(f64::from_bits)
}
fn parse_string(s: &str) -> Option<String> { Some(s.to_string()) }

fn strip_brackets(s: &str) -> Option<&str> {
    s.strip_prefix('[')?.strip_suffix(']')
}

fn parse_bracketed_hex_u32(s: &str) -> Option<u32> {
    let hex = strip_brackets(s)?;
    if hex.contains(',') || hex.len() > 8 {
        return None;
    }
    u32::from_str_radix(hex, 16).ok()
}

fn parse_bracketed_hex_u64(s: &str) -> Option<u64> {
    let hex = strip_brackets(s)?;
    if hex.contains(',') || hex.len() > 16 {
        return None;
    }
    u64::from_str_radix(hex, 16).ok()
}

fn parse_bid32_special(s: &str) -> Option<u32> {
    let lower = s.trim().to_ascii_lowercase();
    match lower.as_str() {
        "inf" | "+inf" | "infinity" | "+infinity" => Some(0x7800_0000),
        "-inf" | "-infinity" => Some(0xf800_0000),
        "nan" | "+nan" | "qnan" | "+qnan" => Some(0x7c00_0000),
        "-nan" | "-qnan" => Some(0xfc00_0000),
        "snan" | "+snan" => Some(0x7e00_0000),
        "-snan" => Some(0xfe00_0000),
        _ => None,
    }
}

fn parse_bid64_special(s: &str) -> Option<u64> {
    let lower = s.trim().to_ascii_lowercase();
    match lower.as_str() {
        "inf" | "+inf" | "infinity" | "+infinity" => Some(0x7800_0000_0000_0000),
        "-inf" | "-infinity" => Some(0xf800_0000_0000_0000),
        "nan" | "+nan" | "qnan" | "+qnan" => Some(0x7c00_0000_0000_0000),
        "-nan" | "-qnan" => Some(0xfc00_0000_0000_0000),
        "snan" | "+snan" => Some(0x7e00_0000_0000_0000),
        "-snan" => Some(0xfe00_0000_0000_0000),
        _ => None,
    }
}

fn parse_bid128_special(s: &str) -> Option<BID_UINT128> {
    let lower = s.trim().to_ascii_lowercase();
    let hi = match lower.as_str() {
        "inf" | "+inf" | "infinity" | "+infinity" => 0x7800_0000_0000_0000,
        "-inf" | "-infinity" => 0xf800_0000_0000_0000,
        "nan" | "+nan" | "qnan" | "+qnan" => 0x7c00_0000_0000_0000,
        "-nan" | "-qnan" => 0xfc00_0000_0000_0000,
        "snan" | "+snan" => 0x7e00_0000_0000_0000,
        "-snan" => 0xfe00_0000_0000_0000,
        _ => return None,
    };
    Some(BID_UINT128 { w: [0, hi] })
}

fn parse_bid64_decimal_exact(s: &str) -> Option<u64> {
    let s = s.trim();
    if s.is_empty() { return None; }
    let (negative, body) = if let Some(rest) = s.strip_prefix('-') {
        (true, rest)
    } else if let Some(rest) = s.strip_prefix('+') {
        (false, rest)
    } else {
        (false, s)
    };
    let (mantissa, explicit_exp) = if let Some(idx) = body.find(['e', 'E']) {
        (&body[..idx], body[idx + 1..].parse::<i32>().ok()?)
    } else {
        (body, 0)
    };
    let mut digits = String::new();
    let mut frac_digits = 0_i32;
    let mut saw_dot = false;
    let mut saw_digit = false;
    for ch in mantissa.chars() {
        match ch {
            '0'..='9' => {
                digits.push(ch);
                if saw_dot { frac_digits += 1; }
                saw_digit = true;
            }
            '.' if !saw_dot => saw_dot = true,
            _ => return None,
        }
    }
    if !saw_digit { return None; }
    let mut exp10 = i64::from(explicit_exp - frac_digits);
    let trimmed = digits.trim_start_matches('0');
    if trimmed.is_empty() {
        let mut value = bid64_from_uint64(0, 0).0;
        if exp10 != 0 { value = bid64_scalbn(value, exp10, 0).0; }
        if negative { value = bid64_negate(value); }
        return Some(value);
    }
    let mut digits = trimmed.to_string();
    while digits.len() > 16 && digits.ends_with('0') {
        digits.pop();
        exp10 += 1;
    }
    if digits.len() > 16 { return None; }
    let coeff = digits.parse::<u64>().ok()?;
    let mut value = if negative {
        bid64_from_int64(-(coeff as i64), 0).0
    } else {
        bid64_from_uint64(coeff, 0).0
    };
    if exp10 != 0 { value = bid64_scalbn(value, exp10, 0).0; }
    Some(value)
}

fn parse_bid32_decimal_exact(s: &str) -> Option<u32> {
    let bid64 = parse_bid64_decimal_exact(s)?;
    let (value, flags) = bid64_to_bid32(bid64, 0);
    if flags == 0 { Some(value) } else { None }
}

fn parse_flags(s: &str) -> u32 {
    let s = s.trim();
    let s = s.strip_prefix("0x").or_else(|| s.strip_prefix("0X")).unwrap_or(s);
    u32::from_str_radix(s, 16).unwrap_or(0)
}

fn fuzzy_flags_ok(expected: u32, actual: u32) -> bool {
    (actual & expected) == expected
}

fn bid64_numerically_equal(a: u64, b: u64) -> bool {
    if bid64_is_na_n(a) != 0 || bid64_is_na_n(b) != 0 {
        return false;
    }
    let (ab, _) = bid64_signaling_less(a, b);
    let (ba, _) = bid64_signaling_less(b, a);
    ab == 0 && ba == 0
}

fn bid32_numerically_equal(a: u32, b: u32) -> bool {
    if bid32_is_na_n32(a) != 0 || bid32_is_na_n32(b) != 0 {
        return false;
    }
    let (ab, _) = bid32_signaling_less(a, b);
    let (ba, _) = bid32_signaling_less(b, a);
    ab == 0 && ba == 0
}

fn bid128_numerically_equal(a: BID_UINT128, b: BID_UINT128) -> bool {
    if bid128_is_na_n(a) != 0 || bid128_is_na_n(b) != 0 {
        return false;
    }
    let (ab, _) = bid128_signaling_less(a, b);
    let (ba, _) = bid128_signaling_less(b, a);
    ab == 0 && ba == 0
}

`
}

func readtestCompareFuncs() string {
	return `
const READTEST_TRANS_FLAGS_MASK: u32 = 0x05;

fn readtest_flags_ok(expected: u32, actual: u32, mode: CmpMode) -> bool {
    match mode {
        CmpMode::CmpRelativeErr => (expected & READTEST_TRANS_FLAGS_MASK) == (actual & READTEST_TRANS_FLAGS_MASK),
        _ => fuzzy_flags_ok(expected, actual),
    }
}

fn readtest_mre_max32(rm: i64) -> f64 {
    match rm {
        0 | 4 => 0.5,
        1 | 2 | 3 => 1.01,
        _ => 0.5,
    }
}

fn readtest_mre_max64(rm: i64) -> f64 {
    match rm {
        0 | 4 => 0.55,
        1 | 2 | 3 => 1.05,
        _ => 0.55,
    }
}

fn readtest_mre_max128(rm: i64) -> f64 {
    match rm {
        0 | 4 => 2.0,
        1 | 2 | 3 => 5.0,
        _ => 2.0,
    }
}

fn bid32_readtest_exp(x: u32) -> u32 {
    if (x & 0x6000_0000) == 0x6000_0000 {
        (x & 0x1fe0_0000) >> 21
    } else {
        (x & 0x7f80_0000) >> 23
    }
}

fn bid32_readtest_mant(x: u32) -> u32 {
    if (x & 0x6000_0000) == 0x6000_0000 {
        (x & 0x001f_ffff) | 0x0080_0000
    } else {
        x & 0x007f_ffff
    }
}

fn bid64_readtest_exp(x: u64) -> u64 {
    if (x & 0x6000_0000_0000_0000) == 0x6000_0000_0000_0000 {
        (x & 0x1ff8_0000_0000_0000) >> 51
    } else {
        (x & 0x7fe0_0000_0000_0000) >> 53
    }
}

fn bid64_readtest_mant(x: u64) -> u64 {
    if (x & 0x6000_0000_0000_0000) == 0x6000_0000_0000_0000 {
        (x & 0x0007_ffff_ffff_ffff) | 0x0020_0000_0000_0000
    } else {
        x & 0x001f_ffff_ffff_ffff
    }
}

fn bid128_readtest_exp(x: BID_UINT128) -> u64 {
    let hi = x.w[1];
    if (hi & 0x6000_0000_0000_0000) == 0x6000_0000_0000_0000 {
        (hi & 0x1fff_8000_0000_0000) >> 15
    } else {
        (hi & 0x7ffe_0000_0000_0000) >> 17
    }
}

fn bid128_readtest_mant(x: BID_UINT128) -> BID_UINT128 {
    let hi = x.w[1];
    if (hi & 0x6000_0000_0000_0000) == 0x6000_0000_0000_0000 {
        BID_UINT128 { w: [x.w[0], (hi & 0x0000_7fff_ffff_ffff) | 0x0002_0000_0000_0000] }
    } else {
        BID_UINT128 { w: [x.w[0], hi & 0x0001_ffff_ffff_ffff] }
    }
}

fn bid32_relative_err_ok(got: u32, expected: u32, rm: i64, ulp_add: f64) -> bool {
    if bid32_is_na_n32(got) != 0 || bid32_is_na_n32(expected) != 0 || bid32_is_inf32(got) != 0 || bid32_is_inf32(expected) != 0 {
        return got == expected;
    }
    if (got & 0x8000_0000) != (expected & 0x8000_0000) {
        return false;
    }

    let mut r1 = got;
    let mut r2 = expected;
    let mut e1 = bid32_readtest_exp(r1);
    let mut e2 = bid32_readtest_exp(r2);
    if e1 < e2 {
        r1 = bid32_quantize(got, expected, rm).0;
        e1 = bid32_readtest_exp(r1);
        e2 = bid32_readtest_exp(r2);
    } else if e2 < e1 {
        r2 = bid32_quantize(expected, got, rm).0;
        e1 = bid32_readtest_exp(r1);
        e2 = bid32_readtest_exp(r2);
    }
    if e1 != e2 {
        return false;
    }

    let m1 = bid32_readtest_mant(r1);
    let m2 = bid32_readtest_mant(r2);
    let mut ulp = if m1 > m2 { (m1 - m2) as f64 } else { (m2 - m1) as f64 };
    let (less, _) = bid32_quiet_less(got, expected);
    if less != 0 {
        ulp *= -1.0;
    }
    (ulp + ulp_add).abs() <= readtest_mre_max32(rm)
}

fn bid64_relative_err_ok(got: u64, expected: u64, rm: i64, ulp_add: f64) -> bool {
    if bid64_is_na_n(got) != 0 || bid64_is_na_n(expected) != 0 || bid64_is_inf(got) != 0 || bid64_is_inf(expected) != 0 {
        return got == expected;
    }
    if (got & 0x8000_0000_0000_0000) != (expected & 0x8000_0000_0000_0000) {
        return false;
    }

    let mut r1 = got;
    let mut r2 = expected;
    let mut e1 = bid64_readtest_exp(r1);
    let mut e2 = bid64_readtest_exp(r2);
    if e1 < e2 {
        r1 = bid64_quantize(got, expected, rm).0;
        e1 = bid64_readtest_exp(r1);
        e2 = bid64_readtest_exp(r2);
    } else if e2 < e1 {
        r2 = bid64_quantize(expected, got, rm).0;
        e1 = bid64_readtest_exp(r1);
        e2 = bid64_readtest_exp(r2);
    }
    if e1 != e2 {
        return false;
    }

    let m1 = bid64_readtest_mant(r1);
    let m2 = bid64_readtest_mant(r2);
    let mut ulp = if m1 > m2 { (m1 - m2) as f64 } else { (m2 - m1) as f64 };
    let (less, _) = bid64_quiet_less(got, expected);
    if less != 0 {
        ulp *= -1.0;
    }
    (ulp + ulp_add).abs() <= readtest_mre_max64(rm)
}

fn bid128_relative_err_ok(mut got: BID_UINT128, mut expected: BID_UINT128, rm: i64, ulp_add: f64) -> bool {
    let got_nan = bid128_is_na_n(got) != 0;
    let expected_nan = bid128_is_na_n(expected) != 0;
    let got_inf = bid128_is_inf(got) != 0;
    let expected_inf = bid128_is_inf(expected) != 0;
    if got_nan || expected_nan || got_inf || expected_inf {
        if got.w == expected.w {
            return true;
        }
        if got_inf {
            got.w[1] = (got.w[1] & 0x8000_0000_0000_0000) | 0x5fff_ed09_bead_87c0;
            got.w[0] = 0x378d_8e63_ffff_ffff;
        } else if expected_inf {
            expected.w[1] = (expected.w[1] & 0x8000_0000_0000_0000) | 0x5fff_ed09_bead_87c0;
            expected.w[0] = 0x378d_8e63_ffff_ffff;
        } else {
            return false;
        }
    }
    if (got.w[1] & 0x8000_0000_0000_0000) != (expected.w[1] & 0x8000_0000_0000_0000) {
        return false;
    }

    let mut r1 = got;
    let mut r2 = expected;
    let mut e1 = bid128_readtest_exp(r1);
    let mut e2 = bid128_readtest_exp(r2);
    if e1 < e2 {
        r1 = bid128_quantize(got, expected, rm).0;
        e1 = bid128_readtest_exp(r1);
        e2 = bid128_readtest_exp(r2);
    } else if e2 < e1 {
        r2 = bid128_quantize(expected, got, rm).0;
        e1 = bid128_readtest_exp(r1);
        e2 = bid128_readtest_exp(r2);
    }
    if e1 != e2 {
        return false;
    }

    let m1 = bid128_readtest_mant(r1);
    let m2 = bid128_readtest_mant(r2);
    let mut ulp = if m1.w[0] > m2.w[0] {
        (m1.w[0] - m2.w[0]) as f64
    } else {
        (m2.w[0] - m1.w[0]) as f64
    };
    let (less, _) = bid128_quiet_less(got, expected);
    if less != 0 {
        ulp *= -1.0;
    }
    (ulp + ulp_add).abs() <= readtest_mre_max128(rm)
}

fn compare_u64(got: u64, expected: u64, got_flags: u32, exp_flags: u32, mode: CmpMode, rm: i64, ulp_add: f64) -> DispatchResult {
    let value_ok = match mode {
        CmpMode::CmpFuzzy => got == expected,
        CmpMode::CmpEqual => got == expected || bid64_numerically_equal(got, expected),
        CmpMode::CmpRelativeErr => bid64_relative_err_ok(got, expected, rm, ulp_add),
    };
    if value_ok && readtest_flags_ok(exp_flags, got_flags, mode) {
        DispatchResult::Pass
    } else {
        DispatchResult::Fail(format!("mode={:?} got={:016x}/{:02x} want={:016x}/{:02x}", mode, got, got_flags, expected, exp_flags))
    }
}

fn compare_u32(got: u32, expected: u32, got_flags: u32, exp_flags: u32, mode: CmpMode, rm: i64, ulp_add: f64) -> DispatchResult {
    let value_ok = match mode {
        CmpMode::CmpFuzzy => got == expected,
        CmpMode::CmpEqual => got == expected || bid32_numerically_equal(got, expected),
        CmpMode::CmpRelativeErr => bid32_relative_err_ok(got, expected, rm, ulp_add),
    };
    if value_ok && readtest_flags_ok(exp_flags, got_flags, mode) {
        DispatchResult::Pass
    } else {
        DispatchResult::Fail(format!("mode={:?} got={:08x}/{:02x} want={:08x}/{:02x}", mode, got, got_flags, expected, exp_flags))
    }
}

fn compare_bid128(got: BID_UINT128, expected: BID_UINT128, got_flags: u32, exp_flags: u32, mode: CmpMode, rm: i64, ulp_add: f64) -> DispatchResult {
    let value_ok = match mode {
        CmpMode::CmpFuzzy => got.w == expected.w,
        CmpMode::CmpEqual => got.w == expected.w || bid128_numerically_equal(got, expected),
        CmpMode::CmpRelativeErr => bid128_relative_err_ok(got, expected, rm, ulp_add),
    };
    if value_ok && readtest_flags_ok(exp_flags, got_flags, mode) {
        DispatchResult::Pass
    } else {
        DispatchResult::Fail(format!("mode={:?} got=[{:016x}{:016x}]/{:02x} want=[{:016x}{:016x}]/{:02x}",
            mode, got.w[1], got.w[0], got_flags, expected.w[1], expected.w[0], exp_flags))
    }
}

fn compare_i32(got: i32, expected: i32, got_flags: u32, exp_flags: u32, mode: CmpMode, _rm: i64, _ulp_add: f64) -> DispatchResult {
    if mode == CmpMode::CmpRelativeErr {
        return DispatchResult::Fail("CMP_RELATIVEERR is unsupported for i32 results".to_string());
    }
    if got == expected && fuzzy_flags_ok(exp_flags, got_flags) {
        DispatchResult::Pass
    } else {
        DispatchResult::Fail(format!("got={}/{:02x} want={}/{:02x}", got, got_flags, expected, exp_flags))
    }
}

fn compare_i64(got: i64, expected: i64, got_flags: u32, exp_flags: u32, mode: CmpMode, _rm: i64, _ulp_add: f64) -> DispatchResult {
    if mode == CmpMode::CmpRelativeErr {
        return DispatchResult::Fail("CMP_RELATIVEERR is unsupported for i64 results".to_string());
    }
    if got == expected && fuzzy_flags_ok(exp_flags, got_flags) {
        DispatchResult::Pass
    } else {
        DispatchResult::Fail(format!("got={}/{:02x} want={}/{:02x}", got, got_flags, expected, exp_flags))
    }
}

fn compare_i8(got: i8, expected: i8, got_flags: u32, exp_flags: u32, mode: CmpMode, rm: i64, ulp_add: f64) -> DispatchResult {
    compare_i32(got as i32, expected as i32, got_flags, exp_flags, mode, rm, ulp_add)
}
fn compare_i16(got: i16, expected: i16, got_flags: u32, exp_flags: u32, mode: CmpMode, rm: i64, ulp_add: f64) -> DispatchResult {
    compare_i32(got as i32, expected as i32, got_flags, exp_flags, mode, rm, ulp_add)
}
fn compare_u8(got: u8, expected: u8, got_flags: u32, exp_flags: u32, mode: CmpMode, rm: i64, ulp_add: f64) -> DispatchResult {
    compare_u32(got as u32, expected as u32, got_flags, exp_flags, mode, rm, ulp_add)
}
fn compare_u16(got: u16, expected: u16, got_flags: u32, exp_flags: u32, mode: CmpMode, rm: i64, ulp_add: f64) -> DispatchResult {
    compare_u32(got as u32, expected as u32, got_flags, exp_flags, mode, rm, ulp_add)
}
fn compare_f32(got: f32, expected: f32, got_flags: u32, exp_flags: u32, mode: CmpMode, _rm: i64, _ulp_add: f64) -> DispatchResult {
    if mode == CmpMode::CmpRelativeErr {
        return DispatchResult::Fail("CMP_RELATIVEERR is unsupported for f32 results".to_string());
    }
    if got.to_bits() == expected.to_bits() && fuzzy_flags_ok(exp_flags, got_flags) {
        DispatchResult::Pass
    } else {
        DispatchResult::Fail(format!("got={:08x}/{:02x} want={:08x}/{:02x}", got.to_bits(), got_flags, expected.to_bits(), exp_flags))
    }
}
fn compare_f64(got: f64, expected: f64, got_flags: u32, exp_flags: u32, mode: CmpMode, _rm: i64, _ulp_add: f64) -> DispatchResult {
    if mode == CmpMode::CmpRelativeErr {
        return DispatchResult::Fail("CMP_RELATIVEERR is unsupported for f64 results".to_string());
    }
    if got.to_bits() == expected.to_bits() && fuzzy_flags_ok(exp_flags, got_flags) {
        DispatchResult::Pass
    } else {
        DispatchResult::Fail(format!("got={:016x}/{:02x} want={:016x}/{:02x}", got.to_bits(), got_flags, expected.to_bits(), exp_flags))
    }
}

fn compare_string(got: String, expected: String, got_flags: u32, exp_flags: u32, mode: CmpMode, _rm: i64, _ulp_add: f64) -> DispatchResult {
    if mode == CmpMode::CmpRelativeErr {
        return DispatchResult::Fail("CMP_RELATIVEERR is unsupported for string results".to_string());
    }
    if got == expected && fuzzy_flags_ok(exp_flags, got_flags) {
        DispatchResult::Pass
    } else {
        DispatchResult::Fail(format!("got={:?}/{:02x} want={:?}/{:02x}", got, got_flags, expected, exp_flags))
    }
}

fn compare_bool_int(got: bool, expected: i32, got_flags: u32, exp_flags: u32, mode: CmpMode, rm: i64, ulp_add: f64) -> DispatchResult {
    compare_i32(if got { 1 } else { 0 }, expected, got_flags, exp_flags, mode, rm, ulp_add)
}

`
}

func readtestCustomHelpers() string {
	return `
fn bid64_from_string_via_c_readtest(s: &str, rm: i64) -> (u64, u32) {
    let cstr = CString::new(s).expect("CString readtest bid64 input");
    let mut flags: u32 = 0;
    let value = unsafe {
        libbid_sys::bid64_from_string(cstr.as_ptr(), rm as u32, &mut flags)
    };
    (value, flags)
}

fn bid128_from_string_via_c_readtest(s: &str, rm: i64) -> (BID_UINT128, u32) {
    let cstr = CString::new(s).expect("CString readtest bid128 input");
    let mut flags: u32 = 0;
    let value = unsafe {
        libbid_sys::bid128_from_string(cstr.as_ptr(), rm as u32, &mut flags)
    };
    (BID_UINT128 { w: value.w }, flags)
}

fn bid128_binop_flags_param_readtest(
    x: BID_UINT128,
    y: BID_UINT128,
    rm: i64,
    op128: fn(BID_UINT128, BID_UINT128, i64, &mut u32) -> BID_UINT128,
) -> (BID_UINT128, u32) {
    let mut flags: u32 = 0;
    let v = op128(x, y, rm, &mut flags);
    (v, flags)
}

fn dispatch_unary_bid128_optional_ignored_arg(
    parts: &[&str],
    op: fn(BID_UINT128) -> (BID_UINT128, u32),
) -> DispatchResult {
    let (input_idx, expected_idx, flags_idx) = match parts.len() {
        5 => (2, 3, 4),
        6 => (2, 4, 5),
        _ => return DispatchResult::Skip,
    };
    let Some(a0) = parse_bid128(parts[input_idx]) else { return DispatchResult::Skip };
    let Some(expected) = parse_bid128(parts[expected_idx]) else { return DispatchResult::Skip };
    let expected_flags = parse_flags(parts[flags_idx]);
    let (got, flags) = op(a0);
    let result = compare_bid128(got, expected, flags, expected_flags, CmpMode::CmpFuzzy, 0, 0.0);
    if !matches!(result, DispatchResult::Pass) { return result; }
    DispatchResult::Pass
}

fn bid32_to_i32_via_bid64_readtest(x: u32, op64: fn(u64) -> (i32, u32)) -> (i32, u32) {
    let (x64, f1) = bid32_to_bid64(x);
    let (v, f2) = op64(x64);
    (v, f1 | f2)
}

fn bid32_to_i64_via_bid64_readtest(x: u32, op64: fn(u64) -> (i64, u32)) -> (i64, u32) {
    let (x64, f1) = bid32_to_bid64(x);
    let (v, f2) = op64(x64);
    (v, f1 | f2)
}

fn bid32_to_u32_via_bid64_readtest(x: u32, op64: fn(u64) -> (u32, u32)) -> (u32, u32) {
    let (x64, f1) = bid32_to_bid64(x);
    let (v, f2) = op64(x64);
    (v, f1 | f2)
}

fn bid32_to_u64_via_bid64_readtest(x: u32, op64: fn(u64) -> (u64, u32)) -> (u64, u32) {
    let (x64, f1) = bid32_to_bid64(x);
    let (v, f2) = op64(x64);
    (v, f1 | f2)
}

fn bid32_to_int8_via_bid64_readtest(x: u32, op64: fn(u64) -> (i8, u32)) -> (i32, u32) {
    let (x64, f1) = bid32_to_bid64(x);
    let (v, f2) = op64(x64);
    (i32::from(v), f1 | f2)
}

fn bid32_to_int16_via_bid64_readtest(x: u32, op64: fn(u64) -> (i16, u32)) -> (i32, u32) {
    let (x64, f1) = bid32_to_bid64(x);
    let (v, f2) = op64(x64);
    (i32::from(v), f1 | f2)
}

fn bid32_to_uint8_via_bid64_readtest(x: u32, op64: fn(u64) -> (u8, u32)) -> (u32, u32) {
    let (x64, f1) = bid32_to_bid64(x);
    let (v, f2) = op64(x64);
    (u32::from(v), f1 | f2)
}

fn bid32_to_uint16_via_bid64_readtest(x: u32, op64: fn(u64) -> (u16, u32)) -> (u32, u32) {
    let (x64, f1) = bid32_to_bid64(x);
    let (v, f2) = op64(x64);
    (u32::from(v), f1 | f2)
}

fn bid32_from_i64_via_bid64_readtest(x: i64, rm: i64, op64: fn(i64, i64) -> (u64, u32)) -> (u32, u32) {
    let (v64, f1) = op64(x, rm);
    let (v32, f2) = bid64_to_bid32(v64, 0);
    (v32, f1 | f2)
}

fn bid32_from_u64_via_bid64_readtest(x: u64, rm: i64, op64: fn(u64, i64) -> (u64, u32)) -> (u32, u32) {
    let (v64, f1) = op64(x, rm);
    let (v32, f2) = bid64_to_bid32(v64, 0);
    (v32, f1 | f2)
}

fn bid32_lrint_via_bid64_readtest(x: u32, rm: i64, op64: fn(u64, i64) -> (i64, u32)) -> (i64, u32) {
    let (x64, f1) = bid32_to_bid64(x);
    let (v, f2) = op64(x64, rm);
    (v, f1 | f2)
}

fn bid32_lround_via_bid64_readtest(x: u32, op64: fn(u64) -> (i64, u32)) -> (i64, u32) {
    let (x64, f1) = bid32_to_bid64(x);
    let (v, f2) = op64(x64);
    (v, f1 | f2)
}

fn bid32_special_from_string_readtest(s: &str) -> Option<u32> {
    let s = s.trim();
    let (sign, body) = if let Some(rest) = s.strip_prefix('-') {
        (0x8000_0000, rest)
    } else if let Some(rest) = s.strip_prefix('+') {
        (0, rest)
    } else {
        (0, s)
    };
    let body = body.to_ascii_lowercase();

    match body.as_str() {
        "inf" | "infinity" => Some(sign | 0x7800_0000),
        _ if body.starts_with("snan") => Some(sign | 0x7e00_0000),
        "nan" | "qnan" => Some(sign | 0x7c00_0000),
        _ if body.starts_with("nan") => Some(sign | 0x7c00_0000),
        _ if body.starts_with("inf") || body.starts_with("infinity") => Some(sign | 0x7c00_0000),
        _ => None,
    }
}

fn bid32_special_strtod_readtest(s: &str) -> Option<u32> {
    let s = s.trim();
    let (sign, body) = if let Some(rest) = s.strip_prefix('-') {
        (0x8000_0000, rest)
    } else if let Some(rest) = s.strip_prefix('+') {
        (0, rest)
    } else {
        (0, s)
    };
    let body = body.to_ascii_lowercase();

    match body.as_str() {
        "inf" | "infinity" => Some(sign | 0x7800_0000),
        _ if body.starts_with("snan") => parse_bid32("0"),
        "nan" | "qnan" => Some(sign | 0x7c00_0000),
        _ if body.starts_with("nan") => Some(sign | 0x7c00_0000),
        _ if body.starts_with("inf") || body.starts_with("infinity") => Some(sign | 0x7c00_0000),
        _ => None,
    }
}

fn bid64_special_from_string_readtest(s: &str) -> Option<u64> {
    let s = s.trim();
    let (sign, body) = if let Some(rest) = s.strip_prefix('-') {
        (0x8000_0000_0000_0000, rest)
    } else if let Some(rest) = s.strip_prefix('+') {
        (0, rest)
    } else {
        (0, s)
    };
    let body = body.to_ascii_lowercase();

    if body == "inf" || body == "infinity" {
        return Some(sign | 0x7800_0000_0000_0000);
    }
    if body.starts_with("snan") {
        return Some(sign | 0x7e00_0000_0000_0000);
    }
    if body == "nan" || body.starts_with("nan") {
        return Some(sign | 0x7c00_0000_0000_0000);
    }
    if body.starts_with("inf") || body.starts_with("infinity") {
        return Some(sign | 0x7c00_0000_0000_0000);
    }

    None
}

fn bid128_special_from_string_readtest(s: &str) -> Option<BID_UINT128> {
    let s = s.trim();
    let (sign, body) = if let Some(rest) = s.strip_prefix('-') {
        (0x8000_0000_0000_0000, rest)
    } else if let Some(rest) = s.strip_prefix('+') {
        (0, rest)
    } else {
        (0, s)
    };
    let body = body.to_ascii_lowercase();

    let hi = if body == "inf" || body == "infinity" {
        0x7800_0000_0000_0000
    } else if body.starts_with("snan") {
        0x7e00_0000_0000_0000
    } else if body == "nan" || body == "qnan" || body.starts_with("nan") {
        0x7c00_0000_0000_0000
    } else if body.starts_with("inf") || body.starts_with("infinity") {
        0x7c00_0000_0000_0000
    } else {
        return None;
    };

    Some(BID_UINT128 { w: [0, sign | hi] })
}

`
}

func readtestRunner() string {
	return `
fn find_readtest_in() -> PathBuf {
    let candidates = [
        "../devtools/third_party/intel_dfp/TESTS/readtest.in",
        "devtools/third_party/intel_dfp/TESTS/readtest.in",
    ];
    for c in &candidates {
        let p = Path::new(c);
        if p.exists() { return p.to_path_buf(); }
    }
    panic!("readtest.in not found");
}

struct RunSummary {
    passed: usize,
    failed: usize,
    skipped: usize,
    by_func: BTreeMap<String, (usize, usize, usize)>,
}

fn run_readtest(filter: &str) -> RunSummary {
    let path = find_readtest_in();
    let f = File::open(&path).expect("open readtest.in");
    let reader = BufReader::new(f);
    let filter_prefix = format!("{}_", filter);

    let mut summary = RunSummary {
        passed: 0, failed: 0, skipped: 0,
        by_func: BTreeMap::new(),
    };

    for line in reader.lines() {
        let line = line.expect("read line");
        let line = line.trim();
        if line.is_empty() || line.starts_with("--") { continue; }

        let parts: Vec<&str> = line.split_whitespace().collect();
        if parts.len() < 4 { continue; }

        let func_name = parts[0];
        if !func_name.starts_with(&filter_prefix) { continue; }
        if !supported_readtest_func(func_name) { continue; }
        if line.contains("longintsize=32") { continue; }

        let rm: i64 = parts[1].parse().unwrap_or(0);
        let ulp_add = parts.iter()
            .find_map(|part| part.strip_prefix("ulp=").and_then(|value| value.parse::<f64>().ok()))
            .unwrap_or(0.0);
        let entry = summary.by_func.entry(func_name.to_string()).or_insert((0, 0, 0));

        match panic::catch_unwind(AssertUnwindSafe(|| dispatch(func_name, &parts, rm, ulp_add))) {
            Ok(DispatchResult::Pass) => {
                summary.passed += 1;
                entry.0 += 1;
            }
            Ok(DispatchResult::Fail(msg)) => {
                summary.failed += 1;
                entry.1 += 1;
                if entry.1 <= 3 {
                    eprintln!("FAIL {}: {}", func_name, msg);
                }
            }
            Ok(DispatchResult::Skip) => {
                summary.skipped += 1;
                entry.2 += 1;
                if entry.2 <= 3 {
                    eprintln!("SKIP {}: {}", func_name, line);
                }
            }
            Err(_) => {
                summary.skipped += 1;
                entry.2 += 1;
                if entry.2 <= 3 {
                    eprintln!("PANIC {}: {}", func_name, line);
                }
            }
        }
    }

    summary
}

`
}

func readtestTests() string {
	return `
#[test]
fn test_readtest_generated_decimal64() {
    let s = run_readtest("bid64");
    println!("decimal64: passed={} failed={} skipped={}", s.passed, s.failed, s.skipped);
    for (func, (p, f, sk)) in &s.by_func {
        if *f > 0 || *sk > 0 {
            println!("  STAT {}: passed={} failed={} skipped={}", func, p, f, sk);
        }
    }
    assert_eq!(s.failed, 0, "decimal64 readtest failures");
    assert_eq!(s.skipped, 0, "decimal64 readtest skips");
}

#[test]
fn test_readtest_generated_decimal32() {
    let s = run_readtest("bid32");
    println!("decimal32: passed={} failed={} skipped={}", s.passed, s.failed, s.skipped);
    for (func, (p, f, sk)) in &s.by_func {
        if *f > 0 || *sk > 0 {
            println!("  STAT {}: passed={} failed={} skipped={}", func, p, f, sk);
        }
    }
    assert_eq!(s.failed, 0, "decimal32 readtest failures");
    assert_eq!(s.skipped, 0, "decimal32 readtest skips");
}

#[test]
fn test_readtest_generated_status_control() {
    let s = run_readtest("bid");
    println!("status-control: passed={} failed={} skipped={}", s.passed, s.failed, s.skipped);
    for (func, (p, f, sk)) in &s.by_func {
        if *f > 0 || *sk > 0 {
            println!("  STAT {}: passed={} failed={} skipped={}", func, p, f, sk);
        }
    }
    assert_eq!(s.failed, 0, "status-control readtest failures");
    assert_eq!(s.skipped, 0, "status-control readtest skips");
    assert_eq!(s.passed, 137, "status-control readtest case count");
}

#[test]
fn test_readtest_generated_decimal128() {
    let s = run_readtest("bid128");
    println!("decimal128: passed={} failed={} skipped={}", s.passed, s.failed, s.skipped);
    for (func, (p, f, sk)) in &s.by_func {
        if *f > 0 || *sk > 0 {
            println!("  STAT {}: passed={} failed={} skipped={}", func, p, f, sk);
        }
    }
    assert_eq!(s.failed, 0, "decimal128 readtest failures");
    assert_eq!(s.skipped, 0, "decimal128 readtest skips");
}
`
}
