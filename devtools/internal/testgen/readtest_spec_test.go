package testgen

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestAppendGeneratedReadCasesUsesSourceLocalStableIDs(t *testing.T) {
	repoRoot := t.TempDir()
	header := "readtest.h"
	source := "readtest.in"
	if err := os.WriteFile(filepath.Join(repoRoot, header), []byte(`if (!strcmp(func, "bid32_add")) {}`), 0o600); err != nil {
		t.Fatalf("write readtest header: %v", err)
	}
	input := "" +
		"-- pinned source comment\n" +
		"bid32_add 0 1 2 [32800003] 00\n" +
		"bid64_add 0 1 2 [31c0000000000003] 00\n" +
		"bid32_add 0 3 4 [32800007] 00\n"
	if err := os.WriteFile(filepath.Join(repoRoot, source), []byte(input), 0o600); err != nil {
		t.Fatalf("write readtest source: %v", err)
	}

	read := ReadTestSpec{
		Name:          "bid32_add",
		Function:      "bid32_add",
		Format:        "decimal32",
		Header:        header,
		Source:        source,
		Kind:          "binary_op",
		OutputType:    "OP_DEC32",
		InputTypes:    []string{"OP_DEC32", "OP_DEC32"},
		CompareGroup:  "CMP_FUZZYSTATUS",
		Statuses:      []string{"00"},
		RoundingModes: []int{0},
	}
	spec := SharedSpec{ReadCases: []GeneratedReadCase{{ID: "unrelated_existing_case"}}}
	count, skips, err := appendGeneratedReadCases(repoRoot, &spec, read)
	if err != nil {
		t.Fatalf("appendGeneratedReadCases: %v", err)
	}
	if count != 2 || len(skips) != 0 {
		t.Fatalf("accepted/skipped = %d/%v, want 2/none", count, skips)
	}

	got := []string{spec.ReadCases[1].ID, spec.ReadCases[2].ID}
	want := []string{"bid32_add_line_2", "bid32_add_line_4"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("generated IDs = %v, want source-local IDs %v", got, want)
	}
}

func TestGroupReadtestShardsRejectsDuplicateCaseIDs(t *testing.T) {
	cases := []GeneratedReadCase{
		{Suite: "bid32_add", ID: "bid32_add_line_2"},
		{Suite: "bid32_add", ID: "bid32_add_line_2"},
	}
	_, err := groupReadtestShards(cases)
	if err == nil || !strings.Contains(err.Error(), `duplicate readtest case ID "bid32_add_line_2"`) {
		t.Fatalf("duplicate readtest case ID error = %v", err)
	}
}

func TestParseReadtestSubsetAcceptsIntelQNaNLiteral(t *testing.T) {
	cases, skips := parseBid32AddReadtestRows(t, "bid32_add 0 QNaN 1 [7c000000] 00\n")
	if len(skips) != 0 {
		t.Fatalf("row skips = %v, want none", skips)
	}

	want := []parsedReadtestCase{
		{
			Function: "bid32_add",
			Rounding: 0,
			Operands: []string{"QNaN", "1"},
			Result:   "[7c000000]",
			Status:   "00",
			Line:     1,
		},
	}
	if !reflect.DeepEqual(cases, want) {
		t.Fatalf("parsed cases mismatch\ngot:  %+v\nwant: %+v", cases, want)
	}
}

func TestParseReadtestSubsetStripsIntelInlineCommentsBeforeTokenizing(t *testing.T) {
	input := "" +
		"bid32_add 0 [32800001] [32800001] [32800002] 00 -- 1\n" +
		"-- bid32_add 0 1 1 [32800002] 00\n"
	cases, skips := parseBid32AddReadtestRows(t, input)
	if len(skips) != 0 {
		t.Fatalf("row skips = %v, want none", skips)
	}

	want := []parsedReadtestCase{
		{
			Function: "bid32_add",
			Rounding: 0,
			Operands: []string{"[32800001]", "[32800001]"},
			Result:   "[32800002]",
			Status:   "00",
			Line:     1,
		},
	}
	if !reflect.DeepEqual(cases, want) {
		t.Fatalf("parsed cases mismatch\ngot:  %+v\nwant: %+v", cases, want)
	}
}

func TestParseReadtestSubsetKeepsMalformedFromStringCases(t *testing.T) {
	spec := ReadTestSpec{
		Function:      "bid32_from_string",
		Kind:          "from_string",
		OutputType:    "OP_DEC32",
		InputTypes:    []string{"OP_DEC32"},
		Statuses:      []string{"00"},
		RoundingModes: []int{0},
	}
	cases, skips := parseReadtestRows(t, spec, "bid32_from_string 0 1.. [7c000000] 00\n")
	if len(skips) != 0 {
		t.Fatalf("row skips = %v, want none", skips)
	}
	if len(cases) != 1 || !reflect.DeepEqual(cases[0].Operands, []string{"1.."}) {
		t.Fatalf("malformed from_string row = %+v, want one executable raw-string case", cases)
	}
}

func TestParseReadtestSubsetKeepsDecimalExpectedFromStringCases(t *testing.T) {
	spec := ReadTestSpec{
		Function:      "bid64_from_string",
		Kind:          "from_string",
		OutputType:    "OP_DEC64",
		InputTypes:    []string{"OP_DEC64"},
		Statuses:      []string{"20"},
		RoundingModes: []int{0},
	}
	input := "bid64_from_string 0 12345678901234565 1234567890123456e1 20\n"
	cases, skips := parseReadtestRows(t, spec, input)
	if len(skips) != 0 || len(cases) != 1 {
		t.Fatalf("cases = %+v, skips = %v, want one decimal-expected from_string case", cases, skips)
	}
}

func TestParseReadtestSubsetMirrorsIntelScanfIntegerTokens(t *testing.T) {
	spec := ReadTestSpec{
		Function:      "bid32_scalbn",
		Kind:          "binary_op",
		OutputType:    "OP_DEC32",
		InputTypes:    []string{"OP_DEC32", "OP_INT32"},
		Statuses:      []string{"00", "28"},
		RoundingModes: []int{0},
	}
	input := "" +
		"bid32_scalbn 0 1.0 -1.0 [3180000a] 00\n" +
		"bid32_scalbn 0 1.0 [60989680] [78000000] 28\n"
	cases, skips := parseReadtestRows(t, spec, input)
	if len(skips) != 0 {
		t.Fatalf("row skips = %v, want none", skips)
	}
	if len(cases) != 2 {
		t.Fatalf("parsed cases = %d, want 2", len(cases))
	}
}

func TestParseReadtestSubsetKeepsBracketedSignedResult(t *testing.T) {
	spec := ReadTestSpec{
		Function:      "bid64_ilogb",
		Kind:          "unary_op",
		OutputType:    "OP_INT32",
		InputTypes:    []string{"OP_DEC64"},
		Statuses:      []string{"01"},
		RoundingModes: []int{0},
	}
	cases, skips := parseReadtestRows(t, spec, "bid64_ilogb 0 0 [80000000] 01\n")
	if len(skips) != 0 || len(cases) != 1 {
		t.Fatalf("cases = %+v, skips = %v, want one executable bracketed-int case", cases, skips)
	}
}

func TestParseReadtestSubsetDropsOnlyHeaderIgnoredOperands(t *testing.T) {
	spec := ReadTestSpec{
		Function:      "bid128_nextdown",
		Kind:          "unary_op",
		OutputType:    "OP_DEC128",
		InputTypes:    []string{"OP_DEC128"},
		Statuses:      []string{"00"},
		RoundingModes: []int{0},
	}
	input := "bid128_nextdown 0 1 [7c00314dc6448d9338c15b0a00000001] [2ffded09bead87c0378d8e63ffffffff] 00\n"
	cases, skips := parseReadtestRows(t, spec, input)
	if len(skips) != 0 || len(cases) != 1 {
		t.Fatalf("cases = %+v, skips = %v, want one executable unary case", cases, skips)
	}
	if !reflect.DeepEqual(cases[0].Operands, []string{"1"}) {
		t.Fatalf("operands = %v, want only the GETTEST1-consumed operand", cases[0].Operands)
	}
}

func TestRepairKnownReadtestOperandsKeepsIntelScanfAcceptedRow(t *testing.T) {
	got := repairKnownReadtestOperands(
		"bid128qd_div",
		11608,
		[]string{"[80000000000000000000000000000001]", "[2fe0000000000005"},
	)
	want := []string{"[80000000000000000000000000000001]", "[2fe0000000000005]"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("repaired operands = %v, want %v", got, want)
	}
	if !supportsReadtestValue("OP_DEC64", got[1], false) {
		t.Fatalf("repaired Decimal64 operand %q is not executable by generated runners", got[1])
	}
}

func TestReadtestValueSupportMatchesPinnedCInputForms(t *testing.T) {
	for _, tc := range []struct {
		kind  string
		value string
	}{
		{kind: "OP_DEC64", value: "[8000000000000000,0000000000000000]"},
		{kind: "OP_INT32", value: "[FFFFEB11]"},
		{kind: "OP_INT32", value: "-1.0e-96"},
		{kind: "OP_BID_UINT64", value: "-9223372036854775809"},
	} {
		if !supportsReadtestValue(tc.kind, tc.value, false) {
			t.Errorf("supportsReadtestValue(%q, %q) = false, want true", tc.kind, tc.value)
		}
	}
	if supportsReadtestValue("OP_INT64", "9223372036854775808", false) {
		t.Error("signed readtest integer above int64 max was accepted")
	}
}

func TestGoportArgExprPreservesReadtestSourceIntegerWidth(t *testing.T) {
	for _, tc := range []struct {
		name      string
		inputType string
		goType    string
		want      string
	}{
		{name: "int32 to native int", inputType: "OP_INT32", goType: "int", want: "int(int32(raw))"},
		{name: "int32 widened to int64", inputType: "OP_INT32", goType: "int64", want: "int64(int32(raw))"},
		{name: "uint32 widened to uint64", inputType: "OP_BID_UINT32", goType: "uint64", want: "uint64(uint32(raw))"},
		{name: "int64 unchanged", inputType: "OP_INT64", goType: "int64", want: "raw"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := goportArgExpr(goportArgPlan{InputType: tc.inputType, GoType: tc.goType}, "raw")
			if err != nil {
				t.Fatalf("goportArgExpr: %v", err)
			}
			if got != tc.want {
				t.Fatalf("goportArgExpr = %q, want %q", got, tc.want)
			}
		})
	}
}

func parseBid32AddReadtestRows(t *testing.T, input string) ([]parsedReadtestCase, map[string]int) {
	t.Helper()
	spec := ReadTestSpec{
		Function:      "bid32_add",
		Kind:          "binary",
		OutputType:    "OP_DEC32",
		InputTypes:    []string{"OP_DEC32", "OP_DEC32"},
		Statuses:      []string{"00"},
		RoundingModes: []int{0},
	}
	return parseReadtestRows(t, spec, input)
}

func parseReadtestRows(t *testing.T, spec ReadTestSpec, input string) ([]parsedReadtestCase, map[string]int) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "readtest.in")
	if err := os.WriteFile(path, []byte(input), 0o600); err != nil {
		t.Fatalf("WriteFile(readtest.in): %v", err)
	}
	cases, skips, err := parseReadtestSubset(path, spec)
	if err != nil {
		t.Fatalf("parseReadtestSubset: %v", err)
	}
	return cases, skips
}
