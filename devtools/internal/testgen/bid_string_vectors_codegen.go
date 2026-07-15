package testgen

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/sky1core/bid754/devtools/internal/genmarker"
)

const (
	bidStringVectorsGoTestPath   = "../bid754-go/internal/bidgo/string_vectors_test.go"
	bidStringVectorsRustTestPath = "../bid754-rs/tests/bid_string_vectors.rs"
)

type bidStringVectorCounts struct {
	Total                int
	FromString           int
	ToString             int
	Decimal32            int
	Decimal64            int
	Decimal128           int
	Decimal32FromString  int
	Decimal64FromString  int
	Decimal128FromString int
	Decimal32ToString    int
	Decimal64ToString    int
	Decimal128ToString   int
}

func WriteBidStringVectorTestOutputs(repoRoot string, spec SharedSpec) error {
	files := GenerateBidStringVectorTestOutputs(spec)
	for path, data := range files {
		fullPath := filepath.Join(repoRoot, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			return fmt.Errorf("mkdir %q: %w", filepath.Dir(fullPath), err)
		}
		if err := os.WriteFile(fullPath, data, 0o644); err != nil {
			return fmt.Errorf("write generated BID string vector test %q: %w", fullPath, err)
		}
	}
	return nil
}

func GenerateBidStringVectorTestOutputs(spec SharedSpec) map[string][]byte {
	counts := countBidStringVectorCases(spec)
	shardFiles := bidStringVectorShardFiles(spec)
	return map[string][]byte{
		bidStringVectorsGoTestPath:   []byte(bidStringVectorsGoTestSource(counts, shardFiles)),
		bidStringVectorsRustTestPath: []byte(bidStringVectorsRustTestSource(counts, shardFiles)),
	}
}

// bidStringVectorShardFiles lists the readtest shard files of every
// from_string/to_string suite in generation order, relative to the
// generated/testspec directory.
func bidStringVectorShardFiles(spec SharedSpec) []string {
	var files []string
	seen := map[string]struct{}{}
	for _, tc := range spec.ReadCases {
		if tc.Kind != "from_string" && tc.Kind != "to_string" {
			continue
		}
		if _, ok := seen[tc.Suite]; ok {
			continue
		}
		seen[tc.Suite] = struct{}{}
		files = append(files, readtestShardRelPath(tc.Suite))
	}
	return files
}

func goBidStringShardFilesLiteral(shardFiles []string) string {
	var b strings.Builder
	for _, file := range shardFiles {
		b.WriteString("\t")
		b.WriteString(strconv.Quote(file))
		b.WriteString(",\n")
	}
	return b.String()
}

func rustBidStringShardFilesLiteral(shardFiles []string) string {
	var b strings.Builder
	for _, file := range shardFiles {
		b.WriteString("    ")
		b.WriteString(strconv.Quote(file))
		b.WriteString(",\n")
	}
	return b.String()
}

func countBidStringVectorCases(spec SharedSpec) bidStringVectorCounts {
	var counts bidStringVectorCounts
	for _, tc := range spec.ReadCases {
		if tc.Kind != "from_string" && tc.Kind != "to_string" {
			continue
		}
		counts.Total++
		switch tc.Kind {
		case "from_string":
			counts.FromString++
		case "to_string":
			counts.ToString++
		}
		switch tc.Format {
		case "decimal32":
			counts.Decimal32++
			if tc.Kind == "from_string" {
				counts.Decimal32FromString++
			} else {
				counts.Decimal32ToString++
			}
		case "decimal64":
			counts.Decimal64++
			if tc.Kind == "from_string" {
				counts.Decimal64FromString++
			} else {
				counts.Decimal64ToString++
			}
		case "decimal128":
			counts.Decimal128++
			if tc.Kind == "from_string" {
				counts.Decimal128FromString++
			} else {
				counts.Decimal128ToString++
			}
		}
	}
	return counts
}

func replaceBidStringCountPlaceholders(src string, counts bidStringVectorCounts) string {
	return strings.NewReplacer(
		"@@TOTAL@@", fmt.Sprint(counts.Total),
		"@@FROM_STRING@@", fmt.Sprint(counts.FromString),
		"@@TO_STRING@@", fmt.Sprint(counts.ToString),
		"@@DECIMAL32@@", fmt.Sprint(counts.Decimal32),
		"@@DECIMAL64@@", fmt.Sprint(counts.Decimal64),
		"@@DECIMAL128@@", fmt.Sprint(counts.Decimal128),
		"@@DECIMAL32_FROM_STRING@@", fmt.Sprint(counts.Decimal32FromString),
		"@@DECIMAL64_FROM_STRING@@", fmt.Sprint(counts.Decimal64FromString),
		"@@DECIMAL128_FROM_STRING@@", fmt.Sprint(counts.Decimal128FromString),
		"@@DECIMAL32_TO_STRING@@", fmt.Sprint(counts.Decimal32ToString),
		"@@DECIMAL64_TO_STRING@@", fmt.Sprint(counts.Decimal64ToString),
		"@@DECIMAL128_TO_STRING@@", fmt.Sprint(counts.Decimal128ToString),
	).Replace(src)
}

func bidStringVectorsGoTestSource(counts bidStringVectorCounts, shardFiles []string) string {
	src := strings.Replace(genmarker.Line("testgen")+`
package bidgo

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

// generatedBIDStringShardFiles lists the readtest shard files of every
// from_string/to_string suite, relative to generated/testspec, in
// generation order.
var generatedBIDStringShardFiles = []string{
@@STRING_SHARD_FILES@@}

type generatedStringShard struct {
	Format   string                     `+"`json:\"format\"`"+`
	Function string                     `+"`json:\"function\"`"+`
	Kind     string                     `+"`json:\"kind\"`"+`
	Cases    []generatedStringShardCase `+"`json:\"cases\"`"+`
}

type generatedStringShardCase struct {
	ID       string   `+"`json:\"id\"`"+`
	Line     int      `+"`json:\"line\"`"+`
	Operands []string `+"`json:\"operands\"`"+`
	Expected string   `+"`json:\"expected\"`"+`
	Status   string   `+"`json:\"status\"`"+`
	Rounding int      `+"`json:\"rounding\"`"+`
}

type generatedStringReadCase struct {
	ID       string
	Line     int
	Format   string
	Function string
	Kind     string
	Operands []string
	Expected string
	Status   string
	Rounding int
}

type generatedBIDStringCounts struct {
	Total                int
	FromString           int
	ToString             int
	Decimal32            int
	Decimal64            int
	Decimal128           int
	Decimal32FromString  int
	Decimal64FromString  int
	Decimal128FromString int
	Decimal32ToString    int
	Decimal64ToString    int
	Decimal128ToString   int
}

var expectedGeneratedBIDStringCounts = generatedBIDStringCounts{
	Total:                @@TOTAL@@,
	FromString:           @@FROM_STRING@@,
	ToString:             @@TO_STRING@@,
	Decimal32:            @@DECIMAL32@@,
	Decimal64:            @@DECIMAL64@@,
	Decimal128:           @@DECIMAL128@@,
	Decimal32FromString:  @@DECIMAL32_FROM_STRING@@,
	Decimal64FromString:  @@DECIMAL64_FROM_STRING@@,
	Decimal128FromString: @@DECIMAL128_FROM_STRING@@,
	Decimal32ToString:    @@DECIMAL32_TO_STRING@@,
	Decimal64ToString:    @@DECIMAL64_TO_STRING@@,
	Decimal128ToString:   @@DECIMAL128_TO_STRING@@,
}

func TestGeneratedBIDStringVectors(t *testing.T) {
	readCases := loadGeneratedBIDStringReadCases(t)
	var counts generatedBIDStringCounts
	for _, tc := range readCases {
		if tc.Kind != "from_string" && tc.Kind != "to_string" {
			continue
		}
		countGeneratedBIDStringCase(&counts, tc)
		t.Run(tc.ID, func(t *testing.T) {
			switch tc.Kind {
			case "from_string":
				gotBits, flags, err := generatedBIDStringFromString(tc)
				if err != nil {
					t.Fatalf("from_string dispatch: %v", err)
				}
				if strings.HasPrefix(strings.TrimSpace(tc.Expected), "[") {
					if normalizeGeneratedBIDStringBits(gotBits) != normalizeGeneratedBIDStringBits(tc.Expected) {
						t.Fatalf("%s line %d: bits got %s, want %s", tc.Function, tc.Line, gotBits, tc.Expected)
					}
				} else {
					equal, err := generatedBIDStringDecimalRowEqual(tc.Format, tc.Expected, gotBits, tc.Rounding)
					if err != nil {
						t.Fatalf("compare from_string decimal result: %v", err)
					}
					if !equal {
						t.Fatalf("%s line %d: bits got %s, want exact cohort parsed from %q", tc.Function, tc.Line, gotBits, tc.Expected)
					}
				}
				if normalizeGeneratedBIDStringStatus(flags) != normalizeGeneratedBIDStringStatus(tc.Status) {
					t.Fatalf("%s line %d: status got %s, want %s", tc.Function, tc.Line, normalizeGeneratedBIDStringStatus(flags), normalizeGeneratedBIDStringStatus(tc.Status))
				}
			case "to_string":
				got, flags, err := generatedBIDStringToString(tc)
				if err != nil {
					t.Fatalf("to_string dispatch: %v", err)
				}
				equal, roundTripStatus, err := generatedBIDStringToStringRowEqual(tc.Format, tc.Expected, got, tc.Rounding)
				if err != nil {
					t.Fatalf("compare to_string result: %v", err)
				}
				if !equal {
					t.Fatalf("%s line %d: got %q, want exact cohort parsed from %q", tc.Function, tc.Line, got, tc.Expected)
				}
				flags, err = combineGeneratedBIDStringStatus(flags, roundTripStatus)
				if err != nil {
					t.Fatalf("combine to_string status: %v", err)
				}
				if normalizeGeneratedBIDStringStatus(flags) != normalizeGeneratedBIDStringStatus(tc.Status) {
					t.Fatalf("%s line %d: status got %s, want %s", tc.Function, tc.Line, normalizeGeneratedBIDStringStatus(flags), normalizeGeneratedBIDStringStatus(tc.Status))
				}
			}
		})
	}
	if counts != expectedGeneratedBIDStringCounts {
		t.Fatalf("generated BID string read case counts changed: got %+v, want %+v", counts, expectedGeneratedBIDStringCounts)
	}
}

// TestGeneratedBIDStringToStringComparatorStrength anchors the generated
// comparator to Intel readtest.c: expected and produced strings are parsed by
// the backend at the row rounding mode and compared as exact BID cohorts, and
// flags from parsing the produced string are accumulated into operation flags.
func TestGeneratedBIDStringToStringComparatorStrength(t *testing.T) {
	equal, _, err := generatedBIDStringToStringRowEqual("decimal64", "+15E+0", "+150E-1", 0)
	if err != nil {
		t.Fatalf("wrong-cohort comparison: %v", err)
	}
	if equal {
		t.Fatal("wrong-cohort string accepted: +150E-1 must not match +15E+0")
	}

	equal, status, err := generatedBIDStringToStringRowEqual("decimal64", "15", "+15E+0", 0)
	if err != nil {
		t.Fatalf("exact-cohort spelling comparison: %v", err)
	}
	if !equal {
		t.Fatal("exact-cohort spelling variant rejected")
	}
	if normalizeGeneratedBIDStringStatus(status) != "00" {
		t.Fatalf("exact-cohort round-trip status = %q, want 00", status)
	}

	equal, status, err = generatedBIDStringToStringRowEqual("decimal64", "5000000000000000E-15", "5.0000000000000000001", 0)
	if err != nil {
		t.Fatalf("inexact round-trip comparison: %v", err)
	}
	if !equal {
		t.Fatal("inexact produced string did not round to the expected cohort")
	}
	if normalizeGeneratedBIDStringStatus(status) != "20" {
		t.Fatalf("inexact round-trip status = %q, want 20", status)
	}
	combined, err := combineGeneratedBIDStringStatus("01", status)
	if err != nil {
		t.Fatalf("combine status: %v", err)
	}
	if normalizeGeneratedBIDStringStatus(combined) != "21" {
		t.Fatalf("combined status = %q, want 21", combined)
	}
}

func countGeneratedBIDStringCase(counts *generatedBIDStringCounts, tc generatedStringReadCase) {
	counts.Total++
	switch tc.Kind {
	case "from_string":
		counts.FromString++
	case "to_string":
		counts.ToString++
	}
	switch tc.Format {
	case "decimal32":
		counts.Decimal32++
		if tc.Kind == "from_string" {
			counts.Decimal32FromString++
		} else {
			counts.Decimal32ToString++
		}
	case "decimal64":
		counts.Decimal64++
		if tc.Kind == "from_string" {
			counts.Decimal64FromString++
		} else {
			counts.Decimal64ToString++
		}
	case "decimal128":
		counts.Decimal128++
		if tc.Kind == "from_string" {
			counts.Decimal128FromString++
		} else {
			counts.Decimal128ToString++
		}
	}
}

func loadGeneratedBIDStringReadCases(t *testing.T) []generatedStringReadCase {
	t.Helper()
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("resolve generated string vector test path")
	}
	baseDir := filepath.Join(filepath.Dir(currentFile), "..", "..", "..", "devtools", "generated", "testspec")
	var readCases []generatedStringReadCase
	for _, shardFile := range generatedBIDStringShardFiles {
		path := filepath.Join(baseDir, filepath.FromSlash(shardFile))
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read string shard %q: %v", path, err)
		}
		var shard generatedStringShard
		if err := json.Unmarshal(data, &shard); err != nil {
			t.Fatalf("parse string shard %q: %v", path, err)
		}
		for _, tc := range shard.Cases {
			readCases = append(readCases, generatedStringReadCase{
				ID:       tc.ID,
				Line:     tc.Line,
				Format:   shard.Format,
				Function: shard.Function,
				Kind:     shard.Kind,
				Operands: tc.Operands,
				Expected: tc.Expected,
				Status:   tc.Status,
				Rounding: tc.Rounding,
			})
		}
	}
	return readCases
}

func generatedBIDStringFromString(tc generatedStringReadCase) (string, string, error) {
	if len(tc.Operands) != 1 {
		return "", "", fmt.Errorf("%s expects 1 operand, got %d", tc.Function, len(tc.Operands))
	}
	return generatedBIDStringParseValue(tc.Format, tc.Operands[0], tc.Rounding)
}

func generatedBIDStringParseValue(format, input string, rounding int) (string, string, error) {
	switch format {
	case "decimal32":
		got, flags := Bid32FromStringRaw(input, rounding)
		return fmt.Sprintf("[%08x]", got), fmt.Sprintf("%02X", flags), nil
	case "decimal64":
		got, flags := Bid64FromString(input, rounding)
		return fmt.Sprintf("[%016x]", got), fmt.Sprintf("%02X", flags), nil
	case "decimal128":
		got, flags := Bid128FromString(input, rounding)
		return formatGeneratedBIDStringBits128(got), fmt.Sprintf("%02X", flags), nil
	default:
		return "", "", fmt.Errorf("unsupported format %q", format)
	}
}

func generatedBIDStringToString(tc generatedStringReadCase) (string, string, error) {
	if len(tc.Operands) != 1 {
		return "", "", fmt.Errorf("%s expects 1 operand, got %d", tc.Function, len(tc.Operands))
	}
	switch tc.Format {
	case "decimal32":
		raw, err := parseGeneratedBIDStringUintBits(tc.Operands[0], 32)
		if err != nil {
			return "", "", err
		}
		return Bid32ToString(uint32(raw)), "00", nil
	case "decimal64":
		raw, err := parseGeneratedBIDStringUintBits(tc.Operands[0], 64)
		if err != nil {
			return "", "", err
		}
		return Bid64ToString(raw), "00", nil
	case "decimal128":
		raw, err := parseGeneratedBIDStringBits128(tc.Operands[0])
		if err != nil {
			return "", "", err
		}
		return Bid128ToString(raw), "00", nil
	default:
		return "", "", fmt.Errorf("unsupported format %q", tc.Format)
	}
}

func generatedBIDStringDecimalRowEqual(format, expected, gotBits string, rounding int) (bool, error) {
	expectedBits, _, err := generatedBIDStringParseValue(format, expected, rounding)
	if err != nil {
		return false, err
	}
	return normalizeGeneratedBIDStringBits(expectedBits) == normalizeGeneratedBIDStringBits(gotBits), nil
}

func generatedBIDStringToStringRowEqual(format, expected, got string, rounding int) (bool, string, error) {
	expectedBits, _, err := generatedBIDStringParseValue(format, expected, rounding)
	if err != nil {
		return false, "", err
	}
	gotBits, roundTripStatus, err := generatedBIDStringParseValue(format, got, rounding)
	if err != nil {
		return false, "", err
	}
	return normalizeGeneratedBIDStringBits(expectedBits) == normalizeGeneratedBIDStringBits(gotBits), roundTripStatus, nil
}

func combineGeneratedBIDStringStatus(a, b string) (string, error) {
	flagsA, err := strconv.ParseUint(normalizeGeneratedBIDStringStatus(a), 16, 32)
	if err != nil {
		return "", fmt.Errorf("parse status %q: %w", a, err)
	}
	flagsB, err := strconv.ParseUint(normalizeGeneratedBIDStringStatus(b), 16, 32)
	if err != nil {
		return "", fmt.Errorf("parse status %q: %w", b, err)
	}
	return fmt.Sprintf("%02X", uint32(flagsA|flagsB)), nil
}

func parseGeneratedBIDStringUintBits(input string, bits int) (uint64, error) {
	trimmed := strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(input), "["), "]")
	trimmed = strings.ReplaceAll(trimmed, ",", "")
	return strconv.ParseUint(trimmed, 16, bits)
}

func parseGeneratedBIDStringBits128(input string) (BID_UINT128, error) {
	var raw BID_UINT128
	literal := strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(input), "["), "]")
	var hiText, loText string
	if strings.Contains(literal, ",") {
		parts := strings.SplitN(literal, ",", 2)
		hiText = strings.TrimSpace(parts[0])
		loText = strings.TrimSpace(parts[1])
	} else {
		compact := strings.ReplaceAll(literal, ",", "")
		if len(compact) > 32 || len(compact) <= 16 {
			return raw, fmt.Errorf("invalid 128-bit hex literal %q", input)
		}
		hiText = compact[:16]
		loText = compact[16:]
	}
	hi, err := strconv.ParseUint(hiText, 16, 64)
	if err != nil {
		return raw, err
	}
	lo, err := strconv.ParseUint(loText, 16, 64)
	if err != nil {
		return raw, err
	}
	raw.lo = lo
	raw.hi = hi
	return raw, nil
}

func formatGeneratedBIDStringBits128(raw BID_UINT128) string {
	return fmt.Sprintf("[%016x%016x]", raw.hi, raw.lo)
}

func normalizeGeneratedBIDStringBits(input string) string {
	trimmed := strings.ToLower(strings.TrimSpace(input))
	trimmed = strings.TrimSuffix(strings.TrimPrefix(trimmed, "["), "]")
	trimmed = strings.ReplaceAll(trimmed, ",", "")
	trimmed = strings.TrimLeft(trimmed, "0")
	if trimmed == "" {
		trimmed = "0"
	}
	return "[" + trimmed + "]"
}

func normalizeGeneratedBIDStringStatus(input string) string {
	trimmed := strings.ToUpper(strings.TrimSpace(input))
	trimmed = strings.TrimPrefix(trimmed, "0X")
	if trimmed == "" {
		return "00"
	}
	if len(trimmed)%2 != 0 {
		trimmed = "0" + trimmed
	}
	return trimmed
}
`, "@@STRING_SHARD_FILES@@", goBidStringShardFilesLiteral(shardFiles), 1)
	return replaceBidStringCountPlaceholders(src, counts)
}

func bidStringVectorsRustTestSource(counts bidStringVectorCounts, shardFiles []string) string {
	src := strings.Replace(genmarker.Line("testgen")+`

use bid754::bid64_from_string_raw;
use bid754::gen_types::BID_UINT128;
use bid754::generated::bid128_string::{bid128_from_string, bid128_to_string};
use bid754::generated::bid32_string::{bid32_from_string_raw, bid32_to_string_raw};
use bid754::generated::string64::bid64_to_string;
use serde::Deserialize;
use std::fs;
use std::path::PathBuf;

/// Readtest shard files of every from_string/to_string suite, relative to
/// generated/testspec, in generation order.
const GENERATED_BID_STRING_SHARD_FILES: &[&str] = &[
@@STRING_SHARD_FILES@@];

#[derive(Deserialize)]
struct ReadtestShard {
    format: String,
    function: String,
    kind: String,
    cases: Vec<ReadtestShardCase>,
}

#[derive(Deserialize)]
struct ReadtestShardCase {
    id: String,
    line: usize,
    operands: Vec<String>,
    expected: String,
    status: String,
    rounding: i64,
}

struct ReadCase {
    id: String,
    line: usize,
    format: String,
    function: String,
    kind: String,
    operands: Vec<String>,
    expected: String,
    status: String,
    rounding: i64,
}

#[derive(Debug, Default, PartialEq, Eq)]
struct GeneratedBIDStringCounts {
    total: usize,
    from_string: usize,
    to_string: usize,
    decimal32: usize,
    decimal64: usize,
    decimal128: usize,
    decimal32_from_string: usize,
    decimal64_from_string: usize,
    decimal128_from_string: usize,
    decimal32_to_string: usize,
    decimal64_to_string: usize,
    decimal128_to_string: usize,
}

const EXPECTED_GENERATED_BID_STRING_COUNTS: GeneratedBIDStringCounts = GeneratedBIDStringCounts {
    total: @@TOTAL@@,
    from_string: @@FROM_STRING@@,
    to_string: @@TO_STRING@@,
    decimal32: @@DECIMAL32@@,
    decimal64: @@DECIMAL64@@,
    decimal128: @@DECIMAL128@@,
    decimal32_from_string: @@DECIMAL32_FROM_STRING@@,
    decimal64_from_string: @@DECIMAL64_FROM_STRING@@,
    decimal128_from_string: @@DECIMAL128_FROM_STRING@@,
    decimal32_to_string: @@DECIMAL32_TO_STRING@@,
    decimal64_to_string: @@DECIMAL64_TO_STRING@@,
    decimal128_to_string: @@DECIMAL128_TO_STRING@@,
};

#[test]
fn test_generated_bid_string_vectors() {
    let read_cases = load_string_read_cases();
    let mut counts = GeneratedBIDStringCounts::default();
    for tc in read_cases.iter().filter(|tc| tc.kind == "from_string" || tc.kind == "to_string") {
        count_bid_string_case(&mut counts, tc);
        match tc.kind.as_str() {
            "from_string" => {
                let (got_bits, flags) = bid_string_from_string(tc);
                if tc.expected.trim_start().starts_with('[') {
                    assert_eq!(
                        normalize_bits(&got_bits),
                        normalize_bits(&tc.expected),
                        "{} line {} bits",
                        tc.function,
                        tc.line
                    );
                } else {
                    assert!(
                        bid_string_decimal_row_equal(
                            &tc.format,
                            &tc.expected,
                            &got_bits,
                            tc.rounding
                        ),
                        "{} line {} bits {:?}, want exact cohort parsed from {:?}",
                        tc.function,
                        tc.line,
                        got_bits,
                        tc.expected
                    );
                }
                assert_eq!(
                    normalize_status(&flags),
                    normalize_status(&tc.status),
                    "{} line {} status",
                    tc.function,
                    tc.line
                );
            }
            "to_string" => {
                let (got, flags) = bid_string_to_string(tc);
                let (equal, round_trip_status) = bid_string_to_string_row_equal(
                    &tc.format,
                    &tc.expected,
                    &got,
                    tc.rounding,
                );
                assert!(
                    equal,
                    "{} line {} result {:?}, want exact cohort parsed from {:?}",
                    tc.function,
                    tc.line,
                    got,
                    tc.expected
                );
                assert_eq!(
                    normalize_status(&combine_bid_string_status(&flags, &round_trip_status)),
                    normalize_status(&tc.status),
                    "{} line {} status",
                    tc.function,
                    tc.line
                );
            }
            _ => unreachable!(),
        }
    }
    assert_eq!(
        counts, EXPECTED_GENERATED_BID_STRING_COUNTS,
        "generated BID string read case counts changed"
    );
}

/// Anchors the generated comparator to Intel readtest.c: expected and produced
/// strings are parsed by the backend at the row rounding mode and compared as
/// exact BID cohorts, and produced-string parse flags accumulate into the
/// operation flags.
#[test]
fn test_bid_string_to_string_comparator_strength() {
    let (equal, _) =
        bid_string_to_string_row_equal("decimal64", "+15E+0", "+150E-1", 0);
    assert!(
        !equal,
        "wrong-cohort string accepted: +150E-1 must not match +15E+0"
    );

    let (equal, status) = bid_string_to_string_row_equal("decimal64", "15", "+15E+0", 0);
    assert!(equal, "exact-cohort spelling variant rejected");
    assert_eq!(normalize_status(&status), "00", "exact-cohort round-trip status");

    let (equal, status) = bid_string_to_string_row_equal(
        "decimal64",
        "5000000000000000E-15",
        "5.0000000000000000001",
        0,
    );
    assert!(
        equal,
        "inexact produced string did not round to the expected cohort"
    );
    assert_eq!(normalize_status(&status), "20", "inexact round-trip status");
    assert_eq!(
        normalize_status(&combine_bid_string_status("01", &status)),
        "21",
        "combined status"
    );
}

fn count_bid_string_case(counts: &mut GeneratedBIDStringCounts, tc: &ReadCase) {
    counts.total += 1;
    match tc.kind.as_str() {
        "from_string" => counts.from_string += 1,
        "to_string" => counts.to_string += 1,
        _ => {}
    }
    match tc.format.as_str() {
        "decimal32" => {
            counts.decimal32 += 1;
            match tc.kind.as_str() {
                "from_string" => counts.decimal32_from_string += 1,
                "to_string" => counts.decimal32_to_string += 1,
                _ => {}
            }
        }
        "decimal64" => {
            counts.decimal64 += 1;
            match tc.kind.as_str() {
                "from_string" => counts.decimal64_from_string += 1,
                "to_string" => counts.decimal64_to_string += 1,
                _ => {}
            }
        }
        "decimal128" => {
            counts.decimal128 += 1;
            match tc.kind.as_str() {
                "from_string" => counts.decimal128_from_string += 1,
                "to_string" => counts.decimal128_to_string += 1,
                _ => {}
            }
        }
        _ => {}
    }
}

fn testspec_dir() -> PathBuf {
    PathBuf::from(env!("CARGO_MANIFEST_DIR")).join("../devtools/generated/testspec")
}

fn load_string_read_cases() -> Vec<ReadCase> {
    let base_dir = testspec_dir();
    let mut read_cases = Vec::new();
    for shard_file in GENERATED_BID_STRING_SHARD_FILES {
        let path = base_dir.join(shard_file);
        let data = fs::read_to_string(&path)
            .unwrap_or_else(|err| panic!("read string shard {}: {err}", path.display()));
        let shard: ReadtestShard = serde_json::from_str(&data)
            .unwrap_or_else(|err| panic!("parse string shard {}: {err}", path.display()));
        for tc in shard.cases {
            read_cases.push(ReadCase {
                id: tc.id,
                line: tc.line,
                format: shard.format.clone(),
                function: shard.function.clone(),
                kind: shard.kind.clone(),
                operands: tc.operands,
                expected: tc.expected,
                status: tc.status,
                rounding: tc.rounding,
            });
        }
    }
    read_cases
}

fn bid_string_from_string(tc: &ReadCase) -> (String, String) {
    assert_eq!(tc.operands.len(), 1, "{} expects one operand", tc.id);
    bid_string_parse_value(&tc.format, &tc.operands[0], tc.rounding)
}

fn bid_string_parse_value(format: &str, input: &str, rounding: i64) -> (String, String) {
    match format {
        "decimal32" => {
            let (got, flags) = bid32_from_string_raw(input, rounding);
            (format!("[{got:08x}]"), format!("{flags:02X}"))
        }
        "decimal64" => {
            let rounding = i32::try_from(rounding).expect("decimal64 rounding mode fits i32");
            let (got, flags) = bid64_from_string_raw(input, rounding);
            (format!("[{got:016x}]"), format!("{flags:02X}"))
        }
        "decimal128" => {
            let (got, flags) = bid128_from_string(input, rounding);
            (format_bits128(got), format!("{flags:02X}"))
        }
        other => panic!("unsupported format {other}"),
    }
}

fn bid_string_to_string(tc: &ReadCase) -> (String, String) {
    assert_eq!(tc.operands.len(), 1, "{} expects one operand", tc.id);
    match tc.format.as_str() {
        "decimal32" => (bid32_to_string_raw(parse_u32_bits(&tc.operands[0])), String::from("00")),
        "decimal64" => (bid64_to_string(parse_u64_bits(&tc.operands[0])), String::from("00")),
        "decimal128" => (bid128_to_string(parse_bits128(&tc.operands[0])), String::from("00")),
        other => panic!("unsupported format {other}"),
    }
}

fn bid_string_decimal_row_equal(
    format: &str,
    expected: &str,
    got_bits: &str,
    rounding: i64,
) -> bool {
    let (expected_bits, _) = bid_string_parse_value(format, expected, rounding);
    normalize_bits(&expected_bits) == normalize_bits(got_bits)
}

fn bid_string_to_string_row_equal(
    format: &str,
    expected: &str,
    got: &str,
    rounding: i64,
) -> (bool, String) {
    let (expected_bits, _) = bid_string_parse_value(format, expected, rounding);
    let (got_bits, round_trip_status) = bid_string_parse_value(format, got, rounding);
    (
        normalize_bits(&expected_bits) == normalize_bits(&got_bits),
        round_trip_status,
    )
}

fn combine_bid_string_status(a: &str, b: &str) -> String {
    let flags_a = u32::from_str_radix(&normalize_status(a), 16)
        .unwrap_or_else(|err| panic!("parse status {a:?}: {err}"));
    let flags_b = u32::from_str_radix(&normalize_status(b), 16)
        .unwrap_or_else(|err| panic!("parse status {b:?}: {err}"));
    format!("{:02X}", flags_a | flags_b)
}

fn parse_u32_bits(input: &str) -> u32 {
    u32::from_str_radix(&normalize_hex(input), 16).unwrap()
}

fn parse_u64_bits(input: &str) -> u64 {
    u64::from_str_radix(&normalize_hex(input), 16).unwrap()
}

fn parse_bits128(input: &str) -> BID_UINT128 {
    let compact = normalize_hex(input);
    assert!(compact.len() > 16 && compact.len() <= 32, "invalid BID128 literal {input}");
    let padded = format!("{compact:0>32}");
    let hi = u64::from_str_radix(&padded[..16], 16).unwrap();
    let lo = u64::from_str_radix(&padded[16..], 16).unwrap();
    BID_UINT128 { lo, hi }
}

fn format_bits128(raw: BID_UINT128) -> String {
    format!("[{:016x}{:016x}]", raw.hi, raw.lo)
}

fn normalize_hex(input: &str) -> String {
    input.trim().trim_start_matches('[').trim_end_matches(']').replace(',', "").to_ascii_lowercase()
}

fn normalize_bits(input: &str) -> String {
    let trimmed = normalize_hex(input);
    let stripped = trimmed.trim_start_matches('0');
    if stripped.is_empty() {
        String::from("[0]")
    } else {
        format!("[{stripped}]")
    }
}

fn normalize_status(input: &str) -> String {
    let mut s = input.trim().trim_start_matches("0x").trim_start_matches("0X").to_ascii_uppercase();
    if s.is_empty() {
        return String::from("00");
    }
    if s.len() % 2 != 0 {
        s = format!("0{s}");
    }
    s
}
`, "@@STRING_SHARD_FILES@@", rustBidStringShardFilesLiteral(shardFiles), 1)
	return replaceBidStringCountPlaceholders(src, counts)
}
