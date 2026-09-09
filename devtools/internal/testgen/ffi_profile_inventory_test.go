package testgen

import (
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// These tests are the hand-written anchor of the FFI profile census. The census
// itself is generated, so a regression inside the generator would regenerate a
// self-consistent artifact and pass verify-generated; what cannot be regenerated
// away is the requirement that the census stay a closed world in both
// directions. Every pinned Intel symbol must land in exactly one accounted
// state, an unaccounted symbol must fail generation, and a registry entry that
// names nothing real must fail generation.

func loadFFIProfileInventoryForTest(t *testing.T) (FFITestSpec, []symbolSpec, map[string]bidgoFuncSig, ffiProfileInventory) {
	t.Helper()
	repoRoot := filepath.Clean(filepath.Join("..", ".."))
	manifest, err := LoadManifest(filepath.Join(repoRoot, "testgen_manifest.json"))
	if err != nil {
		t.Fatalf("LoadManifest error: %v", err)
	}
	if len(manifest.FFITests) != 1 {
		t.Fatalf("manifest declares %d ffi_tests suites, want 1", len(manifest.FFITests))
	}
	spec := manifest.FFITests[0]
	symbols, err := loadSymbolFile(filepath.Join(repoRoot, spec.Symbols))
	if err != nil {
		t.Fatalf("loadSymbolFile error: %v", err)
	}
	sigs, err := loadBidgoFuncSigs(filepath.Join(repoRoot, readtestGoportBidgoSrcDir))
	if err != nil {
		t.Fatalf("loadBidgoFuncSigs error: %v", err)
	}
	inventory, err := buildFFIProfileInventory(spec, symbols.Symbols, sigs)
	if err != nil {
		t.Fatalf("buildFFIProfileInventory error: %v", err)
	}
	return spec, symbols.Symbols, sigs, inventory
}

// TestFFIProfileInventoryAccountsForEveryPinnedSymbol closes the census in the
// "no symbol escapes" direction: one row per pinned Intel symbol, no extra rows,
// every row classified with a non-empty reason, and the header counts equal to
// the rows they summarize.
func TestFFIProfileInventoryAccountsForEveryPinnedSymbol(t *testing.T) {
	_, symbols, _, inventory := loadFFIProfileInventoryForTest(t)

	if inventory.TotalFunctions != len(symbols) {
		t.Errorf("inventory total_functions = %d, pinned symbol file holds %d", inventory.TotalFunctions, len(symbols))
	}
	if len(inventory.Functions) != len(symbols) {
		t.Fatalf("inventory rows = %d, pinned symbol file holds %d", len(inventory.Functions), len(symbols))
	}

	rows := make(map[string]ffiProfileFunctionInventory, len(inventory.Functions))
	for _, row := range inventory.Functions {
		if _, dup := rows[row.Function]; dup {
			t.Errorf("inventory carries a duplicate row for %q", row.Function)
		}
		rows[row.Function] = row
	}
	for _, symbol := range symbols {
		row, ok := rows[symbol.Name]
		if !ok {
			t.Errorf("pinned Intel symbol %q has no census row", symbol.Name)
			continue
		}
		if strings.TrimSpace(row.Classification) == "" {
			t.Errorf("census row %q has an empty classification", symbol.Name)
		}
		if strings.TrimSpace(row.Reason) == "" {
			t.Errorf("census row %q has an empty reason", symbol.Name)
		}
	}

	included, excluded := 0, 0
	byClass := map[string]int{}
	for _, row := range inventory.Functions {
		if row.Included {
			included++
			if row.Classification != ffiProfileClassSelected {
				t.Errorf("included census row %q has classification %q, want %q", row.Function, row.Classification, ffiProfileClassSelected)
			}
			if row.PortFunction == "" {
				t.Errorf("included census row %q names no Go mechanical-port counterpart", row.Function)
			}
			continue
		}
		excluded++
		byClass[row.Classification]++
		if row.Classification == ffiProfileClassSelected {
			t.Errorf("excluded census row %q is classified as selected", row.Function)
		}
	}
	if included != inventory.IncludedFunctions {
		t.Errorf("inventory included_functions = %d, rows say %d", inventory.IncludedFunctions, included)
	}
	if excluded != inventory.ExcludedFunctions {
		t.Errorf("inventory excluded_functions = %d, rows say %d", inventory.ExcludedFunctions, excluded)
	}
	if included+excluded != inventory.TotalFunctions {
		t.Errorf("included %d + excluded %d != total %d", included, excluded, inventory.TotalFunctions)
	}
	if len(byClass) != len(inventory.ExcludedByClassification) {
		t.Errorf("excluded_by_classification has %d classes, rows carry %d", len(inventory.ExcludedByClassification), len(byClass))
	}
	for class, want := range inventory.ExcludedByClassification {
		if byClass[class] != want {
			t.Errorf("excluded_by_classification[%q] = %d, rows say %d", class, want, byClass[class])
		}
	}
}

// TestFFIProfileInventoryRejectsUnaccountedSymbol is the negative control for
// the "neither selected nor excluded" state. Dropping one selected function from
// the suite leaves a pinned symbol whose declaration the FFI case model
// expresses and whose name resolves to a real Go mechanical-port entrypoint, so
// no mechanical exclusion rule can own it; with the registry empty it reaches
// the unclassified state unaccounted and must fail generation rather than be
// silently absorbed.
//
// The probe drives the suite rather than the registry because the registry is
// empty: with nothing to drop there, removing a registry entry could not put any
// symbol into the unaccounted state, and the control would pass vacuously.
func TestFFIProfileInventoryRejectsUnaccountedSymbol(t *testing.T) {
	spec, symbols, sigs, _ := loadFFIProfileInventoryForTest(t)

	probeName := "bid64_nextafter"
	var probe symbolSpec
	for _, symbol := range symbols {
		if symbol.Name == probeName {
			probe = symbol
			break
		}
	}
	if probe.Name == "" {
		t.Fatalf("probe symbol %q is absent from the pinned symbol file", probeName)
	}
	if ffiProfilePortFunction(probe.Name, sigs) == "" {
		t.Fatalf("probe symbol %q has no Go mechanical-port counterpart, so it cannot reach the unclassified state", probeName)
	}
	if reason, ok := ffiProfileUnclassifiedExclusions[probeName]; ok {
		t.Fatalf("probe symbol %q is registered as unclassified debt (%q), so dropping it from the suite would not leave it unaccounted", probeName, reason)
	}

	reduced := spec
	reduced.Functions = make([]string, 0, len(spec.Functions))
	for _, function := range spec.Functions {
		if function == probeName {
			continue
		}
		reduced.Functions = append(reduced.Functions, function)
	}
	if len(reduced.Functions) == len(spec.Functions) {
		t.Fatalf("suite %q does not select probe symbol %q", spec.Name, probeName)
	}

	_, err := buildFFIProfileInventory(reduced, symbols, sigs)
	if err == nil {
		t.Fatalf("census accepted %q as neither selected nor excluded; an unaccounted symbol must fail generation", probeName)
	}
	if !strings.Contains(err.Error(), probeName) {
		t.Fatalf("census failure does not name the unaccounted symbol %q: %v", probeName, err)
	}
}

// TestFFIProfileInventoryRejectsStaleUnclassifiedEntry is the negative control
// for the opposite direction: a registry entry that names no symbol reaching the
// unclassified state must fail generation instead of standing as a reasonless
// permanent allowance. With the registry empty the control injects a synthetic
// entry naming a symbol the suite selects, which can never reach the
// unclassified state, so the stale direction stays exercised.
func TestFFIProfileInventoryRejectsStaleUnclassifiedEntry(t *testing.T) {
	spec, symbols, sigs, _ := loadFFIProfileInventoryForTest(t)

	names := make([]string, 0, len(ffiProfileUnclassifiedExclusions))
	for name := range ffiProfileUnclassifiedExclusions {
		names = append(names, name)
	}
	sort.Strings(names)

	if len(names) > 0 {
		dropped := names[0]
		reduced := make([]symbolSpec, 0, len(symbols))
		for _, symbol := range symbols {
			if symbol.Name == dropped {
				continue
			}
			reduced = append(reduced, symbol)
		}
		if len(reduced) == len(symbols) {
			t.Fatalf("registry entry %q names no pinned symbol; the census world is already broken", dropped)
		}

		_, err := buildFFIProfileInventory(spec, reduced, sigs)
		if err == nil {
			t.Fatalf("census accepted a registry entry for %q that names no reachable symbol", dropped)
		}
		if !strings.Contains(err.Error(), dropped) {
			t.Fatalf("census failure does not name the stale entry %q: %v", dropped, err)
		}
		return
	}

	synthetic := "bid64_add"
	selected := false
	for _, function := range spec.Functions {
		if function == synthetic {
			selected = true
			break
		}
	}
	if !selected {
		t.Fatalf("synthetic stale probe %q is not selected by suite %q, so it could legitimately reach the unclassified state", synthetic, spec.Name)
	}
	restore := ffiProfileUnclassifiedExclusions
	ffiProfileUnclassifiedExclusions = map[string]string{
		synthetic: "synthetic stale entry for the negative control; this symbol is selected by the suite and can never reach the unclassified state",
	}
	defer func() { ffiProfileUnclassifiedExclusions = restore }()

	_, err := buildFFIProfileInventory(spec, symbols, sigs)
	if err == nil {
		t.Fatalf("census accepted a registry entry for selected function %q, which never reaches the unclassified state", synthetic)
	}
	if !strings.Contains(err.Error(), synthetic) {
		t.Fatalf("census failure does not name the stale entry %q: %v", synthetic, err)
	}
}

// TestFFIProfileExclusionRulesStayMechanical keeps the mechanical exclusion
// classes honest: every rule must own at least one excluded symbol (a rule that
// owns nothing is a stale class), and no rule may claim a selected symbol (a
// rule that fires on a selected function is describing the subset instead of a
// structural fact).
func TestFFIProfileExclusionRulesStayMechanical(t *testing.T) {
	spec, symbols, sigs, _ := loadFFIProfileInventoryForTest(t)

	included, err := expandFFIFunctions(spec, symbols)
	if err != nil {
		t.Fatalf("expandFFIFunctions error: %v", err)
	}
	includedSet := make(map[string]bool, len(included))
	for _, function := range included {
		includedSet[function] = true
	}

	matched := map[string]int{}
	for _, symbol := range symbols {
		port := ffiProfilePortFunction(symbol.Name, sigs)
		for _, rule := range ffiProfileExclusionRules {
			reason, ok := rule.reason(symbol, port)
			if !ok {
				continue
			}
			if strings.TrimSpace(reason) == "" {
				t.Errorf("exclusion rule %q produced an empty reason for %q", rule.name, symbol.Name)
			}
			if includedSet[symbol.Name] {
				t.Errorf("exclusion rule %q fires on selected function %q; a mechanical rule must describe a structural blocker, not the manifest subset", rule.name, symbol.Name)
			}
			matched[rule.name]++
			break
		}
	}
	for _, rule := range ffiProfileExclusionRules {
		if matched[rule.name] == 0 {
			t.Errorf("exclusion rule %q matched no symbol; remove the stale rule", rule.name)
		}
	}
}
