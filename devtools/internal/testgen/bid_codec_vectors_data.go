package testgen

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
)

type bidCodecVector struct {
	Type          string `json:"type"`
	Hex           string `json:"hex"`
	HexHi         string `json:"hex_hi,omitempty"`
	Sign          bool   `json:"sign"`
	Coefficient   string `json:"coefficient"`
	Exponent      int32  `json:"exponent"`
	Kind          string `json:"kind"`
	Payload       string `json:"payload,omitempty"`
	DecimalString string `json:"decimal_string"`
	Canonical     bool   `json:"canonical"`
	EncodedHex    string `json:"encoded_hex"`
	EncodedHi     string `json:"encoded_hi,omitempty"`
}

// bidCodecVectorFormatVersion is 5: the `reject_vectors` domain gained the
// width-independent `to_string` channel, whose Components inputs must fail
// through each language's error mechanism rather than silently losing fields
// or rendering outside the shared schema. Version 4 added the top-level
// `string_vectors` success channel. Version 3 widened the Payload
// field to the full BID128 110-bit NaN payload (values at or above 2^64,
// below 10^33) instead of a low 64-bit truncation.
const bidCodecVectorFormatVersion = 5

type bidCodecVectorFile struct {
	FormatVersion int                    `json:"format_version"`
	Vectors       []bidCodecVector       `json:"vectors"`
	RejectVectors []bidCodecRejectVector `json:"reject_vectors"`
	StringVectors []bidCodecStringVector `json:"string_vectors"`
}

func WriteBidCodecVectorDataOutput(repoRoot string, spec BidCodecVectorSpec) error {
	data, err := GenerateBidCodecVectorData(spec)
	if err != nil {
		return err
	}
	fullPath := filepath.Join(repoRoot, spec.Output)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return fmt.Errorf("mkdir %q: %w", filepath.Dir(fullPath), err)
	}
	if err := os.WriteFile(fullPath, data, 0o644); err != nil {
		return fmt.Errorf("write generated BID codec vectors %q: %w", fullPath, err)
	}
	return nil
}

func GenerateBidCodecVectorData(spec BidCodecVectorSpec) ([]byte, error) {
	if spec.RandomCasesPerFormat <= 0 {
		return nil, fmt.Errorf("BID codec vector random_cases_per_format must be positive")
	}

	rng := rand.New(rand.NewSource(spec.Seed))
	vectors := make([]bidCodecVector, 0, bidCodecVectorCapacity(spec.RandomCasesPerFormat))

	for _, value := range bid32BidCodecEdgeValues() {
		vectors = append(vectors, makeBid32BidCodecVector(value))
	}
	for i := 0; i < spec.RandomCasesPerFormat; i++ {
		vectors = append(vectors, makeBid32BidCodecVector(rng.Uint32()))
	}

	for _, value := range bid64BidCodecEdgeValues() {
		vectors = append(vectors, makeBid64BidCodecVector(value))
	}
	for i := 0; i < spec.RandomCasesPerFormat; i++ {
		vectors = append(vectors, makeBid64BidCodecVector(rng.Uint64()))
	}

	for _, value := range bid128BidCodecEdgeValues() {
		vectors = append(vectors, makeBid128BidCodecVector(value.lo, value.hi))
	}
	for i := 0; i < spec.RandomCasesPerFormat; i++ {
		vectors = append(vectors, makeBid128BidCodecVector(rng.Uint64(), rng.Uint64()))
	}

	file := bidCodecVectorFile{
		FormatVersion: bidCodecVectorFormatVersion,
		Vectors:       vectors,
		RejectVectors: bidCodecRejectVectors(),
		StringVectors: bidCodecStringVectors(),
	}
	data, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal BID codec vectors: %w", err)
	}
	data = append(data, '\n')
	return data, nil
}

func bidCodecVectorCapacity(randomCasesPerFormat int) int {
	return len(bid32BidCodecEdgeValues()) + len(bid64BidCodecEdgeValues()) + len(bid128BidCodecEdgeValues()) + randomCasesPerFormat*3
}

// bid32BidCodecEdgeValues is shared with the deterministic FFI edge matrix
// (ffiEdgeValues in ffi_spec.go), whose binary/ternary tables index into it by
// fixed position. Do not insert, remove, or reorder entries here; a
// codec-only edge addition needs the same format/codec split as bid128 below.
func bid32BidCodecEdgeValues() []uint32 {
	return []uint32{
		0x00000000, 0x80000000,
		0x32800000, 0xb2800000,
		0x32800001, 0xb2800001,
		0x3280000a,
		0x32000001,
		0x77f8967f,
		0x00000001,
		0x78000000, 0xf8000000,
		0x7c000000, 0xfc000000,
		0x7c000001,
		0x7e000000,
		0x7e000001,
		0x60000000,
		0x6cb8967f,
	}
}

// bid64FormatEdgeValues is the format-level BID64 edge corpus shared by the
// BID codec vector generator and the deterministic FFI edge matrix
// (ffiEdgeValues in ffi_spec.go). The FFI binary/ternary edge tables index
// into this list by fixed position, so entries must not be inserted, removed,
// or reordered here; BID-codec-only edge additions belong in
// bid64BidCodecEdgeValues below.
func bid64FormatEdgeValues() []uint64 {
	return []uint64{
		0x0000000000000000, 0x8000000000000000,
		0x31c0000000000000, 0xb1c0000000000000,
		0x31c0000000000001, 0xb1c0000000000001,
		0x31c000000000000a,
		0x31a0000000000001,
		0x6c7386f26fc0ffff,
		0x0000000000000001,
		0x7800000000000000, 0xf800000000000000,
		0x7c00000000000000, 0xfc00000000000000,
		0x7c00000000000001,
		0x7e00000000000000,
		0x7e00000000000001,
	}
}

// bid64BidCodecEdgeValues is the BID codec vector edge corpus: the shared
// format edges followed by the closed combinatorial raw-field boundary set.
// Existing format-edge order stays fixed for the FFI matrix; appendUnique64
// adds only boundary words that are not already present.
func bid64BidCodecEdgeValues() []uint64 {
	return appendUnique64(bid64FormatEdgeValues(), bid64CombinatorialBoundaryValues())
}

func appendUnique64(base, extra []uint64) []uint64 {
	result := append([]uint64(nil), base...)
	seen := make(map[uint64]bool, len(base)+len(extra))
	for _, value := range result {
		seen[value] = true
	}
	for _, value := range extra {
		if !seen[value] {
			result = append(result, value)
			seen[value] = true
		}
	}
	return result
}

func bid64CombinatorialBoundaryValues() []uint64 {
	exponents := [...]uint64{0, 1, 2, 397, 398, 399, 766, 767}
	smallCoefficients := [...]uint64{
		0, 1, 2, 9, 10, 11,
		999999999999998, 999999999999999, 1000000000000000, 1000000000000001,
		0x000fffffffffffff, 0x0010000000000000, 0x0010000000000001,
		0x001ffffffffffffe, 0x001fffffffffffff,
	}
	steeringContinuations := [...]uint64{
		0, 1, 2, 9, 10, 11,
		999999999999998, 999999999999999, 1000000000000000, 1000000000000001,
		0x000386f26fc0fffe, 0x000386f26fc0ffff, 0x000386f26fc10000,
		0x0007fffffffffffe, 0x0007ffffffffffff,
	}
	specialHeaders := [...]uint64{
		0x7800000000000000, 0x7900000000000000,
		0x7a00000000000000, 0x7b00000000000000,
		0x7c00000000000000, 0x7d00000000000000,
		0x7e00000000000000, 0x7f00000000000000,
	}
	reserved := [...]uint64{0, 0x0004000000000000, 0x0200000000000000, 0x03fc000000000000}
	payloads := [...]uint64{
		0, 1, 2, 999999999999998, 999999999999999,
		1000000000000000, 1000000000000001,
		0x0003fffffffffffe, 0x0003ffffffffffff,
	}
	values := make(map[uint64]struct{})
	for _, sign := range [...]uint64{0, 0x8000000000000000} {
		for _, exponent := range exponents {
			for _, coefficient := range smallCoefficients {
				values[sign|(exponent<<53)|coefficient] = struct{}{}
			}
			for _, continuation := range steeringContinuations {
				values[sign|0x6000000000000000|(exponent<<51)|continuation] = struct{}{}
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
	result := make([]uint64, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

type bid128BidCodecValue struct {
	lo uint64
	hi uint64
}

// bid128FormatEdgeValues is the format-level BID128 edge corpus shared by the
// BID codec vector generator and the deterministic FFI edge matrix
// (ffiWideEdgeValues in ffi_spec.go). The FFI binary/ternary edge tables index
// into this list by fixed position, so entries must not be inserted, removed,
// or reordered here; BID-codec-only edge additions belong in
// bid128BidCodecEdgeValues below.
func bid128FormatEdgeValues() []bid128BidCodecValue {
	return []bid128BidCodecValue{
		{lo: 0, hi: 0},
		{lo: 0, hi: 0x8000000000000000},
		{lo: 1, hi: uint64(6176) << 49},
		{lo: 1, hi: 0x8000000000000000 | uint64(6176)<<49},
		{lo: 0x378d8e63ffffffff, hi: 0x0001ed09bead87c0 | uint64(6176)<<49},
		{lo: 0, hi: 0x7800000000000000},
		{lo: 0, hi: 0xf800000000000000},
		{lo: 0, hi: 0x7c00000000000000},
		{lo: 1, hi: 0x7c00000000000000},
		{lo: 0, hi: 0x7e00000000000000},
	}
}

// bid128BidCodecEdgeValues preserves the shared format-edge order, then adds
// the closed combinatorial raw-field boundary set without duplicates.
func bid128BidCodecEdgeValues() []bid128BidCodecValue {
	result := append([]bid128BidCodecValue(nil), bid128FormatEdgeValues()...)
	seen := make(map[bid128BidCodecValue]bool, len(result))
	for _, value := range result {
		seen[value] = true
	}
	for _, value := range bid128CombinatorialBoundaryValues() {
		if !seen[value] {
			result = append(result, value)
			seen[value] = true
		}
	}
	return result
}

func bid128CombinatorialBoundaryValues() []bid128BidCodecValue {
	exponents := [...]uint64{0, 1, 2, 6175, 6176, 6177, 12285, 12286, 12287}
	coefficientHighs := [...]uint64{
		0, 1, 2,
		0x0000314dc6448d92, 0x0000314dc6448d93, 0x0000314dc6448d94,
		0x0001ed09bead87bf, 0x0001ed09bead87c0, 0x0001ed09bead87c1,
		0x0000ffffffffffff, 0x0001000000000000, 0x0001ffffffffffff,
	}
	lows := [...]uint64{
		0, 1, 2,
		0x38c15b09ffffffff, 0x38c15b0a00000000, 0x38c15b0a00000001,
		0x378d8e63ffffffff, 0x378d8e6400000000, 0x378d8e6400000001,
		0x7fffffffffffffff, 0x8000000000000000, 0xfffffffffffffffe, 0xffffffffffffffff,
	}
	steeringHighs := [...]uint64{0, 1, 2, 0x00007ffffffffffe, 0x00007fffffffffff}
	specialHeaders := [...]uint64{
		0x7800000000000000, 0x7900000000000000,
		0x7a00000000000000, 0x7b00000000000000,
		0x7c00000000000000, 0x7d00000000000000,
		0x7e00000000000000, 0x7f00000000000000,
	}
	reserved := [...]uint64{0, 0x0000400000000000, 0x0200000000000000, 0x03ffc00000000000}
	payloadHighs := [...]uint64{
		0, 1, 2,
		0x0000314dc6448d92, 0x0000314dc6448d93, 0x0000314dc6448d94,
		0x00003ffffffffffe, 0x00003fffffffffff,
	}
	values := make(map[bid128BidCodecValue]struct{})
	for _, sign := range [...]uint64{0, 0x8000000000000000} {
		for _, exponent := range exponents {
			for _, coefficientHi := range coefficientHighs {
				for _, lo := range lows {
					value := bid128BidCodecValue{lo: lo, hi: sign | (exponent << 49) | coefficientHi}
					values[value] = struct{}{}
				}
			}
			for _, steeringHi := range steeringHighs {
				for _, lo := range lows {
					value := bid128BidCodecValue{
						lo: lo, hi: sign | 0x6000000000000000 | (exponent << 47) | steeringHi,
					}
					values[value] = struct{}{}
				}
			}
		}
		for _, header := range specialHeaders {
			for _, reservedBits := range reserved {
				for _, payloadHi := range payloadHighs {
					for _, lo := range lows {
						value := bid128BidCodecValue{
							lo: lo, hi: sign | header | reservedBits | payloadHi,
						}
						values[value] = struct{}{}
					}
				}
			}
		}
	}
	result := make([]bid128BidCodecValue, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].hi != result[j].hi {
			return result[i].hi < result[j].hi
		}
		return result[i].lo < result[j].lo
	})
	return result
}

func makeBid32BidCodecVector(value uint32) bidCodecVector {
	components := refDecode32(value)
	encoded := refEncode32(components)
	vector := bidCodecVector{
		Type:          "bid32",
		Hex:           fmt.Sprintf("%08x", value),
		Sign:          components.Sign,
		Exponent:      components.Exponent,
		Kind:          bidCodecKindString(components.Kind),
		DecimalString: bidCodecDecimalString(components),
		Canonical:     encoded == value,
		EncodedHex:    fmt.Sprintf("%08x", encoded),
	}
	if components.Coefficient != nil {
		vector.Coefficient = components.Coefficient.String()
	}
	if payloadNonZero(components.Payload) {
		vector.Payload = components.Payload.String()
	}
	return vector
}

func makeBid64BidCodecVector(value uint64) bidCodecVector {
	components := refDecode64(value)
	encoded := refEncode64(components)
	vector := bidCodecVector{
		Type:          "bid64",
		Hex:           fmt.Sprintf("%016x", value),
		Sign:          components.Sign,
		Exponent:      components.Exponent,
		Kind:          bidCodecKindString(components.Kind),
		DecimalString: bidCodecDecimalString(components),
		Canonical:     encoded == value,
		EncodedHex:    fmt.Sprintf("%016x", encoded),
	}
	if components.Coefficient != nil {
		vector.Coefficient = components.Coefficient.String()
	}
	if payloadNonZero(components.Payload) {
		vector.Payload = components.Payload.String()
	}
	return vector
}

func makeBid128BidCodecVector(lo, hi uint64) bidCodecVector {
	components := refDecode128(lo, hi)
	encodedLo, encodedHi := refEncode128(components)
	vector := bidCodecVector{
		Type:          "bid128",
		Hex:           fmt.Sprintf("%016x", lo),
		HexHi:         fmt.Sprintf("%016x", hi),
		Sign:          components.Sign,
		Exponent:      components.Exponent,
		Kind:          bidCodecKindString(components.Kind),
		DecimalString: bidCodecDecimalString(components),
		Canonical:     encodedLo == lo && encodedHi == hi,
		EncodedHex:    fmt.Sprintf("%016x", encodedLo),
		EncodedHi:     fmt.Sprintf("%016x", encodedHi),
	}
	if components.Coefficient != nil {
		vector.Coefficient = components.Coefficient.String()
	}
	if payloadNonZero(components.Payload) {
		vector.Payload = components.Payload.String()
	}
	return vector
}

func bidCodecDecimalString(c bidCodecRefComponents) string {
	prefix := "+"
	if c.Sign {
		prefix = "-"
	}
	switch c.Kind {
	case bidCodecRefInfinity:
		return prefix + "Inf"
	case bidCodecRefQNaN:
		if payloadNonZero(c.Payload) {
			return fmt.Sprintf("%sNaN%s", prefix, c.Payload.String())
		}
		return prefix + "NaN"
	case bidCodecRefSNaN:
		if payloadNonZero(c.Payload) {
			return fmt.Sprintf("%sSNaN%s", prefix, c.Payload.String())
		}
		return prefix + "SNaN"
	case bidCodecRefZero:
		if c.Exponent == 0 {
			return prefix + "0"
		}
		return fmt.Sprintf("%s0E%+d", prefix, c.Exponent)
	}

	digits := c.Coefficient.String()
	exp := int(c.Exponent) + len(digits) - 1
	if len(digits) == 1 {
		return fmt.Sprintf("%s%sE%+d", prefix, digits, exp)
	}
	return fmt.Sprintf("%s%s.%sE%+d", prefix, digits[:1], digits[1:], exp)
}

func bidCodecKindString(kind bidCodecRefKind) string {
	switch kind {
	case bidCodecRefNormal:
		return "normal"
	case bidCodecRefZero:
		return "zero"
	case bidCodecRefInfinity:
		return "inf"
	case bidCodecRefQNaN:
		return "qnan"
	case bidCodecRefSNaN:
		return "snan"
	default:
		return "unknown"
	}
}
