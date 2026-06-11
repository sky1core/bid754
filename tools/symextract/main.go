// tools/symextract/main.go - Extract symbol registry from the bid-go Go implementation path
//
// Uses go/types for full semantic analysis to extract all package-level symbols:
// types, constants, variables (tables), and function signatures.
//
// Output: JSON symbol registry to stdout
//
// Usage: go run tools/symextract/main.go > tools/registry/symbols.json

package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/constant"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Registry is the top-level symbol registry
type Registry struct {
	Version    string              `json:"version"`
	Source     string              `json:"source"`
	Types      map[string]TypeDef  `json:"types"`
	Constants  map[string]ConstDef `json:"constants"`
	Tables     map[string]TableDef `json:"tables"`
	Functions  map[string]FuncDef  `json:"functions"`
	ConstGroup map[string][]string `json:"const_groups"` // group name -> const names
}

type TypeDef struct {
	Fields   []FieldDef `json:"fields"`
	AliasOf  string     `json:"alias_of,omitempty"`
	GoFile   string     `json:"go_file"`
	Exported bool       `json:"exported"`
}

type FieldDef struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type ConstDef struct {
	Value    string `json:"value"`
	Type     string `json:"type"`
	GoFile   string `json:"go_file"`
	Group    string `json:"group,omitempty"`
	AliasOf  string `json:"alias_of,omitempty"`
	Exported bool   `json:"exported"`
}

type TableDef struct {
	ElementType string `json:"element_type"`
	Length      int    `json:"length"`
	GoFile      string `json:"go_file"`
	Exported    bool   `json:"exported"`
	IsSlice     bool   `json:"is_slice,omitempty"`
}

type FuncDef struct {
	Params   []ParamDef `json:"params"`
	Returns  []string   `json:"returns"`
	GoFile   string     `json:"go_file"`
	GoName   string     `json:"go_name"`
	Exported bool       `json:"exported"`
	Group    string     `json:"group,omitempty"`
}

type ParamDef struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Pointer bool   `json:"pointer,omitempty"`
}

func main() {
	projectRoot := findProjectRoot()
	bidGoDir := filepath.Join(projectRoot, "bid-go")

	reg := extractSymbols(bidGoDir)

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(reg); err != nil {
		fmt.Fprintf(os.Stderr, "json encode: %v\n", err)
		os.Exit(1)
	}
}

func findProjectRoot() string {
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			fmt.Fprintln(os.Stderr, "cannot find project root (go.mod)")
			os.Exit(1)
		}
		dir = parent
	}
}

func extractSymbols(srcDir string) *Registry {
	fset := token.NewFileSet()

	filter := func(fi os.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go") &&
			!strings.HasSuffix(fi.Name(), "_cgo.go")
	}

	pkgs, err := parser.ParseDir(fset, srcDir, filter, parser.ParseComments)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse error: %v\n", err)
		os.Exit(1)
	}

	var pkg *ast.Package
	for _, p := range pkgs {
		pkg = p
		break
	}
	if pkg == nil {
		fmt.Fprintln(os.Stderr, "no package found")
		os.Exit(1)
	}

	// Collect files
	var files []*ast.File
	var fileNames []string
	for name, f := range pkg.Files {
		files = append(files, f)
		fileNames = append(fileNames, name)
	}

	// Type-check
	conf := types.Config{
		Importer: importer.Default(),
		Error: func(err error) {
			// Ignore type errors - we still get useful info
		},
	}
	info := &types.Info{
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
		Types: make(map[ast.Expr]types.TypeAndValue),
	}

	tpkg, _ := conf.Check("bid754", fset, files, info)

	reg := &Registry{
		Version:    "1.0",
		Source:     "bid754 Go implementation (Intel BID mechanical port)",
		Types:      make(map[string]TypeDef),
		Constants:  make(map[string]ConstDef),
		Tables:     make(map[string]TableDef),
		Functions:  make(map[string]FuncDef),
		ConstGroup: make(map[string][]string),
	}

	if tpkg == nil {
		fmt.Fprintln(os.Stderr, "warning: type checking failed, falling back to AST-only analysis")
		extractFromAST(reg, pkg, fset, srcDir)
		return reg
	}

	scope := tpkg.Scope()
	for _, name := range scope.Names() {
		obj := scope.Lookup(name)
		goFile := filepath.Base(fset.Position(obj.Pos()).Filename)
		exported := obj.Exported()

		switch o := obj.(type) {
		case *types.TypeName:
			extractType(reg, o, goFile, exported)

		case *types.Const:
			extractConst(reg, o, goFile, exported, info)

		case *types.Var:
			extractVar(reg, o, goFile, exported)

		case *types.Func:
			extractFunc(reg, o, goFile, exported)
		}
	}

	// Detect constant aliases
	detectAliases(reg, pkg)

	// Assign groups
	assignGroups(reg)

	return reg
}

func extractType(reg *Registry, obj *types.TypeName, goFile string, exported bool) {
	underlying := obj.Type().Underlying()
	td := TypeDef{
		GoFile:   goFile,
		Exported: exported,
	}
	if st, ok := underlying.(*types.Struct); ok {
		for i := 0; i < st.NumFields(); i++ {
			f := st.Field(i)
			td.Fields = append(td.Fields, FieldDef{
				Name: f.Name(),
				Type: formatType(f.Type()),
			})
		}
	} else {
		td.AliasOf = formatType(underlying)
	}
	reg.Types[obj.Name()] = td
}

func extractConst(reg *Registry, obj *types.Const, goFile string, exported bool, info *types.Info) {
	val := obj.Val()
	valStr := ""

	switch val.Kind() {
	case constant.Int:
		// Try to get hex representation for large values
		if v, ok := constant.Int64Val(val); ok {
			if v > 255 || v < -255 {
				valStr = fmt.Sprintf("0x%x", uint64(v))
			} else {
				valStr = fmt.Sprintf("%d", v)
			}
		} else if v, ok := constant.Uint64Val(val); ok {
			valStr = fmt.Sprintf("0x%x", v)
		} else {
			valStr = val.ExactString()
		}
	case constant.Float:
		valStr = val.ExactString()
	case constant.String:
		valStr = val.ExactString()
	default:
		valStr = val.ExactString()
	}

	goType := constType(obj)

	reg.Constants[obj.Name()] = ConstDef{
		Value:    valStr,
		Type:     goType,
		GoFile:   goFile,
		Exported: exported,
	}
}

func constType(obj *types.Const) string {
	if basic, ok := obj.Type().(*types.Basic); ok && basic.Kind() == types.UntypedInt {
		if v, ok := constant.Int64Val(obj.Val()); ok {
			if v < 0 {
				if v >= -2147483648 && v <= 2147483647 {
					return "i32"
				}
				return "i64"
			}
			return "u64"
		}
		if _, ok := constant.Uint64Val(obj.Val()); ok {
			return "u64"
		}
	}
	return formatType(obj.Type())
}

func extractVar(reg *Registry, obj *types.Var, goFile string, exported bool) {
	typ := obj.Type()

	// Check if it's an array or slice (table)
	switch t := typ.(type) {
	case *types.Array:
		reg.Tables[obj.Name()] = TableDef{
			ElementType: formatType(t.Elem()),
			Length:      int(t.Len()),
			GoFile:      goFile,
			Exported:    exported,
		}
	case *types.Slice:
		reg.Tables[obj.Name()] = TableDef{
			ElementType: formatType(t.Elem()),
			Length:      -1, // unknown for slices
			GoFile:      goFile,
			Exported:    exported,
			IsSlice:     true,
		}
	default:
		// Non-table variable - still record it
		reg.Tables[obj.Name()] = TableDef{
			ElementType: formatType(typ),
			Length:      0,
			GoFile:      goFile,
			Exported:    exported,
		}
	}
}

func extractFunc(reg *Registry, obj *types.Func, goFile string, exported bool) {
	sig := obj.Type().(*types.Signature)

	// Skip methods (have receiver)
	if sig.Recv() != nil {
		return
	}

	fd := FuncDef{
		GoFile:   goFile,
		GoName:   obj.Name(),
		Exported: exported,
	}

	// Params
	params := sig.Params()
	for i := 0; i < params.Len(); i++ {
		p := params.At(i)
		pType := p.Type()
		isPtr := false
		if ptr, ok := pType.(*types.Pointer); ok {
			pType = ptr.Elem()
			isPtr = true
		}
		fd.Params = append(fd.Params, ParamDef{
			Name:    p.Name(),
			Type:    formatType(pType),
			Pointer: isPtr,
		})
	}

	// Returns
	results := sig.Results()
	for i := 0; i < results.Len(); i++ {
		fd.Returns = append(fd.Returns, formatType(results.At(i).Type()))
	}

	reg.Functions[obj.Name()] = fd
}

func formatType(t types.Type) string {
	switch tt := t.(type) {
	case *types.Basic:
		switch tt.Kind() {
		case types.Uint64:
			return "u64"
		case types.Uint32:
			return "u32"
		case types.Uint16:
			return "u16"
		case types.Uint8:
			return "u8"
		case types.Int64:
			return "i64"
		case types.Int32:
			return "i32"
		case types.Int16:
			return "i16"
		case types.Int8:
			return "i8"
		case types.Int:
			return "i32"
		case types.Uint:
			return "u32"
		case types.Float64:
			return "f64"
		case types.Float32:
			return "f32"
		case types.Bool:
			return "bool"
		case types.String:
			return "String"
		case types.UntypedInt:
			return "u64" // default for untyped int constants
		case types.UntypedFloat:
			return "f64"
		default:
			return tt.Name()
		}
	case *types.Named:
		return tt.Obj().Name()
	case *types.Array:
		return fmt.Sprintf("[%s; %d]", formatType(tt.Elem()), tt.Len())
	case *types.Slice:
		return fmt.Sprintf("Vec<%s>", formatType(tt.Elem()))
	case *types.Struct:
		return "struct"
	case *types.Pointer:
		return fmt.Sprintf("&mut %s", formatType(tt.Elem()))
	default:
		return t.String()
	}
}

func detectAliases(reg *Registry, pkg *ast.Package) {
	// Detect aliases by checking AST: const X = Y where Y is another constant identifier.
	// This is more accurate than value-based detection.
	for _, file := range pkg.Files {
		for _, decl := range file.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.CONST {
				continue
			}
			for _, spec := range gd.Specs {
				vs, ok := spec.(*ast.ValueSpec)
				if !ok || len(vs.Names) == 0 || len(vs.Values) == 0 {
					continue
				}
				for i, nameIdent := range vs.Names {
					if i >= len(vs.Values) {
						break
					}
					// Check if value is a single identifier (reference to another const)
					if ident, ok := vs.Values[i].(*ast.Ident); ok {
						name := nameIdent.Name
						target := ident.Name
						if _, isConst := reg.Constants[target]; isConst {
							if c, exists := reg.Constants[name]; exists {
								c.AliasOf = target
								reg.Constants[name] = c
							}
						}
					}
				}
			}
		}
	}
}

func assignGroups(reg *Registry) {
	for name, c := range reg.Constants {
		group := ""
		switch {
		case strings.Contains(c.GoFile, "bid128"):
			group = "bid128"
		case strings.Contains(c.GoFile, "bid32"):
			group = "bid32"
		case strings.HasSuffix(name, "64") || strings.Contains(c.GoFile, "noncomp64"):
			group = "bid64"
		case strings.HasSuffix(name, "128"):
			group = "bid128"
		case strings.HasSuffix(name, "32"):
			group = "bid32"
		case strings.HasPrefix(name, "BID_ROUNDING"):
			group = "rounding"
		case strings.Contains(name, "EXCEPTION") || strings.Contains(name, "STATUS"):
			group = "flags"
		case strings.Contains(c.GoFile, "internal"):
			group = "internal"
		default:
			group = "misc"
		}
		c.Group = group
		reg.Constants[name] = c

		reg.ConstGroup[group] = append(reg.ConstGroup[group], name)
	}

	// Sort within groups
	for g := range reg.ConstGroup {
		sort.Strings(reg.ConstGroup[g])
	}

	// Assign function groups
	for name, f := range reg.Functions {
		group := ""
		switch {
		case strings.HasPrefix(name, "Bid128") || strings.HasPrefix(name, "bid128"):
			group = "bid128"
		case strings.HasPrefix(name, "Bid64") || strings.HasPrefix(name, "bid64"):
			group = "bid64"
		case strings.HasPrefix(name, "Bid32") || strings.HasPrefix(name, "bid32"):
			group = "bid32"
		case strings.HasPrefix(name, "__"):
			group = "internal_helpers"
		default:
			group = "misc"
		}
		f.Group = group
		reg.Functions[name] = f
	}
}

// Fallback: AST-only extraction when type checking fails
func extractFromAST(reg *Registry, pkg *ast.Package, fset *token.FileSet, bidGoDir string) {
	for fileName, file := range pkg.Files {
		goFile := filepath.Base(fileName)

		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.GenDecl:
				for _, spec := range d.Specs {
					switch s := spec.(type) {
					case *ast.TypeSpec:
						if st, ok := s.Type.(*ast.StructType); ok {
							td := TypeDef{GoFile: goFile, Exported: s.Name.IsExported()}
							for _, field := range st.Fields.List {
								for _, name := range field.Names {
									td.Fields = append(td.Fields, FieldDef{
										Name: name.Name,
										Type: exprToString(field.Type),
									})
								}
							}
							reg.Types[s.Name.Name] = td
						}

					case *ast.ValueSpec:
						if d.Tok == token.CONST {
							for _, name := range s.Names {
								reg.Constants[name.Name] = ConstDef{
									GoFile:   goFile,
									Exported: name.IsExported(),
									Type:     exprToString(s.Type),
								}
							}
						} else if d.Tok == token.VAR {
							for _, name := range s.Names {
								reg.Tables[name.Name] = TableDef{
									GoFile:   goFile,
									Exported: name.IsExported(),
								}
							}
						}
					}
				}

			case *ast.FuncDecl:
				if d.Recv != nil {
					continue // skip methods
				}
				reg.Functions[d.Name.Name] = FuncDef{
					GoFile:   goFile,
					GoName:   d.Name.Name,
					Exported: d.Name.IsExported(),
				}
			}
		}
	}
}

func exprToString(e ast.Expr) string {
	if e == nil {
		return ""
	}
	switch t := e.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.ArrayType:
		return fmt.Sprintf("[%s]%s", exprToString(t.Len), exprToString(t.Elt))
	case *ast.StarExpr:
		return "*" + exprToString(t.X)
	default:
		return fmt.Sprintf("%T", e)
	}
}
