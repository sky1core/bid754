package testgen

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// The FFI profile inventory is the closed-world census of the C FFI exact
// bit-compare domain, in the same shape as the readtest profile inventory: the
// pinned Intel symbol file is the universe, and every symbol in it is either
// selected by the manifest `ffi_tests` suite or excluded with a classification
// and a concrete reason. Neither state is optional — a symbol that reaches
// neither is a generation-time hard failure, which is what makes "the subset is
// named subset and nobody accounts for the rest" impossible to restore.
//
// The exclusion reasons are derived, in priority order, from facts that live
// outside this file:
//
//  1. the Go mechanical-port surface (bid754-go/internal/bidgo), because this
//     domain compares pinned Intel C against that port and a symbol with no
//     exported port counterpart has no second leg to compare;
//  2. the pinned Intel declaration, because the profile's case model emits raw
//     operand bit patterns and integer literals and compares one returned
//     scalar plus the status word — declarations outside that model cannot be
//     expressed as cases at all.
//
// Both are mechanical and non-circular: neither consults the manifest function
// list or the generator's hand-maintained operation switch, so neither can
// justify the current subset by restating it. What survives both is the real
// accounting surface, and it is registered by exact name in
// ffiProfileUnclassifiedExclusions as declared debt rather than given a
// fabricated reason.
const ffiProfileInventoryPath = "generated/testspec/ffi_profile_inventory.json"

const (
	ffiProfileClassSelected   = "selected"
	ffiProfileClassOutOfScope = "out_of_scope_not_required"
	ffiProfileClassUnresolved = "unresolved_required_review"
	ffiProfileSelectionReason = "selected by the FFI exact bit-compare profile"
	ffiProfileFlagsParam      = "_IDEC_flags"
	ffiProfileRoundingParam   = "_IDEC_round"
)

type ffiProfileFunctionInventory struct {
	Function       string `json:"function"`
	Included       bool   `json:"included"`
	PortFunction   string `json:"port_function,omitempty"`
	Classification string `json:"classification"`
	Reason         string `json:"reason"`
}

type ffiProfileInventory struct {
	Version                  string                        `json:"version"`
	Suite                    string                        `json:"suite"`
	Symbols                  string                        `json:"symbols"`
	Source                   string                        `json:"source"`
	TotalFunctions           int                           `json:"total_functions"`
	IncludedFunctions        int                           `json:"included_functions"`
	ExcludedFunctions        int                           `json:"excluded_functions"`
	ExcludedByClassification map[string]int                `json:"excluded_by_classification"`
	Functions                []ffiProfileFunctionInventory `json:"functions"`
}

// ffiProfileExclusionRule is one mechanically derived exclusion class. Rules are
// evaluated in declaration order and each must match at least one excluded
// symbol: a rule that matches nothing is a stale class that would quietly widen
// the census world, so it fails generation instead.
type ffiProfileExclusionRule struct {
	name string
	// reason returns the concrete exclusion reason for this symbol, or ok=false
	// when the rule does not apply. portFunction is the resolved Go
	// mechanical-port entrypoint, empty when the port declares none.
	reason func(symbol symbolSpec, portFunction string) (string, bool)
}

var ffiProfileExclusionRules = []ffiProfileExclusionRule{
	{
		name: "no_port_counterpart",
		reason: func(symbol symbolSpec, portFunction string) (string, bool) {
			if portFunction != "" {
				return "", false
			}
			return "bid754-go/internal/bidgo declares no exported mechanical-port counterpart " +
				"({base, *WithFlags, *Raw} under the shared readtest name normalization), so the " +
				"Intel-C-versus-port exact bit-compare has no second leg for this symbol", true
		},
	},
	{
		name: "non_status_pointer_parameter",
		reason: func(symbol symbolSpec, portFunction string) (string, bool) {
			param, ok := ffiProfileNonStatusPointerParam(symbol)
			if !ok {
				return "", false
			}
			return fmt.Sprintf("the pinned Intel declaration takes pointer parameter %q, which is neither an "+
				"operand nor the trailing status word; the FFI case model emits raw operand bit patterns and "+
				"integer literals only, so this signature has no representable case", param), true
		},
	},
	{
		name: "no_value_operand",
		reason: func(symbol symbolSpec, portFunction string) (string, bool) {
			if ffiProfileHasValueOperand(symbol) {
				return "", false
			}
			return "the pinned Intel declaration takes no value operand (every parameter is a rounding-mode or " +
				"status-flag word), so the FFI operand generator has nothing to vary and the profile's " +
				"operand-driven case model does not apply", true
		},
	},
	{
		name: "platform_width_long_result",
		reason: func(symbol symbolSpec, portFunction string) (string, bool) {
			if symbol.ReturnType != "long int" {
				return "", false
			}
			return "the pinned Intel declaration returns platform-width `long int`, which the FFI case model " +
				"cannot bit-compare as a fixed-width scalar; upstream readtest data carries a `longintsize=32` " +
				"row class for exactly this width dependency, and the generated profile skips those rows", true
		},
	},
}

// ffiProfileUnclassifiedExclusions registers, by exact Intel symbol name, every
// symbol that survives all mechanical exclusion rules and is still absent from
// the manifest FFI function list. These are not classified exclusions: the Go
// mechanical port exposes a counterpart and the pinned declaration fits the
// profile's operand/result model, so nothing structural keeps them out. They are
// recorded as `unresolved_required_review` debt, which docs/SPEC-level scope
// classification defines as required work not closed, rather than being given an
// invented reason.
//
// The registry is currently empty: every symbol that reached this state — the
// fifteen fixed-attribute round-to-integral variants, nexttoward, nextafter,
// fdim, nearbyint, llrint, and llround across all three widths — is now selected
// by the manifest suite. The mechanism stays because the census is generated
// from the pinned symbol file: a future Intel pin, port export, or exclusion-rule
// change can put a new symbol into this state, and generation fails until it is
// either selected or registered here with a written reason.
//
// Closed world in both directions: an entry that names no pinned symbol, names
// an included symbol, names a symbol with no port counterpart, or names a symbol
// a mechanical rule already owns fails generation as stale; a symbol that
// reaches this state without an entry fails generation as unaccounted.
var ffiProfileUnclassifiedExclusions = map[string]string{}

// ffiProfileNonStatusPointerParam returns the first pointer parameter that is
// not the trailing `_IDEC_flags*` status word. That trailing pointer is how
// nearly every selected Intel entrypoint reports flags, so it is not a blocker;
// any other pointer is a string input or an out-parameter, and the profile's
// raw-operand case model cannot express either.
func ffiProfileNonStatusPointerParam(symbol symbolSpec) (string, bool) {
	params := symbol.Parameters
	if n := len(params); n > 0 && strings.HasPrefix(params[n-1], ffiProfileFlagsParam+"*") {
		params = params[:n-1]
	}
	for _, param := range params {
		if strings.Contains(param, "*") {
			return param, true
		}
	}
	return "", false
}

// ffiProfileHasValueOperand reports whether the declaration carries at least one
// parameter that is neither a rounding mode nor a status-flag word, i.e. at
// least one operand the FFI generator can vary.
func ffiProfileHasValueOperand(symbol symbolSpec) bool {
	for _, param := range symbol.Parameters {
		if strings.HasPrefix(param, ffiProfileRoundingParam) || strings.HasPrefix(param, ffiProfileFlagsParam) {
			continue
		}
		return true
	}
	return false
}

// ffiProfilePortFunction resolves the Go mechanical-port entrypoint for an Intel
// symbol through the same {base, *WithFlags, *Raw} candidate search the readtest
// goport dispatch uses, so the census and the dispatch agree on what "the port
// implements this" means.
func ffiProfilePortFunction(function string, sigs map[string]bidgoFuncSig) string {
	candidates := goportCandidateSigs(function, sigs)
	if len(candidates) == 0 {
		return ""
	}
	base := normalizeGoportFuncName(function)
	for _, candidate := range candidates {
		if normalizeGoportFuncName(candidate.Name) == base {
			return candidate.Name
		}
	}
	return candidates[0].Name
}

func buildFFIProfileInventory(spec FFITestSpec, symbols []symbolSpec, sigs map[string]bidgoFuncSig) (ffiProfileInventory, error) {
	included, err := expandFFIFunctions(spec, symbols)
	if err != nil {
		return ffiProfileInventory{}, err
	}
	includedSet := make(map[string]bool, len(included))
	for _, function := range included {
		includedSet[function] = true
	}

	inventory := ffiProfileInventory{
		Version: "1.0",
		Suite:   spec.Name,
		Symbols: filepath.ToSlash(spec.Symbols),
		Source: "devtools/testgen_manifest.json ffi_tests + " + filepath.ToSlash(spec.Symbols) +
			" + bid754-go/internal/bidgo",
		TotalFunctions:           len(symbols),
		ExcludedByClassification: map[string]int{},
		Functions:                make([]ffiProfileFunctionInventory, 0, len(symbols)),
	}

	matchedRules := map[string]bool{}
	usedUnclassified := map[string]bool{}
	for _, symbol := range symbols {
		port := ffiProfilePortFunction(symbol.Name, sigs)
		row := ffiProfileFunctionInventory{
			Function:     symbol.Name,
			PortFunction: port,
		}

		if includedSet[symbol.Name] {
			if port == "" {
				return ffiProfileInventory{}, fmt.Errorf("ffi profile inventory: suite %q selects %q but bid754-go/internal/bidgo declares no exported mechanical-port counterpart, so the differential has no second leg", spec.Name, symbol.Name)
			}
			row.Included = true
			row.Classification = ffiProfileClassSelected
			row.Reason = ffiProfileSelectionReason
			inventory.IncludedFunctions++
			inventory.Functions = append(inventory.Functions, row)
			continue
		}

		inventory.ExcludedFunctions++
		matched := false
		for _, rule := range ffiProfileExclusionRules {
			reason, ok := rule.reason(symbol, port)
			if !ok {
				continue
			}
			if strings.TrimSpace(reason) == "" {
				return ffiProfileInventory{}, fmt.Errorf("ffi profile inventory: exclusion rule %q produced an empty reason for %q", rule.name, symbol.Name)
			}
			matchedRules[rule.name] = true
			row.Classification = ffiProfileClassOutOfScope
			row.Reason = reason
			matched = true
			break
		}
		if !matched {
			reason, ok := ffiProfileUnclassifiedExclusions[symbol.Name]
			if !ok {
				return ffiProfileInventory{}, fmt.Errorf("ffi profile inventory: symbol %q is neither selected by suite %q nor excluded: the Go mechanical port declares counterpart %q and no exclusion rule applies to its pinned declaration. Add it to the suite, or register it in ffiProfileUnclassifiedExclusions with the reason it stays out", symbol.Name, spec.Name, port)
			}
			if strings.TrimSpace(reason) == "" {
				return ffiProfileInventory{}, fmt.Errorf("ffi profile inventory: ffiProfileUnclassifiedExclusions[%q] has an empty reason", symbol.Name)
			}
			usedUnclassified[symbol.Name] = true
			row.Classification = ffiProfileClassUnresolved
			row.Reason = reason
		}
		inventory.ExcludedByClassification[row.Classification]++
		inventory.Functions = append(inventory.Functions, row)
	}

	for _, rule := range ffiProfileExclusionRules {
		if !matchedRules[rule.name] {
			return ffiProfileInventory{}, fmt.Errorf("ffi profile inventory: exclusion rule %q matched no excluded symbol; remove the stale rule so the census world stays closed", rule.name)
		}
	}
	var staleUnclassified []string
	for name := range ffiProfileUnclassifiedExclusions {
		if !usedUnclassified[name] {
			staleUnclassified = append(staleUnclassified, name)
		}
	}
	if len(staleUnclassified) > 0 {
		sort.Strings(staleUnclassified)
		return ffiProfileInventory{}, fmt.Errorf("ffi profile inventory: ffiProfileUnclassifiedExclusions entries %v name no symbol that reaches the unclassified state (absent from %s, selected by the suite, without a port counterpart, or already owned by a mechanical exclusion rule); remove the stale entries", staleUnclassified, inventory.Symbols)
	}

	return inventory, nil
}

// GenerateFFIProfileInventory returns the generated FFI profile census keyed by
// repo-relative path.
func GenerateFFIProfileInventory(repoRoot string, manifest Manifest) (map[string][]byte, error) {
	if len(manifest.FFITests) != 1 {
		return nil, fmt.Errorf("ffi profile inventory: manifest declares %d ffi_tests suites, want exactly 1 so the census has a single closed world", len(manifest.FFITests))
	}
	spec := manifest.FFITests[0]
	symbols, err := loadSymbolFile(filepath.Join(repoRoot, spec.Symbols))
	if err != nil {
		return nil, err
	}
	sigs, err := loadBidgoFuncSigs(filepath.Join(repoRoot, readtestGoportBidgoSrcDir))
	if err != nil {
		return nil, err
	}
	inventory, err := buildFFIProfileInventory(spec, symbols.Symbols, sigs)
	if err != nil {
		return nil, err
	}
	data, err := json.MarshalIndent(inventory, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal ffi profile inventory: %w", err)
	}
	return map[string][]byte{ffiProfileInventoryPath: append(data, '\n')}, nil
}

// WriteFFIProfileInventoryOutput regenerates the FFI profile census from the
// pinned symbol file, the manifest suite, and the Go mechanical-port surface.
func WriteFFIProfileInventoryOutput(repoRoot string, manifest Manifest) error {
	files, err := GenerateFFIProfileInventory(repoRoot, manifest)
	if err != nil {
		return err
	}
	for rel, data := range files {
		fullPath := filepath.Join(repoRoot, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			return fmt.Errorf("mkdir %q: %w", filepath.Dir(fullPath), err)
		}
		if err := os.WriteFile(fullPath, data, 0o644); err != nil {
			return fmt.Errorf("write generated ffi profile inventory %q: %w", fullPath, err)
		}
	}
	return nil
}
