package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sky1core/bid754/devtools/internal/genmarker"
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

	for _, rejected := range []string{"goto done", "TODO: convert goto", "label: done"} {
		if strings.Contains(got, rejected) {
			t.Fatalf("postProcess left %q in output:\n%s", rejected, got)
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

func TestRejectFinalGeneratedFallbacksCatchesMatchArmBreak(t *testing.T) {
	dir := t.TempDir()
	bad := "pub fn f() {\n    loop {\n        match x {\n            1 => {\n                break;\n            }\n            _ => {}\n        }\n    }\n}\n"
	if err := os.WriteFile(filepath.Join(dir, "bad.rs"), []byte(bad), 0o644); err != nil {
		t.Fatalf("WriteFile bad.rs: %v", err)
	}
	if err := rejectFinalGeneratedFallbacks(dir); err == nil {
		t.Fatalf("rejectFinalGeneratedFallbacks accepted a residual match-arm break")
	}
}

func TestRejectResidualMatchArmBreakAcceptsLegitBreaks(t *testing.T) {
	for _, tc := range []struct {
		name string
		code string
	}{
		{name: "loop break", code: "pub fn f() {\n    loop {\n        if x {\n            break;\n        }\n    }\n}\n"},
		{name: "while break", code: "pub fn f() {\n    while (c) {\n        break;\n    }\n}\n"},
		{name: "for break", code: "pub fn f() {\n    for i in 0..n {\n        break;\n    }\n}\n"},
		{name: "labeled loop break", code: "pub fn f() {\n    'roundC2: loop {\n        continue 'roundC2;\n        break;\n    }\n}\n"},
		{name: "labeled block break", code: "pub fn f() {\n    'done: {\n        break 'done;\n    }\n}\n"},
		{name: "match arm no break", code: "pub fn f() {\n    match x {\n        1 => {\n            y = 1;\n        }\n        _ => {}\n    }\n}\n"},
		{name: "non-block arms", code: "pub fn f() {\n    let (v, e) = match r {\n        Ok(v) => (v, None),\n        Err(_) => (0, Some(\"atoi\")),\n    };\n}\n"},
		{name: "loop inside arm break targets inner loop", code: "pub fn f() {\n    loop {\n        match x {\n            1 => {\n                loop {\n                    break;\n                }\n            }\n            _ => {}\n        }\n    }\n}\n"},
		{name: "format string braces", code: "pub fn f() {\n    loop {\n        let s = format!(\"{}\", x);\n        break;\n    }\n}\n"},
		{name: "byte literal brace", code: "pub fn f() {\n    while (c == b'{') {\n        break;\n    }\n}\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := rejectResidualMatchArmBreak("ok.rs", tc.code); err != nil {
				t.Fatalf("unexpected rejection: %v", err)
			}
		})
	}
}

func TestRejectResidualMatchArmBreakFailsOnArmBreak(t *testing.T) {
	for _, tc := range []struct {
		name string
		code string
	}{
		{name: "block arm break in loop", code: "pub fn f() {\n    loop {\n        match x {\n            1 => {\n                break;\n            }\n            _ => {}\n        }\n    }\n}\n"},
		{name: "default arm break in loop", code: "pub fn f() {\n    loop {\n        match x {\n            1 => {}\n            _ => {\n                break;\n            }\n        }\n    }\n}\n"},
		{name: "nested if arm break in loop", code: "pub fn f() {\n    loop {\n        match x {\n            2 => {\n                if cond {\n                    break;\n                }\n            }\n            _ => {}\n        }\n    }\n}\n"},
		{name: "bare arm break in loop", code: "pub fn f() {\n    loop {\n        match x {\n            1 => break,\n            _ => {}\n        }\n    }\n}\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := rejectResidualMatchArmBreak("bad.rs", tc.code); err == nil {
				t.Fatalf("expected match-arm break rejection")
			}
		})
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

	for _, rejected := range []string{"closure TODO", "TODO:"} {
		if strings.Contains(got, rejected) {
			t.Fatalf("converted closure left fallback marker %q:\n%s", rejected, got)
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
	for _, rejected := range []string{
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
		if strings.Contains(code, rejected) {
			t.Fatalf("converted Go int/uint width code contains rejected %q:\n%s", rejected, code)
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
	var err error
	activeSourceFunctions, err = collectSourceFunctionNames([]string{path})
	if err != nil {
		t.Fatalf("collectSourceFunctionNames: %v", err)
	}
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

func TestConvertSwitchFallthroughMergesBareCaseIntoNextArm(t *testing.T) {
	code := convertTypeCheckedTestFile(t, "fallthrough_merge.go", `package bidgo

func classify(rmode int, remainder uint64) uint32 {
	var status uint32 = 32
	switch rmode {
	case 0:
		fallthrough
	case 4:
		if remainder == 0x8000000000000000 {
			status = 0
		}
	case 1:
		fallthrough
	case 3:
		if remainder == 0 {
			status = 0
		}
	default:
		status = 1
	}
	return status
}
`)
	for _, required := range []string{
		"0 | 4 => {",
		"1 | 3 => {",
		"_ => {",
	} {
		if !strings.Contains(code, required) {
			t.Fatalf("converted switch missing %q:\n%s", required, code)
		}
	}
	// The bug this checks against: a bare `case 0: fallthrough` becoming an
	// empty match arm that silently drops the shared body.
	for _, rejected := range []string{
		"0 => {\n            }",
		"1 => {\n            }",
	} {
		if strings.Contains(code, rejected) {
			t.Fatalf("converted switch left empty fallthrough arm %q:\n%s", rejected, code)
		}
	}
	if err := rejectGeneratedFallbacks("fallthrough_merge.go", code); err != nil {
		t.Fatalf("unexpected fallback rejection: %v", err)
	}
}

func TestConvertSwitchFallthroughMergesConsecutiveBareCases(t *testing.T) {
	code := convertTypeCheckedTestFile(t, "fallthrough_consecutive.go", `package bidgo

func widen(rmode int) uint32 {
	var status uint32
	switch rmode {
	case 0:
		fallthrough
	case 1:
		fallthrough
	case 2:
		status = 7
	default:
		status = 1
	}
	return status
}
`)
	// Two consecutive bare fallthroughs must accumulate into one merged
	// pattern, not leave any empty arm behind.
	if !strings.Contains(code, "0 | 1 | 2 => {") {
		t.Fatalf("converted switch missing accumulated pattern %q:\n%s", "0 | 1 | 2 => {", code)
	}
	for _, rejected := range []string{
		"0 => {\n            }",
		"1 => {\n            }",
	} {
		if strings.Contains(code, rejected) {
			t.Fatalf("converted switch left empty fallthrough arm %q:\n%s", rejected, code)
		}
	}
	if err := rejectGeneratedFallbacks("fallthrough_consecutive.go", code); err != nil {
		t.Fatalf("unexpected fallback rejection: %v", err)
	}
}

func TestConvertSwitchFallthroughChainsNonEmptyBodies(t *testing.T) {
	code := convertTypeCheckedTestFile(t, "fallthrough_chain.go", `package bidgo

func chain(mode int) int {
	var n int
	switch mode {
	case 0:
		n += 1
		fallthrough
	case 1:
		n += 2
	case 2:
		fallthrough
	default:
		n += 4
	}
	return n
}
`)
	// case 0 keeps its own statement and then executes case 1's body too.
	arm0 := "0 => {\n            n = n.wrapping_add(1);\n            n = n.wrapping_add(2);\n        }"
	if !strings.Contains(code, arm0) {
		t.Fatalf("converted switch missing chained arm %q:\n%s", arm0, code)
	}
	// bare fallthrough into default duplicates the default body instead of
	// merging into the `_` pattern.
	arm2 := "2 => {\n            n = n.wrapping_add(4);\n        }"
	if !strings.Contains(code, arm2) {
		t.Fatalf("converted switch missing fallthrough-into-default arm %q:\n%s", arm2, code)
	}
	if err := rejectGeneratedFallbacks("fallthrough_chain.go", code); err != nil {
		t.Fatalf("unexpected fallback rejection: %v", err)
	}
}

func TestConvertExpressionlessSwitchFallthroughIsRejected(t *testing.T) {
	code := convertTypeCheckedTestFile(t, "fallthrough_boolswitch.go", `package bidgo

func boolSwitch(n int) int {
	switch {
	case n > 0:
		fallthrough
	default:
		n = 0
	}
	return n
}
`)
	if err := rejectGeneratedFallbacks("fallthrough_boolswitch.go", code); err == nil {
		t.Fatalf("expressionless switch fallthrough was not rejected:\n%s", code)
	}
}

func TestConvertSwitchMidDefaultArmEmittedLast(t *testing.T) {
	code := convertTypeCheckedTestFile(t, "mid_default.go", `package bidgo

func pick(x int) int {
	var n int
	switch x {
	case 1:
		n = 1
	default:
		n = 9
	case 2:
		n = 2
	}
	return n
}
`)
	// Go's default matches only when no case matches, regardless of source
	// position; a `_` arm emitted at the default's mid-switch position would
	// shadow the `2 =>` arm in Rust's top-down match (x == 2 would take the
	// default body).
	armDefault := strings.Index(code, "_ => {")
	armTwo := strings.Index(code, "2 => {")
	if armDefault == -1 || armTwo == -1 {
		t.Fatalf("converted switch missing `_` or `2 =>` arm:\n%s", code)
	}
	if armDefault < armTwo {
		t.Fatalf("`_` arm emitted before `2 =>` arm; later cases are unreachable:\n%s", code)
	}
	// The reordered `_` arm must still carry the default body.
	if !strings.Contains(code[armDefault:], "n = 9") {
		t.Fatalf("`_` arm lost the default body:\n%s", code)
	}
	if err := rejectGeneratedFallbacks("mid_default.go", code); err != nil {
		t.Fatalf("unexpected fallback rejection: %v", err)
	}
}

func TestConvertSwitchFallthroughIntoMidDefaultIsRejected(t *testing.T) {
	code := convertTypeCheckedTestFile(t, "mid_default_fallthrough_in.go", `package bidgo

func pick(x int) int {
	var n int
	switch x {
	case 1:
		n = 1
		fallthrough
	default:
		n = 9
	case 2:
		n = 2
	}
	return n
}
`)
	// A case falling through into a mid-switch default couples the default
	// body to source order; reordering arms is not obviously safe there, so
	// the converter must reject instead of emitting silently.
	if err := rejectGeneratedFallbacks("mid_default_fallthrough_in.go", code); err == nil {
		t.Fatalf("fallthrough into a mid-switch default was not rejected:\n%s", code)
	}
}

func TestConvertSwitchMidDefaultFallingThroughIsRejected(t *testing.T) {
	code := convertTypeCheckedTestFile(t, "mid_default_fallthrough_out.go", `package bidgo

func pick(x int) int {
	var n int
	switch x {
	default:
		n = 9
		fallthrough
	case 2:
		n = 2
	case 3:
		n = 3
	}
	return n
}
`)
	// A mid-switch default that itself falls through into the next case is
	// likewise rejected instead of being reordered.
	if err := rejectGeneratedFallbacks("mid_default_fallthrough_out.go", code); err == nil {
		t.Fatalf("mid-switch default with fallthrough was not rejected:\n%s", code)
	}
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
	for _, rejected := range []string{"x - b'A'", "ps_at!(0) - b'0'", "buffer[i as usize] - b'0'"} {
		if strings.Contains(got, rejected) {
			t.Fatalf("postProcess left byte subtraction %q:\n%s", rejected, got)
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
	// The rebinding here is a whole-string assignment, not a slice: slicing a
	// Go string is rejected outright (TestRejectGoStringSlicing), so a fixture
	// that sliced would no longer reach the binding logic under test.
	src := []byte(`package bidgo

import "strings"

func parse(s string) string {
    s = strings.TrimSpace(s)
    if s[0] == '+' {
        s = strings.TrimSpace(s)
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
		"return s;",
	} {
		if !strings.Contains(code, required) {
			t.Fatalf("converted string-param code missing %q:\n%s", required, code)
		}
	}
}

// TestRejectGoStringSlicing pins the generation-time block on Go string
// slicing. Go slices a string at any byte offset; the generated Rust cuts a
// &str and panics when the cut lands inside a multi-byte character, so a sliced
// probe on the public parse path traps on ordinary rejected input such as
// "1234é" whenever its cut can land mid-character. No Go-side test can observe
// that (Go never faults on the slice), so the block has to happen where the
// Rust is produced.
//
// The block refuses the construct rather than trying to decide whether a
// particular cut is reachable mid-character, because that depends on branches
// elsewhere in the function.
//
// The []byte and array rows are the other half: the block admits exactly the
// bases that have a Rust slice borrow, so it must leave byte slicing alone,
// which is how the bid64/bid128 parse ports read their input.
func TestRejectGoStringSlicing(t *testing.T) {
	cases := []struct {
		name     string
		body     string
		rejected bool
	}{
		{
			name:     "string param open slice",
			body:     "func probe(s string) string {\n    return s[1:]\n}\n",
			rejected: true,
		},
		{
			name:     "string param prefix slice",
			body:     "func probe(s string) string {\n    return s[:4]\n}\n",
			rejected: true,
		},
		{
			name:     "string local slice",
			body:     "func probe(s string) string {\n    t := s\n    return t[1:]\n}\n",
			rejected: true,
		},
		{
			name:     "string self slice assignment",
			body:     "func probe(s string) string {\n    s = s[1:]\n    return s\n}\n",
			rejected: true,
		},
		{
			name:     "string slice inside call argument",
			body:     "func probe(s string) int {\n    return len(s[1:])\n}\n",
			rejected: true,
		},
		// An untyped string constant has types.Basic kind UntypedString, not
		// String. A check that matched on kind let both of these through and
		// emitted a &str cut; the fail-closed form refuses them because neither
		// base is a slice or array.
		{
			name:     "untyped string constant slice",
			body:     "const k = \"abcd\"\n\nfunc probe() string {\n    return k[1:]\n}\n",
			rejected: true,
		},
		{
			name:     "string literal slice",
			body:     "func probe() string {\n    return \"abcd\"[1:]\n}\n",
			rejected: true,
		},
		{
			name:     "named string type slice",
			body:     "type token string\n\nfunc probe(t token) token {\n    return t[1:]\n}\n",
			rejected: true,
		},
		{
			name:     "three index string slice",
			body:     "func probe(s string) string {\n    return s[1:2]\n}\n",
			rejected: true,
		},
		{
			name:     "string slice inside closure",
			body:     "func probe(s string) func() string {\n    return func() string { return s[1:] }\n}\n",
			rejected: true,
		},
		{
			name:     "byte slice param stays allowed",
			body:     "func probe(b []byte) []byte {\n    return b[1:]\n}\n",
			rejected: false,
		},
		{
			name:     "three index byte slice stays allowed",
			body:     "func probe(b []byte) []byte {\n    return b[1:2:2]\n}\n",
			rejected: false,
		},
		{
			name:     "named byte slice type stays allowed",
			body:     "type buf []byte\n\nfunc probe(b buf) buf {\n    return b[1:]\n}\n",
			rejected: false,
		},
		{
			name:     "pointer to array stays allowed",
			body:     "func probe(p *[8]byte) []byte {\n    return p[1:]\n}\n",
			rejected: false,
		},
		{
			name:     "byte array local stays allowed",
			body:     "func probe() []byte {\n    var buf [8]byte\n    return buf[:4]\n}\n",
			rejected: false,
		},
		{
			name:     "string byte index stays allowed",
			body:     "func probe(s string) byte {\n    return s[1]\n}\n",
			rejected: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "slice_probe.go")
			if err := os.WriteFile(path, []byte("package bidgo\n\n"+tc.body), 0o644); err != nil {
				t.Fatalf("WriteFile: %v", err)
			}
			reg := &Registry{
				Types:     map[string]TypeDef{},
				Constants: map[string]ConstDef{},
				Tables:    map[string]TableDef{},
				Functions: map[string]FuncDef{},
			}
			activeRegistry = reg
			activeSourceFunctions = nil
			code, err := convertFile(path, reg)
			if !tc.rejected {
				if err != nil {
					t.Fatalf("convertFile rejected a non-string slice: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("convertFile accepted a Go string slice; generated code:\n%s", code)
			}
			// The message must name the base and the reason, so a future reader
			// of a failing build learns what to change rather than only that
			// something was refused.
			if !strings.Contains(err.Error(), "panics inside a multi-byte character") {
				t.Fatalf("convertFile error = %q, want the string-slicing identity", err)
			}
			if !strings.Contains(err.Error(), "slice_probe.go:") {
				t.Fatalf("convertFile error = %q, want a file:line location", err)
			}
		})
	}
}

// TestRejectGoStringSlicingRequiresTypeInfo keeps the block from degrading into
// a no-op. isGoStringExpr falls back to a parameter-name set when type info is
// absent, and that fallback cannot see string locals, so a pass that ran
// without the type checker would report a clean file it never checked.
func TestRejectGoStringSlicingRequiresTypeInfo(t *testing.T) {
	prevInfo := activeTypeInfo
	activeTypeInfo = nil
	defer func() { activeTypeInfo = prevInfo }()

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "probe.go", "package bidgo\n", parser.ParseComments)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	if err := rejectGoStringSlicing(fset, f, "probe.go"); err == nil {
		t.Fatal("rejectGoStringSlicing passed a file it could not type-check")
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
	for _, name := range []string{"decimal64.go", "nexttoward64.go", "to_binary64.go"} {
		if !shouldConvertFile(name) {
			t.Fatalf("shouldConvertFile(%q) = false, want true for go2rs source conversion", name)
		}
	}
}

func TestConvertFileRejectsReceiverMethods(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "receiver.go")
	src := []byte(`package bidgo

type decimal uint64

func (d decimal) bits() uint64 { return uint64(d) }
`)
	if err := os.WriteFile(path, src, 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	reg := &Registry{
		Types:     map[string]TypeDef{},
		Constants: map[string]ConstDef{},
		Tables:    map[string]TableDef{},
		Functions: map[string]FuncDef{},
	}
	if names, err := collectSourceFunctionNames([]string{path}); err == nil {
		t.Fatalf("collectSourceFunctionNames silently accepted a receiver method: %#v", names)
	} else if !strings.Contains(err.Error(), "receiver method bits") {
		t.Fatalf("collectSourceFunctionNames error = %q, want receiver method identity", err)
	}
	if code, err := convertFile(path, reg); err == nil {
		t.Fatalf("convertFile silently accepted a receiver method; generated code:\n%s", code)
	} else if !strings.Contains(err.Error(), "receiver method bits") {
		t.Fatalf("convertFile error = %q, want receiver method identity", err)
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

func TestRejectStaleGeneratedOwnershipMarkers(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "ok.rs"), []byte(genmarker.Line("go2rs from x.go")+"\n"), 0o644); err != nil {
		t.Fatalf("WriteFile ok.rs: %v", err)
	}
	if err := rejectStaleGeneratedOwnershipMarkers(dir); err != nil {
		t.Fatalf("rejectStaleGeneratedOwnershipMarkers ok dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(dir, "bad.rs"), []byte("// Code generated by tools/codegen rust-optimize; DO NOT EDIT.\n"), 0o644); err != nil {
		t.Fatalf("WriteFile bad.rs: %v", err)
	}
	if err := rejectStaleGeneratedOwnershipMarkers(dir); err == nil {
		t.Fatalf("rejectStaleGeneratedOwnershipMarkers accepted stale codegen marker")
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
			t.Fatalf("%s conversion retained template path:\n%s", name, code)
		}
		if err := rejectGeneratedFallbacks(name, code); err != nil {
			t.Fatalf("%s conversion contains fallback: %v", name, err)
		}
	}
}

// TestBid32StringParseRejectsAllocatingRender pins the failure direction of the
// bid32 parse stage. The stage no longer rewrites the generated parser, so its
// only job is to refuse a rendering that allocates on the parse path; a silent
// pass here would let a Go-side change reintroduce a per-call allocation with
// nothing failing.
func TestBid32StringParseRejectsAllocatingRender(t *testing.T) {
	allocatingRenders := map[string]string{
		"whole-string lowercase copy": `    let mut sl = String::from_utf8_lossy(&s).to_ascii_lowercase();
`,
		"owned byte buffer": `    let s2 = s.as_bytes().to_vec();
`,
		"owned string copy": `    let s3 = s.to_string();
`,
	}
	for name, injected := range allocatingRenders {
		input := `pub fn bid32_from_string_raw(ps: impl AsRef<str>, mut rnd_mode: i64) -> (u32, u32) {
    let ps = ps.as_ref();
    let mut s = (ps).trim_start_matches(|c| " \t".contains(c));
` + injected + `    return (0, 0);
}
`
		if _, err := optimizeRustStringHotpaths("bid32_string.go", input); err == nil {
			t.Fatalf("%s: expected the bid32 parse stage to reject an allocating render, got no error", name)
		}
	}
}

// TestBid32StringParseAcceptsBorrowedRender is the matching success direction:
// the shape the Go port now generates must pass untouched, so the stage stays a
// check rather than turning back into a rewrite.
func TestBid32StringParseAcceptsBorrowedRender(t *testing.T) {
	input := `pub fn bid32_from_string_raw(ps: impl AsRef<str>, mut rnd_mode: i64) -> (u32, u32) {
    let ps = ps.as_ref();
    let mut s = (ps).trim_start_matches(|c| " \t".contains(c));
    if ((s.len() as i64) == 0) {
        return (0x7c000000, 0);
    }
    let mut c = s.as_bytes()[0];
    if ((((c != b'.') && (c != b'-')) && (c != b'+')) && (((c < b'0') || (c > b'9')))) {
        if (equal_fold_ascii(s, "inf") || equal_fold_ascii(s, "infinity")) {
            return (0x78000000, 0);
        }
        if has_prefix_fold_ascii(s, "snan") {
            return (0x7e000000, 0);
        }
        return (0x7c000000, 0);
    }
    if ((s.len() as i64) > 1) {
        let mut sl1 = &s[1 as usize..];
        if equal_fold_ascii(sl1, "nan") {
            return (0x7c000000, 0);
        }
    }
    return (0, 0);
}
`
	got, err := optimizeRustStringHotpaths("bid32_string.go", input)
	if err != nil {
		t.Fatalf("optimizeRustStringHotpaths rejected the borrowed render: %v", err)
	}
	if got != input {
		t.Fatalf("bid32 parse stage rewrote the borrowed render; it must only check.\ngot:\n%s", got)
	}
}

// TestBid32StringParseRequiresParseFunction keeps the stage from degrading into
// a no-op if the parser is renamed or dropped from the generated file.
func TestBid32StringParseRequiresParseFunction(t *testing.T) {
	if _, err := optimizeRustStringHotpaths("bid32_string.go", "pub fn unrelated() {}\n"); err == nil {
		t.Fatal("expected an error when bid32_from_string_raw is absent, got none")
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
	for _, rejected := range []string{"(|| -> i32", "closure TODO", "(n1 as i32), rnd_mode"} {
		if strings.Contains(got, rejected) {
			t.Fatalf("optimizeBid128Misc left %q in output:\n%s", rejected, got)
		}
	}
	for _, required := range []string{"i32::MAX", "i32::MIN", "return bid128_scalbn(x, (n1 as i64), rnd_mode, pfpsf);"} {
		if !strings.Contains(got, required) {
			t.Fatalf("optimizeBid128Misc output missing %q:\n%s", required, got)
		}
	}
}

// TestStringParamIsReboundDetectsMutation pins the analysis that decides whether
// a string parameter is bound by reference or copied into an owned String. A
// false "not rebound" would hand the body a borrow where Go semantics call for
// an independent copy, so every mutation shape must be detected.
func TestStringParamIsReboundDetectsMutation(t *testing.T) {
	cases := []struct {
		name string
		body string
		want bool
	}{
		{"read only", `x := len(s); _ = x`, false},
		{"indexed read", `_ = s[0]`, false},
		{"sliced read", `_ = s[1:]`, false},
		{"range over it", `for i, v := range s { _, _ = i, v }`, false},
		{"direct assign", `s = "x"`, true},
		{"compound assign", `s += "x"`, true},
		{"tuple assign", `a := 0; a, s = 1, "x"; _ = a`, true},
		{"assign in if", `if true { s = "x" }`, true},
		{"assign in loop", `for i := 0; i < 2; i++ { s = "x" }`, true},
		{"assign inside closure", `f := func() { s = "x" }; f()`, true},
		{"assign inside defer", `defer func() { s = "x" }()`, true},
		{"address taken", `p := &s; _ = p`, true},
		{"range assigns into it", `for s = range map[string]int{} { }`, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			src := "package p\nfunc f(s string) {\n" + tc.body + "\n}\n"
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "p.go", src, 0)
			if err != nil {
				t.Fatalf("parse %q: %v", tc.body, err)
			}
			fn := file.Decls[0].(*ast.FuncDecl)
			if got := stringParamIsRebound(fn.Body, "s"); got != tc.want {
				t.Fatalf("stringParamIsRebound(%q) = %v, want %v", tc.body, got, tc.want)
			}
		})
	}
}

// TestStringParamIsReboundTreatsMissingBodyAsRebound keeps the analysis
// conservative when there is no body to inspect.
func TestStringParamIsReboundTreatsMissingBodyAsRebound(t *testing.T) {
	if !stringParamIsRebound(nil, "s") {
		t.Fatal("a nil body must be treated as rebound so the owned form is used")
	}
}

// TestIsGoStringExprFallbackRejectsNonStrings guards the type-info-absent path.
// Misreporting a []byte as a string would turn its mutable Rust slice borrow
// into a shared borrow, so the fallback must only accept known string params.
func TestIsGoStringExprFallbackRejectsNonStrings(t *testing.T) {
	prevInfo := activeTypeInfo
	prevVars := activeStringVars
	activeTypeInfo = nil
	activeStringVars = map[string]bool{"s": true}
	defer func() {
		activeTypeInfo = prevInfo
		activeStringVars = prevVars
	}()

	if !isGoStringExpr(&ast.Ident{Name: "s"}) {
		t.Fatal("a registered string parameter must be reported as a string")
	}
	if isGoStringExpr(&ast.Ident{Name: "buf"}) {
		t.Fatal("an unregistered identifier must not be reported as a string")
	}
	if isGoStringExpr(&ast.CallExpr{Fun: &ast.Ident{Name: "f"}}) {
		t.Fatal("a non-identifier expression must not be reported as a string without type info")
	}
}
