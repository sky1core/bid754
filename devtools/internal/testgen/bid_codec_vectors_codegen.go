package testgen

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sky1core/bid754/devtools/internal/genmarker"
)

const (
	bidCodecVectorsGoTestPath             = "../bid754-codec-go/vector_test.go"
	bidCodecVectorsGoExternalTestPath     = "../bid754-codec-go/testdata/external_vector_test.go"
	bidCodecVectorsGoExhaustive32TestPath = "../bid754-codec-go/exhaustive32_long_test.go"
	bidCodecVectorsGoLong64And128TestPath = "../bid754-codec-go/decimal64_128_long_test.go"
	bidCodecVectorsRustTestPath           = "../bid754-rs/tests/bid_codec_vectors.rs"
	bidCodecVectorsGoFullTestPath         = "../bid754-go/generated_bid_codec_vectors_test.go"
	bidCodecVectorsStandaloneRustTestPath = "../bid754-codec-rs/tests/vectors.rs"
	bidCodecVectorsJavaRunnerPath         = "../bid754-codec-java/src/test/java/io/github/sky1core/bidcodec/VectorRunner.java"
	bidCodecVectorsJavaTestPath           = "../bid754-codec-java/src/test/java/io/github/sky1core/bidcodec/VectorTest.java"
	bidCodecVectorsPythonTestPath         = "../bid754-codec-py/tests/test_vectors.py"
	bidCodecVectorsJSTestPath             = "../bid754-codec-js/src/vectors.test.ts"
	bidCodecVectorsJSRunnerPath           = "../bid754-codec-js/vector_runner.mjs"
	bidCodecVectorsSwiftRunnerPath        = "../bid754-codec-swift/Sources/BidCodecVectorRunner/main.swift"

	bidCodecExpectedVectorTotal            = 23545
	bidCodecExpectedBid32Vectors           = 5019
	bidCodecExpectedBid64Vectors           = 5804
	bidCodecExpectedBid128Vectors          = 12722
	bidCodecExpectedBid32CanonicalVectors  = 4464
	bidCodecExpectedBid64CanonicalVectors  = 4583
	bidCodecExpectedBid128CanonicalVectors = 6047
)

//go:embed bidcodec_templates/*
var bidCodecConsumerTemplates embed.FS

func WriteBidCodecVectorTestOutputs(repoRoot string) error {
	files, err := GenerateBidCodecVectorTestOutputs()
	if err != nil {
		return err
	}
	for path, data := range files {
		fullPath := filepath.Join(repoRoot, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			return fmt.Errorf("mkdir %q: %w", filepath.Dir(fullPath), err)
		}
		if err := os.WriteFile(fullPath, data, 0o644); err != nil {
			return fmt.Errorf("write generated BID codec vector test %q: %w", fullPath, err)
		}
	}
	return nil
}

func bidCodecConsumerTemplate(name, header string) []byte {
	data, err := bidCodecConsumerTemplates.ReadFile(filepath.Join("bidcodec_templates", name))
	if err != nil {
		panic(fmt.Sprintf("read BID codec consumer template %q: %v", name, err))
	}
	data = []byte(applyBidCodecConsumerTemplateReplacements(string(data)))
	out := make([]byte, 0, len(header)+len(data))
	out = append(out, header...)
	out = append(out, data...)
	return out
}

func applyBidCodecConsumerTemplateReplacements(src string) string {
	replacements := map[string]string{
		"{{BID_CODEC_VECTOR_FORMAT_VERSION}}": fmt.Sprintf("%d", bidCodecVectorFormatVersion),
		"{{BID_CODEC_GO_ANCHORS}}":            bidCodecGoAnchorSnippet(),
		"{{BID_CODEC_RUST_ANCHORS}}":          bidCodecRustAnchorSnippet(),
		"{{BID_CODEC_PYTHON_ANCHORS}}":        bidCodecPythonAnchorSnippet(),
		"{{BID_CODEC_JS_ANCHORS}}":            bidCodecJSAnchorSnippet(),
		"{{BID_CODEC_JS_ANCHOR_ARRAY}}":       bidCodecJSAnchorArray(),
		"{{BID_CODEC_JAVA_ANCHOR_JSON}}":      bidCodecJavaAnchorJSONLiteral(),
		"{{BID_CODEC_SWIFT_ANCHORS}}":         bidCodecSwiftAnchorSnippet(),
		"{{BID_CODEC_JAVA_ANCHOR_JSON_RAW}}":  bidCodecAnchorJSON(),
		"{{BID_CODEC_VECTOR_ANCHOR_COUNT}}":   fmt.Sprintf("%d", len(bidCodecVectorAnchors)),
		"{{BID_CODEC_VECTOR_TOTAL}}":          fmt.Sprintf("%d", bidCodecExpectedVectorTotal),
		"{{BID_CODEC_BID32_TOTAL}}":           fmt.Sprintf("%d", bidCodecExpectedBid32Vectors),
		"{{BID_CODEC_BID64_TOTAL}}":           fmt.Sprintf("%d", bidCodecExpectedBid64Vectors),
		"{{BID_CODEC_BID128_TOTAL}}":          fmt.Sprintf("%d", bidCodecExpectedBid128Vectors),
		"{{BID_CODEC_BID32_CANONICAL}}":       fmt.Sprintf("%d", bidCodecExpectedBid32CanonicalVectors),
		"{{BID_CODEC_BID64_CANONICAL}}":       fmt.Sprintf("%d", bidCodecExpectedBid64CanonicalVectors),
		"{{BID_CODEC_BID128_CANONICAL}}":      fmt.Sprintf("%d", bidCodecExpectedBid128CanonicalVectors),
		// reject_vectors domain: total record count (hardcode-verified by every
		// consumer, pinned externally in verification_anchors.json), per-language
		// capability tag sets, and the per-language type-domain reject cases.
		"{{BID_CODEC_REJECT_TOTAL}}": fmt.Sprintf("%d", len(bidCodecRejectVectors())),
		// string_vectors domain: total success-channel record count. The channel
		// is capability-ungated (pure ASCII string pairs), so every language
		// consumer consumes exactly this many records; the same total is pinned
		// externally in verification_anchors.json.
		"{{BID_CODEC_STRING_TOTAL}}":          fmt.Sprintf("%d", len(bidCodecStringVectors())),
		"{{BID_CODEC_GO_REJECT_CAPS}}":        bidCodecRejectCapsElems("go"),
		"{{BID_CODEC_RUST_REJECT_CAPS}}":      bidCodecRejectCapsElems("rust"),
		"{{BID_CODEC_RUST_FULL_REJECT_CAPS}}": bidCodecRejectCapsElems("rust_full"),
		"{{BID_CODEC_JAVA_REJECT_CAPS}}":      bidCodecRejectCapsElems("java"),
		"{{BID_CODEC_PYTHON_REJECT_CAPS}}":    bidCodecRejectCapsElems("python"),
		"{{BID_CODEC_JS_REJECT_CAPS}}":        bidCodecRejectCapsElems("js"),
		"{{BID_CODEC_SWIFT_REJECT_CAPS}}":     bidCodecRejectCapsElems("swift"),
		"{{BID_CODEC_GO_REJECT_TYPE_DOMAIN}}": bidCodecGoTypeDomainElems(),
		// go_full consumer: the bid754-go full library's public parse surface
		// consumes the from_string reject channel and the string_vectors
		// channel against generator-owned expectation classes; the remaining
		// reject channels are channel-skipped (no public Components surface).
		"{{BID_CODEC_GO_FULL_FROM_STRING_CLASSES}}": bidCodecGoFullFromStringClassElems(),
		"{{BID_CODEC_GO_FULL_STRING_CLASSES}}":      bidCodecGoFullStringClassElems(),
		"{{BID_CODEC_PY_REJECT_TYPE_DOMAIN}}":       bidCodecPyTypeDomainElems(),
		"{{BID_CODEC_JS_REJECT_TYPE_DOMAIN}}":       bidCodecJsTypeDomainElems(),
		"{{BID_CODEC_PY_RAW_DECODE_REJECTS}}":       bidCodecPyRawDecodeRejectElems(),
		"{{BID_CODEC_JS_RAW_DECODE_REJECTS}}":       bidCodecJsRawDecodeRejectElems(),
	}
	// go_full consumed/skipped pins are channel-derived, not capability-derived:
	// the from_string channel is the consumed set, every other channel is the
	// skipped set, so the go_full consumer does not join the capability loop
	// below.
	goFullConsumed, goFullSkipped := bidCodecGoFullRejectCounts()
	replacements["{{BID_CODEC_GO_FULL_REJECT_CONSUMED}}"] = fmt.Sprintf("%d", goFullConsumed)
	replacements["{{BID_CODEC_GO_FULL_REJECT_SKIPPED}}"] = fmt.Sprintf("%d", goFullSkipped)
	// Per-language consumed/skipped pins and the unsupported-tag set: each
	// consumer asserts its exact consumption split and that every skipped
	// record's requires tag is a declared-unsupported one, so a generator
	// regression that shifts records into the skipped bucket (or invents a new
	// tag) fails the runner instead of passing as "all skipped".
	for placeholder, lang := range map[string]string{
		"{{BID_CODEC_GO_REJECT":        "go",
		"{{BID_CODEC_RUST_REJECT":      "rust",
		"{{BID_CODEC_RUST_FULL_REJECT": "rust_full",
		"{{BID_CODEC_JAVA_REJECT":      "java",
		"{{BID_CODEC_PYTHON_REJECT":    "python",
		"{{BID_CODEC_JS_REJECT":        "js",
		"{{BID_CODEC_SWIFT_REJECT":     "swift",
	} {
		consumed, skipped := bidCodecRejectExpectedCounts(lang)
		replacements[placeholder+"_CONSUMED}}"] = fmt.Sprintf("%d", consumed)
		replacements[placeholder+"_SKIPPED}}"] = fmt.Sprintf("%d", skipped)
		replacements[placeholder+"_UNSUPPORTED}}"] = bidCodecRejectUnsupportedElems(lang)
	}
	for old, newValue := range replacements {
		src = strings.ReplaceAll(src, old, newValue)
	}
	return src
}

func GenerateBidCodecVectorTestOutputs() (map[string][]byte, error) {
	files := map[string][]byte{
		bidCodecVectorsGoExhaustive32TestPath: bidCodecConsumerTemplate(
			"go_exhaustive32_long_test.go",
			genmarker.Line("testgen")+"\n",
		),
		bidCodecVectorsGoLong64And128TestPath: bidCodecConsumerTemplate(
			"go_decimal64_128_long_test.go",
			genmarker.Line("testgen")+"\n",
		),
		bidCodecVectorsStandaloneRustTestPath: bidCodecConsumerTemplate(
			"standalone_rust_vectors.rs",
			genmarker.Line("testgen")+"\n",
		),
		bidCodecVectorsJavaRunnerPath: bidCodecConsumerTemplate(
			"java_vector_runner.java",
			genmarker.Line("testgen")+"\n",
		),
		bidCodecVectorsJavaTestPath: bidCodecConsumerTemplate(
			"java_vector_test.java",
			genmarker.Line("testgen")+"\n",
		),
		bidCodecVectorsPythonTestPath: bidCodecConsumerTemplate(
			"python_test_vectors.py",
			genmarker.HashLine("testgen")+"\n",
		),
		bidCodecVectorsJSTestPath: bidCodecConsumerTemplate(
			"js_vectors.test.ts",
			genmarker.Line("testgen")+"\n",
		),
		bidCodecVectorsJSRunnerPath: bidCodecConsumerTemplate(
			"js_vector_runner.mjs",
			genmarker.Line("testgen")+"\n",
		),
		bidCodecVectorsSwiftRunnerPath: bidCodecConsumerTemplate(
			"swift_vector_runner.swift",
			genmarker.Line("testgen")+"\n",
		),
		bidCodecVectorsGoTestPath: []byte(applyBidCodecConsumerTemplateReplacements(genmarker.Line("testgen") + `
//go:build bid754_bidcodec_vectors

package bidcodec

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"testing"
)

type vectorEntry struct {
	Type        string ` + "`json:\"type\"`" + `
	Hex         string ` + "`json:\"hex\"`" + `
	HexHi       string ` + "`json:\"hex_hi,omitempty\"`" + `
	Sign        bool   ` + "`json:\"sign\"`" + `
	Coefficient string ` + "`json:\"coefficient\"`" + `
	Exponent    int    ` + "`json:\"exponent\"`" + `
	Kind        string ` + "`json:\"kind\"`" + `
	Payload     string ` + "`json:\"payload,omitempty\"`" + `
	DecimalString string ` + "`json:\"decimal_string\"`" + `
	Canonical   bool   ` + "`json:\"canonical\"`" + `
	EncodedHex  string ` + "`json:\"encoded_hex\"`" + `
	EncodedHi   string ` + "`json:\"encoded_hi,omitempty\"`" + `
}

type vectorFile struct {
	FormatVersion int           ` + "`json:\"format_version\"`" + `
	Vectors       []vectorEntry ` + "`json:\"vectors\"`" + `
	RejectVectors []rejectEntry ` + "`json:\"reject_vectors\"`" + `
	StringVectors []stringEntry ` + "`json:\"string_vectors\"`" + `
}

// stringEntry is one string_vectors success record: FromString(Input) must
// succeed and ToString of the result must equal Expected exactly.
type stringEntry struct {
	Input    string ` + "`json:\"input\"`" + `
	Expected string ` + "`json:\"expected\"`" + `
}

type rejectEntry struct {
	Channel     string ` + "`json:\"channel\"`" + `
	Type        string ` + "`json:\"type\"`" + `
	Input       string ` + "`json:\"input\"`" + `
	Sign        bool   ` + "`json:\"sign\"`" + `
	Kind        string ` + "`json:\"kind\"`" + `
	Coefficient string ` + "`json:\"coefficient\"`" + `
	Exponent    int32  ` + "`json:\"exponent\"`" + `
	Payload     string ` + "`json:\"payload\"`" + `
	Reason      string ` + "`json:\"reason\"`" + `
	Requires    string ` + "`json:\"requires\"`" + `
}

const (
	expectedFormatVersion = {{BID_CODEC_VECTOR_FORMAT_VERSION}}
	expectedRejectTotal = {{BID_CODEC_REJECT_TOTAL}}
	expectedRejectConsumed = {{BID_CODEC_GO_REJECT_CONSUMED}}
	expectedRejectSkipped = {{BID_CODEC_GO_REJECT_SKIPPED}}
	expectedStringVectorTotal = {{BID_CODEC_STRING_TOTAL}}
	expectedVectorTotal = {{BID_CODEC_VECTOR_TOTAL}}
	expectedBid32Vectors = {{BID_CODEC_BID32_TOTAL}}
	expectedBid64Vectors = {{BID_CODEC_BID64_TOTAL}}
	expectedBid128Vectors = {{BID_CODEC_BID128_TOTAL}}
	expectedBid32CanonicalVectors = {{BID_CODEC_BID32_CANONICAL}}
	expectedBid64CanonicalVectors = {{BID_CODEC_BID64_CANONICAL}}
	expectedBid128CanonicalVectors = {{BID_CODEC_BID128_CANONICAL}}
)

func loadVectors(t *testing.T) []vectorEntry {
	t.Helper()
	data, err := os.ReadFile("../bid754-codec-vectors/vectors.json")
	if err != nil {
		t.Fatalf("failed to read vectors.json: %v", err)
	}
	var file vectorFile
	if err := json.Unmarshal(data, &file); err != nil {
		t.Fatalf("failed to parse vectors.json: %v", err)
	}
	if file.FormatVersion != expectedFormatVersion {
		t.Fatalf("unsupported BID codec vectors format_version %d, want %d", file.FormatVersion, expectedFormatVersion)
	}
	return file.Vectors
}

func loadRejectVectors(t *testing.T) []rejectEntry {
	t.Helper()
	data, err := os.ReadFile("../bid754-codec-vectors/vectors.json")
	if err != nil {
		t.Fatalf("failed to read vectors.json: %v", err)
	}
	var file vectorFile
	if err := json.Unmarshal(data, &file); err != nil {
		t.Fatalf("failed to parse vectors.json: %v", err)
	}
	if file.FormatVersion != expectedFormatVersion {
		t.Fatalf("unsupported BID codec vectors format_version %d, want %d", file.FormatVersion, expectedFormatVersion)
	}
	return file.RejectVectors
}

func loadStringVectors(t *testing.T) []stringEntry {
	t.Helper()
	data, err := os.ReadFile("../bid754-codec-vectors/vectors.json")
	if err != nil {
		t.Fatalf("failed to read vectors.json: %v", err)
	}
	var file vectorFile
	if err := json.Unmarshal(data, &file); err != nil {
		t.Fatalf("failed to parse vectors.json: %v", err)
	}
	if file.FormatVersion != expectedFormatVersion {
		t.Fatalf("unsupported BID codec vectors format_version %d, want %d", file.FormatVersion, expectedFormatVersion)
	}
	return file.StringVectors
}

{{BID_CODEC_GO_ANCHORS}}

func parseHex64(s string) uint64 {
	v, _ := strconv.ParseUint(s, 16, 64)
	return v
}

func parseHex32(s string) uint32 {
	v, _ := strconv.ParseUint(s, 16, 32)
	return uint32(v)
}

func parseKind(s string) Kind {
	switch s {
	case "normal":
		return Normal
	case "zero":
		return Zero
	case "inf":
		return Infinity
	case "qnan":
		return QNaN
	case "snan":
		return SNaN
	}
	panic("unknown kind: " + s)
}

func bid32Bytes(s string) []byte {
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], parseHex32(s))
	return b[:]
}

func bid64Bytes(s string) []byte {
	var b [8]byte
	binary.LittleEndian.PutUint64(b[:], parseHex64(s))
	return b[:]
}

func bid128Bytes(loHex, hiHex string) []byte {
	var b [16]byte
	binary.LittleEndian.PutUint64(b[0:8], parseHex64(loHex))
	binary.LittleEndian.PutUint64(b[8:16], parseHex64(hiHex))
	return b[:]
}

func TestBidCodecVectors(t *testing.T) {
	vectors := loadVectors(t)

	var bid32Count, bid64Count, bid128Count int
	var bid32CanonicalCount, bid64CanonicalCount, bid128CanonicalCount int
	var bid32NonCanonicalCount, bid64NonCanonicalCount, bid128NonCanonicalCount int
	var decodePass, decodeFail int
	var roundtripPass, roundtripFail int

	for i, v := range vectors {
		label := fmt.Sprintf("[%d] %s hex=%s", i, v.Type, v.Hex)

		switch v.Type {
		case "bid32":
			bid32Count++
			if v.Canonical {
				bid32CanonicalCount++
			} else {
				bid32NonCanonicalCount++
			}
			raw := parseHex32(v.Hex)
			c := Decode32(raw)

			if !checkDecode(t, label, c, v) {
				decodeFail++
			} else {
				decodePass++
			}

			checkStringRoundtrip(t, label, c, v)
			checkBytesRoundtrip(t, label, c, v)

			if v.Canonical {
				encoded, err := Encode32(c)
				gotHex := fmt.Sprintf("%08x", encoded)
				if err != nil || gotHex != v.EncodedHex {
					t.Errorf("%s roundtrip: got %s err=%v, want %s", label, gotHex, err, v.EncodedHex)
					roundtripFail++
				} else {
					roundtripPass++
				}
			}

		case "bid64":
			bid64Count++
			if v.Canonical {
				bid64CanonicalCount++
			} else {
				bid64NonCanonicalCount++
			}
			raw := parseHex64(v.Hex)
			c := Decode64(raw)

			if !checkDecode(t, label, c, v) {
				decodeFail++
			} else {
				decodePass++
			}

			checkStringRoundtrip(t, label, c, v)
			checkBytesRoundtrip(t, label, c, v)

			if v.Canonical {
				encoded, err := Encode64(c)
				gotHex := fmt.Sprintf("%016x", encoded)
				if err != nil || gotHex != v.EncodedHex {
					t.Errorf("%s roundtrip: got %s err=%v, want %s", label, gotHex, err, v.EncodedHex)
					roundtripFail++
				} else {
					roundtripPass++
				}
			}

		case "bid128":
			bid128Count++
			if v.Canonical {
				bid128CanonicalCount++
			} else {
				bid128NonCanonicalCount++
			}
			lo := parseHex64(v.Hex)
			hi := parseHex64(v.HexHi)
			c := Decode128(lo, hi)

			if !checkDecode(t, label, c, v) {
				decodeFail++
			} else {
				decodePass++
			}

			checkStringRoundtrip(t, label, c, v)
			checkBytesRoundtrip(t, label, c, v)

			if v.Canonical {
				gotLo, gotHi, err := Encode128(c)
				gotLoHex := fmt.Sprintf("%016x", gotLo)
				gotHiHex := fmt.Sprintf("%016x", gotHi)
				if err != nil || gotLoHex != v.EncodedHex || gotHiHex != v.EncodedHi {
					t.Errorf("%s roundtrip: got lo=%s hi=%s err=%v, want lo=%s hi=%s",
						label, gotLoHex, gotHiHex, err, v.EncodedHex, v.EncodedHi)
					roundtripFail++
				} else {
					roundtripPass++
				}
			}

		default:
			t.Fatalf("unknown type: %s", v.Type)
		}
	}

	if len(vectors) != expectedVectorTotal {
		t.Fatalf("vector total = %d, want %d", len(vectors), expectedVectorTotal)
	}
	if bid32Count != expectedBid32Vectors || bid64Count != expectedBid64Vectors || bid128Count != expectedBid128Vectors {
		t.Fatalf("per-format vector counts = bid32:%d bid64:%d bid128:%d, want bid32:%d bid64:%d bid128:%d",
			bid32Count, bid64Count, bid128Count, expectedBid32Vectors, expectedBid64Vectors, expectedBid128Vectors)
	}
	if bid32CanonicalCount != expectedBid32CanonicalVectors || bid64CanonicalCount != expectedBid64CanonicalVectors || bid128CanonicalCount != expectedBid128CanonicalVectors {
		t.Fatalf("canonical vector counts = bid32:%d bid64:%d bid128:%d, want bid32:%d bid64:%d bid128:%d",
			bid32CanonicalCount, bid64CanonicalCount, bid128CanonicalCount,
			expectedBid32CanonicalVectors, expectedBid64CanonicalVectors, expectedBid128CanonicalVectors)
	}
	if bid32NonCanonicalCount != expectedBid32Vectors-expectedBid32CanonicalVectors ||
		bid64NonCanonicalCount != expectedBid64Vectors-expectedBid64CanonicalVectors ||
		bid128NonCanonicalCount != expectedBid128Vectors-expectedBid128CanonicalVectors {
		t.Fatalf("non-canonical vector counts = bid32:%d bid64:%d bid128:%d",
			bid32NonCanonicalCount, bid64NonCanonicalCount, bid128NonCanonicalCount)
	}
	if decodePass != expectedVectorTotal || decodeFail != 0 {
		t.Fatalf("decode coverage = %d pass, %d fail; want %d pass, 0 fail", decodePass, decodeFail, expectedVectorTotal)
	}
	if roundtripPass != expectedBid32CanonicalVectors+expectedBid64CanonicalVectors+expectedBid128CanonicalVectors || roundtripFail != 0 {
		t.Fatalf("roundtrip coverage = %d pass, %d fail", roundtripPass, roundtripFail)
	}

	t.Logf("vectors: %d total (bid32=%d, bid64=%d, bid128=%d)",
		len(vectors), bid32Count, bid64Count, bid128Count)
	t.Logf("decode: %d pass, %d fail", decodePass, decodeFail)
	t.Logf("roundtrip: %d pass, %d fail", roundtripPass, roundtripFail)
}

func TestBidCodecAnchorVectors(t *testing.T) {
	for _, v := range anchorVectors {
		label := fmt.Sprintf("anchor %s hex=%s", v.Type, v.Hex)
		if !v.Canonical {
			t.Fatalf("%s canonical = false, want true", label)
		}
		var c Components
		switch v.Type {
		case "bid32":
			c = Decode32(parseHex32(v.Hex))
		case "bid64":
			c = Decode64(parseHex64(v.Hex))
		case "bid128":
			c = Decode128(parseHex64(v.Hex), parseHex64(v.HexHi))
		default:
			t.Fatalf("unknown anchor type: %s", v.Type)
		}
		if !checkDecode(t, label, c, v) {
			continue
		}
		if c.Exponent != int32(v.Exponent) {
			t.Errorf("%s exponent: got %d, want %d", label, c.Exponent, v.Exponent)
		}
		checkStringRoundtrip(t, label, c, v)
		checkBytesRoundtrip(t, label, c, v)
	}
}

func TestBidCodecErrorSemantics(t *testing.T) {
	if _, err := Decode32Bytes(make([]byte, 3)); err == nil {
		t.Fatal("Decode32Bytes short input succeeded, want error")
	}
	if _, err := Decode32Bytes(make([]byte, 5)); err == nil {
		t.Fatal("Decode32Bytes long input succeeded, want error")
	}
	if _, err := Decode64Bytes(make([]byte, 7)); err == nil {
		t.Fatal("Decode64Bytes short input succeeded, want error")
	}
	if _, err := Decode64Bytes(make([]byte, 9)); err == nil {
		t.Fatal("Decode64Bytes long input succeeded, want error")
	}
	if _, err := Decode128Bytes(make([]byte, 15)); err == nil {
		t.Fatal("Decode128Bytes short input succeeded, want error")
	}
	if _, err := Decode128Bytes(make([]byte, 17)); err == nil {
		t.Fatal("Decode128Bytes long input succeeded, want error")
	}
	// reject_vectors: the generated malformed-input and out-of-range-Components
	// reject domain replaces the previous hardcoded malformed list. Each record
	// must fail through the language error mechanism (parse failure for
	// from_string, encode rejection for encode); records whose Components field
	// types this language cannot construct are skipped with a reported reason.
	rejects := loadRejectVectors(t)
	if len(rejects) != expectedRejectTotal {
		t.Fatalf("reject_vectors total = %d, want %d", len(rejects), expectedRejectTotal)
	}
	caps := map[string]bool{}
	for _, capTag := range []string{ {{BID_CODEC_GO_REJECT_CAPS}} } {
		caps[capTag] = true
	}
	unsupportedCaps := map[string]bool{}
	for _, capTag := range []string{ {{BID_CODEC_GO_REJECT_UNSUPPORTED}} } {
		unsupportedCaps[capTag] = true
	}
	consumed, skipped := 0, 0
	skipReasons := map[string]int{}
	for _, r := range rejects {
		if r.Requires != "" && !caps[r.Requires] {
			if !unsupportedCaps[r.Requires] {
				t.Fatalf("reject record requires tag %q outside the declared capability universe", r.Requires)
			}
			skipped++
			skipReasons[r.Requires]++
			continue
		}
		consumed++
		switch r.Channel {
		case "from_string":
			if _, err := FromString(r.Input); err == nil {
				t.Errorf("reject from_string %q (%s): expected error", r.Input, r.Reason)
			}
		case "encode":
			if err := encodeReject(t, r); err == nil {
				t.Errorf("reject encode %s %s (%s): expected error", r.Type, r.Kind, r.Reason)
			}
		case "to_string":
			if err := toStringReject(t, r); err == nil {
				t.Errorf("reject to_string %s (%s): expected error", r.Kind, r.Reason)
			}
		default:
			t.Fatalf("unknown reject channel %q", r.Channel)
		}
	}
	if consumed != expectedRejectConsumed || skipped != expectedRejectSkipped || consumed+skipped != len(rejects) {
		t.Fatalf("reject consumption = consumed %d skipped %d of %d, want consumed %d skipped %d",
			consumed, skipped, len(rejects), expectedRejectConsumed, expectedRejectSkipped)
	}
	t.Logf("reject_vectors: consumed=%d skipped=%d skipReasons=%v", consumed, skipped, skipReasons)

	// type-domain rejects: values the shared JSON schema cannot express, emitted
	// directly for the languages that can construct them (Go: kind is an integer
	// type, so a value outside the defined set is constructible).
	for _, tc := range []struct {
		id string
		c  Components
	}{ {{BID_CODEC_GO_REJECT_TYPE_DOMAIN}} } {
		if _, err := Encode32(tc.c); err == nil {
			t.Errorf("type-domain reject %s: Encode32 expected error", tc.id)
		}
		if _, err := ToString(tc.c); err == nil {
			t.Errorf("type-domain reject %s: ToString expected error", tc.id)
		}
	}
}

func TestBidCodecStringVectors(t *testing.T) {
	// string_vectors: the generated SUCCESS channel for the string surface.
	// Each record's input must parse and re-render as the exact expected
	// string. This pins FromString→ToString agreement across all language
	// consumers in the encoding-unreachable Components region (above all
	// int32-extreme exponents, where a 32-bit adjusted-exponent computation
	// silently wraps the rendered sign) plus the successful grammar-edge
	// normalizations — territory neither the bit-decode vectors channel nor
	// the failure-side reject_vectors channel reaches. The closure leg then
	// re-parses the expected rendering itself: FromString(expected) must
	// succeed and ToString must reproduce it exactly (parse(render(x)) is
	// total and expected is a rendering fixed point), so a parser that rejects
	// its own renderer's output fails here in every language.
	stringVectors := loadStringVectors(t)
	if len(stringVectors) != expectedStringVectorTotal {
		t.Fatalf("string_vectors total = %d, want %d", len(stringVectors), expectedStringVectorTotal)
	}
	consumed := 0
	for _, sv := range stringVectors {
		consumed++
		c, err := FromString(sv.Input)
		if err != nil {
			t.Errorf("string_vectors FromString(%q): unexpected error: %v", sv.Input, err)
			continue
		}
		if got, err := ToString(c); err != nil || got != sv.Expected {
			t.Errorf("string_vectors %q: ToString got %q, want %q", sv.Input, got, sv.Expected)
			continue
		}
		reparsed, err := FromString(sv.Expected)
		if err != nil {
			t.Errorf("string_vectors closure FromString(%q): rendering not reparseable: %v", sv.Expected, err)
			continue
		}
		if got, err := ToString(reparsed); err != nil || got != sv.Expected {
			t.Errorf("string_vectors closure %q: re-rendered as %q, not a fixed point", sv.Expected, got)
		}
	}
	if consumed != expectedStringVectorTotal {
		t.Fatalf("string_vectors consumption = %d, want %d", consumed, expectedStringVectorTotal)
	}
	t.Logf("string_vectors: consumed=%d", consumed)
}

func encodeReject(t *testing.T, r rejectEntry) error {
	t.Helper()
	c := rejectComponents(t, r)
	switch r.Type {
	case "bid32":
		_, err := Encode32(c)
		return err
	case "bid64":
		_, err := Encode64(c)
		return err
	case "bid128":
		_, _, err := Encode128(c)
		return err
	default:
		t.Fatalf("reject encode: unknown type %q", r.Type)
		return nil
	}
}

func toStringReject(t *testing.T, r rejectEntry) error {
	t.Helper()
	_, err := ToString(rejectComponents(t, r))
	return err
}

func rejectComponents(t *testing.T, r rejectEntry) Components {
	t.Helper()
	c := Components{Sign: r.Sign, Kind: parseKind(r.Kind), Exponent: r.Exponent}
	if r.Coefficient != "" {
		coeff, ok := new(big.Int).SetString(r.Coefficient, 10)
		if !ok {
			t.Fatalf("reject encode %s: bad coefficient %q", r.Type, r.Coefficient)
		}
		c.Coefficient = coeff
	}
	if r.Payload != "" {
		p, ok := new(big.Int).SetString(r.Payload, 10)
		if !ok {
			t.Fatalf("reject encode %s: bad payload %q", r.Type, r.Payload)
		}
		c.Payload = p
	}
	return c
}

func checkDecode(t *testing.T, label string, c Components, v vectorEntry) bool {
	t.Helper()
	ok := true

	if c.Sign != v.Sign {
		t.Errorf("%s sign: got %v, want %v", label, c.Sign, v.Sign)
		ok = false
	}

	wantKind := parseKind(v.Kind)
	if c.Kind != wantKind {
		t.Errorf("%s kind: got %v, want %v", label, c.Kind, wantKind)
		ok = false
	}

	wantCoeff := v.Coefficient
	var gotCoeff string
	if c.Coefficient != nil {
		gotCoeff = c.Coefficient.String()
	}
	if gotCoeff != wantCoeff {
		t.Errorf("%s coefficient: got %q, want %q", label, gotCoeff, wantCoeff)
		ok = false
	}

	if c.Kind == Normal || c.Kind == Zero {
		wantExp := int32(v.Exponent)
		if c.Exponent != wantExp {
			t.Errorf("%s exponent: got %d, want %d", label, c.Exponent, wantExp)
			ok = false
		}
	}

	if v.Payload != "" {
		wantPayload, valid := new(big.Int).SetString(v.Payload, 10)
		gotPayload := c.Payload
		if gotPayload == nil {
			gotPayload = big.NewInt(0)
		}
		if !valid || gotPayload.Cmp(wantPayload) != 0 {
			t.Errorf("%s payload: got %v, want %s", label, c.Payload, v.Payload)
			ok = false
		}
	}

	return ok
}

func checkBytesRoundtrip(t *testing.T, label string, c Components, v vectorEntry) {
	t.Helper()
	switch v.Type {
	case "bid32":
		decoded, err := Decode32Bytes(bid32Bytes(v.Hex))
		if err != nil {
			t.Errorf("%s decode_bytes32: %v", label, err)
			return
		}
		if !checkDecode(t, label+" decode_bytes32", decoded, v) {
			return
		}
		if v.Canonical {
			got, err := Encode32Bytes(c)
			want := bid32Bytes(v.EncodedHex)
			if err != nil || !bytes.Equal(got[:], want) {
				t.Errorf("%s encode_bytes32: got %x err=%v, want %x", label, got, err, want)
			}
		}
	case "bid64":
		decoded, err := Decode64Bytes(bid64Bytes(v.Hex))
		if err != nil {
			t.Errorf("%s decode_bytes64: %v", label, err)
			return
		}
		if !checkDecode(t, label+" decode_bytes64", decoded, v) {
			return
		}
		if v.Canonical {
			got, err := Encode64Bytes(c)
			want := bid64Bytes(v.EncodedHex)
			if err != nil || !bytes.Equal(got[:], want) {
				t.Errorf("%s encode_bytes64: got %x err=%v, want %x", label, got, err, want)
			}
		}
	case "bid128":
		decoded, err := Decode128Bytes(bid128Bytes(v.Hex, v.HexHi))
		if err != nil {
			t.Errorf("%s decode_bytes128: %v", label, err)
			return
		}
		if !checkDecode(t, label+" decode_bytes128", decoded, v) {
			return
		}
		if v.Canonical {
			got, err := Encode128Bytes(c)
			want := bid128Bytes(v.EncodedHex, v.EncodedHi)
			if err != nil || !bytes.Equal(got[:], want) {
				t.Errorf("%s encode_bytes128: got %x err=%v, want %x", label, got, err, want)
			}
		}
	default:
		t.Fatalf("unknown type: %s", v.Type)
	}
}

func checkStringRoundtrip(t *testing.T, label string, c Components, v vectorEntry) {
	t.Helper()
	gotString, err := ToString(c)
	if err != nil {
		t.Errorf("%s to_string: unexpected error: %v", label, err)
		return
	}
	if gotString != v.DecimalString {
		t.Errorf("%s to_string: got %q, want %q", label, gotString, v.DecimalString)
	}
	parsed, err := FromString(v.DecimalString)
	if err != nil {
		t.Errorf("%s from_string(%q): %v", label, v.DecimalString, err)
		return
	}
	switch v.Type {
	case "bid32":
		enc, err := Encode32(parsed)
		got := fmt.Sprintf("%08x", enc)
		if err != nil || got != v.EncodedHex {
			t.Errorf("%s from_string encode32: got %s err=%v, want %s", label, got, err, v.EncodedHex)
		}
	case "bid64":
		enc, err := Encode64(parsed)
		got := fmt.Sprintf("%016x", enc)
		if err != nil || got != v.EncodedHex {
			t.Errorf("%s from_string encode64: got %s err=%v, want %s", label, got, err, v.EncodedHex)
		}
	case "bid128":
		gotLo, gotHi, err := Encode128(parsed)
		gotLoHex := fmt.Sprintf("%016x", gotLo)
		gotHiHex := fmt.Sprintf("%016x", gotHi)
		if err != nil || gotLoHex != v.EncodedHex || gotHiHex != v.EncodedHi {
			t.Errorf("%s from_string encode128: got lo=%s hi=%s err=%v, want lo=%s hi=%s",
				label, gotLoHex, gotHiHex, err, v.EncodedHex, v.EncodedHi)
		}
	}
}
`)),
		bidCodecVectorsRustTestPath: []byte(applyBidCodecConsumerTemplateReplacements(genmarker.Line("testgen") + `
// Additional (non-required) BID codec vector consumer: the bid754-rs full
// library's embedded bid_codec module, which mirrors the standalone
// bid754-codec-rs contract (validating Result encode APIs, strict ASCII
// from_string, full 110-bit u128 payload). It consumes the ` + "`vectors`" + ` array,
// the ` + "`reject_vectors`" + ` domain (same capability rules as the standalone Rust
// consumer), and the ` + "`string_vectors`" + ` success domain, so the two
// intentionally-duplicated Rust codec copies are held to the same generated
// contract.
use bid754::bid_codec::{decode128, decode32, decode64, encode128, encode32, encode64, from_string, to_string, Components, Kind};
use serde::Deserialize;
use std::fs;
use std::path::PathBuf;

#[derive(Debug, Deserialize)]
struct Vector {
    #[serde(rename = "type")]
    typ: String,
    hex: String,
    #[serde(default)]
    hex_hi: Option<String>,
    sign: bool,
    #[serde(default)]
    coefficient: String,
    exponent: i32,
    kind: String,
    decimal_string: String,
    canonical: bool,
    #[serde(default)]
    payload: Option<String>,
    #[serde(default)]
    encoded_hex: Option<String>,
    #[serde(default)]
    encoded_hi: Option<String>,
}

#[derive(Debug, Deserialize)]
struct RejectVector {
    channel: String,
    #[serde(default, rename = "type")]
    typ: String,
    #[serde(default)]
    input: String,
    #[serde(default)]
    sign: bool,
    #[serde(default)]
    kind: String,
    #[serde(default)]
    coefficient: String,
    #[serde(default)]
    exponent: i32,
    #[serde(default)]
    payload: String,
    reason: String,
    #[serde(default)]
    requires: Option<String>,
}

/// One string_vectors success record: from_string(input) must succeed and
/// to_string of the result must equal expected exactly.
#[derive(Debug, Deserialize)]
struct StringVector {
    input: String,
    expected: String,
}

#[derive(Debug, Deserialize)]
struct VectorFile {
    format_version: u32,
    vectors: Vec<Vector>,
    reject_vectors: Vec<RejectVector>,
    string_vectors: Vec<StringVector>,
}

const EXPECTED_FORMAT_VERSION: u32 = {{BID_CODEC_VECTOR_FORMAT_VERSION}};
const EXPECTED_REJECT_TOTAL: usize = {{BID_CODEC_REJECT_TOTAL}};
const EXPECTED_REJECT_CONSUMED: usize = {{BID_CODEC_RUST_FULL_REJECT_CONSUMED}};
const EXPECTED_REJECT_SKIPPED: usize = {{BID_CODEC_RUST_FULL_REJECT_SKIPPED}};
const EXPECTED_STRING_TOTAL: usize = {{BID_CODEC_STRING_TOTAL}};
const EXPECTED_TOTAL: usize = {{BID_CODEC_VECTOR_TOTAL}};
const EXPECTED_BID32: usize = {{BID_CODEC_BID32_TOTAL}};
const EXPECTED_BID64: usize = {{BID_CODEC_BID64_TOTAL}};
const EXPECTED_BID128: usize = {{BID_CODEC_BID128_TOTAL}};
const EXPECTED_BID32_CANONICAL: usize = {{BID_CODEC_BID32_CANONICAL}};
const EXPECTED_BID64_CANONICAL: usize = {{BID_CODEC_BID64_CANONICAL}};
const EXPECTED_BID128_CANONICAL: usize = {{BID_CODEC_BID128_CANONICAL}};

fn vectors_path() -> PathBuf {
    PathBuf::from(env!("CARGO_MANIFEST_DIR")).join("../bid754-codec-vectors/vectors.json")
}

fn load_vectors() -> Vec<Vector> {
    let data = fs::read_to_string(vectors_path()).expect("failed to read BID codec vectors");
    let file: VectorFile = serde_json::from_str(&data).expect("failed to parse BID codec vectors");
    assert_eq!(
        file.format_version, EXPECTED_FORMAT_VERSION,
        "unsupported BID codec vectors format_version"
    );
    file.vectors
}

fn load_reject_vectors() -> Vec<RejectVector> {
    let data = fs::read_to_string(vectors_path()).expect("failed to read BID codec vectors");
    let file: VectorFile = serde_json::from_str(&data).expect("failed to parse BID codec vectors");
    assert_eq!(
        file.format_version, EXPECTED_FORMAT_VERSION,
        "unsupported BID codec vectors format_version"
    );
    file.reject_vectors
}

fn load_string_vectors() -> Vec<StringVector> {
    let data = fs::read_to_string(vectors_path()).expect("failed to read BID codec vectors");
    let file: VectorFile = serde_json::from_str(&data).expect("failed to parse BID codec vectors");
    assert_eq!(
        file.format_version, EXPECTED_FORMAT_VERSION,
        "unsupported BID codec vectors format_version"
    );
    file.string_vectors
}

#[test]
fn test_bid_codec_string_vectors() {
    // string_vectors: the generated SUCCESS channel for the string surface.
    // Each record's input must parse and re-render as the exact expected
    // string, pinning from_string→to_string agreement across all language
    // consumers in the encoding-unreachable Components region (above all
    // int32-extreme exponents whose adjusted exponent exceeds int32) plus the
    // successful grammar-edge normalizations. The closure leg then re-parses
    // the expected rendering itself: from_string(expected) must succeed and
    // to_string must reproduce it exactly (parse(render(x)) is total and
    // expected is a rendering fixed point), so a parser that rejects its own
    // renderer's output fails here. The channel is capability-ungated: every
    // record is consumed.
    let string_vectors = load_string_vectors();
    assert_eq!(string_vectors.len(), EXPECTED_STRING_TOTAL, "string_vectors total changed");
    let mut consumed = 0usize;
    let mut failures = Vec::new();
    for sv in &string_vectors {
        consumed += 1;
        match from_string(&sv.input) {
            Ok(c) => {
                let got = to_string(&c).expect("string_vectors parsed Components must render");
                if got != sv.expected {
                    failures.push(format!(
                        "string_vectors {:?}: to_string got {:?}, want {:?}",
                        sv.input, got, sv.expected
                    ));
                    continue;
                }
                match from_string(&sv.expected) {
                    Ok(reparsed) => {
                        let again = to_string(&reparsed).expect("string_vectors reparsed Components must render");
                        if again != sv.expected {
                            failures.push(format!(
                                "string_vectors closure {:?}: re-rendered as {:?}, not a fixed point",
                                sv.expected, again
                            ));
                        }
                    }
                    Err(err) => {
                        failures.push(format!(
                            "string_vectors closure from_string {:?}: rendering not reparseable: {err}",
                            sv.expected
                        ));
                    }
                }
            }
            Err(err) => {
                failures.push(format!("string_vectors from_string {:?} failed: {err}", sv.input));
            }
        }
    }
    assert!(failures.is_empty(), "string_vectors failures:\n{}", failures.join("\n"));
    assert_eq!(consumed, EXPECTED_STRING_TOTAL, "string_vectors consumed count changed");
    eprintln!("string_vectors: consumed={consumed}");
}

fn reject_components(r: &RejectVector) -> Components {
    // Record-field parsing happens outside the rejection assertion: an altered
    // record fails the harness here instead of passing as an API rejection.
    Components {
        sign: r.sign,
        coefficient: if r.coefficient.is_empty() {
            0
        } else {
            r.coefficient.parse().expect("reject coefficient parse")
        },
        exponent: r.exponent,
        kind: parse_kind(&r.kind),
        payload: if r.payload.is_empty() {
            0
        } else {
            r.payload.parse().expect("reject payload parse")
        },
    }
}

#[test]
fn test_bid_codec_reject_vectors() {
    // Every reject record must fail through the language error mechanism: a
    // parse failure for the from_string channel, an encode/to_string rejection
    // for the corresponding Components channel. Records whose value this module's fixed-width Components
    // types cannot construct (a coefficient at/above 2^128, a negative
    // coefficient, a negative payload) are skipped with a reported reason --
    // the u128 fields have no such capability, so every requires-tagged record
    // is skipped here, exactly like the standalone Rust consumer.
    let rejects = load_reject_vectors();
    assert_eq!(rejects.len(), EXPECTED_REJECT_TOTAL, "reject_vectors total changed");
    let caps: &[&str] = &[{{BID_CODEC_RUST_FULL_REJECT_CAPS}}];
    let unsupported_caps: &[&str] = &[{{BID_CODEC_RUST_FULL_REJECT_UNSUPPORTED}}];
    let mut consumed = 0usize;
    let mut skipped = 0usize;
    let mut skip_reasons: std::collections::BTreeMap<String, usize> = std::collections::BTreeMap::new();
    let mut failures = Vec::new();
    for r in &rejects {
        if let Some(req) = r.requires.as_deref() {
            if !caps.contains(&req) {
                assert!(
                    unsupported_caps.contains(&req),
                    "reject record requires tag {req:?} outside the declared capability universe"
                );
                skipped += 1;
                *skip_reasons.entry(req.to_string()).or_insert(0) += 1;
                continue;
            }
        }
        consumed += 1;
        match r.channel.as_str() {
            "from_string" => {
                if from_string(&r.input).is_ok() {
                    failures.push(format!("from_string {:?} ({}) accepted, want error", r.input, r.reason));
                }
            }
            "encode" => {
                let c = reject_components(r);
                let rejected = match r.typ.as_str() {
                    "bid32" => encode32(&c).is_err(),
                    "bid64" => encode64(&c).is_err(),
                    "bid128" => encode128(&c).is_err(),
                    other => panic!("unknown reject encode type: {other}"),
                };
                if !rejected {
                    failures.push(format!("encode {} {} ({}) accepted, want error", r.typ, r.kind, r.reason));
                }
            }
            "to_string" => {
                let c = reject_components(r);
                if to_string(&c).is_ok() {
                    failures.push(format!("to_string {} ({}) accepted, want error", r.kind, r.reason));
                }
            }
            other => panic!("unknown reject channel: {other}"),
        }
    }
    assert!(failures.is_empty(), "reject_vectors failures:\n{}", failures.join("\n"));
    assert_eq!(consumed, EXPECTED_REJECT_CONSUMED, "reject consumed count changed");
    assert_eq!(skipped, EXPECTED_REJECT_SKIPPED, "reject skipped count changed");
    assert_eq!(consumed + skipped, rejects.len(), "reject consumption does not partition the reject set");
    eprintln!("reject_vectors: consumed={consumed} skipped={skipped} skip_reasons={skip_reasons:?}");
}

{{BID_CODEC_RUST_ANCHORS}}

fn parse_kind(kind: &str) -> Kind {
    match kind {
        "zero" => Kind::Zero,
        "normal" => Kind::Normal,
        "inf" => Kind::Infinity,
        "qnan" => Kind::QNaN,
        "snan" => Kind::SNaN,
        other => panic!("unknown kind: {other}"),
    }
}

fn hex_u32(s: &str) -> u32 {
    u32::from_str_radix(s, 16).unwrap()
}

fn hex_u64(s: &str) -> u64 {
    u64::from_str_radix(s, 16).unwrap()
}

fn coeff_u128(s: &str) -> u128 {
    if s.is_empty() {
        0
    } else {
        s.parse().unwrap()
    }
}

fn payload_u128(v: &Vector) -> u128 {
    v.payload
        .as_deref()
        .map(|s| s.parse().unwrap_or(0))
        .unwrap_or(0)
}

#[test]
fn test_bid_codec_anchor_vectors() {
    let vectors = anchor_vectors();
    assert_eq!(vectors.len(), {{BID_CODEC_VECTOR_ANCHOR_COUNT}}, "BID codec anchor count changed");
    let mut failures = Vec::new();
    for vector in &vectors {
        let expected_kind = parse_kind(&vector.kind);
        match vector.typ.as_str() {
            "bid32" => {
                let decoded = decode32(hex_u32(&vector.hex));
                if decoded.kind != expected_kind
                    || decoded.sign != vector.sign
                    || decoded.exponent != vector.exponent
                    || decoded.coefficient != coeff_u128(&vector.coefficient)
                    || payload_u128(vector) != decoded.payload
                    || to_string(&decoded).expect("BID32 anchor must render") != vector.decimal_string
                    || encode32(&decoded).expect("anchor encode32") != hex_u32(vector.encoded_hex.as_deref().unwrap())
                {
                    failures.push(format!("bid32 anchor {} mismatch: got {:?}", vector.hex, decoded));
                }
            }
            "bid64" => {
                let decoded = decode64(hex_u64(&vector.hex));
                if decoded.kind != expected_kind
                    || decoded.sign != vector.sign
                    || decoded.exponent != vector.exponent
                    || decoded.coefficient != coeff_u128(&vector.coefficient)
                    || payload_u128(vector) != decoded.payload
                    || to_string(&decoded).expect("BID64 anchor must render") != vector.decimal_string
                    || encode64(&decoded).expect("anchor encode64") != hex_u64(vector.encoded_hex.as_deref().unwrap())
                {
                    failures.push(format!("bid64 anchor {} mismatch: got {:?}", vector.hex, decoded));
                }
            }
            "bid128" => {
                let decoded = decode128(hex_u64(&vector.hex), hex_u64(vector.hex_hi.as_deref().unwrap()));
                let (encoded_lo, encoded_hi) = encode128(&decoded).expect("anchor encode128");
                if decoded.kind != expected_kind
                    || decoded.sign != vector.sign
                    || decoded.exponent != vector.exponent
                    || (!matches!(decoded.kind, Kind::QNaN | Kind::SNaN)
                        && decoded.coefficient != coeff_u128(&vector.coefficient))
                    || payload_u128(vector) != decoded.payload
                    || to_string(&decoded).expect("BID128 anchor must render") != vector.decimal_string
                    || encoded_lo != hex_u64(vector.encoded_hex.as_deref().unwrap())
                    || encoded_hi != hex_u64(vector.encoded_hi.as_deref().unwrap())
                {
                    failures.push(format!(
                        "bid128 anchor {}_{} mismatch: got {:?}",
                        vector.hex_hi.as_deref().unwrap_or(""),
                        vector.hex,
                        decoded
                    ));
                }
            }
            other => failures.push(format!("unknown anchor vector type: {other}")),
        }
    }
    assert!(failures.is_empty(), "BID codec anchor failures:\n{}", failures.join("\n"));
}

#[test]
fn test_bid_codec_vectors_encode_decode() {
    let vectors = load_vectors();
    let mut bid32 = 0usize;
    let mut bid64 = 0usize;
    let mut bid128 = 0usize;
    let mut bid32_canonical = 0usize;
    let mut bid64_canonical = 0usize;
    let mut bid128_canonical = 0usize;
    let mut decode_passed = 0usize;
    let mut encode_passed = 0usize;
    let mut skipped = 0usize;
    let mut failures = Vec::new();

    for vector in &vectors {
        let expected_kind = parse_kind(&vector.kind);
        match vector.typ.as_str() {
            "bid32" => {
                bid32 += 1;
                if vector.canonical {
                    bid32_canonical += 1;
                }
                let bits = hex_u32(&vector.hex);
                let decoded = decode32(bits);
                if decoded.kind != expected_kind
                    || decoded.sign != vector.sign
                    || decoded.exponent != vector.exponent
                    || decoded.coefficient != coeff_u128(&vector.coefficient)
                {
                    failures.push(format!(
                        "bid32 decode {}: got kind={:?} sign={} coeff={} exp={}, want kind={:?} sign={} coeff={} exp={}",
                        vector.hex,
                        decoded.kind,
                        decoded.sign,
                        decoded.coefficient,
                        decoded.exponent,
                        expected_kind,
                        vector.sign,
                        coeff_u128(&vector.coefficient),
                        vector.exponent,
                    ));
                    continue;
                }
                if matches!(decoded.kind, Kind::QNaN | Kind::SNaN) && decoded.payload != payload_u128(vector) {
                    failures.push(format!(
                        "bid32 decode payload {}: got {}, want {}",
                        vector.hex,
                        decoded.payload,
                        payload_u128(vector)
                    ));
                    continue;
                }
                decode_passed += 1;
                if to_string(&decoded).expect("decoded BID32 Components must render") != vector.decimal_string {
                    failures.push(format!(
                        "bid32 to_string {}: got {}, want {}",
                        vector.hex,
                        to_string(&decoded).expect("decoded BID32 Components must render"),
                        vector.decimal_string
                    ));
                    continue;
                }
                let parsed = from_string(&vector.decimal_string).unwrap_or_else(|err| {
                    panic!("bid32 from_string {} failed: {}", vector.decimal_string, err)
                });
                let parsed_encoded = encode32(&parsed).expect("from_string encode32");
                let expected_from_string = hex_u32(vector.encoded_hex.as_deref().unwrap_or(&vector.hex));
                if parsed_encoded != expected_from_string {
                    failures.push(format!(
                        "bid32 from_string {}: got {:08x}, want {:08x}",
                        vector.decimal_string,
                        parsed_encoded,
                        expected_from_string
                    ));
                    continue;
                }

                if vector.canonical {
                    let encoded = encode32(&decoded).expect("canonical encode32");
                    let expected = hex_u32(vector.encoded_hex.as_deref().unwrap_or(&vector.hex));
                    if encoded != expected {
                        failures.push(format!(
                            "bid32 encode {}: got {:08x}, want {:08x}",
                            vector.hex, encoded, expected
                        ));
                        continue;
                    }
                    encode_passed += 1;
                } else {
                    skipped += 1;
                }
            }
            "bid64" => {
                bid64 += 1;
                if vector.canonical {
                    bid64_canonical += 1;
                }
                let bits = hex_u64(&vector.hex);
                let decoded = decode64(bits);
                if decoded.kind != expected_kind
                    || decoded.sign != vector.sign
                    || decoded.exponent != vector.exponent
                    || decoded.coefficient != coeff_u128(&vector.coefficient)
                {
                    failures.push(format!(
                        "bid64 decode {}: got kind={:?} sign={} coeff={} exp={}, want kind={:?} sign={} coeff={} exp={}",
                        vector.hex,
                        decoded.kind,
                        decoded.sign,
                        decoded.coefficient,
                        decoded.exponent,
                        expected_kind,
                        vector.sign,
                        coeff_u128(&vector.coefficient),
                        vector.exponent,
                    ));
                    continue;
                }
                if matches!(decoded.kind, Kind::QNaN | Kind::SNaN) && decoded.payload != payload_u128(vector) {
                    failures.push(format!(
                        "bid64 decode payload {}: got {}, want {}",
                        vector.hex,
                        decoded.payload,
                        payload_u128(vector)
                    ));
                    continue;
                }
                decode_passed += 1;
                if to_string(&decoded).expect("decoded BID64 Components must render") != vector.decimal_string {
                    failures.push(format!(
                        "bid64 to_string {}: got {}, want {}",
                        vector.hex,
                        to_string(&decoded).expect("decoded BID64 Components must render"),
                        vector.decimal_string
                    ));
                    continue;
                }
                let parsed = from_string(&vector.decimal_string).unwrap_or_else(|err| {
                    panic!("bid64 from_string {} failed: {}", vector.decimal_string, err)
                });
                let parsed_encoded = encode64(&parsed).expect("from_string encode64");
                let expected_from_string = hex_u64(vector.encoded_hex.as_deref().unwrap_or(&vector.hex));
                if parsed_encoded != expected_from_string {
                    failures.push(format!(
                        "bid64 from_string {}: got {:016x}, want {:016x}",
                        vector.decimal_string,
                        parsed_encoded,
                        expected_from_string
                    ));
                    continue;
                }

                if vector.canonical {
                    let encoded = encode64(&decoded).expect("canonical encode64");
                    let expected = hex_u64(vector.encoded_hex.as_deref().unwrap_or(&vector.hex));
                    if encoded != expected {
                        failures.push(format!(
                            "bid64 encode {}: got {:016x}, want {:016x}",
                            vector.hex, encoded, expected
                        ));
                        continue;
                    }
                    encode_passed += 1;
                } else {
                    skipped += 1;
                }
            }
            "bid128" => {
                bid128 += 1;
                if vector.canonical {
                    bid128_canonical += 1;
                }
                let lo = hex_u64(&vector.hex);
                let hi = hex_u64(vector.hex_hi.as_deref().expect("missing bid128 hex_hi"));
                let decoded = decode128(lo, hi);
                if decoded.kind != expected_kind
                    || decoded.sign != vector.sign
                    || decoded.exponent != vector.exponent
                    || (!matches!(decoded.kind, Kind::QNaN | Kind::SNaN)
                        && decoded.coefficient != coeff_u128(&vector.coefficient))
                {
                    failures.push(format!(
                        "bid128 decode {}_{}: got kind={:?} sign={} coeff={} exp={}, want kind={:?} sign={} coeff={} exp={}",
                        vector.hex_hi.as_deref().unwrap_or(""),
                        vector.hex,
                        decoded.kind,
                        decoded.sign,
                        decoded.coefficient,
                        decoded.exponent,
                        expected_kind,
                        vector.sign,
                        coeff_u128(&vector.coefficient),
                        vector.exponent,
                    ));
                    continue;
                }
                if matches!(decoded.kind, Kind::QNaN | Kind::SNaN) && decoded.payload != payload_u128(vector) {
                    failures.push(format!(
                        "bid128 decode payload {}_{}: got {}, want {}",
                        vector.hex_hi.as_deref().unwrap_or(""),
                        vector.hex,
                        decoded.payload,
                        payload_u128(vector)
                    ));
                    continue;
                }
                decode_passed += 1;
                if to_string(&decoded).expect("decoded BID128 Components must render") != vector.decimal_string {
                    failures.push(format!(
                        "bid128 to_string {}_{}: got {}, want {}",
                        vector.hex_hi.as_deref().unwrap_or(""),
                        vector.hex,
                        to_string(&decoded).expect("decoded BID128 Components must render"),
                        vector.decimal_string
                    ));
                    continue;
                }
                let parsed = from_string(&vector.decimal_string).unwrap_or_else(|err| {
                    panic!("bid128 from_string {} failed: {}", vector.decimal_string, err)
                });
                let (parsed_lo, parsed_hi) = encode128(&parsed).expect("from_string encode128");
                let expected_lo = hex_u64(vector.encoded_hex.as_deref().unwrap_or(&vector.hex));
                let expected_hi = hex_u64(vector.encoded_hi.as_deref().unwrap_or_else(|| vector.hex_hi.as_deref().unwrap()));
                if parsed_lo != expected_lo || parsed_hi != expected_hi {
                    failures.push(format!(
                        "bid128 from_string {}: got {:016x}_{:016x}, want {:016x}_{:016x}",
                        vector.decimal_string,
                        parsed_hi,
                        parsed_lo,
                        expected_hi,
                        expected_lo,
                    ));
                    continue;
                }

                if vector.canonical {
                    let (enc_lo, enc_hi) = encode128(&decoded).expect("canonical encode128");
                    let exp_lo = hex_u64(vector.encoded_hex.as_deref().unwrap_or(&vector.hex));
                    let exp_hi = hex_u64(vector.encoded_hi.as_deref().unwrap_or_else(|| vector.hex_hi.as_deref().unwrap()));
                    if enc_lo != exp_lo || enc_hi != exp_hi {
                        failures.push(format!(
                            "bid128 encode {}_{}: got {:016x}_{:016x}, want {:016x}_{:016x}",
                            vector.hex_hi.as_deref().unwrap_or(""),
                            vector.hex,
                            enc_hi,
                            enc_lo,
                            exp_hi,
                            exp_lo,
                        ));
                        continue;
                    }
                    encode_passed += 1;
                } else {
                    skipped += 1;
                }
            }
            other => failures.push(format!("unknown vector type: {other}")),
        }
    }

    assert_eq!(vectors.len(), EXPECTED_TOTAL, "BID codec vector total changed");
    assert_eq!(bid32, EXPECTED_BID32, "BID32 vector count changed");
    assert_eq!(bid64, EXPECTED_BID64, "BID64 vector count changed");
    assert_eq!(bid128, EXPECTED_BID128, "BID128 vector count changed");
    assert_eq!(bid32_canonical, EXPECTED_BID32_CANONICAL, "BID32 canonical vector count changed");
    assert_eq!(bid64_canonical, EXPECTED_BID64_CANONICAL, "BID64 canonical vector count changed");
    assert_eq!(bid128_canonical, EXPECTED_BID128_CANONICAL, "BID128 canonical vector count changed");
    assert_eq!(skipped, EXPECTED_TOTAL - EXPECTED_BID32_CANONICAL - EXPECTED_BID64_CANONICAL - EXPECTED_BID128_CANONICAL, "non-canonical skip count changed");
    assert_eq!(decode_passed, EXPECTED_TOTAL, "decode coverage changed");
    assert_eq!(encode_passed, EXPECTED_BID32_CANONICAL + EXPECTED_BID64_CANONICAL + EXPECTED_BID128_CANONICAL, "encode coverage changed");

    if !failures.is_empty() {
        let preview = failures.into_iter().take(50).collect::<Vec<_>>().join("\n");
        panic!(
            "BID codec vectors failed: decode_passed={} encode_passed={} skipped={} failures_preview:\n{}",
            decode_passed,
            encode_passed,
            skipped,
            preview,
        );
    }

    eprintln!(
        "BID codec vectors: decode_passed={} encode_passed={} skipped={}",
        decode_passed,
        encode_passed,
        skipped,
    );
}
`)),
		bidCodecVectorsGoFullTestPath: []byte(applyBidCodecConsumerTemplateReplacements(genmarker.Line("testgen") + `
//go:build bid754_bidcodec_vectors

// Additional (non-required) BID codec vector consumer: the bid754-go full
// decimal library's PUBLIC parse surface (NewDecimal{32,64,128},
// NewDecimal*WithFlags, NewDecimal*WithMode). bid754-go has no embedded
// Components codec, so unlike the bid754-rs ` + "`rust_full`" + ` consumer this runner
// does not re-verify the codec contract; it holds the public fromString
// contract to the generated no-silent-failure domain. It consumes the
// ` + "`reject_vectors`" + ` from_string channel and the ` + "`string_vectors`" + ` channel
// against generator-owned expectation classes; the encode and to_string
// channels are channel-skipped (and counted) because the library exposes no
// public Components construction surface. Expected observation classes:
//
//   - exact:    Direct succeeds; WithFlags succeeds with zero flags and the
//               same bits; WithMode(RoundNearestEven) matches WithFlags; and
//               the public render/parse closure holds (Direct(v.String())==v).
//   - rounded:  Direct errors with a zero value (exact-only contract);
//               WithFlags succeeds with nonzero flags (the IEEE flag channel
//               reports the range/precision excursion instead of silence);
//               WithMode matches WithFlags.
//   - rejected: every family errors with a zero value and zero flags (public
//               grammar violation, silent-cohort trap, or a NaN payload
//               outside the width's range).
//
// The external test package makes the public-only access structural: nothing
// below can reach an internal identifier.
package bid754_test

import (
	"encoding/json"
	"os"
	"testing"

	bid754 "github.com/sky1core/bid754/bid754-go"
)

type goFullRejectEntry struct {
	Channel string ` + "`json:\"channel\"`" + `
	Input   string ` + "`json:\"input\"`" + `
	Reason  string ` + "`json:\"reason\"`" + `
}

type goFullStringEntry struct {
	Input    string ` + "`json:\"input\"`" + `
	Expected string ` + "`json:\"expected\"`" + `
}

type goFullVectorFile struct {
	FormatVersion int                 ` + "`json:\"format_version\"`" + `
	RejectVectors []goFullRejectEntry ` + "`json:\"reject_vectors\"`" + `
	StringVectors []goFullStringEntry ` + "`json:\"string_vectors\"`" + `
}

const (
	goFullExpectedFormatVersion = {{BID_CODEC_VECTOR_FORMAT_VERSION}}
	goFullExpectedRejectTotal   = {{BID_CODEC_REJECT_TOTAL}}
	goFullExpectedRejectConsumed = {{BID_CODEC_GO_FULL_REJECT_CONSUMED}}
	goFullExpectedRejectSkipped  = {{BID_CODEC_GO_FULL_REJECT_SKIPPED}}
	goFullExpectedStringTotal    = {{BID_CODEC_STRING_TOTAL}}
)

// goFullFromStringClasses pins the go_full class of every reject_vectors
// from_string input (width-independent on this channel).
var goFullFromStringClasses = map[string]string{ {{BID_CODEC_GO_FULL_FROM_STRING_CLASSES}} }

// goFullStringVectorClasses pins the go_full classes of every string_vectors
// input for widths 32, 64, and 128 in that order.
var goFullStringVectorClasses = map[string][3]string{ {{BID_CODEC_GO_FULL_STRING_CLASSES}} }

func goFullLoadVectors(t *testing.T) goFullVectorFile {
	t.Helper()
	data, err := os.ReadFile("../bid754-codec-vectors/vectors.json")
	if err != nil {
		t.Fatalf("failed to read vectors.json: %v", err)
	}
	var file goFullVectorFile
	if err := json.Unmarshal(data, &file); err != nil {
		t.Fatalf("failed to parse vectors.json: %v", err)
	}
	if file.FormatVersion != goFullExpectedFormatVersion {
		t.Fatalf("unsupported BID codec vectors format_version %d, want %d", file.FormatVersion, goFullExpectedFormatVersion)
	}
	return file
}

func goFullAssertClass32(t *testing.T, label, input, class string) {
	t.Helper()
	d, derr := bid754.NewDecimal32(input)
	w, wf, werr := bid754.NewDecimal32WithFlags(input)
	m, mf, merr := bid754.NewDecimal32WithMode(input, bid754.RoundNearestEven)
	if m != w || mf != wf || (merr == nil) != (werr == nil) {
		t.Errorf("%s %q d32: WithMode(NearestEven) = (%#08x, %v, err=%v) diverges from WithFlags (%#08x, %v, err=%v)",
			label, input, uint32(m), mf, merr, uint32(w), wf, werr)
		return
	}
	switch class {
	case "exact":
		if derr != nil || werr != nil || wf != 0 || w != d {
			t.Errorf("%s %q d32: want exact accept, got Direct(err=%v) WithFlags(%#08x, %v, err=%v) Direct value %#08x",
				label, input, derr, uint32(w), wf, werr, uint32(d))
			return
		}
		rendered := d.String()
		rt, rterr := bid754.NewDecimal32(rendered)
		if rterr != nil || rt != d {
			t.Errorf("%s %q d32: render/parse closure broken: String()=%q reparsed=(%#08x, err=%v), want %#08x",
				label, input, rendered, uint32(rt), rterr, uint32(d))
		}
	case "rounded":
		if derr == nil || d != 0 {
			t.Errorf("%s %q d32: want exact-channel reject, got Direct (%#08x, err=%v)", label, input, uint32(d), derr)
		}
		if werr != nil || wf == 0 {
			t.Errorf("%s %q d32: want flag-channel accept with nonzero flags, got WithFlags (%#08x, %v, err=%v)",
				label, input, uint32(w), wf, werr)
		}
	case "rejected":
		if derr == nil || d != 0 {
			t.Errorf("%s %q d32: want Direct reject, got (%#08x, err=%v)", label, input, uint32(d), derr)
		}
		if werr == nil || w != 0 || wf != 0 {
			t.Errorf("%s %q d32: want WithFlags reject with zero value and flags, got (%#08x, %v, err=%v)",
				label, input, uint32(w), wf, werr)
		}
	default:
		t.Fatalf("%s %q d32: unknown go_full class %q", label, input, class)
	}
}

func goFullAssertClass64(t *testing.T, label, input, class string) {
	t.Helper()
	d, derr := bid754.NewDecimal64(input)
	w, wf, werr := bid754.NewDecimal64WithFlags(input)
	m, mf, merr := bid754.NewDecimal64WithMode(input, bid754.RoundNearestEven)
	if m != w || mf != wf || (merr == nil) != (werr == nil) {
		t.Errorf("%s %q d64: WithMode(NearestEven) = (%#016x, %v, err=%v) diverges from WithFlags (%#016x, %v, err=%v)",
			label, input, uint64(m), mf, merr, uint64(w), wf, werr)
		return
	}
	switch class {
	case "exact":
		if derr != nil || werr != nil || wf != 0 || w != d {
			t.Errorf("%s %q d64: want exact accept, got Direct(err=%v) WithFlags(%#016x, %v, err=%v) Direct value %#016x",
				label, input, derr, uint64(w), wf, werr, uint64(d))
			return
		}
		rendered := d.String()
		rt, rterr := bid754.NewDecimal64(rendered)
		if rterr != nil || rt != d {
			t.Errorf("%s %q d64: render/parse closure broken: String()=%q reparsed=(%#016x, err=%v), want %#016x",
				label, input, rendered, uint64(rt), rterr, uint64(d))
		}
	case "rounded":
		if derr == nil || d != 0 {
			t.Errorf("%s %q d64: want exact-channel reject, got Direct (%#016x, err=%v)", label, input, uint64(d), derr)
		}
		if werr != nil || wf == 0 {
			t.Errorf("%s %q d64: want flag-channel accept with nonzero flags, got WithFlags (%#016x, %v, err=%v)",
				label, input, uint64(w), wf, werr)
		}
	case "rejected":
		if derr == nil || d != 0 {
			t.Errorf("%s %q d64: want Direct reject, got (%#016x, err=%v)", label, input, uint64(d), derr)
		}
		if werr == nil || w != 0 || wf != 0 {
			t.Errorf("%s %q d64: want WithFlags reject with zero value and flags, got (%#016x, %v, err=%v)",
				label, input, uint64(w), wf, werr)
		}
	default:
		t.Fatalf("%s %q d64: unknown go_full class %q", label, input, class)
	}
}

func goFullAssertClass128(t *testing.T, label, input, class string) {
	t.Helper()
	var zero bid754.Decimal128BID
	d, derr := bid754.NewDecimal128(input)
	w, wf, werr := bid754.NewDecimal128WithFlags(input)
	m, mf, merr := bid754.NewDecimal128WithMode(input, bid754.RoundNearestEven)
	if m != w || mf != wf || (merr == nil) != (werr == nil) {
		t.Errorf("%s %q d128: WithMode(NearestEven) = (%x, %v, err=%v) diverges from WithFlags (%x, %v, err=%v)",
			label, input, m.ToBytes(), mf, merr, w.ToBytes(), wf, werr)
		return
	}
	switch class {
	case "exact":
		if derr != nil || werr != nil || wf != 0 || w != d {
			t.Errorf("%s %q d128: want exact accept, got Direct(err=%v) WithFlags(%x, %v, err=%v) Direct value %x",
				label, input, derr, w.ToBytes(), wf, werr, d.ToBytes())
			return
		}
		rendered := d.String()
		rt, rterr := bid754.NewDecimal128(rendered)
		if rterr != nil || rt != d {
			t.Errorf("%s %q d128: render/parse closure broken: String()=%q reparsed=(%x, err=%v), want %x",
				label, input, rendered, rt.ToBytes(), rterr, d.ToBytes())
		}
	case "rounded":
		if derr == nil || d != zero {
			t.Errorf("%s %q d128: want exact-channel reject, got Direct (%x, err=%v)", label, input, d.ToBytes(), derr)
		}
		if werr != nil || wf == 0 {
			t.Errorf("%s %q d128: want flag-channel accept with nonzero flags, got WithFlags (%x, %v, err=%v)",
				label, input, w.ToBytes(), wf, werr)
		}
	case "rejected":
		if derr == nil || d != zero {
			t.Errorf("%s %q d128: want Direct reject, got (%x, err=%v)", label, input, d.ToBytes(), derr)
		}
		if werr == nil || w != zero || wf != 0 {
			t.Errorf("%s %q d128: want WithFlags reject with zero value and flags, got (%x, %v, err=%v)",
				label, input, w.ToBytes(), wf, werr)
		}
	default:
		t.Fatalf("%s %q d128: unknown go_full class %q", label, input, class)
	}
}

func TestGoFullBidCodecRejectVectors(t *testing.T) {
	file := goFullLoadVectors(t)
	if len(file.RejectVectors) != goFullExpectedRejectTotal {
		t.Fatalf("reject_vectors total = %d, want %d", len(file.RejectVectors), goFullExpectedRejectTotal)
	}
	consumed, skipped := 0, 0
	skipChannels := map[string]int{}
	for _, r := range file.RejectVectors {
		switch r.Channel {
		case "from_string":
			consumed++
			class, ok := goFullFromStringClasses[r.Input]
			if !ok {
				t.Fatalf("reject from_string input %q (%s) has no go_full expectation class", r.Input, r.Reason)
			}
			label := "reject from_string (" + r.Reason + ")"
			goFullAssertClass32(t, label, r.Input, class)
			goFullAssertClass64(t, label, r.Input, class)
			goFullAssertClass128(t, label, r.Input, class)
		case "encode", "to_string":
			// Channel skip: bid754-go exposes no public Components
			// construction surface, so these channels have no go_full analog.
			skipped++
			skipChannels[r.Channel]++
		default:
			t.Fatalf("unknown reject channel %q", r.Channel)
		}
	}
	if consumed != goFullExpectedRejectConsumed || skipped != goFullExpectedRejectSkipped || consumed+skipped != len(file.RejectVectors) {
		t.Fatalf("go_full reject consumption = consumed %d skipped %d of %d, want consumed %d skipped %d",
			consumed, skipped, len(file.RejectVectors), goFullExpectedRejectConsumed, goFullExpectedRejectSkipped)
	}
	if len(goFullFromStringClasses) != consumed {
		t.Fatalf("go_full from_string expectation table has %d entries, want %d (stale or missing entries)",
			len(goFullFromStringClasses), consumed)
	}
	t.Logf("go_full reject_vectors: consumed=%d channel_skipped=%d skipChannels=%v", consumed, skipped, skipChannels)
}

func TestGoFullBidCodecStringVectors(t *testing.T) {
	file := goFullLoadVectors(t)
	if len(file.StringVectors) != goFullExpectedStringTotal {
		t.Fatalf("string_vectors total = %d, want %d", len(file.StringVectors), goFullExpectedStringTotal)
	}
	consumed := 0
	for _, sv := range file.StringVectors {
		consumed++
		classes, ok := goFullStringVectorClasses[sv.Input]
		if !ok {
			t.Fatalf("string_vectors input %q has no go_full expectation classes", sv.Input)
		}
		goFullAssertClass32(t, "string_vectors", sv.Input, classes[0])
		goFullAssertClass64(t, "string_vectors", sv.Input, classes[1])
		goFullAssertClass128(t, "string_vectors", sv.Input, classes[2])
	}
	if consumed != goFullExpectedStringTotal {
		t.Fatalf("go_full string_vectors consumption = %d, want %d", consumed, goFullExpectedStringTotal)
	}
	if len(goFullStringVectorClasses) != consumed {
		t.Fatalf("go_full string_vectors expectation table has %d entries, want %d (stale or missing entries)",
			len(goFullStringVectorClasses), consumed)
	}
	t.Logf("go_full string_vectors: consumed=%d", consumed)
}
`)),
	}
	files[bidCodecVectorsGoExternalTestPath] = bidCodecGoExternalVectorTestOutput(files[bidCodecVectorsGoTestPath])
	return formatGeneratedGoOutputs(files)
}

func bidCodecGoExternalVectorTestOutput(repoTest []byte) []byte {
	src := string(repoTest)
	out := strings.Replace(src,
		"package bidcodec\n\n",
		"package main\n\n",
		1,
	)
	if out == src {
		panic("failed to derive external Go BID codec vector test package")
	}
	out = strings.Replace(out,
		"\t\"testing\"\n)",
		"\t\"testing\"\n\n\tcodec \"github.com/sky1core/bid754/bid754-codec-go\"\n)",
		1,
	)
	if !strings.Contains(out, `codec "github.com/sky1core/bid754/bid754-codec-go"`) {
		panic("failed to derive external Go BID codec vector test import")
	}
	out = strings.Replace(out,
		")\n\ntype vectorEntry",
		`)

type Components = codec.Components
type Kind = codec.Kind

const (
	Normal   = codec.Normal
	Zero     = codec.Zero
	Infinity = codec.Infinity
	QNaN     = codec.QNaN
	SNaN     = codec.SNaN
)

var (
	Decode32       = codec.Decode32
	Decode64       = codec.Decode64
	Decode128      = codec.Decode128
	Encode32       = codec.Encode32
	Encode64       = codec.Encode64
	Encode128      = codec.Encode128
	Decode32Bytes  = codec.Decode32Bytes
	Decode64Bytes  = codec.Decode64Bytes
	Decode128Bytes = codec.Decode128Bytes
	Encode32Bytes  = codec.Encode32Bytes
	Encode64Bytes  = codec.Encode64Bytes
	Encode128Bytes = codec.Encode128Bytes
	ToString       = codec.ToString
	FromString     = codec.FromString
)

type vectorEntry`,
		1,
	)
	if !strings.Contains(out, "type Components = codec.Components") {
		panic("failed to derive external Go BID codec vector test aliases")
	}
	return []byte(out)
}
