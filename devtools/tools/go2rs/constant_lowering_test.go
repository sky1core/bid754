package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestIntegerConstantExpressionsExecuteBeforeRustCast(t *testing.T) {
	code := convertTypeCheckedTestFile(t, "constant_probe.go", `package bidgo
func mantissa64() uint64 { return uint64((1 << 52) - 1) }
func maximumFinite64() uint64 { return (uint64(2046) << 52) + uint64((1 << 52) - 1) }
func typedMask() uint64 { return (uint64(1) << 63) | ((uint64(1) << 52) - 1) }
`)
	source := `#![allow(arithmetic_overflow, overflowing_literals, unused_imports, unused_parens)]
mod generated { pub mod prelude {} pub mod probe {
` + code + `
}}
fn main() {
    assert_eq!(generated::probe::mantissa64(), 0x000f_ffff_ffff_ffff);
    assert_eq!(generated::probe::maximum_finite64(), 0x7fef_ffff_ffff_ffff);
    assert_eq!(generated::probe::typed_mask(), 0x800f_ffff_ffff_ffff);
}
`
	dir := t.TempDir()
	path := filepath.Join(dir, "probe.rs")
	binary := filepath.Join(dir, "probe")
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("rustc", "--edition=2021", path, "-o", binary).CombinedOutput(); err != nil {
		t.Fatalf("compile generated constants: %v\n%s\n%s", err, out, source)
	}
	if out, err := exec.Command(binary).CombinedOutput(); err != nil {
		t.Fatalf("execute generated constants: %v\n%s", err, out)
	}
}
