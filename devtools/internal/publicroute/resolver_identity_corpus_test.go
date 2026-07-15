// Resolver identity corpus: the gate-of-the-gate.
//
// The routing gate reads the real tree, so an ambiguous routing pattern cannot
// be placed in the tree as real API. Instead this corpus type-checks synthetic
// in-memory Go snippets with the SAME routeImporter the gate uses, and asserts
// the resolver's verdict on each. It is the permanent regression check for the go/types
// object-identity resolvers: if a future edit relaxes a resolver back toward
// go/ast name matching, the corresponding case fails.
//
// Each case is a known identifier-resolution boundary named in the promotion
// design, with the pre-migration go/ast verdict (the MISS these resolvers were
// promoted to close) recorded in the comment and the post-migration go/types
// verdict (the CATCH) asserted:
//
//  1. import alias            — old missed the aliased port call; new reaches it
//  2. unrelated package alias — old counted `bidgo.X` from a foreign package;
//     new rejects it by package-path identity
//  3. package-function shadow  — old followed the shadowed NAME to the real
//     helper; new resolves the call to the local var
//  4. same-named local in a bidgo body — old recorded a false callee; new
//     excludes the local by object identity
//  5. honest control           — the real plumbing shape; must still resolve to
//     exactly {Bid64Div}
//
// The snippets are self-contained valid Go so the strict type-check accepts
// them; no synthetic case ever touches a real public file.
package publicroute

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"testing"
)

// bidgoImport is the real Go mechanical port, imported so that cases 1/3/5
// resolve to genuine bidgo *types.Func objects (identity against
// bidgoImportPath is the whole point).
const bidgoImport = `"` + bidgoImportPath + `"`

type corpusEnv struct {
	t    *testing.T
	fset *token.FileSet
	ri   *routeImporter
}

func newCorpusEnv(t *testing.T) *corpusEnv {
	fset := token.NewFileSet()
	return &corpusEnv{t: t, fset: fset, ri: newRouteImporter(fset)}
}

// typecheck parses and type-checks src as package "corpus", strict: a
// snippet that does not type-check (unused import, undefined symbol, ...) fails
// the corpus rather than being resolved on partial information.
func (e *corpusEnv) typecheck(name, src string) (*types.Package, *types.Info, *ast.File) {
	e.t.Helper()
	file, err := parser.ParseFile(e.fset, name+".go", src, parser.ParseComments)
	if err != nil {
		e.t.Fatalf("%s: parse corpus snippet: %v", name, err)
	}
	info := newTypesInfo()
	var errs []error
	conf := types.Config{
		Importer: e.ri,
		Error:    func(er error) { errs = append(errs, er) },
	}
	pkg, cerr := conf.Check("corpus", e.fset, []*ast.File{file}, info)
	if cerr != nil && len(errs) == 0 {
		errs = append(errs, cerr)
	}
	if len(errs) > 0 {
		e.t.Fatalf("%s: corpus snippet must be valid Go: %v", name, errors.Join(errs...))
	}
	return pkg, info, file
}

func (e *corpusEnv) newResolver(pkg *types.Package, info *types.Info, file *ast.File) *routeResolver {
	_, _, funcByObj := buildFuncMaps([]*ast.File{file}, info, pkg)
	return &routeResolver{info: info, pkg: pkg, funcByObj: funcByObj}
}

// findDecl returns the first function declaration named name (method or not).
func findDecl(t *testing.T, file *ast.File, name string) *ast.FuncDecl {
	t.Helper()
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == name {
			return fn
		}
	}
	t.Fatalf("corpus snippet has no function %q", name)
	return nil
}

func funcNameSet(s map[*types.Func]bool) map[string]bool {
	out := map[string]bool{}
	for f := range s {
		out[f.Name()] = true
	}
	return out
}

func TestResolverIdentityCorpus(t *testing.T) {
	env := newCorpusEnv(t)

	// --- Case 1: import alias ---------------------------------------------
	// A ...Port helper reaches the real port through an ALIASED import. The
	// retired go/ast resolver keyed off the literal identifier text "bidgo" and
	// reached nothing; the new resolver resolves the selector to the bidgo
	// *types.Func by object identity and reaches Bid64Div.
	t.Run("1_import_alias", func(t *testing.T) {
		src := `package corpus
import port ` + bidgoImport + `
type Decimal64BID uint64
func decimal64BIDDivPort(d, other Decimal64BID) Decimal64BID {
	return Decimal64BID(port.Bid64Div(uint64(d), uint64(other), 0))
}
`
		pkg, info, file := env.typecheck("case1", src)
		entry := findDecl(t, file, "decimal64BIDDivPort")
		reach := env.newResolver(pkg, info, file).reachablePortFuncs(entry)
		t.Logf("new(go/types) reachable=%v (retired go/ast reached [])", sortedSet(reach))
		if !reach["Bid64Div"] {
			t.Errorf("aliased port call not caught: reachable=%v, want it to contain Bid64Div", sortedSet(reach))
		}
	})

	// --- Case 2: unrelated package alias ----------------------------------
	// An unrelated package is imported UNDER THE NAME `bidgo`. The retired
	// resolver counted any `bidgo.X` selector as a port call (reached
	// OnesCount64); the new resolver checks the selected object's package path
	// and reaches nothing.
	t.Run("2_unrelated_package_alias", func(t *testing.T) {
		src := `package corpus
import bidgo "math/bits"
type Decimal64BID uint64
func decimal64BIDDivPort(d, other Decimal64BID) Decimal64BID {
	return Decimal64BID(bidgo.OnesCount64(uint64(d)) + bidgo.OnesCount64(uint64(other)))
}
`
		pkg, info, file := env.typecheck("case2", src)
		entry := findDecl(t, file, "decimal64BIDDivPort")
		reach := env.newResolver(pkg, info, file).reachablePortFuncs(entry)
		t.Logf("new(go/types) reachable=%v (retired go/ast reached [OnesCount64])", sortedSet(reach))
		if len(reach) != 0 {
			t.Errorf("unrelated package alias not rejected: reachable=%v, want empty", sortedSet(reach))
		}
	})

	// --- Case 3: package-function shadow ----------------------------------
	// A public method appears to route through a ...Port helper but a local
	// closure of the same name intercepts the call and returns identity. The
	// retired resolver followed the NAME to the real package helper and reached
	// Bid64Div; the new resolver resolves the call to the local variable and does
	// not follow it.
	t.Run("3_package_func_shadow", func(t *testing.T) {
		src := `package corpus
import ` + bidgoImport + `
type Decimal64BID uint64
func decimal64BIDDivPort(d, other Decimal64BID) Decimal64BID {
	return Decimal64BID(bidgo.Bid64Div(uint64(d), uint64(other), 0))
}
func (d Decimal64BID) Div(other Decimal64BID) Decimal64BID {
	decimal64BIDDivPort := func(a, b Decimal64BID) Decimal64BID { return a }
	return decimal64BIDDivPort(d, other)
}
`
		pkg, info, file := env.typecheck("case3", src)
		entry := findDecl(t, file, "Div")
		reach := env.newResolver(pkg, info, file).reachablePortFuncs(entry)
		t.Logf("new(go/types) reachable=%v (retired go/ast reached [Bid64Div])", sortedSet(reach))
		if reach["Bid64Div"] {
			t.Errorf("shadowed helper not caught: reachable=%v, want it to NOT contain Bid64Div", sortedSet(reach))
		}
	})

	// --- Case 4: same-named local in a bidgo body -------------------------
	// A bidgo body appears to delegate to its oracle-verified sibling, but a
	// local closure of the sibling's name intercepts the call. The retired
	// body-callee resolver keyed off the name and recorded Bid64AddWithFlags; the
	// new resolver resolves the call to the local variable and excludes it.
	t.Run("4_bidgo_body_local_shadow", func(t *testing.T) {
		src := `package corpus
func Bid64AddWithFlags(x, y uint64) (uint64, uint32) { return x + y, 0 }
func Bid64Add(x, y uint64) uint64 {
	Bid64AddWithFlags := func(a, b uint64) uint64 { return a }
	return Bid64AddWithFlags(x, y)
}
`
		pkg, info, file := env.typecheck("case4", src)
		entry := findDecl(t, file, "Bid64Add")
		callees := funcNameSet(env.newResolver(pkg, info, file).bidgoCalleeFuncs(entry))
		t.Logf("new(go/types) callees=%v (retired go/ast recorded [Bid64AddWithFlags])", sortedSet(callees))
		if callees["Bid64AddWithFlags"] {
			t.Errorf("local shadow not excluded: callees=%v, want it to NOT contain Bid64AddWithFlags", sortedSet(callees))
		}
	})

	// --- Case 5: honest control -------------------------------------------
	// The real plumbing shape (method -> ...Port helper -> bidgo.Fn) must resolve
	// to exactly {Bid64Div}: the promotion must not regress honest routing.
	t.Run("5_honest_control", func(t *testing.T) {
		src := `package corpus
import ` + bidgoImport + `
type Decimal64BID uint64
func (d Decimal64BID) ToUint64() uint64 { return uint64(d) }
func decimal64BIDDivPort(d, other Decimal64BID) Decimal64BID {
	return Decimal64BID(bidgo.Bid64Div(d.ToUint64(), other.ToUint64(), 0))
}
func (d Decimal64BID) Div(other Decimal64BID) Decimal64BID { return decimal64BIDDivPort(d, other) }
`
		pkg, info, file := env.typecheck("case5", src)
		entry := findDecl(t, file, "Div")
		reach := env.newResolver(pkg, info, file).reachablePortFuncs(entry)
		t.Logf("new(go/types) reachable=%v", sortedSet(reach))
		if len(reach) != 1 || !reach["Bid64Div"] {
			t.Errorf("honest routing regressed: reachable=%v, want exactly {Bid64Div}", sortedSet(reach))
		}
	})
}
