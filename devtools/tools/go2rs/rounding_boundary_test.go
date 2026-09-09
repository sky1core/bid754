package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/sky1core/bid754/devtools/internal/testgen"
)

func withRoundingContracts(t *testing.T, reg map[string]roundingContract) {
	t.Helper()
	old := roundingContracts
	roundingContracts = reg
	t.Cleanup(func() { roundingContracts = old })
}

func TestCheckedRoundingSourcesPreserveModulesAndRejectUnknownShapes(t *testing.T) {
	t.Run("preserve modules and route internal calls to the port", func(t *testing.T) {
		withRoundingContracts(t, map[string]roundingContract{
			"bid128_add": {params: []string{"u64", "i64"}, modes: []int{1}, maxMode: 4},
		})
		sources := map[string]string{
			"mod.rs":        "pub mod bid128_add;\n",
			"prelude.rs":    "pub use super::bid128_add::*;\n",
			"bid128_add.rs": "pub fn bid128_add(x: u64, rnd_mode: i64) -> (u64, u32) {\n    (x, 0)\n}\n",
			"caller.rs":     "// bid128_add\nconst LABEL: &str = \"bid128_add\";\npub fn caller(x: u64) -> (u64, u32) {\n    super::bid128_add::bid128_add(x, 0)\n}\n",
		}
		got, err := checkedRoundingSources(sources)
		if err != nil {
			t.Fatal(err)
		}
		if got["mod.rs"] != sources["mod.rs"] || got["prelude.rs"] != sources["prelude.rs"] {
			t.Fatal("function lowering changed module paths")
		}
		if !strings.Contains(got["caller.rs"], "super::bid128_add::bid128_add_port(x, 0)") {
			t.Fatal("internal call did not reach the port")
		}
		if !strings.Contains(got["caller.rs"], "// bid128_add\nconst LABEL: &str = \"bid128_add\";") {
			t.Fatal("function lowering changed literal or comment contents")
		}
	})

	t.Run("renamed mode parameter keeps its guard", func(t *testing.T) {
		withRoundingContracts(t, map[string]roundingContract{
			"bid128_add": {params: []string{"u64", "i64"}, modes: []int{1}, maxMode: 4},
		})
		got, err := checkedRoundingSources(map[string]string{
			"bid128_add.rs": "pub fn bid128_add(x: u64, mode: i64) -> (u64, u32) {\n    (x, 0)\n}\n",
		})
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(got["bid128_add.rs"], "if !(0..=4).contains(&mode) {") {
			t.Fatalf("renamed mode parameter dropped its guard:\n%s", got["bid128_add.rs"])
		}
		if !strings.Contains(got["bid128_add.rs"], "pub(crate) fn bid128_add_port(") {
			t.Fatalf("port was not made crate-private:\n%s", got["bid128_add.rs"])
		}
	})

	t.Run("reject a changed parameter type", func(t *testing.T) {
		withRoundingContracts(t, map[string]roundingContract{
			"bid128_add": {params: []string{"u64", "i64"}, modes: []int{1}, maxMode: 4},
		})
		if _, err := checkedRoundingSources(map[string]string{
			"bid128_add.rs": "pub fn bid128_add(x: u32, rnd_mode: i64) -> (u32, u32) {\n    (x, 0)\n}\n",
		}); err == nil {
			t.Fatal("accepted a parameter type that diverges from the contract")
		}
	})

	t.Run("reject a stale contract entry", func(t *testing.T) {
		withRoundingContracts(t, map[string]roundingContract{
			"bid128_add": {params: []string{"u64", "i64"}, modes: []int{1}, maxMode: 4},
			"ghost":      {params: []string{"u64", "i64"}, modes: []int{1}, maxMode: 4},
		})
		if _, err := checkedRoundingSources(map[string]string{
			"bid128_add.rs": "pub fn bid128_add(x: u64, rnd_mode: i64) -> (u64, u32) {\n    (x, 0)\n}\n",
		}); err == nil {
			t.Fatal("accepted a contract entry with no generated function")
		}
	})

	t.Run("fail closed on an unregistered rounding parameter and bad control words", func(t *testing.T) {
		withRoundingContracts(t, map[string]roundingContract{})
		for _, signature := range []string{
			"pub fn f(rnd_mode: i64) -> (bool, u32) {\n(false, 0)\n}\n",
			"pub fn f(rnd_mode: u8) -> u32 {\n0\n}\n",
			"pub fn bid_get_decimal_rounding_direction(rnd_mode: u64) -> u32 {\n0\n}\n",
			"pub fn bid_set_decimal_rounding_direction(rounding_mode: u32, rnd_mode: u32) -> u64 {\n0\n}\n",
		} {
			if _, err := checkedRoundingSources(map[string]string{"bad.rs": signature}); err == nil {
				t.Fatalf("accepted unclassified boundary %s", signature)
			}
		}
	})
}

func TestBinary128ExponentTableArithmeticWidensBeforeSubtract(t *testing.T) {
	src := convertTypeCheckedTestFile(t, "binary128_probe.go", `package bidgo
var bid_exponents_binary128 = [1]int{0}
func exponent(k int) int { return bid_exponents_binary128[0] - k }
`)
	if !strings.Contains(src, "(bid_exponents_binary128[0 as usize] as i64).wrapping_sub(k)") {
		t.Fatalf("missing widening before binary128 exponent arithmetic:\n%s", src)
	}
}

func TestConstantShiftIsEvaluatedBeforeRustContextCast(t *testing.T) {
	src := convertTypeCheckedTestFile(t, "shift_probe.go", `package bidgo
func coefficient(x uint64) uint64 { return (1 << 53) + x }
`)
	if strings.Contains(src, "1 << 53") || !strings.Contains(src, "9007199254740992 as u64") {
		t.Fatalf("untyped Rust shift can overflow before its outer cast:\n%s", src)
	}
}

func copyRoundingProbeTree(t *testing.T, source, dest string) {
	t.Helper()
	err := filepath.WalkDir(source, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dest, rel)
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
}

func newRoundingProbeCrateAt(t *testing.T, repo string) (string, func(args ...string) ([]byte, error)) {
	t.Helper()
	root := t.TempDir()
	for _, rel := range []string{"bid754-go", "bid754-rs/src", "bid754-rs/benches", "devtools/tools/registry"} {
		copyRoundingProbeTree(t, filepath.Join(repo, rel), filepath.Join(root, rel))
	}
	for _, rel := range []string{"devtools/go.mod", "devtools/generated/testspec/public_api_routing_inventory.json", "bid754-rs/Cargo.toml", "bid754-rs/Cargo.lock"} {
		dest := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			t.Fatal(err)
		}
		copyRoundingProbeTree(t, filepath.Join(repo, rel), dest)
	}
	oldRegistry, oldFunctions, oldTypes := activeRegistry, activeSourceFunctions, activeTypeInfo
	t.Cleanup(func() { activeRegistry, activeSourceFunctions, activeTypeInfo = oldRegistry, oldFunctions, oldTypes })
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(filepath.Join(root, "devtools")); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(oldDir); err != nil {
			t.Error(err)
		}
	})
	manifest := filepath.Join(root, "bid754-rs/Cargo.toml")
	target := filepath.Join(root, "target")
	run := func(args ...string) ([]byte, error) {
		cmd := exec.Command("cargo", append(args, "--locked", "--manifest-path", manifest, "--target-dir", target)...)
		return cmd.CombinedOutput()
	}
	return root, run
}

func TestRoundingBoundaryExternalCrate(t *testing.T) {
	repo := filepath.Dir(findProjectRoot())
	root, run := newRoundingProbeCrateAt(t, repo)
	main()

	goExported := make(map[string]bool)
	goRoundingParam := make(map[string]bool)
	goPaths, err := filepath.Glob(filepath.Join(root, "bid754-go/internal/bidgo/*.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range goPaths {
		if !shouldConvertFile(filepath.Base(path)) {
			continue
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || !ast.IsExported(fn.Name.Name) || fn.Type.Params == nil {
				continue
			}
			if fn.Name.Name == "BidGetDecimalRoundingDirection" || fn.Name.Name == "BidSetDecimalRoundingDirection" {
				continue
			}
			rustName := goFuncNameToRust(fn.Name.Name)
			goExported[rustName] = true
			for _, field := range fn.Type.Params.List {
				for _, name := range field.Names {
					switch name.Name {
					case "rndMode", "rnd_mode", "rounding_mode":
						goRoundingParam[rustName] = true
					}
				}
			}
		}
	}
	for name := range roundingContracts {
		if !goExported[name] {
			t.Fatalf("contract entry %s has no exported Go predecessor", name)
		}
	}
	for name := range goRoundingParam {
		if _, ok := roundingContracts[name]; !ok {
			t.Fatalf("exported Go rounding function %s is missing from the contract", name)
		}
	}

	var probe strings.Builder
	probe.WriteString("#![allow(unused_variables, unused_mut)]\nuse bid754::gen_types::BID_UINT128;\n#[test]\nfn reject_all_raw_rounding_boundaries() {\n")
	paths, err := filepath.Glob(filepath.Join(root, "bid754-rs/src/generated/*.rs"))
	if err != nil {
		t.Fatal(err)
	}
	boundaries := 0
	seen := make(map[string]bool)
	var privateCalls []string
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		bs, err := roundingBoundaries(string(data))
		if err != nil {
			t.Fatal(err)
		}
		module := strings.TrimSuffix(filepath.Base(path), ".rs")
		for _, b := range bs {
			seen[b.name] = true
			boundaries++
			params := strings.Split(b.params, ", ")
			for _, invalidMode := range b.modes {
				var invalids string
				var args []string
				for _, param := range params {
					p := strings.SplitN(strings.TrimPrefix(param, "mut "), ": ", 2)
					if p[0] == invalidMode {
						args = append(args, "invalid")
						if p[1] == "u32" {
							invalids = "[5u32, 6, 99, i32::MAX as u32, u32::MAX]"
						} else {
							invalids = fmt.Sprintf("[%di64, 99, -1, i64::MIN, i64::MAX]", b.maxMode+1)
						}
						continue
					}
					switch p[1] {
					case "u32", "u64", "i32", "i64":
						args = append(args, "0")
					case "BID_UINT128":
						args = append(args, "BID_UINT128 { lo: 0, hi: 0 }")
					case "impl AsRef<str>":
						args = append(args, `"1"`)
					case "&mut u32":
						args = append(args, "&mut flags")
					default:
						t.Fatalf("unhandled production boundary parameter %s in %s", param, b.name)
					}
				}
				call := fmt.Sprintf("bid754::generated::%s::%s(%s)", module, b.name, strings.Join(args, ", "))
				fmt.Fprintf(&probe, "for invalid in %s { let mut flags = 0x20u32; let got = %s;\n", invalids, call)
				if strings.HasPrefix(b.result, "Result<") {
					probe.WriteString("assert!(got.is_err());\n")
				} else {
					result, value := b.result, "got"
					if b.flags == "" {
						probe.WriteString("assert_eq!(got.1, 0x01);\n")
						result = strings.TrimSuffix(strings.TrimPrefix(result, "("), ", u32)")
						value = "got.0"
					} else {
						probe.WriteString("assert_eq!(flags, 0x21);\n")
					}
					want, err := b.invalidValue(result)
					if err != nil {
						t.Fatal(err)
					}
					switch result {
					case "BID_UINT128":
						fmt.Fprintf(&probe, "let want = %s; assert_eq!((%s.lo, %s.hi), (want.lo, want.hi));\n", want, value, value)
					case "f32", "f64":
						fmt.Fprintf(&probe, "assert!(%s.is_nan());\n", value)
					default:
						fmt.Fprintf(&probe, "assert_eq!(%s, %s);\n", value, want)
					}
				}
				probe.WriteString("}\n")
				privateCalls = append(privateCalls, "{ let invalid = 99; let mut flags = 0u32; "+strings.Replace(call, b.name+"(", b.name+"_port(", 1)+"; }")
			}
		}
	}
	for name := range roundingContracts {
		if !seen[name] {
			t.Fatalf("contract entry %s produced no generated boundary", name)
		}
	}
	if boundaries != len(roundingContracts) || boundaries == 0 {
		t.Fatalf("generated %d boundaries, contract pins %d", boundaries, len(roundingContracts))
	}
	probe.WriteString("}\n" + validRoundingProbe)
	testDir := filepath.Join(root, "bid754-rs/tests")
	if err := os.MkdirAll(testDir, 0o755); err != nil {
		t.Fatal(err)
	}
	probePath := filepath.Join(testDir, "rounding_boundary.rs")
	if err := os.WriteFile(probePath, []byte(probe.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	if output, err := run("test", "--test", "rounding_boundary"); err != nil {
		t.Fatalf("external production boundary tests: %v\n%s", err, output)
	} else {
		t.Logf("%d checked production entrypoints; %s", boundaries, output)
	}
	if output, err := run("check", "--benches"); err != nil {
		t.Fatalf("production benchmark consumer check: %v\n%s", err, output)
	}
	header := "devtools/third_party/intel_dfp/LIBRARY/src/bid_functions.h"
	if _, err := os.Stat(filepath.Join(repo, header)); err == nil {
		if err := os.MkdirAll(filepath.Dir(filepath.Join(root, header)), 0o755); err != nil {
			t.Fatal(err)
		}
		copyRoundingProbeTree(t, filepath.Join(repo, header), filepath.Join(root, header))
		outputs, err := testgen.GenerateRustPublicParityOutputs(filepath.Join(root, "devtools"))
		if err != nil {
			t.Fatal(err)
		}
		for rel, data := range outputs {
			if err := os.WriteFile(filepath.Join(root, "devtools", rel), data, 0o644); err != nil {
				t.Fatal(err)
			}
		}
		if output, err := run("test", "--test", "public_parity_generated"); err != nil {
			t.Fatalf("generated public parity consumer: %v\n%s", err, output)
		} else {
			t.Logf("generated public parity consumer: %s", output)
		}
	} else if os.IsNotExist(err) {
		t.Log("public parity consumer not run: pinned Intel header is absent")
	} else {
		t.Fatal(err)
	}
	privateProbe := "use bid754::gen_types::BID_UINT128;\nfn main() {\n" + strings.Join(privateCalls, "\n") + "\n}\n"
	if err := os.WriteFile(probePath, []byte(privateProbe), 0o644); err != nil {
		t.Fatal(err)
	}
	output, err := run("check", "--test", "rounding_boundary")
	if err == nil || strings.Count(string(output), "error[E0603]") != len(privateCalls) {
		t.Fatalf("expected %d inaccessible port calls; got %v\n%s", len(privateCalls), err, output)
	}
	t.Logf("%d external port access attempts rejected by Rust privacy", len(privateCalls))
	if err := os.WriteFile(probePath, []byte("fn main() { let _ = bid754::RoundingMode::NearestDown; }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	output, err = run("check", "--test", "rounding_boundary")
	if err == nil || !strings.Contains(string(output), "error[E0599]") {
		t.Fatalf("compatibility rounding leaked into the public enum: %v\n%s", err, output)
	}
}

func renameParamInFunc(t *testing.T, path, funcName, from, to string) {
	t.Helper()
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, src, 0)
	if err != nil {
		t.Fatal(err)
	}
	start, end := 0, 0
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if ok && fn.Name.Name == funcName {
			start = fset.Position(fn.Pos()).Offset
			end = fset.Position(fn.End()).Offset
		}
	}
	if start == end {
		t.Fatalf("function %s not found in %s", funcName, path)
	}
	block := regexp.MustCompile(`\b`+regexp.QuoteMeta(from)+`\b`).ReplaceAllString(string(src[start:end]), to)
	out := string(src[:start]) + block + string(src[end:])
	if out == string(src) {
		t.Fatalf("rename of %s produced no change in %s", from, funcName)
	}
	if err := os.WriteFile(path, []byte(out), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestRoundingBoundaryRenameCannotDropGuard(t *testing.T) {
	repo := filepath.Dir(findProjectRoot())
	root, run := newRoundingProbeCrateAt(t, repo)
	renameParamInFunc(t, filepath.Join(root, "bid754-go/internal/bidgo/lrint64.go"), "Bid64Llrint", "rndMode", "mode")
	main()

	generated, err := os.ReadFile(filepath.Join(root, "bid754-rs/src/generated/lrint64.rs"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(generated), "mode: i64") {
		t.Fatalf("rename did not reach the generated signature:\n%s", generated)
	}
	bs, err := roundingBoundaries(string(generated))
	if err != nil {
		t.Fatalf("rename left the generator unable to classify: %v", err)
	}
	var llrint *roundingBoundary
	for i := range bs {
		if bs[i].name == "bid64_llrint" {
			llrint = &bs[i]
		}
	}
	if llrint == nil {
		t.Fatal("bid64_llrint boundary vanished after the rename")
	}
	if len(llrint.modes) != 1 || llrint.modes[0] != "mode" {
		t.Fatalf("renamed parameter is not the guarded mode: %v", llrint.modes)
	}
	if !strings.Contains(string(generated), "if !(0..=4).contains(&mode) {") {
		t.Fatalf("guard missing after rename:\n%s", generated)
	}

	testDir := filepath.Join(root, "bid754-rs/tests")
	if err := os.MkdirAll(testDir, 0o755); err != nil {
		t.Fatal(err)
	}
	probePath := filepath.Join(testDir, "rounding_rename.rs")
	probe := `#[test]
fn renamed_mode_still_guards_bid64_llrint() {
    use bid754::generated::lrint64::bid64_llrint;
    assert_eq!(bid64_llrint(0, 99), (i64::MIN, 1));
    for invalid in [5i64, -1, i64::MIN, i64::MAX] {
        assert_eq!(bid64_llrint(0, invalid), (i64::MIN, 1));
    }
    for valid in 0..=4 {
        assert_ne!(bid64_llrint(0, valid), (i64::MIN, 1));
    }
}
`
	if err := os.WriteFile(probePath, []byte(probe), 0o644); err != nil {
		t.Fatal(err)
	}
	if output, err := run("test", "--test", "rounding_rename"); err != nil {
		t.Fatalf("renamed-mode guard runtime check: %v\n%s", err, output)
	} else {
		t.Logf("renamed-mode guard runtime check: %s", output)
	}
	if err := os.WriteFile(probePath, []byte("fn main() { let _ = bid754::generated::lrint64::bid64_llrint_port(0, 99); }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	output, err := run("check", "--test", "rounding_rename")
	if err == nil || !strings.Contains(string(output), "error[E0603]") {
		t.Fatalf("expected the renamed port to stay inaccessible; got %v\n%s", err, output)
	}
	t.Logf("renamed port stayed crate-private (E0603) after the rename")
}

const validRoundingProbe = `
#[test]
fn control_words_preserve_canonical_requests_without_bypassing_arithmetic() {
    use bid754::generated::flag_operations::{bid_get_decimal_rounding_direction, bid_set_decimal_rounding_direction};
    use bid754::generated::add64::bid64_add_with_flags;
    for current in [0u32, 1, 2, 3, 4, 5, 6, 8, 99, u32::MAX] {
        assert_eq!(bid_get_decimal_rounding_direction(current), current);
        for requested in [0u32, 1, 2, 3, 4] {
            assert_eq!(bid_set_decimal_rounding_direction(requested, current), requested);
        }
        for requested in [5u32, 6, 8, 99, u32::MAX] {
            assert_eq!(bid_set_decimal_rounding_direction(requested, current), current);
        }
        if current > 4 {
            let mode = bid_set_decimal_rounding_direction(u32::MAX, current);
            assert_eq!(bid64_add_with_flags(0, 0, i64::from(mode)), (0x7c00000000000000, 1));
        }
    }
}

#[test]
fn public_and_compatibility_rounding() {
    use bid754::{bid64_from_string_raw, RoundingMode};
    use bid754::generated::add64::{bid64_add, bid64_add_with_flags};
    use bid754::generated::bid32_string::bid32_from_string_raw;
    use bid754::generated::bid128_string::bid128_from_string;
    let (x, _) = bid64_from_string_raw("1", 0);
    let (y, _) = bid64_from_string_raw("0.00000000000000001", 0);
    let (up, _) = bid64_from_string_raw("1.000000000000001", 0);
    let (down, _) = bid64_from_string_raw("1.000000000000000", 0);
    for mode in 0..=4 {
        let expected = if mode == 2 { up } else { down };
        let (got, flags) = bid64_add_with_flags(x, y, mode);
        assert_eq!((got, flags), (expected, 0x20));
        assert_eq!(bid64_add(x, y, mode), Ok(expected));
        assert!(RoundingMode::try_from(mode as u32).is_ok());
    }
    for invalid in [5, 99, u32::MAX] {
        assert!(RoundingMode::try_from(invalid).is_err());
    }
    for invalid in [99, -1, i32::MIN, i32::MAX] {
        assert_eq!(bid64_from_string_raw("1", invalid), (0x7c00000000000000, 1));
    }
    for invalid in [99, -1, i64::MIN, i64::MAX] {
        assert_eq!(bid32_from_string_raw("1", invalid), (0x7c000000, 1));
        let (got, flags) = bid128_from_string("1", invalid);
        assert_eq!((got.lo, got.hi, flags), (0, 0x7c00000000000000, 1));
        assert_eq!(bid754::generated::to_binary64::bid64_to_binary32(x, invalid), (0x7fc00000, 1));
        assert_eq!(bid754::generated::to_binary64::bid64_to_binary64(x, invalid), (0x7ff8000000000000, 1));
        let (got, flags) = bid754::generated::to_binary64::bid64_to_binary128(x, invalid);
        assert_eq!((got.lo, got.hi, flags), (0, 0x7fff800000000000, 1));
        assert_eq!(bid754::generated::lrint64::bid64_lrint(x, invalid), (i64::MIN, 1));
    }
    for mode in 0..=5 {
        assert_eq!(bid64_from_string_raw("1", mode), (x, 0));
        assert_eq!(bid32_from_string_raw("1", mode as i64), bid32_from_string_raw("1", 0));
        let (got, flags) = bid128_from_string("1", mode as i64);
        let (expected, _) = bid128_from_string("1", 0);
        assert_eq!((got.lo, got.hi, flags), (expected.lo, expected.hi, 0));
    }
    assert_eq!(bid32_from_string_raw("1e-102", 5).1, 0x30);
    assert_eq!(bid64_from_string_raw("1e-399", 5).1, 0x30);
    assert_eq!(bid128_from_string("1e-6177", 5).1, 0x30);
    let (half_down, flags) = bid64_from_string_raw("1.0000000000000005", 5);
    let (truncated, _) = bid64_from_string_raw("1.000000000000000", 0);
    assert_eq!((half_down, flags), (truncated, 0x20));
}
`
