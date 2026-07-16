package testgen

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestEmitMixedRustPublicParityUsesBothOperandWidths(t *testing.T) {
	row := rustParityInventoryRow{
		GoSymbol:      "Add64DQBIDWithMode",
		RustOwner:     "Decimal64",
		RustSurface:   "add_dq_with_mode",
		Shape:         "mixed_binary_mode_flags_dq",
		BidgoFunction: "Bid64dqAdd",
	}
	var b strings.Builder
	_, cases, err := emitRustParityUnit(&b, row, publicParityCorpus{})
	if err != nil {
		t.Fatal(err)
	}
	if want := (len(parityLabelPairs) + 4) * len(publicParityModeOrderNames); cases != want {
		t.Fatalf("mixed Rust parity cases = %d, want %d", cases, want)
	}
	got := b.String()
	for _, want := range []string{
		"let left_pair = PAIRS_64[pair_index]",
		"let right_pair = PAIRS_128[pair_index]",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("mixed Rust parity output missing %q:\n%s", want, got)
		}
	}
	assertRustBoundCallArgs(t, got, "(pv, pf)", "Decimal64::add_dq_with_mode", [][]string{
		{"Decimal64::from_bits(left_bits)", "Decimal128::from_le_bytes(right_bits)", "mode"},
		{"Decimal64::from_bits(left_bits)", "Decimal128::from_le_bytes(right_bits)", "mode"},
	})
	assertRustBoundCallArgs(t, got, "(pr, praw)", "bid754::generated::bid128_add::bid64dq_add", [][]string{
		{"left_bits", "to_port128(right_bits)", "port_mode"},
		{"left_bits", "to_port128(right_bits)", "port_mode"},
	})
}

func TestEmitMixedRustPublicParityDoesNotInventModeDiscriminationForExactDDMul(t *testing.T) {
	row := rustParityInventoryRow{
		GoSymbol:      "Mul128DDBIDWithMode",
		RustOwner:     "Decimal128",
		RustSurface:   "mul_dd_with_mode",
		Shape:         "mixed_binary_mode_flags_dd",
		BidgoFunction: "Bid128ddMul",
	}
	var b strings.Builder
	_, cases, err := emitRustParityUnit(&b, row, publicParityCorpus{})
	if err != nil {
		t.Fatal(err)
	}
	if want := len(parityLabelPairs) * len(publicParityModeOrderNames); cases != want {
		t.Fatalf("exact DD multiplication cases = %d, want %d", cases, want)
	}
	if strings.Contains(b.String(), "mode_seen") {
		t.Fatalf("exact Decimal64-by-Decimal64 to Decimal128 multiplication must not claim a mode-discriminating corpus:\n%s", b.String())
	}
}

func TestEmitMixedFMARustPublicParityPreservesOrderedOperandWidths(t *testing.T) {
	row := rustParityInventoryRow{
		GoSymbol:      "FMA64DQQBIDWithMode",
		RustOwner:     "Decimal64",
		RustSurface:   "fma_dqq_with_mode",
		Shape:         "mixed_ternary_mode_flags_dqq",
		BidgoFunction: "Bid64dqqFma",
	}
	var b strings.Builder
	_, cases, err := emitRustParityUnit(&b, row, publicParityCorpus{})
	if err != nil {
		t.Fatal(err)
	}
	if want := (len(parityLabelTriples)+4)*len(publicParityModeOrderNames) + 1; cases != want {
		t.Fatalf("mixed FMA Rust parity cases = %d, want %d", cases, want)
	}
	got := b.String()
	for _, want := range []string{
		"let x_triple = TRIPLES_64[triple_index]",
		"let y_triple = TRIPLES_128[triple_index]",
		"let z_triple = TRIPLES_128[triple_index]",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("mixed FMA Rust parity output missing %q:\n%s", want, got)
		}
	}
	assertRustBoundCallArgs(t, got, "(pv, pf)", "Decimal64::fma_dqq_with_mode", [][]string{
		{"Decimal64::from_bits(x_bits)", "Decimal128::from_le_bytes(y_bits)", "Decimal128::from_le_bytes(z_bits)", "mode"},
		{"Decimal64::from_bits(x_bits)", "Decimal128::from_le_bytes(y_bits)", "Decimal128::from_le_bytes(z_bits)", "mode"},
	})
	assertRustBoundCallArgs(t, got, "(pr, praw)", "bid754::generated::bid128_fma::bid64dqq_fma", [][]string{
		{"x_bits", "to_port128(y_bits)", "to_port128(z_bits)", "port_mode"},
		{"x_bits", "to_port128(y_bits)", "to_port128(z_bits)", "port_mode"},
	})
	assertRustBoundCallArgs(t, got, "(fused_pv, fused_pf)", "Decimal64::fma_dqq_with_mode", [][]string{{
		"Decimal64::from_bits(fused_x_bits)", "Decimal128::from_le_bytes(fused_y_bits)", "Decimal128::from_le_bytes(fused_z_bits)", "fused_mode",
	}})
	assertRustBoundCallArgs(t, got, "(fused_pr, fused_praw)", "bid754::generated::bid128_fma::bid64dqq_fma", [][]string{{
		"fused_x_bits", "to_port128(fused_y_bits)", "to_port128(fused_z_bits)", "fused_port_mode",
	}})
	assertRustBoundCallArgs(t, got, "(fused_product, fused_mul_raw)", "bid754::generated::bid128_mul::bid128_mul", [][]string{{
		"fused_x_q", "fused_y_q", "fused_port_mode",
	}})
	assertRustBoundCallArgs(t, got, "fused_sum", "bid754::generated::bid128_add::bid128_add", [][]string{{
		"fused_product", "fused_z_q", "fused_port_mode", "&mut composed_raw",
	}})
	assertRustBoundCallArgs(t, got, "(composed_result, fused_narrow_raw)", "bid754::generated::bid128_conversions::bid128_to_bid64", [][]string{{
		"fused_sum", "fused_port_mode",
	}})
	for _, want := range []string{
		"let fused_x_bits: u64 = 0x31c0000000000003u64;",
		"let expected_bits: u64 = 0x2fe38d7ea4c68001u64;",
		"let forbidden_bits: u64 = 0x2fe38d7ea4c68000u64;",
		"let mut composed_raw = fused_x_widen_raw | fused_y_widen_raw | fused_z_widen_raw | fused_mul_raw;",
		"composed_raw |= fused_narrow_raw;",
		"if composed_result == fused_pr && composed_raw == fused_praw {",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("mixed FMA fusedness output missing %q:\n%s", want, got)
		}
	}
}

func TestRustFusednessBitsLiteralPreservesDecimal128RawLittleEndian(t *testing.T) {
	got, err := rustFusednessBitsLiteral(ffiFusednessQ(0x2ffca45894e48295, 0x7efb0aa216fc0001))
	if err != nil {
		t.Fatal(err)
	}
	want := "[0x01, 0x00, 0xfc, 0x16, 0xa2, 0x0a, 0xfb, 0x7e, 0x95, 0x82, 0xe4, 0x94, 0x58, 0xa4, 0xfc, 0x2f]"
	if got != want {
		t.Fatalf("Decimal128 fusedness Rust literal = %q, want raw LE %q", got, want)
	}
}

func TestEmitRustMixedFMAFusednessSentinelRowsPreservesCanonicalOrder(t *testing.T) {
	var b strings.Builder
	emitRustMixedFMAFusednessSentinelRows(&b)
	got := b.String()
	rows := ffiMixedFMAFusednessRows()
	header := fmt.Sprintf("const MIXED_FMA_FUSEDNESS_SENTINEL_ROWS: [&str; %d] = [", len(rows))
	if !strings.HasPrefix(got, header+"\n") {
		t.Fatalf("Rust fusedness sentinel table header missing: want %q\n%s", header, got)
	}
	prior := len(header)
	for i, row := range rows {
		needle := fmt.Sprintf("    %q,", row)
		at := strings.Index(got[prior:], needle)
		if at < 0 {
			t.Fatalf("Rust fusedness sentinel row %d missing or out of order: %q\n%s", i, row, got)
		}
		prior += at + len(needle)
	}
	if strings.Count(got, "\n    \"") != len(rows) {
		t.Fatalf("Rust fusedness sentinel table carries %d quoted rows, want %d\n%s", strings.Count(got, "\n    \""), len(rows), got)
	}
}

func TestGenerateRustPublicParityIncludesClosedMixedFMAFusednessCases(t *testing.T) {
	devtoolsRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	outputs, err := GenerateRustPublicParityOutputs(devtoolsRoot)
	if err != nil {
		t.Fatal(err)
	}
	raw, ok := outputs[rustPublicParityRunnerPath]
	if !ok {
		t.Fatalf("Rust public parity generator omitted %q", rustPublicParityRunnerPath)
	}
	source := string(raw)
	if got, want := strings.Count(source, " fusedness: public result mismatch"), len(ffiMixedFMAFusednessProbes); got != want {
		t.Fatalf("generated Rust public parity fusedness cases = %d, want closed census %d", got, want)
	}
	if got, want := strings.Count(source, "bid754::generated::bid128_mul::bid128_mul(fused_x_q, fused_y_q, fused_port_mode)"), len(ffiMixedFMAFusednessProbes); got != want {
		t.Fatalf("generated Rust fusedness sequential multiplications = %d, want %d", got, want)
	}
	if got, want := strings.Count(source, "bid754::generated::bid128_add::bid128_add(fused_product, fused_z_q, fused_port_mode, &mut composed_raw)"), len(ffiMixedFMAFusednessProbes); got != want {
		t.Fatalf("generated Rust fusedness sequential additions = %d, want %d", got, want)
	}
	if got, want := strings.Count(source, "bid754::generated::bid128_conversions::bid128_to_bid64(fused_sum, fused_port_mode)"), 7; got != want {
		t.Fatalf("generated Rust fusedness Decimal128-to-Decimal64 narrowings = %d, want %d", got, want)
	}
	for shape, wantCases := range map[string]int{
		"mixed_ternary_mode_flags_ddd": 91,
		"mixed_ternary_mode_flags_ddq": 182,
		"mixed_ternary_mode_flags_dqd": 182,
		"mixed_ternary_mode_flags_dqq": 182,
		"mixed_ternary_mode_flags_qdd": 182,
		"mixed_ternary_mode_flags_qdq": 182,
		"mixed_ternary_mode_flags_qqd": 182,
		"mixed_ternary_mode_flags_qqq": 91,
	} {
		row := fmt.Sprintf("(%q, %d),", shape, wantCases)
		if !strings.Contains(source, row) {
			t.Errorf("generated Rust parity case-count table missing %s", row)
		}
	}
	if problems := validateRustPublicParityFusednessStructure(source); len(problems) != 0 {
		t.Fatalf("generated Rust public parity fusedness structure is not closed:\n%s", strings.Join(problems, "\n"))
	}
}

func TestRustPublicParityFusednessStructureRejectsDeadOrDisconnectedDataflow(t *testing.T) {
	devtoolsRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	outputs, err := GenerateRustPublicParityOutputs(devtoolsRoot)
	if err != nil {
		t.Fatal(err)
	}
	raw, ok := outputs[rustPublicParityRunnerPath]
	if !ok {
		t.Fatalf("Rust public parity generator omitted %q", rustPublicParityRunnerPath)
	}
	source := string(raw)
	if problems := validateRustPublicParityFusednessStructure(source); len(problems) != 0 {
		t.Fatalf("baseline generated Rust public parity is invalid:\n%s", strings.Join(problems, "\n"))
	}

	publicCall := "let (fused_pv, fused_pf) = Decimal64::fma_ddq_with_mode(Decimal64::from_bits(fused_x_bits), Decimal64::from_bits(fused_y_bits), Decimal128::from_le_bytes(fused_z_bits), fused_mode);"
	deadPublicCall := "let (fused_pv, fused_pf) = if false { Decimal64::fma_ddq_with_mode(Decimal64::from_bits(fused_x_bits), Decimal64::from_bits(fused_y_bits), Decimal128::from_le_bytes(fused_z_bits), fused_mode) } else { (Decimal64::from_bits(expected_bits), bid754::ExceptionFlags::INEXACT) };"
	driverHeader := "#[test]\nfn generated_public_api_parity() {"
	driverAssertion := `    assert!(
        failures.is_empty(),
        "public API parity failures ({} total):\n{}",
        failures.len(),
        failures.join("\n")
    );`

	mutations := []struct {
		name        string
		old         string
		replacement string
	}{
		{
			name:        "dead fused public call decoy",
			old:         publicCall,
			replacement: deadPublicCall,
		},
		{
			name:        "public result self compare",
			old:         "if fused_pv.to_bits() != expected_bits {",
			replacement: "if fused_pv.to_bits() != fused_pv.to_bits() {",
		},
		{
			name:        "sequential result reuses forbidden outcome",
			old:         "let (composed_result, fused_narrow_raw) = bid754::generated::bid128_conversions::bid128_to_bid64(fused_sum, fused_port_mode);",
			replacement: "let (_sequential_result, fused_narrow_raw) = bid754::generated::bid128_conversions::bid128_to_bid64(fused_sum, fused_port_mode);\n        let composed_result = forbidden_bits;",
		},
		{
			name:        "shared failures sink shadowed",
			old:         "fn parity_fma64_ddqbidwith_mode(failures: &mut Vec<String>) -> usize {",
			replacement: "fn parity_fma64_ddqbidwith_mode(failures: &mut Vec<String>) -> usize {\n    let mut failures: Vec<String> = Vec::new();",
		},
		{
			name:        "failures cleared before final assertion",
			old:         driverAssertion,
			replacement: "    failures.clear();\n" + driverAssertion,
		},
		{
			name:        "final failures assertion removed",
			old:         driverAssertion,
			replacement: "    let _ = failures;",
		},
		{
			name:        "driver test attribute removed",
			old:         driverHeader,
			replacement: "fn generated_public_api_parity() {",
		},
		{
			name:        "driver test ignored",
			old:         driverHeader,
			replacement: "#[test]\n#[ignore]\nfn generated_public_api_parity() {",
		},
		{
			name:        "driver test disabled by cfg",
			old:         driverHeader,
			replacement: "#[cfg(any())]\n#[test]\nfn generated_public_api_parity() {",
		},
		{
			name:        "nested test attribute decoy",
			old:         driverHeader,
			replacement: "mod dead_test_decoy { #[test] fn generated_public_api_parity() {} }\nfn generated_public_api_parity() {",
		},
	}
	for _, mutation := range mutations {
		t.Run(mutation.name, func(t *testing.T) {
			if got := strings.Count(source, mutation.old); got == 0 {
				t.Fatalf("mutation anchor is missing: %q", mutation.old)
			}
			mutated := strings.Replace(source, mutation.old, mutation.replacement, 1)
			if problems := validateRustPublicParityFusednessStructure(mutated); len(problems) == 0 {
				t.Fatalf("structural verifier accepted mutation %q", mutation.name)
			}
		})
	}
	t.Run("integration test crate disabled by inner cfg", func(t *testing.T) {
		mutated := "#![cfg(any())]\n" + source
		if problems := validateRustPublicParityFusednessStructure(mutated); len(problems) == 0 {
			t.Fatal("structural verifier accepted a crate-level inner cfg that disables the entire integration test")
		}
	})

	row128DDQ := `ParityUnit { go_symbol: "FMA128DDQBIDWithMode", shape: "mixed_ternary_mode_flags_ddq", run: parity_fma128_ddqbidwith_mode },`
	row64DDQ := `ParityUnit { go_symbol: "FMA64DDQBIDWithMode", shape: "mixed_ternary_mode_flags_ddq", run: parity_fma64_ddqbidwith_mode },`
	mutateRegistryRows := func(t *testing.T, replacements ...[2]string) string {
		t.Helper()
		mutated := source
		for _, replacement := range replacements {
			if got := strings.Count(mutated, replacement[0]); got != 1 {
				t.Fatalf("registry mutation anchor count = %d, want 1 for %q", got, replacement[0])
			}
			mutated = strings.Replace(mutated, replacement[0], replacement[1], 1)
		}
		return mutated
	}
	assertRegistryMutationRejected := func(t *testing.T, mutated string) {
		t.Helper()
		if problems := validateRustPublicParityFusednessStructure(mutated); len(problems) == 0 {
			t.Fatal("structural verifier accepted disconnected mixed-FMA PARITY_UNITS registration")
		}
	}
	t.Run("same-shape helper registrations swapped", func(t *testing.T) {
		mutated := mutateRegistryRows(t,
			[2]string{row128DDQ, strings.Replace(row128DDQ, "parity_fma128_ddqbidwith_mode", "parity_fma64_ddqbidwith_mode", 1)},
			[2]string{row64DDQ, strings.Replace(row64DDQ, "parity_fma64_ddqbidwith_mode", "parity_fma128_ddqbidwith_mode", 1)},
		)
		assertRegistryMutationRejected(t, mutated)
	})
	t.Run("same-shape helper duplicated and peer left dead", func(t *testing.T) {
		mutated := mutateRegistryRows(t,
			[2]string{row128DDQ, strings.Replace(row128DDQ, "parity_fma128_ddqbidwith_mode", "parity_fma64_ddqbidwith_mode", 1)},
		)
		assertRegistryMutationRejected(t, mutated)
	})
	t.Run("equal-count mixed sqrt helper duplicated and peer left dead", func(t *testing.T) {
		row128D := `ParityUnit { go_symbol: "Sqrt128DBIDWithMode", shape: "mixed_unary_mode_flags_d", run: parity_sqrt128_dbidwith_mode },`
		mutated := mutateRegistryRows(t,
			[2]string{row128D, strings.Replace(row128D, "parity_sqrt128_dbidwith_mode", "parity_sqrt64_qbidwith_mode", 1)},
		)
		assertRegistryMutationRejected(t, mutated)
	})
	t.Run("dead helper row retained only as decoy outside registry", func(t *testing.T) {
		deadRow := strings.Replace(row128DDQ, "parity_fma128_ddqbidwith_mode", "parity_fma64_ddqbidwith_mode", 1)
		mutated := mutateRegistryRows(t, [2]string{row128DDQ, deadRow})
		mutated += "\nconst _DEAD_FMA_PARITY_UNIT: ParityUnit = " + row128DDQ + "\n"
		assertRegistryMutationRejected(t, mutated)
	})
	t.Run("exact helper cfg-disabled and replaced by same-shape alias", func(t *testing.T) {
		const helper = "fn parity_fma128_ddqbidwith_mode(failures: &mut Vec<String>) -> usize {"
		if got := strings.Count(source, helper); got != 1 {
			t.Fatalf("helper mutation anchor count = %d, want 1", got)
		}
		mutated := strings.Replace(source, helper, "#[cfg(any())]\n"+helper, 1)
		mutated += "\nuse self::parity_fma64_ddqbidwith_mode as parity_fma128_ddqbidwith_mode;\n"
		assertRegistryMutationRejected(t, mutated)
	})
	t.Run("exact registry cfg-disabled and replaced by active alias", func(t *testing.T) {
		const tableHeader = "const PARITY_UNITS: &[ParityUnit] = &["
		start := strings.Index(source, tableHeader)
		if start < 0 {
			t.Fatalf("registry table header is missing")
		}
		relEnd := strings.Index(source[start:], "\n];")
		if relEnd < 0 {
			t.Fatalf("registry table terminator is missing")
		}
		end := start + relEnd + len("\n];")
		table := source[start:end]
		active := strings.Replace(table, "const PARITY_UNITS", "const ACTIVE_PARITY_UNITS", 1)
		active = strings.Replace(active, row128DDQ, strings.Replace(row128DDQ, "parity_fma128_ddqbidwith_mode", "parity_fma64_ddqbidwith_mode", 1), 1)
		disabled := "#[cfg(any())]\n" + table
		mutated := source[:start] + disabled + source[end:] + "\n" + active + "\nuse self::ACTIVE_PARITY_UNITS as PARITY_UNITS;\n"
		assertRegistryMutationRejected(t, mutated)
	})
}

type rustParityStructureToken struct {
	kind string
	text string
}

func validateRustPublicParityFusednessStructure(source string) []string {
	tokens, err := lexRustParityStructure(source)
	if err != nil {
		return []string{fmt.Sprintf("lex generated Rust public parity: %v", err)}
	}

	var problems []string
	innerAttributes, err := rustParityModuleLevelInnerAttributeCount(tokens)
	if err != nil {
		problems = append(problems, fmt.Sprintf("module-level inner attributes: %v", err))
	} else if innerAttributes != 0 {
		problems = append(problems, fmt.Sprintf("module-level inner attribute count = %d, want 0; crate-wide cfg/ignore attributes can disable the integration test", innerAttributes))
	}
	if got, want := countRustParityTokenSequence(tokens, []string{"let", "fused_x_bits"}), len(ffiMixedFMAFusednessProbes); got != want {
		problems = append(problems, fmt.Sprintf("fusedness declaration census = %d, want %d", got, want))
	}
	seenFunctions := make(map[string]bool, len(ffiMixedFMAFusednessProbes))
	for _, probe := range ffiMixedFMAFusednessProbes {
		funcName, err := expectedRustFusednessFunctionName(probe.function)
		if err != nil {
			problems = append(problems, err.Error())
			continue
		}
		if seenFunctions[funcName] {
			problems = append(problems, fmt.Sprintf("duplicate expected fusedness function %q", funcName))
			continue
		}
		seenFunctions[funcName] = true
		function, err := rustParityFunctionStructureFor(tokens, funcName)
		if err != nil {
			problems = append(problems, err.Error())
			continue
		}
		if len(function.attributes) != 0 {
			problems = append(problems, fmt.Sprintf("%s: mixed-FMA helper declaration must have no attributes", funcName))
		}
		if want := mustLexRustParityStructure("failures: &mut Vec<String>"); !rustParityTokensEqual(function.params, want) {
			problems = append(problems, fmt.Sprintf("%s: function parameter differs from the exact shared failures sink", funcName))
		}
		if rustParityLetBindsName(function.body, "failures") {
			problems = append(problems, fmt.Sprintf("%s: function shadows the shared failures sink", funcName))
		}
		if countRustParityTokenSequence(function.body, []string{"failures", ".", "clear"}) != 0 {
			problems = append(problems, fmt.Sprintf("%s: function clears the shared failures sink", funcName))
		}
		body := function.body
		block, before, after, err := rustParityLiveFusednessBlock(body)
		if err != nil {
			problems = append(problems, fmt.Sprintf("%s: %v", funcName, err))
			continue
		}
		if len(before) == 0 || before[len(before)-1].text != "}" {
			problems = append(problems, fmt.Sprintf("%s: fusedness block is not a live anonymous top-level block after the generated corpus loops", funcName))
		}
		if !rustParityTokensEqual(after, mustLexRustParityStructure("count")) {
			problems = append(problems, fmt.Sprintf("%s: fusedness block must be the final live statement before the returned count", funcName))
		}
		wantBlock, err := expectedRustFusednessBlock(probe)
		if err != nil {
			problems = append(problems, fmt.Sprintf("%s: build expected fusedness block: %v", funcName, err))
			continue
		}
		if !rustParityTokensEqual(block, mustLexRustParityStructure(wantBlock)) {
			problems = append(problems, fmt.Sprintf("%s: fusedness block dataflow differs from the closed public-FMA/expected/sequential/forbidden contract", funcName))
		}
	}

	problems = append(problems, validateRustMixedParityUnitRegistry(tokens)...)

	driver, err := rustParityFunctionStructureFor(tokens, "generated_public_api_parity")
	if err != nil {
		problems = append(problems, err.Error())
	} else {
		if want := mustLexRustParityStructure("#[test]"); !rustParityTokensEqual(driver.attributes, want) {
			problems = append(problems, "generated_public_api_parity: module-level declaration must carry exactly one contiguous #[test] attribute")
		}
		if want := mustLexRustParityStructure(expectedRustPublicParityDriverBody()); !rustParityTokensEqual(driver.body, want) {
			problems = append(problems, "generated_public_api_parity: driver body differs from the closed failures-collection and final-assertion contract")
		}
	}
	return problems
}

type rustParityUnitRegistryRow struct {
	goSymbol string
	shape    string
	run      string
}

func validateRustMixedParityUnitRegistry(tokens []rustParityStructureToken) []string {
	actual, err := rustParityUnitRegistryRows(tokens)
	if err != nil {
		return []string{fmt.Sprintf("PARITY_UNITS: %v", err)}
	}
	expected, err := expectedRustMixedParityUnitRows()
	if err != nil {
		return []string{err.Error()}
	}

	var problems []string
	for _, want := range expected {
		exactRows := 0
		symbolRows := 0
		runRows := 0
		for _, got := range actual {
			if got.goSymbol == want.goSymbol {
				symbolRows++
			}
			if got.run == want.run {
				runRows++
			}
			if got == want {
				exactRows++
			}
		}
		if symbolRows != 1 {
			problems = append(problems, fmt.Sprintf("PARITY_UNITS: mixed public go_symbol %s row count = %d, want 1", want.goSymbol, symbolRows))
		}
		if runRows != 1 {
			problems = append(problems, fmt.Sprintf("PARITY_UNITS: mixed public run helper %s row count = %d, want 1", want.run, runRows))
		}
		if exactRows != 1 {
			problems = append(problems, fmt.Sprintf("PARITY_UNITS: exact mixed public row (%s, %s, %s) count = %d, want 1", want.goSymbol, want.shape, want.run, exactRows))
		}
		function, functionErr := rustParityFunctionStructureFor(tokens, want.run)
		if functionErr != nil {
			problems = append(problems, fmt.Sprintf("PARITY_UNITS: run helper %s: %v", want.run, functionErr))
			continue
		}
		if len(function.attributes) != 0 {
			problems = append(problems, fmt.Sprintf("PARITY_UNITS: run helper %s must have no attributes", want.run))
		}
		if params := mustLexRustParityStructure("failures: &mut Vec<String>"); !rustParityTokensEqual(function.params, params) {
			problems = append(problems, fmt.Sprintf("PARITY_UNITS: run helper %s parameter differs from the exact shared failures sink", want.run))
		}
	}
	return problems
}

func expectedRustMixedParityUnitRows() ([]rustParityUnitRegistryRow, error) {
	rows := make([]rustParityUnitRegistryRow, 0, len(ffiMixedFMAFusednessProbes)+2)
	seenSymbols := make(map[string]bool, len(ffiMixedFMAFusednessProbes)+2)
	seenRuns := make(map[string]bool, len(ffiMixedFMAFusednessProbes)+2)
	for _, probe := range ffiMixedFMAFusednessProbes {
		resultBits := 0
		for _, width := range []int{64, 128} {
			if strings.HasPrefix(probe.function, fmt.Sprintf("bid%d", width)) {
				resultBits = width
				break
			}
		}
		if resultBits == 0 || !strings.HasSuffix(probe.function, "_fma") {
			return nil, fmt.Errorf("mixed-FMA registry: invalid probe function %q", probe.function)
		}
		code := strings.TrimSuffix(strings.TrimPrefix(probe.function, fmt.Sprintf("bid%d", resultBits)), "_fma")
		if len(code) != 3 || strings.Trim(code, "dq") != "" {
			return nil, fmt.Errorf("mixed-FMA registry: invalid probe width code %q in %q", code, probe.function)
		}
		run, err := expectedRustFusednessFunctionName(probe.function)
		if err != nil {
			return nil, err
		}
		goSymbol := fmt.Sprintf("%q", fmt.Sprintf("FMA%d%sBIDWithMode", resultBits, strings.ToUpper(code)))
		row := rustParityUnitRegistryRow{
			goSymbol: goSymbol,
			shape:    fmt.Sprintf("%q", "mixed_ternary_mode_flags_"+code),
			run:      run,
		}
		if seenSymbols[row.goSymbol] || seenRuns[row.run] {
			return nil, fmt.Errorf("mixed-FMA registry: duplicate independent expected row %#v", row)
		}
		seenSymbols[row.goSymbol] = true
		seenRuns[row.run] = true
		rows = append(rows, row)
	}
	rows = append(rows,
		rustParityUnitRegistryRow{
			goSymbol: `"Sqrt128DBIDWithMode"`,
			shape:    `"mixed_unary_mode_flags_d"`,
			run:      "parity_sqrt128_dbidwith_mode",
		},
		rustParityUnitRegistryRow{
			goSymbol: `"Sqrt64QBIDWithMode"`,
			shape:    `"mixed_unary_mode_flags_q"`,
			run:      "parity_sqrt64_qbidwith_mode",
		},
	)
	return rows, nil
}

func rustParityUnitRegistryRows(tokens []rustParityStructureToken) ([]rustParityUnitRegistryRow, error) {
	indices, err := rustParityModuleLevelNamedTokenIndices(tokens, "const", "PARITY_UNITS")
	if err != nil {
		return nil, err
	}
	if len(indices) != 1 {
		return nil, fmt.Errorf("module-level const declaration count = %d, want 1", len(indices))
	}
	constIndex := indices[0]
	attributes, err := rustParityContiguousModuleAttributes(tokens, constIndex)
	if err != nil {
		return nil, fmt.Errorf("declaration attributes: %w", err)
	}
	if len(attributes) != 0 {
		return nil, fmt.Errorf("module-level const declaration must have no attributes")
	}
	equals := -1
	parenDepth, bracketDepth, braceDepth := 0, 0, 0
	for i := constIndex + 2; i < len(tokens); i++ {
		if parenDepth == 0 && bracketDepth == 0 && braceDepth == 0 {
			if tokens[i].text == "=" {
				equals = i
				break
			}
			if tokens[i].text == ";" {
				break
			}
		}
		updateRustParityDelimiterDepth(tokens[i].text, &parenDepth, &bracketDepth, &braceDepth)
	}
	if equals < 0 || equals+2 >= len(tokens) || tokens[equals+1].text != "&" || tokens[equals+2].text != "[" {
		return nil, fmt.Errorf("initializer is not the exact static &[] array form")
	}
	if want := mustLexRustParityStructure(": &[ParityUnit]"); !rustParityTokensEqual(tokens[constIndex+2:equals], want) {
		return nil, fmt.Errorf("declaration type is not exactly &[ParityUnit]")
	}
	arrayClose, err := matchingRustParityStructureToken(tokens, equals+2, "[", "]")
	if err != nil {
		return nil, err
	}
	if arrayClose+1 >= len(tokens) || tokens[arrayClose+1].text != ";" {
		return nil, fmt.Errorf("static array initializer has tokens after its closing bracket")
	}
	items, err := splitRustParityTopLevelCommaItems(tokens[equals+3 : arrayClose])
	if err != nil {
		return nil, err
	}
	rows := make([]rustParityUnitRegistryRow, 0, len(items))
	for i, item := range items {
		row, err := parseRustParityUnitRegistryRow(item)
		if err != nil {
			return nil, fmt.Errorf("row %d: %w", i, err)
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func parseRustParityUnitRegistryRow(tokens []rustParityStructureToken) (rustParityUnitRegistryRow, error) {
	wantText := []string{"ParityUnit", "{", "go_symbol", ":", "", ",", "shape", ":", "", ",", "run", ":", "", "}"}
	if len(tokens) != len(wantText) {
		return rustParityUnitRegistryRow{}, fmt.Errorf("token count = %d, want %d", len(tokens), len(wantText))
	}
	for i, want := range wantText {
		if i == 4 || i == 8 {
			if tokens[i].kind != "string" {
				return rustParityUnitRegistryRow{}, fmt.Errorf("token %d kind = %q, want string", i, tokens[i].kind)
			}
			continue
		}
		if i == 12 {
			if tokens[i].kind != "ident" {
				return rustParityUnitRegistryRow{}, fmt.Errorf("run token kind = %q, want ident", tokens[i].kind)
			}
			continue
		}
		if tokens[i].text != want {
			return rustParityUnitRegistryRow{}, fmt.Errorf("token %d = %q, want %q", i, tokens[i].text, want)
		}
	}
	return rustParityUnitRegistryRow{goSymbol: tokens[4].text, shape: tokens[8].text, run: tokens[12].text}, nil
}

func expectedRustFusednessFunctionName(function string) (string, error) {
	for _, width := range []string{"64", "128"} {
		prefix := "bid" + width
		if !strings.HasPrefix(function, prefix) || !strings.HasSuffix(function, "_fma") {
			continue
		}
		code := strings.TrimSuffix(strings.TrimPrefix(function, prefix), "_fma")
		if len(code) != 3 || strings.Trim(code, "dq") != "" {
			return "", fmt.Errorf("mixed FMA function %q has invalid width code %q", function, code)
		}
		return "parity_fma" + width + "_" + code + "bidwith_mode", nil
	}
	return "", fmt.Errorf("mixed FMA function %q has no supported result width", function)
}

func expectedRustFusednessBlock(probe ffiFusednessProbe) (string, error) {
	shape, ok := expectedFFIMixedDecimalShapes[probe.function]
	if !ok || shape.operation != "fma" || len(shape.operandBits) != 3 {
		return "", fmt.Errorf("no independent mixed-FMA shape for %q", probe.function)
	}
	prefix := fmt.Sprintf("bid%d", shape.resultBits)
	code := strings.TrimSuffix(strings.TrimPrefix(probe.function, prefix), "_fma")
	if len(code) != 3 {
		return "", fmt.Errorf("mixed-FMA function %q has invalid code %q", probe.function, code)
	}
	expectedType, expectedLiteral, err := expectedRustFusednessBits(probe.expected.bits)
	if err != nil {
		return "", err
	}
	_, forbiddenLiteral, err := expectedRustFusednessBits(probe.forbidden.bits)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	b.WriteString("{\n")
	for i, name := range []string{"fused_x", "fused_y", "fused_z"} {
		typeName, literal, err := expectedRustFusednessBits(probe.operands[i])
		if err != nil {
			return "", err
		}
		fmt.Fprintf(&b, "let %s_bits: %s = %s;\n", name, typeName, literal)
	}
	fmt.Fprintf(&b, "let expected_bits: %s = %s;\n", expectedType, expectedLiteral)
	fmt.Fprintf(&b, "let forbidden_bits: %s = %s;\n", expectedType, forbiddenLiteral)
	fmt.Fprintf(&b, "let expected_raw = 0x%08xu32;\n", probe.expected.flags)
	fmt.Fprintf(&b, "let forbidden_raw = 0x%08xu32;\n", probe.forbidden.flags)
	b.WriteString("let fused_mode = RoundingMode::NearestEven;\n")
	b.WriteString("let fused_port_mode = BIDGO_ROUND_NEAREST_EVEN;\n")

	publicArgs := make([]string, 3)
	portArgs := make([]string, 3)
	for i, name := range []string{"fused_x_bits", "fused_y_bits", "fused_z_bits"} {
		switch shape.operandBits[i] {
		case 64:
			publicArgs[i] = "Decimal64::from_bits(" + name + ")"
			portArgs[i] = name
		case 128:
			publicArgs[i] = "Decimal128::from_le_bytes(" + name + ")"
			portArgs[i] = "to_port128(" + name + ")"
		default:
			return "", fmt.Errorf("mixed-FMA function %q operand %d width = %d", probe.function, i, shape.operandBits[i])
		}
	}
	owner := fmt.Sprintf("Decimal%d", shape.resultBits)
	fmt.Fprintf(&b, "let (fused_pv, fused_pf) = %s::fma_%s_with_mode(%s, fused_mode);\n", owner, code, strings.Join(publicArgs, ", "))
	fmt.Fprintf(&b, "let (fused_pr, fused_praw) = bid754::generated::bid128_fma::%s(%s, fused_port_mode);\n", probe.function, strings.Join(portArgs, ", "))

	pubBits := "fused_pv.to_bits()"
	portBits := "fused_pr"
	composedBits := "composed_result"
	if shape.resultBits == 128 {
		pubBits = "fused_pv.to_le_bytes()"
		portBits = "from_port128(fused_pr)"
		composedBits = "from_port128(composed_result)"
	}
	fmt.Fprintf(&b, "if %s != expected_bits { failures.push(format!(\"\", %s, expected_bits)); }\n", pubBits, pubBits)
	b.WriteString("if fused_pf.bits() != map_port_flags(expected_raw) { failures.push(format!(\"\", fused_pf.bits(), map_port_flags(expected_raw))); }\n")
	fmt.Fprintf(&b, "if %s != expected_bits { failures.push(format!(\"\", %s, expected_bits)); }\n", portBits, portBits)
	b.WriteString("if fused_praw != expected_raw { failures.push(format!(\"\", fused_praw, expected_raw)); }\n")

	for i, name := range []string{"fused_x", "fused_y", "fused_z"} {
		if shape.operandBits[i] == 64 {
			fmt.Fprintf(&b, "let (%s_q, %s_widen_raw) = bid754::generated::to_bid12864::bid64_to_bid128(%s_bits);\n", name, name, name)
		} else {
			fmt.Fprintf(&b, "let %s_q = to_port128(%s_bits);\n", name, name)
			fmt.Fprintf(&b, "let %s_widen_raw = 0u32;\n", name)
		}
	}
	b.WriteString("let (fused_product, fused_mul_raw) = bid754::generated::bid128_mul::bid128_mul(fused_x_q, fused_y_q, fused_port_mode);\n")
	b.WriteString("let mut composed_raw = fused_x_widen_raw | fused_y_widen_raw | fused_z_widen_raw | fused_mul_raw;\n")
	b.WriteString("let fused_sum = bid754::generated::bid128_add::bid128_add(fused_product, fused_z_q, fused_port_mode, &mut composed_raw);\n")
	if shape.resultBits == 64 {
		b.WriteString("let (composed_result, fused_narrow_raw) = bid754::generated::bid128_conversions::bid128_to_bid64(fused_sum, fused_port_mode);\n")
		b.WriteString("composed_raw |= fused_narrow_raw;\n")
	} else {
		b.WriteString("let composed_result = fused_sum;\n")
	}
	fmt.Fprintf(&b, "if %s != forbidden_bits { failures.push(format!(\"\", %s, forbidden_bits)); }\n", composedBits, composedBits)
	b.WriteString("if composed_raw != forbidden_raw { failures.push(format!(\"\", composed_raw, forbidden_raw)); }\n")
	fmt.Fprintf(&b, "if %s == %s && composed_raw == fused_praw { failures.push(\"\".to_string()); }\n", composedBits, portBits)
	b.WriteString("count += 1;\n")
	b.WriteString("}\n")
	return b.String(), nil
}

func expectedRustFusednessBits(bits ffiFusednessBits) (string, string, error) {
	switch bits.width {
	case 64:
		return "u64", fmt.Sprintf("0x%016xu64", bits.lo), nil
	case 128:
		parts := make([]string, 16)
		for i := 0; i < 8; i++ {
			parts[i] = fmt.Sprintf("0x%02x", byte(bits.lo>>uint(8*i)))
			parts[8+i] = fmt.Sprintf("0x%02x", byte(bits.hi>>uint(8*i)))
		}
		return "[u8; 16]", "[" + strings.Join(parts, ", ") + "]", nil
	default:
		return "", "", fmt.Errorf("unsupported fusedness width %d", bits.width)
	}
}

func expectedRustPublicParityDriverBody() string {
	return `
assert_eq!(MIXED_FMA_FUSEDNESS_SENTINEL_ROWS.len(), EXPECTED_MIXED_FMA_FUSEDNESS_SENTINELS, "");
assert_eq!(PARITY_UNITS.len(), EXPECTED_PARITY_WRAPPERS, "");
let mut failures: Vec<String> = Vec::new();
let mut total = 0usize;
let mut by_shape: std::collections::HashMap<&str, usize> = std::collections::HashMap::new();
for unit in PARITY_UNITS {
    let n = (unit.run)(&mut failures);
    if n == 0 { failures.push(format!("", unit.go_symbol)); }
    total += n;
    *by_shape.entry(unit.shape).or_insert(0) += n;
}
if total != EXPECTED_PARITY_CASES {
    failures.push(format!("", EXPECTED_PARITY_CASES, total));
}
for &(shape, want) in EXPECTED_PARITY_CASES_BY_SHAPE {
    let got = by_shape.get(shape).copied().unwrap_or(0);
    if got != want { failures.push(format!("", shape, want, got)); }
}
if by_shape.len() != EXPECTED_PARITY_CASES_BY_SHAPE.len() {
    failures.push(format!("", by_shape.len(), EXPECTED_PARITY_CASES_BY_SHAPE.len()));
}
assert!(failures.is_empty(), "", failures.len(), failures.join(""));
`
}

func lexRustParityStructure(source string) ([]rustParityStructureToken, error) {
	var tokens []rustParityStructureToken
	for i := 0; i < len(source); {
		if source[i] == ' ' || source[i] == '\t' || source[i] == '\r' || source[i] == '\n' {
			i++
			continue
		}
		if i+1 < len(source) && source[i:i+2] == "//" {
			i += 2
			for i < len(source) && source[i] != '\n' {
				i++
			}
			continue
		}
		if i+1 < len(source) && source[i:i+2] == "/*" {
			depth := 1
			i += 2
			for i < len(source) && depth > 0 {
				switch {
				case i+1 < len(source) && source[i:i+2] == "/*":
					depth++
					i += 2
				case i+1 < len(source) && source[i:i+2] == "*/":
					depth--
					i += 2
				default:
					i++
				}
			}
			if depth != 0 {
				return nil, fmt.Errorf("unterminated block comment")
			}
			continue
		}
		if end, ok, err := scanRustParityRawString(source, i); err != nil {
			return nil, err
		} else if ok {
			tokens = append(tokens, rustParityStructureToken{kind: "string", text: source[i:end]})
			i = end
			continue
		}
		if source[i] == '"' {
			start := i
			i++
			closed := false
			for i < len(source) {
				if source[i] == '\\' {
					i += 2
					continue
				}
				if source[i] == '"' {
					i++
					closed = true
					break
				}
				i++
			}
			if !closed {
				return nil, fmt.Errorf("unterminated string literal at byte %d", start)
			}
			tokens = append(tokens, rustParityStructureToken{kind: "string", text: source[start:i]})
			continue
		}
		if source[i] == '\'' {
			if end, ok := scanRustParityCharLiteral(source, i); ok {
				tokens = append(tokens, rustParityStructureToken{kind: "char", text: source[i:end]})
				i = end
				continue
			}
		}
		if isRustParityIdentStart(source[i]) {
			start := i
			i++
			for i < len(source) && isRustParityIdentContinue(source[i]) {
				i++
			}
			tokens = append(tokens, rustParityStructureToken{kind: "ident", text: source[start:i]})
			continue
		}
		if source[i] >= '0' && source[i] <= '9' {
			start := i
			i++
			for i < len(source) && (isRustParityIdentContinue(source[i]) || source[i] == '.') {
				i++
			}
			tokens = append(tokens, rustParityStructureToken{kind: "number", text: source[start:i]})
			continue
		}
		tokens = append(tokens, rustParityStructureToken{kind: "punct", text: source[i : i+1]})
		i++
	}
	return tokens, nil
}

func scanRustParityRawString(source string, start int) (int, bool, error) {
	prefixLen := 0
	switch {
	case strings.HasPrefix(source[start:], "br"):
		prefixLen = 2
	case source[start] == 'r':
		prefixLen = 1
	default:
		return 0, false, nil
	}
	i := start + prefixLen
	hashes := 0
	for i < len(source) && source[i] == '#' {
		hashes++
		i++
	}
	if i >= len(source) || source[i] != '"' {
		return 0, false, nil
	}
	closing := "\"" + strings.Repeat("#", hashes)
	rel := strings.Index(source[i+1:], closing)
	if rel < 0 {
		return 0, false, fmt.Errorf("unterminated raw string literal at byte %d", start)
	}
	return i + 1 + rel + len(closing), true, nil
}

func scanRustParityCharLiteral(source string, start int) (int, bool) {
	i := start + 1
	if i >= len(source) || source[i] == '\n' || source[i] == '\r' {
		return 0, false
	}
	if source[i] == '\\' {
		i++
		if i >= len(source) {
			return 0, false
		}
		switch source[i] {
		case 'u':
			i++
			if i >= len(source) || source[i] != '{' {
				return 0, false
			}
			i++
			for i < len(source) && source[i] != '}' {
				i++
			}
			if i >= len(source) {
				return 0, false
			}
			i++
		case 'x':
			i += 3
		default:
			i++
		}
	} else {
		_, size := utf8.DecodeRuneInString(source[i:])
		if size == 0 {
			return 0, false
		}
		i += size
	}
	if i >= len(source) || source[i] != '\'' {
		return 0, false
	}
	return i + 1, true
}

func isRustParityIdentStart(ch byte) bool {
	return ch == '_' || ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z'
}

func isRustParityIdentContinue(ch byte) bool {
	return isRustParityIdentStart(ch) || ch >= '0' && ch <= '9'
}

func mustLexRustParityStructure(source string) []rustParityStructureToken {
	tokens, err := lexRustParityStructure(source)
	if err != nil {
		panic(err)
	}
	return tokens
}

type rustParityFunctionStructure struct {
	attributes []rustParityStructureToken
	params     []rustParityStructureToken
	body       []rustParityStructureToken
}

func rustParityFunctionStructureFor(tokens []rustParityStructureToken, name string) (rustParityFunctionStructure, error) {
	var functions []rustParityFunctionStructure
	parenDepth, bracketDepth, braceDepth := 0, 0, 0
	for i := 0; i < len(tokens); i++ {
		isTarget := parenDepth == 0 && bracketDepth == 0 && braceDepth == 0 &&
			i+2 < len(tokens) &&
			tokens[i].kind == "ident" && tokens[i].text == "fn" &&
			tokens[i+1].kind == "ident" && tokens[i+1].text == name && tokens[i+2].text == "("
		if !isTarget {
			updateRustParityDelimiterDepth(tokens[i].text, &parenDepth, &bracketDepth, &braceDepth)
			continue
		}
		attributes, err := rustParityContiguousModuleAttributes(tokens, i)
		if err != nil {
			return rustParityFunctionStructure{}, fmt.Errorf("Rust function %s attributes: %w", name, err)
		}
		paramsClose, err := matchingRustParityStructureToken(tokens, i+2, "(", ")")
		if err != nil {
			return rustParityFunctionStructure{}, fmt.Errorf("Rust function %s parameters: %w", name, err)
		}
		bodyOpen := -1
		for j := paramsClose + 1; j < len(tokens); j++ {
			if tokens[j].text == "{" {
				bodyOpen = j
				break
			}
			if tokens[j].text == ";" {
				break
			}
		}
		if bodyOpen < 0 {
			return rustParityFunctionStructure{}, fmt.Errorf("Rust function %s has no body", name)
		}
		bodyClose, err := matchingRustParityStructureToken(tokens, bodyOpen, "{", "}")
		if err != nil {
			return rustParityFunctionStructure{}, fmt.Errorf("Rust function %s body: %w", name, err)
		}
		functions = append(functions, rustParityFunctionStructure{
			attributes: attributes,
			params:     tokens[i+3 : paramsClose],
			body:       tokens[bodyOpen+1 : bodyClose],
		})
		i = bodyClose
	}
	if parenDepth != 0 || bracketDepth != 0 || braceDepth != 0 {
		return rustParityFunctionStructure{}, fmt.Errorf("Rust module delimiters are unbalanced while finding function %s", name)
	}
	if len(functions) != 1 {
		return rustParityFunctionStructure{}, fmt.Errorf("Rust function %s body count = %d, want 1", name, len(functions))
	}
	return functions[0], nil
}

func rustParityContiguousModuleAttributes(tokens []rustParityStructureToken, itemIndex int) ([]rustParityStructureToken, error) {
	start := itemIndex
	for start > 0 && tokens[start-1].text == "]" {
		open, err := matchingRustParityStructureTokenBackward(tokens, start-1, "[", "]")
		if err != nil {
			return nil, err
		}
		if open == 0 || tokens[open-1].text != "#" {
			break
		}
		start = open - 1
	}
	return tokens[start:itemIndex], nil
}

func matchingRustParityStructureTokenBackward(tokens []rustParityStructureToken, close int, left, right string) (int, error) {
	if close < 0 || close >= len(tokens) || tokens[close].text != right {
		return 0, fmt.Errorf("backward delimiter scan starts at %d, want %q", close, right)
	}
	depth := 0
	for i := close; i >= 0; i-- {
		switch tokens[i].text {
		case right:
			depth++
		case left:
			depth--
			if depth == 0 {
				return i, nil
			}
		}
	}
	return 0, fmt.Errorf("unbalanced backward %s%s delimiters", left, right)
}

func rustParityModuleLevelNamedTokenIndices(tokens []rustParityStructureToken, keyword, name string) ([]int, error) {
	var indices []int
	parenDepth, bracketDepth, braceDepth := 0, 0, 0
	for i := 0; i < len(tokens); i++ {
		if parenDepth == 0 && bracketDepth == 0 && braceDepth == 0 &&
			tokens[i].kind == "ident" && tokens[i].text == keyword &&
			i+1 < len(tokens) && tokens[i+1].kind == "ident" && tokens[i+1].text == name {
			indices = append(indices, i)
		}
		updateRustParityDelimiterDepth(tokens[i].text, &parenDepth, &bracketDepth, &braceDepth)
		if parenDepth < 0 || bracketDepth < 0 || braceDepth < 0 {
			return nil, fmt.Errorf("module delimiter depth became negative at token %d", i)
		}
	}
	if parenDepth != 0 || bracketDepth != 0 || braceDepth != 0 {
		return nil, fmt.Errorf("module delimiters are unbalanced")
	}
	return indices, nil
}

func rustParityModuleLevelInnerAttributeCount(tokens []rustParityStructureToken) (int, error) {
	count := 0
	parenDepth, bracketDepth, braceDepth := 0, 0, 0
	for i := 0; i < len(tokens); i++ {
		if parenDepth == 0 && bracketDepth == 0 && braceDepth == 0 &&
			tokens[i].text == "#" && i+2 < len(tokens) && tokens[i+1].text == "!" && tokens[i+2].text == "[" {
			close, err := matchingRustParityStructureToken(tokens, i+2, "[", "]")
			if err != nil {
				return 0, err
			}
			count++
			i = close
			continue
		}
		updateRustParityDelimiterDepth(tokens[i].text, &parenDepth, &bracketDepth, &braceDepth)
		if parenDepth < 0 || bracketDepth < 0 || braceDepth < 0 {
			return 0, fmt.Errorf("module delimiter depth became negative at token %d", i)
		}
	}
	if parenDepth != 0 || bracketDepth != 0 || braceDepth != 0 {
		return 0, fmt.Errorf("module delimiters are unbalanced")
	}
	return count, nil
}

func updateRustParityDelimiterDepth(text string, parenDepth, bracketDepth, braceDepth *int) {
	switch text {
	case "(":
		*parenDepth++
	case ")":
		*parenDepth--
	case "[":
		*bracketDepth++
	case "]":
		*bracketDepth--
	case "{":
		*braceDepth++
	case "}":
		*braceDepth--
	}
}

func splitRustParityTopLevelCommaItems(tokens []rustParityStructureToken) ([][]rustParityStructureToken, error) {
	var items [][]rustParityStructureToken
	start := 0
	parenDepth, bracketDepth, braceDepth := 0, 0, 0
	for i, token := range tokens {
		if token.text == "," && parenDepth == 0 && bracketDepth == 0 && braceDepth == 0 {
			if i == start {
				return nil, fmt.Errorf("empty static array item at token %d", i)
			}
			items = append(items, tokens[start:i])
			start = i + 1
			continue
		}
		updateRustParityDelimiterDepth(token.text, &parenDepth, &bracketDepth, &braceDepth)
		if parenDepth < 0 || bracketDepth < 0 || braceDepth < 0 {
			return nil, fmt.Errorf("array item delimiter depth became negative at token %d", i)
		}
	}
	if parenDepth != 0 || bracketDepth != 0 || braceDepth != 0 {
		return nil, fmt.Errorf("array item delimiters are unbalanced")
	}
	if start < len(tokens) {
		items = append(items, tokens[start:])
	}
	return items, nil
}

func rustParityFunctionBody(tokens []rustParityStructureToken, name string) ([]rustParityStructureToken, error) {
	function, err := rustParityFunctionStructureFor(tokens, name)
	if err != nil {
		return nil, err
	}
	return function.body, nil
}

func rustParityLetBindsName(tokens []rustParityStructureToken, name string) bool {
	for i := 0; i < len(tokens); i++ {
		if tokens[i].kind != "ident" || tokens[i].text != "let" {
			continue
		}
		for j := i + 1; j < len(tokens) && tokens[j].text != "=" && tokens[j].text != ";"; j++ {
			if tokens[j].kind == "ident" && tokens[j].text == name {
				return true
			}
		}
	}
	return false
}

func rustParityLiveFusednessBlock(body []rustParityStructureToken) (block, before, after []rustParityStructureToken, err error) {
	type span struct{ open, close int }
	var candidates []span
	for i := 0; i < len(body); {
		if body[i].text != "{" {
			i++
			continue
		}
		close, matchErr := matchingRustParityStructureToken(body, i, "{", "}")
		if matchErr != nil {
			return nil, nil, nil, matchErr
		}
		if countRustParityTokenSequence(body[i+1:close], []string{"let", "fused_x_bits"}) != 0 {
			candidates = append(candidates, span{open: i, close: close})
		}
		i = close + 1
	}
	if len(candidates) != 1 {
		return nil, nil, nil, fmt.Errorf("live top-level fusedness block count = %d, want 1", len(candidates))
	}
	selected := candidates[0]
	return body[selected.open : selected.close+1], body[:selected.open], body[selected.close+1:], nil
}

func matchingRustParityStructureToken(tokens []rustParityStructureToken, open int, left, right string) (int, error) {
	if open < 0 || open >= len(tokens) || tokens[open].text != left {
		return 0, fmt.Errorf("delimiter scan starts at %d, want %q", open, left)
	}
	depth := 0
	for i := open; i < len(tokens); i++ {
		switch tokens[i].text {
		case left:
			depth++
		case right:
			depth--
			if depth == 0 {
				return i, nil
			}
		}
	}
	return 0, fmt.Errorf("unbalanced %s%s delimiters", left, right)
}

func countRustParityTokenSequence(tokens []rustParityStructureToken, sequence []string) int {
	count := 0
	for i := 0; i+len(sequence) <= len(tokens); i++ {
		match := true
		for j := range sequence {
			if tokens[i+j].text != sequence[j] {
				match = false
				break
			}
		}
		if match {
			count++
		}
	}
	return count
}

func rustParityTokensEqual(got, want []rustParityStructureToken) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i].kind != want[i].kind {
			return false
		}
		if got[i].kind != "string" && got[i].text != want[i].text {
			return false
		}
	}
	return true
}

func TestEmitMixedSqrtRustPublicParityUsesSourceAndResultWidths(t *testing.T) {
	row := rustParityInventoryRow{
		GoSymbol:      "Sqrt128DBIDWithMode",
		RustOwner:     "Decimal128",
		RustSurface:   "sqrt_d_with_mode",
		Shape:         "mixed_unary_mode_flags_d",
		BidgoFunction: "Bid128dSqrt",
	}
	corpus := publicParityCorpus{Bits64: make([]uint64, publicParityCorpusLen)}
	var b strings.Builder
	_, cases, err := emitRustParityUnit(&b, row, corpus)
	if err != nil {
		t.Fatal(err)
	}
	if want := (publicParityCorpusLen + 4) * len(publicParityModeOrderNames); cases != want {
		t.Fatalf("mixed sqrt Rust parity cases = %d, want %d", cases, want)
	}
	got := b.String()
	for _, want := range []string{
		"for &value_bits in CORPUS_64",
		"mode_seen[mi] = pv.to_le_bytes()",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("mixed sqrt Rust parity output missing %q:\n%s", want, got)
		}
	}
	assertRustBoundCallArgs(t, got, "(pv, pf)", "Decimal128::sqrt_d_with_mode", [][]string{
		{"Decimal64::from_bits(value_bits)", "mode"},
		{"Decimal64::from_bits(value_bits)", "mode"},
	})
	assertRustBoundCallArgs(t, got, "(pr, praw)", "bid754::generated::bid128_sqrt::bid128d_sqrt", [][]string{
		{"value_bits", "port_mode"},
		{"value_bits", "port_mode"},
	})
}

// rustBoundCallArgs extracts only calls that are the complete right-hand side
// of a generated `let BINDING = CALLEE(...);` statement. This avoids the old
// substring check accepting an unrelated/commented call, while the balanced
// delimiter scan preserves nested conversion calls and their operand order.
func rustBoundCallArgs(source, binding, callee string) ([][]string, error) {
	if strings.Contains(source, "//") || strings.Contains(source, "/*") {
		return nil, fmt.Errorf("generated Rust function contains comments; structural call parser must be extended before accepting them")
	}
	needle := "let " + binding + " = " + callee + "("
	var calls [][]string
	for searchFrom := 0; ; {
		rel := strings.Index(source[searchFrom:], needle)
		if rel < 0 {
			break
		}
		start := searchFrom + rel
		lineStart := strings.LastIndex(source[:start], "\n") + 1
		if strings.TrimSpace(source[lineStart:start]) != "" {
			searchFrom = start + len(needle)
			continue
		}

		open := start + len(needle) - 1
		close, err := matchingRustDelimiter(source, open, '(', ')')
		if err != nil {
			return nil, fmt.Errorf("%s bound to %s: %w", binding, callee, err)
		}
		lineEnd := strings.IndexByte(source[close+1:], '\n')
		if lineEnd < 0 {
			lineEnd = len(source)
		} else {
			lineEnd += close + 1
		}
		if strings.TrimSpace(source[close+1:lineEnd]) != ";" {
			searchFrom = close + 1
			continue
		}
		args, err := splitTopLevelRustArgs(source[open+1 : close])
		if err != nil {
			return nil, fmt.Errorf("%s bound to %s: %w", binding, callee, err)
		}
		calls = append(calls, args)
		searchFrom = close + 1
	}
	return calls, nil
}

func matchingRustDelimiter(source string, open int, left, right byte) (int, error) {
	depth := 0
	for i := open; i < len(source); i++ {
		switch source[i] {
		case left:
			depth++
		case right:
			depth--
			if depth == 0 {
				return i, nil
			}
		}
	}
	return 0, fmt.Errorf("unbalanced %c%c delimiters", left, right)
}

func splitTopLevelRustArgs(source string) ([]string, error) {
	if strings.TrimSpace(source) == "" {
		return nil, nil
	}
	var args []string
	start := 0
	paren, bracket, brace := 0, 0, 0
	for i := 0; i < len(source); i++ {
		switch source[i] {
		case '(':
			paren++
		case ')':
			paren--
		case '[':
			bracket++
		case ']':
			bracket--
		case '{':
			brace++
		case '}':
			brace--
		case ',':
			if paren == 0 && bracket == 0 && brace == 0 {
				args = append(args, strings.TrimSpace(source[start:i]))
				start = i + 1
			}
		}
		if paren < 0 || bracket < 0 || brace < 0 {
			return nil, fmt.Errorf("unbalanced nested delimiters in %q", source)
		}
	}
	if paren != 0 || bracket != 0 || brace != 0 {
		return nil, fmt.Errorf("unbalanced nested delimiters in %q", source)
	}
	args = append(args, strings.TrimSpace(source[start:]))
	return args, nil
}

func assertRustBoundCallArgs(t *testing.T, source, binding, callee string, want [][]string) {
	t.Helper()
	if total := strings.Count(source, "let "+binding+" = "); total != len(want) {
		t.Fatalf("Rust binding %s call count = %d, want %d; every bound call must use %s\n%s", binding, total, len(want), callee, source)
	}
	got, err := rustBoundCallArgs(source, binding, callee)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(want) {
		t.Fatalf("Rust call %s bound to %s count = %d, want %d; got args %v\n%s", callee, binding, len(got), len(want), got, source)
	}
	for i := range want {
		if len(got[i]) != len(want[i]) {
			t.Errorf("Rust call %s[%d] arg count = %d, want %d; got %v", callee, i, len(got[i]), len(want[i]), got[i])
			continue
		}
		for j := range want[i] {
			if got[i][j] != want[i][j] {
				t.Errorf("Rust call %s[%d] arg %d = %q, want %q", callee, i, j, got[i][j], want[i][j])
			}
		}
	}
}

func TestEmitMixedRustPublicParityRejectsMiswiredManifestRows(t *testing.T) {
	tests := []rustParityInventoryRow{
		{
			GoSymbol: "FMA64DQQBIDWithMode", RustOwner: "Decimal64", RustSurface: "fma_qdq_with_mode",
			Shape: "mixed_ternary_mode_flags_dqq", BidgoFunction: "Bid64dqqFma",
		},
		{
			GoSymbol: "Sqrt128DBIDWithMode", RustOwner: "Decimal128", RustSurface: "sqrt_d_with_mode",
			Shape: "mixed_unary_mode_flags_d", BidgoFunction: "Bid64qSqrt",
		},
	}
	for _, row := range tests {
		var b strings.Builder
		if _, _, err := emitRustParityUnit(&b, row, publicParityCorpus{}); err == nil {
			t.Errorf("emitRustParityUnit(%s) accepted a miswired manifest row", row.GoSymbol)
		}
	}
}
