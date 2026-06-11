// tools/go2rs/main.go - Go to Rust automatic converter for the bid-go implementation path
//
// Reads bid-go/*.go files (excluding tests and tables) and produces Rust source
// files in bid754-rs/src/generated/.
//
// Usage: go run tools/go2rs/main.go

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
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Project paths (relative to project root)
const (
	bidGoDir  = "bid-go"
	outputDir = "bid754-rs/src/generated"
)

type Registry struct {
	Types     map[string]TypeDef  `json:"types"`
	Constants map[string]ConstDef `json:"constants"`
	Tables    map[string]TableDef `json:"tables"`
	Functions map[string]FuncDef  `json:"functions"`
}

type TypeDef struct {
	Fields  []FieldDef `json:"fields"`
	AliasOf string     `json:"alias_of,omitempty"`
}
type FieldDef struct {
	Name string `json:"name"`
	Type string `json:"type"`
}
type ConstDef struct {
	Value   string `json:"value"`
	Type    string `json:"type"`
	AliasOf string `json:"alias_of,omitempty"`
}
type TableDef struct {
	ElementType string `json:"element_type"`
	Length      int    `json:"length"`
	IsSlice     bool   `json:"is_slice,omitempty"`
}
type FuncDef struct{}

var activeRegistry *Registry
var activeSourceFunctions map[string]bool
var activeStringVars map[string]bool
var activeReturnType string
var activeTypeInfo *types.Info

// Go → Rust type mapping
var typeMap = map[string]string{
	"uint64":       "u64",
	"uint32":       "u32",
	"uint16":       "u16",
	"uint8":        "u8",
	"int64":        "i64",
	"int32":        "i32",
	"int16":        "i16",
	"int8":         "i8",
	"int":          "i64",
	"uint":         "u64",
	"bool":         "bool",
	"float64":      "f64",
	"float32":      "f32",
	"byte":         "u8",
	"string":       "String",
	"error":        "&'static str",
	"BID_UINT128":  "BID_UINT128",
	"BID_UINT192":  "BID_UINT192",
	"BID_UINT256":  "BID_UINT256",
	"BID_UINT320":  "BID_UINT320",
	"BID_UINT384":  "BID_UINT384",
	"BID_UINT512":  "BID_UINT512",
	"RoundingMode": "i32",
}

var knownGoPackages = map[string]bool{
	"big":     true,
	"bits":    true,
	"errors":  true,
	"fmt":     true,
	"math":    true,
	"runtime": true,
	"strconv": true,
	"strings": true,
}

var bigIntBorrowArgs = map[string]map[int]bool{
	"bid128_finite_to_binary128_bits":  {2: true},
	"bid_finite_big_to_binary128_bits": {2: true},
	"floor_log2_rat":                   {0: true, 1: true},
	"pack":                             {2: true},
	"round_rat_to_int":                 {0: true, 1: true},
}

var rustKeywords = map[string]bool{
	"as": true, "break": true, "const": true, "continue": true, "crate": true,
	"else": true, "enum": true, "extern": true, "false": true, "fn": true,
	"for": true, "if": true, "impl": true, "in": true, "let": true,
	"loop": true, "match": true, "mod": true, "move": true, "mut": true,
	"pub": true, "ref": true, "return": true, "self": true, "Self": true,
	"static": true, "struct": true, "super": true, "trait": true, "true": true,
	"type": true, "unsafe": true, "use": true, "where": true, "while": true,
	"async": true, "await": true, "dyn": true,
}

// Go → Rust zero values
var zeroValues = map[string]string{
	"u64":          "0",
	"u32":          "0",
	"i64":          "0",
	"i32":          "0",
	"bool":         "false",
	"f64":          "0.0",
	"f32":          "0.0",
	"u8":           "0",
	"String":       "String::new()",
	"&'static str": "\"\"",
	"BigUint":      "BigUint::zero()",
	"BID_UINT128":  "BID_UINT128 { w: [0, 0] }",
	"BID_UINT192":  "BID_UINT192 { w: [0, 0, 0] }",
	"BID_UINT256":  "BID_UINT256 { w: [0, 0, 0, 0] }",
	"BID_UINT320":  "BID_UINT320 { w: [0, 0, 0, 0, 0] }",
	"BID_UINT384":  "BID_UINT384 { w: [0, 0, 0, 0, 0, 0] }",
	"BID_UINT512":  "BID_UINT512 { w: [0, 0, 0, 0, 0, 0, 0, 0] }",
}

var registryOwnedEmptySourceFiles = map[string]string{
	"types.go": "registry-generated gen_types/gen_constants own these declarations",
}

func main() {
	root := findProjectRoot()
	srcDir := filepath.Join(root, bidGoDir)
	outDir := filepath.Join(root, outputDir)
	reg := loadRegistry(filepath.Join(root, "tools", "registry", "symbols.json"))
	activeRegistry = reg
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		fatal("mkdir %s: %v", outDir, err)
	}

	files, err := filepath.Glob(filepath.Join(srcDir, "*.go"))
	if err != nil {
		fatal("glob: %v", err)
	}

	var targets []string
	for _, f := range files {
		base := filepath.Base(f)
		if !shouldConvertFile(base) {
			continue
		}
		targets = append(targets, f)
	}
	sort.Strings(targets)
	activeSourceFunctions = collectSourceFunctionNames(targets)
	packageFset, parsedTargets, typeInfo := parseTypeCheckedPackage(srcDir, targets)
	activeTypeInfo = typeInfo

	fmt.Printf("Converting %d files from %s → %s\n", len(targets), srcDir, outDir)
	type convertedFile struct {
		base   string
		rsName string
		code   string
	}
	var converted []convertedFile
	for _, f := range targets {
		base := filepath.Base(f)
		rsName := goFileToRustFile(base)

		rsCode, err := convertParsedFile(packageFset, parsedTargets[f], f, reg)
		if err != nil {
			fatal("convert %s: %v", base, err)
		}
		if rsCode == "" {
			reason, ok := registryOwnedEmptySourceFiles[base]
			if !ok {
				fatal("convert %s: no Rust output; add an explicit registry-owned entry if another generated artifact owns it", base)
			}
			fmt.Printf("  %s → <registry-owned> (%s)\n", base, reason)
			continue
		}

		rsCode = postProcess(rsCode)
		rsCode, err = optimizeRustStringHotpaths(base, rsCode)
		if err != nil {
			fatal("%v", err)
		}
		if err := rejectGeneratedFallbacks(base, rsCode); err != nil {
			fatal("%v", err)
		}
		converted = append(converted, convertedFile{base: base, rsName: rsName, code: rsCode})
	}

	cleanGeneratedDir(outDir, nil)
	var modFiles []string
	for _, file := range converted {
		outPath := filepath.Join(outDir, file.rsName)
		code := strings.TrimRight(file.code, "\n") + "\n"
		if err := os.WriteFile(outPath, []byte(code), 0644); err != nil {
			fatal("write %s: %v", outPath, err)
		}
		fmt.Printf("  %s → %s\n", file.base, file.rsName)
		modFiles = append(modFiles, file.rsName)
	}

	// Generate mod.rs
	sort.Strings(modFiles)
	generateModRs(outDir, modFiles)
	generatePreludeRs(outDir, modFiles)
	finalizeRustGenerated(root)
	if err := rejectGeneratedOwnershipLeaks(outDir); err != nil {
		fatal("%v", err)
	}
	if err := rejectFinalGeneratedFallbacks(outDir); err != nil {
		fatal("%v", err)
	}
	fmt.Printf("\nDone. %d files converted.\n", len(converted))
}

func collectSourceFunctionNames(files []string) map[string]bool {
	names := make(map[string]bool)
	for _, path := range files {
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			fatal("parse %s for function names: %v", path, err)
		}
		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if ok && fn.Recv == nil {
				names[fn.Name.Name] = true
			}
		}
	}
	return names
}

func parseTypeCheckedPackage(srcDir string, targets []string) (*token.FileSet, map[string]*ast.File, *types.Info) {
	fset := token.NewFileSet()
	files, err := filepath.Glob(filepath.Join(srcDir, "*.go"))
	if err != nil {
		fatal("glob typecheck files: %v", err)
	}
	targetSet := make(map[string]struct{}, len(targets))
	for _, path := range targets {
		targetSet[path] = struct{}{}
	}
	parsedTargets := make(map[string]*ast.File, len(targets))
	var parsedFiles []*ast.File
	for _, path := range files {
		base := filepath.Base(path)
		if strings.HasSuffix(base, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			fatal("parse %s for typecheck: %v", path, err)
		}
		parsedFiles = append(parsedFiles, file)
		if _, ok := targetSet[path]; ok {
			parsedTargets[path] = file
		}
	}
	for _, path := range targets {
		if parsedTargets[path] == nil {
			fatal("missing parsed target %s", path)
		}
	}
	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
	}
	conf := types.Config{Importer: importer.Default()}
	if _, err := conf.Check("github.com/sky1core/bid754/bid-go", fset, parsedFiles, info); err != nil {
		fatal("typecheck bid-go for go2rs: %v", err)
	}
	return fset, parsedTargets, info
}

func cleanGeneratedDir(outDir string, keep map[string]bool) {
	entries, err := os.ReadDir(outDir)
	if err != nil {
		fatal("read generated dir %s: %v", outDir, err)
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".rs") {
			continue
		}
		if e.Name() == "prelude.rs" {
			continue
		}
		if keep[e.Name()] {
			continue
		}
		path := filepath.Join(outDir, e.Name())
		if err := os.Remove(path); err != nil {
			fatal("remove stale generated artifact %s: %v", path, err)
		}
	}
}

func findProjectRoot() string {
	// Walk up from cwd to find go.mod
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			// fallback
			return "."
		}
		dir = parent
	}
}

func loadRegistry(path string) *Registry {
	data, err := os.ReadFile(path)
	if err != nil {
		fatal("read registry: %v", err)
	}
	var reg Registry
	if err := json.Unmarshal(data, &reg); err != nil {
		fatal("parse registry: %v", err)
	}
	if reg.Types == nil {
		reg.Types = make(map[string]TypeDef)
	}
	if reg.Constants == nil {
		reg.Constants = make(map[string]ConstDef)
	}
	if reg.Tables == nil {
		reg.Tables = make(map[string]TableDef)
	}
	if reg.Functions == nil {
		reg.Functions = make(map[string]FuncDef)
	}
	return &reg
}

func goFileToRustFile(goFile string) string {
	name := strings.TrimSuffix(goFile, ".go")
	return name + ".rs"
}

func shouldConvertFile(name string) bool {
	if strings.HasSuffix(name, "_test.go") {
		return false
	}
	if strings.HasPrefix(name, "tables") {
		return false
	}
	// decimal64.go is a pure-Go wrapper rather than the mechanical port surface
	// used as the Rust generation source.
	if name == "decimal64.go" {
		return false
	}
	return true
}

func generateModRs(outDir string, files []string) {
	var sb strings.Builder
	sb.WriteString("// Auto-generated by go2rs. Do not edit.\n\n")
	sb.WriteString("pub mod prelude;\n")
	if len(files) > 0 {
		sb.WriteString("\n")
	}
	for _, f := range files {
		mod := strings.TrimSuffix(f, ".rs")
		if mod == "prelude" {
			continue
		}
		sb.WriteString(fmt.Sprintf("pub mod %s;\n", mod))
	}
	path := filepath.Join(outDir, "mod.rs")
	if err := os.WriteFile(path, []byte(sb.String()), 0644); err != nil {
		fatal("write %s: %v", path, err)
	}
}

func generatePreludeRs(outDir string, files []string) {
	var sb strings.Builder
	sb.WriteString("// Auto-generated by go2rs. Do not edit.\n")
	sb.WriteString("// Prelude: re-exports all shared symbols for generated modules.\n")
	sb.WriteString("// Each generated file uses `use super::prelude::*;` to access these.\n\n")
	sb.WriteString("// Types and constants from registry\n")
	sb.WriteString("pub use crate::gen_types::*;\n")
	sb.WriteString("pub use crate::gen_constants::*;\n")
	sb.WriteString("pub use crate::tables::*;\n\n")
	sb.WriteString("// Shared external support used by generated modules.\n")
	sb.WriteString("pub use num_bigint::BigUint;\n")
	sb.WriteString("pub use num_traits::{One, Zero};\n\n")
	sb.WriteString(`pub type Int = ();

pub fn go_string_from_bytes<B: AsRef<[u8]>>(bytes: B) -> String {
    String::from_utf8_lossy(bytes.as_ref()).into_owned()
}

pub fn go_copy_str(dst: &mut [u8], src: &str) -> usize {
    let src = src.as_bytes();
    let n = dst.len().min(src.len());
    dst[..n].copy_from_slice(&src[..n]);
    n
}

pub fn go_append<T>(mut v: Vec<T>, item: T) -> Vec<T> {
    v.push(item);
    v
}

pub fn go_atoi(s: &str) -> (i64, Option<&'static str>) {
    match s.parse::<i64>() {
        Ok(v) => (v, None),
        Err(_) => (0, Some("atoi")),
    }
}

pub fn go_add64(x: u64, y: u64, carry: u64) -> (u64, u64) {
    let (s1, c1) = x.overflowing_add(y);
    let (s2, c2) = s1.overflowing_add(carry);
    let carry_out = if c1 || c2 { 1 } else { 0 };
    (s2, carry_out)
}

pub fn go_mul64(x: u64, y: u64) -> (u64, u64) {
    let p = (x as u128) * (y as u128);
    ((p >> 64) as u64, p as u64)
}

pub fn go_shift_count_u64(s: u64) -> Option<u32> {
    if s <= u32::MAX as u64 { Some(s as u32) } else { None }
}

pub fn go_shift_count_i64(s: i64) -> Option<u32> {
    if s < 0 {
        panic!("negative shift count")
    }
    go_shift_count_u64(s as u64)
}

pub fn go_checked_shl_u8(x: u8, s: Option<u32>) -> u8 {
    s.and_then(|n| x.checked_shl(n)).unwrap_or(0)
}

pub fn go_checked_shr_u8(x: u8, s: Option<u32>) -> u8 {
    s.and_then(|n| x.checked_shr(n)).unwrap_or(0)
}

pub fn go_checked_shl_u16(x: u16, s: Option<u32>) -> u16 {
    s.and_then(|n| x.checked_shl(n)).unwrap_or(0)
}

pub fn go_checked_shr_u16(x: u16, s: Option<u32>) -> u16 {
    s.and_then(|n| x.checked_shr(n)).unwrap_or(0)
}

pub fn go_checked_shl_u32(x: u32, s: Option<u32>) -> u32 {
    s.and_then(|n| x.checked_shl(n)).unwrap_or(0)
}

pub fn go_checked_shr_u32(x: u32, s: Option<u32>) -> u32 {
    s.and_then(|n| x.checked_shr(n)).unwrap_or(0)
}

pub fn go_checked_shl_u64(x: u64, s: Option<u32>) -> u64 {
    s.and_then(|n| x.checked_shl(n)).unwrap_or(0)
}

pub fn go_checked_shr_u64(x: u64, s: Option<u32>) -> u64 {
    s.and_then(|n| x.checked_shr(n)).unwrap_or(0)
}

pub fn go_checked_shl_usize(x: usize, s: Option<u32>) -> usize {
    s.and_then(|n| x.checked_shl(n)).unwrap_or(0)
}

pub fn go_checked_shr_usize(x: usize, s: Option<u32>) -> usize {
    s.and_then(|n| x.checked_shr(n)).unwrap_or(0)
}

pub fn go_checked_shl_i8(x: i8, s: Option<u32>) -> i8 {
    s.and_then(|n| x.checked_shl(n)).unwrap_or(0)
}

pub fn go_checked_shr_i8(x: i8, s: Option<u32>) -> i8 {
    s.and_then(|n| x.checked_shr(n)).unwrap_or(if x < 0 { -1 } else { 0 })
}

pub fn go_checked_shl_i16(x: i16, s: Option<u32>) -> i16 {
    s.and_then(|n| x.checked_shl(n)).unwrap_or(0)
}

pub fn go_checked_shr_i16(x: i16, s: Option<u32>) -> i16 {
    s.and_then(|n| x.checked_shr(n)).unwrap_or(if x < 0 { -1 } else { 0 })
}

pub fn go_checked_shl_i32(x: i32, s: Option<u32>) -> i32 {
    s.and_then(|n| x.checked_shl(n)).unwrap_or(0)
}

pub fn go_checked_shr_i32(x: i32, s: Option<u32>) -> i32 {
    s.and_then(|n| x.checked_shr(n)).unwrap_or(if x < 0 { -1 } else { 0 })
}

pub fn go_checked_shl_i64(x: i64, s: Option<u32>) -> i64 {
    s.and_then(|n| x.checked_shl(n)).unwrap_or(0)
}

pub fn go_checked_shr_i64(x: i64, s: Option<u32>) -> i64 {
    s.and_then(|n| x.checked_shr(n)).unwrap_or(if x < 0 { -1 } else { 0 })
}

pub fn go_big_bit_len(x: &BigUint) -> i64 {
    let digits = x.to_u64_digits();
    if digits.is_empty() {
        return 0;
    }
    let hi = *digits.last().unwrap();
    (((digits.len() - 1) * 64) + (64usize - (hi.leading_zeros() as usize))) as i64
}

pub fn go_big_to_u64(x: &BigUint) -> u64 {
    let digits = x.to_u64_digits();
    if digits.is_empty() {
        return 0;
    }
    digits[0]
}

pub fn go_big_sign(x: &BigUint) -> i64 {
    if x.is_zero() { 0 } else { 1 }
}

pub fn go_big_cmp(x: &BigUint, y: &BigUint) -> i64 {
    if x < y {
        -1
    } else if x > y {
        1
    } else {
        0
    }
}

pub fn go_big_bit(x: &BigUint, i: u64) -> u64 {
    let bit = (x >> (i as usize)) & BigUint::one();
    if bit.is_zero() { 0 } else { 1 }
}

`)
	sb.WriteString("// Cross-module function access (flat namespace like Go)\n")
	for _, f := range files {
		mod := strings.TrimSuffix(f, ".rs")
		if mod == "prelude" || mod == "mod" {
			continue
		}
		sb.WriteString(fmt.Sprintf("pub use super::%s::*;\n", mod))
	}
	path := filepath.Join(outDir, "prelude.rs")
	if err := os.WriteFile(path, []byte(sb.String()), 0644); err != nil {
		fatal("write %s: %v", path, err)
	}
}

func rejectGeneratedOwnershipLeaks(outDir string) error {
	entries, err := os.ReadDir(outDir)
	if err != nil {
		return fmt.Errorf("read generated dir %s: %w", outDir, err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".rs") {
			continue
		}
		path := filepath.Join(outDir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read generated artifact %s: %w", path, err)
		}
		src := string(data)
		for _, marker := range []string{
			"tools/codegen rust-optimize",
			"Code generated by tools/codegen",
			"rust-optimize",
		} {
			if strings.Contains(src, marker) {
				return fmt.Errorf("%s contains stale generated ownership marker %q", path, marker)
			}
		}
	}
	return nil
}

func rejectFinalGeneratedFallbacks(outDir string) error {
	entries, err := os.ReadDir(outDir)
	if err != nil {
		return fmt.Errorf("read generated dir %s: %w", outDir, err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".rs") {
			continue
		}
		path := filepath.Join(outDir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read generated artifact %s: %w", path, err)
		}
		src := string(data)
		if err := rejectGeneratedFallbacks(entry.Name(), src); err != nil {
			return err
		}
		if strings.Contains(src, "/* ") {
			return fmt.Errorf("go2rs generated unsupported expression fallback in %s", path)
		}
	}
	return nil
}

// convertFile parses a Go file and converts it to Rust source.
func convertFile(path string, reg *Registry) (string, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return "", err
	}
	return convertParsedFile(fset, f, path, reg)
}

func convertParsedFile(fset *token.FileSet, f *ast.File, path string, reg *Registry) (string, error) {
	if activeSourceFunctions == nil {
		activeSourceFunctions = make(map[string]bool)
	}
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if ok && fn.Recv == nil {
			activeSourceFunctions[fn.Name.Name] = true
		}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("// Auto-generated from %s by go2rs. Do not edit.\n\n", filepath.Base(path)))
	sb.WriteString("use super::prelude::*;\n\n")

	hasContent := false
	emitted := make(map[string]bool)

	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			code := convertGenDecl(fset, d, path, reg)
			if code != "" {
				sb.WriteString(code)
				sb.WriteString("\n")
				hasContent = true
			}
		case *ast.FuncDecl:
			rsName := rustIdent(goFuncNameToRust(d.Name.Name))
			if emitted[rsName] {
				continue
			}
			code := convertFuncDecl(fset, d, path)
			if code != "" {
				emitted[rsName] = true
				sb.WriteString(code)
				sb.WriteString("\n")
				hasContent = true
			}
		}
	}

	if !hasContent {
		return "", nil
	}
	return sb.String(), nil
}

// -------------------------------------------------------
// GenDecl conversion (const, var, type)
// -------------------------------------------------------

func convertGenDecl(fset *token.FileSet, d *ast.GenDecl, filePath string, reg *Registry) string {
	switch d.Tok {
	case token.CONST:
		return convertConstBlock(fset, d, filePath, reg)
	case token.VAR:
		return convertVarBlock(fset, d, filePath, reg)
	case token.TYPE:
		return convertTypeBlock(fset, d, filePath, reg)
	default:
		return ""
	}
}

func convertConstBlock(fset *token.FileSet, d *ast.GenDecl, filePath string, reg *Registry) string {
	var sb strings.Builder
	for _, spec := range d.Specs {
		vs, ok := spec.(*ast.ValueSpec)
		if !ok {
			continue
		}
		for i, name := range vs.Names {
			if _, skip := reg.Constants[name.Name]; skip {
				continue
			}
			if !ast.IsExported(name.Name) && !isInternalConst(name.Name) {
				continue
			}
			rsType := ""
			if vs.Type != nil {
				rsType = convertTypeExpr(vs.Type)
			}
			val := ""
			if i < len(vs.Values) {
				val = convertExpr(fset, vs.Values[i], filePath)
			}
			if val == "" {
				continue
			}
			if rsType != "" {
				sb.WriteString(fmt.Sprintf("pub const %s: %s = %s;\n", name.Name, rsType, val))
			} else {
				// Try to infer type from value
				inferredType := inferConstType(val)
				if inferredType != "" {
					sb.WriteString(fmt.Sprintf("pub const %s: %s = %s;\n", name.Name, inferredType, val))
				} else {
					sb.WriteString(fmt.Sprintf("pub const %s: u64 = %s;\n", name.Name, val))
				}
			}
		}
	}
	return sb.String()
}

func isInternalConst(name string) bool {
	// Include all-caps constants and BID_ prefixed
	if strings.HasPrefix(name, "BID_") || strings.HasPrefix(name, "MASK_") ||
		strings.HasPrefix(name, "NAN_") || strings.HasPrefix(name, "SNAN_") ||
		strings.HasPrefix(name, "INFINITY_") || strings.HasPrefix(name, "QUIET_") ||
		strings.HasPrefix(name, "LARGE_") || strings.HasPrefix(name, "SMALL_") ||
		strings.HasPrefix(name, "EXPONENT_") || strings.HasPrefix(name, "SPECIAL_") ||
		strings.HasPrefix(name, "LARGEST_") || strings.HasPrefix(name, "SMALLEST_") ||
		strings.HasPrefix(name, "BINARY_") || strings.HasPrefix(name, "DECIMAL_") ||
		strings.HasPrefix(name, "MAX_") || strings.HasPrefix(name, "MIN_") ||
		strings.HasPrefix(name, "MASK_") || strings.HasPrefix(name, "SINFINITY_") ||
		strings.HasPrefix(name, "SSNAN_") {
		return true
	}
	// All uppercase with underscore
	if isAllUpper(name) {
		return true
	}
	return false
}

func isAllUpper(s string) bool {
	for _, c := range s {
		if c >= 'a' && c <= 'z' {
			return false
		}
	}
	return len(s) > 1
}

func inferConstType(val string) string {
	if strings.HasPrefix(val, "0x") || strings.HasPrefix(val, "0X") {
		return "" // caller picks default
	}
	if strings.Contains(val, ".") {
		return "f64"
	}
	return ""
}

func convertVarBlock(fset *token.FileSet, d *ast.GenDecl, filePath string, reg *Registry) string {
	var sb strings.Builder
	for _, spec := range d.Specs {
		vs, ok := spec.(*ast.ValueSpec)
		if !ok {
			continue
		}
		for i, name := range vs.Names {
			if td, skip := reg.Tables[name.Name]; skip && isRegistryTable(td) {
				continue
			}
			rsType := ""
			if vs.Type != nil {
				rsType = convertTypeExprFull(vs.Type)
			}
			val := ""
			if i < len(vs.Values) {
				val = convertExpr(fset, vs.Values[i], filePath)
			}

			if rsType != "" && val != "" {
				sb.WriteString(fmt.Sprintf("pub static %s: %s = %s;\n", name.Name, rsType, val))
			} else if rsType != "" {
				sb.WriteString(fmt.Sprintf("pub static %s: %s; // TODO: init\n", name.Name, rsType))
			} else if val != "" {
				if inferredType := inferVarType(val); inferredType != "" {
					sb.WriteString(fmt.Sprintf("pub static %s: %s = %s;\n", name.Name, inferredType, val))
				}
			}
		}
	}
	return sb.String()
}

func inferVarType(val string) string {
	if strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"") {
		return "&'static str"
	}
	return ""
}

func convertTypeBlock(fset *token.FileSet, d *ast.GenDecl, filePath string, reg *Registry) string {
	var sb strings.Builder
	for _, spec := range d.Specs {
		ts, ok := spec.(*ast.TypeSpec)
		if !ok {
			continue
		}
		name := ts.Name.Name
		if td, skip := reg.Types[name]; skip && registryTypeOwnsRust(td) {
			continue
		}
		switch t := ts.Type.(type) {
		case *ast.StructType:
			derive := "#[derive(Clone, Copy, Default, Debug)]"
			if structContainsNonCopyField(t) {
				derive = "#[derive(Clone, Default, Debug)]"
			}
			sb.WriteString(fmt.Sprintf("%s\npub struct %s {\n", derive, name))
			if t.Fields != nil {
				for _, field := range t.Fields.List {
					fType := convertTypeExprFull(field.Type)
					for _, fname := range field.Names {
						sb.WriteString(fmt.Sprintf("    pub %s: %s,\n", fname.Name, fType))
					}
				}
			}
			sb.WriteString("}\n")
		case *ast.Ident:
			rsType := convertTypeExpr(t)
			sb.WriteString(fmt.Sprintf("pub type %s = %s;\n", name, rsType))
		case *ast.ArrayType:
			rsType := convertTypeExprFull(t)
			sb.WriteString(fmt.Sprintf("pub type %s = %s;\n", name, rsType))
		case *ast.BasicLit:
			sb.WriteString(fmt.Sprintf("// type %s = ...; // TODO\n", name))
		}
	}
	return sb.String()
}

func registryTypeOwnsRust(t TypeDef) bool {
	if t.AliasOf != "" {
		return true
	}
	for _, f := range t.Fields {
		if registryFieldHasUnsupportedRustType(f.Type) {
			return false
		}
	}
	return true
}

func registryFieldHasUnsupportedRustType(typ string) bool {
	return strings.Contains(typ, "Int") || strings.Contains(typ, "&mut")
}

func structContainsNonCopyField(t *ast.StructType) bool {
	if t.Fields == nil {
		return false
	}
	for _, field := range t.Fields.List {
		if convertTypeExprFull(field.Type) == "BigUint" {
			return true
		}
	}
	return false
}

// -------------------------------------------------------
// Function conversion
// -------------------------------------------------------

func convertFuncDecl(fset *token.FileSet, d *ast.FuncDecl, filePath string) string {
	if d.Recv != nil {
		// Method - skip for now (BID types don't have methods in bid-go)
		return ""
	}

	name := d.Name.Name
	rsName := rustIdent(goFuncNameToRust(name))
	visibility := "pub "
	if !ast.IsExported(name) {
		visibility = "pub(crate) "
	}

	// Parameters
	params, stringParams := convertFuncParams(d.Type.Params)

	// Return type
	retType := convertFuncResults(d.Type.Results)

	// Body
	prevStringVars := activeStringVars
	prevReturnType := activeReturnType
	activeStringVars = collectStringParamNames(d.Type.Params)
	activeReturnType = retType
	body := convertFuncBody(fset, d.Body, filePath, d.Type.Results)
	activeStringVars = prevStringVars
	activeReturnType = prevReturnType

	var sb strings.Builder
	if retType != "" {
		sb.WriteString(fmt.Sprintf("%sfn %s(%s) -> %s {\n", visibility, rsName, params, retType))
	} else {
		sb.WriteString(fmt.Sprintf("%sfn %s(%s) {\n", visibility, rsName, params))
	}
	for _, name := range stringParams {
		sb.WriteString(fmt.Sprintf("    let mut %s = %s.as_ref().to_string();\n", name, name))
	}
	sb.WriteString(body)
	sb.WriteString("}\n")

	return sb.String()
}

func collectStringParamNames(fields *ast.FieldList) map[string]bool {
	names := make(map[string]bool)
	if fields == nil {
		return names
	}
	for _, field := range fields.List {
		if !isGoStringType(field.Type) {
			continue
		}
		for _, name := range field.Names {
			names[rustIdent(name.Name)] = true
		}
	}
	return names
}

func goFuncNameToRust(name string) string {
	// Convert CamelCase to snake_case.
	return camelToSnake(name)
}

func camelToSnake(s string) string {
	// Special handling: preserve underscores at start
	prefix := ""
	for strings.HasPrefix(s, "_") {
		prefix += "_"
		s = s[1:]
	}
	if s == "" {
		return prefix
	}

	// Classify each rune
	type charClass int
	const (
		ccLower charClass = iota
		ccUpper
		ccDigit
		ccOther
	)
	classify := func(r rune) charClass {
		switch {
		case r >= 'a' && r <= 'z':
			return ccLower
		case r >= 'A' && r <= 'Z':
			return ccUpper
		case r >= '0' && r <= '9':
			return ccDigit
		default:
			return ccOther
		}
	}

	runes := []rune(s)
	var result strings.Builder
	for i, r := range runes {
		cc := classify(r)
		if i > 0 && cc == ccUpper {
			prev := classify(runes[i-1])
			if prev == ccLower || prev == ccDigit {
				// aB → a_b, 4B → 4_b
				result.WriteByte('_')
			} else if prev == ccUpper && i+1 < len(runes) && classify(runes[i+1]) == ccLower {
				// ABc → a_bc (end of acronym)
				result.WriteByte('_')
			}
		}
		if cc == ccUpper {
			result.WriteRune(r - 'A' + 'a')
		} else {
			result.WriteRune(r)
		}
	}

	return prefix + result.String()
}

func rustIdent(name string) string {
	if rustKeywords[name] {
		return "r#" + name
	}
	return name
}

func constLiteral(name string) string {
	if activeRegistry == nil {
		return ""
	}
	c, ok := activeRegistry.Constants[name]
	if !ok {
		return ""
	}
	if c.Value != "" {
		return rustConstLiteral(c.Value, c.Type)
	}
	if c.AliasOf != "" && c.AliasOf != name {
		return constLiteral(c.AliasOf)
	}
	return ""
}

func rustConstLiteral(value, rustType string) string {
	if lit, ok := signedHexConstLiteral(value, rustType); ok {
		return lit
	}
	return value
}

func signedHexConstLiteral(value, rustType string) (string, bool) {
	switch rustType {
	case "i8", "i16", "i32", "i64":
	default:
		return "", false
	}
	if !strings.HasPrefix(value, "0x") && !strings.HasPrefix(value, "0X") {
		return "", false
	}
	u, err := strconv.ParseUint(value[2:], 16, 64)
	if err != nil {
		return "", false
	}
	if (u & (uint64(1) << 63)) == 0 {
		return "", false
	}
	s := int64(u)
	if !fitsSignedConstType(s, rustType) {
		return "", false
	}
	return strconv.FormatInt(s, 10), true
}

func fitsSignedConstType(v int64, rustType string) bool {
	switch rustType {
	case "i8":
		return v >= -128 && v <= 127
	case "i16":
		return v >= -32768 && v <= 32767
	case "i32":
		return v >= -2147483648 && v <= 2147483647
	case "i64":
		return true
	default:
		return false
	}
}

func convertFuncParams(fields *ast.FieldList) (string, []string) {
	if fields == nil || len(fields.List) == 0 {
		return "", nil
	}

	var parts []string
	var stringParams []string
	for _, field := range fields.List {
		rsType := convertParamType(field.Type)
		// Check if param is a pointer/ref type
		_, isPtr := field.Type.(*ast.StarExpr)
		isStringParam := isGoStringType(field.Type)
		for _, name := range field.Names {
			rsName := rustIdent(name.Name)
			if isStringParam {
				parts = append(parts, fmt.Sprintf("%s: impl AsRef<str>", rsName))
				stringParams = append(stringParams, rsName)
			} else if isPtr {
				parts = append(parts, fmt.Sprintf("%s: %s", rsName, rsType))
			} else {
				// Make all value parameters mut (Go params are mutable by default)
				parts = append(parts, fmt.Sprintf("mut %s: %s", rsName, rsType))
			}
		}
		if len(field.Names) == 0 {
			parts = append(parts, fmt.Sprintf("_: %s", rsType))
		}
	}
	return strings.Join(parts, ", "), stringParams
}

func isGoStringType(expr ast.Expr) bool {
	ident, ok := expr.(*ast.Ident)
	return ok && ident.Name == "string"
}

func convertParamType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.StarExpr:
		if isBigIntType(t.X) {
			return "&BigUint"
		}
		inner := convertTypeExprFull(t.X)
		if inner == "Int" {
			return "Int"
		}
		return fmt.Sprintf("&mut %s", inner)
	case *ast.ArrayType:
		elt := convertTypeExprFull(t.Elt)
		if t.Len == nil {
			return fmt.Sprintf("&mut [%s]", elt)
		}
		if lit, ok := t.Len.(*ast.BasicLit); ok {
			return fmt.Sprintf("&mut [%s; %s]", elt, lit.Value)
		}
		return fmt.Sprintf("&mut [%s]", elt)
	default:
		return convertTypeExprFull(expr)
	}
}

func convertFuncResults(fields *ast.FieldList) string {
	if fields == nil || len(fields.List) == 0 {
		return ""
	}

	var types []string
	for _, field := range fields.List {
		rsType := convertTypeExpr(field.Type)
		count := len(field.Names)
		if count == 0 {
			count = 1
		}
		for i := 0; i < count; i++ {
			types = append(types, rsType)
		}
	}

	if len(types) == 1 {
		return types[0]
	}
	return "(" + strings.Join(types, ", ") + ")"
}

// -------------------------------------------------------
// Function body conversion (source-level text manipulation)
// -------------------------------------------------------

func convertFuncBody(fset *token.FileSet, body *ast.BlockStmt, filePath string, results *ast.FieldList) string {
	if body == nil {
		fatal("missing function body in %s", filePath)
	}

	// Read the source file
	src, err := os.ReadFile(filePath)
	if err != nil {
		fatal("read %s: %v", filePath, err)
	}

	// Collect named return variables
	namedReturns := collectNamedReturns(results)

	var sb strings.Builder

	// Emit local variable declarations for named returns
	for _, nr := range namedReturns {
		rsType := mapType(nr.typ)
		zero := zeroVal(rsType)
		sb.WriteString(fmt.Sprintf("    let mut %s: %s = %s;\n", nr.name, rsType, zero))
	}

	for _, stmt := range body.List {
		code := convertStmt(fset, stmt, src, 1, namedReturns)
		sb.WriteString(code)
	}

	return sb.String()
}

type namedReturn struct {
	name string
	typ  string
}

func collectNamedReturns(results *ast.FieldList) []namedReturn {
	if results == nil {
		return nil
	}
	var nrs []namedReturn
	for _, field := range results.List {
		if len(field.Names) > 0 {
			typ := convertTypeExprFull(field.Type)
			for _, name := range field.Names {
				nrs = append(nrs, namedReturn{name: rustIdent(name.Name), typ: typ})
			}
		}
	}
	return nrs
}

func typeExprToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + typeExprToString(t.X)
	case *ast.ArrayType:
		if t.Len != nil {
			return "[" + typeExprToString(t.Len) + "]" + typeExprToString(t.Elt)
		}
		return "[]" + typeExprToString(t.Elt)
	case *ast.SelectorExpr:
		return typeExprToString(t.X) + "." + t.Sel.Name
	default:
		return "unknown"
	}
}

func mapType(goType string) string {
	if rt, ok := typeMap[goType]; ok {
		return rt
	}
	return goType
}

func zeroVal(rsType string) string {
	if z, ok := zeroValues[rsType]; ok {
		return z
	}
	if strings.HasPrefix(rsType, "[") && strings.Contains(rsType, ";") && strings.HasSuffix(rsType, "]") {
		inner := rsType[1 : len(rsType)-1]
		parts := strings.SplitN(inner, ";", 2)
		if len(parts) == 2 {
			elemType := strings.TrimSpace(parts[0])
			size := strings.TrimSpace(parts[1])
			elemZero := zeroVal(elemType)
			if elemZero != "Default::default()" && !strings.Contains(elemZero, "String::new()") {
				return fmt.Sprintf("[%s; %s]", elemZero, size)
			}
		}
	}
	return "Default::default()"
}

func isRegistryTable(td TableDef) bool {
	return td.IsSlice || td.Length > 0 || strings.HasPrefix(td.ElementType, "[")
}

// -------------------------------------------------------
// Statement conversion
// -------------------------------------------------------

func convertStmt(fset *token.FileSet, stmt ast.Stmt, src []byte, indent int, namedReturns []namedReturn) string {
	ind := strings.Repeat("    ", indent)

	switch s := stmt.(type) {
	case *ast.DeclStmt:
		return convertDeclStmt(fset, s, src, indent)

	case *ast.AssignStmt:
		return convertAssignStmt(fset, s, src, indent)

	case *ast.ReturnStmt:
		return convertReturnStmt(fset, s, src, indent, namedReturns)

	case *ast.IfStmt:
		return convertIfStmt(fset, s, src, indent, namedReturns)

	case *ast.ForStmt:
		return convertForStmt(fset, s, src, indent, namedReturns)

	case *ast.RangeStmt:
		return convertRangeStmt(fset, s, src, indent, namedReturns)

	case *ast.SwitchStmt:
		return convertSwitchStmt(fset, s, src, indent, namedReturns)

	case *ast.ExprStmt:
		if call, ok := s.X.(*ast.CallExpr); ok {
			if code, ok := convertBigIntMutatingStmt(fset, call, src, indent); ok {
				return code
			}
		}
		expr := convertExprStr(fset, s.X, src)
		return fmt.Sprintf("%s%s;\n", ind, expr)

	case *ast.IncDecStmt:
		x := convertExprStr(fset, s.X, src)
		if s.Tok == token.INC && isIntegerExpr(s.X) {
			return fmt.Sprintf("%s%s = %s.wrapping_add(1);\n", ind, x, x)
		}
		if s.Tok == token.DEC && isIntegerExpr(s.X) {
			return fmt.Sprintf("%s%s = %s.wrapping_sub(1);\n", ind, x, x)
		}
		if s.Tok == token.INC {
			return fmt.Sprintf("%s%s += 1;\n", ind, x)
		}
		return fmt.Sprintf("%s%s -= 1;\n", ind, x)

	case *ast.BranchStmt:
		switch s.Tok {
		case token.BREAK:
			return fmt.Sprintf("%sbreak;\n", ind)
		case token.CONTINUE:
			return fmt.Sprintf("%scontinue;\n", ind)
		case token.GOTO:
			// Rust doesn't have goto natively; we'll use a comment-marker
			// In practice, goto in bid-go maps to labeled loops/blocks
			if s.Label != nil {
				return fmt.Sprintf("%s// goto %s; // TODO: convert goto to loop/break\n", ind, s.Label.Name)
			}
			return fmt.Sprintf("%s// goto; // TODO\n", ind)
		}
		return ""

	case *ast.LabeledStmt:
		label := s.Label.Name
		body := convertStmt(fset, s.Stmt, src, indent, namedReturns)
		return fmt.Sprintf("%s// label: %s\n%s", ind, label, body)

	case *ast.BlockStmt:
		return convertBlockStmt(fset, s, src, indent, namedReturns)

	case *ast.EmptyStmt:
		return ""

	default:
		// Fallback: extract source text
		text := extractSource(fset, stmt, src)
		lines := strings.Split(text, "\n")
		var sb strings.Builder
		for _, line := range lines {
			sb.WriteString(fmt.Sprintf("%s// TODO: %s\n", ind, line))
		}
		return sb.String()
	}
}

func convertBlockStmt(fset *token.FileSet, s *ast.BlockStmt, src []byte, indent int, namedReturns []namedReturn) string {
	var sb strings.Builder
	for _, stmt := range s.List {
		sb.WriteString(convertStmt(fset, stmt, src, indent, namedReturns))
	}
	return sb.String()
}

func sliceExprBaseName(s *ast.SliceExpr) string {
	ident, ok := s.X.(*ast.Ident)
	if !ok {
		return ""
	}
	return rustIdent(ident.Name)
}

func convertBigIntMutatingStmt(fset *token.FileSet, call *ast.CallExpr, src []byte, indent int) (string, bool) {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return "", false
	}
	recvIdent, ok := sel.X.(*ast.Ident)
	if !ok {
		return "", false
	}
	recv := rustIdent(recvIdent.Name)
	ind := strings.Repeat("    ", indent)
	arg := func(i int) string {
		if i >= len(call.Args) {
			return ""
		}
		return convertExprStr(fset, call.Args[i], src)
	}
	shift := func(i int) string {
		return fmt.Sprintf("(%s as usize)", arg(i))
	}

	switch sel.Sel.Name {
	case "Lsh":
		if len(call.Args) != 2 {
			return "", false
		}
		lhs := arg(0)
		if lhs == recv {
			return fmt.Sprintf("%s%s <<= %s;\n", ind, recv, shift(1)), true
		}
		return fmt.Sprintf("%s%s = %s.clone() << %s;\n", ind, recv, lhs, shift(1)), true
	case "Rsh":
		if len(call.Args) != 2 {
			return "", false
		}
		lhs := arg(0)
		if lhs == recv {
			return fmt.Sprintf("%s%s >>= %s;\n", ind, recv, shift(1)), true
		}
		return fmt.Sprintf("%s%s = %s.clone() >> %s;\n", ind, recv, lhs, shift(1)), true
	case "Or":
		if len(call.Args) != 2 {
			return "", false
		}
		lhs := arg(0)
		rhs := arg(1)
		if lhs != recv {
			return fmt.Sprintf("%s%s = %s.clone() | %s.clone();\n", ind, recv, lhs, rhs), true
		}
		return fmt.Sprintf("%s%s |= %s.clone();\n", ind, recv, rhs), true
	case "Add":
		if len(call.Args) != 2 {
			return "", false
		}
		lhs := arg(0)
		rhs := arg(1)
		if lhs != recv {
			return fmt.Sprintf("%s%s = %s.clone() + %s;\n", ind, recv, lhs, rhs), true
		}
		return fmt.Sprintf("%s%s += %s;\n", ind, recv, rhs), true
	case "Sub":
		if len(call.Args) != 2 {
			return "", false
		}
		lhs := arg(0)
		rhs := arg(1)
		if lhs != recv {
			return fmt.Sprintf("%s%s = %s.clone() - %s;\n", ind, recv, lhs, rhs), true
		}
		return fmt.Sprintf("%s%s -= %s;\n", ind, recv, rhs), true
	case "Mul":
		if len(call.Args) != 2 {
			return "", false
		}
		lhs := arg(0)
		rhs := arg(1)
		if lhs != recv {
			return fmt.Sprintf("%s%s = %s.clone() * %s;\n", ind, recv, lhs, rhs), true
		}
		return fmt.Sprintf("%s%s *= %s;\n", ind, recv, rhs), true
	case "Quo":
		if len(call.Args) != 2 {
			return "", false
		}
		lhs := arg(0)
		rhs := arg(1)
		if lhs != recv {
			return fmt.Sprintf("%s%s = %s.clone() / %s;\n", ind, recv, lhs, rhs), true
		}
		return fmt.Sprintf("%s%s /= %s;\n", ind, recv, rhs), true
	case "QuoRem":
		if len(call.Args) != 3 {
			return "", false
		}
		num := arg(0)
		den := arg(1)
		rem := arg(2)
		return fmt.Sprintf("%s%s = %s / %s;\n%s%s = %s %% %s;\n", ind, recv, num, den, ind, rem, num, den), true
	default:
		return "", false
	}
}

// -------------------------------------------------------
// DeclStmt: var declarations inside function body
// -------------------------------------------------------

func convertDeclStmt(fset *token.FileSet, s *ast.DeclStmt, src []byte, indent int) string {
	ind := strings.Repeat("    ", indent)
	gd, ok := s.Decl.(*ast.GenDecl)
	if !ok {
		return ""
	}

	var sb strings.Builder
	for _, spec := range gd.Specs {
		vs, ok := spec.(*ast.ValueSpec)
		if !ok {
			continue
		}
		rsType := ""
		if vs.Type != nil {
			rsType = convertTypeExprFull(vs.Type)
		}
		for i, name := range vs.Names {
			if name.Name == "_" {
				if i < len(vs.Values) {
					val := convertExprStr(fset, vs.Values[i], src)
					sb.WriteString(fmt.Sprintf("%slet _ = %s;\n", ind, val))
				}
				continue
			}
			rsName := rustIdent(name.Name)

			if i < len(vs.Values) {
				val := convertExprStr(fset, vs.Values[i], src)
				if rsType != "" {
					sb.WriteString(fmt.Sprintf("%slet mut %s: %s = %s;\n", ind, rsName, rsType, val))
				} else if inferred := integerRustType(name); inferred != "" && isIntegerConstExpr(vs.Values[i]) {
					sb.WriteString(fmt.Sprintf("%slet mut %s: %s = %s;\n", ind, rsName, inferred, val))
				} else if inferred := inferLocalLiteralType(vs.Values[i]); inferred != "" {
					sb.WriteString(fmt.Sprintf("%slet mut %s: %s = %s;\n", ind, rsName, inferred, val))
				} else {
					sb.WriteString(fmt.Sprintf("%slet mut %s = %s;\n", ind, rsName, val))
				}
			} else {
				if rsType != "" {
					zero := zeroVal(rsType)
					sb.WriteString(fmt.Sprintf("%slet mut %s: %s = %s;\n", ind, rsName, rsType, zero))
				} else {
					sb.WriteString(fmt.Sprintf("%slet mut %s = Default::default();\n", ind, rsName))
				}
			}
		}
	}
	return sb.String()
}

// -------------------------------------------------------
// AssignStmt
// -------------------------------------------------------

func convertAssignStmt(fset *token.FileSet, s *ast.AssignStmt, src []byte, indent int) string {
	ind := strings.Repeat("    ", indent)

	// Short variable declaration :=
	if s.Tok == token.DEFINE {
		if len(s.Lhs) == 1 && len(s.Rhs) == 1 {
			lhs := convertExprStr(fset, s.Lhs[0], src)
			rhs := convertExprStr(fset, s.Rhs[0], src)
			if lhs == "_" {
				return fmt.Sprintf("%slet _ = %s;\n", ind, rhs)
			}
			if inferred := integerRustType(s.Lhs[0]); inferred != "" && isIntegerConstExpr(s.Rhs[0]) {
				return fmt.Sprintf("%slet mut %s: %s = %s;\n", ind, lhs, inferred, rhs)
			}
			if inferred := inferLocalLiteralType(s.Rhs[0]); inferred != "" {
				return fmt.Sprintf("%slet mut %s: %s = %s;\n", ind, lhs, inferred, rhs)
			}
			if casted := castIntegerExprToContext(rhs, rustExprValueType(s.Rhs[0]), integerRustType(s.Lhs[0]), s.Rhs[0]); casted != rhs {
				return fmt.Sprintf("%slet mut %s = %s;\n", ind, lhs, casted)
			}
			return fmt.Sprintf("%slet mut %s = %s;\n", ind, lhs, rhs)
		}
		// Multiple return values
		if len(s.Rhs) == 1 {
			var lhsParts []string
			for _, l := range s.Lhs {
				lhsParts = append(lhsParts, convertExprStr(fset, l, src))
			}
			rhs := convertExprStr(fset, s.Rhs[0], src)
			// Make non-_ variables mutable
			var bindParts []string
			for _, p := range lhsParts {
				if p == "_" {
					bindParts = append(bindParts, "_")
				} else {
					bindParts = append(bindParts, "mut "+p)
				}
			}
			return fmt.Sprintf("%slet (%s) = %s;\n", ind, strings.Join(bindParts, ", "), rhs)
		}
	}

	// Regular assignment
	if len(s.Lhs) == 1 && len(s.Rhs) == 1 {
		lhs := convertExprStr(fset, s.Lhs[0], src)
		rhs := convertExprStr(fset, s.Rhs[0], src)
		if activeStringVars[lhs] {
			if slice, ok := s.Rhs[0].(*ast.SliceExpr); ok && sliceExprBaseName(slice) == lhs {
				return fmt.Sprintf("%s%s = (%s).to_string();\n", ind, lhs, rhs)
			}
		}
		if strings.HasSuffix(lhs, ".coeff") && !strings.HasSuffix(rhs, ".clone()") {
			rhs += ".clone()"
		}
		if s.Tok == token.AND_NOT_ASSIGN {
			return fmt.Sprintf("%s%s &= !%s;\n", ind, lhs, rhs)
		}
		if shifted, ok := checkedShiftAssignExpr(s.Tok, s.Lhs[0], s.Rhs[0], lhs, rhs); ok {
			return fmt.Sprintf("%s%s = %s;\n", ind, lhs, shifted)
		}
		if method, ok := wrappingAssignMethod(s.Tok, s.Lhs[0]); ok {
			rhs = castIntegerExprToContext(rhs, rustExprValueType(s.Rhs[0]), integerRustType(s.Lhs[0]), s.Rhs[0])
			return fmt.Sprintf("%s%s = %s.%s(%s);\n", ind, lhs, lhs, method, rhs)
		}
		if s.Tok == token.ASSIGN {
			rhs = castIntegerExprToContext(rhs, rustExprValueType(s.Rhs[0]), integerRustType(s.Lhs[0]), s.Rhs[0])
		}
		return fmt.Sprintf("%s%s %s %s;\n", ind, lhs, convertAssignOp(s.Tok), rhs)
	}

	// Multi-assign with single RHS (function call)
	if len(s.Rhs) == 1 {
		var lhsParts []string
		for _, l := range s.Lhs {
			lhsParts = append(lhsParts, convertExprStr(fset, l, src))
		}
		rhs := convertExprStr(fset, s.Rhs[0], src)
		lhsStr := "(" + strings.Join(lhsParts, ", ") + ")"
		return fmt.Sprintf("%s%s = %s;\n", ind, lhsStr, rhs)
	}

	// Multi-assign with multiple RHS
	var sb strings.Builder
	for i := range s.Lhs {
		if i < len(s.Rhs) {
			lhs := convertExprStr(fset, s.Lhs[i], src)
			rhs := convertExprStr(fset, s.Rhs[i], src)
			if shifted, ok := checkedShiftAssignExpr(s.Tok, s.Lhs[i], s.Rhs[i], lhs, rhs); ok {
				sb.WriteString(fmt.Sprintf("%s%s = %s;\n", ind, lhs, shifted))
				continue
			}
			if method, ok := wrappingAssignMethod(s.Tok, s.Lhs[i]); ok {
				rhs = castIntegerExprToContext(rhs, rustExprValueType(s.Rhs[i]), integerRustType(s.Lhs[i]), s.Rhs[i])
				sb.WriteString(fmt.Sprintf("%s%s = %s.%s(%s);\n", ind, lhs, lhs, method, rhs))
				continue
			}
			if s.Tok == token.ASSIGN {
				rhs = castIntegerExprToContext(rhs, rustExprValueType(s.Rhs[i]), integerRustType(s.Lhs[i]), s.Rhs[i])
			}
			sb.WriteString(fmt.Sprintf("%s%s %s %s;\n", ind, lhs, convertAssignOp(s.Tok), rhs))
		}
	}
	return sb.String()
}

func inferLocalLiteralType(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.BasicLit:
		switch e.Kind {
		case token.INT:
			if _, ok := parseGoIntLiteral(e.Value); ok {
				return "i64"
			}
			if _, ok := parseGoUintLiteral(e.Value); ok {
				return "u64"
			}
		case token.FLOAT:
			return "f64"
		case token.CHAR:
			return "u8"
		}
	case *ast.UnaryExpr:
		if e.Op == token.SUB {
			if _, ok := e.X.(*ast.BasicLit); ok {
				if _, ok := parseGoSignedLiteral(e); ok {
					return "i64"
				}
			}
		}
	}
	return ""
}

func parseGoSignedLiteral(expr ast.Expr) (int64, bool) {
	switch e := expr.(type) {
	case *ast.BasicLit:
		return parseGoIntLiteral(e.Value)
	case *ast.UnaryExpr:
		if e.Op != token.SUB {
			return 0, false
		}
		lit, ok := e.X.(*ast.BasicLit)
		if !ok {
			return 0, false
		}
		inner, ok := parseGoIntLiteral(lit.Value)
		if !ok {
			return 0, false
		}
		return -inner, true
	default:
		return 0, false
	}
}

func parseGoIntLiteral(value string) (int64, bool) {
	v, err := strconv.ParseInt(value, 0, 64)
	if err == nil {
		return v, true
	}
	return 0, false
}

func parseGoUintLiteral(value string) (uint64, bool) {
	v, err := strconv.ParseUint(value, 0, 64)
	if err == nil {
		return v, true
	}
	return 0, false
}

func convertAssignOp(tok token.Token) string {
	switch tok {
	case token.ASSIGN, token.DEFINE:
		return "="
	case token.ADD_ASSIGN:
		return "+="
	case token.SUB_ASSIGN:
		return "-="
	case token.MUL_ASSIGN:
		return "*="
	case token.QUO_ASSIGN:
		return "/="
	case token.REM_ASSIGN:
		return "%="
	case token.AND_ASSIGN:
		return "&="
	case token.OR_ASSIGN:
		return "|="
	case token.XOR_ASSIGN:
		return "^="
	case token.SHL_ASSIGN:
		return "<<="
	case token.SHR_ASSIGN:
		return ">>="
	default:
		return "="
	}
}

func checkedShiftHelper(expr *ast.BinaryExpr) (string, bool) {
	if isIntegerConstExpr(expr) || !isIntegerExpr(expr.X) {
		return "", false
	}
	helper := ""
	switch expr.Op {
	case token.SHL:
		helper = shiftHelper("shl", expr.X)
	case token.SHR:
		helper = shiftHelper("shr", expr.X)
	default:
		return "", false
	}
	return helper, helper != ""
}

func checkedShiftExpr(tok token.Token, lhs, rhs ast.Expr, left, right string) (string, bool) {
	helper, ok := checkedShiftTokenMethod(tok, lhs)
	if !ok {
		return "", false
	}
	return fmt.Sprintf("%s(%s, %s)", helper, left, shiftCountExpr(rhs, right)), true
}

func checkedShiftAssignExpr(tok token.Token, lhs, rhs ast.Expr, left, right string) (string, bool) {
	switch tok {
	case token.SHL_ASSIGN:
		return checkedShiftExpr(token.SHL, lhs, rhs, left, right)
	case token.SHR_ASSIGN:
		return checkedShiftExpr(token.SHR, lhs, rhs, left, right)
	default:
		return "", false
	}
}

func checkedShiftTokenMethod(tok token.Token, lhs ast.Expr) (string, bool) {
	if !isIntegerExpr(lhs) {
		return "", false
	}
	helper := ""
	switch tok {
	case token.SHL:
		helper = shiftHelper("shl", lhs)
	case token.SHR:
		helper = shiftHelper("shr", lhs)
	default:
		return "", false
	}
	return helper, helper != ""
}

func shiftHelper(direction string, expr ast.Expr) string {
	rustType := integerRustType(expr)
	switch rustType {
	case "u8", "u16", "u32", "u64", "usize", "i8", "i16", "i32", "i64":
		return fmt.Sprintf("go_checked_%s_%s", direction, rustType)
	default:
		return ""
	}
}

func shiftCountExpr(expr ast.Expr, right string) string {
	if isSignedIntegerExpr(expr) && !isIntegerConstExpr(expr) {
		return fmt.Sprintf("go_shift_count_i64((%s) as i64)", right)
	}
	return fmt.Sprintf("go_shift_count_u64((%s) as u64)", right)
}

func wrappingAssignMethod(tok token.Token, lhs ast.Expr) (string, bool) {
	if !isIntegerExpr(lhs) {
		return "", false
	}
	switch tok {
	case token.ADD_ASSIGN:
		return "wrapping_add", true
	case token.SUB_ASSIGN:
		return "wrapping_sub", true
	case token.MUL_ASSIGN:
		return "wrapping_mul", true
	default:
		return "", false
	}
}

func wrappingBinaryMethod(tok token.Token, expr ast.Expr) (string, bool) {
	if isIntegerConstExpr(expr) {
		return "", false
	}
	if !isIntegerExpr(expr) {
		return "", false
	}
	switch tok {
	case token.ADD:
		return "wrapping_add", true
	case token.SUB:
		return "wrapping_sub", true
	case token.MUL:
		return "wrapping_mul", true
	default:
		return "", false
	}
}

func isIntegerExpr(expr ast.Expr) bool {
	if activeTypeInfo == nil {
		return false
	}
	tv, ok := activeTypeInfo.Types[expr]
	if ok && tv.Type != nil {
		basic, ok := tv.Type.Underlying().(*types.Basic)
		return ok && (basic.Info()&types.IsInteger) != 0
	}
	return integerObjectType(expr) != nil
}

func isUnsignedIntegerExpr(expr ast.Expr) bool {
	if activeTypeInfo == nil {
		return false
	}
	tv, ok := activeTypeInfo.Types[expr]
	if ok && tv.Type != nil {
		return isUnsignedIntegerType(tv.Type)
	}
	if typ := integerObjectType(expr); typ != nil {
		return isUnsignedIntegerType(typ)
	}
	return false
}

func isSignedIntegerExpr(expr ast.Expr) bool {
	if activeTypeInfo == nil {
		return false
	}
	tv, ok := activeTypeInfo.Types[expr]
	if ok && tv.Type != nil {
		return isSignedIntegerType(tv.Type)
	}
	if typ := integerObjectType(expr); typ != nil {
		return isSignedIntegerType(typ)
	}
	return false
}

func isSignedIntegerType(typ types.Type) bool {
	basic, ok := typ.Underlying().(*types.Basic)
	if !ok || (basic.Info()&types.IsInteger) == 0 {
		return false
	}
	return (basic.Info() & types.IsUnsigned) == 0
}

func isUnsignedIntegerType(typ types.Type) bool {
	basic, ok := typ.Underlying().(*types.Basic)
	if !ok || (basic.Info()&types.IsUnsigned) == 0 {
		return false
	}
	return (basic.Info() & types.IsInteger) != 0
}

func castIntegerExprToContext(code, fromType, toType string, expr ast.Expr) string {
	if fromType == "" || toType == "" || fromType == toType {
		return code
	}
	if !isIntegerRustType(fromType) || !isIntegerRustType(toType) {
		return code
	}
	if isUntypedIntegerConstExpr(expr) {
		return code
	}
	return fmt.Sprintf("(%s as %s)", code, toType)
}

func isIntegerRustType(t string) bool {
	switch t {
	case "u8", "u16", "u32", "u64", "usize", "i8", "i16", "i32", "i64":
		return true
	default:
		return false
	}
}

func rustExprValueType(expr ast.Expr) string {
	if typ := registryBackedRustValueType(expr); typ != "" {
		return typ
	}
	return integerRustType(expr)
}

func registryBackedRustValueType(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.IndexExpr:
		if typ := registryBackedRustValueType(e.X); strings.HasPrefix(typ, "[") {
			return arrayElementRustType(typ)
		}
		if ident, ok := e.X.(*ast.Ident); ok && activeRegistry != nil {
			if td, ok := activeRegistry.Tables[ident.Name]; ok {
				return tableElementRustType(td.ElementType)
			}
		}
	case *ast.ParenExpr:
		return registryBackedRustValueType(e.X)
	}
	return ""
}

func tableElementRustType(t string) string {
	return t
}

func arrayElementRustType(t string) string {
	if !strings.HasPrefix(t, "[") {
		return ""
	}
	depth := 0
	for i, r := range t {
		switch r {
		case '[':
			depth++
		case ']':
			depth--
		case ';':
			if depth == 1 {
				return strings.TrimSpace(t[1:i])
			}
		}
	}
	return ""
}

func integerRustType(expr ast.Expr) string {
	if activeTypeInfo == nil {
		return ""
	}
	tv, ok := activeTypeInfo.Types[expr]
	if ok && tv.Type != nil {
		if tv.Value != nil {
			if basic, ok := tv.Type.Underlying().(*types.Basic); ok && (basic.Info()&types.IsUntyped) != 0 && (basic.Info()&types.IsInteger) != 0 {
				return rustTypeForUntypedIntegerValue(tv.Value)
			}
		}
		return rustTypeForGoInteger(tv.Type)
	}
	if typ := integerObjectType(expr); typ != nil {
		return rustTypeForGoInteger(typ)
	}
	return ""
}

func integerObjectType(expr ast.Expr) types.Type {
	ident, ok := expr.(*ast.Ident)
	if !ok || activeTypeInfo == nil {
		return nil
	}
	if obj := activeTypeInfo.Defs[ident]; obj != nil && obj.Type() != nil {
		return obj.Type()
	}
	if obj := activeTypeInfo.Uses[ident]; obj != nil && obj.Type() != nil {
		return obj.Type()
	}
	return nil
}

func rustTypeForGoInteger(typ types.Type) string {
	basic, ok := typ.Underlying().(*types.Basic)
	if !ok || (basic.Info()&types.IsInteger) == 0 {
		return ""
	}
	switch basic.Kind() {
	case types.Int:
		return "i64"
	case types.Int8:
		return "i8"
	case types.Int16:
		return "i16"
	case types.Int32:
		return "i32"
	case types.Int64:
		return "i64"
	case types.Uint:
		return "u64"
	case types.Uint8:
		return "u8"
	case types.Uint16:
		return "u16"
	case types.Uint32:
		return "u32"
	case types.Uint64:
		return "u64"
	case types.Uintptr:
		return "usize"
	default:
		return ""
	}
}

func rustTypeForUntypedIntegerValue(value constant.Value) string {
	if value.Kind() != constant.Int {
		return "i64"
	}
	if _, ok := constant.Int64Val(value); ok {
		return "i64"
	}
	if _, ok := constant.Uint64Val(value); ok {
		return "u64"
	}
	return "i64"
}

func isUntypedIntegerConstExpr(expr ast.Expr) bool {
	if activeTypeInfo != nil {
		if tv, ok := activeTypeInfo.Types[expr]; ok && tv.Type != nil && tv.Value != nil {
			if basic, ok := tv.Type.Underlying().(*types.Basic); ok {
				return (basic.Info()&types.IsUntyped) != 0 && (basic.Info()&types.IsInteger) != 0
			}
		}
	}
	switch e := expr.(type) {
	case *ast.BasicLit:
		return e.Kind == token.INT || e.Kind == token.CHAR
	case *ast.UnaryExpr:
		return e.Op == token.SUB && isUntypedIntegerConstExpr(e.X)
	case *ast.ParenExpr:
		return isUntypedIntegerConstExpr(e.X)
	default:
		return false
	}
}

func isIntegerConstExpr(expr ast.Expr) bool {
	if activeTypeInfo != nil {
		if tv, ok := activeTypeInfo.Types[expr]; ok && tv.Type != nil && tv.Value != nil {
			if basic, ok := tv.Type.Underlying().(*types.Basic); ok && (basic.Info()&types.IsInteger) != 0 {
				return true
			}
		}
	}
	switch e := expr.(type) {
	case *ast.BasicLit:
		return e.Kind == token.INT || e.Kind == token.CHAR
	case *ast.UnaryExpr:
		return e.Op == token.SUB && isIntegerConstExpr(e.X)
	case *ast.ParenExpr:
		return isIntegerConstExpr(e.X)
	default:
		return false
	}
}

func isRustIntegerLiteralText(text string) bool {
	return regexp.MustCompile(`^-?(?:0x[0-9a-fA-F_]+|[0-9][0-9_]*)$`).MatchString(text)
}

// -------------------------------------------------------
// ReturnStmt
// -------------------------------------------------------

func convertReturnStmt(fset *token.FileSet, s *ast.ReturnStmt, src []byte, indent int, namedReturns []namedReturn) string {
	ind := strings.Repeat("    ", indent)

	if len(s.Results) == 0 {
		// Naked return with named returns
		if len(namedReturns) == 1 {
			return fmt.Sprintf("%sreturn %s;\n", ind, namedReturns[0].name)
		} else if len(namedReturns) > 1 {
			var names []string
			for _, nr := range namedReturns {
				names = append(names, nr.name)
			}
			return fmt.Sprintf("%sreturn (%s);\n", ind, strings.Join(names, ", "))
		}
		return fmt.Sprintf("%sreturn;\n", ind)
	}

	if len(s.Results) == 1 {
		val := convertExprStr(fset, s.Results[0], src)
		if activeReturnType == "String" {
			if lit, ok := s.Results[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
				val = fmt.Sprintf("%s.to_string()", val)
			}
		} else {
			val = castIntegerExprToContext(val, rustExprValueType(s.Results[0]), activeReturnType, s.Results[0])
		}
		return fmt.Sprintf("%sreturn %s;\n", ind, val)
	}

	var vals []string
	for _, r := range s.Results {
		vals = append(vals, convertExprStr(fset, r, src))
	}
	return fmt.Sprintf("%sreturn (%s);\n", ind, strings.Join(vals, ", "))
}

// -------------------------------------------------------
// IfStmt
// -------------------------------------------------------

func convertIfStmt(fset *token.FileSet, s *ast.IfStmt, src []byte, indent int, namedReturns []namedReturn) string {
	ind := strings.Repeat("    ", indent)
	var sb strings.Builder

	// Init statement
	if s.Init != nil {
		sb.WriteString(convertStmt(fset, s.Init, src, indent, namedReturns))
	}

	cond := convertExprStr(fset, s.Cond, src)
	sb.WriteString(fmt.Sprintf("%sif %s {\n", ind, cond))

	// Body
	for _, stmt := range s.Body.List {
		sb.WriteString(convertStmt(fset, stmt, src, indent+1, namedReturns))
	}

	// Else
	if s.Else != nil {
		switch e := s.Else.(type) {
		case *ast.IfStmt:
			sb.WriteString(fmt.Sprintf("%s} else ", ind))
			// Remove the indent from the recursive call
			elseCode := convertIfStmt(fset, e, src, indent, namedReturns)
			// Strip leading whitespace to join with "} else "
			elseCode = strings.TrimLeft(elseCode, " \t")
			sb.WriteString(elseCode)
			return sb.String()
		case *ast.BlockStmt:
			sb.WriteString(fmt.Sprintf("%s} else {\n", ind))
			for _, stmt := range e.List {
				sb.WriteString(convertStmt(fset, stmt, src, indent+1, namedReturns))
			}
		}
	}

	sb.WriteString(fmt.Sprintf("%s}\n", ind))
	return sb.String()
}

// -------------------------------------------------------
// ForStmt
// -------------------------------------------------------

func convertForStmt(fset *token.FileSet, s *ast.ForStmt, src []byte, indent int, namedReturns []namedReturn) string {
	ind := strings.Repeat("    ", indent)
	var sb strings.Builder

	if s.Init == nil && s.Cond == nil && s.Post == nil {
		// Infinite loop
		sb.WriteString(fmt.Sprintf("%sloop {\n", ind))
		for _, stmt := range s.Body.List {
			sb.WriteString(convertStmt(fset, stmt, src, indent+1, namedReturns))
		}
		sb.WriteString(fmt.Sprintf("%s}\n", ind))
		return sb.String()
	}

	if s.Init == nil && s.Post == nil && s.Cond != nil {
		// while loop
		cond := convertExprStr(fset, s.Cond, src)
		sb.WriteString(fmt.Sprintf("%swhile %s {\n", ind, cond))
		for _, stmt := range s.Body.List {
			sb.WriteString(convertStmt(fset, stmt, src, indent+1, namedReturns))
		}
		sb.WriteString(fmt.Sprintf("%s}\n", ind))
		return sb.String()
	}

	// C-style for loop: emit init, then while loop with post at end
	if s.Init != nil {
		sb.WriteString(convertStmt(fset, s.Init, src, indent, namedReturns))
	}
	cond := "true"
	if s.Cond != nil {
		cond = convertExprStr(fset, s.Cond, src)
	}
	sb.WriteString(fmt.Sprintf("%swhile %s {\n", ind, cond))
	for _, stmt := range s.Body.List {
		sb.WriteString(convertStmt(fset, stmt, src, indent+1, namedReturns))
	}
	if s.Post != nil {
		sb.WriteString(convertStmt(fset, s.Post, src, indent+1, namedReturns))
	}
	sb.WriteString(fmt.Sprintf("%s}\n", ind))
	return sb.String()
}

// -------------------------------------------------------
// RangeStmt
// -------------------------------------------------------

func convertRangeStmt(fset *token.FileSet, s *ast.RangeStmt, src []byte, indent int, namedReturns []namedReturn) string {
	ind := strings.Repeat("    ", indent)
	var sb strings.Builder

	x := convertExprStr(fset, s.X, src)

	key := "_"
	val := "_"
	if s.Key != nil {
		key = convertExprStr(fset, s.Key, src)
	}
	if s.Value != nil {
		val = convertExprStr(fset, s.Value, src)
	}

	if val == "_" && key != "_" {
		sb.WriteString(fmt.Sprintf("%sfor %s in 0..%s.len() {\n", ind, key, x))
	} else if key != "_" && val != "_" {
		sb.WriteString(fmt.Sprintf("%sfor (%s, %s) in %s.iter().enumerate() {\n", ind, key, val, x))
	} else {
		sb.WriteString(fmt.Sprintf("%sfor _ in %s.iter() {\n", ind, x))
	}

	for _, stmt := range s.Body.List {
		sb.WriteString(convertStmt(fset, stmt, src, indent+1, namedReturns))
	}
	sb.WriteString(fmt.Sprintf("%s}\n", ind))
	return sb.String()
}

// -------------------------------------------------------
// SwitchStmt
// -------------------------------------------------------

func convertSwitchStmt(fset *token.FileSet, s *ast.SwitchStmt, src []byte, indent int, namedReturns []namedReturn) string {
	ind := strings.Repeat("    ", indent)
	var sb strings.Builder

	if s.Init != nil {
		sb.WriteString(convertStmt(fset, s.Init, src, indent, namedReturns))
	}

	tag := ""
	if s.Tag != nil {
		tag = convertExprStr(fset, s.Tag, src)
	}

	if tag != "" {
		sb.WriteString(fmt.Sprintf("%smatch %s {\n", ind, tag))
	} else {
		// No tag = switch on boolean conditions
		// Convert to if/else chain
		return convertBoolSwitch(fset, s, src, indent, namedReturns)
	}

	hasDefault := false
	for _, stmt := range s.Body.List {
		cc, ok := stmt.(*ast.CaseClause)
		if !ok {
			continue
		}
		if cc.List == nil {
			// default
			hasDefault = true
			sb.WriteString(fmt.Sprintf("%s    _ => {\n", ind))
		} else {
			var conds []string
			for _, c := range cc.List {
				conds = append(conds, convertExprStr(fset, c, src))
			}
			sb.WriteString(fmt.Sprintf("%s    %s => {\n", ind, strings.Join(conds, " | ")))
		}
		for _, bodyStmt := range cc.Body {
			sb.WriteString(convertStmt(fset, bodyStmt, src, indent+2, namedReturns))
		}
		sb.WriteString(fmt.Sprintf("%s    }\n", ind))
	}
	if !hasDefault {
		sb.WriteString(fmt.Sprintf("%s    _ => {}\n", ind))
	}

	sb.WriteString(fmt.Sprintf("%s}\n", ind))
	return sb.String()
}

func convertBoolSwitch(fset *token.FileSet, s *ast.SwitchStmt, src []byte, indent int, namedReturns []namedReturn) string {
	ind := strings.Repeat("    ", indent)
	var sb strings.Builder

	first := true
	for _, stmt := range s.Body.List {
		cc, ok := stmt.(*ast.CaseClause)
		if !ok {
			continue
		}
		if cc.List == nil {
			sb.WriteString(fmt.Sprintf("%s} else {\n", ind))
		} else {
			var conds []string
			for _, c := range cc.List {
				conds = append(conds, convertExprStr(fset, c, src))
			}
			if first {
				sb.WriteString(fmt.Sprintf("%sif %s {\n", ind, strings.Join(conds, " || ")))
				first = false
			} else {
				sb.WriteString(fmt.Sprintf("%s} else if %s {\n", ind, strings.Join(conds, " || ")))
			}
		}
		for _, bodyStmt := range cc.Body {
			sb.WriteString(convertStmt(fset, bodyStmt, src, indent+1, namedReturns))
		}
	}
	if !first {
		sb.WriteString(fmt.Sprintf("%s}\n", ind))
	}
	return sb.String()
}

// -------------------------------------------------------
// Expression conversion (source-level)
// -------------------------------------------------------

func convertExprStr(fset *token.FileSet, expr ast.Expr, src []byte) string {
	if expr == nil {
		return ""
	}

	switch e := expr.(type) {
	case *ast.Ident:
		return convertIdent(e)

	case *ast.BasicLit:
		return convertBasicLit(e)

	case *ast.BinaryExpr:
		if nilCmp, ok := convertNilComparison(e.X, e.Y, e.Op); ok {
			return nilCmp
		}
		left := convertExprStr(fset, e.X, src)
		right := convertExprStr(fset, e.Y, src)
		if e.Op == token.AND_NOT {
			return fmt.Sprintf("(%s & !%s)", left, right)
		}
		if _, ok := checkedShiftHelper(e); ok {
			if shifted, ok := checkedShiftExpr(e.Op, e.X, e.Y, left, right); ok {
				return fmt.Sprintf("(%s)", shifted)
			}
		}
		if method, ok := wrappingBinaryMethod(e.Op, e); ok {
			if rustType := integerRustType(e); rustType != "" && (isIntegerConstExpr(e.X) || isRustIntegerLiteralText(left)) {
				left = fmt.Sprintf("(%s as %s)", left, rustType)
			} else if rustType := integerRustType(e); rustType != "" {
				left = castIntegerExprToContext(left, rustExprValueType(e.X), rustType, e.X)
			}
			if rustType := integerRustType(e); rustType != "" {
				right = castIntegerExprToContext(right, rustExprValueType(e.Y), rustType, e.Y)
			}
			return fmt.Sprintf("(%s.%s(%s))", left, method, right)
		}
		op := e.Op.String()
		return fmt.Sprintf("(%s %s %s)", left, op, right)

	case *ast.UnaryExpr:
		x := convertExprStr(fset, e.X, src)
		if e.Op == token.AND {
			return fmt.Sprintf("(&mut %s)", strings.TrimPrefix(x, "&mut "))
		}
		if e.Op == token.SUB && isIntegerExpr(e.X) && !isIntegerConstExpr(e.X) {
			return fmt.Sprintf("(%s.wrapping_neg())", x)
		}
		if e.Op == token.XOR {
			// Go's ^x (bitwise complement) = Rust's !x
			return fmt.Sprintf("(!%s)", x)
		}
		return fmt.Sprintf("(%s%s)", e.Op.String(), x)

	case *ast.ParenExpr:
		inner := convertExprStr(fset, e.X, src)
		return fmt.Sprintf("(%s)", inner)

	case *ast.CallExpr:
		return convertCallExpr(fset, e, src)

	case *ast.IndexExpr:
		x := convertExprStr(fset, e.X, src)
		idx := convertExprStr(fset, e.Index, src)
		if ident, ok := e.X.(*ast.Ident); ok && activeStringVars[rustIdent(ident.Name)] {
			return fmt.Sprintf("%s.as_bytes()[%s as usize]", x, idx)
		}
		return fmt.Sprintf("%s[%s as usize]", x, idx)

	case *ast.SelectorExpr:
		if ident, ok := e.X.(*ast.Ident); ok {
			switch ident.Name + "." + e.Sel.Name {
			case "runtime.GOOS":
				return "std::env::consts::OS"
			case "bits.UintSize":
				return "(usize::BITS as i64)"
			}
		}
		x := convertExprStr(fset, e.X, src)
		return fmt.Sprintf("%s.%s", x, e.Sel.Name)

	case *ast.StarExpr:
		x := convertExprStr(fset, e.X, src)
		return fmt.Sprintf("(*%s)", x)

	case *ast.CompositeLit:
		return convertCompositeLit(fset, e, src)

	case *ast.SliceExpr:
		x := convertExprStr(fset, e.X, src)
		low := ""
		high := ""
		if e.Low != nil {
			low = convertExprStr(fset, e.Low, src) + " as usize"
		}
		if e.High != nil {
			high = convertExprStr(fset, e.High, src) + " as usize"
		}
		prefix := "&mut "
		if ident, ok := e.X.(*ast.Ident); ok && activeStringVars[rustIdent(ident.Name)] {
			prefix = "&"
		}
		if low != "" && high != "" {
			return fmt.Sprintf("%s%s[%s..%s]", prefix, x, low, high)
		} else if low != "" {
			return fmt.Sprintf("%s%s[%s..]", prefix, x, low)
		} else if high != "" {
			return fmt.Sprintf("%s%s[..%s]", prefix, x, high)
		}
		return fmt.Sprintf("%s%s[..]", prefix, x)

	case *ast.TypeAssertExpr:
		x := convertExprStr(fset, e.X, src)
		return x // type assertions don't exist in Rust

	case *ast.FuncLit:
		return convertFuncLitExpr(fset, e, src)

	default:
		// Fallback
		text := extractSource(fset, expr, src)
		return "/* " + text + " */"
	}
}

func convertIdent(e *ast.Ident) string {
	switch e.Name {
	case "true", "false":
		return e.Name
	case "nil":
		return "\"\""
	}
	if lit := constLiteral(e.Name); lit != "" {
		return lit
	}
	if activeRegistry != nil {
		if _, ok := activeRegistry.Functions[e.Name]; ok {
			return rustIdent(goFuncNameToRust(e.Name))
		}
	}
	if activeSourceFunctions[e.Name] {
		return rustIdent(goFuncNameToRust(e.Name))
	}
	return rustIdent(e.Name)
}

func convertNilComparison(left, right ast.Expr, op token.Token) (string, bool) {
	if op != token.NEQ && op != token.EQL {
		return "", false
	}
	var other ast.Expr
	if isNilIdent(left) {
		other = right
	} else if isNilIdent(right) {
		other = left
	} else {
		return "", false
	}
	if ident, ok := other.(*ast.Ident); ok && strings.Contains(strings.ToLower(ident.Name), "err") {
		if op == token.NEQ {
			return fmt.Sprintf("%s.is_some()", rustIdent(ident.Name)), true
		}
		return fmt.Sprintf("%s.is_none()", rustIdent(ident.Name)), true
	}
	return "/* unsupported nil comparison */", true
}

func isNilIdent(expr ast.Expr) bool {
	ident, ok := expr.(*ast.Ident)
	return ok && ident.Name == "nil"
}

func convertBasicLit(e *ast.BasicLit) string {
	switch e.Kind {
	case token.INT:
		return e.Value
	case token.FLOAT:
		return e.Value
	case token.CHAR:
		return "b" + e.Value
	case token.STRING:
		return e.Value
	}
	return e.Value
}

func convertCallExpr(fset *token.FileSet, e *ast.CallExpr, src []byte) string {
	if ident, ok := e.Fun.(*ast.Ident); ok {
		switch ident.Name {
		case "len":
			if len(e.Args) == 1 {
				arg := convertExprStr(fset, e.Args[0], src)
				return fmt.Sprintf("(%s.len() as i64)", arg)
			}
		case "new":
			if len(e.Args) == 1 {
				if isBigIntType(e.Args[0]) {
					return "BigUint::zero()"
				}
				return fmt.Sprintf("todo!(\"new(%s)\")", convertExprStr(fset, e.Args[0], src))
			}
		case "copy":
			if len(e.Args) == 2 {
				dst := strings.TrimPrefix(convertExprStr(fset, e.Args[0], src), "&mut ")
				srcArg := convertExprStr(fset, e.Args[1], src)
				return fmt.Sprintf("go_copy_str(&mut %s, &%s)", dst, srcArg)
			}
		case "append":
			if len(e.Args) == 2 {
				vec := convertExprStr(fset, e.Args[0], src)
				item := convertExprStr(fset, e.Args[1], src)
				return fmt.Sprintf("go_append(%s, %s)", vec, item)
			}
		}
	}

	// Check for type casts
	if len(e.Args) == 1 {
		if ident, ok := e.Fun.(*ast.Ident); ok {
			if ident.Name == "string" {
				arg := convertExprStr(fset, e.Args[0], src)
				return fmt.Sprintf("go_string_from_bytes(%s)", arg)
			}
			if ident.Name == "byte" {
				return fmt.Sprintf("(%s as u8)", convertByteCastArg(fset, e.Args[0], src))
			}
			if rsType, isCast := typeMap[ident.Name]; isCast {
				arg := convertExprStr(fset, e.Args[0], src)
				return fmt.Sprintf("(%s as %s)", arg, rsType)
			}
			if registryTypeAlias(ident.Name) {
				arg := convertExprStr(fset, e.Args[0], src)
				return fmt.Sprintf("(%s as %s)", arg, ident.Name)
			}
		}
		if arr, ok := e.Fun.(*ast.ArrayType); ok && arr.Len == nil {
			if elt, ok := arr.Elt.(*ast.Ident); ok && elt.Name == "byte" {
				arg := convertExprStr(fset, e.Args[0], src)
				return fmt.Sprintf("(%s).as_bytes().to_vec()", arg)
			}
		}
	}

	// Check for specific function calls
	if sel, ok := e.Fun.(*ast.SelectorExpr); ok {
		if pkg, ok := sel.X.(*ast.Ident); ok {
			if knownGoPackages[pkg.Name] {
				return convertPkgCall(fset, pkg.Name, sel.Sel.Name, e.Args, src)
			}
		}
		return convertMethodCall(fset, sel, e.Args, src)
	}

	// Regular function call
	rsFunc := convertExprStr(fset, e.Fun, src)
	if ident, ok := e.Fun.(*ast.Ident); ok {
		if activeRegistry != nil {
			if _, exists := activeRegistry.Functions[ident.Name]; exists {
				rsFunc = rustIdent(goFuncNameToRust(ident.Name))
			} else if activeSourceFunctions[ident.Name] {
				rsFunc = rustIdent(goFuncNameToRust(ident.Name))
			} else {
				rsFunc = rustIdent(ident.Name)
			}
		} else if activeSourceFunctions[ident.Name] {
			rsFunc = rustIdent(goFuncNameToRust(ident.Name))
		} else {
			rsFunc = rustIdent(ident.Name)
		}
	}

	var args []string
	for i, arg := range e.Args {
		if ident, ok := e.Fun.(*ast.Ident); ok && ident.Name == "pack" && i == 2 && isNilIdent(arg) {
			args = append(args, "&BigUint::zero()")
			continue
		}
		argCode := convertExprStr(fset, arg, src)
		if shouldBorrowBigIntArg(rsFunc, i) {
			argCode = borrowRustExpr(argCode)
		} else if wantType := callParamRustType(e, i); wantType != "" {
			argCode = castIntegerExprToContext(argCode, rustExprValueType(arg), wantType, arg)
		}
		args = append(args, argCode)
	}

	return fmt.Sprintf("%s(%s)", rsFunc, strings.Join(args, ", "))
}

func convertByteCastArg(fset *token.FileSet, expr ast.Expr, src []byte) string {
	if bin, ok := expr.(*ast.BinaryExpr); ok && (bin.Op == token.ADD || bin.Op == token.SUB) {
		if lit, ok := bin.X.(*ast.BasicLit); ok && lit.Kind == token.CHAR {
			left := convertBasicLit(lit)
			right := convertExprStr(fset, bin.Y, src)
			return fmt.Sprintf("(%s %s (%s as u8))", left, bin.Op.String(), right)
		}
	}
	return convertExprStr(fset, expr, src)
}

func registryTypeAlias(name string) bool {
	if activeRegistry == nil {
		return false
	}
	td, ok := activeRegistry.Types[name]
	return ok && td.AliasOf != ""
}

func shouldBorrowBigIntArg(funcName string, idx int) bool {
	args, ok := bigIntBorrowArgs[funcName]
	if !ok {
		return false
	}
	return args[idx]
}

func callParamRustType(call *ast.CallExpr, idx int) string {
	if activeTypeInfo == nil {
		return ""
	}
	tv, ok := activeTypeInfo.Types[call.Fun]
	if !ok || tv.Type == nil {
		return ""
	}
	sig, ok := tv.Type.Underlying().(*types.Signature)
	if !ok || sig.Params() == nil || idx >= sig.Params().Len() {
		return ""
	}
	return rustTypeForGoParam(sig.Params().At(idx).Type())
}

func rustTypeForGoParam(typ types.Type) string {
	if typ == nil {
		return ""
	}
	if named, ok := typ.(*types.Named); ok {
		if named.Obj() != nil && named.Obj().Name() == "RoundingMode" {
			return "i32"
		}
	}
	return rustTypeForGoInteger(typ)
}

func borrowRustExpr(expr string) string {
	if strings.HasPrefix(expr, "&") {
		return expr
	}
	return "&" + expr
}

func convertMethodCall(fset *token.FileSet, sel *ast.SelectorExpr, args []ast.Expr, src []byte) string {
	recv := convertExprStr(fset, sel.X, src)
	arg := func(i int) string {
		if i >= len(args) {
			return ""
		}
		return convertExprStr(fset, args[i], src)
	}
	shift := func(i int) string {
		return fmt.Sprintf("(%s as usize)", arg(i))
	}

	switch sel.Sel.Name {
	case "SetUint64":
		if len(args) == 1 {
			return fmt.Sprintf("BigUint::from(%s)", arg(0))
		}
	case "Set":
		if len(args) == 1 {
			return fmt.Sprintf("%s.clone()", arg(0))
		}
	case "Exp":
		if len(args) == 3 {
			return fmt.Sprintf("%s.pow(go_big_to_u64(&%s) as u32)", arg(0), arg(1))
		}
	case "Lsh":
		if len(args) == 2 {
			return fmt.Sprintf("(%s.clone() << %s)", arg(0), shift(1))
		}
	case "Rsh":
		if len(args) == 2 {
			return fmt.Sprintf("(%s.clone() >> %s)", arg(0), shift(1))
		}
	case "Or":
		if len(args) == 2 {
			return fmt.Sprintf("(%s.clone() | %s.clone())", arg(0), arg(1))
		}
	case "Add":
		if len(args) == 2 {
			return fmt.Sprintf("(%s.clone() + %s)", arg(0), arg(1))
		}
	case "Sub":
		if len(args) == 2 {
			return fmt.Sprintf("(%s.clone() - %s)", arg(0), arg(1))
		}
	case "Mul":
		if len(args) == 2 {
			return fmt.Sprintf("(%s.clone() * %s)", arg(0), arg(1))
		}
	case "Quo":
		if len(args) == 2 {
			return fmt.Sprintf("(%s.clone() / %s)", arg(0), arg(1))
		}
	case "Sign":
		if len(args) == 0 {
			return fmt.Sprintf("go_big_sign(&%s)", recv)
		}
	case "Cmp":
		if len(args) == 1 {
			return fmt.Sprintf("go_big_cmp(&%s, &%s)", recv, arg(0))
		}
	case "BitLen":
		if len(args) == 0 {
			return fmt.Sprintf("go_big_bit_len(&%s)", recv)
		}
	case "Bit":
		if len(args) == 1 {
			return fmt.Sprintf("go_big_bit(&%s, %s as u64)", recv, arg(0))
		}
	case "Uint64":
		if len(args) == 0 {
			return fmt.Sprintf("go_big_to_u64(&%s)", recv)
		}
	}

	var argStrs []string
	for _, argExpr := range args {
		argStrs = append(argStrs, convertExprStr(fset, argExpr, src))
	}
	return fmt.Sprintf("%s.%s(%s)", recv, rustIdent(camelToSnake(sel.Sel.Name)), strings.Join(argStrs, ", "))
}

func convertFuncLitExpr(fset *token.FileSet, e *ast.FuncLit, src []byte) string {
	params := convertClosureParams(e.Type.Params)
	retType := convertFuncResults(e.Type.Results)
	body := convertBlockStmt(fset, e.Body, src, 1, nil)

	paramList := "||"
	if params != "" {
		paramList = "|" + params + "|"
	}
	if retType != "" {
		return fmt.Sprintf("(%s -> %s {\n%s})", paramList, retType, body)
	}
	return fmt.Sprintf("(%s {\n%s})", paramList, body)
}

func convertClosureParams(fields *ast.FieldList) string {
	if fields == nil || len(fields.List) == 0 {
		return ""
	}

	var parts []string
	for _, field := range fields.List {
		rsType := convertParamType(field.Type)
		_, isPtr := field.Type.(*ast.StarExpr)
		for _, name := range field.Names {
			rsName := rustIdent(name.Name)
			if isPtr {
				parts = append(parts, fmt.Sprintf("%s: %s", rsName, rsType))
			} else {
				parts = append(parts, fmt.Sprintf("mut %s: %s", rsName, rsType))
			}
		}
		if len(field.Names) == 0 {
			parts = append(parts, fmt.Sprintf("_: %s", rsType))
		}
	}
	return strings.Join(parts, ", ")
}

func isBigIntType(expr ast.Expr) bool {
	switch t := expr.(type) {
	case *ast.SelectorExpr:
		pkg, ok := t.X.(*ast.Ident)
		return ok && pkg.Name == "big" && t.Sel.Name == "Int"
	case *ast.StarExpr:
		return isBigIntType(t.X)
	default:
		return false
	}
}

func convertPkgCall(fset *token.FileSet, pkg, fn string, args []ast.Expr, src []byte) string {
	if pkg == "math" {
		switch fn {
		case "Float64bits":
			arg := convertExprStr(fset, args[0], src)
			return fmt.Sprintf("(%s).to_bits()", arg)
		case "Float64frombits":
			arg := convertExprStr(fset, args[0], src)
			return fmt.Sprintf("f64::from_bits(%s)", arg)
		case "Float32bits":
			arg := convertExprStr(fset, args[0], src)
			return fmt.Sprintf("(%s as f32).to_bits()", arg)
		case "Float32frombits":
			arg := convertExprStr(fset, args[0], src)
			return fmt.Sprintf("f32::from_bits(%s)", arg)
		case "Abs":
			arg := convertExprStr(fset, args[0], src)
			return fmt.Sprintf("(%s).abs()", arg)
		case "Sqrt":
			arg := convertExprStr(fset, args[0], src)
			return fmt.Sprintf("(%s).sqrt()", arg)
		case "Floor":
			arg := convertExprStr(fset, args[0], src)
			return fmt.Sprintf("(%s).floor()", arg)
		case "Ceil":
			arg := convertExprStr(fset, args[0], src)
			return fmt.Sprintf("(%s).ceil()", arg)
		case "Log2":
			arg := convertExprStr(fset, args[0], src)
			return fmt.Sprintf("(%s).log2()", arg)
		case "Log10":
			arg := convertExprStr(fset, args[0], src)
			return fmt.Sprintf("(%s).log10()", arg)
		case "Ldexp":
			a := convertExprStr(fset, args[0], src)
			b := convertExprStr(fset, args[1], src)
			return fmt.Sprintf("f64_ldexp(%s, %s)", a, b)
		case "Frexp":
			a := convertExprStr(fset, args[0], src)
			return fmt.Sprintf("f64_frexp(%s)", a)
		}
	}
	if pkg == "bits" {
		switch fn {
		case "Mul64":
			a := convertExprStr(fset, args[0], src)
			b := convertExprStr(fset, args[1], src)
			return fmt.Sprintf("go_mul64(%s, %s)", a, b)
		case "Add64":
			a := convertExprStr(fset, args[0], src)
			b := convertExprStr(fset, args[1], src)
			c := convertExprStr(fset, args[2], src)
			return fmt.Sprintf("go_add64(%s, %s, %s)", a, b, c)
		case "LeadingZeros64":
			a := convertExprStr(fset, args[0], src)
			return fmt.Sprintf("(%s).leading_zeros() as i64", a)
		case "TrailingZeros64":
			a := convertExprStr(fset, args[0], src)
			return fmt.Sprintf("(%s).trailing_zeros() as i64", a)
		case "Len64":
			a := convertExprStr(fset, args[0], src)
			return fmt.Sprintf("(64 - (%s).leading_zeros()) as i64", a)
		}
	}
	if pkg == "strconv" {
		switch fn {
		case "FormatUint":
			a := convertExprStr(fset, args[0], src)
			return fmt.Sprintf("format!(\"{}\", %s)", a)
		case "Itoa":
			a := convertExprStr(fset, args[0], src)
			return fmt.Sprintf("format!(\"{}\", %s)", a)
		case "Atoi":
			a := convertExprStr(fset, args[0], src)
			return fmt.Sprintf("go_atoi(%s)", a)
		}
	}
	if pkg == "strings" {
		switch fn {
		case "Builder":
			return "String::new()"
		case "TrimSpace":
			arg := convertExprStr(fset, args[0], src)
			return fmt.Sprintf("(%s).trim().to_string()", arg)
		case "ToUpper":
			arg := convertExprStr(fset, args[0], src)
			return fmt.Sprintf("(%s).to_ascii_uppercase()", arg)
		case "ToLower":
			arg := convertExprStr(fset, args[0], src)
			return fmt.Sprintf("(%s).to_ascii_lowercase()", arg)
		case "TrimLeft":
			arg := convertExprStr(fset, args[0], src)
			cutset := convertExprStr(fset, args[1], src)
			return fmt.Sprintf("(%s).trim_start_matches(|c| %s.contains(c)).to_string()", arg, cutset)
		case "HasPrefix":
			arg := convertExprStr(fset, args[0], src)
			prefix := convertExprStr(fset, args[1], src)
			return fmt.Sprintf("(%s).starts_with(%s)", arg, prefix)
		}
	}
	if pkg == "fmt" {
		if fn == "Sprintf" && len(args) >= 1 {
			if lit, ok := args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING && lit.Value == "\"%d\"" && len(args) == 2 {
				arg := convertExprStr(fset, args[1], src)
				return fmt.Sprintf("format!(\"{}\", %s)", arg)
			}
		}
	}
	if pkg == "errors" {
		if fn == "New" && len(args) == 1 {
			return convertExprStr(fset, args[0], src)
		}
	}
	if pkg == "big" {
		if fn == "NewInt" && len(args) == 1 {
			arg := convertExprStr(fset, args[0], src)
			return fmt.Sprintf("BigUint::from(%s as u64)", arg)
		}
	}

	// Generic package call
	var argStrs []string
	for _, arg := range args {
		argStrs = append(argStrs, convertExprStr(fset, arg, src))
	}
	return fmt.Sprintf("%s::%s(%s)", rustIdent(pkg), rustIdent(camelToSnake(fn)), strings.Join(argStrs, ", "))
}

func convertCompositeLit(fset *token.FileSet, e *ast.CompositeLit, src []byte) string {
	typeName := ""
	if e.Type != nil {
		typeName = convertTypeExprFull(e.Type)
	}

	var elts []string
	for _, elt := range e.Elts {
		if kv, ok := elt.(*ast.KeyValueExpr); ok {
			key := convertExprStr(fset, kv.Key, src)
			val := convertExprStr(fset, kv.Value, src)
			elts = append(elts, fmt.Sprintf("%s: %s", key, val))
		} else {
			elts = append(elts, convertExprStr(fset, elt, src))
		}
	}

	if typeName != "" {
		// Check if it's an array type
		if strings.HasPrefix(typeName, "[") {
			return fmt.Sprintf("[%s]", strings.Join(elts, ", "))
		}
		// Struct literal
		if len(elts) > 0 && strings.Contains(elts[0], ":") {
			return fmt.Sprintf("%s { %s, ..Default::default() }", typeName, strings.Join(elts, ", "))
		}
		// Array/tuple init
		return fmt.Sprintf("%s { %s }", typeName, strings.Join(elts, ", "))
	}
	return fmt.Sprintf("[%s]", strings.Join(elts, ", "))
}

// -------------------------------------------------------
// Type expression conversion
// -------------------------------------------------------

func convertTypeExpr(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		if rt, ok := typeMap[t.Name]; ok {
			return rt
		}
		return t.Name
	case *ast.StarExpr:
		if isBigIntType(t.X) {
			return "BigUint"
		}
		inner := convertTypeExprFull(t.X)
		if inner == "Int" {
			return "Int"
		}
		return fmt.Sprintf("&mut %s", inner)
	case *ast.FuncType:
		return convertFuncType(t)
	case *ast.ArrayType:
		return convertTypeExprFull(t)
	case *ast.SelectorExpr:
		if pkg, ok := t.X.(*ast.Ident); ok {
			if pkg.Name == "big" && t.Sel.Name == "Int" {
				return "BigUint"
			}
		}
		return convertTypeExpr(t.Sel)
	default:
		return "/* unknown type */"
	}
}

func convertTypeExprFull(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		if rt, ok := typeMap[t.Name]; ok {
			return rt
		}
		return t.Name
	case *ast.ArrayType:
		elt := convertTypeExprFull(t.Elt)
		if t.Len != nil {
			// Fixed-size array
			if lit, ok := t.Len.(*ast.BasicLit); ok {
				return fmt.Sprintf("[%s; %s]", elt, lit.Value)
			}
			return fmt.Sprintf("[%s; %s]", elt, convertExprStr(token.NewFileSet(), t.Len, nil))
		}
		// Slice
		return fmt.Sprintf("Vec<%s>", elt)
	case *ast.StarExpr:
		if isBigIntType(t.X) {
			return "BigUint"
		}
		inner := convertTypeExprFull(t.X)
		if inner == "Int" {
			return "Int"
		}
		return fmt.Sprintf("&mut %s", inner)
	case *ast.MapType:
		key := convertTypeExprFull(t.Key)
		val := convertTypeExprFull(t.Value)
		return fmt.Sprintf("HashMap<%s, %s>", key, val)
	case *ast.FuncType:
		return convertFuncType(t)
	case *ast.SelectorExpr:
		if pkg, ok := t.X.(*ast.Ident); ok {
			if pkg.Name == "big" && t.Sel.Name == "Int" {
				return "BigUint"
			}
		}
		return convertTypeExprFull(t.Sel)
	default:
		return "/* unknown type */"
	}
}

func convertFuncType(t *ast.FuncType) string {
	var params []string
	if t.Params != nil {
		for _, field := range t.Params.List {
			pt := convertTypeExprFull(field.Type)
			count := len(field.Names)
			if count == 0 {
				count = 1
			}
			for i := 0; i < count; i++ {
				params = append(params, pt)
			}
		}
	}
	ret := convertFuncResults(t.Results)
	if ret != "" {
		return fmt.Sprintf("fn(%s) -> %s", strings.Join(params, ", "), ret)
	}
	return fmt.Sprintf("fn(%s)", strings.Join(params, ", "))
}

// -------------------------------------------------------
// Helpers
// -------------------------------------------------------

func extractSource(fset *token.FileSet, node ast.Node, src []byte) string {
	start := fset.Position(node.Pos()).Offset
	end := fset.Position(node.End()).Offset
	if start < 0 || end < 0 || start >= len(src) || end > len(src) {
		return "?"
	}
	return string(src[start:end])
}

// convertExpr for GenDecl context (no src needed, uses ast walking)
func convertExpr(fset *token.FileSet, expr ast.Expr, filePath string) string {
	src, err := os.ReadFile(filePath)
	if err != nil {
		fatal("read %s: %v", filePath, err)
	}
	return convertExprStr(fset, expr, src)
}

func fatal(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "FATAL: "+format+"\n", args...)
	os.Exit(1)
}

func rejectGeneratedFallbacks(base, code string) error {
	for _, marker := range []string{
		"todo!(",
		"TODO:",
		"// empty body",
		"// error reading source",
		"/* closure TODO */",
		"/* unknown type */",
		"// goto ",
		"// label: ",
	} {
		if strings.Contains(code, marker) {
			return fmt.Errorf("go2rs generated unsupported fallback %q in %s", marker, base)
		}
	}
	return nil
}

// Post-process Rust source: fix common patterns
func postProcess(code string) string {
	// Fix ^uint64(0) → !0u64
	code = strings.ReplaceAll(code, "(!((0 as u64)))", "(!0u64)")
	code = strings.ReplaceAll(code, "(!((0 as u32)))", "(!0u32)")

	// Fix array index: [N as usize] → [N] for numeric constants
	reConstIdx := regexp.MustCompile(`\[(\d+) as usize\]`)
	code = reConstIdx.ReplaceAllString(code, "[$1]")

	// Remove redundant double parens in conditions: ((...)) → (...)
	// But be careful not to break tuple returns
	reDoubleParen := regexp.MustCompile(`\(\(([^()]+)\)\)`)
	for i := 0; i < 5; i++ { // multiple passes for nested
		prev := code
		code = reDoubleParen.ReplaceAllString(code, "($1)")
		if code == prev {
			break
		}
	}

	// Fix untyped integer variables that use .wrapping_add()
	// Rust needs explicit type annotation for integer literals calling methods
	reUntypedJ := regexp.MustCompile(`let mut (j|k_lcv) = (\d+);`)
	code = reUntypedJ.ReplaceAllString(code, "let mut $1: i64 = $2;")

	// Go switch default: break means "do nothing"; in Rust match it must not
	// become a loop break.
	reDefaultBreak := regexp.MustCompile(`(?m)^(\s*)_ => \{\n\s*break;\n\s*\}`)
	code = reDefaultBreak.ReplaceAllString(code, "${1}_ => {}")

	code = rewriteRoundC2Loop(code)
	code = rewriteFmaDoneLabelBlock(code)
	code = rewriteAdd128MatchBreak(code)
	code = rewriteByteLiteralSubtractions(code)
	code = strings.ReplaceAll(code, "    let mut -6176: i64 = (-6176);\n", "")
	code = strings.ReplaceAll(code, "    let mut -6176: i32 = (-6176);\n", "")

	return code
}

func rewriteByteLiteralSubtractions(code string) string {
	rePsAt := regexp.MustCompile(`ps_at!\(([^)]*)\) - b'([^']*)'`)
	code = rePsAt.ReplaceAllString(code, `ps_at!($1).wrapping_sub(b'$2')`)

	reSimple := regexp.MustCompile(`([A-Za-z_][A-Za-z0-9_]*(?:\[[^\]\n]+\])?) - b'([^']*)'`)
	return reSimple.ReplaceAllString(code, `$1.wrapping_sub(b'$2')`)
}

func optimizeRustStringHotpaths(base string, code string) (string, error) {
	switch base {
	case "bid32_string.go":
		return optimizeBid32StringParse(code)
	default:
		return code, nil
	}
}

func optimizeBid32StringParse(code string) (string, error) {
	if strings.Contains(code, `let s = (ps).trim_start_matches(|c| " \t".contains(c)).as_bytes();`) &&
		strings.Contains(code, `s.eq_ignore_ascii_case(b"inf")`) {
		return code, nil
	}

	headerRe := regexp.MustCompile(`(?m)^    let mut s = \(ps\)\.trim_start_matches\(\|c\| " \\t"\.contains\(c\)\)\.to_string\(\);\n    if \(\(s\.len\(\) as i64\) == 0\) \{\n        return \(0x7c000000, 0\);\n    \}\n    let mut c = s\[0\];\n`)
	if headerRe.MatchString(code) {
		code = headerRe.ReplaceAllLiteralString(code, `    let s = (ps).trim_start_matches(|c| " \t".contains(c)).as_bytes();
    if ((s.len() as i64) == 0) {
        return (0x7c000000, 0);
    }
    let mut c = s[0];
`)
	} else {
		old := `    let s = (ps).trim_start_matches(|c| " \t".contains(c)).as_bytes().to_vec();
`
		if !strings.Contains(code, old) {
			return "", fmt.Errorf("optimize bid32 string parse: expected parser header not found")
		}
		code = strings.Replace(code, old, `    let s = (ps).trim_start_matches(|c| " \t".contains(c)).as_bytes();
`, 1)
	}

	lineReplacements := [][2]string{
		{`    let mut sl = (s).to_ascii_lowercase();
`, ``},
		{`    let mut sl = String::from_utf8_lossy(&s).to_ascii_lowercase();
`, ``},
		{`        if ((sl == "inf") || (sl == "infinity")) {`, `        if (s.eq_ignore_ascii_case(b"inf") || s.eq_ignore_ascii_case(b"infinity")) {`},
		{`        if (sl).starts_with("snan") {`, `        if ((s.len() >= 4) && s[..4].eq_ignore_ascii_case(b"snan")) {`},
		{`        let mut sl1 = (&mut s[1 as usize..]).to_ascii_lowercase();
`, `        let sl1 = &s[1 as usize..];
`},
		{`        let mut sl1 = String::from_utf8_lossy(&s[1 as usize..]).to_ascii_lowercase();
`, `        let sl1 = &s[1 as usize..];
`},
		{`        if ((sl1 == "inf") || (sl1 == "infinity")) {`, `        if (sl1.eq_ignore_ascii_case(b"inf") || sl1.eq_ignore_ascii_case(b"infinity")) {`},
		{`        if (sl1).starts_with("snan") {`, `        if ((sl1.len() >= 4) && sl1[..4].eq_ignore_ascii_case(b"snan")) {`},
		{`        if (sl1 == "nan") {`, `        if (sl1.eq_ignore_ascii_case(b"nan")) {`},
	}
	for _, repl := range lineReplacements {
		code = strings.ReplaceAll(code, repl[0], repl[1])
	}
	for _, forbidden := range []string{"to_ascii_lowercase", "String::from_utf8_lossy", "as_bytes().to_vec()"} {
		if strings.Contains(code, forbidden) {
			return "", fmt.Errorf("optimize bid32 string parse: leftover %q", forbidden)
		}
	}
	return code, nil
}

func rewriteRoundC2Loop(code string) string {
	label := "                    // label: roundC2\n"
	gotoStmt := "                                // goto roundC2; // TODO: convert goto to loop/break\n"
	footer := "                    C1_hi = C1.w[1];\n"

	labelPos := strings.Index(code, label)
	if labelPos < 0 {
		return code
	}
	gotoSearchStart := labelPos + len(label)
	gotoRel := strings.Index(code[gotoSearchStart:], gotoStmt)
	if gotoRel < 0 {
		return code
	}
	gotoPos := gotoSearchStart + gotoRel
	bodyAfterGoto := gotoPos + len(gotoStmt)
	footerRel := strings.Index(code[bodyAfterGoto:], footer)
	if footerRel < 0 {
		return code
	}
	footerPos := bodyAfterGoto + footerRel

	var sb strings.Builder
	sb.WriteString(code[:labelPos])
	sb.WriteString("                    'roundC2: loop {\n")
	sb.WriteString(code[gotoSearchStart:gotoPos])
	sb.WriteString("                                continue 'roundC2;\n")
	sb.WriteString(code[bodyAfterGoto:footerPos])
	sb.WriteString("                    break;\n")
	sb.WriteString("                    }\n")
	sb.WriteString(code[footerPos:])
	return sb.String()
}

func rewriteFmaDoneLabelBlock(code string) string {
	funcMarker := "pub(crate) fn bid_fma_delta_ge_zero"
	label := "    // label: done\n"
	gotoStmt := "// goto done; // TODO: convert goto to loop/break"

	funcPos := strings.Index(code, funcMarker)
	if funcPos < 0 {
		return code
	}
	labelRel := strings.Index(code[funcPos:], label)
	if labelRel < 0 {
		return code
	}
	labelPos := funcPos + labelRel
	entryRel := -1
	for _, entry := range []string{
		"    if ((p34 <= (delta.wrapping_sub(1)))",
		"    if ((p34 <= (delta - 1))",
	} {
		entryRel = strings.Index(code[funcPos:labelPos], entry)
		if entryRel >= 0 {
			break
		}
	}
	if entryRel < 0 {
		return code
	}
	entryPos := funcPos + entryRel

	body := code[entryPos:labelPos]
	if !strings.Contains(body, gotoStmt) {
		return code
	}
	body = strings.ReplaceAll(body, gotoStmt, "break 'done;")

	var sb strings.Builder
	sb.WriteString(code[:entryPos])
	sb.WriteString("    'done: {\n")
	sb.WriteString(body)
	sb.WriteString("    }\n")
	sb.WriteString(code[labelPos+len(label):])
	return sb.String()
}

func rewriteAdd128MatchBreak(code string) string {
	oldBlocks := []string{
		`        2 => {
            if ((FS.w[1] as i64) < 0) {
                break;
            }
            T2 = bid_power10_table_128[(diff_dec_expon.wrapping_add(extra_digits)) as usize];
            if __unsigned_compare_gt_128(FS, T2) {
                CYh = CYh.wrapping_add(2);
                FS = __sub_128_128(FS, T2);
            } else if ((FS.w[1] == T2.w[1]) && (FS.w[0] == T2.w[0])) {
                CYh = CYh.wrapping_add(1);
                FS.w[1] = 0;
                FS.w[0] = 0;
            } else if ((FS.w[1] | FS.w[0]) != 0) {
                CYh = CYh.wrapping_add(1);
            }
        }
`,
		`        2 => {
            if ((FS.w[1] as i64) < 0) {
                break;
            }
            T2 = bid_power10_table_128[(diff_dec_expon + extra_digits) as usize];
            if __unsigned_compare_gt_128(FS, T2) {
                CYh = CYh.wrapping_add(2);
                FS = __sub_128_128(FS, T2);
            } else if ((FS.w[1] == T2.w[1]) && (FS.w[0] == T2.w[0])) {
                CYh = CYh.wrapping_add(1);
                FS.w[1] = 0;
                FS.w[0] = 0;
            } else if ((FS.w[1] | FS.w[0]) != 0) {
                CYh = CYh.wrapping_add(1);
            }
        }
`,
		`        2 => {
            if ((FS.w[1] as i64) < 0) {
                break;
            }
            T2 = bid_power10_table_128[(diff_dec_expon + extra_digits) as usize];
            if __unsigned_compare_gt_128(FS, T2) {
                CYh += 2;
                FS = __sub_128_128(FS, T2);
            } else if ((FS.w[1] == T2.w[1]) && (FS.w[0] == T2.w[0])) {
                CYh = CYh.wrapping_add(1);
                FS.w[1] = 0;
                FS.w[0] = 0;
            } else if ((FS.w[1] | FS.w[0]) != 0) {
                CYh = CYh.wrapping_add(1);
            }
        }
`,
	}
	newBlock := `        2 => {
            if !((FS.w[1] as i64) < 0) {
                T2 = bid_power10_table_128[(diff_dec_expon.wrapping_add(extra_digits)) as usize];
                if __unsigned_compare_gt_128(FS, T2) {
                    CYh = CYh.wrapping_add(2);
                    FS = __sub_128_128(FS, T2);
                } else if ((FS.w[1] == T2.w[1]) && (FS.w[0] == T2.w[0])) {
                    CYh = CYh.wrapping_add(1);
                    FS.w[1] = 0;
                    FS.w[0] = 0;
                } else if ((FS.w[1] | FS.w[0]) != 0) {
                    CYh = CYh.wrapping_add(1);
                }
            }
        }
`
	for _, oldBlock := range oldBlocks {
		if strings.Contains(code, oldBlock) {
			return strings.Replace(code, oldBlock, newBlock, 1)
		}
	}
	return code
}
