package testgen

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/sky1core/bid754/devtools/internal/genmarker"
)

const readtestRustRegressionsPath = "../bid754-rs/tests/readtest_regressions_generated.rs"

func WriteReadtestRustRegressionsOutput(repoRoot string, spec SharedSpec) error {
	data, err := GenerateReadtestRustRegressions(spec)
	if err != nil {
		return err
	}
	path := filepath.Join(repoRoot, readtestRustRegressionsPath)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write Rust readtest regressions %q: %w", path, err)
	}
	return nil
}

func GenerateReadtestRustRegressions(spec SharedSpec) ([]byte, error) {
	var b strings.Builder
	b.WriteString(genmarker.Line("testgen"))
	b.WriteString("\n#![cfg(feature = \"verification\")]\n\nuse bid754::gen_types::BID_UINT128;\n\n#[test]\nfn readtest_ieee754_regressions() {\n")
	count := 0
	for _, tc := range spec.ReadCases {
		if !isReadtestIEEERegressionGroup(tc.Group) || tc.Kind == "from_string" || tc.Kind == "to_string" {
			continue
		}
		if err := emitRustReadtestRegression(&b, tc); err != nil {
			return nil, fmt.Errorf("readtest regression %s: %w", tc.ID, err)
		}
		count++
	}
	if count == 0 {
		return nil, fmt.Errorf("no non-string IEEE readtest regressions selected")
	}
	b.WriteString("}\n")
	return []byte(b.String()), nil
}

func isReadtestIEEERegressionGroup(group string) bool {
	switch group {
	case "decimal32_ieee754_regressions", "decimal64_ieee754_regressions", "decimal128_ieee754_regressions":
		return true
	default:
		return false
	}
}

func emitRustReadtestRegression(b *strings.Builder, tc GeneratedReadCase) error {
	if tc.InitialStatus&^uint32(0x3d) != 0 || (tc.InitialStatus != 0 &&
		(tc.Function != "bid128_scalbn" && tc.Function != "bid128_scalbln")) {
		return fmt.Errorf("unsupported initial status %x for %s", tc.InitialStatus, tc.Function)
	}
	if tc.CompareGroup != "CMP_FUZZYSTATUS" || tc.UlpAdd != 0 || tc.UnderflowBeforeOnly {
		return fmt.Errorf("unsupported comparison shape %q", tc.CompareGroup)
	}
	if tc.Rounding < 0 || tc.Rounding > 4 {
		return fmt.Errorf("unsupported rounding mode %d", tc.Rounding)
	}
	flags, err := strconv.ParseUint(tc.Status, 16, 32)
	if err != nil || flags&^uint64(0x3d) != 0 {
		return fmt.Errorf("invalid five-flag status %q", tc.Status)
	}
	wantWidth := 128
	operandTypes := []string{"OP_DEC128", "OP_INT32"}
	operands := 2
	call := ""
	switch tc.Function {
	case "bid128_fma":
		if tc.Kind != "ternary_op" || tc.Format != "decimal128" {
			return fmt.Errorf("unsupported fma shape %s/%s", tc.Format, tc.Kind)
		}
		operandTypes = []string{"OP_DEC128", "OP_DEC128", "OP_DEC128"}
		operands = 3
		call = "let (got, flags) = bid754::generated::bid128_fma::bid128_fma(a, b, c, mode);"
	case "bid128_scalbn", "bid128_scalbln", "bid128_ldexp":
		if tc.Kind != "binary_op" || tc.Format != "decimal128" {
			return fmt.Errorf("unsupported scale shape %s/%s", tc.Format, tc.Kind)
		}
		if tc.Function == "bid128_scalbln" {
			operandTypes[1] = "OP_LINT"
		}
		if tc.Function == "bid128_ldexp" {
			call = "let (got, flags) = bid754::generated::bid128_ldexp::bid128_ldexp(a, n, mode);"
		} else {
			call = fmt.Sprintf("let mut flags = %du32; let got = bid754::generated::bid128_misc::%s(a, n, mode, &mut flags);", tc.InitialStatus, tc.Function)
		}
	case "bid32_nextafter":
		if tc.Kind != "binary_op" || tc.Format != "decimal32" {
			return fmt.Errorf("unsupported nextafter shape %s/%s", tc.Format, tc.Kind)
		}
		wantWidth = 32
		operandTypes = []string{"OP_DEC32", "OP_DEC32"}
		call = "let (got, flags) = bid754::generated::bid32_next::bid32_next_after(a, b);"
	case "bid32_quantum":
		if tc.Kind != "unary_op" || tc.Format != "decimal32" || tc.Rounding != 0 {
			return fmt.Errorf("unsupported quantum shape %s/%s/%d", tc.Format, tc.Kind, tc.Rounding)
		}
		wantWidth = 32
		operandTypes = []string{"OP_DEC32"}
		operands = 1
		call = "let got = bid754::generated::bid32_exports::bid32_quantum(a); let flags = 0u32;"
	default:
		return fmt.Errorf("unsupported function %q", tc.Function)
	}
	if tc.OutputType != operandTypes[0] || len(tc.InputTypes) != len(operandTypes) || len(tc.Operands) != operands {
		return fmt.Errorf("unsupported input/output arity or type")
	}
	for i, typ := range operandTypes {
		if tc.InputTypes[i] != typ {
			return fmt.Errorf("unsupported operand %d type %q", i, tc.InputTypes[i])
		}
	}
	var args []string
	for i, raw := range tc.Operands {
		var literal string
		if i > 0 && (tc.Function == "bid128_scalbn" || tc.Function == "bid128_scalbln" || tc.Function == "bid128_ldexp") {
			n, err := strconv.ParseInt(raw, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid integer operand %q: %w", raw, err)
			}
			literal = fmt.Sprintf("%di64", n)
		} else {
			literal, err = rustBIDLiteral(raw, wantWidth)
			if err != nil {
				return err
			}
		}
		args = append(args, literal)
	}
	want, err := rustBIDLiteral(tc.Expected, wantWidth)
	if err != nil {
		return err
	}
	b.WriteString("    {\n")
	for i, arg := range args {
		name := []string{"a", "b", "c"}[i]
		if i == 1 && (tc.Function == "bid128_scalbn" || tc.Function == "bid128_scalbln" || tc.Function == "bid128_ldexp") {
			name = "n"
		}
		fmt.Fprintf(b, "        let %s = %s;\n", name, arg)
	}
	if tc.Function != "bid32_nextafter" && tc.Function != "bid32_quantum" {
		fmt.Fprintf(b, "        let mode = %di64;\n", tc.Rounding)
	}
	fmt.Fprintf(b, "        %s\n", call)
	fmt.Fprintf(b, "        assert_eq!((got, flags), (%s, 0x%xu32), %q);\n", want, flags, tc.ID)
	b.WriteString("    }\n")
	return nil
}

func rustBIDLiteral(raw string, width int) (string, error) {
	hex := strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(raw), "["), "]")
	if len(hex) != width/4 {
		return "", fmt.Errorf("invalid BID%d literal %q", width, raw)
	}
	if width == 32 {
		v, err := strconv.ParseUint(hex, 16, 32)
		if err != nil {
			return "", fmt.Errorf("invalid BID32 literal %q: %w", raw, err)
		}
		return fmt.Sprintf("0x%08xu32", v), nil
	}
	hi, err := strconv.ParseUint(hex[:16], 16, 64)
	if err != nil {
		return "", fmt.Errorf("invalid BID128 literal %q: %w", raw, err)
	}
	lo, err := strconv.ParseUint(hex[16:], 16, 64)
	if err != nil {
		return "", fmt.Errorf("invalid BID128 literal %q: %w", raw, err)
	}
	return fmt.Sprintf("BID_UINT128 { lo: 0x%016xu64, hi: 0x%016xu64 }", lo, hi), nil
}
