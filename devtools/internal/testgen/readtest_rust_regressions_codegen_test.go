package testgen

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestGenerateReadtestRustRegressionsFromSharedSpec(t *testing.T) {
	spec, err := LoadGenerated(filepath.Join("..", "..", "generated", "testspec", "spec_index.json"))
	if err != nil {
		t.Fatal(err)
	}
	data, err := GenerateReadtestRustRegressions(spec)
	if err != nil {
		t.Fatal(err)
	}
	src := string(data)
	anchors := loadVerificationAnchors(t)
	if got := strings.Count(src, "assert_eq!((got, flags)"); got != anchors.RustIEEEReadtestRegressionsTotal {
		t.Fatalf("generated Rust regression rows=%d, anchor=%d", got, anchors.RustIEEEReadtestRegressionsTotal)
	}
	selected := map[string]int{}
	divergences := 0
	for _, tc := range spec.ReadCases {
		if !strings.HasSuffix(tc.Group, "_ieee754_regressions") || tc.Kind == "from_string" || tc.Kind == "to_string" {
			continue
		}
		selected[tc.Function]++
		if tc.NativeCompareSkipReason != "" {
			divergences++
		}
	}
	if !reflect.DeepEqual(selected, anchors.RustIEEEReadtestRegressionsByFunction) || divergences != anchors.RustIEEEReadtestRegressionsNativeDivergences {
		t.Fatalf("Rust regression selection=%v divergences=%d, anchored=%v/%d", selected, divergences, anchors.RustIEEEReadtestRegressionsByFunction, anchors.RustIEEEReadtestRegressionsNativeDivergences)
	}
	for fn, path := range map[string]string{"bid128_fma": "bid128_fma::bid128_fma", "bid128_scalbn": "bid128_misc::bid128_scalbn", "bid128_scalbln": "bid128_misc::bid128_scalbln", "bid128_ldexp": "bid128_ldexp::bid128_ldexp", "bid32_nextafter": "bid32_next::bid32_next_after", "bid32_quantum": "bid32_exports::bid32_quantum"} {
		if got := strings.Count(src, "bid754::generated::"+path+"("); got != selected[fn] {
			t.Errorf("Rust %s rows=%d, selected=%d", fn, got, selected[fn])
		}
	}
	for _, fn := range []string{"bid128_fma::bid128_fma", "bid128_misc::bid128_scalbn", "bid128_misc::bid128_scalbln", "bid128_ldexp::bid128_ldexp", "bid32_next::bid32_next_after", "bid32_exports::bid32_quantum"} {
		if !strings.Contains(src, fn) {
			t.Errorf("missing generated function %s", fn)
		}
	}
	if !strings.Contains(src, "numeric_boundary_bid128_scalbn_threshold_cdiverge") {
		t.Error("missing C-divergent expected-result case")
	}
	if strings.Contains(src, "bid32_from_string") {
		t.Error("string conversion leaked into numeric runner")
	}
	checkedIn, err := os.ReadFile(filepath.Join("..", "..", readtestRustRegressionsPath))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, checkedIn) {
		t.Error("checked-in Rust regression runner differs from shared-spec generation")
	}
}

func TestGenerateReadtestRustRegressionsRejectsUnknownShape(t *testing.T) {
	base := GeneratedReadCase{
		ID: "future_case", Group: "decimal32_ieee754_regressions", Format: "decimal32",
		Kind: "binary_op", Function: "bid32_nextafter", CompareGroup: "CMP_FUZZYSTATUS",
		OutputType: "OP_DEC32", InputTypes: []string{"OP_DEC32", "OP_DEC32"},
		Operands: []string{"[00000000]", "[00000001]"}, Expected: "[00000001]", Status: "00",
	}
	for _, tc := range []GeneratedReadCase{
		func() GeneratedReadCase { c := base; c.Function = "bid32_future"; return c }(),
		func() GeneratedReadCase { c := base; c.InputTypes = []string{"OP_DEC32", "OP_INT32"}; return c }(),
		func() GeneratedReadCase { c := base; c.Status = "40"; return c }(),
		func() GeneratedReadCase { c := base; c.Function = "bid32_quantum"; return c }(),
		func() GeneratedReadCase { c := base; c.Function = "bid32_quantum"; c.Kind = "unary_op"; c.Operands = []string{"[60000000]"}; c.InputTypes = []string{"OP_DEC32"}; c.Rounding = 1; return c }(),
	} {
		if _, err := GenerateReadtestRustRegressions(SharedSpec{ReadCases: []GeneratedReadCase{tc}}); err == nil {
			t.Errorf("accepted unsupported regression shape %+v", tc)
		}
	}
}
