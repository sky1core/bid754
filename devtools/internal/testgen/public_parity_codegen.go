package testgen

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	gotypes "go/types"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// The public-API routing inventory enumerates every exported function, method, and package
// variable on the module-root package `bid754` and records the Go mechanical
// port function (bid754-go/internal/bidgo) each symbol is mapped to, or the
// classified plumbing exclusion. "mapped" is a static routing-census fact,
// not an execution result. The parity runner exercises the mappings, and the
// verification anchors compare the resulting artifact counts.
//
// The census is an exhaustive set: a new public symbol that neither auto-maps by
// name+shape nor carries a manifest override/exclusion is a generation-time
// hard failure, which is how the gate forces every future public entrypoint
// into a routing decision. Exported package variables fail even if an
// exclusion names them: callers could replace process-wide numeric state and
// race concurrent readers. Generated files are part of the census population
// too, so an exported symbol cannot escape the gate by being routed through a
// generated artifact.
const (
	publicParityInventoryPath = "generated/testspec/public_api_routing_inventory.json"
	// publicAPISrcDir is the module-root package `bid754`; loadPublicAPISigs
	// reads every non-test file there. It shares the bidgo source directory
	// constant (readtestGoportBidgoSrcDir) for the port side.
	publicAPISrcDir = "../bid754-go"
)

// publicValueTypeWidth maps a value-type method receiver to its bidgo width
// prefix. Methods on these receivers auto-map by name and flag shape against
// the bidgo Bid<width><Op> surface; methods on any other receiver must be
// classified by an exact symbol override/exclusion.
var publicValueTypeWidth = map[string]string{
	"Decimal32BID":  "32",
	"Decimal64BID":  "64",
	"Decimal128BID": "128",
}

// The public integer-conversion methods dispatch to a per-rounding-mode bidgo
// function rather than a single entrypoint, so a mode_dispatch override records
// the round-to-nearest-even representative and validates that the full variant
// family exists for the width. The Exact variants signal inexact via the X
// prototypes.
var (
	publicConvertRoundingVariants      = []string{"Rnint", "Rninta", "Int", "Ceil", "Floor"}
	publicConvertExactRoundingVariants = []string{"Xrnint", "Xrninta", "Xint", "Xceil", "Xfloor"}
)

// PublicAPISpec is the manifest `public_api` block: the exhaustive set of
// name-family overrides and classified exclusions the census consumes. Every
// entry must match at least one exported symbol (unused entries are a
// generation-time hard failure) and carry a concrete reason.
type PublicAPISpec struct {
	ValueMethodOverrides  []PublicValueMethodOverride  `json:"value_method_overrides"`
	ValueMethodExclusions []PublicValueMethodExclusion `json:"value_method_exclusions"`
	SymbolOverrides       []PublicSymbolOverride       `json:"symbol_overrides"`
	SymbolExclusions      []PublicSymbolExclusion      `json:"symbol_exclusions"`
}

// PublicValueMethodOverride maps a value-type method name (applied across all
// three widths where the method exists) to a bidgo operation whose name the
// automatic rule cannot derive (Remainder->Rem, String->ToString, and so on).
// The generator selects among the Bid<width><BidgoOp>[WithFlags|Raw] variants
// by the public method's flag shape, exactly like the automatic rule.
type PublicValueMethodOverride struct {
	Method string `json:"method"`
	// BidgoOp is the width-independent bidgo operation suffix. For a
	// mode_dispatch override it is the type stem (e.g. "ToInt32") and the
	// generator validates the rounding-variant family.
	BidgoOp      string `json:"bidgo_op"`
	ModeDispatch bool   `json:"mode_dispatch,omitempty"`
	// PreferWithFlags flips the flag-shape variant search to try the
	// Bid<width><Op>WithFlags widening wrapper before the base entrypoint.
	// It is for ops whose base entrypoint already returns a status word by
	// shape (so the default search would stop at it) but whose base pack
	// omits some Intel status flags at some widths, while a *WithFlags
	// widening wrapper carries the exhaustive set: the search then resolves to
	// the wrapper where it exists and falls back to the base where it does
	// not, so per-width flag completeness is expressed by one override.
	PreferWithFlags bool   `json:"prefer_with_flags,omitempty"`
	Reason          string `json:"reason"`
}

// PublicValueMethodExclusion classifies a value-type method that has no port
// counterpart (display formatting, byte-layout access, and similar) by name
// across widths.
type PublicValueMethodExclusion struct {
	Method string `json:"method"`
	Class  string `json:"class"`
	Reason string `json:"reason"`
}

// PublicSymbolOverride maps a single exported symbol (a package function or a
// method on a non-value receiver) to the exact bidgo function it routes
// through or is contractually paired with.
type PublicSymbolOverride struct {
	Symbol        string `json:"symbol"`
	BidgoFunction string `json:"bidgo_function"`
	Reason        string `json:"reason"`
}

// PublicSymbolExclusion classifies a single exported function or method on a
// non-value receiver as plumbing with no port counterpart. Exported package
// variables are rejected before exclusions are considered.
type PublicSymbolExclusion struct {
	Symbol string `json:"symbol"`
	Class  string `json:"class"`
	Reason string `json:"reason"`
}

// publicAPISymbol is one exported census entry, with its declared parameter
// and result types kept for shape matching against the bidgo port surface.
type publicAPISymbol struct {
	Symbol      string // unique id: "Name" for func/var, "Recv.Name" for method
	Kind        string // "func" | "method" | "var"
	Recv        string // method receiver base type, empty otherwise
	RecvPointer bool   // true when the declared receiver is *Recv
	Name        string
	TypeParams  []string // declared type-parameter name+constraint strings, in order
	Params      []string // declared parameter type strings (funcs/methods)
	Results     []string // declared result type strings (funcs/methods)
}

// wantsFlags reports whether the public symbol's declared results include the
// ExceptionFlags type, which pins the flag shape the mapped bidgo variant must
// expose.
func (s publicAPISymbol) wantsFlags() bool {
	for _, r := range s.Results {
		if r == "ExceptionFlags" {
			return true
		}
	}
	return false
}

type publicAPIInventoryRow struct {
	Symbol        string `json:"symbol"`
	Kind          string `json:"kind"`
	Status        string `json:"status"`
	BidgoFunction string `json:"bidgo_function,omitempty"`
	Reason        string `json:"reason,omitempty"`
}

type publicAPIInventory struct {
	Version         string                  `json:"version"`
	Source          string                  `json:"source"`
	Total           int                     `json:"total"`
	Mapped          int                     `json:"mapped"`
	Excluded        int                     `json:"excluded"`
	ExcludedByClass map[string]int          `json:"excluded_by_class"`
	Symbols         []publicAPIInventoryRow `json:"symbols"`
}

// WritePublicParityOutputs regenerates the public-API routing inventory from the
// pinned public source tree, the bidgo port surface, and the manifest
// public_api block.
func WritePublicParityOutputs(repoRoot string, manifest Manifest) error {
	files, err := GeneratePublicParityOutputs(repoRoot, manifest)
	if err != nil {
		return err
	}
	for path, data := range files {
		fullPath := filepath.Join(repoRoot, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			return fmt.Errorf("mkdir %q: %w", filepath.Dir(fullPath), err)
		}
		if err := os.WriteFile(fullPath, data, 0o644); err != nil {
			return fmt.Errorf("write generated public parity output %q: %w", fullPath, err)
		}
	}
	return nil
}

// GeneratePublicParityOutputs returns the generated public-API routing
// artifacts keyed by repo-relative path.
func GeneratePublicParityOutputs(repoRoot string, manifest Manifest) (map[string][]byte, error) {
	if manifest.PublicAPI == nil {
		return nil, fmt.Errorf("public parity: manifest public_api block is required")
	}
	symbols, err := loadPublicAPISigs(filepath.Join(repoRoot, publicAPISrcDir))
	if err != nil {
		return nil, err
	}
	sigs, err := loadBidgoFuncSigs(filepath.Join(repoRoot, readtestGoportBidgoSrcDir))
	if err != nil {
		return nil, err
	}
	inventory, err := resolvePublicAPI(symbols, sigs, manifest.PublicAPI)
	if err != nil {
		return nil, err
	}
	data, err := json.MarshalIndent(inventory, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal public API routing inventory: %w", err)
	}
	out := map[string][]byte{publicParityInventoryPath: append(data, '\n')}

	// The parity dispatch/runner exercises every mapped symbol.
	runner, err := GeneratePublicParityRunnerOutputs(repoRoot, symbols, sigs, inventory)
	if err != nil {
		return nil, err
	}
	for path, contents := range runner {
		out[path] = contents
	}
	return out, nil
}

// loadPublicAPISigs collects every exported function, method, and package
// variable declared in the non-test files of the module-root package
// `bid754`, including generated files: a generated artifact that exports a
// symbol widens the public surface just like a hand-written one, so it must
// be classified too. Build tags are ignored: files are parsed directly, so
// cgo/native-tagged variants contribute their declarations. Exact duplicate
// signatures across build-tag variants collapse to one census entry; a kind
// or signature mismatch fails so a mutable var cannot be hidden behind a
// same-named function from another build variant.
func loadPublicAPISigs(dir string) ([]publicAPISymbol, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read public API source dir %q: %w", dir, err)
	}
	seen := map[string]publicAPISymbol{}
	origins := map[string]string{}
	fset := token.NewFileSet()
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		path := filepath.Join(dir, name)
		file, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if err != nil {
			return nil, fmt.Errorf("parse public API source %q: %w", path, err)
		}
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				if !d.Name.IsExported() {
					continue
				}
				typeParams := publicTypeParamList(d.Type.TypeParams)
				params := publicFieldListTypes(d.Type.Params)
				results := publicFieldListTypes(d.Type.Results)
				if d.Recv == nil {
					sym := publicAPISymbol{Symbol: d.Name.Name, Kind: "func", Name: d.Name.Name, TypeParams: typeParams, Params: params, Results: results}
					if err := recordPublicAPISymbol(seen, origins, sym, path); err != nil {
						return nil, err
					}
					continue
				}
				recv, recvPointer := publicReceiverInfo(d.Recv)
				if recv == "" {
					return nil, fmt.Errorf("public API source %q: cannot resolve receiver type for method %q", path, d.Name.Name)
				}
				id := recv + "." + d.Name.Name
				sym := publicAPISymbol{Symbol: id, Kind: "method", Recv: recv, RecvPointer: recvPointer, Name: d.Name.Name, TypeParams: typeParams, Params: params, Results: results}
				if err := recordPublicAPISymbol(seen, origins, sym, path); err != nil {
					return nil, err
				}
			case *ast.GenDecl:
				if d.Tok != token.VAR {
					continue
				}
				for _, spec := range d.Specs {
					vs, ok := spec.(*ast.ValueSpec)
					if !ok {
						continue
					}
					for _, ident := range vs.Names {
						if ident.IsExported() {
							sym := publicAPISymbol{Symbol: ident.Name, Kind: "var", Name: ident.Name}
							if err := recordPublicAPISymbol(seen, origins, sym, path); err != nil {
								return nil, err
							}
						}
					}
				}
			}
		}
	}
	if len(seen) == 0 {
		return nil, fmt.Errorf("no exported public API symbols found under %q", dir)
	}
	symbols := make([]publicAPISymbol, 0, len(seen))
	for _, sym := range seen {
		symbols = append(symbols, sym)
	}
	sort.Slice(symbols, func(i, j int) bool { return symbols[i].Symbol < symbols[j].Symbol })
	return symbols, nil
}

func recordPublicAPISymbol(seen map[string]publicAPISymbol, origins map[string]string, sym publicAPISymbol, path string) error {
	if existing, ok := seen[sym.Symbol]; ok {
		if samePublicAPISymbol(existing, sym) {
			return nil
		}
		return fmt.Errorf("public API symbol %s has conflicting declarations in %q and %q: %+v vs %+v", sym.Symbol, origins[sym.Symbol], path, existing, sym)
	}
	seen[sym.Symbol] = sym
	origins[sym.Symbol] = path
	return nil
}

func samePublicAPISymbol(a, b publicAPISymbol) bool {
	return a.Symbol == b.Symbol && a.Kind == b.Kind && a.Recv == b.Recv && a.RecvPointer == b.RecvPointer && a.Name == b.Name &&
		samePublicAPITypeList(a.TypeParams, b.TypeParams) && samePublicAPITypeList(a.Params, b.Params) && samePublicAPITypeList(a.Results, b.Results)
}

func samePublicAPITypeList(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func publicReceiverInfo(recv *ast.FieldList) (string, bool) {
	if recv == nil || len(recv.List) == 0 {
		return "", false
	}
	expr := recv.List[0].Type
	pointer := false
	if star, ok := expr.(*ast.StarExpr); ok {
		expr = star.X
		pointer = true
	}
	if ident, ok := expr.(*ast.Ident); ok {
		return ident.Name, pointer
	}
	return "", pointer
}

// publicFieldListTypes flattens a parameter/result field list into one type
// string per declared name (or per anonymous field).
func publicFieldListTypes(fields *ast.FieldList) []string {
	if fields == nil {
		return nil
	}
	var types []string
	for _, field := range fields.List {
		typeName := publicExprString(field.Type)
		n := len(field.Names)
		if n == 0 {
			n = 1
		}
		for i := 0; i < n; i++ {
			types = append(types, typeName)
		}
	}
	return types
}

// publicTypeParamList preserves each type parameter's declared name, order,
// and constraint. Keeping only the constraints would collapse declarations
// such as [T, U any] and [U, T any] even though later parameter references can
// bind to different type-parameter positions.
func publicTypeParamList(fields *ast.FieldList) []string {
	if fields == nil {
		return nil
	}
	var params []string
	for _, field := range fields.List {
		constraint := publicExprString(field.Type)
		if len(field.Names) == 0 {
			params = append(params, constraint)
			continue
		}
		for _, name := range field.Names {
			params = append(params, name.Name+" "+constraint)
		}
	}
	return params
}

func publicExprString(expr ast.Expr) string {
	return gotypes.ExprString(expr)
}

// bidgoExposesFlags reports whether a bidgo port function surfaces the Intel
// status flags: either through a trailing *uint32 pfpsf parameter or a
// trailing extra uint32 result (the two flag shapes buildGoportCallPlan
// recognizes for the goport readtest dispatch).
func bidgoExposesFlags(sig bidgoFuncSig) bool {
	if n := len(sig.Params); n > 0 && sig.Params[n-1].Type == "*uint32" {
		return true
	}
	if n := len(sig.Results); n >= 2 && sig.Results[n-1].Type == "uint32" {
		return true
	}
	return false
}

// resolvePublicAPI applies the four-tier resolution (value-method exclusion,
// value-method override, automatic name+flag-shape rule for value-type
// methods, then exact symbol override/exclusion for everything else) and
// enforces the exhaustive set: unused manifest entries and unresolved symbols
// are hard failures.
func resolvePublicAPI(symbols []publicAPISymbol, bidgoSigs map[string]bidgoFuncSig, spec *PublicAPISpec) (publicAPIInventory, error) {
	var zero publicAPIInventory

	bidgoByNorm := map[string]string{}
	for name := range bidgoSigs {
		norm := NormalizeReadtestFuncName(name)
		if existing, ok := bidgoByNorm[norm]; !ok || name < existing {
			bidgoByNorm[norm] = name
		}
	}
	findBidgo := func(width, op string) (bidgoFuncSig, bool) {
		real, ok := bidgoByNorm[NormalizeReadtestFuncName("Bid"+width+op)]
		if !ok {
			return bidgoFuncSig{}, false
		}
		return bidgoSigs[real], true
	}

	vmoByMethod := map[string]int{}
	for i, o := range spec.ValueMethodOverrides {
		if o.Method == "" || o.BidgoOp == "" || o.Reason == "" {
			return zero, fmt.Errorf("public_api.value_method_overrides[%d] requires method, bidgo_op, and reason", i)
		}
		if _, dup := vmoByMethod[o.Method]; dup {
			return zero, fmt.Errorf("public_api.value_method_overrides duplicate method %q", o.Method)
		}
		vmoByMethod[o.Method] = i
	}
	vmeByMethod := map[string]int{}
	for i, e := range spec.ValueMethodExclusions {
		if e.Method == "" || e.Class == "" || e.Reason == "" {
			return zero, fmt.Errorf("public_api.value_method_exclusions[%d] requires method, class, and reason", i)
		}
		if _, dup := vmeByMethod[e.Method]; dup {
			return zero, fmt.Errorf("public_api.value_method_exclusions duplicate method %q", e.Method)
		}
		if _, dup := vmoByMethod[e.Method]; dup {
			return zero, fmt.Errorf("public_api method %q is both an override and an exclusion", e.Method)
		}
		vmeByMethod[e.Method] = i
	}
	soBySymbol := map[string]int{}
	for i, o := range spec.SymbolOverrides {
		if o.Symbol == "" || o.BidgoFunction == "" || o.Reason == "" {
			return zero, fmt.Errorf("public_api.symbol_overrides[%d] requires symbol, bidgo_function, and reason", i)
		}
		if _, dup := soBySymbol[o.Symbol]; dup {
			return zero, fmt.Errorf("public_api.symbol_overrides duplicate symbol %q", o.Symbol)
		}
		soBySymbol[o.Symbol] = i
	}
	seBySymbol := map[string]int{}
	for i, e := range spec.SymbolExclusions {
		if e.Symbol == "" || e.Class == "" || e.Reason == "" {
			return zero, fmt.Errorf("public_api.symbol_exclusions[%d] requires symbol, class, and reason", i)
		}
		if _, dup := seBySymbol[e.Symbol]; dup {
			return zero, fmt.Errorf("public_api.symbol_exclusions duplicate symbol %q", e.Symbol)
		}
		if _, dup := soBySymbol[e.Symbol]; dup {
			return zero, fmt.Errorf("public_api symbol %q is both an override and an exclusion", e.Symbol)
		}
		seBySymbol[e.Symbol] = i
	}

	vmoUsed := make([]bool, len(spec.ValueMethodOverrides))
	vmeUsed := make([]bool, len(spec.ValueMethodExclusions))
	soUsed := make([]bool, len(spec.SymbolOverrides))
	seUsed := make([]bool, len(spec.SymbolExclusions))

	var rows []publicAPIInventoryRow
	var unresolved []string
	excludedByClass := map[string]int{}
	mapped, excluded := 0, 0

	for _, sym := range symbols {
		if sym.Kind == "var" {
			return zero, fmt.Errorf("public parity census: exported mutable variable %s is rejected; expose a value-returning function instead", sym.Symbol)
		}
		width, isValueMethod := "", false
		if sym.Kind == "method" {
			if w, ok := publicValueTypeWidth[sym.Recv]; ok {
				width, isValueMethod = w, true
			}
		}

		if isValueMethod {
			if idx, ok := vmeByMethod[sym.Name]; ok {
				e := spec.ValueMethodExclusions[idx]
				vmeUsed[idx] = true
				rows = append(rows, publicAPIInventoryRow{Symbol: sym.Symbol, Kind: sym.Kind, Status: "excluded_" + e.Class, Reason: e.Reason})
				excluded++
				excludedByClass[e.Class]++
				continue
			}
			if idx, ok := vmoByMethod[sym.Name]; ok {
				o := spec.ValueMethodOverrides[idx]
				vmoUsed[idx] = true
				bidgoFn, reason, err := resolveValueMethodOverride(sym, width, o, findBidgo)
				if err != nil {
					return zero, err
				}
				rows = append(rows, publicAPIInventoryRow{Symbol: sym.Symbol, Kind: sym.Kind, Status: "mapped", BidgoFunction: bidgoFn, Reason: reason})
				mapped++
				continue
			}
			if bidgoFn, note, ok := autoResolveValueMethod(sym, width, findBidgo); ok {
				rows = append(rows, publicAPIInventoryRow{Symbol: sym.Symbol, Kind: sym.Kind, Status: "mapped", BidgoFunction: bidgoFn, Reason: note})
				mapped++
				continue
			}
			unresolved = append(unresolved, fmt.Sprintf("%s (value-type method: no bidgo name+flag-shape match, no manifest override/exclusion)", sym.Symbol))
			continue
		}

		if idx, ok := soBySymbol[sym.Symbol]; ok {
			o := spec.SymbolOverrides[idx]
			soUsed[idx] = true
			if _, exists := bidgoSigs[o.BidgoFunction]; !exists {
				return zero, fmt.Errorf("public_api.symbol_overrides[%s]: bidgo_function %q does not exist in the bidgo port", o.Symbol, o.BidgoFunction)
			}
			rows = append(rows, publicAPIInventoryRow{Symbol: sym.Symbol, Kind: sym.Kind, Status: "mapped", BidgoFunction: o.BidgoFunction, Reason: o.Reason})
			mapped++
			continue
		}
		if idx, ok := seBySymbol[sym.Symbol]; ok {
			e := spec.SymbolExclusions[idx]
			if err := validatePublicSymbolExclusion(sym, e); err != nil {
				return zero, err
			}
			seUsed[idx] = true
			rows = append(rows, publicAPIInventoryRow{Symbol: sym.Symbol, Kind: sym.Kind, Status: "excluded_" + e.Class, Reason: e.Reason})
			excluded++
			excludedByClass[e.Class]++
			continue
		}
		unresolved = append(unresolved, fmt.Sprintf("%s (%s: no manifest override/exclusion)", sym.Symbol, sym.Kind))
	}

	var unused []string
	for i, used := range vmoUsed {
		if !used {
			unused = append(unused, fmt.Sprintf("value_method_override method=%q", spec.ValueMethodOverrides[i].Method))
		}
	}
	for i, used := range vmeUsed {
		if !used {
			unused = append(unused, fmt.Sprintf("value_method_exclusion method=%q", spec.ValueMethodExclusions[i].Method))
		}
	}
	for i, used := range soUsed {
		if !used {
			unused = append(unused, fmt.Sprintf("symbol_override symbol=%q", spec.SymbolOverrides[i].Symbol))
		}
	}
	for i, used := range seUsed {
		if !used {
			unused = append(unused, fmt.Sprintf("symbol_exclusion symbol=%q", spec.SymbolExclusions[i].Symbol))
		}
	}
	if len(unused) > 0 {
		sort.Strings(unused)
		return zero, fmt.Errorf("public_api manifest has unused entries (remove them or fix the target):\n  %s", strings.Join(unused, "\n  "))
	}
	if len(unresolved) > 0 {
		sort.Strings(unresolved)
		return zero, fmt.Errorf("public parity census: %d exported symbol(s) are unresolved; make them follow the Bid<width><Op> naming rule or add a public_api override/exclusion with a concrete reason:\n  %s", len(unresolved), strings.Join(unresolved, "\n  "))
	}

	sort.Slice(rows, func(i, j int) bool { return rows[i].Symbol < rows[j].Symbol })
	return publicAPIInventory{
		Version:         "1.0",
		Source:          "bid754-go module-root package bid754 (all non-test files, generated included) + bid754-go/internal/bidgo + devtools/testgen_manifest.json public_api",
		Total:           len(symbols),
		Mapped:          mapped,
		Excluded:        excluded,
		ExcludedByClass: excludedByClass,
		Symbols:         rows,
	}, nil
}

func validatePublicSymbolExclusion(sym publicAPISymbol, exclusion PublicSymbolExclusion) error {
	if exclusion.Class != "constant_accessor" {
		return nil
	}
	validResult := len(sym.Results) == 1 && publicValueTypeWidth[sym.Results[0]] != ""
	if sym.Kind != "func" || len(sym.TypeParams) != 0 || len(sym.Params) != 0 || !validResult {
		return fmt.Errorf("public_api.symbol_exclusions[%s]: constant_accessor must be a non-generic package function with no parameters and exactly one Decimal32BID, Decimal64BID, or Decimal128BID result; got %+v", exclusion.Symbol, sym)
	}
	return nil
}

// selectBidgoVariantByFlagShape picks the bidgo variant whose flag exposure
// matches the public symbol's declared flag shape from an ordered candidate
// op list. A flags-returning public symbol requires a flags-exposing variant
// (pfpsf pointer or trailing uint32 result). A value-only public symbol
// prefers a flagless variant, and falls back to a flags-exposing one when the
// port has no flagless entrypoint (the wrapper drops the flags; the fallback
// is annotated in the inventory row).
func selectBidgoVariantByFlagShape(sym publicAPISymbol, width string, ops []string, findBidgo func(width, op string) (bidgoFuncSig, bool)) (string, string, bool) {
	type candidate struct {
		sig   bidgoFuncSig
		flags bool
	}
	var cands []candidate
	seenName := map[string]bool{}
	for _, op := range ops {
		sig, ok := findBidgo(width, op)
		if !ok || seenName[sig.Name] {
			continue
		}
		seenName[sig.Name] = true
		cands = append(cands, candidate{sig: sig, flags: bidgoExposesFlags(sig)})
	}
	if len(cands) == 0 {
		return "", "", false
	}
	if sym.wantsFlags() {
		for _, c := range cands {
			if c.flags {
				return c.sig.Name, "", true
			}
		}
		return "", "", false
	}
	for _, c := range cands {
		if !c.flags {
			return c.sig.Name, "", true
		}
	}
	return cands[0].sig.Name, "value-only wrapper over a flags-exposing port entrypoint (port flags dropped by the public signature)", true
}

// resolveValueMethodOverride turns a value-method override into a concrete
// bidgo function name for a given width, selecting the variant by the public
// flag shape and validating that the target (or the full rounding-variant
// family, for mode_dispatch) exists in the port.
func resolveValueMethodOverride(sym publicAPISymbol, width string, o PublicValueMethodOverride, findBidgo func(width, op string) (bidgoFuncSig, bool)) (string, string, error) {
	if o.ModeDispatch {
		variants := publicConvertRoundingVariants
		representative := "Rnint"
		if strings.HasSuffix(sym.Name, "Exact") {
			variants = publicConvertExactRoundingVariants
			representative = "Xrnint"
		}
		for _, v := range variants {
			sig, ok := findBidgo(width, o.BidgoOp+v)
			if !ok {
				return "", "", fmt.Errorf("public_api.value_method_overrides[%s]: mode-dispatch bidgo_op %q is missing variant Bid%s%s%s", sym.Symbol, o.BidgoOp, width, o.BidgoOp, v)
			}
			if sym.wantsFlags() && !bidgoExposesFlags(sig) {
				return "", "", fmt.Errorf("public_api.value_method_overrides[%s]: %s does not expose flags but the public method returns ExceptionFlags", sym.Symbol, sig.Name)
			}
		}
		rep, ok := findBidgo(width, o.BidgoOp+representative)
		if !ok {
			return "", "", fmt.Errorf("public_api.value_method_overrides[%s]: mode-dispatch bidgo_op %q has no representative Bid%s%s%s", sym.Symbol, o.BidgoOp, width, o.BidgoOp, representative)
		}
		family := make([]string, len(variants))
		for i, v := range variants {
			family[i] = "Bid" + width + o.BidgoOp + v
		}
		return rep.Name, o.Reason + " (mode-dispatched over " + strings.Join(family, ", ") + ")", nil
	}
	variantOps := []string{o.BidgoOp, o.BidgoOp + "WithFlags", o.BidgoOp + "Raw"}
	if o.PreferWithFlags {
		variantOps = []string{o.BidgoOp + "WithFlags", o.BidgoOp, o.BidgoOp + "Raw"}
	}
	name, note, ok := selectBidgoVariantByFlagShape(sym, width, variantOps, findBidgo)
	if !ok {
		return "", "", fmt.Errorf("public_api.value_method_overrides[%s]: bidgo_op %q has no Bid%s%s variant matching the public flag shape (wants flags: %v)", sym.Symbol, o.BidgoOp, width, o.BidgoOp, sym.wantsFlags())
	}
	reason := o.Reason
	if note != "" {
		reason += " (" + note + ")"
	}
	return name, reason, nil
}

// autoResolveValueMethod maps a value-type method whose name matches a bidgo
// Bid<width><Op> entrypoint family ({base, *WithFlags, *Raw}), selecting the
// variant whose flag exposure matches the public method's declared flag
// shape: a flags-returning method must map to a flags-exposing variant (this
// is what forces Decimal32BID.MinNum onto Bid32MinNumWithFlags instead of the
// flagless Bid32MinNum), and a value-only method prefers the flagless
// variant, falling back to a flags-exposing port when no flagless entrypoint
// exists.
func autoResolveValueMethod(sym publicAPISymbol, width string, findBidgo func(width, op string) (bidgoFuncSig, bool)) (string, string, bool) {
	baseOp := strings.TrimSuffix(sym.Name, "WithFlags")
	ops := []string{sym.Name}
	if baseOp != sym.Name {
		ops = append(ops, baseOp)
	}
	ops = append(ops, baseOp+"WithFlags", baseOp+"Raw")
	return selectBidgoVariantByFlagShape(sym, width, ops, findBidgo)
}
