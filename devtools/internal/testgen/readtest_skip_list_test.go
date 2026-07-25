package testgen

// Non-vacuity gate for the hand-maintained exact-function-name list switches the
// readtest selection layer consults (the historical skip list, the current
// declared exclusion list, the Tier 3 CMP_RELATIVEERR adoption list, the status-
// control flag/rounding-direction lists, and the Tier 1/Tier 2 mixed-width
// lists). Each is a closed-world list keyed by exact function name, so an entry
// naming a function that is not in the pinned Intel readtest census is dead code
// that still reads like a live, reasoned decision and invites the next agent to
// extend it.
//
// The census these lists must match is the one the selection code itself
// enumerates: buildReadtestProfileInventory emits one inventory row for every
// function parseReadtestFunctionSpecs finds in the pinned readtest.h, and the
// checked-in generated readtest profile inventory is byte-compared against
// regeneration by `make verify-generated`. Reading the checked-in inventory
// keeps this gate running in environments without the gitignored pinned Intel
// download, where the header itself is absent. The inventory also carries each
// row's comparator group, which is what lets the Tier 3 adoption list be checked
// against comparator identity instead of only against name existence.
//
// The lists are switch arms rather than table vars, so the declared names are
// read out of the readtest_spec.go AST instead of being restated here (a
// restated copy would keep passing while the real switch drifted) and then
// cross-checked against the live predicates in both directions. The registration
// below is hand-maintained too, so a second check walks the selection layer and
// fails on any func(string) bool it reaches that is neither registered as a
// name-keyed list nor excepted with a written reason.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// readtestSpecSourceFile is the readtest selection source whose list switches
// this file reads. Tests run in the package directory.
const readtestSpecSourceFile = "readtest_spec.go"

// readtestNameKeyedList is one hand-maintained exact-function-name list reached
// from the readtest selection roots. requiredCompareGroup, when set, tightens
// the census check from "the name exists upstream" to "the name has an upstream
// row under this comparator", which is the property the list actually depends
// on.
type readtestNameKeyedList struct {
	predicate            func(string) bool
	requiredCompareGroup string
}

// readtestNameKeyedLists registers every exact-name list the selection layer
// consults, paired with its live predicate so the AST-declared names can be
// checked against real behavior instead of against a second transcription.
// Exclusion lists, selection lists, and surface-classification lists are checked
// alike: all are closed name lists whose entries must exist in the pinned census
// regardless of which way they decide.
var readtestNameKeyedLists = map[string]readtestNameKeyedList{
	"isHistoricalReadtestSkipFunction":      {predicate: isHistoricalReadtestSkipFunction},
	"isCurrentSpecReadtestExcludedFunction": {predicate: isCurrentSpecReadtestExcludedFunction},
	// The Tier 3 adoption list exists only to select readtest.h's duplicate
	// CMP_RELATIVEERR comparator blocks, so an entry whose upstream rows are all
	// CMP_FUZZYSTATUS adopts nothing even though the function name is real.
	"isTier3RelativeErrSelectedFunction":         {predicate: isTier3RelativeErrSelectedFunction, requiredCompareGroup: "CMP_RELATIVEERR"},
	"isFlagSubsetReadtestFunction":               {predicate: isFlagSubsetReadtestFunction},
	"isDecimalRoundingDirectionReadtestFunction": {predicate: isDecimalRoundingDirectionReadtestFunction},
	"isTier1MixedWidthIntelReadtestFunction":     {predicate: isTier1MixedWidthIntelReadtestFunction},
	"isTier2MixedWidthIntelReadtestFunction":     {predicate: isTier2MixedWidthIntelReadtestFunction},
}

// readtestNonNameKeyedPredicates are the func(string) bool predicates the
// selection layer reaches that are not exact-function-name lists, each with the
// reason the census-name check cannot apply to it. Closed world in both
// directions: an unlisted predicate fails as unreviewed, and a listed predicate
// the scan no longer reaches fails as stale.
var readtestNonNameKeyedPredicates = map[string]string{
	"isSupportedReadtestScalarOutput":    "keyed by readtest.h OP_* scalar operand/result type tokens, not by function name",
	"isSupportedReadtestDecimalOutput":   "keyed by readtest.h OP_DEC* type tokens, not by function name",
	"isSupportedReadtestInput":           "composes the two OP_* type-token predicates, so it is keyed by type token, not by function name",
	"isMixedWidthIntelReadtestExtension": "matches Intel mixed-width family name prefixes rather than exact function names, so its entries are prefixes with no census row to match; the exact mixed-width surface is carried by the Tier 1/Tier 2 lists registered above",
}

// readtestSelectionRootFuncs are the two readtest_spec.go functions that decide
// per pinned readtest function whether it becomes a generated suite
// (buildProfileReadTest) and why it did not (readtestProfileExclusion). The
// registration scan starts here and follows package-local calls, so the surface
// classification layer these two share is covered as well.
var readtestSelectionRootFuncs = []string{"buildProfileReadTest", "readtestProfileExclusion"}

func TestReadtestNameKeyedListEntriesMatchPinnedFunctionCensus(t *testing.T) {
	census := pinnedReadtestFunctionCensus(t)
	censusNames := census.functionNames()

	for _, predicateName := range sortedNameKeyedListNames(readtestNameKeyedLists) {
		list := readtestNameKeyedLists[predicateName]
		t.Run(predicateName, func(t *testing.T) {
			// An empty list (no case arms) is vacuously consistent: it decides
			// nothing and therefore carries no stale entry.
			declared := readtestNameKeyedListSwitchNames(t, predicateName)
			declaredSet := make(map[string]bool, len(declared))
			for _, name := range declared {
				declaredSet[name] = true
			}

			for _, name := range declared {
				// Extraction fidelity: a name read out of a case arm must
				// actually be matched by the compiled predicate, otherwise the
				// checks below are testing something other than live behavior.
				if !list.predicate(name) {
					t.Errorf("%s(%q) = false although %q appears in its case arms; the AST extraction and the compiled switch disagree", predicateName, name, name)
				}
				if !census.has(name) {
					t.Errorf("%s entry %q is not a function in the pinned readtest census (%s); remove the stale entry", predicateName, name, readtestProfileInventoryRelPath)
					continue
				}
				if list.requiredCompareGroup != "" && !census.hasCompareGroup(name, list.requiredCompareGroup) {
					t.Errorf("%s entry %q has no pinned readtest row with compare group %s (rows: %v); the entry decides nothing, so remove it", predicateName, name, list.requiredCompareGroup, census.compareGroupsOf(name))
				}
			}

			// Closed world in the other direction: every census function the
			// predicate matches must come from a declared case arm, so a switch
			// replaced by prefix/substring matching (which this file cannot
			// enumerate) fails instead of silently escaping the gate.
			for _, name := range censusNames {
				if list.predicate(name) && !declaredSet[name] {
					t.Errorf("%s matches census function %q with no matching case arm; keep the list an enumerable set of exact function names", predicateName, name)
				}
			}
		})
	}
}

// TestReadtestSelectionLayerPredicatesAreRegisteredOrExcepted closes the
// registration itself: both maps above are hand-maintained, so a further
// name-keyed list wired anywhere in the selection layer would otherwise get no
// census check at all. The scan starts at the selection roots and follows
// package-local calls transitively, so a list added one layer down (inside the
// shared surface classification, for instance) is covered without naming that
// layer here.
//
// Boundary: reachability is resolved syntactically over calls made through a
// plain identifier to a func declared in package testgen's non-test sources. A
// predicate defined in another package, reached through a function value,
// method, or interface, or reached only from a selection-layer function that no
// root calls by identifier, is outside this scan and would need its own gate.
// Two more shapes inside the reached layer are outside it as well, both because
// this scan recognizes a list only as a func(string) bool: an exact-name list
// written as func(readtestFunctionSpec) bool — the dominant signature style in
// readtest_spec.go — is not seen as a predicate here, and the inline
// `switch fn.Name` / `fn.Name == "..."` literals that reached functions test
// directly are not predicates at all. Every occurrence of both shapes present
// when this gate was written was audited by hand against the pinned census, but
// neither shape is checked mechanically.
func TestReadtestSelectionLayerPredicatesAreRegisteredOrExcepted(t *testing.T) {
	decls := readtestPackageFuncDecls(t)

	callersOf := make(map[string][]string)
	reached := 0
	for _, name := range readtestSelectionRootFuncs {
		if decl, ok := decls[name]; !ok || decl.Body == nil {
			t.Fatalf("package testgen declares no func %s with a body; the readtest selection root moved and this gate no longer scans it", name)
		}
	}
	for predicate, callers := range stringPredicateCallersReachableFrom(decls, readtestSelectionRootFuncs) {
		callersOf[predicate] = callers
		reached++
	}
	if reached == 0 {
		t.Fatalf("the readtest selection roots %v reach no string predicate at all; this gate lost its subject", readtestSelectionRootFuncs)
	}

	predicateNames := make([]string, 0, len(callersOf))
	for name := range callersOf {
		predicateNames = append(predicateNames, name)
	}
	sort.Strings(predicateNames)
	for _, name := range predicateNames {
		_, registered := readtestNameKeyedLists[name]
		reason, excepted := readtestNonNameKeyedPredicates[name]
		switch {
		case registered && excepted:
			t.Errorf("%s is both registered as a name-keyed list and excepted as non-name-keyed; pick one", name)
		case registered:
		case excepted:
			t.Logf("excepted non-name-keyed predicate %s (called by %v): %s", name, callersOf[name], reason)
		default:
			t.Errorf("the readtest selection layer consults string predicate %s (called by %v), which is neither registered in readtestNameKeyedLists nor excepted with a reason in readtestNonNameKeyedPredicates; register it so its entries are checked against the pinned census, or except it with the reason the census check does not apply", name, callersOf[name])
		}
	}

	for _, name := range sortedNameKeyedListNames(readtestNameKeyedLists) {
		if len(callersOf[name]) == 0 {
			t.Errorf("registered name-keyed list %s is reached by nothing in the readtest selection layer (roots %v); remove the stale registration", name, readtestSelectionRootFuncs)
		}
	}
	for _, name := range sortedStringMapKeys(readtestNonNameKeyedPredicates) {
		if strings.TrimSpace(readtestNonNameKeyedPredicates[name]) == "" {
			t.Errorf("excepted predicate %s carries no written reason", name)
		}
		if len(callersOf[name]) == 0 {
			t.Errorf("excepted predicate %s is reached by nothing in the readtest selection layer (roots %v); remove the stale exception", name, readtestSelectionRootFuncs)
		}
	}
}

// readtestProfileInventoryRelPath is the checked-in generated spec index that
// carries the readtest profile inventory.
var readtestProfileInventoryRelPath = filepath.Join("..", "..", "generated", "testspec", "spec_index.json")

// pinnedReadtestCensus is the pinned readtest function census as the selection
// code enumerates it: every inventoried function name, with the comparator
// groups its upstream rows carry.
type pinnedReadtestCensus struct {
	compareGroups map[string]map[string]bool
}

func (c pinnedReadtestCensus) has(function string) bool {
	_, ok := c.compareGroups[function]
	return ok
}

func (c pinnedReadtestCensus) hasCompareGroup(function, group string) bool {
	return c.compareGroups[function][group]
}

func (c pinnedReadtestCensus) compareGroupsOf(function string) []string {
	return sortedStringSet(c.compareGroups[function])
}

func (c pinnedReadtestCensus) functionNames() []string {
	names := make([]string, 0, len(c.compareGroups))
	for name := range c.compareGroups {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func pinnedReadtestFunctionCensus(t *testing.T) pinnedReadtestCensus {
	t.Helper()
	spec, err := LoadGenerated(readtestProfileInventoryRelPath)
	if err != nil {
		t.Fatalf("load generated spec index: %v", err)
	}
	if len(spec.ReadtestProfileInventory) == 0 {
		t.Fatalf("generated spec index %s carries no readtest profile inventory; this gate lost its census", readtestProfileInventoryRelPath)
	}
	census := pinnedReadtestCensus{compareGroups: make(map[string]map[string]bool)}
	for _, profile := range spec.ReadtestProfileInventory {
		for _, fn := range profile.Functions {
			if fn.Function == "" {
				t.Fatalf("readtest profile %q has an inventory row with no function name", profile.Profile)
			}
			if census.compareGroups[fn.Function] == nil {
				census.compareGroups[fn.Function] = make(map[string]bool)
			}
			if fn.CompareGroup != "" {
				census.compareGroups[fn.Function][fn.CompareGroup] = true
			}
		}
	}
	if len(census.compareGroups) == 0 {
		t.Fatalf("readtest profile inventory in %s lists no functions; this gate lost its census", readtestProfileInventoryRelPath)
	}
	return census
}

func sortedNameKeyedListNames(lists map[string]readtestNameKeyedList) []string {
	names := make([]string, 0, len(lists))
	for name := range lists {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func sortedStringMapKeys(m map[string]string) []string {
	names := make([]string, 0, len(m))
	for name := range m {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// readtestNameKeyedListSwitchNames returns the function names declared in the
// case arms of the named readtest_spec.go list predicate. The predicate must
// stay a single switch on its own parameter with string-literal case arms:
// anything else is a shape this gate cannot enumerate, and is reported as such
// rather than passed as an empty list.
func readtestNameKeyedListSwitchNames(t *testing.T, predicateName string) []string {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, readtestSpecSourceFile, nil, parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("parse %s: %v", readtestSpecSourceFile, err)
	}

	var decl *ast.FuncDecl
	for _, d := range file.Decls {
		fn, ok := d.(*ast.FuncDecl)
		if ok && fn.Recv == nil && fn.Name.Name == predicateName {
			decl = fn
			break
		}
	}
	if decl == nil {
		t.Fatalf("%s declares no func %s; the list moved and this gate no longer reads it", readtestSpecSourceFile, predicateName)
	}
	if decl.Type.Params == nil || len(decl.Type.Params.List) != 1 || len(decl.Type.Params.List[0].Names) != 1 {
		t.Fatalf("%s does not take exactly one named parameter; this gate expects an exact-name predicate", predicateName)
	}
	param := decl.Type.Params.List[0].Names[0].Name

	if decl.Body == nil || len(decl.Body.List) != 1 {
		t.Fatalf("%s body is not a single statement; this gate expects one switch over %s and cannot enumerate any other shape", predicateName, param)
	}
	switchStmt, ok := decl.Body.List[0].(*ast.SwitchStmt)
	if !ok {
		t.Fatalf("%s body statement is %T, not a switch; this gate cannot enumerate the listed names", predicateName, decl.Body.List[0])
	}
	if switchStmt.Init != nil {
		t.Fatalf("%s switch carries an init statement; this gate expects a plain switch over %s", predicateName, param)
	}
	tag, ok := switchStmt.Tag.(*ast.Ident)
	if !ok || tag.Name != param {
		t.Fatalf("%s switches on %v, not on its parameter %s; the switched value must be the function name for exact-name entries to be enumerable", predicateName, switchStmt.Tag, param)
	}

	var names []string
	for _, stmt := range switchStmt.Body.List {
		clause, ok := stmt.(*ast.CaseClause)
		if !ok {
			t.Fatalf("%s switch contains %T, not a case clause", predicateName, stmt)
		}
		for _, expr := range clause.List {
			lit, ok := expr.(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				t.Fatalf("%s case arm at %s is not a string literal; entries must be literal function names to be checkable against the census", predicateName, fset.Position(expr.Pos()))
			}
			name, err := strconv.Unquote(lit.Value)
			if err != nil {
				t.Fatalf("%s case arm %s at %s is not an unquotable string: %v", predicateName, lit.Value, fset.Position(expr.Pos()), err)
			}
			names = append(names, name)
		}
	}
	return names
}

// readtestPackageFuncDecls indexes every package-level func declared in the
// non-test sources of package testgen, so a callee's signature can be resolved
// wherever in the package it lives.
func readtestPackageFuncDecls(t *testing.T) map[string]*ast.FuncDecl {
	t.Helper()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read package directory: %v", err)
	}
	fset := token.NewFileSet()
	decls := make(map[string]*ast.FuncDecl)
	sources := 0
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, name, nil, parser.SkipObjectResolution)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		sources++
		for _, d := range file.Decls {
			fn, ok := d.(*ast.FuncDecl)
			if !ok || fn.Recv != nil {
				continue
			}
			decls[fn.Name.Name] = fn
		}
	}
	if sources == 0 {
		t.Fatal("no non-test Go sources found in the package directory; this gate lost its scan set")
	}
	return decls
}

// stringPredicateCallersReachableFrom walks package-local identifier calls from
// the given roots and returns every reached func(string) bool, mapped to the
// sorted set of reached functions that call it. Traversal enters the predicates
// too, so a predicate composed only inside another predicate is still reached.
func stringPredicateCallersReachableFrom(decls map[string]*ast.FuncDecl, roots []string) map[string][]string {
	callers := make(map[string]map[string]bool)
	visited := make(map[string]bool)
	queue := append([]string(nil), roots...)
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if visited[current] {
			continue
		}
		visited[current] = true
		decl, ok := decls[current]
		if !ok || decl.Body == nil {
			continue
		}
		for _, callee := range packageLocalCallees(decls, decl) {
			if isStringPredicateSignature(decls[callee]) {
				if callers[callee] == nil {
					callers[callee] = make(map[string]bool)
				}
				callers[callee][current] = true
			}
			if !visited[callee] {
				queue = append(queue, callee)
			}
		}
	}

	out := make(map[string][]string, len(callers))
	for predicate, set := range callers {
		out[predicate] = sortedStringSet(set)
	}
	return out
}

// packageLocalCallees returns the package-level funcs the given function calls
// through a plain identifier, deduplicated.
func packageLocalCallees(decls map[string]*ast.FuncDecl, fn *ast.FuncDecl) []string {
	seen := make(map[string]bool)
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		ident, ok := call.Fun.(*ast.Ident)
		if !ok {
			return true
		}
		if _, ok := decls[ident.Name]; ok {
			seen[ident.Name] = true
		}
		return true
	})
	return sortedStringSet(seen)
}

// isStringPredicateSignature reports whether the declaration takes one string
// and returns one bool — the shape a name-keyed list predicate must have, and
// the shape every excepted string predicate shares with it.
func isStringPredicateSignature(decl *ast.FuncDecl) bool {
	if decl == nil {
		return false
	}
	params := decl.Type.Params
	results := decl.Type.Results
	if params == nil || len(params.List) != 1 || len(params.List[0].Names) > 1 {
		return false
	}
	if results == nil || len(results.List) != 1 || len(results.List[0].Names) > 1 {
		return false
	}
	paramType, ok := params.List[0].Type.(*ast.Ident)
	if !ok || paramType.Name != "string" {
		return false
	}
	resultType, ok := results.List[0].Type.(*ast.Ident)
	return ok && resultType.Name == "bool"
}
