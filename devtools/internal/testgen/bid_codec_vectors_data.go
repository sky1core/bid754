package testgen

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
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

const bidCodecVectorFormatVersion = 1

type bidCodecVectorFile struct {
	FormatVersion int              `json:"format_version"`
	Vectors       []bidCodecVector `json:"vectors"`
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

func bid64BidCodecEdgeValues() []uint64 {
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

type bid128BidCodecValue struct {
	lo uint64
	hi uint64
}

func bid128BidCodecEdgeValues() []bid128BidCodecValue {
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
	if components.Payload != 0 {
		vector.Payload = fmt.Sprintf("%d", components.Payload)
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
	if components.Payload != 0 {
		vector.Payload = fmt.Sprintf("%d", components.Payload)
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
	if components.Payload != 0 {
		vector.Payload = fmt.Sprintf("%d", components.Payload)
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
		if c.Payload != 0 {
			return fmt.Sprintf("%sNaN%d", prefix, c.Payload)
		}
		return prefix + "NaN"
	case bidCodecRefSNaN:
		if c.Payload != 0 {
			return fmt.Sprintf("%sSNaN%d", prefix, c.Payload)
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
