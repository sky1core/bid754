package testgen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sky1core/bid754/devtools/internal/genmarker"
)

const (
	readtestGeneratedNativeTestPath = "../bid754-go/generated_readtest_cases_native_test.go"
	readtestGeneratedStubTestPath   = "../bid754-go/generated_readtest_cases_stub_test.go"
)

type readtestGeneratedCaseCounts struct {
	Total              int
	Decimal32          int
	Decimal64          int
	Decimal128         int
	FromString         int
	ToString           int
	UnaryOp            int
	BinaryOp           int
	TernaryOp          int
	StatusControl      int
	Functions          map[string]int
	Groups             map[string]int
	CompareGroups      map[string]int
	NativeCompareSkips map[string]int
}

func WriteReadtestTestOutputs(repoRoot string, spec SharedSpec) error {
	files, err := GenerateReadtestTestOutputs(spec)
	if err != nil {
		return err
	}
	for path, data := range files {
		fullPath := filepath.Join(repoRoot, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			return fmt.Errorf("mkdir %q: %w", filepath.Dir(fullPath), err)
		}
		if err := os.WriteFile(fullPath, data, 0o644); err != nil {
			return fmt.Errorf("write generated readtest test %q: %w", fullPath, err)
		}
	}
	return nil
}

func GenerateReadtestTestOutputs(spec SharedSpec) (map[string][]byte, error) {
	counts := countReadtestGeneratedCases(spec)
	return formatGeneratedGoOutputs(map[string][]byte{
		readtestGeneratedNativeTestPath: []byte(readtestNativeTestSource(counts)),
		readtestGeneratedStubTestPath: []byte(genmarker.Line("testgen") + `
//go:build !cgo || !bid754_native

package bid754

import "testing"

func TestGeneratedReadCases(t *testing.T) {
	t.Skip("generated readtest cases require cgo and bid754_native")
}
`),
	})
}

func countReadtestGeneratedCases(spec SharedSpec) readtestGeneratedCaseCounts {
	counts := readtestGeneratedCaseCounts{
		Functions:          map[string]int{},
		Groups:             map[string]int{},
		CompareGroups:      map[string]int{},
		NativeCompareSkips: map[string]int{},
	}
	for _, tc := range spec.ReadCases {
		counts.Total++
		counts.Functions[tc.Function]++
		counts.Groups[tc.Group]++
		counts.CompareGroups[tc.CompareGroup]++
		if tc.NativeCompareSkipReason != "" {
			counts.NativeCompareSkips[tc.NativeCompareSkipReason]++
		}
		switch tc.Format {
		case "decimal32":
			counts.Decimal32++
		case "decimal64":
			counts.Decimal64++
		case "decimal128":
			counts.Decimal128++
		}
		switch tc.Kind {
		case "from_string":
			counts.FromString++
		case "to_string":
			counts.ToString++
		case "unary_op":
			counts.UnaryOp++
		case "binary_op":
			counts.BinaryOp++
		case "ternary_op":
			counts.TernaryOp++
		case "status_control":
			counts.StatusControl++
		}
	}
	return counts
}

func readtestCountReplacer(counts readtestGeneratedCaseCounts) *strings.Replacer {
	return strings.NewReplacer(
		"@@TOTAL@@", fmt.Sprint(counts.Total),
		"@@DECIMAL32@@", fmt.Sprint(counts.Decimal32),
		"@@DECIMAL64@@", fmt.Sprint(counts.Decimal64),
		"@@DECIMAL128@@", fmt.Sprint(counts.Decimal128),
		"@@FROM_STRING@@", fmt.Sprint(counts.FromString),
		"@@TO_STRING@@", fmt.Sprint(counts.ToString),
		"@@UNARY_OP@@", fmt.Sprint(counts.UnaryOp),
		"@@BINARY_OP@@", fmt.Sprint(counts.BinaryOp),
		"@@TERNARY_OP@@", fmt.Sprint(counts.TernaryOp),
		"@@STATUS_CONTROL@@", fmt.Sprint(counts.StatusControl),
		"@@FUNCTION_COUNTS@@", stringIntMapLiteral(counts.Functions),
		"@@GROUP_COUNTS@@", stringIntMapLiteral(counts.Groups),
		"@@COMPARE_GROUP_COUNTS@@", stringIntMapLiteral(counts.CompareGroups),
		"@@NATIVE_COMPARE_SKIP_COUNTS@@", stringIntMapLiteral(counts.NativeCompareSkips),
	)
}

func readtestNativeTestSource(counts readtestGeneratedCaseCounts) string {
	return readtestCountReplacer(counts).Replace(genmarker.Line("testgen") + `
//go:build cgo && bid754_native

package bid754

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/sky1core/bid754/bid754-go/internal/testspec"
)

type generatedReadCaseCounts struct {
	Total              int
	Decimal32          int
	Decimal64          int
	Decimal128         int
	FromString         int
	ToString           int
	UnaryOp            int
	BinaryOp           int
	TernaryOp          int
	StatusControl      int
	Functions          map[string]int
	Groups             map[string]int
	CompareGroups      map[string]int
	NativeCompareSkips map[string]int
}

var expectedGeneratedReadCaseCounts = generatedReadCaseCounts{
	Total:         @@TOTAL@@,
	Decimal32:     @@DECIMAL32@@,
	Decimal64:     @@DECIMAL64@@,
	Decimal128:    @@DECIMAL128@@,
	FromString:    @@FROM_STRING@@,
	ToString:      @@TO_STRING@@,
	UnaryOp:       @@UNARY_OP@@,
	BinaryOp:      @@BINARY_OP@@,
	TernaryOp:     @@TERNARY_OP@@,
	StatusControl: @@STATUS_CONTROL@@,
	Functions: map[string]int{
@@FUNCTION_COUNTS@@
	},
	Groups: map[string]int{
@@GROUP_COUNTS@@
	},
	CompareGroups: map[string]int{
@@COMPARE_GROUP_COUNTS@@
	},
	NativeCompareSkips: map[string]int{
@@NATIVE_COMPARE_SKIP_COUNTS@@
	},
}

// nativeReadtestStringBackend routes the readtest.c check_results string
// round-trips through the Intel C oracle bid*_from_string, mirroring the
// upstream harness which uses the library's own conversion.
var nativeReadtestStringBackend = readtestStringBackend{
	FromString32:  nativeReadtestBID32FromString,
	FromString64:  nativeReadtestBID64FromString,
	FromString128: nativeReadtestBID128FromString,
}

// nativeReadtestOperationBackend routes the CMP_RELATIVEERR comparator's
// bid*_quantize / bid*_quiet_less calls through the Intel C oracle dispatch,
// mirroring the upstream check32/64/128_rel BIDECIMAL_CALL2 calls.
var nativeReadtestOperationBackend = readtestOperationBackend{
	Dec32:  nativeReadtestGeneratedBID32,
	Dec64:  nativeReadtestGeneratedBID64,
	Dec128: nativeReadtestGeneratedBID128,
	Signed: nativeReadtestGeneratedSigned,
}

func TestGeneratedReadCases(t *testing.T) {
	requireNative(t)
	if testing.Short() {
		t.Skip("generated readtest cases require non-short native run; use make test-native-readtest")
	}
	spec := loadGeneratedReadSpecForTest(t)
	if len(spec.ReadCases) == 0 {
		t.Fatal("expected generated read cases")
	}
	assertGeneratedReadCaseCounts(t, countGeneratedReadCases(spec.ReadCases), expectedGeneratedReadCaseCounts)

	for _, tc := range spec.ReadCases {
		tc := tc
		t.Run(tc.ID, func(t *testing.T) {
			if tc.NativeCompareSkipReason != "" {
				t.Skip(tc.NativeCompareSkipReason)
			}
			switch tc.Kind {
			case "from_string":
				got, status, err := generatedReadCaseBits(tc)
				if err != nil {
					t.Fatalf("generatedReadCaseBits(%q): %v", tc.Operands[0], err)
				}
				if strings.HasPrefix(strings.TrimSpace(tc.Expected), "[") {
					width, err := readtestFormatBitWidth(tc.Format)
					if err != nil {
						t.Fatalf("readtestFormatBitWidth(%q): %v", tc.Format, err)
					}
					if normalizeReadtestBits(got, width) != normalizeReadtestBits(tc.Expected, width) {
						t.Fatalf("generated read case %s line %d: expected bits %q, got %q", tc.ID, tc.Line, normalizeReadtestBits(tc.Expected, width), normalizeReadtestBits(got, width))
					}
				} else {
					equal, err := readtestDecimalRowEqual(tc.Format, tc.Expected, got, tc.Rounding, nativeReadtestStringBackend)
					if err != nil {
						t.Fatalf("readtestDecimalRowEqual(%q): %v", tc.Function, err)
					}
					if !equal {
						t.Fatalf("generated read case %s line %d: expected %q, got bits %q", tc.ID, tc.Line, tc.Expected, got)
					}
				}
				if normalizeReadtestStatus(status) != normalizeReadtestStatus(tc.Status) {
					t.Fatalf("generated read case %s line %d: expected status %q, got %q", tc.ID, tc.Line, normalizeReadtestStatus(tc.Status), normalizeReadtestStatus(status))
				}
			case "to_string":
				got, status, err := generatedReadCaseString(tc)
				if err != nil {
					t.Fatalf("generatedReadCaseString(%q): %v", tc.Operands[0], err)
				}
				equal, roundTripStatus, err := readtestToStringRowEqual(tc.Format, tc.Expected, got, tc.Rounding, nativeReadtestStringBackend)
				if err != nil {
					t.Fatalf("readtestToStringRowEqual(%q): %v", tc.Function, err)
				}
				if !equal {
					t.Fatalf("generated read case %s line %d: expected %q, got %q", tc.ID, tc.Line, tc.Expected, got)
				}
				combined, err := readtestCombineStatus(status, roundTripStatus)
				if err != nil {
					t.Fatalf("readtestCombineStatus(%q, %q): %v", status, roundTripStatus, err)
				}
				if normalizeReadtestStatus(combined) != normalizeReadtestStatus(tc.Status) {
					t.Fatalf("generated read case %s line %d: expected status %q, got %q", tc.ID, tc.Line, normalizeReadtestStatus(tc.Status), normalizeReadtestStatus(combined))
				}
			case "binary_op", "unary_op", "ternary_op", "status_control":
				if isReadtestScalarOutput(tc.OutputType) {
					got, status, err := generatedReadCaseScalarString(tc)
					if err != nil {
						t.Fatalf("generatedReadCaseScalarString(%q): %v", tc.Function, err)
					}
					if !compareReadtestScalarOutput(tc.OutputType, tc.Expected, got) {
						t.Fatalf("generated read case %s line %d: expected %q, got %q", tc.ID, tc.Line, tc.Expected, got)
					}
					if normalizeReadtestStatus(status) != normalizeReadtestStatus(tc.Status) {
						t.Fatalf("generated read case %s line %d: expected status %q, got %q", tc.ID, tc.Line, normalizeReadtestStatus(tc.Status), normalizeReadtestStatus(status))
					}
					return
				}
				got, sec, status, err := generatedReadCaseOperationBits(tc)
				if err != nil {
					t.Fatalf("generatedReadCaseOperationBits(%q): %v", tc.Function, err)
				}
				switch {
				case tc.CompareGroup == "CMP_RELATIVEERR":
					// readtest.c CMP_RELATIVEERR rows compare check*_rel plus the
					// trans_flags_mask-masked status only (readtest.c:1477/1486/1495);
					// no secondary output and no exact status comparison apply.
					equal, err := readtestRelativeErrRowEqual(tc.Format, tc.Expected, got, tc.Rounding, tc.UlpAdd, nativeReadtestStringBackend, nativeReadtestOperationBackend)
					if err != nil {
						t.Fatalf("readtestRelativeErrRowEqual(%q): %v", tc.Function, err)
					}
					if !equal {
						t.Fatalf("generated read case %s line %d: expected relative-error match %q (ulp_add %v), got bits %q", tc.ID, tc.Line, tc.Expected, tc.UlpAdd, got)
					}
					statusEqual, err := readtestRelativeErrStatusEqual(tc.Status, status)
					if err != nil {
						t.Fatalf("readtestRelativeErrStatusEqual(%q, %q): %v", tc.Status, status, err)
					}
					if !statusEqual {
						t.Fatalf("generated read case %s line %d: expected masked status %q, got %q", tc.ID, tc.Line, normalizeReadtestStatus(tc.Status), normalizeReadtestStatus(status))
					}
					return
				case tc.CompareGroup == "CMP_EQUALSTATUS":
					// readtest.c check_results does not compare the frexp/modf
					// secondary output in its CMP_EQUALSTATUS branches, so the
					// secondary check applies only to the value branches below.
					equal, err := readtestQuietEqual(tc.Format, tc.Expected, got, tc.Rounding, nativeReadtestStringBackend, nativeReadtestGeneratedSigned)
					if err != nil {
						t.Fatalf("readtestQuietEqual(%q): %v", tc.Function, err)
					}
					if !equal {
						t.Fatalf("generated read case %s line %d: expected quiet-equal %q, got %q", tc.ID, tc.Line, tc.Expected, got)
					}
				case strings.HasPrefix(strings.TrimSpace(tc.Expected), "["):
					width, err := readtestFormatBitWidth(tc.Format)
					if err != nil {
						t.Fatalf("readtestFormatBitWidth(%q): %v", tc.Format, err)
					}
					if normalizeReadtestBits(got, width) != normalizeReadtestBits(tc.Expected, width) {
						t.Fatalf("generated read case %s line %d: expected bits %q, got %q", tc.ID, tc.Line, normalizeReadtestBits(tc.Expected, width), normalizeReadtestBits(got, width))
					}
					secondaryEqual, err := readtestSecondaryOutputEqual(sec, tc.Operands)
					if err != nil {
						t.Fatalf("readtestSecondaryOutputEqual(%q): %v", tc.Function, err)
					}
					if !secondaryEqual {
						t.Fatalf("generated read case %s line %d: secondary output %+v does not match operand %q", tc.ID, tc.Line, sec, tc.Operands[sec.OperandIndex])
					}
				default:
					equal, err := readtestDecimalRowEqual(tc.Format, tc.Expected, got, tc.Rounding, nativeReadtestStringBackend)
					if err != nil {
						t.Fatalf("readtestDecimalRowEqual(%q): %v", tc.Function, err)
					}
					if !equal {
						t.Fatalf("generated read case %s line %d: expected %q, got bits %q", tc.ID, tc.Line, tc.Expected, got)
					}
					secondaryEqual, err := readtestSecondaryOutputEqual(sec, tc.Operands)
					if err != nil {
						t.Fatalf("readtestSecondaryOutputEqual(%q): %v", tc.Function, err)
					}
					if !secondaryEqual {
						t.Fatalf("generated read case %s line %d: secondary output %+v does not match operand %q", tc.ID, tc.Line, sec, tc.Operands[sec.OperandIndex])
					}
				}
				if normalizeReadtestStatus(status) != normalizeReadtestStatus(tc.Status) {
					t.Fatalf("generated read case %s line %d: expected status %q, got %q", tc.ID, tc.Line, normalizeReadtestStatus(tc.Status), normalizeReadtestStatus(status))
				}
			default:
				t.Fatalf("unsupported generated read kind %q", tc.Kind)
			}
		})
	}
}

func countGeneratedReadCases(cases []testspec.GeneratedReadCase) generatedReadCaseCounts {
	counts := generatedReadCaseCounts{
		Functions:          map[string]int{},
		Groups:             map[string]int{},
		CompareGroups:      map[string]int{},
		NativeCompareSkips: map[string]int{},
	}
	for _, tc := range cases {
		counts.Total++
		counts.Functions[tc.Function]++
		counts.Groups[tc.Group]++
		counts.CompareGroups[tc.CompareGroup]++
		if tc.NativeCompareSkipReason != "" {
			counts.NativeCompareSkips[tc.NativeCompareSkipReason]++
		}
		switch tc.Format {
		case "decimal32":
			counts.Decimal32++
		case "decimal64":
			counts.Decimal64++
		case "decimal128":
			counts.Decimal128++
		}
		switch tc.Kind {
		case "from_string":
			counts.FromString++
		case "to_string":
			counts.ToString++
		case "unary_op":
			counts.UnaryOp++
		case "binary_op":
			counts.BinaryOp++
		case "ternary_op":
			counts.TernaryOp++
		case "status_control":
			counts.StatusControl++
		}
	}
	return counts
}

func assertGeneratedReadCaseCounts(t *testing.T, got, want generatedReadCaseCounts) {
	t.Helper()
	for _, item := range []struct {
		label string
		got   int
		want  int
	}{
		{"total", got.Total, want.Total},
		{"decimal32", got.Decimal32, want.Decimal32},
		{"decimal64", got.Decimal64, want.Decimal64},
		{"decimal128", got.Decimal128, want.Decimal128},
		{"from_string", got.FromString, want.FromString},
		{"to_string", got.ToString, want.ToString},
		{"unary_op", got.UnaryOp, want.UnaryOp},
		{"binary_op", got.BinaryOp, want.BinaryOp},
		{"ternary_op", got.TernaryOp, want.TernaryOp},
		{"status_control", got.StatusControl, want.StatusControl},
	} {
		if item.got != item.want {
			t.Fatalf("generated read case %s count = %d, want %d", item.label, item.got, item.want)
		}
	}
	assertGeneratedReadStringCounts(t, "function", got.Functions, want.Functions)
	assertGeneratedReadStringCounts(t, "group", got.Groups, want.Groups)
	assertGeneratedReadStringCounts(t, "compare_group", got.CompareGroups, want.CompareGroups)
	assertGeneratedReadStringCounts(t, "native_compare_skip", got.NativeCompareSkips, want.NativeCompareSkips)
}

func assertGeneratedReadStringCounts(t *testing.T, label string, got, want map[string]int) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("generated read %s bucket count = %d, want %d", label, len(got), len(want))
	}
	for key, wantValue := range want {
		if got[key] != wantValue {
			t.Fatalf("generated read %s count[%q] = %d, want %d", label, key, got[key], wantValue)
		}
	}
}

func loadGeneratedReadSpecForTest(t *testing.T) testspec.SharedSpec {
	t.Helper()
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("resolve generated readtest file path")
	}
	spec, err := testspec.LoadGenerated(filepath.Join(filepath.Dir(currentFile), "..", "devtools", "generated", "testspec", "spec_index.json"))
	if err != nil {
		t.Fatalf("load shared spec: %v", err)
	}
	return spec
}

func generatedReadCaseBits(tc testspec.GeneratedReadCase) (string, string, error) {
	if len(tc.Operands) != 1 {
		return "", "", fmt.Errorf("%s expects 1 operand, got %d", tc.Kind, len(tc.Operands))
	}
	switch tc.Format {
	case "decimal32":
		raw, status := nativeReadtestBID32FromString(tc.Operands[0], tc.Rounding)
		return fmt.Sprintf("[%08x]", raw), status, nil
	case "decimal64":
		raw, status := nativeReadtestBID64FromString(tc.Operands[0], tc.Rounding)
		return fmt.Sprintf("[%016x]", raw), status, nil
	case "decimal128":
		raw, status := nativeReadtestBID128FromString(tc.Operands[0], tc.Rounding)
		return formatReadtestBits128(raw), status, nil
	default:
		return "", "", fmt.Errorf("unsupported read format %q", tc.Format)
	}
}

func generatedReadCaseString(tc testspec.GeneratedReadCase) (string, string, error) {
	if len(tc.Operands) != 1 {
		return "", "", fmt.Errorf("%s expects 1 operand, got %d", tc.Kind, len(tc.Operands))
	}
	switch tc.Format {
	case "decimal32":
		raw, err := parseReadtestHex(tc.Operands[0], 32)
		if err != nil {
			return "", "", err
		}
		value, status := nativeReadtestBID32ToString(uint32(raw))
		return value, status, nil
	case "decimal64":
		raw, err := parseReadtestHex(tc.Operands[0], 64)
		if err != nil {
			return "", "", err
		}
		value, status := nativeReadtestBID64ToString(raw)
		return value, status, nil
	case "decimal128":
		raw, err := parseReadtestBits128(tc.Operands[0])
		if err != nil {
			return "", "", err
		}
		value, status := nativeReadtestBID128ToString(raw)
		return value, status, nil
	default:
		return "", "", fmt.Errorf("unsupported read format %q", tc.Format)
	}
}

func generatedReadCaseOperationBits(tc testspec.GeneratedReadCase) (string, readtestSecondaryOutput, string, error) {
	switch tc.Format {
	case "decimal32":
		raw, sec, status, err := nativeReadtestGeneratedBID32(tc.Function, tc.Rounding, tc.Operands)
		if err != nil {
			return "", readtestNoSecondaryOutput(), "", err
		}
		return fmt.Sprintf("[%08x]", raw), sec, status, nil
	case "decimal64":
		raw, sec, status, err := nativeReadtestGeneratedBID64(tc.Function, tc.Rounding, tc.Operands)
		if err != nil {
			return "", readtestNoSecondaryOutput(), "", err
		}
		return fmt.Sprintf("[%016x]", raw), sec, status, nil
	case "decimal128":
		raw, sec, status, err := nativeReadtestGeneratedBID128(tc.Function, tc.Rounding, tc.Operands)
		if err != nil {
			return "", readtestNoSecondaryOutput(), "", err
		}
		return formatReadtestBits128(raw), sec, status, nil
	default:
		return "", readtestNoSecondaryOutput(), "", fmt.Errorf("unsupported read format %q", tc.Format)
	}
}

func generatedReadCaseScalarString(tc testspec.GeneratedReadCase) (string, string, error) {
	switch tc.OutputType {
	case "OP_BIN32":
		value, status, err := nativeReadtestGeneratedBinary32(tc.Function, tc.Rounding, tc.Operands)
		if err != nil {
			return "", "", err
		}
		return fmt.Sprintf("[%08x]", value), status, nil
	case "OP_BIN64":
		value, status, err := nativeReadtestGeneratedBinary64(tc.Function, tc.Rounding, tc.Operands)
		if err != nil {
			return "", "", err
		}
		return fmt.Sprintf("[%016x]", value), status, nil
	case "OP_BIN128":
		value, status, err := nativeReadtestGeneratedBinary128(tc.Function, tc.Rounding, tc.Operands)
		if err != nil {
			return "", "", err
		}
		return formatReadtestBits128(value), status, nil
	case "OP_BID_UINT8", "OP_BID_UINT16", "OP_BID_UINT32", "OP_BID_UINT64":
		value, status, err := nativeReadtestGeneratedUnsigned(tc.Function, tc.Rounding, tc.Operands)
		if err != nil {
			return "", "", err
		}
		return strconv.FormatUint(value, 10), status, nil
	case "OP_INT8", "OP_INT16", "OP_INT32", "OP_INT64", "OP_LINT":
		value, status, err := nativeReadtestGeneratedSigned(tc.Function, tc.Rounding, tc.Operands)
		if err != nil {
			return "", "", err
		}
		return strconv.FormatInt(value, 10), status, nil
	default:
		return "", "", fmt.Errorf("unsupported scalar output %q", tc.OutputType)
	}
}
`)
}
