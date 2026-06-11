package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPostProcessRewritesFmaDoneGoto(t *testing.T) {
	input := `pub(crate) fn bid_fma_delta_ge_zero() {
    let mut z_exp = (*z_exp_ptr);
    let mut p_exp = (*p_exp_ptr);
    if ((p34 <= (delta - 1)) || other_condition) {
        // goto done; // TODO: convert goto to loop/break
    }
    // label: done
    (*z_exp_ptr) = z_exp;
    (*p_exp_ptr) = p_exp;
}
`

	got := postProcess(input)

	for _, forbidden := range []string{"goto done", "TODO: convert goto", "label: done"} {
		if strings.Contains(got, forbidden) {
			t.Fatalf("postProcess left %q in output:\n%s", forbidden, got)
		}
	}
	for _, required := range []string{"    'done: {\n", "        break 'done;\n", "    }\n    (*z_exp_ptr) = z_exp;"} {
		if !strings.Contains(got, required) {
			t.Fatalf("postProcess output missing %q:\n%s", required, got)
		}
	}
}

func TestRejectGeneratedFallbacks(t *testing.T) {
	for _, tc := range []struct {
		name string
		code string
	}{
		{name: "todo macro", code: "pub fn f() { todo!(\"x\"); }"},
		{name: "todo comment", code: "// TODO: unsupported\n"},
		{name: "empty body marker", code: "pub fn f() {\n    // empty body\n}\n"},
		{name: "source read marker", code: "pub fn f() {\n    // error reading source\n}\n"},
		{name: "closure marker", code: "let f = /* closure TODO */;"},
		{name: "goto marker", code: "// goto done; // TODO: convert goto to loop/break\n"},
		{name: "label marker", code: "// label: done\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := rejectGeneratedFallbacks("bad.go", tc.code); err == nil {
				t.Fatal("expected fallback rejection")
			}
		})
	}
}

func TestRejectFinalGeneratedFallbacksCatchesExpressionFallback(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "bad.rs"), []byte("pub fn f() { let _x = /* unsupported_call() */ 0; }\n"), 0o644); err != nil {
		t.Fatalf("WriteFile bad.rs: %v", err)
	}
	if err := rejectFinalGeneratedFallbacks(dir); err == nil {
		t.Fatalf("rejectFinalGeneratedFallbacks accepted expression fallback")
	}
}

func TestRejectGeneratedFallbacksAllowsConvertedCode(t *testing.T) {
	code := `pub(crate) fn ok() {
    let mut x = 1;
    x = x.wrapping_add(1);
}
`
	if err := rejectGeneratedFallbacks("ok.go", code); err != nil {
		t.Fatalf("unexpected fallback rejection: %v", err)
	}
}

func TestConvertImmediateFuncLiteralCall(t *testing.T) {
	src := []byte(`func() int32 {
    if n < 0 {
        return int32(0x7fffffff)
    }
    return n
}()`)
	fset := token.NewFileSet()
	expr, err := parser.ParseExprFrom(fset, "inline.go", src, 0)
	if err != nil {
		t.Fatalf("ParseExprFrom: %v", err)
	}

	got := convertExprStr(fset, expr, src)

	for _, forbidden := range []string{"closure TODO", "TODO:"} {
		if strings.Contains(got, forbidden) {
			t.Fatalf("converted closure left fallback marker %q:\n%s", forbidden, got)
		}
	}
	for _, required := range []string{"|| -> i32", "return (0x7fffffff as i32);", "return n;"} {
		if !strings.Contains(got, required) {
			t.Fatalf("converted closure missing %q:\n%s", required, got)
		}
	}
}

func TestConvertAndNotOperators(t *testing.T) {
	fset := token.NewFileSet()
	src := []byte(`x &^ mask`)
	expr, err := parser.ParseExprFrom(fset, "andnot.go", src, 0)
	if err != nil {
		t.Fatalf("ParseExprFrom binary: %v", err)
	}
	if got := convertExprStr(fset, expr, src); got != "(x & !mask)" {
		t.Fatalf("AND_NOT expression = %q", got)
	}

	stmtSrc := []byte(`package p
func f() {
    flags &^= BID_UNDERFLOW_EXCEPTION
}
`)
	file, err := parser.ParseFile(fset, "andnot_stmt.go", stmtSrc, 0)
	if err != nil {
		t.Fatalf("ParseFile stmt: %v", err)
	}
	fn := file.Decls[0].(*ast.FuncDecl)
	got := convertStmt(fset, fn.Body.List[0], stmtSrc, 1, nil)
	if want := "    flags &= !BID_UNDERFLOW_EXCEPTION;\n"; got != want {
		t.Fatalf("AND_NOT assignment = %q, want %q", got, want)
	}
}

func TestTypeCheckedIntegerOpsUseGoOverflowAndShiftSemantics(t *testing.T) {
	code := convertTypeCheckedTestFile(t, "ops.go", `package bidgo

func ops(u uint64, i int32, s uint, big uint64, n int32) (uint64, int32) {
	u = u + 1
	u -= 2
	u <<= s
	u >>= s
	u = u << big
	u = u >> big
	i = i - 1
	i *= 2
	i = i >> s
	i = i >> n
	i >>= s
	i = -i
	return u, i
}
`)
	for _, required := range []string{
		"u = (u.wrapping_add(1));",
		"u = u.wrapping_sub(2);",
		"u = go_checked_shl_u64(u, go_shift_count_u64((s) as u64));",
		"u = go_checked_shr_u64(u, go_shift_count_u64((s) as u64));",
		"u = (go_checked_shl_u64(u, go_shift_count_u64((big) as u64)));",
		"u = (go_checked_shr_u64(u, go_shift_count_u64((big) as u64)));",
		"i = (i.wrapping_sub(1));",
		"i = i.wrapping_mul(2);",
		"i = (go_checked_shr_i32(i, go_shift_count_u64((s) as u64)));",
		"i = (go_checked_shr_i32(i, go_shift_count_i64((n) as i64)));",
		"i = go_checked_shr_i32(i, go_shift_count_u64((s) as u64));",
		"i = (i.wrapping_neg());",
	} {
		if !strings.Contains(code, required) {
			t.Fatalf("converted integer op code missing %q:\n%s", required, code)
		}
	}
}

func TestGoIntUintBuiltinsUseGo64BitSemantics(t *testing.T) {
	code := convertTypeCheckedTestFile(t, "int_width.go", `package bidgo

import (
	"math/bits"
	"strconv"
)

var table = [2]int32{3, 4}

func callee(exp int, raw uint) int {
	return exp + int(raw)
}

func width(xs []byte, u uint64, idx int) (int, uint) {
	n, err := strconv.Atoi("123")
	if err != nil {
		return 0, 0
	}
	digits := len(xs)
	bitLen := bits.Len64(u)
	tableValue := int(table[idx])
	extra := 16 - digits
	raw := uint(bitLen)
	n += tableValue
	n += extra
	return callee(n, raw), raw
}
`)
	for _, required := range []string{
		"pub(crate) fn callee(mut exp: i64, mut raw: u64) -> i64",
		"pub(crate) fn width(mut xs: &mut [u8], mut u: u64, mut idx: i64) -> (i64, u64)",
		`let (mut n, mut err) = go_atoi("123");`,
		"let mut digits = (xs.len() as i64);",
		"let mut bitLen = (64 - (u).leading_zeros()) as i64;",
		"let mut tableValue = (table[idx as usize] as i64);",
		"let mut extra = ((16 as i64).wrapping_sub(digits));",
		"let mut raw = (bitLen as u64);",
		"n = n.wrapping_add(tableValue);",
		"n = n.wrapping_add(extra);",
		"return (callee(n, raw), raw);",
	} {
		if !strings.Contains(code, required) {
			t.Fatalf("converted Go int/uint width code missing %q:\n%s", required, code)
		}
	}
	for _, forbidden := range []string{
		"mut exp: i32",
		"mut raw: u32",
		"-> (i32, u32)",
		`parse::<i32>`,
		"xs.len() as i32",
		"leading_zeros()) as i32",
		"table[idx as usize] as i32",
		"(16 as i32).wrapping_sub(digits)",
		"bitLen as u32",
	} {
		if strings.Contains(code, forbidden) {
			t.Fatalf("converted Go int/uint width code contains forbidden %q:\n%s", forbidden, code)
		}
	}
}

func convertTypeCheckedTestFile(t *testing.T, name, src string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	reg := &Registry{
		Types:     map[string]TypeDef{},
		Constants: map[string]ConstDef{},
		Tables:    map[string]TableDef{},
		Functions: map[string]FuncDef{},
	}
	activeRegistry = reg
	activeSourceFunctions = collectSourceFunctionNames([]string{path})
	fset, parsedTargets, info := parseTypeCheckedPackage(dir, []string{path})
	oldTypeInfo := activeTypeInfo
	activeTypeInfo = info
	t.Cleanup(func() {
		activeTypeInfo = oldTypeInfo
	})
	code, err := convertParsedFile(fset, parsedTargets[path], path, reg)
	if err != nil {
		t.Fatalf("convertParsedFile: %v", err)
	}
	return code
}

func TestConvertByteCastCharArithmetic(t *testing.T) {
	fset := token.NewFileSet()
	src := []byte(`byte('0' + n%10)`)
	expr, err := parser.ParseExprFrom(fset, "bytecast.go", src, 0)
	if err != nil {
		t.Fatalf("ParseExprFrom: %v", err)
	}
	got := convertExprStr(fset, expr, src)
	want := "((b'0' + ((n % 10) as u8)) as u8)"
	if got != want {
		t.Fatalf("byte cast conversion = %q, want %q", got, want)
	}
}

func TestPostProcessRewritesByteLiteralSubtractions(t *testing.T) {
	input := `pub(crate) fn f(mut x: u8) -> u8 {
    if ((x - b'A') <= (b'Z' - b'A')) {
        return ((ps_at!(0) - b'0') + buffer[0].wrapping_sub(b'0'));
    }
    return (buffer[i as usize] - b'0');
}
`
	got := postProcess(input)
	for _, forbidden := range []string{"x - b'A'", "ps_at!(0) - b'0'", "buffer[i as usize] - b'0'"} {
		if strings.Contains(got, forbidden) {
			t.Fatalf("postProcess left byte subtraction %q:\n%s", forbidden, got)
		}
	}
	for _, required := range []string{"x.wrapping_sub(b'A')", "ps_at!(0).wrapping_sub(b'0')", "buffer[i as usize].wrapping_sub(b'0')"} {
		if !strings.Contains(got, required) {
			t.Fatalf("postProcess missing %q:\n%s", required, got)
		}
	}
}

func TestConvertNilComparisons(t *testing.T) {
	fset := token.NewFileSet()
	for _, tc := range []struct {
		src  string
		want string
	}{
		{src: `err != nil`, want: "err.is_some()"},
		{src: `err == nil`, want: "err.is_none()"},
		{src: `frac != nil`, want: "/* unsupported nil comparison */"},
		{src: `frac == nil`, want: "/* unsupported nil comparison */"},
	} {
		expr, err := parser.ParseExprFrom(fset, "nilcmp.go", []byte(tc.src), 0)
		if err != nil {
			t.Fatalf("ParseExprFrom(%q): %v", tc.src, err)
		}
		if got := convertExprStr(fset, expr, []byte(tc.src)); got != tc.want {
			t.Fatalf("convert %q = %q, want %q", tc.src, got, tc.want)
		}
	}
}

func TestConvertMutableStringParam(t *testing.T) {
	path := filepath.Join(t.TempDir(), "string_param.go")
	src := []byte(`package bidgo

import "strings"

func parse(s string) string {
    s = strings.TrimSpace(s)
    if s[0] == '+' {
        s = s[1:]
    }
    return s
}
`)
	if err := os.WriteFile(path, src, 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	activeRegistry = &Registry{
		Types:     map[string]TypeDef{},
		Constants: map[string]ConstDef{},
		Tables:    map[string]TableDef{},
		Functions: map[string]FuncDef{},
	}
	activeSourceFunctions = nil
	code, err := convertFile(path, activeRegistry)
	if err != nil {
		t.Fatalf("convertFile: %v", err)
	}
	for _, required := range []string{
		"let mut s = s.as_ref().to_string();",
		"s.as_bytes()[0 as usize]",
		"s = (&s[1 as usize..]).to_string();",
		"return s;",
	} {
		if !strings.Contains(code, required) {
			t.Fatalf("converted string-param code missing %q:\n%s", required, code)
		}
	}
}

func TestRegistryUnsupportedRustTypeDoesNotOwnGo2rsType(t *testing.T) {
	if registryTypeOwnsRust(TypeDef{Fields: []FieldDef{{Name: "coeff", Type: "&mut Int"}}}) {
		t.Fatal("registryTypeOwnsRust accepted unsupported Rust field type")
	}
	if !registryTypeOwnsRust(TypeDef{Fields: []FieldDef{{Name: "w", Type: "[u64; 2]"}}}) {
		t.Fatal("registryTypeOwnsRust rejected supported fixed-width field type")
	}
}

func TestShouldConvertFileIncludesFormerAlternateGeneratedFiles(t *testing.T) {
	for _, name := range []string{"nexttoward64.go", "to_binary64.go"} {
		if !shouldConvertFile(name) {
			t.Fatalf("shouldConvertFile(%q) = false, want true for go2rs source conversion", name)
		}
	}
}

func TestCleanGeneratedDirRemovesFormerAlternateGeneratedFiles(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"prelude.rs", "nexttoward64.rs", "to_binary64.rs", "old.rs"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(name), 0o644); err != nil {
			t.Fatalf("WriteFile(%q): %v", name, err)
		}
	}

	cleanGeneratedDir(dir, nil)

	for _, name := range []string{"prelude.rs"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Fatalf("expected %s to be preserved: %v", name, err)
		}
	}
	for _, name := range []string{"nexttoward64.rs", "to_binary64.rs", "old.rs"} {
		if _, err := os.Stat(filepath.Join(dir, name)); !os.IsNotExist(err) {
			t.Fatalf("%s stat err = %v, want removed", name, err)
		}
	}
}

func TestRejectGeneratedOwnershipLeaks(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "ok.rs"), []byte("// Auto-generated from x.go by go2rs. Do not edit.\n"), 0o644); err != nil {
		t.Fatalf("WriteFile ok.rs: %v", err)
	}
	if err := rejectGeneratedOwnershipLeaks(dir); err != nil {
		t.Fatalf("rejectGeneratedOwnershipLeaks ok dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(dir, "bad.rs"), []byte("// Auto-generated from x.go by tools/codegen rust-optimize. Do not edit.\n"), 0o644); err != nil {
		t.Fatalf("WriteFile bad.rs: %v", err)
	}
	if err := rejectGeneratedOwnershipLeaks(dir); err == nil {
		t.Fatalf("rejectGeneratedOwnershipLeaks accepted stale codegen marker")
	}
}

func TestFormerAlternateSourcesConvertWithoutFallbacks(t *testing.T) {
	root := findProjectRoot()
	reg := loadRegistry(filepath.Join(root, "tools", "registry", "symbols.json"))
	activeRegistry = reg
	for _, name := range []string{
		"to_binary64.go",
		"nexttoward64.go",
	} {
		code, err := convertFile(filepath.Join(root, bidGoDir, name), reg)
		if err != nil {
			t.Fatalf("convertFile(%s): %v", name, err)
		}
		if strings.Contains(code, "templates/") {
			t.Fatalf("%s conversion leaked template path:\n%s", name, code)
		}
		if err := rejectGeneratedFallbacks(name, code); err != nil {
			t.Fatalf("%s conversion contains fallback: %v", name, err)
		}
	}
}

func TestOptimizeBid32StringParseAvoidsAllocation(t *testing.T) {
	input := `pub fn bid32_from_string_raw(ps: impl AsRef<str>, mut rnd_mode: i64) -> (u32, u32) {
    let ps = ps.as_ref();
    let s = (ps).trim_start_matches(|c| " \t".contains(c)).as_bytes().to_vec();
    if ((s.len() as i64) == 0) {
        return (0x7c000000, 0);
    }
    let mut c = s[0];
    let mut sl = String::from_utf8_lossy(&s).to_ascii_lowercase();
    if ((((c != b'.') && (c != b'-')) && (c != b'+')) && (((c < b'0') || (c > b'9')))) {
        if ((sl == "inf") || (sl == "infinity")) {
            return (0x78000000, 0);
        }
        if (sl).starts_with("snan") {
            return (0x7e000000, 0);
        }
        return (0x7c000000, 0);
    }
    if ((s.len() as i64) > 1) {
        let mut sl1 = String::from_utf8_lossy(&s[1 as usize..]).to_ascii_lowercase();
        if ((sl1 == "inf") || (sl1 == "infinity")) {
            if (c == b'+') {
                return (0x78000000, 0);
            } else if (c == b'-') {
                return (0xf8000000, 0);
            }
            return (0x7c000000, 0);
        }
        if (sl1).starts_with("snan") {
            if (c == b'-') {
                return (0xfe000000, 0);
            }
            return (0x7e000000, 0);
        }
        if (sl1 == "nan") {
            if (c == b'-') {
                return (0xfc000000, 0);
            }
            return (0x7c000000, 0);
        }
    }
    return (0, 0);
}
`
	got, err := optimizeRustStringHotpaths("bid32_string.go", input)
	if err != nil {
		t.Fatalf("optimizeRustStringHotpaths: %v", err)
	}

	for _, forbidden := range []string{
		"as_bytes().to_vec()",
		"String::from_utf8_lossy",
		"to_ascii_lowercase",
	} {
		if strings.Contains(got, forbidden) {
			t.Fatalf("optimized code still contains %q:\n%s", forbidden, got)
		}
	}
	for _, required := range []string{
		`let s = (ps).trim_start_matches(|c| " \t".contains(c)).as_bytes();`,
		`s.eq_ignore_ascii_case(b"inf")`,
		`let sl1 = &s[1 as usize..];`,
		`sl1.eq_ignore_ascii_case(b"nan")`,
	} {
		if !strings.Contains(got, required) {
			t.Fatalf("optimized code missing %q:\n%s", required, got)
		}
	}
}

func TestOptimizeBid128MiscLowersStructuredClosure(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bid128_misc.rs")
	input := `pub fn other() {}

pub fn bid128_scalbln(mut x: BID_UINT128, mut n: i64, mut rnd_mode: i64, pfpsf: &mut u32) -> BID_UINT128 {
    let mut n1 = (n as i32);
    n1 = (|| -> i32 {
    if ((n1 as i64) < n) {
        return (0x7fffffff as i32);
    }
    if ((n1 as i64) > n) {
        return ((-0x80000000) as i32);
    }
    return n1;
})();
    return bid128_scalbn(x, (n1 as i64), rnd_mode, pfpsf);
}

pub fn tail() {}
`
	if err := os.WriteFile(path, []byte(input), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	optimizeBid128Misc(path)

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	got := string(data)
	for _, forbidden := range []string{"(|| -> i32", "closure TODO", "(n1 as i32), rnd_mode"} {
		if strings.Contains(got, forbidden) {
			t.Fatalf("optimizeBid128Misc left %q in output:\n%s", forbidden, got)
		}
	}
	for _, required := range []string{"i32::MAX", "i32::MIN", "return bid128_scalbn(x, (n1 as i64), rnd_mode, pfpsf);"} {
		if !strings.Contains(got, required) {
			t.Fatalf("optimizeBid128Misc output missing %q:\n%s", required, got)
		}
	}
}
