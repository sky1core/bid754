package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestProductionParserInputLowering(t *testing.T) {
	srcDir := filepath.Join("..", "..", "..", "bid754-go", "internal", "bidgo")
	reg := loadRegistry(filepath.Join("..", "registry", "symbols.json"))
	oldRegistry, oldFunctions, oldTypeInfo := activeRegistry, activeSourceFunctions, activeTypeInfo
	t.Cleanup(func() {
		activeRegistry, activeSourceFunctions, activeTypeInfo = oldRegistry, oldFunctions, oldTypeInfo
	})
	activeRegistry = reg
	files, err := filepath.Glob(filepath.Join(srcDir, "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	var targets []string
	for _, path := range files {
		if shouldConvertFile(filepath.Base(path)) {
			targets = append(targets, path)
		}
	}
	activeSourceFunctions, err = collectSourceFunctionNames(targets)
	if err != nil {
		t.Fatal(err)
	}
	fset, parsed, info := parseTypeCheckedPackage(srcDir, targets)
	activeTypeInfo = info
	for _, name := range []string{"bid64_from_string.go", "bid128_string.go"} {
		path := filepath.Join(srcDir, name)
		code, err := convertParsedFile(fset, parsed[path], path, reg)
		if err != nil {
			t.Fatal(err)
		}
		code = postProcess(code)
		if err := rejectGeneratedFallbacks(name, code); err != nil {
			t.Fatal(err)
		}
		for _, want := range []string{"let str = str.as_ref();", "let mut ps: i64 = 0;", "nul_terminated_byte_at(str, ps)", "ps = ps.wrapping_add(1);"} {
			if !strings.Contains(code, want) {
				t.Fatalf("%s missing cursor lowering %q:\n%s", name, want, code)
			}
		}
	}
}

func TestParserNULTerminatedByteLowering(t *testing.T) {
	path := filepath.Join("..", "..", "..", "bid754-go", "internal", "bidgo", "bid64_from_string.go")
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, src, 0)
	if err != nil {
		t.Fatal(err)
	}
	var helper string
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == "nulTerminatedByteAt" {
			helper = string(src[fset.Position(fn.Pos()).Offset:fset.Position(fn.End()).Offset])
		}
	}
	if helper == "" {
		t.Fatal("production parser input helper missing")
	}
	code := convertTypeCheckedTestFile(t, "parser_input.go", "package bidgo\n"+helper)
	want := `pub(crate) fn nul_terminated_byte_at(s: impl AsRef<str>, mut index: i64) -> u8 {
    let s = s.as_ref();
    if ((index < 0) || (index >= (s.len() as i64))) {
        return 0;
    }
    return s.as_bytes()[index as usize];
}`
	if !strings.Contains(code, want) {
		t.Fatalf("unexpected input helper lowering:\n%s", code)
	}
	source := `#![allow(unused_imports, unused_parens, unused_mut)]
mod generated { pub mod prelude {} pub mod probe {
` + code + `
}}
fn main() {
    use generated::probe::nul_terminated_byte_at;
    for s in ["", "+", "-Inf", "sNaN", "1e+", "1\0tail", "é", "1é"] {
        for index in -1..s.len() as i64 + 10 {
            let expected = if index < 0 { 0 } else { *s.as_bytes().get(index as usize).unwrap_or(&0) };
            assert_eq!(nul_terminated_byte_at(s, index), expected);
        }
        assert_eq!(nul_terminated_byte_at(s, i64::MAX), 0);
        assert_eq!(nul_terminated_byte_at(s, i64::MIN), 0);
    }
}
`
	dir := t.TempDir()
	rustPath := filepath.Join(dir, "probe.rs")
	binary := filepath.Join(dir, "probe")
	if err := os.WriteFile(rustPath, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("rustc", "--edition=2021", "-C", "overflow-checks=yes", rustPath, "-o", binary).CombinedOutput(); err != nil {
		t.Fatalf("compile generated parser input helper: %v\n%s", err, out)
	}
	if out, err := exec.Command(binary).CombinedOutput(); err != nil {
		t.Fatalf("execute generated parser input helper: %v\n%s", err, out)
	}
}
