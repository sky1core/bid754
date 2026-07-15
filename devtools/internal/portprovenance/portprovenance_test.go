// Package portprovenance enforces the "Go mechanical port of Intel BID C"
// architecture split structurally: every implementation file and every
// exported function in bid754-go/internal/bidgo must either trace back to the
// pinned Intel BID C tree or carry a documented non-Intel-origin reason.
//
// Four checks, all exhaustive in the tablecrosscheck style (no empty
// reasons, no unused exclusion entries):
//
//  1. File provenance: every non-test bidgo .go file must reference at least
//     one Intel C/H source file in its porting comments, and every referenced
//     file name must actually exist in the pinned Intel tree
//     (devtools/third_party/intel_dfp/LIBRARY/src). This is not a
//     format-string check — a header that names a C file which does not
//     exist upstream fails. Files with no Intel origin (pure Go glue,
//     bid754-authored helpers) must appear in nonIntelOriginFiles with a
//     concrete reason.
//
//  2. Exported-function census: every exported package-level bidgo function
//     must correspond by normalized name to an extern function declared in
//     the pinned Intel bid_functions.h, as extracted into the checked-in
//     devtools/generated/json/intel_dfp_symbols.json (kept in sync with the
//     pinned header by csymbols TestGeneratedSymbolsStayInSync). Functions
//     without an Intel counterpart must appear in nonIntelExportedFunctions
//     with a concrete reason. This makes "add a new hand-written Rust-style
//     alternate implementation surface inside bidgo" impossible without
//     leaving a documented trace.
//
//  3. Exported-type census: the exact exported type names and declaration
//     files are pinned. A convenience value type cannot grow inside the
//     C-shaped port boundary unnoticed.
//
//  4. Receiver-method absence: every non-test implementation declaration is
//     scanned and any receiver method fails. Mechanical-port operations stay
//     package-level functions that go2rs can translate without dropping API.
//
// The Intel source tree itself is downloaded (not committed); when it is
// absent the existence check is skipped with a reason while the reference
// format and census checks still run against checked-in inputs.
package portprovenance

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

const (
	bidgoDirRel        = "../../../bid754-go/internal/bidgo"
	intelSrcDirRel     = "../../third_party/intel_dfp/LIBRARY/src"
	intelSymbolsRel    = "../../generated/json/intel_dfp_symbols.json"
	intelDownloadHint  = "run devtools/scripts/setup_generation_inputs.sh to download the pinned Intel tree"
	withFlagsSuffix    = "WithFlags"
	referencePatternGo = `[A-Za-z][A-Za-z0-9_]*\.[ch]\b`
)

// nonIntelOriginFiles lists the bidgo implementation files that intentionally
// carry no Intel C/H source reference, with the reason there is none. Every
// entry must exist, must still contain zero Intel source references (a stale
// entry that gained references must be removed so the references get
// verified), and must carry a non-empty reason.
var nonIntelOriginFiles = map[string]string{
	"bid32_exports.go":         "bid754-authored exported wrapper glue around the mechanically ported bid32 functions defined in files that carry the Intel references",
	"tables_round_const128.go": "bid754-authored init-time construction of bid_round_const_table_128 from bid_power10_table_128; the value surface is anchored by the tablecrosscheck exclusion entry for bid_round_const_table_128",
	"types.go":                 "Go-side type and rounding-mode constant definitions for the bidgo package; no ported Intel logic",
}

// nonIntelExportedFunctions lists exported package-level bidgo functions with
// no matching extern in the pinned Intel bid_functions.h, with the reason.
// Every entry must exist as an exported bidgo function, must still fail to
// match an Intel extern (a stale entry that started matching must be
// removed), and must carry a non-empty reason.
var nonIntelExportedFunctions = map[string]string{
	"BID_normalize":      "ported from the static helper BID_normalize in Intel bid_inline_add.h; internal static helpers are not extern declarations in bid_functions.h",
	"Bid32FromStringRaw": "raw-uint32 and explicit-status naming variant of the mechanically ported bid32_from_string; Bid32FromString is reserved for the one-result wrapper",
	"Bid32ToStringRaw":   "raw-uint32 naming variant of the ported bid32_to_string",
	"Bid32IsInf32":       "uint32-typed naming variant of the ported bid32_isInf",
	"Bid32IsNaN32":       "uint32-typed naming variant of the ported bid32_isNaN",
	"Bid32IsZero32":      "uint32-typed naming variant of the ported bid32_isZero",
}

// exportedBidgoTypes is a deliberately small closed-world set. The bidgo
// package is the C-shaped mechanical-port boundary, so adding another exported
// value/helper type is an architecture change rather than an incidental Go API
// choice.
var expectedExportedBidgoTypes = map[string]string{
	"BID_UINT128":  "internal.go",
	"BID_UINT192":  "internal.go",
	"BID_UINT256":  "internal.go",
	"BID_UINT320":  "internal.go",
	"BID_UINT384":  "internal.go",
	"BID_UINT512":  "internal.go",
	"DEC_DIGITS":   "tables_intconv.go",
	"RoundingMode": "types.go",
}

// bidgoSubdirExclusions lists subdirectories under bid754-go/internal/bidgo
// that are intentionally outside the mechanical-port provenance/census surface,
// with the reason. bidgoImplementationFiles walks the whole bidgo tree and
// includes .go files in any subdirectory, EXCEPT these — so a hand-written
// alternate implementation cannot escape the provenance and census checks by
// living one package directory below the root (the previous non-recursive scan
// skipped every subdirectory). Exhaustive in the nonIntelOriginFiles style:
// TestBidgoSubdirExclusionsExist fails on a stale entry (an exclusion that
// names no real subdirectory) or an empty reason, and any subdirectory NOT
// listed here is walked into rather than ignored.
var bidgoSubdirExclusions = map[string]string{
	"cexport": "inactive C ABI compatibility snapshot (its own go.mod, isolated build tag); it is not part of the Go mechanical port surface and is checked independently by `make verify-cexport-disabled`",
}

var referencePattern = regexp.MustCompile(referencePatternGo)

// normalizeSourceName canonicalizes a C/H file name for existence matching:
// lowercase with underscores stripped. The porting headers reference Intel
// files by their canonical pinned spelling (e.g. bid64_div.c), but matching
// stays spelling-insensitive as independent verification: a stray CamelCase or
// underscore variant (Bid64Div.c) still has to resolve to a real pinned file
// rather than silently passing. Porting-header text is required for licensing;
// only the pinned file name a header points at is checked here.
func normalizeSourceName(name string) string {
	return strings.ReplaceAll(strings.ToLower(name), "_", "")
}

// normalizeFunctionName canonicalizes a function name for Go-to-C matching:
// the Go WithFlags suffix (explicit-flags variant of the same ported
// function) is dropped, then everything is lowercased with non-alphanumerics
// stripped, so Go CamelCase (Bid128CopySign) meets Intel's mixed
// snake/camel C names (bid128_copySign).
func normalizeFunctionName(name string) string {
	name = strings.TrimSuffix(name, withFlagsSuffix)
	var b strings.Builder
	for _, r := range strings.ToLower(name) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// bidgoImplementationFiles returns the non-test .go files anywhere in the bidgo
// package tree, as slash-separated paths relative to the package root (a
// top-level file stays its bare base name, so nonIntelOriginFiles keys are
// unchanged). The tree is walked recursively: every subdirectory is descended
// into so an implementation file cannot avoid provenance/census by sitting one
// directory below the root, EXCEPT the directories in bidgoSubdirExclusions
// (currently only the disabled cexport module, which has its own gate).
func bidgoImplementationFiles(t *testing.T) []string {
	t.Helper()
	root := filepath.Clean(bidgoDirRel)
	var files []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if path == root {
				return nil
			}
			rel, relErr := filepath.Rel(root, path)
			if relErr != nil {
				return relErr
			}
			if _, excluded := bidgoSubdirExclusions[filepath.ToSlash(rel)]; excluded {
				return fs.SkipDir
			}
			return nil
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			return nil
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		files = append(files, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		t.Fatalf("walk bidgo dir: %v", err)
	}
	sort.Strings(files)
	if len(files) == 0 {
		t.Fatal("no bidgo implementation files found; wrong path?")
	}
	return files
}

// bidgoSubdirs returns the slash-separated relative paths of every
// subdirectory under the bidgo package root (recursively, without applying any
// exclusion), so the exclusion list can be validated for completeness.
func bidgoSubdirs(t *testing.T) map[string]bool {
	t.Helper()
	root := filepath.Clean(bidgoDirRel)
	dirs := map[string]bool{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() || path == root {
			return nil
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		dirs[filepath.ToSlash(rel)] = true
		return nil
	})
	if err != nil {
		t.Fatalf("walk bidgo dir: %v", err)
	}
	return dirs
}

// TestBidgoSubdirExclusionsExist keeps bidgoSubdirExclusions exhaustive: an
// exclusion that names no real subdirectory (a stale skip that would silently
// re-open the hole if a same-named directory ever reappears) or that carries an
// empty reason fails. New non-excluded subdirectories are covered by the
// provenance and census tests via bidgoImplementationFiles' recursive walk.
func TestBidgoSubdirExclusionsExist(t *testing.T) {
	dirs := bidgoSubdirs(t)
	for name, reason := range bidgoSubdirExclusions {
		if strings.TrimSpace(reason) == "" {
			t.Errorf("bidgoSubdirExclusions[%q] has an empty reason; document why the subdirectory is outside the port surface", name)
		}
		if !dirs[name] {
			t.Errorf("bidgoSubdirExclusions entry %q does not exist as a bidgo subdirectory; remove the stale entry", name)
		}
	}
}

// sourceReferences extracts the Intel C/H file names referenced in the
// comments of one bidgo file. Only comment text is scanned so identifiers in
// code cannot fake or hide a reference.
func sourceReferences(t *testing.T, fileName string) []string {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filepath.Join(bidgoDirRel, fileName), nil, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse %s: %v", fileName, err)
	}
	seen := map[string]bool{}
	var refs []string
	for _, group := range file.Comments {
		for _, match := range referencePattern.FindAllString(group.Text(), -1) {
			if !seen[match] {
				seen[match] = true
				refs = append(refs, match)
			}
		}
	}
	sort.Strings(refs)
	return refs
}

func TestBidgoFilesDeclareIntelSourceProvenance(t *testing.T) {
	files := bidgoImplementationFiles(t)
	fileSet := map[string]bool{}
	referencesByFile := map[string][]string{}
	for _, name := range files {
		fileSet[name] = true
		referencesByFile[name] = sourceReferences(t, name)
	}

	// Exhaustive set over the exclusion list.
	for name, reason := range nonIntelOriginFiles {
		if strings.TrimSpace(reason) == "" {
			t.Errorf("nonIntelOriginFiles[%q] has an empty reason; document why the file has no Intel origin", name)
		}
		if !fileSet[name] {
			t.Errorf("nonIntelOriginFiles entry %q does not exist in bidgo; remove the stale entry", name)
			continue
		}
		if refs := referencesByFile[name]; len(refs) > 0 {
			t.Errorf("nonIntelOriginFiles entry %q now references Intel sources %v; remove the entry so the references are verified", name, refs)
		}
	}

	// Every implementation file either references pinned Intel sources or is
	// a documented non-Intel-origin file.
	for _, name := range files {
		if _, excluded := nonIntelOriginFiles[name]; excluded {
			continue
		}
		if len(referencesByFile[name]) == 0 {
			t.Errorf("bidgo file %q has no Intel C/H source reference in its comments; add the porting header or a documented nonIntelOriginFiles entry", name)
		}
	}

	// Reference format: extraction already guarantees the shape, but assert
	// it explicitly so the pattern itself cannot drift into accepting paths
	// or wildcards.
	format := regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]*\.[ch]$`)
	for _, name := range files {
		for _, ref := range referencesByFile[name] {
			if !format.MatchString(ref) {
				t.Errorf("bidgo file %q references %q which is not a plain Intel source file name", name, ref)
			}
		}
	}

	// Existence: every referenced name must resolve (spelling-insensitively)
	// to a real file in the pinned Intel tree. This is what stops a
	// plausible-looking but fabricated porting header.
	t.Run("ReferencesExistInPinnedIntelTree", func(t *testing.T) {
		entries, err := os.ReadDir(intelSrcDirRel)
		if err != nil {
			t.Skipf("pinned Intel tree not present at %s (%v); %s — reference existence not verified in this environment", intelSrcDirRel, err, intelDownloadHint)
		}
		pinned := map[string]bool{}
		for _, entry := range entries {
			if !entry.IsDir() {
				pinned[normalizeSourceName(entry.Name())] = true
			}
		}
		if len(pinned) == 0 {
			t.Fatalf("pinned Intel tree at %s is empty; %s", intelSrcDirRel, intelDownloadHint)
		}
		for _, name := range files {
			for _, ref := range referencesByFile[name] {
				if !pinned[normalizeSourceName(ref)] {
					t.Errorf("bidgo file %q references Intel source %q which does not exist in the pinned tree %s", name, ref, intelSrcDirRel)
				}
			}
		}
	})
}

// intelExternFunctionNames loads the extern function names extracted from the
// pinned Intel bid_functions.h. The JSON artifact is checked in and kept in
// sync with the pinned header by csymbols TestGeneratedSymbolsStayInSync, so
// this census cannot be satisfied by editing a Go-side list.
func intelExternFunctionNames(t *testing.T) map[string]bool {
	t.Helper()
	raw, err := os.ReadFile(intelSymbolsRel)
	if err != nil {
		t.Fatalf("read %s: %v", intelSymbolsRel, err)
	}
	var doc struct {
		Symbols []struct {
			Name string `json:"name"`
		} `json:"symbols"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("decode %s: %v", intelSymbolsRel, err)
	}
	if len(doc.Symbols) == 0 {
		t.Fatalf("%s contains no symbols", intelSymbolsRel)
	}
	names := map[string]bool{}
	for _, symbol := range doc.Symbols {
		names[normalizeFunctionName(symbol.Name)] = true
	}
	return names
}

// exportedBidgoFunctions returns the exported package-level function names
// declared in the non-test bidgo files. Receiver methods are rejected by the
// separate zero-method census below rather than being accepted outside this
// function census.
func exportedBidgoFunctions(t *testing.T) map[string]string {
	t.Helper()
	fset := token.NewFileSet()
	names := map[string]string{}
	for _, fileName := range bidgoImplementationFiles(t) {
		file, err := parser.ParseFile(fset, filepath.Join(bidgoDirRel, fileName), nil, parser.SkipObjectResolution)
		if err != nil {
			t.Fatalf("parse %s: %v", fileName, err)
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv != nil || !fn.Name.IsExported() {
				continue
			}
			names[fn.Name.Name] = fileName
		}
	}
	if len(names) == 0 {
		t.Fatal("no exported bidgo functions found; wrong path?")
	}
	return names
}

func exportedBidgoTypes(t *testing.T) map[string]string {
	t.Helper()
	fset := token.NewFileSet()
	types := map[string]string{}
	for _, fileName := range bidgoImplementationFiles(t) {
		file, err := parser.ParseFile(fset, filepath.Join(bidgoDirRel, fileName), nil, parser.SkipObjectResolution)
		if err != nil {
			t.Fatalf("parse %s: %v", fileName, err)
		}
		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.TYPE {
				continue
			}
			for _, spec := range gen.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if ok && typeSpec.Name.IsExported() {
					types[typeSpec.Name.Name] = fileName
				}
			}
		}
	}
	return types
}

func receiverMethods(t *testing.T) []string {
	t.Helper()
	fset := token.NewFileSet()
	var methods []string
	for _, fileName := range bidgoImplementationFiles(t) {
		file, err := parser.ParseFile(fset, filepath.Join(bidgoDirRel, fileName), nil, parser.SkipObjectResolution)
		if err != nil {
			t.Fatalf("parse %s: %v", fileName, err)
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if ok && fn.Recv != nil {
				methods = append(methods, fn.Name.Name+" ("+fileName+")")
			}
		}
	}
	sort.Strings(methods)
	return methods
}

func TestBidgoExportedTypesMatchClosedWorld(t *testing.T) {
	actual := exportedBidgoTypes(t)
	for name, wantFile := range expectedExportedBidgoTypes {
		gotFile, ok := actual[name]
		if !ok {
			t.Errorf("required exported bidgo type %s is missing", name)
			continue
		}
		if gotFile != wantFile {
			t.Errorf("exported bidgo type %s is declared in %s, want %s", name, gotFile, wantFile)
		}
	}
	for name, fileName := range actual {
		if _, ok := expectedExportedBidgoTypes[name]; !ok {
			t.Errorf("unexpected exported bidgo type %s (%s); the mechanical-port boundary has a closed-world type census", name, fileName)
		}
	}
}

func TestBidgoDeclaresNoReceiverMethods(t *testing.T) {
	for _, method := range receiverMethods(t) {
		t.Errorf("bidgo receiver method %s is outside the C-shaped mechanical-port boundary", method)
	}
}

func TestBidgoExportedFunctionsMatchIntelExterns(t *testing.T) {
	externs := intelExternFunctionNames(t)
	exported := exportedBidgoFunctions(t)

	// Exhaustive set over the exclusion list.
	for name, reason := range nonIntelExportedFunctions {
		if strings.TrimSpace(reason) == "" {
			t.Errorf("nonIntelExportedFunctions[%q] has an empty reason; document why the function has no Intel extern counterpart", name)
		}
		if _, ok := exported[name]; !ok {
			t.Errorf("nonIntelExportedFunctions entry %q is not an exported bidgo function; remove the stale entry", name)
			continue
		}
		if externs[normalizeFunctionName(name)] {
			t.Errorf("nonIntelExportedFunctions entry %q now matches an Intel extern; remove the entry so the match is enforced", name)
		}
	}

	// Every exported function corresponds to a pinned Intel extern or is a
	// documented exception. A new exported bidgo function that implements
	// something Intel does not declare must be added to the exclusion list
	// with its reason, leaving a reviewable trace.
	var unmatched []string
	for name, fileName := range exported {
		if _, excluded := nonIntelExportedFunctions[name]; excluded {
			continue
		}
		if !externs[normalizeFunctionName(name)] {
			unmatched = append(unmatched, name+" ("+fileName+")")
		}
	}
	sort.Strings(unmatched)
	for _, entry := range unmatched {
		t.Errorf("exported bidgo function %s has no matching extern in the pinned Intel bid_functions.h and no nonIntelExportedFunctions entry", entry)
	}
}
