package testgen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sky1core/bid754/devtools/internal/genmarker"
)

const (
	ffiGeneratedNativeSupportPath = "../bid754-go/generated_ffi_bitcompare_native.go"
	ffiGeneratedNativeTestPath    = "../bid754-go/generated_ffi_bitcompare_native_test.go"
	ffiGeneratedStubTestPath      = "../bid754-go/generated_ffi_bitcompare_stub_test.go"
)

type ffiGeneratedCoverageCounts struct {
	Total      int
	Formats    map[string]int
	Operations map[string]int
	Functions  map[string]int
	Roundings  map[int]int
}

func WriteFFITestOutputs(repoRoot string, spec SharedSpec) error {
	files, err := GenerateFFITestOutputs(spec)
	if err != nil {
		return err
	}
	for path, data := range files {
		fullPath := filepath.Join(repoRoot, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			return fmt.Errorf("mkdir %q: %w", filepath.Dir(fullPath), err)
		}
		if err := os.WriteFile(fullPath, data, 0o644); err != nil {
			return fmt.Errorf("write generated ffi test %q: %w", fullPath, err)
		}
	}
	return nil
}

func countFFICoverage(spec SharedSpec) ffiGeneratedCoverageCounts {
	counts := ffiGeneratedCoverageCounts{
		Formats:    map[string]int{},
		Operations: map[string]int{},
		Functions:  map[string]int{},
		Roundings:  map[int]int{},
	}
	for _, tc := range spec.FFICases {
		counts.Total++
		counts.Formats[tc.Format]++
		counts.Operations[tc.Operation]++
		counts.Functions[tc.Function]++
		counts.Roundings[tc.Rounding]++
	}
	return counts
}

func ffiFusednessRowsGoLiteral() string {
	var b strings.Builder
	for _, row := range ffiMixedFMAFusednessRows() {
		fmt.Fprintf(&b, "\t%q,\n", row)
	}
	return b.String()
}

func ffiFusednessPinsGoLiteral() string {
	var b strings.Builder
	for _, probe := range ffiMixedFMAFusednessProbes {
		fmt.Fprintf(&b, "\t%q: {operands: %#v, rounding: %d, expected: %q, forbidden: %q},\n",
			probe.function, probe.ffiOperands(), probe.rounding,
			probe.expected.ffiString(), probe.forbidden.ffiString())
	}
	return b.String()
}

func ffiNativeTestSource(counts ffiGeneratedCoverageCounts) string {
	return strings.NewReplacer(
		"@@FFI_TOTAL@@", fmt.Sprint(counts.Total),
		"@@FFI_FORMAT_COUNTS@@", stringIntMapLiteral(counts.Formats),
		"@@FFI_OPERATION_COUNTS@@", stringIntMapLiteral(counts.Operations),
		"@@FFI_FUNCTION_COUNTS@@", stringIntMapLiteral(counts.Functions),
		"@@FFI_ROUNDING_COUNTS@@", intIntMapLiteral(counts.Roundings),
		"@@FFI_FUSEDNESS_ROWS@@", ffiFusednessRowsGoLiteral(),
		"@@FFI_FUSEDNESS_PINS@@", ffiFusednessPinsGoLiteral(),
	).Replace(genmarker.Line("testgen") + `
//go:build cgo && bid754_native

package bid754

import (
	"encoding/hex"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/sky1core/bid754/bid754-go/internal/testspec"
)

var expectedGeneratedFFIFormatCounts = map[string]int{
@@FFI_FORMAT_COUNTS@@
}

var expectedGeneratedFFIOperationCounts = map[string]int{
@@FFI_OPERATION_COUNTS@@
}

var expectedGeneratedFFIFunctionCounts = map[string]int{
@@FFI_FUNCTION_COUNTS@@
}

var expectedGeneratedFFIRoundingCounts = map[int]int{
@@FFI_ROUNDING_COUNTS@@
}

const generatedFFIRoundingDiscriminantProbe = "rounding-discriminant"
const generatedFFIFusednessProbe = "fusedness"

var mixedFMAFusednessSentinelRows = []string{
@@FFI_FUSEDNESS_ROWS@@
}

type generatedFFIFusednessPin struct {
	operands  []string
	rounding  int
	expected  string
	forbidden string
}

var expectedGeneratedFFIFusednessPins = map[string]generatedFFIFusednessPin{
@@FFI_FUSEDNESS_PINS@@
}

type generatedFFIRoundingProbeGroup struct {
	function         string
	operands         string
	modes            [5]bool
	nativeResultBits map[string]struct{}
}

type generatedFFIRoundingProbeTracker map[string]*generatedFFIRoundingProbeGroup

func validateGeneratedFFIProbeContract(cases []testspec.GeneratedFFICase) (generatedFFIRoundingProbeTracker, error) {
	tracker := generatedFFIRoundingProbeTracker{}
	groupsPerFunction := map[string]int{}
	fusednessSeen := map[string]string{}
	for _, tc := range cases {
		if tc.Probe == "" {
			if tc.ProbeGroup != "" {
				return nil, fmt.Errorf("generated FFI case %s has probe_group %q without a probe", tc.ID, tc.ProbeGroup)
			}
			if tc.Expected != "" || tc.Forbidden != "" {
				return nil, fmt.Errorf("generated FFI case %s has expected/forbidden without a probe", tc.ID)
			}
			continue
		}
		if !generatedFFIMixedDecimalFunction(tc.Function) {
			return nil, fmt.Errorf("generated FFI probe %s targets non-mixed function %q", tc.ID, tc.Function)
		}
		if tc.ProbeGroup == "" {
			return nil, fmt.Errorf("generated FFI probe %s has no probe_group", tc.ID)
		}
		switch tc.Probe {
		case generatedFFIRoundingDiscriminantProbe:
			if tc.Expected != "" || tc.Forbidden != "" {
				return nil, fmt.Errorf("generated FFI rounding probe %s unexpectedly carries expected/forbidden", tc.ID)
			}
			if tc.Rounding < 0 || tc.Rounding >= 5 {
				return nil, fmt.Errorf("generated FFI rounding probe %s has mode %d outside 0..4", tc.ID, tc.Rounding)
			}
			operandKey := strings.Join(tc.Operands, "\x00")
			group := tracker[tc.ProbeGroup]
			if group == nil {
				group = &generatedFFIRoundingProbeGroup{
					function:         tc.Function,
					operands:         operandKey,
					nativeResultBits: map[string]struct{}{},
				}
				tracker[tc.ProbeGroup] = group
				groupsPerFunction[tc.Function]++
			} else if group.function != tc.Function || group.operands != operandKey {
				return nil, fmt.Errorf("generated FFI rounding probe group %q mixes function/operands: first=(%s,%q) case %s=(%s,%q)", tc.ProbeGroup, group.function, group.operands, tc.ID, tc.Function, operandKey)
			}
			if group.modes[tc.Rounding] {
				return nil, fmt.Errorf("generated FFI rounding probe group %q repeats mode %d", tc.ProbeGroup, tc.Rounding)
			}
			group.modes[tc.Rounding] = true
		case generatedFFIFusednessProbe:
			pin, ok := expectedGeneratedFFIFusednessPins[tc.Function]
			if !ok {
				return nil, fmt.Errorf("generated FFI fusedness probe %s targets function %q outside the closed pin census", tc.ID, tc.Function)
			}
			if prior := fusednessSeen[tc.Function]; prior != "" {
				return nil, fmt.Errorf("generated FFI fusedness function %s repeats in cases %s and %s", tc.Function, prior, tc.ID)
			}
			fusednessSeen[tc.Function] = tc.ID
			if tc.ProbeGroup != tc.Function+"/fusedness" {
				return nil, fmt.Errorf("generated FFI fusedness probe %s group = %q, want %q", tc.ID, tc.ProbeGroup, tc.Function+"/fusedness")
			}
			if tc.Expected == "" || tc.Forbidden == "" {
				return nil, fmt.Errorf("generated FFI fusedness probe %s is missing expected or forbidden outcome", tc.ID)
			}
			if strings.Join(tc.Operands, "\x00") != strings.Join(pin.operands, "\x00") || tc.Rounding != pin.rounding || tc.Expected != pin.expected || tc.Forbidden != pin.forbidden {
				return nil, fmt.Errorf("generated FFI fusedness probe %s payload drift: operands=%v mode=%d expected=%q forbidden=%q, want operands=%v mode=%d expected=%q forbidden=%q", tc.ID, tc.Operands, tc.Rounding, tc.Expected, tc.Forbidden, pin.operands, pin.rounding, pin.expected, pin.forbidden)
			}
			if tc.Expected == tc.Forbidden {
				return nil, fmt.Errorf("generated FFI fusedness probe %s expected equals forbidden outcome %q", tc.ID, tc.Expected)
			}
		default:
			return nil, fmt.Errorf("generated FFI case %s has unknown probe %q", tc.ID, tc.Probe)
		}
	}
	for groupName, group := range tracker {
		for mode, present := range group.modes {
			if !present {
				return nil, fmt.Errorf("generated FFI rounding probe group %q is missing mode %d", groupName, mode)
			}
		}
	}
	for function := range expectedGeneratedFFIFunctionCounts {
		if generatedFFIMixedDecimalFunction(function) && groupsPerFunction[function] != 1 {
			return nil, fmt.Errorf("generated mixed FFI function %s has %d rounding-discriminant probe groups, want 1", function, groupsPerFunction[function])
		}
	}
	if len(fusednessSeen) != len(expectedGeneratedFFIFusednessPins) {
		return nil, fmt.Errorf("generated FFI fusedness function census = %d, want %d", len(fusednessSeen), len(expectedGeneratedFFIFusednessPins))
	}
	for function := range expectedGeneratedFFIFusednessPins {
		if fusednessSeen[function] == "" {
			return nil, fmt.Errorf("generated FFI fusedness pin %s has no case", function)
		}
	}
	return tracker, nil
}

func (tracker generatedFFIRoundingProbeTracker) recordCanonicalResult(tc testspec.GeneratedFFICase, native string) error {
	if tc.Probe != generatedFFIRoundingDiscriminantProbe {
		return nil
	}
	group := tracker[tc.ProbeGroup]
	if group == nil {
		return fmt.Errorf("generated FFI rounding probe %s references unknown group %q", tc.ID, tc.ProbeGroup)
	}
	resultBits, status, ok := strings.Cut(native, "/")
	if !ok || len(resultBits) != tc.ResultBits/4 || len(status) != 8 {
		return fmt.Errorf("generated FFI rounding probe %s native result %q is not %d-bit-result/32-bit-status hex", tc.ID, native, tc.ResultBits)
	}
	if _, err := hex.DecodeString(resultBits); err != nil {
		return fmt.Errorf("generated FFI rounding probe %s native result bits %q: %w", tc.ID, resultBits, err)
	}
	if _, err := hex.DecodeString(status); err != nil {
		return fmt.Errorf("generated FFI rounding probe %s native status %q: %w", tc.ID, status, err)
	}
	group.nativeResultBits[resultBits] = struct{}{}
	return nil
}

func (tracker generatedFFIRoundingProbeTracker) validateCanonicalDiscrimination() error {
	for groupName, group := range tracker {
		if len(group.nativeResultBits) < 2 {
			return fmt.Errorf("generated FFI rounding probe group %q (%s operands %q): canonical Intel C produced %d distinct result bit patterns across modes 0..4, want at least 2", groupName, group.function, group.operands, len(group.nativeResultBits))
		}
	}
	return nil
}

func cloneGeneratedFFICases(cases []testspec.GeneratedFFICase) []testspec.GeneratedFFICase {
	cloned := append([]testspec.GeneratedFFICase(nil), cases...)
	for i := range cloned {
		cloned[i].OperandBits = append([]int(nil), cloned[i].OperandBits...)
		cloned[i].Operands = append([]string(nil), cloned[i].Operands...)
	}
	return cloned
}

func testGeneratedFFIProbeValidatorRejectsMutations(t *testing.T, cases []testspec.GeneratedFFICase) {
	t.Helper()
	firstRounding := -1
	firstFusedness := -1
	firstBaseline := -1
	for i, tc := range cases {
		switch tc.Probe {
		case generatedFFIRoundingDiscriminantProbe:
			if firstRounding < 0 {
				firstRounding = i
			}
		case generatedFFIFusednessProbe:
			if firstFusedness < 0 {
				firstFusedness = i
			}
		case "":
			if firstBaseline < 0 {
				firstBaseline = i
			}
		}
	}
	if firstRounding < 0 || firstFusedness < 0 || firstBaseline < 0 {
		t.Fatalf("generated FFI mutation fixtures missing: rounding=%d fusedness=%d baseline=%d", firstRounding, firstFusedness, firstBaseline)
	}
	roundingFunction := cases[firstRounding].Function
	roundingGroup := cases[firstRounding].ProbeGroup
	var roundingGroupIndices []int
	for i, tc := range cases {
		if tc.Probe == generatedFFIRoundingDiscriminantProbe && tc.ProbeGroup == roundingGroup {
			roundingGroupIndices = append(roundingGroupIndices, i)
		}
	}
	if len(roundingGroupIndices) != 5 {
		t.Fatalf("generated FFI rounding mutation group %q has %d cases, want 5", roundingGroup, len(roundingGroupIndices))
	}

	type mutation struct {
		name    string
		wantErr string
		apply   func([]testspec.GeneratedFFICase) []testspec.GeneratedFFICase
	}
	mutations := []mutation{
		{name: "unknown probe", wantErr: "unknown probe", apply: func(got []testspec.GeneratedFFICase) []testspec.GeneratedFFICase {
			got[firstRounding].Probe = "unknown-probe"
			return got
		}},
		{name: "orphan group", wantErr: "without a probe", apply: func(got []testspec.GeneratedFFICase) []testspec.GeneratedFFICase {
			got[firstBaseline].ProbeGroup = "orphan/group"
			return got
		}},
		{name: "empty group", wantErr: "has no probe_group", apply: func(got []testspec.GeneratedFFICase) []testspec.GeneratedFFICase {
			got[firstRounding].ProbeGroup = ""
			return got
		}},
		{name: "non-mixed target", wantErr: "targets non-mixed function", apply: func(got []testspec.GeneratedFFICase) []testspec.GeneratedFFICase {
			got[firstRounding].Function = "bid64_add"
			return got
		}},
		{name: "mode out of range", wantErr: "outside 0..4", apply: func(got []testspec.GeneratedFFICase) []testspec.GeneratedFFICase {
			got[firstRounding].Rounding = 5
			return got
		}},
		{name: "duplicate mode", wantErr: "repeats mode", apply: func(got []testspec.GeneratedFFICase) []testspec.GeneratedFFICase {
			got[roundingGroupIndices[4]].Rounding = got[roundingGroupIndices[0]].Rounding
			return got
		}},
		{name: "missing mode", wantErr: "is missing mode", apply: func(got []testspec.GeneratedFFICase) []testspec.GeneratedFFICase {
			index := roundingGroupIndices[4]
			return append(got[:index], got[index+1:]...)
		}},
		{name: "mixed function", wantErr: "mixes function/operands", apply: func(got []testspec.GeneratedFFICase) []testspec.GeneratedFFICase {
			got[roundingGroupIndices[4]].Function = "bid128d_sqrt"
			return got
		}},
		{name: "mixed operands", wantErr: "mixes function/operands", apply: func(got []testspec.GeneratedFFICase) []testspec.GeneratedFFICase {
			got[roundingGroupIndices[4]].Operands[0] = strings.Repeat("0", len(got[roundingGroupIndices[4]].Operands[0]))
			return got
		}},
		{name: "required function probe missing", wantErr: "rounding-discriminant probe groups, want 1", apply: func(got []testspec.GeneratedFFICase) []testspec.GeneratedFFICase {
			filtered := got[:0]
			for _, tc := range got {
				if tc.Function != roundingFunction || tc.Probe != generatedFFIRoundingDiscriminantProbe {
					filtered = append(filtered, tc)
				}
			}
			return filtered
		}},
		{name: "fused expected drift", wantErr: "payload drift", apply: func(got []testspec.GeneratedFFICase) []testspec.GeneratedFFICase {
			got[firstFusedness].Expected = "0000000000000000/00000000"
			return got
		}},
		{name: "fused forbidden drift", wantErr: "payload drift", apply: func(got []testspec.GeneratedFFICase) []testspec.GeneratedFFICase {
			got[firstFusedness].Forbidden = "0000000000000000/00000000"
			return got
		}},
		{name: "fused expected missing", wantErr: "missing expected or forbidden", apply: func(got []testspec.GeneratedFFICase) []testspec.GeneratedFFICase {
			got[firstFusedness].Expected = ""
			return got
		}},
		{name: "fused forbidden missing", wantErr: "missing expected or forbidden", apply: func(got []testspec.GeneratedFFICase) []testspec.GeneratedFFICase {
			got[firstFusedness].Forbidden = ""
			return got
		}},
		{name: "fusedness probe missing", wantErr: "fusedness function census", apply: func(got []testspec.GeneratedFFICase) []testspec.GeneratedFFICase {
			return append(got[:firstFusedness], got[firstFusedness+1:]...)
		}},
	}
	for _, mutation := range mutations {
		mutation := mutation
		t.Run(mutation.name, func(t *testing.T) {
			mutated := mutation.apply(cloneGeneratedFFICases(cases))
			_, err := validateGeneratedFFIProbeContract(mutated)
			if err == nil || !strings.Contains(err.Error(), mutation.wantErr) {
				t.Fatalf("validateGeneratedFFIProbeContract error = %v, want rejection containing %q", err, mutation.wantErr)
			}
		})
	}

	t.Run("malformed native result", func(t *testing.T) {
		tracker, err := validateGeneratedFFIProbeContract(cases)
		if err != nil {
			t.Fatalf("validateGeneratedFFIProbeContract(valid): %v", err)
		}
		if err := tracker.recordCanonicalResult(cases[firstRounding], "malformed"); err == nil {
			t.Fatal("recordCanonicalResult accepted malformed native result")
		}
	})

	t.Run("canonical result does not discriminate", func(t *testing.T) {
		tracker, err := validateGeneratedFFIProbeContract(cases)
		if err != nil {
			t.Fatalf("validateGeneratedFFIProbeContract(valid): %v", err)
		}
		for _, tc := range cases {
			if tc.Probe != generatedFFIRoundingDiscriminantProbe {
				continue
			}
			native := strings.Repeat("0", tc.ResultBits/4) + "/00000000"
			if err := tracker.recordCanonicalResult(tc, native); err != nil {
				t.Fatalf("recordCanonicalResult(%s): %v", tc.ID, err)
			}
		}
		if err := tracker.validateCanonicalDiscrimination(); err == nil {
			t.Fatal("validateCanonicalDiscrimination accepted one C result bit pattern per group")
		}
	})
}

func TestGeneratedFFIProbeValidatorRejectsMutations(t *testing.T) {
	spec := loadGeneratedFFISpecForTest(t)
	if len(spec.FFICases) == 0 {
		t.Fatal("expected generated ffi cases")
	}
	testGeneratedFFIProbeValidatorRejectsMutations(t, spec.FFICases)
}

func TestGeneratedFFIBitCompareSubset(t *testing.T) {
	requireNative(t)
	if testing.Short() {
		t.Skip("generated FFI bit-compare cases require non-short native run; use make test-native-ffi")
	}
	spec := loadGeneratedFFISpecForTest(t)

	if len(spec.FFICases) == 0 {
		t.Fatal("expected generated ffi cases")
	}
	if len(spec.FFICases) != @@FFI_TOTAL@@ {
		 t.Fatalf("generated ffi case count = %d, want @@FFI_TOTAL@@", len(spec.FFICases))
	}
	assertGeneratedFFICoverage(t, spec.FFICases)
	probeTracker, err := validateGeneratedFFIProbeContract(spec.FFICases)
	if err != nil {
		t.Fatalf("validate generated FFI probe contract: %v", err)
	}

	for _, tc := range spec.FFICases {
		tc := tc
			t.Run(tc.ID, func(t *testing.T) {
			generated := generatedFFICase(tc)
			gotNative, gotExposed, err := runGeneratedFFICase(generated)
			if err != nil {
				t.Fatalf("runGeneratedFFICase(%s): %v", tc.ID, err)
			}
			if gotNative != gotExposed {
				t.Fatalf("%s %s(%s): C=%s exposed=%s", tc.Declaration, tc.Function, strings.Join(tc.Operands, ", "), gotNative, gotExposed)
			}
			if tc.Probe == generatedFFIFusednessProbe {
				if gotNative != tc.Expected || gotExposed != tc.Expected {
					t.Fatalf("fusedness sentinel %s direct mismatch: pinned=%s C=%s Go-port=%s", tc.Function, tc.Expected, gotNative, gotExposed)
				}
				gotComposed, err := runGeneratedFFIMixedFMAComposed(generated)
				if err != nil {
					t.Fatalf("runGeneratedFFIMixedFMAComposed(%s): %v", tc.ID, err)
				}
				if gotComposed != tc.Forbidden {
					t.Fatalf("fusedness sentinel %s sequential composition drift: pinned forbidden=%s composed=%s", tc.Function, tc.Forbidden, gotComposed)
				}
				if gotComposed == gotNative {
					t.Fatalf("fusedness sentinel %s no longer discriminates direct FMA from sequential composition: %s", tc.Function, gotNative)
				}
			}
			if err := probeTracker.recordCanonicalResult(tc, gotNative); err != nil {
				t.Fatalf("record canonical FFI probe result: %v", err)
			}
		})
	}
	if err := probeTracker.validateCanonicalDiscrimination(); err != nil {
		t.Fatal(err)
	}
}

func assertGeneratedFFICoverage(t *testing.T, cases []testspec.GeneratedFFICase) {
	t.Helper()
	formatCounts := map[string]int{}
	operationCounts := map[string]int{}
	functionCounts := map[string]int{}
	roundingCounts := map[int]int{}
	for _, tc := range cases {
		formatCounts[tc.Format]++
		operationCounts[tc.Operation]++
		functionCounts[tc.Function]++
		roundingCounts[tc.Rounding]++
	}
	assertGeneratedFFIStringCounts(t, "format", formatCounts, expectedGeneratedFFIFormatCounts)
	assertGeneratedFFIStringCounts(t, "operation", operationCounts, expectedGeneratedFFIOperationCounts)
	assertGeneratedFFIStringCounts(t, "function", functionCounts, expectedGeneratedFFIFunctionCounts)
	assertGeneratedFFIIntCounts(t, "rounding", roundingCounts, expectedGeneratedFFIRoundingCounts)
}

func assertGeneratedFFIStringCounts(t *testing.T, label string, got, want map[string]int) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("generated ffi %s bucket count = %d, want %d", label, len(got), len(want))
	}
	for key, wantValue := range want {
		if got[key] != wantValue {
			t.Fatalf("generated ffi %s count[%q] = %d, want %d", label, key, got[key], wantValue)
		}
	}
}

func assertGeneratedFFIIntCounts(t *testing.T, label string, got, want map[int]int) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("generated ffi %s bucket count = %d, want %d", label, len(got), len(want))
	}
	for key, wantValue := range want {
		if got[key] != wantValue {
			t.Fatalf("generated ffi %s count[%d] = %d, want %d", label, key, got[key], wantValue)
		}
	}
}

func loadGeneratedFFISpecForTest(t *testing.T) testspec.SharedSpec {
	t.Helper()
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("resolve generated ffi file path")
	}
	spec, err := testspec.LoadGenerated(filepath.Join(filepath.Dir(currentFile), "..", "devtools", "generated", "testspec", "spec_index.json"))
	if err != nil {
		t.Fatalf("load shared spec: %v", err)
	}
	return spec
}
`)
}

func GenerateFFITestOutputs(spec SharedSpec) (map[string][]byte, error) {
	counts := countFFICoverage(spec)
	return formatGeneratedGoOutputs(map[string][]byte{
		ffiGeneratedNativeSupportPath: []byte(genmarker.Line("testgen") + `
//go:build cgo && bid754_native

package bid754

/*
#cgo CFLAGS: -DDECNUMDIGITS=34 -I${SRCDIR}/../devtools/third_party/intel_dfp/src -I${SRCDIR}/../devtools/third_party/intel_dfp/include
#cgo LDFLAGS: -ldecnumber -L${SRCDIR}/../devtools/third_party/intel_dfp/lib -lbid -lm

#include <stdint.h>
#include "bid_conf.h"
#include "bid_functions.h"

*/
import "C"

import (
	"encoding/hex"
	"fmt"
	"math"
	"strconv"
	"strings"
	"unsafe"

	bidgo "github.com/sky1core/bid754/bid754-go/internal/bidgo"
)

func generatedFFICRoundingMode(rounding int) C._IDEC_round {
	return C._IDEC_round(rounding)
}

type generatedFFICase struct {
	Suite       string
	ID          string
	Format      string
	ResultBits  int
	OperandBits []int
	Operation   string
	Function    string
	LinkName    string
	Declaration string
	Source      string
	Probe       string
	ProbeGroup  string
	Expected    string
	Forbidden   string
	Rounding    int
	Operands    []string
}

func runGeneratedFFICase(tc generatedFFICase) (string, string, error) {
	if generatedFFIMixedDecimalFunction(tc.Function) {
		return runGeneratedFFICaseMixedDecimal(tc)
	}
	if generatedFFIBaseIntegerFromOperation(tc.Operation) {
		return runGeneratedFFICaseBaseIntegerFrom(tc)
	}
	if generatedFFIBIDWidthConversionOperation(tc.Operation) {
		return runGeneratedFFICaseBIDWidthConversion(tc)
	}
	if generatedFFIBinaryConversionOperation(tc.Operation) {
		return runGeneratedFFICaseBinaryConversion(tc)
	}
	if generatedFFIBaseIntegerToOperation(tc.Operation) {
		native, err := runGeneratedFFICaseNativeBaseIntegerTo(tc)
		if err != nil {
			return "", "", err
		}
		exposed, err := runGeneratedFFICaseGoBaseIntegerTo(tc)
		if err != nil {
			return "", "", err
		}
		return native, exposed, nil
	}
	if generatedFFIInt64UnaryOperation(tc.Operation) {
		switch tc.Format {
		case "decimal32":
			a, err := parseFFIUint32UnaryOperand(tc)
			if err != nil {
				return "", "", err
			}
			native, nativeFlags, exposed, exposedFlags := runGeneratedFFICase32Int64Unary(tc.Operation, a)
			return formatGeneratedFFIInt64Result(native, nativeFlags), formatGeneratedFFIInt64Result(exposed, exposedFlags), nil
		case "decimal64":
			a, err := parseFFIUint64UnaryOperand(tc)
			if err != nil {
				return "", "", err
			}
			native, nativeFlags, exposed, exposedFlags := runGeneratedFFICase64Int64Unary(tc.Operation, a)
			return formatGeneratedFFIInt64Result(native, nativeFlags), formatGeneratedFFIInt64Result(exposed, exposedFlags), nil
		case "decimal128":
			a, err := parseFFIUint128UnaryOperand(tc)
			if err != nil {
				return "", "", err
			}
			native, nativeFlags, exposed, exposedFlags := runGeneratedFFICase128Int64Unary(tc.Operation, a)
			return formatGeneratedFFIInt64Result(native, nativeFlags), formatGeneratedFFIInt64Result(exposed, exposedFlags), nil
		}
	}
	if generatedFFIIntUnaryOperation(tc.Operation) {
		switch tc.Format {
		case "decimal32":
			a, err := parseFFIUint32UnaryOperand(tc)
			if err != nil {
				return "", "", err
			}
			native, nativeFlags, exposed, exposedFlags := runGeneratedFFICase32IntUnary(tc.Operation, a)
			return formatGeneratedFFIIntResult(native, nativeFlags), formatGeneratedFFIIntResult(exposed, exposedFlags), nil
		case "decimal64":
			a, err := parseFFIUint64UnaryOperand(tc)
			if err != nil {
				return "", "", err
			}
			native, nativeFlags, exposed, exposedFlags := runGeneratedFFICase64IntUnary(tc.Operation, a)
			return formatGeneratedFFIIntResult(native, nativeFlags), formatGeneratedFFIIntResult(exposed, exposedFlags), nil
		case "decimal128":
			a, err := parseFFIUint128UnaryOperand(tc)
			if err != nil {
				return "", "", err
			}
			native, nativeFlags, exposed, exposedFlags := runGeneratedFFICase128IntUnary(tc.Operation, a)
			return formatGeneratedFFIIntResult(native, nativeFlags), formatGeneratedFFIIntResult(exposed, exposedFlags), nil
		}
	}
	if generatedFFIIntBinaryOperation(tc.Operation) {
		switch tc.Format {
		case "decimal32":
			a, b, err := parseFFIUint32BinaryOperands(tc)
			if err != nil {
				return "", "", err
			}
			native, nativeFlags, exposed, exposedFlags := runGeneratedFFICase32IntBinary(tc.Operation, a, b)
			return formatGeneratedFFIIntResult(native, nativeFlags), formatGeneratedFFIIntResult(exposed, exposedFlags), nil
		case "decimal64":
			a, b, err := parseFFIUint64BinaryOperands(tc)
			if err != nil {
				return "", "", err
			}
			native, nativeFlags, exposed, exposedFlags := runGeneratedFFICase64IntBinary(tc.Operation, a, b)
			return formatGeneratedFFIIntResult(native, nativeFlags), formatGeneratedFFIIntResult(exposed, exposedFlags), nil
		case "decimal128":
			a, b, err := parseFFIUint128BinaryOperands(tc)
			if err != nil {
				return "", "", err
			}
			native, nativeFlags, exposed, exposedFlags := runGeneratedFFICase128IntBinary(tc.Operation, a, b)
			return formatGeneratedFFIIntResult(native, nativeFlags), formatGeneratedFFIIntResult(exposed, exposedFlags), nil
		}
	}
	if generatedFFIDecimalFlagUnaryOperation(tc.Operation) {
		switch tc.Format {
		case "decimal32":
			a, err := parseFFIUint32UnaryOperand(tc)
			if err != nil {
				return "", "", err
			}
			native, nativeFlags, exposed, exposedFlags := runGeneratedFFICase32DecimalFlagUnary(tc.Operation, a)
			return fmt.Sprintf("%08x/%08x", native, nativeFlags), fmt.Sprintf("%08x/%08x", exposed, exposedFlags), nil
		case "decimal64":
			a, err := parseFFIUint64UnaryOperand(tc)
			if err != nil {
				return "", "", err
			}
			native, nativeFlags, exposed, exposedFlags := runGeneratedFFICase64DecimalFlagUnary(tc.Operation, a)
			return fmt.Sprintf("%016x/%08x", native, nativeFlags), fmt.Sprintf("%016x/%08x", exposed, exposedFlags), nil
		case "decimal128":
			a, err := parseFFIUint128UnaryOperand(tc)
			if err != nil {
				return "", "", err
			}
			native, nativeFlags, exposed, exposedFlags := runGeneratedFFICase128DecimalFlagUnary(tc.Operation, a)
			return fmt.Sprintf("%s/%08x", formatFFIUint128Bits(native), nativeFlags), fmt.Sprintf("%s/%08x", formatFFIUint128Bits(exposed), exposedFlags), nil
		}
	}
	if generatedFFIDecimalFlagIntOperation(tc.Operation) {
		switch tc.Format {
		case "decimal32":
			a, n, err := parseFFIUint32IntOperands(tc)
			if err != nil {
				return "", "", err
			}
			native, nativeFlags, exposed, exposedFlags := runGeneratedFFICase32DecimalFlagInt(tc.Operation, a, n, tc.Rounding)
			return fmt.Sprintf("%08x/%08x", native, nativeFlags), fmt.Sprintf("%08x/%08x", exposed, exposedFlags), nil
		case "decimal64":
			a, n, err := parseFFIUint64IntOperands(tc)
			if err != nil {
				return "", "", err
			}
			native, nativeFlags, exposed, exposedFlags := runGeneratedFFICase64DecimalFlagInt(tc.Operation, a, n, tc.Rounding)
			return fmt.Sprintf("%016x/%08x", native, nativeFlags), fmt.Sprintf("%016x/%08x", exposed, exposedFlags), nil
		case "decimal128":
			a, n, err := parseFFIUint128IntOperands(tc)
			if err != nil {
				return "", "", err
			}
			native, nativeFlags, exposed, exposedFlags := runGeneratedFFICase128DecimalFlagInt(tc.Operation, a, n, tc.Rounding)
			return fmt.Sprintf("%s/%08x", formatFFIUint128Bits(native), nativeFlags), fmt.Sprintf("%s/%08x", formatFFIUint128Bits(exposed), exposedFlags), nil
		}
	}

	switch tc.Function {
	case "bid32_fma":
		a, b, c, err := parseFFIUint32TernaryOperands(tc)
		if err != nil {
			return "", "", err
		}
		native, exposed := runGeneratedFFICase32Ternary(tc.Function, a, b, c, tc.Rounding)
		return native, exposed, nil
	case "bid32_add", "bid32_sub", "bid32_mul", "bid32_div", "bid32_quantize", "bid32_rem", "bid32_fmod", "bid32_minnum", "bid32_maxnum", "bid32_minnum_mag", "bid32_maxnum_mag", "bid32_copySign":
		a, b, err := parseFFIUint32BinaryOperands(tc)
		if err != nil {
			return "", "", err
		}
		native, exposed := runGeneratedFFICase32Binary(tc.Function, a, b, tc.Rounding)
		return native, exposed, nil
	case "bid32_round_integral_exact", "bid32_copy", "bid32_negate", "bid32_abs", "bid32_sqrt", "bid32_logb", "bid32_nextup", "bid32_nextdown":
		a, err := parseFFIUint32UnaryOperand(tc)
		if err != nil {
			return "", "", err
		}
		native, exposed := runGeneratedFFICase32Unary(tc.Function, a, tc.Rounding)
		return native, exposed, nil
	case "bid32_scalbn":
		a, n, err := parseFFIUint32IntOperands(tc)
		if err != nil {
			return "", "", err
		}
		native, exposed := runGeneratedFFICase32Int(tc.Function, a, n, tc.Rounding)
		return native, exposed, nil
	case "bid64_fma":
		a, b, c, err := parseFFIUint64TernaryOperands(tc)
		if err != nil {
			return "", "", err
		}
		native, exposed := runGeneratedFFICase64Ternary(tc.Function, a, b, c, tc.Rounding)
		return native, exposed, nil
	case "bid64_add", "bid64_sub", "bid64_mul", "bid64_div", "bid64_quantize", "bid64_rem", "bid64_fmod", "bid64_minnum", "bid64_maxnum", "bid64_minnum_mag", "bid64_maxnum_mag", "bid64_copySign":
		a, b, err := parseFFIUint64BinaryOperands(tc)
		if err != nil {
			return "", "", err
		}
		native, exposed := runGeneratedFFICase64Binary(tc.Function, a, b, tc.Rounding)
		return native, exposed, nil
	case "bid64_round_integral_exact", "bid64_copy", "bid64_negate", "bid64_abs", "bid64_sqrt", "bid64_logb", "bid64_nextup", "bid64_nextdown":
		a, err := parseFFIUint64UnaryOperand(tc)
		if err != nil {
			return "", "", err
		}
		native, exposed := runGeneratedFFICase64Unary(tc.Function, a, tc.Rounding)
		return native, exposed, nil
	case "bid64_scalbn":
		a, n, err := parseFFIUint64IntOperands(tc)
		if err != nil {
			return "", "", err
		}
		native, exposed := runGeneratedFFICase64Int(tc.Function, a, n, tc.Rounding)
		return native, exposed, nil
	case "bid128_fma":
		a, b, c, err := parseFFIUint128TernaryOperands(tc)
		if err != nil {
			return "", "", err
		}
		native, exposed := runGeneratedFFICase128Ternary(tc.Function, a, b, c, tc.Rounding)
		return native, exposed, nil
	case "bid128_add", "bid128_sub", "bid128_mul", "bid128_div", "bid128_quantize", "bid128_rem", "bid128_fmod", "bid128_minnum", "bid128_maxnum", "bid128_minnum_mag", "bid128_maxnum_mag", "bid128_copySign":
		a, b, err := parseFFIUint128BinaryOperands(tc)
		if err != nil {
			return "", "", err
		}
		native, exposed := runGeneratedFFICase128Binary(tc.Function, a, b, tc.Rounding)
		return native, exposed, nil
	case "bid128_round_integral_exact", "bid128_copy", "bid128_negate", "bid128_abs", "bid128_sqrt", "bid128_logb", "bid128_nextup", "bid128_nextdown":
		a, err := parseFFIUint128UnaryOperand(tc)
		if err != nil {
			return "", "", err
		}
		native, exposed := runGeneratedFFICase128Unary(tc.Function, a, tc.Rounding)
		return native, exposed, nil
	case "bid128_scalbn":
		a, n, err := parseFFIUint128IntOperands(tc)
		if err != nil {
			return "", "", err
		}
		native, exposed := runGeneratedFFICase128Int(tc.Function, a, n, tc.Rounding)
		return native, exposed, nil
	default:
		return "", "", fmt.Errorf("unsupported generated ffi function %q", tc.Function)
	}
}

type generatedFFIMixedDecimalShape struct {
	format      string
	operation   string
	resultBits  int
	operandBits []int
}

type generatedFFIMixedDecimalOperands struct {
	narrow [3]uint64
	wide   [3]Decimal128BID
}

func generatedFFIMixedDecimalShapeFor(function string) (generatedFFIMixedDecimalShape, bool) {
	switch function {
	case "bid64ddq_fma":
		return generatedFFIMixedDecimalShape{"decimal64", "fma", 64, []int{64, 64, 128}}, true
	case "bid64dqd_fma":
		return generatedFFIMixedDecimalShape{"decimal64", "fma", 64, []int{64, 128, 64}}, true
	case "bid64dqq_fma":
		return generatedFFIMixedDecimalShape{"decimal64", "fma", 64, []int{64, 128, 128}}, true
	case "bid64qdd_fma":
		return generatedFFIMixedDecimalShape{"decimal64", "fma", 64, []int{128, 64, 64}}, true
	case "bid64qdq_fma":
		return generatedFFIMixedDecimalShape{"decimal64", "fma", 64, []int{128, 64, 128}}, true
	case "bid64qqd_fma":
		return generatedFFIMixedDecimalShape{"decimal64", "fma", 64, []int{128, 128, 64}}, true
	case "bid64qqq_fma":
		return generatedFFIMixedDecimalShape{"decimal64", "fma", 64, []int{128, 128, 128}}, true
	case "bid128ddd_fma":
		return generatedFFIMixedDecimalShape{"decimal128", "fma", 128, []int{64, 64, 64}}, true
	case "bid128ddq_fma":
		return generatedFFIMixedDecimalShape{"decimal128", "fma", 128, []int{64, 64, 128}}, true
	case "bid128dqd_fma":
		return generatedFFIMixedDecimalShape{"decimal128", "fma", 128, []int{64, 128, 64}}, true
	case "bid128dqq_fma":
		return generatedFFIMixedDecimalShape{"decimal128", "fma", 128, []int{64, 128, 128}}, true
	case "bid128qdd_fma":
		return generatedFFIMixedDecimalShape{"decimal128", "fma", 128, []int{128, 64, 64}}, true
	case "bid128qdq_fma":
		return generatedFFIMixedDecimalShape{"decimal128", "fma", 128, []int{128, 64, 128}}, true
	case "bid128qqd_fma":
		return generatedFFIMixedDecimalShape{"decimal128", "fma", 128, []int{128, 128, 64}}, true
	case "bid64q_sqrt":
		return generatedFFIMixedDecimalShape{"decimal64", "sqrt", 64, []int{128}}, true
	case "bid128d_sqrt":
		return generatedFFIMixedDecimalShape{"decimal128", "sqrt", 128, []int{64}}, true
	default:
		return generatedFFIMixedDecimalShape{}, false
	}
}

func generatedFFIMixedDecimalFunction(function string) bool {
	_, ok := generatedFFIMixedDecimalShapeFor(function)
	return ok
}

func parseGeneratedFFIMixedDecimalOperands(tc generatedFFICase) (generatedFFIMixedDecimalOperands, error) {
	shape, ok := generatedFFIMixedDecimalShapeFor(tc.Function)
	if !ok {
		return generatedFFIMixedDecimalOperands{}, fmt.Errorf("unsupported mixed decimal ffi function %q", tc.Function)
	}
	if tc.Format != shape.format || tc.Operation != shape.operation || tc.ResultBits != shape.resultBits || !equalGeneratedFFIIntSlices(tc.OperandBits, shape.operandBits) {
		return generatedFFIMixedDecimalOperands{}, fmt.Errorf("%s mixed decimal shape = (%s, %s, %d, %v), want (%s, %s, %d, %v)", tc.Function, tc.Format, tc.Operation, tc.ResultBits, tc.OperandBits, shape.format, shape.operation, shape.resultBits, shape.operandBits)
	}
	if len(tc.Operands) != len(shape.operandBits) {
		return generatedFFIMixedDecimalOperands{}, fmt.Errorf("%s expects %d operands, got %d", tc.Function, len(shape.operandBits), len(tc.Operands))
	}

	var operands generatedFFIMixedDecimalOperands
	for i, bits := range shape.operandBits {
		switch bits {
		case 64:
			value, err := strconv.ParseUint(tc.Operands[i], 16, 64)
			if err != nil {
				return generatedFFIMixedDecimalOperands{}, fmt.Errorf("parse operand %d as BID64 %q: %w", i, tc.Operands[i], err)
			}
			operands.narrow[i] = value
		case 128:
			value, err := parseFFIUint128Bits(tc.Operands[i])
			if err != nil {
				return generatedFFIMixedDecimalOperands{}, fmt.Errorf("parse operand %d as BID128 %q: %w", i, tc.Operands[i], err)
			}
			operands.wide[i] = value
		default:
			return generatedFFIMixedDecimalOperands{}, fmt.Errorf("%s operand %d has unsupported width %d", tc.Function, i, bits)
		}
	}
	return operands, nil
}

func equalGeneratedFFIIntSlices(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func runGeneratedFFICaseMixedDecimal(tc generatedFFICase) (string, string, error) {
	op, err := parseGeneratedFFIMixedDecimalOperands(tc)
	if err != nil {
		return "", "", err
	}
	rounding := generatedFFICRoundingMode(tc.Rounding)
	var flags C._IDEC_flags

	switch tc.Function {
	case "bid64ddq_fma":
		native := uint64(C.bid64ddq_fma(C.BID_UINT64(op.narrow[0]), C.BID_UINT64(op.narrow[1]), ffiUint128ToC(op.wide[2]), rounding, &flags))
		exposed, exposedFlags := bidgo.Bid64ddqFma(op.narrow[0], op.narrow[1], decimal128BIDAsBidgo(op.wide[2]), tc.Rounding)
		return fmt.Sprintf("%016x/%08x", native, uint32(flags)), fmt.Sprintf("%016x/%08x", exposed, exposedFlags), nil
	case "bid64dqd_fma":
		native := uint64(C.bid64dqd_fma(C.BID_UINT64(op.narrow[0]), ffiUint128ToC(op.wide[1]), C.BID_UINT64(op.narrow[2]), rounding, &flags))
		exposed, exposedFlags := bidgo.Bid64dqdFma(op.narrow[0], decimal128BIDAsBidgo(op.wide[1]), op.narrow[2], tc.Rounding)
		return fmt.Sprintf("%016x/%08x", native, uint32(flags)), fmt.Sprintf("%016x/%08x", exposed, exposedFlags), nil
	case "bid64dqq_fma":
		native := uint64(C.bid64dqq_fma(C.BID_UINT64(op.narrow[0]), ffiUint128ToC(op.wide[1]), ffiUint128ToC(op.wide[2]), rounding, &flags))
		exposed, exposedFlags := bidgo.Bid64dqqFma(op.narrow[0], decimal128BIDAsBidgo(op.wide[1]), decimal128BIDAsBidgo(op.wide[2]), tc.Rounding)
		return fmt.Sprintf("%016x/%08x", native, uint32(flags)), fmt.Sprintf("%016x/%08x", exposed, exposedFlags), nil
	case "bid64qdd_fma":
		native := uint64(C.bid64qdd_fma(ffiUint128ToC(op.wide[0]), C.BID_UINT64(op.narrow[1]), C.BID_UINT64(op.narrow[2]), rounding, &flags))
		exposed, exposedFlags := bidgo.Bid64qddFma(decimal128BIDAsBidgo(op.wide[0]), op.narrow[1], op.narrow[2], tc.Rounding)
		return fmt.Sprintf("%016x/%08x", native, uint32(flags)), fmt.Sprintf("%016x/%08x", exposed, exposedFlags), nil
	case "bid64qdq_fma":
		native := uint64(C.bid64qdq_fma(ffiUint128ToC(op.wide[0]), C.BID_UINT64(op.narrow[1]), ffiUint128ToC(op.wide[2]), rounding, &flags))
		exposed, exposedFlags := bidgo.Bid64qdqFma(decimal128BIDAsBidgo(op.wide[0]), op.narrow[1], decimal128BIDAsBidgo(op.wide[2]), tc.Rounding)
		return fmt.Sprintf("%016x/%08x", native, uint32(flags)), fmt.Sprintf("%016x/%08x", exposed, exposedFlags), nil
	case "bid64qqd_fma":
		native := uint64(C.bid64qqd_fma(ffiUint128ToC(op.wide[0]), ffiUint128ToC(op.wide[1]), C.BID_UINT64(op.narrow[2]), rounding, &flags))
		exposed, exposedFlags := bidgo.Bid64qqdFma(decimal128BIDAsBidgo(op.wide[0]), decimal128BIDAsBidgo(op.wide[1]), op.narrow[2], tc.Rounding)
		return fmt.Sprintf("%016x/%08x", native, uint32(flags)), fmt.Sprintf("%016x/%08x", exposed, exposedFlags), nil
	case "bid64qqq_fma":
		native := uint64(C.bid64qqq_fma(ffiUint128ToC(op.wide[0]), ffiUint128ToC(op.wide[1]), ffiUint128ToC(op.wide[2]), rounding, &flags))
		exposed, exposedFlags := bidgo.Bid64qqqFma(decimal128BIDAsBidgo(op.wide[0]), decimal128BIDAsBidgo(op.wide[1]), decimal128BIDAsBidgo(op.wide[2]), tc.Rounding)
		return fmt.Sprintf("%016x/%08x", native, uint32(flags)), fmt.Sprintf("%016x/%08x", exposed, exposedFlags), nil
	case "bid128ddd_fma":
		native := ffiUint128FromC(C.bid128ddd_fma(C.BID_UINT64(op.narrow[0]), C.BID_UINT64(op.narrow[1]), C.BID_UINT64(op.narrow[2]), rounding, &flags))
		exposed, exposedFlags := bidgo.Bid128dddFma(op.narrow[0], op.narrow[1], op.narrow[2], tc.Rounding)
		return fmt.Sprintf("%s/%08x", formatFFIUint128Bits(native), uint32(flags)), fmt.Sprintf("%s/%08x", formatFFIUint128Bits(decimal128BIDFromBidgo(exposed)), exposedFlags), nil
	case "bid128ddq_fma":
		native := ffiUint128FromC(C.bid128ddq_fma(C.BID_UINT64(op.narrow[0]), C.BID_UINT64(op.narrow[1]), ffiUint128ToC(op.wide[2]), rounding, &flags))
		exposed, exposedFlags := bidgo.Bid128ddqFma(op.narrow[0], op.narrow[1], decimal128BIDAsBidgo(op.wide[2]), tc.Rounding)
		return fmt.Sprintf("%s/%08x", formatFFIUint128Bits(native), uint32(flags)), fmt.Sprintf("%s/%08x", formatFFIUint128Bits(decimal128BIDFromBidgo(exposed)), exposedFlags), nil
	case "bid128dqd_fma":
		native := ffiUint128FromC(C.bid128dqd_fma(C.BID_UINT64(op.narrow[0]), ffiUint128ToC(op.wide[1]), C.BID_UINT64(op.narrow[2]), rounding, &flags))
		exposed, exposedFlags := bidgo.Bid128dqdFma(op.narrow[0], decimal128BIDAsBidgo(op.wide[1]), op.narrow[2], tc.Rounding)
		return fmt.Sprintf("%s/%08x", formatFFIUint128Bits(native), uint32(flags)), fmt.Sprintf("%s/%08x", formatFFIUint128Bits(decimal128BIDFromBidgo(exposed)), exposedFlags), nil
	case "bid128dqq_fma":
		native := ffiUint128FromC(C.bid128dqq_fma(C.BID_UINT64(op.narrow[0]), ffiUint128ToC(op.wide[1]), ffiUint128ToC(op.wide[2]), rounding, &flags))
		exposed, exposedFlags := bidgo.Bid128dqqFma(op.narrow[0], decimal128BIDAsBidgo(op.wide[1]), decimal128BIDAsBidgo(op.wide[2]), tc.Rounding)
		return fmt.Sprintf("%s/%08x", formatFFIUint128Bits(native), uint32(flags)), fmt.Sprintf("%s/%08x", formatFFIUint128Bits(decimal128BIDFromBidgo(exposed)), exposedFlags), nil
	case "bid128qdd_fma":
		native := ffiUint128FromC(C.bid128qdd_fma(ffiUint128ToC(op.wide[0]), C.BID_UINT64(op.narrow[1]), C.BID_UINT64(op.narrow[2]), rounding, &flags))
		exposed, exposedFlags := bidgo.Bid128qddFma(decimal128BIDAsBidgo(op.wide[0]), op.narrow[1], op.narrow[2], tc.Rounding)
		return fmt.Sprintf("%s/%08x", formatFFIUint128Bits(native), uint32(flags)), fmt.Sprintf("%s/%08x", formatFFIUint128Bits(decimal128BIDFromBidgo(exposed)), exposedFlags), nil
	case "bid128qdq_fma":
		native := ffiUint128FromC(C.bid128qdq_fma(ffiUint128ToC(op.wide[0]), C.BID_UINT64(op.narrow[1]), ffiUint128ToC(op.wide[2]), rounding, &flags))
		exposed, exposedFlags := bidgo.Bid128qdqFma(decimal128BIDAsBidgo(op.wide[0]), op.narrow[1], decimal128BIDAsBidgo(op.wide[2]), tc.Rounding)
		return fmt.Sprintf("%s/%08x", formatFFIUint128Bits(native), uint32(flags)), fmt.Sprintf("%s/%08x", formatFFIUint128Bits(decimal128BIDFromBidgo(exposed)), exposedFlags), nil
	case "bid128qqd_fma":
		native := ffiUint128FromC(C.bid128qqd_fma(ffiUint128ToC(op.wide[0]), ffiUint128ToC(op.wide[1]), C.BID_UINT64(op.narrow[2]), rounding, &flags))
		exposed, exposedFlags := bidgo.Bid128qqdFma(decimal128BIDAsBidgo(op.wide[0]), decimal128BIDAsBidgo(op.wide[1]), op.narrow[2], tc.Rounding)
		return fmt.Sprintf("%s/%08x", formatFFIUint128Bits(native), uint32(flags)), fmt.Sprintf("%s/%08x", formatFFIUint128Bits(decimal128BIDFromBidgo(exposed)), exposedFlags), nil
	case "bid64q_sqrt":
		native := uint64(C.bid64q_sqrt(ffiUint128ToC(op.wide[0]), rounding, &flags))
		exposed, exposedFlags := bidgo.Bid64qSqrt(decimal128BIDAsBidgo(op.wide[0]), tc.Rounding)
		return fmt.Sprintf("%016x/%08x", native, uint32(flags)), fmt.Sprintf("%016x/%08x", exposed, exposedFlags), nil
	case "bid128d_sqrt":
		native := ffiUint128FromC(C.bid128d_sqrt(C.BID_UINT64(op.narrow[0]), rounding, &flags))
		exposed, exposedFlags := bidgo.Bid128dSqrt(op.narrow[0], tc.Rounding)
		return fmt.Sprintf("%s/%08x", formatFFIUint128Bits(native), uint32(flags)), fmt.Sprintf("%s/%08x", formatFFIUint128Bits(decimal128BIDFromBidgo(exposed)), exposedFlags), nil
	default:
		return "", "", fmt.Errorf("unsupported mixed decimal ffi function %q", tc.Function)
	}
}

func generatedFFIWidenMixedDecimalOperand(op generatedFFIMixedDecimalOperands, width, index int, flags *C._IDEC_flags) (C.BID_UINT128, error) {
	switch width {
	case 64:
		return C.bid64_to_bid128(C.BID_UINT64(op.narrow[index]), flags), nil
	case 128:
		return ffiUint128ToC(op.wide[index]), nil
	default:
		var zero C.BID_UINT128
		return zero, fmt.Errorf("mixed FMA operand %d has unsupported width %d", index, width)
	}
}

// runGeneratedFFIMixedFMAComposed intentionally evaluates the forbidden
// non-fused predecessor sequence. All conversions, multiply, add, and final
// narrowing share one status word so sticky Intel BID flags are observable.
func runGeneratedFFIMixedFMAComposed(tc generatedFFICase) (string, error) {
	shape, ok := generatedFFIMixedDecimalShapeFor(tc.Function)
	if !ok || shape.operation != "fma" || len(shape.operandBits) != 3 {
		return "", fmt.Errorf("unsupported mixed FMA composition %q", tc.Function)
	}
	op, err := parseGeneratedFFIMixedDecimalOperands(tc)
	if err != nil {
		return "", err
	}
	rounding := generatedFFICRoundingMode(tc.Rounding)
	var flags C._IDEC_flags
	operands := [3]C.BID_UINT128{}
	for i, width := range shape.operandBits {
		operands[i], err = generatedFFIWidenMixedDecimalOperand(op, width, i, &flags)
		if err != nil {
			return "", err
		}
	}
	product := C.bid128_mul(operands[0], operands[1], rounding, &flags)
	sum := C.bid128_add(product, operands[2], rounding, &flags)
	switch shape.resultBits {
	case 64:
		result := uint64(C.bid128_to_bid64(sum, rounding, &flags))
		return fmt.Sprintf("%016x/%08x", result, uint32(flags)), nil
	case 128:
		result := ffiUint128FromC(sum)
		return fmt.Sprintf("%s/%08x", formatFFIUint128Bits(result), uint32(flags)), nil
	default:
		return "", fmt.Errorf("mixed FMA %s has unsupported result width %d", tc.Function, shape.resultBits)
	}
}

func generatedFFIIntUnaryOperation(operation string) bool {
	switch operation {
	case "class", "isSigned", "isNormal", "isSubnormal", "isFinite", "isZero", "isInf", "isNaN", "isSignaling", "isCanonical", "radix", "quantexp", "ilogb":
		return true
	default:
		return false
	}
}

func generatedFFIInt64UnaryOperation(operation string) bool {
	return operation == "llquantexp"
}

func generatedFFIIntBinaryOperation(operation string) bool {
	switch operation {
	case "totalOrder", "totalOrderMag", "sameQuantum",
		"quiet_equal", "quiet_greater", "quiet_greater_equal", "quiet_greater_unordered",
		"quiet_less", "quiet_less_equal", "quiet_less_unordered", "quiet_not_equal",
		"quiet_not_greater", "quiet_not_less", "quiet_ordered", "quiet_unordered",
		"signaling_greater", "signaling_greater_equal", "signaling_greater_unordered",
		"signaling_less", "signaling_less_equal", "signaling_less_unordered",
		"signaling_not_greater", "signaling_not_less":
		return true
	default:
		return false
	}
}

func generatedFFIDecimalFlagUnaryOperation(operation string) bool {
	return operation == "quantum"
}

func generatedFFIDecimalFlagIntOperation(operation string) bool {
	return operation == "ldexp" || operation == "scalbln"
}

func formatGeneratedFFIIntResult(value int, flags uint32) string {
	return fmt.Sprintf("%d/%08x", value, flags)
}

func formatGeneratedFFIInt64Result(value int64, flags uint32) string {
	return fmt.Sprintf("%d/%08x", value, flags)
}

func generatedFFIBaseIntegerFromOperation(operation string) bool {
	switch operation {
	case "from_int32", "from_int64", "from_uint32", "from_uint64":
		return true
	default:
		return false
	}
}

func generatedFFIBIDWidthConversionOperation(operation string) bool {
	switch operation {
	case "to_bid32", "to_bid64", "to_bid128":
		return true
	default:
		return false
	}
}

func generatedFFIBinaryConversionOperation(operation string) bool {
	switch operation {
	case "to_binary32", "to_binary64", "to_binary128":
		return true
	default:
		return false
	}
}

func generatedFFIBaseIntegerToOperation(operation string) bool {
	kind, _, ok := parseGeneratedFFIBaseIntegerToOperation(operation)
	return ok && (kind == "int8" || kind == "int16" || kind == "int32" || kind == "int64" || kind == "uint8" || kind == "uint16" || kind == "uint32" || kind == "uint64")
}

func parseGeneratedFFIBaseIntegerToOperation(operation string) (kind string, mode string, ok bool) {
	parts := strings.Split(operation, "_")
	if len(parts) != 3 || parts[0] != "to" {
		return "", "", false
	}
	switch parts[1] {
	case "int8", "int16", "int32", "int64", "uint8", "uint16", "uint32", "uint64":
	default:
		return "", "", false
	}
	switch parts[2] {
	case "ceil", "floor", "int", "rnint", "rninta", "xceil", "xfloor", "xint", "xrnint", "xrninta":
	default:
		return "", "", false
	}
	return parts[1], parts[2], true
}

func runGeneratedFFICaseNativeBaseIntegerTo(tc generatedFFICase) (string, error) {
	kind, _, ok := parseGeneratedFFIBaseIntegerToOperation(tc.Operation)
	if !ok {
		return "", fmt.Errorf("unsupported integer conversion operation %q", tc.Operation)
	}
	operands := generatedFFIReadtestOperands(tc)
	if strings.HasPrefix(kind, "uint") {
		value, status, err := nativeReadtestGeneratedUnsigned(tc.Function, 0, operands)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d/%s", value, status), nil
	}
	value, status, err := nativeReadtestGeneratedSigned(tc.Function, 0, operands)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d/%s", value, status), nil
}

func generatedFFIReadtestOperands(tc generatedFFICase) []string {
	out := make([]string, len(tc.Operands))
	for i, operand := range tc.Operands {
		switch tc.Format {
		case "decimal32":
			out[i] = "[" + strings.ToUpper(operand) + "]"
		case "decimal64":
			out[i] = "[" + strings.ToUpper(operand) + "]"
		case "decimal128":
			out[i] = "[" + strings.ToUpper(operand) + "]"
		default:
			out[i] = operand
		}
	}
	return out
}

func runGeneratedFFICaseBaseIntegerFrom(tc generatedFFICase) (string, string, error) {
	switch tc.Function {
	case "bid32_from_int32":
		x, err := parseFFIInt32Operand(tc)
		if err != nil {
			return "", "", err
		}
		var nativeFlags C._IDEC_flags
		native := uint32(C.bid32_from_int32(C.int(x), generatedFFICRoundingMode(tc.Rounding), &nativeFlags))
		exposed, exposedFlags := bidgo.Bid32FromInt32(x, tc.Rounding)
		return fmt.Sprintf("%08x/%08x", native, uint32(nativeFlags)), fmt.Sprintf("%08x/%08x", exposed, exposedFlags), nil
	case "bid32_from_int64":
		x, err := parseFFIInt64Operand(tc)
		if err != nil {
			return "", "", err
		}
		var nativeFlags C._IDEC_flags
		native := uint32(C.bid32_from_int64(C.BID_SINT64(x), generatedFFICRoundingMode(tc.Rounding), &nativeFlags))
		exposed, exposedFlags := bidgo.Bid32FromInt64(x, tc.Rounding)
		return fmt.Sprintf("%08x/%08x", native, uint32(nativeFlags)), fmt.Sprintf("%08x/%08x", exposed, exposedFlags), nil
	case "bid32_from_uint32":
		x, err := parseFFIUint32DecimalOperand(tc)
		if err != nil {
			return "", "", err
		}
		var nativeFlags C._IDEC_flags
		native := uint32(C.bid32_from_uint32(C.uint(x), generatedFFICRoundingMode(tc.Rounding), &nativeFlags))
		exposed, exposedFlags := bidgo.Bid32FromUint32(x, tc.Rounding)
		return fmt.Sprintf("%08x/%08x", native, uint32(nativeFlags)), fmt.Sprintf("%08x/%08x", exposed, exposedFlags), nil
	case "bid32_from_uint64":
		x, err := parseFFIUint64DecimalOperand(tc)
		if err != nil {
			return "", "", err
		}
		var nativeFlags C._IDEC_flags
		native := uint32(C.bid32_from_uint64(C.BID_UINT64(x), generatedFFICRoundingMode(tc.Rounding), &nativeFlags))
		exposed, exposedFlags := bidgo.Bid32FromUint64(x, tc.Rounding)
		return fmt.Sprintf("%08x/%08x", native, uint32(nativeFlags)), fmt.Sprintf("%08x/%08x", exposed, exposedFlags), nil
	case "bid64_from_int32":
		x, err := parseFFIInt32Operand(tc)
		if err != nil {
			return "", "", err
		}
		native := uint64(C.bid64_from_int32(C.int(x)))
		exposed := bidgo.Bid64FromInt32(x)
		return fmt.Sprintf("%016x", native), fmt.Sprintf("%016x", exposed), nil
	case "bid64_from_int64":
		x, err := parseFFIInt64Operand(tc)
		if err != nil {
			return "", "", err
		}
		var nativeFlags C._IDEC_flags
		native := uint64(C.bid64_from_int64(C.BID_SINT64(x), generatedFFICRoundingMode(tc.Rounding), &nativeFlags))
		exposed, exposedFlags := bidgo.Bid64FromInt64(x, tc.Rounding)
		return fmt.Sprintf("%016x/%08x", native, uint32(nativeFlags)), fmt.Sprintf("%016x/%08x", exposed, exposedFlags), nil
	case "bid64_from_uint32":
		x, err := parseFFIUint32DecimalOperand(tc)
		if err != nil {
			return "", "", err
		}
		native := uint64(C.bid64_from_uint32(C.uint(x)))
		exposed := bidgo.Bid64FromUint32(x)
		return fmt.Sprintf("%016x", native), fmt.Sprintf("%016x", exposed), nil
	case "bid64_from_uint64":
		x, err := parseFFIUint64DecimalOperand(tc)
		if err != nil {
			return "", "", err
		}
		var nativeFlags C._IDEC_flags
		native := uint64(C.bid64_from_uint64(C.BID_UINT64(x), generatedFFICRoundingMode(tc.Rounding), &nativeFlags))
		exposed, exposedFlags := bidgo.Bid64FromUint64(x, tc.Rounding)
		return fmt.Sprintf("%016x/%08x", native, uint32(nativeFlags)), fmt.Sprintf("%016x/%08x", exposed, exposedFlags), nil
	case "bid128_from_int32":
		x, err := parseFFIInt32Operand(tc)
		if err != nil {
			return "", "", err
		}
		native := ffiUint128FromC(C.bid128_from_int32(C.int(x)))
		exposed := decimal128BIDFromBidgo(bidgo.Bid128FromInt32(x))
		return formatFFIUint128Bits(native), formatFFIUint128Bits(exposed), nil
	case "bid128_from_int64":
		x, err := parseFFIInt64Operand(tc)
		if err != nil {
			return "", "", err
		}
		native := ffiUint128FromC(C.bid128_from_int64(C.BID_SINT64(x)))
		exposed := decimal128BIDFromBidgo(bidgo.Bid128FromInt64(x))
		return formatFFIUint128Bits(native), formatFFIUint128Bits(exposed), nil
	case "bid128_from_uint32":
		x, err := parseFFIUint32DecimalOperand(tc)
		if err != nil {
			return "", "", err
		}
		native := ffiUint128FromC(C.bid128_from_uint32(C.uint(x)))
		exposed := decimal128BIDFromBidgo(bidgo.Bid128FromUint32(x))
		return formatFFIUint128Bits(native), formatFFIUint128Bits(exposed), nil
	case "bid128_from_uint64":
		x, err := parseFFIUint64DecimalOperand(tc)
		if err != nil {
			return "", "", err
		}
		native := ffiUint128FromC(C.bid128_from_uint64(C.BID_UINT64(x)))
		exposed := decimal128BIDFromBidgo(bidgo.Bid128FromUint64(x))
		return formatFFIUint128Bits(native), formatFFIUint128Bits(exposed), nil
	default:
		return "", "", fmt.Errorf("unsupported generated ffi integer constructor function %q", tc.Function)
	}
}

func runGeneratedFFICaseBIDWidthConversion(tc generatedFFICase) (string, string, error) {
	switch tc.Function {
	case "bid32_to_bid64":
		a, err := parseFFIUint32UnaryOperand(tc)
		if err != nil {
			return "", "", err
		}
		var nativeFlags C._IDEC_flags
		native := uint64(C.bid32_to_bid64(C.BID_UINT32(a), &nativeFlags))
		exposed, exposedFlags := bidgo.Bid32ToBid64(a)
		return fmt.Sprintf("%016x/%08x", native, uint32(nativeFlags)), fmt.Sprintf("%016x/%08x", exposed, exposedFlags), nil
	case "bid32_to_bid128":
		a, err := parseFFIUint32UnaryOperand(tc)
		if err != nil {
			return "", "", err
		}
		var nativeFlags C._IDEC_flags
		native := ffiUint128FromC(C.bid32_to_bid128(C.BID_UINT32(a), &nativeFlags))
		exposed, exposedFlags := bidgo.Bid32ToBid128(a)
		return fmt.Sprintf("%s/%08x", formatFFIUint128Bits(native), uint32(nativeFlags)), fmt.Sprintf("%s/%08x", formatFFIUint128Bits(decimal128BIDFromBidgo(exposed)), exposedFlags), nil
	case "bid64_to_bid32":
		a, err := parseFFIUint64UnaryOperand(tc)
		if err != nil {
			return "", "", err
		}
		var nativeFlags C._IDEC_flags
		native := uint32(C.bid64_to_bid32(C.BID_UINT64(a), generatedFFICRoundingMode(tc.Rounding), &nativeFlags))
		exposed, exposedFlags := bidgo.Bid64ToBid32(a, tc.Rounding)
		return fmt.Sprintf("%08x/%08x", native, uint32(nativeFlags)), fmt.Sprintf("%08x/%08x", exposed, exposedFlags), nil
	case "bid64_to_bid128":
		a, err := parseFFIUint64UnaryOperand(tc)
		if err != nil {
			return "", "", err
		}
		var nativeFlags C._IDEC_flags
		native := ffiUint128FromC(C.bid64_to_bid128(C.BID_UINT64(a), &nativeFlags))
		exposed, exposedFlags := bidgo.Bid64ToBid128(a)
		return fmt.Sprintf("%s/%08x", formatFFIUint128Bits(native), uint32(nativeFlags)), fmt.Sprintf("%s/%08x", formatFFIUint128Bits(decimal128BIDFromBidgo(exposed)), exposedFlags), nil
	case "bid128_to_bid32":
		a, err := parseFFIUint128UnaryOperand(tc)
		if err != nil {
			return "", "", err
		}
		var nativeFlags C._IDEC_flags
		native := uint32(C.bid128_to_bid32(ffiUint128ToC(a), generatedFFICRoundingMode(tc.Rounding), &nativeFlags))
		exposed, exposedFlags := bidgo.Bid128ToBid32(decimal128BIDAsBidgo(a), tc.Rounding)
		return fmt.Sprintf("%08x/%08x", native, uint32(nativeFlags)), fmt.Sprintf("%08x/%08x", exposed, exposedFlags), nil
	case "bid128_to_bid64":
		a, err := parseFFIUint128UnaryOperand(tc)
		if err != nil {
			return "", "", err
		}
		var nativeFlags C._IDEC_flags
		native := uint64(C.bid128_to_bid64(ffiUint128ToC(a), generatedFFICRoundingMode(tc.Rounding), &nativeFlags))
		exposed, exposedFlags := bidgo.Bid128ToBid64(decimal128BIDAsBidgo(a), tc.Rounding)
		return fmt.Sprintf("%016x/%08x", native, uint32(nativeFlags)), fmt.Sprintf("%016x/%08x", exposed, exposedFlags), nil
	default:
		return "", "", fmt.Errorf("unsupported generated ffi BID width conversion function %q", tc.Function)
	}
}

func runGeneratedFFICaseBinaryConversion(tc generatedFFICase) (string, string, error) {
	switch tc.Function {
	case "bid32_to_binary32":
		a, err := parseFFIUint32UnaryOperand(tc)
		if err != nil {
			return "", "", err
		}
		var nativeFlags C._IDEC_flags
		native := math.Float32bits(float32(C.bid32_to_binary32(C.BID_UINT32(a), generatedFFICRoundingMode(tc.Rounding), &nativeFlags)))
		exposed, exposedFlags := bidgo.Bid32ToBinary32(a, tc.Rounding)
		return fmt.Sprintf("%08x/%08x", native, uint32(nativeFlags)), fmt.Sprintf("%08x/%08x", exposed, exposedFlags), nil
	case "bid32_to_binary64":
		a, err := parseFFIUint32UnaryOperand(tc)
		if err != nil {
			return "", "", err
		}
		var nativeFlags C._IDEC_flags
		native := math.Float64bits(float64(C.bid32_to_binary64(C.BID_UINT32(a), generatedFFICRoundingMode(tc.Rounding), &nativeFlags)))
		exposed, exposedFlags := bidgo.Bid32ToBinary64(a, tc.Rounding)
		return fmt.Sprintf("%016x/%08x", native, uint32(nativeFlags)), fmt.Sprintf("%016x/%08x", exposed, exposedFlags), nil
	case "bid32_to_binary128":
		a, err := parseFFIUint32UnaryOperand(tc)
		if err != nil {
			return "", "", err
		}
		var nativeFlags C._IDEC_flags
		native := ffiUint128FromC(C.bid32_to_binary128(C.BID_UINT32(a), generatedFFICRoundingMode(tc.Rounding), &nativeFlags))
		exposed, exposedFlags := bidgo.Bid32ToBinary128(a, tc.Rounding)
		return fmt.Sprintf("%s/%08x", formatFFIUint128Bits(native), uint32(nativeFlags)), fmt.Sprintf("%s/%08x", formatFFIUint128Bits(decimal128BIDFromBidgo(exposed)), exposedFlags), nil
	case "bid64_to_binary32":
		a, err := parseFFIUint64UnaryOperand(tc)
		if err != nil {
			return "", "", err
		}
		var nativeFlags C._IDEC_flags
		native := math.Float32bits(float32(C.bid64_to_binary32(C.BID_UINT64(a), generatedFFICRoundingMode(tc.Rounding), &nativeFlags)))
		exposed, exposedFlags := bidgo.Bid64ToBinary32(a, tc.Rounding)
		return fmt.Sprintf("%08x/%08x", native, uint32(nativeFlags)), fmt.Sprintf("%08x/%08x", exposed, exposedFlags), nil
	case "bid64_to_binary64":
		a, err := parseFFIUint64UnaryOperand(tc)
		if err != nil {
			return "", "", err
		}
		var nativeFlags C._IDEC_flags
		native := math.Float64bits(float64(C.bid64_to_binary64(C.BID_UINT64(a), generatedFFICRoundingMode(tc.Rounding), &nativeFlags)))
		exposed, exposedFlags := bidgo.Bid64ToBinary64(a, tc.Rounding)
		return fmt.Sprintf("%016x/%08x", native, uint32(nativeFlags)), fmt.Sprintf("%016x/%08x", exposed, exposedFlags), nil
	case "bid64_to_binary128":
		a, err := parseFFIUint64UnaryOperand(tc)
		if err != nil {
			return "", "", err
		}
		var nativeFlags C._IDEC_flags
		native := ffiUint128FromC(C.bid64_to_binary128(C.BID_UINT64(a), generatedFFICRoundingMode(tc.Rounding), &nativeFlags))
		exposed, exposedFlags := bidgo.Bid64ToBinary128(a, tc.Rounding)
		return fmt.Sprintf("%s/%08x", formatFFIUint128Bits(native), uint32(nativeFlags)), fmt.Sprintf("%s/%08x", formatFFIUint128Bits(decimal128BIDFromBidgo(exposed)), exposedFlags), nil
	case "bid128_to_binary32":
		a, err := parseFFIUint128UnaryOperand(tc)
		if err != nil {
			return "", "", err
		}
		var nativeFlags C._IDEC_flags
		native := math.Float32bits(float32(C.bid128_to_binary32(ffiUint128ToC(a), generatedFFICRoundingMode(tc.Rounding), &nativeFlags)))
		var exposedFlags uint32
		exposed := math.Float32bits(bidgo.Bid128ToBinary32(decimal128BIDAsBidgo(a), tc.Rounding, &exposedFlags))
		return fmt.Sprintf("%08x/%08x", native, uint32(nativeFlags)), fmt.Sprintf("%08x/%08x", exposed, exposedFlags), nil
	case "bid128_to_binary64":
		a, err := parseFFIUint128UnaryOperand(tc)
		if err != nil {
			return "", "", err
		}
		var nativeFlags C._IDEC_flags
		native := math.Float64bits(float64(C.bid128_to_binary64(ffiUint128ToC(a), generatedFFICRoundingMode(tc.Rounding), &nativeFlags)))
		var exposedFlags uint32
		exposed := math.Float64bits(bidgo.Bid128ToBinary64(decimal128BIDAsBidgo(a), tc.Rounding, &exposedFlags))
		return fmt.Sprintf("%016x/%08x", native, uint32(nativeFlags)), fmt.Sprintf("%016x/%08x", exposed, exposedFlags), nil
	case "bid128_to_binary128":
		a, err := parseFFIUint128UnaryOperand(tc)
		if err != nil {
			return "", "", err
		}
		var nativeFlags C._IDEC_flags
		native := ffiUint128FromC(C.bid128_to_binary128(ffiUint128ToC(a), generatedFFICRoundingMode(tc.Rounding), &nativeFlags))
		exposed, exposedFlags := bidgo.Bid128ToBinary128(decimal128BIDAsBidgo(a), tc.Rounding)
		return fmt.Sprintf("%s/%08x", formatFFIUint128Bits(native), uint32(nativeFlags)), fmt.Sprintf("%s/%08x", formatFFIUint128Bits(decimal128BIDFromBidgo(exposed)), exposedFlags), nil
	default:
		return "", "", fmt.Errorf("unsupported generated ffi binary conversion function %q", tc.Function)
	}
}

func runGeneratedFFICaseGoBaseIntegerTo(tc generatedFFICase) (string, error) {
	kind, mode, ok := parseGeneratedFFIBaseIntegerToOperation(tc.Operation)
	if !ok {
		return "", fmt.Errorf("unsupported integer conversion operation %q", tc.Operation)
	}
	if strings.HasPrefix(kind, "uint") {
		value, flags, err := runGeneratedFFICaseGoBaseUnsignedIntegerTo(tc, kind, mode)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d/%s", value, formatReadtestStatus(flags)), nil
	}
	value, flags, err := runGeneratedFFICaseGoBaseSignedIntegerTo(tc, kind, mode)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d/%s", value, formatReadtestStatus(flags)), nil
}

func runGeneratedFFICaseGoBaseSignedIntegerTo(tc generatedFFICase, kind, mode string) (int64, uint32, error) {
	switch tc.Format {
	case "decimal32":
		a, err := parseFFIUint32UnaryOperand(tc)
		if err != nil {
			return 0, 0, err
		}
		switch kind {
		case "int8":
			value, flags, err := bid32ToInt8ByMode(a, mode)
			return int64(value), flags, err
		case "int16":
			value, flags, err := bid32ToInt16ByMode(a, mode)
			return int64(value), flags, err
		case "int32":
			value, flags, err := bid32ToInt32ByMode(a, mode)
			return int64(value), flags, err
		case "int64":
			return bid32ToInt64ByMode(a, mode)
		}
	case "decimal64":
		a, err := parseFFIUint64UnaryOperand(tc)
		if err != nil {
			return 0, 0, err
		}
		switch kind {
		case "int8":
			value, flags, err := bid64ToInt8ByMode(a, mode)
			return int64(value), flags, err
		case "int16":
			value, flags, err := bid64ToInt16ByMode(a, mode)
			return int64(value), flags, err
		case "int32":
			value, flags, err := bid64ToInt32ByMode(a, mode)
			return int64(value), flags, err
		case "int64":
			return bid64ToInt64ByMode(a, mode)
		}
	case "decimal128":
		a, err := parseFFIReadtestUint128UnaryOperand(tc)
		if err != nil {
			return 0, 0, err
		}
		ga := decimal128BIDAsBidgo(a)
		switch kind {
		case "int8":
			value, flags, err := bid128ToInt8ByMode(ga, mode)
			return int64(value), flags, err
		case "int16":
			value, flags, err := bid128ToInt16ByMode(ga, mode)
			return int64(value), flags, err
		case "int32":
			value, flags, err := bid128ToInt32ByMode(ga, mode)
			return int64(value), flags, err
		case "int64":
			return bid128ToInt64ByMode(ga, mode)
		}
	}
	return 0, 0, fmt.Errorf("unsupported signed integer conversion %s/%s/%s", tc.Format, kind, mode)
}

func runGeneratedFFICaseGoBaseUnsignedIntegerTo(tc generatedFFICase, kind, mode string) (uint64, uint32, error) {
	switch tc.Format {
	case "decimal32":
		a, err := parseFFIUint32UnaryOperand(tc)
		if err != nil {
			return 0, 0, err
		}
		switch kind {
		case "uint8":
			value, flags, err := bid32ToUint8ByMode(a, mode)
			return uint64(value), flags, err
		case "uint16":
			value, flags, err := bid32ToUint16ByMode(a, mode)
			return uint64(value), flags, err
		case "uint32":
			value, flags, err := bid32ToUint32ByMode(a, mode)
			return uint64(value), flags, err
		case "uint64":
			return bid32ToUint64ByMode(a, mode)
		}
	case "decimal64":
		a, err := parseFFIUint64UnaryOperand(tc)
		if err != nil {
			return 0, 0, err
		}
		switch kind {
		case "uint8":
			value, flags, err := bid64ToUint8ByMode(a, mode)
			return uint64(value), flags, err
		case "uint16":
			value, flags, err := bid64ToUint16ByMode(a, mode)
			return uint64(value), flags, err
		case "uint32":
			value, flags, err := bid64ToUint32ByMode(a, mode)
			return uint64(value), flags, err
		case "uint64":
			return bid64ToUint64ByMode(a, mode)
		}
	case "decimal128":
		a, err := parseFFIReadtestUint128UnaryOperand(tc)
		if err != nil {
			return 0, 0, err
		}
		ga := decimal128BIDAsBidgo(a)
		switch kind {
		case "uint8":
			value, flags, err := bid128ToUint8ByMode(ga, mode)
			return uint64(value), flags, err
		case "uint16":
			value, flags, err := bid128ToUint16ByMode(ga, mode)
			return uint64(value), flags, err
		case "uint32":
			value, flags, err := bid128ToUint32ByMode(ga, mode)
			return uint64(value), flags, err
		case "uint64":
			return bid128ToUint64ByMode(ga, mode)
		}
	}
	return 0, 0, fmt.Errorf("unsupported unsigned integer conversion %s/%s/%s", tc.Format, kind, mode)
}

type generatedFFIBaseInteger interface {
	~int8 | ~int16 | ~int32 | ~int64 | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

func callGeneratedFFIBaseIntegerToMode[X any, R generatedFFIBaseInteger](
	mode string,
	x X,
	ceil func(X) (R, uint32),
	floor func(X) (R, uint32),
	intMode func(X) (R, uint32),
	rnint func(X) (R, uint32),
	rninta func(X) (R, uint32),
	xceil func(X) (R, uint32),
	xfloor func(X) (R, uint32),
	xint func(X) (R, uint32),
	xrnint func(X) (R, uint32),
	xrninta func(X) (R, uint32),
	label string,
) (R, uint32, error) {
	switch mode {
	case "ceil":
		v, f := ceil(x)
		return v, f, nil
	case "floor":
		v, f := floor(x)
		return v, f, nil
	case "int":
		v, f := intMode(x)
		return v, f, nil
	case "rnint":
		v, f := rnint(x)
		return v, f, nil
	case "rninta":
		v, f := rninta(x)
		return v, f, nil
	case "xceil":
		v, f := xceil(x)
		return v, f, nil
	case "xfloor":
		v, f := xfloor(x)
		return v, f, nil
	case "xint":
		v, f := xint(x)
		return v, f, nil
	case "xrnint":
		v, f := xrnint(x)
		return v, f, nil
	case "xrninta":
		v, f := xrninta(x)
		return v, f, nil
	default:
		var zero R
		return zero, 0, fmt.Errorf("unsupported %s mode %q", label, mode)
	}
}

func bid32ToInt8ByMode(x uint32, mode string) (int8, uint32, error) {
	return callGeneratedFFIBaseIntegerToMode(mode, x, bidgo.Bid32ToInt8Ceil, bidgo.Bid32ToInt8Floor, bidgo.Bid32ToInt8Int, bidgo.Bid32ToInt8Rnint, bidgo.Bid32ToInt8Rninta, bidgo.Bid32ToInt8Xceil, bidgo.Bid32ToInt8Xfloor, bidgo.Bid32ToInt8Xint, bidgo.Bid32ToInt8Xrnint, bidgo.Bid32ToInt8Xrninta, "int8")
}

func bid32ToInt16ByMode(x uint32, mode string) (int16, uint32, error) {
	return callGeneratedFFIBaseIntegerToMode(mode, x, bidgo.Bid32ToInt16Ceil, bidgo.Bid32ToInt16Floor, bidgo.Bid32ToInt16Int, bidgo.Bid32ToInt16Rnint, bidgo.Bid32ToInt16Rninta, bidgo.Bid32ToInt16Xceil, bidgo.Bid32ToInt16Xfloor, bidgo.Bid32ToInt16Xint, bidgo.Bid32ToInt16Xrnint, bidgo.Bid32ToInt16Xrninta, "int16")
}

func bid32ToInt32ByMode(x uint32, mode string) (int32, uint32, error) {
	switch mode {
	case "ceil":
		v, f := bidgo.Bid32ToInt32Ceil(x)
		return v, f, nil
	case "floor":
		v, f := bidgo.Bid32ToInt32Floor(x)
		return v, f, nil
	case "int":
		v, f := bidgo.Bid32ToInt32Int(x)
		return v, f, nil
	case "rnint":
		v, f := bidgo.Bid32ToInt32Rnint(x)
		return v, f, nil
	case "rninta":
		v, f := bidgo.Bid32ToInt32Rninta(x)
		return v, f, nil
	case "xceil":
		v, f := bidgo.Bid32ToInt32Xceil(x)
		return v, f, nil
	case "xfloor":
		v, f := bidgo.Bid32ToInt32Xfloor(x)
		return v, f, nil
	case "xint":
		v, f := bidgo.Bid32ToInt32Xint(x)
		return v, f, nil
	case "xrnint":
		v, f := bidgo.Bid32ToInt32Xrnint(x)
		return v, f, nil
	case "xrninta":
		v, f := bidgo.Bid32ToInt32Xrninta(x)
		return v, f, nil
	default:
		return 0, 0, fmt.Errorf("unsupported int32 mode %q", mode)
	}
}

func bid32ToInt64ByMode(x uint32, mode string) (int64, uint32, error) {
	switch mode {
	case "ceil":
		v, f := bidgo.Bid32ToInt64Ceil(x)
		return v, f, nil
	case "floor":
		v, f := bidgo.Bid32ToInt64Floor(x)
		return v, f, nil
	case "int":
		v, f := bidgo.Bid32ToInt64Int(x)
		return v, f, nil
	case "rnint":
		v, f := bidgo.Bid32ToInt64Rnint(x)
		return v, f, nil
	case "rninta":
		v, f := bidgo.Bid32ToInt64Rninta(x)
		return v, f, nil
	case "xceil":
		v, f := bidgo.Bid32ToInt64Xceil(x)
		return v, f, nil
	case "xfloor":
		v, f := bidgo.Bid32ToInt64Xfloor(x)
		return v, f, nil
	case "xint":
		v, f := bidgo.Bid32ToInt64Xint(x)
		return v, f, nil
	case "xrnint":
		v, f := bidgo.Bid32ToInt64Xrnint(x)
		return v, f, nil
	case "xrninta":
		v, f := bidgo.Bid32ToInt64Xrninta(x)
		return v, f, nil
	default:
		return 0, 0, fmt.Errorf("unsupported int64 mode %q", mode)
	}
}

func bid32ToUint8ByMode(x uint32, mode string) (uint8, uint32, error) {
	return callGeneratedFFIBaseIntegerToMode(mode, x, bidgo.Bid32ToUint8Ceil, bidgo.Bid32ToUint8Floor, bidgo.Bid32ToUint8Int, bidgo.Bid32ToUint8Rnint, bidgo.Bid32ToUint8Rninta, bidgo.Bid32ToUint8Xceil, bidgo.Bid32ToUint8Xfloor, bidgo.Bid32ToUint8Xint, bidgo.Bid32ToUint8Xrnint, bidgo.Bid32ToUint8Xrninta, "uint8")
}

func bid32ToUint16ByMode(x uint32, mode string) (uint16, uint32, error) {
	return callGeneratedFFIBaseIntegerToMode(mode, x, bidgo.Bid32ToUint16Ceil, bidgo.Bid32ToUint16Floor, bidgo.Bid32ToUint16Int, bidgo.Bid32ToUint16Rnint, bidgo.Bid32ToUint16Rninta, bidgo.Bid32ToUint16Xceil, bidgo.Bid32ToUint16Xfloor, bidgo.Bid32ToUint16Xint, bidgo.Bid32ToUint16Xrnint, bidgo.Bid32ToUint16Xrninta, "uint16")
}

func bid32ToUint32ByMode(x uint32, mode string) (uint32, uint32, error) {
	switch mode {
	case "ceil":
		v, f := bidgo.Bid32ToUint32Ceil(x)
		return v, f, nil
	case "floor":
		v, f := bidgo.Bid32ToUint32Floor(x)
		return v, f, nil
	case "int":
		v, f := bidgo.Bid32ToUint32Int(x)
		return v, f, nil
	case "rnint":
		v, f := bidgo.Bid32ToUint32Rnint(x)
		return v, f, nil
	case "rninta":
		v, f := bidgo.Bid32ToUint32Rninta(x)
		return v, f, nil
	case "xceil":
		v, f := bidgo.Bid32ToUint32Xceil(x)
		return v, f, nil
	case "xfloor":
		v, f := bidgo.Bid32ToUint32Xfloor(x)
		return v, f, nil
	case "xint":
		v, f := bidgo.Bid32ToUint32Xint(x)
		return v, f, nil
	case "xrnint":
		v, f := bidgo.Bid32ToUint32Xrnint(x)
		return v, f, nil
	case "xrninta":
		v, f := bidgo.Bid32ToUint32Xrninta(x)
		return v, f, nil
	default:
		return 0, 0, fmt.Errorf("unsupported uint32 mode %q", mode)
	}
}

func bid32ToUint64ByMode(x uint32, mode string) (uint64, uint32, error) {
	switch mode {
	case "ceil":
		v, f := bidgo.Bid32ToUint64Ceil(x)
		return v, f, nil
	case "floor":
		v, f := bidgo.Bid32ToUint64Floor(x)
		return v, f, nil
	case "int":
		v, f := bidgo.Bid32ToUint64Int(x)
		return v, f, nil
	case "rnint":
		v, f := bidgo.Bid32ToUint64Rnint(x)
		return v, f, nil
	case "rninta":
		v, f := bidgo.Bid32ToUint64Rninta(x)
		return v, f, nil
	case "xceil":
		v, f := bidgo.Bid32ToUint64Xceil(x)
		return v, f, nil
	case "xfloor":
		v, f := bidgo.Bid32ToUint64Xfloor(x)
		return v, f, nil
	case "xint":
		v, f := bidgo.Bid32ToUint64Xint(x)
		return v, f, nil
	case "xrnint":
		v, f := bidgo.Bid32ToUint64Xrnint(x)
		return v, f, nil
	case "xrninta":
		v, f := bidgo.Bid32ToUint64Xrninta(x)
		return v, f, nil
	default:
		return 0, 0, fmt.Errorf("unsupported uint64 mode %q", mode)
	}
}

func bid64ToInt8ByMode(x uint64, mode string) (int8, uint32, error) {
	return callGeneratedFFIBaseIntegerToMode(mode, x, bidgo.Bid64ToInt8Ceil, bidgo.Bid64ToInt8Floor, bidgo.Bid64ToInt8Int, bidgo.Bid64ToInt8Rnint, bidgo.Bid64ToInt8Rninta, bidgo.Bid64ToInt8Xceil, bidgo.Bid64ToInt8Xfloor, bidgo.Bid64ToInt8Xint, bidgo.Bid64ToInt8Xrnint, bidgo.Bid64ToInt8Xrninta, "int8")
}

func bid64ToInt16ByMode(x uint64, mode string) (int16, uint32, error) {
	return callGeneratedFFIBaseIntegerToMode(mode, x, bidgo.Bid64ToInt16Ceil, bidgo.Bid64ToInt16Floor, bidgo.Bid64ToInt16Int, bidgo.Bid64ToInt16Rnint, bidgo.Bid64ToInt16Rninta, bidgo.Bid64ToInt16Xceil, bidgo.Bid64ToInt16Xfloor, bidgo.Bid64ToInt16Xint, bidgo.Bid64ToInt16Xrnint, bidgo.Bid64ToInt16Xrninta, "int16")
}

func bid64ToInt32ByMode(x uint64, mode string) (int32, uint32, error) {
	switch mode {
	case "ceil":
		v, f := bidgo.Bid64ToInt32Ceil(x)
		return v, f, nil
	case "floor":
		v, f := bidgo.Bid64ToInt32Floor(x)
		return v, f, nil
	case "int":
		v, f := bidgo.Bid64ToInt32Int(x)
		return v, f, nil
	case "rnint":
		v, f := bidgo.Bid64ToInt32Rnint(x)
		return v, f, nil
	case "rninta":
		v, f := bidgo.Bid64ToInt32Rninta(x)
		return v, f, nil
	case "xceil":
		v, f := bidgo.Bid64ToInt32Xceil(x)
		return v, f, nil
	case "xfloor":
		v, f := bidgo.Bid64ToInt32Xfloor(x)
		return v, f, nil
	case "xint":
		v, f := bidgo.Bid64ToInt32Xint(x)
		return v, f, nil
	case "xrnint":
		v, f := bidgo.Bid64ToInt32Xrnint(x)
		return v, f, nil
	case "xrninta":
		v, f := bidgo.Bid64ToInt32Xrninta(x)
		return v, f, nil
	default:
		return 0, 0, fmt.Errorf("unsupported int32 mode %q", mode)
	}
}

func bid64ToInt64ByMode(x uint64, mode string) (int64, uint32, error) {
	switch mode {
	case "ceil":
		v, f := bidgo.Bid64ToInt64Ceil(x)
		return v, f, nil
	case "floor":
		v, f := bidgo.Bid64ToInt64Floor(x)
		return v, f, nil
	case "int":
		v, f := bidgo.Bid64ToInt64Int(x)
		return v, f, nil
	case "rnint":
		v, f := bidgo.Bid64ToInt64Rnint(x)
		return v, f, nil
	case "rninta":
		v, f := bidgo.Bid64ToInt64Rninta(x)
		return v, f, nil
	case "xceil":
		v, f := bidgo.Bid64ToInt64Xceil(x)
		return v, f, nil
	case "xfloor":
		v, f := bidgo.Bid64ToInt64Xfloor(x)
		return v, f, nil
	case "xint":
		v, f := bidgo.Bid64ToInt64Xint(x)
		return v, f, nil
	case "xrnint":
		v, f := bidgo.Bid64ToInt64Xrnint(x)
		return v, f, nil
	case "xrninta":
		v, f := bidgo.Bid64ToInt64Xrninta(x)
		return v, f, nil
	default:
		return 0, 0, fmt.Errorf("unsupported int64 mode %q", mode)
	}
}

func bid64ToUint8ByMode(x uint64, mode string) (uint8, uint32, error) {
	return callGeneratedFFIBaseIntegerToMode(mode, x, bidgo.Bid64ToUint8Ceil, bidgo.Bid64ToUint8Floor, bidgo.Bid64ToUint8Int, bidgo.Bid64ToUint8Rnint, bidgo.Bid64ToUint8Rninta, bidgo.Bid64ToUint8Xceil, bidgo.Bid64ToUint8Xfloor, bidgo.Bid64ToUint8Xint, bidgo.Bid64ToUint8Xrnint, bidgo.Bid64ToUint8Xrninta, "uint8")
}

func bid64ToUint16ByMode(x uint64, mode string) (uint16, uint32, error) {
	return callGeneratedFFIBaseIntegerToMode(mode, x, bidgo.Bid64ToUint16Ceil, bidgo.Bid64ToUint16Floor, bidgo.Bid64ToUint16Int, bidgo.Bid64ToUint16Rnint, bidgo.Bid64ToUint16Rninta, bidgo.Bid64ToUint16Xceil, bidgo.Bid64ToUint16Xfloor, bidgo.Bid64ToUint16Xint, bidgo.Bid64ToUint16Xrnint, bidgo.Bid64ToUint16Xrninta, "uint16")
}

func bid64ToUint32ByMode(x uint64, mode string) (uint32, uint32, error) {
	switch mode {
	case "ceil":
		v, f := bidgo.Bid64ToUint32Ceil(x)
		return v, f, nil
	case "floor":
		v, f := bidgo.Bid64ToUint32Floor(x)
		return v, f, nil
	case "int":
		v, f := bidgo.Bid64ToUint32Int(x)
		return v, f, nil
	case "rnint":
		v, f := bidgo.Bid64ToUint32Rnint(x)
		return v, f, nil
	case "rninta":
		v, f := bidgo.Bid64ToUint32Rninta(x)
		return v, f, nil
	case "xceil":
		v, f := bidgo.Bid64ToUint32Xceil(x)
		return v, f, nil
	case "xfloor":
		v, f := bidgo.Bid64ToUint32Xfloor(x)
		return v, f, nil
	case "xint":
		v, f := bidgo.Bid64ToUint32Xint(x)
		return v, f, nil
	case "xrnint":
		v, f := bidgo.Bid64ToUint32Xrnint(x)
		return v, f, nil
	case "xrninta":
		v, f := bidgo.Bid64ToUint32Xrninta(x)
		return v, f, nil
	default:
		return 0, 0, fmt.Errorf("unsupported uint32 mode %q", mode)
	}
}

func bid64ToUint64ByMode(x uint64, mode string) (uint64, uint32, error) {
	switch mode {
	case "ceil":
		v, f := bidgo.Bid64ToUint64Ceil(x)
		return v, f, nil
	case "floor":
		v, f := bidgo.Bid64ToUint64Floor(x)
		return v, f, nil
	case "int":
		v, f := bidgo.Bid64ToUint64Int(x)
		return v, f, nil
	case "rnint":
		v, f := bidgo.Bid64ToUint64Rnint(x)
		return v, f, nil
	case "rninta":
		v, f := bidgo.Bid64ToUint64Rninta(x)
		return v, f, nil
	case "xceil":
		v, f := bidgo.Bid64ToUint64Xceil(x)
		return v, f, nil
	case "xfloor":
		v, f := bidgo.Bid64ToUint64Xfloor(x)
		return v, f, nil
	case "xint":
		v, f := bidgo.Bid64ToUint64Xint(x)
		return v, f, nil
	case "xrnint":
		v, f := bidgo.Bid64ToUint64Xrnint(x)
		return v, f, nil
	case "xrninta":
		v, f := bidgo.Bid64ToUint64Xrninta(x)
		return v, f, nil
	default:
		return 0, 0, fmt.Errorf("unsupported uint64 mode %q", mode)
	}
}

func bid128ToInt8ByMode(x bidgo.BID_UINT128, mode string) (int8, uint32, error) {
	return callGeneratedFFIBaseIntegerToMode(mode, x, bidgo.Bid128ToInt8Ceil, bidgo.Bid128ToInt8Floor, bidgo.Bid128ToInt8Int, bidgo.Bid128ToInt8Rnint, bidgo.Bid128ToInt8Rninta, bidgo.Bid128ToInt8Xceil, bidgo.Bid128ToInt8Xfloor, bidgo.Bid128ToInt8Xint, bidgo.Bid128ToInt8Xrnint, bidgo.Bid128ToInt8Xrninta, "int8")
}

func bid128ToInt16ByMode(x bidgo.BID_UINT128, mode string) (int16, uint32, error) {
	return callGeneratedFFIBaseIntegerToMode(mode, x, bidgo.Bid128ToInt16Ceil, bidgo.Bid128ToInt16Floor, bidgo.Bid128ToInt16Int, bidgo.Bid128ToInt16Rnint, bidgo.Bid128ToInt16Rninta, bidgo.Bid128ToInt16Xceil, bidgo.Bid128ToInt16Xfloor, bidgo.Bid128ToInt16Xint, bidgo.Bid128ToInt16Xrnint, bidgo.Bid128ToInt16Xrninta, "int16")
}

func bid128ToInt32ByMode(x bidgo.BID_UINT128, mode string) (int32, uint32, error) {
	switch mode {
	case "ceil":
		v, f := bidgo.Bid128ToInt32Ceil(x)
		return v, f, nil
	case "floor":
		v, f := bidgo.Bid128ToInt32Floor(x)
		return v, f, nil
	case "int":
		v, f := bidgo.Bid128ToInt32Int(x)
		return v, f, nil
	case "rnint":
		v, f := bidgo.Bid128ToInt32Rnint(x)
		return v, f, nil
	case "rninta":
		v, f := bidgo.Bid128ToInt32Rninta(x)
		return v, f, nil
	case "xceil":
		v, f := bidgo.Bid128ToInt32Xceil(x)
		return v, f, nil
	case "xfloor":
		v, f := bidgo.Bid128ToInt32Xfloor(x)
		return v, f, nil
	case "xint":
		v, f := bidgo.Bid128ToInt32Xint(x)
		return v, f, nil
	case "xrnint":
		v, f := bidgo.Bid128ToInt32Xrnint(x)
		return v, f, nil
	case "xrninta":
		v, f := bidgo.Bid128ToInt32Xrninta(x)
		return v, f, nil
	default:
		return 0, 0, fmt.Errorf("unsupported int32 mode %q", mode)
	}
}

func bid128ToInt64ByMode(x bidgo.BID_UINT128, mode string) (int64, uint32, error) {
	switch mode {
	case "ceil":
		v, f := bidgo.Bid128ToInt64Ceil(x)
		return v, f, nil
	case "floor":
		v, f := bidgo.Bid128ToInt64Floor(x)
		return v, f, nil
	case "int":
		v, f := bidgo.Bid128ToInt64Int(x)
		return v, f, nil
	case "rnint":
		v, f := bidgo.Bid128ToInt64Rnint(x)
		return v, f, nil
	case "rninta":
		v, f := bidgo.Bid128ToInt64Rninta(x)
		return v, f, nil
	case "xceil":
		v, f := bidgo.Bid128ToInt64Xceil(x)
		return v, f, nil
	case "xfloor":
		v, f := bidgo.Bid128ToInt64Xfloor(x)
		return v, f, nil
	case "xint":
		v, f := bidgo.Bid128ToInt64Xint(x)
		return v, f, nil
	case "xrnint":
		v, f := bidgo.Bid128ToInt64Xrnint(x)
		return v, f, nil
	case "xrninta":
		v, f := bidgo.Bid128ToInt64Xrninta(x)
		return v, f, nil
	default:
		return 0, 0, fmt.Errorf("unsupported int64 mode %q", mode)
	}
}

func bid128ToUint8ByMode(x bidgo.BID_UINT128, mode string) (uint8, uint32, error) {
	return callGeneratedFFIBaseIntegerToMode(mode, x, bidgo.Bid128ToUint8Ceil, bidgo.Bid128ToUint8Floor, bidgo.Bid128ToUint8Int, bidgo.Bid128ToUint8Rnint, bidgo.Bid128ToUint8Rninta, bidgo.Bid128ToUint8Xceil, bidgo.Bid128ToUint8Xfloor, bidgo.Bid128ToUint8Xint, bidgo.Bid128ToUint8Xrnint, bidgo.Bid128ToUint8Xrninta, "uint8")
}

func bid128ToUint16ByMode(x bidgo.BID_UINT128, mode string) (uint16, uint32, error) {
	return callGeneratedFFIBaseIntegerToMode(mode, x, bidgo.Bid128ToUint16Ceil, bidgo.Bid128ToUint16Floor, bidgo.Bid128ToUint16Int, bidgo.Bid128ToUint16Rnint, bidgo.Bid128ToUint16Rninta, bidgo.Bid128ToUint16Xceil, bidgo.Bid128ToUint16Xfloor, bidgo.Bid128ToUint16Xint, bidgo.Bid128ToUint16Xrnint, bidgo.Bid128ToUint16Xrninta, "uint16")
}

func bid128ToUint32ByMode(x bidgo.BID_UINT128, mode string) (uint32, uint32, error) {
	switch mode {
	case "ceil":
		v, f := bidgo.Bid128ToUint32Ceil(x)
		return v, f, nil
	case "floor":
		v, f := bidgo.Bid128ToUint32Floor(x)
		return v, f, nil
	case "int":
		v, f := bidgo.Bid128ToUint32Int(x)
		return v, f, nil
	case "rnint":
		v, f := bidgo.Bid128ToUint32Rnint(x)
		return v, f, nil
	case "rninta":
		v, f := bidgo.Bid128ToUint32Rninta(x)
		return v, f, nil
	case "xceil":
		v, f := bidgo.Bid128ToUint32Xceil(x)
		return v, f, nil
	case "xfloor":
		v, f := bidgo.Bid128ToUint32Xfloor(x)
		return v, f, nil
	case "xint":
		v, f := bidgo.Bid128ToUint32Xint(x)
		return v, f, nil
	case "xrnint":
		v, f := bidgo.Bid128ToUint32Xrnint(x)
		return v, f, nil
	case "xrninta":
		v, f := bidgo.Bid128ToUint32Xrninta(x)
		return v, f, nil
	default:
		return 0, 0, fmt.Errorf("unsupported uint32 mode %q", mode)
	}
}

func bid128ToUint64ByMode(x bidgo.BID_UINT128, mode string) (uint64, uint32, error) {
	switch mode {
	case "ceil":
		v, f := bidgo.Bid128ToUint64Ceil(x)
		return v, f, nil
	case "floor":
		v, f := bidgo.Bid128ToUint64Floor(x)
		return v, f, nil
	case "int":
		v, f := bidgo.Bid128ToUint64Int(x)
		return v, f, nil
	case "rnint":
		v, f := bidgo.Bid128ToUint64Rnint(x)
		return v, f, nil
	case "rninta":
		v, f := bidgo.Bid128ToUint64Rninta(x)
		return v, f, nil
	case "xceil":
		v, f := bidgo.Bid128ToUint64Xceil(x)
		return v, f, nil
	case "xfloor":
		v, f := bidgo.Bid128ToUint64Xfloor(x)
		return v, f, nil
	case "xint":
		v, f := bidgo.Bid128ToUint64Xint(x)
		return v, f, nil
	case "xrnint":
		v, f := bidgo.Bid128ToUint64Xrnint(x)
		return v, f, nil
	case "xrninta":
		v, f := bidgo.Bid128ToUint64Xrninta(x)
		return v, f, nil
	default:
		return 0, 0, fmt.Errorf("unsupported uint64 mode %q", mode)
	}
}

func parseFFIInt32Operand(tc generatedFFICase) (int32, error) {
	if len(tc.Operands) != 1 {
		return 0, fmt.Errorf("%s expects 1 operand, got %d", tc.Function, len(tc.Operands))
	}
	value, err := strconv.ParseInt(tc.Operands[0], 10, 32)
	if err != nil {
		return 0, fmt.Errorf("parse int32 operand %q: %w", tc.Operands[0], err)
	}
	return int32(value), nil
}

func parseFFIInt64Operand(tc generatedFFICase) (int64, error) {
	if len(tc.Operands) != 1 {
		return 0, fmt.Errorf("%s expects 1 operand, got %d", tc.Function, len(tc.Operands))
	}
	value, err := strconv.ParseInt(tc.Operands[0], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse int64 operand %q: %w", tc.Operands[0], err)
	}
	return value, nil
}

func parseFFIUint32DecimalOperand(tc generatedFFICase) (uint32, error) {
	if len(tc.Operands) != 1 {
		return 0, fmt.Errorf("%s expects 1 operand, got %d", tc.Function, len(tc.Operands))
	}
	value, err := strconv.ParseUint(tc.Operands[0], 10, 32)
	if err != nil {
		return 0, fmt.Errorf("parse uint32 operand %q: %w", tc.Operands[0], err)
	}
	return uint32(value), nil
}

func parseFFIUint64DecimalOperand(tc generatedFFICase) (uint64, error) {
	if len(tc.Operands) != 1 {
		return 0, fmt.Errorf("%s expects 1 operand, got %d", tc.Function, len(tc.Operands))
	}
	value, err := strconv.ParseUint(tc.Operands[0], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse uint64 operand %q: %w", tc.Operands[0], err)
	}
	return value, nil
}

func parseFFIUint32UnaryOperand(tc generatedFFICase) (uint32, error) {
	if len(tc.Operands) != 1 {
		return 0, fmt.Errorf("%s expects 1 operand, got %d", tc.Function, len(tc.Operands))
	}
	a, err := strconv.ParseUint(tc.Operands[0], 16, 32)
	if err != nil {
		return 0, fmt.Errorf("parse operand %q: %w", tc.Operands[0], err)
	}
	return uint32(a), nil
}

func parseFFIUint32BinaryOperands(tc generatedFFICase) (uint32, uint32, error) {
	if len(tc.Operands) != 2 {
		return 0, 0, fmt.Errorf("%s expects 2 operands, got %d", tc.Function, len(tc.Operands))
	}
	a, err := strconv.ParseUint(tc.Operands[0], 16, 32)
	if err != nil {
		return 0, 0, fmt.Errorf("parse operand_a %q: %w", tc.Operands[0], err)
	}
	b, err := strconv.ParseUint(tc.Operands[1], 16, 32)
	if err != nil {
		return 0, 0, fmt.Errorf("parse operand_b %q: %w", tc.Operands[1], err)
	}
	return uint32(a), uint32(b), nil
}

func parseFFIUint32IntOperands(tc generatedFFICase) (uint32, int, error) {
	if len(tc.Operands) != 2 {
		return 0, 0, fmt.Errorf("%s expects 2 operands, got %d", tc.Function, len(tc.Operands))
	}
	a, err := strconv.ParseUint(tc.Operands[0], 16, 32)
	if err != nil {
		return 0, 0, fmt.Errorf("parse operand_a %q: %w", tc.Operands[0], err)
	}
	n, err := strconv.Atoi(tc.Operands[1])
	if err != nil {
		return 0, 0, fmt.Errorf("parse operand_n %q: %w", tc.Operands[1], err)
	}
	return uint32(a), n, nil
}

func parseFFIUint32TernaryOperands(tc generatedFFICase) (uint32, uint32, uint32, error) {
	if len(tc.Operands) != 3 {
		return 0, 0, 0, fmt.Errorf("%s expects 3 operands, got %d", tc.Function, len(tc.Operands))
	}
	a, err := strconv.ParseUint(tc.Operands[0], 16, 32)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("parse operand_a %q: %w", tc.Operands[0], err)
	}
	b, err := strconv.ParseUint(tc.Operands[1], 16, 32)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("parse operand_b %q: %w", tc.Operands[1], err)
	}
	c, err := strconv.ParseUint(tc.Operands[2], 16, 32)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("parse operand_c %q: %w", tc.Operands[2], err)
	}
	return uint32(a), uint32(b), uint32(c), nil
}

func parseFFIUint64UnaryOperand(tc generatedFFICase) (uint64, error) {
	if len(tc.Operands) != 1 {
		return 0, fmt.Errorf("%s expects 1 operand, got %d", tc.Function, len(tc.Operands))
	}
	a, err := strconv.ParseUint(tc.Operands[0], 16, 64)
	if err != nil {
		return 0, fmt.Errorf("parse operand %q: %w", tc.Operands[0], err)
	}
	return a, nil
}

func parseFFIUint64BinaryOperands(tc generatedFFICase) (uint64, uint64, error) {
	if len(tc.Operands) != 2 {
		return 0, 0, fmt.Errorf("%s expects 2 operands, got %d", tc.Function, len(tc.Operands))
	}
	a, err := strconv.ParseUint(tc.Operands[0], 16, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("parse operand_a %q: %w", tc.Operands[0], err)
	}
	b, err := strconv.ParseUint(tc.Operands[1], 16, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("parse operand_b %q: %w", tc.Operands[1], err)
	}
	return a, b, nil
}

func parseFFIUint64IntOperands(tc generatedFFICase) (uint64, int, error) {
	if len(tc.Operands) != 2 {
		return 0, 0, fmt.Errorf("%s expects 2 operands, got %d", tc.Function, len(tc.Operands))
	}
	a, err := strconv.ParseUint(tc.Operands[0], 16, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("parse operand_a %q: %w", tc.Operands[0], err)
	}
	n, err := strconv.Atoi(tc.Operands[1])
	if err != nil {
		return 0, 0, fmt.Errorf("parse operand_n %q: %w", tc.Operands[1], err)
	}
	return a, n, nil
}

func parseFFIUint64TernaryOperands(tc generatedFFICase) (uint64, uint64, uint64, error) {
	if len(tc.Operands) != 3 {
		return 0, 0, 0, fmt.Errorf("%s expects 3 operands, got %d", tc.Function, len(tc.Operands))
	}
	a, err := strconv.ParseUint(tc.Operands[0], 16, 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("parse operand_a %q: %w", tc.Operands[0], err)
	}
	b, err := strconv.ParseUint(tc.Operands[1], 16, 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("parse operand_b %q: %w", tc.Operands[1], err)
	}
	c, err := strconv.ParseUint(tc.Operands[2], 16, 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("parse operand_c %q: %w", tc.Operands[2], err)
	}
	return a, b, c, nil
}

func parseFFIUint128UnaryOperand(tc generatedFFICase) (Decimal128BID, error) {
	if len(tc.Operands) != 1 {
		return Decimal128BID{}, fmt.Errorf("%s expects 1 operand, got %d", tc.Function, len(tc.Operands))
	}
	return parseFFIUint128Bits(tc.Operands[0])
}

func parseFFIReadtestUint128UnaryOperand(tc generatedFFICase) (Decimal128BID, error) {
	if len(tc.Operands) != 1 {
		return Decimal128BID{}, fmt.Errorf("%s expects 1 operand, got %d", tc.Function, len(tc.Operands))
	}
	raw, err := parseReadtestBits128(generatedFFIReadtestOperands(tc)[0])
	if err != nil {
		return Decimal128BID{}, err
	}
	return Decimal128BID(raw), nil
}

func parseFFIUint128BinaryOperands(tc generatedFFICase) (Decimal128BID, Decimal128BID, error) {
	if len(tc.Operands) != 2 {
		return Decimal128BID{}, Decimal128BID{}, fmt.Errorf("%s expects 2 operands, got %d", tc.Function, len(tc.Operands))
	}
	a, err := parseFFIUint128Bits(tc.Operands[0])
	if err != nil {
		return Decimal128BID{}, Decimal128BID{}, fmt.Errorf("parse operand_a %q: %w", tc.Operands[0], err)
	}
	b, err := parseFFIUint128Bits(tc.Operands[1])
	if err != nil {
		return Decimal128BID{}, Decimal128BID{}, fmt.Errorf("parse operand_b %q: %w", tc.Operands[1], err)
	}
	return a, b, nil
}

func parseFFIUint128IntOperands(tc generatedFFICase) (Decimal128BID, int, error) {
	if len(tc.Operands) != 2 {
		return Decimal128BID{}, 0, fmt.Errorf("%s expects 2 operands, got %d", tc.Function, len(tc.Operands))
	}
	a, err := parseFFIUint128Bits(tc.Operands[0])
	if err != nil {
		return Decimal128BID{}, 0, fmt.Errorf("parse operand_a %q: %w", tc.Operands[0], err)
	}
	n, err := strconv.Atoi(tc.Operands[1])
	if err != nil {
		return Decimal128BID{}, 0, fmt.Errorf("parse operand_n %q: %w", tc.Operands[1], err)
	}
	return a, n, nil
}

func parseFFIUint128TernaryOperands(tc generatedFFICase) (Decimal128BID, Decimal128BID, Decimal128BID, error) {
	if len(tc.Operands) != 3 {
		return Decimal128BID{}, Decimal128BID{}, Decimal128BID{}, fmt.Errorf("%s expects 3 operands, got %d", tc.Function, len(tc.Operands))
	}
	a, err := parseFFIUint128Bits(tc.Operands[0])
	if err != nil {
		return Decimal128BID{}, Decimal128BID{}, Decimal128BID{}, fmt.Errorf("parse operand_a %q: %w", tc.Operands[0], err)
	}
	b, err := parseFFIUint128Bits(tc.Operands[1])
	if err != nil {
		return Decimal128BID{}, Decimal128BID{}, Decimal128BID{}, fmt.Errorf("parse operand_b %q: %w", tc.Operands[1], err)
	}
	c, err := parseFFIUint128Bits(tc.Operands[2])
	if err != nil {
		return Decimal128BID{}, Decimal128BID{}, Decimal128BID{}, fmt.Errorf("parse operand_c %q: %w", tc.Operands[2], err)
	}
	return a, b, c, nil
}

func parseFFIUint128Bits(input string) (Decimal128BID, error) {
	raw, err := hex.DecodeString(input)
	if err != nil {
		return Decimal128BID{}, err
	}
	if len(raw) != 16 {
		return Decimal128BID{}, fmt.Errorf("decoded %d bytes, want 16", len(raw))
	}
	var out Decimal128BID
	copy(out[:], raw)
	return out, nil
}

func formatFFIUint128Bits(value Decimal128BID) string {
	raw := value.ToBytes()
	return hex.EncodeToString(raw[:])
}

func ffiUint128ToC(value Decimal128BID) C.BID_UINT128 {
	return *(*C.BID_UINT128)(unsafe.Pointer(&value))
}

func ffiUint128FromC(value C.BID_UINT128) Decimal128BID {
	return *(*Decimal128BID)(unsafe.Pointer(&value))
}

func runGeneratedFFICase32DecimalFlagUnary(operation string, a uint32) (uint32, uint32, uint32, uint32) {
	switch operation {
	case "quantum":
		var flags C._IDEC_flags
		native := uint32(C.bid32_quantum(C.BID_UINT32(a), &flags))
		return native, uint32(flags), bidgo.Bid32Quantum(a), 0
	default:
		panic("unsupported 32-bit ffi decimal flag unary operation")
	}
}

func runGeneratedFFICase32DecimalFlagInt(operation string, a uint32, n int, rounding int) (uint32, uint32, uint32, uint32) {
	switch operation {
	case "ldexp":
		var flags C._IDEC_flags
		native := uint32(C.bid32_ldexp(C.BID_UINT32(a), C.int(n), generatedFFICRoundingMode(rounding), &flags))
		exposed, exposedFlags := bidgo.Bid32LdexpWithFlags(a, n, rounding)
		return native, uint32(flags), exposed, exposedFlags
	case "scalbln":
		var flags C._IDEC_flags
		native := uint32(C.bid32_scalbln(C.BID_UINT32(a), C.long(n), generatedFFICRoundingMode(rounding), &flags))
		exposed, exposedFlags := bidgo.Bid32ScalblnWithFlags(a, int64(n), rounding)
		return native, uint32(flags), exposed, exposedFlags
	default:
		panic("unsupported 32-bit ffi decimal flag int operation")
	}
}

func runGeneratedFFICase64DecimalFlagUnary(operation string, a uint64) (uint64, uint32, uint64, uint32) {
	switch operation {
	case "quantum":
		var flags C._IDEC_flags
		native := uint64(C.bid64_quantum(C.BID_UINT64(a), &flags))
		return native, uint32(flags), bidgo.Bid64Quantum(a), 0
	default:
		panic("unsupported 64-bit ffi decimal flag unary operation")
	}
}

func runGeneratedFFICase64DecimalFlagInt(operation string, a uint64, n int, rounding int) (uint64, uint32, uint64, uint32) {
	switch operation {
	case "ldexp":
		var flags C._IDEC_flags
		native := uint64(C.bid64_ldexp(C.BID_UINT64(a), C.int(n), generatedFFICRoundingMode(rounding), &flags))
		exposed, exposedFlags := bidgo.Bid64Ldexp(a, n, rounding)
		return native, uint32(flags), exposed, exposedFlags
	case "scalbln":
		var flags C._IDEC_flags
		native := uint64(C.bid64_scalbln(C.BID_UINT64(a), C.long(n), generatedFFICRoundingMode(rounding), &flags))
		exposed, exposedFlags := bidgo.Bid64Scalbln(a, int64(n), rounding)
		return native, uint32(flags), exposed, exposedFlags
	default:
		panic("unsupported 64-bit ffi decimal flag int operation")
	}
}

func runGeneratedFFICase128DecimalFlagUnary(operation string, a Decimal128BID) (Decimal128BID, uint32, Decimal128BID, uint32) {
	ca := ffiUint128ToC(a)
	ga := decimal128BIDAsBidgo(a)
	switch operation {
	case "quantum":
		var flags C._IDEC_flags
		native := ffiUint128FromC(C.bid128_quantum(ca, &flags))
		return native, uint32(flags), decimal128BIDFromBidgo(bidgo.Bid128Quantum(ga)), 0
	default:
		panic("unsupported 128-bit ffi decimal flag unary operation")
	}
}

func runGeneratedFFICase128DecimalFlagInt(operation string, a Decimal128BID, n int, rounding int) (Decimal128BID, uint32, Decimal128BID, uint32) {
	ca := ffiUint128ToC(a)
	ga := decimal128BIDAsBidgo(a)
	switch operation {
	case "ldexp":
		var flags C._IDEC_flags
		native := ffiUint128FromC(C.bid128_ldexp(ca, C.int(n), generatedFFICRoundingMode(rounding), &flags))
		exposed, exposedFlags := bidgo.Bid128Ldexp(ga, n, rounding)
		return native, uint32(flags), decimal128BIDFromBidgo(exposed), exposedFlags
	case "scalbln":
		var flags C._IDEC_flags
		native := ffiUint128FromC(C.bid128_scalbln(ca, C.long(n), generatedFFICRoundingMode(rounding), &flags))
		var exposedFlags uint32
		exposed := bidgo.Bid128Scalbln(ga, int64(n), rounding, &exposedFlags)
		return native, uint32(flags), decimal128BIDFromBidgo(exposed), exposedFlags
	default:
		panic("unsupported 128-bit ffi decimal flag int operation")
	}
}

func runGeneratedFFICase32Int64Unary(operation string, a uint32) (int64, uint32, int64, uint32) {
	switch operation {
	case "llquantexp":
		var flags C._IDEC_flags
		native := int64(C.bid32_llquantexp(C.BID_UINT32(a), &flags))
		exposed, exposedFlags := bidgo.Bid32LLQuantexp(a)
		return native, uint32(flags), exposed, exposedFlags
	default:
		panic("unsupported 32-bit ffi int64 unary operation")
	}
}

func runGeneratedFFICase64Int64Unary(operation string, a uint64) (int64, uint32, int64, uint32) {
	switch operation {
	case "llquantexp":
		var flags C._IDEC_flags
		native := int64(C.bid64_llquantexp(C.BID_UINT64(a), &flags))
		exposed, exposedFlags := bidgo.Bid64LLQuantexp(a)
		return native, uint32(flags), exposed, exposedFlags
	default:
		panic("unsupported 64-bit ffi int64 unary operation")
	}
}

func runGeneratedFFICase128Int64Unary(operation string, a Decimal128BID) (int64, uint32, int64, uint32) {
	ca := ffiUint128ToC(a)
	ga := decimal128BIDAsBidgo(a)
	switch operation {
	case "llquantexp":
		var flags C._IDEC_flags
		native := int64(C.bid128_llquantexp(ca, &flags))
		var exposedFlags uint32
		exposed := bidgo.Bid128Llquantexp(ga, &exposedFlags)
		return native, uint32(flags), exposed, exposedFlags
	default:
		panic("unsupported 128-bit ffi int64 unary operation")
	}
}

func runGeneratedFFICase32IntUnary(operation string, a uint32) (int, uint32, int, uint32) {
	switch operation {
	case "class":
		return int(C.bid32_class(C.BID_UINT32(a))), 0, bidgo.Bid32Class(a), 0
	case "isSigned":
		return int(C.bid32_isSigned(C.BID_UINT32(a))), 0, bidgo.Bid32IsSigned(a), 0
	case "isNormal":
		return int(C.bid32_isNormal(C.BID_UINT32(a))), 0, bidgo.Bid32IsNormal(a), 0
	case "isSubnormal":
		return int(C.bid32_isSubnormal(C.BID_UINT32(a))), 0, bidgo.Bid32IsSubnormal(a), 0
	case "isFinite":
		return int(C.bid32_isFinite(C.BID_UINT32(a))), 0, bidgo.Bid32IsFinite(a), 0
	case "isZero":
		return int(C.bid32_isZero(C.BID_UINT32(a))), 0, bidgo.Bid32IsZero32(a), 0
	case "isInf":
		return int(C.bid32_isInf(C.BID_UINT32(a))), 0, bidgo.Bid32IsInf32(a), 0
	case "isNaN":
		return int(C.bid32_isNaN(C.BID_UINT32(a))), 0, bidgo.Bid32IsNaN32(a), 0
	case "isSignaling":
		return int(C.bid32_isSignaling(C.BID_UINT32(a))), 0, bidgo.Bid32IsSignaling(a), 0
	case "isCanonical":
		return int(C.bid32_isCanonical(C.BID_UINT32(a))), 0, bidgo.Bid32IsCanonical(a), 0
	case "radix":
		return int(C.bid32_radix(C.BID_UINT32(a))), 0, bidgo.Bid32Radix(), 0
	case "quantexp":
		var flags C._IDEC_flags
		native := int(C.bid32_quantexp(C.BID_UINT32(a), &flags))
		exposed, exposedFlags := bidgo.Bid32Quantexp(a)
		return native, uint32(flags), int(exposed), exposedFlags
	case "ilogb":
		var flags C._IDEC_flags
		native := int(C.bid32_ilogb(C.BID_UINT32(a), &flags))
		exposed, exposedFlags := bidgo.Bid32ILogb(a)
		return native, uint32(flags), exposed, exposedFlags
	default:
		panic("unsupported 32-bit ffi int unary operation")
	}
}

func runGeneratedFFICase32IntBinary(operation string, a, b uint32) (int, uint32, int, uint32) {
	switch operation {
	case "totalOrder":
		return int(C.bid32_totalOrder(C.BID_UINT32(a), C.BID_UINT32(b))), 0, bidgo.Bid32TotalOrder(a, b), 0
	case "totalOrderMag":
		return int(C.bid32_totalOrderMag(C.BID_UINT32(a), C.BID_UINT32(b))), 0, bidgo.Bid32TotalOrderMag(a, b), 0
	case "sameQuantum":
		exposed := 0
		if bidgo.Bid32SameQuantum(a, b) {
			exposed = 1
		}
		return int(C.bid32_sameQuantum(C.BID_UINT32(a), C.BID_UINT32(b))), 0, exposed, 0
	case "quiet_equal":
		var flags C._IDEC_flags
		native := int(C.bid32_quiet_equal(C.BID_UINT32(a), C.BID_UINT32(b), &flags))
		exposed, exposedFlags := bidgo.Bid32QuietEqual(a, b)
		return native, uint32(flags), exposed, exposedFlags
	case "quiet_greater":
		var flags C._IDEC_flags
		native := int(C.bid32_quiet_greater(C.BID_UINT32(a), C.BID_UINT32(b), &flags))
		exposed, exposedFlags := bidgo.Bid32QuietGreater(a, b)
		return native, uint32(flags), exposed, exposedFlags
	case "quiet_greater_equal":
		var flags C._IDEC_flags
		native := int(C.bid32_quiet_greater_equal(C.BID_UINT32(a), C.BID_UINT32(b), &flags))
		exposed, exposedFlags := bidgo.Bid32QuietGreaterEqual(a, b)
		return native, uint32(flags), exposed, exposedFlags
	case "quiet_greater_unordered":
		var flags C._IDEC_flags
		native := int(C.bid32_quiet_greater_unordered(C.BID_UINT32(a), C.BID_UINT32(b), &flags))
		exposed, exposedFlags := bidgo.Bid32QuietGreaterUnordered(a, b)
		return native, uint32(flags), exposed, exposedFlags
	case "quiet_less":
		var flags C._IDEC_flags
		native := int(C.bid32_quiet_less(C.BID_UINT32(a), C.BID_UINT32(b), &flags))
		exposed, exposedFlags := bidgo.Bid32QuietLess(a, b)
		return native, uint32(flags), exposed, exposedFlags
	case "quiet_less_equal":
		var flags C._IDEC_flags
		native := int(C.bid32_quiet_less_equal(C.BID_UINT32(a), C.BID_UINT32(b), &flags))
		exposed, exposedFlags := bidgo.Bid32QuietLessEqual(a, b)
		return native, uint32(flags), exposed, exposedFlags
	case "quiet_less_unordered":
		var flags C._IDEC_flags
		native := int(C.bid32_quiet_less_unordered(C.BID_UINT32(a), C.BID_UINT32(b), &flags))
		exposed, exposedFlags := bidgo.Bid32QuietLessUnordered(a, b)
		return native, uint32(flags), exposed, exposedFlags
	case "quiet_not_equal":
		var flags C._IDEC_flags
		native := int(C.bid32_quiet_not_equal(C.BID_UINT32(a), C.BID_UINT32(b), &flags))
		exposed, exposedFlags := bidgo.Bid32QuietNotEqual(a, b)
		return native, uint32(flags), exposed, exposedFlags
	case "quiet_not_greater":
		var flags C._IDEC_flags
		native := int(C.bid32_quiet_not_greater(C.BID_UINT32(a), C.BID_UINT32(b), &flags))
		exposed, exposedFlags := bidgo.Bid32QuietNotGreater(a, b)
		return native, uint32(flags), exposed, exposedFlags
	case "quiet_not_less":
		var flags C._IDEC_flags
		native := int(C.bid32_quiet_not_less(C.BID_UINT32(a), C.BID_UINT32(b), &flags))
		exposed, exposedFlags := bidgo.Bid32QuietNotLess(a, b)
		return native, uint32(flags), exposed, exposedFlags
	case "quiet_ordered":
		var flags C._IDEC_flags
		native := int(C.bid32_quiet_ordered(C.BID_UINT32(a), C.BID_UINT32(b), &flags))
		exposed, exposedFlags := bidgo.Bid32QuietOrdered(a, b)
		return native, uint32(flags), exposed, exposedFlags
	case "quiet_unordered":
		var flags C._IDEC_flags
		native := int(C.bid32_quiet_unordered(C.BID_UINT32(a), C.BID_UINT32(b), &flags))
		exposed, exposedFlags := bidgo.Bid32QuietUnordered(a, b)
		return native, uint32(flags), exposed, exposedFlags
	case "signaling_greater":
		var flags C._IDEC_flags
		native := int(C.bid32_signaling_greater(C.BID_UINT32(a), C.BID_UINT32(b), &flags))
		exposed, exposedFlags := bidgo.Bid32SignalingGreater(a, b)
		return native, uint32(flags), exposed, exposedFlags
	case "signaling_greater_equal":
		var flags C._IDEC_flags
		native := int(C.bid32_signaling_greater_equal(C.BID_UINT32(a), C.BID_UINT32(b), &flags))
		exposed, exposedFlags := bidgo.Bid32SignalingGreaterEqual(a, b)
		return native, uint32(flags), exposed, exposedFlags
	case "signaling_greater_unordered":
		var flags C._IDEC_flags
		native := int(C.bid32_signaling_greater_unordered(C.BID_UINT32(a), C.BID_UINT32(b), &flags))
		exposed, exposedFlags := bidgo.Bid32SignalingGreaterUnordered(a, b)
		return native, uint32(flags), exposed, exposedFlags
	case "signaling_less":
		var flags C._IDEC_flags
		native := int(C.bid32_signaling_less(C.BID_UINT32(a), C.BID_UINT32(b), &flags))
		exposed, exposedFlags := bidgo.Bid32SignalingLess(a, b)
		return native, uint32(flags), exposed, exposedFlags
	case "signaling_less_equal":
		var flags C._IDEC_flags
		native := int(C.bid32_signaling_less_equal(C.BID_UINT32(a), C.BID_UINT32(b), &flags))
		exposed, exposedFlags := bidgo.Bid32SignalingLessEqual(a, b)
		return native, uint32(flags), exposed, exposedFlags
	case "signaling_less_unordered":
		var flags C._IDEC_flags
		native := int(C.bid32_signaling_less_unordered(C.BID_UINT32(a), C.BID_UINT32(b), &flags))
		exposed, exposedFlags := bidgo.Bid32SignalingLessUnordered(a, b)
		return native, uint32(flags), exposed, exposedFlags
	case "signaling_not_greater":
		var flags C._IDEC_flags
		native := int(C.bid32_signaling_not_greater(C.BID_UINT32(a), C.BID_UINT32(b), &flags))
		exposed, exposedFlags := bidgo.Bid32SignalingNotGreater(a, b)
		return native, uint32(flags), exposed, exposedFlags
	case "signaling_not_less":
		var flags C._IDEC_flags
		native := int(C.bid32_signaling_not_less(C.BID_UINT32(a), C.BID_UINT32(b), &flags))
		exposed, exposedFlags := bidgo.Bid32SignalingNotLess(a, b)
		return native, uint32(flags), exposed, exposedFlags
	default:
		panic("unsupported 32-bit ffi int binary operation")
	}
}

func runGeneratedFFICase64IntUnary(operation string, a uint64) (int, uint32, int, uint32) {
	switch operation {
	case "class":
		return int(C.bid64_class(C.BID_UINT64(a))), 0, bidgo.Bid64Class(a), 0
	case "isSigned":
		return int(C.bid64_isSigned(C.BID_UINT64(a))), 0, bidgo.Bid64IsSigned(a), 0
	case "isNormal":
		return int(C.bid64_isNormal(C.BID_UINT64(a))), 0, bidgo.Bid64IsNormal(a), 0
	case "isSubnormal":
		return int(C.bid64_isSubnormal(C.BID_UINT64(a))), 0, bidgo.Bid64IsSubnormal(a), 0
	case "isFinite":
		return int(C.bid64_isFinite(C.BID_UINT64(a))), 0, bidgo.Bid64IsFinite(a), 0
	case "isZero":
		return int(C.bid64_isZero(C.BID_UINT64(a))), 0, bidgo.Bid64IsZero(a), 0
	case "isInf":
		return int(C.bid64_isInf(C.BID_UINT64(a))), 0, bidgo.Bid64IsInf(a), 0
	case "isNaN":
		return int(C.bid64_isNaN(C.BID_UINT64(a))), 0, bidgo.Bid64IsNaN(a), 0
	case "isSignaling":
		return int(C.bid64_isSignaling(C.BID_UINT64(a))), 0, bidgo.Bid64IsSignaling(a), 0
	case "isCanonical":
		return int(C.bid64_isCanonical(C.BID_UINT64(a))), 0, bidgo.Bid64IsCanonical(a), 0
	case "radix":
		return int(C.bid64_radix(C.BID_UINT64(a))), 0, bidgo.Bid64Radix(), 0
	case "quantexp":
		var flags C._IDEC_flags
		native := int(C.bid64_quantexp(C.BID_UINT64(a), &flags))
		exposed, exposedFlags := bidgo.Bid64Quantexp(a)
		return native, uint32(flags), int(exposed), exposedFlags
	case "ilogb":
		var flags C._IDEC_flags
		native := int(C.bid64_ilogb(C.BID_UINT64(a), &flags))
		exposed, exposedFlags := bidgo.Bid64ILogb(a)
		return native, uint32(flags), exposed, exposedFlags
	default:
		panic("unsupported 64-bit ffi int unary operation")
	}
}

func runGeneratedFFICase64IntBinary(operation string, a, b uint64) (int, uint32, int, uint32) {
	switch operation {
	case "totalOrder":
		return int(C.bid64_totalOrder(C.BID_UINT64(a), C.BID_UINT64(b))), 0, bidgo.Bid64TotalOrder(a, b), 0
	case "totalOrderMag":
		return int(C.bid64_totalOrderMag(C.BID_UINT64(a), C.BID_UINT64(b))), 0, bidgo.Bid64TotalOrderMag(a, b), 0
	case "sameQuantum":
		return int(C.bid64_sameQuantum(C.BID_UINT64(a), C.BID_UINT64(b))), 0, bidgo.Bid64SameQuantum(a, b), 0
	case "quiet_equal":
		var flags C._IDEC_flags
		native := int(C.bid64_quiet_equal(C.BID_UINT64(a), C.BID_UINT64(b), &flags))
		exposed, exposedFlags := bidgo.Bid64QuietEqual(a, b)
		return native, uint32(flags), exposed, exposedFlags
	case "quiet_greater":
		var flags C._IDEC_flags
		native := int(C.bid64_quiet_greater(C.BID_UINT64(a), C.BID_UINT64(b), &flags))
		exposed, exposedFlags := bidgo.Bid64QuietGreater(a, b)
		return native, uint32(flags), exposed, exposedFlags
	case "quiet_greater_equal":
		var flags C._IDEC_flags
		native := int(C.bid64_quiet_greater_equal(C.BID_UINT64(a), C.BID_UINT64(b), &flags))
		exposed, exposedFlags := bidgo.Bid64QuietGreaterEqual(a, b)
		return native, uint32(flags), exposed, exposedFlags
	case "quiet_greater_unordered":
		var flags C._IDEC_flags
		native := int(C.bid64_quiet_greater_unordered(C.BID_UINT64(a), C.BID_UINT64(b), &flags))
		exposed, exposedFlags := bidgo.Bid64QuietGreaterUnordered(a, b)
		return native, uint32(flags), exposed, exposedFlags
	case "quiet_less":
		var flags C._IDEC_flags
		native := int(C.bid64_quiet_less(C.BID_UINT64(a), C.BID_UINT64(b), &flags))
		exposed, exposedFlags := bidgo.Bid64QuietLess(a, b)
		return native, uint32(flags), exposed, exposedFlags
	case "quiet_less_equal":
		var flags C._IDEC_flags
		native := int(C.bid64_quiet_less_equal(C.BID_UINT64(a), C.BID_UINT64(b), &flags))
		exposed, exposedFlags := bidgo.Bid64QuietLessEqual(a, b)
		return native, uint32(flags), exposed, exposedFlags
	case "quiet_less_unordered":
		var flags C._IDEC_flags
		native := int(C.bid64_quiet_less_unordered(C.BID_UINT64(a), C.BID_UINT64(b), &flags))
		exposed, exposedFlags := bidgo.Bid64QuietLessUnordered(a, b)
		return native, uint32(flags), exposed, exposedFlags
	case "quiet_not_equal":
		var flags C._IDEC_flags
		native := int(C.bid64_quiet_not_equal(C.BID_UINT64(a), C.BID_UINT64(b), &flags))
		exposed, exposedFlags := bidgo.Bid64QuietNotEqual(a, b)
		return native, uint32(flags), exposed, exposedFlags
	case "quiet_not_greater":
		var flags C._IDEC_flags
		native := int(C.bid64_quiet_not_greater(C.BID_UINT64(a), C.BID_UINT64(b), &flags))
		exposed, exposedFlags := bidgo.Bid64QuietNotGreater(a, b)
		return native, uint32(flags), exposed, exposedFlags
	case "quiet_not_less":
		var flags C._IDEC_flags
		native := int(C.bid64_quiet_not_less(C.BID_UINT64(a), C.BID_UINT64(b), &flags))
		exposed, exposedFlags := bidgo.Bid64QuietNotLess(a, b)
		return native, uint32(flags), exposed, exposedFlags
	case "quiet_ordered":
		var flags C._IDEC_flags
		native := int(C.bid64_quiet_ordered(C.BID_UINT64(a), C.BID_UINT64(b), &flags))
		exposed, exposedFlags := bidgo.Bid64QuietOrdered(a, b)
		return native, uint32(flags), exposed, exposedFlags
	case "quiet_unordered":
		var flags C._IDEC_flags
		native := int(C.bid64_quiet_unordered(C.BID_UINT64(a), C.BID_UINT64(b), &flags))
		exposed, exposedFlags := bidgo.Bid64QuietUnordered(a, b)
		return native, uint32(flags), exposed, exposedFlags
	case "signaling_greater":
		var flags C._IDEC_flags
		native := int(C.bid64_signaling_greater(C.BID_UINT64(a), C.BID_UINT64(b), &flags))
		exposed, exposedFlags := bidgo.Bid64SignalingGreater(a, b)
		return native, uint32(flags), exposed, exposedFlags
	case "signaling_greater_equal":
		var flags C._IDEC_flags
		native := int(C.bid64_signaling_greater_equal(C.BID_UINT64(a), C.BID_UINT64(b), &flags))
		exposed, exposedFlags := bidgo.Bid64SignalingGreaterEqual(a, b)
		return native, uint32(flags), exposed, exposedFlags
	case "signaling_greater_unordered":
		var flags C._IDEC_flags
		native := int(C.bid64_signaling_greater_unordered(C.BID_UINT64(a), C.BID_UINT64(b), &flags))
		exposed, exposedFlags := bidgo.Bid64SignalingGreaterUnordered(a, b)
		return native, uint32(flags), exposed, exposedFlags
	case "signaling_less":
		var flags C._IDEC_flags
		native := int(C.bid64_signaling_less(C.BID_UINT64(a), C.BID_UINT64(b), &flags))
		exposed, exposedFlags := bidgo.Bid64SignalingLess(a, b)
		return native, uint32(flags), exposed, exposedFlags
	case "signaling_less_equal":
		var flags C._IDEC_flags
		native := int(C.bid64_signaling_less_equal(C.BID_UINT64(a), C.BID_UINT64(b), &flags))
		exposed, exposedFlags := bidgo.Bid64SignalingLessEqual(a, b)
		return native, uint32(flags), exposed, exposedFlags
	case "signaling_less_unordered":
		var flags C._IDEC_flags
		native := int(C.bid64_signaling_less_unordered(C.BID_UINT64(a), C.BID_UINT64(b), &flags))
		exposed, exposedFlags := bidgo.Bid64SignalingLessUnordered(a, b)
		return native, uint32(flags), exposed, exposedFlags
	case "signaling_not_greater":
		var flags C._IDEC_flags
		native := int(C.bid64_signaling_not_greater(C.BID_UINT64(a), C.BID_UINT64(b), &flags))
		exposed, exposedFlags := bidgo.Bid64SignalingNotGreater(a, b)
		return native, uint32(flags), exposed, exposedFlags
	case "signaling_not_less":
		var flags C._IDEC_flags
		native := int(C.bid64_signaling_not_less(C.BID_UINT64(a), C.BID_UINT64(b), &flags))
		exposed, exposedFlags := bidgo.Bid64SignalingNotLess(a, b)
		return native, uint32(flags), exposed, exposedFlags
	default:
		panic("unsupported 64-bit ffi int binary operation")
	}
}

func runGeneratedFFICase128IntUnary(operation string, a Decimal128BID) (int, uint32, int, uint32) {
	ca := ffiUint128ToC(a)
	ga := decimal128BIDAsBidgo(a)
	switch operation {
	case "class":
		return int(C.bid128_class(ca)), 0, bidgo.Bid128Class(ga), 0
	case "isSigned":
		return int(C.bid128_isSigned(ca)), 0, bidgo.Bid128IsSigned(ga), 0
	case "isNormal":
		return int(C.bid128_isNormal(ca)), 0, bidgo.Bid128IsNormal(ga), 0
	case "isSubnormal":
		return int(C.bid128_isSubnormal(ca)), 0, bidgo.Bid128IsSubnormal(ga), 0
	case "isFinite":
		return int(C.bid128_isFinite(ca)), 0, bidgo.Bid128IsFinite(ga), 0
	case "isZero":
		return int(C.bid128_isZero(ca)), 0, bidgo.Bid128IsZero(ga), 0
	case "isInf":
		return int(C.bid128_isInf(ca)), 0, bidgo.Bid128IsInf(ga), 0
	case "isNaN":
		return int(C.bid128_isNaN(ca)), 0, bidgo.Bid128IsNaN(ga), 0
	case "isSignaling":
		return int(C.bid128_isSignaling(ca)), 0, bidgo.Bid128IsSignaling(ga), 0
	case "isCanonical":
		return int(C.bid128_isCanonical(ca)), 0, bidgo.Bid128IsCanonical(ga), 0
	case "radix":
		return int(C.bid128_radix(ca)), 0, bidgo.Bid128Radix(), 0
	case "quantexp":
		var flags C._IDEC_flags
		native := int(C.bid128_quantexp(ca, &flags))
		var exposedFlags uint32
		exposed := bidgo.Bid128Quantexp(ga, &exposedFlags)
		return native, uint32(flags), int(exposed), exposedFlags
	case "ilogb":
		var flags C._IDEC_flags
		native := int(C.bid128_ilogb(ca, &flags))
		var exposedFlags uint32
		exposed := bidgo.Bid128Ilogb(ga, &exposedFlags)
		return native, uint32(flags), exposed, exposedFlags
	default:
		panic("unsupported 128-bit ffi int unary operation")
	}
}

func runGeneratedFFICase128IntBinary(operation string, a, b Decimal128BID) (int, uint32, int, uint32) {
	ca := ffiUint128ToC(a)
	cb := ffiUint128ToC(b)
	ga := decimal128BIDAsBidgo(a)
	gb := decimal128BIDAsBidgo(b)
	switch operation {
	case "totalOrder":
		return int(C.bid128_totalOrder(ca, cb)), 0, bidgo.Bid128TotalOrder(ga, gb), 0
	case "totalOrderMag":
		return int(C.bid128_totalOrderMag(ca, cb)), 0, bidgo.Bid128TotalOrderMag(ga, gb), 0
	case "sameQuantum":
		return int(C.bid128_sameQuantum(ca, cb)), 0, bidgo.Bid128SameQuantum(ga, gb), 0
	case "quiet_equal":
		var flags C._IDEC_flags
		native := int(C.bid128_quiet_equal(ca, cb, &flags))
		exposed, exposedFlags := bidgo.Bid128QuietEqual(ga, gb)
		return native, uint32(flags), exposed, exposedFlags
	case "quiet_greater":
		var flags C._IDEC_flags
		native := int(C.bid128_quiet_greater(ca, cb, &flags))
		exposed, exposedFlags := bidgo.Bid128QuietGreater(ga, gb)
		return native, uint32(flags), exposed, exposedFlags
	case "quiet_greater_equal":
		var flags C._IDEC_flags
		native := int(C.bid128_quiet_greater_equal(ca, cb, &flags))
		exposed, exposedFlags := bidgo.Bid128QuietGreaterEqual(ga, gb)
		return native, uint32(flags), exposed, exposedFlags
	case "quiet_greater_unordered":
		var flags C._IDEC_flags
		native := int(C.bid128_quiet_greater_unordered(ca, cb, &flags))
		exposed, exposedFlags := bidgo.Bid128QuietGreaterUnordered(ga, gb)
		return native, uint32(flags), exposed, exposedFlags
	case "quiet_less":
		var flags C._IDEC_flags
		native := int(C.bid128_quiet_less(ca, cb, &flags))
		exposed, exposedFlags := bidgo.Bid128QuietLess(ga, gb)
		return native, uint32(flags), exposed, exposedFlags
	case "quiet_less_equal":
		var flags C._IDEC_flags
		native := int(C.bid128_quiet_less_equal(ca, cb, &flags))
		exposed, exposedFlags := bidgo.Bid128QuietLessEqual(ga, gb)
		return native, uint32(flags), exposed, exposedFlags
	case "quiet_less_unordered":
		var flags C._IDEC_flags
		native := int(C.bid128_quiet_less_unordered(ca, cb, &flags))
		exposed, exposedFlags := bidgo.Bid128QuietLessUnordered(ga, gb)
		return native, uint32(flags), exposed, exposedFlags
	case "quiet_not_equal":
		var flags C._IDEC_flags
		native := int(C.bid128_quiet_not_equal(ca, cb, &flags))
		exposed, exposedFlags := bidgo.Bid128QuietNotEqual(ga, gb)
		return native, uint32(flags), exposed, exposedFlags
	case "quiet_not_greater":
		var flags C._IDEC_flags
		native := int(C.bid128_quiet_not_greater(ca, cb, &flags))
		exposed, exposedFlags := bidgo.Bid128QuietNotGreater(ga, gb)
		return native, uint32(flags), exposed, exposedFlags
	case "quiet_not_less":
		var flags C._IDEC_flags
		native := int(C.bid128_quiet_not_less(ca, cb, &flags))
		exposed, exposedFlags := bidgo.Bid128QuietNotLess(ga, gb)
		return native, uint32(flags), exposed, exposedFlags
	case "quiet_ordered":
		var flags C._IDEC_flags
		native := int(C.bid128_quiet_ordered(ca, cb, &flags))
		exposed, exposedFlags := bidgo.Bid128QuietOrdered(ga, gb)
		return native, uint32(flags), exposed, exposedFlags
	case "quiet_unordered":
		var flags C._IDEC_flags
		native := int(C.bid128_quiet_unordered(ca, cb, &flags))
		exposed, exposedFlags := bidgo.Bid128QuietUnordered(ga, gb)
		return native, uint32(flags), exposed, exposedFlags
	case "signaling_greater":
		var flags C._IDEC_flags
		native := int(C.bid128_signaling_greater(ca, cb, &flags))
		exposed, exposedFlags := bidgo.Bid128SignalingGreater(ga, gb)
		return native, uint32(flags), exposed, exposedFlags
	case "signaling_greater_equal":
		var flags C._IDEC_flags
		native := int(C.bid128_signaling_greater_equal(ca, cb, &flags))
		exposed, exposedFlags := bidgo.Bid128SignalingGreaterEqual(ga, gb)
		return native, uint32(flags), exposed, exposedFlags
	case "signaling_greater_unordered":
		var flags C._IDEC_flags
		native := int(C.bid128_signaling_greater_unordered(ca, cb, &flags))
		exposed, exposedFlags := bidgo.Bid128SignalingGreaterUnordered(ga, gb)
		return native, uint32(flags), exposed, exposedFlags
	case "signaling_less":
		var flags C._IDEC_flags
		native := int(C.bid128_signaling_less(ca, cb, &flags))
		exposed, exposedFlags := bidgo.Bid128SignalingLess(ga, gb)
		return native, uint32(flags), exposed, exposedFlags
	case "signaling_less_equal":
		var flags C._IDEC_flags
		native := int(C.bid128_signaling_less_equal(ca, cb, &flags))
		exposed, exposedFlags := bidgo.Bid128SignalingLessEqual(ga, gb)
		return native, uint32(flags), exposed, exposedFlags
	case "signaling_less_unordered":
		var flags C._IDEC_flags
		native := int(C.bid128_signaling_less_unordered(ca, cb, &flags))
		exposed, exposedFlags := bidgo.Bid128SignalingLessUnordered(ga, gb)
		return native, uint32(flags), exposed, exposedFlags
	case "signaling_not_greater":
		var flags C._IDEC_flags
		native := int(C.bid128_signaling_not_greater(ca, cb, &flags))
		exposed, exposedFlags := bidgo.Bid128SignalingNotGreater(ga, gb)
		return native, uint32(flags), exposed, exposedFlags
	case "signaling_not_less":
		var flags C._IDEC_flags
		native := int(C.bid128_signaling_not_less(ca, cb, &flags))
		exposed, exposedFlags := bidgo.Bid128SignalingNotLess(ga, gb)
		return native, uint32(flags), exposed, exposedFlags
	default:
		panic("unsupported 128-bit ffi int binary operation")
	}
}

func runGeneratedFFICase32Unary(function string, a uint32, rounding int) (string, string) {
	switch function {
	case "bid32_round_integral_exact":
		var flags C._IDEC_flags
		native := uint32(C.bid32_round_integral_exact(C.BID_UINT32(a), generatedFFICRoundingMode(rounding), &flags))
		exposed, exposedFlags := bidgo.Bid32RoundIntegralExact(a, rounding)
		return fmt.Sprintf("%08x/%08x", native, uint32(flags)), fmt.Sprintf("%08x/%08x", exposed, exposedFlags)
	case "bid32_copy":
		return fmt.Sprintf("%08x", uint32(C.bid32_copy(C.BID_UINT32(a)))), fmt.Sprintf("%08x", bidgo.Bid32Copy(a))
	case "bid32_negate":
		return fmt.Sprintf("%08x", uint32(C.bid32_negate(C.BID_UINT32(a)))), fmt.Sprintf("%08x", bidgo.Bid32Negate(a))
	case "bid32_abs":
		return fmt.Sprintf("%08x", uint32(C.bid32_abs(C.BID_UINT32(a)))), fmt.Sprintf("%08x", bidgo.Bid32Abs(a))
	case "bid32_sqrt":
		var flags C._IDEC_flags
		native := uint32(C.bid32_sqrt(C.BID_UINT32(a), generatedFFICRoundingMode(rounding), &flags))
		exposed, exposedFlags := bidgo.Bid32Sqrt(a, rounding)
		return fmt.Sprintf("%08x/%08x", native, uint32(flags)), fmt.Sprintf("%08x/%08x", exposed, exposedFlags)
	case "bid32_logb":
		var flags C._IDEC_flags
		native := uint32(C.bid32_logb(C.BID_UINT32(a), &flags))
		exposed, exposedFlags := bidgo.Bid32Logb(a)
		return fmt.Sprintf("%08x/%08x", native, uint32(flags)), fmt.Sprintf("%08x/%08x", exposed, exposedFlags)
	case "bid32_nextup":
		var flags C._IDEC_flags
		native := uint32(C.bid32_nextup(C.BID_UINT32(a), &flags))
		exposed, exposedFlags := bidgo.Bid32NextUp(a)
		return fmt.Sprintf("%08x/%08x", native, uint32(flags)), fmt.Sprintf("%08x/%08x", exposed, exposedFlags)
	case "bid32_nextdown":
		var flags C._IDEC_flags
		native := uint32(C.bid32_nextdown(C.BID_UINT32(a), &flags))
		exposed, exposedFlags := bidgo.Bid32NextDown(a)
		return fmt.Sprintf("%08x/%08x", native, uint32(flags)), fmt.Sprintf("%08x/%08x", exposed, exposedFlags)
	default:
		panic("unsupported 32-bit ffi unary function")
	}
}

func runGeneratedFFICase32Binary(function string, a uint32, b uint32, rounding int) (string, string) {
	switch function {
	case "bid32_add":
		var flags C._IDEC_flags
		native := uint32(C.bid32_add(C.BID_UINT32(a), C.BID_UINT32(b), generatedFFICRoundingMode(rounding), &flags))
		exposed, exposedFlags := bidgo.Bid32AddWithFlags(a, b, rounding)
		return fmt.Sprintf("%08x/%08x", native, uint32(flags)), fmt.Sprintf("%08x/%08x", exposed, exposedFlags)
	case "bid32_sub":
		var flags C._IDEC_flags
		native := uint32(C.bid32_sub(C.BID_UINT32(a), C.BID_UINT32(b), generatedFFICRoundingMode(rounding), &flags))
		exposed, exposedFlags := bidgo.Bid32SubWithFlags(a, b, rounding)
		return fmt.Sprintf("%08x/%08x", native, uint32(flags)), fmt.Sprintf("%08x/%08x", exposed, exposedFlags)
	case "bid32_mul":
		var flags C._IDEC_flags
		native := uint32(C.bid32_mul(C.BID_UINT32(a), C.BID_UINT32(b), generatedFFICRoundingMode(rounding), &flags))
		exposed, exposedFlags := bidgo.Bid32MulWithFlags(a, b, rounding)
		return fmt.Sprintf("%08x/%08x", native, uint32(flags)), fmt.Sprintf("%08x/%08x", exposed, exposedFlags)
	case "bid32_div":
		var flags C._IDEC_flags
		native := uint32(C.bid32_div(C.BID_UINT32(a), C.BID_UINT32(b), generatedFFICRoundingMode(rounding), &flags))
		exposed, exposedFlags := bidgo.Bid32DivWithFlags(a, b, rounding)
		return fmt.Sprintf("%08x/%08x", native, uint32(flags)), fmt.Sprintf("%08x/%08x", exposed, exposedFlags)
	case "bid32_quantize":
		var flags C._IDEC_flags
		native := uint32(C.bid32_quantize(C.BID_UINT32(a), C.BID_UINT32(b), generatedFFICRoundingMode(rounding), &flags))
		exposed, exposedFlags := bidgo.Bid32Quantize(a, b, rounding)
		return fmt.Sprintf("%08x/%08x", native, uint32(flags)), fmt.Sprintf("%08x/%08x", exposed, exposedFlags)
	case "bid32_minnum":
		var flags C._IDEC_flags
		native := uint32(C.bid32_minnum(C.BID_UINT32(a), C.BID_UINT32(b), &flags))
		exposed, exposedFlags := bidgo.Bid32MinNumWithFlags(a, b)
		return fmt.Sprintf("%08x/%08x", native, uint32(flags)), fmt.Sprintf("%08x/%08x", exposed, exposedFlags)
	case "bid32_maxnum":
		var flags C._IDEC_flags
		native := uint32(C.bid32_maxnum(C.BID_UINT32(a), C.BID_UINT32(b), &flags))
		exposed, exposedFlags := bidgo.Bid32MaxNumWithFlags(a, b)
		return fmt.Sprintf("%08x/%08x", native, uint32(flags)), fmt.Sprintf("%08x/%08x", exposed, exposedFlags)
	case "bid32_minnum_mag":
		var flags C._IDEC_flags
		native := uint32(C.bid32_minnum_mag(C.BID_UINT32(a), C.BID_UINT32(b), &flags))
		exposed, exposedFlags := bidgo.Bid32MinNumMagWithFlags(a, b)
		return fmt.Sprintf("%08x/%08x", native, uint32(flags)), fmt.Sprintf("%08x/%08x", exposed, exposedFlags)
	case "bid32_maxnum_mag":
		var flags C._IDEC_flags
		native := uint32(C.bid32_maxnum_mag(C.BID_UINT32(a), C.BID_UINT32(b), &flags))
		exposed, exposedFlags := bidgo.Bid32MaxNumMagWithFlags(a, b)
		return fmt.Sprintf("%08x/%08x", native, uint32(flags)), fmt.Sprintf("%08x/%08x", exposed, exposedFlags)
	case "bid32_copySign":
		return fmt.Sprintf("%08x", uint32(C.bid32_copySign(C.BID_UINT32(a), C.BID_UINT32(b)))), fmt.Sprintf("%08x", bidgo.Bid32CopySign(a, b))
	case "bid32_rem":
		var flags C._IDEC_flags
		native := uint32(C.bid32_rem(C.BID_UINT32(a), C.BID_UINT32(b), &flags))
		exposed, exposedFlags := bidgo.Bid32Rem(a, b)
		return fmt.Sprintf("%08x/%08x", native, uint32(flags)), fmt.Sprintf("%08x/%08x", exposed, exposedFlags)
	case "bid32_fmod":
		var flags C._IDEC_flags
		native := uint32(C.bid32_fmod(C.BID_UINT32(a), C.BID_UINT32(b), &flags))
		exposed, exposedFlags := bidgo.Bid32Fmod(a, b)
		return fmt.Sprintf("%08x/%08x", native, uint32(flags)), fmt.Sprintf("%08x/%08x", exposed, exposedFlags)
	default:
		panic("unsupported 32-bit ffi binary function")
	}
}

func runGeneratedFFICase32Int(function string, a uint32, n int, rounding int) (string, string) {
	switch function {
	case "bid32_scalbn":
		var flags C._IDEC_flags
		native := uint32(C.bid32_scalbn(C.BID_UINT32(a), C.int(n), generatedFFICRoundingMode(rounding), &flags))
		exposed, exposedFlags := bidgo.Bid32ScalbnWithFlags(a, n, rounding)
		return fmt.Sprintf("%08x/%08x", native, uint32(flags)), fmt.Sprintf("%08x/%08x", exposed, exposedFlags)
	default:
		panic("unsupported 32-bit ffi int function")
	}
}

func runGeneratedFFICase32Ternary(function string, a, b, c uint32, rounding int) (string, string) {
	switch function {
	case "bid32_fma":
		var flags C._IDEC_flags
		native := uint32(C.bid32_fma(C.BID_UINT32(a), C.BID_UINT32(b), C.BID_UINT32(c), generatedFFICRoundingMode(rounding), &flags))
		exposed, exposedFlags := bidgo.Bid32Fma(a, b, c, rounding)
		return fmt.Sprintf("%08x/%08x", native, uint32(flags)), fmt.Sprintf("%08x/%08x", exposed, exposedFlags)
	default:
		panic("unsupported 32-bit ffi ternary function")
	}
}

func runGeneratedFFICase64Unary(function string, a uint64, rounding int) (string, string) {
	switch function {
	case "bid64_round_integral_exact":
		var flags C._IDEC_flags
		native := uint64(C.bid64_round_integral_exact(C.BID_UINT64(a), generatedFFICRoundingMode(rounding), &flags))
		exposed, exposedFlags := bidgo.Bid64RoundIntegralExact(a, rounding)
		return fmt.Sprintf("%016x/%08x", native, uint32(flags)), fmt.Sprintf("%016x/%08x", exposed, exposedFlags)
	case "bid64_copy":
		return fmt.Sprintf("%016x", uint64(C.bid64_copy(C.BID_UINT64(a)))), fmt.Sprintf("%016x", bidgo.Bid64Copy(a))
	case "bid64_negate":
		return fmt.Sprintf("%016x", uint64(C.bid64_negate(C.BID_UINT64(a)))), fmt.Sprintf("%016x", bidgo.Bid64Negate(a))
	case "bid64_abs":
		return fmt.Sprintf("%016x", uint64(C.bid64_abs(C.BID_UINT64(a)))), fmt.Sprintf("%016x", bidgo.Bid64Abs(a))
	case "bid64_sqrt":
		var flags C._IDEC_flags
		native := uint64(C.bid64_sqrt(C.BID_UINT64(a), generatedFFICRoundingMode(rounding), &flags))
		exposed, exposedFlags := bidgo.Bid64Sqrt(a, rounding)
		return fmt.Sprintf("%016x/%08x", native, uint32(flags)), fmt.Sprintf("%016x/%08x", exposed, exposedFlags)
	case "bid64_logb":
		var flags C._IDEC_flags
		native := uint64(C.bid64_logb(C.BID_UINT64(a), &flags))
		exposed, exposedFlags := bidgo.Bid64Logb(a)
		return fmt.Sprintf("%016x/%08x", native, uint32(flags)), fmt.Sprintf("%016x/%08x", exposed, exposedFlags)
	case "bid64_nextup":
		var flags C._IDEC_flags
		native := uint64(C.bid64_nextup(C.BID_UINT64(a), &flags))
		exposed, exposedFlags := bidgo.Bid64NextUp(a)
		return fmt.Sprintf("%016x/%08x", native, uint32(flags)), fmt.Sprintf("%016x/%08x", exposed, exposedFlags)
	case "bid64_nextdown":
		var flags C._IDEC_flags
		native := uint64(C.bid64_nextdown(C.BID_UINT64(a), &flags))
		exposed, exposedFlags := bidgo.Bid64NextDown(a)
		return fmt.Sprintf("%016x/%08x", native, uint32(flags)), fmt.Sprintf("%016x/%08x", exposed, exposedFlags)
	default:
		panic("unsupported 64-bit ffi unary function")
	}
}

func runGeneratedFFICase64Binary(function string, a uint64, b uint64, rounding int) (string, string) {
	switch function {
	case "bid64_add":
		var flags C._IDEC_flags
		native := uint64(C.bid64_add(C.BID_UINT64(a), C.BID_UINT64(b), generatedFFICRoundingMode(rounding), &flags))
		exposed, exposedFlags := bidgo.Bid64AddWithFlags(a, b, rounding)
		return fmt.Sprintf("%016x/%08x", native, uint32(flags)), fmt.Sprintf("%016x/%08x", exposed, exposedFlags)
	case "bid64_sub":
		var flags C._IDEC_flags
		native := uint64(C.bid64_sub(C.BID_UINT64(a), C.BID_UINT64(b), generatedFFICRoundingMode(rounding), &flags))
		exposed, exposedFlags := bidgo.Bid64SubWithFlags(a, b, rounding)
		return fmt.Sprintf("%016x/%08x", native, uint32(flags)), fmt.Sprintf("%016x/%08x", exposed, exposedFlags)
	case "bid64_mul":
		var flags C._IDEC_flags
		native := uint64(C.bid64_mul(C.BID_UINT64(a), C.BID_UINT64(b), generatedFFICRoundingMode(rounding), &flags))
		exposed, exposedFlags := bidgo.Bid64MulWithFlags(a, b, rounding)
		return fmt.Sprintf("%016x/%08x", native, uint32(flags)), fmt.Sprintf("%016x/%08x", exposed, exposedFlags)
	case "bid64_div":
		var flags C._IDEC_flags
		native := uint64(C.bid64_div(C.BID_UINT64(a), C.BID_UINT64(b), generatedFFICRoundingMode(rounding), &flags))
		exposed, exposedFlags := bidgo.Bid64DivWithFlags(a, b, rounding)
		return fmt.Sprintf("%016x/%08x", native, uint32(flags)), fmt.Sprintf("%016x/%08x", exposed, exposedFlags)
	case "bid64_quantize":
		var flags C._IDEC_flags
		native := uint64(C.bid64_quantize(C.BID_UINT64(a), C.BID_UINT64(b), generatedFFICRoundingMode(rounding), &flags))
		exposed, exposedFlags := bidgo.Bid64Quantize(a, b, rounding)
		return fmt.Sprintf("%016x/%08x", native, uint32(flags)), fmt.Sprintf("%016x/%08x", exposed, exposedFlags)
	case "bid64_minnum":
		var flags C._IDEC_flags
		native := uint64(C.bid64_minnum(C.BID_UINT64(a), C.BID_UINT64(b), &flags))
		exposed, exposedFlags := bidgo.Bid64MinNum(a, b)
		return fmt.Sprintf("%016x/%08x", native, uint32(flags)), fmt.Sprintf("%016x/%08x", exposed, exposedFlags)
	case "bid64_maxnum":
		var flags C._IDEC_flags
		native := uint64(C.bid64_maxnum(C.BID_UINT64(a), C.BID_UINT64(b), &flags))
		exposed, exposedFlags := bidgo.Bid64MaxNum(a, b)
		return fmt.Sprintf("%016x/%08x", native, uint32(flags)), fmt.Sprintf("%016x/%08x", exposed, exposedFlags)
	case "bid64_minnum_mag":
		var flags C._IDEC_flags
		native := uint64(C.bid64_minnum_mag(C.BID_UINT64(a), C.BID_UINT64(b), &flags))
		exposed, exposedFlags := bidgo.Bid64MinNumMag(a, b)
		return fmt.Sprintf("%016x/%08x", native, uint32(flags)), fmt.Sprintf("%016x/%08x", exposed, exposedFlags)
	case "bid64_maxnum_mag":
		var flags C._IDEC_flags
		native := uint64(C.bid64_maxnum_mag(C.BID_UINT64(a), C.BID_UINT64(b), &flags))
		exposed, exposedFlags := bidgo.Bid64MaxNumMag(a, b)
		return fmt.Sprintf("%016x/%08x", native, uint32(flags)), fmt.Sprintf("%016x/%08x", exposed, exposedFlags)
	case "bid64_copySign":
		return fmt.Sprintf("%016x", uint64(C.bid64_copySign(C.BID_UINT64(a), C.BID_UINT64(b)))), fmt.Sprintf("%016x", bidgo.Bid64CopySign(a, b))
	case "bid64_rem":
		var flags C._IDEC_flags
		native := uint64(C.bid64_rem(C.BID_UINT64(a), C.BID_UINT64(b), &flags))
		exposed, exposedFlags := bidgo.Bid64Rem(a, b)
		return fmt.Sprintf("%016x/%08x", native, uint32(flags)), fmt.Sprintf("%016x/%08x", exposed, exposedFlags)
	case "bid64_fmod":
		var flags C._IDEC_flags
		native := uint64(C.bid64_fmod(C.BID_UINT64(a), C.BID_UINT64(b), &flags))
		exposed, exposedFlags := bidgo.Bid64Fmod(a, b)
		return fmt.Sprintf("%016x/%08x", native, uint32(flags)), fmt.Sprintf("%016x/%08x", exposed, exposedFlags)
	default:
		panic("unsupported 64-bit ffi binary function")
	}
}

func runGeneratedFFICase64Int(function string, a uint64, n int, rounding int) (string, string) {
	switch function {
	case "bid64_scalbn":
		var flags C._IDEC_flags
		native := uint64(C.bid64_scalbn(C.BID_UINT64(a), C.int(n), generatedFFICRoundingMode(rounding), &flags))
		exposed, exposedFlags := bidgo.Bid64Scalbn(a, n, rounding)
		return fmt.Sprintf("%016x/%08x", native, uint32(flags)), fmt.Sprintf("%016x/%08x", exposed, exposedFlags)
	default:
		panic("unsupported 64-bit ffi int function")
	}
}

func runGeneratedFFICase64Ternary(function string, a, b, c uint64, rounding int) (string, string) {
	switch function {
	case "bid64_fma":
		var flags C._IDEC_flags
		native := uint64(C.bid64_fma(C.BID_UINT64(a), C.BID_UINT64(b), C.BID_UINT64(c), generatedFFICRoundingMode(rounding), &flags))
		exposed, exposedFlags := bidgo.Bid64Fma(a, b, c, rounding)
		return fmt.Sprintf("%016x/%08x", native, uint32(flags)), fmt.Sprintf("%016x/%08x", exposed, exposedFlags)
	default:
		panic("unsupported 64-bit ffi ternary function")
	}
}

func runGeneratedFFICase128Unary(function string, a Decimal128BID, rounding int) (string, string) {
	ca := ffiUint128ToC(a)
	ga := decimal128BIDAsBidgo(a)
	switch function {
	case "bid128_round_integral_exact":
		var flags C._IDEC_flags
		native := ffiUint128FromC(C.bid128_round_integral_exact(ca, generatedFFICRoundingMode(rounding), &flags))
		var exposedFlags uint32
		exposed := bidgo.Bid128RoundIntegralExact(ga, rounding, &exposedFlags)
		return fmt.Sprintf("%s/%08x", formatFFIUint128Bits(native), uint32(flags)), fmt.Sprintf("%s/%08x", formatFFIUint128Bits(decimal128BIDFromBidgo(exposed)), exposedFlags)
	case "bid128_copy":
		return formatFFIUint128Bits(ffiUint128FromC(C.bid128_copy(ca))), formatFFIUint128Bits(decimal128BIDFromBidgo(bidgo.Bid128Copy(ga)))
	case "bid128_negate":
		return formatFFIUint128Bits(ffiUint128FromC(C.bid128_negate(ca))), formatFFIUint128Bits(decimal128BIDFromBidgo(bidgo.Bid128Negate(ga)))
	case "bid128_abs":
		return formatFFIUint128Bits(ffiUint128FromC(C.bid128_abs(ca))), formatFFIUint128Bits(decimal128BIDFromBidgo(bidgo.Bid128Abs(ga)))
	case "bid128_sqrt":
		var flags C._IDEC_flags
		native := ffiUint128FromC(C.bid128_sqrt(ca, generatedFFICRoundingMode(rounding), &flags))
		exposed, exposedFlags := bidgo.Bid128Sqrt(ga, rounding)
		return fmt.Sprintf("%s/%08x", formatFFIUint128Bits(native), uint32(flags)), fmt.Sprintf("%s/%08x", formatFFIUint128Bits(decimal128BIDFromBidgo(exposed)), exposedFlags)
	case "bid128_logb":
		var flags C._IDEC_flags
		native := ffiUint128FromC(C.bid128_logb(ca, &flags))
		var exposedFlags uint32
		exposed := bidgo.Bid128Logb(ga, &exposedFlags)
		return fmt.Sprintf("%s/%08x", formatFFIUint128Bits(native), uint32(flags)), fmt.Sprintf("%s/%08x", formatFFIUint128Bits(decimal128BIDFromBidgo(exposed)), exposedFlags)
	case "bid128_nextup":
		var flags C._IDEC_flags
		native := ffiUint128FromC(C.bid128_nextup(ca, &flags))
		exposed, exposedFlags := bidgo.Bid128NextUp(ga)
		return fmt.Sprintf("%s/%08x", formatFFIUint128Bits(native), uint32(flags)), fmt.Sprintf("%s/%08x", formatFFIUint128Bits(decimal128BIDFromBidgo(exposed)), exposedFlags)
	case "bid128_nextdown":
		var flags C._IDEC_flags
		native := ffiUint128FromC(C.bid128_nextdown(ca, &flags))
		exposed, exposedFlags := bidgo.Bid128NextDown(ga)
		return fmt.Sprintf("%s/%08x", formatFFIUint128Bits(native), uint32(flags)), fmt.Sprintf("%s/%08x", formatFFIUint128Bits(decimal128BIDFromBidgo(exposed)), exposedFlags)
	default:
		panic("unsupported 128-bit ffi unary function")
	}
}

func runGeneratedFFICase128Binary(function string, a, b Decimal128BID, rounding int) (string, string) {
	ca := ffiUint128ToC(a)
	cb := ffiUint128ToC(b)
	ga := decimal128BIDAsBidgo(a)
	gb := decimal128BIDAsBidgo(b)
	switch function {
	case "bid128_add":
		var flags C._IDEC_flags
		native := ffiUint128FromC(C.bid128_add(ca, cb, generatedFFICRoundingMode(rounding), &flags))
		var exposedFlags uint32
		exposed := bidgo.Bid128Add(ga, gb, rounding, &exposedFlags)
		return fmt.Sprintf("%s/%08x", formatFFIUint128Bits(native), uint32(flags)), fmt.Sprintf("%s/%08x", formatFFIUint128Bits(decimal128BIDFromBidgo(exposed)), exposedFlags)
	case "bid128_sub":
		var flags C._IDEC_flags
		native := ffiUint128FromC(C.bid128_sub(ca, cb, generatedFFICRoundingMode(rounding), &flags))
		var exposedFlags uint32
		exposed := bidgo.Bid128Sub(ga, gb, rounding, &exposedFlags)
		return fmt.Sprintf("%s/%08x", formatFFIUint128Bits(native), uint32(flags)), fmt.Sprintf("%s/%08x", formatFFIUint128Bits(decimal128BIDFromBidgo(exposed)), exposedFlags)
	case "bid128_mul":
		var flags C._IDEC_flags
		native := ffiUint128FromC(C.bid128_mul(ca, cb, generatedFFICRoundingMode(rounding), &flags))
		exposed, exposedFlags := bidgo.Bid128Mul(ga, gb, rounding)
		return fmt.Sprintf("%s/%08x", formatFFIUint128Bits(native), uint32(flags)), fmt.Sprintf("%s/%08x", formatFFIUint128Bits(decimal128BIDFromBidgo(exposed)), exposedFlags)
	case "bid128_div":
		var flags C._IDEC_flags
		native := ffiUint128FromC(C.bid128_div(ca, cb, generatedFFICRoundingMode(rounding), &flags))
		exposed, exposedFlags := bidgo.Bid128Div(ga, gb, rounding)
		return fmt.Sprintf("%s/%08x", formatFFIUint128Bits(native), uint32(flags)), fmt.Sprintf("%s/%08x", formatFFIUint128Bits(decimal128BIDFromBidgo(exposed)), exposedFlags)
	case "bid128_quantize":
		var flags C._IDEC_flags
		native := ffiUint128FromC(C.bid128_quantize(ca, cb, generatedFFICRoundingMode(rounding), &flags))
		exposed, exposedFlags := bidgo.Bid128Quantize(ga, gb, rounding)
		return fmt.Sprintf("%s/%08x", formatFFIUint128Bits(native), uint32(flags)), fmt.Sprintf("%s/%08x", formatFFIUint128Bits(decimal128BIDFromBidgo(exposed)), exposedFlags)
	case "bid128_minnum":
		var flags C._IDEC_flags
		native := ffiUint128FromC(C.bid128_minnum(ca, cb, &flags))
		var exposedFlags uint32
		exposed := bidgo.Bid128Minnum(ga, gb, &exposedFlags)
		return fmt.Sprintf("%s/%08x", formatFFIUint128Bits(native), uint32(flags)), fmt.Sprintf("%s/%08x", formatFFIUint128Bits(decimal128BIDFromBidgo(exposed)), exposedFlags)
	case "bid128_maxnum":
		var flags C._IDEC_flags
		native := ffiUint128FromC(C.bid128_maxnum(ca, cb, &flags))
		var exposedFlags uint32
		exposed := bidgo.Bid128Maxnum(ga, gb, &exposedFlags)
		return fmt.Sprintf("%s/%08x", formatFFIUint128Bits(native), uint32(flags)), fmt.Sprintf("%s/%08x", formatFFIUint128Bits(decimal128BIDFromBidgo(exposed)), exposedFlags)
	case "bid128_minnum_mag":
		var flags C._IDEC_flags
		native := ffiUint128FromC(C.bid128_minnum_mag(ca, cb, &flags))
		var exposedFlags uint32
		exposed := bidgo.Bid128MinnumMag(ga, gb, &exposedFlags)
		return fmt.Sprintf("%s/%08x", formatFFIUint128Bits(native), uint32(flags)), fmt.Sprintf("%s/%08x", formatFFIUint128Bits(decimal128BIDFromBidgo(exposed)), exposedFlags)
	case "bid128_maxnum_mag":
		var flags C._IDEC_flags
		native := ffiUint128FromC(C.bid128_maxnum_mag(ca, cb, &flags))
		var exposedFlags uint32
		exposed := bidgo.Bid128MaxnumMag(ga, gb, &exposedFlags)
		return fmt.Sprintf("%s/%08x", formatFFIUint128Bits(native), uint32(flags)), fmt.Sprintf("%s/%08x", formatFFIUint128Bits(decimal128BIDFromBidgo(exposed)), exposedFlags)
	case "bid128_copySign":
		return formatFFIUint128Bits(ffiUint128FromC(C.bid128_copySign(ca, cb))), formatFFIUint128Bits(decimal128BIDFromBidgo(bidgo.Bid128CopySign(ga, gb)))
	case "bid128_rem":
		var flags C._IDEC_flags
		native := ffiUint128FromC(C.bid128_rem(ca, cb, &flags))
		exposed, exposedFlags := bidgo.Bid128Rem(ga, gb)
		return fmt.Sprintf("%s/%08x", formatFFIUint128Bits(native), uint32(flags)), fmt.Sprintf("%s/%08x", formatFFIUint128Bits(decimal128BIDFromBidgo(exposed)), exposedFlags)
	case "bid128_fmod":
		var flags C._IDEC_flags
		native := ffiUint128FromC(C.bid128_fmod(ca, cb, &flags))
		exposed, exposedFlags := bidgo.Bid128Fmod(ga, gb)
		return fmt.Sprintf("%s/%08x", formatFFIUint128Bits(native), uint32(flags)), fmt.Sprintf("%s/%08x", formatFFIUint128Bits(decimal128BIDFromBidgo(exposed)), exposedFlags)
	default:
		panic("unsupported 128-bit ffi binary function")
	}
}

func runGeneratedFFICase128Int(function string, a Decimal128BID, n int, rounding int) (string, string) {
	ca := ffiUint128ToC(a)
	ga := decimal128BIDAsBidgo(a)
	switch function {
	case "bid128_scalbn":
		var flags C._IDEC_flags
		native := ffiUint128FromC(C.bid128_scalbn(ca, C.int(n), generatedFFICRoundingMode(rounding), &flags))
		var exposedFlags uint32
		exposed := bidgo.Bid128Scalbn(ga, n, rounding, &exposedFlags)
		return fmt.Sprintf("%s/%08x", formatFFIUint128Bits(native), uint32(flags)), fmt.Sprintf("%s/%08x", formatFFIUint128Bits(decimal128BIDFromBidgo(exposed)), exposedFlags)
	default:
		panic("unsupported 128-bit ffi int function")
	}
}

func runGeneratedFFICase128Ternary(function string, a, b, c Decimal128BID, rounding int) (string, string) {
	ca := ffiUint128ToC(a)
	cb := ffiUint128ToC(b)
	cc := ffiUint128ToC(c)
	ga := decimal128BIDAsBidgo(a)
	gb := decimal128BIDAsBidgo(b)
	gc := decimal128BIDAsBidgo(c)
	switch function {
	case "bid128_fma":
		var flags C._IDEC_flags
		native := ffiUint128FromC(C.bid128_fma(ca, cb, cc, generatedFFICRoundingMode(rounding), &flags))
		exposed, exposedFlags := bidgo.Bid128Fma(ga, gb, gc, rounding)
		return fmt.Sprintf("%s/%08x", formatFFIUint128Bits(native), uint32(flags)), fmt.Sprintf("%s/%08x", formatFFIUint128Bits(decimal128BIDFromBidgo(exposed)), exposedFlags)
	default:
		panic("unsupported 128-bit ffi ternary function")
	}
}
`),
		ffiGeneratedNativeTestPath: []byte(ffiNativeTestSource(counts)),
		ffiGeneratedStubTestPath: []byte(genmarker.Line("testgen") + `
//go:build !cgo || !bid754_native

package bid754

import "testing"

func TestGeneratedFFIBitCompareSubset(t *testing.T) {
	t.Skip("generated ffi bit-compare cases require cgo and bid754_native")
}
`),
	})
}
