package testgen

import (
	"embed"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/sky1core/bid754/devtools/internal/genmarker"
)

const (
	tier1CompareConversionLongGeneratedPath     = "../bid754-go/generated_ffi_bitcompare_tier1_compare_conversion_long_test.go"
	tier1CompareConversionRustLongGeneratedPath = "../bid754-rs/ffi-verify/tests/tier1_compare_conversion_long_generated.rs"
)

//go:embed tier1_compare_conversion_templates/*
var tier1CompareConversionTemplates embed.FS

func WriteTier1CompareConversionLongOutputs(repoRoot string) error {
	files, err := GenerateTier1CompareConversionLongOutputs()
	if err != nil {
		return err
	}
	for path, data := range files {
		fullPath := filepath.Join(repoRoot, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			return fmt.Errorf("mkdir %q: %w", filepath.Dir(fullPath), err)
		}
		if err := os.WriteFile(fullPath, data, 0o644); err != nil {
			return fmt.Errorf("write generated Tier 1 compare/conversion long test %q: %w", fullPath, err)
		}
	}
	return nil
}

func GenerateTier1CompareConversionLongOutputs() (map[string][]byte, error) {
	template, err := tier1CompareConversionTemplates.ReadFile("tier1_compare_conversion_templates/go_tier1_compare_conversion_long_test.go")
	if err != nil {
		return nil, fmt.Errorf("read Tier 1 compare/conversion long template: %w", err)
	}

	semantic32, semantic64, semantic128, err := tier1ConversionSemanticInputs()
	if err != nil {
		return nil, err
	}
	int32Inputs := tier1ConstructorInt32Inputs()
	uint32Inputs := tier1ConstructorUint32Inputs()
	int64Inputs := tier1ConstructorInt64Inputs()
	uint64Inputs := tier1ConstructorUint64Inputs()
	rustTemplate, err := tier1CompareConversionTemplates.ReadFile("tier1_compare_conversion_templates/rust_tier1_compare_conversion_long.rs")
	if err != nil {
		return nil, fmt.Errorf("read Rust Tier 1 compare/conversion long template: %w", err)
	}
	boundary32 := tier1ArithmeticBoundary32Values()
	boundary64 := bid64BidCodecEdgeValues()
	boundary128 := bid128BidCodecEdgeValues()
	counts := tier1CompareConversionCountsFor(
		uint64(len(boundary32)), uint64(len(boundary64)), uint64(len(boundary128)),
		uint64(len(semantic32)), uint64(len(semantic64)), uint64(len(semantic128)),
		uint64(len(int32Inputs)), uint64(len(uint32Inputs)), uint64(len(int64Inputs)), uint64(len(uint64Inputs)),
	)
	streamHashes := map[string]string{
		"@@TIER1_COMPARE_RANDOM_STREAM_HASH32@@":     fmt.Sprint(tier1CompareConversionRandomStreamHashForGeneration(32, 32, 0xdec75432c04d5001, uint64(1)<<20, 2)),
		"@@TIER1_COMPARE_RANDOM_STREAM_HASH64@@":     fmt.Sprint(tier1CompareConversionRandomStreamHashForGeneration(64, 64, 0xdec75464c04d5001, uint64(1)<<20, 2)),
		"@@TIER1_COMPARE_RANDOM_STREAM_HASH128@@":    fmt.Sprint(tier1CompareConversionRandomStreamHashForGeneration(128, 128, 0xdec754c0c04d5001, uint64(1)<<19, 2)),
		"@@TIER1_TO_INTEGER_RANDOM_STREAM_HASH32@@":  fmt.Sprint(tier1CompareConversionRandomStreamHashForGeneration(32, 32, 0xdec75432c0a70001, uint64(1)<<18, 1)),
		"@@TIER1_TO_INTEGER_RANDOM_STREAM_HASH64@@":  fmt.Sprint(tier1CompareConversionRandomStreamHashForGeneration(64, 64, 0xdec75464c0a70001, uint64(1)<<18, 1)),
		"@@TIER1_TO_INTEGER_RANDOM_STREAM_HASH128@@": fmt.Sprint(tier1CompareConversionRandomStreamHashForGeneration(128, 128, 0xdec754c0c0a70001, uint64(1)<<17, 1)),
		"@@TIER1_WIDTH_RANDOM_STREAM_HASH32@@":       fmt.Sprint(tier1CompareConversionRandomStreamHashForGeneration(32, 32, 0xdec75432c0de0001, uint64(1)<<18, 1)),
		"@@TIER1_WIDTH_RANDOM_STREAM_HASH64@@":       fmt.Sprint(tier1CompareConversionRandomStreamHashForGeneration(64, 64, 0xdec75464c0de0001, uint64(1)<<18, 1)),
		"@@TIER1_WIDTH_RANDOM_STREAM_HASH128@@":      fmt.Sprint(tier1CompareConversionRandomStreamHashForGeneration(128, 128, 0xdec754c0c0de0001, uint64(1)<<17, 1)),
		"@@TIER1_BINARY_RANDOM_STREAM_HASH32@@":      fmt.Sprint(tier1CompareConversionRandomStreamHashForGeneration(32, 32, 0xdec75432c0b10001, uint64(1)<<18, 1)),
		"@@TIER1_BINARY_RANDOM_STREAM_HASH64@@":      fmt.Sprint(tier1CompareConversionRandomStreamHashForGeneration(64, 64, 0xdec75464c0b10001, uint64(1)<<18, 1)),
		"@@TIER1_BINARY_RANDOM_STREAM_HASH128@@":     fmt.Sprint(tier1CompareConversionRandomStreamHashForGeneration(128, 128, 0xdec754c0c0b10001, uint64(1)<<17, 1)),
		"@@TIER1_CONSTRUCTOR_RANDOM_STREAM_HASH@@":   fmt.Sprint(tier1CompareConversionRandomStreamHashForGeneration(64, 0, 0xdec754c0c0570001, uint64(1)<<20, 1)),
	}
	streamHashReplacements := make([]string, 0, len(streamHashes)*2)
	for token, value := range streamHashes {
		streamHashReplacements = append(streamHashReplacements, token, value)
	}
	streamHashReplacer := strings.NewReplacer(streamHashReplacements...)

	// Routing sentinels: deterministic known-answer rows selected and
	// self-asserted by tier1_sentinel_cc_codegen.go; a selection failure
	// aborts the whole generation run with no partial output.
	sentinelRows, err := GenerateTier1CompareConversionSentinelRows()
	if err != nil {
		return nil, err
	}

	replacer := strings.NewReplacer(
		"@@TIER1_CC_SENTINEL_COUNT@@", fmt.Sprint(len(sentinelRows)),
		"@@TIER1_CC_SENTINEL_ROWS@@", tier1SentinelGoRowLiterals(sentinelRows),
		"@@TIER1_BOUNDARY32_COUNT@@", fmt.Sprint(len(tier1ArithmeticBoundary32Values())),
		"@@TIER1_BOUNDARY64_COUNT@@", fmt.Sprint(len(bid64BidCodecEdgeValues())),
		"@@TIER1_BOUNDARY128_COUNT@@", fmt.Sprint(len(bid128BidCodecEdgeValues())),
		"@@TIER1_CONVERSION_SEMANTIC32_COUNT@@", fmt.Sprint(len(semantic32)),
		"@@TIER1_CONVERSION_SEMANTIC64_COUNT@@", fmt.Sprint(len(semantic64)),
		"@@TIER1_CONVERSION_SEMANTIC128_COUNT@@", fmt.Sprint(len(semantic128)),
		"@@TIER1_CONVERSION_SEMANTIC32_VALUES@@", tier1ArithmeticUint32Literals(semantic32),
		"@@TIER1_CONVERSION_SEMANTIC64_VALUES@@", tier1ArithmeticUint64Literals(semantic64),
		"@@TIER1_CONVERSION_SEMANTIC128_VALUES@@", tier1ArithmeticUint128Literals(semantic128),
		"@@TIER1_CONSTRUCTOR_INT32_COUNT@@", fmt.Sprint(len(int32Inputs)),
		"@@TIER1_CONSTRUCTOR_UINT32_COUNT@@", fmt.Sprint(len(uint32Inputs)),
		"@@TIER1_CONSTRUCTOR_INT64_COUNT@@", fmt.Sprint(len(int64Inputs)),
		"@@TIER1_CONSTRUCTOR_UINT64_COUNT@@", fmt.Sprint(len(uint64Inputs)),
		"@@TIER1_CONSTRUCTOR_INT32_VALUES@@", tier1Int32Literals(int32Inputs),
		"@@TIER1_CONSTRUCTOR_UINT32_VALUES@@", tier1Uint32Literals(uint32Inputs),
		"@@TIER1_CONSTRUCTOR_INT64_VALUES@@", tier1Int64Literals(int64Inputs),
		"@@TIER1_CONSTRUCTOR_UINT64_VALUES@@", tier1Uint64Literals(uint64Inputs),
	)
	source := []byte(genmarker.Line("testgen") + "\n" + streamHashReplacer.Replace(replacer.Replace(string(template))))
	goOutputs, err := formatGeneratedGoOutputs(map[string][]byte{
		tier1CompareConversionLongGeneratedPath: source,
	})
	if err != nil {
		return nil, err
	}
	rustReplacer := strings.NewReplacer(
		"@@TIER1_CC_SENTINEL_COUNT@@", fmt.Sprint(len(sentinelRows)),
		"@@TIER1_CC_SENTINEL_ROWS@@", tier1SentinelRustRowLiterals(sentinelRows),
		"@@TIER1_BOUNDARY32_COUNT@@", fmt.Sprint(len(boundary32)),
		"@@TIER1_BOUNDARY64_COUNT@@", fmt.Sprint(len(boundary64)),
		"@@TIER1_BOUNDARY128_COUNT@@", fmt.Sprint(len(boundary128)),
		"@@TIER1_BOUNDARY32_VALUES@@", tier1RustUint32Literals(boundary32),
		"@@TIER1_BOUNDARY64_VALUES@@", tier1RustUint64Literals(boundary64),
		"@@TIER1_BOUNDARY128_VALUES@@", tier1RustUint128Literals(boundary128),
		"@@TIER1_CONVERSION_SEMANTIC32_COUNT@@", fmt.Sprint(len(semantic32)),
		"@@TIER1_CONVERSION_SEMANTIC64_COUNT@@", fmt.Sprint(len(semantic64)),
		"@@TIER1_CONVERSION_SEMANTIC128_COUNT@@", fmt.Sprint(len(semantic128)),
		"@@TIER1_CONVERSION_SEMANTIC32_VALUES@@", tier1RustUint32Literals(semantic32),
		"@@TIER1_CONVERSION_SEMANTIC64_VALUES@@", tier1RustUint64Literals(semantic64),
		"@@TIER1_CONVERSION_SEMANTIC128_VALUES@@", tier1RustUint128Literals(semantic128),
		"@@TIER1_CONSTRUCTOR_INT32_COUNT@@", fmt.Sprint(len(int32Inputs)),
		"@@TIER1_CONSTRUCTOR_UINT32_COUNT@@", fmt.Sprint(len(uint32Inputs)),
		"@@TIER1_CONSTRUCTOR_INT64_COUNT@@", fmt.Sprint(len(int64Inputs)),
		"@@TIER1_CONSTRUCTOR_UINT64_COUNT@@", fmt.Sprint(len(uint64Inputs)),
		"@@TIER1_CONSTRUCTOR_INT32_VALUES@@", tier1RustInt32Literals(int32Inputs),
		"@@TIER1_CONSTRUCTOR_UINT32_VALUES@@", tier1RustUint32ConstructorLiterals(uint32Inputs),
		"@@TIER1_CONSTRUCTOR_INT64_VALUES@@", tier1RustInt64Literals(int64Inputs),
		"@@TIER1_CONSTRUCTOR_UINT64_VALUES@@", tier1RustUint64ConstructorLiterals(uint64Inputs),
		"@@TIER1_COMPARE_STRUCTURED32_COUNT@@", fmt.Sprint(counts.compareStructured32),
		"@@TIER1_COMPARE_STRUCTURED64_COUNT@@", fmt.Sprint(counts.compareStructured64),
		"@@TIER1_COMPARE_STRUCTURED128_COUNT@@", fmt.Sprint(counts.compareStructured128),
		"@@TIER1_COMPARE_RANDOM32_COUNT@@", fmt.Sprint(counts.compareRandom32),
		"@@TIER1_COMPARE_RANDOM64_COUNT@@", fmt.Sprint(counts.compareRandom64),
		"@@TIER1_COMPARE_RANDOM128_COUNT@@", fmt.Sprint(counts.compareRandom128),
		"@@TIER1_COMPARE_TOTAL32_COUNT@@", fmt.Sprint(counts.compareTotal32),
		"@@TIER1_COMPARE_TOTAL64_COUNT@@", fmt.Sprint(counts.compareTotal64),
		"@@TIER1_COMPARE_TOTAL128_COUNT@@", fmt.Sprint(counts.compareTotal128),
		"@@TIER1_TO_INT_STRUCTURED32_COUNT@@", fmt.Sprint(counts.toIntStructured32),
		"@@TIER1_TO_INT_STRUCTURED64_COUNT@@", fmt.Sprint(counts.toIntStructured64),
		"@@TIER1_TO_INT_STRUCTURED128_COUNT@@", fmt.Sprint(counts.toIntStructured128),
		"@@TIER1_TO_INT_TOTAL32_COUNT@@", fmt.Sprint(counts.toIntTotal32),
		"@@TIER1_TO_INT_TOTAL64_COUNT@@", fmt.Sprint(counts.toIntTotal64),
		"@@TIER1_TO_INT_TOTAL128_COUNT@@", fmt.Sprint(counts.toIntTotal128),
		"@@TIER1_WIDTH_STRUCTURED32_COUNT@@", fmt.Sprint(counts.widthStructured32),
		"@@TIER1_WIDTH_STRUCTURED64_COUNT@@", fmt.Sprint(counts.widthStructured64),
		"@@TIER1_WIDTH_STRUCTURED128_COUNT@@", fmt.Sprint(counts.widthStructured128),
		"@@TIER1_WIDTH_TOTAL32_COUNT@@", fmt.Sprint(counts.widthTotal32),
		"@@TIER1_WIDTH_TOTAL64_COUNT@@", fmt.Sprint(counts.widthTotal64),
		"@@TIER1_WIDTH_TOTAL128_COUNT@@", fmt.Sprint(counts.widthTotal128),
		"@@TIER1_CONSTRUCTOR_STRUCTURED_COUNT@@", fmt.Sprint(counts.constructorStructured),
		"@@TIER1_CONSTRUCTOR_TOTAL_COUNT@@", fmt.Sprint(counts.constructorTotal),
		"@@TIER1_CONSTRUCTOR_CONVENIENCE_COUNT@@", fmt.Sprint(counts.constructorConvenience),
		"@@TIER1_CONVERSION_STRUCTURED_COUNT@@", fmt.Sprint(counts.conversionStructured),
		"@@TIER1_CONVERSION_RANDOM_COUNT@@", fmt.Sprint(counts.conversionRandom),
		"@@TIER1_CONVERSION_TOTAL_COUNT@@", fmt.Sprint(counts.conversionTotal),
		"@@TIER1_TO_INT_EXTERN_DECLS@@", tier1RustToIntExternDecls(),
		"@@TIER1_TO_INT_NATIVE_DISPATCH@@", tier1RustToIntNativeDispatch(),
		"@@TIER1_RANDOM_SAMPLE0@@", fmt.Sprintf("0x%016x", tier1ArithmeticRandomWordForGeneration(0xdec75432c04d5001, 0, 0)),
		"@@TIER1_RANDOM_SAMPLE1@@", fmt.Sprintf("0x%016x", tier1ArithmeticRandomWordForGeneration(0xdec75464c0a70001, (uint64(1)<<18)-1, 0)),
		"@@TIER1_RANDOM_SAMPLE2@@", fmt.Sprintf("0x%016x", tier1ArithmeticRandomWordForGeneration(0xdec754c0c0de0001, (uint64(1)<<17)-1, 1)),
	)
	rustSource, err := formatGeneratedRustOutput([]byte(genmarker.Line("testgen") + "\n" + streamHashReplacer.Replace(rustReplacer.Replace(string(rustTemplate)))))
	if err != nil {
		return nil, err
	}
	goOutputs[tier1CompareConversionRustLongGeneratedPath] = rustSource
	return goOutputs, nil
}

// tier1CompareConversionMixOperandForGeneration folds one deterministic-random
// operand draw into the shared FNV stream digest exactly as both generated
// runners consume it: 32-bit draws keep the 32-bit truncation and pad with a
// zero high word, 64-bit draws pad with a zero high word, and 128-bit draws
// split one logical lane into the word lanes (lane*2, lane*2+1).
func tier1CompareConversionMixOperandForGeneration(digest uint64, bits int, seed, caseIndex, lane uint64) uint64 {
	var lo, hi uint64
	switch bits {
	case 32:
		lo = uint64(uint32(tier1ArithmeticRandomWordForGeneration(seed, caseIndex, lane)))
	case 64:
		lo = tier1ArithmeticRandomWordForGeneration(seed, caseIndex, lane)
	case 128:
		lo = tier1ArithmeticRandomWordForGeneration(seed, caseIndex, lane*2)
		hi = tier1ArithmeticRandomWordForGeneration(seed, caseIndex, lane*2+1)
	default:
		panic(fmt.Sprintf("unsupported Tier 1 compare/conversion random stream width %d", bits))
	}
	digest = tier1ArithmeticScaleTupleHashMix(digest, lo)
	return tier1ArithmeticScaleTupleHashMix(digest, hi)
}

// tier1CompareConversionRandomStreamHashForGeneration replicates one
// deterministic-random block of both compare/conversion runners: widthTag is
// mixed first, then operandsPerCase draws per case index in logical-lane order,
// then the case count. Both generated runners accumulate the same digest at
// consumption time and assert it against this generator-anchored constant, so a
// seed, lane, truncation, or draw-order drift in either template fails against
// the anchor. Mode/operation selection is deliberately not folded in: the
// operation tables are pinned by their own count anchors.
func tier1CompareConversionRandomStreamHashForGeneration(bits int, widthTag uint64, seed, cases, operandsPerCase uint64) uint64 {
	digest := tier1ArithmeticScaleTupleHashMix(tier1ArithmeticScaleTupleHashOffset, widthTag)
	for i := uint64(0); i < cases; i++ {
		for lane := uint64(0); lane < operandsPerCase; lane++ {
			digest = tier1CompareConversionMixOperandForGeneration(digest, bits, seed, i, lane)
		}
	}
	return tier1ArithmeticScaleTupleHashMix(digest, cases)
}

type tier1CompareConversionCounts struct {
	compareStructured32, compareStructured64, compareStructured128  uint64
	compareRandom32, compareRandom64, compareRandom128              uint64
	compareTotal32, compareTotal64, compareTotal128                 uint64
	toIntStructured32, toIntStructured64, toIntStructured128        uint64
	toIntTotal32, toIntTotal64, toIntTotal128                       uint64
	widthStructured32, widthStructured64, widthStructured128        uint64
	widthTotal32, widthTotal64, widthTotal128                       uint64
	constructorStructured, constructorTotal, constructorConvenience uint64
	conversionStructured, conversionRandom, conversionTotal         uint64
}

func tier1CompareConversionCountsFor(
	boundary32, boundary64, boundary128 uint64,
	semantic32, semantic64, semantic128 uint64,
	int32Count, uint32Count, int64Count, uint64Count uint64,
) tier1CompareConversionCounts {
	const (
		probes = uint64(12)
		// 6 quiet predicates + 4 min/max operations (minnum, maxnum,
		// minnum_mag, maxnum_mag) per visited pair.
		compareOps = uint64(10)
		toIntOps          = uint64(80)
		toIntRandom32     = uint64(1) << 18
		toIntRandom64     = uint64(1) << 18
		toIntRandom128    = uint64(1) << 17
		widthRandom32     = uint64(1) << 18
		widthRandom64     = uint64(1) << 18
		widthRandom128    = uint64(1) << 17
		// One-way BID -> binary interchange: 3 targets x 5 rounding modes.
		binaryOps         = uint64(15)
		binaryRandom32    = uint64(1) << 18
		binaryRandom64    = uint64(1) << 18
		binaryRandom128   = uint64(1) << 17
		constructorRandom = uint64(1) << 20
	)
	compareStructured := func(boundary uint64) uint64 {
		return (boundary*probes*2 + probes*probes) * compareOps
	}
	result := tier1CompareConversionCounts{
		compareStructured32:    compareStructured(boundary32),
		compareStructured64:    compareStructured(boundary64),
		compareStructured128:   compareStructured(boundary128),
		compareRandom32:        (uint64(1) << 20) * compareOps,
		compareRandom64:        (uint64(1) << 20) * compareOps,
		compareRandom128:       (uint64(1) << 19) * compareOps,
		toIntStructured32:      (boundary32 + semantic32) * toIntOps,
		toIntStructured64:      (boundary64 + semantic64) * toIntOps,
		toIntStructured128:     (boundary128 + semantic128) * toIntOps,
		widthStructured32:      (boundary32 + semantic32) * 2,
		widthStructured64:      (boundary64 + semantic64) * 6,
		widthStructured128:     (boundary128 + semantic128) * 10,
		constructorStructured:  int32Count*7 + uint32Count*7 + int64Count*11 + uint64Count*11,
		constructorConvenience: int32Count + 2*int64Count,
	}
	result.compareTotal32 = result.compareStructured32 + result.compareRandom32
	result.compareTotal64 = result.compareStructured64 + result.compareRandom64
	result.compareTotal128 = result.compareStructured128 + result.compareRandom128
	result.toIntTotal32 = result.toIntStructured32 + toIntRandom32
	result.toIntTotal64 = result.toIntStructured64 + toIntRandom64
	result.toIntTotal128 = result.toIntStructured128 + toIntRandom128
	result.widthTotal32 = result.widthStructured32 + widthRandom32
	result.widthTotal64 = result.widthStructured64 + widthRandom64
	result.widthTotal128 = result.widthStructured128 + widthRandom128
	result.constructorTotal = result.constructorStructured + constructorRandom
	result.conversionStructured = result.toIntStructured32 + result.toIntStructured64 + result.toIntStructured128 +
		result.widthStructured32 + result.widthStructured64 + result.widthStructured128 +
		(boundary32+semantic32)*binaryOps + (boundary64+semantic64)*binaryOps + (boundary128+semantic128)*binaryOps +
		result.constructorStructured
	result.conversionRandom = toIntRandom32 + toIntRandom64 + toIntRandom128 +
		widthRandom32 + widthRandom64 + widthRandom128 +
		binaryRandom32 + binaryRandom64 + binaryRandom128 + constructorRandom
	result.conversionTotal = result.conversionStructured + result.conversionRandom
	return result
}

type tier1DecimalInputSpec struct {
	negative    bool
	coefficient string
	exponent    int32
}

func tier1ConversionSemanticInputs() ([]uint32, []uint64, []bid128BidCodecValue, error) {
	// These are conversion thresholds, half-way cases, and the nearest
	// representable values bracketing integer limits at each decimal width.
	// Special values and raw encoding boundaries come from the shared Tier 1
	// arithmetic boundary corpus and are intentionally not repeated here.
	specs32 := []tier1DecimalInputSpec{
		{coefficient: "1", exponent: -2}, {negative: true, coefficient: "1", exponent: -2},
		{coefficient: "1", exponent: -1}, {negative: true, coefficient: "1", exponent: -1},
		{coefficient: "5", exponent: -1}, {negative: true, coefficient: "5", exponent: -1},
		{coefficient: "15", exponent: -1}, {negative: true, coefficient: "15", exponent: -1},
		{coefficient: "127", exponent: 0}, {negative: true, coefficient: "127", exponent: 0},
		{coefficient: "1275", exponent: -1}, {negative: true, coefficient: "1275", exponent: -1},
		{coefficient: "128", exponent: 0}, {negative: true, coefficient: "128", exponent: 0},
		{coefficient: "255", exponent: 0}, {negative: true, coefficient: "255", exponent: 0},
		{coefficient: "2555", exponent: -1}, {negative: true, coefficient: "2555", exponent: -1},
		{coefficient: "256", exponent: 0}, {negative: true, coefficient: "256", exponent: 0},
		{coefficient: "32767", exponent: 0}, {negative: true, coefficient: "32767", exponent: 0},
		{coefficient: "327675", exponent: -1}, {negative: true, coefficient: "327675", exponent: -1},
		{coefficient: "32768", exponent: 0}, {negative: true, coefficient: "32768", exponent: 0},
		{coefficient: "65535", exponent: 0}, {negative: true, coefficient: "65535", exponent: 0},
		{coefficient: "655355", exponent: -1}, {negative: true, coefficient: "655355", exponent: -1},
		{coefficient: "65536", exponent: 0}, {negative: true, coefficient: "65536", exponent: 0},
		{coefficient: "2147483", exponent: 3}, {negative: true, coefficient: "2147483", exponent: 3},
		{coefficient: "2147484", exponent: 3}, {negative: true, coefficient: "2147484", exponent: 3},
		{coefficient: "4294967", exponent: 3}, {negative: true, coefficient: "4294967", exponent: 3},
		{coefficient: "4294968", exponent: 3}, {negative: true, coefficient: "4294968", exponent: 3},
		{coefficient: "9223372", exponent: 12}, {negative: true, coefficient: "9223372", exponent: 12},
		{coefficient: "9223373", exponent: 12}, {negative: true, coefficient: "9223373", exponent: 12},
		{coefficient: "1844674", exponent: 13}, {negative: true, coefficient: "1844674", exponent: 13},
		{coefficient: "1844675", exponent: 13}, {negative: true, coefficient: "1844675", exponent: 13},
	}
	specs64 := append(append([]tier1DecimalInputSpec{}, specs32...),
		tier1DecimalInputSpec{coefficient: "2147483647", exponent: 0},
		tier1DecimalInputSpec{negative: true, coefficient: "2147483648", exponent: 0},
		tier1DecimalInputSpec{coefficient: "21474836475", exponent: -1},
		tier1DecimalInputSpec{negative: true, coefficient: "21474836485", exponent: -1},
		tier1DecimalInputSpec{coefficient: "4294967295", exponent: 0},
		tier1DecimalInputSpec{coefficient: "42949672955", exponent: -1},
		tier1DecimalInputSpec{coefficient: "9223372036854775", exponent: 3},
		tier1DecimalInputSpec{negative: true, coefficient: "9223372036854775", exponent: 3},
		tier1DecimalInputSpec{coefficient: "9223372036854776", exponent: 3},
		tier1DecimalInputSpec{negative: true, coefficient: "9223372036854776", exponent: 3},
		tier1DecimalInputSpec{coefficient: "1844674407370955", exponent: 4},
		tier1DecimalInputSpec{coefficient: "1844674407370956", exponent: 4},
		// Decimal64 -> Decimal32 precision ties and adjacent values.
		tier1DecimalInputSpec{coefficient: "10000001", exponent: -7},
		tier1DecimalInputSpec{coefficient: "10000005", exponent: -7},
		tier1DecimalInputSpec{coefficient: "99999995", exponent: -7},
		tier1DecimalInputSpec{negative: true, coefficient: "10000001", exponent: -7},
		// Decimal32 overflow and minimum-subnormal rounding boundaries.
		tier1DecimalInputSpec{coefficient: "99999995", exponent: 89},
		tier1DecimalInputSpec{negative: true, coefficient: "99999995", exponent: 89},
		tier1DecimalInputSpec{coefficient: "1", exponent: -102},
		tier1DecimalInputSpec{coefficient: "5", exponent: -102},
		tier1DecimalInputSpec{coefficient: "9", exponent: -102},
		tier1DecimalInputSpec{negative: true, coefficient: "1", exponent: -102},
		tier1DecimalInputSpec{negative: true, coefficient: "5", exponent: -102},
		tier1DecimalInputSpec{negative: true, coefficient: "9", exponent: -102},
	)
	specs128 := append(append([]tier1DecimalInputSpec{}, specs64...),
		tier1DecimalInputSpec{coefficient: "9223372036854775807", exponent: 0},
		tier1DecimalInputSpec{negative: true, coefficient: "9223372036854775808", exponent: 0},
		tier1DecimalInputSpec{coefficient: "92233720368547758075", exponent: -1},
		tier1DecimalInputSpec{coefficient: "92233720368547758085", exponent: -1},
		tier1DecimalInputSpec{negative: true, coefficient: "92233720368547758075", exponent: -1},
		tier1DecimalInputSpec{negative: true, coefficient: "92233720368547758085", exponent: -1},
		tier1DecimalInputSpec{coefficient: "18446744073709551615", exponent: 0},
		tier1DecimalInputSpec{coefficient: "184467440737095516155", exponent: -1},
		tier1DecimalInputSpec{coefficient: "18446744073709551616", exponent: 0},
		// Decimal128 -> Decimal64 precision ties and adjacent values.
		tier1DecimalInputSpec{coefficient: "10000000000000001", exponent: -16},
		tier1DecimalInputSpec{coefficient: "10000000000000005", exponent: -16},
		tier1DecimalInputSpec{coefficient: "99999999999999995", exponent: -16},
		tier1DecimalInputSpec{negative: true, coefficient: "10000000000000001", exponent: -16},
		// Decimal64 overflow and minimum-subnormal rounding boundaries.
		tier1DecimalInputSpec{coefficient: "99999999999999995", exponent: 368},
		tier1DecimalInputSpec{negative: true, coefficient: "99999999999999995", exponent: 368},
		tier1DecimalInputSpec{coefficient: "1", exponent: -399},
		tier1DecimalInputSpec{coefficient: "5", exponent: -399},
		tier1DecimalInputSpec{coefficient: "9", exponent: -399},
		tier1DecimalInputSpec{negative: true, coefficient: "1", exponent: -399},
		tier1DecimalInputSpec{negative: true, coefficient: "5", exponent: -399},
		tier1DecimalInputSpec{negative: true, coefficient: "9", exponent: -399},
	)

	boundary32 := make(map[uint32]struct{})
	for _, value := range tier1ArithmeticBoundary32Values() {
		boundary32[value] = struct{}{}
	}
	boundary64 := make(map[uint64]struct{})
	for _, value := range bid64BidCodecEdgeValues() {
		boundary64[value] = struct{}{}
	}
	boundary128 := make(map[bid128BidCodecValue]struct{})
	for _, value := range bid128BidCodecEdgeValues() {
		boundary128[value] = struct{}{}
	}

	values32 := make(map[uint32]struct{})
	for _, spec := range specs32 {
		value, err := tier1EncodeSemantic32(spec)
		if err != nil {
			return nil, nil, nil, err
		}
		if _, exists := boundary32[value]; !exists {
			values32[value] = struct{}{}
		}
	}
	values64 := make(map[uint64]struct{})
	for _, spec := range specs64 {
		value, err := tier1EncodeSemantic64(spec)
		if err != nil {
			return nil, nil, nil, err
		}
		if _, exists := boundary64[value]; !exists {
			values64[value] = struct{}{}
		}
	}
	values128 := make(map[bid128BidCodecValue]struct{})
	for _, spec := range specs128 {
		value, err := tier1EncodeSemantic128(spec)
		if err != nil {
			return nil, nil, nil, err
		}
		if _, exists := boundary128[value]; !exists {
			values128[value] = struct{}{}
		}
	}

	out32 := make([]uint32, 0, len(values32))
	for value := range values32 {
		out32 = append(out32, value)
	}
	sort.Slice(out32, func(i, j int) bool { return out32[i] < out32[j] })
	out64 := make([]uint64, 0, len(values64))
	for value := range values64 {
		out64 = append(out64, value)
	}
	sort.Slice(out64, func(i, j int) bool { return out64[i] < out64[j] })
	out128 := make([]bid128BidCodecValue, 0, len(values128))
	for value := range values128 {
		out128 = append(out128, value)
	}
	sort.Slice(out128, func(i, j int) bool {
		if out128[i].hi != out128[j].hi {
			return out128[i].hi < out128[j].hi
		}
		return out128[i].lo < out128[j].lo
	})
	return out32, out64, out128, nil
}

func tier1SemanticComponents(spec tier1DecimalInputSpec, maxCoefficient string, minExponent, maxExponent int32) (bidCodecRefComponents, error) {
	coefficient, ok := new(big.Int).SetString(spec.coefficient, 10)
	if !ok || coefficient.Sign() <= 0 {
		return bidCodecRefComponents{}, fmt.Errorf("Tier 1 semantic coefficient %q is not a positive decimal integer", spec.coefficient)
	}
	max, _ := new(big.Int).SetString(maxCoefficient, 10)
	if coefficient.Cmp(max) > 0 {
		return bidCodecRefComponents{}, fmt.Errorf("Tier 1 semantic coefficient %s exceeds format maximum %s", coefficient, max)
	}
	if spec.exponent < minExponent || spec.exponent > maxExponent {
		return bidCodecRefComponents{}, fmt.Errorf("Tier 1 semantic exponent %d is outside [%d,%d]", spec.exponent, minExponent, maxExponent)
	}
	return bidCodecRefComponents{
		Sign:        spec.negative,
		Coefficient: coefficient,
		Exponent:    spec.exponent,
		Kind:        bidCodecRefNormal,
	}, nil
}

func tier1EncodeSemantic32(spec tier1DecimalInputSpec) (uint32, error) {
	components, err := tier1SemanticComponents(spec, "9999999", -101, 90)
	if err != nil {
		return 0, fmt.Errorf("encode Decimal32 Tier 1 semantic input: %w", err)
	}
	return refEncode32(components), nil
}

func tier1EncodeSemantic64(spec tier1DecimalInputSpec) (uint64, error) {
	components, err := tier1SemanticComponents(spec, "9999999999999999", -398, 369)
	if err != nil {
		return 0, fmt.Errorf("encode Decimal64 Tier 1 semantic input: %w", err)
	}
	return refEncode64(components), nil
}

func tier1EncodeSemantic128(spec tier1DecimalInputSpec) (bid128BidCodecValue, error) {
	components, err := tier1SemanticComponents(spec, "9999999999999999999999999999999999", -6176, 6111)
	if err != nil {
		return bid128BidCodecValue{}, fmt.Errorf("encode Decimal128 Tier 1 semantic input: %w", err)
	}
	lo, hi := refEncode128(components)
	return bid128BidCodecValue{lo: lo, hi: hi}, nil
}

func tier1ConstructorInt32Inputs() []int32 {
	return []int32{
		-2147483648, -2147483647, -10000001, -10000000, -9999999, -8388609,
		-8388608, -128, -1, 0, 1, 127, 8388607, 8388608, 9999999,
		10000000, 10000001, 2147483646, 2147483647,
	}
}

func tier1ConstructorUint32Inputs() []uint32 {
	return []uint32{
		0, 1, 127, 255, 65535, 8388607, 8388608, 9999999, 10000000,
		10000001, 2147483647, 2147483648, 4294967294, 4294967295,
	}
}

func tier1ConstructorInt64Inputs() []int64 {
	return []int64{
		-9223372036854775808, -9223372036854775807,
		-10000000000000001, -10000000000000000, -9999999999999999,
		-10000001, -10000000, -9999999, -1, 0, 1, 9999999, 10000000,
		10000001, 9999999999999999, 10000000000000000, 10000000000000001,
		9223372036854775806, 9223372036854775807,
	}
}

func tier1ConstructorUint64Inputs() []uint64 {
	return []uint64{
		0, 1, 127, 255, 65535, 9999999, 10000000, 10000001,
		9999999999999999, 10000000000000000, 10000000000000001,
		9223372036854775807, 9223372036854775808,
		18446744073709551614, 18446744073709551615,
	}
}

func tier1Int32Literals(values []int32) string {
	var out strings.Builder
	for _, value := range values {
		fmt.Fprintf(&out, "\t%d,\n", value)
	}
	return out.String()
}

func tier1Uint32Literals(values []uint32) string {
	var out strings.Builder
	for _, value := range values {
		fmt.Fprintf(&out, "\t%d,\n", value)
	}
	return out.String()
}

func tier1Int64Literals(values []int64) string {
	var out strings.Builder
	for _, value := range values {
		fmt.Fprintf(&out, "\t%d,\n", value)
	}
	return out.String()
}

func tier1Uint64Literals(values []uint64) string {
	var out strings.Builder
	for _, value := range values {
		fmt.Fprintf(&out, "\t%d,\n", value)
	}
	return out.String()
}

func tier1RustInt32Literals(values []int32) string {
	var out strings.Builder
	for _, value := range values {
		fmt.Fprintf(&out, "    %d,\n", value)
	}
	return out.String()
}

func tier1RustUint32ConstructorLiterals(values []uint32) string {
	var out strings.Builder
	for _, value := range values {
		fmt.Fprintf(&out, "    %d,\n", value)
	}
	return out.String()
}

func tier1RustInt64Literals(values []int64) string {
	var out strings.Builder
	for _, value := range values {
		fmt.Fprintf(&out, "    %d,\n", value)
	}
	return out.String()
}

func tier1RustUint64ConstructorLiterals(values []uint64) string {
	var out strings.Builder
	for _, value := range values {
		fmt.Fprintf(&out, "    %d,\n", value)
	}
	return out.String()
}

type tier1RustIntegerKind struct {
	name     string
	enumName string
	rustType string
	signed   bool
}

var tier1RustIntegerKinds = []tier1RustIntegerKind{
	{name: "int8", enumName: "I8", rustType: "i8", signed: true},
	{name: "int16", enumName: "I16", rustType: "i16", signed: true},
	{name: "int32", enumName: "I32", rustType: "i32", signed: true},
	{name: "int64", enumName: "I64", rustType: "i64", signed: true},
	{name: "uint8", enumName: "U8", rustType: "u8"},
	{name: "uint16", enumName: "U16", rustType: "u16"},
	{name: "uint32", enumName: "U32", rustType: "u32"},
	{name: "uint64", enumName: "U64", rustType: "u64"},
}

type tier1RustIntegerSuffix struct {
	name     string
	enumName string
}

var tier1RustIntegerSuffixes = []tier1RustIntegerSuffix{
	{name: "rnint", enumName: "Rnint"},
	{name: "rninta", enumName: "Rninta"},
	{name: "int", enumName: "Int"},
	{name: "ceil", enumName: "Ceil"},
	{name: "floor", enumName: "Floor"},
	{name: "xrnint", enumName: "Xrnint"},
	{name: "xrninta", enumName: "Xrninta"},
	{name: "xint", enumName: "Xint"},
	{name: "xceil", enumName: "Xceil"},
	{name: "xfloor", enumName: "Xfloor"},
}

func tier1RustToIntExternDecls() string {
	var out strings.Builder
	for _, width := range []int{32, 64, 128} {
		operandType := map[int]string{32: "u32", 64: "u64", 128: "C128"}[width]
		for _, kind := range tier1RustIntegerKinds {
			for _, suffix := range tier1RustIntegerSuffixes {
				name := fmt.Sprintf("bid%d_to_%s_%s", width, kind.name, suffix.name)
				fmt.Fprintf(&out, "    #[link_name = \"__%s\"]\n", name)
				fmt.Fprintf(&out, "    fn native_%s(x: %s, flags: *mut u32) -> %s;\n", name, operandType, kind.rustType)
			}
		}
	}
	return out.String()
}

func tier1RustToIntNativeDispatch() string {
	var out strings.Builder
	for _, width := range []int{32, 64, 128} {
		variant := map[int]string{32: "D32", 64: "D64", 128: "D128"}[width]
		operand := "x"
		if width == 128 {
			operand = "c128(x)"
		}
		for _, kind := range tier1RustIntegerKinds {
			for _, suffix := range tier1RustIntegerSuffixes {
				name := fmt.Sprintf("bid%d_to_%s_%s", width, kind.name, suffix.name)
				conversion := " as u64"
				if kind.signed {
					conversion = " as i64 as u64"
				}
				fmt.Fprintf(&out, "        (RawDecimal::%s(x), IntKind::%s, IntSuffix::%s) => native_%s(%s, &mut flags)%s,\n", variant, kind.enumName, suffix.enumName, name, operand, conversion)
			}
		}
	}
	return out.String()
}
