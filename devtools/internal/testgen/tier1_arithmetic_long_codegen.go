package testgen

import (
	"embed"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/sky1core/bid754/devtools/internal/genmarker"
)

const (
	tier1ArithmeticLongGeneratedPath             = "../bid754-go/generated_ffi_bitcompare_tier1_arithmetic_long_test.go"
	tier1ArithmeticRustLongGeneratedPath         = "../bid754-rs/ffi-verify/tests/tier1_arithmetic_long_generated.rs"
	tier1ArithmeticScaleFiniteTransitionLimit32  = int64(197)
	tier1ArithmeticScaleFiniteTransitionLimit64  = int64(782)
	tier1ArithmeticScaleFiniteTransitionLimit128 = int64(12320)
	tier1ArithmeticScaleRandomStrata             = uint64(4)
	tier1ArithmeticScaleModeCross                = uint64(5)
	tier1ArithmeticScaleTupleHashOffset          = uint64(0xcbf29ce484222325)
	tier1ArithmeticScaleTupleHashPrime           = uint64(0x100000001b3)
)

//go:embed tier1_arithmetic_templates/*
var tier1ArithmeticTemplates embed.FS

// tier1ArithmeticProbes* are the single source for the probe operand sets
// injected into both language runners. The generated stream hashes below are
// computed over these values plus the boundary corpus, so a probe edit that
// reaches only one consumer fails the anchored stream-hash contract.
var tier1ArithmeticProbes32Values = []uint32{
	0x00000000, 0x80000000, 0x32800001, 0xb2800001, 0x00000001, 0x77f8967f,
	0xf7f8967f, 0x78000000, 0x7c000001, 0x7e000001, 0x60000000, 0x5f800000,
}

var tier1ArithmeticProbes64Values = []uint64{
	0x0000000000000000, 0x8000000000000000, 0x31c0000000000001, 0xb1c0000000000001,
	0x0000000000000001, 0x77fb86f26fc0ffff, 0xf7fb86f26fc0ffff, 0x7800000000000000,
	0x7c00000000000001, 0x7e00000000000001, 0x6000000000000000, 0x5fe0000000000000,
}

var tier1ArithmeticProbes128Values = []bid128BidCodecValue{
	{lo: 0x0000000000000000, hi: 0x0000000000000000},
	{lo: 0x0000000000000000, hi: 0x8000000000000000},
	{lo: 0x0000000000000001, hi: 0x3040000000000000},
	{lo: 0x0000000000000001, hi: 0xb040000000000000},
	{lo: 0x0000000000000001, hi: 0x0000000000000000},
	{lo: 0x378d8e63ffffffff, hi: 0x5fffed09bead87c0},
	{lo: 0x378d8e63ffffffff, hi: 0xdfffed09bead87c0},
	{lo: 0x0000000000000000, hi: 0x7800000000000000},
	{lo: 0x0000000000000001, hi: 0x7c00000000000000},
	{lo: 0x0000000000000001, hi: 0x7e00000000000000},
	{lo: 0x0000000000000000, hi: 0x6000000000000000},
	{lo: 0x0000000000000000, hi: 0x5ffe000000000000},
}

func WriteTier1ArithmeticLongOutputs(repoRoot string) error {
	files, err := GenerateTier1ArithmeticLongOutputs()
	if err != nil {
		return err
	}
	for path, data := range files {
		fullPath := filepath.Join(repoRoot, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			return fmt.Errorf("mkdir %q: %w", filepath.Dir(fullPath), err)
		}
		if err := os.WriteFile(fullPath, data, 0o644); err != nil {
			return fmt.Errorf("write generated Tier 1 arithmetic long test %q: %w", fullPath, err)
		}
	}
	return nil
}

func GenerateTier1ArithmeticLongOutputs() (map[string][]byte, error) {
	template, err := tier1ArithmeticTemplates.ReadFile("tier1_arithmetic_templates/go_tier1_arithmetic_long_test.go.tmpl")
	if err != nil {
		return nil, fmt.Errorf("read Tier 1 arithmetic long template: %w", err)
	}

	boundary32 := tier1ArithmeticBoundary32Values()
	boundary64 := bid64BidCodecEdgeValues()
	boundary128 := bid128BidCodecEdgeValues()
	semantic, err := tier1ArithmeticSemanticCorpus()
	if err != nil {
		return nil, err
	}
	rustTemplate, err := tier1ArithmeticTemplates.ReadFile("tier1_arithmetic_templates/rust_tier1_arithmetic_long.rs")
	if err != nil {
		return nil, fmt.Errorf("read Rust Tier 1 arithmetic long template: %w", err)
	}
	counts := tier1ArithmeticCountsFor(
		uint64(len(boundary32)), uint64(len(boundary64)), uint64(len(boundary128)),
		uint64(len(semantic.rounded32)), uint64(len(semantic.rounded64)), uint64(len(semantic.rounded128)),
		uint64(len(semantic.scale32)), uint64(len(semantic.scale64)), uint64(len(semantic.scale128)),
		uint64(len(semantic.remainder32)), uint64(len(semantic.remainder64)), uint64(len(semantic.remainder128)),
		uint64(len(semantic.fma32)), uint64(len(semantic.fma64)), uint64(len(semantic.fma128)),
		uint64(len(semantic.sqrt32)), uint64(len(semantic.sqrt64)), uint64(len(semantic.sqrt128)),
	)
	boundaryWords32 := make([]bid128BidCodecValue, len(boundary32))
	for i, value := range boundary32 {
		boundaryWords32[i] = bid128BidCodecValue{lo: uint64(value)}
	}
	boundaryWords64 := make([]bid128BidCodecValue, len(boundary64))
	for i, value := range boundary64 {
		boundaryWords64[i] = bid128BidCodecValue{lo: value}
	}
	probeWords32 := make([]bid128BidCodecValue, len(tier1ArithmeticProbes32Values))
	for i, value := range tier1ArithmeticProbes32Values {
		probeWords32[i] = bid128BidCodecValue{lo: uint64(value)}
	}
	probeWords64 := make([]bid128BidCodecValue, len(tier1ArithmeticProbes64Values))
	for i, value := range tier1ArithmeticProbes64Values {
		probeWords64[i] = bid128BidCodecValue{lo: value}
	}
	streamHashes := map[string]string{
		"@@TIER1_PAIR_STREAM_HASH32@@":        fmt.Sprint(tier1ArithmeticPairStreamHashForGeneration(32, boundaryWords32, probeWords32)),
		"@@TIER1_PAIR_STREAM_HASH64@@":        fmt.Sprint(tier1ArithmeticPairStreamHashForGeneration(64, boundaryWords64, probeWords64)),
		"@@TIER1_PAIR_STREAM_HASH128@@":       fmt.Sprint(tier1ArithmeticPairStreamHashForGeneration(128, boundary128, tier1ArithmeticProbes128Values)),
		"@@TIER1_FMA_TRIPLE_STREAM_HASH32@@":  fmt.Sprint(tier1ArithmeticTripleStreamHashForGeneration(32, boundaryWords32, probeWords32)),
		"@@TIER1_FMA_TRIPLE_STREAM_HASH64@@":  fmt.Sprint(tier1ArithmeticTripleStreamHashForGeneration(64, boundaryWords64, probeWords64)),
		"@@TIER1_FMA_TRIPLE_STREAM_HASH128@@": fmt.Sprint(tier1ArithmeticTripleStreamHashForGeneration(128, boundary128, tier1ArithmeticProbes128Values)),
		"@@TIER1_RANDOM_STREAM_HASH32@@":      fmt.Sprint(tier1ArithmeticRandomStreamHashForGeneration(32, uint64(1)<<20)),
		"@@TIER1_RANDOM_STREAM_HASH64@@":      fmt.Sprint(tier1ArithmeticRandomStreamHashForGeneration(64, uint64(1)<<20)),
		"@@TIER1_RANDOM_STREAM_HASH128@@":     fmt.Sprint(tier1ArithmeticRandomStreamHashForGeneration(128, uint64(1)<<19)),
	}
	streamHashReplacements := make([]string, 0, len(streamHashes)*2)
	for token, value := range streamHashes {
		streamHashReplacements = append(streamHashReplacements, token, value)
	}
	// Routing sentinels: deterministic known-answer rows selected and
	// self-asserted (S1/S2/S3) by tier1_sentinel_codegen.go; a selection
	// failure aborts the whole generation run with no partial output.
	sentinelRows, err := GenerateTier1ArithmeticSentinelRows()
	if err != nil {
		return nil, err
	}
	replacer := strings.NewReplacer(
		"@@TIER1_ARITH_SENTINEL_COUNT@@", fmt.Sprint(len(sentinelRows)),
		"@@TIER1_ARITH_SENTINEL_ROWS@@", tier1SentinelGoRowLiterals(sentinelRows),
		"@@TIER1_BOUNDARY32_COUNT@@", fmt.Sprint(len(boundary32)),
		"@@TIER1_BOUNDARY64_COUNT@@", fmt.Sprint(len(boundary64)),
		"@@TIER1_BOUNDARY128_COUNT@@", fmt.Sprint(len(boundary128)),
		"@@TIER1_BOUNDARY32_VALUES@@", tier1ArithmeticUint32Literals(boundary32),
		"@@TIER1_BOUNDARY64_VALUES@@", tier1ArithmeticUint64Literals(boundary64),
		"@@TIER1_BOUNDARY128_VALUES@@", tier1ArithmeticUint128Literals(boundary128),
		"@@TIER1_SEMANTIC_ROUNDED32_COUNT@@", fmt.Sprint(len(semantic.rounded32)),
		"@@TIER1_SEMANTIC_ROUNDED64_COUNT@@", fmt.Sprint(len(semantic.rounded64)),
		"@@TIER1_SEMANTIC_ROUNDED128_COUNT@@", fmt.Sprint(len(semantic.rounded128)),
		"@@TIER1_SEMANTIC_SCALE32_COUNT@@", fmt.Sprint(len(semantic.scale32)),
		"@@TIER1_SEMANTIC_SCALE64_COUNT@@", fmt.Sprint(len(semantic.scale64)),
		"@@TIER1_SEMANTIC_SCALE128_COUNT@@", fmt.Sprint(len(semantic.scale128)),
		"@@TIER1_SEMANTIC_REMAINDER32_COUNT@@", fmt.Sprint(len(semantic.remainder32)),
		"@@TIER1_SEMANTIC_REMAINDER64_COUNT@@", fmt.Sprint(len(semantic.remainder64)),
		"@@TIER1_SEMANTIC_REMAINDER128_COUNT@@", fmt.Sprint(len(semantic.remainder128)),
		"@@TIER1_SEMANTIC_FMA32_COUNT@@", fmt.Sprint(len(semantic.fma32)),
		"@@TIER1_SEMANTIC_FMA64_COUNT@@", fmt.Sprint(len(semantic.fma64)),
		"@@TIER1_SEMANTIC_FMA128_COUNT@@", fmt.Sprint(len(semantic.fma128)),
		"@@TIER1_SEMANTIC_SQRT32_COUNT@@", fmt.Sprint(len(semantic.sqrt32)),
		"@@TIER1_SEMANTIC_SQRT64_COUNT@@", fmt.Sprint(len(semantic.sqrt64)),
		"@@TIER1_SEMANTIC_SQRT128_COUNT@@", fmt.Sprint(len(semantic.sqrt128)),
		"@@TIER1_SEMANTIC_FMA32_VALUES@@", tier1ArithmeticTriple32Literals(semantic.fma32),
		"@@TIER1_SEMANTIC_FMA64_VALUES@@", tier1ArithmeticTriple64Literals(semantic.fma64),
		"@@TIER1_SEMANTIC_FMA128_VALUES@@", tier1ArithmeticTriple128Literals(semantic.fma128),
		"@@TIER1_SEMANTIC_SQRT32_VALUES@@", tier1ArithmeticUint32Literals(semantic.sqrt32),
		"@@TIER1_SEMANTIC_SQRT64_VALUES@@", tier1ArithmeticUint64Literals(semantic.sqrt64),
		"@@TIER1_SEMANTIC_SQRT128_VALUES@@", tier1ArithmeticUint128Literals(semantic.sqrt128),
		"@@TIER1_SEMANTIC_ROUNDED32_VALUES@@", tier1ArithmeticRounded32Literals(semantic.rounded32),
		"@@TIER1_SEMANTIC_ROUNDED64_VALUES@@", tier1ArithmeticRounded64Literals(semantic.rounded64),
		"@@TIER1_SEMANTIC_ROUNDED128_VALUES@@", tier1ArithmeticRounded128Literals(semantic.rounded128),
		"@@TIER1_SEMANTIC_SCALE32_VALUES@@", tier1ArithmeticScale32Literals(semantic.scale32),
		"@@TIER1_SEMANTIC_SCALE64_VALUES@@", tier1ArithmeticScale64Literals(semantic.scale64),
		"@@TIER1_SEMANTIC_SCALE128_VALUES@@", tier1ArithmeticScale128Literals(semantic.scale128),
		"@@TIER1_SEMANTIC_REMAINDER32_VALUES@@", tier1ArithmeticPair32Literals(semantic.remainder32),
		"@@TIER1_SEMANTIC_REMAINDER64_VALUES@@", tier1ArithmeticPair64Literals(semantic.remainder64),
		"@@TIER1_SEMANTIC_REMAINDER128_VALUES@@", tier1ArithmeticPair128Literals(semantic.remainder128),
		"@@TIER1_SCALE_FINITE_TRANSITION_LIMIT32@@", fmt.Sprint(tier1ArithmeticScaleFiniteTransitionLimit32),
		"@@TIER1_SCALE_FINITE_TRANSITION_LIMIT64@@", fmt.Sprint(tier1ArithmeticScaleFiniteTransitionLimit64),
		"@@TIER1_SCALE_FINITE_TRANSITION_LIMIT128@@", fmt.Sprint(tier1ArithmeticScaleFiniteTransitionLimit128),
		"@@TIER1_SCALE_RANDOM_STRATA@@", fmt.Sprint(tier1ArithmeticScaleRandomStrata),
		"@@TIER1_SCALE_MODE_CROSS@@", fmt.Sprint(tier1ArithmeticScaleModeCross),
		"@@TIER1_SCALE_MODE_CROSS_GROUPS32@@", fmt.Sprint(tier1ArithmeticScaleModeCrossGroups(tier1ArithmeticScaleFiniteTransitionLimit32)),
		"@@TIER1_SCALE_MODE_CROSS_GROUPS64@@", fmt.Sprint(tier1ArithmeticScaleModeCrossGroups(tier1ArithmeticScaleFiniteTransitionLimit64)),
		"@@TIER1_SCALE_MODE_CROSS_GROUPS128@@", fmt.Sprint(tier1ArithmeticScaleModeCrossGroups(tier1ArithmeticScaleFiniteTransitionLimit128)),
		"@@TIER1_SCALE_TUPLE_HASH32@@", fmt.Sprint(tier1ArithmeticScaleTupleHashForGeneration(32, 0xdec7543253414c45, 1, tier1ArithmeticScaleFiniteTransitionLimit32, uint64(1)<<20)),
		"@@TIER1_SCALE_TUPLE_HASH64@@", fmt.Sprint(tier1ArithmeticScaleTupleHashForGeneration(64, 0xdec7546453414c45, 1, tier1ArithmeticScaleFiniteTransitionLimit64, uint64(1)<<20)),
		"@@TIER1_SCALE_TUPLE_HASH128@@", fmt.Sprint(tier1ArithmeticScaleTupleHashForGeneration(128, 0xdec7541253414c45, 2, tier1ArithmeticScaleFiniteTransitionLimit128, uint64(1)<<19)),
		"@@TIER1_RANDOM_SAMPLE0@@", fmt.Sprintf("0x%016x", tier1ArithmeticRandomWordForGeneration(0xdec7543200000000, 0, 0)),
		"@@TIER1_RANDOM_SAMPLE1@@", fmt.Sprintf("0x%016x", tier1ArithmeticRandomWordForGeneration(0xdec7546400000004, (uint64(1)<<20)-1, 1)),
		"@@TIER1_RANDOM_SAMPLE2@@", fmt.Sprintf("0x%016x", tier1ArithmeticRandomWordForGeneration(0xdec7541253414c45, (uint64(1)<<19)-1, 2)),
	)
	goStreamReplacer := strings.NewReplacer(append(append([]string{}, streamHashReplacements...),
		"@@TIER1_PROBES32_VALUES@@", tier1ArithmeticUint32Literals(tier1ArithmeticProbes32Values),
		"@@TIER1_PROBES64_VALUES@@", tier1ArithmeticUint64Literals(tier1ArithmeticProbes64Values),
		"@@TIER1_PROBES128_VALUES@@", tier1ArithmeticUint128Literals(tier1ArithmeticProbes128Values),
	)...)
	source := []byte(genmarker.Line("testgen") + "\n" + goStreamReplacer.Replace(replacer.Replace(string(template))))
	goOutputs, err := formatGeneratedGoOutputs(map[string][]byte{
		tier1ArithmeticLongGeneratedPath: source,
	})
	if err != nil {
		return nil, err
	}
	rustReplacer := strings.NewReplacer(
		"@@TIER1_ARITH_SENTINEL_COUNT@@", fmt.Sprint(len(sentinelRows)),
		"@@TIER1_ARITH_SENTINEL_ROWS@@", tier1SentinelRustRowLiterals(sentinelRows),
		"@@TIER1_BOUNDARY32_COUNT@@", fmt.Sprint(len(boundary32)),
		"@@TIER1_BOUNDARY64_COUNT@@", fmt.Sprint(len(boundary64)),
		"@@TIER1_BOUNDARY128_COUNT@@", fmt.Sprint(len(boundary128)),
		"@@TIER1_STRUCTURED32_COUNT@@", fmt.Sprint(counts.structured32),
		"@@TIER1_STRUCTURED64_COUNT@@", fmt.Sprint(counts.structured64),
		"@@TIER1_STRUCTURED128_COUNT@@", fmt.Sprint(counts.structured128),
		"@@TIER1_RANDOM32_COUNT@@", fmt.Sprint(counts.random32),
		"@@TIER1_RANDOM64_COUNT@@", fmt.Sprint(counts.random64),
		"@@TIER1_RANDOM128_COUNT@@", fmt.Sprint(counts.random128),
		"@@TIER1_RANDOM_CASES_PER_OP32@@", fmt.Sprint(uint64(1)<<20),
		"@@TIER1_RANDOM_CASES_PER_OP64@@", fmt.Sprint(uint64(1)<<20),
		"@@TIER1_RANDOM_CASES_PER_OP128@@", fmt.Sprint(uint64(1)<<19),
		"@@TIER1_TOTAL32_COUNT@@", fmt.Sprint(counts.total32),
		"@@TIER1_TOTAL64_COUNT@@", fmt.Sprint(counts.total64),
		"@@TIER1_TOTAL128_COUNT@@", fmt.Sprint(counts.total128),
		"@@TIER1_SEMANTIC_ROUNDED32_COUNT@@", fmt.Sprint(len(semantic.rounded32)),
		"@@TIER1_SEMANTIC_ROUNDED64_COUNT@@", fmt.Sprint(len(semantic.rounded64)),
		"@@TIER1_SEMANTIC_ROUNDED128_COUNT@@", fmt.Sprint(len(semantic.rounded128)),
		"@@TIER1_SEMANTIC_SCALE32_COUNT@@", fmt.Sprint(len(semantic.scale32)),
		"@@TIER1_SEMANTIC_SCALE64_COUNT@@", fmt.Sprint(len(semantic.scale64)),
		"@@TIER1_SEMANTIC_SCALE128_COUNT@@", fmt.Sprint(len(semantic.scale128)),
		"@@TIER1_SEMANTIC_REMAINDER32_COUNT@@", fmt.Sprint(len(semantic.remainder32)),
		"@@TIER1_SEMANTIC_REMAINDER64_COUNT@@", fmt.Sprint(len(semantic.remainder64)),
		"@@TIER1_SEMANTIC_REMAINDER128_COUNT@@", fmt.Sprint(len(semantic.remainder128)),
		"@@TIER1_SEMANTIC_FMA32_COUNT@@", fmt.Sprint(len(semantic.fma32)),
		"@@TIER1_SEMANTIC_FMA64_COUNT@@", fmt.Sprint(len(semantic.fma64)),
		"@@TIER1_SEMANTIC_FMA128_COUNT@@", fmt.Sprint(len(semantic.fma128)),
		"@@TIER1_SEMANTIC_SQRT32_COUNT@@", fmt.Sprint(len(semantic.sqrt32)),
		"@@TIER1_SEMANTIC_SQRT64_COUNT@@", fmt.Sprint(len(semantic.sqrt64)),
		"@@TIER1_SEMANTIC_SQRT128_COUNT@@", fmt.Sprint(len(semantic.sqrt128)),
		"@@TIER1_SEMANTIC_FMA32_VALUES@@", tier1RustTriple32Literals(semantic.fma32),
		"@@TIER1_SEMANTIC_FMA64_VALUES@@", tier1RustTriple64Literals(semantic.fma64),
		"@@TIER1_SEMANTIC_FMA128_VALUES@@", tier1RustTriple128Literals(semantic.fma128),
		"@@TIER1_SEMANTIC_SQRT32_VALUES@@", tier1RustUint32Literals(semantic.sqrt32),
		"@@TIER1_SEMANTIC_SQRT64_VALUES@@", tier1RustUint64Literals(semantic.sqrt64),
		"@@TIER1_SEMANTIC_SQRT128_VALUES@@", tier1RustUint128Literals(semantic.sqrt128),
		"@@TIER1_BOUNDARY32_VALUES@@", tier1RustUint32Literals(boundary32),
		"@@TIER1_BOUNDARY64_VALUES@@", tier1RustUint64Literals(boundary64),
		"@@TIER1_BOUNDARY128_VALUES@@", tier1RustUint128Literals(boundary128),
		"@@TIER1_SEMANTIC_ROUNDED32_VALUES@@", tier1RustRounded32Literals(semantic.rounded32),
		"@@TIER1_SEMANTIC_ROUNDED64_VALUES@@", tier1RustRounded64Literals(semantic.rounded64),
		"@@TIER1_SEMANTIC_ROUNDED128_VALUES@@", tier1RustRounded128Literals(semantic.rounded128),
		"@@TIER1_SEMANTIC_SCALE32_VALUES@@", tier1RustScale32Literals(semantic.scale32),
		"@@TIER1_SEMANTIC_SCALE64_VALUES@@", tier1RustScale64Literals(semantic.scale64),
		"@@TIER1_SEMANTIC_SCALE128_VALUES@@", tier1RustScale128Literals(semantic.scale128),
		"@@TIER1_SEMANTIC_REMAINDER32_VALUES@@", tier1RustPair32Literals(semantic.remainder32),
		"@@TIER1_SEMANTIC_REMAINDER64_VALUES@@", tier1RustPair64Literals(semantic.remainder64),
		"@@TIER1_SEMANTIC_REMAINDER128_VALUES@@", tier1RustPair128Literals(semantic.remainder128),
		"@@TIER1_SCALE_FINITE_TRANSITION_LIMIT32@@", fmt.Sprint(tier1ArithmeticScaleFiniteTransitionLimit32),
		"@@TIER1_SCALE_FINITE_TRANSITION_LIMIT64@@", fmt.Sprint(tier1ArithmeticScaleFiniteTransitionLimit64),
		"@@TIER1_SCALE_FINITE_TRANSITION_LIMIT128@@", fmt.Sprint(tier1ArithmeticScaleFiniteTransitionLimit128),
		"@@TIER1_SCALE_RANDOM_STRATA@@", fmt.Sprint(tier1ArithmeticScaleRandomStrata),
		"@@TIER1_SCALE_MODE_CROSS@@", fmt.Sprint(tier1ArithmeticScaleModeCross),
		"@@TIER1_SCALE_MODE_CROSS_GROUPS32@@", fmt.Sprint(tier1ArithmeticScaleModeCrossGroups(tier1ArithmeticScaleFiniteTransitionLimit32)),
		"@@TIER1_SCALE_MODE_CROSS_GROUPS64@@", fmt.Sprint(tier1ArithmeticScaleModeCrossGroups(tier1ArithmeticScaleFiniteTransitionLimit64)),
		"@@TIER1_SCALE_MODE_CROSS_GROUPS128@@", fmt.Sprint(tier1ArithmeticScaleModeCrossGroups(tier1ArithmeticScaleFiniteTransitionLimit128)),
		"@@TIER1_SCALE_TUPLE_HASH32@@", fmt.Sprint(tier1ArithmeticScaleTupleHashForGeneration(32, 0xdec7543253414c45, 1, tier1ArithmeticScaleFiniteTransitionLimit32, uint64(1)<<20)),
		"@@TIER1_SCALE_TUPLE_HASH64@@", fmt.Sprint(tier1ArithmeticScaleTupleHashForGeneration(64, 0xdec7546453414c45, 1, tier1ArithmeticScaleFiniteTransitionLimit64, uint64(1)<<20)),
		"@@TIER1_SCALE_TUPLE_HASH128@@", fmt.Sprint(tier1ArithmeticScaleTupleHashForGeneration(128, 0xdec7541253414c45, 2, tier1ArithmeticScaleFiniteTransitionLimit128, uint64(1)<<19)),
		"@@TIER1_RANDOM_SAMPLE0@@", fmt.Sprintf("0x%016x", tier1ArithmeticRandomWordForGeneration(0xdec7543200000000, 0, 0)),
		"@@TIER1_RANDOM_SAMPLE1@@", fmt.Sprintf("0x%016x", tier1ArithmeticRandomWordForGeneration(0xdec7546400000004, (uint64(1)<<20)-1, 1)),
		"@@TIER1_RANDOM_SAMPLE2@@", fmt.Sprintf("0x%016x", tier1ArithmeticRandomWordForGeneration(0xdec7541253414c45, (uint64(1)<<19)-1, 2)),
	)
	rustStreamReplacer := strings.NewReplacer(append(append([]string{}, streamHashReplacements...),
		"@@TIER1_PROBES32_VALUES@@", tier1RustUint32Literals(tier1ArithmeticProbes32Values),
		"@@TIER1_PROBES64_VALUES@@", tier1RustUint64Literals(tier1ArithmeticProbes64Values),
		"@@TIER1_PROBES128_VALUES@@", tier1RustUint128Literals(tier1ArithmeticProbes128Values),
	)...)
	rustSource, err := formatGeneratedRustOutput([]byte(genmarker.Line("testgen") + "\n" + rustStreamReplacer.Replace(rustReplacer.Replace(string(rustTemplate)))))
	if err != nil {
		return nil, err
	}
	goOutputs[tier1ArithmeticRustLongGeneratedPath] = rustSource
	return goOutputs, nil
}

func tier1ArithmeticRandomWordForGeneration(seed, caseIndex, lane uint64) uint64 {
	value := seed ^ caseIndex*0xd1342543de82ef95 ^ lane*0x9e3779b97f4a7c15
	value += 0x9e3779b97f4a7c15
	value = (value ^ (value >> 30)) * 0xbf58476d1ce4e5b9
	value = (value ^ (value >> 27)) * 0x94d049bb133111eb
	return value ^ (value >> 31)
}

// tier1ArithmeticRandomScaleExponentForGeneration mirrors the exponent stream
// emitted into the generated Go and Rust Tier 1 ScaleB runners.
func tier1ArithmeticScaleModeCrossGroups(transitionLimit int64) uint64 {
	return uint64(transitionLimit*2 + 1)
}

func tier1ArithmeticScaleModeCrossCase(caseIndex uint64, transitionLimit int64) bool {
	if caseIndex%tier1ArithmeticScaleRandomStrata != 1 {
		return false
	}
	slot := caseIndex / tier1ArithmeticScaleRandomStrata
	return slot < tier1ArithmeticScaleModeCrossGroups(transitionLimit)*tier1ArithmeticScaleModeCross
}

func tier1ArithmeticScaleModeCrossGroup(caseIndex uint64) uint64 {
	return (caseIndex / tier1ArithmeticScaleRandomStrata) / tier1ArithmeticScaleModeCross
}

func tier1ArithmeticRandomScaleExponentForGeneration(seed, caseIndex, lane uint64, transitionLimit int64) int64 {
	switch caseIndex % tier1ArithmeticScaleRandomStrata {
	case 0:
		return int64(tier1ArithmeticRandomWordForGeneration(seed, caseIndex, lane))
	case 1:
		if tier1ArithmeticScaleModeCrossCase(caseIndex, transitionLimit) {
			return int64(tier1ArithmeticScaleModeCrossGroup(caseIndex)) - transitionLimit
		}
		return int64(tier1ArithmeticRandomWordForGeneration(seed, caseIndex, lane)%uint64(transitionLimit*2+1)) - transitionLimit
	case 2:
		return int64(tier1ArithmeticRandomWordForGeneration(seed, caseIndex, lane)%uint64(transitionLimit*4+1)) - transitionLimit*2
	default:
		return int64(tier1ArithmeticRandomWordForGeneration(seed, caseIndex, lane)%uint64(transitionLimit*2+1)) - transitionLimit
	}
}

// tier1ArithmeticRandomScaleOperandIndexForGeneration mirrors the operand
// stream index emitted into the generated Go and Rust Tier 1 ScaleB runners.
func tier1ArithmeticRandomScaleOperandIndexForGeneration(caseIndex uint64, transitionLimit int64) uint64 {
	if tier1ArithmeticScaleModeCrossCase(caseIndex, transitionLimit) {
		return tier1ArithmeticScaleModeCrossGroup(caseIndex)
	}
	return caseIndex
}

type tier1ArithmeticScaleWords128 struct {
	lo uint64
	hi uint64
}

func tier1ArithmeticScaleModeCrossSign(group uint64) uint64 {
	return group & 1
}

func tier1ArithmeticRandomScaleOperand32ForGeneration(seed, caseIndex uint64, transitionLimit int64) uint32 {
	if tier1ArithmeticScaleModeCrossCase(caseIndex, transitionLimit) {
		group := tier1ArithmeticScaleModeCrossGroup(caseIndex)
		exponent := int64(group) - transitionLimit
		sign := uint32(tier1ArithmeticScaleModeCrossSign(group) << 31)
		switch {
		case exponent > 0:
			return sign | 0x00000001
		case exponent < 0:
			return sign | 0x77f8967f
		default:
			return sign | 0x32800001
		}
	}
	raw := uint32(tier1ArithmeticRandomWordForGeneration(seed, tier1ArithmeticRandomScaleOperandIndexForGeneration(caseIndex, transitionLimit), 0))
	if caseIndex%tier1ArithmeticScaleRandomStrata != 0 {
		raw = raw&^0x20000000 | 1
	}
	return raw
}

func tier1ArithmeticRandomScaleOperand64ForGeneration(seed, caseIndex uint64, transitionLimit int64) uint64 {
	if tier1ArithmeticScaleModeCrossCase(caseIndex, transitionLimit) {
		group := tier1ArithmeticScaleModeCrossGroup(caseIndex)
		exponent := int64(group) - transitionLimit
		sign := tier1ArithmeticScaleModeCrossSign(group) << 63
		switch {
		case exponent > 0:
			return sign | 0x0000000000000001
		case exponent < 0:
			return sign | 0x77fb86f26fc0ffff
		default:
			return sign | 0x31c0000000000001
		}
	}
	raw := tier1ArithmeticRandomWordForGeneration(seed, tier1ArithmeticRandomScaleOperandIndexForGeneration(caseIndex, transitionLimit), 0)
	if caseIndex%tier1ArithmeticScaleRandomStrata != 0 {
		raw = raw&^0x2000000000000000 | 1
	}
	return raw
}

func tier1ArithmeticRandomScaleOperand128ForGeneration(seed, caseIndex uint64, transitionLimit int64) tier1ArithmeticScaleWords128 {
	if tier1ArithmeticScaleModeCrossCase(caseIndex, transitionLimit) {
		group := tier1ArithmeticScaleModeCrossGroup(caseIndex)
		exponent := int64(group) - transitionLimit
		sign := tier1ArithmeticScaleModeCrossSign(group) << 63
		switch {
		case exponent > 0:
			return tier1ArithmeticScaleWords128{lo: 0x0000000000000001, hi: sign}
		case exponent < 0:
			return tier1ArithmeticScaleWords128{lo: 0x378d8e63ffffffff, hi: sign | 0x5fffed09bead87c0}
		default:
			return tier1ArithmeticScaleWords128{lo: 0x0000000000000001, hi: sign | 0x3040000000000000}
		}
	}
	operandIndex := tier1ArithmeticRandomScaleOperandIndexForGeneration(caseIndex, transitionLimit)
	raw := tier1ArithmeticScaleWords128{
		lo: tier1ArithmeticRandomWordForGeneration(seed, operandIndex, 0),
		hi: tier1ArithmeticRandomWordForGeneration(seed, operandIndex, 1),
	}
	if caseIndex%tier1ArithmeticScaleRandomStrata != 0 {
		raw.lo |= 1
		raw.hi &^= 0x2001000000000000
	}
	return raw
}

func tier1ArithmeticScaleTupleHashMix(digest, word uint64) uint64 {
	return (digest ^ word) * tier1ArithmeticScaleTupleHashPrime
}

func tier1ArithmeticScaleTupleHashForGeneration(bits int, seed, exponentLane uint64, transitionLimit int64, cases uint64) uint64 {
	digest := tier1ArithmeticScaleTupleHashMix(tier1ArithmeticScaleTupleHashOffset, uint64(bits))
	nativeModes := [...]uint64{0, 4, 3, 2, 1}
	for caseIndex := uint64(0); caseIndex < cases; caseIndex++ {
		var lo, hi uint64
		switch bits {
		case 32:
			lo = uint64(tier1ArithmeticRandomScaleOperand32ForGeneration(seed, caseIndex, transitionLimit))
		case 64:
			lo = tier1ArithmeticRandomScaleOperand64ForGeneration(seed, caseIndex, transitionLimit)
		case 128:
			words := tier1ArithmeticRandomScaleOperand128ForGeneration(seed, caseIndex, transitionLimit)
			lo, hi = words.lo, words.hi
		default:
			panic(fmt.Sprintf("unsupported Tier 1 ScaleB tuple width %d", bits))
		}
		exponent := tier1ArithmeticRandomScaleExponentForGeneration(seed, caseIndex, exponentLane, transitionLimit)
		digest = tier1ArithmeticScaleTupleHashMix(digest, caseIndex)
		digest = tier1ArithmeticScaleTupleHashMix(digest, lo)
		digest = tier1ArithmeticScaleTupleHashMix(digest, hi)
		digest = tier1ArithmeticScaleTupleHashMix(digest, uint64(exponent))
		digest = tier1ArithmeticScaleTupleHashMix(digest, nativeModes[caseIndex%uint64(len(nativeModes))])
	}
	return tier1ArithmeticScaleTupleHashMix(digest, cases)
}

// tier1ArithmeticPairStreamHashForGeneration replicates both runners'
// visit-pairs iteration (boundary×probes in both operand orders, then the
// full probes×probes cross) and folds every operand word plus the final visit
// count into the ScaleB FNV mix. The generated Go and Rust corpus-contract
// tests recompute this digest over their own visit functions, so an
// iteration-order or probe drift in either template fails against the anchor.
func tier1ArithmeticPairStreamHashForGeneration(bits uint64, boundary, probes []bid128BidCodecValue) uint64 {
	digest := tier1ArithmeticScaleTupleHashMix(tier1ArithmeticScaleTupleHashOffset, bits)
	var visits uint64
	visit := func(x, y bid128BidCodecValue) {
		digest = tier1ArithmeticScaleTupleHashMix(digest, x.lo)
		digest = tier1ArithmeticScaleTupleHashMix(digest, x.hi)
		digest = tier1ArithmeticScaleTupleHashMix(digest, y.lo)
		digest = tier1ArithmeticScaleTupleHashMix(digest, y.hi)
		visits++
	}
	for _, x := range boundary {
		for _, y := range probes {
			visit(x, y)
			visit(y, x)
		}
	}
	for _, x := range probes {
		for _, y := range probes {
			visit(x, y)
		}
	}
	return tier1ArithmeticScaleTupleHashMix(digest, visits)
}

// tier1ArithmeticTripleStreamHashForGeneration replicates the fma
// visit-triples iteration (boundary value rotated through all three operand
// slots against (i+j)%probes companion pairs, then the full probes^3 cross).
func tier1ArithmeticTripleStreamHashForGeneration(bits uint64, boundary, probes []bid128BidCodecValue) uint64 {
	digest := tier1ArithmeticScaleTupleHashMix(tier1ArithmeticScaleTupleHashOffset, bits)
	var visits uint64
	visit := func(x, y, z bid128BidCodecValue) {
		digest = tier1ArithmeticScaleTupleHashMix(digest, x.lo)
		digest = tier1ArithmeticScaleTupleHashMix(digest, x.hi)
		digest = tier1ArithmeticScaleTupleHashMix(digest, y.lo)
		digest = tier1ArithmeticScaleTupleHashMix(digest, y.hi)
		digest = tier1ArithmeticScaleTupleHashMix(digest, z.lo)
		digest = tier1ArithmeticScaleTupleHashMix(digest, z.hi)
		visits++
	}
	for i, x := range boundary {
		for j, y := range probes {
			z := probes[(i+j)%len(probes)]
			visit(x, y, z)
			visit(y, z, x)
			visit(z, x, y)
		}
	}
	for _, x := range probes {
		for _, y := range probes {
			for _, z := range probes {
				visit(x, y, z)
			}
		}
	}
	return tier1ArithmeticScaleTupleHashMix(digest, visits)
}

// tier1ArithmeticRandomStreamHashForGeneration replicates the deterministic
// random blocks of both runners for every non-ScaleB operation (rounded add/
// sub/mul/div/quantize, unrounded remainder/fmod, fma, sqrt — ScaleB keeps
// its own dedicated tuple hash), folding the exact operand words the runners
// consume (including the 32-bit truncation) and the native rounding mode of
// each moded case. A seed, lane, operand-width, or mode-rotation drift in
// either template fails against the anchor.
func tier1ArithmeticRandomStreamHashForGeneration(bits int, cases uint64) uint64 {
	var roundedBase, unroundedBase, fmaSeed, sqrtSeed uint64
	switch bits {
	case 32:
		roundedBase, unroundedBase = 0xdec7543200000000, 0xdec7543210000000
		fmaSeed, sqrtSeed = 0xdec7543220000000, 0xdec7543230000000
	case 64:
		roundedBase, unroundedBase = 0xdec7546400000000, 0xdec7546410000000
		fmaSeed, sqrtSeed = 0xdec7546420000000, 0xdec7546430000000
	case 128:
		roundedBase, unroundedBase = 0xdec7541200000000, 0xdec7541210000000
		fmaSeed, sqrtSeed = 0xdec7541220000000, 0xdec7541230000000
	default:
		panic(fmt.Sprintf("unsupported Tier 1 random stream width %d", bits))
	}
	nativeModes := [...]uint64{0, 4, 3, 2, 1}
	digest := tier1ArithmeticScaleTupleHashMix(tier1ArithmeticScaleTupleHashOffset, uint64(bits))
	operand := func(seed, caseIndex, lane uint64) (lo, hi uint64) {
		switch bits {
		case 32:
			return uint64(uint32(tier1ArithmeticRandomWordForGeneration(seed, caseIndex, lane))), 0
		case 64:
			return tier1ArithmeticRandomWordForGeneration(seed, caseIndex, lane), 0
		default:
			return tier1ArithmeticRandomWordForGeneration(seed, caseIndex, lane*2),
				tier1ArithmeticRandomWordForGeneration(seed, caseIndex, lane*2+1)
		}
	}
	mixOperand := func(seed, caseIndex, lane uint64) {
		lo, hi := operand(seed, caseIndex, lane)
		digest = tier1ArithmeticScaleTupleHashMix(digest, lo)
		digest = tier1ArithmeticScaleTupleHashMix(digest, hi)
	}
	var total uint64
	for opIndex := uint64(0); opIndex < 5; opIndex++ {
		seed := roundedBase ^ opIndex
		for i := uint64(0); i < cases; i++ {
			mixOperand(seed, i, 0)
			mixOperand(seed, i, 1)
			digest = tier1ArithmeticScaleTupleHashMix(digest, nativeModes[i%uint64(len(nativeModes))])
			total++
		}
	}
	for opIndex := uint64(0); opIndex < 2; opIndex++ {
		seed := unroundedBase ^ opIndex
		for i := uint64(0); i < cases; i++ {
			mixOperand(seed, i, 0)
			mixOperand(seed, i, 1)
			total++
		}
	}
	for i := uint64(0); i < cases; i++ {
		mixOperand(fmaSeed, i, 0)
		mixOperand(fmaSeed, i, 1)
		mixOperand(fmaSeed, i, 2)
		digest = tier1ArithmeticScaleTupleHashMix(digest, nativeModes[i%uint64(len(nativeModes))])
		total++
	}
	for i := uint64(0); i < cases; i++ {
		mixOperand(sqrtSeed, i, 0)
		digest = tier1ArithmeticScaleTupleHashMix(digest, nativeModes[i%uint64(len(nativeModes))])
		total++
	}
	return tier1ArithmeticScaleTupleHashMix(digest, total)
}

type tier1ArithmeticCounts struct {
	structured32, structured64, structured128 uint64
	random32, random64, random128             uint64
	total32, total64, total128                uint64
}

func tier1ArithmeticCountsFor(
	boundary32, boundary64, boundary128 uint64,
	rounded32, rounded64, rounded128 uint64,
	scale32, scale64, scale128 uint64,
	remainder32, remainder64, remainder128 uint64,
	fma32, fma64, fma128 uint64,
	sqrt32, sqrt64, sqrt128 uint64,
) tier1ArithmeticCounts {
	const (
		probes         = uint64(12)
		roundingModes  = uint64(5)
		roundedOps     = uint64(5)
		unroundedOps   = uint64(2)
		fmaSlots       = uint64(3)
		scaleExponents = uint64(25)
		randomOps      = uint64(10)
		randomCases32  = uint64(1) << 20
		randomCases64  = uint64(1) << 20
		randomCases128 = uint64(1) << 19
	)
	structured := func(boundary, rounded, scale, remainder, fma, sqrt uint64) uint64 {
		pairs := boundary*probes*2 + probes*probes
		fmaTriples := boundary*probes*fmaSlots + probes*probes*probes
		return pairs*(roundedOps*roundingModes+unroundedOps) +
			(fmaTriples+fma)*roundingModes +
			(boundary+sqrt)*roundingModes +
			boundary*scaleExponents*roundingModes +
			(rounded+scale)*roundingModes + remainder*unroundedOps
	}
	result := tier1ArithmeticCounts{
		structured32:  structured(boundary32, rounded32, scale32, remainder32, fma32, sqrt32),
		structured64:  structured(boundary64, rounded64, scale64, remainder64, fma64, sqrt64),
		structured128: structured(boundary128, rounded128, scale128, remainder128, fma128, sqrt128),
		random32:      randomCases32 * randomOps,
		random64:      randomCases64 * randomOps,
		random128:     randomCases128 * randomOps,
	}
	result.total32 = result.structured32 + result.random32
	result.total64 = result.structured64 + result.random64
	result.total128 = result.structured128 + result.random128
	return result
}

type tier1ArithmeticRounded32Spec struct {
	operation string
	x         uint32
	y         uint32
}

type tier1ArithmeticRounded64Spec struct {
	operation string
	x         uint64
	y         uint64
}

type tier1ArithmeticRounded128Spec struct {
	operation string
	x         bid128BidCodecValue
	y         bid128BidCodecValue
}

type tier1ArithmeticScale32Spec struct {
	x        uint32
	exponent int
}

type tier1ArithmeticScale64Spec struct {
	x        uint64
	exponent int
}

type tier1ArithmeticScale128Spec struct {
	x        bid128BidCodecValue
	exponent int
}

type tier1ArithmeticPair32Spec struct {
	x uint32
	y uint32
}

type tier1ArithmeticPair64Spec struct {
	x uint64
	y uint64
}

type tier1ArithmeticPair128Spec struct {
	x bid128BidCodecValue
	y bid128BidCodecValue
}

type tier1ArithmeticTriple32Spec struct {
	x uint32
	y uint32
	z uint32
}

type tier1ArithmeticTriple64Spec struct {
	x uint64
	y uint64
	z uint64
}

type tier1ArithmeticTriple128Spec struct {
	x bid128BidCodecValue
	y bid128BidCodecValue
	z bid128BidCodecValue
}

type tier1ArithmeticSemanticSpec struct {
	rounded32    []tier1ArithmeticRounded32Spec
	rounded64    []tier1ArithmeticRounded64Spec
	rounded128   []tier1ArithmeticRounded128Spec
	scale32      []tier1ArithmeticScale32Spec
	scale64      []tier1ArithmeticScale64Spec
	scale128     []tier1ArithmeticScale128Spec
	remainder32  []tier1ArithmeticPair32Spec
	remainder64  []tier1ArithmeticPair64Spec
	remainder128 []tier1ArithmeticPair128Spec
	fma32        []tier1ArithmeticTriple32Spec
	fma64        []tier1ArithmeticTriple64Spec
	fma128       []tier1ArithmeticTriple128Spec
	sqrt32       []uint32
	sqrt64       []uint64
	sqrt128      []bid128BidCodecValue
}

func tier1ArithmeticSemanticCorpus() (tier1ArithmeticSemanticSpec, error) {
	var result tier1ArithmeticSemanticSpec
	for _, operation := range []string{"Add", "Sub", "Mul", "Div", "Quantize"} {
		publicName := strings.ToLower(operation)
		for _, width := range []int{32, 64, 128} {
			pairs, err := modeBinaryDiscriminantOperands(operation, width)
			if err != nil {
				return result, fmt.Errorf("Tier 1 arithmetic long semantic corpus: %w", err)
			}
			for _, pair := range pairs {
				switch width {
				case 32:
					x, err := encodeModeDiscOperand32(pair[0])
					if err != nil {
						return result, err
					}
					y, err := encodeModeDiscOperand32(pair[1])
					if err != nil {
						return result, err
					}
					result.rounded32 = append(result.rounded32, tier1ArithmeticRounded32Spec{operation: publicName, x: x, y: y})
				case 64:
					x, err := encodeModeDiscOperand64(pair[0])
					if err != nil {
						return result, err
					}
					y, err := encodeModeDiscOperand64(pair[1])
					if err != nil {
						return result, err
					}
					result.rounded64 = append(result.rounded64, tier1ArithmeticRounded64Spec{operation: publicName, x: x, y: y})
				case 128:
					x, err := encodeModeDiscOperand128(pair[0])
					if err != nil {
						return result, err
					}
					y, err := encodeModeDiscOperand128(pair[1])
					if err != nil {
						return result, err
					}
					result.rounded128 = append(result.rounded128, tier1ArithmeticRounded128Spec{
						operation: publicName,
						x:         bid128BidCodecValue{lo: binary.LittleEndian.Uint64(x[0:8]), hi: binary.LittleEndian.Uint64(x[8:16])},
						y:         bid128BidCodecValue{lo: binary.LittleEndian.Uint64(y[0:8]), hi: binary.LittleEndian.Uint64(y[8:16])},
					})
				}
			}
		}
	}
	// Pinned Intel readtest lines 126095-126104 exercise multiplication that is
	// tiny before rounding but rounds to the minimum normal BID32 value. Cross
	// both signs with every rounding mode so Tier 1 pins the configured
	// DECIMAL_TINY_DETECTION_AFTER_ROUNDING=0 Underflow|Inexact status.
	for _, pair := range [][2]modeDiscOperand{
		{mdo(1010101, -95), mdo(99, -8)},
		{mdo(1010101, -95), mdoNeg(99, -8)},
	} {
		x, err := encodeModeDiscOperand32(pair[0])
		if err != nil {
			return result, err
		}
		y, err := encodeModeDiscOperand32(pair[1])
		if err != nil {
			return result, err
		}
		result.rounded32 = append(result.rounded32, tier1ArithmeticRounded32Spec{operation: "mul", x: x, y: y})
	}
	for _, width := range []int{32, 64, 128} {
		cases, err := modeScaleBDiscriminantCases("ScaleB", width)
		if err != nil {
			return result, fmt.Errorf("Tier 1 arithmetic long semantic scaleB corpus: %w", err)
		}
		for _, tc := range cases {
			switch width {
			case 32:
				x, err := encodeModeDiscOperand32(tc.Val)
				if err != nil {
					return result, err
				}
				result.scale32 = append(result.scale32, tier1ArithmeticScale32Spec{x: x, exponent: tc.Exp})
			case 64:
				x, err := encodeModeDiscOperand64(tc.Val)
				if err != nil {
					return result, err
				}
				result.scale64 = append(result.scale64, tier1ArithmeticScale64Spec{x: x, exponent: tc.Exp})
			case 128:
				x, err := encodeModeDiscOperand128(tc.Val)
				if err != nil {
					return result, err
				}
				result.scale128 = append(result.scale128, tier1ArithmeticScale128Spec{
					x:        bid128BidCodecValue{lo: binary.LittleEndian.Uint64(x[0:8]), hi: binary.LittleEndian.Uint64(x[8:16])},
					exponent: tc.Exp,
				})
			}
		}
	}
	for _, width := range []int{32, 64, 128} {
		triples, err := modeTernaryDiscriminantOperands("FMA", width)
		if err != nil {
			return result, fmt.Errorf("Tier 1 arithmetic long semantic FMA corpus: %w", err)
		}
		for _, triple := range triples {
			switch width {
			case 32:
				x, err := encodeModeDiscOperand32(triple[0])
				if err != nil {
					return result, err
				}
				y, err := encodeModeDiscOperand32(triple[1])
				if err != nil {
					return result, err
				}
				z, err := encodeModeDiscOperand32(triple[2])
				if err != nil {
					return result, err
				}
				result.fma32 = append(result.fma32, tier1ArithmeticTriple32Spec{x: x, y: y, z: z})
			case 64:
				x, err := encodeModeDiscOperand64(triple[0])
				if err != nil {
					return result, err
				}
				y, err := encodeModeDiscOperand64(triple[1])
				if err != nil {
					return result, err
				}
				z, err := encodeModeDiscOperand64(triple[2])
				if err != nil {
					return result, err
				}
				result.fma64 = append(result.fma64, tier1ArithmeticTriple64Spec{x: x, y: y, z: z})
			case 128:
				x, err := encodeModeDiscOperand128(triple[0])
				if err != nil {
					return result, err
				}
				y, err := encodeModeDiscOperand128(triple[1])
				if err != nil {
					return result, err
				}
				z, err := encodeModeDiscOperand128(triple[2])
				if err != nil {
					return result, err
				}
				result.fma128 = append(result.fma128, tier1ArithmeticTriple128Spec{
					x: bid128BidCodecValue{lo: binary.LittleEndian.Uint64(x[0:8]), hi: binary.LittleEndian.Uint64(x[8:16])},
					y: bid128BidCodecValue{lo: binary.LittleEndian.Uint64(y[0:8]), hi: binary.LittleEndian.Uint64(y[8:16])},
					z: bid128BidCodecValue{lo: binary.LittleEndian.Uint64(z[0:8]), hi: binary.LittleEndian.Uint64(z[8:16])},
				})
			}
		}
	}
	for _, width := range []int{32, 64, 128} {
		operands, err := modeUnaryDiscriminantOperands("Sqrt", width)
		if err != nil {
			return result, fmt.Errorf("Tier 1 arithmetic long semantic sqrt corpus: %w", err)
		}
		for _, operand := range operands {
			switch width {
			case 32:
				x, err := encodeModeDiscOperand32(operand)
				if err != nil {
					return result, err
				}
				result.sqrt32 = append(result.sqrt32, x)
			case 64:
				x, err := encodeModeDiscOperand64(operand)
				if err != nil {
					return result, err
				}
				result.sqrt64 = append(result.sqrt64, x)
			case 128:
				x, err := encodeModeDiscOperand128(operand)
				if err != nil {
					return result, err
				}
				result.sqrt128 = append(result.sqrt128, bid128BidCodecValue{
					lo: binary.LittleEndian.Uint64(x[0:8]),
					hi: binary.LittleEndian.Uint64(x[8:16]),
				})
			}
		}
	}
	remainderPairs := [][2]modeDiscOperand{
		{mdo(5, 0), mdo(3, 0)},
		{mdo(7, 0), mdo(4, 0)},
		{mdoNeg(5, 0), mdo(3, 0)},
		{mdo(5, 0), mdoNeg(3, 0)},
	}
	for _, pair := range remainderPairs {
		x32, err := encodeModeDiscOperand32(pair[0])
		if err != nil {
			return result, err
		}
		y32, err := encodeModeDiscOperand32(pair[1])
		if err != nil {
			return result, err
		}
		result.remainder32 = append(result.remainder32, tier1ArithmeticPair32Spec{x: x32, y: y32})

		x64, err := encodeModeDiscOperand64(pair[0])
		if err != nil {
			return result, err
		}
		y64, err := encodeModeDiscOperand64(pair[1])
		if err != nil {
			return result, err
		}
		result.remainder64 = append(result.remainder64, tier1ArithmeticPair64Spec{x: x64, y: y64})

		x128, err := encodeModeDiscOperand128(pair[0])
		if err != nil {
			return result, err
		}
		y128, err := encodeModeDiscOperand128(pair[1])
		if err != nil {
			return result, err
		}
		result.remainder128 = append(result.remainder128, tier1ArithmeticPair128Spec{
			x: bid128BidCodecValue{lo: binary.LittleEndian.Uint64(x128[0:8]), hi: binary.LittleEndian.Uint64(x128[8:16])},
			y: bid128BidCodecValue{lo: binary.LittleEndian.Uint64(y128[0:8]), hi: binary.LittleEndian.Uint64(y128[8:16])},
		})
	}
	return result, nil
}

func tier1ArithmeticBoundary32Values() []uint32 {
	values := make(map[uint32]struct{})
	for _, value := range bid32BidCodecEdgeValues() {
		values[value] = struct{}{}
	}

	exponents := [...]uint32{0, 1, 2, 99, 100, 101, 102, 189, 190, 191}
	smallCoefficients := [...]uint32{
		0, 1, 2, 4, 5, 9, 10, 11, 49, 50, 51, 99, 100, 101,
		999998, 999999, 1000000, 1000001,
		0x003ffffe, 0x003fffff, 0x00400000, 0x00400001,
		0x007ffffe, 0x007fffff,
	}
	steeringContinuations := [...]uint32{
		0, 1, 2, 4, 5, 9, 10, 11,
		999998, 999999, 1000000, 1000001,
		0x0018967e, 0x0018967f, 0x00189680,
		0x001ffffe, 0x001fffff,
	}
	specialHeaders := [...]uint32{
		0x78000000, 0x79000000, 0x7a000000, 0x7b000000,
		0x7c000000, 0x7d000000, 0x7e000000, 0x7f000000,
	}
	reserved := [...]uint32{0, 0x00100000, 0x02000000, 0x03f00000}
	payloads := [...]uint32{
		0, 1, 2, 999998, 999999, 1000000, 1000001,
		0x000ffffe, 0x000fffff,
	}
	for _, sign := range [...]uint32{0, 0x80000000} {
		for _, exponent := range exponents {
			for _, coefficient := range smallCoefficients {
				values[sign|(exponent<<23)|coefficient] = struct{}{}
			}
			for _, continuation := range steeringContinuations {
				values[sign|0x60000000|(exponent<<21)|continuation] = struct{}{}
			}
		}
		for _, header := range specialHeaders {
			for _, reservedBits := range reserved {
				for _, payload := range payloads {
					values[sign|header|reservedBits|payload] = struct{}{}
				}
			}
		}
	}

	result := make([]uint32, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

func tier1ArithmeticUint32Literals(values []uint32) string {
	var out strings.Builder
	for _, value := range values {
		fmt.Fprintf(&out, "\t0x%08x,\n", value)
	}
	return out.String()
}

func tier1ArithmeticUint64Literals(values []uint64) string {
	var out strings.Builder
	for _, value := range values {
		fmt.Fprintf(&out, "\t0x%016x,\n", value)
	}
	return out.String()
}

func tier1ArithmeticUint128Literals(values []bid128BidCodecValue) string {
	var out strings.Builder
	for _, value := range values {
		fmt.Fprintf(&out, "\t{lo: 0x%016x, hi: 0x%016x},\n", value.lo, value.hi)
	}
	return out.String()
}

func tier1ArithmeticRounded32Literals(values []tier1ArithmeticRounded32Spec) string {
	var out strings.Builder
	for _, value := range values {
		fmt.Fprintf(&out, "\t{operation: %q, x: 0x%08x, y: 0x%08x},\n", value.operation, value.x, value.y)
	}
	return out.String()
}

func tier1ArithmeticRounded64Literals(values []tier1ArithmeticRounded64Spec) string {
	var out strings.Builder
	for _, value := range values {
		fmt.Fprintf(&out, "\t{operation: %q, x: 0x%016x, y: 0x%016x},\n", value.operation, value.x, value.y)
	}
	return out.String()
}

func tier1ArithmeticRounded128Literals(values []tier1ArithmeticRounded128Spec) string {
	var out strings.Builder
	for _, value := range values {
		fmt.Fprintf(&out, "\t{operation: %q, x: tier1Arithmetic128Words{lo: 0x%016x, hi: 0x%016x}, y: tier1Arithmetic128Words{lo: 0x%016x, hi: 0x%016x}},\n", value.operation, value.x.lo, value.x.hi, value.y.lo, value.y.hi)
	}
	return out.String()
}

func tier1ArithmeticScale32Literals(values []tier1ArithmeticScale32Spec) string {
	var out strings.Builder
	for _, value := range values {
		fmt.Fprintf(&out, "\t{x: 0x%08x, exponent: %d},\n", value.x, value.exponent)
	}
	return out.String()
}

func tier1ArithmeticScale64Literals(values []tier1ArithmeticScale64Spec) string {
	var out strings.Builder
	for _, value := range values {
		fmt.Fprintf(&out, "\t{x: 0x%016x, exponent: %d},\n", value.x, value.exponent)
	}
	return out.String()
}

func tier1ArithmeticScale128Literals(values []tier1ArithmeticScale128Spec) string {
	var out strings.Builder
	for _, value := range values {
		fmt.Fprintf(&out, "\t{x: tier1Arithmetic128Words{lo: 0x%016x, hi: 0x%016x}, exponent: %d},\n", value.x.lo, value.x.hi, value.exponent)
	}
	return out.String()
}

func tier1ArithmeticPair32Literals(values []tier1ArithmeticPair32Spec) string {
	var out strings.Builder
	for _, value := range values {
		fmt.Fprintf(&out, "\t{x: 0x%08x, y: 0x%08x},\n", value.x, value.y)
	}
	return out.String()
}

func tier1ArithmeticPair64Literals(values []tier1ArithmeticPair64Spec) string {
	var out strings.Builder
	for _, value := range values {
		fmt.Fprintf(&out, "\t{x: 0x%016x, y: 0x%016x},\n", value.x, value.y)
	}
	return out.String()
}

func tier1ArithmeticPair128Literals(values []tier1ArithmeticPair128Spec) string {
	var out strings.Builder
	for _, value := range values {
		fmt.Fprintf(&out, "\t{x: tier1Arithmetic128Words{lo: 0x%016x, hi: 0x%016x}, y: tier1Arithmetic128Words{lo: 0x%016x, hi: 0x%016x}},\n", value.x.lo, value.x.hi, value.y.lo, value.y.hi)
	}
	return out.String()
}

func tier1ArithmeticTriple32Literals(values []tier1ArithmeticTriple32Spec) string {
	var out strings.Builder
	for _, value := range values {
		fmt.Fprintf(&out, "\t{x: 0x%08x, y: 0x%08x, z: 0x%08x},\n", value.x, value.y, value.z)
	}
	return out.String()
}

func tier1ArithmeticTriple64Literals(values []tier1ArithmeticTriple64Spec) string {
	var out strings.Builder
	for _, value := range values {
		fmt.Fprintf(&out, "\t{x: 0x%016x, y: 0x%016x, z: 0x%016x},\n", value.x, value.y, value.z)
	}
	return out.String()
}

func tier1ArithmeticTriple128Literals(values []tier1ArithmeticTriple128Spec) string {
	var out strings.Builder
	for _, value := range values {
		fmt.Fprintf(&out, "\t{x: tier1Arithmetic128Words{lo: 0x%016x, hi: 0x%016x}, y: tier1Arithmetic128Words{lo: 0x%016x, hi: 0x%016x}, z: tier1Arithmetic128Words{lo: 0x%016x, hi: 0x%016x}},\n", value.x.lo, value.x.hi, value.y.lo, value.y.hi, value.z.lo, value.z.hi)
	}
	return out.String()
}

func tier1RustUint32Literals(values []uint32) string {
	var out strings.Builder
	for _, value := range values {
		fmt.Fprintf(&out, "    0x%08x,\n", value)
	}
	return out.String()
}

func tier1RustUint64Literals(values []uint64) string {
	var out strings.Builder
	for _, value := range values {
		fmt.Fprintf(&out, "    0x%016x,\n", value)
	}
	return out.String()
}

func tier1RustUint128Literals(values []bid128BidCodecValue) string {
	var out strings.Builder
	for _, value := range values {
		fmt.Fprintf(&out, "    Words { lo: 0x%016x, hi: 0x%016x },\n", value.lo, value.hi)
	}
	return out.String()
}

func tier1RustRoundedOp(operation string) string {
	switch operation {
	case "add":
		return "RoundedOp::Add"
	case "sub":
		return "RoundedOp::Sub"
	case "mul":
		return "RoundedOp::Mul"
	case "div":
		return "RoundedOp::Div"
	case "quantize":
		return "RoundedOp::Quantize"
	default:
		panic("unknown Tier 1 rounded operation: " + operation)
	}
}

func tier1RustRounded32Literals(values []tier1ArithmeticRounded32Spec) string {
	var out strings.Builder
	for _, value := range values {
		fmt.Fprintf(&out, "    Rounded32 { op: %s, x: 0x%08x, y: 0x%08x },\n", tier1RustRoundedOp(value.operation), value.x, value.y)
	}
	return out.String()
}

func tier1RustRounded64Literals(values []tier1ArithmeticRounded64Spec) string {
	var out strings.Builder
	for _, value := range values {
		fmt.Fprintf(&out, "    Rounded64 { op: %s, x: 0x%016x, y: 0x%016x },\n", tier1RustRoundedOp(value.operation), value.x, value.y)
	}
	return out.String()
}

func tier1RustRounded128Literals(values []tier1ArithmeticRounded128Spec) string {
	var out strings.Builder
	for _, value := range values {
		fmt.Fprintf(&out, "    Rounded128 { op: %s, x: Words { lo: 0x%016x, hi: 0x%016x }, y: Words { lo: 0x%016x, hi: 0x%016x } },\n", tier1RustRoundedOp(value.operation), value.x.lo, value.x.hi, value.y.lo, value.y.hi)
	}
	return out.String()
}

func tier1RustScale32Literals(values []tier1ArithmeticScale32Spec) string {
	var out strings.Builder
	for _, value := range values {
		fmt.Fprintf(&out, "    Scale32 { x: 0x%08x, exponent: %d },\n", value.x, value.exponent)
	}
	return out.String()
}

func tier1RustScale64Literals(values []tier1ArithmeticScale64Spec) string {
	var out strings.Builder
	for _, value := range values {
		fmt.Fprintf(&out, "    Scale64 { x: 0x%016x, exponent: %d },\n", value.x, value.exponent)
	}
	return out.String()
}

func tier1RustScale128Literals(values []tier1ArithmeticScale128Spec) string {
	var out strings.Builder
	for _, value := range values {
		fmt.Fprintf(&out, "    Scale128 { x: Words { lo: 0x%016x, hi: 0x%016x }, exponent: %d },\n", value.x.lo, value.x.hi, value.exponent)
	}
	return out.String()
}

func tier1RustTriple32Literals(values []tier1ArithmeticTriple32Spec) string {
	var out strings.Builder
	for _, value := range values {
		fmt.Fprintf(&out, "    Triple32 { x: 0x%08x, y: 0x%08x, z: 0x%08x },\n", value.x, value.y, value.z)
	}
	return out.String()
}

func tier1RustTriple64Literals(values []tier1ArithmeticTriple64Spec) string {
	var out strings.Builder
	for _, value := range values {
		fmt.Fprintf(&out, "    Triple64 { x: 0x%016x, y: 0x%016x, z: 0x%016x },\n", value.x, value.y, value.z)
	}
	return out.String()
}

func tier1RustTriple128Literals(values []tier1ArithmeticTriple128Spec) string {
	var out strings.Builder
	for _, value := range values {
		fmt.Fprintf(&out, "    Triple128 { x: Words { lo: 0x%016x, hi: 0x%016x }, y: Words { lo: 0x%016x, hi: 0x%016x }, z: Words { lo: 0x%016x, hi: 0x%016x } },\n", value.x.lo, value.x.hi, value.y.lo, value.y.hi, value.z.lo, value.z.hi)
	}
	return out.String()
}

func tier1RustPair32Literals(values []tier1ArithmeticPair32Spec) string {
	var out strings.Builder
	for _, value := range values {
		fmt.Fprintf(&out, "    Pair32 { x: 0x%08x, y: 0x%08x },\n", value.x, value.y)
	}
	return out.String()
}

func tier1RustPair64Literals(values []tier1ArithmeticPair64Spec) string {
	var out strings.Builder
	for _, value := range values {
		fmt.Fprintf(&out, "    Pair64 { x: 0x%016x, y: 0x%016x },\n", value.x, value.y)
	}
	return out.String()
}

func tier1RustPair128Literals(values []tier1ArithmeticPair128Spec) string {
	var out strings.Builder
	for _, value := range values {
		fmt.Fprintf(&out, "    Pair128 { x: Words { lo: 0x%016x, hi: 0x%016x }, y: Words { lo: 0x%016x, hi: 0x%016x } },\n", value.x.lo, value.x.hi, value.y.lo, value.y.hi)
	}
	return out.String()
}
