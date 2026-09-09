package portprovenance

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// mathBigImportPath is the quoted import path of the one standard-library
// package that has no counterpart anywhere in pinned Intel BID C: the C library
// computes exclusively in fixed-width 32/64/128/192/256-bit limbs and macros, so
// a bidgo file reaching for arbitrary-precision arithmetic is computing
// something the C source does not.
const mathBigImportPath = `"math/big"`

// nonCanonicalCoreEntry pins one bidgo file whose operation core is
// bid754-authored rather than a mechanical port, against what the file actually
// contains.
type nonCanonicalCoreEntry struct {
	// headerDeclaration is the file's own statement that part of it departs
	// from the pinned C, quoted from the porting header. It is matched as a
	// whitespace-normalized substring of the comment block above the package
	// clause, so deleting or rewriting the self-declaration fails the gate
	// instead of silently leaving the file looking like a plain port.
	headerDeclaration string
	// coreFunctions are the package-level functions that make up the
	// bid754-authored core. Each must be declared in that file and must
	// reference math/big, so an entry cannot be satisfied by naming an
	// incidental helper.
	coreFunctions []string
	// reason states which Intel predecessor the core would have to map to and
	// why it does not.
	reason string
}

// nonCanonicalCoreFiles is the third provenance axis: bidgo files whose porting
// header points at real pinned Intel sources — so the file-provenance check
// passes and the exported-function census matches by name — while the operation
// core inside them is bid754-authored. nonIntelOriginFiles cannot hold these,
// because an entry there must carry zero Intel source references and these files
// legitimately carry them for their special-value and wrapper regions. Without
// this axis the transitional debt that docs/ARCHITECTURE_SPEC.md defines has no
// mechanical inventory at all: both existing axes pass on a file whose entire
// finite-value path replaces the Intel algorithm.
//
// Entries are checked against file contents by
// TestBidgoNonCanonicalCoreRegistryMatchesFileContents; unregistered files
// carrying the detectable signature are surfaced by
// TestBidgoNonCanonicalCoreSignatureFilesAreRegisteredOrExcepted. Membership is
// bounded by that same signature in both directions: a file this package cannot
// check mechanically is refused rather than admitted on the strength of its
// entry text, so a core written in the port's own fixed-width limbs has to
// extend the scan before it can be registered.
var nonCanonicalCoreFiles = map[string]nonCanonicalCoreEntry{}

// nonCanonicalSignatureExceptions lists bidgo files that carry the detectable
// non-canonical signature but whose use of it is a local semantics-preserving
// representation substitution of a specific pinned Intel construct, not a
// replacement core. An entry is an assertion that the substituted C construct
// was identified by name; it is not a place to park an unjudged file. Closed
// world in both directions with nonCanonicalCoreFiles: a file may be in exactly
// one of the two, an entry naming a file that no longer carries the signature
// fails as stale, and an entry without a written reason fails.
var nonCanonicalSignatureExceptions = map[string]string{}

// parseBidgoFile parses one bidgo implementation file with comments retained.
func parseBidgoFile(t *testing.T, fileName string) *ast.File {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), filepath.Join(bidgoDirRel, fileName), nil, parser.ParseComments|parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("parse %s: %v", fileName, err)
	}
	return file
}

// collapseWhitespace reduces every run of whitespace to a single space so a
// pinned declaration keeps matching after the porting header is re-wrapped.
func collapseWhitespace(text string) string {
	return strings.Join(strings.Fields(text), " ")
}

// bidgoFileHeaderText returns the comment text above the package clause,
// whitespace-collapsed. The porting header is not the file's doc comment (a
// blank line separates the two in this package), so every comment group that
// starts before the package keyword is taken.
func bidgoFileHeaderText(file *ast.File) string {
	var parts []string
	for _, group := range file.Comments {
		if group.Pos() >= file.Package {
			break
		}
		parts = append(parts, group.Text())
	}
	return collapseWhitespace(strings.Join(parts, " "))
}

// mathBigLocalName reports whether the file imports math/big and under which
// local name. The import is matched by path, so renaming the import does not
// hide the file from the signature scan; the returned name is only used for the
// per-function attribution below.
func mathBigLocalName(file *ast.File) (string, bool) {
	for _, spec := range file.Imports {
		if spec.Path == nil || spec.Path.Value != mathBigImportPath {
			continue
		}
		if spec.Name != nil {
			return spec.Name.Name, true
		}
		return "big", true
	}
	return "", false
}

// bidgoFuncsReferencingPackage returns the package-level functions whose
// declaration (signature or body) selects through the given local package name.
func bidgoFuncsReferencingPackage(file *ast.File, local string) map[string]bool {
	found := map[string]bool{}
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv != nil {
			continue
		}
		ast.Inspect(fn, func(node ast.Node) bool {
			sel, ok := node.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == local {
				found[fn.Name.Name] = true
			}
			return true
		})
	}
	return found
}

// bidgoFilePackageFuncs returns the package-level function names declared in one
// bidgo file, exported or not.
func bidgoFilePackageFuncs(file *ast.File) map[string]bool {
	names := map[string]bool{}
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Recv == nil {
			names[fn.Name.Name] = true
		}
	}
	return names
}

func sortedNonCanonicalCoreFileNames() []string {
	names := make([]string, 0, len(nonCanonicalCoreFiles))
	for name := range nonCanonicalCoreFiles {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func sortedSignatureExceptionNames() []string {
	names := make([]string, 0, len(nonCanonicalSignatureExceptions))
	for name := range nonCanonicalSignatureExceptions {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// TestBidgoNonCanonicalCoreRegistryMatchesFileContents holds every registry
// entry against the file it names, so the inventory cannot drift into a list of
// names: the file must exist and must still be on this axis (Intel references
// present, and not also claimed by nonIntelOriginFiles), the pinned
// self-declaration must still be in the porting header, the file must still
// carry the math/big signature the axis is detectable by, and every registered
// core function must still be declared in that file and still reference
// math/big, so an entry cannot be satisfied by naming an incidental helper after
// the real core moved.
//
// The signature requirement runs in both directions on purpose. Without it a
// file with no signature at all could be registered as debt — over-claiming a
// file that is not debt, and leaving a stale entry standing after a refactor
// moves the core off math/big — because nothing in the entry would be checkable
// beyond its own text.
func TestBidgoNonCanonicalCoreRegistryMatchesFileContents(t *testing.T) {
	existing := map[string]bool{}
	for _, name := range bidgoImplementationFiles(t) {
		existing[name] = true
	}

	for _, fileName := range sortedNonCanonicalCoreFileNames() {
		entry := nonCanonicalCoreFiles[fileName]
		if !existing[fileName] {
			t.Errorf("nonCanonicalCoreFiles entry %q does not exist in bidgo; remove the stale entry", fileName)
			continue
		}
		if strings.TrimSpace(entry.reason) == "" {
			t.Errorf("nonCanonicalCoreFiles[%q] has an empty reason; state which Intel predecessor the core would have to map to and why it does not", fileName)
		}
		if _, alsoNonIntel := nonIntelOriginFiles[fileName]; alsoNonIntel {
			t.Errorf("nonCanonicalCoreFiles entry %q is also in nonIntelOriginFiles; a file with no Intel origin at all belongs on that axis only", fileName)
		}
		if len(sourceReferences(t, fileName)) == 0 {
			t.Errorf("nonCanonicalCoreFiles entry %q references no Intel source; this axis is for files that do carry a porting header, so move it to nonIntelOriginFiles", fileName)
		}

		file := parseBidgoFile(t, fileName)
		if strings.TrimSpace(entry.headerDeclaration) == "" {
			t.Errorf("nonCanonicalCoreFiles[%q] pins no header declaration; quote the file's own statement that it departs from the pinned C", fileName)
		} else if !strings.Contains(bidgoFileHeaderText(file), collapseWhitespace(entry.headerDeclaration)) {
			t.Errorf("bidgo file %q no longer declares %q in its porting header; the file and the registry disagree about whether its core is bid754-authored", fileName, collapseWhitespace(entry.headerDeclaration))
		}

		if len(entry.coreFunctions) == 0 {
			t.Errorf("nonCanonicalCoreFiles[%q] names no core function; the entry claims debt without saying where it is", fileName)
		}
		declared := bidgoFilePackageFuncs(file)
		local, importsMathBig := mathBigLocalName(file)
		referencing := map[string]bool{}
		if importsMathBig {
			referencing = bidgoFuncsReferencingPackage(file, local)
		} else {
			t.Errorf("nonCanonicalCoreFiles entry %q does not import math/big, the one signature this axis can check a core by; remove the entry if the file is not debt or no longer holds it, and if a bid754-authored core written in the port's own fixed-width limbs has been found, extend the signature scan to reach it instead of registering an unverifiable entry", fileName)
		}
		seen := map[string]bool{}
		for _, fn := range entry.coreFunctions {
			if seen[fn] {
				t.Errorf("nonCanonicalCoreFiles[%q] lists core function %s twice", fileName, fn)
				continue
			}
			seen[fn] = true
			if !declared[fn] {
				t.Errorf("nonCanonicalCoreFiles[%q] lists core function %s, which is not a package-level function declared in that file; remove the stale name or point it at the function that carries the core", fileName, fn)
				continue
			}
			if importsMathBig && !referencing[fn] {
				t.Errorf("nonCanonicalCoreFiles[%q] lists core function %s, which no longer uses %s in a file that still imports math/big; if the core moved, re-judge the entry against the function that now holds it", fileName, fn, local)
			}
		}
	}
}

// TestBidgoNonCanonicalCoreSignatureFilesAreRegisteredOrExcepted supplies the
// detection direction the registry cannot: a newly written bid754-authored core
// is invisible to a hand-maintained list. Importing math/big is the one
// mechanically checkable signature of computing outside the pinned C's
// fixed-width limb model, so every bidgo file that does must land in exactly one
// of the two maps, both closed worlds.
//
// Boundary. This surfaces files, not functions: the per-function debt list stays
// hand-registered above, checked for existence and math/big use but never proven
// exhaustive within its file, and no reachability census records which exported
// operations route into a registered core, so routing one more operation through
// existing debt is not surfaced here. The signature also bounds what the axis
// can hold at all: a bid754-authored core written with the same uint64/
// BID_UINT128 limbs the port uses elsewhere, or against math/bits, produces no
// signature, so it is invisible to this scan and is refused by the registry
// rather than admitted as an unverifiable entry — reaching it means extending
// the scan, and until that happens such a core leaves no trace on this axis.
// Conversely math/big alone does not prove debt, which is why an exception
// carrying the substituted C construct exists. Per-function
// attribution is syntactic (a selector through an identifier equal to the
// import's local name, not a type-checked reference), so a local variable
// shadowing that name inside a function counts as a math/big reference. Scope is
// the bidgo tree that bidgoImplementationFiles walks: _test.go files and
// bidgoSubdirExclusions are outside it, so the cexport snapshot's copy of the
// same core is left to its own gate.
func TestBidgoNonCanonicalCoreSignatureFilesAreRegisteredOrExcepted(t *testing.T) {
	carriesSignature := map[string]bool{}
	for _, fileName := range bidgoImplementationFiles(t) {
		if _, ok := mathBigLocalName(parseBidgoFile(t, fileName)); !ok {
			continue
		}
		carriesSignature[fileName] = true

		_, registered := nonCanonicalCoreFiles[fileName]
		reason, excepted := nonCanonicalSignatureExceptions[fileName]
		switch {
		case registered && excepted:
			t.Errorf("bidgo file %q is both registered in nonCanonicalCoreFiles and excepted in nonCanonicalSignatureExceptions; pick one", fileName)
		case registered:
		case excepted:
			t.Logf("excepted math/big use in %s: %s", fileName, reason)
		default:
			t.Errorf("bidgo file %q imports math/big, which pinned Intel BID C has no counterpart for, but is in neither nonCanonicalCoreFiles nor nonCanonicalSignatureExceptions; register the bid754-authored core it holds, or except it naming the pinned C construct its math/big use locally substitutes", fileName)
		}
	}

	for _, fileName := range sortedSignatureExceptionNames() {
		if strings.TrimSpace(nonCanonicalSignatureExceptions[fileName]) == "" {
			t.Errorf("nonCanonicalSignatureExceptions[%q] carries no written reason; name the pinned C construct the math/big use substitutes", fileName)
		}
		if !carriesSignature[fileName] {
			t.Errorf("nonCanonicalSignatureExceptions entry %q no longer imports math/big (or no longer exists); remove the stale exception", fileName)
		}
	}
}
