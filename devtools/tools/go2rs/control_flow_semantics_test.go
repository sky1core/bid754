package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestParallelAssignmentAndNestedContinueExecuteInRust(t *testing.T) {
	code := convertTypeCheckedTestFile(t, "semantics.go", `package bidgo
func swap(x, y uint64) (uint64, uint64) { x, y = y, x; return x, y }
func repeatedTarget(x uint64) uint64 { x, x = x + 1, x + 2; return x }
func mixedDefine(x uint64) uint64 { y := x; y, z := y + 1, y + 2; return y*10 + z }
func pair() (uint64, uint64) { return 4, 5 }
func pointerTuple(p *uint64) uint64 { var q uint64; *p, q = pair(); return q }
func nestedContinue() int {
 n := 0
 for i := 0; i < 3; i++ {
  for j := 0; j < 2; j++ { if j == 0 { continue }; n++ }
  if i == 1 { continue }
  n += 10
 }
 return n
}
`)
	source := `#![allow(unused_imports, unused_parens, unused_mut, unused_assignments)]
mod generated { pub mod prelude {} pub mod probe {
` + code + `
}}
fn main() {
 assert_eq!(generated::probe::swap(1, 2), (2, 1));
 assert_eq!(generated::probe::repeated_target(1), 3);
 assert_eq!(generated::probe::mixed_define(1), 23);
 assert_eq!(generated::probe::nested_continue(), 23);
 let mut value = 1;
 assert_eq!(generated::probe::pointer_tuple(&mut value), 5);
 assert_eq!(value, 4);
}
`
	dir := t.TempDir()
	path := filepath.Join(dir, "probe.rs")
	binary := filepath.Join(dir, "probe")
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("rustc", "--edition=2021", path, "-o", binary).CombinedOutput(); err != nil {
		t.Fatalf("compile generated Rust: %v\n%s\n%s", err, out, source)
	}
	if out, err := exec.Command(binary).CombinedOutput(); err != nil {
		t.Fatalf("execute generated Rust: %v\n%s\n%s", err, out, source)
	}
}

func TestParallelAssignmentRejectsIndexedTarget(t *testing.T) {
	for _, source := range []string{
		`package bidgo
func indexed(a *[2]uint64, i int, v uint64) { a[i], v = v, a[i] }`,
		`package bidgo
func pair() (uint64, uint64) { return 1, 2 }
func indexed(a *[2]uint64, i int) { a[i], a[i+1] = pair() }`,
		`package bidgo
func rebound(p, q *uint64) { p, *p = q, 7 }`,
		`package bidgo
func change(p **uint64, q *uint64) uint64 { *p = q; return 7 }
func rebound(p, q *uint64) { *p, *q = change(&p, q), 8 }`,
	} {
		t.Run(source, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "unsupported.go")
			if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
				t.Fatal(err)
			}
			reg := &Registry{Types: map[string]TypeDef{}, Constants: map[string]ConstDef{}, Tables: map[string]TableDef{}, Functions: map[string]FuncDef{}}
			activeRegistry = reg
			fset, parsed, info := parseTypeCheckedPackage(dir, []string{path})
			previous := activeTypeInfo
			activeTypeInfo = info
			t.Cleanup(func() { activeTypeInfo = previous })
			_, err := convertParsedFile(fset, parsed[path], path, reg)
			if err == nil || !strings.Contains(err.Error(), "parallel assignment") {
				t.Fatalf("expected explicit parallel assignment error, got %v", err)
			}
		})
	}
}

func TestLabeledLoopBranchesAreRejected(t *testing.T) {
	for _, branch := range []string{"continue", "break"} {
		source := "package bidgo\nfunc f() { outer: for i:=0;i<2;i++ { for j:=0;j<2;j++ { " + branch + " outer } } }"
		dir := t.TempDir()
		path := filepath.Join(dir, "labels.go")
		if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
			t.Fatal(err)
		}
		reg := &Registry{Types: map[string]TypeDef{}, Constants: map[string]ConstDef{}, Tables: map[string]TableDef{}, Functions: map[string]FuncDef{}}
		activeRegistry = reg
		fset, parsed, info := parseTypeCheckedPackage(dir, []string{path})
		previous := activeTypeInfo
		activeTypeInfo = info
		_, err := convertParsedFile(fset, parsed[path], path, reg)
		activeTypeInfo = previous
		if err == nil || !strings.Contains(err.Error(), "labeled loop branches") {
			t.Fatalf("%s label was silently discarded: %v", branch, err)
		}
	}
}
