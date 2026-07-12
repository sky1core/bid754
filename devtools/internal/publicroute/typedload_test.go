// Typed loader and object-identity resolvers for the public-API routing gate.
//
// Checks 2, 3, and the decTest port-mode adapter value-flow resolve identifiers
// and calls by go/types OBJECT IDENTITY rather than by syntactic name. This
// closes the identifier-resolution ambiguity that pure go/ast leaves open:
// import aliases, unrelated package aliases, package-function-name shadowing,
// same-named locals inside bidgo bodies, and foreign receivers can no longer make code
// that only LOOKS like it routes through the Go mechanical port pass the gate.
//
// The type-check covers exactly the two source packages the public runtime path
// compiles from: the module-root package `bid754` under Config A (the default
// build: bid754_native off, cgo off) and its one internal dependency
// `internal/bidgo` (pure Go, stdlib-only). Both are type-checked from source
// with go/importer's "source" compiler for stdlib, following the proven recipe
// in devtools/internal/tablecrosscheck/tablecrosscheck_test.go. The load is
// strict (any type error fails the gate), single (sync.Once), and shared
// by every promoted check.
//
// What stays go/ast (documented residual, see the package doc in
// publicroute_test.go): the census enumeration (check 1), the cgo check and
// native-only-symbol scan (check 4), the FFI-surface extraction and native
// (cgo && bid754_native) adapter cgo-reachability walk (their files
// `import "C"`, which go/types cannot type-check from source), and the
// flagless-variant equivalence gate pair extraction (a bidgo external test
// package). These surfaces are bounded structurally (byte-pinned generated
// artifacts and exhaustive file/target lists), not by object identity.
package publicroute

import (
	"errors"
	"fmt"
	"go/ast"
	"go/build/constraint"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

const (
	// rootImportPath is the module-root public package under test.
	rootImportPath = "github.com/sky1core/bid754/bid754-go"
	// bidgoImportPath is the canonical Go mechanical port. Object identity
	// against this exact path (not the identifier text "bidgo") is what makes
	// import-alias and unrelated-package ambiguity impossible.
	bidgoImportPath = "github.com/sky1core/bid754/bid754-go/internal/bidgo"
)

// ---------------------------------------------------------------------------
// routeImporter: stdlib via the source compiler, bidgo type-checked from source
// once. Any other non-stdlib path is a strict error (Config A is verified
// to import only stdlib + bidgo; an unexpected internal import must fail loudly,
// never be silently delegated to the stdlib importer).
// ---------------------------------------------------------------------------

type routeImporter struct {
	fset *token.FileSet
	std  types.Importer

	once       sync.Once
	bidgoPkg   *types.Package
	bidgoInfo  *types.Info
	bidgoFiles []*ast.File
	bidgoErr   error
}

func newRouteImporter(fset *token.FileSet) *routeImporter {
	return &routeImporter{fset: fset, std: importer.ForCompiler(fset, "source", nil)}
}

func (ri *routeImporter) Import(path string) (*types.Package, error) {
	if path == bidgoImportPath {
		ri.once.Do(ri.checkBidgo)
		return ri.bidgoPkg, ri.bidgoErr
	}
	if !isStdlibPath(path) {
		return nil, fmt.Errorf("routeImporter: unexpected non-stdlib import %q; Config A is verified to import only stdlib + bidgo, so this must reject unresolved input rather than resolve silently", path)
	}
	return ri.std.Import(path)
}

func (ri *routeImporter) checkBidgo() {
	files, err := parseBidgoFiles(ri.fset)
	if err != nil {
		ri.bidgoErr = err
		return
	}
	info := newTypesInfo()
	var errs []error
	conf := types.Config{
		Importer: ri,
		Error:    func(e error) { errs = append(errs, e) },
	}
	pkg, cerr := conf.Check(bidgoImportPath, ri.fset, files, info)
	if cerr != nil && len(errs) == 0 {
		errs = append(errs, cerr)
	}
	if len(errs) > 0 {
		ri.bidgoErr = fmt.Errorf("type-check bidgo from source: %w", errors.Join(errs...))
		return
	}
	ri.bidgoPkg = pkg
	ri.bidgoInfo = info
	ri.bidgoFiles = files
}

// isStdlibPath reports whether path is a standard-library import path. Stdlib
// paths never contain a dot in their first segment; module paths always do.
func isStdlibPath(path string) bool {
	first := path
	if i := strings.IndexByte(path, '/'); i >= 0 {
		first = path[:i]
	}
	return !strings.Contains(first, ".")
}

func newTypesInfo() *types.Info {
	return &types.Info{
		Types:      map[ast.Expr]types.TypeAndValue{},
		Defs:       map[*ast.Ident]types.Object{},
		Uses:       map[*ast.Ident]types.Object{},
		Selections: map[*ast.SelectorExpr]*types.Selection{},
		Implicits:  map[ast.Node]types.Object{},
	}
}

// ---------------------------------------------------------------------------
// File selection and parsing (non-*testing.T so the load can be a plain func)
// ---------------------------------------------------------------------------

func listRootGoFiles() ([]string, error) {
	entries, err := os.ReadDir(publicSrcDirRel)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", publicSrcDirRel, err)
	}
	var files []string
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		files = append(files, name)
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no non-test .go files under %s", publicSrcDirRel)
	}
	return files, nil
}

// buildConstraintOf mirrors fileBuildConstraint without a *testing.T so the
// typed load can run outside a test body.
func buildConstraintOf(file *ast.File) (constraint.Expr, error) {
	for _, cg := range file.Comments {
		if cg.Pos() >= file.Package {
			continue
		}
		for _, c := range cg.List {
			if constraint.IsGoBuild(c.Text) {
				expr, err := constraint.Parse(c.Text)
				if err != nil {
					return nil, fmt.Errorf("parse build constraint %q: %w", c.Text, err)
				}
				return expr, nil
			}
		}
	}
	return nil, nil
}

// parseConfigARootFiles parses the module-root non-test files that participate
// in the default public runtime build (Config A: bid754_native off, cgo off) -
// the untagged files plus the no-cgo stubs, excluding the bid754_native oracle
// files. This is the exact file set the public runtime path compiles, and the
// same set the go/ast checks select via requiresNative.
func parseConfigARootFiles(fset *token.FileSet) ([]*ast.File, error) {
	names, err := listRootGoFiles()
	if err != nil {
		return nil, err
	}
	var files []*ast.File
	for _, name := range names {
		file, err := parser.ParseFile(fset, filepath.Join(publicSrcDirRel, name), nil, parser.ParseComments)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", name, err)
		}
		expr, err := buildConstraintOf(file)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		if requiresNative(expr) {
			continue // bid754_native oracle path, excluded from the public build
		}
		files = append(files, file)
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no Config A root files parsed")
	}
	return files, nil
}

// parseBidgoFiles parses the direct non-test .go files of the bidgo package
// (the cexport subpackage keeps its own go.mod and is a different package, so
// it is not part of this compilation unit).
func parseBidgoFiles(fset *token.FileSet) ([]*ast.File, error) {
	entries, err := os.ReadDir(bidgoDirRel)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", bidgoDirRel, err)
	}
	var files []*ast.File
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, filepath.Join(bidgoDirRel, name), nil, parser.ParseComments)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", name, err)
		}
		files = append(files, file)
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no bidgo source files under %s", bidgoDirRel)
	}
	return files, nil
}

// ---------------------------------------------------------------------------
// typedBuild: the cached, shared typed load
// ---------------------------------------------------------------------------

type typedBuild struct {
	fset *token.FileSet

	rootPkg           *types.Package
	rootInfo          *types.Info
	rootFiles         []*ast.File
	rootFuncsByName   map[string]*ast.FuncDecl // receiver-less package funcs, keyed by name
	rootMethodsByName map[string]*ast.FuncDecl // "Recv.Method"
	rootFuncsByObj    map[*types.Func]*ast.FuncDecl

	bidgoPkg         *types.Package
	bidgoInfo        *types.Info
	bidgoFiles       []*ast.File
	bidgoFuncsByName map[string]*ast.FuncDecl
	bidgoFuncsByObj  map[*types.Func]*ast.FuncDecl
}

var (
	typedBuildOnce sync.Once
	typedBuildVal  *typedBuild
	typedBuildErr  error
)

// loadTypedPublicBuild type-checks Config A + bidgo from source exactly once and
// returns the shared result. Strict: any type error (a genuinely broken
// tree, an undefined native-only reference included in the public build, a new
// unexpected import) fails the gate here.
func loadTypedPublicBuild(t *testing.T) *typedBuild {
	t.Helper()
	typedBuildOnce.Do(func() { typedBuildVal, typedBuildErr = buildTypedPublicBuild() })
	if typedBuildErr != nil {
		t.Fatalf("load typed public build: %v", typedBuildErr)
	}
	return typedBuildVal
}

func buildTypedPublicBuild() (*typedBuild, error) {
	fset := token.NewFileSet()
	ri := newRouteImporter(fset)

	// Type-check bidgo first (also primes the importer cache the root check
	// consumes). A bidgo type error surfaces here.
	bidgoPkg, err := ri.Import(bidgoImportPath)
	if err != nil {
		return nil, err
	}

	rootFiles, err := parseConfigARootFiles(fset)
	if err != nil {
		return nil, err
	}
	rootInfo := newTypesInfo()
	var rootErrs []error
	conf := types.Config{
		Importer: ri,
		Error:    func(e error) { rootErrs = append(rootErrs, e) },
	}
	rootPkg, cerr := conf.Check(rootImportPath, fset, rootFiles, rootInfo)
	if cerr != nil && len(rootErrs) == 0 {
		rootErrs = append(rootErrs, cerr)
	}
	if len(rootErrs) > 0 {
		return nil, fmt.Errorf("Config A type-check: %w", errors.Join(rootErrs...))
	}

	rootFuncsByName, rootMethodsByName, rootFuncsByObj := buildFuncMaps(rootFiles, rootInfo, rootPkg)
	bidgoFuncsByName, _, bidgoFuncsByObj := buildFuncMaps(ri.bidgoFiles, ri.bidgoInfo, bidgoPkg)

	return &typedBuild{
		fset:              fset,
		rootPkg:           rootPkg,
		rootInfo:          rootInfo,
		rootFiles:         rootFiles,
		rootFuncsByName:   rootFuncsByName,
		rootMethodsByName: rootMethodsByName,
		rootFuncsByObj:    rootFuncsByObj,
		bidgoPkg:          bidgoPkg,
		bidgoInfo:         ri.bidgoInfo,
		bidgoFiles:        ri.bidgoFiles,
		bidgoFuncsByName:  bidgoFuncsByName,
		bidgoFuncsByObj:   bidgoFuncsByObj,
	}, nil
}

// buildFuncMaps indexes a package's function declarations. Receiver-less
// package functions are keyed both by name (entry-point lookup, which the
// verification/exception tables address by string) and by *types.Func object (edge
// following, which resolves calls by identity). Methods are keyed "Recv.Method"
// by name for entry-point lookup only; reachability never follows into methods.
func buildFuncMaps(files []*ast.File, info *types.Info, pkg *types.Package) (funcsByName, methodsByName map[string]*ast.FuncDecl, funcsByObj map[*types.Func]*ast.FuncDecl) {
	funcsByName = map[string]*ast.FuncDecl{}
	methodsByName = map[string]*ast.FuncDecl{}
	funcsByObj = map[*types.Func]*ast.FuncDecl{}
	for _, file := range files {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			if fn.Recv == nil {
				funcsByName[fn.Name.Name] = fn
				if obj, ok := info.Defs[fn.Name].(*types.Func); ok {
					funcsByObj[obj] = fn
				}
				continue
			}
			if recv := receiverTypeName(fn.Recv); recv != "" {
				methodsByName[recv+"."+fn.Name.Name] = fn
			}
		}
	}
	return funcsByName, methodsByName, funcsByObj
}

// ---------------------------------------------------------------------------
// routeResolver: object-identity resolution over one package's bodies
// ---------------------------------------------------------------------------

// routeResolver resolves identifiers and calls inside the bodies of one package
// (root or bidgo) by go/types object identity. It carries the package's
// type-check info, the package object itself (so "local package function" and
// "value-type receiver" are identity questions, not name questions), and the
// package's receiver-less func declarations for edge following.
type routeResolver struct {
	info      *types.Info
	pkg       *types.Package
	funcByObj map[*types.Func]*ast.FuncDecl
}

func (tb *typedBuild) rootResolver() *routeResolver {
	return &routeResolver{info: tb.rootInfo, pkg: tb.rootPkg, funcByObj: tb.rootFuncsByObj}
}

func (tb *typedBuild) bidgoResolver() *routeResolver {
	return &routeResolver{info: tb.bidgoInfo, pkg: tb.bidgoPkg, funcByObj: tb.bidgoFuncsByObj}
}

// localFunc returns the *types.Func an identifier resolves to IFF it is a
// receiver-less package-level function of r.pkg. A shadowing local variable,
// parameter, function literal bound to a name, or a function from another
// package resolves to a different object (or a non-func) and returns nil, so
// the edge is not followed.
func (r *routeResolver) localFunc(ident *ast.Ident) *types.Func {
	fn, ok := r.info.Uses[ident].(*types.Func)
	if !ok {
		return nil
	}
	if fn.Pkg() != r.pkg {
		return nil
	}
	// Package-level, receiver-less: the package scope maps its name to exactly
	// this object. A method lives in a type's method set, not the package
	// scope; a local shadow already failed the *types.Func assertion above.
	if r.pkg.Scope().Lookup(fn.Name()) != fn {
		return nil
	}
	return fn
}

// bidgoFuncName returns the name of the bidgo package function a selector call
// resolves to, or "" if the selector is not a receiver-less bidgo package
// function. Identity against bidgoImportPath is what rejects an unrelated
// package aliased as `bidgo` and accepts the real port imported under any
// alias.
func (r *routeResolver) bidgoFuncName(sel *ast.SelectorExpr) string {
	fn, ok := r.info.Uses[sel.Sel].(*types.Func)
	if !ok {
		return ""
	}
	if fn.Pkg() == nil || fn.Pkg().Path() != bidgoImportPath {
		return ""
	}
	if sig, ok := fn.Type().(*types.Signature); ok && sig.Recv() != nil {
		return "" // a method on a bidgo type is not a package-qualified port call
	}
	return fn.Name()
}

// reachablePortFuncs walks start's body, following package-level function calls
// (the wrapper -> ...Port helper plumbing) by object identity and collecting
// every bidgo package function reachable through them. Value-method calls and
// foreign-package calls are leaves: the port routing in this package always
// flows through package-level functions, so a wrapper that instead routed
// through a value method would leave its reachable set short of the pinned
// expectation and fail check 2 rather than pass silently.
func (r *routeResolver) reachablePortFuncs(start *ast.FuncDecl) map[string]bool {
	result := map[string]bool{}
	visited := map[*types.Func]bool{}
	stack := []*ast.FuncDecl{start}
	for len(stack) > 0 {
		fn := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if fn.Body == nil {
			continue
		}
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			switch f := call.Fun.(type) {
			case *ast.Ident:
				if obj := r.localFunc(f); obj != nil && !visited[obj] {
					if decl := r.funcByObj[obj]; decl != nil {
						visited[obj] = true
						stack = append(stack, decl)
					}
				}
			case *ast.SelectorExpr:
				if name := r.bidgoFuncName(f); name != "" {
					result[name] = true
				}
			}
			return true
		})
	}
	return result
}

// bidgoCalleeFuncs returns the set of receiver-less bidgo package functions
// fn's body calls by bare identifier (the local-call plumbing inside bidgo).
// Same-named locals, parameters, and function-value arguments resolve to
// non-package-func objects and are excluded by identity.
func (r *routeResolver) bidgoCalleeFuncs(fn *ast.FuncDecl) map[*types.Func]bool {
	out := map[*types.Func]bool{}
	if fn == nil || fn.Body == nil {
		return out
	}
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if ident, ok := call.Fun.(*ast.Ident); ok {
			if obj := r.localFunc(ident); obj != nil {
				out[obj] = true
			}
		}
		return true
	})
	return out
}

func (r *routeResolver) lookupFunc(name string) *types.Func {
	fn, _ := r.pkg.Scope().Lookup(name).(*types.Func)
	return fn
}

// ---------------------------------------------------------------------------
// Typed value-flow tracking (object-keyed)
// ---------------------------------------------------------------------------

// valueSourceEventsTyped is the go/types object-keyed successor of the
// name-keyed valueSourceEvents: it tracks, per resolved object, how many times
// the object receives a value from the target call versus any other source. A
// shadowing re-declaration produces a DIFFERENT object, so its bindings can no
// longer merge with the outer identifier's - the class of confusion the
// name-keyed tracker could not tell apart.
//
// Analysis boundary (unchanged in kind from the name-keyed tracker):
// go/types resolves identifier and call identity, but a value moved outside
// tracked bindings - notably through unsafe.Pointer bit reinterpretation, which
// has no function/method object - is still opaque. That is an accepted residual
// documented on the package, not a hole this tracker claims to close.
type valueSourceEventsTyped struct {
	fromTarget   map[types.Object]int
	fromOther    map[types.Object]int
	localDefined map[types.Object]bool
}

func (ev valueSourceEventsTyped) derivedFromTarget(obj types.Object) bool {
	return obj != nil && ev.fromTarget[obj] == 1 && ev.fromOther[obj] == 0
}

// objOf returns the object a binding-or-use identifier denotes: a defining
// occurrence (fresh :=, var, param, named result, range/type-switch binding)
// is in Defs; a re-declaration/assignment or a reference is in Uses.
func (r *routeResolver) objOf(id *ast.Ident) types.Object {
	if id == nil {
		return nil
	}
	if o := r.info.Defs[id]; o != nil {
		return o
	}
	return r.info.Uses[id]
}

// collectValueSourceEvents is the object-keyed port of collectValueSourceEvents
// (see the name-keyed original's contract). isTargetCall identifies the port
// call whose result is being traced; allResults selects whether every result of
// a multi-value target binding counts as target-derived (adapter: value+flags
// both compared) or only the first (equivalence proof: the value result only).
func (r *routeResolver) collectValueSourceEvents(fn *ast.FuncDecl, isTargetCall func(*ast.CallExpr) bool, allResults bool) valueSourceEventsTyped {
	targetResult := 0
	if allResults {
		targetResult = -1
	}
	return r.collectValueSourceResultEvents(fn, isTargetCall, targetResult)
}

// collectValueSourceResultEvents is the result-position-selective form used
// when only one result of a multi-value call is admissible provenance.
// targetResult >= 0 selects that zero-based result; -1 selects every result.
func (r *routeResolver) collectValueSourceResultEvents(fn *ast.FuncDecl, isTargetCall func(*ast.CallExpr) bool, targetResult int) valueSourceEventsTyped {
	ev := valueSourceEventsTyped{
		fromTarget:   map[types.Object]int{},
		fromOther:    map[types.Object]int{},
		localDefined: map[types.Object]bool{},
	}
	if fn == nil {
		return ev
	}
	if fn.Type.Params != nil {
		for _, field := range fn.Type.Params.List {
			for _, name := range field.Names {
				if obj := r.objOf(name); obj != nil {
					ev.localDefined[obj] = true
				}
			}
		}
	}
	// Named results start life as zero values: an ambiguous origin.
	if fn.Type.Results != nil {
		for _, field := range fn.Type.Results.List {
			for _, name := range field.Names {
				if obj := r.objOf(name); obj != nil {
					ev.localDefined[obj] = true
					ev.fromOther[obj]++
				}
			}
		}
	}
	if fn.Body == nil {
		return ev
	}
	bindIdent := func(expr ast.Expr, target bool) {
		id, ok := expr.(*ast.Ident)
		if !ok || id.Name == "_" {
			return
		}
		obj := r.objOf(id)
		if obj == nil {
			return
		}
		ev.localDefined[obj] = true
		if target {
			ev.fromTarget[obj]++
		} else {
			ev.fromOther[obj]++
		}
	}
	isTarget := func(expr ast.Expr) bool {
		call, ok := expr.(*ast.CallExpr)
		return ok && isTargetCall(call)
	}
	selectedTargetResult := func(index, rhsCount int, fromTargetCall bool) bool {
		if !fromTargetCall {
			return false
		}
		if targetResult < 0 {
			return index == 0 || rhsCount == 1
		}
		if targetResult == 0 {
			return index == 0
		}
		return rhsCount == 1 && index == targetResult
	}
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		switch s := n.(type) {
		case *ast.AssignStmt:
			if len(s.Lhs) == 0 {
				return true
			}
			fromTargetCall := len(s.Rhs) >= 1 && isTarget(s.Rhs[0])
			for i, lhs := range s.Lhs {
				bindIdent(lhs, selectedTargetResult(i, len(s.Rhs), fromTargetCall))
			}
		case *ast.GenDecl:
			if s.Tok != token.VAR {
				return true
			}
			for _, spec := range s.Specs {
				vs, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				fromTargetCall := len(vs.Values) >= 1 && isTarget(vs.Values[0])
				for i, name := range vs.Names {
					if len(vs.Values) == 0 {
						bindIdent(name, false) // zero-value declaration: ambiguous origin
						continue
					}
					bindIdent(name, selectedTargetResult(i, len(vs.Values), fromTargetCall))
				}
			}
		case *ast.RangeStmt:
			if s.Key != nil {
				bindIdent(s.Key, false)
			}
			if s.Value != nil {
				bindIdent(s.Value, false)
			}
		case *ast.TypeSwitchStmt:
			// A type-switch variable `v := x.(type)` has one implicit object
			// PER case clause (go/types stores them in Implicits, not Defs on
			// the check ident), so objOf on the check would miss them. Mark
			// each as an other-origin local so a value moved through a type
			// switch is never mistaken for safe constant data (strict).
			for _, stmt := range s.Body.List {
				clause, ok := stmt.(*ast.CaseClause)
				if !ok {
					continue
				}
				if obj := r.info.Implicits[clause]; obj != nil {
					ev.localDefined[obj] = true
					ev.fromOther[obj]++
				}
			}
		}
		return true
	})
	return ev
}

// firstResultFlowsFromCall is the object-keyed port: callee identity is a
// *types.Func matched through info.Uses, so a same-named unrelated function in call position
// is not the callee, and a shadowed return identifier is a different object than
// the tracked binding.
func (r *routeResolver) firstResultFlowsFromCall(fn *ast.FuncDecl, callee *types.Func) bool {
	if fn == nil || fn.Body == nil || callee == nil {
		return false
	}
	isCalleeCallExpr := func(call *ast.CallExpr) bool {
		id, ok := call.Fun.(*ast.Ident)
		return ok && r.info.Uses[id] == callee
	}
	isCalleeCall := func(expr ast.Expr) bool {
		call, ok := expr.(*ast.CallExpr)
		return ok && isCalleeCallExpr(call)
	}
	ev := r.collectValueSourceEvents(fn, isCalleeCallExpr, false)
	sawReturn := false
	shapeOK := true
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		ret, ok := n.(*ast.ReturnStmt)
		if !ok {
			return true
		}
		sawReturn = true
		if len(ret.Results) == 0 {
			shapeOK = false // bare return (named results): origin not provable
			return true
		}
		if isCalleeCall(ret.Results[0]) {
			return true
		}
		if id, ok := ret.Results[0].(*ast.Ident); ok && ev.derivedFromTarget(r.info.Uses[id]) {
			return true
		}
		shapeOK = false
		return true
	})
	return sawReturn && shapeOK
}

// proveVariantEquivalence is the object-identity port of the name-based proof.
// It resolves fn and its WithFlags/flagless sibling to real bidgo *types.Func
// objects and proves value-path sharing through one of the three provable
// structures (see the name-based doc). Callee identity via bidgoCalleeFuncs and
// value flow via firstResultFlowsFromCall are both object-based, so a same-named
// alternate cannot stand in for the surface sibling or the shared delegate.
func (r *routeResolver) proveVariantEquivalence(fn string, covered func(string) bool) (string, bool) {
	var sibling string
	if strings.HasSuffix(fn, "WithFlags") {
		sibling = strings.TrimSuffix(fn, "WithFlags")
	} else {
		sibling = fn + "WithFlags"
	}
	if !covered(sibling) {
		return "", false
	}
	fnObj, sibObj := r.lookupFunc(fn), r.lookupFunc(sibling)
	if fnObj == nil || sibObj == nil {
		return "", false
	}
	fnDecl, sibDecl := r.funcByObj[fnObj], r.funcByObj[sibObj]
	if fnDecl == nil || sibDecl == nil {
		return "", false
	}
	fnCallees := r.bidgoCalleeFuncs(fnDecl)
	sibCallees := r.bidgoCalleeFuncs(sibDecl)
	if len(fnCallees) == 1 && fnCallees[sibObj] && r.firstResultFlowsFromCall(fnDecl, sibObj) {
		return fmt.Sprintf("returns the value of surface sibling %s (sole callee)", sibling), true
	}
	if sibCallees[fnObj] && r.firstResultFlowsFromCall(sibDecl, fnObj) {
		return fmt.Sprintf("surface sibling %s returns its value", sibling), true
	}
	if len(fnCallees) == 1 {
		var sole *types.Func
		for c := range fnCallees {
			sole = c
		}
		if sibCallees[sole] && r.firstResultFlowsFromCall(fnDecl, sole) && r.firstResultFlowsFromCall(sibDecl, sole) {
			return fmt.Sprintf("both return the value of shared delegate %s (surface sibling %s)", sole.Name(), sibling), true
		}
	}
	return "", false
}

// ---------------------------------------------------------------------------
// Typed decTest port-mode adapter value-flow
// ---------------------------------------------------------------------------

// valueTypeRecvName returns the module-root value type name for t (possibly a
// pointer), or "" when t is not Decimal32BID/Decimal64BID/Decimal128BID from
// r.pkg.
func (r *routeResolver) valueTypeRecvName(t types.Type) string {
	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}
	named, ok := t.(*types.Named)
	if !ok {
		return ""
	}
	obj := named.Obj()
	if obj.Pkg() != r.pkg || valueTypeWidths[obj.Name()] == "" {
		return ""
	}
	return obj.Name()
}

// isValueTypeMethodCall reports whether call is `x.<method>(...)` resolving,
// through info.Selections, to a real method of the required module-root value
// type. This is the object-identity replacement for the
// name-plus-"not-C"-receiver check: a foreign or wrong-width type that merely
// happens to expose a same-named method is rejected.
func (r *routeResolver) isValueTypeMethodCall(call *ast.CallExpr, valueType, method string) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != method {
		return false
	}
	seln, ok := r.info.Selections[sel]
	if !ok || seln.Kind() != types.MethodVal {
		return false
	}
	return r.valueTypeRecvName(seln.Recv()) == valueType
}

type decTestOperandParserSpec struct {
	parser   string
	delegate string
}

var decTestOperandParserSpecs = map[string]decTestOperandParserSpec{
	"Decimal32BID":  {parser: "parseDecTestDecimal32Operand", delegate: "parseDecimal32BIDPortMode"},
	"Decimal64BID":  {parser: "parseDecTestDecimal64Operand", delegate: "parseDecimal64BIDPortMode"},
	"Decimal128BID": {parser: "parseDecTestDecimal128Operand", delegate: "parseDecimal128BIDPortMode"},
}

// isVerifiedDecTestOperandParserCall recognizes only the generated decTest
// operand loaders whose complete multi-result return directly delegates to the
// width-matched mechanical from-string port helper, forwarding the normalized
// input and rounding parameter unchanged. The adapter call itself must take
// operand zero from its decTest case (or, in the isolated resolver identity corpus, a
// direct adapter parameter) and a case-derived rounding value. Operand
// conversion flags are legitimate supplemental provenance in an adapter's
// compared Flags field, but they must not count as evidence that the operation
// method itself is live.
func (r *routeResolver) isVerifiedDecTestOperandParserCall(
	call *ast.CallExpr,
	fn *ast.FuncDecl,
	valueType string,
	caseParam types.Object,
	roundingValues valueSourceEventsTyped,
) bool {
	id, ok := call.Fun.(*ast.Ident)
	if !ok {
		return false
	}
	parserObj, ok := r.info.Uses[id].(*types.Func)
	if !ok || parserObj.Pkg() != r.pkg {
		return false
	}
	spec, ok := decTestOperandParserSpecs[valueType]
	if !ok || parserObj.Name() != spec.parser || len(call.Args) != 2 {
		return false
	}
	delegateObj, ok := r.pkg.Scope().Lookup(spec.delegate).(*types.Func)
	if !ok {
		return false
	}
	decl := r.funcByObj[parserObj]
	if decl == nil || decl.Body == nil || len(decl.Body.List) != 1 {
		return false
	}
	ret, ok := decl.Body.List[0].(*ast.ReturnStmt)
	if !ok || len(ret.Results) != 1 {
		return false
	}
	delegateCall, ok := ret.Results[0].(*ast.CallExpr)
	if !ok {
		return false
	}
	delegateID, ok := delegateCall.Fun.(*ast.Ident)
	if !ok || r.info.Uses[delegateID] != delegateObj || len(delegateCall.Args) != 2 {
		return false
	}
	parserSig, ok := parserObj.Type().(*types.Signature)
	if !ok || parserSig.Params().Len() != 2 {
		return false
	}
	operandNormalizer, ok := r.pkg.Scope().Lookup("decTestOperandString").(*types.Func)
	if !ok {
		return false
	}
	normalizeCall, ok := delegateCall.Args[0].(*ast.CallExpr)
	if !ok || len(normalizeCall.Args) != 1 {
		return false
	}
	normalizeID, ok := normalizeCall.Fun.(*ast.Ident)
	if !ok || r.info.Uses[normalizeID] != operandNormalizer {
		return false
	}
	normalizeArg, ok := normalizeCall.Args[0].(*ast.Ident)
	if !ok || r.info.Uses[normalizeArg] != parserSig.Params().At(0) {
		return false
	}
	roundingArg, ok := delegateCall.Args[1].(*ast.Ident)
	if !ok || r.info.Uses[roundingArg] != parserSig.Params().At(1) {
		return false
	}
	if !r.isVerifiedDecTestOperandArgument(call.Args[0], fn, caseParam) {
		return false
	}
	adapterRounding, ok := call.Args[1].(*ast.Ident)
	if !ok {
		return false
	}
	roundingObj := r.objOf(adapterRounding)
	return roundingValues.derivedFromTarget(roundingObj) || (caseParam == nil && r.isFunctionParameter(fn, roundingObj))
}

func (r *routeResolver) isFunctionParameter(fn *ast.FuncDecl, obj types.Object) bool {
	if fn == nil || fn.Type.Params == nil || obj == nil {
		return false
	}
	for _, field := range fn.Type.Params.List {
		for _, name := range field.Names {
			if r.objOf(name) == obj {
				return true
			}
		}
	}
	return false
}

func (r *routeResolver) decTestCaseParameter(fn *ast.FuncDecl) types.Object {
	if fn == nil || fn.Type.Params == nil {
		return nil
	}
	for _, field := range fn.Type.Params.List {
		for _, name := range field.Names {
			obj := r.objOf(name)
			v, ok := obj.(*types.Var)
			if !ok {
				continue
			}
			named, ok := v.Type().(*types.Named)
			if ok && named.Obj().Pkg() == r.pkg && named.Obj().Name() == "decTestCase" {
				return obj
			}
		}
	}
	return nil
}

func (r *routeResolver) isVerifiedDecTestCaseRoundingCall(call *ast.CallExpr, caseParam types.Object) bool {
	if caseParam == nil || len(call.Args) != 1 {
		return false
	}
	id, ok := call.Fun.(*ast.Ident)
	if !ok {
		return false
	}
	callee, ok := r.pkg.Scope().Lookup("decTestCaseRoundingMode").(*types.Func)
	if !ok || r.info.Uses[id] != callee {
		return false
	}
	arg, ok := call.Args[0].(*ast.Ident)
	return ok && r.info.Uses[arg] == caseParam
}

// isVerifiedDecTestOperandArgument pins the unary Reduce evidence to operand
// zero of the adapter's decTestCase. Isolated resolver snippets without that
// project type may use a direct string parameter, but never a literal or a
// transformed expression.
func (r *routeResolver) isVerifiedDecTestOperandArgument(expr ast.Expr, fn *ast.FuncDecl, caseParam types.Object) bool {
	if caseParam == nil {
		id, ok := expr.(*ast.Ident)
		return ok && r.isFunctionParameter(fn, r.objOf(id))
	}
	index, ok := expr.(*ast.IndexExpr)
	if !ok {
		return false
	}
	lit, ok := index.Index.(*ast.BasicLit)
	if !ok || lit.Kind != token.INT || lit.Value != "0" {
		return false
	}
	sel, ok := index.X.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "Operands" {
		return false
	}
	root, ok := sel.X.(*ast.Ident)
	if !ok || r.info.Uses[root] != caseParam {
		return false
	}
	selection, ok := r.info.Selections[sel]
	return ok && selection.Kind() == types.FieldVal && selection.Obj().Name() == "Operands"
}

// adapterReturnsPortMethodValue is the object-identity port of the name-based
// adapter value-flow proof (see that doc for the full contract). The port call
// is identified by isValueTypeMethodCall (Selections receiver typing), and value
// provenance is tracked by object, not by name.
func (r *routeResolver) adapterReturnsPortMethodValue(fn *ast.FuncDecl, valueType, method string, comparedFields []string) string {
	if fn == nil || fn.Body == nil {
		return "adapter has no body"
	}
	isMethodCall := func(call *ast.CallExpr) bool { return r.isValueTypeMethodCall(call, valueType, method) }
	ev := r.collectValueSourceEvents(fn, isMethodCall, true)
	caseParam := r.decTestCaseParameter(fn)
	caseRoundingCall := func(call *ast.CallExpr) bool { return r.isVerifiedDecTestCaseRoundingCall(call, caseParam) }
	roundingValues := r.collectValueSourceEvents(fn, caseRoundingCall, false)
	operandParserCall := func(call *ast.CallExpr) bool {
		return r.isVerifiedDecTestOperandParserCall(call, fn, valueType, caseParam, roundingValues)
	}
	operandParseFlags := r.collectValueSourceResultEvents(fn, operandParserCall, 1)

	var usesDerived func(expr ast.Expr) bool
	usesDerived = func(expr ast.Expr) bool {
		uses := false
		ast.Inspect(expr, func(n ast.Node) bool {
			if id, ok := n.(*ast.Ident); ok && ev.derivedFromTarget(r.objOf(id)) {
				uses = true
			}
			if call, ok := n.(*ast.CallExpr); ok && isMethodCall(call) {
				uses = true
			}
			return true
		})
		return uses
	}

	var safe func(expr ast.Expr) bool
	safe = func(expr ast.Expr) bool {
		switch e := expr.(type) {
		case *ast.BasicLit:
			return true
		case *ast.Ident:
			if e.Name == "nil" || e.Name == "true" || e.Name == "false" {
				return true
			}
			obj := r.objOf(e)
			if ev.derivedFromTarget(obj) {
				return true
			}
			if operandParseFlags.derivedFromTarget(obj) {
				return true
			}
			// A local identifier that is not target-derived (parameter, other
			// call's result, ambiguous origin) is unsafe; a package-level
			// const/var (not a local object) is safe constant data.
			return obj == nil || !ev.localDefined[obj]
		case *ast.SelectorExpr:
			return safe(e.X)
		case *ast.CallExpr:
			if sel, ok := e.Fun.(*ast.SelectorExpr); ok {
				if !safe(sel.X) {
					return false
				}
			} else if _, ok := e.Fun.(*ast.Ident); !ok {
				return false
			}
			for _, arg := range e.Args {
				if !safe(arg) {
					return false
				}
			}
			return true
		case *ast.CompositeLit:
			for _, elt := range e.Elts {
				if !safe(elt) {
					return false
				}
			}
			return true
		case *ast.KeyValueExpr:
			return safe(e.Value)
		case *ast.ParenExpr:
			return safe(e.X)
		case *ast.UnaryExpr:
			return safe(e.X)
		case *ast.BinaryExpr:
			return safe(e.X) && safe(e.Y)
		case *ast.IndexExpr:
			return safe(e.X) && safe(e.Index)
		default:
			return false // strict: unhandled expression shapes are unproven
		}
	}

	sawSuccess := false
	fieldDerivedSeen := map[string]bool{}
	failure := ""
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		ret, ok := n.(*ast.ReturnStmt)
		if !ok || failure != "" {
			return failure == ""
		}
		if len(ret.Results) < 2 {
			failure = "adapter has a return without the (value, error) shape; success paths are not identifiable"
			return false
		}
		last, ok := ret.Results[len(ret.Results)-1].(*ast.Ident)
		if !ok || last.Name != "nil" {
			return true // error path: its value is never compared
		}
		sawSuccess = true
		lit, ok := ret.Results[0].(*ast.CompositeLit)
		if !ok {
			failure = "a success return's value is not a keyed composite literal; the compared fields cannot be traced (strict)"
			return false
		}
		fieldValues := map[string]ast.Expr{}
		for _, elt := range lit.Elts {
			kv, ok := elt.(*ast.KeyValueExpr)
			if !ok {
				failure = "a success return uses a positional composite literal; the compared fields cannot be traced (strict)"
				return false
			}
			key, ok := kv.Key.(*ast.Ident)
			if !ok {
				failure = "a success return uses a non-identifier field key; the compared fields cannot be traced (strict)"
				return false
			}
			fieldValues[key.Name] = kv.Value
		}
		for _, field := range comparedFields {
			value, present := fieldValues[field]
			if !present {
				continue // zero value: constant-equivalent, contributes no derivation
			}
			if !safe(value) {
				failure = fmt.Sprintf("a success return's compared field %q does not provably flow from the port method call (or from literals/constants derived with it)", field)
				return false
			}
			if usesDerived(value) {
				fieldDerivedSeen[field] = true
			}
		}
		return true
	})
	if failure != "" {
		return failure
	}
	if !sawSuccess {
		return "adapter has no success return; nothing flows to the comparison"
	}
	for _, field := range comparedFields {
		if !fieldDerivedSeen[field] {
			return fmt.Sprintf("no success return derives compared field %q from .%s(...); the port method call is dead with respect to the compared %s", field, method, field)
		}
	}
	return ""
}

// ---------------------------------------------------------------------------
// decTest dispatch entrypoint reference/flag-check resolution (object identity)
// ---------------------------------------------------------------------------

// rootFuncObj returns the receiver-less module-root package function object of
// the given name, or nil.
func (tb *typedBuild) rootFuncObj(name string) *types.Func {
	fn, _ := tb.rootPkg.Scope().Lookup(name).(*types.Func)
	return fn
}

// entrypointReferencesFunc reports whether the decTest dispatch entrypoint's
// body references the named adapter function BY OBJECT IDENTITY (adapter exec
// functions are passed by value, so this is a reference check). Resolving
// through info.Uses means a same-named unrelated identifier cannot satisfy the
// reference.
func entrypointReferencesFunc(t *testing.T, tb *typedBuild, name string) bool {
	t.Helper()
	entry := tb.rootFuncsByName[dectestEntrypoint]
	if entry == nil {
		t.Fatalf("decTest dispatch entrypoint %q not found in the public build; the decTest coverage evidence lost its anchor", dectestEntrypoint)
	}
	target := tb.rootFuncObj(name)
	if target == nil {
		t.Fatalf("adapter %q is not a module-root package function; the decTest coverage evidence lost its subject", name)
	}
	refs := false
	ast.Inspect(entry.Body, func(n ast.Node) bool {
		if id, ok := n.(*ast.Ident); ok && tb.rootInfo.Uses[id] == target {
			refs = true
		}
		return true
	})
	return refs
}

// dectestAdapterFlagCheckModes returns the flag-check mode identifiers the
// generated decTest dispatch entrypoint passes alongside the named adapter (the
// last argument of every dispatch call that references the adapter BY OBJECT
// IDENTITY). The caller derives the compared-field requirement from these modes
// instead of hardcoding it, so the requirement tracks the checked-in dispatch
// artifact. Mode identifiers stay name-valued: they are compared against a
// single known constant and any unrecognized mode falls back to requiring both
// fields (strict), so an alternate mode name can only tighten the check.
func dectestAdapterFlagCheckModes(t *testing.T, tb *typedBuild, adapterFunc string) map[string]bool {
	t.Helper()
	entry := tb.rootFuncsByName[dectestEntrypoint]
	if entry == nil {
		t.Fatalf("decTest dispatch entrypoint %q not found in the public build", dectestEntrypoint)
	}
	target := tb.rootFuncObj(adapterFunc)
	if target == nil {
		t.Fatalf("adapter %q is not a module-root package function", adapterFunc)
	}
	modes := map[string]bool{}
	ast.Inspect(entry.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok || len(call.Args) == 0 {
			return true
		}
		usesAdapter := false
		for _, arg := range call.Args {
			if id, ok := arg.(*ast.Ident); ok && tb.rootInfo.Uses[id] == target {
				usesAdapter = true
			}
		}
		if !usesAdapter {
			return true
		}
		if last, ok := call.Args[len(call.Args)-1].(*ast.Ident); ok {
			modes[last.Name] = true
		} else {
			modes["<non-identifier flag-check argument>"] = true
		}
		return true
	})
	if len(modes) == 0 {
		t.Errorf("decTest dispatch entrypoint passes adapter %q to no call; cannot derive its compared fields", adapterFunc)
	}
	return modes
}

// ---------------------------------------------------------------------------
// Phase 0 sanity: the typed load succeeds and populates the maps a known call
// depends on. Kept permanently as a strict corroborator of the loader.
// ---------------------------------------------------------------------------

func TestTypedLoadSanity(t *testing.T) {
	tb := loadTypedPublicBuild(t)

	if tb.rootPkg.Path() != rootImportPath {
		t.Fatalf("root package path = %q, want %q", tb.rootPkg.Path(), rootImportPath)
	}
	if tb.bidgoPkg.Path() != bidgoImportPath {
		t.Fatalf("bidgo package path = %q, want %q", tb.bidgoPkg.Path(), bidgoImportPath)
	}

	// A known plumbing call must resolve by object identity: the package-level
	// helper decimal64BIDDivPort calls bidgo.Bid64Div.
	helper := tb.rootFuncsByName["decimal64BIDDivPort"]
	if helper == nil {
		t.Fatal("decimal64BIDDivPort not found among Config A package functions")
	}
	reach := tb.rootResolver().reachablePortFuncs(helper)
	if !reach["Bid64Div"] {
		t.Fatalf("decimal64BIDDivPort reachable port set = %v, want it to contain Bid64Div", sortedSet(reach))
	}

	// The public method Div must reach the same port through the helper.
	div := tb.rootMethodsByName["Decimal64BID.Div"]
	if div == nil {
		t.Fatal("Decimal64BID.Div method not found")
	}
	if reachDiv := tb.rootResolver().reachablePortFuncs(div); !reachDiv["Bid64Div"] {
		t.Fatalf("Decimal64BID.Div reachable port set = %v, want it to contain Bid64Div", sortedSet(reachDiv))
	}

	// Selections must be populated: the reduce adapter's a.Reduce() call.
	adapter := tb.rootFuncsByName["executeDecTestReduceOperation"]
	if adapter == nil {
		t.Fatal("executeDecTestReduceOperation adapter not found in Config A")
	}
	foundReduceSelection := false
	ast.Inspect(adapter.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if tb.rootResolver().isValueTypeMethodCall(call, "Decimal64BID", "Reduce") {
			foundReduceSelection = true
		}
		return true
	})
	if !foundReduceSelection {
		t.Fatal("reduce adapter's Decimal64BID.Reduce() call did not resolve as a value-type method selection")
	}
}
