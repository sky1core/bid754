package testgen

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/sky1core/bid754/devtools/internal/genmarker"
)

func TestBidCodecVectorGeneratorDoesNotImportBidCodecUnderTest(t *testing.T) {
	files, err := filepath.Glob("bid_codec*.go")
	if err != nil {
		t.Fatalf("Glob bid_codec*.go: %v", err)
	}
	for _, path := range files {
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("ParseFile(%q): %v", path, err)
		}
		for _, imp := range file.Imports {
			if imp.Path.Value == `"github.com/sky1core/bid754/bid754-codec-go"` {
				t.Fatalf("%s imports github.com/sky1core/bid754/bid754-codec-go; BID codec vectors must use the independent reference codec, not the package under test", path)
			}
		}
	}
}

func TestBidCodecToStringRejectVectorsCloseSharedSchema(t *testing.T) {
	vectors := bidCodecToStringRejectVectors()
	if len(vectors) != 17 {
		t.Fatalf("to_string reject vector count = %d, want 17", len(vectors))
	}
	want := map[string]bool{
		"normal||0||missing_coefficient_not_representable_for_normal|":                  true,
		"normal|0|0||zero_coefficient_not_representable_for_normal|":                    true,
		"normal|10000000000000000000000000000000000|0||coefficient_exceeds_schema_max|": true,
		"normal|1|0|1|payload_not_representable_for_normal|":                            true,
		"zero|1|0||coefficient_not_representable_for_zero|":                             true,
		"zero||0|1|payload_not_representable_for_zero|":                                 true,
		"inf|1|0||coefficient_not_representable_for_infinity|":                          true,
		"inf||1||exponent_not_representable_for_infinity|":                              true,
		"inf||0|1|payload_not_representable_for_infinity|":                              true,
		"qnan|1|0||coefficient_not_representable_for_qnan|":                             true,
		"qnan||1||exponent_not_representable_for_qnan|":                                 true,
		"qnan||0|1000000000000000000000000000000000|nan_payload_exceeds_schema_max|":    true,
		"snan|1|0||coefficient_not_representable_for_snan|":                             true,
		"snan||1||exponent_not_representable_for_snan|":                                 true,
		"snan||0|1000000000000000000000000000000000|nan_payload_exceeds_schema_max|":    true,
		"normal|-1|0||negative_coefficient|negative_coefficient":                        true,
		"qnan||0|-1|negative_payload|negative_payload":                                  true,
	}
	for _, v := range vectors {
		if v.Channel != "to_string" || v.Type != "" || v.Input != nil {
			t.Fatalf("to_string reject has channel/type/input = %q/%q/%v", v.Channel, v.Type, v.Input)
		}
		key := fmt.Sprintf("%s|%s|%d|%s|%s|%s", v.Kind, v.Coefficient, v.Exponent, v.Payload, v.Reason, v.Requires)
		if !want[key] {
			t.Fatalf("unexpected to_string reject vector: %#v", v)
		}
		delete(want, key)
	}
	if len(want) != 0 {
		t.Fatalf("missing to_string reject vectors: %v", want)
	}

	wantCounts := map[string][2]int{
		"go": {98, 0}, "java": {98, 0}, "python": {98, 0}, "js": {98, 0},
		"rust": {93, 5}, "rust_full": {93, 5}, "swift": {93, 5},
	}
	for lang, counts := range wantCounts {
		consumed, skipped := bidCodecRejectExpectedCounts(lang)
		if consumed != counts[0] || skipped != counts[1] {
			t.Errorf("%s reject counts = consumed %d, skipped %d; want %d/%d", lang, consumed, skipped, counts[0], counts[1])
		}
	}

	wantTypeDomainIDs := map[string]bool{
		"non_boolean_sign": true, "non_integral_exponent": true, "nan_exponent": true,
		"exponent_above_int32": true, "exponent_below_int32": true,
		"non_integer_coefficient": true, "non_integer_payload": true, "unrecognized_kind": true,
		"boolean_coefficient": true, "boolean_exponent": true, "boolean_payload": true,
	}
	pyCases, jsCases, goCases := 0, 0, 0
	for _, tc := range bidCodecTypeDomainRejectCases {
		if !wantTypeDomainIDs[tc.ID] {
			t.Fatalf("unexpected type-domain reject %q", tc.ID)
		}
		delete(wantTypeDomainIDs, tc.ID)
		if tc.Py != "" {
			pyCases++
		}
		if tc.JS != "" {
			jsCases++
		}
		if tc.Go != "" {
			goCases++
		}
	}
	if len(wantTypeDomainIDs) != 0 || pyCases != 11 || jsCases != 11 || goCases != 1 {
		t.Fatalf("type-domain rejects missing=%v counts py/js/go=%d/%d/%d, want 11/11/1", wantTypeDomainIDs, pyCases, jsCases, goCases)
	}

	wantRawDecodeIDs := map[string]bool{
		"decode32_boolean": true, "decode32_string": true, "decode32_fraction": true,
		"decode32_nan": true, "decode32_infinity": true, "decode32_negative": true,
		"decode32_overflow": true, "decode32_bigint": true,
		"decode64_boolean": true, "decode64_string": true, "decode64_wrong_numeric_type": true,
		"decode64_negative": true, "decode64_overflow": true,
		"decode128_lo_boolean": true, "decode128_lo_string": true,
		"decode128_lo_wrong_numeric_type": true, "decode128_lo_negative": true,
		"decode128_lo_overflow": true, "decode128_hi_boolean": true,
		"decode128_hi_string": true, "decode128_hi_wrong_numeric_type": true,
		"decode128_hi_negative": true, "decode128_hi_overflow": true,
	}
	pyRawCases, jsRawCases := 0, 0
	for _, tc := range bidCodecRawDecodeRejectCases {
		if !wantRawDecodeIDs[tc.ID] {
			t.Fatalf("unexpected raw-decode reject %q", tc.ID)
		}
		delete(wantRawDecodeIDs, tc.ID)
		if tc.Py != "" {
			pyRawCases++
		}
		if tc.JS != "" {
			jsRawCases++
		}
	}
	if len(wantRawDecodeIDs) != 0 || pyRawCases != 22 || jsRawCases != 23 {
		t.Fatalf("raw-decode rejects missing=%v counts py/js=%d/%d, want 22/23", wantRawDecodeIDs, pyRawCases, jsRawCases)
	}
}

func TestDectestExecutorRootFilesAreGenerated(t *testing.T) {
	repoRoot := filepath.Clean(filepath.Join("..", ".."))
	generatedExecutorPaths := make(map[string]bool, len(dectestExecutorTemplates))
	for _, item := range dectestExecutorTemplates {
		generatedExecutorPaths[filepath.Join(repoRoot, item.OutputPath)] = true
	}
	files, err := filepath.Glob(filepath.Join(repoRoot, "..", "bid754-go", "dectest_*.go"))
	if err != nil {
		t.Fatalf("Glob dectest_*.go: %v", err)
	}
	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("ReadFile(%q): %v", path, err)
		}
		hasExecutor := bytes.Contains(data, []byte("func executeDecTest")) ||
			bytes.Contains(data, []byte("func executePortable"))
		hasGeneratedHeader := bytes.HasPrefix(data, []byte(genmarker.Line("testgen")+"\n"))
		isGeneratedExecutorPath := generatedExecutorPaths[path]
		if hasExecutor && !hasGeneratedHeader {
			t.Fatalf("%s contains a decTest executor but is not generated by testgen", path)
		}
		if hasGeneratedHeader && !isGeneratedExecutorPath {
			t.Fatalf("%s has a testgen generated header but is not owned by dectestExecutorTemplates", path)
		}
		if isGeneratedExecutorPath && !hasGeneratedHeader {
			t.Fatalf("%s is listed in dectestExecutorTemplates but lacks the generated header", path)
		}
	}
}

func TestGeneratedSharedSpecStaysInSync(t *testing.T) {
	repoRoot := filepath.Clean(filepath.Join("..", ".."))
	manifest, err := LoadManifest(filepath.Join(repoRoot, "testgen_manifest.json"))
	if err != nil {
		t.Fatalf("LoadManifest error: %v", err)
	}
	requireGenerationInputsForTest(t, repoRoot, manifest)
	spec, err := Generate(repoRoot, manifest)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	specFiles, err := EncodeSpecFiles(manifest, spec)
	if err != nil {
		t.Fatalf("EncodeSpecFiles error: %v", err)
	}
	for relativePath, want := range specFiles {
		assertGeneratedFileMatches(t, repoRoot, filepath.FromSlash(relativePath), want)
	}
	assertNoStaleShardFiles(t, repoRoot, manifest, specFiles)
	readtestTestOutputs, err := GenerateReadtestTestOutputs(spec)
	if err != nil {
		t.Fatalf("GenerateReadtestTestOutputs error: %v", err)
	}
	assertGeneratedFileMatches(t, repoRoot, readtestGeneratedNativeTestPath, readtestTestOutputs[readtestGeneratedNativeTestPath])
	assertGeneratedFileMatches(t, repoRoot, readtestGeneratedStubTestPath, readtestTestOutputs[readtestGeneratedStubTestPath])
	readDispatchOutputs, err := GenerateReadtestDispatchOutputs(repoRoot, manifest)
	if err != nil {
		t.Fatalf("GenerateReadtestDispatchOutputs error: %v", err)
	}
	assertGeneratedFileMatches(t, repoRoot, readtestGeneratedNativePath, readDispatchOutputs[readtestGeneratedNativePath])
	assertGeneratedFileMatches(t, repoRoot, readtestGeneratedStubPath, readDispatchOutputs[readtestGeneratedStubPath])
	assertGeneratedFileMatches(t, repoRoot, readtestGeneratedSharedPath, readDispatchOutputs[readtestGeneratedSharedPath])
	goportOutputs, err := GenerateReadtestGoportOutputs(repoRoot, manifest, spec)
	if err != nil {
		t.Fatalf("GenerateReadtestGoportOutputs error: %v", err)
	}
	assertGeneratedFileMatches(t, repoRoot, readtestGoportDispatchPath, goportOutputs[readtestGoportDispatchPath])
	assertGeneratedFileMatches(t, repoRoot, readtestGoportCasesPath, goportOutputs[readtestGoportCasesPath])
	assertGeneratedFileMatches(t, repoRoot, readtestGoportInventoryPath, goportOutputs[readtestGoportInventoryPath])
	ffiOutputs, err := GenerateFFITestOutputs(spec)
	if err != nil {
		t.Fatalf("GenerateFFITestOutputs error: %v", err)
	}
	assertGeneratedFileMatches(t, repoRoot, ffiGeneratedNativeSupportPath, ffiOutputs[ffiGeneratedNativeSupportPath])
	assertGeneratedFileMatches(t, repoRoot, ffiGeneratedNativeTestPath, ffiOutputs[ffiGeneratedNativeTestPath])
	assertGeneratedFileMatches(t, repoRoot, ffiGeneratedStubTestPath, ffiOutputs[ffiGeneratedStubTestPath])
	tier1ArithmeticOutputs, err := GenerateTier1ArithmeticLongOutputs()
	if err != nil {
		t.Fatalf("GenerateTier1ArithmeticLongOutputs error: %v", err)
	}
	assertGeneratedOutputSet(t, "Tier 1 arithmetic long", tier1ArithmeticOutputs,
		tier1ArithmeticLongGeneratedPath,
		tier1ArithmeticRustLongGeneratedPath,
	)
	for path, data := range tier1ArithmeticOutputs {
		assertGeneratedFileMatches(t, repoRoot, path, data)
	}
	tier1CompareConversionOutputs, err := GenerateTier1CompareConversionLongOutputs()
	if err != nil {
		t.Fatalf("GenerateTier1CompareConversionLongOutputs error: %v", err)
	}
	assertGeneratedOutputSet(t, "Tier 1 compare/conversion long", tier1CompareConversionOutputs,
		tier1CompareConversionLongGeneratedPath,
		tier1CompareConversionRustLongGeneratedPath,
	)
	for path, data := range tier1CompareConversionOutputs {
		assertGeneratedFileMatches(t, repoRoot, path, data)
	}
	dectestOutputs, err := GenerateDectestTestOutputs(repoRoot, spec)
	if err != nil {
		t.Fatalf("GenerateDectestTestOutputs error: %v", err)
	}
	for path, data := range dectestOutputs {
		assertGeneratedFileMatches(t, repoRoot, path, data)
	}
	bidCodecOutputs, err := GenerateBidCodecVectorTestOutputs()
	if err != nil {
		t.Fatalf("GenerateBidCodecVectorTestOutputs error: %v", err)
	}
	for path, data := range bidCodecOutputs {
		assertGeneratedFileMatches(t, repoRoot, path, data)
	}
	bidStringOutputs := GenerateBidStringVectorTestOutputs(spec)
	assertGeneratedFileMatches(t, repoRoot, bidStringVectorsGoTestPath, bidStringOutputs[bidStringVectorsGoTestPath])
	assertGeneratedFileMatches(t, repoRoot, bidStringVectorsRustTestPath, bidStringOutputs[bidStringVectorsRustTestPath])
	bidCodecVectorData, err := GenerateBidCodecVectorData(*manifest.BidCodecVectors)
	if err != nil {
		t.Fatalf("GenerateBidCodecVectorData error: %v", err)
	}
	assertGeneratedFileMatches(t, repoRoot, manifest.BidCodecVectors.Output, bidCodecVectorData)

	if len(spec.DectestSuites) != 4 {
		t.Fatalf("generated %d dectest suites, want 4", len(spec.DectestSuites))
	}
	wantSuiteFiles := map[string]int{
		"ds*.decTest": 1,
		"dd*.decTest": 32,
		"dq*.decTest": 33,
		"*.decTest":   10,
	}
	for _, suite := range spec.DectestSuites {
		want, ok := wantSuiteFiles[suite.Pattern]
		if !ok {
			t.Fatalf("unexpected generated dectest suite pattern %q", suite.Pattern)
		}
		if len(suite.Files) != want {
			t.Fatalf("generated suite %q has %d files, want %d", suite.Pattern, len(suite.Files), want)
		}
		if len(suite.IgnoredOperations) != 1 || suite.IgnoredOperations[0] != "apply" {
			t.Fatalf("generated suite %q ignored operations = %v, want [apply]", suite.Pattern, suite.IgnoredOperations)
		}
		delete(wantSuiteFiles, suite.Pattern)
	}
	if len(wantSuiteFiles) != 0 {
		t.Fatalf("missing generated dectest suites for patterns %v", wantSuiteFiles)
	}
	assertSuiteContainsFile(t, spec.DectestSuites, "dd*.decTest", "tests/ddQuantize.decTest")
	assertSuiteContainsFile(t, spec.DectestSuites, "dd*.decTest", "tests/ddToIntegral.decTest")
	assertSuiteContainsFile(t, spec.DectestSuites, "dd*.decTest", "tests/ddCopySign.decTest")
	assertSuiteContainsFile(t, spec.DectestSuites, "dd*.decTest", "tests/ddClass.decTest")
	assertSuiteContainsFile(t, spec.DectestSuites, "dd*.decTest", "tests/ddSameQuantum.decTest")
	assertSuiteContainsFile(t, spec.DectestSuites, "dd*.decTest", "tests/ddMin.decTest")
	assertSuiteContainsFile(t, spec.DectestSuites, "dd*.decTest", "tests/ddMax.decTest")
	assertSuiteContainsFile(t, spec.DectestSuites, "dd*.decTest", "tests/ddMinMag.decTest")
	assertSuiteContainsFile(t, spec.DectestSuites, "dd*.decTest", "tests/ddMaxMag.decTest")
	assertSuiteContainsFile(t, spec.DectestSuites, "dd*.decTest", "tests/ddCompareTotal.decTest")
	assertSuiteContainsFile(t, spec.DectestSuites, "dd*.decTest", "tests/ddCompareTotalMag.decTest")
	assertSuiteContainsFile(t, spec.DectestSuites, "dd*.decTest", "tests/ddAbs.decTest")
	assertSuiteContainsFile(t, spec.DectestSuites, "dd*.decTest", "tests/ddPlus.decTest")
	assertSuiteContainsFile(t, spec.DectestSuites, "dd*.decTest", "tests/ddMinus.decTest")
	assertSuiteContainsFile(t, spec.DectestSuites, "dd*.decTest", "tests/ddFMA.decTest")
	assertSuiteContainsFile(t, spec.DectestSuites, "dd*.decTest", "tests/ddLogB.decTest")
	assertSuiteContainsFile(t, spec.DectestSuites, "dd*.decTest", "tests/ddScaleB.decTest")
	assertSuiteContainsFile(t, spec.DectestSuites, "dd*.decTest", "tests/ddRemainder.decTest")
	assertSuiteContainsFile(t, spec.DectestSuites, "dd*.decTest", "tests/ddRemainderNear.decTest")
	assertSuiteContainsFile(t, spec.DectestSuites, "dd*.decTest", "tests/ddNextToward.decTest")
	assertSuiteContainsFile(t, spec.DectestSuites, "dd*.decTest", "tests/ddNextPlus.decTest")
	assertSuiteContainsFile(t, spec.DectestSuites, "dd*.decTest", "tests/ddNextMinus.decTest")
	assertSuiteMissingFile(t, spec.DectestSuites, "dd*.decTest", "tests/ddReduce.decTest")
	assertSuiteContainsFile(t, spec.DectestSuites, "dq*.decTest", "tests/dqQuantize.decTest")
	assertSuiteContainsFile(t, spec.DectestSuites, "dq*.decTest", "tests/dqToIntegral.decTest")
	assertSuiteContainsFile(t, spec.DectestSuites, "dq*.decTest", "tests/dqCopySign.decTest")
	assertSuiteContainsFile(t, spec.DectestSuites, "dq*.decTest", "tests/dqClass.decTest")
	assertSuiteContainsFile(t, spec.DectestSuites, "dq*.decTest", "tests/dqSameQuantum.decTest")
	assertSuiteContainsFile(t, spec.DectestSuites, "dq*.decTest", "tests/dqMin.decTest")
	assertSuiteContainsFile(t, spec.DectestSuites, "dq*.decTest", "tests/dqMax.decTest")
	assertSuiteContainsFile(t, spec.DectestSuites, "dq*.decTest", "tests/dqMinMag.decTest")
	assertSuiteContainsFile(t, spec.DectestSuites, "dq*.decTest", "tests/dqMaxMag.decTest")
	assertSuiteContainsFile(t, spec.DectestSuites, "dq*.decTest", "tests/dqCompareTotal.decTest")
	assertSuiteContainsFile(t, spec.DectestSuites, "dq*.decTest", "tests/dqCompareTotalMag.decTest")
	assertSuiteContainsFile(t, spec.DectestSuites, "dq*.decTest", "tests/dqAbs.decTest")
	assertSuiteContainsFile(t, spec.DectestSuites, "dq*.decTest", "tests/dqPlus.decTest")
	assertSuiteContainsFile(t, spec.DectestSuites, "dq*.decTest", "tests/dqMinus.decTest")
	assertSuiteContainsFile(t, spec.DectestSuites, "dq*.decTest", "tests/dqFMA.decTest")
	assertSuiteContainsFile(t, spec.DectestSuites, "dq*.decTest", "tests/dqLogB.decTest")
	assertSuiteContainsFile(t, spec.DectestSuites, "dq*.decTest", "tests/dqScaleB.decTest")
	assertSuiteContainsFile(t, spec.DectestSuites, "dq*.decTest", "tests/dqRemainder.decTest")
	assertSuiteContainsFile(t, spec.DectestSuites, "dq*.decTest", "tests/dqRemainderNear.decTest")
	assertSuiteContainsFile(t, spec.DectestSuites, "dq*.decTest", "tests/dqNextToward.decTest")
	assertSuiteContainsFile(t, spec.DectestSuites, "dq*.decTest", "tests/dqNextPlus.decTest")
	assertSuiteContainsFile(t, spec.DectestSuites, "dq*.decTest", "tests/dqNextMinus.decTest")
	assertSuiteContainsFile(t, spec.DectestSuites, "*.decTest", "tests/quantize.decTest")
	assertSuiteContainsFile(t, spec.DectestSuites, "*.decTest", "tests/tointegral.decTest")
	assertSuiteContainsFile(t, spec.DectestSuites, "*.decTest", "tests/tointegralx.decTest")
	assertSuiteContainsFile(t, spec.DectestSuites, "*.decTest", "tests/comparesig.decTest")
	assertSuiteMissingFile(t, spec.DectestSuites, "*.decTest", "tests/randoms.decTest")
	assertSuiteMissingFile(t, spec.DectestSuites, "*.decTest", "tests/testall.decTest")
	if len(spec.DectestFileInventories) != 144 {
		t.Fatalf("generated dectest inventory file count = %d, want official decTest file count 144", len(spec.DectestFileInventories))
	}
	selectedDectestInventoryFiles := 0
	selectedDectestFiles := map[string]struct{}{}
	for _, suite := range spec.DectestSuites {
		for _, file := range suite.Files {
			selectedDectestFiles[file] = struct{}{}
		}
	}
	inventoriedFiles := map[string]struct{}{}
	unsupportedDectestInventoryFiles := 0
	zeroOperationInventoryFiles := 0
	unsupportedDectestClassifications := map[string]int{}
	for _, inventory := range spec.DectestFileInventories {
		inventoriedFiles[inventory.File] = struct{}{}
		if len(inventory.Operations) == 0 {
			zeroOperationInventoryFiles++
		}
		if len(inventory.SelectedSuites) > 0 {
			selectedDectestInventoryFiles++
		}
		if len(inventory.UnsupportedBySuite) > 0 {
			unsupportedDectestInventoryFiles++
		}
		for suite, unsupported := range inventory.UnsupportedBySuite {
			reasons := inventory.UnsupportedReasonsBySuite[suite]
			classifications := inventory.UnsupportedClassificationsBySuite[suite]
			if len(reasons) != len(unsupported) {
				t.Fatalf("generated dectest inventory %q suite %q reason count = %d, want %d", inventory.File, suite, len(reasons), len(unsupported))
			}
			if len(classifications) != len(unsupported) {
				t.Fatalf("generated dectest inventory %q suite %q classification count = %d, want %d", inventory.File, suite, len(classifications), len(unsupported))
			}
			for _, op := range unsupported {
				if reasons[op] == "" {
					t.Fatalf("generated dectest inventory %q suite %q unsupported op %q missing reason", inventory.File, suite, op)
				}
				unsupportedDectestClassifications[classifications[op]]++
				switch classifications[op] {
				case "out_of_scope_not_required", "optional_not_required", "optional_scope_gap":
				default:
					t.Fatalf("generated dectest inventory %q suite %q unsupported op %q classification = %q", inventory.File, suite, op, classifications[op])
				}
			}
		}
	}
	if selectedDectestInventoryFiles != 76 {
		t.Fatalf("generated dectest selected inventory file count = %d, want current subset count 76", selectedDectestInventoryFiles)
	}
	if unsupportedDectestInventoryFiles != 61 {
		t.Fatalf("generated dectest unsupported inventory file count = %d, want current unsupported count 61", unsupportedDectestInventoryFiles)
	}
	if zeroOperationInventoryFiles != 4 {
		t.Fatalf("generated dectest zero-operation inventory file count = %d, want current metadata-file count 4", zeroOperationInventoryFiles)
	}
	assertCountMap(t, "generated dectest unsupported classifications", unsupportedDectestClassifications, map[string]int{
		"out_of_scope_not_required": 58,
		"optional_not_required":     9,
	})
	const reduceReason = "IBM GDA reduce has no canonical Intel BID predecessor and is outside the strict Intel BID surface"
	assertDectestUnsupportedOperation(t, spec.DectestFileInventories, "tests/ddReduce.decTest", "Decimal64", "reduce", reduceReason, "out_of_scope_not_required")
	assertDectestUnsupportedOperation(t, spec.DectestFileInventories, "tests/dqReduce.decTest", "Decimal128", "reduce", reduceReason, "out_of_scope_not_required")
	assertDectestUnsupportedOperation(t, spec.DectestFileInventories, "tests/reduce.decTest", "General", "reduce", reduceReason, "out_of_scope_not_required")
	for file := range selectedDectestFiles {
		if _, ok := inventoriedFiles[file]; !ok {
			t.Fatalf("generated dectest selected file %q is missing from inventory", file)
		}
	}
	if len(spec.DectestRuntimeSkipInventory) != 4 {
		t.Fatalf("generated dectest runtime skip inventory count = %d, want 4", len(spec.DectestRuntimeSkipInventory))
	}
	dectestRuntimeSkipInventory := map[string]GeneratedDectestRuntimeSkipInventory{}
	for _, inventory := range spec.DectestRuntimeSkipInventory {
		dectestRuntimeSkipInventory[inventory.Suite] = inventory
	}
	assertGeneratedDectestRuntimeSkipInventory(t, dectestRuntimeSkipInventory, "Decimal32", 909, map[string]int{})
	assertGeneratedDectestRuntimeSkipInventory(t, dectestRuntimeSkipInventory, "Decimal64", 11940, map[string]int{
		"fma_nan_payload_precedence":                              13,
		"fma_unsupported_rounding":                                58,
		"ignored_operation_apply":                                 4,
		"minmax_nan_payload_precedence":                           8,
		"minmax_zero_tie":                                         50,
		"nexttoward_nan_payload_precedence":                       2,
		"remainder_gda_division_impossible_context_semantics":     7,
		"remainder_nan_payload_precedence":                        1,
		"remaindernear_gda_division_impossible_context_semantics": 7,
		"remaindernear_nan_payload_precedence":                    1,
		"tagged_to_integral":                                      2,
	})
	assertGeneratedDectestRuntimeSkipInventory(t, dectestRuntimeSkipInventory, "Decimal128", 12313, map[string]int{
		"fma_nan_payload_precedence":                              13,
		"fma_unsupported_rounding":                                74,
		"ignored_operation_apply":                                 371,
		"minmax_nan_payload_precedence":                           8,
		"minmax_zero_tie":                                         50,
		"nexttoward_nan_payload_precedence":                       2,
		"remainder_gda_division_impossible_context_semantics":     7,
		"remainder_nan_payload_precedence":                        1,
		"remaindernear_gda_division_impossible_context_semantics": 7,
		"remaindernear_nan_payload_precedence":                    1,
		"tagged_literal":                                          43,
	})
	assertGeneratedDectestRuntimeSkipInventory(t, dectestRuntimeSkipInventory, "General", 7490, map[string]int{
		"ignored_operation_apply": 20,
		"precision_over_general":  81,
		"tagged_literal":          24,
	})
	if len(spec.ReadCases) == 0 {
		t.Fatal("expected generated read cases")
	}
	if len(spec.ReadCases) != 86250 {
		t.Fatalf("generated read case count = %d, want pinned current-surface plus IEEE regression case count 86250", len(spec.ReadCases))
	}
	assertGeneratedReadtestProfileInventory(t, spec)
	expectedReads := make(map[string]ReadTestSpec)
	for _, read := range manifest.ReadTests {
		expectedReads[read.Name] = read
	}
	for _, group := range manifest.ReadTestGroups {
		for _, read := range expandReadTestGroup(group) {
			expectedReads[read.Name] = read
		}
	}
	for _, profile := range manifest.ReadProfiles {
		reads, err := expandReadTestProfile(repoRoot, profile)
		if err != nil {
			t.Fatalf("expandReadTestProfile(%q): %v", profile.Name, err)
		}
		for _, read := range reads {
			expectedReads[read.Name] = read
		}
	}
	if len(expectedReads) == 0 {
		t.Fatal("expected generated readtest specs")
	}
	seenSuites := map[string]struct{}{}
	readCaseCompareCounts := map[string]int{}
	readCaseFormatCounts := map[string]int{}
	readCaseKindCounts := map[string]int{}
	readCaseGroupCounts := map[string]int{}
	statusControlCaseCounts := map[string]int{}
	for _, tc := range spec.ReadCases {
		read, ok := expectedReads[tc.Suite]
		if !ok {
			t.Fatalf("generated read case suite = %q, not selected by manifest readtests/readtest profile", tc.Suite)
		}
		if tc.Source != filepath.ToSlash(read.Source) {
			t.Fatalf("generated read case source = %q, want %q", tc.Source, filepath.ToSlash(read.Source))
		}
		if tc.Header != filepath.ToSlash(read.Header) {
			t.Fatalf("generated read case header = %q, want %q", tc.Header, filepath.ToSlash(read.Header))
		}
		if tc.OutputType == "" || len(tc.InputTypes) == 0 || tc.CompareGroup == "" {
			t.Fatalf("generated read case metadata is incomplete for %q", tc.Function)
		}
		if tc.OutputType != read.OutputType || tc.CompareGroup != read.CompareGroup {
			t.Fatalf("generated read case metadata for %q = %q/%q, want %q/%q", tc.Function, tc.OutputType, tc.CompareGroup, read.OutputType, read.CompareGroup)
		}
		if strings.Join(tc.InputTypes, ",") != strings.Join(read.InputTypes, ",") {
			t.Fatalf("generated read case input types for %q = %v, want %v", tc.Function, tc.InputTypes, read.InputTypes)
		}
		if tc.Kind != read.Kind {
			t.Fatalf("generated read case kind = %q, want profile kind %q", tc.Kind, read.Kind)
		}
		if len(tc.Operands) == 0 {
			t.Fatal("generated read case is missing operands")
		}
		if tc.Function == "" || tc.Status == "" {
			t.Fatalf("generated read case function/status = %q/%q, want non-empty function and status", tc.Function, tc.Status)
		}
		if tc.Group != read.Group {
			t.Fatalf("generated read case group = %q, want profile group %q", tc.Group, read.Group)
		}
		if tc.Rounding < 0 || tc.Rounding > 4 {
			t.Fatalf("generated read case rounding = %d, want Intel rounding mode 0..4", tc.Rounding)
		}
		seenSuites[tc.Suite] = struct{}{}
		readCaseCompareCounts[tc.CompareGroup]++
		readCaseFormatCounts[tc.Format]++
		readCaseKindCounts[tc.Kind]++
		readCaseGroupCounts[tc.Group]++
		if tc.Kind == "status_control" {
			statusControlCaseCounts[tc.Function]++
		}
	}
	if len(seenSuites) != len(expectedReads) {
		t.Fatalf("generated read suite count = %d, want %d manifest-selected suites", len(seenSuites), len(expectedReads))
	}
	for suite := range expectedReads {
		if _, ok := seenSuites[suite]; !ok {
			t.Fatalf("generated read cases do not include manifest-selected suite %q", suite)
		}
	}
	assertCountMap(t, "generated readtest case compare groups", readCaseCompareCounts, map[string]int{
		"CMP_EQUALSTATUS": 1027,
		"CMP_FUZZYSTATUS": 85223,
	})
	assertCountMap(t, "generated readtest case formats", readCaseFormatCounts, map[string]int{
		"decimal32":  20862,
		"decimal64":  21528,
		"decimal128": 43723,
		"status":     137,
	})
	assertCountMap(t, "generated readtest case kinds", readCaseKindCounts, map[string]int{
		"unary_op":       61299,
		"binary_op":      23028,
		"ternary_op":     1445,
		"from_string":    278,
		"to_string":      63,
		"status_control": 137,
	})
	assertCountMap(t, "generated readtest case groups", readCaseGroupCounts, map[string]int{
		"decimal32_operations":           20737,
		"decimal32_strings":              110,
		"decimal32_ieee754_regressions":  15,
		"decimal64_ieee754_regressions":  15,
		"decimal128_ieee754_regressions": 15,
		"decimal64_operations":           21428,
		"decimal64_strings":              85,
		"decimal128_operations":          43607,
		"decimal128_strings":             101,
		"status_control_operations":      137,
	})
	assertCountMap(t, "generated readtest status-control case functions", statusControlCaseCounts, map[string]int{
		"bid_getDecimalRoundingDirection": 5,
		"bid_lowerFlags":                  17,
		"bid_restoreFlags":                14,
		"bid_saveFlags":                   12,
		"bid_setDecimalRoundingDirection": 26,
		"bid_signalException":             23,
		"bid_testFlags":                   21,
		"bid_testSavedFlags":              19,
	})
	if len(spec.FuzzCases) == 0 {
		t.Fatal("expected generated fuzz cases")
	}
	if len(spec.FFICases) == 0 {
		t.Fatal("expected generated ffi cases")
	}
	if len(spec.FFICases) != 22800 {
		t.Fatalf("generated %d ffi cases, want 22800", len(spec.FFICases))
	}
	ffiSymbols, err := loadSymbolFile(filepath.Join(repoRoot, "generated", "json", "intel_dfp_symbols.json"))
	if err != nil {
		t.Fatalf("load ffi symbols: %v", err)
	}
	ffiSymbolsByName := make(map[string]symbolSpec, len(ffiSymbols.Symbols))
	for _, symbol := range ffiSymbols.Symbols {
		ffiSymbolsByName[symbol.Name] = symbol
	}
	ffiFunctionCaseCounts := map[string]int{}
	ffiFunctionOperations := map[string]string{}
	ffiFunctionBits := map[string]int{}
	ffiFormatCaseCounts := map[string]int{}
	ffiOperationCaseCounts := map[string]int{}
	for _, tc := range spec.FFICases {
		if tc.Source != "generated/json/intel_dfp_symbols.json" {
			t.Fatalf("generated ffi case source = %q, want Intel symbol source", tc.Source)
		}
		if tc.Format != "decimal32" && tc.Format != "decimal64" && tc.Format != "decimal128" {
			t.Fatalf("generated ffi case format = %q, want decimal32, decimal64, or decimal128", tc.Format)
		}
		operationKind, ok := classifyFFIOperation(tc.Operation)
		if !ok {
			t.Fatalf("generated ffi case operation = %q, want supported BID exact-compare subset", tc.Operation)
		}
		if len(tc.Operands) != operationKind.arity {
			t.Fatalf("generated ffi case %q operand count = %d, want %d", tc.Function, len(tc.Operands), operationKind.arity)
		}
		if tc.Operation == "scalbn" || tc.Operation == "ldexp" || tc.Operation == "scalbln" {
			if _, err := strconv.Atoi(tc.Operands[1]); err != nil {
				t.Fatalf("generated ffi case %q scale operand %q is not decimal int: %v", tc.Function, tc.Operands[1], err)
			}
		}
		if tc.Function == "bid128_quantum" && generatedFFIBID128OperandIsNaN(t, tc.Operands[0]) {
			t.Fatalf("generated ffi case %q uses BID128 NaN operand %q; quantum NaN payload exact-compare is intentionally excluded", tc.Function, tc.Operands[0])
		}
		symbol, ok := ffiSymbolsByName[tc.Function]
		if !ok {
			t.Fatalf("generated ffi case %q has no extracted symbol", tc.Function)
		}
		if ffiSymbolHasRoundingParam(symbol) {
			if tc.Rounding < 0 || tc.Rounding > 4 {
				t.Fatalf("generated ffi case %q rounding = %d, want Intel BID rounding mode 0..4", tc.Function, tc.Rounding)
			}
		} else if tc.Rounding != 0 {
			t.Fatalf("generated ffi case %q rounding = %d, want 0 for non-rounding symbol", tc.Function, tc.Rounding)
		}
		ffiFunctionCaseCounts[tc.Function]++
		ffiFunctionOperations[tc.Function] = tc.Operation
		switch tc.Format {
		case "decimal32":
			ffiFunctionBits[tc.Function] = 32
		case "decimal64":
			ffiFunctionBits[tc.Function] = 64
		case "decimal128":
			ffiFunctionBits[tc.Function] = 128
		}
		ffiFormatCaseCounts[tc.Format]++
		ffiOperationCaseCounts[tc.Operation]++
	}
	if len(ffiFunctionCaseCounts) != 453 {
		t.Fatalf("generated ffi function count = %d, want 453 (counts: %v)", len(ffiFunctionCaseCounts), ffiFunctionCaseCounts)
	}
	for function, count := range ffiFunctionCaseCounts {
		want := 48 + 4*ffiTier1RoundingEdgeCaseCount(ffiFunctionOperations[function], ffiFunctionBits[function])
		if count != want {
			t.Fatalf("generated ffi function %q has %d cases, want %d", function, count, want)
		}
	}
	assertCountMap(t, "ffi formats", ffiFormatCaseCounts, map[string]int{
		"decimal32":  7600,
		"decimal64":  7600,
		"decimal128": 7600,
	})
	expectedFFIOperations := map[string]int{
		"abs":                         144,
		"add":                         324,
		"class":                       144,
		"copy":                        144,
		"copySign":                    144,
		"div":                         324,
		"fma":                         144,
		"fmod":                        144,
		"from_int32":                  144,
		"from_int64":                  144,
		"from_uint32":                 144,
		"from_uint64":                 144,
		"isCanonical":                 144,
		"isFinite":                    144,
		"isInf":                       144,
		"isNaN":                       144,
		"isNormal":                    144,
		"isSignaling":                 144,
		"isSigned":                    144,
		"isSubnormal":                 144,
		"isZero":                      144,
		"ilogb":                       144,
		"ldexp":                       144,
		"logb":                        144,
		"llquantexp":                  144,
		"maxnum":                      144,
		"maxnum_mag":                  144,
		"minnum":                      144,
		"minnum_mag":                  144,
		"mul":                         324,
		"negate":                      144,
		"nextdown":                    144,
		"nextup":                      144,
		"quantize":                    324,
		"quiet_equal":                 144,
		"quiet_greater":               144,
		"quiet_greater_equal":         144,
		"quiet_greater_unordered":     144,
		"quiet_less":                  144,
		"quiet_less_equal":            144,
		"quiet_less_unordered":        144,
		"quiet_not_equal":             144,
		"quiet_not_greater":           144,
		"quiet_not_less":              144,
		"quiet_ordered":               144,
		"quiet_unordered":             144,
		"quantexp":                    144,
		"quantum":                     144,
		"radix":                       144,
		"rem":                         144,
		"round_integral_exact":        144,
		"scalbn":                      144,
		"scalbln":                     300,
		"sameQuantum":                 144,
		"signaling_greater":           144,
		"signaling_greater_equal":     144,
		"signaling_greater_unordered": 144,
		"signaling_less":              144,
		"signaling_less_equal":        144,
		"signaling_less_unordered":    144,
		"signaling_not_greater":       144,
		"signaling_not_less":          144,
		"sqrt":                        144,
		"sub":                         324,
		"totalOrder":                  144,
		"totalOrderMag":               144,
		"to_bid128":                   96,
		"to_bid32":                    96,
		"to_bid64":                    96,
		"to_binary128":                144,
		"to_binary32":                 144,
		"to_binary64":                 144,
	}
	for _, kind := range []string{"int8", "int16", "int32", "int64", "uint8", "uint16", "uint32", "uint64"} {
		for _, mode := range []string{"ceil", "floor", "int", "rnint", "rninta", "xceil", "xfloor", "xint", "xrnint", "xrninta"} {
			expectedFFIOperations["to_"+kind+"_"+mode] = 144
		}
	}
	assertCountMap(t, "ffi operations", ffiOperationCaseCounts, expectedFFIOperations)
}

func TestDeterministicFFIGeneratorUsesFormatSpecificEdges(t *testing.T) {
	unaryTests := []struct {
		name  string
		bits  int
		index int
		want  string
	}{
		{name: "bid32_positive_infinity", bits: 32, index: 10, want: "78000000"},
		{name: "bid32_signaling_nan", bits: 32, index: 15, want: "7e000000"},
		{name: "bid64_positive_infinity", bits: 64, index: 10, want: "7800000000000000"},
		{name: "bid64_signaling_nan", bits: 64, index: 15, want: "7e00000000000000"},
		{name: "bid128_negative_zero", bits: 128, index: 1, want: "80000000000000000000000000000000"},
		{name: "bid128_positive_infinity", bits: 128, index: 5, want: "78000000000000000000000000000000"},
		{name: "bid128_signaling_nan", bits: 128, index: 9, want: "7e000000000000000000000000000000"},
	}
	for _, tt := range unaryTests {
		t.Run(tt.name, func(t *testing.T) {
			generator := newDeterministicFFIGenerator(754, "bid32_add", tt.bits)
			got := generator.nextOperands(tt.index, 1)[0]
			if got != tt.want {
				t.Fatalf("format edge operand = %q, want %q", got, tt.want)
			}
		})
	}

	binaryTests := []struct {
		name  string
		bits  int
		index int
		want  []string
	}{
		{name: "bid32_snan_left_finite_right", bits: 32, index: 9, want: []string{"7e000000", "32800001"}},
		{name: "bid32_finite_left_snan_right", bits: 32, index: 10, want: []string{"32800001", "7e000000"}},
		{name: "bid64_noncanonical_left_finite_right", bits: 64, index: 13, want: []string{"6000000000000000", "31c0000000000001"}},
		{name: "bid128_qnan_right", bits: 128, index: 8, want: []string{"30400000000000000000000000000001", "7c000000000000000000000000000000"}},
		{name: "bid128_snan_left", bits: 128, index: 9, want: []string{"7e000000000000000000000000000000", "30400000000000000000000000000001"}},
	}
	for _, tt := range binaryTests {
		t.Run(tt.name, func(t *testing.T) {
			generator := newDeterministicFFIGenerator(754, "bid32_add", tt.bits)
			got := generator.nextOperands(tt.index, 2)
			if strings.Join(got, ",") != strings.Join(tt.want, ",") {
				t.Fatalf("format binary edge operands = %v, want %v", got, tt.want)
			}
		})
	}

	ternaryTests := []struct {
		name  string
		bits  int
		index int
		want  []string
	}{
		{name: "bid32_snan_left", bits: 32, index: 7, want: []string{"7e000000", "32800001", "32800001"}},
		{name: "bid64_snan_right", bits: 64, index: 8, want: []string{"31c0000000000001", "7e00000000000000", "31c0000000000001"}},
		{name: "bid128_snan_left", bits: 128, index: 7, want: []string{"7e000000000000000000000000000000", "30400000000000000000000000000001", "30400000000000000000000000000001"}},
	}
	for _, tt := range ternaryTests {
		t.Run(tt.name, func(t *testing.T) {
			generator := newDeterministicFFIGenerator(754, "bid32_fma", tt.bits)
			got := generator.nextOperands(tt.index, 3)
			if strings.Join(got, ",") != strings.Join(tt.want, ",") {
				t.Fatalf("format ternary edge operands = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildFFICasesCrossesTier1EdgesWithEveryRoundingMode(t *testing.T) {
	repoRoot := filepath.Clean(filepath.Join("..", ".."))
	var functions []string
	for _, bits := range []int{32, 64, 128} {
		for _, operation := range []string{"add", "sub", "mul", "div", "quantize", "scalbln"} {
			functions = append(functions, fmt.Sprintf("bid%d_%s", bits, operation))
		}
	}
	spec := FFITestSpec{
		Name:             "tier1_rounding_edges",
		Symbols:          "generated/json/intel_dfp_symbols.json",
		Functions:        functions,
		CasesPerFunction: 48,
		Seed:             754,
	}
	cases, err := buildFFICases(repoRoot, spec)
	if err != nil {
		t.Fatalf("buildFFICases error: %v", err)
	}

	seen := map[string]bool{}
	for _, tc := range cases {
		key := fmt.Sprintf("%s|%d|%s", tc.Function, tc.Rounding, strings.Join(tc.Operands, ","))
		if seen[key] {
			t.Fatalf("duplicate function/operands/rounding tuple: %s", key)
		}
		seen[key] = true
	}

	for _, bits := range []int{32, 64, 128} {
		for _, operation := range []string{"add", "sub", "mul", "div", "quantize", "scalbln"} {
			function := fmt.Sprintf("bid%d_%s", bits, operation)
			t.Run(function, func(t *testing.T) {
				generator := newDeterministicFFIGenerator(spec.Seed, function, bits)
				edgeCaseCount := ffiTier1RoundingEdgeCaseCount(operation, bits)
				for edgeIndex := 0; edgeIndex < edgeCaseCount; edgeIndex++ {
					wantOperands := generator.nextOperandsForOperation(edgeIndex, operation, 2)
					modes := map[int]bool{}
					for _, tc := range cases {
						if tc.Function == function && reflect.DeepEqual(tc.Operands, wantOperands) {
							modes[tc.Rounding] = true
						}
					}
					for mode := 0; mode < ffiRoundingModeCount; mode++ {
						if !modes[mode] {
							t.Fatalf("edge case %d operands %v missing rounding mode %d (modes: %v)", edgeIndex, wantOperands, mode, modes)
						}
					}
				}
			})
		}
	}
}

func tier1ArithmeticScaleOperand32IsCanonicalNonzeroFinite(raw uint32) bool {
	if raw&0x60000000 == 0x60000000 {
		if raw&0x78000000 == 0x78000000 {
			return false
		}
		coefficient := raw&0x001fffff | 0x00800000
		return coefficient < 10000000
	}
	return raw&0x007fffff != 0
}

func tier1ArithmeticScaleOperand64IsCanonicalNonzeroFinite(raw uint64) bool {
	if raw&0x6000000000000000 == 0x6000000000000000 {
		if raw&0x7800000000000000 == 0x7800000000000000 {
			return false
		}
		coefficient := raw&0x0007ffffffffffff | 0x0020000000000000
		return coefficient < 10000000000000000
	}
	return raw&0x001fffffffffffff != 0
}

func tier1ArithmeticScaleOperand128IsCanonicalNonzeroFinite(raw tier1ArithmeticScaleWords128) bool {
	if raw.hi&0x7800000000000000 >= 0x6000000000000000 {
		return false
	}
	coefficientHi := raw.hi & 0x0001ffffffffffff
	if coefficientHi == 0 && raw.lo == 0 {
		return false
	}
	return coefficientHi < 0x0001ed09bead87c0 ||
		coefficientHi == 0x0001ed09bead87c0 && raw.lo <= 0x378d8e63ffffffff
}

func TestTier1ArithmeticRandomScaleCorpusCoversFiniteTransitionExponentModes(t *testing.T) {
	tests := []struct {
		name  string
		bits  int
		seed  uint64
		lane  uint64
		limit int64
		cases uint64
	}{
		{name: "decimal32", bits: 32, seed: 0xdec7543253414c45, lane: 1, limit: tier1ArithmeticScaleFiniteTransitionLimit32, cases: uint64(1) << 20},
		{name: "decimal64", bits: 64, seed: 0xdec7546453414c45, lane: 1, limit: tier1ArithmeticScaleFiniteTransitionLimit64, cases: uint64(1) << 20},
		{name: "decimal128", bits: 128, seed: 0xdec7541253414c45, lane: 2, limit: tier1ArithmeticScaleFiniteTransitionLimit128, cases: uint64(1) << 19},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			exponentCount := tier1ArithmeticScaleModeCrossGroups(tc.limit)
			seen := make([]bool, exponentCount*tier1ArithmeticScaleModeCross)
			modeSets := make([]uint8, exponentCount)
			operandLo := make([]uint64, exponentCount)
			operandHi := make([]uint64, exponentCount)
			operandSeen := make([]bool, exponentCount)
			var modeCrossCases uint64
			for i := uint64(0); i < tc.cases; i++ {
				exponent := tier1ArithmeticRandomScaleExponentForGeneration(tc.seed, i, tc.lane, tc.limit)
				var lo, hi uint64
				var finite bool
				switch tc.bits {
				case 32:
					raw := tier1ArithmeticRandomScaleOperand32ForGeneration(tc.seed, i, tc.limit)
					lo, finite = uint64(raw), tier1ArithmeticScaleOperand32IsCanonicalNonzeroFinite(raw)
				case 64:
					raw := tier1ArithmeticRandomScaleOperand64ForGeneration(tc.seed, i, tc.limit)
					lo, finite = raw, tier1ArithmeticScaleOperand64IsCanonicalNonzeroFinite(raw)
				case 128:
					raw := tier1ArithmeticRandomScaleOperand128ForGeneration(tc.seed, i, tc.limit)
					lo, hi, finite = raw.lo, raw.hi, tier1ArithmeticScaleOperand128IsCanonicalNonzeroFinite(raw)
				default:
					t.Fatalf("unsupported width %d", tc.bits)
				}

				switch i % tier1ArithmeticScaleRandomStrata {
				case 0:
					want := int64(tier1ArithmeticRandomWordForGeneration(tc.seed, i, tc.lane))
					if exponent != want {
						t.Fatalf("ScaleB full-domain exponent=%d want=%d at case %d", exponent, want, i)
					}
				case 1:
					if !tier1ArithmeticScaleModeCrossCase(i, tc.limit) {
						if exponent < -tc.limit || exponent > tc.limit || !finite {
							t.Fatalf("ScaleB surplus in-range case %d exponent=%d finite=%t", i, exponent, finite)
						}
						continue
					}
					modeCrossCases++
					group := tier1ArithmeticScaleModeCrossGroup(i)
					wantExponent := int64(group) - tc.limit
					if exponent != wantExponent {
						t.Fatalf("ScaleB mode-cross exponent=%d want=%d at case %d", exponent, wantExponent, i)
					}
					if got := tier1ArithmeticRandomScaleOperandIndexForGeneration(i, tc.limit); got != group {
						t.Fatalf("ScaleB mode-cross operand index=%d want=%d at case %d", got, group, i)
					}
					if !finite {
						t.Fatalf("ScaleB mode-cross operand is not canonical nonzero finite at case %d: hi=%016x lo=%016x", i, hi, lo)
					}
					if operandSeen[group] {
						if lo != operandLo[group] || hi != operandHi[group] {
							t.Fatalf("ScaleB mode-cross operand drift in group %d: got=%016x:%016x want=%016x:%016x", group, hi, lo, operandHi[group], operandLo[group])
						}
					} else {
						operandLo[group], operandHi[group], operandSeen[group] = lo, hi, true
					}
					mode := i % tier1ArithmeticScaleModeCross
					modeSets[group] |= 1 << mode
					seen[group*tier1ArithmeticScaleModeCross+mode] = true
				case 2:
					if exponent < -tc.limit*2 || exponent > tc.limit*2 || !finite {
						t.Fatalf("ScaleB transition-window case %d exponent=%d finite=%t", i, exponent, finite)
					}
				case 3:
					if exponent < -tc.limit || exponent > tc.limit || !finite {
						t.Fatalf("ScaleB in-range case %d exponent=%d finite=%t", i, exponent, finite)
					}
				}
			}
			wantModeCrossCases := exponentCount * tier1ArithmeticScaleModeCross
			if modeCrossCases != wantModeCrossCases {
				t.Fatalf("ScaleB complete mode-cross cases=%d want=%d", modeCrossCases, wantModeCrossCases)
			}
			missing := 0
			for _, covered := range seen {
				if !covered {
					missing++
				}
			}
			if missing != 0 {
				t.Fatalf("ScaleB random corpus misses %d/%d finite-transition exponent×rounding-mode cells", missing, len(seen))
			}
			for group, modes := range modeSets {
				if modes != 0x1f {
					t.Fatalf("ScaleB group %d native-mode index set=%05b want=11111", group, modes)
				}
			}
		})
	}
}

func TestBuildFFICasesIncludesTier1MinNumMaxNum(t *testing.T) {
	repoRoot := filepath.Clean(filepath.Join("..", ".."))
	functions := []string{
		"bid32_minnum", "bid32_maxnum",
		"bid64_minnum", "bid64_maxnum",
		"bid128_minnum", "bid128_maxnum",
	}
	cases, err := buildFFICases(repoRoot, FFITestSpec{
		Name:             "tier1_minmax",
		Symbols:          "generated/json/intel_dfp_symbols.json",
		Functions:        functions,
		CasesPerFunction: 48,
		Seed:             754,
	})
	if err != nil {
		t.Fatalf("buildFFICases error: %v", err)
	}

	counts := map[string]int{}
	for _, tc := range cases {
		counts[tc.Function]++
		if tc.Operation != "minnum" && tc.Operation != "maxnum" {
			t.Fatalf("%s operation = %q, want minnum or maxnum", tc.Function, tc.Operation)
		}
		if tc.Rounding != 0 {
			t.Fatalf("%s rounding = %d, want 0", tc.Function, tc.Rounding)
		}
	}
	for _, function := range functions {
		if counts[function] != 48 {
			t.Fatalf("%s cases = %d, want 48", function, counts[function])
		}
	}
}

func assertGeneratedFileMatches(t *testing.T, repoRoot, relativePath string, want []byte) {
	t.Helper()

	got, err := os.ReadFile(filepath.Join(repoRoot, relativePath))
	if err != nil {
		t.Fatalf("ReadFile(%q) error: %v", relativePath, err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("%s is out of date; run `go run ./cmd/testgen -manifest testgen_manifest.json`", relativePath)
	}
}

func assertGeneratedOutputSet(t *testing.T, label string, outputs map[string][]byte, want ...string) {
	t.Helper()
	if len(outputs) != len(want) {
		t.Fatalf("%s generated output count = %d, want %d exact outputs %v", label, len(outputs), len(want), want)
	}
	for _, path := range want {
		if _, ok := outputs[path]; !ok {
			t.Fatalf("%s generator omitted required output %q", label, path)
		}
	}
}

// assertNoStaleShardFiles fails when the checked-in shard directories contain
// JSON files that the current generation no longer produces.
func assertNoStaleShardFiles(t *testing.T, repoRoot string, manifest Manifest, specFiles map[string][]byte) {
	t.Helper()

	outputDir := filepath.ToSlash(filepath.Dir(manifest.Output))
	for _, shardDir := range []string{readtestShardDir, ffiShardDir} {
		dirPath := filepath.Join(repoRoot, filepath.FromSlash(outputDir), shardDir)
		entries, err := os.ReadDir(dirPath)
		if err != nil {
			t.Fatalf("ReadDir(%q) error: %v", dirPath, err)
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
				continue
			}
			rel := outputDir + "/" + shardDir + "/" + entry.Name()
			if _, ok := specFiles[rel]; !ok {
				t.Fatalf("%s is a stale shard file; run `go run ./cmd/testgen -manifest testgen_manifest.json`", rel)
			}
		}
	}
}

// TestSpecIndexShardRoundTripReconstructsSharedSpec writes the generated spec
// through the production WriteOutput into a temporary directory, loads it back
// with the production LoadGenerated, and requires the reconstructed SharedSpec
// to be deeply equal to the original.
func TestSpecIndexShardRoundTripReconstructsSharedSpec(t *testing.T) {
	repoRoot := filepath.Clean(filepath.Join("..", ".."))
	manifest, err := LoadManifest(filepath.Join(repoRoot, "testgen_manifest.json"))
	if err != nil {
		t.Fatalf("LoadManifest error: %v", err)
	}
	requireGenerationInputsForTest(t, repoRoot, manifest)
	spec, err := Generate(repoRoot, manifest)
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	if len(spec.ReadCases) == 0 || len(spec.FFICases) == 0 {
		t.Fatalf("generated spec has %d read cases and %d ffi cases; round trip needs both", len(spec.ReadCases), len(spec.FFICases))
	}

	tempRoot := t.TempDir()
	if err := WriteOutput(tempRoot, manifest, spec); err != nil {
		t.Fatalf("WriteOutput error: %v", err)
	}
	loaded, err := LoadGenerated(filepath.Join(tempRoot, filepath.FromSlash(manifest.Output)))
	if err != nil {
		t.Fatalf("LoadGenerated error: %v", err)
	}
	if !reflect.DeepEqual(spec, loaded) {
		t.Fatalf("LoadGenerated reconstruction differs from the generated SharedSpec")
	}
}

func assertSuiteContainsFile(t *testing.T, suites []GeneratedDectestSuite, pattern, wantFile string) {
	t.Helper()

	for _, suite := range suites {
		if suite.Pattern != pattern {
			continue
		}
		for _, file := range suite.Files {
			if file == wantFile {
				return
			}
		}
		t.Fatalf("generated suite %q does not include %q", pattern, wantFile)
	}

	t.Fatalf("generated suite %q not found", pattern)
}

func assertSuiteMissingFile(t *testing.T, suites []GeneratedDectestSuite, pattern, unwantedFile string) {
	t.Helper()

	for _, suite := range suites {
		if suite.Pattern != pattern {
			continue
		}
		for _, file := range suite.Files {
			if file == unwantedFile {
				t.Fatalf("generated suite %q unexpectedly includes %q", pattern, unwantedFile)
			}
		}
		return
	}

	t.Fatalf("generated suite %q not found", pattern)
}

func assertDectestUnsupportedOperation(
	t *testing.T,
	inventories []GeneratedDectestFileInventory,
	file, suite, operation, wantReason, wantClassification string,
) {
	t.Helper()

	for _, inventory := range inventories {
		if inventory.File != file {
			continue
		}
		if got := inventory.UnsupportedReasonsBySuite[suite][operation]; got != wantReason {
			t.Fatalf("generated dectest inventory %q suite %q operation %q reason = %q, want %q", file, suite, operation, got, wantReason)
		}
		if got := inventory.UnsupportedClassificationsBySuite[suite][operation]; got != wantClassification {
			t.Fatalf("generated dectest inventory %q suite %q operation %q classification = %q, want %q", file, suite, operation, got, wantClassification)
		}
		return
	}

	t.Fatalf("generated dectest inventory %q not found", file)
}

func TestExpandReadTestProfileUsesMechanicalReadtestScope(t *testing.T) {
	repoRoot := filepath.Clean(filepath.Join("..", ".."))
	manifest, err := LoadManifest(filepath.Join(repoRoot, "testgen_manifest.json"))
	if err != nil {
		t.Fatalf("LoadManifest error: %v", err)
	}
	if len(manifest.ReadProfiles) != 1 {
		t.Fatalf("manifest readtest profile count = %d, want 1", len(manifest.ReadProfiles))
	}
	profile := manifest.ReadProfiles[0]
	requireReadtestProfileInputsForTest(t, repoRoot, profile)
	reads, err := expandReadTestProfile(repoRoot, profile)
	if err != nil {
		t.Fatalf("expandReadTestProfile(%q): %v", profile.Name, err)
	}
	if profile.Selection != "repo_supported_surface" {
		t.Fatalf("readtest profile selection = %q, want repo_supported_surface", profile.Selection)
	}
	if len(reads) == 0 {
		t.Fatal("expected mechanically selected readtest functions")
	}
	if len(reads) != 542 {
		t.Fatalf("mechanically selected readtest function count = %d, want pinned Intel Tier 1 surface count 542", len(reads))
	}
	allowedFormats := make(map[string]struct{}, len(profile.Formats))
	for _, format := range profile.Formats {
		allowedFormats[strings.ToLower(strings.TrimSpace(format))] = struct{}{}
	}

	compareCounts := map[string]int{}
	formatCounts := map[string]int{}
	kindCounts := map[string]int{}

	var hasFMA bool
	var hasDecimal128 bool
	var hasBinary32 bool
	var hasBinary64 bool
	var hasBinary128 bool
	var hasToBid128 bool
	var hasNextToward bool
	selectedTier1Mixed := 0
	requiredStatusControl := map[string]bool{
		"bid_getDecimalRoundingDirection": false,
		"bid_lowerFlags":                  false,
		"bid_restoreFlags":                false,
		"bid_saveFlags":                   false,
		"bid_setDecimalRoundingDirection": false,
		"bid_signalException":             false,
		"bid_testFlags":                   false,
		"bid_testSavedFlags":              false,
	}
	for _, read := range reads {
		if isTier1MixedWidthIntelReadtestFunction(read.Function) {
			selectedTier1Mixed++
		}
		if read.OutputType == "" || len(read.InputTypes) == 0 || read.CompareGroup == "" {
			t.Fatalf("selected readtest function %q is missing signature metadata", read.Function)
		}
		if read.CompareGroup != "CMP_FUZZYSTATUS" && read.CompareGroup != "CMP_EQUALSTATUS" {
			t.Fatalf("selected readtest function %q compare group = %q, want CMP_FUZZYSTATUS or CMP_EQUALSTATUS", read.Function, read.CompareGroup)
		}
		compareCounts[read.CompareGroup]++
		formatCounts[read.Format]++
		kindCounts[read.Kind]++
		switch read.Kind {
		case "from_string", "to_string", "binary_op", "unary_op", "ternary_op", "status_control":
		default:
			t.Fatalf("selected readtest kind = %q, want supported generated kind", read.Kind)
		}
		switch read.Kind {
		case "unary_op":
			if len(read.InputTypes) != 1 {
				t.Fatalf("selected unary readtest %q input type count = %d, want 1", read.Function, len(read.InputTypes))
			}
		case "binary_op":
			if len(read.InputTypes) != 2 {
				t.Fatalf("selected binary readtest %q input type count = %d, want 2", read.Function, len(read.InputTypes))
			}
		case "ternary_op":
			if len(read.InputTypes) != 3 {
				t.Fatalf("selected ternary readtest %q input type count = %d, want 3", read.Function, len(read.InputTypes))
			}
			if strings.HasSuffix(read.Function, "_fma") {
				hasFMA = true
			}
		}
		if _, ok := allowedFormats[read.Format]; !ok {
			if read.Format != "status" {
				t.Fatalf("selected readtest format = %q, want one of manifest formats %v or status-control format", read.Format, profile.Formats)
			}
		}
		if read.Format == "decimal128" {
			hasDecimal128 = true
		}
		switch read.Function {
		case "bid32_to_binary32", "bid64_to_binary32", "bid128_to_binary32":
			hasBinary32 = true
		case "bid32_to_binary64", "bid64_to_binary64", "bid128_to_binary64":
			hasBinary64 = true
		case "bid32_to_binary128", "bid64_to_binary128", "bid128_to_binary128":
			hasBinary128 = true
		case "bid32_to_bid128", "bid64_to_bid128":
			hasToBid128 = true
		case "bid32_nexttoward", "bid64_nexttoward":
			hasNextToward = true
		}
		if _, ok := requiredStatusControl[read.Function]; ok {
			requiredStatusControl[read.Function] = true
		}
	}
	if !hasFMA {
		t.Fatal("expected mechanically selected readtest surface to include fma")
	}
	if !hasDecimal128 {
		t.Fatal("expected mechanically selected readtest surface to include decimal128")
	}
	if !hasBinary32 {
		t.Fatal("expected mechanically selected readtest surface to include binary32 conversions")
	}
	if !hasBinary64 {
		t.Fatal("expected mechanically selected readtest surface to include binary64 conversions")
	}
	if !hasBinary128 {
		t.Fatal("expected mechanically selected readtest surface to include binary128 conversions")
	}
	if !hasToBid128 {
		t.Fatal("expected mechanically selected readtest surface to include bid128 widening conversions")
	}
	if !hasNextToward {
		t.Fatal("expected mechanically selected readtest surface to include nexttoward")
	}
	if selectedTier1Mixed != 24 {
		t.Fatalf("selected mixed-width Intel Tier 1 functions = %d, want 24", selectedTier1Mixed)
	}
	for function, seen := range requiredStatusControl {
		if !seen {
			t.Fatalf("expected mechanically selected readtest surface to include status-control function %q", function)
		}
	}
	assertCountMap(t, "readtest compare groups", compareCounts, map[string]int{
		"CMP_FUZZYSTATUS": 530,
		"CMP_EQUALSTATUS": 12,
	})
	assertCountMap(t, "readtest formats", formatCounts, map[string]int{
		"decimal32":  170,
		"decimal64":  182,
		"decimal128": 182,
		"status":     8,
	})
	assertCountMap(t, "readtest kinds", kindCounts, map[string]int{
		"unary_op":       372,
		"binary_op":      153,
		"ternary_op":     3,
		"from_string":    3,
		"to_string":      3,
		"status_control": 8,
	})
}

func requireGenerationInputsForTest(t *testing.T, repoRoot string, manifest Manifest) {
	t.Helper()

	checkedDectestDirs := map[string]struct{}{}
	for _, suite := range manifest.DectestSuites {
		if _, ok := checkedDectestDirs[suite.Directory]; ok {
			continue
		}
		checkedDectestDirs[suite.Directory] = struct{}{}
		matches, err := filepath.Glob(filepath.Join(repoRoot, suite.Directory, "*.decTest"))
		if err != nil {
			t.Fatalf("glob decTest inputs for %q: %v", suite.Directory, err)
		}
		if len(matches) == 0 {
			skipMissingGenerationInput(t, filepath.Join(suite.Directory, "*.decTest"))
		}
	}

	for _, read := range manifest.ReadTests {
		requireGenerationInputPath(t, repoRoot, read.Header)
		requireGenerationInputPath(t, repoRoot, read.Source)
	}
	for _, group := range manifest.ReadTestGroups {
		requireGenerationInputPath(t, repoRoot, group.Header)
		requireGenerationInputPath(t, repoRoot, group.Source)
	}
	for _, profile := range manifest.ReadProfiles {
		requireReadtestProfileInputsForTest(t, repoRoot, profile)
	}
	for _, fuzz := range manifest.FuzzTests {
		for _, source := range fuzz.Sources {
			requireGenerationInputPath(t, repoRoot, source)
		}
	}
}

func requireReadtestProfileInputsForTest(t *testing.T, repoRoot string, profile ReadTestProfileSpec) {
	t.Helper()
	requireGenerationInputPath(t, repoRoot, profile.Header)
	requireGenerationInputPath(t, repoRoot, profile.Source)
}

func requireGenerationInputPath(t *testing.T, repoRoot string, relativePath string) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(repoRoot, relativePath)); err != nil {
		if os.IsNotExist(err) {
			skipMissingGenerationInput(t, relativePath)
		}
		t.Fatalf("stat generation input %q: %v", relativePath, err)
	}
}

func skipMissingGenerationInput(t *testing.T, relativePath string) {
	t.Helper()
	t.Skipf("authoritative generator input %q is not present; run `make setup-generation-inputs` or `make verify-generated` for generated reproducibility checks", relativePath)
}

func assertGeneratedReadtestProfileInventory(t *testing.T, spec SharedSpec) {
	t.Helper()

	if len(spec.ReadtestProfileInventory) != 1 {
		t.Fatalf("generated readtest profile inventory count = %d, want 1", len(spec.ReadtestProfileInventory))
	}
	inventory := spec.ReadtestProfileInventory[0]
	if inventory.Profile != "intel_readtest_current_surface" {
		t.Fatalf("readtest inventory profile = %q, want intel_readtest_current_surface", inventory.Profile)
	}
	if inventory.TotalFunctions != 680 || inventory.SelectedFunctions != 542 || inventory.ExcludedFunctions != 138 {
		t.Fatalf("readtest inventory counts = total %d selected %d excluded %d, want 680/542/138", inventory.TotalFunctions, inventory.SelectedFunctions, inventory.ExcludedFunctions)
	}
	if len(inventory.Functions) != inventory.TotalFunctions {
		t.Fatalf("readtest inventory function entries = %d, want total %d", len(inventory.Functions), inventory.TotalFunctions)
	}

	classificationCounts := map[string]int{}
	unresolvedRequired := map[string]string{}
	for _, fn := range inventory.Functions {
		if fn.Function == "" || fn.Reason == "" || fn.Classification == "" {
			t.Fatalf("readtest inventory entry has incomplete function/reason/classification: %+v", fn)
		}
		classificationCounts[fn.Classification]++
		if fn.Classification == "unresolved_required_review" {
			t.Fatalf("readtest inventory leaves %q as unresolved_required_review: %s", fn.Function, fn.Reason)
		}
		if strings.HasPrefix(fn.Classification, "unresolved_required") {
			unresolvedRequired[fn.Function] = fn.Classification
		}
	}
	assertCountMap(t, "readtest inventory classifications", classificationCounts, map[string]int{
		"selected":                  542,
		"optional_not_required":     87,
		"optional_scope_gap":        16,
		"out_of_scope_not_required": 35,
	})
	if len(unresolvedRequired) != 0 {
		t.Fatalf("readtest inventory unresolved-required functions = %v, want none", unresolvedRequired)
	}
}

func mapStringCounts(values map[string]string) map[string]int {
	counts := make(map[string]int, len(values))
	for key := range values {
		counts[key]++
	}
	return counts
}

func generatedFFIBID128OperandIsNaN(t *testing.T, operand string) bool {
	t.Helper()
	if len(operand) != 32 {
		t.Fatalf("BID128 operand %q length = %d, want 32 hex chars", operand, len(operand))
	}
	raw, err := hex.DecodeString(operand)
	if err != nil {
		t.Fatalf("parse BID128 operand %q: %v", operand, err)
	}
	hi := binary.LittleEndian.Uint64(raw[8:16])
	const bid128NaNMask uint64 = 0x7c00000000000000
	return hi&bid128NaNMask == bid128NaNMask
}

func assertCountMap(t *testing.T, label string, got, want map[string]int) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s count keys = %v, want %v", label, got, want)
	}
	for key, wantCount := range want {
		if got[key] != wantCount {
			t.Fatalf("%s[%q] = %d, want %d (all counts: %v)", label, key, got[key], wantCount, got)
		}
	}
}

func assertGeneratedDectestRuntimeSkipInventory(t *testing.T, inventories map[string]GeneratedDectestRuntimeSkipInventory, suite string, cases int, skipReasons map[string]int) {
	t.Helper()
	inventory, ok := inventories[suite]
	if !ok {
		t.Fatalf("generated dectest runtime skip inventory missing suite %q", suite)
	}
	if inventory.Cases != cases {
		t.Fatalf("generated dectest runtime skip inventory suite %q cases = %d, want %d", suite, inventory.Cases, cases)
	}
	assertCountMap(t, "generated dectest runtime skip inventory "+suite, inventory.SkipReasons, skipReasons)
}
