package testgen

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// verificationAnchors mirrors devtools/verification_anchors.json, the
// hand-maintained case-count anchor file that lives outside every generation
// path. Generated artifacts pin their own counts, so a generator regression
// that shrinks a generated case set also shrinks the generated pins and
// passes silently; this test compares the checked-in generated artifacts
// against the external anchors instead.
type verificationAnchors struct {
	Comment                                      []string                     `json:"comment"`
	ReadtestCasesTotal                           int                          `json:"readtest_cases_total"`
	ReadtestStatusControlCases                   int                          `json:"readtest_status_control_cases"`
	ReadtestNativeCompareSkipCases               int                          `json:"readtest_native_compare_skip_cases"`
	ReadtestProfileRowsAccepted                  int                          `json:"readtest_profile_rows_accepted"`
	ReadtestProfileRowSkipReasons                map[string]int               `json:"readtest_profile_row_skip_reasons"`
	GoportReadtestExecutedCases                  int                          `json:"goport_readtest_executed_cases"`
	FFIBitcompareCasesTotal                      int                          `json:"ffi_bitcompare_cases_total"`
	Tier1ArithmeticBoundaryValues                map[string]uint64            `json:"tier1_arithmetic_long_boundary_values_by_width"`
	Tier1ArithmeticSemanticRounded               map[string]uint64            `json:"tier1_arithmetic_long_semantic_rounded_pairs_by_width"`
	Tier1ArithmeticSemanticScale                 map[string]uint64            `json:"tier1_arithmetic_long_semantic_scale_cases_by_width"`
	Tier1ArithmeticSemanticRemainder             map[string]uint64            `json:"tier1_arithmetic_long_semantic_remainder_pairs_by_width"`
	Tier1ArithmeticSemanticFma                   map[string]uint64            `json:"tier1_arithmetic_long_semantic_fma_triples_by_width"`
	Tier1ArithmeticSemanticSqrt                  map[string]uint64            `json:"tier1_arithmetic_long_semantic_sqrt_cases_by_width"`
	Tier1ArithmeticStructuredCases               map[string]uint64            `json:"tier1_arithmetic_long_structured_comparisons_by_width"`
	Tier1ArithmeticRandomOperations              uint64                       `json:"tier1_arithmetic_long_random_operations"`
	Tier1ArithmeticRandomCasesPerOp              map[string]uint64            `json:"tier1_arithmetic_long_random_cases_per_operation_by_width"`
	Tier1ArithmeticScaleFiniteTransitionLimits   map[string]uint64            `json:"tier1_arithmetic_long_scale_finite_transition_limit_by_width"`
	Tier1ArithmeticScaleRandomStrata             uint64                       `json:"tier1_arithmetic_long_scale_random_strata"`
	Tier1ArithmeticScaleModeCross                uint64                       `json:"tier1_arithmetic_long_scale_meaningful_mode_cross"`
	Tier1ArithmeticScaleModeCrossGroups          map[string]uint64            `json:"tier1_arithmetic_long_scale_mode_cross_groups_by_width"`
	Tier1ArithmeticScaleTupleHashes              map[string]uint64            `json:"tier1_arithmetic_long_scale_tuple_hash_by_width"`
	Tier1ArithmeticPairStreamHashes              map[string]uint64            `json:"tier1_arithmetic_long_pair_stream_hash_by_width"`
	Tier1ArithmeticFmaTripleStreamHashes         map[string]uint64            `json:"tier1_arithmetic_long_fma_triple_stream_hash_by_width"`
	Tier1ArithmeticRandomStreamHashes            map[string]uint64            `json:"tier1_arithmetic_long_random_stream_hash_by_width"`
	Tier1ArithmeticTotalComparisons              map[string]uint64            `json:"tier1_arithmetic_long_total_comparisons_by_width"`
	Tier1ArithmeticLongConsumers                 uint64                       `json:"tier1_arithmetic_long_consumers"`
	Tier1CompareConversionBoundaryValues         map[string]uint64            `json:"tier1_compare_conversion_long_boundary_values_by_width"`
	Tier1CompareConversionSemanticValues         map[string]uint64            `json:"tier1_compare_conversion_long_semantic_values_by_width"`
	Tier1CompareConversionQuietPredicates        uint64                       `json:"tier1_compare_conversion_long_quiet_predicates"`
	Tier1CompareConversionMinMaxOperations       uint64                       `json:"tier1_compare_conversion_long_minmax_operations"`
	Tier1CompareConversionComparisonStructured   map[string]uint64            `json:"tier1_compare_conversion_long_comparison_structured_by_width"`
	Tier1CompareConversionComparisonRandom       map[string]uint64            `json:"tier1_compare_conversion_long_comparison_random_by_width"`
	Tier1CompareConversionComparisonRandomHashes map[string]uint64            `json:"tier1_compare_conversion_long_comparison_random_stream_hash_by_width"`
	Tier1CompareConversionComparisonTotal        map[string]uint64            `json:"tier1_compare_conversion_long_comparison_total_by_width"`
	Tier1CompareConversionToIntegerOperations    uint64                       `json:"tier1_compare_conversion_long_to_integer_operations"`
	Tier1CompareConversionToIntegerTotal         map[string]uint64            `json:"tier1_compare_conversion_long_to_integer_total_by_width"`
	Tier1CompareConversionToIntegerRandomHashes  map[string]uint64            `json:"tier1_compare_conversion_long_to_integer_random_stream_hash_by_width"`
	Tier1CompareConversionWidthOperations        map[string]uint64            `json:"tier1_compare_conversion_long_width_operations_by_source"`
	Tier1CompareConversionWidthTotal             map[string]uint64            `json:"tier1_compare_conversion_long_width_total_by_source"`
	Tier1CompareConversionWidthRandomHashes      map[string]uint64            `json:"tier1_compare_conversion_long_width_random_stream_hash_by_width"`
	Tier1CompareConversionBinaryRandomHashes     map[string]uint64            `json:"tier1_compare_conversion_long_binary_random_stream_hash_by_width"`
	Tier1CompareConversionConstructorOperations  uint64                       `json:"tier1_compare_conversion_long_constructor_operations"`
	Tier1CompareConversionConstructorTotal       uint64                       `json:"tier1_compare_conversion_long_constructor_total"`
	Tier1CompareConversionConstructorConvenience uint64                       `json:"tier1_compare_conversion_long_constructor_convenience_checks"`
	Tier1CompareConversionConstructorRandomHash  uint64                       `json:"tier1_compare_conversion_long_constructor_random_stream_hash"`
	Tier1CompareConversionStructured             uint64                       `json:"tier1_compare_conversion_long_conversion_structured"`
	Tier1CompareConversionRandom                 uint64                       `json:"tier1_compare_conversion_long_conversion_random"`
	Tier1CompareConversionTotal                  uint64                       `json:"tier1_compare_conversion_long_conversion_total"`
	Tier1CompareConversionLongConsumers          uint64                       `json:"tier1_compare_conversion_long_consumers"`
	DectestSuiteCases                            map[string]int               `json:"dectest_suite_cases"`
	GoportDectestExecutedCases                   map[string]int               `json:"goport_dectest_executed_cases"`
	GoportDectestSkippedCases                    map[string]int               `json:"goport_dectest_skipped_cases"`
	GoportDectestFlagExemptCases                 map[string]int               `json:"goport_dectest_flag_exempt_cases"`
	RustDectestExecutedCases                     map[string]int               `json:"rust_dectest_executed_cases"`
	RustDectestSkippedCases                      map[string]int               `json:"rust_dectest_skipped_cases"`
	RustDectestFlagExemptCases                   map[string]int               `json:"rust_dectest_flag_exempt_cases"`
	ReadtestProfileFunctionsTotal                int                          `json:"readtest_profile_functions_total"`
	ReadtestProfileFunctionsSelected             int                          `json:"readtest_profile_functions_selected"`
	GoportReadtestDispatchedFuncs                int                          `json:"goport_readtest_dispatched_functions"`
	RustReadtestDispatchedFuncs                  int                          `json:"rust_readtest_dispatched_functions"`
	RustReadtestSuitePasses                      map[string]int               `json:"rust_readtest_suite_passes"`
	BidCodecVectorsTotal                         int                          `json:"bid_codec_vectors_total"`
	BidCodecVectorsByWidth                       map[string]int               `json:"bid_codec_vectors_by_width"`
	BidCodecCanonicalByWidth                     map[string]int               `json:"bid_codec_canonical_vectors_by_width"`
	BidCodecRejectVectorsTotal                   int                          `json:"bid_codec_reject_vectors_total"`
	BidCodecRejectChannels                       map[string]int               `json:"bid_codec_reject_vectors_channels"`
	BidCodecRejectRequires                       map[string]int               `json:"bid_codec_reject_vectors_requires"`
	BidCodecRejectConsumedByLanguage             map[string]int               `json:"bid_codec_reject_vectors_consumed_by_language"`
	BidCodecRejectGoFullConsumed                 int                          `json:"bid_codec_reject_vectors_go_full_consumed"`
	BidCodecRejectGoFullChannelSkipped           int                          `json:"bid_codec_reject_vectors_go_full_channel_skipped"`
	BidCodecRejectRustFullParseConsumed          int                          `json:"bid_codec_reject_vectors_rust_full_parse_consumed"`
	BidCodecRejectRustFullParseChannelSkipped    int                          `json:"bid_codec_reject_vectors_rust_full_parse_channel_skipped"`
	BidCodecStringGoFullConsumed                 int                          `json:"bid_codec_string_vectors_go_full_consumed"`
	BidCodecStringVectorsTotal                   int                          `json:"bid_codec_string_vectors_total"`
	BidCodecStringConsumedByLanguage             map[string]int               `json:"bid_codec_string_vectors_consumed_by_language"`
	BidCodecD32ExhaustiveRawPatterns             uint64                       `json:"bid_codec_decimal32_exhaustive_raw_patterns"`
	BidCodecD32StructuredCases                   uint64                       `json:"bid_codec_decimal32_structured_cases"`
	BidCodecD32ExhaustiveRawClasses              map[string]uint64            `json:"bid_codec_decimal32_exhaustive_raw_classes"`
	BidCodecD32StructuredClasses                 map[string]uint64            `json:"bid_codec_decimal32_structured_classes"`
	BidCodecD64128DifferentialCases              uint64                       `json:"bid_codec_decimal64_128_differential_cases_per_width"`
	BidCodecD64128DifferentialClass              map[string]map[string]uint64 `json:"bid_codec_decimal64_128_differential_classes"`
	BidCodecD64128BoundaryRawCases               map[string]uint64            `json:"bid_codec_decimal64_128_boundary_raw_cases"`
	BidCodecD64128StructuredCases                map[string]uint64            `json:"bid_codec_decimal64_128_structured_cases"`
	BidCodecD64128StructuredClasses              map[string]map[string]uint64 `json:"bid_codec_decimal64_128_structured_classes"`
	PublicAPISymbolsTotal                        int                          `json:"public_api_symbols_total"`
	PublicAPIParityWrappers                      int                          `json:"public_api_parity_wrappers"`
	PublicAPIParityExcluded                      int                          `json:"public_api_parity_excluded"`
	PublicAPIParityCasesTotal                    int                          `json:"public_api_parity_cases_total"`
	PublicAPIFlaglessSiblingTargets              int                          `json:"public_api_flagless_sibling_targets"`
	PublicAPIFlaglessSiblingCasesTotal           int                          `json:"public_api_flagless_sibling_cases_total"`
	RustPublicAPIParityWrappers                  int                          `json:"rust_public_api_parity_wrappers"`
	RustPublicAPIDirectRoutingWrappers           int                          `json:"rust_public_api_direct_routing_wrappers"`
	RustPublicAPIDelegatedRoutingWrappers        int                          `json:"rust_public_api_delegated_routing_wrappers"`
	RustPublicAPIParityCasesTotal                int                          `json:"rust_public_api_parity_cases_total"`
	RustPublicAPIConstantsTotal                  int                          `json:"rust_public_api_constants_total"`
	RustPublicAPIFlaglessSiblingTargets          int                          `json:"rust_public_api_flagless_sibling_targets"`
	RustPublicAPIFlaglessSiblingCasesTotal       int                          `json:"rust_public_api_flagless_sibling_cases_total"`
	RustPublicAPIParityCasesByShape              map[string]int               `json:"rust_public_api_parity_cases_by_shape"`
	DecnumberDiffBoundaryValues                  map[string]uint64            `json:"decnumber_differential_boundary_values_by_width"`
	DecnumberDiffProbeValues                     map[string]uint64            `json:"decnumber_differential_probe_values_by_width"`
	DecnumberDiffExactProductValues              map[string]uint64            `json:"decnumber_differential_exact_product_values_by_width"`
	DecnumberDiffExactProductAddends             map[string]uint64            `json:"decnumber_differential_exact_product_addends_by_width"`
	DecnumberDiffRandomPairsPerOp                uint64                       `json:"decnumber_differential_random_pairs_per_operation"`
	DecnumberDiffStructuredComparisons           map[string]uint64            `json:"decnumber_differential_structured_comparisons_by_width"`
	DecnumberDiffStructuredFmaExcluded           map[string]uint64            `json:"decnumber_differential_structured_fma_excluded_by_width"`
	DecnumberDiffStructuredKnownDivergences      map[string]uint64            `json:"decnumber_differential_structured_known_divergences_by_width"`
	DecnumberDiffStructuredStreamHashes          map[string]uint64            `json:"decnumber_differential_structured_stream_hash_by_width"`
	DecnumberDiffRandomComparisons               map[string]uint64            `json:"decnumber_differential_random_comparisons_by_width"`
	DecnumberDiffRandomFmaExcluded               map[string]uint64            `json:"decnumber_differential_random_fma_excluded_by_width"`
	DecnumberDiffRandomKnownDivergences          map[string]uint64            `json:"decnumber_differential_random_known_divergences_by_width"`
	DecnumberDiffRandomStreamHashes              map[string]uint64            `json:"decnumber_differential_random_stream_hash_by_width"`
	DecnumberDiffTotalComparisons                map[string]uint64            `json:"decnumber_differential_total_comparisons_by_width"`
	DecnumberDiffConsumers                       uint64                       `json:"decnumber_differential_consumers"`
	D32ExhaustiveOperations                      uint64                       `json:"d32_exhaustive_unary_operations"`
	D32ExhaustiveLanes                           uint64                       `json:"d32_exhaustive_unary_lanes"`
	D32ExhaustiveCasesPerLane                    uint64                       `json:"d32_exhaustive_unary_cases_per_lane"`
	D32ExhaustiveTotalComparisons                uint64                       `json:"d32_exhaustive_unary_total_comparisons"`
	D32ExhaustiveDigestByLane                    map[string]uint64            `json:"d32_exhaustive_unary_result_digest_by_lane"`
	D32ExhaustiveConsumers                       uint64                       `json:"d32_exhaustive_unary_consumers"`
	VerificationArtifactSHA256                   map[string]string            `json:"verification_artifact_sha256"`
}

type dispatchInventoryCounts struct {
	Dispatched int `json:"dispatched"`
}

func loadVerificationAnchors(t *testing.T) verificationAnchors {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "verification_anchors.json"))
	if err != nil {
		t.Fatalf("read verification_anchors.json: %v", err)
	}
	var anchors verificationAnchors
	decoder := json.NewDecoder(bytes.NewReader(raw))
	// A key this struct does not declare is an anchor nothing enforces;
	// fail closed instead of silently carrying dead or misspelled anchors.
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&anchors); err != nil {
		t.Fatalf("unmarshal verification_anchors.json (unknown keys are rejected): %v", err)
	}
	return anchors
}

// verificationSentinels mirrors devtools/verification_sentinels.json, the
// hand-maintained known-answer row pin that lives outside every
// generation path (GUARDRAILS: no generator reads or writes it). The rows
// are the anchor payload themselves — not counts — so the comparison below
// requires exact, ordered, byte-equal row sets across the pin file and both
// generated runner literals.
type verificationSentinels struct {
	Comment                           []string `json:"comment"`
	Tier1ArithmeticRoutingRows        []string `json:"tier1_arithmetic_long_routing_sentinel_rows"`
	Tier1CompareConversionRoutingRows []string `json:"tier1_compare_conversion_long_routing_sentinel_rows"`
	MixedFMAFusednessRows             []string `json:"mixed_fma_fusedness_rows"`
	MixedFormatFFIRoutingRows         []string `json:"mixed_format_ffi_routing_sentinel_rows"`
	DecnumberDifferentialRows         []string `json:"decnumber_differential_sentinel_rows"`
	D32ExhaustiveRows                 []string `json:"d32_exhaustive_sentinel_rows"`
}

func loadVerificationSentinels(t *testing.T) verificationSentinels {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "verification_sentinels.json"))
	if err != nil {
		t.Fatalf("read verification_sentinels.json: %v", err)
	}
	var sentinels verificationSentinels
	decoder := json.NewDecoder(bytes.NewReader(raw))
	// A key this struct does not declare is a pin nothing enforces; fail
	// closed instead of silently carrying dead or misspelled sentinel keys.
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&sentinels); err != nil {
		t.Fatalf("unmarshal verification_sentinels.json (unknown keys are rejected): %v", err)
	}
	return sentinels
}

// loadGeneratedGoStringSliceLiteral extracts the ordered string elements of
// one `var <name> = []string{...}` declaration from a generated Go artifact.
// The existing scalar extractor evaluates integer constants only, so the
// sentinel rows need this dedicated string-slice reader.
func loadGeneratedGoStringSliceLiteral(t *testing.T, path, varName string) []string {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		t.Fatalf("parse generated artifact %s: %v", path, err)
	}
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.VAR {
			continue
		}
		for _, spec := range gen.Specs {
			values, ok := spec.(*ast.ValueSpec)
			if !ok || len(values.Names) != 1 || values.Names[0].Name != varName || len(values.Values) != 1 {
				continue
			}
			composite, ok := values.Values[0].(*ast.CompositeLit)
			if !ok {
				t.Fatalf("generated %s: %s is not a composite literal", path, varName)
			}
			rows := make([]string, 0, len(composite.Elts))
			for _, element := range composite.Elts {
				literal, ok := element.(*ast.BasicLit)
				if !ok || literal.Kind != token.STRING {
					t.Fatalf("generated %s: %s carries a non-string element %T", path, varName, element)
				}
				row, err := strconv.Unquote(literal.Value)
				if err != nil {
					t.Fatalf("generated %s: %s element %s: %v", path, varName, literal.Value, err)
				}
				rows = append(rows, row)
			}
			return rows
		}
	}
	t.Fatalf("generated %s: string-slice literal %s not found", path, varName)
	return nil
}

// loadGeneratedGoUintConstant extracts one `const <name> = uint64(<n>)` value
// from a generated Go artifact so a count constant can be bound to the literal
// length outside the native runtime path.
func loadGeneratedGoUintConstant(t *testing.T, path, name string) uint64 {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		t.Fatalf("parse generated artifact %s: %v", path, err)
	}
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.CONST {
			continue
		}
		for _, spec := range gen.Specs {
			values, ok := spec.(*ast.ValueSpec)
			if !ok || len(values.Names) != 1 || values.Names[0].Name != name || len(values.Values) != 1 {
				continue
			}
			value, err := evalGeneratedUintConstant(values.Values[0], map[string]uint64{})
			if err != nil {
				t.Fatalf("generated %s: constant %s: %v", path, name, err)
			}
			return value
		}
	}
	t.Fatalf("generated %s: uint64 constant %s not found", path, name)
	return 0
}

var rustRoutingSentinelRowRe = regexp.MustCompile(`^\s*"([ -~]+)",$`)

// loadRustStringArrayLiteral extracts an ordered generated Rust [&str; N]
// literal and requires its declared length to match the extracted row count.
func loadRustStringArrayLiteral(t *testing.T, path, constName string) []string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read generated Rust artifact %s: %v", path, err)
	}
	headerRe := regexp.MustCompile(`^const ` + regexp.QuoteMeta(constName) + `: \[&str; (\d+)\] = \[$`)
	lines := strings.Split(string(raw), "\n")
	declared := -1
	rows := []string(nil)
	for i := 0; i < len(lines); i++ {
		header := headerRe.FindStringSubmatch(lines[i])
		if header == nil {
			continue
		}
		if declared >= 0 {
			t.Fatalf("generated %s: %s declared twice", path, constName)
		}
		count, err := strconv.Atoi(header[1])
		if err != nil {
			t.Fatalf("generated %s: bad %s length %q: %v", path, constName, header[1], err)
		}
		declared = count
		rows = []string{}
		for i++; i < len(lines); i++ {
			if strings.TrimSpace(lines[i]) == "];" {
				break
			}
			// The generated token-injection layout leaves one blank line
			// before the closing bracket; it carries no row.
			if strings.TrimSpace(lines[i]) == "" {
				continue
			}
			match := rustRoutingSentinelRowRe.FindStringSubmatch(lines[i])
			if match == nil {
				t.Fatalf("generated %s: unparseable %s line %q", path, constName, lines[i])
			}
			rows = append(rows, match[1])
		}
	}
	if declared < 0 {
		t.Fatalf("generated %s: %s block not found", path, constName)
	}
	if len(rows) != declared {
		t.Fatalf("generated %s: %s declares %d rows but carries %d", path, constName, declared, len(rows))
	}
	return rows
}

// loadRustRoutingSentinelRows extracts the Tier 1 routing rows from their
// fixed generated ffi-verify location.
func loadRustRoutingSentinelRows(t *testing.T, filename string) []string {
	t.Helper()
	return loadRustStringArrayLiteral(t,
		filepath.Join("..", "..", "..", "bid754-rs", "ffi-verify", "tests", filename),
		"ROUTING_SENTINEL_ROWS")
}

// firstSentinelRowDivergence renders the first index where two sentinel row
// lists differ, as a diagnostic suffix for the three-way comparison.
func firstSentinelRowDivergence(got, want []string) string {
	limit := len(got)
	if len(want) < limit {
		limit = len(want)
	}
	for i := 0; i < limit; i++ {
		if got[i] != want[i] {
			return fmt.Sprintf("; first divergence at row %d: generated %q, pinned %q", i, got[i], want[i])
		}
	}
	return ""
}

type bidCodecD32HarnessInventory struct {
	RawPatterns       uint64
	StructuredCases   uint64
	RawClasses        map[string]uint64
	StructuredClasses map[string]uint64
}

func evalGeneratedUintConstant(expr ast.Expr, values map[string]uint64) (uint64, error) {
	switch e := expr.(type) {
	case *ast.BasicLit:
		if e.Kind != token.INT {
			return 0, fmt.Errorf("literal %q is not an integer", e.Value)
		}
		value, err := strconv.ParseUint(e.Value, 0, 64)
		if err != nil {
			return 0, fmt.Errorf("parse integer literal %q: %w", e.Value, err)
		}
		return value, nil
	case *ast.Ident:
		value, ok := values[e.Name]
		if !ok {
			return 0, fmt.Errorf("constant %q is not defined before use", e.Name)
		}
		return value, nil
	case *ast.ParenExpr:
		return evalGeneratedUintConstant(e.X, values)
	case *ast.CallExpr:
		name, ok := e.Fun.(*ast.Ident)
		if !ok || name.Name != "uint64" || len(e.Args) != 1 {
			return 0, fmt.Errorf("unsupported constant conversion %T", e.Fun)
		}
		return evalGeneratedUintConstant(e.Args[0], values)
	case *ast.BinaryExpr:
		left, err := evalGeneratedUintConstant(e.X, values)
		if err != nil {
			return 0, err
		}
		right, err := evalGeneratedUintConstant(e.Y, values)
		if err != nil {
			return 0, err
		}
		switch e.Op {
		case token.ADD:
			result := left + right
			if result < left {
				return 0, fmt.Errorf("uint64 addition overflow: %d + %d", left, right)
			}
			return result, nil
		case token.MUL:
			if left != 0 && right > ^uint64(0)/left {
				return 0, fmt.Errorf("uint64 multiplication overflow: %d * %d", left, right)
			}
			return left * right, nil
		case token.SHL:
			if right >= 64 || left > ^uint64(0)>>right {
				return 0, fmt.Errorf("uint64 left shift overflow: %d << %d", left, right)
			}
			return left << right, nil
		case token.OR:
			return left | right, nil
		default:
			return 0, fmt.Errorf("unsupported constant operator %s", e.Op)
		}
	default:
		return 0, fmt.Errorf("unsupported constant expression %T", expr)
	}
}

type bidCodecD64128HarnessInventory struct {
	DifferentialCasesPerWidth uint64
	DifferentialClasses       map[string]map[string]uint64
	BoundaryRawCases          map[string]uint64
	StructuredCases           map[string]uint64
	StructuredClasses         map[string]map[string]uint64
}

type tier1ArithmeticLongInventory struct {
	BoundaryValues              map[string]uint64
	SemanticRounded             map[string]uint64
	SemanticScale               map[string]uint64
	SemanticRemainder           map[string]uint64
	SemanticFma                 map[string]uint64
	SemanticSqrt                map[string]uint64
	StructuredCases             map[string]uint64
	RandomOperations            uint64
	RandomCasesPerOp            map[string]uint64
	ScaleFiniteTransitionLimits map[string]uint64
	ScaleRandomStrata           uint64
	ScaleModeCross              uint64
	ScaleModeCrossGroups        map[string]uint64
	ScaleTupleHashes            map[string]uint64
	PairStreamHashes            map[string]uint64
	FmaTripleStreamHashes       map[string]uint64
	RandomStreamHashes          map[string]uint64
	TotalComparisons            map[string]uint64
}

func loadTier1ArithmeticLongInventory(t *testing.T) tier1ArithmeticLongInventory {
	t.Helper()
	path := filepath.Join("..", "..", "..", "bid754-go", "generated_ffi_bitcompare_tier1_arithmetic_long_test.go")
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		t.Fatalf("parse generated Tier 1 arithmetic long harness: %v", err)
	}
	constants := map[string]uint64{}
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.CONST {
			continue
		}
		for _, spec := range gen.Specs {
			values, ok := spec.(*ast.ValueSpec)
			if !ok || len(values.Names) != len(values.Values) {
				t.Fatalf("generated Tier 1 arithmetic long harness has an implicit or malformed const declaration")
			}
			for i, name := range values.Names {
				if !strings.HasPrefix(name.Name, "tier1Arithmetic") {
					continue
				}
				value, err := evalGeneratedUintConstant(values.Values[i], constants)
				if err != nil {
					t.Fatalf("evaluate generated Tier 1 arithmetic constant %s: %v", name.Name, err)
				}
				constants[name.Name] = value
			}
		}
	}
	require := func(name string) uint64 {
		t.Helper()
		value, ok := constants[name]
		if !ok {
			t.Fatalf("generated Tier 1 arithmetic long harness is missing constant %s", name)
		}
		return value
	}
	byWidth := func(prefix string) map[string]uint64 {
		return map[string]uint64{
			"decimal32":  require(prefix + "32"),
			"decimal64":  require(prefix + "64"),
			"decimal128": require(prefix + "128"),
		}
	}
	return tier1ArithmeticLongInventory{
		BoundaryValues: map[string]uint64{
			"decimal32":  require("tier1ArithmeticBoundary32Count"),
			"decimal64":  require("tier1ArithmeticBoundary64Count"),
			"decimal128": require("tier1ArithmeticBoundary128Count"),
		},
		SemanticRounded: map[string]uint64{
			"decimal32":  require("tier1ArithmeticSemanticRounded32Count"),
			"decimal64":  require("tier1ArithmeticSemanticRounded64Count"),
			"decimal128": require("tier1ArithmeticSemanticRounded128Count"),
		},
		SemanticScale: map[string]uint64{
			"decimal32":  require("tier1ArithmeticSemanticScale32Count"),
			"decimal64":  require("tier1ArithmeticSemanticScale64Count"),
			"decimal128": require("tier1ArithmeticSemanticScale128Count"),
		},
		SemanticRemainder: map[string]uint64{
			"decimal32":  require("tier1ArithmeticSemanticRemainder32Count"),
			"decimal64":  require("tier1ArithmeticSemanticRemainder64Count"),
			"decimal128": require("tier1ArithmeticSemanticRemainder128Count"),
		},
		SemanticFma: map[string]uint64{
			"decimal32":  require("tier1ArithmeticSemanticFma32Count"),
			"decimal64":  require("tier1ArithmeticSemanticFma64Count"),
			"decimal128": require("tier1ArithmeticSemanticFma128Count"),
		},
		SemanticSqrt: map[string]uint64{
			"decimal32":  require("tier1ArithmeticSemanticSqrt32Count"),
			"decimal64":  require("tier1ArithmeticSemanticSqrt64Count"),
			"decimal128": require("tier1ArithmeticSemanticSqrt128Count"),
		},
		StructuredCases:             byWidth("tier1ArithmeticStructuredComparisons"),
		RandomOperations:            require("tier1ArithmeticRandomOps"),
		RandomCasesPerOp:            byWidth("tier1ArithmeticRandomCasesPerOp"),
		ScaleFiniteTransitionLimits: byWidth("tier1ArithmeticScaleFiniteTransitionLimit"),
		ScaleRandomStrata:           require("tier1ArithmeticScaleRandomStrata"),
		ScaleModeCross:              require("tier1ArithmeticScaleModeCross"),
		ScaleModeCrossGroups:        byWidth("tier1ArithmeticScaleModeCrossGroups"),
		ScaleTupleHashes:            byWidth("tier1ArithmeticScaleTupleHash"),
		PairStreamHashes:            byWidth("tier1ArithmeticPairStreamHash"),
		FmaTripleStreamHashes:       byWidth("tier1ArithmeticFmaTripleStreamHash"),
		RandomStreamHashes:          byWidth("tier1ArithmeticRandomStreamHash"),
		TotalComparisons:            byWidth("tier1ArithmeticTotalComparisons"),
	}
}

type tier1CompareConversionLongInventory struct {
	BoundaryValues         map[string]uint64
	SemanticValues         map[string]uint64
	QuietPredicates        uint64
	MinMaxOperations       uint64
	ComparisonStructured   map[string]uint64
	ComparisonRandom       map[string]uint64
	ComparisonRandomHashes map[string]uint64
	ComparisonTotal        map[string]uint64
	ToIntegerOperations    uint64
	ToIntegerTotal         map[string]uint64
	ToIntegerRandomHashes  map[string]uint64
	WidthOperations        map[string]uint64
	WidthTotal             map[string]uint64
	WidthRandomHashes      map[string]uint64
	BinaryRandomHashes     map[string]uint64
	ConstructorOperations  uint64
	ConstructorTotal       uint64
	ConstructorConvenience uint64
	ConstructorRandomHash  uint64
	ConversionStructured   uint64
	ConversionRandom       uint64
	ConversionTotal        uint64
}

func loadTier1CompareConversionLongInventory(t *testing.T) tier1CompareConversionLongInventory {
	t.Helper()
	path := filepath.Join("..", "..", "..", "bid754-go", "generated_ffi_bitcompare_tier1_compare_conversion_long_test.go")
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		t.Fatalf("parse generated Tier 1 compare/conversion long harness: %v", err)
	}
	constants := map[string]uint64{}
	prefixes := []string{
		"tier1Binary", "tier1Compare", "tier1Conversion", "tier1Constructor", "tier1Quiet",
		"tier1MinMax", "tier1ToInteger", "tier1Width",
	}
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.CONST {
			continue
		}
		for _, spec := range gen.Specs {
			values, ok := spec.(*ast.ValueSpec)
			if !ok || len(values.Names) != len(values.Values) {
				t.Fatalf("generated Tier 1 compare/conversion long harness has an implicit or malformed const declaration")
			}
			for i, name := range values.Names {
				selected := false
				for _, prefix := range prefixes {
					if strings.HasPrefix(name.Name, prefix) {
						selected = true
						break
					}
				}
				if !selected {
					continue
				}
				value, err := evalGeneratedUintConstant(values.Values[i], constants)
				if err != nil {
					t.Fatalf("evaluate generated Tier 1 compare/conversion constant %s: %v", name.Name, err)
				}
				constants[name.Name] = value
			}
		}
	}
	require := func(name string) uint64 {
		t.Helper()
		value, ok := constants[name]
		if !ok {
			t.Fatalf("generated Tier 1 compare/conversion long harness is missing constant %s", name)
		}
		return value
	}
	byWidth := func(prefix string) map[string]uint64 {
		return map[string]uint64{
			"decimal32":  require(prefix + "32"),
			"decimal64":  require(prefix + "64"),
			"decimal128": require(prefix + "128"),
		}
	}
	return tier1CompareConversionLongInventory{
		BoundaryValues: map[string]uint64{
			"decimal32":  require("tier1CompareConversionBoundary32Count"),
			"decimal64":  require("tier1CompareConversionBoundary64Count"),
			"decimal128": require("tier1CompareConversionBoundary128Count"),
		},
		SemanticValues: map[string]uint64{
			"decimal32":  require("tier1ConversionSemantic32Count"),
			"decimal64":  require("tier1ConversionSemantic64Count"),
			"decimal128": require("tier1ConversionSemantic128Count"),
		},
		QuietPredicates:        require("tier1QuietPredicateCount"),
		MinMaxOperations:       require("tier1MinMaxOperationCount"),
		ComparisonStructured:   byWidth("tier1CompareStructured"),
		ComparisonRandom:       byWidth("tier1CompareRandomComparisons"),
		ComparisonRandomHashes: byWidth("tier1CompareRandomStreamHash"),
		ComparisonTotal:        byWidth("tier1CompareTotal"),
		ToIntegerOperations:    require("tier1ToIntegerOperationCount"),
		ToIntegerTotal:         byWidth("tier1ToIntegerTotal"),
		ToIntegerRandomHashes:  byWidth("tier1ToIntegerRandomStreamHash"),
		WidthOperations:        byWidth("tier1WidthOperationsFrom"),
		WidthTotal:             byWidth("tier1WidthTotal"),
		WidthRandomHashes:      byWidth("tier1WidthRandomStreamHash"),
		BinaryRandomHashes:     byWidth("tier1BinaryRandomStreamHash"),
		ConstructorOperations:  require("tier1ConstructorOperationCount"),
		ConstructorTotal:       require("tier1ConstructorTotal"),
		ConstructorConvenience: require("tier1ConstructorConvenienceChecks"),
		ConstructorRandomHash:  require("tier1ConstructorRandomStreamHash"),
		ConversionStructured:   require("tier1ConversionStructured"),
		ConversionRandom:       require("tier1ConversionRandom"),
		ConversionTotal:        require("tier1ConversionTotal"),
	}
}

func loadRustTier1LongConstants(t *testing.T, filename string) map[string]uint64 {
	t.Helper()
	path := filepath.Join("..", "..", "..", "bid754-rs", "ffi-verify", "tests", filename)
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read generated Rust Tier 1 long harness %s: %v", filename, err)
	}
	pattern := regexp.MustCompile(`(?m)^const ([A-Z][A-Z0-9_]*): (?:u64|usize) = ([0-9]+);$`)
	constants := map[string]uint64{}
	for _, match := range pattern.FindAllSubmatch(raw, -1) {
		value, err := strconv.ParseUint(string(match[2]), 10, 64)
		if err != nil {
			t.Fatalf("parse generated Rust Tier 1 constant %s: %v", match[1], err)
		}
		constants[string(match[1])] = value
	}
	if len(constants) == 0 {
		t.Fatalf("generated Rust Tier 1 long harness %s has no literal count constants", filename)
	}
	return constants
}

func loadRustTier1ArithmeticLongInventory(t *testing.T) tier1ArithmeticLongInventory {
	t.Helper()
	constants := loadRustTier1LongConstants(t, "tier1_arithmetic_long_generated.rs")
	require := func(name string) uint64 {
		t.Helper()
		value, ok := constants[name]
		if !ok {
			available := make([]string, 0, len(constants))
			for key := range constants {
				available = append(available, key)
			}
			sort.Strings(available)
			t.Fatalf("generated Rust Tier 1 arithmetic long harness is missing constant %s; available=%v", name, available)
		}
		return value
	}
	byWidth := func(prefix, suffix string) map[string]uint64 {
		return map[string]uint64{
			"decimal32":  require(prefix + "32" + suffix),
			"decimal64":  require(prefix + "64" + suffix),
			"decimal128": require(prefix + "128" + suffix),
		}
	}
	return tier1ArithmeticLongInventory{
		BoundaryValues:              byWidth("BOUNDARY", "_COUNT"),
		SemanticRounded:             byWidth("SEMANTIC_ROUNDED", "_COUNT"),
		SemanticScale:               byWidth("SEMANTIC_SCALE", "_COUNT"),
		SemanticRemainder:           byWidth("SEMANTIC_REMAINDER", "_COUNT"),
		SemanticFma:                 byWidth("SEMANTIC_FMA", "_COUNT"),
		SemanticSqrt:                byWidth("SEMANTIC_SQRT", "_COUNT"),
		StructuredCases:             byWidth("STRUCTURED", "_COUNT"),
		RandomOperations:            require("RANDOM_OPS"),
		RandomCasesPerOp:            byWidth("RANDOM_CASES", ""),
		ScaleFiniteTransitionLimits: byWidth("SCALE_FINITE_TRANSITION_LIMIT", ""),
		ScaleRandomStrata:           require("SCALE_RANDOM_STRATA"),
		ScaleModeCross:              require("SCALE_MODE_CROSS"),
		ScaleModeCrossGroups:        byWidth("SCALE_MODE_CROSS_GROUPS", ""),
		ScaleTupleHashes:            byWidth("SCALE_TUPLE_HASH", ""),
		PairStreamHashes:            byWidth("PAIR_STREAM_HASH", ""),
		FmaTripleStreamHashes:       byWidth("FMA_TRIPLE_STREAM_HASH", ""),
		RandomStreamHashes:          byWidth("RANDOM_STREAM_HASH", ""),
		TotalComparisons:            byWidth("TOTAL", "_COUNT"),
	}
}

func loadRustTier1CompareConversionLongInventory(t *testing.T) tier1CompareConversionLongInventory {
	t.Helper()
	constants := loadRustTier1LongConstants(t, "tier1_compare_conversion_long_generated.rs")
	require := func(name string) uint64 {
		t.Helper()
		value, ok := constants[name]
		if !ok {
			t.Fatalf("generated Rust Tier 1 compare/conversion long harness is missing constant %s", name)
		}
		return value
	}
	byWidth := func(prefix, suffix string) map[string]uint64 {
		return map[string]uint64{
			"decimal32":  require(prefix + "32" + suffix),
			"decimal64":  require(prefix + "64" + suffix),
			"decimal128": require(prefix + "128" + suffix),
		}
	}
	return tier1CompareConversionLongInventory{
		BoundaryValues:         byWidth("BOUNDARY", "_COUNT"),
		SemanticValues:         byWidth("SEMANTIC", "_COUNT"),
		QuietPredicates:        require("QUIET_PREDICATE_COUNT"),
		MinMaxOperations:       require("MINMAX_OPERATION_COUNT"),
		ComparisonStructured:   byWidth("COMPARE_STRUCTURED", ""),
		ComparisonRandom:       byWidth("COMPARE_RANDOM", ""),
		ComparisonRandomHashes: byWidth("COMPARE_RANDOM_STREAM_HASH", ""),
		ComparisonTotal:        byWidth("COMPARE_TOTAL", ""),
		ToIntegerOperations:    require("TO_INT_OPERATION_COUNT"),
		ToIntegerTotal:         byWidth("TO_INT_TOTAL", ""),
		ToIntegerRandomHashes:  byWidth("TO_INT_RANDOM_STREAM_HASH", ""),
		WidthOperations:        byWidth("WIDTH_OPERATION_COUNT", ""),
		WidthTotal:             byWidth("WIDTH_TOTAL", ""),
		WidthRandomHashes:      byWidth("WIDTH_RANDOM_STREAM_HASH", ""),
		BinaryRandomHashes:     byWidth("BINARY_RANDOM_STREAM_HASH", ""),
		ConstructorOperations:  require("CONSTRUCTOR_OPERATION_COUNT"),
		ConstructorTotal:       require("CONSTRUCTOR_TOTAL"),
		ConstructorConvenience: require("CONSTRUCTOR_CONVENIENCE"),
		ConstructorRandomHash:  require("CONSTRUCTOR_RANDOM_STREAM_HASH"),
		ConversionStructured:   require("CONVERSION_STRUCTURED"),
		ConversionRandom:       require("CONVERSION_RANDOM"),
		ConversionTotal:        require("CONVERSION_TOTAL"),
	}
}

func loadBidCodecD64128HarnessInventory(t *testing.T) bidCodecD64128HarnessInventory {
	t.Helper()
	path := filepath.Join("..", "..", "..", "bid754-codec-go", "decimal64_128_long_test.go")
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		t.Fatalf("parse generated Decimal64/128 long harness: %v", err)
	}
	constants := map[string]uint64{}
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.CONST {
			continue
		}
		for _, spec := range gen.Specs {
			values, ok := spec.(*ast.ValueSpec)
			if !ok || len(values.Names) != len(values.Values) {
				t.Fatalf("generated Decimal64/128 long harness has an implicit or malformed const declaration")
			}
			for i, name := range values.Names {
				if !strings.HasPrefix(name.Name, "decimal64") && !strings.HasPrefix(name.Name, "decimal128") {
					continue
				}
				value, err := evalGeneratedUintConstant(values.Values[i], constants)
				if err != nil {
					t.Fatalf("evaluate generated constant %s: %v", name.Name, err)
				}
				constants[name.Name] = value
			}
		}
	}
	require := func(name string) uint64 {
		t.Helper()
		value, ok := constants[name]
		if !ok {
			t.Fatalf("generated Decimal64/128 long harness is missing constant %s", name)
		}
		return value
	}
	classes := func(prefix string) map[string]uint64 {
		return map[string]uint64{
			"negative":     require(prefix + "Negative"),
			"normal":       require(prefix + "Normal"),
			"zero":         require(prefix + "Zero"),
			"infinity":     require(prefix + "Infinity"),
			"qnan":         require(prefix + "QNaN"),
			"snan":         require(prefix + "SNaN"),
			"canonical":    require(prefix + "Canonical"),
			"noncanonical": require(prefix + "Noncanonical"),
		}
	}
	structuredClasses := func(prefix string) map[string]uint64 {
		return map[string]uint64{
			"negative":     require(prefix + "Negative"),
			"normal":       require(prefix + "Normal"),
			"zero":         require(prefix + "Zero"),
			"infinity":     require(prefix + "Infinity"),
			"qnan":         require(prefix + "QNaN"),
			"snan":         require(prefix + "SNaN"),
			"canonical":    require(prefix + "Canonical"),
			"noncanonical": require(prefix + "Noncanonical"),
		}
	}
	return bidCodecD64128HarnessInventory{
		DifferentialCasesPerWidth: require("decimal64And128DifferentialCasesPerWidth"),
		DifferentialClasses: map[string]map[string]uint64{
			"decimal64":  classes("decimal64Differential"),
			"decimal128": classes("decimal128Differential"),
		},
		BoundaryRawCases: map[string]uint64{
			"decimal64":  require("decimal64BoundaryRawCases"),
			"decimal128": require("decimal128BoundaryRawCases"),
		},
		StructuredCases: map[string]uint64{
			"decimal64":  require("decimal64StructuredTotal"),
			"decimal128": require("decimal128StructuredTotal"),
		},
		StructuredClasses: map[string]map[string]uint64{
			"decimal64":  structuredClasses("decimal64Structured"),
			"decimal128": structuredClasses("decimal128Structured"),
		},
	}
}

func loadBidCodecD32HarnessInventory(t *testing.T) bidCodecD32HarnessInventory {
	t.Helper()
	path := filepath.Join("..", "..", "..", "bid754-codec-go", "exhaustive32_long_test.go")
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		t.Fatalf("parse generated Decimal32 exhaustive harness: %v", err)
	}
	constants := map[string]uint64{}
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.CONST {
			continue
		}
		for _, spec := range gen.Specs {
			values, ok := spec.(*ast.ValueSpec)
			if !ok || len(values.Names) != len(values.Values) {
				t.Fatalf("generated Decimal32 exhaustive harness has an implicit or malformed const declaration")
			}
			for i, name := range values.Names {
				if !strings.HasPrefix(name.Name, "exhaustive32") {
					continue
				}
				value, err := evalGeneratedUintConstant(values.Values[i], constants)
				if err != nil {
					t.Fatalf("evaluate generated constant %s: %v", name.Name, err)
				}
				constants[name.Name] = value
			}
		}
	}
	require := func(name string) uint64 {
		t.Helper()
		value, ok := constants[name]
		if !ok {
			t.Fatalf("generated Decimal32 exhaustive harness is missing constant %s", name)
		}
		return value
	}
	return bidCodecD32HarnessInventory{
		RawPatterns:     require("exhaustive32Total"),
		StructuredCases: require("exhaustive32StructuredTotal"),
		RawClasses: map[string]uint64{
			"normal":       require("exhaustive32RawNormal"),
			"zero":         require("exhaustive32RawZero"),
			"infinity":     require("exhaustive32RawInfinity"),
			"qnan":         require("exhaustive32RawQNaN"),
			"snan":         require("exhaustive32RawSNaN"),
			"canonical":    require("exhaustive32RawCanonical"),
			"noncanonical": require("exhaustive32RawNoncanonical"),
		},
		StructuredClasses: map[string]uint64{
			"normal":   require("exhaustive32StructuredNormal"),
			"zero":     require("exhaustive32StructuredZero"),
			"infinity": require("exhaustive32StructuredInfinity"),
			"qnan":     require("exhaustive32StructuredQNaN"),
			"snan":     require("exhaustive32StructuredSNaN"),
		},
	}
}

func TestVerificationAnchorsMatchGeneratedArtifacts(t *testing.T) {
	anchors := loadVerificationAnchors(t)

	spec, err := LoadGenerated(filepath.Join("..", "..", "generated", "testspec", "spec_index.json"))
	if err != nil {
		t.Fatalf("load generated spec: %v", err)
	}

	readTotal := len(spec.ReadCases)
	statusControl := 0
	nativeCompareSkips := 0
	for _, tc := range spec.ReadCases {
		if tc.Kind == "status_control" {
			statusControl++
		}
		if tc.NativeCompareSkipReason != "" {
			nativeCompareSkips++
		}
	}
	if readTotal != anchors.ReadtestCasesTotal {
		t.Errorf("generated readtest case total = %d, anchor = %d", readTotal, anchors.ReadtestCasesTotal)
	}
	if statusControl != anchors.ReadtestStatusControlCases {
		t.Errorf("generated readtest status_control case count = %d, anchor = %d", statusControl, anchors.ReadtestStatusControlCases)
	}
	if nativeCompareSkips != anchors.ReadtestNativeCompareSkipCases {
		t.Errorf("generated readtest native-compare skip case count = %d, anchor = %d", nativeCompareSkips, anchors.ReadtestNativeCompareSkipCases)
	}
	// The goport gate executes every generated readtest row, status_control
	// included (the section 5.7.4 status-control operations are ported into
	// bidgo and driven with an explicit status word and rounding mode), so the
	// executed count is the full corpus with nothing subtracted.
	if readTotal != anchors.GoportReadtestExecutedCases {
		t.Errorf("goport executed readtest case count = %d, anchor = %d", readTotal, anchors.GoportReadtestExecutedCases)
	}
	if len(spec.FFICases) != anchors.FFIBitcompareCasesTotal {
		t.Errorf("generated FFI bit-compare case total = %d, anchor = %d", len(spec.FFICases), anchors.FFIBitcompareCasesTotal)
	}
	tier1ArithmeticOutputs, err := GenerateTier1ArithmeticLongOutputs()
	if err != nil {
		t.Fatalf("generate Tier 1 arithmetic long outputs for consumer anchor: %v", err)
	}
	assertGeneratedOutputSet(t, "Tier 1 arithmetic long", tier1ArithmeticOutputs,
		tier1ArithmeticLongGeneratedPath,
		tier1ArithmeticRustLongGeneratedPath,
	)
	if got := uint64(len(tier1ArithmeticOutputs)); got != anchors.Tier1ArithmeticLongConsumers {
		t.Errorf("Tier 1 arithmetic long generated consumers = %d, anchor = %d", got, anchors.Tier1ArithmeticLongConsumers)
	}
	tier1Arithmetic := loadTier1ArithmeticLongInventory(t)
	for _, check := range []struct {
		label string
		got   map[string]uint64
		want  map[string]uint64
	}{
		{"Tier 1 arithmetic boundary values", tier1Arithmetic.BoundaryValues, anchors.Tier1ArithmeticBoundaryValues},
		{"Tier 1 arithmetic semantic rounded pairs", tier1Arithmetic.SemanticRounded, anchors.Tier1ArithmeticSemanticRounded},
		{"Tier 1 arithmetic semantic scale cases", tier1Arithmetic.SemanticScale, anchors.Tier1ArithmeticSemanticScale},
		{"Tier 1 arithmetic semantic remainder pairs", tier1Arithmetic.SemanticRemainder, anchors.Tier1ArithmeticSemanticRemainder},
		{"Tier 1 arithmetic semantic fma triples", tier1Arithmetic.SemanticFma, anchors.Tier1ArithmeticSemanticFma},
		{"Tier 1 arithmetic semantic sqrt cases", tier1Arithmetic.SemanticSqrt, anchors.Tier1ArithmeticSemanticSqrt},
		{"Tier 1 arithmetic structured comparisons", tier1Arithmetic.StructuredCases, anchors.Tier1ArithmeticStructuredCases},
		{"Tier 1 arithmetic random cases per operation", tier1Arithmetic.RandomCasesPerOp, anchors.Tier1ArithmeticRandomCasesPerOp},
		{"Tier 1 arithmetic ScaleB finite-transition limits", tier1Arithmetic.ScaleFiniteTransitionLimits, anchors.Tier1ArithmeticScaleFiniteTransitionLimits},
		{"Tier 1 arithmetic ScaleB mode-cross groups", tier1Arithmetic.ScaleModeCrossGroups, anchors.Tier1ArithmeticScaleModeCrossGroups},
		{"Tier 1 arithmetic ScaleB tuple hashes", tier1Arithmetic.ScaleTupleHashes, anchors.Tier1ArithmeticScaleTupleHashes},
		{"Tier 1 arithmetic pair stream hashes", tier1Arithmetic.PairStreamHashes, anchors.Tier1ArithmeticPairStreamHashes},
		{"Tier 1 arithmetic fma triple stream hashes", tier1Arithmetic.FmaTripleStreamHashes, anchors.Tier1ArithmeticFmaTripleStreamHashes},
		{"Tier 1 arithmetic random stream hashes", tier1Arithmetic.RandomStreamHashes, anchors.Tier1ArithmeticRandomStreamHashes},
		{"Tier 1 arithmetic total comparisons", tier1Arithmetic.TotalComparisons, anchors.Tier1ArithmeticTotalComparisons},
	} {
		if !reflect.DeepEqual(check.got, check.want) {
			t.Errorf("generated %s = %#v, anchor = %#v", check.label, check.got, check.want)
		}
	}
	if tier1Arithmetic.ScaleRandomStrata != anchors.Tier1ArithmeticScaleRandomStrata {
		t.Errorf("generated Tier 1 arithmetic ScaleB random strata = %d, anchor = %d", tier1Arithmetic.ScaleRandomStrata, anchors.Tier1ArithmeticScaleRandomStrata)
	}
	if tier1Arithmetic.ScaleModeCross != anchors.Tier1ArithmeticScaleModeCross {
		t.Errorf("generated Tier 1 arithmetic ScaleB meaningful mode cross = %d, anchor = %d", tier1Arithmetic.ScaleModeCross, anchors.Tier1ArithmeticScaleModeCross)
	}
	if tier1Arithmetic.RandomOperations != anchors.Tier1ArithmeticRandomOperations {
		t.Errorf("generated Tier 1 arithmetic random operations = %d, anchor = %d", tier1Arithmetic.RandomOperations, anchors.Tier1ArithmeticRandomOperations)
	}
	rustTier1Arithmetic := loadRustTier1ArithmeticLongInventory(t)
	for _, check := range []struct {
		label string
		got   map[string]uint64
		want  map[string]uint64
	}{
		{"Rust Tier 1 arithmetic boundary values", rustTier1Arithmetic.BoundaryValues, anchors.Tier1ArithmeticBoundaryValues},
		{"Rust Tier 1 arithmetic semantic rounded pairs", rustTier1Arithmetic.SemanticRounded, anchors.Tier1ArithmeticSemanticRounded},
		{"Rust Tier 1 arithmetic semantic scale cases", rustTier1Arithmetic.SemanticScale, anchors.Tier1ArithmeticSemanticScale},
		{"Rust Tier 1 arithmetic semantic remainder pairs", rustTier1Arithmetic.SemanticRemainder, anchors.Tier1ArithmeticSemanticRemainder},
		{"Rust Tier 1 arithmetic semantic fma triples", rustTier1Arithmetic.SemanticFma, anchors.Tier1ArithmeticSemanticFma},
		{"Rust Tier 1 arithmetic semantic sqrt cases", rustTier1Arithmetic.SemanticSqrt, anchors.Tier1ArithmeticSemanticSqrt},
		{"Rust Tier 1 arithmetic structured comparisons", rustTier1Arithmetic.StructuredCases, anchors.Tier1ArithmeticStructuredCases},
		{"Rust Tier 1 arithmetic random cases per operation", rustTier1Arithmetic.RandomCasesPerOp, anchors.Tier1ArithmeticRandomCasesPerOp},
		{"Rust Tier 1 arithmetic ScaleB finite-transition limits", rustTier1Arithmetic.ScaleFiniteTransitionLimits, anchors.Tier1ArithmeticScaleFiniteTransitionLimits},
		{"Rust Tier 1 arithmetic ScaleB mode-cross groups", rustTier1Arithmetic.ScaleModeCrossGroups, anchors.Tier1ArithmeticScaleModeCrossGroups},
		{"Rust Tier 1 arithmetic ScaleB tuple hashes", rustTier1Arithmetic.ScaleTupleHashes, anchors.Tier1ArithmeticScaleTupleHashes},
		{"Rust Tier 1 arithmetic pair stream hashes", rustTier1Arithmetic.PairStreamHashes, anchors.Tier1ArithmeticPairStreamHashes},
		{"Rust Tier 1 arithmetic fma triple stream hashes", rustTier1Arithmetic.FmaTripleStreamHashes, anchors.Tier1ArithmeticFmaTripleStreamHashes},
		{"Rust Tier 1 arithmetic random stream hashes", rustTier1Arithmetic.RandomStreamHashes, anchors.Tier1ArithmeticRandomStreamHashes},
		{"Rust Tier 1 arithmetic total comparisons", rustTier1Arithmetic.TotalComparisons, anchors.Tier1ArithmeticTotalComparisons},
	} {
		if !reflect.DeepEqual(check.got, check.want) {
			t.Errorf("generated %s = %#v, anchor = %#v", check.label, check.got, check.want)
		}
	}
	if rustTier1Arithmetic.ScaleRandomStrata != anchors.Tier1ArithmeticScaleRandomStrata {
		t.Errorf("generated Rust Tier 1 arithmetic ScaleB random strata = %d, anchor = %d", rustTier1Arithmetic.ScaleRandomStrata, anchors.Tier1ArithmeticScaleRandomStrata)
	}
	if rustTier1Arithmetic.ScaleModeCross != anchors.Tier1ArithmeticScaleModeCross {
		t.Errorf("generated Rust Tier 1 arithmetic ScaleB meaningful mode cross = %d, anchor = %d", rustTier1Arithmetic.ScaleModeCross, anchors.Tier1ArithmeticScaleModeCross)
	}
	if rustTier1Arithmetic.RandomOperations != anchors.Tier1ArithmeticRandomOperations {
		t.Errorf("generated Rust Tier 1 arithmetic random operations = %d, anchor = %d", rustTier1Arithmetic.RandomOperations, anchors.Tier1ArithmeticRandomOperations)
	}
	// decNumber differential gate: the checked-in shared constants, the
	// generated inventory JSON, and the external anchors must agree in every
	// direction; the regenerated output set pins the closed artifact world.
	manifest, err := LoadManifest(filepath.Join("..", "..", "testgen_manifest.json"))
	if err != nil {
		t.Fatalf("load manifest for decNumber differential consumer anchor: %v", err)
	}
	decnumberDiffOutputs, err := GenerateDecnumberDifferentialOutputs(manifest)
	if err != nil {
		t.Fatalf("generate decNumber differential outputs for consumer anchor: %v", err)
	}
	assertGeneratedOutputSet(t, "decNumber differential", decnumberDiffOutputs,
		decnumberDiffSharedGeneratedPath,
		decnumberDiffNativeShimGeneratedPath,
		decnumberDiffRunnerGeneratedPath,
		decnumberDiffStubGeneratedPath,
		manifest.DecnumberDifferential.InventoryOutput,
	)
	decnumberDiffRunnerConsumers := uint64(0)
	for path := range decnumberDiffOutputs {
		if strings.HasSuffix(path, "_native_test.go") {
			decnumberDiffRunnerConsumers++
		}
	}
	if decnumberDiffRunnerConsumers != anchors.DecnumberDiffConsumers {
		t.Errorf("decNumber differential generated runner consumers = %d, anchor = %d", decnumberDiffRunnerConsumers, anchors.DecnumberDiffConsumers)
	}
	decnumberDiff := loadDecnumberDifferentialArtifactInventory(t)
	for _, check := range []struct {
		label string
		got   map[string]uint64
		want  map[string]uint64
	}{
		{"decNumber differential boundary values", decnumberDiff.BoundaryValues, anchors.DecnumberDiffBoundaryValues},
		{"decNumber differential probe values", decnumberDiff.ProbeValues, anchors.DecnumberDiffProbeValues},
		{"decNumber differential exact-product values", decnumberDiff.ExactProductValues, anchors.DecnumberDiffExactProductValues},
		{"decNumber differential exact-product addends", decnumberDiff.ExactProductAddends, anchors.DecnumberDiffExactProductAddends},
		{"decNumber differential structured comparisons", decnumberDiff.StructuredComparisons, anchors.DecnumberDiffStructuredComparisons},
		{"decNumber differential structured fma exclusions", decnumberDiff.StructuredFmaExcluded, anchors.DecnumberDiffStructuredFmaExcluded},
		{"decNumber differential structured known divergences", decnumberDiff.StructuredKnownDivergences, anchors.DecnumberDiffStructuredKnownDivergences},
		{"decNumber differential structured stream hashes", decnumberDiff.StructuredStreamHashes, anchors.DecnumberDiffStructuredStreamHashes},
		{"decNumber differential random comparisons", decnumberDiff.RandomComparisons, anchors.DecnumberDiffRandomComparisons},
		{"decNumber differential random fma exclusions", decnumberDiff.RandomFmaExcluded, anchors.DecnumberDiffRandomFmaExcluded},
		{"decNumber differential random known divergences", decnumberDiff.RandomKnownDivergences, anchors.DecnumberDiffRandomKnownDivergences},
		{"decNumber differential random stream hashes", decnumberDiff.RandomStreamHashes, anchors.DecnumberDiffRandomStreamHashes},
		{"decNumber differential total comparisons", decnumberDiff.TotalComparisons, anchors.DecnumberDiffTotalComparisons},
	} {
		if !reflect.DeepEqual(check.got, check.want) {
			t.Errorf("generated %s = %#v, anchor = %#v", check.label, check.got, check.want)
		}
	}
	if decnumberDiff.RandomPairsPerOp != anchors.DecnumberDiffRandomPairsPerOp {
		t.Errorf("generated decNumber differential random pairs per operation = %d, anchor = %d",
			decnumberDiff.RandomPairsPerOp, anchors.DecnumberDiffRandomPairsPerOp)
	}
	decnumberDiffJSON := loadDecnumberDifferentialInventoryJSON(t)
	for _, widthEntry := range decnumberDiffJSON.Widths {
		label := widthEntry.Width
		for _, check := range []struct {
			name string
			got  uint64
			want uint64
		}{
			{"boundary_included_values", uint64(widthEntry.BoundaryIncluded), anchors.DecnumberDiffBoundaryValues[label]},
			{"probe_values", uint64(widthEntry.ProbeValues), anchors.DecnumberDiffProbeValues[label]},
			{"exact_product_values", uint64(widthEntry.ExactProductValues), anchors.DecnumberDiffExactProductValues[label]},
			{"structured_comparisons", widthEntry.StructuredComparisons, anchors.DecnumberDiffStructuredComparisons[label]},
			{"structured_known_divergences", widthEntry.StructuredKnownDivergences, anchors.DecnumberDiffStructuredKnownDivergences[label]},
			{"structured_stream_hash", widthEntry.StructuredStreamHash, anchors.DecnumberDiffStructuredStreamHashes[label]},
			{"random_comparisons", widthEntry.RandomComparisons, anchors.DecnumberDiffRandomComparisons[label]},
			{"random_known_divergences", widthEntry.RandomKnownDivergences, anchors.DecnumberDiffRandomKnownDivergences[label]},
			{"random_stream_hash", widthEntry.RandomStreamHash, anchors.DecnumberDiffRandomStreamHashes[label]},
			{"total_comparisons", widthEntry.TotalComparisons, anchors.DecnumberDiffTotalComparisons[label]},
		} {
			if check.got != check.want {
				t.Errorf("decNumber differential inventory %s %s = %d, anchor = %d", label, check.name, check.got, check.want)
			}
		}
	}

	// Routing sentinels: the hand-pinned rows in verification_sentinels.json
	// must be byte-equal, in order, with both generated runner literals. The
	// generator cannot touch the pin file, so a selection change (or a hand
	// edit to either runner) fails here until a human re-audits and re-pins.
	sentinels := loadVerificationSentinels(t)
	if len(sentinels.MixedFMAFusednessRows) == 0 {
		t.Errorf("verification_sentinels.json pins no mixed FMA fusedness rows")
	}
	goFusednessRows := loadGeneratedGoStringSliceLiteral(t,
		filepath.Join("..", "..", "..", "bid754-go", "generated_ffi_bitcompare_native_test.go"),
		"mixedFMAFusednessSentinelRows")
	if !reflect.DeepEqual(goFusednessRows, sentinels.MixedFMAFusednessRows) {
		t.Errorf("generated Go mixed FMA fusedness rows diverge from verification_sentinels.json: generated %d rows, pinned %d rows%s",
			len(goFusednessRows), len(sentinels.MixedFMAFusednessRows),
			firstSentinelRowDivergence(goFusednessRows, sentinels.MixedFMAFusednessRows))
	}
	rustFusednessRows := loadRustStringArrayLiteral(t,
		filepath.Join("..", "..", "..", "bid754-rs", "tests", "public_parity_generated.rs"),
		"MIXED_FMA_FUSEDNESS_SENTINEL_ROWS")
	if !reflect.DeepEqual(rustFusednessRows, sentinels.MixedFMAFusednessRows) {
		t.Errorf("generated Rust mixed FMA fusedness rows diverge from verification_sentinels.json: generated %d rows, pinned %d rows%s",
			len(rustFusednessRows), len(sentinels.MixedFMAFusednessRows),
			firstSentinelRowDivergence(rustFusednessRows, sentinels.MixedFMAFusednessRows))
	}
	// Mixed-format FFI operand-swap routing sentinels are a Go-only native
	// differential domain (no Rust FFI leg exists), so — like the decNumber
	// differential sentinel — only the generated Go literal is bound. The
	// generated row count constant is bound to the literal length here so the
	// count cannot drift outside the native runtime replay.
	if len(sentinels.MixedFormatFFIRoutingRows) == 0 {
		t.Errorf("verification_sentinels.json pins no mixed-format FFI routing sentinel rows")
	}
	goMixedFFIRoutingRows := loadGeneratedGoStringSliceLiteral(t,
		filepath.Join("..", "..", "..", "bid754-go", "generated_ffi_bitcompare_native_test.go"),
		"mixedFormatFFIRoutingSentinelRows")
	if !reflect.DeepEqual(goMixedFFIRoutingRows, sentinels.MixedFormatFFIRoutingRows) {
		t.Errorf("generated Go mixed-format FFI routing sentinel rows diverge from verification_sentinels.json: generated %d rows, pinned %d rows%s",
			len(goMixedFFIRoutingRows), len(sentinels.MixedFormatFFIRoutingRows),
			firstSentinelRowDivergence(goMixedFFIRoutingRows, sentinels.MixedFormatFFIRoutingRows))
	}
	mixedFFIRoutingRowCount := loadGeneratedGoUintConstant(t,
		filepath.Join("..", "..", "..", "bid754-go", "generated_ffi_bitcompare_native_test.go"),
		"mixedFormatFFIRoutingSentinelRowCount")
	if got := uint64(len(goMixedFFIRoutingRows)); got != mixedFFIRoutingRowCount {
		t.Errorf("generated mixed-format FFI routing sentinel row literal count %d diverges from the generated constant %d", got, mixedFFIRoutingRowCount)
	}
	if len(sentinels.Tier1ArithmeticRoutingRows) == 0 {
		t.Errorf("verification_sentinels.json pins no Tier 1 arithmetic routing sentinel rows")
	}
	goSentinelRows := loadGeneratedGoStringSliceLiteral(t,
		filepath.Join("..", "..", "..", "bid754-go", "generated_ffi_bitcompare_tier1_arithmetic_long_test.go"),
		"tier1ArithmeticRoutingSentinelRows")
	if !reflect.DeepEqual(goSentinelRows, sentinels.Tier1ArithmeticRoutingRows) {
		t.Errorf("generated Go Tier 1 arithmetic routing sentinel rows diverge from verification_sentinels.json: generated %d rows, pinned %d rows%s",
			len(goSentinelRows), len(sentinels.Tier1ArithmeticRoutingRows),
			firstSentinelRowDivergence(goSentinelRows, sentinels.Tier1ArithmeticRoutingRows))
	}
	rustSentinelRows := loadRustRoutingSentinelRows(t, "tier1_arithmetic_long_generated.rs")
	if !reflect.DeepEqual(rustSentinelRows, sentinels.Tier1ArithmeticRoutingRows) {
		t.Errorf("generated Rust Tier 1 arithmetic routing sentinel rows diverge from verification_sentinels.json: generated %d rows, pinned %d rows%s",
			len(rustSentinelRows), len(sentinels.Tier1ArithmeticRoutingRows),
			firstSentinelRowDivergence(rustSentinelRows, sentinels.Tier1ArithmeticRoutingRows))
	}
	if len(sentinels.Tier1CompareConversionRoutingRows) == 0 {
		t.Errorf("verification_sentinels.json pins no Tier 1 compare/conversion routing sentinel rows")
	}
	goCCSentinelRows := loadGeneratedGoStringSliceLiteral(t,
		filepath.Join("..", "..", "..", "bid754-go", "generated_ffi_bitcompare_tier1_compare_conversion_long_test.go"),
		"tier1CompareConversionRoutingSentinelRows")
	if !reflect.DeepEqual(goCCSentinelRows, sentinels.Tier1CompareConversionRoutingRows) {
		t.Errorf("generated Go Tier 1 compare/conversion routing sentinel rows diverge from verification_sentinels.json: generated %d rows, pinned %d rows%s",
			len(goCCSentinelRows), len(sentinels.Tier1CompareConversionRoutingRows),
			firstSentinelRowDivergence(goCCSentinelRows, sentinels.Tier1CompareConversionRoutingRows))
	}
	rustCCSentinelRows := loadRustRoutingSentinelRows(t, "tier1_compare_conversion_long_generated.rs")
	if !reflect.DeepEqual(rustCCSentinelRows, sentinels.Tier1CompareConversionRoutingRows) {
		t.Errorf("generated Rust Tier 1 compare/conversion routing sentinel rows diverge from verification_sentinels.json: generated %d rows, pinned %d rows%s",
			len(rustCCSentinelRows), len(sentinels.Tier1CompareConversionRoutingRows),
			firstSentinelRowDivergence(rustCCSentinelRows, sentinels.Tier1CompareConversionRoutingRows))
	}
	if len(sentinels.DecnumberDifferentialRows) == 0 {
		t.Errorf("verification_sentinels.json pins no decNumber differential sentinel rows")
	}
	decnumberDiffSentinelRowsLiteral := loadGeneratedGoStringSliceLiteral(t,
		filepath.Join("..", "..", "..", "bid754-go", "generated_decnumber_differential_native_test.go"),
		"decnumberDiffSentinelRows")
	if !reflect.DeepEqual(decnumberDiffSentinelRowsLiteral, sentinels.DecnumberDifferentialRows) {
		t.Errorf("generated decNumber differential sentinel rows diverge from verification_sentinels.json: generated %d rows, pinned %d rows%s",
			len(decnumberDiffSentinelRowsLiteral), len(sentinels.DecnumberDifferentialRows),
			firstSentinelRowDivergence(decnumberDiffSentinelRowsLiteral, sentinels.DecnumberDifferentialRows))
	}
	if got := uint64(len(decnumberDiffSentinelRowsLiteral)); got != decnumberDiff.SentinelRowCount {
		t.Errorf("generated decNumber differential sentinel row literal count %d diverges from the shared constant %d", got, decnumberDiff.SentinelRowCount)
	}

	// Decimal32 unary exhaustive gate: the generated lane/count constants,
	// the generated lane-name inventory, and the hand-pinned digest-map key
	// set must agree with the external anchors; the regenerated output set
	// pins the closed artifact world. The digest VALUES are runtime
	// execution evidence (they require running pinned Intel C over the full
	// space), so they are bound to the runner's log lines by cmd/verifylog,
	// not compared against any artifact here; this test pins their key set
	// and rejects zero (unpinned) digests.
	d32ExhaustiveOutputs, err := GenerateD32ExhaustiveOutputs()
	if err != nil {
		t.Fatalf("generate d32 exhaustive outputs for consumer anchor: %v", err)
	}
	assertGeneratedOutputSet(t, "d32 exhaustive", d32ExhaustiveOutputs,
		d32ExhaustiveNativeShimGeneratedPath,
		d32ExhaustiveRunnerGeneratedPath,
		d32ExhaustiveStubGeneratedPath,
		d32ExhaustiveRustRunnerGeneratedPath,
	)
	// The consumers anchor counts the sweep-executing runner artifacts: the
	// Go native long runner plus the generated Rust leg (shim and stub are
	// support files, not sweep consumers).
	d32ExhaustiveRunnerConsumers := uint64(0)
	for path := range d32ExhaustiveOutputs {
		if strings.HasSuffix(path, "_long_test.go") || strings.HasSuffix(path, "_long_generated.rs") {
			d32ExhaustiveRunnerConsumers++
		}
	}
	if d32ExhaustiveRunnerConsumers != anchors.D32ExhaustiveConsumers {
		t.Errorf("d32 exhaustive generated runner consumers = %d, anchor = %d", d32ExhaustiveRunnerConsumers, anchors.D32ExhaustiveConsumers)
	}
	d32Exhaustive := loadD32ExhaustiveArtifactInventory(t)
	for _, check := range []struct {
		label string
		got   uint64
		want  uint64
	}{
		{"d32 exhaustive lanes", d32Exhaustive.Lanes, anchors.D32ExhaustiveLanes},
		{"d32 exhaustive operations", d32Exhaustive.Operations, anchors.D32ExhaustiveOperations},
		{"d32 exhaustive cases per lane", d32Exhaustive.CasesPerLane, anchors.D32ExhaustiveCasesPerLane},
		{"d32 exhaustive total comparisons", d32Exhaustive.TotalComparisons, anchors.D32ExhaustiveTotalComparisons},
	} {
		if check.got != check.want {
			t.Errorf("generated %s = %d, anchor = %d", check.label, check.got, check.want)
		}
	}
	if anchors.D32ExhaustiveLanes*anchors.D32ExhaustiveCasesPerLane != anchors.D32ExhaustiveTotalComparisons {
		t.Errorf("d32 exhaustive anchors are internally inconsistent: lanes %d x cases-per-lane %d != total %d",
			anchors.D32ExhaustiveLanes, anchors.D32ExhaustiveCasesPerLane, anchors.D32ExhaustiveTotalComparisons)
	}
	d32ExhaustiveLaneNameRows := loadGeneratedGoStringSliceLiteral(t,
		filepath.Join("..", "..", "..", "bid754-go", "generated_d32_exhaustive_long_test.go"),
		"d32ExhaustiveLaneNames")
	if got := uint64(len(d32ExhaustiveLaneNameRows)); got != anchors.D32ExhaustiveLanes {
		t.Errorf("generated d32 exhaustive lane-name inventory carries %d lanes, anchor = %d", got, anchors.D32ExhaustiveLanes)
	}
	for _, lane := range d32ExhaustiveLaneNameRows {
		digest, pinned := anchors.D32ExhaustiveDigestByLane[lane]
		if !pinned {
			t.Errorf("d32 exhaustive anchors pin no result digest for generated lane %q", lane)
			continue
		}
		if digest == 0 {
			t.Errorf("d32 exhaustive anchors pin a zero result digest for lane %q (an unpinned digest binds nothing)", lane)
		}
	}
	if len(anchors.D32ExhaustiveDigestByLane) != len(d32ExhaustiveLaneNameRows) {
		t.Errorf("d32 exhaustive anchors pin %d lane digests for %d generated lanes (stale or extra digest keys)",
			len(anchors.D32ExhaustiveDigestByLane), len(d32ExhaustiveLaneNameRows))
	}
	if len(sentinels.D32ExhaustiveRows) == 0 {
		t.Errorf("verification_sentinels.json pins no d32 exhaustive sentinel rows")
	}
	d32ExhaustiveSentinelRowsLiteral := loadGeneratedGoStringSliceLiteral(t,
		filepath.Join("..", "..", "..", "bid754-go", "generated_d32_exhaustive_long_test.go"),
		"d32ExhaustiveSentinelRows")
	if !reflect.DeepEqual(d32ExhaustiveSentinelRowsLiteral, sentinels.D32ExhaustiveRows) {
		t.Errorf("generated d32 exhaustive sentinel rows diverge from verification_sentinels.json: generated %d rows, pinned %d rows%s",
			len(d32ExhaustiveSentinelRowsLiteral), len(sentinels.D32ExhaustiveRows),
			firstSentinelRowDivergence(d32ExhaustiveSentinelRowsLiteral, sentinels.D32ExhaustiveRows))
	}
	if got := uint64(len(d32ExhaustiveSentinelRowsLiteral)); got != d32Exhaustive.SentinelRowCount {
		t.Errorf("generated d32 exhaustive sentinel row literal count %d diverges from the generated constant %d", got, d32Exhaustive.SentinelRowCount)
	}
	rustD32ExhaustiveSentinelRows := loadRustStringArrayLiteral(t,
		filepath.Join("..", "..", "..", "bid754-rs", "ffi-verify", "tests", "d32_exhaustive_long_generated.rs"),
		"D32_EXHAUSTIVE_SENTINEL_ROWS")
	if !reflect.DeepEqual(rustD32ExhaustiveSentinelRows, sentinels.D32ExhaustiveRows) {
		t.Errorf("generated Rust d32 exhaustive sentinel rows diverge from verification_sentinels.json: generated %d rows, pinned %d rows%s",
			len(rustD32ExhaustiveSentinelRows), len(sentinels.D32ExhaustiveRows),
			firstSentinelRowDivergence(rustD32ExhaustiveSentinelRows, sentinels.D32ExhaustiveRows))
	}
	rustD32ExhaustiveLaneNames := loadRustStringArrayLiteral(t,
		filepath.Join("..", "..", "..", "bid754-rs", "ffi-verify", "tests", "d32_exhaustive_long_generated.rs"),
		"D32_EXHAUSTIVE_LANE_NAMES")
	if !reflect.DeepEqual(rustD32ExhaustiveLaneNames, d32ExhaustiveLaneNameRows) {
		t.Errorf("generated Rust d32 exhaustive lane-name inventory diverges from the Go runner inventory: Rust %d lanes, Go %d lanes%s",
			len(rustD32ExhaustiveLaneNames), len(d32ExhaustiveLaneNameRows),
			firstSentinelRowDivergence(rustD32ExhaustiveLaneNames, d32ExhaustiveLaneNameRows))
	}

	tier1CompareConversionOutputs, err := GenerateTier1CompareConversionLongOutputs()
	if err != nil {
		t.Fatalf("generate Tier 1 compare/conversion long outputs for consumer anchor: %v", err)
	}
	assertGeneratedOutputSet(t, "Tier 1 compare/conversion long", tier1CompareConversionOutputs,
		tier1CompareConversionLongGeneratedPath,
		tier1CompareConversionRustLongGeneratedPath,
	)
	if got := uint64(len(tier1CompareConversionOutputs)); got != anchors.Tier1CompareConversionLongConsumers {
		t.Errorf("Tier 1 compare/conversion long generated consumers = %d, anchor = %d", got, anchors.Tier1CompareConversionLongConsumers)
	}
	tier1CompareConversion := loadTier1CompareConversionLongInventory(t)
	for _, check := range []struct {
		label string
		got   map[string]uint64
		want  map[string]uint64
	}{
		{"Tier 1 compare/conversion boundary values", tier1CompareConversion.BoundaryValues, anchors.Tier1CompareConversionBoundaryValues},
		{"Tier 1 compare/conversion semantic values", tier1CompareConversion.SemanticValues, anchors.Tier1CompareConversionSemanticValues},
		{"Tier 1 compare/minmax structured comparisons", tier1CompareConversion.ComparisonStructured, anchors.Tier1CompareConversionComparisonStructured},
		{"Tier 1 compare/minmax random comparisons", tier1CompareConversion.ComparisonRandom, anchors.Tier1CompareConversionComparisonRandom},
		{"Tier 1 compare/minmax random stream hashes", tier1CompareConversion.ComparisonRandomHashes, anchors.Tier1CompareConversionComparisonRandomHashes},
		{"Tier 1 compare/minmax total comparisons", tier1CompareConversion.ComparisonTotal, anchors.Tier1CompareConversionComparisonTotal},
		{"Tier 1 BID-to-integer total comparisons", tier1CompareConversion.ToIntegerTotal, anchors.Tier1CompareConversionToIntegerTotal},
		{"Tier 1 BID-to-integer random stream hashes", tier1CompareConversion.ToIntegerRandomHashes, anchors.Tier1CompareConversionToIntegerRandomHashes},
		{"Tier 1 BID-width operation counts", tier1CompareConversion.WidthOperations, anchors.Tier1CompareConversionWidthOperations},
		{"Tier 1 BID-width total comparisons", tier1CompareConversion.WidthTotal, anchors.Tier1CompareConversionWidthTotal},
		{"Tier 1 BID-width random stream hashes", tier1CompareConversion.WidthRandomHashes, anchors.Tier1CompareConversionWidthRandomHashes},
		{"Tier 1 BID-to-binary random stream hashes", tier1CompareConversion.BinaryRandomHashes, anchors.Tier1CompareConversionBinaryRandomHashes},
	} {
		if !reflect.DeepEqual(check.got, check.want) {
			t.Errorf("generated %s = %#v, anchor = %#v", check.label, check.got, check.want)
		}
	}
	rustTier1CompareConversion := loadRustTier1CompareConversionLongInventory(t)
	for _, check := range []struct {
		label string
		got   map[string]uint64
		want  map[string]uint64
	}{
		{"Rust Tier 1 compare/conversion boundary values", rustTier1CompareConversion.BoundaryValues, anchors.Tier1CompareConversionBoundaryValues},
		{"Rust Tier 1 compare/conversion semantic values", rustTier1CompareConversion.SemanticValues, anchors.Tier1CompareConversionSemanticValues},
		{"Rust Tier 1 compare/minmax structured comparisons", rustTier1CompareConversion.ComparisonStructured, anchors.Tier1CompareConversionComparisonStructured},
		{"Rust Tier 1 compare/minmax random comparisons", rustTier1CompareConversion.ComparisonRandom, anchors.Tier1CompareConversionComparisonRandom},
		{"Rust Tier 1 compare/minmax random stream hashes", rustTier1CompareConversion.ComparisonRandomHashes, anchors.Tier1CompareConversionComparisonRandomHashes},
		{"Rust Tier 1 compare/minmax total comparisons", rustTier1CompareConversion.ComparisonTotal, anchors.Tier1CompareConversionComparisonTotal},
		{"Rust Tier 1 BID-to-integer total comparisons", rustTier1CompareConversion.ToIntegerTotal, anchors.Tier1CompareConversionToIntegerTotal},
		{"Rust Tier 1 BID-to-integer random stream hashes", rustTier1CompareConversion.ToIntegerRandomHashes, anchors.Tier1CompareConversionToIntegerRandomHashes},
		{"Rust Tier 1 BID-width operation counts", rustTier1CompareConversion.WidthOperations, anchors.Tier1CompareConversionWidthOperations},
		{"Rust Tier 1 BID-width total comparisons", rustTier1CompareConversion.WidthTotal, anchors.Tier1CompareConversionWidthTotal},
		{"Rust Tier 1 BID-width random stream hashes", rustTier1CompareConversion.WidthRandomHashes, anchors.Tier1CompareConversionWidthRandomHashes},
		{"Rust Tier 1 BID-to-binary random stream hashes", rustTier1CompareConversion.BinaryRandomHashes, anchors.Tier1CompareConversionBinaryRandomHashes},
	} {
		if !reflect.DeepEqual(check.got, check.want) {
			t.Errorf("generated %s = %#v, anchor = %#v", check.label, check.got, check.want)
		}
	}
	for _, check := range []struct {
		label string
		got   uint64
		want  uint64
	}{
		{"Rust Tier 1 quiet predicate count", rustTier1CompareConversion.QuietPredicates, anchors.Tier1CompareConversionQuietPredicates},
		{"Rust Tier 1 min/max operation count", rustTier1CompareConversion.MinMaxOperations, anchors.Tier1CompareConversionMinMaxOperations},
		{"Rust Tier 1 BID-to-integer operation count", rustTier1CompareConversion.ToIntegerOperations, anchors.Tier1CompareConversionToIntegerOperations},
		{"Rust Tier 1 integer-constructor operation count", rustTier1CompareConversion.ConstructorOperations, anchors.Tier1CompareConversionConstructorOperations},
		{"Rust Tier 1 integer-constructor total comparisons", rustTier1CompareConversion.ConstructorTotal, anchors.Tier1CompareConversionConstructorTotal},
		{"Rust Tier 1 integer-constructor convenience checks", rustTier1CompareConversion.ConstructorConvenience, anchors.Tier1CompareConversionConstructorConvenience},
		{"Rust Tier 1 integer-constructor random stream hash", rustTier1CompareConversion.ConstructorRandomHash, anchors.Tier1CompareConversionConstructorRandomHash},
		{"Rust Tier 1 conversion structured comparisons", rustTier1CompareConversion.ConversionStructured, anchors.Tier1CompareConversionStructured},
		{"Rust Tier 1 conversion random comparisons", rustTier1CompareConversion.ConversionRandom, anchors.Tier1CompareConversionRandom},
		{"Rust Tier 1 conversion total comparisons", rustTier1CompareConversion.ConversionTotal, anchors.Tier1CompareConversionTotal},
	} {
		if check.got != check.want {
			t.Errorf("generated %s = %d, anchor = %d", check.label, check.got, check.want)
		}
	}
	for _, check := range []struct {
		label string
		got   uint64
		want  uint64
	}{
		{"Tier 1 quiet predicate count", tier1CompareConversion.QuietPredicates, anchors.Tier1CompareConversionQuietPredicates},
		{"Tier 1 min/max operation count", tier1CompareConversion.MinMaxOperations, anchors.Tier1CompareConversionMinMaxOperations},
		{"Tier 1 BID-to-integer operation count", tier1CompareConversion.ToIntegerOperations, anchors.Tier1CompareConversionToIntegerOperations},
		{"Tier 1 integer-constructor operation count", tier1CompareConversion.ConstructorOperations, anchors.Tier1CompareConversionConstructorOperations},
		{"Tier 1 integer-constructor total comparisons", tier1CompareConversion.ConstructorTotal, anchors.Tier1CompareConversionConstructorTotal},
		{"Tier 1 integer-constructor convenience checks", tier1CompareConversion.ConstructorConvenience, anchors.Tier1CompareConversionConstructorConvenience},
		{"Tier 1 integer-constructor random stream hash", tier1CompareConversion.ConstructorRandomHash, anchors.Tier1CompareConversionConstructorRandomHash},
		{"Tier 1 conversion structured comparisons", tier1CompareConversion.ConversionStructured, anchors.Tier1CompareConversionStructured},
		{"Tier 1 conversion random comparisons", tier1CompareConversion.ConversionRandom, anchors.Tier1CompareConversionRandom},
		{"Tier 1 conversion total comparisons", tier1CompareConversion.ConversionTotal, anchors.Tier1CompareConversionTotal},
	} {
		if check.got != check.want {
			t.Errorf("generated %s = %d, anchor = %d", check.label, check.got, check.want)
		}
	}

	if len(spec.DectestRuntimeSkipInventory) != len(anchors.DectestSuiteCases) {
		t.Errorf("dectest runtime skip inventory suite count = %d, anchor suites = %d",
			len(spec.DectestRuntimeSkipInventory), len(anchors.DectestSuiteCases))
	}
	for _, inventory := range spec.DectestRuntimeSkipInventory {
		want, ok := anchors.DectestSuiteCases[inventory.Suite]
		if !ok {
			t.Errorf("dectest suite %q has no anchor entry", inventory.Suite)
			continue
		}
		if inventory.Cases != want {
			t.Errorf("dectest suite %q case count = %d, anchor = %d", inventory.Suite, inventory.Cases, want)
		}
	}

	// Goport decTest leg: executed and skipped per fixed-width suite, pinned outside
	// the generation path so a self-consistent generator regression that shrinks the
	// executed oracle-op set moves a count here. executed = cases - sum(skip_reasons).
	if len(spec.DectestGoportRuntimeSkipInventory) != len(anchors.GoportDectestExecutedCases) {
		t.Errorf("goport dectest runtime skip inventory suite count = %d, executed-anchor suites = %d",
			len(spec.DectestGoportRuntimeSkipInventory), len(anchors.GoportDectestExecutedCases))
	}
	// Check the skipped-anchor map size too, so a stale extra suite key in
	// goport_dectest_skipped_cases cannot survive unnoticed (the executed map size
	// check alone would not catch it).
	if len(spec.DectestGoportRuntimeSkipInventory) != len(anchors.GoportDectestSkippedCases) {
		t.Errorf("goport dectest runtime skip inventory suite count = %d, skipped-anchor suites = %d",
			len(spec.DectestGoportRuntimeSkipInventory), len(anchors.GoportDectestSkippedCases))
	}
	// Flag-exemption anchors are exhaustive against the same suite set as the
	// executed anchors: every verified suite needs an entry, and a stale extra
	// suite key cannot survive.
	if len(spec.DectestGoportRuntimeSkipInventory) != len(anchors.GoportDectestFlagExemptCases) {
		t.Errorf("goport dectest runtime skip inventory suite count = %d, flag-exempt-anchor suites = %d",
			len(spec.DectestGoportRuntimeSkipInventory), len(anchors.GoportDectestFlagExemptCases))
	}
	for _, inventory := range spec.DectestGoportRuntimeSkipInventory {
		skipped := 0
		for _, n := range inventory.SkipReasons {
			skipped += n
		}
		executed := inventory.Cases - skipped
		if want, ok := anchors.GoportDectestExecutedCases[inventory.Suite]; !ok {
			t.Errorf("goport dectest suite %q has no executed anchor entry", inventory.Suite)
		} else if executed != want {
			t.Errorf("goport dectest suite %q executed = %d, anchor = %d", inventory.Suite, executed, want)
		}
		if want, ok := anchors.GoportDectestSkippedCases[inventory.Suite]; !ok {
			t.Errorf("goport dectest suite %q has no skipped anchor entry", inventory.Suite)
		} else if skipped != want {
			t.Errorf("goport dectest suite %q skipped = %d, anchor = %d", inventory.Suite, skipped, want)
		}
		flagExempt := 0
		for _, n := range inventory.FlagExemptions {
			flagExempt += n
		}
		if want, ok := anchors.GoportDectestFlagExemptCases[inventory.Suite]; !ok {
			t.Errorf("goport dectest suite %q has no flag-exempt anchor entry", inventory.Suite)
		} else if flagExempt != want {
			t.Errorf("goport dectest suite %q flag-exempt = %d, anchor = %d", inventory.Suite, flagExempt, want)
		}
		// Exempt cases must stay inside the executed set (their value and
		// quantum assertions still run): an exemption bucket larger than the
		// executed count marks a broken classifier.
		if flagExempt > executed {
			t.Errorf("goport dectest suite %q flag-exempt %d exceeds executed %d", inventory.Suite, flagExempt, executed)
		}
	}

	// Rust decTest leg: dectest_rust_codegen.go's runner generator calls the
	// identical countDectestGoportSuiteCoverage function the goport leg above is
	// pinned against, so rust_dectest_{executed,skipped,flag_exempt}_cases must be
	// value-identical to their goport_dectest_* counterparts against the SAME
	// spec.DectestGoportRuntimeSkipInventory source -- not an independent
	// re-measurement, but a direct check against the two anchor tables silently
	// drifting apart (which would mean the two legs' generators no longer agree on
	// which decTest cases execute). TestRustAndGoportDectestDispatchAgrees
	// (dectest_rust_crosscheck_test.go) separately re-derives this from the live
	// generator call and also cross-checks every dispatch function name.
	if len(spec.DectestGoportRuntimeSkipInventory) != len(anchors.RustDectestExecutedCases) {
		t.Errorf("goport dectest runtime skip inventory suite count = %d, rust executed-anchor suites = %d",
			len(spec.DectestGoportRuntimeSkipInventory), len(anchors.RustDectestExecutedCases))
	}
	if len(spec.DectestGoportRuntimeSkipInventory) != len(anchors.RustDectestSkippedCases) {
		t.Errorf("goport dectest runtime skip inventory suite count = %d, rust skipped-anchor suites = %d",
			len(spec.DectestGoportRuntimeSkipInventory), len(anchors.RustDectestSkippedCases))
	}
	if len(spec.DectestGoportRuntimeSkipInventory) != len(anchors.RustDectestFlagExemptCases) {
		t.Errorf("goport dectest runtime skip inventory suite count = %d, rust flag-exempt-anchor suites = %d",
			len(spec.DectestGoportRuntimeSkipInventory), len(anchors.RustDectestFlagExemptCases))
	}
	for _, inventory := range spec.DectestGoportRuntimeSkipInventory {
		skipped := 0
		for _, n := range inventory.SkipReasons {
			skipped += n
		}
		executed := inventory.Cases - skipped
		flagExempt := 0
		for _, n := range inventory.FlagExemptions {
			flagExempt += n
		}
		if want, ok := anchors.RustDectestExecutedCases[inventory.Suite]; !ok {
			t.Errorf("rust dectest suite %q has no executed anchor entry", inventory.Suite)
		} else if executed != want {
			t.Errorf("rust dectest suite %q executed = %d, anchor = %d", inventory.Suite, executed, want)
		} else if goportWant := anchors.GoportDectestExecutedCases[inventory.Suite]; executed != goportWant {
			t.Errorf("rust dectest suite %q executed = %d diverged from goport_dectest_executed_cases = %d", inventory.Suite, executed, goportWant)
		}
		if want, ok := anchors.RustDectestSkippedCases[inventory.Suite]; !ok {
			t.Errorf("rust dectest suite %q has no skipped anchor entry", inventory.Suite)
		} else if skipped != want {
			t.Errorf("rust dectest suite %q skipped = %d, anchor = %d", inventory.Suite, skipped, want)
		} else if goportWant := anchors.GoportDectestSkippedCases[inventory.Suite]; skipped != goportWant {
			t.Errorf("rust dectest suite %q skipped = %d diverged from goport_dectest_skipped_cases = %d", inventory.Suite, skipped, goportWant)
		}
		if want, ok := anchors.RustDectestFlagExemptCases[inventory.Suite]; !ok {
			t.Errorf("rust dectest suite %q has no flag-exempt anchor entry", inventory.Suite)
		} else if flagExempt != want {
			t.Errorf("rust dectest suite %q flag-exempt = %d, anchor = %d", inventory.Suite, flagExempt, want)
		} else if goportWant := anchors.GoportDectestFlagExemptCases[inventory.Suite]; flagExempt != goportWant {
			t.Errorf("rust dectest suite %q flag-exempt = %d diverged from goport_dectest_flag_exempt_cases = %d", inventory.Suite, flagExempt, goportWant)
		}
	}

	if len(spec.ReadtestProfileInventory) != 1 {
		t.Fatalf("readtest profile inventory count = %d, want 1", len(spec.ReadtestProfileInventory))
	}
	profile := spec.ReadtestProfileInventory[0]
	if profile.TotalFunctions != anchors.ReadtestProfileFunctionsTotal {
		t.Errorf("readtest profile total functions = %d, anchor = %d", profile.TotalFunctions, anchors.ReadtestProfileFunctionsTotal)
	}
	if profile.SelectedFunctions != anchors.ReadtestProfileFunctionsSelected {
		t.Errorf("readtest profile selected functions = %d, anchor = %d", profile.SelectedFunctions, anchors.ReadtestProfileFunctionsSelected)
	}

	// Row-level skip verification: the generator records, per named reason, readtest.in
	// rows that belong to a selected function but were dropped, plus the accepted
	// row count. These are pinned outside the generation path (like the decTest
	// suite-case anchors) so a parser change that silently drops rows moves a
	// reason count here instead of only shrinking the total.
	if profile.RowsAccepted != anchors.ReadtestProfileRowsAccepted {
		t.Errorf("readtest profile rows accepted = %d, anchor = %d", profile.RowsAccepted, anchors.ReadtestProfileRowsAccepted)
	}
	if len(profile.RowSkipReasons) != len(anchors.ReadtestProfileRowSkipReasons) {
		t.Errorf("readtest profile row skip reason count = %d, anchor reasons = %d (%v vs %v)",
			len(profile.RowSkipReasons), len(anchors.ReadtestProfileRowSkipReasons), profile.RowSkipReasons, anchors.ReadtestProfileRowSkipReasons)
	}
	for reason, want := range anchors.ReadtestProfileRowSkipReasons {
		got, ok := profile.RowSkipReasons[reason]
		if !ok {
			t.Errorf("readtest profile row skip reason %q missing from generated inventory", reason)
			continue
		}
		if got != want {
			t.Errorf("readtest profile row skip reason %q = %d, anchor = %d", reason, got, want)
		}
	}
	// Sanity: accepted rows plus IEEE deviation supplement must equal the pinned
	// readtest case total, tying this anchor to the existing total anchor.
	if profile.RowsAccepted > anchors.ReadtestCasesTotal {
		t.Errorf("readtest profile rows accepted %d exceeds readtest case total %d", profile.RowsAccepted, anchors.ReadtestCasesTotal)
	}

	for _, item := range []struct {
		label  string
		path   string
		anchor int
	}{
		{"goport readtest dispatch inventory", filepath.Join("..", "..", "generated", "testspec", "goport_readtest_dispatch_inventory.json"), anchors.GoportReadtestDispatchedFuncs},
		{"rust readtest dispatch inventory", filepath.Join("..", "..", "generated", "testspec", "rust_readtest_dispatch_inventory.json"), anchors.RustReadtestDispatchedFuncs},
	} {
		raw, err := os.ReadFile(item.path)
		if err != nil {
			t.Fatalf("read %s: %v", item.label, err)
		}
		var inventory dispatchInventoryCounts
		if err := json.Unmarshal(raw, &inventory); err != nil {
			t.Fatalf("unmarshal %s: %v", item.label, err)
		}
		if inventory.Dispatched != item.anchor {
			t.Errorf("%s dispatched = %d, anchor = %d", item.label, inventory.Dispatched, item.anchor)
		}
	}

	// The Rust readtest generator pins per-suite passed-case counts inside the
	// generated runner and records the same counts in its dispatch inventory; the
	// inventory counts must match the hand-maintained anchors so a symmetric
	// generator regression (runner template and generation-time counter losing
	// the same rows) cannot move the pins silently.
	rustInventoryRaw, err := os.ReadFile(filepath.Join("..", "..", "generated", "testspec", "rust_readtest_dispatch_inventory.json"))
	if err != nil {
		t.Fatalf("read rust readtest dispatch inventory: %v", err)
	}
	var rustInventory struct {
		SuitePasses []struct {
			Suite          string `json:"suite"`
			Filter         string `json:"filter"`
			ExpectedPasses int    `json:"expected_passes"`
		} `json:"suite_expected_passes"`
	}
	if err := json.Unmarshal(rustInventoryRaw, &rustInventory); err != nil {
		t.Fatalf("unmarshal rust readtest dispatch inventory: %v", err)
	}
	if len(rustInventory.SuitePasses) != len(anchors.RustReadtestSuitePasses) {
		t.Errorf("rust readtest inventory suite count = %d, anchor suites = %d",
			len(rustInventory.SuitePasses), len(anchors.RustReadtestSuitePasses))
	}
	seenSuites := map[string]bool{}
	for _, suite := range rustInventory.SuitePasses {
		if seenSuites[suite.Suite] {
			t.Errorf("rust readtest inventory duplicates suite %q", suite.Suite)
			continue
		}
		seenSuites[suite.Suite] = true
		want, ok := anchors.RustReadtestSuitePasses[suite.Suite]
		if !ok {
			t.Errorf("rust readtest suite %q (filter %q) has no anchor entry", suite.Suite, suite.Filter)
			continue
		}
		if suite.ExpectedPasses != want {
			t.Errorf("rust readtest suite %q expected passes = %d, anchor = %d", suite.Suite, suite.ExpectedPasses, want)
		}
	}

	vectorsRaw, err := os.ReadFile(filepath.Join("..", "..", "..", "bid754-codec-vectors", "vectors.json"))
	if err != nil {
		t.Fatalf("read bid754-codec-vectors/vectors.json: %v", err)
	}
	var vectors struct {
		Vectors []struct {
			Type      string `json:"type"`
			Canonical bool   `json:"canonical"`
		} `json:"vectors"`
		RejectVectors []struct {
			Channel  string `json:"channel"`
			Requires string `json:"requires"`
		} `json:"reject_vectors"`
		StringVectors []struct {
			Input    string `json:"input"`
			Expected string `json:"expected"`
		} `json:"string_vectors"`
	}
	if err := json.Unmarshal(vectorsRaw, &vectors); err != nil {
		t.Fatalf("unmarshal vectors.json: %v", err)
	}
	if len(vectors.Vectors) != anchors.BidCodecVectorsTotal {
		t.Errorf("checked-in BID codec vector count = %d, anchor = %d", len(vectors.Vectors), anchors.BidCodecVectorsTotal)
	}
	vectorsByWidth := map[string]int{"bid32": 0, "bid64": 0, "bid128": 0}
	canonicalByWidth := map[string]int{"bid32": 0, "bid64": 0, "bid128": 0}
	for _, vector := range vectors.Vectors {
		if _, ok := vectorsByWidth[vector.Type]; !ok {
			t.Errorf("checked-in BID codec vector has unknown width %q", vector.Type)
			continue
		}
		vectorsByWidth[vector.Type]++
		if vector.Canonical {
			canonicalByWidth[vector.Type]++
		}
	}
	if !reflect.DeepEqual(vectorsByWidth, anchors.BidCodecVectorsByWidth) {
		t.Errorf("checked-in BID codec vectors by width = %v, anchors = %v", vectorsByWidth, anchors.BidCodecVectorsByWidth)
	}
	if !reflect.DeepEqual(canonicalByWidth, anchors.BidCodecCanonicalByWidth) {
		t.Errorf("checked-in canonical BID codec vectors by width = %v, anchors = %v", canonicalByWidth, anchors.BidCodecCanonicalByWidth)
	}
	if len(vectors.RejectVectors) != anchors.BidCodecRejectVectorsTotal {
		t.Errorf("checked-in BID codec reject vector count = %d, anchor = %d", len(vectors.RejectVectors), anchors.BidCodecRejectVectorsTotal)
	}
	if len(vectors.StringVectors) != anchors.BidCodecStringVectorsTotal {
		t.Errorf("checked-in BID codec string vector count = %d, anchor = %d", len(vectors.StringVectors), anchors.BidCodecStringVectorsTotal)
	}
	d32Inventory := loadBidCodecD32HarnessInventory(t)
	if d32Inventory.RawPatterns != anchors.BidCodecD32ExhaustiveRawPatterns {
		t.Errorf("generated decimal32 exhaustive raw patterns = %d, anchor = %d", d32Inventory.RawPatterns, anchors.BidCodecD32ExhaustiveRawPatterns)
	}
	if d32Inventory.StructuredCases != anchors.BidCodecD32StructuredCases {
		t.Errorf("generated decimal32 structured cases = %d, anchor = %d", d32Inventory.StructuredCases, anchors.BidCodecD32StructuredCases)
	}
	if !reflect.DeepEqual(d32Inventory.RawClasses, anchors.BidCodecD32ExhaustiveRawClasses) {
		t.Errorf("generated decimal32 exhaustive raw classes = %v, anchors = %v", d32Inventory.RawClasses, anchors.BidCodecD32ExhaustiveRawClasses)
	}
	if !reflect.DeepEqual(d32Inventory.StructuredClasses, anchors.BidCodecD32StructuredClasses) {
		t.Errorf("generated decimal32 structured classes = %v, anchors = %v", d32Inventory.StructuredClasses, anchors.BidCodecD32StructuredClasses)
	}
	d64128Inventory := loadBidCodecD64128HarnessInventory(t)
	if d64128Inventory.DifferentialCasesPerWidth != anchors.BidCodecD64128DifferentialCases {
		t.Errorf("generated decimal64/128 differential cases per width = %d, anchor = %d",
			d64128Inventory.DifferentialCasesPerWidth, anchors.BidCodecD64128DifferentialCases)
	}
	if !reflect.DeepEqual(d64128Inventory.DifferentialClasses, anchors.BidCodecD64128DifferentialClass) {
		t.Errorf("generated decimal64/128 differential classes = %v, anchors = %v",
			d64128Inventory.DifferentialClasses, anchors.BidCodecD64128DifferentialClass)
	}
	if !reflect.DeepEqual(d64128Inventory.BoundaryRawCases, anchors.BidCodecD64128BoundaryRawCases) {
		t.Errorf("generated decimal64/128 boundary raw cases = %v, anchors = %v",
			d64128Inventory.BoundaryRawCases, anchors.BidCodecD64128BoundaryRawCases)
	}
	if !reflect.DeepEqual(d64128Inventory.StructuredCases, anchors.BidCodecD64128StructuredCases) {
		t.Errorf("generated decimal64/128 structured cases = %v, anchors = %v",
			d64128Inventory.StructuredCases, anchors.BidCodecD64128StructuredCases)
	}
	if !reflect.DeepEqual(d64128Inventory.StructuredClasses, anchors.BidCodecD64128StructuredClasses) {
		t.Errorf("generated decimal64/128 structured classes = %v, anchors = %v",
			d64128Inventory.StructuredClasses, anchors.BidCodecD64128StructuredClasses)
	}
	const bidCodecRandomCasesPerWidth = 5000
	if vectorsByWidth["bid64"] != bidCodecRandomCasesPerWidth+int(d64128Inventory.BoundaryRawCases["decimal64"]) ||
		vectorsByWidth["bid128"] != bidCodecRandomCasesPerWidth+int(d64128Inventory.BoundaryRawCases["decimal128"]) {
		t.Errorf("shared vector widths do not contain the full long-harness boundary inventories plus %d random cases: vectors=%v boundaries=%v",
			bidCodecRandomCasesPerWidth, vectorsByWidth, d64128Inventory.BoundaryRawCases)
	}
	rawKindTotal := d32Inventory.RawClasses["normal"] + d32Inventory.RawClasses["zero"] +
		d32Inventory.RawClasses["infinity"] + d32Inventory.RawClasses["qnan"] + d32Inventory.RawClasses["snan"]
	if rawKindTotal != d32Inventory.RawPatterns {
		t.Errorf("generated decimal32 raw Kind classes sum to %d, raw domain = %d", rawKindTotal, d32Inventory.RawPatterns)
	}
	if d32Inventory.RawClasses["canonical"]+d32Inventory.RawClasses["noncanonical"] != d32Inventory.RawPatterns {
		t.Errorf("generated decimal32 canonical classes do not partition raw domain")
	}
	structuredKindTotal := d32Inventory.StructuredClasses["normal"] + d32Inventory.StructuredClasses["zero"] +
		d32Inventory.StructuredClasses["infinity"] + d32Inventory.StructuredClasses["qnan"] + d32Inventory.StructuredClasses["snan"]
	if structuredKindTotal != d32Inventory.StructuredCases {
		t.Errorf("generated decimal32 structured Kind classes sum to %d, structured domain = %d", structuredKindTotal, d32Inventory.StructuredCases)
	}
	for _, width := range []string{"decimal64", "decimal128"} {
		differential := d64128Inventory.DifferentialClasses[width]
		kindTotal := differential["normal"] + differential["zero"] + differential["infinity"] +
			differential["qnan"] + differential["snan"]
		if kindTotal != d64128Inventory.DifferentialCasesPerWidth {
			t.Errorf("generated %s differential Kind classes sum to %d, domain = %d",
				width, kindTotal, d64128Inventory.DifferentialCasesPerWidth)
		}
		if differential["canonical"]+differential["noncanonical"] != d64128Inventory.DifferentialCasesPerWidth {
			t.Errorf("generated %s differential canonical classes do not partition domain", width)
		}
		structured := d64128Inventory.StructuredClasses[width]
		structuredTotal := structured["normal"] + structured["zero"] + structured["infinity"] +
			structured["qnan"] + structured["snan"]
		if structuredTotal != d64128Inventory.StructuredCases[width] {
			t.Errorf("generated %s structured Kind classes sum to %d, domain = %d",
				width, structuredTotal, d64128Inventory.StructuredCases[width])
		}
		if structured["canonical"]+structured["noncanonical"] != d64128Inventory.StructuredCases[width] {
			t.Errorf("generated %s structured canonical classes do not partition domain", width)
		}
	}

	// Reject distribution pins. Pinning only the reject total would let a
	// generator regression that tags every record with an unsupported requires
	// capability pass as "all skipped" in every consumer; the per-channel,
	// per-requires-tag, and per-language consumed counts below close that.
	rejectChannelCounts := map[string]int{}
	rejectRequiresCounts := map[string]int{}
	for _, r := range vectors.RejectVectors {
		rejectChannelCounts[r.Channel]++
		if r.Requires != "" {
			rejectRequiresCounts[r.Requires]++
		}
	}
	if len(rejectChannelCounts) != len(anchors.BidCodecRejectChannels) {
		t.Errorf("reject channel kinds = %v, anchor channels = %v", rejectChannelCounts, anchors.BidCodecRejectChannels)
	}
	for channel, want := range anchors.BidCodecRejectChannels {
		if got := rejectChannelCounts[channel]; got != want {
			t.Errorf("reject channel %q count = %d, anchor = %d", channel, got, want)
		}
	}
	if len(rejectRequiresCounts) != len(anchors.BidCodecRejectRequires) {
		t.Errorf("reject requires tags = %v, anchor tags = %v", rejectRequiresCounts, anchors.BidCodecRejectRequires)
	}
	for tag, want := range anchors.BidCodecRejectRequires {
		if got := rejectRequiresCounts[tag]; got != want {
			t.Errorf("reject requires tag %q count = %d, anchor = %d", tag, got, want)
		}
	}

	// Hand-maintained language -> capability map, written out independently of
	// the generator table in bid_codec_reject_vectors.go. The two are
	// cross-checked both ways so a capability edit on either side that is not
	// mirrored (and re-approved via this anchor test) fails here, outside the
	// generation path.
	handRejectCapabilities := map[string][]string{
		// Go's Components.Payload is now a signed *big.Int, so Go can construct a
		// negative payload (negative_payload) in addition to a big/negative
		// coefficient. Rust (u128) and Swift (UInt64 pair) stay unsigned
		// fixed-width and skip every requires-tagged record. "rust_full" is the
		// additional consumer generated for the bid754-rs full library's embedded
		// bid_codec module (same u128 field types as the standalone Rust codec,
		// so the same empty capability set).
		"go":        {"bignum_coefficient", "negative_coefficient", "negative_payload"},
		"rust":      {},
		"rust_full": {},
		"java":      {"bignum_coefficient", "negative_coefficient", "negative_payload"},
		"python":    {"bignum_coefficient", "negative_coefficient", "negative_payload"},
		"js":        {"bignum_coefficient", "negative_coefficient", "negative_payload"},
		"swift":     {},
	}
	for lang, handCaps := range handRejectCapabilities {
		genCaps, ok := bidCodecRejectCapabilities[lang]
		if !ok {
			t.Errorf("language %q in hand capability map missing from generator capability table", lang)
			continue
		}
		handSet := map[string]bool{}
		for _, c := range handCaps {
			handSet[c] = true
		}
		genSet := map[string]bool{}
		for _, c := range genCaps {
			genSet[c] = true
		}
		if len(handSet) != len(genSet) {
			t.Errorf("language %q capabilities: hand map %v, generator table %v", lang, handCaps, genCaps)
			continue
		}
		for c := range handSet {
			if !genSet[c] {
				t.Errorf("language %q capability %q in hand map missing from generator table", lang, c)
			}
		}
	}
	for lang := range bidCodecRejectCapabilities {
		if _, ok := handRejectCapabilities[lang]; !ok {
			t.Errorf("language %q in generator capability table missing from hand capability map", lang)
		}
	}

	// Per-language consumed counts, recomputed from the checked-in JSON requires
	// tags plus the hand-maintained capability map, must equal the anchors. The
	// generated runners pin the same numbers as generation-time constants, which
	// a generator regression would shift together; this recomputation is the
	// independent copy.
	if len(anchors.BidCodecRejectConsumedByLanguage) != len(handRejectCapabilities) {
		t.Errorf("anchor consumed-by-language has %d languages, hand capability map has %d",
			len(anchors.BidCodecRejectConsumedByLanguage), len(handRejectCapabilities))
	}
	for lang, want := range anchors.BidCodecRejectConsumedByLanguage {
		handCaps, ok := handRejectCapabilities[lang]
		if !ok {
			t.Errorf("anchor consumed-by-language language %q missing from hand capability map", lang)
			continue
		}
		capsSet := map[string]bool{}
		for _, c := range handCaps {
			capsSet[c] = true
		}
		consumed := 0
		for _, r := range vectors.RejectVectors {
			if r.Requires == "" || capsSet[r.Requires] {
				consumed++
			}
		}
		if consumed != want {
			t.Errorf("language %q recomputed reject consumed = %d, anchor = %d", lang, consumed, want)
		}
	}

	// go_full consumer pins. The bid754-go full library's public parse surface
	// consumes the from_string reject channel and every string_vectors record;
	// the encode/to_string channels are channel-skipped because bid754-go has
	// no public Components construction surface. The split is channel-derived,
	// not capability-derived, so go_full stays outside the consumed-by-language
	// map above and is recomputed here directly from the checked-in JSON.
	goFullConsumed, goFullSkipped := 0, 0
	for _, r := range vectors.RejectVectors {
		if r.Channel == "from_string" {
			goFullConsumed++
		} else {
			goFullSkipped++
		}
	}
	if goFullConsumed != anchors.BidCodecRejectGoFullConsumed {
		t.Errorf("go_full recomputed reject consumed = %d, anchor = %d", goFullConsumed, anchors.BidCodecRejectGoFullConsumed)
	}
	if goFullSkipped != anchors.BidCodecRejectGoFullChannelSkipped {
		t.Errorf("go_full recomputed reject channel-skipped = %d, anchor = %d", goFullSkipped, anchors.BidCodecRejectGoFullChannelSkipped)
	}
	if len(vectors.StringVectors) != anchors.BidCodecStringGoFullConsumed {
		t.Errorf("go_full string_vectors consumed anchor = %d, want the ungated record total %d",
			anchors.BidCodecStringGoFullConsumed, len(vectors.StringVectors))
	}

	// rust_full_parse consumer: the bid754-rs public parse surface. It splits
	// the reject channels exactly like go_full (the public Decimal API exposes
	// no Components construction surface in either language), so its counts are
	// equal to the go_full counts by construction. They are pinned separately
	// anyway: the equality is the contract being checked, and collapsing the
	// two onto one anchor would let a Rust leg that silently stopped consuming
	// the channel keep passing on the Go leg's count.
	if goFullConsumed != anchors.BidCodecRejectRustFullParseConsumed {
		t.Errorf("rust_full_parse recomputed reject consumed = %d, anchor = %d", goFullConsumed, anchors.BidCodecRejectRustFullParseConsumed)
	}
	if goFullSkipped != anchors.BidCodecRejectRustFullParseChannelSkipped {
		t.Errorf("rust_full_parse recomputed reject channel-skipped = %d, anchor = %d", goFullSkipped, anchors.BidCodecRejectRustFullParseChannelSkipped)
	}
	if anchors.BidCodecRejectRustFullParseConsumed != anchors.BidCodecRejectGoFullConsumed {
		t.Errorf("rust_full_parse consumed anchor = %d, go_full = %d; the two public parse surfaces must consume the same channel",
			anchors.BidCodecRejectRustFullParseConsumed, anchors.BidCodecRejectGoFullConsumed)
	}

	// string_vectors consumed pins. The success channel is capability-ungated
	// by design: every record is a pure ASCII input/expected string pair
	// constructible in every language, so each consumer must consume every
	// record. Each per-language anchor must equal the checked-in record total,
	// and the anchor map is exhaustive against the same seven-language set as
	// the reject consumed map, so a language cannot silently drop out of the
	// success channel (or carry a stale entry) without failing here.
	for i, sv := range vectors.StringVectors {
		if sv.Input == "" || sv.Expected == "" {
			t.Errorf("string_vectors[%d] carries an empty input or expected field: %+v", i, sv)
		}
	}
	if len(anchors.BidCodecStringConsumedByLanguage) != len(handRejectCapabilities) {
		t.Errorf("anchor string consumed-by-language has %d languages, hand capability map has %d",
			len(anchors.BidCodecStringConsumedByLanguage), len(handRejectCapabilities))
	}
	for lang, want := range anchors.BidCodecStringConsumedByLanguage {
		if _, ok := handRejectCapabilities[lang]; !ok {
			t.Errorf("anchor string consumed-by-language language %q missing from hand capability map", lang)
			continue
		}
		if want != len(vectors.StringVectors) {
			t.Errorf("language %q string_vectors consumed anchor = %d, want the ungated record total %d",
				lang, want, len(vectors.StringVectors))
		}
	}
	for lang := range handRejectCapabilities {
		if _, ok := anchors.BidCodecStringConsumedByLanguage[lang]; !ok {
			t.Errorf("language %q missing from anchor string consumed-by-language map", lang)
		}
	}

	// Public-API routing gate: the census inventory (symbol/mapped/excluded counts)
	// and the generated parity runner (wrapper/case constants) both pin the same
	// surface at generation time, so a generator regression that shrinks either
	// one also shrinks its own pin and passes locally. Compare all three surfaces
	// -- the external anchors, the checked-in inventory JSON, and the checked-in
	// runner constants -- so any divergence between them fails here, outside the
	// generation path. The mapped wrapper count lives in both artifacts and both
	// must equal the single external anchor.
	t.Run("PublicAPIRoutingGate", func(t *testing.T) {
		inventoryRaw, err := os.ReadFile(filepath.Join("..", "..", "generated", "testspec", "public_api_routing_inventory.json"))
		if err != nil {
			t.Fatalf("read public_api_routing_inventory.json: %v", err)
		}
		var inventory struct {
			Total    int `json:"total"`
			Mapped   int `json:"mapped"`
			Excluded int `json:"excluded"`
		}
		if err := json.Unmarshal(inventoryRaw, &inventory); err != nil {
			t.Fatalf("unmarshal public_api_routing_inventory.json: %v", err)
		}

		runnerWrappers, runnerCases := loadPublicParityRunnerConstants(t)

		if inventory.Total != anchors.PublicAPISymbolsTotal {
			t.Errorf("public API inventory total symbols = %d, anchor public_api_symbols_total = %d", inventory.Total, anchors.PublicAPISymbolsTotal)
		}
		if inventory.Mapped != anchors.PublicAPIParityWrappers {
			t.Errorf("public API inventory mapped wrappers = %d, anchor public_api_parity_wrappers = %d", inventory.Mapped, anchors.PublicAPIParityWrappers)
		}
		if inventory.Excluded != anchors.PublicAPIParityExcluded {
			t.Errorf("public API inventory excluded = %d, anchor public_api_parity_excluded = %d", inventory.Excluded, anchors.PublicAPIParityExcluded)
		}
		if runnerWrappers != anchors.PublicAPIParityWrappers {
			t.Errorf("generated runner expectedPublicParityWrappers = %d, anchor public_api_parity_wrappers = %d", runnerWrappers, anchors.PublicAPIParityWrappers)
		}
		if runnerCases != anchors.PublicAPIParityCasesTotal {
			t.Errorf("generated runner expectedPublicParityCases = %d, anchor public_api_parity_cases_total = %d", runnerCases, anchors.PublicAPIParityCasesTotal)
		}

		flaglessTargets, flaglessCases := loadPublicParityFlaglessConstants(t)
		if flaglessTargets != anchors.PublicAPIFlaglessSiblingTargets {
			t.Errorf("generated runner expectedPublicParityFlaglessSiblingTargets = %d, anchor public_api_flagless_sibling_targets = %d", flaglessTargets, anchors.PublicAPIFlaglessSiblingTargets)
		}
		if flaglessCases != anchors.PublicAPIFlaglessSiblingCasesTotal {
			t.Errorf("generated runner expectedPublicParityFlaglessSiblingCases = %d, anchor public_api_flagless_sibling_cases_total = %d", flaglessCases, anchors.PublicAPIFlaglessSiblingCasesTotal)
		}
		// Belt-and-suspenders: the two generated artifacts must agree with each
		// other on the wrapper count independent of the anchor comparison above.
		if inventory.Mapped != runnerWrappers {
			t.Errorf("public API inventory mapped wrappers = %d disagrees with generated runner expectedPublicParityWrappers = %d", inventory.Mapped, runnerWrappers)
		}
		if inventory.Total != inventory.Mapped+inventory.Excluded {
			t.Errorf("public API inventory total %d != mapped %d + excluded %d", inventory.Total, inventory.Mapped, inventory.Excluded)
		}
	})

	// Rust public-API parity gate: devtools/generated/testspec/
	// rust_api_surface_inventory.json's emitted-row count and the generated Rust
	// runner's own pinned wrapper/case/by-shape constants both pin the same
	// surface at generation time (a generator regression that shrinks either
	// one also shrinks its own pin and passes locally). Compare all three
	// surfaces -- the external anchors, the checked-in inventory JSON, and the
	// checked-in Rust runner constants -- so any divergence fails here,
	// outside the generation path.
	t.Run("RustPublicAPIParityGate", func(t *testing.T) {
		inventoryRaw, err := os.ReadFile(filepath.Join("..", "..", "generated", "testspec", "rust_api_surface_inventory.json"))
		if err != nil {
			t.Fatalf("read rust_api_surface_inventory.json: %v", err)
		}
		var inventory struct {
			Emitted int `json:"emitted"`
			Rows    []struct {
				Status string `json:"status"`
			} `json:"rows"`
		}
		if err := json.Unmarshal(inventoryRaw, &inventory); err != nil {
			t.Fatalf("unmarshal rust_api_surface_inventory.json: %v", err)
		}
		emittedRows := 0
		for _, row := range inventory.Rows {
			if row.Status == "emitted" {
				emittedRows++
			}
		}
		if emittedRows != inventory.Emitted {
			t.Errorf("rust_api_surface_inventory.json emitted field = %d, recounted emitted rows = %d", inventory.Emitted, emittedRows)
		}

		runnerWrappers, runnerCases, runnerByShape := loadRustParityRunnerConstants(t)

		if inventory.Emitted != anchors.RustPublicAPIParityWrappers {
			t.Errorf("rust API surface inventory emitted = %d, anchor rust_public_api_parity_wrappers = %d", inventory.Emitted, anchors.RustPublicAPIParityWrappers)
		}
		if runnerWrappers != anchors.RustPublicAPIParityWrappers {
			t.Errorf("generated rust runner EXPECTED_PARITY_WRAPPERS = %d, anchor rust_public_api_parity_wrappers = %d", runnerWrappers, anchors.RustPublicAPIParityWrappers)
		}
		if runnerCases != anchors.RustPublicAPIParityCasesTotal {
			t.Errorf("generated rust runner EXPECTED_PARITY_CASES = %d, anchor rust_public_api_parity_cases_total = %d", runnerCases, anchors.RustPublicAPIParityCasesTotal)
		}
		// Belt-and-suspenders: the two generated artifacts must agree with each
		// other on the wrapper count independent of the anchor comparison above.
		if inventory.Emitted != runnerWrappers {
			t.Errorf("rust API surface inventory emitted = %d disagrees with generated rust runner EXPECTED_PARITY_WRAPPERS = %d", inventory.Emitted, runnerWrappers)
		}
		if len(runnerByShape) != len(anchors.RustPublicAPIParityCasesByShape) {
			t.Errorf("generated rust runner has %d shapes, anchor rust_public_api_parity_cases_by_shape has %d", len(runnerByShape), len(anchors.RustPublicAPIParityCasesByShape))
		}
		sumByShape := 0
		for shape, want := range anchors.RustPublicAPIParityCasesByShape {
			got, ok := runnerByShape[shape]
			if !ok {
				t.Errorf("generated rust runner is missing shape %q present in the anchor", shape)
				continue
			}
			if got != want {
				t.Errorf("generated rust runner shape %q cases = %d, anchor = %d", shape, got, want)
			}
			sumByShape += want
		}
		if sumByShape != anchors.RustPublicAPIParityCasesTotal {
			t.Errorf("sum of anchor rust_public_api_parity_cases_by_shape = %d != anchor rust_public_api_parity_cases_total = %d", sumByShape, anchors.RustPublicAPIParityCasesTotal)
		}
	})

	// Rust flagless-sibling equivalence leg: the Rust mirror of the Go leg
	// checked above. Both legs are emitted from the same generator tables
	// (mutation_witness_corpus.go) and the same shared corpus, so their target
	// and case counts are equal by construction; that equality is asserted
	// here, outside the generation path, so a change that shrinks one language's
	// leg without the other fails even though each runner's own self-check
	// would still pass.
	t.Run("RustPublicAPIFlaglessSiblingLeg", func(t *testing.T) {
		rustTargets, rustCases := loadRustParityFlaglessConstants(t)
		if rustTargets != anchors.RustPublicAPIFlaglessSiblingTargets {
			t.Errorf("generated rust runner EXPECTED_FLAGLESS_SIBLING_TARGETS = %d, anchor rust_public_api_flagless_sibling_targets = %d", rustTargets, anchors.RustPublicAPIFlaglessSiblingTargets)
		}
		if rustCases != anchors.RustPublicAPIFlaglessSiblingCasesTotal {
			t.Errorf("generated rust runner EXPECTED_FLAGLESS_SIBLING_CASES = %d, anchor rust_public_api_flagless_sibling_cases_total = %d", rustCases, anchors.RustPublicAPIFlaglessSiblingCasesTotal)
		}
		if anchors.RustPublicAPIFlaglessSiblingTargets != anchors.PublicAPIFlaglessSiblingTargets {
			t.Errorf("anchor rust_public_api_flagless_sibling_targets = %d != anchor public_api_flagless_sibling_targets = %d; the two legs are emitted from the same generator tables and must stay equal", anchors.RustPublicAPIFlaglessSiblingTargets, anchors.PublicAPIFlaglessSiblingTargets)
		}
		if anchors.RustPublicAPIFlaglessSiblingCasesTotal != anchors.PublicAPIFlaglessSiblingCasesTotal {
			t.Errorf("anchor rust_public_api_flagless_sibling_cases_total = %d != anchor public_api_flagless_sibling_cases_total = %d; the two legs consume the same corpus, seeds, and witness rows and must stay equal", anchors.RustPublicAPIFlaglessSiblingCasesTotal, anchors.PublicAPIFlaglessSiblingCasesTotal)
		}
		goTargets, goCases := loadPublicParityFlaglessConstants(t)
		if rustTargets != goTargets {
			t.Errorf("generated rust runner EXPECTED_FLAGLESS_SIBLING_TARGETS = %d disagrees with the Go runner's expectedPublicParityFlaglessSiblingTargets = %d", rustTargets, goTargets)
		}
		if rustCases != goCases {
			t.Errorf("generated rust runner EXPECTED_FLAGLESS_SIBLING_CASES = %d disagrees with the Go runner's expectedPublicParityFlaglessSiblingCases = %d", rustCases, goCases)
		}
	})

	// Rust public-API constants gate: rust_api_surface_inventory.json's
	// SEPARATE "constants" section (the 12 ZERO/ONE/PI/E census excluded_constant_accessor
	// symbols -- outside the mapped public-symbol set the gate above covers) and the generated
	// Rust runner's own per-width EXPECTED_CONST_PARITY_CASES_<digits> self-checks both
	// pin the same 12-constant surface at generation time. Compare all three surfaces --
	// the external anchor, the checked-in inventory JSON, and the checked-in Rust runner
	// constants -- so any divergence fails here, outside the generation path.
	t.Run("RustPublicAPIConstantsGate", func(t *testing.T) {
		inventoryRaw, err := os.ReadFile(filepath.Join("..", "..", "generated", "testspec", "rust_api_surface_inventory.json"))
		if err != nil {
			t.Fatalf("read rust_api_surface_inventory.json: %v", err)
		}
		var inventory struct {
			ConstantsEmitted int `json:"constants_emitted"`
			Constants        []struct {
				Status string `json:"status"`
			} `json:"constants"`
		}
		if err := json.Unmarshal(inventoryRaw, &inventory); err != nil {
			t.Fatalf("unmarshal rust_api_surface_inventory.json: %v", err)
		}
		emittedConstants := 0
		for _, row := range inventory.Constants {
			if row.Status == "emitted" {
				emittedConstants++
			}
		}
		if emittedConstants != inventory.ConstantsEmitted {
			t.Errorf("rust_api_surface_inventory.json constants_emitted field = %d, recounted emitted constants rows = %d", inventory.ConstantsEmitted, emittedConstants)
		}
		if inventory.ConstantsEmitted != anchors.RustPublicAPIConstantsTotal {
			t.Errorf("rust API surface inventory constants_emitted = %d, anchor rust_public_api_constants_total = %d", inventory.ConstantsEmitted, anchors.RustPublicAPIConstantsTotal)
		}

		byWidth := loadRustConstParityRunnerConstants(t)
		sum := 0
		for _, n := range byWidth {
			sum += n
		}
		if sum != anchors.RustPublicAPIConstantsTotal {
			t.Errorf("generated rust runner EXPECTED_CONST_PARITY_CASES_* sum = %d, anchor rust_public_api_constants_total = %d", sum, anchors.RustPublicAPIConstantsTotal)
		}
	})
}

// loadPublicParityRunnerConstants extracts the two generation-time count
// constants the parity runner pins in bid754-go/generated_public_parity_cases_test.go.
// The file is a _test.go in a different module, so its constants cannot be
// imported; they are read from source by AST so the external anchor is compared
// against the exact literals the generator emitted.
func loadPublicParityRunnerConstants(t *testing.T) (wrappers, cases int) {
	t.Helper()
	path := filepath.Join("..", "..", "..", "bid754-go", "generated_public_parity_cases_test.go")
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	found := map[string]int{}
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.CONST {
			continue
		}
		for _, spec := range gen.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for i, name := range vs.Names {
				if name.Name != "expectedPublicParityWrappers" && name.Name != "expectedPublicParityCases" {
					continue
				}
				if i >= len(vs.Values) {
					t.Fatalf("const %s in %s has no value", name.Name, path)
				}
				lit, ok := vs.Values[i].(*ast.BasicLit)
				if !ok || lit.Kind != token.INT {
					t.Fatalf("const %s in %s is not an integer literal", name.Name, path)
				}
				v, err := strconv.Atoi(lit.Value)
				if err != nil {
					t.Fatalf("const %s in %s = %q: %v", name.Name, path, lit.Value, err)
				}
				found[name.Name] = v
			}
		}
	}
	w, okW := found["expectedPublicParityWrappers"]
	c, okC := found["expectedPublicParityCases"]
	if !okW || !okC {
		t.Fatalf("expected both parity constants in %s, found %v", path, found)
	}
	return w, c
}

// loadPublicParityFlaglessConstants extracts the generation-time count
// constants of the flagless-sibling equivalence leg from the generated parity
// cases file, mirroring loadPublicParityRunnerConstants.
func loadPublicParityFlaglessConstants(t *testing.T) (targets, cases int) {
	t.Helper()
	path := filepath.Join("..", "..", "..", "bid754-go", "generated_public_parity_cases_test.go")
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	found := map[string]int{}
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.CONST {
			continue
		}
		for _, spec := range gen.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for i, name := range vs.Names {
				if name.Name != "expectedPublicParityFlaglessSiblingTargets" && name.Name != "expectedPublicParityFlaglessSiblingCases" {
					continue
				}
				if i >= len(vs.Values) {
					t.Fatalf("const %s in %s has no value", name.Name, path)
				}
				lit, ok := vs.Values[i].(*ast.BasicLit)
				if !ok || lit.Kind != token.INT {
					t.Fatalf("const %s in %s is not an integer literal", name.Name, path)
				}
				v, err := strconv.Atoi(lit.Value)
				if err != nil {
					t.Fatalf("const %s in %s = %q: %v", name.Name, path, lit.Value, err)
				}
				found[name.Name] = v
			}
		}
	}
	targetsV, okT := found["expectedPublicParityFlaglessSiblingTargets"]
	casesV, okC := found["expectedPublicParityFlaglessSiblingCases"]
	if !okT || !okC {
		t.Fatalf("expected both flagless-sibling constants in %s, found %v", path, found)
	}
	return targetsV, casesV
}

var (
	rustParityWrappersConstRe = regexp.MustCompile(`(?m)^pub\(crate\) const EXPECTED_PARITY_WRAPPERS: usize = (\d+);$`)
	rustParityCasesConstRe    = regexp.MustCompile(`(?m)^pub\(crate\) const EXPECTED_PARITY_CASES: usize = (\d+);$`)
	rustParityByShapeRowRe    = regexp.MustCompile(`(?m)^\s*\("([a-z0-9_]+)",\s*(\d+)\),$`)
)

// loadRustParityRunnerConstants extracts the generation-time count constants
// bid754-rs/tests/public_parity_generated.rs pins (EXPECTED_PARITY_WRAPPERS,
// EXPECTED_PARITY_CASES, EXPECTED_PARITY_CASES_BY_SHAPE). The file is Rust
// source in a different module/language, so it is read by regex on the fixed
// literal-constant forms this generator always emits, rather than by an AST
// parser (there is no Rust parser in the Go standard toolchain), so the
// external anchor is compared against the exact literals the generator wrote.
func loadRustParityRunnerConstants(t *testing.T) (wrappers, cases int, byShape map[string]int) {
	t.Helper()
	path := filepath.Join("..", "..", "..", "bid754-rs", "tests", "public_parity_generated.rs")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	src := string(data)

	wm := rustParityWrappersConstRe.FindStringSubmatch(src)
	if wm == nil {
		t.Fatalf("EXPECTED_PARITY_WRAPPERS constant not found in %s", path)
	}
	wrappers, err = strconv.Atoi(wm[1])
	if err != nil {
		t.Fatalf("EXPECTED_PARITY_WRAPPERS in %s: %v", path, err)
	}

	cm := rustParityCasesConstRe.FindStringSubmatch(src)
	if cm == nil {
		t.Fatalf("EXPECTED_PARITY_CASES constant not found in %s", path)
	}
	cases, err = strconv.Atoi(cm[1])
	if err != nil {
		t.Fatalf("EXPECTED_PARITY_CASES in %s: %v", path, err)
	}

	tableStart := strings.Index(src, "const EXPECTED_PARITY_CASES_BY_SHAPE")
	if tableStart < 0 {
		t.Fatalf("EXPECTED_PARITY_CASES_BY_SHAPE table not found in %s", path)
	}
	tableEnd := strings.Index(src[tableStart:], "];")
	if tableEnd < 0 {
		t.Fatalf("EXPECTED_PARITY_CASES_BY_SHAPE table in %s has no closing \"];\"", path)
	}
	tableSrc := src[tableStart : tableStart+tableEnd]
	rows := rustParityByShapeRowRe.FindAllStringSubmatch(tableSrc, -1)
	if len(rows) == 0 {
		t.Fatalf("EXPECTED_PARITY_CASES_BY_SHAPE table in %s has no parseable rows", path)
	}
	byShape = make(map[string]int, len(rows))
	for _, row := range rows {
		n, err := strconv.Atoi(row[2])
		if err != nil {
			t.Fatalf("EXPECTED_PARITY_CASES_BY_SHAPE row %q in %s: %v", row[0], path, err)
		}
		if _, dup := byShape[row[1]]; dup {
			t.Fatalf("EXPECTED_PARITY_CASES_BY_SHAPE in %s duplicates shape %q", path, row[1])
		}
		byShape[row[1]] = n
	}
	return wrappers, cases, byShape
}

var (
	rustFlaglessTargetsConstRe = regexp.MustCompile(`(?m)^const EXPECTED_FLAGLESS_SIBLING_TARGETS: usize = (\d+);$`)
	rustFlaglessCasesConstRe   = regexp.MustCompile(`(?m)^const EXPECTED_FLAGLESS_SIBLING_CASES: usize = (\d+);$`)
)

// loadRustParityFlaglessConstants extracts the generation-time count constants
// of the Rust flagless-sibling equivalence leg
// (rust_public_parity_flagless_emit.go) from the generated runner, mirroring
// loadRustParityRunnerConstants's fixed-literal-form regex approach (there is
// no Rust parser in the Go standard toolchain).
func loadRustParityFlaglessConstants(t *testing.T) (targets, cases int) {
	t.Helper()
	path := filepath.Join("..", "..", "..", "bid754-rs", "tests", "public_parity_generated.rs")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	src := string(data)

	tm := rustFlaglessTargetsConstRe.FindStringSubmatch(src)
	if tm == nil {
		t.Fatalf("EXPECTED_FLAGLESS_SIBLING_TARGETS constant not found in %s", path)
	}
	targets, err = strconv.Atoi(tm[1])
	if err != nil {
		t.Fatalf("EXPECTED_FLAGLESS_SIBLING_TARGETS in %s: %v", path, err)
	}

	cm := rustFlaglessCasesConstRe.FindStringSubmatch(src)
	if cm == nil {
		t.Fatalf("EXPECTED_FLAGLESS_SIBLING_CASES constant not found in %s", path)
	}
	cases, err = strconv.Atoi(cm[1])
	if err != nil {
		t.Fatalf("EXPECTED_FLAGLESS_SIBLING_CASES in %s: %v", path, err)
	}
	return targets, cases
}

// rustConstParityCasesConstRe extracts a per-width EXPECTED_CONST_PARITY_CASES_
// self-check constant (e.g. "const EXPECTED_CONST_PARITY_CASES_64: usize = 4;") from
// bid754-rs/tests/public_parity_generated.rs, mirroring rustParityCasesConstRe's
// fixed-literal-form regex approach (no Rust parser in the Go standard toolchain).
var rustConstParityCasesConstRe = regexp.MustCompile(`(?m)^const EXPECTED_CONST_PARITY_CASES_(\d+): usize = (\d+);$`)

// loadRustConstParityRunnerConstants extracts every EXPECTED_CONST_PARITY_CASES_<digits>
// constant the constants-parity generator pins (rust_public_parity_emit.go's
// emitRustConstParityTest), keyed by width digits ("32"/"64"/"128"), so
// TestVerificationAnchorsMatchGeneratedArtifacts's RustPublicAPIConstantsGate can sum
// them against the external rust_public_api_constants_total anchor.
func loadRustConstParityRunnerConstants(t *testing.T) map[string]int {
	t.Helper()
	path := filepath.Join("..", "..", "..", "bid754-rs", "tests", "public_parity_generated.rs")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	matches := rustConstParityCasesConstRe.FindAllStringSubmatch(string(data), -1)
	if len(matches) == 0 {
		t.Fatalf("EXPECTED_CONST_PARITY_CASES_* constants not found in %s", path)
	}
	byWidth := make(map[string]int, len(matches))
	for _, m := range matches {
		n, err := strconv.Atoi(m[2])
		if err != nil {
			t.Fatalf("EXPECTED_CONST_PARITY_CASES_%s in %s: %v", m[1], path, err)
		}
		if _, dup := byWidth[m[1]]; dup {
			t.Fatalf("EXPECTED_CONST_PARITY_CASES_%s appears twice in %s", m[1], path)
		}
		byWidth[m[1]] = n
	}
	return byWidth
}

// generatedMarkerRegexp mirrors the marker line that
// devtools/scripts/check_generated_marker_coverage.sh discovers generated
// artifacts with. TestVerificationArtifactContentHashes reuses the same
// Git-visible marker set as the exhaustive universe it must classify.
const generatedMarkerRegexp = `^(//|#) Code generated .* DO NOT EDIT\.$`

// verificationArtifactDigestListingCap bounds how many per-file digest lines a
// content-hash mismatch prints inline. The tree hash is the actual check; the
// per-file listing is only a diagnostic aid for a committed regression that git
// status cannot point at, so large buckets (the ~1000-file testspec tree) print
// the changed-files hint and a pointer instead of a wall of digests.
const verificationArtifactDigestListingCap = 40

// implementationExclusionRules name the generated-marker artifacts that are
// deliberately NOT content-hashed. These are implementation-path outputs (the
// generated Go/Rust value types, the DFP tables, and the full go2rs Rust
// implementation tree): a content substitution in them is already caught by the
// readtest / decTest / FFI behavior domains and devtools/internal/tablecrosscheck,
// so pinning their bytes here would duplicate that coverage and churn on every
// legitimate implementation regeneration. Each rule must still match at least
// one universe file, so a rule that no longer matches anything fails as a stale
// exclusion rather than silently dropping an artifact out of the exhaustive set.
var implementationExclusionRules = []struct {
	name  string
	match func(rel string) bool
}{
	{"bid754-go/generated_types.go", func(r string) bool { return r == "bid754-go/generated_types.go" }},
	{"bid754-go/internal/bidgo/tables_binarydecimal.go", func(r string) bool {
		return r == "bid754-go/internal/bidgo/tables_binarydecimal.go"
	}},
	{"devtools/generated/go/intel_dfp_tables.go", func(r string) bool { return r == "devtools/generated/go/intel_dfp_tables.go" }},
	{"bid754-rs/src/gen_types.rs", func(r string) bool { return r == "bid754-rs/src/gen_types.rs" }},
	{"bid754-rs/src/gen_constants.rs", func(r string) bool { return r == "bid754-rs/src/gen_constants.rs" }},
	{"bid754-rs/src/tables.rs", func(r string) bool { return r == "bid754-rs/src/tables.rs" }},
	{"bid754-rs/src/intel_dfp_tables.rs", func(r string) bool { return r == "bid754-rs/src/intel_dfp_tables.rs" }},
	{"bid754-rs/src/generated/", func(r string) bool { return strings.HasPrefix(r, "bid754-rs/src/generated/") }},
	// The crate root is now a go2rs apiemit output (it re-exports the generated
	// public API and keeps the internal modules as doc(hidden) compat surface).
	// Like the generated implementation tree it is routing/plumbing over the
	// generated port, not a verification runner: its reproducibility is covered
	// by verify-generated and its behavior by the cargo test domains, so it is
	// excluded from content hashing rather than pinned here.
	{"bid754-rs/src/lib.rs", func(r string) bool { return r == "bid754-rs/src/lib.rs" }},
}

// classifyVerificationArtifact maps a repo-relative, slash-separated path drawn
// from the generated-marker universe to its content-hash bucket. It returns
// (bucket, "") for a hashed verification artifact, ("", ruleName) for an
// implementation-path artifact excluded from hashing, and ("", "") for a path
// that matches nothing (an unclassified generated artifact that must fail the
// exhaustive check). The testspec tree and the codec vectors file carry no
// marker line and are added to the hashed set separately by the test.
func classifyVerificationArtifact(rel string) (bucket, exclusionRule string) {
	switch rel {
	case "bid754-codec-go/vector_test.go",
		"bid754-codec-go/testdata/external_vector_test.go",
		"bid754-codec-go/exhaustive32_long_test.go",
		"bid754-codec-go/decimal64_128_long_test.go",
		"bid754-codec-rs/tests/vectors.rs",
		"bid754-rs/tests/bid_codec_vectors.rs",
		"bid754-rs/tests/bid_codec_parse_vectors.rs",
		"bid754-go/generated_bid_codec_vectors_test.go",
		"bid754-rs/tests/bid_string_vectors.rs",
		"bid754-codec-java/src/test/java/io/github/sky1core/bidcodec/VectorRunner.java",
		"bid754-codec-java/src/test/java/io/github/sky1core/bidcodec/VectorTest.java",
		"bid754-codec-py/tests/test_vectors.py",
		"bid754-codec-js/src/vectors.test.ts",
		"bid754-codec-js/vector_runner.mjs",
		"bid754-codec-swift/Sources/BidCodecVectorRunner/main.swift":
		return "bidcodec_harnesses", ""
	case "bid754-go/internal/testspec/spec.go",
		"bid754-go/internal/testspec/spec_io.go",
		"bid754-go/internal/testspec/spec_io_strict_test.go":
		return "spec_loader", ""
	case "bid754-rs/ffi-verify/tests/readtest_generated.rs":
		return "rust_readtest_runner", ""
	case "bid754-rs/ffi-verify/tests/tier1_arithmetic_long_generated.rs",
		"bid754-rs/ffi-verify/tests/tier1_compare_conversion_long_generated.rs":
		return "rust_tier1_long_runners", ""
	case "bid754-rs/ffi-verify/tests/d32_exhaustive_long_generated.rs":
		// The generated Rust leg of the Decimal32 unary exhaustive gate hashes
		// with the Go-side d32 exhaustive artifacts: one gate, one bucket.
		return "d32_exhaustive_runners", ""
	case "bid754-rs/tests/public_parity_generated.rs":
		return "rust_public_parity_runner", ""
	case "bid754-rs/tests/dectest_generated.rs":
		return "rust_dectest_runner", ""
	case "bid754-go/internal/bidgo/string_vectors_test.go":
		return "goport_verification_runners", ""
	}
	const goRoot = "bid754-go/"
	if strings.HasPrefix(rel, goRoot) {
		if base := rel[len(goRoot):]; !strings.Contains(base, "/") {
			switch {
			case strings.HasPrefix(base, "generated_dectest_goport_"):
				// The portable Go mechanical-port decTest cross-check runners are
				// goport verification runners, not decNumber-oracle executor files;
				// this case must precede the generic generated_dectest_ arm.
				return "goport_verification_runners", ""
			case strings.HasPrefix(base, "generated_decnumber_differential_"):
				// The decNumber third-oracle differential gate artifacts (shared
				// support, cgo shim, runner, stub) form their own hashed bucket;
				// this case must precede the generic generated_dectest_ arm's
				// sibling prefixes for clarity even though the prefixes differ.
				return "decnumber_differential_runners", ""
			case strings.HasPrefix(base, "generated_d32_exhaustive_"):
				// The Decimal32 unary exhaustive gate artifacts (cgo shim,
				// runner, stub) form their own hashed bucket.
				return "d32_exhaustive_runners", ""
			case strings.HasPrefix(base, "generated_readtest_"),
				strings.HasPrefix(base, "generated_ffi_bitcompare_"),
				strings.HasPrefix(base, "generated_public_parity_"):
				return "goport_verification_runners", ""
			case strings.HasPrefix(base, "generated_dectest_"),
				strings.HasPrefix(base, "dectest_") && strings.HasSuffix(base, ".go"):
				return "dectest_executor", ""
			}
		}
	}
	for _, rule := range implementationExclusionRules {
		if rule.match(rel) {
			return "", rule.name
		}
	}
	return "", ""
}

// TestVerificationArtifactContentHashes promotes the anchor scheme from
// count-only pins to count+content pins. verify-generated proves "checked-in ==
// regenerated" (reproducibility), but it cannot see a self-consistent generator
// regression: if the generator and its own hardcoded pins shrink or relax
// together, the regenerated artifact still matches the checked-in one. This test
// pins the CONTENT of the verification-participating generated artifacts against
// hashes held outside every generation path (verification_anchors.json), so a
// count-preserving content substitution (swapping hard cases for easy ones, or
// relaxing a runner template's assertions) fails here instead of passing every
// gate. It also enforces an exhaustive set: every generated-marker artifact must be
// either content-hashed or explicitly excluded as an implementation-path output.
func TestVerificationArtifactContentHashes(t *testing.T) {
	anchors := loadVerificationAnchors(t)
	if len(anchors.VerificationArtifactSHA256) == 0 {
		t.Fatalf("verification_anchors.json is missing the verification_artifact_sha256 content-hash anchors")
	}
	repoRoot := verificationRepoRoot(t)

	universe := markerArtifactUniverse(t, repoRoot)
	changed := gitStatusPaths(repoRoot)

	bucketFiles := map[string][]string{}
	matchedExclusions := map[string]bool{}
	var unclassified []string
	for _, rel := range universe {
		bucket, exclusion := classifyVerificationArtifact(rel)
		switch {
		case bucket != "":
			bucketFiles[bucket] = append(bucketFiles[bucket], rel)
		case exclusion != "":
			matchedExclusions[exclusion] = true
		default:
			unclassified = append(unclassified, rel)
		}
	}
	if len(unclassified) > 0 {
		t.Errorf("incomplete mapping: %d generated-marker artifact(s) are neither content-hashed nor excluded.\n"+
			"Classify each into a verification_artifact_sha256 bucket, or add an implementation-path exclusion in classifyVerificationArtifact:\n  %s",
			len(unclassified), strings.Join(unclassified, "\n  "))
	}
	for _, rule := range implementationExclusionRules {
		if !matchedExclusions[rule.name] {
			t.Errorf("stale exclusion: implementationExclusionRules entry %q matched no generated-marker artifact; remove it or fix the rule", rule.name)
		}
	}

	// The shared verification spec tree and the BID codec vectors file are JSON
	// artifacts with no generated-code marker line, so they join the hashed set
	// explicitly rather than through the marker universe. The generated spec
	// tree includes both tracked files and untracked non-ignored candidates: a
	// newly generated shard must move the external hash before it is staged,
	// rather than disappearing from this gate until Git happens to track it.
	// The codec vectors file is a single fixed tracked path.
	bucketFiles["testspec_tree"] = trackedAndUntrackedNonIgnoredFilesUnder(t, repoRoot, "devtools/generated/testspec")
	bucketFiles["bidcodec_vectors"] = []string{"bid754-codec-vectors/vectors.json"}

	unusedAnchors := map[string]bool{}
	for key := range anchors.VerificationArtifactSHA256 {
		unusedAnchors[key] = true
	}
	for bucket, files := range bucketFiles {
		got, digests := verificationTreeHash(t, repoRoot, files)
		t.Logf("verification_artifact_sha256.%s = %s  (%d files)", bucket, got, len(files))
		want, ok := anchors.VerificationArtifactSHA256[bucket]
		if !ok {
			t.Errorf("content-hash bucket %q has no verification_artifact_sha256 anchor entry", bucket)
			continue
		}
		delete(unusedAnchors, bucket)
		if got == want {
			continue
		}
		t.Errorf("content-hash mismatch for verification domain %q:\n"+
			"  computed sha256 = %s\n"+
			"  anchor   sha256 = %s\n"+
			"%s"+
			"  If this content change is intended, set verification_anchors.json verification_artifact_sha256.%s to the computed sha256 and record the input change that moved it.\n"+
			"%s",
			bucket, got, want,
			changedFilesHint(files, changed),
			bucket,
			digestDiagnostic(bucket, files, digests))
	}
	for key := range unusedAnchors {
		t.Errorf("verification_artifact_sha256 anchor %q matches no content-hash bucket (stale anchor key)", key)
	}
}

func verificationRepoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	return root
}

// markerArtifactUniverse returns the tracked and untracked non-ignored
// generated-marker files minus the documented marker exceptions, i.e. exactly
// the marker-bearing files that verify-generated is expected to compare. git is
// required: the marker set is the independent exhaustive universe, and silently
// skipping it would disable the check.
func markerArtifactUniverse(t *testing.T, repoRoot string) []string {
	t.Helper()
	cmd := exec.Command("git", "-C", repoRoot, "grep", "--untracked", "-lIE", generatedMarkerRegexp, "--", ".")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git grep for generated markers failed (git is required to enumerate the exhaustive universe): %v", err)
	}
	marks := splitNonEmptyLines(string(out))
	if len(marks) == 0 {
		t.Fatalf("git grep found no generated-marker files; the marker regexp %q no longer matches the tree", generatedMarkerRegexp)
	}
	allow := readMarkerExceptions(t, repoRoot)
	var universe []string
	for _, m := range marks {
		if allow[m] {
			continue
		}
		universe = append(universe, m)
	}
	sort.Strings(universe)
	return universe
}

func TestMarkerArtifactUniverseIncludesUntrackedNonIgnoredFiles(t *testing.T) {
	repoRoot := t.TempDir()
	runGit := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", repoRoot}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
		}
	}
	runGit("init", "-q")
	for relative, data := range map[string]string{
		".gitignore": "ignored.go\n",
		"devtools/scripts/generated_marker_exceptions.txt": "",
		"tracked.go":   "// Code generated by test; DO NOT EDIT.\n",
		"untracked.go": "// Code generated by test; DO NOT EDIT.\n",
		"ignored.go":   "// Code generated by test; DO NOT EDIT.\n",
	} {
		fullPath := filepath.Join(repoRoot, filepath.FromSlash(relative))
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(fullPath, []byte(data), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	runGit("add", "--", ".gitignore", "tracked.go")

	got := markerArtifactUniverse(t, repoRoot)
	want := []string{"tracked.go", "untracked.go"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Git-visible generated-marker files = %v, want %v", got, want)
	}
}

// readMarkerExceptions parses devtools/scripts/generated_marker_exceptions.txt the
// same way check_generated_marker_coverage.sh does, so a file that is
// intentionally excluded from verify-generated comparison is also excluded from
// the content-hash universe.
func readMarkerExceptions(t *testing.T, repoRoot string) map[string]bool {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(repoRoot, "devtools", "scripts", "generated_marker_exceptions.txt"))
	if err != nil {
		t.Fatalf("read generated_marker_exceptions.txt: %v", err)
	}
	allow := map[string]bool{}
	for _, line := range strings.Split(string(raw), "\n") {
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = line[:i]
		}
		if line = strings.TrimSpace(line); line != "" {
			allow[line] = true
		}
	}
	return allow
}

// trackedAndUntrackedNonIgnoredFilesUnder returns the Git-visible candidate
// files under relDir as sorted repo-relative slash paths. Including cached and
// untracked non-ignored files closes the pre-staging gap: a newly generated
// shard participates in the hash as soon as it appears in the working tree.
func trackedAndUntrackedNonIgnoredFilesUnder(t *testing.T, repoRoot, relDir string) []string {
	t.Helper()
	out, err := exec.Command("git", "-C", repoRoot, "ls-files", "-z", "--cached", "--others", "--exclude-standard", "--", relDir).Output()
	if err != nil {
		t.Fatalf("git ls-files under %s failed (git is required to enumerate the generated spec tree): %v", relDir, err)
	}
	var files []string
	// git ls-files -z emits each path verbatim terminated by NUL, so split on
	// NUL and keep non-empty entries (the trailing NUL yields one empty tail).
	// No trimming: a path is used byte-for-byte as git tracks it.
	for _, f := range strings.Split(string(out), "\x00") {
		if f != "" {
			files = append(files, f)
		}
	}
	if len(files) == 0 {
		t.Fatalf("no tracked or untracked non-ignored files under %s; the verification spec tree is empty", relDir)
	}
	sort.Strings(files)
	return files
}

func TestTrackedAndUntrackedNonIgnoredFilesUnder(t *testing.T) {
	repoRoot := t.TempDir()
	runGit := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", repoRoot}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
		}
	}
	runGit("init", "-q")
	if err := os.MkdirAll(filepath.Join(repoRoot, "tree"), 0o755); err != nil {
		t.Fatal(err)
	}
	for rel, data := range map[string]string{
		".gitignore":          "tree/ignored.json\n",
		"tree/tracked.json":   "tracked\n",
		"tree/untracked.json": "untracked\n",
		"tree/ignored.json":   "ignored\n",
	} {
		if err := os.WriteFile(filepath.Join(repoRoot, filepath.FromSlash(rel)), []byte(data), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	runGit("add", "--", ".gitignore", "tree/tracked.json")

	got := trackedAndUntrackedNonIgnoredFilesUnder(t, repoRoot, "tree")
	want := []string{"tree/tracked.json", "tree/untracked.json"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Git-visible generated files = %v, want %v", got, want)
	}
}

// verificationTreeHash computes the bucket content hash as the sha256 of the
// sorted "<repo-relative-path>  <file sha256>" lines, and returns those lines
// for diagnostics.
func verificationTreeHash(t *testing.T, repoRoot string, relPaths []string) (string, []string) {
	t.Helper()
	sorted := append([]string(nil), relPaths...)
	sort.Strings(sorted)
	var manifest strings.Builder
	digests := make([]string, 0, len(sorted))
	for _, rel := range sorted {
		data, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(rel)))
		if err != nil {
			t.Fatalf("read hashed artifact %s: %v", rel, err)
		}
		sum := sha256.Sum256(data)
		line := fmt.Sprintf("%s  %x", rel, sum)
		manifest.WriteString(line)
		manifest.WriteByte('\n')
		digests = append(digests, line)
	}
	top := sha256.Sum256([]byte(manifest.String()))
	return hex.EncodeToString(top[:]), digests
}

// gitStatusPaths returns the working-tree paths git reports as changed, used
// only as a best-effort hint to point at the file behind a mismatch. It returns
// nil if git is unavailable; the tree hash remains the real check.
func gitStatusPaths(repoRoot string) map[string]bool {
	out, err := exec.Command("git", "-C", repoRoot, "status", "--porcelain").Output()
	if err != nil {
		return nil
	}
	changed := map[string]bool{}
	for _, line := range strings.Split(string(out), "\n") {
		if len(line) < 4 {
			continue
		}
		p := strings.TrimSpace(line[3:])
		if i := strings.Index(p, " -> "); i >= 0 {
			p = p[i+len(" -> "):]
		}
		changed[strings.Trim(p, "\"")] = true
	}
	return changed
}

func changedFilesHint(files []string, changed map[string]bool) string {
	if changed == nil {
		return "  (git status unavailable; inspect this domain's files below)\n"
	}
	var hits []string
	for _, f := range files {
		if changed[f] {
			hits = append(hits, f)
		}
	}
	if len(hits) == 0 {
		return "  No uncommitted change among this domain's files: the divergence is in committed content — review the digest below.\n"
	}
	return "  Locally modified file(s) in this domain (candidate changed artifacts):\n    " + strings.Join(hits, "\n    ") + "\n"
}

func digestDiagnostic(bucket string, files, digests []string) string {
	if len(files) > verificationArtifactDigestListingCap {
		return fmt.Sprintf("  (%d files in %q; per-file digest omitted — run devtools/scripts/compute_verification_artifact_hashes.sh or git diff to locate the change)\n",
			len(files), bucket)
	}
	var b strings.Builder
	fmt.Fprintf(&b, "  Per-file digest for domain %q (%d files):\n", bucket, len(files))
	for _, d := range digests {
		b.WriteString("    ")
		b.WriteString(d)
		b.WriteByte('\n')
	}
	return b.String()
}

func splitNonEmptyLines(s string) []string {
	var out []string
	for _, line := range strings.Split(s, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			out = append(out, line)
		}
	}
	return out
}

// decnumberDifferentialArtifactInventory carries the anchored constants
// parsed from the checked-in generated decNumber differential shared
// support file (bid754-go/generated_decnumber_differential_shared_test.go).
type decnumberDifferentialArtifactInventory struct {
	BoundaryValues             map[string]uint64
	ProbeValues                map[string]uint64
	ExactProductValues         map[string]uint64
	ExactProductAddends        map[string]uint64
	RandomPairsPerOp           uint64
	StructuredComparisons      map[string]uint64
	StructuredFmaExcluded      map[string]uint64
	StructuredKnownDivergences map[string]uint64
	StructuredStreamHashes     map[string]uint64
	RandomComparisons          map[string]uint64
	RandomFmaExcluded          map[string]uint64
	RandomKnownDivergences     map[string]uint64
	RandomStreamHashes         map[string]uint64
	TotalComparisons           map[string]uint64
	SentinelRowCount           uint64
}

// loadDecnumberDifferentialArtifactInventory evaluates every
// `decnumberDiff*` integer constant of the checked-in shared support file.
// Constants whose expressions fall outside the evaluator's operator set are
// skipped; every constant this inventory needs must resolve.
func loadDecnumberDifferentialArtifactInventory(t *testing.T) decnumberDifferentialArtifactInventory {
	t.Helper()
	path := filepath.Join("..", "..", "..", "bid754-go", "generated_decnumber_differential_shared_test.go")
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		t.Fatalf("parse generated decNumber differential shared support: %v", err)
	}
	constants := map[string]uint64{}
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.CONST {
			continue
		}
		for _, spec := range gen.Specs {
			values, ok := spec.(*ast.ValueSpec)
			if !ok || len(values.Names) != len(values.Values) {
				continue
			}
			for i, name := range values.Names {
				if !strings.HasPrefix(name.Name, "decnumberDiff") {
					continue
				}
				value, err := evalGeneratedUintConstant(values.Values[i], constants)
				if err != nil {
					continue // non-arithmetic support constant; not anchored here
				}
				constants[name.Name] = value
			}
		}
	}
	require := func(name string) uint64 {
		t.Helper()
		value, ok := constants[name]
		if !ok {
			t.Fatalf("generated decNumber differential shared support is missing constant %s", name)
		}
		return value
	}
	widthMap := func(prefix string) map[string]uint64 {
		return map[string]uint64{
			"decimal32":  require(prefix + "32"),
			"decimal64":  require(prefix + "64"),
			"decimal128": require(prefix + "128"),
		}
	}
	return decnumberDifferentialArtifactInventory{
		BoundaryValues:             map[string]uint64{"decimal32": require("decnumberDiffBoundary32Count"), "decimal64": require("decnumberDiffBoundary64Count"), "decimal128": require("decnumberDiffBoundary128Count")},
		ProbeValues:                map[string]uint64{"decimal32": require("decnumberDiffProbes32Count"), "decimal64": require("decnumberDiffProbes64Count"), "decimal128": require("decnumberDiffProbes128Count")},
		ExactProductValues:         map[string]uint64{"decimal32": require("decnumberDiffExactProduct32Count"), "decimal64": require("decnumberDiffExactProduct64Count"), "decimal128": require("decnumberDiffExactProduct128Count")},
		ExactProductAddends:        map[string]uint64{"decimal32": require("decnumberDiffExactProductZ32Count"), "decimal64": require("decnumberDiffExactProductZ64Count"), "decimal128": require("decnumberDiffExactProductZ128Count")},
		RandomPairsPerOp:           require("decnumberDiffRandomPairsPerOp"),
		StructuredComparisons:      widthMap("decnumberDiffStructuredComparisons"),
		StructuredFmaExcluded:      widthMap("decnumberDiffStructuredFmaExcluded"),
		StructuredKnownDivergences: widthMap("decnumberDiffStructuredKnownDivergences"),
		StructuredStreamHashes:     widthMap("decnumberDiffStructuredStreamHash"),
		RandomComparisons:          widthMap("decnumberDiffRandomComparisons"),
		RandomFmaExcluded:          widthMap("decnumberDiffRandomFmaExcluded"),
		RandomKnownDivergences:     widthMap("decnumberDiffRandomKnownDivergences"),
		RandomStreamHashes:         widthMap("decnumberDiffRandomStreamHash"),
		TotalComparisons:           widthMap("decnumberDiffTotalComparisons"),
		SentinelRowCount:           require("decnumberDiffSentinelRowCount"),
	}
}

// d32ExhaustiveArtifactInventory carries the anchored constants parsed from
// the checked-in generated d32 exhaustive runner
// (bid754-go/generated_d32_exhaustive_long_test.go).
type d32ExhaustiveArtifactInventory struct {
	Lanes            uint64
	Operations       uint64
	CasesPerLane     uint64
	TotalComparisons uint64
	SentinelRowCount uint64
}

// loadD32ExhaustiveArtifactInventory evaluates every `d32Exhaustive*`
// integer constant of the checked-in generated runner. Constants whose
// expressions fall outside the evaluator's operator set are skipped; every
// constant this inventory needs must resolve.
func loadD32ExhaustiveArtifactInventory(t *testing.T) d32ExhaustiveArtifactInventory {
	t.Helper()
	path := filepath.Join("..", "..", "..", "bid754-go", "generated_d32_exhaustive_long_test.go")
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		t.Fatalf("parse generated d32 exhaustive runner: %v", err)
	}
	constants := map[string]uint64{}
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.CONST {
			continue
		}
		for _, spec := range gen.Specs {
			values, ok := spec.(*ast.ValueSpec)
			if !ok || len(values.Names) != len(values.Values) {
				continue
			}
			for i, name := range values.Names {
				if !strings.HasPrefix(name.Name, "d32Exhaustive") {
					continue
				}
				value, err := evalGeneratedUintConstant(values.Values[i], constants)
				if err != nil {
					continue // non-arithmetic support constant; not anchored here
				}
				constants[name.Name] = value
			}
		}
	}
	require := func(name string) uint64 {
		t.Helper()
		value, ok := constants[name]
		if !ok {
			t.Fatalf("generated d32 exhaustive runner is missing constant %s", name)
		}
		return value
	}
	return d32ExhaustiveArtifactInventory{
		Lanes:            require("d32ExhaustiveLaneCount"),
		Operations:       require("d32ExhaustiveOperationCount"),
		CasesPerLane:     require("d32ExhaustiveCasesPerLane"),
		TotalComparisons: require("d32ExhaustiveTotalComparisons"),
		SentinelRowCount: require("d32ExhaustiveSentinelRowCount"),
	}
}

// loadDecnumberDifferentialInventoryJSON reads the generated closed-world
// exclusion inventory as the third agreement point (anchors == shared
// constants == inventory JSON).
func loadDecnumberDifferentialInventoryJSON(t *testing.T) decnumberDiffInventory {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "generated", "testspec", "decnumber_differential_inventory.json"))
	if err != nil {
		t.Fatalf("read decNumber differential inventory: %v", err)
	}
	var inventory decnumberDiffInventory
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&inventory); err != nil {
		t.Fatalf("unmarshal decNumber differential inventory: %v", err)
	}
	return inventory
}
